import { Box, Text, useInput } from "ink";
import { useEffect, useState } from "react";

function stripMouseSequences(str: string): string {
	// Ink 的 useInput 会剥掉 \x1b 前缀，所以实际收到的是 [<0;10;5M 形式
	return str.replace(/\x1b?\[<[\d;]*[Mm]/g, "");
}

interface TextInputProps {
	value: string;
	onChange: (value: string) => void;
	onSubmit?: (value: string) => void;
	placeholder?: string;
	focus?: boolean;
	showCursor?: boolean;
	onLineCountChange?: (count: number) => void;
}

export function TextInput({
	value,
	onChange,
	onSubmit,
	placeholder = "",
	focus = true,
	showCursor = true,
	onLineCountChange,
}: TextInputProps) {
	const [cursor, setCursor] = useState(value.length);

	const lineCount = value.split("\n").length;
	useEffect(() => {
		onLineCountChange?.(lineCount);
	}, [lineCount, onLineCountChange]);

	useInput(
		(input, key) => {
			const cleanInput = stripMouseSequences(input);

			if (key.return) {
				if (key.shift) {
					// Shift+Enter → 插入换行（Kitty 协议终端）
					const next = `${value.slice(0, cursor)}\n${value.slice(cursor)}`;
					setCursor((c) => c + 1);
					onChange(next);
				} else {
					// Enter → 提交
					onSubmit?.(value);
				}
				return;
			}

			// Ctrl+J → 插入换行（所有终端通用 fallback）
			if (key.ctrl && cleanInput === "j") {
				const next = `${value.slice(0, cursor)}\n${value.slice(cursor)}`;
				setCursor((c) => c + 1);
				onChange(next);
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

			if (key.upArrow) {
				// 多行：移动到上一行同列位置
				const textBefore = value.slice(0, cursor);
				const linesBefore = textBefore.split("\n");
				if (linesBefore.length > 1) {
					const currentCol = linesBefore[linesBefore.length - 1]!.length;
					const prevLineLength = linesBefore[linesBefore.length - 2]!.length;
					const targetCol = Math.min(currentCol, prevLineLength);
					setCursor(cursor - currentCol - 1 - (prevLineLength - targetCol));
				}
				return;
			}

			if (key.downArrow) {
				// 多行：移动到下一行同列位置
				const textBefore = value.slice(0, cursor);
				const linesBefore = textBefore.split("\n");
				const currentCol = linesBefore[linesBefore.length - 1]!.length;
				const afterCursor = value.slice(cursor);
				const nextNewline = afterCursor.indexOf("\n");
				if (nextNewline !== -1) {
					const afterNextNewline = afterCursor.slice(nextNewline + 1);
					const followingNewline = afterNextNewline.indexOf("\n");
					const nextLineLength =
						followingNewline === -1
							? afterNextNewline.length
							: followingNewline;
					const targetCol = Math.min(currentCol, nextLineLength);
					setCursor(cursor + nextNewline + 1 + targetCol);
				}
				return;
			}

			if (
				key.tab ||
				key.escape ||
				(key.ctrl &&
					(cleanInput === "c" || cleanInput === "d" || cleanInput === "r"))
			) {
				return;
			}

			if (cleanInput) {
				const next = value.slice(0, cursor) + cleanInput + value.slice(cursor);
				setCursor((c) => c + cleanInput.length);
				onChange(next);
			}
		},
		{ isActive: focus },
	);

	// 计算光标在多行中的位置
	const textBeforeCursor = value.slice(0, cursor);
	const linesBeforeCursor = textBeforeCursor.split("\n");
	const cursorLine = linesBeforeCursor.length - 1;
	const cursorCol = linesBeforeCursor[cursorLine]!.length;
	const allLines = value.split("\n");

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
		if (allLines.length === 1) {
			return <Text>{value}</Text>;
		}
		return (
			<Box flexDirection="column">
				{allLines.map((line, i) => (
					<Text key={i}>{line || " "}</Text>
				))}
			</Box>
		);
	}

	// 单行带光标渲染
	if (allLines.length === 1) {
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

	// 多行带光标渲染
	return (
		<Box flexDirection="column">
			{allLines.map((line, i) => {
				if (i === cursorLine) {
					const before = line.slice(0, cursorCol);
					const cursorChar = line[cursorCol] ?? " ";
					const after = line.slice(cursorCol + 1);
					return (
						<Text key={i}>
							{before}
							<Text inverse>{cursorChar}</Text>
							{after}
						</Text>
					);
				}
				return <Text key={i}>{line || " "}</Text>;
			})}
		</Box>
	);
}
