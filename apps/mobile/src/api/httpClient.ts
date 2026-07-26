import type { AuthClient } from "../auth/authClient";
import type { MemoryAccessToken } from "../auth/tokenStore";

export class HttpClient {
  constructor(
    private readonly baseURL: string,
    private readonly access: MemoryAccessToken,
    private readonly auth: AuthClient,
    private readonly fetchImpl: typeof fetch = fetch,
  ) {}

  async request(path: string, init: RequestInit = {}): Promise<Response> {
    const headers = new Headers(init.headers);
    const token = this.access.get();
    if (token) headers.set("Authorization", `Bearer ${token}`);
    let res = await this.fetchImpl(`${this.baseURL.replace(/\/$/, "")}${path}`, {
      ...init,
      headers,
    });
    if (res.status === 401) {
      const ok = await this.auth.refreshOnce();
      if (!ok) throw new Error("unauthorized");
      const retryHeaders = new Headers(init.headers);
      const next = this.access.get();
      if (next) retryHeaders.set("Authorization", `Bearer ${next}`);
      res = await this.fetchImpl(`${this.baseURL.replace(/\/$/, "")}${path}`, {
        ...init,
        headers: retryHeaders,
      });
    }
    return res;
  }
}
