# Workbook stages 1–2 golden refresh record

Status: technical validation complete. Agent review supports the authorized design
change; human design acceptance and participant evaluation remain outstanding.

## Accepted trigger

The user authorized Design §§7.1/7.5/8.3/12 and Core 03 §14.9 amendments for
incident/view hierarchy, a conditional query strip, compact grid footer, quiet
commands and collapsed saved-view Startup. This intentionally changes shell
geometry and control presentation throughout retained production fixtures.

The pinned renderer, viewports, zoom, density, masks, scroll normalization and
screenshot scopes are unchanged. New study attachments use explicit top/left
anchors; they are not goldens and do not replace required fixtures. Existing
default-shell focus intentionally scrolls to Activity Synopsis. The inspector
bound assertion now uses the complete primary-grid slot, which includes its
footer, rather than only the shorter grid scrollport.

## Evidence and review

Ordinary reconciliation was reviewed before refresh at
`20260921T220859Z-p27254`, complemented by the three completed default-shell
captures in `20260921T221613Z-p51035`. Neither artifact had missing or ambiguous
golden mappings. No golden was deleted.

The first canonical update passed all 12 units at `20260921T221842Z-p64677`;
reconciliation accounted for 255 captures and 255 active goldens, with zero
orphans, missing goldens or ambiguous mappings. It changed 172 images and the
generated golden manifest. Agent review covered all changed images in 20 indexed
contact sheets, plus full-size shell, pressure and saved-view states. The review
identified remaining native Sort reorder/remove buttons, which were then migrated
to the same quiet treatment. The final Sort image was reviewed at full size; its quiet controls and unboxed
sections retain complete keyboard targets and readable order. The second canonical
update passed at `20260921T222849Z-p38514` with the same complete 255-capture
reconciliation. Two fresh ordinary `make browser-e2e-visual` runs then passed all
12 units at `20260921T223437Z-p88272` and `20260921T223438Z-p88479`.

Observed intentional effects: incident labels use spare width; saved-view labels
receive more width; routine tools lose repeated borders; desktop chips occupy
their own row; browsing moves below the grid; Startup is collapsed. Existing
conflict, retained-draft, read recovery, inspector and extension surfaces retain
their distinct context and actions. Eight Network Analysis images change only
within top chrome; its resource workflow is untouched.

Review sheets and a pixel-bound inventory are retained under
`.cartulary/design-studies/workbook-overhaul/golden-review/`. This is diagnostic
review material; only canonical Playwright PNGs supply golden inputs.

## Complete changed-golden inventory

Paths below are within `apps/web/e2e/workbook.visual.spec.ts-snapshots/`. Each
owner row and fixture identity comes from the reconciled authored catalog; an
em dash means a valid active capture without a registered fixture.

### module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1

| Golden filename | Registered fixture IDs |
| --- | --- |
| `collaboration-conflict-resolver-linux.png` | `visual.fixture.same_field_conflict` |
| `collaboration-presence-markers-linux.png` | `visual.fixture.presence_overflow` |

### module.collaboration.visual.the_visual_harness_asserts_deterministic_timelin_22b64f5dec

| Golden filename | Registered fixture IDs |
| --- | --- |
| `collaboration-grid-presence-markers-linux.png` | — |

### module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c

| Golden filename | Registered fixture IDs |
| --- | --- |
| `collaboration-grid-conflict-resolver-compact-linux.png` | — |
| `collaboration-grid-conflict-resolver-linux.png` | — |
| `collaboration-grid-conflict-resolver-narrow-linux.png` | — |

### module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc

| Golden filename | Registered fixture IDs |
| --- | --- |
| `collaboration-grid-blocked-conflict-linux.png` | — |

### module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7

| Golden filename | Registered fixture IDs |
| --- | --- |
| `entity-mention-chip-states-linux.png` | `visual.fixture.mention_chip_state_matrix` |

### module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7

| Golden filename | Registered fixture IDs |
| --- | --- |
| `record-relationships-mention-chips-linux.png` | — |

### module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4

| Golden filename | Registered fixture IDs |
| --- | --- |
| `evidence-affordance-states-linux.png` | `visual.fixture.evidence_affordance` |
| `evidence-timeline-evidence-count-linux.png` | — |

### module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830

| Golden filename | Registered fixture IDs |
| --- | --- |
| `evidence-grid-blocked-preview-linux.png` | — |
| `evidence-grid-timeline-evidence-badge-linux.png` | — |

### module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc

| Golden filename | Registered fixture IDs |
| --- | --- |
| `record-relationships-evidence-access-linux.png` | — |

### module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af

| Golden filename | Registered fixture IDs |
| --- | --- |
| `evidence-grid-available-evidence-linux.png` | — |
| `evidence-grid-requested-evidence-linux.png` | — |

### module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6

| Golden filename | Registered fixture IDs |
| --- | --- |
| `network-flow-analysis-accepted-inspector-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-compact-saved-graphs-linux.png` | `visual.fixture.claimed_network_analysis_compact_workspace` |
| `network-flow-analysis-delete-dialog-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-graph-contributors-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-mapping-dialog-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-narrow-query-controls-linux.png` | `visual.fixture.claimed_network_analysis_narrow_workspace` |
| `network-flow-analysis-rejected-diagnostics-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-saved-graph-result-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |

### module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc

| Golden filename | Registered fixture IDs |
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
| `workbook-view-bar-filter-editing-overflow-linux.png` | — |
| `workbook-view-bar-long-columns-linux.png` | — |
| `workbook-view-bar-maximum-pressure-base-linux.png` | — |
| `workbook-view-bar-maximum-pressure-compact-linux.png` | — |
| `workbook-view-bar-maximum-pressure-narrow-linux.png` | — |
| `workbook-view-bar-ordered-maximum-sort-linux.png` | — |
| `workbook-view-bar-saved-view-actions-linux.png` | — |
| `workbook-view-bar-saved-view-clean-linux.png` | — |
| `workbook-view-bar-saved-view-modified-linux.png` | — |
| `workbook-view-bar-text-spacing-linux.png` | — |
| `workbook-view-bar-zoom-200-linux.png` | — |

### module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810

| Golden filename | Registered fixture IDs |
| --- | --- |
| `timeline-grid-grouped-grid-linux.png` | — |

### module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206

| Golden filename | Registered fixture IDs |
| --- | --- |
| `timeline-grid-timeline-default-linux.png` | — |

### module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf

| Golden filename | Registered fixture IDs |
| --- | --- |
| `timeline-grid-active-edit-cell-linux.png` | `visual.fixture.edit_cell` |

### module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0

| Golden filename | Registered fixture IDs |
| --- | --- |
| `incident-directory-compact-desktop-workbook-shell-linux.png` | `visual.fixture.compact_desktop_workbook_shell` |
| `incident-directory-default-timeline-workbook-shell-linux.png` | `visual.fixture.default_timeline_workbook_shell` |
| `incident-directory-narrow-desktop-workbook-shell-linux.png` | `visual.fixture.narrow_desktop_workbook_shell` |

### module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea

| Golden filename | Registered fixture IDs |
| --- | --- |
| `workbook-inspector-attached-edit-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-compact-actions-linux.png` | `visual.fixture.inspector_compact_actions` |
| `workbook-inspector-details-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-history-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-public-error-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-relationships-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-retained-draft-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `visual.fixture.base_inspector` |

### module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67

| Golden filename | Registered fixture IDs |
| --- | --- |
| `timeline-mutation-active-edit-cell-linux.png` | `visual.fixture.edit_cell` |
| `timeline-mutation-empty-timeline-query-linux.png` | — |
| `timeline-mutation-pending-replay-status-linux.png` | — |
| `timeline-mutation-transaction-recovery-panel-compact-linux.png` | — |
| `timeline-mutation-transaction-recovery-panel-linux.png` | — |
| `timeline-mutation-transaction-recovery-panel-narrow-linux.png` | — |

### module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc

| Golden filename | Registered fixture IDs |
| --- | --- |
| `record-relationships-task-requests-linux.png` | `visual.fixture.task_requests_or_decisions` |

### module.workbook.visual.contextual_task_decision_creation

| Golden filename | Registered fixture IDs |
| --- | --- |
| `contextual-decision-authoring-linux.png` | — |
| `contextual-decision-authoring-narrow-linux.png` | — |
| `contextual-decision-recovery-linux.png` | — |
| `contextual-decision-recovery-narrow-linux.png` | — |
| `contextual-decision-references-narrow-linux.png` | — |
| `contextual-task-request-authoring-linux.png` | — |
| `contextual-task-request-authoring-narrow-linux.png` | — |
| `contextual-task-request-recovery-linux.png` | — |
| `contextual-task-request-recovery-narrow-linux.png` | — |
| `contextual-task-request-references-narrow-linux.png` | — |

### module.workbook.visual.coordination_create_authoring_recovery

| Golden filename | Registered fixture IDs |
| --- | --- |
| `coordination-comm-log-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-handoff-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-lesson-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-narrow-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-source-narrow-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-status-review-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |

### module.workbook.visual.decision_supersession_review_recovery

| Golden filename | Registered fixture IDs |
| --- | --- |
| `decision-supersession-accepted-linux.png` | — |
| `decision-supersession-review-linux.png` | — |
| `decision-supersession-review-narrow-linux.png` | — |

### module.workbook.visual.indicator_lifecycle_authoring

| Golden filename | Registered fixture IDs |
| --- | --- |
| `indicator-lifecycle-authoring-linux.png` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-lifecycle-authoring-narrow-linux.png` | `visual.fixture.indicator_lifecycle_authoring` |

### module.workbook.visual.indicator_observations_authoring

| Golden filename | Registered fixture IDs |
| --- | --- |
| `indicator-observation-authoring-linux.png` | `visual.fixture.indicator_observations_authoring` |
| `indicator-observation-authoring-narrow-linux.png` | `visual.fixture.indicator_observations_authoring` |

### module.workbook.visual.note_create_authoring_recovery

| Golden filename | Registered fixture IDs |
| --- | --- |
| `linked-note-authoring-linux.png` | — |
| `linked-note-authoring-narrow-linux.png` | — |
| `linked-note-recovery-linux.png` | — |
| `linked-note-recovery-narrow-linux.png` | — |
| `linked-note-source-narrow-linux.png` | — |

### module.workbook.visual.ordinary_create_authoring_recovery

| Golden filename | Registered fixture IDs |
| --- | --- |
| `ordinary-closed-retained-narrow-linux.png` | — |
| `ordinary-recovery-1280-linux.png` | — |
| `ordinary-recovery-390-linux.png` | — |
| `ordinary-reference-authoring-linux.png` | — |

### module.workbook.visual.preferences

| Golden filename | Registered fixture IDs |
| --- | --- |
| `workbook-preferences-comfortable-linux.png` | — |
| `workbook-preferences-compact-linux.png` | — |
| `workbook-preferences-confirmed-stale-linux.png` | — |
| `workbook-preferences-uncertain-linux.png` | — |
| `workbook-preferences-uncertain-narrow-linux.png` | — |
| `workbook-preferences-unset-linux.png` | — |

### module.workbook.visual.timeline_capture_actions

| Golden filename | Registered fixture IDs |
| --- | --- |
| `timeline-supersession-accepted-linux.png` | — |
| `timeline-supersession-authoring-linux.png` | — |
| `timeline-supersession-review-linux.png` | — |
| `timeline-supersession-review-narrow-linux.png` | — |

### module.workbook.visual.timeline_related_evidence

| Golden filename | Registered fixture IDs |
| --- | --- |
| `timeline-related-evidence-authoring-linux.png` | — |
| `timeline-related-evidence-authoring-narrow-linux.png` | — |
| `timeline-related-evidence-partial-linux.png` | — |
| `timeline-related-evidence-partial-narrow-linux.png` | — |
| `timeline-related-evidence-party-narrow-linux.png` | — |

### package.grid_adapter.visual.capture_test_only_grid_adapter_support_specimens_9c222633ba

| Golden filename | Registered fixture IDs |
| --- | --- |
| `timeline-grid-adapter-fixtures-linux.png` | `visual.fixture.drag_fill_handle`, `visual.fixture.edit_cell`, `visual.fixture.frozen_column`, `visual.fixture.resize_handle`, `visual.fixture.tree_group_row` |

### web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc

| Golden filename | Registered fixture IDs |
| --- | --- |
| `account-menu-controls-compact-linux.png` | — |
| `account-menu-controls-narrow-linux.png` | — |
| `account-menu-controls-short-linux.png` | — |
| `account-menu-long-label-text-spacing-linux.png` | — |
| `account-menu-long-label-zoom-linux.png` | — |
| `account-menu-workbook-root-linux.png` | — |

### web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c

| Golden filename | Registered fixture IDs |
| --- | --- |
| `design-delayed-initial-loading-linux.png` | `visual.fixture.delayed_initial_loading` |
| `design-exposed-theme-states-linux.png` | `visual.fixture.component_state_matrix` |
| `design-grid-background-refresh-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-grid-closed-read-only-rows-linux.png` | — |
| `design-grid-stale-refresh-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-grid-unavailable-initial-load-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-immediate-initial-loading-linux.png` | `visual.fixture.delayed_initial_loading` |

### web.design.visual.lifecycle

| Golden filename | Registered fixture IDs |
| --- | --- |
| `lifecycle-closed-linux.png` | — |
| `lifecycle-comfortable-linux.png` | — |
| `lifecycle-compact-linux.png` | — |
| `lifecycle-confirmed-refresh-failure-linux.png` | — |
| `lifecycle-pending-linux.png` | — |
| `lifecycle-reason-linux.png` | — |
| `lifecycle-review-linux.png` | — |
| `lifecycle-review-narrow-linux.png` | — |
| `lifecycle-review-spacing-linux.png` | — |
| `lifecycle-review-zoom-linux.png` | — |
| `lifecycle-uncertain-linux.png` | — |

### web.design.visual.membership_audit_browsing

| Golden filename | Registered fixture IDs |
| --- | --- |
| `membership-audit-comfortable-linux.png` | — |
| `membership-audit-compact-linux.png` | — |
| `membership-audit-cursor-recovery-linux.png` | — |
| `membership-audit-empty-linux.png` | — |
| `membership-audit-inspected-linux.png` | — |
| `membership-audit-inspected-narrow-linux.png` | — |
| `membership-audit-inspected-spacing-linux.png` | — |
| `membership-audit-inspected-zoom-linux.png` | — |
| `membership-audit-loading-linux.png` | — |
| `membership-audit-stale-linux.png` | — |

### web.design.visual.membership_management_visual

| Golden filename | Registered fixture IDs |
| --- | --- |
| `membership-management-comfortable-linux.png` | — |
| `membership-management-compact-linux.png` | — |
| `membership-management-confirmed-refresh-failure-linux.png` | — |
| `membership-management-loading-linux.png` | — |
| `membership-management-pending-linux.png` | — |
| `membership-management-removal-narrow-linux.png` | — |
| `membership-management-removal-spacing-linux.png` | — |
| `membership-management-removal-zoom-linux.png` | — |
| `membership-management-role-linux.png` | — |
| `membership-management-uncertain-linux.png` | — |

### web.design.visual.metadata_editing

| Golden filename | Registered fixture IDs |
| --- | --- |
| `metadata-closed-linux.png` | — |
| `metadata-comfortable-linux.png` | — |
| `metadata-compact-linux.png` | — |
| `metadata-confirmed-refresh-failure-linux.png` | — |
| `metadata-conflict-linux.png` | — |
| `metadata-dirty-linux.png` | — |
| `metadata-loading-linux.png` | — |
| `metadata-review-narrow-linux.png` | — |
| `metadata-review-spacing-linux.png` | — |
| `metadata-review-zoom-linux.png` | — |
| `metadata-saving-linux.png` | — |
| `metadata-uncertain-linux.png` | — |
