package twitch

type Authorization struct {
	CreatedAt string   `json:"created_at"`
	Scopes    []string `json:"scopes"`
	UpdatedAt string   `json:"updated_at"`
}

type Links struct {
	Channel  string `json:"channel"`
	Channels string `json:"channels"`
	Chat     string `json:"chat"`
	Ingests  string `json:"ingests"`
	Streams  string `json:"streams"`
	Teams    string `json:"teams"`
	User     string `json:"user"`
	Users    string `json:"users"`
}

type TokenInfo struct {
	Athorization Authorization `json:"authorization"`
	ClientID     string        `json:"client_id"`
	Username     string        `json:"user_name"`
	Valid        bool          `json:"valid"`
}

type TokenValidation struct {
	Identified bool      `json:"identified"`
	Links      Links     `json:"_links"`
	Token      TokenInfo `json:"token"`
}
