package sdp

import (
	"encoding/json"

	"github.com/Jeffail/gabs"
	"github.com/jiyeyuran/mediasoup-go"
)

func GetCname(offerMediaObject interface{}) string {
	j, _ := json.Marshal(offerMediaObject)
	jsonMedia, _ := gabs.ParseJSON(j)
	if children, err := jsonMedia.S("ssrcs").Children(); err == nil {
		for _, ssrc := range children {
			if ssrc.Path("attribute").Data().(string) == "cname" {
				return ssrc.Path("value").Data().(string)
			}
		}
	}
	return ""
}

func GetRtpEncodings(offerMediaObject interface{}) []mediasoup.RtpEncodingParameters {
	j, _ := json.Marshal(offerMediaObject)
	jsonMedia, _ := gabs.ParseJSON(j)

	ssrcs := make(map[uint32]bool, 0)
	if children, err := jsonMedia.S("ssrcs").Children(); err == nil {
		for _, ssrc := range children {
			var id float64 = ssrc.Path("id").Data().(float64)
			ssrcs[uint32(id)] = true
		}
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
