/**
 * @jest-environment jsdom
 */

import { openAffiliate } from "../../src/guardrails/user-gesture";

test("mở affiliate không user gesture -> ném lỗi", () => {
  expect(() => openAffiliate(undefined, "https://go.aff.x")).toThrow(/trusted user gesture/);
});

test("untrusted event -> ném lỗi", () => {
  // @ts-ignore
  expect(() => openAffiliate({ isTrusted: false } as any, "https://go.aff.x")).toThrow();
});

test("trusted event -> gọi window.open", () => {
    const originalOpen = window.open;
    const mockOpen = jest.fn();
    window.open = mockOpen;
    try {
        // @ts-ignore
        openAffiliate({ isTrusted: true } as any, "https://go.aff.x");
        expect(mockOpen).toHaveBeenCalledWith("https://go.aff.x", "_blank", "noopener");
    } finally {
        window.open = originalOpen;
    }
});
