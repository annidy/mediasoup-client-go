package client

import (
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/pion/webrtc/v4"
	"github.com/rs/zerolog/log"
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
	extendedRtpCapabilities *RtpCapabilitiesEx
	closed                  bool

	handler        *PionHandler
	producers      map[string]*Producer
	producerIdChan chan string
}

type DeviceCreateTransportOptions struct {
	direction               string
	Id                      string          `json:"id,omitempty"`
	IceParameters           *IceParameters  `json:"iceParameters,omitempty"`
	IceCandidates           []*IceCandidate `json:"iceCandidates,omitempty"`
	DtlsParameters          *DtlsParameters `json:"dtlsParameters,omitempty"`
	SctpParameters          *SctpParameters `json:"sctpParameters,omitempty"`
	extendedRtpCapabilities *RtpCapabilitiesEx
}

func newTransport(options DeviceCreateTransportOptions) *Transport {
	transport := &Transport{
		IEventEmitter:           mediasoup.NewEventEmitter(),
		id:                      options.Id,
		direction:               options.direction,
		extendedRtpCapabilities: options.extendedRtpCapabilities,
		handler:                 NewPionHandler(),
		producers:               make(map[string]*Producer),
		producerIdChan:          make(chan string, 1), // 发消息是阻塞的，这里必现带缓冲
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

type TransportProduceOptions struct {
	Track        webrtc.TrackLocal
	codecOptions []*mediasoup.RtpCodecParameters
	Codec        *mediasoup.RtpCodecParameters
	OnRtpSender  func(*webrtc.RTPSender)
	AppData      any
}

func (t *Transport) Produce(options TransportProduceOptions) *Producer {
	track, codecOptions, codec, onRtpSender, appData := options.Track, options.codecOptions, options.Codec, options.OnRtpSender, options.AppData

	localId, rtpParameters, _ := t.handler.send(HandlerSendOptions{
		track:        track,
		codecOptions: codecOptions,
		codec:        codec,
		onRtpSender:  onRtpSender,
	})

	// This will fill rtpParameters's missing fields with default values.
	ortc.validateRtpParameters(rtpParameters)

	if !t.Emit("produce", MediaKind(track.Kind().String()), rtpParameters, appData) {
		log.Error().Msg("produce signaling failed")
		return nil
	}

	id := <-t.producerIdChan

	producer := NewProducer(ProducerOptions{
		Id:            id,
		LocalId:       localId,
		Kind:          mediasoup.MediaKind(track.Kind().String()),
		RtpParameters: rtpParameters,
	})

	t.producers[id] = producer
	t.handleProducer(producer)

	t.SafeEmit("newproducer", producer)

	return producer
}

func (t *Transport) ProduceData(options DataProducerOptions) *DataProducer {
	return nil
}

func (t *Transport) Id() string {
	return t.id
}

func (t *Transport) Closed() bool {
	return t.closed
}

func (t *Transport) Direction() string {
	return t.direction
}

func (t *Transport) RestartIce(iceParameters IceParameters) {

}

func (t *Transport) handleHandler() {
	handler := t.handler
	handler.On("@connect", func(dtlsParameters *DtlsParameters) {
		if !t.closed {
			t.SafeEmit("connect", dtlsParameters)
		}
	})

	handler.On("@iceconnectionstatechange", func(state string) {
		if !t.closed {
			t.SafeEmit("iceconnectionstatechange", state)
		}
	})

	handler.On("@connectionstatechange", func(state string) {
		if !t.closed {
			t.SafeEmit("connectionstatechange", state)
		}
	})
}

func (t *Transport) handleProducer(producer *Producer) {
	producer.On("@close", func() {
		delete(t.producers, producer.Id())

		if t.closed {
			return
		}
	})

	producer.On("@pause", func() {
		// TODO: handler pauseSending
	})

	producer.On("@resume", func() {
		// TODO: handler resumeSending
	})
	producer.On("@replacetrack", func() {
		// TODO: handler replaceTrack
	})
	producer.On("@setmaxspatiallayers", func() {
		// TODO: handler setMaxSpatialLayers
	})
	producer.On("@setrtpencodingparameters", func() {
		// TODO: handler setRtpEncodingParameters
	})

	producer.On("@getstats", func() {
		// TODO: handler getStats
	})
}

func (t *Transport) ProducerIdChan() chan<- string {
	return t.producerIdChan
}
