import type {
  CreateViewRowResponse,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  rowCellTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { openIncidentAsTrackedUser } from "./pages/incidentDirectory";
import { csrfHeaders } from "./support/auth/browserSession";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { publicHttpOperation } from "./support/transport/publicHttpOperationClient";
import { atJsonOrigin } from "./support/transport/publicJsonClient";
import { fetchFullRecordHistory } from "./support/workbook/history";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openRecoveryItem } from "./support/workbook/recovery";
import { openTimelineInspector } from "./support/workbook/rowMutations";

const raw = "Original Timeline text: preserve  spacing and source spelling.";
async function fixture(page: Page) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("TRE"),
    "Timeline Evidence recovery",
  );
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "timeline.activity_synopsis_text": "Original source",
    "timeline.raw_activity_text": raw,
  });
  const reviewed = await publicHttpOperation({
    operationID: "markTimelineRecordReviewed",
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    pathParameters: { record_id: source.record_id },
    body: {
      base_row_version: source.row_version,
      client_txn_id: uniqueTxn("review"),
    },
  });
  expect(reviewed.ok).toBe(true);
  const current = (
    await queryViewRows(page, incident, timelineViewSchemaId)
  ).find((row) => row.record_id === source.record_id);
  if (!current) throw new Error("Missing original source");
  expect(current.cells["timeline.capture_state"]?.value).toBe("reviewed");
  const before = await fetchFullRecordHistory(page, source.record_id);
  return { incident, source: current, before };
}
async function begin(page: Page, incident: string, sourceId: string) {
  await page.goto(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await openTimelineInspector(page, sourceId);
  await expect(
    page.getByTestId(
      workbookInspectorPanelTestId(timelineViewSchemaId, "workflow"),
    ),
  ).toBeVisible();
  await page
    .getByTestId(
      workbookInspectorFeatureActionTestId(
        timelineViewSchemaId,
        "create_related.evidence",
      ),
    )
    .click();
  await page
    .getByTestId(genericCreateFieldTestId("evidence.collector_party_text"))
    .fill("Preserved collector text");
}
async function submit(page: Page) {
  await page
    .getByTestId(genericCreateSubmitTestId(evidenceViewSchemaId))
    .click();
}
async function recovery(page: Page) {
  await openRecoveryItem(page, /^Timeline Evidence (draft|creation) ·/);
  const region = page.getByRole("region", {
    name: "Retained Timeline Evidence creation",
    exact: true,
  });
  await expect(region).toBeVisible();
  return region;
}
async function verifyLinked(
  page: Page,
  incident: string,
  sourceId: string,
  targetId: string,
  version: number,
) {
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incident, timelineViewSchemaId)).find(
          (row) => row.record_id === sourceId,
        )?.row_version,
    )
    .toBe(version);
  const source = (
    await queryViewRows(page, incident, timelineViewSchemaId)
  ).find((row) => row.record_id === sourceId);
  expect(source?.cells["timeline.raw_activity_text"]?.value).toBe(raw);
  expect(source?.cells["timeline.capture_state"]?.value).toBe("enriched");
  const links = source?.cells["timeline.attached_evidence_ids"]?.value as {
    items: { linked_record_id: string }[];
  };
  expect(
    links.items.filter((item) => item.linked_record_id === targetId),
  ).toHaveLength(1);
  const target = (
    await queryViewRows(page, incident, evidenceViewSchemaId)
  ).find((row) => row.record_id === targetId);
  expect(target?.cells["evidence.linked_record_count"]?.value).toBe(1);
  expect(target?.cells["evidence.lifecycle_state"]?.value).toBe("requested");
  expect(target?.cells["evidence.blob_hash"]?.value).toBeNull();
}

test("Timeline Evidence creation survives a real collection conflict and links only the retained Evidence after review", async ({
  page,
}) => {
  const f = await fixture(page);
  const existing = await createViewRow(page, f.incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("existing"),
    "evidence.title": "Existing Evidence",
  });
  const creations: CreateViewRowResponse[] = [],
    links: string[] = [];
  await page.route(`**/views/${evidenceViewSchemaId}/rows`, async (route) => {
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    creations.push(await response.json());
    await route.fulfill({ response });
  });
  await page.route(`**/api/v1/records/${f.source.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    links.push(route.request().postData() ?? "");
    if (links.length === 1)
      await patchRecord(page, f.source.record_id, {
        view_schema_id: timelineViewSchemaId,
        base_row_version: f.source.row_version,
        client_txn_id: uniqueTxn("concurrent-link"),
        changes: [
          {
            field_key: "timeline.attached_evidence_ids",
            action_payload: {
              kind: "collection_actions_v1",
              actions: [
                { op: "add_record_ref", linked_record_id: existing.record_id },
              ],
            },
          },
        ],
      });
    const response = await route.fetch();
    expect(response.status()).toBe(links.length === 1 ? 409 : 200);
    await route.fulfill({ response });
  });
  await begin(page, f.incident, f.source.record_id);
  const conflictedLink = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${f.source.record_id}`) &&
      response.status() === 409,
  );
  await submit(page);
  await conflictedLink;
  await expect.poll(() => creations.length).toBe(1);
  const retained = await recovery(page);
  await expect.poll(() => links.length).toBe(1);
  expect(creations).toHaveLength(1);
  const target = creations[0]?.data.row;
  if (!target) throw new Error("Missing creation receipt");
  const targetBefore = await fetchFullRecordHistory(page, target.record_id);
  await expect(
    page.getByRole("button", { name: "Discard local draft", exact: true }),
  ).toBeVisible();
  await page
    .getByRole("button", { name: "Discard local draft", exact: true })
    .click();
  await expect(
    page.getByRole("region", { name: "Recovery navigation", exact: true }),
  ).toBeHidden();
  await recovery(page);
  await expect(retained).toContainText(
    "Evidence created; Timeline link incomplete.",
  );
  await retained
    .getByRole("button", { name: "Review original Timeline link", exact: true })
    .click();
  await retained
    .getByRole("button", { name: "Link created Evidence", exact: true })
    .click();
  await expect.poll(() => links.length).toBe(2);
  expect(creations).toHaveLength(1);
  await verifyLinked(
    page,
    f.incident,
    f.source.record_id,
    target.record_id,
    f.source.row_version + 2,
  );
  const sourceAfter = await fetchFullRecordHistory(page, f.source.record_id);
  expect(sourceAfter.items).toHaveLength(f.before.items.length + 4);
  expect(
    new Set(sourceAfter.items.map((item) => item.change_set_id)).size,
  ).toBe(new Set(f.before.items.map((item) => item.change_set_id)).size + 2);
  const targetAfter = await fetchFullRecordHistory(page, target.record_id);
  expect(targetAfter.items).toHaveLength(targetBefore.items.length + 1);
  expect(
    targetAfter.items.filter((item) =>
      item.diff_summary.units.some(
        (unit) =>
          unit.kind === "evidence_association" && unit.operation === "add",
      ),
    ),
  ).toHaveLength(1);
  expect(targetAfter.items.slice(1)).toEqual(targetBefore.items);
  const current = (
    await queryViewRows(page, f.incident, timelineViewSchemaId)
  ).find((row) => row.record_id === f.source.record_id);
  expect(
    JSON.stringify(current?.cells["timeline.attached_evidence_ids"]?.value),
  ).toContain(existing.record_id);
});

test("Timeline Evidence lost creation response replays exactly and recovers the original record", async ({
  page,
}) => {
  const f = await fixture(page),
    bodies: string[] = [],
    receipts: CreateViewRowResponse[] = [];
  await page.route(`**/views/${evidenceViewSchemaId}/rows`, async (route) => {
    bodies.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    receipts.push(await response.json());
    if (bodies.length === 1) await route.abort("failed");
    else await route.fulfill({ response });
  });
  await begin(page, f.incident, f.source.record_id);
  await submit(page);
  await expect.poll(() => receipts.length).toBe(1);
  const targetId = receipts[0]?.data.row.record_id;
  if (!targetId) throw new Error("Missing committed target");
  const before = await fetchFullRecordHistory(page, targetId);
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  const retained = await recovery(page);
  await retained
    .getByRole("button", { name: "Recover Evidence creation", exact: true })
    .click();
  await expect.poll(() => receipts.length).toBe(2);
  expect(bodies[1]).toBe(bodies[0]);
  expect(receipts[1]?.data).toEqual(receipts[0]?.data);
  expect((await fetchFullRecordHistory(page, targetId)).items).toEqual(
    before.items,
  );
  await retained
    .getByRole("button", { name: "Review original Timeline link", exact: true })
    .click();
  await retained
    .getByRole("button", { name: "Link created Evidence", exact: true })
    .click();
  await verifyLinked(
    page,
    f.incident,
    f.source.record_id,
    targetId,
    f.source.row_version + 1,
  );
  expect(
    await queryViewRows(page, f.incident, evidenceViewSchemaId),
  ).toHaveLength(1);
});

test("Timeline Evidence lost link response replays without duplicate relationship history revision or capture effects", async ({
  page,
}) => {
  const f = await fixture(page),
    bodies: string[] = [],
    receipts: CreateViewRowResponse[] = [];
  await page.route(`**/api/v1/records/${f.source.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    bodies.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    receipts.push(await response.json());
    if (bodies.length === 1) await route.abort("failed");
    else await route.fulfill({ response });
  });
  await begin(page, f.incident, f.source.record_id);
  await submit(page);
  await expect.poll(() => receipts.length).toBe(1);
  const evidence = await queryViewRows(page, f.incident, evidenceViewSchemaId);
  const target = evidence[0];
  if (!target) throw new Error("Missing Evidence");
  const before = await fetchFullRecordHistory(page, f.source.record_id),
    targetBefore = await fetchFullRecordHistory(page, target.record_id);
  const retained = await recovery(page);
  await retained
    .getByRole("button", { name: "Recover Timeline link", exact: true })
    .click();
  await expect.poll(() => receipts.length).toBe(2);
  expect(bodies[1]).toBe(bodies[0]);
  expect(receipts[1]?.data).toEqual(receipts[0]?.data);
  await verifyLinked(
    page,
    f.incident,
    f.source.record_id,
    target.record_id,
    f.source.row_version + 1,
  );
  expect(
    (await fetchFullRecordHistory(page, f.source.record_id)).items,
  ).toEqual(before.items);
  expect((await fetchFullRecordHistory(page, target.record_id)).items).toEqual(
    targetBefore.items,
  );
  expect(
    await queryViewRows(page, f.incident, evidenceViewSchemaId),
  ).toHaveLength(1);
});

test("Timeline Evidence navigation and source edits between commits require renewed review of the original row", async ({
  page,
}) => {
  const f = await fixture(page);
  const other = await createViewRow(page, f.incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("other"),
    "timeline.activity_synopsis_text": "Other selected row",
  });
  let release = () => {};
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  let committed: ViewRow | undefined;
  let links = 0;
  page.on("request", (request) => {
    if (
      request.method() === "PATCH" &&
      request.url().endsWith(`/records/${f.source.record_id}`)
    )
      links++;
  });
  await page.route(`**/views/${evidenceViewSchemaId}/rows`, async (route) => {
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    committed = ((await response.json()) as CreateViewRowResponse).data.row;
    await gate;
    await route.fulfill({ response });
  });
  await begin(page, f.incident, f.source.record_id);
  await submit(page);
  await expect.poll(() => committed?.record_id).toBeTruthy();
  await openTimelineInspector(page, other.record_id);
  await patchRecord(page, f.source.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: f.source.row_version,
    client_txn_id: uniqueTxn("source-edit"),
    changes: [
      {
        field_key: "timeline.activity_synopsis_text",
        value: "Remote source edit retained",
      },
    ],
  });
  release();
  const retained = await recovery(page);
  await expect(retained).toContainText(
    "Evidence created; Timeline link incomplete.",
  );
  expect(links).toBe(0);
  await retained
    .getByRole("button", { name: "Review original Timeline link", exact: true })
    .click();
  await retained
    .getByRole("button", { name: "Link created Evidence", exact: true })
    .click();
  if (!committed) throw new Error("Missing committed Evidence");
  await verifyLinked(
    page,
    f.incident,
    f.source.record_id,
    committed.record_id,
    f.source.row_version + 2,
  );
  const current = (
    await queryViewRows(page, f.incident, timelineViewSchemaId)
  ).find((row) => row.record_id === f.source.record_id);
  expect(current?.cells["timeline.activity_synopsis_text"]?.value).toBe(
    "Remote source edit retained",
  );
  const untouched = (
    await queryViewRows(page, f.incident, timelineViewSchemaId)
  ).find((row) => row.record_id === other.record_id);
  expect(untouched?.row_version).toBe(other.row_version);
  await expect(
    page.getByTestId(
      rowCellTestId(other.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toBeVisible();
});

test("Timeline Evidence accepted writes survive failed projections and recover through reads only", async ({
  page,
}) => {
  const f = await fixture(page);
  let failedReads = false,
    writes = 0;
  page.on("request", (request) => {
    if (
      (request.method() === "PATCH" &&
        request.url().endsWith(`/records/${f.source.record_id}`)) ||
      request.url().endsWith(`/views/${evidenceViewSchemaId}/rows`)
    )
      writes++;
  });
  await page.route(`**/views/*/query`, async (route) => {
    if (failedReads) await route.abort("failed");
    else await route.continue();
  });
  await page.route(`**/api/v1/records/${f.source.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    failedReads = true;
    await route.fulfill({ response });
  });
  await begin(page, f.incident, f.source.record_id);
  const linkedResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/records/${f.source.record_id}`),
  );
  await submit(page);
  expect((await linkedResponse).ok()).toBe(true);
  const retained = await recovery(page);
  await expect(retained).toContainText("Projection refresh is incomplete");
  expect(writes).toBe(2);
  failedReads = false;
  await retained
    .getByRole("button", { name: "Refresh Evidence result", exact: true })
    .click();
  await expect(
    retained.getByText("Timeline link: saved; views refreshed.", {
      exact: true,
    }),
  ).toBeVisible();
  expect(writes).toBe(2);
  const target = (
    await queryViewRows(page, f.incident, evidenceViewSchemaId)
  )[0];
  if (!target) throw new Error("Missing Evidence");
  await verifyLinked(
    page,
    f.incident,
    f.source.record_id,
    target.record_id,
    f.source.row_version + 1,
  );
});

test("Timeline Evidence retains partial success when its source is deleted or superseded or its target becomes unavailable", async ({
  page,
}) => {
  for (const unavailable of [
    "deleted_source",
    "superseded_source",
    "deleted_target",
  ] as const) {
    const f = await fixture(page);
    let release = () => {},
      committed: ViewRow | undefined,
      links = 0;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    const path = `**/incidents/${f.incident}/views/${evidenceViewSchemaId}/rows`;
    await page.route(path, async (route) => {
      const response = await route.fetch();
      expect(response.ok()).toBe(true);
      committed = ((await response.json()) as CreateViewRowResponse).data.row;
      await gate;
      await route.fulfill({ response });
    });
    page.on("request", (request) => {
      if (
        request.method() === "PATCH" &&
        request.url().endsWith(`/records/${f.source.record_id}`)
      )
        links++;
    });
    await begin(page, f.incident, f.source.record_id);
    await submit(page);
    await expect.poll(() => committed?.record_id).toBeTruthy();
    if (!committed) throw new Error("Missing committed Evidence");
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
    const affected = unavailable === "deleted_target" ? committed : f.source;
    const result = await publicHttpOperation({
      operationID:
        unavailable === "superseded_source"
          ? "supersedeRecord"
          : "deleteRecord",
      request: atJsonOrigin(page.request, apiBase),
      pathParameters: { record_id: affected.record_id },
      headers: await csrfHeaders(page),
      body: {
        base_row_version: affected.row_version,
        client_txn_id: uniqueTxn("unavailable"),
        reason: "Isolated recovery fixture",
      },
    });
    expect(result.ok).toBe(true);
    release();
    const retained = await recovery(page);
    await expect(retained).toContainText(
      "Evidence created; Timeline link incomplete.",
    );
    await retained
      .getByRole("button", {
        name: "Review original Timeline link",
        exact: true,
      })
      .click();
    await expect(retained).toContainText(/unavailable|superseded/);
    await expect(
      retained.getByRole("button", {
        name: "Link created Evidence",
        exact: true,
      }),
    ).toHaveCount(0);
    expect(links).toBe(0);
    expect(
      (await fetchFullRecordHistory(page, committed.record_id)).items.some(
        (item) => item.operation === "create",
      ),
    ).toBe(true);
    if (unavailable !== "deleted_target")
      expect(
        (await queryViewRows(page, f.incident, evidenceViewSchemaId)).some(
          (row) => row.record_id === committed?.record_id,
        ),
      ).toBe(true);
    await page.unroute(path);
  }
});

test("Timeline Evidence rechecks a real role loss between commits and preserves the accepted record", async ({
  page,
  browser,
  sessionTracker,
}) => {
  const f = await fixture(page),
    email = uniqueEmail("tre-editor"),
    password = "TimelineEvidenceFixture!234";
  const user = await createIncidentMemberUser(page, f.incident, {
    email,
    display_name: "Evidence author",
    initial_password: password,
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const editor = await openIncidentAsTrackedUser(browser, sessionTracker, {
    createdBy: "Timeline Evidence role recovery",
    email,
    password,
    incidentId: f.incident,
    userId: user.user_id,
    purpose: "Verify authority between Evidence creation and Timeline linking",
  });
  let release = () => {},
    committed: ViewRow | undefined,
    links = 0;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  await editor.route(`**/views/${evidenceViewSchemaId}/rows`, async (route) => {
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    committed = ((await response.json()) as CreateViewRowResponse).data.row;
    await gate;
    await route.fulfill({ response });
  });
  editor.on("request", (request) => {
    if (request.method() === "PATCH") links++;
  });
  await begin(editor, f.incident, f.source.record_id);
  await submit(editor);
  await expect.poll(() => committed?.record_id).toBeTruthy();
  const loss = await publicHttpOperation({
    operationID: "patchIncidentMembership",
    request: atJsonOrigin(page.request, apiBase),
    headers: await csrfHeaders(page),
    pathParameters: { incident_id: f.incident, user_id: user.user_id },
    body: { base_membership_version: 1, role: "viewer" },
  });
  expect(loss.ok).toBe(true);
  release();
  const retained = await recovery(editor);
  await expect(retained).toContainText(
    "Evidence created; Timeline link incomplete.",
  );
  await expect(
    retained.getByRole("button", {
      name: "Review original Timeline link",
      exact: true,
    }),
  ).toBeDisabled();
  expect(links).toBe(0);
  expect(
    (await queryViewRows(page, f.incident, evidenceViewSchemaId)).some(
      (row) => row.record_id === committed?.record_id,
    ),
  ).toBe(true);
  expect(
    (await fetchFullRecordHistory(page, f.source.record_id)).items,
  ).toEqual(f.before.items);
});
