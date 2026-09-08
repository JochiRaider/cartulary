import { afterEach, describe, expect, it, vi } from "vitest";
import {
  networkFlowApprovalRequest,
  networkFlowApprovedPreviewMatches,
} from "../networkFlow/networkFlowImportModel";
import { ImportClient, importFailureMessage } from "../services/importClient";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { errorResponse, jsonResponse } from "../testing/fetchMockTestSupport";
import {
  importTestEnvelope as envelope,
  importTestIds as ids,
  importTestApprovedUnit,
  importTestJob,
  importTestPreview,
  importTestScope,
  importTestSession,
  importTestUnit,
} from "../testing/workbookImportTestSupport";
import {
  loadImportResources,
  observeImportJob,
  submitImportAttempt,
} from "./importCoordinator";
import { captureImportWrite, captureWorkbookUpload } from "./importRequests";

const candidate = {
  target_kind: "network_flow_table",
  extension_profile_id: "network_flow_activity",
  owner_mapping_schema_id: "cartulary.network_flow.mapping_candidate.v1",
  owner_mapping: { mapping_kind: "characterized" },
};
const discovery = {
  sessionId: ids.session,
  unit: importTestUnit(),
  preview: importTestPreview(),
};
const approved = (fingerprint = "a".repeat(64)) =>
  importTestApprovedUnit({
    mapping_fingerprint: fingerprint,
    approved_mapping: {
      ...candidate,
      source_columns: [
        {
          source_column_ordinal: 1,
          source_header_text: "Activity Synopsis",
          field_key: null,
          entity_binding_mode: null,
          transform_id: null,
          transform_options: {},
          empty_value_policy: "omit_field",
        },
      ],
    },
  });
const service = () =>
  new ImportClient({
    availability: readyExtensionAvailability(ids.incident),
    incidentId: ids.incident,
  });
const signal = () => new AbortController().signal;
afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("extension import coordinator stages", () => {
  it("keeps discovery, side-effect-free preview, approval, selection, and apply as explicit stages", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=import-csrf",
    );
    const fetch = vi
      .fn()
      .mockImplementation(
        async (input: RequestInfo | URL, init?: RequestInit) => {
          const url = String(input);
          if (url.endsWith("/import-sessions") || url.endsWith("/apply"))
            return jsonResponse(envelope(importTestJob()), 202);
          if (url.includes("/jobs/"))
            return jsonResponse(envelope(importTestJob("succeeded")));
          if (url.endsWith(`/import-sessions/${ids.session}`))
            return jsonResponse(envelope(importTestSession()));
          if (url.includes("/units?"))
            return jsonResponse({
              data: { import_units: [importTestUnit()] },
              meta: {
                request_id: "test",
                paging: { limit: 50, has_more: false, next_cursor: null },
              },
            });
          if (url.endsWith("/mapping-preview"))
            return jsonResponse(
              envelope({
                schema_id:
                  "cartulary.imports.extension_mapping_preview_result.v1",
                import_session_id: ids.session,
                import_unit_id: ids.unit,
                target_kind: candidate.target_kind,
                extension_profile_id: candidate.extension_profile_id,
                owner_result_schema_id:
                  "cartulary.network_flow.import_preview_result.v1",
                owner_result: { mapping_fingerprint: "a".repeat(64) },
              }),
            );
          if (url.endsWith("/mapping")) {
            expect(new Headers(init?.headers).get("X-CSRF-Token")).toBe(
              "import-csrf",
            );
            return jsonResponse(envelope(approved()));
          }
          if (url.endsWith("/select"))
            return jsonResponse(
              envelope({
                import_session_id: ids.session,
                session_status: "ready_to_apply",
                selected_unit_ids: [ids.unit],
                unit: approved(),
              }),
            );
          return jsonResponse(envelope(importTestPreview()));
        },
      );
    vi.stubGlobal("fetch", fetch);
    const client = service();
    const receipt = await submitImportAttempt({
      signal: signal(),
      upload: true,
      send: (s) =>
        client.send(
          captureWorkbookUpload(
            importTestScope,
            new File(["data"], "source.csv"),
          ),
          s,
        ),
    });
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(receipt.kind).toBe("accepted");
    if (receipt.kind !== "accepted" || receipt.receipt.kind !== "job")
      throw new Error("Missing fixture job");
    const job = await observeImportJob({
      initial: receipt.receipt.job,
      signal: signal(),
      read: (id, s) => client.readJob(id, s),
      onJob: vi.fn(),
    });
    expect(job.kind).toBe("terminal");
    expect(
      (await loadImportResources(client, ids.session, signal())).kind,
    ).toBe("received");
    const preview = await client.previewMapping(
      discovery.unit,
      candidate,
      signal(),
    );
    expect(preview.kind).toBe("received");
    const previewCall = fetch.mock.calls.find(([url]) =>
      String(url).endsWith("/mapping-preview"),
    );
    expect(JSON.parse(String(previewCall?.[1]?.body))).toEqual(candidate);
    expect(
      fetch.mock.calls.some(([url]) => String(url).endsWith("/mapping")),
    ).toBe(false);
    const mapping = await client.send(
      captureImportWrite({
        kind: "mapping",
        scope: importTestScope,
        sessionId: ids.session,
        unitId: ids.unit,
        body: networkFlowApprovalRequest(discovery, candidate, "mapping"),
      }),
      signal(),
    );
    expect(mapping.kind).toBe("accepted");
    expect(
      (
        await client.send(
          captureImportWrite({
            kind: "select",
            scope: importTestScope,
            sessionId: ids.session,
            unitId: ids.unit,
            body: { client_txn_id: "select" },
          }),
          signal(),
        )
      ).kind,
    ).toBe("accepted");
    expect(
      (
        await client.send(
          captureImportWrite({
            kind: "apply",
            scope: importTestScope,
            sessionId: ids.session,
            body: { client_txn_id: "apply", selected_unit_ids: [ids.unit] },
          }),
          signal(),
        )
      ).kind,
    ).toBe("accepted");
  });
  it("blocks selection and apply when durable approval returns a stale fingerprint", async () => {
    const fetch = vi
      .fn()
      .mockResolvedValue(jsonResponse(envelope(approved("b".repeat(64)))));
    vi.stubGlobal("fetch", fetch);
    const result = await service().send(
      captureImportWrite({
        kind: "mapping",
        scope: importTestScope,
        sessionId: ids.session,
        unitId: ids.unit,
        body: networkFlowApprovalRequest(discovery, candidate, "mapping"),
      }),
      signal(),
    );
    expect(result.kind).toBe("accepted");
    if (result.kind === "accepted" && result.receipt.kind === "unit")
      expect(
        networkFlowApprovedPreviewMatches(result.receipt.unit, "a".repeat(64)),
      ).toBe(false);
    expect(fetch).toHaveBeenCalledTimes(1);
  });
  it("creates and previews a durable operator-selected region", async () => {
    const unit = importTestUnit({
      import_unit_id: ids.secondUnit,
      locator_kind: "operator_region",
      locator: { sheet_name: "Sheet1" },
    });
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse(envelope(unit), 201))
      .mockResolvedValueOnce(jsonResponse(envelope(importTestPreview(unit))));
    vi.stubGlobal("fetch", fetch);
    const client = service();
    const result = await client.send(
      captureImportWrite({
        kind: "region",
        scope: importTestScope,
        sessionId: ids.session,
        unitId: ids.unit,
        body: {
          client_txn_id: "region",
          source_rect: {
            start_row: 1,
            start_column: 1,
            end_row: 2,
            end_column: 1,
          },
        },
      }),
      signal(),
    );
    expect(result).toMatchObject({
      kind: "accepted",
      receipt: { kind: "unit", unit: { import_unit_id: ids.secondUnit } },
    });
    expect(fetch).toHaveBeenCalledTimes(1);
    expect((await client.preview(unit, signal())).kind).toBe("received");
    expect(String(fetch.mock.calls[1]?.[0])).toContain(ids.secondUnit);
  });
  it("preserves public preview errors and creates no durable mapping", async () => {
    const fetch = vi.fn().mockResolvedValue(
      errorResponse("invalid_import_mapping", 422, {
        reason_code: "required_field_missing",
      }),
    );
    vi.stubGlobal("fetch", fetch);
    const result = await service().previewMapping(
      discovery.unit,
      candidate,
      signal(),
    );
    expect(result).toMatchObject({
      kind: "failed",
      failure: {
        code: "invalid_import_mapping",
        reason: "required_field_missing",
        status: 422,
      },
    });
    expect(fetch).toHaveBeenCalledTimes(1);
  });
  it("sanitizes unsafe import transport errors through the shared public error view", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockRejectedValue(new Error("stack trace /home/service/secret")),
    );
    const result = await service().send(
      captureWorkbookUpload(importTestScope, new File(["data"], "source.csv")),
      signal(),
    );
    expect(result.kind).toBe("uncertain");
    if (result.kind !== "accepted")
      expect(importFailureMessage(result.failure)).not.toContain("secret");
  });
});
