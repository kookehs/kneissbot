package server

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/kookehs/kneissbot/os/exec"
	"golang.org/x/oauth2"
)

type TwitchAuthServer struct {
	Channel chan string
	Config  *oauth2.Config
	Server  *http.Server
	State   string
}

func NewTwitchAuthServer(channel chan string, config *oauth2.Config) *TwitchAuthServer {
	twitchAuthServer := new(TwitchAuthServer)
	twitchAuthServer.Channel = channel
	twitchAuthServer.Config = config
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/twitch", twitchAuthServer.TwitchAuthorization)
	server := new(http.Server)
	server.Addr = ":8080"
	server.Handler = serveMux
	twitchAuthServer.Server = server
	return twitchAuthServer
}

func (tas *TwitchAuthServer) Authenticate() error {
	key := make([]byte, 64)

	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return err
	}

	tas.State = base64.StdEncoding.EncodeToString(key)
	authCodeURL := tas.Config.AuthCodeURL(tas.State, oauth2.AccessTypeOffline)
	tas.RedirectToURL(authCodeURL)
	return nil
}

func (tas *TwitchAuthServer) Close() error {
	if err := tas.Server.Close(); err != nil {
		return err
	}

	return nil
}

func (tas *TwitchAuthServer) ListenAndServe() {
	if err := tas.Server.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func (tas *TwitchAuthServer) RedirectToURL(s string) error {
	if err := exec.OpenBrowser(s); err != nil {
		return err
	}

	return nil
}

func (tas *TwitchAuthServer) TwitchAuthorization(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	state := values.Get("state")

	if strings.Compare(tas.State, state) == 0 {
		tas.Channel <- values.Get("code")
	}
}
