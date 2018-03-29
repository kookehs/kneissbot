package twitch

const (
	// URLs
	Origin = "http://twitch.tv"
	IRC    = "ws://irc-ws.chat.twitch.tv:80"

	// Capabilities
	CommandsCapability   = "twitch.tv/commands"
	MembershipCapability = "twitch.tv/membership"
	TagsCapability       = "twitch.tv/tags"
)

// Chatters contains various roles of chatters in a stream.
type Chatters struct {
	Admins     []string `json:"admins`
	GlobalMods []string `json:"global_mods"`
	Moderators []string `json:"moderators"`
	Staff      []string `json:"staff"`
	Viewers    []string `json:"viewers"`
}
