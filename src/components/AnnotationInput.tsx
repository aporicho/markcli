import { Box, useInput, useStdin } from "ink";
import { useEffect, useRef, useState } from "react";
import { disableMouseTracking, enableMouseTracking } from "../utils/mouse.js";
import { TextInput } from "./TextInput.js";

interface AnnotationInputProps {
	onSubmit: (comment: string) => void;
	onCancel: () => void;
	onReselect?: (currentComment: string) => void;
	onDelete?: () => void;
	isEditing?: boolean;
	top?: number;
	left?: number;
	width?: number;
	initialValue?: string;
}

export function AnnotationInput({
	onSubmit,
	onCancel,
	onReselect,
	onDelete,
	isEditing,
	top,
	left,
	width,
	initialValue,
}: AnnotationInputProps) {
	const [value, setValue] = useState(initialValue ?? "");
	const [ready, setReady] = useState(false);

	const stdinContext = useStdin();
	const emitter = (
		stdinContext as unknown as {
			internal_eventEmitter: import("node:events").EventEmitter;
		}
	).internal_eventEmitter;
	const onCancelRef = useRef(onCancel);
	onCancelRef.current = onCancel;

	const readyRef = useRef(false);

	useEffect(() => {
		const timer = setTimeout(() => {
			setReady(true);
			readyRef.current = true;
		}, 50);
		return () => clearTimeout(timer);
	}, []);

	// 立即开始消费 SGR 鼠标数据（防止泄漏到 TextInput），但点击外部取消需等 ready
	useEffect(() => {
		enableMouseTracking();

		const boxTop = (top ?? 0) + 1; // +1: marginTop 是 0-based，终端行是 1-based
		const boxHeight = 3;
		const boxLeft = (left ?? 0) + 1;
		const boxWidth = width ?? 40;

		const handleInput = (data: Buffer | string) => {
			const str = typeof data === "string" ? data : data.toString();
			const sgrRegex = /\x1b\[<(\d+);(\d+);(\d+)([Mm])/g;
			let match: RegExpExecArray | null;
			while ((match = sgrRegex.exec(str)) !== null) {
				const button = Number.parseInt(match[1]!, 10);
				const col = Number.parseInt(match[2]!, 10);
				const row = Number.parseInt(match[3]!, 10);
				const isRelease = match[4] === "m";

				// 只响应左键点击，忽略拖拽/滚轮/release
				if (button !== 0 || isRelease) continue;

				// ready 之前只消费数据不触发动作，防止双击残留事件误触取消
				if (!readyRef.current) continue;

				const inBox =
					row >= boxTop &&
					row < boxTop + boxHeight &&
					col >= boxLeft &&
					col < boxLeft + boxWidth;

				if (!inBox) {
					onCancelRef.current();
					return;
				}
			}
		};

		emitter?.on("input", handleInput);

		return () => {
			disableMouseTracking();
			emitter?.removeListener("input", handleInput);
		};
	}, [top, left, width, emitter]);

	useInput((input, key) => {
		if (key.escape) {
			onCancel();
			return;
		}
		if (key.ctrl && input === "r" && onReselect) {
			onReselect(value);
			return;
		}
		if (key.ctrl && input === "d" && onDelete) {
			onDelete();
			return;
		}
	});

	return (
		<Box
			borderStyle="round"
			borderColor="cyan"
			backgroundColor="black"
			paddingX={1}
			position="absolute"
			marginTop={top}
			marginLeft={left}
			width={width}
			height={3}
			flexDirection="column"
		>
			<TextInput
				value={value}
				onChange={setValue}
				onSubmit={(val) => {
					if (isEditing) {
						onSubmit(val.trim());
					} else if (val.trim()) {
						onSubmit(val.trim());
					}
				}}
				placeholder=" "
				focus={ready}
			/>
		</Box>
	);
}
