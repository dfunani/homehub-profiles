# Legacy Docker builder (no BuildKit) compatible: no RUN --mount.

ARG GO_VERSION=1.25.1
FROM golang:${GO_VERSION} AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

ARG TARGETARCH
COPY src ./src
RUN if [ -n "$TARGETARCH" ]; then export GOARCH="$TARGETARCH"; else export GOARCH="$(go env GOARCH)"; fi && \
    CGO_ENABLED=0 go build -o /bin/server ./src

################################################################################
FROM alpine:latest AS final

RUN apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

COPY --from=build /bin/server /bin/

EXPOSE 8000

ENTRYPOINT [ "/bin/server" ]
