package core

// MovingAverage contains logic related to a moving average.
type MovingAverage struct {
	EMAs   []float64
	Period float64
	SMAs   []float64
	Values []float64
}

// NewMovingAverage returns an initialized MovingAverage for storing values.
func NewMovingAverage(period float64) *MovingAverage {
	return &MovingAverage{
		EMAs:   make([]float64, 0),
		Period: period,
		SMAs:   make([]float64, 0),
		Values: make([]float64, 0),
	}
}

// Multiplier returns the multiplier used for EMA.
func Multiplier(period float64) float64 {
	return 2 / (period + 1)
}

// Append adds to the slices of values.
func (ma *MovingAverage) Append(value float64) {
	ma.Values = append(ma.Values, value)
}

// EMA calculates the exponential moving average based on the given period.
func (ma *MovingAverage) EMA() float64 {
	var prev float64
	length := len(ma.EMAs)

	if length == 0 {
		prev = ma.SMAs[len(ma.SMAs)-1]
	} else {
		prev = ma.EMAs[length-1]
	}

	closing := ma.Values[len(ma.Values)]
	ema := ((closing - prev) * Multiplier(ma.Period)) + prev
	ma.EMAs = append(ma.EMAs, ema)
	return ema
}

// SMA calculates the simple moving average based on the given period.
func (ma *MovingAverage) SMA() float64 {
	var sum float64
	start := len(ma.Values)

	if start > int(ma.Period) {
		start -= int(ma.Period)
	}

	for _, value := range ma.Values[start:] {
		sum += value
	}

	sma := sum / float64(ma.Period)
	ma.SMAs = append(ma.SMAs, sma)
	return sma
}

// Update updates the moving averages for both SMA and EMA.
func (ma *MovingAverage) Update() (float64, float64) {
	return ma.SMA(), ma.EMA()
}
