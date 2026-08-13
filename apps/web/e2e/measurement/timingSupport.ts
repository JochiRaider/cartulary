import { createHash } from "node:crypto";

import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  type CartularyAc043PredicateId,
  cartularyAc043PerformanceContract,
  cartularyInteractiveP95MeasurementPolicy,
  draftCellTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";

export const interactiveMeasurementSamplePolicy =
  cartularyInteractiveP95MeasurementPolicy;

export type ServerTimingMetric = {
  attributes: Record<string, string | true>;
  durationMs: number | null;
  name: string;
  raw: string;
};

export type ClientTimingEvent = {
  at: number;
  name: string;
  [key: string]: unknown;
};

export type MeasurementStages = Record<string, number>;

export type MeasurementSample = {
  stagesMs: MeasurementStages;
  totalMs: number;
};

type CommittedRowSummaryMatch = { recordId: string; rowVersion: number };
const workbookTimingMarkPrefix = "cartulary.workbook.";

export function performancePredicate(predicateId: CartularyAc043PredicateId) {
  const predicate = cartularyAc043PerformanceContract.predicates.find(
    (candidate) => candidate.predicateId === predicateId,
  );
  if (predicate === undefined) {
    throw new Error(`unknown AC-043 performance predicate ${predicateId}`);
  }
  return predicate;
}

export function ac043FixtureDigest() {
  return `sha256:${createHash("sha256")
    .update(canonicalJSON(cartularyAc043PerformanceContract.fixture))
    .digest("hex")}`;
}

export async function measureSelectionDown(
  page: Page,
  options: { fromRecordId: string; toRecordId: string },
): Promise<MeasurementSample> {
  const predicate = performancePredicate(
    "perf.timeline_summary_selection_down.v1",
  );
  const sourceTestId = rowCellTestId(options.fromRecordId, predicate.fieldKey);
  const targetTestId = rowCellTestId(options.toRecordId, predicate.fieldKey);
  await resetWorkbookClientTiming(page);
  const driverDispatchAt = await page.evaluate(() => performance.now());
  const completion = waitForCellPaint(page, {
    mode: "selected",
    startMark: predicate.startMark,
    stopMark: predicate.stopMark,
    testId: targetTestId,
    timeoutMs: 5_000,
  });
  await page.getByTestId(sourceTestId).press(predicate.initiatingKey);
  return sampleFromCompletion(await completion, driverDispatchAt);
}

export async function measureFocusEdit(
  page: Page,
  options: { recordId: string },
): Promise<MeasurementSample> {
  const predicate = performancePredicate("perf.timeline_summary_focus_edit.v1");
  const editorTestId = timelineScalarEditorTestId({
    fieldKey: predicate.fieldKey,
    recordId: options.recordId,
    surface: "grid",
  });
  const triggerTestId = rowCellTestId(options.recordId, predicate.fieldKey);
  await resetWorkbookClientTiming(page);
  const driverDispatchAt = await page.evaluate(() => performance.now());
  const completion = waitForCellPaint(page, {
    mode: "editor",
    startMark: predicate.startMark,
    stopMark: predicate.stopMark,
    testId: editorTestId,
    timeoutMs: 5_000,
  });
  await page.getByTestId(triggerTestId).press(predicate.initiatingKey);
  return sampleFromCompletion(await completion, driverDispatchAt);
}

export async function measureTypingAck(
  page: Page,
  options: { recordId: string },
): Promise<MeasurementSample> {
  const predicate = performancePredicate("perf.typing_ack.v1");
  const testId = timelineScalarEditorTestId({
    fieldKey: predicate.fieldKey,
    recordId: options.recordId,
    surface: "grid",
  });
  const input = page.getByTestId(testId);
  const currentValue = await input.inputValue();
  await resetWorkbookClientTiming(page);
  const driverDispatchAt = await page.evaluate(() => performance.now());
  const completion = waitForCellPaint(page, {
    expectedValue: `${currentValue}${predicate.initiatingKey}`,
    mode: "editor",
    startMark: predicate.startMark,
    stopMark: predicate.stopMark,
    testId,
    timeoutMs: 5_000,
  });
  await input.press(predicate.initiatingKey);
  return sampleFromCompletion(await completion, driverDispatchAt);
}

export async function measureBlankRowCreate(
  page: Page,
  expectedSummary: string,
): Promise<MeasurementSample & { status: number }> {
  const predicate = performancePredicate("perf.timeline_blank_row_create.v1");
  const draftSummary = page.getByTestId(draftCellTestId(predicate.fieldKey));
  await draftSummary.focus();
  await resetWorkbookClientTiming(page);
  const driverDispatchAt = await page.evaluate(() => performance.now());
  const completion = waitForBlankRowPaint(page, {
    expectedSummary,
    startMark: predicate.startMark,
    stopMark: predicate.stopMark,
    timeoutMs: 10_000,
  });
  await draftSummary.press(predicate.initiatingKey);
  const visible = await completion;
  const events = await readWorkbookClientTiming(page);
  const statusEvent = events.find(
    (event) => event.name === "pending_fetch_response",
  );
  const status =
    typeof statusEvent?.status === "number" ? statusEvent.status : 0;
  return {
    ...sampleFromCompletion(visible, driverDispatchAt, events),
    status,
  };
}

export function percentile95(
  samples: number[],
  options: { minimumSampleCount?: number; sampleLabel?: string } = {},
) {
  const minimumSampleCount =
    options.minimumSampleCount ??
    interactiveMeasurementSamplePolicy.measuredSamples;
  return nearestRankPercentile(samples, 95, {
    minimumSampleCount,
    sampleLabel: options.sampleLabel ?? "samples",
  });
}

export function nearestRankPercentile(
  samples: number[],
  percentile: number,
  options: { minimumSampleCount?: number; sampleLabel?: string } = {},
) {
  const label = options.sampleLabel ?? "samples";
  const minimum = options.minimumSampleCount ?? 1;
  if (!Number.isFinite(percentile) || percentile <= 0 || percentile > 100) {
    throw new Error(
      `percentile for ${label} must be finite and within (0, 100]`,
    );
  }
  if (samples.length < minimum) {
    throw new Error(
      `cannot compute percentile for ${label}: expected at least ${minimum} samples, got ${samples.length}`,
    );
  }
  if (samples.some((sample) => !Number.isFinite(sample) || sample < 0)) {
    throw new Error(`${label} must contain only finite non-negative samples`);
  }
  const sorted = [...samples].sort((left, right) => left - right);
  return sorted[Math.ceil((percentile / 100) * sorted.length) - 1] as number;
}

export async function attachMeasurementSummary(
  testInfo: TestInfo,
  options: {
    failureReason?: string;
    fixtureDigest: string;
    predicateId: CartularyAc043PredicateId;
    samples: MeasurementSample[];
  },
) {
  const predicate = performancePredicate(options.predicateId);
  const measured = options.samples.slice(
    interactiveMeasurementSamplePolicy.warmupPasses,
  );
  const complete =
    options.samples.length ===
      interactiveMeasurementSamplePolicy.totalSamples &&
    measured.length === interactiveMeasurementSamplePolicy.measuredSamples;
  const p50 = complete
    ? nearestRankPercentile(
        measured.map((sample) => sample.totalMs),
        50,
      )
    : null;
  const p95 = complete
    ? percentile95(
        measured.map((sample) => sample.totalMs),
        {
          sampleLabel: options.predicateId,
        },
      )
    : null;
  const quietQualified =
    process.env.CARTULARY_BROWSER_RESOURCE_PROFILE_ID ===
    "browser_measurement_quiet";
  const outcome = !quietQualified
    ? "environment_not_qualified"
    : !complete
      ? "incomplete"
      : (p95 as number) <= predicate.thresholdMs
        ? "passed"
        : "threshold_failed";
  const warmupSamples = Math.min(
    options.samples.length,
    interactiveMeasurementSamplePolicy.warmupPasses,
  );
  const measuredSamples = Math.max(0, options.samples.length - warmupSamples);
  const summary = {
    schema_id: "cartulary.frontend_measurement_summary.v1",
    criterion_id: cartularyAc043PerformanceContract.criterionId,
    predicate_id: predicate.predicateId,
    fixture_id: cartularyAc043PerformanceContract.fixture.fixtureId,
    fixture_digest: options.fixtureDigest,
    measurement_policy_id:
      cartularyAc043PerformanceContract.measurementPolicyId,
    threshold_ms: predicate.thresholdMs,
    warmup_samples: warmupSamples,
    measured_samples: measuredSamples,
    percentile: interactiveMeasurementSamplePolicy.percentile,
    p50_ms: p50,
    p95_ms: p95,
    outcome,
    qualification: {
      quiet_profile_id: "browser_measurement_quiet",
      scheduler_overlap_count: quietQualified ? 0 : null,
      analyst_sessions:
        cartularyAc043PerformanceContract.fixture.analystSessions,
      background_updates_per_second:
        cartularyAc043PerformanceContract.fixture.backgroundUpdatesPerSecond,
    },
    samples: options.samples.map((sample, sampleIndex) => ({
      sample_index: sampleIndex,
      warmup: sampleIndex < interactiveMeasurementSamplePolicy.warmupPasses,
      total_ms: sample.totalMs,
      stages_ms: sample.stagesMs,
    })),
    ...(options.failureReason === undefined
      ? {}
      : { failure_reason: options.failureReason.slice(0, 512) }),
  };
  await testInfo.attach(
    `cartulary.frontend_measurement_summary.v1.${predicate.predicateId}`,
    {
      body: Buffer.from(`${JSON.stringify(summary)}\n`, "utf8"),
      contentType:
        "application/vnd.cartulary.frontend-measurement-summary+json",
    },
  );
  return summary;
}

type PaintCompletion = {
  acceptedAt: number;
  stoppedAt: number;
};

async function waitForCellPaint(
  page: Page,
  options: {
    expectedValue?: string;
    mode: "selected" | "editor";
    startMark: string;
    stopMark: string;
    testId: string;
    timeoutMs: number;
  },
) {
  return page.evaluate<PaintCompletion, typeof options>(
    observeCellPaintInBrowser,
    options,
  );
}

export function observeCellPaintInBrowser(options: {
  expectedValue?: string;
  mode: "selected" | "editor";
  startMark: string;
  stopMark: string;
  testId: string;
  timeoutMs: number;
}): Promise<PaintCompletion> {
  const { expectedValue, mode, startMark, stopMark, testId, timeoutMs } =
    options;
  return new Promise((resolve, reject) => {
    const deadline = performance.now() + timeoutMs;
    let priorTarget: Element | null = null;
    let consecutiveFrames = 0;
    let lastQualification = "not_observed";
    const visible = (target: Element) => {
      const cell = target.closest<HTMLElement>('[role="gridcell"]');
      const grid = target.closest<HTMLElement>('[role="grid"]');
      if (cell === null || grid === null) return false;
      const targetRect = target.getBoundingClientRect();
      const gridRect = grid.getBoundingClientRect();
      const style = getComputedStyle(target);
      return (
        target.isConnected &&
        style.display !== "none" &&
        style.visibility !== "hidden" &&
        Number.parseFloat(style.opacity || "1") > 0 &&
        targetRect.width > 0 &&
        targetRect.height > 0 &&
        targetRect.right > gridRect.left &&
        targetRect.left < gridRect.right &&
        targetRect.bottom > gridRect.top &&
        targetRect.top < gridRect.bottom &&
        (mode !== "selected" ||
          document.activeElement === cell ||
          cell.contains(document.activeElement))
      );
    };
    const tick = () => {
      const start = performance.getEntriesByName(startMark, "mark").at(-1);
      const target =
        Array.from(document.querySelectorAll("[data-testid]")).find(
          (element) => element.getAttribute("data-testid") === testId,
        ) ?? null;
      const input =
        target instanceof HTMLInputElement ||
        target instanceof HTMLTextAreaElement
          ? target
          : null;
      const editorReady =
        mode !== "editor" ||
        (input !== null &&
          document.activeElement === input &&
          input.selectionStart === input.value.length &&
          input.selectionEnd === input.value.length &&
          (expectedValue === undefined || input.value === expectedValue));
      lastQualification = JSON.stringify({
        accepted_mark: start !== undefined,
        active: target !== null && document.activeElement === target,
        caret_end:
          input !== null &&
          input.selectionStart === input.value.length &&
          input.selectionEnd === input.value.length,
        expected_value:
          expectedValue === undefined || input?.value === expectedValue,
        target_present: target !== null,
        target_visible: target !== null && visible(target),
      });
      if (
        start !== undefined &&
        target !== null &&
        visible(target) &&
        editorReady
      ) {
        consecutiveFrames = target === priorTarget ? consecutiveFrames + 1 : 1;
        priorTarget = target;
        if (consecutiveFrames >= 2) {
          const stoppedAt = performance.now();
          performance.mark(stopMark, {
            detail: {
              field: "timeline.activity_synopsis_text",
              surface: "grid",
            },
          });
          resolve({ acceptedAt: start.startTime, stoppedAt });
          return;
        }
      } else {
        consecutiveFrames = 0;
        priorTarget = null;
      }
      if (performance.now() > deadline) {
        reject(
          new Error(
            `timed out waiting for paint-qualified ${testId}; qualification=${lastQualification}`,
          ),
        );
        return;
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  });
}

async function waitForBlankRowPaint(
  page: Page,
  options: {
    expectedSummary: string;
    startMark: string;
    stopMark: string;
    timeoutMs: number;
  },
) {
  return page.evaluate<PaintCompletion, typeof options>(
    observeBlankRowPaintInBrowser,
    options,
  );
}

export function observeBlankRowPaintInBrowser(options: {
  expectedSummary: string;
  startMark: string;
  stopMark: string;
  timeoutMs: number;
}): Promise<PaintCompletion> {
  const { expectedSummary, startMark, stopMark, timeoutMs } = options;
  return new Promise((resolve, reject) => {
    const deadline = performance.now() + timeoutMs;
    let priorRecordId: string | null = null;
    let priorVersion: number | null = null;
    let consecutiveFrames = 0;
    let lastDiagnostics = {
      acceptedMark: false,
      gridPresent: false,
      mountedRows: 0,
      summaryMatch: false,
      versionedSummaryMatch: false,
      visibleSummaryMatch: false,
    };
    const tick = () => {
      const start = performance.getEntriesByName(startMark, "mark").at(-1);
      const grid =
        Array.from(
          document.querySelectorAll<HTMLElement>("[data-testid]"),
        ).find(
          (element) =>
            element.getAttribute("data-testid") ===
            "cartulary.view.timeline.v2-grid-shell",
        ) ?? null;
      let match: { recordId: string; rowVersion: number } | null = null;
      let summaryMatch = false;
      let versionedSummaryMatch = false;
      let visibleSummaryMatch = false;
      const mountedRows =
        grid?.querySelectorAll<HTMLElement>("[data-grid-record-id]") ?? [];
      if (grid !== null) {
        for (const row of mountedRows) {
          const recordId = row.dataset.gridRecordId;
          if (!recordId) continue;
          const summary = Array.from(
            row.querySelectorAll<HTMLElement>("[data-testid]"),
          ).find(
            (element) =>
              element.getAttribute("data-testid") ===
              `row-${recordId}-timeline.activity_synopsis_text`,
          );
          const value =
            summary instanceof HTMLInputElement ||
            summary instanceof HTMLTextAreaElement
              ? summary.value
              : summary?.textContent?.trim();
          const versionNode =
            Array.from(row.querySelectorAll<HTMLElement>("[data-testid]")).find(
              (element) =>
                element.getAttribute("data-testid") ===
                `row-${recordId}-row_version`,
            ) ??
            Array.from(
              document.querySelectorAll<HTMLElement>("[data-testid]"),
            ).find(
              (element) =>
                element.getAttribute("data-testid") ===
                `row-${recordId}-row_version`,
            );
          const rowVersion = Number.parseInt(
            versionNode?.textContent ?? "",
            10,
          );
          if (value === expectedSummary && summary !== undefined) {
            summaryMatch = true;
          }
          if (
            value !== expectedSummary ||
            !Number.isInteger(rowVersion) ||
            rowVersion < 1 ||
            summary === undefined
          )
            continue;
          versionedSummaryMatch = true;
          const targetRect = summary.getBoundingClientRect();
          const gridRect = grid.getBoundingClientRect();
          const style = getComputedStyle(summary);
          if (
            summary.isConnected &&
            style.display !== "none" &&
            style.visibility !== "hidden" &&
            targetRect.width > 0 &&
            targetRect.height > 0 &&
            targetRect.right > gridRect.left &&
            targetRect.left < gridRect.right &&
            targetRect.bottom > gridRect.top &&
            targetRect.top < gridRect.bottom
          ) {
            visibleSummaryMatch = true;
            match = { recordId, rowVersion };
            break;
          }
        }
      }
      lastDiagnostics = {
        acceptedMark: start !== undefined,
        gridPresent: grid !== null,
        mountedRows: mountedRows.length,
        summaryMatch,
        versionedSummaryMatch,
        visibleSummaryMatch,
      };
      if (start !== undefined && match !== null) {
        consecutiveFrames =
          match.recordId === priorRecordId && match.rowVersion === priorVersion
            ? consecutiveFrames + 1
            : 1;
        priorRecordId = match.recordId;
        priorVersion = match.rowVersion;
        if (consecutiveFrames >= 2) {
          const stoppedAt = performance.now();
          performance.mark(stopMark, {
            detail: {
              field: "timeline.activity_synopsis_text",
              surface: "grid",
            },
          });
          resolve({ acceptedAt: start.startTime, stoppedAt });
          return;
        }
      } else {
        consecutiveFrames = 0;
        priorRecordId = null;
        priorVersion = null;
      }
      if (performance.now() > deadline) {
        reject(
          new Error(
            `timed out waiting for paint-qualified committed Timeline row: ${JSON.stringify(lastDiagnostics)}`,
          ),
        );
        return;
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  });
}

function sampleFromCompletion(
  completion: PaintCompletion,
  driverDispatchAt: number,
  events: ClientTimingEvent[] = [],
): MeasurementSample {
  const byName = new Map(events.map((event) => [event.name, event.at]));
  const stagesMs: MeasurementStages = {};
  addDuration(
    stagesMs,
    "driver_dispatch",
    driverDispatchAt,
    completion.acceptedAt,
  );
  addDuration(
    stagesMs,
    "accepted_action_to_request",
    completion.acceptedAt,
    byName.get("pending_fetch_request"),
  );
  addDuration(
    stagesMs,
    "request_round_trip",
    byName.get("pending_fetch_request"),
    byName.get("pending_fetch_response"),
  );
  addDuration(
    stagesMs,
    "response_decode",
    byName.get("pending_fetch_response"),
    byName.get("pending_fetch_json_parsed"),
  );
  addDuration(
    stagesMs,
    "client_apply",
    byName.get("apply_row_mutation_start"),
    byName.get("apply_row_mutation_end"),
  );
  addDuration(
    stagesMs,
    "apply_to_visible_paint",
    byName.get("apply_row_mutation_end"),
    completion.stoppedAt,
  );
  const totalMs = completion.stoppedAt - completion.acceptedAt;
  if (!Number.isFinite(totalMs) || totalMs < 0) {
    throw new Error("measurement total must be finite and non-negative");
  }
  return { stagesMs, totalMs };
}

function addDuration(
  target: MeasurementStages,
  name: string,
  start: number | undefined,
  end: number | undefined,
) {
  if (start === undefined || end === undefined) return;
  const duration = end - start;
  if (Number.isFinite(duration) && duration >= 0) target[name] = duration;
}

export function parseServerTiming(header: string): ServerTimingMetric[] {
  return splitServerTimingHeader(header)
    .map((rawPart) => rawPart.trim())
    .filter((rawPart) => rawPart !== "")
    .map((raw) => {
      const [rawName = "", ...rawAttributes] = raw.split(";");
      const attributes: Record<string, string | true> = {};
      let durationMs: number | null = null;
      for (const rawAttribute of rawAttributes) {
        const attribute = rawAttribute.trim();
        if (attribute === "") continue;
        const separatorIndex = attribute.indexOf("=");
        if (separatorIndex < 0) {
          attributes[attribute] = true;
          continue;
        }
        const key = attribute.slice(0, separatorIndex).trim();
        const value = unquoteServerTimingValue(
          attribute.slice(separatorIndex + 1).trim(),
        );
        attributes[key] = value;
        if (key.toLowerCase() === "dur") {
          const parsed = Number.parseFloat(value);
          durationMs = Number.isFinite(parsed) ? parsed : null;
        }
      }
      return { attributes, durationMs, name: rawName.trim(), raw };
    });
}

export async function waitForCommittedRowSummary(
  page: Page,
  options: {
    expectedSummary: string;
    startedAtMs?: number;
    surface: string;
    timeoutMs: number;
  },
) {
  await scrollGridTargetIntoView({
    page,
    surface: options.surface,
    targetTestId: gridSortHeaderTestId(
      options.surface,
      "timeline.activity_synopsis_text",
    ),
    timeoutMs: options.timeoutMs,
  });
  const startedAtMs =
    options.startedAtMs ?? (await page.evaluate(() => performance.now()));
  return page.evaluate(
    ({ expectedSummary, startedAtMs, timeoutMs }) =>
      new Promise<{ durationMs: number; recordId: string; rowVersion: number }>(
        (resolve, reject) => {
          const deadline = startedAtMs + timeoutMs;
          const tick = () => {
            for (const row of document.querySelectorAll<HTMLElement>(
              "[data-grid-record-id]",
            )) {
              const recordId = row.dataset.gridRecordId;
              if (!recordId) continue;
              const summary = Array.from(
                row.querySelectorAll<HTMLElement>("[data-testid]"),
              ).find(
                (element) =>
                  element.getAttribute("data-testid") ===
                  `row-${recordId}-timeline.activity_synopsis_text`,
              );
              const value =
                summary instanceof HTMLInputElement ||
                summary instanceof HTMLTextAreaElement
                  ? summary.value
                  : summary?.textContent?.trim();
              const version =
                Array.from(
                  row.querySelectorAll<HTMLElement>("[data-testid]"),
                ).find(
                  (element) =>
                    element.getAttribute("data-testid") ===
                    `row-${recordId}-row_version`,
                ) ??
                Array.from(
                  document.querySelectorAll<HTMLElement>("[data-testid]"),
                ).find(
                  (element) =>
                    element.getAttribute("data-testid") ===
                    `row-${recordId}-row_version`,
                );
              const rowVersion = Number.parseInt(
                version?.textContent ?? "",
                10,
              );
              if (
                value === expectedSummary &&
                Number.isInteger(rowVersion) &&
                rowVersion > 0
              ) {
                resolve({
                  durationMs: performance.now() - startedAtMs,
                  recordId,
                  rowVersion,
                });
                return;
              }
            }
            if (performance.now() > deadline) {
              reject(new Error("timed out waiting for committed row summary"));
            } else requestAnimationFrame(tick);
          };
          requestAnimationFrame(tick);
        },
      ),
    {
      expectedSummary: options.expectedSummary,
      startedAtMs,
      timeoutMs: options.timeoutMs,
    },
  );
}

export function findCommittedRowSummaryInRoot(
  root: ParentNode,
  options: { expectedSummary: string; surface: string },
): CommittedRowSummaryMatch | null {
  const grid = findElementByTestId(root, gridShellTestId(options.surface));
  if (grid === null) return null;
  const rows = Array.from(
    grid.querySelectorAll<HTMLElement>(gridSavedRowsSelector()),
  );
  for (const row of rows) {
    const recordId = row.getAttribute("data-grid-record-id");
    if (recordId === null || recordId.trim() === "") continue;
    const candidate = findElementByTestId<
      HTMLInputElement | HTMLTextAreaElement
    >(row, rowCellTestId(recordId, "timeline.activity_synopsis_text"));
    if (candidate?.value !== options.expectedSummary) continue;
    const rowVersionText = findElementByTestId(
      row,
      timelineRowVersionTestId(recordId),
    )?.textContent;
    const rowVersion = Number.parseInt(rowVersionText ?? "", 10);
    if (Number.isInteger(rowVersion) && rowVersion >= 1) {
      return { recordId, rowVersion };
    }
  }
  return null;
}

async function resetWorkbookClientTiming(page: Page) {
  await page.evaluate(() => {
    for (const entry of performance.getEntriesByType("mark")) {
      if (entry.name.startsWith("cartulary.workbook.")) {
        performance.clearMarks(entry.name);
      }
    }
  });
}

async function readWorkbookClientTiming(
  page: Page,
): Promise<ClientTimingEvent[]> {
  return page.evaluate<ClientTimingEvent[], string>(
    (markPrefix) =>
      performance
        .getEntriesByType("mark")
        .filter((entry) => entry.name.startsWith(markPrefix))
        .map((entry) => {
          const detail =
            "detail" in entry &&
            entry.detail !== null &&
            typeof entry.detail === "object"
              ? (entry.detail as Record<string, unknown>)
              : {};
          return {
            at: entry.startTime,
            name: entry.name.slice(markPrefix.length),
            ...detail,
          };
        }),
    workbookTimingMarkPrefix,
  );
}

function splitServerTimingHeader(header: string) {
  const parts: string[] = [];
  let current = "";
  let quoted = false;
  for (const character of header) {
    if (character === '"') {
      quoted = !quoted;
      current += character;
    } else if (character === "," && !quoted) {
      parts.push(current);
      current = "";
    } else current += character;
  }
  parts.push(current);
  return parts;
}

function unquoteServerTimingValue(value: string) {
  if (value.length >= 2 && value.startsWith('"') && value.endsWith('"')) {
    return value.slice(1, -1).replace(/\\"/gu, '"');
  }
  return value;
}

function findElementByTestId<T extends Element = HTMLElement>(
  root: ParentNode,
  testId: string,
): T | null {
  return (
    Array.from(root.querySelectorAll<T>("[data-testid]")).find(
      (element) => element.getAttribute("data-testid") === testId,
    ) ?? null
  );
}

function canonicalJSON(value: unknown): string {
  if (Array.isArray(value)) {
    return `[${value.map(canonicalJSON).join(",")}]`;
  }
  if (value !== null && typeof value === "object") {
    return `{${Object.entries(value)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, entry]) => `${JSON.stringify(key)}:${canonicalJSON(entry)}`)
      .join(",")}}`;
  }
  return JSON.stringify(value);
}
