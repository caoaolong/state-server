<script setup lang="ts">
/**
 * 流程编辑器：仅提供路由参数与主题，具体逻辑在 FlowCanvas 内部实现
 */
import { computed, inject, ref } from "vue";
import { useRoute } from "vue-router";
import FlowCanvas from "./FlowCanvas.vue";

const route = useRoute();
const dark = inject<ReturnType<typeof ref<boolean>>>("app-dark-mode", ref(false));

const stateMachineId = computed(() => route.params.id as string | undefined);
</script>

<template>
	<div style="height: 100%">
		<FlowCanvas :state-machine-id="stateMachineId" :dark="dark" />
	</div>
</template>

<style>
/* VueFlow 全局样式如需可保留 */
</style>
