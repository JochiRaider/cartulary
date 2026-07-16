import {
  scrollGridCellIntoView,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils";
import {
  draftCellTestId,
  entityInspectButtonTestId,
  gridShellTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowHistoryActionTestId,
  rowHistoryDeleteButtonTestId,
  rowHistoryDestructiveConfirmButtonTestId,
  rowHistoryDestructiveConfirmPanelTestId,
  rowHistoryMessageTestId,
  rowHistoryOpenButtonTestId,
  rowHistoryPanelTestId,
  rowHistoryRestoreButtonTestId,
  rowHistoryRollbackCancelButtonTestId,
  rowHistoryRollbackConfirmButtonTestId,
  rowHistoryRollbackPreviewTestId,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  saveStateTestId,
  surfaceTabTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineMutationSubstrateReadyTestId,
  timelineScalarEditorTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createIncidentMemberUser,
  createSavedView,
  createViewRow,
  csrfHeaders,
  patchRecord,
  queryViewRows,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
  waitForCommittedRowSummary,
} from "./helpers";
import {
  clickTimelineRowAction,
  collectionActionsPayload,
  collectionItems,
  evidenceViewSchemaId,
  hostRefsFieldKey,
  hostsViewSchemaId,
  openTimelineInspector,
  requireItemByRawText,
  timelineViewSchemaId,
  type ViewRow,
} from "./phase4Helpers";

type HistoryItem = {
  actor_user_id: string;
  available_rollback_actions: Array<
    "history_entry" | "change_set" | "row_restore"
  >;
  change_set_id: string;
  committed_at: string;
  diff_summary: { summary: string; units: Array<Record<string, unknown>> };
  history_entry_ref?: string;
  history_item_ref: string;
  operation: string;
  reversible: boolean;
  revision_no?: number;
};

type HistoryData = {
  deleted: boolean;
  items: HistoryItem[];
  record_id: string;
  row_version: number;
};

type IncidentMembershipRecord = {
  membership_version: number;
  role: string;
  user_id: string;
};

function attachedEvidencePayload(recordId: string) {
  return {
    kind: "collection_actions_v1",
    actions: [
      {
        op: "add_record_ref",
        linked_record_id: recordId,
      },
    ],
  };
}

function historyActionTestId(
  item: HistoryItem,
  action: HistoryItem["available_rollback_actions"][number],
) {
  return rowHistoryActionTestId({
    action,
    historyItemRef: item.history_item_ref,
  });
}

function rollbackPreviewAnchor(
  item: HistoryItem,
  action: HistoryItem["available_rollback_actions"][number],
) {
  return {
    action,
    historyItemRef: item.history_item_ref,
  };
}

function requireHistoryEntryAction(history: HistoryData) {
  const item =
    history.items.find(
      (candidate) =>
        candidate.available_rollback_actions.includes("history_entry") &&
        typeof candidate.history_entry_ref === "string" &&
        candidate.history_entry_ref.length > 0,
    ) ?? null;
  if (item === null) {
    throw new Error("missing history_entry rollback item");
  }
  return item;
}

test("FE-I-P9-01 Verify history and rollback preview/action use public route contracts, preserve retained history, and render public error envelopes.", async ({
  page,
}) => {
  await disableWorkbookSockets(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEIP901"),
    "FE-I-P9-01 rollback public envelope",
  );
  const evidence = (await createViewRow(
    page,
    incidentId,
    evidenceViewSchemaId,
    {
      client_txn_id: uniqueTxn("feip901-evidence"),
      "evidence.collector_party_text": "FE-I-P9 collector",
      "evidence.title": "FE-I-P9 attached evidence",
    },
  )) as unknown as ViewRow;
  const row = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("feip901-row"),
    "timeline.activity_synopsis_text": "FE-I-P9 rollback row",
  })) as unknown as ViewRow;
  const linkedRow = (await patchRecord(page, row.record_id, {
    base_row_version: row.row_version,
    changes: [
      {
        action_payload: attachedEvidencePayload(evidence.record_id),
        field_key: "timeline.attached_evidence_ids",
      },
    ],
    client_txn_id: uniqueTxn("feip901-link"),
    view_schema_id: timelineViewSchemaId,
  })) as unknown as ViewRow;
  const history = await fetchRecordHistory(page, row.record_id);
  const rollbackItem = requireHistoryEntryAction(history);
  const rollbackAnchor = rollbackPreviewAnchor(rollbackItem, "history_entry");

  await openTimelineSurface(page, incidentId);
  await clickTimelineRowAction(
    page,
    row.record_id,
    rowHistoryOpenButtonTestId(row.record_id),
  );
  await expect(page.getByTestId(rowHistoryPanelTestId())).toBeVisible();
  await page
    .getByTestId(historyActionTestId(rollbackItem, "history_entry"))
    .click();
  await expect(
    page.getByTestId(rowHistoryRollbackPreviewTestId(rollbackAnchor)),
  ).toContainText(rollbackItem.history_item_ref);
  await page
    .getByTestId(rowHistoryRollbackCancelButtonTestId(rollbackAnchor))
    .click();
  await expect(
    page.getByTestId(rowHistoryRollbackPreviewTestId(rollbackAnchor)),
  ).toHaveCount(0);

  await page
    .getByTestId(historyActionTestId(rollbackItem, "history_entry"))
    .click();
  await patchRecord(page, row.record_id, {
    base_row_version: linkedRow.row_version,
    changes: [
      {
        field_key: "timeline.activity_synopsis_text",
        value: "FE-I-P9 stale server update",
      },
    ],
    client_txn_id: uniqueTxn("feip901-stale"),
    view_schema_id: timelineViewSchemaId,
  });

  const rollbackResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${row.record_id}/rollback`),
  );
  await page
    .getByTestId(rowHistoryRollbackConfirmButtonTestId(rollbackAnchor))
    .click();
  const response = await rollbackResponse;
  expect(response.status()).toBe(409);
  const responseBody = (await response.json()) as {
    error: { code: string };
  };
  expect(responseBody.error.code).toBe("row_version_conflict");
  const requestBody = JSON.parse(response.request().postData() ?? "{}");
  expect(requestBody.base_row_version).toBe(history.row_version);
  expect(requestBody.target).toEqual({
    history_entry_ref: rollbackItem.history_entry_ref,
    kind: "history_entry",
  });
  expect(String(requestBody.client_txn_id)).toMatch(/^timeline-client-/u);
  await expect(page.getByTestId(rowHistoryMessageTestId())).toContainText(
    "row_version_conflict",
  );
  await expect(
    page.getByTestId(historyActionTestId(rollbackItem, "history_entry")),
  ).toBeVisible();
});

test("FE-E-P9-01 Verify inspector Details, Relationships, Evidence, History, rollback, and destructive-action authorization through public browser routes.", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEEP901"),
    "FE-E-P9-01 inspector row-local actions",
  );
  await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("feep901-fallback"),
    "timeline.activity_synopsis_text": "FE-E-P9 fallback row",
  });
  const evidence = (await createViewRow(
    page,
    incidentId,
    evidenceViewSchemaId,
    {
      client_txn_id: uniqueTxn("feep901-evidence"),
      "evidence.collector_party_text": "FE-E-P9 collector",
      "evidence.title": "FE-E-P9 attached evidence",
    },
  )) as unknown as ViewRow;
  const target = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    [hostRefsFieldKey]: collectionActionsPayload(["FE-E-P9 stable host"]),
    client_txn_id: uniqueTxn("feep901-target"),
    "timeline.raw_activity_text": "FE-E-P9 detailed inspector body",
    "timeline.activity_synopsis_text": "FE-E-P9 selected row",
  })) as unknown as ViewRow;
  const linkedTarget = (await patchRecord(page, target.record_id, {
    base_row_version: target.row_version,
    changes: [
      {
        action_payload: attachedEvidencePayload(evidence.record_id),
        field_key: "timeline.attached_evidence_ids",
      },
    ],
    client_txn_id: uniqueTxn("feep901-link"),
    view_schema_id: timelineViewSchemaId,
  })) as unknown as ViewRow;
  const hostItem = requireItemByRawText(
    collectionItems(linkedTarget, hostRefsFieldKey),
    "FE-E-P9 stable host",
  );
  const attachedEvidenceItem =
    collectionItems(linkedTarget, "timeline.attached_evidence_ids").find(
      (item) => item.linked_record_id === evidence.record_id,
    ) ?? null;
  expect(String(attachedEvidenceItem?.item_ref ?? "")).not.toBe("");
  const history = await fetchRecordHistory(page, target.record_id);
  const rollbackItem = requireHistoryEntryAction(history);
  const rollbackAnchor = rollbackPreviewAnchor(rollbackItem, "history_entry");

  await openTimelineSurface(page, incidentId);
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: target.record_id,
    surface: timelineViewSchemaId,
  });
  await page
    .getByTestId(
      rowCellTestId(target.record_id, "timeline.activity_synopsis_text"),
    )
    .focus();
  await openTimelineInspector(page, target.record_id);
  for (const section of [
    "operational-text",
    "relationships",
    "evidence",
    "history",
  ] as const) {
    await expect(
      page.getByTestId(timelineInspectorSectionTestId(section)),
    ).toBeVisible();
  }
  await expect(
    page.getByTestId(
      timelineScalarEditorTestId({
        fieldKey: "timeline.raw_activity_text",
        recordId: target.record_id,
        surface: "inspector",
      }),
    ),
  ).toHaveValue("FE-E-P9 detailed inspector body");
  await expect(
    page
      .getByTestId(relationshipItemsTestId(target.record_id, hostRefsFieldKey))
      .getByTestId(relationshipChipTestId(String(hostItem.item_ref))),
  ).toBeVisible();
  await expect(
    page.getByTestId(timelineInspectorSectionTestId("evidence")),
  ).toContainText("Attached evidence count: 0");

  await clickTimelineRowAction(
    page,
    target.record_id,
    rowHistoryOpenButtonTestId(target.record_id),
  );
  await expect(
    page.getByTestId(historyActionTestId(rollbackItem, "history_entry")),
  ).toBeVisible();
  await page
    .getByTestId(historyActionTestId(rollbackItem, "history_entry"))
    .click();
  await expect(
    page.getByTestId(rowHistoryRollbackPreviewTestId(rollbackAnchor)),
  ).toContainText(rollbackItem.history_item_ref);
  await page
    .getByTestId(rowHistoryRollbackCancelButtonTestId(rollbackAnchor))
    .click();

  await page.getByTestId(rowHistoryDeleteButtonTestId()).click();
  await expect(
    page.getByTestId(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "delete" }),
    ),
  ).toContainText(target.record_id);
  const deleteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "DELETE" &&
      response.url().endsWith(`/api/v1/records/${target.record_id}`),
  );
  await page
    .getByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "delete" }),
    )
    .click();
  const deleted = await deleteResponse;
  expect(deleted.ok()).toBeTruthy();
  const deleteBody = JSON.parse(deleted.request().postData() ?? "{}");
  expect(deleteBody.base_row_version).toBe(linkedTarget.row_version);
  expect(String(deleteBody.client_txn_id)).toMatch(/^timeline-client-/u);
  const deletedEnvelope = (await deleted.json()) as {
    data: { row_version: number };
  };
  await expect(
    page.getByTestId(
      rowCellTestId(target.record_id, "timeline.activity_synopsis_text"),
    ),
  ).toHaveCount(0);
  await expect(page.getByTestId(rowHistoryPanelTestId())).toContainText(
    `Record ${target.record_id}`,
  );
  await expect(page.getByTestId(rowHistoryRestoreButtonTestId())).toBeVisible();
  const restoreResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${target.record_id}/restore`),
  );
  await page.getByTestId(rowHistoryRestoreButtonTestId()).click();
  await expect(
    page.getByTestId(
      rowHistoryDestructiveConfirmPanelTestId({ operation: "restore" }),
    ),
  ).toContainText(target.record_id);
  await page
    .getByTestId(
      rowHistoryDestructiveConfirmButtonTestId({ operation: "restore" }),
    )
    .click();
  const restored = await restoreResponse;
  expect(restored.ok()).toBeTruthy();
  const restoreBody = JSON.parse(restored.request().postData() ?? "{}");
  expect(restoreBody.base_row_version).toBe(deletedEnvelope.data.row_version);
  expect(String(restoreBody.client_txn_id)).toMatch(/^timeline-client-/u);

  const memberPassword = "Phase9Reviewer1!";
  const member = await createIncidentMemberUser(page, incidentId, {
    display_name: "FE-E-P9 reviewer",
    email: uniqueEmail("fe-e-p9-reviewer"),
    initial_password: memberPassword,
    role: "reviewer",
  });
  const memberContext = await browser.newContext();
  const memberPage = await memberContext.newPage();
  try {
    await disableWorkbookSockets(memberPage);
    await sessionTracker.loginTrackedUser(memberPage, {
      createdBy: "FE-E-P9-01",
      email: member.email,
      password: memberPassword,
      purpose: "row-local rollback authorization denial",
      userId: member.user_id,
    });
    const retainedHistory = await fetchRecordHistory(page, target.record_id);
    const retainedRollbackItem = requireHistoryEntryAction(retainedHistory);
    const retainedRollbackAnchor = rollbackPreviewAnchor(
      retainedRollbackItem,
      "history_entry",
    );
    await openTimelineSurface(memberPage, incidentId);
    await clickTimelineRowAction(
      memberPage,
      target.record_id,
      rowHistoryOpenButtonTestId(target.record_id),
    );
    await expect(
      memberPage.getByTestId(
        historyActionTestId(retainedRollbackItem, "history_entry"),
      ),
    ).toBeVisible();
    await memberPage
      .getByTestId(historyActionTestId(retainedRollbackItem, "history_entry"))
      .click();
    await expect(
      memberPage.getByTestId(
        rowHistoryRollbackPreviewTestId(retainedRollbackAnchor),
      ),
    ).toContainText(retainedRollbackItem.history_item_ref);

    const memberMembership = await loadIncidentMembership(
      page,
      incidentId,
      member.user_id,
    );
    await patchIncidentMembershipRole(page, incidentId, {
      baseMembershipVersion: memberMembership.membership_version,
      role: "editor",
      userId: member.user_id,
    });

    const rollbackResponse = memberPage.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().endsWith(`/api/v1/records/${target.record_id}/rollback`),
    );
    await memberPage
      .getByTestId(
        rowHistoryRollbackConfirmButtonTestId(retainedRollbackAnchor),
      )
      .click();
    const denied = await rollbackResponse;
    expect(denied.status()).toBe(403);
    const deniedBody = (await denied.json()) as {
      error: { code: string };
    };
    expect(deniedBody.error.code).toBe("authorization_denied");
    await expect(
      memberPage.getByTestId(rowHistoryMessageTestId()),
    ).toContainText("authorization_denied");
  } finally {
    await memberContext.close();
  }
});

test("FE-E-P9-02 Verify default-closed inspector state, no-row state, surface switch config changes, saved-view switch over the same view_schema_id keeps the same config, closed incident read-only behavior, server-denied action behavior, and Timeline create/edit/paste without opening the inspector.", async ({
  browser,
  page,
  sessionTracker,
}) => {
  await disableWorkbookSockets(page);
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEEP902"),
    "FE-E-P9-02 view-schema inspector config",
  );
  const timelineSavedView = await createSavedView(page, incidentId, {
    display_name: "FE-E-P9-02 Timeline saved view",
    scope: "shared",
    view_schema_id: timelineViewSchemaId,
  });
  const timelineSeed = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("feep902-seed"),
      "timeline.raw_activity_text": "FE-E-P9-02 seed raw details",
      "timeline.activity_synopsis_text": "FE-E-P9-02 seed summary",
    },
  )) as unknown as ViewRow;
  const hostSeed = (await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn("feep902-host"),
    "host.display_name": "FE-E-P9-02 host",
    "host.hostname": "fe-e-p9-02.example.test",
  })) as unknown as ViewRow;

  await openTimelineSurface(page, incidentId);
  await expect(
    page.getByTestId(timelineMutationSubstrateReadyTestId()),
  ).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  await expect(page.getByTestId(timelineInspectorTestId())).toContainText(
    "no_row_selected",
  );
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

  await openTimelineInspector(page, timelineSeed.record_id);
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
  const sameSurfaceSelector = page.getByTestId(
    savedViewSelectorTestId(timelineViewSchemaId),
  );
  await expect(
    sameSurfaceSelector.getByTestId(
      savedViewOptionTestId(
        timelineViewSchemaId,
        timelineSavedView.saved_view_id,
      ),
    ),
  ).toHaveAttribute("data-view-schema-id", timelineViewSchemaId);
  await sameSurfaceSelector.selectOption(timelineSavedView.saved_view_id);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await openTimelineInspector(page, timelineSeed.record_id);
  await expect(
    page.getByTestId(timelineInspectorSectionTestId("relationships")),
  ).toBeVisible();
  await page.reload();
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

  await openTimelineInspector(page, timelineSeed.record_id);
  await page.getByTestId(surfaceTabTestId(hostsViewSchemaId)).click();
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  await expect(
    page.getByTestId(gridShellTestId(hostsViewSchemaId)),
  ).toBeVisible();
  const hostInspectButtonTestId = entityInspectButtonTestId(
    "host",
    hostSeed.record_id,
  );
  await scrollGridTargetIntoView({
    page,
    surface: hostsViewSchemaId,
    targetTestId: hostInspectButtonTestId,
  });
  await page.getByTestId(hostInspectButtonTestId).click();
  await expect(page.getByTestId("host-inspector")).toContainText(
    "FE-E-P9-02 host",
  );
  await page.getByTestId(surfaceTabTestId(timelineViewSchemaId)).click();
  await expect(page.getByTestId("host-inspector")).toHaveCount(0);
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

  const draftSummaryTestId = draftCellTestId("timeline.activity_synopsis_text");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftSummaryTestId,
  });
  const draftSummary = page.getByTestId(draftSummaryTestId);
  await draftSummary.fill("FE-E-P9-02 hot path created");
  await draftSummary.press("Enter");
  const created = await waitForCommittedRowSummary(page, {
    expectedSummary: "FE-E-P9-02 hot path created",
    surface: timelineViewSchemaId,
    timeoutMs: 5_000,
  });
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

  const createdSummary = page.getByTestId(
    rowCellTestId(created.recordId, "timeline.activity_synopsis_text"),
  );
  await createdSummary.fill("FE-E-P9-02 hot path edited");
  await createdSummary.press("Tab");
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);

  await createdSummary.focus();
  const pasteResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .includes(
          `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/clipboard-paste`,
        ),
  );
  await createdSummary.evaluate((element) => {
    const data = new DataTransfer();
    data.setData(
      "text/plain",
      ["FE-E-P9-02 pasted first", "FE-E-P9-02 pasted second"].join("\n"),
    );
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  });
  await expect((await pasteResponse).ok()).toBeTruthy();
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  const timelineRows = await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  expect(
    timelineRows.filter((row) =>
      String(
        (row.cells as Record<string, { value: unknown }>)[
          "timeline.activity_synopsis_text"
        ]?.value ?? "",
      ).startsWith("FE-E-P9-02 pasted "),
    ),
  ).toHaveLength(2);

  const memberPassword = "Phase9InspectorViewer1!";
  const member = await createIncidentMemberUser(page, incidentId, {
    display_name: "FE-E-P9-02 viewer",
    email: uniqueEmail("fe-e-p9-02-viewer"),
    initial_password: memberPassword,
    role: "viewer",
  });
  const memberContext = await browser.newContext();
  const memberPage = await memberContext.newPage();
  try {
    await disableWorkbookSockets(memberPage);
    await sessionTracker.loginTrackedUser(memberPage, {
      createdBy: "FE-E-P9-02",
      email: member.email,
      password: memberPassword,
      purpose: "inspector-backed record action denial",
      userId: member.user_id,
    });
    const deniedDelete = await memberPage.request.delete(
      `${apiBase}/api/v1/records/${timelineSeed.record_id}`,
      {
        headers: await csrfHeaders(memberPage),
        data: {
          base_row_version: timelineSeed.row_version,
          client_txn_id: uniqueTxn("feep902-denied-delete"),
        },
      },
    );
    expect(deniedDelete.status()).toBe(403);
    const deniedBody = (await deniedDelete.json()) as {
      error: { code: string };
    };
    expect(deniedBody.error.code).toBe("authorization_denied");
  } finally {
    await memberContext.close();
  }
});

async function openTimelineSurface(page: Page, incidentId: string) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      timelineViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
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

async function loadIncidentMembership(
  page: Page,
  incidentId: string,
  userId: string,
) {
  const response = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
    { headers: await csrfHeaders(page) },
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { memberships: IncidentMembershipRecord[] };
  };
  const membership =
    body.data.memberships.find((candidate) => candidate.user_id === userId) ??
    null;
  if (membership === null) {
    throw new Error(`missing incident membership for ${userId}`);
  }
  return membership;
}

async function patchIncidentMembershipRole(
  page: Page,
  incidentId: string,
  options: {
    baseMembershipVersion: number;
    role: string;
    userId: string;
  },
) {
  const response = await page.request.patch(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${options.userId}`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_membership_version: options.baseMembershipVersion,
        role: options.role,
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

async function disableWorkbookSockets(page: Page) {
  await page.addInitScript(() => {
    class Phase9InspectorClosedWebSocket {
      static readonly CONNECTING = 0;
      static readonly OPEN = 1;
      static readonly CLOSING = 2;
      static readonly CLOSED = 3;

      readonly url: string;
      readyState = Phase9InspectorClosedWebSocket.CONNECTING;
      onclose: ((event: CloseEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onopen: ((event: Event) => void) | null = null;

      constructor(url: string | URL) {
        this.url = String(url);
      }

      close() {
        this.readyState = Phase9InspectorClosedWebSocket.CLOSED;
        this.onclose?.(new CloseEvent("close"));
      }

      send() {}
    }

    Object.defineProperty(window, "WebSocket", {
      configurable: true,
      value: Phase9InspectorClosedWebSocket,
    });
  });
}
