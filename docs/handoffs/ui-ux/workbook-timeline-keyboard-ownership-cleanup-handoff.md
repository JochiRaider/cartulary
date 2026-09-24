# Timeline editor keyboard ownership cleanup

## Scope and authority

This is an interaction correction on `main` from a clean `812d77e0d` checkout.
Core 03 §13.1 REQ-03-218/300 governs accepted editor departure, rejection,
native controls, composition and focus. Core 03 §7 REQ-03-111/099/100/298
governs trailing capture and retained authoring. Design §8.6 supplies the
keyboard presentation. Inspector Details keeps its separate Ctrl/Cmd+Enter
submission contract. The September 13 UI/UX digest was rechecked against
current source and used only as advisory navigation.

`packages/grid-adapter` owns neutral grid event admission and the RDG binding;
`web.workbook` owns Timeline intent, scalar authoring and generic Workbook
controls. `module.timeline`, `module.workbook`, `web.workbook` and
`package.grid_adapter` independently route verification. Current source/import
ownership and all four owner task guides were consulted. No controlling active
tracker covers this bounded slice. No API, schema, storage, authorization,
query, mutation-settlement or migration change was made.

## Reproduction and decision

Before correction, the production Adapter capture handler consumed Ctrl+Enter
as ordinary Enter. The new Adapter regression failed because it called commit
once; the production-wired Timeline controller regression failed because it
queued scalar save. The browser regression saw an extra write after its first
modified key. The browser probe attached to the final run shows Chromium under
Playwright delivered Control, Meta and Alt combined with Enter and Tab to the
page, all six with the probe input still focused. This establishes page-event
behavior in this automation environment; it does not claim that real browser or
OS tab-switching shortcuts will reach a page.

The common decision is whether an Enter/Tab chord is an ordinary grid
departure: Alt, Ctrl and Meta reject it; Shift only reverses admitted movement.
The existing Adapter interaction boundary now makes that decision for the
production capture handler and test-support binding. A rejected page event
stops before child and RDG key handlers, without preventing the browser's own
default. Composition, nested interactions, native select, multiline
Shift+Enter, Alt+ArrowDown and Escape retain their existing precedence.
Timeline's controller and intent model and Generic/Entity Workbook child editor
use the same admission decision. TimelineScalarEditor no longer has a competing
managed Enter/Tab departure branch. The controller's production draft-navigation
callback is required and present in its test fixture.

The boundary is small and reusable by another writable Adapter consumer without
a new keyboard framework. Leaving the old checks would permit child or vendor
commit after a parent rejection. Generic and Entity use the shared Workbook
adapter; their child Enter action now honors the same decision and prior event
prevention. Network Analysis uses SemanticDataGrid without an editor and has no
changed departure path. No compatibility alias or data migration is needed.
Rollback is the code and authored-routing diff in this slice; generated
derivatives must be regenerated from their owners.

## Verification

Regression routing was authored in `tools/test_families` before `make generate`
updated the tool-managed browser batch and execution topology. Tests and runtime
do not read Markdown or the digest.

| Check | Result and artifact |
| --- | --- |
| Red Adapter, controller and browser baselines | Product failures: `.cartulary/test-results/20260924T204210Z-p62917`, `20260924T204244Z-p63693`, `20260924T204818Z-p3637`. An earlier browser fixture timed out before reaching the target because the virtual cell needed scrolling; corrected in the test. |
| Final focused `make test-slice` | `package.grid_adapter` PASS `.cartulary/test-results/20260924T210415Z-p70668`; `web.workbook` PASS `20260924T210415Z-p70694`; `module.timeline` scalar component PASS `20260924T210433Z-p72133`. |
| Final `make service-backed-test-slice` | Timeline production editor/capture PASS `.cartulary/test-results/20260924T210443Z-p72753`; separate Timeline scalar clipboard characterization, including Inspector Ctrl+Enter, PASS `20260924T210717Z-p41162`; Workbook autosave and live shortcuts PASS `20260924T210549Z-p6126`. |
| Browser assertions | Committed editor and trailing capture retained exact raw drafts and focus without refocusing between key assertions. The committed semantic anchor and capture cell selection remained stable. Capture's input-triggered POST was counted before keys and held; rejected gestures produced no additional POST/PATCH before or after release. Existing ordinary/reverse traversal, acceptance/rejection, multiline, native select, composition, Alt+ArrowDown and Escape cases remained green in selected owner rows. |
| Terminal checks | `make generate`, `make test-catalog-check`, `make generate-drift`, `make generated-artifact-policy-check`, `make frontend-typecheck`, `make frontend-import-boundary-check` and `make lint-biome` PASS. `make format` applied repository formatting. Generation drift artifact: `.cartulary/test-results/20260924T210330Z-p65740`; type, import and lint artifacts: `20260924T210135Z-p56978`, `20260924T210221Z-p58979`, `20260924T210300Z-p65068`. |

An initial `make agent-finalize` failed on stale generated topology after a new
test title. Regeneration fixed it; the rerun passed at
`.cartulary/test-results/20260924T205902Z-p50926` before broader checks; the
final rerun passed at `.cartulary/test-results/20260924T211003Z-p74776`.
Broader type checking then exposed two typing mistakes in the new helper and
controller; both were corrected and type checking passed. Two initial lint and
import attempts lacked the bundled Node path; the reruns used the repository's
local Node runtime and passed. An attempted component row ID was wrong and was
rerun with the authored ID. These were verification setup issues, not remaining
product failures. `make lint-markdown` passed at
`.cartulary/test-results/20260924T211026Z-p78658`. `RESULTS_DIR` was unset, so
retained-run maintenance was skipped. No visual fixture or measurement target
applies to this key-admission correction.

## Digest acceptance disposition

| Rows | Disposition and evidence |
| --- | --- |
| A001–A003 | PASS. Exact Core/design clauses, source/verification boundaries, clean baseline, current wiring, narrow common decision, retirement, compatibility and rollback recorded above. |
| A007, A011, A014, A019 | PASS for this slice. Browser and owner regressions cover capture without a duplicate create, exact raw draft, semantic focus/selection, keyboard admission and recovery. |
| A023–A027 | PASS. Stable semantic selectors, Markdown-independent tests, authored routing followed by generation/drift, unchanged external contracts, and this handoff are evidenced above. |
| A004–A006, A008–A010, A015–A018, A020–A022 | N/A. Token/theme/density, shell/inspector layout and dispatch, conflict/query/authorization/evidence state, component geometry, virtualization and visual fixtures belong to unchanged owners and were not modified by this key-admission slice. A021 is N/A to virtualization claims; the focused semantic focus assertion is recorded under A011. |
| A012–A013 | N/A. Mutation identity, replay, queue recovery and acknowledgement owners were not changed; the selected write-count and acceptance tests check that rejected keys do not enter those paths. |

The acceptance result covers the observed page events and selected owner routes;
real browser/OS interception varies by platform and was not automated here.
