#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

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
}

sign_macos_binary() {
  local binary="$1"
  if [ "$OS" = "darwin" ]; then
    xattr -dr com.apple.quarantine "$binary" 2>/dev/null || true
    xattr -dr com.apple.provenance "$binary" 2>/dev/null || true
    codesign --force --sign - "$binary" 2>/dev/null || true
  fi
}

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

# --- 主流程 ---

detect_platform

echo "🔧 MarkCLI 本地开发安装"
echo ""

# 1. 构建
BUILD_LOG=$(mktemp)
(cd "$ROOT" && go build -o dist/mark ./cmd/mark) > "$BUILD_LOG" 2>&1 &
if ! spinner $! "构建二进制..."; then
  echo ""
  echo "构建输出："
  cat "$BUILD_LOG"
  rm -f "$BUILD_LOG"
  exit 1
fi
rm -f "$BUILD_LOG"

BINARY="$ROOT/dist/mark"
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="${INSTALL_DIR}/mark"

echo "   系统: ${OS}-$(uname -m)"

# 2. 安装
(mkdir -p "$INSTALL_DIR" && cp "$BINARY" "$INSTALL_PATH" && chmod +x "$INSTALL_PATH") &
spinner $! "安装到 ${INSTALL_PATH}..."
sign_macos_binary "$INSTALL_PATH"

# 3. Shell 补全
SHELL_NAME="${SHELL##*/}"
case "$SHELL_NAME" in
  zsh)
    LINE='eval "$(mark completion zsh)"'
    if ! grep -qF "$LINE" ~/.zshrc 2>/dev/null; then
      echo "$LINE" >> ~/.zshrc
      echo "  ✅ 补全已写入 ~/.zshrc"
    else
      echo "  ✅ 补全已配置"
    fi
    ;;
  bash)
    LINE='eval "$(mark completion bash)"'
    if ! grep -qF "$LINE" ~/.bashrc 2>/dev/null; then
      echo "$LINE" >> ~/.bashrc
      echo "  ✅ 补全已写入 ~/.bashrc"
    else
      echo "  ✅ 补全已配置"
    fi
    ;;
esac

# 4. MCP 配置
if command -v claude &>/dev/null; then
  (claude mcp remove mark --scope user 2>/dev/null || true; claude mcp add --scope user mark -- "$INSTALL_PATH" mcp) &>/dev/null &
  spinner $! "配置 Claude Code MCP..."
else
  echo "  ⚠️  未找到 claude 命令，跳过 MCP 配置"
  echo "   安装后手动运行: claude mcp add mark -- ${INSTALL_PATH} mcp"
fi

# 5. 验证
echo ""
echo "  🔍 验证"
export PATH="${INSTALL_DIR}:$PATH"
echo "   版本: $("$INSTALL_PATH" --version)"
"$INSTALL_PATH" doctor

# --- 摘要 ---
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 本地开发安装完成！"
echo ""
echo "💡 重启 Claude Code 对话即可使用"
