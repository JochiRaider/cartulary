import { afterEach, describe, expect, it, vi } from "vitest";
import {
  importProfileId,
  importRouteFamily,
  networkFlowActivityProfileId,
  networkFlowRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import { captureWorkbookUpload } from "../imports/importRequests";
import { ImportClient } from "../services/importClient";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { jsonResponse } from "../testing/fetchMockTestSupport";
import {
  importTestEnvelope,
  importTestJob,
  importTestScope,
} from "../testing/workbookImportTestSupport";

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});
describe("Network Flow import dispatch admission", () => {
  it("rechecks the analytical claim after shared queue admission before transmitting bytes", async () => {
    const availability = readyExtensionAvailability(importTestScope.incidentId);
    const fetch = vi.fn(async () =>
      jsonResponse(importTestEnvelope(importTestJob()), 202),
    );
    vi.stubGlobal("fetch", fetch);
    let release!: () => void;
    const queue = availability.runProfileRequest(
      importProfileId,
      importRouteFamily,
      () =>
        new Promise<void>((resolve) => {
          release = resolve;
        }),
    );
    await Promise.resolve();
    const client = new ImportClient({
      incidentId: importTestScope.incidentId,
      availability,
      canDispatch: () =>
        availability.isRouteAvailable(
          networkFlowActivityProfileId,
          networkFlowRouteFamily,
        ),
    });
    const pending = client.send(
      captureWorkbookUpload(
        importTestScope,
        new File(["protected source"], "flows.csv"),
      ),
      new AbortController().signal,
    );
    availability.setDiscovery([
      {
        profile_id: importProfileId,
        claimed: true,
        contract_major: 1,
        route_families: [importRouteFamily],
        workspace_keys: [],
        capabilities: [],
      },
    ]);
    release();
    await queue.catch(() => {});
    await pending;
    expect(fetch).not.toHaveBeenCalled();
  });
});
