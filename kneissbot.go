package main

import (
	"bufio"
	"os"
	"time"

	"github.com/kookehs/kneissbot/net/irc"
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
	origin := "http://twitch.tv"
	url := "ws://irc-ws.chat.twitch.tv:80"
	twitchIRC, err := irc.NewTwitch(origin, url)

	if err != nil {
		panic(err)
	}

	defer twitchIRC.Close()
	twitchIRC.Start()

	if ok := twitchIRC.Connect("kneissbot", token); !ok {
		panic("Unable to connect to IRC")
	}

	channel := "lirik"

	if ok := twitchIRC.Join(channel); !ok {
		panic("Unable to join IRC channel")
	}

	twitchIRC.Cap([]string{irc.Commands, irc.Membership, irc.Tags})
	go twitchIRC.HandleCommands()

	path := "data/" + channel + ".txt"
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

	if err != nil {
		panic(err)
	}

	defer file.Close()
	writer := bufio.NewWriter(file)

	for {
		in := <-twitchIRC.In
		out := time.Now().UTC().Format(time.RFC3339) + " " + string(in)

		if len(out)+writer.Buffered() >= writer.Available() {
			if err := writer.Flush(); err != nil {
				panic(err)
			}
		}

		writer.WriteString(out)
	}

	// management := core.NewManagement()

	// for {
	// 	split := strings.Index(messages, "Z ")
	// 	messages := messages[split+2:]
	// 	individual := strings.Split(messages, "\n")
	// 	var bans, messages int
	// 	if strings.Contains(message, "@ban-duration") {}
	// 	fmt.Scanf("%d %d\n", &messages, &bans)
	// 	management.Bans = bans
	// 	management.Messages = uint64(messages)
	// 	management.Update()
	// 	log.Printf("Mods: %v", management.Moderators)
	// }
}
