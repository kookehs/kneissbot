package main

import (
	"github.com/kookehs/kneissbot/core"
	"github.com/kookehs/kneissbot/net/server"
)

func main() {
	output := make(chan string)
	defer close(output)
	twitchAuth := server.NewTwitchAuth(output)
	go twitchAuth.ListenAndServe()

	if err := twitchAuth.Authenticate(); err != nil {
		panic(err)
	}

	defer twitchAuth.Close()
	token := <-output
	twitch, err := core.NewTwitch()

	if err != nil {
		panic(err)
	}

	defer twitch.Close()
	twitch.Start()

	if ok := twitch.Connect("kneissbot", token); !ok {
		panic("Unable to connect to IRC")
	}

	channel := "lirik"

	if ok := twitch.Join(channel); !ok {
		panic("Unable to join IRC channel")
	}

	twitch.Cap(core.DefaultCapabilities)
	select {}
}
