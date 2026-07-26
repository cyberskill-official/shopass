import type { MemoryAccessToken, SecureStore } from "./tokenStore";

export interface AuthClientOptions {
  baseURL: string;
  secure: SecureStore;
  access: MemoryAccessToken;
  fetchImpl?: typeof fetch;
}

export class AuthClient {
  private readonly baseURL: string;
  private readonly secure: SecureStore;
  private readonly access: MemoryAccessToken;
  private readonly fetchImpl: typeof fetch;

  constructor(opts: AuthClientOptions) {
    this.baseURL = opts.baseURL.replace(/\/$/, "");
    this.secure = opts.secure;
    this.access = opts.access;
    this.fetchImpl = opts.fetchImpl ?? fetch;
  }

  async login(email: string, password: string): Promise<void> {
    const res = await this.fetchImpl(`${this.baseURL}/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) throw new Error("login_failed");
    const body = (await res.json()) as { access_token: string; refresh_token: string };
    this.access.set(body.access_token);
    await this.secure.setRefresh(body.refresh_token);
  }

  async refreshOnce(): Promise<boolean> {
    const refresh = await this.secure.getRefresh();
    if (!refresh) return false;
    const res = await this.fetchImpl(`${this.baseURL}/v1/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refresh }),
    });
    if (!res.ok) {
      this.access.set(null);
      await this.secure.clearRefresh();
      return false;
    }
    const body = (await res.json()) as { access_token: string; refresh_token?: string };
    this.access.set(body.access_token);
    if (body.refresh_token) await this.secure.setRefresh(body.refresh_token);
    return true;
  }

  async logout(): Promise<void> {
    this.access.set(null);
    await this.secure.clearRefresh();
  }
}
