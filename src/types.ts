export interface Annotation {
  id: string;
  startLine: number;
  endLine: number;
  startCol?: number;
  endCol?: number;
  selectedText: string;
  comment: string;
  createdAt: string;
  // 文本锚定（W3C Web Annotation 标准）
  quote?: string;
  prefix?: string;
  suffix?: string;
}

export interface SelectionPos {
  line: number;
  col: number; // 0-indexed position in stripped text
}

export interface AnnotationFile {
  file: string;
  annotations: Annotation[];
}

export type AppMode = "reading" | "selecting" | "annotating" | "deleting";
