# MarkCLI 测试文档

这是一个用于测试 MarkCLI 的 Markdown 文件，包含多种格式和足够长度用于滚动测试。

## 功能列表

1. **Markdown 渲染** - 在终端中漂亮地显示 Markdown
2. **文本选择** - 鼠标拖拽或键盘操作选中文本
3. **批注系统** - 对选中的文本添加批注
4. **数据持久化** - 批注保存为 JSON 文件
5. **文本锚定** - 批注绑定到文本内容而非行号

## 代码示例

```javascript
function hello() {
  console.log("Hello, MarkCLI!");
}

class MarkdownReader {
  constructor(filePath) {
    this.filePath = filePath;
    this.annotations = [];
  }

  addAnnotation(startLine, endLine, comment) {
    this.annotations.push({ startLine, endLine, comment });
  }

  render() {
    const content = fs.readFileSync(this.filePath, "utf-8");
    return marked(content);
  }
}
```

## 使用说明

使用 `mark open test.md` 打开此文件进行阅读。

- 使用 **↑↓** 键或鼠标滚轮滚动内容
- 按 **v** 进入选中模式，用方向键调整范围
- 鼠标点击并拖拽可以直接选择文本
- 选中后按 **a** 添加批注
- 按 **Esc** 取消当前选择
- 按 **q** 退出阅读器

## 设计理念

MarkCLI 的核心设计理念是**让终端成为阅读和批注的第一现场**。不同于传统的编辑器插件，MarkCLI 是一个独立的 CLI 工具，专注于阅读体验和批注管理。

### 文本锚定

传统的批注工具通常以行号标记批注位置，但这种方式在文件被编辑后就会失效。MarkCLI 采用了 **W3C Web Annotation** 标准中的文本锚定方案：

- **quote**: 被选中的原文文本
- **prefix**: 选中文本之前的上下文
- **suffix**: 选中文本之后的上下文

通过这三个字段，即使文件内容发生变化（增删行、修改段落），只要原文还在，批注就能自动重新定位到正确位置。

### 字符级选择

终端中的字符选择需要处理多种复杂情况：

1. **ANSI 转义序列** - Markdown 渲染后的文本包含颜色代码，选择时需要正确映射到原始字符位置
2. **双宽字符** - 中日韩文字在终端中占两列，鼠标列号到字符索引的转换必须考虑字符宽度
3. **分段渲染** - 同一行可能同时存在选择高亮和批注高亮，需要将行按区间切分后分别着色

### 分段渲染算法

`buildSegments()` 函数的工作流程：

1. 收集所有「切割点」——选择区间和批注区间的起止位置
2. 按位置排序去重
3. 在相邻切割点之间生成文本段
4. 为每段标记 `selected` 和 `annotated` 属性
5. 渲染时根据标记应用不同样式

## 技术栈

| 技术 | 用途 |
|------|------|
| React + Ink | 终端 UI 框架 |
| TypeScript | 类型安全 |
| marked | Markdown 解析 |
| marked-terminal | 终端渲染 |
| approx-string-match | 模糊匹配 |

## 快捷键一览

| 按键 | 阅读模式 | 选择模式 |
|------|---------|---------|
| ↑/↓ | 滚动 | 移动选择端点 |
| PgUp/PgDn | 翻页 | 翻页 |
| v | 进入选择模式 | - |
| a / Enter | - | 确认选择并批注 |
| Esc | - | 取消选择 |
| h/l | - | 左右移动列 |
| q | 退出 | - |

## 路线图

- [x] 基础 Markdown 渲染
- [x] 鼠标滚轮滚动
- [x] 键盘滚动 (↑↓/PgUp/PgDn)
- [x] 行选择（鼠标 + 键盘）
- [x] 字符级选择
- [x] 批注输入面板
- [x] JSON 持久化
- [x] 文本锚定 (quote/prefix/suffix)
- [ ] 批注列表查看
- [ ] 批注导出 (Markdown/HTML)
- [ ] 多文件批注管理
- [ ] 搜索功能
- [ ] 主题自定义

## 示例段落

以下是一些用于测试长文本滚动和选择的段落。

> 「工欲善其事，必先利其器。」—— 论语
>
> 好的工具不仅提高效率，更能改变工作方式。MarkCLI 希望成为开发者阅读和思考的好帮手。

在软件开发中，代码阅读的时间往往远超编写的时间。一个好的阅读工具应该能够帮助开发者快速理解代码结构，标记重要段落，并留下自己的思考笔记。这正是 MarkCLI 想要解决的问题。

```python
# 这是一段 Python 代码示例
def fibonacci(n):
    """计算斐波那契数列"""
    if n <= 1:
        return n
    a, b = 0, 1
    for _ in range(2, n + 1):
        a, b = b, a + b
    return b

# 测试
for i in range(10):
    print(f"fib({i}) = {fibonacci(i)}")
```

### 关于终端 UI

终端用户界面有其独特的魅力。在纯文本的世界里，每一个字符都有其位置和意义。没有复杂的布局引擎，没有 CSS 的层叠规则，有的只是行与列、前景与背景、粗体与下划线。

这种简洁反而带来了一种专注感——当你在终端里阅读一篇文章时，没有广告弹窗，没有消息通知，只有你和文字之间的对话。

## 结尾

感谢使用 MarkCLI！如果你有任何建议或发现 bug，欢迎提交 issue。

---

*此文档同时作为 MarkCLI 的功能测试文件和使用说明。*
