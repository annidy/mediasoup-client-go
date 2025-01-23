package proto

import (
	"github.com/jiyeyuran/mediasoup-go"
)

type Consumer struct {
	AppData        interface{}             `json:"appData,omitempty"`
	Id             string                  `json:"id,omitempty"`
	PeerId         string                  `json:"peerId,omitempty"`
	ProducerId     string                  `json:"producerId,omitempty"`
	Kind           mediasoup.MediaKind     `json:"kind,omitempty"`
	Type           mediasoup.ConsumerType  `json:"type,omitempty"`
	RtpParameters  mediasoup.RtpParameters `json:"rtpParameters,omitempty"`
	ProducerPaused bool                    `json:"producerPaused,omitempty"`
}
