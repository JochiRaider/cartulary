import type {
  CreateViewRowResponse,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import {
  authTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridRowTestId,
  gridRowVersionAttribute,
  gridShellTestId,
  incidentLandingTestId,
  surfaceTabTestId,
  timelineInspectorTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  handoffViewSchemaId,
  lessonViewSchemaId,
  requireViewContract,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { collectionItems } from "./support/entities/mentions";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { responseBarrier } from "./support/runtime/cleanup";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import {
  awaitResponseBarrier,
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import { disableWorkbookSockets } from "./support/transport/socketIsolation";
import { setGenericCreateField } from "./support/workbook/authoring";
import {
  coordinationMatrix,
  fillCoordinationMinimum,
  openCoordinationFixture,
  retainCoordinationUncertainResult,
} from "./support/workbook/coordinationCreate";
import { fetchFullRecordHistory } from "./support/workbook/history";
import {
  createViewRow,
  queryViewRows,
  readWorkbookMutation,
  waitForViewRowByCell,
} from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { activateCandidateIdentities } from "./support/workbook/references";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test("All twelve contextual coordination actions create one owner-defined source link without semantic seeds", async ({
  page,
}) => {
  for (const [view, variants] of Object.entries(coordinationMatrix))
    for (const variant of variants) {
      const f = await openCoordinationFixture(page, variant, view);
      const before = await fetchFullRecordHistory(page, f.source.record_id);
      await f.form
        .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
        .press("Enter");
      await expect(f.form.getByRole("alert").first()).toBeVisible();
      await fillCoordinationMinimum(f);
      const response = page.waitForResponse(
        (r) =>
          r.request().method() === "POST" &&
          r.url().endsWith(`/views/${f.target.viewSchemaId}/rows`),
      );
      await f.form
        .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
        .press("Enter");
      const accepted = await response;
      expect(accepted.ok()).toBe(true);
      const receipt = (await accepted.json()) as CreateViewRowResponse;
      expect(receipt.data).toMatchObject({
        view_schema_id: f.target.viewSchemaId,
        source_record_id: f.source.record_id,
        link_type: "references_artifact",
      });
      expect(receipt.meta.request_id).toBe(accepted.headers()["x-request-id"]);
      await expect(f.form).toHaveCount(0);
      await expect(f.action).toBeFocused();
      const rows = await queryViewRows(page, f.incident, f.target.viewSchemaId);
      expect(rows).toHaveLength(1);
      expect(rows[0]?.record_id).toBe(receipt.data.row.record_id);
      for (const field of f.target.fields.filter(
        (field) => field.readKind === "collection",
      )) {
        const cell = rows[0]?.cells[field.fieldKey];
        expect(cell?.value).toMatchObject({
          kind: "collection_value_v1",
          items: [],
        });
      }
      const after = await fetchFullRecordHistory(page, f.source.record_id);
      const added = after.items.filter(
        (item) =>
          !before.items.some(
            (prior) => prior.history_item_ref === item.history_item_ref,
          ),
      );
      expect(added.length).toBeGreaterThan(0);
      expect(new Set(added.map((item) => item.change_set_id))).toEqual(
        new Set([receipt.data.change_set_id]),
      );
      const sources = await queryViewRows(page, f.incident, view);
      expect(
        sources.find((row) => row.record_id === f.source.record_id)
          ?.row_version,
      ).toBe(f.source.row_version);
    }
});

test("Coordination drafts survive navigation and source replacement or clear preserves every authored value", async ({
  page,
}) => {
  const f = await openCoordinationFixture(
    page,
    "lesson",
    taskRequestsViewSchemaId,
  );
  const replacement = await createViewRow(
    page,
    f.incident,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("replacement"),
      "timeline.activity_synopsis_text": "Chosen coordination source",
    },
  );
  await fillCoordinationMinimum(f);
  await f.form
    .getByRole("button", { name: "Choose source", exact: true })
    .click();
  const picker = f.form.getByRole("region", {
    name: "Choose source",
    exact: true,
  });
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(timelineViewSchemaId);
  await expect(
    picker.getByRole("radio", {
      name: `Chosen coordination source (${replacement.record_id})`,
      exact: true,
    }),
  ).toBeAttached();
  await picker
    .getByRole("group", { name: "Source", exact: true })
    .locator(`input[type="radio"][value="${replacement.record_id}"]`)
    .check();
  await picker.press("Escape");
  await expect(
    f.form.getByRole("button", { name: "Choose source", exact: true }),
  ).toBeFocused();
  await f.form
    .getByRole("button", { name: "Choose source", exact: true })
    .click();
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(timelineViewSchemaId);
  await expect(
    picker.getByRole("radio", {
      name: `Chosen coordination source (${replacement.record_id})`,
      exact: true,
    }),
  ).toBeAttached();
  await picker
    .getByRole("group", { name: "Source", exact: true })
    .locator(`input[type="radio"][value="${replacement.record_id}"]`)
    .check();
  await picker
    .getByRole("button", { name: "Apply references", exact: true })
    .click();
  await expect(f.form).toContainText("Chosen coordination source");
  await page.getByTestId(workbookInspectorCloseButtonTestId(f.view)).click();
  await switchSheet(page, statusReviewViewSchemaId);
  await openRecoveryItem(page, /^Coordination draft ·/);
  const recovery = page.getByRole("region", {
    name: "Retained Coordination authoring",
    exact: true,
  });
  await recovery
    .getByRole("button", { name: "Resume Coordination draft", exact: true })
    .click();
  await expect(
    recovery.getByTestId(genericCreateFieldTestId("lesson.summary")),
  ).toHaveValue("  Authored coordination summary  ");
  await expect(recovery).toContainText("Chosen coordination source");
  await recovery
    .getByRole("button", { name: "Clear source", exact: true })
    .click();
  await expect(recovery).toContainText("No source link will be saved.");
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      r.url().endsWith(`/views/${lessonViewSchemaId}/rows`),
  );
  await recovery
    .getByTestId(genericCreateSubmitTestId(lessonViewSchemaId))
    .click();
  const receipt = (await (await response).json()) as CreateViewRowResponse;
  expect(receipt.data).not.toHaveProperty("source_record_id");
  expect(receipt.data.row.cells["lesson.summary"]?.value).toBe(
    "Authored coordination summary",
  );
});

test("Coordination response loss after server commit recovers the exact request without another artifact", async ({
  page,
}) => {
  const f = await openCoordinationFixture(page);
  await fillCoordinationMinimum(f);
  const bodies: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "POST" &&
      request.url().endsWith(`/views/${f.target.viewSchemaId}/rows`)
    )
      bodies.push(request.postData() ?? "");
  });
  const { path, recovery } = await retainCoordinationUncertainResult(page, f);
  expect(
    await queryViewRows(page, f.incident, f.target.viewSchemaId),
  ).toHaveLength(1);
  await removeOwnedRoute(page, path);
  const response = page.waitForResponse(
    (r) =>
      r.request().method() === "POST" &&
      r.url().endsWith(`/views/${f.target.viewSchemaId}/rows`),
  );
  await recovery
    .getByRole("button", { name: "Recover submission", exact: true })
    .press("Enter");
  const replay = await response;
  expect(replay.status()).toBe(200);
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  expect(
    await queryViewRows(page, f.incident, f.target.viewSchemaId),
  ).toHaveLength(1);
  await expect(recovery).not.toBeVisible();
  await expect(
    page.getByRole("button", { name: /^Refresh saved view ·/ }),
  ).toBeVisible();
});

test("Accepted coordination refresh recovery sends reads only and keeps the source selection", async ({
  page,
}) => {
  const f = await openCoordinationFixture(page);
  await fillCoordinationMinimum(f);
  let committed = false,
    creates = 0;
  const query = `**/incidents/${f.incident}/views/${f.target.viewSchemaId}/query`;
  await installOwnedRoute(page, query, async (route) => {
    if (committed) await route.abort("failed");
    else await route.continue();
  });
  await installOwnedRoute(
    page,
    `**/incidents/${f.incident}/views/${f.target.viewSchemaId}/rows`,
    async (route) => {
      creates++;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await route.fulfill({ response });
    },
  );
  const createdResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/views/${f.target.viewSchemaId}/rows`),
  );
  await f.form
    .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
    .click();
  // Source validation may still be reading after the click. Activating
  // Recovery detaches that form, so wait for accepted creation first.
  expect((await createdResponse).ok()).toBe(true);
  await openRecoveryItem(page, /^Coordination creation ·/);
  const recovery = page.getByRole("region", {
    name: "Retained Coordination authoring",
    exact: true,
  });
  await expect(recovery).toContainText("views need refresh");
  await removeOwnedRoute(page, query);
  await recovery
    .getByRole("button", { name: "Retry refresh", exact: true })
    .press("Enter");
  await expect(recovery).not.toBeVisible();
  await expect(
    page.getByRole("button", { name: /^Refresh saved view ·/ }),
  ).toBeVisible();
  expect(creates).toBe(1);
  await expect(page.getByTestId(gridShellTestId(f.view))).toBeVisible();
  expect(
    await queryViewRows(page, f.incident, f.target.viewSchemaId),
  ).toHaveLength(1);
});

test("Incident revocation conceals coordination drafts and late acceptance while preserving the account session", async ({
  page,
  sessionTracker,
  testActorRequest,
}) => {
  let memberId = "";
  const f = await openCoordinationFixture(
    page,
    "lesson",
    taskRequestsViewSchemaId,
    async (url) => {
      const incident = new URL(url, "http://fixture").searchParams.get(
        "incident_id",
      );
      if (!incident) throw new Error("Missing incident");
      const member = await createIncidentMemberUser(page, incident, {
        email: uniqueEmail("coordination-revoked"),
        display_name: "Coordination author",
        initial_password: "CoordinationMember1!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      });
      memberId = member.user_id;
      await sessionTracker.loginTrackedUser(page, {
        createdBy: "coordination-revocation",
        email: member.email,
        password: member.initial_password,
        purpose: "retained coordination revocation",
        userId: memberId,
      });
      const sockets = installIncidentSocketMonitor(page, incident);
      await page.goto(url);
      await sockets.waitForAcceptedSocket();
    },
  );
  await fillCoordinationMinimum(f);
  let committed = false;
  let release: () => void = () => {};
  const delayedBarrier = responseBarrier(page);
  release = delayedBarrier.release;
  const delayed = delayedBarrier.promise;
  const suffix = `/views/${f.target.viewSchemaId}/rows`;
  await installOwnedRoute(
    page,
    `**/incidents/${f.incident}${suffix}`,
    async (route) => {
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      if (!(await awaitResponseBarrier(route, delayed))) return;
      await route.fulfill({ response });
    },
  );
  await f.form
    .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
    .click();
  await expect.poll(() => committed).toBe(true);
  const response = await testActorRequest.get(
    `/api/v1/incidents/${f.incident}/memberships`,
  );
  const members = (await response.json()) as {
    data: { memberships: { user_id: string; membership_version: number }[] };
  };
  const member = members.data.memberships.find(
    (item) => item.user_id === memberId,
  );
  if (!member) throw new Error("Missing membership");
  const revoked = await testActorRequest.delete(
    `/api/v1/incidents/${f.incident}/memberships/${memberId}`,
    { data: { base_membership_version: member.membership_version } },
  );
  expect(revoked.status()).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(f.form).toHaveCount(0);
  const late = page.waitForResponse((r) => r.url().endsWith(suffix));
  release();
  await late;
  await expect(page.getByTestId(authTestId("shell"))).toHaveCount(0);
  await expect(
    page.getByRole("region", {
      name: "Retained Coordination authoring",
      exact: true,
    }),
  ).toHaveCount(0);
  await expect(page).not.toHaveURL(/incident_id=/u);
});

async function switchSheet(page: Page, view: string) {
  const tab = page.getByTestId(surfaceTabTestId(view));
  if (await tab.count()) await tab.click();
  else {
    await page.getByRole("button", { name: "More views", exact: true }).click();
    await page
      .locator(`[role="menuitemradio"][data-view-schema-id="${view}"]`)
      .click();
  }
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
}

async function expectTimelineWorkflowShell(page: Page) {
  await expect(page).toHaveURL(
    new RegExp(`view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`),
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
  await expect(
    page.getByTestId(
      workbookInspectorPanelTestId(timelineViewSchemaId, "workflow"),
    ),
  ).toBeVisible();
}

async function createFromTimelineWorkflow(
  page: Page,
  incidentId: string,
  options: TimelineWorkflowCreateOptions,
) {
  await expectTimelineWorkflowShell(page);
  const action = page.getByTestId(
    workbookInspectorFeatureActionTestId(
      timelineViewSchemaId,
      options.actionKey,
    ),
  );
  await expect(action).toBeVisible();
  await action.click();
  await expect(
    page.getByTestId(genericCreateSubmitTestId(options.targetViewSchemaId)),
  ).toBeVisible();
  await expect(action).toHaveCount(1);
  const commandRow = page.locator(
    `[data-inspector-feature="${options.actionKey}"]`,
  );
  const authoring = commandRow.locator(
    `[data-inspector-feature-content="${options.actionKey}"]`,
  );
  await expect(
    authoring.getByTestId(
      genericCreateSubmitTestId(options.targetViewSchemaId),
    ),
  ).toBeVisible();
  await test.info().attach(`inspector-workflow-${options.actionKey}`, {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  for (const [fieldKey, value] of options.fields) {
    const field = requireViewContract(options.targetViewSchemaId).fieldMap[
      fieldKey
    ];
    const coordinationReference =
      [
        commLogViewSchemaId,
        handoffViewSchemaId,
        statusReviewViewSchemaId,
        lessonViewSchemaId,
      ].some((view) => view === options.targetViewSchemaId) &&
      field?.directReferenceContractId;
    if (coordinationReference) {
      await page
        .getByRole("button", {
          name: `Choose ${field.label.toLowerCase()}`,
          exact: true,
        })
        .click();
      await activateCandidateIdentities(
        page.getByTestId(genericCreateFieldTestId(fieldKey)),
        value,
      );
      await page
        .getByRole("button", { name: "Apply references", exact: true })
        .click();
    } else await setGenericCreateField(page, fieldKey, value);
  }
  await page
    .getByTestId(genericCreateSubmitTestId(options.targetViewSchemaId))
    .click();
  const created = await waitForViewRowByCell(
    page,
    incidentId,
    options.targetViewSchemaId,
    options.waitFieldKey,
    options.waitValue,
  );
  await expectTimelineWorkflowShell(page);
  return created;
}

type TimelineWorkflowCreateOptions = {
  actionKey: string;
  fields: ReadonlyArray<readonly [string, string]>;
  targetViewSchemaId: string;
  waitFieldKey: string;
  waitValue: string;
};

function expectCollectionReferencesRecord(
  row: ViewRow,
  fieldKey: string,
  recordId: string,
) {
  expect(JSON.stringify(collectionItems(row, fieldKey))).toContain(recordId);
}

test("Verify Timeline inspector Workflow create-related actions stay in the workbook shell.", async ({
  page,
  testActor,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("WORKBOOKWORKFLOW"),
    "end-to-end.inspector-history.row-03 Timeline inspector create-related workflows",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("workbook-workflow-source"),
    "timeline.activity_synopsis_text":
      "end-to-end.inspector-history.row-03 create-related source",
    "timeline.raw_activity_text": "Inspector Workflow create-related source.",
  });

  await disableWorkbookSockets(page);
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      timelineViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await openTimelineInspector(page, source.record_id as string);
  await expectTimelineWorkflowShell(page);

  const task = await createFromTimelineWorkflow(page, incidentId, {
    actionKey: "create_related.task_request",
    targetViewSchemaId: taskRequestsViewSchemaId,
    fields: [
      ["task.title", "end-to-end.inspector-history.row-03 workflow task"],
      ["task.task_kind", "follow_up"],
    ],
    waitFieldKey: "task.title",
    waitValue: "end-to-end.inspector-history.row-03 workflow task",
  });
  expectCollectionReferencesRecord(
    task,
    "task.linked_record_ids",
    source.record_id as string,
  );

  const decision = await createFromTimelineWorkflow(page, incidentId, {
    actionKey: "create_related.decision",
    targetViewSchemaId: decisionsViewSchemaId,
    fields: [
      [
        "decision.summary",
        "end-to-end.inspector-history.row-03 workflow decision",
      ],
      ["decision.decision_type", "containment"],
      ["decision.rationale", "Decision created from the Timeline inspector."],
    ],
    waitFieldKey: "decision.summary",
    waitValue: "end-to-end.inspector-history.row-03 workflow decision",
  });
  expectCollectionReferencesRecord(
    decision,
    "decision.support_refs",
    source.record_id as string,
  );

  const evidenceLinkResponse = page.waitForResponse((response) => {
    const request = response.request();
    if (
      request.method() !== "PATCH" ||
      !request.url().endsWith(`/api/v1/records/${source.record_id}`)
    ) {
      return false;
    }
    const body = request.postDataJSON() as {
      changes?: Array<{ field_key?: string }>;
    };
    return (
      body.changes?.some(
        (change) => change.field_key === "timeline.attached_evidence_ids",
      ) ?? false
    );
  });
  const evidence = await createFromTimelineWorkflow(page, incidentId, {
    actionKey: "create_related.evidence",
    targetViewSchemaId: evidenceViewSchemaId,
    fields: [
      [
        "evidence.title",
        "end-to-end.inspector-history.row-03 workflow evidence",
      ],
    ],
    waitFieldKey: "evidence.title",
    waitValue: "end-to-end.inspector-history.row-03 workflow evidence",
  });
  const evidenceLink = await evidenceLinkResponse;
  expect(evidenceLink.ok()).toBeTruthy();
  const linkedTimeline = await readWorkbookMutation(
    evidenceLink,
    "patchRecord",
  );
  expect(evidenceLink.request().postDataJSON()).toMatchObject({
    view_schema_id: timelineViewSchemaId,
    changes: [
      {
        field_key: "timeline.attached_evidence_ids",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "add_record_ref",
              linked_record_id: evidence.record_id,
            },
          ],
        },
      },
    ],
  });
  await openRecoveryItem(page, /^Timeline Evidence creation ·/);
  await expect(
    page.getByRole("region", {
      name: "Retained Timeline Evidence creation",
      exact: true,
    }),
  ).toContainText(
    "Evidence created and linked to the original Timeline record.",
  );
  await page
    .getByRole("button", { name: "Close recovery", exact: true })
    .click();

  // The evidence link advances the source version and invalidates old authoring.
  // Open the next workflow only after that committed version reaches the grid.
  await expect(
    page.getByTestId(
      gridRowTestId(timelineViewSchemaId, source.record_id as string),
    ),
  ).toHaveAttribute(
    gridRowVersionAttribute,
    String(linkedTimeline.data.row.row_version),
  );
  await openTimelineInspector(page, source.record_id as string);
  const comm = await createFromTimelineWorkflow(page, incidentId, {
    actionKey: "create_related.comm_log",
    targetViewSchemaId: commLogViewSchemaId,
    fields: [
      ["comm_log.comm_type", "briefing"],
      ["comm_log.audience", "end-to-end.inspector-history.row-03 responders"],
      ["comm_log.channel_or_meeting", "Timeline inspector"],
      [
        "comm_log.summary",
        "end-to-end.inspector-history.row-03 workflow communication",
      ],
    ],
    waitFieldKey: "comm_log.summary",
    waitValue: "end-to-end.inspector-history.row-03 workflow communication",
  });
  expect(comm.cells["comm_log.comm_type"]?.value).toBe("briefing");

  const handoff = await createFromTimelineWorkflow(page, incidentId, {
    actionKey: "create_related.handoff",
    targetViewSchemaId: handoffViewSchemaId,
    fields: [
      ["handoff.incoming_owner_user_id", testActor.user_id],
      [
        "handoff.current_state_summary",
        "end-to-end.inspector-history.row-03 workflow handoff",
      ],
    ],
    waitFieldKey: "handoff.current_state_summary",
    waitValue: "end-to-end.inspector-history.row-03 workflow handoff",
  });
  expect(handoff.cells["handoff.incoming_owner_user_id"]?.value).toBe(
    testActor.user_id,
  );

  const statusReview = await createFromTimelineWorkflow(page, incidentId, {
    actionKey: "create_related.status_review",
    targetViewSchemaId: statusReviewViewSchemaId,
    fields: [
      [
        "status_review.current_state_summary",
        "end-to-end.inspector-history.row-03 workflow status review",
      ],
    ],
    waitFieldKey: "status_review.current_state_summary",
    waitValue: "end-to-end.inspector-history.row-03 workflow status review",
  });
  expect(statusReview.cells["status_review.current_state_summary"]?.value).toBe(
    "end-to-end.inspector-history.row-03 workflow status review",
  );

  const lesson = await createFromTimelineWorkflow(page, incidentId, {
    actionKey: "create_related.lesson",
    targetViewSchemaId: lessonViewSchemaId,
    fields: [
      ["lesson.summary", "end-to-end.inspector-history.row-03 workflow lesson"],
    ],
    waitFieldKey: "lesson.summary",
    waitValue: "end-to-end.inspector-history.row-03 workflow lesson",
  });
  expect(lesson.cells["lesson.summary"]?.value).toBe(
    "end-to-end.inspector-history.row-03 workflow lesson",
  );
});
