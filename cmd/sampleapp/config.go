package main

import "os"

const defaultPort = "8080"

// port is the listen port from the PORT env var, defaulting to defaultPort.
func port() string {
	p := os.Getenv("PORT")
	if p == "" {
		return defaultPort
	}
	return p
}
