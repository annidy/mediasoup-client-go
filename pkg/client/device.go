package client

import (
	"sync/atomic"

	"github.com/annidy/mediasoup-client/internal/util"
	"github.com/jiyeyuran/mediasoup-go"
	"github.com/rs/zerolog/log"
)

type Device struct {
	mediasoup.IEventEmitter
	Name                  string
	recvRtpCapabilities   mediasoup.RtpCapabilities
	sctpCapabilities      mediasoup.SctpCapabilities
	extendRtpCapabilities RtpCapabilitiesEx
	loaded                atomic.Bool
	handler               *PionHandler
	canProduceByKind      map[mediasoup.MediaKind]bool
}

func NewDevice() *Device {
	handler := NewPionHandler()

	return &Device{
		Name:    "pion",
		handler: handler,
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
	util.Clone(routerRtpCapabilities, &clonedRouterRtpCapabilities)
	if err := ortc.validateRtpCapabilities(&clonedRouterRtpCapabilities); err != nil {
		panic(err)
	}

	nativeRtpCapabilities := d.handler.getNativeRouterRtpCapabilities()
	var clonedNativeRtpCapabilities mediasoup.RtpCapabilities
	util.Clone(nativeRtpCapabilities, &clonedNativeRtpCapabilities)
	if err := ortc.validateRtpCapabilities(&clonedNativeRtpCapabilities); err != nil {
		panic(err)
	}

	d.extendRtpCapabilities = ortc.getExtendedRtpCapabilities(clonedNativeRtpCapabilities, clonedRouterRtpCapabilities)

	// TODO: check whether we can produce audio/video

	d.recvRtpCapabilities = ortc.getRecvRtpCapabilities(d.extendRtpCapabilities)
	if err := ortc.validateRtpCapabilities(&d.recvRtpCapabilities); err != nil {
		panic(err)
	}
	d.sctpCapabilities = d.handler.getNativeSctpCapabilities()
}

func (d *Device) CreateSendTransport(transportInfo WebrtcTransportInfo) *Transport {
	return nil
}

func (d *Device) CreateRecvTransport(transportInfo WebrtcTransportInfo) *Transport {
	return nil
}

func (d *Device) createTransport() {

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
