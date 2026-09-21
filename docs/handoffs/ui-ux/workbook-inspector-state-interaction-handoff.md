# Workbook inspector state and interaction remediation

## Baseline and authority

User-approved implementation on `codex/workbook-inspector-state-interaction`,
from clean `main` at `fb36c6a52de8d38d603b021dda1746f5c9560c91`.
Core 03 §2.3A owns interaction and retained lifetimes; Core 01 §§7.4–7.4.1A
owns declared panels, capabilities and routes; Core 02 owns source semantics;
Core 04 owns authorization and acceptance. Design §§7.3, 12.4–12.8 and 14.2
own state composition, local feedback, saved-first Details and relationship
grouping. The prior inspector remediation handoff remains historical evidence.

## Workstream ledger

| Workstream | Status | Exit |
| --- | --- | --- |
| W0 Owners and projections | PASS | Adopted behavior, generated presentation, contribution inventory and traceability agree. |
| W1 Truthful regions and recovery | PASS | All four production families preserve independent read states, context, local recovery and concealed content. |
| W2 Saved-first Details | PASS | Generic, Entity and Timeline use explicit field editing with retained drafts and no blur submission. |
| W3 Relationship grouping | PASS | Hosts, Identities and Tags retain semantic commands and local context without duplicated groups. |
| W4 Verification and handoff | Product gates PASS; handoff delivered with sequencing deviation | Focused/browser checks, two fresh visual passes, retirement and review evidence complete; strict historical phase ordering was not met. |

The implementation order follows owner closure, shared presentation, ordinary
Details and grouped relationships. Some implementation and validation overlapped;
this ledger does not claim that every earlier workstream had its complete exit
before dependent code was authored. All final product exits now have current passing evidence. This does not
retroactively satisfy the strict historical phase-ordering criterion in A027.

## Remediation and ownership decisions

| Gap | Adopted fix and changed areas | Rationale and durable benefit | Compatibility and unresolved-risk control | Completion evidence |
| --- | --- | --- | --- | --- |
| Competing presentation authorities | Core 03 §2.3A, Core 01 navigation, Core 04 AC-456, Design §§8.4–8.5, 12.7 and 14.2; authored design projection and schema | One behavior owner; machine consumers receive closed presentation facts without Markdown dependencies | Existing public schema identities and routes preserved; drift checks prevent projection divergence | Generation, projection tests and final drift checks |
| Fabricated ready/empty states | Nonempty owner regions; History, Notes, Evidence, Entity preview and specialized Indicator read adapters | Independent data and recovery remain cohesive with their existing readers; no duplicate request/cache owner | Five states retained; neutral not-requested History; previously empty accepted pages survive failed refresh as stale observations | Production composition, independent Notes retry, History, Evidence and Indicator regressions |
| Cross-subject or duplicate feedback | Captured semantic destinations and owner-retained announcement ledgers; local field/relationship/read outcomes | Attempt and transition identity survives presentation detachment; a new failed recovery is distinguishable from an old notice | Original receipt and replay owners remain authoritative; ordinary field errors retain revision fencing | Feedback identity, ledger remount/retarget, exact replay and read-only recovery checks |
| Hidden access loss and fresh mutation admission | Concealed variants have no payload; all family facades withdraw protected headers on absent incident authority; file/drop/paste admission is owner checked | Read availability and mutation eligibility remain separate | Read-only/closed reading remains; no new replay permission is inferred | Authorization, lifecycle, Evidence and browser permission scenarios |
| Editing before reading | Contract-ordered saved field overview and explicit one-field attachment | Accepted values remain independent of local typing; long values wrap; null, missing, empty text, zero and false remain distinguishable | Assessment append forms remain specialized; field selectors retire in favor of named Edit actions | Entity and Timeline production tests; browser subject retention and accessible narrow-layout recovery |
| Timeline implicit scalar writes | Timeline-owned source-preserving PATCH contribution using ordinary drafts; Update/Ctrl/Cmd+Enter only | One ordinary inspector lifetime and submission model; source text is never title-normalized | Intentional removal of inspector blur/paste departure saves; grid autosave and collection token semantics stay separately owned | Native clipboard, draft continuity, explicit request count and source-write sequencing checks |
| Disconnected relationship controls | Field-keyed Hosts, Identities and Tags contributions; selected details and outcomes inside each group | Stable source/field/item identity guides controls, focus and recovery without another editor | Raw mentions, resolved links, target lookup failures and session-only dismissed observations retain their domain meaning | Mention, collection overflow, stateful and accessibility regressions; reviewed visuals |
| Duplicate or obsolete alternatives | Removed scalar inspector renderer/binding list, monolithic relationship editor slot, duplicate headings/prompts and bottom-channel local notices | Fewer independent presentation and submission paths to maintain | Shared Timeline registry still serves grid drafts and inspector collection authoring; it is not used by ordinary scalar Details | Source/caller review, type/import checks and regression routing |
| Unsupported readiness claims | This controlling handoff records commands, failures, visual reconciliation and acceptance dispositions | Reproducible review and migration evidence replaces historical-pass assumptions | No durable-draft, reload-persistence, endpoint, dependency or storage claim | Final A001–A027 assessment and visual gate |

Private subscribed-region callbacks deliver owner models directly to the shared
renderer without moving read lifetimes. Owner components retain typed operation
commands, admission and disabled explanations; the shared renderer neither
constructs routes nor authorizes operations. Ordinary explicit recovery uses a
presentation lease only to avoid displaying the same outcome twice; leases are
not a mutation cache and cannot cancel or acknowledge an attempt.

## Contribution inventory

Every supported schema is selected through its canonical inspector config.
Executable coverage uses those machine declarations, not this table.

| Family / schemas | Panel / region | Data and command owner | States and recovery | Semantic selector |
| --- | --- | --- | --- | --- |
| Generic, all registered generic schemas | Details / saved fields, ordinary editor | Canonical row; Inspector drafts and explicit PATCH | Accepted fields; independent authoring/review and read-only recovery | schema, record, field |
| Generic / Notes | Relationships / sources, related notes; Evidence / evidence associations | Note association owner | Loading, ready empty/populated, refreshing, stale, unavailable, concealed; exact owner read/replay | schema, record, association kind |
| Generic / other schemas | Relationships and Evidence / declared reference summaries, Party and Evidence contributions | Reference projection; Party/Evidence owner | Accepted row or owner read/access state; owner commands only | schema, record, field/feature |
| Entity / Hosts, Identities | Details / saved fields, aliases, reusable identifiers | Canonical entity row and ordinary/alias owners | Accepted values and retained authoring; no preview-derived emptiness | schema, record, field/item |
| Entity | Relationships / Timeline preview, merge; Evidence / projected count | Preview reader and merge owner; row projection | Scoped preview loading/ready/stale/unavailable; captured merge review | schema, record, panel |
| Timeline | Details / saved fields, ordinary editor | Canonical row; Inspector drafts and explicit PATCH | Accepted fields plus retained explicit authoring | schema, record, field |
| Timeline | Relationships / host refs, identity refs, tags, observations | Collection/mention and Indicator owners | Source-backed membership, observed dismissals, independent observation reads | record, field, item/feature |
| Timeline | Evidence / metadata, attachment recovery | Row projection and file owner | Accepted count and independent captured file operations | record, evidence/file identity |
| All record families | History / browsing and corrective action | Retained History owner | Not requested, loading, ready, refreshing, stale, unavailable, concealed; owner continuation/recovery | schema, record, history action |
| All families | Workflow / declared actions and attached creation | Source-specific creation owners | Admitted action catalogue and separately retained authoring/outcomes | schema, feature, attachment |
| Assessment | Details / accepted assessment; Relationships / support; Workflow / creation | Assessment owner | Read-only accepted values; explicit standalone/follow-on creation context | schema, record or creation attachment |

## Compatibility, retirement and risks

No routes, public discovery, source semantics, storage or durable-draft migration.
Timeline inspector scalar blur-save intentionally becomes explicit Update or
Ctrl/Cmd+Enter. Grid autosave, token departure and specialized workflows retain
their distinct owners. Retire opaque success wrappers, misplaced notices,
ordinary field selectors and the superseded inspector scalar queue path during
their cutovers. No legacy preference or permanent dual renderer is retained.

Shared presentation owns rendering and semantic destinations, not requests,
authorization, source writes, receipts or a second accepted-data cache. Principal
risks are duplicated reads/announcements, cross-subject outcomes, dropped newer
authoring and accidental replay after acknowledgement. Production tests must
observe request counts and source identities, not only displayed messages.

## Verification ledger

Final product verification passed. Current task guides for `web.workbook`, `module.workbook`,
`web.design`, `package.ui` and architecture checks supplied routing.
`make agent-finalize` ran before broad final verification. Retained-run
maintenance was skipped because no eligible full warm `RESULTS_DIR` was supplied.
Visual changes follow the visual golden maintenance guide, including ordinary
reconciliation, manual candidate review and two fresh ordinary comparison passes.
The acceptance assessment below uses current implementation evidence and
identifies the historical process deviation separately from passing product gates.

Focused evidence obtained during implementation:

- Initial region/feedback slice: `.cartulary/test-results/20260921T013624Z-p84772`,
  8/8 units passed.
- First complete `web.workbook` slice:
  `.cartulary/test-results/20260921T020351Z-p12442`, 286/288 units passed.
  The old scalar Escape assertion and concealed History alert assertion were
  migrated to the adopted behavior; their focused rerun passed.
- Timeline production Details tests:
  `.cartulary/test-results/20260921T021437Z-p93317`, 2/2 units passed.
  They caught and drove repair of accidental use of title normalization for
  `timeline_visible_text_v1`.
- Inspector region/notice/Details/recovery slice:
  `.cartulary/test-results/20260921T022136Z-p11206`, 5/5 units passed.
- Module production inspector composition slice:
  `.cartulary/test-results/20260921T021600Z-p98427`, 5/5 units passed.
- First browser interaction slice:
  `.cartulary/test-results/20260921T021259Z-p58829`; all selected scenarios except
  Entity Find passed. Entity Find needed to open the new Manage aliases disclosure.
- Browser, accessibility and stateful inspector rerun:
  `.cartulary/test-results/20260921T021652Z-p99587`, 19/19 units passed.
- Clipboard first run: `.cartulary/test-results/20260921T021842Z-p41551`;
  continuity/composition passed, characterization/departure caught Activity
  Synopsis incorrectly using a single-line control. The form control now follows
  Timeline's source-preserving multiline contract. Characterization/departure
  rerun passed 11/11 units at `.cartulary/test-results/20260921T022726Z-p73723`.
- Type and import checks passed at
  `.cartulary/test-results/20260921T022516Z-p36352` and
  `.cartulary/test-results/20260921T022516Z-p36359`.
- `make agent-finalize` passed at
  `.cartulary/test-results/20260921T022124Z-p8366`.
  Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
- Focused ordinary visual comparison:
  `.cartulary/test-results/20260921T021843Z-p41776`; intended inspector changes
  differed from old goldens. Visual review also restored History controls above
  entries and removed duplicate Workflow prompts. No golden was promoted.
- The first full ordinary visual attempt
  `.cartulary/test-results/20260921T022204Z-p13492` stopped at build-artifact
  validation because source changed after its input snapshot. No product or
  visual result is inferred; subsequent builds freeze frontend source until
  their artifact receipt is complete.

## Final verification records

All commands run from the repository root with the pinned local Node runtime
on `PATH`. Unit counts below are harness execution units, not individual assertions.
The retained `run-manifest.json` records exact `OWNER`, `ROWS` and worker inputs;
`run-summary.json` and the target/group results record terminal status.

| Command / scope | Result | Run root beneath `.cartulary/test-results/` |
| --- | --- | --- |
| `make test-slice OWNER=web.workbook` | PASS 289/289 | `20260921T022516Z-p36294` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_timeline_explicit_details,web.workbook.regression.inspector_ordinary_recovery,web.workbook.regression.explicit_task_patch_recovery,web.workbook.regression.inspector_explicit_panel_states` | PASS 5/5 after final recovery change | `20260921T023659Z-p6219` |
| `make test-slice OWNER=module.workbook ROWS=…` production inspector compositions | PASS 5/5; exact row selection in retained manifest | `20260921T021600Z-p98427` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=…` | Production editing, local feedback, retained subject, replay, acknowledgement, collections and Find; one obsolete Entity alias entry assertion fixed and rerun | `20260921T021259Z-p58829` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=…` | PASS 19/19: Entity Find, inspector accessibility and stateful coverage | `20260921T021652Z-p99587` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=…` | Clipboard composition/continuity passed in first run; characterization/departure passed 11/11 on rerun | `20260921T021842Z-p41551`, `20260921T022726Z-p73723` |
| `make test-slice OWNER=web.design` | PASS 19/19, including global accessibility and responsive frame | `20260921T023659Z-p6241` |
| `make test-slice OWNER=package.ui` | PASS 10/10 | `20260921T022939Z-p41663` |
| `make frontend-typecheck` | PASS 2/2 | `20260921T023712Z-p41775` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260921T023712Z-p41816` |
| `make lint-biome` | PASS 2/2 | `20260921T023712Z-p42012` |
| `make generate` | PASS, authored presentation/catalog inputs regenerated | `20260921T023127Z-p86538` |
| `make generate-drift` | PASS 4/4 | `20260921T024410Z-p53472` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260921T024322Z-p87880` |
| `make json-shape-check` | PASS 3/3 | `20260921T024322Z-p87844` |
| `make lint-markdown` | PASS after final review/acceptance text | `20260921T024934Z-p65522` |
| `make browser-e2e-visual` (two fresh post-promotion runs) | PASS 12/12 each | `20260921T024322Z-p88150`, `20260921T024322Z-p88151` |
| `make agent-finalize` | PASS 1/1 before final broad checks; retained-run maintenance skipped because `RESULTS_DIR` was unset | `20260921T023553Z-p96829` |

Additional failures resolved during final review:

- `web.design` run `20260921T022939Z-p41627` failed an old test that tried
  to focus an ordinary editor before invoking Edit. The migrated assertion now
  verifies Edit focus, attachment, Escape return to Edit, and subsequent close;
  the entire design slice passed afterward.
- Timeline recovery test run `20260921T023538Z-p96032` incorrectly expected
  Dismiss after a replay rejection. The existing operation owner correctly keeps
  the original uncertain request: a later rejection cannot prove that the first
  request failed. The corrected test verifies exact request object/bytes,
  retained source text, local replay control and one new failure announcement.
- Final source and dependency review found no Markdown dependency in added tests,
  generators or runtime. Production source reads typed contracts; this handoff and
  the digest remain human review material.

Full backend/release suites and unrelated module suites were not run: no backend,
wire, database, dependency or release behavior changed. The selected real-service
browser checks cover the affected frontend/operation boundaries. No reload
persistence or future Jump to section experiment is claimed.

The exact selected production/browser invocations referenced above were:

```sh
make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_active_view_schema_id_selects_inspector_c_82f9499272,module.workbook.frontend_unit.verify_active_view_schema_id_selects_inspector_c_fed994e037,module.workbook.frontend_unit.verify_inspector_panels_and_feature_groups_rende_b947f0007c,module.workbook.frontend_unit.verify_inspector_selection_tab_state_details_rel_d2dc82a4bb
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_collection_feedback,module.workbook.browser.inspector_detached_refresh,module.workbook.browser.inspector_exact_replay,module.workbook.browser.inspector_local_feedback,module.workbook.browser.inspector_persistent_header,module.workbook.browser.inspector_subject_retention,module.workbook.browser.grid_autosave_inspector,module.workbook.browser.entity_find_inspector,module.workbook.browser.find_source_editors
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.entity_find_inspector,module.workbook.accessibility.inspector_edit_recovery,module.workbook.accessibility.verify_inspector_tabs_relationship_links_evidenc_9ee9fd9ea2,module.workbook.accessibility.verify_keyboard_open_close_panel_navigation_esc_42b98cf08e,module.workbook.browser_stateful.verify_default_closed_inspector_state_no_row_sta_897d604d9e,module.workbook.browser_stateful.verify_timeline_inspector_workflow_create_relate_7fc4833af4
make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.scalar_clipboard_characterization,module.timeline.browser.scalar_clipboard_composition,module.timeline.browser.scalar_clipboard_continuity,module.timeline.browser.scalar_clipboard_departure
make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.scalar_clipboard_characterization,module.timeline.browser.scalar_clipboard_departure
```

## Visual refresh record

Accepted trigger: adopted saved-first Details, grouped relationships, independent
read states, neutral unloaded History, and removal of duplicated Workflow prompts.
No viewport, browser zoom, density, masks, scroll normalization, screenshot scope,
renderer/font pin or comparison tolerance changed. The two source assertion edits
replace the retired operational-text/editor surface with saved-field inspection;
all other functional assertions remain in force.

The full ordinary comparison at `20260921T022725Z-p73596` failed only on the
expected old-golden comparisons. Its reconciliation accounted for all 252 active
capture intents and committed PNGs, 29 registered fixtures, and no orphan,
missing, ambiguous or unresolved mapping. The initial focused reconciliation was
not used to justify mutation of unselected images.

`make browser-e2e-visual-update` passed 12/12 units at
`.cartulary/test-results/20260921T023659Z-p6477`. Candidate promotion retained
222 passing goldens and updated the 30 listed below plus the tool-owned golden
manifest. Every changed promoted image was opened and visually inspected.
Review confirmed intended grouping/state/spacing changes, readable wrapped
content, preserved visible focus, header/close access and local controls across
desktop/narrow fixtures. Content outside each declared scrollport remains
scrollable; capture anchors and crops were preserved. No unexplained typography,
theme or overflow change was accepted.

The exact owner/fixture mappings below come from that run's reconciliation,
not filename inference. Nonregistry active captures remain valid catalog-owned
consumers. Two fresh `make browser-e2e-visual` runs passed 12/12 units each at
`.cartulary/test-results/20260921T024322Z-p88150` and
`.cartulary/test-results/20260921T024322Z-p88151`. Both reconcile all 252 captures
against the promoted manifest, with no missing or ambiguous mapping. They ran
independently against the same frozen frontend source and golden inputs.

| Golden filename | Catalog owner row | Stable fixture ID |
| --- | --- | --- |
| `coordination-comm-log-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-handoff-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-lesson-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-source-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-status-review-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `decision-supersession-review-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | Active nonregistry capture |
| `decision-supersession-review-narrow-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | Active nonregistry capture |
| `entity-mention-chip-states-linux.png` | `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `visual.fixture.mention_chip_state_matrix` |
| `evidence-affordance-states-linux.png` | `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `visual.fixture.evidence_affordance` |
| `indicator-lifecycle-authoring-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-lifecycle-authoring-narrow-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-observation-authoring-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `indicator-observation-authoring-narrow-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `linked-note-authoring-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | Active nonregistry capture |
| `linked-note-authoring-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | Active nonregistry capture |
| `linked-note-source-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | Active nonregistry capture |
| `record-relationships-mention-chips-linux.png` | `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | Active nonregistry capture |
| `timeline-related-evidence-authoring-linux.png` | `module.workbook.visual.timeline_related_evidence` | Active nonregistry capture |
| `timeline-related-evidence-authoring-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | Active nonregistry capture |
| `timeline-related-evidence-party-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | Active nonregistry capture |
| `timeline-supersession-authoring-linux.png` | `module.workbook.visual.timeline_capture_actions` | Active nonregistry capture |
| `timeline-supersession-review-linux.png` | `module.workbook.visual.timeline_capture_actions` | Active nonregistry capture |
| `timeline-supersession-review-narrow-linux.png` | `module.workbook.visual.timeline_capture_actions` | Active nonregistry capture |
| `workbook-inspector-compact-actions-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_compact_actions` |
| `workbook-inspector-destructive-confirmation-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.destructive_actions` |
| `workbook-inspector-history-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-public-error-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-relationships-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |

## Digest acceptance assessment

The digest supplies review criteria, not executable product authority. Evidence
references here resolve to the verification and visual records above. No advisory
TSV was edited or used as a test input. No row is excluded as irrelevant merely
because its implementation owner was unchanged.

| Row | Status | Scope and current evidence |
| --- | --- | --- |
| A001 | PASS | Exact Core/design clauses and independent source/verification owners mapped above; attachment and historical handoffs remained supporting material. |
| A002 | PASS | Shared decision is known data/access/authoring/outcome presentation. Source-specific read/mutation ownership remains separate. Future independent regions can use the same union without expanding a request registry. Scalar inspector and monolithic relationship alternatives retired. |
| A003 | PASS | Clean baseline, current branch, pnpm/React/TypeScript/Go manifests, source guides, authored ownership and generated boundaries inspected; final import-boundary check passes. Current source/catalog overrides digest snapshot navigation. |
| A004 | PASS | New presentation uses existing spacing, ink, surface, border and radius tokens. No palette, density or theme registry added; package.ui token/projection checks pass. |
| A005 | PASS | Current dark_graphite theme retained; full visual fixture inventory includes the existing theme/token specimen. No theme or font changes in diff. |
| A006 | PASS | Shared density owner retained. Existing compact/base visual fixtures and responsive design/browser checks exercise inspector controls and saved/read-only values; no component-local density selection added. |
| A007 | PASS | Required creation contexts are explicit, including Assessment without a fabricated record. Assessment, contextual creation and capability rows pass in workbook/design suites; specialized payload/authoring owners remain intact. |
| A008 | PASS | Responsive controller/accessors unchanged; web.design frame geometry and inspector accessibility pass. Full visual inventory retains desktop, narrow and constrained fixtures with declared dimensions and focus assertions. |
| A009 | PASS | Inspector body and grid retain their independent scrollports; header, close, status strip and application navigation remain reachable in design/browser checks and manual image review. |
| A010 | PASS | Machine-configured declaration/order, capability routes and disabled reasons retained; required missing contributions fail. Production module composition, role/state, subject, review and detachment regressions pass. |
| A011 | PASS | Subject-retention, detached-refresh, Find and native clipboard browser scenarios pass. Saved-row refresh retains caret/selection; source readiness sees ordinary inspector drafts; late outcomes stay owner/capture-bound. |
| A012 | PASS | Existing transaction-ID factory retained. Explicit recovery tests and browser exact replay observe original request identity/bytes, duplicate admission fencing and late receipt handling. |
| A013 | PASS | Existing queue retry policy unchanged. Ordinary PATCH recovery tests prove acknowledgement survives detachment/new authoring and repeated failed-refresh recovery sends no additional write. Rejected uncertain replay remains uncertain under its existing owner policy. |
| A014 | PASS | Production ordinary Details, retained draft/field feedback, browser editing and native clipboard scenarios pass. Timeline inspector deliberately uses explicit Update/Ctrl/Cmd+Enter; grid and collection behavior stays independent. |
| A015 | PASS | Existing cell conflict owner retained; explicit PATCH source coordination and captured-revision tests pass. Saved overview and unsaved editor are separate; local action/field feedback remains reachable. |
| A016 | PASS | Five-state typed projection and explicit access variants exercised through production Notes, History, Evidence, Entity and specialized reads. Empty accepted observations, stale siblings and separate mutation admission tested. |
| A017 | PASS | Workbook authority/draft lifetime and role regressions pass; protected family headers/regions withdraw on read loss. Same-record refresh differs from retargeting; readonly/closed values remain readable. Account/session lifecycle owners remain unchanged. |
| A018 | PASS | Evidence owner regressions and visual matrix retain lifecycle/overlay/handle distinctions; metadata and handle access are separate regions. File selection/drop/paste callbacks check current admission. |
| A019 | PASS | Module inspector accessibility/stateful checks and complete web.design slice pass. Notice ledger tests cover remount, retarget and a new failure transition; explicit replay regression has one new assertive announcement. No new conformance profile claimed. |
| A020 | PASS | Existing component variants, long content, text spacing, zoom, overflow and density fixtures retain their assertions and parameters. Every changed golden reviewed; ordinary editor uses existing form controls and safe wrapping. |
| A021 | PASS | Virtualization and row realization are unchanged. Production grid/autosave, clipboard continuity, Find and stateful inspector checks preserve semantic source/focus across edits and navigation; no synthetic rows or second grid cache introduced. Unrelated throughput benchmarks were not rerun. |
| A022 | PASS | Full ordinary reconciliation and all 30 changed promoted images reviewed; both fresh ordinary runs pass 12/12. Visual artifacts remain implementation-support evidence. |
| A023 | PASS | Stable schema/record/field/item selectors retained; new field entry/attachment selectors use field keys. Named Edit actions are accessible controls. All package.ui selector checks pass. |
| A024 | PASS | Added runtime/tests/generator inputs use machine-owned JSON/TypeScript. Diff/dependency review found no Markdown reads, stats, hashes or prose-bound behavior. Documentation lint is separate. |
| A025 | PASS | Authored presentation/schema/generator and source/test ownership inputs precede generated facades/topology. Post-promotion JSON, generated-policy and drift checks pass; visual manifest changed only through its update target. |
| A026 | PASS | No public schema identity, route, storage, dependency or persistence migration. Explicit Timeline save behavior is the authorized correction. Existing route-specific replay, source semantics and collection consumers justify retained shared owners. Rollback is a coherent frontend revert. |
| A027 | NOT MET for historical phase ordering; handoff delivered | All requested handoff material and final product evidence are present. The strict requirement to finish each earlier exit before dependent implementation was not met; no blanket claim of full skill-process compliance is made. |

Process deviation: some dependent implementation was authored while earlier
workstream validation was still running. The skill's strict requirement to finish
and record each exit before starting its dependent was not met retroactively.
All implementation exits are re-evaluated against current evidence before final
handoff; this deviation does not create an alternate runtime path or justify
omitting a product gate.

## Deferred navigation

Jump to section stays absent. A later prototype must use the existing menu,
at least three readable declared sections, canonical order, one heading target,
and only the inspector body scrollport. No persistence or alternate controller.
Gate: counterbalanced evaluation with eight representative users at typical and
minimum widths; History and Workflow median acquisition each improves at least
20%, errors do not increase, Details/Close medians regress at most 10%, and all
keyboard, access, authoring and scroll checks pass. These are product decision
thresholds, not research claims; inconclusive results retain the simpler shell.

## Rollback and advisory disposition

Revert coherent presentation/interaction cutovers without changing committed
records, receipts or source history. Retained authoring has no reload guarantee.
R002/R006/R008 focus, local feedback and owner recovery are ADOPT; R012 state
placement is ADAPT to retained owner lifetimes. No new framework, design system,
universal retry engine or navigation chrome is introduced.

Delivery state: changes remain uncommitted on the implementation branch. The next
action is normal code review and a coherent commit of specs, authored/generated
contracts, source, tests, reviewed goldens and this handoff. No product validation
gate remains pending; the historical A027 sequencing deviation remains disclosed.
