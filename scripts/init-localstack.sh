#!/bin/bash


# ============================================================================
# Set environment variables
# ============================================================================

export AWS_ENDPOINT_URL=http://localhost:4566
echo "Checking LocalStack health..."
MAX_RETRIES=5
RETRY_COUNT=0
HEALTH_CHECK_URL="http://localhost:4566/_localstack/health"
BUCKET_NAME=${S3_BUCKET_NAME:-profiles-api-local}
REGION=${AWS_DEFAULT_REGION:-us-east-1}
LAMBDA_ROLE="arn:aws:iam::000000000000:role/lambda-role"
LAMBDA_PACKAGE_PATH="/tmp/lambda-package/lambda-package.zip"
LAMBDA_PACKAGE_DIR="/tmp/lambda-package"
LAMBDA_FUNCTION_NAME="${LAMBDA_FUNCTION_NAME:-profiles-api}"
# Go custom runtime (bootstrap binary in zip). LocalStack accepts provided.al2 like AWS.
LAMBDA_RUNTIME="${LAMBDA_RUNTIME:-provided.al2}"
LAMBDA_HANDLER="${LAMBDA_HANDLER:-bootstrap}"


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

echo "LocalStack initialization complete!"
