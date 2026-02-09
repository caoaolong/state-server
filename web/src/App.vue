<script setup lang="ts">
import { h, ref, computed, provide, watch, onMounted, onUnmounted, type Component } from "vue";
import { useRouter, useRoute } from "vue-router";
import { NConfigProvider, darkTheme, zhCN, dateZhCN, NLayout, NLayoutSider, NLayoutContent, NMenu, NIcon, type MenuOption } from "naive-ui";
import { BookOutline, SettingsOutline, ShareSocialOutline } from "@vicons/ionicons5";

const router = useRouter();
const route = useRoute();

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) });
}

// Global Theme Management
const dark = ref(window.matchMedia("(prefers-color-scheme: dark)").matches);
const theme = computed(() => (dark.value ? darkTheme : null));
provide("app-dark-mode", dark);

const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
const handleThemeChange = (e: MediaQueryListEvent) => {
  dark.value = e.matches;
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
      { label: "列表", key: "/state-machines/list" },
      { label: "设计", key: "/state-machines/design" },
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
    if (path.startsWith("/state-machines")) keys.push("state-machines");
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
  </n-config-provider>
</template>

<style>
#app {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  width: 100%;
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
