package core

type TwitchConfig struct {
	AccessToken string
	Username    string
}

func NewTwitchConfig() *TwitchConfig {
	return &TwitchConfig{}
}
