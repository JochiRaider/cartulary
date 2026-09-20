import {
  applyFilterChip,
  removeFilterChip,
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  gridScrollportSelector,
  gridShellTestId,
  relationshipItemsTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  workbookInspectorCloseButtonTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { installVisualPreferences } from "./support/auth/visualPreferences";
import {
  editTimelineSummary,
  installPatchController,
} from "./support/collaboration/replay";
import { currentLifecycle, lifecycleAction } from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { createTimelineFillers } from "./support/timeline/fixtures";
import { holdBrowserRequest } from "./support/transport/requestInterception";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import { fetchRecordHistoryCount } from "./support/workbook/history";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { openTimelineInspector } from "./support/workbook/rowMutations";

const summary = "timeline.activity_synopsis_text";
const tagInput = (page: Page) =>
  page.getByRole("textbox", { name: "Tag for selected Timeline records" });
const checkbox = (page: Page, id: string) =>
  page.getByRole("checkbox", { name: `Select record ${id}`, exact: true });
const settle = (page: Page) =>
  page.evaluate(
    () =>
      new Promise<void>((resolve) =>
        requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
      ),
  );

async function seed(page: Page) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("BTI"),
    "Timeline bulk tag interaction",
  );
  const rows = [];
  for (const label of [
    "First selected",
    "Second selected",
    "Unselected record",
  ]) {
    rows.push(
      await createViewRow(page, incident, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("bulk-tag-row"),
        [summary]: label,
        "timeline.activity_utc_text": "2026-04-01T00:00:00Z",
      }),
    );
  }
  await page.goto(`/?incident_id=${incident}`);
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  const first = rows[0],
    second = rows[1];
  if (!first || !second) throw new Error("Missing fixture rows");
  return { incident, first, second };
}

async function observeRenders(page: Page) {
  await page.addInitScript(() => {
    type Fiber = {
      type?: unknown;
      flags: number;
      child?: Fiber;
      sibling?: Fiber;
      memoizedProps?: Record<string, unknown> & { binding?: { kind?: string } };
      alternate?: Fiber;
    };
    const observed = window as unknown as {
      __btiCounts: Record<string, number>;
      __REACT_DEVTOOLS_GLOBAL_HOOK__: unknown;
    };
    const counts: Record<string, number> = {};
    observed.__btiCounts = counts;
    let previous = new WeakSet<object>();
    observed.__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
      supportsFiber: true,
      inject: () => 1,
      onCommitFiberUnmount: () => {},
      onCommitFiberRoot: (_renderer: number, root: { current: Fiber }) => {
        const increment = (key: string) => {
          counts[key] = (counts[key] ?? 0) + 1;
        };
        increment("commits");
        const current = new WeakSet<object>();
        const visit = (fiber: Fiber | undefined) => {
          if (!fiber) return;
          current.add(fiber);
          if (!previous.has(fiber)) {
            if (
              typeof fiber.type === "function" &&
              (fiber.flags & 1) !== 0 &&
              Array.isArray(fiber.memoizedProps?.columns) &&
              Array.isArray(fiber.memoizedProps?.rows)
            )
              increment("gridRenders");
            if (
              typeof fiber.type === "function" &&
              (fiber.flags & 1) !== 0 &&
              fiber.memoizedProps?.binding?.kind === "collection"
            )
              increment("collectionRenders");
            for (const key of ["columns", "rows"])
              if (
                Array.isArray(fiber.memoizedProps?.[key]) &&
                fiber.alternate &&
                fiber.memoizedProps?.[key] !==
                  fiber.alternate.memoizedProps?.[key]
              )
                increment(`${key}Replacements`);
          }
          visit(fiber.child);
          visit(fiber.sibling);
        };
        visit(root.current);
        previous = current;
      },
    };
  });
}

const renderCounts = (page: Page) =>
  page.evaluate(() => ({
    ...(window as unknown as { __btiCounts: Record<string, number> })
      .__btiCounts,
  }));

test("Timeline bulk tag production characterization", async ({
  page,
}, testInfo) => {
  test.setTimeout(180_000);
  await page.setViewportSize({ width: 1440, height: 900 });
  await observeRenders(page);
  const { incident, first, second } = await seed(page);
  const requests: string[] = [];
  page.on("request", (request) => {
    if (new URL(request.url()).pathname.startsWith("/api/"))
      requests.push(`${request.method()} ${new URL(request.url()).pathname}`);
  });
  const observations: unknown[] = [];
  const snapshot = async (phase: string) => {
    const selected = [];
    for (const row of [first, second])
      selected.push({
        recordId: row.record_id,
        checkbox: await checkbox(page, row.record_id).count(),
        selected:
          (await checkbox(page, row.record_id).count()) > 0 &&
          (await checkbox(page, row.record_id).isChecked()),
      });
    observations.push({
      phase,
      selected,
      input: await tagInput(page)
        .evaluate((node: HTMLInputElement) => ({
          text: node.value,
          focused: document.activeElement === node,
          start: node.selectionStart,
          end: node.selectionEnd,
          direction: node.selectionDirection,
        }))
        .catch(() => null),
      scroll: await page
        .locator(gridScrollportSelector())
        .evaluate((node) => ({ left: node.scrollLeft, top: node.scrollTop })),
      grid: await page
        .getByTestId(gridShellTestId(timelineViewSchemaId))
        .boundingBox(),
      bar: await page
        .getByRole("region", { name: "Workbook query and action controls" })
        .boundingBox(),
      requests: [...requests],
    });
  };
  await checkbox(page, first.record_id).check();
  await checkbox(page, second.record_id).check();
  await tagInput(page).fill("  retained tag Ω 東京  ");
  await tagInput(page).focus();
  await settle(page);
  const before = await renderCounts(page),
    requestCount = requests.length;
  await tagInput(page).pressSequentially("0123456789");
  await settle(page);
  const after = await renderCounts(page);
  observations.push({
    phase: "typing-diagnostics-not-latency",
    requests: requests.length - requestCount,
    work: Object.fromEntries(
      [...new Set([...Object.keys(before), ...Object.keys(after)])].map(
        (key) => [key, (after[key] ?? 0) - (before[key] ?? 0)],
      ),
    ),
  });
  await tagInput(page).evaluate((node: HTMLInputElement) => {
    node.setSelectionRange(3, 8, "backward");
    (window as unknown as { __btiInput: HTMLInputElement }).__btiInput = node;
  });
  await patchRecord(page, first.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: first.row_version,
    client_txn_id: uniqueTxn("bulk-tag-live"),
    changes: [{ field_key: "timeline.analyst_text", value: "Another analyst" }],
  });
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incident, timelineViewSchemaId)).find(
          (row) => row.record_id === first.record_id,
        )?.row_version,
    )
    .toBeGreaterThan(first.row_version);
  await settle(page);
  await snapshot("passive-row-update");
  observations.push({
    phase: "input-identity",
    sameNode: await tagInput(page).evaluate(
      (node) =>
        node ===
        (window as unknown as { __btiInput: HTMLInputElement }).__btiInput,
    ),
  });
  await checkbox(page, second.record_id).uncheck();
  await tagInput(page).focus();
  await tagInput(page).evaluate((node: HTMLInputElement) =>
    node.setSelectionRange(3, 8, "backward"),
  );
  await patchRecord(page, first.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: first.row_version + 1,
    client_txn_id: uniqueTxn("bulk-tag-single-live"),
    changes: [{ field_key: "timeline.analyst_text", value: "Updated analyst" }],
  });
  await settle(page);
  await expect(tagInput(page)).toBeFocused();
  await snapshot("single-selected-row-update");
  await checkbox(page, second.record_id).check();
  for (const width of [1440, 1024, 768]) {
    await page.setViewportSize({ width, height: 900 });
    await tagInput(page).fill("long-tag-Ω".repeat(35));
    await settle(page);
    await snapshot(`layout-${width}`);
    await testInfo.attach(`bulk-tag-workbook-${width}`, {
      body: await page.screenshot(),
      contentType: "image/png",
    });
  }
  await page.setViewportSize({ width: 1440, height: 900 });
  const patches = await installPatchController(page);
  try {
    const held = patches.holdNextPatch({ recordId: first.record_id });
    await editTimelineSummary(page, first.record_id, "Selected edit held", {
      outcome: "queued",
    });
    await held.waitForHit;
    await snapshot("held-save");
    held.release();
    await held.waitForCompletion;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await snapshot("acknowledged-save");
    await checkbox(page, first.record_id).check();
    patches.failNextPatch(409, "client_txn_conflict", {
      recordId: first.record_id,
    });
    await editTimelineSummary(page, first.record_id, "Rejected selected edit", {
      outcome: "queued",
    });
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await snapshot("rejected-save");
  } finally {
    await patches.dispose();
  }
  await testInfo.attach("bulk-tag-observations", {
    body: JSON.stringify(observations, null, 2),
    contentType: "application/json",
  });
});

test("Timeline selected records survive autosave and explicit tagging waits for the complete captured set", async ({
  page,
}) => {
  const { incident, first, second } = await seed(page);
  const requests: Record<string, unknown>[] = [];
  page.on("request", (request) => {
    if (request.url().endsWith("/bulk-mutations"))
      requests.push(request.postDataJSON());
  });
  const patches = await installPatchController(page);
  try {
    await checkbox(page, first.record_id).check();
    await checkbox(page, second.record_id).check();
    await tagInput(page).fill("  triaged Ω  ");
    const held = patches.holdNextPatch({ recordId: first.record_id });
    await editTimelineSummary(page, first.record_id, "Edited while selected", {
      outcome: "queued",
    });
    await held.waitForHit;
    await expect(checkbox(page, first.record_id)).toBeChecked();
    await expect(checkbox(page, second.record_id)).toBeChecked();
    await expect(tagInput(page)).toHaveValue("  triaged Ω  ");
    await tagInput(page).press("Enter");
    await expect(
      page.getByRole("form", { name: "Timeline bulk record actions" }),
    ).toContainText("waiting for earlier work");
    expect(requests).toHaveLength(0);
    await checkbox(page, first.record_id).uncheck();
    await tagInput(page).fill("newer unsubmitted text");
    await tagInput(page).focus();
    await tagInput(page).evaluate((node: HTMLInputElement) =>
      node.setSelectionRange(2, 7, "backward"),
    );
    const scrollBefore = await page
      .locator(gridScrollportSelector())
      .evaluate((node) => [node.scrollLeft, node.scrollTop]);
    const response = page.waitForResponse(
      (response) =>
        response.url().endsWith("/bulk-mutations") &&
        response.request().method() === "POST",
    );
    held.release();
    await held.waitForCompletion;
    expect((await response).ok()).toBe(true);
    expect(requests).toHaveLength(1);
    expect(requests[0]).toMatchObject({
      kind: "multi_row_tag_assignment_v1",
      tag_name: "triaged Ω",
      targets: expect.arrayContaining([
        { record_id: first.record_id, base_row_version: first.row_version + 1 },
        { record_id: second.record_id, base_row_version: second.row_version },
      ]),
    });
    await expect(tagInput(page)).toHaveValue("newer unsubmitted text");
    await expect(tagInput(page)).toBeFocused();
    expect(
      await page
        .locator(gridScrollportSelector())
        .evaluate((node) => [node.scrollLeft, node.scrollTop]),
    ).toEqual(scrollBefore);
    expect(
      await tagInput(page).evaluate((node: HTMLInputElement) => [
        node.selectionStart,
        node.selectionEnd,
        node.selectionDirection,
      ]),
    ).toEqual([2, 7, "backward"]);
    await expect(checkbox(page, first.record_id)).not.toBeChecked();
    await expect(checkbox(page, second.record_id)).toBeChecked();
    const saved = await queryViewRows(page, incident, timelineViewSchemaId);
    for (const id of [first.record_id, second.record_id])
      expect(
        JSON.stringify(
          saved.find((row) => row.record_id === id)?.cells["timeline.tags"],
        ),
      ).toContain("triaged Ω");
    await settle(page);
    expect(requests).toHaveLength(1);
    // A later captured prerequisite can fail; readiness must never bypass it.
    await checkbox(page, first.record_id).check();
    const failedPrerequisite = patches.holdNextPatch({
      recordId: first.record_id,
    });
    await editTimelineSummary(page, first.record_id, "Later selected edit", {
      outcome: "queued",
    });
    await failedPrerequisite.waitForHit;
    await tagInput(page).fill("waiting behind failed edit");
    await tagInput(page).press("Enter");
    await expect(
      page.getByRole("form", { name: "Timeline bulk record actions" }),
    ).toContainText("waiting for earlier work");
    const version = (
      await queryViewRows(page, incident, timelineViewSchemaId)
    ).find((row) => row.record_id === first.record_id)?.row_version;
    if (!version) throw new Error("Missing current prerequisite version");
    await patchRecord(page, first.record_id, {
      view_schema_id: timelineViewSchemaId,
      client_txn_id: uniqueTxn("failed-prerequisite"),
      base_row_version: version,
      changes: [{ field_key: summary, value: "Conflicting server edit" }],
    });
    failedPrerequisite.release();
    await failedPrerequisite.waitForCompletion;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await expect(
      page.getByRole("form", { name: "Timeline bulk record actions" }),
    ).toContainText("needs recovery");
    await expect(checkbox(page, first.record_id)).toBeChecked();
    await expect(checkbox(page, second.record_id)).toBeChecked();
    await expect(tagInput(page)).toHaveValue("waiting behind failed edit");
    expect(requests).toHaveLength(1);
  } finally {
    await patches.dispose();
  }
});

test("Timeline failed selected edits retain authoring and block tag dispatch without hiding controls", async ({
  page,
}, testInfo) => {
  const { first, second } = await seed(page);
  let batches = 0;
  page.on("request", (request) => {
    if (request.url().endsWith("/bulk-mutations")) batches++;
  });
  const patches = await installPatchController(page);
  try {
    await checkbox(page, first.record_id).check();
    await checkbox(page, second.record_id).check();
    const raw = "  long retained tag Ω 東京 ".repeat(20);
    await tagInput(page).fill(raw);
    patches.failNextPatch(409, "client_txn_conflict", {
      recordId: first.record_id,
    });
    await editTimelineSummary(
      page,
      first.record_id,
      "Failed edit still selected",
      { outcome: "queued" },
    );
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await expect(checkbox(page, first.record_id)).toBeChecked();
    await expect(checkbox(page, second.record_id)).toBeChecked();
    await expect(tagInput(page)).toHaveValue(raw);
    const form = page.getByRole("form", {
      name: "Timeline bulk record actions",
    });
    await expect(form).toContainText("needs recovery");
    await expect(
      form.getByRole("button", { name: "Assign tag", exact: true }),
    ).toBeDisabled();
    for (const [width, zoom] of [
      [1440, 1],
      [1024, 1],
      [768, 1],
      [1440, 2],
    ] as const) {
      await page.setViewportSize({ width, height: 900 });
      await page.evaluate((zoom) => {
        document.documentElement.style.zoom = String(zoom);
      }, zoom);
      await tagInput(page).focus();
      await tagInput(page).evaluate((node: HTMLInputElement) =>
        node.setSelectionRange(3, 8, "backward"),
      );
      await settle(page);
      await expect(tagInput(page)).toBeFocused();
      await expect(tagInput(page)).toHaveValue(raw);
      const query = page.getByRole("region", {
        name: "Workbook query and action controls",
      });
      const queryBox = await query.boundingBox(),
        formBox = await tagInput(page).boundingBox();
      expect(
        queryBox && formBox && formBox.y >= queryBox.y + queryBox.height,
      ).toBe(true);
      for (const name of [
        "Find in loaded rows",
        "Clear contents",
        "Open inspector",
        "Add row",
      ])
        await expect(
          query.getByRole("button", { name, exact: true }),
        ).toBeVisible();
      expect(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= window.innerWidth,
        ),
      ).toBe(true);
      await testInfo.attach(`bulk-tag-rejection-${width}-${zoom}`, {
        body: await page.screenshot(),
        contentType: "image/png",
      });
    }
    expect(batches).toBe(0);
  } finally {
    await patches.dispose();
  }
});

test("Timeline bulk selection follows accepted query windows without expansion and prunes lost authority", async ({
  page,
}) => {
  test.setTimeout(240_000);
  const { incident, first, second } = await seed(page);
  const form = page.getByRole("form", { name: "Timeline bulk record actions" });
  const browsing = page.getByRole("group", { name: "Workbook browsing" });
  await checkbox(page, first.record_id).check();
  await checkbox(page, second.record_id).check();
  await tagInput(page).fill("retained across membership");
  const queryPath = `**/api/v1/incidents/${incident}/views/${timelineViewSchemaId}/query`;
  await page.route(queryPath, (route) => route.abort("failed"));
  await browsing.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(
    page.locator('[data-grid-data-state="stale_error"]'),
  ).toBeVisible();
  await expect(checkbox(page, first.record_id)).toBeChecked();
  await expect(checkbox(page, second.record_id)).toBeChecked();
  await page.unroute(queryPath);
  await createTimelineFillers(page, incident, "bulk-window", 397, {
    occurredAtStart: "2026-04-02T00:00:00Z",
  });
  await browsing.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(browsing).toContainText("100 records loaded");
  await expect(form).toContainText("2 selected");
  for (const count of [200, 300]) {
    await browsing
      .getByRole("button", { name: "Load more", exact: true })
      .click();
    await expect(browsing).toContainText(`${count} records loaded`);
    await expect(form).toContainText("2 selected");
  }
  await browsing
    .getByRole("button", { name: "Load more", exact: true })
    .click();
  await expect(form).toContainText("0 selected");
  await expect(tagInput(page)).toHaveValue("retained across membership");
  await browsing
    .getByRole("button", { name: "Earlier rows", exact: true })
    .click();
  await expect(checkbox(page, first.record_id)).not.toBeChecked();
  await checkbox(page, first.record_id).check();
  await checkbox(page, second.record_id).check();
  await patchRecord(page, second.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: second.row_version,
    client_txn_id: uniqueTxn("bulk-membership-tag"),
    changes: [
      {
        field_key: "timeline.tags",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [{ op: "add_tag", tag_name: "membership" }],
        },
      },
    ],
  });
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.tags",
    "membership",
  );
  await expect(browsing).toContainText("1 records loaded");
  await expect(form).toContainText("1 selected");
  await expect(checkbox(page, second.record_id)).toBeChecked();
  await removeFilterChip(page, timelineViewSchemaId, "timeline.tags");
  await expect(checkbox(page, first.record_id)).not.toBeChecked();
  await expect(checkbox(page, second.record_id)).toBeChecked();
  const lifecycle = await currentLifecycle(page, incident);
  expect(
    (
      await lifecycleAction(page, incident, "closeIncident", {
        client_txn_id: uniqueTxn("bulk-tag-close"),
        base_incident_version: lifecycle.incident_version,
        reason: "Bulk selection authority",
      })
    ).ok,
  ).toBe(true);
  await expect(form).toContainText("0 selected");
  await expect(tagInput(page)).toHaveValue("retained across membership");
  await expect(tagInput(page)).toHaveAttribute("readonly", "");
  await expect(
    form.getByRole("button", { name: "Assign tag", exact: true }),
  ).toBeDisabled();
  await expect(
    page.getByRole("checkbox", {
      name: "Select all loaded records",
      exact: true,
    }),
  ).toHaveCount(0);
});

test("Timeline bulk tag delivery identity and ordinary collection entry remain independent", async ({
  page,
}) => {
  const { incident, first, second } = await seed(page);
  const batches: Record<string, unknown>[] = [];
  page.on("request", (request) => {
    if (request.url().endsWith("/bulk-mutations"))
      batches.push(request.postDataJSON());
  });
  await checkbox(page, first.record_id).check();
  await checkbox(page, second.record_id).check();
  await showTimelineCollectionColumns(page, ["Tags"]);
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: relationshipItemsTestId(
      first.record_id,
      "timeline.tags",
      "grid",
    ),
  });
  const cell = page
    .getByRole("group", { name: "Tags collection cell", exact: true })
    .filter({
      has: page.getByTestId(
        relationshipItemsTestId(first.record_id, "timeline.tags", "grid"),
      ),
    });
  await cell
    .getByRole("button", { name: "Add tags token", exact: true })
    .click();
  const collection = page.getByTestId(
    timelineCollectionInputTestId(first.record_id, "timeline.tags", "grid"),
  );
  await collection.fill("ordinary collection tag");
  await collection.press("Enter");
  await expect
    .poll(async () =>
      JSON.stringify(
        (await queryViewRows(page, incident, timelineViewSchemaId)).find(
          (row) => row.record_id === first.record_id,
        )?.cells["timeline.tags"],
      ),
    )
    .toContain("ordinary collection tag");
  await expect(checkbox(page, first.record_id)).toBeChecked();
  await expect(checkbox(page, second.record_id)).toBeChecked();
  await tagInput(page).fill("repeated tag");
  const form = page.getByRole("form", { name: "Timeline bulk record actions" });
  await expect(
    form.getByRole("button", { name: "Assign tag", exact: true }),
  ).toBeEnabled();
  const response = () =>
    page.waitForResponse(
      (response) =>
        response.url().endsWith("/bulk-mutations") &&
        response.request().method() === "POST",
    );
  let accepted = response();
  await form.evaluate((element) => {
    const delivery = new Event("submit", { bubbles: true, cancelable: true });
    element.dispatchEvent(delivery);
    element.dispatchEvent(delivery);
  });
  expect((await accepted).ok()).toBe(true);
  await settle(page);
  expect(batches).toHaveLength(1);
  accepted = response();
  await tagInput(page).press("Enter");
  expect((await accepted).ok()).toBe(true);
  expect(batches).toHaveLength(2);
  expect(batches[0]?.client_txn_id).not.toBe(batches[1]?.client_txn_id);
  await scrollGridCellIntoView({
    page,
    surface: timelineViewSchemaId,
    recordId: first.record_id,
    cellKey: summary,
  });
  await editTimelineSummary(
    page,
    first.record_id,
    "Ordinary scalar after tagging",
    { outcome: "accepted" },
  );
  await expect(checkbox(page, first.record_id)).toBeChecked();
  await expect(checkbox(page, second.record_id)).toBeChecked();
  expect(batches).toHaveLength(2);
  const path = `/api/v1/incidents/${incident}/views/${timelineViewSchemaId}/bulk-mutations`;
  const held = await holdBrowserRequest(page, { method: "POST", path });
  try {
    await tagInput(page).fill("partially accepted tag");
    const partialResponse = response();
    await tagInput(page).press("Enter");
    await held.waitForHit;
    const current = (
      await queryViewRows(page, incident, timelineViewSchemaId)
    ).find((row) => row.record_id === first.record_id);
    if (!current) throw new Error("Missing partial conflict target");
    await patchRecord(page, first.record_id, {
      view_schema_id: timelineViewSchemaId,
      base_row_version: current.row_version,
      client_txn_id: uniqueTxn("bulk-partial-conflict"),
      changes: [
        {
          field_key: "timeline.tags",
          action_payload: {
            kind: "collection_actions_v1",
            actions: [{ op: "add_tag", tag_name: "concurrent server tag" }],
          },
        },
      ],
    });
    held.release();
    const partial = await (await partialResponse).json();
    expect(
      partial.data.rows.map((row: { record_id: string }) => row.record_id),
    ).toEqual([second.record_id]);
    expect(partial.data.conflicts).toHaveLength(1);
    expect(partial.data.conflicts[0]).toMatchObject({
      record_id: first.record_id,
      conflict_resolution_class: "collection_review",
    });
    await expect(form).toContainText(
      "A selected record has an edit that needs recovery.",
    );
    await expect(
      form.getByRole("button", { name: "Assign tag", exact: true }),
    ).toBeDisabled();
    await expect(checkbox(page, first.record_id)).toBeChecked();
    await expect(checkbox(page, second.record_id)).toBeChecked();
    await expect(tagInput(page)).toHaveValue("partially accepted tag");
    expect(batches).toHaveLength(3);
  } finally {
    await held.dispose();
  }
});

test("Timeline tag assignment replays the captured attempt and recovers accepted reads without resending", async ({
  page,
}) => {
  const { incident, first, second } = await seed(page);
  const path = `/api/v1/incidents/${incident}/views/${timelineViewSchemaId}/bulk-mutations`;
  const queryPath = `/api/v1/incidents/${incident}/views/${timelineViewSchemaId}/query`;
  const attempts: string[] = [],
    changes: string[] = [];
  const originalHistory = new Map(
    await Promise.all(
      [first, second].map(
        async (row) =>
          [
            row.record_id,
            await fetchRecordHistoryCount(page, row.record_id),
          ] as const,
      ),
    ),
  );
  let failReads = false;
  await page.route(`**${queryPath}`, async (route) => {
    if (failReads)
      await route.fulfill({
        status: 503,
        json: {
          error: { code: "internal_error", message: "Temporary read failure" },
        },
      });
    else await route.continue();
  });
  await page.route(`**${path}`, async (route) => {
    attempts.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.status()).toBe(200);
    changes.push((await response.json()).data.change_set_id);
    if (attempts.length === 1) await route.abort("connectionfailed");
    else {
      failReads = true;
      await route.fulfill({ response });
    }
  });
  try {
    await checkbox(page, first.record_id).check();
    await checkbox(page, second.record_id).check();
    await tagInput(page).fill("captured replay tag");
    await tagInput(page).press("Enter");
    await expect(
      page.getByRole("form", { name: "Timeline bulk record actions" }),
    ).toContainText("outcome is uncertain");
    const acceptedHistory = new Map(
      await Promise.all(
        [first, second].map(
          async (row) =>
            [
              row.record_id,
              await fetchRecordHistoryCount(page, row.record_id),
            ] as const,
        ),
      ),
    );
    for (const row of [first, second])
      expect(acceptedHistory.get(row.record_id)).toBeGreaterThan(
        originalHistory.get(row.record_id) ?? 0,
      );
    await checkbox(page, second.record_id).uncheck();
    await tagInput(page).fill("newer raw tag remains");
    await openRecoveryItem(page, /^Tag assignment ·/);
    await page
      .getByRole("button", { name: "Retry tag assignment", exact: true })
      .click();
    const refresh = page.getByRole("button", {
      name: "Retry refresh",
      exact: true,
    });
    await expect(refresh).toBeVisible();
    await expect(tagInput(page)).toHaveValue("newer raw tag remains");
    await expect(checkbox(page, first.record_id)).toBeChecked();
    await expect(checkbox(page, second.record_id)).not.toBeChecked();
    failReads = false;
    await refresh.click();
    await expect(refresh).toHaveCount(0);
    expect(attempts).toEqual([attempts[0], attempts[0]]);
    expect(changes).toEqual([changes[0], changes[0]]);
    for (const row of [first, second]) {
      expect(await fetchRecordHistoryCount(page, row.record_id)).toBe(
        acceptedHistory.get(row.record_id),
      );
      expect(
        JSON.stringify(
          (await queryViewRows(page, incident, timelineViewSchemaId)).find(
            (value) => value.record_id === row.record_id,
          )?.cells["timeline.tags"],
        ),
      ).toContain("captured replay tag");
    }
  } finally {
    failReads = false;
    await page.unroute(`**${path}`);
    await page.unroute(`**${queryPath}`);
  }
});

test("Timeline bulk tag controls fit density Inspector zoom and text spacing without losing authoring", async ({
  page,
  workerAdmin,
}, testInfo) => {
  test.setTimeout(240_000);
  const preferences = await installVisualPreferences(page, workerAdmin.user_id);
  const raw = "long raw Ω 東京 tag ".repeat(32);
  for (const density of ["compact", "default", "comfortable"] as const) {
    preferences.select(density);
    const { first, second } = await seed(page);
    await checkbox(page, first.record_id).check();
    await checkbox(page, second.record_id).check();
    await tagInput(page).fill(raw);
    for (const profile of [
      { name: "base", width: 1440, height: 900, inspector: false, zoom: 1 },
      {
        name: "minimum-inspector",
        width: 1440,
        height: 900,
        inspector: true,
        zoom: 1,
        resize: "Home",
      },
      {
        name: "maximum-inspector",
        width: 1440,
        height: 900,
        inspector: true,
        zoom: 1,
        resize: "End",
      },
      { name: "narrow", width: 1024, height: 720, inspector: true, zoom: 1 },
      { name: "compact", width: 768, height: 640, inspector: true, zoom: 1 },
      {
        name: "text-spacing",
        width: 1024,
        height: 720,
        inspector: false,
        zoom: 1,
      },
      {
        name: "supported-zoom",
        width: 2048,
        height: 1440,
        inspector: false,
        zoom: 2,
      },
      { name: "zoom", width: 1440, height: 900, inspector: false, zoom: 2 },
    ]) {
      const close = page.getByTestId(
        workbookInspectorCloseButtonTestId(timelineViewSchemaId),
      );
      if (await close.count()) await close.click();
      await page.setViewportSize({
        width: profile.width,
        height: profile.height,
      });
      await page.evaluate((profile) => {
        document.documentElement.style.zoom = String(profile.zoom);
        document.body.style.letterSpacing =
          profile.name === "text-spacing" ? "0.12em" : "";
        document.body.style.wordSpacing =
          profile.name === "text-spacing" ? "0.16em" : "";
      }, profile);
      if (profile.inspector) await openTimelineInspector(page, first.record_id);
      if (profile.resize)
        await page
          .getByRole("separator", { name: "Resize inspector" })
          .press(profile.resize);
      await tagInput(page).focus();
      await tagInput(page).evaluate((input: HTMLInputElement) =>
        input.setSelectionRange(3, 8, "backward"),
      );
      await settle(page);
      await expect(tagInput(page)).toHaveValue(raw);
      await expect(tagInput(page)).toBeFocused();
      await expect(checkbox(page, first.record_id)).toBeChecked();
      await expect(checkbox(page, second.record_id)).toBeChecked();
      const form = page.getByRole("form", {
        name: "Timeline bulk record actions",
      });
      for (const name of ["Assign tag", "Clear tag draft"])
        await form
          .getByRole("button", { name, exact: true })
          .click({ trial: true });
      const query = page.getByRole("region", {
        name: "Workbook query and action controls",
      });
      if (profile.name !== "zoom") {
        await expect(
          query.getByRole("button", { name: "Saved view", exact: true }),
        ).toBeInViewport();
        await expect(
          query.getByRole("button", {
            name: "Filters, 0 active filters",
            exact: true,
          }),
        ).toBeInViewport();
        await expect(
          query.getByRole("button", { name: "Columns", exact: true }),
        ).toBeInViewport();
      }
      for (const name of ["Find in loaded rows", "Clear contents", "Add row"])
        await expect(
          query.getByRole("button", { name, exact: true }),
        ).toBeInViewport();
      const rect = await page
        .getByTestId(gridShellTestId(timelineViewSchemaId))
        .boundingBox();
      expect(rect?.height).toBeGreaterThan(200);
      expect(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= window.innerWidth,
        ),
      ).toBe(true);
      await testInfo.attach(`bulk-tag-${density}-${profile.name}`, {
        body: await page.screenshot(),
        contentType: "image/png",
      });
    }
    await page.evaluate(() => {
      document.documentElement.style.zoom = "1";
      document.body.style.letterSpacing = "";
      document.body.style.wordSpacing = "";
    });
  }
});
