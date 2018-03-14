package core

import (
	"encoding/gob"
	"io"
)

// Config contains configuration variables for the bot.
type Config struct {
	Twitch *TwitchConfig
}

// NewConfig creates and initializes a new Config. NewConfig is intended
// to initialize a Config structure for storage of variables.
func NewConfig() *Config {
	return &Config{
		Twitch: &TwitchConfig{},
	}
}

// Deserialize decodes byte data encoded by gob.
func (c *Config) Deserialize(r io.Reader) error {
	decoder := gob.NewDecoder(r)
	return decoder.Decode(c)
}

// Serialize encodes to byte data using gob.
func (c *Config) Serialize(w io.Writer) error {
	encoder := gob.NewEncoder(w)
	return encoder.Encode(c)
}

// TwitchConfig contains variables for Twitch related configurations.
type TwitchConfig struct {
	AccessToken string
	Username    string
}
