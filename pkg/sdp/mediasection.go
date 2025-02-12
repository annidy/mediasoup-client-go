package sdp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/annidy/mediasoup-client/internal/gptr"
	"github.com/annidy/mediasoup-client/internal/utils"
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type MediaSection struct {
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
	return ms.mediaObject.Port == 0
}

func (ms *MediaSection) Mid() string {
	return ms.mediaObject.Mid
}

func (ms *MediaSection) MediaObject() MediaObject {
	return ms.mediaObject
}

type ProducerCodecOptions struct {
	OpusStereo              *bool
	OpusFec                 *bool
	OpusDtx                 *bool
	OpusMaxPlaybackRate     *int
	OpusMaxAverageBitrate   *int
	OpusPtime               *int
	OpusNack                *bool
	VideoGoogleStartBitrate *int
	VideoGoogleMaxBitrate   *int
	VideoGoogleMinBitrate   *int
}

type AnswerMediaSectionOptions struct {
	MediaSectionOptions
	OfferMediaObject    MediaObject
	AnswerRtpParameters mediasoup.RtpParameters
	OfferRtpParameters  mediasoup.RtpParameters
	CodecOptions        ProducerCodecOptions
	ExtmapAllowMixed    bool
}

type AnswerMediaSection struct {
	MediaSection
}

func NewAnswerMediaSection(options AnswerMediaSectionOptions) *AnswerMediaSection {
	offerMediaObject, answerRtpParameters, extmapAllowMixed, codecOptions := options.OfferMediaObject, options.AnswerRtpParameters, options.ExtmapAllowMixed, options.CodecOptions
	ms := &AnswerMediaSection{
		MediaSection: *NewMediaSection(options.MediaSectionOptions),
	}

	ms.mediaObject.Mid = offerMediaObject.Mid
	ms.mediaObject.Type = offerMediaObject.Type
	ms.mediaObject.Protocol = offerMediaObject.Protocol

	// No plainRtpParameters
	ms.mediaObject.Connection = gptr.Of(struct {
		IP      string `json:"ip"`
		Version int    `json:"version"`
	}{
		IP:      "127.0.0.1",
		Version: 4,
	})
	ms.mediaObject.Port = 7

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
				Codec:   getCodecName(codec),
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

			opusStereo, opusFec, opusDtx, opusMaxPlaybackRate, opusMaxAverageBitrate, opusPtime, opusNack, videoGoogleStartBitrate, videoGoogleMaxBitrate, videoGoogleMinBitrate := codecOptions.OpusStereo, codecOptions.OpusFec, codecOptions.OpusDtx, codecOptions.OpusMaxPlaybackRate, codecOptions.OpusMaxAverageBitrate, codecOptions.OpusPtime, codecOptions.OpusNack, codecOptions.VideoGoogleStartBitrate, codecOptions.VideoGoogleMaxBitrate, codecOptions.VideoGoogleMinBitrate

			switch strings.ToLower(codec.MimeType) {
			case "audio/opus", "audio/multiopus":
				if opusStereo != nil {
					codecParameters.SpropStereo = utils.Bool2Type[uint8](*opusStereo)
				}
				if opusFec != nil {
					codecParameters.Useinbandfec = utils.Bool2Type[uint8](*opusFec)
				}
				if opusDtx != nil {
					codecParameters.Usedtx = utils.Bool2Type[uint8](*opusDtx)
				}
				if opusMaxPlaybackRate != nil {
					codecParameters.Maxplaybackrate = uint32(*opusMaxPlaybackRate)
				}
				if opusMaxAverageBitrate != nil {
					// TODO: no maxaveragebitrate in codecParameters
				}
				if opusPtime != nil {
					// TODO: no ptime in codecParameters
				}

				// If opusNack is not set, we must remove NACK support for OPUS.
				// Otherwise it would be enabled for those handlers that artificially
				// announce it in their RTP capabilities.
				if opusNack != nil {
					codecRtcpFeedback = lo.Filter(codecRtcpFeedback, func(fb mediasoup.RtcpFeedback, index int) bool {
						return fb.Type != "nack" || fb.Parameter != ""
					})
				}

			case "video/vp8", "video/vp9", "video/h264", "video/h265":
				if videoGoogleStartBitrate != nil {
					codecParameters.XGoogleStartBitrate = uint32(*videoGoogleStartBitrate)
				}
				if videoGoogleMaxBitrate != nil {
					codecParameters.XGoogleMaxBitrate = uint32(*videoGoogleMaxBitrate)
				}
				if videoGoogleMinBitrate != nil {
					codecParameters.XGoogleMinBitrate = uint32(*videoGoogleMinBitrate)
				}
			}

			fmtp := struct {
				Config  string `json:"config,omitempty"`
				Payload int    `json:"payload,omitempty"`
			}{
				Payload: int(codec.PayloadType),
			}

			jc, _ := json.Marshal(codecParameters)
			jsonParsed, _ := gabs.ParseJSON(jc)
			contaner, _ := jsonParsed.ChildrenMap()
			for key, value := range contaner {
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
		}
	default:
		log.Warn().Msgf("ignoring media section with unsupported type: %s", offerMediaObject.Type)
	}

	return ms
}

func getCodecName(codec *mediasoup.RtpCodecParameters) string {
	mineTypeRegex := regexp.MustCompile(`(audio|video)/(.+)`)
	matches := mineTypeRegex.FindStringSubmatch(codec.MimeType)
	if len(matches) < 3 {
		return ""
	}
	return matches[2]
}
