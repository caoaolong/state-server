<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import { useRouter } from "vue-router";
import { NDataTable, NCard, NButton, NSpace, type DataTableColumns, useMessage } from "naive-ui";
import { getStateMachineList, deleteStateMachine, type StateMachineListItem } from "../api";
import { formatDateTime } from "../utils/date";

const router = useRouter();
const message = useMessage();

const loading = ref(false);
const data = ref<StateMachineListItem[]>([]);

const columns: DataTableColumns<StateMachineListItem> = [
  { title: "ID", key: "id", width: 120, ellipsis: { tooltip: true } },
  { title: "名称", key: "name", width: 160 },
  { title: "描述", key: "description", ellipsis: { tooltip: true } },
  { title: "创建时间", key: "createdAt", width: 180, render: (row) => formatDateTime(row.createdAt) },
  { title: "更新时间", key: "updatedAt", width: 180, render: (row) => formatDateTime(row.updatedAt) },
  {
    title: "操作",
    key: "actions",
    width: 220,
    render(_row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { quaternary: true, size: "small", type: "primary", onClick: () => router.push(`/state-machines/design/${_row.id}`) }, () => "设计"),
          h(NButton, { quaternary: true, size: "small" }, () => "查看"),
          h(NButton, { quaternary: true, size: "small", type: "error", loading: deletingId.value === _row.id, onClick: () => handleDelete(_row) }, () => "删除"),
        ],
      });
    },
  },
];

async function fetchList() {
  loading.value = true;
  try {
    const res = await getStateMachineList();
    data.value = res.list ?? [];
  } catch (e) {
    message.error(e instanceof Error ? e.message : "加载列表失败");
    data.value = [];
  } finally {
    loading.value = false;
  }
}

const deletingId = ref<string | null>(null);
async function handleDelete(row: StateMachineListItem) {
  if (!window.confirm(`确定要删除状态机「${row.name}」吗？`)) return;
  deletingId.value = row.id;
  try {
    await deleteStateMachine(row.id);
    message.success("删除成功");
    data.value = data.value.filter((r) => r.id !== row.id);
  } catch (e) {
    message.error(e instanceof Error ? e.message : "删除失败");
  } finally {
    deletingId.value = null;
  }
}

onMounted(() => {
  fetchList();
});
</script>

<template>
  <div class="page state-machine-list">
    <n-card title="状态机列表" :bordered="false">
      <n-data-table :columns="columns" :data="data" :loading="loading" :bordered="false" striped />
    </n-card>
  </div>
</template>

<style scoped>
.state-machine-list {
  padding: 16px;
  height: auto;
  overflow: auto;
}
</style>
