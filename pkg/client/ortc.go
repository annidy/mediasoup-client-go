package client

import (
	"reflect"
	"strings"

	"github.com/annidy/mediasoup-client/internal/util"
	"github.com/jiyeyuran/mediasoup-go"
)

type Ortc struct{}

var ortc Ortc

var DYNAMIC_PAYLOAD_TYPES = [...]byte{
	100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115,
	116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 96, 97, 98, 99, 77,
	78, 79, 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 35, 36,
	37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56,
	57, 58, 59, 60, 61, 62, 63, 64, 65, 66, 67, 68, 69, 70, 71,
}

type matchOptions struct {
	strict bool
	modify bool
}

type RtpMapping struct {
	Codecs    []RtpMappingCodec    `json:"codecs,omitempty"`
	Encodings []RtpMappingEncoding `json:"encodings,omitempty"`
}

type RtpMappingCodec struct {
	PayloadType       byte `json:"payloadType"`
	MappedPayloadType byte `json:"mappedPayloadType"`
}

type RtpMappingEncoding struct {
	Ssrc            uint32 `json:"ssrc,omitempty"`
	Rid             string `json:"rid,omitempty"`
	ScalabilityMode string `json:"scalabilityMode,omitempty"`
	MappedSsrc      uint32 `json:"mappedSsrc"`
}

// validateRtpCapabilities validates RtpCapabilities. It may modify given data by adding missing
// fields with default values.
func (o Ortc) validateRtpCapabilities(params *mediasoup.RtpCapabilities) (err error) {
	if params.Codecs == nil {
		params.Codecs = make([]*mediasoup.RtpCodecCapability, 0)
	}
	for _, codec := range params.Codecs {
		if err = o.validateRtpCodecCapability(codec); err != nil {
			return
		}
	}
	if params.HeaderExtensions == nil {
		params.HeaderExtensions = make([]*mediasoup.RtpHeaderExtension, 0)
	}
	for _, ext := range params.HeaderExtensions {
		if err = o.validateRtpHeaderExtension(ext); err != nil {
			return
		}
	}

	return
}

// validateRtpCodecCapability validates RtpCodecCapability. It may modify given data by adding
// missing fields with default values.
func (o Ortc) validateRtpCodecCapability(code *mediasoup.RtpCodecCapability) (err error) {
	mimeType := strings.ToLower(code.MimeType)

	//  mimeType is mandatory.
	if !strings.HasPrefix(mimeType, "audio/") && !strings.HasPrefix(mimeType, "video/") {
		return mediasoup.NewTypeError("invalid codec.mimeType")
	}

	code.Kind = mediasoup.MediaKind(strings.Split(mimeType, "/")[0])

	// clockRate is mandatory.
	if code.ClockRate == 0 {
		return mediasoup.NewTypeError("missing codec.clockRate")
	}

	// channels is optional. If unset, set it to 1 (just if audio).
	if code.Kind == mediasoup.MediaKind_Audio && code.Channels == 0 {
		code.Channels = 1
	}

	for _, fb := range code.RtcpFeedback {
		if err = o.validateRtcpFeedback(fb); err != nil {
			return
		}
	}

	return
}

/**
 * Generate extended RTP capabilities for sending and receiving.
 */
func (o Ortc) getExtendedRtpCapabilities(localCaps, remoteCaps mediasoup.RtpCapabilities) mediasoup.RtpCapabilities {
	var extendedRtpCapabilities mediasoup.RtpCapabilities

	// TODO: 融合本地和远端的RtpCapabilities
	util.Clone(remoteCaps, &extendedRtpCapabilities)

	return extendedRtpCapabilities
}

func (o Ortc) getRecvRtpCapabilities(extendedRtpCapabilities RtpCapabilities) RtpCapabilities {

	return extendedRtpCapabilities
}

// validateRtcpFeedback validates RtcpFeedback. It may modify given data by adding missing
// fields with default values.
func (o Ortc) validateRtcpFeedback(fb mediasoup.RtcpFeedback) error {
	if len(fb.Type) == 0 {
		return mediasoup.NewTypeError("missing fb.type")
	}
	return nil
}

// validateRtpHeaderExtension validates RtpHeaderExtension. It may modify given data by adding
// missing fields with default values.
func (o Ortc) validateRtpHeaderExtension(ext *mediasoup.RtpHeaderExtension) (err error) {
	if len(ext.Kind) > 0 && ext.Kind != mediasoup.MediaKind_Audio && ext.Kind != mediasoup.MediaKind_Video {
		return mediasoup.NewTypeError("invalid ext.kind")
	}

	// uri is mandatory.
	if len(ext.Uri) == 0 {
		return mediasoup.NewTypeError("missing ext.uri")
	}

	// preferredId is mandatory.
	if ext.PreferredId == 0 {
		return mediasoup.NewTypeError("missing ext.preferredId")
	}

	// direction is optional. If unset set it to sendrecv.
	if len(ext.Direction) == 0 {
		ext.Direction = mediasoup.Direction_Sendrecv
	}

	return
}

// validateRtpParameters validates RtpParameters. It may modify given data by adding missing
// fields with default values.
func (o Ortc) validateRtpParameters(params *mediasoup.RtpParameters) (err error) {
	for _, codec := range params.Codecs {
		if err = o.validateRtpCodecParameters(codec); err != nil {
			return
		}
	}

	for _, ext := range params.HeaderExtensions {
		if err = o.validateRtpHeaderExtensionParameters(ext); err != nil {
			return
		}
	}

	return o.validateRtcpParameters(&params.Rtcp)
}

// validateRtpCodecParameters validates RtpCodecParameters. It may modify given data by adding
// missing fields with default values.
func (o Ortc) validateRtpCodecParameters(code *mediasoup.RtpCodecParameters) (err error) {
	mimeType := strings.ToLower(code.MimeType)

	//  mimeType is mandatory.
	if !strings.HasPrefix(mimeType, "audio/") && !strings.HasPrefix(mimeType, "video/") {
		return mediasoup.NewTypeError("invalid codec.mimeType")
	}

	// clockRate is mandatory.
	if code.ClockRate == 0 {
		return mediasoup.NewTypeError("missing codec.clockRate")
	}

	kind := mediasoup.MediaKind(strings.Split(mimeType, "/")[0])

	// channels is optional. If unset, set it to 1 (just if audio).
	if kind == mediasoup.MediaKind_Audio && code.Channels == 0 {
		code.Channels = 1
	}

	for _, fb := range code.RtcpFeedback {
		if err = o.validateRtcpFeedback(fb); err != nil {
			return
		}
	}

	return
}

// validateRtpHeaderExtensionParameters validates RtpHeaderExtension. It may modify given data by
// adding missing fields with default values.
func (o Ortc) validateRtpHeaderExtensionParameters(ext mediasoup.RtpHeaderExtensionParameters) (err error) {
	// uri is mandatory.
	if len(ext.Uri) == 0 {
		return mediasoup.NewTypeError("missing ext.uri")
	}

	// preferredId is mandatory.
	if ext.Id == 0 {
		return mediasoup.NewTypeError("missing ext.id")
	}

	return
}

// validateRtcpParameters validates RtcpParameters. It may modify given data by adding missing
// fields with default values.
func (o Ortc) validateRtcpParameters(rtcp *mediasoup.RtcpParameters) (err error) {
	// reducedSize is optional. If unset set it to true.
	if rtcp.ReducedSize == nil {
		rtcp.ReducedSize = mediasoup.Bool(true)
	}

	return
}

// validateSctpCapabilities validates SctpCapabilities. It may modify given data by adding missing
// fields with default values.
func (o Ortc) validateSctpCapabilities(caps mediasoup.SctpCapabilities) (err error) {
	// numStreams is mandatory.
	if reflect.DeepEqual(caps.NumStreams, mediasoup.NumSctpStreams{}) {
		return mediasoup.NewTypeError("missing caps.numStreams")
	}

	return o.validateNumSctpStreams(caps.NumStreams)
}

// validateNumSctpStreams validates NumSctpStreams. It may modify given data by adding missing
// fields with default values.
func (o Ortc) validateNumSctpStreams(numStreams mediasoup.NumSctpStreams) (err error) {
	// OS is mandatory.
	if numStreams.OS == 0 {
		return mediasoup.NewTypeError("missing numStreams.OS")
	}
	// MIS is mandatory.
	if numStreams.MIS == 0 {
		return mediasoup.NewTypeError("missing numStreams.MIS")
	}

	return
}

// validateSctpParameters validates SctpParameters. It may modify given data by adding missing
// fields with default values.
func (o Ortc) validateSctpParameters(params mediasoup.SctpParameters) (err error) {
	// port is mandatory.
	if params.Port == 0 {
		return mediasoup.NewTypeError("missing params.port")
	}

	// OS is mandatory.
	if params.OS == 0 {
		return mediasoup.NewTypeError("missing params.OS")
	}
	// MIS is mandatory.
	if params.MIS == 0 {
		return mediasoup.NewTypeError("missing params.MIS")
	}

	// maxMessageSize is mandatory.
	if params.MaxMessageSize == 0 {
		return mediasoup.NewTypeError("missing params.maxMessageSize")
	}

	return
}

// validateSctpStreamParameters validates SctpStreamParameters. It may modify given data by adding
// missing fields with default values.
func (o Ortc) validateSctpStreamParameters(params *mediasoup.SctpStreamParameters) (err error) {
	if params == nil {
		return mediasoup.NewTypeError("params is nil")
	}
	orderedGiven := params.Ordered != nil

	if params.Ordered == nil {
		params.Ordered = mediasoup.Bool(true)
	}

	if params.MaxPacketLifeTime > 0 && params.MaxRetransmits > 0 {
		return mediasoup.NewTypeError("cannot provide both maxPacketLifeTime and maxRetransmits")
	}

	if orderedGiven && *params.Ordered &&
		(params.MaxPacketLifeTime > 0 || params.MaxRetransmits > 0) {
		return mediasoup.NewTypeError("cannot be ordered with maxPacketLifeTime or maxRetransmits")
	} else if !orderedGiven && (params.MaxPacketLifeTime > 0 || params.MaxRetransmits > 0) {
		params.Ordered = mediasoup.Bool(false)
	}

	return
}

func (o Ortc) addNackSupportForOpus(caps *RtpCapabilities) {
	for _, codec := range caps.Codecs {
		if strings.ToLower(codec.MimeType) == "audio/opus" {
			for _, fb := range codec.RtcpFeedback {
				if fb.Type == "nack" {
					return
				}
			}
			codec.RtcpFeedback = append(codec.RtcpFeedback, mediasoup.RtcpFeedback{Type: "nack"})
		}
	}
}
