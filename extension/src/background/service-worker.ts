import { onInstalled, onStartup } from "./lifecycle";
import { registerAlarms, onAlarm } from "./alarms";
import { onMessage } from "../shared/messaging";

chrome.runtime.onInstalled.addListener(onInstalled);
chrome.runtime.onStartup.addListener(onStartup);
chrome.alarms.onAlarm.addListener(onAlarm);
chrome.runtime.onMessage.addListener(onMessage);

registerAlarms();
