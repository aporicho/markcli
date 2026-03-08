import { nanoid } from "nanoid";
import { useCallback, useEffect, useState } from "react";
import type { Annotation, AnnotationFile } from "../types.js";
import { loadAnnotations, saveAnnotations } from "../utils/storage.js";

export function useAnnotations(filePath: string) {
	const [data, setData] = useState<AnnotationFile>(() =>
		loadAnnotations(filePath),
	);

	// filePath 变化时重新加载批注
	useEffect(() => {
		setData(loadAnnotations(filePath));
	}, [filePath]);

	const addAnnotation = useCallback(
		(params: {
			startLine: number;
			endLine: number;
			startCol?: number;
			endCol?: number;
			selectedText: string;
			comment: string;
			quote?: string;
			prefix?: string;
			suffix?: string;
		}) => {
			const annotation: Annotation = {
				id: nanoid(6),
				startLine: params.startLine,
				endLine: params.endLine,
				...(params.startCol !== undefined ? { startCol: params.startCol } : {}),
				...(params.endCol !== undefined ? { endCol: params.endCol } : {}),
				selectedText: params.selectedText,
				comment: params.comment,
				createdAt: new Date().toISOString(),
				...(params.quote
					? {
							quote: params.quote,
							prefix: params.prefix ?? "",
							suffix: params.suffix ?? "",
						}
					: {}),
			};
			setData((prev) => {
				const next = {
					...prev,
					annotations: [...prev.annotations, annotation],
				};
				saveAnnotations(filePath, next);
				return next;
			});
		},
		[filePath],
	);

	const updateAnnotation = useCallback(
		(id: string, comment: string) => {
			setData((prev) => {
				const next = {
					...prev,
					annotations: prev.annotations.map((a) =>
						a.id === id ? { ...a, comment } : a,
					),
				};
				saveAnnotations(filePath, next);
				return next;
			});
		},
		[filePath],
	);

	const removeAnnotation = useCallback(
		(id: string) => {
			setData((prev) => {
				const next = {
					...prev,
					annotations: prev.annotations.filter((a) => a.id !== id),
				};
				saveAnnotations(filePath, next);
				return next;
			});
		},
		[filePath],
	);

	return {
		annotations: data.annotations,
		addAnnotation,
		updateAnnotation,
		removeAnnotation,
	};
}
