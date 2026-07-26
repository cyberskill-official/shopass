import type { MemoryAccessToken, SecureStore } from "./tokenStore";
import type { HttpClient } from "../api/httpClient";
import type { PendingReferral } from "../deeplink/pendingReferral";

export interface AuthClientOptions {
  baseURL: string;
  secure: SecureStore;
  access: MemoryAccessToken;
  fetchImpl?: typeof fetch;
  onLogoutDevice?: () => Promise<void>;
  pendingReferral?: PendingReferral;
  http?: HttpClient;
}

export class AuthClient {
  private readonly baseURL: string;
  private readonly secure: SecureStore;
  private readonly access: MemoryAccessToken;
  private readonly fetchImpl: typeof fetch;
  private readonly onLogoutDevice?: () => Promise<void>;
  private readonly pendingReferral?: PendingReferral;
  private readonly http?: HttpClient;

  constructor(opts: AuthClientOptions) {
    this.baseURL = opts.baseURL.replace(/\/$/, "");
    this.secure = opts.secure;
    this.access = opts.access;
    this.fetchImpl = opts.fetchImpl ?? fetch;
    this.onLogoutDevice = opts.onLogoutDevice;
    this.pendingReferral = opts.pendingReferral;
    this.http = opts.http;
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

  async register(email: string, password: string): Promise<void> {
    const body: Record<string, string> = { email, password };
    const pending = this.pendingReferral?.peek();
    if (pending) body.ref = pending;

    const res = await this.fetchImpl(`${this.baseURL}/v1/auth/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!res.ok) throw new Error("register_failed");
    const tokens = (await res.json()) as { access_token: string; refresh_token: string };
    this.access.set(tokens.access_token);
    await this.secure.setRefresh(tokens.refresh_token);

    // Attribute pending referral via BILL-004 after signup (DEC-MOBILE-23).
    if (pending && this.http) {
      await this.http.request("/v1/referral/attribute", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ ref: pending }),
      });
      this.pendingReferral?.clear();
    }
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
    if (this.onLogoutDevice) {
      try {
        await this.onLogoutDevice();
      } catch {
        // still clear local session
      }
    }
    this.access.set(null);
    await this.secure.clearRefresh();
  }
}
