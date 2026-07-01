export interface ExtState {
  schemaVersion: number;
  installedAt: number;
  lastSyncAt?: number;
  pendingReads: string[];
}

const KEY = "sandeal:state";

export async function getState(): Promise<ExtState> {
  const obj = await chrome.storage.local.get(KEY);
  return (obj[KEY] as ExtState) ?? defaultState();
}

export async function setState(next: ExtState): Promise<void> {
  await chrome.storage.local.set({ [KEY]: next });
}

export function defaultState(): ExtState {
  return { schemaVersion: 1, installedAt: Date.now(), pendingReads: [] };
}
