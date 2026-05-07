# AGENTS.md - Using `@netpiedev/drift` In Another Codebase

This guide is for LLM/code agents integrating Drift into an existing repository.

## What Drift Is

Drift is a CLI migration tool (PostgreSQL-first) with:
- Go migration engine
- npm/bun wrapper (`@netpiedev/drift`)
- explicit up/down migration pair requirements
- safety checks (ordering, checksums, rollback controls)

## Agent Goal In A Target Repo

When adding Drift to another codebase, the agent should:
1. install Drift
2. initialize migration structure
3. configure database connection safely
4. create first migration pair
5. verify/lint/preview before apply
6. apply in controlled environments

## Install Options

### Option A: npm/bun wrapper (recommended for JS/TS repos)

```bash
npm install -D @netpiedev/drift
# or
bun add -d @netpiedev/drift
```

Run via:

```bash
npx @netpiedev/drift --help
# or
bunx @netpiedev/drift --help
```

### Option B: Go binary

```bash
go install github.com/netpiedev/drift/core/cmd/drift@latest
drift --help
```

## Minimal Setup In Target Repo

From target project root:

```bash
npx @netpiedev/drift init
```

This creates baseline folders/config (`drift.yaml`, migrations, seeds, snapshots).

Set DB URL in environment:

```bash
export DATABASE_URL='postgres://user:pass@host:5432/dbname?sslmode=disable'
```

## First Migration Workflow

Create migration pair:

```bash
npx @netpiedev/drift make create_users
```

Edit generated files:
- `*.up.sql` -> forward schema changes
- `*.down.sql` -> explicit rollback SQL

Validate before apply:

```bash
npx @netpiedev/drift migrate verify
npx @netpiedev/drift migrate lint
npx @netpiedev/drift migrate up --dry-run
```

Apply:

```bash
npx @netpiedev/drift migrate up
```

## Safety Rules Agents Must Follow

1. Never ship an `up` migration without a valid `down` migration.
2. Never edit already-applied historical migrations; create a new migration instead.
3. Always run `verify` and `--dry-run` before `migrate up` in shared/prod-like envs.
4. Prefer additive schema changes first; destructive operations need explicit operator review.
5. In production, do not run rollback unless explicitly requested and confirmed.

## Operational Commands Agents Should Know

Status and validation:

```bash
npx @netpiedev/drift migrate status
npx @netpiedev/drift migrate verify
npx @netpiedev/drift migrate doctor
npx @netpiedev/drift migrate lint
```

Rollback:

```bash
npx @netpiedev/drift migrate down
npx @netpiedev/drift migrate rollback 2
```

Schema diff (DB-to-DB):

```bash
npx @netpiedev/drift diff --from-url <source> --to-url <target>
npx @netpiedev/drift diff --to-url <target> --write --name sync_schema
```

## CI Pattern For Target Repo

Recommended migration validation step in CI:

```bash
npx @netpiedev/drift migrate verify
npx @netpiedev/drift migrate lint
npx @netpiedev/drift migrate doctor
```

If CI has DB access, also run dry-run:

```bash
npx @netpiedev/drift migrate up --dry-run
```

## Troubleshooting

### Binary download 404 from wrapper
- Check package/release version alignment.
- Try explicit version:
  - `bunx @netpiedev/drift@<version> ...`
- Override release tag if needed:
  - `DRIFT_RELEASE_TAG=v0.1.4 bunx @netpiedev/drift migrate status`

### Wrong DB targeted
- Confirm `DATABASE_URL` and `drift.yaml` values before applying.
- Run `migrate status` and `--dry-run` first.

## Agent Output Expectations

When an agent performs migration tasks in another repo, it should report:
1. files created/edited
2. commands run
3. dry-run/verify/lint result
4. any destructive-risk warnings
5. exact next action for human approval (if production-impacting)
