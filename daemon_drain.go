package main

import "time"

// daemonOutputDrainGrace bounds how long a daemon request waits, AFTER its
// command has finished and the pipe write ends are closed, for the stdout and
// stderr copiers to drain. Normally they finish within microseconds (at most
// one pipe buffer is still in flight). The bound only matters when a background
// process spawned by the command inherited the pipe and keeps it open — the
// request must still complete instead of hanging the client forever.
const daemonOutputDrainGrace = 5 * time.Second

// waitForOutputDrain blocks until every channel is closed, or until grace has
// elapsed. It returns true when all streams drained completely.
func waitForOutputDrain(grace time.Duration, streams ...<-chan struct{}) bool {
	timer := time.NewTimer(grace)
	defer timer.Stop()
	for _, stream := range streams {
		select {
		case <-stream:
		case <-timer.C:
			return false
		}
	}
	return true
}
