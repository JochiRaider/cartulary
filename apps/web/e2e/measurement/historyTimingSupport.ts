import {
  type CartularyHistoryPredicateId,
  cartularyLargeGridFixture,
  cartularyHistoryPerformanceContract as contract,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { ac043FixtureDigest, resetWorkbookClientTiming } from "./timingSupport";

export type HistoryKind = (typeof contract.predicates)[number]["kind"];
export type HistoryPaintInput = {
  kind: HistoryKind;
  recordId: string;
  actorId: string;
  priorVersion: number;
  priorHistoryRefs: string[];
  priorChangeSetIds: string[];
  gridSelector: string;
  historySelector: string;
  cells: Array<{ selector: string; fieldKey: string; value: string }>;
  startMark: string;
  timeoutMs: number;
};
export type HistoryPaintCompletion = {
  elapsed: number;
  rowVersion: number;
  historyRef: string;
  changeSetId: string;
  committedAt: string;
};

// Serialized unchanged into Chromium; the DOM regression exercises this observer.
export function observeHistoryPaintInBrowser(
  input: HistoryPaintInput,
): Promise<HistoryPaintCompletion> {
  const origin = performance.timeOrigin;
  const deadline = performance.now() + input.timeoutMs;
  const fixedGrid = document.querySelector<HTMLElement>(input.gridSelector);
  if (!fixedGrid)
    return Promise.reject(new Error("History measurement grid is missing"));
  // Product-owned inspector focus reveal is part of visible completion. Only
  // the grid and its ancestors are fixed; the driver never scrolls either pane.
  const fixedNodes = new Set<HTMLElement>([
    fixedGrid,
    ...fixedGrid.querySelectorAll<HTMLElement>("*"),
  ]);
  for (
    let parent = fixedGrid.parentElement;
    parent;
    parent = parent.parentElement
  )
    fixedNodes.add(parent);
  const scroll = [...fixedNodes]
    .filter(
      (node) =>
        node.scrollHeight > node.clientHeight ||
        node.scrollWidth > node.clientWidth,
    )
    .map((node) => ({ node, left: node.scrollLeft, top: node.scrollTop }));
  return new Promise((resolve, reject) => {
    let prior: Element[] = [],
      priorRef = "",
      frames = 0;
    function visible(node: HTMLElement, viewport: HTMLElement) {
      if (!node.isConnected || !viewport.contains(node)) return false;
      const clip = viewport.getBoundingClientRect();
      let left = Math.max(0, clip.left),
        top = Math.max(0, clip.top),
        right = Math.min(innerWidth, clip.right),
        bottom = Math.min(innerHeight, clip.bottom);
      for (
        let ancestor: HTMLElement | null = node;
        ancestor;
        ancestor = ancestor.parentElement
      ) {
        const style = getComputedStyle(ancestor);
        if (
          ancestor.hidden ||
          ancestor.hasAttribute("inert") ||
          ancestor.getAttribute("aria-hidden") === "true" ||
          style.display === "none" ||
          style.visibility === "hidden" ||
          Number(style.opacity || 1) <= 0
        )
          return false;
        if (
          ancestor !== node &&
          style.display !== "contents" &&
          /auto|scroll|hidden|clip/u.test(
            `${style.overflow} ${style.overflowX} ${style.overflowY}`,
          )
        ) {
          const box = ancestor.getBoundingClientRect();
          left = Math.max(left, box.left);
          top = Math.max(top, box.top);
          right = Math.min(right, box.right);
          bottom = Math.min(bottom, box.bottom);
        }
      }
      const box = node.getBoundingClientRect();
      return (
        box.width > 0 &&
        box.height > 0 &&
        right > left &&
        bottom > top &&
        box.left < right &&
        box.right > left &&
        box.top < bottom &&
        box.bottom > top
      );
    }
    const tick = () => {
      if (
        performance.timeOrigin !== origin ||
        document.querySelector(input.gridSelector) !== fixedGrid ||
        innerWidth !== 1920 ||
        innerHeight !== 1080 ||
        devicePixelRatio !== 1 ||
        scroll.some(
          ({ node, left, top }) =>
            node.isConnected &&
            (node.scrollLeft !== left || node.scrollTop !== top),
        )
      ) {
        reject(
          new Error("History measurement document, geometry or scroll changed"),
        );
        return;
      }
      const marks = performance.getEntriesByName(
        input.startMark,
        "mark",
      ) as PerformanceMark[];
      if (marks.length > 1) {
        reject(new Error("History action was repeated"));
        return;
      }
      const mark = marks[0];
      if (
        mark &&
        (mark.detail?.recordId !== input.recordId ||
          mark.detail.rowVersion !== input.priorVersion ||
          mark.detail.action !== input.kind ||
          mark.detail.viewSchemaId !== "cartulary.view.timeline.v2")
      ) {
        reject(new Error("History acceptance target does not match"));
        return;
      }
      const grid = document.querySelector<HTMLElement>(input.gridSelector);
      const history = document.querySelector<HTMLElement>(
        input.historySelector,
      );
      const rows = [
        ...(grid?.querySelectorAll<HTMLElement>("[data-grid-record-id]") ?? []),
      ].filter((row) => row.dataset.gridRecordId === input.recordId);
      const row = rows.length === 1 ? rows[0] : undefined;
      const version = Number(row?.dataset.gridRowVersion);
      const cells = input.cells.map((cell) =>
        row?.querySelector<HTMLElement>(cell.selector),
      );
      const entries = [
        ...(history?.querySelectorAll<HTMLElement>("[data-history-item-ref]") ??
          []),
      ].filter(
        (entry) =>
          !input.priorHistoryRefs.includes(
            entry.dataset.historyItemRef ?? "",
          ) &&
          !input.priorChangeSetIds.includes(
            entry.dataset.historyChangeSetId ?? "",
          ) &&
          !!entry.dataset.historyItemRef &&
          !!entry.dataset.historyChangeSetId &&
          entry.dataset.historyActorId === input.actorId &&
          entry.dataset.historyOperation ===
            (input.kind === "row_restore" ? "row_restore" : "rollback") &&
          Number(entry.dataset.historyRevision) === version,
      );
      const entry = entries.length === 1 ? entries[0] : undefined;
      const attribution = entry
        ? [
            entry.querySelector<HTMLElement>(":scope > strong"),
            entry.querySelector<HTMLElement>(
              '[data-generated-metadata="history-actor"]',
            ),
            entry.querySelector<HTMLTimeElement>(
              '[data-generated-metadata="history-time"]',
            ),
          ]
        : [];
      const timestamp = attribution[2]?.getAttribute("datetime");
      const ready =
        !!mark &&
        !!grid &&
        !!history &&
        history.dataset.historyRecordId === input.recordId &&
        !!row &&
        !!entry &&
        Number.isSafeInteger(version) &&
        version === input.priorVersion + 1 &&
        cells.length === 2 &&
        cells.every((cell, index) => {
          const target = input.cells[index];
          return (
            !!cell &&
            !!target &&
            cell.closest<HTMLElement>("[data-grid-field-key]")?.dataset
              .gridFieldKey === target.fieldKey &&
            cell.textContent === target.value &&
            visible(cell, grid)
          );
        }) &&
        attribution.length === 3 &&
        attribution.every(
          (node) =>
            !!node && !!node.textContent?.trim() && visible(node, history),
        ) &&
        !!timestamp &&
        Number.isFinite(Date.parse(timestamp));
      const nodes =
        ready && row && entry
          ? [
              row,
              ...cells.filter((cell): cell is HTMLElement => cell != null),
              entry,
              ...attribution.filter(
                (node): node is HTMLElement => node != null,
              ),
            ]
          : [];
      const ref = entry?.dataset.historyItemRef ?? "";
      frames = ready
        ? ref === priorRef &&
          nodes.every((node, index) => node === prior[index])
          ? frames + 1
          : 1
        : 0;
      prior = nodes;
      priorRef = ready ? ref : "";
      if (frames === 2) {
        if (
          !mark ||
          !entry ||
          !timestamp ||
          !entry.dataset.historyChangeSetId
        ) {
          reject(new Error("History completion lost its observation binding"));
          return;
        }
        const elapsed = performance.now() - mark.startTime;
        if (!Number.isFinite(elapsed) || elapsed < 0) {
          reject(new Error("History elapsed interval is invalid"));
          return;
        }
        resolve({
          elapsed,
          rowVersion: version,
          historyRef: ref,
          changeSetId: entry.dataset.historyChangeSetId,
          committedAt: timestamp,
        });
        return;
      }
      if (performance.now() >= deadline) {
        reject(
          new Error(
            "History row and attribution did not qualify across two visible frames",
          ),
        );
        return;
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  });
}

export async function measureHistory(
  page: Page,
  input: Omit<HistoryPaintInput, "startMark" | "timeoutMs">,
  action: () => Promise<void>,
) {
  await resetWorkbookClientTiming(page);
  const observation = page
    .evaluate(observeHistoryPaintInBrowser, {
      ...input,
      startMark: contract.start_mark,
      timeoutMs: 5_000,
    })
    .then(
      (value) => ({ value }),
      (error) => ({ error }),
    );
  await action();
  const result = await observation;
  if ("error" in result) throw result.error;
  return result.value;
}

export async function attachHistoryObservation(
  testInfo: TestInfo,
  predicateId: CartularyHistoryPredicateId,
  target?: {
    kind: HistoryKind;
    recordId: string;
    actorId: string;
    priorVersion: number;
  },
  completion?: HistoryPaintCompletion,
  failureReason?: string,
) {
  const elapsed = completion?.elapsed ?? null;
  const observation = {
    schema_id: "cartulary.frontend_measurement_observation.v4",
    fixture_profile_id: "timeline_large_grid_snapshot_v2",
    criterion_id: contract.criterion_id,
    predicate_id: predicateId,
    fixture_id: cartularyLargeGridFixture.fixtureId,
    fixture_digest: ac043FixtureDigest(),
    measurement_policy_id: contract.measurement_policy_id,
    threshold_ms: contract.threshold_ms,
    warmup_samples: 0,
    measured_samples: completion ? 1 : 0,
    statistics: { kind: "single", elapsed_ms: elapsed },
    outcome:
      process.env.CARTULARY_BROWSER_RESOURCE_PROFILE_ID !==
      "browser_measurement_quiet"
        ? "environment_not_qualified"
        : failureReason || elapsed === null
          ? "incomplete"
          : elapsed < contract.threshold_ms
            ? "passed"
            : "threshold_failed",
    traffic: {
      counts: [
        {
          fact_id: "analyst_sessions",
          value: cartularyLargeGridFixture.analystSessions,
        },
        {
          fact_id: "background_sessions",
          value: cartularyLargeGridFixture.backgroundSessions,
        },
      ],
      rates: [
        {
          fact_id: "background_updates_per_second",
          value: cartularyLargeGridFixture.backgroundUpdatesPerSecond,
        },
      ],
      conditions: [
        { fact_id: "presence_enabled", value: true },
        { fact_id: "target_row_excluded", value: true },
      ],
    },
    samples: completion
      ? [
          {
            sample_index: 0,
            warmup: false,
            total_ms: completion.elapsed,
            stages_ms: [],
          },
        ]
      : [],
    predicate_evidence:
      completion && target
        ? {
            schema_id: "cartulary.history_measurement_proof.v2",
            kind: "timeline_history",
            action_kind: target.kind,
            record_id: target.recordId,
            actor_user_id: target.actorId,
            row_version: completion.rowVersion,
            prior_row_version: target.priorVersion,
            view_schema_id: contract.view_schema_id,
            visible_cells: 2,
            consecutive_frames: 2,
            new_history_entry: true,
            attribution_verified: true,
            binding_stable: true,
            geometry_unchanged: true,
            grid_scroll_unchanged: true,
            action_driver: contract.workload.action_driver,
            inspector_scroll_policy: contract.workload.inspector_scroll_policy,
            action_count: 1,
            viewport_width: 1920,
            viewport_height: 1080,
            device_scale_factor: 1,
          }
        : null,
    ...(failureReason ? { failure_reason: failureReason } : {}),
  };
  await testInfo.attach(`${observation.schema_id}.${predicateId}`, {
    body: JSON.stringify(observation),
    contentType: "application/json",
  });
}
