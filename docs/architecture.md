# Drift Architecture

## Layers

- CLI (`internal/cli`): Cobra command tree and I/O
- Application (`internal/app`): app bootstrap
- Core services (`internal/migration/*`, `internal/seed`, `internal/snapshot`): migration planning, execution, safety checks, diff, lint, graphs
- Infrastructure (`internal/db`, `internal/config`, `internal/logging`, `internal/telemetry`): adapters for PostgreSQL, config, logging, observability

## Design Principles

- Modular packages (no giant files)
- Explicit dependency wiring
- No global mutable state
- Clear extension points for future DB drivers and plugin system
