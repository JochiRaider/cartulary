import {
  deploymentUserRowTestId,
  landingAdminCommandTestId,
  landingAdminPanelTestId,
  landingAdminShellTestId,
  landingAdminTabTestId,
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1LandingTestId,
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
      (await screen.findByTestId(phase1LandingTestId("empty-state")))
        .textContent,
    ).toContain("No incidents are visible");
    expect(screen.getByTestId(landingAdminShellTestId("shell"))).toBeTruthy();
    expect(
      screen
        .getByTestId(landingAdminShellTestId("tablist"))
        .getAttribute("role"),
    ).toBe("tablist");
    expect(
      screen
        .getByTestId(landingAdminTabTestId("incidents"))
        .getAttribute("aria-selected"),
    ).toBe("true");
    expect(
      screen
        .getByTestId(landingAdminPanelTestId("incidents"))
        .getAttribute("role"),
    ).toBe("tabpanel");
    expect(
      (
        screen.getByTestId(
          landingAdminCommandTestId("incidents", "open selected"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      screen.getByTestId(
        landingAdminCommandTestId("incidents", "new incident"),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(phase1LandingTestId("incidents-count")).textContent,
    ).toBe("0");
    expect(
      screen.getByTestId(phase1LandingTestId("current-user")).textContent,
    ).toBe("Bootstrap Admin · deployment admin");
    await expectStableFetchCount(fetchMock, 3);
  });

  it("switches ribbon panels and exposes contextual commands", async () => {
    installLandingShellFetch(fetchMock, {
      session: sessionResource({
        display_name: "Operator",
      }),
    });

    renderApp();

    await screen.findByTestId(phase1LandingTestId("empty-state"));
    fireEvent.click(
      screen.getByTestId(landingAdminTabTestId("account-security")),
    );
    expect(
      screen
        .getByTestId(landingAdminTabTestId("account-security"))
        .getAttribute("aria-selected"),
    ).toBe("true");
    expect(
      screen
        .getByTestId(landingAdminPanelTestId("account-security"))
        .getAttribute("aria-labelledby"),
    ).toBe(landingAdminTabTestId("account-security"));
    expect(
      screen.getByTestId(
        landingAdminCommandTestId("account-security", "refresh account"),
      ),
    ).toBeTruthy();

    fireEvent.keyDown(screen.getByTestId(landingAdminShellTestId("tablist")), {
      key: "End",
    });
    await waitFor(() => {
      expect(
        screen
          .getByTestId(landingAdminTabTestId("reference-packs"))
          .getAttribute("aria-selected"),
      ).toBe("true");
    });
    expect(
      (
        screen.getByTestId(
          landingAdminCommandTestId("reference-packs", "import"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);

    fireEvent.click(
      screen.getByTestId(landingAdminTabTestId("deployment-users")),
    );
    expect(
      screen.getByTestId(phase1AdminTestId("access-note")).textContent,
    ).toContain("Deployment admin access is required");
    expect(
      (
        screen.getByTestId(
          landingAdminCommandTestId("deployment-users", "save target"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
  });

  it("enables deployment-user ribbon target commands after selecting a loaded user", async () => {
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
      screen.getByTestId(landingAdminTabTestId("deployment-users")),
    );
    expect(
      (
        screen.getByTestId(
          landingAdminCommandTestId("deployment-users", "save target"),
        ) as HTMLButtonElement
      ).disabled,
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
      (
        screen.getByTestId(
          landingAdminCommandTestId("deployment-users", "save target"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(false);
    expect(
      (
        screen.getByTestId(
          landingAdminCommandTestId("deployment-users", "reset password"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(false);
    expect(
      (
        screen.getByTestId(
          landingAdminCommandTestId("deployment-users", "revoke sessions"),
        ) as HTMLButtonElement
      ).disabled,
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
    await waitFor(() => {
      expect(
        screen
          .getByTestId(landingIncidentCardTestId("incident-1"))
          .getAttribute("data-selected"),
      ).toBe("true");
    });

    const secondIncidentSelector = screen
      .getByTestId(landingIncidentCardTestId("incident-2"))
      .querySelector("button");
    if (secondIncidentSelector === null) {
      throw new Error("Missing incident selection button");
    }
    fireEvent.click(secondIncidentSelector);
    expect(
      screen
        .getByTestId(landingIncidentCardTestId("incident-2"))
        .getAttribute("data-selected"),
    ).toBe("true");
    const openSelectedCommand = screen.getByTestId(
      landingAdminCommandTestId("incidents", "open selected"),
    ) as HTMLButtonElement;
    expect(openSelectedCommand.disabled).toBe(false);
    fireEvent.click(openSelectedCommand);

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
      screen.getByTestId(phase1LandingTestId("status")).textContent,
    ).toContain("no longer visible");
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
      screen.getByTestId(phase1LandingTestId("status")).textContent,
    ).toContain("no longer visible");
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
