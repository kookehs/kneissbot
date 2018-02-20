package twitch

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	HelixAPI  = "https://api.twitch.tv/helix"
	KrakenAPI = "https://api.twitch.tv/kraken"
	GetUsers  = "https://api.twitch.tv/helix/users"
)

// API is a structure used to communicate with the Twitch API. Stores the
// access token as well as a http.Client.
type API struct {
	Client *http.Client
	Token  string
}

// NewAPI creates and initilaizes an API. NewAPI accepts an access token
// as its parameter.
func NewAPI(token string) *API {
	return &API{
		Client: new(http.Client),
		Token:  token,
	}
}

// AuthType returns appropriate authorization type based on given URL.
func AuthType(url string) string {
	if strings.Contains(url, "helix") {
		return "Bearer"
	}

	return "OAuth"
}

// Get sends a GET request to the specified URL returning the body
// as bytes or an error.
func (a *API) Get(url string) ([]byte, error) {
	if valid, err := a.ValidToken(); !valid || err != nil {
		if !valid {
			return nil, errors.New("invalid token")
		}

		return nil, err
	}

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

// GetUsers returns a UserResponse with the supplied arguments.
// If no arguments are given, then the user to which the access
// token belongs to is returned.
func (a *API) GetUsers(id, login []string) (*UsersResponse, error) {
	query := make(url.Values)

	for _, v := range id {
		query.Add("id", v)
	}

	for _, v := range login {
		query.Add("login", v)
	}

	body, err := a.Get(GetUsers + "?" + query.Encode())

	if err != nil {
		return nil, err
	}

	log.Println(string(body))
	resp := new(UsersResponse)

	if err = json.Unmarshal(body, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// ValidToken sends a request to the root URL to check if access
// token is still valid. Tokens must be validated before each request.
func (a *API) ValidToken() (bool, error) {
	req, err := http.NewRequest(http.MethodGet, KrakenAPI, nil)

	if err != nil {
		return false, err
	}

	req.Header.Add("Authorization", AuthType(KrakenAPI)+" "+a.Token)
	resp, err := a.Client.Do(req)

	if err != nil {
		return false, err
	}

	decoder := json.NewDecoder(resp.Body)
	token := new(TokenResponse)
	err = decoder.Decode(token)
	resp.Body.Close()

	if err != nil {
		return false, nil
	}

	return token.Token.Valid, nil
}
