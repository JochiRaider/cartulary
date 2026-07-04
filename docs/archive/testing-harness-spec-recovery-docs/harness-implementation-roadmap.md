---
doc_id: THR-S7-HARNESS-IMPLEMENTATION-ROADMAP
title: Testing Harness Implementation Roadmap
status: active
role: implementation-roadmap
---

# Testing Harness Implementation Roadmap

## Document role

This roadmap records the practical implementation sequence for applying the
2026-05-09 maintainer decisions. It is sequenced so documentation authority is
set first, implementation changes follow, and verification closes with generated
artifact checks.

## Phase 0: Preflight and evidence lock

Goal: record maintainer decisions, preserve unresolved source limits, and name
S7 deliverables before behavior changes.

Concrete changes: add `maintainer-decision-summary-2026-05-09.md`; name the
S7 NLSpec, acceptance matrix, roadmap, and review packet paths; carry forward
`SL-*`, `AMB-*`, `AUTH-*`, `PRES-*`, `MSC-*`, `FAIL-*`, and `TODO:*_unknown`
rows.

Dependencies: S0-S6 audit register, source-limit log, authority map,
preservation matrix, and selected S7 runtime evidence.

Risks: closing source limits by wording or treating WSL/Linux observations as a
complete platform matrix.

Validation and exit: decided rows cite maintainer decisions; unresolved rows
remain labeled; deliverable paths exist.

Still limited: env precedence, visual refresh bounds, parent-death cleanup,
active DB cleanup, CI provider details, and final evidence-gated `MUST` text.

## Phase 1: Harness contract draft

Goal: draft the harness contract before relying on implementation details.

Concrete changes: create `harness-nlspec.md`; document Make authority,
local-dev verification behavior, logical scheduler lanes, downstream generated
artifacts, source-observed env/defaults, observed WSL/Linux profile, retained
artifact identity, cleanup tiers, snapshot validation-only status, and
provider-neutral CI.

Dependencies: Phase 0.

Risks: over-specifying env precedence, local-dev cleanup, direct package
scripts, or provider-specific CI behavior.

Validation and exit: every candidate requirement has an evidence label or a
source-limit blocker; NLSpec contains scope, non-goals, authority, command,
local-dev, scheduler, env, artifacts, cleanup, CI, and acceptance sections.

Still limited: env precedence, non-Linux guarantees, provider CI behavior, and
direct package-script first-class status.

## Phase 2: Reset route ownership move

Goal: move the harness-owned reset route out of production application ownership
while preserving browser reset behavior.

Concrete changes: move bootstrap restoration support to
`internal/platform/bootstrap`; move reset route implementation and tests to
`internal/testutil/testruntime`; keep `cmd/server` as wiring only; preserve
`CARTULARY_ENABLE_TEST_ROUTES=1`; keep reset schema
`cartulary.test.runtime_reset.v1`.

Affected files: `internal/app/runtime.go`,
`internal/platform/bootstrap/**`, `internal/testutil/testruntime/**`,
`cmd/server/main.go`, and related tests.

Dependencies: Phase 1 reset-route authority wording.

Risks: import cycles, accidental product API exposure, or browser E2E reset
breakage.

Validation and exit: default reset route returns 404; enabled test route resets
DB/object state and restores bootstrap admin; no reset implementation remains
in `internal/app`.

Still limited: product/public reset API semantics and partial reset failure
semantics.

## Phase 3: Cleanup and destructive safety

Goal: encode cleanup as best-effort except where selected evidence proves
stronger behavior, and constrain stale janitors.

Concrete changes: document cleanup guarantee tiers; require generated name,
harness metadata or lease, conservative age or completed-summary evidence, and
scope-limited resource type before stale deletion; keep external Go caches out
of default cleanup; preserve parent-death and active DB cleanup source limits.

Affected subsystems: `tools/testservices/**`,
`internal/testutil/suiteservices/**`, `internal/testutil/pgtest/**`,
`internal/testutil/s3test/**`, Make cleanup docs, and recovery docs.

Dependencies: Phase 1 cleanup wording and selected cleanup evidence.

Risks: over-broad deletion rules, treating delayed after-state as synchronous
reaper completion, or adding `/tmp/cartulary-go-*` cleanup.

Validation and exit: janitor eligibility tests or static assertions prove proof
gates; docs distinguish best-effort from guaranteed cleanup; Go caches remain
out of `clean` and `distclean`.

Still limited: parent-death cleanup, active DB cleanup, and detached reaper
completion as a hard guarantee.

## Phase 4: Artifact, fixture, snapshot, and schema stabilization

Goal: make maintenance workflows safe without inventing unsupported refresh
authority.

Concrete changes: define fixture/golden refresh workflow; keep visual snapshots
validation-only; list known schema IDs as stable; preserve Playwright internals,
CI provider annotations, and shell log contents as tool-defined or unknown.

Affected files: `structured-output-schema-notes.md`,
`observable-interface-map.md`, `artifact-ownership-matrix.md`,
`harness-acceptance-matrix.md`, and snapshot docs only.

Dependencies: Phase 1 artifact authority.

Risks: blessing snapshot updates without platform/browser bounds, or treating
external tool report contents as harness-owned.

Validation and exit: no snapshot baseline files change; unknown schemas stay
explicitly labeled; refresh workflow is binary and reviewable.

Still limited: visual refresh OS/browser/version/command, CI provider
annotations, and Playwright tool report internals.

## Phase 5: CI and stale harness smoke demotion

Goal: make phase advancement independent of the stale extended harness smoke
failure while preserving useful smoke checks.

Concrete changes: remove `run-harness-smoke-extended` from blocking `ci` and
`release-check` sequence steps in `tools/execution_topology_manifest.json`;
regenerate downstream task-surface outputs through the generator; keep extended
smoke as an explicit diagnostic target; update sequence tests.

Affected files: `tools/execution_topology_manifest.json`,
`tools/task_surface_manifest.json`,
`tools/execution_topology_render_index.json`, and
`scripts/test-run-make-sequence.sh`.

Dependencies: generated-artifact authority from Phase 1 and maintainer decision
`MD-S7-0016`.

Risks: hand-editing generated outputs, removing useful fast smoke, or hiding
release readiness failures.

Validation and exit: `make ci` and `make release-check` no longer include
`run-harness-smoke-extended` as blocking sequence work; the diagnostic target
remains runnable.

Still limited: provider CI workflows/annotations and release readiness beyond
smoke demotion.

## Phase 6: Acceptance matrix and review packet

Goal: bind the contract and implementation to binary validation and maintainer
handoff.

Concrete changes: create `harness-acceptance-matrix.md`; create
`harness-review-packet.md`; list implemented decisions, deferred source limits,
generated artifact boundaries, CI/smoke disposition, reset-route move, cleanup
proof rules, and fixture/snapshot workflows.

Dependencies: Phases 1 through 5.

Risks: vague acceptance criteria or missing source-limited blockers.

Validation and exit: every final normative requirement has a criterion or
source-limit/owner-required blocker; Phase 7 and Phase 8 remain future planned
work only.

Still limited: all unresolved Phase 0 items unless later decisions or selected
evidence close them.

## Phase 7: Verification and finalization

Goal: verify without masking generated-output ownership or drift.

Concrete changes: run `make agent-finalize`, targeted reset and cleanup tests,
`make generated-artifact-policy-check`, `make json-shape-check`,
`make task-surface-report TASK_SURFACE_REPORT_ARGS=--all`, `make test-fast`,
`make ci`, and `make release-check` if CI passes or release readiness is being
claimed.

Dependencies: all implementation phases.

Risks: generated outputs changing unexpectedly, release readiness remaining
blocked by unrelated release artifacts, or local Docker/Playwright readiness
being unavailable.

Validation and exit: no hand-edited generated Go/TS/task artifacts; no
hand-edited lockfiles; CI behavior matches provider-neutral contract; any
failure is classified as blocking, source-limited, or non-blocking roadmap work.

Still limited: any source-limited item that lacks selected evidence or owner
decision after verification.

