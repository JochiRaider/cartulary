import { afterEach, describe, expect, it, vi } from "vitest";
import { hostsViewSchemaId } from "../models/workbookSurfaceRegistry";
import { createWorkbookIncidentAdapter } from "./createWorkbookIncidentAdapter";
import { createWorkbookPreferenceAdapter } from "./createWorkbookPreferenceAdapter";
import { createWorkbookStartupAdapter } from "./createWorkbookStartupAdapter";

const incidentId = "00000000-0000-4000-8000-000000000001";
const otherIncidentId = "00000000-0000-4000-8000-000000000002";
const userId = "00000000-0000-4000-8000-000000000101";
const now = "2026-07-31T20:00:00Z";

function envelope(data: unknown) {
  return { data, meta: { request_id: "req-workbook-port" } };
}

function response(data: unknown, status = 200): Response {
  return new Response(JSON.stringify(data), {
    status,
    headers: { "content-type": "application/json" },
  });
}

function incidentResource(id = incidentId) {
  return {
    closed_at: null,
    created_at: now,
    created_by_user_id: userId,
    current_phase: "containment",
    description: "Incident description",
    incident_id: id,
    incident_key: "INC-001",
    incident_version: 4,
    primary_external_case_ref: null,
    severity: "high",
    status: "active",
    title: "Incident title",
    tlp: "TLP:AMBER",
    updated_at: now,
    updated_by_user_id: userId,
  };
}

function membershipResource(id = incidentId) {
  return {
    added_by_user_id: userId,
    display_name: "Incident Owner",
    incident_id: id,
    joined_at: now,
    membership_version: 2,
    role: "admin",
    updated_at: now,
    updated_by_user_id: userId,
    user_id: userId,
  };
}

function startupResource(id = incidentId) {
  return {
    cleared_pointers: [],
    default_sheet_ref: null,
    extension_workspace_availability: {
      incident_id: id,
      schema_id: "cartulary.extension_workspace_availability.v1",
      workspaces: [
        {
          extension_profile_id: "network_flow_activity",
          workspace_key: "network_analysis",
        },
      ],
    },
    home_sheet_ref: null,
    incident_id: id,
    selected_saved_view: null,
    selected_sheet_ref: { kind: "view_schema", id: hostsViewSchemaId },
    selected_view_schema_id: hostsViewSchemaId,
    source: "explicit",
  };
}

function authorizationDenied(): Response {
  return response(
    {
      error: {
        code: "authorization_denied",
        details: {},
        message: "Access denied.",
        request_id: "req-denied",
        retryable: false,
        status: 403,
      },
    },
    403,
  );
}

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("Workbook incident adapter", () => {
  it("uses projected routes and returns only correlated semantic identity and members", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(envelope(incidentResource())))
      .mockResolvedValueOnce(
        response(envelope({ memberships: [membershipResource()] })),
      );
    vi.stubGlobal("fetch", fetchMock);
    const incident = createWorkbookIncidentAdapter({
      apiBase: "/base",
      incidentId,
    });

    await expect(
      incident.getIdentity({ signal: new AbortController().signal }),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: { incident_id: incidentId, incident_key: "INC-001" },
    });
    await expect(
      incident.listMembers({ signal: new AbortController().signal }),
    ).resolves.toEqual({
      kind: "accepted",
      value: {
        members: [{ displayName: "Incident Owner", userId }],
      },
    });
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      `/base/api/v1/incidents/${incidentId}`,
      expect.objectContaining({ method: "GET" }),
    );
    expect(fetchMock).toHaveBeenNthCalledWith(
      2,
      `/base/api/v1/incidents/${incidentId}/memberships`,
      expect.objectContaining({ method: "GET" }),
    );
  });

  it("fails closed on malformed or cross-incident resources and classifies access loss", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(envelope({ incident_id: incidentId })))
      .mockResolvedValueOnce(
        response(envelope(incidentResource(otherIncidentId))),
      )
      .mockResolvedValueOnce(
        response(
          envelope({ memberships: [membershipResource(otherIncidentId)] }),
        ),
      )
      .mockResolvedValueOnce(authorizationDenied());
    vi.stubGlobal("fetch", fetchMock);
    const incident = createWorkbookIncidentAdapter({
      apiBase: undefined,
      incidentId,
    });

    for (const load of [
      () => incident.getIdentity({ signal: new AbortController().signal }),
      () => incident.getIdentity({ signal: new AbortController().signal }),
      () => incident.listMembers({ signal: new AbortController().signal }),
    ]) {
      await expect(load()).resolves.toMatchObject({
        kind: "rejected",
        failure: { kind: "invalid_contract" },
      });
    }
    await expect(
      incident.getIdentity({ signal: new AbortController().signal }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "authorization_lost" },
    });
  });
});

describe("Workbook startup adapter", () => {
  it("derives the exact projected query and accepts correlated availability", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(envelope(startupResource())));
    vi.stubGlobal("fetch", fetchMock);
    const startup = createWorkbookStartupAdapter({
      apiBase: "/base/",
      incidentId,
    });

    await expect(
      startup.load({
        query: {
          extensionProfileId: "network_flow_activity",
          sheetRefId: "network_analysis",
          sheetRefKind: "extension_workspace",
          viewSchemaId: hostsViewSchemaId,
        },
        signal: new AbortController().signal,
      }),
    ).resolves.toMatchObject({
      kind: "accepted",
      value: {
        availability: {
          workspaces: [
            {
              extensionProfileId: "network_flow_activity",
              workspaceKey: "network_analysis",
            },
          ],
        },
        selection: {
          selectedSheetRef: { kind: "view_schema", id: hostsViewSchemaId },
        },
      },
    });
    expect(fetchMock).toHaveBeenCalledWith(
      `/base/api/v1/incidents/${incidentId}/workbook-startup?extension_profile_id=network_flow_activity&sheet_ref_id=network_analysis&sheet_ref_kind=extension_workspace&view_schema_id=${encodeURIComponent(hostsViewSchemaId)}`,
      expect.objectContaining({ method: "GET" }),
    );
  });

  it("rejects malformed and mismatched startup resources and contains aborts", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(envelope({ incident_id: incidentId })))
      .mockResolvedValueOnce(
        response(envelope(startupResource(otherIncidentId))),
      )
      .mockImplementationOnce(
        (_input: RequestInfo | URL, init?: RequestInit) =>
          new Promise<Response>((_resolve, reject) => {
            init?.signal?.addEventListener("abort", () => {
              reject(new DOMException("Aborted", "AbortError"));
            });
          }),
      );
    vi.stubGlobal("fetch", fetchMock);
    const startup = createWorkbookStartupAdapter({
      apiBase: undefined,
      incidentId,
    });
    const input = {
      query: {},
      signal: new AbortController().signal,
    };

    await expect(startup.load(input)).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    await expect(startup.load(input)).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    const controller = new AbortController();
    const pending = startup.load({ query: {}, signal: controller.signal });
    controller.abort();
    await expect(pending).resolves.toEqual({ kind: "aborted" });
  });
});

describe("Workbook preference adapter", () => {
  it("uses PUT plus CSRF and accepts only exact incident and sheet correlations", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=csrf-token",
    );
    const sheetRef = { kind: "view_schema" as const, id: hostsViewSchemaId };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        response(
          envelope({
            created_at: now,
            home_sheet_ref: sheetRef,
            incident_id: incidentId,
            updated_at: now,
            user_id: userId,
          }),
        ),
      )
      .mockResolvedValueOnce(
        response(
          envelope({
            created_at: now,
            default_sheet_ref: sheetRef,
            incident_id: incidentId,
            updated_at: now,
            updated_by_user_id: userId,
          }),
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const preferences = createWorkbookPreferenceAdapter({
      apiBase: "/base",
      incidentId,
    });

    await expect(
      preferences.setHomeSheet({
        sheetRef,
        signal: new AbortController().signal,
      }),
    ).resolves.toEqual({ kind: "accepted", value: undefined });
    await expect(
      preferences.setDefaultSheet({
        sheetRef,
        signal: new AbortController().signal,
      }),
    ).resolves.toEqual({ kind: "accepted", value: undefined });

    for (const [call, route, requestKey] of [
      [1, "me", "home_sheet_ref"],
      [2, "default", "default_sheet_ref"],
    ] as const) {
      const [, init] = fetchMock.mock.calls[call - 1] as [string, RequestInit];
      expect(fetchMock).toHaveBeenNthCalledWith(
        call,
        `/base/api/v1/incidents/${incidentId}/workbook-preferences/${route}`,
        expect.objectContaining({ method: "PUT" }),
      );
      expect(new Headers(init.headers).get("X-CSRF-Token")).toBe("csrf-token");
      expect(JSON.parse(String(init.body))).toEqual({
        [requestKey]: sheetRef,
      });
    }
  });

  it("rejects malformed, cross-incident, and mismatched preference acknowledgements", async () => {
    const requested = {
      kind: "view_schema" as const,
      id: hostsViewSchemaId,
    };
    const otherSheet = {
      kind: "view_schema" as const,
      id: "cartulary.view.timeline.v2",
    };
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(response(envelope({ incident_id: incidentId })))
      .mockResolvedValueOnce(
        response(
          envelope({
            created_at: now,
            home_sheet_ref: requested,
            incident_id: otherIncidentId,
            updated_at: now,
            user_id: userId,
          }),
        ),
      )
      .mockResolvedValueOnce(
        response(
          envelope({
            created_at: now,
            default_sheet_ref: otherSheet,
            incident_id: incidentId,
            updated_at: now,
            updated_by_user_id: userId,
          }),
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const preferences = createWorkbookPreferenceAdapter({
      apiBase: undefined,
      incidentId,
    });

    await expect(
      preferences.setHomeSheet({
        sheetRef: requested,
        signal: new AbortController().signal,
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    await expect(
      preferences.setHomeSheet({
        sheetRef: requested,
        signal: new AbortController().signal,
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
    await expect(
      preferences.setDefaultSheet({
        sheetRef: requested,
        signal: new AbortController().signal,
      }),
    ).resolves.toMatchObject({
      kind: "rejected",
      failure: { kind: "invalid_contract" },
    });
  });
});
