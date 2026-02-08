import { ref, type Ref } from "vue";

export interface UseWebSocketOptions {
  /** WebSocket 路径，默认 '/ws' */
  path?: string;
  /** 连接成功回调 */
  onopen?: () => void;
  /** 收到消息回调 */
  onmessage?: (data: string) => void;
  /** 错误回调 */
  onerror?: (event: Event) => void;
  /** 关闭回调 */
  onclose?: () => void;
}

export interface UseWebSocketReturn {
  /** 是否已连接 */
  isConnected: Ref<boolean>;
  /** 最近一条消息（可选，便于调试或展示） */
  lastMessage: Ref<string | null>;
  /** 建立连接 */
  connect: () => void;
  /** 关闭连接 */
  disconnect: () => void;
  /** 发送消息 */
  send: (data: string | ArrayBufferLike | Blob) => void;
}

/**
 * WebSocket 封装：连接、断开、消息收发与生命周期回调。
 * 使用当前页面的 protocol/host，path 会走 Vite 代理（如 /ws -> localhost:8080/ws）。
 */
export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const {
    path = "/ws",
    onopen,
    onmessage,
    onerror,
    onclose,
  } = options;

  const isConnected = ref(false);
  const lastMessage = ref<string | null>(null);
  let ws: WebSocket | null = null;

  function getWsUrl(): string {
    const protocol = location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${location.host}${path}`;
  }

  function connect(): void {
    if (ws != null) return;
    const url = getWsUrl();
    ws = new WebSocket(url);

    ws.onopen = () => {
      isConnected.value = true;
      console.log("[WS] 已连接", url);
      onopen?.();
    };

    ws.onmessage = (event: MessageEvent) => {
      const data = typeof event.data === "string" ? event.data : "";
      lastMessage.value = data;
      console.log("[WS] 收到消息", data);
      onmessage?.(data);
    };

    ws.onerror = (event: Event) => {
      console.error("[WS] 错误", event);
      onerror?.(event);
    };

    ws.onclose = () => {
      isConnected.value = false;
      ws = null;
      console.log("[WS] 已断开");
      onclose?.();
    };
  }

  function disconnect(): void {
    if (ws == null) return;
    ws.close();
    ws = null;
    isConnected.value = false;
  }

  function send(data: string | ArrayBufferLike | Blob): void {
    if (ws == null || ws.readyState !== WebSocket.OPEN) {
      console.warn("[WS] 未连接，无法发送");
      return;
    }
    ws.send(data);
  }

  return {
    isConnected,
    lastMessage,
    connect,
    disconnect,
    send,
  };
}
