const store = new Map<string, { username: string; password: string }>();

export const ACCESSIBLE = {
  WHEN_UNLOCKED_THIS_DEVICE_ONLY: "WHEN_UNLOCKED_THIS_DEVICE_ONLY",
};

export async function setGenericPassword(
  username: string,
  password: string,
  opts?: { service?: string },
): Promise<boolean> {
  store.set(opts?.service ?? "default", { username, password });
  return true;
}

export async function getGenericPassword(opts?: {
  service?: string;
}): Promise<false | { username: string; password: string }> {
  const v = store.get(opts?.service ?? "default");
  return v ?? false;
}

export async function resetGenericPassword(opts?: { service?: string }): Promise<boolean> {
  store.delete(opts?.service ?? "default");
  return true;
}

export function __resetMockKeychain(): void {
  store.clear();
}
