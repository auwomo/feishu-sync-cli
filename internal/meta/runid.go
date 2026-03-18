package meta

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// NewRunID returns a reasonably unique run id without external deps.
func NewRunID() string {
	ts := time.Now().UTC().Format("20060102T150405Z")
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", ts, hex.EncodeToString(b))
}
