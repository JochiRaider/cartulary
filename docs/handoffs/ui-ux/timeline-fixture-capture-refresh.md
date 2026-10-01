# Timeline fixture and capture refresh record

Follow-up: [all Timeline shell captures](timeline-shell-capture-expansion.md) expands
the rich dataset beyond the three primary shells and records its own promotion evidence.

Status: reviewed, promoted and verified by two fresh ordinary visual passes.

Trigger: adopt Timeline fixture/capture remediation on `main` baseline `6456013e8`.
Reviewer: Codex implementation review, 2026-10-01.
Dataset: `timeline-investigation-rich-v1`, revision 1; recipe schema v1.
The rejected post-promotion comparisons and their preparation corrections are
recorded in the remediation handoff. All 30 Timeline inspector capture identities now declare
Recorded and Edited read-only metadata explicitly. The Evidence affordance capture
also declares its inspector Edited metadata. Their wrapping cannot vary the
following section's fractional scroll origin. Authored and historical values stay intact.
The focused visual slice passes against the same promoted images.
No new goldens or D-VFIX identities. Timeline v2 and product defaults are unchanged.

## Renderer and admission

Renderer: `visual.renderer.playwright_1_63_0_chromium_1243_linux_amd64`.
Container: `mcr.microsoft.com/playwright@sha256:bc6ab0d6d44ff4826e4cb8c1e6d801e185bfc42bb0753f8e2a30efc70db054c7`.
Fonts: `c21f8663e6c8fe72681b2be644aa8398538afc59a0f0cda06b94d46d5fbba5fe`.
Locale en-US, device scale 1, Chromium 153.0.8010.12; existing per-capture
viewport, zoom, theme, density and comparison tolerances are unchanged.

Previous golden manifest SHA-256: `5155ea46750774fda1900edbc92690c8fecf923bbb4b85e5ab119a1516282d82`.
Normalization policy SHA-256: `a8e78deac29c1cbbe57bdb437e57d3b84b033c7d4a979fb91b877d8cef4bca24`.

Unchanged baseline: `20261001T125324Z-p97418`, ordinary visual pass, 12/12 units.
Reviewed candidate: `20261001T143356Z-p28904`, all 255 captures executed;
255 active goldens, zero missing/orphan/ambiguous entries, all 29 registry fixtures
resolved. There were 72 image mismatches and no functional/preparation failures.
The target correctly failed comparison; that failure is not a product-pass claim.

## Review dispositions

| Code | Accepted change and review finding |
| --- | --- |
| R1 | Rich twelve-event core and deterministic continuations; canonical semantic selection and visible Date Entered focus at top-left. Narrow/compact control pressure retained. |
| R2 | History/rollback now selects an explicit RAW source edit after unchanged sparse reading captures. Authored ISO/UUID before/after values survive; only declared attribution/history metadata changes. Confirmation and error states remain legible. |
| R3 | Auxiliary column selection omits runtime Owner/Outgoing Owner references; target picker and review/recovery content remain. Scoped supersession metadata preserves reasons, labels and states. |
| R4 | Fixed authored Evidence Requested/Received timestamps replace reliance on broad timestamp rewriting. Source excerpts, labels and relevant controls remain visible. |
| R5 | Authored audit values and example timestamps retained. Expanded incident-ID metadata has an exact scoped rule. Published values and redaction remain intact. |
| R6 | Authored activity times retained; technical conflict/supersession references normalized only in declared leaves. Conflict, review and accepted states remain distinct. |
| R7 | Removal of page-wide text-node rewrites changes text rendering in the scrolled narrow preference recovery capture. Wording, controls and layout remain intact; both final repeat runs confirm stability. |
| R8 | Source-preservation refresh: authored dates, UUID-shaped fixture values, incident form keys or fixed synthetic account labels survive. Existing capture purpose and controls remain intact. |

Every listed image was inspected through digest-verified Make UI-review imports.
Sixty final candidate images were byte-identical to the inspected preceding
candidate; the remaining twelve were inspected again. Review also covered all
six rich inspector attachments. The twelve layout-study attachments remain
available in the candidate report. Supporting observations add no goldens.

## Catalog ownership

| Ref | Owner | Catalog row | Scenario |
| --- | --- | --- | --- |
| S01 | `module.collaboration` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `scenario_f710780a7d03` |
| S02 | `module.collaboration` | `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc` | `scenario_9cc4be31d2aa` |
| S03 | `module.evidence` | `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830` | `scenario_78faa758f272` |
| S04 | `module.evidence` | `module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc` | `scenario_5c893193b257` |
| S05 | `module.evidence` | `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af` | `scenario_870610bb7b2c` |
| S06 | `module.networkflow` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `scenario_252dd2b0ff55` |
| S07 | `module.savedviews` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `scenario_5974e42bb8fb` |
| S08 | `module.timeline` | `module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810` | `scenario_b058629ebe04` |
| S09 | `module.timeline` | `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206` | `scenario_5f3469f1fe3b` |
| S10 | `module.workbook` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `scenario_e6ffdeb2d292` |
| S11 | `module.workbook` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `scenario_cb54783764e3` |
| S12 | `module.workbook` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | `scenario_79fa228c8b2e` |
| S13 | `module.workbook` | `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc` | `scenario_2ad2818dac42` |
| S14 | `module.workbook` | `module.workbook.visual.contextual_task_decision_creation` | `scenario_4a08391df78c` |
| S15 | `module.workbook` | `module.workbook.visual.decision_supersession_review_recovery` | `scenario_d87078727b34` |
| S16 | `module.workbook` | `module.workbook.visual.ordinary_create_authoring_recovery` | `scenario_627d716b1de2` |
| S17 | `module.workbook` | `module.workbook.visual.preferences` | `scenario_6f57bd00c7a7` |
| S18 | `module.workbook` | `module.workbook.visual.timeline_capture_actions` | `scenario_a2fe4114db52` |
| S19 | `web.design` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | `scenario_d840806c01cd` |
| S20 | `web.design` | `web.design.visual.administrative_audit_browsing` | `scenario_2da065feffeb` |
| S21 | `web.design` | `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `scenario_ddd90978bbf5` |
| S22 | `web.design` | `web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde` | `scenario_46ea092e7f2b` |
| S23 | `web.design` | `web.design.visual.membership_audit_browsing` | `scenario_d4214d0e8b76` |
| S24 | `web.design` | `web.design.visual.membership_management_visual` | `scenario_71b349a3a8ce` |

## Image-by-image disposition

Paths below are under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
Registry design identities apply only to their declared primary fixture mappings;
secondary artifacts do not create another adoption of that identity.

| Image | Capture identity | Catalog | Registry fixture / design owner | Outcome |
| --- | --- | --- | --- | --- |
| `account-menu-deployment-root-linux.png` | `visual.capture.17e0b9ae241d550045ee` | S19 | Implementation-only capture | Accepted — R8 |
| `administrative-audit-empty-linux.png` | `visual.capture.ded375bfa54c4f384b4e` | S20 | Implementation-only capture | Accepted — R8 |
| `administrative-audit-inspected-linux.png` | `visual.capture.8ce15444b98aab68fa04` | S20 | Implementation-only capture | Accepted — R8 |
| `administrative-audit-loading-linux.png` | `visual.capture.eec377109485f5df9948` | S20 | Implementation-only capture | Accepted — R8 |
| `administrative-audit-page-two-linux.png` | `visual.capture.cbad2d56dfefaa711a07` | S20 | Implementation-only capture | Accepted — R8 |
| `administrative-audit-stale-linux.png` | `visual.capture.72a8a9282945100c7b29` | S20 | Implementation-only capture | Accepted — R8 |
| `administrative-audit-unavailable-linux.png` | `visual.capture.e06bc6dff803d369747e` | S20 | Implementation-only capture | Accepted — R8 |
| `collaboration-conflict-resolver-linux.png` | `visual.capture.0a1171abb1b3e2e4a62c` | S01 | `visual.fixture.same_field_conflict` / `D-VFIX-003` | Accepted — R6 |
| `collaboration-grid-blocked-conflict-linux.png` | `visual.capture.2ccc3a1200064e755088` | S02 | Implementation-only capture | Accepted — R6 |
| `contextual-decision-authoring-linux.png` | `visual.capture.44e851b61fcacf540838` | S14 | Implementation-only capture | Accepted — R4 |
| `contextual-decision-recovery-linux.png` | `visual.capture.597c2cd3c6d4b9ed069b` | S14 | Implementation-only capture | Accepted — R4 |
| `contextual-task-request-authoring-linux.png` | `visual.capture.010db75b102f203f2286` | S14 | Implementation-only capture | Accepted — R4 |
| `contextual-task-request-recovery-linux.png` | `visual.capture.0a3f3690592b622a6288` | S14 | Implementation-only capture | Accepted — R4 |
| `decision-supersession-accepted-linux.png` | `visual.capture.47ccc5e6f032f3f4a9d3` | S15 | Implementation-only capture | Accepted — R3 |
| `decision-supersession-review-linux.png` | `visual.capture.ee4d0843cfc3a8f32681` | S15 | Implementation-only capture | Accepted — R3 |
| `decision-supersession-review-narrow-linux.png` | `visual.capture.1a8b2c9007adbde61581` | S15 | Implementation-only capture | Accepted — R3 |
| `design-grid-background-refresh-linux.png` | `visual.capture.c179d064c15077e70ac8` | S21 | `visual.fixture.error_presentation_loci` / `D-VFIX-014` | Accepted — R8 |
| `design-grid-closed-read-only-rows-linux.png` | `visual.capture.cebd310a63d721320aad` | S21 | Implementation-only capture | Accepted — R8 |
| `design-grid-stale-refresh-linux.png` | `visual.capture.c82adffc16722fbe66a5` | S21 | `visual.fixture.error_presentation_loci` / `D-VFIX-014` | Accepted — R8 |
| `evidence-grid-available-evidence-linux.png` | `visual.capture.ed2766061fb0b5e26465` | S05 | Implementation-only capture | Accepted — R4 |
| `evidence-grid-blocked-preview-linux.png` | `visual.capture.08fb74af7c121f4c746b` | S03 | Implementation-only capture | Accepted — R4 |
| `evidence-grid-requested-evidence-linux.png` | `visual.capture.45f7ce6d292620442137` | S05 | Implementation-only capture | Accepted — R4 |
| `incident-create-expanded-details-linux.png` | `visual.capture.10c44722b7aa59411f31` | S22 | Implementation-only capture | Accepted — R8 |
| `incident-create-pending-linux.png` | `visual.capture.555eb9a04b92ff7930c3` | S22 | Implementation-only capture | Accepted — R8 |
| `incident-create-recovery-short-linux.png` | `visual.capture.8c4a9ca3c4fc73840423` | S22 | Implementation-only capture | Accepted — R8 |
| `incident-create-recovery-text-spacing-linux.png` | `visual.capture.d62db66130daec6e0428` | S22 | Implementation-only capture | Accepted — R8 |
| `incident-create-recovery-zoom-linux.png` | `visual.capture.a2f89c5bd4c5c2483f0f` | S22 | Implementation-only capture | Accepted — R8 |
| `incident-directory-compact-desktop-workbook-shell-linux.png` | `visual.capture.5895c38a72f7e019fbed` | S10 | `visual.fixture.compact_desktop_workbook_shell` / `D-VFIX-011` | Accepted — R1 |
| `incident-directory-default-timeline-workbook-shell-linux.png` | `visual.capture.7314c63e882d78cd3cf8` | S10 | `visual.fixture.default_timeline_workbook_shell` / `D-VFIX-001` | Accepted — R1 |
| `incident-directory-narrow-desktop-workbook-shell-linux.png` | `visual.capture.474f3c17ecf77ef9d8b2` | S10 | `visual.fixture.narrow_desktop_workbook_shell` / `D-VFIX-010` | Accepted — R1 |
| `membership-audit-compact-linux.png` | `visual.capture.1ba7079890196b4a0003` | S23 | Implementation-only capture | Accepted — R5 |
| `membership-audit-cursor-recovery-linux.png` | `visual.capture.004d9c4b1ba9285c9a45` | S23 | Implementation-only capture | Accepted — R5 |
| `membership-audit-empty-linux.png` | `visual.capture.188569fc954860444a99` | S23 | Implementation-only capture | Accepted — R5 |
| `membership-audit-inspected-linux.png` | `visual.capture.a11484007de0a9246fea` | S23 | Implementation-only capture | Accepted — R5 |
| `membership-audit-loading-linux.png` | `visual.capture.60f7acbd4fc5682e8346` | S23 | Implementation-only capture | Accepted — R5 |
| `membership-audit-stale-linux.png` | `visual.capture.2751e548dfa24d7fb9a3` | S23 | Implementation-only capture | Accepted — R5 |
| `membership-management-comfortable-linux.png` | `visual.capture.b6fae0a185503be4b49e` | S24 | Implementation-only capture | Accepted — R8 |
| `membership-management-compact-linux.png` | `visual.capture.be1865e801a7fdab4caa` | S24 | Implementation-only capture | Accepted — R8 |
| `membership-management-confirmed-refresh-failure-linux.png` | `visual.capture.65f37803d46189a11a85` | S24 | Implementation-only capture | Accepted — R8 |
| `membership-management-pending-linux.png` | `visual.capture.278029016c6b60a9745a` | S24 | Implementation-only capture | Accepted — R8 |
| `membership-management-removal-narrow-linux.png` | `visual.capture.00f7c068d36017c70d3c` | S24 | Implementation-only capture | Accepted — R8 |
| `membership-management-removal-spacing-linux.png` | `visual.capture.f94b83236deaeebca37d` | S24 | Implementation-only capture | Accepted — R8 |
| `membership-management-removal-zoom-linux.png` | `visual.capture.ffee60c42f4759ff4b91` | S24 | Implementation-only capture | Accepted — R8 |
| `membership-management-role-linux.png` | `visual.capture.97d5ba9d8a7cde127c5c` | S24 | Implementation-only capture | Accepted — R8 |
| `membership-management-uncertain-linux.png` | `visual.capture.b11f46b442198c4d0995` | S24 | Implementation-only capture | Accepted — R8 |
| `network-flow-analysis-accepted-inspector-linux.png` | `visual.capture.3c29ea9e2d992ca33e21` | S06 | `visual.fixture.claimed_network_analysis_workspace_states` | Accepted — R8 |
| `network-flow-analysis-graph-contributors-linux.png` | `visual.capture.fe36571af4277f846e89` | S06 | `visual.fixture.claimed_network_analysis_workspace_states` | Accepted — R8 |
| `network-flow-analysis-narrow-query-controls-linux.png` | `visual.capture.4e8d00b0f1c95a3f8c6a` | S06 | `visual.fixture.claimed_network_analysis_narrow_workspace` | Accepted — R8 |
| `ordinary-recovery-1280-linux.png` | `visual.capture.f2810371a0de5f28e842` | S16 | Implementation-only capture | Accepted — R4 |
| `ordinary-reference-authoring-linux.png` | `visual.capture.a90717d56123e5a1f4ed` | S16 | Implementation-only capture | Accepted — R3 |
| `record-relationships-evidence-access-linux.png` | `visual.capture.b18314ed6d6b90ca6c33` | S04 | Implementation-only capture | Accepted — R4 |
| `record-relationships-task-requests-linux.png` | `visual.capture.01b96e83320b93460e07` | S13 | `visual.fixture.task_requests_or_decisions` | Accepted — R3 |
| `timeline-grid-grouped-grid-linux.png` | `visual.capture.28368d3fafa98a438776` | S08 | Implementation-only capture | Accepted — R8 |
| `timeline-grid-timeline-default-linux.png` | `visual.capture.16689335246c227e88d8` | S09 | Implementation-only capture | Accepted — R8 |
| `timeline-mutation-active-edit-cell-linux.png` | `visual.capture.45b0f9bac4c7f6bc0412` | S12 | `visual.fixture.edit_cell` | Accepted — R8 |
| `timeline-mutation-pending-replay-status-linux.png` | `visual.capture.a5356d72e06448c07560` | S12 | Implementation-only capture | Accepted — R8 |
| `timeline-supersession-accepted-linux.png` | `visual.capture.37b80519177fc24904ff` | S18 | Implementation-only capture | Accepted — R6 |
| `timeline-supersession-authoring-linux.png` | `visual.capture.5fb1a56e0f004366997f` | S18 | Implementation-only capture | Accepted — R6 |
| `timeline-supersession-review-linux.png` | `visual.capture.e66b3638a68983c4947a` | S18 | Implementation-only capture | Accepted — R6 |
| `workbook-inspector-destructive-confirmation-linux.png` | `visual.capture.06903c2d94c98d2d2eaf` | S11 | `visual.fixture.destructive_actions` / `D-VFIX-008` | Accepted — R2 |
| `workbook-inspector-history-linux.png` | `visual.capture.21a59ffec6a398738353` | S11 | `visual.fixture.base_inspector` / `D-VFIX-002` | Accepted — R2 |
| `workbook-inspector-public-error-linux.png` | `visual.capture.1aebc4c7b4cfc2965a5f` | S11 | `visual.fixture.base_inspector` / `D-VFIX-002` | Accepted — R2 |
| `workbook-inspector-rollback-preview-linux.png` | `visual.capture.3451d6d508d3782f0df3` | S11 | `visual.fixture.base_inspector` / `D-VFIX-002` | Accepted — R2 |
| `workbook-preferences-uncertain-narrow-linux.png` | `visual.capture.81c8c654c29ea74cd73e` | S17 | Implementation-only capture | Accepted — R7 |
| `workbook-query-saved-view-query-controls-linux.png` | `visual.capture.b6524c89920690198ffb` | S07 | `visual.fixture.saved_view_query_controls_and_grouped_result` | Accepted — R8 |
| `workbook-view-bar-long-columns-linux.png` | `visual.capture.1f0c5e239a5b50226642` | S07 | Implementation-only capture | Accepted — R8 |
| `workbook-view-bar-maximum-pressure-base-linux.png` | `visual.capture.255f84878da966bb6a2f` | S07 | Implementation-only capture | Accepted — R8 |
| `workbook-view-bar-ordered-maximum-sort-linux.png` | `visual.capture.d994111394c03b53cee4` | S07 | Implementation-only capture | Accepted — R8 |
| `workbook-view-bar-saved-view-actions-linux.png` | `visual.capture.f08b4d814d30429d51f3` | S07 | Implementation-only capture | Accepted — R8 |
| `workbook-view-bar-saved-view-clean-linux.png` | `visual.capture.4f5e321cc67de64bdcfa` | S07 | Implementation-only capture | Accepted — R8 |
| `workbook-view-bar-saved-view-modified-linux.png` | `visual.capture.41ad8eafac6d081ac683` | S07 | Implementation-only capture | Accepted — R8 |
| `workbook-view-bar-text-spacing-linux.png` | `visual.capture.82fbeaf5595c1c0b0e18` | S07 | Implementation-only capture | Accepted — R8 |

## Promotion, repetition and cleanup

`make browser-e2e-visual-update` passed 12/12 units at `20261001T144454Z-p79181`.
It promoted exactly the 72 listed images. Sixty-eight matched the reviewed
candidate bytes; the four remaining promoted PNGs were separately imported,
digest-verified, visually inspected and accepted with unchanged content and
usable geometry. No unreviewed image entered the promoted inventory.

Final manifest SHA-256:
`ce1508c7630322fd903941cf70c6d1abdc0487ff390fff947193ec9925e226bd`.
Two fresh ordinary visual passes against this same manifest completed successfully; their exact roots are recorded below.

Earlier rejected candidates are documented in the remediation handoff. No image
was promoted to hide a readiness failure, unexplained source mutation or product
regression. Artifact-review sessions use owned cleanup; temporary bundles expire
and private caller scratch is removed at closure.

The handoff records semantic coverage, owner validation, skipped broad/release
claims, and rollback. Restore source, projections, goldens and manifest together;
no application-data migration is needed.

## Final ordinary verification

`make browser-e2e-visual` passed 12/12 units independently at
`20261001T154606Z-p29821` and `20261001T154606Z-p29828`. Both use the final manifest
SHA-256 above. Each resolves all 255 captures, 255 active goldens and 29 registry
fixtures with zero missing, orphaned or ambiguous joins. All 47 Playwright
scenarios passed without skips, retries or flaky outcomes. Each retains all
12 layout-study and 6 rich-inspector attachments; owned cleanup passed.

The final metadata policy SHA-256 is
`a8e78deac29c1cbbe57bdb437e57d3b84b033c7d4a979fb91b877d8cef4bca24`.
The additional exact timestamp declarations required no further image changes:
the final inventory remains 72 refreshed existing images and zero added images.
