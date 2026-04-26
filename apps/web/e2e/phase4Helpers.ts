import {
  assertGridFocusContinuity,
  gridSavedRowsSelector,
  gridShellTestId,
  rowInspectButtonTestId,
} from "@cartulary/test-utils";
import type { Page, Response } from "@playwright/test";

import { expect } from "./fixtures";
import { createViewRow, queryViewRows, uniqueTxn } from "./helpers";

export const timelineViewSchemaId = "cartulary.view.timeline.v1";
export const hostsViewSchemaId = "cartulary.view.hosts.v1";
export const identitiesViewSchemaId = "cartulary.view.identities.v1";
export const assessmentsViewSchemaId = "cartulary.view.assessments.v1";
export const notesViewSchemaId = "cartulary.view.notes.v1";
export const taskRequestsViewSchemaId = "cartulary.view.task_requests.v1";
export const decisionsViewSchemaId = "cartulary.view.decisions.v1";
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

export async function createTimelineFillers(
  page: Page,
  incidentId: string,
  prefix: string,
  count: number,
) {
  for (let index = 1; index <= count; index += 1) {
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn(`${prefix}-${index}`),
      "timeline.summary": `${prefix} ${index}`,
    });
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

export async function addRelationshipTokenViaUI(
  page: Page,
  recordId: string,
  draftKey: "hostRefs" | "identityRefs",
  rawText: string,
) {
  const responsePromise = waitForTimelinePatch(page, recordId);
  await page.getByTestId(`row-${recordId}-${draftKey}-input`).fill(rawText);
  await page.getByTestId(`row-${recordId}-${draftKey}-input`).press("Enter");
  await waitForSaveState(page, "Saved");
  return readTimelineMutation(await responsePromise);
}

export async function openTimelineInspector(page: Page, recordId: string) {
  await page.getByTestId(rowInspectButtonTestId(recordId)).click();
  await expect(page.getByTestId("timeline-inspector")).toContainText(recordId);
}

export function waitForTimelinePatch(page: Page, recordId: string) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "PATCH" &&
      response.url().endsWith(`/api/v1/records/${recordId}`),
  );
}

export function waitForMergeResponse(page: Page, survivorRecordId: string) {
  return page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response.url().endsWith(`/api/v1/records/${survivorRecordId}/merge`),
  );
}

export async function readTimelineMutation(response: Response) {
  expect(response.ok()).toBeTruthy();
  return (await response.json()) as TimelineMutationEnvelope;
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
    .toBe("timeline");
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
) {
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
  return match as ViewRow;
}

export async function editGenericCell(
  page: Page,
  viewSchemaId: string,
  recordId: string,
  fieldKey: string,
  value: string,
) {
  await page
    .getByTestId(`generic-edit-record-${viewSchemaId}`)
    .selectOption(recordId);
  await page
    .getByTestId(`generic-edit-field-${viewSchemaId}`)
    .selectOption(fieldKey);
  await page.getByTestId(`generic-edit-value-${viewSchemaId}`).fill(value);
  await page.getByTestId(`generic-edit-submit-${viewSchemaId}`).click();
}
