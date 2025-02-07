package client

import (
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/pion/webrtc/v4"
)

type IceParameters = mediasoup.IceParameters
type IceCandidate = mediasoup.IceCandidate
type DtlsParameters = mediasoup.DtlsParameters
type RtpCapabilities = mediasoup.RtpCapabilities
type RtpCodecCapability = mediasoup.RtpCodecCapability
type RtpCodecParameters = mediasoup.RtpCodecParameters
type MediaKind = mediasoup.MediaKind
type RtpParameters = mediasoup.RtpParameters
type SctpCapabilities = mediasoup.SctpCapabilities
type SctpParameters = mediasoup.SctpParameters
type SctpStreamParameters = mediasoup.SctpStreamParameters
type RtpCodecSpecificParameters = mediasoup.RtpCodecSpecificParameters
type RtcpFeedback = mediasoup.RtcpFeedback
type RtpHeaderExtension = mediasoup.RtpHeaderExtension

type RtpCodec struct {
	c *RtpCodecCapability
	p *RtpCodecParameters
}

func (r RtpCodec) MimeType() string {
	if r.c != nil {
		return r.c.MimeType
	}
	return r.p.MimeType
}

func (r RtpCodec) Kind() mediasoup.MediaKind {
	if r.c != nil {
		return r.c.Kind
	}
	panic("type error")
}

func (r RtpCodec) HasKind() bool {
	return r.c != nil
}

func (r RtpCodec) ClockRate() int {
	if r.c != nil {
		return r.c.ClockRate
	}
	return r.p.ClockRate
}

func (r RtpCodec) Channels() int {
	if r.c != nil {
		return r.c.Channels
	}
	return r.p.Channels
}

func (r RtpCodec) Parameters() RtpCodecSpecificParameters {
	if r.c != nil {
		return r.c.Parameters
	}
	return r.p.Parameters
}

type RTCRtpSender interface {
}

type RtpCapabilitiesEx struct {
	Codecs           []*RtpCodecCapabilityEx `json:"codecs,omitempty"`
	HeaderExtensions []*RtpHeaderExtensionEx `json:"headerExtensions,omitempty"`
}
type RtpCodecCapabilityEx struct {
	RtpCodecCapability
	LocalPayloadType     byte
	LocalRtxPayloadType  byte
	LocalParameters      RtpCodecSpecificParameters
	RemotePayloadType    byte
	RemoteRtxPayloadType byte
	RemoteParameters     RtpCodecSpecificParameters
}

type RtpHeaderExtensionEx struct {
	RtpHeaderExtension
	SendId  int
	RecvId  int
	Encrypt bool
}

type WebrtcTransportInfo struct {
	Id               string                      `json:"id,omitempty"`
	DtlsParameters   *mediasoup.DtlsParameters   `json:"dtlsParameters,omitempty"`
	SctpCapabilities *mediasoup.SctpCapabilities `json:"sctpCapabilities,omitempty"`
}

type DataProducerOptions struct {
	Ordered        bool
	MaxRetransmits int
	Label          string
	Priority       string
	AppData        any
}

type TrackProducerOptions struct {
	Track *webrtc.TrackLocal
}

type DeviceInfo struct {
	Flag    string `json:"flag,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}
