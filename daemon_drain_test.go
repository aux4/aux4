package main

import (
	"testing"
	"time"
)

func TestWaitForOutputDrainWaitsForEveryStream(t *testing.T) {
	stdout := make(chan struct{})
	stderr := make(chan struct{})
	close(stderr) // stderr finishes first, stdout is still copying

	go func() {
		time.Sleep(50 * time.Millisecond)
		close(stdout)
	}()

	start := time.Now()
	if !waitForOutputDrain(time.Second, stdout, stderr) {
		t.Fatal("expected both streams to drain")
	}
	if time.Since(start) < 40*time.Millisecond {
		t.Fatal("returned before stdout finished draining")
	}
}

func TestWaitForOutputDrainGivesUpAfterGrace(t *testing.T) {
	held := make(chan struct{}) // e.g. a background child keeps the pipe open
	done := make(chan struct{})
	close(done)

	start := time.Now()
	if waitForOutputDrain(30*time.Millisecond, done, held) {
		t.Fatal("expected a timeout when a stream never drains")
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("grace not honoured, waited %v", elapsed)
	}
}
