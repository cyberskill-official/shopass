const store: Record<string, unknown> = {};

export function fakeChromeStorage() {
  return {
    storage: {
      local: {
        get: async (key: string) => ({ [key]: store[key] }),
        set: async (items: Record<string, unknown>) =>
          Object.assign(store, items),
      },
    },
  } as unknown as typeof chrome;
}
