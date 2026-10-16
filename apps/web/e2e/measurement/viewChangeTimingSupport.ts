import type { CartularyAc044PredicateId } from "@cartulary/ui-contracts";
import { cartularyAc044PerformanceContract } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import {
  type MeasurementSample,
  resetWorkbookClientTiming,
} from "./timingSupport";

export function viewChangePredicate(id: CartularyAc044PredicateId) {
  const predicate = cartularyAc044PerformanceContract.predicates.find(
    (value) => value.predicate_id === id,
  );
  if (!predicate) throw new Error("Unknown view-change predicate");
  return predicate;
}

export type ViewChangePaintInput = {
  operation: "sort" | "filter" | "group";
  viewSchemaId: string;
  startMark: string;
  stopMark: string;
  completion: "first_useful_viewport" | "stable_viewport";
  expectedRows: readonly { recordId: string; minimumVersion: number }[];
  groupId: string | null;
  timeoutMs: number;
};

// Self-contained so the same predicate runs in the real browser and adversarial DOM fixtures.
export function observeViewChangePaintInBrowser(
  input: ViewChangePaintInput,
): Promise<{ acceptedAt: number; stoppedAt: number }> {
  return new Promise((resolve, reject) => {
    const deadline = performance.now() + input.timeoutMs;
    const expected = new Map(
      input.expectedRows.map((row) => [row.recordId, row.minimumVersion]),
    );
    let prior: { element: Element; id: string; version: number }[] = [];
    let priorRevision = -1;
    let frames = 0;
    let diagnostic = "not_observed";
    const tick = () => {
      const marks = performance.getEntriesByName(
        input.startMark,
        "mark",
      ) as PerformanceMark[];
      if (marks.length > 1) {
        reject(new Error("View-change action was accepted more than once"));
        return;
      }
      const start = marks[0];
      const frame = [
        ...document.querySelectorAll<HTMLElement>(
          "[data-workbook-query-revision]",
        ),
      ].find(
        (element) =>
          element.dataset.workbookViewSchemaId === input.viewSchemaId,
      );
      const grid = frame?.querySelector<HTMLElement>(
        '[role="grid"], [role="treegrid"]',
      );
      const revision = Number(frame?.dataset.workbookQueryRevision);
      const viewport = grid?.getBoundingClientRect();
      const clipped = viewport && {
        left: Math.max(0, viewport.left),
        right: Math.min(innerWidth, viewport.right),
        top: Math.max(0, viewport.top),
        bottom: Math.min(innerHeight, viewport.bottom),
      };
      const visible = (element: Element) => {
        if (!clipped || !element.isConnected) return false;
        const rect = element.getBoundingClientRect();
        for (
          let node: Element | null = element;
          node;
          node = node.parentElement
        ) {
          const style = getComputedStyle(node);
          if (
            node.hasAttribute("hidden") ||
            node.hasAttribute("inert") ||
            style.display === "none" ||
            style.visibility === "hidden" ||
            Number(style.opacity || 1) <= 0
          )
            return false;
        }
        return (
          rect.width > 0 &&
          rect.height > 0 &&
          rect.right > clipped.left &&
          rect.left < clipped.right &&
          rect.bottom > clipped.top &&
          rect.top < clipped.bottom
        );
      };
      const rows = Array.from(
        grid?.querySelectorAll<HTMLElement>(
          '[role="row"][data-grid-record-id]',
        ) ?? [],
      )
        .flatMap((element) => {
          // Semantic rows may use display:contents. Painted cells own geometry.
          const cell = [
            ...element.querySelectorAll<HTMLElement>(
              '[role="gridcell"], [role="rowheader"]',
            ),
          ].find(visible);
          if (!cell) return [];
          const { top, bottom } = cell.getBoundingClientRect();
          return [
            {
              element,
              id: element.dataset.gridRecordId ?? "",
              version: Number(element.dataset.gridRowVersion),
              top,
              bottom,
              height: bottom - top,
            },
          ];
        })
        .sort((a, b) => a.top - b.top);
      const group = grid?.querySelector<HTMLElement>(
        "[data-cartulary-grid-group-id]",
      );
      const groupReady =
        input.groupId === null
          ? !group
          : !!group &&
            visible(group) &&
            group.dataset.cartularyGridGroupId === input.groupId &&
            group.getAttribute("aria-expanded") === "true";
      const entry = grid?.matches('[tabindex="0"]')
        ? grid
        : grid?.querySelector<HTMLElement>('[tabindex="0"]');
      const keyboardReady =
        !!entry &&
        visible(entry) &&
        !entry.hasAttribute("disabled") &&
        entry.getAttribute("aria-disabled") !== "true";
      const bound =
        rows.length >= 2 &&
        rows.length <= input.expectedRows.length &&
        new Set(rows.map((row) => row.id)).size === rows.length &&
        rows.every(
          (row) =>
            Number.isSafeInteger(row.version) &&
            row.version > 0 &&
            expected.has(row.id) &&
            row.version >= (expected.get(row.id) ?? Infinity),
        );
      const ordered = rows.every(
        (row, index) => row.id === input.expectedRows[index]?.recordId,
      );
      const firstRow = rows[0];
      const filled =
        !!clipped &&
        firstRow !== undefined &&
        (rows.at(-1)?.bottom ?? 0) >= clipped.bottom - 1 &&
        firstRow.top <=
          clipped.top +
            firstRow.height * (input.groupId === null ? 2 : 3) +
            1 &&
        rows.slice(1).every((row, index) => {
          const previousRow = rows[index];
          return (
            previousRow !== undefined &&
            Math.abs(row.top - previousRow.bottom) <= 1
          );
        });
      const queryReady =
        frame?.dataset.workbookQueryFilterCount ===
          (input.operation === "filter" ? "1" : "0") &&
        frame?.dataset.workbookQuerySortCount ===
          (input.operation === "sort" ? "1" : "0") &&
        frame?.dataset.workbookQueryGroupField ===
          (input.operation === "group"
            ? "timeline.activity_time_pair_state"
            : "");
      const accepted =
        !!start &&
        start.detail?.viewSchemaId === input.viewSchemaId &&
        Number.isSafeInteger(start.detail?.afterRevision) &&
        start.detail.afterRevision >= 0 &&
        Number.isSafeInteger(revision) &&
        revision > start.detail.afterRevision;
      const settled =
        frame?.dataset.workbookQueryPending === "false" &&
        grid?.getAttribute("aria-busy") !== "true";
      const ready =
        accepted &&
        queryReady &&
        bound &&
        filled &&
        groupReady &&
        keyboardReady &&
        (input.completion !== "stable_viewport" || (ordered && settled));
      diagnostic = JSON.stringify({
        accepted,
        queryReady,
        bound,
        filled,
        groupReady,
        keyboardReady,
        ordered,
        settled,
        visibleRows: rows.length,
      });
      if (ready) {
        const same =
          revision === priorRevision &&
          rows.length === prior.length &&
          rows.every(
            (row, index) =>
              row.element === prior[index]?.element &&
              row.id === prior[index]?.id &&
              row.version === prior[index]?.version,
          );
        frames = same ? frames + 1 : 1;
        prior = rows;
        priorRevision = revision;
        if (frames >= 2 && start) {
          const stoppedAt = performance.now();
          performance.mark(input.stopMark);
          resolve({ acceptedAt: start.startTime, stoppedAt });
          return;
        }
      } else {
        frames = 0;
        prior = [];
        priorRevision = -1;
      }
      if (performance.now() >= deadline) {
        reject(new Error(`View-change paint did not qualify: ${diagnostic}`));
        return;
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  });
}

export async function measureViewChange(
  page: Page,
  predicateId: CartularyAc044PredicateId,
  expectedRows: ViewChangePaintInput["expectedRows"],
  action: () => Promise<void>,
): Promise<MeasurementSample> {
  const predicate = viewChangePredicate(predicateId);
  await resetWorkbookClientTiming(page);
  const completion = page.evaluate(observeViewChangePaintInBrowser, {
    operation: predicate.operation,
    viewSchemaId: cartularyAc044PerformanceContract.view_schema_id,
    startMark: predicate.start_mark,
    stopMark: predicate.stop_mark,
    completion: predicate.completion,
    expectedRows,
    groupId: predicate.operation === "group" ? "s:disabled" : null,
    timeoutMs: 5_000,
  });
  // Attach rejection handling before the driver yields; missing marks never become an unhandled promise.
  const observed = completion.then(
    (value) => ({ value }),
    (error) => ({ error }),
  );
  await action();
  const result = await observed;
  if ("error" in result) throw result.error;
  return {
    totalMs: result.value.stoppedAt - result.value.acceptedAt,
    stagesMs: {},
  };
}
