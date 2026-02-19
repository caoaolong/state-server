<script setup lang="ts">
<<<<<<< HEAD
import { ref, onMounted, onUnmounted, computed, nextTick, inject, reactive } from "vue";
import { useRoute, useRouter } from "vue-router";
import { type Node, type Edge, type Connection, type NodeChange, type EdgeChange, useVueFlow } from "@vue-flow/core";
import { VueFlow } from "@vue-flow/core";
import { Background } from "@vue-flow/background";
import { NButton, NIcon, NDropdown, NMenu, NModal, NCard, NForm, NFormItem, NInput, NInputNumber, NDivider, useMessage, NSplit, NTabs, NTabPane, NCode, NEmpty } from "naive-ui";
import type { DropdownOption } from "naive-ui";
import type { MenuOption } from "naive-ui";
import { Play, Stop, Add, GridOutline, SaveOutline } from "@vicons/ionicons5";
import { getFlowData, getStateMachineById, saveFlow, createStateMachine, updateStateMachine } from "../api";
import { useLayout } from "../core/layout";
import FlowNode from "./FlowNode.vue";
import type { FlowNodeData } from "./FlowNode.vue";
=======
/**
 * 流程编辑器：仅提供路由参数与主题，具体逻辑在 FlowCanvas 内部实现
 */
import { computed, inject, ref } from "vue";
import { useRoute } from "vue-router";
import FlowCanvas from "./FlowCanvas.vue";
>>>>>>> 7921434093a2c8557483be064053f6e791d6c826

const route = useRoute();
const dark = inject<ReturnType<typeof ref<boolean>>>("app-dark-mode", ref(false));

<<<<<<< HEAD
/** 状态机信息对象 */
interface StateMachineInfo {
  /** 状态机 ID（来自路由 /state-machines/design/:id），无则为新建 */
  id: string | undefined;
  /** 状态机名称（编辑时从接口加载，用于保存对话框预填） */
  name: string;
  /** 状态机描述（编辑时从接口加载，用于保存对话框预填） */
  description: string;
  /** 状态机 baseUrl（请求基础地址，与节点请求路径拼接） */
  baseUrl: string;
  /** 状态机标识符（用于唯一识别状态机） */
  identifier: string;
}

const stateMachineInfo = reactive<StateMachineInfo>({
  id: undefined,
  name: "",
  description: "",
  baseUrl: "",
  identifier: "",
});

/** 当前编辑的状态机 ID（来自路由 /state-machines/design/:id），无则为新建 - 计算属性 */
const stateMachineId = computed(() => route.params.id as string | undefined);

const nodes = ref<Node[]>([]);
const edges = ref<Edge[]>([]);
const { layout } = useLayout();
const { fitView, findNode, project } = useVueFlow();

// 当前选中的节点ID
const currentSelectionId = ref<string | null>(null);

// 获取当前选中节点的 data
const currentSelectionNode = computed(() => {
  if (!currentSelectionId.value) return null;
  return nodes.value.find((n) => n.id === currentSelectionId.value);
});

// 节点点击事件
function onNodeClick(event: { node: Node }) {
  currentSelectionId.value = event.node.id;
}

// 侧边栏点击或其他区域点击时，可能需要清除选中？暂时不需要，保持选中方便查看日志

const layoutGraph = async (direction: string) => {
	nodes.value = layout(nodes.value, edges.value, direction);
	nextTick(() => fitView());
};

// 节点连线：用户从 handle 拖到另一节点 handle 时添加边
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

// 将拖拽后的节点位置同步回 nodes，避免添加新节点时已有节点位置被“重置”
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

// 边变更：选中边后按 Delete/Backspace 删除时同步到 edges
function onEdgesChange(changes: EdgeChange[]) {
	const removeIds = new Set(
		changes.filter((c): c is EdgeChange & { type: "remove" } => c.type === "remove").map((c) => c.id)
	);
	if (removeIds.size === 0) return;
	edges.value = edges.value.filter((e) => !removeIds.has(e.id));
}

// 节点创建：唯一 id 计数
const idCounters = ref({ sceneStart: 0, sceneEnd: 0, sceneDefault: 0, choice: 0, result: 0 });
function nextId(kind: keyof typeof idCounters.value): string {
	idCounters.value[kind]++;
	const n = idCounters.value[kind];
	const prefix =
		kind === "sceneStart"
			? "scene-start"
			: kind === "sceneEnd"
				? "scene-end"
				: kind === "sceneDefault"
					? "scene"
					: kind === "choice"
						? "choice"
						: "result";
	return `${prefix}-${n}`;
}

function hasNodeKind(kind: string): boolean {
	return nodes.value.some(
		(n) =>
			(n.data as FlowNodeData)?.nodeCategory === "scene" &&
			(n.data as FlowNodeData)?.nodeKind === kind
	);
}

/** 根据节点 data 计算样式 class（回显时恢复颜色） */
function getNodeClass(data: FlowNodeData | undefined): string {
	if (!data) return "";
	const cat = data.nodeCategory;
	const kind = data.nodeKind;
	if (cat === "scene") return `flow-type-scene flow-type-scene-${kind ?? "default"}`;
	if (cat === "choice") return "flow-type-choice";
	if (cat === "result") return "flow-type-result";
	return "";
}

/** 在画布坐标 position 处创建节点 */
function addNodeAt(
	position: { x: number; y: number },
	option: "scene-start" | "scene-end" | "scene-default" | "choice" | "result"
) {
	if (option === "scene-start" && hasNodeKind("start")) return;
	if (option === "scene-end" && hasNodeKind("end")) return;

	const id =
		option === "scene-start"
			? nextId("sceneStart")
			: option === "scene-end"
				? nextId("sceneEnd")
				: option === "scene-default"
					? nextId("sceneDefault")
					: option === "choice"
						? nextId("choice")
						: nextId("result");

	const labels: Record<typeof option, string> = {
		"scene-start": "开始",
		"scene-end": "结束",
		"scene-default": "普通场景",
		choice: "选择",
		result: "结果",
	};

	let data: FlowNodeData;
	let nodeClass = "";

	if (option.startsWith("scene-")) {
		const nodeKind =
			option === "scene-start" ? "start" : option === "scene-end" ? "end" : "default";
		data = {
			label: labels[option],
			nodeCategory: "scene",
			nodeKind,
		};
		nodeClass = `flow-type-scene flow-type-scene-${nodeKind}`;
	} else if (option === "choice") {
		data = {
			label: labels.choice,
			nodeCategory: "choice",
			outputCount: 2,
		};
		nodeClass = "flow-type-choice";
	} else {
		data = {
			label: labels.result,
			nodeCategory: "result",
			inputCount: 2,
		};
		nodeClass = "flow-type-result";
	}

	const newNode: Node = {
		id,
		type: "flow",
		position,
		data,
		class: nodeClass,
	};
	nodes.value = [...nodes.value, newNode];
}

// 右键菜单：画布坐标
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
	if (lastContextEvent && (key === "scene-start" || key === "scene-end" || key === "scene-default" || key === "choice" || key === "result")) {
		const flowPos = project({ x: lastContextEvent.clientX, y: lastContextEvent.clientY });
		addNodeAt(flowPos, key);
	}
	contextMenuShow.value = false;
	lastContextEvent = null;
}

// 边的右键菜单：删除连线
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

function handleEdgeContextMenuSelect(key: string) {
	if (key === "delete-edge" && edgeContextMenuEdgeId.value) {
		edges.value = edges.value.filter((e) => e.id !== edgeContextMenuEdgeId.value);
	}
	closeEdgeContextMenu();
}

function closeEdgeContextMenu() {
	edgeContextMenuShow.value = false;
	edgeContextMenuEdgeId.value = null;
}

// 左上角悬浮按钮菜单（在画布默认位置创建）
const DEFAULT_ADD_POSITION = { x: 200, y: 150 };
function handleToolbarAdd(key: string) {
	if (key === "scene-start" || key === "scene-end" || key === "scene-default" || key === "choice" || key === "result") {
		addNodeAt({ ...DEFAULT_ADD_POSITION }, key);
	}
}

// 节点工具栏：复制、编辑、删除
function onNodeCopy(nodeId: string) {
	const node = nodes.value.find((n) => n.id === nodeId);
	if (!node) return;
	const data = node.data as FlowNodeData;

	// 禁止复制开始和结束节点
	if (data?.nodeCategory === "scene" && (data?.nodeKind === "start" || data?.nodeKind === "end")) {
		return;
	}

	const category = data?.nodeCategory ?? "scene";
	const kind =
		category === "scene"
			? data?.nodeKind === "start"
				? "sceneStart"
				: data?.nodeKind === "end"
					? "sceneEnd"
					: "sceneDefault"
			: category === "choice"
				? "choice"
				: "result";
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
}

// 节点编辑
const showEditModal = ref(false);
const editingNodeId = ref<string | null>(null);

// 定义表单模型接口
interface EditFormModel {
	label: string;
	description: string;
	// 请求配置（不包含 base_url）
	requestPath: string;
	requestMethod: string;
	requestData: string;
	// Choice
	outputCount: number;
	options: Array<{ label: string; description: string }>;
	// Result
	inputCount: number;
	results: Array<{ label: string; description: string }>;
	// Meta
	nodeCategory: string;
	nodeKind?: string;
}

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
	// Deep copy relevant data to form model
	editFormModel.value = {
		label: data.label ?? "",
		description: data.description ?? "",
		requestPath: data.requestPath ?? "",
		requestMethod: data.requestMethod ?? "",
		requestData: data.requestData ?? "",
		outputCount: data.outputCount ?? 1,
		options: data.options ? JSON.parse(JSON.stringify(data.options)) : [],
		inputCount: data.inputCount ?? 1,
		results: data.results ? JSON.parse(JSON.stringify(data.results)) : [],
		nodeCategory: data.nodeCategory,
		nodeKind: data.nodeKind,
	};

	showEditModal.value = true;
}

// 监听 outputCount 变化，动态调整 options 数组
function onOptionCountChange(val: number | null) {
	if (val === null) return;
	const count = Math.max(1, val);
	editFormModel.value.outputCount = count;

	const current = editFormModel.value.options;
	if (current.length < count) {
		const add = Array.from({ length: count - current.length }, (_, i) => ({
			label: `选项 ${current.length + i + 1}`,
			description: "",
		}));
		editFormModel.value.options = [...current, ...add];
	} else if (current.length > count) {
		editFormModel.value.options = current.slice(0, count);
	}
}

// 监听 inputCount 变化，动态调整 results 数组
function onResultCountChange(val: number | null) {
	if (val === null) return;
	const count = Math.max(1, val);
	editFormModel.value.inputCount = count;

	const current = editFormModel.value.results;
	if (current.length < count) {
		const add = Array.from({ length: count - current.length }, (_, i) => ({
			label: `结果 ${current.length + i + 1}`,
			description: "",
		}));
		editFormModel.value.results = [...current, ...add];
	} else if (current.length > count) {
		editFormModel.value.results = current.slice(0, count);
	}
}

function handleEditSave() {
	if (!editingNodeId.value) return;

	nodes.value = nodes.value.map((n) => {
		if (n.id === editingNodeId.value) {
			const newData: FlowNodeData = {
				...n.data,
				label: editFormModel.value.label,
				requestPath: editFormModel.value.requestPath || undefined,
				requestMethod: editFormModel.value.requestMethod || undefined,
				requestData: editFormModel.value.requestData || undefined,
			};

			// Update specific fields based on type
			if (editFormModel.value.nodeCategory === "scene") {
				// 所有场景类型（包括开始、结束、普通）都可以编辑描述
				newData.description = editFormModel.value.description;
			} else if (editFormModel.value.nodeCategory === "choice") {
				newData.outputCount = editFormModel.value.outputCount;
				newData.options = JSON.parse(JSON.stringify(editFormModel.value.options));
			} else if (editFormModel.value.nodeCategory === "result") {
				newData.inputCount = editFormModel.value.inputCount;
				newData.results = JSON.parse(JSON.stringify(editFormModel.value.results));
			}

			return {
				...n,
				data: newData,
			};
		}
		return n;
	});

	showEditModal.value = false;
	editingNodeId.value = null;
}

function onNodeDelete(nodeId: string) {
	nodes.value = nodes.value.filter((n) => n.id !== nodeId);
	edges.value = edges.value.filter(
		(e) => e.source !== nodeId && e.target !== nodeId
	);
}

const addNodeOptions = computed<DropdownOption[]>(() => {
	const hasStart = hasNodeKind("start");
	const hasEnd = hasNodeKind("end");
	return [
		{
			label: "添加场景",
			key: "scene",
			type: "group",
			children: [
				{ label: "开始", key: "scene-start", disabled: hasStart },
				{ label: "结束", key: "scene-end", disabled: hasEnd },
				{ label: "普通场景", key: "scene-default" },
			],
		},
		{ label: "添加选择", key: "choice" },
		{ label: "添加结果", key: "result" },
	];
});

const contextMenuOptions = computed<MenuOption[]>(() => {
	const hasStart = hasNodeKind("start");
	const hasEnd = hasNodeKind("end");
	return [
		{
			label: "添加场景",
			key: "scene",
			children: [
				{ label: "开始", key: "scene-start", disabled: hasStart },
				{ label: "结束", key: "scene-end", disabled: hasEnd },
				{ label: "普通场景", key: "scene-default" },
			],
		},
		{ label: "添加选择", key: "choice" },
		{ label: "添加结果", key: "result" },
	];
});

// 测试运行
const runningNodeId = ref<string | null>(null);
const manualRunningNodeId = ref<string | null>(null);
type NodeState = "running" | "paused" | "completed" | "failed";
const currentNodeState = ref<NodeState | null>(null);
let runTimerId: ReturnType<typeof setInterval> | null = null;
const RUN_INTERVAL_MS = 5000;

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
	const nextIds = edges.value
		.filter((e) => e.source === currentId)
		.map((e) => e.target)
		.filter((id) => nodes.value.some((n) => n.id === id));
	return [...new Set(nextIds)];
}

function focusNode(nodeId: string) {
	nextTick(() => {
		const graphNode = findNode(nodeId);
		if (graphNode) fitView({ nodes: [nodeId], padding: 0.3, duration: 300 });
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
		const nextId = nextIds[Math.floor(Math.random() * nextIds.length)];
		runningNodeId.value = nextId;
		focusNode(nextId);
	}, RUN_INTERVAL_MS);
}

const displayNodes = computed(() => {
	const runId = runningNodeId.value;
    const manualId = manualRunningNodeId.value;
	const state = currentNodeState.value;
	return nodes.value.map((n) => {
		const isAutoRunning = runId === n.id;
        const isManualRunning = manualId === n.id;
        const isRunning = isAutoRunning || isManualRunning;
		const baseClass = n.class ?? "";
		return {
			...n,
			class: [baseClass, isRunning ? "running" : ""].filter(Boolean).join(" "),
			data: {
				...n.data,
				isRunning,
				nodeState: isAutoRunning ? state : null,
				...(typeof (n.data as { onRun?: () => void })?.onRun === "function" && {
					onRun: (n.data as { onRun: () => void }).onRun,
				}),
			},
		};
	});
});

const saveLoading = ref(false);
const showSaveModal = ref(false);

function openSaveModal() {
	if (!stateMachineInfo.name) {
		stateMachineInfo.name = "未命名状态机";
	}
	showSaveModal.value = true;
}

async function doSaveConfirm() {
	const name = stateMachineInfo.name.trim();
	if (!name) {
		message.warning("请输入流程图名称");
		return;
	}
	const description = stateMachineInfo.description.trim();
	let id = stateMachineInfo.id;
	const isNew = !id;
	saveLoading.value = true;
	try {
		if (!id) {
			const created = await createStateMachine({ name, description: description || undefined });
			id = created.id;
			stateMachineInfo.id = id;
			router.replace(`/state-machines/design/${id}`);
		} else {
			const baseUrl = stateMachineInfo.baseUrl.trim();
			await updateStateMachine(id, {
				name,
				description: description || undefined,
				baseUrl: baseUrl || undefined,
				identifier: stateMachineInfo.identifier.trim() || undefined,
			});
		}
		await saveFlow(id, { nodes: nodes.value, edges: edges.value });
		stateMachineInfo.name = name;
		stateMachineInfo.description = description;
		stateMachineInfo.baseUrl = stateMachineInfo.baseUrl.trim();
		showSaveModal.value = false;
		message.success(isNew ? "新建并保存成功" : "保存成功");
	} catch (e) {
		message.error(e instanceof Error ? e.message : "保存失败");
	} finally {
		saveLoading.value = false;
	}
}

onMounted(async () => {
	const id = stateMachineId.value;
	try {
		const data = await getFlowData(id);
		nodes.value = data.nodes.map((n) => ({
			...n,
			class: n.class || getNodeClass(n.data as FlowNodeData),
		}));
		console.log(nodes.value);
		edges.value = data.edges;
	} catch {
		nodes.value = [];
		edges.value = [];
	} finally {
		if (!hasNodeKind("start")) {
			addNodeAt({ x: 100, y: 200 }, "scene-start");
		}
		if (!hasNodeKind("end")) {
			addNodeAt({ x: 500, y: 200 }, "scene-end");
		}
	}
	if (id) {
		try {
			const detail = await getStateMachineById(id);
			stateMachineInfo.id = detail.id;
			stateMachineInfo.name = detail.name ?? "";
			stateMachineInfo.description = detail.description ?? "";
			stateMachineInfo.baseUrl = detail.baseUrl ?? "";
			stateMachineInfo.identifier = detail.identifier; // 使用 ID 作为标识符
		} catch {
			stateMachineInfo.id = undefined;
			stateMachineInfo.name = "";
			stateMachineInfo.description = "";
			stateMachineInfo.baseUrl = "";
			stateMachineInfo.identifier = "";
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
=======
const stateMachineId = computed(() => route.params.id as string | undefined);
>>>>>>> 7921434093a2c8557483be064053f6e791d6c826
</script>

<template>
	<div style="height: 100%">
<<<<<<< HEAD
		<n-split direction="vertical" :default-size="0.7" style="height: 100%">
			<template #1>
				<div style="height: 100%; position: relative;">
					<VueFlow :fit-view-on-init="false" @connect="onConnect" @nodes-change="onNodesChange"
						@edges-change="onEdgesChange" @pane-context-menu="onPaneContextMenu"
						@edge-context-menu="onEdgeContextMenu" @node-click="onNodeClick" :nodes="displayNodes" :edges="edges"
						:class="{ 'vue-flow-dark': dark }" :style="{ backgroundColor: dark ? '#333' : '#EEE' }">
						<template #node-flow="flowNodeProps">
							<FlowNode v-bind="flowNodeProps" :base-url="stateMachineInfo.baseUrl" @copy="onNodeCopy" @edit="onNodeEdit"
								@delete="onNodeDelete" />
						</template>
						<Background />

						<!-- 左上角：带菜单的悬浮按钮 -->
						<div class="flow-toolbar flow-toolbar--left" style="
          position: absolute;
          left: 12px;
          top: 12px;
          z-index: 4;
          display: flex;
          align-items: center;
          gap: 8px;
        ">
							<n-dropdown trigger="click" :options="addNodeOptions" @select="handleToolbarAdd">
								<n-button circle type="primary" title="添加节点">
									<template #icon>
										<n-icon>
											<Add />
										</n-icon>
									</template>
								</n-button>
							</n-dropdown>
						</div>

						<!-- 右键菜单：固定位置 NMenu -->
						<Teleport to="body">
							<div v-if="contextMenuShow" class="flow-context-menu-backdrop"
								style="position: fixed; inset: 0; z-index: 9997;" @click="contextMenuShow = false"
								@contextmenu.prevent="contextMenuShow = false" />
							<div v-if="contextMenuShow" class="flow-context-menu-panel"
								:class="{ 'flow-context-menu-panel--dark': dark }" :style="{
              position: 'fixed',
              left: contextMenuPosition.x + 'px',
              top: contextMenuPosition.y + 'px',
              zIndex: 9999,
              minWidth: '160px',
              maxHeight: '70vh',
              overflow: 'auto',
            }" @click.stop>
								<n-menu :options="contextMenuOptions" :value="null" @update:value="handleContextMenuSelect" />
							</div>
						</Teleport>

						<!-- 边的右键菜单：删除连线 -->
						<Teleport to="body">
							<div v-if="edgeContextMenuShow" class="flow-context-menu-backdrop"
								style="position: fixed; inset: 0; z-index: 9997;" @click="closeEdgeContextMenu"
								@contextmenu.prevent="edgeContextMenuShow = false" />
							<div v-if="edgeContextMenuShow" class="flow-context-menu-panel"
								:class="{ 'flow-context-menu-panel--dark': dark }" :style="{
              position: 'fixed',
              left: edgeContextMenuPosition.x + 'px',
              top: edgeContextMenuPosition.y + 'px',
              zIndex: 9999,
              minWidth: '120px',
            }" @click.stop>
								<n-menu :options="edgeContextMenuOptions" :value="null" @update:value="handleEdgeContextMenuSelect" />
							</div>
						</Teleport>

						<div style="
          position: absolute;
          right: 10px;
          top: 10px;
          z-index: 4;
          display: flex;
          align-items: center;
          gap: 8px;
        ">
							<n-button circle title="整理布局" @click="layoutGraph('LR')">
								<template #icon>
									<n-icon>
										<GridOutline />
									</n-icon>
								</template>
							</n-button>
							<n-button circle type="primary" :loading="saveLoading" title="保存流程（无 ID 时先新建再保存）" @click="openSaveModal">
								<template #icon>
									<n-icon>
										<SaveOutline />
									</n-icon>
								</template>
							</n-button>
							<n-button circle :type="runningNodeId ? 'error' : 'primary'"
								:title="runningNodeId ? '停止测试运行' : '测试运行'" @click="startTestRun">
								<template #icon>
									<n-icon>
										<Stop v-if="runningNodeId" />
										<Play v-else />
									</n-icon>
								</template>
							</n-button>
						</div>

						<!-- 保存流程图对话框 -->
						<n-modal v-model:show="showSaveModal" :mask-closable="false">
							<n-card style="width: 400px" title="保存流程图" :bordered="false" role="dialog" aria-modal="true">
								<n-form-item label="流程图名称">
								<n-input v-model:value="stateMachineInfo.name" placeholder="请输入流程图名称" clearable
									@keyup.enter="doSaveConfirm" />
							</n-form-item>
							<n-form-item label="描述">
								<n-input v-model:value="stateMachineInfo.description" type="textarea" placeholder="请输入描述（选填）"
									:autosize="{ minRows: 2, maxRows: 6 }" clearable />
							</n-form-item>
							<n-form-item label="标识符">
								<n-input v-model:value="stateMachineInfo.identifier" placeholder="请输入标识符" clearable />
							</n-form-item>
							<n-form-item label="Base URL">
								<n-input v-model:value="stateMachineInfo.baseUrl" placeholder="请求基础地址，与节点请求路径拼接（选填）" clearable />
							</n-form-item>
							<template #footer>
								<div style="display: flex; justify-content: flex-end; gap: 8px">
									<n-button @click="showSaveModal = false">取消</n-button>
									<n-button type="primary" :loading="saveLoading" @click="doSaveConfirm">确定</n-button>
								</div>
							</template>
						</n-card>
					</n-modal>

						<!-- 节点编辑模态框 -->
						<n-modal v-model:show="showEditModal">
							<n-card style="width: 600px" title="编辑节点" :bordered="false" size="huge" role="dialog"
								aria-modal="true">
								<n-form ref="formRef" :model="editFormModel" label-placement="left" label-width="auto"
									require-mark-placement="right-hanging">
									<!-- 公共项：节点名称 -->
									<n-form-item label="节点名称" path="label">
										<n-input v-model:value="editFormModel.label" placeholder="请输入节点显示内容" />
									</n-form-item>

									<!-- 普通场景、开始、结束：描述 -->
									<template v-if="editFormModel.nodeCategory === 'scene'">
										<n-form-item label="描述" path="description">
											<n-input v-model:value="editFormModel.description" type="textarea" placeholder="请输入描述"
												:autosize="{ minRows: 2, maxRows: 6 }" />
										</n-form-item>
									</template>

									<!-- 请求配置（不包含 base_url） -->
									<n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">
										请求配置
									</n-divider>
									<n-form-item label="请求路径" path="requestPath">
										<n-input v-model:value="editFormModel.requestPath" placeholder="如 /api/scene/enter，不含 base_url"
											clearable />
									</n-form-item>
									<n-form-item label="请求方法" path="requestMethod">
										<n-input v-model:value="editFormModel.requestMethod" placeholder="如 GET、POST" clearable />
									</n-form-item>
									<n-form-item label="请求体/参数" path="requestData">
										<n-input v-model:value="editFormModel.requestData" type="textarea"
											placeholder="JSON 或其它请求参数（选填）" :autosize="{ minRows: 2, maxRows: 6 }" clearable />
									</n-form-item>

									<!-- 选择节点 -->
									<template v-if="editFormModel.nodeCategory === 'choice'">
										<n-form-item label="选项个数">
											<n-input-number :value="editFormModel.outputCount" @update:value="onOptionCountChange"
												:min="1" style="width: 100%" />
										</n-form-item>
										<n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">
											选项列表
										</n-divider>
										<div v-for="(opt, index) in editFormModel.options" :key="index" style="
                  border: 1px solid var(--n-border-color);
                  border-radius: 4px;
                  padding: 8px;
                  margin-bottom: 8px;
                ">
											<div style="font-size: 12px; margin-bottom: 4px; opacity: 0.8">
												选项 {{ index + 1 }}
											</div>
											<n-form-item label="名称" label-placement="left" :show-feedback="false"
												style="margin-bottom: 8px">
												<n-input v-model:value="opt.label" placeholder="选项名称" size="small" />
											</n-form-item>
											<n-form-item label="描述" label-placement="left" :show-feedback="false">
												<n-input v-model:value="opt.description" placeholder="选项描述" size="small" />
											</n-form-item>
										</div>
									</template>

									<!-- 结果节点 -->
									<template v-if="editFormModel.nodeCategory === 'result'">
										<n-form-item label="结果数量">
											<n-input-number :value="editFormModel.inputCount" @update:value="onResultCountChange"
												:min="1" style="width: 100%" />
										</n-form-item>
										<n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">
											结果列表
										</n-divider>
										<div v-for="(res, index) in editFormModel.results" :key="index" style="
                  border: 1px solid var(--n-border-color);
                  border-radius: 4px;
                  padding: 8px;
                  margin-bottom: 8px;
                ">
											<div style="font-size: 12px; margin-bottom: 4px; opacity: 0.8">
												结果 {{ index + 1 }}
											</div>
											<n-form-item label="名称" label-placement="left" :show-feedback="false"
												style="margin-bottom: 8px">
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
				<n-tabs type="line" style="height: 100%;" pane-style="height: calc(100% - 40px); overflow: auto; padding: 12px;">
					<n-tab-pane name="log" tab="日志">
						<template v-if="currentSelectionNode">
							<div style="display: flex; justify-content: start; gap: 20px; align-items: center; margin-bottom: 12px;">
								<span style="font-weight: bold; font-size: 14px;">{{ currentSelectionNode.data.label || '未命名节点' }}</span>
							</div>
							<div v-if="currentSelectionNode.data.requestData || currentSelectionNode.data.runResult">
								<div v-if="currentSelectionNode.data.requestData">
									<n-divider title-placement="left"
										style="margin-top: 0; margin-bottom: 8px; font-size: 12px;">请求体</n-divider>
									<n-code :code="currentSelectionNode.data.requestData" language="json" word-wrap
										style="font-size: 12px; margin-bottom: 12px;" />
								</div>
								<div v-if="currentSelectionNode.data.runResult">
									<n-divider title-placement="left"
										style="margin-top: 0; margin-bottom: 8px; font-size: 12px;">响应结果</n-divider>
									<n-code :code="currentSelectionNode.data.runResult" language="json" word-wrap
										style="font-size: 12px;" />
								</div>
							</div>
							<n-empty v-else description="该节点暂无请求或运行数据" style="margin-top: 24px;" />
						</template>
						<n-empty v-else description="请选择一个节点以查看日志" style="margin-top: 24px;" />
					</n-tab-pane>
				</n-tabs>
			</template>
		</n-split>
=======
		<FlowCanvas :state-machine-id="stateMachineId" :dark="dark" />
>>>>>>> 7921434093a2c8557483be064053f6e791d6c826
	</div>
</template>

<style>
/* VueFlow 全局样式如需可保留 */
</style>
