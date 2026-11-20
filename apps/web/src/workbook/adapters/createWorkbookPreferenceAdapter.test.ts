import { afterEach, describe, expect, it, vi } from "vitest";
import { metadataJSON } from "../../testing/incidentMetadataTestSupport";
import {
  defaultPreference,
  homePreference,
  preferenceActorId,
  preferenceHosts,
  preferenceIncidentId,
} from "../../testing/workbookPreferenceTestSupport";
import { createWorkbookPreferenceAdapter } from "./createWorkbookPreferenceAdapter";

const envelope = (data: unknown) => ({
  data,
  meta: { request_id: "preferences" },
});
const create = () =>
  createWorkbookPreferenceAdapter({
    incidentId: preferenceIncidentId,
    actorId: preferenceActorId,
    apiBase: "/base",
  });
const signal = () => new AbortController().signal;
describe("Workbook preference transport", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });
  it("uses all four generated operations and explicit nullable PUT bodies with CSRF", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=csrf-token",
    );
    const fetch = vi.fn(async (url: string, init?: RequestInit) => {
      const home = url.endsWith("/me");
      const body = init?.body ? JSON.parse(String(init.body)) : {};
      return metadataJSON(
        envelope(
          home
            ? homePreference(body.home_sheet_ref ?? null)
            : defaultPreference(body.default_sheet_ref ?? null),
        ),
      );
    });
    vi.stubGlobal("fetch", fetch);
    const port = create();
    expect(await port.readHome({ signal: signal() })).toEqual({
      kind: "accepted",
      value: homePreference(),
    });
    expect(await port.readDefault({ signal: signal() })).toEqual({
      kind: "accepted",
      value: defaultPreference(),
    });
    for (const sheetRef of [
      preferenceHosts,
      { kind: "saved_view", id: "saved-1" },
      {
        kind: "extension_workspace",
        extension_profile_id: "network_flow_activity",
        workspace_key: "network_analysis",
      },
      null,
    ] as const) {
      expect(await port.setHomeSheet({ signal: signal(), sheetRef })).toEqual({
        kind: "accepted",
        value: homePreference(sheetRef),
      });
      expect(
        await port.setDefaultSheet({ signal: signal(), sheetRef }),
      ).toEqual({ kind: "accepted", value: defaultPreference(sheetRef) });
    }
    expect(fetch).toHaveBeenCalledTimes(10);
    for (const [url, init] of fetch.mock.calls.slice(2)) {
      expect(init?.method).toBe("PUT");
      expect(new Headers(init?.headers).get("X-CSRF-Token")).toBe("csrf-token");
      const body = JSON.parse(String(init?.body));
      expect(Object.keys(body)).toEqual([
        url.endsWith("/me") ? "home_sheet_ref" : "default_sheet_ref",
      ]);
      expect(url).toMatch(
        new RegExp(
          `/base/api/v1/incidents/${preferenceIncidentId}/workbook-preferences/(me|default)$`,
        ),
      );
    }
  });
  it("rejects wrong incident wrong user missing and malformed GET resources", async () => {
    const invalid = [
      {
        ...homePreference(),
        incident_id: "00000000-0000-4000-8000-000000009999",
      },
      { ...homePreference(), user_id: "00000000-0000-4000-8000-000000000099" },
      { ...homePreference(), home_sheet_ref: undefined },
      { ...homePreference(), home_sheet_ref: { kind: "view_schema", id: "" } },
      { ...homePreference(), created_at: undefined },
      {
        ...homePreference(),
        home_sheet_ref: {
          kind: "saved_view",
          id: "saved-1",
          view_schema_id: "wrong",
        },
      },
    ];
    const port = create();
    for (const resource of invalid) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () => metadataJSON(envelope(resource))),
      );
      expect(await port.readHome({ signal: signal() })).toMatchObject({
        kind: "rejected",
        problem: { code: "invalid_public_contract_response" },
      });
    }
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        metadataJSON(
          envelope({
            ...defaultPreference(),
            incident_id: "00000000-0000-4000-8000-000000009999",
          }),
        ),
      ),
    );
    expect(await port.readDefault({ signal: signal() })).toMatchObject({
      kind: "rejected",
    });
  });
  it("requires an exact nullable pointer and current actor in PUT acknowledgement", async () => {
    const port = create();
    for (const resource of [
      homePreference(),
      {
        ...homePreference(preferenceHosts),
        user_id: "00000000-0000-4000-8000-000000000099",
      },
      {
        ...homePreference(preferenceHosts),
        incident_id: "00000000-0000-4000-8000-000000009999",
      },
    ]) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () => metadataJSON(envelope(resource))),
      );
      expect(
        await port.setHomeSheet({
          signal: signal(),
          sheetRef: preferenceHosts,
        }),
      ).toMatchObject({
        kind: "uncertain",
        problem: { code: "invalid_public_contract_response" },
      });
    }
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        metadataJSON(envelope(homePreference(preferenceHosts))),
      ),
    );
    expect(
      await port.setHomeSheet({ signal: signal(), sheetRef: null }),
    ).toMatchObject({ kind: "uncertain" });
  });
  it("returns no-op attribution unchanged and detaches immutable accepted resources", async () => {
    const resource = {
      ...defaultPreference(preferenceHosts),
      updated_by_user_id: "00000000-0000-4000-8000-000000000099",
    };
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => metadataJSON(envelope(resource))),
    );
    const result = await create().setDefaultSheet({
      signal: signal(),
      sheetRef: preferenceHosts,
    });
    expect(result).toEqual({ kind: "accepted", value: resource });
    if (result.kind !== "accepted") throw new Error("Missing acknowledgement");
    expect(Object.isFrozen(result.value)).toBe(true);
    expect(Object.isFrozen(result.value.default_sheet_ref)).toBe(true);
  });
  it("separates known rejection from uncertain server transport and invalid success outcomes", async () => {
    const port = create();
    // An undeclared code is a client contract failure (HTTP 502 locally).
    for (const [write, status, code, kind, resultStatus, problemCode] of [
      [
        "home",
        400,
        "invalid_mutation_payload",
        "rejected",
        400,
        "invalid_mutation_payload",
      ],
      ["home", 401, "session_required", "rejected", 401, "session_required"],
      [
        "home",
        403,
        "csrf_verification_failed",
        "rejected",
        403,
        "csrf_verification_failed",
      ],
      [
        "default",
        403,
        "authorization_denied",
        "rejected",
        403,
        "authorization_denied",
      ],
      [
        "home",
        403,
        "authorization_denied",
        "uncertain",
        502,
        "invalid_public_contract_response",
      ],
      [
        "home",
        404,
        "incident_not_found",
        "rejected",
        404,
        "incident_not_found",
      ],
      ["home", 500, "internal_error", "uncertain", 500, "internal_error"],
      [
        "home",
        503,
        "internal_error",
        "uncertain",
        502,
        "invalid_public_contract_response",
      ],
      [
        "home",
        400,
        "secret_error_value",
        "uncertain",
        502,
        "invalid_public_contract_response",
      ],
    ] as const) {
      vi.stubGlobal(
        "fetch",
        vi.fn(async () =>
          metadataJSON(
            {
              error: {
                code,
                status,
                message: "sensitive server detail",
                request_id: "request-test",
                retryable: false,
                details: { raw: "sensitive" },
              },
            },
            status,
          ),
        ),
      );
      const result = await (write === "home"
        ? port.setHomeSheet({ signal: signal(), sheetRef: null })
        : port.setDefaultSheet({ signal: signal(), sheetRef: null }));
      expect(result).toEqual({
        kind,
        status: resultStatus,
        problem: { code: problemCode },
      });
      expect(JSON.stringify(result)).not.toContain("sensitive");
      expect(JSON.stringify(result)).not.toContain("secret_error_value");
    }
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new TypeError("private transport detail");
      }),
    );
    expect(
      await port.setHomeSheet({ signal: signal(), sheetRef: null }),
    ).toMatchObject({ kind: "uncertain", problem: { code: "transport" } });
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => metadataJSON({ data: {} })),
    );
    expect(
      await port.setHomeSheet({ signal: signal(), sheetRef: null }),
    ).toMatchObject({ kind: "uncertain" });
  });
});
