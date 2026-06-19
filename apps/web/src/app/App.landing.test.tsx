import {
  deploymentUserRowTestId,
  landingAdminMenuItemTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1LandingTestId,
  phase1RouteTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("../workbook/WorkbookShell", async () => {
  const React = await import("react");

  return {
    WorkbookShell: ({
      incidentId,
      onIncidentAccessLost,
    }: {
      incidentId: string;
      onIncidentAccessLost?: () => void;
    }) => {
      const [currentRole, setCurrentRole] = React.useState("");

      React.useEffect(() => {
        let active = true;
        void fetch("/api/v1/auth/session")
          .then(async (response) => {
            const payload = (await response.json()) as {
              data?: {
                memberships?: Array<{
                  incident_id: string;
                  role: string;
                }>;
              };
            };
            if (!active) {
              return;
            }
            const membership =
              payload.data?.memberships?.find(
                (entry) => entry.incident_id === incidentId,
              ) ?? null;
            setCurrentRole(membership?.role ?? "");
          })
          .catch(() => {
            if (active) {
              setCurrentRole("");
            }
          });

        return () => {
          active = false;
        };
      }, [incidentId]);

      return (
        <section data-testid="mock-workbook">
          <p data-testid="mock-workbook-incident">{incidentId}</p>
          <p data-testid="mock-workbook-role">{currentRole}</p>
          <button
            data-testid="mock-access-lost"
            type="button"
            onClick={() => {
              onIncidentAccessLost?.();
            }}
          >
            Lose access
          </button>
        </section>
      );
    },
    buildCreatePayload: vi.fn(),
    createDraftRow: vi.fn(),
    ensureDraftRow: vi.fn(),
    TimelineWorkbook: vi.fn(),
  };
});

vi.mock("./Phase1Harness", () => ({
  Phase1Harness: () => <section data-testid="mock-phase1-harness" />,
}));

vi.mock("./Phase2Harness", () => ({
  Phase2Harness: () => <section data-testid="mock-phase2-harness" />,
}));

import {
  incidentResource,
  installLandingShellFetch,
  sessionResource,
} from "../testing/appShellTestSupport";
import {
  abortablePendingResponse,
  expectStableFetchCount,
  jsonResponse,
} from "../testing/fetchMockTestSupport";
import { AppRoot } from "./AppRoot";

describe("Incident landing", () => {
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

  it("shows the zero-incident landing state", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Bootstrap Admin",
        is_deployment_admin: true,
      }),
    });

    renderApp();

    expect(
      await screen.findByTestId(phase1LandingTestId("empty-state")),
    ).toBeTruthy();
    const adminShell = screen.getByTestId(landingAdminShellTestId("shell"));
    expect(adminShell).toBeTruthy();
    expect((adminShell as HTMLElement).style.width).toBe("100%");
    expect((adminShell as HTMLElement).style.margin).toBe("");
    expect((adminShell as HTMLElement).style.borderRadius).toBe("");
    expect(window.getComputedStyle(document.body).margin).toBe("0px");
    expect(
      Array.from(document.querySelectorAll("style")).some((style) =>
        Boolean(
          style.textContent?.includes("#root") &&
            style.textContent.includes("margin: 0") &&
            style.textContent.includes("--ct-app-viewport-block-size: 100vh") &&
            style.textContent.includes("@supports (height: 100dvh)") &&
            style.textContent.includes(
              "scroll-margin-block-start: var(--cartulary-grid-scroll-margin-block-start)",
            ),
        ),
      ),
    ).toBe(true);
    expect(
      screen.getByTestId(landingAdminShellTestId("menu")).getAttribute("role"),
    ).toBe(null);
    expect(
      screen
        .getByTestId(landingAdminMenuItemTestId("incidents"))
        .getAttribute("aria-pressed"),
    ).toBe("true");
    expect(
      screen
        .getByTestId(landingAdminPanelTestId("incidents"))
        .getAttribute("role"),
    ).toBe(null);
    expect(screen.queryByText("Open selected")).toBe(null);
    expect(
      screen.getByTestId(phase1LandingTestId("incidents-count")).textContent,
    ).toBe("0");
    expect(
      screen.getByTestId(phase1LandingTestId("current-user")).textContent,
    ).toBe("Bootstrap Admin · deployment admin");
    await expectStableFetchCount(fetchMock, 3);
  });

  it("switches menu panels and leaves actions inside each panel", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
    });

    renderApp();

    await screen.findByTestId(phase1LandingTestId("empty-state"));
    fireEvent.click(
      screen.getByTestId(landingAdminMenuItemTestId("account-security")),
    );
    expect(
      screen
        .getByTestId(landingAdminMenuItemTestId("account-security"))
        .getAttribute("aria-pressed"),
    ).toBe("true");
    expect(
      screen
        .getByTestId(landingAdminPanelTestId("account-security"))
        .getAttribute("aria-labelledby"),
    ).toBe(landingAdminMenuItemTestId("account-security"));
    expect(
      screen.getByTestId(phase1AccountTestId("refresh-state")),
    ).toBeTruthy();
    expect(screen.getByTestId(phase1AccountTestId("logout"))).toBeTruthy();

    fireEvent.keyDown(screen.getByTestId(landingAdminShellTestId("menu")), {
      key: "End",
    });
    await waitFor(() => {
      expect(
        screen
          .getByTestId(landingAdminMenuItemTestId("reference-packs"))
          .getAttribute("aria-pressed"),
      ).toBe("true");
    });
    expect(document.body.textContent).toContain(
      "Deployment admin access is required for reference-pack",
    );

    fireEvent.click(
      screen.getByTestId(landingAdminMenuItemTestId("deployment-users")),
    );
    expect(
      screen.getByTestId(phase1AdminTestId("access-note")).textContent,
    ).toContain("Deployment admin access is required");
    expect(screen.queryByTestId(phase1AdminTestId("patch-user"))).toBe(null);
  });

  it("enables deployment-user panel target actions after selecting a loaded user", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Deployment Admin",
        is_deployment_admin: true,
      }),
      extraRoutes: [
        {
          method: "GET",
          url: "/api/v1/users?limit=100",
          handler: () =>
            jsonResponse({
              data: {
                users: [
                  {
                    user_id: "user-2",
                    email: "target@example.test",
                    display_name: "Target User",
                    user_version: 3,
                    is_active: true,
                    mfa_required: true,
                    is_deployment_admin: false,
                  },
                ],
              },
              meta: {
                paging: {
                  limit: 100,
                  has_more: false,
                  next_cursor: null,
                },
                request_id: "req-users",
              },
            }),
        },
      ],
    });

    renderApp();

    await screen.findByTestId(phase1LandingTestId("empty-state"));
    fireEvent.click(
      screen.getByTestId(landingAdminMenuItemTestId("deployment-users")),
    );
    expect(
      (screen.getByTestId(phase1AdminTestId("patch-user")) as HTMLButtonElement)
        .disabled,
    ).toBe(true);
    const userRow = await screen.findByTestId(
      deploymentUserRowTestId("user-2"),
    );
    fireEvent.click(userRow);
    await waitFor(() => {
      expect(
        screen.getByTestId(phase1AdminTestId("target-user-id")).textContent,
      ).toBe("user-2");
    });
    expect(
      (screen.getByTestId(phase1AdminTestId("patch-user")) as HTMLButtonElement)
        .disabled,
    ).toBe(false);
    expect(
      (
        screen.getByTestId(
          phase1AdminTestId("password-reset"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(false);
    expect(
      (screen.getByTestId(phase1AdminTestId("revoke-all")) as HTMLButtonElement)
        .disabled,
    ).toBe(false);
  });

  it("renders one or many visible incidents on the landing screen", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: [
        incidentResource("incident-1", "IR-201", "First Incident"),
        incidentResource("incident-2", "IR-202", "Second Incident"),
      ],
    });

    renderApp();

    expect(
      await screen.findByTestId(phase1LandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("incidents-count")).textContent,
    ).toBe("2");
    expect(
      screen.getByTestId(landingIncidentCardTestId("incident-1")).textContent,
    ).toContain("First Incident");
    expect(
      screen.getByTestId(landingIncidentCardTestId("incident-2")).textContent,
    ).toContain("Second Incident");
    expect(
      screen
        .getByTestId(landingIncidentCardTestId("incident-1"))
        .getAttribute("data-selected"),
    ).toBe(null);
    fireEvent.click(
      screen.getByTestId(landingIncidentOpenButtonTestId("incident-2")),
    );

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "incident-2",
    );
  });

  it("Phase 2 U-2-11 ordinary landing shell creates an incident, refreshes session-visible membership, routes to the workbook by incident_id, and falls back when a stale incident selection is no longer visible", async () => {
    let created = false;
    installLandingShellFetch(fetchMock, {
      session: () =>
        sessionResource({
          display_name: "Operator",
          memberships: created
            ? [
                {
                  incident_id: "incident-created",
                  role: "admin",
                },
              ]
            : [],
        }),
      incidents: () =>
        created
          ? [incidentResource("incident-created", "IR-203", "Created Incident")]
          : [],
      onCreateIncident: () => {
        created = true;
        return incidentResource(
          "incident-created",
          "IR-203",
          "Created Incident",
        );
      },
    });

    renderApp();

    await screen.findByTestId(phase1LandingTestId("empty-state"));
    fireEvent.change(screen.getByTestId(phase1LandingTestId("incident-key")), {
      target: { value: "IR-203" },
    });
    fireEvent.change(
      screen.getByTestId(phase1LandingTestId("incident-title")),
      {
        target: { value: "Created Incident" },
      },
    );
    fireEvent.click(screen.getByTestId(phase1LandingTestId("create-button")));

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "incident-created",
    );
    await waitFor(() => {
      expect(screen.getByTestId("mock-workbook-role").textContent).toBe(
        "admin",
      );
    });
    expect(window.location.search).toContain("incident_id=incident-created");

    cleanup();
    window.history.replaceState({}, "", "/?incident_id=incident-stale");
    fetchMock.mockReset();
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
      incidents: [incidentResource("incident-live", "IR-204", "Live Incident")],
    });

    renderApp();

    expect(
      await screen.findByTestId(phase1LandingTestId("incident-list")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(landingIncidentOpenButtonTestId("incident-live")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("status")).textContent?.trim(),
    ).not.toBe("");
    expect(window.location.search).not.toContain("incident_id=");
  });

  it("renders the workbook directly from an authenticated incident route under StrictMode", async () => {
    window.history.replaceState({}, "", "/?incident_id=incident-5");
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
        memberships: [
          {
            incident_id: "incident-5",
            role: "admin",
          },
        ],
      }),
      incidents: [incidentResource("incident-5", "IR-205", "Visible Incident")],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    expect(screen.getByTestId("mock-workbook-incident").textContent).toBe(
      "incident-5",
    );
    const appShell = screen.getByTestId(phase1RouteTestId("app-shell"));
    expect(appShell.style.blockSize).toBe("var(--ct-app-viewport-block-size)");
    expect(appShell.style.overflow).toBe("hidden");
    expect(["0", "0px"]).toContain(appShell.style.minBlockSize);
    expect(["0", "0px"]).toContain(appShell.style.minHeight);
    const workbookFrame = screen.getByTestId("mock-workbook").parentElement;
    expect(workbookFrame).toBeInstanceOf(HTMLElement);
    expect((workbookFrame as HTMLElement).style.display).toBe("grid");
    expect((workbookFrame as HTMLElement).style.blockSize).toBe("100%");
    expect((workbookFrame as HTMLElement).style.overflow).toBe("hidden");
    await expectStableFetchCount(fetchMock, 5);
  });

  it("returns to the landing screen when workbook access is lost", async () => {
    window.history.replaceState({}, "", "/?incident_id=incident-5");
    let accessLost = false;
    installLandingShellFetch(fetchMock, {
      session: () =>
        sessionResource({
          display_name: "Operator",
          memberships: accessLost
            ? []
            : [
                {
                  incident_id: "incident-5",
                  role: "admin",
                },
              ],
        }),
      incidents: () =>
        accessLost
          ? []
          : [incidentResource("incident-5", "IR-205", "Visible Incident")],
    });

    renderApp();

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    await expectStableFetchCount(fetchMock, 5);
    accessLost = true;
    fireEvent.click(screen.getByTestId("mock-access-lost"));

    await waitFor(() => {
      expect(
        screen.getByTestId(phase1LandingTestId("empty-state")),
      ).toBeTruthy();
    });
    expect(
      screen.getByTestId(phase1LandingTestId("status")).textContent?.trim(),
    ).not.toBe("");
    expect(window.location.search).not.toContain("incident_id=");
    await expectStableFetchCount(fetchMock, 8);
  });

  it("cancels an in-flight shell refresh when the app unmounts", async () => {
    const abortedSignals: AbortSignal[] = [];
    fetchMock.mockImplementation((input, init) => {
      if (String(input) === "/api/v1/auth/session") {
        return abortablePendingResponse(
          init?.signal as AbortSignal | undefined,
          (signal) => {
            abortedSignals.push(signal);
          },
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
    });

    const view = renderApp();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });
    view.unmount();

    await waitFor(() => {
      expect(abortedSignals).toHaveLength(1);
    });
    await expectStableFetchCount(fetchMock, 1);
  });

  it("aborts an in-flight refresh when entering the debug harness and loads the ordinary shell after leaving it", async () => {
    const abortedSignals: AbortSignal[] = [];
    let readyToLoad = false;
    installLandingShellFetch(fetchMock, {
      session: ({ init }) => {
        if (!readyToLoad) {
          return abortablePendingResponse(
            init?.signal as AbortSignal | undefined,
            (signal) => {
              abortedSignals.push(signal);
            },
          );
        }
        return sessionResource({
          display_name: "Operator",
        });
      },
    });

    renderApp();

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });
    expect(screen.getByTestId(phase1AuthTestId("status")).textContent).toBe(
      "Checking current session…",
    );

    window.history.pushState({}, "", "/?debug=harness");
    fireEvent.popState(window);

    expect(await screen.findByText("Debug harness shell")).toBeTruthy();
    await waitFor(() => {
      expect(abortedSignals).toHaveLength(1);
    });

    readyToLoad = true;
    window.history.pushState({}, "", "/");
    fireEvent.popState(window);

    expect(
      await screen.findByTestId(phase1LandingTestId("empty-state")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("current-user")).textContent,
    ).toBe("Operator");
    await expectStableFetchCount(fetchMock, 4);
  });
});

function renderApp() {
  return render(<AppRoot />);
}
