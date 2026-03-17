package cli

import "fmt"

func authLoginRedirectURI(host string, port int, callbackPath string) string {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 18900
	}
	if callbackPath == "" {
		callbackPath = "/callback"
	}
	if callbackPath[0] != '/' {
		callbackPath = "/" + callbackPath
	}
	return fmt.Sprintf("http://%s:%d%s", host, port, callbackPath)
}
