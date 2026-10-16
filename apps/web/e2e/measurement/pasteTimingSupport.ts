import {
  cartularyLargeGridFixture,
  cartularyAc003PerformanceContract as contract,
  dataTestIdSelector,
  draftCellTestId,
  gridFieldCellSelector,
  gridRowVersionAttribute,
  gridSavedRowsSelector,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { ac043FixtureDigest, resetWorkbookClientTiming } from "./timingSupport";

export type PastePaintCompletion = {
  acceptedAt: number;
  stoppedAt: number;
  createdRecords: Array<{ record_id: string; row_version: number }>;
};

type PastePaintOptions = {
  gridSelector: string;
  rowSelector: string;
  fieldSelectors: string[];
  versionAttribute: string;
  existingRecordIds: string[];
  expectedRows: string[][];
  startMark: string;
  stopMark: string;
  viewportWidth: number;
  viewportHeight: number;
  deviceScaleFactor: number;
  timeoutMs: number;
};

// This function is serialized into the browser. Keep its inputs explicit and
// its paint observation independent of fixture setup and evidence serialization.
export function observePastePaintInBrowser(
  options: PastePaintOptions,
): Promise<PastePaintCompletion> {
  return new Promise((resolve, reject) => {
    if (
      options.expectedRows.length !== 20 ||
      options.fieldSelectors.length !== 5 ||
      options.expectedRows.some(
        (row) => row.length !== 5 || row.some((value) => value.length === 0),
      )
    ) {
      reject(
        new Error(
          "paste measurement requires exactly 20 rows and five nonempty cells",
        ),
      );
      return;
    }
    const existing = new Set(options.existingRecordIds);
    const deadline = performance.now() + options.timeoutMs;
    let prior: Array<{
      row: Element;
      cells: Element[];
      record_id: string;
      row_version: number;
    }> = [];
    let lastState = "not_observed";
    const visible = (element: HTMLElement, grid: HTMLElement) => {
      if (
        !element.isConnected ||
        element.closest("[hidden], [aria-hidden='true']")
      )
        return false;
      const rect = element.getBoundingClientRect();
      const clip = grid.getBoundingClientRect();
      let left = Math.max(0, clip.left),
        top = Math.max(0, clip.top);
      let right = Math.min(innerWidth, clip.right),
        bottom = Math.min(innerHeight, clip.bottom);
      for (
        let ancestor: HTMLElement | null = element;
        ancestor;
        ancestor = ancestor.parentElement
      ) {
        const style = getComputedStyle(ancestor);
        if (
          style.display === "none" ||
          style.visibility === "hidden" ||
          Number.parseFloat(style.opacity || "1") === 0
        )
          return false;
        if (
          ancestor !== element &&
          /auto|scroll|hidden|clip/u.test(
            `${style.overflow} ${style.overflowX} ${style.overflowY}`,
          )
        ) {
          const bounds = ancestor.getBoundingClientRect();
          left = Math.max(left, bounds.left);
          right = Math.min(right, bounds.right);
          top = Math.max(top, bounds.top);
          bottom = Math.min(bottom, bounds.bottom);
        }
      }
      return (
        rect.width > 0 &&
        rect.height > 0 &&
        right > left &&
        bottom > top &&
        rect.right > left &&
        rect.left < right &&
        rect.bottom > top &&
        rect.top < bottom
      );
    };
    const tick = () => {
      if (
        innerWidth !== options.viewportWidth ||
        innerHeight !== options.viewportHeight ||
        devicePixelRatio !== options.deviceScaleFactor
      ) {
        reject(
          new Error("paste measurement geometry changed or was not qualified"),
        );
        return;
      }
      const marks = performance.getEntriesByName(options.startMark, "mark");
      if (marks.length > 1) {
        reject(new Error("paste measurement action was repeated"));
        return;
      }
      const start = marks[0];
      const grid = document.querySelector<HTMLElement>(options.gridSelector);
      const matches: typeof prior = [];
      const identities = new Set<string>();
      if (start && grid) {
        const rows = [
          ...grid.querySelectorAll<HTMLElement>(options.rowSelector),
        ];
        for (const expected of options.expectedRows) {
          const candidates = rows.flatMap((row) => {
            const record_id = row.dataset.gridRecordId ?? "";
            const row_version = Number(
              row.getAttribute(options.versionAttribute),
            );
            if (
              !/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/u.test(
                record_id,
              ) ||
              existing.has(record_id) ||
              !Number.isSafeInteger(row_version) ||
              row_version < 1
            )
              return [];
            const cells: Element[] = [];
            for (const [index, selector] of options.fieldSelectors.entries()) {
              const field = row.querySelector<HTMLElement>(selector);
              const cell =
                field?.querySelector<HTMLInputElement | HTMLTextAreaElement>(
                  "input, textarea",
                ) ?? field;
              if (!cell || !visible(cell, grid)) return [];
              const value =
                cell instanceof HTMLInputElement ||
                cell instanceof HTMLTextAreaElement
                  ? cell.value
                  : cell.textContent?.trim();
              if (value !== expected[index]) return [];
              cells.push(cell);
            }
            return [{ row, cells, record_id, row_version }];
          });
          const candidate = candidates[0];
          if (
            candidates.length !== 1 ||
            candidate === undefined ||
            identities.has(candidate.record_id)
          )
            break;
          identities.add(candidate.record_id);
          matches.push(candidate);
        }
      }
      const complete = matches.length === options.expectedRows.length;
      const stable =
        complete &&
        prior.length === matches.length &&
        matches.every((entry, index) => {
          const previous = prior[index];
          return (
            previous !== undefined &&
            entry.row === previous.row &&
            entry.record_id === previous.record_id &&
            entry.row_version === previous.row_version &&
            entry.cells.every(
              (cell, cellIndex) => cell === previous.cells[cellIndex],
            )
          );
        });
      if (start && stable) {
        performance.mark(options.stopMark);
        resolve({
          acceptedAt: start.startTime,
          stoppedAt: performance.now(),
          createdRecords: matches.map(({ record_id, row_version }) => ({
            record_id,
            row_version,
          })),
        });
        return;
      }
      prior = complete ? matches : [];
      lastState = start
        ? `qualified_rows=${matches.length}`
        : "missing_acceptance_mark";
      if (performance.now() > deadline) {
        reject(new Error(`paste visible completion failed: ${lastState}`));
        return;
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  });
}

export async function measureTimelinePaste(
  page: Page,
  existingRecordIds: string[],
) {
  await resetWorkbookClientTiming(page);
  const completion = page.evaluate(observePastePaintInBrowser, {
    gridSelector: dataTestIdSelector(gridShellTestId(contract.view_schema_id)),
    rowSelector: gridSavedRowsSelector(),
    fieldSelectors: contract.field_keys.map(gridFieldCellSelector),
    versionAttribute: gridRowVersionAttribute,
    existingRecordIds,
    expectedRows: contract.rows.map((row) => [...row]),
    startMark: contract.start_mark,
    stopMark: contract.stop_mark,
    viewportWidth: contract.viewport_width,
    viewportHeight: contract.viewport_height,
    deviceScaleFactor: contract.device_scale_factor,
    timeoutMs: 10_000,
  });
  const payload = contract.rows.map((row) => row.join("\t")).join("\n");
  await page
    .getByTestId(draftCellTestId(contract.field_keys[0]))
    .locator("xpath=ancestor::*[@role='gridcell'][1]")
    .evaluate((element, text) => {
      const data = new DataTransfer();
      data.setData("text/plain", text);
      element.dispatchEvent(
        new ClipboardEvent("paste", {
          bubbles: true,
          cancelable: true,
          clipboardData: data,
        }),
      );
    }, payload);
  return completion;
}

export async function attachPasteObservation(
  testInfo: TestInfo,
  completion?: PastePaintCompletion,
  failureReason?: string,
) {
  const elapsed = completion
    ? completion.stoppedAt - completion.acceptedAt
    : null;
  const quiet =
    process.env.CARTULARY_BROWSER_RESOURCE_PROFILE_ID ===
    "browser_measurement_quiet";
  const fixture = cartularyLargeGridFixture;
  const outcome = !quiet
    ? "environment_not_qualified"
    : elapsed === null || failureReason
      ? "incomplete"
      : elapsed < contract.threshold_ms
        ? "passed"
        : "threshold_failed";
  const observation = {
    schema_id: "cartulary.frontend_measurement_observation.v4",
    fixture_profile_id: "timeline_large_grid_snapshot_v2",
    criterion_id: contract.criterion_id,
    predicate_id: contract.predicate_id,
    fixture_id: fixture.fixtureId,
    fixture_digest: ac043FixtureDigest(),
    measurement_policy_id: contract.measurement_policy_id,
    threshold_ms: contract.threshold_ms,
    warmup_samples: 0,
    measured_samples: elapsed === null ? 0 : 1,
    statistics: { kind: "single", elapsed_ms: elapsed },
    outcome,
    traffic: {
      counts: [
        { fact_id: "analyst_sessions", value: fixture.analystSessions },
        { fact_id: "background_sessions", value: fixture.backgroundSessions },
      ],
      rates: [
        {
          fact_id: "background_updates_per_second",
          value: fixture.backgroundUpdatesPerSecond,
        },
      ],
      conditions: [
        { fact_id: "presence_enabled", value: true },
        { fact_id: "target_row_excluded", value: true },
      ],
    },
    samples:
      elapsed === null
        ? []
        : [
            {
              sample_index: 0,
              warmup: false,
              total_ms: elapsed,
              stages_ms: [
                { fact_id: "accepted_to_visible_paint", value: elapsed },
              ],
            },
          ],
    predicate_evidence: completion
      ? {
          kind: "timeline_paste_20x5",
          clipboard_format: contract.clipboard_format,
          clipboard_mime: contract.clipboard_mime,
          clipboard_delivery: "synthesized_event",
          viewport_width: contract.viewport_width,
          viewport_height: contract.viewport_height,
          device_scale_factor: contract.device_scale_factor,
          row_count: 20,
          cell_count: 100,
          consecutive_frames: 2,
          all_new_ids: true,
          geometry_unchanged: true,
          target_bindings_stable: true,
          created_records: completion.createdRecords,
        }
      : null,
    ...(failureReason ? { failure_reason: failureReason.slice(0, 512) } : {}),
  };
  await testInfo.attach(`${observation.schema_id}.${contract.predicate_id}`, {
    body: Buffer.from(`${JSON.stringify(observation)}\n`),
    contentType:
      "application/vnd.cartulary.frontend-measurement-observation+json",
  });
  return observation;
}
