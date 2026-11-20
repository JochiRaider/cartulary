import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  referencePackFixture,
  referencePackJobFixture,
  referencePackTestJobId,
  referencePackTestValidationRef,
  referencePackValidationFixture,
} from "../testing/referencePackTestSupport";
import {
  captureReferencePackAttempt,
  loadReferencePackJob,
  readReferencePackValidationSummary,
  submitReferencePackAttempt,
  validReferencePackJob,
} from "./referencePacks";

const fetchMock = vi.fn<typeof fetch>();
const signal = () => new AbortController().signal;
const envelope = (data: unknown, status: number) =>
  new Response(JSON.stringify({ data, meta: { request_id: "response" } }), {
    status,
    headers: { "content-type": "application/json" },
  });
beforeEach(() => {
  fetchMock.mockReset();
  vi.stubGlobal("fetch", fetchMock);
});
afterEach(() => vi.unstubAllGlobals());

describe("Reference Pack transport", () => {
  it("validates bounded diagnostic summaries and never follows an arbitrary reference", async () => {
    expect(
      (
        await readReferencePackValidationSummary(
          "https://other.example/private",
          signal(),
        )
      ).kind,
    ).toBe("failed");
    expect(fetchMock).not.toHaveBeenCalled();
    const summary = referencePackValidationFixture();
    fetchMock.mockResolvedValueOnce(envelope(summary, 200));
    expect(
      await readReferencePackValidationSummary(
        referencePackTestValidationRef,
        signal(),
      ),
    ).toEqual({ kind: "read", value: summary });
    expect(String(fetchMock.mock.calls[0]?.[0])).toBe(
      `/api/v1/reference-packs/validation-summaries/${referencePackTestValidationRef}`,
    );
    for (const changes of [
      { issues: [] },
      { retained_issue_count: 2 },
      { total_issue_count: 2 },
      { issues_truncated: true },
      { primary_issue_id: `rpi_${"b".repeat(64)}` },
      { extra: "hostile" },
      { issues: [{ ...summary.issues[0], safe_details: null }] },
    ]) {
      fetchMock.mockResolvedValueOnce(
        envelope({ ...summary, ...changes }, 200),
      );
      expect(
        (
          await readReferencePackValidationSummary(
            referencePackTestValidationRef,
            signal(),
          )
        ).kind,
      ).toBe("failed");
    }
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          error: {
            code: "forbidden",
            message: "Denied",
            retryable: false,
            details: {},
          },
          meta: { request_id: "denied" },
        }),
        { status: 403, headers: { "content-type": "application/json" } },
      ),
    );
    // An invalid error envelope is the client contract failure: its raw 403
    // status must not select the access-failure outcome.
    expect(
      await readReferencePackValidationSummary(
        referencePackTestValidationRef,
        signal(),
      ),
    ).toEqual({
      kind: "failed",
      problem: {
        kind: "contract",
        status: 403,
        code: "invalid_public_contract_response",
      },
    });
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          error: {
            code: "authorization_denied",
            status: 403,
            message: "Denied",
            request_id: "request-test",
            retryable: false,
            details: {},
          },
        }),
        { status: 403, headers: { "content-type": "application/json" } },
      ),
    );
    expect(
      await readReferencePackValidationSummary(
        referencePackTestValidationRef,
        signal(),
      ),
    ).toEqual({ kind: "access_failed", status: 403 });
  });

  it("validates inline and asynchronous activation and disable by successful HTTP status", async () => {
    for (const kind of ["activate", "disable"] as const) {
      const attempt = captureReferencePackAttempt(
        { kind, target: { pack_key: "type_registry.host", pack_version: "1" } },
        `txn-${kind}`,
      );
      fetchMock.mockResolvedValueOnce(
        envelope(
          {
            pack_version: referencePackFixture({
              active: kind === "activate",
              pack_version_state:
                kind === "disable" ? "disabled" : "verified_available",
            }),
          },
          200,
        ),
      );
      expect((await submitReferencePackAttempt(attempt, signal())).kind).toBe(
        "committed",
      );
      fetchMock.mockResolvedValueOnce(envelope(referencePackJobFixture(), 202));
      expect((await submitReferencePackAttempt(attempt, signal())).kind).toBe(
        "accepted",
      );
      fetchMock.mockResolvedValueOnce(envelope(referencePackJobFixture(), 200));
      expect((await submitReferencePackAttempt(attempt, signal())).kind).toBe(
        "uncertain",
      );
      fetchMock.mockResolvedValueOnce(
        envelope({ pack_version: referencePackFixture() }, 202),
      );
      expect((await submitReferencePackAttempt(attempt, signal())).kind).toBe(
        "uncertain",
      );
      fetchMock.mockResolvedValueOnce(
        envelope(
          { pack_version: referencePackFixture({ pack_version: "other" }) },
          200,
        ),
      );
      expect((await submitReferencePackAttempt(attempt, signal())).kind).toBe(
        "uncertain",
      );
    }
  });

  it("keeps all omission separate from deduplicated explicit selected scope during replay", async () => {
    fetchMock.mockResolvedValue(envelope(referencePackJobFixture(), 202));
    const all = captureReferencePackAttempt({ kind: "refresh_all" }, "txn-all");
    await submitReferencePackAttempt(all, signal());
    const selected = captureReferencePackAttempt(
      { kind: "refresh_selected", packKeys: ["outside", "outside", "loaded"] },
      "txn-selected",
    );
    await submitReferencePackAttempt(selected, signal());
    await submitReferencePackAttempt(selected, signal(), true);
    expect(
      fetchMock.mock.calls.map(([, init]) => JSON.parse(String(init?.body))),
    ).toEqual([
      { client_txn_id: "txn-all" },
      { client_txn_id: "txn-selected", pack_keys: ["loaded", "outside"] },
      { client_txn_id: "txn-selected", pack_keys: ["loaded", "outside"] },
    ]);
    expect(() =>
      captureReferencePackAttempt(
        { kind: "refresh_selected", packKeys: [] },
        "bad",
      ),
    ).toThrow();
  });

  it("preserves upload bytes filename media hint and staged-only metadata across lost-response replay", async () => {
    const bytes = new Uint8Array([0, 1, 255, 128, 10, 13]);
    const file = new File([bytes], "exact.zip", { type: "application/zip" });
    const attempt = captureReferencePackAttempt(
      { kind: "import", filename: file.name },
      "upload-transaction",
      file,
    );
    fetchMock.mockRejectedValueOnce(new TypeError("lost response"));
    expect((await submitReferencePackAttempt(attempt, signal())).kind).toBe(
      "uncertain",
    );
    fetchMock.mockResolvedValueOnce(envelope(referencePackJobFixture(), 202));
    expect(
      (await submitReferencePackAttempt(attempt, signal(), true)).kind,
    ).toBe("accepted");
    for (const [, init] of fetchMock.mock.calls) {
      const body = init?.body as FormData;
      const uploaded = body.get("file") as File;
      const metadata = body.get("metadata") as Blob;
      expect(uploaded.name).toBe("exact.zip");
      expect(uploaded.type).toBe("application/zip");
      expect(new Uint8Array(await readBlob(uploaded))).toEqual(bytes);
      expect(
        JSON.parse(new TextDecoder().decode(await readBlob(metadata))),
      ).toEqual({
        client_txn_id: "upload-transaction",
        activation_policy: "staged_only",
      });
    }
  });

  it("distinguishes definitive rejection conflict malformed success and transport uncertainty", async () => {
    const attempt = captureReferencePackAttempt(
      { kind: "refresh_all" },
      "txn-outcomes",
    );
    // Each conflict code is sent by an operation that declares it for 409.
    const activation = captureReferencePackAttempt(
      {
        kind: "activate",
        target: { pack_key: "type_registry.host", pack_version: "1" },
      },
      "txn-activation-outcomes",
    );
    const upload = captureReferencePackAttempt(
      { kind: "import", filename: "pack.zip" },
      "txn-upload-outcomes",
      new File(["pack"], "pack.zip", { type: "application/zip" }),
    );
    for (const [submitted, code, kind] of [
      [attempt, "reference_pack_operation_rejected", "operation_rejected"],
      [activation, "reference_pack_activation_rejected", "activation_rejected"],
      [upload, "reference_pack_verification_failed", "verification_failed"],
      [upload, "client_txn_conflict", "transaction_conflict"],
    ] as const) {
      fetchMock.mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            error: {
              code,
              status: 409,
              message: "unsafe wire message",
              request_id: "request-test",
              retryable: false,
              details: { reason_code: "unsafe reason" },
            },
          }),
          { status: 409, headers: { "content-type": "application/json" } },
        ),
      );
      expect(
        await submitReferencePackAttempt(submitted, signal()),
      ).toMatchObject({
        kind: "rejected",
        problem: { kind, status: 409, code },
      });
    }
    // No operation declares reference_pack_state_conflict: it is a client
    // contract failure and never selects the state-conflict outcome.
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          error: {
            code: "reference_pack_state_conflict",
            status: 409,
            message: "unsafe wire message",
            request_id: "request-test",
            retryable: false,
            details: {},
          },
        }),
        { status: 409, headers: { "content-type": "application/json" } },
      ),
    );
    const undeclared = await submitReferencePackAttempt(attempt, signal());
    expect(undeclared).toMatchObject({
      problem: { code: "invalid_public_contract_response" },
    });
    if ("problem" in undeclared)
      expect(undeclared.problem.kind).not.toBe("state_conflict");
    for (const status of [400, 401, 403, 409, 500, 202]) {
      fetchMock.mockResolvedValueOnce(
        new Response("{malformed", {
          status,
          headers: { "content-type": "application/json" },
        }),
      );
      // A malformed body is the client contract failure whatever its raw
      // status, handled exactly like an invalid success response.
      const outcome = await submitReferencePackAttempt(attempt, signal());
      expect(outcome).toMatchObject({
        kind: "uncertain",
        problem: {
          kind: "contract",
          status,
          code: "invalid_public_contract_response",
        },
      });
    }
    fetchMock.mockRejectedValueOnce(new TypeError("network"));
    expect(await submitReferencePackAttempt(attempt, signal())).toMatchObject({
      kind: "uncertain",
      problem: { kind: "transport", status: 0 },
    });
    fetchMock.mockResolvedValueOnce(
      envelope({ job_id: referencePackTestJobId, status: "queued" }, 202),
    );
    expect(await submitReferencePackAttempt(attempt, signal())).toMatchObject({
      kind: "uncertain",
      problem: { kind: "contract", status: 202 },
    });
  });

  it("validates complete job envelopes while accepting unknown additive reference kinds", async () => {
    for (const status of [
      "queued",
      "running",
      "cancel_requested",
      "succeeded",
      "failed",
      "canceled",
    ] as const) {
      expect(validReferencePackJob(referencePackJobFixture(status))).toBe(true);
    }
    const extra = referencePackJobFixture("succeeded", {
      result_summary: {
        code: "reference_packs_refreshed",
        message: "complete",
        resource_refs: [{ kind: "future_reference_kind", id: "opaque" }],
      },
    });
    fetchMock.mockResolvedValueOnce(envelope(extra, 200));
    expect(
      (await loadReferencePackJob(referencePackTestJobId, signal())).kind,
    ).toBe("read");
    expect(
      validReferencePackJob(
        referencePackJobFixture("running", {
          progress: { completed: 3, total: null },
        }),
      ),
    ).toBe(true);
    expect(
      validReferencePackJob(
        referencePackJobFixture("running", {
          progress: { completed: 3, total: 2 },
        }),
      ),
    ).toBe(false);
    expect(
      validReferencePackJob(
        referencePackJobFixture("succeeded", { retained_until: null }),
      ),
    ).toBe(false);
    expect(
      validReferencePackJob(
        referencePackJobFixture("succeeded", {
          result_summary: {
            code: "reference_pack_imported",
            message: "done",
            resource_refs: [
              {
                kind: "reference_pack_version",
                id: "arbitrary",
                route: "/api/v1/reference-packs/a/b",
              },
            ],
          },
        }),
      ),
    ).toBe(false);
    fetchMock.mockResolvedValueOnce(
      envelope(
        { ...extra, job_id: "33333333-3333-4333-8333-333333333333" },
        200,
      ),
    );
    expect(
      (await loadReferencePackJob(referencePackTestJobId, signal())).kind,
    ).toBe("failed");
  });
});

function readBlob(blob: Blob): Promise<ArrayBuffer> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result as ArrayBuffer);
    reader.onerror = () => reject(reader.error);
    reader.readAsArrayBuffer(blob);
  });
}
