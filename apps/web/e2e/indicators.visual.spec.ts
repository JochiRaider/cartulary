import {
  indicatorsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect } from "./fixtures";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import { uniqueTxn } from "./support/runtime/fixtureIdentity";
import {
  assertViewportVisualRegression,
  type VisualAnchor,
} from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import { createRichTimelineSource } from "./support/visual/timelineVisualFixture";
import {
  createLifecycleFixture,
  openLifecycleEditor,
} from "./support/workbook/indicatorLifecycle";
import {
  createObservationFixture,
  observationRawText,
  openObservationEditor,
  selectRepeatedObservation,
} from "./support/workbook/indicatorObservations";

test("Capture Indicator lifecycle UTC authoring at desktop and narrow widths", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const { incidentId, indicator } = await createLifecycleFixture(page);
  const editor = await openLifecycleEditor(
    page,
    incidentId,
    indicator.record_id,
    (url) => navigateVisualApplication(page, url),
  );
  await editor
    .getByLabel("Effective from (UTC)", { exact: true })
    .fill("2026-09-11T12:00");
  await editor.getByLabel("Confidence (optional)", { exact: true }).fill("0");
  await editor
    .getByLabel("Rationale (optional)", { exact: true })
    .fill(
      "Reviewed evidence supports this interval. Earlier assessments remain in History.",
    );
  await editor
    .getByLabel("Assessor text (optional, analyst entered)", { exact: true })
    .fill("Incident analyst");
  await expect(
    editor.getByText("No matching supporting records.", { exact: true }),
  ).toBeVisible();
  const anchor: VisualAnchor = {
    locator: editor,
    align: "start",
    scrollportSelector: `aside[data-view-schema-id="${indicatorsViewSchemaId}"] [data-inspector-scroll-body]`,
  };
  await assertViewportVisualRegression(page, "indicator-lifecycle-authoring", {
    anchor,
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await assertViewportVisualRegression(
    page,
    "indicator-lifecycle-authoring-narrow",
    { anchor },
  );
});

test("Capture Indicator observation source selection at desktop and narrow widths", async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const supplied = await createRichTimelineSource(page, "indicator-source", {
    client_txn_id: uniqueTxn("indicator-source"),
    "timeline.raw_activity_text": observationRawText,
    "timeline.activity_synopsis_text": "Repeated source observation",
  });
  const { incidentId, source } = await createObservationFixture(page, supplied);
  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);
  const editor = await openObservationEditor(page, source.record_id);
  await selectRepeatedObservation(page);
  await expect(
    editor.getByRole("button", { name: "Create observation", exact: true }),
  ).toBeEnabled();
  const anchor: VisualAnchor = {
    locator: editor,
    align: "start",
    scrollportSelector: `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`,
  };
  await assertViewportVisualRegression(
    page,
    "indicator-observation-authoring",
    { anchor },
  );
  await testInfo.attach("indicator-observation-authoring-review", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await assertViewportVisualRegression(
    page,
    "indicator-observation-authoring-narrow",
    { anchor },
  );
  await testInfo.attach("indicator-observation-authoring-narrow-review", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });
});
