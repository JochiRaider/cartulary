import {
  scrollGridCellIntoView,
  sortByHeader,
} from "@cartulary/test-utils/grid";
import {
  authTestId,
  genericEditFieldSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridShellTestId,
  incidentLandingTestId,
  rowCellTestId,
  timelineScalarEditorTestId,
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
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
import { revokeAllSessions } from "./support/auth/sessions";
import { csrfHeaderName } from "./support/auth/storageState";
import {
  confirmLifecycle,
  currentLifecycle,
  lifecycleAction,
  openLifecycle,
} from "./support/incidentLifecycle";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { fetchRecordHistoryCount } from "./support/workbook/history";
import { switchOrdinarySheet } from "./support/workbook/ordinaryCreate";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import {
  createSavedView,
  selectSavedView,
} from "./support/workbook/savedViews";

function gate() {
  let release!: () => void;
  const promise = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { promise, release };
}

async function fixture(page: Page, view: string) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("GEA"),
    "Committed grid autosave evidence",
  );
  const field =
    view === hostsViewSchemaId
      ? "host.display_name"
      : view === timelineViewSchemaId
        ? "timeline.activity_synopsis_text"
        : "evidence.title";
  const other = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("gea-other"),
    [field]: "Other row",
  });
  const row = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("gea-row"),
    [field]: "Original row",
  });
  const sockets = installIncidentSocketMonitor(page, incident);
  await page.goto(`/?incident_id=${incident}&view_schema_id=${view}`);
  await sockets.waitForAcceptedSocket();
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
  return { incident, view, field, row, other, sockets };
}

const editor = (page: Page, view: string, record: string, field: string) =>
  page.getByTestId(
    view === timelineViewSchemaId
      ? timelineScalarEditorTestId({
          recordId: record,
          fieldKey: field,
          surface: "grid",
        })
      : `grid-editor-${record}-${field}`,
  );

async function activate(
  page: Page,
  view: string,
  record: string,
  field: string,
) {
  const contract = requireViewContract(view).fieldMap[field];
  if (contract?.defaultHidden) {
    await page.getByTestId(workbookColumnsMenuTriggerTestId(view)).click();
    const option = page
      .getByTestId(workbookColumnsMenuTestId(view))
      .getByRole("menuitemcheckbox", { name: contract.label, exact: true });
    if ((await option.getAttribute("aria-checked")) !== "true")
      await option.click();
    await page.getByTestId(workbookColumnsMenuTriggerTestId(view)).click();
  }
  await scrollGridCellIntoView({
    page,
    surface: view,
    recordId: record,
    cellKey: field,
  });
  await page.getByTestId(rowCellTestId(record, field)).click();
  const input = editor(page, view, record, field);
  await expect(input).toBeFocused();
  return input;
}

test("Committed grid A acknowledgement preserves B authoring and sequences real same-record writes", async ({
  page,
}) => {
  for (const view of [
    hostsViewSchemaId,
    evidenceViewSchemaId,
    timelineViewSchemaId,
  ]) {
    const f = await fixture(page, view),
      a = gate(),
      b = gate(),
      bodies: string[] = [];
    if (view === timelineViewSchemaId) {
      // Empty activity timestamps tie on opaque record IDs. Capture an explicit
      // visible order for this authoritative-acceptance navigation assertion.
      await sortByHeader(page, view, "timeline.activity_synopsis_text");
    }
    const receipts: Array<{
      data: { change_set_id: string; row: { row_version: number } };
    }> = [];
    await page.route(`**/api/v1/records/${f.row.record_id}`, async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      const index = bodies.push(route.request().postData() ?? "") - 1;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      receipts[index] = await response.json();
      await (index === 0 ? a.promise : b.promise);
      await route.fulfill({ response });
    });
    try {
      const input = await activate(page, view, f.row.record_id, f.field);
      await input.fill("First accepted value");
      await input.press("Enter");
      await expect.poll(() => receipts.length).toBe(1);
      await expect(input).toBeFocused();
      await input.fill("Newer second value");
      await input.press("Enter");
      expect(bodies).toHaveLength(1);
      a.release();
      await expect.poll(() => receipts.length).toBe(2);
      await expect(input).toHaveValue("Newer second value");
      await expect(input).toBeFocused();
      const requests = bodies.map((body) => JSON.parse(body));
      expect(requests.map((body) => body.base_row_version)).toEqual([1, 2]);
      expect(new Set(requests.map((body) => body.client_txn_id)).size).toBe(2);
      expect(requests.map((body) => body.changes)).toEqual([
        [{ field_key: f.field, value: "First accepted value" }],
        [{ field_key: f.field, value: "Newer second value" }],
      ]);
      expect(receipts.map((receipt) => receipt.data.row.row_version)).toEqual([
        2, 3,
      ]);
      expect(
        new Set(receipts.map((receipt) => receipt.data.change_set_id)).size,
      ).toBe(2);
      b.release();
      await expect(input).toHaveCount(0);
      await expect(
        page.getByTestId(rowCellTestId(f.row.record_id, f.field)),
      ).toHaveText("Newer second value");
      expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(3);
      const rows = await queryViewRows(page, f.incident, view);
      expect(
        rows.find((row) => row.record_id === f.row.record_id)?.cells[f.field]
          ?.value,
      ).toBe("Newer second value");
      expect(
        rows.find((row) => row.record_id === f.other.record_id)?.row_version,
      ).toBe(1);
      await expect(
        page
          .getByTestId(rowCellTestId(f.other.record_id, f.field))
          .locator("xpath=ancestor::*[@role='gridcell'][1]"),
      ).toBeFocused();
    } finally {
      a.release();
      b.release();
      await page.unroute(`**/api/v1/records/${f.row.record_id}`);
    }
  }
});

test("Committed grid response loss replays exact real commits while retaining detached newer drafts", async ({
  page,
}) => {
  for (const [view, malformed] of [
    [hostsViewSchemaId, false],
    [evidenceViewSchemaId, true],
    [timelineViewSchemaId, false],
  ] as const) {
    const f = await fixture(page, view),
      replay = gate(),
      bodies: string[] = [],
      changes: string[] = [];
    await page.route(`**/api/v1/records/${f.row.record_id}`, async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      const index = bodies.push(route.request().postData() ?? "") - 1;
      if (index > 0) await replay.promise;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      changes.push((await response.json()).data.change_set_id);
      if (index > 0) await route.fulfill({ response });
      else if (malformed)
        await route.fulfill({
          response,
          body: JSON.stringify({
            data: { row: {} },
            meta: { request_id: "gea-lost" },
          }),
        });
      else await route.abort("failed");
    });
    try {
      const input = await activate(page, view, f.row.record_id, f.field);
      await input.fill("One durable change");
      await input.press("Enter");
      await expect.poll(() => changes.length).toBe(1);
      await input.fill("  unfinished newer text  ");
      const away =
        view === timelineViewSchemaId
          ? hostsViewSchemaId
          : timelineViewSchemaId;
      await switchOrdinarySheet(page, away);
      replay.release();
      await expect.poll(() => changes.length).toBe(2);
      expect(bodies).toHaveLength(2);
      expect(bodies[1]).toBe(bodies[0]);
      expect(changes[1]).toBe(changes[0]);
      expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(2);
      await switchOrdinarySheet(page, view);
      await expect(editor(page, view, f.row.record_id, f.field)).toHaveCount(0);
      const restored = await activate(page, view, f.row.record_id, f.field);
      await expect(restored).toHaveValue("  unfinished newer text  ");
      await expect(restored).toBeFocused();
      expect(
        (await queryViewRows(page, f.incident, view)).find(
          (row) => row.record_id === f.row.record_id,
        )?.cells[f.field]?.value,
      ).toBe("One durable change");
      await restored.press("Escape");
      expect(bodies).toHaveLength(2);
    } finally {
      replay.release();
      await page.unroute(`**/api/v1/records/${f.row.record_id}`);
    }
  }
});

test("Committed grid invalid timestamps survive detachment and clear only with explicit intent", async ({
  page,
}) => {
  const f = await fixture(page, evidenceViewSchemaId),
    field = "evidence.requested_at",
    bodies: string[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.row.record_id}`)
    )
      bodies.push(request.postData() ?? "");
  });
  const input = await activate(page, f.view, f.row.record_id, field);
  await input.fill(" 2026-02-30T12:00:00 ");
  await input.press("Enter");
  await expect(input).toBeFocused();
  expect(bodies).toHaveLength(0);
  await switchOrdinarySheet(page, hostsViewSchemaId);
  await switchOrdinarySheet(page, f.view);
  await expect(editor(page, f.view, f.row.record_id, field)).toHaveCount(0);
  const restored = await activate(page, f.view, f.row.record_id, field);
  await expect(restored).toHaveValue(" 2026-02-30T12:00:00 ");
  await restored.fill("2026-06-01T08:04:05.123456789+02:00");
  await restored.press("Enter");
  await expect(restored).toHaveCount(0);
  expect(JSON.parse(bodies[0] ?? "{}").changes).toEqual([
    { field_key: field, value: "2026-06-01T08:04:05.123456789+02:00" },
  ]);
  const clear = await activate(page, f.view, f.row.record_id, field);
  await clear.fill("");
  await clear.press("Enter");
  await expect(clear).toBeFocused();
  expect(bodies).toHaveLength(1);
  await page
    .getByRole("button", {
      name: `Clear ${requireViewContract(f.view).fieldMap[field]?.label}`,
      exact: true,
    })
    .click();
  await clear.press("Enter");
  await expect(clear).toHaveCount(0);
  expect(JSON.parse(bodies[1] ?? "{}").changes).toEqual([
    { field_key: field, value: null },
  ]);
  expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(3);
});

test("Committed grid direct editors cover all fifteen adopted surfaces", async ({
  page,
  workerAdmin,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("GEA-MATRIX"),
    "Committed field surface matrix",
  );
  const matrix: Array<{
    view: string;
    seed: Record<string, string>;
    field: string;
  }> = [
    {
      view: hostsViewSchemaId,
      seed: { "host.display_name": "Host" },
      field: "host.display_name",
    },
    {
      view: identitiesViewSchemaId,
      seed: { "identity.display_name": "Identity" },
      field: "identity.display_name",
    },
    {
      view: timelineViewSchemaId,
      seed: { "timeline.activity_synopsis_text": "Event" },
      field: "timeline.activity_synopsis_text",
    },
    {
      view: evidenceViewSchemaId,
      seed: { "evidence.title": "Evidence" },
      field: "evidence.title",
    },
    {
      view: partiesViewSchemaId,
      seed: { "party.display_name": "Party", "party.party_kind": "person" },
      field: "party.display_name",
    },
    {
      view: notesViewSchemaId,
      seed: { "note.title": "Note" },
      field: "note.title",
    },
    {
      view: taskRequestsViewSchemaId,
      seed: { "task.title": "Task", "task.task_kind": "question" },
      field: "task.title",
    },
    {
      view: decisionsViewSchemaId,
      seed: {
        "decision.summary": "Decision",
        "decision.decision_type": "scope",
        "decision.rationale": "Rationale",
      },
      field: "decision.summary",
    },
    {
      view: findingsViewSchemaId,
      seed: { "finding.statement": "Finding" },
      field: "finding.statement",
    },
    {
      view: forensicKeywordsViewSchemaId,
      seed: {
        "forensic_keyword.pattern": "Needle",
        "forensic_keyword.reason": "Reason",
      },
      field: "forensic_keyword.pattern",
    },
    {
      view: investigativeQueriesViewSchemaId,
      seed: {
        "investigative_query.platform": "SQL",
        "investigative_query.purpose": "Purpose",
        "investigative_query.query_text": "select value",
      },
      field: "investigative_query.platform",
    },
    {
      view: commLogViewSchemaId,
      seed: {
        "comm_log.comm_type": "briefing",
        "comm_log.audience": "Operations",
        "comm_log.channel_or_meeting": "Bridge",
        "comm_log.summary": "Summary",
      },
      field: "comm_log.summary",
    },
    {
      view: handoffViewSchemaId,
      seed: {
        "handoff.incoming_owner_user_id": workerAdmin.user_id,
        "handoff.current_state_summary": "State",
      },
      field: "handoff.current_state_summary",
    },
    {
      view: statusReviewViewSchemaId,
      seed: { "status_review.current_state_summary": "State" },
      field: "status_review.current_state_summary",
    },
    {
      view: lessonViewSchemaId,
      seed: { "lesson.summary": "Lesson" },
      field: "lesson.summary",
    },
  ];
  for (const entry of matrix)
    await test.step(entry.view, async () => {
      const row = await createViewRow(page, incident, entry.view, {
        client_txn_id: uniqueTxn("gea-matrix"),
        ...entry.seed,
      });
      await page.goto(`/?incident_id=${incident}&view_schema_id=${entry.view}`);
      await expect(page.getByTestId(gridShellTestId(entry.view))).toBeVisible();
      const input = await activate(
        page,
        entry.view,
        row.record_id,
        entry.field,
      );
      const request = page.waitForRequest(
        (request) =>
          request.method() === "PATCH" &&
          request.url().endsWith(`/records/${row.record_id}`),
      );
      await input.fill("Edited directly in the grid");
      await input.press("Enter");
      expect((await request).postDataJSON()).toMatchObject({
        view_schema_id: entry.view,
        base_row_version: 1,
        changes: [
          { field_key: entry.field, value: "Edited directly in the grid" },
        ],
      });
      await expect(input).toHaveCount(0);
      await expect(
        page.getByTestId(rowCellTestId(row.record_id, entry.field)),
      ).toHaveText("Edited directly in the grid");
      expect(await fetchRecordHistoryCount(page, row.record_id)).toBe(2);
    });
});

test("Committed grid numeric boolean enum multiline and stable-reference input retains exact local intent", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("GEA-FAMILIES"),
    "Committed editor families",
  );
  const findingView = findingsViewSchemaId,
    keywordView = forensicKeywordsViewSchemaId,
    noteView = notesViewSchemaId;
  const finding = await createViewRow(page, incident, findingView, {
    client_txn_id: uniqueTxn("gea-finding"),
    ...{ "finding.statement": "Finding" },
  });
  const keyword = await createViewRow(page, incident, keywordView, {
    client_txn_id: uniqueTxn("gea-keyword"),
    ...{
      "forensic_keyword.pattern": "Needle",
      "forensic_keyword.reason": "Reason",
    },
  });
  const note = await createViewRow(page, incident, noteView, {
    client_txn_id: uniqueTxn("gea-note"),
    "note.title": "Note",
  });
  const party = await createViewRow(page, incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("gea-party"),
    "party.display_name": "Source party",
    "party.party_kind": "person",
  });
  const evidence = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("gea-evidence"),
    "evidence.title": "Evidence",
  });
  const findingKind =
    requireViewContract(findingView).fieldMap["finding.kind"]?.enumValues?.[1];
  if (!findingKind)
    throw new Error("Finding kind contract has no second enum value");
  const cases = [
    {
      view: findingView,
      row: finding,
      field: "finding.confidence_score",
      raw: "42",
      invalid: "1e",
      expected: 42,
    },
    {
      view: findingView,
      row: finding,
      field: "finding.kind",
      raw: findingKind,
      expected: findingKind,
    },
    {
      view: keywordView,
      row: keyword,
      field: "forensic_keyword.case_sensitive",
      raw: "true",
      expected: true,
    },
    {
      view: noteView,
      row: note,
      field: "note.body",
      raw: "  Line one\nLine two  ",
      expected: "Line one\nLine two",
    },
    {
      view: evidenceViewSchemaId,
      row: evidence,
      field: "evidence.source_party_id",
      raw: party.record_id,
      expected: party.record_id,
    },
  ];
  for (const entry of cases)
    await test.step(entry.field, async () => {
      await page.goto(`/?incident_id=${incident}&view_schema_id=${entry.view}`);
      await expect(page.getByTestId(gridShellTestId(entry.view))).toBeVisible();
      let input = await activate(
        page,
        entry.view,
        entry.row.record_id,
        entry.field,
      );
      if (entry.invalid) {
        await input.fill(entry.invalid);
        await input.press("Enter");
        await expect(input).toBeFocused();
        await switchOrdinarySheet(page, hostsViewSchemaId);
        await switchOrdinarySheet(page, entry.view);
        input = await activate(
          page,
          entry.view,
          entry.row.record_id,
          entry.field,
        );
        await expect(input).toHaveValue(entry.invalid);
      }
      const tag = await input.evaluate((element) => element.tagName);
      if (tag === "SELECT") await input.selectOption(entry.raw);
      else if (entry.field === "forensic_keyword.case_sensitive")
        await input.check();
      else await input.fill(entry.raw);
      const request = page.waitForRequest(
        (request) =>
          request.method() === "PATCH" &&
          request.url().endsWith(`/records/${entry.row.record_id}`),
      );
      await input.press("Enter");
      expect((await request).postDataJSON().changes).toEqual([
        { field_key: entry.field, value: entry.expected },
      ]);
      await expect(input).toHaveCount(0);
      const saved = (await queryViewRows(page, incident, entry.view)).find(
        (row) => row.record_id === entry.row.record_id,
      );
      expect(saved?.cells[entry.field]?.value).toBe(entry.expected);
    });
});

test("Committed grid accepted rows survive failed stale and missing refresh with read-only remount recovery", async ({
  page,
}) => {
  for (const [view, mode] of [
    [hostsViewSchemaId, "failed"],
    [evidenceViewSchemaId, "stale"],
    [timelineViewSchemaId, "missing"],
  ] as const) {
    const f = await fixture(page, view),
      ack = gate();
    let committed = false,
      damageReads = true,
      reads = 0,
      writes = 0;
    let oldRows: unknown[] = [];
    await page.route(`**/views/${view}/query`, async (route) => {
      reads++;
      if (committed && damageReads && mode === "failed")
        return route.abort("failed");
      const response = await route.fetch(),
        payload = await response.json();
      if (!committed) oldRows = payload.data.rows;
      if (committed && damageReads)
        payload.data.rows = mode === "missing" ? [] : oldRows;
      await route.fulfill({ response, json: payload });
    });
    await page.reload();
    await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
    if (view === timelineViewSchemaId) {
      await scrollGridCellIntoView({
        page,
        surface: view,
        recordId: f.row.record_id,
        cellKey: f.field,
      });
      await page
        .getByRole("columnheader", { name: "Activity Synopsis", exact: true })
        .click();
    }
    await page.route(`**/api/v1/records/${f.row.record_id}`, async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      writes++;
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = true;
      await ack.promise;
      await route.fulfill({ response });
    });
    try {
      const input = await activate(page, view, f.row.record_id, f.field);
      await input.fill("Accepted despite refresh");
      await input.press("Enter");
      await expect.poll(() => committed).toBe(true);
      await input.fill("  newer local text  ");
      ack.release();
      await expect.poll(() => reads).toBeGreaterThan(1);
      if (mode === "missing") {
        await expect(input).toHaveCount(0);
        await page.getByText("Unsaved cells (1)", { exact: true }).click();
        await expect(
          page.getByRole("textbox", {
            name: "Retained Activity Synopsis",
            exact: true,
          }),
        ).toHaveValue("  newer local text  ");
      } else {
        await expect(input).toHaveValue("  newer local text  ");
        await expect(input).toBeFocused();
      }
      await switchOrdinarySheet(
        page,
        view === timelineViewSchemaId
          ? hostsViewSchemaId
          : timelineViewSchemaId,
      );
      damageReads = false;
      const beforeReturn = reads;
      await switchOrdinarySheet(page, view);
      await expect.poll(() => reads).toBeGreaterThan(beforeReturn);
      await expect(editor(page, view, f.row.record_id, f.field)).toHaveCount(0);
      await scrollGridCellIntoView({
        page,
        surface: view,
        recordId: f.row.record_id,
        cellKey: f.field,
      });
      await expect(
        page.getByTestId(rowCellTestId(f.row.record_id, f.field)),
      ).toHaveText("Accepted despite refresh");
      const restored = await activate(page, view, f.row.record_id, f.field);
      await expect(restored).toHaveValue("  newer local text  ");
      await restored.press("Escape");
      expect(writes).toBe(1);
      expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(2);
    } finally {
      ack.release();
      await page.unroute(`**/views/${view}/query`);
      await page.unroute(`**/api/v1/records/${f.row.record_id}`);
    }
  }
});

test("Committed grid closure retains copyable original-cell drafts and reopening requires explicit activation", async ({
  page,
}) => {
  for (const view of [
    hostsViewSchemaId,
    evidenceViewSchemaId,
    timelineViewSchemaId,
  ]) {
    const f = await fixture(page, view);
    let writes = 0;
    page.on("request", (request) => {
      if (
        request.method() === "PATCH" &&
        request.url().endsWith(`/records/${f.row.record_id}`)
      )
        writes++;
    });
    const input = await activate(page, view, f.row.record_id, f.field);
    await input.fill("  private unfinished text  ");
    const before = await currentLifecycle(page, f.incident);
    const closed = await lifecycleAction(page, f.incident, "closeIncident", {
      client_txn_id: uniqueTxn("gea-close"),
      base_incident_version: before.incident_version,
      reason: "Retain grid authoring",
    });
    expect(closed.ok).toBe(true);
    await expect(input).toHaveCount(0);
    await page.getByText("Unsaved cells (1)", { exact: true }).click();
    const retained = page.getByRole("textbox", {
      name: `Retained ${requireViewContract(view).fieldMap[f.field]?.label}`,
      exact: true,
    });
    await expect(retained).toHaveValue("  private unfinished text  ");
    await expect(retained).toHaveAttribute("readonly", "");
    expect(writes).toBe(0);
    await openLifecycle(page);
    await confirmLifecycle(page, "Reopen", "Resume original target");
    await page
      .getByRole("button", { name: "Close incident controls", exact: true })
      .click();
    await expect(retained).toHaveCount(0);
    await expect(editor(page, view, f.row.record_id, f.field)).toHaveCount(0);
    const restored = await activate(page, view, f.row.record_id, f.field);
    await expect(restored).toHaveValue("  private unfinished text  ");
    await restored.press("Escape");
    expect(writes).toBe(0);
    expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(1);
  }
});

test("Committed grid session suspension conceals raw work and account replacement retires it", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const f = await fixture(page, evidenceViewSchemaId);
  const members = [];
  for (const name of ["original", "replacement"])
    members.push(
      await createIncidentMemberUser(page, f.incident, {
        email: uniqueEmail(`gea-${name}`),
        display_name: name,
        initial_password: "RetainedGrid1!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      }),
    );
  const original = members[0],
    replacement = members[1];
  if (!original || !replacement) throw new Error("Missing test members");
  const login = (member: typeof original, recovery = false) =>
    sessionTracker.loginTrackedUser(page, {
      recovery,
      createdBy: "grid-autosave",
      email: member.email,
      password: member.initial_password,
      purpose: "Committed grid authority lifetime",
      userId: member.user_id,
    });
  await login(original);
  const sockets = installIncidentSocketMonitor(page, f.incident);
  await page.goto(`/?incident_id=${f.incident}&view_schema_id=${f.view}`);
  await sockets.waitForAcceptedSocket();
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.row.record_id}`)
    )
      writes++;
  });
  const input = await activate(page, f.view, f.row.record_id, f.field);
  await input.fill("Protected exact local text");
  await revokeAllSessions(
    workerAdminRequest,
    original.user_id,
    "Grid suspension evidence",
  );
  await expect(page.getByTestId(authTestId("shell"))).toBeVisible();
  await expect(input).toHaveCount(0);
  await expect(
    page.getByText("Protected exact local text", { exact: true }),
  ).toHaveCount(0);
  await login(original, true);
  await expect(page.getByTestId(gridShellTestId(f.view))).toBeVisible();
  await expect(input).toHaveCount(0);
  const restored = await activate(page, f.view, f.row.record_id, f.field);
  await expect(restored).toHaveValue("Protected exact local text");
  await revokeAllSessions(
    workerAdminRequest,
    original.user_id,
    "Replace author account",
  );
  await expect(page.getByTestId(authTestId("shell"))).toBeVisible();
  await login(replacement, true);
  await expect(page.getByTestId(gridShellTestId(f.view))).toBeVisible();
  const fresh = await activate(page, f.view, f.row.record_id, f.field);
  await expect(fresh).toHaveValue("Original row");
  await fresh.press("Escape");
  expect(writes).toBe(0);
  const history = await workerAdminRequest.get(
    `/api/v1/records/${f.row.record_id}/history`,
  );
  expect(history.ok()).toBe(true);
  expect((await history.json()).data.items).toHaveLength(1);
});

test("Committed grid role loss retains readable text and incident revocation retires protected presentation", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const f = await fixture(page, evidenceViewSchemaId);
  const member = await createIncidentMemberUser(page, f.incident, {
    email: uniqueEmail("gea-role"),
    display_name: "Grid author",
    initial_password: "RetainedGrid1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "grid-autosave-role",
    email: member.email,
    password: member.initial_password,
    purpose: "Role and incident lifetime",
    userId: member.user_id,
  });
  const sockets = installIncidentSocketMonitor(page, f.incident);
  await page.goto(`/?incident_id=${f.incident}&view_schema_id=${f.view}`);
  await sockets.waitForAcceptedSocket();
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.row.record_id}`)
    )
      writes++;
  });
  const input = await activate(page, f.view, f.row.record_id, f.field);
  await input.fill("  authoring before role loss  ");
  const path = `/api/v1/incidents/${f.incident}/memberships/${member.user_id}`;
  expect(
    (
      await workerAdminRequest.patch(path, {
        data: { base_membership_version: 1, role: "viewer" },
      })
    ).ok(),
  ).toBe(true);
  await input.press("Enter");
  await expect(input).toHaveCount(0);
  await page.getByText("Unsaved cells (1)", { exact: true }).click();
  const retained = page.getByRole("textbox", {
    name: "Retained Title",
    exact: true,
  });
  await expect(retained).toHaveValue("  authoring before role loss  ");
  await expect(retained).toHaveAttribute("readonly", "");
  expect(writes).toBe(1);
  expect(
    (
      await workerAdminRequest.delete(path, {
        data: { base_membership_version: 2 },
      })
    ).status(),
  ).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(retained).toHaveCount(0);
  await expect(page.getByTestId(authTestId("shell"))).toHaveCount(0);
  await expect(page).not.toHaveURL(/incident_id=/u);
  expect(writes).toBe(1);
});

test("Committed grid dependent inspector writes wait for the accepted row version", async ({
  page,
}) => {
  const f = await fixture(page, hostsViewSchemaId),
    ack = gate(),
    bodies: string[] = [];
  let committed = false;
  await page.route(`**/api/v1/records/${f.row.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    const index = bodies.push(route.request().postData() ?? "") - 1;
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    if (index === 0) {
      committed = true;
      await ack.promise;
    }
    await route.fulfill({ response });
  });
  try {
    const input = await activate(page, f.view, f.row.record_id, f.field);
    await input.fill("Grid first");
    await input.press("Enter");
    await expect.poll(() => committed).toBe(true);
    await page.getByTestId(workbookInspectorToggleTestId(f.view)).click();
    await page
      .getByTestId(genericEditFieldSelectTestId(f.view))
      .selectOption("host.location");
    const inspector = page.getByTestId(genericEditValueTestId(f.view));
    await inspector.fill("Inspector second");
    await page.getByTestId(genericEditSubmitTestId(f.view)).click();
    expect(bodies).toHaveLength(1);
    ack.release();
    await expect.poll(() => bodies.length).toBe(2);
    const requests = bodies.map((body) => JSON.parse(body));
    expect(requests.map((body) => body.base_row_version)).toEqual([1, 2]);
    expect(requests[1].changes).toEqual([
      { field_key: "host.location", value: "Inspector second" },
    ]);
    expect(requests[1].client_txn_id).not.toBe(requests[0].client_txn_id);
    await expect
      .poll(() => fetchRecordHistoryCount(page, f.row.record_id))
      .toBe(3);
    const saved = (await queryViewRows(page, f.incident, f.view)).find(
      (row) => row.record_id === f.row.record_id,
    );
    expect(saved?.cells[f.field]?.value).toBe("Grid first");
    expect(saved?.cells["host.location"]?.value).toBe("Inspector second");
  } finally {
    ack.release();
    await page.unroute(`**/api/v1/records/${f.row.record_id}`);
  }
});

test("a11y.grid-autosave correction actions preserve keyboard access composition and narrow-layout focus", async ({
  page,
}, info) => {
  const f = await fixture(page, evidenceViewSchemaId),
    field = "evidence.requested_at";
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.row.record_id}`)
    )
      writes++;
  });
  await page.emulateMedia({ reducedMotion: "reduce" });
  for (const [width, zoom] of [
    [1280, 1],
    [390, 1],
    [1280, 2],
  ] as const) {
    await page.setViewportSize({ width, height: 720 });
    await page.evaluate((zoom) => {
      document.documentElement.style.zoom = String(zoom);
    }, zoom);
    const input = await activate(page, f.view, f.row.record_id, field);
    await input.dispatchEvent("compositionstart");
    await input.fill("  unfinished timestamp  ");
    await input.dispatchEvent("keydown", {
      key: "Enter",
      isComposing: true,
      bubbles: true,
    });
    await expect(input).toBeFocused();
    await expect(input).toHaveValue("  unfinished timestamp  ");
    await input.dispatchEvent("compositionend", {
      data: "  unfinished timestamp  ",
    });
    await input.press("Enter");
    await expect(input).toHaveAttribute("aria-invalid", "true");
    await expect(input).toBeInViewport({ ratio: 1 });
    await input.press("Alt+ArrowDown");
    const clear = page.getByRole("button", {
      name: "Clear Requested",
      exact: true,
    });
    await expect(clear).toBeFocused();
    await expect(clear).toBeInViewport({ ratio: 1 });
    await clear.press("Tab");
    await expect(
      page.getByRole("button", { name: "Commit", exact: true }),
    ).toBeFocused();
    await page.keyboard.press("Shift+Tab");
    await expect(clear).toBeFocused();
    await info.attach(`grid-autosave-${width}-${zoom}`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
    await info.attach(`grid-autosave-aria-${width}-${zoom}`, {
      body: await page
        .getByRole("group", { name: "Edit Requested", exact: true })
        .ariaSnapshot(),
      contentType: "text/plain",
    });
    await clear.press("Escape");
    await expect(input).toHaveCount(0);
    expect(writes).toBe(0);
  }
});

test("Committed grid different-field successors preserve collaboration versions and require review for changed authoring fields", async ({
  page,
}) => {
  const f = await fixture(page, hostsViewSchemaId),
    ack = gate(),
    bodies: string[] = [];
  let committed = false;
  await page.route(`**/api/v1/records/${f.row.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    const index = bodies.push(route.request().postData() ?? "") - 1,
      response = await route.fetch();
    expect(response.ok()).toBe(true);
    if (index === 0) {
      committed = true;
      await ack.promise;
    }
    await route.fulfill({ response });
  });
  try {
    const a = await activate(page, f.view, f.row.record_id, f.field);
    await a.fill("Grid A");
    await a.press("Enter");
    await expect.poll(() => committed).toBe(true);
    await switchOrdinarySheet(page, timelineViewSchemaId);
    await switchOrdinarySheet(page, f.view);
    const b = await activate(page, f.view, f.row.record_id, "host.location");
    await b.fill("Different field B");
    await patchRecord(page, f.row.record_id, {
      view_schema_id: f.view,
      base_row_version: 2,
      client_txn_id: uniqueTxn("gea-remote-os"),
      changes: [{ field_key: "host.os_platform", value: "Remote OS" }],
    });
    await f.sockets.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload.record_id === f.row.record_id &&
        message.payload.row_version === 3,
    });
    await switchOrdinarySheet(page, timelineViewSchemaId);
    await switchOrdinarySheet(page, f.view);
    const restored = await activate(
      page,
      f.view,
      f.row.record_id,
      "host.location",
    );
    await expect(restored).toHaveValue("Different field B");
    await restored.press("Enter");
    expect(bodies).toHaveLength(1);
    ack.release();
    await expect(restored).toHaveCount(0);
    expect(bodies.map((body) => JSON.parse(body).base_row_version)).toEqual([
      1, 3,
    ]);
    expect(JSON.parse(bodies[1] ?? "{}").changes).toEqual([
      { field_key: "host.location", value: "Different field B" },
    ]);
    expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(4);
    const local = await activate(
      page,
      f.view,
      f.row.record_id,
      "host.location",
    );
    await local.fill("Local review value");
    await patchRecord(page, f.row.record_id, {
      view_schema_id: f.view,
      base_row_version: 4,
      client_txn_id: uniqueTxn("gea-remote-location"),
      changes: [{ field_key: "host.location", value: "Remote location" }],
    });
    await f.sockets.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload.record_id === f.row.record_id &&
        message.payload.row_version === 5,
    });
    await switchOrdinarySheet(page, timelineViewSchemaId);
    await switchOrdinarySheet(page, f.view);
    const review = await activate(
      page,
      f.view,
      f.row.record_id,
      "host.location",
    );
    await expect(review).toHaveValue("Local review value");
    await review.press("Enter");
    expect(bodies).toHaveLength(2);
    await page
      .getByRole("button", { name: "Keep draft Location", exact: true })
      .click();
    await review.press("Enter");
    await expect(review).toHaveCount(0);
    expect(JSON.parse(bodies[2] ?? "{}").base_row_version).toBe(5);
    expect(
      new Set(bodies.map((body) => JSON.parse(body).client_txn_id)).size,
    ).toBe(3);
    expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(6);
    const saved = (await queryViewRows(page, f.incident, f.view)).find(
      (row) => row.record_id === f.row.record_id,
    );
    expect(saved?.row_version).toBe(6);
    expect(saved?.cells["host.os_platform"]?.value).toBe("Remote OS");
    expect(saved?.cells["host.location"]?.value).toBe("Local review value");
  } finally {
    ack.release();
    await page.unroute(`**/api/v1/records/${f.row.record_id}`);
  }
});

test("Committed grid drafts stay on their original target through saved views hidden fields filtering and deletion", async ({
  page,
}) => {
  const f = await fixture(page, evidenceViewSchemaId),
    field = "evidence.requested_at";
  await patchRecord(page, f.other.record_id, {
    view_schema_id: f.view,
    base_row_version: 1,
    client_txn_id: uniqueTxn("gea-filter-fixture"),
    changes: [{ field_key: "evidence.storage_ref", value: "kept-result" }],
  });
  const originalView = await createSavedView(page, f.incident, {
    display_name: "Original eligible cells",
    view_schema_id: f.view,
    query_json: {},
    layout_json: {
      layout_schema_id: "cartulary.layout.v1",
      column_order: requireViewContract(f.view).fields.map(
        (field) => field.fieldKey,
      ),
      column_widths: [],
      hidden_field_keys: [],
    },
  });
  const hidden = await createSavedView(page, f.incident, {
    display_name: "Hidden original field",
    view_schema_id: f.view,
    layout_json: {
      layout_schema_id: "cartulary.layout.v1",
      column_order: requireViewContract(f.view).fields.map(
        (field) => field.fieldKey,
      ),
      column_widths: [],
      hidden_field_keys: [field],
    },
  });
  const filtered = await createSavedView(page, f.incident, {
    display_name: "Other row only",
    view_schema_id: f.view,
    query_json: {
      filters: [
        {
          field_key: "evidence.storage_ref",
          op: "eq",
          arg: { value: "kept-result" },
        },
      ],
      sort: [{ field_key: "evidence.requested_at", direction: "asc" }],
    },
  });
  await page.reload();
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.row.record_id}`)
    )
      writes++;
  });
  const input = await activate(page, f.view, f.row.record_id, field);
  await input.fill("  unfinished original timestamp  ");
  for (const saved of [hidden, filtered]) {
    await selectSavedView(page, f.view, saved.saved_view_id);
    await expect(input).toHaveCount(0);
    await page.getByText("Unsaved cells (1)", { exact: true }).click();
    await expect(
      page.getByRole("textbox", { name: "Retained Requested", exact: true }),
    ).toHaveValue("  unfinished original timestamp  ");
    expect(writes).toBe(0);
    await selectSavedView(page, f.view, originalView.saved_view_id);
    await expect(input).toHaveCount(0);
    await page.getByTestId(workbookColumnsMenuTriggerTestId(f.view)).click();
    const requested = page.getByRole("menuitemcheckbox", {
      name: "Requested",
      exact: true,
    });
    if ((await requested.getAttribute("aria-checked")) === "false")
      await requested.click();
    await page.getByTestId(workbookColumnsMenuTriggerTestId(f.view)).click();
    const restored = await activate(page, f.view, f.row.record_id, field);
    await expect(restored).toHaveValue("  unfinished original timestamp  ");
  }
  const removed = await publicHttpOperation({
    operationID: "deleteRecord",
    pathParameters: { record_id: f.row.record_id },
    body: {
      base_row_version: 1,
      client_txn_id: uniqueTxn("gea-delete"),
      reason: "Original target deletion",
    },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(removed.ok).toBe(true);
  await expect(input).toHaveCount(0);
  const summary = page.getByText("Unsaved cells (1)", { exact: true });
  await summary.click();
  await expect(
    page.getByRole("textbox", { name: "Retained Requested", exact: true }),
  ).toHaveValue("  unfinished original timestamp  ");
  await page
    .getByRole("button", { name: "Discard Requested draft", exact: true })
    .click();
  await expect(summary).toHaveCount(0);
  expect(writes).toBe(0);
  expect(
    (await queryViewRows(page, f.incident, f.view)).find(
      (row) => row.record_id === f.other.record_id,
    )?.row_version,
  ).toBe(2);
});

test("Committed grid CSRF denial recovers current authorization and replays the same request without losing newer text", async ({
  page,
}) => {
  const f = await fixture(page, evidenceViewSchemaId),
    ack = gate(),
    bodies: string[] = [];
  let denied = false,
    accepted = false;
  await page.route(`**/api/v1/records/${f.row.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    const index = bodies.push(route.request().postData() ?? "") - 1;
    const response = await route.fetch(
      index === 0
        ? {
            headers: {
              ...Object.fromEntries(
                Object.entries(route.request().headers()).filter(
                  ([name]) =>
                    name.toLowerCase() !== csrfHeaderName.toLowerCase(),
                ),
              ),
            },
          }
        : {},
    );
    if (index === 0) {
      expect(response.status()).toBe(403);
      expect((await response.json()).error.code).toBe(
        "csrf_verification_failed",
      );
      denied = true;
      await ack.promise;
    } else {
      expect(response.ok()).toBe(true);
      accepted = true;
    }
    await route.fulfill({ response });
  });
  try {
    const input = await activate(page, f.view, f.row.record_id, f.field);
    await input.fill("Accepted after CSRF recovery");
    await input.press("Enter");
    await expect.poll(() => denied).toBe(true);
    await input.fill("  newer local after denial  ");
    ack.release();
    await expect.poll(() => accepted).toBe(true);
    await expect(page.getByTestId(authTestId("shell"))).toHaveCount(0);
    expect(bodies).toHaveLength(2);
    expect(bodies[1]).toBe(bodies[0]);
    expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(2);
    await switchOrdinarySheet(page, timelineViewSchemaId);
    await switchOrdinarySheet(page, f.view);
    const restored = await activate(page, f.view, f.row.record_id, f.field);
    await expect(restored).toHaveValue("  newer local after denial  ");
    await restored.press("Escape");
    expect(
      (await queryViewRows(page, f.incident, f.view)).find(
        (row) => row.record_id === f.row.record_id,
      )?.cells[f.field]?.value,
    ).toBe("Accepted after CSRF recovery");
  } finally {
    ack.release();
    await page.unroute(`**/api/v1/records/${f.row.record_id}`);
  }
});

test("Committed Timeline preparation reconciles collaboration between a real commit and its HTTP acknowledgement", async ({
  page,
}) => {
  const f = await fixture(page, timelineViewSchemaId),
    ack = gate(),
    bodies: string[] = [];
  let committed = false;
  await page.route(`**/api/v1/records/${f.row.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    const index = bodies.push(route.request().postData() ?? "") - 1;
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    if (index === 0) {
      committed = true;
      await ack.promise;
    }
    await route.fulfill({ response });
  });
  try {
    const input = await activate(page, f.view, f.row.record_id, f.field);
    await input.fill("Committed A");
    await input.press("Enter");
    await expect.poll(() => committed).toBe(true);
    await input.fill("Later B");
    await patchRecord(page, f.row.record_id, {
      view_schema_id: f.view,
      base_row_version: 2,
      client_txn_id: uniqueTxn("gea-timeline-remote"),
      changes: [
        {
          field_key: "timeline.data_source_text",
          value: "Collaborator source",
        },
      ],
    });
    await f.sockets.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload.record_id === f.row.record_id &&
        message.payload.row_version === 3,
    });
    await input.press("Enter");
    expect(bodies).toHaveLength(1);
    ack.release();
    await expect(input).toHaveCount(0);
    expect(bodies.map((body) => JSON.parse(body).base_row_version)).toEqual([
      1, 3,
    ]);
    expect(
      new Set(bodies.map((body) => JSON.parse(body).client_txn_id)).size,
    ).toBe(2);
    const row = (await queryViewRows(page, f.incident, f.view)).find(
      (row) => row.record_id === f.row.record_id,
    );
    expect(row?.row_version).toBe(4);
    expect(row?.cells[f.field]?.value).toBe("Later B");
    expect(row?.cells["timeline.data_source_text"]?.value).toBe(
      "Collaborator source",
    );
    expect(await fetchRecordHistoryCount(page, f.row.record_id)).toBe(4);
  } finally {
    ack.release();
    await page.unroute(`**/api/v1/records/${f.row.record_id}`);
  }
});

test("Existing reference cell picker keeps keyboard browsing local and commits exact later-page Party identity once", async ({
  page,
}) => {
  test.setTimeout(180_000);
  const f = await fixture(page, evidenceViewSchemaId);
  for (let offset = 0; offset < 105; offset += 5)
    await Promise.all(
      Array.from({ length: 5 }, (_, index) =>
        createViewRow(page, f.incident, partiesViewSchemaId, {
          client_txn_id: uniqueTxn("rsr-party"),
          "party.display_name": `Picker Party ${String(offset + index).padStart(3, "0")}`,
          "party.party_kind": "person",
        }),
      ),
    );
  const field = "evidence.source_party_id";
  const input = await activate(page, f.view, f.row.record_id, field);
  const writes: Record<string, unknown>[] = [];
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.row.record_id}`)
    )
      writes.push(request.postDataJSON());
  });
  await input.fill("");
  await input.press("Enter");
  await expect(input).toBeFocused();
  expect(writes).toHaveLength(0);
  const trigger = input
    .locator("..")
    .getByRole("button", { name: "Choose source party", exact: true });
  await trigger.focus();
  await trigger.press("Enter");
  const popup = page.getByRole("dialog", {
    name: "Choose source party",
    exact: true,
  });
  await expect(popup).toContainText("Page 1: 100 candidates; more available");
  const next = popup.getByRole("button", { name: "Next", exact: true });
  await next.focus();
  await next.press("Enter");
  await expect(popup).toContainText("Page 2: 5 candidates; end of this source");
  const list = popup.getByRole("listbox", { name: "Source Party candidates" });
  await list.focus();
  await list.press("ArrowDown");
  const chosen = await list.inputValue();
  expect(chosen).toMatch(/^party:/);
  await list.press("Enter");
  expect(writes).toHaveLength(0);
  await list.press("Escape");
  await expect(popup).toHaveCount(0);
  await expect(input).toBeFocused();
  await expect(input).toHaveValue("");
  await trigger.focus();
  await trigger.press("Enter");
  await expect(popup).toContainText("Page 1: 100 candidates");
  await next.focus();
  await next.press("Enter");
  await expect(popup).toContainText("Page 2: 5 candidates");
  await list.selectOption(chosen);
  const accept = popup.getByRole("button", {
    name: "Use selection",
    exact: true,
  });
  await accept.focus();
  await accept.press("Enter");
  await expect(input).toHaveCount(0);
  expect(writes).toHaveLength(1);
  expect(writes[0]?.changes).toEqual([
    { field_key: field, value: chosen.replace("party:", "") },
  ]);
  const clearInput = await activate(page, f.view, f.row.record_id, field);
  await page
    .getByRole("button", { name: "Clear Source Party", exact: true })
    .click();
  await page.getByRole("button", { name: "Commit", exact: true }).click();
  await expect(clearInput).toHaveCount(0);
  expect(writes).toHaveLength(2);
  expect(writes[1]?.changes).toEqual([{ field_key: field, value: null }]);
});

test("Existing reference drafts survive same-account recovery while staged late lookups and replaced accounts are fenced", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const f = await fixture(page, evidenceViewSchemaId);
  const party = await createViewRow(page, f.incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("rsr-session-party"),
    "party.display_name": "Protected reference label",
    "party.party_kind": "person",
  });
  const members = [];
  for (const name of ["original", "replacement"])
    members.push(
      await createIncidentMemberUser(page, f.incident, {
        email: uniqueEmail(`rsr-${name}`),
        display_name: name,
        initial_password: "ReferenceRecovery1!",
        role: "editor",
        is_deployment_admin: false,
        mfa_required: false,
      }),
    );
  const original = members[0],
    replacement = members[1];
  if (!original || !replacement) throw new Error("Missing reference authors");
  const login = (member: typeof original, recovery = false) =>
    sessionTracker.loginTrackedUser(page, {
      recovery,
      createdBy: "reference-recovery",
      email: member.email,
      password: member.initial_password,
      purpose: "Reference draft authority",
      userId: member.user_id,
    });
  await login(original);
  const sockets = installIncidentSocketMonitor(page, f.incident);
  await page.goto(`/?incident_id=${f.incident}&view_schema_id=${f.view}`);
  await sockets.waitForAcceptedSocket();
  const field = "evidence.source_party_id";
  const input = await activate(page, f.view, f.row.record_id, field);
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.row.record_id}`)
    )
      writes++;
  });
  await input.fill(party.record_id);
  const delayed = gate();
  let lookupStarted = false;
  await page.route(`**/views/${partiesViewSchemaId}/query`, async (route) => {
    lookupStarted = true;
    await delayed.promise;
    await route.continue().catch(() => {});
  });
  await input
    .locator("..")
    .getByRole("button", { name: "Choose source party", exact: true })
    .click();
  await expect.poll(() => lookupStarted).toBe(true);
  await revokeAllSessions(
    workerAdminRequest,
    original.user_id,
    "Reference read suspension",
  );
  await expect(page.getByTestId(authTestId("shell"))).toBeVisible();
  await expect(
    page.getByRole("dialog", { name: "Choose source party" }),
  ).toHaveCount(0);
  await expect(input).toHaveCount(0);
  delayed.release();
  await login(original, true);
  await expect(page.getByTestId(gridShellTestId(f.view))).toBeVisible();
  await expect(input).toHaveCount(0);
  const restored = await activate(page, f.view, f.row.record_id, field);
  await expect(restored).toHaveValue(party.record_id);
  await expect(
    page.getByRole("dialog", { name: "Choose source party" }),
  ).toHaveCount(0);
  await revokeAllSessions(
    workerAdminRequest,
    original.user_id,
    "Reference account replacement",
  );
  await expect(page.getByTestId(authTestId("shell"))).toBeVisible();
  await login(replacement, true);
  await expect(page.getByTestId(gridShellTestId(f.view))).toBeVisible();
  const fresh = await activate(page, f.view, f.row.record_id, field);
  await expect(fresh).toHaveValue("");
  await fresh.press("Escape");
  expect(writes).toBe(0);
});

test("Existing reference closure role changes and access loss separate readable choices from write admission", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const f = await fixture(page, evidenceViewSchemaId);
  const party = await createViewRow(page, f.incident, partiesViewSchemaId, {
    client_txn_id: uniqueTxn("rsr-role-party"),
    "party.display_name": "Readable reference target",
    "party.party_kind": "person",
  });
  const field = "evidence.source_party_id";
  const input = await activate(page, f.view, f.row.record_id, field);
  await input.fill(party.record_id);
  await input
    .locator("..")
    .getByRole("button", { name: "Choose source party", exact: true })
    .click();
  const popup = page.getByRole("dialog", {
    name: "Choose source party",
    exact: true,
  });
  await expect(popup).toContainText("Readable reference target");
  let writes = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.row.record_id}`)
    )
      writes++;
  });
  const before = await currentLifecycle(page, f.incident);
  expect(
    (
      await lifecycleAction(page, f.incident, "closeIncident", {
        client_txn_id: uniqueTxn("rsr-close"),
        base_incident_version: before.incident_version,
        reason: "Reference lifecycle evidence",
      })
    ).ok,
  ).toBe(true);
  await expect(popup).toHaveCount(0);
  await expect(input).toHaveCount(0);
  await page.getByText("Unsaved cells (1)", { exact: true }).click();
  const retained = page.getByRole("textbox", {
    name: "Retained Source Party",
    exact: true,
  });
  await expect(retained).toHaveValue(party.record_id);
  await expect(retained).toHaveAttribute("readonly", "");
  expect(writes).toBe(0);
  await openLifecycle(page);
  await confirmLifecycle(page, "Reopen", "Resume reference authoring");
  await page
    .getByRole("button", { name: "Close incident controls", exact: true })
    .click();
  await expect(input).toHaveCount(0);
  const restored = await activate(page, f.view, f.row.record_id, field);
  await expect(restored).toHaveValue(party.record_id);
  await restored.press("Escape");
  const member = await createIncidentMemberUser(page, f.incident, {
    email: uniqueEmail("rsr-role"),
    display_name: "Reference author",
    initial_password: "ReferenceRole1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  await sessionTracker.loginTrackedUser(page, {
    createdBy: "reference-role",
    email: member.email,
    password: member.initial_password,
    purpose: "Reference role and access lifetime",
    userId: member.user_id,
  });
  const sockets = installIncidentSocketMonitor(page, f.incident);
  await page.goto(`/?incident_id=${f.incident}&view_schema_id=${f.view}`);
  await sockets.waitForAcceptedSocket();
  const draft = await activate(page, f.view, f.row.record_id, field);
  await draft.fill(party.record_id);
  await draft
    .locator("..")
    .getByRole("button", { name: "Choose source party", exact: true })
    .click();
  await expect(popup).toContainText("Readable reference target");
  const membership = `/api/v1/incidents/${f.incident}/memberships/${member.user_id}`;
  expect(
    (
      await workerAdminRequest.patch(membership, {
        data: { base_membership_version: 1, role: "viewer" },
      })
    ).ok(),
  ).toBe(true);
  // Role discovery follows the existing denied-write recovery owner; the read
  // remains authorized while the next attempted mutation is rejected.
  await popup
    .getByRole("button", { name: "Cancel references", exact: true })
    .click();
  await draft.press("Enter");
  await expect(draft).toHaveCount(0);
  await page.getByText("Unsaved cells (1)", { exact: true }).click();
  await expect(retained).toHaveValue(party.record_id);
  await expect(retained).toHaveAttribute("readonly", "");
  expect(writes).toBe(1);
  expect(
    (
      await workerAdminRequest.delete(membership, {
        data: { base_membership_version: 2 },
      })
    ).status(),
  ).toBe(204);
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(retained).toHaveCount(0);
  await expect(popup).toHaveCount(0);
  await expect(page.getByTestId(authTestId("shell"))).toHaveCount(0);
  expect(writes).toBe(1);
});
