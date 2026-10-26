import type {
  CreateViewRowRequest,
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts/http";

import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, Route } from "@playwright/test";

import { createIncident } from "../incidents/fixtures";

import { uniqueIncidentKey, uniqueTxn } from "../runtime/fixtureIdentity";

import { installOwnedRoute } from "../transport/requestInterception";

import { createViewRow } from "./query";

export async function observeQuery(page: Page, incident: string, view: string) {
  const reads: {
    request: QueryWorkbookViewRequest;
    response: QueryWorkbookViewResponse;
  }[] = [];
  const path = `**/incidents/${incident}/views/${view}/query`;
  const read = async (route: Route) => {
    const response = await route.fetch();
    if (response.ok())
      reads.push({
        request: route.request().postDataJSON() as QueryWorkbookViewRequest,
        response: (await response.json()) as QueryWorkbookViewResponse,
      });
    await route.fulfill({ response });
  };
  await installOwnedRoute(page, path, read);
  return reads;
}

export async function timestampTaskFixture(page: Page) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TIMESTAMP"),
    "Timestamp range membership",
  );
  const values = [
    "2026-04-17T23:59:59.999Z",
    "2026-04-18T00:00:00Z",
    "2026-04-18T00:00:00.05Z",
    "2026-04-18T00:00:00.1Z",
    "2026-04-18T00:00:00.101Z",
  ];
  const rows = [];
  for (const value of values)
    rows.push(
      await createViewRow(page, incident, taskRequestsViewSchemaId, {
        client_txn_id: uniqueTxn("timestamp-task"),
        "task.title": `Due ${value}`,
        "task.task_kind": "question",
        "task.due_at": value,
      }),
    );
  return { incident, rows };
}

export const fixtureFields: Readonly<
  Record<
    string,
    {
      readonly field: string;
      readonly minimum: Readonly<Record<string, unknown>>;
    }
  >
> = {
  [timelineViewSchemaId]: {
    field: "timeline.activity_synopsis_text",
    minimum: {},
  },
  [hostsViewSchemaId]: { field: "host.display_name", minimum: {} },
  [identitiesViewSchemaId]: { field: "identity.display_name", minimum: {} },
  [notesViewSchemaId]: {
    field: "note.title",
    minimum: {
      "note.body": "Workbook browsing fixture",
      "note.tags": {
        kind: "collection_actions_v1",
        actions: [
          { op: "add_tag", tag_name: "alpha" },
          { op: "add_tag", tag_name: "beta" },
        ],
      },
    },
  },
  [evidenceViewSchemaId]: {
    field: "evidence.title",
    minimum: { "evidence.collector_party_text": "BRIDGE Operations" },
  },
  [indicatorsViewSchemaId]: {
    field: "indicator.display_value",
    minimum: {
      "indicator.indicator_type": "domain_name",
      "indicator.value_kind": "atomic",
    },
  },
  [partiesViewSchemaId]: {
    field: "party.display_name",
    minimum: { "party.party_kind": "person" },
  },
  [taskRequestsViewSchemaId]: {
    field: "task.title",
    minimum: { "task.task_kind": "question" },
  },
  [decisionsViewSchemaId]: {
    field: "decision.summary",
    minimum: {
      "decision.decision_type": "scope",
      "decision.rationale": "Query browsing evidence",
    },
  },
  [commLogViewSchemaId]: {
    field: "comm_log.summary",
    minimum: {
      "comm_log.comm_type": "briefing",
      "comm_log.audience": "Operations",
      "comm_log.channel_or_meeting": "Bridge",
    },
  },
  [handoffViewSchemaId]: {
    field: "handoff.current_state_summary",
    minimum: {},
  },
  [statusReviewViewSchemaId]: {
    field: "status_review.current_state_summary",
    minimum: {},
  },
  [lessonViewSchemaId]: { field: "lesson.summary", minimum: {} },
  [findingsViewSchemaId]: { field: "finding.statement", minimum: {} },
  [investigativeQueriesViewSchemaId]: {
    field: "investigative_query.purpose",
    minimum: {
      "investigative_query.platform": "SQL",
      "investigative_query.query_text": "select value",
    },
  },
  [forensicKeywordsViewSchemaId]: {
    field: "forensic_keyword.pattern",
    minimum: { "forensic_keyword.reason": "Query browsing evidence" },
  },
  [assessmentsViewSchemaId]: {
    field: "assessment.rationale",
    minimum: {
      "assessment.subject_type": "host",
      "assessment.assessment_state": "confirmed",
      "assessment.confidence_score": 85,
      "assessment.assessed_at": "2026-09-10T00:00:00Z",
    },
  },
};

export async function seed(
  page: Page,
  incident: string,
  view: string,
  count: number,
  actorId: string,
  afterCreated?: (recordId: string) => Promise<unknown>,
) {
  const fixture = fixtureFields[view];
  if (!fixture)
    throw new Error(`Uncovered registered workbook surface: ${view}`);
  const subject =
    view === assessmentsViewSchemaId
      ? await createViewRow(page, incident, hostsViewSchemaId, {
          client_txn_id: uniqueTxn("wqc-subject"),
          "host.hostname": "subject.example.test",
        })
      : null;
  let next = 0;
  // Bounded fixture creation; these isolated writes are not the application's browsing path.
  await Promise.all(
    Array.from({ length: 8 }, async () => {
      for (;;) {
        const index = next++;
        if (index >= count) return;
        const ordinal = String(index).padStart(5, "0");
        const payload: CreateViewRowRequest = {
          client_txn_id: uniqueTxn(`wqc-${index}`),
          ...fixture.minimum,
          [fixture.field]: view.includes("indicators")
            ? `wqc-${ordinal}.example.test`
            : `WQC ${ordinal}`,
          ...(view === hostsViewSchemaId
            ? { "host.hostname": `wqc-${ordinal}.example.test` }
            : {}),
          ...(view === identitiesViewSchemaId
            ? {
                "identity.sid": `S-1-5-21-123456789-123456789-123456789-${1000 + index}`,
              }
            : {}),
          ...(view.includes("handoff")
            ? { "handoff.incoming_owner_user_id": actorId }
            : {}),
          ...(subject ? { "assessment.subject_ref": subject.record_id } : {}),
        };
        const row = await createViewRow(page, incident, view, payload);
        await afterCreated?.(row.record_id);
      }
    }),
  );
  return fixture;
}
