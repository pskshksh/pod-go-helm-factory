// Command sampleapp is a tiny HTTP service used to exercise generated charts
// end-to-end. It serves a liveness probe at /healthz, a readiness probe at
// /readyz (the paths the generated chart's probes hit), and a greeting at /.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := run(ctx, port())
	if err != nil {
		log.Fatal(err)
	}
}
