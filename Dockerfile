# syntax=docker/dockerfile:1

# Comments are provided throughout this file to help you get started.
# If you need more help, visit the Dockerfile reference guide at
# https://docs.docker.com/go/dockerfile-reference/

# Want to help us make this template better? Share your feedback here: https://forms.gle/ybq9Krt8jtBL3iCk7

################################################################################
# Create a stage for building the application.
ARG GO_VERSION=1.25.1
# Omit --platform=$BUILDPLATFORM: many hosts (e.g. CapRover) do not set BUILDPLATFORM,
# which yields --platform= and breaks the parser. Default platform is the builder's.
FROM golang:${GO_VERSION} AS build
WORKDIR /src

# Download dependencies as a separate step to take advantage of layer caching.
COPY go.mod go.sum ./
RUN go mod download -x

COPY . .

# Build an ARM64 binary to match the aarch64 runtime VM.
ARG TARGETARCH=arm64
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/server ./src

################################################################################
# Create a new stage for running the application that contains the minimal
# runtime dependencies for the application. This often uses a different base
# image from the build stage where the necessary files are copied from the build
# stage.
#
# The example below uses the alpine image as the foundation for running the app.
# By specifying the "latest" tag, it will also use whatever happens to be the
# most recent version of that image when you build your Dockerfile. If
# reproducibility is important, consider using a versioned tag
# (e.g., alpine:3.17.2) or SHA (e.g., alpine@sha256:c41ab5c992deb4fe7e5da09f67a8804a46bd0592bfdf0b1847dde0e0889d2bff).
FROM alpine:latest AS final

# Install runtime dependencies needed to run the application.
RUN apk --update add \
        ca-certificates \
        tzdata \
        && \
        update-ca-certificates

# Create a non-privileged user that the app will run under.
# See https://docs.docker.com/go/dockerfile-user-best-practices/
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

WORKDIR /app

# Copy the executable from the "build" stage.
COPY --from=build /bin/server /app/server

RUN chown -R appuser:appuser /app

USER appuser

# Expose the port that the application listens on.
EXPOSE 80

# What the container should run when it is started.
ENTRYPOINT [ "/app/server" ]