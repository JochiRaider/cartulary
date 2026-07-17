import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  dataTestIdSelector,
  draftCellTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  timelineRowVersionTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";
import { timelineViewSchemaId } from "../support/contracts/workbookSurfaces";

export const ordinaryMeasurementSamplePolicy = {
  warmupSamples: 1,
  measuredSamples: 25,
  totalSamples: 26,
} as const;

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

type CommittedRowSummaryMatch = { recordId: string; rowVersion: number };
type WorkbookTimingProbeWindow = Window & {
  __cartularyWorkbookTimingProbe?: { events: ClientTimingEvent[] };
};

export async function measureTypingAck(
  page: Page,
  testId: string,
  appendedCharacter: string,
) {
  const input = page.getByTestId(testId);
  const currentValue = await input.inputValue();
  const completion = waitForInputValue(page, {
    testId,
    expectedValue: `${currentValue}${appendedCharacter}`,
    requireFocus: true,
    timeoutMs: 5_000,
  });
  await input.press(appendedCharacter);
  return completion;
}

export async function measureBlankRowCreate(
  page: Page,
  expectedSummary: string,
) {
  const draftSummary = page.getByTestId(
    draftCellTestId("timeline.activity_synopsis_text"),
  );
  await expect(draftSummary).toHaveValue(expectedSummary);
  await draftSummary.focus();
  await resetWorkbookClientTiming(page);
  const browserStart = await page.evaluate(() => performance.now());
  const completion = waitForCommittedRowSummary(page, {
    expectedSummary,
    startedAtMs: browserStart,
    surface: timelineViewSchemaId,
    timeoutMs: 5_000,
  });
  await draftSummary.press("Enter");
  let committed: Awaited<ReturnType<typeof waitForCommittedRowSummary>>;
  try {
    committed = await completion;
  } catch (error) {
    const clientTimingEvents = await readWorkbookClientTiming(page).catch(
      () => [],
    );
    throw new Error(
      `${error instanceof Error ? error.message : String(error)}\nClient timing events:\n${JSON.stringify(
        clientTimingEvents,
        null,
        2,
      )}`,
    );
  }
  const clientTimingEvents = await readWorkbookClientTiming(page);
  const responseEvent = clientTimingEvents.find(
    (event) => event.name === "pending_fetch_response",
  );
  const status =
    typeof responseEvent?.status === "number" ? responseEvent.status : 0;
  expect(status, JSON.stringify({ clientTimingEvents }, null, 2)).toBe(201);
  const serverTiming =
    typeof responseEvent?.serverTiming === "string"
      ? responseEvent.serverTiming
      : "";
  const networkDurationMs =
    typeof responseEvent?.at === "number"
      ? responseEvent.at - browserStart
      : Number.NaN;
  return {
    clientTimingEvents,
    committedDurationMs: committed.durationMs,
    networkDurationMs,
    recordId: committed.recordId,
    rowVersion: committed.rowVersion,
    serverTiming,
    serverTimingMetrics: parseServerTiming(serverTiming),
    status,
  };
}

export function percentile95(
  samples: number[],
  options: { minimumSampleCount?: number; sampleLabel?: string } = {},
) {
  const minimumSampleCount =
    options.minimumSampleCount ??
    ordinaryMeasurementSamplePolicy.measuredSamples;
  const sampleLabel = options.sampleLabel ?? "samples";
  if (samples.length === 0) {
    throw new Error("cannot compute percentile95 for an empty sample set");
  }
  if (samples.length < minimumSampleCount) {
    throw new Error(
      `cannot compute percentile95 for ${sampleLabel}: expected at least ${minimumSampleCount} samples, got ${samples.length}`,
    );
  }
  const sorted = [...samples].sort((left, right) => left - right);
  const index = Math.max(0, Math.ceil(sorted.length * 0.95) - 1);
  return sorted[index] ?? sorted[sorted.length - 1];
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
        if (attribute === "") {
          continue;
        }
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
    surface: WorkbookSurface;
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
    ({
      expectedSummary,
      gridTestId,
      rowVersionFieldKey,
      savedRowsSelector,
      startedAtMs,
      summaryFieldKey,
      timeoutMs,
    }) =>
      new Promise<{
        durationMs: number;
        recordId: string;
        rowVersion: number;
      }>((resolve, reject) => {
        const deadline = startedAtMs + timeoutMs;
        let animationFrame: number | null = null;
        let settled = false;
        const findByTestId = <T extends Element>(
          root: ParentNode,
          testId: string,
        ): T | null =>
          Array.from(root.querySelectorAll<T>("[data-testid]")).find(
            (element) => element.getAttribute("data-testid") === testId,
          ) ?? null;
        const rowCellTestIdFor = (recordId: string, fieldKey: string) =>
          `row-${recordId}-${fieldKey}`;
        const findMatch = (): CommittedRowSummaryMatch | null => {
          const grid = findByTestId<HTMLElement>(document, gridTestId);
          if (grid === null) {
            return null;
          }
          const rows = Array.from(
            grid.querySelectorAll<HTMLElement>(savedRowsSelector),
          );
          for (const row of rows) {
            const recordId = row.getAttribute("data-grid-record-id");
            if (recordId === null || recordId.trim() === "") {
              continue;
            }
            const candidate = findByTestId<HTMLElement>(
              row,
              rowCellTestIdFor(recordId, summaryFieldKey),
            );
            const candidateValue =
              candidate instanceof HTMLInputElement ||
              candidate instanceof HTMLTextAreaElement
                ? candidate.value
                : candidate?.textContent?.trim();
            if (candidateValue !== expectedSummary) {
              continue;
            }
            const rowVersionText =
              findByTestId(row, rowCellTestIdFor(recordId, rowVersionFieldKey))
                ?.textContent ??
              findByTestId(
                document,
                rowCellTestIdFor(recordId, rowVersionFieldKey),
              )?.textContent;
            const rowVersion = Number.parseInt(rowVersionText ?? "", 10);
            if (!Number.isInteger(rowVersion) || rowVersion < 1) {
              continue;
            }
            return { recordId, rowVersion };
          }
          return null;
        };
        const observer = new MutationObserver(() => checkForMatch());
        const finish = (match: CommittedRowSummaryMatch) => {
          if (settled) return;
          settled = true;
          if (animationFrame !== null) {
            window.cancelAnimationFrame(animationFrame);
          }
          observer.disconnect();
          const durationMs = performance.now() - startedAtMs;
          const target = window as WorkbookTimingProbeWindow;
          target.__cartularyWorkbookTimingProbe?.events.push({
            at: performance.now(),
            name: "committed_row_visible",
            recordId: match.recordId,
            rowVersion: match.rowVersion,
          });
          resolve({
            durationMs,
            recordId: match.recordId,
            rowVersion: match.rowVersion,
          });
        };
        const fail = () => {
          if (settled) return;
          settled = true;
          if (animationFrame !== null) {
            window.cancelAnimationFrame(animationFrame);
          }
          observer.disconnect();
          reject(
            new Error(
              `timed out waiting for committed row summary ${expectedSummary}`,
            ),
          );
        };
        const checkForMatch = () => {
          const match = findMatch();
          if (match !== null) {
            finish(match);
          } else if (performance.now() > deadline) {
            fail();
          }
        };
        const tick = () => {
          if (settled) return;
          checkForMatch();
          if (!settled) animationFrame = window.requestAnimationFrame(tick);
        };
        observer.observe(document.documentElement, {
          attributes: true,
          childList: true,
          subtree: true,
        });
        tick();
      }),
    {
      expectedSummary: options.expectedSummary,
      gridTestId: gridShellTestId(options.surface),
      rowVersionFieldKey: "row_version",
      savedRowsSelector: gridSavedRowsSelector(),
      startedAtMs,
      summaryFieldKey: "timeline.activity_synopsis_text",
      timeoutMs: options.timeoutMs,
    },
  );
}

export function findCommittedRowSummaryInRoot(
  root: ParentNode,
  options: { expectedSummary: string; surface: WorkbookSurface },
): CommittedRowSummaryMatch | null {
  const grid = findElementByTestId(root, gridShellTestId(options.surface));
  if (grid === null) {
    return null;
  }
  const rows = Array.from(
    grid.querySelectorAll<HTMLElement>(gridSavedRowsSelector()),
  );
  for (const row of rows) {
    const recordId = row.getAttribute("data-grid-record-id");
    if (recordId === null || recordId.trim() === "") {
      continue;
    }
    const candidate = findElementByTestId<
      HTMLInputElement | HTMLTextAreaElement
    >(row, rowCellTestId(recordId, "timeline.activity_synopsis_text"));
    if (candidate?.value !== options.expectedSummary) {
      continue;
    }
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
    const target = window as WorkbookTimingProbeWindow;
    target.__cartularyWorkbookTimingProbe = { events: [] };
  });
}

async function readWorkbookClientTiming(page: Page) {
  return page.evaluate(() => {
    const target = window as WorkbookTimingProbeWindow;
    return [...(target.__cartularyWorkbookTimingProbe?.events ?? [])];
  });
}

async function waitForInputValue(
  page: Page,
  options: {
    testId: string;
    expectedValue: string;
    requireFocus: boolean;
    timeoutMs: number;
  },
) {
  const start = await page.evaluate(() => performance.now());
  const selector = dataTestIdSelector(options.testId);
  return page.evaluate(
    ({ expectedValue, requireFocus, selector, startMark, testId, timeoutMs }) =>
      new Promise<number>((resolve, reject) => {
        const deadline = startMark + timeoutMs;
        const tick = () => {
          const element = document.querySelector(selector);
          const isTextInput =
            element instanceof HTMLInputElement ||
            element instanceof HTMLTextAreaElement;
          if (
            isTextInput &&
            element.value === expectedValue &&
            (!requireFocus || document.activeElement === element)
          ) {
            resolve(performance.now() - startMark);
          } else if (performance.now() > deadline) {
            reject(
              new Error(
                `timed out waiting for ${testId} to reach ${expectedValue}`,
              ),
            );
          } else {
            requestAnimationFrame(tick);
          }
        };
        requestAnimationFrame(tick);
      }),
    {
      expectedValue: options.expectedValue,
      requireFocus: options.requireFocus,
      selector,
      startMark: start,
      testId: options.testId,
      timeoutMs: options.timeoutMs,
    },
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
    } else {
      current += character;
    }
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
