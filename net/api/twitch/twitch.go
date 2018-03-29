package twitch

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	// HelixAPI is the root URL for the Helix API
	HelixAPI = "https://api.twitch.tv/helix"
	// KrakenAPI is the root URL for the Kraken API
	KrakenAPI = "https://api.twitch.tv/kraken"
	// GetUsers is the endpoint for retrieving user information
	GetUsers = "https://api.twitch.tv/helix/users"
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
	if response, err := a.ValidToken(); err != nil || !response.Token.Valid {
		if !response.Token.Valid {
			return nil, errors.New("Invalid token")
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

	if err != nil {
		return nil, err
	}

	if err = resp.Body.Close(); err != nil {
		return nil, err
	}

	return body, nil
}

// GetChatters returns a ChattersResponse for the given channel.
func (a *API) GetChatters(channel string) (*ChattersResponse, error) {
	body, err := a.Get("http://tmi.twitch.tv/group/user/" + channel + "/chatters")

	if err != nil {
		return nil, err
	}

	resp := new(ChattersResponse)

	if err = json.Unmarshal(body, resp); err != nil {
		return nil, err
	}

	return resp, nil
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

	resp := new(UsersResponse)

	if err = json.Unmarshal(body, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// ValidToken sends a request to the root URL to check if access
// token is still valid. Tokens must be validated before each request.
func (a *API) ValidToken() (*TokenResponse, error) {
	req, err := http.NewRequest(http.MethodGet, KrakenAPI, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", AuthType(KrakenAPI)+" "+a.Token)
	resp, err := a.Client.Do(req)

	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(resp.Body)
	response := new(TokenResponse)
	err = decoder.Decode(response)

	if err != nil {
		return nil, err
	}

	if err = resp.Body.Close(); err != nil {
		return nil, err
	}

	return response, nil
}
