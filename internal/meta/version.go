package meta

// These are injected at build time via -ldflags.
// Fallbacks are "dev" when not provided.

var (
	Version = "dev"
	Commit  = "dev"
	Date    = "dev"
)
