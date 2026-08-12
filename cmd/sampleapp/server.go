package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

const shutdownTimeout = 10 * time.Second

// run starts the HTTP server on port and blocks until ctx is cancelled (SIGTERM
// from Kubernetes), then shuts down gracefully. It returns early if the server
// fails to start.
func run(ctx context.Context, port string) error {
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errc := make(chan error, 1)
	go func() {
		log.Printf("sampleapp listening on :%s", port)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errc <- err
			return
		}
		errc <- nil
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}

	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
