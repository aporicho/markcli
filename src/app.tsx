import React, { useState, useCallback, useMemo, useEffect } from "react";
import { Box, useStdout, useApp } from "ink";
import { MarkdownViewer, type ViewerStatus } from "./components/MarkdownViewer.js";
import { AnnotationInput } from "./components/AnnotationInput.js";
import { StatusBar } from "./components/StatusBar.js";
import { DeleteOverlay } from "./components/DeleteOverlay.js";
import { getSelectedText } from "./components/LineSelector.js";
import { useAnnotations } from "./hooks/useAnnotations.js";
import { renderMarkdownWrapped } from "./utils/markdown.js";
import {
  extractAnchor,
  relocateAnchor,
  lineColToOffset,
  offsetToLineCol,
} from "./utils/textAnchor.js";
import { stripAnsi } from "./utils/ranges.js";
import { disableMouseTracking, cleanupMouseOnExit } from "./utils/mouse.js";
import type { AppMode, Annotation } from "./types.js";

interface AppProps {
  filePath: string;
  content: string;
}

export function App({ filePath, content }: AppProps) {
  const { exit } = useApp();
  const { stdout } = useStdout();
  const viewportHeight = (stdout?.rows ?? 24) - 1; // -1 for StatusBar

  const termWidth = stdout?.columns ?? 80;
  const renderedLines = useMemo(
    () => renderMarkdownWrapped(content, termWidth),
    [content, termWidth],
  );

  // stripped 行 + 行长度数组（用于偏移量转换）
  const strippedLines = useMemo(
    () => renderedLines.map((l) => stripAnsi(l)),
    [renderedLines],
  );
  const lineLengths = useMemo(
    () => strippedLines.map((l) => l.length),
    [strippedLines],
  );
  const fullStrippedText = useMemo(
    () => strippedLines.join("\n"),
    [strippedLines],
  );

  const [mode, setMode] = useState<AppMode>("reading");
  const [viewerStatus, setViewerStatus] = useState<ViewerStatus>({
    scrollOffset: 0,
    selecting: false,
  });
  const [pendingSelection, setPendingSelection] = useState<{
    startLine: number;
    endLine: number;
    startCol: number;
    endCol: number;
  } | null>(null);

  const { annotations, addAnnotation, removeAnnotation } =
    useAnnotations(filePath);

  // ---- 将存储的批注解析到当前渲染位置 ----
  const resolvedAnnotations = useMemo(() => {
    return annotations.map((ann): Annotation => {
      // 有文本锚定 → 重新定位
      if (ann.quote) {
        const range = relocateAnchor(fullStrippedText, {
          quote: ann.quote,
          prefix: ann.prefix ?? "",
          suffix: ann.suffix ?? "",
        });
        if (range) {
          const start = offsetToLineCol(lineLengths, range.start);
          const end = offsetToLineCol(lineLengths, range.end);
          return {
            ...ann,
            startLine: start.line,
            endLine: end.line,
            startCol: start.col,
            endCol: end.col,
          };
        }
      }
      // 无锚定或定位失败 → 使用存储的 line/col（旧批注兼容）
      return ann;
    });
  }, [annotations, fullStrippedText, lineLengths]);

  // 进程退出兜底清理鼠标
  useEffect(() => {
    return cleanupMouseOnExit();
  }, []);

  // MarkdownViewer 选中完成 → 进入批注模式
  const handleSelect = useCallback(
    (startLine: number, endLine: number, startCol: number, endCol: number) => {
      setPendingSelection({ startLine, endLine, startCol, endCol });
      setMode("annotating");
    },
    [],
  );

  // 退出
  const handleQuit = useCallback(() => {
    disableMouseTracking();
    exit();
  }, [exit]);

  // 批注提交
  const handleAnnotationSubmit = useCallback(
    (comment: string) => {
      if (!pendingSelection) return;
      const { startLine, endLine, startCol, endCol } = pendingSelection;
      const selectedText = getSelectedText(strippedLines, startLine, endLine, startCol, endCol);

      // 提取文本锚定
      const offsetStart = lineColToOffset(lineLengths, startLine, startCol);
      const offsetEnd = lineColToOffset(lineLengths, endLine, endCol);
      const anchor = extractAnchor(fullStrippedText, offsetStart, offsetEnd);

      addAnnotation({
        startLine, endLine, startCol, endCol,
        selectedText, comment,
        quote: anchor.quote,
        prefix: anchor.prefix,
        suffix: anchor.suffix,
      });
      setPendingSelection(null);
      setMode("reading");
    },
    [pendingSelection, strippedLines, lineLengths, fullStrippedText, addAnnotation],
  );

  // 批注取消
  const handleAnnotationCancel = useCallback(() => {
    setPendingSelection(null);
    setMode("reading");
  }, []);

  // 删除模式
  const handleEnterDeleteMode = useCallback(() => {
    setMode("deleting");
  }, []);

  const handleDelete = useCallback((id: string) => {
    removeAnnotation(id);
    setMode("reading");
  }, [removeAnnotation]);

  const handleDeleteCancel = useCallback(() => {
    setMode("reading");
  }, []);

  const isViewing = mode === "reading" || mode === "selecting";

  // StatusBar 显示的模式：优先用 viewer 的 selecting 状态
  const displayMode: AppMode = mode === "annotating"
    ? "annotating"
    : mode === "deleting"
      ? "deleting"
      : viewerStatus.selecting
        ? "selecting"
        : "reading";

  // StatusBar 显示的选区行号
  const selStart = mode === "annotating"
    ? pendingSelection?.startLine ?? null
    : viewerStatus.selStartLine ?? null;
  const selEnd = mode === "annotating"
    ? pendingSelection?.endLine ?? null
    : viewerStatus.selEndLine ?? null;

  return (
    <Box flexDirection="column" height={stdout?.rows ?? 24}>
      <MarkdownViewer
        lines={renderedLines}
        viewportHeight={viewportHeight}
        annotations={resolvedAnnotations}
        active={isViewing}
        onSelect={handleSelect}
        onQuit={handleQuit}
        onStatusChange={setViewerStatus}
        onDeleteMode={handleEnterDeleteMode}
      />
      {mode === "deleting" && (
        <DeleteOverlay
          annotations={resolvedAnnotations}
          onDelete={handleDelete}
          onCancel={handleDeleteCancel}
        />
      )}
      {mode === "annotating" && pendingSelection && (
        <AnnotationInput
          selectedText={getSelectedText(
            strippedLines,
            pendingSelection.startLine,
            pendingSelection.endLine,
            pendingSelection.startCol,
            pendingSelection.endCol,
          )}
          onSubmit={handleAnnotationSubmit}
          onCancel={handleAnnotationCancel}
        />
      )}
      <StatusBar
        mode={displayMode}
        currentLine={viewerStatus.scrollOffset + 1}
        totalLines={renderedLines.length}
        selectionStart={selStart}
        selectionEnd={selEnd}
        annotationCount={resolvedAnnotations.length}
      />
    </Box>
  );
}
