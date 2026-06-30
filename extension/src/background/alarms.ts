import { getState, setState } from "../shared/storage";

export const TICK = "sandeal:tick";

export function registerAlarms(): void {
  // 0.5 minutes = 30 seconds
  chrome.alarms.create(TICK, { periodInMinutes: 0.5 });
}

export async function onAlarm(alarm: chrome.alarms.Alarm): Promise<void> {
  if (alarm.name !== TICK) return;
  const state = await getState();
  await setState({ ...state, lastSyncAt: Date.now() });
}
