package daemon

import (
	"bufio"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const SocketName = ".aux4.daemon.sock"
const IdleTimeout = 30 * time.Minute

// daemonSessionEnv carries a per-top-level-command token into the environment of
// the command the daemon runs. Any `aux4` that command shells out inherits it
// (see output.ChildEnv) and forwards it back, letting the server recognize the
// re-entrant/nested call and run it WITHOUT waiting on the command mutex the
// parent already holds — the fix for the nested-call deadlock.
const daemonSessionEnv = "AUX4_DAEMON_SESSION"

// ExecuteFunc is called by the server to execute a command.
// It receives args, stdin reader, and stdout/stderr writers.
// Returns the exit code.
type ExecuteFunc func(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int

type Server struct {
	socketPath string
	listener   net.Listener
	executeFn  ExecuteFunc
	mu         sync.Mutex // serializes independent top-level commands
	sessMu     sync.Mutex // guards activeSession
	// activeSession is the token of the top-level command currently holding mu
	// (empty when idle). A request carrying this token is a nested call from that
	// command and is allowed to run re-entrantly without waiting on mu.
	activeSession string
	idleTimer     *time.Timer
	done          chan struct{}
}

// newSessionID returns a random token identifying a top-level command run.
func newSessionID() string {
	b := make([]byte, 16)
	if _, err := crand.Read(b); err != nil {
		return fmt.Sprintf("sess-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// SocketPathEnv, when set, is the absolute socket path used by both the daemon
// server and its clients, overriding the CWD-relative discovery below. It lets a
// caller place the socket in a known writable location (e.g. an api server that
// starts a warm daemon in /tmp and points its command clients at it) WITHOUT
// changing anyone's working directory — the read-only-CWD case (AWS Lambda) and
// callers that must preserve CWD both rely on this.
const SocketPathEnv = "AUX4_DAEMON_SOCKET"

func FindSocketPath(fromDir string) string {
	if sock := os.Getenv(SocketPathEnv); sock != "" {
		return sock
	}

	dir, err := filepath.Abs(fromDir)
	if err != nil {
		dir = fromDir
	}

	for {
		aux4File := filepath.Join(dir, ".aux4")
		if _, err := os.Stat(aux4File); err == nil {
			return filepath.Join(dir, SocketName)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: use the original directory
	abs, err := filepath.Abs(fromDir)
	if err != nil {
		abs = fromDir
	}
	return filepath.Join(abs, SocketName)
}

func StartServer(socketPath string, executeFn ExecuteFunc) error {
	// Remove stale socket if it exists
	if _, err := os.Stat(socketPath); err == nil {
		// Try connecting to check if it's active
		conn, err := net.Dial("unix", socketPath)
		if err == nil {
			conn.Close()
			return fmt.Errorf("daemon already running at %s", socketPath)
		}
		// Stale socket, remove it
		os.Remove(socketPath)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", socketPath, err)
	}

	server := &Server{
		socketPath: socketPath,
		listener:   listener,
		executeFn:  executeFn,
		done:       make(chan struct{}),
	}

	// Set up idle timeout
	server.idleTimer = time.AfterFunc(IdleTimeout, func() {
		fmt.Fprintln(os.Stderr, "daemon idle timeout, shutting down")
		server.Shutdown()
	})

	fmt.Fprintln(os.Stdout, "daemon started at", socketPath)
	fmt.Fprintln(os.Stdout, "idle timeout:", IdleTimeout)

	// Write PID file
	pidPath := socketPath + ".pid"
	os.WriteFile(pidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)

	// Accept connections
	go server.acceptLoop()

	// Wait for shutdown
	<-server.done

	// Cleanup
	os.Remove(socketPath)
	os.Remove(pidPath)
	fmt.Fprintln(os.Stdout, "daemon stopped")

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				continue
			}
		}

		// Reset idle timer on each connection
		s.idleTimer.Reset(IdleTimeout)

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	line, err := reader.ReadBytes('\n')
	if err != nil {
		s.sendFrame(conn, &Frame{Type: "stderr", Data: "failed to read request"})
		s.sendFrame(conn, &Frame{Type: "exit", Code: 1})
		return
	}

	req, err := DecodeRequest(line)
	if err != nil {
		s.sendFrame(conn, &Frame{Type: "stderr", Data: "invalid request format"})
		s.sendFrame(conn, &Frame{Type: "exit", Code: 1})
		return
	}

	switch req.Action {
	case "shutdown":
		s.sendFrame(conn, &Frame{Type: "stdout", Data: "daemon shutting down\n"})
		s.sendFrame(conn, &Frame{Type: "exit", Code: 0})
		s.Shutdown()
	case "execute":
		s.executeCommand(conn, reader, req)
	case "ping":
		s.sendFrame(conn, &Frame{Type: "stdout", Data: "pong"})
		s.sendFrame(conn, &Frame{Type: "exit", Code: 0})
	default:
		s.sendFrame(conn, &Frame{Type: "stderr", Data: "unknown action: " + req.Action})
		s.sendFrame(conn, &Frame{Type: "exit", Code: 1})
	}
}

func (s *Server) executeCommand(conn net.Conn, reader *bufio.Reader, req *Request) {
	// A nested `aux4` call — one shelled out by a command the daemon is already
	// running — forwards back into the daemon carrying the parent's session token
	// in its environment. It must NOT wait on s.mu: the parent holds it and is
	// parked in the very shell-out that spawned this call, so blocking here would
	// deadlock (and wedge the daemon for all later commands). Detect the token
	// match and skip the lock. The parent is parked, so at most one execution is
	// ever ACTIVELY touching the process-global cwd/env/stdio at a time.
	reqSession := ""
	if req.Env != nil {
		reqSession = req.Env[daemonSessionEnv]
	}

	s.sessMu.Lock()
	isNested := reqSession != "" && reqSession == s.activeSession
	s.sessMu.Unlock()

	var session string
	if !isNested {
		// Top-level command: serialize against other independent top-level
		// commands (they share process-global cwd/env/stdio) and publish a fresh
		// session token so this command's own nested calls re-enter.
		s.mu.Lock()
		defer s.mu.Unlock()

		session = newSessionID()
		s.sessMu.Lock()
		s.activeSession = session
		s.sessMu.Unlock()
		defer func() {
			s.sessMu.Lock()
			s.activeSession = ""
			s.sessMu.Unlock()
		}()
	}

	// Change to requested working directory
	if req.Cwd != "" {
		origDir, _ := os.Getwd()
		if err := os.Chdir(req.Cwd); err != nil {
			s.sendFrame(conn, &Frame{Type: "stderr", Data: fmt.Sprintf("failed to chdir to %s: %v", req.Cwd, err)})
			s.sendFrame(conn, &Frame{Type: "exit", Code: 1})
			return
		}
		defer os.Chdir(origDir)
	}

	// Set environment variables
	var origEnv []string
	if req.Env != nil {
		origEnv = os.Environ()
		for k, v := range req.Env {
			os.Setenv(k, v)
		}
	}
	// Publish the session token into the process environment of the command we
	// run so any `aux4` it shells out inherits it (via output.ChildEnv) and
	// re-enters instead of deadlocking. Restored with the rest of the env below.
	if !isNested {
		if origEnv == nil {
			origEnv = os.Environ()
		}
		os.Setenv(daemonSessionEnv, session)
	}

	// Create a pipe for stdin — client sends stdin frames, we pipe them to the command
	stdinR, stdinW := io.Pipe()

	// Read stdin frames from the client in background
	go func() {
		defer stdinW.Close()
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return
			}
			frame, err := DecodeFrame(line)
			if err != nil {
				continue
			}
			switch frame.Type {
			case "stdin":
				stdinW.Write([]byte(frame.Data))
			case "stdin_eof":
				return
			}
		}
	}()

	// Create writers that stream to the client
	stdoutW := &frameWriter{conn: conn, outputType: "stdout"}
	stderrW := &frameWriter{conn: conn, outputType: "stderr"}

	// Execute the command
	exitCode := s.executeFn(req.Args, stdinR, stdoutW, stderrW)

	// Restore environment
	if origEnv != nil {
		os.Clearenv()
		for _, env := range origEnv {
			parts := splitEnvVar(env)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
	}

	s.sendFrame(conn, &Frame{Type: "exit", Code: exitCode})
}

func (s *Server) sendFrame(conn net.Conn, f *Frame) {
	data, err := json.Marshal(f)
	if err != nil {
		return
	}
	data = append(data, '\n')
	conn.Write(data)
}

func (s *Server) Shutdown() {
	s.idleTimer.Stop()
	s.listener.Close()
	close(s.done)
}

// frameWriter implements io.Writer and streams data back to the client as frames
type frameWriter struct {
	conn       net.Conn
	outputType string
}

func (w *frameWriter) Write(p []byte) (n int, err error) {
	f := &Frame{Type: w.outputType, Data: string(p)}
	data, err := json.Marshal(f)
	if err != nil {
		return 0, err
	}
	data = append(data, '\n')
	_, err = w.conn.Write(data)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func splitEnvVar(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env}
}
