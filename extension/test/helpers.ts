export function fakeChromeStorage() {
  const store = new Map<string, any>();
  return {
    local: {
      get: jest.fn(async (keys: string | string[]) => {
        if (typeof keys === "string") {
          return store.has(keys) ? { [keys]: store.get(keys) } : {};
        }
        return {};
      }),
      set: jest.fn(async (items: any) => {
        for (const [k, v] of Object.entries(items)) {
          store.set(k, v);
        }
      }),
      remove: jest.fn(async (keys: string | string[]) => {
        if (typeof keys === "string") store.delete(keys);
      })
    }
  };
}
