/**
 * API Key 管理：列表、创建、刷新、删除、查看明文（复制用）
 */

import { http } from "./request";

export interface ApiKeyItem {
  id: string;
  name: string;
  apiKey: string;
  createdAt: string;
}

const BASE = import.meta.env.VITE_API_BASE ?? "";

/** 获取 API Key 列表（apiKey 为脱敏） */
export async function getApiKeyList(): Promise<ApiKeyItem[]> {
  const res = await http<{ list: ApiKeyItem[] }>({
    method: "GET",
    url: `${BASE}/api-keys`,
  });
  return res?.list ?? [];
}

/** 创建 API Key，返回新建项（含完整 apiKey，仅此一次） */
export async function createApiKey(name: string): Promise<ApiKeyItem> {
  return http<ApiKeyItem>({
    method: "POST",
    url: `${BASE}/api-keys`,
    body: { name: name.trim() || "未命名" },
  });
}

/** 刷新（重新生成）Key，返回完整 apiKey（仅此一次） */
export async function refreshApiKey(id: string): Promise<{ apiKey: string }> {
  return http<{ apiKey: string }>({
    method: "PUT",
    url: `${BASE}/api-keys/${id}/refresh`,
  });
}

/** 查看完整 Key（用于复制） */
export async function revealApiKey(id: string): Promise<{ apiKey: string }> {
  return http<{ apiKey: string }>({
    method: "GET",
    url: `${BASE}/api-keys/${id}/reveal`,
  });
}

/** 删除 API Key */
export async function deleteApiKey(id: string): Promise<void> {
  await http<void>({
    method: "DELETE",
    url: `${BASE}/api-keys/${id}`,
  });
}
