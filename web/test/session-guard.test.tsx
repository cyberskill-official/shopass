/**
 * @jest-environment jsdom
 */
import React from "react";
import { act, render, screen, waitFor } from "@testing-library/react";
import { SessionGuard } from "../components/session-guard";

jest.mock("../lib/auth", () => ({
  tryRefreshOnce: jest.fn(),
  getAccessToken: jest.fn(),
}));

import { getAccessToken, tryRefreshOnce } from "../lib/auth";

const tryRefreshOnceMock = tryRefreshOnce as jest.MockedFunction<typeof tryRefreshOnce>;
const getAccessTokenMock = getAccessToken as jest.MockedFunction<typeof getAccessToken>;

describe("SessionGuard", () => {
  beforeEach(() => {
    tryRefreshOnceMock.mockReset();
    getAccessTokenMock.mockReset();
  });

  it("renders children when access token is already present", async () => {
    getAccessTokenMock.mockReturnValue("tok");

    render(
      <SessionGuard>
        <div>secure</div>
      </SessionGuard>,
    );

    await waitFor(() => {
      expect(screen.getByText("secure")).toBeTruthy();
    });
    expect(tryRefreshOnceMock).not.toHaveBeenCalled();
  });

  it("refreshes then renders children when access token is restored", async () => {
    getAccessTokenMock.mockReturnValue(null);
    tryRefreshOnceMock.mockResolvedValue("refreshed");

    render(
      <SessionGuard>
        <div>secure</div>
      </SessionGuard>,
    );

    await waitFor(() => {
      expect(screen.getByText("secure")).toBeTruthy();
    });
    expect(tryRefreshOnceMock).toHaveBeenCalledTimes(1);
  });

  it("does not render children when refresh is invalid", async () => {
    getAccessTokenMock.mockReturnValue(null);
    tryRefreshOnceMock.mockResolvedValue("invalid");

    // jsdom cannot navigate; swallow Location.replace not-implemented noise.
    const err = jest.spyOn(console, "error").mockImplementation(() => undefined);

    await act(async () => {
      render(
        <SessionGuard>
          <div>secure</div>
        </SessionGuard>,
      );
    });

    await waitFor(() => {
      expect(tryRefreshOnceMock).toHaveBeenCalledTimes(1);
    });
    expect(screen.queryByText("secure")).toBeNull();
    err.mockRestore();
  });

  it("shows retry UI when refresh is transiently unavailable", async () => {
    getAccessTokenMock.mockReturnValue(null);
    tryRefreshOnceMock.mockResolvedValue("transient");

    render(
      <SessionGuard>
        <div>secure</div>
      </SessionGuard>,
    );

    await waitFor(() => {
      expect(screen.getByText(/Không thể xác thực phiên đăng nhập/)).toBeTruthy();
    });
    expect(screen.queryByText("secure")).toBeNull();
  });
});
