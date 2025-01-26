package client

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/annidy/mediasoup-client/pkg/sdp"
	"github.com/jiyeyuran/mediasoup-go"
)

func extractRtpCapabilities(sdpObject sdp.Sdp) (rtpCapabilities RtpCapabilities) {
	var gotAudio, gotVideo bool
	codecsMap := make(map[byte]*RtpCodecCapability)
	headerExtensions := make([]*RtpHeaderExtension, 0)
	for _, media := range sdpObject.Media {
		// only consider one audio and video
		if media.Type == "audio" {
			if gotAudio {
				continue
			}
			gotAudio = true
		} else if media.Type == "video" {
			if gotVideo {
				continue
			}
			gotVideo = true
		} else {
			continue
		}
		for _, rtp := range media.Rtp {
			channels, _ := strconv.ParseInt(rtp.Encoding, 10, 64)
			codec := &RtpCodecCapability{
				Kind:                 mediasoup.MediaKind(media.Type),
				MimeType:             fmt.Sprintf("%s/%s", media.Type, rtp.Codec),
				PreferredPayloadType: byte(rtp.Payload),
				ClockRate:            rtp.Rate,
				Channels:             int(channels),
				RtcpFeedback:         []RtcpFeedback{},
			}
			codecsMap[codec.PreferredPayloadType] = codec
		}
		// get codec parameters
		for _, fmtp := range media.Fmtp {
			parameters := sdp.ParseParams(fmtp.Config)
			if codec, ok := codecsMap[byte(fmtp.Payload)]; ok {
				codec.Parameters.ProfileLevelId = parameters.ProfileLevelId
				if parameters.PacketizationMode > 0 {
					var PacketizationMode uint8 = uint8(parameters.PacketizationMode)
					codec.Parameters.PacketizationMode = &PacketizationMode
				}
				codec.Parameters.LevelAsymmetryAllowed = uint8(parameters.LevelAsymmetryAllowed)
				codec.Parameters.Apt = byte(parameters.Apt)
				codec.Parameters.Minptime = uint8(parameters.Minptime)
				codec.Parameters.Useinbandfec = uint8(parameters.Useinbandfec)
				if parameters.ProfileId > 0 {
					var ProfileId uint8 = uint8(parameters.ProfileId)
					codec.Parameters.ProfileId = &ProfileId
				}
				// TODO: 拷贝其他属性
			}
		}
		// get rtcp feedback
		for _, fb := range media.RtcpFb {
			feedback := RtcpFeedback{
				Type:      fb.Type,
				Parameter: fb.Subtype,
			}
			if fb.Payload != "*" {
				payload, err := strconv.ParseInt(fb.Payload, 10, 64)
				if err != nil {
					continue
				}
				if codec, ok := codecsMap[byte(payload)]; ok {
					codec.RtcpFeedback = append(codec.RtcpFeedback, feedback)
				}
			} else {
				panic("not supported")
			}
		}
		// get rtp header extensions
		for _, ext := range media.Ext {
			if len(ext.EncryptUri) > 0 {
				continue
			}
			headerExtension := &RtpHeaderExtension{
				Kind:        mediasoup.MediaKind(media.Type),
				Uri:         ext.Uri,
				PreferredId: ext.Value,
			}
			headerExtensions = append(headerExtensions, headerExtension)
		}
	}
	return RtpCapabilities{
		Codecs:           slices.Collect(maps.Values(codecsMap)),
		HeaderExtensions: headerExtensions,
	}
}

func isRtxCodec(codec RtpCodecCapability) bool {
	return strings.HasSuffix(codec.MimeType, "/rtx")
}

func matchCodec(aCodec, bCodec RtpCodecCapability, strict, modify bool) bool {
	aMinType := strings.ToLower(aCodec.MimeType)
	bMinType := strings.ToLower(bCodec.MimeType)
	if aMinType != bMinType {
		return false
	}
	if aCodec.Kind == mediasoup.MediaKind_Audio {
		if aCodec.ClockRate != bCodec.ClockRate {
			return false
		}
		if aCodec.Channels != bCodec.Channels {
			return false
		}
	}
	switch aMinType {
	case "video/h264":
		if strict {
			// TODO: check profile-level-id
		}
	case "video/vp9":
		if strict {
			// TODO: check profile-id
		}
	}
	return true
}

func reduceRtcpFeedback(aCodec, bCodec RtpCodecCapability) []RtcpFeedback {
	var result []RtcpFeedback
	for _, fb := range bCodec.RtcpFeedback {
		if slices.Contains(aCodec.RtcpFeedback, fb) {
			result = append(result, fb)
		}
	}
	return result
}

func matchHeaderExtension(aExt, bExt RtpHeaderExtension) bool {
	return aExt.Kind == bExt.Kind && aExt.Uri == bExt.Uri
}
