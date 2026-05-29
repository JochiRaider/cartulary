import {
  landingIncidentCardTestId,
  landingIncidentOpenButtonTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./WorkbookShell", async () => {
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

import { AppRoot } from "./AppRoot";
import {
  incidentResource,
  installLandingShellFetch,
  sessionResource,
} from "./appShellTestSupport";
import {
  abortablePendingResponse,
  expectStableFetchCount,
} from "./fetchMockTestSupport";

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
      (await screen.findByTestId("landing-empty-state")).textContent,
    ).toContain("No incidents are visible");
    expect(screen.getByTestId("landing-incidents-count").textContent).toBe("0");
    expect(screen.getByTestId("landing-current-user").textContent).toBe(
      "Bootstrap Admin · deployment admin",
    );
    await expectStableFetchCount(fetchMock, 3);
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

    expect(await screen.findByTestId("landing-incident-list")).toBeTruthy();
    expect(screen.getByTestId("landing-incidents-count").textContent).toBe("2");
    expect(
      screen.getByTestId(landingIncidentCardTestId("incident-1")).textContent,
    ).toContain("First Incident");
    expect(
      screen.getByTestId(landingIncidentCardTestId("incident-2")).textContent,
    ).toContain("Second Incident");
    await expectStableFetchCount(fetchMock, 3);
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

    await screen.findByTestId("landing-empty-state");
    fireEvent.change(screen.getByTestId("landing-incident-key"), {
      target: { value: "IR-203" },
    });
    fireEvent.change(screen.getByTestId("landing-incident-title"), {
      target: { value: "Created Incident" },
    });
    fireEvent.click(screen.getByTestId("landing-create-button"));

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

    expect(await screen.findByTestId("landing-incident-list")).toBeTruthy();
    expect(
      screen.getByTestId(landingIncidentOpenButtonTestId("incident-live")),
    ).toBeTruthy();
    expect(screen.getByTestId("landing-status").textContent).toContain(
      "no longer visible",
    );
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
      expect(screen.getByTestId("landing-empty-state")).toBeTruthy();
    });
    expect(screen.getByTestId("landing-status").textContent).toContain(
      "no longer visible",
    );
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
    expect(screen.getByTestId("auth-status").textContent).toBe(
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

    expect(await screen.findByTestId("landing-empty-state")).toBeTruthy();
    expect(screen.getByTestId("landing-current-user").textContent).toBe(
      "Operator",
    );
    await expectStableFetchCount(fetchMock, 4);
  });
});

function renderApp() {
  return render(<AppRoot />);
}
