import type {
  QueryWorkbookViewRequest,
  QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts/http";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridRowTestId,
  workbookFilterOperatorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookFocusAnchorTestId,
  workbookInspectorCloseButtonTestId,
  workbookQueryEntryTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  requireViewContract,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";

import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";

import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { observePageRequests } from "./support/runtime/observations";
import {
  holdBrowserRequest,
  installOwnedRoute,
} from "./support/transport/requestInterception";
import {
  authorLiteralMembers,
  expectLiteralMembership,
  literalSetFixture,
} from "./support/workbook/literalSetFilters";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import { createViewRow } from "./support/workbook/query";
import {
  fixtureFields,
  observeQuery,
  timestampTaskFixture,
} from "./support/workbook/queryBrowsing";
import {
  openGenericInspectorForRecord,
  openTimelineInspector,
} from "./support/workbook/rowMutations";
import {
  createSavedView,
  selectSavedView,
} from "./support/workbook/savedViews";

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
  observePageRequests(page, (request) => {
    if (request.method() === "POST" && request.url().endsWith(endpoint))
      requests.push(request.postDataJSON() as QueryWorkbookViewRequest);
  });
  let failNext = false;
  await installOwnedRoute(page, `**${endpoint}`, async (route) => {
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
  {
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
    if (!(await trigger.isVisible()))
      await page.getByText("Query ▾", { exact: true }).click();
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
    await expect(
      page.getByRole("button", { name: "Query controls", exact: true }),
    ).toBeVisible();
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
    if (!(await trigger.isVisible()))
      await page.getByText("Query ▾", { exact: true }).click();
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

    await heldDraft?.dispose();
  }
}

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
  if (!(await trigger.isVisible()))
    await page.getByText("Query ▾", { exact: true }).click();
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
  await installOwnedRoute(
    page,
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
  await expect(
    page.getByRole("button", { name: "Query controls", exact: true }),
  ).toBeVisible();
  if (!(await trigger.isVisible()))
    await page.getByText("Query ▾", { exact: true }).click();
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
  observePageRequests(page, (request) => {
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
  if (!(await trigger.isVisible()))
    await page.getByText("Query ▾", { exact: true }).click();
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
  await installOwnedRoute(
    page,
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
  if (!(await trigger.isVisible()))
    await page.getByText("Query ▾", { exact: true }).click();
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
  await expect(
    page.getByRole("button", { name: "Query controls", exact: true }),
  ).toBeVisible();
  if (!(await trigger.isVisible()))
    await page.getByText("Query ▾", { exact: true }).click();
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
  await installOwnedRoute(page, `**${endpoint}`, async (route) => {
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
  if (!(await trigger.isVisible()))
    await page.getByText("Query ▾", { exact: true }).click();
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
  await expect(
    page.getByRole("button", { name: "Query controls", exact: true }),
  ).toBeVisible();
  if (!(await trigger.isVisible()))
    await page.getByText("Query ▾", { exact: true }).click();
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
