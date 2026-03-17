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

# 3) pull (sync)
feishu-sync pull
```

## Workspace convention

`feishu-sync init` creates a folder:

- `./.feishu-sync/` — config + auth token + incremental state

By default, output is written to `./backup/` (relative to the workspace root).

## Security

- Do NOT commit `.feishu-sync/token.json` / `.feishu-sync/state.json`.
- Do NOT put app secrets in `config.yaml`.
- Secret resolution priority: env (`app.secret_env`) > file (`app.secret_file`, default `.feishu-sync/secret`).
- Use `feishu-sync secret show` to confirm which source is used.
