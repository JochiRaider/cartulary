import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  handoffViewSchemaId,
  lessonViewSchemaId,
  partiesViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";
import type { APIResponse, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createViewRow, type ViewApiRow } from "./support/workbook/query";

type PublicEnvelope<TData> = {
  data: TData;
  error?: never;
  meta?: Record<string, unknown>;
};

type PublicErrorEnvelope = {
  error: {
    code: string;
    details?: Record<string, unknown>;
  };
};

type IncidentMembershipRecord = {
  membership_version: number;
  role: string;
  user_id: string;
};

type CoordinationCase = {
  createValues: Record<string, unknown>;
  label: string;
  patchFieldKey: string;
  patchValue: unknown;
  viewSchemaId: string;
};

const coordinationCases: CoordinationCase[] = [
  {
    createValues: {
      "task.task_kind": "follow_up",
      "task.title": "end-to-end.coordination-review public task",
    },
    label: "task",
    patchFieldKey: "task.priority",
    patchValue: "high",
    viewSchemaId: taskRequestsViewSchemaId,
  },
  {
    createValues: {
      "decision.decision_type": "containment",
      "decision.rationale":
        "end-to-end.coordination-review public decision rationale",
      "decision.summary": "end-to-end.coordination-review public decision",
    },
    label: "decision",
    patchFieldKey: "decision.status",
    patchValue: "approved",
    viewSchemaId: decisionsViewSchemaId,
  },
  {
    createValues: {
      "party.display_name": "end-to-end.coordination-review Public Party",
      "party.party_kind": "team",
    },
    label: "party",
    patchFieldKey: "party.notes",
    patchValue: "end-to-end.coordination-review party route edit",
    viewSchemaId: partiesViewSchemaId,
  },
  {
    createValues: {
      "comm_log.audience": "end-to-end.coordination-review response leadership",
      "comm_log.channel_or_meeting": "end-to-end.coordination-review bridge",
      "comm_log.comm_type": "briefing",
      "comm_log.summary": "end-to-end.coordination-review public communication",
    },
    label: "comm-log",
    patchFieldKey: "comm_log.summary",
    patchValue: "end-to-end.coordination-review edited communication",
    viewSchemaId: commLogViewSchemaId,
  },
  {
    createValues: {
      "handoff.current_state_summary":
        "end-to-end.coordination-review handoff state",
    },
    label: "handoff",
    patchFieldKey: "handoff.next_checks",
    patchValue: "end-to-end.coordination-review verify owner handoff",
    viewSchemaId: handoffViewSchemaId,
  },
  {
    createValues: {
      "status_review.current_state_summary":
        "end-to-end.coordination-review status review state",
    },
    label: "status-review",
    patchFieldKey: "status_review.active_risks_summary",
    patchValue: "end-to-end.coordination-review risk review",
    viewSchemaId: statusReviewViewSchemaId,
  },
  {
    createValues: {
      "lesson.summary": "end-to-end.coordination-review public lesson",
    },
    label: "lesson",
    patchFieldKey: "lesson.closure_state",
    patchValue: "closed",
    viewSchemaId: lessonViewSchemaId,
  },
];

test("Verify coordination rows can be queried and edited through public view/row mutation contracts with current-role authorization.", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOKCOORDINATION"),
    "end-to-end.coordination-review.row-01 public coordination route contracts",
  );
  const memberPassword = "BackupRestoreEditor1!";
  const member = await createIncidentMemberUser(page, incidentId, {
    display_name: "end-to-end.coordination-review editor",
    email: uniqueEmail("end-to-end.coordination-review-editor"),
    initial_password: memberPassword,
    role: "editor",
  });

  const seededRows = new Map<string, ViewApiRow>();
  for (const entry of coordinationCases) {
    const createPayload = {
      client_txn_id: uniqueTxn(`workbook-coordination-create-${entry.label}`),
      ...entry.createValues,
      ...(entry.viewSchemaId === handoffViewSchemaId
        ? { "handoff.incoming_owner_user_id": member.user_id }
        : {}),
    };
    const row = await createViewRow(
      page,
      incidentId,
      entry.viewSchemaId,
      createPayload,
    );
    seededRows.set(entry.viewSchemaId, row);
  }

  const memberContext = await browser.newContext();
  const memberPage = await memberContext.newPage();
  try {
    await sessionTracker.loginTrackedUser(memberPage, {
      createdBy: "end-to-end.coordination-review.row-01",
      email: member.email,
      password: memberPassword,
      purpose: "coordination public-route query and edit",
      userId: member.user_id,
    });

    const patchedRows = new Map<string, ViewApiRow>();
    for (const entry of coordinationCases) {
      const seeded = requiredSeededRow(seededRows, entry.viewSchemaId);
      const queried = await queryPublicRows(
        memberPage,
        incidentId,
        entry.viewSchemaId,
      );
      const queriedRow = requiredQueriedRow(
        queried.rows,
        entry.viewSchemaId,
        seeded.record_id,
      );
      expect(queried.view_schema_id).toBe(entry.viewSchemaId);
      expect(queriedRow.record_id).toBe(seeded.record_id);
      expect(Number.isSafeInteger(queriedRow.row_version)).toBeTruthy();

      const txnId = uniqueTxn(`workbook-coordination-patch-${entry.label}`);
      const patchBody = {
        base_row_version: queriedRow.row_version,
        changes: [
          {
            field_key: entry.patchFieldKey,
            value: entry.patchValue,
          },
        ],
        client_txn_id: txnId,
        view_schema_id: entry.viewSchemaId,
      };
      const patchResponse = await memberPage.request.patch(
        `${apiBase}/api/v1/records/${queriedRow.record_id}`,
        {
          data: patchBody,
          headers: await csrfHeaders(memberPage),
        },
      );
      expect(patchResponse.ok()).toBeTruthy();
      const patched = await readSuccessEnvelope<{
        change_set_id: string;
        row: ViewApiRow;
      }>(patchResponse);
      expect(typeof patched.change_set_id).toBe("string");
      expect(patched.row.record_id).toBe(queriedRow.record_id);
      expect(patched.row.row_version).toBeGreaterThan(queriedRow.row_version);
      expect(patched.row.cells[entry.patchFieldKey]?.value).toEqual(
        entry.patchValue,
      );
      patchedRows.set(entry.viewSchemaId, patched.row);

      const replayed = await queryPublicRows(
        memberPage,
        incidentId,
        entry.viewSchemaId,
      );
      const replayedRow = requiredQueriedRow(
        replayed.rows,
        entry.viewSchemaId,
        queriedRow.record_id,
      );
      expect(replayedRow.row_version).toBe(patched.row.row_version);
      expect(replayedRow.cells[entry.patchFieldKey]?.value).toEqual(
        entry.patchValue,
      );
    }

    const memberMembership = await loadIncidentMembership(
      page,
      incidentId,
      member.user_id,
    );
    await patchIncidentMembershipRole(page, incidentId, {
      baseMembershipVersion: memberMembership.membership_version,
      role: "viewer",
      userId: member.user_id,
    });

    const taskCase = coordinationCases[0];
    if (taskCase === undefined) {
      throw new Error("missing task route case");
    }
    const currentTask = requiredSeededRow(patchedRows, taskCase.viewSchemaId);
    const viewerQuery = await queryPublicRows(
      memberPage,
      incidentId,
      taskCase.viewSchemaId,
    );
    expect(
      viewerQuery.rows.some((row) => row.record_id === currentTask.record_id),
    ).toBeTruthy();

    const deniedPatchBody = {
      base_row_version: currentTask.row_version,
      changes: [
        {
          field_key: taskCase.patchFieldKey,
          value: "urgent",
        },
      ],
      client_txn_id: uniqueTxn("workbook-coordination-denied-current-role"),
      view_schema_id: taskCase.viewSchemaId,
    };
    const denied = await memberPage.request.patch(
      `${apiBase}/api/v1/records/${currentTask.record_id}`,
      {
        data: deniedPatchBody,
        headers: await csrfHeaders(memberPage),
      },
    );
    expect(denied.status()).toBe(403);
    const deniedEnvelope = (await denied.json()) as PublicErrorEnvelope;
    expect(deniedEnvelope.error.code).toBe("authorization_denied");
    expect(deniedEnvelope.error.details?.required_role).toBe(
      "editor|reviewer|admin",
    );

    const afterDenied = await queryPublicRows(
      page,
      incidentId,
      taskCase.viewSchemaId,
    );
    const unchangedTask = requiredQueriedRow(
      afterDenied.rows,
      taskCase.viewSchemaId,
      currentTask.record_id,
    );
    expect(unchangedTask.row_version).toBe(currentTask.row_version);
    expect(unchangedTask.cells[taskCase.patchFieldKey]?.value).toBe(
      taskCase.patchValue,
    );
  } finally {
    await memberContext.close();
  }
});

async function queryPublicRows(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
) {
  const response = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
    { data: {} },
  );
  expect(response.ok()).toBeTruthy();
  return readSuccessEnvelope<{
    incident_id: string;
    rows: ViewApiRow[];
    view_schema_id: string;
  }>(response);
}

async function readSuccessEnvelope<TData>(
  response: APIResponse,
): Promise<TData> {
  const envelope = (await response.json()) as PublicEnvelope<TData>;
  return envelope.data;
}

function requiredSeededRow(
  rows: Map<string, ViewApiRow>,
  viewSchemaId: string,
) {
  const row = rows.get(viewSchemaId) ?? null;
  if (row === null) {
    throw new Error(`missing seeded row for ${viewSchemaId}`);
  }
  return row;
}

function requiredQueriedRow(
  rows: ViewApiRow[],
  viewSchemaId: string,
  recordId: string,
) {
  const row =
    rows.find((candidate) => candidate.record_id === recordId) ?? null;
  if (row === null) {
    throw new Error(`missing ${recordId} in public ${viewSchemaId} query`);
  }
  return row;
}

async function loadIncidentMembership(
  page: Page,
  incidentId: string,
  userId: string,
) {
  const response = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
    { headers: await csrfHeaders(page) },
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { memberships: IncidentMembershipRecord[] };
  };
  const membership =
    body.data.memberships.find((candidate) => candidate.user_id === userId) ??
    null;
  if (membership === null) {
    throw new Error(`missing incident membership for ${userId}`);
  }
  return membership;
}

async function patchIncidentMembershipRole(
  page: Page,
  incidentId: string,
  options: {
    baseMembershipVersion: number;
    role: string;
    userId: string;
  },
) {
  const response = await page.request.patch(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${options.userId}`,
    {
      data: {
        base_membership_version: options.baseMembershipVersion,
        role: options.role,
      },
      headers: await csrfHeaders(page),
    },
  );
  expect(response.ok()).toBeTruthy();
}
