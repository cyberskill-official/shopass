import { ConsentPurpose, ConsentRecord } from "./types";

const CONSENT_KEY = "sandeal:consent";

export async function getConsent(): Promise<ConsentRecord | null> {
  const obj = await chrome.storage.local.get(CONSENT_KEY);
  return (obj[CONSENT_KEY] as ConsentRecord) ?? null;
}

export async function setConsent(purposes: ConsentPurpose[]): Promise<void> {
  const record: ConsentRecord = {
    policyVersion: "2026-06-27",
    decidedAt: Date.now(),
    granted: purposes
  };
  await chrome.storage.local.set({ [CONSENT_KEY]: record });
}

export async function ensureConsent(purpose: ConsentPurpose): Promise<boolean> {
  const consent = await getConsent();
  if (!consent) return false;
  return consent.granted.includes(purpose);
}
