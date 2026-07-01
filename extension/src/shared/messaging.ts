import type { Message } from "./types";

export function sendMessage(msg: Message): Promise<unknown> {
  return chrome.runtime.sendMessage(msg);
}

export function onMessage(
  handler: (msg: Message, sender: chrome.runtime.MessageSender) => void
): void {
  chrome.runtime.onMessage.addListener(handler);
}
