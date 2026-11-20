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
  rowInspectButtonTestId,
  surfaceTabTestId,
  timelineInspectorTestId,
  workbookActiveSurfaceFocusTargetTestId,
  workbookImportAssistantTestId,
  workbookRowContextMenuTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { admitIRCorpus } from "../../../tools/harness/fixtures/ir-incident.mjs";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { authHeadersForStorageState } from "./support/auth/storageState";
import { createIncident } from "./support/incidents/fixtures";
import { seedWorkbookImport } from "./support/incidents/workbookImportSeed";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import {
  fulfillOperationError,
  installOwnedRoute,
} from "./support/transport/requestInterception";
import { queryViewRows } from "./support/workbook/query";
import { openTimelineRowActions } from "./support/workbook/rowMutations";

// TH-HARNESS-REQ-864: an import job's settlement wait has its own bound,
// about twice a normal apply, instead of the whole test budget.
const importSettlementTimeoutMs = 30_000;

type IRCorpus = Awaited<ReturnType<typeof admitIRCorpus>>;

function irSource(corpus: IRCorpus, name: string) {
  const source = corpus.manifest.source_files.find(
    (candidate) => candidate.logical_path === name,
  );
  if (!source) throw new Error(`Missing admitted IR source ${name}`);
  return source;
}

async function queryIRCorpusView(
  page: Page,
  incidentId: string,
  view: string,
): Promise<ViewRow[]> {
  const rows: ViewRow[] = [];
  let body: QueryWorkbookViewRequest = { limit: 200 };
  for (let pageNumber = 0; pageNumber < 20; pageNumber++) {
    const response = await publicHttpOperation({
      operationID: "queryWorkbookView",
      body,
      headers: authHeadersForStorageState(await page.context().storageState()),
      pathParameters: { incident_id: incidentId, view_schema_id: view },
      request: atJsonOrigin(page.request, apiBase),
    });
    if (!response.ok)
      throw new Error(`IR corpus query failed: ${response.status}`);
    rows.push(...response.payload.data.rows);
    const paging = response.payload.meta.paging;
    if (!paging) throw new Error("Missing corpus query paging");
    if (!paging.has_more) {
      expect(new Set(rows.map((row) => row.record_id)).size).toBe(rows.length);
      return rows;
    }
    if (!paging.next_cursor) throw new Error("Missing corpus continuation");
    body = { cursor_token: paging.next_cursor };
  }
  throw new Error("Unbounded corpus query");
}

async function attachIRFixtureAdmission(info: TestInfo, corpus: IRCorpus) {
  await info.attach("ir-fixture-admission", {
    body: JSON.stringify({
      fixture_id: corpus.manifest.fixture_id,
      revision: corpus.manifest.revision,
      manifest_sha256: corpus.manifestSHA256,
    }),
    contentType: "application/json",
  });
}

type IRExpectations = {
  readonly records: Readonly<Record<string, number>>;
  readonly time_pairs: number;
  readonly empty_times: number;
  readonly multiline_raw: number;
  readonly multiline_synopsis: number;
  readonly tabbed_raw: number;
  readonly tabbed_synopsis: number;
  readonly timeline_cases: readonly {
    readonly source_record: number;
    readonly source_identity: string;
    readonly cells: readonly string[];
  }[];
};

const cellText = (row: ViewRow, field: string) =>
  String(row.cells[field]?.value ?? "");

test("IR corpus imports the original Core CSV files through the Import Assistant with exact source text", async ({
  page,
}, info) => {
  const corpus = await admitIRCorpus(info.outputPath("ir-inputs"));
  const expected = JSON.parse(
    corpus.read("ir-incident/v1/expected.json").toString(),
  ) as IRExpectations;
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IR-CORPUS"),
    "Supplied incident documentation",
  );
  const sources = [
    { name: "hosts.csv", view: hostsViewSchemaId, title: "Hosts" },
    {
      name: "identities.csv",
      view: "cartulary.view.identities.v1",
      title: "Identities",
    },
    { name: "timeline.csv", view: timelineViewSchemaId, title: "Timeline" },
  ];
  for (const selected of sources) {
    const source = irSource(corpus, selected.name);
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
      const mapping = assistant.getByRole("combobox", {
        name: header,
        exact: true,
      });
      await mapping.selectOption(field);
      await expect(mapping).toHaveValue(field);
    }
    await assistant
      .getByRole("button", { name: "Approve mapping and select" })
      .click();
    await expect(assistant.getByRole("status")).toContainText("ready to apply");
    await assistant
      .getByRole("button", { name: "Apply 1 selected unit" })
      .click();
    // A normal apply settles in about 15 s; this bound leaves the rest of
    // the test budget to report a stalled job at this step.
    await expect(assistant.getByRole("status")).toContainText(
      "Import completed",
      { timeout: importSettlementTimeoutMs },
    );
    await assistant
      .getByRole("button", { name: "Open " + selected.title, exact: true })
      .click();
    await expect(
      page.getByTestId(surfaceTabTestId(selected.view)),
    ).toHaveAttribute("aria-selected", "true");
    const rows = await queryIRCorpusView(page, incidentId, selected.view);
    expect(rows).toHaveLength(source.records);
    expect(rows).toHaveLength(expected.records[selected.name] ?? -1);
  }

  const timeline = await queryIRCorpusView(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  const columns = irSource(corpus, "timeline.csv").field_keys;
  for (const selected of expected.timeline_cases) {
    const matches = timeline.filter((row) =>
      cellText(row, "timeline.raw_activity_text").includes(
        `evidence_ref="${selected.source_identity}"`,
      ),
    );
    expect(matches, `source record ${selected.source_record}`).toHaveLength(1);
    const row = matches[0];
    if (!row) throw new Error("Missing source record");
    for (const [i, field] of columns.entries())
      expect(cellText(row, field)).toBe(selected.cells[i]);
  }
  // Corpus-wide text shape: exact line breaks, tabs, and paired or empty times.
  const count = (predicate: (row: ViewRow) => boolean) =>
    timeline.filter(predicate).length;
  const raw = (row: ViewRow) => cellText(row, "timeline.raw_activity_text");
  const synopsis = (row: ViewRow) =>
    cellText(row, "timeline.activity_synopsis_text");
  const utc = (row: ViewRow) => cellText(row, "timeline.activity_utc_text");
  const local = (row: ViewRow) => cellText(row, "timeline.activity_local_text");
  expect({
    multiline_raw: count((row) => raw(row).includes("\n")),
    multiline_synopsis: count((row) => synopsis(row).includes("\n")),
    tabbed_raw: count((row) => raw(row).includes("\t")),
    tabbed_synopsis: count((row) => synopsis(row).includes("\t")),
    empty_times: count((row) => utc(row) === "" && local(row) === ""),
    time_pairs: count((row) => utc(row) !== "" && local(row) !== ""),
  }).toEqual({
    multiline_raw: expected.multiline_raw,
    multiline_synopsis: expected.multiline_synopsis,
    tabbed_raw: expected.tabbed_raw,
    tabbed_synopsis: expected.tabbed_synopsis,
    empty_times: expected.empty_times,
    time_pairs: expected.time_pairs,
  });
  const identities = timeline.map(
    (row) => /evidence_ref="([^"]+)"/u.exec(raw(row))?.[1],
  );
  expect(identities.every((identity) => identity !== undefined)).toBe(true);
  expect(new Set(identities).size).toBe(timeline.length);
  await attachIRFixtureAdmission(info, corpus);
});

test("IR corpus source history reopens the exact committed Timeline source for an API-seeded import", async ({
  page,
}, info) => {
  const corpus = await admitIRCorpus(info.outputPath("ir-inputs"));
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IR-SOURCES"),
    "Supplied incident documentation",
  );
  const source = irSource(corpus, "timeline.csv");
  const seeded = await seedWorkbookImport(page, {
    incidentId,
    targetViewSchemaId: timelineViewSchemaId,
    fieldKeys: source.field_keys,
    file: {
      name: "timeline.csv",
      mimeType: "text/csv",
      buffer: corpus.read("ir-incident/v1/timeline.csv"),
    },
    settlementTimeoutMs: importSettlementTimeoutMs,
  });
  const timeline = await queryIRCorpusView(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  expect(timeline).toHaveLength(source.records);
  const firstRecord = timeline[0];
  if (!firstRecord) throw new Error("Missing imported Timeline record");

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await openTimelineRowActions(page, firstRecord.record_id);
  await expect(
    page.getByTestId(
      workbookRowContextMenuTestId(timelineViewSchemaId, firstRecord.record_id),
    ),
  ).toBeVisible();
  const inspect = page.getByTestId(
    rowInspectButtonTestId(firstRecord.record_id),
  );
  await expect(inspect).toBeFocused();
  await inspect.click();
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();

  await page
    .getByRole("button", { name: "Find original import sources" })
    .click();
  const sourceHistory = page.getByRole("region", { name: "Source history" });
  await expect(
    sourceHistory.getByText("Showing sources that produced this record."),
  ).toBeVisible();
  await expect(sourceHistory.getByRole("listitem")).toHaveCount(1);
  await sourceHistory
    .getByRole("button", { name: "timeline.csv", exact: true })
    .click();
  await expect(
    sourceHistory.getByRole("heading", { name: "timeline.csv", exact: true }),
  ).toBeVisible();
  const detail = (label: string) =>
    sourceHistory
      .locator("dl > div")
      .filter({
        has: page.getByRole("term").filter({
          hasText: new RegExp(
            `^${label.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}$`,
          ),
        }),
      })
      .getByRole("definition");
  await expect(detail("Source parser")).toHaveText(
    "cartulary.import.workbook.v2",
  );
  await expect(detail("Applied target")).toHaveText(timelineViewSchemaId);
  await expect(detail("Original import session")).toHaveText(
    seeded.importSessionId,
  );
  await expect(detail("Original import unit")).toHaveText(seeded.importUnitId);
  await expect(detail("Size in bytes")).toHaveText(String(source.size_bytes));
  await expect(detail("SHA-256")).toHaveText(source.sha256);
  await sourceHistory.getByRole("button", { name: "Back to sources" }).click();
  await expect(
    sourceHistory.getByRole("button", { name: "timeline.csv", exact: true }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Close incident controls" }).click();
  // The shell admits one secondary panel at a time. Source inspection replaces
  // the inspector, so its removed invoker returns to the active surface.
  await expect(
    page.getByTestId(workbookActiveSurfaceFocusTargetTestId()),
  ).toBeFocused();
  await attachIRFixtureAdmission(info, corpus);
});

test("IR corpus Cisco Network Flow imports retain the table list, exact source download and accessible source history", async ({
  page,
}, info) => {
  const corpus = await admitIRCorpus(info.outputPath("ir-inputs"));
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IR-FLOWS"),
    "Supplied incident documentation",
  );
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
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
  const names = ["day1", "day2", "overlap"];
  for (const name of names) {
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
    // workflow. Observe its published table within its own bound, below the
    // test budget, before asserting the mapping panel's terminal presentation.
    await expect(
      page.getByRole("tab", { name: new RegExp(`IR ${name}`) }),
    ).toBeVisible({ timeout: importSettlementTimeoutMs });
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
  const sourceHistory = page.getByRole("region", { name: "Source history" });
  const listed = sourceHistory.getByRole("listitem").getByRole("button");
  await expect(listed).toHaveCount(names.length);
  expect((await listed.allTextContents()).sort()).toEqual(
    names.map((name) => `${name}.csv`),
  );
  await attachIRFixtureAdmission(info, corpus);
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
      await fulfillOperationError(route, "getJob", 500, "internal_error");
    } else await route.continue();
  });
  await installOwnedRoute(
    page,
    "**/api/v1/import-sessions/*/units/*/select",
    async (route) => {
      selections++;
      if (selections === 1)
        await fulfillOperationError(
          route,
          "selectImportUnit",
          409,
          "import_apply_blocked",
          { details: { reason_code: "unit_not_ready" } },
        );
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
