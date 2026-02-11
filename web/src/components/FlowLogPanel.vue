<script setup lang="ts">
/**
 * 流程日志面板：展示当前选中节点的请求体与响应结果
 */
import type { Node } from "@vue-flow/core";
import { NCode, NEmpty, NIcon, NTag } from "naive-ui";
import { DocumentTextOutline } from "@vicons/ionicons5";
import type { FlowNodeData } from "./FlowNode.vue";

defineProps<{
	/** 当前选中的节点，用于展示请求/响应日志 */
	selectionNode?: Node<FlowNodeData> | null;
}>();
</script>

<template>
	<div class="flow-log-panel">
		<div class="flow-log-panel__header">
			<n-icon :component="DocumentTextOutline" :size="18" class="flow-log-panel__icon" />
			<span class="flow-log-panel__title">节点日志</span>
		</div>

		<template v-if="selectionNode">
			<div class="flow-log-panel__node-header">
				<span class="flow-log-panel__node-label">{{ selectionNode.data?.label ?? "未命名节点" }}</span>
				<n-tag v-if="selectionNode.data?.nodeCategory" size="small" :bordered="false" type="info">
					{{ selectionNode.data.nodeCategory === "scene" ? "场景" : selectionNode.data.nodeCategory === "choice" ? "选择" : "结果" }}
				</n-tag>
			</div>

			<template v-if="selectionNode.data?.requestData || selectionNode.data?.runResult">
				<div class="flow-log-panel__row">
					<div class="flow-log-panel__section">
						<div class="flow-log-panel__section-title">请求体</div>
						<div v-if="selectionNode.data.requestData" class="flow-log-panel__code-wrap">
							<n-code
								:code="selectionNode.data.requestData"
								language="json"
								word-wrap
								class="flow-log-panel__code"
							/>
						</div>
						<div v-else class="flow-log-panel__empty-line">无</div>
					</div>
					<div class="flow-log-panel__section">
						<div class="flow-log-panel__section-title">响应结果</div>
						<div v-if="selectionNode.data.runResult" class="flow-log-panel__code-wrap">
							<n-code
								:code="selectionNode.data.runResult"
								language="json"
								word-wrap
								class="flow-log-panel__code"
							/>
						</div>
						<div v-else class="flow-log-panel__empty-line">无</div>
					</div>
				</div>
			</template>

			<n-empty
				v-else
				description="该节点暂无请求或运行数据"
				size="medium"
				class="flow-log-panel__empty"
			/>
		</template>

		<n-empty
			v-else
			description="在画布中点击节点以查看日志"
			size="large"
			class="flow-log-panel__empty flow-log-panel__empty--no-node"
		/>
	</div>
</template>

<style scoped>
.flow-log-panel {
	height: 100%;
	display: flex;
	flex-direction: column;
	min-height: 0;
	background: var(--n-color);
}

.flow-log-panel__header {
	display: flex;
	align-items: center;
	gap: 8px;
	padding: 12px 16px;
	border-bottom: 1px solid var(--n-border-color);
	flex-shrink: 0;
}

.flow-log-panel__icon {
	opacity: 0.85;
}

.flow-log-panel__title {
	font-weight: 600;
	font-size: 14px;
	color: var(--n-text-color);
}

.flow-log-panel__node-header {
	display: flex;
	align-items: center;
	gap: 8px;
	padding: 12px 16px;
	flex-shrink: 0;
	border-bottom: 1px solid var(--n-border-color);
}

.flow-log-panel__node-label {
	font-weight: 600;
	font-size: 13px;
	color: var(--n-text-color);
}

.flow-log-panel__row {
	flex: 1;
	min-height: 0;
	display: flex;
	flex-direction: row;
	gap: 12px;
	padding: 12px 16px;
}

.flow-log-panel__section {
	flex: 1;
	min-width: 0;
	min-height: 0;
	display: flex;
	flex-direction: column;
}

.flow-log-panel__section-title {
	font-size: 12px;
	color: var(--n-text-color-3);
	margin-bottom: 8px;
	font-weight: 500;
}

.flow-log-panel__code-wrap {
	flex: 1;
	min-height: 80px;
	max-height: 240px;
	overflow: auto;
	border-radius: 6px;
	background: var(--n-color-modal);
	border: 1px solid var(--n-border-color);
	padding: 10px 12px;
}

.flow-log-panel__code {
	font-size: 12px;
	line-height: 1.5;
}

.flow-log-panel__code-wrap :deep(pre) {
	margin: 0;
	white-space: pre-wrap;
	word-break: break-all;
}

.flow-log-panel__empty-line {
	font-size: 12px;
	color: var(--n-text-color-3);
	padding: 8px 0;
}

.flow-log-panel__empty {
	flex: 1;
	display: flex;
	flex-direction: column;
	justify-content: center;
	align-items: center;
	padding: 24px;
	min-height: 160px;
}

.flow-log-panel__empty--no-node {
	min-height: 200px;
}
</style>
