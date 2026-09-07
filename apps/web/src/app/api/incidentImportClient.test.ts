import { afterEach, describe, expect, it, vi } from "vitest";
import {
  errorResponse,
  jsonResponse,
  readHeader,
} from "../../testing/fetchMockTestSupport";
import {
  importActorID,
  importedIncidentID,
  importJob,
  importJobID,
  jobEnvelope,
  readTestBlob,
} from "../../testing/incidentImportTestSupport";
import {
  cancelImportJob,
  captureImport,
  importedIncidentTarget,
  importIncidentBundle,
  importJobAdvances,
  readImportJob,
} from "./incidentImportClient";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
describe("incident import client", () => {
  it("preserves exact multipart bytes metadata and csrf on admission and replay", async () => {
    const fetch = vi
      .fn()
      .mockImplementation(async () => jsonResponse(jobEnvelope(), 202));
    vi.stubGlobal("fetch", fetch);
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=import-csrf",
    );
    const file = new File(["exact\u0000bytes\r\n"], "archive.zip", {
      type: "application/zip",
    });
    const attempt = captureImport(file, "txn-captured");
    expect(attempt.file).toBe(file);
    const signal = new AbortController().signal;
    for (let replay = 0; replay < 2; replay++) {
      expect((await importIncidentBundle(attempt, signal)).ok).toBe(true);
      const [path, init] = fetch.mock.calls[replay] ?? [];
      expect(path).toBe("/api/v1/incident-bundles/import");
      expect(init.signal).toBe(signal);
      expect(init.credentials).toBe("include");
      expect(readHeader(init, "X-CSRF-Token")).toBe("import-csrf");
      expect(readHeader(init, "Content-Type")).toBe("");
      const body = init.body as FormData;
      expect([...body.keys()].sort()).toEqual(["file", "metadata"]);
      expect(await readTestBlob(body.get("metadata") as Blob)).toBe(
        '{"client_txn_id":"txn-captured"}',
      );
      expect(await readTestBlob(body.get("file") as Blob)).toBe(
        "exact\u0000bytes\r\n",
      );
    }
    const unknown = captureImport(
      new File(["same"], "hint.unknown", { type: "text/plain" }),
      "txn-hint",
    );
    expect(unknown.file.type).toBe("application/octet-stream");
    expect(await readTestBlob(unknown.file)).toBe("same");
  });
  it("requires complete queued or running 202 admission and validates terminal reads", async () => {
    const signal = new AbortController().signal;
    const attempt = captureImport(new File(["archive"], "archive.tar"), "txn");
    for (const status of ["queued", "running"] as const) {
      vi.stubGlobal(
        "fetch",
        vi
          .fn()
          .mockResolvedValue(jsonResponse(jobEnvelope(importJob(status)), 202)),
      );
      expect((await importIncidentBundle(attempt, signal)).ok).toBe(true);
    }
    for (const [payload, status] of [
      [{ data: { job_id: importJobID } }, 202],
      [jobEnvelope(), 200],
      [jobEnvelope(importJob("succeeded")), 202],
      [{ ...jobEnvelope(), meta: {} }, 202],
    ] as const) {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(jsonResponse(payload, status)),
      );
      expect(await importIncidentBundle(attempt, signal)).toMatchObject({
        ok: false,
        payload: { error: { code: "invalid_public_contract_response" } },
      });
    }
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockImplementation(async () =>
          jsonResponse(jobEnvelope(importJob("succeeded")), 202),
        ),
    );
    expect((await importIncidentBundle(attempt, signal, true)).ok).toBe(true);
    for (const status of [
      "queued",
      "running",
      "cancel_requested",
      "succeeded",
      "failed",
      "canceled",
    ] as const) {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(jsonResponse(jobEnvelope(importJob(status)))),
      );
      expect((await readImportJob(importJobID, signal)).ok).toBe(true);
    }
  });
  it("rejects inconsistent jobs and preserves opaque cancellation idempotency", async () => {
    const signal = new AbortController().signal;
    const invalids = [
      importJob("running", { job_id: importActorID }),
      importJob("running", {
        scope: { kind: "incident", incident_id: importedIncidentID },
      }),
      importJob("running", { progress: { completed: 4, total: 2 } }),
      importJob("succeeded", { cancelable: true }),
      importJob("canceled", {
        result_summary: { code: "incorrect", message: "" },
      }),
      importJob("succeeded", { retained_until: "2026-09-08T12:00:01Z" }),
    ];
    for (const job of invalids) {
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(jsonResponse(jobEnvelope(job))),
      );
      expect((await readImportJob(importJobID, signal)).ok).toBe(false);
    }
    const fetch = vi
      .fn()
      .mockResolvedValue(
        jsonResponse(jobEnvelope(importJob("cancel_requested"))),
      );
    vi.stubGlobal("fetch", fetch);
    await cancelImportJob(importJobID, { client_txn_id: "cancel-id" }, signal);
    expect(fetch.mock.calls[0]?.[0]).toBe(`/api/v1/jobs/${importJobID}/cancel`);
    expect(fetch.mock.calls[0]?.[1].body).toBe('{"client_txn_id":"cancel-id"}');
    fetch.mockResolvedValue(errorResponse("job_not_found", 404));
    expect(await readImportJob(importJobID, signal)).toMatchObject({
      ok: false,
      status: 404,
    });
  });
  it("ignores additive result kinds without allowing ambiguous or premature navigation", () => {
    const succeeded = importJob("succeeded");
    const refs = succeeded.result_summary?.resource_refs ?? [];
    expect(importedIncidentTarget(succeeded)).toBe(importedIncidentID);
    expect(
      importedIncidentTarget({
        ...succeeded,
        result_summary: {
          code: "incident_bundle_imported",
          message: "",
          resource_refs: [
            ...refs,
            {
              kind: "future_output",
              id: "opaque",
              route: "/api/v1/future/opaque",
            },
          ],
        },
      }),
    ).toBe(importedIncidentID);
    for (const job of [
      { ...succeeded, status: "failed" as const },
      { ...succeeded, status: "running" as const },
      {
        ...succeeded,
        result_summary: {
          code: "incident_bundle_exported",
          message: "",
          resource_refs: refs,
        },
      },
      {
        ...succeeded,
        result_summary: {
          code: "incident_bundle_imported",
          message: "",
          resource_refs: [...refs, ...refs],
        },
      },
      {
        ...succeeded,
        result_summary: {
          code: "incident_bundle_imported",
          message: "",
          resource_refs: [],
        },
      },
      {
        ...succeeded,
        result_summary: {
          code: "incident_bundle_imported",
          message: "",
          resource_refs: [{ kind: "incident", id: importedIncidentID }],
        },
      },
    ])
      expect(importedIncidentTarget(job)).toBeNull();
    expect(importJobAdvances(importJob(), succeeded)).toBe(true);
    expect(importJobAdvances(succeeded, importJob())).toBe(false);
    expect(
      importJobAdvances(
        importJob("running", { progress: { completed: 3, total: 8 } }),
        importJob("running", { progress: { completed: 2, total: null } }),
      ),
    ).toBe(false);
  });
});
