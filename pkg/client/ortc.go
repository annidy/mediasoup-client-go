package client

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

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
func (o Ortc) getExtendedRtpCapabilities(localCaps, remoteCaps RtpCapabilities) RtpCapabilitiesEx {
	extendedRtpCapabilities := RtpCapabilitiesEx{
		Codecs:           make([]*RtpCodecCapabilityEx, 0),
		HeaderExtensions: make([]*RtpHeaderExtensionEx, 0),
	}
	// Match media codecs and keep the order preferred by remoteCaps.
	usedMap := make(map[*mediasoup.RtpCodecCapability]bool)
	for _, remoteCodec := range remoteCaps.Codecs {
		if isRtxCodec(*remoteCodec) {
			continue
		}
		li := slices.IndexFunc(localCaps.Codecs, func(localCodec *mediasoup.RtpCodecCapability) bool {
			if matchCodec(*localCodec, *remoteCodec, false, false) {
				// 剔除重复的codec
				if _, ok := usedMap[localCodec]; ok {
					return false
				}
				usedMap[localCodec] = true
				return true
			}
			return false
		})
		if li < 0 {
			continue
		}
		matchingLocalCodec := localCaps.Codecs[li]
		extenedCodec := RtpCodecCapabilityEx{
			RtpCodecCapability: mediasoup.RtpCodecCapability{
				MimeType:     matchingLocalCodec.MimeType,
				Kind:         matchingLocalCodec.Kind,
				ClockRate:    matchingLocalCodec.ClockRate,
				Channels:     matchingLocalCodec.Channels,
				RtcpFeedback: reduceRtcpFeedback(*matchingLocalCodec, *remoteCodec),
			},
			LocalPayloadType:  matchingLocalCodec.PreferredPayloadType,
			RemotePayloadType: remoteCodec.PreferredPayloadType,
			LocalParameters:   matchingLocalCodec.Parameters,
			RemoteParameters:  remoteCodec.Parameters,
		}

		extendedRtpCapabilities.Codecs = append(extendedRtpCapabilities.Codecs, &extenedCodec)
	}

	// Match RTX codecs.
	for _, extendCodec := range extendedRtpCapabilities.Codecs {
		li := slices.IndexFunc(localCaps.Codecs, func(localCodec *mediasoup.RtpCodecCapability) bool {
			return isRtxCodec(*localCodec) && localCodec.Parameters.Apt == extendCodec.LocalPayloadType
		})
		ri := slices.IndexFunc(remoteCaps.Codecs, func(remoteCodec *mediasoup.RtpCodecCapability) bool {
			return isRtxCodec(*remoteCodec) && remoteCodec.Parameters.Apt == extendCodec.RemotePayloadType
		})
		if li > 0 && ri > 0 {
			extendCodec.LocalRtxPayloadType = localCaps.Codecs[li].PreferredPayloadType
			extendCodec.RemoteRtxPayloadType = remoteCaps.Codecs[ri].PreferredPayloadType
		}
	}

	// Match header extensions.
	for _, remoteExt := range remoteCaps.HeaderExtensions {
		li := slices.IndexFunc(localCaps.HeaderExtensions, func(localExt *mediasoup.RtpHeaderExtension) bool {
			return matchHeaderExtension(*localExt, *remoteExt)
		})
		if li < 0 {
			continue
		}
		extendedExt := &RtpHeaderExtensionEx{
			RtpHeaderExtension: RtpHeaderExtension{
				Kind:             remoteExt.Kind,
				Uri:              remoteExt.Uri,
				PreferredId:      remoteExt.PreferredId,
				PreferredEncrypt: remoteExt.PreferredEncrypt,
				Direction:        remoteExt.Direction,
			},
			SendId:  localCaps.HeaderExtensions[li].PreferredId,
			RecvId:  remoteExt.PreferredId,
			Encrypt: localCaps.HeaderExtensions[li].PreferredEncrypt,
		}
		// sendonly/recvonly 反向
		if extendedExt.Direction == mediasoup.Direction_Sendonly {
			extendedExt.Direction = mediasoup.Direction_Recvonly
		} else if extendedExt.Direction == mediasoup.Direction_Recvonly {
			extendedExt.Direction = mediasoup.Direction_Sendonly
		}

		extendedRtpCapabilities.HeaderExtensions = append(extendedRtpCapabilities.HeaderExtensions, extendedExt)
	}

	return extendedRtpCapabilities
}

func (o Ortc) getRecvRtpCapabilities(extendedRtpCapabilities RtpCapabilitiesEx) RtpCapabilities {
	rtpCapabilities := RtpCapabilities{
		Codecs:           make([]*RtpCodecCapability, 0),
		HeaderExtensions: make([]*RtpHeaderExtension, 0),
	}
	for _, extendedCodec := range extendedRtpCapabilities.Codecs {
		codec := RtpCodecCapability{
			MimeType:             extendedCodec.MimeType,
			Kind:                 extendedCodec.Kind,
			PreferredPayloadType: extendedCodec.RemotePayloadType,
			ClockRate:            extendedCodec.ClockRate,
			Channels:             extendedCodec.Channels,
			Parameters:           extendedCodec.LocalParameters,
			RtcpFeedback:         extendedCodec.RtcpFeedback,
		}
		rtpCapabilities.Codecs = append(rtpCapabilities.Codecs, &codec)
		if extendedCodec.RemoteRtxPayloadType > 0 {
			rtpCapabilities.Codecs = append(rtpCapabilities.Codecs, &RtpCodecCapability{
				MimeType:             fmt.Sprintf("%s/rtx", extendedCodec.Kind),
				Kind:                 extendedCodec.Kind,
				PreferredPayloadType: extendedCodec.RemoteRtxPayloadType,
				ClockRate:            extendedCodec.ClockRate,
				Channels:             extendedCodec.Channels,
				Parameters:           RtpCodecSpecificParameters{Apt: extendedCodec.RemotePayloadType},
				RtcpFeedback:         make([]mediasoup.RtcpFeedback, 0),
			})
		}
	}

	for _, extendedExt := range extendedRtpCapabilities.HeaderExtensions {
		// Ignore RTP extensions not valid for receiving.
		if extendedExt.Direction != mediasoup.Direction_Sendrecv && extendedExt.Direction != mediasoup.Direction_Recvonly {
			continue
		}
		ext := RtpHeaderExtension{
			Kind:             extendedExt.Kind,
			Uri:              extendedExt.Uri,
			PreferredId:      extendedExt.RecvId,
			PreferredEncrypt: extendedExt.Encrypt,
			Direction:        extendedExt.Direction,
		}
		rtpCapabilities.HeaderExtensions = append(rtpCapabilities.HeaderExtensions, &ext)
	}
	return rtpCapabilities
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

	// TODO: validate encodings

	return o.validateRtcpParameters(&params.Rtcp)
}

// validateRtpCodecParameters validates RtpCodecParameters. It may modify given data by adding
// missing fields with default values.
func (o Ortc) validateRtpCodecParameters(code *mediasoup.RtpCodecParameters) (err error) {
	mimeType := strings.ToLower(code.MimeType)

	//  mimeType is mandatory.
	if !strings.HasPrefix(mimeType, "audio/") && !strings.HasPrefix(mimeType, "video/") {
		return mediasoup.NewTypeError("invalid remoteCodec.mimeType")
	}

	// clockRate is mandatory.
	if code.ClockRate == 0 {
		return mediasoup.NewTypeError("missing remoteCodec.clockRate")
	}

	// channels is optional. If unset, set it to 1 (just if audio).
	if strings.HasPrefix(mimeType, "audio/") && code.Channels == 0 {
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
	for _, remoteCodec := range caps.Codecs {
		if strings.ToLower(remoteCodec.MimeType) == "audio/opus" {
			for _, fb := range remoteCodec.RtcpFeedback {
				if fb.Type == "nack" {
					return
				}
			}
			remoteCodec.RtcpFeedback = append(remoteCodec.RtcpFeedback, mediasoup.RtcpFeedback{Type: "nack"})
		}
	}
}
