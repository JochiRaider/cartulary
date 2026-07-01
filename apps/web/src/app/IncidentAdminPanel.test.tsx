import {
  incidentControlsActionMessageTestId,
  incidentControlsStatusTestId,
  incidentControlsSurfaceTestId,
  incidentMembershipAdminNoteTestId,
  incidentMembershipCreateButtonTestId,
  incidentMembershipDeleteButtonTestId,
  incidentMembershipEmailInputTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleDisplayTestId,
  incidentMembershipRoleInputTestId,
  incidentMembershipRowTestId,
  incidentMembershipVersionTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { IncidentAdminPanel } from "./IncidentAdminPanel";

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

  it("Phase 2 U-2-12 ordinary incident shell gates promoted-field controls by incident role, hides membership-admin controls from non-admin members, and returns to landing when incident access is lost", async () => {
    const onIncidentAccessLost = vi.fn();
    const memberships = [
      membershipRecord("user-1", "Operator", "admin", 1),
      membershipRecord("user-2", "Viewer Analyst", "viewer", 1),
    ];

    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/incident-1") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (url === "/api/v1/incidents/incident-1/memberships") {
        return Promise.resolve(
          jsonResponse({
            data: {
              memberships,
            },
          }),
        );
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/default") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              default_sheet_ref: null,
            },
          }),
        );
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/me") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              user_id: "user-1",
              home_sheet_ref: null,
            },
          }),
        );
      }
      if (url.startsWith("/api/v1/incidents/incident-lost")) {
        return Promise.resolve(errorResponse("incident_not_found", 404));
      }
      throw new Error(`unexpected fetch: ${url}`);
    });

    const view = render(
      <IncidentAdminPanel
        activeSection="incident-fields"
        currentIncidentRole="viewer"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );

    await screen.findByText("Incident controls synced.");
    expect(screen.queryByTestId("incident-patch-button")).toBeNull();
    expect(
      screen.getByTestId("incident-patch-readonly-note").textContent,
    ).toContain("read-only");

    view.rerender(
      <IncidentAdminPanel
        activeSection="memberships"
        currentIncidentRole="viewer"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    await screen.findByTestId(incidentMembershipRoleDisplayTestId("user-2"));
    expect(
      screen.queryByTestId(incidentMembershipCreateButtonTestId()),
    ).toBeNull();
    expect(
      screen.getByTestId(incidentMembershipAdminNoteTestId()).textContent,
    ).toContain("Only incident admins");
    expect(
      screen.getByTestId(incidentMembershipRoleDisplayTestId("user-2"))
        .textContent,
    ).toBe("viewer");

    view.rerender(
      <IncidentAdminPanel
        activeSection="incident-fields"
        currentIncidentRole="reviewer"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    expect(screen.getByTestId("incident-patch-button")).toBeTruthy();

    view.rerender(
      <IncidentAdminPanel
        activeSection="memberships"
        currentIncidentRole="reviewer"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    expect(
      screen.queryByTestId(incidentMembershipCreateButtonTestId()),
    ).toBeNull();
    expect(
      screen.getByTestId(incidentMembershipAdminNoteTestId()),
    ).toBeTruthy();

    view.rerender(
      <IncidentAdminPanel
        activeSection="memberships"
        currentIncidentRole="admin"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    await screen.findByTestId(incidentMembershipPatchButtonTestId("user-2"));
    expect(
      screen.getByTestId(incidentMembershipCreateButtonTestId()),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentMembershipPatchButtonTestId("user-2")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentMembershipDeleteButtonTestId("user-2")),
    ).toBeTruthy();
    expect(
      screen.queryByTestId(incidentMembershipAdminNoteTestId()),
    ).toBeNull();

    view.rerender(
      <IncidentAdminPanel
        activeSection="incident-fields"
        currentIncidentRole="admin"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    expect(screen.getByTestId("incident-patch-button")).toBeTruthy();

    view.rerender(
      <IncidentAdminPanel
        activeSection="summary"
        currentIncidentRole="admin"
        incidentId="incident-lost"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    await waitFor(() => {
      expect(onIncidentAccessLost).toHaveBeenCalledTimes(1);
    });
  });

  it("renders null and view-schema workbook preference sheet refs without treating refs as strings", async () => {
    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/incident-1") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/default") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              default_sheet_ref: null,
            },
          }),
        );
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/me") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              user_id: "user-1",
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
        incidentId="incident-1"
      />,
    );

    await screen.findByText("Incident controls synced.");
    expect(
      screen.getByTestId("incident-pref-default-sheet-ref").textContent,
    ).toBe("Unset");
    expect(screen.getByTestId("incident-pref-home-sheet-ref").textContent).toBe(
      "View schema: Timeline (cartulary.view.timeline.v2)",
    );
  });

  it("renders saved-view workbook preference sheet refs and marks malformed refs unavailable", async () => {
    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/incident-1") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/default") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              default_sheet_ref: {
                kind: "saved_view",
                id: "saved-view-1",
              },
            },
          }),
        );
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/me") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              user_id: "user-1",
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
        incidentId="incident-1"
      />,
    );

    await screen.findByText(
      "Incident summary synced; workbook preferences unavailable.",
    );
    expect(screen.getByTestId("incident-summary-key").textContent).toBe(
      "IR-201",
    );
    expect(
      screen.getByTestId("incident-pref-default-sheet-ref").textContent,
    ).toBe("Saved view: saved-view-1");
    expect(screen.getByTestId("incident-pref-home-sheet-ref").textContent).toBe(
      "Unavailable",
    );
    expect(screen.getByTestId("incident-admin-error-code").textContent).toBe(
      "",
    );
  });

  it("patches required incident metadata fields and restricts TLP to canonical tokens", async () => {
    let patchBody: Record<string, unknown> | null = null;
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/incidents/incident-1" && method === "GET") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (url === "/api/v1/incidents/incident-1" && method === "PATCH") {
        patchBody = JSON.parse(String(init?.body ?? "{}")) as Record<
          string,
          unknown
        >;
        return Promise.resolve(
          jsonResponse({
            data: incidentSummary({
              description: "Escalated investigation",
              severity: "critical",
              tlp: "TLP:RED",
              current_phase: "containment",
              primary_external_case_ref: "CASE-999",
              incident_version: 2,
            }),
          }),
        );
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    render(
      <IncidentAdminPanel
        activeSection="incident-fields"
        currentIncidentRole="reviewer"
        incidentId="incident-1"
      />,
    );

    await screen.findByText("Incident controls synced.");
    fireEvent.change(screen.getByTestId("incident-patch-description"), {
      target: { value: "Escalated investigation" },
    });
    fireEvent.change(screen.getByTestId("incident-patch-severity"), {
      target: { value: "critical" },
    });
    fireEvent.change(screen.getByTestId("incident-patch-tlp"), {
      target: { value: "TLP:RED" },
    });
    fireEvent.change(screen.getByTestId("incident-patch-current-phase"), {
      target: { value: "containment" },
    });
    fireEvent.change(screen.getByTestId("incident-patch-external-case"), {
      target: { value: "CASE-999" },
    });
    fireEvent.click(screen.getByTestId("incident-patch-button"));

    await waitFor(() => {
      expect(patchBody).toEqual({
        base_incident_version: 1,
        description: "Escalated investigation",
        severity: "critical",
        tlp: "TLP:RED",
        current_phase: "containment",
        primary_external_case_ref: "CASE-999",
      });
    });
    await screen.findByText("Saved promoted incident fields.");
    expect(screen.getByTestId(incidentControlsStatusTestId()).textContent).toBe(
      "Incident controls synced.",
    );
    expect(
      screen.getByTestId(incidentControlsActionMessageTestId()).textContent,
    ).toBe("Saved promoted incident fields.");
    expect(
      screen
        .getByTestId(incidentControlsSurfaceTestId())
        .getAttribute("data-incident-controls-load-state"),
    ).toBe("synced");
  });

  it("keeps readiness on the current incident-controls section when a patch finishes after a section switch", async () => {
    const patchResponse = deferredResponse();
    let defaultPreferenceReads = 0;
    let userPreferenceReads = 0;

    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/incidents/incident-1" && method === "GET") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (url === "/api/v1/incidents/incident-1" && method === "PATCH") {
        return patchResponse.promise;
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/default") {
        defaultPreferenceReads += 1;
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              default_sheet_ref: null,
            },
          }),
        );
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/me") {
        userPreferenceReads += 1;
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              user_id: "user-1",
              home_sheet_ref: null,
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    const view = render(
      <IncidentAdminPanel
        activeSection="incident-fields"
        currentIncidentRole="reviewer"
        incidentId="incident-1"
      />,
    );

    await screen.findByText("Incident controls synced.");
    fireEvent.click(screen.getByTestId("incident-patch-button"));
    expect(
      screen.getByTestId(incidentControlsActionMessageTestId()).textContent,
    ).toBe("Saving promoted incident fields…");

    view.rerender(
      <IncidentAdminPanel
        activeSection="summary"
        currentIncidentRole="reviewer"
        incidentId="incident-1"
      />,
    );
    await waitFor(() => {
      const surface = screen.getByTestId(incidentControlsSurfaceTestId());
      expect(surface.getAttribute("data-incident-controls-section")).toBe(
        "summary",
      );
      expect(surface.getAttribute("data-incident-controls-load-state")).toBe(
        "synced",
      );
    });
    expect(defaultPreferenceReads).toBe(1);
    expect(userPreferenceReads).toBe(1);

    await act(async () => {
      patchResponse.resolve(
        jsonResponse({
          data: incidentSummary({
            current_phase: "containment",
            incident_version: 2,
          }),
        }),
      );
      await patchResponse.promise;
    });

    await screen.findByText("Saved promoted incident fields.");
    await waitFor(() => {
      expect(defaultPreferenceReads).toBe(2);
      expect(userPreferenceReads).toBe(2);
    });
    expect(
      screen
        .getByTestId(incidentControlsSurfaceTestId())
        .getAttribute("data-incident-controls-section"),
    ).toBe("summary");
    expect(screen.getByTestId(incidentControlsStatusTestId()).textContent).toBe(
      "Incident controls synced.",
    );
    expect(screen.queryByTestId("incident-patch-button")).toBeNull();
  });

  it("keeps membership audit placement inside incident admin controls", async () => {
    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/incident-1") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      throw new Error(`unexpected fetch: ${url}`);
    });

    render(
      <IncidentAdminPanel
        activeSection="membership-audit"
        currentIncidentRole="admin"
        incidentId="incident-1"
      />,
    );

    expect(await screen.findByTestId("membership-audit-note")).toBeTruthy();
    expect(screen.getByTestId("membership-audit-note").textContent).toContain(
      "incident-scoped",
    );
    expect(
      fetchMock.mock.calls.some(([input]) =>
        String(input).includes("administrative-audit"),
      ),
    ).toBe(false);
  });

  it("keeps incident summary visible when a workbook preference route fails", async () => {
    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/incident-1") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/default") {
        return Promise.resolve(errorResponse("preference_unavailable", 500));
      }
      if (url === "/api/v1/incidents/incident-1/workbook-preferences/me") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              user_id: "user-1",
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
        incidentId="incident-1"
      />,
    );

    await screen.findByText(
      "Incident summary synced; workbook preferences unavailable.",
    );
    expect(screen.getByTestId("incident-summary-title").textContent).toBe(
      "Incident 201",
    );
    expect(
      screen.getByTestId("incident-pref-default-sheet-ref").textContent,
    ).toBe("Unavailable");
    expect(screen.getByTestId("incident-pref-home-sheet-ref").textContent).toBe(
      "Unset",
    );
  });

  it("Phase 2 U-2-13 ordinary incident shell issues membership create, patch, and delete requests with versioned payloads and refreshes session role after each mutation", async () => {
    const onSessionRoleChange = vi.fn().mockResolvedValue(undefined);
    const requests: Array<{
      method: string;
      url: string;
      body: Record<string, unknown> | null;
      headers: Headers;
    }> = [];
    let memberships = [membershipRecord("user-1", "Operator", "admin", 1)];

    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();

      if (url === "/api/v1/incidents/incident-1" && method === "GET") {
        return Promise.resolve(jsonResponse({ data: incidentSummary() }));
      }
      if (
        url === "/api/v1/incidents/incident-1/memberships" &&
        method === "GET"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              memberships,
            },
          }),
        );
      }
      if (
        url === "/api/v1/incidents/incident-1/workbook-preferences/default" &&
        method === "GET"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              default_sheet_ref: null,
            },
          }),
        );
      }
      if (
        url === "/api/v1/incidents/incident-1/workbook-preferences/me" &&
        method === "GET"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "incident-1",
              user_id: "user-1",
              home_sheet_ref: null,
            },
          }),
        );
      }
      if (
        url === "/api/v1/incidents/incident-1/memberships" &&
        method === "POST"
      ) {
        const body = JSON.parse(String(init?.body ?? "{}")) as Record<
          string,
          unknown
        >;
        requests.push({
          method,
          url,
          body,
          headers: new Headers(init?.headers),
        });
        memberships = [
          ...memberships,
          membershipRecord("user-2", "Analyst", String(body.role), 1),
        ];
        return Promise.resolve(
          jsonResponse(
            { data: membershipRecord("user-2", "Analyst", "viewer", 1) },
            201,
          ),
        );
      }
      if (
        url === "/api/v1/incidents/incident-1/memberships/user-2" &&
        method === "PATCH"
      ) {
        const body = JSON.parse(String(init?.body ?? "{}")) as Record<
          string,
          unknown
        >;
        requests.push({
          method,
          url,
          body,
          headers: new Headers(init?.headers),
        });
        memberships = memberships.map((membership) =>
          membership.user_id === "user-2"
            ? membershipRecord("user-2", "Analyst", String(body.role), 2)
            : membership,
        );
        return Promise.resolve(
          jsonResponse({
            data: membershipRecord("user-2", "Analyst", "reviewer", 2),
          }),
        );
      }
      if (
        url === "/api/v1/incidents/incident-1/memberships/user-2" &&
        method === "DELETE"
      ) {
        const body = JSON.parse(String(init?.body ?? "{}")) as Record<
          string,
          unknown
        >;
        requests.push({
          method,
          url,
          body,
          headers: new Headers(init?.headers),
        });
        memberships = memberships.filter(
          (membership) => membership.user_id !== "user-2",
        );
        return Promise.resolve(new Response(null, { status: 204 }));
      }

      throw new Error(`unexpected fetch: ${method} ${url}`);
    });

    render(
      <IncidentAdminPanel
        activeSection="memberships"
        currentIncidentRole="admin"
        incidentId="incident-1"
        onSessionRoleChange={onSessionRoleChange}
      />,
    );

    await screen.findByText("Incident controls synced.");

    fireEvent.change(screen.getByTestId(incidentMembershipEmailInputTestId()), {
      target: { value: " analyst@example.test " },
    });
    fireEvent.click(screen.getByTestId(incidentMembershipCreateButtonTestId()));

    await waitFor(() => {
      expect(onSessionRoleChange).toHaveBeenCalledTimes(1);
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(incidentMembershipRowTestId("user-2")),
      ).toBeTruthy();
    });

    fireEvent.change(
      screen.getByTestId(incidentMembershipRoleInputTestId("user-2")),
      {
        target: { value: "reviewer" },
      },
    );
    fireEvent.click(
      screen.getByTestId(incidentMembershipPatchButtonTestId("user-2")),
    );

    await waitFor(() => {
      expect(onSessionRoleChange).toHaveBeenCalledTimes(2);
    });
    expect(
      screen.getByTestId(incidentMembershipVersionTestId("user-2")).textContent,
    ).toContain("Version 2");

    fireEvent.click(
      screen.getByTestId(incidentMembershipDeleteButtonTestId("user-2")),
    );

    await waitFor(() => {
      expect(onSessionRoleChange).toHaveBeenCalledTimes(3);
    });
    await waitFor(() => {
      expect(
        screen.queryByTestId(incidentMembershipRowTestId("user-2")),
      ).toBeNull();
    });

    expect(requests).toHaveLength(3);
    expect(requests[0]?.method).toBe("POST");
    expect(requests[0]?.url).toBe("/api/v1/incidents/incident-1/memberships");
    expect(requests[0]?.body?.email).toBe("analyst@example.test");
    expect(requests[0]?.body?.role).toBe("viewer");
    expect(typeof requests[0]?.body?.client_txn_id).toBe("string");

    expect(requests[1]?.method).toBe("PATCH");
    expect(requests[1]?.body).toEqual({
      base_membership_version: 1,
      role: "reviewer",
    });

    expect(requests[2]?.method).toBe("DELETE");
    expect(requests[2]?.body).toEqual({
      base_membership_version: 2,
    });

    for (const request of requests) {
      expect(request.headers.get("Content-Type")).toBe("application/json");
    }
  });
});

function incidentSummary(overrides?: Record<string, unknown>) {
  return {
    incident_id: "incident-1",
    incident_key: "IR-201",
    title: "Incident 201",
    description: null,
    severity: null,
    tlp: "TLP:AMBER",
    current_phase: "triage",
    primary_external_case_ref: "CASE-201",
    incident_version: 1,
    status: "active",
    closed_at: null,
    ...overrides,
  };
}

function membershipRecord(
  userId: string,
  displayName: string,
  role: string,
  membershipVersion: number,
) {
  return {
    incident_id: "incident-1",
    user_id: userId,
    display_name: displayName,
    role,
    membership_version: membershipVersion,
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

function deferredResponse() {
  let resolve: (response: Response) => void = () => {};
  const promise = new Promise<Response>((next) => {
    resolve = next;
  });
  return { promise, resolve };
}
