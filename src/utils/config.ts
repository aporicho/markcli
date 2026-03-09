import fs from "node:fs";
import os from "node:os";
import path from "node:path";

export function loadConfig(): { theme?: string } {
	const configPath = path.join(
		os.homedir(),
		".config",
		"markcli",
		"config.json",
	);
	try {
		const raw = fs.readFileSync(configPath, "utf-8");
		const parsed = JSON.parse(raw);
		if (parsed && typeof parsed === "object") {
			return {
				theme: typeof parsed.theme === "string" ? parsed.theme : undefined,
			};
		}
	} catch {
		// 文件不存在或解析失败
	}
	return {};
}
