/**
 * @jest-environment node
 */
describe("extension endpoint config", () => {
  const prev = process.env.SHOPASS_ENV;

  afterEach(() => {
    jest.resetModules();
    if (prev === undefined) delete process.env.SHOPASS_ENV;
    else process.env.SHOPASS_ENV = prev;
  });

  it("defaults to production Shopass domains", async () => {
    delete process.env.SHOPASS_ENV;
    const mod = await import("../src/shared/config");
    expect(mod.SYNC_URL).toBe("https://shopass.cyberskill.world/v1/ext/sync");
    expect(mod.WSS_URL).toBe("wss://shopass.cyberskill.world/v1/ext/ws");
    expect(mod.DSAR_URL).toBe("https://shopass.cyberskill.world/dsar");
    expect(mod.SYNC_URL).not.toMatch(/sandeal\.vn/);
  });

  it("development points at local gateway", async () => {
    process.env.SHOPASS_ENV = "development";
    const mod = await import("../src/shared/config");
    expect(mod.SYNC_URL).toBe("http://127.0.0.1:8080/v1/ext/sync");
    expect(mod.WSS_URL).toBe("ws://127.0.0.1:8080/v1/ext/ws");
  });
});
