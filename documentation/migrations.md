# Database migrations (Atlas)

Schema is owned by **versioned SQL** under `migrations/` (with `atlas.sum`). Apply with **[Atlas](https://atlasgo.io/)** (`atlas migrate apply`). The app uses **GORM** for queries; keep models aligned with the latest migration.

## Docker Compose

Migrations do **not** run from `docker-entrypoint-initdb.d` scripts that call Atlas there: the Postgres image has **no `migrations/` mount**, `file://migrations` would be wrong, and `atlas schema apply` with `docker://…` as dev URL cannot run Docker-in-Docker inside Postgres.

Instead, **`compose.yaml`** defines a **`migrate` service** (`arigaio/atlas:latest`) that:

1. **`depends_on`** `database` with **`condition: service_healthy`**
2. Mounts **`./migrations:/migrations`**
3. Runs **`atlas migrate apply --url … --dir file:///migrations`**

Bring the stack up (migrations run after Postgres is ready):

```bash
docker compose up -d
```

`atlas migrate apply` only applies **pending** revisions; safe to run on every compose up.

**Note:** Scripts in `/docker-entrypoint-initdb.d/` run **only on first init** (empty `pgdata`). The `migrate` service runs whenever you start compose (after healthcheck passes).

## Local / CI (without Compose)

```bash
atlas migrate apply \
  --url "postgres://homehubuser:homehubpassword@localhost:5432/homehub?sslmode=disable" \
  --dir "file://$(pwd)/migrations"
```

Use the same URL as your app. Requires the [Atlas CLI](https://atlasgo.io/getting-started) installed.

For **AWS RDS in a private VPC**, use a **bastion SSH tunnel** (or SSM port-forward) and point the URL at `127.0.0.1` and your local forward port (for example `5433`). Use **`sslmode=require`** for RDS (not `disable`). IAM database auth tokens expire in ~15 minutes. See [docs/aws-rds-connectivity-iam-bastion.md](../docs/aws-rds-connectivity-iam-bastion.md).

**Example (two terminals):**

```bash
# Terminal 1 — leave running
ssh -i ~/.ssh/your-bastion.pem -N -L 5433:your-db.xxxx.region.rds.amazonaws.com:5432 ec2-user@BASTION_PUBLIC_IP
```

```bash
# Terminal 2 — from repository root
export DATABASE_URL="postgres://DB_USER:DB_PASSWORD@127.0.0.1:5433/your_database_name?sslmode=require"
./initdb/migrations.sh
```

Or call Atlas directly: `atlas migrate apply --url "$DATABASE_URL" --dir "file://$(pwd)/migrations"`.

## Apply migrations from Go (same URL as the app)

Atlas applies SQL from disk using the **Atlas CLI** (not the GORM pool). The helper **`database.RunAtlasMigrations`** shells out to `atlas migrate apply` with the same URL as **`PostgresURLFromConfig`**, so you target the same database as GORM.

```bash
# From repo root (Atlas CLI on PATH)
go run ./cmd/atlas-apply
```

Or in code (after loading config):

```go
// import dbConfig "dfunani/homehub-profiles/src/config"
// import "dfunani/homehub-profiles/src/database"
ctx := context.Background()
cfg := dbConfig.GetDatabaseConfig()
url := database.PostgresURLFromConfig(&cfg)
if err := database.RunAtlasMigrations(ctx, url, "migrations"); err != nil {
    log.Fatal(err)
}
db := database.Connect(&cfg)
```

If you **add or edit** a file under `migrations/`, refresh checksums before `apply`:

```bash
atlas migrate hash --dir file://migrations
```

## Adding a migration

Use Atlas (e.g. `atlas migrate diff` with `atlas.hcl` and GORM models under `src/database/models`) so `migrations/` and `atlas.sum` stay in sync. After hand-editing SQL, run `atlas migrate hash` as above. Do not edit `atlas.sum` by hand unless you know the checksum format.

## Notes

- If Atlas reports a **dirty** or **checksum** mismatch, see [Atlas troubleshooting](https://atlasgo.io/versioned/troubleshoot).
- Avoid GORM `AutoMigrate` for production schema; treat SQL migrations as source of truth.
