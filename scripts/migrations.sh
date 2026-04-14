#!/usr/bin/env bash
# Run Atlas migrations against RDS through an SSH tunnel.
#
# Terminal 1 — keep open (tunnel). Use 5433 if local Postgres/DBeaver uses 5432.
#   ssh -i ~/.ssh/bastion-homehub-eu-north-1.pem -N \
#     -L 5433:YOUR_RDS_ENDPOINT:5432 ec2-user@YOUR_BASTION_IP
#
# Terminal 2 — from repository root (Atlas CLI installed: https://atlasgo.io/getting-started)
#   ./initdb/migrations.sh
#
# Password special characters must be URL-encoded in DATABASE_URL.

set -euo pipefail
cd "$(dirname "$0")/.."

: "${DATABASE_URL:=}"
if [[ -z "${DATABASE_URL}" ]]; then
  echo "Set DATABASE_URL for the tunneled endpoint, e.g."
  echo '  export DATABASE_URL="postgres://USER:PASSWORD@127.0.0.1:5433/homehub_profiles?sslmode=require"'
  echo "Use sslmode=require for RDS. Port must match your ssh -L local port."
  exit 1
fi

atlas migrate apply --url "${DATABASE_URL}" --dir "file://$(pwd)/migrations"
