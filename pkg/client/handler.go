package client

import (
	"fmt"

	"github.com/annidy/mediasoup-client/internal/utils"
	"github.com/annidy/mediasoup-client/pkg/sdp"
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/mediadevices/pkg/codec/x264"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
)

type HandlerRunOptions struct {
	direction              string
	iceParameters          *IceParameters
	iceCandidates          []*IceCandidate
	dtlsParameters         *DtlsParameters
	sctpParameters         *SctpParameters
	extenedRtpCapabilities *RtpCapabilitiesEx
}

type Handler interface {
	getNativeRtpCapabilities() mediasoup.RtpCapabilities
	getNativeRouterRtpCapabilities() mediasoup.RtpCapabilities
	getNativeSctpCapabilities() mediasoup.SctpCapabilities
	run(options HandlerRunOptions)
	restartIce(iceParameters *mediasoup.IceParameters)
	close() error
}

type HandlerFactory func() Handler

type PionHandler struct {
	mediasoup.IEventEmitter
	direction                        string
	remoteSdp                        *RemoteSdp
	sendingRtpParametersByKind       map[string]*RtpParameters
	sendingRemoteRtpParametersByKind map[string]*RtpParameters
	pc                               *webrtc.PeerConnection
	mapMidTransceiver                map[string]*webrtc.RTPTransceiver
	transportReady                   bool
	api                              *webrtc.API
}

var _ Handler = (*PionHandler)(nil)

func NewPionHandler() *PionHandler {
	x264Params, err := x264.NewParams()
	if err != nil {
		panic(err)
	}
	opusParams, err := opus.NewParams()
	if err != nil {
		panic(err)
	}
	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&x264Params),
		mediadevices.WithAudioEncoders(&opusParams),
	)

	mediaEngine := webrtc.MediaEngine{}
	codecSelector.Populate(&mediaEngine)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))

	return &PionHandler{
		IEventEmitter:     mediasoup.NewEventEmitter(),
		mapMidTransceiver: make(map[string]*webrtc.RTPTransceiver),
		api:               api,
	}
}

func (h *PionHandler) getNativeRtpCapabilities() RtpCapabilities {
	return h.getNativeRouterRtpCapabilities()
}

func (h *PionHandler) getNativeRouterRtpCapabilities() RtpCapabilities {
	pc, err := h.api.NewPeerConnection(webrtc.Configuration{
		ICEServers:         []webrtc.ICEServer{},
		ICETransportPolicy: webrtc.ICETransportPolicyAll,
		BundlePolicy:       webrtc.BundlePolicyBalanced,
		RTCPMuxPolicy:      webrtc.RTCPMuxPolicyRequire,
		SDPSemantics:       webrtc.SDPSemanticsUnifiedPlan,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if cErr := pc.Close(); cErr != nil {
			log.Err(cErr).Msg("cannot close pc")
		}
	}()
	pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
	pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		panic(err)
	}
	pc.Close()

	fmt.Println(offer.SDP)

	sdpObject := sdp.Parse(offer.SDP)
	nativeRtpCapabilities := extractRtpCapabilities(sdpObject)

	// libwebrtc supports NACK for OPUS but doesn't announce it.
	ortc.addNackSupportForOpus(&nativeRtpCapabilities)

	return nativeRtpCapabilities
}

func (h *PionHandler) getNativeSctpCapabilities() mediasoup.SctpCapabilities {
	return mediasoup.SctpCapabilities{
		NumStreams: mediasoup.NumSctpStreams{
			OS:  1024,
			MIS: 1024,
		},
	}
}

func (h *PionHandler) run(options HandlerRunOptions) {
	iceParameters, iceCandidates, dtlsParameters, sctpParameters, direction, extenedRtpCapabilities := options.iceParameters, options.iceCandidates, options.dtlsParameters, options.sctpParameters, options.direction, options.extenedRtpCapabilities

	h.direction = direction
	h.remoteSdp = NewRemoteSdp(RemoteSdpOptions{
		iceParameters:  iceParameters,
		iceCandidates:  iceCandidates,
		dtlsParameters: dtlsParameters,
		sctpParameters: sctpParameters,
	})
	h.sendingRtpParametersByKind = map[string]*RtpParameters{
		"audio": ortc.getSendingRtpParameters(mediasoup.MediaKind_Audio, extenedRtpCapabilities),
		"video": ortc.getSendingRtpParameters(mediasoup.MediaKind_Video, extenedRtpCapabilities),
	}
	h.sendingRemoteRtpParametersByKind = map[string]*RtpParameters{
		"audio": ortc.getSendingRemoteRtpParameters(mediasoup.MediaKind_Audio, extenedRtpCapabilities),
		"video": ortc.getSendingRemoteRtpParameters(mediasoup.MediaKind_Video, extenedRtpCapabilities),
	}

	config := webrtc.Configuration{
		// TODO: iceServers
	}

	pc, err := h.api.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	pc.OnICEGatheringStateChange(func(state webrtc.ICEGatheringState) {
		log.Info().Str("state", state.String()).Msg("ICEGatheringStateChange")
		h.Emit("@icegatheringstatechange", state)
	})
	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		log.Info().Str("state", state.String()).Msg("ConnectionStateChange")
		h.Emit("@connectionstatechange", state)
	})
	h.pc = pc
}

type HandlerSendOptions struct {
	track        webrtc.TrackLocal
	codecOptions sdp.ProducerCodecOptions
	codec        *mediasoup.RtpCodecParameters
	encodings    []mediasoup.RtpEncodingParameters
	onRtpSender  func(*webrtc.RTPSender)
}

func (h *PionHandler) send(options HandlerSendOptions) (localId string, rtpParameters *RtpParameters, rtpSender *webrtc.RTPSender) {
	track, codecOptions, codec, onRtpSender, encodings := options.track, options.codecOptions, options.codec, options.onRtpSender, options.encodings

	trackKind := track.Kind().String()
	log.Debug().Str("kind", trackKind).Str("track.id", track.ID()).Msg("send()")

	var sendingRtpParameters, sendingRemoteRtpParameters RtpParameters
	utils.Clone(h.sendingRtpParametersByKind[trackKind], &sendingRtpParameters)

	sendingRtpParameters.Codecs = ortc.reduceCodecs(sendingRtpParameters.Codecs, codec)

	utils.Clone(h.sendingRemoteRtpParametersByKind[trackKind], &sendingRemoteRtpParameters)
	sendingRemoteRtpParameters.Codecs = ortc.reduceCodecs(sendingRemoteRtpParameters.Codecs, codec)

	mediaSectionIdx, mediaSectionReuseMid := h.remoteSdp.getNextMediaSectionIdx()

	var webrtcEncodings []webrtc.RTPEncodingParameters
	for _, encoding := range encodings {
		webrtcEncodings = append(webrtcEncodings, webrtc.RTPEncodingParameters{
			RTPCodingParameters: webrtc.RTPCodingParameters{
				RID: encoding.Rid,
			},
			// TODO: pion不支持scaleResolutionDownBy、maxBitrate等，它实现simulcast的方案和标准不太一样
		})
	}

	transceiver, err := h.pc.AddTransceiverFromTrack(track, webrtc.RTPTransceiverInit{
		Direction:     webrtc.RTPTransceiverDirectionSendonly,
		SendEncodings: webrtcEncodings,
	})
	if err != nil {
		panic(err)
	}
	if onRtpSender != nil {
		onRtpSender(transceiver.Sender())
	}

	offer, err := h.pc.CreateOffer(nil)
	if err != nil {
		panic(err)
	}
	localSdpObject := sdp.Parse(offer.SDP)

	// fmt.Println(offer.SDP)

	if !h.transportReady {
		h.setupTransport(mediasoup.DtlsRole_Client, localSdpObject)
	}

	if err := h.pc.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	// We can now get the transceiver.mid.
	localId = transceiver.Mid()

	// Set MID
	sendingRtpParameters.Mid = localId

	// maybe this is different with previous sdp
	localSdpObject = sdp.Parse(h.pc.LocalDescription().SDP)

	offerMediaObject := localSdpObject.Media[mediaSectionIdx]

	sendingRtpParameters.Rtcp.Cname = sdp.GetCname(offerMediaObject)

	// Set RTP encodings by parsing the SDP offer if no encodings are given.
	sendingRtpParameters.Encodings = sdp.GetRtpEncodings(offerMediaObject)

	h.remoteSdp.send(SendTransportOptions{
		offerMediaObject:    offerMediaObject,
		reuseMid:            mediaSectionReuseMid,
		offerRtpParameters:  sendingRtpParameters,
		answerRtpParameters: sendingRemoteRtpParameters,
		codecOptions:        codecOptions,
		extmapAllowMixed:    true,
	})

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  h.remoteSdp.getSdp(),
	}

	if err := h.pc.SetRemoteDescription(answer); err != nil {
		panic(err)
	}

	// Store in the map.
	h.mapMidTransceiver[localId] = transceiver

	return localId, &sendingRtpParameters, transceiver.Sender()
}

func (h *PionHandler) restartIce(iceParameters *mediasoup.IceParameters) {
	// Provide the remote SDP handler with new remote ICE parameters.
	h.remoteSdp.updateIceParameters(iceParameters)
}

func (h *PionHandler) setupTransport(localDtlsRole mediasoup.DtlsRole, localSdpObject sdp.Sdp) {

	dtlsParameters := extractDtlsParameters(localSdpObject)
	dtlsParameters.Role = localDtlsRole

	if localDtlsRole == mediasoup.DtlsRole_Client {
		h.remoteSdp.updateDtlsRole(mediasoup.DtlsRole_Server)
	} else {
		h.remoteSdp.updateDtlsRole(mediasoup.DtlsRole_Client)
	}

	h.SafeEmit("@connect", dtlsParameters)

	h.transportReady = true
}

func (h *PionHandler) close() error {
	if h.pc != nil {
		if err := h.pc.Close(); err != nil {
			log.Err(err).Msg("cannot close pc")
			return err
		}
		h.pc = nil
	}
	return nil
}
