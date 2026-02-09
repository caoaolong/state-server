<script setup lang="ts">
import { ref, reactive } from "vue";
import { NCard, NForm, NFormItem, NInput, NInputNumber, NButton, NSpace, NSelect, NSwitch } from "naive-ui";

const loading = ref(false);

const form = reactive({
  appName: "状态机平台",
  apiBaseUrl: "http://localhost:3000",
  theme: "auto",
  language: "zh-CN",
  enableNotifications: true,
  maxHistoryItems: 100,
});

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
  setTimeout(() => {
    loading.value = false;
    console.log("保存设置", form);
  }, 500);
}

function onReset() {
  form.appName = "状态机平台";
  form.apiBaseUrl = "http://localhost:3000";
  form.theme = "auto";
  form.language = "zh-CN";
  form.enableNotifications = true;
  form.maxHistoryItems = 100;
}
</script>

<template>
  <div class="page settings-page">
    <n-card title="设置" :bordered="false">
      <n-form :model="form" label-placement="left" label-width="120" style="max-width: 520px">
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
</template>

<style scoped>
.settings-page {
  padding: 16px;
  height: auto;
  overflow: auto;
}
</style>
