package client

import (
	"strings"

	"github.com/annidy/mediasoup-client/internal/gptr"
	"github.com/annidy/mediasoup-client/pkg/sdp"
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type RemoteSdp struct {
	iceParameters  *IceParameters
	iceCandidates  []*IceCandidate
	dtlsParameters *DtlsParameters
	sctpParameters *SctpParameters
	// TODO: palin rtp parameters
	IsPlanB bool
	// MediaSection sdp.MediaSection
	midToIndex map[string]int
	firstMid   string
	sdpObject  sdp.Sdp

	mediaSections []*sdp.MediaSection
}

type RemoteSdpOptions struct {
	iceParameters  *IceParameters
	iceCandidates  []*IceCandidate
	dtlsParameters *DtlsParameters
	sctpParameters *SctpParameters
}

func NewRemoteSdp(options RemoteSdpOptions) *RemoteSdp {
	iceParameters, iceCandidates, dtlsParameters, sctpParameters := options.iceParameters, options.iceCandidates, options.dtlsParameters, options.sctpParameters

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
		midToIndex: make(map[string]int),
	}
	if iceParameters != nil && iceParameters.IceLite {
		rdp.sdpObject.IceLite = "ice-lite"
	}
	if dtlsParameters != nil {
		rdp.sdpObject.MsidSemantic = gptr.Of(struct {
			Semantics string `json:"semantic"`
			Token     string `json:"token"`
		}{
			Semantics: "WMS",
			Token:     "*",
		})

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
	log.Debug().Msgf("updateIceParameters [iceParameters:%v]", iceParameters)
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
	offerMediaObject    sdp.MediaObject
	reuseMid            string
	offerRtpParameters  mediasoup.RtpParameters
	answerRtpParameters mediasoup.RtpParameters
	codecOptions        sdp.ProducerCodecOptions
	extmapAllowMixed    bool
}

func (rdp *RemoteSdp) send(options SendTransportOptions) {
	offerMediaObject, answerRtpParameters, codecOptions, reuseMid, extmapAllowMixed := options.offerMediaObject, options.answerRtpParameters, options.codecOptions, options.reuseMid, options.extmapAllowMixed
	mediaSection := sdp.NewAnswerMediaSection(sdp.AnswerMediaSectionOptions{
		MediaSectionOptions: sdp.MediaSectionOptions{
			IceParameters:  rdp.iceParameters,
			IceCandidates:  rdp.iceCandidates,
			DtlsParameters: rdp.dtlsParameters,
			PlanB:          rdp.IsPlanB,
		},
		OfferMediaObject:    offerMediaObject,
		AnswerRtpParameters: answerRtpParameters,
		CodecOptions:        codecOptions,
		ExtmapAllowMixed:    extmapAllowMixed,
	})

	// Unified-Plan with closed media section replacement.
	if reuseMid != "" {
		rdp.replaceMediaSection(&mediaSection.MediaSection, reuseMid)
	} else if _, exist := rdp.midToIndex[mediaSection.Mid()]; !exist {
		// Unified-Plan or Plan-B with different media kind.
		rdp.addMediaSection(&mediaSection.MediaSection)
	} else {
		// Plan-B with same media kind.
		rdp.replaceMediaSection(&mediaSection.MediaSection, "")
	}
}

func (rdp *RemoteSdp) getSdp() string {
	rdp.sdpObject.Origin.SessionVersion++

	return sdp.Write(rdp.sdpObject)
}

func (rdp *RemoteSdp) addMediaSection(newMediaSection *sdp.MediaSection) {
	if rdp.firstMid == "" {
		rdp.firstMid = newMediaSection.Mid()
	}
	rdp.mediaSections = append(rdp.mediaSections, newMediaSection)

	rdp.midToIndex[newMediaSection.Mid()] = len(rdp.mediaSections) - 1

	rdp.sdpObject.Media = append(rdp.sdpObject.Media, newMediaSection.MediaObject())

	rdp.regenerateBundleMids()
}

func (rdp *RemoteSdp) replaceMediaSection(newMediaSection *sdp.MediaSection, reuseMid string) {
	if reuseMid != "" {
		// Get the index of the old media section.
		idx, exist := rdp.midToIndex[reuseMid]
		if !exist {
			panic("no media section found for reuseMid")
		}
		oldMediaSection := rdp.mediaSections[idx]
		// Replace the media section.
		rdp.mediaSections[idx] = newMediaSection
		delete(rdp.midToIndex, oldMediaSection.Mid())
		rdp.midToIndex[newMediaSection.Mid()] = idx

		rdp.sdpObject.Media[idx] = newMediaSection.MediaObject()
	} else {
		// Get the index of the old media section.
		idx, exist := rdp.midToIndex[newMediaSection.Mid()]
		if !exist {
			panic("no media section found for newMediaSection.Mid()")
		}
		// Replace the media section.
		rdp.mediaSections[idx] = newMediaSection

		rdp.sdpObject.Media[idx] = newMediaSection.MediaObject()
	}

}

func (rdp *RemoteSdp) regenerateBundleMids() {
	if rdp.dtlsParameters == nil {
		return
	}
	mids := lo.Map(lo.Filter(rdp.mediaSections, func(mediaSection *sdp.MediaSection, index int) bool {
		return !mediaSection.Closed()
	}), func(mediaSection *sdp.MediaSection, index int) string {
		return mediaSection.Mid()
	})
	rdp.sdpObject.Groups[0].Mids = strings.Join(mids, " ")
}

func (rdp *RemoteSdp) updateDtlsRole(role mediasoup.DtlsRole) {
	rdp.dtlsParameters.Role = role

	for _, mediaSection := range rdp.mediaSections {
		mediaSection.SetDtlsRole(role)
	}
}
