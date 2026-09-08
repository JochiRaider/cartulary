import {
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  surfaceTabTestId,
  workbookImportAssistantTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { queryViewRows } from "./support/workbook/query";

test("Workbook Import Assistant discovers, maps, selects, applies, and navigates from the claimed production surface", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("IMPORT-ASSISTANT"),
    "Workbook Import Assistant production path",
  );
  const importedSummary = uniqueTxn("assistant-timeline-row");

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await expect(
    page.getByTestId(surfaceTabTestId(hostsViewSchemaId)),
  ).toHaveAttribute("aria-current", "page");
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
      `Activity Synopsis,Unmapped source note\n${importedSummary},retained raw source\n`,
    ),
  });
  await assistant.getByRole("button", { name: "Upload and discover" }).click();
  await expect(assistant.getByRole("status")).toContainText(
    "Discovered 1 import unit",
  );
  await expect(assistant.getByLabel("Target view")).toHaveValue(
    timelineViewSchemaId,
  );
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
  ).toHaveAttribute("aria-current", "page");
  await expect
    .poll(async () =>
      (await queryViewRows(page, incidentId, timelineViewSchemaId)).some(
        (row) =>
          row.cells["timeline.activity_synopsis_text"]?.value ===
          importedSummary,
      ),
    )
    .toBe(true);
});

async function openAssistant(page: Page) {
  const account = page.getByLabel("Account and application navigation");
  await account.focus();
  await account.press("Enter");
  await page.getByTestId(incidentControlsTriggerTestId()).press("Enter");
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
  await page.route("**/api/v1/jobs/*", async (route) => {
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
  await page.route(
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
  await expect(
    assistant.getByRole("heading", { name: "Outcomes", exact: true }),
  ).toBeVisible();
  await assistant.getByRole("button", { name: "Open Timeline" }).press("Enter");
  await expect(page.getByTestId(incidentControlsPanelTestId())).toHaveCount(0);
  await expect(
    page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
  ).toHaveAttribute("aria-current", "page");
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
  ).toHaveAttribute("aria-current", "page");
});

test("Workbook Import Assistant is absent when the Import profile is unclaimed", async ({
  page,
}) => {
  await page.route("**/api/v1/extensions", async (route) => {
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
