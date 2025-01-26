package room

import (
	"github.com/jiyeyuran/go-protoo"
	"github.com/rs/zerolog/log"
)

type Protoo struct {
	*protoo.Peer
}

func (p *Protoo) RequestData(method string, data interface{}, resp interface{}) (err error) {
	log.Info().Str("method", method).Interface("data", data).Msg("request")
	rsp := p.Request(method, data)
	if rsp.Err() != nil {
		err = rsp.Err()
		return
	}
	err = json.Unmarshal(rsp.Data(), resp)
	log.Info().Str("method", method).Interface("data", resp).Msg("response")
	return
}
