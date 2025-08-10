#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="anwerj"
REPO_NAME="youtube-uploader-mcp"
BINARY_BASENAME="youtube-uploader-mcp"
INSTALL_VERSION="${INSTALL_VERSION:-latest}"
LINUX_DEFAULT_BIN="$HOME/.local/bin"
MAC_DEFAULT_BIN="/usr/local/bin"

OS=""
ARCH=""
BINARY_PATH=""

info() { printf "\033[1;34m[INFO]\033[0m %s\n" "$*"; }
warn() { printf "\033[1;33m[WARN]\033[0m %s\n" "$*"; }
err()  { printf "\033[1;31m[ERR]\033[0m %s\n" "$*" >&2; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || { err "Missing required command: $1"; exit 1; }
}

detect_os_arch() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "$os" in
    Darwin) OS="darwin" ;;
    Linux)  OS="linux"  ;;
    *) err "Unsupported OS: $os"; exit 1 ;;
  esac

  case "$arch" in
    x86_64|amd64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) err "Unsupported architecture: $arch"; exit 1 ;;
  esac

  info "Detected OS/Arch: ${OS}/${ARCH}"
}

gh_release_url() {
  local os="$1" arch="$2" version="$3"
  local asset="${BINARY_BASENAME}-${os}-${arch}"

  if [ "$version" = "latest" ]; then
    echo "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${asset}"
  else
    echo "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${asset}"
  fi
}

select_install_dir() {
  if [ "$OS" = "linux" ]; then
    DEFAULT_DIR="$LINUX_DEFAULT_BIN"
  else
    DEFAULT_DIR="$MAC_DEFAULT_BIN"
  fi

  read -r -p "Install directory [${DEFAULT_DIR}]: " CHOSEN_DIR
  CHOSEN_DIR="${CHOSEN_DIR:-$DEFAULT_DIR}"
  mkdir -p "$CHOSEN_DIR"
  echo "$CHOSEN_DIR"
}

remove_quarantine_mac() {
  local path="$1"
  if [ "$OS" = "darwin" ]; then
    if command -v xattr >/dev/null 2>&1; then
      xattr -dr com.apple.quarantine "$path" || true
    fi
  fi
}

validate_client_secret() {
  local file="$1"
  if ! command -v jq >/dev/null 2>&1; then
    warn "jq not found; skipping deep validation. Proceeding with file existence check only."
    [ -s "$file" ] || { err "client secret file is empty or missing."; return 1; }
    return 0
  fi

  if ! jq -e . "$file" >/dev/null 2>&1; then
    err "Invalid JSON in client secret."
    return 1
  fi

  if jq -e 'has("installed") or has("web")' "$file" >/dev/null; then
    if jq -e '.installed.client_id or .web.client_id' "$file" >/dev/null; then
      return 0
    fi
  fi

  err "client_secret.json does not look like a Google OAuth client secret."
  return 1
}

install_binary() {
  require_cmd curl

  local url
  url="$(gh_release_url "$OS" "$ARCH" "$INSTALL_VERSION")"

  local dest_dir="$HOME/youtube-uploader-mcp"
  mkdir -p "$dest_dir"
  BINARY_PATH="$dest_dir/${BINARY_BASENAME}-${OS}-${ARCH}"

  info "Downloading $url"
  curl -fL "$url" -o "$BINARY_PATH"

  chmod +x "$BINARY_PATH"
  remove_quarantine_mac "$BINARY_PATH"
}

check_client_secret() {
  local CSRC
  CSRC="$(find . "$HOME/Downloads" -maxdepth 1 -type f -name 'client_secret*.apps.googleusercontent.com.json' | head -n 1)"

  if [ -z "$CSRC" ]; then
    err "Could not find client_secret*.apps.googleusercontent.com.json in current directory or ~/Downloads."
    exit 1
  fi

  validate_client_secret "$CSRC"

  local cfg_dir="$HOME/.config/youtube-uploader-mcp"
  mkdir -p "$cfg_dir"
  local dst="$cfg_dir/$(basename "$CSRC")"
  cp "$CSRC" "$dst"
  chmod 600 "$dst"
  echo "$dst"
}

maybe_integrate_claude() {
  echo
  read -r -p "Integrate with Claude Desktop now? [y/N]: " yn
  case "$yn" in
    [Yy]*)
      local claude_cfg
      if [ "$OS" = "darwin" ]; then
        claude_cfg="$HOME/Library/Application Support/Claude/claude_desktop_config.json"
      else
        read -r -p "Enter path to Claude config (claude_desktop_config.json): " claude_cfg
      fi

      if [ -z "${claude_cfg:-}" ]; then
        warn "No config path provided; skipping Claude integration."
        return 0
      fi

      mkdir -p "$(dirname "$claude_cfg")"
      if [ -f "$claude_cfg" ]; then
        cp "$claude_cfg" "${claude_cfg}.bak.$(date +%s)"
      fi

      if ! command -v jq >/dev/null 2>&1; then
        err "jq is required to safely edit JSON. Please install jq and re-run this step."
        if [ "$OS" = "darwin" ]; then
          info "Install jq with: brew install jq"
        elif [ "$OS" = "linux" ]; then
          info "Install jq with: sudo apt-get install jq   # or: sudo yum install jq"
        fi
        return 1
      fi

      local tmp="${claude_cfg}.tmp"
      if [ -f "$claude_cfg" ] && jq empty "$claude_cfg" >/dev/null 2>&1; then
        cp "$claude_cfg" "$tmp"
      else
        echo '{}' > "$tmp"
      fi

      local cmd="$1"
      local client_secret="$2"

      jq --arg cmd "$cmd" \
         --arg cs "$client_secret" \
         '
         .mcpServers = (.mcpServers // {}) |
         .mcpServers["youtube-uploader-mcp"] = {
           "command": $cmd,
           "args": ["-client_secret_file", $cs]
         }
         ' "$tmp" > "${tmp}.new" && mv "${tmp}.new" "$tmp"

      mv "$tmp" "$claude_cfg"
      info "Claude config updated at: $claude_cfg"
      info "If Claude is running, restart it to load the new MCP server."
      ;;
    *)
      info "Skipping Claude Desktop integration."
      ;;
  esac
}

maybe_help_custom() {
  echo
  read -r -p "Show VS Code/Cursor MCP config JSON now? [y/N]: " yn
  case "$yn" in
    [Yy]*)
      local name="youtube-uploader-mcp"
      local config
      config=$(jq -n \
        --arg cmd "$1" \
        --arg cs "$2" \
        '{
          command: $cmd,
          args: ["-client_secret_file", $cs],
          env: {},
          disabled: false,
          autoApprove: []
        }')
      echo
      info "Add the following to your MCP config:"
      echo "\"$name\": $config"
      ;;
    *)
      info "Skipping step."
      ;;
  esac
}

main() {
  info "Detecting OS and architecture..."
  detect_os_arch

  info "Checking location of client secret file"
  info "If client_secret not downloaded yet, Please watch https://youtu.be/fcywz5FIUpM for very detailed steps to download"

  local client_secret
  client_secret="$(check_client_secret)"
  info "Client secret stored at: $client_secret"

  info "Installing ${REPO_OWNER}/${REPO_NAME} (${INSTALL_VERSION})"
  install_binary
  info "Installed binary: $BINARY_PATH"

  maybe_integrate_claude "$BINARY_PATH" "$client_secret"

  maybe_help_custom "$BINARY_PATH" "$client_secret"

  echo
  info "Setup complete."
}

main "$@"
