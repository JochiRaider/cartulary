import { requireViewContract } from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { publicWorkbookSchema } from "../../testing/workbookSchemaTestSupport";
import { ordinaryCreateContributions } from "../features/ordinary/ordinaryCreateContributions";
import { createOrdinaryCreateTransport } from "./createOrdinaryCreateTransport";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
const view = "cartulary.view.evidence.v1";
function fixture() {
  const transport = createOrdinaryCreateTransport(undefined);
  const target = requireViewContract(view);
  const attempt = transport.capture({
    authority: {
      incidentId: "10000000-0000-4000-8000-000000000001",
      actorId: "99999999-9999-4999-8999-999999999999",
      role: "admin",
      closed: false,
      sessionIdentity: "same-account",
    },
    target,
    draft: {
      id: 1,
      revision: 2,
      values: { "evidence.title": "exact" },
      references: {},
    },
    request: {
      client_txn_id: "00000000-0000-4000-8000-000000000001",
      "evidence.title": "exact",
    },
    clientTxnId: "00000000-0000-4000-8000-000000000001",
  });
  const receipt = {
    data: {
      view_schema_id: view,
      change_set_id: "30000000-0000-4000-8000-000000000001",
      row: fullWorkbookViewRow(
        target,
        "20000000-0000-4000-8000-000000000001",
        1,
        { "evidence.title": "exact" },
      ),
    },
    meta: { request_id: "request-create" },
  };
  return { transport, attempt, receipt };
}
describe("ordinary create transport", () => {
  it("verifies all fourteen discovered contracts and rejects changed fields, inputs and minima", async () => {
    const transport = createOrdinaryCreateTransport(undefined);
    for (const view of ordinaryCreateContributions.flatMap(
      (item) => item.views,
    )) {
      const target = requireViewContract(view);
      const schema = publicWorkbookSchema(target);
      for (const [key, value] of Object.entries({
        valid: schema,
        unavailable: { ...schema, create_capable: false },
        fields: { ...schema, fields: [] },
        inputs: { ...schema, create_inputs: [{ input_key: "fabricated" }] },
        minima: {
          ...schema,
          inline_create: {
            ...schema.inline_create,
            minimum_create_field_sets: [["unavailable.required"]],
          },
        },
      })) {
        vi.stubGlobal(
          "fetch",
          vi.fn(
            async () =>
              new Response(
                JSON.stringify({
                  data: value,
                  meta: { request_id: "discovery" },
                }),
                {
                  status: 200,
                  headers: { "content-type": "application/json" },
                },
              ),
          ),
        );
        const verified = transport.verify(target, new AbortController().signal);
        if (key === "valid")
          await expect(verified, view).resolves.toBeUndefined();
        else await expect(verified, `${view} ${key}`).rejects.toThrow();
      }
    }
  });
  it("preserves complete receipts and HTTP outcomes with unchanged captured replay bytes", async () => {
    const { transport, attempt, receipt } = fixture();
    const cookie = vi
      .spyOn(document, "cookie", "get")
      .mockReturnValue("cartulary_csrf=first");
    const fetch = vi.fn<typeof globalThis.fetch>(
      async () =>
        new Response(JSON.stringify(receipt), {
          status: 200,
          headers: {
            "content-type": "application/json",
            "X-Request-ID": "request-create",
          },
        }),
    );
    vi.stubGlobal("fetch", fetch);
    const result = await transport.send(attempt, new AbortController().signal);
    expect(result).toEqual({ kind: "accepted", status: 200, receipt });
    cookie.mockReturnValue("cartulary_csrf=recovered");
    await transport.send(attempt, new AbortController().signal);
    expect(fetch.mock.calls).toHaveLength(2);
    expect(
      fetch.mock.calls.map(([url, init]) => [
        url,
        init?.body,
        init?.credentials,
      ]),
    ).toEqual([
      [attempt.path, attempt.body, "include"],
      [attempt.path, attempt.body, "include"],
    ]);
    expect(
      fetch.mock.calls.map(([, init]) =>
        new Headers(init?.headers).get("X-CSRF-Token"),
      ),
    ).toEqual(["first", "recovered"]);
    expect(attempt.body).toBe(JSON.stringify(attempt.request));
    expect(Object.isFrozen(attempt)).toBe(true);
  });
  it("classifies transport failures, server failures and malformed success as uncertain", async () => {
    const { transport, attempt, receipt } = fixture();
    for (const response of [
      () => Promise.reject(new TypeError("lost response")),
      async () =>
        new Response(
          JSON.stringify({
            error: { code: "internal", request_id: "request-create" },
          }),
          { status: 500, headers: { "content-type": "application/json" } },
        ),
      async () =>
        new Response(
          JSON.stringify({ ...receipt, data: { ...receipt.data, row: {} } }),
          { status: 201, headers: { "content-type": "application/json" } },
        ),
      async () =>
        new Response(JSON.stringify({ ...receipt, meta: {} }), {
          status: 201,
          headers: { "content-type": "application/json" },
        }),
      async () =>
        new Response(JSON.stringify(receipt), {
          status: 201,
          headers: {
            "content-type": "application/json",
            "X-Request-ID": "wrong-request",
          },
        }),
      async () =>
        new Response(JSON.stringify(receipt), {
          status: 202,
          headers: { "content-type": "application/json" },
        }),
    ]) {
      vi.stubGlobal("fetch", vi.fn(response));
      await expect(
        transport.send(attempt, new AbortController().signal),
      ).resolves.toEqual({ kind: "uncertain" });
    }
  });
});
