// biome-ignore lint/correctness/noUndeclaredDependencies: the workspace root pins the browser accessibility engine used by the Make-owned harness.
import AxeBuilder from "@axe-core/playwright";
import type {
  QueryWorkbookViewRequest,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import { networkFlowDecoders } from "@cartulary/protocol-ts/network-flow";
import {
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  networkAnalysisTestId,
  surfaceTabTestId,
  workbookActiveSurfaceFocusTargetTestId,
  workbookImportAssistantTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { admitIRCorpus } from "../../../tools/harness/fixtures/ir-incident.mjs";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { authHeadersForStorageState } from "./support/auth/storageState";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { installOwnedRoute } from "./support/transport/requestInterception";
import { queryViewRows } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test("IR corpus imports original Core and Cisco files and reopens exact source text", async ({
  page,
}, info) => {
  const corpus = await admitIRCorpus(info.outputPath("ir-inputs"));
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IR-CORPUS"),
    "Supplied incident documentation",
  );
  const sources = [
    { name: "hosts.csv", view: "cartulary.view.hosts.v1", title: "Hosts" },
    {
      name: "identities.csv",
      view: "cartulary.view.identities.v1",
      title: "Identities",
    },
    { name: "timeline.csv", view: timelineViewSchemaId, title: "Timeline" },
  ];
  const queryAll = async (view: string): Promise<ViewRow[]> => {
    const rows: ViewRow[] = [];
    let body: QueryWorkbookViewRequest = { limit: 200 };
    for (let pageNumber = 0; pageNumber < 20; pageNumber++) {
      const response = await publicHttpOperation({
        operationID: "queryWorkbookView",
        body,
        headers: authHeadersForStorageState(
          await page.context().storageState(),
        ),
        pathParameters: { incident_id: incidentId, view_schema_id: view },
        request: atJsonOrigin(page.request, apiBase),
      });
      if (!response.ok)
        throw new Error(`IR corpus query failed: ${response.status}`);
      rows.push(...response.payload.data.rows);
      const paging = response.payload.meta.paging;
      if (!paging) throw new Error("Missing corpus query paging");
      if (!paging.has_more) {
        expect(new Set(rows.map((row) => row.record_id)).size).toBe(
          rows.length,
        );
        return rows;
      }
      if (!paging.next_cursor) throw new Error("Missing corpus continuation");
      body = { cursor_token: paging.next_cursor };
    }
    throw new Error("Unbounded corpus query");
  };
  for (const selected of sources) {
    const source = corpus.manifest.source_files.find(
      (source) => source.logical_path === selected.name,
    );
    if (!source) throw new Error("Missing admitted import source");
    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    const assistant = await openAssistant(page);
    await assistant.getByLabel("Source workbook").setInputFiles({
      name: selected.name,
      mimeType: "text/csv",
      buffer: corpus.read("ir-incident/v1/" + selected.name),
    });
    await assistant
      .getByRole("button", { name: "Upload and discover" })
      .click();
    await expect(assistant.getByRole("status")).toContainText(
      "Discovered 1 import unit",
    );
    await assistant.getByLabel("Target view").selectOption(selected.view);
    for (const [i, header] of source.columns.entries()) {
      const field = source.field_keys[i];
      if (!field) throw new Error("Missing admitted source mapping");
      await assistant
        .getByRole("combobox", { name: header, exact: true })
        .selectOption(field);
    }
    await assistant
      .getByRole("button", { name: "Approve mapping and select" })
      .click();
    await expect(assistant.getByRole("status")).toContainText("ready to apply");
    await assistant
      .getByRole("button", { name: "Apply 1 selected unit" })
      .click();
    await expect(assistant.getByRole("status")).toContainText(
      "Import completed",
      { timeout: info.timeout },
    );
    await assistant
      .getByRole("button", { name: "Open " + selected.title, exact: true })
      .click();
    expect(await queryAll(selected.view)).toHaveLength(source.records);
  }
  await page.reload();
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const timeline = await queryAll(timelineViewSchemaId);
  expect(timeline).toHaveLength(1500);
  const expected = JSON.parse(
    corpus.read("ir-incident/v1/expected.json").toString(),
  ) as {
    timeline_cases: {
      source_record: number;
      source_identity: string;
      cells: string[];
    }[];
  };
  const columns = corpus.manifest.source_files.find(
    (source) => source.logical_path === "timeline.csv",
  )?.field_keys;
  if (!columns) throw new Error("Missing Timeline mapping");
  for (const selected of expected.timeline_cases) {
    const matches = timeline.filter((row) =>
      String(row.cells["timeline.raw_activity_text"]?.value ?? "").includes(
        `evidence_ref="${selected.source_identity}"`,
      ),
    );
    expect(matches, `source record ${selected.source_record}`).toHaveLength(1);
    const row = matches[0];
    if (!row) throw new Error("Missing source record");
    for (const [i, field] of columns.entries())
      expect(row.cells[field]?.value ?? "").toBe(selected.cells[i]);
  }

  const firstRecord = timeline[0];
  if (!firstRecord) throw new Error("Missing imported Timeline record");
  await openTimelineInspector(page, firstRecord.record_id);
  const recordSourceTrigger = page.getByRole("button", {
    name: "Find original import sources",
  });
  await recordSourceTrigger.click();
  const sourceHistory = page.getByRole("region", { name: "Source history" });
  await expect(
    sourceHistory.getByText("Showing sources that produced this record."),
  ).toBeVisible();
  await sourceHistory
    .getByRole("button", { name: "timeline.csv", exact: true })
    .click();
  await expect(
    sourceHistory.getByRole("heading", { name: "timeline.csv", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Close incident controls" }).click();
  // The shell admits one secondary panel at a time. Source inspection replaces
  // the inspector, so its removed invoker returns to the active surface.
  await expect(
    page.getByTestId(workbookActiveSurfaceFocusTargetTestId()),
  ).toBeFocused();

  await page.getByRole("button", { name: "More views", exact: true }).click();
  await page.getByTestId(networkAnalysisTestId("tab")).click();
  const mappings = [
    ["firstActiveTime", "flow_start_utc"],
    ["lastActiveTime", "flow_end_utc"],
    ["searchSubject.ipAddress", "input_first_ip"],
    ["peer.ipAddress", "input_second_ip"],
    ["searchSubject.portProtocol.port", "input_first_port"],
    ["peer.portProtocol.port", "input_second_port"],
    ["searchSubject.portProtocol.protocol", "ip_protocol"],
    ["connection.transferBytes", "bytes_count"],
    ["connection.transferPackets", "packets_count"],
    ["searchSubject.orientation", "input_first_orientation"],
    ["peer.orientation", "input_second_orientation"],
  ] as const;
  for (const name of ["day1", "day2", "overlap"]) {
    await page
      .getByTestId(networkAnalysisTestId("import-input"))
      .setInputFiles({
        name: `${name}.csv`,
        mimeType: "text/csv",
        buffer: corpus.read(`network-flow-orientation/v1/${name}.csv`),
      });
    const dialog = page.getByTestId(networkAnalysisTestId("mapping-dialog"));
    await expect(dialog).toBeVisible();
    await dialog
      .getByTestId(networkAnalysisTestId("mapping-display-name"))
      .fill(`IR ${name}`);
    await dialog
      .getByTestId(networkAnalysisTestId("mapping-profile"))
      .selectOption("cisco_sna_flow_search_csv_v2");
    for (const [header, field] of mappings) {
      const escaped = header.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
      const mapping = dialog.getByRole("combobox", {
        name: new RegExp(`^Target for ${escaped} · column [0-9]+$`),
      });
      await mapping.selectOption(`network_flow.${field}`);
      await expect(mapping).toHaveValue(`network_flow.${field}`);
    }
    await dialog.getByTestId(networkAnalysisTestId("mapping-preview")).click();
    await expect(
      dialog.getByTestId(networkAnalysisTestId("mapping-preview-summary")),
    ).toBeVisible();
    await dialog.getByTestId(networkAnalysisTestId("mapping-apply")).click();
    // Approval, selection, apply, and the workspace handoff are one async
    // workflow. Observe its published table within the existing journey budget
    // before asserting the mapping panel's terminal presentation.
    await expect(
      page.getByRole("tab", { name: new RegExp(`IR ${name}`) }),
    ).toBeVisible({ timeout: info.timeout });
    await expect(dialog).toHaveCount(0);
  }
  await page.reload();
  await expect(
    page.getByTestId(networkAnalysisTestId("workspace")),
  ).toBeVisible();
  const tableResponse = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/network-flow/tables`,
  );
  expect(tableResponse.ok()).toBe(true);
  const tableList = networkFlowDecoders.tableList.decode(
    (await tableResponse.json()).data,
  );
  if (!tableList.ok) throw new Error("Invalid retained flow table resource");
  expect(
    tableList.value.tables.map((table) => [
      table.display_name,
      table.row_count_accepted,
    ]),
  ).toEqual([
    ["IR day1", 2000],
    ["IR day2", 2000],
    ["IR overlap", 500],
  ]);
  expect(
    tableList.value.tables.reduce(
      (sum, table) => sum + table.row_count_accepted,
      0,
    ),
  ).toBe(4500);
  await page.getByRole("tab", { name: /IR overlap/ }).click();
  const sourceTrigger = page.getByRole("button", {
    name: "Inspect original source",
  });
  await sourceTrigger.click();
  await expect(
    page.getByRole("heading", { name: "Source history" }),
  ).toBeFocused();
  await expect(
    page.getByRole("heading", { name: "overlap.csv", exact: true }),
  ).toBeVisible();
  await page
    .getByRole("link", { name: "Download original source" })
    .scrollIntoViewIfNeeded();
  const sourceAccessibility = await new AxeBuilder({ page })
    .include('section[aria-labelledby="committed-sources-heading"]')
    .analyze();
  expect(sourceAccessibility.violations).toEqual([]);
  const downloaded = page.waitForEvent("download");
  await page.getByRole("link", { name: "Download original source" }).click();
  const stream = await (await downloaded).createReadStream();
  if (!stream) throw new Error("Source download unavailable");
  const chunks: Buffer[] = [];
  for await (const chunk of stream) chunks.push(Buffer.from(chunk));
  expect(Buffer.concat(chunks)).toEqual(
    corpus.read("network-flow-orientation/v1/overlap.csv"),
  );
  await page.getByRole("button", { name: "Back to sources" }).click();
  await page.getByRole("button", { name: "timeline.csv", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "timeline.csv", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByText("cartulary.import.workbook.v2", { exact: true }),
  ).toBeVisible();
  await info.attach("ir-fixture-admission", {
    body: JSON.stringify({
      fixture_id: corpus.manifest.fixture_id,
      revision: corpus.manifest.revision,
      manifest_sha256: corpus.manifestSHA256,
    }),
    contentType: "application/json",
  });
});

test("Workbook Import Assistant discovers, maps, selects, applies, and navigates from the claimed production surface", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IMPORT-ASSISTANT"),
    "Workbook Import Assistant production path",
  );
  const importedSummary = uniqueTxn("assistant-timeline-row");
  const shortSummary = uniqueTxn("assistant-short-row");

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await expect(
    page.getByTestId(surfaceTabTestId(hostsViewSchemaId)),
  ).toHaveAttribute("aria-selected", "true");
  await page.getByLabel("Account and application navigation").click();
  await page.getByTestId(incidentControlsTriggerTestId()).click();
  await expect(page.getByTestId(incidentControlsMenuTestId())).toBeVisible();
  await page
    .getByTestId(incidentControlsMenuItemTestId("import-assistant"))
    .click();

  const assistant = page.getByTestId(workbookImportAssistantTestId());
  await expect(assistant).toBeVisible();
  await assistant.getByLabel("Source workbook").setInputFiles({
    name: "timeline-import.csv",
    mimeType: "text/csv",
    buffer: Buffer.from(
      `Activity Synopsis,Unmapped source note,Absent source field\n${importedSummary},retained raw source\n${shortSummary}\n`,
    ),
  });
  await assistant.getByRole("button", { name: "Upload and discover" }).click();
  await expect(assistant.getByRole("status")).toContainText(
    "Discovered 1 import unit",
  );
  await expect(assistant.getByLabel("Target view")).toHaveValue(
    timelineViewSchemaId,
  );
  const previewRows = assistant.getByRole("table").locator("tbody tr");
  await expect(assistant.getByRole("columnheader")).toHaveCount(3);
  await expect(previewRows).toHaveCount(2);
  await expect(previewRows.nth(0).getByRole("cell")).toHaveCount(2);
  await expect(previewRows.nth(1).getByRole("cell")).toHaveCount(1);
  await expect(previewRows.nth(1).getByRole("cell")).toHaveText(shortSummary);
  await assistant
    .getByRole("button", { name: "Approve mapping and select" })
    .click();
  await expect(assistant.getByRole("status")).toContainText("ready to apply");
  await assistant
    .getByRole("button", { name: "Apply 1 selected unit" })
    .click();
  await expect(assistant.getByRole("status")).toContainText("Import completed");

  await assistant.getByRole("button", { name: "Open Timeline" }).click();
  await expect(page.getByTestId(incidentControlsPanelTestId())).toHaveCount(0);
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toHaveAttribute("aria-selected", "true");
  await expect
    .poll(async () =>
      (await queryViewRows(page, incidentId, timelineViewSchemaId)).some(
        (row) =>
          row.cells["timeline.activity_synopsis_text"]?.value ===
          importedSummary,
      ),
    )
    .toBe(true);
  await expect
    .poll(async () =>
      (await queryViewRows(page, incidentId, timelineViewSchemaId)).some(
        (row) =>
          row.cells["timeline.activity_synopsis_text"]?.value === shortSummary,
      ),
    )
    .toBe(true);
});

async function openAssistant(page: Page) {
  const account = page.getByLabel("Account and application navigation");
  await account.focus();
  await account.press("Enter");
  await expect(page.getByTestId(incidentControlsTriggerTestId())).toBeVisible();
  await page.getByTestId(incidentControlsTriggerTestId()).press("Enter");
  await expect(page.getByTestId(incidentControlsMenuTestId())).toBeVisible();
  await expect(
    page.getByTestId(incidentControlsMenuItemTestId("import-assistant")),
  ).toBeVisible();
  await page
    .getByTestId(incidentControlsMenuItemTestId("import-assistant"))
    .press("Enter");
  const assistant = page.getByTestId(workbookImportAssistantTestId());
  await expect(assistant).toBeVisible();
  return assistant;
}

test("Workbook Import Assistant contains long mapping content in every workbook density", async ({
  page,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IMPORT-DENSITY"),
    "Workbook Import density",
  );
  const original = (
    await (
      await page.request.get(`${apiBase}/api/v1/account/preferences`)
    ).json()
  ).data;
  const setDensity = async (density: string) => {
    const current = (
      await (
        await page.request.get(`${apiBase}/api/v1/account/preferences`)
      ).json()
    ).data;
    const response = await page.request.put(
      `${apiBase}/api/v1/account/preferences`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_preferences_version: current.preferences_version,
          client_txn_id: uniqueTxn("import-density"),
          density_mode: density,
        },
      },
    );
    expect(response.ok()).toBe(true);
  };
  try {
    for (const density of ["compact", "default", "comfortable"]) {
      await setDensity(density);
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
      const assistant = await openAssistant(page);
      await expect(assistant).toHaveAttribute("data-density", density);
      await assistant.getByLabel("Source workbook").setInputFiles({
        name: "long-columns.csv",
        mimeType: "text/csv",
        buffer: Buffer.from(
          `Activity Synopsis,${"Long header ".repeat(30)}\nDensity preview,${"Long value ".repeat(80)}\n`,
        ),
      });
      await assistant
        .getByRole("button", { name: "Upload and discover" })
        .click();
      await expect(
        assistant.getByRole("button", { name: "Approve mapping and select" }),
      ).toBeEnabled();
      await page.setViewportSize({ width: 390, height: 640 });
      expect(
        await assistant.evaluate(
          (element) => element.scrollWidth <= element.clientWidth + 1,
        ),
      ).toBe(true);
      const approve = assistant.getByRole("button", {
        name: "Approve mapping and select",
      });
      await approve.focus();
      await expect(approve).toBeFocused();
      expect(
        await approve.evaluate(
          (element) => getComputedStyle(element).outlineStyle,
        ),
      ).not.toBe("none");
      await page.screenshot({
        path: testInfo.outputPath(`import-${density}-narrow.png`),
      });
      await page.setViewportSize({ width: 1280, height: 720 });
    }
  } finally {
    await setDensity(original.density_mode);
  }
});

test("Workbook Import Assistant resumes its known job and recovers only selection after drawer closure using the keyboard", async ({
  page,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IMPORT-RECOVERY"),
    "Workbook Import recovery",
  );
  let uploads = 0,
    mappings = 0,
    selections = 0,
    failJob = true;
  const observedJobs: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith("/import-sessions")
    )
      uploads++;
    if (request.method() === "PUT" && request.url().endsWith("/mapping"))
      mappings++;
  });
  await installOwnedRoute(page, "**/api/v1/jobs/*", async (route) => {
    if (route.request().method() !== "GET") {
      await route.continue();
      return;
    }
    observedJobs.push(route.request().url());
    if (failJob) {
      failJob = false;
      await route.fulfill({
        status: 503,
        json: {
          error: { code: "service_unavailable", status: 503, retryable: true },
        },
      });
    } else await route.continue();
  });
  await installOwnedRoute(
    page,
    "**/api/v1/import-sessions/*/units/*/select",
    async (route) => {
      selections++;
      if (selections === 1)
        await route.fulfill({
          status: 409,
          json: {
            error: {
              code: "invalid_import_state",
              status: 409,
              details: { reason_code: "unit_not_ready" },
            },
          },
        });
      else await route.continue();
    },
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  let assistant = await openAssistant(page);
  const file = assistant.getByLabel("Source workbook");
  await file.setInputFiles({
    name: "recovery.csv",
    mimeType: "text/csv",
    buffer: Buffer.from(
      "Activity Synopsis,Long source detail\nRecovery observation," +
        "long source content ".repeat(90) +
        "\n",
    ),
  });
  await file.focus();
  await page.keyboard.press("Tab");
  await expect(
    assistant.getByRole("button", { name: "Upload and discover" }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(assistant.getByRole("alert")).toContainText("Observation:");
  const close = page.getByRole("button", { name: "Close incident controls" });
  await close.focus();
  await page.keyboard.press("Escape");
  await expect(
    page.getByLabel("Account and application navigation"),
  ).toBeFocused();
  assistant = await openAssistant(page);
  await assistant
    .getByRole("button", { name: "Refresh / Resume job" })
    .press("Enter");
  await expect(assistant.getByLabel("Target view")).toHaveValue(
    timelineViewSchemaId,
  );
  expect(uploads).toBe(1);
  expect(new Set(observedJobs).size).toBe(1);
  await assistant
    .getByRole("button", { name: "Approve mapping and select" })
    .press("Enter");
  await expect(assistant.getByRole("alert")).toContainText(
    "Mapping approval is retained",
  );
  await close.focus();
  await page.keyboard.press("Escape");
  assistant = await openAssistant(page);
  await assistant
    .getByRole("button", { name: "Retry selection" })
    .press("Enter");
  await expect(
    assistant.getByRole("button", { name: "Apply 1 selected unit" }),
  ).toBeDisabled();
  await assistant
    .getByRole("button", { name: "Refresh session and outcomes" })
    .press("Enter");
  await expect(
    assistant.getByRole("button", { name: "Apply 1 selected unit" }),
  ).toBeEnabled();
  expect(mappings).toBe(1);
  expect(selections).toBe(2);
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const size of [
    { width: 390, height: 640 },
    { width: 1024, height: 420 },
  ]) {
    await page.setViewportSize(size);
    await expect(page.getByTestId(incidentControlsPanelTestId())).toBeVisible();
    expect(
      await assistant.evaluate(
        (element) => element.scrollWidth <= element.clientWidth + 1,
      ),
    ).toBe(true);
    await close.focus();
    await expect(close).toBeFocused();
  }
  await page.setViewportSize({ width: 1280, height: 900 });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "200%";
  });
  expect(
    await assistant.evaluate(
      (element) => element.scrollWidth <= element.clientWidth + 1,
    ),
  ).toBe(true);
  await testInfo.attach("import-recovery-accessibility-tree", {
    body: await assistant.ariaSnapshot(),
    contentType: "text/plain",
  });
  await page.screenshot({
    path: testInfo.outputPath("import-recovery-zoom.png"),
  });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "";
  });
  await assistant
    .getByRole("button", { name: "Apply 1 selected unit" })
    .press("Enter");
  await expect(assistant.getByRole("status")).toContainText("Import completed");
  await assistant.getByRole("button", { name: "Open Timeline" }).press("Enter");
  await expect(page.getByTestId(incidentControlsPanelTestId())).toHaveCount(0);
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toHaveAttribute("aria-selected", "true");
});

test("Workbook Import Assistant reviews XLSX partial outcomes and navigates only to committed targets", async ({
  page,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IMPORT-PARTIAL"),
    "Workbook Import partial outcomes",
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const assistant = await openAssistant(page);
  await assistant
    .getByLabel("Source workbook")
    .setInputFiles(
      new URL("./testdata/import-assistant-partial.xlsx", import.meta.url)
        .pathname,
    );
  await assistant.getByRole("button", { name: "Upload and discover" }).click();
  await expect(assistant.getByRole("status")).toContainText(
    "Discovered 2 import units",
  );
  const timeline = assistant.getByRole("region", {
    name: "Import unit 1",
    exact: true,
  });
  await timeline
    .getByRole("button", { name: "Approve mapping and select" })
    .click();
  await expect(
    timeline.getByRole("button", {
      name: "Mapping approved and selected",
      exact: true,
    }),
  ).toBeVisible();
  const evidence = assistant.getByRole("region", {
    name: "Import unit 2",
    exact: true,
  });
  await evidence.getByRole("button", { name: "Load unit preview" }).click();
  await evidence
    .getByLabel("Target view")
    .selectOption("cartulary.view.evidence.v1");
  await evidence
    .getByRole("combobox", { name: "Title", exact: true })
    .selectOption("evidence.title");
  await evidence
    .getByRole("combobox", { name: "Requested At", exact: true })
    .selectOption("evidence.requested_at");
  await evidence
    .getByRole("button", { name: "Approve mapping and select" })
    .click();
  await expect(
    assistant.getByRole("button", { name: "Apply 2 selected units" }),
  ).toBeEnabled();
  await assistant
    .getByRole("button", { name: "Apply 2 selected units" })
    .click();
  await expect(
    assistant.getByText(
      "Some units were applied and others failed. Applied work remains committed.",
    ),
  ).toBeVisible();
  await expect(timeline).toContainText("Outcome: applied");
  await expect(evidence).toContainText("Outcome: failed");
  await expect(
    assistant.getByRole("button", { name: "Open Evidence" }),
  ).toHaveCount(0);
  await expect(
    assistant.getByRole("button", { name: "Open Timeline" }),
  ).toBeVisible();
  expect(
    (await queryViewRows(page, incidentId, timelineViewSchemaId)).some(
      (row) =>
        row.cells["timeline.activity_synopsis_text"]?.value ===
        "Workbook import retained observation",
    ),
  ).toBe(true);
  await page.screenshot({
    path: testInfo.outputPath("import-partial-outcomes.png"),
  });
  await assistant.getByRole("button", { name: "Open Timeline" }).click();
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toHaveAttribute("aria-selected", "true");
});

test("Workbook Import Assistant is absent when the Import profile is unclaimed", async ({
  page,
}) => {
  await installOwnedRoute(page, "**/api/v1/extensions", async (route) => {
    const response = await route.fetch();
    const body = (await response.json()) as {
      data: {
        extensions: Array<{ claimed: boolean; profile_id: string }>;
      };
    };
    await route.fulfill({
      response,
      json: {
        ...body,
        data: {
          ...body.data,
          extensions: body.data.extensions.map((extension) =>
            extension.profile_id === "import"
              ? { ...extension, claimed: false }
              : extension,
          ),
        },
      },
    });
  });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IMPORT-BASE"),
    "Workbook Base profile fallback",
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page.getByLabel("Account and application navigation").click();
  await page.getByTestId(incidentControlsTriggerTestId()).click();
  await expect(page.getByTestId(incidentControlsMenuTestId())).toBeVisible();
  await expect(
    page.getByTestId(incidentControlsMenuItemTestId("import-assistant")),
  ).toHaveCount(0);
  await expect(page.getByTestId(workbookImportAssistantTestId())).toHaveCount(
    0,
  );
});
