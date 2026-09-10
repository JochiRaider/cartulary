import { Buffer } from "node:buffer";
import {
  networkAnalysisSavedGraphTestId,
  networkAnalysisTestId,
  surfaceTabTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import type { Request } from "@playwright/test";

import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  expectNetworkFlowRuntimeProfile,
  importNetworkFlowCSV,
  networkFlowMinimalCSV,
  openClaimedNetworkAnalysis,
  openNetworkFlowIncident,
} from "./support/extensions/network_flow_activity/workspace";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { apiBase } from "./support/runtime/configuration";
import { uniqueTxn } from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";

test("Network Analysis authors lossless typed queries and applies exact row and graph scopes", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFQUERY");
  const fixture =
    [
      "Start Time,End Time,Source IP,Destination IP,Source Port,Destination Port,Protocol,Bytes,Packets,Input Interface,Output Interface",
      "2026-07-10T12:00:00Z,2026-07-10T12:00:00Z,192.0.2.1,192.0.2.2,443,80,6,18446744073709551615,1,01,out",
      "2026-07-10T12:00:00.000001Z,2026-07-10T12:01:00Z,192.0.2.1,192.0.2.2,53,80,17,9007199254740993,2,1,out",
      "2026-07-10T11:59:59Z,2026-07-10T12:00:00Z,192.0.2.1,192.0.2.2,0,80,1,0,3,,out",
      "2026-07-10T12:01:00Z,2026-07-10T12:01:05Z,192.0.2.1,192.0.2.2,443,80,6,42,4,01,out",
      "2026-07-10T12:00:30Z,2026-07-10T12:00:31Z,192.0.2.1,192.0.2.2,53,80,17,100,5,1,out",
      "2026-07-10T12:00:30Z,2026-07-10T12:00:31Z,invalid,192.0.2.2,53,80,17,100,5,1,out",
      "2026-07-10T12:00:30Z,2026-07-10T12:00:31Z,192.0.2.1,invalid,53,80,17,100,5,1,out",
    ].join("\n") + "\n";
  await importNetworkFlowCSV(page, {
    displayName: "query-a",
    file: {
      name: "query.csv",
      mimeType: "text/csv",
      buffer: Buffer.from(fixture),
    },
  });
  const advanced = page.getByTestId(networkAnalysisTestId("advanced-filters"));
  const apply = page.getByRole("button", { name: "Apply query", exact: true });
  const requests: string[] = [];
  page.on("request", (request) => {
    if (/\/tables\/nft_[^/]+\/query$/.test(request.url()))
      requests.push(request.postData() ?? "");
  });
  const closeAdvanced = async () => {
    if ((await advanced.getAttribute("open")) !== null)
      await advanced.locator("summary").click();
  };
  const add = async (field: string, op: string) => {
    if ((await advanced.getAttribute("open")) === null)
      await advanced.locator("summary").click();
    await advanced
      .getByRole("button", { name: "Add filter", exact: true })
      .click();
    const editor = advanced
      .locator("fieldset")
      .filter({ has: page.getByLabel("Field", { exact: true }) })
      .last();
    await editor.getByLabel("Field", { exact: true }).selectOption(field);
    await editor.getByLabel("Operator", { exact: true }).selectOption(op);
    return editor;
  };
  const applyRows = async (expected: number[]) => {
    await closeAdvanced();
    const response = page.waitForResponse(
      (r) =>
        /\/tables\/nft_[^/]+\/query$/.test(r.url()) &&
        r.request().method() === "POST",
    );
    await apply.click();
    const result = await response;
    expect(result.status()).toBe(200);
    const body = await result.json();
    expect(
      body.data.rows
        .map((row: { source_row_number: number }) => row.source_row_number)
        .sort((a: number, b: number) => a - b),
    ).toEqual(expected);
    await expect(
      page.getByRole("status").filter({ hasText: "Query applied." }),
    ).toBeVisible();
    return body.data.rows as {
      "network_flow.bytes_count": string;
      source_row_number: number;
    }[];
  };
  const clear = async () => {
    await closeAdvanced();
    const response = page.waitForResponse((r) =>
      /\/tables\/nft_[^/]+\/query$/.test(r.url()),
    );
    await page
      .getByRole("button", { name: "Clear query", exact: true })
      .click();
    expect((await response).status()).toBe(200);
  };
  let editor = await add("network_flow.ip_protocol", "in");
  await editor.getByLabel("Value 1", { exact: true }).fill("6");
  await editor.getByRole("button", { name: "Add value", exact: true }).click();
  await editor.getByLabel("Value 2", { exact: true }).fill("17");
  await applyRows([2, 3, 5, 6]);
  const count = requests.length;
  await apply.focus();
  await apply.press("Enter");
  await expect(
    page.getByRole("status").filter({ hasText: "Query applied." }),
  ).toBeVisible();
  expect(requests).toHaveLength(count);
  await page.getByLabel("Endpoint IP value").fill("192.0.2.1");
  await applyRows([2, 3, 5, 6]);
  expect(JSON.parse(requests.at(-1) ?? "{}").filters).toContainEqual({
    field_key: "network_flow.ip_protocol",
    op: "in",
    value: [6, 17],
  });
  editor = await add("network_flow.bytes_count", "eq");
  await editor
    .getByLabel("Value", { exact: true })
    .fill("18446744073709551615");
  expect((await applyRows([2]))[0]?.["network_flow.bytes_count"]).toBe(
    "18446744073709551615",
  );
  await advanced.locator("summary").click();
  await advanced.getByRole("button", { name: /Edit bytes count eq/ }).click();
  await advanced
    .getByLabel("Value", { exact: true })
    .fill("18446744073709551616");
  await closeAdvanced();
  const beforeInvalid = requests.length;
  await apply.focus();
  await apply.press("Enter");
  await expect(advanced.getByLabel("Value", { exact: true })).toBeFocused();
  await expect(advanced.getByLabel("Value", { exact: true })).toHaveAttribute(
    "aria-invalid",
    "true",
  );
  expect(requests).toHaveLength(beforeInvalid);
  await expect(
    advanced.getByLabel("Value", { exact: true }),
  ).toHaveAccessibleDescription(
    /Enter an integer from 0 through 18446744073709551615/,
  );
  await expect(
    page.getByRole("alert").filter({ hasText: "Enter an integer" }),
  ).toBeVisible();
  await closeAdvanced();
  const appliedSummary = page.locator("details").filter({
    has: page.locator("summary").filter({ hasText: /^Applied filters/ }),
  });
  await appliedSummary.locator("summary").click();
  await expect(appliedSummary).toContainText(
    'bytes count eq "18446744073709551615"',
  );
  await appliedSummary.locator("summary").click();
  await clear();
  editor = await add("network_flow.bytes_count", "range");
  await editor.getByLabel("At most (inclusive)").fill("42");
  await applyRows([4, 5]);
  await advanced.locator("summary").click();
  await editor.getByLabel("At least (inclusive)").fill("42");
  await applyRows([5]);
  await clear();
  editor = await add("network_flow.input_interface", "is_null");
  await applyRows([4]);
  expect(JSON.parse(requests.at(-1) ?? "{}").filters).toEqual([
    { field_key: "network_flow.input_interface", op: "is_null" },
  ]);
  await clear();
  await page.getByTestId(networkAnalysisTestId("mode-rejected")).click();
  await page
    .getByRole("button", { name: "Add error code", exact: true })
    .click();
  await page
    .getByLabel("Error code 1", { exact: true })
    .selectOption("network_flow_invalid_ip");
  await page
    .getByRole("button", { name: "Add field key", exact: true })
    .click();
  await page
    .getByLabel("Field key 1", { exact: true })
    .selectOption("network_flow.src_ip");
  await page
    .getByRole("button", { name: "Add field key", exact: true })
    .click();
  await page
    .getByLabel("Field key 2", { exact: true })
    .selectOption("network_flow.dst_ip");
  await page.getByLabel("First source row").fill("7");
  await page.getByLabel("Last source row").fill("8");
  const diagnosticsResponse = page.waitForResponse((r) =>
    r.url().endsWith("/rejected-rows/query"),
  );
  await page
    .getByRole("button", { name: "Apply diagnostics query", exact: true })
    .click();
  const diagnostics = await diagnosticsResponse;
  expect(diagnostics.status()).toBe(200);
  expect(
    (await diagnostics.json()).data.diagnostics
      .map((d: { source_row_number: number }) => d.source_row_number)
      .toSorted(),
  ).toEqual([7, 8]);
  expect(
    JSON.parse(diagnostics.request().postData() ?? "{}").field_keys,
  ).toEqual(["network_flow.src_ip", "network_flow.dst_ip"]);
  await page.getByTestId(networkAnalysisTestId("mode-rows")).click();
  editor = await add("network_flow.input_interface", "eq");
  await editor.getByLabel("Value", { exact: true }).fill("x".repeat(256));
  const popup = advanced.locator(".network-flow-popover");
  expect(
    await popup.evaluate(
      (element) => element.scrollWidth <= element.clientWidth + 1,
    ),
  ).toBe(true);
  await applyRows([]);
  await advanced.locator("summary").click();
  const beforeRemoval = requests.length;
  await advanced
    .getByRole("button", { name: /Remove input interface eq/ })
    .click();
  expect(requests).toHaveLength(beforeRemoval);
  await applyRows([2, 3, 4, 5, 6]);
  await page.getByLabel("Flow overlap starts at").fill("2026-07-10T12:00:00Z");
  await page
    .getByLabel("Flow overlap ends before")
    .fill("2026-07-10T12:01:00Z");
  await applyRows([2, 3, 4, 6]);
  const graphResponse = page.waitForResponse((r) =>
    r.url().endsWith("/graphs/query"),
  );
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  const graph = await (await graphResponse).json();
  expect(
    graph.data.edge_annotations.reduce(
      (sum: number, edge: { example_refs_total_count: number }) =>
        sum + edge.example_refs_total_count,
      0,
    ),
  ).toBe(4);
  await page.getByLabel("Time buckets").check();
  const temporalResponse = page.waitForResponse((r) =>
    r.url().endsWith("/graphs/query"),
  );
  await apply.click();
  const temporal = await temporalResponse;
  expect(temporal.status()).toBe(200);
  expect(
    (await temporal.json()).data.result_variant.time_buckets[0]
      .contributing_row_count,
  ).toBe(3);
  await page.getByTestId(networkAnalysisTestId("mode-rows")).click();
  await clear();
  await importNetworkFlowCSV(page, { displayName: "query-b" });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  await expect(page.getByTestId(/^network-flow-vertex-/).first()).toBeVisible();
  const scope = page.getByTestId(networkAnalysisTestId("graph-scope"));
  await scope.getByLabel("Selected tables", { exact: true }).check();
  await scope.getByLabel("query-a", { exact: true }).check();
  await scope.getByLabel("query-b", { exact: true }).check();
  const scopedResponse = page.waitForResponse((r) =>
    r.url().endsWith("/graphs/query"),
  );
  await apply.click();
  const scoped = await scopedResponse;
  expect(scoped.status()).toBe(200);
  const scopedBody = await scoped.json();
  const selected = JSON.parse(scoped.request().postData() ?? "{}").table_scope
    .selected_table_ids;
  expect(selected).toHaveLength(2);
  expect(scopedBody.data.semantic_query.selected_table_ids.toSorted()).toEqual(
    selected.toSorted(),
  );
  expect(
    scopedBody.data.edge_annotations.reduce(
      (sum: number, edge: { example_refs_total_count: number }) =>
        sum + edge.example_refs_total_count,
      0,
    ),
  ).toBe(8);
});

test("Network Analysis links compatible targets and recovers exact committed requests across workspace departure", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFLINKRECOVERY");
  await importNetworkFlowCSV(page, { displayName: "link-source" });
  const sourceCell = page
    .getByRole("gridcell", { name: /Source IP: 192\.0\.2\.10/u })
    .first();
  await sourceCell.click();
  const trigger = page.getByRole("button", {
    name: "Link 1 selected row",
    exact: true,
  });
  await trigger.focus();
  await trigger.press("Enter");
  const dialog = page.getByTestId(
    networkAnalysisTestId("indicator-link-dialog"),
  );
  const confirmation = page.getByTestId(
    networkAnalysisTestId("indicator-link-confirmation"),
  );
  const submit = page.getByTestId(
    networkAnalysisTestId("indicator-link-submit"),
  );
  await expect(confirmation).toBeFocused();
  await confirmation.fill("192.0.2.10 ");
  await confirmation.press("Enter");
  await expect(confirmation).toHaveAttribute("aria-invalid", "true");
  await expect(
    dialog.getByText(/Enter the canonical candidate exactly/u),
  ).toBeVisible();
  await page.keyboard.press("Escape");
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
  await trigger.press("Enter");
  await confirmation.fill("192.0.2.10");
  const requests: string[] = [];
  let loseReceipt = true;
  let committedBinding = "";
  await page.route("**/network-flow/indicator-links", async (route) => {
    requests.push(route.request().postData() ?? "");
    const response = await route.fetch();
    if (loseReceipt) {
      loseReceipt = false;
      expect(response.status()).toBe(201);
      committedBinding = (await response.json()).data.binding
        .network_flow_indicator_binding_id;
      await route.abort("failed");
    } else await route.fulfill({ response });
  });
  await submit.press("Enter");
  await expect(dialog.getByRole("alert")).toContainText("may have committed");
  await page.keyboard.press("Escape");
  await page
    .getByTestId(surfaceTabTestId("cartulary.view.timeline.v2"))
    .click();
  const recovery = page.getByRole("button", {
    name: "Review retained indicator link",
  });
  await expect(recovery).toBeVisible();
  await recovery.focus();
  await recovery.press("Enter");
  await expect(dialog).toBeVisible();
  await dialog
    .getByRole("button", { name: "Replay exact request", exact: true })
    .press("Enter");
  await expect(dialog.getByRole("status")).toHaveText(
    "Indicator binding created.",
  );
  await expect(dialog).toContainText(committedBinding);
  await expect(dialog).toContainText(
    "Recovered by replaying the exact original request.",
  );
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  await dialog
    .getByRole("button", { name: "Done", exact: true })
    .press("Enter");
  await expect(recovery).toHaveCount(0);
  await expect(page.getByTestId(networkAnalysisTestId("tab"))).toBeFocused();
  await page.getByTestId(networkAnalysisTestId("tab")).click();
  await sourceCell.click();
  await trigger.click();
  await dialog.getByLabel("Existing indicator", { exact: true }).check();
  const choices = dialog.getByLabel("Compatible indicators on this page", {
    exact: true,
  });
  await expect(choices.locator("option")).toHaveCount(2);
  await choices.focus();
  await choices.press("ArrowDown");
  await choices.press("Enter");
  await expect(
    dialog.getByLabel("Known indicator ID (alternative)"),
  ).not.toHaveValue("");
  await confirmation.fill("192.0.2.10");
  await confirmation.press("Enter");
  await expect(dialog.getByRole("status")).toHaveText(
    "Existing indicator binding reused.",
  );
  await expect(dialog).toContainText(committedBinding);
  expect(requests).toHaveLength(3);
  expect(JSON.parse(requests[2] ?? "{}").client_txn_id).not.toBe(
    JSON.parse(requests[0] ?? "{}").client_txn_id,
  );
  await test.info().attach("indicator-link-reused-receipt", {
    body: await dialog.screenshot(),
    contentType: "image/png",
  });
  await dialog
    .getByRole("button", { name: "Done", exact: true })
    .press("Enter");
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  await page
    .getByTestId(/^network-flow-edge-/u)
    .first()
    .getByRole("button", { name: /^Select edge/u })
    .click();
  await page
    .getByRole("button", { name: "Link destination", exact: true })
    .click();
  const value = await dialog.locator(".network-flow-link-value").innerText();
  await confirmation.fill(value);
  await confirmation.press("Enter");
  await expect(dialog.getByRole("status")).toHaveText(
    "Indicator binding created.",
  );
  await dialog
    .getByRole("button", { name: "Done", exact: true })
    .press("Enter");
});

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
  const incidentId = await openClaimedNetworkAnalysis(page, "NFAC002");

  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toHaveAttribute("data-extension-profile-id", "network_flow_activity");
  await expect(
    page.getByLabel("Empty Network Analysis workspace"),
  ).toBeVisible();
  await expect(
    page.getByTestId(networkAnalysisTestId("import-trigger")),
  ).toHaveAttribute("data-network-flow-variant", "primary");
  const duplicate = await page.context().newPage();
  try {
    await duplicate.goto(`/?incident_id=${incidentId}`);
    await duplicate.getByTestId(networkAnalysisTestId("tab")).click();
    await expect(
      duplicate.getByTestId(networkAnalysisTestId("workspace")),
    ).toBeVisible();
    await expect(
      page.getByTestId(workbookPresenceSummaryTestId()),
    ).toHaveAccessibleName(/^1 collaborator present on this sheet/);
    await expect(
      page.getByRole("img", {
        name: /^\d+ collaborators? (on this row|editing .+ on this row)/,
      }),
    ).toHaveCount(0);
    await duplicate.goto(
      `/?incident_id=${incidentId}&view_schema_id=cartulary.view.timeline.v2`,
    );
    await expect(
      page.getByTestId(workbookPresenceSummaryTestId()),
    ).toHaveAccessibleName(/^0 collaborators/);
  } finally {
    await duplicate.close();
  }
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

test("Network Analysis recovers discovery selection exact apply and published table handoff", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFRECOVERY");
  let uploads = 0,
    approvals = 0,
    sourceFailures = 0,
    observationFailures = 0;
  let applyJobId = "",
    blockTables = false;
  const selections: unknown[] = [],
    applies: unknown[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith("/api/v1/import-sessions")
    )
      uploads++;
    if (request.method() === "PUT" && request.url().endsWith("/mapping"))
      approvals++;
  });
  await page.route(
    "**/api/v1/import-sessions/*/units/*/preview",
    async (route) => {
      if (sourceFailures++ === 0) await route.abort("failed");
      else await route.continue();
    },
  );
  await page.route(
    "**/api/v1/import-sessions/*/units/*/select",
    async (route) => {
      selections.push(route.request().postDataJSON());
      if (selections.length === 1)
        await route.fulfill({
          status: 409,
          contentType: "application/json",
          body: JSON.stringify({
            error: {
              code: "import_apply_blocked",
              message: "Selection must be retried.",
              retryable: false,
              details: { reason_code: "unit_not_ready" },
            },
          }),
        });
      else await route.continue();
    },
  );
  await page.route("**/api/v1/import-sessions/*/apply", async (route) => {
    applies.push(route.request().postDataJSON());
    if (applies.length === 1) {
      const response = await route.fetch();
      expect(response.status()).toBe(202);
      applyJobId = (await response.json()).data.job_id;
      blockTables = true;
      await route.abort("failed");
    } else await route.continue();
  });
  await page.route("**/api/v1/jobs/*", async (route) => {
    if (
      applyJobId &&
      route.request().url().endsWith(applyJobId) &&
      observationFailures++ === 0
    )
      await route.abort("failed");
    else await route.continue();
  });
  await page.route("**/network-flow/tables", async (route) => {
    if (blockTables) await route.abort("failed");
    else await route.continue();
  });
  await page
    .getByTestId(networkAnalysisTestId("import-input"))
    .setInputFiles(networkFlowMinimalCSV);
  const dialog = page.getByTestId(networkAnalysisTestId("mapping-dialog"));
  await dialog
    .getByTestId(networkAnalysisTestId("import-source-reload"))
    .click();
  await dialog
    .getByTestId(networkAnalysisTestId("mapping-display-name"))
    .fill("recovered-table");
  await dialog.getByTestId(networkAnalysisTestId("mapping-preview")).click();
  await expect(
    dialog.getByTestId(networkAnalysisTestId("mapping-preview-summary")),
  ).toContainText("only the preview slice");
  await dialog.getByTestId(networkAnalysisTestId("mapping-apply")).click();
  await expect(
    dialog.getByTestId(networkAnalysisTestId("import-progress")),
  ).toContainText("selecting");
  await dialog.getByTestId(networkAnalysisTestId("import-replay")).click();
  await expect(
    dialog.getByTestId(networkAnalysisTestId("import-replay")),
  ).toHaveText("Retry exact request");
  await expect(
    dialog.getByTestId(networkAnalysisTestId("mapping-preview")),
  ).toBeDisabled();
  await dialog.getByTestId(networkAnalysisTestId("import-replay")).click();
  await dialog.getByTestId(networkAnalysisTestId("import-resume")).click();
  await expect(
    dialog.getByTestId(networkAnalysisTestId("import-progress")),
  ).toContainText("Import succeeded");
  blockTables = false;
  await dialog.getByTestId(networkAnalysisTestId("import-handoff")).click();
  await expect(dialog).toHaveCount(0);
  await expect(
    page.getByRole("tab", { name: /recovered-table/u }),
  ).toHaveAttribute("aria-selected", "true");
  expect(uploads).toBe(1);
  expect(approvals).toBe(1);
  expect(selections).toHaveLength(2);
  expect(applies).toHaveLength(2);
  expect(selections[0]).not.toEqual(selections[1]);
  expect(applies[0]).toEqual(applies[1]);
  await page.getByTestId(networkAnalysisTestId("import-trigger")).click();
  await dialog
    .getByRole("button", { name: "Start another import", exact: true })
    .click();
  await expect(dialog).toHaveCount(0);
  await expect(
    page.getByTestId(networkAnalysisTestId("import-trigger")),
  ).toContainText("Import");
  expect(uploads).toBe(1);
});

test("Network Analysis retains mapping review across workspace departure and incident closure", async ({
  page,
}) => {
  const incidentId = await openClaimedNetworkAnalysis(page, "NFRETAIN");
  const trigger = page.getByTestId(networkAnalysisTestId("import-trigger"));
  await trigger.focus();
  await page
    .getByTestId(networkAnalysisTestId("import-input"))
    .setInputFiles(networkFlowMinimalCSV);
  const dialog = page.getByTestId(networkAnalysisTestId("mapping-dialog"));
  await dialog
    .getByTestId(networkAnalysisTestId("mapping-display-name"))
    .fill("copyable-retained-mapping");
  await dialog.press("Escape");
  await expect(trigger).toBeFocused();
  await page
    .getByTestId(surfaceTabTestId("cartulary.view.timeline.v2"))
    .click();
  await page.getByTestId(networkAnalysisTestId("import-recovery")).click();
  await expect(
    dialog.getByTestId(networkAnalysisTestId("mapping-display-name")),
  ).toHaveValue("copyable-retained-mapping");
  await dialog.press("Escape");
  await page.getByTestId(networkAnalysisTestId("tab")).click();
  await trigger.click();
  await expect(
    dialog.getByTestId(networkAnalysisTestId("mapping-display-name")),
  ).toHaveValue("copyable-retained-mapping");
  await dialog.press("Escape");
  const incident = await currentLifecycle(page, incidentId);
  expect(
    (
      await lifecycleAction(page, incidentId, "closeIncident", {
        base_incident_version: incident.incident_version,
        client_txn_id: uniqueTxn("nf-close"),
        reason: "Verify retained analytical draft",
      })
    ).ok,
  ).toBe(true);
  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toHaveCount(0);
  await page.getByTestId(networkAnalysisTestId("import-recovery")).click();
  await expect(dialog).toContainText("Closed, read-only");
  await expect(
    dialog.getByLabel("Retained mapping draft (copy only)"),
  ).toContainText("copyable-retained-mapping");
  await expect(
    dialog.getByTestId(networkAnalysisTestId("mapping-preview")),
  ).toBeDisabled();
  await expect(
    dialog.getByTestId(networkAnalysisTestId("mapping-apply")),
  ).toBeDisabled();
  await expect(
    dialog.getByTestId(networkAnalysisTestId("import-cancel")),
  ).toHaveCount(0);
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
  await expect(scope.getByRole("checkbox", { name: "scope-b" })).toBeChecked();
  await scope.getByRole("checkbox", { name: "scope-a" }).check();
  await expect(scope.getByRole("checkbox", { name: "scope-a" })).toBeChecked();
  await page.getByRole("button", { name: "Apply query", exact: true }).click();
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
  await vertex.getByRole("button", { name: /^Select vertex/u }).click();
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
  await page.getByRole("button", { name: "Apply query", exact: true }).click();

  const edge = page.getByTestId(/^network-flow-edge-/).first();
  await expect(edge).toBeVisible();
  await edge.getByRole("button", { name: /^Select edge/u }).click();
  const drawer = page.getByTestId(networkAnalysisTestId("contributor-drawer"));
  await expect(drawer).toBeVisible();
  await expect(drawer).toContainText("contributors-a");
  await expect(drawer).toContainText("contributors-b");
});

test("Network Analysis saved graphs complete exact-result lifecycle through the real application", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFGRAPHVIEW");
  await importNetworkFlowCSV(page, {
    displayName: "saved-graph-source",
    file: networkFlowMinimalCSV,
  });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  await expect(page.getByTestId(/^network-flow-vertex-/).first()).toBeVisible();
  await page.getByLabel("Time buckets").check();
  await page.getByRole("button", { name: "Apply query", exact: true }).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")).getByRole("alert"),
  ).toContainText("require both UTC range bounds");
  await page.getByLabel("Flow starts at or after").fill("2026-07-10T12:00:00Z");
  await page.getByLabel("Flow starts before").fill("2026-07-10T14:00:00Z");
  const temporalGraphRequest = page.waitForRequest(
    (request) =>
      request.url().endsWith("/network-flow/graphs/query") &&
      request.method() === "POST" &&
      requestJSON(request).schema_id ===
        "cartulary.network_flow.graph_query_request.v2",
  );
  await page.getByTestId(networkAnalysisTestId("accepted-query-apply")).click();
  const temporalBody = requestJSON(await temporalGraphRequest);
  expect(temporalBody.aggregation).toMatchObject({
    mode: "time_bucket_v1",
    bucket_width_seconds: 3600,
  });
  const bucketNavigation = page.getByRole("navigation", {
    name: "Time bucket navigation",
  });
  await expect(bucketNavigation).toContainText("Bucket 1 of 2");
  await expect(page.getByTestId(/^network-flow-edge-/).first()).toBeVisible();
  await bucketNavigation.getByRole("button", { name: "Next bucket" }).focus();
  await page.keyboard.press("Enter");
  await expect(bucketNavigation).toContainText("0 vertices · 0 edges · 0 rows");
  await expect(page.getByTestId(/^network-flow-edge-/)).toHaveCount(0);
  await bucketNavigation
    .getByRole("button", { name: "Previous bucket" })
    .focus();
  await page.keyboard.press("Enter");
  await expect(page.getByTestId(/^network-flow-edge-/).first()).toBeVisible();
  await page.getByRole("button", { name: "Saved graphs" }).click();

  const panel = page.getByRole("region", {
    name: "Saved Network Flow graphs",
  });
  await panel.getByRole("button", { name: "Save current graph" }).click();
  await page
    .getByRole("textbox", { name: "Display name" })
    .fill("Evidence graph");
  await page.getByRole("button", { name: "Save graph" }).click();
  await expect(
    panel.getByTestId(networkAnalysisTestId("saved-graph-heading")),
  ).toHaveText("Evidence graph");
  await expect(
    panel.getByText("Materialization succeeded.", { exact: true }),
  ).toBeVisible({ timeout: 15_000 });
  await expect(
    page.getByTestId(/^network-flow-saved-graph-vertex-/u).first(),
  ).toBeVisible();
  await expect(
    page.getByTestId(/^network-flow-saved-graph-edge-/u).first(),
  ).toBeVisible();

  await page
    .getByTestId(/^network-flow-saved-graph-vertex-/u)
    .first()
    .getByRole("button")
    .click();
  const contributors = page.getByRole("complementary", {
    name: "Saved graph contributors",
  });
  await expect(contributors).toBeVisible();
  await expect(contributors).toContainText("Row");
  await contributors.getByRole("button", { name: "Close" }).click();

  await panel.getByRole("button", { name: "Rename" }).click();
  await page
    .getByRole("textbox", { name: "Display name" })
    .fill("Renamed evidence graph");
  await page.getByRole("button", { name: "Rename graph" }).click();
  await expect(
    panel.getByTestId(networkAnalysisTestId("saved-graph-heading")),
  ).toHaveText("Renamed evidence graph");

  await panel.getByRole("button", { name: "Refresh" }).click();
  await page.getByRole("button", { name: "Refresh graph" }).click();
  await expect(
    panel.getByText(
      "Showing the last successful result while refresh continues.",
    ),
  ).toBeVisible();
  await expect(
    page.getByTestId(/^network-flow-saved-graph-vertex-/u).first(),
  ).toBeVisible();
  await expect(
    panel.getByText("Materialization succeeded.", { exact: true }),
  ).toBeVisible({ timeout: 15_000 });

  await panel.getByRole("button", { name: "Retire" }).click();
  await page.getByRole("button", { name: "Retire graph" }).click();
  await expect(panel.getByText("No saved graphs yet.")).toBeVisible();
  await expect(
    panel.getByRole("button", { name: "Reload", exact: true }),
  ).toBeFocused();
});

test("Network Analysis saved graphs recover denied replay and withdrawn read authority", async ({
  page,
}) => {
  await openClaimedNetworkAnalysis(page, "NFGRAPHRECOVERY");
  await importNetworkFlowCSV(page, {
    displayName: "recovery-source",
    file: networkFlowMinimalCSV,
  });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  await expect(page.getByTestId(/^network-flow-vertex-/).first()).toBeVisible();
  await page.getByRole("button", { name: "Saved graphs" }).click();
  const panel = page.getByRole("region", { name: "Saved Network Flow graphs" });
  const writes: string[] = [];
  let denyList = false;
  let failOperationReads = false;
  await page.route("**/api/v1/jobs/*", async (route) => {
    if (failOperationReads && route.request().method() === "GET")
      await route.abort("failed");
    else await route.continue();
  });
  await page.route("**/network-flow/graph-views/*", async (route) => {
    if (failOperationReads && route.request().method() === "GET")
      await route.abort("failed");
    else await route.continue();
  });
  await page.route("**/network-flow/graph-views", async (route) => {
    if (route.request().method() === "GET") {
      if (failOperationReads) await route.abort("failed");
      else if (denyList)
        await route.fulfill({
          status: 403,
          json: {
            error: {
              code: "authorization_denied",
              status: 403,
              message: "Current read access was withdrawn.",
              request_id: "req-saved-read-denial",
              retryable: false,
              details: {},
            },
          },
        });
      else await route.continue();
      return;
    }
    writes.push(route.request().postData() ?? "");
    if (writes.length === 1) {
      const response = await route.fetch();
      expect(response.status()).toBe(202);
      await route.abort("failed");
    } else if (writes.length === 2) {
      await route.fulfill({
        status: 403,
        json: {
          error: {
            code: "authorization_denied",
            status: 403,
            message: "Replay recovery request denied.",
            request_id: "req-saved-replay-denial",
            retryable: false,
            details: {},
          },
        },
      });
    } else {
      failOperationReads = true;
      await route.continue();
    }
  });
  await panel.getByRole("button", { name: "Save current graph" }).click();
  const dialog = page.getByRole("dialog", { name: "Save current graph" });
  await dialog
    .getByRole("textbox", { name: "Display name" })
    .fill("Recovered evidence");
  await dialog.getByRole("button", { name: "Save graph", exact: true }).click();
  await dialog.getByRole("button", { name: "Replay exact attempt" }).click();
  await expect(dialog.getByRole("alert")).toContainText(
    "Replay recovery request denied",
  );
  await expect(
    dialog.getByRole("textbox", { name: "Display name" }),
  ).toBeDisabled();
  await expect(dialog).toContainText("The outcome is uncertain");
  await dialog.getByRole("button", { name: "Replay exact attempt" }).click();
  expect(writes).toHaveLength(3);
  expect(new Set(writes).size).toBe(1);
  const recovery = panel.getByRole("region", { name: "Saved graph operation" });
  await expect(recovery).toContainText("Recovered evidence");
  await expect(
    recovery.getByRole("button", { name: "Resume operation observation" }),
  ).toBeEnabled();
  await expect(dialog).toHaveCount(0);
  failOperationReads = false;
  await recovery
    .getByRole("button", { name: "Resume operation observation" })
    .click();
  await recovery
    .getByRole("button", { name: "Reload operation graph" })
    .click();
  await expect(
    panel.getByRole("heading", { name: "Recovered evidence" }),
  ).toBeVisible();
  const result = panel.getByTestId(networkAnalysisTestId("saved-graph-result"));
  await expect(result).toBeVisible();
  const vertex = page
    .getByTestId(/^network-flow-saved-graph-vertex-/u)
    .first()
    .getByRole("button");
  let staleContributors = true;
  await page.route(
    "**/network-flow/graph-views/*/contributors/query",
    async (route) => {
      if (staleContributors) {
        staleContributors = false;
        await route.fulfill({
          status: 409,
          json: {
            error: {
              code: "network_flow_graph_query_stale",
              status: 409,
              message: "The binding requires current-state review.",
              request_id: "req-saved-binding",
              retryable: false,
              details: {},
            },
          },
        });
      } else await route.continue();
    },
  );
  await vertex.click();
  await expect(result).toHaveCount(0);
  const declarationRead = page.waitForRequest(
    (request) =>
      request.method() === "GET" &&
      /\/graph-views\/nfgv_[a-f0-9]+$/.test(request.url()),
  );
  await panel.getByRole("button", { name: "Reload saved graph" }).click();
  await declarationRead;
  await expect(result).toBeVisible();
  await vertex.click();
  const contributors = page.getByRole("complementary", {
    name: "Saved graph contributors",
  });
  await expect(contributors).toContainText("Row");
  const retained = await vertex.elementHandle();
  await panel.getByRole("button", { name: "Reload", exact: true }).click();
  await expect(contributors).toContainText("Row");
  expect(
    await vertex.evaluate((node, previous) => node === previous, retained),
  ).toBe(true);
  denyList = true;
  await panel.getByRole("button", { name: "Reload", exact: true }).click();
  await expect(panel.getByRole("alert")).toContainText(
    "Current read access was withdrawn",
  );
  await expect(result).toHaveCount(0);
  await expect(contributors).toHaveCount(0);
  await expect(panel.getByText("No saved graphs yet.")).toHaveCount(0);
  denyList = false;
  await panel.getByRole("button", { name: "Reload", exact: true }).click();
  await expect(result).toBeVisible();
  await retained?.dispose();
});

test("Network Analysis saved graphs fence deferred results and contributors across stable-ID selection", async ({
  page,
}) => {
  const incidentId = await openClaimedNetworkAnalysis(page, "NFGRAPHRACES");
  await importNetworkFlowCSV(page, {
    displayName: "race-source",
    file: networkFlowMinimalCSV,
  });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  await expect(page.getByTestId(/^network-flow-vertex-/).first()).toBeVisible();
  await page.getByRole("button", { name: "Saved graphs" }).click();
  const panel = page.getByRole("region", { name: "Saved Network Flow graphs" });
  const ids: string[] = [];
  for (let index = 0; index < 2; index++) {
    await panel.getByRole("button", { name: "Save current graph" }).click();
    await page
      .getByRole("textbox", { name: "Display name" })
      .fill("Duplicate graph");
    const accepted = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().endsWith("/graph-views"),
    );
    await page.getByRole("button", { name: "Save graph", exact: true }).click();
    ids.push((await (await accepted).json()).data.graph_view.graph_view_id);
    await expect(
      panel.getByTestId(networkAnalysisTestId("saved-graph-result")),
    ).toBeVisible();
  }
  expect(ids[0]).not.toBe(ids[1]);
  const [firstId, secondId] = ids;
  if (!firstId || !secondId)
    throw new Error("Both saved declarations must be acknowledged.");
  const graphA = panel.getByTestId(networkAnalysisSavedGraphTestId(firstId));
  const graphB = panel.getByTestId(networkAnalysisSavedGraphTestId(secondId));
  await expect(graphA).toContainText("Duplicate graph");
  await expect(graphB).toContainText("Duplicate graph");
  await panel.getByRole("button", { name: "Rename", exact: true }).click();
  await page
    .getByRole("textbox", { name: "Display name" })
    .fill("Second graph");
  const confirmation = page.getByRole("dialog", { name: "Rename saved graph" });
  const other = await page.context().newPage();
  try {
    await other.goto(`/?incident_id=${incidentId}`);
    await other.getByTestId(networkAnalysisTestId("tab")).click();
    await other.getByTestId(networkAnalysisTestId("mode-graph")).click();
    await other.getByRole("button", { name: "Saved graphs" }).click();
    await other.getByTestId(networkAnalysisSavedGraphTestId(secondId)).click();
    await other
      .getByRole("region", { name: "Saved Network Flow graphs" })
      .getByRole("button", { name: "Rename", exact: true })
      .click();
    await other
      .getByRole("textbox", { name: "Display name" })
      .fill("Remote graph");
    await other
      .getByRole("button", { name: "Rename graph", exact: true })
      .click();
    await expect(graphB).toContainText("Remote graph");
    await expect(confirmation).toContainText("Target: Duplicate graph");
    await confirmation
      .getByRole("button", { name: "Rename graph", exact: true })
      .click();
    await expect(confirmation.getByRole("alert")).toContainText("Review");
    await confirmation
      .getByRole("button", { name: "Review current graph" })
      .click();
    await expect(confirmation).toContainText("Target: Remote graph");
    await expect(
      confirmation.getByRole("textbox", { name: "Display name" }),
    ).toHaveValue("Second graph");
    await confirmation
      .getByRole("button", { name: "Rename graph", exact: true })
      .click();
    await expect(graphB).toContainText("Second graph");
  } finally {
    await other.close();
  }
  let releaseResult = () => {};
  const resultGate = new Promise<void>((resolve) => {
    releaseResult = resolve;
  });
  let resultStarted = () => {};
  const resultWaiting = new Promise<void>((resolve) => {
    resultStarted = resolve;
  });
  const resultRoute = `**/network-flow/graph-views/${ids[0]}/result`;
  await page.route(resultRoute, async (route) => {
    const response = await route.fetch();
    resultStarted();
    await resultGate;
    await route.fulfill({ response });
  });
  try {
    await graphA.click();
    await resultWaiting;
    await graphB.click();
    await expect(graphB).toHaveAttribute("aria-current", "true");
    releaseResult();
    await expect(
      panel.getByRole("heading", { name: "Second graph" }),
    ).toBeVisible();
    await expect(
      panel.getByTestId(networkAnalysisTestId("saved-graph-result")),
    ).toBeVisible();
  } finally {
    releaseResult();
    await page.unroute(resultRoute);
  }
  await graphA.click();
  await expect(
    panel.getByRole("heading", { name: "Duplicate graph" }),
  ).toBeVisible();
  const vertex = page
    .getByTestId(/^network-flow-saved-graph-vertex-/u)
    .first()
    .getByRole("button");
  await expect(vertex).toBeVisible();
  let releaseContributors = () => {};
  const contributorGate = new Promise<void>((resolve) => {
    releaseContributors = resolve;
  });
  let contributorsStarted = () => {};
  const contributorsWaiting = new Promise<void>((resolve) => {
    contributorsStarted = resolve;
  });
  const contributorRoute = `**/network-flow/graph-views/${ids[0]}/contributors/query`;
  await page.route(contributorRoute, async (route) => {
    const response = await route.fetch();
    contributorsStarted();
    await contributorGate;
    await route.fulfill({ response });
  });
  try {
    await vertex.click();
    await contributorsWaiting;
    await graphB.click();
    releaseContributors();
    await expect(
      panel.getByRole("heading", { name: "Second graph" }),
    ).toBeVisible();
    await expect(
      panel.getByTestId(networkAnalysisTestId("saved-graph-result")),
    ).toBeVisible();
    await expect(
      page.getByRole("complementary", { name: "Saved graph contributors" }),
    ).toHaveCount(0);
    await expect(graphB).toHaveAttribute("aria-current", "true");
  } finally {
    releaseContributors();
    await page.unroute(contributorRoute);
  }
  await page
    .getByTestId(/^network-flow-saved-graph-vertex-/u)
    .first()
    .getByRole("button")
    .click();
  const retainedContributors = page.getByRole("complementary", {
    name: "Saved graph contributors",
  });
  await expect(retainedContributors).toContainText("Row");
  await page.getByTestId(networkAnalysisTestId("delete-trigger")).click();
  await page
    .getByTestId(networkAnalysisTestId("delete-confirmation"))
    .fill("race-source");
  await page.getByTestId(networkAnalysisTestId("delete-confirm")).click();
  await expect(
    panel.getByTestId(networkAnalysisTestId("saved-graph-result")),
  ).toHaveCount(0);
  await expect(retainedContributors).toHaveCount(0);
  await expect(
    panel.getByTestId(/^network-flow-saved-graph-vertex-/u),
  ).toHaveCount(0);
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
  await page.setViewportSize({ width: 390, height: 480 });
  await expect(
    dialog.getByRole("region", { name: "Mapping requirements" }),
  ).toContainText("Resolve alias collisions");
  expect(
    await dialog.evaluate((node) => node.scrollWidth <= node.clientWidth + 1),
  ).toBe(true);
  await page.setViewportSize({ width: 1440, height: 900 });
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
  await edge.getByRole("button", { name: /^Select edge/u }).click();
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
  await expect(page.getByText("lifecycle-source was deleted.")).toBeVisible();

  await importNetworkFlowCSV(page, { displayName: "recovery-source" });
  await expect(
    page.getByTestId(networkAnalysisTestId("mode-graph")),
  ).toHaveAttribute("aria-pressed", "true");
  await expect(
    page.getByTestId(networkAnalysisTestId("status-strip")),
  ).toContainText("graph stale");
  await page
    .getByRole("button", { name: "Recompute graph", exact: true })
    .click();
  await expect(page.getByText("Graph ready", { exact: true })).toBeVisible();
  await page
    .getByTestId(/^network-flow-vertex-/)
    .first()
    .getByRole("button", { name: /^Select vertex/u })
    .click();
  await expect(
    page.getByTestId(networkAnalysisTestId("contributor-drawer")),
  ).toBeVisible();
  await expect(
    page.getByTestId(networkAnalysisTestId("page-status")),
  ).toHaveText("Page 1");
  await page.getByTestId(networkAnalysisTestId("mode-rows")).click();
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
  await edge.getByRole("button", { name: /^Select edge/u }).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("contributor-drawer")),
  ).toBeVisible();
  await page.getByTestId(networkAnalysisTestId("contributor-close")).click();
  await expect(
    page.getByTestId(networkAnalysisTestId("rename-trigger")),
  ).toHaveAttribute("data-network-flow-variant", "secondary");
  await expect(
    page.getByTestId(networkAnalysisTestId("delete-trigger")),
  ).toHaveAttribute("data-network-flow-variant", "danger");
  await expect(
    page.getByTestId(networkAnalysisTestId("mode-graph")),
  ).toHaveAttribute("aria-pressed", "true");
  const appearances = await page
    .getByTestId(networkAnalysisTestId("workspace"))
    .locator("[data-network-flow-control]:visible")
    .evaluateAll((controls) =>
      controls.map((control) => {
        const style = getComputedStyle(control);
        return {
          backgroundColor: style.backgroundColor,
          color: style.color,
        };
      }),
    );
  expect(appearances.length).toBeGreaterThan(10);
  for (const appearance of appearances) {
    expect(appearance.backgroundColor).not.toBe("rgb(255, 255, 255)");
    expect(appearance.color).not.toBe("rgb(0, 0, 0)");
  }
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

test("Network Analysis table dialogs review peer changes preserve graph context and replay exact lost receipts", async ({
  page,
}) => {
  const incidentId = await openClaimedNetworkAnalysis(
    page,
    "NFTABLECONTINUITY",
  );
  await importNetworkFlowCSV(page, { displayName: "table-source" });
  const collection = `${apiBase}/api/v1/incidents/${incidentId}/network-flow/tables`;
  const catalog = await (await page.request.get(collection)).json();
  const tableId = catalog.data.tables[0].network_flow_table_id as string;
  const resource = `${collection}/${tableId}`;
  const graphRequests: Request[] = [];
  page.on("request", (request) => {
    if (request.url().endsWith("/network-flow/graphs/query"))
      graphRequests.push(request);
  });
  await page.getByTestId(networkAnalysisTestId("mode-graph")).click();
  const edge = page
    .getByTestId(/^network-flow-edge-/u)
    .first()
    .getByRole("button", { name: /^Select edge/u });
  await edge.click();
  const contributors = page.getByRole("complementary", {
    name: /contributors/iu,
  });
  await expect(contributors).toBeVisible();
  const count = graphRequests.length;
  const trigger = page.getByTestId(networkAnalysisTestId("rename-trigger"));
  await trigger.focus();
  await trigger.press("Enter");
  const dialog = page.getByTestId(networkAnalysisTestId("rename-dialog"));
  const input = page.getByTestId(networkAnalysisTestId("rename-input"));
  await expect(input).toBeFocused();
  await input.fill("😀".repeat(65));
  await input.press("Enter");
  await expect(dialog.getByRole("alert")).toContainText("64 Unicode");
  await input.fill(" Cafe\u0301 ");
  const peer = await page.request.patch(resource, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("table-peer"),
      base_table_version: 1,
      display_name: "Peer table",
    },
  });
  expect(peer.status()).toBe(200);
  await expect(
    dialog.getByRole("button", { name: "Review current table" }),
  ).toBeEnabled();
  await expect(input).toHaveValue(" Cafe\u0301 ");
  await expect(
    dialog.getByTestId(networkAnalysisTestId("rename-submit")),
  ).toBeDisabled();
  await dialog
    .getByRole("button", { name: "Review current table" })
    .press("Enter");
  await page.getByTestId(networkAnalysisTestId("rename-input")).press("Enter");
  await expect(dialog).toHaveCount(0);
  await expect(trigger).toBeFocused();
  await expect(page.getByRole("tab", { name: /Café/u })).toBeVisible();
  await expect(contributors).toBeVisible();
  expect(graphRequests).toHaveLength(count);

  const requests: string[] = [];
  let lose = true;
  await page.route(`**/network-flow/tables/${tableId}`, async (route) => {
    if (route.request().method() !== "PATCH") {
      await route.continue();
      return;
    }
    requests.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.status()).toBe(200);
    if (lose) {
      lose = false;
      await route.abort("failed");
    } else await route.fulfill({ response });
  });
  await trigger.click();
  await input.fill("Recovered table");
  await input.press("Enter");
  await expect(dialog.getByRole("alert")).toContainText("may have committed");
  await page.keyboard.press("Escape");
  await page
    .getByTestId(surfaceTabTestId("cartulary.view.timeline.v2"))
    .click();
  // A later authorized mutation changes current metadata before historical replay.
  const latest = await page.request.patch(resource, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("table-later"),
      base_table_version: 4,
      display_name: "Latest table",
    },
  });
  expect(latest.status()).toBe(200);
  await page
    .getByRole("button", { name: "Review retained table change" })
    .click();
  await dialog.getByRole("button", { name: "Replay exact request" }).click();
  await expect(dialog).toHaveCount(0);
  expect(requests).toHaveLength(2);
  expect(requests[1]).toBe(requests[0]);
  await page.getByTestId(networkAnalysisTestId("tab")).click();
  await expect(page.getByRole("tab", { name: /Latest table/u })).toBeVisible();
  await expect(page.getByRole("tab", { name: /Recovered table/u })).toHaveCount(
    0,
  );
});
