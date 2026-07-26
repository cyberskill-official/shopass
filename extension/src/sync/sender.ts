import { authedFetch, refreshJwt, NoAuthError } from "./auth-bridge";
import * as queue from "./queue";
import { SYNC_URL } from "../shared/config";

export { SYNC_URL };

export const metrics = {
  sent: 0,
  retryCount: 0,
  failClosedCount: 0,
  sentLog: () => { metrics.sent++; },
  retry: () => { metrics.retryCount++; },
  failClosed: () => { metrics.failClosedCount++; }
};

export async function flushQueue(): Promise<void> {
  const items = await queue.peekAll();
  for (const item of items) {
    try {
      const res = await authedFetch(SYNC_URL, item.env);
      if (res.status === 401) {
        await refreshJwt();
        continue; // thử lại sau refresh
      }
      if (res.ok) {
        await queue.ack(item.id);
        metrics.sentLog(); // chỉ ack khi 2xx
      } else {
        metrics.retry(); // 5xx -> giữ lại
      }
    } catch (e) {
      if (e instanceof NoAuthError) {
        metrics.failClosed();
        return; // thiếu JWT -> dừng, giữ hàng đợi (fail-closed)
      }
      metrics.retry(); // lỗi mạng -> giữ lại, backoff
    }
  }
}
