import {
  assertMarkerAnchoredToGridTarget,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  cellPresenceMarkerTestId,
  gridRowGutterTestId,
  rowCellTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
  timelineScalarEditorTestId,
  workbookConflictControlTestId,
  workbookConflictLocalValueTestId,
  workbookConflictResolverTestId,
  workbookConflictSavedValueTestId,
  workbookConflictSummaryTestId,
  workbookPresenceSummaryTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { expectAllInteractiveControlsNamed } from "./support/accessibility/assertions";
import { expectNoFocusTrap } from "./support/accessibility/keyboard";
import { withPresentation } from "./support/accessibility/presentation";
import { expectAndRecordContrast } from "./support/accessibility/records";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  driveRealTimelineSummaryConflict,
  focusRemoteTimelineCellAndWaitForPresence,
  installPatchController,
  openIncidentAsTrackedUserReady,
  requireRecordId,
} from "./support/collaboration/replay";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { semanticGridCell } from "./support/workbook/keyboardGrid";
import { expectResolverSurfaceGeometry } from "./support/workbook/presentationGeometry";
import { createViewRow } from "./support/workbook/query";
import { observeSaveEvents, saveEvents } from "./support/workbook/saveStatus";

test("a11y.presence remains named and anchored without save or live-region noise", async ({
  browserScope,
  page,
  sessionTracker,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YCOLLABORATION"),
    "a11y.collaboration conflict accessibility",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Accessible Analyst",
    email: uniqueEmail("a11y.collaboration-remote"),
    initial_password: "A11yCollaborationRemotePass!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.collaboration-row"),
    "timeline.activity_synopsis_text": "a11y.collaboration conflict base",
  });
  const recordId = requireRecordId(row);
  const primarySocket = installIncidentSocketMonitor(page, incidentId);
  let remotePage: Page | null = null;
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await primarySocket.waitForAcceptedSocket();
  await observeSaveEvents(page);
  const remoteSession = await openIncidentAsTrackedUserReady(
    browserScope,
    sessionTracker,
    {
      createdBy: "a11y.collaboration.row-01",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "a11y.collaboration remote presence analyst",
      readyRecordId: recordId,
      userId: remote.user_id,
    },
  );
  remotePage = remoteSession.page;
  await focusRemoteTimelineCellAndWaitForPresence({
    actorText: "AA",
    fieldKey: "timeline.activity_synopsis_text",
    primaryPage: page,
    recordId,
    remotePage,
    socketMonitor: primarySocket,
  });
  expect(await saveEvents(page)).toEqual([]);
  const presenceIds = [
    workbookPresenceSummaryTestId(),
    rowPresenceMarkerTestId(recordId),
    cellPresenceMarkerTestId(recordId, "timeline.activity_synopsis_text"),
  ] as const;
  const checkPresenceGeometry = async (scrollPage = false) => {
    await scrollGridCellIntoView({
      page,
      surface: timelineViewSchemaId,
      recordId,
      cellKey: "timeline.activity_synopsis_text",
    });
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: presenceIds[2],
    });
    for (const id of presenceIds) {
      const marker = page.getByTestId(id);
      if (scrollPage) await marker.scrollIntoViewIfNeeded();
      await expect(marker).toBeInViewport({ ratio: 1 });
      await expect(marker).toHaveAccessibleName(
        /^1 collaborator .*Accessible Analyst editing/,
      );
      expect(
        await marker.evaluate((element) => {
          for (
            let node: Element | null = element;
            node;
            node = node.parentElement
          ) {
            if (
              ["status", "alert", "log"].includes(
                node.getAttribute("role") ?? "",
              ) ||
              ["polite", "assertive"].includes(
                node.getAttribute("aria-live") ?? "",
              )
            )
              return false;
          }
          return true;
        }),
      ).toBe(true);
    }
    expect(await saveEvents(page)).toEqual([]);
    await assertMarkerAnchoredToGridTarget({
      page,
      surface: timelineViewSchemaId,
      anchorKind: "row-gutter",
      markerTestId: presenceIds[1],
      targetTestId: gridRowGutterTestId(timelineViewSchemaId, recordId),
    });
    await assertMarkerAnchoredToGridTarget({
      page,
      surface: timelineViewSchemaId,
      anchorKind: "cell",
      markerTestId: presenceIds[2],
      targetTestId: rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    });
  };
  const originalPreferences = (
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
          client_txn_id: uniqueTxn("presence-density"),
          density_mode: density,
        },
      },
    );
    expect(response.ok()).toBeTruthy();
    await page.reload();
    await observeSaveEvents(page);
  };
  try {
    for (const density of ["compact", "default", "comfortable"]) {
      await setDensity(density);
      for (const viewport of [
        { width: 1440, height: 900 },
        { width: 1024, height: 720 },
        { width: 768, height: 640 },
      ]) {
        await page.setViewportSize(viewport);
        await checkPresenceGeometry();
      }
      // Keep the effective viewport within the supported desktop band.
      await page.setViewportSize({ width: 1600, height: 1440 });
      await withPresentation(page, { zoom: "200%" }, async () => {
        // The existing CSS zoom exercise retains CSS viewport height;
        // reach the responsive footer through normal document scrolling.
        await checkPresenceGeometry(true);
      });
      await page.setViewportSize({ width: 1440, height: 900 });
      await withPresentation(
        page,
        {
          textSpacing:
            "#root * { letter-spacing: 0.12em !important; line-height: 1.5 !important; word-spacing: 0.16em !important; } #root p { margin-bottom: 2em !important; }",
        },
        async () => {
          await checkPresenceGeometry();
        },
      );
      await page
        .getByTestId(rowCellTestId(recordId, "timeline.activity_synopsis_text"))
        .click();
      const editor = page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId,
          surface: "grid",
        }),
      );
      await expect(editor).toBeFocused();
      const editorBox = await editor.boundingBox();
      const markerBox = await page.getByTestId(presenceIds[2]).boundingBox();
      expect(editorBox).not.toBeNull();
      expect(markerBox).not.toBeNull();
      expect((editorBox?.x ?? 0) + (editorBox?.width ?? 0)).toBeLessThanOrEqual(
        markerBox?.x ?? 0,
      );
      await editor.press("Escape");
    }
  } finally {
    await page.setViewportSize({ width: 1440, height: 900 });
    await setDensity(originalPreferences.density_mode);
  }
  if (remotePage) await browserScope.closeContext(remotePage.context());
});

test("a11y.conflict recovery retains values focus and reachable controls", async ({
  browserScope,
  page,
  sessionTracker,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("A11YCOLLABORATION"),
    "a11y.collaboration conflict accessibility",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Accessible Analyst",
    email: uniqueEmail("a11y.collaboration-remote"),
    initial_password: "A11yCollaborationRemotePass!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("a11y.collaboration-row"),
    "timeline.activity_synopsis_text": "a11y.collaboration conflict base",
  });
  const recordId = requireRecordId(row);
  const patchController = await installPatchController(page);
  const primarySocket = installIncidentSocketMonitor(page, incidentId);
  let remotePage: Page | null = null;
  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await primarySocket.waitForAcceptedSocket();
  const remoteSession = await openIncidentAsTrackedUserReady(
    browserScope,
    sessionTracker,
    {
      createdBy: "a11y.collaboration.row-01",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "a11y.collaboration remote presence analyst",
      readyRecordId: recordId,
      userId: remote.user_id,
    },
  );
  remotePage = remoteSession.page;
  await focusRemoteTimelineCellAndWaitForPresence({
    actorText: "AA",
    fieldKey: "timeline.activity_synopsis_text",
    primaryPage: page,
    recordId,
    remotePage,
    socketMonitor: primarySocket,
  });
  const localConflictValue = `a11y_collaboration_local_${"L".repeat(96)}`;
  const savedConflictValue = `a11y_collaboration_saved_${"S".repeat(96)}`;
  await driveRealTimelineSummaryConflict({
    baseRowVersion: 1,
    localValue: localConflictValue,
    page,
    patchController,
    recordId,
    remotePatchPage: remotePage,
    remoteValue: savedConflictValue,
    txnPrefix: "fea11y-collaboration-conflict",
  });
  const resolver = page.getByTestId(workbookConflictResolverTestId());
  await expect(resolver).toHaveAttribute(
    "aria-label",
    "Workbook conflict recovery",
  );
  const summary = page.getByTestId(workbookConflictSummaryTestId());
  const panelHeading = page
    .getByRole("region", { name: "Recovery navigation", exact: true })
    .locator(":scope > h2");
  await expect(panelHeading).toBeFocused();
  await expect(resolver).toHaveAttribute(
    "data-conflict-field-key",
    "timeline.activity_synopsis_text",
  );
  await expect(
    page.getByTestId(workbookConflictSavedValueTestId()),
  ).toHaveValue(savedConflictValue);
  await expect(
    page.getByTestId(workbookConflictLocalValueTestId()),
  ).toHaveValue(localConflictValue);
  await expect(
    page.getByRole("button", {
      name: "Close conflict recovery",
      exact: true,
    }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Discard local draft" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Use my unsaved value" }),
  ).toBeVisible();
  await expect(
    page.getByRole("button", { name: "Use merged value" }),
  ).toBeVisible();
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 720 },
    { width: 768, height: 640 },
  ]) {
    await page.setViewportSize(viewport);
    await expectResolverSurfaceGeometry(page);
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  await withPresentation(page, { zoom: "200%" }, async () => {
    await expectResolverSurfaceGeometry(page, {
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
      await expectResolverSurfaceGeometry(page);
    },
  );
  await page.setViewportSize({ width: 1440, height: 900 });
  await resolver.evaluate((element) => {
    element.scrollTop = 0;
  });
  await summary.press("Enter");
  await expect(resolver).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
  const retainedConflictEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId,
      surface: "grid",
    }),
  );
  const retainedConflictMarker = semanticGridCell(
    retainedConflictEditor,
  ).locator('[data-grid-state-marker="conflicted"]');
  await expect(retainedConflictMarker).toBeVisible();
  await expect(retainedConflictMarker).toHaveAttribute(
    "aria-label",
    "Conflict on Activity Synopsis",
  );
  await expectAllInteractiveControlsNamed(page);
  await expectNoFocusTrap(page);
  await expectAndRecordContrast(page, [
    saveStateTestId(),
    rowPresenceMarkerTestId(recordId),
    cellPresenceMarkerTestId(recordId, "timeline.activity_synopsis_text"),
    workbookConflictControlTestId("close"),
    workbookConflictControlTestId("keep-saved"),
    workbookConflictControlTestId("use-unsaved"),
    workbookConflictControlTestId("use-merged"),
  ]);
  await panelHeading.focus();
  await page.keyboard.press("Escape");
  await expect(resolver).toHaveCount(0);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
  await expect(
    semanticGridCell(retainedConflictEditor).getByRole("img", {
      name: "Conflict on Activity Synopsis",
    }),
  ).toBeVisible();
  await patchController.dispose();
  if (remotePage) await browserScope.closeContext(remotePage.context());
});
