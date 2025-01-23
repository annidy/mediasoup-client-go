package client

import "github.com/jiyeyuran/mediasoup-go"

type Transport struct {
	mediasoup.ITransport
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
