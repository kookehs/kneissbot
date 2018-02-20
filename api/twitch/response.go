package twitch

// AuthorizationResponse is the json structure returned by the
// Twitch API. Contains variables related to token authorization.
type AuthorizationResponse struct {
	CreatedAt string   `json:"created_at"`
	Scopes    []string `json:"scopes"`
	UpdatedAt string   `json:"updated_at"`
}

// LinksResponse is the json structure returned by the
// Twitch API. Contains variables related to various links.
type LinksResponse struct {
	Channel  string `json:"channel"`
	Channels string `json:"channels"`
	Chat     string `json:"chat"`
	Ingests  string `json:"ingests"`
	Streams  string `json:"streams"`
	Teams    string `json:"teams"`
	User     string `json:"user"`
	Users    string `json:"users"`
}

// TokenInfoResponse is the json structure returned by the
// Twitch API. Contains variables related to token bearer.
type TokenInfoResponse struct {
	Authorization AuthorizationResponse `json:"authorization"`
	ClientID      string                `json:"client_id"`
	Username      string                `json:"user_name"`
	Valid         bool                  `json:"valid"`
}

// TokenResponse is the json structure returned by the
// Twitch API. Contains variables related to the OAuth 2 token.
type TokenResponse struct {
	Identified bool              `json:"identified"`
	Links      LinksResponse     `json:"_links"`
	Token      TokenInfoResponse `json:"token"`
}

// UserResponse is the json structure returned by the
// Twitch API. Contains variables related to user retrieved.
type UserResponse struct {
	BroadcasterType string `json:"broadcaster_type"`
	Description     string `json:"description"`
	DisplayName     string `json:"display_name"`
	ID              string `json:"id"`
	Login           string `json:"login"`
	OfflineImageURL string `json:"offline_image_url"`
	ProfileImageURL string `json:"profile_image_url"`
	Type            string `json:"type"`
	ViewCount       uint64 `json:"view_count"`
}

// UsersResponse is the json structure returned by the
// Twitch API. Contains variables related to users retrieved.
type UsersResponse struct {
	Data []UserResponse `json:"data"`
}
