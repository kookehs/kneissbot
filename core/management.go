package core

import (
	watchmen "github.com/kookehs/watchmen/core"
)

// Management handles logic related to dynamically managing moderators.
type Management struct {
	// Magement variables
	MovingAverage *MovingAverage

	// Twitch related variables
	Bans       int
	Messages   uint64
	Moderators int
	Ratio      float64
	Timeouts   int

	// R9K and Slow may not be needed if we just account for message.
	R9K  bool
	Slow bool

	// Watchmen modules
	DPoS   *watchmen.DPoS
	Ledger *watchmen.Ledger
	Node   *watchmen.Node
}

// NewManagement creates and initializes a new Management.
func NewManagement() *Management {
	ledger := watchmen.NewLedger()
	dpos := watchmen.NewDPoS()
	node := watchmen.NewNode(dpos, ledger)

	return &Management{
		DPoS:   dpos,
		Ledger: ledger,
		Node:   node,
	}
}

// Heuristic is used to estimate the number of moderators needed.
// Take a conversative, slow approach to making people moderators.
// DPoS will handle the quality of moderators.
func (m *Management) Heuristic() int {
	m.MovingAverage.Append(m.Score())
	// sma, ema, crossed := m.MovingAverage.Update()
	// diff := math.Abs(sma - ema)

	// if !crossed {
	// 	return m.Moderators
	// }

	// switch m.MovingAverage.Signal {
	// case -1:
	// 	effectiveness := diff / float64(m.Moderators)
	// 	mods := math.Log10(effectiveness)
	// 	return m.Moderators - 1 - int(mods)
	// case 0:
	// 	// No action should be taken as we are unsure which way the trend will go.
	// case 1:
	// 	return m.Moderators + 1
	// }

	// TODO: Factor in the trend into number of moderators.
	return -1
}

// Score helps to quanitify the effectiveness of moderators.
func (m *Management) Score() float64 {
	// Calculate infractions relative to messages.
	score := float64(m.Messages)
	infractions := float64(m.Bans + m.Timeouts)

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
