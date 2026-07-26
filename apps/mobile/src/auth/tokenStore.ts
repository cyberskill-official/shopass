/** Secure token storage contract (Keychain/Keystore in RN). Access token stays in memory. */

export interface SecureStore {
  getRefresh(): Promise<string | null>;
  setRefresh(token: string): Promise<void>;
  clearRefresh(): Promise<void>;
}

export class MemoryAccessToken {
  private access: string | null = null;

  get(): string | null {
    return this.access;
  }

  set(token: string | null): void {
    this.access = token;
  }
}
