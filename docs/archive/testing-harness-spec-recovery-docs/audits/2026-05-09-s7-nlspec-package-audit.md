---
doc_id: THR-S7-NLSPEC-PACKAGE-AUDIT-2026-05-09
title: S7 NLSpec Package Audit
status: complete
role: audit
---

# S7 NLSpec Package Audit

## Audit objective

Audit whether S7 converted the completed S0 through S6 recovery work, S8
authority and preservation follow-up work, and the 2026-05-09 maintainer
decisions into a reviewable NLSpec-style harness specification package.

## Verdict

`block`

The S7 package is mostly reviewable and source-bounded, but acceptance should be
blocked until three issues are corrected or explicitly waived by an owner:

1. The expected final handoff note or final handoff section is missing.
2. The package does not explicitly list missing fixtures/goldens/snapshots or
   state that none were found.
3. A historical S7 implementation commit changed browser test logic without that
   change being represented in the S7 review packet or maintainer-decision
   summary.

This audit made no implementation, fixture, generated-artifact, CI, cleanup,
lockfile, package-manager, service, or runtime changes.

## Audit boundary

Scope inspected:

- S7 package outputs: `harness-nlspec.md`,
  `harness-acceptance-matrix.md`, `harness-implementation-roadmap.md`, and
  `harness-review-packet.md`.
- S0 through S6 recovery outputs, handoffs, and audits.
- S8 authority and preservation outputs.
- Cross-cutting registers: `ambiguity-register.md`, `source-limit-log.md`,
  `runtime-evidence-register.md`, `cleanup-signal-evidence-register.md`, and
  `maintainer-decision-summary-2026-05-09.md`.
- Main authority context through Core 00 through Core 04 routing documents and
  `docs/domain.md` vocabulary boundaries.

Non-scope honored:

- No harness rewrite.
- No implementation behavior change.
- No fixture, generated artifact, CI, cleanup, lockfile, or package-manager
  edits.
- No additional runtime evidence collection.

Inspection commands were limited to non-mutating file reads/searches and git
history inspection, including `rg`, `sed`, `nl`, `find`, `git status`,
`git diff --name-only`, `git log`, and `git show`.

## Reviewed S7 outputs

| Output | Audit result | Notes |
|---|---|---|
| Harness NLSpec draft | `pass` | Exists with S7 doc metadata, scope, non-goals, authority, command surface, local-dev, scheduler, environment/platform, retained artifacts, reset route, cleanup, fixture/snapshot, CI, and acceptance-hook sections. |
| Acceptance matrix | `pass_with_blocking_gap` | Exists and has `HAC-0001` through `HAC-0020` plus `HAC-GAP-*` rows, but the missing-fixture/golden/snapshot list is not explicit. |
| Missing fixture list | `block` | No explicit missing-fixture or no-missing-fixtures statement was found. `HAC-GAP-0002` covers visual snapshot refresh authority, but not fixture/golden absence or completeness. |
| Future CI gate list | `pass` | `HAC-GAP-0005` preserves provider CI annotations/uploads as future owner/evidence work, and the NLSpec keeps CI provider-neutral. |
| Harness roadmap | `pass` | Roadmap phases are written as implementation/future sequencing and preserve source-limited items. |
| Recovery review packet | `pass_with_blocking_gap` | Packet summarizes implemented decisions, deferred source limits, and verification results, but lacks the required final handoff section. |
| Final handoff note | `block` | No S7 handoff file exists under `handoffs/`, and `harness-review-packet.md` has no final handoff section despite `03-sprint-plan.md` requiring one. |

## Traceability summary

| S7 claim area | Traceability result |
|---|---|
| Make is canonical | Traces to `MD-S7-0001`, `AUTH-0002`, `AUTH-0003`, `PRES-0001`, and `PRES-0011`; acceptance row `HAC-0001`. |
| Generated artifacts are downstream inputs | Traces to `MD-S7-0003`, `AUTH-0010`, `PRES-0002`, `MSC-0005`; acceptance row `HAC-0002`. |
| Phase 7/8 are future work | Traces to `MD-S7-0002`, `AMB-0004`, `SL-0005`; acceptance row `HAC-0003`. |
| Reset route is harness-owned | Traces to `MD-S7-0004`, `AUTH-0004`, `PRES-0010`, `MSC-0001`; acceptance rows `HAC-0004` through `HAC-0006`. |
| Local-dev behavior is verification scope | Traces to `MD-S7-0005`, `AUTH-0005`, `PRES-0012`, `MSC-0003`; acceptance row `HAC-0007`. |
| Scheduler lanes are logical | Traces to `MD-S7-0006`, `AUTH-0011`, `PRES-0003`; acceptance row `HAC-0008`. |
| Cleanup is bounded and best effort | Traces to `MD-S7-0008`, `SL-0014`, `SL-FU-0002`, `SL-FU-0008`, and cleanup evidence; acceptance row `HAC-0010`. |
| Environment/platform limits | Traces to `MD-S7-0010`, `MD-S7-0011`, `SL-0013`, `SL-0015`, `AMB-0028`; acceptance rows `HAC-0012` and `HAC-0013`. |
| Retained artifacts need explicit identity | Traces to `MD-S7-0012`, `AUTH-0013`, `PRES-0017`, retained artifact rows; acceptance row `HAC-0014`. |
| Fixture/golden/snapshot workflow | Traces to `MD-S7-0013`, `AUTH-0014`, `PRES-0018`, artifact rows; acceptance rows `HAC-0015` and `HAC-0016`, with blocking gap for explicit missing-fixture/golden/snapshot list. |
| Unknown schemas remain unknown | Traces to `MD-S7-0014`, `SCHEMA-*`, `OI-*`; acceptance row `HAC-0017`. |
| CI remains provider-neutral | Traces to `MD-S7-0015`, `SL-0001`, `AUTH-0015`, `PRES-0019`, `MSC-0010`; acceptance row `HAC-0018`. |
| Extended smoke is demoted | Traces to `MD-S7-0016`, `FAIL-0028`, selected runtime evidence, and sequence-manifest changes; acceptance row `HAC-0019`. |
| Final `MUST` language is evidence-gated | Traces to `MD-S7-0017`; acceptance row `HAC-0020`. |

## Blocking findings

| Finding ID | Finding | Evidence | Required resolution |
|---|---|---|---|
| S7-BLOCK-001 | The final handoff note is missing. | `03-sprint-plan.md` says S7 must produce a final handoff and a `harness-review-packet.md` final handoff section. `harness-review-packet.md` ends with review questions and has no final handoff section. No S7 handoff file exists under `handoffs/`. | Add a final handoff section to the review packet or create a S7 handoff file. It should include status, changed paths, verification summary, source limits, open owner decisions, implementation-change audit, and exact next actions. |
| S7-BLOCK-002 | Missing-fixture/golden/snapshot output is not explicit. | `03-sprint-plan.md` requires identifying missing tests, missing fixtures, missing golden files, and future CI gates. `harness-acceptance-matrix.md` has `HAC-GAP-*` rows, including visual snapshot refresh authority, but no explicit missing-fixture/golden/snapshot list or statement that no missing fixture/golden files were found. | Add an explicit missing fixture/golden/snapshot list or a reviewed `none identified` statement with evidence source. Keep visual snapshot update authority source-limited unless owner bounds are supplied. |
| S7-BLOCK-003 | One historical S7 browser test-logic change is not accounted for in S7 handoff/review materials. | Commit `6f7ec1c` changed `apps/web/e2e/phase6.session-recovery.spec.ts` by raising the concurrency eviction login burst from `5` to `10`. The S7 review packet accounts for reset-route movement and stale-smoke demotion, but does not mention this test change; no matching `MD-S7-*` decision was found. | Owner must confirm the change was separately authorized or outside S7 scope, or the S7 review packet/final handoff must record the rationale, evidence, and authorization. If not authorized, record remediation before accepting S7. |

## Follow-up findings

These do not block S7 package acceptance if they remain clearly labeled as
future work, source-limited, or owner-required:

| Follow-up ID | Follow-up | Needed owner/evidence |
|---|---|---|
| S7-FU-001 | Environment-variable precedence remains unresolved. | Harness config owner plus Core 04 owner where product config overlaps; override matrix or owner decision. |
| S7-FU-002 | Visual snapshot refresh authority remains unresolved. | Browser/harness owner must name OS, browser, version, and update command before any snapshot update workflow is normative. |
| S7-FU-003 | Parent-death cleanup, active DB cleanup, and detached reaper completion remain source-limited. | Controlled cleanup evidence or owner decision. |
| S7-FU-004 | CI provider annotations/uploads remain source-limited while `.github/**` is absent. | CI owner decision or provider workflow source. |
| S7-FU-005 | Playwright report/trace/video/screenshot internals remain `schema_unknown` or tool-owned. | Tool-schema adoption decision or selected failure artifacts. |
| S7-FU-006 | Release readiness beyond stale-smoke demotion remains a claim boundary. | Passing release evidence or explicit non-readiness classification for any future failure. |

## Authority and evidence checks

- Observable behavior is generally specified over mechanism. Mechanism appears
  where it is contract-bearing, such as Make authority, generated artifact
  ownership, reset-route location, cleanup proof gates, and CI sequence behavior.
- The NLSpec separates harness validation behavior from product behavior. It
  states that Core 00 through Core 04 own product behavior and that reset routes
  are test hooks, not product APIs.
- Source-limited behavior is preserved for environment precedence, platform
  support, cleanup guarantees, provider CI, visual snapshot refresh, retained
  artifacts, and unknown tool schemas.
- Maintainer decisions are cited for Make authority, generated artifacts, reset
  ownership, local-dev scope, scheduler lanes, cleanup strength, external caches,
  environment/platform limits, retained artifacts, fixture/snapshot workflow,
  machine schemas, CI, stale-smoke demotion, and final evidence-gated `MUST`
  language.
- No source-limited runtime behavior was found promoted to an unqualified
  product or platform guarantee.

## Acceptance checks

The NLSpec contains lower-case normative `must` / `must not` language. The audit
maps those claims as follows:

| NLSpec claim area | Acceptance or owner routing |
|---|---|
| Evidence split must be preserved | `HAC-0020`, `MD-S7-0017`. |
| Provider CI behavior must not be invented | `HAC-0018`, `MD-S7-0015`, `SL-0001`. |
| Release readiness failures must be separate from stale-smoke demotion | `HAC-0019`, `HAC-GAP-0007`, `MD-S7-0016`. |
| Local-dev behavior must not become production guarantees | `HAC-0007`, `MD-S7-0005`, `MSC-0003`. |
| Env precedence must remain unresolved | `HAC-0012`, `MD-S7-0010`, `SL-0015`. |
| Reset route must stay harness-owned, disabled by default, and test-wired only | `HAC-0004` through `HAC-0006`, `MD-S7-0004`, `MSC-0001`. |
| Reset route must not be documented as product API | `MD-S7-0004`, `MSC-0001`; acceptable as owner-routed but should be kept visible in final handoff. |
| `.github` workflow behavior must not be invented | `HAC-0018`, `MD-S7-0015`, `SL-0001`. |
| Every final `must` / `must not` needs acceptance or owner routing | `HAC-0020`, `MD-S7-0017`. |

No vague pass predicates such as unqualified "works", "robust", "appropriate",
or "reasonable" were found in the S7 acceptance pass conditions.

## Roadmap checks

- Roadmap work is sequenced as future or implementation work and does not claim
  source-limited behavior as completed.
- Phase 4 preserves visual snapshot refresh bounds as unresolved and requires no
  snapshot baseline changes.
- Phase 5 clearly ties stale-smoke demotion to `MD-S7-0016`.
- Phase 7 records verification as finalization work and preserves failure
  classification as blocking, source-limited, or non-blocking roadmap work.
- Roadmap language does not claim provider CI behavior, full platform support,
  cleanup guarantees beyond selected evidence, or snapshot update authority.

## No-implementation-change confirmation

Current audit actions:

- Worktree was clean before writing this audit document.
- This audit did not run test, CI, browser, service-backed, cleanup, generator,
  formatter, baseline, package-manager, or runtime evidence commands.
- This audit adds only this audit document.

Historical S7 implementation changes:

- Reset-route movement and bootstrap extraction are represented in the review
  packet and authorized by `MD-S7-0004`.
- CI/release stale-smoke demotion and sequence-test updates are represented in
  the review packet and authorized by `MD-S7-0016`.
- Generated task/schedule artifacts are downstream inputs and are covered by
  `MD-S7-0003` and the generated-artifact policy.
- The browser test change in
  `apps/web/e2e/phase6.session-recovery.spec.ts` is not represented in the S7
  review packet or maintainer-decision summary and is blocking until accounted
  for.

## Validation criteria

| Criterion | Result | Notes |
|---|---|---|
| Required S7 deliverables are present and mutually consistent. | `fail` | Main files exist, but final handoff and explicit missing-fixture/golden/snapshot list are missing. |
| NLSpec content is reviewable, source-bounded, and authority-aware. | `pass` | Draft is reviewable and preserves source limits/authority boundaries. |
| Normative language is acceptance-mapped or owner-routed. | `pass_with_followup` | Claims are broadly mapped; keep reset-route product-API boundary explicit in final handoff. |
| Source limits and open decisions are consolidated. | `pass` | Review packet and acceptance matrix summarize the main open limits. |
| Roadmap and future gates are future work only. | `pass` | Roadmap avoids implying source-limited work is complete. |
| No unauthorized implementation or fixture/generated/runtime changes occurred. | `fail` | Audit made none, but historical S7 browser test logic change is unaccounted for. |

## Final audit checklist

- [x] S7 expected outputs exist and are reviewable.
- [x] NLSpec draft is complete enough for maintainer review.
- [x] Observable behavior is specified over mechanism unless contract-bearing.
- [x] Evidence labels are preserved and not conflated.
- [x] Harness spec does not redefine main-spec product behavior.
- [x] Every `MUST`/`MUST NOT` has binary acceptance or owner routing.
- [ ] Missing fixtures/goldens/snapshots and future CI gates are listed or explicitly routed.
- [x] Roadmap items are future work only.
- [x] Source limits and owner decisions are summarized.
- [ ] No unauthorized implementation, fixture, generated, CI, cleanup, or lockfile changes occurred.
- [x] Audit verdict and follow-up list are recorded.
