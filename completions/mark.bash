_mark() {
  local cur prev commands
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  commands="open list show clear"

  case "$COMP_CWORD" in
    1)
      # 第一个参数：补全子命令或 .md 文件
      COMPREPLY=( $(compgen -W "$commands" -- "$cur") )
      COMPREPLY+=( $(compgen -f -X '!*.md' -- "$cur") )
      compopt -o filenames 2>/dev/null
      ;;
    2)
      # 第二个参数：补全 .md 文件
      COMPREPLY=( $(compgen -f -X '!*.md' -- "$cur") )
      compopt -o filenames 2>/dev/null
      ;;
  esac
}

complete -F _mark mark
