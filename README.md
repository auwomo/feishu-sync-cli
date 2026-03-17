# feishu-sync

A minimal, git-like CLI to back up Feishu (Lark) docs to local files.

## Quick start

```bash
# 1) init a workspace in any folder
feishu-sync init

# 2) set secret (choose one)
export FEISHU_APP_SECRET='***'   # recommended (CI-friendly)
# OR
printf '%s' '***' | feishu-sync secret set

# 3) choose auth mode
# tenant mode (default): uses tenant_access_token
# user mode: OAuth login to get user_access_token

# If using user mode:
#   - set auth.mode: user in .feishu-sync/config.yaml
#   - login once (opens browser)
#     Note: add redirect URL `http://127.0.0.1:18900/callback` in Feishu app settings.
feishu-sync auth login --port 18900

# 4) explore Drive folders
feishu-sync drive roots
feishu-sync drive ls --folder <folder_token> --depth 2

# 5) pull (backup)
# Dry-run prints the manifest (discovery only)
feishu-sync pull --dry-run

# Real backup exports files into ./drive/... and writes meta files:
#   _meta/manifest.json
#   _meta/errors-YYYYMMDD.jsonl
feishu-sync pull
```

## Workspace convention

`feishu-sync init` creates a folder:

- `./.feishu-sync/` — config + auth token + incremental state

By default, output is written to `./` (relative to the workspace root).

## Auth modes

### Tenant mode (default)

Uses `tenant_access_token` (bot mode).

In tenant-only mode, Feishu Drive does not expose a universal tenant-wide root folder that can be enumerated across all users.
To start Drive discovery, you must provide one or more starting folder tokens in `.feishu-sync/config.yaml`:

```yaml
scope:
  drive_folder_tokens:
    - "fldxxxxx"  # folder token
```

### User mode (OAuth)

Set:

```yaml
auth:
  mode: user
```

Then run (ensure your Feishu app has redirect URL `http://127.0.0.1:18900/callback`):

```bash
feishu-sync auth login --port 18900
```

In user mode, if `scope.drive_folder_tokens` is empty, `pull` / `pull --dry-run` will auto-discover the current user's Drive root folder.

## Security

- Do NOT commit `.feishu-sync/token.json`.
- Do NOT put app secrets in `config.yaml`.
- Secret resolution priority: env (`app.secret_env`) > file (`app.secret_file`, default `.feishu-sync/secret`).
- Use `feishu-sync secret show` to confirm which source is used.
