import { Buffer } from "node:buffer";
import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import { expect } from "./fixtures";
import {
  type NetworkFlowFixtureFields,
  networkFlowFixtureSetup,
  openClaimedNetworkAnalysis,
} from "./support/extensions/network_flow_activity/workspace";
import {
  assertViewportVisualRegression,
  assertVisualRegression,
} from "./support/visual/capture";
import { test as base } from "./support/visual/fixtures";

const test = base.extend<NetworkFlowFixtureFields>(networkFlowFixtureSetup);

test("Capture claimed Network Analysis accepted inspector, rejected diagnostics, and graph contributor drawer at the deterministic desktop viewport.", async ({
  page,
  networkFlowMinimalCSV,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await openClaimedNetworkAnalysis(page, "NETWORKFLOWVISUAL");
  const fixture = networkFlowMinimalCSV.buffer.toString("utf8");
  const lines = fixture.trimEnd().split("\n");
  const invalidRow = lines.at(-1)?.replace("192.0.2.10", "not-an-ip") ?? "";
  await page.getByTestId(networkAnalysisTestId("import-input")).setInputFiles({
    name: "visual-flow-with-a-deliberately-long-source-name.csv",
    mimeType: "text/csv",
    buffer: Buffer.from(`${fixture.trimEnd()}\n${invalidRow}\n`),
  });
  const mappingDialog = page.getByTestId(
    networkAnalysisTestId("mapping-dialog"),
  );
  await expect(mappingDialog).toBeVisible();
  await page
    .getByTestId(networkAnalysisTestId("mapping-display-name"))
    .fill("visual-flow");
  await assertViewportVisualRegression(
    page,
    "network-flow-analysis-mapping-dialog",
  );
  await page.getByTestId(networkAnalysisTestId("mapping-preview")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("mapping-preview-summary")),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("mapping-apply")).click();
  await expect(mappingDialog).toHaveCount(0);
  await page.getByLabel("Endpoint IP value").fill("203.0.113.255");
  await page.getByTestId(networkAnalysisTestId("accepted-query-apply")).click();
  const networkFilteredEmpty = page
    .getByTestId(networkAnalysisTestId("accepted-grid"))
    .locator('[data-grid-data-state="filtered_empty"]');
  await expect(networkFilteredEmpty).toBeVisible();
  await expect(networkFilteredEmpty).toContainText(
    "No rows match the current filters.",
  );
  await assertVisualRegression(
    page,
    "network-flow-analysis-filtered-empty-grid",
    page.getByTestId(networkAnalysisTestId("accepted-grid")),
  );
  await networkFilteredEmpty
    .getByRole("button", { name: "Clear filters" })
    .click();
  await page
    .getByRole("gridcell", { name: /Source IP:/u })
    .first()
    .click();
  await expect(
    page.getByTestId(networkAnalysisTestId("inspector")),
  ).toBeVisible();

  await assertViewportVisualRegression(
    page,
    "network-flow-analysis-accepted-inspector",
  );
  await page.getByTestId(networkAnalysisTestId("mode-rejected")).click();
  await assertViewportVisualRegression(
    page,
    "network-flow-analysis-rejected-diagnostics",
  );
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  const edge = page.getByTestId(/^network-flow-edge-/).first();
  await expect(edge).toBeVisible();
  await edge.getByRole("button", { name: /^Select edge/u }).click();
  const contributorDrawer = page.getByTestId(
    networkAnalysisTestId("contributor-drawer"),
  );
  await expect(contributorDrawer).toBeVisible();
  await expect(contributorDrawer).toContainText("visual-flow");
  await assertViewportVisualRegression(
    page,
    "network-flow-analysis-graph-contributors",
  );
  await page.getByTestId(networkAnalysisTestId("contributor-close")).click();
  await page.getByRole("button", { name: "Saved graphs" }).click();
  await page.getByTestId(networkAnalysisTestId("saved-graph-create")).click();
  await page
    .getByTestId(networkAnalysisTestId("saved-graph-name"))
    .fill("Visual saved graph");
  await page.getByRole("button", { name: "Save graph" }).click();
  const savedGraphs = page.getByTestId(networkAnalysisTestId("saved-graphs"));
  await expect(
    savedGraphs.getByText("Materialization succeeded.", { exact: true }),
  ).toBeVisible({ timeout: 15_000 });
  await expect(
    page.getByTestId(/^network-flow-saved-graph-edge-/u).first(),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "network-flow-analysis-saved-graph-result",
  );
  await page.setViewportSize({ width: 1024, height: 720 });
  await page
    .getByTestId(networkAnalysisTestId("graph-surface-explore"))
    .click();
  await page.getByTestId(networkAnalysisTestId("mode-rows")).click();
  await page
    .getByLabel("Endpoint IP value")
    .fill(`2001:db8:${"longsegment".repeat(12)}`);
  await page
    .getByTestId(networkAnalysisTestId("advanced-filters"))
    .locator("summary")
    .click();
  await assertViewportVisualRegression(
    page,
    "network-flow-analysis-narrow-query-controls",
  );
  await page
    .getByTestId(networkAnalysisTestId("advanced-filters"))
    .locator("summary")
    .click();
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  // Returning from Rows starts a fresh exploration read. Let it finish before
  // detaching exploration so Save current graph has a deterministic source.
  await expect(page.getByTestId(/^network-flow-edge-/).first()).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("graph-surface-saved")).click();
  await page.setViewportSize({ width: 768, height: 640 });
  await assertViewportVisualRegression(
    page,
    "network-flow-analysis-compact-saved-graphs",
  );
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.getByTestId(networkAnalysisTestId("delete-trigger")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("delete-dialog")),
  ).toBeVisible();
  await assertViewportVisualRegression(
    page,
    "network-flow-analysis-delete-dialog",
  );
});
