package main

import (
	"bufio"
	"os"

	"github.com/kookehs/kneissbot/core"
)

func main() {
	bot, err := core.NewBot()

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
