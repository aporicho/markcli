// IPC 消息类型定义（NDJSON over Unix socket）

// MCP → TUI 请求
export interface GetAnnotationsRequest {
	type: "get_annotations";
	file?: string;
}

export interface GetStatusRequest {
	type: "get_status";
}

export interface GetSelectionRequest {
	type: "get_selection";
}

export interface OpenFileRequest {
	type: "open_file";
	path: string;
}

export interface AddAnnotationRequest {
	type: "add_annotation";
	file?: string;
	selectedText: string;
	comment: string;
}

export interface UpdateAnnotationRequest {
	type: "update_annotation";
	id: string;
	comment: string;
}

export interface RemoveAnnotationRequest {
	type: "remove_annotation";
	id: string;
}

export type IpcRequest =
	| GetAnnotationsRequest
	| GetStatusRequest
	| GetSelectionRequest
	| OpenFileRequest
	| AddAnnotationRequest
	| UpdateAnnotationRequest
	| RemoveAnnotationRequest;

// TUI → MCP 响应
export interface AnnotationsResponse {
	type: "annotations";
	data: {
		file: string;
		annotations: Array<{
			id: string;
			selectedText: string;
			comment: string;
			startLine: number;
			endLine: number;
		}>;
	};
}

export interface StatusResponse {
	type: "status";
	data: {
		file: string;
		mode: string;
		annotationCount: number;
	};
}

export interface SelectionResponse {
	type: "selection";
	data: {
		file: string;
		selectedText: string | null;
	};
}

export interface OkResponse {
	type: "ok";
	message: string;
}

export interface ErrorResponse {
	type: "error";
	message: string;
}

export type IpcResponse =
	| AnnotationsResponse
	| StatusResponse
	| SelectionResponse
	| OkResponse
	| ErrorResponse;
