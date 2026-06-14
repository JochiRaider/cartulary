import {
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
  phase1LandingTestId,
  phase1RouteTestId,
} from "@cartulary/ui-contracts";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../workbook/WorkbookShell", () => ({
  WorkbookShell: vi.fn(),
  buildCreatePayload: vi.fn(),
  createDraftRow: vi.fn(),
  ensureDraftRow: vi.fn(),
  TimelineWorkbook: vi.fn(),
}));

import {
  credentialStateResource,
  installLandingShellFetch,
  sessionResource,
} from "../testing/appShellTestSupport";
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
      if (String(input) === "/api/v1/auth/providers") {
        return Promise.resolve(jsonResponse({ data: { providers: [] } }));
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
    });

    render(<AppRoot />);

    expect(
      await screen.findByTestId(phase1AuthTestId("login-username")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1AuthTestId("shell-message")).textContent,
    ).toContain("Sign in with your local account");
    await waitFor(() => {
      expect(screen.getByTestId(phase1AuthTestId("status")).textContent).toBe(
        "Ready to sign in.",
      );
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("FE-S-P1-01 Verify bootstrap route selectors and error-state selectors use stable test-id builders.", async () => {
    installLandingShellFetch(fetchMock, {
      credentialState: credentialStateResource(),
      incidents: [],
      session: sessionResource({
        display_name: "Phase 1 Support Operator",
      }),
    });

    render(<AppRoot />);

    expect(
      await screen.findByTestId(phase1RouteTestId("app-shell")),
    ).toBeTruthy();
    expect(screen.getByTestId(phase1LandingTestId("shell"))).toBeTruthy();
    expect(screen.getByTestId(phase1LandingTestId("status"))).toBeTruthy();
    expect(
      screen.getByTestId(phase1AccountTestId("session-user-id")).textContent,
    ).not.toBe("");
    expect(screen.getByTestId(phase1AdminTestId("access-note"))).toBeTruthy();
    expect(screen.getByTestId(phase1ErrorCodeTestId("landing"))).toBeTruthy();
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("landing").container),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("account").message),
    ).toBeTruthy();

    cleanup();
    fetchMock.mockReset();
    installAnonymousSessionRequiredFetch(fetchMock);
    window.history.replaceState({}, "", "/");
    render(<AppRoot />);

    expect(
      await screen.findByTestId(phase1AuthTestId("login-username")),
    ).toBeTruthy();
    expect(screen.getByTestId(phase1AuthTestId("shell-message"))).toBeTruthy();
    expect(screen.getByTestId(phase1AuthTestId("status"))).toBeTruthy();
    expect(screen.getByTestId(phase1ErrorCodeTestId("auth"))).toBeTruthy();
    expect(
      screen.getByTestId(phase1ErrorSummaryTestIds("auth").container),
    ).toBeTruthy();
  });
});

function installAnonymousSessionRequiredFetch(
  fetchMock: ReturnType<typeof vi.fn>,
) {
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
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
