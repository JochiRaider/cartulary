import {
  gridShellTestId,
  rowInspectButtonTestId,
} from "@cartulary/ui-contracts";
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
  let timelineRows: Array<{
    record_id: string;
    row_version: number;
    cells: Record<string, { value: unknown }>;
  }>;
  let uploadShouldFail: boolean;

  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    evidenceRows = [];
    timelineRows = [];
    uploadShouldFail = false;
    fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
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
      if (
        method === "POST" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/views/${evidenceViewSchemaId}/rows`,
        )
      ) {
        return successEnvelope(
          {
            view_schema_id: evidenceViewSchemaId,
            change_set_id: "change-evidence",
            row: evidenceRow("evidence-created", 1, "Attached screenshot"),
          },
          201,
        );
      }
      if (method === "POST" && url.endsWith("/api/v1/object-blobs")) {
        return successEnvelope(
          {
            object_blob_id: "blob-created",
            upload_target: {
              href: "/api/v1/object-uploads/test-token",
              method: "PUT",
            },
          },
          201,
        );
      }
      if (
        method === "PUT" &&
        url.endsWith("/api/v1/object-uploads/test-token")
      ) {
        return new Response(null, { status: uploadShouldFail ? 500 : 200 });
      }
      if (
        method === "POST" &&
        url.endsWith("/api/v1/evidence-records/evidence-created/attach-blob")
      ) {
        return successEnvelope({
          record_id: "evidence-created",
          row_version: 2,
        });
      }
      if (method === "PATCH" && url.endsWith("/api/v1/records/timeline-1")) {
        const row = timelineRow("timeline-1", 2, "Selected row", 1);
        timelineRows = [row];
        return successEnvelope({
          view_schema_id: timelineViewSchemaId,
          change_set_id: "change-timeline",
          row,
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
            rows:
              viewSchemaId === evidenceViewSchemaId
                ? evidenceRows
                : viewSchemaId === timelineViewSchemaId
                  ? timelineRows
                  : [],
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

  it("Phase 4 U-4-WB-03 issues opaque evidence preview and download handles from the evidence surface", async () => {
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

  it("Phase 5 E-5-01 orchestrates selected Timeline evidence attachment inline", async () => {
    timelineRows = [timelineRow("timeline-1", 1, "Selected row", 0)];

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(await screen.findByTestId("surface-tab-timeline"));
    fireEvent.click(
      await screen.findByTestId(rowInspectButtonTestId("timeline-1")),
    );
    const input = await screen.findByTestId(
      "timeline-evidence-file-timeline-1",
    );
    fireEvent.change(input, {
      target: {
        files: [
          new File(["screenshot body"], "screenshot.txt", {
            type: "text/plain",
          }),
        ],
      },
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${evidenceViewSchemaId}/rows`),
        expect.objectContaining({ method: "POST" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/object-blobs"),
        expect.objectContaining({ method: "POST" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/object-uploads/test-token"),
        expect.objectContaining({ method: "PUT" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(
          "/api/v1/evidence-records/evidence-created/attach-blob",
        ),
        expect.objectContaining({ method: "POST" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/records/timeline-1"),
        expect.objectContaining({ method: "PATCH" }),
      );
    });
    expect(
      (await screen.findByTestId("timeline-inspector-message")).textContent,
    ).toBe("Evidence attached.");
  });

  it("Phase 5 E-5-01 surfaces upload failures inline without issuing Timeline patches", async () => {
    uploadShouldFail = true;
    timelineRows = [timelineRow("timeline-1", 1, "Selected row", 0)];

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(await screen.findByTestId("surface-tab-timeline"));
    fireEvent.click(
      await screen.findByTestId(rowInspectButtonTestId("timeline-1")),
    );
    fireEvent.change(
      await screen.findByTestId("timeline-evidence-file-timeline-1"),
      {
        target: {
          files: [
            new File(["screenshot body"], "screenshot.txt", {
              type: "text/plain",
            }),
          ],
        },
      },
    );

    await waitFor(() => {
      expect(screen.getByTestId("timeline-inspector-message").textContent).toBe(
        "upload_failed_500",
      );
    });
    expect(
      fetchMock.mock.calls.some(([input, init]) => {
        return (
          String(input).endsWith("/api/v1/records/timeline-1") &&
          ((init as RequestInit | undefined)?.method ?? "GET") === "PATCH"
        );
      }),
    ).toBe(false);
  });
});

function timelineRow(
  recordId: string,
  rowVersion: number,
  summary: string,
  evidenceCount: number,
) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "timeline.occurred_at": { value: "" },
      "timeline.summary": { value: summary },
      "timeline.details": { value: "" },
      "timeline.source_text": { value: "" },
      "timeline.host_refs": {
        value: { kind: "collection_value_v1", ordered: true, items: [] },
      },
      "timeline.identity_refs": {
        value: { kind: "collection_value_v1", ordered: true, items: [] },
      },
      "timeline.evidence_count": { value: evidenceCount },
      "timeline.tags": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "timeline.edited_at": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.recorded_at": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.sort_ts": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.capture_state": { value: "rough" },
      "timeline.replacement_record_id": { value: null },
      "timeline.occurred_day": { value: null },
      "timeline.recorded_day": { value: "2026-04-24" },
      "timeline.has_evidence": { value: evidenceCount > 0 },
      "timeline.attached_evidence_ids": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "timeline.has_unresolved_mentions": { value: false },
    },
  };
}

function evidenceRow(recordId: string, rowVersion: number, title: string) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "evidence.title": { value: title },
      "evidence.lifecycle_state": { value: "requested" },
      "evidence.requested_at": { value: null },
      "evidence.received_at": { value: null },
      "evidence.storage_ref": { value: "" },
      "evidence.blob_hash": { value: "" },
      "evidence.collector_party_text": { value: "Workbook upload" },
      "evidence.source_party_text": { value: "" },
      "evidence.upload_state": { value: "pending" },
      "evidence.linked_record_count": { value: 0 },
      "evidence.edited_at": { value: null },
    },
  };
}

describe("generic workbook mutation payloads", () => {
  it("Phase 4 U-4-WB-04 builds required creates with direct values, timestamps, and explicit clears", () => {
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

  it("Phase 4 U-4-WB-05 builds direct clears and typed collection actions", () => {
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
        actions: [{ op: "add_tag", tag_name: "urgent" }],
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
