# Cartulary Visual Golden Maintenance Guide

**Status**: Implementation-support guide
**Authority**: Core 00-04 own product behavior. `docs/testing-harness-nlspec.md` owns harness conformance when adopted. This guide does not promote visual snapshot refresh into current harness conformance.

## Purpose

Visual goldens are committed validation inputs for browser-rendered workbook states. They help detect UI drift in `browser-e2e-visual`, but they are not product behavior owners and they are not claim-bearing evidence by themselves.

## Canonical Surface

- Use `make browser-e2e-visual` for the canonical validation target.
- Visual workbook tests live in `apps/web/e2e/workbook.visual.spec.ts`.
- Committed Playwright goldens live beside the spec under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
- Retained actual/diff artifacts from failed runs live under the run root reported by the harness, usually `.cartulary/test-results/<run-id>/.../playwright-output/`.

## Accepted Refresh Triggers

Refresh a golden only when at least one of these is true:

- the UI contract intentionally changed;
- the visual harness intentionally changed viewport, masking, scroll normalization, or screenshot scope;
- a dependency, browser, or platform pin changed and the rendered output change is accepted;
- the previous golden is stale relative to already-validated functional behavior.

Do not refresh a golden to hide an unexplained product regression, broken functional assertion, unstable dynamic text, missing data, or a browser/runtime mismatch.

## Workbook Grid Scroll Contract

Every workbook-grid visual snapshot must declare its scroll normalization through the test helper. For `GridVisualScrollState.left`:

- `"left"` means `scrollLeft = 0`.
- `"right"` means the live computed maximum, `scrollWidth - clientWidth`, clamped at zero.
- a number means that exact requested offset, clamped into the live scroll range.

Use `"right"` for snapshots whose purpose is to validate far-right columns, badges, or status fields. Do not encode historical pixel offsets unless the specific pixel position is the behavior under test.

## Review Expectations

When visual normalization helpers change, the handoff or pull request must state:

- which helper behavior changed;
- which snapshots were affected;
- which goldens were regenerated, or why none were affected;
- whether the visual diff is explained by the intended contract change.

Skipped work units after a `browser-e2e-visual/visual` failure should be treated as cascade skips unless their own summaries show independent failures. Scheduler `resource:host_cpu`, `resource:host_io`, and dependency blocker counts are scheduling metadata, not root-cause failures.

## Validation Checklist

Before handoff:

- run the narrow Playwright update flow for the affected visual test;
- inspect the generated diff and confirm it matches the intended visual contract;
- run `make browser-e2e-visual`;
- run `make agent-finalize`, passing `RESULTS_DIR=<successful retained run root>` when a successful retained run should refresh and validate timing maintenance inputs;
- report whether `agent-finalize` ran unchanged, updated generated artifacts, skipped retained-run maintenance because `RESULTS_DIR` was unset, or failed.

`RESULTS_DIR` must identify a successful, uncontaminated retained run. A failed run, including a `make check` run whose summary reports `failure_class=product`, is invalid finalizer input; running `make agent-finalize` without `RESULTS_DIR` validates only the non-retained-run maintenance path.
