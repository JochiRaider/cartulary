import { scrollGridToBottom } from "@cartulary/test-utils";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  patchTimelineRecord,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  collectionActionsPayload,
  collectionItems,
  createTimelineFillers,
  expectTimelineContinuity,
  findRow,
  hostRefsFieldKey,
  hostsViewSchemaId,
  openTimelineInspector,
  readTimelineMutation,
  requireItemByRawText,
  sanitizeTestId,
  timelineFixtureOccurredAt,
  timelineViewSchemaId,
  type ViewRow,
  waitForTimelinePatch,
} from "./phase4Helpers";

test("E-4-02 dismisses and ordinarily restores a mention without relinking", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E402"),
    "Phase 4 E-4-02",
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

  await createTimelineFillers(page, incidentId, "E-4-02 filler before", 6, {
    occurredAtStart: timelineFixtureOccurredAt(0),
  });
  const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e402-row"),
    "timeline.occurred_at": timelineFixtureOccurredAt(6),
    "timeline.summary": "E-4-02 lifecycle row",
    [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
  })) as ViewRow;
  await createTimelineFillers(page, incidentId, "E-4-02 filler after", 6, {
    occurredAtStart: timelineFixtureOccurredAt(7),
  });
  const seededMention = requireItemByRawText(
    collectionItems(row, hostRefsFieldKey),
    "WS-023?",
  );
  await patchTimelineRecord(page, row.record_id, {
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
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  await expect(
    page
      .getByTestId(`row-${row.record_id}-hostRefs-items`)
      .getByLabel("Resolved WS-023"),
  ).toBeVisible();
  const initialTimelineRows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const rowIndexBeforeDismiss = initialTimelineRows.findIndex(
    (candidate) => candidate.record_id === row.record_id,
  );
  expect(rowIndexBeforeDismiss).toBeGreaterThanOrEqual(0);
  expect(rowIndexBeforeDismiss).toBeGreaterThan(0);
  expect(rowIndexBeforeDismiss).toBeLessThan(initialTimelineRows.length - 1);

  await openTimelineInspector(page, row.record_id);
  await page
    .getByTestId(`mention-${sanitizeTestId(String(seededMention.item_ref))}`)
    .click();

  const dismissScroll = await scrollGridToBottom(page, "timeline");
  expect(dismissScroll.top).toBeGreaterThan(0);
  const dismissResponsePromise = waitForTimelinePatch(page, row.record_id);
  await page.getByRole("button", { name: "Dismiss" }).click();
  const dismissEnvelope = await readTimelineMutation(
    await dismissResponsePromise,
  );

  await expect(
    page.getByTestId(`row-${row.record_id}-hostRefs-items`),
  ).toContainText("No items");
  await expect(
    page
      .getByTestId(`mention-${sanitizeTestId(String(seededMention.item_ref))}`)
      .getByLabel("Dismissed WS-023?"),
  ).toBeVisible();
  await expectTimelineContinuity(page, row.record_id, dismissScroll, {
    requireExactVerticalScroll: false,
  });
  expect(
    collectionItems(dismissEnvelope.data.row, hostRefsFieldKey),
  ).toHaveLength(0);

  const restoreScroll = await scrollGridToBottom(page, "timeline");
  const restoreResponsePromise = waitForTimelinePatch(page, row.record_id);
  await page.getByRole("button", { name: "Restore to unresolved" }).click();
  const restoreResponse = await restoreResponsePromise;
  const restoreEnvelope = await readTimelineMutation(restoreResponse);
  const restoreBody = JSON.parse(
    restoreResponse.request().postData() ?? "{}",
  ) as {
    changes: Array<{
      action_payload: {
        actions: Array<{ item_ref: string }>;
      };
    }>;
  };
  const restoredItem = requireItemByRawText(
    collectionItems(restoreEnvelope.data.row, hostRefsFieldKey),
    "WS-023?",
  );

  await expect(
    page
      .getByTestId(`row-${row.record_id}-hostRefs-items`)
      .getByLabel("Unresolved WS-023?"),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(`row-${row.record_id}-hostRefs-items`)
      .getByLabel(/^Resolved WS-023$/),
  ).toHaveCount(0);
  await expectTimelineContinuity(page, row.record_id, restoreScroll, {
    requireExactVerticalScroll: false,
  });
  expect(restoreBody.changes[0]?.action_payload.actions[0]?.item_ref).toBe(
    seededMention.item_ref,
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

  expect(String(restoredItem.item_kind)).toBe("unresolved_mention");
  expect(restoredItem.resolved_record_id).toBeUndefined();
  expect(String(restoredRowItem.item_kind)).toBe("unresolved_mention");
  expect(restoredRowItem.resolved_record_id).toBeUndefined();
});
