import React, { useEffect } from "react";
import { Box, Text, useInput } from "ink";
import { disableMouseTracking } from "../utils/mouse.js";
import type { Annotation } from "../types.js";

interface DeleteOverlayProps {
  annotations: Annotation[];
  onDelete: (id: string) => void;
  onCancel: () => void;
}

export function DeleteOverlay({ annotations, onDelete, onCancel }: DeleteOverlayProps) {
  useEffect(() => {
    disableMouseTracking();
  }, []);

  useInput((input, key) => {
    if (key.escape) {
      onCancel();
      return;
    }
    const num = parseInt(input, 10);
    if (num >= 1 && num <= Math.min(9, annotations.length)) {
      const ann = annotations[num - 1]!;
      onDelete(ann.id);
    }
  });

  if (annotations.length === 0) {
    return (
      <Box flexDirection="column" borderStyle="single" borderColor="red" paddingX={1}>
        <Text color="red" bold>删除批注</Text>
        <Text dimColor>（暂无批注）按 Esc 返回</Text>
      </Box>
    );
  }

  const display = annotations.slice(0, 9);

  return (
    <Box flexDirection="column" borderStyle="single" borderColor="red" paddingX={1}>
      <Text color="red" bold>删除批注（按数字键删除，Esc 取消）</Text>
      {display.map((ann, i) => {
        const range = ann.startLine === ann.endLine
          ? `L${ann.startLine}`
          : `L${ann.startLine}-${ann.endLine}`;
        const textPreview = ann.selectedText.length > 30
          ? ann.selectedText.slice(0, 27) + "..."
          : ann.selectedText;
        const commentPreview = ann.comment.length > 30
          ? ann.comment.slice(0, 27) + "..."
          : ann.comment;
        return (
          <Text key={ann.id}>
            <Text color="yellow">[{i + 1}]</Text>
            <Text> {range} </Text>
            <Text dimColor>"{textPreview.replace(/\n/g, "↵")}"</Text>
            <Text> → {commentPreview}</Text>
          </Text>
        );
      })}
    </Box>
  );
}
