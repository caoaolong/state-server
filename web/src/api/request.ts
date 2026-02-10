/**
 * 请求封装层：统一 HTTP 调用入口，便于后续切换实现（网页版 fetch / Wails 桌面版 Go 绑定）。
 * 业务接口只依赖 request，不直接使用 fetch。
 */

export type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";

export interface RequestConfig {
  method: HttpMethod;
  /** 完整 URL 或相对路径（会与 baseURL 拼接，由实现决定） */
  url: string;
  /** JSON 请求体，仅 POST/PUT/PATCH 使用 */
  body?: unknown;
  headers?: Record<string, string>;
}

/**
 * HTTP 客户端接口：同一抽象，可替换为 fetch 实现或 Wails 桥接实现。
 */
export interface HttpClient {
  <T = unknown>(config: RequestConfig): Promise<T>;
}

const API_KEYS_STORAGE_KEY = "state-server-api-keys";

export interface ApiKeyItem {
  id: string;
  name: string;
  apiKey: string;
  createdAt: string;
}

function loadApiKeys(): ApiKeyItem[] {
  try {
    const raw = localStorage.getItem(API_KEYS_STORAGE_KEY);
    if (!raw) return [];
    const list = JSON.parse(raw) as ApiKeyItem[];
    return Array.isArray(list) ? list : [];
  } catch {
    return [];
  }
}

function saveApiKeys(list: ApiKeyItem[]): void {
  localStorage.setItem(API_KEYS_STORAGE_KEY, JSON.stringify(list));
}

const API_KEY_PREFIX = "smKey-";
const API_KEY_TOTAL_LENGTH = 36;
const API_KEY_RANDOM_LENGTH = API_KEY_TOTAL_LENGTH - API_KEY_PREFIX.length;

/** 生成随机 API Key，格式 smKey-xxx，总长度 32 字符 */
function generateRandomKey(): string {
  const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
  let s = "";
  for (let i = 0; i < API_KEY_RANDOM_LENGTH; i++) s += chars.charAt(Math.floor(Math.random() * chars.length));
  return API_KEY_PREFIX + s;
}

/** 获取全部 API Key 列表 */
export function getApiKeys(): ApiKeyItem[] {
  return loadApiKeys();
}

/** 当前请求使用的 API Key（取列表中第一个） */
export function getApiKey(): string {
  const list = loadApiKeys();
  return list.length > 0 ? list[0].apiKey : "";
}

/** 创建新 API Key，返回新建项（含生成的 key） */
export function createApiKey(name: string): ApiKeyItem {
  const list = loadApiKeys();
  const item: ApiKeyItem = {
    id: crypto.randomUUID(),
    name: name.trim() || "未命名",
    apiKey: generateRandomKey(),
    createdAt: new Date().toISOString(),
  };
  list.unshift(item);
  saveApiKeys(list);
  return item;
}

/** 删除 API Key */
export function deleteApiKey(id: string): void {
  const list = loadApiKeys().filter((x) => x.id !== id);
  saveApiKeys(list);
}

/** 刷新（重新生成）指定 API Key，返回更新后的项 */
export function refreshApiKey(id: string): ApiKeyItem | null {
  const list = loadApiKeys();
  const idx = list.findIndex((x) => x.id === id);
  if (idx === -1) return null;
  list[idx] = { ...list[idx], apiKey: generateRandomKey(), createdAt: new Date().toISOString() };
  saveApiKeys(list);
  return list[idx];
}

/**
 * 网页版：基于 fetch 的实现。
 * 鉴权：从本地读取 API Key，以 X-API-Key 请求头发送。
 * 桌面版可替换为调用 window.go.main.App.XXX 等 Wails 绑定，实现同一 HttpClient 接口。
 */
export const request: HttpClient = async <T = unknown>(config: RequestConfig): Promise<T> => {
  const { method, url, body, headers: customHeaders } = config;
  const apiKey = getApiKey();
  const headers: Record<string, string> = {
    ...(body != null && { "Content-Type": "application/json" }),
    ...(apiKey && { "X-API-Key": apiKey }),
    ...customHeaders,
  };
  const res = await fetch(url, {
    method,
    headers: Object.keys(headers).length ? headers : undefined,
    body: body != null ? JSON.stringify(body) : undefined,
  });
  if (!res.ok) throw new Error(res.statusText);
  if (res.status === 204) return undefined as T;
  const text = await res.text();
  return (text ? JSON.parse(text) : undefined) as T;
};

/** 当前使用的客户端，入口处可替换为 Wails 版本，例如： setHttpClient(wailsClient) */
let currentClient: HttpClient = request;

export function getHttpClient(): HttpClient {
  return currentClient;
}

export function setHttpClient(client: HttpClient): void {
  currentClient = client;
}

/**
 * 业务代码统一使用的请求方法（走当前注入的 client）。
 */
export function http<T = unknown>(config: RequestConfig): Promise<T> {
  return currentClient(config);
}
