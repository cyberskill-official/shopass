import { WSS_URL } from "../shared/config";

export { WSS_URL };

let ws: WebSocket | null = null;

export function openRealtime(): void {
  if (ws) return; // không mở trùng
  // @ts-ignore
  ws = new WebSocket(WSS_URL); // CHỈ khi cần realtime
  ws.onclose = () => {
    ws = null;
  };
}

export function closeRealtime(): void {
  ws?.close(); // đóng khi xong; KHÔNG mở thường trực
  ws = null;
}

export function isWsOpen(): boolean {
  return ws !== null && ws.readyState === 1; // 1 is OPEN
}
