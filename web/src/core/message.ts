export interface FlowNodeStateMessage {
    nodeId: string;
    state: 'running' | 'paused' | 'completed' | 'failed' | 'normal';
}