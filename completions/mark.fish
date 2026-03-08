# 禁用默认文件补全，由我们自己控制
complete -c mark -f

# 子命令
complete -c mark -n '__fish_use_subcommand' -a 'open'  -d '打开文件进行阅读和批注'
complete -c mark -n '__fish_use_subcommand' -a 'list'  -d '查看文件的所有批注'
complete -c mark -n '__fish_use_subcommand' -a 'show'  -d '输出格式化的批注摘要'
complete -c mark -n '__fish_use_subcommand' -a 'clear' -d '清除所有批注'

# 第一个参数也补全 .md 文件（支持 mark README.md 省略 open）
complete -c mark -n '__fish_use_subcommand' -F -a '(__fish_complete_suffix .md)'

# 子命令后补全 .md 文件
complete -c mark -n '__fish_seen_subcommand_from open list show clear' -F -a '(__fish_complete_suffix .md)'

# 选项
complete -c mark -s V -l version -d '显示版本号'
complete -c mark -l completions -xa 'zsh bash fish' -d '输出 shell 补全脚本'
