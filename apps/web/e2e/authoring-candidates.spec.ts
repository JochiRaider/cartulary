import {
  assessmentCreateControlTestId,
  genericCreateFieldTestId,
  gridShellTestId,
  workbookAddRowButtonTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  partiesViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { openNoteFixture } from "./support/workbook/noteCreate";
import {
  ordinaryField,
  switchOrdinarySheet,
} from "./support/workbook/ordinaryCreate";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openTimelineEvidenceFixture } from "./support/workbook/timelineRelatedEvidence";

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
  const option = select.getByRole("option").nth(multiple ? 0 : 1);
  await expect(option).toBeAttached();
  const id = await option.getAttribute("value");
  if (!id) throw new Error("Missing exact candidate identity");
  await select.selectOption(id);
  return id;
}
function button(scope: Locator | Page, name: string) {
  return scope.getByRole("button", { name, exact: true });
}

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
  const select = picker.getByRole("combobox", {
    name: "Collector Party",
    exact: true,
  });
  await expect(select.getByRole("option")).toHaveCount(101);
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
  await expect(select.getByRole("option")).toHaveCount(7);
  expect(reads.at(-1)).toEqual(failedRead);
  await expect(
    picker.getByRole("button", { name: /^Remove selected Collector Party / }),
  ).toHaveCount(1);
  await button(picker, "Apply Party").click();
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
    source.getByRole("combobox", { name: "Source Party", exact: true }),
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
    name: "Create Related Task Request",
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
  const records = references.getByRole("listbox", {
    name: "Linked Records",
    exact: true,
  });
  await expect(records.getByRole("option")).toHaveCount(100);
  await chooseFirst(records, true);
  await button(references, "Next candidates").click();
  await expect(records.getByRole("option")).toHaveCount(6);
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
  const ordinaryCandidates = ordinary.getByRole("combobox", {
    name: "Collector Party",
    exact: true,
  });
  await expect(ordinaryCandidates.getByRole("option")).toHaveCount(101);
  await button(ordinary, "Next candidates").click();
  await expect(ordinaryCandidates.getByRole("option")).toHaveCount(7);
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
  const select = picker.getByRole("combobox", {
    name: "Note source",
    exact: true,
  });
  await expect(select.getByRole("option")).toHaveCount(101);
  await button(picker, "Next candidates").click();
  await expect.poll(() => pending).toBe(true);
  await expect(button(picker, "Apply source")).toBeEnabled();
  await picker
    .getByRole("combobox", { name: "Source sheet", exact: true })
    .selectOption(timelineViewSchemaId);
  await expect(select.getByRole("option")).toHaveCount(2);
  release();
  await expect(
    picker.getByRole("button", { name: /^Remove selected Note source / }),
  ).toHaveCount(1);
  await expect(picker.getByRole("alert")).toHaveCount(0);
  await button(picker, "Apply source").click();
  await expect(
    f.form.getByRole("textbox", { name: "Title", exact: true }),
  ).toHaveValue("Retained Note text");
  await button(f.form, "Choose source").click();
  await expect(
    picker.getByRole("combobox", { name: "Source sheet", exact: true }),
  ).toHaveValue(hostsViewSchemaId);
  await expect(select.getByRole("option")).toHaveCount(101);
  await chooseFirst(select);
  await select.press("Escape");
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
  await expect(subject.getByRole("option")).toHaveCount(101);
  const subjectId = await chooseFirst(subject);
  expect(hosts).toContain(subjectId);
  const rationale = page.getByTestId(
    assessmentCreateControlTestId("rationale"),
  );
  await rationale.fill("Retained assessment rationale");
  await button(page, "Next candidates").click();
  await expect(subject.getByRole("option")).toHaveCount(6);
  await expect(
    page.getByRole("button", { name: /^Remove selected Subject / }),
  ).toHaveCount(1);
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
  await expect(subject.getByRole("option")).toHaveCount(6);
  await button(page, "Apply candidate query").click();
  await expect(subject.getByRole("option")).toHaveCount(1);
  await expect(
    page.getByRole("button", { name: /^Remove selected Subject / }),
  ).toHaveCount(1);
  await button(page, "Choose support").click();
  const support = page.getByRole("group", {
    name: "Choose assessment support",
    exact: true,
  });
  const candidates = support.getByRole("listbox", {
    name: "Timeline support candidates",
    exact: true,
  });
  await expect(candidates.getByRole("option")).toHaveCount(100);
  const first = await chooseFirst(candidates, true);
  await button(support, "Next candidates").click();
  await expect(candidates.getByRole("option")).toHaveCount(5);
  const second = await chooseFirst(candidates, true);
  expect(second).not.toBe(first);
  await expect(
    support.getByRole("button", {
      name: /^Remove selected Timeline support candidates /,
    }),
  ).toHaveCount(2);
  await button(support, "Apply support selection").click();
  await expect(
    page.getByRole("region", {
      name: "Assessment supporting records",
      exact: true,
    }),
  ).toContainText("Supporting records (2/64)");
  await button(page, "Choose support").click();
  await expect(candidates.getByRole("option")).toHaveCount(100);
  await candidates.selectOption([]);
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
