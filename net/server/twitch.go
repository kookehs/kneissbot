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
	// ClientID is provided by Twitch
	ClientID = "2qt0hvdtidd4o2p7r0ndjajnawb080"
	// RedirectURI should match the URL entered when registering app on Twitch
	RedirectURI = "http://localhost:8080/twitch"
	// ResponseType must be token for OAuth 2 Implicit Code Flow
	ResponseType = "token"
)

var (
	// Scopes based on requirements of the app
	Scopes = []string{"channel_check_subscription", "channel_subscriptions", "chat_login", "communities_moderate"}
)

// TwitchAuth contains variables need to set up an OAuth 2 connection
// with the Twitch API.
type TwitchAuth struct {
	Channel chan string
	Server  *http.Server
	State   string
}

// NewTwitchAuth creates and initializes a local server used to
// authorize an OAuth Implicit Code flow.
func NewTwitchAuth(channel chan string) *TwitchAuth {
	twitchAuth := new(TwitchAuth)
	twitchAuth.Channel = channel
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/token", twitchAuth.TokenRetrieval)
	serveMux.HandleFunc("/twitch", twitchAuth.TwitchAuthorization)
	server := new(http.Server)
	server.Addr = ":8080"
	server.Handler = serveMux
	twitchAuth.Server = server
	return twitchAuth
}

// Authenticate generates a state key and redirects the user to
// Twitch's authorization page.
func (ta *TwitchAuth) Authenticate() error {
	key := make([]byte, 64)

	if _, err := io.ReadFull(rand.Reader, key[:]); err != nil {
		return err
	}

	ta.State = base64.StdEncoding.EncodeToString(key)
	query := ta.BuildQuery()
	href := twitch.Endpoint.AuthURL + "?" + query.Encode()
	return ta.RedirectToURL(href)
}

// BuildQuery constructs the query part of the URL
func (ta *TwitchAuth) BuildQuery() url.Values {
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
	query.Add("state", url.QueryEscape(ta.State))
	return query
}

// Close forcefully shuts down the underlying server.
func (ta *TwitchAuth) Close() error {
	return ta.Server.Close()
}

// ListenAndServe instructs the underlying server to begin listening
// and handling requests.
func (ta *TwitchAuth) ListenAndServe() error {
	return ta.Server.ListenAndServe()
}

// RedirectToURL opens the given URL with the user's default browser.
func (ta *TwitchAuth) RedirectToURL(url string) error {
	return exec.OpenBrowser(url)
}

// TokenRetrieval checks the state variable and extracts the
// access token from the URL.
func (ta *TwitchAuth) TokenRetrieval(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println(err)
		return
	}

	if err = r.Body.Close(); err != nil {
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

	if strings.Compare(ta.State, query["state"][0]) == 0 {
		token := fragment[strings.IndexByte(fragment, '=')+1:]
		ta.Channel <- token
	}
}

// TwitchAuthorization handles the Twitch redirect by serving a HTML file.
func (ta *TwitchAuth) TwitchAuthorization(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/twitch.html")
}
