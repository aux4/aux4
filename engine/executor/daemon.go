package executor

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"aux4.dev/aux4/core"
	"aux4.dev/aux4/daemon"
	"aux4.dev/aux4/engine"
	"aux4.dev/aux4/engine/param"
	"aux4.dev/aux4/output"
)

type Aux4DaemonStartExecutor struct{}

func (executor *Aux4DaemonStartExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	socketPath := daemon.FindSocketPath(".")

	if daemon.Ping(socketPath) {
		output.Out(output.StdOut).Println("daemon is already running at", socketPath)
		return nil
	}

	execPath, err := os.Executable()
	if err != nil {
		return core.InternalError("failed to find aux4 executable", err)
	}

	cmd := exec.Command(execPath, "-daemon-server", socketPath)
	setDetachedProcess(cmd)

	logPath := socketPath + ".log"
	logFile, err := os.Create(logPath)
	if err != nil {
		return core.InternalError("failed to create log file", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Dir, _ = os.Getwd()

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return core.InternalError("failed to start daemon", err)
	}

	pid := cmd.Process.Pid
	cmd.Process.Release()
	logFile.Close()

	output.Out(output.StdOut).Println("daemon started (pid:", pid, ")")
	output.Out(output.StdOut).Println("socket:", socketPath)
	output.Out(output.StdOut).Println("log:", logPath)

	return nil
}

type Aux4DaemonStopExecutor struct{}

func (executor *Aux4DaemonStopExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	socketPath := daemon.FindSocketPath(".")

	err := daemon.Shutdown(socketPath)
	if err != nil {
		pidPath := socketPath + ".pid"
		data, readErr := os.ReadFile(pidPath)
		if readErr == nil {
			pid, parseErr := strconv.Atoi(string(data))
			if parseErr == nil {
				process, findErr := os.FindProcess(pid)
				if findErr == nil {
					process.Signal(syscall.SIGTERM)
					output.Out(output.StdOut).Println("daemon stopped (pid:", pid, ")")
					os.Remove(socketPath)
					os.Remove(pidPath)
					return nil
				}
			}
		}

		output.Out(output.StdErr).Println("daemon is not running")
		return nil
	}

	return nil
}

type Aux4DaemonStatusExecutor struct{}

func (executor *Aux4DaemonStatusExecutor) Execute(env *engine.VirtualEnvironment, command core.Command, actions []string, params *param.Parameters) error {
	socketPath := daemon.FindSocketPath(".")

	if daemon.Ping(socketPath) {
		pidPath := socketPath + ".pid"
		data, err := os.ReadFile(pidPath)
		pid := "unknown"
		if err == nil {
			pid = string(data)
		}
		output.Out(output.StdOut).Println("daemon is running")
		output.Out(output.StdOut).Println("  pid:", pid)
		output.Out(output.StdOut).Println("  socket:", socketPath)
	} else {
		output.Out(output.StdOut).Println("daemon is not running")
	}

	return nil
}
