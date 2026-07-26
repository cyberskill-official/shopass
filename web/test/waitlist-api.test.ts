import { submitWaitlist } from "../lib/billing/waitlist";

describe("waitlist client", () => {
  const originalFetch = global.fetch;

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it("posts to /api/waitlist", async () => {
    const fetchMock = jest.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ ok: true, id: 3 }),
    } as Response);
    global.fetch = fetchMock;

    await expect(
      submitWaitlist({
        email: "a@b.co",
        source: "pricing",
        tier_interest: "premium_pro",
      }),
    ).resolves.toEqual({ ok: true, id: 3 });

    expect(fetchMock).toHaveBeenCalledWith("/api/waitlist", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        email: "a@b.co",
        source: "pricing",
        tier_interest: "premium_pro",
      }),
    });
  });

  it("surfaces upstream error", async () => {
    global.fetch = jest.fn().mockResolvedValue({
      ok: false,
      json: async () => ({ error: "invalid email" }),
    } as Response);

    await expect(submitWaitlist({ email: "nope" })).rejects.toThrow("invalid email");
  });
});
