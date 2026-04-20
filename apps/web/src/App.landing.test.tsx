import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("./WorkbookShell", () => ({
  WorkbookShell: ({
    incidentId,
    onIncidentAccessLost,
  }: {
    incidentId: string;
    onIncidentAccessLost?: () => void;
  }) => (
    <section data-testid="mock-workbook">
      <p data-testid="mock-workbook-incident">{incidentId}</p>
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
  ),
  buildCreatePayload: vi.fn(),
  createDraftRow: vi.fn(),
  ensureDraftRow: vi.fn(),
  TimelineWorkbook: vi.fn(),
}));

import { App } from "./App";

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
            data: {
              user_id: "user-1",
              display_name: "Bootstrap Admin",
              is_deployment_admin: true,
              memberships: [],
            },
          }),
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

    render(<App />);

    expect(
      (await screen.findByTestId("landing-empty-state")).textContent,
    ).toContain("No incidents are visible");
    expect(screen.getByTestId("landing-incidents-count").textContent).toBe("0");
    expect(screen.getByTestId("landing-current-user").textContent).toBe(
      "Bootstrap Admin · deployment admin",
    );
  });

  it("renders one or many visible incidents on the landing screen", async () => {
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: {
              user_id: "user-1",
              display_name: "Operator",
              is_deployment_admin: false,
              memberships: [],
            },
          }),
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

    render(<App />);

    expect(await screen.findByTestId("landing-incident-list")).toBeTruthy();
    expect(screen.getByTestId("landing-incidents-count").textContent).toBe("2");
    expect(
      screen.getByTestId("landing-incident-incident-1").textContent,
    ).toContain("First Incident");
    expect(
      screen.getByTestId("landing-incident-incident-2").textContent,
    ).toContain("Second Incident");
  });

  it("navigates into the workbook after create succeeds", async () => {
    let created = false;
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: {
              user_id: "user-1",
              display_name: "Operator",
              is_deployment_admin: false,
              memberships: created
                ? [
                    {
                      incident_id: "incident-created",
                      role: "admin",
                    },
                  ]
                : [],
            },
          }),
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

    render(<App />);

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
    expect(window.location.search).toContain("incident_id=incident-created");
  });

  it("falls back to the landing screen when the requested incident is stale", async () => {
    window.history.replaceState({}, "", "/?incident_id=incident-stale");
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: {
              user_id: "user-1",
              display_name: "Operator",
              is_deployment_admin: false,
              memberships: [],
            },
          }),
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

    render(<App />);

    expect(await screen.findByTestId("landing-incident-list")).toBeTruthy();
    expect(screen.getByTestId("landing-status").textContent).toContain(
      "no longer visible",
    );
    expect(window.location.search).not.toContain("incident_id=");
  });

  it("returns to the landing screen when workbook access is lost", async () => {
    window.history.replaceState({}, "", "/?incident_id=incident-5");
    let accessLost = false;
    fetchMock.mockImplementation((input) => {
      if (String(input) === "/api/v1/auth/session") {
        return Promise.resolve(
          jsonResponse({
            data: {
              user_id: "user-1",
              display_name: "Operator",
              is_deployment_admin: false,
              memberships: accessLost
                ? []
                : [
                    {
                      incident_id: "incident-5",
                      role: "admin",
                    },
                  ],
            },
          }),
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

    render(<App />);

    expect(await screen.findByTestId("mock-workbook")).toBeTruthy();
    accessLost = true;
    fireEvent.click(screen.getByTestId("mock-access-lost"));

    await waitFor(() => {
      expect(screen.getByTestId("landing-empty-state")).toBeTruthy();
    });
    expect(screen.getByTestId("landing-status").textContent).toContain(
      "no longer visible",
    );
    expect(window.location.search).not.toContain("incident_id=");
  });
});

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

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
