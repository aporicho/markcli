#!/usr/bin/env node
import { execSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import * as esbuild from "esbuild";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");
const DIST = path.join(ROOT, "dist");
const BUNDLE = path.join(DIST, "bundle.mjs");
const BIN_DIR = path.join(DIST, "binaries");

const TARGETS = {
	"darwin-arm64": "bun-darwin-arm64",
	"darwin-x64": "bun-darwin-x64",
	"linux-x64": "bun-linux-x64",
	"linux-arm64": "bun-linux-arm64",
};

// --- 阶段 1: esbuild 打包 ---
async function bundle() {
	console.log("📦 esbuild: 打包中...");
	await esbuild.build({
		entryPoints: [path.join(ROOT, "src/cli.tsx")],
		bundle: true,
		platform: "node",
		format: "esm",
		outfile: BUNDLE,
		target: "node20",
		alias: {
			"react-devtools-core": path.join(
				ROOT,
				"scripts/shims/react-devtools-core.mjs",
			),
		},
		// Bun 运行时自带这些
		external: [],
		banner: {
			js: 'import { createRequire as __createRequire } from "module"; const require = __createRequire(import.meta.url); var parcelRequire;',
		},
	});
	console.log(`✅ bundle → ${path.relative(ROOT, BUNDLE)}`);
}

// --- 阶段 2: Bun 编译 ---
function compile(targetKey) {
	const bunTarget = TARGETS[targetKey];
	const outfile = path.join(BIN_DIR, `mark-${targetKey}`);

	console.log(`🔨 bun compile: ${targetKey}...`);
	execSync(
		`bun build --compile --target=${bunTarget} ${BUNDLE} --outfile ${outfile}`,
		{ stdio: "inherit", cwd: ROOT },
	);
	console.log(`✅ binary → ${path.relative(ROOT, outfile)}`);
}

// --- 主流程 ---
async function main() {
	const args = process.argv.slice(2);
	const buildAll = args.includes("--all");
	const targetArg = args.find((a) => a.startsWith("--target="));

	// 确保输出目录存在
	fs.mkdirSync(BIN_DIR, { recursive: true });

	// 阶段 1
	await bundle();

	// 阶段 2
	if (buildAll) {
		for (const key of Object.keys(TARGETS)) {
			compile(key);
		}
	} else if (targetArg) {
		const key = targetArg.split("=")[1];
		if (!TARGETS[key]) {
			console.error(
				`❌ 未知目标: ${key}\n可用: ${Object.keys(TARGETS).join(", ")}`,
			);
			process.exit(1);
		}
		compile(key);
	} else {
		// 默认：编译当前平台
		const os = process.platform === "darwin" ? "darwin" : "linux";
		const arch = process.arch === "arm64" ? "arm64" : "x64";
		const key = `${os}-${arch}`;
		compile(key);
	}

	console.log("\n🎉 构建完成！");
}

main().catch((err) => {
	console.error(err);
	process.exit(1);
});
