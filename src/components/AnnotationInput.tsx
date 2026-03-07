import React, { useState, useEffect } from "react";
import { Box, Text, useInput } from "ink";
import { TextInput } from "./TextInput.js";
import { disableMouseTracking } from "../utils/mouse.js";

interface AnnotationInputProps {
  selectedText: string;
  onSubmit: (comment: string) => void;
  onCancel: () => void;
}

export function AnnotationInput({
  selectedText,
  onSubmit,
  onCancel,
}: AnnotationInputProps) {
  const [value, setValue] = useState("");

  useEffect(() => {
    disableMouseTracking();
  }, []);

  // 截取预览，最多显示 40 字符
  const preview = selectedText.length > 40
    ? selectedText.slice(0, 37) + "..."
    : selectedText;
  // 替换换行为可见符号
  const previewLabel = preview.replace(/\n/g, "↵");

  useInput((_input, key) => {
    if (key.escape) {
      onCancel();
    }
  });

  return (
    <Box
      flexDirection="column"
      borderStyle="single"
      borderColor="cyan"
      paddingX={1}
    >
      <Text color="gray">"{previewLabel}"</Text>
      <Box>
        <TextInput
          value={value}
          onChange={setValue}
          onSubmit={(val) => {
            if (val.trim()) {
              onSubmit(val.trim());
            }
          }}
          placeholder="输入批注，Enter 确认，Esc 取消"
        />
      </Box>
    </Box>
  );
}
