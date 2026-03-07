import { useEffect, useRef } from "react";
import { useStdin } from "ink";
import { stripAnsi, termColToCharIndex } from "../utils/ranges.js";
import { enableMouseTracking, disableMouseTracking } from "../utils/mouse.js";

export interface MouseClickEvent {
  type: "click";
  lineNum: number;
  textCol: number;
}

export interface MouseDragEvent {
  type: "drag";
  lineNum: number;
  textCol: number;
}

export interface MouseScrollEvent {
  type: "scroll";
  direction: "up" | "down";
}

export type MouseEvent = MouseClickEvent | MouseDragEvent | MouseScrollEvent;

interface UseMouseOptions {
  active: boolean;
  lines: string[];
  scrollOffset: number;
  onEvent: (event: MouseEvent) => void;
}

export function useMouse({ active, lines, scrollOffset, onEvent }: UseMouseOptions) {
  const stdinContext = useStdin();
  const emitter = (stdinContext as any).internal_eventEmitter;

  const stateRef = useRef({ scrollOffset, active, lines, onEvent });
  stateRef.current = { scrollOffset, active, lines, onEvent };

  useEffect(() => {
    if (!active) {
      disableMouseTracking();
      return;
    }

    enableMouseTracking();

    const handleInput = (data: Buffer | string) => {
      const str = typeof data === "string" ? data : data.toString();
      const s = stateRef.current;
      const sgrRegex = /\x1b\[<(\d+);(\d+);(\d+)([Mm])/g;
      let match;
      while ((match = sgrRegex.exec(str)) !== null) {
        const button = parseInt(match[1]!, 10);
        const col = parseInt(match[2]!, 10);
        const row = parseInt(match[3]!, 10);
        const isRelease = match[4] === "m";

        // 滚轮
        if (button === 64) { s.onEvent({ type: "scroll", direction: "up" }); continue; }
        if (button === 65) { s.onEvent({ type: "scroll", direction: "down" }); continue; }

        // 计算行号和文本列
        const lineNum = s.scrollOffset + row;
        if (lineNum < 1 || lineNum > s.lines.length) continue;
        const lineText = s.lines[lineNum - 1] ?? "";
        const stripped = stripAnsi(lineText);
        // col 是 1-based 终端列，减 1 转 0-based
        const termCol = Math.max(0, col - 1);
        const textCol = termColToCharIndex(stripped, termCol);

        // 左键点击
        if (button === 0 && !isRelease) {
          s.onEvent({ type: "click", lineNum, textCol });
          continue;
        }

        // 左键拖拽
        if (button === 32) {
          s.onEvent({ type: "drag", lineNum, textCol });
          continue;
        }
      }
    };

    emitter?.on("input", handleInput);

    return () => {
      disableMouseTracking();
      emitter?.removeListener("input", handleInput);
    };
  }, [active, emitter]);
}
