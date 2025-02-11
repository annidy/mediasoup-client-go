package client

import (
	"sync/atomic"

	"github.com/annidy/mediasoup-client/internal/utils"
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/rs/zerolog/log"
)

type Device struct {
	mediasoup.IEventEmitter
	Name                    string
	recvRtpCapabilities     mediasoup.RtpCapabilities
	sctpCapabilities        mediasoup.SctpCapabilities
	extendedRtpCapabilities *RtpCapabilitiesEx
	loaded                  atomic.Bool
	handlerFactory          func() *PionHandler
	canProduceByKind        map[mediasoup.MediaKind]bool
}

func NewDevice() *Device {

	return &Device{
		Name:           "pion",
		handlerFactory: NewPionHandler,
		canProduceByKind: map[mediasoup.MediaKind]bool{
			mediasoup.MediaKind_Audio: true,
			mediasoup.MediaKind_Video: true,
		},
	}
}

func (d *Device) Load(routerRtpCapabilities RtpCapabilities) {
	if !d.loaded.CompareAndSwap(false, true) {
		log.Info().Msg("already loaded")
		return
	}
	var clonedRouterRtpCapabilities mediasoup.RtpCapabilities
	utils.Clone(routerRtpCapabilities, &clonedRouterRtpCapabilities)
	if err := ortc.validateRtpCapabilities(&clonedRouterRtpCapabilities); err != nil {
		panic(err)
	}

	handler := d.handlerFactory()
	defer handler.close()

	nativeRtpCapabilities := handler.getNativeRouterRtpCapabilities()
	var clonedNativeRtpCapabilities mediasoup.RtpCapabilities
	utils.Clone(nativeRtpCapabilities, &clonedNativeRtpCapabilities)
	if err := ortc.validateRtpCapabilities(&clonedNativeRtpCapabilities); err != nil {
		panic(err)
	}

	d.extendedRtpCapabilities = ortc.getExtendedRtpCapabilities(clonedNativeRtpCapabilities, clonedRouterRtpCapabilities)

	// TODO: check whether we can produce audio/video

	d.recvRtpCapabilities = ortc.getRecvRtpCapabilities(d.extendedRtpCapabilities)
	if err := ortc.validateRtpCapabilities(&d.recvRtpCapabilities); err != nil {
		panic(err)
	}
	d.sctpCapabilities = handler.getNativeSctpCapabilities()
}

func (d *Device) CreateSendTransport(options DeviceCreateTransportOptions) *Transport {
	options.Direction = "send"
	return d.createTransport(options)
}

func (d *Device) CreateRecvTransport(options DeviceCreateTransportOptions) *Transport {
	options.Direction = "recv"
	return d.createTransport(options)
}

func (d *Device) createTransport(options DeviceCreateTransportOptions) *Transport {
	options.extendedRtpCapabilities = d.extendedRtpCapabilities
	transport := newTransport(options)

	transport.SafeEmit("newtransport", transport)

	return transport
}

func (d *Device) RtpCapabilities() *RtpCapabilities {
	if !d.loaded.Load() {
		panic("not loaded")
	}
	return &d.recvRtpCapabilities
}

func (d *Device) SctpCapabilities() *SctpCapabilities {
	if !d.loaded.Load() {
		panic("not loaded")
	}
	return &d.sctpCapabilities
}

func (d *Device) DeviceInfo() DeviceInfo {
	return DeviceInfo{
		Name:    "pion",
		Version: "0.0.0",
		Flag:    "pion",
	}
}
