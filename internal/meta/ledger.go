package meta

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Ledger struct {
	runID string
	mu    sync.Mutex
	w     *os.File
}

type Entry struct {
	RunID         string `json:"run_id"`
	ResourceType  string `json:"resource_type"`
	ResourceToken string `json:"resource_token"`
	Action        string `json:"action"`
	Status        string `json:"status"` // ok|skipped|error
	StartedAt     string `json:"started_at"`
	EndedAt       string `json:"ended_at"`
	DurationMS    int64  `json:"duration_ms"`
	Bytes         int64  `json:"bytes,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
	Retryable     *bool  `json:"retryable,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
}

func OpenLedger(path string, runID string) (*Ledger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, err
	}
	return &Ledger{runID: runID, w: f}, nil
}

func (l *Ledger) Close() error {
	if l == nil || l.w == nil {
		return nil
	}
	return l.w.Close()
}

func (l *Ledger) Write(e Entry) {
	if l == nil || l.w == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = json.NewEncoder(l.w).Encode(e)
}

func NowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func Trunc(s string, max int) string {
	if max <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

// Timer helps record started/ended/duration.
type Timer struct {
	Start time.Time
}

func StartTimer() Timer {
	return Timer{Start: time.Now()}
}

func (t Timer) Done() (startedAt string, endedAt string, durMS int64) {
	end := time.Now()
	return t.Start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339), end.Sub(t.Start).Milliseconds()
}
