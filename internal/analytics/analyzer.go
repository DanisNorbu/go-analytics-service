package analytics

import (
	"math"
	"sync"
	"time"
)

type Metric struct {
	Timestamp time.Time `json:"timestamp"`
	CPU       float64   `json:"cpu"`
	RPS       float64   `json:"rps"`
}

type Result struct {
	Count          int       `json:"count"`
	AvgCPU         float64   `json:"avg_cpu"`
	AvgRPS         float64   `json:"avg_rps"`
	LastTimestamp  time.Time `json:"last_timestamp"`
	LastRPS        float64   `json:"last_rps"`
	IsAnomalyRPS   bool      `json:"is_anomaly_rps"`
	ZScoreRPS      float64   `json:"zscore_rps"`
	WindowSizeUsed int       `json:"window_size_used"`
}

type Analyzer struct {
	mu      sync.RWMutex
	window  []Metric
	maxSize int
}

func NewAnalyzer(size int) *Analyzer {
	if size <= 0 {
		size = 50
	}
	return &Analyzer{
		window:  make([]Metric, 0, size),
		maxSize: size,
	}
}

func (a *Analyzer) AddMetric(m Metric) Result {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.window = append(a.window, m)
	if len(a.window) > a.maxSize {
		a.window = a.window[1:]
	}

	return a.computeLocked()
}

func (a *Analyzer) GetLastResult() Result {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.computeLocked()
}

func (a *Analyzer) computeLocked() Result {
	n := len(a.window)
	if n == 0 {
		return Result{}
	}

	var sumCPU, sumRPS float64
	for _, m := range a.window {
		sumCPU += m.CPU
		sumRPS += m.RPS
	}
	avgCPU := sumCPU / float64(n)
	avgRPS := sumRPS / float64(n)

	var variance float64
	for _, m := range a.window {
		d := m.RPS - avgRPS
		variance += d * d
	}
	variance /= float64(n)
	std := math.Sqrt(variance)

	last := a.window[n-1]
	z := 0.0
	isAnomaly := false
	if std > 0 {
		z = (last.RPS - avgRPS) / std
		if math.Abs(z) > 2.0 {
			isAnomaly = true
		}
	}

	return Result{
		Count:          n,
		AvgCPU:         avgCPU,
		AvgRPS:         avgRPS,
		LastTimestamp:  last.Timestamp,
		LastRPS:        last.RPS,
		IsAnomalyRPS:   isAnomaly,
		ZScoreRPS:      z,
		WindowSizeUsed: n,
	}
}
