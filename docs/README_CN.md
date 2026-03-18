# feishu-sync（中文文档）

一个小型 CLI，用于将飞书 / Lark 的内容（云空间 Drive + 知识库 Wiki）导出到本地文件。

- **认证**：user 或 tenant
- **范围**：all / drive / wiki
- **输出**：导出到本地目录

---

## Install

### macOS / Linux（bash）

默认安装目录：**~/.local/bin**（无需 sudo）。

安装：

```bash
curl -fsSL https://raw.githubusercontent.com/auwomo/feishu-sync-cli/main/scripts/install.sh | sh
```

卸载：

```bash
curl -fsSL https://raw.githubusercontent.com/auwomo/feishu-sync-cli/main/scripts/install.sh | sh -s -- --uninstall
```

### Windows（PowerShell）

默认安装到：**$env:LOCALAPPDATA\feishu-sync\bin**。

安装：

```powershell
irm https://raw.githubusercontent.com/auwomo/feishu-sync-cli/main/scripts/install.ps1 | iex
```

卸载：

```powershell
install.ps1 -Uninstall
```

> 说明：Windows 二进制文件可能尚未发布；如果没有可用版本，脚本会输出友好提示。

---

## Quickstart

### 1) 初始化工作区

```bash
feishu-sync init --app-id cli_xxx
```

会创建：

- `.feishu-sync/config.yaml`
- `.feishu-sync/secret`

> 你可以随时编辑 `.feishu-sync/config.yaml`。

### 2) 配置凭证

```bash
feishu-sync config
```

### 3) 登录

```bash
feishu-sync login
```

### 4) 预览 / 导出

```bash
feishu-sync pull --dry-run
feishu-sync pull
```

---

## 命令概览

### `feishu-sync init`

```bash
feishu-sync init [--app-id cli_xxx] [--force] [-C DIR]
```

### `feishu-sync config`

```bash
feishu-sync config
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

---

## Troubleshooting

- **找不到命令 / Command not found**：确认安装目录在 PATH 中（macOS/Linux：`~/.local/bin`；Windows：`$env:LOCALAPPDATA\feishu-sync\bin`）。
- **Windows 没有可用版本**：脚本会提示当前没有发布 Windows 二进制包；可以先在 macOS/Linux 使用，或自行从源码编译。
- **权限 / 授权失败**：检查 `.feishu-sync/config.yaml` 中的 `auth.mode`、`scope.mode` 是否正确；确认 App 权限与可访问范围匹配。
- **导出结果不符合预期**：建议先运行 `feishu-sync pull --dry-run` 预览将要导出的内容。
