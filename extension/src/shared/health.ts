export interface HealthSignal {
  platform: string;
  broke: string;
  source: string;
}

export function reportHealth(signal: HealthSignal) {
  // Try sending health signal to background
  try {
    chrome.runtime.sendMessage({ type: "HEALTH_SIGNAL", ...signal });
  } catch (err) {
    console.debug("SănDeal: failed to send health signal", err);
  }
}
