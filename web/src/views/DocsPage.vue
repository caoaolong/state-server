<script setup lang="ts">
import { NCard } from "naive-ui";
</script>

<template>
  <div class="page docs-page">
    <n-card title="操作文档" :bordered="false">
      <div class="docs-content">
        <h2>状态机</h2>
        <h3>列表</h3>
        <p>在状态机列表页可查看所有状态机，支持进入「设计」编辑流程图、点击「删除」删除状态机。</p>
        <h3>设计</h3>
        <p>在设计页可拖拽添加节点（开始、结束、普通场景、选择、结果），连线表示流程走向。支持：</p>
        <ul>
          <li>画布右键：添加节点</li>
          <li>节点工具栏：复制、编辑、删除</li>
          <li>连线：选中后按 Delete 或右键「删除连线」移除</li>
          <li>整理布局：右上角按钮自动排列节点</li>
          <li>保存：填写名称与描述后保存流程图（新建时先创建状态机再保存流程）</li>
        </ul>

        <h2>会话</h2>
        <h3>列表</h3>
        <p>查看当前会话列表及状态。</p>
        <h3>历史</h3>
        <p>查看会话状态变更历史记录。</p>

        <h2>设置</h2>
        <p>配置应用名称、API 地址、主题、语言、消息通知、历史记录条数等。</p>
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

.docs-content {
  max-width: 720px;
  line-height: 1.6;
}

.docs-content h2 {
  margin-top: 24px;
  margin-bottom: 12px;
  font-size: 1.25rem;
}

.docs-content h2:first-child {
  margin-top: 0;
}

.docs-content h3 {
  margin-top: 16px;
  margin-bottom: 8px;
  font-size: 1rem;
}

.docs-content p {
  margin: 0 0 12px;
  color: var(--n-text-color);
}

.docs-content ul {
  margin: 0 0 12px;
  padding-left: 1.5em;
}
</style>
