<script setup lang="ts">
import { ref, reactive, onMounted, inject, h } from "vue";
import type { Ref } from "vue";
import {
  NCard,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NButton,
  NIcon,
  NSpace,
  NSelect,
  NSwitch,
  NDataTable,
  NModal,
  useMessage,
  type DataTableColumns,
} from "naive-ui";
import { CopyOutline, RefreshOutline, TrashOutline } from "@vicons/ionicons5";
import {
  getApiKeyList,
  createApiKey as createApiKeyApi,
  deleteApiKey as deleteApiKeyApi,
  refreshApiKey as refreshApiKeyApi,
  revealApiKey,
  type ApiKeyItem,
} from "../api/apikey";
import { formatDateTime } from "../utils/date";
import type { ThemeOption } from "../api/settings";

const themeOption = inject<Ref<ThemeOption>>("app-theme-option")!;
const themeStorageKey = inject<string>("app-theme-storage-key", "app-theme");

const message = useMessage();
const loading = ref(false);
const apiKeyLoading = ref(false);

const form = reactive({
  appName: "状态机平台",
  apiBaseUrl: "http://localhost:3000",
  theme: "auto",
  language: "zh-CN",
  enableNotifications: true,
  maxHistoryItems: 100,
});

const apiKeyList = ref<ApiKeyItem[]>([]);
const showCreateModal = ref(false);
const createName = ref("");
const creating = ref(false);
const refreshingId = ref<string | null>(null);
const deletingId = ref<string | null>(null);
const revealingId = ref<string | null>(null);

async function fetchApiKeyList() {
  apiKeyLoading.value = true;
  try {
    apiKeyList.value = await getApiKeyList();
  } catch (e) {
    message.error(e instanceof Error ? e.message : "加载 API Key 列表失败");
    apiKeyList.value = [];
  } finally {
    apiKeyLoading.value = false;
  }
}

onMounted(() => {
  form.theme = themeOption.value;
  fetchApiKeyList();
});

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).then(
    () => message.success("已复制到剪贴板"),
    () => message.error("复制失败")
  );
}

function handleCreate() {
  showCreateModal.value = true;
  createName.value = "";
}

async function confirmCreate(): Promise<boolean> {
  const name = createName.value.trim() || "未命名";
  creating.value = true;
  try {
    const item = await createApiKeyApi(name);
    await fetchApiKeyList();
    message.success("创建成功，请妥善保存 Key（仅显示一次）");
    copyToClipboard(item.apiKey);
    return true;
  } catch (e) {
    message.error(e instanceof Error ? e.message : "创建失败");
    return false;
  } finally {
    creating.value = false;
  }
}

async function handleRefresh(row: ApiKeyItem) {
  refreshingId.value = row.id;
  try {
    const res = await refreshApiKeyApi(row.id);
    await fetchApiKeyList();
    message.success("已重新生成 Key，请更新使用处");
    copyToClipboard(res.apiKey);
  } catch (e) {
    message.error(e instanceof Error ? e.message : "刷新失败");
  } finally {
    refreshingId.value = null;
  }
}

async function handleDelete(row: ApiKeyItem) {
  if (!window.confirm(`确定要删除「${row.name}」的 API Key 吗？`)) return;
  deletingId.value = row.id;
  try {
    await deleteApiKeyApi(row.id);
    await fetchApiKeyList();
    message.success("已删除");
  } catch (e) {
    message.error(e instanceof Error ? e.message : "删除失败");
  } finally {
    deletingId.value = null;
  }
}

async function handleCopy(row: ApiKeyItem) {
  if (revealingId.value) return;
  revealingId.value = row.id;
  try {
    const res = await revealApiKey(row.id);
    copyToClipboard(res.apiKey);
  } catch (e) {
    message.error(e instanceof Error ? e.message : "获取 Key 失败");
  } finally {
    revealingId.value = null;
  }
}

const apiKeyColumns: DataTableColumns<ApiKeyItem> = [
  { title: "序号", key: "index", width: 70, render: (_row, index) => (index ?? 0) + 1 },
  { title: "名称", key: "name", width: 140, ellipsis: { tooltip: true } },
  {
    title: "ApiKey",
    key: "apiKey",
    width: 280,
    ellipsis: { tooltip: true },
    render(row) {
      const displayText = row.apiKey
        ? (row.apiKey.length > 12 ? row.apiKey.slice(0, 8) + "••••••••" : row.apiKey)
        : "已脱敏，点击复制获取";
      return h(NSpace, { align: "center", size: 8 }, {
        default: () => [
          h("span", { class: "api-key-mask" }, displayText),
          h(NButton, {
            quaternary: true,
            circle: true,
            size: "small",
            title: "复制",
            loading: revealingId.value === row.id,
            onClick: () => handleCopy(row),
          }, { icon: () => h(NIcon, { component: CopyOutline }) }),
        ],
      });
    },
  },
  { title: "创建时间", key: "createdAt", width: 180, render: (row) => formatDateTime(row.createdAt) },
  {
    title: "操作",
    key: "actions",
    width: 120,
    render(row) {
      return h(NSpace, { size: 4 }, {
        default: () => [
          h(NButton, {
            quaternary: true,
            circle: true,
            size: "small",
            title: "刷新",
            loading: refreshingId.value === row.id,
            onClick: () => handleRefresh(row),
          }, { icon: () => h(NIcon, { component: RefreshOutline }) }),
          h(NButton, {
            quaternary: true,
            circle: true,
            size: "small",
            type: "error",
            title: "删除",
            loading: deletingId.value === row.id,
            onClick: () => handleDelete(row),
          }, { icon: () => h(NIcon, { component: TrashOutline }) }),
        ],
      });
    },
  },
];

const themeOptions = [
  { label: "跟随系统", value: "auto" },
  { label: "浅色", value: "light" },
  { label: "深色", value: "dark" },
];

const languageOptions = [
  { label: "简体中文", value: "zh-CN" },
  { label: "English", value: "en" },
];

function onSubmit() {
  loading.value = true;
  themeOption.value = form.theme as ThemeOption;
  localStorage.setItem(themeStorageKey, form.theme);
  setTimeout(() => {
    loading.value = false;
    message.success("设置已保存");
  }, 300);
}

function onReset() {
  form.appName = "状态机平台";
  form.apiBaseUrl = "http://localhost:3000";
  form.theme = themeOption.value;
  form.language = "zh-CN";
  form.enableNotifications = true;
  form.maxHistoryItems = 100;
}
</script>

<template>
  <div class="page settings-page">
    <div class="settings-layout">
      <!-- API Key 管理 -->
      <n-card class="settings-card" title="API Key 管理" :bordered="false">
        <template #header-extra>
          <n-button type="primary" size="small" @click="handleCreate">创建 API Key</n-button>
        </template>
        <p class="settings-card-desc">用于接口鉴权，请求时将作为 X-API-Key 请求头发送。Key 格式：smKey-xxx（32 位）。</p>
        <n-data-table
          :columns="apiKeyColumns"
          :data="apiKeyList"
          :loading="apiKeyLoading"
          :bordered="false"
          striped
          size="small"
          :single-line="false"
          class="api-key-table"
        />
      </n-card>

      <!-- 通用设置 -->
      <n-card class="settings-card" title="通用设置" :bordered="false">
        <n-form :model="form" label-placement="left" label-width="120" class="settings-form">
          <n-form-item label="应用名称">
            <n-input v-model:value="form.appName" placeholder="应用名称" />
          </n-form-item>
          <n-form-item label="API 地址">
            <n-input v-model:value="form.apiBaseUrl" placeholder="http://localhost:3000" />
          </n-form-item>
          <n-form-item label="主题">
            <n-select v-model:value="form.theme" :options="themeOptions" />
          </n-form-item>
          <n-form-item label="语言">
            <n-select v-model:value="form.language" :options="languageOptions" />
          </n-form-item>
          <n-form-item label="消息通知">
            <n-switch v-model:value="form.enableNotifications" />
          </n-form-item>
          <n-form-item label="历史记录条数">
            <n-input-number v-model:value="form.maxHistoryItems" :min="10" :max="1000" placeholder="100" />
          </n-form-item>
          <n-form-item>
            <n-space>
              <n-button type="primary" :loading="loading" @click="onSubmit">保存</n-button>
              <n-button @click="onReset">重置</n-button>
            </n-space>
          </n-form-item>
        </n-form>
      </n-card>
    </div>

    <n-modal v-model:show="showCreateModal" preset="dialog" title="创建 API Key" positive-text="创建" negative-text="取消" :loading="creating" @positive-click="confirmCreate">
      <n-form label-placement="left" label-width="80" style="margin-top: 16px">
        <n-form-item label="名称">
          <n-input v-model:value="createName" placeholder="例如：开发环境、生产环境" clearable />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<style scoped>
.settings-page {
  padding: 20px 24px;
  height: auto;
  overflow: auto;
}

.settings-layout {
  max-width: 900px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.settings-card {
  border-radius: 8px;
}

.settings-card :deep(.n-card__content) {
  padding-top: 4px;
}

.settings-card-desc {
  margin: 0 0 16px;
  font-size: 13px;
  color: var(--n-text-color-3);
  line-height: 1.5;
}

.api-key-table {
  width: 100%;
}

.settings-form {
  max-width: 480px;
}

.api-key-mask {
  font-family: monospace;
  margin-right: 8px;
}
</style>
