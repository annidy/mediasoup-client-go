package client

import (
	"github.com/annidy/mediasoup-client/pkg/sdp"
	"github.com/jiyeyuran/mediasoup-go"
)

type RemoteSdp struct {
	iceParameters  *IceParameters
	iceCandidates  []*IceCandidate
	dtlsParameters *DtlsParameters
	sctpParameters *SctpParameters
	// TODO: palin rtp parameters
	IsPlanB bool
	// MediaSection sdp.MediaSection
	MidToIndex map[string]int
	firstMid   string
	sdpObject  sdp.Sdp

	mediaSections []*sdp.MediaSection
}

// TODO: iceParameters、iceCandidates、sctpParameters
func NewRemoteSdp(iceParameters *IceParameters, iceCandidates []*IceCandidate, dtlsParameters *DtlsParameters, sctpParameters *SctpParameters) *RemoteSdp {
	rdp := &RemoteSdp{
		iceParameters:  iceParameters,
		iceCandidates:  iceCandidates,
		dtlsParameters: dtlsParameters,
		sctpParameters: sctpParameters,
		sdpObject: sdp.Sdp{
			Version: 0,
			Origin: struct {
				Address        string `json:"address,omitempty"`
				IPVer          int    `json:"ipVer,omitempty"`
				NetType        string `json:"netType,omitempty"`
				SessionID      int    `json:"sessionId,omitempty"`
				SessionVersion int    `json:"sessionVersion,omitempty"`
				Username       string `json:"username,omitempty"`
			}{
				Address:        "0.0.0.0",
				IPVer:          4,
				NetType:        "IN",
				SessionID:      10000,
				SessionVersion: 0,
				Username:       "mediasoup-client-go",
			},
			Name: "-",
		},
	}
	if iceParameters != nil {
		if iceParameters.IceLite {
			rdp.sdpObject.IceLite = "ice-lite"
		}
	}
	if dtlsParameters != nil {
		rdp.sdpObject.MsidSemantic = struct {
			Semantics string `json:"semantics,omitempty"`
			Token     string `json:"token,omitempty"`
		}{
			Semantics: "WMS",
			Token:     "*",
		}

		i := len(dtlsParameters.Fingerprints) - 1

		rdp.sdpObject.Fingerprint = struct {
			Hash string `json:"hash,omitempty"`
			Type string `json:"type,omitempty"`
		}{
			Type: dtlsParameters.Fingerprints[i].Algorithm,
			Hash: dtlsParameters.Fingerprints[i].Value,
		}

		rdp.sdpObject.Groups = append(rdp.sdpObject.Groups, struct {
			Mids string `json:"mids"`
			Type string `json:"type"`
		}{
			Mids: "",
			Type: "BUNDLE",
		})
	}
	// TODO: support plain RTP parameters

	return rdp
}

func (rdp *RemoteSdp) getNextMediaSectionIdx() (idx int, reuseMid string) {
	for i, sec := range rdp.mediaSections {
		if sec.Closed() {
			return i, sec.Mid()
		}
	}
	return len(rdp.mediaSections), ""
}

func (rdp *RemoteSdp) updateIceParameters(iceParameters *mediasoup.IceParameters) {
	rdp.iceParameters = iceParameters
	if iceParameters.IceLite {
		rdp.sdpObject.IceLite = "ice-lite"
	} else {
		rdp.sdpObject.IceLite = ""
	}
	for _, mediaSection := range rdp.mediaSections {
		mediaSection.SetIceParameters(iceParameters)
	}
}

type SendTransportOptions struct {
	offerMediaObject    interface{}
	reuseMid            string
	offerRtpParameters  mediasoup.RtpParameters
	answerRtpParameters mediasoup.RtpParameters
	codecOptions        []*mediasoup.RtpCodecParameters
	extmapAllowMixed    bool
}

func (rdp *RemoteSdp) send(options SendTransportOptions) {

}

func (rdp *RemoteSdp) getSdp() string {
	rdp.sdpObject.Origin.SessionVersion++

	return sdp.Write(rdp.sdpObject)
}
