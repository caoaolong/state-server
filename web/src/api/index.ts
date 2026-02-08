import type { Node, Edge } from '@vue-flow/core'

export interface FlowData {
  nodes: Node[]
  edges: Edge[]
}

/**
 * 从 JSON 加载流程数据（仅支持 { nodes, edges } 格式）。
 * 若请求失败或数据无效则返回空流程。
 */
export const loadFlowData = async (url: string = '/data/flow.json'): Promise<FlowData> => {
  try {
    const response = await fetch(url)
    if (!response.ok) return { nodes: [], edges: [] }
    const data = await response.json()
    if (!data || !Array.isArray(data.nodes) || !Array.isArray(data.edges)) {
      return { nodes: [], edges: [] }
    }
    return { nodes: data.nodes, edges: data.edges }
  } catch {
    return { nodes: [], edges: [] }
  }
}
