<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import { useRouter } from "vue-router";
import { NDataTable, NCard, NButton, NSpace, type DataTableColumns } from "naive-ui";

const router = useRouter();

interface StateMachineRow {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

const loading = ref(false);
const data = ref<StateMachineRow[]>([]);

const columns: DataTableColumns<StateMachineRow> = [
  { title: "ID", key: "id", width: 120, ellipsis: { tooltip: true } },
  { title: "名称", key: "name", width: 160 },
  { title: "描述", key: "description", ellipsis: { tooltip: true } },
  { title: "创建时间", key: "createdAt", width: 180 },
  { title: "更新时间", key: "updatedAt", width: 180 },
  {
    title: "操作",
    key: "actions",
    width: 160,
    render(_row) {
      return h(NSpace, null, {
        default: () => [
          h(NButton, { quaternary: true, size: "small", type: "primary", onClick: () => router.push("/state-machines/design") }, () => "设计"),
          h(NButton, { quaternary: true, size: "small" }, () => "查看"),
        ],
      });
    },
  },
];

onMounted(() => {
  loading.value = true;
  // 模拟数据，后续可接真实 API
  setTimeout(() => {
    data.value = [
      { id: "sm-001", name: "示例状态机", description: "演示用", createdAt: "2025-02-01 10:00:00", updatedAt: "2025-02-09 12:00:00" },
    ];
    loading.value = false;
  }, 300);
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
