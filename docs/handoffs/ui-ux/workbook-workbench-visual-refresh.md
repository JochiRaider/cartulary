# Workbook workbench visual refresh record

Status: all changed images reviewed; two final ordinary comparisons passed. See the
[implementation handoff](workbook-workbench-implementation.md).
This is implementation-support evidence, not behavioral or release authority.

## Accepted change and evidence

The adopted Design 0.6 and Core 03 workbench requirements intentionally change
the incident/view/footer bands, auxiliary destination, inspector reading surface,
and command/query menus. Existing administration and directory consumers of the
shared top-bar minimum-height token also change: the reviewed audit header/content
origin moves from 64 to 58 pixels without losing controls or data. The corrective
refresh keeps the closed lifecycle label within the 40-pixel incident band at
390 pixels. No renderer, font, viewport, zoom, masking, scroll normalization,
screenshot scope, or comparison tolerance was changed for this refresh.

The initial ordinary baseline is
`.cartulary/test-results/20261007T035002Z-p13696`. Its reconciliation reported
no missing or ambiguously mapped goldens; expected pixel comparisons failed.
The initial transactional update passed 12/12 units at
`.cartulary/test-results/20261007T044457Z-p76167`. The corrective transactional
update passed 12/12 units at
`.cartulary/test-results/20261007T050710Z-p2162`.

The final reconciliation accounts for 255 active captures/goldens and 29 registered
fixtures with zero missing goldens, ambiguous mappings, or orphans. Of these,
232 goldens changed relative to the starting commit. Each final changed image was
visually consumed through a private artifact session after byte-length and SHA-256
verification. Review included the single corrected lifecycle image after the
second update. No unexplained visual difference remains. Standalone image imports
provide pixels; they do not provide DOM, accessibility, or performance evidence.
All review sessions were stopped with successful cleanup receipts.

The final `tools/frontend_visual_golden_manifest.json` SHA-256 is
`233a2d274e7cb3abc16035d761f401c404c0292ad1f0415a688418f96133b854`.
Renderer identity remains
`visual.renderer.playwright_1_63_0_chromium_1243_linux_amd64`.
The two final ordinary comparisons both passed 12/12 units against this exact
manifest at `.cartulary/test-results/20261007T051745Z-p41032` and
`.cartulary/test-results/20261007T051745Z-p41039`. Their reconciliations account for
the same 255 active goldens with no missing, ambiguous, or orphan entries. The
implementation handoff records broader functional/accessibility evidence.

## Exact changed-golden ownership

The following mappings were read from the final reconciliation's exact
`golden_path`, `catalog_row_ids`, `owner_ids`, and `fixture_ids` fields, after
matching the changed paths and SHA-256 values. Ownership was not inferred from
filenames. Every filename below is relative to
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`. “Active nonregistry capture”
means an admitted capture without a registry fixture, not an orphan.

### `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1`

Owner: `module.collaboration`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `collaboration-conflict-resolver-linux.png` | `visual.fixture.same_field_conflict` |
| `collaboration-conflict-strip-linux.png` | `visual.fixture.save_state_strip` |
| `collaboration-presence-markers-linux.png` | `visual.fixture.presence_overflow` |
| `collaboration-recovered-saved-strip-linux.png` | `visual.fixture.save_state_strip` |

### `module.collaboration.visual.the_visual_harness_asserts_deterministic_timelin_22b64f5dec`

Owner: `module.collaboration`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `collaboration-grid-presence-markers-linux.png` | Active nonregistry capture |

### `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c`

Owner: `module.collaboration`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `collaboration-grid-conflict-resolver-compact-linux.png` | Active nonregistry capture |
| `collaboration-grid-conflict-resolver-linux.png` | Active nonregistry capture |
| `collaboration-grid-conflict-resolver-narrow-linux.png` | Active nonregistry capture |

### `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc`

Owner: `module.collaboration`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `collaboration-grid-blocked-conflict-linux.png` | Active nonregistry capture |
| `collaboration-grid-recovered-saved-strip-linux.png` | Active nonregistry capture |
| `collaboration-grid-saved-strip-linux.png` | Active nonregistry capture |
| `collaboration-grid-syncing-strip-linux.png` | `visual.fixture.save_state_strip` |

### `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`

Owner: `module.entities`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `entity-mention-chip-states-linux.png` | `visual.fixture.mention_chip_state_matrix` |

### `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7`

Owner: `module.entities`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `record-relationships-mention-chips-linux.png` | Active nonregistry capture |

### `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`

Owner: `module.evidence`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `evidence-affordance-states-linux.png` | `visual.fixture.evidence_affordance` |
| `evidence-timeline-evidence-count-linux.png` | Active nonregistry capture |

### `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830`

Owner: `module.evidence`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `evidence-grid-blocked-preview-linux.png` | Active nonregistry capture |
| `evidence-grid-timeline-evidence-badge-linux.png` | Active nonregistry capture |

### `module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc`

Owner: `module.evidence`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `record-relationships-evidence-access-linux.png` | Active nonregistry capture |

### `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af`

Owner: `module.evidence`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `evidence-grid-available-evidence-linux.png` | Active nonregistry capture |
| `evidence-grid-requested-evidence-linux.png` | Active nonregistry capture |

### `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`

Owner: `module.networkflow`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `network-flow-analysis-accepted-inspector-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-compact-saved-graphs-linux.png` | `visual.fixture.claimed_network_analysis_compact_workspace` |
| `network-flow-analysis-delete-dialog-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-filtered-empty-grid-linux.png` | Active nonregistry capture |
| `network-flow-analysis-graph-contributors-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-mapping-dialog-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-narrow-query-controls-linux.png` | `visual.fixture.claimed_network_analysis_narrow_workspace` |
| `network-flow-analysis-rejected-diagnostics-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-saved-graph-result-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |

### `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc`

Owner: `module.savedviews`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `workbook-query-empty-closed-read-only-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-compact-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-density-comfortable-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-density-compact-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-narrow-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-successful-query-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-text-spacing-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-zoom-200-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-filtered-empty-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-saved-view-query-controls-linux.png` | `visual.fixture.saved_view_query_controls_and_grouped_result` |
| `workbook-view-bar-filter-editing-overflow-linux.png` | Active nonregistry capture |
| `workbook-view-bar-long-columns-linux.png` | Active nonregistry capture |
| `workbook-view-bar-maximum-pressure-base-linux.png` | Active nonregistry capture |
| `workbook-view-bar-maximum-pressure-compact-linux.png` | Active nonregistry capture |
| `workbook-view-bar-maximum-pressure-narrow-linux.png` | Active nonregistry capture |
| `workbook-view-bar-ordered-maximum-sort-linux.png` | Active nonregistry capture |
| `workbook-view-bar-saved-view-actions-linux.png` | Active nonregistry capture |
| `workbook-view-bar-saved-view-clean-linux.png` | Active nonregistry capture |
| `workbook-view-bar-saved-view-modified-linux.png` | Active nonregistry capture |
| `workbook-view-bar-text-spacing-linux.png` | Active nonregistry capture |
| `workbook-view-bar-zoom-200-linux.png` | Active nonregistry capture |

### `module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810`

Owner: `module.timeline`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-grid-grouped-grid-linux.png` | Active nonregistry capture |

### `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206`

Owner: `module.timeline`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-grid-timeline-default-linux.png` | Active nonregistry capture |

### `module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf`

Owner: `module.timeline`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-grid-active-edit-cell-linux.png` | `visual.fixture.edit_cell` |
| `timeline-grid-conflict-strip-linux.png` | Active nonregistry capture |
| `timeline-grid-saved-strip-linux.png` | Active nonregistry capture |
| `timeline-grid-syncing-strip-linux.png` | Active nonregistry capture |

### `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `incident-directory-compact-desktop-workbook-shell-linux.png` | `visual.fixture.compact_desktop_workbook_shell` |
| `incident-directory-default-timeline-workbook-shell-linux.png` | `visual.fixture.default_timeline_workbook_shell` |
| `incident-directory-narrow-desktop-workbook-shell-linux.png` | `visual.fixture.narrow_desktop_workbook_shell` |

### `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `workbook-inspector-attached-edit-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-compact-actions-linux.png` | `visual.fixture.inspector_compact_actions` |
| `workbook-inspector-destructive-confirmation-linux.png` | `visual.fixture.destructive_actions` |
| `workbook-inspector-details-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-history-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-public-error-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-relationships-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-retained-draft-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `visual.fixture.base_inspector` |

### `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-mutation-active-edit-cell-linux.png` | `visual.fixture.edit_cell` |
| `timeline-mutation-empty-timeline-query-linux.png` | Active nonregistry capture |
| `timeline-mutation-pending-replay-status-linux.png` | Active nonregistry capture |
| `timeline-mutation-transaction-recovery-panel-compact-linux.png` | Active nonregistry capture |
| `timeline-mutation-transaction-recovery-panel-linux.png` | Active nonregistry capture |
| `timeline-mutation-transaction-recovery-panel-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `record-relationships-task-requests-linux.png` | `visual.fixture.task_requests_or_decisions` |

### `module.workbook.visual.contextual_task_decision_creation`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `contextual-decision-authoring-linux.png` | Active nonregistry capture |
| `contextual-decision-authoring-narrow-linux.png` | Active nonregistry capture |
| `contextual-decision-recovery-linux.png` | Active nonregistry capture |
| `contextual-decision-recovery-narrow-linux.png` | Active nonregistry capture |
| `contextual-decision-references-narrow-linux.png` | Active nonregistry capture |
| `contextual-task-request-authoring-linux.png` | Active nonregistry capture |
| `contextual-task-request-authoring-narrow-linux.png` | Active nonregistry capture |
| `contextual-task-request-recovery-linux.png` | Active nonregistry capture |
| `contextual-task-request-recovery-narrow-linux.png` | Active nonregistry capture |
| `contextual-task-request-references-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.coordination_create_authoring_recovery`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `coordination-comm-log-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-handoff-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-lesson-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-narrow-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-source-narrow-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-status-review-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |

### `module.workbook.visual.decision_supersession_review_recovery`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `decision-supersession-accepted-linux.png` | Active nonregistry capture |
| `decision-supersession-review-linux.png` | Active nonregistry capture |
| `decision-supersession-review-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.indicator_lifecycle_authoring`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `indicator-lifecycle-authoring-linux.png` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-lifecycle-authoring-narrow-linux.png` | `visual.fixture.indicator_lifecycle_authoring` |

### `module.workbook.visual.indicator_observations_authoring`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `indicator-observation-authoring-linux.png` | `visual.fixture.indicator_observations_authoring` |
| `indicator-observation-authoring-narrow-linux.png` | `visual.fixture.indicator_observations_authoring` |

### `module.workbook.visual.note_create_authoring_recovery`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `linked-note-authoring-linux.png` | Active nonregistry capture |
| `linked-note-authoring-narrow-linux.png` | Active nonregistry capture |
| `linked-note-recovery-linux.png` | Active nonregistry capture |
| `linked-note-recovery-narrow-linux.png` | Active nonregistry capture |
| `linked-note-source-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.ordinary_create_authoring_recovery`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `ordinary-closed-retained-narrow-linux.png` | Active nonregistry capture |
| `ordinary-recovery-1280-linux.png` | Active nonregistry capture |
| `ordinary-recovery-390-linux.png` | Active nonregistry capture |
| `ordinary-reference-authoring-linux.png` | Active nonregistry capture |

### `module.workbook.visual.preferences`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `workbook-preferences-comfortable-linux.png` | Active nonregistry capture |
| `workbook-preferences-compact-linux.png` | Active nonregistry capture |
| `workbook-preferences-confirmed-stale-linux.png` | Active nonregistry capture |
| `workbook-preferences-uncertain-linux.png` | Active nonregistry capture |
| `workbook-preferences-uncertain-narrow-linux.png` | Active nonregistry capture |
| `workbook-preferences-unset-linux.png` | Active nonregistry capture |

### `module.workbook.visual.timeline_capture_actions`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-supersession-accepted-linux.png` | Active nonregistry capture |
| `timeline-supersession-authoring-linux.png` | Active nonregistry capture |
| `timeline-supersession-review-linux.png` | Active nonregistry capture |
| `timeline-supersession-review-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.timeline_related_evidence`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-related-evidence-authoring-linux.png` | Active nonregistry capture |
| `timeline-related-evidence-authoring-narrow-linux.png` | Active nonregistry capture |
| `timeline-related-evidence-partial-linux.png` | Active nonregistry capture |
| `timeline-related-evidence-partial-narrow-linux.png` | Active nonregistry capture |
| `timeline-related-evidence-party-narrow-linux.png` | Active nonregistry capture |

### `package.grid_adapter.visual.capture_test_only_grid_adapter_support_specimens_9c222633ba`

Owner: `package.grid_adapter`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-grid-adapter-fixtures-linux.png` | `visual.fixture.drag_fill_handle`, `visual.fixture.edit_cell`, `visual.fixture.frozen_column`, `visual.fixture.resize_handle`, `visual.fixture.tree_group_row` |

### `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `account-menu-controls-compact-linux.png` | Active nonregistry capture |
| `account-menu-controls-narrow-linux.png` | Active nonregistry capture |
| `account-menu-controls-short-linux.png` | Active nonregistry capture |
| `account-menu-deployment-root-linux.png` | Active nonregistry capture |
| `account-menu-directory-root-linux.png` | Active nonregistry capture |
| `account-menu-long-label-text-spacing-linux.png` | Active nonregistry capture |
| `account-menu-long-label-zoom-linux.png` | Active nonregistry capture |
| `account-menu-workbook-root-linux.png` | Active nonregistry capture |

### `web.design.visual.account_settings_editing_states_5af731dc29`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `account-settings-appearance-conflict-linux.png` | Active nonregistry capture |
| `account-settings-appearance-dirty-linux.png` | Active nonregistry capture |
| `account-settings-appearance-pending-linux.png` | Active nonregistry capture |
| `account-settings-appearance-recovery-linux.png` | Active nonregistry capture |
| `account-settings-appearance-short-linux.png` | Active nonregistry capture |
| `account-settings-appearance-zoom-linux.png` | Active nonregistry capture |
| `account-settings-profile-conflict-linux.png` | Active nonregistry capture |
| `account-settings-profile-dirty-linux.png` | Active nonregistry capture |
| `account-settings-profile-long-name-linux.png` | Active nonregistry capture |
| `account-settings-profile-pending-linux.png` | Active nonregistry capture |
| `account-settings-profile-recovery-linux.png` | Active nonregistry capture |

### `web.design.visual.administrative_audit_browsing`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `administrative-audit-empty-linux.png` | Active nonregistry capture |
| `administrative-audit-inspected-linux.png` | Active nonregistry capture |
| `administrative-audit-loading-linux.png` | Active nonregistry capture |
| `administrative-audit-page-two-linux.png` | Active nonregistry capture |
| `administrative-audit-stale-linux.png` | Active nonregistry capture |
| `administrative-audit-unavailable-linux.png` | Active nonregistry capture |

### `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `design-delayed-initial-loading-linux.png` | `visual.fixture.delayed_initial_loading` |
| `design-exposed-theme-states-linux.png` | `visual.fixture.component_state_matrix` |
| `design-grid-background-refresh-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-grid-closed-read-only-rows-linux.png` | Active nonregistry capture |
| `design-grid-stale-refresh-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-grid-unavailable-initial-load-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-immediate-initial-loading-linux.png` | `visual.fixture.delayed_initial_loading` |

### `web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `incident-create-confirmed-handoff-failure-linux.png` | Active nonregistry capture |
| `incident-create-expanded-details-linux.png` | Active nonregistry capture |
| `incident-create-initial-linux.png` | Active nonregistry capture |
| `incident-create-pending-linux.png` | Active nonregistry capture |
| `incident-create-recovery-short-linux.png` | Active nonregistry capture |
| `incident-create-recovery-zoom-linux.png` | Active nonregistry capture |
| `incident-create-required-errors-linux.png` | Active nonregistry capture |

### `web.design.visual.incident_import_workflow`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `incident-import-access-retry-linux.png` | Active nonregistry capture |
| `incident-import-action-checking-linux.png` | Active nonregistry capture |
| `incident-import-action-unconfirmed-linux.png` | Active nonregistry capture |
| `incident-import-admission-pending-linux.png` | Active nonregistry capture |
| `incident-import-cancel-requested-linux.png` | Active nonregistry capture |
| `incident-import-canceled-linux.png` | Active nonregistry capture |
| `incident-import-empty-linux.png` | Active nonregistry capture |
| `incident-import-failed-linux.png` | Active nonregistry capture |
| `incident-import-handoff-unavailable-linux.png` | Active nonregistry capture |
| `incident-import-observation-unavailable-linux.png` | Active nonregistry capture |
| `incident-import-queued-indeterminate-linux.png` | Active nonregistry capture |
| `incident-import-required-linux.png` | Active nonregistry capture |
| `incident-import-running-determinate-linux.png` | Active nonregistry capture |
| `incident-import-succeeded-linux.png` | Active nonregistry capture |

### `web.design.visual.lifecycle`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `lifecycle-closed-linux.png` | Active nonregistry capture |
| `lifecycle-comfortable-linux.png` | Active nonregistry capture |
| `lifecycle-compact-linux.png` | Active nonregistry capture |
| `lifecycle-confirmed-refresh-failure-linux.png` | Active nonregistry capture |
| `lifecycle-pending-linux.png` | Active nonregistry capture |
| `lifecycle-reason-linux.png` | Active nonregistry capture |
| `lifecycle-review-linux.png` | Active nonregistry capture |
| `lifecycle-review-narrow-linux.png` | Active nonregistry capture |
| `lifecycle-review-spacing-linux.png` | Active nonregistry capture |
| `lifecycle-review-zoom-linux.png` | Active nonregistry capture |
| `lifecycle-uncertain-linux.png` | Active nonregistry capture |

### `web.design.visual.membership_audit_browsing`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `membership-audit-comfortable-linux.png` | Active nonregistry capture |
| `membership-audit-compact-linux.png` | Active nonregistry capture |
| `membership-audit-cursor-recovery-linux.png` | Active nonregistry capture |
| `membership-audit-empty-linux.png` | Active nonregistry capture |
| `membership-audit-inspected-linux.png` | Active nonregistry capture |
| `membership-audit-inspected-narrow-linux.png` | Active nonregistry capture |
| `membership-audit-inspected-spacing-linux.png` | Active nonregistry capture |
| `membership-audit-inspected-zoom-linux.png` | Active nonregistry capture |
| `membership-audit-loading-linux.png` | Active nonregistry capture |
| `membership-audit-stale-linux.png` | Active nonregistry capture |

### `web.design.visual.membership_management_visual`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `membership-management-comfortable-linux.png` | Active nonregistry capture |
| `membership-management-compact-linux.png` | Active nonregistry capture |
| `membership-management-confirmed-refresh-failure-linux.png` | Active nonregistry capture |
| `membership-management-loading-linux.png` | Active nonregistry capture |
| `membership-management-pending-linux.png` | Active nonregistry capture |
| `membership-management-removal-narrow-linux.png` | Active nonregistry capture |
| `membership-management-removal-spacing-linux.png` | Active nonregistry capture |
| `membership-management-removal-zoom-linux.png` | Active nonregistry capture |
| `membership-management-role-linux.png` | Active nonregistry capture |
| `membership-management-uncertain-linux.png` | Active nonregistry capture |

### `web.design.visual.metadata_editing`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `metadata-closed-linux.png` | Active nonregistry capture |
| `metadata-comfortable-linux.png` | Active nonregistry capture |
| `metadata-compact-linux.png` | Active nonregistry capture |
| `metadata-confirmed-refresh-failure-linux.png` | Active nonregistry capture |
| `metadata-conflict-linux.png` | Active nonregistry capture |
| `metadata-dirty-linux.png` | Active nonregistry capture |
| `metadata-loading-linux.png` | Active nonregistry capture |
| `metadata-review-narrow-linux.png` | Active nonregistry capture |
| `metadata-review-spacing-linux.png` | Active nonregistry capture |
| `metadata-review-zoom-linux.png` | Active nonregistry capture |
| `metadata-saving-linux.png` | Active nonregistry capture |
| `metadata-uncertain-linux.png` | Active nonregistry capture |

### `web.design.visual.reference_pack_administration`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `reference-pack-cancel-requested-linux.png` | Active nonregistry capture |
| `reference-pack-canceled-linux.png` | Active nonregistry capture |
| `reference-pack-catalog-linux.png` | Active nonregistry capture |
| `reference-pack-failed-linux.png` | Active nonregistry capture |
| `reference-pack-observation-recovery-linux.png` | Active nonregistry capture |
| `reference-pack-queued-linux.png` | Active nonregistry capture |
| `reference-pack-running-linux.png` | Active nonregistry capture |
| `reference-pack-selection-linux.png` | Active nonregistry capture |
| `reference-pack-submitted-linux.png` | Active nonregistry capture |
| `reference-pack-succeeded-linux.png` | Active nonregistry capture |
