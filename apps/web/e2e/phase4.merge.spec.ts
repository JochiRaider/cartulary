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
  findRow,
  hostRefsFieldKey,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  identityRefsFieldKey,
  readMergeEnvelope,
  requireItemByRawText,
  timelineViewSchemaId,
  type ViewRow,
  waitForMergeResponse,
} from "./phase4Helpers";

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

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      hostsViewSchemaId,
    )}`,
  );
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

  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      identitiesViewSchemaId,
    )}`,
  );
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
