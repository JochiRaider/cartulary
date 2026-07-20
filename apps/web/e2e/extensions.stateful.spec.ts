import {
  networkAnalysisTestId,
  surfaceTabTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import {
  expectNetworkFlowRuntimeProfile,
  openNetworkFlowIncident,
} from "./support/extensions/network_flow_activity/workspace";
import { apiBase } from "./support/runtime/configuration";

const timelineViewSchemaId = "cartulary.view.timeline.v2";

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
        profile_id: "network_flow_activity",
        supported_contract_majors: [2],
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
  await page.getByTestId(networkAnalysisTestId("tab")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toBeVisible();
  expect(networkFlowChunks).toHaveLength(1);

  await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toHaveCount(0);
  expect(
    await page.evaluate(() =>
      window.sessionStorage.getItem("cartulary.client_instance_id"),
    ),
  ).toBe(clientInstanceID);
});
