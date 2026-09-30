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
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, test } from "./fixtures";
import {
  collectionActionsPayload,
  collectionItems,
  findRow,
  hostRefsFieldKey,
  identityRefsFieldKey,
  readMentionAction,
  readMentionActionRequest,
  requireItemByRawText,
  waitForMentionAction,
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
  expectTimelineMutationContinuity,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

test("dismisses and ordinarily restores a mention without relinking", async ({
  page,
}, testInfo) => {
  for (const entityType of ["host", "identity"] as const) {
    const fieldKey =
      entityType === "host" ? hostRefsFieldKey : identityRefsFieldKey;
    const targetDisplayName = entityType === "host" ? "WS-023" : "Alex Analyst";
    const rawText =
      entityType === "host" ? "WS-023?" : "alex-source@example.test?";
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("MENTION-LIFECYCLE"),
      "Record relationships entity-resolution",
    );
    const existingEntity = await createViewRow(
      page,
      incidentId,
      entityType === "host" ? hostsViewSchemaId : identitiesViewSchemaId,
      {
        client_txn_id: uniqueTxn("e402-host"),
        ...(entityType === "host"
          ? {
              "host.display_name": targetDisplayName,
              "host.hostname": "ws-023.corp.example.test",
            }
          : {
              "identity.display_name": targetDisplayName,
              "identity.upn": "alex@example.test",
            }),
      },
    );

    await createTimelineFillers(
      page,
      incidentId,
      "entity-resolution filler before",
      6,
      {
        occurredAtStart: timelineFixtureOccurredAt(0),
      },
    );
    const row = await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("e402-row"),
      "timeline.activity_utc_text": timelineFixtureOccurredAt(6),
      "timeline.activity_synopsis_text": "entity-resolution lifecycle row",
      [fieldKey]: collectionActionsPayload([rawText]),
    });
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
      collectionItems(row, fieldKey),
      rawText,
    );
    await patchRecord(page, row.record_id, {
      view_schema_id: timelineViewSchemaId,
      base_row_version: row.row_version,
      client_txn_id: uniqueTxn("e402-resolve-setup"),
      changes: [
        {
          field_key: fieldKey,
          action_payload: {
            kind: "collection_actions_v1",
            actions: [
              {
                op: "resolve_item",
                item_ref: String(seededMention.item_ref),
                resolved_record_id: existingEntity.record_id,
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
        .getByTestId(relationshipItemsTestId(row.record_id, fieldKey))
        .getByLabel(`Resolved ${entityType}: ${targetDisplayName}`, {
          exact: true,
        }),
    ).toBeVisible();
    const initialTimelineRows = await queryViewRows(
      page,
      incidentId,
      timelineViewSchemaId,
    );
    const rowBeforeDismiss = findRow(initialTimelineRows, row.record_id);
    const mentionBeforeDismiss = requireItemByRawText(
      collectionItems(rowBeforeDismiss, fieldKey),
      rawText,
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

    const selectedCorrection = page.getByRole("region", {
      name:
        entityType === "host"
          ? "Selected Hosts item"
          : "Selected Identities item",
    });
    await expect(selectedCorrection).toBeVisible();
    expect(
      await selectedCorrection.evaluate((element) =>
        element.previousElementSibling?.getAttribute("data-testid"),
      ),
    ).toBe(mentionItemTestId(String(seededMention.item_ref)));
    await testInfo.attach(`inspector-relationship-correction-${entityType}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    const dismissScroll = await scrollGridToBottom(page, timelineViewSchemaId);
    const dismissResponsePromise = waitForMentionAction(page, seededMention);
    await page.getByTestId(mentionDismissButtonTestId()).click();
    const dismissResponse = await dismissResponsePromise;
    const dismissEnvelope = await readMentionAction(
      dismissResponse,
      row.record_id,
    );
    const dismissBody = readMentionActionRequest(dismissResponse);

    await expect(
      page.getByTestId(relationshipItemsTestId(row.record_id, fieldKey)),
    ).toContainText("No items");
    await expect(
      page.getByTestId(mentionItemTestId(String(seededMention.item_ref))),
    ).toHaveAccessibleName(`Dismissed mention: ${rawText}`);
    await expectTimelineMutationContinuity(
      page,
      row.record_id,
      dismissEnvelope.data.source_record.row_version,
      dismissScroll,
      {
        requireExactVerticalScroll: false,
      },
    );
    expect(dismissBody).toMatchObject({
      base_mention_row_version: mentionBeforeDismiss.mention_row_version,
      action: "dismiss_item",
    });
    expect(dismissBody).not.toHaveProperty("resolved_record_id");
    expect(dismissEnvelope.data.entity_mention.resolution_status).toBe(
      "dismissed",
    );
    expect(dismissEnvelope.data.entity_mention).toMatchObject({
      raw_text: rawText,
      resolved_record_id: null,
      resolution_method: null,
    });
    const rowsAfterDismiss = await queryViewRows(
      page,
      incidentId,
      timelineViewSchemaId,
    );
    expect(
      findRow(rowsAfterDismiss, row.record_id).cells[
        "timeline.has_unresolved_mentions"
      ],
    ).toMatchObject({ value: false });
    expect(
      collectionItems(findRow(rowsAfterDismiss, row.record_id), fieldKey),
    ).toHaveLength(0);

    expect(
      await selectedCorrection.evaluate((element) =>
        element.previousElementSibling?.getAttribute("data-testid"),
      ),
    ).toBe(mentionItemTestId(String(seededMention.item_ref)));
    await testInfo.attach(`inspector-relationship-dismissed-${entityType}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
    const restoreScroll = await scrollGridToBottom(page, timelineViewSchemaId);
    const restoreResponsePromise = waitForMentionAction(page, seededMention);
    await page.getByTestId(mentionRestoreUnresolvedButtonTestId()).click();
    const restoreResponse = await restoreResponsePromise;
    const restoreEnvelope = await readMentionAction(
      restoreResponse,
      row.record_id,
    );
    const restoreBody = readMentionActionRequest(restoreResponse);

    await expect(
      page
        .getByTestId(relationshipItemsTestId(row.record_id, fieldKey))
        .getByLabel(`Unresolved ${entityType} mention: ${rawText}`, {
          exact: true,
        }),
    ).toBeVisible();
    await expect(
      page
        .getByTestId(relationshipItemsTestId(row.record_id, fieldKey))
        .getByLabel(`Resolved ${entityType}: ${targetDisplayName}`, {
          exact: true,
        }),
    ).toHaveCount(0);
    await expectTimelineMutationContinuity(
      page,
      row.record_id,
      restoreEnvelope.data.source_record.row_version,
      restoreScroll,
      {
        requireExactVerticalScroll: false,
      },
    );
    expect(restoreBody).toMatchObject({
      base_mention_row_version: dismissEnvelope.data.entity_mention.row_version,
      action: "revert_to_unresolved",
    });
    expect(restoreBody).not.toHaveProperty("resolved_record_id");
    expect(restoreEnvelope.data.entity_mention.resolution_status).toBe(
      "unresolved",
    );

    expect(restoreEnvelope.data.entity_mention).toMatchObject({
      raw_text: rawText,
      resolved_record_id: null,
      resolution_method: null,
    });
    const timelineRows = await queryViewRows(
      page,
      incidentId,
      timelineViewSchemaId,
    );
    const restoredRow = findRow(timelineRows, row.record_id);
    const restoredRowItem = requireItemByRawText(
      collectionItems(restoredRow, fieldKey),
      rawText,
    );

    expect(String(restoredRowItem.item_kind)).toBe("unresolved_mention");
    expect(restoredRowItem.resolved_record_id).toBeUndefined();
  }
});
