import { test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  patchRecord,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  aliasCollectionActionsPayload,
  collectionActionsPayload,
  collectionItems,
  exerciseEntityMerge,
  hostRefsFieldKey,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  identityRefsFieldKey,
  requireItemByRawText,
  timelineViewSchemaId,
  type ViewRow,
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
    "host.aliases": aliasCollectionActionsPayload(["Workstation 23"]),
  })) as ViewRow;
  const loser = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e403-loser"),
    "host.display_name": "WS-023 duplicate",
    "host.hostname": "ws-023-dup.corp.example.test",
    "host.aliases": aliasCollectionActionsPayload(["Workstation 23"]),
  })) as ViewRow;
  const identitySurvivor = (await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-identity-survivor"),
      "identity.display_name": "Alex Analyst",
      "identity.email": "alex.analyst@example.test",
      "identity.aliases": aliasCollectionActionsPayload(["Case Owner"]),
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
      "identity.aliases": aliasCollectionActionsPayload(["Case Owner"]),
    },
  )) as ViewRow;
  const dependentRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-row"),
      "timeline.activity_synopsis_text": "E-4-03 dependent row",
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
      "timeline.activity_synopsis_text": "E-4-03 identity dependent row",
      [identityRefsFieldKey]: collectionActionsPayload(["Case Owner"]),
    },
  )) as ViewRow;
  const identityDependentMention = requireItemByRawText(
    collectionItems(identityDependentRow, identityRefsFieldKey),
    "Case Owner",
  );
  await patchRecord(page, dependentRow.record_id, {
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
  await patchRecord(page, identityDependentRow.record_id, {
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

  await exerciseEntityMerge(page, {
    dependentRow,
    entityType: "host",
    expectAdminRole: true,
    fieldKey: hostRefsFieldKey,
    incidentId,
    loser,
    loserLabel: "WS-023 duplicate",
    mergeReason: "Phase 4 E-4-03 duplicate merge",
    rawText: "Workstation 23",
    recordType: "host",
    resolvedLabel: "WS-023",
    survivor,
    survivorLabel: "WS-023",
    viewSchemaId: hostsViewSchemaId,
  });

  await exerciseEntityMerge(page, {
    dependentRow: identityDependentRow,
    entityType: "identity",
    fieldKey: identityRefsFieldKey,
    incidentId,
    loser: identityLoser,
    loserLabel: "Alex Analyst duplicate",
    mergeReason: "Phase 4 E-4-03 identity duplicate merge",
    rawText: "Case Owner",
    recordType: "identity",
    resolvedLabel: "Alex Analyst",
    survivor: identitySurvivor,
    survivorLabel: "Alex Analyst",
    viewSchemaId: identitiesViewSchemaId,
  });
});
