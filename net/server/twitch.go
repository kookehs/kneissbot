package server

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/kookehs/kneissbot/os/exec"
	"golang.org/x/oauth2/twitch"
)

const (
	// Values retrieved from Twitch Apps
	ClientID     = "2qt0hvdtidd4o2p7r0ndjajnawb080"
	RedirectURI  = "http://localhost:8080/twitch"
	ResponseType = "token"
)

var (
	Scopes = []string{"channel_check_subscription", "channel_subscriptions", "chat_login", "communities_moderate"}
)

// TwitchAuthServer contains variables need to set up an OAuth 2 connection
// with the Twitch API.
type TwitchAuthServer struct {
	Channel chan string
	Server  *http.Server
	State   string
}

// NewTwitchAuthServer creates and initializes a local server used to
// authorize an OAuth Implicit Code flow.
func NewTwitchAuthServer(channel chan string) *TwitchAuthServer {
	twitchAuthServer := new(TwitchAuthServer)
	twitchAuthServer.Channel = channel
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/token", twitchAuthServer.TokenRetrieval)
	serveMux.HandleFunc("/twitch", twitchAuthServer.TwitchAuthorization)
	server := new(http.Server)
	server.Addr = ":8080"
	server.Handler = serveMux
	twitchAuthServer.Server = server
	return twitchAuthServer
}

// Authenticate generates a state key and redirects the user to
// Twitch's authorization page.
func (tas *TwitchAuthServer) Authenticate() error {
	key := make([]byte, 64)

	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return err
	}

	tas.State = base64.StdEncoding.EncodeToString(key)
	query := tas.BuildQuery()
	href := twitch.Endpoint.AuthURL + "?" + query.Encode()
	tas.RedirectToURL(href)
	return nil
}

// BuildQuery constructs the query part of the URL
func (tas *TwitchAuthServer) BuildQuery() url.Values {
	var scopes bytes.Buffer

	for i, v := range Scopes {
		if i != 0 {
			scopes.WriteByte(' ')
		}

		scopes.WriteString(v)
	}

	query := make(url.Values)
	query.Add("client_id", ClientID)
	query.Add("redirect_uri", RedirectURI)
	query.Add("response_type", ResponseType)
	query.Add("scope", scopes.String())
	query.Add("state", url.QueryEscape(tas.State))
	return query
}

// Close forcefully shuts down the underlying server.
func (tas *TwitchAuthServer) Close() error {
	if err := tas.Server.Close(); err != nil {
		return err
	}

	return nil
}

// ListenAndServe instructs the underlying server to begin listening
// and handling requests.
func (tas *TwitchAuthServer) ListenAndServe() error {
	if err := tas.Server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

// RedirectToURL opens the given URL with the user's default browser.
func (tas *TwitchAuthServer) RedirectToURL(url string) error {
	if err := exec.OpenBrowser(url); err != nil {
		return err
	}

	return nil
}

// TokenRetrieval checks the state variable and extracts the
// access token from the URL.
func (tas *TwitchAuthServer) TokenRetrieval(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		log.Println(err)
		return
	}

	href, err := url.Parse(string(body[1 : len(body)-1]))

	if err != nil {
		log.Println(err)
		return
	}

	split := strings.IndexByte(href.Fragment, '&')
	fragment := href.Fragment[:split]
	query, err := url.ParseQuery(href.Fragment[split:])

	if err != nil {
		log.Println(err)
		return
	}

	if strings.Compare(tas.State, query["state"][0]) == 0 {
		token := fragment[strings.IndexByte(fragment, '=')+1:]
		tas.Channel <- token
	}
}

// TwitchAuthorization handles the Twitch redirect by serving a HTML file.
func (tas *TwitchAuthServer) TwitchAuthorization(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/twitch.html")
}
