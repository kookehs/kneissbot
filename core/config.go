package core

type Config struct {
	Twitch *TwitchConfig
}

func NewConfig() *Config {
	return &Config{
		Twitch: NewTwitchConfig(),
	}
}
