package cli

import "strings"

func authMode(cfgAuthMode string) string {
	m := strings.TrimSpace(strings.ToLower(cfgAuthMode))
	if m == "" {
		return "tenant"
	}
	return m
}
