package client

import (
	"fmt"

	"github.com/annidy/mediasoup-client/internal/utils"
	"github.com/annidy/mediasoup-client/pkg/sdp"
	"github.com/jiyeyuran/mediasoup-go"
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
}

type PionHandler struct {
	mediasoup.IEventEmitter
	direction                        string
	remoteSdp                        *RemoteSdp
	sendingRtpParametersByKind       map[string]*RtpParameters
	sendingRemoteRtpParametersByKind map[string]*RtpParameters
	pc                               *webrtc.PeerConnection
	mapMidTransceiver                map[string]*webrtc.RTPTransceiver
	transportReady                   bool
}

var _ Handler = (*PionHandler)(nil)

func NewPionHandler() *PionHandler {
	return &PionHandler{
		IEventEmitter:     mediasoup.NewEventEmitter(),
		mapMidTransceiver: make(map[string]*webrtc.RTPTransceiver),
	}
}

func (h *PionHandler) getNativeRtpCapabilities() RtpCapabilities {
	return h.getNativeRouterRtpCapabilities()
}

func (h *PionHandler) getNativeRouterRtpCapabilities() RtpCapabilities {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
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

	// log.Info().Str("SDP", offer.SDP).Msg("CreateOffer")

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
	h.direction = options.direction

	h.remoteSdp = NewRemoteSdp(options.iceParameters, options.iceCandidates, options.dtlsParameters, options.sctpParameters)
	h.sendingRtpParametersByKind = map[string]*RtpParameters{
		"audio": ortc.getSendingRtpParameters(mediasoup.MediaKind_Audio, options.extenedRtpCapabilities),
		"video": ortc.getSendingRtpParameters(mediasoup.MediaKind_Video, options.extenedRtpCapabilities),
	}
	h.sendingRemoteRtpParametersByKind = map[string]*RtpParameters{
		"audio": ortc.getSendingRemoteRtpParameters(mediasoup.MediaKind_Audio, options.extenedRtpCapabilities),
		"video": ortc.getSendingRemoteRtpParameters(mediasoup.MediaKind_Video, options.extenedRtpCapabilities),
	}

	config := webrtc.Configuration{
		// TODO: iceServers
	}

	pc, err := webrtc.NewPeerConnection(config)
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
	codecOptions []*mediasoup.RtpCodecParameters
	codec        *mediasoup.RtpCodecParameters
	onRtpSender  func(*webrtc.RTPSender)
}

func (h *PionHandler) send(options HandlerSendOptions) (localId string, rtpParameters *RtpParameters, rtpSender RTCRtpSender) {
	track, codecOptions, codec, onRtpSender := options.track, options.codecOptions, options.codec, options.onRtpSender

	var sendingRtpParameters, sendingRemoteRtpParameters RtpParameters
	utils.Clone(h.sendingRtpParametersByKind[track.Kind().String()], &sendingRtpParameters)

	sendingRtpParameters.Codecs = ortc.reduceCodecs(sendingRtpParameters.Codecs, codec)

	utils.Clone(h.sendingRemoteRtpParametersByKind[track.Kind().String()], &sendingRemoteRtpParameters)
	sendingRemoteRtpParameters.Codecs = ortc.reduceCodecs(sendingRemoteRtpParameters.Codecs, codec)

	mediaSectionIdx, mediaSectionReuseMid := h.remoteSdp.getNextMediaSectionIdx()

	// TODO: 获取Media过来的encodings
	transceiver, err := h.pc.AddTransceiverFromTrack(track, webrtc.RTPTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionSendonly,
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

	fmt.Println(offer.SDP)

	if !h.transportReady {
		h.setupTransport(mediasoup.DtlsRole_Client, localSdpObject)
	}

	h.pc.SetLocalDescription(offer)

	// We can now get the transceiver.mid.
	localId = transceiver.Mid()

	// Set MID
	sendingRtpParameters.Mid = localId

	// Why reparse? This is wrong.
	// localSdpObject = sdp.Parse(h.pc.LocalDescription().SDP)

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

	h.pc.SetRemoteDescription(answer)

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
