package sdp

/*
#cgo LDFLAGS: /usr/local/lib/libsdptransform.a -lstdc++
#include <sdptransform/transform.h>
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"unsafe"
)

func Write(spdObject Sdp) (spdStr string) {
	jsonStr, _ := json.Marshal(spdObject)
	cJson := C.CString(string(jsonStr))
	defer C.free(unsafe.Pointer(cJson))
	cResult := C.sdptransform_parse(cJson)
	spdStr = C.GoString(cResult)
	C.free(unsafe.Pointer(cResult))
	return
}

func Parse(sdpStr string) (spdObject Sdp) {
	cSdp := C.CString(sdpStr)
	defer C.free(unsafe.Pointer(cSdp))
	cResult := C.sdptransform_parse(cSdp)
	jsonStr := C.GoString(cResult)
	C.free(unsafe.Pointer(cResult))
	json.Unmarshal([]byte(jsonStr), &spdObject)
	return
}

func ParseParams(params string) (spdParamsObject SpdParams) {
	cStr := C.CString(params)
	defer C.free(unsafe.Pointer(cStr))
	cResult := C.sdptransform_parse_params(cStr)
	jsonStr := C.GoString(cResult)
	C.free(unsafe.Pointer(cResult))
	json.Unmarshal([]byte(jsonStr), &spdParamsObject)
	return
}

type Sdp struct {
	Connection struct {
		IP      string `json:"ip,omitempty"`
		Version int    `json:"version,omitempty"`
	} `json:"connection,omitempty"`
	Fingerprint struct {
		Hash string `json:"hash,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"fingerprint,omitempty"`
	IcePwd   string `json:"icePwd,omitempty"`
	IceUfrag string `json:"iceUfrag,omitempty"`
	Media    []struct {
		Candidates []struct {
			Component  int    `json:"component,omitempty"`
			Foundation string `json:"foundation,omitempty"`
			IP         string `json:"ip,omitempty"`
			Port       int    `json:"port,omitempty"`
			Priority   int    `json:"priority,omitempty"`
			Transport  string `json:"transport,omitempty"`
			Type       string `json:"type,omitempty"`
		} `json:"candidates,omitempty"`
		Direction string `json:"direction,omitempty"`
		Fmtp      []struct {
			Config  string `json:"config,omitempty"`
			Payload int    `json:"payload,omitempty"`
		} `json:"fmtp,omitempty"`
		Payloads string `json:"payloads,omitempty"`
		Port     int    `json:"port,omitempty"`
		Protocol string `json:"protocol,omitempty"`
		Ptime    int    `json:"ptime,omitempty"`
		Rtp      []struct {
			Codec   string `json:"codec,omitempty"`
			Payload int    `json:"payload,omitempty"`
			Rate    int    `json:"rate,omitempty"`
		} `json:"rtp,omitempty"`
		Type   string `json:"type,omitempty"`
		RtcpFb []struct {
			Payload string `json:"payload,omitempty"`
			Type    string `json:"type,omitempty"`
			Subtype string `json:"subtype,omitempty"`
		} `json:"rtcpFb,omitempty"`
		RtcpFbTrrInt []struct {
			Payload string `json:"payload,omitempty"`
			Value   int    `json:"value,omitempty"`
		} `json:"rtcpFbTrrInt,omitempty"`
		Ssrcs []struct {
			Attribute string `json:"attribute,omitempty"`
			ID        int    `json:"id,omitempty"`
			Value     string `json:"value,omitempty"`
		} `json:"ssrcs,omitempty"`
		Ext []struct {
			EncryptUri string `json:"encrypt-uri,omitempty"`
			Uri        string `json:"uri,omitempty"`
			Value      int    `json:"value,omitempty"`
		} `json:"ext,omitempty"`
	} `json:"media,omitempty"`
	Name   string `json:"name,omitempty"`
	Origin struct {
		Address        string `json:"address,omitempty"`
		IPVer          int    `json:"ipVer,omitempty"`
		NetType        string `json:"netType,omitempty"`
		SessionID      int    `json:"sessionId,omitempty"`
		SessionVersion int    `json:"sessionVersion,omitempty"`
		Username       string `json:"username,omitempty"`
	} `json:"origin,omitempty"`
	Timing struct {
		Start int `json:"start,omitempty"`
		Stop  int `json:"stop,omitempty"`
	} `json:"timing,omitempty"`
	Version int `json:"version,omitempty"`
}

type SpdParams struct {
	ProfileLevelId        string `json:"profile-level-id,omitempty"`
	PacketizationMode     int    `json:"packetization-mode,omitempty"`
	LevelAsymmetryAllowed int    `json:"level-asymmetry-allowed,omitempty"`
}
