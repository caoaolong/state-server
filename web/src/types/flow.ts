/**
 * 流程编辑相关共享类型与常量
 */

/** 节点编辑表单模型（用于编辑弹窗与 FlowCanvas） */
export interface EditFormModel {
	label: string;
	description: string;
	requestPath: string;
	requestMethod: string;
	requestData: string;
	outputCount: number;
	options: Array<{ label: string; description: string }>;
	inputCount: number;
	results: Array<{ label: string; description: string }>;
	nodeCategory: string;
	nodeKind?: string;
}

/** 添加节点菜单选项 key（工具栏/右键菜单统一） */
export type AddNodeOptionKey = "scene-start" | "scene-end" | "scene-default" | "choice" | "result";

export const ADD_NODE_KEYS: AddNodeOptionKey[] = [
	"scene-start",
	"scene-end",
	"scene-default",
	"choice",
	"result",
];

export function isAddNodeOptionKey(key: string): key is AddNodeOptionKey {
	return ADD_NODE_KEYS.includes(key as AddNodeOptionKey);
}
