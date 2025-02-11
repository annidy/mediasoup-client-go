package client

import (
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/pion/webrtc/v4"
)

type ProducerOptions struct {
	Id            string
	LocalId       string
	Track         webrtc.TrackLocal
	RtpParameters *mediasoup.RtpParameters
	AppData       any
}
type Producer struct {
	mediasoup.IEventEmitter
	id            string
	localId       string
	track         webrtc.TrackLocal
	rtpParameters *mediasoup.RtpParameters
	appData       any
	closed        bool
}

func NewProducer(options ProducerOptions) *Producer {
	id, localId, track, rtpParameters, appData := options.Id, options.LocalId, options.Track, options.RtpParameters, options.AppData
	p := &Producer{
		IEventEmitter: mediasoup.NewEventEmitter(),
		id:            id,
		localId:       localId,
		track:         track,
		rtpParameters: rtpParameters,
		appData:       appData,
	}
	p.handleTrack()
	return p
}

func (p *Producer) Close() {
	if p.closed {
		return
	}
	p.closed = true

	p.destroyTrack()

	p.Emit("close")
}

func (p *Producer) TransportClosed() {
	if p.closed {
		return
	}
	p.closed = true

	p.destroyTrack()

	p.Emit("transportclose")
	p.SafeEmit("close")
}

func (p *Producer) destroyTrack() {
	// TODO: stop track
}

func (p *Producer) Id() string {
	return p.id
}

func (p *Producer) Pause() {
}

func (p *Producer) Resume() {
}

func (p *Producer) handleTrack() {
	// TODO: listen track ended event
}

type DataProducer struct {
	mediasoup.IEventEmitter
}

func (d *DataProducer) Close() {}

func (d *DataProducer) Id() string {
	return ""
}

func (d *DataProducer) Send(data []byte) {}
