package client

import (
	"fmt"
	"maps"
	"slices"
	"strconv"

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
			codec := &RtpCodecCapability{
				Kind:                 mediasoup.MediaKind(media.Type),
				MimeType:             fmt.Sprintf("%s/%s", media.Type, rtp.Codec),
				PreferredPayloadType: byte(rtp.Payload),
				ClockRate:            rtp.Rate,
				Channels:             1,
				RtcpFeedback:         []RtcpFeedback{},
			}
			codecsMap[codec.PreferredPayloadType] = codec
		}
		// get profile-level-id
		for _, fmtp := range media.Fmtp {
			parameters := sdp.ParseParams(fmtp.Config)
			if codec, ok := codecsMap[byte(fmtp.Payload)]; ok {
				codec.Parameters.ProfileLevelId = parameters.ProfileLevelId
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
