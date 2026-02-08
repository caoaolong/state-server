<script setup lang="ts">
import { h, ref, computed, provide, onMounted, onUnmounted, type Component } from "vue";
import { NConfigProvider, darkTheme, zhCN, dateZhCN, NLayout, NLayoutSider, NLayoutContent, NMenu, NIcon, type MenuOption } from "naive-ui";
import { BookOutline, SettingsOutline, ShareSocialOutline } from "@vicons/ionicons5";
import FlowEditor from "./components/FlowEditor.vue";

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) });
}

// Global Theme Management
const dark = ref(window.matchMedia("(prefers-color-scheme: dark)").matches);
const theme = computed(() => (dark.value ? darkTheme : null));
provide('app-dark-mode', dark);

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
    label: "流程设计",
    key: "flow-design",
    icon: renderIcon(ShareSocialOutline),
  },
  {
    label: "我的项目",
    key: "projects",
    icon: renderIcon(BookOutline),
  },
  {
    label: "系统设置",
    key: "settings",
    icon: renderIcon(SettingsOutline),
  },
];

const activeKey = ref("flow-design");
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
          v-model:value="activeKey"
          :collapsed="collapsed"
          :collapsed-width="64"
          :collapsed-icon-size="22"
          :options="menuOptions"
        />
      </n-layout-sider>
      <n-layout-content class="content-layout">
        <FlowEditor />
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
