package client

import (
	"github.com/annidy/mediasoup-client/pkg/sdp"
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
	Direction               string
	Id                      string          `json:"id,omitempty"`
	IceParameters           *IceParameters  `json:"iceParameters,omitempty"`
	IceCandidates           []*IceCandidate `json:"iceCandidates,omitempty"`
	DtlsParameters          *DtlsParameters `json:"dtlsParameters,omitempty"`
	SctpParameters          *SctpParameters `json:"sctpParameters,omitempty"`
	extendedRtpCapabilities *RtpCapabilitiesEx
}

func newTransport(options DeviceCreateTransportOptions) *Transport {
	direction, id, iceParameters, iceCandidates, dtlsParameters, sctpParameters, extendedRtpCapabilities := options.Direction, options.Id, options.IceParameters, options.IceCandidates, options.DtlsParameters, options.SctpParameters, options.extendedRtpCapabilities
	transport := &Transport{
		IEventEmitter:           mediasoup.NewEventEmitter(),
		id:                      id,
		direction:               direction,
		extendedRtpCapabilities: extendedRtpCapabilities,
		handler:                 NewPionHandler(),
		producers:               make(map[string]*Producer),
		producerIdChan:          make(chan string, 1), // 发消息是阻塞的，这里必现带缓冲
	}

	transport.handler.run(HandlerRunOptions{
		direction:              direction,
		iceParameters:          iceParameters,
		iceCandidates:          iceCandidates,
		dtlsParameters:         dtlsParameters,
		sctpParameters:         sctpParameters,
		extenedRtpCapabilities: extendedRtpCapabilities,
	})

	transport.handleHandler()

	return transport
}

type TransportProduceOptions struct {
	Track        webrtc.TrackLocal
	CodecOptions sdp.ProducerCodecOptions
	Codec        *mediasoup.RtpCodecParameters
	Encodings    []mediasoup.RtpEncodingParameters
	OnRtpSender  func(*webrtc.RTPSender)
	AppData      any
}

func (t *Transport) Produce(options TransportProduceOptions) *Producer {
	track, codecOptions, codec, onRtpSender, appData, encodings := options.Track, options.CodecOptions, options.Codec, options.OnRtpSender, options.AppData, options.Encodings

	localId, rtpParameters, _ := t.handler.send(HandlerSendOptions{
		track:        track,
		codecOptions: codecOptions,
		codec:        codec,
		encodings:    encodings,
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
		Track:         track,
		RtpParameters: rtpParameters,
		AppData:       appData,
	})

	t.producers[id] = producer
	t.handleProducer(producer)

	t.SafeEmit("newproducer", producer)

	return producer
}

type TransportProduceDataOptions struct {
	Ordered        bool
	MaxRetransmits int
	Label          string
	Priority       string
	AppData        any
}

func (t *Transport) ProduceData(options TransportProduceDataOptions) *DataProducer {
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

	handler.On("@icegatheringstatechange", func(state string) {
		if !t.closed {
			t.SafeEmit("icegatheringstatechange", state)
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

func (t *Transport) Close() {
	if t.closed {
		return
	}

	t.closed = true

	t.handler.close()

	for _, producer := range t.producers {
		producer.TransportClosed()
	}
	clear(t.producers)

	t.SafeEmit("close")
}
