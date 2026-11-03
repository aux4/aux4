//go:build !windows

package cmd

func resolveShell(shell string) (string, error) {
	if shell == "" {
		return "/bin/sh", nil
	}
	return shell, nil
}

func resolveBash() (string, error) {
	return "bash", nil
}
