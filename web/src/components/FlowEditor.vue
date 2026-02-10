<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, nextTick, inject } from "vue";
import { useRoute, useRouter } from "vue-router";
import { type Node, type Edge, type Connection, type NodeChange, type EdgeChange, useVueFlow } from "@vue-flow/core";
import { VueFlow } from "@vue-flow/core";
import { Background } from "@vue-flow/background";
import { NButton, NIcon, NDropdown, NMenu, NModal, NCard, NForm, NFormItem, NInput, NInputNumber, NDivider, useMessage } from "naive-ui";
import type { DropdownOption } from "naive-ui";
import type { MenuOption } from "naive-ui";
import { Play, Stop, Add, GridOutline, SaveOutline } from "@vicons/ionicons5";
import { getFlowData, getStateMachineById, saveFlow, createStateMachine, updateStateMachine } from "../api";
import { useLayout } from "../core/layout";
import FlowNode from "./FlowNode.vue";
import type { FlowNodeData } from "./FlowNode.vue";

const route = useRoute();
const router = useRouter();
const message = useMessage();
const dark = inject('app-dark-mode', ref(false));

/** 当前编辑的状态机 ID（来自路由 /state-machines/design/:id），无则为新建 */
const stateMachineId = computed(() => route.params.id as string | undefined);
/** 当前状态机名称（编辑时从接口加载，用于保存对话框预填） */
const stateMachineName = ref("");
/** 当前状态机描述（编辑时从接口加载，用于保存对话框预填） */
const stateMachineDescription = ref("");
/** 当前状态机 baseUrl（请求基础地址，与节点请求路径拼接） */
const stateMachineBaseUrl = ref("");

const nodes = ref<Node[]>([]);
const edges = ref<Edge[]>([]);
const { layout } = useLayout();
const { fitView, findNode, project } = useVueFlow();

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
  
  // 禁止编辑开始和结束节点
  if (data?.nodeCategory === "scene" && (data?.nodeKind === "start" || data?.nodeKind === "end")) {
    return;
  }
  
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
      if (editFormModel.value.nodeCategory === "scene" && editFormModel.value.nodeKind === "default") {
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
  const state = currentNodeState.value;
  return nodes.value.map((n) => {
    const isRunning = runId === n.id;
    const baseClass = n.class ?? "";
    return {
      ...n,
      class: [baseClass, isRunning ? "running" : ""].filter(Boolean).join(" "),
      data: {
        ...n.data,
        isRunning,
        nodeState: isRunning ? state : null,
        ...(typeof (n.data as { onRun?: () => void })?.onRun === "function" && {
          onRun: (n.data as { onRun: () => void }).onRun,
        }),
      },
    };
  });
});

const saveLoading = ref(false);
const showSaveModal = ref(false);
const saveFormName = ref("");
const saveFormDescription = ref("");
const saveFormBaseUrl = ref("");

function openSaveModal() {
  saveFormName.value = stateMachineId.value ? stateMachineName.value || "未命名状态机" : "未命名状态机";
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
  let id = stateMachineId.value;
  const isNew = !id;
  saveLoading.value = true;
  try {
    if (!id) {
      const created = await createStateMachine({ name, description: description || undefined });
      id = created.id;
      router.replace(`/state-machines/design/${id}`);
    } else {
      await updateStateMachine(id, {
        name,
        description: description || undefined,
        baseUrl: saveFormBaseUrl.value.trim() || undefined,
      });
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
    <VueFlow
      :fit-view-on-init="false"
      @connect="onConnect"
      @nodes-change="onNodesChange"
      @edges-change="onEdgesChange"
      @pane-context-menu="onPaneContextMenu"
      @edge-context-menu="onEdgeContextMenu"
      :nodes="displayNodes"
      :edges="edges"
      :class="{ 'vue-flow-dark': dark }"
      :style="{ backgroundColor: dark ? '#333' : '#EEE' }"
    >
      <template #node-flow="flowNodeProps">
        <FlowNode
          v-bind="flowNodeProps"
          @copy="onNodeCopy"
          @edit="onNodeEdit"
          @delete="onNodeDelete"
        />
      </template>
      <Background />

      <!-- 左上角：带菜单的悬浮按钮 -->
      <div
        class="flow-toolbar flow-toolbar--left"
        style="
          position: absolute;
          left: 12px;
          top: 12px;
          z-index: 4;
          display: flex;
          align-items: center;
          gap: 8px;
        "
      >
        <n-dropdown
          trigger="click"
          :options="addNodeOptions"
          @select="handleToolbarAdd"
        >
          <n-button circle type="primary" title="添加节点">
            <template #icon>
              <n-icon><Add /></n-icon>
            </template>
          </n-button>
        </n-dropdown>
      </div>

      <!-- 右键菜单：固定位置 NMenu -->
      <Teleport to="body">
          <div
            v-if="contextMenuShow"
            class="flow-context-menu-backdrop"
            style="position: fixed; inset: 0; z-index: 9997;"
            @click="contextMenuShow = false"
            @contextmenu.prevent="contextMenuShow = false"
          />
          <div
            v-if="contextMenuShow"
            class="flow-context-menu-panel"
            :class="{ 'flow-context-menu-panel--dark': dark }"
            :style="{
              position: 'fixed',
              left: contextMenuPosition.x + 'px',
              top: contextMenuPosition.y + 'px',
              zIndex: 9999,
              minWidth: '160px',
              maxHeight: '70vh',
              overflow: 'auto',
            }"
            @click.stop
          >
            <n-menu
              :options="contextMenuOptions"
              :value="null"
              @update:value="handleContextMenuSelect"
            />
          </div>
          <!-- 边的右键菜单：删除连线 -->
          <div
            v-if="edgeContextMenuShow"
            class="flow-context-menu-backdrop"
            style="position: fixed; inset: 0; z-index: 9997;"
            @click="closeEdgeContextMenu"
            @contextmenu.prevent="edgeContextMenuShow = false"
          />
          <div
            v-if="edgeContextMenuShow"
            class="flow-context-menu-panel"
            :class="{ 'flow-context-menu-panel--dark': dark }"
            :style="{
              position: 'fixed',
              left: edgeContextMenuPosition.x + 'px',
              top: edgeContextMenuPosition.y + 'px',
              zIndex: 9999,
              minWidth: '120px',
            }"
            @click.stop
          >
            <n-menu
              :options="edgeContextMenuOptions"
              :value="null"
              @update:value="handleEdgeContextMenuSelect"
            />
          </div>
      </Teleport>

      <div
        style="
          position: absolute;
          right: 10px;
          top: 10px;
          z-index: 4;
          display: flex;
          align-items: center;
          gap: 8px;
        "
      >
        <n-button
          circle
          title="整理布局"
          @click="layoutGraph('LR')"
        >
          <template #icon>
            <n-icon><GridOutline /></n-icon>
          </template>
        </n-button>
        <n-button
          circle
          type="primary"
          :loading="saveLoading"
          title="保存流程（无 ID 时先新建再保存）"
          @click="openSaveModal"
        >
          <template #icon>
            <n-icon><SaveOutline /></n-icon>
          </template>
        </n-button>
        <n-button
          circle
          :type="runningNodeId ? 'error' : 'primary'"
          :title="runningNodeId ? '停止测试运行' : '测试运行'"
          @click="startTestRun"
        >
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
        <n-card
          style="width: 400px"
          title="保存流程图"
          :bordered="false"
          role="dialog"
          aria-modal="true"
        >
          <n-form-item label="流程图名称">
            <n-input
              v-model:value="saveFormName"
              placeholder="请输入流程图名称"
              clearable
              @keyup.enter="doSaveConfirm"
            />
          </n-form-item>
          <n-form-item label="描述">
            <n-input
              v-model:value="saveFormDescription"
              type="textarea"
              placeholder="请输入描述（选填）"
              :autosize="{ minRows: 2, maxRows: 6 }"
              clearable
            />
          </n-form-item>
          <n-form-item label="Base URL">
            <n-input
              v-model:value="saveFormBaseUrl"
              placeholder="请求基础地址，与节点请求路径拼接（选填）"
              clearable
            />
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
        <n-card
          style="width: 600px"
          title="编辑节点"
          :bordered="false"
          size="huge"
          role="dialog"
          aria-modal="true"
        >
          <n-form
            ref="formRef"
            :model="editFormModel"
            label-placement="left"
            label-width="auto"
            require-mark-placement="right-hanging"
          >
            <!-- 公共项：节点名称 -->
            <n-form-item label="节点名称" path="label">
              <n-input v-model:value="editFormModel.label" placeholder="请输入节点显示内容" />
            </n-form-item>

            <!-- 普通场景：描述 -->
            <template v-if="editFormModel.nodeCategory === 'scene' && editFormModel.nodeKind === 'default'">
              <n-form-item label="描述" path="description">
                <n-input
                  v-model:value="editFormModel.description"
                  type="textarea"
                  placeholder="请输入描述"
                  :autosize="{ minRows: 2, maxRows: 6 }"
                />
              </n-form-item>
            </template>

            <!-- 请求配置（不包含 base_url） -->
            <n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">
              请求配置
            </n-divider>
            <n-form-item label="请求路径" path="requestPath">
              <n-input
                v-model:value="editFormModel.requestPath"
                placeholder="如 /api/scene/enter，不含 base_url"
                clearable
              />
            </n-form-item>
            <n-form-item label="请求方法" path="requestMethod">
              <n-input
                v-model:value="editFormModel.requestMethod"
                placeholder="如 GET、POST"
                clearable
              />
            </n-form-item>
            <n-form-item label="请求体/参数" path="requestData">
              <n-input
                v-model:value="editFormModel.requestData"
                type="textarea"
                placeholder="JSON 或其它请求参数（选填）"
                :autosize="{ minRows: 2, maxRows: 6 }"
                clearable
              />
            </n-form-item>

            <!-- 选择节点 -->
            <template v-if="editFormModel.nodeCategory === 'choice'">
              <n-form-item label="选项个数">
                <n-input-number
                  :value="editFormModel.outputCount"
                  @update:value="onOptionCountChange"
                  :min="1"
                  style="width: 100%"
                />
              </n-form-item>
              <n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">
                选项列表
              </n-divider>
              <div
                v-for="(opt, index) in editFormModel.options"
                :key="index"
                style="
                  border: 1px solid var(--n-border-color);
                  border-radius: 4px;
                  padding: 8px;
                  margin-bottom: 8px;
                "
              >
                <div style="font-size: 12px; margin-bottom: 4px; opacity: 0.8">
                  选项 {{ index + 1 }}
                </div>
                <n-form-item label="名称" label-placement="left" :show-feedback="false" style="margin-bottom: 8px">
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
                <n-input-number
                  :value="editFormModel.inputCount"
                  @update:value="onResultCountChange"
                  :min="1"
                  style="width: 100%"
                />
              </n-form-item>
              <n-divider title-placement="left" style="margin: 12px 0; font-size: 12px; color: #999">
                结果列表
              </n-divider>
              <div
                v-for="(res, index) in editFormModel.results"
                :key="index"
                style="
                  border: 1px solid var(--n-border-color);
                  border-radius: 4px;
                  padding: 8px;
                  margin-bottom: 8px;
                "
              >
                <div style="font-size: 12px; margin-bottom: 4px; opacity: 0.8">
                  结果 {{ index + 1 }}
                </div>
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

<style>
.vue-flow__node.running {
  border-width: 2px !important;
  border-color: rgb(59, 130, 246) !important;
  animation: running-breathe 1.5s ease-in-out infinite;
}
@keyframes running-breathe {
  0%,
  100% {
    box-shadow: 0 0 8px rgba(59, 130, 246, 0.4);
  }
  50% {
    box-shadow: 0 0 24px rgba(59, 130, 246, 0.85);
  }
}

/* 节点顶部工具栏：复制、编辑、删除 */
.flow-node-toolbar {
  display: flex;
  align-items: center;
  gap: 6px;
}

.flow-context-menu-panel {
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  padding: 4px 0;
  /* 浅色主题 */
  background: #fff;
  border: 1px solid #e0e0e0;
}
.flow-context-menu-panel--dark {
  background: #18181b;
  border-color: #3f3f46;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.5);
}
</style>
