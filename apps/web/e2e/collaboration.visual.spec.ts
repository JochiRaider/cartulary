import {
  assertMarkerAnchoredToGridTarget,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  cellPresenceMarkerTestId,
  gridRowGutterTestId,
  gridRowTestId,
  gridShellTestId,
  pendingReplayCountAttribute,
  rowCellTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
  timelineScalarEditorTestId,
  workbookConflictControlTestId,
  workbookConflictResolverTestId,
  workbookPresenceSummaryTestId,
  workbookShellReadyTestId,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect } from "./fixtures";
import { withPresentation } from "./support/accessibility/presentation";
import { selectVisualDensity } from "./support/auth/visualPreferences";
import {
  driveRealTimelineSummaryConflict,
  focusRemoteTimelineCellAndWaitForPresence,
  installPatchController,
  openIncidentAsTrackedUserReady,
} from "./support/collaboration/replay";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import {
  navigateVisualApplication,
  reloadVisualApplication,
} from "./support/runtime/applicationReadiness";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import {
  assertViewportVisualRegression,
  assertVisualRegression,
  prepareVisualPresentation,
} from "./support/visual/capture";
import { initializeVisualPage, test } from "./support/visual/fixtures";
import {
  createRichTimelineIncident,
  createVisualTimelineRow,
} from "./support/visual/timelineVisualFixture";
import {
  assertStatusStripVisualFixture,
  assertStatusStripVisualRegression,
  stabilizeConflictResolverVisual,
} from "./support/visual/workbookCaptures";
import {
  normalizeWorkbookGridVisualState,
  readWorkbookGridScroll,
} from "./support/visual/workbookGeometry";
import { mountedGridCell } from "./support/workbook/keyboardGrid";

async function prepareFeP7ConflictVisual(
  page: Page,
  options: {
    incidentKeyPrefix: string;
    title: string;
  },
) {
  await page.setViewportSize({ width: 1280, height: 720 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey(options.incidentKeyPrefix),
    options.title,
  );
  const conflictRow = await createVisualTimelineRow(
    page,
    incidentId,
    "same-field-conflict",
    {
      client_txn_id: uniqueTxn(`${options.incidentKeyPrefix}-CONFLICT`),
      "timeline.activity_utc_text": "2025-03-07T10:00:00Z",
      "timeline.activity_synopsis_text": "Conflict visual base",
    },
  );
  const patchController = await installPatchController(page);

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await driveRealTimelineSummaryConflict({
    baseRowVersion: conflictRow.row_version,
    localValue: "Conflict visual local",
    page,
    patchController,
    recordId: conflictRow.record_id,
    remoteValue: "Conflict visual server",
    txnPrefix: `${options.incidentKeyPrefix.toLowerCase()}-conflict`,
  });
  const originalEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: conflictRow.record_id,
      surface: "grid",
    }),
  );
  await expect(
    page
      .getByRole("region", { name: "Recovery navigation", exact: true })
      .locator(":scope > h2"),
  ).toBeFocused();
  expect(
    await originalEditor.evaluate(
      (element) => element.closest('[inert], [aria-hidden="true"]') === null,
    ),
  ).toBe(true);
  return { conflictRow, patchController };
}

test("Capture deterministic presence at the workbook header, row gutter, and cell with exact overflow counts.", async ({
  browserScope,
  page,
  sessionTracker,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALCOLLABORATION"),
    "browser.collaboration visual collaboration states",
  );
  const remoteActors = await Promise.all(
    (
      [
        ["Alpha Analyst", "AA"],
        ["Bravo Analyst", "BA"],
        ["Charlie Analyst", "CA"],
        ["Delta Analyst", "DA"],
        ["Echo Analyst", "EA"],
        ["Foxtrot Analyst", "FA"],
      ] as const
    ).map(async ([displayName, actorText], index) => ({
      actorText,
      user: await createIncidentMemberUser(page, incidentId, {
        display_name: displayName,
        email: uniqueEmail(`collaboration-remote-${index + 1}`),
        initial_password: "FeVP7RemotePass!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      }),
    })),
  );
  const presenceRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualcollaboration-presence",
    {
      client_txn_id: uniqueTxn("VISUALCOLLABORATION-PRESENCE"),
      "timeline.activity_synopsis_text": "Presence visual row",
    },
  );
  const primarySocket = installIncidentSocketMonitor(page, incidentId);

  const remotePages: Page[] = [];
  {
    await navigateVisualApplication(page, `/?incident_id=${incidentId}`);
    await primarySocket.waitForAcceptedSocket();

    for (const [index, remoteActor] of remoteActors.entries()) {
      const remoteSession = await openIncidentAsTrackedUserReady(
        browserScope,
        sessionTracker,
        {
          preparePage: async (visualPage) => {
            await initializeVisualPage(visualPage, remoteActor.user.user_id);
          },
          createdBy: `visual.collaboration.presence-${index + 1}`,
          email: remoteActor.user.email,
          incidentId,
          password: remoteActor.user.initial_password,
          purpose: "browser.collaboration visual presence overflow analyst",
          readyRecordId: presenceRow.record_id,
          userId: remoteActor.user.user_id,
        },
      );
      remotePages.push(remoteSession.page);
      await focusRemoteTimelineCellAndWaitForPresence({
        fieldKey: "timeline.activity_synopsis_text",
        primaryPage: page,
        recordId: presenceRow.record_id,
        remotePage: remoteSession.page,
        socketMonitor: primarySocket,
        ...(index === 0 ? { actorText: remoteActor.actorText } : {}),
      });
      if (index === 0) {
        const duplicatePage = await remoteSession.page.context().newPage();
        await initializeVisualPage(duplicatePage, remoteActor.user.user_id);
        const duplicateSocket = installIncidentSocketMonitor(
          duplicatePage,
          incidentId,
        );
        await duplicatePage.goto(`/?incident_id=${incidentId}`);
        await duplicateSocket.waitForAcceptedSocket();
        await focusRemoteTimelineCellAndWaitForPresence({
          fieldKey: "timeline.activity_synopsis_text",
          primaryPage: page,
          recordId: presenceRow.record_id,
          remotePage: duplicatePage,
          socketMonitor: primarySocket,
        });
      }
    }
    for (const density of ["compact", "comfortable", "default"] as const) {
      await selectVisualDensity(page, density);
      await test.step("reload visual application", () =>
        reloadVisualApplication(page));
      await withPresentation(
        page,
        {
          textSpacing:
            "#root * { letter-spacing: 0.12em !important; line-height: 1.5 !important; word-spacing: 0.16em !important; }",
        },
        async () => {
          await scrollGridTargetIntoView({
            page,
            surface: timelineViewSchemaId,
            targetTestId: cellPresenceMarkerTestId(
              presenceRow.record_id,
              "timeline.activity_synopsis_text",
            ),
          });
          for (const id of [
            rowPresenceMarkerTestId(presenceRow.record_id),
            cellPresenceMarkerTestId(
              presenceRow.record_id,
              "timeline.activity_synopsis_text",
            ),
          ]) {
            const marker = page.getByTestId(id);
            await expect(marker).toHaveAccessibleName(/^6 collaborators/);
            expect(
              await marker.evaluate((element) => {
                const cell = element
                  .closest("[data-grid-field-key]")
                  ?.getBoundingClientRect();
                if (!cell) return false;
                return Array.from(element.children).every((child) => {
                  const rect = child.getBoundingClientRect();
                  return (
                    rect.left >= cell.left &&
                    rect.right <= cell.right &&
                    rect.top >= cell.top &&
                    rect.bottom <= cell.bottom
                  );
                });
              }),
            ).toBe(true);
          }
        },
      );
    }
    await selectVisualDensity(page, null);
    await test.step("reload visual application", () =>
      reloadVisualApplication(page));

    // Nearest-target scrolling must start at a known origin. Density and
    // reload exercises can otherwise leave a different, already-visible
    // viewport that remains stable but does not match this capture's anchor.
    const preparePresenceAnchor = async () => {
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { left: "left", top: 0 },
      });
      await scrollGridTargetIntoView({
        page,
        surface: timelineViewSchemaId,
        targetTestId: cellPresenceMarkerTestId(
          presenceRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      });
      // Include the neighboring field so the edited field's marker has visible
      // context on both sides, rather than sitting against the viewport edge.
      await scrollGridCellIntoView({
        page,
        surface: timelineViewSchemaId,
        recordId: presenceRow.record_id,
        cellKey: "timeline.data_source_text",
      });
    };
    await prepareVisualPresentation(page);
    await preparePresenceAnchor();
    const presenceScroll = await readWorkbookGridScroll(
      page,
      timelineViewSchemaId,
    );
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "row-gutter",
      markerTestId: rowPresenceMarkerTestId(presenceRow.record_id),
      page,
      surface: timelineViewSchemaId,
      targetTestId: gridRowGutterTestId(
        timelineViewSchemaId,
        presenceRow.record_id,
      ),
    });
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "cell",
      markerTestId: cellPresenceMarkerTestId(
        presenceRow.record_id,
        "timeline.activity_synopsis_text",
      ),
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(
        presenceRow.record_id,
        "timeline.activity_synopsis_text",
      ),
    });
    await expect(
      page.getByTestId(workbookPresenceSummaryTestId()),
    ).toContainText("+1");
    await expect(
      page.getByTestId(rowPresenceMarkerTestId(presenceRow.record_id)),
    ).toContainText("+3");
    await expect(
      page.getByTestId(
        cellPresenceMarkerTestId(
          presenceRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      ),
    ).toContainText("+4");
    for (const testId of [
      workbookPresenceSummaryTestId(),
      rowPresenceMarkerTestId(presenceRow.record_id),
      cellPresenceMarkerTestId(
        presenceRow.record_id,
        "timeline.activity_synopsis_text",
      ),
    ]) {
      await expect(page.getByTestId(testId)).toBeInViewport({ ratio: 1 });
      await expect(page.getByTestId(testId)).toHaveAccessibleName(
        /^6 collaborators/,
      );
    }
    await assertViewportVisualRegression(
      page,
      "collaboration-presence-markers",
      {
        renderSurface: timelineViewSchemaId,
        prepareAnchor: preparePresenceAnchor,
        verifyFraming: async () => {
          expect(
            await readWorkbookGridScroll(page, timelineViewSchemaId),
          ).toEqual(presenceScroll);
        },
      },
    );

    await Promise.all(
      remotePages.map((remotePage) =>
        browserScope.closeContext(remotePage.context()),
      ),
    );
  }
});

test("Capture same-field conflict resolver.", async ({ page }) => {
  const fixture = await prepareFeP7ConflictVisual(page, {
    incidentKeyPrefix: "VISUALCOLLABORATIONRESOLVE",
    title: "browser.collaboration visual conflict resolver",
  });
  {
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: gridRowTestId(
        timelineViewSchemaId,
        fixture.conflictRow.record_id,
      ),
    });
    await expect(
      page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: fixture.conflictRow.record_id,
          surface: "grid",
        }),
      ),
    ).toHaveValue("Conflict visual local");
    await stabilizeConflictResolverVisual(page);
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "right" },
    });
    const conflictFixture = page.getByTestId(workbookShellReadyTestId());
    await conflictFixture.evaluate((element) => {
      element.setAttribute("data-design-fixture", "conflict");
    });
    await assertVisualRegression(
      page,
      "collaboration-conflict-resolver",
      conflictFixture,
    );

    await fixture.patchController.dispose();
  }
});

test("Capture conflict save-state strip.", async ({ page }) => {
  const fixture = await prepareFeP7ConflictVisual(page, {
    incidentKeyPrefix: "VISUALCOLLABORATIONCONFLICTSTRIP",
    title: "browser.collaboration visual conflict strip",
  });
  await page.getByTestId(workbookConflictControlTestId("close")).click();
  await expect(page.getByTestId(workbookConflictResolverTestId())).toHaveCount(
    0,
  );
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
  await expect(
    page.getByTestId(workbookShellSlotTestId("status-strip")),
  ).toContainText("Conflict");
  await assertStatusStripVisualFixture(page, "collaboration-conflict-strip");

  await fixture.patchController.dispose();
});

test("Capture recovered saved-state strip.", async ({ page }) => {
  const fixture = await prepareFeP7ConflictVisual(page, {
    incidentKeyPrefix: "VISUALCOLLABORATIONRECOVERED",
    title: "browser.collaboration visual recovered strip",
  });
  await page.getByTestId(workbookConflictControlTestId("keep-saved")).click();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(page.getByTestId(saveStateTestId())).toHaveAttribute(
    pendingReplayCountAttribute,
    "0",
  );
  await assertStatusStripVisualFixture(
    page,
    "collaboration-recovered-saved-strip",
  );

  await fixture.patchController.dispose();
});

test("regresses collaboration row-gutter and same-cell presence markers", async ({
  browserScope,
  page,
  sessionTracker,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALCOLLABORATIONPRESENCE"),
    "Collaboration visual presence markers",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Visual Analyst",
    email: uniqueEmail("collaboration-v6grid01-remote"),
    initial_password: "CollaborationV6Grid01!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const timelineRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualcollaborationpresence-row",
    {
      client_txn_id: uniqueTxn("VISUALCOLLABORATIONPRESENCE-ROW"),
      "timeline.activity_synopsis_text": "Presence visual row",
    },
  );
  const primarySocket = installIncidentSocketMonitor(page, incidentId);

  let remotePage: Page | null = null;
  {
    await navigateVisualApplication(page, `/?incident_id=${incidentId}`);
    await primarySocket.waitForAcceptedSocket();

    await expect(
      await mountedGridCell(
        page,
        timelineViewSchemaId,
        timelineRow.record_id,
        "timeline.activity_synopsis_text",
      ),
    ).toHaveText("Presence visual row");

    const remoteSession = await openIncidentAsTrackedUserReady(
      browserScope,
      sessionTracker,
      {
        preparePage: async (visualPage) => {
          await initializeVisualPage(visualPage, remote.user_id);
        },
        createdBy: "collaboration-visual",
        email: remote.email,
        incidentId,
        password: remote.initial_password,
        purpose: "Collaboration visual presence analyst",
        readyRecordId: timelineRow.record_id,
        userId: remote.user_id,
      },
    );
    remotePage = remoteSession.page;
    await focusRemoteTimelineCellAndWaitForPresence({
      actorText: "VA",
      fieldKey: "timeline.activity_synopsis_text",
      primaryPage: page,
      recordId: timelineRow.record_id,
      remotePage,
      socketMonitor: primarySocket,
    });
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "row-gutter",
      markerTestId: rowPresenceMarkerTestId(timelineRow.record_id),
      page,
      surface: timelineViewSchemaId,
      targetTestId: gridRowGutterTestId(
        timelineViewSchemaId,
        timelineRow.record_id,
      ),
    });
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "cell",
      markerTestId: cellPresenceMarkerTestId(
        timelineRow.record_id,
        "timeline.activity_synopsis_text",
      ),
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(
        timelineRow.record_id,
        "timeline.activity_synopsis_text",
      ),
    });

    await expect(
      page.getByTestId(rowPresenceMarkerTestId(timelineRow.record_id)),
    ).toBeInViewport({ ratio: 1 });
    await expect(
      page.getByTestId(
        cellPresenceMarkerTestId(
          timelineRow.record_id,
          "timeline.activity_synopsis_text",
        ),
      ),
    ).toBeInViewport({ ratio: 1 });
    await assertVisualRegression(
      page,
      "collaboration-grid-presence-markers",
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
      { renderSurface: timelineViewSchemaId },
    );

    if (remotePage) await browserScope.closeContext(remotePage.context());
  }
});

test("regresses collaboration same-field conflict marker resolver and Conflict strip", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALCOLLABORATIONCONFLICT"),
    "Collaboration visual conflict resolver",
  );
  const timelineRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualcollaborationconflict-row",
    {
      client_txn_id: uniqueTxn("VISUALCOLLABORATIONCONFLICT-ROW"),
      "timeline.activity_synopsis_text": "Conflict visual base",
    },
  );

  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  const patchController = await installPatchController(page);
  {
    const localValue = `Conflict_visual_local_${"L".repeat(96)}`;
    const remoteValue = `Conflict_visual_saved_${"S".repeat(96)}`;
    await driveRealTimelineSummaryConflict({
      baseRowVersion: timelineRow.row_version,
      localValue,
      page,
      patchController,
      recordId: timelineRow.record_id,
      remoteValue,
      txnPrefix: "visual-collaboration-conflict",
    });
    await stabilizeConflictResolverVisual(page);
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "right" },
    });

    await assertViewportVisualRegression(
      page,
      "collaboration-grid-conflict-resolver",
    );

    await page.setViewportSize({ width: 1024, height: 720 });
    await stabilizeConflictResolverVisual(page);
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "right" },
    });
    await assertViewportVisualRegression(
      page,
      "collaboration-grid-conflict-resolver-narrow",
    );

    await page.setViewportSize({ width: 768, height: 640 });
    await stabilizeConflictResolverVisual(page);
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "right" },
    });
    await assertViewportVisualRegression(
      page,
      "collaboration-grid-conflict-resolver-compact",
    );

    await patchController.dispose();
  }
});

test("regresses collaboration pending-queue save-state transitions", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const incidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUALCOLLABORATIONSAVE"),
    "Collaboration visual pending queue",
  );
  const syncRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualcollaborationsave-row",
    {
      client_txn_id: uniqueTxn("VISUALCOLLABORATIONSAVE-ROW"),
      "timeline.activity_utc_text": "2025-03-06T10:00:00Z",
      "timeline.activity_synopsis_text": "Pending visual base",
    },
  );
  const conflictRow = await createVisualTimelineRow(
    page,
    incidentId,
    "visualcollaborationsave-conflict-row",
    {
      client_txn_id: uniqueTxn("VISUALCOLLABORATIONSAVE-CONFLICT-ROW"),
      "timeline.activity_utc_text": "2025-03-06T10:05:00Z",
      "timeline.activity_synopsis_text": "Pending conflict visual base",
    },
  );
  await navigateVisualApplication(page, `/?incident_id=${incidentId}`);

  const summaryInput = await mountedGridCell(
    page,
    timelineViewSchemaId,
    syncRow.record_id,
    "timeline.activity_synopsis_text",
  );
  const saveState = page.getByTestId(saveStateTestId());
  const patchController = await installPatchController(page);
  const hold = patchController.holdNextPatch({ recordId: syncRow.record_id });

  {
    await summaryInput.click();
    const summaryEditor = page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.activity_synopsis_text",
        recordId: syncRow.record_id,
        surface: "grid",
      }),
    );
    await expect(summaryEditor).toBeFocused();
    await summaryEditor.fill("Pending visual syncing");
    await summaryEditor.press("Enter");
    await hold.waitForHit;
    await expect(saveState).toHaveText("Syncing");
    await assertStatusStripVisualRegression(
      page,
      "collaboration-grid-syncing-strip",
    );
    await hold.release();
    await hold.waitForCompletion;
    await expect(saveState).toHaveText("Saved");
    await assertStatusStripVisualRegression(
      page,
      "collaboration-grid-saved-strip",
    );

    await driveRealTimelineSummaryConflict({
      baseRowVersion: conflictRow.row_version,
      localValue: "Pending visual blocked",
      page,
      patchController,
      recordId: conflictRow.record_id,
      remoteValue: "Pending visual server",
      txnPrefix: "visual-collaboration-pending-conflict",
    });
    await stabilizeConflictResolverVisual(page);
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "right" },
    });
    await assertViewportVisualRegression(
      page,
      "collaboration-grid-blocked-conflict",
      {
        anchor: {
          locator: page.getByRole("region", {
            name: "Recovery navigation",
            exact: true,
          }),
          align: "start",
          scrollportSelector: 'section[aria-label="Workbook recovery"]',
        },
      },
    );

    await page.getByTestId(workbookConflictControlTestId("keep-saved")).click();
    await expect(saveState).toHaveText("Saved");

    await patchController.dispose();
  }

  await expect(saveState).toHaveText("Saved");
  await expect(page.getByTestId(saveStateTestId())).toHaveAttribute(
    pendingReplayCountAttribute,
    "0",
  );
  await assertStatusStripVisualRegression(
    page,
    "collaboration-grid-recovered-saved-strip",
  );
});
