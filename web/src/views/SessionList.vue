<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import { NDataTable, NCard, NButton, NTag, type DataTableColumns } from "naive-ui";

interface SessionRow {
  id: string;
  sessionId: string;
  stateMachineId: string;
  status: string;
  createdAt: string;
}

const loading = ref(false);
const data = ref<SessionRow[]>([]);

const columns: DataTableColumns<SessionRow> = [
  { title: "ID", key: "id", width: 80 },
  { title: "会话 ID", key: "sessionId", width: 200, ellipsis: { tooltip: true } },
  { title: "状态机", key: "stateMachineId", width: 140 },
  {
    title: "状态",
    key: "status",
    width: 100,
    render(row) {
      const type = row.status === "running" ? "success" : row.status === "ended" ? "default" : "warning";
      return h(NTag, { type, size: "small" }, () => row.status);
    },
  },
  { title: "创建时间", key: "createdAt", width: 180 },
  {
    title: "操作",
    key: "actions",
    width: 120,
    render() {
      return h(NButton, { size: "small" }, () => "详情");
    },
  },
];

onMounted(() => {
  loading.value = true;
  setTimeout(() => {
    data.value = [
      { id: "1", sessionId: "sess-abc-001", stateMachineId: "sm-001", status: "running", createdAt: "2025-02-09 14:00:00" },
      { id: "2", sessionId: "sess-abc-002", stateMachineId: "sm-001", status: "ended", createdAt: "2025-02-09 13:30:00" },
    ];
    loading.value = false;
  }, 300);
});
</script>

<template>
  <div class="page session-list">
    <n-card title="会话列表" :bordered="false">
      <n-data-table :columns="columns" :data="data" :loading="loading" :bordered="false" striped />
    </n-card>
  </div>
</template>

<style scoped>
.session-list {
  padding: 16px;
  height: auto;
  overflow: auto;
}
</style>
