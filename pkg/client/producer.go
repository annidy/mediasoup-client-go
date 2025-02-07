package client

import "github.com/jiyeyuran/mediasoup-go"

type ProducerOptions struct {
	Id            string
	LocalId       string
	Kind          mediasoup.MediaKind
	RtpParameters *mediasoup.RtpParameters
}
type Producer struct {
	mediasoup.IEventEmitter
}

func NewProducer(options ProducerOptions) *Producer {
	return &Producer{
		IEventEmitter: mediasoup.NewEventEmitter(),
	}
}

func (p *Producer) Close() {}

func (p *Producer) Id() string {
	return ""
}

func (p *Producer) Pause() {
}

func (p *Producer) Resume() {
}

type DataProducer struct {
	mediasoup.IEventEmitter
}

func (d *DataProducer) Close() {}

func (d *DataProducer) Id() string {
	return ""
}

func (d *DataProducer) Send(data []byte) {}
