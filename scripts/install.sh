#!/usr/bin/env sh
set -eu

REPO="auwomo/feishu-sync-cli"
BIN_NAME="feishu-sync"

# Defaults
: "${INSTALL_DIR:=$HOME/.local/bin}"
: "${VERSION:=latest}"

# --- UI ---
BOLD='\033[1m'
DIM='\033[2m'
GREEN='\033[32m'
YELLOW='\033[33m'
RED='\033[31m'
RESET='\033[0m'

say() { printf "%s\n" "$*"; }
info() { say "${DIM}$*${RESET}"; }
ok() { say "${GREEN}$*${RESET}"; }
warn() { say "${YELLOW}$*${RESET}"; }
err() { say "${RED}error:${RESET} $*"; }

usage() {
  cat <<EOF
${BOLD}Usage${RESET}:
  install.sh [--uninstall] [--dir DIR]

${BOLD}Env${RESET}:
  VERSION=latest|vX.Y.Z   (default: latest)
  INSTALL_DIR=~/.local/bin (default)

${BOLD}Examples${RESET}:
  curl -fsSL https://raw.githubusercontent.com/$REPO/main/scripts/install.sh | sh
  VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/$REPO/main/scripts/install.sh | sh
  curl -fsSL https://raw.githubusercontent.com/$REPO/main/scripts/install.sh | sh -s -- --uninstall
EOF
}

need_cmd() { command -v "$1" >/dev/null 2>&1; }

http_get() {
  url="$1"
  if need_cmd curl; then
    curl -fsSL "$url"
  elif need_cmd wget; then
    wget -qO- "$url"
  else
    err "missing curl or wget"
    exit 2
  fi
}

download_to() {
  url="$1"; out="$2"
  if need_cmd curl; then
    curl -fL --retry 3 --retry-delay 1 -o "$out" "$url"
  else
    wget -qO "$out" "$url"
  fi
}

os() {
  u=$(uname -s 2>/dev/null || echo unknown)
  case "$u" in
    Darwin) echo darwin;;
    Linux) echo linux;;
    *) echo unknown;;
  esac
}

arch() {
  u=$(uname -m 2>/dev/null || echo unknown)
  case "$u" in
    x86_64|amd64) echo amd64;;
    aarch64|arm64) echo arm64;;
    *) echo unknown;;
  esac
}

installed_version() {
  if [ -x "$INSTALL_DIR/$BIN_NAME" ]; then
    "$INSTALL_DIR/$BIN_NAME" version 2>/dev/null | awk '{print $NF}' | head -n1 || true
  else
    true
  fi
}

resolve_version() {
  if [ "$VERSION" != "latest" ]; then
    echo "$VERSION"
    return
  fi
  tag=$(http_get "https://api.github.com/repos/$REPO/releases/latest" \
    | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n1)
  if [ -z "${tag:-}" ]; then
    err "failed to resolve latest version from GitHub"
    exit 3
  fi
  echo "$tag"
}

pick_asset() {
  tag="$1"; os_="$2"; arch_="$3"

  # common naming patterns, try most specific first
  candidates="
${BIN_NAME}_${tag#v}_${os_}_${arch_}.tar.gz
${BIN_NAME}_${os_}_${arch_}.tar.gz
${BIN_NAME}-${tag#v}-${os_}-${arch_}.tar.gz
${BIN_NAME}-${os_}-${arch_}.tar.gz
${BIN_NAME}_${os_}_${arch_}
"

  rel_json=$(http_get "https://api.github.com/repos/$REPO/releases/tags/$tag")

  for name in $candidates; do
    url=$(printf "%s" "$rel_json" | sed -n "s#.*\"browser_download_url\"[[:space:]]*:[[:space:]]*\"\([^\"]*${name//./\\.}[^\"]*\)\".*#\1#p" | head -n1)
    if [ -n "${url:-}" ]; then
      say "$url"
      return
    fi
  done

  # fallback: find first asset containing os + arch
  url=$(printf "%s" "$rel_json" | sed -n "s#.*\"browser_download_url\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*#\1#p" \
    | grep -E "${os_}.*${arch_}" \
    | head -n1 || true)
  if [ -n "${url:-}" ]; then
    say "$url"
    return
  fi

  return 1
}

verify_checksum() {
  file="$1"; sums="$2"
  base=$(basename "$file")
  expected=$(grep -E "[[:xdigit:]]{64}[[:space:]]+\*?${base}$" "$sums" 2>/dev/null | awk '{print $1}' | head -n1 || true)
  if [ -z "${expected:-}" ]; then
    warn "no checksum entry found for ${BOLD}${base}${RESET} (skipping verification)"
    return 0
  fi

  if need_cmd sha256sum; then
    actual=$(sha256sum "$file" | awk '{print $1}')
  elif need_cmd shasum; then
    actual=$(shasum -a 256 "$file" | awk '{print $1}')
  else
    warn "sha256sum/shasum not found; skipping checksum verification"
    return 0
  fi

  if [ "$expected" != "$actual" ]; then
    err "checksum mismatch for ${base}"
    exit 4
  fi
  ok "checksum verified"
}

uninstall() {
  target="$INSTALL_DIR/$BIN_NAME"
  if [ -f "$target" ]; then
    rm -f "$target"
    ok "removed ${BOLD}${target}${RESET}"
  else
    info "not installed: ${BOLD}${target}${RESET}"
  fi
}

main() {
  action="install"

  while [ $# -gt 0 ]; do
    case "$1" in
      -h|--help) usage; exit 0;;
      --uninstall) action="uninstall"; shift;;
      --dir) INSTALL_DIR="$2"; shift 2;;
      *) err "unknown arg: $1"; usage; exit 2;;
    esac
  done

  if [ "$action" = "uninstall" ]; then
    uninstall
    exit 0
  fi

  os_=$(os); arch_=$(arch)
  if [ "$os_" = "unknown" ] || [ "$arch_" = "unknown" ]; then
    err "unsupported platform: $(uname -s)/$(uname -m)"
    exit 2
  fi

  tag=$(resolve_version)

  current=$(installed_version || true)
  if [ -n "${current:-}" ] && [ "$current" = "$tag" ]; then
    ok "${BIN_NAME} ${BOLD}${tag}${RESET} already installed at ${BOLD}${INSTALL_DIR}${RESET}"
    exit 0
  fi

  url=$(pick_asset "$tag" "$os_" "$arch_" || true)
  if [ -z "${url:-}" ]; then
    err "no release asset found for ${os_}/${arch_} (${tag})"
    exit 3
  fi

  tmp=$(mktemp -d 2>/dev/null || mktemp -d -t feishu-sync)
  trap 'rm -rf "$tmp"' EXIT

  archive="$tmp/asset"
  sums="$tmp/checksums.txt"

  info "version: ${BOLD}${tag}${RESET}"
  info "install dir: ${BOLD}${INSTALL_DIR}${RESET}"
  info "download: ${DIM}${url}${RESET}"

  download_to "$url" "$archive"

  # checksums
  sums_url="https://github.com/$REPO/releases/download/$tag/checksums.txt"
  if download_to "$sums_url" "$sums" 2>/dev/null; then
    verify_checksum "$archive" "$sums"
  else
    warn "checksums.txt not found; skipping checksum verification"
  fi

  mkdir -p "$INSTALL_DIR"

  extracted="$tmp/$BIN_NAME"
  if file "$archive" 2>/dev/null | grep -qi "tar"; then
    tar -xzf "$archive" -C "$tmp"
    if [ -x "$tmp/$BIN_NAME" ]; then
      extracted="$tmp/$BIN_NAME"
    else
      # try find first matching binary
      extracted=$(find "$tmp" -maxdepth 2 -type f -name "$BIN_NAME" -perm -111 | head -n1 || true)
    fi
  else
    mv "$archive" "$extracted"
    chmod +x "$extracted" || true
  fi

  if [ ! -x "${extracted}" ]; then
    err "failed to extract ${BIN_NAME} from asset"
    exit 5
  fi

  install_path="$INSTALL_DIR/$BIN_NAME"
  mv "$extracted" "$install_path"
  chmod +x "$install_path" || true

  ok "installed ${BOLD}${BIN_NAME}${RESET} to ${BOLD}${install_path}${RESET}"

  if ! printf "%s" ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
    warn "${BOLD}${INSTALL_DIR}${RESET} is not on PATH"
    say "  add this to your shell profile:"
    say "    export PATH=\"$INSTALL_DIR:\$PATH\""
  fi

  exit 0
}

main "$@"
