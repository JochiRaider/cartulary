import {
  assertGridFocusContinuity,
  scrollGridTargetIntoView,
} from "@cartulary/test-utils/grid";
import {
  dataTestIdSelector,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  gridScrollportSelector,
  gridShellTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  rowInspectorFieldTestId,
  saveStateTestId,
  timelineInspectorTestId,
  timelineRowVersionTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import type { Page, Response } from "@playwright/test";

import { expect } from "@playwright/test";
import { timelineViewSchemaId } from "../contracts/workbookSurfaces";
import { queryViewRows, type ViewApiRow } from "./query";

type ViewRow = ViewApiRow;

type TimelineMutationEnvelope = {
  data: {
    change_set_id: string;
    row: ViewRow;
  };
};

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
  await anchor.evaluate((element) => {
    const gridCell = element.closest<HTMLElement>('[role="gridcell"]');
    if (gridCell === null)
      throw new Error("Expected Timeline action anchor cell");
    gridCell.focus();
  });
  await page.keyboard.press("Shift+F10");
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

export async function expectTimelineMutationContinuity(
  page: Page,
  recordId: string,
  minimumRowVersion: number,
  preservedScroll: { left: number; top: number },
  options: {
    requireExactHorizontalScroll?: boolean;
    requireExactVerticalScroll?: boolean;
  } = {},
) {
  await expect
    .poll(() => new URL(page.url()).searchParams.get("surface"))
    .toBeNull();
  await expect
    .poll(
      async () => {
        const rendered = await page
          .getByTestId(timelineRowVersionTestId(recordId))
          .textContent();
        const rowVersion = Number(rendered);
        return Number.isSafeInteger(rowVersion) ? rowVersion : -1;
      },
      {
        message: [
          "Timeline mutation projection did not reach its response version",
          `record_id=${recordId}`,
          `minimum_row_version=${minimumRowVersion}`,
        ].join("\n"),
      },
    )
    .toBeGreaterThanOrEqual(minimumRowVersion);
  const focusTestId = rowCellTestId(
    recordId,
    "timeline.activity_synopsis_text",
  );
  try {
    await assertGridFocusContinuity({
      allowContainingGridCell: true,
      focusTestId,
      page,
      preservedScroll,
      requireExactHorizontalScroll:
        options.requireExactHorizontalScroll ?? false,
      requireExactVerticalScroll: options.requireExactVerticalScroll ?? false,
      surface: timelineViewSchemaId,
    });
  } catch (error) {
    const diagnostic = await timelineContinuityDiagnosticSnapshot(
      page,
      recordId,
      minimumRowVersion,
    );
    throw new Error(
      `${error instanceof Error ? error.message : String(error)}\n${diagnostic}`,
      { cause: error },
    );
  }
}

async function timelineContinuityDiagnosticSnapshot(
  page: Page,
  recordId: string,
  minimumRowVersion: number,
) {
  const snapshot = await page.evaluate(
    ({ gridSelector, scrollportSelector, targetSelector }) => {
      const active = document.activeElement;
      const grid = document
        .querySelector(gridSelector)
        ?.querySelector<HTMLElement>(scrollportSelector);
      return {
        activeElement: {
          role: active?.getAttribute("role") ?? null,
          tag: active?.tagName.toLowerCase() ?? null,
          testId: active?.getAttribute("data-testid") ?? null,
        },
        mountedRowIds: Array.from(
          document.querySelectorAll<HTMLElement>("[data-grid-record-id]"),
        )
          .map((element) => element.dataset.gridRecordId)
          .filter(
            (candidate): candidate is string =>
              candidate !== undefined && candidate !== "",
          )
          .filter((candidate, index, all) => all.indexOf(candidate) === index),
        scroll: {
          clientHeight: grid?.clientHeight ?? null,
          clientWidth: grid?.clientWidth ?? null,
          left: grid?.scrollLeft ?? null,
          scrollHeight: grid?.scrollHeight ?? null,
          scrollWidth: grid?.scrollWidth ?? null,
          top: grid?.scrollTop ?? null,
        },
        targetPresent: document.querySelector(targetSelector) !== null,
      };
    },
    {
      gridSelector: dataTestIdSelector(gridShellTestId(timelineViewSchemaId)),
      scrollportSelector: gridScrollportSelector(),
      targetSelector: dataTestIdSelector(
        rowCellTestId(recordId, "timeline.activity_synopsis_text"),
      ),
    },
  );
  return [
    `record_id=${recordId}`,
    `minimum_row_version=${minimumRowVersion}`,
    `target_present=${snapshot.targetPresent}`,
    `active_element=${JSON.stringify(snapshot.activeElement)}`,
    `mounted_row_ids=${JSON.stringify(snapshot.mountedRowIds)}`,
    `scroll_geometry=${JSON.stringify(snapshot.scroll)}`,
  ].join("\n");
}

function findRow(rows: ViewRow[], recordId: string) {
  const row = rows.find((candidate) => candidate.record_id === recordId);
  if (!row) {
    throw new Error(`missing row ${recordId}`);
  }
  return row;
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
