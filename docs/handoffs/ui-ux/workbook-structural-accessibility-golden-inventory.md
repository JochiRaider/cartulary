# Structural accessibility golden refresh inventory

The user explicitly authorized a full-catalog golden refresh. The supported
`make browser-e2e-visual-update CARTULARY_HARNESS_CACHE_MODE=off` passed 12/12
units at `20261001T002025Z-p90787`. Reconciliation accounts for all 255 active
goldens/captures and 29 registered fixtures without missing or ambiguous mappings.
112 PNGs and their tool-produced manifest changed; 143 PNGs were retained.

Every changed image was reviewed. Two fresh ordinary validations passed 12/12:
`20261001T002656Z-p42334` and `20261001T003638Z-p44113`. An intervening interrupted
attempt is recorded separately in the correction handoff and is not a pass.

Viewport, browser zoom, masks, scroll normalization, crop scope, renderer pins,
and fixture identities were unchanged. The visible Row default is the primary
trigger. The user also authorized the inherited retained-authoring, Inspector,
dismissed-mention and Evidence differences documented in the correction handoff.
This inventory is human review evidence, never an executable harness input.

Paths below are relative to `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
Rows are taken from reconciliation and checked against the authored catalog.
A dash means no registered fixture; the exact active capture mapping still exists.

| Golden | Owner row | Registered fixture |
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
| `collaboration-grid-presence-markers-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_deterministic_timelin_22b64f5dec` | — |
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
| `coordination-status-review-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `decision-supersession-accepted-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | — |
| `decision-supersession-review-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | — |
| `design-delayed-initial-loading-linux.png` | `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `visual.fixture.delayed_initial_loading` |
| `design-grid-background-refresh-linux.png` | `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `visual.fixture.error_presentation_loci` |
| `design-grid-closed-read-only-rows-linux.png` | `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | — |
| `design-grid-stale-refresh-linux.png` | `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `visual.fixture.error_presentation_loci` |
| `design-grid-unavailable-initial-load-linux.png` | `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `visual.fixture.error_presentation_loci` |
| `design-immediate-initial-loading-linux.png` | `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `visual.fixture.delayed_initial_loading` |
| `entity-mention-chip-states-linux.png` | `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `visual.fixture.mention_chip_state_matrix` |
| `evidence-affordance-states-linux.png` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` |
| `evidence-grid-available-evidence-linux.png` | `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af` | — |
| `evidence-grid-blocked-preview-linux.png` | `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830` | — |
| `evidence-grid-requested-evidence-linux.png` | `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af` | — |
| `incident-directory-default-timeline-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.default_timeline_workbook_shell` |
| `indicator-lifecycle-authoring-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-observation-authoring-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `lifecycle-closed-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-comfortable-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-compact-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-confirmed-refresh-failure-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-pending-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-reason-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-uncertain-linux.png` | `web.design.visual.lifecycle` | — |
| `linked-note-authoring-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-recovery-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-recovery-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `membership-audit-cursor-recovery-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-empty-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-loading-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-stale-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-management-confirmed-refresh-failure-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-loading-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-pending-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-role-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-uncertain-linux.png` | `web.design.visual.membership_management_visual` | — |
| `metadata-closed-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-conflict-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-dirty-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-loading-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-saving-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-uncertain-linux.png` | `web.design.visual.metadata_editing` | — |
| `ordinary-closed-retained-narrow-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-recovery-1280-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-recovery-390-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-reference-authoring-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `record-relationships-evidence-access-linux.png` | `module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc` | — |
| `record-relationships-mention-chips-linux.png` | `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | — |
| `record-relationships-task-requests-linux.png` | `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc` | `visual.fixture.task_requests_or_decisions` |
| `timeline-grid-active-edit-cell-linux.png` | `module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf` | `visual.fixture.edit_cell` |
| `timeline-grid-timeline-default-linux.png` | `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206` | — |
| `timeline-mutation-active-edit-cell-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | `visual.fixture.edit_cell` |
| `timeline-mutation-empty-timeline-query-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-pending-replay-status-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-compact-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-narrow-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-related-evidence-authoring-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-partial-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-partial-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-supersession-accepted-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `timeline-supersession-authoring-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `timeline-supersession-review-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `workbook-inspector-attached-edit-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-details-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-history-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-public-error-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-relationships-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-retained-draft-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-preferences-confirmed-stale-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-uncertain-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-unset-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-query-empty-closed-read-only-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-compact-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-density-comfortable-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-density-compact-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-narrow-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-successful-query-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-text-spacing-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-zoom-200-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-filtered-empty-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
