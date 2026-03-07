import fs from "node:fs";
import path from "node:path";
import type { AnnotationFile } from "../types.js";

function getAnnotationPath(filePath: string): string {
  const dir = path.dirname(filePath);
  const base = path.basename(filePath);
  return path.join(dir, `${base}.markcli.json`);
}

export function loadAnnotations(filePath: string): AnnotationFile {
  const annoPath = getAnnotationPath(filePath);
  if (fs.existsSync(annoPath)) {
    const raw = fs.readFileSync(annoPath, "utf-8");
    return JSON.parse(raw) as AnnotationFile;
  }
  return {
    file: path.basename(filePath),
    annotations: [],
  };
}

export function saveAnnotations(
  filePath: string,
  data: AnnotationFile
): void {
  const annoPath = getAnnotationPath(filePath);
  fs.writeFileSync(annoPath, JSON.stringify(data, null, 2), "utf-8");
}

export function clearAnnotations(filePath: string): void {
  const annoPath = getAnnotationPath(filePath);
  if (fs.existsSync(annoPath)) {
    fs.unlinkSync(annoPath);
  }
}
