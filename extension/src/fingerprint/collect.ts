import { hashDeviceSignals } from "./hash";

/**
 * Collect coarse, non-PII signals and return a single client hash.
 * Raw UA / fonts / canvas pixels MUST NOT leave the device.
 */
export async function collectDeviceHash(): Promise<string> {
  const coarse = [
    typeof navigator !== "undefined" ? String(navigator.language || "") : "",
    typeof screen !== "undefined" ? `${screen.width}x${screen.height}` : "",
    typeof Intl !== "undefined"
      ? Intl.DateTimeFormat().resolvedOptions().timeZone || ""
      : "",
  ];
  return hashDeviceSignals(coarse);
}
