import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  referencePackFixture,
  referencePackJobFixture,
  referencePackTestJobId,
} from "../testing/referencePackTestSupport";
import {
  captureReferencePackAttempt,
  loadReferencePackJob,
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
    for (const [code, kind] of [
      ["reference_pack_state_conflict", "state_conflict"],
      ["reference_pack_activation_rejected", "activation_rejected"],
      ["reference_pack_verification_failed", "verification_failed"],
      ["client_txn_conflict", "transaction_conflict"],
    ]) {
      fetchMock.mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            error: {
              code,
              message: "unsafe wire message",
              details: { reason_code: "unsafe reason" },
            },
          }),
          { status: 409, headers: { "content-type": "application/json" } },
        ),
      );
      expect(await submitReferencePackAttempt(attempt, signal())).toMatchObject(
        { kind: "rejected", problem: { kind, status: 409, code } },
      );
    }
    for (const status of [400, 401, 403, 409, 500, 202]) {
      fetchMock.mockResolvedValueOnce(
        new Response("{malformed", {
          status,
          headers: { "content-type": "application/json" },
        }),
      );
      const outcome = await submitReferencePackAttempt(attempt, signal());
      expect(outcome.kind).toBe(
        [401, 403].includes(status)
          ? "access_failed"
          : status < 500 && status >= 400
            ? "rejected"
            : "uncertain",
      );
      if ("problem" in outcome) expect(outcome.problem.status).toBe(status);
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
