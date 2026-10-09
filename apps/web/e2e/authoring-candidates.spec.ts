import {
  assessmentCreateControlTestId,
  coordinationWorkflowTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  workbookAddRowButtonTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { readCurrentSession } from "./support/auth/sessions";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { openContextualCreationFixture } from "./support/workbook/contextualCreate";
import {
  authorLiteralMembers,
  literalSetFixture,
} from "./support/workbook/literalSetFilters";
import { openNoteFixture } from "./support/workbook/noteCreate";
import {
  ordinaryField,
  switchOrdinarySheet,
} from "./support/workbook/ordinaryCreate";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { activateCandidateIdentities } from "./support/workbook/references";
import {
  openGenericInspectorForRecord,
  openTimelineInspector,
} from "./support/workbook/rowMutations";
import { openTimelineEvidenceFixture } from "./support/workbook/timelineRelatedEvidence";

test("Literal candidate filters stage exact members and retain contextual selected identities", async ({
  page,
}) => {
  const f = await literalSetFixture(page);
  const first = f.rows[0],
    selected = f.rows[1];
  if (!first || !selected) throw new Error("Missing candidate fixture records");
  await openTimelineInspector(page, first.record_id);
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.task_request",
      ),
    )
    .click();
  const task = page.getByRole("region", {
    name: "Create task request",
    exact: true,
  });
  await button(task, "Choose Linked Records").click();
  const picker = task.getByRole("region", {
    name: "Choose Linked Records",
    exact: true,
  });
  const candidates = picker.getByRole("group", {
    name: "Linked Records",
    exact: true,
  });
  await activateCandidateIdentities(candidates, selected.record_id);
  const requests: unknown[] = [],
    memberships: string[][] = [];
  let failNext = false;
  await page.route(
    `**/incidents/${f.incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      requests.push(route.request().postDataJSON());
      if (failNext) {
        failNext = false;
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      memberships.push(
        (await response.json()).data.rows.map(
          (row: { record_id: string }) => row.record_id,
        ),
      );
      await route.fulfill({ response });
    },
  );
  await picker
    .getByText("Linked Records ordering and filters", { exact: true })
    .click();
  await picker
    .getByLabel("Linked Records filter field", { exact: true })
    .selectOption("timeline.tags");
  await authorLiteralMembers(
    picker,
    ["review,priority"],
    "Linked Records filter value",
  );
  await button(picker, "Add filter").click();
  expect(requests).toHaveLength(0);
  await button(picker, "Apply candidate query").click();
  await expect.poll(() => memberships.at(-1)).toEqual([first.record_id]);
  await expect(candidates.getByRole("checkbox")).toHaveCount(1);
  const retained = picker.getByRole("button", {
    name: /Remove selected Linked Records .*B component/,
  });
  await expect(retained).toBeVisible();
  failNext = true;
  await picker
    .getByLabel("Linked Records order")
    .selectOption("timeline.date_entered_sort_day:asc");
  await button(picker, "Apply candidate query").click();
  await expect.poll(() => requests.length).toBe(2);
  await expect(retained).toBeVisible();
  await button(picker, "Cancel references").click();
  await expect(button(task, "Choose Linked Records")).toBeFocused();
});

test("Literal Assessment support filters stage exact members and retain selected identities", async ({
  page,
}) => {
  const f = await literalSetFixture(page);
  const [first, selected, separate] = f.rows;
  if (!first || !selected || !separate)
    throw new Error("Missing support fixture records");
  await page.goto(
    `/?incident_id=${f.incident}&view_schema_id=${assessmentsViewSchemaId}`,
  );
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Literal support draft");
  await button(page, "Choose support").click();
  const support = page.getByRole("group", {
    name: "Choose assessment support",
    exact: true,
  });
  const candidates = support.getByRole("group", {
    name: "Timeline support candidates",
    exact: true,
  });
  await expect(candidates.getByRole("checkbox")).toHaveCount(3);
  await activateCandidateIdentities(candidates, selected.record_id);
  const requests: unknown[] = [],
    memberships: string[][] = [];
  await page.route(
    `**/incidents/${f.incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      requests.push(route.request().postDataJSON());
      const response = await route.fetch();
      memberships.push(
        (await response.json()).data.rows
          .map((row: { record_id: string }) => row.record_id)
          .sort(),
      );
      await route.fulfill({ response });
    },
  );
  await support
    .getByText("Timeline support candidates ordering and filters", {
      exact: true,
    })
    .click();
  await support
    .getByLabel("Timeline support candidates filter field", { exact: true })
    .selectOption("timeline.tags");
  for (const op of ["contains_any", "contains_all"] as const) {
    for (const values of [["review,priority"], ["review", "priority"]]) {
      const remove = support.getByRole("button", {
        name: "Remove filter Tags",
        exact: true,
      });
      if (await remove.count()) await remove.click();
      // Start a fresh draft so member removal remains deliberate.
      await support
        .getByLabel("Timeline support candidates filter field", { exact: true })
        .selectOption("timeline.activity_time_pair_state");
      await support
        .getByLabel("Timeline support candidates filter field", { exact: true })
        .selectOption("timeline.tags");
      await support
        .getByLabel("Timeline support candidates filter operator", {
          exact: true,
        })
        .selectOption(op);
      const count = requests.length;
      await authorLiteralMembers(
        support,
        values,
        "Timeline support candidates filter value",
      );
      await button(support, "Add filter").click();
      expect(requests).toHaveLength(count);
      await button(support, "Apply candidate query").click();
      const expected =
        values.length === 1
          ? [first.record_id]
          : op === "contains_any"
            ? [selected.record_id, separate.record_id]
            : [separate.record_id];
      await expect.poll(() => memberships.at(-1)).toEqual(expected.sort());
      await expect(candidates.getByRole("checkbox")).toHaveCount(
        expected.length,
      );
      const optionIds = await candidates
        .getByRole("checkbox")
        .evaluateAll((options) =>
          options.map((option) => (option as HTMLInputElement).value).sort(),
        );
      expect(optionIds).toEqual(expected.sort());
      await expect(
        support.getByRole("button", {
          name: /Remove selected Timeline support candidates .*B component/,
        }),
      ).toBeVisible();
    }
  }
  await button(support, "Cancel support selection").click();
  await expect(button(page, "Choose support")).toBeFocused();
  await expect(
    page.getByRole("region", {
      name: "Assessment supporting records",
      exact: true,
    }),
  ).toContainText("Supporting records (0/64)");
});

async function seed(page: Page, incident: string, view: string, field: string) {
  const ids: string[] = [];
  for (let offset = 0; offset < 105; offset += 5) {
    const rows = await Promise.all(
      Array.from({ length: 5 }, (_, index) =>
        createViewRow(page, incident, view, {
          client_txn_id: uniqueTxn("candidate"),
          [field]: `ACD ${String(offset + index).padStart(3, "0")} ${"Long reference label ".repeat(5)}`,
          ...(view === partiesViewSchemaId
            ? { "party.party_kind": "person" }
            : {}),
        }),
      ),
    );
    ids.push(...rows.map((row) => row.record_id));
  }
  return ids;
}
async function chooseFirst(select: Locator, multiple = false) {
  const option = multiple
    ? select.getByRole("checkbox").first()
    : select.getByRole("radio").first();
  await expect(option).toBeAttached();
  const id = await option.getAttribute("value");
  if (!id) throw new Error("Missing exact candidate identity");
  await activateCandidateIdentities(select, id);
  return id;
}
function button(scope: Locator | Page, name: string) {
  return scope.getByRole("button", { name, exact: true });
}

test("Party timestamp candidate filters stage explicitly and preserve selected identities", async ({
  page,
}) => {
  const f = await openTimelineEvidenceFixture(page);
  const party = await createViewRow(page, f.incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("timestamp-party"),
    "party.display_name": "Timestamp Party",
    "party.party_kind": "person",
  });
  await button(f.form, "Keep draft and close").click();
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.task_request",
      ),
    )
    .click();
  const task = page.getByRole("region", {
    name: "Create task request",
    exact: true,
  });
  await button(task, "Choose Requester Party").click();
  const picker = task.getByRole("region", {
    name: "Choose Requester Party",
    exact: true,
  });
  const select = picker.getByRole("group", {
    name: "Requester Party",
    exact: true,
  });
  await expect(
    select.getByRole("radio", {
      name: `Timestamp Party (${party.record_id})`,
      exact: true,
    }),
  ).toBeAttached();
  await activateCandidateIdentities(select, party.record_id);
  const queries: { filters?: unknown }[] = [];
  await page.route(
    `**/incidents/${f.incident}/views/${partiesViewSchemaId}/query`,
    async (route) => {
      queries.push(route.request().postDataJSON());
      await route.continue();
    },
  );
  await picker
    .getByText("Requester Party ordering and filters", { exact: true })
    .click();
  await picker
    .getByLabel("Requester Party filter field", { exact: true })
    .selectOption("party.updated_at");
  const value = picker.getByLabel("Requester Party filter value", {
    exact: true,
  });
  await value.fill("tomorrow");
  await expect(button(picker, "Add filter")).toBeDisabled();
  await expect(value).toHaveAttribute("aria-invalid", "true");
  expect(queries).toHaveLength(0);
  await value.fill("0001-01-01T00:00:00.000000001Z");
  await button(picker, "Add filter").click();
  expect(queries).toHaveLength(0);
  await button(picker, "Apply candidate query").click();
  await expect.poll(() => queries.length).toBe(1);
  expect(queries[0]?.filters).toEqual([
    {
      field_key: "party.updated_at",
      op: "eq",
      arg: { value: "0001-01-01T00:00:00.000000001Z" },
    },
  ]);
  await expect(
    select.getByRole("radio", {
      name: `Timestamp Party (${party.record_id})`,
      exact: true,
    }),
  ).toHaveCount(0);
  await expect(
    button(picker, "Remove selected Requester Party Timestamp Party"),
  ).toBeVisible();
  await button(picker, "Apply references").click();
  await expect(
    task.getByRole("group", {
      name: "Requester Party selected references",
      exact: true,
    }),
  ).toContainText("Timestamp Party");
});

test("Boolean support filters stage typed queries and retain selected identities across pages", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("BOOL-SUPPORT"),
    "Boolean support filtering",
  );
  const ids = await seed(
    page,
    incident,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  let submissions = 0;
  await page.route(
    `**/incidents/${incident}/views/${assessmentsViewSchemaId}/rows`,
    async (route) => {
      submissions++;
      await route.continue();
    },
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${assessmentsViewSchemaId}`,
  );
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Independent boolean support draft");
  await button(page, "Choose support").click();
  const support = page.getByRole("group", {
    name: "Choose assessment support",
    exact: true,
  });
  const candidates = support.getByRole("group", {
    name: "Timeline support candidates",
    exact: true,
  });
  await expect(candidates.getByRole("checkbox")).toHaveCount(100);
  const first = await chooseFirst(candidates, true);
  expect(ids).toContain(first);
  await button(support, "Next candidates").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(5);
  const second = await chooseFirst(candidates, true);
  expect(ids).toContain(second);
  expect(first).not.toBe(second);
  const selected = support.getByRole("button", {
    name: /^Remove selected Timeline support candidates /,
  });
  const identities = await selected.allTextContents();
  const queries: { filters?: unknown }[] = [];
  const accepted: unknown[] = [];
  let failNext = false;
  await page.route(
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      queries.push(route.request().postDataJSON());
      if (failNext) {
        failNext = false;
        await route.abort("failed");
        return;
      }
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      accepted.push((await response.json()).meta.query.filters);
      await route.fulfill({ response });
    },
  );
  await page.setViewportSize({ width: 768, height: 640 });
  await support
    .getByText("Timeline support candidates ordering and filters", {
      exact: true,
    })
    .click();
  await support
    .getByRole("combobox", {
      name: "Timeline support candidates filter field",
      exact: true,
    })
    .selectOption("timeline.has_evidence");
  const mode = support.getByRole("combobox", {
    name: "Timeline support candidates equality operand",
    exact: true,
  });
  await mode.selectOption("values");
  await expect(button(support, "Add filter")).toBeDisabled();
  const trueChoice = support.getByRole("checkbox", {
    name: "true",
    exact: true,
  });
  const falseChoice = support.getByRole("checkbox", {
    name: "false",
    exact: true,
  });
  await trueChoice.focus();
  await page.keyboard.press("Tab");
  await expect(falseChoice).toBeFocused();
  await page.keyboard.press("Space");
  await page.keyboard.press("Tab");
  await expect(button(support, "Add filter")).toBeFocused();
  await page.keyboard.press("Enter");
  expect(queries).toHaveLength(0);
  await expect(candidates.getByRole("checkbox")).toHaveCount(5);
  await button(support, "Apply candidate query").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(100);
  const filters = [
    { field_key: "timeline.has_evidence", op: "eq", arg: { values: [false] } },
  ];
  expect(queries).toHaveLength(1);
  expect(queries[0]?.filters).toEqual(filters);
  expect(accepted[0]).toEqual(filters);
  expect(await selected.allTextContents()).toEqual(identities);
  await falseChoice.uncheck();
  await trueChoice.check();
  await button(support, "Add filter").click();
  expect(queries).toHaveLength(1);
  failNext = true;
  await button(support, "Apply candidate query").click();
  await expect(button(support, "Retry candidates")).toBeVisible();
  expect(await selected.allTextContents()).toEqual(identities);
  await button(support, "Retry candidates").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(0);
  expect(queries.at(-1)?.filters).toEqual([
    { field_key: "timeline.has_evidence", op: "eq", arg: { values: [true] } },
  ]);
  expect(await selected.allTextContents()).toEqual(identities);
  await button(support, "Apply support selection").click();
  await expect(button(page, "Choose support")).toBeFocused();
  await expect(
    page.getByRole("region", {
      name: "Assessment supporting records",
      exact: true,
    }),
  ).toContainText("Supporting records (2/64)");
  await button(page, "Choose support").click();
  await activateCandidateIdentities(candidates, []);
  await candidates.press("Escape");
  await expect(button(page, "Choose support")).toBeFocused();
  await expect(
    page.getByRole("region", {
      name: "Assessment supporting records",
      exact: true,
    }),
  ).toContainText("Supporting records (2/64)");
  expect(submissions).toBe(0);
});

test("Ordinary reference boolean filtering preserves the selected source until explicit acceptance", async ({
  page,
}) => {
  const f = await openNoteFixture(page);
  await f.form
    .getByRole("textbox", { name: "Title", exact: true })
    .fill("Boolean reference draft");
  await button(f.form, "Choose source").click();
  const picker = f.form.getByRole("region", {
    name: "Choose Note source",
    exact: true,
  });
  const select = picker.getByRole("group", {
    name: "Note source",
    exact: true,
  });
  await expect(
    select.locator(`input[value="${f.source.record_id}"]`),
  ).toBeChecked();
  const selected = picker.getByRole("button", {
    name: /^Remove selected Note source /,
  });
  const selection = await selected.allTextContents();
  const queries: { filters?: unknown }[] = [];
  const canonical: unknown[] = [];
  let writes = 0;
  await page.route(
    `**/records/${f.source.record_id}/linked-notes`,
    async (route) => {
      writes++;
      await route.continue();
    },
  );
  await page.route(
    `**/incidents/${f.incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      queries.push(route.request().postDataJSON());
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      canonical.push((await response.json()).meta.query.filters);
      await route.fulfill({ response });
    },
  );
  await picker
    .getByText("Note source ordering and filters", { exact: true })
    .click();
  await picker
    .getByRole("combobox", { name: "Note source filter field", exact: true })
    .selectOption("timeline.has_evidence");
  const value = picker.getByRole("combobox", {
    name: "Note source filter value",
    exact: true,
  });
  await expect(value).toHaveValue("");
  await expect(button(picker, "Add filter")).toBeDisabled();
  await value.selectOption("true");
  await button(picker, "Add filter").click();
  expect(queries).toHaveLength(0);
  await button(picker, "Apply candidate query").click();
  await expect(picker).toContainText("No candidates match this query.");
  expect(queries[0]?.filters).toEqual([
    { field_key: "timeline.has_evidence", op: "eq", arg: { value: true } },
  ]);
  expect(canonical[0]).toEqual(queries[0]?.filters);
  expect(await selected.allTextContents()).toEqual(selection);
  await button(picker, "Cancel source").click();
  await expect(button(f.form, "Choose source")).toBeFocused();
  await button(f.form, "Choose source").click();
  await expect(
    select.locator(`input[value="${f.source.record_id}"]`),
  ).toBeChecked();
  await button(picker, "Apply source").click();
  await expect(button(f.form, "Choose source")).toBeFocused();
  await expect(
    f.form.getByRole("textbox", { name: "Title", exact: true }),
  ).toHaveValue("Boolean reference draft");
  expect(writes).toBe(0);
});

test("Assessment Timeline support enum filtering preserves staged selection", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ENUM-SUPPORT"),
    "Explicit support filtering",
  );
  await createViewRow(page, incident, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("enum-host"),
    "host.display_name": "Assessment subject",
  });
  const ids = await seed(
    page,
    incident,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  let submissions = 0;
  const queries: { filters?: unknown }[] = [];
  let failNext = false;
  await page.route(
    `**/incidents/${incident}/views/${assessmentsViewSchemaId}/rows`,
    async (route) => {
      submissions++;
      await route.continue();
    },
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${assessmentsViewSchemaId}`,
  );
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  await page
    .getByTestId(assessmentCreateControlTestId("rationale"))
    .fill("Independent assessment draft");
  await button(page, "Choose support").click();
  const support = page.getByRole("group", {
    name: "Choose assessment support",
    exact: true,
  });
  const candidates = support.getByRole("group", {
    name: "Timeline support candidates",
    exact: true,
  });
  await expect(candidates.getByRole("checkbox")).toHaveCount(100);
  const first = await chooseFirst(candidates, true);
  expect(ids).toContain(first);
  await button(support, "Next candidates").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(5);
  const second = await chooseFirst(candidates, true);
  expect(second).not.toBe(first);
  await page.setViewportSize({ width: 768, height: 640 });
  await page.route(
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      queries.push(route.request().postDataJSON());
      if (failNext) {
        failNext = false;
        await route.abort("failed");
        return;
      }
      await route.fulfill({ response: await route.fetch() });
    },
  );
  await support
    .getByText("Timeline support candidates ordering and filters", {
      exact: true,
    })
    .click();
  await support
    .getByRole("combobox", {
      name: "Timeline support candidates filter field",
      exact: true,
    })
    .selectOption("timeline.activity_time_pair_state");
  await support
    .getByRole("combobox", {
      name: "Timeline support candidates filter value",
      exact: true,
    })
    .selectOption("disabled");
  expect(queries).toHaveLength(0);
  await button(support, "Add filter").click();
  expect(queries).toHaveLength(0);
  await expect(candidates.getByRole("checkbox")).toHaveCount(5);
  await button(support, "Apply candidate query").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(100);
  expect(queries).toHaveLength(1);
  expect(queries[0]?.filters).toEqual([
    {
      field_key: "timeline.activity_time_pair_state",
      op: "eq",
      arg: { value: "disabled" },
    },
  ]);
  await expect(
    support.getByRole("button", {
      name: /^Remove selected Timeline support candidates /,
    }),
  ).toHaveCount(2);
  await support
    .getByRole("combobox", {
      name: "Timeline support candidates equality operand",
      exact: true,
    })
    .selectOption("values");
  await support.getByRole("checkbox", { name: "empty", exact: true }).check();
  await button(support, "Custom literals").click();
  await button(support, "Add literal").click();
  await support
    .getByRole("textbox", {
      name: "Timeline support candidates filter value literal 2",
      exact: true,
    })
    .fill("long_custom_literal,".repeat(12));
  const spacing = await page.addStyleTag({
    content: `
    * { line-height: 1.5 !important; letter-spacing: 0.12em !important; word-spacing: 0.16em !important; }
    p { margin-block-end: 2em !important; }
  `,
  });
  await button(support, "Apply candidate query").focus();
  const applyBox = await button(support, "Apply candidate query").boundingBox();
  expect(applyBox).not.toBeNull();
  if (!applyBox) throw new Error("Missing candidate Apply geometry");
  expect(applyBox.y).toBeGreaterThanOrEqual(0);
  expect(applyBox.y + applyBox.height).toBeLessThanOrEqual(640);
  await test.info().attach("enum-candidate-text-spacing", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await spacing.evaluate((element) => element.parentNode?.removeChild(element));
  await button(support, "Add filter").click();
  failNext = true;
  await button(support, "Apply candidate query").click();
  await expect(button(support, "Retry candidates")).toBeVisible();
  await expect(
    support.getByRole("button", {
      name: /^Remove selected Timeline support candidates /,
    }),
  ).toHaveCount(2);
  await button(support, "Retry candidates").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(0);
  await expect(
    support.getByRole("button", {
      name: /^Remove selected Timeline support candidates /,
    }),
  ).toHaveCount(2);
  expect(submissions).toBe(0);
  const selected = page.getByRole("region", {
    name: "Assessment supporting records",
    exact: true,
  });
  await expect(selected).toContainText("Supporting records (0/64)");
  await button(support, "Cancel support selection").click();
  await expect(button(page, "Choose support")).toBeFocused();
  await expect(selected).toContainText("Supporting records (0/64)");
  await button(page, "Choose support").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(100);
  await activateCandidateIdentities(candidates, first);
  await button(support, "Apply support selection").click();
  await expect(selected).toContainText("Supporting records (1/64)");
  await button(page, "Choose support").click();
  await activateCandidateIdentities(candidates, []);
  await candidates.press("Escape");
  await expect(selected).toContainText("Supporting records (1/64)");
  await expect(
    page.getByTestId(assessmentCreateControlTestId("rationale")),
  ).toHaveValue("Independent assessment draft");
  expect(submissions).toBe(0);
});

test("Timeline retained Owner reconciles accepted membership labels only on Apply and remains readable through Recovery", async ({
  page,
}) => {
  const actor = await readCurrentSession(page);
  let writes = 0;
  await page.route(
    `**/views/${taskRequestsViewSchemaId}/rows`,
    async (route) => {
      writes++;
      await route.continue();
    },
  );
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 720 },
  ]) {
    await page.setViewportSize(viewport);
    const f = await openTimelineEvidenceFixture(page);
    await button(f.form, "Keep draft and close").click();
    await page
      .getByTestId(
        workbookInspectorFeatureActionTestId(
          timelineViewSchemaId,
          "create_related.task_request",
        ),
      )
      .click();
    const task = page.getByRole("region", {
      name: "Create task request",
      exact: true,
    });
    const title = task.getByTestId(genericCreateFieldTestId("task.title"));
    const authored = `Retained Owner title at ${viewport.width}`;
    await title.fill(authored);
    const owner = task.getByRole("group", {
      name: "Owner selected references",
      exact: true,
    });
    await expect(owner.getByRole("list")).toContainText("Current actor");
    const choose = button(owner, "Choose Owner");
    await choose.focus();
    await choose.press("Enter");
    const picker = task.getByRole("region", {
      name: "Choose Owner",
      exact: true,
    });
    const select = picker.getByRole("group", { name: "Owner", exact: true });
    await expect(select).toBeEnabled();
    await expect(
      select.locator(`input[value="${actor.user_id}"]`),
    ).toBeChecked();
    await expect(
      select.locator(`input[value="${actor.user_id}"]`),
    ).toHaveAccessibleName(`${actor.display_name} (${actor.user_id})`);
    await expect(picker.getByRole("list")).toContainText(actor.display_name);
    const remove = picker.getByRole("button", {
      name: /^Remove selected Owner /,
    });
    await expect(remove).toHaveAccessibleName(
      `Remove selected Owner ${actor.display_name}`,
    );
    await expect(owner.getByRole("list").first()).toContainText(
      "Current actor",
    );
    await select.getByRole("radio", { checked: true }).focus();
    await page.keyboard.press("Escape");
    await expect(choose).toBeFocused();
    await expect(owner.getByRole("list")).toContainText("Current actor");
    await choose.press("Enter");
    await expect(select).toBeEnabled();
    await button(picker, "Apply references").click();
    await expect(owner.getByRole("list")).toContainText(actor.display_name);
    await expect(
      owner.getByRole("button", { name: /^Remove Owner / }),
    ).toHaveAccessibleName(`Remove Owner ${actor.display_name}`);
    await expect(title).toHaveValue(authored);
    await button(task, "Keep draft and close").click();
    await openRecoveryItem(page, /^Task Requests draft ·/);
    const recovery = page.getByRole("region", {
      name: "Retained contextual creation",
      exact: true,
    });
    await button(recovery, "Resume contextual draft").click();
    await expect(
      recovery.getByTestId(genericCreateFieldTestId("task.title")),
    ).toHaveValue(authored);
    const resumedOwner = recovery.getByRole("group", {
      name: "Owner selected references",
      exact: true,
    });
    await expect(resumedOwner.getByRole("list")).toContainText(
      actor.display_name,
    );
    await button(resumedOwner, "Choose Owner").click();
    await expect(
      recovery
        .getByRole("group", { name: "Owner", exact: true })
        .locator(`input[value="${actor.user_id}"]`),
    ).toBeChecked();
    await button(recovery, "Cancel references").click();
    await button(recovery, "Discard draft").click();
  }
  expect(writes).toBe(0);
});

test("Timeline retained Requester Party updates accepted presentation through refresh and page replacement", async ({
  page,
}) => {
  const f = await openTimelineEvidenceFixture(page);
  await seed(page, f.incident, partiesViewSchemaId, "party.display_name");
  await button(f.form, "Keep draft and close").click();
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.task_request",
      ),
    )
    .click();
  const task = page.getByRole("region", {
    name: "Create task request",
    exact: true,
  });
  const choose = button(task, "Choose Requester Party");
  await choose.click();
  const picker = task.getByRole("region", {
    name: "Choose Requester Party",
    exact: true,
  });
  const select = picker.getByRole("group", {
    name: "Requester Party",
    exact: true,
  });
  await expect(select.getByRole("radio")).toHaveCount(100);
  const id = await chooseFirst(select);
  const original =
    (
      await select.locator(`input[value="${id}"]`).getAttribute("aria-label")
    )?.replace(` (${id})`, "") ?? "";
  await button(picker, "Apply references").click();
  const retained = task.getByRole("group", {
    name: "Requester Party selected references",
    exact: true,
  });
  await expect(retained.getByRole("list")).toContainText(original);
  const rows = await queryViewRows(page, f.incident, partiesViewSchemaId);
  const row = rows.find((row) => row.record_id === id);
  if (!row) throw new Error("Selected fixture Party missing from first page");
  const renamed = "ACD 000 Accepted renamed Party";
  await patchRecord(page, id, {
    base_row_version: row.row_version,
    client_txn_id: uniqueTxn("rename-party"),
    view_schema_id: partiesViewSchemaId,
    changes: [{ field_key: "party.display_name", value: renamed }],
  });
  await choose.click();
  await expect(select.locator(`input[value="${id}"]`)).toHaveAccessibleName(
    `${renamed} (${id})`,
  );
  const remove = picker.getByRole("button", {
    name: /^Remove selected Requester Party /,
  });
  await expect(remove).toHaveAccessibleName(
    `Remove selected Requester Party ${renamed}`,
  );
  const refresh = button(picker, "Refresh candidates");
  await refresh.focus();
  await refresh.press("Enter");
  await expect(refresh).toHaveAttribute("aria-busy", "false");
  await expect(refresh).toBeFocused();
  await expect(remove).toHaveAccessibleName(
    `Remove selected Requester Party ${renamed}`,
  );
  await button(picker, "Next candidates").click();
  await expect(select.locator(`input[value="${id}"]`)).toHaveCount(0);
  await expect(remove).toHaveAccessibleName(
    `Remove selected Requester Party ${renamed}`,
  );
  await button(picker, "Previous candidates").click();
  await expect(select.locator(`input[value="${id}"]`)).toBeChecked();
  await button(picker, "Apply references").click();
  await expect(retained.getByRole("list")).toContainText(renamed);
  await expect(
    retained.getByRole("button", { name: /^Remove Requester Party / }),
  ).toHaveAccessibleName(`Remove Requester Party ${renamed}`);
});

test("Authoring Party pages retain ordinary contextual and related Evidence selections through failed continuation", async ({
  page,
}, info) => {
  const f = await openTimelineEvidenceFixture(page);
  const parties = await seed(
    page,
    f.incident,
    partiesViewSchemaId,
    "party.display_name",
  );
  const reads: { cursor_token?: string }[] = [];
  let failed = false;
  await page.route(
    `**/incidents/${f.incident}/views/${partiesViewSchemaId}/query`,
    async (route) => {
      const request = route.request().postDataJSON() as {
        cursor_token?: string;
      };
      reads.push(request);
      if (request.cursor_token && !failed) {
        failed = true;
        await route.abort("failed");
      } else await route.continue();
    },
  );
  await button(f.form, "Choose Collector Party").click();
  const picker = f.form.getByRole("region", {
    name: "Choose Collector Party",
    exact: true,
  });
  const select = picker.getByRole("group", {
    name: "Collector Party",
    exact: true,
  });
  await expect(select.getByRole("radio")).toHaveCount(100);
  const selected = await chooseFirst(select);
  expect(parties).toContain(selected);
  await button(picker, "Next candidates").click();
  await expect(picker.getByRole("alert")).toContainText(
    "accepted page remains available",
  );
  await expect(select).toBeEnabled();
  await expect(button(picker, "Apply Party")).toBeEnabled();
  const failedRead = reads.at(-1);
  await button(picker, "Retry candidates").click();
  await expect(select.getByRole("radio")).toHaveCount(6);
  expect(reads.at(-1)).toEqual(failedRead);
  await expect(
    picker.getByRole("button", { name: /^Remove selected Collector Party / }),
  ).toHaveCount(1);
  await button(picker, "Apply Party").click();
  await expect(picker).toHaveCount(0);
  await expect(
    f.form.getByTestId(
      genericCreateFieldTestId("evidence.collector_party_text"),
    ),
  ).toHaveValue("Response team collection log");
  const parentCollectorRemove = button(f.form, "Remove Collector Party");
  await parentCollectorRemove.focus();
  await page.keyboard.press("Enter");
  await expect(button(f.form, "Choose Collector Party")).toBeFocused();
  await expect(
    f.form.getByTestId(
      genericCreateFieldTestId("evidence.collector_party_text"),
    ),
  ).toHaveValue("Response team collection log");
  await button(f.form, "Choose Source Party").click();
  const source = f.form.getByRole("region", {
    name: "Choose Source Party",
    exact: true,
  });
  await expect(
    source.getByRole("group", { name: "Source Party", exact: true }),
  ).toBeEnabled();
  await button(source, "Cancel Party selection").click();
  await expect(button(f.form, "Choose Source Party")).toBeFocused();
  await button(f.form, "Keep draft and close").click();

  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.task_request",
      ),
    )
    .click();
  const task = page.getByRole("region", {
    name: "Create task request",
    exact: true,
  });
  await task
    .getByTestId(genericCreateFieldTestId("task.title"))
    .fill("Keep contextual intent");
  await button(task, "Choose Linked Records").click();
  const references = task.getByRole("region", {
    name: "Choose Linked Records",
    exact: true,
  });
  await references
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(partiesViewSchemaId);
  const records = references.getByRole("group", {
    name: "Linked Records",
    exact: true,
  });
  await expect(records.getByRole("checkbox")).toHaveCount(100);
  await chooseFirst(records, true);
  await button(references, "Next candidates").click();
  await expect(records.getByRole("checkbox")).toHaveCount(6);
  await chooseFirst(records, true);
  await expect(
    references.getByRole("button", {
      name: /^Remove selected Linked Records /,
    }),
  ).toHaveCount(3);
  await button(references, "Apply references").click();
  await expect(
    task.getByTestId(genericCreateFieldTestId("task.title")),
  ).toHaveValue("Keep contextual intent");

  await switchOrdinarySheet(page, evidenceViewSchemaId);
  const raw = await ordinaryField(
    page,
    evidenceViewSchemaId,
    "evidence.collector_party_id",
  );
  await raw.fill(selected);
  const cell = page.getByRole("gridcell").filter({ has: raw });
  const choose = cell.getByRole("button", {
    name: "Choose collector party",
    exact: true,
  });
  // Hidden fields may be authored in the supplementary form, with the same semantic trigger.
  const trigger = (await choose.count())
    ? choose
    : page.getByRole("button", { name: "Choose collector party", exact: true });
  await trigger.click();
  const ordinary = page.getByRole("region", {
    name: "Choose collector party",
    exact: true,
  });
  const ordinaryCandidates = ordinary.getByRole("group", {
    name: "Collector Party",
    exact: true,
  });
  await expect(ordinaryCandidates.getByRole("radio")).toHaveCount(100);
  const compactRemove = ordinary.getByRole("button", {
    name: /^Remove selected Collector Party /,
  });
  await expect(compactRemove).toHaveCount(1);
  await ordinaryCandidates.getByRole("radio").first().focus();
  await page.keyboard.press("Tab");
  await expect(compactRemove).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(ordinaryCandidates.getByRole("radio").first()).toBeFocused();
  await button(ordinary, "Cancel references").click();
  await expect(raw).toHaveValue(selected);
  await trigger.click();
  await expect(ordinaryCandidates.getByRole("radio")).toHaveCount(100);
  await button(ordinary, "Next candidates").click();
  await expect(ordinaryCandidates.getByRole("radio")).toHaveCount(6);
  await button(ordinary, "Apply references").click();
  await expect(raw).toHaveValue(selected);
  await expect(trigger).toBeFocused();
  expect(
    await queryViewRows(page, f.incident, evidenceViewSchemaId),
  ).toHaveLength(0);
  await info.attach("ordinary-and-Party-candidate-review", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
});

test("Timeline contextual Party candidate reads keep keyboard focus through pending retry and exhaustion", async ({
  page,
}) => {
  const f = await openTimelineEvidenceFixture(page);
  await seed(page, f.incident, partiesViewSchemaId, "party.display_name");
  await button(f.form, "Keep draft and close").click();
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.task_request",
      ),
    )
    .click();
  const task = page.getByRole("region", {
    name: "Create task request",
    exact: true,
  });
  const title = task.getByTestId(genericCreateFieldTestId("task.title"));
  await title.fill("Keyboard retained task draft");
  await button(task, "Choose Requester Party").click();
  const picker = task.getByRole("region", {
    name: "Choose Requester Party",
    exact: true,
  });
  const candidates = picker.getByRole("group", {
    name: "Requester Party",
    exact: true,
  });
  await expect(candidates.getByRole("radio")).toHaveCount(100);

  type Gate = {
    readonly wait: Promise<void>;
    readonly requested: () => void;
    readonly fail: boolean;
    readonly release: () => void;
  };
  let queued: Gate | null = null;
  let reads = 0;
  await page.route(
    `**/incidents/${f.incident}/views/${partiesViewSchemaId}/query`,
    async (route) => {
      reads++;
      const gate = queued;
      queued = null;
      if (gate) {
        gate.requested();
        await gate.wait;
        if (gate.fail) await route.abort("failed");
        else await route.continue();
      } else await route.continue();
    },
  );
  function hold(fail = false) {
    let release!: () => void;
    let requested!: () => void;
    const wait = new Promise<void>((resolve) => {
      release = resolve;
    });
    const entered = new Promise<void>((resolve) => {
      requested = resolve;
    });
    queued = { wait, requested, fail, release };
    return { entered, release };
  }
  async function admit(name: string, fail = false) {
    const gate = hold(fail);
    const control = button(picker, name);
    await control.focus();
    await page.keyboard.press("Enter");
    await gate.entered;
    await expect(control).toBeFocused();
    await expect(control).toHaveAttribute("aria-busy", "true");
    await expect(control).toHaveAttribute("aria-disabled", "true");
    await expect(control).not.toHaveAttribute("disabled");
    return { control, release: gate.release };
  }

  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 720 },
  ]) {
    await page.setViewportSize(viewport);
    const read = await admit("Refresh candidates");
    const count = reads;
    await page.keyboard.press("Enter");
    await page.keyboard.press("Space");
    const bounds = await read.control.boundingBox();
    if (!bounds) throw new Error("Missing focused Refresh control geometry");
    await page.mouse.click(
      bounds.x + bounds.width / 2,
      bounds.y + bounds.height / 2,
    );
    expect(reads).toBe(count);
    await read.control.focus();
    await expect(read.control).toBeFocused();
    read.release();
    await expect(
      picker.getByRole("status").filter({ hasText: /Page|Loading candidates/ }),
    ).toContainText("Page 1");
    await expect(read.control).toBeFocused();
    await expect(read.control).toHaveAttribute("aria-busy", "false");
  }

  const selected = await chooseFirst(candidates);
  const next = await admit("Next candidates", true);
  next.release();
  await expect(picker.getByRole("alert")).toContainText(
    "accepted page remains available",
  );
  await expect(next.control).toBeFocused();
  await expect(candidates).toBeEnabled();
  await expect(button(picker, "Apply references")).toBeEnabled();

  const failedRetry = await admit("Retry candidates", true);
  failedRetry.release();
  await expect(picker.getByRole("alert")).toBeVisible();
  await expect(failedRetry.control).toBeFocused();
  await expect(failedRetry.control).toHaveAttribute("aria-disabled", "false");

  const successfulRetry = await admit("Retry candidates");
  successfulRetry.release();
  await expect(candidates.getByRole("radio")).toHaveCount(6);
  await expect(successfulRetry.control).toBeFocused();
  await expect(successfulRetry.control).toHaveAttribute(
    "aria-disabled",
    "true",
  );
  await page.keyboard.press("Tab");
  await expect(button(picker, "Apply references")).toBeFocused();
  await expect(button(picker, "Retry candidates")).toBeDisabled();
  await expect(button(picker, "Next candidates")).toBeDisabled();
  await expect(
    picker.getByRole("button", { name: /^Remove selected Requester Party / }),
  ).toHaveCount(1);

  const previous = await admit("Previous candidates");
  previous.release();
  await expect(candidates.getByRole("radio")).toHaveCount(100);
  await expect(previous.control).toBeFocused();
  await expect(previous.control).toHaveAttribute("aria-disabled", "true");
  await page.keyboard.press("Tab");
  await expect(previous.control).toBeDisabled();

  const first = await admit("First candidates");
  first.release();
  await expect(
    picker.getByRole("status").filter({ hasText: /Page|Loading candidates/ }),
  ).toContainText("Page 1");
  await expect(first.control).toBeFocused();
  await expect(title).toHaveValue("Keyboard retained task draft");
  await button(picker, "Cancel references").click();
  await expect(title).toHaveValue("Keyboard retained task draft");
  expect(selected).toBeTruthy();
  expect(
    await queryViewRows(page, f.incident, taskRequestsViewSchemaId),
  ).toHaveLength(0);
});

test("Timeline contextual reference removal keeps keyboard focus in the reference field", async ({
  page,
}, info) => {
  const f = await openTimelineEvidenceFixture(page);
  await seed(page, f.incident, partiesViewSchemaId, "party.display_name");
  await button(f.form, "Keep draft and close").click();
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.task_request",
      ),
    )
    .click();
  const task = page.getByRole("region", {
    name: "Create task request",
    exact: true,
  });
  const title = task.getByTestId(genericCreateFieldTestId("task.title"));
  await title.fill("Retained keyboard authoring");
  await page.setViewportSize({ width: 430, height: 700 });
  const choose = button(task, "Choose Linked Records");
  const parentRemove = task.getByRole("button", {
    name: /^Remove Linked Records /,
  });
  await expect(parentRemove).toHaveCount(1);
  await choose.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Shift+Tab");
  await expect(parentRemove).toBeFocused();
  await page.keyboard.press("Enter");
  const parentFocus = await page.evaluate(() =>
    document.activeElement?.tagName === "BUTTON"
      ? document.activeElement.textContent?.trim()
      : document.activeElement?.tagName,
  );

  await choose.click();
  const picker = task.getByRole("region", {
    name: "Choose Linked Records",
    exact: true,
  });
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(partiesViewSchemaId);
  const select = picker.getByRole("group", {
    name: "Linked Records",
    exact: true,
  });
  await expect(select.getByRole("checkbox")).toHaveCount(100);
  const firstPageIds = await select
    .getByRole("checkbox")
    .evaluateAll((options) =>
      options.slice(0, 3).map((option) => (option as HTMLInputElement).value),
    );
  await activateCandidateIdentities(select, firstPageIds);
  const remove = picker.getByRole("button", {
    name: /^Remove selected Linked Records /,
  });
  await expect(remove).toHaveCount(3);
  await button(picker, "Next candidates").click();
  await expect(select.getByRole("checkbox")).toHaveCount(6);
  await expect(remove).toHaveCount(3);
  const lastId = await select
    .getByRole("checkbox")
    .first()
    .getAttribute("value");
  if (!lastId) throw new Error("Missing off-page candidate identity");
  await activateCandidateIdentities(select, lastId);
  await expect(remove).toHaveCount(4);
  const scrollOffsets = async () => ({
    ...(await page.evaluate(() => ({
      page: document.scrollingElement?.scrollTop ?? 0,
      grid: document.querySelector('[role="grid"]')?.scrollTop ?? 0,
    }))),
    inspector: await task.evaluate((element) => {
      for (
        let parent = element.parentElement;
        parent;
        parent = parent.parentElement
      ) {
        const style = getComputedStyle(parent);
        if (
          /(auto|scroll)/.test(style.overflowY) &&
          parent.scrollHeight > parent.clientHeight
        )
          return parent.scrollTop;
      }
      return null;
    }),
  });
  const fallbackName = await select
    .getByRole("checkbox")
    .first()
    .getAttribute("aria-label");
  const observations: {
    position: string;
    focus: string | null;
    scroll: unknown;
  }[] = [];
  for (const [position, tabCount, expected] of [
    ["first", 1, 3],
    ["middle", 2, 2],
    ["last", 2, 1],
    ["sole", 1, 0],
  ] as const) {
    await select.getByRole("checkbox").last().focus();
    for (let index = 0; index < tabCount; index++)
      await page.keyboard.press("Tab");
    await expect(remove.nth(tabCount - 1)).toBeFocused();
    const before = await scrollOffsets();
    await page.keyboard.press(position === "middle" ? "Space" : "Enter");
    const focus = await page.evaluate(
      () =>
        document.activeElement?.getAttribute("aria-label") ??
        document.activeElement?.tagName ??
        null,
    );
    const after = await scrollOffsets();
    observations.push({ position, focus, scroll: { before, after } });
    await expect(remove).toHaveCount(expected);
    await expect(page.locator(":focus")).toBeInViewport();
    expect(after.page).toBe(before.page);
    expect(after.grid).toBe(before.grid);
  }
  await info.attach("selected-reference-removal-observations", {
    body: Buffer.from(JSON.stringify({ parentFocus, observations }, null, 2)),
    contentType: "application/json",
  });
  expect(parentFocus).toBe("Choose Linked Records");
  expect(observations.map((item) => item.focus)).toEqual([
    expect.stringMatching(/^Remove selected Linked Records /),
    expect.stringMatching(/^Remove selected Linked Records /),
    expect.stringMatching(/^Remove selected Linked Records /),
    fallbackName,
  ]);
  await button(picker, "Apply references").click();
  await expect(title).toHaveValue("Retained keyboard authoring");
  await expect(parentRemove).toHaveCount(0);

  await choose.click();
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(partiesViewSchemaId);
  await activateCandidateIdentities(select, firstPageIds.slice(0, 2));
  await expect(remove).toHaveCount(2);
  await button(picker, "Cancel references").click();
  await expect(parentRemove).toHaveCount(0);
  await expect(title).toHaveValue("Retained keyboard authoring");

  await choose.click();
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(partiesViewSchemaId);
  await activateCandidateIdentities(select, firstPageIds.slice(0, 2));
  await button(picker, "Apply references").click();
  await expect(parentRemove).toHaveCount(2);
  await choose.focus();
  await page.keyboard.press("Shift+Tab");
  await page.keyboard.press("Shift+Tab");
  await expect(parentRemove.last()).toBeFocused();
  await page.keyboard.press("Space");
  await expect(parentRemove).toHaveCount(1);
  await expect(parentRemove).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(parentRemove).toHaveCount(0);
  await expect(choose).toBeFocused();
  await expect(choose).toBeInViewport();
  await expect(title).toHaveValue("Retained keyboard authoring");
});

test("Note source browsing preserves reviewed identity and text across delayed replacement and cancellation", async ({
  page,
}, info) => {
  const f = await openNoteFixture(page, hostsViewSchemaId);
  await seed(page, f.incident, hostsViewSchemaId, "host.display_name");
  await createViewRow(page, f.incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline"),
    "timeline.activity_synopsis_text": "Replacement surface observation",
  });
  await f.form
    .getByRole("textbox", { name: "Title", exact: true })
    .fill("Retained Note text");
  let release = () => {},
    pending = false;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route(
    `**/incidents/${f.incident}/views/${hostsViewSchemaId}/query`,
    async (route) => {
      if (route.request().postDataJSON().cursor_token) {
        pending = true;
        await gate;
        await route.abort("failed");
      } else await route.continue();
    },
  );
  await button(f.form, "Choose source").click();
  const picker = f.form.getByRole("region", {
    name: "Choose Note source",
    exact: true,
  });
  const select = picker.getByRole("group", {
    name: "Note source",
    exact: true,
  });
  await expect(select.getByRole("radio")).toHaveCount(100);
  const heldNext = button(picker, "Next candidates");
  await heldNext.focus();
  await page.keyboard.press("Enter");
  await expect.poll(() => pending).toBe(true);
  await expect(heldNext).toBeFocused();
  await expect(heldNext).toHaveAttribute("aria-busy", "true");
  await expect(heldNext).toHaveAttribute("aria-disabled", "true");
  await expect(button(picker, "Apply source")).toBeEnabled();
  await picker
    .getByRole("combobox", { name: "Source sheet", exact: true })
    .selectOption(timelineViewSchemaId);
  await expect(select.getByRole("radio")).toHaveCount(1);
  const externalFocus = button(picker, "Apply source");
  await externalFocus.focus();
  await expect(externalFocus).toBeFocused();
  release();
  await expect(
    picker.getByRole("button", { name: /^Remove selected Note source / }),
  ).toHaveCount(1);
  await expect(picker.getByRole("alert")).toHaveCount(0);
  await expect(externalFocus).toBeFocused();
  await button(picker, "Apply source").click();
  await expect(
    f.form.getByRole("textbox", { name: "Title", exact: true }),
  ).toHaveValue("Retained Note text");
  await button(f.form, "Choose source").click();
  await expect(
    picker.getByRole("combobox", { name: "Source sheet", exact: true }),
  ).toHaveValue(hostsViewSchemaId);
  await expect(select.getByRole("radio")).toHaveCount(100);
  await chooseFirst(select);
  let releaseClosedRead!: () => void;
  let requestedClosedRead!: () => void;
  const closedRead = new Promise<void>((resolve) => {
    releaseClosedRead = resolve;
  });
  const closedRequest = new Promise<void>((resolve) => {
    requestedClosedRead = resolve;
  });
  await page.route(
    `**/incidents/${f.incident}/views/${hostsViewSchemaId}/query`,
    async (route) => {
      requestedClosedRead();
      await closedRead;
      await route.continue().catch(() => {});
    },
  );
  const closedRefresh = button(picker, "Refresh candidates");
  await closedRefresh.focus();
  await page.keyboard.press("Enter");
  await closedRequest;
  await expect(closedRefresh).toBeFocused();
  await expect(closedRefresh).toHaveAttribute("aria-busy", "true");
  await page.keyboard.press("Escape");
  await expect(button(f.form, "Choose source")).toBeFocused();
  releaseClosedRead();
  await expect(button(f.form, "Choose source")).toBeFocused();
  await button(f.form, "Choose source").click();
  await expect(
    picker.getByRole("button", {
      name: /^Remove selected Note source Reviewed investigation source/,
    }),
  ).toHaveCount(1);
  await page.setViewportSize({ width: 390, height: 640 });
  await button(picker, "Cancel source").scrollIntoViewIfNeeded();
  await expect(button(picker, "Cancel source")).toBeInViewport();
  await info.attach("Note-source-candidate-narrow-review", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await button(picker, "Cancel source").click();
  const clearSource = button(f.form, "Clear source");
  await clearSource.focus();
  await page.keyboard.press("Space");
  await expect(button(f.form, "Choose source")).toBeFocused();
  await expect(button(f.form, "Choose source")).toBeInViewport();
  await expect(
    f.form.getByRole("textbox", { name: "Title", exact: true }),
  ).toHaveValue("Retained Note text");
});

test("Assessment subjects and Timeline support retain deliberate identities through page eviction and explicit filtering", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("ACD-ASSESSMENT"),
    "Bounded Assessment discovery",
  );
  const hosts = await seed(
    page,
    incident,
    hostsViewSchemaId,
    "host.display_name",
  );
  await seed(
    page,
    incident,
    timelineViewSchemaId,
    "timeline.activity_synopsis_text",
  );
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${assessmentsViewSchemaId}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(assessmentsViewSchemaId)),
  ).toBeVisible();
  await page
    .getByTestId(workbookAddRowButtonTestId(assessmentsViewSchemaId))
    .click();
  const subject = page.getByTestId(assessmentCreateControlTestId("subject"));
  await expect(subject.getByRole("radio")).toHaveCount(100);
  const subjectId = await chooseFirst(subject);
  expect(hosts).toContain(subjectId);
  const rationale = page.getByTestId(
    assessmentCreateControlTestId("rationale"),
  );
  await rationale.fill("Retained assessment rationale");
  await button(page, "Next candidates").click();
  await expect(subject.getByRole("radio")).toHaveCount(5);
  await expect(
    page.getByRole("button", { name: /^Remove selected Subject / }),
  ).toHaveCount(1);
  await subject.getByRole("radio").last().focus();
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("button", { name: /^Remove selected Subject / }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(subject.getByRole("radio").first()).toBeFocused();
  await chooseFirst(subject);
  await expect(rationale).toHaveValue("Retained assessment rationale");
  await page.getByText("Subject ordering and filters", { exact: true }).click();
  await page
    .getByRole("combobox", { name: "Subject filter field", exact: true })
    .selectOption("host.host_state");
  await page
    .getByRole("combobox", { name: "Subject filter operator", exact: true })
    .selectOption("eq");
  await page
    .getByRole("textbox", { name: "Subject filter value", exact: true })
    .fill("excluded");
  await button(page, "Add filter").click();
  await expect(subject.getByRole("radio")).toHaveCount(5);
  await button(page, "Apply candidate query").click();
  await expect(subject.getByRole("radio")).toHaveCount(0);
  await expect(
    page.getByRole("button", { name: /^Remove selected Subject / }),
  ).toHaveCount(1);
  await button(page, "Choose support").click();
  const support = page.getByRole("group", {
    name: "Choose assessment support",
    exact: true,
  });
  const candidates = support.getByRole("group", {
    name: "Timeline support candidates",
    exact: true,
  });
  await expect(candidates.getByRole("checkbox")).toHaveCount(100);
  let releaseSupport!: () => void;
  let requestedSupport!: () => void;
  const supportGate = new Promise<void>((resolve) => {
    releaseSupport = resolve;
  });
  const supportRequested = new Promise<void>((resolve) => {
    requestedSupport = resolve;
  });
  await page.route(
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
    async (route) => {
      requestedSupport();
      await supportGate;
      await route.continue();
    },
  );
  const supportRefresh = button(support, "Refresh candidates");
  await supportRefresh.focus();
  await page.keyboard.press("Enter");
  await supportRequested;
  await expect(supportRefresh).toBeFocused();
  await expect(supportRefresh).toHaveAttribute("aria-busy", "true");
  await expect(supportRefresh).toHaveAttribute("aria-disabled", "true");
  releaseSupport();
  await expect(supportRefresh).toBeFocused();
  await expect(supportRefresh).toHaveAttribute("aria-busy", "false");
  await page.unroute(
    `**/incidents/${incident}/views/${timelineViewSchemaId}/query`,
  );
  const first = await chooseFirst(candidates, true);
  await button(support, "Next candidates").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(5);
  const second = await chooseFirst(candidates, true);
  expect(second).not.toBe(first);
  await expect(
    support.getByRole("button", {
      name: /^Remove selected Timeline support candidates /,
    }),
  ).toHaveCount(2);
  const supportRemove = support.getByRole("button", {
    name: /^Remove selected Timeline support candidates /,
  });
  await candidates.getByRole("checkbox").last().focus();
  await page.keyboard.press("Tab");
  await expect(supportRemove.first()).toBeFocused();
  await page.keyboard.press("Space");
  await expect(supportRemove).toHaveCount(1);
  await expect(supportRemove).toBeFocused();
  await button(support, "Cancel support selection").click();
  await expect(button(page, "Choose support")).toBeFocused();
  await button(page, "Choose support").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(100);
  const retainedSupport = await candidates
    .getByRole("checkbox")
    .evaluateAll((options) =>
      options.slice(0, 2).map((option) => (option as HTMLInputElement).value),
    );
  await activateCandidateIdentities(candidates, retainedSupport);
  await button(support, "Apply support selection").click();
  await expect(
    page.getByRole("region", {
      name: "Assessment supporting records",
      exact: true,
    }),
  ).toContainText("Supporting records (2/64)");
  await button(page, "Choose support").click();
  await expect(candidates.getByRole("checkbox")).toHaveCount(100);
  await activateCandidateIdentities(candidates, []);
  await candidates.press("Escape");
  await expect(button(page, "Choose support")).toBeFocused();
  await expect(
    page.getByRole("region", {
      name: "Assessment supporting records",
      exact: true,
    }),
  ).toContainText("Supporting records (2/64)");
  await expect(rationale).toHaveValue("Retained assessment rationale");
  await info.attach("Assessment-retained-candidates-review", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
});

test("Single-target equal-label identities remain inspectable through contextual acceptance and explicit Party linking", async ({
  page,
}, info) => {
  const f = await openContextualCreationFixture(
    page,
    "task_request",
    undefined,
    {
      viewSchemaId: timelineViewSchemaId,
      label: "Identity-aware requester source",
    },
  );
  const label =
    "Response coordination team — " +
    "long readable requester label ".repeat(4) +
    "RequesterIdentityToken".repeat(4);
  const parties = [];
  for (let index = 0; index < 2; index++)
    parties.push(
      await createViewRow(page, f.incident, partiesViewSchemaId, {
        client_txn_id: uniqueTxn(`same-label-${index}`),
        "party.display_name": label,
        "party.party_kind": "team",
      }),
    );
  const [first, chosen] = parties;
  if (!first || !chosen) throw new Error("Missing equal-label Parties");
  const task = page.getByRole("region", {
    name: "Create task request",
    exact: true,
  });
  const wording = "  Requester wording remains independent  ";
  const raw = task.getByTestId(
    genericCreateFieldTestId("task.requester_party_text"),
  );
  await raw.fill(wording);
  const choose = button(task, "Choose Requester Party");
  const picker = task.getByRole("region", {
    name: "Choose Requester Party",
    exact: true,
  });
  const candidates = picker.getByRole("group", {
    name: "Requester Party",
    exact: true,
  });
  const selected = task.getByRole("group", {
    name: "Requester Party selected references",
    exact: true,
  });
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" ||
      (request.method() === "POST" &&
        request.url().includes(`/views/${taskRequestsViewSchemaId}/rows`))
    )
      writes++;
  });
  for (const viewport of [
    { width: 1440, height: 900 },
    { width: 1024, height: 768 },
    { width: 768, height: 640 },
  ]) {
    await page.setViewportSize(viewport);
    await choose.click();
    const radio = candidates.getByRole("radio", {
      name: `${label.trim()} (${chosen.record_id})`,
      exact: true,
    });
    await expect(candidates.getByRole("radio")).toHaveCount(2);
    await expect(candidates.getByRole("radio", { checked: true })).toHaveCount(
      0,
    );
    for (const party of parties)
      await expect(
        candidates.getByText(party.record_id, { exact: true }),
      ).toBeVisible();
    await expect(candidates.getByText(label, { exact: true })).toHaveCount(2);
    expect(
      await candidates.evaluate(
        (element) => element.scrollWidth <= element.clientWidth + 1,
      ),
    ).toBe(true);
    await radio.focus();
    await expect(radio).not.toBeChecked();
    await page.keyboard.press("Space");
    await expect(radio).toBeChecked();
    await page.keyboard.press("Enter");
    expect(writes).toBe(0);
    await button(picker, "Cancel references").click();
    await expect(choose).toBeFocused();
    await expect(raw).toHaveValue(wording);
    await expect(selected).toContainText("No references selected");
  }
  await choose.click();
  await candidates.locator(`input[value="${chosen.record_id}"]`).check();
  await button(picker, "Apply references").click();
  await expect(selected).toContainText(chosen.record_id);
  await expect(selected).toContainText(label.trim());
  await expect(raw).toHaveValue(wording);
  expect(writes).toBe(0);
  expect(
    await queryViewRows(page, f.incident, taskRequestsViewSchemaId),
  ).toHaveLength(0);
  await task
    .getByTestId(genericCreateSubmitTestId(taskRequestsViewSchemaId))
    .click();
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, f.incident, taskRequestsViewSchemaId))
          .length,
    )
    .toBe(1);
  const created = (
    await queryViewRows(page, f.incident, taskRequestsViewSchemaId)
  )[0];
  if (!created) throw new Error("Missing created task");
  expect(created.cells["task.requester_party_id"]?.value).toBe(
    chosen.record_id,
  );
  expect(created.cells["task.requester_party_text"]?.value).toBe(
    wording.trim(),
  );
  await switchOrdinarySheet(page, taskRequestsViewSchemaId);
  await openGenericInspectorForRecord(
    page,
    taskRequestsViewSchemaId,
    created.record_id,
  );
  await page
    .getByTestId(coordinationWorkflowTestId("party-pair"))
    .selectOption("task.requester_party_text:task.requester_party_id");
  const existing = page.getByTestId(
    coordinationWorkflowTestId("party-existing"),
  );
  const target = existing.locator(`input[value="${first.record_id}"]`);
  await target.check();
  const writesBeforeLink = writes;
  await page
    .getByRole("button", { name: "Clear selected party", exact: true })
    .click();
  await expect(existing.getByRole("radio").first()).toBeFocused();
  await target.focus();
  await page.keyboard.press("Space");
  await expect(target).toBeChecked();
  expect(writes).toBe(writesBeforeLink);
  await page
    .getByTestId(coordinationWorkflowTestId("party-link-existing"))
    .click();
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, f.incident, taskRequestsViewSchemaId))[0]
          ?.cells["task.requester_party_id"]?.value,
    )
    .toBe(first.record_id);
  expect(
    (await queryViewRows(page, f.incident, taskRequestsViewSchemaId))[0]?.cells[
      "task.requester_party_text"
    ]?.value,
  ).toBe(wording.trim());
  await page.setViewportSize({ width: 1440, height: 900 });
  const linked = page.getByRole("region", {
    name: "Requester Party",
    exact: true,
  });
  await expect(linked).toContainText(first.record_id);
  expect(
    await linked.evaluate(
      (element) => element.scrollWidth <= element.clientWidth + 1,
    ),
  ).toBe(true);
  await linked
    .locator("p")
    .filter({ hasText: "Party link:" })
    .scrollIntoViewIfNeeded();
  await info.attach("single-target-linked-long-identity", {
    body: await page.screenshot({ animations: "disabled", caret: "hide" }),
    contentType: "image/png",
  });
});
