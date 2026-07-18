import { scrollGridToBottom } from "@cartulary/test-utils/grid";
import {
  mentionCreateEntityButtonTestId,
  mentionItemTestId,
  mentionResolveExistingButtonTestId,
  mentionResolveTargetSelectTestId,
  relationshipItemsTestId,
  rowCellTestId,
  timelineInspectorTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { Page, Response } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "./support/contracts/workbookSurfaces";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  collectionItems,
  findRow,
  hostRefsFieldKey,
  identityRefsFieldKey,
  requireItemByRawText,
  type ViewRow,
} from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createTimelineFillers } from "./support/timeline/fixtures";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import {
  ensureTimelineGridTargetVisible,
  expectNoPendingQueueAuthPause,
  expectTimelineContinuity,
  openTimelineInspector,
  waitForViewRow,
} from "./support/workbook/rowMutations";

test("resolves and creates entities from Timeline mentions in the inspector", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E401"),
    "Record relationships E-4-01",
  );
  const existingHost = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn("e401-host"),
      "host.display_name": "WS-023",
      "host.hostname": "ws-023.corp.example.test",
    },
  )) as ViewRow;

  await createTimelineFillers(page, incidentId, "E-4-01 filler", 12);
  const siblingRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e401-sibling"),
      "timeline.activity_synopsis_text": "E-4-01 sibling unresolved",
      [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
    },
  )) as ViewRow;
  const mainRow = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e401-main"),
    "timeline.activity_synopsis_text": "E-4-01 workbook row",
  })) as ViewRow;
  const identitiesBefore = await queryViewRows(
    page,
    incidentId,
    identitiesViewSchemaId,
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const mainSummaryTestId = rowCellTestId(
    mainRow.record_id,
    "timeline.activity_synopsis_text",
  );
  await ensureTimelineGridTargetVisible(page, mainSummaryTestId);
  await expect(page.getByTestId(mainSummaryTestId)).toHaveText(
    "E-4-01 workbook row",
  );

  const hostAddEnvelope = await addRelationshipTokenViaUI(
    page,
    mainRow.record_id,
    "hostRefs",
    "WS-023?",
  );
  const identityAddEnvelope = await addRelationshipTokenViaUI(
    page,
    mainRow.record_id,
    "identityRefs",
    "vpn.user@example.test",
  );
  const hostMention = requireItemByRawText(
    collectionItems(hostAddEnvelope.data.row, hostRefsFieldKey),
    "WS-023?",
  );
  const identityMention = requireItemByRawText(
    collectionItems(identityAddEnvelope.data.row, identityRefsFieldKey),
    "vpn.user@example.test",
  );

  await expect(
    page
      .getByTestId(relationshipItemsTestId(mainRow.record_id, hostRefsFieldKey))
      .getByLabel("Unresolved WS-023?"),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(mainRow.record_id, identityRefsFieldKey),
      )
      .getByLabel("Unresolved vpn.user@example.test"),
  ).toBeVisible();

  await openTimelineInspector(page, mainRow.record_id);
  await page
    .getByTestId(mentionItemTestId(String(hostMention.item_ref)))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    "Raw token",
  );
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    "WS-023?",
  );

  const resolveScroll = await scrollGridToBottom(page, timelineViewSchemaId);
  await expectNoPendingQueueAuthPause(page, "before resolving host mention");
  const resolveResponsePromise = waitForMentionAction(
    page,
    hostMention.item_ref,
  );
  await page
    .getByTestId(mentionResolveTargetSelectTestId())
    .selectOption(existingHost.record_id);
  await page.getByTestId(mentionResolveExistingButtonTestId()).click();
  const resolveResponse = await resolveResponsePromise;
  const resolveEnvelope = await readMentionAction(resolveResponse);
  const resolveBody = readMentionActionRequest(resolveResponse);

  await expect(
    page
      .getByTestId(relationshipItemsTestId(mainRow.record_id, hostRefsFieldKey))
      .getByLabel("Resolved WS-023"),
  ).toBeVisible();
  await expectTimelineContinuity(page, mainRow.record_id, resolveScroll);

  await page
    .getByTestId(mentionItemTestId(String(identityMention.item_ref)))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    "vpn.user@example.test",
  );

  const createScroll = await scrollGridToBottom(page, timelineViewSchemaId);
  await expectNoPendingQueueAuthPause(page, "before creating identity mention");
  const createResponsePromise = waitForMentionAction(
    page,
    identityMention.item_ref,
  );
  await page.getByTestId(mentionCreateEntityButtonTestId("identity")).click();
  const createResponse = await createResponsePromise;
  const createEnvelope = await readMentionAction(createResponse);
  const createdIdentityRecordId = String(
    createEnvelope.data.entity_mention.resolved_record_id,
  );
  expect(createdIdentityRecordId).not.toBe("null");
  expect(createdIdentityRecordId).not.toBe("");
  const createBody = readMentionActionRequest(createResponse);

  await expect(
    page
      .getByTestId(
        relationshipItemsTestId(mainRow.record_id, identityRefsFieldKey),
      )
      .getByLabel("Resolved vpn.user@example.test"),
  ).toBeVisible();
  await expectTimelineContinuity(page, mainRow.record_id, createScroll);

  const timelineRows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const mainRowAfter = findRow(timelineRows, mainRow.record_id);
  const siblingRowAfter = findRow(timelineRows, siblingRow.record_id);
  const mainHostAfter = requireItemByRawText(
    collectionItems(mainRowAfter, hostRefsFieldKey),
    "WS-023?",
  );
  const mainIdentityAfter = requireItemByRawText(
    collectionItems(mainRowAfter, identityRefsFieldKey),
    "vpn.user@example.test",
  );
  const siblingHostAfter = requireItemByRawText(
    collectionItems(siblingRowAfter, hostRefsFieldKey),
    "WS-023?",
  );
  const createdIdentityRow = await waitForViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    createdIdentityRecordId,
  );

  expect(identitiesBefore).toHaveLength(0);
  expect(resolveBody).toMatchObject({
    base_mention_row_version: hostMention.mention_row_version,
    action: "resolve_item",
    resolved_record_id: existingHost.record_id,
  });
  expect(createBody).toMatchObject({
    base_mention_row_version: identityMention.mention_row_version,
    action: "resolve_item",
    resolved_record_id: createdIdentityRecordId,
  });
  expect(typeof resolveBody.client_txn_id).toBe("string");
  expect(resolveEnvelope.data.entity_mention.resolved_record_id).toBe(
    existingHost.record_id,
  );
  expect(String(mainHostAfter.item_kind)).toBe("resolved_ref");
  expect(String(mainHostAfter.raw_text)).toBe("WS-023?");
  expect(String(mainHostAfter.resolved_record_id)).toBe(existingHost.record_id);
  expect(String(mainIdentityAfter.item_kind)).toBe("resolved_ref");
  expect(String(mainIdentityAfter.raw_text)).toBe("vpn.user@example.test");
  expect(String(mainIdentityAfter.resolved_record_id)).toBe(
    createdIdentityRecordId,
  );
  expect(String(siblingHostAfter.item_kind)).toBe("unresolved_mention");
  expect(siblingHostAfter.resolved_record_id).toBeUndefined();
  expect(createdIdentityRow.record_id).toBe(createdIdentityRecordId);
});

type MentionActionEnvelope = {
  data: {
    entity_mention: {
      resolved_record_id: string | null;
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
