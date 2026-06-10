package coverage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func resetCollector() {
	collector = nil
	once = sync.Once{}
}

func TestIsEnabled(t *testing.T) {
	os.Unsetenv("AUX4_COVERAGE_FILE")
	if IsEnabled() {
		t.Error("expected disabled when env var is not set")
	}

	os.Setenv("AUX4_COVERAGE_FILE", "/tmp/test-cov.json")
	defer os.Unsetenv("AUX4_COVERAGE_FILE")
	if !IsEnabled() {
		t.Error("expected enabled when env var is set")
	}
}

func TestRecordStep(t *testing.T) {
	resetCollector()
	os.Setenv("AUX4_COVERAGE_FILE", "/tmp/test-cov.json")
	defer os.Unsetenv("AUX4_COVERAGE_FILE")

	RecordStep("pkg", "main", "build", 0, "echo hello", 5*time.Millisecond)
	RecordStep("pkg", "main", "build", 0, "echo hello", 10*time.Millisecond)

	c := ensureCollector()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(c.steps))
	}

	key := stepKey("pkg", "main", "build", 0)
	hit := c.steps[key]
	if hit.Hits != 2 {
		t.Errorf("expected 2 hits, got %d", hit.Hits)
	}
	if len(hit.Durations) != 2 {
		t.Errorf("expected 2 durations, got %d", len(hit.Durations))
	}
	if hit.Durations[0] != 5 || hit.Durations[1] != 10 {
		t.Errorf("unexpected durations: %v", hit.Durations)
	}
}

func TestRecordBranch(t *testing.T) {
	resetCollector()
	os.Setenv("AUX4_COVERAGE_FILE", "/tmp/test-cov.json")
	defer os.Unsetenv("AUX4_COVERAGE_FILE")

	RecordBranch("pkg", "main", "deploy", 2, "when:env=prod:backup", "true", 3*time.Millisecond)
	RecordBranch("pkg", "main", "deploy", 2, "when:env=prod:backup", "false", 1*time.Millisecond)
	RecordBranch("pkg", "main", "deploy", 2, "when:env=prod:backup", "false", 2*time.Millisecond)

	c := ensureCollector()
	c.mu.Lock()
	defer c.mu.Unlock()

	key := stepKey("pkg", "main", "deploy", 2)
	hit := c.steps[key]
	if hit.Hits != 3 {
		t.Errorf("expected 3 hits, got %d", hit.Hits)
	}
	if hit.Branches["true"] != 1 {
		t.Errorf("expected 1 true branch, got %d", hit.Branches["true"])
	}
	if hit.Branches["false"] != 2 {
		t.Errorf("expected 2 false branches, got %d", hit.Branches["false"])
	}
}

func TestRecordIteration(t *testing.T) {
	resetCollector()
	os.Setenv("AUX4_COVERAGE_FILE", "/tmp/test-cov.json")
	defer os.Unsetenv("AUX4_COVERAGE_FILE")

	iterDurs := []time.Duration{2 * time.Millisecond, 3 * time.Millisecond, 5 * time.Millisecond}
	RecordIteration("pkg", "main", "process", 1, "each:handle ${item}", 10*time.Millisecond, 3, iterDurs)

	c := ensureCollector()
	c.mu.Lock()
	defer c.mu.Unlock()

	key := stepKey("pkg", "main", "process", 1)
	hit := c.steps[key]
	if hit.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", hit.Hits)
	}
	if len(hit.Iterations) != 1 {
		t.Fatalf("expected 1 iteration record, got %d", len(hit.Iterations))
	}
	if hit.Iterations[0].Count != 3 {
		t.Errorf("expected 3 iterations, got %d", hit.Iterations[0].Count)
	}
	if len(hit.Iterations[0].Durations) != 3 {
		t.Errorf("expected 3 iteration durations, got %d", len(hit.Iterations[0].Durations))
	}
}

func TestFlushWritesFile(t *testing.T) {
	resetCollector()
	tmpFile := filepath.Join(t.TempDir(), "cov.json")
	os.Setenv("AUX4_COVERAGE_FILE", tmpFile)
	defer os.Unsetenv("AUX4_COVERAGE_FILE")

	RecordStep("pkg", "main", "build", 0, "echo hello", 5*time.Millisecond)
	RecordStep("pkg", "main", "build", 1, "echo done", 2*time.Millisecond)

	Flush()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("expected coverage file to exist: %v", err)
	}

	var report CoverageReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("expected valid JSON: %v", err)
	}
	if len(report.Steps) != 2 {
		t.Errorf("expected 2 steps in report, got %d", len(report.Steps))
	}
	if report.Timestamp == "" {
		t.Error("expected timestamp to be set")
	}
}

func TestFlushMergesWithExisting(t *testing.T) {
	resetCollector()
	tmpFile := filepath.Join(t.TempDir(), "cov.json")
	os.Setenv("AUX4_COVERAGE_FILE", tmpFile)
	defer os.Unsetenv("AUX4_COVERAGE_FILE")

	// Write first report
	RecordStep("pkg", "main", "build", 0, "echo hello", 5*time.Millisecond)
	Flush()

	// Reset and write second report — should merge
	resetCollector()
	RecordStep("pkg", "main", "build", 0, "echo hello", 10*time.Millisecond)
	RecordStep("pkg", "main", "test", 0, "go test", 100*time.Millisecond)
	Flush()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("expected coverage file: %v", err)
	}

	var report CoverageReport
	json.Unmarshal(data, &report)

	if len(report.Steps) != 2 {
		t.Fatalf("expected 2 steps after merge, got %d", len(report.Steps))
	}

	// Find the build step
	var buildStep *StepHit
	for _, s := range report.Steps {
		if s.Command == "build" {
			buildStep = s
			break
		}
	}
	if buildStep == nil {
		t.Fatal("build step not found in merged report")
	}
	if buildStep.Hits != 2 {
		t.Errorf("expected 2 merged hits, got %d", buildStep.Hits)
	}
	if len(buildStep.Durations) != 2 {
		t.Errorf("expected 2 merged durations, got %d", len(buildStep.Durations))
	}
}

func TestFlushNoopWhenDisabled(t *testing.T) {
	resetCollector()
	os.Unsetenv("AUX4_COVERAGE_FILE")

	// Should not panic
	Flush()
}

func TestFlushNoopWhenEmpty(t *testing.T) {
	resetCollector()
	tmpFile := filepath.Join(t.TempDir(), "cov.json")
	os.Setenv("AUX4_COVERAGE_FILE", tmpFile)
	defer os.Unsetenv("AUX4_COVERAGE_FILE")

	// Ensure collector exists but empty
	ensureCollector()
	Flush()

	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("expected no file to be written when there are no steps")
	}
}

func TestRecoverFromPanic(t *testing.T) {
	// Coverage should never crash the calling code
	resetCollector()
	os.Setenv("AUX4_COVERAGE_FILE", "/tmp/test-cov.json")
	defer os.Unsetenv("AUX4_COVERAGE_FILE")

	// These should all complete without panic even with empty values
	RecordStep("", "", "", 0, "", 0)
	RecordBranch("", "", "", 0, "", "", 0)
	RecordIteration("", "", "", 0, "", 0, 0, nil)
}

func TestStepKeyUniqueness(t *testing.T) {
	k1 := stepKey("pkg", "main", "build", 0)
	k2 := stepKey("pkg", "main", "build", 1)
	k3 := stepKey("pkg", "main", "test", 0)
	k4 := stepKey("other", "main", "build", 0)

	keys := map[string]bool{k1: true, k2: true, k3: true, k4: true}
	if len(keys) != 4 {
		t.Errorf("expected 4 unique keys, got %d", len(keys))
	}
}

func TestMultipleStepsPreserveOrder(t *testing.T) {
	resetCollector()
	os.Setenv("AUX4_COVERAGE_FILE", filepath.Join(t.TempDir(), "cov.json"))
	defer os.Unsetenv("AUX4_COVERAGE_FILE")

	RecordStep("pkg", "main", "cmd", 0, "step-a", time.Millisecond)
	RecordStep("pkg", "main", "cmd", 1, "step-b", time.Millisecond)
	RecordStep("pkg", "main", "cmd", 2, "step-c", time.Millisecond)
	// Re-hit step 0
	RecordStep("pkg", "main", "cmd", 0, "step-a", time.Millisecond)

	c := ensureCollector()
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.order) != 3 {
		t.Fatalf("expected 3 ordered entries, got %d", len(c.order))
	}

	hit0 := c.steps[c.order[0]]
	if hit0.Step != "step-a" {
		t.Errorf("expected first step to be step-a, got %s", hit0.Step)
	}
	if hit0.Hits != 2 {
		t.Errorf("expected step-a to have 2 hits, got %d", hit0.Hits)
	}
}
