import fs from "node:fs";
import { useCallback, useEffect, useRef } from "react";

/**
 * 监听文件变化，debounce 后调用 onChange。
 * 用于 Claude 编辑文件后 Mark 自动刷新。
 */
export function useFileWatcher(
	filePath: string,
	onChange: (newContent: string) => void,
): void {
	const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const onChangeRef = useRef(onChange);
	onChangeRef.current = onChange;

	const handleChange = useCallback(() => {
		if (timerRef.current) clearTimeout(timerRef.current);
		timerRef.current = setTimeout(() => {
			try {
				const content = fs.readFileSync(filePath, "utf-8");
				onChangeRef.current(content);
			} catch {
				// 文件可能被删除，忽略
			}
		}, 100);
	}, [filePath]);

	useEffect(() => {
		let watcher: fs.FSWatcher | null = null;
		try {
			watcher = fs.watch(filePath, handleChange);
		} catch {
			// 文件不存在或不支持 watch
		}

		// polling fallback：确保 WSL2/网络文件系统等环境也能检测变化
		fs.watchFile(filePath, { interval: 1000 }, handleChange);

		return () => {
			if (timerRef.current) clearTimeout(timerRef.current);
			watcher?.close();
			fs.unwatchFile(filePath, handleChange);
		};
	}, [filePath, handleChange]);
}
