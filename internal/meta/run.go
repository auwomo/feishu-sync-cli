package meta

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/your-org/feishu-sync/internal/config"
)

// Run holds per-invocation metadata for pull/run.
// It intentionally avoids secrets.
//
// Meta files live under <output>/_meta/.
// - manifest.json: summary of the run and discovery counts
// - ledger.jsonl: JSONL events for discovery/export
//
// NOTE: This is a minimal implementation: no big refactor, best-effort writes.

type Run struct {
	RunID     string
	StartedAt time.Time
	EndedAt   time.Time

	Mode  string
	Scope RunScope

	ConfigFingerprint string

	Version string
	Commit  string
}

type RunScope struct {
	DriveFolderTokens []string
	WikiSpaceIDs      []string
}

type Manifest struct {
	RunID     string `json:"run_id"`
	StartedAt string `json:"started_at"`
	EndedAt   string `json:"ended_at"`
	Mode      string `json:"mode"`
	Scope     struct {
		DriveFolderTokens []string `json:"drive_folder_tokens"`
		WikiSpaceIDs      []string `json:"wiki_space_ids"`
	} `json:"scope"`
	Counts struct {
		Drive struct {
			RootsDiscovered int `json:"roots_discovered"`
			ItemsDiscovered int `json:"items_discovered"`
			Exported        int `json:"exported"`
			Unsupported     int `json:"unsupported"`
			Errors          int `json:"errors"`
		} `json:"drive"`
		Wiki struct {
			SpacesDiscovered int `json:"spaces_discovered"`
			ItemsDiscovered  int `json:"items_discovered"`
			Exported         int `json:"exported"`
			Errors           int `json:"errors"`
		} `json:"wiki"`
	} `json:"counts"`
	ConfigFingerprint string `json:"config_fingerprint"`
	Version           string `json:"version,omitempty"`
	Commit            string `json:"commit,omitempty"`
}

// FingerprintConfig hashes a sanitized subset of config so the fingerprint is stable
// and avoids secrets.
func FingerprintConfig(cfg *config.Config) string {
	s := struct {
		AppID string `json:"app_id"`
		Auth  struct {
			Mode string `json:"mode"`
		} `json:"auth"`
		Scope struct {
			Mode              string   `json:"mode"`
			DriveFolderTokens []string `json:"drive_folder_tokens"`
			WikiSpaceIDs      []string `json:"wiki_space_ids"`
		} `json:"scope"`
		Output struct {
			Dir string `json:"dir"`
		} `json:"output"`
		Runtime struct {
			Concurrency  int  `json:"concurrency"`
			RateLimitQPS int  `json:"rate_limit_qps"`
			Incremental  bool `json:"incremental"`
		} `json:"runtime"`
	}{
		AppID: cfg.App.ID,
	}
	s.Auth.Mode = cfg.Auth.Mode
	s.Scope.Mode = cfg.Scope.Mode
	s.Scope.DriveFolderTokens = cfg.Scope.DriveFolderTokens
	s.Scope.WikiSpaceIDs = cfg.Scope.WikiSpaceIDs
	s.Output.Dir = cfg.Output.Dir
	s.Runtime.Concurrency = cfg.Runtime.Concurrency
	s.Runtime.RateLimitQPS = cfg.Runtime.RateLimitQPS
	s.Runtime.Incremental = cfg.Runtime.Incremental

	b, _ := json.Marshal(s)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])[:16]
}

func NewRun(runID string, cfg *config.Config) *Run {
	r := &Run{RunID: runID, StartedAt: time.Now(), Mode: cfg.Scope.Mode}
	r.Scope = RunScope{DriveFolderTokens: append([]string(nil), cfg.Scope.DriveFolderTokens...), WikiSpaceIDs: append([]string(nil), cfg.Scope.WikiSpaceIDs...)}
	r.ConfigFingerprint = FingerprintConfig(cfg)
	r.Version = Version
	r.Commit = Commit
	return r
}

func (r *Run) End() {
	r.EndedAt = time.Now()
}

func (r *Run) RFC3339(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func (r *Run) BuildManifest() Manifest {
	var m Manifest
	m.RunID = r.RunID
	m.StartedAt = r.RFC3339(r.StartedAt)
	m.EndedAt = r.RFC3339(r.EndedAt)
	m.Mode = r.Mode
	m.Scope.DriveFolderTokens = r.Scope.DriveFolderTokens
	m.Scope.WikiSpaceIDs = r.Scope.WikiSpaceIDs
	m.ConfigFingerprint = r.ConfigFingerprint
	m.Version = r.Version
	m.Commit = r.Commit
	return m
}
