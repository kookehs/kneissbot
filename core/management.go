package core

import (
	"log"
	"math"

	watchmen "github.com/kookehs/watchmen/core"
)

// DefaultModerators is the default number of moderators.
const DefaultModerators = 3

// Management handles logic related to dynamically managing moderators.
type Management struct {
	// Magement variables
	MovingAverage *MovingAverage

	// Twitch related variables
	Bans       int
	Messages   uint64
	Moderators int
	Timeouts   int

	// Watchmen modules
	DPoS   *watchmen.DPoS
	Ledger *watchmen.Ledger
	Node   *watchmen.Node
}

// NewManagement creates and initializes a new Management.
func NewManagement() *Management {
	dpos := watchmen.NewDPoS()
	ledger := watchmen.NewLedger()
	ma := NewMovingAverage(10)
	node := watchmen.NewNode(dpos, ledger)

	return &Management{
		DPoS:          dpos,
		Ledger:        ledger,
		Moderators:    DefaultModerators,
		MovingAverage: ma,
		Node:          node,
	}
}

// Heuristic is used to estimate the number of moderators needed.
// Take a conversative, slow approach to making people moderators.
// DPoS will handle the quality of moderators.
func (m *Management) Heuristic() int {
	m.MovingAverage.Append(m.Score())
	sma, ema, _ := m.MovingAverage.Update()
	diff := math.Abs(sma - ema)
	effectiveness := diff / float64(m.Moderators) * 2
	mods := float64(m.Moderators)
	multiplier := 1.0

	log.Printf("EMA: %v, SMA: %v, Signal: %v", ema, sma, m.MovingAverage.Signal)

	switch m.MovingAverage.Signal {
	case -1:
		// A lower sma than ema indicates a good trend.
		// Quarter circle shifted up and right by 1.
		// A peak value at x = 1 means a 1:1 ratio in trend to moderators.
		scaling := math.Sqrt(math.Abs(1-math.Pow(effectiveness, 2))) + 1
		multiplier = 1 / scaling
	case 0:
		// No action should be taken as we are unsure which way the trend will go.
	case 1:
		// A higher sma than ema indicates a bad trend.
		// Take a slow approach as the spread grows.
		multiplier = math.Pow(effectiveness, 1/4)
	}

	if mods *= multiplier; mods < DefaultModerators {
		mods = DefaultModerators
	}

	// Truncate value to avoid rounding up in order to not overestimate.
	return int(mods)
}

// Score helps to quanitify the effectiveness of moderators.
// A low score indicates a high amount of infractions or low activity in chat.
func (m *Management) Score() float64 {
	// Calculate infractions relative to messages.
	score := float64(m.Messages)
	// TODO: Add weights for bans and timeouts.
	infractions := float64(m.Bans + m.Timeouts)

	// TODO: Messages need a bigger weight.
	if infractions > 0 {
		score /= infractions
	}

	return score
}

// Update updates variables and calculates the
// number of moderators based on the heuristic.
func (m *Management) Update() {
	m.Moderators = m.Heuristic()
}
