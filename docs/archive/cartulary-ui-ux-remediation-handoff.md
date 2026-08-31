# Cartulary UI/UX Remediation Handoff

**Date:** 2026-08-30 EDT  
**Baseline commit:** `19343318c1fd2e637b42e2a60bf5d12b36cf03aa`  
**Scope:** Create discovery/readiness, inspector dispatch, density geometry,
responsive metrics, and their verification projections.

## Authority and scope record

The implementation followed `AGENTS.md`, then the order in
`docs/cartulary-ui-ux-refactor-digest/cartulary/START_HERE.md`. The digest is an
immutable advisory snapshot: no file below
`docs/cartulary-ui-ux-refactor-digest/` changed. Its package manifest and sample
offline query were already green at authorization. Narrow offline UX queries
were reviewed as questions only. Keyboard/focus, non-color state, local error,
loading/recovery, stable virtualization, and reserved-space advice was adopted;
generic responsive/density guidance was adapted to Cartulary owners and tokens;
generic mobile/card conversion, Tailwind snippets, generated palettes, blanket
loader timing, and toast-only critical state were rejected. No upstream
`--design-system`, `--persist`, or `--force` option was used.

Repository revalidation found the digest's expected `packages/ui` absent,
`packages/view-contracts/src/generated` active as a managed generated root,
visual support active, and current owner/schema projections newer than the
localized snapshot. `REPO_MAP.tsv` owner and acceptance rows were checked against
the current sources before mutation. The initial tracked worktree was clean.
Product authorization was the request to implement all numbered plan slices;
`make task-guide ROLE=module-author OWNER=<owner>` was run for each affected
owner before its slice.

## Owner-to-change map

| Workstream | Authoritative owner and projections | Implementation and evidence |
| --- | --- | --- |
| Public create discovery | Core 01; authored `platform.viewschema` OpenAPI; view-schema contracts | Additive `create_capable`, `inline_create`, and total `create_writable`; validation of the true-only create override; runtime registry, generated Go/TypeScript, release change set, and public-shape tests |
| Contract-driven creation | Adopted source-owner create rules; view-schema projections | Ordinary field-set minima projected for twelve additional surfaces; all create entry points use capability, interaction authority, and declared create inputs; payloads use only create-writable fields/inputs; Evidence and Indicators retain owner-specific validation |
| Inspector dispatch | Existing `inspector_config_v1` feature groups and route bindings | Stable semantic dispatcher keyed by view schema, group, route kind/owner, and action; completeness, role/state, confirmation, invalidation, history, related-create, record action/patch, Indicator, Evidence, Entity, Assessment, Generic, and Timeline coverage |
| Density geometry | `package.ui`, `package.grid_adapter`, and `web.design` tokens/owners | One density box controls row/header height, block/inline padding, typography, gutters, saved/draft/read-only content, and full-cell editor geometry; Timeline overrides removed |
| Responsive geometry | Existing UI layout tokens and workbook owner algorithms | Validated CSS-length accessors, exact token-backed thresholds and inspector clamp/ARIA, and real `innerWidth`/`innerHeight` fallback when `visualViewport` is absent |
| Verification | Current owner catalog, test families, visual fixture registry | Negative/stale expectations replaced; contract completeness, create minima, viewport fallback, token consumption, computed geometry, browser, accessibility, service-backed, and visual evidence added or refreshed |

No database, stored-data, route, ACL, generic workflow endpoint, theme, mobile/card
conversion, or compatibility shim was introduced. The public discovery change is
response-additive and retains `view_schema_resource_v2`; existing clients must
continue ignoring unknown additive members. Surface pivots remain visibly
disabled when the declared `pivot_target_unavailable` state is true; the client
does not invent a query field or route to bypass owner availability. Transaction
recovery, editing, paste state, draft retention, conflict resolution,
virtualization, and continuity behavior were retained as non-regression gates.

## Visual refresh record

The accepted trigger was the intentional density contract correction plus
complete rendering of current inspector feature groups. Header and cell metrics
now follow the selected density; inspector snapshots expose declared actions.
Viewport, browser zoom, masks, scroll normalization, screenshot scope, browser
pin, and font bundle did not change. Review found only the intended row/header
geometry, text spacing, additional compact visible-row capacity, and inspector
controls; no unexpected focus, clipping, overflow, loading, error, responsive,
or typography change was accepted.

The ordinary validation reconciliation reports 60 capture intents, 60 committed
goldens, 60 active mappings, no orphan or missing golden, no ambiguous mapping,
22 registered fixtures, and no unresolved registered fixture. Its artifact is
`20260830T212653Z-p3214414/browser-e2e-visual/frontend-visual-reconciliation.json`.

Affected semantic owner rows:

- `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1`
- `module.collaboration.visual.the_visual_harness_asserts_deterministic_timelin_22b64f5dec`
- `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c`
- `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc`
- `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`
- `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7`
- `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`
- `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830`
- `module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc`
- `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af`
- `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`
- `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc`
- `module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810`
- `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206`
- `module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf`
- `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0`
- `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`
- `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67`
- `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc`

Affected registered fixture IDs:

- `visual.fixture.base_inspector`
- `visual.fixture.claimed_network_analysis_workspace_states`
- `visual.fixture.compact_desktop_workbook_shell`
- `visual.fixture.default_timeline_workbook_shell`
- `visual.fixture.edit_cell`
- `visual.fixture.empty_successful_query`
- `visual.fixture.evidence_affordance`
- `visual.fixture.mention_chip_state_matrix`
- `visual.fixture.narrow_desktop_workbook_shell`
- `visual.fixture.presence_overflow`
- `visual.fixture.same_field_conflict`
- `visual.fixture.saved_view_query_controls_and_grouped_result`
- `visual.fixture.task_requests_or_decisions`

Changed golden basenames:

- `collaboration-conflict-resolver-linux.png`
- `collaboration-grid-blocked-conflict-linux.png`
- `collaboration-grid-conflict-resolver-linux.png`
- `collaboration-grid-presence-markers-linux.png`
- `collaboration-presence-markers-linux.png`
- `entity-mention-chip-states-linux.png`
- `evidence-affordance-states-linux.png`
- `evidence-grid-available-evidence-linux.png`
- `evidence-grid-blocked-preview-linux.png`
- `evidence-grid-requested-evidence-linux.png`
- `evidence-grid-timeline-evidence-badge-linux.png`
- `evidence-timeline-evidence-count-linux.png`
- `incident-directory-compact-desktop-workbook-shell-linux.png`
- `incident-directory-default-timeline-workbook-shell-linux.png`
- `incident-directory-narrow-desktop-workbook-shell-linux.png`
- `network-flow-analysis-accepted-inspector-linux.png`
- `network-flow-analysis-graph-contributors-linux.png`
- `network-flow-analysis-rejected-diagnostics-linux.png`
- `record-relationships-evidence-access-linux.png`
- `record-relationships-mention-chips-linux.png`
- `record-relationships-task-requests-linux.png`
- `timeline-grid-active-edit-cell-linux.png`
- `timeline-grid-grouped-grid-linux.png`
- `timeline-grid-timeline-default-linux.png`
- `timeline-mutation-active-edit-cell-linux.png`
- `timeline-mutation-empty-timeline-query-linux.png`
- `timeline-mutation-pending-replay-status-linux.png`
- `timeline-mutation-transaction-recovery-panel-linux.png`
- `workbook-inspector-history-linux.png`
- `workbook-inspector-public-error-linux.png`
- `workbook-inspector-relationships-linux.png`
- `workbook-inspector-rollback-preview-linux.png`
- `workbook-query-empty-successful-query-linux.png`
- `workbook-query-saved-view-query-controls-linux.png`

Refresh passed 12/12 at `.cartulary/test-results/20260830T212452Z-p3169561`.
Independent validation passed 12/12 at
`.cartulary/test-results/20260830T212653Z-p3214414`.

## Acceptance disposition

| Rows | Status | Evidence |
| --- | --- | --- |
| A001-A003 | PASS | Owner map above; immutable digest; current repository/generated/owner drift and clean baseline recorded; workstreams remained bounded with no unrelated refactor |
| A004-A005 | PASS | Existing UI token registry/accessors only; no second registry, theme, or generated palette; `package.ui` 10/10 and visual review |
| A006 | PASS | Compact/default/comfortable computed geometry covers row/header, padding, font, line height, gutters, values, anchors, and editors; grid owner 38/38 and service-backed 13/13 |
| A007 | PASS | Capability/auth/create-input gating, exact Hosts/Identities minima, alias/default rejection, paste parity, false-capability fixtures, and browser entry-path tests |
| A008-A009 | PASS | Token boundary and missing-`visualViewport` tests; 1280x720, 1024x720, 768x640, below-minimum, vertical-only, overflow, navigation, and inspector clamp fixtures |
| A010-A011 | PASS | Current-contract semantic-dispatch completeness; role/state-disabled rendering, stable confirmation identity/version, selection/invalidation, history, surface transition, refresh, scroll, row, and focus tests |
| A012-A018 | PASS | Preserved transaction randomness/replay, blocked retry/discard, editing/paste/Escape, conflict loci, async/refresh matrices, and Evidence states; stateful 34/34 and webserver-backed 60/60 |
| A019-A023 | PASS | Accessibility 12/12, measurement 22/22, deterministic component/virtualization fixtures, visual 12/12, stable selector contracts, and UI-contract 10/10 |
| A024 | PASS | Test audit and ownership checks; product tests consume projections/runtime interfaces and do not read, stat, or hash the digest or Markdown |
| A025 | PASS | Generator-owned outputs only; `make generate`, generation drift 4/4, and artifact policy 3/3 |
| A026 | PASS | Additive public projection was adopted in Core 01; no unowned route/schema/write/auth/storage behavior; unavailable pivots fail closed through the declared token |
| A027 | PASS | This record captures owners, paths, behavior, compatibility, commands, visual review, diagnostic failures, and follow-up posture |

## Retained validation evidence

| Command | Result and run root |
| --- | --- |
| `make generate` | PASS; `.cartulary/test-results/20260830T213534Z-p3317210` |
| `make generate-drift` | 4/4; `.cartulary/test-results/20260830T213552Z-p3320407` |
| `make generated-artifact-policy-check` | 3/3; `.cartulary/test-results/20260830T213552Z-p3320431` |
| `make test-slice OWNER=platform.viewschema` | 1/1; `.cartulary/test-results/20260830T213552Z-p3320604` |
| `make test-slice OWNER=package.view_contracts` | 5/5; `.cartulary/test-results/20260830T213552Z-p3320660` |
| `make test-slice OWNER=package.ui` | 10/10; `.cartulary/test-results/20260830T213552Z-p3320706` |
| `make test-slice OWNER=package.grid_adapter` | 38/38; `.cartulary/test-results/20260830T213552Z-p3320733` |
| `make test-slice OWNER=web.workbook` | 138/138; `.cartulary/test-results/20260830T213552Z-p3320780` |
| `make test-slice OWNER=web.design` | 15/15; `.cartulary/test-results/20260830T213552Z-p3320824` |
| `make service-backed-test-slice OWNER=module.workbook` | 39/39; `.cartulary/test-results/20260830T213730Z-p3428815` |
| `make service-backed-test-slice OWNER=package.grid_adapter` | 13/13; `.cartulary/test-results/20260830T212026Z-p3033822` |
| `make service-backed-test-slice OWNER=web.design` | 15/15; `.cartulary/test-results/20260830T212026Z-p3033821` |
| `make test-fast` | 441/441; `.cartulary/test-results/20260830T213731Z-p3428952` |
| `make browser-e2e-webserver-backed` | 60/60; `.cartulary/test-results/20260830T212838Z-p3259809` |
| `make browser-e2e-stateful` | 34/34; `.cartulary/test-results/20260830T214004Z-p3500360` |
| `make browser-e2e-measurement` | 22/22; `.cartulary/test-results/20260830T214004Z-p3500377` |
| `make browser-e2e-a11y` | 12/12; `.cartulary/test-results/20260830T214004Z-p3500513` |
| `make browser-e2e-visual` | 12/12; `.cartulary/test-results/20260830T212653Z-p3214414` |

Earlier webserver-backed roots `20260830T204821Z-p2459925` and
`20260830T211114Z-p2878408` retained the density-header assertion failures that
led to the explicit logical-edge fix. Visual diagnostic root
`20260830T212136Z-p3123262` retained the expected pre-refresh differences. The
corrected owner rows and the complete canonical targets pass above; these failed
roots are diagnostic only and were not used as acceptance evidence.

Final static gates also pass: `make agent-finalize` 1/1 at
`.cartulary/test-results/20260830T214636Z-p3649317`, `make lint-markdown` at
`.cartulary/test-results/20260830T214654Z-p3652447`, frontend typecheck 2/2 at
`.cartulary/test-results/20260830T214654Z-p3652407`, and frontend import-boundary
check 2/2 at `.cartulary/test-results/20260830T214654Z-p3652436`. `RESULTS_DIR`
was intentionally unset because no successful full warm `make check` root from
this exact source was retained; retained-run maintenance was therefore skipped.

## Handoff and follow-up posture

There is no data migration or rollback choreography. Reverting requires keeping
Core 01, authored projections, generated artifacts, runtime registry, client
facades, UI behavior, tests, and reviewed goldens coherent; do not retain a
partial public capability shape. Existing additive-response compatibility and
server authorization remain authoritative.

No remediation slice is deferred. Future feature groups must either use a
registered current route owner or be unknown additive groups governed by
`unsupported_feature_behavior=omit_feature`; current-profile omissions fail the
dispatcher completeness test. Future surface pivots require an owner-declared,
executable view-query seed before their availability token may be cleared. The
next product slice, if any, must obtain its own owner-specific authorization and
task guide.
