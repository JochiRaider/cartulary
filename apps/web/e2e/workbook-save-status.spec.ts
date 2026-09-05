import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  saveStateActionButtonTestId,
  saveStateTestId,
  surfaceTabTestId,
  workbookConflictControlTestId,
  workbookConflictResolverTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryTestId,
  workbookInspectorToggleTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, Route } from "@playwright/test";
import { expect, test } from "./fixtures";
import { openSystemSurfaceBySwitcher } from "./pages/workbookInspector";
import {
  editTimelineSummary,
  installPatchController,
  installPatchTransportFailureController,
} from "./support/collaboration/replay";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { safelyRemoveRoute } from "./support/transport/requestInterception";
import { createViewRow, patchRecord } from "./support/workbook/query";
import {
  createSavedView,
  selectSavedView,
} from "./support/workbook/savedViews";
import { observeSaveEvents, saveEvents } from "./support/workbook/saveStatus";

async function selectSurface(page: Page, schema: string) {
  if (schema === assessmentsViewSchemaId)
    await openSystemSurfaceBySwitcher(page, schema);
  else await page.getByTestId(surfaceTabTestId(schema)).click();
  await expect(page.getByTestId(gridShellTestId(schema))).toBeVisible();
}

async function seed(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("SAVESTATUS"),
    "Workbook save-status transitions and recovery",
  );
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("status-row"),
    "timeline.activity_synopsis_text": "Status base",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  return { incidentId, row };
}

test("Preserve save transitions and exact saved-view conflict scope across workbook surfaces.", async ({
  page,
}) => {
  const { incidentId, row } = await seed(page);
  const first = await createSavedView(page, incidentId, {
    display_name: "First status view",
    view_schema_id: timelineViewSchemaId,
  });
  const second = await createSavedView(page, incidentId, {
    display_name: "Second status view",
    view_schema_id: timelineViewSchemaId,
  });
  // Reload only to load authored saved-view fixtures; all subsequent navigation uses production controls.
  await page.reload();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await observeSaveEvents(page);
  await selectSavedView(page, timelineViewSchemaId, first.saved_view_id);
  const patches = await installPatchController(page);
  const held = patches.holdNextPatch({ recordId: row.record_id });
  try {
    await editTimelineSummary(page, row.record_id, "Pending across surfaces", {
      expectValueAfterCommit: false,
    });
    await held.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    for (const schema of [
      notesViewSchemaId,
      hostsViewSchemaId,
      assessmentsViewSchemaId,
      timelineViewSchemaId,
    ]) {
      await selectSurface(page, schema);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    }
    await expect
      .poll(() => saveEvents(page))
      .toEqual([{ priority: "polite", message: "Syncing changes" }]);
    held.release();
    await held.waitForCompletion;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect
      .poll(() => saveEvents(page))
      .toEqual([
        { priority: "polite", message: "Syncing changes" },
        { priority: "polite", message: "Saved" },
      ]);

    await selectSavedView(page, timelineViewSchemaId, first.saved_view_id);
    const conflicted = patches.holdNextPatch({ recordId: row.record_id });
    try {
      await editTimelineSummary(page, row.record_id, "Unsaved conflict", {
        expectValueAfterCommit: false,
      });
      await conflicted.waitForHit;
      await patchRecord(page, row.record_id, {
        view_schema_id: timelineViewSchemaId,
        base_row_version: 2,
        client_txn_id: uniqueTxn("status-remote"),
        changes: [
          {
            field_key: "timeline.activity_synopsis_text",
            value: "Saved remote value",
          },
        ],
      });
      conflicted.release();
      await conflicted.waitForCompletion;
      await expect(
        page.getByTestId(workbookConflictResolverTestId()),
      ).toBeVisible();
      await page.keyboard.press("Escape");
      const strip = page.getByTestId(workbookShellSlotTestId("status-strip"));
      await expect(strip).toContainText("1 same-field conflict needs review.");
      await selectSavedView(page, timelineViewSchemaId, second.saved_view_id);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      await expect(strip).not.toContainText("same-field conflict needs review");
      await selectSurface(page, notesViewSchemaId);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      const action = page.getByTestId(saveStateActionButtonTestId());
      await action.focus();
      await action.press("Enter");
      const resolver = page.getByTestId(workbookConflictResolverTestId());
      await expect(resolver).toBeVisible();
      await expect
        .poll(() =>
          resolver.evaluate((element) =>
            element.contains(document.activeElement),
          ),
        )
        .toBe(true);
      await page.keyboard.press("Escape");
      await expect(action).toBeFocused();
      await action.press("Enter");
      await page
        .getByTestId(workbookConflictControlTestId("keep-saved"))
        .click();
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect
        .poll(() => saveEvents(page))
        .toEqual([
          { priority: "polite", message: "Syncing changes" },
          { priority: "polite", message: "Saved" },
          { priority: "polite", message: "Syncing changes" },
          { priority: "assertive", message: "Conflict. 1 unresolved" },
          { priority: "polite", message: "Syncing changes" },
          { priority: "polite", message: "Saved" },
        ]);
    } finally {
      conflicted.release();
    }
  } finally {
    held.release();
    await held.waitForCompletion;
    await patches.dispose();
  }
});

test("Keep a global FIFO blocker above concurrent local work and activate its existing recovery.", async ({
  page,
}) => {
  const { incidentId, row } = await seed(page);
  await observeSaveEvents(page);
  let releaseCreate = () => {};
  const holdCreate = new Promise<void>((resolve) => {
    releaseCreate = resolve;
  });
  let createHit = false;
  const pattern = `**/api/v1/incidents/${incidentId}/views/${notesViewSchemaId}/rows`;
  const handler = async (route: Route) => {
    createHit = true;
    await holdCreate;
    await route.continue();
  };
  await page.route(pattern, handler);
  const patches = await installPatchController(page);
  try {
    await selectSurface(page, notesViewSchemaId);
    await page
      .getByTestId(workbookInspectorToggleTestId(notesViewSchemaId))
      .click();
    await page
      .getByTestId(genericCreateFieldTestId("note.title"))
      .fill("Concurrent local write");
    await page
      .getByTestId(genericCreateSubmitTestId(notesViewSchemaId))
      .click();
    await expect.poll(() => createHit).toBe(true);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    await selectSurface(page, timelineViewSchemaId);
    patches.failNextPatch(409, "client_txn_conflict", {
      recordId: row.record_id,
    });
    await editTimelineSummary(page, row.record_id, "Blocked queued write", {
      expectValueAfterCommit: false,
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await selectSurface(page, hostsViewSchemaId);
    const action = page.getByTestId(saveStateActionButtonTestId());
    const explanation =
      "A queued edit could not be replayed safely. Retry it with a new request ID, or discard the blocked edit to continue.";
    const eventsBeforeLayout = await saveEvents(page);
    const expectPrimaryGeometry = async () => {
      const primary = page.getByTestId(saveStateTestId());
      await expect(primary).toBeVisible();
      const geometry = await primary.evaluate((element) => {
        const box = element.getBoundingClientRect();
        const strip = element
          .closest('[aria-label="Status strip"]')
          ?.getBoundingClientRect();
        return {
          left: box.left,
          right: box.right,
          top: box.top,
          bottom: box.bottom,
          stripLeft: strip?.left,
          stripRight: strip?.right,
          stripTop: strip?.top,
          stripBottom: strip?.bottom,
        };
      });
      expect(geometry.left).toBeGreaterThanOrEqual(geometry.stripLeft ?? 0);
      expect(geometry.right).toBeLessThanOrEqual(geometry.stripRight ?? 0);
      expect(geometry.top).toBeGreaterThanOrEqual(geometry.stripTop ?? 0);
      expect(geometry.bottom).toBeLessThanOrEqual(geometry.stripBottom ?? 0);
    };
    for (const width of [1440, 1024, 768, 640]) {
      await page.setViewportSize({ width, height: 900 });
      await expectPrimaryGeometry();
      if (width >= 768)
        await expect(action).toHaveAccessibleDescription(
          `Conflict ${explanation}`,
        );
      expect(await saveEvents(page)).toEqual(eventsBeforeLayout);
    }
    await page.setViewportSize({ width: 1600, height: 1440 });
    await page.evaluate(() => {
      document.documentElement.style.zoom = "200%";
    });
    await expectPrimaryGeometry();
    await page.evaluate(() => {
      document.documentElement.style.zoom = "";
    });
    await page.setViewportSize({ width: 768, height: 900 });
    const spacing = await page.addStyleTag({
      content:
        "#root * { letter-spacing: 0.12em !important; line-height: 1.5 !important; word-spacing: 0.16em !important; } #root p { margin-bottom: 2em !important; }",
    });
    await page.emulateMedia({ reducedMotion: "reduce" });
    await expectPrimaryGeometry();
    await expect(action).toHaveAccessibleDescription(`Conflict ${explanation}`);
    expect(await saveEvents(page)).toEqual(eventsBeforeLayout);
    await spacing.evaluate((element) =>
      element.parentNode?.removeChild(element),
    );
    await page.setViewportSize({ width: 1440, height: 900 });
    await action.focus();
    await action.press("Enter");
    await expect(page.getByTestId(workbookEditRecoveryTestId())).toBeFocused();
    await page.getByTestId(workbookEditRecoveryDiscardButtonTestId()).click();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    await expect
      .poll(() => saveEvents(page))
      .toEqual([
        { priority: "polite", message: "Syncing changes" },
        { priority: "assertive", message: "Conflict" },
        { priority: "polite", message: "Syncing changes" },
      ]);
    releaseCreate();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect
      .poll(() => saveEvents(page))
      .toEqual([
        { priority: "polite", message: "Syncing changes" },
        { priority: "assertive", message: "Conflict" },
        { priority: "polite", message: "Syncing changes" },
        { priority: "polite", message: "Saved" },
      ]);
    await selectSurface(page, timelineViewSchemaId);
    patches.failNextPatch(409, "future_terminal_public_error", {
      recordId: row.record_id,
    });
    await editTimelineSummary(page, row.record_id, "Terminal blocked edit", {
      expectValueAfterCommit: false,
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await selectSurface(page, assessmentsViewSchemaId);
    const terminalAction = page.getByTestId(saveStateActionButtonTestId());
    await expect(terminalAction).toHaveAccessibleDescription(
      "Conflict A queued edit could not be completed safely. Discard the blocked edit to continue with later queued edits.",
    );
    await terminalAction.click();
    await expect(page.getByTestId(workbookEditRecoveryTestId())).toBeFocused();
    await page.getByTestId(workbookEditRecoveryDiscardButtonTestId()).click();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  } finally {
    releaseCreate();
    await safelyRemoveRoute(page, pattern, handler);
    await patches.dispose();
  }
});

test("Keep queue overflow globally accessible after real editor admission reaches capacity.", async ({
  page,
}) => {
  const { incidentId, row } = await seed(page);
  // The fixture fills the current 64-unit FIFO and attempts one additional edit.
  const rows = [row];
  for (let index = 1; index <= 64; index += 1) {
    rows.push(
      await createViewRow(page, incidentId, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("overflow-row"),
        "timeline.activity_synopsis_text": `Overflow fixture ${index}`,
      }),
    );
  }
  await page.reload();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await observeSaveEvents(page);
  const patches = await installPatchTransportFailureController(page);
  try {
    patches.disconnect();
    for (const [index, target] of rows.entries()) {
      await editTimelineSummary(
        page,
        target.record_id,
        `Queued edit ${index}`,
        { expectValueAfterCommit: false },
      );
    }
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await selectSurface(page, notesViewSchemaId);
    const action = page.getByTestId(saveStateActionButtonTestId());
    await expect(action).toHaveAccessibleDescription(
      "Conflict The local pending queue is full. Existing queued edits are retained; the current edit remains unsaved local work.",
    );
    await action.click();
    const overflow = page.getByRole("complementary", {
      name: "Workbook queued edit overflow",
    });
    await expect(overflow).toBeFocused();
    await expect(overflow).toContainText(
      "current edit remains unsaved local work",
    );
    await expect
      .poll(() => saveEvents(page))
      .toEqual([
        { priority: "polite", message: "Syncing changes" },
        { priority: "assertive", message: "Conflict" },
      ]);
  } finally {
    patches.connect();
    await patches.dispose();
  }
});
