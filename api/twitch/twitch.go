package twitch

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	HelixAPI  = "https://api.twitch.tv/helix"
	KrakenAPI = "https://api.twitch.tv/kraken"
)

type API struct {
	Client *http.Client
	Token  string
}

func NewAPI(token string) *API {
	return &API{
		Client: new(http.Client),
		Token:  token,
	}
}

func AuthType(url string) string {
	if strings.Contains(url, "helix") {
		return "Bearer"
	}

	return "OAuth"
}

func (a *API) Get(url string) ([]byte, error) {
	// TODO: Validate every token before each request
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", AuthType(url)+" "+a.Token)
	resp, err := a.Client.Do(req)

	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return body, nil
}

func (a *API) ValidToken(url string) (bool, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", AuthType(url)+" "+a.Token)
	resp, err := a.Client.Do(req)

	if err != nil {
		return false, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// TODO: Build structure for token validation
	log.Println(string(body))
	return false, nil
}
