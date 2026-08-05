import {
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackErrorTestId,
  referencePackFileInputTestId,
  referencePackJobStatusTestId,
  referencePackListStatusTestId,
  referencePackRefreshAllButtonTestId,
  referencePackRowTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  deferred,
  errorResponse,
  jsonResponse,
} from "../testing/fetchMockTestSupport";
import type { SessionData } from "./api/appShellClient";
import { ReferencePackAdminPanel } from "./ReferencePackAdminPanel";

describe("ReferencePackAdminPanel", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("hides deployment Reference Pack controls from non-admin sessions", () => {
    render(<ReferencePackAdminPanel session={session(false)} />);
    expect(
      screen.getByTestId(referencePackAdminPanelTestId()).textContent,
    ).toContain("Deployment admin access is required");
    expect(screen.queryByTestId(referencePackFileInputTestId())).toBeNull();
  });

  it("shows job progress and cancel controls for deployment-admin Reference Pack work", async () => {
    const jobID = "11111111-1111-4111-8111-111111111111";
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url.startsWith("/api/v1/reference-packs?") && method === "GET") {
        const parsed = new URL(url, "http://cartulary.test");
        const search = parsed.searchParams.get("search");
        const filtered =
          search === "identity"
            ? [
                packResource({
                  pack_key: "type_registry.identity",
                  pack_version_state: "staged",
                  active: false,
                  verification_result: "pending",
                }),
              ]
            : [
                packResource({
                  pack_key: "type_registry.host",
                  pack_version_state: "verified_available",
                  active: true,
                  verification_result: "passed",
                }),
                packResource({
                  pack_key: "type_registry.identity",
                  pack_version_state: "staged",
                  active: false,
                  verification_result: "pending",
                }),
              ];
        return Promise.resolve(jsonResponse(packListEnvelope(filtered)));
      }
      if (url === "/api/v1/reference-packs/refresh" && method === "POST") {
        return Promise.resolve(
          jsonResponse({
            data: jobResource(jobID, "queued", true, 0, 1),
            meta: { request_id: "request-1" },
          }),
        );
      }
      if (url === `/api/v1/jobs/${jobID}` && method === "GET") {
        return Promise.resolve(
          jsonResponse({
            data: jobResource(jobID, "running", true, 0, 1),
            meta: { request_id: "request-2" },
          }),
        );
      }
      if (url === `/api/v1/jobs/${jobID}/cancel` && method === "POST") {
        return Promise.resolve(
          jsonResponse({
            data: jobResource(jobID, "cancel_requested", false, 0, 1),
            meta: { request_id: "request-3" },
          }),
        );
      }
      throw new Error(`unexpected fetch ${method} ${url}`);
    });

    render(<ReferencePackAdminPanel session={session(true)} />);
    expect(screen.getByTestId(referencePackAdminPanelTestId())).toBeTruthy();
    expect(screen.queryByTestId(referencePackFileInputTestId())).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Import pack" }));
    expect(screen.getByTestId(referencePackFileInputTestId())).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByText("Filter loaded packs")).toBeNull();

    const packRow = await screen.findByTestId(
      referencePackRowTestId("type_registry.host", "1"),
    );
    expect(packRow.textContent).toContain("type_registry.host@1");
    expect(
      screen.getByTestId(referencePackRowTestId("type_registry.identity", "1")),
    ).toBeTruthy();

    fireEvent.change(screen.getByLabelText("Search reference packs"), {
      target: { value: "identity" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Search" }));
    await waitFor(() => {
      expect(
        screen.queryByTestId(referencePackRowTestId("type_registry.host", "1")),
      ).toBeNull();
    });
    expect(
      screen.getByTestId(referencePackRowTestId("type_registry.identity", "1")),
    ).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Search reference packs"), {
      target: { value: "" },
    });
    fireEvent.change(screen.getByLabelText("Reference pack state"), {
      target: { value: "verified_available" },
    });
    fireEvent.change(
      screen.getByLabelText("Reference pack verification result"),
      {
        target: { value: "passed" },
      },
    );
    fireEvent.change(screen.getByLabelText("Reference pack active state"), {
      target: { value: "true" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Search" }));
    await waitFor(() => {
      expect(
        fetchMock.mock.calls.some(
          ([input]) =>
            String(input) ===
            "/api/v1/reference-packs?active=true&limit=100&pack_version_state=verified_available&verification_result=passed",
        ),
      ).toBe(true);
    });

    fireEvent.click(screen.getByTestId(referencePackRefreshAllButtonTestId()));
    await waitFor(() => {
      expect(
        screen.getByTestId(referencePackJobStatusTestId()).textContent,
      ).toContain("running");
    });
    expect(screen.getByTestId(referencePackCancelButtonTestId())).toBeTruthy();

    fireEvent.click(screen.getByTestId(referencePackCancelButtonTestId()));
    await waitFor(() => {
      expect(
        screen.getByTestId(referencePackJobStatusTestId()).textContent,
      ).toMatch(/cancel/i);
    });
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        `/api/v1/jobs/${jobID}/cancel`,
        expect.objectContaining({ method: "POST" }),
      );
    });
  });

  it("discards stale reference pack search responses", async () => {
    const initialResponse = deferred<Response>();
    const searchResponse = deferred<Response>();
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/reference-packs?limit=100" && method === "GET") {
        return initialResponse.promise;
      }
      if (
        url === "/api/v1/reference-packs?limit=100&search=identity" &&
        method === "GET"
      ) {
        return searchResponse.promise;
      }
      throw new Error(`unexpected fetch ${method} ${url}`);
    });

    render(<ReferencePackAdminPanel session={session(true)} />);
    fireEvent.change(screen.getByLabelText("Search reference packs"), {
      target: { value: "identity" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Search" }));

    searchResponse.resolve(
      jsonResponse(
        packListEnvelope([
          packResource({
            pack_key: "type_registry.identity",
            pack_version_state: "staged",
            active: false,
            verification_result: "pending",
          }),
        ]),
      ),
    );
    expect(
      await screen.findByTestId(
        referencePackRowTestId("type_registry.identity", "1"),
      ),
    ).toBeTruthy();

    initialResponse.resolve(
      jsonResponse(
        packListEnvelope([
          packResource({
            pack_key: "type_registry.host",
            pack_version_state: "verified_available",
            active: true,
            verification_result: "passed",
          }),
        ]),
      ),
    );

    await waitFor(() => {
      expect(
        screen.queryByTestId(referencePackRowTestId("type_registry.host", "1")),
      ).toBe(null);
    });
  });

  it("retains accepted rows while searching, resets paging on Enter, and ignores stale terminal results", async () => {
    const appendResponse = deferred<Response>();
    const searchResponse = deferred<Response>();
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/reference-packs?limit=100" && method === "GET") {
        return Promise.resolve(
          jsonResponse(
            packListEnvelope(
              [packResource({ pack_key: "type_registry.host" })],
              {
                limit: 100,
                has_more: true,
                next_cursor: "cursor-a",
              },
            ),
          ),
        );
      }
      if (
        url === "/api/v1/reference-packs?cursor_token=cursor-a&limit=100" &&
        method === "GET"
      ) {
        return appendResponse.promise;
      }
      if (
        url === "/api/v1/reference-packs?limit=100&search=identity" &&
        method === "GET"
      ) {
        return searchResponse.promise;
      }
      throw new Error(`unexpected fetch ${method} ${url}`);
    });

    render(<ReferencePackAdminPanel session={session(true)} />);
    expect(
      await screen.findByTestId(
        referencePackRowTestId("type_registry.host", "1"),
      ),
    ).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "Load more" }));
    fireEvent.change(screen.getByLabelText("Search reference packs"), {
      target: { value: "identity" },
    });
    fireEvent.keyDown(screen.getByLabelText("Search reference packs"), {
      key: "Enter",
    });

    expect(
      screen.getByTestId(referencePackListStatusTestId()).textContent,
    ).toBe("Searching reference packs");
    expect(
      screen.getByTestId(referencePackRowTestId("type_registry.host", "1")),
    ).toBeTruthy();
    expect(
      fetchMock.mock.calls.some(
        ([input]) =>
          String(input) === "/api/v1/reference-packs?limit=100&search=identity",
      ),
    ).toBe(true);

    searchResponse.resolve(errorResponse("reference_pack_search_failed", 500));
    await waitFor(() => {
      expect(screen.getByTestId(referencePackErrorTestId()).textContent).toBe(
        "reference_pack_search_failed",
      );
    });
    appendResponse.resolve(
      jsonResponse(
        packListEnvelope([packResource({ pack_key: "type_registry.stale" })]),
      ),
    );
    await waitFor(() => {
      expect(
        screen.queryByTestId(
          referencePackRowTestId("type_registry.stale", "1"),
        ),
      ).toBeNull();
    });
    expect(screen.getByTestId(referencePackErrorTestId()).textContent).toBe(
      "reference_pack_search_failed",
    );
  });

  it("clears protected list state and invalidates pending generations on authorization loss", async () => {
    const pendingResponse = deferred<Response>();
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/reference-packs?limit=100" && method === "GET") {
        return Promise.resolve(
          jsonResponse(
            packListEnvelope([
              packResource({ pack_key: "type_registry.host" }),
            ]),
          ),
        );
      }
      if (
        url === "/api/v1/reference-packs?limit=100&search=identity" &&
        method === "GET"
      ) {
        return pendingResponse.promise;
      }
      throw new Error(`unexpected fetch ${method} ${url}`);
    });

    const view = render(<ReferencePackAdminPanel session={session(true)} />);
    expect(
      await screen.findByTestId(
        referencePackRowTestId("type_registry.host", "1"),
      ),
    ).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Search reference packs"), {
      target: { value: "identity" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Search" }));

    view.rerender(<ReferencePackAdminPanel session={session(false)} />);
    expect(
      screen.getByTestId(referencePackAdminPanelTestId()).textContent,
    ).toContain("Deployment admin access is required");
    expect(
      screen.queryByTestId(referencePackRowTestId("type_registry.host", "1")),
    ).toBeNull();

    pendingResponse.resolve(
      jsonResponse(
        packListEnvelope([
          packResource({ pack_key: "type_registry.identity" }),
        ]),
      ),
    );
    await waitFor(() => {
      expect(
        screen.queryByTestId(
          referencePackRowTestId("type_registry.identity", "1"),
        ),
      ).toBeNull();
    });
  });

  it("rejects malformed successful Reference Pack responses", async () => {
    fetchMock.mockResolvedValue(
      jsonResponse({
        data: { pack_versions: [{ pack_key: "incomplete" }] },
        meta: { request_id: "request-invalid" },
      }),
    );

    render(<ReferencePackAdminPanel session={session(true)} />);
    await waitFor(() => {
      expect(screen.getByTestId(referencePackErrorTestId()).textContent).toBe(
        "invalid_public_contract_response",
      );
    });
    expect(
      screen.getByTestId(referencePackListStatusTestId()).textContent,
    ).toBe("Reference packs unavailable");
  });
});

function session(isDeploymentAdmin: boolean): SessionData {
  return {
    user_id: "user-1",
    display_name: "Operator",
    provider_type: "local",
    mfa_state: "satisfied",
    is_deployment_admin: isDeploymentAdmin,
    authenticated_at: "2026-05-24T00:00:00Z",
    idle_expires_at: "2026-05-24T01:00:00Z",
    absolute_expires_at: "2026-05-24T02:00:00Z",
    session_expires_at: "2026-05-24T01:00:00Z",
    memberships: [],
  };
}

function jobResource(
  jobID: string,
  status: string,
  cancelable: boolean,
  completed: number,
  total: number,
) {
  return {
    job_id: jobID,
    status,
    cancelable,
    scope: { kind: "deployment" },
    status_route: `/api/v1/jobs/${jobID}`,
    submitted_by_user_id: "22222222-2222-4222-8222-222222222222",
    submitted_at: "2026-05-24T00:00:00Z",
    updated_at: "2026-05-24T00:00:00Z",
    progress: { completed, total },
    started_at: status === "queued" ? null : "2026-05-24T00:00:01Z",
    finished_at: null,
    retained_until: null,
    result_summary: null,
    error_summary: null,
  };
}

function packResource(
  overrides: Partial<{
    active: boolean;
    pack_key: string;
    pack_version_state:
      | "staged"
      | "verified_available"
      | "disabled"
      | "failed"
      | "missing";
    verification_result: "pending" | "passed" | "failed";
  }> = {},
) {
  return {
    activated_at: null,
    activated_by_user_id: null,
    active: false,
    imported_at: "2026-05-24T00:00:00Z",
    imported_by_user_id: null,
    manifest_sha256: "a".repeat(64),
    pack_contract_version: "1",
    pack_key: "type_registry.host",
    pack_kind: "type_registry",
    pack_version: "1",
    pack_version_state: "verified_available" as const,
    payload_sha256: "b".repeat(64),
    previous_active_version: null,
    signer_key_id: null,
    source_identifier: null,
    verification_method: "sha256",
    verification_result: "passed" as const,
    ...overrides,
  };
}

function packListEnvelope(
  packVersions: ReturnType<typeof packResource>[],
  paging = { limit: 100, has_more: false, next_cursor: null as string | null },
) {
  return {
    data: { pack_versions: packVersions },
    meta: {
      request_id: "request-list",
      paging,
    },
  };
}
