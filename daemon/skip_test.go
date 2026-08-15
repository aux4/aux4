package daemon

import (
	"os"
	"reflect"
	"testing"
)

func TestExtractNoDaemonFlag(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		wantEnabled bool
		wantArgs    []string
	}{
		{
			name:        "absent leaves args unchanged",
			args:        []string{"aux4", "version"},
			wantEnabled: false,
			wantArgs:    []string{"aux4", "version"},
		},
		{
			name:        "bare flag before command is stripped and does not swallow command",
			args:        []string{"--noDaemon", "mcp"},
			wantEnabled: true,
			wantArgs:    []string{"mcp"},
		},
		{
			name:        "flag after command",
			args:        []string{"aux4", "version", "--noDaemon"},
			wantEnabled: true,
			wantArgs:    []string{"aux4", "version"},
		},
		{
			name:        "explicit true",
			args:        []string{"--noDaemon=true", "mcp"},
			wantEnabled: true,
			wantArgs:    []string{"mcp"},
		},
		{
			name:        "explicit false is off but still stripped",
			args:        []string{"--noDaemon=false", "mcp"},
			wantEnabled: false,
			wantArgs:    []string{"mcp"},
		},
		{
			name:        "other flags preserved",
			args:        []string{"--noDaemon", "aux4", "version", "--raw"},
			wantEnabled: true,
			wantArgs:    []string{"aux4", "version", "--raw"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			enabled, args := ExtractNoDaemonFlag(c.args)
			if enabled != c.wantEnabled {
				t.Fatalf("enabled: got %v, want %v", enabled, c.wantEnabled)
			}
			if !reflect.DeepEqual(args, c.wantArgs) {
				t.Fatalf("args: got %v, want %v", args, c.wantArgs)
			}
		})
	}
}

func TestSkipForwarding(t *testing.T) {
	original, hadEnv := os.LookupEnv(NoDaemonEnv)
	os.Unsetenv(NoDaemonEnv)
	defer func() {
		if hadEnv {
			os.Setenv(NoDaemonEnv, original)
		} else {
			os.Unsetenv(NoDaemonEnv)
		}
	}()

	t.Run("no flag and no env forwards", func(t *testing.T) {
		os.Unsetenv(NoDaemonEnv)
		if SkipForwarding(false) {
			t.Fatal("expected forwarding (SkipForwarding=false) with no flag and no env")
		}
	})

	t.Run("flag skips forwarding", func(t *testing.T) {
		os.Unsetenv(NoDaemonEnv)
		if !SkipForwarding(true) {
			t.Fatal("expected SkipForwarding=true when flag is set")
		}
	})

	t.Run("env=1 skips forwarding", func(t *testing.T) {
		os.Setenv(NoDaemonEnv, "1")
		if !SkipForwarding(false) {
			t.Fatal("expected SkipForwarding=true when AUX4_NO_DAEMON=1")
		}
	})

	t.Run("env other than 1 forwards", func(t *testing.T) {
		os.Setenv(NoDaemonEnv, "0")
		if SkipForwarding(false) {
			t.Fatal("expected forwarding when AUX4_NO_DAEMON is not exactly 1")
		}
		os.Setenv(NoDaemonEnv, "true")
		if SkipForwarding(false) {
			t.Fatal("expected forwarding when AUX4_NO_DAEMON=true (only \"1\" enables)")
		}
	})
}
