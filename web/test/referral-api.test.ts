import {
  attributeReferral,
  capturePendingReferral,
  clearPendingReferral,
  readPendingReferral,
} from "../lib/referral/api";
import { apiFetch } from "../lib/api";

jest.mock("../lib/api", () => ({
  apiFetch: jest.fn(),
}));

describe("referral client", () => {
  const apiFetchMock = apiFetch as jest.MockedFunction<typeof apiFetch>;

  beforeEach(() => {
    jest.clearAllMocks();
    clearPendingReferral();
  });

  it("stores pending ref from share links", () => {
    capturePendingReferral("sdab12");
    expect(readPendingReferral()).toBe("SDAB12");
  });

  it("maps self_referral errors", async () => {
    apiFetchMock.mockResolvedValue({
      ok: false,
      json: async () => ({ error: "self_referral" }),
    } as Response);
    await expect(attributeReferral("SDSELF")).rejects.toThrow(/tự giới thiệu/i);
  });
});
