import type { HttpClient } from "../api/httpClient";

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
