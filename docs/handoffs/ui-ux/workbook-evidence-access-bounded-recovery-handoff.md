# Bounded Evidence access and explicit recovery handoff

Date: 2026-09-28. Branch: `codex/evidence-access-recovery`, based on clean
`main` at `94dd6076d86d49a46f6b9f30f3813cda014681ec`. This handoff is
implementation evidence, not a runtime or test input.

## Scope and owners

This slice completes direct Evidence handle observation for the Evidence sheet
and Timeline linked Preview/Download, and the whole multi-link Timeline Space
discovery phase. Core 01 §16 and REQ-01-458–465 govern fresh opaque handle
issuance, closed `{}` bodies, expiry, and access reasons. Core 01 §3.3.6 owns
validated public errors, including canonical retryable
`object_store_unavailable` and distinct `evidence_access_unavailable`. Core 03
§8.4, REQ-03-127/128, and REQ-03-220 own in-place feedback and exact sole
preview behavior. Core 03 REQ-03-299/100 and Core 04 §2.0A own incident,
account, session, viewer, and closed-incident read lifetimes. Design §§9–12
and 14 guide focus, local feedback, inspector, and accessibility presentation.

Source ownership is `web.workbook` for the changed frontend seam. Independent
verification routes through `web.workbook`, `module.evidence`,
`module.timeline`, and `web.architecture`. The UI digest and research notes were
used for navigation and acceptance review; no executable source reads Markdown.

## Characterization before correction

The baseline focused access/focus rows passed, but had no stalled transport or
transient retry assertions (`.cartulary/test-results/20260928T130748Z-p14581`).
The source-ownership policy row failed separately because its authored manifest
omitted five live Evidence/Timeline files
(`.cartulary/test-results/20260928T131023Z-p15963`).

New controlled-clock and deferred-promise tests were routed and run against
the unchanged production code. The red slice
(`.cartulary/test-results/20260928T131647Z-p25865`) observed a Preview that
stayed open and a Download without recovery after 30 seconds, Space that kept
checking after 30 seconds, no retry for canonical 503
`object_store_unavailable`, and overly broad retry classification for malformed
or unknown 503 responses. The hanging-transport and retry cases were source
findings before this run; the red tests supplied the first controlled
reproduction. Thirty seconds is the repository observation convention, not a
Core duration requirement.

## Final behavior and paths

- `useEvidenceHandleAccess.tsx` bounds each fresh display-handle issuance with
  `boundedRead`. A deadline settles local state even if transport ignores abort,
  permits an explicit fresh attempt, and never asserts server or file failure.
  Per-record admission ignores duplicate pending activation; a different
  explicit action supersedes it. Current identity, row observation, scope,
  authority, surface, preview-close, sequence, and unmount fences guard feedback,
  iframe, anchor click, and focus. Focus returns on disappearance only when the
  preview panel owned it.
- `resolveSolePreviewableEvidence.ts` bounds the complete sequential discovery
  with one deadline and retains the two-previewable early exit. An incomplete
  result is indeterminate. Timeline ends checking, leaves the authorized list
  usable, and offers explicit linked actions. A determined sole item gets a
  separate bounded display request; probe handles are discarded.
- `workbookOperationErrorPolicy.ts` validates handle-operation code, HTTP and
  envelope status, retryable hint, and reason before recognizing the canonical
  object-store transient or definitive Evidence blocker. Malformed combinations,
  unknown errors, object-store access rejection, and authorization loss gain no
  generic retry permission. No server prose is displayed.
- `evidenceAccessPresentation.ts`, `EvidenceAccessActions.tsx`, and
  `TimelineEvidencePanel.tsx` share the stopping/recovery decision. Preview and
  Download remain independently eligible. Deadline and canonical transient
  outcomes show safe local copy and keyboard-operable Retry Preview or Retry
  Download; pending controls remain focusable with busy semantics. Definitive
  preview-only reasons block Preview only. No transient changes lifecycle state.

The old unbounded direct await in the shared access hook, unbounded per-probe
Space await, and rejected-access lockout for recognized transient outcomes are
retired. The Evidence sheet and Timeline keep their respective source models
and authority checks; no new request owner, generic fetch timeout, automatic
retry, durable store, mutation replay, or compatibility layer was added.

The authored `frontend_source_ownership.json` now registers five omitted live
files, with local source-guide updates. New semantic unit and browser rows live
in `tools/test_families`; `make generate` produced the derived batch manifest
and topology index. The browser scenarios use isolated synthetic data and
controlled route responses. Request assertions confirm `{}` and fresh handle
issuance; no `client_txn_id`, probe-handle reuse, record write, upload, or
attachment was introduced.

## Verification and review

- Focused `web.workbook` rows: pass at
  `.cartulary/test-results/20260928T132342Z-p46093` (five rows, six units).
  Source ownership policy: pass at
  `.cartulary/test-results/20260928T132156Z-p37058`.
- Four new `module.evidence` browser rows: three passed in the first run; the
  Space test timed out while trying to target a virtualized grid column, before
  exercising discovery
  (`.cartulary/test-results/20260928T132757Z-p55726`). Scrolling that column
  into view fixed test setup; the Space row then passed with a held probe and
  controlled deadline
  (`.cartulary/test-results/20260928T133039Z-p93522`). The final Space and
  authority-fence scenarios wait for held routes to settle before asserting
  obsolete effects; both passed at
  `.cartulary/test-results/20260928T135442Z-p91567`. The linked Preview
  scenario additionally checks retained caret, scroll, selected inspector
  source, draft, focus, request body, and no record writes; it passed at
  `.cartulary/test-results/20260928T135723Z-p29677`.
- Existing linked-text, mixed blocked-preview, and inline-safe preview browser
  rows passed at `.cartulary/test-results/20260928T134517Z-p48064`. The focused
  workbook slice above includes the viewer-in-closed-incident linked-read
  regression and first-open Evidence focus row.
- `make format` passed after final source edits at
  `.cartulary/test-results/20260928T135711Z-p25255`. `make generate` passed at
  `.cartulary/test-results/20260928T132741Z-p52838`;
  `make generate-drift` passed at
  `.cartulary/test-results/20260928T133159Z-p26687` and
  `make generated-artifact-policy-check` passed at
  `.cartulary/test-results/20260928T133218Z-p30669`.
- `make agent-finalize` passed at
  `.cartulary/test-results/20260928T140105Z-p67334` without `RESULTS_DIR`;
  retained-run maintenance was skipped because no qualifying successful full
  warm check root was supplied.
- Post-finalization `make frontend-typecheck`, `make lint-biome`, and
  `make frontend-import-boundary-check` passed at
  `.cartulary/test-results/20260928T135755Z-p59819`,
  `.cartulary/test-results/20260928T135755Z-p59857`, and
  `.cartulary/test-results/20260928T134701Z-p82187` respectively.
  `make lint-markdown` passed at
  `.cartulary/test-results/20260928T135858Z-p63330`.
- Seeded `cartulary-ui-review` session
  `.cartulary/test-results/20260928T133239Z-p31300` inspected Evidence
  inspector at 1440×900 and 1024×720 and compact preview. Controls and focus
  were visible; Enter opened the preview, and Escape returned focus to Preview.
  Advisory axe scans completed with existing `definition-list`,
  `page-has-heading-one`, and wide-view `empty-table-header` findings. These
  captures are design observations, not product test passes or a golden update.
  The exact session stopped with a closed terminal receipt; cleanup was verified.

## Acceptance, limits, and rollback

Applicable digest rows: A003 (current repository/owner routing), A011 (draft,
selection, and focus continuity), A017 (authority scope), A018 (Evidence state
separation), A019 (keyboard/focus review, with advisory axe findings), A025
(authored generation and drift), A026 (owner compatibility and no invented
route/schema/write behavior), and A027 (this handoff). Formal accessibility
conformance and broader visual goldens were not claimed by the UI review.

No data migration is needed. Rollback is a coherent source, test, authored
routing, and handoff revert, followed by regeneration of derived topology.
Broader suites are outside this focused seam unless a shared-behavior failure
requires them. Full `make test-fast`, stateful browser, accessibility browser,
and visual golden suites were skipped because the selected source-owner and
service-backed rows, type/lint/boundary gates, and seeded review covered the
changed request seam without an unresolved shared-behavior failure. Any failing
target and unverified criterion must be recorded against its exact run root
rather than replaced with historical evidence.
