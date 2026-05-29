package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
)

// Connect attempts to connect to the daemon socket.
// Returns the connection or nil if daemon is not running.
func Connect(socketPath string) net.Conn {
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		// Socket exists but daemon is not responding, clean up stale socket
		os.Remove(socketPath)
		os.Remove(socketPath + ".pid")
		return nil
	}

	return conn
}

// Forward sends a command to the daemon and streams stdin/stdout/stderr bidirectionally.
// Returns the exit code.
func Forward(conn net.Conn, args []string) int {
	defer conn.Close()

	cwd, _ := os.Getwd()

	// Build environment map
	env := make(map[string]string)
	for _, e := range os.Environ() {
		parts := splitEnvVar(e)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	req := &Request{
		Action: "execute",
		Args:   args,
		Cwd:    cwd,
		Env:    env,
	}

	data, err := EncodeRequest(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "daemon client: failed to encode request:", err)
		return 1
	}

	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		fmt.Fprintln(os.Stderr, "daemon client: failed to send request:", err)
		return 1
	}

	// Forward stdin to daemon in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		forwardStdin(conn)
	}()

	// Read responses
	exitCode := 1
	scanner := bufio.NewScanner(conn)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		frame, err := DecodeFrame(scanner.Bytes())
		if err != nil {
			continue
		}

		switch frame.Type {
		case "stdout":
			fmt.Fprint(os.Stdout, frame.Data)
		case "stderr":
			fmt.Fprint(os.Stderr, frame.Data)
		case "exit":
			exitCode = frame.Code
			// Signal stdin forwarder to stop by closing the write side
			if tc, ok := conn.(*net.UnixConn); ok {
				tc.CloseWrite()
			}
			return exitCode
		}
	}

	return exitCode
}

func forwardStdin(conn net.Conn) {
	buf := make([]byte, 4096)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			frame := &Frame{Type: "stdin", Data: string(buf[:n])}
			data, _ := json.Marshal(frame)
			data = append(data, '\n')
			if _, werr := conn.Write(data); werr != nil {
				return
			}
		}
		if err != nil {
			// Send EOF marker
			frame := &Frame{Type: "stdin_eof"}
			data, _ := json.Marshal(frame)
			data = append(data, '\n')
			conn.Write(data)
			return
		}
	}
}

// Shutdown sends a shutdown request to the daemon.
func Shutdown(socketPath string) error {
	conn := Connect(socketPath)
	if conn == nil {
		return fmt.Errorf("daemon is not running")
	}
	defer conn.Close()

	req := &Request{Action: "shutdown"}
	data, err := EncodeRequest(req)
	if err != nil {
		return err
	}

	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		return err
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		frame, err := DecodeFrame(scanner.Bytes())
		if err != nil {
			continue
		}

		switch frame.Type {
		case "stdout":
			fmt.Fprint(os.Stdout, frame.Data)
			fmt.Fprintln(os.Stdout)
		case "stderr":
			fmt.Fprint(os.Stderr, frame.Data)
			fmt.Fprintln(os.Stderr)
		case "exit":
			return nil
		}
	}

	return nil
}

// Ping checks if the daemon is alive.
func Ping(socketPath string) bool {
	conn := Connect(socketPath)
	if conn == nil {
		return false
	}
	defer conn.Close()

	req := &Request{Action: "ping"}
	data, _ := EncodeRequest(req)
	data = append(data, '\n')
	conn.Write(data)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		frame, err := DecodeFrame(scanner.Bytes())
		if err != nil {
			continue
		}
		if frame.Type == "exit" {
			return frame.Code == 0
		}
	}

	return false
}
