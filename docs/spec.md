# feishu-sync spec (draft)

## Goals

- Pure CLI (no GUI).
- Git-like workflow: `init` once, then `pull` in the workspace.
- Standardized config + stable output layout.
- Easy to wrap as an OpenClaw skill (skill calls the CLI).

## Commands

### `feishu-sync init`

Creates `.feishu-sync/` in the current directory and writes a config template.

Options (draft):
- `--out <relpath>`: default output directory (relative)
- `--force`: overwrite existing `.feishu-sync/`

### `feishu-sync pull`

Runs sync using the nearest workspace root (walk up to find `.feishu-sync/`).

Options (draft):
- `-C <dir>`: run as if started in `<dir>` (like `git -C`)
- `-c <file>`: explicit config file path (advanced)

### `feishu-sync validate`

Validates config + env + basic API access.

## Workspace layout

At workspace root:

- `.feishu-sync/`
  - `config.yaml`
  - `token.json` (OAuth token cache)
  - `state.json` (incremental state)
  - `logs/`
- `backup/` (default output dir)
  - `drive/`
  - `wiki/`
  - `_meta/`
    - `manifest.json`
    - `errors-YYYYMMDD.jsonl`

## Config schema (config.yaml)

```yaml
app:
  id: cli_xxx
  secret_env: FEISHU_APP_SECRET

scope:
  mode: all               # all | drive | wiki
  drive_folder_tokens: []
  wiki_space_ids: []

output:
  dir: backup             # must be relative to workspace root

runtime:
  concurrency: 6
  rate_limit_qps: 8
  incremental: true
```

## Security

- `app.secret` must not be stored in config.
- `token.json` should be `.gitignore`'d by default.
