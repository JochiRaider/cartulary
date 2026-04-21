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
import type { CredentialState, SessionData } from "./phase1Client";

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
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Bootstrap Admin",
              is_deployment_admin: true,
            }),
          }),
        );
      }
      if (String(input) === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({ data: credentialStateResource() }),
        );
      }
      if (String(input) === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [],
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
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
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Operator",
            }),
          }),
        );
      }
      if (String(input) === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({ data: credentialStateResource() }),
        );
      }
      if (String(input) === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [
                incidentResource("incident-1", "IR-201", "First Incident"),
                incidentResource("incident-2", "IR-202", "Second Incident"),
              ],
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
    });

    renderApp();

    expect(await screen.findByTestId("landing-incident-list")).toBeTruthy();
    expect(screen.getByTestId("landing-incidents-count").textContent).toBe("2");
    expect(
      screen.getByTestId("landing-incident-incident-1").textContent,
    ).toContain("First Incident");
    expect(
      screen.getByTestId("landing-incident-incident-2").textContent,
    ).toContain("Second Incident");
    await expectStableFetchCount(fetchMock, 3);
  });

  it("Phase 2 U-2-11 ordinary landing shell creates an incident, refreshes session-visible membership, routes to the workbook by incident_id, and falls back when a stale incident selection is no longer visible", async () => {
    let created = false;
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
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
          }),
        );
      }
      if (url === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({ data: credentialStateResource() }),
        );
      }
      if (url === "/api/v1/incidents" && method === "GET") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: created
                ? [
                    incidentResource(
                      "incident-created",
                      "IR-203",
                      "Created Incident",
                    ),
                  ]
                : [],
            },
          }),
        );
      }
      if (url === "/api/v1/incidents" && method === "POST") {
        created = true;
        return Promise.resolve(
          jsonResponse({
            data: incidentResource(
              "incident-created",
              "IR-203",
              "Created Incident",
            ),
          }),
        );
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
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
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Operator",
            }),
          }),
        );
      }
      if (String(input) === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({ data: credentialStateResource() }),
        );
      }
      if (String(input) === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [
                incidentResource("incident-live", "IR-204", "Live Incident"),
              ],
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
    });

    renderApp();

    expect(await screen.findByTestId("landing-incident-list")).toBeTruthy();
    expect(screen.getByTestId("landing-status").textContent).toContain(
      "no longer visible",
    );
    expect(window.location.search).not.toContain("incident_id=");
  });

  it("renders the workbook directly from an authenticated incident route under StrictMode", async () => {
    window.history.replaceState({}, "", "/?incident_id=incident-5");
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Operator",
              memberships: [
                {
                  incident_id: "incident-5",
                  role: "admin",
                },
              ],
            }),
          }),
        );
      }
      if (String(input) === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({ data: credentialStateResource() }),
        );
      }
      if (String(input) === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [
                incidentResource("incident-5", "IR-205", "Visible Incident"),
              ],
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
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
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
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
          }),
        );
      }
      if (String(input) === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({ data: credentialStateResource() }),
        );
      }
      if (String(input) === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: accessLost
                ? []
                : [
                    incidentResource(
                      "incident-5",
                      "IR-205",
                      "Visible Incident",
                    ),
                  ],
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
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
        const signal = init?.signal as AbortSignal | undefined;
        return new Promise<Response>((_, reject) => {
          const abort = () => {
            if (signal) {
              abortedSignals.push(signal);
            }
            reject(new DOMException("Aborted", "AbortError"));
          };
          if (signal?.aborted) {
            abort();
            return;
          }
          signal?.addEventListener("abort", abort, { once: true });
        });
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
    await flushMicrotasks();
    expect(fetchMock).toHaveBeenCalledTimes(1);
  });

  it("aborts an in-flight refresh when entering the debug harness and loads the ordinary shell after leaving it", async () => {
    const abortedSignals: AbortSignal[] = [];
    let readyToLoad = false;
    fetchMock.mockImplementation((input, init) => {
      if (String(input) === "/api/v1/auth/session") {
        if (!readyToLoad) {
          const signal = init?.signal as AbortSignal | undefined;
          return new Promise<Response>((_, reject) => {
            const abort = () => {
              if (signal) {
                abortedSignals.push(signal);
              }
              reject(new DOMException("Aborted", "AbortError"));
            };
            if (signal?.aborted) {
              abort();
              return;
            }
            signal?.addEventListener("abort", abort, { once: true });
          });
        }
        return Promise.resolve(
          jsonResponse({
            data: sessionResource({
              display_name: "Operator",
            }),
          }),
        );
      }
      if (String(input) === "/api/v1/auth/credential-state") {
        return Promise.resolve(
          jsonResponse({ data: credentialStateResource() }),
        );
      }
      if (String(input) === "/api/v1/incidents") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incidents: [],
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${String(input)}`);
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

function incidentResource(
  incidentId: string,
  incidentKey: string,
  title: string,
) {
  return {
    incident_id: incidentId,
    incident_key: incidentKey,
    title,
    description: null,
    severity: null,
    tlp: null,
    current_phase: null,
    primary_external_case_ref: null,
    incident_version: 1,
  };
}

function sessionResource(overrides?: Partial<SessionData>): SessionData {
  return {
    user_id: "user-1",
    display_name: "Operator",
    provider_type: "local",
    mfa_state: "not_required",
    is_deployment_admin: false,
    authenticated_at: "2026-04-20T12:00:00Z",
    idle_expires_at: "2026-04-20T12:30:00Z",
    absolute_expires_at: "2026-04-20T20:00:00Z",
    session_expires_at: "2026-04-20T12:30:00Z",
    memberships: [],
    ...overrides,
  };
}

function credentialStateResource(): CredentialState {
  return {
    user_id: "user-1",
    auth_kind: "local",
    recovery_model: "admin_assisted",
    password_changed_at: "2026-04-20T12:00:00Z",
    totp: {
      state: "not_enrolled",
      enrolled_at: null,
      pending_expires_at: null,
    },
  };
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}

async function expectStableFetchCount(
  fetchMock: ReturnType<typeof vi.fn>,
  expectedCount: number,
) {
  await waitFor(() => {
    expect(fetchMock).toHaveBeenCalledTimes(expectedCount);
  });
  await flushMicrotasks();
  expect(fetchMock).toHaveBeenCalledTimes(expectedCount);
}

async function flushMicrotasks() {
  await Promise.resolve();
  await Promise.resolve();
}
