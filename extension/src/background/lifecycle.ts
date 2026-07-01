import { getState, setState, defaultState } from "../shared/storage";

export async function onInstalled(details: chrome.runtime.InstalledDetails): Promise<void> {
  if (details.reason === "install") {
    await setState(defaultState());
    chrome.tabs.create({ url: "ui/onboarding.html" });
  }
}

export async function onStartup(): Promise<void> {
  const state = await getState();
  // rehydrate if needed
}
