import { Buffer } from "node:buffer";
import type { ViewRow } from "@cartulary/protocol-ts/http";
import {
  applyFilterChip,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  draftTimelineCollectionInputTestId,
  gridRowTestId,
  relationshipItemsTestId,
  rowCellTestId,
  timelineCollectionInputTestId,
  timelineEvidenceFileInputTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";

import { csrfHeaders } from "./support/auth/browserSession";
import { openTimelineSurface } from "./support/evidence/browser";
import {
  chooseEvidenceFile,
  openTimelineAttachmentFeedback,
} from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import { responseBarrier } from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  observePageRequests,
  rejectUnexpectedPageErrors,
} from "./support/runtime/observations";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import {
  awaitResponseBarrier,
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import { showTimelineCollectionColumns } from "./support/workbook/collections";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

async function fileTargetFixture(page: Page, extraRows = 0) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("FILE-TARGET"),
    "File gesture source identity",
  );
  const records: ViewRow[] = [];
  for (const [index, label] of [
    "A active",
    "B target",
    "C Inspector",
    ...Array.from({ length: extraRows }, (_, i) => `Virtual row ${i}`),
  ].entries())
    records.push(
      await createViewRow(page, incident, timelineViewSchemaId, {
        client_txn_id: uniqueTxn("file-target-source"),
        "timeline.activity_utc_text": new Date(
          Date.UTC(2026, 8, 23, 0, index),
        ).toISOString(),
        "timeline.activity_synopsis_text": label,
      }),
    );
  const [a, b, c] = records;
  const last = records.at(-1);
  if (!a || !b || !c || !last)
    throw new Error("Expected three distinct file sources");
  await openTimelineSurface(page, incident);
  await openTimelineInspector(page, c.record_id);
  const aCell = page.getByRole("gridcell").filter({
    has: page.getByTestId(
      rowCellTestId(a.record_id, "timeline.activity_synopsis_text"),
    ),
  });
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: rowCellTestId(a.record_id, "timeline.activity_synopsis_text"),
  });
  await aCell.click({ modifiers: ["Shift"], position: { x: 3, y: 3 } });
  await expect(aCell).toBeFocused();
  await expect(
    page.getByTestId(timelineEvidenceFileInputTestId(c.record_id)),
  ).toBeAttached();
  return { incident, a, b, c, last };
}

function observeFileWrites(page: Page) {
  const writes = {
    slots: 0,
    transfers: 0,
    evidence: 0,
    timeline: 0,
    links: [] as string[],
  };
  const watch = (request: import("@playwright/test").Request) => {
    if (request.method() === "POST" && request.url().endsWith("/object-blobs"))
      writes.slots++;
    if (
      request.method() === "PUT" &&
      request.url().includes("/object-uploads/")
    )
      writes.transfers++;
    if (
      request.method() === "POST" &&
      request.url().endsWith(`/views/${evidenceViewSchemaId}/rows`)
    )
      writes.evidence++;
    if (
      request.method() === "POST" &&
      request.url().endsWith(`/views/${timelineViewSchemaId}/rows`)
    )
      writes.timeline++;
    if (request.method() === "PATCH" && request.url().includes("/records/"))
      writes.links.push(request.postData() ?? "");
  };
  const stopObserving = observePageRequests(page, watch);
  return { writes, stop: stopObserving };
}

async function fileGesture(
  target: import("@playwright/test").Locator,
  kind: "drop" | "paste",
  name: string,
  count = 1,
) {
  await target.evaluate(
    (element, input) => {
      const data = new DataTransfer();
      for (let i = 0; i < input.count; i++)
        data.items.add(
          new File(["capture"], `${input.name}-${i}.txt`, {
            type: "text/plain",
          }),
        );
      data.setData("text/plain", "File content must take precedence");
      element.dispatchEvent(
        input.kind === "drop"
          ? new DragEvent("drop", {
              bubbles: true,
              cancelable: true,
              dataTransfer: data,
            })
          : new ClipboardEvent("paste", {
              bubbles: true,
              cancelable: true,
              clipboardData: data,
            }),
      );
    },
    { kind, name, count },
  );
}

async function attachedTimelineIds(page: Page, incident: string) {
  return (await queryViewRows(page, incident, timelineViewSchemaId))
    .filter((row) => Number(row.cells["timeline.evidence_count"]?.value) > 0)
    .map((row) => row.record_id)
    .sort();
}

test.beforeEach(({ page }) => {
  rejectUnexpectedPageErrors(page);
});

async function exerciseRowLocalGesture(
  page: Page,
  info: TestInfo,
  entry: "drop" | "scalar-paste" | "collection-paste",
) {
  test.setTimeout(120_000);

  const f = await fileTargetFixture(page);
  if (entry === "collection-paste") {
    await showTimelineCollectionColumns(page, ["Tags"]);
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: relationshipItemsTestId(
        f.a.record_id,
        "timeline.tags",
        "grid",
      ),
    });
    const active = page.getByRole("gridcell").filter({
      has: page.getByTestId(
        relationshipItemsTestId(f.a.record_id, "timeline.tags", "grid"),
      ),
    });
    await active.click({ modifiers: ["Shift"], position: { x: 3, y: 3 } });
    await expect(active).toBeFocused();
  }
  const observed = observeFileWrites(page);
  await fileGesture(
    page.getByTestId(
      entry === "collection-paste"
        ? relationshipItemsTestId(f.b.record_id, "timeline.tags", "grid")
        : rowCellTestId(f.b.record_id, "timeline.activity_synopsis_text"),
    ),
    entry === "drop" ? "drop" : "paste",
    `row-local-${entry}`,
  );
  await expect
    .poll(() => attachedTimelineIds(page, f.incident))
    .toEqual([f.b.record_id]);
  await info.attach(`association-identities-${entry}`, {
    body: JSON.stringify({
      associated: await attachedTimelineIds(page, f.incident),
      writes: observed.writes,
    }),
    contentType: "application/json",
  });
  expect(observed.writes).toMatchObject({
    slots: 1,
    transfers: 1,
    evidence: 1,
    timeline: 0,
  });
  expect(observed.writes.links).toHaveLength(1);
  await expect(
    page.getByTestId(timelineEvidenceFileInputTestId(f.c.record_id)),
  ).toBeAttached();
  observed.stop();
}

test("Timeline row-local drop targets B independently of active A and Inspector C", async ({
  page,
}, info) => {
  await exerciseRowLocalGesture(page, info, "drop");
});

test("Timeline row-local scalar-paste targets B independently of active A and Inspector C", async ({
  page,
}, info) => {
  await exerciseRowLocalGesture(page, info, "scalar-paste");
});

test("Timeline row-local collection-paste targets B independently of active A and Inspector C", async ({
  page,
}, info) => {
  await exerciseRowLocalGesture(page, info, "collection-paste");
});

async function exerciseDraftGesture(
  page: Page,
  info: TestInfo,
  field: "scalar" | "collection",
  kind: "drop" | "paste",
) {
  test.setTimeout(120_000);

  const f = await fileTargetFixture(page);
  if (field === "collection")
    await showTimelineCollectionColumns(page, ["Tags"]);
  const observed = observeFileWrites(page);
  const id =
    field === "scalar"
      ? draftCellTestId("timeline.activity_synopsis_text")
      : draftTimelineCollectionInputTestId("timeline.tags");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: id,
  });
  await fileGesture(page.getByTestId(id), kind, `${field}-${kind}`);
  await expect
    .poll(() => attachedTimelineIds(page, f.incident))
    .toHaveLength(1);
  const rows = await queryViewRows(page, f.incident, timelineViewSchemaId);
  const newRows = rows.filter(
    (row) =>
      ![f.a.record_id, f.b.record_id, f.c.record_id].includes(row.record_id),
  );
  await info.attach(`${field}-${kind}-identities`, {
    body: JSON.stringify({
      associated: await attachedTimelineIds(page, f.incident),
      newRows: newRows.map((row) => row.record_id),
      writes: observed.writes,
    }),
    contentType: "application/json",
  });
  expect.soft(newRows).toHaveLength(1);
  expect
    .soft(await attachedTimelineIds(page, f.incident))
    .toEqual(newRows.map((row) => row.record_id));
  expect
    .soft(observed.writes)
    .toMatchObject({ slots: 1, transfers: 1, evidence: 1, timeline: 1 });
  await expect(
    page.getByTestId(timelineEvidenceFileInputTestId(f.c.record_id)),
  ).toBeAttached();
  observed.stop();
}

test("Timeline draft scalar drop retains its source instead of Inspector context", async ({
  page,
}, info) => {
  await exerciseDraftGesture(page, info, "scalar", "drop");
});

test("Timeline draft scalar paste retains its source instead of Inspector context", async ({
  page,
}, info) => {
  await exerciseDraftGesture(page, info, "scalar", "paste");
});

test("Timeline draft collection drop retains its source instead of Inspector context", async ({
  page,
}, info) => {
  await exerciseDraftGesture(page, info, "collection", "drop");
});

test("Timeline draft collection paste retains its source instead of Inspector context", async ({
  page,
}, info) => {
  await exerciseDraftGesture(page, info, "collection", "paste");
});

test("Timeline background file gestures require an eligible active grid target", async ({
  page,
}) => {
  const f = await fileTargetFixture(page);
  const observed = observeFileWrites(page);
  const area = page.getByRole("region", {
    name: "Timeline file work area",
    exact: true,
  });
  await fileGesture(area, "drop", "background-active");
  await expect
    .poll(() => attachedTimelineIds(page, f.incident))
    .toEqual([f.a.record_id]);
  expect(observed.writes).toMatchObject({
    slots: 1,
    transfers: 1,
    evidence: 1,
    timeline: 0,
  });
  await fileGesture(
    page.getByRole("columnheader").first(),
    "drop",
    "not-a-source",
  );
  await expect(
    page
      .getByRole("region", { name: "Timeline attachments", exact: true })
      .getByText("Select the Timeline row or draft for this file.", {
        exact: true,
      }),
  ).toBeVisible();
  await fileGesture(
    page.getByTestId(
      rowCellTestId(f.b.record_id, "timeline.activity_synopsis_text"),
    ),
    "paste",
    "multiple-on-row",
    2,
  );
  await expect(
    page
      .getByRole("region", { name: "Timeline attachments", exact: true })
      .getByText("Choose one file at a time.", { exact: true }),
  ).toBeVisible();
  expect(observed.writes.slots).toBe(1);
  observed.stop();
  const empty = await createIncident(
    page,
    uniqueIncidentKey("FILE-NO-TARGET"),
    "No active file source",
  );
  await openTimelineSurface(page, empty);
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftCellTestId("timeline.activity_synopsis_text"),
  });
  const absent = observeFileWrites(page);
  await fileGesture(area, "paste", "background-missing");
  await expect(
    page
      .getByRole("region", { name: "Timeline attachments", exact: true })
      .getByText("Select the Timeline row or draft for this file.", {
        exact: true,
      }),
  ).toBeVisible();
  await fileGesture(area, "drop", "multiple", 2);
  await expect(
    page
      .getByRole("region", { name: "Timeline attachments", exact: true })
      .getByText("Choose one file at a time.", { exact: true }),
  ).toBeVisible();
  expect(absent.writes).toEqual({
    slots: 0,
    transfers: 0,
    evidence: 0,
    timeline: 0,
    links: [],
  });
  await page
    .getByTestId(draftCellTestId("timeline.activity_synopsis_text"))
    .focus();
  await fileGesture(area, "paste", "background-active-draft");
  await expect.poll(() => attachedTimelineIds(page, empty)).toHaveLength(1);
  expect(absent.writes).toMatchObject({
    slots: 1,
    transfers: 1,
    evidence: 1,
    timeline: 1,
  });
  expect(absent.writes.links).toHaveLength(0);
  absent.stop();
});

test("Timeline picker completion keeps the invoking source after Inspector replacement", async ({
  page,
}, info) => {
  const f = await fileTargetFixture(page);
  const observed = observeFileWrites(page);
  const chooserPromise = page.waitForEvent("filechooser");
  await page
    .getByRole("button", {
      name: "Attach file to this Timeline record",
      exact: true,
    })
    .click();
  const chooser = await chooserPromise;
  await openTimelineInspector(page, f.b.record_id);
  await applyFilterChip(
    page,
    timelineViewSchemaId,
    "timeline.capture_state",
    "reviewed",
  );
  await expect(
    page.getByTestId(gridRowTestId(timelineViewSchemaId, f.c.record_id)),
  ).toHaveCount(0);
  await chooser.setFiles({
    name: "invoked-C.txt",
    mimeType: "text/plain",
    buffer: Buffer.from("capture"),
  });
  await expect
    .poll(() => attachedTimelineIds(page, f.incident))
    .toHaveLength(1);
  await info.attach("picker-associations", {
    body: JSON.stringify({
      associated: await attachedTimelineIds(page, f.incident),
      writes: observed.writes,
    }),
    contentType: "application/json",
  });
  expect(await attachedTimelineIds(page, f.incident)).toEqual([f.c.record_id]);
  expect(observed.writes).toMatchObject({
    slots: 1,
    transfers: 1,
    evidence: 1,
    timeline: 0,
  });
  observed.stop();
});

async function exerciseNativePaste(
  page: Page,
  info: TestInfo,
  kind: "scalar" | "collection",
) {
  test.setTimeout(120_000);
  await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);

  const f = await fileTargetFixture(page);
  if (kind === "collection")
    await showTimelineCollectionColumns(page, ["Tags"]);
  const target =
    kind === "scalar"
      ? rowCellTestId(f.b.record_id, "timeline.activity_synopsis_text")
      : relationshipItemsTestId(f.b.record_id, "timeline.tags", "grid");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: target,
  });
  if (kind === "scalar") await page.getByTestId(target).click();
  else
    await page
      .getByRole("group", { name: "Tags collection cell", exact: true })
      .filter({ has: page.getByTestId(target) })
      .getByRole("button", { name: "Add tags token", exact: true })
      .click();
  const input = page.getByTestId(
    kind === "scalar"
      ? timelineScalarEditorTestId({
          recordId: f.b.record_id,
          fieldKey: "timeline.activity_synopsis_text",
          surface: "grid",
        })
      : timelineCollectionInputTestId(f.b.record_id, "timeline.tags", "grid"),
  );
  await input.fill("native raw Ω authoring");
  await input.evaluate((element: HTMLInputElement) =>
    element.setSelectionRange(7, 10, "backward"),
  );
  await page.evaluate(() => navigator.clipboard.writeText("text"));
  await page.keyboard.press("Control+v");
  await expect(input).toHaveValue("native text Ω authoring");
  if (kind === "scalar")
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, f.incident, timelineViewSchemaId)).find(
            (row) => row.record_id === f.b.record_id,
          )?.cells["timeline.activity_synopsis_text"]?.value,
      )
      .toBe("native text Ω authoring");
  await input.evaluate((element: HTMLInputElement) =>
    element.setSelectionRange(3, 6, "backward"),
  );
  const cdp = await page.context().newCDPSession(page);
  await cdp.send("Input.imeSetComposition", {
    text: "仮",
    selectionStart: 1,
    selectionEnd: 1,
  });
  const state = (element: HTMLInputElement) => ({
    value: element.value,
    start: element.selectionStart,
    end: element.selectionEnd,
    direction: element.selectionDirection,
    focused: document.activeElement === element,
  });
  const before = await input.evaluate(state);
  const observed = observeFileWrites(page);
  let release = () => {};
  const gateBarrier = responseBarrier(page);
  release = gateBarrier.release;
  const gate = gateBarrier.promise;
  const path = "**/api/v1/object-uploads/**";
  await installOwnedRoute(page, path, async (route) => {
    if (!(await awaitResponseBarrier(route, gate))) return;
    await route.continue();
  });
  try {
    await chooseEvidenceFile(
      page.getByTestId(timelineEvidenceFileInputTestId(f.b.record_id)),
      [],
    );
    expect(await input.evaluate(state)).toEqual(before);
    expect(observed.writes).toEqual({
      slots: 0,
      transfers: 0,
      evidence: 0,
      timeline: 0,
      links: [],
    });
    await fileGesture(input, "paste", `native-${kind}`);
    await expect.poll(() => observed.writes.transfers).toBe(1);
    expect(await input.evaluate(state)).toEqual(before);
    expect(observed.writes).toMatchObject({
      slots: 1,
      transfers: 1,
      evidence: 0,
      timeline: 0,
      links: [],
    });
    await cdp.send("Input.imeSetComposition", {
      text: "仮名",
      selectionStart: 2,
      selectionEnd: 2,
    });
    await expect(input).toHaveValue(/仮名/);
    await cdp.send("Input.imeSetComposition", {
      text: "",
      selectionStart: 0,
      selectionEnd: 0,
    });
    await input.press("Escape");
    release();
    await openTimelineAttachmentFeedback(page);
    const recovery = page.getByRole("group", {
      includeHidden: true,
      name: `File recovery: native-${kind}-0.txt`,
      exact: true,
    });
    await expect(recovery).toContainText(
      /Evidence attached\.|Review the original Timeline row/,
    );
    // Existing source-edit/presentation review can win the native editing race.
    const review = recovery.getByRole("button", {
      name: "Review original source",
      exact: true,
    });
    if (await review.isVisible()) {
      await review.click();
      await recovery
        .getByRole("button", { name: "Use reviewed source", exact: true })
        .click();
      await recovery
        .getByRole("button", { name: "Resume", exact: true })
        .click();
    }
    await expect
      .poll(() => attachedTimelineIds(page, f.incident))
      .toEqual([f.b.record_id]);
    expect(observed.writes).toMatchObject({
      slots: 1,
      transfers: 1,
      evidence: 1,
      timeline: 0,
    });
    expect(observed.writes.links).toHaveLength(1);
    await info.attach(`native-${kind}`, {
      body: JSON.stringify({ before, writes: observed.writes }),
      contentType: "application/json",
    });

    release();

    await removeOwnedRoute(page, path);
  } finally {
    gateBarrier.cancel();
    observed.stop();
    await cdp.detach();
  }
}

test("Timeline file paste preserves native scalar editing during upload", async ({
  page,
}, info) => {
  await exerciseNativePaste(page, info, "scalar");
});

test("Timeline file paste preserves native collection editing during upload", async ({
  page,
}, info) => {
  await exerciseNativePaste(page, info, "collection");
});

test("Timeline admitted file work keeps its source through selection filtering and detachment", async ({
  page,
}, info) => {
  test.setTimeout(120_000);
  const f = await fileTargetFixture(page, 60);
  const observed = observeFileWrites(page);
  let release = () => {};
  const gateBarrier = responseBarrier(page);
  release = gateBarrier.release;
  const gate = gateBarrier.promise;
  const path = "**/api/v1/object-uploads/**";
  await installOwnedRoute(page, path, async (route) => {
    if (!(await awaitResponseBarrier(route, gate))) return;
    await route.continue();
  });
  try {
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(
        f.b.record_id,
        "timeline.activity_synopsis_text",
      ),
    });
    await fileGesture(
      page.getByTestId(
        rowCellTestId(f.b.record_id, "timeline.activity_synopsis_text"),
      ),
      "drop",
      "filtered-source",
    );
    await expect.poll(() => observed.writes.transfers).toBe(1);
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(
        f.last.record_id,
        "timeline.activity_synopsis_text",
      ),
    });
    await expect(
      page.getByTestId(gridRowTestId(timelineViewSchemaId, f.b.record_id)),
    ).toHaveCount(0);
    await openTimelineInspector(page, f.a.record_id);
    await applyFilterChip(
      page,
      timelineViewSchemaId,
      "timeline.capture_state",
      "reviewed",
    );
    await expect(
      page.getByTestId(
        rowCellTestId(f.b.record_id, "timeline.activity_synopsis_text"),
      ),
    ).toHaveCount(0);
    // A disappeared active source is not replaced by Inspector selection or the draft.
    await fileGesture(
      page.getByRole("region", {
        name: "Timeline file work area",
        exact: true,
      }),
      "paste",
      "stale-background",
    );
    await expect(
      page
        .getByRole("region", { name: "Timeline attachments", exact: true })
        .getByText("Select the Timeline row or draft for this file.", {
          exact: true,
        }),
    ).toBeVisible();
    expect(observed.writes.slots).toBe(1);
    const focused = await page.evaluateHandle(() => document.activeElement);
    release();
    const recovery = page.getByRole("group", {
      name: "File recovery: filtered-source-0.txt",
      exact: true,
    });
    await expect(
      page.getByRole("button", { name: /^Attachments:/ }),
    ).toContainText("1 need attention");
    expect(
      await focused.evaluate((element) => element === document.activeElement),
    ).toBe(true);
    expect(await attachedTimelineIds(page, f.incident)).toEqual([]);
    expect(observed.writes.links).toEqual([]);
    await openTimelineAttachmentFeedback(page);
    await expect(recovery).toContainText(
      "Evidence saved. Review the original Timeline row before linking.",
    );
    await recovery
      .getByRole("button", { name: "Review original source", exact: true })
      .click();
    await expect(recovery).toContainText("B target");
    await recovery
      .getByRole("button", { name: "Use reviewed source", exact: true })
      .click();
    await recovery.getByRole("button", { name: "Resume", exact: true }).click();
    await expect
      .poll(() => attachedTimelineIds(page, f.incident))
      .toEqual([f.b.record_id]);
    expect(observed.writes).toMatchObject({
      slots: 1,
      transfers: 1,
      evidence: 1,
      timeline: 0,
    });
    expect(observed.writes.links).toHaveLength(1);
    await info.attach("filtered-original-source", {
      body: JSON.stringify({
        associated: await attachedTimelineIds(page, f.incident),
        writes: observed.writes,
      }),
      contentType: "application/json",
    });

    release();

    await removeOwnedRoute(page, path);
  } finally {
    gateBarrier.cancel();
    observed.stop();
  }
});

test("Timeline keyboard picker cancellation and removed originals start no upload", async ({
  page,
}, info) => {
  const f = await fileTargetFixture(page);
  const observed = observeFileWrites(page);
  const input = page.getByTestId(
    timelineEvidenceFileInputTestId(f.c.record_id),
  );
  const button = page.getByRole("button", {
    name: (await input.getAttribute("aria-label")) ?? "",
    exact: true,
  });
  await button.focus();
  const cancelled = page.waitForEvent("filechooser");
  await button.press("Enter");
  await (await cancelled).setFiles([]);
  await expect(button).toBeFocused();
  expect(observed.writes).toEqual({
    slots: 0,
    transfers: 0,
    evidence: 0,
    timeline: 0,
    links: [],
  });
  const removedChooser = page.waitForEvent("filechooser");
  await button.press("Space");
  const chooser = await removedChooser;
  const removed = await publicHttpOperation({
    operationID: "deleteRecord",
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    pathParameters: { record_id: f.c.record_id },
    body: {
      base_row_version: f.c.row_version,
      client_txn_id: uniqueTxn("file-remove-original"),
      reason: "File chooser original removal fixture",
    },
  });
  expect(removed.ok).toBe(true);
  await openTimelineInspector(page, f.b.record_id);
  await chooser.setFiles({
    name: "removed-original.txt",
    mimeType: "text/plain",
    buffer: Buffer.from("capture"),
  });
  await openTimelineAttachmentFeedback(page);
  await expect(
    page
      .getByRole("region", { name: "Timeline attachments", exact: true })
      .getByText(
        "The original Timeline source is unavailable. No upload started.",
        { exact: true },
      ),
  ).toBeVisible();
  expect(await attachedTimelineIds(page, f.incident)).toEqual([]);
  expect(observed.writes).toEqual({
    slots: 0,
    transfers: 0,
    evidence: 0,
    timeline: 0,
    links: [],
  });
  await info.attach("cancelled-and-removed", {
    body: JSON.stringify(observed.writes),
    contentType: "application/json",
  });
  observed.stop();
});

test("Timeline picker follows its invoking draft promotion and accepts a zero-byte file", async ({
  page,
}, info) => {
  const f = await fileTargetFixture(page);
  // Keep the invoking editor beside the row action. Horizontal virtualization
  // may otherwise unmount it when the native picker button scrolls into view.
  await showTimelineCollectionColumns(page, ["Activity Synopsis"]);
  const draft = page.getByTestId(
    draftCellTestId("timeline.activity_synopsis_text"),
  );
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftCellTestId("timeline.activity_synopsis_text"),
  });
  await draft.focus();
  const observed = observeFileWrites(page);
  const chooserPromise = page.waitForEvent("filechooser");
  await page
    .getByRole("button", {
      name: "Attach evidence to draft timeline row",
      exact: true,
    })
    .click();
  const chooser = await chooserPromise;
  await expect(draft).toBeFocused();
  expect(observed.writes.timeline).toBe(0);
  await draft.fill("Promoted while the chooser is open");
  const originalIds = [f.a.record_id, f.b.record_id, f.c.record_id];
  const created = async () =>
    (await queryViewRows(page, f.incident, timelineViewSchemaId)).filter(
      (row) => !originalIds.includes(row.record_id),
    );
  await expect.poll(async () => (await created()).length).toBe(1);
  const promoted = (await created())[0];
  if (!promoted) throw new Error("Expected the invoking draft to be promoted");
  await chooser.setFiles({
    name: "zero-byte.txt",
    mimeType: "text/plain",
    buffer: Buffer.alloc(0),
  });
  await expect
    .poll(() => attachedTimelineIds(page, f.incident))
    .toEqual([promoted.record_id]);
  expect(await created()).toHaveLength(1);
  expect(observed.writes).toMatchObject({
    slots: 1,
    transfers: 1,
    evidence: 1,
    timeline: 1,
  });
  expect(observed.writes.links).toHaveLength(1);
  expect(
    (await created())[0]?.cells["timeline.activity_synopsis_text"]?.value,
  ).toBe("Promoted while the chooser is open");
  await info.attach("invoking-draft-promotion", {
    body: JSON.stringify({
      originalIds,
      promoted: promoted.record_id,
      writes: observed.writes,
    }),
    contentType: "application/json",
  });
  observed.stop();
});
