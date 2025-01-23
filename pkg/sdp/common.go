package sdp

import (
	"fmt"

	"github.com/annidy/mediasoup-client/pkg/client"
	"github.com/jiyeyuran/mediasoup-go"
)

func ExtractRtpCapabilities(sdpObject Spd) (rtpCapabilities client.RtpCapabilities) {
	var gotAudio, gotVideo bool
	codecsMap := make(map[byte]*client.RtpCodecCapability)
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
			codec := &client.RtpCodecCapability{
				Kind:                 mediasoup.MediaKind(media.Type),
				MimeType:             fmt.Sprintf("%s/%s", media.Type, rtp.Codec),
				PreferredPayloadType: byte(rtp.Payload),
				ClockRate:            rtp.Rate,
				Channels:             1,
				RtcpFeedback:         []client.RtcpFeedback{},
			}
			codecsMap[codec.PreferredPayloadType] = codec
		}
		for _, fmtp := range media.Fmtp {
		}
	}
}
