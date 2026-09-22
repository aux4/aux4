package main

import (
	"strings"
	"testing"
)

// swapRunCommandForTest substitutes runCommandFn (the package-level
// indirection run() calls) with a test stand-in, and returns a function that
// restores the original. Used to exercise run()'s recover() wiring without a
// real, hard-to-trigger internal panic.
func swapRunCommandForTest(fn func() int) func() {
	original := runCommandFn
	runCommandFn = fn
	return func() {
		runCommandFn = original
	}
}

// TestFormatCrashReport_WithoutDebug ensures the default (AUX4_DEBUG unset)
// message is actionable — names the panic value, says it's an aux4 bug (not
// the user's command), points at where to report it, and tells the user how
// to get the full trace — but does NOT include the raw Go stack trace.
func TestFormatCrashReport_WithoutDebug(t *testing.T) {
	stack := []byte("goroutine 1 [running]:\nmain.someInternalFunction(...)\n")
	message := formatCrashReport("index out of range [3] with length 2", stack, false)

	if !strings.Contains(message, "index out of range [3] with length 2") {
		t.Errorf("message does not include the panic value: %q", message)
	}
	if !strings.Contains(message, "bug in aux4") {
		t.Errorf("message does not tell the user this is an aux4 bug: %q", message)
	}
	if !strings.Contains(message, "github.com/aux4/aux4/issues") {
		t.Errorf("message does not point at where to report the bug: %q", message)
	}
	if !strings.Contains(message, "AUX4_DEBUG") {
		t.Errorf("message does not mention AUX4_DEBUG for the full trace: %q", message)
	}
	if strings.Contains(message, "goroutine 1 [running]") {
		t.Errorf("message leaked the raw stack trace without AUX4_DEBUG=true: %q", message)
	}
}

// TestFormatCrashReport_WithDebug ensures AUX4_DEBUG=true preserves full
// diagnosability — the raw stack trace must be present.
func TestFormatCrashReport_WithDebug(t *testing.T) {
	stack := []byte("goroutine 1 [running]:\nmain.someInternalFunction(...)\n")
	message := formatCrashReport("nil pointer dereference", stack, true)

	if !strings.Contains(message, "goroutine 1 [running]") {
		t.Errorf("message did not include the stack trace with AUX4_DEBUG=true: %q", message)
	}
	if !strings.Contains(message, "nil pointer dereference") {
		t.Errorf("message does not include the panic value: %q", message)
	}
}

// TestReportCrash_ExitCodeIsNonZero ensures the backstop always reports a
// non-zero exit code — a caught panic must never look like success.
func TestReportCrash_ExitCodeIsNonZero(t *testing.T) {
	code := reportCrash("boom", []byte("stack"))
	if code == 0 {
		t.Fatalf("reportCrash returned exit code 0, want non-zero")
	}
}

// TestRun_RecoversFromPanicAndReturnsNonZero exercises the actual defer/
// recover wiring in run(): if runCommand panics, run() must not propagate the
// panic to its caller (main), and must return a non-zero exit code.
func TestRun_RecoversFromPanicAndReturnsNonZero(t *testing.T) {
	restore := swapRunCommandForTest(func() int {
		panic("simulated crash for CORE-036/CORE-037 backstop test")
	})
	defer restore()

	var exitCode int
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("run() let a panic escape instead of recovering: %v", r)
			}
		}()
		exitCode = run()
	}()

	if exitCode == 0 {
		t.Errorf("run() returned exit code 0 after recovering a panic, want non-zero")
	}
}

// TestRun_NormalPathIsUnaffected ensures the recover() wiring is transparent
// when nothing panics — the real exit code from runCommand must still flow
// through untouched.
func TestRun_NormalPathIsUnaffected(t *testing.T) {
	restore := swapRunCommandForTest(func() int {
		return 42
	})
	defer restore()

	if got := run(); got != 42 {
		t.Errorf("run() = %d, want 42", got)
	}
}
