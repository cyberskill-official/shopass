import type { Message } from "./types";

export function sendMessage(msg: Message): Promise<any> {
  return chrome.runtime.sendMessage(msg);
}

export function onMessage(message: any, sender: chrome.runtime.MessageSender, sendResponse: (response?: any) => void): boolean | void {
  const msg = message as Message;
  if (msg.type === "PING") {
    sendResponse({ pong: true });
    return false;
  }
  if (msg.type === "CART_READ") {
    // Process cart read
    sendResponse({ success: true });
    return false;
  }
  return false;
}
