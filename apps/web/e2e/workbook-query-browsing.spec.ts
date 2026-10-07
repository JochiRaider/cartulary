import type {
  CreateViewRowRequest,
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts/http";
import {
  applyFilterChip,
  changeGrouping,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  authTestId,
  draftCellTestId,
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
  workbookFocusAnchorTestId,
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
import type { Locator, Page, Route } from "@playwright/test";
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
import { holdBrowserRequest } from "./support/transport/requestInterception";
import {
  authorLiteralMembers,
  expectLiteralMembership,
  literalSetFixture,
} from "./support/workbook/literalSetFilters";
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
  updateSavedViewFromCurrentSurface,
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

test("Literal set tag filters select distinct exact Timeline and Notes memberships", async ({
  page,
}) => {
  for (const view of [timelineViewSchemaId, notesViewSchemaId]) {
    const f = await literalSetFixture(page, view);
    const reads = await observeQuery(page, f.incident, view);
    const ids = f.rows.map((row) => row.record_id);
    const [a, , c] = ids;
    if (!a || !c) throw new Error("Missing literal fixture identities");
    for (const op of ["contains_any", "contains_all"] as const) {
      for (const values of [["review,priority"], ["review", "priority"]]) {
        const chip = page.getByTestId(
          workbookQueryEntryTestId(view, "filter", f.tags),
        );
        if (await chip.count()) {
          await chip.focus();
          await page.keyboard.press("Delete");
          await expect(chip).toHaveCount(0);
        }
        await page
          .getByTestId(workbookFilterPopoverTriggerTestId(view))
          .click();
        await page
          .getByTestId(gridFilterFieldTestId(view))
          .selectOption(f.tags);
        await page
          .getByTestId(workbookFilterOperatorTestId(view))
          .selectOption(op);
        await authorLiteralMembers(page.getByRole("dialog"), values);
        await page.getByTestId(gridFilterApplyTestId(view)).click();
        const expected =
          values.length === 1
            ? [a]
            : op === "contains_any"
              ? ids.slice(1)
              : [c];
        await expect
          .poll(() =>
            reads
              .at(-1)
              ?.response.data.rows.map((row) => row.record_id)
              .sort(),
          )
          .toEqual([...expected].sort());
        await expectLiteralMembership(page, view, expected, ids);
        await expect(chip).toHaveAccessibleName(
          new RegExp(
            values.length === 1
              ? '\\["review,priority"\\]'
              : '\\["priority", "review"\\]',
          ),
        );
      }
    }
  }
});

test("Literal saved arrays reload reopen and reapply without changing membership", async ({
  page,
}) => {
  const f = await literalSetFixture(page);
  const view = timelineViewSchemaId,
    ids = f.rows.map((row) => row.record_id);
  const a = ids[0];
  if (!a) throw new Error("Missing literal fixture identity");
  const reads = await observeQuery(page, f.incident, view);
  for (const op of ["contains_any", "contains_all"] as const) {
    const filter = {
      field_key: f.tags,
      op,
      arg: { values: ["review,priority"] },
    };
    const saved = await createSavedView(page, f.incident, {
      display_name: `Literal ${op}`,
      view_schema_id: view,
      query_json: { filters: [filter], sort: [] },
    });
    expect(saved.query_json.filters).toEqual([filter]);
    await selectSavedView(page, view, saved.saved_view_id);
    await expectLiteralMembership(page, view, [a], ids);
    await page.reload();
    await selectSavedView(page, view, saved.saved_view_id);
    await expectLiteralMembership(page, view, [a], ids);
    await page
      .getByTestId(workbookQueryEntryTestId(view, "filter", f.tags))
      .click();
    await page.getByTestId(gridFilterApplyTestId(view)).click();
    await expect
      .poll(() => reads.at(-1)?.response.meta.query.filters)
      .toEqual([filter]);
    await expectLiteralMembership(page, view, [a], ids);
    const persisted = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().endsWith(`/incidents/${f.incident}/saved-views`),
    );
    await setSavedViewDraftName(page, view, `Authored literal ${op}`);
    await createSavedViewFromCurrentSurface(page, view);
    const persistence = await persisted;
    expect(persistence.ok()).toBe(true);
    const resource = (await persistence.json()).data;
    expect(resource.query_json.filters).toEqual([filter]);
    await page.reload();
    await selectSavedView(page, view, resource.saved_view_id);
    await expectLiteralMembership(page, view, [a], ids);
    await page
      .getByTestId(workbookQueryEntryTestId(view, "filter", f.tags))
      .click();
    await expect(
      page.getByRole("textbox", { name: "Value 1", exact: true }),
    ).toHaveValue("review,priority");
    await expect(page.getByRole("textbox", { name: /^Value / })).toHaveCount(1);
    await page.getByTestId(gridFilterApplyTestId(view)).click();
    await expectLiteralMembership(page, view, [a], ids);
  }
});

test("Ordinary text equality sets retain comma-containing organization members", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("LITERAL-PARTY"),
    "Organization membership",
  );
  const rows = [];
  for (const value of ["Northwind, Inc.", "Northwind", "Inc."])
    rows.push(
      await createViewRow(page, incident, partiesViewSchemaId, {
        client_txn_id: uniqueTxn("literal-party"),
        "party.display_name": value,
        "party.party_kind": "organization",
        "party.organization_name": value,
      }),
    );
  const first = rows[0];
  if (!first) throw new Error("Missing organization fixture identity");
  const reads = await observeQuery(page, incident, partiesViewSchemaId);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${partiesViewSchemaId}`,
  );
  await page
    .getByTestId(workbookFilterPopoverTriggerTestId(partiesViewSchemaId))
    .click();
  await page
    .getByTestId(gridFilterFieldTestId(partiesViewSchemaId))
    .selectOption("party.organization_name");
  await page.getByLabel("Equality operand kind").selectOption("values");
  await authorLiteralMembers(page.getByRole("dialog"), ["Northwind, Inc."]);
  await page.getByTestId(gridFilterApplyTestId(partiesViewSchemaId)).click();
  await expect
    .poll(() => reads.at(-1)?.response.data.rows.map((row) => row.record_id))
    .toEqual([first.record_id]);
  await expectLiteralMembership(
    page,
    partiesViewSchemaId,
    [first.record_id],
    rows.map((row) => row.record_id),
  );
  await page
    .getByTestId(
      workbookQueryEntryTestId(
        partiesViewSchemaId,
        "filter",
        "party.organization_name",
      ),
    )
    .click();
  await page.getByTestId(gridFilterApplyTestId(partiesViewSchemaId)).click();
  await expectLiteralMembership(
    page,
    partiesViewSchemaId,
    [first.record_id],
    rows.map((row) => row.record_id),
  );
  await page
    .getByTestId(
      workbookQueryEntryTestId(
        partiesViewSchemaId,
        "filter",
        "party.organization_name",
      ),
    )
    .click();
  await page
    .getByRole("textbox", { name: "Value 1", exact: true })
    .fill("Northwind");
  await page.getByRole("button", { name: "Add value", exact: true }).click();
  await page
    .getByRole("textbox", { name: "Value 2", exact: true })
    .fill("Inc.");
  await page.getByTestId(gridFilterApplyTestId(partiesViewSchemaId)).click();
  const components = rows.slice(1).map((row) => row.record_id);
  await expect
    .poll(() =>
      reads
        .at(-1)
        ?.response.data.rows.map((row) => row.record_id)
        .sort(),
    )
    .toEqual(components.sort());
  await expectLiteralMembership(
    page,
    partiesViewSchemaId,
    components,
    rows.map((row) => row.record_id),
  );
  expect(reads.at(-1)?.response.meta.query.filters[0]?.arg).toEqual({
    values: ["inc.", "northwind"],
  });
});

async function timestampTaskFixture(page: Page) {
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

test("Timestamp drafts correct locally and fractional ranges select exact populated workbook membership", async ({
  page,
}) => {
  const { incident, rows } = await timestampTaskFixture(page);
  const note = await createViewRow(page, incident, notesViewSchemaId, {
    client_txn_id: uniqueTxn("timestamp-note"),
    "note.title": "Timestamp correction",
  });
  const notes = await observeQuery(page, incident, notesViewSchemaId);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${notesViewSchemaId}`,
  );
  await expect(
    page.getByTestId(gridRowTestId(notesViewSchemaId, note.record_id)),
  ).toBeVisible();
  await page
    .getByTestId(workbookFilterPopoverTriggerTestId(notesViewSchemaId))
    .click();
  await page
    .getByTestId(gridFilterFieldTestId(notesViewSchemaId))
    .selectOption("note.updated_at");
  const value = page.getByTestId(gridFilterValueTestId(notesViewSchemaId));
  const applyNote = page.getByTestId(gridFilterApplyTestId(notesViewSchemaId));
  const before = notes.length;
  await value.fill(" tomorrow ");
  await expect(applyNote).toBeDisabled();
  await expect(value).toHaveValue(" tomorrow ");
  await expect(value).toHaveAttribute("aria-invalid", "true");
  await expect(page.getByRole("dialog")).toContainText(
    "numeric timezone offset",
  );
  expect(notes.length).toBe(before);
  const updated = note.cells["note.updated_at"]?.value;
  if (typeof updated !== "string")
    throw new Error("Missing authoritative Note timestamp");
  await value.fill(updated);
  await applyNote.click();
  await expect
    .poll(() => notes.at(-1)?.response.meta.query.filters[0]?.arg)
    .toEqual({ value: updated });
  await expect(
    page.getByTestId(gridRowTestId(notesViewSchemaId, note.record_id)),
  ).toBeVisible();
  const reads = await observeQuery(page, incident, taskRequestsViewSchemaId);
  await switchOrdinarySheet(page, taskRequestsViewSchemaId);
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  await expect(browsing).toContainText(
    "5 records loaded; end of current results.",
  );
  const trigger = page.getByTestId(
    workbookFilterPopoverTriggerTestId(taskRequestsViewSchemaId),
  );
  await trigger.click();
  await page
    .getByTestId(gridFilterFieldTestId(taskRequestsViewSchemaId))
    .selectOption("task.due_at");
  await page
    .getByTestId(workbookFilterOperatorTestId(taskRequestsViewSchemaId))
    .selectOption("range");
  const lower = page.getByRole("textbox", {
    name: "Lower-bound value",
    exact: true,
  });
  const upper = page.getByRole("textbox", {
    name: "Upper-bound value",
    exact: true,
  });
  const apply = page.getByTestId(
    gridFilterApplyTestId(taskRequestsViewSchemaId),
  );
  await lower.fill("2026-04-18T00:00:00Z");
  await upper.fill("2026-04-18T00:00:00.1Z");
  await apply.click();
  const expected = rows
    .slice(1, 4)
    .map((row) => row.record_id)
    .sort();
  await expect
    .poll(() =>
      reads
        .at(-1)
        ?.response.data.rows.map((row) => row.record_id)
        .sort(),
    )
    .toEqual(expected);
  const chip = page.getByTestId(
    workbookQueryEntryTestId(taskRequestsViewSchemaId, "filter", "task.due_at"),
  );
  await expect(chip).toContainText("2026-04-18T00:00:00.1Z");
  for (const row of rows)
    await expect(
      page.getByTestId(gridRowTestId(taskRequestsViewSchemaId, row.record_id)),
    ).toHaveCount(expected.includes(row.record_id) ? 1 : 0);
  await chip.click();
  const acceptedReads = reads.length;
  await lower.fill("2026-04-18T00:00:00.1Z");
  await upper.fill("2026-04-18T00:00:00Z");
  await expect(apply).toBeDisabled();
  expect(reads.length).toBe(acceptedReads);
  await expect(lower).toHaveAttribute("aria-invalid", "true");
  await lower.fill("2026-04-18T00:00:00Z");
  await upper.fill("2026-04-18T00:00:00.1Z");
  await page
    .getByLabel("Lower-bound comparison", { exact: true })
    .selectOption("gt");
  await page
    .getByLabel("Upper-bound comparison", { exact: true })
    .selectOption("lt");
  await apply.click();
  await expect
    .poll(() => reads.at(-1)?.response.data.rows.map((row) => row.record_id))
    .toEqual([rows[2]?.record_id]);
  // A genuine failed read still retains its accepted row and chips.
  await chip.click();
  await upper.fill("2026-04-18T00:00:00.2Z");
  await page.route(
    `**/incidents/${incident}/views/${taskRequestsViewSchemaId}/query`,
    (route) => route.abort("failed"),
    { times: 1 },
  );
  await apply.click();
  await expect(
    browsing.getByRole("button", { name: "Retry", exact: true }),
  ).toBeVisible();
  await expect(chip).toContainText("2026-04-18T00:00:00.1Z");
  await expect(browsing).toContainText("1 records loaded");
  await browsing.getByRole("button", { name: "Revert", exact: true }).click();
  await page.setViewportSize({ width: 768, height: 640 });
  await trigger.click();
  await page.getByRole("button", { name: /^Edit Filter 1, Due/ }).click();
  await lower.fill("tomorrow");
  await expect(apply).toBeDisabled();
  await lower.focus();
  await expect(lower).toBeInViewport();
  await expect(page.getByRole("dialog")).toContainText("2026-04-18T00:00:00Z");
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
});

test("Timestamp saved queries reload and reopen without losing fractional instant meaning", async ({
  page,
}) => {
  const { incident } = await timestampTaskFixture(page);
  const reads = await observeQuery(page, incident, taskRequestsViewSchemaId);
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${taskRequestsViewSchemaId}`,
  );
  await expect(
    page.getByRole("group", { name: "Workbook browsing" }),
  ).toContainText("5 records loaded");
  await page
    .getByTestId(workbookFilterPopoverTriggerTestId(taskRequestsViewSchemaId))
    .click();
  await page
    .getByTestId(gridFilterFieldTestId(taskRequestsViewSchemaId))
    .selectOption("task.due_at");
  await page
    .getByTestId(workbookFilterOperatorTestId(taskRequestsViewSchemaId))
    .selectOption("range");
  await page
    .getByLabel("Lower-bound value", { exact: true })
    .fill("2026-04-17T19:59:59.999999999-04:00");
  await page
    .getByLabel("Upper-bound value", { exact: true })
    .fill("2026-04-18T00:00:00.1Z");
  await page
    .getByTestId(gridFilterApplyTestId(taskRequestsViewSchemaId))
    .click();
  const arg = {
    gte: "2026-04-17T23:59:59.999999999Z",
    lte: "2026-04-18T00:00:00.1Z",
  };
  await expect
    .poll(() => reads.at(-1)?.response.meta.query.filters[0]?.arg)
    .toEqual(arg);
  const saved = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/incidents/${incident}/saved-views`),
  );
  await setSavedViewDraftName(
    page,
    taskRequestsViewSchemaId,
    "Precise timestamp window",
  );
  await createSavedViewFromCurrentSurface(page, taskRequestsViewSchemaId);
  const response = await saved;
  expect(response.ok()).toBe(true);
  const resource = (await response.json()).data;
  expect(resource.query_json.filters).toEqual([
    { field_key: "task.due_at", op: "range", arg },
  ]);
  await page.reload();
  await expect(
    page.getByRole("group", { name: "Workbook browsing" }),
  ).toContainText("3 records loaded");
  await page
    .getByTestId(
      workbookQueryEntryTestId(
        taskRequestsViewSchemaId,
        "filter",
        "task.due_at",
      ),
    )
    .click();
  await expect(
    page.getByLabel("Lower-bound value", { exact: true }),
  ).toHaveValue(arg.gte);
  await expect(
    page.getByLabel("Upper-bound value", { exact: true }),
  ).toHaveValue(arg.lte);
  await expect(
    page.getByTestId(gridFilterApplyTestId(taskRequestsViewSchemaId)),
  ).toBeEnabled();
  await page.keyboard.press("Escape");
  // Model a resource read from before the fix; the integration fixture separately
  // persists this older predicate in PostgreSQL and verifies it is never repaired.
  const savedPath = `/incidents/${incident}/saved-views/${resource.saved_view_id}`;
  const older = {
    ...resource,
    query_json: {
      ...resource.query_json,
      filters: [
        {
          field_key: "task.due_at",
          op: "range",
          arg: {
            gte: "2026-04-18T00:00:00.1Z",
            lte: "2026-04-18T00:00:00Z",
          },
        },
      ],
    },
  };
  // Startup carries the selected resource. Keep all read paths consistent until
  // the explicit correction, including repeated reads after cancellation.
  const savedReads = `**/incidents/${incident}/saved-views**`;
  const startupRead = `**/incidents/${incident}/workbook-startup**`;
  const olderRead = async (route: Route) => {
    if (route.request().method() !== "GET") return route.continue();
    const response = await route.fetch();
    const envelope = await response.json();
    const data =
      "selected_saved_view" in envelope.data
        ? { ...envelope.data, selected_saved_view: older }
        : Array.isArray(envelope.data.saved_views)
          ? {
              ...envelope.data,
              saved_views: envelope.data.saved_views.map(
                (view: { saved_view_id: string }) =>
                  view.saved_view_id === resource.saved_view_id ? older : view,
              ),
            }
          : envelope.data.saved_view_id === resource.saved_view_id
            ? older
            : envelope.data;
    await route.fulfill({ response, json: { ...envelope, data } });
  };
  await page.route(savedReads, olderRead);
  await page.route(startupRead, olderRead);
  const rejected = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/views/${taskRequestsViewSchemaId}/query`) &&
      response.status() === 400,
  );
  await page.reload();
  await rejected;
  await page
    .getByTestId(workbookFilterPopoverTriggerTestId(taskRequestsViewSchemaId))
    .click();
  await page.getByRole("button", { name: /^Edit Filter 1, Due/ }).click();
  const lower = page.getByLabel("Lower-bound value", { exact: true });
  const upper = page.getByLabel("Upper-bound value", { exact: true });
  await expect(lower).toHaveValue("2026-04-18T00:00:00.1Z");
  await expect(upper).toHaveValue("2026-04-18T00:00:00Z");
  await expect(
    page.getByTestId(gridFilterApplyTestId(taskRequestsViewSchemaId)),
  ).toBeDisabled();
  // Choose a correction distinct from the valid backing fixture so this is a
  // real versioned write, not an unchanged PATCH against the underlying row.
  const corrected = { gte: "2026-04-18T00:00:00Z", lte: arg.lte };
  await lower.fill(corrected.gte);
  await upper.fill(corrected.lte);
  await page
    .getByTestId(gridFilterApplyTestId(taskRequestsViewSchemaId))
    .click();
  await expect(
    page.getByRole("group", { name: "Workbook browsing" }),
  ).toContainText("3 records loaded");
  await page.unroute(savedReads, olderRead);
  await page.unroute(startupRead, olderRead);
  const updated = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(savedPath),
  );
  await updateSavedViewFromCurrentSurface(
    page,
    taskRequestsViewSchemaId,
    resource.saved_view_id,
  );
  const updateResponse = await updated;
  expect(updateResponse.ok()).toBe(true);
  const updatedResource = (await updateResponse.json()).data;
  expect(updatedResource.query_json.filters).toEqual([
    { field_key: "task.due_at", op: "range", arg: corrected },
  ]);
  expect(updatedResource.saved_view_version).toBeGreaterThan(
    resource.saved_view_version,
  );
});

test("Boolean filters preserve typed saved and accepted operands across matching modes", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("BOOL-QUERY"),
    "Typed boolean filters",
  );
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("bool-timeline"),
    "timeline.activity_synopsis_text": "Boolean fixture without evidence",
  });
  await createViewRow(page, incident, taskRequestsViewSchemaId, {
    client_txn_id: uniqueTxn("bool-task"),
    "task.title": "Boolean task fixture",
    "task.task_kind": "question",
  });
  const field = "timeline.has_evidence";
  const savedSet = await createSavedView(page, incident, {
    display_name: "Boolean set",
    view_schema_id: timelineViewSchemaId,
    query_json: {
      filters: [
        { field_key: field, op: "eq", arg: { values: [true, false, true] } },
      ],
      sort: [],
    },
  });
  const savedNull = await createSavedView(page, incident, {
    display_name: "Boolean empty",
    view_schema_id: timelineViewSchemaId,
    query_json: {
      filters: [{ field_key: field, op: "eq", arg: { value: null } }],
      sort: [],
    },
  });
  const lifecycle = await currentLifecycle(page, incident);
  expect(
    (
      await lifecycleAction(page, incident, "closeIncident", {
        client_txn_id: uniqueTxn("bool-close"),
        base_incident_version: lifecycle.incident_version,
        reason: "Boolean filtering requires read access only",
      })
    ).ok,
  ).toBe(true);
  const reads = await observeQuery(page, incident, timelineViewSchemaId);
  let requests = 0;
  page.on("request", (request) => {
    if (request.url().endsWith(`/views/${timelineViewSchemaId}/query`))
      requests++;
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${timelineViewSchemaId}`,
  );
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  await expect(browsing).toContainText(
    "1 records loaded; end of current results.",
  );
  const trigger = page.getByTestId(
    workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
  );
  const apply = page.getByTestId(gridFilterApplyTestId(timelineViewSchemaId));
  const mode = page.getByRole("combobox", {
    name: "Equality operand kind",
    exact: true,
  });
  const chip = page.getByTestId(
    workbookQueryEntryTestId(timelineViewSchemaId, "filter", field),
  );
  const accept = async (arg: Record<string, unknown>, count: number) => {
    await expect
      .poll(() => reads.at(-1)?.request.filters?.[0]?.arg)
      .toEqual(arg);
    await expect
      .poll(() => reads.at(-1)?.response.meta.query.filters[0]?.arg)
      .toEqual(arg);
    await expect(browsing).toContainText(
      `${count} records loaded; end of current results.`,
    );
  };
  await trigger.click();
  await page
    .getByTestId(gridFilterFieldTestId(timelineViewSchemaId))
    .selectOption(field);
  const before = requests;
  await expect(
    page.getByTestId(gridFilterValueTestId(timelineViewSchemaId)),
  ).toHaveValue("");
  await expect(apply).toBeDisabled();
  await mode.selectOption("values");
  const trueChoice = page.getByRole("checkbox", { name: "true", exact: true });
  const falseChoice = page.getByRole("checkbox", {
    name: "false",
    exact: true,
  });
  await expect(apply).toBeDisabled();
  await trueChoice.focus();
  await page.keyboard.press("Space");
  await page.keyboard.press("Tab");
  await expect(falseChoice).toBeFocused();
  await page.keyboard.press("Shift+Tab");
  await expect(trueChoice).toBeFocused();
  expect(requests).toBe(before);
  await apply.click();
  await accept({ values: [true] }, 0);
  await expect(trigger).toBeFocused();
  await chip.click();
  await expect(trueChoice).toBeChecked();
  await trueChoice.uncheck();
  await expect(apply).toBeDisabled();
  await falseChoice.check();
  await apply.click();
  await accept({ values: [false] }, 1);
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, row.record_id)),
  ).toBeVisible();
  await chip.click();
  await falseChoice.uncheck();
  await trueChoice.check();
  await page.route(
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    (route) => route.abort("failed"),
    { times: 1 },
  );
  await apply.click();
  await expect(
    browsing.getByRole("button", { name: "Retry", exact: true }),
  ).toBeVisible();
  await expect(chip).toContainText("false");
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, row.record_id)),
  ).toBeVisible();
  await trigger.click();
  await page
    .getByRole("button", { name: /Edit unapplied.*Has Evidence/i })
    .click();
  await expect(trueChoice).toBeChecked();
  await expect(falseChoice).not.toBeChecked();
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  await browsing.getByRole("button", { name: "Revert", exact: true }).click();
  await selectSavedView(page, timelineViewSchemaId, savedSet.saved_view_id);
  // The server owns deduplication and false-before-true canonical ordering.
  await expect
    .poll(() => reads.at(-1)?.response.meta.query.filters[0]?.arg)
    .toEqual({ values: [false, true] });
  await chip.click();
  await expect(trueChoice).toBeChecked();
  await expect(falseChoice).toBeChecked();
  await apply.click();
  await accept({ values: [false, true] }, 1);
  await selectSavedView(page, timelineViewSchemaId, savedNull.saved_view_id);
  await accept({ value: null }, 0);
  await chip.click();
  await expect(mode).toHaveValue("null");
  await mode.selectOption("value");
  const scalar = page.getByTestId(gridFilterValueTestId(timelineViewSchemaId));
  await expect(scalar).toHaveValue("");
  await expect(apply).toBeDisabled();
  await scalar.selectOption("false");
  await apply.click();
  await accept({ value: false }, 1);
  await page.setViewportSize({ width: 768, height: 640 });
  await trigger.click();
  await page
    .getByRole("button", { name: /^Edit Filter 1, Has Evidence/ })
    .click();
  await mode.selectOption("values");
  await expect(apply).toBeDisabled();
  await falseChoice.check();
  await apply.focus();
  await expect(apply).toBeInViewport();
  await page.keyboard.press("Shift+Tab");
  await expect(
    page.getByRole("button", { name: "Cancel", exact: true }),
  ).toBeFocused();
  await page.keyboard.press("Escape");
  await expect(trigger).toBeFocused();
  await page.setViewportSize({ width: 1440, height: 900 });
  const tasks = await observeQuery(page, incident, taskRequestsViewSchemaId);
  await switchOrdinarySheet(page, taskRequestsViewSchemaId);
  await expect(browsing).toContainText(
    "1 records loaded; end of current results.",
  );
  await page
    .getByTestId(workbookFilterPopoverTriggerTestId(taskRequestsViewSchemaId))
    .click();
  await page
    .getByTestId(gridFilterFieldTestId(taskRequestsViewSchemaId))
    .selectOption("task.no_owner");
  const taskValue = page.getByTestId(
    gridFilterValueTestId(taskRequestsViewSchemaId),
  );
  await expect(taskValue).toHaveValue("");
  await taskValue.focus();
  await page.keyboard.press("ArrowDown");
  await expect(taskValue).toHaveValue("true");
  await page
    .getByTestId(gridFilterApplyTestId(taskRequestsViewSchemaId))
    .click();
  await expect
    .poll(() => tasks.at(-1)?.request.filters?.[0]?.arg)
    .toEqual({ value: true });
  await expect
    .poll(() => tasks.at(-1)?.response.meta.query.filters[0]?.arg)
    .toEqual({ value: true });
  await expect(
    browsing.getByRole("button", { name: "Retry", exact: true }),
  ).toHaveCount(0);
});

test("Enum equality choices remain explicit and preserve custom queries", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ENUM-QUERY"),
    "Enum query choices",
  );
  const row = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("enum-timeline"),
    "timeline.activity_synopsis_text": "Enum filter fixture",
  });
  await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("enum-party"),
    "party.display_name": "Enum reuse fixture",
    "party.party_kind": "person",
  });
  const before = await currentLifecycle(page, incident);
  expect(
    (
      await lifecycleAction(page, incident, "closeIncident", {
        client_txn_id: uniqueTxn("enum-close"),
        base_incident_version: before.incident_version,
        reason: "Read-only enum query evidence",
      })
    ).ok,
  ).toBe(true);
  const endpoint = `/incidents/${incident}/views/${timelineViewSchemaId}/query`;
  const requests: QueryWorkbookViewRequest[] = [];
  const accepted: QueryWorkbookViewResponse[] = [];
  let failNext = false;
  await page.route(`**${endpoint}`, async (route) => {
    requests.push(route.request().postDataJSON() as QueryWorkbookViewRequest);
    if (failNext) {
      failNext = false;
      await route.abort("failed");
      return;
    }
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    accepted.push((await response.json()) as QueryWorkbookViewResponse);
    await route.fulfill({ response });
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${timelineViewSchemaId}`,
  );
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  await expect(browsing).toContainText(
    "1 records loaded; end of current results.",
  );
  const trigger = page.getByTestId(
    workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
  );
  const apply = page.getByTestId(gridFilterApplyTestId(timelineViewSchemaId));
  const value = page.getByTestId(gridFilterValueTestId(timelineViewSchemaId));
  const field = "timeline.activity_time_pair_state";
  const start = requests.length;
  await trigger.click();
  await page
    .getByTestId(gridFilterFieldTestId(timelineViewSchemaId))
    .selectOption(field);
  await expect(value).toHaveValue("");
  await expect(value.getByRole("option")).toHaveText([
    "Choose a value",
    ...(requireViewContract(timelineViewSchemaId).fieldMap[field]?.enumValues ??
      []),
  ]);
  await value.focus();
  await page.keyboard.press("ArrowDown");
  await expect(value).toHaveValue("disabled");
  expect(requests.length).toBe(start);
  await apply.click();
  await expect(browsing).toContainText(
    "1 records loaded; end of current results.",
  );
  await expect
    .poll(() => requests.at(-1)?.filters)
    .toEqual([{ field_key: field, op: "eq", arg: { value: "disabled" } }]);
  await expect
    .poll(() => accepted.at(-1)?.meta.query.filters)
    .toEqual([{ field_key: field, op: "eq", arg: { value: "disabled" } }]);
  await expect(trigger).toBeFocused();
  const chip = () =>
    page.getByRole("button", { name: /^Filter 1, Activity Time Pair State/ });
  await chip().click();
  await page
    .getByRole("combobox", { name: "Equality operand kind", exact: true })
    .selectOption("values");
  await page.getByRole("checkbox", { name: "disabled", exact: true }).focus();
  await page.keyboard.press("Space");
  await page.getByRole("checkbox", { name: "empty", exact: true }).check();
  await apply.click();
  await expect
    .poll(() => requests.at(-1)?.filters?.[0]?.arg)
    .toEqual({ values: ["disabled", "empty"] });
  await expect(browsing).toContainText(
    "1 records loaded; end of current results.",
  );
  await chip().click();
  await page
    .getByRole("button", { name: "Custom literals", exact: true })
    .click();
  await page.getByRole("button", { name: "Add literal", exact: true }).click();
  await page
    .getByRole("textbox", { name: "Value literal 3", exact: true })
    .fill("Disabled");
  await page.getByRole("button", { name: "Add literal", exact: true }).click();
  await page
    .getByRole("textbox", { name: "Value literal 4", exact: true })
    .fill("custom,token");
  await apply.click();
  await expect
    .poll(() => requests.at(-1)?.filters?.[0]?.arg.values)
    .toEqual(
      expect.arrayContaining(["disabled", "empty", "Disabled", "custom,token"]),
    );
  await expect(browsing).toContainText(
    "1 records loaded; end of current results.",
  );
  await chip().click();
  await expect(
    page.getByRole("button", { name: "Custom literals", exact: true }),
  ).toHaveAttribute("aria-expanded", "true");
  expect(
    await page
      .getByRole("textbox", { name: /^Value literal \d+$/ })
      .evaluateAll((inputs) =>
        inputs.map((input) => (input as HTMLInputElement).value),
      ),
  ).toContain("custom,token");
  await page
    .getByRole("combobox", { name: "Equality operand kind", exact: true })
    .selectOption("value");
  await page
    .getByRole("button", { name: "Custom literals", exact: true })
    .click();
  await page
    .getByRole("textbox", { name: "Value literal", exact: true })
    .fill("disable");
  failNext = true;
  await apply.click();
  await expect(
    browsing.getByRole("button", { name: "Retry", exact: true }),
  ).toBeVisible();
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, row.record_id)),
  ).toBeVisible();
  await expect(chip()).toContainText("custom,token");
  await browsing.getByRole("button", { name: "Revert", exact: true }).click();
  await expect(chip()).toContainText("custom,token");
  await chip().click();
  await page
    .getByRole("combobox", { name: "Equality operand kind", exact: true })
    .selectOption("value");
  await page
    .getByRole("button", { name: "Custom literals", exact: true })
    .click();
  await page
    .getByRole("textbox", { name: "Value literal", exact: true })
    .fill("disable");
  failNext = true;
  await apply.click();
  await expect(
    browsing.getByRole("button", { name: "Retry", exact: true }),
  ).toBeVisible();
  await browsing.getByRole("button", { name: "Retry", exact: true }).click();
  await expect(browsing).toContainText(
    "0 records loaded; end of current results.",
  );
  await expect(chip()).toHaveAccessibleName(
    "Filter 1, Activity Time Pair State, equals disable",
  );
  await chip().click();
  await expect(
    page.getByRole("textbox", { name: "Value literal", exact: true }),
  ).toHaveValue("disable");
  await page
    .getByRole("textbox", { name: "Value literal", exact: true })
    .fill("Disabled");
  await apply.click();
  await expect(browsing).toContainText(
    "1 records loaded; end of current results.",
  );
  await expect
    .poll(() => accepted.at(-1)?.meta.query.filters?.[0]?.arg)
    .toEqual({ value: "Disabled" });
  await page.setViewportSize({ width: 768, height: 640 });
  await trigger.click();
  await page
    .getByRole("button", { name: /^Edit Filter 1, Activity Time Pair State/ })
    .click();
  await page
    .getByRole("combobox", { name: "Equality operand kind", exact: true })
    .selectOption("values");
  await page
    .getByRole("checkbox", { name: "conversion_unavailable", exact: true })
    .check();
  await page
    .getByRole("button", { name: "Custom literals", exact: true })
    .click();
  await page.getByRole("button", { name: "Add literal", exact: true }).click();
  await page
    .getByRole("textbox", { name: "Value literal 2", exact: true })
    .fill("long_custom_literal,".repeat(12));
  const spacing = await page.addStyleTag({
    content: `
    * { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; }
    p { margin-block-end: 2em !important; }
  `,
  });
  await apply.focus();
  await page.keyboard.press("Shift+Tab");
  const cancel = page.getByRole("button", { name: "Cancel", exact: true });
  await expect(cancel).toBeFocused();
  for (const action of [apply, cancel]) {
    const box = await action.boundingBox();
    expect(box).not.toBeNull();
    if (!box) throw new Error("Missing enum action geometry");
    expect(box.x).toBeGreaterThanOrEqual(0);
    expect(box.y).toBeGreaterThanOrEqual(0);
    expect(box.x + box.width).toBeLessThanOrEqual(768);
    expect(box.y + box.height).toBeLessThanOrEqual(640);
  }
  await test.info().attach("enum-filter-text-spacing", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await cancel.click();
  await spacing.evaluate((element) => element.parentNode?.removeChild(element));
  await page.setViewportSize({ width: 1440, height: 900 });
  await switchOrdinarySheet(page, partiesViewSchemaId);
  const partyReads = await observeQuery(page, incident, partiesViewSchemaId);
  await page
    .getByTestId(workbookFilterPopoverTriggerTestId(partiesViewSchemaId))
    .click();
  await page
    .getByTestId(gridFilterFieldTestId(partiesViewSchemaId))
    .selectOption("party.party_kind");
  await page
    .getByTestId(gridFilterValueTestId(partiesViewSchemaId))
    .selectOption("person");
  await page.getByTestId(gridFilterApplyTestId(partiesViewSchemaId)).click();
  await expect
    .poll(() => partyReads.at(-1)?.request.filters?.[0]?.arg)
    .toEqual({ value: "person" });
  await expect(browsing).toContainText(
    "1 records loaded; end of current results.",
  );
});
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

const queryObserverCleanup = new WeakMap<Page, Array<() => Promise<void>>>();

test.afterEach(async ({ page }) => {
  // Finish only this observer's work before Playwright disposes fetched bodies.
  // Deliberately held routes in other scenarios keep their own lifecycle.
  for (const dispose of queryObserverCleanup.get(page) ?? []) await dispose();
  queryObserverCleanup.delete(page);
});

async function observeQuery(page: Page, incident: string, view: string) {
  const reads: {
    request: QueryWorkbookViewRequest;
    response: QueryWorkbookViewResponse;
  }[] = [];
  const pending = new Set<Promise<void>>();
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
  const handler = async (route: Route) => {
    const completion = read(route);
    pending.add(completion);
    try {
      await completion;
    } finally {
      pending.delete(completion);
    }
  };
  await page.route(path, handler);
  const cleanup = queryObserverCleanup.get(page) ?? [];
  cleanup.push(async () => {
    await page.unroute(path, handler);
    await Promise.all(pending);
  });
  queryObserverCleanup.set(page, cleanup);
  return reads;
}

async function reachButtonByTab(page: Page, button: Locator) {
  if (
    await page.evaluate(() =>
      Boolean(document.activeElement?.closest('[role="grid"]')),
    )
  )
    await page.keyboard.press("Control+End");
  for (let step = 0; step < 160; step++) {
    if (await button.evaluate((element) => document.activeElement === element))
      return;
    await page.keyboard.press("Tab");
  }
  throw new Error("Keyboard Tab did not reach the workbook browsing action");
}

async function expectVisibleKeyboardFocus(button: Locator) {
  await expect(button).toBeFocused();
  expect(
    await button.evaluate(
      (element) =>
        element.matches(":focus-visible") &&
        getComputedStyle(element).outlineStyle !== "none" &&
        Number.parseFloat(getComputedStyle(element).outlineWidth) > 0,
    ),
  ).toBe(true);
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
  const url = `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`;
  const fixtureSocket = installIncidentSocketMonitor(page, incident);
  await page.goto(url);
  await fixtureSocket.waitForAcceptedSocket();
  // A committed fixture write may still await collaboration sequencing. Observe
  // each publication before measuring a fresh page's bounded startup reads.
  const fixture = await seed(page, incident, view, count, actorId, (recordId) =>
    fixtureSocket.waitForMessage("record_changed", {
      matches: (message) => message.payload.record_id === recordId,
    }),
  );
  await page.goto("about:blank");
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
  await page
    .getByTestId(workbookFilterPopoverTriggerTestId(notesViewSchemaId))
    .click();
  await page
    .getByTestId(gridFilterFieldTestId(notesViewSchemaId))
    .selectOption("note.tags");
  await authorLiteralMembers(page.getByRole("dialog"), [
    " beta ",
    "alpha",
    "beta",
  ]);
  await page.getByTestId(gridFilterApplyTestId(notesViewSchemaId)).click();
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

test("Workbook query recovery retains keyboard focus through Timeline replacement and continuation", async ({
  page,
  workerAdmin,
}) => {
  test.setTimeout(180_000);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-FOCUS"),
    "Workbook query recovery focus",
  );
  await seed(page, incident, timelineViewSchemaId, 105, workerAdmin.user_id);
  const writes: string[] = [];
  page.on("request", (request) => {
    if (
      ["PATCH", "PUT", "DELETE"].includes(request.method()) &&
      request.url().includes("/records/")
    )
      writes.push(`${request.method()} ${request.url()}`);
    if (
      request.method() === "POST" &&
      request.url().includes(`/incidents/${incident}/views/`) &&
      request.url().endsWith("/rows")
    )
      writes.push(`POST ${request.url()}`);
  });
  const reads: QueryWorkbookViewRequest[] = [];
  let next: "ordinary" | "fail" | "hold_failure" | "hold_success" = "ordinary";
  let release = () => {};
  let gate = Promise.resolve();
  const hold = (outcome: "hold_failure" | "hold_success") => {
    gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    next = outcome;
  };
  await page.route(
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      const request = route
        .request()
        .postDataJSON() as QueryWorkbookViewRequest;
      const behavior = next;
      next = "ordinary";
      reads.push(request);
      if (behavior === "fail") {
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      if (behavior === "hold_failure" || behavior === "hold_success") {
        await gate;
        if (behavior === "hold_failure") {
          await route.abort("failed");
          return;
        }
      }
      await route.fulfill({ response });
    },
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
  );
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  const retry = controls.getByRole("button", { name: "Retry", exact: true });
  const revert = controls.getByRole("button", { name: "Revert", exact: true });
  const refresh = controls.getByRole("button", {
    name: "Refresh",
    exact: true,
  });
  const more = controls.getByRole("button", { name: "Load more", exact: true });
  await expect(controls).toContainText("100 records loaded; more available.");
  try {
    next = "fail";
    await sortByHeader(
      page,
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    );
    await expect(retry).toBeVisible();
    await expect(revert).toBeVisible();
    const failedReplacement = reads.at(-1);
    expect(failedReplacement?.cursor_token).toBeUndefined();
    expect(failedReplacement?.limit).toBe(100);
    expect(failedReplacement?.sort?.[0]?.field_key).toBe(
      "timeline.activity_synopsis_text",
    );
    await expect(controls).toContainText("100 records loaded");
    await reachButtonByTab(page, retry);
    hold("hold_failure");
    const firstRetry = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(firstRetry + 1);
    expect(reads.at(-1)).toEqual(failedReplacement);
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-busy", "true");
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    await expect(controls).toContainText("Retrying records…");
    await expect(controls).not.toContainText("Query changes are unapplied.");
    await page.keyboard.press("Enter");
    expect(reads.length).toBe(firstRetry + 1);
    release();
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-disabled", "false");
    await expect(controls).toContainText("Query changes are unapplied.");

    hold("hold_success");
    const secondRetry = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(secondRetry + 1);
    expect(reads.at(-1)).toEqual(failedReplacement);
    await expect(retry).toBeFocused();
    release();
    await expect(controls).toContainText("100 records loaded; more available.");
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    await page.keyboard.press("Tab");
    await expect(retry).toHaveCount(0);

    await reachButtonByTab(page, more);
    next = "fail";
    await page.keyboard.press("Enter");
    await expect(retry).toBeVisible();
    const failedContinuation = reads.at(-1);
    expect(failedContinuation?.cursor_token).toBeTruthy();
    expect(failedContinuation?.limit).toBe(100);
    expect(failedContinuation?.sort).toEqual(failedReplacement?.sort);
    await expect(controls).toContainText("100 records loaded");
    await reachButtonByTab(page, retry);
    hold("hold_success");
    const continuationRetry = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(continuationRetry + 1);
    expect(reads.at(-1)).toEqual(failedContinuation);
    await expectVisibleKeyboardFocus(retry);
    await expect(controls).toContainText("Retrying records…");
    await page.keyboard.press("Shift+Tab");
    await expect(refresh).toBeFocused();
    release();
    await expect(controls).toContainText(
      "105 records loaded; end of current results.",
    );
    await expect(refresh).toBeFocused();
    await expect(retry).toHaveCount(0);

    next = "fail";
    await sortByHeader(
      page,
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    );
    await expect(revert).toBeVisible();
    await reachButtonByTab(page, revert);
    hold("hold_success");
    const reverted = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(reverted + 1);
    await expectVisibleKeyboardFocus(revert);
    await expect(revert).toHaveAttribute("aria-disabled", "true");
    release();
    await expect(controls).toContainText(
      "105 records loaded; end of current results.",
    );
    await expectVisibleKeyboardFocus(revert);
    await page.keyboard.press("Tab");
    await expect(revert).toHaveCount(0);
    expect(writes).toEqual([]);
  } finally {
    release();
  }
});

test("Workbook query recovery keeps keyboard focus with accepted-empty Notes", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-EMPTY-FOCUS"),
    "Workbook empty recovery focus",
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(notesViewSchemaId)}`,
  );
  const controls = page.getByRole("group", { name: "Workbook browsing" });
  await expect(controls).toContainText(
    "0 records loaded; end of current results.",
  );
  let next: "ordinary" | "fail" | "hold" = "ordinary";
  let release = () => {};
  let gate = Promise.resolve();
  const reads: QueryWorkbookViewRequest[] = [];
  await page.route(
    `**/incidents/${incident}/views/${notesViewSchemaId}/query`,
    async (route) => {
      const request = route
        .request()
        .postDataJSON() as QueryWorkbookViewRequest;
      const behavior = next;
      next = "ordinary";
      reads.push(request);
      if (behavior === "fail") {
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      if (behavior === "hold") await gate;
      await route.fulfill({ response });
    },
  );
  const refresh = controls.getByRole("button", {
    name: "Refresh",
    exact: true,
  });
  const retry = controls.getByRole("button", { name: "Retry", exact: true });
  try {
    await reachButtonByTab(page, refresh);
    next = "fail";
    await page.keyboard.press("Enter");
    await expect(retry).toBeVisible();
    const failed = reads.at(-1);
    expect(failed?.cursor_token).toBeUndefined();
    await expect(controls).toContainText("0 records loaded");
    await reachButtonByTab(page, retry);
    gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    next = "hold";
    const count = reads.length;
    await page.keyboard.press("Enter");
    await expect.poll(() => reads.length).toBe(count + 1);
    expect(reads.at(-1)).toEqual(failed);
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-busy", "true");
    await expect(controls).toContainText("0 records loaded. Retrying records…");
    release();
    await expect(controls).toContainText(
      "0 records loaded; end of current results.",
    );
    await expectVisibleKeyboardFocus(retry);
    await expect(retry).toHaveAttribute("aria-disabled", "true");
    await page.keyboard.press("Tab");
    await expect(retry).toHaveCount(0);
  } finally {
    release();
  }
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

async function dateFilterCorrection(page: Page, view: string) {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incident = await createIncident(
    page,
    uniqueIncidentKey("DATE-QUERY"),
    "Date filter correction",
  );
  const fixture = fixtureFields[view];
  if (!fixture) throw new Error("Missing date query fixture");
  const dateField = requireViewContract(view).fields.find(
    (field) => field.readKind === "date" && field.filterOps.includes("eq"),
  )?.fieldKey;
  if (!dateField) throw new Error("Missing declared date filter");
  const rows = [];
  for (const day of ["18", "19"]) {
    rows.push(
      await createViewRow(page, incident, view, {
        client_txn_id: uniqueTxn(`date-query-${day}`),
        ...fixture.minimum,
        [fixture.field]: `Date correction ${day}`,
        ...(view === timelineViewSchemaId
          ? { "timeline.date_entered_text": `2026-04-${day}` }
          : { "comm_log.timestamp_utc": `2026-04-${day}T12:00:00Z` }),
      }),
    );
  }
  const first = rows[0];
  const second = rows[1];
  if (!first || !second) throw new Error("Missing seeded date rows");
  const endpoint = `/incidents/${incident}/views/${view}/query`;
  // Observe admission when requests begin, including attempts that fail or abort.
  const requests: QueryWorkbookViewRequest[] = [];
  page.on("request", (request) => {
    if (request.method() === "POST" && request.url().endsWith(endpoint))
      requests.push(request.postDataJSON() as QueryWorkbookViewRequest);
  });
  let failNext = false;
  await page.route(`**${endpoint}`, async (route) => {
    if (failNext) {
      failNext = false;
      await route.abort("failed");
      return;
    }
    await route.fulfill({ response: await route.fetch() });
  });
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  await expect(browsing).toContainText(
    "2 records loaded; end of current results.",
  );
  const firstRow = page.getByTestId(gridRowTestId(view, first.record_id));
  const secondRow = page.getByTestId(gridRowTestId(view, second.record_id));
  await expect(firstRow).toBeVisible();
  await expect(secondRow).toBeVisible();
  let heldDraft: Awaited<ReturnType<typeof holdBrowserRequest>> | undefined;
  try {
    if (view === timelineViewSchemaId) {
      heldDraft = await holdBrowserRequest(page, {
        method: "POST",
        path: `/api/v1/incidents/${incident}/views/${view}/rows`,
      });
      await openTimelineInspector(page, first.record_id);
      await scrollGridTargetIntoView({
        page,
        surface: view,
        targetTestId: draftCellTestId("timeline.analyst_text"),
      });
      await page
        .getByRole("textbox", { name: "Analyst draft row", exact: true })
        .fill("Independent raw draft");
      await heldDraft.waitForHit;
      await page
        .getByRole("checkbox", {
          name: "Select all loaded records",
          exact: true,
        })
        .check();
    } else await openGenericInspectorForRecord(page, view, first.record_id);
    const anchor = page.getByTestId(workbookFocusAnchorTestId());
    const previousAnchor = await anchor.textContent();
    const trigger = page.getByTestId(workbookFilterPopoverTriggerTestId(view));
    await trigger.click();
    await page.getByTestId(gridFilterFieldTestId(view)).selectOption(dateField);
    const value = page.getByRole("textbox", {
      name: "Date value",
      exact: true,
    });
    const apply = page.getByTestId(gridFilterApplyTestId(view));
    const count = requests.length;
    await value.fill(" 2026-04-31 ");
    await page.setViewportSize({ width: 768, height: 640 });
    const bounds = await page
      .getByRole("dialog", { name: "Add filter" })
      .boundingBox();
    if (!bounds) throw new Error("Missing filter editor bounds");
    expect(bounds.x).toBeGreaterThanOrEqual(0);
    expect(bounds.x + bounds.width).toBeLessThanOrEqual(768);
    await page.setViewportSize({ width: 1440, height: 900 });
    await expect(apply).toBeDisabled();
    await expect(value).toHaveAttribute("aria-invalid", "true");
    const feedback = await value.getAttribute("aria-describedby");
    expect(feedback).toBeTruthy();
    await expect(page.locator(`[id="${feedback}"]`)).toContainText(
      "YYYY-MM-DD",
    );
    await value.press("Enter");
    await value.press("Tab");
    await expect(
      page.getByRole("dialog", { name: "Add filter" }),
    ).toBeVisible();
    await expect(value).toHaveValue(" 2026-04-31 ");
    await expect(anchor).toHaveText(previousAnchor ?? "");
    await expect(firstRow).toBeVisible();
    await expect(secondRow).toBeVisible();
    await expect(
      page.getByTestId(workbookInspectorCloseButtonTestId(view)),
    ).toBeVisible();
    if (view === timelineViewSchemaId) {
      await expect(
        page.getByRole("textbox", { name: "Analyst draft row", exact: true }),
      ).toHaveValue("Independent raw draft");
      await expect(
        page.getByText("2 records selected.", { exact: true }),
      ).toBeVisible();
    }
    expect(requests).toHaveLength(count);
    await value.fill("2026-04-18");
    await expect(apply).toBeEnabled();
    await expect(value).not.toHaveAttribute("aria-invalid");
    await value.press("Tab");
    await page.keyboard.press("Tab");
    await expect(apply).toBeFocused();
    const acceptedResponse = page.waitForResponse(
      (response) =>
        response.url().endsWith(endpoint) && response.status() === 200,
    );
    await page.keyboard.press("Enter");
    await acceptedResponse;
    await expect(page.getByRole("dialog")).toHaveCount(0);
    await expect(trigger).toBeFocused();
    await expect(browsing).toContainText(
      "1 records loaded; end of current results.",
    );
    await expect(firstRow).toBeVisible();
    await expect(secondRow).toHaveCount(0);
    const chip = page.getByTestId(
      workbookQueryEntryTestId(view, "filter", dateField),
    );
    await expect(chip).toContainText("2026-04-18");
    await chip.click();
    await page
      .getByTestId(workbookFilterOperatorTestId(view))
      .selectOption("range");
    const lower = page.getByRole("textbox", {
      name: "Lower-bound value",
      exact: true,
    });
    const upper = page.getByRole("textbox", {
      name: "Upper-bound value",
      exact: true,
    });
    await lower.fill("2026-04-19");
    await upper.fill("2026-04-18");
    const rangeCount = requests.length;
    await expect(apply).toBeDisabled();
    await expect(lower).toHaveAttribute("aria-invalid", "true");
    await expect(upper).toHaveAttribute(
      "aria-describedby",
      (await lower.getAttribute("aria-describedby")) ?? "",
    );
    await lower.press("Enter");
    await expect(chip).toContainText("2026-04-18");
    await expect(firstRow).toBeVisible();
    expect(requests).toHaveLength(rangeCount);
    await lower.fill("2026-04-18");
    await page
      .getByRole("combobox", { name: "Lower-bound comparison", exact: true })
      .selectOption("gt");
    await expect(apply).toBeDisabled();
    await expect(
      page.getByRole("combobox", {
        name: "Upper-bound comparison",
        exact: true,
      }),
    ).toHaveAttribute("aria-invalid", "true");
    await page
      .getByRole("combobox", { name: "Lower-bound comparison", exact: true })
      .selectOption("gte");
    // A genuine failed read still enters existing recovery; its replacement is editable in place.
    failNext = true;
    await apply.click();
    await expect(
      browsing.getByRole("button", { name: "Retry", exact: true }),
    ).toBeVisible();
    await expect(chip).toContainText("equals 2026-04-18");
    await trigger.click();
    await page
      .getByRole("button", { name: /Edit unapplied.*2026-04-18/ })
      .click();
    const failedCount = requests.length;
    await lower.fill("2026-04-31");
    await expect(apply).toBeDisabled();
    await lower.press("Enter");
    await expect(lower).toHaveValue("2026-04-31");
    await expect(chip).toContainText("equals 2026-04-18");
    await expect(firstRow).toBeVisible();
    expect(requests).toHaveLength(failedCount);
    await lower.fill("2026-04-17");
    const correctionResponse = page.waitForResponse(
      (response) =>
        response.url().endsWith(endpoint) && response.status() === 200,
    );
    await apply.click();
    await correctionResponse;
    await expect(
      browsing.getByRole("button", { name: "Retry", exact: true }),
    ).toHaveCount(0);
    await expect(chip).toContainText("2026-04-17");
    await expect(firstRow).toBeVisible();
    await chip.click();
    await lower.fill("bad date");
    await lower.press("Escape");
    await expect(chip).toBeFocused();
    await expect(page.getByRole("dialog")).toHaveCount(0);
  } finally {
    await heldDraft?.dispose();
  }
}

test("Timeline date filter drafts remain locally correctable before query admission", async ({
  page,
}) => {
  await dateFilterCorrection(page, timelineViewSchemaId);
});

test("Communications Log date filter drafts remain locally correctable before query admission", async ({
  page,
}) => {
  await dateFilterCorrection(page, commLogViewSchemaId);
});
