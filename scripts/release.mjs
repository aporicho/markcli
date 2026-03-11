#!/usr/bin/env node
import { execSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");

function run(cmd) {
	console.log(`$ ${cmd}`);
	execSync(cmd, { stdio: "inherit", cwd: ROOT });
}

function bump(version, type) {
	const [major, minor, patch] = version.split(".").map(Number);
	switch (type) {
		case "major":
			return `${major + 1}.0.0`;
		case "minor":
			return `${major}.${minor + 1}.0`;
		case "patch":
			return `${major}.${minor}.${patch + 1}`;
		default:
			return type; // 直接指定版本号
	}
}

const arg = process.argv[2];
if (!arg) {
	console.log(`用法: node scripts/release.mjs <patch|minor|major|x.y.z>`);
	console.log(`示例: node scripts/release.mjs patch`);
	process.exit(1);
}

// 从 git tag 获取当前版本（不再依赖 package.json）
let oldVersion;
try {
	oldVersion = execSync("git describe --tags --abbrev=0", {
		cwd: ROOT,
		encoding: "utf-8",
	})
		.trim()
		.replace(/^v/, "");
} catch {
	oldVersion = "0.0.0";
}

const newVersion = bump(oldVersion, arg);

console.log(`\n📦 发布 v${oldVersion} → v${newVersion}\n`);

// 打 tag + push（Go 版通过 ldflags 注入版本，无需更新文件）
run(`git tag v${newVersion}`);
run(`git push origin v${newVersion}`);

console.log(`\n✅ v${newVersion} 已发布！CI 正在构建二进制...`);
console.log(`   查看进度: gh run list --repo aporicho/markcli`);
