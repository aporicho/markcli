import React, { useState } from "react";
import { Text, useInput } from "ink";

function stripMouseSequences(str: string): string {
  return str.replace(/\x1b\[<[\d;]*[Mm]/g, "");
}

interface TextInputProps {
  value: string;
  onChange: (value: string) => void;
  onSubmit?: (value: string) => void;
  placeholder?: string;
  focus?: boolean;
  showCursor?: boolean;
}

export function TextInput({
  value,
  onChange,
  onSubmit,
  placeholder = "",
  focus = true,
  showCursor = true,
}: TextInputProps) {
  const [cursor, setCursor] = useState(value.length);

  useInput(
    (input, key) => {
      const cleanInput = stripMouseSequences(input);

      if (key.return) {
        onSubmit?.(value);
        return;
      }

      if (key.backspace || key.delete) {
        if (cursor > 0) {
          const next = value.slice(0, cursor - 1) + value.slice(cursor);
          setCursor((c) => c - 1);
          onChange(next);
        }
        return;
      }

      if (key.leftArrow) {
        setCursor((c) => Math.max(0, c - 1));
        return;
      }

      if (key.rightArrow) {
        setCursor((c) => Math.min(value.length, c + 1));
        return;
      }

      if (
        key.upArrow ||
        key.downArrow ||
        key.tab ||
        key.escape ||
        (key.ctrl && cleanInput === "c")
      ) {
        return;
      }

      if (cleanInput) {
        const next =
          value.slice(0, cursor) + cleanInput + value.slice(cursor);
        setCursor((c) => c + cleanInput.length);
        onChange(next);
      }
    },
    { isActive: focus }
  );

  // 空值显示 placeholder
  if (!value && placeholder) {
    if (showCursor && focus) {
      return (
        <Text>
          <Text inverse>{placeholder[0]}</Text>
          <Text dimColor>{placeholder.slice(1)}</Text>
        </Text>
      );
    }
    return <Text dimColor>{placeholder}</Text>;
  }

  // 不显示光标
  if (!showCursor || !focus) {
    return <Text>{value}</Text>;
  }

  // 带光标渲染
  const before = value.slice(0, cursor);
  const cursorChar = value[cursor] ?? " ";
  const after = value.slice(cursor + 1);

  return (
    <Text>
      {before}
      <Text inverse>{cursorChar}</Text>
      {after}
    </Text>
  );
}
