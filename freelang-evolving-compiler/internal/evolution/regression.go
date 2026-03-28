// Package evolution detects performance regressions
package evolution

import (
	"fmt"
	"math"
)

// RegressionAlert indicates a detected performance problem
type RegressionAlert struct {
	Message      string
	AvgNs        int64
	LatestNs     int64
	DegradePct   float64
	PreviousNs   int64
	Severity     string // "low", "medium", "high"
}

// RegressionDetector analyzes build history for performance issues
type RegressionDetector struct {
	recorder *EvolutionRecorder
}

// NewRegressionDetector creates a detector for an evolution recorder
func NewRegressionDetector(recorder *EvolutionRecorder) *RegressionDetector {
	return &RegressionDetector{
		recorder: recorder,
	}
}

// DetectRegression checks if latest build is significantly slower
func (rd *RegressionDetector) DetectRegression(threshold float64) *RegressionAlert {
	metrics := rd.recorder.GetMetrics()
	if len(metrics) < 5 {
		return nil // Need at least 5 builds for baseline
	}

	// Calculate average of all but last build (baseline)
	var baselineSum int64
	for i := 0; i < len(metrics)-1; i++ {
		baselineSum += metrics[i].BuildTimeNs
	}
	baselineAvg := baselineSum / int64(len(metrics)-1)

	latest := metrics[len(metrics)-1].BuildTimeNs

	if baselineAvg == 0 {
		return nil
	}

	ratio := float64(latest) / float64(baselineAvg)

	if ratio > threshold {
		degradePct := (ratio - 1.0) * 100
		return &RegressionAlert{
			Message:    fmt.Sprintf("Build time regression: %.2fx slower", ratio),
			AvgNs:      baselineAvg,
			LatestNs:   latest,
			DegradePct: degradePct,
			Severity:   rd.calculateSeverity(ratio),
		}
	}

	return nil
}

// DetectTrendRegression detects if performance is gradually degrading
func (rd *RegressionDetector) DetectTrendRegression(windowSize int) *RegressionAlert {
	metrics := rd.recorder.GetMetrics()
	if len(metrics) < windowSize*2 {
		return nil
	}

	// Compare two windows: recent vs older
	recentStart := len(metrics) - windowSize
	olderStart := recentStart - windowSize

	// Calculate average of older window
	var olderSum int64
	for i := olderStart; i < recentStart; i++ {
		olderSum += metrics[i].BuildTimeNs
	}
	olderAvg := olderSum / int64(windowSize)

	// Calculate average of recent window
	var recentSum int64
	for i := recentStart; i < len(metrics); i++ {
		recentSum += metrics[i].BuildTimeNs
	}
	recentAvg := recentSum / int64(windowSize)

	if olderAvg == 0 {
		return nil
	}

	ratio := float64(recentAvg) / float64(olderAvg)

	// Trend threshold is lower (catches gradual degradation)
	const trendThreshold = 1.1 // 10% degradation

	if ratio > trendThreshold {
		degradePct := (ratio - 1.0) * 100
		return &RegressionAlert{
			Message:    fmt.Sprintf("Trending regression: %.2fx slower", ratio),
			AvgNs:      olderAvg,
			LatestNs:   recentAvg,
			DegradePct: degradePct,
			PreviousNs: olderAvg,
			Severity:   rd.calculateSeverity(ratio),
		}
	}

	return nil
}

// DetectOutlier finds unexpectedly slow individual builds
func (rd *RegressionDetector) DetectOutlier(stdDevs float64) *RegressionAlert {
	metrics := rd.recorder.GetMetrics()
	if len(metrics) < 3 {
		return nil
	}

	// Calculate mean and std dev
	var sum int64
	for _, m := range metrics {
		sum += m.BuildTimeNs
	}
	mean := float64(sum) / float64(len(metrics))

	var sumSq float64
	for _, m := range metrics {
		diff := float64(m.BuildTimeNs) - mean
		sumSq += diff * diff
	}
	variance := sumSq / float64(len(metrics))
	stdDev := math.Sqrt(variance)

	latest := float64(metrics[len(metrics)-1].BuildTimeNs)
	zScore := (latest - mean) / stdDev

	if zScore > stdDevs {
		return &RegressionAlert{
			Message:    fmt.Sprintf("Outlier detected: %.1f std devs above mean", zScore),
			AvgNs:      int64(mean),
			LatestNs:   int64(latest),
			DegradePct: ((latest - mean) / mean) * 100,
			Severity:   "medium",
		}
	}

	return nil
}

// calculateSeverity determines alert level based on degradation ratio
func (rd *RegressionDetector) calculateSeverity(ratio float64) string {
	switch {
	case ratio > 2.0:
		return "high"
	case ratio > 1.5:
		return "medium"
	default:
		return "low"
	}
}

// GetHealthStatus returns overall performance health
func (rd *RegressionDetector) GetHealthStatus() string {
	metrics := rd.recorder.GetMetrics()
	if len(metrics) == 0 {
		return "unknown"
	}

	// Check for any regression
	if rd.DetectRegression(1.2) != nil {
		return "degraded"
	}

	// Check for trends
	if rd.DetectTrendRegression(3) != nil {
		return "degrading"
	}

	// Check for outliers
	if rd.DetectOutlier(2.0) != nil {
		return "unstable"
	}

	return "healthy"
}

// AnalyzeFull performs comprehensive regression analysis
func (rd *RegressionDetector) AnalyzeFull() []RegressionAlert {
	var alerts []RegressionAlert

	// Check for absolute regression
	if alert := rd.DetectRegression(1.2); alert != nil {
		alerts = append(alerts, *alert)
	}

	// Check for trending regression
	if alert := rd.DetectTrendRegression(3); alert != nil {
		alerts = append(alerts, *alert)
	}

	// Check for outliers
	if alert := rd.DetectOutlier(2.0); alert != nil {
		alerts = append(alerts, *alert)
	}

	return alerts
}
