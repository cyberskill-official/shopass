/**
 * Secure token storage contract (Keychain/Keystore in RN).
 * Access token stays in process memory only (DEC-MOBILE-02).
 */
import * as Keychain from "react-native-keychain";

const SERVICE = "shopass.refresh";

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

/** Refresh token in OS secure storage — never AsyncStorage/cleartext. */
export class KeychainSecureStore implements SecureStore {
  async setRefresh(token: string): Promise<void> {
    await Keychain.setGenericPassword("refresh", token, {
      service: SERVICE,
      accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
    });
  }

  async getRefresh(): Promise<string | null> {
    const creds = await Keychain.getGenericPassword({ service: SERVICE });
    return creds ? creds.password : null;
  }

  async clearRefresh(): Promise<void> {
    await Keychain.resetGenericPassword({ service: SERVICE });
  }
}

/** In-memory SecureStore for unit tests. */
export class MemorySecureStore implements SecureStore {
  private refresh: string | null = null;

  async getRefresh(): Promise<string | null> {
    return this.refresh;
  }

  async setRefresh(token: string): Promise<void> {
    this.refresh = token;
  }

  async clearRefresh(): Promise<void> {
    this.refresh = null;
  }
}
