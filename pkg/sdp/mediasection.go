package sdp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/annidy/mediasoup-client/internal/utils"
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type MediaSection struct {
	closed      bool
	mid         string
	mediaObject MediaObject
	planB       bool
}

type MediaSectionOptions struct {
	IceParameters  *mediasoup.IceParameters
	IceCandidates  []*mediasoup.IceCandidate
	DtlsParameters *mediasoup.DtlsParameters
	PlanB          bool
}

func NewMediaSection(options MediaSectionOptions) *MediaSection {
	iceParameters, iceCandidates, dtlsParameters, planB := options.IceParameters, options.IceCandidates, options.DtlsParameters, options.PlanB

	s := MediaSection{
		planB: planB,
	}

	if iceParameters != nil {
		s.SetIceParameters(iceParameters)
	}
	if iceCandidates != nil {
		s.mediaObject.Candidates = make([]MeidaCandidates, len(iceCandidates))

		for i, candidate := range iceCandidates {
			s.mediaObject.Candidates[i] = MeidaCandidates{
				Component:  1,
				Foundation: candidate.Foundation,
				IP:         candidate.Ip,
				Port:       int(candidate.Port),
				Priority:   int(candidate.Priority),
				Transport:  string(candidate.Protocol),
				Type:       candidate.Type,
				Tcptype:    candidate.TcpType,
			}
		}
		s.mediaObject.EndOfCandidates = "end-of-candidates"
		s.mediaObject.IceOptions = "renomination"
	}

	if dtlsParameters != nil {
		s.SetDtlsRole(dtlsParameters.Role)
	}

	return &s
}

func (ms *MediaSection) SetIceParameters(iceParameters *mediasoup.IceParameters) {
	ms.mediaObject.IceUfrag = iceParameters.UsernameFragment
	ms.mediaObject.IcePwd = iceParameters.Password
}

func (ms *MediaSection) SetDtlsRole(role mediasoup.DtlsRole) {
	switch role {
	case "auto":
		ms.mediaObject.Setup = "actpass"
	case "client":
		ms.mediaObject.Setup = "active"
	case "server":
		ms.mediaObject.Setup = "passive"
	}
}

func (ms *MediaSection) Closed() bool {
	return ms.closed
}

func (ms *MediaSection) Mid() string {
	return ms.mid
}

func (ms *MediaSection) MediaObject() MediaObject {
	return ms.mediaObject
}

type AnswerMediaSectionOptions struct {
	MediaSectionOptions
	OfferMediaObject    MediaObject
	AnswerRtpParameters mediasoup.RtpParameters
	CodecOptions        []*mediasoup.RtpCodecParameters
	ExtmapAllowMixed    bool
}

type AnswerMediaSection struct {
	MediaSection
	mediaObject MediaObject
}

func NewAnswerMediaSection(options AnswerMediaSectionOptions) *AnswerMediaSection {
	offerMediaObject, answerRtpParameters, extmapAllowMixed := options.OfferMediaObject, options.AnswerRtpParameters, options.ExtmapAllowMixed
	ms := &AnswerMediaSection{
		MediaSection: *NewMediaSection(options.MediaSectionOptions),
	}

	ms.mediaObject.Mid = offerMediaObject.Mid
	ms.mediaObject.Type = offerMediaObject.Type
	ms.mediaObject.Protocol = offerMediaObject.Protocol

	// No plainRtpParameters
	ms.mediaObject.Connection.IP = "127.0.0.1"
	ms.mediaObject.Connection.Version = 4

	switch offerMediaObject.Type {
	case "audio", "video":
		ms.mediaObject.Direction = "recvonly"

		for _, codec := range answerRtpParameters.Codecs {
			rtp := struct {
				Codec    string `json:"codec,omitempty"`
				Payload  int    `json:"payload,omitempty"`
				Rate     int    `json:"rate,omitempty"`
				Encoding string `json:"encoding,omitempty"`
			}{
				Codec:   codec.MimeType,
				Payload: int(codec.PayloadType),
				Rate:    codec.ClockRate,
			}
			if codec.Channels > 0 {
				// ????
				rtp.Encoding = fmt.Sprintf("%d", codec.Channels)
			}

			ms.mediaObject.Rtp = append(ms.mediaObject.Rtp, rtp)

			var codecParameters mediasoup.RtpCodecSpecificParameters
			utils.Clone(codec.Parameters, &codecParameters)
			var codecRtcpFeedback []mediasoup.RtcpFeedback
			utils.Clone(codec.RtcpFeedback, &codecRtcpFeedback)

			fmtp := struct {
				Config  string `json:"config,omitempty"`
				Payload int    `json:"payload,omitempty"`
			}{
				Payload: int(codec.PayloadType),
			}

			jc, _ := json.Marshal(codecParameters)
			var jcodecParameters map[string]string
			json.Unmarshal(jc, &jcodecParameters)

			for key, value := range jcodecParameters {
				if len(value) == 0 {
					continue
				}
				if len(fmtp.Config) > 0 {
					fmtp.Config += "; "
				}
				fmtp.Config = fmt.Sprintf("%s=%s", key, value)
			}
			if len(fmtp.Config) > 0 {
				ms.mediaObject.Fmtp = append(ms.mediaObject.Fmtp, fmtp)
			}
			for _, fb := range codecRtcpFeedback {
				rtcpFb := struct {
					Payload string `json:"payload,omitempty"`
					Type    string `json:"type,omitempty"`
					Subtype string `json:"subtype,omitempty"`
				}{
					Payload: fmt.Sprintf("%d", codec.PayloadType),
					Type:    fb.Type,
					Subtype: fb.Parameter,
				}
				ms.mediaObject.RtcpFb = append(ms.mediaObject.RtcpFb, rtcpFb)
			}
		}

		ms.mediaObject.Payloads = strings.Join(lo.Map(answerRtpParameters.Codecs, func(codec *mediasoup.RtpCodecParameters, index int) string {
			return fmt.Sprintf("%d", codec.PayloadType)
		}), " ")

		for _, ext := range answerRtpParameters.HeaderExtensions {
			var found bool
			for _, offerExt := range offerMediaObject.Ext {
				if offerExt.Uri == ext.Uri {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			ms.mediaObject.Ext = append(ms.mediaObject.Ext, struct {
				EncryptUri string `json:"encrypt-uri,omitempty"`
				Uri        string `json:"uri,omitempty"`
				Value      int    `json:"value,omitempty"`
			}{
				Value: ext.Id,
				Uri:   ext.Uri,
			})

			if extmapAllowMixed && offerMediaObject.ExtmapAllowMixed == "extmap-allow-mixed" {
				ms.mediaObject.ExtmapAllowMixed = "extmap-allow-mixed"
			}
			// TODO: Simulcast

			ms.mediaObject.RtcpMux = "rtcp-mux"
			ms.mediaObject.RtcpRsize = "rtcp-rsize"
		}
	default:
		log.Warn().Msgf("ignoring media section with unsupported type: %s", offerMediaObject.Type)
	}

	return ms
}
