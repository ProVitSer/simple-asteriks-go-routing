package main

import (
	"context"

	"github.com/CyCoreSystems/ari/v6"
	"github.com/CyCoreSystems/ari/v6/client/native"
	"github.com/inconshreveable/log15"
)

var ariApp = "test"

var log = log15.New()

var bridge *ari.BridgeHandle

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Info("Connecting to ARI")

	cl, err := native.Connect(&native.Options{
		Application:  "codims",
		Username:     "codims",
		Password:     "b0098bd73db283d4911a746ddaa6d56f",
		URL:          "http://localhost:8088/ari",
		WebsocketURL: "ws://localhost:8088/ari/events",
	})

	if err != nil {
		log.Error("Failed to build ARI client", "error", err)
		return
	}

	log.Info("Starting listener app")
	log.Info("Listening for new calls")

	sub := cl.Bus().Subscribe(nil, "StasisStart")

	for {
		select {
		case e := <-sub.Events():
			v := e.(*ari.StasisStart)
			log.Info("Got stasis start", "channel", v.Channel.ID)
			go app(ctx, cl, cl.Channel().Get(v.Key(ari.ChannelKey, v.Channel.ID)))
		case <-ctx.Done():
			return
		}
	}
}

func app(ctx context.Context, cl ari.Client, h *ari.ChannelHandle) {

}
