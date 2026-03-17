# feishu-sync spec (draft)

## Goals

- Pure CLI (no GUI).
- Git-like workflow: `init` once, then `pull` in the workspace.
- Standardized config + stable output layout.
- Easy to wrap as an OpenClaw skill (skill calls the CLI).

## Commands

### `feishu-sync wiki ls`

Lists accessible Wiki spaces. With `--nodes`, also lists nodes under each space.

Options:
- `-C <dir>`: run as if started in `<dir>`
- `-c <file>`: explicit config file path (advanced)
- `--space <space_id>`: filter to one space
- `--nodes`: also list nodes



### `feishu-sync auth login`

User-mode OAuth login.

Options:
- `--host <ip>`: callback listen host (default `127.0.0.1`)
- `--port <int>`: callback listen port (default `18900`)
- `--callback-path <path>`: callback path (default `/callback`)

Setup:
- Add redirect URL `http://127.0.0.1:18900/callback` to your Feishu app.
- Run: `feishu-sync auth login --port 18900`

### `feishu-sync init`

Creates `.feishu-sync/` in the target directory and writes a config template.

Options (draft):
- `-C <dir>`: initialize workspace in `<dir>` (like `git -C`)
- `--out <relpath>`: default output directory (relative)
- `--force`: overwrite existing `.feishu-sync/`

### `feishu-sync pull`

Runs sync using the nearest workspace root (walk up to find `.feishu-sync/`).

Options (draft):
- `-C <dir>`: run as if started in `<dir>` (like `git -C`)
- `-c <file>`: explicit config file path (advanced)
- `--dry-run`: discover scope and print a manifest, without downloading

### `feishu-sync drive roots`

Prints guidance about Drive roots.

Notes:
- This CLI currently uses `tenant_access_token` (bot mode).
- In tenant-only mode, Feishu Drive does not expose a universal "root folder" that can be enumerated across all users.
- You must start discovery from one or more known folder tokens (configure `scope.drive_folder_tokens`).

### `feishu-sync drive ls`

Lists items in a Drive folder (and optionally recurses), to help find folder tokens and verify permissions.

Options:
- `-C <dir>`: run as if started in `<dir>`
- `-c <file>`: explicit config file path (advanced)
- `--folder <token>`: folder token to list (if omitted, uses the first configured `scope.drive_folder_tokens[0]`)
- `--depth <N>`: recursion depth (0 = only this folder; default 1)

### `feishu-sync validate`

Validates config + env + basic API access.

## Workspace layout

At workspace root:

- `.feishu-sync/`
  - `config.yaml`
  - `token.json` (OAuth token cache)
  - `state.json` (incremental state)
  - `logs/`
- `` (default output dir)
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
  secret_file: .feishu-sync/secret

scope:
  mode: all               # all | drive | wiki
  drive_folder_tokens: []
  wiki_space_ids: []

output:
  dir: .             # must be relative to workspace root

runtime:
  concurrency: 6
  rate_limit_qps: 8
  incremental: true
```

## Manifest (dry-run)

`feishu-sync pull --dry-run` prints a JSON manifest.

Drive section includes:
- `drive.roots`: the starting folder tokens used for discovery
- `drive.summary.folder_count`: how many starting folders were successfully scanned
- `drive.summary.item_count`: total items returned across scanned folders
- `drive.errors[]`: discovery errors

## Security

- `app.secret` must not be stored in config.
- Use either:
  - `app.secret_env` (recommended for CI), or
  - `app.secret_file` (local dev; default `.feishu-sync/secret`)
- Priority: env (`secret_env`) > file (`secret_file`).
- `token.json` should be `.gitignore`'d by default.
