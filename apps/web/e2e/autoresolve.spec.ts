import { scrollGridToBottom } from "@cartulary/test-utils/grid";
import {
  autoResolutionNoticeFamilySelector,
  autoResolutionNoticeTestId,
  autoResolutionReviewButtonTestId,
  autoResolutionUndoButtonTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  timelineCollectionInputTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import {
  addRelationshipTokenViaUI,
  aliasCollectionActionsPayload,
  collectionItems,
  findRow,
  hostRefsFieldKey,
  readMentionAction,
  readMentionActionRequest,
  requireItemByRawText,
  type ViewRow,
  waitForMentionAction,
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
  expectTimelineMutationContinuity,
  openTimelineInspector,
  readTimelineMutation,
  waitForTimelinePatch,
} from "./support/workbook/rowMutations";

test("auto-resolves only eligible exact-match Timeline tokens", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ENTITY-AUTORESOLUTION"),
    "Record relationships entity-resolution",
  );
  const autoTarget = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e404-auto"),
    "host.display_name": "Gateway node",
    "host.hostname": "gateway-node.example.test",
    "host.aliases": aliasCollectionActionsPayload(["VPN Gateway"]),
  })) as ViewRow;
  await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e404-competing-a"),
    "host.display_name": "WS-023 A",
    "host.hostname": "ws-023-a.example.test",
    "host.aliases": aliasCollectionActionsPayload(["WS-023"]),
  });
  await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e404-competing-b"),
    "host.display_name": "WS-023 B",
    "host.hostname": "ws-023-b.example.test",
    "host.aliases": aliasCollectionActionsPayload(["WS-023"]),
  });

  await createTimelineFillers(page, incidentId, "entity-resolution filler", 32);
  const suppressedRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e404-suppressed"),
      "timeline.activity_synopsis_text": "entity-resolution suppressed row",
    },
  )) as ViewRow;
  const eligibleRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e404-eligible"),
      "timeline.activity_synopsis_text": "entity-resolution eligible row",
    },
  )) as ViewRow;

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  const eligibleHostRefsInputTestId = timelineCollectionInputTestId(
    eligibleRow.record_id,
    hostRefsFieldKey,
  );
  const autoScroll = await ensureTimelineGridTargetVisible(
    page,
    rowCellTestId(eligibleRow.record_id, "timeline.activity_synopsis_text"),
  );
  await openTimelineInspector(page, eligibleRow.record_id);
  const eligibleHostRefsInput = page.getByTestId(eligibleHostRefsInputTestId);
  // Capture continuity baselines only after the specific row control exists;
  // the shell can render before this hydrated input is ready.
  await eligibleHostRefsInput.focus();
  await expect(eligibleHostRefsInput).toBeVisible();

  const eligibleResponsePromise = waitForTimelinePatch(
    page,
    eligibleRow.record_id,
  );
  await eligibleHostRefsInput.fill(" vpn   gateway ");
  await eligibleHostRefsInput.press("Enter");
  const eligibleEnvelope = await readTimelineMutation(
    await eligibleResponsePromise,
  );
  const eligibleItem = requireItemByRawText(
    collectionItems(eligibleEnvelope.data.row, hostRefsFieldKey),
    " vpn   gateway ",
  );
  const eligibleChipId = relationshipChipTestId(String(eligibleItem.item_ref));
  const eligibleRowItems = page.getByTestId(
    relationshipItemsTestId(eligibleRow.record_id, hostRefsFieldKey),
  );
  const autoNotice = page.getByTestId(
    autoResolutionNoticeTestId(String(eligibleItem.item_ref)),
  );

  await expect(eligibleRowItems.getByTestId(eligibleChipId)).toContainText(
    "Auto",
  );
  await expect(autoNotice).toContainText("vpn gateway");
  await expect(autoNotice).toContainText("Gateway node");
  await expect(autoNotice).toContainText("VPN Gateway");
  await expect(
    autoNotice.getByTestId(
      autoResolutionUndoButtonTestId(String(eligibleItem.item_ref)),
    ),
  ).toBeVisible();
  await expect(
    autoNotice.getByTestId(
      autoResolutionReviewButtonTestId(String(eligibleItem.item_ref)),
    ),
  ).toBeVisible();
  await expectTimelineMutationContinuity(
    page,
    eligibleRow.record_id,
    eligibleEnvelope.data.row.row_version,
    autoScroll,
  );

  const undoScroll = await scrollGridToBottom(page, timelineViewSchemaId);
  const undoResponsePromise = waitForMentionAction(page, eligibleItem.item_ref);
  await autoNotice
    .getByTestId(autoResolutionUndoButtonTestId(String(eligibleItem.item_ref)))
    .click();
  const undoResponse = await undoResponsePromise;
  const undoEnvelope = await readMentionAction(
    undoResponse,
    eligibleRow.record_id,
  );
  const undoBody = readMentionActionRequest(undoResponse);

  await expect(autoNotice).toHaveCount(0);
  await expect(eligibleRowItems.getByTestId(eligibleChipId)).not.toContainText(
    "Auto",
  );
  await expectTimelineMutationContinuity(
    page,
    eligibleRow.record_id,
    undoEnvelope.data.source_record.row_version,
    undoScroll,
  );
  expect(undoBody).toMatchObject({
    base_mention_row_version: eligibleItem.mention_row_version,
    action: "revert_to_unresolved",
  });
  expect(undoBody).not.toHaveProperty("resolved_record_id");
  expect(undoEnvelope.data.entity_mention.resolution_status).toBe("unresolved");

  const suppressedTokens = [
    "WS-023",
    "WS-023?",
    "WS-023??",
    "WS-023 ~",
    "WS-023 maybe",
    "WS-023 prob",
    "WS-023 probably",
    "WS-023 approx",
    "WS-023 approximately",
    "(WS-023)",
    "WS-023.",
    "WS-023,",
    "WS-023 likely",
  ];
  let expectedSuppressedBaseRowVersion = suppressedRow.row_version;
  for (const token of suppressedTokens) {
    let patchBaseRowVersion: unknown;
    const envelope = await addRelationshipTokenViaUI(
      page,
      suppressedRow.record_id,
      "hostRefs",
      token,
      {
        onPatchRequest: (payload) => {
          patchBaseRowVersion = payload.base_row_version;
        },
      },
    );
    expect(patchBaseRowVersion).toBe(expectedSuppressedBaseRowVersion);
    expectedSuppressedBaseRowVersion = envelope.data.row.row_version;
    await expect(
      page.locator(autoResolutionNoticeFamilySelector()),
    ).toHaveCount(0);
  }

  const timelineRows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const eligibleRowAfter = findRow(timelineRows, eligibleRow.record_id);
  const suppressedRowAfter = findRow(timelineRows, suppressedRow.record_id);
  const eligibleItemAfterUndo = requireItemByRawText(
    collectionItems(eligibleRowAfter, hostRefsFieldKey),
    " vpn   gateway ",
  );
  const suppressedItems = collectionItems(suppressedRowAfter, hostRefsFieldKey);

  expect(String(eligibleItem.item_kind)).toBe("resolved_ref");
  expect(String(eligibleItem.resolved_record_id)).toBe(autoTarget.record_id);
  expect(eligibleItem.auto_resolved).toBe(true);
  expect(String(eligibleItem.provenance)).toBe("auto_match");
  expect(eligibleItem.confidence).toBe(100);
  expect(String(eligibleItem.matched_alias_text)).toBe("VPN Gateway");
  expect(String(eligibleItemAfterUndo.item_kind)).toBe("unresolved_mention");
  expect(eligibleItemAfterUndo.resolved_record_id).toBeUndefined();

  for (const token of suppressedTokens) {
    const item = requireItemByRawText(suppressedItems, token);
    expect(String(item.item_kind)).toBe("unresolved_mention");
    expect(item.resolved_record_id).toBeUndefined();
    expect(item.provenance).toBeUndefined();
    expect(item.confidence).toBeUndefined();
    expect(item.matched_alias_text).toBeUndefined();
  }
});
