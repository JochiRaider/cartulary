import {
  cartularyLargeGridFixture,
  cartularyAnchorTraceContract as contract,
} from "@cartulary/ui-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { ac043FixtureDigest } from "./timingSupport";

type Step = (typeof contract.steps)[number];
export type AnchorTraceInput = {
  recordId: string;
  initialVersion: number;
  gridSelector: string;
  editorSelector: string;
  draftValue: string;
  totalRows: number;
};
export type AnchorTraceProof = {
  schema_id: "cartulary.anchor_trace_proof.v1";
  kind: "timeline_anchor_trace";
  record_id: string;
  request_record_id: string | null;
  initial_row_version: number;
  committed_row_version: number | null;
  final_row_version: number | null;
  max_mounted_rows: number;
  frame_count: number;
  grid_replacements: number;
  violations: string[];
  checkpoints: {
    step: Step;
    frame_count: number;
    row_version: number;
    mounted_rows: number;
  }[];
};

declare global {
  interface Window {
    cartularyAnchorTrace?: {
      draft: boolean;
      checkpoint: (step: Step) => void;
      read: () => AnchorTraceProof;
      stop: () => void;
    };
  }
}

// This function is serialized into Chromium. All observation mechanics live in
// its closure; it reads the DOM but never changes the product or its rendering.
export function installAnchorTrace(input: AnchorTraceInput) {
  if (window.cartularyAnchorTrace)
    throw new Error("Anchor trace already installed");
  let root = document.querySelector(input.gridSelector);
  if (!root) throw new Error("Anchor trace has no grid");
  const proof: AnchorTraceProof = {
    schema_id: "cartulary.anchor_trace_proof.v1",
    kind: "timeline_anchor_trace",
    record_id: input.recordId,
    request_record_id: null,
    initial_row_version: input.initialVersion,
    committed_row_version: null,
    final_row_version: null,
    max_mounted_rows: 0,
    frame_count: 0,
    grid_replacements: 0,
    violations: [],
    checkpoints: [],
  };
  const versions = new Map<string, number>([
    [input.recordId, input.initialVersion],
  ]);
  let frame = 0,
    lastCheckpoint = 0,
    mountedRows = 0,
    rowVersion = input.initialVersion,
    stopped = false;
  const fail = (code: string) => {
    if (!proof.violations.includes(code)) proof.violations.push(code);
  };
  const control = {
    draft: true,
    checkpoint(step: Step) {
      const count = proof.frame_count - lastCheckpoint;
      if (count < 2)
        throw new Error("Anchor checkpoint requires two observed frames");
      proof.checkpoints.push({
        step,
        frame_count: count,
        row_version: rowVersion,
        mounted_rows: mountedRows,
      });
      lastCheckpoint = proof.frame_count;
    },
    read: () => structuredClone(proof),
    stop() {
      stopped = true;
      cancelAnimationFrame(frame);
    },
  };
  const observe = () => {
    if (stopped) return;
    proof.frame_count++;
    const currentRoot = document.querySelector(input.gridSelector);
    if (!currentRoot) {
      fail("grid_missing");
      frame = requestAnimationFrame(observe);
      return;
    }
    if (currentRoot !== root) proof.grid_replacements++;
    root = currentRoot;
    const rows = root.querySelectorAll<HTMLElement>(
      '[role="row"][data-grid-record-id]',
    );
    mountedRows = rows.length;
    proof.max_mounted_rows = Math.max(proof.max_mounted_rows, mountedRows);
    if (mountedRows >= input.totalRows) fail("full_sheet_mounted");
    let found = false;
    for (const row of rows) {
      const id = row.getAttribute("data-grid-record-id");
      const version = Number(row.getAttribute("data-grid-row-version"));
      if (!id || !Number.isSafeInteger(version) || version < 1) {
        fail("row_binding_invalid");
        continue;
      }
      if (version < (versions.get(id) ?? 0)) fail("row_version_regressed");
      versions.set(id, Math.max(version, versions.get(id) ?? 0));
      if (id === input.recordId) {
        found = true;
        rowVersion = version;
      }
    }
    // Virtualization may unmount an accepted row outside the viewport. An active
    // editor must remain pinned, even when its row is scrolled out of view.
    if (control.draft && !found) fail("anchor_missing");
    const editors = document.querySelectorAll<
      HTMLInputElement | HTMLTextAreaElement
    >(input.editorSelector);
    // Pending commits may retain the editor. Follow its semantic binding even
    // after submission, instead of assuming the control became display text.
    if (control.draft || editors.length > 0) {
      const editor = editors[0];
      if (
        editors.length !== 1 ||
        editor
          ?.closest("[data-grid-record-id]")
          ?.getAttribute("data-grid-record-id") !== input.recordId
      )
        fail("editor_retargeted");
      if (editor?.value !== input.draftValue) fail("draft_changed");
    }
    frame = requestAnimationFrame(observe);
  };
  window.cartularyAnchorTrace = control;
  frame = requestAnimationFrame(observe);
}

export async function anchorCheckpoint(page: Page, step: Step) {
  await page.evaluate(async (name) => {
    await new Promise<void>((resolve) =>
      requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
    );
    const trace = window.cartularyAnchorTrace;
    if (!trace) throw new Error("Anchor trace is not installed");
    trace.checkpoint(name);
  }, step);
}

export async function attachAnchorTrace(
  info: TestInfo,
  proof: AnchorTraceProof | null,
  failureReason?: string,
) {
  const observation = {
    schema_id: "cartulary.frontend_measurement_observation.v4",
    fixture_profile_id: "timeline_large_grid_snapshot_v2",
    criterion_id: contract.criterion_id,
    predicate_id: contract.predicate_id,
    fixture_id: cartularyLargeGridFixture.fixtureId,
    fixture_digest: ac043FixtureDigest(),
    measurement_policy_id: contract.measurement_policy_id,
    threshold_ms: null,
    warmup_samples: 0,
    measured_samples: 0,
    samples: [],
    statistics: {
      kind: "seeded_trace",
      completed_steps: proof?.checkpoints.length ?? 0,
      expected_steps: contract.steps.length,
    },
    outcome:
      process.env.CARTULARY_BROWSER_RESOURCE_PROFILE_ID !==
      "browser_measurement_quiet"
        ? "environment_not_qualified"
        : proof?.violations.length
          ? "predicate_failed"
          : failureReason || proof?.checkpoints.length !== contract.steps.length
            ? "incomplete"
            : "passed",
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
    predicate_evidence: proof,
    ...(failureReason ? { failure_reason: failureReason } : {}),
  };
  await info.attach(`${observation.schema_id}.${contract.predicate_id}`, {
    body: JSON.stringify(observation),
    contentType: "application/json",
  });
  return observation;
}
