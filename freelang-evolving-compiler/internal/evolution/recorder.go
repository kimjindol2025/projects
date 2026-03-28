// Package evolution tracks compiler evolution and performance
package evolution

import (
	"fmt"
	"time"
)

// EvolutionMetrics captures a single build's metrics
type EvolutionMetrics struct {
	BuildID      string
	Timestamp    time.Time
	BuildTimeNs  int64
	OptsPassed   int    // number of optimization rules that modified code
	OptsApplied  []string
	CodeSizeBy   int // output code size in bytes
	SourceHash   string
}

// EvolutionRecorder accumulates build metrics over time
type EvolutionRecorder struct {
	metrics []EvolutionMetrics
	startID int
}

// NewEvolutionRecorder creates a new recorder
func NewEvolutionRecorder() *EvolutionRecorder {
	return &EvolutionRecorder{
		metrics: make([]EvolutionMetrics, 0),
		startID: 1,
	}
}

// RecordBuild adds a new build record
func (er *EvolutionRecorder) RecordBuild(buildTime int64, optsApplied []string,
	codeSize int, sourceHash string) EvolutionMetrics {

	metric := EvolutionMetrics{
		BuildID:     fmt.Sprintf("build_%d", len(er.metrics)+er.startID),
		Timestamp:   time.Now(),
		BuildTimeNs: buildTime,
		OptsPassed:  len(optsApplied),
		OptsApplied: optsApplied,
		CodeSizeBy:  codeSize,
		SourceHash:  sourceHash,
	}

	er.metrics = append(er.metrics, metric)
	return metric
}

// GetMetrics returns all recorded metrics
func (er *EvolutionRecorder) GetMetrics() []EvolutionMetrics {
	return er.metrics
}

// GetLastBuild returns the most recent build, or nil if none recorded
func (er *EvolutionRecorder) GetLastBuild() *EvolutionMetrics {
	if len(er.metrics) == 0 {
		return nil
	}
	m := er.metrics[len(er.metrics)-1]
	return &m
}

// GetBuildCount returns the number of recorded builds
func (er *EvolutionRecorder) GetBuildCount() int {
	return len(er.metrics)
}

// AverageBuildTime returns mean build duration in nanoseconds
func (er *EvolutionRecorder) AverageBuildTime() int64 {
	if len(er.metrics) == 0 {
		return 0
	}

	total := int64(0)
	for _, m := range er.metrics {
		total += m.BuildTimeNs
	}
	return total / int64(len(er.metrics))
}

// AverageOptimizationsApplied returns mean number of optimizations per build
func (er *EvolutionRecorder) AverageOptimizationsApplied() float64 {
	if len(er.metrics) == 0 {
		return 0
	}

	total := 0
	for _, m := range er.metrics {
		total += m.OptsPassed
	}
	return float64(total) / float64(len(er.metrics))
}

// GetOptimizationFrequency returns how many times each optimization was applied
func (er *EvolutionRecorder) GetOptimizationFrequency() map[string]int {
	freq := make(map[string]int)
	for _, m := range er.metrics {
		for _, opt := range m.OptsApplied {
			freq[opt]++
		}
	}
	return freq
}

// GetBuildsSince returns builds recorded after a specific time
func (er *EvolutionRecorder) GetBuildsSince(t time.Time) []EvolutionMetrics {
	var result []EvolutionMetrics
	for _, m := range er.metrics {
		if m.Timestamp.After(t) {
			result = append(result, m)
		}
	}
	return result
}

// LatestBuildTime returns the most recent build duration
func (er *EvolutionRecorder) LatestBuildTime() int64 {
	if len(er.metrics) == 0 {
		return 0
	}
	return er.metrics[len(er.metrics)-1].BuildTimeNs
}

// Summary returns a human-readable summary
func (er *EvolutionRecorder) Summary() string {
	if len(er.metrics) == 0 {
		return "No builds recorded"
	}

	return fmt.Sprintf(
		"Evolution: %d builds, avg %.0fus, %d optimizations/build avg, latest %dns",
		len(er.metrics),
		float64(er.AverageBuildTime())/1000,
		int(er.AverageOptimizationsApplied()),
		er.LatestBuildTime(),
	)
}
