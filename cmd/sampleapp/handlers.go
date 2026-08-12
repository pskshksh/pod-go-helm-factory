package main

import "net/http"

// routes builds the HTTP handler: liveness, readiness, and a root greeting.
func routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", ok)
	mux.HandleFunc("/readyz", ok)
	mux.HandleFunc("/", root)
	return mux
}

// ok answers health and readiness probes with 200 OK.
func ok(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

// root greets requests to / and 404s anything else (the catch-all pattern).
func root(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	_, _ = w.Write([]byte("hello from sampleapp\n"))
}
