import {
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackFileInputTestId,
  referencePackJobStatusTestId,
  referencePackRefreshAllButtonTestId,
  referencePackReloadButtonTestId,
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

import { jsonResponse } from "../testing/fetchMockTestSupport";
import type { SessionData } from "./phase1Client";
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

  it("Phase 11 U-11-REFERENCE-PACK-04 hides deployment Reference Pack controls from non-admin sessions", () => {
    render(<ReferencePackAdminPanel session={session(false)} />);
    expect(
      screen.getByTestId(referencePackAdminPanelTestId()).textContent,
    ).toContain("Deployment admin access is required");
    expect(screen.queryByTestId(referencePackFileInputTestId())).toBeNull();
  });

  it("Phase 11 U-11-REFERENCE-PACK-06 shows job progress and cancel controls for deployment-admin Reference Pack work", async () => {
    fetchMock.mockImplementation((input, init) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/reference-packs?limit=100" && method === "GET") {
        return Promise.resolve(
          jsonResponse({
            data: {
              pack_versions: [
                {
                  pack_key: "type_registry.host",
                  pack_version: "1",
                  pack_kind: "type_registry",
                  pack_version_state: "verified_available",
                  active: true,
                  verification_result: "passed",
                },
                {
                  pack_key: "type_registry.identity",
                  pack_version: "1",
                  pack_kind: "type_registry",
                  pack_version_state: "draft",
                  active: false,
                  verification_result: "pending",
                },
              ],
            },
          }),
        );
      }
      if (url === "/api/v1/reference-packs/refresh" && method === "POST") {
        return Promise.resolve(
          jsonResponse({
            data: jobResource("job-1", "queued", true, 0, 1),
          }),
        );
      }
      if (url === "/api/v1/jobs/job-1" && method === "GET") {
        return Promise.resolve(
          jsonResponse({
            data: jobResource("job-1", "running", true, 0, 1),
          }),
        );
      }
      if (url === "/api/v1/jobs/job-1/cancel" && method === "POST") {
        return Promise.resolve(
          jsonResponse({
            data: jobResource("job-1", "cancel_requested", false, 0, 1),
          }),
        );
      }
      throw new Error(`unexpected fetch ${method} ${url}`);
    });

    render(<ReferencePackAdminPanel session={session(true)} />);
    const panel = screen.getByTestId(referencePackAdminPanelTestId());
    expect(panel.style.background).toBe("var(--ct-colors-surface-2)");
    expect(panel.style.border).toBe("var(--ct-border-hairline)");
    expect(panel.style.boxSizing).toBe("border-box");
    const fileInput = screen.getByTestId(referencePackFileInputTestId());
    expect(fileInput.style.boxSizing).toBe("border-box");
    expect(fileInput.style.maxWidth).toBe("100%");
    expect(fileInput.style.minWidth).toBe("0");

    fireEvent.click(screen.getByTestId(referencePackReloadButtonTestId()));
    const packRow = await screen.findByTestId(
      referencePackRowTestId("type_registry.host", "1"),
    );
    expect(packRow.textContent).toContain("type_registry.host@1");
    expect(
      screen.getByTestId(referencePackRowTestId("type_registry.identity", "1")),
    ).toBeTruthy();

    fireEvent.change(screen.getByLabelText("Filter reference packs"), {
      target: { value: "identity" },
    });
    expect(
      screen.queryByTestId(referencePackRowTestId("type_registry.host", "1")),
    ).toBeNull();
    expect(
      screen.getByTestId(referencePackRowTestId("type_registry.identity", "1")),
    ).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Filter reference packs"), {
      target: { value: "" },
    });

    fireEvent.click(screen.getByTestId(referencePackRefreshAllButtonTestId()));
    await waitFor(() => {
      expect(
        screen.getByTestId(referencePackJobStatusTestId()).textContent,
      ).toContain("running · 0/1");
    });
    expect(screen.getByTestId(referencePackCancelButtonTestId())).toBeTruthy();

    fireEvent.click(screen.getByTestId(referencePackCancelButtonTestId()));
    await waitFor(() => {
      expect(
        screen.getByTestId(referencePackJobStatusTestId()).textContent,
      ).toContain("Cancel requested");
    });
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        "/api/v1/jobs/job-1/cancel",
        expect.objectContaining({ method: "POST" }),
      );
    });
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
    submitted_by_user_id: "user-1",
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
