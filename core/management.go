package core

import (
	"log"
	"math"
	"time"

	watchmen "github.com/kookehs/watchmen/core"
)

// DefaultModerators is the default number of moderators.
const DefaultModerators = 3

// Management handles logic related to dynamically managing moderators.
type Management struct {
	// Management variables
	MovingAverage *MovingAverage
	Timer         *time.Timer

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
		Timer:         time.NewTimer(10 * time.Second),
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

	// TODO: Figure out why values aren't correct.
	log.Printf("[Management]: EMA - %v, SMA - %v, Signal - %v", ema, sma, m.MovingAverage.Signal)

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
	// A few infractions should not bring the score down substantially.
	adjustment := infractions / 2 * math.Log(infractions+1)
	return score / adjustment
}

// Update updates variables and resets counters.
// Update should be called as a goroutine.
func (m *Management) Update() {
	for {
		<-m.Timer.C
		m.Moderators = m.Heuristic()
		log.Printf("[Management]: Mods - %v", m.Moderators)
		m.Bans = 0
		m.Messages = 0
		m.Timeouts = 0
		m.Timer.Reset(10 * time.Second)
	}
}
