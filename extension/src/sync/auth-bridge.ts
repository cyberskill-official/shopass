import { SyncEnvelope } from "../shared/types";

export class NoAuthError extends Error {
  constructor() {
    super("Missing SănDeal JWT");
    this.name = "NoAuthError";
  }
}

let memoryJwt: string | undefined = undefined;

// In a real environment, this gets the JWT from chrome.storage.session
export async function getJwt(): Promise<string | undefined> {
  if (typeof chrome !== "undefined" && chrome.storage && chrome.storage.session) {
    return new Promise((resolve) => {
      chrome.storage.session.get(["sandeal_jwt"], (res: any) => resolve(res.sandeal_jwt));
    });
  }
  return memoryJwt;
}

export async function setJwt(jwt: string | undefined): Promise<void> {
  if (typeof chrome !== "undefined" && chrome.storage && chrome.storage.session) {
    if (jwt === undefined) {
      return new Promise((resolve) => chrome.storage.session.remove(["sandeal_jwt"], resolve));
    }
    return new Promise((resolve) => {
      chrome.storage.session.set({ sandeal_jwt: jwt }, resolve);
    });
  }
  memoryJwt = jwt;
}

export async function refreshJwt(): Promise<void> {
  // Logic to refresh JWT using refresh token (from FR-AUTH-002)
  // Mocked for now
}

export async function authedFetch(url: string, env: SyncEnvelope): Promise<Response> {
  const jwt = await getJwt(); // storage.session
  if (!jwt) throw new NoAuthError(); // fail-closed (DEC-EXT-25)

  return fetch(url, {
    method: "POST",
    headers: {
      "Authorization": `Bearer ${jwt}`, // JWT SănDeal, KHÔNG token sàn
      "Content-Type": "application/json"
    },
    body: JSON.stringify(env) // body chỉ OutboundPayload đã sạch + clientTs
  });
}
