#compdef mark

_mark() {
  local -a commands
  commands=(
    'open:打开文件进行阅读和批注'
    'list:查看文件的所有批注（JSON 格式）'
    'show:输出格式化的批注摘要'
    'clear:清除所有批注'
  )

  _arguments -C \
    '(-V --version)'{-V,--version}'[显示版本号]' \
    '--completions[输出 shell 补全脚本]:shell:(zsh bash fish)' \
    '1:命令或文件:->first' \
    '2:Markdown 文件:_files -g "*.md(-.)"' \
    && return

  case "$state" in
    first)
      _alternative \
        'commands:命令:((${commands}))' \
        'files:Markdown 文件:_files -g "*.md(-.)"'
      ;;
  esac
}

_mark "$@"
