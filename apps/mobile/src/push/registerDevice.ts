import type { HttpClient } from "../api/httpClient";

export type MessagingLike = {
  requestPermission(): Promise<boolean>;
  getToken(): Promise<string | null>;
  onTokenRefresh(cb: (token: string) => void): () => void;
};

/** Register FCM device token with backend (DEC-MOBILE-03/04). */
export async function registerDevice(
  http: HttpClient,
  fcmToken: string,
  platform: "ios" | "android",
): Promise<void> {
  const res = await http.request("/v1/devices", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ fcm_token: fcmToken, platform }),
  });
  if (!res.ok) throw new Error("device_register_failed");
}

export async function unregisterDevice(http: HttpClient, fcmToken: string): Promise<void> {
  const res = await http.request("/v1/devices", {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ fcm_token: fcmToken }),
  });
  if (!res.ok && res.status !== 204) throw new Error("device_unregister_failed");
}

/**
 * Ask OS permission, fetch token, POST /v1/devices.
 * Denial is non-fatal — app remains usable without push (DEC-MOBILE- §1 #7).
 */
export async function bootstrapPush(
  http: HttpClient,
  messaging: MessagingLike,
  platform: "ios" | "android",
): Promise<string | null> {
  const allowed = await messaging.requestPermission();
  if (!allowed) return null;
  const token = await messaging.getToken();
  if (!token) return null;
  await registerDevice(http, token, platform);
  return token;
}

export function watchTokenRefresh(
  http: HttpClient,
  messaging: MessagingLike,
  platform: "ios" | "android",
): () => void {
  return messaging.onTokenRefresh((token) => {
    void registerDevice(http, token, platform);
  });
}
