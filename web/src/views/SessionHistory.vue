<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import { NDataTable, NCard, NButton, type DataTableColumns } from "naive-ui";
import { formatDateTime } from "../utils/date";

interface HistoryRow {
  id: string;
  sessionId: string;
  event: string;
  fromState: string;
  toState: string;
  createdAt: string;
}

const loading = ref(false);
const data = ref<HistoryRow[]>([]);

const columns: DataTableColumns<HistoryRow> = [
  { title: "ID", key: "id", width: 80 },
  { title: "会话 ID", key: "sessionId", width: 180, ellipsis: { tooltip: true } },
  { title: "事件", key: "event", width: 120 },
  { title: "原状态", key: "fromState", width: 120 },
  { title: "目标状态", key: "toState", width: 120 },
  { title: "时间", key: "createdAt", width: 180, render: (row) => formatDateTime(row.createdAt) },
  {
    title: "操作",
    key: "actions",
    width: 100,
    render() {
      return h(NButton, { size: "small" }, () => "查看");
    },
  },
];

onMounted(() => {
  loading.value = true;
  setTimeout(() => {
    data.value = [
      { id: "1", sessionId: "sess-001", event: "submit", fromState: "draft", toState: "pending", createdAt: "2025-02-09 14:05:00" },
      { id: "2", sessionId: "sess-001", event: "approve", fromState: "pending", toState: "approved", createdAt: "2025-02-09 14:10:00" },
    ];
    loading.value = false;
  }, 300);
});
</script>

<template>
  <div class="page session-history">
    <n-card title="会话历史" :bordered="false">
      <n-data-table :columns="columns" :data="data" :loading="loading" :bordered="false" striped />
    </n-card>
  </div>
</template>

<style scoped>
.session-history {
  padding: 16px;
  height: auto;
  overflow: auto;
}
</style>
