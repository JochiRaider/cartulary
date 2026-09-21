import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import {
  collectionActionsPayload,
  hostRefsFieldKey,
} from "../entities/mentions";
import { createIncident } from "../incidents/fixtures";
import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";
import { createViewRow, patchRecord } from "../workbook/query";

/** The same sparse production fixture serves visual capture and reading geometry. */
export async function createInspectorReadingFixture(page: Page) {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALINSPECTORHISTORY"),
    "browser.inspector-history visual inspector actions",
  );
  const evidence = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("VISUALINSPECTORHISTORY-EVIDENCE"),
    "evidence.collector_party_text":
      "browser.inspector-history visual collector",
    "evidence.title": "browser.inspector-history visual attached evidence",
  });
  const target = await createViewRow(page, incidentId, timelineViewSchemaId, {
    [hostRefsFieldKey]: collectionActionsPayload([
      "browser.inspector-history visual host",
    ]),
    client_txn_id: uniqueTxn("VISUALINSPECTORHISTORY-TARGET"),
    "timeline.raw_activity_text":
      "browser.inspector-history visual inspector details",
    "timeline.activity_synopsis_text":
      "browser.inspector-history visual inspector target",
  });
  const linkedTarget = await patchRecord(page, target.record_id, {
    base_row_version: target.row_version,
    changes: [
      {
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            { op: "add_record_ref", linked_record_id: evidence.record_id },
          ],
        },
        field_key: "timeline.attached_evidence_ids",
      },
    ],
    client_txn_id: uniqueTxn("VISUALINSPECTORHISTORY-LINK"),
    view_schema_id: timelineViewSchemaId,
  });
  return { incidentId, evidence, target, linkedTarget };
}
