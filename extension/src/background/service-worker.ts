import { onInstalled, onStartup } from "./lifecycle";
import { registerAlarms, onAlarm } from "./alarms";
import { onMessage } from "../shared/messaging";
import { flushQueue } from "../sync/sender";

chrome.runtime.onInstalled.addListener(onInstalled);
chrome.runtime.onStartup.addListener(onStartup);
chrome.alarms.onAlarm.addListener(onAlarm);

// Single message listener: CART_READ → minimize → enqueue/flush (see messaging.ts).
chrome.runtime.onMessage.addListener(onMessage);

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === "sandeal_sync_flush") {
    flushQueue();
  }
});

registerAlarms();
