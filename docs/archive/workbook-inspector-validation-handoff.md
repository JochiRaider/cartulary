# Workbook inspector validation and handoff

Status: **DONE**, including WS-10 validation and handoff (2026-09-21). The [controlling tracker](workbook-inspector-ui-ux-audit.md)
owns workstream status, decisions and sequencing. This document records evidence;
it does not define product requirements or become an executable input.

## Delivered boundary

The adopted presentation now separates saved record reading, connections and
explicit local actions. One admitted descriptor sequence drives mounted inspector
sections and navigation. The fixed record frame keeps Close and Sections available;
passive scrolling changes only the current-section indication. Details preserve
schema order, distinguish unloaded/null/empty/false/zero, and disclose long values
and field-local editors. Closing retains unfinished work; Update submits it.
Relationships retain source identity while disclosing correction controls. Evidence
separates accepted information, authorized access and attachment recovery. Workflow
rows explain whether an action opens authoring, opens review or navigates.
Specialized forms share presentation while keeping their source-owned lifetimes.

History now uses a closed semantic contract with source-owned canonical-fact
projectors, checked catalog admission, deterministic identities/order, typed
before/after presence, public references and imported attribution. Ten record
snapshot types and fourteen mutation target kinds are contributed by nine owners.
No central source-type switch, raw snapshot disclosure, legacy decoder or response
adapter was introduced. Required detail cannot fall back to a successful empty
result. Browsing generations are compared before committed content; disposable
reads restart without retiring captured operations or drafts.

The exact owner-to-change map, G00–G11 rationale/migration/risks, WS-00–09 checkpoints,
source coverage, seventeen-schema inventory and optional-feature N/A decisions are
in controlling-tracker §§4–10. Core 01 REQ-01-052A, Core 02 REQ-02-218/265,
Core 03 REQ-03-139 and design §12.7/D-AC-082–086 own the adopted changes.
Domain vocabulary/navigation required no correction; `docs/domain.md` is unchanged.

## Repository and retirement review

Baseline: `main` at `67f3d98ac01b42a1da4409c0a6630d08d24e1761`.
Only the pre-staged controlling audit existed at entry. Implementation was made in
the shared working tree; the index was preserved. There is no commit, PR or deployment.
An isolated detached baseline checkout under `tmp/workbook-inspector-baseline`
provides rollback build/smoke evidence; it does not change the working branch.

Removed the old always-expanded field form, Done editing wording, flat Workflow
button presentation, repeated History action layout and arbitrary History-unit
object contract. Removed `observationStyles.ts` and `indicatorLifecycleStyles.ts`
and migrated thirteen Indicator consumers to shared form roles. No replacement
form engine, retry engine, theme, density preference or durable draft store exists.

Retained shared exports have concrete consumers: Timeline style aliases serve its
grid/draft controls and consume the shared role values; the compact grid action role
serves Evidence access and attachment siblings inside density-owned grid geometry;
the generic related-row adapter/seed renderer still serves `view_row_create` routes;
feature-owned draft/operation/recovery stores retain materially distinct lifetimes.
History selectors, mutation envelopes, `/api/v1` routes and idempotency remain owned
by their existing contracts. Generated outputs were produced through Make from
changed authored OpenAPI/design inputs, never hand-edited. No SQL migration or
stored-history rewrite is required.

## Validation ledger

All run identifiers below are relative to `.cartulary/test-results/` unless a
baseline checkout is named. `run-manifest.json`, `run-summary.json`, row receipts,
unit logs and browser group reports preserve the exact expanded selections.
Commands ran from the repository root with the provisioned Node runtime on PATH.
Workstream-specific commands and results remain in the controlling checkpoints.

| Command / selection | Result | Run or artifact |
| --- | --- | --- |
| `make agent-finalize` | PASS before broader verification; retained-run maintenance SKIPPED because RESULTS_DIR was unset | `20260921T073043Z-p91957` (earlier checks also passed) |
| `make test-fast` | PASS 708/708 units on final source, including clipboard, canonical-cell and access-scope corrections | `20260921T073129Z-p97071` |
| `make browser-e2e-stateful` | Initial PASS 42/42; final source 39/42, with two isolated scenario resolutions below | Initial `20260921T062456Z-p87023`; final `20260921T073129Z-p96952` |
| `make browser-e2e-a11y` | Initial PASS 20/20; final source 18/20, one coordination focus scenario passed unchanged in isolation; all other groups pass | Initial `20260921T065248Z-p10254`; final `20260921T073453Z-p32834`; resolution `20260921T074150Z-p32164` |
| `make browser-e2e-webserver-backed` | Initial 127/137 units; all nine failed scenarios resolved with focused reruns below | `20260921T061637Z-p18330` |
| `make service-backed-test-slice OWNER=module.revisions` | Initial canonical-fixture failures resolved by three selected rows, 4/4 units; remaining rows passed full run | Full `20260921T061747Z-p63365`; resolution `20260921T062140Z-p68836` |
| `make generate` | PASS after final fixture registration | `20260921T073025Z-p88182` |
| `make format frontend-typecheck lint-biome` | PASS after final browser fixture stabilization | `20260921T074057Z-p26499/-p26458/-p26486` |
| `make generate-drift generated-artifact-policy-check json-shape-check openapi-compatibility-check` | PASS | Drift `20260921T073129Z-p96719`; policy/shape/compatibility `20260921T073025Z-p88186/-p88188/-p88194` |
| `make lint-markdown`; staged/unstaged `git diff --check` | PASS | Markdown `20260921T075527Z-p42495`; final whitespace review clean |
| `make frontend-import-boundary-check` | PASS | `20260921T073026Z-p88246` |
| `make build-server build-web` | PASS 4/4 and 2/2; final paired candidate | `20260921T073129Z-p97267/-p97300` |
| `make browser-e2e-visual-update` | PASS 12/12; final fixture-stability refresh and all changed-image review complete | `20260921T074355Z-p30667` |
| `make browser-e2e-visual` | PASS 12/12 in each of two fresh ordinary confirmations after the final reviewed refresh | `20260921T074928Z-p69256`, `20260921T074928Z-p69271` |

### Failure resolution

No production admission was relaxed to make a fixture pass. Synthetic backend Host
and tag snapshots were corrected to complete canonical retained facts; pagination
uses valid operation/target identities. The nonreversible rollback atomicity fixture
now uses a canonical retained rollback transition instead of a malformed tag or
missing-current-row case with the wrong expected status. Production rejects schema-less
and malformed snapshots, as Core 02 requires.

Browser fixtures now explicitly open fields, Update ordinary edits, open alias or
mention disclosure and inspect exact timestamp detail. Grid autosave scenarios still
use grid editing. The clipboard scenario exposed loss of a trailing newline in native
rendered-text copy; the saved-value renderer now copies a wholly contained selection's
exact text range, leaving cross-field and collapsed selections to the browser.
The test continues to check exact bytes, read-only copy, and no cut/paste mutation.
A second integrated failure exposed canonical nulls being reconstructed as empty
text by Timeline committed-version tracking. That reconstruction and the parallel
discard-recovery raw-cell overlay were removed. Accepted `rawRow` cells now remain
canonical; display strings and queued work cannot become saved facts. Unit coverage
checks null, empty and absent cells plus pending FIFO work; the post-review edit
browser regression exercises the authoritative baseline end to end.
A later stateful run exposed an access-scope error hidden by the former raw-cell
reconstruction: a saved-view presentation reset retired a still-authorized source
object. All three retained-row consumers now distinguish the incident/account/access
epoch from the presentation reset key. Retention retires the previous observation,
allowing a fresh authorized row to arrive with a new scope while rejecting the old
row after access loss. The retention/Timeline composition regression passed 3/3 at
`20260921T072924Z-p22616`. The saved-view/role/closure browser row passed 11/11
at `20260921T072925Z-p22843` after this correction. The earlier full stateful
run `20260921T071155Z-p70862` (40/42) and focused attempts
`20260921T071708Z-p29741` / `20260921T072104Z-p76080` remain failed evidence,
resolved by that correction and the passing inspector scenarios in the final full rerun.
History receipt reconciliation may settle while its Refresh control is being
activated; the test accepts disappearance only after asserting the completed receipt.
No acknowledged write is replayed to recover a read.

The initial apparent visual orphan was an active mention scenario failing before
capture. Correcting its source-group assertion restored its mapping; no PNG was
deleted. Initial accessibility failures were stale always-expanded controls and a
combined Evidence-container assumption. The geometry check now scrolls the complete
access/attachment sibling group into view and still checks all four controls.
No contrast, clipping or interaction expectation was weakened.

The final stateful run also encountered an unrelated Network Flow transient-state
assertion: materialization had already succeeded before the test observed its
refresh-in-progress message (`20260921T073129Z-p96952`). The failure snapshot
shows the accepted refresh, preserved graph and completed materialization. The
unchanged owner row passed 11/11 at `20260921T073603Z-p78561` using
`make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle`.
No inspector fix changes that lifecycle; the full-target failure is not relabeled
as a full-target pass.

The final stateful Timeline transaction-recovery scenario timed out while opening
a subsequent grid editor after Retry; its unchanged complete owner row passed
11/11 at `20260921T074157Z-p46106` using
`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.verify_timeline_public_mutations_remount_safe_tr_84f645d6c5`.
The final accessibility run had one transient Coordination Retry refresh focus
assertion; the unchanged full scenario passed 11/11 at `20260921T074150Z-p32164`
using `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.coordination_create_authoring_recovery`.
All other scenarios in those final runs passed. These isolated passes resolve the
observed failures without weakening focus or recovery assertions; intermittent
timing remains recorded for future harness hardening, not erased from the run history.

Transient build-snapshot races while formatting (`20260921T062837Z-p83116`) and the
cancelled repeated fast run (`20260921T063720Z-p89235`, 201/708 completed) provide no
passing evidence. Final-source fast verification passed 708/708 at `20260921T073129Z-p97071`. A premature `explain-run` on the still-running browser
target had no completed summary to read. An OWNER argument to the full visual target
was rejected; scoped visual selection subsequently used the public owner slice.

### Full browser run: resolved rows

Every failing browser scenario from `20260921T061637Z-p18330` is accounted for
below. These focused public reruns resolve the initial failures; the original full
run remains recorded as failed, not relabeled as a full-target pass.

| `make service-backed-test-slice` selection | Resolution run | Result |
| --- | --- | --- |
| `OWNER=module.entities ROWS=module.entities.browser.verify_hosts_identities_and_notes_grids_render_c_7531b57a50` | `20260921T063316Z-p24751` | PASS 11/11; explicit Generic field editor |
| `OWNER=module.evidence ROWS=module.evidence.browser.requested_evidence_can_be_tracked_before_the_blo_6a619b85fd` | `20260921T063317Z-p24970` | PASS 11/11; explicit Generic field editor |
| `OWNER=module.collaboration ROWS=module.collaboration.browser.live_updates_never_retarget_pending_local_edits_11699e9b2d` | `20260921T063320Z-p27085` | PASS 11/11; real clipboard paste and explicit grid Enter |
| `OWNER=module.workbook ROWS=module.workbook.browser.entity_find_characterization` | `20260921T063718Z-p88920` | PASS 11/11; native Manage aliases disclosure |
| `OWNER=module.workbook ROWS=module.workbook.browser.indicator_observations_paging_source_edits` | `20260921T065243Z-p5038` | PASS 11/11; explicit source Edit/Update, independent target retained |
| `OWNER=module.revisions ROWS=module.revisions.browser.history_recovery_074ea859ee8b` | `20260921T070156Z-p92528` | PASS 11/11; exact replay, completed receipt and read-only reconciliation |
| `OWNER=module.tasksdecisions ROWS=module.tasksdecisions.browser.the_browser_workbook_opens_task_requests_and_dec_8048a75ceb` | `20260921T070700Z-p60108` | PASS 11/11; close inspector for grid query, reopen for selected Task workflow |
| `OWNER=module.timeline ROWS=module.timeline.browser.a_reviewer_session_browser_flow_visibly_performs_c3aaf62d89,module.timeline.browser.scalar_clipboard_composition` | `20260921T071112Z-p35368` | PASS 13/13; reviewer edit/demotion/supersede and exact multiline read-only clipboard |

The final Timeline source correction also passed
`make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_mutation_models_a110000001,module.timeline.frontend.timeline_row_mutation_coordinator_1a7e2c9b44`
at `20260921T071111Z-p35060` (3/3 units). Temporary preparation diagnostics used to
identify the null/empty mismatch were removed; no diagnostic output remains.

## Visual refresh review

Accepted trigger: adopted shared typography/component roles and the inspector
reading/navigation/action redesign. The review includes shared-role consumers in
account, lifecycle, metadata and Network surfaces; those changes are expected
consequences of their named shared-role consumption.

Ordinary pre-update reconciliation is retained in
`20260921T055649Z-p58675/browser-e2e-visual/frontend-visual-reconciliation.json`.
Focused ordinary runs `20260921T060903Z-p1247` and `20260921T061549Z-p53155`
reached all base/mention capture intents before the public refresh. The first update
passed at `20260921T061746Z-p63050`. All **151 changed images** were manually reviewed
using 26 six-image sheets plus full-resolution critical Details/editor/History views.
A subsequent review found an unmasked formatted History clock and the old editor
legend. The final update changed only attached-edit, History and public-error images;
all three were reviewed again at full resolution. The canonical-null correction
then changed Details/attached-edit/retained-draft text, and the explicit Evidence
anchor changed its scroll placement. All four were reviewed at full resolution
in `20260921T073129Z-p97021` (12/12 update units). This update also exposed two
Network Flow captures racing an exploration read; the fixture now awaits its
visible edge before detaching exploration. Final ordinary reconciliation and
public refresh passed 12/12 at `20260921T074355Z-p30667`; its only two replaced
images were reviewed at full resolution and match the previously reviewed state.
The corrected Network Flow ordinary captures
at `20260921T074150Z-p32184` are byte-identical to their previously reviewed
goldens; no Network Flow product behavior or expected saved-result state changed.
Manual review records are retained in `tmp/workbook-inspector-visual-review.json`,
`tmp/workbook-inspector-final-image-review.json`,
`tmp/workbook-inspector-canonical-image-review.json` and
`tmp/workbook-inspector-stability-image-review.json`; these are local human-review
notes, not executable requirements or release certification.

Final update reconciliation accounts for **255 captures, 255 committed goldens,
255 active mappings, zero orphans, zero missing goldens, zero ambiguous mappings,
29 registered fixtures and zero unresolved fixtures**. Existing viewport, zoom,
renderer, screenshot crop and tolerances are unchanged. The Evidence fixture now
applies its declared section-start anchor after fonts and dynamic masks settle;
this removes timing-dependent scroll placement without hiding content.
The existing clock mask now also handles the adopted `UTC +00:00` display format.
Three new registered base-inspector captures show saved Details, an attached editor
and a detached retained draft at 1280×720, 100% zoom, dark_graphite/compact.

The table below is derived from the final update's exact reconciliation artifact,
not filename-based ownership guesses. Every listed image was reviewed and accepted.
A dash in the fixture column means an active nonregistry capture with an exact
catalog/scenario/project mapping; it is not an orphan. Committed paths share
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`.

| Golden filename | Semantic owner row | Stable fixture |
| --- | --- | --- |
| `account-menu-controls-compact-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-controls-narrow-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-controls-short-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-long-label-text-spacing-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-long-label-zoom-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-workbook-root-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `collaboration-conflict-resolver-linux.png` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.same_field_conflict` |
| `collaboration-grid-blocked-conflict-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc` | — |
| `collaboration-grid-conflict-resolver-compact-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-grid-conflict-resolver-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-grid-conflict-resolver-narrow-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-presence-markers-linux.png` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.presence_overflow` |
| `contextual-decision-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-decision-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-decision-recovery-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-decision-recovery-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-decision-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-recovery-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-recovery-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `coordination-comm-log-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-handoff-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-lesson-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-source-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-status-review-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `decision-supersession-accepted-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | — |
| `decision-supersession-review-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | — |
| `decision-supersession-review-narrow-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | — |
| `entity-mention-chip-states-linux.png` | `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `visual.fixture.mention_chip_state_matrix` |
| `evidence-affordance-states-linux.png` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` |
| `evidence-grid-available-evidence-linux.png` | `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af` | — |
| `evidence-grid-blocked-preview-linux.png` | `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830` | — |
| `evidence-grid-requested-evidence-linux.png` | `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af` | — |
| `incident-directory-compact-desktop-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.compact_desktop_workbook_shell` |
| `incident-directory-default-timeline-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.default_timeline_workbook_shell` |
| `incident-directory-narrow-desktop-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.narrow_desktop_workbook_shell` |
| `indicator-lifecycle-authoring-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-lifecycle-authoring-narrow-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-observation-authoring-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `indicator-observation-authoring-narrow-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `lifecycle-closed-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-comfortable-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-compact-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-confirmed-refresh-failure-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-pending-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-reason-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-narrow-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-spacing-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-zoom-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-uncertain-linux.png` | `web.design.visual.lifecycle` | — |
| `linked-note-authoring-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-authoring-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-recovery-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-recovery-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-source-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `membership-audit-comfortable-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-compact-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-cursor-recovery-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-empty-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-narrow-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-spacing-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-zoom-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-loading-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-stale-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-management-comfortable-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-compact-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-confirmed-refresh-failure-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-loading-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-pending-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-removal-narrow-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-removal-spacing-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-removal-zoom-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-role-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-uncertain-linux.png` | `web.design.visual.membership_management_visual` | — |
| `metadata-closed-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-comfortable-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-compact-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-confirmed-refresh-failure-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-conflict-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-dirty-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-loading-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-review-narrow-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-review-spacing-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-review-zoom-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-saving-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-uncertain-linux.png` | `web.design.visual.metadata_editing` | — |
| `network-flow-analysis-accepted-inspector-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-compact-saved-graphs-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_compact_workspace` |
| `network-flow-analysis-delete-dialog-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-graph-contributors-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-mapping-dialog-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-narrow-query-controls-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_narrow_workspace` |
| `network-flow-analysis-rejected-diagnostics-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-saved-graph-result-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `ordinary-closed-retained-narrow-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-recovery-1280-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-recovery-390-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-reference-authoring-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `record-relationships-evidence-access-linux.png` | `module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc` | — |
| `record-relationships-mention-chips-linux.png` | `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | — |
| `record-relationships-task-requests-linux.png` | `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc` | `visual.fixture.task_requests_or_decisions` |
| `timeline-grid-timeline-default-linux.png` | `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206` | — |
| `timeline-mutation-pending-replay-status-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-compact-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-narrow-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-related-evidence-authoring-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-authoring-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-partial-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-partial-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-party-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-supersession-accepted-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `timeline-supersession-authoring-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `timeline-supersession-review-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `timeline-supersession-review-narrow-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `workbook-inspector-attached-edit-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-compact-actions-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_compact_actions` |
| `workbook-inspector-destructive-confirmation-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.destructive_actions` |
| `workbook-inspector-details-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-history-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-public-error-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-relationships-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-retained-draft-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-preferences-comfortable-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-compact-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-confirmed-stale-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-uncertain-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-uncertain-narrow-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-unset-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-query-empty-text-spacing-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-zoom-200-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-saved-view-query-controls-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.saved_view_query_controls_and_grouped_result` |
| `workbook-view-bar-filter-editing-overflow-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-long-columns-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-maximum-pressure-base-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-maximum-pressure-compact-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-maximum-pressure-narrow-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-ordered-maximum-sort-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-saved-view-actions-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-saved-view-clean-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-saved-view-modified-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-text-spacing-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-zoom-200-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |

## Five-task comparison and usability limits

Method: baseline source/audit walkthrough, retained baseline History browser smoke,
candidate automated interaction tests, and manual review of the actual candidate
captures. No recruited participants or timed paired task study were available.
Automated assertion success is not a measurement of human completion, incorrect
actions, scroll distance, or understanding of saving versus closing.

| Same task | Baseline observation | Candidate behavior and completion evidence | Incorrect actions, scroll travel and interpretation |
| --- | --- | --- | --- |
| Locate Evidence | Discovery depends on scrolling to an admitted section | Sections → Evidence reaches the mounted region in two activations; count scope, access and attachment remain distinct; navigation and Evidence browser/a11y rows complete | Human errors and scroll pixels not measured; reduced search travel is a hypothesis. Count is not presented as a richer item catalog. |
| Edit, close and resume | Existing editing dominates reading; Done editing obscures whether work was saved | Explicit Edit → Update or Close editor; retained cue at the original field; Resume restores owner-held authoring. Details unit and inspector browser scenarios complete | No implicit submit/discard from blur, Tab, Escape or navigation in automated tests. Human interpretation of the retention explanation remains unmeasured. |
| Correct a mention | Expanded correction controls consume vertical space | Compact raw/state/authorized-target summary; source-local disclosure and selection; paging, exact replay, late outcomes and return focus complete in relationship/recovery tests | No observed assertion failure after fixture migration; participant incorrect-action rate and scroll travel not measured. Compactness benefit is a hypothesis. |
| Inspect History and reversal scope | Coarse/raw detail and repeated controls require interpretation | Actor, UTC display, complete typed units, exact event detail, event-local reversal and trailing Record actions; legal scopes/paging/recovery checks complete | Automated scopes/selectors checked; human understanding of whole-change-set reach and time display remains unmeasured. No measured time/error improvement claimed. |
| Create a related task | Equal-weight workflow controls and local form treatments compete | Canonical ordered action row explains opening authoring; shared form presentation retains explicit source and replacement/clearing semantics; ordinary/contextual creation and Task owner checks complete | Duplicate invocation and changed-source behavior are asserted. Participant completion, incorrect actions, scroll travel and source-link interpretation not measured. |

Next usability study should run those exact five tasks on the retained baseline and
candidate with the same seeded records, role and viewport. Record completion,
incorrect actions, actual scroll travel and the participant's explanation of
Update/Close/Resume. Until then, readability and effort reductions are residual
usability hypotheses, not product correctness blockers or asserted empirical gains.

## Coordinated rollout and rollback

The public History representation change is explicitly reviewed in the pending
OpenAPI **2.0.0** boundary (eleven reviewed changes). The immutable **1.0.0** baseline
is unchanged. `/api/v1`, opaque rollback selectors, write envelopes and idempotency
are unchanged. Deploy browser assets and server as one coherent bundle; there is no
dual protocol or legacy response adapter.

1. Review the working-tree source/spec/contract changes and their green acceptance
   evidence. Preserve the existing canonical database/object store and operation
   recovery state. Verify the bundle manifest and archive digest before promotion.
2. Deploy the candidate server and matching browser assets together. Existing
   admitted browser reads compare representation generation before content equality.
   Changed generations restart disposable reads; unsupported payloads fail locally.
   Do not force a reload that would lose local drafts, captured requests or receipts.
3. Smoke-test saved Details, explicit edit/close/resume, Evidence access, complete
   History detail/paging and one permitted scoped reversal using canonical records.
   Recheck imported source attribution and strict malformed-payload rejection.
4. If the deployment must be reverted, restore both members of the retained baseline
   bundle together. Do not rewrite history or re-key acknowledged writes. A browser
   with an incompatible representation fails the read safely; operation owners keep
   their captured work. Return to a matching browser bundle at a user-controlled
   transition that preserves/resolves in-session authoring, rather than automatic
   page reload.
5. Unsupported schema-less legacy databases remain at the existing operator-managed
   reset boundary in Core 02 REQ-02-265. This effort performs no automatic reset,
   shape inference, backfill, audit-data conversion or account-directory lookup.

Baseline builds passed (`build-server` 4/4, `build-web` 2/2) in the isolated checkout
at `20260921T061929Z-p41881/-p41889`. Baseline legal History opening and opaque entry
selectors passed 11/11 at `20260921T062437Z-p33489`; pagination also passed in the
initial baseline smoke. That first smoke (`20260921T062141Z-p69150`) had an existing
stale review-message assertion; the baseline was not modified to make it pass.
The baseline archive is `tmp/workbook-inspector-bundles/baseline.tar.gz`, SHA-256
`4ca835ef1e4612d5a9a83c4cc0a2f0b74c3fbb0b90443e694dfa80c88b65c519`.
Its companion manifest lists 73 server/web artifact hashes.
The final candidate builds passed at `20260921T073129Z-p97267/-p97300`;
`tmp/workbook-inspector-bundles/candidate.tar.gz` has SHA-256
`fd57095553cdbe1574aef01139b40f8bd08586824e9a0bd4a8105cc8eea86d68`.
All 73 archived server/web files were verified against `candidate-manifest.json`. These are retained local build
and browser smoke artifacts, not evidence of an external production deployment.

## Digest acceptance assessment

The digest is advisory. Its rows assess implementation evidence against the adopted
owners; no executable input reads the TSV or this report. Optional unbound features
are N/A only where the tracker WS-08 inventory cites Core 01 REQ-01-615
`omit_feature`. The seventeen configurations and specialized families remain in scope.

| Row | Outcome | Evidence and limit |
| --- | --- | --- |
| A001 Authority | PASS | Tracker §4 owner/source/verification map; adopted WS-00 clauses; Core 02 legacy prohibition corrected before source implementation. |
| A002 Scope | PASS | G00–G11 selection rationale, benefit, extension and retirement decisions; shared admission/navigation and presentation slots hide real common decisions without merging source lifetimes. |
| A003 Repository state | PASS | Baseline/index inventory, current source ownership manifest and import checks; generated boundaries and Grid Adapter ownership inspected; no direct vendor integration added. |
| A004 Tokens | PASS | Authored presentation projection and shared role migration; no theme/density registry or local visual palette; compact Evidence geometry consumes density-owned values. |
| A005 Theme | PASS | dark_graphite visual fixture coverage; all changed images manually reviewed; no additional theme exposed. |
| A006 Density | PASS | Complete a11y run `20260921T065248Z-p10254`, computed role/grid fixtures and ordinary visuals cover compact/default/comfortable and grid-specific geometry. No timing improvement claimed. |
| A007 Creation | PASS | WS-09 family checks and stateful coverage retain source replacement/clearing, ordinary minima and Evidence/Indicator exceptions; final Task workflow `20260921T070700Z-p60108` passes. |
| A008 Responsive | PASS | Full a11y and both ordinary visual passes; base/narrow/compact/below-minimum, width clamps, vertical resize, fallback and ARIA fixtures retain existing geometry ownership. |
| A009 Overflow | PASS | Shared one-body inspector, persistent header/navigation, overlay and account navigation fixtures; a11y/visual/stateful checks. |
| A010 Inspector | PASS | One descriptor sequence, semantic action deduplication, role/concealment and review invalidation tests; final Timeline source edit/review browser pass `20260921T071112Z-p35368`. |
| A011 Continuity | PASS | Passing stateful scenarios and source-local WS-03–09 suites cover retained drafts, explicit source transitions, late receipts, semantic focus and navigation without passive reads. |
| A012 Transactions | PASS | Existing Web Crypto ID owner unchanged; explicit-patch, History and specialized owner tests fence duplicate dispatch and preserve exact captured request identity/bytes. |
| A013 Acknowledgement and recovery | PASS | History before-commit exact replay + projection read recovery `20260921T070156Z-p92528`, stateful authorization/recovery and owner-specific acknowledged-write tests; only reads recover accepted writes. |
| A014 Editing | PASS | Explicit submission and retained drafts in WS-03; exact native clipboard and post-review edit pass `20260921T071112Z-p35368`; source-null/empty/absent preservation unit `20260921T071111Z-p35060` and copy unit `20260921T070154Z-p91918`. |
| A015 Conflict | PASS | Cell/field-local saved-versus-draft cues, retained recovery and same-field conflict browser/stateful/visual cases; toast is not sole unresolved state. |
| A016 Query and interaction | PASS | Typed presentation, independent sibling reads and owner matrices; full stateful and a11y coverage keep query state separate from permission and lifecycle. |
| A017 Authorization scope | PASS | Same-account recovery, account replacement, incident concealment/access loss and late-response tests in WS-03–09 plus full stateful run; no generic clear-all or stale protected content. |
| A018 Evidence | PASS | File lifecycle, preview/download independence, count scope, uncertain/saved-unlinked/accepted-unrefreshed cases; full a11y and Evidence browser resolution `20260921T063317Z-p24970`. |
| A019 Accessibility | PASS | All final accessibility scenarios pass through the full run plus focused Coordination resolution `20260921T074150Z-p32164`; keyboard scenarios, nested Escape/focus return, density/zoom/spacing/reduced-motion and visible non-color state fixtures; implementation-support posture retained. |
| A020 Components | PASS | Shared role/state unit fixtures, complete a11y and reviewed long-content/zoom/density captures; expansion retains mounted work and recovery controls. |
| A021 Virtualization | PASS | Production Grid Adapter continuity rows in fast/stateful/full browser runs, paging and source identity checks; no virtualization/query-size algorithm change. Unchanged performance certification is outside this slice; no new throughput/latency claim. |
| A022 Visual fixtures | PASS | Exact 255-capture reconciliation, 151 manual image reviews and complete re-review of later changes. Final public update `20260921T074355Z-p30667` passes; two fresh ordinary runs `20260921T074928Z-p69256` and `20260921T074928Z-p69271` each pass 12/12 against the promoted manifest. |
| A023 Selectors | PASS | Semantic field/record/view selectors and authored fixture identities; public native disclosure controls replace removed presentations; package.ui verification and generation. |
| A024 Test authority | PASS | Runtime/test/generator dependency review and import/boundary checks; machine facts remain in authored contracts; Markdown is human authority/support only. |
| A025 Generation | PASS | Authored design/OpenAPI/schema inputs precede generated facades; generation/drift/policy/JSON-shape checks, immutable compatibility baseline retained. |
| A026 Compatibility | PASS | Pending OpenAPI 2.0.0 reviewed break, canonical retained/imported data tests, generation-before-equality, baseline/candidate paired bundle builds and rollback guidance; no migration, adapter, backfill or automatic reset. |
| A027 Handoff | PASS | Serial DONE checkpoints WS-00–09 preserved; mandatory final WS-10 exit saved after checks. Complete changed-file inventory, exact evidence/failure resolutions, compatibility, retirement, verified candidate/baseline bundles, skipped-check reasons and next action; final Markdown/whitespace checks pass. |

## Completion and next maintainer action

WS-00–WS-10 are DONE. All applicable digest rows A001–A027 are PASS, and all
required gap exits have execution evidence. The controlling tracker records each
serial checkpoint and the mandatory final validation/handoff exit. The final
source passes 708 fast units, all selected behavioral scenarios through full and
focused runs, contract/static checks, paired builds, and two ordinary full visual
runs after the reviewed refresh. Failed full runs remain explicitly identified
with their successful scenario resolutions; they are not reported as full passes.

Retained-run maintenance was SKIPPED because `RESULTS_DIR` was unset; no qualifying
full warm check was supplied. Broad `make check`, release certification and a new
performance/measurement campaign were not run: the selected owner suites cover the
changed contracts/presentation, and this work changes no virtualization, query-size
or throughput algorithm. No external deployment or participant usability study was
performed. The five-task study remains the explicit next usability investigation.

The next maintainer should review and stage the coherent source/spec/contract/test
change, preserving the user's existing staged audit. Promote the paired candidate
only through the rollout sequence above and retain the matching baseline bundle.
Do not split the semantic History server/browser cutover, add a legacy decoder,
force a work-losing reload, or reset stored data. The recorded intermittent browser
timing failures merit harness follow-up; their unchanged isolated reruns passed.

## Changed-file inventory

This is the final authored/generated file inventory for human review. The 151
changed or new PNGs are individually mapped in the visual table above. Deleted
Indicator style paths are intentionally retained in this inventory as retirements.
Ignored test outputs, dependency installs and local bundle archives are excluded.

### Adopted owners and source guidance

- `apps/web/src/workbook/features/evidence/README.md`
- `apps/web/src/workbook/features/indicators/README.md`
- `apps/web/src/workbook/inspector/presentation/README.md`
- `apps/web/src/workbook/timeline/README.md`
- `docs/design.md`
- `docs/handoffs/workbook-inspector-ui-ux-audit.md`
- `docs/handoffs/workbook-inspector-validation-handoff.md`
- `docs/spec/01_architecture_storage_and_view_contracts.md`
- `docs/spec/02_domain_model_schema_and_history.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `internal/modules/revisions/README.md`

### Frontend production

- `apps/web/src/workbook/adapters/workbookHistoryResponse.ts`
- `apps/web/src/workbook/components/EntityWorkbookSurface.tsx`
- `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`
- `apps/web/src/workbook/components/workbookFormStyles.ts`
- `apps/web/src/workbook/features/assessments/AssessmentWorkbookInspector.tsx`
- `apps/web/src/workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`
- `apps/web/src/workbook/features/coordination/ContextualCreateForm.tsx`
- `apps/web/src/workbook/features/coordination/CoordinationCreateForm.tsx`
- `apps/web/src/workbook/features/coordination/CoordinationWorkflowBindings.tsx`
- `apps/web/src/workbook/features/coordination/DecisionSupersessionEditor.tsx`
- `apps/web/src/workbook/features/entities/EntityWorkbookInspector.tsx`
- `apps/web/src/workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`
- `apps/web/src/workbook/features/evidence/EvidenceAccessActions.tsx`
- `apps/web/src/workbook/features/evidence/EvidenceAttachmentEntry.tsx`
- `apps/web/src/workbook/features/evidence/EvidenceFileRecovery.tsx`
- `apps/web/src/workbook/features/evidence/TimelineRelatedEvidenceForm.tsx`
- `apps/web/src/workbook/features/evidence/useEvidenceWorkbookBindings.tsx`
- `apps/web/src/workbook/features/generic/GenericWorkbookInspector.tsx`
- `apps/web/src/workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`
- `apps/web/src/workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`
- `apps/web/src/workbook/features/indicators/IndicatorCanonicalAuthoring.tsx`
- `apps/web/src/workbook/features/indicators/IndicatorCreateFromObservation.tsx`
- `apps/web/src/workbook/features/indicators/IndicatorCreateOperationStatus.tsx`
- `apps/web/src/workbook/features/indicators/IndicatorInspectorWorkflow.tsx`
- `apps/web/src/workbook/features/indicators/IndicatorLifecycleOperationStatus.tsx`
- `apps/web/src/workbook/features/indicators/IndicatorLifecycleSupportPicker.tsx`
- `apps/web/src/workbook/features/indicators/IndicatorLifecycleWorkflow.tsx`
- `apps/web/src/workbook/features/indicators/ObservationCaptureEditor.tsx`
- `apps/web/src/workbook/features/indicators/ObservationDetails.tsx`
- `apps/web/src/workbook/features/indicators/ObservationOperationStatus.tsx`
- `apps/web/src/workbook/features/indicators/ObservationTargetPicker.tsx`
- `apps/web/src/workbook/features/indicators/WorkbookIndicatorCreateRecovery.tsx`
- `apps/web/src/workbook/features/indicators/WorkbookObservationRecovery.tsx`
- `apps/web/src/workbook/features/indicators/indicatorLifecycleStyles.ts`
- `apps/web/src/workbook/features/indicators/observationStyles.ts`
- `apps/web/src/workbook/features/notes/NoteCreateForm.tsx`
- `apps/web/src/workbook/features/notes/NoteSheetAuthoring.tsx`
- `apps/web/src/workbook/features/notes/NoteSourceControl.tsx`
- `apps/web/src/workbook/history/HistoryPageLookup.ts`
- `apps/web/src/workbook/history/workbookHistoryBrowsing.ts`
- `apps/web/src/workbook/history/workbookHistoryItem.ts`
- `apps/web/src/workbook/history/workbookHistoryTestFixtures.ts`
- `apps/web/src/workbook/inspector/InspectorCreateRelatedWorkflow.tsx`
- `apps/web/src/workbook/inspector/WorkbookInspectorContextualActions.tsx`
- `apps/web/src/workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx`
- `apps/web/src/workbook/inspector/WorkbookInspectorDetails.tsx`
- `apps/web/src/workbook/inspector/WorkbookInspectorDraftStore.ts`
- `apps/web/src/workbook/inspector/WorkbookInspectorRecordHistory.tsx`
- `apps/web/src/workbook/inspector/WorkbookRecordHistoryPresentation.tsx`
- `apps/web/src/workbook/inspector/presentation/WorkbookHistoryPresentation.tsx`
- `apps/web/src/workbook/inspector/presentation/WorkbookInspectorActions.tsx`
- `apps/web/src/workbook/inspector/presentation/WorkbookInspectorFeedback.tsx`
- `apps/web/src/workbook/inspector/presentation/WorkbookInspectorShell.tsx`
- `apps/web/src/workbook/inspector/presentation/workbookInspectorPresentationModel.ts`
- `apps/web/src/workbook/inspector/useRetainedInspectorRow.ts`
- `apps/web/src/workbook/inspector/useWorkbookInspectorEditDraft.ts`
- `apps/web/src/workbook/inspector/useWorkbookRecordHistoryController.ts`
- `apps/web/src/workbook/inspector/useWorkbookRecordHistoryFocus.ts`
- `apps/web/src/workbook/inspector/workbookHistoryPresentationModel.ts`
- `apps/web/src/workbook/inspector/workbookRecordHistoryModel.ts`
- `apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx`
- `apps/web/src/workbook/layout/workbookInspectorNavigation.ts`
- `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.tsx`
- `apps/web/src/workbook/timeline/components/TimelineInspectorDetails.tsx`
- `apps/web/src/workbook/timeline/components/TimelineMentionActionControls.tsx`
- `apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookInspectorSections.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookStyles.ts`
- `apps/web/src/workbook/timeline/composition/useTimelineInspectorStateComposition.ts`
- `apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineInspectorSelection.ts`
- `apps/web/src/workbook/timeline/models/timelineCommittedVersionLedger.ts`
- `apps/web/src/workbook/timeline/models/timelineDiscardedReconciliation.ts`
- `apps/web/src/workbook/timeline/models/timelineRowModel.ts`

### Frontend unit and browser validation

- `apps/web/e2e/batch-history-review.spec.ts`
- `apps/web/e2e/collaboration.spec.ts`
- `apps/web/e2e/entity-find.spec.ts`
- `apps/web/e2e/evidence.spec.ts`
- `apps/web/e2e/history-browsing.spec.ts`
- `apps/web/e2e/history-recovery.spec.ts`
- `apps/web/e2e/history.spec.ts`
- `apps/web/e2e/indicator-lifecycle.spec.ts`
- `apps/web/e2e/indicator-observations.spec.ts`
- `apps/web/e2e/inspector-actions.spec.ts`
- `apps/web/e2e/keyboard.spec.ts`
- `apps/web/e2e/mention-lifecycle.spec.ts`
- `apps/web/e2e/mentions.recovery.spec.ts`
- `apps/web/e2e/merge-recovery.spec.ts`
- `apps/web/e2e/sentinel.spec.ts`
- `apps/web/e2e/support/workbook/history.ts`
- `apps/web/e2e/support/workbook/rowMutations.ts`
- `apps/web/e2e/timeline-related-evidence.spec.ts`
- `apps/web/e2e/timeline-scalar-clipboard.spec.ts`
- `apps/web/e2e/timeline-workbook.spec.ts`
- `apps/web/e2e/workbook-inspector-edit.spec.ts`
- `apps/web/e2e/workbook-ordinary-create.spec.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.generic.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/src/workbook/WorkbookShell.actionSequencing.test.tsx`
- `apps/web/src/workbook/WorkbookShell.assessments.test.tsx`
- `apps/web/src/workbook/WorkbookShell.autosave.test.tsx`
- `apps/web/src/workbook/WorkbookShell.history.test.tsx`
- `apps/web/src/workbook/WorkbookShell.inspector.test.tsx`
- `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`
- `apps/web/src/workbook/adapters/createWorkbookRecordHistoryAdapter.test.ts`
- `apps/web/src/workbook/features/coordination/decisionSupersessionReconciliation.test.tsx`
- `apps/web/src/workbook/features/entities/entityInspectorEditing.test.tsx`
- `apps/web/src/workbook/features/evidence/useEvidenceWorkbookBindings.test.tsx`
- `apps/web/src/workbook/features/indicators/IndicatorInspectorWorkflow.test.tsx`
- `apps/web/src/workbook/features/indicators/indicatorCreateReconciliation.test.ts`
- `apps/web/src/workbook/features/indicators/indicatorLifecycleReconciliation.test.tsx`
- `apps/web/src/workbook/features/indicators/observationReconciliation.test.tsx`
- `apps/web/src/workbook/history/HistoryActionLookup.test.ts`
- `apps/web/src/workbook/history/WorkbookBatchHistoryReview.test.tsx`
- `apps/web/src/workbook/history/WorkbookHistoryRecovery.test.tsx`
- `apps/web/src/workbook/history/WorkbookRecordHistoryOwner.test.ts`
- `apps/web/src/workbook/history/workbookHistoryBrowsing.characterization.test.tsx`
- `apps/web/src/workbook/history/workbookHistoryBrowsing.test.ts`
- `apps/web/src/workbook/inspector/WorkbookInspectorRecordHistory.test.tsx`
- `apps/web/src/workbook/inspector/presentation/WorkbookInspectorPresentation.test.tsx`
- `apps/web/src/workbook/inspector/useRetainedInspectorRow.test.ts`
- `apps/web/src/workbook/inspector/useWorkbookInspectorEditDraft.test.tsx`
- `apps/web/src/workbook/inspector/workbookHistoryPresentationModel.test.ts`
- `apps/web/src/workbook/inspector/workbookHistoryRecovery.characterization.test.tsx`
- `apps/web/src/workbook/inspector/workbookRecordHistoryModel.test.ts`
- `apps/web/src/workbook/timeline/actions/timelineMentionCandidates.test.tsx`
- `apps/web/src/workbook/timeline/adapters/createTimelineActionAdapters.test.ts`
- `apps/web/src/workbook/timeline/components/TimelineCollectionCell.test.tsx`
- `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.test.tsx`
- `apps/web/src/workbook/timeline/components/TimelineInspectorDetails.test.tsx`
- `apps/web/src/workbook/timeline/models/timelineMutationModels.test.ts`
- `apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.test.tsx`
- `apps/web/src/workbook/timeline/useTimelineCompositionLifecycle.test.tsx`

### Source-owner implementation and backend validation

- `internal/app/revisionassembly/history_projection_test.go`
- `internal/modules/artifacts/history_projection.go`
- `internal/modules/artifacts/revision_provider_contribution.go`
- `internal/modules/assessments/history_projection.go`
- `internal/modules/assessments/revision_provider_contribution.go`
- `internal/modules/entities/boundary_guard_test.go`
- `internal/modules/entities/history_projection.go`
- `internal/modules/entities/revision_provider_contribution.go`
- `internal/modules/evidence/history_projection.go`
- `internal/modules/evidence/revision_provider_contribution.go`
- `internal/modules/incidentbundles/routes_helpers_integration_test.go`
- `internal/modules/indicators/boundary_guard_test.go`
- `internal/modules/indicators/history_projection.go`
- `internal/modules/indicators/revision_provider_contribution.go`
- `internal/modules/links/history_projection.go`
- `internal/modules/links/revision_provider_contribution.go`
- `internal/modules/links/test_helpers_test.go`
- `internal/modules/parties/history_projection.go`
- `internal/modules/parties/revision_provider_contribution.go`
- `internal/modules/revisions/catalog_admission_test.go`
- `internal/modules/revisions/command_service_test.go`
- `internal/modules/revisions/history_components_test.go`
- `internal/modules/revisions/history_helpers_test.go`
- `internal/modules/revisions/history_materializer.go`
- `internal/modules/revisions/history_model.go`
- `internal/modules/revisions/history_projection.go`
- `internal/modules/revisions/history_service.go`
- `internal/modules/revisions/history_test.go`
- `internal/modules/revisions/historycontract/history.go`
- `internal/modules/revisions/httpapi/routes.go`
- `internal/modules/revisions/integration_test.go`
- `internal/modules/revisions/provider_contributions.go`
- `internal/modules/revisions/rollback_test.go`
- `internal/modules/revisions/target_semantics_catalog.go`
- `internal/modules/revisions/target_semantics_compiler.go`
- `internal/modules/revisions/target_semantics_nonrow_catalog_test.go`
- `internal/modules/revisions/target_semantics_row_catalog_test.go`
- `internal/modules/tasksdecisions/history_projection.go`
- `internal/modules/tasksdecisions/revision_provider_contribution.go`
- `internal/modules/timeline/history_projection.go`
- `internal/modules/timeline/revision_provider_contribution.go`
- `internal/modules/workbook/note_associations_integration_test.go`

### Authored contracts and routing

- `contracts/design/presentation.v1.json`
- `contracts/openapi-releases/2.0.0.change-set.json`
- `contracts/openapi-source/owners/module.revisions/openapi.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_fixture_registry.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/schemas/cartulary.design_presentation.v1.schema.json`
- `tools/test_families/module.revisions.json`
- `tools/test_families/module.timeline.json`
- `tools/test_families/web.workbook.json`

### Shared packages and generated projections

- `contracts/openapi/cartulary.openapi.yaml`
- `internal/gen/contractopenapi/artifacts_gen.go`
- `internal/gen/openapioperations/catalog_gen.go`
- `packages/protocol-ts/src/generated/core-http-types.ts`
- `packages/protocol-ts/src/generated/core-http-validators.ts`
- `packages/ui-contracts/src/design-presentation.test.ts`
- `packages/ui-contracts/src/generated/design-presentation.ts`
- `tools/harness/generated-artifacts/design-presentation/design-presentation.mjs`
