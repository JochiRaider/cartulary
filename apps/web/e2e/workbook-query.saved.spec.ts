import { applyFilterChip, changeGrouping } from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  savedViewModifiedTestId,
  workbookFilterOperatorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookQueryEntryTestId,
  workbookSortMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  notesViewSchemaId,
  requireViewContract,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Route } from "@playwright/test";
import { expect, test } from "./fixtures";

import { createIncident } from "./support/incidents/fixtures";

import { uniqueIncidentKey } from "./support/runtime/fixtureIdentity";

import {
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import {
  authorLiteralMembers,
  expectLiteralMembership,
  literalSetFixture,
} from "./support/workbook/literalSetFilters";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import {
  observeQuery,
  seed,
  timestampTaskFixture,
} from "./support/workbook/queryBrowsing";
import {
  createSavedView,
  createSavedViewFromCurrentSurface,
  selectSavedView,
  setSavedViewDraftName,
  updateSavedViewFromCurrentSurface,
} from "./support/workbook/savedViews";

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
    // Complete the adopted creation-first activation before another query gesture.
    await expect(
      page.getByTestId(draftCellTestId("timeline.date_entered_text")),
    ).toBeFocused();
    await expectLiteralMembership(page, view, [a], ids);
    await page.reload();
    await selectSavedView(page, view, saved.saved_view_id);
    // Complete the adopted creation-first activation before another query gesture.
    await expect(
      page.getByTestId(draftCellTestId("timeline.date_entered_text")),
    ).toBeFocused();
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
    await expect(
      page.getByTestId(draftCellTestId("timeline.date_entered_text")),
    ).toBeFocused();
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
  await installOwnedRoute(page, savedReads, olderRead);
  await installOwnedRoute(page, startupRead, olderRead);
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
  await removeOwnedRoute(page, savedReads, olderRead);
  await removeOwnedRoute(page, startupRead, olderRead);
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

test("Workbook canonical filters continue without feedback and saved views retain authored sort intent", async ({
  page,
  testActor,
}) => {
  test.setTimeout(180_000);
  const incident = await createIncident(
    page,
    uniqueIncidentKey("WQC-CANON"),
    "Canonical workbook continuation",
  );
  await seed(page, incident, notesViewSchemaId, 105, testActor.user_id);
  await seed(page, incident, timelineViewSchemaId, 105, testActor.user_id);
  const evidenceView = evidenceViewSchemaId;
  await seed(page, incident, evidenceView, 105, testActor.user_id);
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
