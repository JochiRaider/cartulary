# Cartulary Visual Golden Reduction Ledger

**Status**: Human review record for visual-golden consolidation
**Authority**: Implementation-support record only. The authored catalog,
fixture registry, normalization policy, Timeline profiles and reconciliation
artifacts remain the executable owners. No test, generator, conformance check or
release evidence may read this file. Retirement decisions follow the
[maintenance guide](cartulary_visual_golden_maintenance.md).

## Method

Reduce maintained pixel baselines by consolidating visual obligations, not by
targeting a file count. Every retired golden names its retained pixel evidence or
replacement nonpixel assertion, and is deleted only after ordinary
reconciliation classifies it as a zero-consumer `orphan`. Renderer pins,
tolerances, masks and retained capture normalization stay unchanged.

## Baseline

| Measure | Value |
| --- | ---: |
| Source revision | `b5cd96adfca711a3e401d9826aafee008d8be592` |
| Committed goldens | 255 |
| Committed golden bytes | 18,398,270 |
| Timeline capture profiles | 146 (123 rich, 13 empty, 10 sparse) |
| Ordinary Visual run | `grr-p0-visual-baseline-20261114`: fail |
| Pilot row run | `grr-p0-coordination-row-baseline-20261114`: fail |

Byte-identical groups (four surplus copies by SHA-256). These are deferred to a
later batch and need consumer review before any decision:

- `collaboration-grid-saved-strip` and `collaboration-grid-recovered-saved-strip`
- `workbook-query-empty-density-compact` and
  `workbook-query-empty-successful-query`
- `workbook-query-saved-view-query-controls`,
  `workbook-view-bar-maximum-pressure-base` and
  `workbook-view-bar-saved-view-modified`

### Baseline failures at clean HEAD

| Row | Golden | Difference | Cause |
| --- | --- | --- | --- |
| `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | `account-menu-controls-narrow` | 7,813 px | New Original import sources menu item |
| `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `network-flow-analysis-mapping-dialog` | 40 px | Parser/adapter identifiers changed from `v1` to `v2` |
| `module.workbook.visual.coordination_create_authoring_recovery` | four `coordination-*-authoring` and `coordination-source-narrow` | 3,872 to 7,441 px each | Anchored inspector form content renders about 1 px higher |

The account-menu failure ended its browser group early. Its 36 remaining
scenarios did not run, so the baseline full run reported 157 unconsumed goldens.
That reconciliation is not valid orphan evidence.

## Refresh record

| Field | Value |
| --- | --- |
| Accepted trigger | The repository owner chose to accept the current rendered output instead of investigating the failures first |
| Update run | `grr-p0-visual-update-20261114` (`make browser-e2e-visual-update`) |
| Changed goldens | 49 retained goldens (50 including the later-retired comm-log golden) |
| Capture-contract changes | None. Viewport, zoom, masks, scroll normalization and screenshot scope are unchanged. |
| Fresh ordinary passes | `grr-p0-refresh-pass1-20261114` and `grr-p0-refresh-pass2-20261114`: both pass, reconciliation 255 of 255 active |

Affected rows, with changed-golden counts:

| Row | Goldens |
| --- | ---: |
| `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | 8 |
| `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | 8 |
| `module.workbook.visual.contextual_task_decision_creation` | 6 |
| `module.workbook.visual.coordination_create_authoring_recovery` | 5 |
| `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | 5 |
| `module.workbook.visual.note_create_authoring_recovery` | 3 |
| `module.workbook.visual.timeline_capture_actions` | 3 |
| `module.workbook.visual.timeline_related_evidence` | 3 |
| `module.workbook.visual.decision_supersession_review_recovery` | 2 |
| `module.workbook.visual.indicator_lifecycle_authoring` | 2 |
| `module.workbook.visual.indicator_observations_authoring` | 2 |
| `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | 1 |
| `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | 1 |
| `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | 1 |

Affected registered fixtures: `base_inspector`,
`claimed_network_analysis_compact_workspace`,
`claimed_network_analysis_narrow_workspace`,
`claimed_network_analysis_workspace_states`,
`contextual_coordination_creation`, `evidence_affordance`,
`indicator_lifecycle_authoring`, `indicator_observations_authoring`,
`inspector_compact_actions` and `mention_chip_state_matrix`. All carry the
`visual.fixture.` prefix.

Reviewer outcome by change class:

- **Inspector and authoring panels:** about 1 px vertical offset; no content
  change. Accepted.
- **Account menu:** a new Original import sources item shifts later items.
  Accepted as current product content.
- **Network Analysis header and Inspector Details:** a new Inspect original source
  button, and a new Import history row with a Find original import sources
  button. Both buttons are unstyled native `<button>` elements in
  `apps/web/src/imports/CommittedSourceContext.tsx` and render with light default
  chrome in the dark theme. The refresh accepts this as current output. It is a
  product styling follow-up, not a validated design.
- **Network Analysis mapping dialog:** parser identifier digits changed from
  `v1` to `v2`. Accepted.

## Pilot: contextual coordination creation

Fixture `visual.fixture.contextual_coordination_creation`, row
`module.workbook.visual.coordination_create_authoring_recovery`, scenario
"Capture contextual coordination target authoring source selection and retained
recovery". It has no `design_contract_id`, so no design amendment is needed.

All four authoring captures render `CoordinationCreateForm` at 1280x720 in
compact density. Each is a full-viewport capture, so only the top of the form is
compared. The form is field-contract driven, with branches for multiline text,
enum selects, direct references, `comm_log.audience`, `handoff.open_risk_refs`
and `lesson.closure_state`.

| Golden | Visible form treatments | Decision | Retained or replacement evidence |
| --- | --- | --- | --- |
| `coordination-comm-log-authoring` | Timestamp, filled required select, filled text inputs, Audience help text | Retired | Pixel: filled select and text input in `lesson-authoring`; timestamp and helper text in all retained authoring captures. Nonpixel: legend, field-control presence, Audience help text, focus and horizontal-fit assertions in the same scenario; all-field keyboard reachability and visible focus in `coordination.a11y.spec.ts` |
| `coordination-handoff-authoring` | Empty owner chooser, selected reference with Remove, required-reference message | Keep | Only selected-reference treatment in this form; now the fixture's `golden_filename` |
| `coordination-status-review-authoring` | Empty owner chooser, filled multiline textarea, empty collection | Keep | Only filled textarea in this form |
| `coordination-lesson-authoring` | Filled text input, owner chooser, `closure_state` default option | Keep | Unique default-option branch; precedes lesson-based captures |
| `coordination-source-narrow` | Source picker at 768x640 | Keep | Distinct width and state |
| `coordination-recovery-narrow` | Uncertain recovery at 390x480 | Keep | Distinct width and state |
| `coordination-recovery` | Uncertain recovery at 1280x720 | Keep | Distinct state |

Authored changes:

- `apps/web/e2e/coordination.visual.spec.ts`: skips the comm-log pixel capture.
  Still drives the full comm-log fixture and asserts its obligations.
- `tools/frontend_visual_fixture_registry.json`: removes the path and profile;
  `golden_filename` now points to the handoff golden.
- `tools/frontend_visual_normalization.json`: removes the capture declaration and
  binding.
- `tools/frontend_visual_timeline_profiles.json`: removes the rich profile. The
  profiles are now 145: 122 rich, 13 empty, 10 sparse.

The PNG was deleted and the manifest regenerated with `make generate`.

Defect sensitivity (disposable product perturbations, reverted; `apps/web/src`
matches HEAD):

| Perturbation | Run | Result |
| --- | --- | --- |
| Audience input `minWidth: 900` | `grr-p2-sensitivity-overflow-20261114` | Fails `contentFits` |
| Audience help text removed | `grr-p2-sensitivity-helptext-20261114` | Fails help-text visibility |

## Results

| Measure | Before | After |
| --- | ---: | ---: |
| Committed goldens | 255 | 254 |
| Committed golden bytes | 18,398,270 | 18,273,843 |
| Pilot fixture goldens | 7 | 6 |
| Full ordinary Visual wall time | 435,514 ms (refresh pass 1) | 440,886 ms |

The pilot proves the lifecycle; it does not reduce runtime. The comm-log fixture
flow still runs, and the wall-time difference is within run-to-run variation.

| Evidence | Run | Outcome |
| --- | --- | --- |
| Pilot row with replacements | `grr-p2-coordination-row-b-20261114` | Pass; comm-log golden is the only coordination orphan |
| Orphan check | `grr-p2-orphan-check-20261114` | 48 of 48 rows pass; exactly one orphan with zero consumers and zero declared refs |
| Final ordinary Visual | `grr-p2-final-visual-20261114` | Pass; 48 of 48 rows; reconciliation 254 of 254 active, 0 orphan, missing or ambiguous |

Reconciliation v4 emits `permitted_action: blocked` for every nonactive golden.
Testing Harness NLSpec allows a zero-consumer orphan to proceed to an authorized
deletion slice. This pilot was that slice.

## Next candidates

- Contextual Task and Decision family (10 goldens): authoring, references and
  recovery at two widths. Evaluate the same way.
- The byte-identical groups above.
- Fix the unstyled import-source buttons, then refresh their goldens under a
  normal design-change trigger.
