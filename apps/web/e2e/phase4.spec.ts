import {
  applyFilterChip,
  assertGridFocusContinuity,
  gridSavedRowsSelector,
  gridShellTestId,
  removeFilterChip,
  rowCellTestId,
  rowInspectButtonTestId,
  scrollGridToBottom,
} from "@cartulary/test-utils";
import type { Page, Response } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  patchTimelineRecord,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";
const hostsViewSchemaId = "cartulary.view.hosts.v1";
const identitiesViewSchemaId = "cartulary.view.identities.v1";
const assessmentsViewSchemaId = "cartulary.view.assessments.v1";
const hostRefsFieldKey = "timeline.host_refs";
const identityRefsFieldKey = "timeline.identity_refs";

type ViewRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
};

type TimelineMutationEnvelope = {
  data: {
    change_set_id: string;
    row: ViewRow;
  };
};

type MergeEnvelope = {
  data: {
    survivor_record_id: string;
    loser_record_id: string;
    merged_into_record_id: string;
    merge_summary: {
      record_type: string;
      repointed_mention_resolution_count: number;
      repointed_link_count: number;
    };
  };
};

type CollectionItem = Record<string, unknown>;

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
  await expect(
    page.getByTestId(`row-${mainRow.record_id}-summary`),
  ).toHaveValue("E-4-01 workbook row");

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

  await createTimelineFillers(page, incidentId, "E-4-02 filler before", 6);
  const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e402-row"),
    "timeline.summary": "E-4-02 lifecycle row",
    [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
  })) as ViewRow;
  await createTimelineFillers(page, incidentId, "E-4-02 filler after", 6);
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

test("E-4-03 merges duplicate entities from the inspector and preserves survivor identity", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E403"),
    "Phase 4 E-4-03",
  );
  const survivor = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e403-survivor"),
    "host.display_name": "WS-023",
    "host.hostname": "ws-023.corp.example.test",
    "host.aliases": collectionActionsPayload(["Workstation 23"]),
  })) as ViewRow;
  const loser = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e403-loser"),
    "host.display_name": "WS-023 duplicate",
    "host.hostname": "ws-023-dup.corp.example.test",
    "host.aliases": collectionActionsPayload(["Workstation 23"]),
  })) as ViewRow;
  const identitySurvivor = (await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-identity-survivor"),
      "identity.display_name": "Alex Analyst",
      "identity.email": "alex.analyst@example.test",
      "identity.aliases": collectionActionsPayload(["Case Owner"]),
    },
  )) as ViewRow;
  const identityLoser = (await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-identity-loser"),
      "identity.display_name": "Alex Analyst duplicate",
      "identity.email": "alex.duplicate@example.test",
      "identity.aliases": collectionActionsPayload(["Case Owner"]),
    },
  )) as ViewRow;
  const dependentRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-row"),
      "timeline.summary": "E-4-03 dependent row",
      [hostRefsFieldKey]: collectionActionsPayload(["Workstation 23"]),
    },
  )) as ViewRow;
  const dependentMention = requireItemByRawText(
    collectionItems(dependentRow, hostRefsFieldKey),
    "Workstation 23",
  );
  const identityDependentRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-identity-row"),
      "timeline.summary": "E-4-03 identity dependent row",
      [identityRefsFieldKey]: collectionActionsPayload(["Case Owner"]),
    },
  )) as ViewRow;
  const identityDependentMention = requireItemByRawText(
    collectionItems(identityDependentRow, identityRefsFieldKey),
    "Case Owner",
  );
  await patchTimelineRecord(page, dependentRow.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: dependentRow.row_version,
    client_txn_id: uniqueTxn("e403-resolve"),
    changes: [
      {
        field_key: hostRefsFieldKey,
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "resolve_item",
              item_ref: dependentMention.item_ref,
              resolved_record_id: loser.record_id,
            },
          ],
        },
      },
    ],
  });
  await patchTimelineRecord(page, identityDependentRow.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: identityDependentRow.row_version,
    client_txn_id: uniqueTxn("e403-identity-resolve"),
    changes: [
      {
        field_key: identityRefsFieldKey,
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "resolve_item",
              item_ref: identityDependentMention.item_ref,
              resolved_record_id: identityLoser.record_id,
            },
          ],
        },
      },
    ],
  });

  await page.goto(`/?incident_id=${incidentId}&surface=hosts`);
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  await expect(page.getByText("Current incident role: admin")).toBeVisible();
  await expect(
    page.getByTestId(`host-row-${survivor.record_id}`),
  ).toBeVisible();
  await expect(page.getByTestId(`host-row-${loser.record_id}`)).toBeVisible();

  await page.getByTestId(`inspect-host-${survivor.record_id}`).click();
  await expect(page.getByTestId("host-inspector")).toContainText("WS-023");
  await expect(page.getByTestId("merge-start")).toBeVisible();
  await page.getByTestId("merge-start").click();
  await page.getByTestId("merge-loser-record").selectOption(loser.record_id);
  await page.getByTestId("merge-reason").fill("Phase 4 E-4-03 duplicate merge");
  await expect(page.getByTestId("merge-plan")).toContainText("Survivor WS-023");
  await expect(page.getByTestId("merge-plan")).toContainText(
    "loser WS-023 duplicate",
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    survivor.record_id,
  );
  await expect(page.getByTestId("merge-plan")).toContainText(loser.record_id);

  const mergeResponsePromise = waitForMergeResponse(page, survivor.record_id);
  await page.getByTestId("merge-confirm").click();
  const mergeEnvelope = await readMergeEnvelope(await mergeResponsePromise);

  await expect(page.getByTestId(`host-row-${loser.record_id}`)).toHaveCount(0);
  await expect(page.getByTestId("host-inspector")).toContainText("WS-023");
  await expect(page.getByTestId("merge-message")).toContainText(
    "Merged WS-023 duplicate into WS-023",
  );
  await expect(
    page.getByTestId(`timeline-preview-row-${dependentRow.record_id}`),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(`timeline-preview-row-${dependentRow.record_id}`)
      .getByLabel("Resolved WS-023"),
  ).toBeVisible();

  const hostRowsAfter = (await queryViewRows(
    page,
    incidentId,
    hostsViewSchemaId,
  )) as ViewRow[];
  const timelineRowsAfter = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const dependentRowAfter = findRow(timelineRowsAfter, dependentRow.record_id);
  const dependentHostAfter = requireItemByRawText(
    collectionItems(dependentRowAfter, hostRefsFieldKey),
    "Workstation 23",
  );

  expect(mergeEnvelope.data.survivor_record_id).toBe(survivor.record_id);
  expect(mergeEnvelope.data.loser_record_id).toBe(loser.record_id);
  expect(mergeEnvelope.data.merged_into_record_id).toBe(survivor.record_id);
  expect(mergeEnvelope.data.merge_summary.record_type).toBe("host");
  expect(
    mergeEnvelope.data.merge_summary.repointed_mention_resolution_count,
  ).toBeGreaterThan(0);
  expect(mergeEnvelope.data.merge_summary.repointed_link_count).toBeGreaterThan(
    0,
  );
  expect(
    hostRowsAfter.some((row) => row.record_id === survivor.record_id),
  ).toBeTruthy();
  expect(
    hostRowsAfter.some((row) => row.record_id === loser.record_id),
  ).toBeFalsy();
  expect(String(dependentHostAfter.item_kind)).toBe("resolved_ref");
  expect(String(dependentHostAfter.raw_text)).toBe("Workstation 23");
  expect(String(dependentHostAfter.resolved_record_id)).toBe(
    survivor.record_id,
  );

  await page.goto(`/?incident_id=${incidentId}&surface=identities`);
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  await expect(
    page.getByTestId(`identity-row-${identitySurvivor.record_id}`),
  ).toBeVisible();
  await expect(
    page.getByTestId(`identity-row-${identityLoser.record_id}`),
  ).toBeVisible();

  await page
    .getByTestId(`inspect-identity-${identitySurvivor.record_id}`)
    .click();
  await expect(page.getByTestId("identity-inspector")).toContainText(
    "Alex Analyst",
  );
  await expect(page.getByTestId("merge-start")).toBeVisible();
  await page.getByTestId("merge-start").click();
  await page
    .getByTestId("merge-loser-record")
    .selectOption(identityLoser.record_id);
  await page
    .getByTestId("merge-reason")
    .fill("Phase 4 E-4-03 identity duplicate merge");
  await expect(page.getByTestId("merge-plan")).toContainText(
    "Survivor Alex Analyst",
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    "loser Alex Analyst duplicate",
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    identitySurvivor.record_id,
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    identityLoser.record_id,
  );

  const identityMergeResponsePromise = waitForMergeResponse(
    page,
    identitySurvivor.record_id,
  );
  await page.getByTestId("merge-confirm").click();
  const identityMergeEnvelope = await readMergeEnvelope(
    await identityMergeResponsePromise,
  );

  await expect(
    page.getByTestId(`identity-row-${identityLoser.record_id}`),
  ).toHaveCount(0);
  await expect(page.getByTestId("identity-inspector")).toContainText(
    "Alex Analyst",
  );
  await expect(page.getByTestId("merge-message")).toContainText(
    "Merged Alex Analyst duplicate into Alex Analyst",
  );
  await expect(
    page.getByTestId(`timeline-preview-row-${identityDependentRow.record_id}`),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(`timeline-preview-row-${identityDependentRow.record_id}`)
      .getByLabel("Resolved Alex Analyst"),
  ).toBeVisible();

  const identityRowsAfter = (await queryViewRows(
    page,
    incidentId,
    identitiesViewSchemaId,
  )) as ViewRow[];
  const timelineRowsAfterIdentity = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const identityDependentRowAfter = findRow(
    timelineRowsAfterIdentity,
    identityDependentRow.record_id,
  );
  const dependentIdentityAfter = requireItemByRawText(
    collectionItems(identityDependentRowAfter, identityRefsFieldKey),
    "Case Owner",
  );

  expect(identityMergeEnvelope.data.survivor_record_id).toBe(
    identitySurvivor.record_id,
  );
  expect(identityMergeEnvelope.data.loser_record_id).toBe(
    identityLoser.record_id,
  );
  expect(identityMergeEnvelope.data.merged_into_record_id).toBe(
    identitySurvivor.record_id,
  );
  expect(identityMergeEnvelope.data.merge_summary.record_type).toBe("identity");
  expect(
    identityMergeEnvelope.data.merge_summary.repointed_mention_resolution_count,
  ).toBeGreaterThan(0);
  expect(
    identityMergeEnvelope.data.merge_summary.repointed_link_count,
  ).toBeGreaterThan(0);
  expect(
    identityRowsAfter.some(
      (row) => row.record_id === identitySurvivor.record_id,
    ),
  ).toBeTruthy();
  expect(
    identityRowsAfter.some((row) => row.record_id === identityLoser.record_id),
  ).toBeFalsy();
  expect(String(dependentIdentityAfter.item_kind)).toBe("resolved_ref");
  expect(String(dependentIdentityAfter.raw_text)).toBe("Case Owner");
  expect(String(dependentIdentityAfter.resolved_record_id)).toBe(
    identitySurvivor.record_id,
  );
});

test("E-4-04 auto-resolves only eligible exact-match Timeline tokens", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E404"),
    "Phase 4 E-4-04",
  );
  const autoTarget = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e404-auto"),
    "host.display_name": "Gateway node",
    "host.hostname": "gateway-node.example.test",
    "host.aliases": collectionActionsPayload(["VPN Gateway"]),
  })) as ViewRow;
  await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e404-competing-a"),
    "host.display_name": "WS-023 A",
    "host.hostname": "ws-023-a.example.test",
    "host.aliases": collectionActionsPayload(["WS-023"]),
  });
  await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e404-competing-b"),
    "host.display_name": "WS-023 B",
    "host.hostname": "ws-023-b.example.test",
    "host.aliases": collectionActionsPayload(["WS-023"]),
  });

  await createTimelineFillers(page, incidentId, "E-4-04 filler", 12);
  const suppressedRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e404-suppressed"),
      "timeline.summary": "E-4-04 suppressed row",
    },
  )) as ViewRow;
  const eligibleRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e404-eligible"),
      "timeline.summary": "E-4-04 eligible row",
    },
  )) as ViewRow;

  await page.goto(`/?incident_id=${incidentId}`);
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  const eligibleHostRefsInput = page.getByTestId(
    `row-${eligibleRow.record_id}-hostRefs-input`,
  );
  // Capture continuity baselines only after the specific row control exists;
  // the shell can render before this hydrated input is ready.
  await expect(eligibleHostRefsInput).toBeVisible();

  const autoScroll = await scrollGridToBottom(page, "timeline");
  expect(autoScroll.top).toBeGreaterThan(0);
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
  const eligibleChipId = `chip-${sanitizeTestId(String(eligibleItem.item_ref))}`;
  const eligibleRowItems = page.getByTestId(
    `row-${eligibleRow.record_id}-hostRefs-items`,
  );
  const autoNotice = page.getByTestId(
    `auto-resolution-notice-${sanitizeTestId(String(eligibleItem.item_ref))}`,
  );

  await expect(eligibleRowItems.getByTestId(eligibleChipId)).toContainText(
    "Auto",
  );
  await expect(autoNotice).toContainText("vpn gateway");
  await expect(autoNotice).toContainText("Gateway node");
  await expect(autoNotice).toContainText("VPN Gateway");
  await expect(autoNotice.getByRole("button", { name: "Undo" })).toBeVisible();
  await expect(
    autoNotice.getByRole("button", { name: "Review" }),
  ).toBeVisible();
  await expect(
    page.getByTestId(rowInspectButtonTestId(eligibleRow.record_id)),
  ).toBeFocused();
  await expectTimelineContinuity(page, eligibleRow.record_id, autoScroll);

  const undoScroll = await scrollGridToBottom(page, "timeline");
  const undoResponsePromise = waitForTimelinePatch(page, eligibleRow.record_id);
  await autoNotice.getByRole("button", { name: "Undo" }).click();
  const undoEnvelope = await readTimelineMutation(await undoResponsePromise);
  const undoneItem = requireItemByRawText(
    collectionItems(undoEnvelope.data.row, hostRefsFieldKey),
    " vpn   gateway ",
  );

  await expect(autoNotice).toHaveCount(0);
  await expect(eligibleRowItems.getByTestId(eligibleChipId)).not.toContainText(
    "Auto",
  );
  await expect(
    page.getByTestId(rowInspectButtonTestId(eligibleRow.record_id)),
  ).toBeFocused();
  await expectTimelineContinuity(page, eligibleRow.record_id, undoScroll);

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
  for (const token of suppressedTokens) {
    await addRelationshipTokenViaUI(
      page,
      suppressedRow.record_id,
      "hostRefs",
      token,
    );
    await expect(
      page.locator('[data-testid^="auto-resolution-notice-"]'),
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
  expect(String(undoneItem.item_kind)).toBe("unresolved_mention");
  expect(undoneItem.resolved_record_id).toBeUndefined();
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

test("E-4-05 creates append-only assessment history through the workbook UI", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E4ASSESS"),
    "Phase 4 assessment workbook E2E",
  );
  const subject = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-host"),
    "host.display_name": "Assessment Host",
    "host.hostname": "assessment-host.example.test",
  })) as ViewRow;
  const support = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("assessment-support"),
    "timeline.summary": "Assessment support event",
  })) as ViewRow;

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      assessmentsViewSchemaId,
    )}`,
  );
  await expect(page.getByText("Timeline workbook shell")).toBeVisible();
  await expect(page.getByTestId("assessment-create-panel")).toBeVisible();
  await expect(page.getByTestId("assessment-create-subject")).toHaveValue(
    subject.record_id,
  );

  const created: Record<string, ViewRow> = {};
  for (const entry of [
    {
      state: "unknown",
      band: "unset",
      assessedAt: "2026-04-24T10:00:00Z",
      supportRefs: [] as string[],
    },
    {
      state: "suspected",
      band: "low",
      assessedAt: "2026-04-24T11:00:00Z",
      supportRefs: [] as string[],
    },
    {
      state: "confirmed",
      band: "medium",
      assessedAt: "2026-04-24T12:00:00Z",
      supportRefs: [support.record_id],
    },
    {
      state: "disproven",
      band: "medium",
      assessedAt: "2026-04-24T13:00:00Z",
      supportRefs: [] as string[],
    },
    {
      state: "cleared",
      band: "high",
      assessedAt: "2026-04-24T14:00:00Z",
      supportRefs: [] as string[],
    },
  ]) {
    created[entry.state] = await createAssessmentViaUI(page, {
      assessedAt: entry.assessedAt,
      confidenceBand: entry.band,
      rationale: `Assessment ${entry.state} rationale.`,
      state: entry.state,
      supportRecordIds: entry.supportRefs,
    });
  }

  await expectAssessmentGridOrder(page, [
    created.cleared.record_id,
    created.disproven.record_id,
    created.confirmed.record_id,
    created.suspected.record_id,
    created.unknown.record_id,
  ]);

  await expect(
    page.getByTestId(
      rowCellTestId(created.unknown.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("unset");
  await expect(
    page.getByTestId(
      rowCellTestId(created.suspected.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("low");
  await expect(
    page.getByTestId(
      rowCellTestId(created.confirmed.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("medium");
  await expect(
    page.getByTestId(
      rowCellTestId(created.cleared.record_id, "assessment.confidence_band"),
    ),
  ).toHaveText("high");
  await expect(
    page.getByTestId(
      rowCellTestId(
        created.confirmed.record_id,
        "assessment.supporting_link_count",
      ),
    ),
  ).toHaveText("1");

  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.confidence_band",
    "high",
  );
  await expectAssessmentGridOrder(page, [created.cleared.record_id]);
  await removeFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.confidence_band",
  );
  await expectAssessmentGridOrder(page, [
    created.cleared.record_id,
    created.disproven.record_id,
    created.confirmed.record_id,
    created.suspected.record_id,
    created.unknown.record_id,
  ]);

  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
    "disproven",
  );
  await expectAssessmentGridOrder(page, [created.disproven.record_id]);
  await removeFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
  );
  await applyFilterChip(
    page,
    assessmentsViewSchemaId,
    "assessment.assessment_state",
    "cleared",
  );
  await expectAssessmentGridOrder(page, [created.cleared.record_id]);
});

async function createTimelineFillers(
  page: Page,
  incidentId: string,
  prefix: string,
  count: number,
) {
  for (let index = 1; index <= count; index += 1) {
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn(`${prefix}-${index}`),
      "timeline.summary": `${prefix} ${index}`,
    });
  }
}

async function createAssessmentViaUI(
  page: Page,
  options: {
    assessedAt: string;
    confidenceBand: string;
    rationale: string;
    state: string;
    supportRecordIds: string[];
  },
) {
  await page.getByTestId("assessment-create-state").selectOption(options.state);
  await page
    .getByTestId("assessment-create-confidence-band")
    .selectOption(options.confidenceBand);
  await page.getByTestId("assessment-create-rationale").fill(options.rationale);
  await page
    .getByTestId("assessment-create-assessed-at")
    .fill(options.assessedAt);
  if (options.supportRecordIds.length > 0) {
    await expect(
      page.getByTestId("assessment-create-support-refs").locator("option"),
    ).toHaveCount(options.supportRecordIds.length);
    await page
      .getByTestId("assessment-create-support-refs")
      .selectOption(options.supportRecordIds);
  }

  const responsePromise = waitForAssessmentCreate(page);
  await page.getByTestId("assessment-create-submit").click();
  const envelope = await readTimelineMutation(await responsePromise);
  await expect(
    page.getByTestId(`assessment-row-${envelope.data.row.record_id}`),
  ).toBeVisible();
  return envelope.data.row;
}

function waitForAssessmentCreate(page: Page) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/views/${assessmentsViewSchemaId}/rows`),
  );
}

async function expectAssessmentGridOrder(page: Page, expected: string[]) {
  const grid = page.getByTestId(gridShellTestId(assessmentsViewSchemaId));
  await expect
    .poll(async () =>
      grid
        .locator(gridSavedRowsSelector())
        .evaluateAll((rows) =>
          rows.map((row) => row.getAttribute("data-grid-record-id") ?? ""),
        ),
    )
    .toEqual(expected);
}

function collectionActionsPayload(rawTexts: string[]) {
  return {
    kind: "collection_actions_v1",
    actions: rawTexts.map((rawText) => ({
      op: "add_token",
      raw_text: rawText,
    })),
  };
}

function collectionItems(
  row: ViewRow | Record<string, unknown>,
  fieldKey: string,
) {
  const cells = (row as { cells: Record<string, { value: unknown }> }).cells;
  const cellValue = cells[fieldKey]?.value;
  if (
    !cellValue ||
    typeof cellValue !== "object" ||
    Array.isArray(cellValue) ||
    !("items" in cellValue)
  ) {
    return [] as CollectionItem[];
  }
  const items = (cellValue as { items?: unknown }).items;
  if (!Array.isArray(items)) {
    return [] as CollectionItem[];
  }
  return items.filter(
    (item): item is CollectionItem =>
      typeof item === "object" && item !== null && !Array.isArray(item),
  );
}

function findRow(rows: ViewRow[], recordId: string) {
  const row = rows.find((candidate) => candidate.record_id === recordId);
  if (!row) {
    throw new Error(`missing row ${recordId}`);
  }
  return row;
}

function requireItemByRawText(items: CollectionItem[], rawText: string) {
  const item = items.find((candidate) => candidate.raw_text === rawText);
  if (!item) {
    throw new Error(`missing collection item raw_text=${rawText}`);
  }
  return item;
}

function sanitizeTestId(value: string) {
  return value.replace(/[^a-zA-Z0-9_-]+/gu, "-");
}

async function waitForSaveState(
  page: Page,
  value: "Saved" | "Syncing" | "Conflict",
) {
  await expect(page.getByTestId("save-state")).toHaveText(value);
}

async function addRelationshipTokenViaUI(
  page: Page,
  recordId: string,
  draftKey: "hostRefs" | "identityRefs",
  rawText: string,
) {
  const responsePromise = waitForTimelinePatch(page, recordId);
  await page.getByTestId(`row-${recordId}-${draftKey}-input`).fill(rawText);
  await page.getByTestId(`row-${recordId}-${draftKey}-input`).press("Enter");
  await waitForSaveState(page, "Saved");
  return readTimelineMutation(await responsePromise);
}

async function openTimelineInspector(page: Page, recordId: string) {
  await page.getByTestId(rowInspectButtonTestId(recordId)).click();
  await expect(page.getByTestId("timeline-inspector")).toContainText(recordId);
}

function waitForTimelinePatch(page: Page, recordId: string) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${recordId}`),
  );
}

function waitForMergeResponse(page: Page, survivorRecordId: string) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${survivorRecordId}/merge`),
  );
}

async function readTimelineMutation(response: Response) {
  expect(response.ok()).toBeTruthy();
  return (await response.json()) as TimelineMutationEnvelope;
}

async function readMergeEnvelope(response: Response) {
  expect(response.ok()).toBeTruthy();
  return (await response.json()) as MergeEnvelope;
}

async function expectTimelineContinuity(
  page: Page,
  recordId: string,
  preservedScroll: { left: number; top: number },
  options: {
    requireExactHorizontalScroll?: boolean;
    requireExactVerticalScroll?: boolean;
  } = {},
) {
  await expect
    .poll(() => new URL(page.url()).searchParams.get("surface"))
    .toBe("timeline");
  await assertGridFocusContinuity({
    focusTestId: rowInspectButtonTestId(recordId),
    page,
    preservedScroll,
    requireExactHorizontalScroll: options.requireExactHorizontalScroll ?? false,
    requireExactVerticalScroll: options.requireExactVerticalScroll ?? false,
    surface: "timeline",
  });
}

async function waitForViewRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  recordId: string,
) {
  await expect
    .poll(async () => {
      const rows = (await queryViewRows(page, incidentId, viewSchemaId)) as
        | ViewRow[]
        | Array<Record<string, unknown>>;
      return rows.some(
        (candidate) =>
          typeof candidate === "object" &&
          candidate !== null &&
          "record_id" in candidate &&
          candidate.record_id === recordId,
      );
    })
    .toBe(true);
  const rows = (await queryViewRows(
    page,
    incidentId,
    viewSchemaId,
  )) as ViewRow[];
  return findRow(rows, recordId);
}
