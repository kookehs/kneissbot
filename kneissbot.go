package main

import (
	"bufio"
	"os"

	"github.com/kookehs/kneissbot/api/twitch"
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
	api := twitch.NewAPI(token)
	response, err := api.ValidToken()

	if err != nil || !response.Token.Valid {
		if !response.Token.Valid {
			panic("Invalid token")
		}

		panic(err)
	}

	username := response.Token.Username
	bot, err := core.NewBot(username)

	if err != nil {
		panic(err)
	}

	defer bot.Close()
	bot.Start()

	if ok := bot.Connect(username, token); !ok {
		panic("Unable to connect to IRC")
	}

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
