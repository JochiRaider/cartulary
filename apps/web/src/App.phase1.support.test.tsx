import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./WorkbookShell", () => ({
  WorkbookShell: vi.fn(),
  buildCreatePayload: vi.fn(),
  createDraftRow: vi.fn(),
  ensureDraftRow: vi.fn(),
  TimelineWorkbook: vi.fn(),
}));

import { AppRoot } from "./AppRoot";

describe("Phase 1 ordinary shell support", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    window.history.replaceState({}, "", "/");
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("shows the anonymous login surface when the browser has no session", async () => {
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse(
            {
              error: {
                code: "session_required",
                status: 401,
                details: {
                  reason_code: "no_session",
                },
              },
            },
            401,
          ),
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
    });

    render(<AppRoot />);

    expect(await screen.findByTestId("auth-login-username")).toBeTruthy();
    expect(screen.getByTestId("auth-shell-message").textContent).toContain(
      "Sign in with your local account",
    );
    await waitFor(() => {
      expect(screen.getByTestId("auth-status").textContent).toBe(
        "Ready to sign in.",
      );
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });
});

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
