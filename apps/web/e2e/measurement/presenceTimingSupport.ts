import {
  type CartularyPresencePredicateId,
  cartularyLargeGridFixture,
  cartularyPresencePerformanceContract as contract,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { ac043FixtureDigest, resetWorkbookClientTiming } from "./timingSupport";

export type PresenceKind = "surface" | "row" | "cell";
export type PresenceTarget = {
  kind: PresenceKind;
  actorUserId: string;
  markerSelector: string;
  recordId: string | null;
  fieldKey: string | null;
};
export type PresencePaintInput = PresenceTarget & {
  viewSchemaId: string;
  timeoutMs: number;
};
export type PresencePaint = {
  origin: number;
  stoppedAt: number;
  rowVersion: number | null;
};

// The same observer is serialized into Chromium and exercised by negative DOM fixtures.
export function observePresencePaintInBrowser(
  input: PresencePaintInput,
): Promise<PresencePaint> {
  const origin = performance.timeOrigin;
  return new Promise((resolve, reject) => {
    const deadline = performance.now() + input.timeoutMs;
    const glyph = () =>
      [...document.querySelectorAll<HTMLElement>(input.markerSelector)]
        .flatMap((marker) => [
          ...marker.querySelectorAll<HTMLElement>("[data-presence-user-id]"),
        ])
        .filter(
          (node) =>
            node.dataset.presenceUserId === input.actorUserId &&
            (input.kind === "surface"
              ? ["viewing", "focused", "editing", "idle"].includes(
                  node.dataset.presenceMode ?? "",
                )
              : node.dataset.presenceMode ===
                (input.kind === "row" ? "focused" : "editing")),
        );
    if (glyph().length !== 0) {
      reject(
        new Error("Presence target was already present before the action"),
      );
      return;
    }
    let previous: HTMLElement | null = null;
    let previousVersion: number | null = null;
    let frames = 0;
    let diagnostic = "no_glyph";
    const tick = () => {
      if (
        performance.timeOrigin !== origin ||
        innerWidth !== 1920 ||
        innerHeight !== 1080 ||
        devicePixelRatio !== 1
      ) {
        reject(
          new Error("Presence measurement clock origin or geometry changed"),
        );
        return;
      }
      const nodes = glyph();
      const node = nodes.length === 1 ? (nodes[0] ?? null) : null;
      const frame = [
        ...document.querySelectorAll<HTMLElement>(
          "[data-workbook-view-schema-id]",
        ),
      ].find(
        (value) => value.dataset.workbookViewSchemaId === input.viewSchemaId,
      );
      const row = node?.closest<HTMLElement>("[data-grid-record-id]");
      const rowVersion = row ? Number(row.dataset.gridRowVersion) : null;
      let ready = !!node && !!frame && node.isConnected;
      diagnostic = !node
        ? "no_unique_glyph"
        : !frame
          ? "no_view_frame"
          : "binding";
      if (input.kind !== "surface") {
        ready =
          ready &&
          !!row &&
          frame?.contains(row) === true &&
          row.dataset.gridRecordId === input.recordId &&
          Number.isSafeInteger(rowVersion) &&
          rowVersion !== null &&
          rowVersion > 0;
      }
      if (input.kind === "cell")
        ready =
          ready &&
          node?.closest<HTMLElement>("[data-grid-field-key]")?.dataset
            .gridFieldKey === input.fieldKey;
      if (ready && node) {
        diagnostic = "visibility";
        const bounds = node.getBoundingClientRect();
        let left = 0,
          top = 0,
          right = innerWidth,
          bottom = innerHeight;
        for (
          let ancestor: HTMLElement | null = node;
          ancestor;
          ancestor = ancestor.parentElement
        ) {
          const style = getComputedStyle(ancestor);
          if (
            ancestor.hidden ||
            ancestor.hasAttribute("inert") ||
            style.display === "none" ||
            style.visibility === "hidden" ||
            Number(style.opacity || 1) <= 0
          )
            ready = false;
          if (
            ancestor !== node &&
            style.display !== "contents" &&
            /auto|scroll|hidden|clip/u.test(
              `${style.overflow} ${style.overflowX} ${style.overflowY}`,
            )
          ) {
            const box = ancestor.getBoundingClientRect();
            left = Math.max(left, box.left);
            right = Math.min(right, box.right);
            top = Math.max(top, box.top);
            bottom = Math.min(bottom, box.bottom);
          }
        }
        if (input.kind !== "surface") {
          const grid = row?.closest<HTMLElement>(
            '[role="grid"], [role="treegrid"]',
          );
          if (!grid) ready = false;
          else {
            const box = grid.getBoundingClientRect();
            left = Math.max(left, box.left);
            right = Math.min(right, box.right);
            top = Math.max(top, box.top);
            bottom = Math.min(bottom, box.bottom);
          }
        }
        ready =
          ready &&
          bounds.width > 0 &&
          bounds.height > 0 &&
          bounds.left < right &&
          bounds.right > left &&
          bounds.top < bottom &&
          bounds.bottom > top &&
          (node.textContent?.trim().length ?? 0) > 0;
        if (!ready)
          diagnostic = `visibility:box=${[bounds.left, bounds.top, bounds.right, bounds.bottom].map(Math.round)};clip=${[left, top, right, bottom].map(Math.round)}`;
      }
      frames = ready
        ? previous === node && previousVersion === rowVersion
          ? frames + 1
          : 1
        : 0;
      previous = ready ? node : null;
      previousVersion = ready ? rowVersion : null;
      if (frames === 2) {
        resolve({ origin, stoppedAt: performance.now(), rowVersion });
        return;
      }
      if (performance.now() >= deadline) {
        reject(
          new Error(
            `Presence target did not qualify across two visible frames (${diagnostic})`,
          ),
        );
        return;
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  });
}

export function presenceElapsed(
  start: { origin: number; acceptedAt: number },
  stop: PresencePaint,
): number {
  const elapsed =
    stop.origin - start.origin + (stop.stoppedAt - start.acceptedAt);
  if (
    ![
      start.origin,
      start.acceptedAt,
      stop.origin,
      stop.stoppedAt,
      elapsed,
    ].every(Number.isFinite) ||
    start.origin <= 0 ||
    stop.origin <= 0 ||
    start.acceptedAt < 0 ||
    stop.stoppedAt < 0 ||
    elapsed < 0
  )
    throw new Error("Presence clocks do not define a valid monotonic interval");
  return elapsed;
}

export async function measurePresence(
  actor: Page,
  observer: Page,
  target: PresenceTarget,
  action: () => Promise<void>,
) {
  if (
    actor.context() === observer.context() ||
    actor.context().browser() !== observer.context().browser()
  )
    throw new Error(
      "Presence measurement requires distinct authenticated contexts in one browser",
    );
  await resetWorkbookClientTiming(actor);
  const actorOrigin = await actor.evaluate(() => performance.timeOrigin);
  const observed = observer
    .evaluate(observePresencePaintInBrowser, {
      ...target,
      viewSchemaId: contract.view_schema_id,
      timeoutMs: 5_000,
    })
    .then(
      (value) => ({ value }),
      (error) => ({ error }),
    );
  await action();
  const result = await observed;
  if ("error" in result) {
    const diagnostic = await Promise.all(
      [actor, observer].map((page) =>
        page.evaluate(
          ({ target, mark }) => {
            const selected = document.activeElement?.closest<HTMLElement>(
              "[data-grid-record-id]",
            );
            const glyphs = [
              ...document.querySelectorAll<HTMLElement>(
                "[data-presence-user-id]",
              ),
            ].filter(
              (node) => node.dataset.presenceUserId === target.actorUserId,
            );
            return {
              activeRowMatches:
                selected?.dataset.gridRecordId === target.recordId,
              targetRowMounted: [
                ...document.querySelectorAll<HTMLElement>(
                  "[data-grid-record-id]",
                ),
              ].some((node) => node.dataset.gridRecordId === target.recordId),
              actorGlyphs: glyphs.length,
              targetGlyphs: glyphs
                .filter(
                  (node) =>
                    node.closest<HTMLElement>("[data-grid-record-id]")?.dataset
                      .gridRecordId === target.recordId,
                )
                .map((node) => ({
                  mode: ["viewing", "focused", "editing", "idle"].includes(
                    node.dataset.presenceMode ?? "",
                  )
                    ? node.dataset.presenceMode
                    : "unknown",
                })),
              acceptedTargets: performance
                .getEntriesByName(mark, "mark")
                .map((entry) => {
                  const detail = (entry as PerformanceMark).detail;
                  return {
                    recordMatches: detail?.recordId === target.recordId,
                    fieldMatches: detail?.fieldKey === target.fieldKey,
                    mode: ["viewing", "editing", "idle"].includes(detail?.mode)
                      ? detail.mode
                      : "unknown",
                  };
                })
                .slice(0, 8),
            };
          },
          { target, mark: contract.start_mark },
        ),
      ),
    );
    throw new Error(
      `${result.error instanceof Error ? result.error.message : "Presence paint failed"}; bounded state=${JSON.stringify(diagnostic)}`,
    );
  }
  const start = await actor.evaluate(
    (mark) => ({
      origin: performance.timeOrigin,
      marks: performance.getEntriesByName(mark, "mark").map((value) => ({
        acceptedAt: value.startTime,
        detail: (value as PerformanceMark).detail,
      })),
    }),
    contract.start_mark,
  );
  const matching = start.marks.filter(
    (mark) =>
      mark.detail?.sheetRef?.kind === "view_schema" &&
      mark.detail.sheetRef.id === contract.view_schema_id &&
      mark.detail.recordId === target.recordId &&
      mark.detail.fieldKey === target.fieldKey &&
      mark.detail.mode === (target.kind === "cell" ? "editing" : "viewing"),
  );
  const mark = matching[0];
  if (
    start.origin !== actorOrigin ||
    matching.length !== 1 ||
    mark === undefined
  )
    throw new Error(
      "Presence acceptance mark is missing, repeated, or from a new document",
    );
  return {
    elapsed: presenceElapsed(
      { origin: start.origin, acceptedAt: mark.acceptedAt },
      result.value,
    ),
    rowVersion: result.value.rowVersion,
  };
}

export async function attachPresenceObservation(
  testInfo: TestInfo,
  predicateId: CartularyPresencePredicateId,
  target?: PresenceTarget,
  completion?: { elapsed: number; rowVersion: number | null },
  failureReason?: string,
) {
  const elapsed = completion?.elapsed ?? null;
  const observation = {
    schema_id: "cartulary.frontend_measurement_observation.v4",
    fixture_profile_id: "timeline_large_grid_snapshot_v2",
    criterion_id: contract.criterion_ids[0],
    predicate_id: predicateId,
    fixture_id: cartularyLargeGridFixture.fixtureId,
    fixture_digest: ac043FixtureDigest(),
    measurement_policy_id: contract.measurement_policy_id,
    threshold_ms: contract.threshold_ms,
    warmup_samples: 0,
    measured_samples: completion ? 1 : 0,
    outcome:
      process.env.CARTULARY_BROWSER_RESOURCE_PROFILE_ID !==
      "browser_measurement_quiet"
        ? "environment_not_qualified"
        : failureReason
          ? "incomplete"
          : elapsed === null
            ? "incomplete"
            : elapsed <= contract.threshold_ms
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
    statistics: { kind: "single", elapsed_ms: elapsed },
    predicate_evidence:
      completion && target
        ? {
            schema_id: "cartulary.presence_measurement_proof.v1",
            kind: "timeline_presence",
            presence_kind: target.kind,
            actor_user_id: target.actorUserId,
            view_schema_id: contract.view_schema_id,
            record_id: target.recordId,
            row_version: completion.rowVersion,
            field_key: target.fieldKey,
            consecutive_frames: 2,
            clock_domain: "browser_monotonic_time_origin",
            origins_unchanged: true,
            initially_absent: true,
            binding_stable: true,
            action_count: 1,
            viewport_width: 1920,
            viewport_height: 1080,
            device_scale_factor: 1,
          }
        : null,
    ...(failureReason
      ? { failure_reason: "Presence measurement did not complete" }
      : {}),
  };
  await testInfo.attach(
    `cartulary.frontend_measurement_observation.v4.${predicateId}`,
    {
      body: JSON.stringify(observation),
      contentType: "application/json",
    },
  );
  return observation;
}
