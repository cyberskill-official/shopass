import type { Message } from "./types";
import { ensureConsent } from "../consent/consent-gate";
import { minimize } from "../pipeline/minimize";
import { enqueue } from "../sync/queue";
import { flushQueue } from "../sync/sender";
import { openRealtime, closeRealtime } from "../sync/ws-client";

export function sendMessage(msg: Message): Promise<any> {
  return chrome.runtime.sendMessage(msg);
}

/** Host patterns matching manifest host_permissions (content-script origin only). */
const ALLOWED_CART_HOST_RE =
  /^(https:\/\/shopee\.vn\/|https:\/\/([a-z0-9-]+\.)*tiktok\.com\/|https:\/\/www\.lazada\.vn\/)/i;

/**
 * CART_READ must come from a content script on an allowlisted shop host.
 * Reject extension pages / external senders (defense in depth).
 */
export function isTrustedCartSender(sender: chrome.runtime.MessageSender): boolean {
  if (sender.id && typeof chrome !== "undefined" && chrome.runtime?.id) {
    if (sender.id !== chrome.runtime.id) return false;
  }
  const url = sender.tab?.url ?? sender.url;
  if (!url || typeof url !== "string") return false;
  return ALLOWED_CART_HOST_RE.test(url);
}

export function onMessage(
  message: any,
  sender: chrome.runtime.MessageSender,
  sendResponse: (response?: any) => void
): boolean | void {
  const msg = message as Message & { type?: string };

  if (msg?.type === "PING") {
    sendResponse({ pong: true });
    return false;
  }

  if (msg?.type === "START_REALTIME") {
    openRealtime();
    sendResponse({ ok: true });
    return false;
  }

  if (msg?.type === "STOP_REALTIME") {
    closeRealtime();
    sendResponse({ ok: true });
    return false;
  }

  if (msg?.type === "CART_READ") {
    handleCartRead(msg as Extract<Message, { type: "CART_READ" }>, sender)
      .then(sendResponse)
      .catch((err) => {
        sendResponse({
          success: false,
          reason: "handler_error",
          error: err instanceof Error ? err.message : String(err),
        });
      });
    return true; // async response
  }

  return false;
}

async function handleCartRead(
  msg: Extract<Message, { type: "CART_READ" }>,
  sender: chrome.runtime.MessageSender
): Promise<{ success: boolean; reason?: string }> {
  if (!isTrustedCartSender(sender)) {
    return { success: false, reason: "untrusted_sender" };
  }

  if (!(await ensureConsent("sync_backend"))) {
    return { success: false, reason: "no_consent" };
  }

  // Live path: CART_READ → minimize → queue → flush (PIPELINE_DONE removed).
  const payload = minimize({
    type: "CART_READ",
    platform: msg.platform,
    items: msg.items as unknown as Array<Record<string, unknown>>,
    vouchers: msg.vouchers as unknown as Array<Record<string, unknown>>,
  });

  if (!payload) {
    return { success: false, reason: "minimize_rejected" };
  }

  await enqueue({
    payload,
    clientTs: Date.now(),
  });
  await flushQueue();
  return { success: true };
}
