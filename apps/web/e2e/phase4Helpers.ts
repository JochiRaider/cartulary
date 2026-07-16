import {
  assertGridFocusContinuity,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils";
import {
  currentIncidentRoleTestId,
  entityInspectButtonTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridRowTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  relationshipChipTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  saveStateTestId,
  timelineCollectionInputTestId,
  timelineInspectorTestId,
  timelinePreviewRowTestId,
  timelineRowVersionTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
} from "@cartulary/ui-contracts";
import type { Page, Response } from "@playwright/test";

import { expect } from "./fixtures";
import { createViewRow, queryViewRows, uniqueTxn } from "./helpers";

export const timelineViewSchemaId = "cartulary.view.timeline.v2";
export const hostsViewSchemaId = "cartulary.view.hosts.v1";
export const identitiesViewSchemaId = "cartulary.view.identities.v1";
export const assessmentsViewSchemaId = "cartulary.view.assessments.v1";
export const partiesViewSchemaId = "cartulary.view.parties.v1";
export const evidenceViewSchemaId = "cartulary.view.evidence.v1";
export const notesViewSchemaId = "cartulary.view.notes.v1";
export const indicatorsViewSchemaId = "cartulary.view.indicators.v1";
export const taskRequestsViewSchemaId = "cartulary.view.task_requests.v1";
export const decisionsViewSchemaId = "cartulary.view.decisions.v1";
export const commLogViewSchemaId = "cartulary.view.comm_log.v1";
export const handoffViewSchemaId = "cartulary.view.handoff.v1";
export const statusReviewViewSchemaId = "cartulary.view.status_review.v1";
export const lessonViewSchemaId = "cartulary.view.lesson.v1";
export const hostRefsFieldKey = "timeline.host_refs";
export const identityRefsFieldKey = "timeline.identity_refs";

export type ViewRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
};

type TimelineMutationEnvelope = {
  data: {
    change_set_id: string;
    row: ViewRow;
  };
};

type TimelinePatchRequestPayload = {
  base_row_version?: unknown;
};

type MergeEnvelope = {
  data: {
    survivor_record_id: string;
    loser_record_id: string;
    merged_into_record_id: string;
    merge_summary: {
      record_type: string;
      repointed_mention_resolution_count: number;
      repointed_link_count: number;
    };
  };
};

type CollectionItem = Record<string, unknown>;

export const timelineFixtureBaseOccurredAt = "2026-04-10T10:00:00.000Z";

type TimelineFillerOptions = {
  occurredAtStart?: string;
  occurredAtStepMinutes?: number;
};

export function timelineFixtureOccurredAt(
  offsetMinutes: number,
  baseOccurredAt = timelineFixtureBaseOccurredAt,
) {
  const baseMs = Date.parse(baseOccurredAt);
  if (!Number.isFinite(baseMs)) {
    throw new Error(
      `invalid timeline fixture base timestamp: ${baseOccurredAt}`,
    );
  }
  return new Date(baseMs + offsetMinutes * 60_000).toISOString();
}

export async function createTimelineFillers(
  page: Page,
  incidentId: string,
  prefix: string,
  count: number,
  options: TimelineFillerOptions = {},
) {
  const occurredAtStartMs =
    options.occurredAtStart === undefined
      ? null
      : Date.parse(options.occurredAtStart);
  if (occurredAtStartMs !== null && !Number.isFinite(occurredAtStartMs)) {
    throw new Error(
      `invalid timeline filler start timestamp: ${options.occurredAtStart}`,
    );
  }
  const occurredAtStepMs = (options.occurredAtStepMinutes ?? 1) * 60_000;

  for (let index = 1; index <= count; index += 1) {
    const payload: Record<string, unknown> = {
      client_txn_id: uniqueTxn(`${prefix}-${index}`),
      "timeline.activity_synopsis_text": `${prefix} ${index}`,
    };
    if (occurredAtStartMs !== null) {
      payload["timeline.activity_utc_text"] = new Date(
        occurredAtStartMs + (index - 1) * occurredAtStepMs,
      ).toISOString();
    }
    await createViewRow(page, incidentId, timelineViewSchemaId, payload);
  }
}

export async function createAssessmentViaUI(
  page: Page,
  options: {
    assessedAt: string;
    confidenceBand: string;
    rationale: string;
    state: string;
    supportRecordIds: string[];
  },
) {
  await page.getByTestId("assessment-create-state").selectOption(options.state);
  await page
    .getByTestId("assessment-create-confidence-band")
    .selectOption(options.confidenceBand);
  await page.getByTestId("assessment-create-rationale").fill(options.rationale);
  await page
    .getByTestId("assessment-create-assessed-at")
    .fill(options.assessedAt);
  if (options.supportRecordIds.length > 0) {
    await expect(
      page.getByTestId("assessment-create-support-refs").locator("option"),
    ).toHaveCount(options.supportRecordIds.length);
    await page
      .getByTestId("assessment-create-support-refs")
      .selectOption(options.supportRecordIds);
  }

  const responsePromise = waitForAssessmentCreate(page);
  await page.getByTestId("assessment-create-submit").click();
  const envelope = await readTimelineMutation(await responsePromise);
  await expect(
    page.getByTestId(
      gridRowTestId(assessmentsViewSchemaId, envelope.data.row.record_id),
    ),
  ).toBeVisible();
  return envelope.data.row;
}

export function waitForAssessmentCreate(page: Page) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/views/${assessmentsViewSchemaId}/rows`),
  );
}

export async function expectAssessmentGridOrder(
  page: Page,
  expected: string[],
) {
  const grid = page.getByTestId(gridShellTestId(assessmentsViewSchemaId));
  await expect
    .poll(async () =>
      grid
        .locator(gridSavedRowsSelector())
        .evaluateAll((rows) =>
          rows.map((row) => row.getAttribute("data-grid-record-id") ?? ""),
        ),
    )
    .toEqual(expected);
}

export function collectionActionsPayload(rawTexts: string[]) {
  return {
    kind: "collection_actions_v1",
    actions: rawTexts.map((rawText) => ({
      op: "add_token",
      raw_text: rawText,
    })),
  };
}

export function aliasCollectionActionsPayload(aliasTexts: string[]) {
  return {
    kind: "collection_actions_v1",
    actions: aliasTexts.map((aliasText) => ({
      op: "add_alias",
      alias_text: aliasText,
    })),
  };
}

export function resolvedRefPayload(rawText: string, resolvedRecordId: string) {
  return {
    kind: "collection_actions_v1",
    actions: [
      {
        op: "add_resolved_ref",
        raw_text: rawText,
        resolved_record_id: resolvedRecordId,
      },
    ],
  };
}

export async function seedHostMentionStateFixture(
  page: Page,
  incidentId: string,
  options: {
    displayPrefix: string;
    hostnamePrefix: string;
    occurredAt: {
      auto: string;
      dismissed: string;
      manual: string;
      resolved: string;
      unresolved: string;
    };
    rawTextPrefix: string;
    summary: {
      auto: string;
      dismissed: string;
      manual: string;
      resolved: string;
      unresolved: string;
    };
    txnPrefix: string;
  },
) {
  const resolvedTarget = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-resolved-target`),
      "host.display_name": `${options.displayPrefix} Resolved Target`,
      "host.hostname": `${options.hostnamePrefix}-resolved-target.example.test`,
    },
  )) as ViewRow;
  const manualTarget = (await createViewRow(
    page,
    incidentId,
    hostsViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-manual-target`),
      "host.display_name": `${options.displayPrefix} Manual Target`,
      "host.hostname": `${options.hostnamePrefix}-manual-target.example.test`,
    },
  )) as ViewRow;
  await createViewRow(page, incidentId, hostsViewSchemaId, {
    client_txn_id: uniqueTxn(`${options.txnPrefix}-auto-target`),
    "host.display_name": `${options.displayPrefix} Auto Target`,
    "host.hostname": `${options.hostnamePrefix}-auto-target.example.test`,
    "host.aliases": aliasCollectionActionsPayload([
      `${options.rawTextPrefix} Auto Alias`,
    ]),
  });

  const unresolvedRawText = `${options.rawTextPrefix} Unresolved?`;
  const resolvedRawText = `${options.rawTextPrefix} Resolved Raw`;
  const manualRawText = `${options.rawTextPrefix} Manual Raw`;
  const autoRawText = `${options.rawTextPrefix} Auto Alias`;
  const dismissedRawText = `${options.rawTextPrefix} Dismissed Raw`;
  const unresolvedRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-unresolved-row`),
      "timeline.activity_utc_text": options.occurredAt.unresolved,
      "timeline.activity_synopsis_text": options.summary.unresolved,
      [hostRefsFieldKey]: collectionActionsPayload([unresolvedRawText]),
    },
  )) as ViewRow;
  const resolvedRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-resolved-row`),
      "timeline.activity_utc_text": options.occurredAt.resolved,
      "timeline.activity_synopsis_text": options.summary.resolved,
      [hostRefsFieldKey]: resolvedRefPayload(
        resolvedRawText,
        resolvedTarget.record_id,
      ),
    },
  )) as ViewRow;
  const manualRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-manual-row`),
      "timeline.activity_utc_text": options.occurredAt.manual,
      "timeline.activity_synopsis_text": options.summary.manual,
      [hostRefsFieldKey]: collectionActionsPayload([manualRawText]),
    },
  )) as ViewRow;
  const autoRow = (await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn(`${options.txnPrefix}-auto-row`),
    "timeline.activity_utc_text": options.occurredAt.auto,
    "timeline.activity_synopsis_text": options.summary.auto,
  })) as ViewRow;
  const dismissedRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn(`${options.txnPrefix}-dismissed-row`),
      "timeline.activity_utc_text": options.occurredAt.dismissed,
      "timeline.activity_synopsis_text": options.summary.dismissed,
      [hostRefsFieldKey]: collectionActionsPayload([dismissedRawText]),
    },
  )) as ViewRow;

  return {
    autoRawText,
    autoRow,
    dismissedMention: requireItemByRawText(
      collectionItems(dismissedRow, hostRefsFieldKey),
      dismissedRawText,
    ),
    dismissedRawText,
    dismissedRow,
    manualMention: requireItemByRawText(
      collectionItems(manualRow, hostRefsFieldKey),
      manualRawText,
    ),
    manualRawText,
    manualRow,
    manualTarget,
    resolvedMention: requireItemByRawText(
      collectionItems(resolvedRow, hostRefsFieldKey),
      resolvedRawText,
    ),
    resolvedRawText,
    resolvedRow,
    resolvedTarget,
    unresolvedMention: requireItemByRawText(
      collectionItems(unresolvedRow, hostRefsFieldKey),
      unresolvedRawText,
    ),
    unresolvedRawText,
    unresolvedRow,
  };
}

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByLabel(
    "Account and application navigation",
  );
  await accountMenuTrigger.click();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toHaveText(
    roleText,
  );
  await accountMenuTrigger.click();
}

export async function exerciseEntityMerge(
  page: Page,
  options: {
    dependentRow: ViewRow;
    entityType: "host" | "identity";
    expectAdminRole?: boolean;
    fieldKey: string;
    incidentId: string;
    loser: ViewRow;
    loserLabel: string;
    mergeReason: string;
    rawText: string;
    recordType: string;
    resolvedLabel: string;
    survivor: ViewRow;
    survivorLabel: string;
    viewSchemaId: string;
  },
) {
  await page.goto(
    `/?incident_id=${options.incidentId}&view_schema_id=${encodeURIComponent(
      options.viewSchemaId,
    )}`,
  );
  await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
  if (options.expectAdminRole === true) {
    await expectCurrentIncidentRole(page, "Current incident role: admin");
  }
  await expect(
    page.getByTestId(
      gridRowTestId(options.viewSchemaId, options.survivor.record_id),
    ),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      gridRowTestId(options.viewSchemaId, options.loser.record_id),
    ),
  ).toBeVisible();

  const inspectButtonTestId = entityInspectButtonTestId(
    options.entityType,
    options.survivor.record_id,
  );
  await scrollGridTargetIntoView({
    page,
    surface: options.viewSchemaId,
    targetTestId: inspectButtonTestId,
  });
  await page.getByTestId(inspectButtonTestId).click();
  const inspectorTestId = entityInspectorLocalTestId(options.entityType);
  await expect(page.getByTestId(inspectorTestId)).toContainText(
    options.survivorLabel,
  );
  await expect(page.getByTestId("merge-start")).toBeVisible();
  await page.getByTestId("merge-start").click();
  await page
    .getByTestId("merge-loser-record")
    .selectOption(options.loser.record_id);
  await page.getByTestId("merge-reason").fill(options.mergeReason);
  await expect(page.getByTestId("merge-plan")).toContainText(
    `Survivor ${options.survivorLabel}`,
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    `loser ${options.loserLabel}`,
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    options.survivor.record_id,
  );
  await expect(page.getByTestId("merge-plan")).toContainText(
    options.loser.record_id,
  );

  const mergeResponsePromise = waitForMergeResponse(
    page,
    options.survivor.record_id,
  );
  await page.getByTestId("merge-confirm").click();
  const mergeEnvelope = await readMergeEnvelope(await mergeResponsePromise);

  await expect(
    page.getByTestId(
      gridRowTestId(options.viewSchemaId, options.loser.record_id),
    ),
  ).toHaveCount(0);
  await expect(page.getByTestId(inspectorTestId)).toContainText(
    options.survivorLabel,
  );
  await expect(page.getByTestId("merge-message")).toContainText(
    `Merged ${options.loserLabel} into ${options.survivorLabel}`,
  );
  await expect(
    page.getByTestId(timelinePreviewRowTestId(options.dependentRow.record_id)),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(timelinePreviewRowTestId(options.dependentRow.record_id))
      .getByLabel(`Resolved ${options.resolvedLabel}`),
  ).toBeVisible();

  const entityRowsAfter = (await queryViewRows(
    page,
    options.incidentId,
    options.viewSchemaId,
  )) as ViewRow[];
  const timelineRowsAfter = (await queryViewRows(
    page,
    options.incidentId,
    timelineViewSchemaId,
  )) as ViewRow[];
  const dependentRowAfter = findRow(
    timelineRowsAfter,
    options.dependentRow.record_id,
  );
  const dependentItemAfter = requireItemByRawText(
    collectionItems(dependentRowAfter, options.fieldKey),
    options.rawText,
  );

  expect(mergeEnvelope.data.survivor_record_id).toBe(
    options.survivor.record_id,
  );
  expect(mergeEnvelope.data.loser_record_id).toBe(options.loser.record_id);
  expect(mergeEnvelope.data.merged_into_record_id).toBe(
    options.survivor.record_id,
  );
  expect(mergeEnvelope.data.merge_summary.record_type).toBe(options.recordType);
  expect(
    mergeEnvelope.data.merge_summary.repointed_mention_resolution_count,
  ).toBeGreaterThan(0);
  expect(mergeEnvelope.data.merge_summary.repointed_link_count).toBeGreaterThan(
    0,
  );
  expect(
    entityRowsAfter.some((row) => row.record_id === options.survivor.record_id),
  ).toBeTruthy();
  expect(
    entityRowsAfter.some((row) => row.record_id === options.loser.record_id),
  ).toBeFalsy();
  expect(String(dependentItemAfter.item_kind)).toBe("resolved_ref");
  expect(String(dependentItemAfter.raw_text)).toBe(options.rawText);
  expect(String(dependentItemAfter.resolved_record_id)).toBe(
    options.survivor.record_id,
  );

  return {
    dependentItemAfter,
    dependentRowAfter,
    entityRowsAfter,
    mergeEnvelope,
    timelineRowsAfter,
  };
}

function entityInspectorLocalTestId(entityType: "host" | "identity") {
  return entityType === "host" ? "host-inspector" : "identity-inspector";
}

export function collectionItems(
  row: ViewRow | Record<string, unknown>,
  fieldKey: string,
) {
  const cells = (row as { cells: Record<string, { value: unknown }> }).cells;
  const cellValue = cells[fieldKey]?.value;
  if (
    !cellValue ||
    typeof cellValue !== "object" ||
    Array.isArray(cellValue) ||
    !("items" in cellValue)
  ) {
    return [] as CollectionItem[];
  }
  const items = (cellValue as { items?: unknown }).items;
  if (!Array.isArray(items)) {
    return [] as CollectionItem[];
  }
  return items.filter(
    (item): item is CollectionItem =>
      typeof item === "object" && item !== null && !Array.isArray(item),
  );
}

export function findRow(rows: ViewRow[], recordId: string) {
  const row = rows.find((candidate) => candidate.record_id === recordId);
  if (!row) {
    throw new Error(`missing row ${recordId}`);
  }
  return row;
}

export function requireItemByRawText(items: CollectionItem[], rawText: string) {
  const item = items.find((candidate) => candidate.raw_text === rawText);
  if (!item) {
    throw new Error(`missing collection item raw_text=${rawText}`);
  }
  return item;
}

export async function waitForSaveState(
  page: Page,
  value: "Saved" | "Syncing" | "Conflict",
) {
  await expect(page.getByTestId(saveStateTestId())).toHaveText(value);
}

export async function expectNoPendingQueueAuthPause(
  page: Page,
  context: string,
) {
  const snapshot = await pendingQueueDiagnosticSnapshot(page);
  if (!snapshot.authPaused) {
    return;
  }
  throw new Error(formatPendingQueueAuthPause(context, snapshot));
}

export async function ensureTimelineGridTargetVisible(
  page: Page,
  targetTestId: string,
) {
  return scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId,
  });
}

export async function addRelationshipTokenViaUI(
  page: Page,
  recordId: string,
  draftKey: "hostRefs" | "identityRefs",
  rawText: string,
  options: {
    onPatchRequest?: (payload: TimelinePatchRequestPayload) => void;
    requireVisibleChip?: boolean;
  } = {},
) {
  const fieldKey =
    draftKey === "identityRefs" ? identityRefsFieldKey : hostRefsFieldKey;
  const inputTestId = timelineCollectionInputTestId(recordId, fieldKey);
  await openTimelineInspector(page, recordId);
  const input = page.getByTestId(inputTestId);
  await input.focus();
  await expect(input).toBeVisible();
  const responsePromise = waitForTimelinePatch(page, recordId);
  await input.fill(rawText);
  await input.press("Enter");
  const response = await responsePromise;
  const requestPayload = readRequestPayload(response);
  const envelope = await readTimelineMutation(response);
  options.onPatchRequest?.(requestPayload);
  const item = requireItemByRawText(
    collectionItems(envelope.data.row, fieldKey),
    rawText,
  );
  if (options.requireVisibleChip === true) {
    await expect(
      page
        .getByTestId(relationshipItemsTestId(recordId, fieldKey))
        .getByTestId(relationshipChipTestId(String(item.item_ref))),
    ).toBeVisible();
  }
  await expect
    .poll(
      async () => ({
        inputValue: await input.inputValue().catch((error: unknown) => {
          return `<<failed to read input value: ${String(error)}>>`;
        }),
        pendingQueueNoticeCount: await page
          .getByTestId(pendingQueueNoticeTestId())
          .count(),
        renderedRowVersion: await page
          .getByTestId(timelineRowVersionTestId(recordId))
          .textContent()
          .catch((error: unknown) => {
            return `<<failed to read row version: ${String(error)}>>`;
          }),
      }),
      {
        message: [
          "relationship token commit did not converge",
          `record_id=${recordId}`,
          `draft_key=${draftKey}`,
          `raw_text=${JSON.stringify(rawText)}`,
          `request_payload=${JSON.stringify(requestPayload)}`,
          `response_row_version=${envelope.data.row.row_version}`,
        ].join("\n"),
      },
    )
    .toEqual({
      inputValue: "",
      pendingQueueNoticeCount: 0,
      renderedRowVersion: String(envelope.data.row.row_version),
    });
  return envelope;
}

export async function commitInspectorScalarEdit(
  page: Page,
  recordId: string,
  fieldKey: string,
  value: string,
) {
  const input = page.getByTestId(rowInspectorFieldTestId(recordId, fieldKey));
  await expect(input).toBeVisible();
  const responsePromise = waitForTimelinePatch(page, recordId);
  await input.fill(value);
  await input.press("Tab");
  const response = await responsePromise;
  const envelope = await readTimelineMutation(response);
  await expect(page.getByTestId(timelineRowVersionTestId(recordId))).toHaveText(
    String(envelope.data.row.row_version),
  );
  await waitForSaveState(page, "Saved");
  return envelope;
}

export async function openTimelineInspector(page: Page, recordId: string) {
  const closeInspector = page.getByTestId(
    workbookInspectorCloseButtonTestId(timelineViewSchemaId),
  );
  if ((await closeInspector.count()) > 0) {
    await closeInspector.click();
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  }
  const inspectButtonTestId = rowInspectButtonTestId(recordId);
  await clickTimelineRowAction(page, recordId, inspectButtonTestId);
  await expect(page.getByTestId(timelineInspectorTestId())).toBeVisible();
}

export async function openTimelineRowActions(page: Page, recordId: string) {
  const closeInspector = page.getByTestId(
    workbookInspectorCloseButtonTestId(timelineViewSchemaId),
  );
  if ((await closeInspector.count()) > 0) {
    await closeInspector.click();
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
  }
  const anchorTestId = rowCellTestId(
    recordId,
    "timeline.activity_synopsis_text",
  );
  await ensureTimelineGridTargetVisible(page, anchorTestId);
  const anchor = page.getByTestId(anchorTestId);
  await anchor.focus();
  await anchor.press("Shift+F10");
}

export async function clickTimelineRowAction(
  page: Page,
  recordId: string,
  actionTestId: string,
) {
  await openTimelineRowActions(page, recordId);
  await page.getByTestId(actionTestId).click();
}

export function waitForTimelinePatch(page: Page, recordId: string) {
  const responsePromise = page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${recordId}`),
  );
  return Promise.race([
    responsePromise,
    waitForPendingQueueAuthPause(page, `timeline PATCH ${recordId}`),
  ]);
}

export function waitForMergeResponse(page: Page, survivorRecordId: string) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${survivorRecordId}/merge`),
  );
}

export async function readTimelineMutation(response: Response) {
  if (!response.ok()) {
    const request = response.request();
    const responseBody = await response.text().catch((error: unknown) => {
      return `<<failed to read response body: ${String(error)}>>`;
    });
    const requestBody = request.postData() ?? "";
    expect(
      response.ok(),
      [
        `timeline mutation failed with HTTP ${response.status()}`,
        `method=${request.method()}`,
        `url=${response.url()}`,
        `request_body=${truncateDiagnostic(requestBody)}`,
        `response_body=${truncateDiagnostic(responseBody)}`,
      ].join("\n"),
    ).toBeTruthy();
  }
  return (await response.json()) as TimelineMutationEnvelope;
}

function readRequestPayload(response: Response): TimelinePatchRequestPayload {
  const postData = response.request().postData();
  if (!postData) {
    return {};
  }
  try {
    const parsed = JSON.parse(postData) as unknown;
    if (!parsed || typeof parsed !== "object" || Array.isArray(parsed)) {
      return {};
    }
    return parsed as TimelinePatchRequestPayload;
  } catch {
    return {};
  }
}

async function waitForPendingQueueAuthPause(
  page: Page,
  context: string,
): Promise<Response> {
  const notice = page
    .getByTestId(pendingQueueNoticeTestId())
    .filter({ hasText: "Authentication is required before queued edits" });
  return notice
    .waitFor({ state: "visible", timeout: 60_000 })
    .then(async () => {
      throw new Error(
        formatPendingQueueAuthPause(
          context,
          await pendingQueueDiagnosticSnapshot(page),
        ),
      );
    })
    .catch((error: unknown) => {
      if (String(error).includes("Timeout")) {
        return new Promise<Response>(() => undefined);
      }
      throw error;
    });
}

async function pendingQueueDiagnosticSnapshot(page: Page) {
  const notice = page.getByTestId(pendingQueueNoticeTestId());
  const noticeCount = await notice.count().catch(() => -1);
  const noticeText =
    noticeCount > 0
      ? await notice
          .first()
          .textContent()
          .then((value) => value ?? "")
          .catch((error: unknown) => {
            return `<<failed to read pending queue notice: ${String(error)}>>`;
          })
      : "";
  const pendingUnits =
    noticeCount > 0
      ? await page
          .getByTestId(pendingQueueCountTestId())
          .textContent()
          .then((value) => value ?? "")
          .catch((error: unknown) => {
            return `<<failed to read pending queue count: ${String(error)}>>`;
          })
      : "";
  const saveState = await page
    .getByTestId(saveStateTestId())
    .textContent()
    .then((value) => value ?? "")
    .catch((error: unknown) => {
      return `<<failed to read save state: ${String(error)}>>`;
    });
  return {
    authPaused: noticeText.includes(
      "Authentication is required before queued edits",
    ),
    noticeCount,
    noticeText,
    pendingUnits,
    saveState,
    url: page.url(),
  };
}

function formatPendingQueueAuthPause(
  context: string,
  snapshot: Awaited<ReturnType<typeof pendingQueueDiagnosticSnapshot>>,
) {
  return [
    "pending queue entered auth-paused state before timeline mutation completed",
    `context=${context}`,
    `url=${snapshot.url}`,
    `save_state=${JSON.stringify(snapshot.saveState)}`,
    `pending_queue_notice_count=${snapshot.noticeCount}`,
    `pending_queue_notice=${truncateDiagnostic(snapshot.noticeText)}`,
    `pending_queue_units=${truncateDiagnostic(snapshot.pendingUnits)}`,
  ].join("\n");
}

function truncateDiagnostic(value: string) {
  const limit = 4000;
  if (value.length <= limit) {
    return value;
  }
  return `${value.slice(0, limit)}...<truncated ${value.length - limit} chars>`;
}

export async function readMergeEnvelope(response: Response) {
  expect(response.ok()).toBeTruthy();
  return (await response.json()) as MergeEnvelope;
}

export async function expectTimelineContinuity(
  page: Page,
  recordId: string,
  preservedScroll: { left: number; top: number },
  options: {
    requireExactHorizontalScroll?: boolean;
    requireExactVerticalScroll?: boolean;
  } = {},
) {
  await expect
    .poll(() => new URL(page.url()).searchParams.get("surface"))
    .toBeNull();
  const focusTestId = rowCellTestId(
    recordId,
    "timeline.activity_synopsis_text",
  );
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: focusTestId,
  });
  await page.getByTestId(focusTestId).scrollIntoViewIfNeeded();
  await assertGridFocusContinuity({
    focusTestId,
    page,
    preservedScroll,
    requireExactHorizontalScroll: options.requireExactHorizontalScroll ?? false,
    requireExactVerticalScroll: options.requireExactVerticalScroll ?? false,
    surface: timelineViewSchemaId,
  });
}

export async function waitForViewRow(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  recordId: string,
) {
  await expect
    .poll(async () => {
      const rows = (await queryViewRows(page, incidentId, viewSchemaId)) as
        | ViewRow[]
        | Array<Record<string, unknown>>;
      return rows.some(
        (candidate) =>
          typeof candidate === "object" &&
          candidate !== null &&
          "record_id" in candidate &&
          candidate.record_id === recordId,
      );
    })
    .toBe(true);
  const rows = (await queryViewRows(
    page,
    incidentId,
    viewSchemaId,
  )) as ViewRow[];
  return findRow(rows, recordId);
}

export async function waitForViewRowByCell(
  page: Page,
  incidentId: string,
  viewSchemaId: string,
  fieldKey: string,
  value: string,
): Promise<ViewRow> {
  let match: ViewRow | null = null;
  await expect
    .poll(async () => {
      const rows = (await queryViewRows(
        page,
        incidentId,
        viewSchemaId,
      )) as ViewRow[];
      match = rows.find((row) => row.cells[fieldKey]?.value === value) ?? null;
      return match !== null;
    })
    .toBe(true);
  return requirePolledViewRow(match, viewSchemaId, fieldKey, value);
}

function requirePolledViewRow(
  row: ViewRow | null,
  viewSchemaId: string,
  fieldKey: string,
  value: string,
) {
  if (row === null) {
    throw new Error(`missing ${viewSchemaId} row where ${fieldKey}=${value}`);
  }
  return row;
}

export async function editGenericCell(
  page: Page,
  viewSchemaId: string,
  recordId: string,
  fieldKey: string,
  value: string | string[],
) {
  await page.getByTestId(workbookInspectorToggleTestId(viewSchemaId)).click();
  await page
    .getByTestId(genericEditRecordSelectTestId(viewSchemaId))
    .selectOption(recordId);
  await page
    .getByTestId(genericEditFieldSelectTestId(viewSchemaId))
    .selectOption(fieldKey);
  const input = page.getByTestId(genericEditValueTestId(viewSchemaId));
  const tagName = await input.evaluate((element) => element.tagName);
  if (tagName === "SELECT") {
    await input.selectOption(value);
  } else {
    await input.fill(Array.isArray(value) ? value.join("\n") : value);
  }
  await submitGenericEditAndWait(page, viewSchemaId, recordId);
}

export function waitForRecordPatch(
  page: Page,
  recordId: string,
): Promise<Response> {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${recordId}`),
  );
}

export async function submitGenericEditAndWait(
  page: Page,
  viewSchemaId: string,
  recordId: string,
): Promise<Response> {
  const patchResponse = waitForRecordPatch(page, recordId);
  await page.getByTestId(genericEditSubmitTestId(viewSchemaId)).click();
  const response = await patchResponse;
  expect(response.ok()).toBeTruthy();
  return response;
}
