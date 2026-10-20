import { afterEach, describe, expect, it, vi } from "vitest";
import {
  metadataActorId,
  metadataEnvelope,
  metadataIncident,
  metadataIncidentId,
  metadataJSON,
} from "../../testing/incidentMetadataTestSupport";
import {
  patchIncidentMetadata,
  readIncidentMetadata,
} from "./incidentMetadataClient";

const authority = {
  incidentId: metadataIncidentId,
  actorId: metadataActorId,
  lifetime: "test",
  role: "reviewer" as const,
};
const signal = () => new AbortController().signal;
afterEach(() => vi.unstubAllGlobals());
describe("Incident metadata transport", () => {
  it("uses generated paths and preserves sparse null and exact Unicode input", async () => {
    const fetch = vi.fn<typeof globalThis.fetch>(async () =>
      metadataJSON(metadataEnvelope()),
    );
    vi.stubGlobal("fetch", fetch);
    await readIncidentMetadata({ authority, signal: signal() });
    const request = {
      base_incident_version: 1,
      description: " e\u0301\r\n\tline ",
      tlp: null,
      ignored: "never send",
    };
    await patchIncidentMetadata({
      authority,
      payload: request,
      signal: signal(),
    });
    expect(fetch.mock.calls[0]?.[0]).toBe(
      `/api/v1/incidents/${metadataIncidentId}`,
    );
    const init = fetch.mock.calls[1]?.[1];
    expect(JSON.parse(String(init?.body))).toEqual({
      base_incident_version: 1,
      description: " e\u0301\r\n\tline ",
      tlp: null,
    });
    expect(init?.method).toBe("PATCH");
    expect(init?.signal).toBeInstanceOf(AbortSignal);
  });
  it("accepts authoritative no-op responses without demanding a version increase", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => metadataJSON(metadataEnvelope())),
    );
    expect(
      await patchIncidentMetadata({
        authority,
        payload: { base_incident_version: 1 },
        signal: signal(),
      }),
    ).toEqual({ ok: true, resource: metadataIncident() });
  });
  it("rejects malformed wrong-incident regressing and unexpected-status success responses", async () => {
    for (const [data, status] of [
      [{}, 200],
      [metadataIncident({ incident_id: metadataActorId }), 200],
      [metadataIncident({ incident_version: 0 }), 200],
      [metadataIncident(), 201],
      [metadataIncident(), 200],
    ] as const) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () =>
          metadataJSON({ data, meta: { request_id: "test" } }, status),
        ),
      );
      expect(
        await patchIncidentMetadata({
          authority,
          payload: { base_incident_version: 2, severity: "critical" },
          signal: signal(),
        }),
      ).toMatchObject({
        ok: false,
        problem: { code: "invalid_public_contract_response" },
      });
    }
  });
  it("retains only safe known error identities and bounded field reasons", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        metadataJSON(
          {
            error: {
              code: "invalid_incident_patch",
              message: "private SQL",
              details: {
                field: "severity",
                reason_code: "forbidden_field",
                secret: "private",
              },
            },
          },
          400,
        ),
      ),
    );
    expect(
      await patchIncidentMetadata({
        authority,
        payload: { base_incident_version: 1 },
        signal: signal(),
      }),
    ).toEqual({
      ok: false,
      status: 400,
      problem: {
        code: "invalid_incident_patch",
        field: "severity",
        reason: undefined,
      },
    });
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        metadataJSON(
          { error: { code: "private future error", message: "private" } },
          500,
        ),
      ),
    );
    expect(
      await readIncidentMetadata({ authority, signal: signal() }),
    ).toMatchObject({ problem: { code: "unknown_public_error" } });
  });
});
