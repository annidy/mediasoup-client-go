package sdp

import (
	"github.com/jiyeyuran/mediasoup-go"
)

func GetCname(offerMediaObject MediaObject) string {
	for _, ssrc := range offerMediaObject.Ssrcs {
		if ssrc.Attribute == "cname" {
			return ssrc.Value
		}
	}
	return ""
}

func GetRtpEncodings(offerMediaObject MediaObject) []mediasoup.RtpEncodingParameters {

	ssrcs := make(map[uint32]bool, 0)
	for _, ssrc := range offerMediaObject.Ssrcs {
		ssrcs[uint32(ssrc.ID)] = true
	}

	if len(ssrcs) == 0 {
		panic("no a=ssrc line found")
	}

	ssrcToRtxSsrc := make(map[uint32]uint32, 0)

	// TODO: 处理rtx。但是mediasObject没有ssrcGroups，所以暂时不处理

	for ssrc := range ssrcs {
		ssrcToRtxSsrc[ssrc] = 0
	}

	encodings := make([]mediasoup.RtpEncodingParameters, 0)
	for ssrc, rtxSsrc := range ssrcToRtxSsrc {
		encoding := mediasoup.RtpEncodingParameters{
			Ssrc: ssrc,
		}
		if rtxSsrc != 0 {
			encoding.Rtx = &mediasoup.RtpEncodingRtx{
				Ssrc: rtxSsrc,
			}
		}
		encodings = append(encodings, encoding)
	}
	return encodings
}
