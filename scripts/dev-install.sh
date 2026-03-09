#!/usr/bin/env bash
set -euo pipefail

echo "🔧 MarkCLI 本地开发安装"
echo ""

# 1. 构建二进制（当前平台）
echo "📦 构建二进制..."
npm run build:binary

# 2. 检测 OS / ARCH（复用 install.sh 逻辑）
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

BINARY="dist/binaries/mark-${OS}-${ARCH}"
INSTALL_DIR="$HOME/.local/bin"
INSTALL_PATH="${INSTALL_DIR}/mark"

echo "   系统: ${OS}-${ARCH}"
echo "   二进制: ${BINARY}"

if [ ! -f "$BINARY" ]; then
  echo "错误：构建产物不存在: ${BINARY}"
  exit 1
fi

# 3. 安装（和 install.sh 完全一致）
echo ""
echo "📥 安装到 ${INSTALL_PATH}..."
mkdir -p "$INSTALL_DIR"
cp "$BINARY" "$INSTALL_PATH"
chmod +x "$INSTALL_PATH"

# macOS: 清除隔离属性并重新签名
if [ "$OS" = "darwin" ]; then
  xattr -dr com.apple.quarantine "$INSTALL_PATH" 2>/dev/null || true
  xattr -dr com.apple.provenance "$INSTALL_PATH" 2>/dev/null || true
  codesign --force --sign - "$INSTALL_PATH" 2>/dev/null || true
fi

echo "✅ 已安装到 ${INSTALL_PATH}"

# 4. MCP 配置（先 remove 再 add，确保指向最新）
echo ""
echo "🔗 配置 Claude Code MCP..."
if command -v claude &>/dev/null; then
  claude mcp remove mark 2>/dev/null || true
  claude mcp add mark -- "$INSTALL_PATH" mcp
  echo "✅ MCP 已配置"
else
  echo "⚠️  未找到 claude 命令，跳过 MCP 配置"
  echo "   安装后手动运行: claude mcp add mark -- ${INSTALL_PATH} mcp"
fi

# 5. 验证
echo ""
echo "🔍 验证..."
echo "   版本: $(mark --version)"
mark doctor

echo ""
echo "✅ 本地开发安装完成！重启 Claude Code 对话即可使用。"
