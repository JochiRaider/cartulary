import { timelineViewSchemaId } from "@cartulary/view-contracts";
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
  openNoteFixture,
  retainNoteUncertainResult,
} from "./support/workbook/noteCreate";

test("Capture linked Note authoring source selection and retained atomic recovery", async ({
  page,
}) => {
  const capture = async (
    name: string,
    options: { anchor?: VisualAnchor } = {},
  ) => {
    await assertViewportVisualRegression(page, name, options);
    await test.info().attach(`${name}-review`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  };
  await page.setViewportSize({ width: 1280, height: 720 });
  const supplied = await createRichTimelineSource(page, "note-source", {
    client_txn_id: uniqueTxn("note-source"),
    "timeline.activity_synopsis_text": "Reviewed investigation source",
  });
  const f = await openNoteFixture(
    page,
    timelineViewSchemaId,
    (url) => navigateVisualApplication(page, url),
    supplied,
  );
  await f.form
    .getByRole("textbox", { name: "Title", exact: true })
    .fill("Retained investigation Note");
  await f.form
    .getByRole("textbox", { name: "Body", exact: true })
    .fill("Unfinished analysis remains attached to its chosen source.");
  const anchor: VisualAnchor = {
    locator: f.form,
    align: "start",
    scrollportSelector: `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`,
  };
  await capture("linked-note-authoring", { anchor });
  await page.setViewportSize({ width: 768, height: 640 });
  await capture("linked-note-authoring-narrow", { anchor });
  await f.form
    .getByRole("button", { name: "Choose source", exact: true })
    .click();
  const picker = f.form.getByRole("region", {
    name: "Choose Note source",
    exact: true,
  });
  await expect(
    picker.getByRole("button", { name: "Apply source", exact: true }),
  ).toBeEnabled();
  await capture("linked-note-source-narrow", {
    anchor: { ...anchor, locator: picker },
  });
  await picker
    .getByRole("button", { name: "Cancel source", exact: true })
    .click();
  const { recovery } = await retainNoteUncertainResult(
    page,
    f.source.record_id,
    f.view,
  );
  await capture("linked-note-recovery-narrow");
  await page.setViewportSize({ width: 1280, height: 720 });
  await capture("linked-note-recovery");
  await test.info().attach("linked-note-recovery-tree", {
    body: await recovery.ariaSnapshot(),
    contentType: "text/plain",
  });
});
