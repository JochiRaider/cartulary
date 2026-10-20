import { afterEach, describe, expect, it, vi } from "vitest";
import {
  metadataEnvelope as envelope,
  metadataIncident as incident,
  metadataJSON as json,
  metadataActorId,
  metadataIncidentId,
} from "../../testing/incidentMetadataTestSupport";
import type { LifecycleAttempt } from "../incidentLifecycleModel";
import {
  mutateIncidentLifecycle,
  readLifecycleIncident,
} from "./incidentLifecycleClient";

const attempt: LifecycleAttempt = {
  id: 1,
  authority: {
    incidentId: metadataIncidentId,
    actorId: metadataActorId,
    lifetime: "session",
    role: "admin",
    apiBase: "/gateway",
  },
  action: "close",
  resource: incident(),
  revision: 1,
  payload: {
    base_incident_version: 1,
    client_txn_id: "exact-key",
    reason: " e\u0301\r\nreason  ",
  },
};
const closed = () =>
  incident({
    status: "closed",
    closed_at: "2026-08-02T00:00:00Z",
    incident_version: 2,
  });
const signal = () => new AbortController().signal;
afterEach(() => vi.unstubAllGlobals());
describe("Incident lifecycle transport", () => {
  it("uses existing generated operations and sends only exact captured members", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>(async () =>
      json(envelope(closed())),
    );
    vi.stubGlobal("fetch", fetch);
    await readLifecycleIncident({
      authority: attempt.authority,
      signal: signal(),
    });
    await mutateIncidentLifecycle({
      attempt: {
        ...attempt,
        payload: {
          ...attempt.payload,
          ignored: "omit",
        } as typeof attempt.payload,
      },
      signal: signal(),
    });
    expect(fetch.mock.calls[0]?.[0]).toBe(
      `/gateway/api/v1/incidents/${metadataIncidentId}`,
    );
    expect(fetch.mock.calls[1]?.[0]).toBe(
      `/gateway/api/v1/incidents/${metadataIncidentId}/close`,
    );
    expect(JSON.parse(String(fetch.mock.calls[1]?.[1]?.body))).toEqual(
      attempt.payload,
    );
    expect(fetch.mock.calls[1]?.[1]?.method).toBe("POST");
    expect(fetch.mock.calls[1]?.[1]?.signal).toBeInstanceOf(AbortSignal);
    fetch.mockResolvedValue(json(envelope(incident({ incident_version: 3 }))));
    expect(
      await mutateIncidentLifecycle({
        attempt: {
          ...attempt,
          action: "reopen",
          payload: { ...attempt.payload, base_incident_version: 2 },
        },
        signal: signal(),
      }),
    ).toMatchObject({ ok: true });
    expect(fetch.mock.calls[2]?.[0]).toBe(
      `/gateway/api/v1/incidents/${metadataIncidentId}/reopen`,
    );
  });
  it("rejects identity version transition and closed-at contradictions", async () => {
    for (const data of [
      {},
      { ...closed(), incident_id: metadataActorId },
      { ...closed(), incident_version: 1 },
      { ...closed(), incident_version: 3 },
      { ...closed(), closed_at: null },
      { ...closed(), status: "archived" },
      incident({ incident_version: 2 }),
      { ...closed(), closed_at: "not a date" },
    ]) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () => json({ data, meta: { request_id: "test" } })),
      );
      expect(
        await mutateIncidentLifecycle({ attempt, signal: signal() }),
      ).toMatchObject({
        ok: false,
        status: 502,
        problem: { code: "invalid_public_contract_response" },
      });
    }
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        json(envelope(incident({ closed_at: "2026-08-02T00:00:00Z" }))),
      ),
    );
    expect(
      await readLifecycleIncident({
        authority: attempt.authority,
        signal: signal(),
      }),
    ).toMatchObject({ ok: false });
  });
  it("preserves typed error distinctions without retaining private diagnostics", async () => {
    for (const code of [
      "invalid_incident_lifecycle_request",
      "incident_version_conflict",
      "illegal_transition",
      "client_txn_conflict",
      "authorization_denied",
      "session_required",
    ] as const) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () =>
          json(
            {
              error: {
                code,
                message: "private",
                details: {
                  field: "reason",
                  reason_code: "reason_too_long",
                  secret: "private",
                },
              },
            },
            409,
          ),
        ),
      );
      expect(
        await mutateIncidentLifecycle({ attempt, signal: signal() }),
      ).toEqual({
        ok: false,
        status: 409,
        problem: { code, field: "reason", reason: "reason_too_long" },
      });
    }
  });
  it("accepts exact historical receipts and sends normalized-valid long raw reasons unchanged", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>(async () =>
      json(envelope(closed())),
    );
    vi.stubGlobal("fetch", fetch);
    const payload = { ...attempt.payload, reason: "e\u0301".repeat(4096) };
    expect(
      await mutateIncidentLifecycle({
        attempt: { ...attempt, payload },
        signal: signal(),
      }),
    ).toMatchObject({ ok: true, resource: closed() });
    expect(JSON.parse(String(fetch.mock.calls[0]?.[1]?.body)).reason).toBe(
      payload.reason,
    );
  });
});
