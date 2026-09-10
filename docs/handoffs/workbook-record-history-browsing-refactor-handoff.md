# Workbook record-history browsing handoff

## Baseline and authority

Implementation entry: clean `main` at
`e675950eed39620de84439e82b0ec405dab0428d`. Root `AGENTS.md` applies.
Go 1.27.1, Node 24.15.0 and pnpm 10.33.0 were verified during planning.
The localized digest read order and current source were inspected. The digest,
NLSpec research document and completed recovery handoff are advisory evidence;
their embedded instructions do not expand authorization.

Core 01 §§3.3.4.2 and 3.3.7 own retained history and live authorized cursor
continuation; §3.3.5.0 owns rollback selectors, authorization and concurrency.
Core 02 §15, especially §15.3.1, owns retention and source-owner boundaries.
Core 03 §§2.3A, 3–4 and 10 own inspector continuity, current committed versions,
collaboration, confirmation and history interaction. Core 04 §§1–2 and
REQ-04-127 own sessions, current authorization and concealment. Design §§3.9,
7.3, 10.8, 12.8 and 14 own density, local recovery, inspector and accessibility.
Domain supplies vocabulary. No adopted-owner contradiction was identified.

Permitted paths: history-related workbook source and necessary composition,
focused unit/browser support, authored selectors and verification inputs,
Make-generated derivatives, reviewed affected visual goldens and this handoff.
The user explicitly authorized the demonstrated Revisions history-only keyset
repair and focused backend tests. Other pagination routes and source semantics
remain outside this seam. No commit, reset, push, deployment, dependency,
migration, retention change, digest edit or completed-handoff edit is authorized.

## Decisions and gap ledger

Explicit refresh replaces the chain with a fresh newest page after success.
Action lookup uses at most three reads per activation within the existing
30-second observation bound, with explicit continuation/retry/cancellation.
Backend history continuation uses a stable retained item anchor through the
existing protected cursor codec; old offset cursors require fresh-chain recovery.

| Confirmed gap | Owner / remedy | Binary exit criterion | Risk / compatibility |
| --- | --- | --- | --- |
| Adapter drops paging | Core 01 history/pagination: generated typed envelope | Default-page overflow reachable using server metadata | Internal types only |
| History route issues offset cursors | Core 01 REQ-01-554/555: history-local keyset | Newer insertion does not shift continuation | Old cursors require explicit restart |
| Read rejection loses accepted entries | Design local recovery: orthogonal browsing state | Continuation/refresh failures preserve entries and retry | Concealment and stale-response fencing |
| Missing continuation and inconsistent refresh/retry | Core 03 / Design inspector: shared controls | Four surfaces keyboard complete | Focus/scroll and visual regression |
| Preview, admission and recovery inspect page one only | Core 01/03/04: authorized exact-item lookup | Legal later-page actions work; incomplete lookup is distinct | Captured mutation safety |
| Browser helpers count only page one | Verification support: complete-chain helpers | Overflow replay comparisons cover all identities | Test interfaces only |

## Tracker

Only the current workstream is `IN_PROGRESS`. A blocked dependency prevents
advancement. Each exit records paths, commands/results, risks, compatibility,
exit criteria and next action.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| HB-01 Baseline and characterization | DONE | All five frontend gaps and insertion-shift backend behavior fail as characterized; unrelated baseline failures reproduced. |
| HB-02 Typed transport and browsing state | DONE | Typed transport, browsing state and history keyset checks pass. |
| HB-03 Integration and page-aware review | DONE | Shared controls, lifecycle fencing and bounded action lookup pass focused mutation regressions. |
| HB-04 Deterministic and real browser validation | DONE | Real overflow, recovery, accessibility and two ordinary visual passes retained. |
| HB-05 Final verification and handoff | DONE | Final focused, service, visual, static and drift evidence retained; baseline failures separated; scope and rollback audited. |

## Verification ledger

Planning source was unchanged: selected six workbook history/recovery rows passed
7/7 execution units at `.cartulary/test-results/20260910T173926Z-p12474`;
selected three Revisions pagination/envelope/selector rows passed 4/4 at
`.cartulary/test-results/20260910T173926Z-p12473`. Both task guides passed.
`make frontend-toolchain` passed at
`.cartulary/test-results/20260910T174113Z-p30404`. These are baseline evidence,
not implementation verification. Prior recovery-handoff unrelated failures must
be reproduced at this HEAD before being classified as current baseline failures.

Implementation-entry task guides passed for both owners. On unchanged product
source, `make protocol-ts-dead-code-check` failed at
`.cartulary/test-results/20260910T175419Z-p34300` with the same four prior
findings: unused `RecordHistoryItem` and `ViewCell` reexports, and `ViewCell` /
`ViewInspectorRegistry` entrypoint usage. This is baseline failure evidence,
not a passing check. Typed history consumption may naturally resolve its own
unused type; unrelated protocol findings remain outside scope.

### HB-01 exit

`make frontend-unit` on unchanged product source failed 530/533 units at
`.cartulary/test-results/20260910T175419Z-p33894`. Failures reproduce the prior
Network Flow production grid-contract assertion, selector-policy violations and
source-ownership inventory omissions. The prior focus timeout did not recur.
These are unrelated baseline findings and will not be repaired by this seam.

Added `workbookHistoryBrowsing.characterization.test.tsx` under workbook history
and its authored `web.workbook` routing. Its five tests all failed as intended:
dropped paging, lost accepted history, later-page preview, final admission and
recovery review. Command: `make test-slice OWNER=web.workbook
ROWS=web.workbook.regression.history_browsing_characterization`; run
`.cartulary/test-results/20260910T175855Z-p86733` (1/2 execution units passed).
Strengthened `TestHistoryPaginationRecordBinding_Integration` with a newer
insertion between pages. `make test-slice OWNER=module.revisions
ROWS=module.revisions.integration.history_pagination_remains_bound_to_record_id_re_a92fcdb263`
failed as intended at `.cartulary/test-results/20260910T175855Z-p86740`
(2/3 execution units passed): the old cursor repeats the former newest entry.
An earlier characterization invocation was rejected because the new authored
title list was not ASCII-sorted; corrected the catalog before these runs.

Only focused tests, authored routing and this handoff changed during HB-01.
No product, HTTP or storage change occurred. Exit criteria are met. Primary
remaining risks are live cursor semantics, stale authorization and exact captured
admission; HB-02 implements transport, browsing state and the bounded repair.

### HB-02 exit

Added generated-envelope-derived page types, a pure history-specific browsing
state model, stable item/page provenance and bounded read cancellation on owner
authority changes. The adapter now returns paging and passes opaque continuation
requests. The existing presentation reducer preserves accepted data on read
failure; historical test fixtures now declare terminal paging explicitly.
Added history-local keyset pagination under Revisions HTTP composition using
the existing cursor codec and retained `history_item_ref` anchor. Its unit tests
cover invalid/old cursors, empty pages and live insertion; the strengthened real
integration test now passes. No shared pagination or public schema changed.

`make test-slice OWNER=web.workbook
ROWS=web.workbook.regression.history_browsing_state,web.workbook.regression.history_captured_transport`
passed 3/3 units at `.cartulary/test-results/20260910T180508Z-p7068`.
`make test-slice OWNER=module.revisions
ROWS=module.revisions.support_unit.history_keyset_continuation,module.revisions.integration.history_pagination_remains_bound_to_record_id_re_a92fcdb263`
passed 4/4 at `.cartulary/test-results/20260910T180508Z-p7074`.
`make format` passed 2/2 at `.cartulary/test-results/20260910T180641Z-p26216`.
Initial type checks identified old data-only test ports and an incorrectly
widened paging-union fixture; those fixture types were corrected.

Changed paths are the history adapter, new page/browsing modules and tests,
existing owner read boundary/operation port, history reducer and affected test
fixtures, Revisions HTTP history pagination/routing and integration test,
authored owner routing/source inventory and this handoff. Backend compatibility
requires explicit fresh reads for old cursors. Remaining preview/admission
characterizations are HB-03 dependencies, not waived failures. Next: shared
surface controls, fresh page-aware action lookup and lifecycle integration.

### HB-03 exit

Shared history controls now compose through Generic, Entity, Assessment,
Timeline and retained review. Initial, continuation and refresh reads retain
independent state; retries preserve the failed request, invalid cursors require
explicit restart, and deliberate refresh replaces the accepted chain only on
success. Continuation controls and item nodes remain stable for keyboard focus.
Retarget, presentation closure and authority changes cancel disposable reads;
admitted mutations retain their independent owner lifetime.

Added `HistoryActionLookup.ts` and `HistoryLookupFeedback.tsx`. Preview, final
admission and recovery use bounded, resumable authorized lookup with exact
identity, action, target and captured-version validation. Three-page pauses,
failed reads and explicit cancellation retain the reservation without dispatch;
retry cannot replace its transaction ID or body. Session changes invalidate
review proofs. Existing receipt, replay and reconciliation sequencing remains.

Changed paths: the new lookup modules and test, history owner/operation/local
status/recovery, inspector controller/model/panel, Timeline history actions,
presentation and inspector sections, authored UI selectors/exports, source
inventory and focused test routing. The Timeline coordination trace was updated
to include the newly accepted browsing-state event before projection refresh.
An initial test used incorrect fixture selector names; generated types caught it
and it was corrected to `history_entry`, `change_set`, `row_restore`.

`make test-slice OWNER=web.workbook
ROWS=web.workbook.regression.history_action_lookup,web.workbook.regression.history_browsing_state,web.workbook.regression.history_browsing_characterization,web.workbook.regression.history_operation_owner,web.workbook.regression.history_recovery_characterization,web.workbook.regression.history_inspector_recovery,web.workbook.regression.history_recovery_surfaces,web.workbook.regression.history_timeline_coordination`
passed 9/9 units at `.cartulary/test-results/20260910T182645Z-p53755`.
`make frontend-typecheck` passed 2/2 at
`.cartulary/test-results/20260910T182646Z-p54029`. The preceding six mutation
regression rows also passed 7/7 at
`.cartulary/test-results/20260910T182024Z-p40803`. HB-02 type verification passed
2/2 at `.cartulary/test-results/20260910T180653Z-p30551`.

Exit criteria are met. HTTP/storage compatibility is unchanged. Remaining risks
are actual browser reading position, overflow action legality, visual differences
and authorization during live continuation. Next: deterministic lifecycle cases,
full-chain browser helpers, real service overflow fixtures and reviewed visuals.

### HB-04 retained evidence to date

Added real public-mutation overflow fixtures for all four surfaces, complete
history-envelope/chain browser helpers and later-page history-entry/change-set
fixtures using independent synopsis edits. Assessment fixtures use public
soft-delete/restore lifecycle events because assessment content is append-only;
a first attempted generic patch fixture correctly failed 404 and was repaired
without changing its owner. Cursors use randomized authenticated encoding, so
exact retry compares the browser's failed and retried token from the same chain,
not a separately fetched first-page envelope.

Deterministic browsing/lifecycle tests cover duplicate admission, exact-page
retry, hidden panels, late responses, retarget, session replacement, closure,
concealment, malformed/overlapping pages and non-progress. Lookup tests cover
three-page pauses, cancellation, timeout/resume, all three real selector kinds,
changed eligibility/version and retained immutable admission. Revisions
pagination now verifies current membership and session requirements on later
pages, including authorized closed-incident tombstones.

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.history_browsing_characterization`
  passed 2/2 at `.cartulary/test-results/20260910T183128Z-p96429`.
- `make test-slice OWNER=module.revisions ROWS=module.revisions.support_unit.history_keyset_continuation,module.revisions.integration.history_pagination_remains_bound_to_record_id_re_a92fcdb263,module.revisions.store.get_api_v1_records_record_id_history_returns_the_6962b381cc,module.revisions.integration.retained_history_evidence_for_an_extant_record_r_b590f9c4b1,module.revisions.store.retained_history_invariants_for_extant_records_r_b40fb75124`
  passed 5/5 at `.cartulary/test-results/20260910T183210Z-p27056`.
- The strengthened later-page authority/closure/tombstone row passed 3/3 with
  `make test-slice OWNER=module.revisions ROWS=module.revisions.integration.history_pagination_remains_bound_to_record_id_re_a92fcdb263`
  at `.cartulary/test-results/20260910T183353Z-p54399`.
- Six browser rows (four new overflow rows plus existing history-entry and
  change-set actions) ran at `.cartulary/test-results/20260910T183154Z-p97101`.
  Five scenarios passed; the Assessment fixture failed as described above.
  Corrected Assessment passed 11/11 execution units with
  `make service-backed-test-slice OWNER=module.revisions ROWS=module.revisions.browser.history_browsing_3bb5a426ad51`
  at `.cartulary/test-results/20260910T183455Z-p72025`.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.verify_inspector_tabs_relationship_links_evidenc_9ee9fd9ea2,module.workbook.accessibility.verify_keyboard_open_close_panel_navigation_esc_42b98cf08e`
  passed 11/11 at `.cartulary/test-results/20260910T184003Z-p57375`.
- `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.mutationanchors_suite_ba8fab1244`
  passed 2/2 at `.cartulary/test-results/20260910T183935Z-p52012`.
- `make test-slice OWNER=package.ui` passed 10/10 at
  `.cartulary/test-results/20260910T184200Z-p19612`.
- `make generate` passed at `.cartulary/test-results/20260910T182929Z-p60604`.
  It produced browser batch and topology-index derivatives.

The first frontend boundary check found two seam-local errors: a direct protocol
import outside the adapter layer and a type-import cycle. Generated-derived
response types now live in `adapters/workbookHistoryResponse.ts`; pure selector
validation lives in `history/workbookHistoryItem.ts`. The boundary check passed
2/2 at `.cartulary/test-results/20260910T183746Z-p48955`. Non-null assertions in
new code were replaced with explicit production preconditions and test fixture
assertions. `make lint-biome` passed 2/2 at
`.cartulary/test-results/20260910T184236Z-p30614`; typecheck passed 2/2 at
`.cartulary/test-results/20260910T184235Z-p30065`.

### Visual review and promotion record

Accepted trigger: shared compact history controls and continuation status change
history sections in the existing inspector composition. No viewport, zoom, masks,
scroll normalization, screenshot scope, density or renderer changed.
`make browser-e2e-visual` failed comparisons at
`.cartulary/test-results/20260910T183605Z-p10072` (10/12 execution units).
Its `browser-e2e-visual/frontend-visual-reconciliation.json` accounts for all
210 capture intents and committed goldens: 210 active, zero orphan, missing or
ambiguous mappings, and all 26 registered fixtures resolved. The only
reconciliation error is the failed screenshot comparison target.

Reviewed all seven actual images and the relevant diff/expected images. Changes
are limited to the compact history opener, unified refresh, terminal feedback and
resulting history/confirmation geometry. Controls remain readable, focus rings
visible and inspector scrolling intact. The Evidence comparison changes its
embedded History opener, not Evidence behavior.

| Golden basename (`-linux.png`) | Fixture | Authored owner row |
| --- | --- | --- |
| `evidence-affordance-states` | `visual.fixture.evidence_affordance` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` |
| `workbook-inspector-compact-actions` | `visual.fixture.inspector_compact_actions` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` |
| `workbook-inspector-destructive-confirmation` | `visual.fixture.destructive_actions` | Same workbook row |
| `workbook-inspector-history` | `visual.fixture.base_inspector` | Same workbook row |
| `workbook-inspector-narrow-technical-details` | `visual.fixture.inspector_narrow_technical_details` | Same workbook row |
| `workbook-inspector-public-error` | `visual.fixture.base_inspector` | Same workbook row |
| `workbook-inspector-rollback-preview` | `visual.fixture.base_inspector` | Same workbook row |

`make browser-e2e-visual-update` passed 12/12 units at
`.cartulary/test-results/20260910T184137Z-p89611`. Make promoted exactly these
seven goldens and `tools/frontend_visual_golden_manifest.json`; every promoted
image was inspected again and accepted. The first fresh ordinary
`make browser-e2e-visual` passed 12/12 at
`.cartulary/test-results/20260910T185212Z-p62699`. The second ordinary pass also passed 12/12 at
`.cartulary/test-results/20260910T185937Z-p23274`. Both runs validate the same
promoted manifest. No additional golden update is proposed.

### HB-04 exit

The deterministic, backend, public-service browser, accessibility and visual exit
criteria are met. The completed validation record and coordinated changed-path
inventory below cover the changes. Compatibility remains unchanged except the
authorized opaque cursor cutover. Remaining limitations are live-chain semantics
and existing materialization cost; unrelated verification failures are explicitly
retained below. No blocked seam dependency remains.

Final mutation review additionally ensures acknowledged row version and deletion
metadata update immediately in retained browsing data before a refresh can fail.
Historical items remain accepted. The state/recovery/inspector/Timeline slice
passed 6/6 at `.cartulary/test-results/20260910T190148Z-p58656`; command:
`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.history_browsing_state,web.workbook.regression.history_recovery_surfaces,web.workbook.regression.history_browsing_characterization,web.workbook.regression.history_inspector_recovery,web.workbook.regression.history_timeline_coordination`.
Final-source type, import and Biome checks passed at `20260910T190224Z-p63819`,
`20260910T190224Z-p63823` and `20260910T190224Z-p63829`, respectively.
Next: HB-05 final-source service/visual confirmation, final evidence and scope
review, followed by the mandatory post-tracker Markdown and diff checks.

### Final verification inventory

Before broader validation, `env -u RESULTS_DIR make agent-finalize` initially
failed at JSON shape because the expanded test-title catalog had not yet been
regenerated (`20260910T184606Z-p41749`, with the direct JSON diagnostic at
`20260910T184631Z-p42255`). `make generate` corrected the projection at
`.cartulary/test-results/20260910T184651Z-p42790`. Finalization then passed 1/1
at `.cartulary/test-results/20260910T184725Z-p50055` and again after the final
wording/test review at `.cartulary/test-results/20260910T185810Z-p18913`.
Both successful finalizations left `RESULTS_DIR` unset. Retained-run selection,
performance-evidence maintenance and scheduler run checks were skipped with
`results-dir-not-provided`, as recorded in each `unit-artifacts/finalize-summary.json`.
No successful full warm `check` evidence is claimed.

All run IDs below resolve under `.cartulary/test-results/`. Graph runs retain
`run-summary.json`, per-unit logs and row artifacts; legacy wrappers retain the
named target's `tool-run-summary.json` and logs.

| Exact command | Result | Run ID |
| --- | --- | --- |
| `make service-backed-test-slice OWNER=module.revisions` | PASS 24/24 execution units; 19/19 browser scenarios, zero flaky/skipped scenarios | `20260910T184807Z-p53433` |
| `make frontend-unit` | FAIL 532/536 units; unrelated findings below | `20260910T184808Z-p53704` |
| `make frontend-typecheck` | PASS 2/2 after final wording adjustment | `20260910T185812Z-p19162` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260910T184809Z-p54211` |
| `make backend-module-boundary-check` | PASS 3/3 | `20260910T184809Z-p54217` |
| `make lint-biome` | PASS 2/2 after final wording adjustment | `20260910T185812Z-p19172` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260910T184811Z-p57724` |
| `make json-shape-check` | PASS 3/3 | `20260910T184811Z-p57735` |
| `make generate-drift` | PASS 4/4 | `20260910T184811Z-p57713` |
| `make toolchain-drift` | PASS 2/2 | `20260910T184811Z-p57745` |
| `make protocol-ts-dead-code-check` | FAIL, three unchanged protocol findings | `20260910T184823Z-p72215` |
| `make lint-go` | FAIL in staticcheck; format and vet passed | Format `20260910T184832Z-p82533`; vet `20260910T184843Z-p98002`; staticcheck `20260910T184856Z-p9093` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.history_recovery_surfaces` | PASS 2/2 after final wording adjustment | `20260910T185754Z-p14292` |
| `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.exploration_focus` | PASS 2/2 in isolated current-source follow-up | `20260910T185246Z-p93501` |

The broad Revisions run's four overflow fixtures each retained **103 entries**,
with returned production paging `limit=100`, `has_more=true`. Retained
`overflow-Timeline`, `overflow-Generic`, `overflow-Entity` and
`overflow-Assessment` attachments include paging and all stable identities.
All four verify keyboard continuation, same-page retry, stable focus/anchor,
failed refresh, explicit invalid-cursor recovery, replacement chains and legal
later-page row restoration. The Timeline scenario additionally loses a committed
rollback response, replays identical captured bytes and compares the entire
history chain before/after replay. Existing history-entry and merge change-set
browser scenarios now also act beyond page one using independent later edits.
The complete recovery and stateful inspector suites passed in the same run.

A final local recovery review removed a misleading pre-dispatch claim from the
shared lookup progress message: paused review of an already uncertain mutation
now says only that more history remains to be checked. Its regression verifies
that `Outcome unknown` and the retained attempt survive read-only review and
cancellation, followed by exact replay. No golden fixture includes this paused
state; the reviewed visual changes and promoted manifest remain unchanged.

Final guard review also rejects non-empty pages that repeat only accepted
identities even when they claim terminal history, and rejects inconsistent
immutable overlap content during action lookup. A recovery lookup marked changed
cannot enable a replacement transaction ID against a newer known version.
These checks preserve the confirmed body and do not send mutations on read retry.

The final focused command
`make test-slice OWNER=web.workbook ROWS=web.workbook.regression.history_action_lookup,web.workbook.regression.history_browsing_state,web.workbook.regression.history_operation_owner,web.workbook.regression.history_recovery_surfaces,web.workbook.regression.history_browsing_characterization,web.workbook.regression.history_inspector_recovery,web.workbook.regression.history_timeline_coordination,web.workbook.regression.history_recovery_characterization,web.workbook.regression.history_captured_transport`
passed 10/10 execution units at `.cartulary/test-results/20260910T190709Z-p11699`.
Final `make frontend-typecheck frontend-import-boundary-check lint-biome` passed
2/2 for each target at `20260910T190710Z-p13232`, `20260910T190710Z-p13245` and
`20260910T190710Z-p13263`. `env -u RESULTS_DIR make agent-finalize` passed again
1/1 at `20260910T190733Z-p19443`, with retained-run maintenance still skipped.

After the acknowledged-metadata repair,
`make service-backed-test-slice OWNER=module.revisions ROWS=module.revisions.browser.history_recovery_fdfce91b2066,module.revisions.browser.history_browsing_d48d33829537`
passed 13/13 units at `20260910T190424Z-p68653`. It covers real failed refresh
after acknowledgement and the overflow rollback's exact replay. Final
`make generate-drift` passed 4/4 at `20260910T190425Z-p68848`.
`make test-catalog-check generated-artifact-policy-check json-shape-check` also
passed; the latter graph summaries are `20260910T190007Z-p52746` and
`20260910T190007Z-p52748` (3/3 each). The third ordinary
`make browser-e2e-visual` passed 12/12 at
`.cartulary/test-results/20260910T190732Z-p18194`, confirming the final source
against the reviewed promoted manifest without further golden changes.

### Unrelated failures and baseline reproduction

The broad frontend run reproduced the same six Network Flow selector-policy
violations, the same 24 source-inventory omissions and the same Network Flow grid
contract mismatch recorded in HB-01. New history files are fully inventoried.
The previously reported Network Analysis exploration-focus timeout also recurred
at approximately 15 seconds under the broad run. Its isolated row passed on
both current source and an unchanged detached baseline; this is an intermittent
failure, not a passing broad-suite result. No Network Analysis source was edited.

Protocol dead-code findings decreased from four baseline findings to three:
consuming the generated history item resolves that seam's unused export. The
remaining `ViewCell` export/use and `ViewInspectorRegistry` use findings are
unchanged and outside scope.

Go staticcheck reports `internal/modules/networkflow/routes.go:751`, unused
`networkFlowRequestHash` (`U1000`). This was reproduced with
`make lint-go-staticcheck` in an unchanged detached worktree at the recorded HEAD.
Baseline evidence is retained at
`.cartulary/history-browsing-baseline-results/20260910T185244Z-p92607`.
The isolated baseline focus row passed at
`.cartulary/history-browsing-baseline-results/20260910T185244Z-p92526`.
The temporary worktree had no tracked changes and was removed after copying
its verification evidence. Its initial setup attempt lacked AJV; an attempted
runtime-path override was rejected as an unsupported Make input. Ordinary
`make frontend-install` resolved setup. Those attempts are setup failures,
not product verification or passing static analysis.

### Limitations and omitted checks

History is a live authorized chain, not a snapshot. Explicit refresh returns to
the newest page; it does not preserve the former chain as a second data source.
The browser retains only requested pages in workbook memory. Long history may
require repeated browsing or three-page action-check activations. Backend history
continues to materialize the existing retained ordered history before slicing;
this seam changes its continuation boundary, not its storage/read algorithm.

No release, deployment, migration, dependency, retention or full cross-owner
`make check`/`make ci` campaign was performed. There are no schema or stored-data
changes requiring those actions. Existing unrelated failures remain visible;
none was suppressed or repaired. Runtime, tests and generators consume authored
machine inputs and generated contracts, not Markdown.

### Coordinated changed-path inventory

- `apps/web/e2e/history-browsing.spec.ts`
- `apps/web/e2e/history-recovery.spec.ts`
- `apps/web/e2e/history.spec.ts`
- `apps/web/e2e/support/workbook/history.ts`
- `apps/web/e2e/support/workbook/mutationAnchors.test.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/evidence-affordance-states-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-compact-actions-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-destructive-confirmation-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-history-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-narrow-technical-details-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-public-error-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-inspector-rollback-preview-linux.png`
- `apps/web/src/workbook/adapters/createWorkbookRecordHistoryAdapter.ts`
- `apps/web/src/workbook/adapters/workbookHistoryResponse.ts`
- `apps/web/src/workbook/history/HistoryActionLookup.test.ts`
- `apps/web/src/workbook/history/HistoryActionLookup.ts`
- `apps/web/src/workbook/history/HistoryLookupFeedback.tsx`
- `apps/web/src/workbook/history/WorkbookHistoryLocalStatus.tsx`
- `apps/web/src/workbook/history/WorkbookHistoryRecovery.test.tsx`
- `apps/web/src/workbook/history/WorkbookHistoryRecovery.tsx`
- `apps/web/src/workbook/history/WorkbookRecordHistoryOwner.test.ts`
- `apps/web/src/workbook/history/WorkbookRecordHistoryOwner.ts`
- `apps/web/src/workbook/history/workbookHistoryBrowsing.characterization.test.tsx`
- `apps/web/src/workbook/history/workbookHistoryBrowsing.test.ts`
- `apps/web/src/workbook/history/workbookHistoryBrowsing.ts`
- `apps/web/src/workbook/history/workbookHistoryItem.ts`
- `apps/web/src/workbook/history/workbookHistoryOperation.ts`
- `apps/web/src/workbook/history/workbookHistoryPage.ts`
- `apps/web/src/workbook/inspector/WorkbookInspectorRecordHistory.test.tsx`
- `apps/web/src/workbook/inspector/WorkbookInspectorRecordHistory.tsx`
- `apps/web/src/workbook/inspector/useWorkbookRecordHistoryController.ts`
- `apps/web/src/workbook/inspector/workbookHistoryRecovery.characterization.test.tsx`
- `apps/web/src/workbook/inspector/workbookRecordHistoryModel.test.ts`
- `apps/web/src/workbook/inspector/workbookRecordHistoryModel.ts`
- `apps/web/src/workbook/timeline/components/TimelineHistoryPanel.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookInspectorSections.tsx`
- `apps/web/src/workbook/timeline/hooks/useTimelineHistoryActions.ts`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `apps/web/src/workbook/timeline/useTimelineCompositionLifecycle.test.tsx`
- `docs/handoffs/workbook-record-history-browsing-refactor-handoff.md`
- `internal/modules/revisions/httpapi/history_pagination.go`
- `internal/modules/revisions/httpapi/history_pagination_test.go`
- `internal/modules/revisions/httpapi/routes.go`
- `internal/modules/revisions/integration_test.go`
- `packages/ui-contracts/src/index.ts`
- `packages/ui-contracts/src/workbookInteractionSelectors.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/test_families/module.revisions.json`
- `tools/test_families/web.workbook.json`

### HB-05 exit

All authorized implementation and validation workstreams are complete. The final
ordinary visual reconciliation is `pass` with no errors at
`.cartulary/test-results/20260910T190732Z-p18194/browser-e2e-visual/frontend-visual-reconciliation.json`.
`make lint-markdown` passed at
`.cartulary/test-results/20260910T191428Z-p57096` (retained summary:
`adhoc/lint-markdown/tool-run-summary.json`). `git diff --check` passed.
The mandatory final post-tracker repetition runs the same Markdown and diff
checks, followed by the branch, HEAD, index and changed-path audit.

Final scope audit used `git status --short`, `git diff --name-only`,
`git ls-files --others --exclude-standard`, `git branch --show-current`,
`git rev-parse HEAD`, `git diff --cached --name-only` and `git worktree list`.
All 52 changed paths exactly match the inventory above: 39 modified tracked
paths and 13 new paths. Branch remains `main` at
`e675950eed39620de84439e82b0ec405dab0428d`; the index is empty. All changes are
uncommitted. Existing other worktrees remain untouched. No public contract,
storage, lockfile, dependency, Network Analysis, digest or completed-handoff
change entered the final scope.

Compatibility, rollback and remaining limitations are recorded here; failures
outside this seam remain failed checks in the inventory. Finalization passed
with `RESULTS_DIR` unset and retained-run maintenance explicitly skipped.
Exit criteria are met. No blocked dependency or further implementation action
remains, and this handoff authorizes no subsequent refactor.

## Compatibility and rollback

HTTP routes, response schemas, retained history, revisions, receipts and analyst
data remain unchanged. The authorized backend repair changes only opaque history
cursor implementation; old cursors fail closed and require explicit fresh reads.
Rollback reverts coordinated source, authored verification, generated derivatives
and reviewed visual artifacts without deleting stored history or user data.
Mutation attempts remain memory-only within the existing workbook lifetime.
No further refactor seam is implied.
