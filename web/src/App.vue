<script setup lang="ts">
import { h, ref, computed, provide, watch, onMounted, onUnmounted, type Component } from "vue";
import { useRouter, useRoute } from "vue-router";
import { NConfigProvider, NMessageProvider, NDialogProvider, darkTheme, zhCN, dateZhCN, NLayout, NLayoutSider, NLayoutContent, NMenu, NIcon, type MenuOption } from "naive-ui";
import { BookOutline, DocumentTextOutline, SettingsOutline, ShareSocialOutline } from "@vicons/ionicons5";
import type { ThemeOption } from "./api/settings";

const router = useRouter();
const route = useRoute();

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) });
}

// 主题：由设置页控制，持久化到 localStorage
const THEME_STORAGE_KEY = "app-theme";

const themeOption = ref<ThemeOption>(
  (localStorage.getItem(THEME_STORAGE_KEY) as ThemeOption) ?? "auto"
);
const systemDark = ref(window.matchMedia("(prefers-color-scheme: dark)").matches);
const dark = computed(
  () =>
    themeOption.value === "dark" ? true : themeOption.value === "light" ? false : systemDark.value
);
const theme = computed(() => (dark.value ? darkTheme : null));
provide("app-dark-mode", dark);
provide("app-theme-option", themeOption);
provide("app-theme-storage-key", THEME_STORAGE_KEY);

const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
const handleThemeChange = (e: MediaQueryListEvent) => {
  systemDark.value = e.matches;
};

onMounted(() => {
  mediaQuery.addEventListener("change", handleThemeChange);
});

onUnmounted(() => {
  mediaQuery.removeEventListener("change", handleThemeChange);
});

const menuOptions: MenuOption[] = [
  {
    label: "状态机",
    key: "state-machines",
    icon: renderIcon(ShareSocialOutline),
    children: [
      { label: "列表", key: "/flow/list" },
      { label: "设计", key: "/flow/design" },
    ],
  },
  {
    label: "会话",
    key: "sessions",
    icon: renderIcon(BookOutline),
    children: [
      { label: "列表", key: "/sessions/list" },
      { label: "历史", key: "/sessions/history" },
    ],
  },
  {
    label: "文档",
    key: "/docs",
    icon: renderIcon(DocumentTextOutline),
  },
  {
    label: "设置",
    key: "/settings",
    icon: renderIcon(SettingsOutline),
  },
];

const activeKey = computed(() => route.path);

const expandedKeys = ref<string[]>([]);
watch(
  () => route.path,
  (path) => {
    const keys: string[] = [];
    if (path.startsWith("/flow")) keys.push("state-machines");
    if (path.startsWith("/sessions")) keys.push("sessions");
    expandedKeys.value = keys;
  },
  { immediate: true }
);

function handleMenuSelect(key: string) {
  if (key.startsWith("/")) router.push(key);
}

const collapsed = ref(false);
</script>

<template>
  <n-config-provider :theme="theme" :locale="zhCN" :date-locale="dateZhCN" class="app-provider">
    <n-message-provider>
      <n-dialog-provider>
      <n-layout has-sider class="main-layout">
      <n-layout-sider
        bordered
        collapse-mode="width"
        :collapsed-width="64"
        :width="240"
        :collapsed="collapsed"
        show-trigger
        @collapse="collapsed = true"
        @expand="collapsed = false"
      >
        <n-menu
          :value="activeKey"
          v-model:expanded-keys="expandedKeys"
          :collapsed="collapsed"
          :collapsed-width="64"
          :collapsed-icon-size="22"
          :options="menuOptions"
          @update:value="handleMenuSelect"
        />
      </n-layout-sider>
      <n-layout-content class="content-layout">
        <router-view />
      </n-layout-content>
    </n-layout>
      </n-dialog-provider>
    </n-message-provider>
  </n-config-provider>
</template>

<style>
/* 全局隐藏滚动条，保留滚动 */
* {
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE/Edge */
}
*::-webkit-scrollbar {
  display: none; /* Chrome, Safari, Opera */
}

#app {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  width: 100%;
  min-width: 1280px;
  overflow: hidden;
}

.app-provider {
  height: 100%;
}

.main-layout {
  height: 100%;
}

.content-layout {
  height: 100%;
  display: flex;
  flex-direction: column;
}
</style>
