package client

import (
	"github.com/jiyeyuran/mediasoup-go"
)

type Transport struct {
	mediasoup.IEventEmitter
	id                      string
	direction               string
	IceParameters           *IceParameters
	IceCandidates           *[]IceCandidate
	DtlsParameters          *DtlsParameters
	SctpParameters          *SctpParameters
	IceServers              []string
	extendedRtpCapabilities RtpCapabilitiesEx

	handler *PionHandler
}

type TransportOptions struct {
	direction               string
	Id                      string          `json:"id,omitempty"`
	IceParameters           *IceParameters  `json:"iceParameters,omitempty"`
	IceCandidates           []*IceCandidate `json:"iceCandidates,omitempty"`
	DtlsParameters          *DtlsParameters `json:"dtlsParameters,omitempty"`
	SctpParameters          *SctpParameters `json:"sctpParameters,omitempty"`
	extendedRtpCapabilities RtpCapabilitiesEx
}

func NewTransport(options TransportOptions) *Transport {
	transport := &Transport{
		IEventEmitter:           mediasoup.NewEventEmitter(),
		id:                      options.Id,
		direction:               options.direction,
		extendedRtpCapabilities: options.extendedRtpCapabilities,
		handler:                 NewPionHandler(),
	}

	transport.handler.run(HandlerRunOptions{
		direction:              options.direction,
		iceParameters:          options.IceParameters,
		iceCandidates:          options.IceCandidates,
		dtlsParameters:         options.DtlsParameters,
		sctpParameters:         options.SctpParameters,
		extenedRtpCapabilities: options.extendedRtpCapabilities,
	})

	transport.handleHandler()

	return transport
}

func (t *Transport) Produce() *Producer {
	return nil
}

func (t *Transport) ProduceData(options DataProducerOptions) *DataProducer {
	return nil
}

func (t *Transport) Id() string {
	return ""
}

func (t *Transport) RestartIce(iceParameters IceParameters) {

}

func (t *Transport) handleHandler() {

}
