import React, { useState, useCallback, useEffect, useRef } from "react";
import { Box, Text } from "ink";
import { useMouse } from "../hooks/useMouse.js";
import { useSelection } from "../hooks/useSelection.js";
import { useKeyboard } from "../hooks/useKeyboard.js";
import {
  stripAnsi,
  getAnnotationRangesForLine,
  getSelectionRangeForLine,
  buildSegments,
} from "../utils/ranges.js";
import type { Annotation } from "../types.js";

export interface ViewerStatus {
  scrollOffset: number;
  selecting: boolean;
  selStartLine?: number;
  selEndLine?: number;
}

interface MarkdownViewerProps {
  lines: string[];
  viewportHeight: number;
  annotations: Annotation[];
  active: boolean;
  onSelect: (startLine: number, endLine: number, startCol: number, endCol: number) => void;
  onQuit: () => void;
  onStatusChange?: (status: ViewerStatus) => void;
  onDeleteMode?: () => void;
}

export function MarkdownViewer({
  lines,
  viewportHeight,
  annotations,
  active,
  onSelect,
  onQuit,
  onStatusChange,
  onDeleteMode,
}: MarkdownViewerProps) {
  const [scrollOffset, setScrollOffset] = useState(0);
  const maxOffset = Math.max(0, lines.length - viewportHeight);

  // resize 时修正 scrollOffset
  const prevLinesLenRef = useRef(lines.length);
  useEffect(() => {
    if (prevLinesLenRef.current !== lines.length) {
      prevLinesLenRef.current = lines.length;
      setScrollOffset((prev) => Math.min(prev, Math.max(0, lines.length - viewportHeight)));
    }
  }, [lines.length, viewportHeight]);

  // ---- 滚动 ----
  const scrollUp = useCallback(
    (n = 1) => setScrollOffset((p) => Math.max(0, p - n)),
    [],
  );
  const scrollDown = useCallback(
    (n = 1) => setScrollOffset((p) => Math.min(maxOffset, p + n)),
    [maxOffset],
  );

  // ---- 选择 ----
  const selection = useSelection({
    lines,
    viewportHeight,
    scrollOffset,
    onConfirm: onSelect,
  });

  // ---- 鼠标 ----
  const handleMouseEvent = useCallback(
    (event: import("../hooks/useMouse.js").MouseEvent) => {
      if (event.type === "scroll") {
        if (event.direction === "up") scrollUp(3);
        else scrollDown(3);
      } else if (event.type === "click") {
        selection.startSelection({ line: event.lineNum, col: event.textCol });
      } else if (event.type === "drag") {
        selection.updateEnd({ line: event.lineNum, col: event.textCol });
      }
    },
    [scrollUp, scrollDown, selection],
  );

  useMouse({ active, lines, scrollOffset, onEvent: handleMouseEvent });

  // ---- 键盘 ----
  useKeyboard({
    active,
    selecting: selection.selecting,
    viewportHeight,
    onScrollUp: scrollUp,
    onScrollDown: scrollDown,
    onQuit,
    onEnterSelection: selection.enterWithKeyboard,
    onCancelSelection: selection.cancel,
    onConfirmSelection: selection.confirm,
    onMoveLineBy: (delta) => {
      const newLine = selection.moveLineBy(delta);
      if (newLine !== null) {
        if (newLine <= scrollOffset) scrollUp();
        if (newLine > scrollOffset + viewportHeight) scrollDown();
      }
    },
    onMoveColBy: selection.moveColBy,
    onDeleteMode,
  });

  // ---- 状态上报 ----
  useEffect(() => {
    onStatusChange?.({
      scrollOffset,
      selecting: selection.selecting,
      selStartLine: selection.normStart?.line,
      selEndLine: selection.normEnd?.line,
    });
  }, [scrollOffset, selection.selecting, selection.normStart, selection.normEnd, onStatusChange]);

  // ---- 渲染 ----
  const visibleLines = lines.slice(scrollOffset, scrollOffset + viewportHeight);
  const { normStart, normEnd } = selection;

  return (
    <Box flexDirection="column" flexGrow={1} flexShrink={1} flexBasis={0} overflow="hidden">
      {visibleLines.map((line, idx) => {
        const lineNum = scrollOffset + idx + 1;
        const stripped = stripAnsi(line) || " ";

        // 计算高亮区间
        const selRange = normStart && normEnd
          ? getSelectionRangeForLine(lineNum, stripped.length, normStart, normEnd)
          : null;
        const annRanges = getAnnotationRangesForLine(annotations, lineNum, stripped.length);

        // 无高亮 → 直接渲染原始行（保留 ANSI 颜色）
        if (!selRange && annRanges.length === 0) {
          return (
            <Box key={lineNum} flexDirection="row">
              <Text>{line || " "}</Text>
            </Box>
          );
        }

        // 有高亮 → 分段渲染
        const segments = buildSegments(stripped, selRange, annRanges);

        return (
          <Box key={lineNum} flexDirection="row">
            {segments.map((seg, i) => {
              if (seg.selected) {
                return <Text key={i} backgroundColor="blue" color="white">{seg.text}</Text>;
              }
              if (seg.annotated) {
                return <Text key={i} backgroundColor="#3a3a2a" color="yellow">{seg.text}</Text>;
              }
              return <Text key={i}>{seg.text}</Text>;
            })}
          </Box>
        );
      })}
    </Box>
  );
}
