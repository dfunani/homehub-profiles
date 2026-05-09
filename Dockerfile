# syntax=docker/dockerfile:1

# Match your module’s Go version (adjust if go.mod differs).
ARG GO_VERSION=1.25.1

################################################################################
# Build the Go binary
################################################################################
FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Default amd64; override for ARM hosts / CapRover ARM:
#   docker build --build-arg TARGETARCH=arm64 .
ARG TARGETOS=linux
ARG TARGETARCH=armd64

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -ldflags="-s -w" -o /bin/server ./src

################################################################################
# Atlas CLI binary (copied into final image)
################################################################################
FROM arigaio/atlas:latest AS atlas

################################################################################
# Runtime image
################################################################################
FROM alpine:3.20 AS final

RUN apk add --no-cache \
        ca-certificates \
        tzdata \
        postgresql16-client \
    && update-ca-certificates

ARG UID=10001
RUN adduser -D -h /home/appuser -u "${UID}" appuser

ENV HOME=/home/appuser
ENV GOCACHE=/home/appuser/.cache/go-build
ENV GOMODCACHE=/home/appuser/go/pkg/mod

WORKDIR /app

COPY --from=build /bin/server /app/server

# Official image keeps the binary here; if COPY fails, run:
#   docker run --rm --entrypoint ls arigaio/atlas:latest -la /
COPY --from=atlas /atlas /usr/local/bin/atlas

COPY migrations /app/migrations
COPY atlas.hcl /app/atlas.hcl
COPY scripts/entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh \
 && chown -R appuser:appuser /app /home/appuser

USER appuser

EXPOSE 80

ENTRYPOINT ["/app/entrypoint.sh"]