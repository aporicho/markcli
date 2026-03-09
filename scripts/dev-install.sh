#!/usr/bin/env bash
set -euo pipefail

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

# --- 主流程 ---

echo "🔧 MarkCLI 本地开发安装"
echo ""

# 1. 构建
BUILD_LOG=$(mktemp)
npm run build:binary > "$BUILD_LOG" 2>&1 &
if ! spinner $! "构建二进制..."; then
  echo ""
  echo "构建输出："
  cat "$BUILD_LOG"
  rm -f "$BUILD_LOG"
  exit 1
fi
rm -f "$BUILD_LOG"

# 2. 检测平台
detect_platform

BINARY="dist/binaries/mark-${OS}-${ARCH}"
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="${INSTALL_DIR}/mark"

echo "   系统: ${OS}-${ARCH}"

if [ ! -f "$BINARY" ]; then
  echo "错误：构建产物不存在: ${BINARY}"
  exit 1
fi

# 3. 安装
(mkdir -p "$INSTALL_DIR" && cp "$BINARY" "$INSTALL_PATH" && chmod +x "$INSTALL_PATH") &
spinner $! "安装到 ${INSTALL_PATH}..."
sign_macos_binary "$INSTALL_PATH"

# 4. MCP 配置（先 remove 再 add，确保指向最新）
if command -v claude &>/dev/null; then
  (claude mcp remove mark 2>/dev/null || true; claude mcp add mark -- "$INSTALL_PATH" mcp) &>/dev/null &
  spinner $! "配置 Claude Code MCP..."
else
  echo "  ⚠️  未找到 claude 命令，跳过 MCP 配置"
  echo "   安装后手动运行: claude mcp add mark -- ${INSTALL_PATH} mcp"
fi

# 5. 验证
echo ""
echo "  🔍 验证"
echo "   版本: $(mark --version)"
mark doctor

# --- 摘要 ---
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ 本地开发安装完成！"
echo ""
echo "💡 重启 Claude Code 对话即可使用"
