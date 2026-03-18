# feishu-sync for agents

This doc is a **minimal, agent-friendly** guide to install and run `feishu-sync`.

## Install

From repo root:

```bash
go build ./cmd/feishu-sync
```

Or install to your GOPATH/bin:

```bash
go install ./cmd/feishu-sync
```

## Minimal run (export)

In an empty folder (workspace):

```bash
1) Initialize workspace:    feishu-sync init --app-id cli_xxx
2) Configure credentials:   feishu-sync config
3) OAuth login (user mode): feishu-sync login
4) Export:                  feishu-sync pull
```

Skill template (OpenClaw): `skills/feishu-sync/SKILL.md`

Preview manifest only (no downloads):

```bash
feishu-sync pull --dry-run
```

## Outputs (paths)

All paths are **relative to the workspace root** unless configured otherwise.

- Config: `.feishu-sync/config.yaml`
- Secret (default): `.feishu-sync/secret`
- Export output dir: `./export/` (default; see `output.dir`)
- Run artifacts (inside output dir):
  - Manifest: `<output.dir>/_meta/manifest.json`
  - Ledger (append-only JSONL): `<output.dir>/_meta/ledger.jsonl`
  - Errors (JSONL): `<output.dir>/_meta/errors-YYYYMMDD.jsonl`

## Suggested agent output format

When running `feishu-sync pull`, print a compact summary like:

- workspace: <abs path>
- mode: export | dry-run
- scope: all | drive | wiki
- output: <abs path>
- run_id: <id>
- manifest: <path>
- ledger: <path>
- errors: <path>
- counts:
  - drive: discovered=X exported=Y errors=Z
  - wiki: discovered=X exported=Y errors=Z

Avoid dumping full manifest content unless explicitly requested.
