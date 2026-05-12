import { gridShellTestId, rowCellTestId } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createViewRow,
  csrfHeaders,
  patchTimelineRecord,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  collectionActionsPayload,
  collectionItems,
  evidenceViewSchemaId,
  findRow,
  hostRefsFieldKey,
  hostsViewSchemaId,
  requireItemByRawText,
  timelineViewSchemaId,
  type ViewRow,
} from "./phase4Helpers";
import { installIncidentSocketMonitor } from "./phase6Harness";

type HistoryItem = {
  actor_user_id: string;
  committed_at: string;
  operation: string;
  diff_summary: { summary: string; units: Array<Record<string, unknown>> };
  change_set_id: string;
  reversible: boolean;
  available_rollback_actions: Array<
    "history_entry" | "change_set" | "row_restore"
  >;
  history_entry_ref?: string;
  revision_no?: number;
};

type HistoryData = {
  record_id: string;
  row_version: number;
  deleted: boolean;
  items: HistoryItem[];
};

test("E-7-01 opens row history from the workbook surface with legal rollback actions", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E701"),
    "Phase 7 E-7-01 row history",
  );
  const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e701-row"),
    "timeline.summary": "E-7-01 before",
  })) as unknown as ViewRow;
  await patchTimelineRecord(page, row.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.row_version,
    client_txn_id: uniqueTxn("e701-update"),
    changes: [{ field_key: "timeline.summary", value: "E-7-01 after" }],
  });
  const history = await fetchRecordHistory(page, row.record_id);
  const visibleItemIndex = history.items.findIndex(
    (item) => item.available_rollback_actions.length > 0,
  );
  expect(visibleItemIndex).toBeGreaterThanOrEqual(0);
  const visibleItem = history.items[visibleItemIndex] as HistoryItem;

  await openTimelineSurface(page, incidentId);
  await page.getByTestId(`row-history-open-${row.record_id}`).click();
  await expect(page.getByTestId("row-history-panel")).toContainText(
    visibleItem.actor_user_id,
  );
  await expect(page.getByTestId("row-history-panel")).toContainText(
    visibleItem.operation,
  );
  await expect(page.getByTestId("row-history-panel")).toContainText(
    visibleItem.diff_summary.summary,
  );
  await expect(page.getByTestId("row-history-panel")).toContainText(
    new Date(visibleItem.committed_at).toISOString(),
  );

  for (const action of visibleItem.available_rollback_actions) {
    await expect(
      page.getByTestId(`row-history-action-${visibleItemIndex}-${action}`),
    ).toBeVisible();
  }
});

test("E-7-02 rolls back one attached-evidence mutation without reverting later unrelated edits", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E702"),
    "Phase 7 E-7-02 rollback evidence link",
  );
  const evidence = (await createViewRow(
    page,
    incidentId,
    evidenceViewSchemaId,
    {
      client_txn_id: uniqueTxn("e702-evidence"),
      "evidence.title": "E-7-02 linked evidence",
      "evidence.collector_party_text": "Reviewer",
    },
  )) as unknown as ViewRow;
  const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e702-row"),
    "timeline.summary": "E-7-02 original summary",
  })) as unknown as ViewRow;
  const linkedRow = (await patchTimelineRecord(page, row.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.row_version,
    client_txn_id: uniqueTxn("e702-link-evidence"),
    changes: [
      {
        field_key: "timeline.attached_evidence_ids",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "add_record_ref",
              linked_record_id: evidence.record_id,
            },
          ],
        },
      },
    ],
  })) as unknown as ViewRow;
  const currentRow = (await patchTimelineRecord(page, row.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: linkedRow.row_version,
    client_txn_id: uniqueTxn("e702-unrelated"),
    changes: [
      { field_key: "timeline.summary", value: "E-7-02 unrelated later edit" },
    ],
  })) as unknown as ViewRow;

  const history = await fetchRecordHistory(page, row.record_id);
  const rollbackIndex = history.items.findIndex(
    (item) =>
      item.available_rollback_actions.includes("history_entry") &&
      item.diff_summary.units.some(
        (unit) => unit.target_kind === "record_link",
      ),
  );
  expect(rollbackIndex).toBeGreaterThanOrEqual(0);
  const rollbackItem = history.items[rollbackIndex] as HistoryItem;

  const listener = await page.context().newPage();
  try {
    const socketMonitor = installIncidentSocketMonitor(listener, incidentId);
    await openTimelineSurface(listener, incidentId);
    await socketMonitor.waitForMessage("hello_ack");
    await openTimelineSurface(page, incidentId);
    await page.getByTestId(`row-history-open-${row.record_id}`).click();
    await expect(
      page.getByTestId(`row-history-action-${rollbackIndex}-history_entry`),
    ).toBeVisible();
    const rollbackResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().endsWith(`/api/v1/records/${row.record_id}/rollback`),
    );
    await page
      .getByTestId(`row-history-action-${rollbackIndex}-history_entry`)
      .click();
    const response = await rollbackResponse;
    expect(response.ok()).toBeTruthy();
    const rollbackBody = JSON.parse(response.request().postData() ?? "{}");
    expect(rollbackBody.target).toEqual({
      kind: "history_entry",
      history_entry_ref: rollbackItem.history_entry_ref,
    });
    expect(Object.keys(rollbackBody.target).sort()).toEqual([
      "history_entry_ref",
      "kind",
    ]);
    await socketMonitor.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload?.record_id === row.record_id &&
        Array.isArray(message.payload?.affected_views) &&
        message.payload.affected_views.some(
          (view: { change_kind?: string; view_schema_id?: string }) =>
            view.view_schema_id === timelineViewSchemaId &&
            view.change_kind === "invalidate",
        ),
    });

    const rows = (await queryViewRows(
      page,
      incidentId,
      timelineViewSchemaId,
    )) as unknown as ViewRow[];
    const afterRollback = findRow(rows, row.record_id);
    expect(
      collectionItems(afterRollback, "timeline.attached_evidence_ids").some(
        (item) => item.linked_record_id === evidence.record_id,
      ),
    ).toBe(false);
    expect(afterRollback.cells["timeline.summary"]?.value).toBe(
      "E-7-02 unrelated later edit",
    );
    expect(currentRow.cells["timeline.summary"]?.value).toBe(
      "E-7-02 unrelated later edit",
    );
  } finally {
    await listener.close();
  }
});

test("E-7-03 soft-deletes and restores a row with tombstone concurrency", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E703"),
    "Phase 7 E-7-03 delete restore",
  );
  const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e703-row"),
    "timeline.summary": "E-7-03 delete restore row",
  })) as unknown as ViewRow;

  const listener = await page.context().newPage();
  const socketMonitor = installIncidentSocketMonitor(listener, incidentId);
  await openTimelineSurface(listener, incidentId);
  await socketMonitor.waitForMessage("hello_ack");
  await openTimelineSurface(page, incidentId);
  await page.getByTestId(`row-history-open-${row.record_id}`).click();
  await expect(page.getByTestId("row-history-delete")).toBeVisible();

  const deleteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "DELETE" &&
      response.url().endsWith(`/api/v1/records/${row.record_id}`),
  );
  await page.getByTestId("row-history-delete").click();
  const deleted = await deleteResponse;
  expect(deleted.ok()).toBeTruthy();
  const deleteBody = JSON.parse(deleted.request().postData() ?? "{}");
  expect(deleteBody.base_row_version).toBe(row.row_version);
  const removeMessage = await socketMonitor.waitForMessage("record_changed", {
    matches: (message) =>
      message.payload?.record_id === row.record_id &&
      Array.isArray(message.payload?.affected_views) &&
      message.payload.affected_views.some(
        (view: { change_kind?: string }) => view.change_kind === "remove",
      ),
  });
  const tombstoneRowVersion = Number(removeMessage.payload.row_version);
  expect(tombstoneRowVersion).toBeGreaterThan(row.row_version);

  await expect(page.getByTestId("row-history-restore")).toBeVisible();
  const restoreResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${row.record_id}/restore`),
  );
  await page.getByTestId("row-history-restore").click();
  const restored = await restoreResponse;
  expect(restored.ok()).toBeTruthy();
  const restoreBody = JSON.parse(restored.request().postData() ?? "{}");
  expect(restoreBody.base_row_version).toBe(tombstoneRowVersion);
  await socketMonitor.waitForMessage("record_changed", {
    matches: (message) =>
      message.payload?.record_id === row.record_id &&
      Array.isArray(message.payload?.affected_views) &&
      message.payload.affected_views.some(
        (view: { change_kind?: string }) => view.change_kind === "invalidate",
      ),
  });

  await openTimelineSurface(page, incidentId);
  await expect(
    page.getByTestId(rowCellTestId(row.record_id, "summary")),
  ).toHaveValue("E-7-03 delete restore row");
  await listener.close();
});

test("E-7-04 whole-row restore appends a new attributed revision", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E704"),
    "Phase 7 E-7-04 whole row restore",
  );
  const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("e704-row"),
    "timeline.summary": "E-7-04 original",
  })) as unknown as ViewRow;
  const snapshot = (await patchTimelineRecord(page, row.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: row.row_version,
    client_txn_id: uniqueTxn("e704-snapshot"),
    changes: [
      {
        field_key: "timeline.summary",
        value: "E-7-04 historical snapshot",
      },
    ],
  })) as unknown as ViewRow;
  await patchTimelineRecord(page, row.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: snapshot.row_version,
    client_txn_id: uniqueTxn("e704-current"),
    changes: [
      {
        field_key: "timeline.summary",
        value: "E-7-04 current value",
      },
    ],
  });
  const historyBefore = await fetchRecordHistory(page, row.record_id);
  const restoreIndex = historyBefore.items.findIndex(
    (item) =>
      item.available_rollback_actions.includes("row_restore") &&
      item.revision_no === snapshot.row_version,
  );
  expect(restoreIndex).toBeGreaterThanOrEqual(0);
  const restoreItem = historyBefore.items[restoreIndex] as HistoryItem;

  await openTimelineSurface(page, incidentId);
  await page.getByTestId(`row-history-open-${row.record_id}`).click();
  await expect(
    page.getByTestId(`row-history-action-${restoreIndex}-row_restore`),
  ).toBeVisible();
  const restoreResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${row.record_id}/rollback`),
  );
  await page
    .getByTestId(`row-history-action-${restoreIndex}-row_restore`)
    .click();
  const response = await restoreResponse;
  expect(response.ok()).toBeTruthy();
  const requestBody = JSON.parse(response.request().postData() ?? "{}");
  expect(requestBody.target).toEqual({
    kind: "row_restore",
    restore_to_revision_no: restoreItem.revision_no,
  });

  await openTimelineSurface(page, incidentId);
  await expect(
    page.getByTestId(rowCellTestId(row.record_id, "summary")),
  ).toHaveValue("E-7-04 historical snapshot");
  const historyAfter = await fetchRecordHistory(page, row.record_id);
  expect(historyAfter.items.length).toBeGreaterThan(historyBefore.items.length);
  expect(
    historyAfter.items.some(
      (item) => item.change_set_id === restoreItem.change_set_id,
    ),
  ).toBe(true);
  const appended = historyAfter.items.find(
    (item) =>
      item.operation === "row_restore" &&
      item.revision_no === historyAfter.row_version,
  ) as HistoryItem | undefined;
  expect(appended).toBeDefined();
  const appendedItem = appended as HistoryItem;
  expect(appendedItem.change_set_id).not.toBe(restoreItem.change_set_id);
  expect(appendedItem.operation).toBe("row_restore");
  expect(appendedItem.actor_user_id).not.toBe("");
  expect(appendedItem.committed_at).not.toBe("");
  expect(appendedItem.committed_at).toMatch(
    /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z$/,
  );
  expect(Number.isNaN(new Date(appendedItem.committed_at).getTime())).toBe(
    false,
  );
});

test("E-7-05 rolls back a merge change set from row history", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E705"),
    "Phase 7 E-7-05 merge rollback",
  );
  const survivor = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e705-survivor"),
    "host.display_name": "E-7-05 survivor",
    "host.hostname": "e705-survivor",
  })) as ViewRow;
  const loser = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("e705-loser"),
    "host.display_name": "E-7-05 loser",
    "host.hostname": "e705-loser",
    "host.fqdn": "e705-loser.example.test",
    "host.aliases": collectionActionsPayload(["E-7-05 loser alias"]),
  })) as ViewRow;
  const timeline = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e705-timeline"),
      "timeline.summary": "E-7-05 dependent timeline row",
      [hostRefsFieldKey]: collectionActionsPayload(["E-7-05 loser alias"]),
    },
  )) as ViewRow;
  const hostMention = requireItemByRawText(
    collectionItems(timeline, hostRefsFieldKey),
    "E-7-05 loser alias",
  );
  await patchTimelineRecord(page, timeline.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: timeline.row_version,
    client_txn_id: uniqueTxn("e705-resolve-host"),
    changes: [
      {
        field_key: hostRefsFieldKey,
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "resolve_item",
              item_ref: hostMention.item_ref,
              resolved_record_id: loser.record_id,
            },
          ],
        },
      },
    ],
  });

  const mergeResponse = await page.request.post(
    `${apiBase}/api/v1/records/${survivor.record_id}/merge`,
    {
      headers: await csrfHeaders(page),
      data: {
        loser_record_id: loser.record_id,
        survivor_base_row_version: survivor.row_version,
        loser_base_row_version: loser.row_version,
        client_txn_id: uniqueTxn("e705-merge"),
        reason: "Phase 7 E-7-05 browser merge",
      },
    },
  );
  expect(mergeResponse.ok()).toBeTruthy();
  const mergeData = (await mergeResponse.json()) as {
    data: { change_set_id: string; survivor_row_version: number };
  };

  const history = await fetchRecordHistory(page, timeline.record_id);
  const mergeIndex = history.items.findIndex(
    (item) =>
      item.change_set_id === mergeData.data.change_set_id &&
      item.available_rollback_actions.includes("change_set"),
  );
  expect(mergeIndex).toBeGreaterThanOrEqual(0);

  await openTimelineSurface(page, incidentId);
  await page.getByTestId(`row-history-open-${timeline.record_id}`).click();
  await expect(
    page.getByTestId(`row-history-action-${mergeIndex}-change_set`),
  ).toBeVisible();

  const rollbackResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${timeline.record_id}/rollback`),
  );
  await page.getByTestId(`row-history-action-${mergeIndex}-change_set`).click();
  const response = await rollbackResponse;
  expect(response.ok()).toBeTruthy();
  const rollbackBody = JSON.parse(response.request().postData() ?? "{}");
  expect(rollbackBody.target).toEqual({
    kind: "change_set",
    change_set_id: mergeData.data.change_set_id,
  });
  expect(rollbackBody.base_row_version).toBe(history.row_version);

  const rows = (await queryViewRows(
    page,
    incidentId,
    hostsViewSchemaId,
  )) as unknown as ViewRow[];
  const survivorAfter = findRow(rows, survivor.record_id);
  const loserAfter = findRow(rows, loser.record_id);
  expect(survivorAfter.row_version).toBeGreaterThan(
    mergeData.data.survivor_row_version,
  );
  expect(survivorAfter.cells["host.fqdn"]?.value ?? "").toBe("");
  expect(loserAfter.cells["host.fqdn"]?.value).toBe("e705-loser.example.test");
  await openHostSurface(page, incidentId);
  await expect(page.getByTestId(`host-row-${loser.record_id}`)).toBeVisible();
});

async function openTimelineSurface(page: Page, incidentId: string) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      timelineViewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(gridShellTestId("timeline"))).toBeVisible();
}

async function openHostSurface(page: Page, incidentId: string) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      hostsViewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(gridShellTestId("hosts"))).toBeVisible();
}

async function fetchRecordHistory(
  page: Page,
  recordId: string,
): Promise<HistoryData> {
  const response = await page.request.get(
    `${apiBase}/api/v1/records/${recordId}/history`,
    { headers: await csrfHeaders(page) },
  );
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: HistoryData }).data;
}
