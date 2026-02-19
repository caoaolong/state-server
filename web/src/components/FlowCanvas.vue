<script setup lang="ts">
/**
 * 流程画布：内部持有全部状态与逻辑（加载/保存、节点 CRUD、测试运行、弹窗等），
 * 仅接收 stateMachineId 与 dark 两个入参。左侧画布 + 右侧日志面板。
 */
import { ref, onMounted, onUnmounted, computed, nextTick } from "vue";
import { useRouter } from "vue-router";
import { type Node, type Edge, type Connection, type NodeChange, type EdgeChange, useVueFlow } from "@vue-flow/core";
import { VueFlow } from "@vue-flow/core";
import { Background } from "@vue-flow/background";
import {
	NButton,
	NIcon,
	NDropdown,
	NMenu,
	NModal,
	NCard,
	NForm,
	NFormItem,
	NInput,
	NInputNumber,
	NDivider,
	useMessage,
	NSplit,
} from "naive-ui";
import type { DropdownOption, MenuOption } from "naive-ui";
import { Play, Stop, Add, GridOutline, SaveOutline } from "@vicons/ionicons5";
import { getFlowData, getStateMachineById, saveFlow, createStateMachine, updateStateMachine, updateNode, createNode, http } from "../api";
import { useLayout } from "../core/layout";
import { type EditFormModel, isAddNodeOptionKey } from "../types/flow";
import FlowNode from "./FlowNode.vue";
import FlowLogPanel from "./FlowLogPanel.vue";
import type { FlowNodeData } from "./FlowNode.vue";

const props = withDefaults(
	defineProps<{
		/** 当前编辑的状态机 ID，无则为新建 */
		stateMachineId?: string;
		dark?: boolean;
	}>(),
	{ dark: false }
);

const router = useRouter();
const message = useMessage();
const { layout } = useLayout();
const { fitView, findNode, project } = useVueFlow();

// ---------- 状态机元数据 ----------
const stateMachineName = ref("");
const stateMachineDescription = ref("");
const stateMachineBaseUrl = ref("");

function getEffectiveBaseUrl(base: string): string {
	if (!base) return base;
	if (import.meta.env.DEV && (base.startsWith("http://localhost:5000") || base.startsWith("http://127.0.0.1:5000"))) {
		return `${window.location.origin}/dev-5000`;
	}
	return base;
}
const effectiveBaseUrl = computed(() => getEffectiveBaseUrl(stateMachineBaseUrl.value));

// ---------- 画布数据与布局 ----------
const nodes = ref<Node[]>([]);
const edges = ref<Edge[]>([]);

// ---------- 选中节点（供右侧日志面板） ----------
const currentSelectionId = ref<string | null>(null);
const currentSelectionNode = computed(() => {
	if (!currentSelectionId.value) return null;
	return nodes.value.find((n) => n.id === currentSelectionId.value);
});

function onNodeClick(event: { node: Node }) {
	currentSelectionId.value = event.node.id;
}

function onNodeRunResult(payload: { nodeId: string; runResult: string; nodeState: string }) {
	const { nodeId, runResult, nodeState } = payload;
	nodes.value = nodes.value.map((n) =>
		n.id === nodeId ? { ...n, data: { ...n.data, runResult, nodeState } } : n
	);
}

// ---------- 节点请求 ----------
const manualRunningNodeId = ref<string | null>(null);
async function runNodeRequest(nodeId: string) {
	const node = nodes.value.find((n) => n.id === nodeId);
	if (!node) return;
	const data = node.data as FlowNodeData;
	const baseUrl = (effectiveBaseUrl.value || "").replace(/\/$/, "");
	const path = (data.requestPath || "").trim();
	if (!path) {
		message.warning("请先配置该节点的请求路径");
		return;
	}
	const url = path.startsWith("http") ? path : `${baseUrl || ""}${path.startsWith("/") ? path : `/${path}`}`;
	const method = (data.requestMethod || "GET").toUpperCase() as "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
	let requestBody: unknown = undefined;
	const rawBody = (data.requestData || "").trim();
	if (rawBody && ["POST", "PUT", "PATCH"].includes(method)) {
		try {
			requestBody = JSON.parse(rawBody);
		} catch {
			requestBody = rawBody;
		}
	}
	const requestPayloadStr = rawBody || (requestBody != null ? JSON.stringify(requestBody, null, 2) : "");
	currentSelectionId.value = nodeId;
	manualRunningNodeId.value = nodeId;
	try {
		const res = await http({ method, url, body: requestBody });
		const resultStr = typeof res === "string" ? res : JSON.stringify(res, null, 2);
		nodes.value = nodes.value.map((n) =>
			n.id === nodeId
				? { ...n, data: { ...n.data, requestData: requestPayloadStr, runResult: resultStr } }
				: n
		);
		message.success("请求成功");
	} catch (e) {
		const errMsg = e instanceof Error ? e.message : String(e);
		nodes.value = nodes.value.map((n) =>
			n.id === nodeId
				? { ...n, data: { ...n.data, requestData: requestPayloadStr, runResult: JSON.stringify({ error: errMsg }, null, 2) } }
				: n
		);
		message.error(errMsg);
	} finally {
		manualRunningNodeId.value = null;
	}
}

// ---------- 连线与画布变更 ----------
function layoutGraph(direction: string) {
	nodes.value = layout(nodes.value, edges.value, direction);
	nextTick(() => fitView());
}

function onConnect(conn: Connection) {
	if (!conn.source || !conn.target) return;
	const id = `edge-${conn.source}-${conn.sourceHandle ?? "out"}-${conn.target}-${conn.targetHandle ?? "in"}`;
	const newEdge: Edge = {
		id,
		source: conn.source,
		target: conn.target,
		...(conn.sourceHandle != null && { sourceHandle: conn.sourceHandle }),
		...(conn.targetHandle != null && { targetHandle: conn.targetHandle }),
	};
	if (edges.value.some((e) => e.id === newEdge.id)) return;
	edges.value = [...edges.value, newEdge];
}

function onNodesChange(changes: NodeChange[]) {
	const positionChanges = changes.filter(
		(c): c is NodeChange & { type: "position"; id: string; position: { x: number; y: number } } =>
			c.type === "position" && "position" in c && c.position != null
	);
	if (positionChanges.length === 0) return;
	nodes.value = nodes.value.map((n) => {
		const ch = positionChanges.find((c) => c.id === n.id);
		return ch ? { ...n, position: { ...ch.position } } : n;
	});
}

function onEdgesChange(changes: EdgeChange[]) {
	const removeIds = new Set(
		changes.filter((c): c is EdgeChange & { type: "remove" } => c.type === "remove").map((c) => c.id)
	);
	if (removeIds.size === 0) return;
	edges.value = edges.value.filter((e) => !removeIds.has(e.id));
}

// ---------- 节点 CRUD ----------
const idCounters = ref({ sceneStart: 0, sceneEnd: 0, sceneDefault: 0, choice: 0, result: 0, task: 0 });
const DEFAULT_ADD_POSITION = { x: 200, y: 150 };

function nextId(kind: keyof typeof idCounters.value): string {
	idCounters.value[kind]++;
	const n = idCounters.value[kind];
	const prefix =
		kind === "sceneStart" ? "scene-start"
		: kind === "sceneEnd" ? "scene-end"
		: kind === "sceneDefault" ? "scene"
		: kind === "choice" ? "choice"
		: kind === "result" ? "result"
		: "task";
	return `${prefix}-${n}`;
}

function hasNodeKind(kind: string): boolean {
	return nodes.value.some(
		(n) => (n.data as FlowNodeData)?.nodeCategory === "scene" && (n.data as FlowNodeData)?.nodeKind === kind
	);
}

function getNodeClass(data: FlowNodeData | undefined): string {
	if (!data) return "";
	const cat = data.nodeCategory;
	const kind = data.nodeKind;
	if (cat === "scene") return `flow-type-scene flow-type-scene-${kind ?? "default"}`;
	if (cat === "choice") return "flow-type-choice";
	if (cat === "result") return "flow-type-result";
	if (cat === "task") return "flow-type-task";
	return "";
}

function addNodeAt(
	position: { x: number; y: number },
	option: import("../types/flow").AddNodeOptionKey
) {
	if (option === "scene-start" && hasNodeKind("start")) return;
	if (option === "scene-end" && hasNodeKind("end")) return;
	const id =
		option === "scene-start" ? nextId("sceneStart")
		: option === "scene-end" ? nextId("sceneEnd")
		: option === "scene-default" ? nextId("sceneDefault")
		: option === "choice" ? nextId("choice")
		: option === "result" ? nextId("result")
		: nextId("task");
	const labels: Record<string, string> = {
		"scene-start": "开始",
		"scene-end": "结束",
		"scene-default": "普通场景",
		choice: "选择",
		result: "结果",
		task: "任务",
	};
	let data: FlowNodeData;
	let nodeClass: string;
	if (option.startsWith("scene-")) {
		const nodeKind = option === "scene-start" ? "start" : option === "scene-end" ? "end" : "default";
		data = { label: labels[option], nodeCategory: "scene", nodeKind };
		nodeClass = `flow-type-scene flow-type-scene-${nodeKind}`;
	} else if (option === "choice") {
		data = {
			label: labels.choice,
			nodeCategory: "choice",
			outputCount: 2,
			options: [
				{ label: "选项 1", description: "" },
				{ label: "选项 2", description: "" },
			],
		};
		nodeClass = "flow-type-choice";
	} else if (option === "result") {
		data = {
			label: labels.result,
			nodeCategory: "result",
			inputCount: 2,
			results: [
				{ label: "结果 1", description: "" },
				{ label: "结果 2", description: "" },
			],
		};
		nodeClass = "flow-type-result";
	} else {
		data = { label: labels.task, nodeCategory: "task", description: "" };
		nodeClass = "flow-type-task";
	}
	const newNode: Node = { id, type: "flow", position, data, class: nodeClass };
	nodes.value = [...nodes.value, newNode];

	// 若有状态机 id，调用创建节点接口持久化
	if (props.stateMachineId) {
		createNode(props.stateMachineId, {
			id: newNode.id,
			type: newNode.type ?? "flow",
			position: newNode.position ?? { x: 0, y: 0 },
			data: newNode.data as Record<string, unknown>,
		}).catch((e) => message.error(e instanceof Error ? e.message : "保存节点失败"));
	}
}

// ---------- 画布/边右键菜单 ----------
const contextMenuShow = ref(false);
const contextMenuPosition = ref({ x: 0, y: 0 });
let lastContextEvent: MouseEvent | null = null;

function onPaneContextMenu(event: MouseEvent) {
	event.preventDefault();
	edgeContextMenuShow.value = false;
	lastContextEvent = event;
	contextMenuPosition.value = { x: event.clientX, y: event.clientY };
	contextMenuShow.value = true;
}

function handleContextMenuSelect(key: string) {
	if (key !== "__close__" && lastContextEvent && isAddNodeOptionKey(key)) {
		const flowPos = project({ x: lastContextEvent.clientX, y: lastContextEvent.clientY });
		addNodeAt(flowPos, key);
	}
	contextMenuShow.value = false;
	lastContextEvent = null;
}

const edgeContextMenuShow = ref(false);
const edgeContextMenuPosition = ref({ x: 0, y: 0 });
const edgeContextMenuEdgeId = ref<string | null>(null);

function onEdgeContextMenu(ev: { event: MouseEvent | TouchEvent; edge: { id: string } }) {
	ev.event.preventDefault();
	contextMenuShow.value = false;
	edgeContextMenuEdgeId.value = ev.edge.id;
	const e = ev.event as MouseEvent;
	edgeContextMenuPosition.value = { x: e.clientX, y: e.clientY };
	edgeContextMenuShow.value = true;
}

const edgeContextMenuOptions: MenuOption[] = [{ label: "删除连线", key: "delete-edge" }];

function closeEdgeContextMenu() {
	edgeContextMenuShow.value = false;
	edgeContextMenuEdgeId.value = null;
}

function handleEdgeContextMenuSelect(key: string) {
	if (key === "delete-edge" && edgeContextMenuEdgeId.value) {
		edges.value = edges.value.filter((e) => e.id !== edgeContextMenuEdgeId.value);
	}
	closeEdgeContextMenu();
}

function handleToolbarAdd(key: string) {
	if (isAddNodeOptionKey(key)) addNodeAt({ ...DEFAULT_ADD_POSITION }, key);
}

function onNodeCopy(nodeId: string) {
	const node = nodes.value.find((n) => n.id === nodeId);
	if (!node) return;
	const data = node.data as FlowNodeData;
	if (data?.nodeCategory === "scene" && (data?.nodeKind === "start" || data?.nodeKind === "end")) return;
	const category = data?.nodeCategory ?? "scene";
	const kind =
		category === "scene"
			? data?.nodeKind === "start" ? "sceneStart" : data?.nodeKind === "end" ? "sceneEnd" : "sceneDefault"
			: category === "choice" ? "choice" : category === "result" ? "result" : "task";
	const newId = nextId(kind);
	const pos = node.position ?? { x: 0, y: 0 };
	const newNode: Node = {
		...node,
		id: newId,
		position: { x: pos.x + 30, y: pos.y + 30 },
		data: { ...data },
		class: node.class ?? "",
	};
	nodes.value = [...nodes.value, newNode];
	if (props.stateMachineId) {
		createNode(props.stateMachineId, {
			id: newNode.id,
			type: newNode.type ?? "flow",
			position: newNode.position ?? { x: 0, y: 0 },
			data: newNode.data as Record<string, unknown>,
		}).catch((e) => message.error(e instanceof Error ? e.message : "保存节点失败"));
	}
}

// ---------- 节点编辑弹窗 ----------
const showEditModal = ref(false);
const editingNodeId = ref<string | null>(null);
const editFormModel = ref<EditFormModel>({
	label: "",
	description: "",
	requestPath: "",
	requestMethod: "",
	requestData: "",
	outputCount: 1,
	options: [],
	inputCount: 1,
	results: [],
	nodeCategory: "scene",
});

function onNodeEdit(nodeId: string) {
	const node = nodes.value.find((n) => n.id === nodeId);
	if (!node) return;
	const data = node.data as FlowNodeData;
	editingNodeId.value = nodeId;
	const outputCount = Math.max(1, data.outputCount ?? 1);
	const inputCount = Math.max(1, data.inputCount ?? 1);
	let options = data.options ? JSON.parse(JSON.stringify(data.options)) : [];
	let results = data.results ? JSON.parse(JSON.stringify(data.results)) : [];
	// 补齐 options/results 长度与 outputCount/inputCount 一致，使表单与节点显示一致
	if (options.length < outputCount) {
		options = [...options, ...Array.from({ length: outputCount - options.length }, (_, i) => ({
			label: `选项 ${options.length + i + 1}`,
			description: "",
		}))];
	} else if (options.length > outputCount) {
		options = options.slice(0, outputCount);
	}
	if (results.length < inputCount) {
		results = [...results, ...Array.from({ length: inputCount - results.length }, (_, i) => ({
			label: `结果 ${results.length + i + 1}`,
			description: "",
		}))];
	} else if (results.length > inputCount) {
		results = results.slice(0, inputCount);
	}
	editFormModel.value = {
		label: data.label ?? "",
		description: data.description ?? "",
		requestPath: data.requestPath ?? "",
		requestMethod: data.requestMethod ?? "",
		requestData: data.requestData ?? "",
		outputCount,
		options,
		inputCount,
		results,
		nodeCategory: data.nodeCategory,
		nodeKind: data.nodeKind,
	};
	showEditModal.value = true;
}

function onOptionCountChange(val: number | null) {
	if (val === null) return;
	const count = Math.max(1, val);
	editFormModel.value.outputCount = count;
	const current = editFormModel.value.options;
	if (current.length < count) {
		editFormModel.value.options = [...current, ...Array.from({ length: count - current.length }, (_, i) => ({
			label: `选项 ${current.length + i + 1}`,
			description: "",
		}))];
	} else {
		editFormModel.value.options = current.slice(0, count);
	}
}

function onResultCountChange(val: number | null) {
	if (val === null) return;
	const count = Math.max(1, val);
	editFormModel.value.inputCount = count;
	const current = editFormModel.value.results;
	if (current.length < count) {
		editFormModel.value.results = [...current, ...Array.from({ length: count - current.length }, (_, i) => ({
			label: `结果 ${current.length + i + 1}`,
			description: "",
		}))];
	} else {
		editFormModel.value.results = current.slice(0, count);
	}
}

async function handleEditSave() {
	if (!editingNodeId.value) return;
	const nodeId = editingNodeId.value;
	nodes.value = nodes.value.map((n) => {
		if (n.id !== nodeId) return n;
		const newData: FlowNodeData = {
			...n.data,
			label: editFormModel.value.label,
			requestPath: editFormModel.value.requestPath || undefined,
			requestMethod: editFormModel.value.requestMethod || undefined,
			requestData: editFormModel.value.requestData || undefined,
		};
		if (editFormModel.value.nodeCategory === "scene") newData.description = editFormModel.value.description;
		else if (editFormModel.value.nodeCategory === "choice") {
			newData.outputCount = editFormModel.value.outputCount;
			newData.options = JSON.parse(JSON.stringify(editFormModel.value.options));
		} else if (editFormModel.value.nodeCategory === "result") {
			newData.inputCount = editFormModel.value.inputCount;
			newData.results = JSON.parse(JSON.stringify(editFormModel.value.results));
		} else if (editFormModel.value.nodeCategory === "task") {
			newData.description = editFormModel.value.description;
		}
		return { ...n, data: newData };
	});
	showEditModal.value = false;
	editingNodeId.value = null;

	// 若有状态机 id，调用更新节点接口持久化
	const flowId = props.stateMachineId;
	if (flowId) {
		const updatedNode = nodes.value.find((n) => n.id === nodeId);
		if (updatedNode) {
			try {
				await updateNode(flowId, nodeId, {
					id: updatedNode.id,
					type: updatedNode.type ?? "flow",
					position: updatedNode.position ?? { x: 0, y: 0 },
					data: updatedNode.data as Record<string, unknown>,
				});
				message.success("节点已保存");
			} catch (e) {
				message.error(e instanceof Error ? e.message : "保存节点失败");
			}
		}
	}
}

function onNodeDelete(nodeId: string) {
	nodes.value = nodes.value.filter((n) => n.id !== nodeId);
	edges.value = edges.value.filter((e) => e.source !== nodeId && e.target !== nodeId);
}

const sceneMenuChildren = computed(() => {
	const hasStart = hasNodeKind("start");
	const hasEnd = hasNodeKind("end");
	return [
		{ label: "开始", key: "scene-start", disabled: hasStart },
		{ label: "结束", key: "scene-end", disabled: hasEnd },
		{ label: "普通场景", key: "scene-default" },
	];
});

const addNodeOptions = computed<DropdownOption[]>(() => [
	{ label: "添加场景", key: "scene", type: "group", children: sceneMenuChildren.value },
	{ label: "添加选择", key: "choice" },
	{ label: "添加结果", key: "result" },
	{ label: "添加任务", key: "task" },
]);

const contextMenuOptions = computed<MenuOption[]>(() => [
	{ label: "添加场景", key: "scene", children: sceneMenuChildren.value },
	{ label: "添加选择", key: "choice" },
	{ label: "添加结果", key: "result" },
	{ label: "添加任务", key: "task" },
]);

// ---------- 测试运行 ----------
const RUN_INTERVAL_MS = 5000;
const runningNodeId = ref<string | null>(null);
type RunNodeState = "running" | "paused" | "completed" | "failed";
const currentNodeState = ref<RunNodeState | null>(null);
let runTimerId: ReturnType<typeof setInterval> | null = null;

function getFirstNodeId(): string | null {
	const targets = new Set(edges.value.map((e) => e.target));
	const entries = nodes.value.filter((n) => !targets.has(n.id));
	if (entries.length === 0) return null;
	const start = entries.find(
		(n) => (n.data as FlowNodeData)?.nodeCategory === "scene" && (n.data as FlowNodeData)?.nodeKind === "start"
	);
	return (start ?? entries[0]).id;
}

function getNextNodeIds(currentId: string): string[] {
	return [...new Set(
		edges.value.filter((e) => e.source === currentId).map((e) => e.target).filter((id) => nodes.value.some((n) => n.id === id))
	)];
}

function focusNode(nodeId: string) {
	nextTick(() => {
		if (findNode(nodeId)) fitView({ nodes: [nodeId], padding: 0.3, duration: 300 });
	});
}

function startTestRun() {
	if (runTimerId != null) {
		clearInterval(runTimerId);
		runTimerId = null;
		currentNodeState.value = "paused";
		runningNodeId.value = null;
		return;
	}
	const firstId = getFirstNodeId();
	if (firstId == null) return;
	runningNodeId.value = firstId;
	currentNodeState.value = "running";
	focusNode(firstId);
	runTimerId = setInterval(() => {
		const current = runningNodeId.value;
		if (current == null) return;
		const nextIds = getNextNodeIds(current);
		if (nextIds.length === 0) {
			if (runTimerId != null) clearInterval(runTimerId);
			runTimerId = null;
			currentNodeState.value = "completed";
			return;
		}
		runningNodeId.value = nextIds[Math.floor(Math.random() * nextIds.length)];
		focusNode(runningNodeId.value!);
	}, RUN_INTERVAL_MS);
}

function isFormNodeWithRun(data: FlowNodeData): boolean {
	if (!data?.nodeCategory) return false;
	if (data.nodeCategory === "scene" && data.nodeKind === "default") return true;
	if (data.nodeCategory === "choice" || data.nodeCategory === "result" || data.nodeCategory === "task") return true;
	return false;
}

const displayNodes = computed(() => {
	const runId = runningNodeId.value;
	const manualId = manualRunningNodeId.value;
	const state = currentNodeState.value;
	return nodes.value.map((n) => {
		const isAutoRunning = runId === n.id;
		const isManualRunning = manualId === n.id;
		const isRunning = isAutoRunning || isManualRunning;
		const d = n.data as FlowNodeData;
		const hasRun = isFormNodeWithRun(d);
		return {
			...n,
			class: [n.class ?? "", isRunning ? "running" : ""].filter(Boolean).join(" "),
			data: {
				...n.data,
				isRunning,
				nodeState: isAutoRunning ? state : null,
				...(hasRun && { onRun: () => runNodeRequest(n.id) }),
			},
		};
	});
});

// ---------- 保存流程 ----------
const saveLoading = ref(false);
const showSaveModal = ref(false);
const saveFormName = ref("");
const saveFormDescription = ref("");
const saveFormBaseUrl = ref("");

function openSaveModal() {
	saveFormName.value = props.stateMachineId ? stateMachineName.value || "未命名状态机" : "未命名状态机";
	saveFormDescription.value = stateMachineDescription.value;
	saveFormBaseUrl.value = stateMachineBaseUrl.value;
	showSaveModal.value = true;
}

async function doSaveConfirm() {
	const name = saveFormName.value.trim();
	if (!name) {
		message.warning("请输入流程图名称");
		return;
	}
	const description = saveFormDescription.value.trim();
	let id = props.stateMachineId;
	const isNew = !id;
	saveLoading.value = true;
	try {
		if (!id) {
			const created = await createStateMachine({ name, description: description || undefined });
			id = created.id;
			router.replace(`/flow/design/${id}`);
		} else {
			await updateStateMachine(id, { name, description: description || undefined, baseUrl: saveFormBaseUrl.value.trim() || undefined });
		}
		await saveFlow(id, { nodes: nodes.value, edges: edges.value });
		stateMachineName.value = name;
		stateMachineDescription.value = description;
		stateMachineBaseUrl.value = saveFormBaseUrl.value.trim();
		showSaveModal.value = false;
		message.success(isNew ? "新建并保存成功" : "保存成功");
	} catch (e) {
		message.error(e instanceof Error ? e.message : "保存失败");
	} finally {
		saveLoading.value = false;
	}
}

// ---------- 生命周期 ----------
onMounted(async () => {
	const id = props.stateMachineId;
	try {
		const data = await getFlowData(id);
		nodes.value = data.nodes.map((n) => ({ ...n, class: n.class || getNodeClass(n.data as FlowNodeData) }));
		edges.value = data.edges;
	} catch {
		nodes.value = [];
		edges.value = [];
	} finally {
		if (!hasNodeKind("start")) addNodeAt({ x: 100, y: 200 }, "scene-start");
		if (!hasNodeKind("end")) addNodeAt({ x: 500, y: 200 }, "scene-end");
	}
	if (id) {
		try {
			const detail = await getStateMachineById(id);
			stateMachineName.value = detail.name ?? "";
			stateMachineDescription.value = detail.description ?? "";
			stateMachineBaseUrl.value = detail.baseUrl ?? "";
		} catch {
			stateMachineName.value = "";
			stateMachineDescription.value = "";
			stateMachineBaseUrl.value = "";
		}
	}
});

onUnmounted(() => {
	if (runTimerId != null) {
		clearInterval(runTimerId);
		runTimerId = null;
	}
	currentNodeState.value = null;
	runningNodeId.value = null;
});
</script>

<template>
	<div style="height: 100%">
		<n-split direction="vertical" :default-size="0.7" style="height: 100%" :max="0.96">
			<template #1>
				<div style="height: 100%; position: relative">
					<VueFlow
						:fit-view-on-init="false"
						:nodes="displayNodes"
						:edges="edges"
						:class="{ 'vue-flow-dark': dark }"
						:style="{ backgroundColor: dark ? '#333' : '#EEE' }"
						@connect="onConnect"
						@nodes-change="onNodesChange"
						@edges-change="onEdgesChange"
						@pane-context-menu="onPaneContextMenu"
						@edge-context-menu="onEdgeContextMenu"
						@node-click="onNodeClick"
					>
						<template #node-flow="flowNodeProps">
							<FlowNode
								v-bind="flowNodeProps"
								:base-url="effectiveBaseUrl"
								@copy="onNodeCopy"
								@edit="onNodeEdit"
								@delete="onNodeDelete"
								@update:run-result="onNodeRunResult"
							/>
						</template>
						<Background />

						<div
							class="flow-toolbar flow-toolbar--left"
							style="position: absolute; left: 12px; top: 12px; z-index: 4; display: flex; align-items: center; gap: 8px"
						>
							<n-dropdown trigger="click" :options="addNodeOptions" @select="handleToolbarAdd">
								<n-button circle type="primary" title="添加节点">
									<template #icon><n-icon><Add /></n-icon></template>
								</n-button>
							</n-dropdown>
						</div>

						<Teleport to="body">
							<div
								v-if="contextMenuShow"
								class="flow-context-menu-backdrop"
								style="position: fixed; inset: 0; z-index: 9997"
								@click="handleContextMenuSelect('__close__')"
								@contextmenu.prevent="handleContextMenuSelect('__close__')"
							/>
							<div
								v-if="contextMenuShow"
								class="flow-context-menu-panel"
								:class="{ 'flow-context-menu-panel--dark': dark }"
								:style="{ position: 'fixed', left: contextMenuPosition.x + 'px', top: contextMenuPosition.y + 'px', zIndex: 9999, minWidth: '160px', maxHeight: '70vh', overflow: 'auto' }"
								@click.stop
							>
								<n-menu :options="contextMenuOptions" :value="null" @update:value="handleContextMenuSelect" />
							</div>
						</Teleport>

						<Teleport to="body">
							<div
								v-if="edgeContextMenuShow"
								class="flow-context-menu-backdrop"
								style="position: fixed; inset: 0; z-index: 9997"
								@click="closeEdgeContextMenu"
								@contextmenu.prevent="closeEdgeContextMenu"
							/>
							<div
								v-if="edgeContextMenuShow"
								class="flow-context-menu-panel"
								:class="{ 'flow-context-menu-panel--dark': dark }"
								:style="{ position: 'fixed', left: edgeContextMenuPosition.x + 'px', top: edgeContextMenuPosition.y + 'px', zIndex: 9999, minWidth: '120px' }"
								@click.stop
							>
								<n-menu :options="edgeContextMenuOptions" :value="null" @update:value="handleEdgeContextMenuSelect" />
							</div>
						</Teleport>

						<div style="position: absolute; right: 10px; top: 10px; z-index: 4; display: flex; align-items: center; gap: 8px">
							<n-button circle title="整理布局" @click="layoutGraph('LR')">
								<template #icon><n-icon><GridOutline /></n-icon></template>
							</n-button>
							<n-button circle :loading="saveLoading" title="保存" @click="openSaveModal">
								<template #icon><n-icon><SaveOutline /></n-icon></template>
							</n-button>
							<n-button
								circle
								:type="runningNodeId ? 'error' : 'primary'"
								:title="runningNodeId ? '停止测试运行' : '测试运行'"
								@click="startTestRun"
							>
								<template #icon>
									<n-icon><Stop v-if="runningNodeId" /><Play v-else /></n-icon>
								</template>
							</n-button>
						</div>

						<n-modal v-model:show="showSaveModal" :mask-closable="false">
							<n-card style="width: 400px" title="保存流程图" :bordered="false" role="dialog" aria-modal="true">
								<n-form-item label="流程图名称">
									<n-input v-model:value="saveFormName" placeholder="请输入流程图名称" clearable @keyup.enter="doSaveConfirm" />
								</n-form-item>
								<n-form-item label="描述">
									<n-input v-model:value="saveFormDescription" type="textarea" placeholder="请输入描述（选填）" :autosize="{ minRows: 2, maxRows: 6 }" clearable />
								</n-form-item>
								<n-form-item label="Base URL">
									<n-input v-model:value="saveFormBaseUrl" placeholder="请求基础地址，与节点请求路径拼接（选填）" clearable />
								</n-form-item>
								<template #footer>
									<div style="display: flex; justify-content: flex-end; gap: 8px">
										<n-button @click="showSaveModal = false">取消</n-button>
										<n-button type="primary" :loading="saveLoading" @click="doSaveConfirm">确定</n-button>
									</div>
								</template>
							</n-card>
						</n-modal>

						<n-modal v-model:show="showEditModal">
							<n-card style="width: 600px" title="编辑节点" :bordered="false" size="huge" role="dialog" aria-modal="true">
								<n-form :model="editFormModel" label-placement="left" label-width="auto" require-mark-placement="right-hanging">
									<n-form-item label="节点名称" path="label">
										<n-input v-model:value="editFormModel.label" placeholder="请输入节点显示内容" />
									</n-form-item>
									<template v-if="editFormModel.nodeCategory === 'scene'">
										<n-form-item label="描述" path="description">
											<n-input v-model:value="editFormModel.description" type="textarea" placeholder="请输入描述" :autosize="{ minRows: 2, maxRows: 6 }" />
										</n-form-item>
									</template>
									<template v-if="editFormModel.nodeCategory === 'task'">
										<n-form-item label="描述" path="description">
											<n-input v-model:value="editFormModel.description" type="textarea" placeholder="请输入描述" :autosize="{ minRows: 2, maxRows: 6 }" />
										</n-form-item>
									</template>
									<n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">请求配置</n-divider>
									<n-form-item label="请求路径" path="requestPath">
										<n-input v-model:value="editFormModel.requestPath" placeholder="如 /api/scene/enter，不含 base_url" clearable />
									</n-form-item>
									<n-form-item label="请求方法" path="requestMethod">
										<n-input v-model:value="editFormModel.requestMethod" placeholder="如 GET、POST" clearable />
									</n-form-item>
									<n-form-item label="请求体/参数" path="requestData">
										<n-input v-model:value="editFormModel.requestData" type="textarea" placeholder="JSON 或其它请求参数（选填）" :autosize="{ minRows: 2, maxRows: 6 }" clearable />
									</n-form-item>
									<template v-if="editFormModel.nodeCategory === 'choice'">
										<n-form-item label="选项个数">
											<n-input-number :value="editFormModel.outputCount" :min="1" style="width: 100%" @update:value="onOptionCountChange" />
										</n-form-item>
										<n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">选项列表</n-divider>
										<div v-for="(opt, index) in editFormModel.options" :key="index" style="border: 1px solid var(--n-border-color); border-radius: 4px; padding: 8px; margin-bottom: 8px">
											<div style="font-size: 12px; margin-bottom: 4px; opacity: 0.8">选项 {{ index + 1 }}</div>
											<n-form-item label="名称" label-placement="left" :show-feedback="false" style="margin-bottom: 8px">
												<n-input v-model:value="opt.label" placeholder="选项名称" size="small" />
											</n-form-item>
											<n-form-item label="描述" label-placement="left" :show-feedback="false">
												<n-input v-model:value="opt.description" placeholder="选项描述" size="small" />
											</n-form-item>
										</div>
									</template>
									<template v-if="editFormModel.nodeCategory === 'result'">
										<n-form-item label="结果数量">
											<n-input-number :value="editFormModel.inputCount" :min="1" style="width: 100%" @update:value="onResultCountChange" />
										</n-form-item>
										<n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">结果列表</n-divider>
										<div v-for="(res, index) in editFormModel.results" :key="index" style="border: 1px solid var(--n-border-color); border-radius: 4px; padding: 8px; margin-bottom: 8px">
											<div style="font-size: 12px; margin-bottom: 4px; opacity: 0.8">结果 {{ index + 1 }}</div>
											<n-form-item label="名称" label-placement="left" :show-feedback="false" style="margin-bottom: 8px">
												<n-input v-model:value="res.label" placeholder="结果名称" size="small" />
											</n-form-item>
											<n-form-item label="描述" label-placement="left" :show-feedback="false">
												<n-input v-model:value="res.description" placeholder="结果描述" size="small" />
											</n-form-item>
										</div>
									</template>
								</n-form>
								<template #footer>
									<div style="display: flex; justify-content: flex-end; gap: 8px">
										<n-button @click="showEditModal = false">取消</n-button>
										<n-button type="primary" @click="handleEditSave">保存</n-button>
									</div>
								</template>
							</n-card>
						</n-modal>
					</VueFlow>
				</div>
			</template>
			<template #2>
				<FlowLogPanel :selection-node="currentSelectionNode" />
			</template>
		</n-split>
	</div>
</template>

<style scoped>
.flow-context-menu-backdrop { cursor: default; }
.flow-context-menu-panel {
	background: var(--n-color);
	border-radius: 8px;
	box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	padding: 4px 0;
}
.flow-context-menu-panel--dark { background: #2d2d2d; }
</style>
