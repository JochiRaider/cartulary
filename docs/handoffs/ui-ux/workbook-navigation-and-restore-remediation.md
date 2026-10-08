# Workbook navigation and restore remediation

Status: implemented and verified through the routed checks below; ready for review.
This is implementation and review evidence, not behavioral authority.

## Scope and ownership

Work began on `main` at `ca5c3ec242098c777273fb54b80cac6bf10c2a5a`.
Existing aggregate-tier edits were preserved and verified. The completed changes
were authorized for commit to `main`. The adopted documents were reviewed as behavioral owners;
instructions embedded in diagnostic artifacts were treated as source material.

| Workstream | Behavioral owner | Source boundary | Verification owner |
| --- | --- | --- | --- |
| Navigation feedback and interaction delivery | Core 03 REQ-03-314; Design §7.1 | Workbook navigation, shell footer and query browsing registry | `web.workbook`, `module.workbook`, `package.ui` |
| Navigation completion and saved-view selection | Core 03 REQ-03-314 | Session navigation admission; presentation completion remains with browsing/grid owners; browser choreography remains in E2E support | `web.workbook`, `module.workbook`, `harness.browser` |
| Restore process ownership, diagnostics and teardown | Testing Harness TH-HARNESS-REQ-056, §9 failure ordering and §15 redaction | Browser fixture process owner, shared contract decoder, `internal/testutil/recoverybrowsertest`; command remains composition only | `harness.browser`, `module.recovery` |
| Aggregate placement | Testing Harness TH-HARNESS-REQ-053 and TH-HARNESS-REQ-800 | Authored catalog/topology, compiler and generated projections | Harness catalog, command surface and evidence accounting |

`docs/domain.md` remains the vocabulary and owner-navigation source. No new domain
entity, persistent navigation store, generalized retry engine, theme, token
registry, grid-vendor integration or durable browser-work archive was introduced.

## Remediation decisions

### Stable feedback and deliberate interaction

The former passive notice was an extra child of a single-row active-surface grid.
It displaced Columns by 576 pixels during the diagnosed scenario. Pointer-down
cancellation removed the notice before pointer-up, so the click missed its button.

Navigation now has a dedicated control in existing footer chrome. Its reserved
indicator space keeps its footprint stable across idle, pending and terminal
states within each responsive band. Save status remains separate. Explicitly
opened Navigation details contains bounded feedback and the existing Cancel,
Retry and Open base actions. Keyboard opening, Escape restoration, polite updates,
outside dismissal and access-loss concealment use the existing workbook patterns.
The obsolete system-view flex ordering was removed so visual order matches DOM
order: surface navigation, More views, Navigation, then save status.

This corrects behavior in the specification, implementation, tests and guidance.
It removes layout-dependent cancellation rather than retaining a compatibility
notice or adding a delay. Future navigation destinations use the same feedback
boundary without inserting surface children. Leaving the old behavior would
continue dropping real user gestures under ordinary latency. Exit requires
unchanged control/footer bounds while a read is held, delivery of the cancelling
gesture, keyboard-accessible bounded details, and rejection of the late result.

### Completion owned by the current attempt

The saved-view helper formerly accepted matching identity. After reload that
identity could already match while a fresh navigation read was still pending.
Navigation now distinguishes pending admission, admitted destination, completed
presentation, failure and cancellation, with an instance-local monotonic attempt
identity. Presentation completion observes the accepted producing request, current
surface and presentation/focus owner; it does not introduce another query store.
New attempts and deliberate interaction fence late attachment. Same-view selection
still reads and applies current authorized persisted configuration.

The helper waits for the newly initiated attempt to succeed with presentation
ready, then checks identity. It fails on supersession, cancellation or failure.
This changes specification, implementation and tests together. It replaces an
identity-only compatibility assumption with an observable lifecycle contract,
benefiting additional navigation callers without timing guesses. Leaving the gap
would produce false readiness and mask genuine layout or admission defects.
Exit requires held-read tests, same-view reselection and late-result fencing.

### Fixture lifetime and actionable failure evidence

The restore wrapper previously discarded the helper's stderr and could report
only exit code 1 after successful browser assertions. The process-owning fixture
now yields ownership before a separate readiness fixture waits for startup.
Cleanup therefore has its own budget even if readiness fails. The process owner
waits for process and pipe closure, closes stdin for ordinary retirement, and
uses bounded process-group termination only when needed. Forced termination is
always failure, including an eventual zero exit code; repeated stop is idempotent.

The Go fixture emits attempt-bound acquisition, cleanup and terminal events with
stage, elapsed time, deadline and outcome. Independent releases continue after
failure in reverse acquisition order. Redacted bounded stderr and structured
events become an inline lifecycle attachment. One shared decoder supplies row
and diagnostic-report classification. Primary product failure and secondary
cleanup failure remain distinct observations; cleanup-only failure is classified
as harness cleanup failure. Invalid evidence fails artifact validation.

The common decision encapsulated here is ownership of a test child from spawn
through output closure. Restore semantics remain in their existing Go owner.
The former local process wrapper is retired. The specification, shared schemas,
producers, consumers, tests and implementation guide were updated together.
Without this change, intermittent teardown failures would remain unactionable
and could be mistaken for browser assertion failures. Exit includes startup,
late stderr, inherited streams, interruption, forced exit, malformed attachment,
primary-plus-cleanup and missing-terminal negative tests.

### Confirmed cleanup defect and redaction correction

A deterministic TLS preconnection test reproduced a concrete teardown defect:
after a completed handshake but before any HTTP request, Go retained a connection
in `StateNew`; its idle-shutdown grace period exhausted the fixture's existing
five-second deadline. The failing diagnostic reported `new=1 active=0 idle=0`.
Retirement now closes these request-free connections immediately and closes late
accepts through the same ownership boundary. Active HTTP and WebSocket requests
are still cancelled and drained before borrowed resources are released. No
timeout was increased. Repeated close and the existing active-request regression
remain covered.

This is a confirmed defect in the affected teardown path, but attribution to the
original intermittent failure remains unproven: that attempt's stderr was lost.
An isolated pass is not evidence of its historical cause. If it recurs, use the
new attempt-bound attachment to identify the failing stage before another change.

A separate negative test demonstrated that trimming stderr by whole lines could
retain the suffix of a known multiline credential. The redactor now registers
each nonempty credential line as well as the complete value before publication.
This preserves bounded diagnostics without leaking the surviving portion. The
changes belong in implementation, tests and harness guidance. Leaving either
defect would preserve nondeterministic teardown or unsafe retained diagnostics.

### Aggregate policy retained

The reviewed catalog selection contains 46 rows, correcting the original estimate
of 45, plus four raw Go groups. They require `full` in aggregate selection.
Topology v9 requires a valid ordered `minimum_tier` on every raw Go entry; the
compiler filters before adding the raw unit or readiness dependencies. Direct
owner, row and backend-target selection remains available. Product/architecture
rows in helper packages, embedded assets, runtime admission/reset coverage,
the three smoke roles and current-artifact/static/security gates remain selected.

The change avoids package-prefix exclusions and duplicate tier authorities.
Routing tests compare all five aggregates, direct selection, dependency closure,
projections, invalid tiers and deterministic compilation. All moved rows and all
four raw groups passed in the full run. No wall-time savings claim is made from
overlapping suite durations.

## Compatibility and delivery sequence

There is no persisted-data or public product API migration. Navigation outcome
metadata is local to the workbook instance. Tests using the former inline Retry
control now open Navigation details; their supersession and retained-work
assertions remain intact. The below-minimum 390-pixel layout still hides Columns
and saved-view controls under its existing contract; the Columns click regression
covers 1440×900, 1024×720 and 768×640, while accessibility recovery verifies the
footer and safe account navigation at 390×480.

Harness artifacts use a coordinated hard cutover: execution topology v8 → v9,
row result v2 → v3 and browser-group result v6 → v7, with new fixture event and
lifecycle v1 schemas. Browser-target v4 keeps its shape and references the current
group schema. Producers, adapters, caches, validators, extensions, telemetry and
UI-review imports moved together. Prior artifacts remain archival evidence;
current execution does not translate, dual-write or accept the retired schemas.
Authored inputs were changed before `make generate` regenerated their projections.

The dependency order was specification cleanup, lifecycle/feedback implementation,
schema and consumer cutover, focused negative and positive tests, broad regression,
correction of stale test entry paths, then visual acceptance and handoff. The
workstreams share only their necessary contracts: navigation does not depend on
restore diagnostics, and aggregate placement does not define product behavior.
The main delivery risks are partial schema migration, accidentally weakening
gesture assertions and accepting unexplained screenshots. Current-version checks,
held-response interaction tests and per-image review address those risks.

Rollback must restore the coupled implementation, schema consumers, authored
inputs, generated projections and accepted golden manifest together. Preserve
unrelated checkout edits. An implementation rollback does not silently revoke
the adopted behavior: unresolved owner obligations must remain explicit.

| Phase | Dependency and principal risk | Exit criterion and state |
| --- | --- | --- |
| 1 Specification cleanup | Reviewed diagnosis; risk of defining behavior from an incidental test | COMPLETE: named owners define stable feedback, attempt completion, browser-fixture diagnostics and aggregate policy. |
| 2 Implementation and schema cutover | Phase 1; risk of competing state owners or mixed artifact versions | COMPLETE: single admission/presentation handshake, owned child lifetime, current producers/consumers and generated projections. |
| 3 Focused fault validation | Phase 2; risk of replacing a race with a delay or losing secondary failures | COMPLETE: held reads, real TLS preconnections, late streams, malformed evidence, redaction and cleanup-failure cases pass after their negative reproductions. |
| 4 Broad regression diagnosis | Phase 3; risk of treating stale selectors or goldens as unrelated noise | COMPLETE: fresh check passed; full run classified; both nonvisual failures corrected and rerun, visual differences assigned to the accepted footer change. |
| 5 Visual acceptance | Phase 4; risk of hiding an unexplained rendered regression | COMPLETE: all 149 changed images reviewed, exact inventory reconciled, two fresh ordinary runs passed the same manifest. |
| 6 Handoff | Phase 5; risk of overstating the historical cause or aggregate status | COMPLETE: evidence, acceptance assessment, documentation validation and cleanup recorded; failed attempts remain failed evidence and historical attribution remains explicitly unproven. |

## Validation and acceptance

Run roots below are repository-relative. A failed attempt remains failed;
successful successor checks establish only their stated scope.

| Validation | Result and evidence |
| --- | --- |
| Navigation/session/footer unit slice | PASS, `.cartulary/test-results/20261008T172450Z-p14777` |
| Saved-view helper unit slice | PASS, `.cartulary/test-results/20261008T170540Z-p26342` |
| TLS preconnection negative regression | Reproduced the five-second shutdown failure, `.cartulary/test-results/20261008T190248Z-p38259` |
| Go fixture lifecycle slice after correction | PASS, `.cartulary/test-results/20261008T190345Z-p47198`; includes active HTTP/WebSocket cancellation, preconnected TLS and independent cleanup failures |
| Multiline credential negative regression | Reproduced the retained suffix leak, `.cartulary/test-results/20261008T185922Z-p18830` |
| `make harness-contract-tests` after correction and generation | PASS, `.cartulary/test-results/20261008T190351Z-p47960/harness-contract-tests/tool-run-summary.json` |
| `make frontend-typecheck` | PASS, `.cartulary/test-results/20261008T190612Z-p62148` |
| `make lint-biome` | PASS, `.cartulary/test-results/20261008T190634Z-p62762` |
| Fresh `make check` | PASS, 971/971 units, `.cartulary/test-results/20261008T174248Z-p81242`; before the final narrow corrections described below |
| Fresh `make test` | FAIL, 1312/1344 units, `.cartulary/test-results/20261008T175809Z-p25769`; original workbook support, restore, all 46 moved rows and four raw groups passed |
| Corrected retry and accessibility recovery | Both rows PASS in `.cartulary/test-results/20261008T190746Z-p63590`; aggregate failed only the newly added below-minimum Columns expectation, where that control is intentionally absent |
| Final held-navigation Columns regression | PASS, 11/11 units, `.cartulary/test-results/20261008T191402Z-p10199` |
| Restore browser slice after TLS retirement correction | PASS, 11/11 units, `.cartulary/test-results/20261008T191503Z-p49457`; 15 cleanup stages succeeded, streams closed, exit 0, no forced termination |
| `make json-shape-check` | PASS, `.cartulary/test-results/20261008T191621Z-p91080` |
| `make generate-drift` | PASS, `.cartulary/test-results/20261008T191628Z-p91627` |
| `make agent-finalize` before visual verification | PASS, `.cartulary/test-results/20261008T191648Z-p96128`; includes catalog/tier validation |
| Post-refresh JSON shape, catalog and generated-artifact policy | PASS; schema root `.cartulary/test-results/20261008T194345Z-p78424`, policy root `.cartulary/test-results/20261008T194354Z-p79350`; standalone catalog check exited 0 |
| Post-refresh generated drift | PASS, `.cartulary/test-results/20261008T194400Z-p79931` |
| First final ordinary visual comparison | PASS, 96/96 units, `.cartulary/test-results/20261008T194420Z-p84772`; all 255 accepted digests and capture mappings retained |
| Second final ordinary visual comparison | PASS, 96/96 units, `.cartulary/test-results/20261008T195843Z-p66714`; same 255 digests and capture declarations as the first pass and current checkout |
| Documentation lint | PASS, `.cartulary/test-results/20261008T201528Z-p48855`; final status and limitations update also checked before delivery |

The full run's two nonvisual assertion failures were stale entry/geometry
expectations: Retry had moved into Navigation details, and the 390-pixel recovery
test compared only horizontal bounds of controls now in different bands. Both
were corrected without removing their workflow assertions. Other failures were
visual comparisons and their target summaries. Review of every changed image
and two fresh ordinary visual passes now close that exit.

The source-impact boundary after the full run comprises the sole-use footer
ordering, these test entry paths, separate restore process/readiness fixtures,
unused TLS connection retirement and multiline redaction. Focused frontend,
browser, Go, harness-contract and schema checks above cover those changes.
Unchanged passing full-run rows remain retained evidence, not a claim that the
failed full run became green. The whole `make test` aggregate and `make check`
were not repeated after the narrow corrections: their changed scope was checked
through the focused slices and two complete ordinary visual runs. CI and release
aggregate membership were validated by the routing regressions; those release
aggregates were not executed for this implementation handoff. Retained-run
maintenance was skipped by
`agent-finalize` because `RESULTS_DIR` was unset. No release qualification,
publication claim or benchmark comparison is made.

## Acceptance assessment

The offline advisory query for layout shift and keyboard focus returned four
results without fallback. R002 is adopted only for the existing owner-defined
visible/restored focus obligations; enhanced AAA advice does not establish a new
conformance claim. Font-loading advice is outside this slice. R010 and R014 align
with the existing stable-layout and late-navigation-fencing obligations. No
upstream palette, component system or framework prescription was imported.

| Criterion | Assessment and evidence |
| --- | --- |
| A001 Authority | PASS: exact owner/source/verification map above; documents remain outside executable inputs. |
| A002 Scope | PASS: feedback placement, attempt completion and child-process lifetime each hide a real owned decision; former inline notice and local process wrapper retired. |
| A003 Repository state | PASS: recorded branch/commit, existing edits preserved; current source/import ownership and generated policy inspected; new footer test registered to `web.workbook`. |
| A004 Tokens | PASS: existing workbook styles, tokens, placement and responsive bands; reserved indicator space follows its font. No token registry added. |
| A005 Theme | PASS: graphite renderer retained; no theme controls or palette introduced. Final pixel review and both ordinary visual runs passed. |
| A006 Density | N/A: density resolution, row/header/editor geometry and selection are unchanged. Existing density fixtures remain in visual validation. |
| A007 Creation | N/A: no capabilities, writable inputs, source identity or creation policy changed. Existing ordinary recovery scenario remains selected. |
| A008 Responsive | PASS: held-response geometry at 1440×900, 1024×720 and 768×640; 390×480 footer/account reachability and bounded keyboard details. Responsive thresholds and inspector clamps are unchanged. |
| A009 Overflow | PASS: no document scroll or footer/control displacement in the held-response scenario; narrow footer/save visibility retained. |
| A010 Inspector | N/A: inspector dispatch, review ownership and draft lifetime are unchanged; handoff supersession still exercises retained Work. |
| A011 Continuity | PASS: current accepted-request/presentation matching; cancellation fences late attachment; original saved-view persistence and corrected supersession scenarios pass. |
| A012 Transactions | N/A: no mutation identity, captured bytes, replay or write admission changes. |
| A013 Acknowledgement and recovery | N/A: navigation read recovery does not alter mutation acknowledgement or retry semantics. |
| A014 Editing | N/A: editor commit/cancel and raw-draft ownership are unchanged; navigation cancels only its attachment/focus intent. |
| A015 Conflict | N/A: no cell conflict presentation or saved/local value change. |
| A016 Query and interaction states | PASS: completion observes the existing accepted-query and presentation owners; permission remains independently owned. No new data or permission store. |
| A017 Refresh and authorization | PASS: readable guard prevents completion/reveals after loss; footer concealment is unit-tested, and retained authorization/navigation rows passed in the full run. No revocation policy changed. |
| A018 Evidence | N/A: evidence lifecycle, preview and source meaning unchanged. Harness lifecycle diagnostics do not become product evidence semantics. |
| A019 Accessibility | PASS: semantic named controls, polite status, stable non-color indicator, keyboard details/escape restoration and corrected narrow recovery. No modal trap or animation introduced. |
| A020 Components | PASS: fixed footprint, bounded details, desktop/compact states, and final reviewed density/zoom/spacing/overflow fixtures. |
| A021 Virtualization | N/A: Grid Adapter row identity, virtualization and query paging remain unchanged; new completion observes existing semantic focus. |
| A022 Visual fixtures | PASS: transactional update, all 149 changed images inspected, unchanged capture declarations, and two fresh ordinary passes against the same manifest. |
| A023 Selectors | PASS: shared navigation selector and attempt/outcome state; saved-view/view-schema IDs and semantic roles retained. |
| A024 Test authority | PASS: no tests, runtime, generators or evidence machinery depend on Markdown; advisory lookup was manual review only. |
| A025 Generated artifacts | PASS: authored catalog/schema/topology first, Make generation and drift/schema checks afterward. Final golden publication uses only its Make owner. |
| A026 Compatibility | PASS: no persisted-data migration; coordinated artifact-schema cutover; archival prior runs are not translated into current evidence; rollback scope recorded. |
| A027 Handoff | PASS: final visual exit, exact refresh inventory, documentation validation, migration/rollback and limitations recorded; every review session has complete cleanup evidence. |

## Visual refresh inventory

The accepted trigger is the adopted stable Navigation control in existing footer
chrome. The transactional update passed 96/96 units in
`.cartulary/test-results/20261008T191721Z-p1108`. Its reconciliation accounts for
255 active captures/goldens and 29 registered fixtures, with zero missing goldens,
orphans, ambiguous mappings or unresolved fixtures. All 255 capture declarations
and consumer mappings match the preceding ordinary run. No viewport, browser
zoom, font, renderer pin, mask, scroll normalization, screenshot scope or
comparison tolerance changed.

The 149 changed files span 27 semantic owner rows and 18 registered fixtures.
Every changed image was imported through the artifact-review harness, checked
against its returned byte length and SHA-256, and visually inspected at useful
resolution. Crops covered every exact differing pixel; full 390-pixel workbook
frames and an old/current lifecycle comparison supplied surrounding context.
There were 116 footer-only changes. The remaining 33 images additionally contain
1–51 differing pixels outside the bottom 64 source pixels: channel deltas are
almost entirely 1–2 levels, with one 14-level delta. These do not change content
or geometry and were not independent reasons to refresh a golden. Review found
no unexpected clipping, overlap, focus, typography, overflow or missing data.
All actual footer changes follow the adopted Navigation placement; existing
comparison tolerances remain unchanged.

The first private review session, `.cartulary/test-results/20261008T191555Z-p88024`,
was interrupted with exit 143; its sender/cause was not established. Its terminal
receipt confirms complete cleanup. Unconsumed imports were reimported, not counted
as inspected. Review completed in the following artifacts-mode sessions, each
explicitly stopped with a digest-verified terminal receipt reporting complete
cleanup and a successful foreground exit:

- `.cartulary/test-results/20261008T193555Z-p6742`
- `.cartulary/test-results/20261008T193928Z-p33153`
- `.cartulary/test-results/20261008T194143Z-p58924`

All caller-owned review request scratch was removed afterward. Private images
and links expired on cleanup; the committed goldens and structural receipts are
the durable references. Image imports provided no DOM or axe channel; browser
and accessibility assertions above supply separate interaction evidence. No
interactive HTML report was needed or inspected.

All filenames below are relative to
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`. This is a human refresh record,
not an additional fixture registry. “Unregistered capture” means a valid
catalog-bound capture without a claimed registered fixture.

**`module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1`** — owner `module.collaboration`.

- `collaboration-conflict-resolver-linux.png` — `visual.fixture.same_field_conflict`.
- `collaboration-conflict-strip-linux.png` — `visual.fixture.save_state_strip`.
- `collaboration-presence-markers-linux.png` — `visual.fixture.presence_overflow`.
- `collaboration-recovered-saved-strip-linux.png` — `visual.fixture.save_state_strip`.

**`module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c`** — owner `module.collaboration`.

- `collaboration-grid-conflict-resolver-compact-linux.png` — unregistered capture.
- `collaboration-grid-conflict-resolver-linux.png` — unregistered capture.
- `collaboration-grid-conflict-resolver-narrow-linux.png` — unregistered capture.

**`module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc`** — owner `module.collaboration`.

- `collaboration-grid-blocked-conflict-linux.png` — unregistered capture.
- `collaboration-grid-recovered-saved-strip-linux.png` — unregistered capture.
- `collaboration-grid-saved-strip-linux.png` — unregistered capture.
- `collaboration-grid-syncing-strip-linux.png` — `visual.fixture.save_state_strip`.

**`module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`** — owner `module.entities`.

- `entity-mention-chip-states-linux.png` — `visual.fixture.mention_chip_state_matrix`.

**`module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7`** — owner `module.entities`.

- `record-relationships-mention-chips-linux.png` — unregistered capture.

**`module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`** — owner `module.networkflow`.

- `network-flow-analysis-accepted-inspector-linux.png` — `visual.fixture.claimed_network_analysis_workspace_states`.
- `network-flow-analysis-compact-saved-graphs-linux.png` — `visual.fixture.claimed_network_analysis_compact_workspace`.
- `network-flow-analysis-delete-dialog-linux.png` — `visual.fixture.claimed_network_analysis_workspace_states`.
- `network-flow-analysis-graph-contributors-linux.png` — `visual.fixture.claimed_network_analysis_workspace_states`.
- `network-flow-analysis-mapping-dialog-linux.png` — `visual.fixture.claimed_network_analysis_workspace_states`.
- `network-flow-analysis-narrow-query-controls-linux.png` — `visual.fixture.claimed_network_analysis_narrow_workspace`.
- `network-flow-analysis-rejected-diagnostics-linux.png` — `visual.fixture.claimed_network_analysis_workspace_states`.
- `network-flow-analysis-saved-graph-result-linux.png` — `visual.fixture.claimed_network_analysis_workspace_states`.

**`module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc`** — owner `module.savedviews`.

- `workbook-query-empty-text-spacing-linux.png` — `visual.fixture.empty_successful_query`.
- `workbook-query-empty-zoom-200-linux.png` — `visual.fixture.empty_successful_query`.
- `workbook-query-saved-view-query-controls-linux.png` — `visual.fixture.saved_view_query_controls_and_grouped_result`.
- `workbook-view-bar-filter-editing-overflow-linux.png` — unregistered capture.
- `workbook-view-bar-long-columns-linux.png` — unregistered capture.
- `workbook-view-bar-maximum-pressure-base-linux.png` — unregistered capture.
- `workbook-view-bar-maximum-pressure-compact-linux.png` — unregistered capture.
- `workbook-view-bar-maximum-pressure-narrow-linux.png` — unregistered capture.
- `workbook-view-bar-ordered-maximum-sort-linux.png` — unregistered capture.
- `workbook-view-bar-saved-view-actions-linux.png` — unregistered capture.
- `workbook-view-bar-saved-view-clean-linux.png` — unregistered capture.
- `workbook-view-bar-saved-view-modified-linux.png` — unregistered capture.
- `workbook-view-bar-text-spacing-linux.png` — unregistered capture.
- `workbook-view-bar-zoom-200-linux.png` — unregistered capture.

**`module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206`** — owner `module.timeline`.

- `timeline-grid-timeline-default-linux.png` — unregistered capture.

**`module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf`** — owner `module.timeline`.

- `timeline-grid-conflict-strip-linux.png` — unregistered capture.
- `timeline-grid-saved-strip-linux.png` — unregistered capture.
- `timeline-grid-syncing-strip-linux.png` — unregistered capture.

**`module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0`** — owner `module.workbook`.

- `incident-directory-compact-desktop-workbook-shell-linux.png` — `visual.fixture.compact_desktop_workbook_shell`.
- `incident-directory-default-timeline-workbook-shell-linux.png` — `visual.fixture.default_timeline_workbook_shell`.
- `incident-directory-narrow-desktop-workbook-shell-linux.png` — `visual.fixture.narrow_desktop_workbook_shell`.

**`module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`** — owner `module.workbook`.

- `workbook-inspector-attached-edit-linux.png` — `visual.fixture.base_inspector`.
- `workbook-inspector-compact-actions-linux.png` — `visual.fixture.inspector_compact_actions`.
- `workbook-inspector-details-linux.png` — `visual.fixture.base_inspector`.
- `workbook-inspector-history-linux.png` — `visual.fixture.base_inspector`.
- `workbook-inspector-narrow-technical-details-linux.png` — `visual.fixture.inspector_narrow_technical_details`.
- `workbook-inspector-public-error-linux.png` — `visual.fixture.base_inspector`.
- `workbook-inspector-relationships-linux.png` — `visual.fixture.base_inspector`.
- `workbook-inspector-retained-draft-linux.png` — `visual.fixture.base_inspector`.
- `workbook-inspector-rollback-preview-linux.png` — `visual.fixture.base_inspector`.

**`module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67`** — owner `module.workbook`.

- `timeline-mutation-pending-replay-status-linux.png` — unregistered capture.
- `timeline-mutation-transaction-recovery-panel-compact-linux.png` — unregistered capture.
- `timeline-mutation-transaction-recovery-panel-linux.png` — unregistered capture.
- `timeline-mutation-transaction-recovery-panel-narrow-linux.png` — unregistered capture.

**`module.workbook.visual.contextual_task_decision_creation`** — owner `module.workbook`.

- `contextual-decision-authoring-linux.png` — unregistered capture.
- `contextual-decision-authoring-narrow-linux.png` — unregistered capture.
- `contextual-decision-recovery-linux.png` — unregistered capture.
- `contextual-decision-recovery-narrow-linux.png` — unregistered capture.
- `contextual-decision-references-narrow-linux.png` — unregistered capture.
- `contextual-task-request-authoring-linux.png` — unregistered capture.
- `contextual-task-request-authoring-narrow-linux.png` — unregistered capture.
- `contextual-task-request-recovery-linux.png` — unregistered capture.
- `contextual-task-request-recovery-narrow-linux.png` — unregistered capture.
- `contextual-task-request-references-narrow-linux.png` — unregistered capture.

**`module.workbook.visual.coordination_create_authoring_recovery`** — owner `module.workbook`.

- `coordination-comm-log-authoring-linux.png` — `visual.fixture.contextual_coordination_creation`.
- `coordination-handoff-authoring-linux.png` — `visual.fixture.contextual_coordination_creation`.
- `coordination-lesson-authoring-linux.png` — `visual.fixture.contextual_coordination_creation`.
- `coordination-recovery-linux.png` — `visual.fixture.contextual_coordination_creation`.
- `coordination-recovery-narrow-linux.png` — `visual.fixture.contextual_coordination_creation`.
- `coordination-source-narrow-linux.png` — `visual.fixture.contextual_coordination_creation`.
- `coordination-status-review-authoring-linux.png` — `visual.fixture.contextual_coordination_creation`.

**`module.workbook.visual.decision_supersession_review_recovery`** — owner `module.workbook`.

- `decision-supersession-accepted-linux.png` — unregistered capture.
- `decision-supersession-review-linux.png` — unregistered capture.
- `decision-supersession-review-narrow-linux.png` — unregistered capture.

**`module.workbook.visual.indicator_lifecycle_authoring`** — owner `module.workbook`.

- `indicator-lifecycle-authoring-linux.png` — `visual.fixture.indicator_lifecycle_authoring`.
- `indicator-lifecycle-authoring-narrow-linux.png` — `visual.fixture.indicator_lifecycle_authoring`.

**`module.workbook.visual.indicator_observations_authoring`** — owner `module.workbook`.

- `indicator-observation-authoring-linux.png` — `visual.fixture.indicator_observations_authoring`.
- `indicator-observation-authoring-narrow-linux.png` — `visual.fixture.indicator_observations_authoring`.

**`module.workbook.visual.note_create_authoring_recovery`** — owner `module.workbook`.

- `linked-note-authoring-linux.png` — unregistered capture.
- `linked-note-authoring-narrow-linux.png` — unregistered capture.
- `linked-note-recovery-linux.png` — unregistered capture.
- `linked-note-recovery-narrow-linux.png` — unregistered capture.
- `linked-note-source-narrow-linux.png` — unregistered capture.

**`module.workbook.visual.ordinary_create_authoring_recovery`** — owner `module.workbook`.

- `ordinary-closed-retained-narrow-linux.png` — unregistered capture.
- `ordinary-recovery-1280-linux.png` — unregistered capture.
- `ordinary-recovery-390-linux.png` — unregistered capture.
- `ordinary-reference-authoring-linux.png` — unregistered capture.

**`module.workbook.visual.preferences`** — owner `module.workbook`.

- `workbook-preferences-comfortable-linux.png` — unregistered capture.
- `workbook-preferences-compact-linux.png` — unregistered capture.
- `workbook-preferences-confirmed-stale-linux.png` — unregistered capture.
- `workbook-preferences-uncertain-linux.png` — unregistered capture.
- `workbook-preferences-uncertain-narrow-linux.png` — unregistered capture.
- `workbook-preferences-unset-linux.png` — unregistered capture.

**`module.workbook.visual.timeline_capture_actions`** — owner `module.workbook`.

- `timeline-supersession-accepted-linux.png` — unregistered capture.
- `timeline-supersession-authoring-linux.png` — unregistered capture.
- `timeline-supersession-review-linux.png` — unregistered capture.
- `timeline-supersession-review-narrow-linux.png` — unregistered capture.

**`module.workbook.visual.timeline_related_evidence`** — owner `module.workbook`.

- `timeline-related-evidence-authoring-linux.png` — unregistered capture.
- `timeline-related-evidence-authoring-narrow-linux.png` — unregistered capture.
- `timeline-related-evidence-partial-linux.png` — unregistered capture.
- `timeline-related-evidence-partial-narrow-linux.png` — unregistered capture.
- `timeline-related-evidence-party-narrow-linux.png` — unregistered capture.

**`web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc`** — owner `web.design`.

- `account-menu-controls-compact-linux.png` — unregistered capture.
- `account-menu-controls-narrow-linux.png` — unregistered capture.
- `account-menu-workbook-root-linux.png` — unregistered capture.

**`web.design.visual.lifecycle`** — owner `web.design`.

- `lifecycle-closed-linux.png` — unregistered capture.
- `lifecycle-comfortable-linux.png` — unregistered capture.
- `lifecycle-compact-linux.png` — unregistered capture.
- `lifecycle-confirmed-refresh-failure-linux.png` — unregistered capture.
- `lifecycle-pending-linux.png` — unregistered capture.
- `lifecycle-reason-linux.png` — unregistered capture.
- `lifecycle-review-linux.png` — unregistered capture.
- `lifecycle-review-narrow-linux.png` — unregistered capture.
- `lifecycle-review-spacing-linux.png` — unregistered capture.
- `lifecycle-review-zoom-linux.png` — unregistered capture.
- `lifecycle-uncertain-linux.png` — unregistered capture.

**`web.design.visual.membership_audit_browsing`** — owner `web.design`.

- `membership-audit-comfortable-linux.png` — unregistered capture.
- `membership-audit-compact-linux.png` — unregistered capture.
- `membership-audit-cursor-recovery-linux.png` — unregistered capture.
- `membership-audit-empty-linux.png` — unregistered capture.
- `membership-audit-inspected-linux.png` — unregistered capture.
- `membership-audit-inspected-narrow-linux.png` — unregistered capture.
- `membership-audit-inspected-spacing-linux.png` — unregistered capture.
- `membership-audit-inspected-zoom-linux.png` — unregistered capture.
- `membership-audit-loading-linux.png` — unregistered capture.
- `membership-audit-stale-linux.png` — unregistered capture.

**`web.design.visual.membership_management_visual`** — owner `web.design`.

- `membership-management-comfortable-linux.png` — unregistered capture.
- `membership-management-compact-linux.png` — unregistered capture.
- `membership-management-confirmed-refresh-failure-linux.png` — unregistered capture.
- `membership-management-loading-linux.png` — unregistered capture.
- `membership-management-pending-linux.png` — unregistered capture.
- `membership-management-removal-narrow-linux.png` — unregistered capture.
- `membership-management-removal-spacing-linux.png` — unregistered capture.
- `membership-management-removal-zoom-linux.png` — unregistered capture.
- `membership-management-role-linux.png` — unregistered capture.
- `membership-management-uncertain-linux.png` — unregistered capture.

**`web.design.visual.metadata_editing`** — owner `web.design`.

- `metadata-closed-linux.png` — unregistered capture.
- `metadata-comfortable-linux.png` — unregistered capture.
- `metadata-compact-linux.png` — unregistered capture.
- `metadata-confirmed-refresh-failure-linux.png` — unregistered capture.
- `metadata-conflict-linux.png` — unregistered capture.
- `metadata-dirty-linux.png` — unregistered capture.
- `metadata-loading-linux.png` — unregistered capture.
- `metadata-review-narrow-linux.png` — unregistered capture.
- `metadata-review-spacing-linux.png` — unregistered capture.
- `metadata-review-zoom-linux.png` — unregistered capture.
- `metadata-saving-linux.png` — unregistered capture.
- `metadata-uncertain-linux.png` — unregistered capture.
