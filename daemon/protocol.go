package daemon

import "encoding/json"

// Request is sent from the client to the daemon
type Request struct {
	Action string            `json:"action"` // "execute" or "shutdown"
	Args   []string          `json:"args,omitempty"`
	Cwd    string            `json:"cwd,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
}

// Frame is used for bidirectional streaming after the initial request.
// Client→Daemon: type "stdin" (data) or "stdin_eof"
// Daemon→Client: type "stdout", "stderr", or "exit"
type Frame struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Code int    `json:"code,omitempty"`
}

func EncodeRequest(req *Request) ([]byte, error) {
	return json.Marshal(req)
}

func DecodeRequest(data []byte) (*Request, error) {
	var req Request
	err := json.Unmarshal(data, &req)
	return &req, err
}

func EncodeFrame(f *Frame) ([]byte, error) {
	return json.Marshal(f)
}

func DecodeFrame(data []byte) (*Frame, error) {
	var f Frame
	err := json.Unmarshal(data, &f)
	return &f, err
}
