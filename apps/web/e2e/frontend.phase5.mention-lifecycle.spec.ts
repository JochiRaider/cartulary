import {
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  mentionRestoreUnresolvedButtonTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  timelineInspectorTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { Page, Response, Route } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  addRelationshipTokenViaUI,
  aliasCollectionActionsPayload,
  collectionActionsPayload,
  collectionItems,
  findRow,
  hostRefsFieldKey,
  hostsViewSchemaId,
  openTimelineInspector,
  requireItemByRawText,
  timelineViewSchemaId,
  type ViewRow,
} from "./phase4Helpers";

const exactScenarioTitle =
  "FE-E-P5-01 Verify manual mention resolution, dismissal, auto-resolution disclosure, and undo through public mutation routes and refreshed rows.";

test(exactScenarioTitle, async ({ page }) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEEP501"),
    "Phase 5 FE-E-P5-01",
  );
  const manualTarget = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn("feep501-manual-target"),
      "host.display_name": "FE-E-P5 Manual Target",
      "host.hostname": "feep501-manual-target.example.test",
    },
  )) as ViewRow;
  const autoTarget = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("feep501-auto-target"),
    "host.display_name": "FE-E-P5 Auto Target",
    "host.hostname": "feep501-auto-target.example.test",
    "host.aliases": aliasCollectionActionsPayload(["FEEP501 Auto Alias"]),
  })) as ViewRow;
  const correctionTarget = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn("feep501-correction-target"),
      "host.display_name": "FE-E-P5 Corrected Target",
      "host.hostname": "feep501-corrected-target.example.test",
    },
  )) as ViewRow;
  const manualRawText = "FEEP501 Manual Raw";
  const autoRawText = "FEEP501 Auto Alias";
  const manualRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("feep501-manual-row"),
      "timeline.activity_synopsis_text": "FE-E-P5 manual lifecycle row",
      [hostRefsFieldKey]: collectionActionsPayload([manualRawText]),
    },
  )) as ViewRow;
  const autoCorrectionRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("feep501-auto-correction-row"),
      "timeline.activity_synopsis_text": "FE-E-P5 auto correction row",
    },
  )) as ViewRow;
  const autoUndoRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("feep501-auto-undo-row"),
      "timeline.activity_synopsis_text": "FE-E-P5 auto undo row",
    },
  )) as ViewRow;
  const manualMention = requireItemByRawText(
    collectionItems(manualRow, hostRefsFieldKey),
    manualRawText,
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();

  await openTimelineInspector(page, manualRow.record_id);
  await page
    .getByTestId(mentionItemTestId(String(manualMention.item_ref)))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    manualRawText,
  );

  const manualResolveResponsePromise = waitForMentionAction(
    page,
    manualMention.item_ref,
  );
  await page
    .getByTestId(mentionResolveTargetSelectTestId())
    .selectOption(manualTarget.record_id);
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  const manualResolveResponse = await manualResolveResponsePromise;
  const manualResolveEnvelope = await readMentionAction(manualResolveResponse);
  const manualResolveBody = readMentionActionRequest(manualResolveResponse);
  expect(manualResolveBody).toMatchObject({
    base_mention_row_version: manualMention.mention_row_version,
    action: "resolve_item",
    resolved_record_id: manualTarget.record_id,
  });
  expect(typeof manualResolveBody.client_txn_id).toBe("string");
  expect(manualResolveEnvelope.data.entity_mention.resolution_status).toBe(
    "resolved",
  );

  const manualResolvedChip = page
    .getByTestId(relationshipItemsTestId(manualRow.record_id, hostRefsFieldKey))
    .getByTestId(relationshipChipTestId(String(manualMention.item_ref)));
  await expect(manualResolvedChip).toContainText("Manual");
  await expect(manualResolvedChip).not.toContainText("Auto");
  await page
    .getByTestId(mentionItemTestId(String(manualMention.item_ref)))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    manualRawText,
  );
  const manualResolvedRow = await refreshedTimelineRow(
    page,
    incidentId,
    manualRow.record_id,
  );
  const manualResolvedItem = requireItemByRawText(
    collectionItems(manualResolvedRow, hostRefsFieldKey),
    manualRawText,
  );
  expect(String(manualResolvedItem.item_kind)).toBe("resolved_ref");
  expect(String(manualResolvedItem.resolved_record_id)).toBe(
    manualTarget.record_id,
  );
  expect(manualResolvedItem.auto_resolved).not.toBe(true);

  const dismissResponsePromise = waitForMentionAction(
    page,
    manualMention.item_ref,
  );
  await page.getByTestId(mentionDismissButtonTestId()).click();
  const dismissResponse = await dismissResponsePromise;
  const dismissEnvelope = await readMentionAction(dismissResponse);
  const dismissBody = readMentionActionRequest(dismissResponse);
  expect(dismissBody).toMatchObject({
    base_mention_row_version: manualResolvedItem.mention_row_version,
    action: "dismiss_item",
  });
  expect(dismissBody).not.toHaveProperty("resolved_record_id");
  expect(dismissEnvelope.data.entity_mention.resolution_status).toBe(
    "dismissed",
  );
  await expect(
    page.getByTestId(
      relationshipItemsTestId(manualRow.record_id, hostRefsFieldKey),
    ),
  ).toContainText("No items");
  await expect(
    page
      .getByTestId(mentionItemTestId(String(manualMention.item_ref)))
      .getByLabel(`Dismissed ${manualRawText}`),
  ).toBeVisible();
  const manualDismissedRow = await refreshedTimelineRow(
    page,
    incidentId,
    manualRow.record_id,
  );
  expect(collectionItems(manualDismissedRow, hostRefsFieldKey)).toHaveLength(0);

  const restoreResponsePromise = waitForMentionAction(
    page,
    manualMention.item_ref,
  );
  await page.getByTestId(mentionRestoreUnresolvedButtonTestId()).click();
  const restoreResponse = await restoreResponsePromise;
  const restoreEnvelope = await readMentionAction(restoreResponse);
  const restoreBody = readMentionActionRequest(restoreResponse);
  expect(restoreBody).toMatchObject({
    base_mention_row_version: dismissEnvelope.data.entity_mention.row_version,
    action: "revert_to_unresolved",
  });
  expect(restoreBody).not.toHaveProperty("resolved_record_id");
  expect(restoreEnvelope.data.entity_mention.resolution_status).toBe(
    "unresolved",
  );
  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(manualRow.record_id, hostRefsFieldKey),
      )
      .getByLabel(`Unresolved ${manualRawText}`),
  ).toBeVisible();
  const manualRestoredRow = await refreshedTimelineRow(
    page,
    incidentId,
    manualRow.record_id,
  );
  const manualRestoredItem = requireItemByRawText(
    collectionItems(manualRestoredRow, hostRefsFieldKey),
    manualRawText,
  );
  expect(String(manualRestoredItem.item_kind)).toBe("unresolved_mention");

  const autoCorrectionEnvelope = await addRelationshipTokenViaUI(
    page,
    autoCorrectionRow.record_id,
    "hostRefs",
    autoRawText,
  );
  const autoCorrectionItem = requireItemByRawText(
    collectionItems(autoCorrectionEnvelope.data.row, hostRefsFieldKey),
    autoRawText,
  );
  const autoCorrectionChip = page
    .getByTestId(
      relationshipItemsTestId(autoCorrectionRow.record_id, hostRefsFieldKey),
    )
    .getByTestId(relationshipChipTestId(String(autoCorrectionItem.item_ref)));
  await expect(autoCorrectionChip).toContainText("Auto");
  const autoCorrectionNotice = page.getByTestId(
    autoResolutionNoticeTestId(String(autoCorrectionItem.item_ref)),
  );
  await expect(autoCorrectionNotice).toContainText("FE-E-P5 Auto Target");
  await expect(autoCorrectionNotice).toContainText("FEEP501 Auto Alias");
  await autoCorrectionNotice
    .getByTestId(
      autoResolutionReviewButtonTestId(String(autoCorrectionItem.item_ref)),
    )
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    autoRawText,
  );
  await expect(autoCorrectionNotice).toBeVisible();

  const removeFailedCorrectionRoute = await routeFailedMentionActionOnce(
    page,
    autoCorrectionItem.item_ref,
    "feep501-auto-correction-conflict",
  );
  const failedCorrectionResponsePromise = waitForMentionAction(
    page,
    autoCorrectionItem.item_ref,
  );
  await page
    .getByTestId(mentionResolveTargetSelectTestId())
    .selectOption(correctionTarget.record_id);
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  const failedCorrectionResponse = await failedCorrectionResponsePromise;
  expect(failedCorrectionResponse.ok()).toBeFalsy();
  await expect(autoCorrectionNotice).toBeVisible();
  await removeFailedCorrectionRoute();

  const correctionResponsePromise = waitForMentionAction(
    page,
    autoCorrectionItem.item_ref,
  );
  await page
    .getByTestId(mentionResolveTargetSelectTestId())
    .selectOption(correctionTarget.record_id);
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  const correctionResponse = await correctionResponsePromise;
  const correctionEnvelope = await readMentionAction(correctionResponse);
  const correctionBody = readMentionActionRequest(correctionResponse);
  expect(correctionBody).toMatchObject({
    base_mention_row_version: autoCorrectionItem.mention_row_version,
    action: "resolve_item",
    resolved_record_id: correctionTarget.record_id,
  });
  expect(correctionEnvelope.data.entity_mention.resolution_status).toBe(
    "resolved",
  );
  await expect(autoCorrectionChip).toContainText("Manual");
  await expect(autoCorrectionChip).not.toContainText("Auto");
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    autoRawText,
  );
  await expect(autoCorrectionNotice).toHaveCount(0);
  const autoCorrectedRow = await refreshedTimelineRow(
    page,
    incidentId,
    autoCorrectionRow.record_id,
  );
  const autoCorrectedItem = requireItemByRawText(
    collectionItems(autoCorrectedRow, hostRefsFieldKey),
    autoRawText,
  );
  expect(String(autoCorrectedItem.resolved_record_id)).toBe(
    correctionTarget.record_id,
  );
  expect(autoCorrectedItem.auto_resolved).not.toBe(true);

  const autoUndoEnvelope = await addRelationshipTokenViaUI(
    page,
    autoUndoRow.record_id,
    "hostRefs",
    autoRawText,
  );
  const autoUndoItem = requireItemByRawText(
    collectionItems(autoUndoEnvelope.data.row, hostRefsFieldKey),
    autoRawText,
  );
  expect(String(autoUndoItem.resolved_record_id)).toBe(autoTarget.record_id);
  expect(autoUndoItem.auto_resolved).toBe(true);
  const autoUndoNotice = page.getByTestId(
    autoResolutionNoticeTestId(String(autoUndoItem.item_ref)),
  );
  await expect(autoUndoNotice).toContainText("FE-E-P5 Auto Target");
  const removeFailedUndoRoute = await routeFailedMentionActionOnce(
    page,
    autoUndoItem.item_ref,
    "feep501-auto-undo-conflict",
  );
  const failedUndoResponsePromise = waitForMentionAction(
    page,
    autoUndoItem.item_ref,
  );
  await autoUndoNotice
    .getByTestId(autoResolutionUndoButtonTestId(String(autoUndoItem.item_ref)))
    .click();
  const failedUndoResponse = await failedUndoResponsePromise;
  expect(failedUndoResponse.ok()).toBeFalsy();
  await expect(autoUndoNotice).toBeVisible();
  await removeFailedUndoRoute();

  const undoResponsePromise = waitForMentionAction(page, autoUndoItem.item_ref);
  await autoUndoNotice
    .getByTestId(autoResolutionUndoButtonTestId(String(autoUndoItem.item_ref)))
    .click();
  const undoResponse = await undoResponsePromise;
  const undoEnvelope = await readMentionAction(undoResponse);
  const undoBody = readMentionActionRequest(undoResponse);
  expect(undoBody).toMatchObject({
    base_mention_row_version: autoUndoItem.mention_row_version,
    action: "revert_to_unresolved",
  });
  expect(undoBody).not.toHaveProperty("resolved_record_id");
  expect(undoEnvelope.data.entity_mention.resolution_status).toBe("unresolved");
  await expect(autoUndoNotice).toHaveCount(0);
  const autoUndoRefreshedRow = await refreshedTimelineRow(
    page,
    incidentId,
    autoUndoRow.record_id,
  );
  const autoUndoRefreshedItem = requireItemByRawText(
    collectionItems(autoUndoRefreshedRow, hostRefsFieldKey),
    autoRawText,
  );
  expect(String(autoUndoRefreshedItem.item_kind)).toBe("unresolved_mention");
  expect(autoUndoRefreshedItem.resolved_record_id).toBeUndefined();
  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(autoUndoRow.record_id, hostRefsFieldKey),
      )
      .getByLabel(`Unresolved ${autoRawText}`),
  ).toBeVisible();
});

type MentionActionEnvelope = {
  data: {
    entity_mention: {
      entity_mention_id: string;
      raw_text: string;
      resolution_status: string;
      resolved_record_id: string | null;
      row_version: number;
    };
    source_record: {
      record_id: string;
      row_version: number;
    };
    change_set_id: string;
  };
};

async function routeFailedMentionActionOnce(
  page: Page,
  itemRef: unknown,
  requestId: string,
) {
  const mentionId = entityMentionIdFromItemRef(itemRef);
  const routePattern = `**/api/v1/entity-mentions/${mentionId}/resolve`;
  let handled = false;
  const routeHandler = async (route: Route) => {
    if (handled || route.request().method().toUpperCase() !== "POST") {
      await route.fallback();
      return;
    }
    handled = true;
    await route.fulfill({
      status: 409,
      contentType: "application/json",
      body: JSON.stringify({
        error: {
          status: 409,
          code: "row_version_conflict",
          message: "row version conflict",
          request_id: requestId,
          retryable: false,
          details: {
            reason_code: "stale_row_version",
          },
        },
      }),
    });
  };
  await page.route(routePattern, routeHandler);
  return async () => {
    await page.unroute(routePattern, routeHandler);
  };
}

function waitForMentionAction(page: Page, itemRef: unknown) {
  const mentionId = entityMentionIdFromItemRef(itemRef);
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/entity-mentions/${mentionId}/resolve`),
  );
}

async function readMentionAction(response: Response) {
  expect(response.ok()).toBeTruthy();
  return (await response.json()) as MentionActionEnvelope;
}

function readMentionActionRequest(response: Response) {
  return JSON.parse(response.request().postData() ?? "{}") as Record<
    string,
    unknown
  >;
}

function entityMentionIdFromItemRef(itemRef: unknown) {
  const value = String(itemRef);
  expect(value.startsWith("entity_mention:")).toBe(true);
  return value.slice("entity_mention:".length);
}

async function refreshedTimelineRow(
  page: Page,
  incidentId: string,
  recordId: string,
) {
  const rows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  return findRow(rows, recordId);
}
