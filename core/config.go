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
		Twitch: NewTwitchConfig(),
	}
}

// Deserialize decodes byte data encoded by gob to Config structure.
func (c *Config) Deserialize(r io.Reader) error {
	decoder := gob.NewDecoder(r)
	return decoder.Decode(c)
}

// Serialize encodes Config structure to byte data using gob.
func (c *Config) Serialize(w io.Writer) error {
	encoder := gob.NewEncoder(w)
	return encoder.Encode(c)
}
