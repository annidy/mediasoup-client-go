package proto

import (
	"github.com/annidy/mediasoup-client/pkg/client"
	"github.com/jiyeyuran/mediasoup-go"
)

type GetRouterRtpCapabilitiesResponse struct {
	Codecs           []*mediasoup.RtpCodecCapability `json:"codecs,omitempty"`
	HeaderExtensions []*mediasoup.RtpHeaderExtension `json:"headerExtensions,omitempty"`
}

type WebrtcTransportProducerRequest struct {
	TransportId   string                   `json:"transportId,omitempty"`
	Kind          mediasoup.MediaKind      `json:"kind,omitempty"`
	RtpParameters *mediasoup.RtpParameters `json:"rtpParameters,omitempty"`
	AppData       interface{}              `json:"appData,omitempty"`
}

type WebrtcTransportProducerDataRequest struct {
	TransportId          string `json:"transportId,omitempty"`
	SctpStreamParameters *mediasoup.SctpStreamParameters
	Label                string `json:"label,omitempty"`
	Protocol             string `json:"protocol,omitempty"`
	AppData              any    `json:"appData,omitempty"`
}

type CreateWebrtcTransportRequest struct {
	ForceTcp         bool                        `json:"forceTcp,omitempty"`
	Producing        bool                        `json:"producing,omitempty"`
	Consuming        bool                        `json:"consuming,omitempty"`
	SctpCapabilities *mediasoup.SctpCapabilities `json:"sctpCapabilities,omitempty"`
}

type ConnectWebRtcTransportRequest struct {
	TransportId    string                    `json:"transportId,omitempty"`
	DtlsParameters *mediasoup.DtlsParameters `json:"dtlsParameters,omitempty"`
}

// 加入房间请求的参数就是PeerInfo
type JoinRequest struct {
	DisplayName      string                      `json:"displayName,omitempty"`
	Device           client.DeviceInfo           `json:"device,omitempty"`
	RtpCapabilities  *mediasoup.RtpCapabilities  `json:"rtpCapabilities,omitempty"`
	SctpCapabilities *mediasoup.SctpCapabilities `json:"sctpCapabilities,omitempty"`
}

type CloseProducerRequest struct {
	ProducerId string `json:"producerId,omitempty"`
}

type PauseProducerRequest struct {
	ProducerId string `json:"producerId,omitempty"`
}

type ResumeProducerRequest struct {
	ProducerId string `json:"producerId,omitempty"`
}

type RestartIceRequest struct {
	TransportId string `json:"transportId,omitempty"`
}
