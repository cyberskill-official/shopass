import { onInstalled, onStartup } from "./lifecycle";
import { registerAlarms, onAlarm } from "./alarms";
import { onMessage } from "../shared/messaging";
import { enqueue } from "../sync/queue";
import { flushQueue } from "../sync/sender";
import { openRealtime, closeRealtime } from "../sync/ws-client";

chrome.runtime.onInstalled.addListener(onInstalled);
chrome.runtime.onStartup.addListener(onStartup);
chrome.alarms.onAlarm.addListener(onAlarm);
chrome.runtime.onMessage.addListener(onMessage);

chrome.runtime.onMessage.addListener((msg, sender, sendResponse) => {
  if (msg.type === "PIPELINE_DONE") {
    enqueue({
      payload: msg.payload,
      clientTs: Date.now()
    }).then(() => flushQueue());
  } else if (msg.type === "START_REALTIME") {
    openRealtime();
  } else if (msg.type === "STOP_REALTIME") {
    closeRealtime();
  }
});

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === "sandeal_sync_flush") {
    flushQueue();
  }
});

registerAlarms();

