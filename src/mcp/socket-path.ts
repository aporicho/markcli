import os from "node:os";
import path from "node:path";

export function getSocketPath(): string {
	const uid = os.userInfo().uid;
	return path.join(os.tmpdir(), `mark-${uid}.sock`);
}
