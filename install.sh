#!/usr/bin/env bash
set -euo pipefail

REPO="aporicho/markcli"
BINARY_NAME="mark"

# --- 辅助函数 ---

detect_platform() {
  case "$(uname -s)" in
    Darwin) OS="darwin" ;;
    Linux)  OS="linux" ;;
    *)
      echo "错误：不支持的操作系统 $(uname -s)"
      exit 1
      ;;
  esac

  case "$(uname -m)" in
    arm64|aarch64) ARCH="arm64" ;;
    x86_64|amd64)  ARCH="x64" ;;
    *)
      echo "错误：不支持的架构 $(uname -m)"
      exit 1
      ;;
  esac
}

sign_macos_binary() {
  local binary="$1"
  if [ "$OS" = "darwin" ]; then
    xattr -dr com.apple.quarantine "$binary" 2>/dev/null || true
    xattr -dr com.apple.provenance "$binary" 2>/dev/null || true
    codesign --force --sign - "$binary" 2>/dev/null || true
  fi
}

get_shell_config() {
  case "${SHELL##*/}" in
    zsh)  echo "$HOME/.zshrc" ;;
    bash) echo "$HOME/.bashrc" ;;
    fish) echo "$HOME/.config/fish/config.fish" ;;
    *)    echo "$HOME/.bashrc" ;;
  esac
}

# 向 shell 配置文件追加一行（如果不存在）
# 返回 0=新增，1=已存在
ensure_line_in_config() {
  local config_file="$1"
  local line="$2"

  if [ -f "$config_file" ] && grep -qF "$line" "$config_file"; then
    return 1
  fi

  mkdir -p "$(dirname "$config_file")"
  echo "$line" >> "$config_file"
  return 0
}

# Braille spinner 动画
spinner() {
  local pid=$1 msg=$2
  local frames='⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
  local i=0
  while kill -0 "$pid" 2>/dev/null; do
    printf "\r  %s %s" "${frames:i++%${#frames}:1}" "$msg"
    sleep 0.1
  done
  wait "$pid"
  local rc=$?
  printf "\r  %s %s\n" "$([ $rc -eq 0 ] && echo '✅' || echo '❌')" "$msg"
  return $rc
}

# 后台运行命令并显示 spinner
run_with_spinner() {
  local msg=$1
  shift
  "$@" &>/dev/null &
  spinner $! "$msg"
}

# 跨平台获取文件大小
file_size() {
  stat -f%z "$1" 2>/dev/null || stat -c%s "$1" 2>/dev/null || echo 0
}

# Unicode 进度条下载
download_with_progress() {
  local url=$1 output=$2 label=$3
  local total
  total=$(curl -fsSLI "$url" 2>/dev/null | grep -i '^content-length:' | tail -1 | tr -dc '0-9')

  if [ -z "$total" ] || [ "${total:-0}" -eq 0 ] 2>/dev/null; then
    # 无法获取大小，降级为 spinner
    curl -fsSL -o "$output" "$url" &
    spinner $! "$label"
    return
  fi

  curl -fsSL -o "$output" "$url" &
  local pid=$!
  local blocks='▏▎▍▌▋▊▉█'
  local width=30

  while kill -0 "$pid" 2>/dev/null; do
    local current
    current=$(file_size "$output")
    if [ "${current:-0}" -gt 0 ] 2>/dev/null; then
      local pct=$((current * 100 / total))
      if [ "$pct" -gt 100 ]; then pct=100; fi
      local filled=$((current * width * 8 / total))
      if [ "$filled" -gt $((width * 8)) ]; then filled=$((width * 8)); fi
      local full=$((filled / 8))
      local frac=$((filled % 8))
      local bar=""
      for ((j=0; j<full; j++)); do bar+="█"; done
      if [ "$frac" -gt 0 ] && [ "$full" -lt "$width" ]; then
        bar+="${blocks:frac-1:1}"
        full=$((full + 1))
      fi
      local empty=$((width - full))
      for ((j=0; j<empty; j++)); do bar+=" "; done
      printf "\r  ⬇️  %s [%s] %3d%%" "$label" "$bar" "$pct"
    fi
    sleep 0.1
  done

  wait "$pid"
  local rc=$?
  if [ $rc -eq 0 ]; then
    local bar=""
    for ((j=0; j<width; j++)); do bar+="█"; done
    printf "\r  ✅ %s [%s] 100%%\n" "$label" "$bar"
  else
    printf "\r  ❌ %s\n" "$label"
  fi
  return $rc
}

# --- 主流程 ---

detect_platform

ASSET="mark-${OS}-${ARCH}"
SHELL_NAME="${SHELL##*/}"
SHELL_CONFIG=$(get_shell_config)

echo "📦 MarkCLI 安装程序"
echo "   系统: ${OS}-${ARCH}"

# 检测已安装版本
CURRENT_VERSION=""
if command -v "$BINARY_NAME" &>/dev/null; then
  CURRENT_VERSION=$("$BINARY_NAME" --version 2>/dev/null || echo "")
fi

# 获取最新版本
LATEST_FILE=$(mktemp)
(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' > "$LATEST_FILE") &
spinner $! "获取最新版本..." || true
LATEST=$(cat "$LATEST_FILE")
rm -f "$LATEST_FILE"
if [ -z "$LATEST" ]; then
  echo "错误：无法获取最新版本"
  exit 1
fi

IS_UPDATE=false
if [ -n "$CURRENT_VERSION" ]; then
  echo "   当前: ${CURRENT_VERSION} → 最新: ${LATEST}"
  IS_UPDATE=true
else
  echo "   版本: ${LATEST}"
fi

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"

# 安装目录
if [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
  INSTALL_DIR="$HOME/.local/bin"
else
  INSTALL_DIR="/usr/local/bin"
fi
INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"

# 下载（带进度条）
download_with_progress "$DOWNLOAD_URL" "$INSTALL_PATH" "$ASSET"

chmod +x "$INSTALL_PATH"
sign_macos_binary "$INSTALL_PATH"

# PATH 配置
PATH_ADDED=false
printf "  ⠋ 配置 PATH...\r"
if ! echo "$PATH" | tr ':' '\n' | grep -q "^${INSTALL_DIR}$"; then
  if [ "$SHELL_NAME" = "fish" ]; then
    PATH_LINE="fish_add_path ${INSTALL_DIR}"
  else
    PATH_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""
  fi
  if ensure_line_in_config "$SHELL_CONFIG" "$PATH_LINE"; then
    PATH_ADDED=true
  fi
fi
printf "\r  ✅ 配置 PATH       \n"

# Tab 补全
COMPLETION_ADDED=false
printf "  ⠋ 配置补全...\r"
case "$SHELL_NAME" in
  zsh)
    # shellcheck disable=SC2016
    if ensure_line_in_config "$SHELL_CONFIG" 'eval "$(mark completion zsh)"'; then
      COMPLETION_ADDED=true
    fi
    ;;
  bash)
    # shellcheck disable=SC2016
    if ensure_line_in_config "$SHELL_CONFIG" 'eval "$(mark completion bash)"'; then
      COMPLETION_ADDED=true
    fi
    ;;
  fish)
    FISH_COMP_DIR="$HOME/.config/fish/completions"
    mkdir -p "$FISH_COMP_DIR"
    if "$INSTALL_PATH" completion fish > "$FISH_COMP_DIR/mark.fish" 2>/dev/null; then
      COMPLETION_ADDED=true
    fi
    ;;
esac
printf "\r  ✅ 配置补全       \n"

# Claude Code MCP 配置
MCP_CONFIGURED=false
MCP_SKIPPED=false
if command -v claude &>/dev/null; then
  claude mcp add --scope user mark -- "$INSTALL_PATH" mcp &>/dev/null &
  if spinner $! "配置 MCP..."; then
    MCP_CONFIGURED=true
  fi
else
  MCP_SKIPPED=true
fi

# 验证安装
export PATH="${INSTALL_DIR}:$PATH"
"$INSTALL_PATH" --version &>/dev/null &
spinner $! "验证安装..."
INSTALLED_VERSION=$("$INSTALL_PATH" --version 2>/dev/null || echo "未知")

# --- 摘要 ---
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ "$IS_UPDATE" = true ]; then
  echo "✅ 更新完成: ${CURRENT_VERSION} → ${INSTALLED_VERSION}"
else
  echo "✅ 安装完成: ${INSTALLED_VERSION}"
fi
echo "   路径: ${INSTALL_PATH}"

NOTES=()

if [ "$PATH_ADDED" = true ]; then
  NOTES+=("PATH 已写入 ${SHELL_CONFIG}，运行 source ${SHELL_CONFIG} 或重开终端生效")
fi

if [ "$COMPLETION_ADDED" = true ]; then
  if [ "$SHELL_NAME" = "fish" ]; then
    NOTES+=("Fish 补全已安装")
  else
    NOTES+=("Tab 补全已写入 ${SHELL_CONFIG}")
  fi
fi

if [ "$MCP_CONFIGURED" = true ]; then
  NOTES+=("MCP 已配置，重启 Claude Code 对话生效")
fi

if [ "$MCP_SKIPPED" = true ]; then
  NOTES+=("配合 Claude Code 使用: claude mcp add mark -- ${BINARY_NAME} mcp")
fi

if [ ${#NOTES[@]} -gt 0 ]; then
  echo ""
  for note in "${NOTES[@]}"; do
    echo "💡 ${note}"
  done
fi
