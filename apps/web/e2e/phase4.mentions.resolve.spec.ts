import { scrollGridToBottom } from "@cartulary/test-utils";

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
  collectionActionsPayload,
  collectionItems,
  createTimelineFillers,
  ensureTimelineGridTargetVisible,
  expectNoPendingQueueAuthPause,
  expectTimelineContinuity,
  findRow,
  hostRefsFieldKey,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  identityRefsFieldKey,
  openTimelineInspector,
  readTimelineMutation,
  requireItemByRawText,
  sanitizeTestId,
  timelineViewSchemaId,
  type ViewRow,
  waitForTimelinePatch,
  waitForViewRow,
} from "./phase4Helpers";

test("E-4-01 resolves and creates entities from Timeline mentions in the inspector", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E401"),
    "Phase 4 E-4-01",
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
      "timeline.summary": "E-4-01 sibling unresolved",
      [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
    },
  )) as ViewRow;
  const mainRow = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e401-main"),
    "timeline.summary": "E-4-01 workbook row",
  })) as ViewRow;
  const identitiesBefore = await queryViewRows(
    page,
    incidentId,
    identitiesViewSchemaId,
  );

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  const mainSummaryTestId = `row-${mainRow.record_id}-summary`;
  await ensureTimelineGridTargetVisible(page, mainSummaryTestId);
  await expect(page.getByTestId(mainSummaryTestId)).toHaveValue(
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
      .getByTestId(`row-${mainRow.record_id}-hostRefs-items`)
      .getByLabel("Unresolved WS-023?"),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(`row-${mainRow.record_id}-identityRefs-items`)
      .getByLabel("Unresolved vpn.user@example.test"),
  ).toBeVisible();

  await openTimelineInspector(page, mainRow.record_id);
  await page
    .getByTestId(`mention-${sanitizeTestId(String(hostMention.item_ref))}`)
    .click();
  await expect(page.getByTestId("timeline-inspector")).toContainText(
    "Raw token",
  );
  await expect(page.getByTestId("timeline-inspector")).toContainText("WS-023?");

  const resolveScroll = await scrollGridToBottom(page, "timeline");
  await expectNoPendingQueueAuthPause(page, "before resolving host mention");
  const resolveResponsePromise = waitForTimelinePatch(page, mainRow.record_id);
  await page
    .getByTestId("inspector-resolve-target")
    .selectOption(existingHost.record_id);
  await page.getByRole("button", { name: "Resolve to existing" }).click();
  const resolveEnvelope = await readTimelineMutation(
    await resolveResponsePromise,
  );
  const resolvedHostItem = requireItemByRawText(
    collectionItems(resolveEnvelope.data.row, hostRefsFieldKey),
    "WS-023?",
  );

  await expect(
    page
      .getByTestId(`row-${mainRow.record_id}-hostRefs-items`)
      .getByLabel("Resolved WS-023"),
  ).toBeVisible();
  await expectTimelineContinuity(page, mainRow.record_id, resolveScroll);

  await page
    .getByTestId(`mention-${sanitizeTestId(String(identityMention.item_ref))}`)
    .click();
  await expect(page.getByTestId("timeline-inspector")).toContainText(
    "vpn.user@example.test",
  );

  const createScroll = await scrollGridToBottom(page, "timeline");
  await expectNoPendingQueueAuthPause(page, "before creating identity mention");
  const createResponsePromise = waitForTimelinePatch(page, mainRow.record_id);
  await page.getByRole("button", { name: "Create identity" }).click();
  const createEnvelope = await readTimelineMutation(
    await createResponsePromise,
  );
  const createdIdentityItem = requireItemByRawText(
    collectionItems(createEnvelope.data.row, identityRefsFieldKey),
    "vpn.user@example.test",
  );
  const createdIdentityRecordId = String(
    createdIdentityItem.resolved_record_id,
  );

  await expect(
    page
      .getByTestId(`row-${mainRow.record_id}-identityRefs-items`)
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
  expect(String(resolvedHostItem.item_kind)).toBe("resolved_ref");
  expect(String(resolvedHostItem.resolved_record_id)).toBe(
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
