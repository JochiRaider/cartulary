import type {
  CreateViewRowRequest,
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts/http";
import {
  applyFilterChip,
  changeGrouping,
  scrollGridCellIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  authTestId,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridGroupRowsSelector,
  gridGroupRowTestId,
  gridRowTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  incidentLandingTestId,
  rowCellTestId,
  savedViewModifiedTestId,
  workbookFilterOperatorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookQueryEntryTestId,
  workbookSortMenuTriggerTestId,
} from "@cartulary/ui-contracts";
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
  listWorkbookSurfaceContracts,
  notesViewSchemaId,
  partiesViewSchemaId,
  requireViewContract,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { readCurrentSession } from "./support/auth/sessions";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { TestClock } from "./support/runtime/testClock";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import { createViewRow, patchRecord } from "./support/workbook/query";
import {
  activateCommittedGridCell,
  openGenericInspectorForRecord,
  openTimelineInspector,
} from "./support/workbook/rowMutations";
import {
  createSavedView,
  createSavedViewFromCurrentSurface,
  selectSavedView,
  setSavedViewDraftName,
} from "./support/workbook/savedViews";

const schemas = [
  timelineViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  evidenceViewSchemaId,
  indicatorsViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  decisionsViewSchemaId,
  commLogViewSchemaId,
  handoffViewSchemaId,
  statusReviewViewSchemaId,
  lessonViewSchemaId,
  findingsViewSchemaId,
  investigativeQueriesViewSchemaId,
  forensicKeywordsViewSchemaId,
  assessmentsViewSchemaId,
];
const fixtureFields: Readonly<
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

async function seed(
  page: Page,
  incident: string,
  view: string,
  count: number,
  actorId: string,
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
        await createViewRow(page, incident, view, payload);
      }
    }),
  );
  return fixture;
}

async function observeQuery(page: Page, incident: string, view: string) {
  const reads: {
    request: QueryWorkbookViewRequest;
    response: QueryWorkbookViewResponse;
  }[] = [];
  await page.route(
    `**/incidents/${incident}/views/${view}/query`,
    async (route) => {
      const response = await route.fetch();
      if (response.ok())
        reads.push({
          request: route.request().postDataJSON() as QueryWorkbookViewRequest,
          response: (await response.json()) as QueryWorkbookViewResponse,
        });
      await route.fulfill({ response });
    },
  );
  return reads;
}

async function exerciseSurface(page: Page, view: string, actorId: string) {
  test.setTimeout(180_000);
  const pageErrors: string[] = [];
  page.on("pageerror", (error) => pageErrors.push(error.message));
  expect([...schemas].sort()).toEqual(
    listWorkbookSurfaceContracts()
      .map((entry) => entry.viewSchemaId)
      .sort(),
  );
  const count = [
    timelineViewSchemaId,
    hostsViewSchemaId,
    identitiesViewSchemaId,
    assessmentsViewSchemaId,
    notesViewSchemaId,
  ].some((candidate) => candidate === view)
    ? 405
    : 105;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC"),
    `Workbook continuation ${view}`,
  );
  const fixture = await seed(page, incident, view, count, actorId);
  const reads = await observeQuery(page, incident, view);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText("100 records loaded; more available.");
  expect(reads.length).toBeLessThanOrEqual(3);
  const first = reads.at(-1);
  if (!first) throw new Error("Missing first production response");
  expect(first.request.cursor_token).toBeUndefined();
  expect(first.request.limit).toBe(100);
  if (view === timelineViewSchemaId) {
    await page
      .getByRole("checkbox", { name: "Select all loaded records", exact: true })
      .check();
    await expect(
      page.getByText("100 records selected.", { exact: true }),
    ).toBeVisible();
  }
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  await more.focus();
  await more.press("Enter");
  await expect(controls).toContainText(
    `${Math.min(200, count)} records loaded`,
  );
  await expect(more).toBeFocused();
  if (view === timelineViewSchemaId)
    await expect(
      page.getByText("100 records selected.", { exact: true }),
    ).toBeVisible();
  const later = reads.at(-1);
  if (!later) throw new Error("Missing continuation response");
  const producing = reads
    .slice(0, reads.indexOf(later))
    .reverse()
    .find(
      (read) =>
        read.response.meta.paging.next_cursor === later.request.cursor_token,
    );
  expect(
    producing,
    "Continuation preserves a producing response's cursor bytes",
  ).toBeDefined();
  expect(later.response.meta.query).toEqual(producing?.response.meta.query);
  expect(later.request.sort).toBeUndefined();
  expect(later.request.group_by).toBeUndefined();
  const row = later.response.data.rows[0];
  if (!row) throw new Error("Fixture must cross the route page boundary");
  await scrollGridCellIntoView({
    page,
    surface: view,
    recordId: row.record_id,
    cellKey: fixture.field,
  });
  await expect(
    page.getByTestId(rowCellTestId(row.record_id, fixture.field)),
  ).toContainText(String(row.cells[fixture.field]?.value));
  if (view === timelineViewSchemaId)
    await openTimelineInspector(page, row.record_id);
  else await openGenericInspectorForRecord(page, view, row.record_id);
  await expect(
    page.getByTestId(workbookInspectorCloseButtonTestId(view)),
  ).toBeVisible();
  await page.getByTestId(workbookInspectorCloseButtonTestId(view)).click();
  // Assessment rows are append-only; Indicator display values have their own lifecycle editor.
  if (view !== assessmentsViewSchemaId && !view.includes("indicators")) {
    await scrollGridCellIntoView({
      page,
      surface: view,
      recordId: row.record_id,
      cellKey: fixture.field,
    });
    const cell = page
      .getByTestId(rowCellTestId(row.record_id, fixture.field))
      .locator('xpath=ancestor-or-self::*[@role="gridcell"][1]');
    await cell.click();
    const editor = page
      .getByTestId(gridRowTestId(view, row.record_id))
      .getByRole("textbox")
      .first();
    await expect(editor).toBeVisible();
    const accepted = page.waitForResponse(
      (response) =>
        response.request().method() === "PATCH" &&
        response.url().endsWith(`/records/${row.record_id}`),
    );
    await editor.fill(`${String(row.cells[fixture.field]?.value)} edited`);
    await editor.press("Enter");
    expect((await accepted).ok()).toBe(true);
    await expect(
      page.getByTestId(gridRowTestId(view, row.record_id)),
    ).toHaveAttribute("data-grid-row-version", "2");
  }
  if (count > 300) {
    await more.click();
    await expect(controls).toContainText("300 records loaded; more available.");
    await more.click();
    await expect
      .poll(() => reads.at(-1)?.request.cursor_token)
      .not.toBe(later.request.cursor_token);
    await expect(
      controls.getByRole("button", { name: "Earlier rows" }),
    ).toHaveAttribute("aria-disabled", "false");
    const earlier = controls.getByRole("button", { name: "Earlier rows" });
    await earlier.focus();
    await earlier.press("Enter");
    await expect(controls).toContainText("100 records loaded; more available.");
    expect(reads.at(-1)?.request.cursor_token).toBeUndefined();
    await expect(earlier).toBeFocused();
  } else {
    await expect(controls).toContainText(
      `${count} records loaded; end of current results.`,
    );
    // The exhausted control stays only while it owns focus. Inspector departure
    // releases that focus, so the unavailable action is no longer rendered.
    await expect(more).toHaveCount(0);
  }
  await expect(
    page.getByTestId(gridShellTestId(view)).locator(gridSavedRowsSelector()),
  ).not.toHaveCount(0);
  expect(
    await page
      .getByTestId(gridShellTestId(view))
      .locator(gridSavedRowsSelector())
      .count(),
  ).toBeLessThan(100);
  expect(requireViewContract(view).viewSchemaId).toBe(view);
  // One initial read, explicit continuation/return, and bounded edit reconciliation.
  expect(reads.length).toBeLessThanOrEqual(12);
  expect(pageErrors).toEqual([]);
}

test("Workbook Group retains requested keyboard choice during delayed replacement", async ({
  page,
  workerAdmin,
}, testInfo) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-GROUP-PENDING"),
    "Workbook pending group choice",
  );
  await seed(page, incident, timelineViewSchemaId, 2, workerAdmin.user_id);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
  );
  const grouping = page.getByTestId(
    gridGroupingSelectTestId(timelineViewSchemaId),
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  const acceptedChip = page.getByTestId(
    workbookQueryEntryTestId(
      timelineViewSchemaId,
      "group",
      "timeline.capture_state",
    ),
  );
  const acceptedGroup = page.getByTestId(
    gridGroupRowTestId(timelineViewSchemaId, "timeline.capture_state", "rough"),
  );
  await expect(acceptedChip).toBeVisible();
  await expect(acceptedGroup).toBeVisible();
  let releaseOlder = () => {};
  let releaseLatest = () => {};
  const olderGate = new Promise<void>((resolve) => {
    releaseOlder = resolve;
  });
  const latestGate = new Promise<void>((resolve) => {
    releaseLatest = resolve;
  });
  const held = new Set<string>();
  const settled = new Set<string>();
  const requested: QueryWorkbookViewRequest[] = [];
  await page.route(
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      const request = route
        .request()
        .postDataJSON() as QueryWorkbookViewRequest;
      const groupBy = request.group_by;
      if (
        groupBy !== "timeline.has_evidence" &&
        groupBy !== "timeline.has_unresolved_mentions"
      ) {
        await route.continue();
        return;
      }
      requested.push(request);
      const response = await route.fetch();
      held.add(groupBy);
      await (groupBy === "timeline.has_evidence" ? olderGate : latestGate);
      try {
        await route.fulfill({ response });
      } catch (error) {
        if (route.request().failure() === null) throw error;
      } finally {
        settled.add(groupBy);
      }
    },
  );
  try {
    await grouping.click();
    await grouping.press("ArrowDown");
    await grouping.press("Enter");
    await expect.poll(() => held.has("timeline.has_evidence")).toBe(true);
    await testInfo.attach("group-pending-baseline", {
      body: JSON.stringify({
        selected: await grouping.inputValue(),
        focused: await grouping.evaluate(
          (element) => document.activeElement === element,
        ),
        acceptedChip: await acceptedChip.isVisible(),
        acceptedGroup: await acceptedGroup.isVisible(),
        requested,
      }),
      contentType: "application/json",
    });
    await expect(grouping).toHaveValue("timeline.has_evidence");
    await expect(grouping).toBeFocused();
    await expect(acceptedChip).toBeVisible();
    await expect(acceptedGroup).toBeVisible();
    await grouping.press("ArrowDown");
    await grouping.press("Enter");
    await expect
      .poll(() => held.has("timeline.has_unresolved_mentions"))
      .toBe(true);
    await expect(grouping).toHaveValue("timeline.has_unresolved_mentions");
    await expect(grouping).toBeFocused();
    await expect(acceptedChip).toBeVisible();
    await expect(acceptedGroup).toBeVisible();
    releaseLatest();
    const latestChip = page.getByTestId(
      workbookQueryEntryTestId(
        timelineViewSchemaId,
        "group",
        "timeline.has_unresolved_mentions",
      ),
    );
    await expect(latestChip).toBeVisible();
    await expect(
      page.locator(
        gridGroupRowsSelector(
          timelineViewSchemaId,
          "timeline.has_unresolved_mentions",
        ),
      ),
    ).not.toHaveCount(0);
    await expect(grouping).toBeFocused();
    releaseOlder();
    await expect.poll(() => settled.has("timeline.has_evidence")).toBe(true);
    await expect(grouping).toHaveValue("timeline.has_unresolved_mentions");
    await expect(latestChip).toBeVisible();
    await expect(acceptedChip).toHaveCount(0);
    await expect(acceptedGroup).toHaveCount(0);
    await expect(grouping).toBeFocused();
    expect(requested.map((entry) => entry.group_by)).toEqual([
      "timeline.has_evidence",
      "timeline.has_unresolved_mentions",
    ]);
    await latestChip.click();
    await expect(grouping).toBeFocused();
    await grouping.press("Escape");
    await expect(latestChip).toBeFocused();
    await expect(grouping).toHaveValue("timeline.has_unresolved_mentions");
  } finally {
    releaseLatest();
    releaseOlder();
  }
});

test("Workbook Group retains requested keyboard choice after failed replacement", async ({
  page,
  workerAdmin,
}, testInfo) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-GROUP-FAIL"),
    "Workbook failed group choice",
  );
  await seed(page, incident, timelineViewSchemaId, 2, workerAdmin.user_id);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
  );
  const grouping = page.getByTestId(
    gridGroupingSelectTestId(timelineViewSchemaId),
  );
  await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
  const acceptedChip = page.getByTestId(
    workbookQueryEntryTestId(
      timelineViewSchemaId,
      "group",
      "timeline.capture_state",
    ),
  );
  const acceptedGroup = page.getByTestId(
    gridGroupRowTestId(timelineViewSchemaId, "timeline.capture_state", "rough"),
  );
  await expect(acceptedChip).toBeVisible();
  await expect(acceptedGroup).toBeVisible();
  const requests: QueryWorkbookViewRequest[] = [];
  let failEvidence = true;
  let failNone = false;
  let releaseNone = () => {};
  const noneGate = new Promise<void>((resolve) => {
    releaseNone = resolve;
  });
  await page.route(
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      const request = route
        .request()
        .postDataJSON() as QueryWorkbookViewRequest;
      requests.push(request);
      if (request.group_by === "timeline.has_evidence" && failEvidence) {
        failEvidence = false;
        await route.abort("failed");
        return;
      }
      if (request.group_by === undefined && failNone) {
        failNone = false;
        await noneGate;
        await route.abort("failed");
        return;
      }
      await route.continue();
    },
  );
  await grouping.click();
  await grouping.press("ArrowDown");
  await grouping.press("Enter");
  await expect
    .poll(() => requests.at(-1)?.group_by)
    .toBe("timeline.has_evidence");
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  const retry = browsing.getByRole("button", { name: "Retry", exact: true });
  await expect(retry).toBeVisible();
  await testInfo.attach("group-failed-baseline", {
    body: JSON.stringify({
      selected: await grouping.inputValue(),
      focused: await grouping.evaluate(
        (element) => document.activeElement === element,
      ),
      acceptedChip: await acceptedChip.isVisible(),
      acceptedGroup: await acceptedGroup.isVisible(),
      acceptedGroupCount: await page
        .locator(
          gridGroupRowsSelector(timelineViewSchemaId, "timeline.capture_state"),
        )
        .count(),
      requested: requests.at(-1),
    }),
    contentType: "application/json",
  });
  await expect(grouping).toHaveValue("timeline.has_evidence");
  await expect(grouping).toBeFocused();
  await expect(acceptedChip).toBeVisible();
  await expect(acceptedGroup).toBeVisible();
  await expect(grouping).toHaveAttribute("aria-describedby", /unapplied/u);
  const groupStatus = page.locator(
    `[id="${gridGroupingSelectTestId(timelineViewSchemaId)}-unapplied"]`,
  );
  await expect(groupStatus).toContainText(
    "retained results grouped by Capture State",
  );
  await retry.click();
  const evidenceChip = page.getByTestId(
    workbookQueryEntryTestId(
      timelineViewSchemaId,
      "group",
      "timeline.has_evidence",
    ),
  );
  await expect(evidenceChip).toBeVisible();
  await expect(
    page.locator(
      gridGroupRowsSelector(timelineViewSchemaId, "timeline.has_evidence"),
    ),
  ).not.toHaveCount(0);
  await expect(grouping).toHaveValue("timeline.has_evidence");
  await expect(acceptedChip).toHaveCount(0);
  expect(requests.at(-1)?.group_by).toBe("timeline.has_evidence");

  failNone = true;
  const beforeNone = requests.length;
  await grouping.click();
  await grouping.selectOption("");
  try {
    await expect.poll(() => requests.length).toBeGreaterThan(beforeNone);
    expect(requests.at(-1)?.group_by).toBeUndefined();
    await expect(grouping).toHaveValue("");
    await expect(evidenceChip).toBeVisible();
    await expect(
      page.locator(
        gridGroupRowsSelector(timelineViewSchemaId, "timeline.has_evidence"),
      ),
    ).not.toHaveCount(0);
  } finally {
    releaseNone();
  }
  await expect(retry).toBeVisible();
  await expect(grouping).toHaveValue("");
  await expect(evidenceChip).toBeVisible();
  await browsing.getByRole("button", { name: "Revert", exact: true }).click();
  await expect(grouping).toHaveValue("timeline.has_evidence");
  await expect(evidenceChip).toBeVisible();
  await expect(groupStatus).toHaveCount(0);

  await grouping.click();
  await grouping.selectOption("");
  await expect(grouping).toHaveValue("");
  await expect(evidenceChip).toHaveCount(0);
  await expect(
    page.locator(
      gridGroupRowsSelector(timelineViewSchemaId, "timeline.has_evidence"),
    ),
  ).toHaveCount(0);
  expect(requests.at(-1)?.group_by).toBeUndefined();
});

test("Workbook Group pointer selection stays scoped to Hosts", async ({
  page,
  workerAdmin,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-GROUP-HOST"),
    "Workbook Host group choice",
  );
  await seed(page, incident, hostsViewSchemaId, 2, workerAdmin.user_id);
  await seed(page, incident, timelineViewSchemaId, 1, workerAdmin.user_id);
  const requests: QueryWorkbookViewRequest[] = [];
  await page.route(
    `**/incidents/${incident}/views/${hostsViewSchemaId}/query`,
    async (route) => {
      requests.push(route.request().postDataJSON() as QueryWorkbookViewRequest);
      await route.continue();
    },
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(hostsViewSchemaId)}`,
  );
  const hostGrouping = page.getByTestId(
    gridGroupingSelectTestId(hostsViewSchemaId),
  );
  await expect(hostGrouping).toHaveValue("");
  await hostGrouping.click();
  await hostGrouping.selectOption("host.host_state");
  await expect(hostGrouping).toHaveValue("host.host_state");
  await expect(
    page.getByTestId(
      workbookQueryEntryTestId(hostsViewSchemaId, "group", "host.host_state"),
    ),
  ).toBeVisible();
  await expect(
    page.locator(gridGroupRowsSelector(hostsViewSchemaId, "host.host_state")),
  ).not.toHaveCount(0);
  expect(requests.at(-1)?.group_by).toBe("host.host_state");
  await switchOrdinarySheet(page, timelineViewSchemaId);
  await expect(
    page.getByTestId(gridGroupingSelectTestId(timelineViewSchemaId)),
  ).toHaveValue("");
  await expect(hostGrouping).toHaveCount(0);
});

test("Workbook continuation reaches and returns later cartulary.view.timeline.v2 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, timelineViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.hosts.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, hostsViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.identities.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, identitiesViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.notes.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, notesViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.evidence.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, evidenceViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.indicators.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, indicatorsViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.parties.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, partiesViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.task_requests.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, taskRequestsViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.decisions.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, decisionsViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.comm_log.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, commLogViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.handoff.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, handoffViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.status_review.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, statusReviewViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.lesson.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, lessonViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.findings.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, findingsViewSchemaId, workerAdmin.user_id);
});
test("Workbook continuation reaches and returns later cartulary.view.investigative_queries.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(
    page,
    investigativeQueriesViewSchemaId,
    workerAdmin.user_id,
  );
});
test("Workbook continuation reaches and returns later cartulary.view.forensic_keywords.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(
    page,
    forensicKeywordsViewSchemaId,
    workerAdmin.user_id,
  );
});
test("Workbook continuation reaches and returns later cartulary.view.assessments.v1 records through the real route", async ({
  page,
  workerAdmin,
}) => {
  await exerciseSurface(page, assessmentsViewSchemaId, workerAdmin.user_id);
});

test("Workbook browsing retains off-window drafts and sheet anchors through twenty-checkpoint exhaustion", async ({
  page,
  workerAdmin,
}) => {
  test.setTimeout(300_000);
  const view = notesViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-LONG"),
    "Long live workbook traversal",
  );
  await seed(page, incident, view, 2505, workerAdmin.user_id);
  const reads = await observeQuery(page, incident, view);
  let writes = 0;
  page.on("request", (request) => {
    if (request.method() === "PATCH" && request.url().includes("/records/"))
      writes++;
  });
  await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText("100 records loaded; more available.");
  const original = reads.at(-1)?.response.data.rows[0];
  if (!original) throw new Error("Missing original source");
  const originalCell = page.getByTestId(
    rowCellTestId(original.record_id, "note.title"),
  );
  await originalCell.click();
  const editor = page
    .getByTestId(gridRowTestId(view, original.record_id))
    .getByRole("textbox")
    .first();
  await editor.fill("Exact retained draft after query eviction");
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  for (let index = 0; index < 3; index++) {
    const previous = reads.length;
    await more.click();
    await expect.poll(() => reads.length).toBe(previous + 1);
    await expect(more).toHaveAttribute("aria-disabled", "false");
  }
  expect(writes).toBe(0);
  await expect(
    page.getByTestId(gridRowTestId(view, original.record_id)),
  ).toHaveCount(0);
  await controls.getByRole("button", { name: "Earlier rows" }).click();
  await expect(controls).toContainText("100 records loaded; more available.");
  await expect(
    page
      .getByTestId(gridRowTestId(view, original.record_id))
      .getByRole("textbox"),
  ).toHaveCount(0);
  await originalCell.click();
  await expect(editor).toHaveValue("Exact retained draft after query eviction");
  await editor.press("Escape");
  expect(writes).toBe(0);
  for (let index = 0; index < 25; index++) {
    const previous = reads.length;
    await more.focus();
    await more.press("Enter");
    await expect.poll(() => reads.length).toBe(previous + 1);
    await expect(controls).not.toContainText("Loading");
    await expect(more).toBeFocused();
    const loaded = Number(
      (await controls.innerText()).match(/(\d+) records loaded/u)?.[1],
    );
    expect(loaded).toBeLessThanOrEqual(300);
  }
  await expect(controls).toContainText(
    "205 records loaded; end of current results.",
  );
  await expect(more).toHaveAttribute("aria-disabled", "true");
  const last = reads.at(-1)?.response.data.rows.at(-1);
  if (!last) throw new Error("Missing terminal record");
  await scrollGridCellIntoView({
    page,
    surface: view,
    recordId: last.record_id,
    cellKey: "note.title",
  });
  const lastCell = page.getByTestId(
    rowCellTestId(last.record_id, "note.title"),
  );
  await activateCommittedGridCell(
    lastCell.locator('xpath=ancestor::*[@role="gridcell"][1]'),
  );
  await switchOrdinarySheet(page, timelineViewSchemaId);
  const beforeReturn = reads.length;
  await switchOrdinarySheet(page, view);
  await expect(controls).toContainText(
    "5 records loaded; end of current results.",
  );
  expect(reads.length).toBe(beforeReturn + 1);
  expect(reads.at(-1)?.request.cursor_token).toBe(
    reads[beforeReturn - 1]?.request.cursor_token,
  );
  await expect(lastCell).toBeVisible();
  for (let index = 0; index < 20; index++) {
    const before = reads.length;
    await controls.getByRole("button", { name: "Earlier rows" }).click();
    await expect.poll(() => reads.length).toBe(before + 1);
    await expect(controls).not.toContainText("Loading");
  }
  await expect(
    controls.getByRole("button", { name: "Earlier rows" }),
  ).toHaveAttribute("aria-disabled", "true");
  await expect(controls).toContainText("Earlier history limit reached");
  await page.setViewportSize({ width: 360, height: 720 });
  const rect = await controls.boundingBox();
  if (!rect) throw new Error("Browsing controls are missing");
  expect(rect.x).toBeGreaterThanOrEqual(0);
  expect(rect.x + rect.width).toBeLessThanOrEqual(360);
  const refresh = controls.getByRole("button", {
    name: "Refresh",
    exact: true,
  });
  await refresh.focus();
  await refresh.press("Enter");
  await expect(controls).toContainText("100 records loaded; more available.");
  expect(reads.at(-1)?.request.cursor_token).toBeUndefined();
  await expect(refresh).toBeFocused();
  await page.setViewportSize({ width: 720, height: 1280 });
  await page.evaluate(() => {
    document.documentElement.style.zoom = "200%";
  });
  await refresh.scrollIntoViewIfNeeded();
  await refresh.focus();
  await refresh.press("Enter");
  await expect(controls).toContainText("100 records loaded; more available.");
  await expect(refresh).toBeFocused();
  const zoomed = await controls.boundingBox();
  expect(zoomed && zoomed.x >= 0 && zoomed.x + zoomed.width <= 720).toBe(true);
  await page.evaluate(() => {
    document.documentElement.style.zoom = "";
  });
});

test("Workbook canonical filters continue without feedback and saved views retain authored sort intent", async ({
  page,
  workerAdmin,
}) => {
  test.setTimeout(180_000);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-CANON"),
    "Canonical workbook continuation",
  );
  await seed(page, incident, notesViewSchemaId, 105, workerAdmin.user_id);
  await seed(page, incident, timelineViewSchemaId, 105, workerAdmin.user_id);
  const evidenceView = evidenceViewSchemaId;
  await seed(page, incident, evidenceView, 105, workerAdmin.user_id);
  const evidenceReads = await observeQuery(page, incident, evidenceView);
  const notes = await observeQuery(page, incident, notesViewSchemaId);
  const timeline = await observeQuery(page, incident, timelineViewSchemaId);
  const sorts = requireViewContract(timelineViewSchemaId)
    .sortFields.filter((field) => field !== "timeline.activity_sort_ts")
    .slice(0, 8)
    .map((field_key) => ({ field_key, direction: "asc" as const }));
  const grouped = await createSavedView(page, incident, {
    display_name: "Eight authored sorts",
    view_schema_id: timelineViewSchemaId,
    query_json: {
      filters: [],
      sort: sorts,
      group_by: "timeline.capture_state",
    },
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(notesViewSchemaId)}`,
  );
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText("100 records loaded; more available.");
  const before = notes.length;
  await applyFilterChip(
    page,
    notesViewSchemaId,
    "note.full_text",
    "  BROWSING   fixture browsing  ",
  );
  await expect
    .poll(() => notes.at(-1)?.response.meta.query.filters.length)
    .toBe(1);
  await expect(controls).toContainText("100 records loaded; more available.");
  let normalized = notes.at(-1);
  if (!normalized) throw new Error("Missing canonical read");
  const arg = normalized.response.meta.query.filters[0]?.arg;
  await expect(
    page.getByTestId(
      workbookQueryEntryTestId(notesViewSchemaId, "filter", "note.full_text"),
    ),
  ).toContainText(String(arg?.query));
  expect(notes.length).toBe(before + 1);
  await applyFilterChip(
    page,
    notesViewSchemaId,
    "note.tags",
    " beta,alpha,beta ",
  );
  await expect
    .poll(() => notes.at(-1)?.response.meta.query.filters.length)
    .toBe(2);
  await expect(controls).toContainText("100 records loaded; more available.");
  normalized = notes.at(-1);
  if (!normalized) throw new Error("Missing set-like canonical read");
  expect(
    normalized.response.meta.query.filters.find(
      (filter) => filter.field_key === "note.tags",
    )?.arg,
  ).toEqual({ values: ["alpha", "beta"] });
  expect(notes.length).toBe(before + 2);
  await controls
    .getByRole("button", { name: "Load more", exact: true })
    .click();
  await expect(controls).toContainText(
    "105 records loaded; end of current results.",
  );
  expect(notes.at(-1)?.request.filters).toEqual(
    normalized.response.meta.query.filters,
  );
  expect(notes.at(-1)?.request.cursor_token).toBe(
    normalized.response.meta.paging.next_cursor,
  );
  const savedRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith(`/incidents/${incident}/saved-views`),
  );
  await setSavedViewDraftName(
    page,
    notesViewSchemaId,
    "Canonical filtered notes",
  );
  await createSavedViewFromCurrentSurface(page, notesViewSchemaId);
  const persisted = (await savedRequest).postDataJSON();
  expect(persisted.query_json.filters).toEqual(
    normalized.response.meta.query.filters,
  );
  expect(persisted.query_json.sort).toEqual([]);
  expect(persisted.query_json.group_by).toBeUndefined();
  await expect(
    page.getByTestId(savedViewModifiedTestId(notesViewSchemaId)),
  ).toHaveCount(0);
  await switchOrdinarySheet(page, timelineViewSchemaId);
  await selectSavedView(page, timelineViewSchemaId, grouped.saved_view_id);
  await expect.poll(() => timeline.at(-1)?.request.sort).toEqual(sorts);
  await expect(controls).toContainText("100 records loaded; more available.");
  expect(timeline.at(-1)?.request.group_by).toBe("timeline.capture_state");
  expect(timeline.at(-1)?.response.meta.query.sort).toHaveLength(10);
  await expect(
    page.getByTestId(savedViewModifiedTestId(timelineViewSchemaId)),
  ).toHaveCount(0);
  await controls
    .getByRole("button", { name: "Load more", exact: true })
    .click();
  await expect(controls).toContainText(
    "105 records loaded; end of current results.",
  );
  expect(timeline.at(-1)?.request.sort).toEqual(sorts);
  await page
    .getByTestId(workbookSortMenuTriggerTestId(timelineViewSchemaId))
    .click();
  for (let remaining = 8; remaining > 0; remaining--) {
    await page
      .getByRole("menuitem", { name: /^Remove .+ sort$/u })
      .first()
      .click();
    await expect
      .poll(() => timeline.at(-1)?.request.sort?.length ?? 0)
      .toBe(remaining - 1);
  }
  await page
    .getByTestId(workbookSortMenuTriggerTestId(timelineViewSchemaId))
    .press("Escape");
  expect(timeline.at(-1)?.request.sort).toBeUndefined();
  await changeGrouping(page, timelineViewSchemaId, "");
  await expect.poll(() => timeline.at(-1)?.request.group_by).toBeUndefined();
  expect(timeline.at(-1)?.request.sort).toBeUndefined();
  const clearedRequest = page.waitForRequest(
    (request) =>
      request.method() === "POST" &&
      request.url().endsWith(`/incidents/${incident}/saved-views`),
  );
  await setSavedViewDraftName(page, timelineViewSchemaId, "Cleared overrides");
  await createSavedViewFromCurrentSurface(page, timelineViewSchemaId);
  const cleared = (await clearedRequest).postDataJSON();
  expect(cleared.query_json.sort).toEqual([]);
  expect(cleared.query_json.group_by).toBeUndefined();
  await switchOrdinarySheet(page, evidenceView);
  await expect(controls).toContainText("100 records loaded; more available.");
  await page
    .getByTestId(workbookFilterPopoverTriggerTestId(evidenceView))
    .click();
  await page
    .getByTestId(gridFilterFieldTestId(evidenceView))
    .selectOption("evidence.collector_party_text");
  await page
    .getByTestId(workbookFilterOperatorTestId(evidenceView))
    .selectOption("prefix");
  await page.getByTestId(gridFilterValueTestId(evidenceView)).fill("BRIDGE");
  await page.getByTestId(gridFilterApplyTestId(evidenceView)).click();
  await expect
    .poll(() => evidenceReads.at(-1)?.response.meta.query.filters.length)
    .toBe(1);
  await expect(controls).toContainText("100 records loaded; more available.");
  const prefix = evidenceReads.at(-1);
  if (!prefix) throw new Error("Missing prefix response");
  expect(prefix.response.meta.query.filters[0]?.arg).toEqual({
    value: "bridge",
  });
  await expect(
    page.getByTestId(
      workbookQueryEntryTestId(
        evidenceView,
        "filter",
        "evidence.collector_party_text",
      ),
    ),
  ).toContainText("bridge");
  await controls
    .getByRole("button", { name: "Load more", exact: true })
    .click();
  await expect(controls).toContainText(
    "105 records loaded; end of current results.",
  );
  expect(evidenceReads.at(-1)?.request.filters).toEqual(
    prefix.response.meta.query.filters,
  );
  expect(evidenceReads.at(-1)?.request.cursor_token).toBe(
    prefix.response.meta.paging.next_cursor,
  );
});

test("Workbook real-route recovery preserves accepted labels and reconciles live placement without draining pages", async ({
  page,
  workerAdmin,
}) => {
  test.setTimeout(180_000);
  const view = notesViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-LIVE"),
    "Workbook live continuation recovery",
  );
  await seed(page, incident, view, 405, workerAdmin.user_id);
  const reads: {
    request: QueryWorkbookViewRequest;
    response?: QueryWorkbookViewResponse;
    status: number;
    delivered?: boolean;
  }[] = [];
  let mode: "normal" | "abort" | "malformed" | "invalid_cursor" | "hold" =
    "normal";
  let held = false;
  let release: () => void = () => {};
  let gate = Promise.resolve();
  const path = `**/incidents/${incident}/views/${view}/query`;
  await page.route(path, async (route) => {
    const request = route.request().postDataJSON() as QueryWorkbookViewRequest;
    const behavior = mode;
    mode = "normal";
    if (behavior === "abort") {
      reads.push({ request, status: 0 });
      await route.abort("failed");
      return;
    }
    const response = await route.fetch(
      behavior === "invalid_cursor"
        ? { postData: { ...request, cursor_token: "not-a-core-cursor" } }
        : {},
    );
    const payload = await response.json();
    const read = {
      request,
      status: response.status(),
      ...(response.ok() ? { response: payload } : {}),
      delivered: false,
    };
    reads.push(read);
    if (behavior === "hold") {
      held = true;
      await gate;
    }
    if (behavior === "malformed") {
      delete payload.meta.paging;
      await route.fulfill({ response, json: payload });
    } else await route.fulfill({ response });
    // route.fetch also completes for reads cancelled during startup. Only a
    // response delivered to the page can own the next continuation token.
    const delivered = await route.request().response();
    read.delivered =
      delivered !== null && (await delivered.finished()) === null;
  });
  const sockets = installIncidentSocketMonitor(page, incident);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await sockets.waitForAcceptedSocket();
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  await expect(controls).toContainText("100 records loaded; more available.");
  await expect(more).toBeEnabled();
  await expect.poll(() => reads.some((read) => read.delivered)).toBe(true);
  const first = reads.filter((read) => read.delivered).at(-1)?.response;
  if (!first) throw new Error("Missing delivered first page");
  mode = "abort";
  const failedIndex = reads.length;
  await more.evaluate((button) => {
    (button as HTMLButtonElement).click();
    (button as HTMLButtonElement).click();
  });
  await expect(
    controls.getByRole("button", { name: "Retry", exact: true }),
  ).toBeVisible();
  expect(reads.length).toBe(failedIndex + 1);
  expect(reads.at(-1)?.request.cursor_token).toBe(
    first.meta.paging.next_cursor,
  );
  await expect(controls).toContainText("100 records loaded");
  const failed = reads.at(-1)?.request;
  await controls.getByRole("button", { name: "Retry", exact: true }).click();
  await expect(controls).toContainText("200 records loaded; more available.");
  expect(reads.at(-1)?.request).toEqual(failed);
  mode = "invalid_cursor";
  const invalidIndex = reads.length;
  await more.click();
  await expect(controls).toContainText("100 records loaded; more available.");
  expect(reads.slice(invalidIndex).map((read) => read.status)).toEqual([
    400, 200,
  ]);
  expect(reads.at(-1)?.request.cursor_token).toBeUndefined();
  mode = "malformed";
  await more.click();
  await expect(
    controls.getByRole("button", { name: "Retry", exact: true }),
  ).toBeVisible();
  await expect(controls).toContainText("100 records loaded");
  await controls.getByRole("button", { name: "Retry", exact: true }).click();
  await expect(controls).toContainText("200 records loaded; more available.");
  mode = "abort";
  await sortByHeader(page, view, "note.title");
  await expect(controls).toContainText("Query changes are unapplied.");
  await expect(
    page.getByTestId(workbookSortMenuTriggerTestId(view)),
  ).toHaveAttribute("aria-label", "Sort, no user sorts");
  await expect(controls).toContainText("200 records loaded");
  await controls.getByRole("button", { name: "Revert", exact: true }).click();
  await expect(controls).not.toContainText("Query changes are unapplied.");
  await expect(controls).toContainText("200 records loaded; more available.");
  const liveStart = reads.length;
  const inserted = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("wqc-live-insert"),
    "note.title": "Live inserted",
    "note.body": "Live source",
  });
  await expect(
    page.getByTestId(rowCellTestId(inserted.record_id, "note.title")),
  ).toHaveText("Live inserted");
  await expect(controls).toContainText("200 records loaded; more available.");
  expect(reads.length - liveStart).toBeLessThanOrEqual(4);
  const mutate = async (
    operationID: "deleteRecord" | "restoreRecord",
    version: number,
  ) =>
    publicHttpOperation({
      operationID,
      pathParameters: { record_id: inserted.record_id },
      headers: await csrfHeaders(page),
      request: atJsonOrigin(page.request, apiBase),
      body: {
        client_txn_id: uniqueTxn(operationID),
        base_row_version: version,
        reason: "Workbook query reconciliation",
      },
    });
  const deleted = await mutate("deleteRecord", 1);
  expect(deleted.ok).toBe(true);
  await expect(
    page.getByTestId(rowCellTestId(inserted.record_id, "note.title")),
  ).toHaveCount(0);
  expect((await mutate("restoreRecord", 2)).ok).toBe(true);
  await expect(
    page.getByTestId(rowCellTestId(inserted.record_id, "note.title")),
  ).toHaveText("Live inserted");
  gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  mode = "hold";
  await controls.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect.poll(() => held).toBe(true);
  await patchRecord(page, inserted.record_id, {
    client_txn_id: uniqueTxn("wqc-newer"),
    base_row_version: 3,
    view_schema_id: view,
    changes: [{ field_key: "note.title", value: "Newer accepted live source" }],
  });
  release();
  await expect(
    page.getByTestId(rowCellTestId(inserted.record_id, "note.title")),
  ).toHaveText("Newer accepted live source");
  await expect(
    page.getByTestId(gridRowTestId(view, inserted.record_id)),
  ).toHaveAttribute("data-grid-row-version", "4");
  await expect(controls).toContainText("100 records loaded; more available.");
});

test("Workbook continuation revalidates role and membership and fences a late authorized page", async ({
  page,
  workerAdmin,
  workerAdminRequest,
  sessionTracker,
}) => {
  test.setTimeout(180_000);
  const view = notesViewSchemaId;
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-AUTH"),
    "Workbook continuation authority",
  );
  await seed(page, incident, view, 405, workerAdmin.user_id);
  const member = await createIncidentMemberUser(page, incident, {
    email: uniqueEmail("wqc-authority"),
    display_name: "Workbook paging analyst",
    initial_password: "WorkbookPaging1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "wqc-authority",
    email: member.email,
    password: member.initial_password,
    purpose: "live query authority",
    userId: member.user_id,
  });
  const reads = await observeQuery(page, incident, view);
  const sockets = installIncidentSocketMonitor(page, incident);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await sockets.waitForAcceptedSocket();
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  await expect(controls).toContainText("100 records loaded; more available.");
  await more.click();
  await expect(controls).toContainText("200 records loaded; more available.");
  const membershipPath = `/api/v1/incidents/${incident}/memberships/${member.user_id}`;
  expect(
    (
      await workerAdminRequest.patch(membershipPath, {
        data: { base_membership_version: 1, role: "viewer" },
      })
    ).ok(),
  ).toBe(true);
  await more.click();
  await expect(controls).toContainText("300 records loaded; more available.");
  const later = reads.at(-1)?.response.data.rows[0];
  if (!later) throw new Error("Missing live-authorized later row");
  const deniedWrite = await publicHttpOperation({
    operationID: "patchRecord",
    pathParameters: { record_id: later.record_id },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
    body: {
      view_schema_id: view,
      base_row_version: later.row_version,
      client_txn_id: uniqueTxn("wqc-denied"),
      changes: [{ field_key: "note.title", value: "Must not change" }],
    },
  });
  expect(deniedWrite.ok).toBe(false);
  expect(deniedWrite.status).toBe(403);
  let release: () => void = () => {};
  let held = false;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route(
    `**/incidents/${incident}/views/${view}/query`,
    async (route) => {
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      held = true;
      await gate;
      await route.fulfill({ response });
    },
    { times: 1 },
  );
  await more.click();
  await expect.poll(() => held).toBe(true);
  expect(
    (
      await workerAdminRequest.delete(membershipPath, {
        data: { base_membership_version: 2 },
      })
    ).status(),
  ).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  release();
  await expect(controls).toHaveCount(0);
  await expect(page.getByTestId(gridShellTestId(view))).toHaveCount(0);
  await expect(page).not.toHaveURL(/incident_id=/u);
});

test("Workbook expiry clears continuation and account replacement starts a readable closed incident without an old cursor", async ({
  page,
  workerAdmin,
  sessionTracker,
}) => {
  test.setTimeout(180_000);
  const clock = new TestClock(page);
  await clock.reset();
  try {
    const view = notesViewSchemaId;
    const incident = await createIncident(
      page,
      uniqueIncidentKey("WQC-SESSION"),
      "Workbook paging session lifetime",
    );
    await seed(page, incident, view, 305, workerAdmin.user_id);
    const members = [];
    for (const label of ["first", "replacement"])
      members.push(
        await createIncidentMemberUser(page, incident, {
          email: uniqueEmail(`wqc-${label}`),
          display_name: `Workbook ${label}`,
          initial_password: "WorkbookPaging1!",
          role: "viewer",
          is_deployment_admin: false,
          mfa_required: false,
        }),
      );
    const before = await currentLifecycle(page, incident);
    expect(
      (
        await lifecycleAction(page, incident, "closeIncident", {
          client_txn_id: uniqueTxn("wqc-close"),
          base_incident_version: before.incident_version,
          reason: "Closed readable continuation evidence",
        })
      ).ok,
    ).toBe(true);
    const first = members[0],
      second = members[1];
    if (!first || !second) throw new Error("Missing account fixtures");
    await sessionTracker.loginTrackedUser(page, {
      createdBy: "wqc-session",
      email: first.email,
      password: first.initial_password,
      purpose: "closed workbook browsing",
      userId: first.user_id,
    });
    const reads = await observeQuery(page, incident, view);
    await page.goto(
      `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
    );
    const controls = page.getByRole("group", { name: "Workbook browsing" });
    await expect(controls).toContainText("100 records loaded; more available.");
    await controls
      .getByRole("button", { name: "Load more", exact: true })
      .click();
    await expect(controls).toContainText("200 records loaded; more available.");
    const session = await readCurrentSession(page);
    await clock.setAfter(session.session_expires_at);
    // Closed incidents terminate their collaboration stream; an explicit read observes expiry.
    if (await controls.isVisible())
      await controls
        .getByRole("button", { name: "Refresh", exact: true })
        .click();
    await expect(page.getByTestId(authTestId("shell"))).toBeVisible({
      timeout: 30_000,
    });
    await expect(controls).toHaveCount(0);
    await expect(page.getByTestId(gridShellTestId(view))).toHaveCount(0);
    await clock.reset();
    const beforeReplacement = reads.length;
    await sessionTracker.loginTrackedUser(page, {
      createdBy: "wqc-session",
      email: second.email,
      password: second.initial_password,
      purpose: "replacement account browsing",
      userId: second.user_id,
      recovery: true,
    });
    // Account replacement may return to landing; enter the same readable incident explicitly.
    if (!new URL(page.url()).searchParams.has("incident_id"))
      await page.goto(
        `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
      );
    await expect(controls).toContainText("100 records loaded; more available.");
    expect(
      reads
        .slice(beforeReplacement)
        .every((read) => read.request.cursor_token === undefined),
    ).toBe(true);
    await controls
      .getByRole("button", { name: "Load more", exact: true })
      .click();
    await expect(controls).toContainText("200 records loaded; more available.");
  } finally {
    await clock.reset();
  }
});
