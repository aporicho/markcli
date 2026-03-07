// LineSelector is integrated into the main app logic
// This module exports selection helper utilities

export function getSelectedText(
  lines: string[],
  startLine: number,
  endLine: number,
  startCol?: number,
  endCol?: number,
): string {
  // Lines are 1-based
  if (startLine === endLine) {
    const line = lines[startLine - 1] ?? "";
    if (startCol !== undefined && endCol !== undefined) {
      return line.slice(startCol, endCol);
    }
    return line;
  }

  const result: string[] = [];
  for (let i = startLine; i <= endLine; i++) {
    const line = lines[i - 1] ?? "";
    if (i === startLine && startCol !== undefined) {
      result.push(line.slice(startCol));
    } else if (i === endLine && endCol !== undefined) {
      result.push(line.slice(0, endCol));
    } else {
      result.push(line);
    }
  }
  return result.join("\n");
}
