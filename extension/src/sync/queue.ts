import { SyncEnvelope } from "../shared/types";

export interface QueueItem {
  id: string;
  env: SyncEnvelope;
}

const QUEUE_KEY = "sandeal:queue";

// Helper để giả lập storage trong test nếu không có chrome.storage
let memoryStorage: { [key: string]: any } = {};

async function getStorage(key: string): Promise<any> {
  if (typeof chrome !== "undefined" && chrome.storage && chrome.storage.local) {
    return new Promise((resolve) => {
      chrome.storage.local.get([key], (res) => resolve(res[key]));
    });
  }
  return memoryStorage[key];
}

async function setStorage(key: string, val: any): Promise<void> {
  if (typeof chrome !== "undefined" && chrome.storage && chrome.storage.local) {
    return new Promise((resolve) => {
      chrome.storage.local.set({ [key]: val }, resolve);
    });
  }
  memoryStorage[key] = val;
}

export async function peekAll(): Promise<QueueItem[]> {
  const data = await getStorage(QUEUE_KEY);
  return data || [];
}

export async function enqueue(env: SyncEnvelope): Promise<void> {
  const items = await peekAll();
  items.push({
    id: Math.random().toString(36).substring(2, 15),
    env
  });
  await setStorage(QUEUE_KEY, items);
}

export async function ack(id: string): Promise<void> {
  const items = await peekAll();
  const filtered = items.filter(i => i.id !== id);
  await setStorage(QUEUE_KEY, filtered);
}
