// @ts-ignore no type declarations
import cliMarkdown from "cli-markdown";
import wrapAnsi from "wrap-ansi";

// chalk 依赖 isTTY 检测颜色支持，ink 子进程中可能为 false，强制启用
if (!process.env.FORCE_COLOR) {
  process.env.FORCE_COLOR = "1";
}

export function renderMarkdown(content: string): string {
  return cliMarkdown(content);
}

/**
 * 渲染 Markdown 并按终端宽度预换行，保证每个逻辑行 <= columns 列。
 * 解决 Ink <Text> 自动换行与分段渲染不一致的问题。
 */
export function renderMarkdownWrapped(content: string, columns: number): string[] {
  const raw = renderMarkdown(content);
  const wrapped = wrapAnsi(raw, columns, { hard: true, trim: false });
  return wrapped.split("\n");
}
