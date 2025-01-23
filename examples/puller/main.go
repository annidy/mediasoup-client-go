package main

import (
	"net/url"

	"github.com/annidy/mediasoup-client/internal/util"
	"github.com/annidy/mediasoup-client/pkg/room"
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Scheme string
		Host   string
		Path   string
	}
}

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	config := Config{}
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	u := url.URL{Scheme: config.Server.Scheme, Host: config.Server.Host, Path: config.Server.Path}
	q := u.Query()
	q.Set("roomId", "888389527")
	q.Set("peerId", util.RandomAlpha(6))
	u.RawQuery = q.Encode()

	roomClient := &room.RoomClient{}

	roomClient.Join(u.String())

	select {}
}
