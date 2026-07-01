import type { Message } from "./types";

export function sendMessage(msg: Message): Promise<any> {
  return chrome.runtime.sendMessage(msg);
}

import { ensureConsent } from "../consent/consent-gate";

export function onMessage(message: any, sender: chrome.runtime.MessageSender, sendResponse: (response?: any) => void): boolean | void {
  const msg = message as Message;
  if (msg.type === "PING") {
    sendResponse({ pong: true });
    return false;
  }
  if (msg.type === "CART_READ") {
    // Defense-in-depth: double check consent before processing
    ensureConsent("sync_backend").then(granted => {
      if (!granted) {
        sendResponse({ success: false, reason: "no_consent" });
        return;
      }
      // Process cart read
      sendResponse({ success: true });
    });
    return true; // indicates async response
  }
  return false;
}
