package client

import "github.com/jiyeyuran/mediasoup-go"

type IceParameters = mediasoup.IceParameters
type DtlsParameters = mediasoup.DtlsParameters
type RtpCapabilities = mediasoup.RtpCapabilities
type RtpCodecCapability = mediasoup.RtpCodecCapability
type MediaKind = mediasoup.MediaKind
type RtpParameters = mediasoup.RtpParameters
type SctpCapabilities = mediasoup.SctpCapabilities
type SctpStreamParameters = mediasoup.SctpStreamParameters
type RtpCodecSpecificParameters = mediasoup.RtpCodecSpecificParameters
type RtcpFeedback = mediasoup.RtcpFeedback

type WebrtcTransportInfo struct {
	Id             string                    `json:"id,omitempty"`
	DtlsParameters *mediasoup.DtlsParameters `json:"dtlsParameters,omitempty"`
	IceCandidates  *mediasoup.IceCandidate   `json:"port,omitempty"`
	SctpParameters *mediasoup.SctpParameters `json:"rtcpPort,omitempty"`
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
