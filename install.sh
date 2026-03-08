#!/usr/bin/env bash
set -euo pipefail

REPO="aporicho/markcli"
BINARY_NAME="mark"

# 检测 OS
case "$(uname -s)" in
  Darwin) OS="darwin" ;;
  Linux)  OS="linux" ;;
  *)
    echo "错误：不支持的操作系统 $(uname -s)"
    exit 1
    ;;
esac

# 检测架构
case "$(uname -m)" in
  arm64|aarch64) ARCH="arm64" ;;
  x86_64|amd64)  ARCH="x64" ;;
  *)
    echo "错误：不支持的架构 $(uname -m)"
    exit 1
    ;;
esac

ASSET="mark-${OS}-${ARCH}"
echo "📦 MarkCLI 安装/更新程序"
echo "   系统: ${OS}-${ARCH}"

# 获取最新版本
echo "🔍 获取最新版本..."
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "错误：无法获取最新版本"
  exit 1
fi
echo "   版本: ${LATEST}"

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"

# ���择安装目录
if [ -d "$HOME/.local/bin" ] || mkdir -p "$HOME/.local/bin" 2>/dev/null; then
  INSTALL_DIR="$HOME/.local/bin"
else
  INSTALL_DIR="/usr/local/bin"
fi

INSTALL_PATH="${INSTALL_DIR}/${BINARY_NAME}"

# 下载
echo "⬇️  下载 ${DOWNLOAD_URL}..."
if command -v curl &>/dev/null; then
  curl -fsSL -o "$INSTALL_PATH" "$DOWNLOAD_URL"
elif command -v wget &>/dev/null; then
  wget -qO "$INSTALL_PATH" "$DOWNLOAD_URL"
else
  echo "错误：需要 curl 或 wget"
  exit 1
fi

chmod +x "$INSTALL_PATH"

# 验证
if "$INSTALL_PATH" --help &>/dev/null; then
  echo "✅ 安装成功！"
else
  echo "✅ 已安装到 ${INSTALL_PATH}"
fi

echo "   路径: ${INSTALL_PATH}"

# 检查 PATH
if ! echo "$PATH" | tr ':' '\n' | grep -q "^${INSTALL_DIR}$"; then
  echo ""
  echo "⚠️  ${INSTALL_DIR} 不在 PATH 中，请添加："
  echo "   echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc"
fi

# 配置 Claude Code MCP 集成
echo ""
if command -v claude &>/dev/null; then
  echo "🔗 配置 Claude Code MCP 集成..."
  if claude mcp add mark -- "$INSTALL_PATH" mcp 2>/dev/null; then
    echo "✅ MCP 已配置，Claude Code 可直接读取批注"
  else
    echo "⚠️  MCP 自动配置失败，请手动运行："
    echo "   claude mcp add mark -- ${INSTALL_PATH} mcp"
  fi
else
  echo "💡 配合 Claude Code 使用（安装 Claude Code 后运行）："
  echo "   claude mcp add mark -- ${BINARY_NAME} mcp"
fi

# 提示启用 shell 补全
echo ""
echo "💡 启用 Tab 补全，在 shell 配置中添加一行："
case "${SHELL##*/}" in
  zsh)  echo "   echo 'eval \"\$(mark --completions zsh)\"' >> ~/.zshrc" ;;
  bash) echo "   echo 'eval \"\$(mark --completions bash)\"' >> ~/.bashrc" ;;
  fish) echo "   mark --completions fish > ~/.config/fish/completions/mark.fish" ;;
  *)    echo "   eval \"\$(mark --completions zsh)\"  # zsh/bash/fish" ;;
esac
