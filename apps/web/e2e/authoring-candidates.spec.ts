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
    const select = picker.getByRole("combobox", { name: "Owner", exact: true });
    await expect(select).toBeEnabled();
    await expect(select).toHaveValue(actor.user_id);
    await expect(select.locator(`option[value="${actor.user_id}"]`)).toHaveText(
      actor.display_name,
    );
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
    await select.focus();
    await select.press("Escape");
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
      recovery.getByRole("combobox", { name: "Owner", exact: true }),
    ).toHaveValue(actor.user_id);
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
  const select = picker.getByRole("combobox", {
    name: "Requester Party",
    exact: true,
  });
  await expect(select.getByRole("option")).toHaveCount(101);
  const id = await chooseFirst(select);
  const original = await select.locator(`option[value="${id}"]`).innerText();
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
  await expect(select.locator(`option[value="${id}"]`)).toHaveText(renamed);
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
  await expect(select.locator(`option[value="${id}"]`)).toHaveCount(0);
  await expect(remove).toHaveAccessibleName(
    `Remove selected Requester Party ${renamed}`,
  );
  await button(picker, "Previous candidates").click();
  await expect(select).toHaveValue(id);
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
  const compactRemove = ordinary.getByRole("button", {
    name: /^Remove selected Collector Party /,
  });
  await expect(compactRemove).toHaveCount(1);
  await ordinaryCandidates.focus();
  await page.keyboard.press("Tab");
  await expect(compactRemove).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(ordinaryCandidates).toBeFocused();
  await button(ordinary, "Cancel references").click();
  await expect(raw).toHaveValue(selected);
  await trigger.click();
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
  const candidates = picker.getByRole("combobox", {
    name: "Requester Party",
    exact: true,
  });
  await expect(candidates.getByRole("option")).toHaveCount(101);

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
    await expect(picker.getByRole("status")).toContainText("Page 1");
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
  await expect(candidates.getByRole("option")).toHaveCount(7);
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
  await expect(candidates.getByRole("option")).toHaveCount(101);
  await expect(previous.control).toBeFocused();
  await expect(previous.control).toHaveAttribute("aria-disabled", "true");
  await page.keyboard.press("Tab");
  await expect(previous.control).toBeDisabled();

  const first = await admit("First candidates");
  first.release();
  await expect(picker.getByRole("status")).toContainText("Page 1");
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
  const select = picker.getByRole("listbox", {
    name: "Linked Records",
    exact: true,
  });
  const firstPageIds = await select
    .getByRole("option")
    .evaluateAll((options) =>
      options.slice(0, 3).map((option) => (option as HTMLOptionElement).value),
    );
  await select.selectOption(firstPageIds);
  const remove = picker.getByRole("button", {
    name: /^Remove selected Linked Records /,
  });
  await expect(remove).toHaveCount(3);
  await button(picker, "Next candidates").click();
  await expect(select.getByRole("option")).toHaveCount(6);
  await expect(remove).toHaveCount(3);
  const lastId = await select.getByRole("option").first().getAttribute("value");
  if (!lastId) throw new Error("Missing off-page candidate identity");
  await select.selectOption(lastId);
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
    await select.focus();
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
    "SELECT",
  ]);
  await button(picker, "Apply references").click();
  await expect(title).toHaveValue("Retained keyboard authoring");
  await expect(parentRemove).toHaveCount(0);

  await choose.click();
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(partiesViewSchemaId);
  await select.selectOption(firstPageIds.slice(0, 2));
  await expect(remove).toHaveCount(2);
  await button(picker, "Cancel references").click();
  await expect(parentRemove).toHaveCount(0);
  await expect(title).toHaveValue("Retained keyboard authoring");

  await choose.click();
  await picker
    .getByRole("combobox", { name: "Reference surface", exact: true })
    .selectOption(partiesViewSchemaId);
  await select.selectOption(firstPageIds.slice(0, 2));
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
  const select = picker.getByRole("combobox", {
    name: "Note source",
    exact: true,
  });
  await expect(select.getByRole("option")).toHaveCount(101);
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
  await expect(select.getByRole("option")).toHaveCount(2);
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
  await expect(select.getByRole("option")).toHaveCount(101);
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
  await subject.focus();
  await page.keyboard.press("Tab");
  await expect(
    page.getByRole("button", { name: /^Remove selected Subject / }),
  ).toBeFocused();
  await page.keyboard.press("Enter");
  await expect(subject).toBeFocused();
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
  await expect(candidates.getByRole("option")).toHaveCount(5);
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
  await candidates.focus();
  await page.keyboard.press("Tab");
  await expect(supportRemove.first()).toBeFocused();
  await page.keyboard.press("Space");
  await expect(supportRemove).toHaveCount(1);
  await expect(supportRemove).toBeFocused();
  await button(support, "Cancel support selection").click();
  await expect(button(page, "Choose support")).toBeFocused();
  await button(page, "Choose support").click();
  await expect(candidates.getByRole("option")).toHaveCount(100);
  const retainedSupport = await candidates
    .getByRole("option")
    .evaluateAll((options) =>
      options.slice(0, 2).map((option) => (option as HTMLOptionElement).value),
    );
  await candidates.selectOption(retainedSupport);
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
