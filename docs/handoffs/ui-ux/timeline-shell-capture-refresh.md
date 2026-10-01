# All Timeline shell captures: refresh record

Trigger: populate every populated Timeline grid with the shared twelve-event
investigation, including backgrounds behind foreground interfaces. Reviewer:
Codex implementation review, 2026-10-01. This extends the
[initial refresh](timeline-fixture-capture-refresh.md); it does not replace that
historical evidence. See the [handoff](timeline-shell-capture-expansion.md) for
implementation, exceptions and verification routing.

## Promotion and review

`make browser-e2e-visual-update` passed 12/12 units at
`20261001T172554Z-p62130`. It refreshed 111 existing PNGs, added none, and kept
all 23 intentional empty/sparse images byte-identical to the follow-up baseline.
All 47 scenarios passed. The three primary 48-row shells retain the already-rich
images from the preceding revision. No new D-VFIX identity was added.

The final anchor promotion passed 12/12 units at `20261001T180405Z-p12672`.
It changed exactly one additional image, the reviewed sparse rollback preview,
for 112 changed PNGs overall. All 23 exception specimens retain their source
inputs and purpose; 22 remain byte-identical and rollback has the explicit
framing correction described below.

All 111 initial changed images were reviewed. Of these, 104 promoted PNGs exactly match
the inspected ordinary candidate `20261001T170003Z-p32794`. Seven promoted
images were imported and inspected separately. The two editor captures now show
the focused Synopsis editor. Exact reference comparisons for the other five
localized 1, 16, 6, 6 and 12 isolated edge pixels respectively, with unchanged
source wording and geometry. No comparison tolerance was increased.
The final rollback image exactly matches the reviewed centered candidate,
SHA-256 `fe64730e7023d4558497a3fe6458f8c73ef1c40293aba05965ade115ef84b05e`.

| Outcome | Meaning |
| --- | --- |
| A | Accepted: rich core and explicit scenario records populate the existing shell/background; original interaction, controls and viewport purpose retained. |
| B | Accepted after correction: active Synopsis editor remains fully visible and focused at its natural scroll position; initial offscreen candidate rejected. |
| C | Accepted: richer fixture as A; final promoted image separately inspected, with isolated edge-pixel variation relative to the reviewed candidate. |
| D | Accepted: saved-view filters still select one matching scenario record or three reviewed records; all selected source fields are populated, full core remains seeded. |
| E | Accepted after correction: sparse rollback confirmation is centered after metadata normalization, with Cancel focused; source data, controls and workflow remain unchanged. |

## Renderer, policies and manifest

Dataset `timeline-investigation-rich-v1`, revision 1; recipe schema v1.
Renderer `visual.renderer.playwright_1_63_0_chromium_1243_linux_amd64`, Chromium
153.0.8010.12, en-US, scale 1. Existing per-capture sizes, density, zoom, theme
and tolerances remain unchanged.
Container digest: `bc6ab0d6d44ff4826e4cb8c1e6d801e185bfc42bb0753f8e2a30efc70db054c7`.
Font digest: `c21f8663e6c8fe72681b2be644aa8398538afc59a0f0cda06b94d46d5fbba5fe`.
Starting manifest: `ce1508c7630322fd903941cf70c6d1abdc0487ff390fff947193ec9925e226bd`.
First promotion manifest: `8c5b07c83844d1a95098c7066e1701af559fc5dcd2d3bfb0fa73d6b87cb8690f`.
Final manifest: `785598423f199c91a965c89ae8aa648f1b3b0653aefc2983cefbdb02f8d7a2fc`.
Unchanged normalization policy: `a8e78deac29c1cbbe57bdb437e57d3b84b033c7d4a979fb91b877d8cef4bca24`.
Capture-data profiles: `cc5baaeeb0a1a621fd087630543a201d8a194a2476d5e8cbd2f08a202023e4bb`.

No source data was normalized. The visual Evidence upload helper now supplies a
fixed authored Received timestamp through its owner; its promoted image is
unchanged from the baseline. A repeat comparison exposed the same omission in
Decision Decided source values; fixed authored target/replacement timestamps
now stabilize the existing accepted Decision framing without another promotion.
Subsequent ordinary comparisons exposed sparse inspector anchors applied before
metadata normalization. Moving History and public-error anchors into the declared
preparation stage preserves their accepted images. Rollback now has an explicit
centered confirmation anchor and retains Cancel focus. This one reviewed framing
change preserves sparse data and workflow. Canonical Date Entered focus remains
at top-left; the two active-edit framing corrections remain as reviewed above.

## Catalog and scenario identities

| Ref | Owner | Catalog row | Scenario |
| --- | --- | --- | --- |
| C01 | `module.collaboration` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `scenario_f710780a7d03` |
| C02 | `module.collaboration` | `module.collaboration.visual.the_visual_harness_asserts_deterministic_timelin_22b64f5dec` | `scenario_4f30b3239772` |
| C03 | `module.collaboration` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | `scenario_62a496085448` |
| C04 | `module.collaboration` | `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc` | `scenario_9cc4be31d2aa` |
| C05 | `module.entities` | `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `scenario_7795c5381ab3` |
| C06 | `module.entities` | `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | `scenario_4447178c020a` |
| C07 | `module.evidence` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `scenario_2bff3f0ed45f` |
| C08 | `module.evidence` | `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830` | `scenario_78faa758f272` |
| C09 | `module.savedviews` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `scenario_5974e42bb8fb` |
| C10 | `module.timeline` | `module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810` | `scenario_b058629ebe04` |
| C11 | `module.timeline` | `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206` | `scenario_5f3469f1fe3b` |
| C12 | `module.timeline` | `module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf` | `scenario_65b64494c412` |
| C13 | `module.workbook` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | `scenario_79fa228c8b2e` |
| C14 | `module.workbook` | `module.workbook.visual.coordination_create_authoring_recovery` | `scenario_d41974eeeb29` |
| C15 | `module.workbook` | `module.workbook.visual.indicator_observations_authoring` | `scenario_4d24a63a7d35` |
| C16 | `module.workbook` | `module.workbook.visual.note_create_authoring_recovery` | `scenario_e55e1a363e58` |
| C17 | `module.workbook` | `module.workbook.visual.preferences` | `scenario_6f57bd00c7a7` |
| C18 | `module.workbook` | `module.workbook.visual.timeline_capture_actions` | `scenario_a2fe4114db52` |
| C19 | `module.workbook` | `module.workbook.visual.timeline_related_evidence` | `scenario_46f536da95f0` |
| C20 | `web.design` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | `scenario_d840806c01cd` |
| C21 | `web.design` | `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c` | `scenario_ddd90978bbf5` |
| C22 | `web.design` | `web.design.visual.lifecycle` | `scenario_4c28f69d060d` |
| C23 | `web.design` | `web.design.visual.membership_audit_browsing` | `scenario_d4214d0e8b76` |
| C24 | `web.design` | `web.design.visual.membership_management_visual` | `scenario_71b349a3a8ce` |
| C25 | `web.design` | `web.design.visual.metadata_editing` | `scenario_7462a6753b02` |
| C26 | `module.workbook` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `scenario_cb54783764e3` |

## Image-by-image disposition

Paths are under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
Fixture mappings remain those of registry v6. Secondary implementation captures
do not adopt another fixture's design identity.

| Image | Exact capture identity | Catalog | Registry fixture | Outcome |
| --- | --- | --- | --- | --- |
| `account-menu-controls-compact-linux.png` | `visual.capture.b550822d7c0b51ba5540` | C20 | Implementation capture | A |
| `account-menu-controls-narrow-linux.png` | `visual.capture.fcea5276d0d0cc990c04` | C20 | Implementation capture | A |
| `account-menu-controls-short-linux.png` | `visual.capture.a9c5fd6a6e5dac6b88e3` | C20 | Implementation capture | A |
| `account-menu-long-label-text-spacing-linux.png` | `visual.capture.e6d45ff45f81e05b1b8d` | C20 | Implementation capture | A |
| `account-menu-long-label-zoom-linux.png` | `visual.capture.acb89fcc3686f8cda4af` | C20 | Implementation capture | A |
| `account-menu-workbook-root-linux.png` | `visual.capture.f30e99b1ce0e75d45851` | C20 | Implementation capture | A |
| `collaboration-conflict-resolver-linux.png` | `visual.capture.0a1171abb1b3e2e4a62c` | C01 | `visual.fixture.same_field_conflict` | A |
| `collaboration-grid-blocked-conflict-linux.png` | `visual.capture.2ccc3a1200064e755088` | C04 | Implementation capture | A |
| `collaboration-grid-conflict-resolver-compact-linux.png` | `visual.capture.fce1904bc3c8d3d577c0` | C03 | Implementation capture | A |
| `collaboration-grid-conflict-resolver-linux.png` | `visual.capture.0ce471473d73f40fe38e` | C03 | Implementation capture | A |
| `collaboration-grid-conflict-resolver-narrow-linux.png` | `visual.capture.712b87e4dbe40f10414b` | C03 | Implementation capture | A |
| `collaboration-grid-presence-markers-linux.png` | `visual.capture.706ab0c0f6658d6f4faf` | C02 | Implementation capture | C |
| `collaboration-presence-markers-linux.png` | `visual.capture.dc2e1b1d5d65f20c8d92` | C01 | `visual.fixture.presence_overflow` | A |
| `coordination-comm-log-authoring-linux.png` | `visual.capture.415d4003c382c81562c4` | C14 | `visual.fixture.contextual_coordination_creation` | A |
| `coordination-handoff-authoring-linux.png` | `visual.capture.b2de4c89fcab66bd6049` | C14 | `visual.fixture.contextual_coordination_creation` | A |
| `coordination-lesson-authoring-linux.png` | `visual.capture.60689ee6ad50bc7ad6ba` | C14 | `visual.fixture.contextual_coordination_creation` | A |
| `coordination-recovery-linux.png` | `visual.capture.b8cb242e972f24566480` | C14 | `visual.fixture.contextual_coordination_creation` | A |
| `coordination-source-narrow-linux.png` | `visual.capture.7361b79164bda73b917e` | C14 | `visual.fixture.contextual_coordination_creation` | A |
| `coordination-status-review-authoring-linux.png` | `visual.capture.ca30bc4813c00e464e4a` | C14 | `visual.fixture.contextual_coordination_creation` | A |
| `design-grid-background-refresh-linux.png` | `visual.capture.c179d064c15077e70ac8` | C21 | `visual.fixture.error_presentation_loci` | A |
| `design-grid-closed-read-only-rows-linux.png` | `visual.capture.cebd310a63d721320aad` | C21 | Implementation capture | A |
| `design-grid-stale-refresh-linux.png` | `visual.capture.c82adffc16722fbe66a5` | C21 | `visual.fixture.error_presentation_loci` | A |
| `entity-mention-chip-states-linux.png` | `visual.capture.bda571d19ea0163393fc` | C05 | `visual.fixture.mention_chip_state_matrix` | C |
| `evidence-grid-timeline-evidence-badge-linux.png` | `visual.capture.47ccc6b491fd5e186555` | C08 | Implementation capture | A |
| `evidence-timeline-evidence-count-linux.png` | `visual.capture.b780497df76862578d0c` | C07 | Implementation capture | A |
| `indicator-observation-authoring-linux.png` | `visual.capture.e084a3414b80dbabab3d` | C15 | `visual.fixture.indicator_observations_authoring` | A |
| `indicator-observation-authoring-narrow-linux.png` | `visual.capture.42d65cdf1d6d6cc7a834` | C15 | `visual.fixture.indicator_observations_authoring` | A |
| `lifecycle-closed-linux.png` | `visual.capture.864872af69315a1b0719` | C22 | Implementation capture | A |
| `lifecycle-comfortable-linux.png` | `visual.capture.78f22d6a0e40b2fd79f4` | C22 | Implementation capture | A |
| `lifecycle-compact-linux.png` | `visual.capture.e9fd1ad9befc46b25dc6` | C22 | Implementation capture | A |
| `lifecycle-confirmed-refresh-failure-linux.png` | `visual.capture.35f4b8ca4813aa639b4d` | C22 | Implementation capture | A |
| `lifecycle-pending-linux.png` | `visual.capture.814c4b925d8d773ee912` | C22 | Implementation capture | A |
| `lifecycle-reason-linux.png` | `visual.capture.048bb8a82f5679c1acf3` | C22 | Implementation capture | A |
| `lifecycle-review-linux.png` | `visual.capture.009315438240bf6d5a7d` | C22 | Implementation capture | A |
| `lifecycle-review-narrow-linux.png` | `visual.capture.e6a743acbe50272f803f` | C22 | Implementation capture | A |
| `lifecycle-review-spacing-linux.png` | `visual.capture.b7b216bd63de6abdf531` | C22 | Implementation capture | A |
| `lifecycle-review-zoom-linux.png` | `visual.capture.fc885a71e33d7313e603` | C22 | Implementation capture | A |
| `lifecycle-uncertain-linux.png` | `visual.capture.ccd98280a4e93586c904` | C22 | Implementation capture | A |
| `linked-note-authoring-linux.png` | `visual.capture.9bb90770e53a11be45c3` | C16 | Implementation capture | A |
| `linked-note-authoring-narrow-linux.png` | `visual.capture.f49e759c383c4e0f1f5a` | C16 | Implementation capture | A |
| `linked-note-recovery-linux.png` | `visual.capture.a1a830e781756be72b5f` | C16 | Implementation capture | A |
| `linked-note-recovery-narrow-linux.png` | `visual.capture.0f6082dcb8e7ec9c062f` | C16 | Implementation capture | A |
| `linked-note-source-narrow-linux.png` | `visual.capture.2e7d95ce6b3840cc81a7` | C16 | Implementation capture | A |
| `membership-audit-comfortable-linux.png` | `visual.capture.9f8764a09ee93ac496e6` | C23 | Implementation capture | A |
| `membership-audit-compact-linux.png` | `visual.capture.1ba7079890196b4a0003` | C23 | Implementation capture | A |
| `membership-audit-cursor-recovery-linux.png` | `visual.capture.004d9c4b1ba9285c9a45` | C23 | Implementation capture | A |
| `membership-audit-empty-linux.png` | `visual.capture.188569fc954860444a99` | C23 | Implementation capture | A |
| `membership-audit-inspected-linux.png` | `visual.capture.a11484007de0a9246fea` | C23 | Implementation capture | A |
| `membership-audit-inspected-narrow-linux.png` | `visual.capture.914ce670aa5017b301a7` | C23 | Implementation capture | A |
| `membership-audit-inspected-spacing-linux.png` | `visual.capture.5574ccaedbb0e2586598` | C23 | Implementation capture | A |
| `membership-audit-inspected-zoom-linux.png` | `visual.capture.773aa40b61a17f34df0f` | C23 | Implementation capture | A |
| `membership-audit-loading-linux.png` | `visual.capture.60f7acbd4fc5682e8346` | C23 | Implementation capture | A |
| `membership-audit-stale-linux.png` | `visual.capture.2751e548dfa24d7fb9a3` | C23 | Implementation capture | A |
| `membership-management-comfortable-linux.png` | `visual.capture.b6fae0a185503be4b49e` | C24 | Implementation capture | A |
| `membership-management-compact-linux.png` | `visual.capture.be1865e801a7fdab4caa` | C24 | Implementation capture | A |
| `membership-management-confirmed-refresh-failure-linux.png` | `visual.capture.65f37803d46189a11a85` | C24 | Implementation capture | A |
| `membership-management-loading-linux.png` | `visual.capture.7a5985aa2e474a629d10` | C24 | Implementation capture | A |
| `membership-management-pending-linux.png` | `visual.capture.278029016c6b60a9745a` | C24 | Implementation capture | A |
| `membership-management-removal-narrow-linux.png` | `visual.capture.00f7c068d36017c70d3c` | C24 | Implementation capture | A |
| `membership-management-removal-spacing-linux.png` | `visual.capture.f94b83236deaeebca37d` | C24 | Implementation capture | A |
| `membership-management-removal-zoom-linux.png` | `visual.capture.ffee60c42f4759ff4b91` | C24 | Implementation capture | A |
| `membership-management-role-linux.png` | `visual.capture.97d5ba9d8a7cde127c5c` | C24 | Implementation capture | A |
| `membership-management-uncertain-linux.png` | `visual.capture.b11f46b442198c4d0995` | C24 | Implementation capture | A |
| `metadata-closed-linux.png` | `visual.capture.136f66613d4eb5ebb348` | C25 | Implementation capture | A |
| `metadata-comfortable-linux.png` | `visual.capture.f2768bd50460059330f7` | C25 | Implementation capture | A |
| `metadata-compact-linux.png` | `visual.capture.e2466c4700ff818cf4aa` | C25 | Implementation capture | A |
| `metadata-confirmed-refresh-failure-linux.png` | `visual.capture.59b95b912d49f30a9759` | C25 | Implementation capture | A |
| `metadata-conflict-linux.png` | `visual.capture.f77270fae8efcf6d0e28` | C25 | Implementation capture | A |
| `metadata-dirty-linux.png` | `visual.capture.4030c6a79227345943e0` | C25 | Implementation capture | A |
| `metadata-loading-linux.png` | `visual.capture.3b6710f5709fc093a054` | C25 | Implementation capture | A |
| `metadata-review-narrow-linux.png` | `visual.capture.b870f1c3884a0a5b59d1` | C25 | Implementation capture | A |
| `metadata-review-spacing-linux.png` | `visual.capture.68e190a75befd7f69f33` | C25 | Implementation capture | A |
| `metadata-review-zoom-linux.png` | `visual.capture.00ea8c6347526eb87775` | C25 | Implementation capture | A |
| `metadata-saving-linux.png` | `visual.capture.048441e5821b032d0f67` | C25 | Implementation capture | A |
| `metadata-uncertain-linux.png` | `visual.capture.dc789d498fccc0471bc9` | C25 | Implementation capture | A |
| `record-relationships-mention-chips-linux.png` | `visual.capture.cf8fe458f89df653e074` | C06 | Implementation capture | A |
| `timeline-grid-active-edit-cell-linux.png` | `visual.capture.7e1ef25fc40811162aba` | C12 | `visual.fixture.edit_cell` | B |
| `timeline-grid-grouped-grid-linux.png` | `visual.capture.28368d3fafa98a438776` | C10 | Implementation capture | A |
| `timeline-grid-timeline-default-linux.png` | `visual.capture.16689335246c227e88d8` | C11 | Implementation capture | A |
| `timeline-mutation-active-edit-cell-linux.png` | `visual.capture.45b0f9bac4c7f6bc0412` | C13 | `visual.fixture.edit_cell` | B |
| `timeline-mutation-pending-replay-status-linux.png` | `visual.capture.a5356d72e06448c07560` | C13 | Implementation capture | A |
| `timeline-mutation-transaction-recovery-panel-compact-linux.png` | `visual.capture.af13dd93abd19940247c` | C13 | Implementation capture | A |
| `timeline-mutation-transaction-recovery-panel-linux.png` | `visual.capture.124cf27d3d0177dea828` | C13 | Implementation capture | C |
| `timeline-mutation-transaction-recovery-panel-narrow-linux.png` | `visual.capture.c49072efd841f8cb0cea` | C13 | Implementation capture | C |
| `timeline-related-evidence-authoring-linux.png` | `visual.capture.3d782a7e7d58929c9e62` | C19 | Implementation capture | A |
| `timeline-related-evidence-authoring-narrow-linux.png` | `visual.capture.07abf9bcf805add36738` | C19 | Implementation capture | A |
| `timeline-related-evidence-partial-linux.png` | `visual.capture.58335086ceda35c551d2` | C19 | Implementation capture | A |
| `timeline-related-evidence-partial-narrow-linux.png` | `visual.capture.6bf80687111d542513ec` | C19 | Implementation capture | A |
| `timeline-related-evidence-party-narrow-linux.png` | `visual.capture.55a22539e04b7331aa15` | C19 | Implementation capture | A |
| `timeline-supersession-accepted-linux.png` | `visual.capture.37b80519177fc24904ff` | C18 | Implementation capture | A |
| `timeline-supersession-authoring-linux.png` | `visual.capture.5fb1a56e0f004366997f` | C18 | Implementation capture | A |
| `timeline-supersession-review-linux.png` | `visual.capture.e66b3638a68983c4947a` | C18 | Implementation capture | A |
| `timeline-supersession-review-narrow-linux.png` | `visual.capture.e75993e88b8b3f8b2591` | C18 | Implementation capture | A |
| `workbook-preferences-comfortable-linux.png` | `visual.capture.1fbdc3fd547f52dc8eab` | C17 | Implementation capture | A |
| `workbook-preferences-compact-linux.png` | `visual.capture.8dc97521e5c4f473e3be` | C17 | Implementation capture | A |
| `workbook-preferences-confirmed-stale-linux.png` | `visual.capture.8a036912cdcb3f180f97` | C17 | Implementation capture | C |
| `workbook-preferences-uncertain-linux.png` | `visual.capture.80a3b23b316daf1b67d2` | C17 | Implementation capture | A |
| `workbook-preferences-uncertain-narrow-linux.png` | `visual.capture.81c8c654c29ea74cd73e` | C17 | Implementation capture | A |
| `workbook-preferences-unset-linux.png` | `visual.capture.0e33fae0c136cd1a3309` | C17 | Implementation capture | A |
| `workbook-query-saved-view-query-controls-linux.png` | `visual.capture.b6524c89920690198ffb` | C09 | `visual.fixture.saved_view_query_controls_and_grouped_result` | D |
| `workbook-view-bar-filter-editing-overflow-linux.png` | `visual.capture.9dece7e0dfcaa9014a74` | C09 | Implementation capture | D |
| `workbook-view-bar-long-columns-linux.png` | `visual.capture.1f0c5e239a5b50226642` | C09 | Implementation capture | D |
| `workbook-view-bar-maximum-pressure-base-linux.png` | `visual.capture.255f84878da966bb6a2f` | C09 | Implementation capture | D |
| `workbook-view-bar-maximum-pressure-compact-linux.png` | `visual.capture.31051420009eef38c992` | C09 | Implementation capture | D |
| `workbook-view-bar-maximum-pressure-narrow-linux.png` | `visual.capture.000d8edec05917c8680f` | C09 | Implementation capture | D |
| `workbook-view-bar-ordered-maximum-sort-linux.png` | `visual.capture.d994111394c03b53cee4` | C09 | Implementation capture | D |
| `workbook-view-bar-saved-view-actions-linux.png` | `visual.capture.f08b4d814d30429d51f3` | C09 | Implementation capture | D |
| `workbook-view-bar-saved-view-clean-linux.png` | `visual.capture.4f5e321cc67de64bdcfa` | C09 | Implementation capture | D |
| `workbook-view-bar-saved-view-modified-linux.png` | `visual.capture.41ad8eafac6d081ac683` | C09 | Implementation capture | D |
| `workbook-view-bar-text-spacing-linux.png` | `visual.capture.82fbeaf5595c1c0b0e18` | C09 | Implementation capture | D |
| `workbook-view-bar-zoom-200-linux.png` | `visual.capture.67b5c747c2e07b27ca14` | C09 | Implementation capture | D |

The final anchor promotion adds this separately reviewed image:

| Image | Exact capture identity | Catalog | Registry fixture | Outcome |
| --- | --- | --- | --- | --- |
| `workbook-inspector-rollback-preview-linux.png` | `visual.capture.3451d6d508d3782f0df3` | C26 | `visual.fixture.base_inspector` (`D-VFIX-002`) | E |

## Final ordinary verification and cleanup

Two fresh `make browser-e2e-visual` runs passed against the final manifest above:
`20261001T181153Z-p49823` and `20261001T181154Z-p50017`. Each passed 12/12
harness units and all 47 scenarios, with no skipped or flaky scenarios. Both
reconcile 255 captures to 255 goldens and 29 registered fixtures without errors.
All 112 changed PNGs match exact reviewed bytes. UI review used digest-verified
Make imports; standalone reference comparisons carry no DOM or accessibility
claim. Final source and exception checks, completed session/scratch cleanup and
skipped broader claims are recorded in the handoff.
