package core

// TwitchConfig contains variables for Twitch related configurations.
type TwitchConfig struct {
	AccessToken string
	Username    string
}

// NewTwitchConfig creates and initializes a new NewTwitchConfig.
func NewTwitchConfig() *TwitchConfig {
	return &TwitchConfig{}
}
