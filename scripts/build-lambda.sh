#!/usr/bin/env bash
# Build a Linux/amd64 bootstrap binary and zip for AWS Lambda (provided.al2) / LocalStack.
# Run from the homehub-profiles repo root before starting LocalStack, or rely on init-localstack.sh
# if the LocalStack image has Go installed (uncommon).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
OUT_DIR="${1:-lambda-package}"
mkdir -p "$OUT_DIR"

echo "Building Lambda bootstrap (linux/amd64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "${OUT_DIR}/bootstrap" ./lambda

echo "Zipping..."
( cd "$OUT_DIR" && rm -f lambda-package.zip && zip -j lambda-package.zip bootstrap )
echo "OK: ${OUT_DIR}/lambda-package.zip"
