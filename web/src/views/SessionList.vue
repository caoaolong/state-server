<script setup lang="ts">
import { ref, onMounted, h } from "vue";
import { useRouter } from "vue-router";
import { NDataTable, NCard, NButton, NTag, NSpace, NSelect, type DataTableColumns, useMessage } from "naive-ui";
import { getSessionList, getStateMachineList, type SessionListItem, type StateMachineListItem } from "../api";
import { formatDateTime } from "../utils/date";

const router = useRouter();
const message = useMessage();

const loading = ref(false);
const data = ref<SessionListItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const stateMachineId = ref<string | null>(null);
const status = ref<string | null>(null);
const stateMachineOptions = ref<{ label: string; value: string }[]>([]);

const statusOptions = [
  { label: "全部", value: "" },
  { label: "运行中", value: "running" },
  { label: "已结束", value: "ended" },
  { label: "已挂起", value: "suspended" },
];

const columns: DataTableColumns<SessionListItem> = [
  { title: "ID", key: "id", width: 80 },
  { title: "会话 ID", key: "sessionId", width: 200, ellipsis: { tooltip: true } },
  { title: "状态机 ID", key: "stateMachineId", width: 140 },
  {
    title: "状态",
    key: "status",
    width: 100,
    render(row) {
      const type = row.status === "running" ? "success" : row.status === "ended" ? "default" : "warning";
      return h(NTag, { type, size: "small" }, () => row.status);
    },
  },
  { title: "创建时间", key: "createdAt", width: 180, render: (row) => formatDateTime(row.createdAt) },
  {
    title: "操作",
    key: "actions",
    width: 120,
    render(row) {
      return h(
        NButton,
        { quaternary: true, size: "small", onClick: () => router.push({ path: "/sessions/history", query: { sessionId: row.sessionId } }) },
        () => "历史"
      );
    },
  },
];

async function fetchStateMachineOptions() {
  try {
    const res = await getStateMachineList({ pageSize: 200 });
    stateMachineOptions.value = [
      { label: "全部", value: "" },
      ...(res.list ?? []).map((r: StateMachineListItem) => ({ label: r.name || r.id, value: r.id })),
    ];
  } catch {
    stateMachineOptions.value = [{ label: "全部", value: "" }];
  }
}

async function fetchList() {
  loading.value = true;
  try {
    const res = await getSessionList({
      page: page.value,
      pageSize: pageSize.value,
      stateMachineId: stateMachineId.value || undefined,
      status: (status.value as SessionListItem["status"]) || undefined,
    });
    data.value = res.list ?? [];
    total.value = res.total ?? 0;
  } catch (e) {
    message.error(e instanceof Error ? e.message : "加载列表失败");
    data.value = [];
    total.value = 0;
  } finally {
    loading.value = false;
  }
}

function onPageChange(p: number) {
  page.value = p;
  fetchList();
}

function onPageSizeChange(ps: number) {
  pageSize.value = ps;
  page.value = 1;
  fetchList();
}

function onSearch() {
  page.value = 1;
  fetchList();
}

onMounted(() => {
  fetchStateMachineOptions();
  fetchList();
});
</script>

<template>
  <div class="page session-list">
    <n-card title="会话列表" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-select
            v-model:value="stateMachineId"
            :options="stateMachineOptions"
            placeholder="状态机"
            clearable
            style="width: 160px"
          />
          <n-select
            v-model:value="status"
            :options="statusOptions"
            placeholder="状态"
            clearable
            style="width: 120px"
          />
          <n-button type="primary" @click="onSearch">查询</n-button>
        </n-space>
      </template>
      <n-data-table
        :columns="columns"
        :data="data"
        :loading="loading"
        :bordered="false"
        striped
        :pagination="{
          page,
          pageSize,
          itemCount: total,
          showSizePicker: true,
          pageSizes: [10, 20, 50],
          onUpdatePage: onPageChange,
          onUpdatePageSize: onPageSizeChange,
        }"
      />
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
