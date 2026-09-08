import { incidentAdministrationTestId } from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import { LifecycleTestSurface } from "./incidentLifecycleTestSurface";

describe("IncidentAdminPanel", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("renders null and view-schema workbook preference sheet refs without treating refs as strings", async () => {
    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/00000000-0000-4000-8000-000000001001") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (
        url ===
        "/api/v1/incidents/00000000-0000-4000-8000-000000001001/workbook-preferences/default"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "00000000-0000-4000-8000-000000001001",
              default_sheet_ref: null,
            },
          }),
        );
      }
      if (
        url ===
        "/api/v1/incidents/00000000-0000-4000-8000-000000001001/workbook-preferences/me"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "00000000-0000-4000-8000-000000001001",
              user_id: "00000000-0000-4000-8000-000000000001",
              home_sheet_ref: {
                kind: "view_schema",
                id: "cartulary.view.timeline.v2",
              },
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${url}`);
    });

    render(
      <IncidentAdminPanel
        activeSection="summary"
        currentIncidentRole="admin"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );

    await screen.findByText("Incident controls synced.");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-default-sheet-ref"))
        .textContent,
    ).toBe("Unset");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
        .textContent,
    ).toBe("View schema: Timeline (cartulary.view.timeline.v2)");
  });

  it("renders saved-view workbook preference sheet refs and marks malformed refs unavailable", async () => {
    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/00000000-0000-4000-8000-000000001001") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (
        url ===
        "/api/v1/incidents/00000000-0000-4000-8000-000000001001/workbook-preferences/default"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "00000000-0000-4000-8000-000000001001",
              default_sheet_ref: {
                kind: "saved_view",
                id: "saved-view-1",
              },
            },
          }),
        );
      }
      if (
        url ===
        "/api/v1/incidents/00000000-0000-4000-8000-000000001001/workbook-preferences/me"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "00000000-0000-4000-8000-000000001001",
              user_id: "00000000-0000-4000-8000-000000000001",
              home_sheet_ref: {
                kind: "legacy_workspace",
                id: "legacy-1",
              },
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${url}`);
    });

    render(
      <IncidentAdminPanel
        activeSection="summary"
        currentIncidentRole="admin"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );

    await screen.findByText(
      "Incident summary synced; workbook preferences unavailable.",
    );
    expect(
      screen.getByTestId(incidentAdministrationTestId("summary-key"))
        .textContent,
    ).toBe("IR-201");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-default-sheet-ref"))
        .textContent,
    ).toBe("Saved view: saved-view-1");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
        .textContent,
    ).toBe("Unavailable");
    expect(
      screen.getByTestId(incidentAdministrationTestId("admin-error-code"))
        .textContent,
    ).toBe("");
  });

  it("composes reviewed lifecycle actions with summary and preserves preferences during reconciliation", async () => {
    const lifecycleRequests: Array<Record<string, unknown>> = [];
    let currentIncident = incidentSummary();
    let incidentReads = 0;

    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (
        url === "/api/v1/incidents/00000000-0000-4000-8000-000000001001" &&
        method === "GET"
      ) {
        incidentReads += 1;
        return Promise.resolve(
          jsonResponse({ data: currentIncident, meta: { request_id: "read" } }),
        );
      }
      if (
        (url ===
          "/api/v1/incidents/00000000-0000-4000-8000-000000001001/workbook-preferences/default" ||
          url ===
            "/api/v1/incidents/00000000-0000-4000-8000-000000001001/workbook-preferences/me") &&
        method === "GET"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: url.endsWith("/default")
              ? {
                  default_sheet_ref: null,
                  incident_id: "00000000-0000-4000-8000-000000001001",
                }
              : {
                  home_sheet_ref: null,
                  incident_id: "00000000-0000-4000-8000-000000001001",
                  user_id: "00000000-0000-4000-8000-000000000001",
                },
          }),
        );
      }
      if (
        url ===
          "/api/v1/incidents/00000000-0000-4000-8000-000000001001/close" &&
        method === "POST"
      ) {
        lifecycleRequests.push(JSON.parse(String(init?.body)));
        currentIncident = incidentSummary({
          closed_at: "2026-07-26T12:00:00Z",
          incident_version: 2,
          status: "closed",
        });
        return Promise.resolve(
          jsonResponse({
            data: currentIncident,
            meta: { request_id: "close-request" },
          }),
        );
      }
      if (
        url ===
          "/api/v1/incidents/00000000-0000-4000-8000-000000001001/reopen" &&
        method === "POST"
      ) {
        lifecycleRequests.push(JSON.parse(String(init?.body)));
        currentIncident = incidentSummary({
          closed_at: "2026-07-26T12:00:00Z",
          incident_version: 3,
          status: "closed",
        });
        return Promise.resolve(errorResponse("incident_version_conflict", 409));
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    render(
      <LifecycleTestSurface
        activeSection="summary"
        currentIncidentRole="admin"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );

    await screen.findByText("Incident controls synced.");
    fireEvent.change(
      screen.getByTestId(incidentAdministrationTestId("lifecycle-reason")),
      {
        target: { value: "  containment complete  " },
      },
    );
    await waitFor(() =>
      expect(
        screen
          .getByTestId(incidentAdministrationTestId("close-button"))
          .getAttribute("aria-disabled"),
      ).toBe("false"),
    );
    fireEvent.click(
      screen.getByTestId(incidentAdministrationTestId("close-button")),
    );
    expect(lifecycleRequests).toHaveLength(0);
    fireEvent.click(
      screen.getByTestId(incidentAdministrationTestId("lifecycle-confirm")),
    );
    await screen.findByText("Close confirmed.");
    expect(
      screen.getByTestId(incidentAdministrationTestId("summary-status"))
        .textContent,
    ).toBe("Closed, read-only");
    expect(lifecycleRequests[0]).toMatchObject({
      base_incident_version: 1,
      reason: "  containment complete  ",
    });
    expect(typeof lifecycleRequests[0]?.client_txn_id).toBe("string");

    fireEvent.change(
      screen.getByTestId(incidentAdministrationTestId("lifecycle-reason")),
      {
        target: { value: "new evidence" },
      },
    );
    await waitFor(() =>
      expect(
        screen
          .getByTestId(incidentAdministrationTestId("reopen-button"))
          .getAttribute("aria-disabled"),
      ).toBe("false"),
    );
    fireEvent.click(
      screen.getByTestId(incidentAdministrationTestId("reopen-button")),
    );
    fireEvent.click(
      screen.getByTestId(incidentAdministrationTestId("lifecycle-confirm")),
    );
    await screen.findByText(/incident version changed.*review/i);
    expect(lifecycleRequests[1]).toMatchObject({
      base_incident_version: 2,
      reason: "new evidence",
    });
    expect(incidentReads).toBeGreaterThanOrEqual(2);
    expect(
      fetchMock.mock.calls.filter(([url]) =>
        String(url).includes("workbook-preferences"),
      ),
    ).toHaveLength(2);
    expect(
      screen.getByTestId(incidentAdministrationTestId("summary-version"))
        .textContent,
    ).toContain("3");
    expect(
      screen
        .getByTestId(incidentAdministrationTestId("reopen-button"))
        .getAttribute("disabled"),
    ).toBeNull();
  });

  it("keeps incident summary visible when a workbook preference route fails", async () => {
    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/00000000-0000-4000-8000-000000001001") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (
        url ===
        "/api/v1/incidents/00000000-0000-4000-8000-000000001001/workbook-preferences/default"
      ) {
        return Promise.resolve(errorResponse("preference_unavailable", 500));
      }
      if (
        url ===
        "/api/v1/incidents/00000000-0000-4000-8000-000000001001/workbook-preferences/me"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "00000000-0000-4000-8000-000000001001",
              user_id: "00000000-0000-4000-8000-000000000001",
              home_sheet_ref: null,
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${url}`);
    });

    render(
      <IncidentAdminPanel
        activeSection="summary"
        currentIncidentRole="admin"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );

    await screen.findByText(
      "Incident summary synced; workbook preferences unavailable.",
    );
    expect(
      screen.getByTestId(incidentAdministrationTestId("summary-title"))
        .textContent,
    ).toBe("Incident 201");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-default-sheet-ref"))
        .textContent,
    ).toBe("Unavailable");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
        .textContent,
    ).toBe("Unset");
  });
});

function incidentSummary(overrides?: Record<string, unknown>) {
  return {
    incident_id: "00000000-0000-4000-8000-000000001001",
    incident_key: "IR-201",
    title: "Incident 201",
    created_at: "2026-07-26T10:00:00Z",
    created_by_user_id: "00000000-0000-4000-8000-000000000010",
    description: null,
    severity: null,
    tlp: "TLP:AMBER",
    current_phase: "triage",
    primary_external_case_ref: "CASE-201",
    incident_version: 1,
    status: "active",
    closed_at: null,
    updated_at: "2026-07-26T11:00:00Z",
    updated_by_user_id: "00000000-0000-4000-8000-000000000010",
    ...overrides,
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

function errorResponse(code: string, status: number) {
  return jsonResponse(
    {
      error: {
        code,
      },
    },
    status,
  );
}
