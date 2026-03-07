import stringWidth from "string-width";
import type { SelectionPos, Annotation } from "../types.js";

// ---- ANSI 处理 ----

export function stripAnsi(str: string): string {
  return str.replace(/\x1b\[[0-9;]*m/g, "");
}

/** 终端列号（0-based）→ 字符索引，正确处理双宽字符 */
export function termColToCharIndex(text: string, termCol: number): number {
  let colAcc = 0;
  for (let i = 0; i < text.length; i++) {
    const charW = stringWidth(text[i]!);
    if (colAcc + charW > termCol) return i;
    colAcc += charW;
  }
  return text.length;
}

/** 字符索引 → 终端列宽（用于 lineLength） */
export function displayWidth(text: string): number {
  return stringWidth(text);
}

// ---- 位置规范化 ----

/** 比较两个 SelectionPos，返回 [start, end] 使 start <= end */
export function normalizePos(a: SelectionPos, b: SelectionPos): [SelectionPos, SelectionPos] {
  if (a.line < b.line || (a.line === b.line && a.col <= b.col)) {
    return [a, b];
  }
  return [b, a];
}

// ---- 区间计算 ----

/** 计算某行上的批注高亮区间（合并重叠） */
export function getAnnotationRangesForLine(
  annotations: Annotation[],
  lineNum: number,
  lineLength: number,
): Array<[number, number]> {
  const ranges: Array<[number, number]> = [];

  for (const ann of annotations) {
    if (lineNum < ann.startLine || lineNum > ann.endLine) continue;

    let start: number;
    let end: number;

    const hasCol = ann.startCol !== undefined && ann.endCol !== undefined;

    if (ann.startLine === ann.endLine) {
      start = hasCol ? ann.startCol! : 0;
      end = hasCol ? ann.endCol! : lineLength;
    } else if (lineNum === ann.startLine) {
      start = hasCol ? ann.startCol! : 0;
      end = lineLength;
    } else if (lineNum === ann.endLine) {
      start = 0;
      end = hasCol ? ann.endCol! : lineLength;
    } else {
      start = 0;
      end = lineLength;
    }

    ranges.push([Math.max(0, start), Math.min(end, lineLength)]);
  }

  if (ranges.length === 0) return [];

  // 合并重叠区间
  ranges.sort((a, b) => a[0] - b[0]);
  const merged: Array<[number, number]> = [ranges[0]!];
  for (let i = 1; i < ranges.length; i++) {
    const last = merged[merged.length - 1]!;
    const cur = ranges[i]!;
    if (cur[0] <= last[1]) {
      last[1] = Math.max(last[1], cur[1]);
    } else {
      merged.push(cur);
    }
  }
  return merged;
}

/** 计算某行在选择范围内的列区间 */
export function getSelectionRangeForLine(
  lineNum: number,
  lineLength: number,
  normStart: SelectionPos,
  normEnd: SelectionPos,
): [number, number] | null {
  if (lineNum < normStart.line || lineNum > normEnd.line) return null;

  const isSingleLine = normStart.line === normEnd.line;
  let s: number, e: number;

  if (isSingleLine) {
    s = normStart.col;
    e = normEnd.col;
  } else if (lineNum === normStart.line) {
    s = normStart.col;
    e = lineLength;
  } else if (lineNum === normEnd.line) {
    s = 0;
    e = normEnd.col;
  } else {
    s = 0;
    e = lineLength;
  }

  s = Math.max(0, Math.min(s, lineLength));
  e = Math.max(s, Math.min(e, lineLength));
  return e > s ? [s, e] : null;
}

// ---- 分段渲染 ----

export interface Segment {
  text: string;
  selected: boolean;
  annotated: boolean;
}

/** 将一行文本按选择和批注区间切分为样式段 */
export function buildSegments(
  stripped: string,
  selRange: [number, number] | null,
  annRanges: Array<[number, number]>,
): Segment[] {
  if (stripped.length === 0) return [{ text: " ", selected: false, annotated: false }];

  const cuts = new Set<number>();
  cuts.add(0);
  cuts.add(stripped.length);

  if (selRange) {
    cuts.add(selRange[0]);
    cuts.add(selRange[1]);
  }
  for (const [s, e] of annRanges) {
    cuts.add(s);
    cuts.add(e);
  }

  const sortedCuts = [...cuts].sort((a, b) => a - b);
  const segments: Segment[] = [];

  for (let i = 0; i < sortedCuts.length - 1; i++) {
    const start = sortedCuts[i]!;
    const end = sortedCuts[i + 1]!;
    if (start >= end) continue;

    const text = stripped.slice(start, end);
    const mid = start;

    const selected = selRange !== null && mid >= selRange[0] && mid < selRange[1];
    const annotated = annRanges.some(([s, e]) => mid >= s && mid < e);

    segments.push({ text, selected, annotated });
  }

  return segments.length > 0 ? segments : [{ text: " ", selected: false, annotated: false }];
}
