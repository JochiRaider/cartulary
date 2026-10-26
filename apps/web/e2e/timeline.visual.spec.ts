import {
  gridRowTestId,
  gridRowVersionAttribute,
  saveStateTestId,
  timelineCaptureActionTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { expect } from "./fixtures";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  assertViewportVisualRegression,
  type VisualAnchor,
} from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import {
  createRichTimelineIncident,
  createRichTimelineSource,
  createVisualTimelineRow,
  visualTimelineRecord,
} from "./support/visual/timelineVisualFixture";
import { normalizeWorkbookGridVisualState } from "./support/visual/workbookGeometry";
import { mountedGridCell } from "./support/workbook/keyboardGrid";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
import { openTimelineSupersessionFixture } from "./support/workbook/timelineCaptureActions";

test("captures the Timeline default viewport with stable row version and save-state strip", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALTIMELINEDEFAULT"),
    "Timeline mutation visual default",
  );
  const timelineRow = visualTimelineRecord(
    incidentId,
    "authentication-anomaly",
  );

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  const summaryCell = await mountedGridCell(
    page,
    timelineViewSchemaId,
    timelineRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expect(
    page.getByTestId(
      gridRowTestId(timelineViewSchemaId, timelineRow.record_id),
    ),
  ).toHaveAttribute(gridRowVersionAttribute, String(timelineRow.row_version));
  await expect(summaryCell).toHaveText(
    String(timelineRow.cells["timeline.activity_synopsis_text"]?.value),
  );
  await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
    scroll: { top: 0, left: "left" },
  });

  await assertViewportVisualRegression(page, "timeline-grid-timeline-default", {
    renderSurface: timelineViewSchemaId,
  });
});

test("Capture Timeline supersession authoring review and accepted replacement", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const supplied = await createRichTimelineSource(page, "supersession-target", {
    client_txn_id: uniqueTxn("supersession-target"),
    "timeline.activity_synopsis_text": "Investigated observation",
    "timeline.device_object_text": "Workstation A",
    "timeline.activity_utc_text": "2025-02-17T11:00:00Z",
  });
  const replacementRow = await createVisualTimelineRow(
    page,
    supplied.incidentId,
    "supersession-replacement",
    {
      client_txn_id: uniqueTxn("supersession-replacement"),
      "timeline.activity_synopsis_text": "Investigated observation",
      "timeline.device_object_text": "Workstation B",
      "timeline.activity_utc_text": "2025-02-17T11:05:00Z",
    },
  );
  const { target, replacement } = await openTimelineSupersessionFixture(
    page,
    (url) => navigateVisualApplication(page, url),
    { ...supplied, replacement: replacementRow },
  );
  const editor = page.getByTestId(
    timelineCaptureActionTestId("editor", target.record_id),
  );
  const anchor: VisualAnchor = {
    locator: editor,
    align: "start",
    scrollportSelector: `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`,
  };
  await assertViewportVisualRegression(page, "timeline-supersession-review", {
    anchor,
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await assertViewportVisualRegression(
    page,
    "timeline-supersession-review-narrow",
    { anchor },
  );
  await page.setViewportSize({ width: 1280, height: 720 });
  await editor.getByRole("button", { name: "Back", exact: true }).click();
  await page
    .getByTestId(timelineCaptureActionTestId("replacement", target.record_id))
    .selectOption("");
  await assertViewportVisualRegression(
    page,
    "timeline-supersession-authoring",
    { anchor },
  );
  await page
    .getByTestId(timelineCaptureActionTestId("replacement", target.record_id))
    .selectOption(replacement.record_id);
  await page
    .getByTestId(timelineCaptureActionTestId("review", target.record_id))
    .click();
  await page
    .getByTestId(timelineCaptureActionTestId("confirm", target.record_id))
    .click();
  await expect(recoveryEntry(page)).toHaveText("Recovery (0)");
  await openRecoveryItem(page, /^Timeline action ·/);
  await expect(
    page.getByTestId(timelineCaptureActionTestId("result", target.record_id)),
  ).toHaveText("Timeline supersession completed.");
  await assertViewportVisualRegression(page, "timeline-supersession-accepted");
});
