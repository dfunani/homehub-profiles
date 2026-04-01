#!/bin/bash


# ============================================================================
# Set environment variables
# ============================================================================

export AWS_ENDPOINT_URL=http://localhost:4566
echo "Checking LocalStack health..."
MAX_RETRIES=5
RETRY_COUNT=0
HEALTH_CHECK_URL="http://localhost:4566/_localstack/health"
BUCKET_NAME=${S3_BUCKET_NAME:-user-api-local}
REGION=${AWS_DEFAULT_REGION:-us-east-1}
LAMBDA_ROLE="arn:aws:iam::000000000000:role/lambda-role"
LAMBDA_PACKAGE_PATH="/tmp/lambda-package/lambda-package.zip"
LAMBDA_PACKAGE_DIR="/tmp/lambda-package"
LAMBDA_FUNCTION_NAME="${LAMBDA_FUNCTION_NAME:-user-api}"
# Go custom runtime (bootstrap binary in zip). LocalStack accepts provided.al2 like AWS.
LAMBDA_RUNTIME="${LAMBDA_RUNTIME:-provided.al2}"
LAMBDA_HANDLER="${LAMBDA_HANDLER:-bootstrap}"

# RDS (LocalStack emulated Postgres — see https://docs.localstack.cloud/aws/services/rds/)
# Master username cannot be "postgres" on LocalStack. Override via .env / compose.
RDS_DB_INSTANCE_IDENTIFIER="${RDS_DB_INSTANCE_IDENTIFIER:-user-api-rds}"
RDS_MASTER_USERNAME="${RDS_MASTER_USERNAME:-homehubadmin}"
RDS_MASTER_PASSWORD="${RDS_MASTER_PASSWORD:-homehubpassword}"
RDS_DB_NAME="${RDS_DB_NAME:-homehub}"
RDS_INSTANCE_CLASS="${RDS_INSTANCE_CLASS:-db.t3.micro}"
RDS_ALLOCATED_STORAGE="${RDS_ALLOCATED_STORAGE:-20}"
# Where to write connection hints (compose mounts ./localstack here)
LOCALSTACK_DATA_DIR="${LOCALSTACK_DATA_DIR:-/var/lib/localstack/data}"
# RDS APIs require LocalStack Pro (or a licensed plan). Community edition returns 501.
# Set ENABLE_LOCALSTACK_RDS=1 only when your license includes RDS.
ENABLE_LOCALSTACK_RDS="${ENABLE_LOCALSTACK_RDS:-0}"

# ============================================================================
# Run LocalStack health check
# ============================================================================

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  if curl -f "$HEALTH_CHECK_URL" >/dev/null 2>&1; then
    echo "LocalStack is ready!"
    break
  else
    RETRY_COUNT=$((RETRY_COUNT + 1))
    if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
      echo "Waiting for LocalStack to become healthy... (attempt $RETRY_COUNT/$MAX_RETRIES)"
      sleep 2
    else
      echo "ERROR: LocalStack health check failed after $MAX_RETRIES attempts."
      echo "LocalStack is not running or not healthy at $HEALTH_CHECK_URL"
      exit 1
    fi
  fi
done

# ============================================================================
# Create S3 bucket
# ============================================================================


echo "Creating S3 bucket: $BUCKET_NAME in region: $REGION"

# Use awslocal to create bucket (idempotent - will skip if exists)
awslocal s3 mb s3://$BUCKET_NAME --region $REGION 2>&1 || {
  # Check if bucket already exists
  if awslocal s3 ls s3://$BUCKET_NAME 2>&1 | grep -q "NoSuchBucket"; then
    echo "Bucket creation failed"
    exit 1
  else
    echo "Bucket already exists or was created successfully"
  fi
}

echo "S3 bucket '$BUCKET_NAME' is ready!"

# awslocal s3 ls s3://rc-faas-offer-local-offer --endpoint-url http://faas-offer-localstack:4566 --recursive --human-readable
# awslocal s3 cp s3://rc-faas-offer-local-offer/doc/truid/52667c0d-f138-4d42-af05-dd045f758114/results/result.json ./result.json


# ============================================================================
# Create RDS PostgreSQL instance (LocalStack Pro only — opt-in)
# ============================================================================
#
# Community LocalStack returns 501 for rds.* APIs ("not included in your license plan").
# Enable only with Pro + auth token: ENABLE_LOCALSTACK_RDS=1 and SERVICES including rds.
#
# For plain Postgres locally, use the Compose `database` service instead.
# Docs: https://docs.localstack.cloud/aws/services/rds/

if [ "${ENABLE_LOCALSTACK_RDS}" = "1" ]; then
  echo "Ensuring RDS instance: $RDS_DB_INSTANCE_IDENTIFIER"

  if awslocal rds describe-db-instances --db-instance-identifier "$RDS_DB_INSTANCE_IDENTIFIER" \
    --query 'DBInstances[0].DBInstanceIdentifier' --output text 2>/dev/null | grep -q "$RDS_DB_INSTANCE_IDENTIFIER"; then
    echo "RDS instance '$RDS_DB_INSTANCE_IDENTIFIER' already exists."
  else
    echo "Creating RDS PostgreSQL instance..."
    if ! awslocal rds create-db-instance \
      --db-instance-identifier "$RDS_DB_INSTANCE_IDENTIFIER" \
      --db-instance-class "$RDS_INSTANCE_CLASS" \
      --engine postgres \
      --master-username "$RDS_MASTER_USERNAME" \
      --master-user-password "$RDS_MASTER_PASSWORD" \
      --db-name "$RDS_DB_NAME" \
      --allocated-storage "$RDS_ALLOCATED_STORAGE" \
      --region "$REGION" 2>&1; then
      echo "WARNING: RDS create-db-instance failed (Pro license / engine / Docker socket). Continuing."
    fi
  fi

  if awslocal rds describe-db-instances --db-instance-identifier "$RDS_DB_INSTANCE_IDENTIFIER" \
    --query 'DBInstances[0].DBInstanceIdentifier' --output text 2>/dev/null | grep -q "$RDS_DB_INSTANCE_IDENTIFIER"; then
    echo "Waiting for RDS instance to become available..."
    for _ in $(seq 1 90); do
      RDS_STATUS=$(awslocal rds describe-db-instances --db-instance-identifier "$RDS_DB_INSTANCE_IDENTIFIER" \
        --query 'DBInstances[0].DBInstanceStatus' --output text 2>/dev/null || echo "unknown")
      if [ "$RDS_STATUS" = "available" ]; then
        echo "RDS status: available"
        break
      fi
      sleep 2
    done

    RDS_HOST=$(awslocal rds describe-db-instances --db-instance-identifier "$RDS_DB_INSTANCE_IDENTIFIER" \
      --query 'DBInstances[0].Endpoint.Address' --output text 2>/dev/null || echo "")
    RDS_PORT=$(awslocal rds describe-db-instances --db-instance-identifier "$RDS_DB_INSTANCE_IDENTIFIER" \
      --query 'DBInstances[0].Endpoint.Port' --output text 2>/dev/null || echo "5432")

    if [ -n "$RDS_HOST" ] && [ "$RDS_HOST" != "None" ]; then
      mkdir -p "$LOCALSTACK_DATA_DIR"
      RDS_URL="postgres://${RDS_MASTER_USERNAME}:${RDS_MASTER_PASSWORD}@${RDS_HOST}:${RDS_PORT}/${RDS_DB_NAME}?sslmode=disable"
      {
        echo "# Generated by init-localstack.sh — LocalStack RDS endpoint (source from host: ./localstack/rds-endpoint.env)"
        echo "export RDS_HOST=${RDS_HOST}"
        echo "export RDS_PORT=${RDS_PORT}"
        echo "export RDS_DB_NAME=${RDS_DB_NAME}"
        echo "export RDS_MASTER_USERNAME=${RDS_MASTER_USERNAME}"
        printf 'export DATABASE_URL=%q\n' "$RDS_URL"
      } > "${LOCALSTACK_DATA_DIR}/rds-endpoint.env"
      echo "Wrote ${LOCALSTACK_DATA_DIR}/rds-endpoint.env"
      echo "RDS endpoint: ${RDS_HOST}:${RDS_PORT} database=${RDS_DB_NAME} user=${RDS_MASTER_USERNAME}"
    else
      echo "Could not read RDS endpoint from describe-db-instances."
    fi
  else
    echo "RDS instance not found after create; skipping endpoint file."
  fi
else
  echo "Skipping LocalStack RDS (ENABLE_LOCALSTACK_RDS is not 1). Use Compose Postgres or enable Pro + RDS."
fi


# ============================================================================
# Create Main Lambda function (Go, provided.al2 / bootstrap)
# ============================================================================
#
# Expected layout:
#   - Mount the Go module at /tmp/source (repo root with go.mod and cmd/lambda).
#   - Mount ./lambda-package -> /tmp/lambda-package for the zip output.
#
# Recommended before compose up (LocalStack image usually has no Go toolchain):
#   ./scripts/build-lambda.sh
#   → produces lambda-package/lambda-package.zip
#
# If lambda-package.zip is missing but `go` exists in PATH, we build inside the container.

echo "Creating Lambda function: $LAMBDA_FUNCTION_NAME (runtime=$LAMBDA_RUNTIME handler=$LAMBDA_HANDLER)"

LAMBDA_PACKAGE_PATH=""
USE_S3=false
S3_KEY=""

mkdir -p "$LAMBDA_PACKAGE_DIR"

if [ -f "$LAMBDA_PACKAGE_DIR/lambda-package.zip" ] && [ -s "$LAMBDA_PACKAGE_DIR/lambda-package.zip" ]; then
  LAMBDA_PACKAGE_PATH="$LAMBDA_PACKAGE_DIR/lambda-package.zip"
  echo "Using pre-built Lambda zip: $LAMBDA_PACKAGE_PATH"
elif [ -f /tmp/source/go.mod ] && command -v go >/dev/null 2>&1; then
  echo "Building Go Lambda from /tmp/source (linux/amd64)..."
  WORK="/tmp/lambda-go-build"
  rm -rf "$WORK"
  mkdir -p "$WORK"
  export GOCACHE=/tmp/go-build-cache
  export GOMODCACHE=/tmp/go-mod-cache
  mkdir -p "$GOCACHE" "$GOMODCACHE"
  # /tmp/source is often mounted read-only; caches must be writable
  if ( cd /tmp/source && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$WORK/bootstrap" ./cmd/lambda ); then
    if ! command -v zip >/dev/null 2>&1; then
      echo "ERROR: zip not found; cannot package Lambda"
    else
      ( cd "$WORK" && zip -j "$LAMBDA_PACKAGE_DIR/lambda-package.zip" bootstrap )
      LAMBDA_PACKAGE_PATH="$LAMBDA_PACKAGE_DIR/lambda-package.zip"
      echo "Lambda zip created: $LAMBDA_PACKAGE_PATH"
    fi
  else
    echo "ERROR: go build ./cmd/lambda failed"
  fi
else
  echo "ERROR: No lambda-package/lambda-package.zip and cannot build (need /tmp/source/go.mod + go in PATH)."
  echo "Run from host: ./scripts/build-lambda.sh"
fi

# Optional: upload large zips to S3 (same threshold as before)
if [ -n "$LAMBDA_PACKAGE_PATH" ] && [ -f "$LAMBDA_PACKAGE_PATH" ]; then
  PACKAGE_SIZE_BYTES=$(du -b "$LAMBDA_PACKAGE_PATH" | cut -f1)
  echo "Package size: $(du -h "$LAMBDA_PACKAGE_PATH" | cut -f1)"
  if [ "${PACKAGE_SIZE_BYTES:-0}" -gt 52428800 ]; then
    echo "⚠ WARNING: Package exceeds 50MB; uploading to S3..."
    TIMESTAMP=$(date +%s)
    S3_KEY="lambda-packages/${LAMBDA_FUNCTION_NAME}-${TIMESTAMP}.zip"
    if awslocal s3 cp "$LAMBDA_PACKAGE_PATH" "s3://${BUCKET_NAME}/${S3_KEY}" --region "$REGION" 2>&1; then
      echo "Package uploaded: s3://${BUCKET_NAME}/${S3_KEY}"
      USE_S3=true
    else
      echo "ERROR: S3 upload failed; skipping Lambda"
      LAMBDA_PACKAGE_PATH=""
    fi
  fi
fi

# Proceed with Lambda function creation if package exists
if [ -n "$LAMBDA_PACKAGE_PATH" ] && [ -f "$LAMBDA_PACKAGE_PATH" ]; then
  echo "Lambda package found, creating/updating function..."

  if [ -n "${LAMBDA_ENV_VARS_FROM_FILE:-}" ] && [ -f "${LAMBDA_ENV_VARS_FROM_FILE}" ]; then
    echo "Loading environment variables from file: ${LAMBDA_ENV_VARS_FROM_FILE}"
    set -a
    # shellcheck disable=SC1090
    source "${LAMBDA_ENV_VARS_FROM_FILE}"
    set +a
  fi

  LAMBDA_ENV_ARGS=()
  ENV_VARS_ARRAY=()

  while IFS= read -r line; do
    if echo "$line" | grep -qE '^(AWS_|LOCALSTACK_|HOME=|PATH=|PWD=|SHELL=|USER=|SHLVL=|_=|TERM=|HOSTNAME=|LAMBDA_|S3_|REGION=|BUCKET_|FUNCTION_)'; then
      continue
    fi
    var="${line%%=*}"
    value="${line#*=}"
    if [ -z "$var" ] || [ -z "$value" ]; then
      continue
    fi
    ENV_VARS_ARRAY+=("${var}=${value}")
  done < <(env | sort -u)

  if [ ${#ENV_VARS_ARRAY[@]} -gt 0 ]; then
    JSON_VARS=""
    FIRST=true
    for var_value in "${ENV_VARS_ARRAY[@]}"; do
      var="${var_value%%=*}"
      value="${var_value#*=}"
      escaped_value=$(echo "$value" | sed 's/\\/\\\\/g' | sed 's/"/\\"/g')
      if [ "$FIRST" = true ]; then
        JSON_VARS="\"${var}\":\"${escaped_value}\""
        FIRST=false
      else
        JSON_VARS="${JSON_VARS},\"${var}\":\"${escaped_value}\""
      fi
    done
    ENV_JSON="{\"Variables\":{${JSON_VARS}}}"
    LAMBDA_ENV_ARGS=("--environment" "$ENV_JSON")
    echo "Passing ${#ENV_VARS_ARRAY[@]} environment variable(s) to Lambda function..."
  else
    echo "No environment variables to pass to Lambda function"
  fi

  CREATE_CODE_ARGS=()
  UPDATE_CODE_ARGS=()
  if [ "$USE_S3" = true ] && [ -n "$S3_KEY" ]; then
    CREATE_CODE_ARGS=("--code" "S3Bucket=${BUCKET_NAME},S3Key=${S3_KEY}")
    UPDATE_CODE_ARGS=("--s3-bucket" "${BUCKET_NAME}" "--s3-key" "${S3_KEY}")
    echo "Using S3 for Lambda code: s3://${BUCKET_NAME}/${S3_KEY}"
  else
    CREATE_CODE_ARGS=("--zip-file" "fileb://$LAMBDA_PACKAGE_PATH")
    UPDATE_CODE_ARGS=("--zip-file" "fileb://$LAMBDA_PACKAGE_PATH")
    echo "Using local zip for Lambda code"
  fi

  if awslocal lambda get-function --function-name "$LAMBDA_FUNCTION_NAME" 2>&1 | grep -q "ResourceNotFoundException"; then
    echo "Creating Lambda function..."
    if [ ${#LAMBDA_ENV_ARGS[@]} -gt 0 ]; then
      awslocal lambda create-function \
        --function-name "$LAMBDA_FUNCTION_NAME" \
        --runtime "$LAMBDA_RUNTIME" \
        --role "$LAMBDA_ROLE" \
        --handler "$LAMBDA_HANDLER" \
        "${CREATE_CODE_ARGS[@]}" \
        --timeout 300 \
        --memory-size 512 \
        "${LAMBDA_ENV_ARGS[@]}" \
        --region "$REGION" 2>&1 && {
        echo "Lambda function '$LAMBDA_FUNCTION_NAME' created successfully with environment variables!"
      } || {
        echo "Lambda function creation failed"
      }
    else
      awslocal lambda create-function \
        --function-name "$LAMBDA_FUNCTION_NAME" \
        --runtime "$LAMBDA_RUNTIME" \
        --role "$LAMBDA_ROLE" \
        --handler "$LAMBDA_HANDLER" \
        "${CREATE_CODE_ARGS[@]}" \
        --timeout 300 \
        --memory-size 512 \
        --region "$REGION" 2>&1 && {
        echo "Lambda function '$LAMBDA_FUNCTION_NAME' created successfully!"
      } || {
        echo "Lambda function creation failed"
      }
    fi
  else
    echo "Lambda function '$LAMBDA_FUNCTION_NAME' already exists, updating code..."
    awslocal lambda update-function-code \
      --function-name "$LAMBDA_FUNCTION_NAME" \
      "${UPDATE_CODE_ARGS[@]}" \
      --region "$REGION" 2>&1 && {
      echo "Lambda function code updated!"
      if [ ${#LAMBDA_ENV_ARGS[@]} -gt 0 ]; then
        echo "Updating Lambda function environment variables..."
        awslocal lambda update-function-configuration \
          --function-name "$LAMBDA_FUNCTION_NAME" \
          "${LAMBDA_ENV_ARGS[@]}" \
          --region "$REGION" 2>&1 && {
          echo "Lambda function environment variables updated!"
        } || {
          echo "Failed to update Lambda function environment variables"
        }
      fi
    } || {
      echo "Lambda function update failed"
    }
  fi
else
  echo "Lambda package not available, skipping Main Lambda function creation"
fi

echo "LocalStack initialization complete!"
