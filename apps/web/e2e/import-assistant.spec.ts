import {
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  surfaceTabTestId,
  workbookImportAssistantTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { queryViewRows } from "./support/workbook/query";

const timelineViewSchemaId = "cartulary.view.timeline.v2";
const hostsViewSchemaId = "cartulary.view.hosts.v1";

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
