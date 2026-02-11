/** 会话状态 */
export type SessionStatus = "running" | "ended" | "suspended";

/** 会话列表项（会话列表页表格行） */
export interface SessionListItem {
  id: string;
  sessionId: string;
  stateMachineId: string;
  status: SessionStatus;
  createdAt: string;
}

/** 会话历史记录项（会话历史页表格行） */
export interface SessionHistoryItem {
  id: string;
  sessionId: string;
  event: string;
  fromState: string;
  toState: string;
  createdAt: string;
}

/** 获取会话列表 - 请求 */
export interface SessionListRequest {
  page?: number;
  pageSize?: number;
  stateMachineId?: string;
  status?: SessionStatus;
}

/** 获取会话列表 - 返回 */
export interface SessionListResponse {
  list: SessionListItem[];
  total: number;
}

/** 获取会话历史 - 请求 */
export interface SessionHistoryRequest {
  sessionId?: string;
  page?: number;
  pageSize?: number;
}

/** 获取会话历史 - 返回 */
export interface SessionHistoryResponse {
  list: SessionHistoryItem[];
  total: number;
}

import { http } from "./request";

const BASE = import.meta.env.VITE_API_BASE ?? "";

/**
 * 创建会话（设计页进入时调用，sessionId 固定为 0）
 * POST /sessions  body: { stateMachineId, sessionId: 0 }
 */
export interface CreateSessionRequest {
  stateMachineId: string;
  sessionId?: number;
}

export interface CreateSessionResponse {
  id: string;
  sessionId: number;
  stateMachineId: string;
  status: SessionStatus;
  createdAt: string;
}

export async function createSession(req: CreateSessionRequest): Promise<CreateSessionResponse> {
  return http<CreateSessionResponse>({
    method: "POST",
    url: `${BASE}/sessions`,
    body: { stateMachineId: req.stateMachineId, sessionId: req.sessionId ?? 0 },
  });
}

/**
 * 获取会话列表
 * GET /sessions?page=1&pageSize=10&stateMachineId=xxx&status=running
 */
export async function getSessionList(
  req: SessionListRequest = {}
): Promise<SessionListResponse> {
  const params = new URLSearchParams();
  if (req.page != null) params.set("page", String(req.page));
  if (req.pageSize != null) params.set("pageSize", String(req.pageSize));
  if (req.stateMachineId) params.set("stateMachineId", req.stateMachineId);
  if (req.status) params.set("status", req.status);
  const query = params.toString();
  return http<SessionListResponse>({
    method: "GET",
    url: `${BASE}/sessions${query ? `?${query}` : ""}`,
  });
}

/**
 * 获取单个会话详情
 * GET /sessions/:id
 */
export async function getSessionById(sessionId: string): Promise<SessionListItem> {
  return http<SessionListItem>({
    method: "GET",
    url: `${BASE}/sessions/${sessionId}`,
  });
}

/**
 * 获取会话历史记录（可按 sessionId 筛选）
 * GET /sessions/history?sessionId=xxx&page=1&pageSize=20
 */
export async function getSessionHistory(
  req: SessionHistoryRequest = {}
): Promise<SessionHistoryResponse> {
  const params = new URLSearchParams();
  if (req.sessionId) params.set("sessionId", req.sessionId);
  if (req.page != null) params.set("page", String(req.page));
  if (req.pageSize != null) params.set("pageSize", String(req.pageSize));
  const query = params.toString();
  return http<SessionHistoryResponse>({
    method: "GET",
    url: `${BASE}/sessions/history${query ? `?${query}` : ""}`,
  });
}
