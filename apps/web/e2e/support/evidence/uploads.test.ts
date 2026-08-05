// @vitest-environment node

import { Buffer } from "node:buffer";
import type { Page } from "@playwright/test";
import { describe, expect, it, vi } from "vitest";

import { csrfCookieName } from "../auth/storageState";
import { apiBase } from "../runtime/configuration";
import type { JsonRequestContextLike } from "../transport/publicJsonClient";
import { createAndUploadObjectBlob } from "./uploads";

describe("evidence upload adapter", () => {
  it("sends validated same-origin PUT targets with exactly the returned headers", async () => {
    const uploadTarget = validUploadTarget({
      headers: {
        "Content-Type": "application/cartulary-evidence",
        "X-Cartulary-Upload": "opaque-header",
      },
    });
    const { fetch, page, put } = evidenceUploadPage(
      objectBlobEnvelope(uploadTarget),
    );
    const body = Buffer.from("invalid-target-body", "utf8");

    await expect(
      createAndUploadObjectBlob(page, {
        body,
        clientTxnId: "evidence-upload-valid",
        contentType: "text/plain",
        filename: "evidence.txt",
        incidentId: "11111111-1111-4111-8111-111111111111",
      }),
    ).resolves.toEqual(
      expect.objectContaining({
        object_blob_id: "22222222-2222-4222-8222-222222222222",
      }),
    );
    expect(fetch).toHaveBeenCalledWith(`${apiBase}/api/v1/object-blobs`, {
      data: {
        byte_size: body.byteLength,
        client_txn_id: "evidence-upload-valid",
        content_type_hint: "text/plain",
        filename_hint: "evidence.txt",
        incident_id: "11111111-1111-4111-8111-111111111111",
      },
      headers: { "X-CSRF-Token": "csrf-token" },
      method: "POST",
    });
    expect(put).toHaveBeenCalledWith(
      `${apiBase}/api/v1/object-uploads/upl_valid-token_1`,
      { data: body, headers: uploadTarget.headers },
    );
  });

  it("does not synthesize headers when the returned header map is empty", async () => {
    const { page, put } = evidenceUploadPage(
      objectBlobEnvelope(
        validUploadTarget({
          headers: {},
          href: "/api/v1/object-uploads/upl_content-type-fallback",
        }),
      ),
    );

    await createAndUploadObjectBlob(page, evidenceUploadOptions());

    expect(put).toHaveBeenCalledWith(
      `${apiBase}/api/v1/object-uploads/upl_content-type-fallback`,
      {
        data: expect.any(Buffer),
        headers: {},
      },
    );
  });

  it("rejects malformed targets and envelopes before upload dispatch", async () => {
    const invalidTargets: unknown[] = [
      validUploadTarget({
        href: "https://attacker.test/api/v1/object-uploads/token",
      }),
      validUploadTarget({
        href: "//attacker.test/api/v1/object-uploads/token",
      }),
      validUploadTarget({
        href: "https://user:password@attacker.test/api/v1/object-uploads/token",
      }),
      validUploadTarget({ href: "not-an-upload-path" }),
      validUploadTarget({
        href: "/api/v1/object-uploads/token?credential=secret",
      }),
      validUploadTarget({ href: "/api/v1/object-uploads/token#fragment" }),
      validUploadTarget({ href: "/api/v1/object-uploads/token/extra" }),
      validUploadTarget({ href: "/api/v1/object-uploads/.." }),
      validUploadTarget({ method: "POST" }),
      {
        expires_at: "2026-08-05T02:30:00Z",
        headers: {},
        href: "/api/v1/object-uploads/token",
      },
    ];

    for (const uploadTarget of invalidTargets) {
      const { page, put } = evidenceUploadPage(
        objectBlobEnvelope(uploadTarget),
      );
      await expect(
        createAndUploadObjectBlob(page, evidenceUploadOptions()),
      ).rejects.toThrow();
      expect(put).not.toHaveBeenCalled();
    }

    const malformed = evidenceUploadPage({ data: {}, meta: {} });
    await expect(
      createAndUploadObjectBlob(malformed.page, evidenceUploadOptions()),
    ).rejects.toThrow();
    expect(malformed.put).not.toHaveBeenCalled();
  });
});

function evidenceUploadOptions() {
  return {
    body: Buffer.from("invalid-target-body", "utf8"),
    clientTxnId: "evidence-upload-invalid",
    contentType: "text/plain",
    filename: "evidence.txt",
    incidentId: "11111111-1111-4111-8111-111111111111",
  };
}

function validUploadTarget(
  overrides: Record<string, unknown> = {},
): Record<string, unknown> {
  return {
    expires_at: "2026-08-05T02:30:00Z",
    headers: {},
    href: "/api/v1/object-uploads/upl_valid-token_1",
    method: "PUT",
    ...overrides,
  };
}

function objectBlobEnvelope(uploadTarget: unknown) {
  return {
    data: {
      accepted_contract: {
        byte_size: Buffer.byteLength("invalid-target-body"),
        content_type_hint: "text/plain",
        filename_hint: "evidence.txt",
        incident_id: "11111111-1111-4111-8111-111111111111",
        sha256_hex: null,
      },
      incident_id: "11111111-1111-4111-8111-111111111111",
      object_blob_id: "22222222-2222-4222-8222-222222222222",
      pending_expires_at: "2026-08-06T01:30:00Z",
      target_expires_at: "2026-08-05T02:30:00Z",
      upload_state: "pending",
      upload_target: uploadTarget,
    },
    meta: { request_id: "request-evidence-upload" },
  };
}

function evidenceUploadPage(payload: unknown) {
  const fetch = vi.fn<JsonRequestContextLike["fetch"]>(async () => ({
    headers: () => ({}),
    json: async () => payload,
    ok: () => true,
    status: () => 201,
  }));
  const put = vi.fn(async () => ({
    ok: () => true,
    status: () => 204,
  }));
  const page = {
    context: () => ({
      cookies: async () => [{ name: csrfCookieName, value: "csrf-token" }],
    }),
    request: { fetch, put },
  } as unknown as Page;
  return { fetch, page, put };
}
