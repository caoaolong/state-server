<script setup lang="ts">
import { ref, onMounted, watch, computed } from "vue";
import { useRoute } from "vue-router";
import { NDataTable, NCard, NButton, NInput, NSpace, type DataTableColumns, useMessage } from "naive-ui";
import { getSessionHistory, type SessionHistoryItem } from "../api";
import { formatDateTime } from "../utils/date";

const route = useRoute();
const message = useMessage();

const loading = ref(false);
const data = ref<SessionHistoryItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);
const sessionIdInput = ref("");

const sessionIdFromQuery = computed(() => (route.query.sessionId as string) ?? "");

const columns: DataTableColumns<SessionHistoryItem> = [
  { title: "ID", key: "id", width: 80 },
  { title: "会话 ID", key: "sessionId", width: 180, ellipsis: { tooltip: true } },
  { title: "事件", key: "event", width: 120 },
  { title: "原状态", key: "fromState", width: 120 },
  { title: "目标状态", key: "toState", width: 120 },
  { title: "时间", key: "createdAt", width: 180, render: (row) => formatDateTime(row.createdAt) },
];

async function fetchList() {
  loading.value = true;
  try {
    const sid = sessionIdInput.value.trim() || undefined;
    const res = await getSessionHistory({
      sessionId: sid,
      page: page.value,
      pageSize: pageSize.value,
    });
    data.value = res.list ?? [];
    total.value = res.total ?? 0;
  } catch (e) {
    message.error(e instanceof Error ? e.message : "加载历史失败");
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
  if (sessionIdFromQuery.value) {
    sessionIdInput.value = sessionIdFromQuery.value;
  }
  fetchList();
});

watch(sessionIdFromQuery, (v) => {
  if (v) sessionIdInput.value = v;
});
</script>

<template>
  <div class="page session-history">
    <n-card title="会话历史" :bordered="false">
      <template #header-extra>
        <n-space>
          <n-input
            v-model:value="sessionIdInput"
            placeholder="会话 ID 筛选"
            clearable
            style="width: 200px"
            @keyup.enter="onSearch"
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
.session-history {
  padding: 16px;
  height: auto;
  overflow: auto;
}
</style>
