import { afterEach, describe, expect, it, vi } from "vitest";
import {
  membershipAuthority as authority,
  membershipEnvelope as envelope,
  membershipJSON as json,
  membershipFixture as member,
} from "../../testing/incidentMembershipManagementTestSupport";
import {
  listIncidentMembershipPage,
  mutateIncidentMembership,
} from "./incidentMembershipManagementClient";

const signal = () => new AbortController().signal;
afterEach(() => vi.unstubAllGlobals());
describe("Incident membership management typed transport", () => {
  it("uses only limit and opaque cursor and retains the server ordering", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>(async () =>
      json(envelope([member()], "next")),
    );
    vi.stubGlobal("fetch", fetch);
    const result = await listIncidentMembershipPage({
      authority,
      cursor: "opaque%token",
      signal: signal(),
    });
    expect(result.ok).toBe(true);
    const url = new URL(String(fetch.mock.calls[0]?.[0]), "http://localhost");
    expect(Object.fromEntries(url.searchParams)).toEqual({
      limit: "100",
      cursor_token: "opaque%token",
    });
  });
  it("rejects wrong incident duplicate unordered and malformed paged resources", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => json(envelope([member()], "repeat"))),
    );
    expect(
      await listIncidentMembershipPage({
        authority,
        cursor: "repeat",
        signal: signal(),
      }),
    ).toMatchObject({
      ok: false,
      problem: { code: "invalid_pagination_request" },
    });
    const cases = [
      envelope([
        member({ incident_id: "00000000-0000-4000-8000-000000001002" }),
      ]),
      envelope([member(), member()]),
      envelope([member(), member({ user_id: authority.actorId })]),
      {
        data: { memberships: [member()] },
        meta: { request_id: "missing-paging" },
      },
      {
        ...envelope(),
        meta: {
          request_id: "limit",
          paging: { limit: 500, has_more: false, next_cursor: null },
        },
      },
    ];
    for (const payload of cases) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () => json(payload)),
      );
      expect(
        await listIncidentMembershipPage({
          authority,
          cursor: null,
          signal: signal(),
        }),
      ).toMatchObject({ ok: false, status: 502 });
    }
  });
  it("preserves 201 and 200 create outcomes and sends one email selector", async () => {
    for (const status of [200, 201]) {
      const fetch = vi.fn<typeof globalThis.fetch>(async () =>
        json({ data: member(), meta: { request_id: "create" } }, status),
      );
      vi.stubGlobal("fetch", fetch);
      const payload = {
        email: "analyst@example.test",
        role: "viewer" as const,
        client_txn_id: "exact-create",
      };
      expect(
        await mutateIncidentMembership({
          authority,
          input: { kind: "add", payload },
          signal: signal(),
        }),
      ).toMatchObject({ ok: true, status });
      expect(JSON.parse(String(fetch.mock.calls[0]?.[1]?.body))).toEqual(
        payload,
      );
    }
  });
  it("uses captured PATCH and DELETE bodies and accepts a bodyless 204", async () => {
    const fetch = vi
      .fn<typeof globalThis.fetch>()
      .mockResolvedValueOnce(
        json({ data: member(), meta: { request_id: "patch" } }),
      )
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    vi.stubGlobal("fetch", fetch);
    await mutateIncidentMembership({
      authority,
      input: {
        kind: "role",
        userId: member().user_id,
        payload: { base_membership_version: 1, role: "viewer" },
      },
      signal: signal(),
    });
    expect(
      await mutateIncidentMembership({
        authority,
        input: {
          kind: "remove",
          userId: member().user_id,
          payload: { base_membership_version: 1 },
        },
        signal: signal(),
      }),
    ).toEqual({ ok: true, status: 204, resource: null });
    expect(JSON.parse(String(fetch.mock.calls[0]?.[1]?.body))).toEqual({
      base_membership_version: 1,
      role: "viewer",
    });
    expect(JSON.parse(String(fetch.mock.calls[1]?.[1]?.body))).toEqual({
      base_membership_version: 1,
    });
  });
  it("does not accept mismatched resource identities roles or undeclared success statuses", async () => {
    for (const response of [
      json({
        data: member({ user_id: authority.actorId }),
        meta: { request_id: "wrong-id" },
      }),
      json({
        data: member({ role: "admin" }),
        meta: { request_id: "wrong-role" },
      }),
      json({ data: member(), meta: { request_id: "wrong-status" } }, 201),
    ]) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () => response),
      );
      expect(
        await mutateIncidentMembership({
          authority,
          input: {
            kind: "role",
            userId: member().user_id,
            payload: { base_membership_version: 1, role: "viewer" },
          },
          signal: signal(),
        }),
      ).toMatchObject({ ok: false, status: 502 });
    }
  });
  it("retains safe error identities without raw server message or arbitrary details", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        json({ error: { code: "private-secret", message: "hidden" } }, 409),
      ),
    );
    expect(
      await mutateIncidentMembership({
        authority,
        input: {
          kind: "remove",
          userId: member().user_id,
          payload: { base_membership_version: 1 },
        },
        signal: signal(),
      }),
    ).toEqual({
      ok: false,
      status: 502,
      problem: { code: "invalid_public_contract_response" },
    });

    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        json(
          {
            error: {
              code: "membership_exists_use_patch",
              message: "SQL secret",
              details: { field: "role", secret: "hidden" },
            },
          },
          409,
        ),
      ),
    );
    expect(
      await mutateIncidentMembership({
        authority,
        input: {
          kind: "add",
          payload: {
            email: "analyst@example.test",
            role: "viewer",
            client_txn_id: "request",
          },
        },
        signal: signal(),
      }),
    ).toEqual({
      ok: false,
      status: 409,
      problem: { code: "membership_exists_use_patch", field: "role" },
    });
  });
});
