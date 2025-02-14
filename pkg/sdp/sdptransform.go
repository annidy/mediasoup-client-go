package sdp

/*
#cgo LDFLAGS: /usr/local/lib/libsdptransform.a -lstdc++
#include <sdptransform/transform.h>
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"fmt"
	"unsafe"
)

func Write(spdObject Sdp) (spdStr string) {
	jsonStr, _ := json.Marshal(spdObject)
	fmt.Println(string(jsonStr))
	cJson := C.CString(string(jsonStr))
	defer C.free(unsafe.Pointer(cJson))
	cResult := C.sdptransform_write(cJson)
	spdStr = C.GoString(cResult)
	C.free(unsafe.Pointer(cResult))
	fmt.Println(spdStr)
	return
}

func Parse(sdpStr string) (spdObject Sdp) {
	cSdp := C.CString(sdpStr)
	defer C.free(unsafe.Pointer(cSdp))
	cResult := C.sdptransform_parse(cSdp)
	jsonStr := C.GoString(cResult)
	C.free(unsafe.Pointer(cResult))
	json.Unmarshal([]byte(jsonStr), &spdObject)
	checkMissingKeys[Sdp](jsonStr)
	fmt.Println(jsonStr)
	return
}

func ParseParams(params string) (spdParamsObject SpdParams) {
	cStr := C.CString(params)
	defer C.free(unsafe.Pointer(cStr))
	cResult := C.sdptransform_parse_params(cStr)
	jsonStr := C.GoString(cResult)
	C.free(unsafe.Pointer(cResult))
	json.Unmarshal([]byte(jsonStr), &spdParamsObject)
	checkMissingKeys[SpdParams](jsonStr)
	return
}

type MeidaCandidates struct {
	Component  int    `json:"component"`
	Foundation string `json:"foundation"`
	Generation int    `json:"generation,omitempty"`
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Priority   int    `json:"priority"`
	Transport  string `json:"transport"`
	Type       string `json:"type"`
	Tcptype    string `json:"tcptype,omitempty"`
	Raddr      string `json:"raddr,omitempty"`
	Rport      int    `json:"rport,omitempty"`
}

type MediaObject struct {
	Candidates []MeidaCandidates `json:"candidates,omitempty"`
	Connection struct {
		IP      string `json:"ip"`
		Version int    `json:"version"`
	} `json:"connection,omitzero"`
	Crypto []struct {
		Config string `json:"config"`
		ID     int    `json:"id"`
		Suite  string `json:"suite"`
	} `json:"crypto,omitempty"`
	Direction string `json:"direction,omitempty"`
	Fmtp      []struct {
		Config  string `json:"config,omitempty"`
		Payload int    `json:"payload,omitempty"`
	} `json:"fmtp,omitempty"`
	IceOptions      string `json:"iceOptions,omitempty"`
	EndOfCandidates string `json:"endOfCandidates,omitempty"`
	IcePwd          string `json:"icePwd,omitempty"`
	IceUfrag        string `json:"iceUfrag,omitempty"`
	Maxptime        int    `json:"maxptime,omitempty"`
	Mid             string `json:"mid,omitempty"`
	Payloads        string `json:"payloads,omitempty"`
	Port            int    `json:"port,omitempty"`
	Protocol        string `json:"protocol,omitempty"`
	Ptime           int    `json:"ptime,omitempty"`
	Rtcp            struct {
		Address string `json:"address,omitempty"`
		IPVer   int    `json:"ipVer,omitempty"`
		NetType string `json:"netType,omitempty"`
		Port    int    `json:"port,omitempty"`
	} `json:"rtcp,omitempty"`
	RtcpMux   string `json:"rtcpMux,omitempty"`
	RtcpRsize string `json:"rtcpRsize,omitempty"`
	Rtp       []struct {
		Codec    string `json:"codec,omitempty"`
		Payload  int    `json:"payload,omitempty"`
		Rate     int    `json:"rate,omitempty"`
		Encoding string `json:"encoding,omitempty"`
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
	Framerate float64 `json:"framerate,omitempty"`
	Ext       []struct {
		EncryptUri string `json:"encrypt-uri,omitempty"`
		Uri        string `json:"uri,omitempty"`
		Value      int    `json:"value,omitempty"`
	} `json:"ext,omitempty"`
	Bandwidth []struct {
		Limit int    `json:"limit,omitempty"`
		Type  string `json:"type,omitempty"`
	} `json:"bandwidth,omitempty"`
	Fingerprint struct {
		Hash string `json:"hash"`
		Type string `json:"type"`
	} `json:"fingerprint,omitzero"`
	Sctpmap struct {
		App            string `json:"app,omitempty"`
		MaxMessageSize int    `json:"maxMessageSize,omitempty"`
		SctpmapNumber  int    `json:"sctpmapNumber,omitempty"`
	} `json:"sctpmap,omitzero"`
	Setup            string `json:"setup,omitempty"`
	ExtmapAllowMixed string `json:"extmapAllowMixed,omitempty"`
}

type Sdp struct {
	Groups []struct {
		Mids string `json:"mids"`
		Type string `json:"type"`
	} `json:"groups"`
	Connection *struct {
		IP      string `json:"ip,omitempty"`
		Version int    `json:"version,omitempty"`
	} `json:"connection,omitempty"`
	Fingerprint struct {
		Hash string `json:"hash,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"fingerprint,omitempty"`
	IcePwd       string        `json:"icePwd,omitempty"`
	IceUfrag     string        `json:"iceUfrag,omitempty"`
	Media        []MediaObject `json:"media,omitempty"`
	MsidSemantic *struct {
		Semantics string `json:"semantic"`
		Token     string `json:"token"`
	} `json:"msidSemantic,omitempty"`
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
		Start int `json:"start"`
		Stop  int `json:"stop"`
	} `json:"timing,omitempty"`
	IceLite string `json:"icelite,omitempty"`
	Version int    `json:"version,omitempty"`
	// 暂时不知道有什么用
	ExtmapAllowMixed string `json:"extmapAllowMixed,omitempty"`
	Invalid          []struct {
		Value string `json:"value,omitempty"`
	} `json:"invalid,omitempty"`
}

type SpdParams struct {
	Apt                   int    `json:"apt,omitempty"`
	Minptime              int    `json:"minptime,omitempty"`
	Useinbandfec          int    `json:"useinbandfec,omitempty"`
	ProfileId             int    `json:"profile-id,omitempty"`
	ProfileLevelId        string `json:"profile-level-id,omitempty"`
	PacketizationMode     *int   `json:"packetization-mode,omitempty"`
	LevelAsymmetryAllowed int    `json:"level-asymmetry-allowed,omitempty"`
}
