import type { Node, Edge } from "@vue-flow/core";
import { http } from "./request";

/** 状态机列表项（列表页表格行） */
export interface StateMachineListItem {
  id: string;
  name: string;
  description: string;
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
  flowData?: FlowData;
}

/** 更新状态机 - 返回 */
export interface StateMachineUpdateResponse {
  id: string;
  name: string;
  description: string;
  updatedAt: string;
}

const BASE = import.meta.env.VITE_API_BASE ?? "";

/**
 * 获取状态机列表
 * GET /state-machines?page=1&pageSize=10&keyword=xxx
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
    url: `${BASE}/state-machines${query ? `?${query}` : ""}`,
  });
}

/**
 * 获取单个状态机详情（含流程数据，用于设计页）
 * GET /state-machines/:id
 */
export async function getStateMachineById(id: string): Promise<StateMachineDetail> {
  return http<StateMachineDetail>({
    method: "GET",
    url: `${BASE}/state-machines/${id}`,
  });
}

/**
 * 获取状态机的流程数据（设计页加载画布）
 * GET /state-machines/:id/flow
 * 返回: { nodes, edges }
 * 未传 stateMachineId 时返回空流程。
 */
export async function getFlowData(stateMachineId?: string): Promise<FlowData> {
  if (!stateMachineId) return { nodes: [], edges: [] };
  try {
    const data = await http<FlowData>({
      method: "GET",
      url: `${BASE}/state-machines/${stateMachineId}/flow`,
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
 * 创建状态机
 * POST /state-machines
 * Body: StateMachineCreateRequest
 */
export async function createStateMachine(
  req: StateMachineCreateRequest
): Promise<StateMachineCreateResponse> {
  return http<StateMachineCreateResponse>({
    method: "POST",
    url: `${BASE}/state-machines`,
    body: req,
  });
}

/**
 * 更新状态机（可含流程数据保存）
 * PUT /state-machines/:id
 * Body: StateMachineUpdateRequest
 */
export async function updateStateMachine(
  id: string,
  req: StateMachineUpdateRequest
): Promise<StateMachineUpdateResponse> {
  return http<StateMachineUpdateResponse>({
    method: "PUT",
    url: `${BASE}/state-machines/${id}`,
    body: req,
  });
}

/**
 * 删除状态机
 * DELETE /state-machines/:id
 */
export async function deleteStateMachine(id: string): Promise<void> {
  await http<void>({
    method: "DELETE",
    url: `${BASE}/state-machines/${id}`,
  });
}
