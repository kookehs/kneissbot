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
	// Minimum number of moderators based on messages incoming.
	baseline := float64(m.Messages) * m.Ratio
	m.MovingAverage.Append(m.Score())
	sma, ema := m.MovingAverage.Update()

	// TODO: Factor in the trend into number of moderators.
}

// Score quanitifies the effectiveness of moderators.
func (m *Management) Score() float64 {
	// Count infractions relative to messages.
	infractions := float64(m.Bans + m.Timeouts)
	score := float64(m.Messages)

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
