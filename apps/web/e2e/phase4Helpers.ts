import {
  assertGridFocusContinuity,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils";
import {
  gridSavedRowsSelector,
  gridShellTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  relationshipItemsTestId,
  rowInspectButtonTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import type { Page, Response } from "@playwright/test";

import { expect } from "./fixtures";
import { createViewRow, queryViewRows, uniqueTxn } from "./helpers";

export const timelineViewSchemaId = "cartulary.view.timeline.v1";
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
      "timeline.summary": `${prefix} ${index}`,
    };
    if (occurredAtStartMs !== null) {
      payload["timeline.occurred_at"] = new Date(
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
    page.getByTestId(`assessment-row-${envelope.data.row.record_id}`),
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

export function sanitizeTestId(value: string) {
  return value.replace(/[^a-zA-Z0-9_-]+/gu, "-");
}

export async function waitForSaveState(
  page: Page,
  value: "Saved" | "Syncing" | "Conflict",
) {
  await expect(page.getByTestId("save-state")).toHaveText(value);
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
    surface: "timeline",
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
  } = {},
) {
  const inputTestId = `row-${recordId}-${draftKey}-input`;
  await ensureTimelineGridTargetVisible(page, inputTestId);
  const input = page.getByTestId(inputTestId);
  const responsePromise = waitForTimelinePatch(page, recordId);
  await input.fill(rawText);
  await input.press("Enter");
  const response = await responsePromise;
  const requestPayload = readRequestPayload(response);
  const envelope = await readTimelineMutation(response);
  options.onPatchRequest?.(requestPayload);
  const fieldKey =
    draftKey === "identityRefs" ? identityRefsFieldKey : hostRefsFieldKey;
  const item = requireItemByRawText(
    collectionItems(envelope.data.row, fieldKey),
    rawText,
  );
  await expect(
    page
      .getByTestId(relationshipItemsTestId(recordId, draftKey))
      .getByTestId(`chip-${sanitizeTestId(String(item.item_ref))}`),
  ).toBeVisible();
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
        saveState: await page
          .getByTestId("save-state")
          .textContent()
          .catch((error: unknown) => {
            return `<<failed to read save state: ${String(error)}>>`;
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
      saveState: "Saved",
    });
  return envelope;
}

export async function openTimelineInspector(page: Page, recordId: string) {
  const inspectButtonTestId = rowInspectButtonTestId(recordId);
  await ensureTimelineGridTargetVisible(page, inspectButtonTestId);
  await page.getByTestId(inspectButtonTestId).click();
  await expect(page.getByTestId("timeline-inspector")).toContainText(recordId);
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
    .getByTestId("save-state")
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
  await assertGridFocusContinuity({
    focusTestId: rowInspectButtonTestId(recordId),
    page,
    preservedScroll,
    requireExactHorizontalScroll: options.requireExactHorizontalScroll ?? false,
    requireExactVerticalScroll: options.requireExactVerticalScroll ?? false,
    surface: "timeline",
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
  await page
    .getByTestId(`generic-edit-record-${viewSchemaId}`)
    .selectOption(recordId);
  await page
    .getByTestId(`generic-edit-field-${viewSchemaId}`)
    .selectOption(fieldKey);
  const input = page.getByTestId(`generic-edit-value-${viewSchemaId}`);
  const tagName = await input.evaluate((element) => element.tagName);
  if (tagName === "SELECT") {
    await input.selectOption(value);
  } else {
    await input.fill(Array.isArray(value) ? value.join("\n") : value);
  }
  await page.getByTestId(`generic-edit-submit-${viewSchemaId}`).click();
}
