# feishu-sync

A minimal, git-like CLI to back up Feishu (Lark) docs (Drive + Wiki) to local files.

> Install: **TODO** (for now, build from source).

## 最短路径 (Quickstart)

### 0) 准备 Feishu 开放平台应用

你需要一个“自建应用”，并拿到：

- `app.id`（形如 `cli_xxx`）
- `app secret`

如果你打算用 **OAuth 用户登录（user mode）**，还需要在应用后台配置重定向地址：

- `http://127.0.0.1:18900/callback`（默认）

### 1) 初始化工作区

在任意空目录：

```bash
feishu-sync init
# 生成: ./.feishu-sync/config.yaml
```

### 2) 配置 secret（推荐用环境变量）

```bash
export FEISHU_APP_SECRET='***'
```

（或写入本地文件：`printf '%s' '***' | feishu-sync secret set`）

### 3) 编辑配置文件

打开 `./.feishu-sync/config.yaml`，至少改两处：

- `app.id: cli_xxx` → 你的 app id
- `auth.mode`：
  - `tenant`（默认，免登录；但 Drive 需要你提供起始 folder token）
  - `user`（需要浏览器登录；Drive 可自动发现“我的云空间”根目录）

> 默认 `output.dir: .`，输出会写到当前工作区根目录下。

### 4)（可选）OAuth 登录（仅 user mode）

```bash
# 默认监听 127.0.0.1:18900/callback
feishu-sync auth login
```

### 5) 拉取备份

```bash
# 先 dry-run 看将要同步的范围（只做发现，不下载）
feishu-sync pull --dry-run

# 真正导出到本地（默认写入 ./drive ./wiki ./_meta）
feishu-sync pull
```

---

## 配置 (config.yaml)

`feishu-sync init` 会生成一个可运行的模板（默认值如下）：

```yaml
app:
  id: cli_xxx
  secret_env: FEISHU_APP_SECRET
  secret_file: .feishu-sync/secret

auth:
  mode: tenant            # tenant | user

scope:
  mode: all               # all | drive | wiki
  drive_folder_tokens: []
  wiki_space_ids: []

output:
  dir: .                  # 相对 workspace root（默认就是 workspace root）

runtime:
  concurrency: 6
  rate_limit_qps: 8
  incremental: true
```

关键字段说明：

- `app.id`: Feishu 应用的 App ID（`cli_xxx`）
- `app.secret_env`: 从环境变量读取 app secret（推荐）
- `app.secret_file`: secret 文件路径（相对 workspace root），默认 `.feishu-sync/secret`
- `auth.mode`:
  - `user`（默认）：走 OAuth，保存 `user_access_token` 到 `./.feishu-sync/token.json`
  - `tenant`：使用 `tenant_access_token`（机器人/租户模式）
- `scope.mode`（必填）：`all | drive | wiki`
- `scope.drive_folder_tokens`：Drive 扫描起点（tenant 模式下通常必填）
- `scope.wiki_space_ids`：只同步指定知识空间（留空=全部可见空间）
- `output.dir`：输出目录（必须是相对路径；默认 `.`）

---

## 命令概览

- `feishu-sync init`
  - 初始化 `./.feishu-sync/`，写入 `config.yaml`（默认 `output.dir: .`）
- `feishu-sync secret set` / `feishu-sync secret show`
  - 管理 app secret（优先级：`secret_env` > `secret_file`）
- `feishu-sync auth login`
  - OAuth 登录（仅 `auth.mode: user` 需要）
  - 默认参数：`--host 127.0.0.1 --port 18900 --callback-path /callback`
- `feishu-sync drive roots`
  - tenant 模式下给出“如何获取 folder token”的说明
- `feishu-sync drive ls --folder <folder_token> --depth 2`
  - 递归列目录（默认 depth=1；0 表示只列当前目录）
- `feishu-sync wiki ls ...`
  - 列 Wiki（帮助见 `feishu-sync wiki ls -h`，如实现支持）
- `feishu-sync pull [--dry-run]`
  - 发现并导出（`--dry-run` 只打印 manifest JSON）
- `feishu-sync validate`
  - 校验配置（例如 `output.dir`、`scope.mode`、secret 是否可解析）

---

## 输出目录结构

默认 `output.dir: .`，所以导出会落在工作区根目录：

- `./drive/` — Drive 导出内容
- `./wiki/` — Wiki 导出内容
- `./_meta/` — 元数据
  - `manifest.json` — 本次发现/同步清单
  - `errors-YYYYMMDD.jsonl` — 导出错误日志（逐行 JSON）

`./.feishu-sync/` 只存工作区状态：

- `./.feishu-sync/config.yaml`
- `./.feishu-sync/token.json`（user mode 才会生成）
- `./.feishu-sync/secret`（如果你选择写文件）

---

## Troubleshooting

### 1) `20029 redirect_uri mismatch`

原因：Feishu 应用后台配置的 OAuth 重定向地址与本地监听地址不一致。

默认是：

- `http://127.0.0.1:18900/callback`

解决：

- 在 Feishu 开放平台应用设置里添加上述 URL
- 或使用 `feishu-sync auth login --host ... --port ... --callback-path ...` 并把对应的 URL 加到后台

### 2) `token missing` / `workspace not found`

- `workspace not found (no .feishu-sync in parents)`：你不在工作区内运行，先 `feishu-sync init`。
- user mode 下如果未登录：先 `feishu-sync auth login` 生成 `./.feishu-sync/token.json`。

### 3) Wiki 404 / 找不到节点

- 这通常是权限/可见性问题：确保该用户/应用对对应知识空间有权限。
- 如果你只想同步某些空间，填 `scope.wiki_space_ids`。

### 4) Drive 同步不到“根目录”

- `auth.mode: tenant` 下无法枚举所有人的 Drive root：必须提供 `scope.drive_folder_tokens` 作为起点。
  - `feishu-sync drive roots` 会提示如何从 Cloud Space URL 中提取 `folder_token`。
- `auth.mode: user` 下如果 `scope.drive_folder_tokens` 为空，会自动发现当前用户的 Drive roots。

### 5) `permission denied` / 导出文件失败

- 确认 `output.dir` 指向的目录可写。
- `output.dir` 必须是相对路径（安全考虑）。

---

## Security notes

- **不要提交**：`./.feishu-sync/token.json`、`./.feishu-sync/secret`、以及导出的敏感文档内容。
- 推荐用环境变量保存 secret：`FEISHU_APP_SECRET`。
- 建议在仓库里加 `.gitignore`（示例）：

```gitignore
.feishu-sync/
_drive/
_wiki/
_meta/
drive/
wiki/

# 如果你把输出放在别处，也别忘了忽略
```

---

## Remote auth ideas (设计草案，暂未实现)

当你在服务器上跑 `feishu-sync`，但需要在本机完成登录授权时，可以考虑两种思路：

1) **Public callback URL 模式**
   - 服务端：`feishu-sync auth login --listen 0.0.0.0 --port ...` 并提供 `--public-url https://.../callback`
   - 将 `redirect_uri` 指向公网可访问的回调 URL（可能需要反向代理 / TLS）
   - 风险：回调暴露在公网，必须有 state 校验、短超时、最小暴露面

2) **手动 copy-code 模式**
   - 服务端只打印授权 URL（不启动回调 HTTP server）
   - 用户在本机打开 URL 登录后拿到 `code`，再复制回服务器终端
   - 好处：不需要暴露回调端口
   - 需要 CLI 支持一个 `--paste-code`/交互输入的流程

---

## Install / distribution (TODO)

- Homebrew / Go install / Releases
- Prebuilt binaries
- Shell completion
