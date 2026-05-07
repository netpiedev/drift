AGENTS.md - Drift LLM Agent Guide

Purpose
- This file tells an LLM/code agent how to work in this repository safely and productively.
- Project: @netpiedev/drift (Go migration engine + npm/bun wrapper).

Primary Goal
- Keep Drift releasable and production-safe.
- Preserve migration safety guarantees while improving developer experience.

Repository Map
- core/: Go CLI engine (Cobra, pgx, zerolog)
- npm/: TypeScript wrapper package for npx/bunx
- docs/: documentation
- examples/: sample migrations and seeds
- .github/workflows/: CI/release
- .goreleaser.yaml: binary release settings

Agent Operating Rules
1) Never break explicit rollback support.
- Every migration version must have both up and down variants.
- Validation and checksum protections must remain intact.

2) Prefer small, auditable changes.
- Patch only the required files.
- Keep behavior backward compatible unless instructed otherwise.

3) Keep npm wrapper aligned with release artifacts.
- Binary naming and download URL contract must match GoReleaser outputs.
- If release workflow changes, verify wrapper download logic still works.

4) Protect production safety features.
- Do not remove environment checks, confirmations, or readonly safeguards.

5) Do not add web UI/AI features.
- This project is CLI-only.

Standard Agent Workflow
1. Read task and identify target layer: core, npm wrapper, CI/release, docs.
2. Inspect relevant files before editing.
3. Implement focused patch.
4. Run local verification commands:
   - bun run build
   - bun run test
   - (cd core && go test ./...)
   - (cd core && go build ./cmd/drift)
5. Summarize exactly what changed and why.

Release/Publish Checklist (Agent)
- npm/package.json
  - name is correct scope/package
  - version is incremented for new publish
  - repository.url points to https://github.com/netpiedev/drift
  - publishConfig.registry is https://registry.npmjs.org/
- .github/workflows/release.yml
  - npm auth uses NODE_AUTH_TOKEN from secrets.NPM_TOKEN
  - provenance flag behavior matches repo visibility policy
- .goreleaser.yaml
  - hooks and paths work from repo root
  - build artifacts match npm wrapper expectations

Migration Safety Checklist (Agent)
- parser/validate still enforces up/down pairs.
- verify still checks order and checksum integrity.
- rollback logic still removes applied state correctly.
- dry-run output still reports warnings/locks/tables.

Testing Notes
- Integration tests may require DRIFT_TEST_DATABASE_URL.
- CI provides postgres service; keep tests deterministic.

Expected Response Format for Agent PRs
- What changed
- Why it changed
- Validation performed
- Any follow-up risks or actions

Out of Scope
- No database dashboard
- No hosted service
- No unrelated refactors

