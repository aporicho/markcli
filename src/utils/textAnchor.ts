import search from "approx-string-match";

const CONTEXT_CHARS = 30;

export interface TextAnchor {
  quote: string;
  prefix: string;
  suffix: string;
}

export interface TextRange {
  start: number; // offset in fullText
  end: number;   // exclusive
}

// ---- 偏移量 ↔ 行列 转换 ----

/** 将 (line, col) 转为拼接文本中的字符偏移。line 为 1-based，col 为 0-based。 */
export function lineColToOffset(lineLengths: number[], line: number, col: number): number {
  let offset = 0;
  for (let i = 0; i < line - 1 && i < lineLengths.length; i++) {
    offset += lineLengths[i]! + 1; // +1 for \n
  }
  return offset + col;
}

/** 将拼接文本中的字符偏移转为 (line, col)。返回的 line 为 1-based，col 为 0-based。 */
export function offsetToLineCol(lineLengths: number[], offset: number): { line: number; col: number } {
  let remaining = offset;
  for (let i = 0; i < lineLengths.length; i++) {
    const len = lineLengths[i]! + 1; // +1 for \n
    if (remaining < len || i === lineLengths.length - 1) {
      return { line: i + 1, col: Math.min(remaining, lineLengths[i]!) };
    }
    remaining -= len;
  }
  return { line: 1, col: 0 };
}

// ---- 锚定提取 ----

/** 从全文中提取文本锚定 */
export function extractAnchor(fullText: string, start: number, end: number): TextAnchor {
  return {
    quote: fullText.slice(start, end),
    prefix: fullText.slice(Math.max(0, start - CONTEXT_CHARS), start),
    suffix: fullText.slice(end, Math.min(fullText.length, end + CONTEXT_CHARS)),
  };
}

// ---- 锚定重定位 ----

/** 在全文中重新定位锚定，返回匹配位置 */
export function relocateAnchor(fullText: string, anchor: TextAnchor): TextRange | null {
  const { quote, prefix, suffix } = anchor;
  if (!quote) return null;

  // 策略 1：精确搜索 quote
  const exactMatches: number[] = [];
  let idx = 0;
  while ((idx = fullText.indexOf(quote, idx)) !== -1) {
    exactMatches.push(idx);
    idx++;
  }

  if (exactMatches.length === 1) {
    return { start: exactMatches[0]!, end: exactMatches[0]! + quote.length };
  }

  if (exactMatches.length > 1) {
    // 策略 2：用 prefix/suffix 消歧
    let bestIdx = exactMatches[0]!;
    let bestScore = -1;

    for (const matchIdx of exactMatches) {
      let score = 0;
      if (prefix) {
        const before = fullText.slice(Math.max(0, matchIdx - prefix.length), matchIdx);
        score += commonSuffixLength(before, prefix);
      }
      if (suffix) {
        const after = fullText.slice(matchIdx + quote.length, matchIdx + quote.length + suffix.length);
        score += commonPrefixLength(after, suffix);
      }
      if (score > bestScore) {
        bestScore = score;
        bestIdx = matchIdx;
      }
    }

    return { start: bestIdx, end: bestIdx + quote.length };
  }

  // 策略 3：模糊匹配（允许 ~20% 误差）
  const maxErrors = Math.max(1, Math.floor(quote.length * 0.2));
  const matches = search(fullText, quote, maxErrors);

  if (matches.length > 0) {
    // 选择误差最少的匹配
    const best = matches.reduce((a, b) => (a.errors < b.errors ? a : b));
    return { start: best.start, end: best.end };
  }

  return null;
}

// ---- 辅助 ----

function commonSuffixLength(a: string, b: string): number {
  let n = 0;
  const max = Math.min(a.length, b.length);
  for (let i = 0; i < max; i++) {
    if (a[a.length - 1 - i] === b[b.length - 1 - i]) n++;
    else break;
  }
  return n;
}

function commonPrefixLength(a: string, b: string): number {
  let n = 0;
  const max = Math.min(a.length, b.length);
  for (let i = 0; i < max; i++) {
    if (a[i] === b[i]) n++;
    else break;
  }
  return n;
}
