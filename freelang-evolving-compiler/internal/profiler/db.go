// Package profiler implements pattern database persistence
package profiler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// BuildRecord stores metrics from a single build execution
type BuildRecord struct {
	ID         string    `json:"id"`
	Timestamp  time.Time `json:"timestamp"`
	SourceHash string    `json:"source_hash"`
	BuildTimeNs int64    `json:"build_time_ns"`
	PatternsHit []string `json:"patterns_hit"`
	OutputSize int      `json:"output_size"`
}

// PatternEntry is persisted pattern data
type PatternEntry struct {
	Kind      string `json:"kind"`
	Signature string `json:"signature"`
	Count     int64  `json:"count"`
	SavedNs   int64  `json:"saved_ns"`
}

// Database stores patterns and build history
type Database struct {
	Patterns     []PatternEntry `json:"patterns"`
	TotalBuilds  int            `json:"total_builds"`
	LastUpdated  time.Time      `json:"last_updated"`
	BuildHistory []BuildRecord  `json:"build_history"`
}

// NewDatabase creates an empty database
func NewDatabase() *Database {
	return &Database{
		Patterns:     []PatternEntry{},
		BuildHistory: []BuildRecord{},
		LastUpdated:  time.Now(),
	}
}

// LoadFromFile loads database from JSON file
func LoadFromFile(path string) (*Database, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewDatabase(), nil
		}
		return nil, err
	}

	var db Database
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

// SaveToFile persists database to JSON file
func (db *Database) SaveToFile(path string) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

// UpdateFromCollector merges collector results into database
func (db *Database) UpdateFromCollector(collector *Collector, sourceCode string) {
	// Create unique ID for this build
	buildID := fmt.Sprintf("build_%d", db.TotalBuilds+1)

	// Calculate source hash
	hash := sha256.Sum256([]byte(sourceCode))
	sourceHash := fmt.Sprintf("%x", hash)

	// Create build record
	record := BuildRecord{
		ID:          buildID,
		Timestamp:   time.Now(),
		SourceHash:  sourceHash,
		BuildTimeNs: collector.GetBuildTimeNs(),
		PatternsHit: []string{},
		OutputSize:  0,
	}

	// Update patterns
	newPatterns := collector.GetPatterns()
	for _, newP := range newPatterns {
		found := false
		for i, existing := range db.Patterns {
			if existing.Signature == newP.Signature {
				db.Patterns[i].Count += newP.Count
				db.Patterns[i].SavedNs += newP.SavedNs
				record.PatternsHit = append(record.PatternsHit, newP.Signature)
				found = true
				break
			}
		}

		if !found {
			db.Patterns = append(db.Patterns, PatternEntry{
				Kind:      newP.Kind.String(),
				Signature: newP.Signature,
				Count:     newP.Count,
				SavedNs:   newP.SavedNs,
			})
			record.PatternsHit = append(record.PatternsHit, newP.Signature)
		}
	}

	// Update metadata
	db.TotalBuilds++
	db.LastUpdated = time.Now()
	db.BuildHistory = append(db.BuildHistory, record)

	// Keep only last 100 builds to avoid unbounded growth
	if len(db.BuildHistory) > 100 {
		db.BuildHistory = db.BuildHistory[len(db.BuildHistory)-100:]
	}
}

// TopPatterns returns the top N patterns by frequency
func (db *Database) TopPatterns(n int) []PatternEntry {
	patterns := make([]PatternEntry, len(db.Patterns))
	copy(patterns, db.Patterns)

	// Sort by count descending
	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].Count > patterns[i].Count {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	if n > len(patterns) {
		n = len(patterns)
	}
	return patterns[:n]
}

// AverageBuildTime returns the average build time in nanoseconds
func (db *Database) AverageBuildTime() int64 {
	if len(db.BuildHistory) == 0 {
		return 0
	}

	total := int64(0)
	for _, record := range db.BuildHistory {
		total += record.BuildTimeNs
	}
	return total / int64(len(db.BuildHistory))
}

// LatestBuildTime returns the most recent build time
func (db *Database) LatestBuildTime() int64 {
	if len(db.BuildHistory) == 0 {
		return 0
	}
	return db.BuildHistory[len(db.BuildHistory)-1].BuildTimeNs
}

// DetectRegression detects if latest build is significantly slower than average
func (db *Database) DetectRegression(threshold float64) (bool, float64) {
	if len(db.BuildHistory) < 5 {
		return false, 0
	}

	avg := db.AverageBuildTime()
	latest := db.LatestBuildTime()

	if avg == 0 {
		return false, 0
	}

	ratio := float64(latest) / float64(avg)
	if ratio > threshold {
		return true, ratio
	}
	return false, ratio
}

// GetPatternStats returns detailed statistics for a specific pattern
func (db *Database) GetPatternStats(signature string) (PatternEntry, bool) {
	for _, p := range db.Patterns {
		if p.Signature == signature {
			return p, true
		}
	}
	return PatternEntry{}, false
}

// Summary returns a human-readable summary of database state
func (db *Database) Summary() string {
	return fmt.Sprintf("Database: %d patterns learned, %d builds recorded, last updated %s",
		len(db.Patterns),
		db.TotalBuilds,
		db.LastUpdated.Format(time.RFC3339))
}
