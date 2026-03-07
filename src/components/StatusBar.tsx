import React from "react";
import { Box, Text } from "ink";
import type { AppMode } from "../types.js";

interface StatusBarProps {
  mode: AppMode;
  currentLine: number;
  totalLines: number;
  selectionStart: number | null;
  selectionEnd: number | null;
  annotationCount: number;
}

export function StatusBar({
  mode,
  currentLine,
  totalLines,
  selectionStart,
  selectionEnd,
  annotationCount,
}: StatusBarProps) {
  const modeLabel =
    mode === "reading"
      ? "阅读"
      : mode === "selecting"
        ? "选中"
        : mode === "deleting"
          ? "删除"
          : "批注";

  const modeColor =
    mode === "reading"
      ? "green"
      : mode === "selecting"
        ? "yellow"
        : mode === "deleting"
          ? "red"
          : "cyan";

  return (
    <Box flexDirection="row" justifyContent="space-between" width="100%">
      <Box>
        <Text color={modeColor} bold>
          [{modeLabel}]
        </Text>
        <Text> </Text>
        {selectionStart !== null && (
          <Text color="yellow">
            {" "}
            L{selectionStart}
            {selectionEnd !== null && selectionEnd !== selectionStart
              ? `-${selectionEnd}`
              : ""}
          </Text>
        )}
      </Box>
      <Box>
        <Text dimColor>
          {annotationCount} 批注 | 行 {currentLine}/{totalLines} |{" "}
          {mode === "reading"
            ? "v:选中 d:删除 q:退出 ↑↓:滚动"
            : mode === "selecting"
              ? "a:批注 Esc:取消 ↑↓:扩展"
              : mode === "deleting"
                ? "1-9:删除 Esc:取消"
                : "Enter:确认 Esc:取��"}
        </Text>
      </Box>
    </Box>
  );
}
