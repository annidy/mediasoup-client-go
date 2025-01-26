package client

import "github.com/jiyeyuran/mediasoup-go"

type IceParameters = mediasoup.IceParameters
type IceCandidate = mediasoup.IceCandidate
type DtlsParameters = mediasoup.DtlsParameters
type RtpCapabilities = mediasoup.RtpCapabilities
type RtpCodecCapability = mediasoup.RtpCodecCapability
type MediaKind = mediasoup.MediaKind
type RtpParameters = mediasoup.RtpParameters
type SctpCapabilities = mediasoup.SctpCapabilities
type SctpParameters = mediasoup.SctpParameters
type SctpStreamParameters = mediasoup.SctpStreamParameters
type RtpCodecSpecificParameters = mediasoup.RtpCodecSpecificParameters
type RtcpFeedback = mediasoup.RtcpFeedback
type RtpHeaderExtension = mediasoup.RtpHeaderExtension

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

type DeviceInfo struct {
	Flag    string `json:"flag,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}
