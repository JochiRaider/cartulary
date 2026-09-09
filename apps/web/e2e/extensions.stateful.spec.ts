import {
  incidentAdministrationTestId,
  networkAnalysisTestId,
  saveStateTestId,
  surfaceTabTestId,
  workbookPreferenceTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";

import { expect, test } from "./fixtures";
import { openIncidentControls } from "./pages/deploymentAdministration";
import {
  createTimelineRow,
  editTimelineSummary,
  installPatchController,
} from "./support/collaboration/replay";
import {
  expectNetworkFlowRuntimeProfile,
  openClaimedNetworkAnalysis,
  openNetworkFlowIncident,
} from "./support/extensions/network_flow_activity/workspace";
import { apiBase } from "./support/runtime/configuration";

test("Verify extension availability bootstrap, no-store startup, lazy Network Analysis loading, and Base client identity continuity.", async ({
  page,
}) => {
  expectNetworkFlowRuntimeProfile("network_flow_claimed");
  const networkFlowChunks: string[] = [];
  page.on("request", (request) => {
    if (request.url().includes("/assets/NetworkFlowFeature-")) {
      networkFlowChunks.push(request.url());
    }
  });
  const startupResponsePromise = page.waitForResponse((response) =>
    response.url().includes("/workbook-startup"),
  );
  const incidentId = await openNetworkFlowIncident(page, "EXTAVAILABILITY");
  const startupResponse = await startupResponsePromise;

  expect(startupResponse.headers()["cache-control"]).toBe("no-store");
  const startup = (await startupResponse.json()) as {
    data: {
      extension_workspace_availability: {
        incident_id: string;
        schema_id: string;
        workspaces: Array<{
          extension_profile_id: string;
          workspace_key: string;
        }>;
      };
    };
  };
  expect(startup.data.extension_workspace_availability).toEqual({
    schema_id: "cartulary.extension_workspace_availability.v1",
    incident_id: incidentId,
    workspaces: [
      {
        extension_profile_id: "network_flow_activity",
        workspace_key: "network_analysis",
      },
    ],
  });
  const packagedRootResponse = await page.request.get(`${apiBase}/`);
  expect(packagedRootResponse.ok()).toBe(true);
  const packagedRoot = await packagedRootResponse.text();
  const supportMatch = packagedRoot.match(
    /<script id="cartulary-client-extension-support-registry" type="application\/json">([^<]+)<\/script>/u,
  );
  expect(supportMatch).not.toBeNull();
  expect(JSON.parse(supportMatch?.[1] ?? "null")).toMatchObject({
    asset_set_sha256: expect.stringMatching(/^[0-9a-f]{64}$/u),
    client_build_class: "standard",
    profiles: [
      {
        profile_id: "import",
        supported_contract_majors: [1],
        workspace_keys: [],
        capability_ids: [],
      },
      {
        profile_id: "network_flow_activity",
        supported_contract_majors: [6],
        workspace_keys: ["network_analysis"],
        capability_ids: [],
      },
    ],
  });
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  expect(networkFlowChunks).toEqual([]);

  const clientInstanceID = await page.evaluate(() =>
    window.sessionStorage.getItem("cartulary.client_instance_id"),
  );
  expect(clientInstanceID).toMatch(/^[0-9a-f-]{36}$/u);
  const row = await createTimelineRow(
    page,
    incidentId,
    "Pending Base work across extension navigation",
  );
  await page.reload();
  const patchController = await installPatchController(page);
  const held = patchController.holdNextPatch({ recordId: row.record_id });
  try {
    await editTimelineSummary(page, row.record_id, "Retained Base work", {
      expectValueAfterCommit: false,
    });
    await held.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    await page.getByTestId(networkAnalysisTestId("tab")).click();
    await expect(
      page.getByTestId(networkAnalysisTestId("workspace")),
    ).toBeVisible();
    expect(networkFlowChunks).toHaveLength(1);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

    await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
    await expect(
      page.getByTestId(networkAnalysisTestId("workspace")),
    ).toHaveCount(0);
    expect(
      await page.evaluate(() =>
        window.sessionStorage.getItem("cartulary.client_instance_id"),
      ),
    ).toBe(clientInstanceID);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    held.release();
    await held.waitForCompletion;
    try {
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    } catch (error) {
      await test.info().attach("save-settlement-timing.json", {
        body: JSON.stringify(
          await page.evaluate(() =>
            performance
              .getEntriesByType("mark")
              .filter((entry) =>
                entry.name.startsWith("cartulary.workbook.pending_"),
              )
              .map((entry) => ({
                ...entry.toJSON(),
                detail: (entry as PerformanceMark).detail,
              })),
          ),
        ),
        contentType: "application/json",
      });
      throw error;
    }
  } finally {
    held.release();
    await held.waitForCompletion;
    await patchController.dispose();
  }
});

test("preferences store and clear the claimed authorized extension workspace without changing active identity", async ({
  page,
}) => {
  const incidentId = await openClaimedNetworkAnalysis(page, "WP-EXTENSION");
  const original = page.url();
  const requests: unknown[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PUT" &&
      request.url().includes("workbook-preferences")
    )
      requests.push(request.postDataJSON());
  });
  await openIncidentControls(page);
  for (const kind of ["home", "default"] as const) {
    await expect(
      page.getByTestId(workbookPreferenceTestId(kind, "set")),
    ).toHaveAttribute("aria-disabled", "false");
    await page.getByTestId(workbookPreferenceTestId(kind, "set")).click();
    await expect(
      page.getByTestId(workbookPreferenceTestId(kind, "outcome")),
    ).toHaveText(
      `${kind === "home" ? "Home" : "Incident default"} update confirmed.`,
    );
    await expect(
      page.getByTestId(
        incidentAdministrationTestId(
          kind === "home" ? "pref-home-sheet-ref" : "pref-default-sheet-ref",
        ),
      ),
    ).toContainText("network_flow_activity/network_analysis");
    await expect(
      page.getByTestId(workbookPreferenceTestId(kind, "clear")),
    ).toHaveAttribute("aria-disabled", "false");
    await page.getByTestId(workbookPreferenceTestId(kind, "clear")).click();
    await expect(
      page.getByTestId(workbookPreferenceTestId(kind, "outcome")),
    ).toHaveText(
      `${kind === "home" ? "Home" : "Incident default"} clear confirmed.`,
    );
  }
  expect(requests).toEqual([
    {
      home_sheet_ref: {
        kind: "extension_workspace",
        extension_profile_id: "network_flow_activity",
        workspace_key: "network_analysis",
      },
    },
    { home_sheet_ref: null },
    {
      default_sheet_ref: {
        kind: "extension_workspace",
        extension_profile_id: "network_flow_activity",
        workspace_key: "network_analysis",
      },
    },
    { default_sheet_ref: null },
  ]);
  expect(page.url()).toBe(original);
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toBeVisible();
  expect(incidentId).not.toBe("");
});
