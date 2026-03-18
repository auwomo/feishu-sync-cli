# feishu-sync

A small CLI to export Feishu/Lark content (Drive + Wiki) to local files.

- **Auth**: user or tenant
- **Scope**: all / drive / wiki
- **Output**: export to a local directory

---

## 安装 / Install

### macOS / Linux (recommended)

Default install dir: **~/.local/bin** (no sudo).

```bash
curl -fsSL https://raw.githubusercontent.com/auwomo/feishu-sync-cli/main/scripts/install.sh | sh
```

Install a specific version:

```bash
VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/auwomo/feishu-sync-cli/main/scripts/install.sh | sh
```

Uninstall:

```bash
curl -fsSL https://raw.githubusercontent.com/auwomo/feishu-sync-cli/main/scripts/install.sh | sh -s -- --uninstall
```

### Windows (PowerShell)

Installs to **$env:LOCALAPPDATA\feishu-sync\bin**.

```powershell
irm https://raw.githubusercontent.com/auwomo/feishu-sync-cli/main/scripts/install.ps1 | iex
```

Install a specific version:

```powershell
irm https://raw.githubusercontent.com/auwomo/feishu-sync-cli/main/scripts/install.ps1 | iex; install.ps1 -Version v0.1.0
```

Uninstall:

```powershell
install.ps1 -Uninstall
```

> Note: Windows binaries may not be published yet. The script will print a friendly message if none are available.


---

## 快速开始 / Quickstart

### 1) 初始化工作区 / Init workspace

```bash
feishu-sync init --app-id cli_xxx
```

This creates:

- `.feishu-sync/config.yaml`
- `.feishu-sync/secret`

> You can edit `.feishu-sync/config.yaml` anytime.

### 2) 设置 App Secret / Set secret

Preferred (safe, avoids shell history):

```bash
printf '%s' 'YOUR_APP_SECRET' | feishu-sync secret set
```

Or (less safe; may be stored in shell history):

```bash
feishu-sync secret set --value 'YOUR_APP_SECRET'
```

### 3) 登录 / Login

```bash
feishu-sync login
```

### 4) 预览/导出 / Preview & export

```bash
feishu-sync pull --dry-run
feishu-sync pull
```

---

## 配置 / Configuration

Workspace config lives at `.feishu-sync/config.yaml`.

Key fields:

- `app.id`: your app id (e.g. `cli_xxx`)
- `app.secret_env`: optional env var name for secret
- `app.secret_file`: secret file path (relative to workspace root)
- `auth.mode`: `user` (default) or `tenant`
- `scope.mode`: `all | drive | wiki`
- `output.dir`: where files are written

---

## 命令 / Commands

### `feishu-sync init`

```bash
feishu-sync init [--app-id cli_xxx] [--force] [-C DIR]
```

### `feishu-sync secret`

```bash
# set from stdin (recommended)
printf '%s' 'YOUR_APP_SECRET' | feishu-sync secret set

# set from flag (unsafe)
feishu-sync secret set --value 'YOUR_APP_SECRET'

# show (hidden by default)
feishu-sync secret show
feishu-sync secret show --reveal
```

### `feishu-sync login`

```bash
feishu-sync login
```

### `feishu-sync pull`

```bash
feishu-sync pull [--dry-run]
```

### `feishu-sync drive`

```bash
feishu-sync drive roots
feishu-sync drive ls --folder FOLDER_TOKEN [--depth N]
```

### `feishu-sync wiki`

```bash
feishu-sync wiki ls
```

### `feishu-sync validate`

```bash
feishu-sync validate
```
