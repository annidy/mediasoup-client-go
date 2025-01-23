package room

import (
	"crypto/tls"
	"net/http"

	"github.com/annidy/mediasoup-client/internal/util"
	"github.com/annidy/mediasoup-client/pkg/client"
	"github.com/annidy/mediasoup-client/pkg/proto"

	"github.com/gorilla/websocket"
	"github.com/jiyeyuran/go-protoo"
	"github.com/jiyeyuran/go-protoo/transport"
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
)

func init() {
	extra.RegisterFuzzyDecoders()
}

var json jsoniter.API = jsoniter.ConfigCompatibleWithStandardLibrary

type RoomClient struct {
	produce          bool // 是否推流
	consume          bool // 是否拉流
	peer             *Protoo
	peers            map[string]*protoo.Peer // 房间中的其它人
	device           *client.Device          // 包装设备
	sendTransport    *client.Transport
	recvTransport    *client.Transport
	closed           uint32
	micProducer      *client.Producer
	webcamProducer   *client.Producer
	chatDataProducer *client.DataProducer
}

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
		log.Info().Msgf("notification %v", notification)
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

func (r *RoomClient) EnableMic() {
	// TODO:
	// - getUserMedia
	// - getAudioTrack
	// - createProducer
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
	// TODO:
	// - getUserMedia
	// - getVideoTrack
	// - createProducer
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

func (r *RoomClient) enableChatDataProducer() {
	r.chatDataProducer = r.sendTransport.ProduceData(client.DataProducerOptions{
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
		var transportInfo client.WebrtcTransportInfo

		if err := r.peer.RequestData("createWebRtcTransport", proto.CreateWebrtcTransportRequest{
			ForceTcp:         false,
			Producing:        true,
			Consuming:        false,
			SctpCapabilities: r.device.SctpCapabilities(),
		}, &transportInfo); err != nil {
			log.Err(err).Msg("createWebRtcTransport")
			return
		}

		r.sendTransport = r.device.CreateSendTransport(transportInfo)

		r.sendTransport.On("connect", func(dtlsParameters *client.DtlsParameters) {
			r.peer.Request("connectWebRtcTransport", proto.ConnectWebRtcTransportRequest{
				TransportId:    transportInfo.Id,
				DtlsParameters: dtlsParameters,
			})
		})
		r.sendTransport.On("produce", func(kind client.MediaKind, rtpParameters *client.RtpParameters, appData any) {
			r.peer.Request("produce", proto.WebrtcTransportProducerRequest{
				TransportId:   transportInfo.Id,
				Kind:          kind,
				RtpParameters: rtpParameters,
				AppData:       appData,
			})
		})
		r.sendTransport.On("producedata", func(sctpStreamParameters *client.SctpStreamParameters, label string, protocol string, appData any) {
			r.peer.Request("produceData", proto.WebrtcTransportProducerDataRequest{
				TransportId:          transportInfo.Id,
				SctpStreamParameters: sctpStreamParameters,
				Label:                label,
				Protocol:             protocol,
				AppData:              appData,
			})
		})
	}

	if r.consume {

		var transportInfo client.WebrtcTransportInfo

		if err := r.peer.RequestData("createWebRtcTransport", proto.CreateWebrtcTransportRequest{
			ForceTcp:         false,
			Producing:        false,
			Consuming:        true,
			SctpCapabilities: r.device.SctpCapabilities(),
		}, &transportInfo); err != nil {
			log.Err(err).Msg("createWebRtcTransport")
			return
		}

		r.recvTransport = r.device.CreateRecvTransport(transportInfo)
		r.recvTransport.On("connect", func(dtlsParameters *client.DtlsParameters) {
			r.peer.Request("connectWebRtcTransport", proto.ConnectWebRtcTransportRequest{
				TransportId:    transportInfo.Id,
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
		DisplayName:      util.RandomAlpha(8),
		Device:           r.device.DeviceInfo(),
		RtpCapabilities:  r.device.RtpCapabilities(),
		SctpCapabilities: r.device.SctpCapabilities(),
	}, &peers); err != nil {
		log.Err(err).Msg("join")
		return
	}
	for _, peer := range peers.Peers {
		r.peers[peer.Id] = protoo.NewPeer(peer.Id, peer, nil)
	}

	// Enable mic/webcam.
	if r.produce {
		r.EnableMic()
		r.EnableWebcam()

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
