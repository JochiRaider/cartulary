import {
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
        currentIncidentRole="reviewer"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    expect(screen.getByTestId("incident-patch-button")).toBeTruthy();
    expect(
      screen.queryByTestId(incidentMembershipCreateButtonTestId()),
    ).toBeNull();
    expect(
      screen.getByTestId(incidentMembershipAdminNoteTestId()),
    ).toBeTruthy();

    view.rerender(
      <IncidentAdminPanel
        currentIncidentRole="admin"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    expect(screen.getByTestId("incident-patch-button")).toBeTruthy();
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
        currentIncidentRole="admin"
        incidentId="incident-lost"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    await waitFor(() => {
      expect(onIncidentAccessLost).toHaveBeenCalledTimes(1);
    });
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

function incidentSummary() {
  return {
    incident_id: "incident-1",
    incident_key: "IR-201",
    title: "Incident 201",
    description: null,
    severity: null,
    tlp: "amber",
    current_phase: "triage",
    primary_external_case_ref: "CASE-201",
    incident_version: 1,
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
