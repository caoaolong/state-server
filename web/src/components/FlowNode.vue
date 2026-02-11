<script setup lang="ts">
/**
 * 流程节点卡片：场景/选择/结果，支持复制/编辑/删除、执行请求、多 handle
 */
import { watch, computed, ref } from "vue";
import { Handle, Position, type NodeProps } from "@vue-flow/core";
import { NodeToolbar } from "@vue-flow/node-toolbar";
import { NCard, NIcon, NButton, NSpace, useMessage } from "naive-ui";
import {
  PlayCircle,
  PlayCircleOutline,
  PauseCircle,
  CheckmarkCircle,
  CloseCircle,
  EllipseOutline,
  CopyOutline,
  CreateOutline,
  TrashOutline,
} from "@vicons/ionicons5";

/** 节点状态（用于 Card 头部展示） */
export type NodeState = "normal" | "running" | "paused" | "completed" | "failed";

/** 节点大类：场景、选择、结果 */
export type NodeCategory = "scene" | "choice" | "result";

/** 场景子类型：开始、结束、普通 */
export type SceneKind = "start" | "end" | "default";

/** 通用流程节点 data 类型 */
export interface FlowNodeData {
  label: string;
  /** 节点大类，用于样式与 handle 数量 */
  nodeCategory: NodeCategory;
  /** 场景子类型：开始仅输出、结束仅输入、普通双端 */
  nodeKind?: SceneKind;
  /** 选择节点：输出端数量（多个输出） */
  outputCount?: number;
  /** 结果节点：输入端数量（多个输入） */
  inputCount?: number;
  nodeType?: string;
  option?: string;
  description?: string;
  options?: Array<{ label: string; description: string }>;
  results?: Array<{ label: string; description: string }>;
  /** 请求地址（不包含 base_url） */
  requestPath?: string;
  /** 请求方法，如 GET/POST */
  requestMethod?: string;
  /** 请求体/参数（如 JSON 字符串） */
  requestData?: string;
  isRunning?: boolean;
  nodeState?: Exclude<NodeState, "normal"> | null;
  onRun?: () => void;
  /** 运行结果 */
  runResult?: string;
}

const STATE_CONFIG: Record<
  NodeState,
  { label: string; icon: typeof PlayCircle; class: string }
> = {
  normal: { label: "等待中", icon: EllipseOutline, class: "flow-node__state--normal" },
  running: { label: "运行中", icon: PlayCircle, class: "flow-node__state--running" },
  paused: { label: "已暂停", icon: PauseCircle, class: "flow-node__state--paused" },
  completed: { label: "已完成", icon: CheckmarkCircle, class: "flow-node__state--completed" },
  failed: { label: "失败", icon: CloseCircle, class: "flow-node__state--failed" },
};

const props = withDefaults(
  defineProps<NodeProps<FlowNodeData> & { baseUrl?: string }>(),
  { baseUrl: "" }
);

const message = useMessage();
const isRunning = ref(false);

const stateConfig = computed(() =>
  props.data?.nodeState ? STATE_CONFIG[props.data.nodeState] : STATE_CONFIG.normal
);

watch(
  () => props.data?.isRunning,
  (running) => {
    if (running && typeof props.data?.onRun === "function") {
      props.data.onRun();
    }
  }
);

const category = computed(() => props.data?.nodeCategory ?? "scene");
const sceneKind = computed(() => props.data?.nodeKind ?? "default");

// 场景：开始无输入、结束无输出、普通一入一出
const isScene = computed(() => category.value === "scene");
const isStartOrEnd = computed(() => isScene.value && (sceneKind.value === "start" || sceneKind.value === "end"));
const isNormalScene = computed(() => isScene.value && sceneKind.value === "default");
const isFormNode = computed(
  () =>
    isNormalScene.value ||
    isChoice.value ||
    isResult.value
);
const showSceneTarget = computed(
  () => isScene.value && (sceneKind.value === "end" || sceneKind.value === "default")
);
const showSceneSource = computed(
  () => isScene.value && (sceneKind.value === "start" || sceneKind.value === "default")
);

// 多连接桩：从顶部起算，间隔 30px
const HANDLE_TOP_OFFSET_PX = 50;
const HANDLE_GAP_PX = 31;

// 选择：一个输入，多个输出
const isChoice = computed(() => category.value === "choice");
const choiceOutputCount = computed(() => Math.max(1, props.data?.outputCount ?? 2));
const choiceOutputHandles = computed(() =>
  Array.from({ length: choiceOutputCount.value }, (_, i) => ({
    id: `source-${i}`,
    topPx: HANDLE_TOP_OFFSET_PX + i * HANDLE_GAP_PX,
  }))
);

// 结果：多个输入，一个输出
const isResult = computed(() => category.value === "result");
const resultInputCount = computed(() => Math.max(1, props.data?.inputCount ?? 2));
const resultInputHandles = computed(() =>
  Array.from({ length: resultInputCount.value }, (_, i) => ({
    id: `target-${i}`,
    topPx: HANDLE_TOP_OFFSET_PX + i * HANDLE_GAP_PX,
  }))
);

const cardContentStyle = { padding: "8px 12px" };

const emit = defineEmits<{
  copy: [nodeId: string];
  edit: [nodeId: string];
  delete: [nodeId: string];
  "update:run-result": [payload: { nodeId: string; runResult: string; nodeState: string }];
}>();

function onCopy() {
  emit("copy", props.id);
}
function onEdit() {
  emit("edit", props.id);
}
function onDelete() {
  emit("delete", props.id);
}

function onRun() {
  props.data?.onRun?.();
}

async function executeRequest() {
  const data = props.data;
  if (!data) return;

  const setErrorResult = (msg: string) => {
    data.runResult = `错误: ${msg}`;
    data.nodeState = "failed";
    emit("update:run-result", { nodeId: props.id, runResult: data.runResult, nodeState: "failed" });
  };

  if (!data.requestPath) {
    setErrorResult("请配置请求路径");
    message.error("请配置请求路径");
    return;
  }

  isRunning.value = true;
  data.nodeState = "running";

  try {
    const BASE = import.meta.env.VITE_API_BASE ?? "";
    const res = await fetch(`${BASE}/nodes/run`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        node: {
          id: props.id,
          type: props.type ?? "default",
          position: props.position,
          data: {
            requestPath: data.requestPath ?? "",
            requestMethod: data.requestMethod ?? "GET",
            requestData: data.requestData ?? "",
          },
        },
        sessionId: 0,
      }),
    });
    const text = await res.text();
    let result: { ok?: boolean; statusCode?: number; body?: string; error?: string };
    try {
      result = text ? JSON.parse(text) : {};
    } catch {
      result = { ok: false, statusCode: res.status, body: text, error: res.statusText || `HTTP ${res.status}` };
    }
    if (!res.ok && result.error == null) {
      result.error = text || `HTTP ${res.status}`;
    }

    if (result.ok !== false && result.body !== undefined) {
      try {
        data.runResult = JSON.stringify(JSON.parse(result.body), null, 2);
      } catch {
        data.runResult = result.body;
      }
      data.nodeState = "completed";
      message.success(`请求成功 (${result.statusCode ?? res.status})`);
    } else {
      const errMsg = (result.error ?? result.body ?? text) || "请求失败";
      data.runResult = `错误: ${errMsg}${result.body ? `\n\n响应: ${result.body}` : ""}`;
      data.nodeState = "failed";
      message.error(result.error ?? "请求失败");
    }
  } catch (error) {
    const errMsg = error instanceof Error ? error.message : String(error);
    data.runResult = `错误: ${errMsg}`;
    data.nodeState = "failed";
    message.error("请求失败");
  } finally {
    isRunning.value = false;
    emit("update:run-result", {
      nodeId: props.id,
      runResult: data.runResult ?? "",
      nodeState: data.nodeState ?? "normal",
    });
  }
}
</script>

<template>
  <div
    class="flow-node"
    :class="[
      `flow-node--${category}`,
      isScene ? `flow-node--scene-${sceneKind}` : '',
    ]"
    :data-node-category="category"
    :data-node-kind="isScene ? sceneKind : undefined"
  >
    <NodeToolbar :position="Position.Top" :offset="8" align="center" class="flow-node-toolbar">
      <n-button v-if="!isStartOrEnd" quaternary size="small" title="复制" @click.stop="onCopy">
        <template #icon>
          <n-icon :component="CopyOutline" />
        </template>
        复制
      </n-button>
      <n-button quaternary size="small" title="编辑" @click.stop="onEdit">
        <template #icon>
          <n-icon :component="CreateOutline" />
        </template>
        编辑
      </n-button>
      <n-button quaternary size="small" type="error" title="删除" @click.stop="onDelete">
        <template #icon>
          <n-icon :component="TrashOutline" />
        </template>
        删除
      </n-button>
    </NodeToolbar>
    <!-- 场景：左侧一个 target -->
    <Handle
      v-if="showSceneTarget"
      type="target"
      :position="Position.Left"
      class="flow-node__handle"
    />
    <!-- 结果：左侧多个 target，从顶部起算、间隔 30px -->
    <Handle
      v-for="h in resultInputHandles"
      v-else-if="isResult"
      :key="h.id"
      :id="h.id"
      type="target"
      :position="Position.Left"
      :style="{ top: h.topPx + 'px', left: 0, transform: 'translate(-50%, -50%)' }"
      class="flow-node__handle flow-node__handle--multi"
    />
    <!-- 选择：左侧一个 target -->
    <Handle
      v-else-if="isChoice"
      type="target"
      :position="Position.Left"
      class="flow-node__handle"
    />

    <n-card
      size="small"
      :bordered="false"
      :segmented="{ content: true }"
      :hoverable="false"
      class="flow-node__card"
      :content-style="cardContentStyle"
    >
      <template #header>
        <div class="flow-node__header" :class="stateConfig.class">
          <n-icon :component="stateConfig.icon" class="flow-node__header-icon" />
          <span class="flow-node__header-label">{{ stateConfig.label }}</span>
        </div>
      </template>
      <div class="flow-node__content">
        <template v-if="isFormNode">
          <!-- 普通场景：描述 -->
          <template v-if="isNormalScene">
            <div class="flow-node__text-content" v-if="data.description">
              {{ data.description }}
            </div>
            <div class="flow-node__text-empty" v-else>
              暂无描述
            </div>
          </template>

          <!-- 选择节点：显示选项列表摘要 -->
          <template v-else-if="isChoice">
            <div class="flow-node__list-content">
              <div v-for="(opt, index) in data.options" :key="index" class="flow-node__list-item">
                <span class="flow-node__list-index">{{ index + 1 }}.</span>
                <span class="flow-node__list-label">{{ opt.label || '未命名' }}</span>
              </div>
              <div v-if="!data.options?.length" class="flow-node__text-empty">
                暂无选项
              </div>
            </div>
          </template>

          <!-- 结果节点：显示结果列表摘要 -->
          <template v-else-if="isResult">
            <div class="flow-node__list-content">
              <div v-for="(res, index) in data.results" :key="index" class="flow-node__list-item">
                <span class="flow-node__list-index">{{ index + 1 }}.</span>
                <span class="flow-node__list-label">{{ res.label || '未命名' }}</span>
              </div>
              <div v-if="!data.results?.length" class="flow-node__text-empty">
                暂无结果
              </div>
            </div>
          </template>
        </template>
        <template v-else>
          <span class="flow-node__label">{{ data?.label ?? id }}</span>
          <span v-if="data?.option" class="flow-node__option">{{ data.option }}</span>
        </template>
      </div>
      <template #footer>
        <div class="flow-node__footer" :class="{ 'flow-node__footer--between': isFormNode }">
          <span v-if="isFormNode" class="flow-node__footer-label" :title="data?.label">{{ data?.label }}</span>
          <n-space :size="4">
            <n-button
              v-if="data?.onRun"
              quaternary
              circle
              size="small"
              title="运行"
              type="primary"
              @click.stop="onRun"
            >
              <template #icon>
                <n-icon :component="PlayCircle" />
              </template>
            </n-button>
            <!-- 自定义节点的运行按钮 -->
            <n-button
              v-if="(isScene || isFormNode) && data?.requestPath && data?.requestMethod"
              quaternary
              circle
              size="small"
              title="执行请求"
              :loading="isRunning"
              @click.stop="executeRequest"
            >
              <template #icon>
                <n-icon :component="PlayCircleOutline" />
              </template>
            </n-button>
          </n-space>
        </div>
      </template>
    </n-card>

    <!-- 场景：右侧一个 source -->
    <Handle
      v-if="showSceneSource"
      type="source"
      :position="Position.Right"
      class="flow-node__handle"
    />
    <!-- 选择：右侧多个 source，从顶部起算、间隔 30px -->
    <Handle
      v-for="h in choiceOutputHandles"
      v-else-if="isChoice"
      :key="h.id"
      :id="h.id"
      type="source"
      :position="Position.Right"
      :style="{ top: h.topPx + 'px', right: 0, left: 'auto', transform: 'translate(50%, -50%)' }"
      class="flow-node__handle flow-node__handle--multi"
    />
    <!-- 结果：右侧一个 source -->
    <Handle
      v-else-if="isResult"
      type="source"
      :position="Position.Right"
      class="flow-node__handle"
    />
  </div>
</template>

<style scoped>
.flow-node {
  position: relative;
  display: flex;
  align-items: center;
  gap: 0;
  width: 200px;
  min-width: 200px;
  max-width: 400px;
  color: var(--vf-node-text, #f1f1f1);
  font-size: 12px;
  box-sizing: border-box;
  cursor: pointer;
  pointer-events: auto;
}

.flow-node :deep(*) {
  pointer-events: auto;
}

.flow-node__card {
  flex: 1;
  min-width: 0;
  background: transparent;
}

.flow-node__card :deep(.n-card__content) {
  padding: 8px 12px;
  background: transparent;
}

.flow-node__card :deep(.n-card-header) {
  padding: 4px 10px;
  background: transparent;
}

.flow-node__card :deep(.n-card__footer) {
  padding: 2px 10px 4px;
  background: transparent;
}

.flow-node__footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 4px;
}

.flow-node__footer--between {
  justify-content: space-between;
}

.flow-node__footer-label {
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
  margin-right: 8px;
  color: var(--vf-node-text, #f1f1f1);
}

.flow-node__header {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 500;
}

.flow-node__header-icon {
  font-size: 14px;
}

.flow-node__header-label {
  color: inherit;
}

.flow-node__header.flow-node__state--normal {
  color: #999;
}
.flow-node__header.flow-node__state--normal .flow-node__header-icon {
  color: #999;
}
.flow-node__header.flow-node__state--running {
  color: #18a058;
}
.flow-node__header.flow-node__state--running .flow-node__header-icon {
  color: #18a058;
}
.flow-node__header.flow-node__state--paused {
  color: #f0a020;
}
.flow-node__header.flow-node__state--paused .flow-node__header-icon {
  color: #f0a020;
}
.flow-node__header.flow-node__state--completed {
  color: #18a058;
}
.flow-node__header.flow-node__state--completed .flow-node__header-icon {
  color: #18a058;
}
.flow-node__header.flow-node__state--failed {
  color: #d03050;
}
.flow-node__header.flow-node__state--failed .flow-node__header-icon {
  color: #d03050;
}

.flow-node__content {
  min-height: 20px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.flow-node__label {
  font-weight: 500;
}

.flow-node__option {
  font-size: 11px;
  opacity: 0.85;
}

.flow-node__handle {
  width: 8px;
  height: 8px;
  background: var(--vf-node-color, #999);
  border: 2px solid var(--vf-node-bg, #1a192b);
  flex-shrink: 0;
}

/* 多 handle：从顶部起算、间隔 30px，top/left/right 由内联 style 设置 */
.flow-node__handle--multi {
  position: absolute;
}

/* ========== 三种节点样式区分 ========== */
/* 场景：绿色系 */
.flow-node--scene-start {
  --flow-node-accent: #059669;
}
.flow-node--scene-end {
  --flow-node-accent: #dc2626;
}
.flow-node--scene-default {
  --flow-node-accent: #0891b2;
}
.flow-node--scene-start .flow-node__card :deep(.n-card-header),
.flow-node--scene-start .flow-node__handle { border-color: #059669; color: #059669; }
.flow-node--scene-end .flow-node__card :deep(.n-card-header),
.flow-node--scene-end .flow-node__handle { border-color: #dc2626; color: #dc2626; }
.flow-node--scene-default .flow-node__card :deep(.n-card-header),
.flow-node--scene-default .flow-node__handle { border-color: #0891b2; color: #0891b2; }

/* 选择：橙色系，多输出 */
.flow-node--choice .flow-node__card :deep(.n-card-header),
.flow-node--choice .flow-node__handle {
  border-color: #d97706;
  color: #d97706;
}

/* 结果：紫色系，多输入 */
.flow-node--result .flow-node__card :deep(.n-card-header),
.flow-node--result .flow-node__handle {
  border-color: #7c3aed;
  color: #7c3aed;
}

.flow-node__text-content {
  font-size: 12px;
  line-height: 1.5;
  color: var(--vf-node-text, #f1f1f1);
  opacity: 0.9;
  word-break: break-all;
}

.flow-node__text-empty {
  font-size: 12px;
  color: var(--vf-node-text, #f1f1f1);
  opacity: 0.5;
  font-style: italic;
}

.flow-node__list-content {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 0;
}

.flow-node__list-item {
  height: 30px;
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  border-bottom: 1px dashed var(--vf-node-color, #999);
}

.flow-node__list-index {
  opacity: 0.6;
  min-width: 16px;
}

.flow-node__list-label {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

</style>
