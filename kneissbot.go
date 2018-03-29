package main

import (
	"bufio"
	"os"

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
	bot, err := core.NewBot(token)

	if err != nil {
		panic(err)
	}

	defer bot.Close()
	bot.Start()

	if ok := bot.Connect(); !ok {
		panic("Unable to connect to IRC")
	}

	// TODO: Remove test code
	reader := bufio.NewReader(os.Stdin)
	channel, err := reader.ReadString('\n')

	if err != nil {
		panic(err)
	}

	if ok := bot.Join(channel); !ok {
		panic("Unable to join IRC channel")
	}

	bot.Cap(nil)
	select {}
}
