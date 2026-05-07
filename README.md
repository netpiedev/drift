# Drift (`@netpiedev/drift`)

Database migration CLI focused on PostgreSQL safety and developer experience.

## Features

- Ordered migration system (`.sql`, `.ts`, `.js`, `.py`, `.go`)
- Required explicit rollback files
- Safety checks (checksum, ordering, duplicate detection)
- `--dry-run` preview with SQL and lock warnings
- PostgreSQL schema diff (`drift diff --to-url ...`)
- Migration lint and doctor commands
- Seed support by environment
- Snapshot create/restore via `pg_dump` and `pg_restore`
- Dependency graph output (terminal + JSON)
- Prometheus metrics and OpenTelemetry bootstrap hooks
- npm wrapper with automatic platform binary download

## Quick Start

```bash
# Wrapper build
bun install
bun run build

# Core build
cd core
go build -o ./bin/drift ./cmd/drift

# Initialize workspace
./bin/drift init
./bin/drift make add_user_indexes
./bin/drift migrate up --dry-run
```

## CLI Commands

- `drift init`
- `drift make <name>`
- `drift migrate up`
- `drift migrate down`
- `drift migrate rollback <n>`
- `drift migrate status`
- `drift migrate verify`
- `drift migrate doctor`
- `drift migrate lint`
- `drift migrate diff --from-url <source_db_url> --to-url <target_db_url>`
- `drift migrate diff --to-url <target_db_url> --write --name add_indexes`
- `drift diff --from-url <source_db_url> --to-url <target_db_url>`
- `drift seed dev`
- `drift seed staging`
- `drift snapshot create`
- `drift snapshot restore --file <dump>`
- `drift graph tables [--json]`
- `drift graph migrations [--json]`

## Config (`drift.yaml`)

```yaml
environment: dev

database:
  url: ${DATABASE_URL}

migrations:
  dir: ./examples/migrations

safety:
  require_confirm_prod: true
  readonly_prod: false
  prod_fingerprint: ""
```

## Install From npm

```bash
npx @netpiedev/drift migrate status
bunx @netpiedev/drift migrate status
```

The wrapper resolves in this order:

1. `DRIFT_BINARY_PATH` if set
2. local binary candidates (`./core/bin/drift`, `./bin/drift`, `./drift`)
3. cached/downloaded binary from GitHub Releases for the current platform

## Quality

- Clean package boundaries
- Structured logging with `zerolog`
- SQL execution telemetry (Prometheus + OTel hooks)
- Unit and integration test scaffolding
- CI + release pipelines for Go binaries and npm publish
