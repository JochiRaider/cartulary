import { gridShellTestId } from "@cartulary/test-utils";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { successEnvelope } from "./timelineWorkbookTestSupport";
import { WorkbookShell } from "./WorkbookShell";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const hostsViewSchemaId = "cartulary.view.hosts.v1";
const identitiesViewSchemaId = "cartulary.view.identities.v1";
const evidenceViewSchemaId = "cartulary.view.evidence.v1";
const indicatorsViewSchemaId = "cartulary.view.indicators.v1";

describe("WorkbookShell surface selection", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [{ incident_id: "incident-1", role: "admin" }],
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1")) {
        return successEnvelope({
          incident_id: "incident-1",
          incident_key: "IR-1",
          title: "Incident 1",
          description: null,
          severity: null,
          tlp: null,
          current_phase: null,
          primary_external_case_ref: null,
          incident_version: 1,
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1/memberships")) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "incident-1",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-preferences/")) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      for (const viewSchemaId of [
        timelineViewSchemaId,
        hostsViewSchemaId,
        identitiesViewSchemaId,
        evidenceViewSchemaId,
        indicatorsViewSchemaId,
      ]) {
        if (url.includes(`/views/${viewSchemaId}/query`)) {
          return successEnvelope({
            incident_id: "incident-1",
            view_schema_id: viewSchemaId,
            rows: [],
          });
        }
      }
      return successEnvelope({});
    });
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal(
      "WebSocket",
      class {
        onmessage: ((event: MessageEvent) => void) | null = null;

        close() {}
      } as unknown as typeof WebSocket,
    );
  });

  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("selects required built-in and system view surfaces by view_schema_id", async () => {
    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(await screen.findByTestId("surface-tab-evidence"));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${evidenceViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(
      screen.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeTruthy();

    fireEvent.change(screen.getByTestId("system-view-selector"), {
      target: { value: indicatorsViewSchemaId },
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${indicatorsViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(window.location.search).toContain(
      `view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`,
    );
  });
});
