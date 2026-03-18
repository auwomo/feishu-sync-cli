# Skill: feishu-sync (template)

> Skeleton only (not published).

## Purpose

Help an agent run `feishu-sync` to export Feishu/Lark Drive + Wiki to local files, producing a manifest + ledger.

## Repo

- Repo root: `/Users/openclaw/code/feishu-sync` (adjust as needed)

## Quick commands

```bash
# build
cd /Users/openclaw/code/feishu-sync
go build ./cmd/feishu-sync

# init workspace
mkdir -p /tmp/fs-work && cd /tmp/fs-work
feishu-sync init --app-id cli_xxx
feishu-sync config
feishu-sync login

# preview
feishu-sync pull --dry-run

# export
feishu-sync pull
```

## Artifacts to report

- `<output.dir>/_meta/manifest.json`
- `<output.dir>/_meta/ledger.jsonl`
- `<output.dir>/_meta/errors-YYYYMMDD.jsonl`

## Agent reporting format

Return:

- run_id
- scope/mode/output dir
- counts (drive/wiki discovered/exported/errors)
- paths to manifest/ledger/errors
