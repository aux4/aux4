package coverage

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"
)

var (
	collector *Collector
	once      sync.Once
)

func IsEnabled() bool {
	return os.Getenv("AUX4_COVERAGE_FILE") != ""
}

func ensureCollector() *Collector {
	once.Do(func() {
		collector = &Collector{
			steps: make(map[string]*StepHit),
		}
	})
	return collector
}

type StepHit struct {
	Package    string            `json:"package"`
	Profile    string            `json:"profile"`
	Command    string            `json:"command"`
	Index      int               `json:"index"`
	Step       string            `json:"step"`
	Hits       int               `json:"hits"`
	Durations  []float64         `json:"durations"`
	Iterations []IterationRecord `json:"iterations,omitempty"`
	Branches   map[string]int    `json:"branches,omitempty"`
}

type IterationRecord struct {
	Count     int       `json:"count"`
	Durations []float64 `json:"durations"`
}

type CoverageReport struct {
	Timestamp string     `json:"timestamp"`
	Steps     []*StepHit `json:"steps"`
}

type Collector struct {
	mu    sync.Mutex
	steps map[string]*StepHit
	order []string
}

func stepKey(pkg, profile, command string, index int) string {
	buf := make([]byte, 0, len(pkg)+len(profile)+len(command)+10)
	buf = append(buf, pkg...)
	buf = append(buf, '|')
	buf = append(buf, profile...)
	buf = append(buf, '|')
	buf = append(buf, command...)
	buf = append(buf, '|')
	buf = appendInt(buf, index)
	return string(buf)
}

func appendInt(buf []byte, n int) []byte {
	if n < 10 {
		return append(buf, byte('0'+n))
	}
	digits := make([]byte, 0, 4)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return append(buf, digits...)
}

func getOrCreate(c *Collector, pkg, profile, command string, index int, step string) *StepHit {
	key := stepKey(pkg, profile, command, index)
	hit, exists := c.steps[key]
	if !exists {
		hit = &StepHit{
			Package: pkg,
			Profile: profile,
			Command: command,
			Index:   index,
			Step:    step,
		}
		c.steps[key] = hit
		c.order = append(c.order, key)
	}
	return hit
}

func RecordStep(pkg, profile, command string, index int, step string, duration time.Duration) {
	defer recover()

	c := ensureCollector()
	c.mu.Lock()
	defer c.mu.Unlock()

	hit := getOrCreate(c, pkg, profile, command, index, step)
	hit.Hits++
	hit.Durations = append(hit.Durations, float64(duration.Milliseconds()))
}

func RecordBranch(pkg, profile, command string, index int, step string, branch string, duration time.Duration) {
	defer recover()

	c := ensureCollector()
	c.mu.Lock()
	defer c.mu.Unlock()

	hit := getOrCreate(c, pkg, profile, command, index, step)
	hit.Hits++
	hit.Durations = append(hit.Durations, float64(duration.Milliseconds()))
	if hit.Branches == nil {
		hit.Branches = make(map[string]int)
	}
	hit.Branches[branch]++
}

func RecordIteration(pkg, profile, command string, index int, step string, totalDuration time.Duration, iterationCount int, iterationDurations []time.Duration) {
	defer recover()

	c := ensureCollector()
	c.mu.Lock()
	defer c.mu.Unlock()

	hit := getOrCreate(c, pkg, profile, command, index, step)
	hit.Hits++
	hit.Durations = append(hit.Durations, float64(totalDuration.Milliseconds()))

	iterDurs := make([]float64, len(iterationDurations))
	for i, d := range iterationDurations {
		iterDurs[i] = float64(d.Milliseconds())
	}
	hit.Iterations = append(hit.Iterations, IterationRecord{
		Count:     iterationCount,
		Durations: iterDurs,
	})
}

func Flush() {
	defer recover()

	filePath := os.Getenv("AUX4_COVERAGE_FILE")
	if filePath == "" || collector == nil {
		return
	}
	collector.mu.Lock()
	defer collector.mu.Unlock()

	if len(collector.order) == 0 {
		return
	}

	report := CoverageReport{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Steps:     make([]*StepHit, 0, len(collector.order)),
	}

	for _, key := range collector.order {
		report.Steps = append(report.Steps, collector.steps[key])
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return
	}

	// Lock the file to prevent concurrent processes from losing each other's data
	lockPath := filePath + ".lock"
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX)
		defer func() {
			syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
			lockFile.Close()
		}()
	}

	// Merge with existing file (supports multiple process invocations)
	existing, readErr := os.ReadFile(filePath)
	if readErr == nil && len(existing) > 0 {
		var existingReport CoverageReport
		if json.Unmarshal(existing, &existingReport) == nil {
			merged := mergeReports(existingReport, report)
			data, _ = json.MarshalIndent(merged, "", "  ")
		}
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "[coverage] failed to write %s: %v\n", filePath, err)
	}
}

func mergeReports(existing, incoming CoverageReport) CoverageReport {
	merged := CoverageReport{
		Timestamp: incoming.Timestamp,
		Steps:     make([]*StepHit, 0),
	}

	index := make(map[string]int)
	for i, step := range existing.Steps {
		key := stepKey(step.Package, step.Profile, step.Command, step.Index)
		index[key] = i
		merged.Steps = append(merged.Steps, step)
	}

	for _, step := range incoming.Steps {
		key := stepKey(step.Package, step.Profile, step.Command, step.Index)
		if idx, exists := index[key]; exists {
			target := merged.Steps[idx]
			target.Hits += step.Hits
			target.Durations = append(target.Durations, step.Durations...)
			target.Iterations = append(target.Iterations, step.Iterations...)
			if step.Branches != nil {
				if target.Branches == nil {
					target.Branches = make(map[string]int)
				}
				for k, v := range step.Branches {
					target.Branches[k] += v
				}
			}
		} else {
			merged.Steps = append(merged.Steps, step)
		}
	}

	return merged
}
