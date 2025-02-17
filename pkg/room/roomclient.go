package room

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/annidy/mediasoup-client/internal/gptr"
	"github.com/annidy/mediasoup-client/internal/utils"
	"github.com/annidy/mediasoup-client/pkg/client"
	"github.com/annidy/mediasoup-client/pkg/proto"
	"github.com/annidy/mediasoup-client/pkg/sdp"

	"github.com/gorilla/websocket"
	"github.com/jiyeyuran/go-protoo"
	"github.com/jiyeyuran/go-protoo/transport"
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/jiyeyuran/mediasoup-go/h264"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/h264reader"
	"github.com/pion/webrtc/v4/pkg/media/oggreader"
	"github.com/rs/zerolog/log"

	"github.com/pion/mediadevices/pkg/codec/opus" // This is required to use opus audio encoder
	"github.com/pion/mediadevices/pkg/codec/x264"
	_ "github.com/pion/mediadevices/pkg/driver/camera"     // This is required to register camera adapter
	_ "github.com/pion/mediadevices/pkg/driver/microphone" // This is required to register microphone adapter
	"github.com/pion/mediadevices/pkg/prop"
)

func init() {
	extra.RegisterFuzzyDecoders()
}

var json jsoniter.API = jsoniter.ConfigCompatibleWithStandardLibrary

type RoomClient struct {
	produce          bool // 是否推流
	consume          bool // 是否拉流
	peer             *Protoo
	peers            map[string]*proto.PeerData // 房间中的其它人
	device           *client.Device             // 包装设备
	sendTransport    *client.Transport
	recvTransport    *client.Transport
	closed           uint32
	micProducer      *client.Producer
	webcamProducer   *client.Producer
	chatDataProducer *client.DataProducer
	displayName      string
	enableMic        bool
	enableWebcam     bool
	enableLocalFile  bool

	codecSelector *mediadevices.CodecSelector
	mediaEngine   *webrtc.MediaEngine
}

func NewRoomClient() *RoomClient {
	x264Params, err := x264.NewParams()
	if err != nil {
		panic(err)
	}
	x264Params.Preset = x264.PresetMedium
	x264Params.BitRate = 500_000 // 500kbps

	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}
	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
		mediadevices.WithAudioEncoders(&opusParams),
	)

	return &RoomClient{
		produce:         true,
		consume:         true,
		displayName:     utils.RandomAlpha(8),
		peers:           make(map[string]*proto.PeerData),
		enableMic:       true,
		enableWebcam:    true,
		enableLocalFile: false,
		codecSelector:   codecSelector,
	}
}

const (
	audioFileName   = "output.ogg"
	videoFileName   = "output.h264"
	oggPageDuration = time.Millisecond * 20
)

// 加入房间
func (r *RoomClient) Join(wss string) {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", "protoo")

	log.Info().Msgf("Dial %s", wss)
	conn, _, err := dialer.Dial(wss, header)
	if err != nil {
		log.Err(err).Msg("Dial")
		return
	}

	transport := transport.NewWebsocketTransport(conn)

	peerData := proto.NewPeerData()
	peer := Protoo{Peer: protoo.NewPeer("-", peerData, transport)}
	r.peer = &peer

	peer.On("request", func(request protoo.Message, accept func(data interface{}), reject func(err error)) {
		log.Info().Str("method", request.Method).Msg("request event")

		err := r.handleProtooRequest(request, accept)
		if err != nil {
			reject(err)
		}
	})

	peer.On("open", func() {
		r.joinRoom()
	})

	peer.On("notification", func(notification protoo.Message) {
		log.Info().Interface("notification", notification).Msgf("notification")
	})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Err(r.(error)).Msg("Run panic")
			}
		}()
		err := transport.Run()
		log.Info().Err(err).Msg("Run stop")
	}()
	peer.Emit("open")
}

// 离开房间
func (r *RoomClient) Close() {
	// TODO: 发leaveRoom消息
	r.peer.Close()
}

// 拉流
func (r *RoomClient) Consume(peerId string) {
	r.consume = true
}

// https://github.com/pion/mediadevices/tree/master/examples/webrtc
func (r *RoomClient) EnableMic() {
	if r.micProducer != nil {
		return
	}
	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Audio: func(c *mediadevices.MediaTrackConstraints) {
			c.ChannelCount = prop.Int(2)
			c.SampleRate = prop.Int(48000)
		},
		Codec: r.codecSelector,
	})
	if err != nil {
		panic(err)
	}
	track := s.GetAudioTracks()[0]

	_ = r.sendTransport.Produce(client.TransportProduceOptions{
		Track: track,
		Codec: &mediasoup.RtpCodecParameters{
			MimeType:  webrtc.MimeTypeOpus,
			ClockRate: 48000,
			Channels:  2,
			Parameters: mediasoup.RtpCodecSpecificParameters{
				SpropStereo:  uint8(1),
				Useinbandfec: uint8(1),
				Usedtx:       uint8(1),
			},
		},
		CodecOptions: sdp.ProducerCodecOptions{
			OpusStereo: gptr.Of(true),
			OpusDtx:    gptr.Of(true),
			OpusFec:    gptr.Of(true),
			OpusNack:   gptr.Of(true),
		},
		OnRtpSender: nil,
		AppData:     map[string]any{},
	})
}

func (r *RoomClient) DisableMic() {
	if r.micProducer == nil {
		return
	}
	r.micProducer.Close()

	r.peer.Request("closeProducer", proto.CloseProducerRequest{
		ProducerId: r.micProducer.Id(),
	})

	r.micProducer = nil
}

func (r *RoomClient) MuteMic() {
	if r.micProducer == nil {
		return
	}
	r.micProducer.Pause()

	r.peer.Request("pauseProducer", proto.PauseProducerRequest{
		ProducerId: r.micProducer.Id(),
	})
}

func (r *RoomClient) UnmuteMic() {
	if r.micProducer == nil {
		return
	}
	r.micProducer.Resume()

	r.peer.Request("resumeProducer", proto.ResumeProducerRequest{
		ProducerId: r.micProducer.Id(),
	})
}

func (r *RoomClient) EnableWebcam() {
	if r.webcamProducer != nil {
		return
	}
	s, err := mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(c *mediadevices.MediaTrackConstraints) {
			c.FrameFormat = prop.FrameFormat(frame.FormatI420)
			c.Width = prop.Int(640)
			c.Height = prop.Int(480)
		},
		Audio: nil,
		Codec: r.codecSelector,
	})
	if err != nil {
		panic(err)
	}

	track := s.GetVideoTracks()[0]

	_ = r.sendTransport.Produce(client.TransportProduceOptions{
		Track: track,
		Codec: &mediasoup.RtpCodecParameters{
			MimeType:  webrtc.MimeTypeH264,
			ClockRate: 90000,
			Parameters: mediasoup.RtpCodecSpecificParameters{
				RtpParameter: h264.RtpParameter{
					ProfileLevelId:        "42e01f",
					PacketizationMode:     gptr.Of(uint8(1)),
					LevelAsymmetryAllowed: uint8(1),
				},
			},
		},
		CodecOptions: sdp.ProducerCodecOptions{
			VideoGoogleStartBitrate: gptr.Of(1000),
		},
		OnRtpSender: nil,
		AppData:     map[string]any{},
	})
}

func (r *RoomClient) MuteWebcam() {
	if r.webcamProducer == nil {
		return
	}
	r.webcamProducer.Pause()

	r.peer.Request("pauseProducer", proto.PauseProducerRequest{
		ProducerId: r.webcamProducer.Id(),
	})
}

func (r *RoomClient) UnmuteWebcam() {
	if r.webcamProducer == nil {
		return
	}
	r.webcamProducer.Resume()

	r.peer.Request("resumeProducer", proto.ResumeProducerRequest{
		ProducerId: r.webcamProducer.Id(),
	})
}

func (r *RoomClient) DisableWebcam() {
	if r.webcamProducer == nil {
		return
	}
	r.webcamProducer.Close()

	r.peer.Request("closeProducer", proto.CloseProducerRequest{
		ProducerId: r.webcamProducer.Id(),
	})

	r.webcamProducer = nil
}

func (r *RoomClient) EnableLocalFile() {
	_, err := os.Stat(videoFileName)
	haveVideoFile := !os.IsNotExist(err)

	_, err = os.Stat(audioFileName)
	haveAudioFile := !os.IsNotExist(err)

	if !haveAudioFile && !haveVideoFile {
		panic("Could not find `" + audioFileName + "` or `" + videoFileName + "`")
	}

	if haveVideoFile {
		// Create a video track
		videoTrack, videoTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
		if videoTrackErr != nil {
			panic(videoTrackErr)
		}

		_ = r.sendTransport.Produce(client.TransportProduceOptions{
			Track: videoTrack,
			Codec: &mediasoup.RtpCodecParameters{
				MimeType:  webrtc.MimeTypeH264,
				ClockRate: 90000,
				Parameters: mediasoup.RtpCodecSpecificParameters{
					RtpParameter: h264.RtpParameter{
						ProfileLevelId:        "42e01f",
						PacketizationMode:     gptr.Of(uint8(1)),
						LevelAsymmetryAllowed: uint8(1),
					},
				},
			},
			CodecOptions: sdp.ProducerCodecOptions{
				VideoGoogleStartBitrate: gptr.Of(1000),
			},
			OnRtpSender: func(rtpSender *webrtc.RTPSender) {
				// Read incoming RTCP packets
				// Before these packets are returned they are processed by interceptors. For things
				// like NACK this needs to be called.
				go func() {
					rtcpBuf := make([]byte, 1500)
					for {
						if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
							return
						}
					}
				}()
			},
			AppData: map[string]any{},
		})

		go func() {
			file, h264Err := os.Open(videoFileName)
			if h264Err != nil {
				panic(h264Err)
			}

			h264, h264Err := h264reader.NewReader(file)
			if h264Err != nil {
				panic(h264Err)
			}

			ticker := time.NewTicker(time.Millisecond * time.Duration(1000/24))
			defer ticker.Stop()
			for ; true; <-ticker.C {
				nal, h264Err := h264.NextNAL()
				if errors.Is(h264Err, io.EOF) {
					fmt.Printf("All video frames parsed and sent")
					os.Exit(0)
				}

				if h264Err != nil {
					panic(h264Err)
				}

				if h264Err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: time.Second}); h264Err != nil {
					panic(h264Err)
				}
			}
		}()
	}

	if haveAudioFile {
		file, oggErr := os.Open(audioFileName)
		if oggErr != nil {
			panic(oggErr)
		}

		_, header, oggErr := oggreader.NewWith(file)
		if oggErr != nil {
			panic(oggErr)
		}
		file.Close()

		// Create a audio track
		audioTrack, audioTrackErr := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
			MimeType:  webrtc.MimeTypeOpus,
			ClockRate: header.SampleRate,
			Channels:  uint16(header.Channels),
		}, "audio", "pion")
		fmt.Printf("%+v", audioTrack.Codec())
		if audioTrackErr != nil {
			panic(audioTrackErr)
		}

		_ = r.sendTransport.Produce(client.TransportProduceOptions{
			Track: audioTrack,
			Codec: &mediasoup.RtpCodecParameters{
				MimeType:  webrtc.MimeTypeOpus,
				ClockRate: int(header.SampleRate),
				Channels:  int(header.Channels),
				Parameters: mediasoup.RtpCodecSpecificParameters{
					SpropStereo:  uint8(1),
					Useinbandfec: uint8(1),
					Usedtx:       uint8(1),
				},
			},
			CodecOptions: sdp.ProducerCodecOptions{
				OpusStereo: gptr.Of(true),
				OpusDtx:    gptr.Of(true),
				OpusFec:    gptr.Of(true),
				OpusNack:   gptr.Of(true),
			},
			OnRtpSender: func(rtpSender *webrtc.RTPSender) {
				// Read incoming RTCP packets
				// Before these packets are returned they are processed by interceptors. For things
				// like NACK this needs to be called.
				go func() {
					rtcpBuf := make([]byte, 1500)
					for {
						if _, _, rtcpErr := rtpSender.Read(rtcpBuf); rtcpErr != nil {
							return
						}
					}
				}()
			},
			AppData: map[string]any{},
		})

		go func() {
			// Open a OGG file and start reading using our OGGReader
			file, oggErr := os.Open(audioFileName)
			if oggErr != nil {
				panic(oggErr)
			}

			// Open on oggfile in non-checksum mode.
			ogg, _, oggErr := oggreader.NewWith(file)
			if oggErr != nil {
				panic(oggErr)
			}

			// Keep track of last granule, the difference is the amount of samples in the buffer
			var lastGranule uint64

			// It is important to use a time.Ticker instead of time.Sleep because
			// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
			// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
			ticker := time.NewTicker(oggPageDuration)
			defer ticker.Stop()
			for ; true; <-ticker.C {
				pageData, pageHeader, oggErr := ogg.ParseNextPage()
				if errors.Is(oggErr, io.EOF) {
					fmt.Printf("All audio pages parsed and sent")
					os.Exit(0)
				}

				if oggErr != nil {
					panic(oggErr)
				}

				// The amount of samples is the difference between the last and current timestamp
				sampleCount := float64(pageHeader.GranulePosition - lastGranule)
				lastGranule = pageHeader.GranulePosition
				sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

				if oggErr = audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); oggErr != nil {
					panic(oggErr)
				}
			}
		}()
	}
}

func (r *RoomClient) enableChatDataProducer() {
	r.chatDataProducer = r.sendTransport.ProduceData(client.TransportProduceDataOptions{
		Ordered:        false,
		MaxRetransmits: 1,
		Label:          "chat",
		Priority:       "medium",
		AppData:        map[string]any{"info": "pion-chat"},
	})
	r.chatDataProducer.On("transportclose", func() {
		r.chatDataProducer = nil
	})
	r.chatDataProducer.On("close", func() {
		r.chatDataProducer = nil
	})
}

func (r *RoomClient) RestartIce() {
	if r.sendTransport != nil {
		var iceParameters client.IceParameters
		err := r.peer.RequestData("restartIce", proto.RestartIceRequest{
			TransportId: r.sendTransport.Id(),
		}, &iceParameters)

		if err != nil {
			log.Err(err).Msg("restartIce")
			return
		}
		r.sendTransport.RestartIce(iceParameters)
	}

	if r.recvTransport != nil {
		var iceParameters client.IceParameters
		err := r.peer.RequestData("restartIce", proto.RestartIceRequest{
			TransportId: r.recvTransport.Id(),
		}, &iceParameters)
		if err != nil {
			log.Err(err).Msg("restartIce")
			return
		}
		r.recvTransport.RestartIce(iceParameters)
	}
}

func (r *RoomClient) SendChat(msg string) {
	if r.chatDataProducer == nil {
		return
	}
	r.chatDataProducer.Send([]byte(msg))
}

// --

func (r *RoomClient) joinRoom() {
	r.device = client.NewDevice()

	var rtpCapabilities client.RtpCapabilities
	if err := r.peer.RequestData("getRouterRtpCapabilities", nil, &rtpCapabilities); err != nil {
		log.Err(err).Msg("getRouterRtpCapabilities")
		return
	}

	r.device.Load(rtpCapabilities)

	if r.produce {
		var transportOptions client.DeviceCreateTransportOptions

		if err := r.peer.RequestData("createWebRtcTransport", proto.CreateWebrtcTransportRequest{
			ForceTcp:         false,
			Producing:        true,
			Consuming:        false,
			SctpCapabilities: r.device.SctpCapabilities(),
		}, &transportOptions); err != nil {
			log.Err(err).Msg("createWebRtcTransport")
			return
		}

		r.sendTransport = r.device.CreateSendTransport(transportOptions)

		r.sendTransport.On("connect", func(dtlsParameters *client.DtlsParameters) {
			if err := r.peer.RequestData("connectWebRtcTransport", proto.ConnectWebRtcTransportRequest{
				TransportId:    transportOptions.Id,
				DtlsParameters: dtlsParameters,
			}, nil); err != nil {
				log.Err(err).Msg("connectWebRtcTransport")
				return
			}
		})
		r.sendTransport.On("produce", func(kind client.MediaKind, rtpParameters *client.RtpParameters, appData any) {
			type ProduceResponse struct {
				Id string
			}
			var rsp ProduceResponse

			if err := r.peer.RequestData("produce", proto.WebrtcTransportProducerRequest{
				TransportId:   transportOptions.Id,
				Kind:          kind,
				RtpParameters: rtpParameters,
				AppData:       appData,
			}, &rsp); err != nil {
				log.Err(err).Msg("produce")
				return
			}
			r.sendTransport.ProducerIdChan() <- rsp.Id
		})
		r.sendTransport.On("producedata", func(sctpStreamParameters *client.SctpStreamParameters, label string, protocol string, appData any) {
			r.peer.Request("produceData", proto.WebrtcTransportProducerDataRequest{
				TransportId:          transportOptions.Id,
				SctpStreamParameters: sctpStreamParameters,
				Label:                label,
				Protocol:             protocol,
				AppData:              appData,
			})
		})
	}

	if r.consume {

		var transportOptions client.DeviceCreateTransportOptions

		if err := r.peer.RequestData("createWebRtcTransport", proto.CreateWebrtcTransportRequest{
			ForceTcp:         false,
			Producing:        false,
			Consuming:        true,
			SctpCapabilities: r.device.SctpCapabilities(),
		}, &transportOptions); err != nil {
			log.Err(err).Msg("createWebRtcTransport")
			return
		}

		r.recvTransport = r.device.CreateRecvTransport(transportOptions)
		r.recvTransport.On("connect", func(dtlsParameters *client.DtlsParameters) {
			r.peer.Request("connectWebRtcTransport", proto.ConnectWebRtcTransportRequest{
				TransportId:    transportOptions.Id,
				DtlsParameters: dtlsParameters,
			})
		})
	}

	// 发Join消息
	type Peers struct {
		Peers []*proto.PeerData
	}
	var peers Peers
	if err := r.peer.RequestData("join", proto.JoinRequest{
		DisplayName:      r.displayName,
		Device:           r.device.DeviceInfo(),
		RtpCapabilities:  r.device.RtpCapabilities(),
		SctpCapabilities: r.device.SctpCapabilities(),
	}, &peers); err != nil {
		log.Err(err).Msg("join")
		return
	}
	for _, peer := range peers.Peers {
		r.peers[peer.Id] = peer
	}

	// Enable mic/webcam.
	if r.produce {
		if r.enableMic {
			r.EnableMic()
		}
		if r.enableWebcam {
			r.EnableWebcam()
		}
		if r.enableLocalFile {
			r.EnableLocalFile()
		}

		r.sendTransport.On("connectionstatechange", func(connectionState webrtc.PeerConnectionState) {
			if connectionState == webrtc.PeerConnectionStateConnected {
				r.enableChatDataProducer()
			}
		})
	}
}

func (r *RoomClient) handleProtooRequest(request protoo.Message, accept func(data interface{})) (err error) {
	peerData := r.peer.Data().(*proto.PeerData)
	ret := map[string]interface{}{}
	switch request.Method {
	case "newConsumer":
		consumer := proto.Consumer{}
		err = json.Unmarshal(request.Data, &consumer)
		if err != nil {
			log.Err(err).Msg("newConsumer")
			return
		}
		peerData.AddConsumer(&consumer)
	}
	accept(ret)
	return
}
