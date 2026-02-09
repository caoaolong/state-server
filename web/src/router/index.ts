import { createRouter, createWebHistory } from "vue-router";
import type { RouteRecordRaw } from "vue-router";

declare module "vue-router" {
  interface RouteMeta {
    title?: string;
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: "/",
    redirect: "/state-machines/list",
  },
  {
    path: "/state-machines",
    children: [
      {
        path: "list",
        name: "StateMachineList",
        component: () => import("../views/StateMachineList.vue"),
        meta: { title: "状态机列表" },
      },
      {
        path: "design",
        name: "StateMachineDesign",
        component: () => import("../views/StateMachineDesign.vue"),
        meta: { title: "状态机设计" },
      },
    ],
  },
  {
    path: "/sessions",
    children: [
      {
        path: "list",
        name: "SessionList",
        component: () => import("../views/SessionList.vue"),
        meta: { title: "会话列表" },
      },
      {
        path: "history",
        name: "SessionHistory",
        component: () => import("../views/SessionHistory.vue"),
        meta: { title: "会话历史" },
      },
    ],
  },
  {
    path: "/settings",
    name: "Settings",
    component: () => import("../views/SettingsPage.vue"),
    meta: { title: "设置" },
  },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.afterEach((to) => {
  const title = to.meta.title;
  document.title = title ? `StateServer - ${title}` : "StateServer";
});
