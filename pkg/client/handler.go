package client

import (
	"fmt"

	"github.com/annidy/mediasoup-client/pkg/sdp"
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
)

type HandlerRunOptions struct {
}

type Handler interface {
	getNativeRtpCapabilities() mediasoup.RtpCapabilities
	getNativeRouterRtpCapabilities() mediasoup.RtpCapabilities
	getNativeSctpCapabilities() mediasoup.SctpCapabilities
	run(options HandlerRunOptions)
	restartIce(iceServers []mediasoup.IceParameters)
}

type PionHandler struct {
}

func NewPionHandler() *PionHandler {
	return &PionHandler{}
}

func (h *PionHandler) getNativeRtpCapabilities() mediasoup.RtpCapabilities {
	return mediasoup.RtpCapabilities{}
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
