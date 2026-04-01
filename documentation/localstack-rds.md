# LocalStack RDS (PostgreSQL)

**RDS is not available on the free Community image.** The `rds` APIs return `501` unless your plan includes RDS (typically **LocalStack Pro** with an auth token).

This repo keeps RDS **off by default**:

- `ENABLE_LOCALSTACK_RDS=0` (default) — init script skips all `awslocal rds` calls (no error spam).
- `SERVICES` defaults to `s3,secretsmanager,lambda` (no `rds`).

To try RDS with a **licensed** LocalStack:

1. Set `ENABLE_LOCALSTACK_RDS=1` in `.env`.
2. Set `LOCALSTACK_SERVICES=rds,s3,secretsmanager,lambda` (or add `rds` to your list).
3. Configure Pro/auth per [LocalStack docs](https://docs.localstack.cloud/getting-started/auth-token/).

For everyday local Postgres **without** RDS APIs, use the Compose **`database`** service.

---

When RDS is enabled and supported, `scripts/init-localstack.sh` calls **`awslocal rds create-db-instance`** after the S3 bucket step. LocalStack starts a real Postgres process (needs **Docker socket** on the LocalStack service, already mounted).

## Configuration

Set in `.env` or `compose.yaml` (defaults shown):

| Variable | Default | Notes |
|----------|---------|--------|
| `RDS_DB_INSTANCE_IDENTIFIER` | `user-api-rds` | |
| `RDS_MASTER_USERNAME` | `homehubadmin` | **Cannot be `postgres`** (LocalStack restriction). |
| `RDS_MASTER_PASSWORD` | `homehubpassword` | |
| `RDS_DB_NAME` | `homehub` | Avoid using `postgres` as DB name on older LocalStack. |
| `RDS_INSTANCE_CLASS` | `db.t3.micro` | |
| `RDS_ALLOCATED_STORAGE` | `20` | GB (CLI requirement) |

## Endpoint file

After the instance is **available**, the script writes:

`localstack/rds-endpoint.env` (host path; `LOCALSTACK_DATA_DIR` in the container)

Example contents: `RDS_HOST`, `RDS_PORT`, `RDS_DB_NAME`, `RDS_MASTER_USERNAME`, `DATABASE_URL`.

From the project root:

```bash
set -a && source ./localstack/rds-endpoint.env && set +a
psql "$DATABASE_URL" -c 'SELECT 1'
```

`Endpoint.Address` is often `localhost` with a **dynamic port** (not necessarily `5432`).

## Pro vs Community

Some RDS features require **LocalStack Pro**; if `create-db-instance` fails, check the container logs and [LocalStack RDS docs](https://docs.localstack.cloud/aws/services/rds/).

## Relation to `database` Compose service

The separate **`database`** service is plain Docker Postgres for local app/migrate. **RDS in LocalStack** is a second, AWS-shaped endpoint for tests that call the RDS API or connect through the emulated endpoint.
