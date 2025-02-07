package sdp

import "github.com/jiyeyuran/mediasoup-go"

type MediaSection struct {
	closed      bool
	mid         string
	mediaObject MediaObject
	planB       bool
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
	} `json:"connection"`
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
	EndOfCandidates string `json:"end-of-candidates,omitempty"`
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
	RtcpMux string `json:"rtcpMux,omitempty"`
	Rtp     []struct {
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
		Hash string `json:"hash,omitempty"`
		Type string `json:"type,omitempty"`
	} `json:"fingerprint,omitempty"`
	Sctpmap struct {
		App            string `json:"app,omitempty"`
		MaxMessageSize int    `json:"maxMessageSize,omitempty"`
		SctpmapNumber  int    `json:"sctpmapNumber,omitempty"`
	} `json:"sctpmap,omitempty"`
	Setup string `json:"setup,omitempty"`
}

type MediaSectionOptions struct {
	iceParameters  *mediasoup.IceParameters
	iceCandidates  []*mediasoup.IceCandidate
	dtlsParameters *mediasoup.DtlsParameters
	palnB          bool
}

func NewMediaSection(options MediaSectionOptions) *MediaSection {
	iceParameters, iceCandidates, dtlsParameters, planB := options.iceParameters, options.iceCandidates, options.dtlsParameters, options.palnB

	s := MediaSection{
		planB: planB,
	}

	if iceParameters != nil {
		s.SetIceParameters(iceParameters)
	}
	if iceCandidates != nil {
		s.mediaObject.Candidates = make([]MeidaCandidates, len(iceCandidates))

		for i, candidate := range iceCandidates {
			s.mediaObject.Candidates[i] = MeidaCandidates{
				Component:  1,
				Foundation: candidate.Foundation,
				IP:         candidate.Ip,
				Port:       int(candidate.Port),
				Priority:   int(candidate.Priority),
				Transport:  string(candidate.Protocol),
				Type:       candidate.Type,
				Tcptype:    candidate.TcpType,
			}
		}
		s.mediaObject.EndOfCandidates = "end-of-candidates"
		s.mediaObject.IceOptions = "renomination"
	}

	if dtlsParameters != nil {
		s.SetDtlsRole(dtlsParameters.Role)
	}

	return &s
}

func (ms *MediaSection) SetIceParameters(iceParameters *mediasoup.IceParameters) {
	ms.mediaObject.IceUfrag = iceParameters.UsernameFragment
	ms.mediaObject.IcePwd = iceParameters.Password
}

func (ms *MediaSection) SetDtlsRole(role mediasoup.DtlsRole) {
	switch role {
	case "auto":
		ms.mediaObject.Setup = "actpass"
	case "client":
		ms.mediaObject.Setup = "active"
	case "server":
		ms.mediaObject.Setup = "passive"
	}
}

func (ms *MediaSection) Closed() bool {
	return ms.closed
}

func (ms *MediaSection) Mid() string {
	return ms.mid
}

type AnswerMediaSection struct {
	MediaSection
}
