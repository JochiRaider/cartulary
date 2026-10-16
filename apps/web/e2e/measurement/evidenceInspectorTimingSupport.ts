import { createHash } from "node:crypto";
import {
  cartularyEvidenceInspectorContract as contract,
  dataTestIdSelector,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
} from "@cartulary/ui-contracts";
import type { Page, Request, TestInfo } from "@playwright/test";
import {
  type EvidenceInspectorTarget,
  evidenceInspectorProfileId,
} from "../support/performance/evidenceInspectorSnapshot";
import { readHttpOperationResponse } from "../support/transport/publicHttpOperationClient";

export type EvidenceInspectorPaint = {
  accepted_at: number;
  first_frame_at: number;
  second_frame_at: number;
  metadata_started_at: number;
  metadata_completed_at: number;
};
export type EvidenceInspectorSample = EvidenceInspectorPaint & {
  sample_index: number;
  warmup: boolean;
  total_ms: number;
};
export type EvidencePaintOptions = {
  inspectorSelector: string;
  evidenceSelector: string;
  metadataPath: string;
  target: EvidenceInspectorTarget;
  startMark: string;
  stopMark: string;
  timeoutMs: number;
};

// Serialized into Chromium. Product acceptance, resource observation and paint
// proof share its monotonic clock; fixture setup and Node assertions do not.
export function observeEvidenceInspectorPaintInBrowser(
  options: EvidencePaintOptions,
): Promise<EvidenceInspectorPaint> {
  return new Promise((resolve, reject) => {
    if (
      options.target.items.length !== 5 ||
      new Set(options.target.items.map((item) => item.record_id)).size !== 5 ||
      !Number.isSafeInteger(options.target.row_version) ||
      options.target.row_version < 1 ||
      options.target.items.some(
        (item) =>
          !Number.isSafeInteger(item.row_version) ||
          item.row_version < 1 ||
          !item.title,
      )
    ) {
      reject(
        new Error("Evidence inspector requires five exact stable bindings"),
      );
      return;
    }
    const deadline = performance.now() + options.timeoutMs;
    let prior: HTMLElement[] = [],
      firstFrame: number | null = null;
    let visibilityReason = "none";
    let visibilityGeometry = "none";
    const visible = (element: HTMLElement, inspector: HTMLElement) => {
      if (
        !element.isConnected ||
        element.closest("[hidden],[aria-hidden='true']")
      ) {
        visibilityReason = "detached_or_hidden";
        return false;
      }
      const rect = element.getBoundingClientRect(),
        clip = inspector.getBoundingClientRect();
      let left = Math.max(0, clip.left),
        right = Math.min(innerWidth, clip.right),
        top = Math.max(0, clip.top),
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
        ) {
          visibilityReason = "style_hidden";
          return false;
        }
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
      const intersects =
        rect.width > 0 &&
        rect.height > 0 &&
        right > left &&
        bottom > top &&
        rect.right > left &&
        rect.left < right &&
        rect.bottom > top &&
        rect.top < bottom;
      if (!intersects) {
        visibilityReason =
          right <= left || bottom <= top ? "empty_clip" : "outside_clip";
        visibilityGeometry = [
          rect.left,
          rect.top,
          rect.right,
          rect.bottom,
          left,
          top,
          right,
          bottom,
        ]
          .map((value) => Math.round(value))
          .join(",");
      }
      return intersects;
    };
    const tick = () => {
      if (
        innerWidth !== 1920 ||
        innerHeight !== 1080 ||
        devicePixelRatio !== 1
      ) {
        reject(new Error("Evidence inspector geometry is not qualified"));
        return;
      }
      const marks = performance.getEntriesByName(options.startMark, "mark");
      if (marks.length > 1) {
        reject(new Error("Evidence inspector action was repeated"));
        return;
      }
      const start = marks[0];
      const resources = performance.getEntriesByType(
        "resource",
      ) as PerformanceResourceTiming[];
      const metadata = resources.filter(
        (entry) =>
          new URL(entry.name, location.href).pathname === options.metadataPath,
      );
      if (
        metadata.length > 1 ||
        resources.some((entry) =>
          /\/evidence-(?:records\/[^/]+\/(?:preview|download)-handle|handles\/)/u.test(
            new URL(entry.name, location.href).pathname,
          ),
        )
      ) {
        reject(
          new Error(
            `Evidence inspector measured request set is contaminated: metadata=${metadata.length}, access=${resources.filter((entry) => /\/evidence-(?:records\/[^/]+\/(?:preview|download)-handle|handles\/)/u.test(new URL(entry.name, location.href).pathname)).length}`,
          ),
        );
        return;
      }
      const read = metadata[0];
      if (
        start &&
        read &&
        (read.startTime < start.startTime || read.responseEnd < read.startTime)
      ) {
        reject(new Error("Evidence inspector metadata predates acceptance"));
        return;
      }
      const inspector = document.querySelector<HTMLElement>(
        options.inspectorSelector,
      );
      const parts: HTMLElement[] = [];
      let condition = !start
        ? "acceptance_missing"
        : !read
          ? "metadata_missing"
          : "target_binding";
      if (
        start &&
        read &&
        read.responseEnd > 0 &&
        inspector?.dataset.recordId === options.target.record_id &&
        inspector.dataset.rowVersion === String(options.target.row_version)
      ) {
        condition = "heading_or_panel";
        const heading = inspector.querySelector<HTMLElement>("h2");
        const panel = inspector.querySelector<HTMLElement>(
          options.evidenceSelector,
        );
        if (heading?.textContent?.trim() === options.target.summary && panel) {
          condition = "counts_or_window";
          const counts = [...panel.querySelectorAll<HTMLElement>("p")].filter(
            (element) =>
              ["Attached evidence count: 100", "Linked evidence: 100"].includes(
                element.textContent?.trim() ?? "",
              ),
          );
          const entries = [
            ...panel.querySelectorAll<HTMLElement>("[data-evidence-record-id]"),
          ];
          if (counts.length === 2 && entries.length === 5) {
            condition = "item_binding";
            parts.push(heading, ...counts);
            for (const [index, item] of options.target.items.entries()) {
              const entry = entries[index];
              const paragraphs = entry
                ? [...entry.querySelectorAll<HTMLElement>("p")]
                : [];
              const preview = entry?.querySelector<HTMLButtonElement>(
                "button[aria-label^='Preview ']",
              );
              if (
                !entry ||
                entry.dataset.evidenceRecordId !== item.record_id ||
                entry.dataset.evidenceRowVersion !== String(item.row_version) ||
                entry.dataset.evidenceMetadataState !== "ready" ||
                paragraphs[0]?.textContent?.trim() !== item.title ||
                paragraphs[1]?.textContent?.trim() !==
                  "Available · File available" ||
                !preview ||
                preview.disabled ||
                preview.getAttribute("aria-disabled") === "true"
              ) {
                parts.length = 0;
                break;
              }
              parts.push(entry, ...paragraphs, preview);
            }
          }
        }
      }
      let invisiblePart = -1;
      if (inspector && parts.length > 3) {
        invisiblePart = parts.findIndex((part) => !visible(part, inspector));
        condition = invisiblePart < 0 ? "second_frame" : "visibility";
      }
      const complete = !!inspector && parts.length > 3 && invisiblePart < 0;
      const stable =
        complete &&
        parts.length === prior.length &&
        parts.every((part, index) => part === prior[index]);
      const now = performance.now();
      if (
        start &&
        read &&
        stable &&
        firstFrame !== null &&
        now > firstFrame &&
        read.responseEnd <= firstFrame
      ) {
        performance.mark(options.stopMark);
        resolve({
          accepted_at: start.startTime,
          first_frame_at: firstFrame,
          second_frame_at: now,
          metadata_started_at: read.startTime,
          metadata_completed_at: read.responseEnd,
        });
        return;
      }
      prior = complete ? parts : [];
      firstFrame = complete ? now : null;
      if (now >= deadline) {
        reject(
          new Error(
            `Evidence inspector did not reach exact visible completion: condition=${condition}, part=${invisiblePart}, visibility=${visibilityReason}, geometry=${visibilityGeometry}`,
          ),
        );
        return;
      }
      requestAnimationFrame(tick);
    };
    requestAnimationFrame(tick);
  });
}

export async function measureEvidenceInspectorPaint(
  page: Page,
  incidentId: string,
  target: EvidenceInspectorTarget,
  action: () => Promise<void>,
) {
  await page.evaluate(() => {
    performance.clearMarks();
    performance.clearMeasures();
    performance.clearResourceTimings();
  });
  const metadataPath = `/api/v1/incidents/${incidentId}/evidence-metadata`;
  let reads = 0,
    accessRequests = 0;
  const observe = (request: Request) => {
    const path = new URL(request.url()).pathname;
    if (path === metadataPath) {
      reads++;
      if (request.method() !== "POST") reads++;
    }
    if (
      /\/evidence-(?:records\/[^/]+\/(?:preview|download)-handle|handles\/)/u.test(
        path,
      )
    )
      accessRequests++;
  };
  page.on("request", observe);
  try {
    const response = page.waitForResponse(
      (response) => new URL(response.url()).pathname === metadataPath,
      { timeout: 10000 },
    );
    const completed = page.evaluate(observeEvidenceInspectorPaintInBrowser, {
      inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
      evidenceSelector: dataTestIdSelector(
        timelineInspectorSectionTestId("evidence"),
      ),
      metadataPath,
      target,
      startMark: contract.start_mark,
      stopMark: contract.stop_mark,
      timeoutMs: 10000,
    });
    const [paint, read] = await Promise.all([completed, response, action()]);
    const body = await readHttpOperationResponse(read, "queryEvidenceMetadata");
    if (
      reads !== 1 ||
      accessRequests !== 0 ||
      body.data.items.length !== target.items.length ||
      body.data.items.some((item, index) => {
        const expected = target.items[index];
        return (
          !expected ||
          item.record_id !== expected.record_id ||
          item.row_version !== expected.row_version ||
          item.title !== expected.title ||
          item.lifecycle_state !== "available" ||
          item.upload_state !== "available" ||
          !item.has_blob ||
          !item.preview_eligibility.eligible
        );
      })
    )
      throw new Error(
        "Evidence inspector fresh metadata or access request proof failed",
      );
    return paint;
  } finally {
    page.off("request", observe);
  }
}

function canonicalJSON(value: unknown): string {
  if (Array.isArray(value)) return `[${value.map(canonicalJSON).join(",")}]`;
  if (value !== null && typeof value === "object")
    return `{${Object.entries(value)
      .sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0))
      .map(([key, entry]) => `${JSON.stringify(key)}:${canonicalJSON(entry)}`)
      .join(",")}}`;
  return JSON.stringify(value);
}
export function evidenceInspectorFixtureDigest() {
  return `sha256:${createHash("sha256").update(canonicalJSON(contract.fixture)).digest("hex")}`;
}

export async function attachEvidenceInspectorObservation(
  info: TestInfo,
  samples: EvidenceInspectorSample[],
  target?: EvidenceInspectorTarget,
  failure?: string,
) {
  const values = samples
    .filter((sample) => !sample.warmup)
    .map((sample) => sample.total_ms)
    .sort((a, b) => a - b);
  const complete =
    samples.length === 101 && values.length === 100 && !!target && !failure;
  const p50 = values.length ? values[Math.ceil(values.length * 0.5) - 1] : null;
  const p95 = values.length
    ? values[Math.ceil(values.length * 0.95) - 1]
    : null;
  const quiet =
    process.env.CARTULARY_BROWSER_RESOURCE_PROFILE_ID ===
    "browser_measurement_quiet";
  const outcome = !quiet
    ? "environment_not_qualified"
    : !complete
      ? "incomplete"
      : p95 !== null && p95 !== undefined && p95 <= contract.threshold_ms
        ? "passed"
        : "threshold_failed";
  const observation = {
    schema_id: "cartulary.frontend_measurement_observation.v4",
    fixture_profile_id: evidenceInspectorProfileId,
    criterion_id: contract.criterion_id,
    predicate_id: contract.predicate_id,
    fixture_id: contract.fixture.fixture_id,
    fixture_digest: evidenceInspectorFixtureDigest(),
    measurement_policy_id: contract.measurement_policy_id,
    threshold_ms: contract.threshold_ms,
    warmup_samples: samples.filter((sample) => sample.warmup).length,
    measured_samples: values.length,
    statistics: {
      kind: "percentile",
      percentile: 95,
      p50_ms: p50,
      p95_ms: p95,
    },
    outcome,
    traffic: {
      counts: [
        { fact_id: "analyst_sessions", value: 1 },
        { fact_id: "background_sessions", value: 0 },
      ],
      rates: [],
      conditions: [],
    },
    samples: samples.map(({ sample_index, warmup, total_ms }) => ({
      sample_index,
      warmup,
      total_ms,
      stages_ms: [{ fact_id: "accepted_to_visible_paint", value: total_ms }],
    })),
    predicate_evidence:
      complete && target
        ? {
            schema_id: "cartulary.evidence_inspector_measurement_proof.v1",
            kind: "evidence_inspector",
            consecutive_frames: 2,
            all_content_visible: true,
            bindings_stable: true,
            fresh_metadata_each_open: true,
            target: {
              record_id: target.record_id,
              row_version: target.row_version,
            },
            items: target.items.map(({ record_id, row_version }) => ({
              record_id,
              row_version,
            })),
            viewport_width: 1920,
            viewport_height: 1080,
            device_scale_factor: 1,
            owning_viewport: "inspector",
            preview_preflight_records: 100,
            preview_preflight_body_verification: "complete_size_and_sha256",
            measured_access_handle_requests: 0,
            samples,
          }
        : null,
    ...(failure ? { failure_reason: failure.slice(0, 512) } : {}),
  };
  await info.attach(
    `cartulary.frontend_measurement_observation.v4.${contract.predicate_id}`,
    {
      body: Buffer.from(JSON.stringify(observation)),
      contentType: "application/json",
    },
  );
  return observation;
}
