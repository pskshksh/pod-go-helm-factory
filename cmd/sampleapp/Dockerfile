# Build a static sampleapp binary and ship it on a minimal non-root base.
# Build context is the repo root: docker build -f cmd/sampleapp/Dockerfile .
FROM golang:1.26 AS build
WORKDIR /src

# Stdlib-only module: copying go.mod is enough for a cache-friendly download.
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/sampleapp ./cmd/sampleapp

# distroless/static:nonroot has no shell and runs as an unprivileged user,
# matching the chart's hardened securityContext (non-root, read-only rootfs).
FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/sampleapp /sampleapp
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/sampleapp"]
