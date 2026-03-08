/** 统一鼠标生命周期管理 */

export function enableMouseTracking() {
	process.stdout.write("\x1b[?1000h\x1b[?1002h\x1b[?1006h");
	process.stdout.write("\x1b[5 q"); // 文本光标设为竖线
	process.stdout.write("\x1b]22;text\x1b\\"); // OSC 22: 鼠标指针设为 I-beam
}

export function disableMouseTracking() {
	process.stdout.write("\x1b[?1006l\x1b[?1002l\x1b[?1000l\x1b[?1003l");
	process.stdout.write("\x1b[0 q"); // 恢复默认文本光标
	process.stdout.write("\x1b]22;\x1b\\"); // OSC 22: 恢复默认鼠标指针
}

/** 注册 process exit 时自动关闭鼠标追踪，返回取消函数 */
export function cleanupMouseOnExit(): () => void {
	process.on("exit", disableMouseTracking);
	return () => {
		process.removeListener("exit", disableMouseTracking);
	};
}
