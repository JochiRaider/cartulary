import { pasteGridMatrix } from "@cartulary/test-utils/grid";
import {
  gridGroupingSelectTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  pendingReplayCountAttribute,
  rowCellTestId,
  rowInspectorFieldTestId,
  saveStateActionButtonTestId,
  saveStateTestId,
  surfaceTabTestId,
  timelineInspectorTestId,
  timelineScalarEditorTestId,
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryRetryButtonTestId,
  workbookEditRecoveryTestId,
  workbookFilterPopoverTriggerTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  handoffViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import {
  activateTimelineGridEditor,
  expectVisibleSemanticGridCellFocus,
} from "./support/accessibility/grid";
import {
  expectNoFocusTrap,
  expectTabOrderIncludes,
  expectVisibleFocus,
} from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { expectWorkbookSavePresentation } from "./support/accessibility/status";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  installPatchController,
  installPatchTransportFailureController,
  successfulPatchCalls,
} from "./support/collaboration/replay";
import { createIncident } from "./support/incidents/fixtures";
import { responseBarrier } from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  awaitResponseBarrier,
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import { expectDecisionControlReachable } from "./support/workbook/decisionSupersession";
import {
  mountedGridCell,
  mountedGridTarget,
  semanticGridCell,
} from "./support/workbook/keyboardGrid";
import {
  fillOrdinaryField,
  openOrdinaryFixture,
  ordinaryField,
  retainOrdinaryUncertainty,
  switchOrdinarySheet,
} from "./support/workbook/ordinaryCreate";
import { expectRecoverySurfaceGeometry } from "./support/workbook/presentationGeometry";
import { createViewRow } from "./support/workbook/query";
import { openRecoveryItem, recoveryEntry } from "./support/workbook/recovery";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test("a11y.mutation-lifecycle.row-01 Verify grid navigation, edit entry/exit, paste feedback, validation feedback, save-state communication, and Esc priority are keyboard and screen-reader safe.", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YMUTATIONLIFECYCLE"),
    "a11y.mutation-lifecycle.row-01 Timeline accessibility",
  );
  const editRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.mutation-lifecycle-01-edit"),
    "timeline.activity_utc_text": "2026-06-03T10:00:00Z",
    "timeline.activity_synopsis_text":
      "browser.mutation-lifecycle edit accessibility row",
    "timeline.raw_activity_text": "Escape priority details",
  });
  const pasteRow = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.mutation-lifecycle-01-paste"),
    "timeline.activity_utc_text": "2026-06-03T10:05:00Z",
    "timeline.activity_synopsis_text":
      "browser.mutation-lifecycle paste accessibility row",
  });
  const pendingRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.mutation-lifecycle-01-pending"),
      "timeline.activity_utc_text": "2026-06-03T10:10:00Z",
      "timeline.activity_synopsis_text":
        "browser.mutation-lifecycle pending accessibility row",
    },
  );
  const validationRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("a11y.mutation-lifecycle-01-validation"),
      "timeline.activity_utc_text": "2026-06-03T10:15:00Z",
      "timeline.activity_synopsis_text":
        "browser.mutation-lifecycle validation accessibility row",
    },
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expectWorkbookSavePresentation(page);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await mountedGridCell(
    page,
    timelineViewSchemaId,
    editRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expectTabOrderIncludes(page, [
    workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
    gridGroupingSelectTestId(timelineViewSchemaId),
    gridShellTestId(timelineViewSchemaId),
  ]);

  const editSummary = await activateTimelineGridEditor(
    page,
    editRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expect(editSummary).toHaveAttribute(
    "aria-label",
    `Activity Synopsis ${editRow.record_id}`,
  );
  await editSummary.fill(
    "browser.mutation-lifecycle accessibility committed edit",
  );
  await editSummary.press("Enter");
  await expect(
    page.getByTestId(
      rowCellTestId(editRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("browser.mutation-lifecycle accessibility committed edit");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

  await pasteGridMatrix({
    fieldKey: "timeline.activity_synopsis_text",
    matrix: [
      [
        "browser.mutation-lifecycle accessibility pasted summary",
        "a11y-host.example",
      ],
    ],
    page,
    recordId: pasteRow.record_id,
    surface: timelineViewSchemaId,
  });
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(
    page.getByTestId(
      rowCellTestId(pasteRow.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText("browser.mutation-lifecycle accessibility pasted summary");

  const patchController = await installPatchTransportFailureController(page);
  {
    patchController.disconnect();
    const pendingSummary = await activateTimelineGridEditor(
      page,
      pendingRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await pendingSummary.fill(
      "browser.mutation-lifecycle accessibility pending replay",
    );
    await pendingSummary.press("Enter");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
    const syncingStatus = await page
      .getByRole("region", {
        name: workbookShellSlotLabel("status-strip"),
        exact: true,
      })
      .ariaSnapshot();
    expect(syncingStatus).toContain("Syncing");
    expect(syncingStatus).not.toContain(`${timelineViewSchemaId}:`);
    expect(syncingStatus).not.toContain("cleared");
    const saveStatus = page.getByTestId(saveStateTestId());
    await expect(saveStatus).toBeVisible();
    await expect(saveStatus).not.toHaveAttribute("role");
    await expect(
      saveStatus.locator('[aria-live], [role="status"]'),
    ).toHaveCount(0);
    await expect(
      page.getByRole("status", { name: "Workbook save updates" }),
    ).toHaveText("Syncing changes");
    await expect(page.getByTestId(saveStateTestId())).toHaveAttribute(
      pendingReplayCountAttribute,
      "1",
    );

    patchController.connect();
    await expect
      .poll(() => successfulPatchCalls(patchController.calls).length)
      .toBe(1);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(page.getByTestId(saveStateTestId())).toHaveAttribute(
      pendingReplayCountAttribute,
      "0",
    );

    await patchController.dispose();
  }

  const originSummary = await mountedGridCell(
    page,
    timelineViewSchemaId,
    editRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await expectVisibleSemanticGridCellFocus(originSummary);
  await openTimelineInspector(page, editRow.record_id);
  const inspectorOriginSummary = await mountedGridCell(
    page,
    timelineViewSchemaId,
    editRow.record_id,
    "timeline.activity_synopsis_text",
  );
  const originSummaryCell = semanticGridCell(inspectorOriginSummary);
  await semanticGridCell(inspectorOriginSummary).focus();
  await expect(page.getByTestId(workbookFocusAnchorTestId())).toHaveText(
    `${timelineViewSchemaId}:${editRow.record_id}:timeline.activity_synopsis_text`,
  );
  const editRaw = page.locator(
    '[data-inspector-edit-field="timeline.raw_activity_text"]',
  );
  await editRaw.click();
  const inspectorDetails = page.getByTestId(
    rowInspectorFieldTestId(editRow.record_id, "timeline.raw_activity_text"),
  );
  await expectVisibleFocus(inspectorDetails);
  await page.keyboard.press("Escape");
  await expect(editRaw).toBeFocused();
  await editRaw.press("Escape");
  await expect(originSummaryCell).toBeFocused();

  const validationCell = await activateTimelineGridEditor(
    page,
    validationRow.record_id,
    "timeline.activity_utc_text",
  );
  await validationCell.fill("not-a-timestamp");
  await validationCell.press("Enter");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(page.getByTestId(saveStateTestId())).toHaveAttribute(
    pendingReplayCountAttribute,
    "0",
  );
  await expect(
    page.getByTestId(
      rowCellTestId(validationRow.record_id, "timeline.activity_utc_text"),
    ),
  ).toHaveText("not-a-timestamp");

  const recoveryController = await installPatchController(page);
  {
    const closeInspector = page.getByTestId(
      workbookInspectorCloseButtonTestId(timelineViewSchemaId),
    );
    if (await closeInspector.isVisible()) {
      await closeInspector.click();
      await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    }
    recoveryController.failNextPatch(409, "client_txn_conflict", {
      recordId: pendingRow.record_id,
    });
    const blockedSummary = await activateTimelineGridEditor(
      page,
      pendingRow.record_id,
      "timeline.activity_synopsis_text",
    );
    await blockedSummary.fill(
      "browser.mutation-lifecycle accessibility blocked edit",
    );
    await blockedSummary.press("Enter");
    await expect.poll(() => recoveryController.calls.length).toBe(1);

    await expect(page.getByTestId(workbookEditRecoveryTestId())).toHaveCount(0);
    const statusAction = page.getByTestId(saveStateActionButtonTestId());
    const recoveryMessage =
      "A queued edit could not be replayed safely. Retry it with a new request ID, or discard the blocked edit to continue.";
    await expect(statusAction).toHaveAccessibleName("Open conflict recovery");
    await expect(statusAction).toHaveAccessibleDescription(
      `Conflict ${recoveryMessage}`,
    );
    await expectVisibleFocus(statusAction);
    await statusAction.press("Enter");
    const recoveryPanel = page.getByTestId(workbookEditRecoveryTestId());
    const retryButton = page.getByTestId(
      workbookEditRecoveryRetryButtonTestId(),
    );
    const discardButton = page.getByTestId(
      workbookEditRecoveryDiscardButtonTestId(),
    );
    await expect(recoveryPanel).toHaveRole("region");
    await expect(recoveryPanel).toHaveAccessibleName("Workbook edit recovery");
    await expect(recoveryPanel).not.toBeFocused();
    await expect(retryButton).toHaveAccessibleName(
      "Retry with a new request ID",
    );
    await expect(discardButton).toHaveAccessibleName("Discard blocked edit");
    await expect(retryButton).toBeEnabled();
    await expect(discardButton).toBeEnabled();
    await expect(
      recoveryPanel.getByRole("status", {
        name: "Queued edit recovery message",
      }),
    ).toHaveAttribute("aria-live", "polite");
    await expect(
      recoveryPanel.getByRole("status", {
        name: "Queued edit recovery message",
      }),
    ).toHaveAttribute("aria-atomic", "true");
    await expect(recoveryPanel.getByRole("status")).toHaveCount(1);
    await expect(recoveryPanel).toContainText("Queued edits");
    await expect(recoveryPanel).toContainText(
      "A queued edit could not be replayed safely. Retry it with a new request ID, or discard the blocked edit to continue.",
    );
    await expect(recoveryPanel).not.toContainText("client_txn_conflict");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");

    for (const viewport of [
      { width: 1440, height: 900 },
      { width: 1024, height: 720 },
      { width: 768, height: 640 },
    ]) {
      await page.setViewportSize(viewport);
      await expectRecoverySurfaceGeometry(page);
      const accessibleStatus = await page
        .getByRole("region", {
          name: workbookShellSlotLabel("status-strip"),
          exact: true,
        })
        .ariaSnapshot();
      expect(accessibleStatus).toContain("Conflict");
      expect(accessibleStatus).toContain(recoveryMessage);
      expect(accessibleStatus).not.toContain(`${timelineViewSchemaId}:`);
      expect(accessibleStatus).not.toContain("cleared");
      await expect(statusAction).toHaveAccessibleDescription(
        new RegExp(recoveryMessage.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")),
      );
    }
    await page.setViewportSize({ width: 1440, height: 900 });
    await withPresentation(page, { zoom: "200%" }, async () => {
      await expectRecoverySurfaceGeometry(page, {
        enforceDocumentBlockExtent: false,
      });
    });
    await page.setViewportSize({ width: 768, height: 640 });
    await withPresentation(
      page,
      {
        textSpacing: `
          #root * {
            letter-spacing: 0.12em !important;
            line-height: 1.5 !important;
            word-spacing: 0.16em !important;
          }
          #root p {
            margin-bottom: 2em !important;
          }
        `,
      },
      async () => {
        await expectRecoverySurfaceGeometry(page);
      },
    );
    await page.setViewportSize({ width: 1440, height: 900 });

    await page.getByTestId(saveStateActionButtonTestId()).click();
    await expect(
      page
        .getByRole("region", { name: "Recovery navigation", exact: true })
        .locator(":scope > h2"),
    ).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(retryButton).toBeFocused();
    await page.keyboard.press("Tab");
    await expect(discardButton).toBeFocused();

    await expect(
      await mountedGridTarget(
        page,
        timelineViewSchemaId,
        gridSortHeaderTestId(
          timelineViewSchemaId,
          "timeline.activity_synopsis_text",
        ),
      ),
    ).toContainText("Activity Synopsis");
    await expectAllInteractiveControlsNamed(page);
    await expectNoFocusTrap(page);
    await expectAndRecordContrast(page, [
      workbookShellSlotTestId("status-strip"),
      workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
      gridGroupingSelectTestId(timelineViewSchemaId),
      gridSortHeaderTestId(
        timelineViewSchemaId,
        "timeline.activity_synopsis_text",
      ),
      rowCellTestId(editRow.record_id, "timeline.activity_synopsis_text"),
      rowCellTestId(validationRow.record_id, "timeline.activity_utc_text"),
      workbookEditRecoveryTestId(),
      workbookEditRecoveryRetryButtonTestId(),
      workbookEditRecoveryDiscardButtonTestId(),
      saveStateTestId(),
    ]);

    await discardButton.focus();
    await expect(discardButton).toBeFocused();
    await page.keyboard.press("Space");
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(recoveryPanel).toHaveCount(0);
    await expect(
      page
        .getByRole("region", { name: "Recovery navigation", exact: true })
        .locator(":scope > h2"),
    ).toBeFocused();
    expect(recoveryController.calls).toHaveLength(1);

    await recoveryController.dispose();
  }
});

test("a11y.ordinary grid references and retained recovery support keyboard focus and narrow layouts", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const preferences = async () =>
    (
      await (
        await page.request.get(`${apiBase}/api/v1/account/preferences`)
      ).json()
    ).data;
  const priorPreferences = await preferences();
  const selectDensity = async (density: string | null) => {
    const current = await preferences();
    const response = await page.request.put(
      `${apiBase}/api/v1/account/preferences`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_preferences_version: current.preferences_version,
          client_txn_id: uniqueTxn("ordinary-recovery-density"),
          density_mode: density,
        },
      },
    );
    expect(response.ok()).toBe(true);
  };
  try {
    await selectDensity("default");
    const handoff = handoffViewSchemaId;
    const { incident } = await openOrdinaryFixture(page, handoff);
    const referenceInput = await ordinaryField(
      page,
      handoff,
      "handoff.incoming_owner_user_id",
    );
    const reference = page
      .getByRole("gridcell")
      .filter({ has: referenceInput });
    const choose = reference.getByRole("button", {
      name: "Choose incoming owner",
      exact: true,
    });
    await expectDecisionControlReachable(page, choose);
    await choose.press("Enter");
    const picker = reference.getByRole("group", {
      name: "Incoming Owner",
      exact: true,
    });
    const candidate = picker.getByRole("radio").last();
    await expect(candidate).toBeEnabled();
    await expect(
      reference.getByRole("button", { name: "Cancel references", exact: true }),
    ).toBeFocused();
    await page.keyboard.press("Shift+Tab");
    await expect(
      reference.getByRole("button", { name: "Apply references", exact: true }),
    ).toBeFocused();
    await page.keyboard.press("Shift+Tab");
    await expect(
      reference.getByRole("button", {
        name: "Refresh candidates",
        exact: true,
      }),
    ).toBeFocused();
    await page.keyboard.press("Shift+Tab");
    await expect(
      reference.getByRole("button", { name: "First candidates", exact: true }),
    ).toBeFocused();
    await page.keyboard.press("Shift+Tab");
    await expect(candidate).toBeFocused();
    await expectVisibleFocus(candidate);
    const originalViewport = page.viewportSize();
    if (!originalViewport) throw new Error("Missing browser viewport");
    await page.setViewportSize({ width: originalViewport.width, height: 480 });
    await expect(candidate).toBeFocused();
    await expect(candidate).toBeInViewport({ ratio: 1 });
    await expect(
      reference.getByRole("button", { name: "Cancel references", exact: true }),
    ).toBeInViewport({ ratio: 1 });
    await page.setViewportSize(originalViewport);
    await picker.press("Escape");
    await expect(choose).toBeFocused();
    await expect(picker).toHaveCount(0);
    await switchOrdinarySheet(page, evidenceViewSchemaId);
    await fillOrdinaryField(
      page,
      evidenceViewSchemaId,
      "evidence.title",
      "Accessible retained draft",
    );
    const { recovery } = await retainOrdinaryUncertainty(page, incident);
    const recover = recovery.getByRole("button", {
      name: "Recover submission",
      exact: true,
    });
    const account = page.getByRole("button", {
      name: "Account and application navigation",
      exact: true,
    });
    for (const viewport of [
      { width: 1280, height: 720 },
      { width: 390, height: 480 },
    ]) {
      await page.setViewportSize(viewport);
      await expectDecisionControlReachable(page, recover);
      await expectVisibleFocus(recover);
      await expectAllInteractiveControlsNamed(page);
      await testInfo.attach(`ordinary-recovery-${viewport.width}`, {
        body: await page.screenshot({ animations: "disabled", caret: "hide" }),
        contentType: "image/png",
      });
    }
    await testInfo.attach("ordinary-recovery-tree", {
      body: await recovery.ariaSnapshot(),
      contentType: "text/plain",
    });
    // Account navigation and footer navigation remain reachable in their own bands.
    await expect(account).toBeInViewport({ ratio: 1 });
    await recover.press("Enter");
    await expect(recovery).toContainText("Row accepted.");
    await expect(account).toBeInViewport({ ratio: 1 });
    const footer = page.getByTestId(workbookShellSlotTestId("status-strip"));
    const systemViews = footer.getByRole("button", {
      name: "More views",
      exact: true,
    });
    // The toolbar owns focus reveal; do not scroll the control from the test.
    await systemViews.focus();
    await expect(systemViews).toBeFocused();
    await expectVisibleFocus(systemViews);
    await expect(systemViews).toBeInViewport({ ratio: 1 });
    await expect(account).toBeInViewport({ ratio: 1 });
    const systemBounds = await systemViews.boundingBox();
    const accountBounds = await account.boundingBox();
    if (!systemBounds || !accountBounds)
      throw new Error("Workbook chrome is missing");
    expect(systemBounds.y).toBeGreaterThanOrEqual(
      accountBounds.y + accountBounds.height,
    );
    const navigation = footer.getByRole("button", {
      name: "Navigation",
      exact: true,
    });
    await expectDecisionControlReachable(page, navigation);
    await expectVisibleFocus(navigation);
    await expect(footer.getByTestId(saveStateTestId())).toBeInViewport({
      ratio: 1,
    });
    await navigation.press("Enter");
    await expect(
      page.getByRole("dialog", { name: "Navigation details" }),
    ).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(navigation).toBeFocused();
    await testInfo.attach("ordinary-toolbar-focus-390", {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
    await expectDecisionControlReachable(page, account);
    await account.press("Enter");
    await expect(
      page.getByRole("menuitem", { name: "Incidents", exact: true }),
    ).toBeFocused();
    await page.keyboard.press("Escape");
    await expect(account).toBeFocused();
  } finally {
    await selectDensity(priorPreferences.density_mode);
  }
});

test("a11y.workbook-batch retained paste retry remains reachable across narrow layouts", async ({
  page,
}, testInfo) => {
  await page.emulateMedia({ reducedMotion: "reduce" });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YBATCH"),
    "Workbook batch recovery",
  );
  const field = "timeline.activity_synopsis_text";
  const created = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("batch-a11y-seed"),
    [field]: "Batch anchor",
  });
  await page.goto(`/?incident_id=${incidentId}`);
  await openTimelineInspector(page, created.record_id);
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.note",
      ),
    )
    .click();
  await page
    .getByRole("region", { name: "Create Note", exact: true })
    .getByRole("textbox", { name: "Title", exact: true })
    .fill("  Retained Note alongside an uncertain batch  ");
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await expect(recoveryEntry(page)).toHaveText("Recovery (1)");
  const anchor = await mountedGridCell(
    page,
    timelineViewSchemaId,
    created.record_id,
    field,
  );
  await anchor.click();
  await page.keyboard.press("Escape");
  const path = `**/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`;
  let fail = true;
  const attempts: string[] = [];
  const replayGateBarrier = responseBarrier(page);
  const releaseReplay = replayGateBarrier.release;
  const replayGate = replayGateBarrier.promise;
  await installOwnedRoute(page, path, async (route) => {
    attempts.push(route.request().postData() ?? "");
    const response = await route.fetch();
    if (fail) await route.abort("connectionfailed");
    else {
      if (!(await awaitResponseBarrier(route, replayGate))) return;
      await route.fulfill({ response });
    }
  });
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
  await page.evaluate(() =>
    navigator.clipboard.writeText("Retained update\nRetained create"),
  );
  await page.keyboard.press("Control+v");
  const trigger = recoveryEntry(page);
  await expect(trigger).toHaveText("Recovery (2)");
  await expectDecisionControlReachable(page, trigger);
  await openRecoveryItem(page, /^Paste ·/);
  const recovery = page.getByRole("region", {
    name: "Workbook recovery",
    exact: true,
  });
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
  await openRecoveryItem(page, /^Note draft ·/);
  await page
    .getByRole("button", { name: "Resume Note draft", exact: true })
    .click();
  await expect(
    page
      .getByRole("region", { name: "Retained Note authoring", exact: true })
      .getByRole("textbox", { name: "Title", exact: true }),
  ).toHaveValue("  Retained Note alongside an uncertain batch  ");
  await openRecoveryItem(page, /^Paste ·/);
  const retry = recovery.getByRole("button", {
    name: "Retry paste",
    exact: true,
  });
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 768, height: 640 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    await expectDecisionControlReachable(page, retry);
    await expectVisibleFocus(retry);
    await expectAllInteractiveControlsNamed(page);
    await expect(
      recovery.getByRole("textbox", { name: "Original batch input" }),
    ).toHaveValue("Retained update\nRetained create");
    await testInfo.attach(`batch-recovery-${viewport.width}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectDecisionControlReachable(page, retry);
    await expectVisibleFocus(retry);
  });
  await page.setViewportSize({ width: 768, height: 640 });
  await withPresentation(
    page,
    {
      textSpacing: `
      #root * {
        letter-spacing: 0.12em !important;
        line-height: 1.5 !important;
        word-spacing: 0.16em !important;
      }
      #root p { margin-bottom: 2em !important; }
    `,
    },
    async () => {
      await expectDecisionControlReachable(page, retry);
      await expectVisibleFocus(retry);
      await expectDecisionControlReachable(
        page,
        recovery.getByRole("button", { name: "Close recovery", exact: true }),
      );
    },
  );
  await expectDecisionControlReachable(page, retry);
  await retry.press("Escape");
  await expect(trigger).toBeFocused();
  await openRecoveryItem(page, /^Paste ·/);
  await expectDecisionControlReachable(page, retry);
  fail = false;
  await retry.press("Enter");
  try {
    await expect.poll(() => attempts.length).toBe(2);
    await expect(retry).toBeFocused();
    await expect(retry).toHaveAttribute("aria-busy", "true");
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    for (const profile of [
      { width: 390, height: 480, zoom: "", spacing: false },
      { width: 1280, height: 720, zoom: "200%", spacing: false },
      { width: 768, height: 640, zoom: "", spacing: true },
    ]) {
      await page.setViewportSize({
        width: profile.width,
        height: profile.height,
      });
      await withPresentation(page, { zoom: profile.zoom }, async () => {
        await withPresentation(
          page,
          profile.spacing
            ? {
                textSpacing:
                  "#root * { letter-spacing: 0.12em !important; line-height: 1.5 !important; word-spacing: 0.16em !important; }",
              }
            : {},
          async () => {
            await expect(retry).toBeVisible();
            await expect(retry).toBeFocused();
            await expect(retry).toHaveAttribute("aria-busy", "true");
            await expect(
              recovery.getByRole("textbox", { name: "Original batch input" }),
            ).toHaveValue("Retained update\nRetained create");
            await testInfo.attach(
              `batch-recovery-pending-${profile.width}${profile.spacing ? "-spacing" : profile.zoom ? "-zoom" : ""}`,
              {
                body: await page.screenshot({
                  animations: "disabled",
                  caret: "hide",
                }),
                contentType: "image/png",
              },
            );
          },
        );
      });
    }

    releaseReplay();
  } finally {
    replayGateBarrier.cancel();
  }
  await expect(recovery).toContainText("Saved changes.");
  expect(attempts).toHaveLength(2);
  expect(attempts[0]).toBe(attempts[1]);
  await testInfo.attach("batch-recovery-tree", {
    body: await recovery.ariaSnapshot(),
    contentType: "text/plain",
  });
  await recovery
    .getByRole("button", { name: "Close recovery", exact: true })
    .click();
  await page.setViewportSize({ width: 1440, height: 900 });
  const additional = [];
  for (let index = 0; index < 8; index++) {
    additional.push(
      await createViewRow(page, incidentId, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("many-recovery-seed"),
        [field]: `Recovery target ${index + 1}`,
      }),
    );
  }
  await page.getByTestId(surfaceTabTestId(notesViewSchemaId)).click();
  await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
  fail = true;
  for (const [index, row] of additional.entries()) {
    const target = await mountedGridCell(
      page,
      timelineViewSchemaId,
      row.record_id,
      field,
    );
    await target.click();
    await expect(
      page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: field,
          recordId: row.record_id,
          surface: "grid",
        }),
      ),
    ).toBeFocused();
    await page.keyboard.press("Escape");
    await page.evaluate(
      (value) => navigator.clipboard.writeText(value),
      `Batch sample ${index + 1} with a long retained description to distinguish this recovery obligation\tBatch source ${index + 1}`,
    );
    await page.keyboard.press("Control+v");
    try {
      await expect(trigger).toHaveText(`Recovery (${index + 2})`);
    } catch (error) {
      await testInfo.attach("many-recovery-failure-tree", {
        body: await page.locator("#root").ariaSnapshot(),
        contentType: "text/plain",
      });
      throw error;
    }
  }
  await trigger.click();
  const list = page.getByRole("region", {
    name: "Recovery navigation",
    exact: true,
  });
  await expect(list.getByRole("button", { name: /^Paste ·/ })).toHaveCount(8);
  for (const viewport of [
    { width: 1280, height: 720 },
    { width: 390, height: 480 },
  ]) {
    await page.setViewportSize(viewport);
    const last = list
      .getByRole("listitem")
      .filter({ hasText: "Batch sample 8" })
      .getByRole("button", { name: /^Paste ·/ });
    await expect(last).toHaveAccessibleDescription(/Batch sample 8/);
    await expectDecisionControlReachable(page, last);
    await expectVisibleFocus(last);
    await testInfo.attach(`many-recovery-${viewport.width}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
  }
  await list
    .getByRole("listitem")
    .filter({ hasText: "Batch sample 8" })
    .getByRole("button", { name: /^Paste ·/ })
    .press("Enter");
  await expect(
    recovery.getByRole("textbox", { name: "Original batch input" }),
  ).toHaveValue(/Batch sample 8/);
  await recovery
    .getByRole("button", { name: "Close recovery", exact: true })
    .press("Escape");
  await expect(trigger).toBeFocused();
  await removeOwnedRoute(page, path);
});
