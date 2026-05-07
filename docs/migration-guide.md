# Migration Guide

## File naming

`<version>_<name>.<up|down>.<sql|ts|js|py|go>`

Examples:

- `001_init.up.sql`
- `001_init.down.sql`
- `202605081230_add_user_indexes.up.sql`

## Commands

- `drift init`
- `drift make add_user_indexes`
- `drift migrate up`
- `drift migrate down`
- `drift migrate rollback 2`
- `drift migrate verify`
- `drift migrate lint`
- `drift migrate doctor`
- `drift migrate status`
- `drift diff --from-url <source_url> --to-url <target_url>`
- `drift diff --to-url <target_url> --write --name sync_schema`
