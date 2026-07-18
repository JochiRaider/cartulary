import {
  scrollGridCellIntoView,
  scrollGridToBottom,
} from "@cartulary/test-utils/grid";
import {
  mentionDismissButtonTestId,
  mentionItemTestId,
  mentionRestoreUnresolvedButtonTestId,
  relationshipItemsTestId,
  rowCellTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { Page, Response } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "./support/contracts/workbookSurfaces";
import {
  collectionActionsPayload,
  collectionItems,
  findRow,
  hostRefsFieldKey,
  requireItemByRawText,
  type ViewRow,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  createTimelineFillers,
  timelineFixtureOccurredAt,
} from "./support/timeline/fixtures";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import {
  expectTimelineContinuity,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

test("dismisses and ordinarily restores a mention without relinking", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("MENTION-LIFECYCLE"),
    "Record relationships entity-resolution",
  );
  const existingHost = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e402-host"),
      "host.display_name": "WS-023",
      "host.hostname": "ws-023.corp.example.test",
    },
  )) as ViewRow;

  await createTimelineFillers(
    page,
    incidentId,
    "entity-resolution filler before",
    6,
    {
      occurredAtStart: timelineFixtureOccurredAt(0),
    },
  );
  const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e402-row"),
    "timeline.activity_utc_text": timelineFixtureOccurredAt(6),
    "timeline.activity_synopsis_text": "entity-resolution lifecycle row",
    [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
  })) as ViewRow;
  await createTimelineFillers(
    page,
    incidentId,
    "entity-resolution filler after",
    6,
    {
      occurredAtStart: timelineFixtureOccurredAt(7),
    },
  );
  const seededMention = requireItemByRawText(
    collectionItems(row, hostRefsFieldKey),
    "WS-023?",
  );
  await patchRecord(page, row.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.row_version,
    client_txn_id: uniqueTxn("e402-resolve-setup"),
    changes: [
      {
        field_key: hostRefsFieldKey,
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "resolve_item",
              item_ref: seededMention.item_ref,
              resolved_record_id: existingHost.record_id,
            },
          ],
        },
      },
    ],
  });

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: row.record_id,
    surface: timelineViewSchemaId,
  });
  await page
    .getByTestId(
      rowCellTestId(row.record_id, "timeline.activity_synopsis_text"),
    )
    .focus();
  await openTimelineInspector(page, row.record_id);
  await expect(
    page
      .getByTestId(relationshipItemsTestId(row.record_id, hostRefsFieldKey))
      .getByLabel("Resolved WS-023"),
  ).toBeVisible();
  const initialTimelineRows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const rowBeforeDismiss = findRow(initialTimelineRows, row.record_id);
  const mentionBeforeDismiss = requireItemByRawText(
    collectionItems(rowBeforeDismiss, hostRefsFieldKey),
    "WS-023?",
  );
  const rowIndexBeforeDismiss = initialTimelineRows.findIndex(
    (candidate) => candidate.record_id === row.record_id,
  );
  expect(rowIndexBeforeDismiss).toBeGreaterThanOrEqual(0);
  expect(rowIndexBeforeDismiss).toBeGreaterThan(0);
  expect(rowIndexBeforeDismiss).toBeLessThan(initialTimelineRows.length - 1);

  await page
    .getByTestId(mentionItemTestId(String(seededMention.item_ref)))
    .click();

  const dismissScroll = await scrollGridToBottom(page, timelineViewSchemaId);
  const dismissResponsePromise = waitForMentionAction(
    page,
    seededMention.item_ref,
  );
  await page.getByTestId(mentionDismissButtonTestId()).click();
  const dismissResponse = await dismissResponsePromise;
  const dismissEnvelope = await readMentionAction(dismissResponse);
  const dismissBody = readMentionActionRequest(dismissResponse);

  await expect(
    page.getByTestId(relationshipItemsTestId(row.record_id, hostRefsFieldKey)),
  ).toContainText("No items");
  await expect(
    page
      .getByTestId(mentionItemTestId(String(seededMention.item_ref)))
      .getByLabel("Dismissed WS-023?"),
  ).toBeVisible();
  await expectTimelineContinuity(page, row.record_id, dismissScroll, {
    requireExactVerticalScroll: false,
  });
  expect(dismissBody).toMatchObject({
    base_mention_row_version: mentionBeforeDismiss.mention_row_version,
    action: "dismiss_item",
  });
  expect(dismissBody).not.toHaveProperty("resolved_record_id");
  expect(dismissEnvelope.data.entity_mention.resolution_status).toBe(
    "dismissed",
  );
  const rowsAfterDismiss = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  expect(
    collectionItems(findRow(rowsAfterDismiss, row.record_id), hostRefsFieldKey),
  ).toHaveLength(0);

  const restoreScroll = await scrollGridToBottom(page, timelineViewSchemaId);
  const restoreResponsePromise = waitForMentionAction(
    page,
    seededMention.item_ref,
  );
  await page.getByTestId(mentionRestoreUnresolvedButtonTestId()).click();
  const restoreResponse = await restoreResponsePromise;
  const restoreEnvelope = await readMentionAction(restoreResponse);
  const restoreBody = readMentionActionRequest(restoreResponse);

  await expect(
    page
      .getByTestId(relationshipItemsTestId(row.record_id, hostRefsFieldKey))
      .getByLabel("Unresolved WS-023?"),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(relationshipItemsTestId(row.record_id, hostRefsFieldKey))
      .getByLabel(/^Resolved WS-023$/),
  ).toHaveCount(0);
  await expectTimelineContinuity(page, row.record_id, restoreScroll, {
    requireExactVerticalScroll: false,
  });
  expect(restoreBody).toMatchObject({
    base_mention_row_version: dismissEnvelope.data.entity_mention.row_version,
    action: "revert_to_unresolved",
  });
  expect(restoreBody).not.toHaveProperty("resolved_record_id");
  expect(restoreEnvelope.data.entity_mention.resolution_status).toBe(
    "unresolved",
  );

  const timelineRows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const restoredRow = findRow(timelineRows, row.record_id);
  const restoredRowItem = requireItemByRawText(
    collectionItems(restoredRow, hostRefsFieldKey),
    "WS-023?",
  );

  expect(String(restoredRowItem.item_kind)).toBe("unresolved_mention");
  expect(restoredRowItem.resolved_record_id).toBeUndefined();
});

type MentionActionEnvelope = {
  data: {
    entity_mention: {
      resolution_status: string;
      row_version: number;
    };
  };
};

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
