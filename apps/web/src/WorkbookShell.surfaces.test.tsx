import { gridShellTestId } from "@cartulary/test-utils";
import {
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { successEnvelope } from "./timelineWorkbookTestSupport";
import {
  buildGenericCreatePayload,
  buildGenericPatchChange,
  WorkbookShell,
} from "./WorkbookShell";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const hostsViewSchemaId = "cartulary.view.hosts.v1";
const identitiesViewSchemaId = "cartulary.view.identities.v1";
const evidenceViewSchemaId = "cartulary.view.evidence.v1";
const indicatorsViewSchemaId = "cartulary.view.indicators.v1";

function requireField(
  contract: ViewContract,
  fieldKey: string,
): ViewFieldContract {
  const field = contract.fieldMap[fieldKey];
  if (!field) {
    throw new Error(`Missing field ${fieldKey} on ${contract.viewSchemaId}`);
  }
  return field;
}

describe("WorkbookShell surface selection", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let evidenceRows: Array<{
    record_id: string;
    row_version: number;
    cells: Record<string, { value: unknown }>;
  }>;

  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    evidenceRows = [];
    fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [{ incident_id: "incident-1", role: "admin" }],
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1")) {
        return successEnvelope({
          incident_id: "incident-1",
          incident_key: "IR-1",
          title: "Incident 1",
          description: null,
          severity: null,
          tlp: null,
          current_phase: null,
          primary_external_case_ref: null,
          incident_version: 1,
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1/memberships")) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "incident-1",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-preferences/")) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (url.endsWith("/api/v1/evidence-records/evidence-1/preview-handle")) {
        return successEnvelope({
          href: "/api/v1/evidence-handles/preview-token",
          method: "GET",
          filename: "evidence.txt",
          preview_kind: "text_inline",
          content_type: "text/plain",
        });
      }
      if (url.endsWith("/api/v1/evidence-records/evidence-1/download-handle")) {
        return successEnvelope({
          href: "/api/v1/evidence-handles/download-token",
          method: "GET",
          filename: "evidence.txt",
          content_type: "text/plain",
        });
      }
      for (const viewSchemaId of [
        timelineViewSchemaId,
        hostsViewSchemaId,
        identitiesViewSchemaId,
        evidenceViewSchemaId,
        indicatorsViewSchemaId,
      ]) {
        if (url.includes(`/views/${viewSchemaId}/query`)) {
          return successEnvelope({
            incident_id: "incident-1",
            view_schema_id: viewSchemaId,
            rows: viewSchemaId === evidenceViewSchemaId ? evidenceRows : [],
          });
        }
      }
      return successEnvelope({});
    });
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal(
      "WebSocket",
      class {
        onmessage: ((event: MessageEvent) => void) | null = null;

        close() {}
      } as unknown as typeof WebSocket,
    );
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("selects required built-in and system view surfaces by view_schema_id", async () => {
    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(await screen.findByTestId("surface-tab-evidence"));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${evidenceViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(
      screen.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeTruthy();

    fireEvent.change(screen.getByTestId("system-view-selector"), {
      target: { value: indicatorsViewSchemaId },
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${indicatorsViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(window.location.search).toContain(
      `view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`,
    );
  });

  it("issues opaque evidence preview and download handles from the evidence surface", async () => {
    evidenceRows = [
      {
        record_id: "evidence-1",
        row_version: 4,
        cells: {
          "evidence.title": { value: "EDR package" },
          "evidence.lifecycle_state": { value: "available" },
          "evidence.requested_at": { value: null },
          "evidence.received_at": { value: null },
          "evidence.storage_ref": { value: "slot" },
          "evidence.blob_hash": { value: "sha" },
          "evidence.collector_party_text": { value: "IR" },
          "evidence.source_party_text": { value: "Endpoint" },
          "evidence.upload_state": { value: "available" },
          "evidence.linked_record_count": { value: 0 },
          "evidence.edited_at": { value: null },
        },
      },
    ];
    const anchorClick = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => undefined);

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(await screen.findByTestId("surface-tab-evidence"));
    fireEvent.click(await screen.findByTestId("evidence-preview-evidence-1"));

    const frame = await screen.findByTestId(
      "evidence-preview-frame-evidence-1",
    );
    expect(frame.getAttribute("src")).toBe(
      "/api/v1/evidence-handles/preview-token",
    );
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining(
        "/api/v1/evidence-records/evidence-1/preview-handle",
      ),
      expect.objectContaining({ method: "POST", body: "{}" }),
    );

    fireEvent.click(screen.getByTestId("evidence-download-evidence-1"));

    await waitFor(() => {
      expect(anchorClick).toHaveBeenCalled();
    });
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining(
        "/api/v1/evidence-records/evidence-1/download-handle",
      ),
      expect.objectContaining({ method: "POST", body: "{}" }),
    );
  });
});

describe("generic workbook mutation payloads", () => {
  it("builds required creates with direct values, timestamps, and explicit clears", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);

    expect(
      buildGenericCreatePayload(evidence, {}, "txn-evidence-missing"),
    ).toBeNull();
    expect(
      buildGenericCreatePayload(
        evidence,
        {
          "evidence.title": " Endpoint package ",
          "evidence.requested_at": "2026-04-24T12:00:00Z",
          "evidence.collector_party_id": "",
        },
        "txn-evidence-create",
      ),
    ).toMatchObject({
      client_txn_id: "txn-evidence-create",
      "evidence.title": "Endpoint package",
      "evidence.requested_at": "2026-04-24T12:00:00Z",
      "evidence.collector_party_id": null,
    });
  });

  it("builds direct clears and typed collection actions", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);
    const notes = requireViewContract("cartulary.view.notes.v1");
    const commLog = requireViewContract("cartulary.view.comm_log.v1");
    const handoff = requireViewContract("cartulary.view.handoff.v1");
    const decisions = requireViewContract("cartulary.view.decisions.v1");

    expect(
      buildGenericPatchChange(
        requireField(evidence, "evidence.source_party_id"),
        "",
      ),
    ).toEqual({ field_key: "evidence.source_party_id", value: null });
    expect(
      buildGenericPatchChange(requireField(notes, "note.tags"), " urgent "),
    ).toEqual({
      field_key: "note.tags",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "add_token", raw_text: "urgent" }],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(commLog, "comm_log.audience_party_ids"),
        "party-1",
      ),
    ).toEqual({
      field_key: "comm_log.audience_party_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "add_party_ref", party_id: "party-1" }],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(decisions, "decision.support_refs"),
        "record-1",
      ),
    ).toEqual({
      field_key: "decision.support_refs",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "add_record_ref", linked_record_id: "record-1" }],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(handoff, "handoff.open_risk_refs"),
        "risk_ref:abc",
        "remove",
      ),
    ).toEqual({
      field_key: "handoff.open_risk_refs",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "remove_risk_ref", item_ref: "risk_ref:abc" }],
      },
    });
  });
});
