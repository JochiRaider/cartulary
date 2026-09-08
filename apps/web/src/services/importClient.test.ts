import { afterEach, describe, expect, it, vi } from "vitest";
import {
  captureImportWrite,
  captureWorkbookUpload,
} from "../imports/importRequests";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { jsonResponse } from "../testing/fetchMockTestSupport";
import {
  importTestIds as ids,
  importTestApprovedUnit,
  importTestEnvelope,
  importTestJob,
  importTestMapping,
  importTestPreview,
  importTestScope,
  importTestUnit,
} from "../testing/workbookImportTestSupport";
import { ImportClient, importFailureMessage } from "./importClient";
import { validWorkbookImportJob } from "./importJobContract";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
const signal = () => new AbortController().signal;
const client = () =>
  new ImportClient({
    availability: readyExtensionAvailability(ids.incident),
    incidentId: ids.incident,
  });

describe("Workbook import typed transport", () => {
  it("captures exact upload bytes and secure identity once for an exact replay", async () => {
    const fetch = vi
      .fn()
      .mockImplementation(async () =>
        jsonResponse(importTestEnvelope(importTestJob()), 202),
      );
    vi.stubGlobal("fetch", fetch);
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=csrf-test",
    );
    const attempt = captureWorkbookUpload(
      importTestScope,
      new File(["original bytes"], "one.csv", { type: "invalid/type" }),
    );
    const service = client();
    expect((await service.send(attempt, signal())).kind).toBe("accepted");
    expect((await service.send(attempt, signal(), true)).kind).toBe("accepted");
    const bodies = fetch.mock.calls.map((call) => call[1].body as FormData);
    const read = (blob: Blob) =>
      new Promise((resolve) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result);
        reader.readAsText(blob);
      });
    expect(await read(bodies[0]?.get("file") as Blob)).toBe("original bytes");
    expect(await read(bodies[1]?.get("metadata") as Blob)).toBe(
      await read(bodies[0]?.get("metadata") as Blob),
    );
    expect(attempt.file.type).toBe("application/octet-stream");
    expect(
      new Headers(fetch.mock.calls[0]?.[1].headers).get("X-CSRF-Token"),
    ).toBe("csrf-test");
    expect(
      captureWorkbookUpload(importTestScope, new File(["new"], "two.csv"))
        .metadata.client_txn_id,
    ).not.toBe(attempt.metadata.client_txn_id);
  });

  it("separates acceptance from resource loading and retains complete selection receipts", async () => {
    const unit = importTestApprovedUnit();
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse(
          importTestEnvelope({
            import_session_id: ids.session,
            session_status: "ready_to_apply",
            selected_unit_ids: [ids.unit],
            unit,
          }),
        ),
      ),
    );
    const result = await client().send(
      captureImportWrite({
        kind: "select",
        scope: importTestScope,
        sessionId: ids.session,
        unitId: ids.unit,
        body: { client_txn_id: "selection" },
      }),
      signal(),
    );
    expect(result).toMatchObject({
      kind: "accepted",
      receipt: {
        kind: "selection",
        selection: { selected_unit_ids: [ids.unit], unit },
      },
    });
  });

  it("keeps malformed success and server uncertainty distinct from public rejection", async () => {
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const attempt = captureWorkbookUpload(
      importTestScope,
      new File(["a"], "a.csv"),
    );
    const service = client();
    for (const status of [200, 500, 408]) {
      fetch.mockResolvedValue(
        jsonResponse(
          status === 200 ? importTestEnvelope(importTestJob()) : {},
          status,
        ),
      );
      expect((await service.send(attempt, signal())).kind).toBe("uncertain");
    }
    fetch.mockResolvedValue(
      jsonResponse(
        {
          error: {
            code: "invalid_import_request",
            status: 400,
            message: "private internal content",
            details: {
              field: "source_columns",
              reason_code: "duplicate_target_field",
            },
          },
        },
        400,
      ),
    );
    const result = await service.send(attempt, signal());
    expect(result).toMatchObject({
      kind: "rejected",
      failure: { field: "source_columns", reason: "duplicate_target_field" },
    });
    if (result.kind !== "accepted")
      expect(importFailureMessage(result.failure)).toBe(
        "Each target field can be mapped only once.",
      );
  });

  it("preserves authentication authorization conflict and resource denial without private exceptions", async () => {
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const service = client();
    for (const status of [401, 403, 404, 409]) {
      fetch.mockResolvedValue(new Response("not JSON", { status }));
      const result = await service.readJob(ids.job, signal());
      expect(result).toMatchObject({ kind: "failed", failure: { status } });
    }
    fetch.mockRejectedValue(new Error("stack trace /home/private/secret"));
    const result = await service.readJob(ids.job, signal());
    expect(JSON.stringify(result)).not.toContain("secret");
  });

  it("rejects foreign jobs contradictory mappings and cancel responses", async () => {
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const service = client();
    fetch.mockResolvedValue(
      jsonResponse(
        importTestEnvelope(
          importTestJob("running", { job_id: ids.secondUnit }),
        ),
      ),
    );
    expect((await service.readJob(ids.job, signal())).kind).toBe("failed");
    fetch.mockResolvedValue(
      jsonResponse(
        importTestEnvelope(
          importTestUnit({ mapping_fingerprint: "a".repeat(64) }),
        ),
      ),
    );
    expect((await service.readUnit(ids.session, ids.unit, signal())).kind).toBe(
      "failed",
    );
    fetch.mockResolvedValue(
      jsonResponse(
        importTestEnvelope(importTestUnit({ unit_status: "ready" })),
      ),
    );
    expect((await service.readUnit(ids.session, ids.unit, signal())).kind).toBe(
      "failed",
    );
    fetch.mockResolvedValue(
      jsonResponse(
        importTestEnvelope(
          importTestJob("cancel_requested", { cancelable: true }),
        ),
      ),
    );
    expect(
      (
        await service.send(
          captureImportWrite({
            kind: "cancel",
            scope: importTestScope,
            jobId: ids.job,
            body: { client_txn_id: "cancel" },
          }),
          signal(),
        )
      ).kind,
    ).toBe("uncertain");
    expect(
      validWorkbookImportJob(
        importTestJob("succeeded", { retained_until: "2026-09-09T10:00:02Z" }),
        ids.incident,
      ),
    ).toBe(false);
    expect(
      validWorkbookImportJob(
        importTestJob("succeeded"),
        ids.incident,
        ids.job,
        ids.session,
        "apply",
      ),
    ).toBe(false);
    expect(
      validWorkbookImportJob(
        importTestJob("succeeded"),
        ids.incident,
        ids.job,
        ids.session,
        "discovery",
      ),
    ).toBe(true);
  });

  it("preserves opaque paging and accepts successful empty discovery", async () => {
    const page = (
      units: unknown[],
      has_more: boolean,
      next_cursor: string | null,
    ) => ({
      data: { import_units: units },
      meta: {
        request_id: "test",
        paging: { limit: 50, has_more, next_cursor },
      },
    });
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(
        jsonResponse(page([importTestUnit()], true, "opaque /+ cursor")),
      )
      .mockResolvedValueOnce(
        jsonResponse(
          page(
            [importTestUnit({ import_unit_id: ids.secondUnit })],
            false,
            null,
          ),
        ),
      );
    vi.stubGlobal("fetch", fetch);
    const service = client();
    expect(await service.listUnits(ids.session, signal())).toMatchObject({
      kind: "received",
      value: [{ import_unit_id: ids.unit }, { import_unit_id: ids.secondUnit }],
    });
    expect(String(fetch.mock.calls[1]?.[0])).toContain(
      "cursor_token=opaque%20%2F%2B%20cursor",
    );
    fetch.mockResolvedValue(jsonResponse(page([], false, null)));
    expect(await service.listUnits(ids.session, signal())).toEqual({
      kind: "received",
      value: [],
    });
    for (const body of [
      page([importTestUnit()], true, "repeated"),
      page([importTestUnit()], false, "bad"),
      page(
        [importTestUnit({ import_session_id: ids.secondUnit })],
        false,
        null,
      ),
    ]) {
      fetch.mockResolvedValue(jsonResponse(body));
      expect((await service.listUnits(ids.session, signal())).kind).toBe(
        "failed",
      );
    }
  });

  it("rejects mismatched and oversized previews and snapshots mapping input", async () => {
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const service = client();
    const preview = importTestPreview();
    for (const p of [
      { ...preview, import_unit_id: ids.secondUnit },
      { ...preview, preview_rows: Array(51).fill(preview.preview_rows[0]) },
    ]) {
      fetch.mockResolvedValue(jsonResponse(importTestEnvelope(p)));
      expect((await service.preview(importTestUnit(), signal())).kind).toBe(
        "failed",
      );
    }
    const body = {
      ...JSON.parse(JSON.stringify(importTestMapping)),
      client_txn_id: "mapping",
      header_row_ref: 1,
      data_start_row_ref: 2,
    };
    const attempt = captureImportWrite({
      kind: "mapping",
      scope: importTestScope,
      sessionId: ids.session,
      unitId: ids.unit,
      body,
    });
    body.source_columns[0].field_key = null;
    expect(attempt.body.source_columns[0]?.field_key).toBe(
      "timeline.activity_synopsis_text",
    );
  });
});
