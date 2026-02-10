<script setup lang="ts">
import { NCard, NCode } from "naive-ui";

const baseUrl = import.meta.env.VITE_API_BASE || "/api";
const authHeaderCode = "X-API-Key: <your-api-key>";

/** 生成接口锚点 id（用于右侧导航定位） */
function getApiAnchorId(api: { method: string; path: string }): string {
  const pathPart = api.path
    .replace(/^\//, "")
    .replace(/\//g, "-")
    .replace(/:(\w+)/g, "$1");
  return `${api.method.toLowerCase()}-${pathPart}`;
}

const apiSections = [
  {
    title: "状态机",
    apis: [
      {
        name: "获取状态机列表",
        method: "GET",
        path: "/state-machines",
        auth: false,
        request: {
          type: "query",
          params: [
            { name: "page", type: "number", required: false, desc: "页码" },
            { name: "pageSize", type: "number", required: false, desc: "每页条数" },
            { name: "keyword", type: "string", required: false, desc: "关键词搜索" },
          ],
        },
        response: `{
  "list": [
    {
      "id": "sm-001",
      "name": "示例状态机",
      "description": "描述",
      "createdAt": "2026-02-09T08:16:40.049Z",
      "updatedAt": "2026-02-09T08:16:40.049Z"
    }
  ],
  "total": 1
}`,
        responseLanguage: "json",
      },
      {
        name: "获取状态机详情",
        method: "GET",
        path: "/state-machines/:id",
        auth: false,
        request: { type: "path", params: [{ name: "id", type: "string", required: true, desc: "状态机 ID" }] },
        response: `{
  "id": "sm-001",
  "name": "示例状态机",
  "description": "描述",
  "createdAt": "2026-02-09T08:16:40.049Z",
  "updatedAt": "2026-02-09T08:16:40.049Z",
  "flowData": {
    "nodes": [],
    "edges": []
  }
}`,
        responseLanguage: "json",
      },
      {
        name: "获取流程数据",
        method: "GET",
        path: "/state-machines/:id/flow",
        auth: false,
        request: { type: "path", params: [{ name: "id", type: "string", required: true, desc: "状态机 ID" }] },
        response: `{
  "nodes": [
    {
      "id": "scene-start-1",
      "type": "default",
      "position": { "x": 100, "y": 200 },
      "data": { "label": "开始", "nodeCategory": "scene", "nodeKind": "start" }
    }
  ],
  "edges": [
    { "id": "edge-1", "source": "scene-start-1", "target": "scene-1" }
  ]
}`,
        responseLanguage: "json",
      },
      {
        name: "创建状态机",
        method: "POST",
        path: "/state-machines",
        auth: false,
        request: {
          type: "body",
          params: [
            { name: "name", type: "string", required: true, desc: "名称" },
            { name: "description", type: "string", required: false, desc: "描述" },
          ],
        },
        response: `{
  "id": "sm-001",
  "name": "未命名状态机",
  "description": "",
  "createdAt": "2026-02-09T08:16:40.049Z",
  "updatedAt": "2026-02-09T08:16:40.049Z"
}`,
        responseLanguage: "json",
      },
      {
        name: "更新状态机",
        method: "PUT",
        path: "/state-machines/:id",
        auth: false,
        request: {
          type: "path+body",
          pathParams: [{ name: "id", type: "string", required: true, desc: "状态机 ID" }],
          bodyParams: [
            { name: "name", type: "string", required: false, desc: "名称" },
            { name: "description", type: "string", required: false, desc: "描述" },
            { name: "flowData", type: "FlowData", required: false, desc: "流程数据 { nodes, edges }" },
          ],
        },
        response: `{
  "id": "sm-001",
  "name": "示例状态机",
  "description": "描述",
  "updatedAt": "2026-02-09T08:16:40.049Z"
}`,
        responseLanguage: "json",
      },
      {
        name: "保存流程",
        method: "PUT",
        path: "/state-machines/:id/flow",
        auth: false,
        request: {
          type: "path+body",
          pathParams: [{ name: "id", type: "string", required: true, desc: "状态机 ID" }],
          bodyParams: [
            { name: "nodes", type: "Node[]", required: true, desc: "节点列表" },
            { name: "edges", type: "Edge[]", required: true, desc: "边列表" },
          ],
        },
        response: `{
  "ok": true,
  "updatedAt": "2026-02-09T08:16:40.049Z"
}`,
        responseLanguage: "json",
      },
      {
        name: "删除状态机",
        method: "DELETE",
        path: "/state-machines/:id",
        auth: false,
        request: { type: "path", params: [{ name: "id", type: "string", required: true, desc: "状态机 ID" }] },
        response: "无 body，成功为 2xx",
        responseLanguage: "text",
      },
    ],
  },
  {
    title: "会话",
    apis: [
      {
        name: "获取会话列表",
        method: "GET",
        path: "/sessions",
        auth: false,
        request: {
          type: "query",
          params: [
            { name: "page", type: "number", required: false, desc: "页码" },
            { name: "pageSize", type: "number", required: false, desc: "每页条数" },
            { name: "stateMachineId", type: "string", required: false, desc: "状态机 ID" },
            { name: "status", type: "string", required: false, desc: "running | ended | suspended" },
          ],
        },
        response: `{
  "list": [
    {
      "id": "1",
      "sessionId": "sess-001",
      "stateMachineId": "sm-001",
      "status": "running",
      "createdAt": "2026-02-09T08:16:40.049Z"
    }
  ],
  "total": 1
}`,
        responseLanguage: "json",
      },
      {
        name: "获取会话详情",
        method: "GET",
        path: "/sessions/:id",
        auth: false,
        request: { type: "path", params: [{ name: "id", type: "string", required: true, desc: "会话 ID" }] },
        response: `{
  "id": "1",
  "sessionId": "sess-001",
  "stateMachineId": "sm-001",
  "status": "running",
  "createdAt": "2026-02-09T08:16:40.049Z"
}`,
        responseLanguage: "json",
      },
      {
        name: "获取会话历史",
        method: "GET",
        path: "/sessions/history",
        auth: false,
        request: {
          type: "query",
          params: [
            { name: "sessionId", type: "string", required: false, desc: "会话 ID" },
            { name: "page", type: "number", required: false, desc: "页码" },
            { name: "pageSize", type: "number", required: false, desc: "每页条数" },
          ],
        },
        response: `{
  "list": [
    {
      "id": "1",
      "sessionId": "sess-001",
      "event": "submit",
      "fromState": "draft",
      "toState": "pending",
      "createdAt": "2026-02-09T08:16:40.049Z"
    }
  ],
  "total": 1
}`,
        responseLanguage: "json",
      },
    ],
  },
  {
    title: "设置",
    apis: [
      {
        name: "获取设置",
        method: "GET",
        path: "/settings",
        auth: false,
        request: { type: "none" },
        response: `{
  "appName": "状态机平台",
  "apiBaseUrl": "http://localhost:8080",
  "theme": "auto",
  "language": "zh-CN",
  "enableNotifications": true,
  "maxHistoryItems": 100
}`,
        responseLanguage: "json",
      },
      {
        name: "保存设置",
        method: "PUT",
        path: "/settings",
        auth: false,
        request: {
          type: "body",
          params: [
            { name: "appName", type: "string", required: true, desc: "应用名称" },
            { name: "apiBaseUrl", type: "string", required: true, desc: "API 地址" },
            { name: "theme", type: "string", required: true, desc: "auto|light|dark" },
            { name: "language", type: "string", required: true, desc: "zh-CN|en" },
            { name: "enableNotifications", type: "boolean", required: true, desc: "是否开启通知" },
            { name: "maxHistoryItems", type: "number", required: true, desc: "历史记录条数" },
          ],
        },
        response: `{
  "ok": true
}`,
        responseLanguage: "json",
      },
    ],
  },
];
</script>

<template>
  <div class="page docs-page">
    <n-card title="API 文档" :bordered="false">
      <div class="docs-layout">
        <div class="docs-content">
          <!-- 接口鉴权说明 -->
          <section id="doc-auth" class="doc-section">
            <h2>接口鉴权说明（API Key 校验）</h2>
          <p>接口采用 <strong>API Key</strong> 鉴权：在请求头中携带 <n-code :code="authHeaderCode" inline />，服务端校验通过后允许访问。</p>
          <ul>
            <li>在本应用的 <strong>设置</strong> 页面中配置并保存 API Key，后续所有请求会自动在请求头中附带 <n-code code="X-API-Key" inline />。</li>
            <li>若未配置 API Key 或服务端未启用鉴权，请求可能被拒绝或按服务端策略处理。</li>
          </ul>
          <p>请求地址基础路径为：<n-code :code="baseUrl" inline />（开发环境经代理转发）。</p>
        </section>

        <!-- API 列表 -->
        <section v-for="section in apiSections" :key="section.title" class="doc-section" :id="'section-' + section.title">
          <h2>{{ section.title }}</h2>
          <div v-for="api in section.apis" :key="api.method + api.path" :id="getApiAnchorId(api)" class="api-block">
            <h3>
              <span class="api-method" :class="api.method">{{ api.method }}</span>
              {{ api.name }}
            </h3>
            <div class="api-path">
              <span class="label">请求地址</span>
              <n-code :code="baseUrl + api.path" inline />
            </div>
            <div v-if="api.request && api.request.type !== 'none'" class="api-request">
              <span class="label">请求参数</span>
              <template v-if="api.request.type === 'query'">
                <p>Query：</p>
                <ul>
                  <li v-for="p in api.request.params" :key="p.name">
                    <n-code :code="p.name" inline /> ({{ p.type }}) {{ p.required ? "必填" : "选填" }} — {{ p.desc }}
                  </li>
                </ul>
              </template>
              <template v-else-if="api.request.type === 'path'">
                <p>路径参数：</p>
                <ul>
                  <li v-for="p in api.request.params" :key="p.name">
                    <n-code :code="p.name" inline /> ({{ p.type }}) — {{ p.desc }}
                  </li>
                </ul>
              </template>
              <template v-else-if="api.request.type === 'body'">
                <p>Body (JSON)：</p>
                <ul>
                  <li v-for="p in api.request.params" :key="p.name">
                    <n-code :code="p.name" inline /> ({{ p.type }}) {{ p.required ? "必填" : "选填" }} — {{ p.desc }}
                  </li>
                </ul>
              </template>
              <template v-else-if="api.request.type === 'path+body' && 'pathParams' in api.request && 'bodyParams' in api.request">
                <p>路径参数：</p>
                <ul>
                  <li v-for="p in api.request.pathParams" :key="p.name">
                    <n-code :code="p.name" inline /> ({{ p.type }}) — {{ p.desc }}
                  </li>
                </ul>
                <p>Body (JSON)：</p>
                <ul>
                  <li v-for="p in api.request.bodyParams" :key="p.name">
                    <n-code :code="p.name" inline /> ({{ p.type }}) {{ p.required ? "必填" : "选填" }} — {{ p.desc }}
                  </li>
                </ul>
              </template>
            </div>
            <div class="api-response">
              <span class="label">响应格式</span>
              <n-code :code="api.response" :language="api.responseLanguage ?? 'json'" :word-wrap="true" />
            </div>
          </div>
        </section>
        </div>
        <!-- 右侧导航 -->
        <nav class="docs-nav">
          <div class="docs-nav-title">导航</div>
          <a href="#doc-auth" class="docs-nav-link">接口鉴权说明</a>
          <template v-for="section in apiSections" :key="section.title">
            <div class="docs-nav-section">{{ section.title }}</div>
            <a
              v-for="api in section.apis"
              :key="api.method + api.path"
              :href="'#' + getApiAnchorId(api)"
              class="docs-nav-link"
            >
              <span class="docs-nav-method" :class="api.method">{{ api.method }}</span>
              {{ api.name }}
            </a>
          </template>
        </nav>
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.docs-page {
  padding: 16px;
  height: 100%;
  overflow: auto;
}

.docs-layout {
  display: flex;
  gap: 24px;
  align-items: flex-start;
}

.docs-content {
  flex: 1;
  min-width: 0;
  max-width: 800px;
  line-height: 1.6;
}

.docs-nav {
  position: sticky;
  top: 16px;
  width: 220px;
  flex-shrink: 0;
  padding: 12px 0;
  border-left: 1px solid var(--n-border-color);
  padding-left: 16px;
}

.docs-nav-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--n-text-color-2);
  margin-bottom: 12px;
}

.docs-nav-section {
  font-size: 12px;
  font-weight: 600;
  color: var(--n-text-color-3);
  margin-top: 16px;
  margin-bottom: 8px;
}

.docs-nav-section:first-child {
  margin-top: 0;
}

.docs-nav-link {
  display: block;
  font-size: 12px;
  color: var(--n-text-color);
  text-decoration: none;
  padding: 4px 0;
  line-height: 1.4;
  border-radius: 4px;
  transition: background 0.15s;
}

.docs-nav-link:hover {
  color: var(--n-primary-color);
  background: var(--n-color-hover);
}

.docs-nav-method {
  display: inline-block;
  margin-right: 6px;
  padding: 1px 4px;
  border-radius: 2px;
  font-size: 10px;
  font-weight: 600;
}

.docs-nav-method.GET { background: #e8f5e9; color: #2e7d32; }
.docs-nav-method.POST { background: #e3f2fd; color: #1565c0; }
.docs-nav-method.PUT { background: #fff3e0; color: #e65100; }
.docs-nav-method.DELETE { background: #ffebee; color: #c62828; }

.doc-section {
  margin-bottom: 32px;
}

.doc-section h2 {
  margin-top: 28px;
  margin-bottom: 16px;
  font-size: 1.25rem;
  border-bottom: 1px solid var(--n-border-color);
  padding-bottom: 8px;
}

.doc-section h2:first-child {
  margin-top: 0;
}

.api-block {
  margin-bottom: 24px;
  padding: 16px;
  background: var(--n-color-modal);
  border-radius: 8px;
  border: 1px solid var(--n-border-color);
}

.api-block h3 {
  margin: 0 0 12px 0;
  font-size: 1rem;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.api-method {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
}

.api-method.GET {
  background: #e8f5e9;
  color: #2e7d32;
}

.api-method.POST {
  background: #e3f2fd;
  color: #1565c0;
}

.api-method.PUT {
  background: #fff3e0;
  color: #e65100;
}

.api-method.DELETE {
  background: #ffebee;
  color: #c62828;
}

.label {
  display: block;
  font-size: 12px;
  color: var(--n-text-color-3);
  margin-bottom: 6px;
  font-weight: 500;
}

.api-path {
  margin-bottom: 12px;
}

.api-path :deep(.n-code) {
  font-size: 13px;
  word-break: break-all;
}

.api-request {
  margin-bottom: 12px;
}

.api-request p {
  margin: 4px 0 4px 0;
  font-size: 13px;
}

.api-request ul {
  margin: 0 0 8px 0;
  padding-left: 1.5em;
  font-size: 13px;
}

.api-request li {
  margin-bottom: 4px;
}

.api-response :deep(.n-code) {
  margin: 0;
  border-radius: 4px;
  font-size: 12px;
}

.docs-content p {
  margin: 0 0 10px 0;
  color: var(--n-text-color);
  font-size: 14px;
}

.docs-content ul {
  margin: 0 0 10px 0;
  padding-left: 1.5em;
  font-size: 14px;
}

.docs-content :deep(.n-code.n-code--inline) {
  font-size: 13px;
}
</style>
