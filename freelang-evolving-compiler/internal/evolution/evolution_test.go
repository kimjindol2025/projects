package evolution

import (
	"testing"
	"time"
)

// TestRecorderCreation validates initialization
func TestRecorderCreation(t *testing.T) {
	rec := NewEvolutionRecorder()
	if rec == nil {
		t.Errorf("failed to create recorder")
	}
	if rec.GetBuildCount() != 0 {
		t.Errorf("expected 0 builds, got %d", rec.GetBuildCount())
	}
}

// TestRecordBuild validates recording a single build
func TestRecordBuild(t *testing.T) {
	rec := NewEvolutionRecorder()

	metric := rec.RecordBuild(1000000, []string{"ConstantFolding"}, 512, "abc123")

	if metric.BuildID == "" {
		t.Errorf("expected build ID")
	}
	if metric.BuildTimeNs != 1000000 {
		t.Errorf("expected 1000000ns, got %d", metric.BuildTimeNs)
	}
	if metric.OptsPassed != 1 {
		t.Errorf("expected 1 opt, got %d", metric.OptsPassed)
	}
	if metric.CodeSizeBy != 512 {
		t.Errorf("expected 512 bytes, got %d", metric.CodeSizeBy)
	}
}

// TestMultipleBuilds validates accumulation
func TestMultipleBuilds(t *testing.T) {
	rec := NewEvolutionRecorder()

	rec.RecordBuild(1000000, []string{"ConstantFolding"}, 512, "abc")
	rec.RecordBuild(900000, []string{"ConstantFolding"}, 480, "def")
	rec.RecordBuild(950000, []string{"ConstantFolding", "DeadCodeElimination"}, 450, "ghi")

	if rec.GetBuildCount() != 3 {
		t.Errorf("expected 3 builds, got %d", rec.GetBuildCount())
	}
}

// TestGetLastBuild validates latest build retrieval
func TestGetLastBuild(t *testing.T) {
	rec := NewEvolutionRecorder()
	rec.RecordBuild(1000000, []string{}, 512, "first")
	rec.RecordBuild(2000000, []string{}, 512, "second")

	last := rec.GetLastBuild()
	if last == nil {
		t.Errorf("expected last build, got nil")
	}
	if last.BuildTimeNs != 2000000 {
		t.Errorf("expected 2000000ns, got %d", last.BuildTimeNs)
	}
}

// TestAverageBuildTime validates computation
func TestAverageBuildTime(t *testing.T) {
	rec := NewEvolutionRecorder()
	rec.RecordBuild(1000000, []string{}, 0, "")
	rec.RecordBuild(2000000, []string{}, 0, "")
	rec.RecordBuild(3000000, []string{}, 0, "")

	avg := rec.AverageBuildTime()
	expected := int64(2000000)
	if avg != expected {
		t.Errorf("expected average %d, got %d", expected, avg)
	}
}

// TestAverageOptimizationsApplied validates count
func TestAverageOptimizationsApplied(t *testing.T) {
	rec := NewEvolutionRecorder()
	rec.RecordBuild(1000000, []string{"A"}, 0, "")
	rec.RecordBuild(1000000, []string{"A", "B"}, 0, "")
	rec.RecordBuild(1000000, []string{}, 0, "")

	avg := rec.AverageOptimizationsApplied()
	expected := 1.0 // (1 + 2 + 0) / 3
	if avg != expected {
		t.Errorf("expected average %.1f, got %.1f", expected, avg)
	}
}

// TestGetOptimizationFrequency validates frequency map
func TestGetOptimizationFrequency(t *testing.T) {
	rec := NewEvolutionRecorder()
	rec.RecordBuild(1000000, []string{"A", "B"}, 0, "")
	rec.RecordBuild(1000000, []string{"A"}, 0, "")
	rec.RecordBuild(1000000, []string{"B", "C"}, 0, "")

	freq := rec.GetOptimizationFrequency()

	if freq["A"] != 2 {
		t.Errorf("expected A=2, got %d", freq["A"])
	}
	if freq["B"] != 2 {
		t.Errorf("expected B=2, got %d", freq["B"])
	}
	if freq["C"] != 1 {
		t.Errorf("expected C=1, got %d", freq["C"])
	}
}

// TestGetBuildsSince validates time filtering
func TestGetBuildsSince(t *testing.T) {
	rec := NewEvolutionRecorder()

	start := time.Now()
	rec.RecordBuild(1000000, []string{}, 0, "first")
	time.Sleep(10 * time.Millisecond)
	midpoint := time.Now()
	time.Sleep(10 * time.Millisecond)
	rec.RecordBuild(1000000, []string{}, 0, "second")

	builds := rec.GetBuildsSince(midpoint)

	if len(builds) != 1 {
		t.Errorf("expected 1 build since midpoint, got %d", len(builds))
	}

	allBuilds := rec.GetBuildsSince(start.Add(-1 * time.Second))
	if len(allBuilds) != 2 {
		t.Errorf("expected 2 builds since start, got %d", len(allBuilds))
	}
}

// TestDetectionCreation validates detector initialization
func TestDetectionCreation(t *testing.T) {
	rec := NewEvolutionRecorder()
	det := NewRegressionDetector(rec)
	if det == nil {
		t.Errorf("failed to create detector")
	}
}

// TestDetectRegression validates threshold detection
func TestDetectRegression(t *testing.T) {
	rec := NewEvolutionRecorder()

	// Add baseline builds
	for i := 0; i < 5; i++ {
		rec.RecordBuild(1000000, []string{}, 0, "")
	}

	// Add regression build (2x slower)
	rec.RecordBuild(2500000, []string{}, 0, "")

	det := NewRegressionDetector(rec)
	alert := det.DetectRegression(1.5)

	if alert == nil {
		t.Errorf("expected regression alert")
	}
	if alert.DegradePct <= 0 {
		t.Errorf("expected positive degradation percentage")
	}
}

// TestNoRegressionBeforeBaseline validates minimum history requirement
func TestNoRegressionBeforeBaseline(t *testing.T) {
	rec := NewEvolutionRecorder()
	rec.RecordBuild(1000000, []string{}, 0, "")
	rec.RecordBuild(2000000, []string{}, 0, "")
	rec.RecordBuild(3000000, []string{}, 0, "")

	det := NewRegressionDetector(rec)
	alert := det.DetectRegression(1.5)

	if alert != nil {
		t.Errorf("expected no alert with insufficient history")
	}
}

// TestDetectTrendRegression validates gradual degradation
func TestDetectTrendRegression(t *testing.T) {
	rec := NewEvolutionRecorder()

	// Older window: fast builds
	rec.RecordBuild(1000000, []string{}, 0, "")
	rec.RecordBuild(1000000, []string{}, 0, "")
	rec.RecordBuild(1000000, []string{}, 0, "")

	// Recent window: slower builds
	rec.RecordBuild(1200000, []string{}, 0, "")
	rec.RecordBuild(1200000, []string{}, 0, "")
	rec.RecordBuild(1200000, []string{}, 0, "")

	det := NewRegressionDetector(rec)
	alert := det.DetectTrendRegression(3)

	if alert == nil {
		t.Errorf("expected trend regression alert")
	}
}

// TestDetectOutlier validates outlier detection
func TestDetectOutlier(t *testing.T) {
	rec := NewEvolutionRecorder()

	// Normal builds
	rec.RecordBuild(1000000, []string{}, 0, "")
	rec.RecordBuild(1000000, []string{}, 0, "")
	rec.RecordBuild(1000000, []string{}, 0, "")

	// Outlier
	rec.RecordBuild(5000000, []string{}, 0, "")

	det := NewRegressionDetector(rec)
	alert := det.DetectOutlier(1.5)

	if alert == nil {
		t.Errorf("expected outlier alert")
	}
}

// TestHealthStatusHealthy validates good performance
func TestHealthStatusHealthy(t *testing.T) {
	rec := NewEvolutionRecorder()

	// Consistent fast builds
	for i := 0; i < 10; i++ {
		rec.RecordBuild(1000000, []string{}, 0, "")
	}

	det := NewRegressionDetector(rec)
	status := det.GetHealthStatus()

	if status != "healthy" {
		t.Errorf("expected healthy status, got %s", status)
	}
}

// TestHealthStatusDegraded validates poor performance
func TestHealthStatusDegraded(t *testing.T) {
	rec := NewEvolutionRecorder()

	// Baseline
	for i := 0; i < 5; i++ {
		rec.RecordBuild(1000000, []string{}, 0, "")
	}

	// Regression
	rec.RecordBuild(3000000, []string{}, 0, "")

	det := NewRegressionDetector(rec)
	status := det.GetHealthStatus()

	if status != "degraded" {
		t.Errorf("expected degraded status, got %s", status)
	}
}

// TestAnalyzeFull validates comprehensive analysis
func TestAnalyzeFull(t *testing.T) {
	rec := NewEvolutionRecorder()

	// Multiple alert conditions
	for i := 0; i < 5; i++ {
		rec.RecordBuild(1000000, []string{}, 0, "")
	}
	rec.RecordBuild(2500000, []string{}, 0, "")

	det := NewRegressionDetector(rec)
	alerts := det.AnalyzeFull()

	if len(alerts) == 0 {
		t.Errorf("expected alerts from full analysis")
	}
}

// TestMetricsFields validates metric structure
func TestMetricsFields(t *testing.T) {
	rec := NewEvolutionRecorder()
	opts := []string{"Opt1", "Opt2"}
	metric := rec.RecordBuild(12345, opts, 9999, "hash123")

	if metric.BuildTimeNs != 12345 {
		t.Errorf("BuildTimeNs mismatch")
	}
	if metric.OptsPassed != 2 {
		t.Errorf("OptsPassed mismatch")
	}
	if len(metric.OptsApplied) != 2 {
		t.Errorf("OptsApplied length mismatch")
	}
	if metric.CodeSizeBy != 9999 {
		t.Errorf("CodeSizeBy mismatch")
	}
	if metric.SourceHash != "hash123" {
		t.Errorf("SourceHash mismatch")
	}
}

// TestLatestBuildTime validates latest retrieval
func TestLatestBuildTime(t *testing.T) {
	rec := NewEvolutionRecorder()
	rec.RecordBuild(1000000, []string{}, 0, "")
	rec.RecordBuild(2000000, []string{}, 0, "")
	rec.RecordBuild(3000000, []string{}, 0, "")

	latest := rec.LatestBuildTime()
	if latest != 3000000 {
		t.Errorf("expected 3000000, got %d", latest)
	}
}

// TestSummaryOutput validates human-readable output
func TestSummaryOutput(t *testing.T) {
	rec := NewEvolutionRecorder()
	rec.RecordBuild(1000000, []string{"A"}, 0, "")
	rec.RecordBuild(1000000, []string{}, 0, "")

	summary := rec.Summary()
	if summary == "" {
		t.Errorf("expected non-empty summary")
	}
	if len(summary) < 10 {
		t.Errorf("summary too short: %s", summary)
	}
}
