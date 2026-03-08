import { useInput } from "ink";

interface UseKeyboardOptions {
	active: boolean;
	selecting: boolean;
	viewportHeight: number;
	onScrollUp: (n?: number) => void;
	onScrollDown: (n?: number) => void;
	onQuit: () => void;
	// 选择模式操作
	onEnterSelection: () => void;
	onCancelSelection: () => void;
	onConfirmSelection: () => void;
	onMoveLineBy: (delta: number) => void;
	onMoveColBy: (delta: number) => void;
	onOverviewMode?: () => void;
}

export function useKeyboard({
	active,
	selecting,
	viewportHeight,
	onScrollUp,
	onScrollDown,
	onQuit,
	onEnterSelection,
	onCancelSelection,
	onConfirmSelection,
	onMoveLineBy,
	onMoveColBy,
	onOverviewMode,
}: UseKeyboardOptions) {
	useInput((input, key) => {
		if (!active) return;

		// ---- 上下方向键 ----
		if (key.upArrow) {
			if (selecting) {
				onMoveLineBy(-1);
				// 如果需要滚动会在外部处理
			} else {
				onScrollUp();
			}
			return;
		}
		if (key.downArrow) {
			if (selecting) {
				onMoveLineBy(1);
			} else {
				onScrollDown();
			}
			return;
		}

		// ---- 翻页 ----
		if (key.pageUp) {
			onScrollUp(viewportHeight - 2);
			return;
		}
		if (key.pageDown) {
			onScrollDown(viewportHeight - 2);
			return;
		}

		if (!selecting) {
			// 阅读模式
			if (input === "q") {
				onQuit();
				return;
			}
			if (input === "d" && onOverviewMode) {
				onOverviewMode();
				return;
			}
			if (input === "v") {
				onEnterSelection();
				return;
			}
		} else {
			// 选中模式
			if (key.escape) {
				onCancelSelection();
				return;
			}
			if (key.leftArrow || input === "h") {
				onMoveColBy(-1);
				return;
			}
			if (key.rightArrow || input === "l") {
				onMoveColBy(1);
				return;
			}
			if (input === "a" || key.return) {
				onConfirmSelection();
				return;
			}
		}
	});
}
