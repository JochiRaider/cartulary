import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { test } from "./fixtures";
import {
  aliasCollectionActionsPayload,
  collectionActionsPayload,
  collectionItems,
  hostRefsFieldKey,
  identityRefsFieldKey,
  requireItemByRawText,
} from "./support/entities/mentions";
import { exerciseEntityMerge } from "./support/entities/merge";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow, patchRecord } from "./support/workbook/query";

test("merges duplicate entities from the inspector and preserves survivor identity", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("ENTITY-MERGE"),
    "Record relationships entity-resolution",
  );
  const survivor = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e403-survivor"),
    "host.display_name": "WS-023",
    "host.hostname": "ws-023.corp.example.test",
    "host.aliases": aliasCollectionActionsPayload(["Workstation 23"]),
  });
  const loser = await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e403-loser"),
    "host.display_name": "WS-023 duplicate",
    "host.hostname": "ws-023-dup.corp.example.test",
    "host.aliases": aliasCollectionActionsPayload(["Workstation 23"]),
  });
  const identitySurvivor = await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-identity-survivor"),
      "identity.display_name": "Alex Analyst",
      "identity.email": "alex.analyst@example.test",
      "identity.aliases": aliasCollectionActionsPayload(["Case Owner"]),
    },
  );
  const identityLoser = await createViewRow(
    page,
    incidentId,
    identitiesViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-identity-loser"),
      "identity.display_name": "Alex Analyst duplicate",
      "identity.email": "alex.duplicate@example.test",
      "identity.aliases": aliasCollectionActionsPayload(["Case Owner"]),
    },
  );
  const dependentRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-row"),
      "timeline.activity_synopsis_text": "entity-resolution dependent row",
      [hostRefsFieldKey]: collectionActionsPayload(["Workstation 23"]),
    },
  );
  const dependentMention = requireItemByRawText(
    collectionItems(dependentRow, hostRefsFieldKey),
    "Workstation 23",
  );
  const identityDependentRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e403-identity-row"),
      "timeline.activity_synopsis_text":
        "entity-resolution identity dependent row",
      [identityRefsFieldKey]: collectionActionsPayload(["Case Owner"]),
    },
  );
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
              item_ref: String(dependentMention.item_ref),
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
              item_ref: String(identityDependentMention.item_ref),
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
    mergeReason: "Record relationships entity-resolution duplicate merge",
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
    mergeReason:
      "Record relationships entity-resolution identity duplicate merge",
    rawText: "Case Owner",
    recordType: "identity",
    resolvedLabel: "Alex Analyst",
    survivor: identitySurvivor,
    survivorLabel: "Alex Analyst",
    viewSchemaId: identitiesViewSchemaId,
  });
});
