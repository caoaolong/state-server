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

/**
 * 网页版：基于 fetch 的实现。
 * 桌面版可替换为调用 window.go.main.App.XXX 等 Wails 绑定，实现同一 HttpClient 接口。
 */
export const request: HttpClient = async <T = unknown>(config: RequestConfig): Promise<T> => {
  const { method, url, body, headers: customHeaders } = config;
  const headers: Record<string, string> = {
    ...(body != null && { "Content-Type": "application/json" }),
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
