/// <reference types="vite/client" />
import type { Node, Edge } from "@vue-flow/core";
import { http } from "./request";

/** 状态机列表项（列表页表格行） */
export interface StateMachineListItem {
  id: string;
  name: string;
  description: string;
  baseUrl?: string;
  createdAt: string;
  updatedAt: string;
}

/** 状态机详情（含流程数据，用于设计页加载） */
export interface StateMachineDetail extends StateMachineListItem {
  /** 流程节点与边，设计页编辑器使用 */
  flowData: FlowData;
}

export interface FlowData {
  nodes: Node[];
  edges: Edge[];
}

/** 获取状态机列表 - 请求 */
export interface StateMachineListRequest {
  page?: number;
  pageSize?: number;
  keyword?: string;
}

/** 获取状态机列表 - 返回 */
export interface StateMachineListResponse {
  list: StateMachineListItem[];
  total: number;
}

/** 创建状态机 - 请求 */
export interface StateMachineCreateRequest {
  name: string;
  description?: string;
}

/** 创建状态机 - 返回 */
export interface StateMachineCreateResponse {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

/** 更新状态机（含流程保存）- 请求 */
export interface StateMachineUpdateRequest {
  name?: string;
  description?: string;
  baseUrl?: string;
  flowData?: FlowData;
}

/** 更新状态机 - 返回 */
export interface StateMachineUpdateResponse {
  id: string;
  name: string;
  description: string;
  baseUrl?: string;
  updatedAt: string;
}

/** 保存流程 - 请求（与 FlowData 一致：nodes + edges） */
export type FlowSaveRequest = FlowData;

/** 保存流程 - 返回 */
export interface FlowSaveResponse {
  ok: boolean;
  updatedAt?: string;
}

/** 更新单个节点 - 请求（与 nodes 数组中单条一致） */
export interface UpdateNodeRequest {
  id: string;
  type: string;
  position: { x: number; y: number };
  data: Record<string, unknown>;
}

/** 更新单个节点 - 返回 */
export interface UpdateNodeResponse {
  ok: boolean;
}

const BASE = import.meta.env.VITE_API_BASE ?? "";

/**
 * 获取状态机列表
 * GET /flow?page=1&pageSize=10&keyword=xxx
 */
export async function getStateMachineList(
  req: StateMachineListRequest = {}
): Promise<StateMachineListResponse> {
  const params = new URLSearchParams();
  if (req.page != null) params.set("page", String(req.page));
  if (req.pageSize != null) params.set("pageSize", String(req.pageSize));
  if (req.keyword) params.set("keyword", req.keyword);
  const query = params.toString();
  return http<StateMachineListResponse>({
    method: "GET",
    url: `${BASE}/flow${query ? `?${query}` : ""}`,
  });
}

/**
 * 获取单个状态机详情（含流程数据，用于设计页）
 * GET /flow/:id
 */
export async function getStateMachineById(id: string): Promise<StateMachineDetail> {
  return http<StateMachineDetail>({
    method: "GET",
    url: `${BASE}/flow/${id}`,
  });
}

/**
 * 获取状态机的流程数据（设计页加载画布）
 * GET /flow/:id/flow
 * 返回: { nodes, edges }
 * 未传 stateMachineId 时返回空流程。
 */
export async function getFlowData(stateMachineId?: string): Promise<FlowData> {
  if (!stateMachineId) return { nodes: [], edges: [] };
  try {
    const data = await http<FlowData>({
      method: "GET",
      url: `${BASE}/flow/${stateMachineId}/flow`,
    });
    if (!data || !Array.isArray(data.nodes) || !Array.isArray(data.edges)) {
      return { nodes: [], edges: [] };
    }
    return { nodes: data.nodes, edges: data.edges };
  } catch {
    return { nodes: [], edges: [] };
  }
}

/**
 * 保存状态机流程数据（设计页保存画布）
 * PUT /flow/:id/flow
 * 请求体: FlowSaveRequest { nodes: Node[], edges: Edge[] }
 * 返回: FlowSaveResponse { ok: boolean, updatedAt?: string }
 */
export async function saveFlow(
  stateMachineId: string,
  flowData: FlowSaveRequest
): Promise<FlowSaveResponse> {
  return http<FlowSaveResponse>({
    method: "PUT",
    url: `${BASE}/flow/${stateMachineId}/flow`,
    body: flowData,
  });
}

/**
 * 更新单个节点（编辑窗口保存时调用）
 * PUT /flow/:id/nodes/:nodeId
 */
export async function updateNode(
  stateMachineId: string,
  nodeId: string,
  payload: UpdateNodeRequest
): Promise<UpdateNodeResponse> {
  return http<UpdateNodeResponse>({
    method: "PUT",
    url: `${BASE}/flow/${stateMachineId}/nodes/${encodeURIComponent(nodeId)}`,
    body: payload,
  });
}

/**
 * 创建单个节点（创建节点时保存到服务端）
 * POST /flow/:id/nodes
 */
export async function createNode(
  stateMachineId: string,
  payload: UpdateNodeRequest
): Promise<UpdateNodeResponse> {
  return http<UpdateNodeResponse>({
    method: "POST",
    url: `${BASE}/flow/${stateMachineId}/nodes`,
    body: payload,
  });
}

/**
 * 创建状态机
 * POST /flow
 * Body: StateMachineCreateRequest
 */
export async function createStateMachine(
  req: StateMachineCreateRequest
): Promise<StateMachineCreateResponse> {
  return http<StateMachineCreateResponse>({
    method: "POST",
    url: `${BASE}/flow`,
    body: req,
  });
}

/**
 * 更新状态机（可含流程数据保存）
 * PUT /flow/:id
 * Body: StateMachineUpdateRequest
 */
export async function updateStateMachine(
  id: string,
  req: StateMachineUpdateRequest
): Promise<StateMachineUpdateResponse> {
  return http<StateMachineUpdateResponse>({
    method: "PUT",
    url: `${BASE}/flow/${id}`,
    body: req,
  });
}

/**
 * 删除状态机
 * DELETE /flow/:id
 */
export async function deleteStateMachine(id: string): Promise<void> {
  await http<void>({
    method: "DELETE",
    url: `${BASE}/flow/${id}`,
  });
}
