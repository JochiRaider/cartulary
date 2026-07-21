import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import type { Request } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  expectNetworkFlowRuntimeProfile,
  importNetworkFlowCSV,
  networkFlowMinimalCSV,
  openClaimedNetworkAnalysis,
  openNetworkFlowIncident,
} from "./support/extensions/network_flow_activity/workspace";
import { apiBase } from "./support/runtime/configuration";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";

test("Network Flow unclaimed workspace remains unavailable", async ({
  page,
}) => {
  expectNetworkFlowRuntimeProfile("default");
  const incidentId = await openNetworkFlowIncident(page, "NFAC001");

  await expect(page.getByTestId(networkAnalysisTestId("tab"))).toHaveCount(0);
  const response = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/network-flow/tables`,
  );
  expect(response.status()).toBe(404);
});

test("Network Analysis claimed empty state exposes import entry", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFAC002");

  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toHaveAttribute("data-extension-profile-id", "network_flow_activity");
  await expect(
    page.getByLabel("Empty Network Analysis workspace"),
  ).toBeVisible();
  await expect(
    page.getByTestId(networkAnalysisTestId("import-trigger")),
  ).toBeVisible();
});

test("Network Analysis import creates one inner table tab", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFAC006");
  await importNetworkFlowCSV(page, {
    displayName: "cisco-sna-minimal",
    file: networkFlowMinimalCSV,
  });

  await expect(
    page.getByRole("tab", { name: /cisco-sna-minimal/ }),
  ).toHaveCount(1);
  await expect(
    page.getByTestId(networkAnalysisTestId("status-strip")),
  ).toContainText("1 active table");
  await expect(
    page.getByTestId(networkAnalysisTestId("accepted-grid")),
  ).toBeVisible();
});

test("Network Analysis soft delete removes active table tab", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFAC023");
  await importNetworkFlowCSV(page, {
    displayName: "delete-me",
    file: networkFlowMinimalCSV,
  });

  await page.getByTestId(networkAnalysisTestId("delete-trigger")).click();
  await page
    .getByTestId(networkAnalysisTestId("delete-confirmation"))
    .fill("delete-me");
  await page.getByTestId(networkAnalysisTestId("delete-confirm")).click();

  await expect(page.getByRole("tab", { name: /delete-me/ })).toHaveCount(0);
  await expect(
    page.getByTestId(networkAnalysisTestId("accepted-grid")),
  ).toHaveCount(0);
  await expect(
    page.getByLabel("Empty Network Analysis workspace"),
  ).toBeVisible();
  await expect(
    page.getByTestId(networkAnalysisTestId("status-strip")),
  ).toContainText("delete-me was deleted");
});

test("Network Analysis graph mode exposes table selection controls", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFAC029");
  await importNetworkFlowCSV(page, {
    displayName: "scope-a",
    file: networkFlowMinimalCSV,
  });
  await importNetworkFlowCSV(page, {
    displayName: "scope-b",
    file: networkFlowMinimalCSV,
  });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();

  const scope = page.getByTestId(networkAnalysisTestId("graph-scope"));
  await expect(scope).toBeVisible();
  await scope.getByLabel("Selected tables").check();
  await expect(scope.getByRole("checkbox", { name: "scope-a" })).toBeChecked();
  await scope.getByRole("checkbox", { name: "scope-b" }).check();
  await expect(scope.getByRole("checkbox", { name: "scope-b" })).toBeChecked();
});

test("Network Analysis table graph defaults to active-table scope", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFAC030");
  await importNetworkFlowCSV(page, {
    displayName: "active-scope",
    file: networkFlowMinimalCSV,
  });
  const graphRequest = page.waitForRequest((request) =>
    request.url().endsWith("/network-flow/graphs/query"),
  );
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();

  const request = await graphRequest;
  const body = requestJSON(request);
  expect(body.table_scope).toMatchObject({ mode: "active_table" });
  expect(
    String((body.table_scope as { active_table_id?: string }).active_table_id),
  ).toMatch(/^nft_[a-z0-9]+$/);
  await expect(
    page
      .getByTestId(networkAnalysisTestId("graph-scope"))
      .getByRole("radio", { name: "Active table", exact: true }),
  ).toBeChecked();
});

test("Network Analysis vertex selection uses stable graph identity", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFAC036");
  await importNetworkFlowCSV(page, {
    displayName: "vertex-source",
    file: networkFlowMinimalCSV,
  });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();

  const vertex = page.getByTestId(/^network-flow-vertex-/).first();
  await expect(vertex).toBeVisible();
  const semanticTestId = await vertex.getAttribute("data-testid");
  expect(semanticTestId).toMatch(/^network-flow-vertex-/);
  await vertex.getByRole("button", { name: "Select vertex" }).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("contributor-drawer")),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("contributor-close")).click();
  await page.getByTestId(networkAnalysisTestId("mode-rows")).click();
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  await expect(
    page.getByTestId(semanticTestId ?? "missing-semantic-id"),
  ).toBeVisible();
});

test("Network Analysis edge selection opens ordered contributor drawer", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFAC037");
  await importNetworkFlowCSV(page, {
    displayName: "contributors-a",
    file: networkFlowMinimalCSV,
  });
  await importNetworkFlowCSV(page, {
    displayName: "contributors-b",
    file: networkFlowMinimalCSV,
  });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  await page
    .getByTestId(networkAnalysisTestId("graph-scope"))
    .getByLabel("All active tables")
    .check();

  const edge = page.getByTestId(/^network-flow-edge-/).first();
  await expect(edge).toBeVisible();
  await edge.getByRole("button", { name: "Select edge" }).click();
  const drawer = page.getByTestId(networkAnalysisTestId("contributor-drawer"));
  await expect(drawer).toBeVisible();
  await expect(drawer).toContainText("contributors-a");
  await expect(drawer).toContainText("contributors-b");
});

test("Network Analysis alias collision requires explicit approval", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFAC074");
  await page.getByTestId(networkAnalysisTestId("import-input")).setInputFiles({
    name: "alias-collision.csv",
    mimeType: "text/csv",
    buffer: Buffer.from(aliasCollisionCSV()),
  });

  const dialog = page.getByTestId(networkAnalysisTestId("mapping-dialog"));
  await expect(dialog).toBeVisible();
  await expect(
    dialog.getByRole("alert", { name: "Alias collision" }),
  ).toBeVisible();
  await expect(
    page.getByTestId(networkAnalysisTestId("mapping-preview")),
  ).toBeDisabled();
  await dialog
    .getByLabel("Target for Source IP · column 3")
    .selectOption("network_flow.src_ip");
  await dialog
    .getByLabel("Target for Source IP · column 4")
    .selectOption("__ignore__");
  await expect(
    dialog.getByRole("alert", { name: "Alias collision" }),
  ).toHaveCount(0);
  await page
    .getByTestId(networkAnalysisTestId("mapping-display-name"))
    .fill("alias-collision-approved");
  await page.getByTestId(networkAnalysisTestId("mapping-preview")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("mapping-preview-summary")),
  ).toBeVisible();
  await expect(
    page.getByTestId(networkAnalysisTestId("mapping-apply")),
  ).toBeEnabled();
  await page.getByTestId(networkAnalysisTestId("mapping-apply")).click();
  await expect(
    page.getByRole("tab", { name: /alias-collision-approved/ }),
  ).toBeVisible();
});

test("Verify Network Analysis clears protected grid, inspector, graph, contributor, and selection state after lifecycle loss and refetches after recovery.", async ({
  page,
}) => {
  const socketMonitorRef: {
    current: ReturnType<typeof installIncidentSocketMonitor> | null;
  } = { current: null };
  await openClaimedNetworkAnalysis(page, "NETWORKFLOWSTATE", {
    onIncidentCreated: (incidentId) => {
      socketMonitorRef.current = installIncidentSocketMonitor(page, incidentId);
    },
  });
  const lifecycleSocket = socketMonitorRef.current;
  if (lifecycleSocket === null) {
    throw new Error("network_flow_lifecycle_socket_monitor_not_installed");
  }
  await lifecycleSocket.waitForMessage("hello_ack");
  await importNetworkFlowCSV(page, { displayName: "lifecycle-source" });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  const edge = page.getByTestId(/^network-flow-edge-/).first();
  await expect(edge).toBeVisible();
  await edge.getByRole("button", { name: "Select edge" }).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("contributor-drawer")),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("contributor-close")).click();

  const deleteEventStartAt = lifecycleSocket.messageCount();
  await page.getByTestId(networkAnalysisTestId("delete-trigger")).click();
  await page
    .getByTestId(networkAnalysisTestId("delete-confirmation"))
    .fill("lifecycle-source");
  await page.getByTestId(networkAnalysisTestId("delete-confirm")).click();

  await expect(
    page.getByTestId(networkAnalysisTestId("contributor-drawer")),
  ).toHaveCount(0);
  await expect(page.getByTestId(/^network-flow-(?:edge|vertex)-/)).toHaveCount(
    0,
  );
  await expect(
    page.getByLabel("Empty Network Analysis workspace"),
  ).toBeVisible();
  await lifecycleSocket.waitForMessage("extension_resource_changed", {
    startAt: deleteEventStartAt,
    matches: (message) =>
      message.payload.extension_profile_id === "network_flow_activity" &&
      message.payload.resource_kind === "network_flow_table" &&
      message.payload.change_kind === "remove" &&
      typeof message.payload.resource_id === "string" &&
      message.payload.resource_id.length > 0,
  });

  await importNetworkFlowCSV(page, { displayName: "recovery-source" });
  await expect(
    page.getByTestId(networkAnalysisTestId("accepted-grid")),
  ).toBeVisible();
  await expect(page.getByRole("tab", { name: /lifecycle-source/ })).toHaveCount(
    0,
  );
  await expect(
    page.getByRole("tab", { name: /recovery-source/ }),
  ).toBeVisible();
});

test("Verify claimed Network Analysis discovery, import, mapping approval, semantic grids, graph, contributors, and lifecycle controls through the real application.", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NETWORKFLOWBROWSER");
  await importNetworkFlowCSV(page, { displayName: "browser-evidence" });
  await expect(
    page.getByTestId(networkAnalysisTestId("accepted-grid")),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("mode-rejected")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("rejected-grid")),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  const edge = page.getByTestId(/^network-flow-edge-/).first();
  await expect(edge).toBeVisible();
  await edge.getByRole("button", { name: "Select edge" }).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("contributor-drawer")),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("contributor-close")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("rename-trigger")),
  ).toBeVisible();
  await expect(
    page.getByTestId(networkAnalysisTestId("delete-trigger")),
  ).toBeVisible();
});

function requestJSON(request: Request): Record<string, unknown> {
  return JSON.parse(request.postData() ?? "{}") as Record<string, unknown>;
}

function aliasCollisionCSV(): string {
  return [
    "Start Time,End Time,Source IP,Source IP,Destination IP,Source Port,Destination Port,Protocol,Bytes,Packets",
    "2026-07-10T12:00:00Z,2026-07-10T12:00:05Z,192.0.2.10,192.0.2.10,192.0.2.20,443,51515,6,1200,12",
  ].join("\n");
}
