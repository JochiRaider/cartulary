# Timeline file-gesture targeting cleanup

Date: 2026-09-23. Scope: the authorized Timeline file-admission correction,
chooser invocation lifetime, and the original-source availability gate before
first upload. Work is on `main` at
`1597f2e0da7ecfb22adf666f509b2211b2f1ba9e`. The inspected baseline was clean;
all final dirty paths belong to this slice. No user changes were overwritten.
No commit, push, deployment, dependencies, endpoints, persistence, or migrations.

## Owners and navigation

The September 13 digest at `59fa79e04035a4f29251fcb2337376f29c40f1f4`
is advisory navigation. Its START_HERE, LOCAL_AGENT_PROMPT, REPO_MAP, OWNER_MAP,
rules, acceptance, and QUERY_RECIPES were consulted with AGENTS.md and the
Cartulary UI/UX skill. Historical handoffs are regression context, not fresh
results. Current authored inputs and source guides were revalidated. React
19.2.5, TypeScript 6.0.2, Vite 8.0.8, Playwright 1.59.1, and Vitest 4.1.4 remain
as declared in `apps/web/package.json`; no stack change. Direct grid-vendor
integration remains inside `packages/grid-adapter`.

| Concern | Behavioral owner | Implementation owner | Independent verification owner |
| --- | --- | --- | --- |
| Gesture target and retained original | Core 03 §8.1, REQ-03-116–119 and adjoining retention/admission clauses; bounded requested clarification recorded there | Source owner `web.workbook`: Timeline grid, pure attachment-plan resolver, editor registry, explicit-source command | `module.evidence` browser rows; `module.timeline` resolver row; `web.workbook` keyboard/surface rows |
| Draft creation and promotion | Core 03 §7 REQ-03-111–115, §8.1 shared-draft ordering | Existing TimelineCaptureLifecycle and mutation runtime; DraftRowActions captures the existing draft key | `module.evidence` retained draft unit and production ordering rows |
| Clipboard and native editing | Core 03 §11.1 REQ-03-145–147, §13; design §§8,12.5,14 | Timeline grid capture handler; existing scalar/collection editor registrations and external-action focus convention | New production paste/composition rows; `web.workbook` keyboard regression |
| Upload, finalization, association, replay and receipts | Core 01 §§3.3.5,3.3.8; Core 03 §8.1 | Existing WorkbookTimelineFileOwner, EvidenceUploadSession, EvidenceFileFinalization, Timeline mutation owner | `module.evidence` retained_file_recovery, timeline_file_draft_recovery, file_stage_recovery, existing_file_recovery |
| Authority and suspension | Core 04 §§1–2,4; Core 03 REQ-03-100/299 | Existing incident/account runtime and file-owner rechecks; grid interaction mode remains a separate gate | Retained owner recovery units and shared Evidence binding regression |
| Feedback and Evidence rendering | Design §§8,11,12.5,14,15 within Core behavior | Existing file recovery and form/grid tokens; shared EvidenceAttachmentEntry | `module.evidence.accessibility.file_recovery` and four existing visual rows |

Source assignment is authored in `tools/frontend_source_ownership.json` and
imports in `tools/frontend_import_boundaries.json`. Verification routing is
separately authored in `tools/test_families/module.evidence.json` and
`module.timeline.json`, downstream of the verification contracts/catalog.
These artifacts do not establish product requirements. `docs/domain.md` remains
vocabulary/navigation; `docs/design.md` remains design direction.

`tools/generated_artifact_policy.json` was the exhaustive generated registry.
Only repository generation rewrote `tools/browser_e2e_batch_manifest.json` and
`tools/execution_topology_render_index.json`; authored routing preceded it.
No generated View Contracts, dependencies, or lockfiles were edited.

## Characterization and selection rubric

The production fixture uses separate committed A, B and C, an independently
open Inspector C, and the retained trailing draft. A semantic gridcell is
activated using the grid's range gesture without selecting Inspector C.
Association assertions query persisted Timeline rows; request observers count
slot creation, transfer, Evidence creation, Timeline creation and record writes.

| Gap and classification | Evidence before production changes | Remediation, rationale and exit |
| --- | --- | --- |
| **Confirmed defect:** drop on B used active A | Production run `20260923T162235Z-p56001`; persisted association went to A | Semantic B identity now precedes the active grid target. PASS requires one association to B and none to A/C for drop and file paste. |
| **Confirmed defect:** scalar draft drop and collection draft drop/paste inherited committed context | `20260923T162453Z-p89886`; scalar draft paste already passed, the other three variants associated a committed row | Resolve registered editor/draft identity uniformly; retain its lifecycle key. PASS requires one new intended Timeline row for each variant. |
| **Confirmed defect:** detached chooser completion was lost; replacing draft context could retarget | Production detached-picker run `20260923T162235Z-p56001`; chooser units `20260923T161924Z-p88347` and `20260923T162258Z-p85785` | Capture callback/source at invocation and retain native completion/cancel listeners through React detachment. PASS requires the original committed/draft identity, no operation on cancellation. |
| **Confirmed defect:** unavailable original still created an upload slot | New owner unit failed at `20260923T162021Z-p21377`; prior recovery tests in that run passed | Use existing incident-bound reader/draft port before first slot creation. No editor settlement at this gate. PASS requires zero upload/finalization/coordination for unavailable, superseded or unverifiable originals. |
| **Structural weakness:** optional scalar-only identity, remembered anchor, Inspector fallback, and dead selected-target planner formed competing admission paths | Source inspection; not a claim of every possible browser outcome | One small pure resolver returns resolved/missing/unavailable/ambiguous. Grid owns semantic event extraction and one scoped active-target reference. Callers hand off a required TimelineFileSource. |
| **Hypotheses investigated:** incidental blur, raw text/composition damage, filtering/virtualization retargeting, background ambiguity | Not inferred from historical handoffs or source shape | Held-upload production checks examine native text, selection, composition, source unmount/filter and local rejection. Ambiguous duplicate identities are tested at the pure boundary because valid production grid presentation exposes one semantic active cell. |

This is one coherent admission boundary, not a new upload framework. The common
decision is whether a gesture names exactly one eligible Timeline source and
which immutable snapshot to hand off. This improves correctness and makes future
Timeline entry points reuse a small source contract without another selection
fallback or attachment store. A future entry point can pass an existing semantic
row/draft identity; no speculative extension was implemented.

Capability value and retention are explicit: single-file admission, zero-byte
eligibility, native clipboard editing, keyboard chooser access, retained original
review, exact replay, independent receipts and read-only refresh recovery remain
because their adopted owners require them. Leaving the confirmed defects would
associate evidence with the wrong record; leaving the competing paths would
make future controls inherit the same maintenance risk.

The authorized behavior correction is separate from structural movement:
Core 03 §8.1 now specifies local/background identities, chooser capture and
initial availability. Upload and recovery behavior is not redefined. Source
movement is confined to the existing Timeline resolver, required callback
signature and shared picker hook. There is no second queue or retained attachment
store, no bulk upload and no new confirmation step.

## Cutover and retained paths

- `TimelineWorkbookGrid.tsx` resolves registered scalar/collection editors and
  semantic grid record/draft attributes. Known headers/groups are unavailable.
  Row-local identity wins; background uses only the current scoped active grid
  identity, checked against presentation membership. Missing/ambiguous/unavailable
  targets report “Select the Timeline row or draft for this file.” locally.
- `timelineEvidenceAttachmentPlan.ts` replaces the unused selected-target planner
  with explicit source resolution, accepted draft aliases and snapshot capture.
  Inspector selection, bulk selection, text, indexes and vendor coordinates are
  absent from resolver inputs.
- Presentation, Inspector panel/sections, draft actions and
  `useTimelineEvidenceAttach.ts` now use required captured TimelineFileSource.
  Retired: `fileAnchor`, `selectedRowId` attachment fallback, optional
  `editorRowKey`, scalar-only `data-timeline-file-source`, unused planner exports
  and the unused `attachEvidenceFileToTimeline` command alias. The remaining
  selectedRowKey action context names presentation for existing review invalidation;
  it does not select a file target.
- `EvidenceAttachmentEntry.tsx` and DraftRowActions capture invocation before the
  native chooser opens. Completion belongs to that invocation even if its input
  detaches; cancel/change clears listeners. Completion without invocation is not
  admission. External-action markers and mousedown focus borrowing avoid unrelated
  blur submission, while Enter/Space remain supported.
- `WorkbookTimelineFileOwner.ts` checks original availability before first upload
  through its existing unfiltered reader/draft port. It checks generation and
  authority after the read. Original-source write coordination and review remain
  at pre-link, including existing review after presentation detachment. Accepted
  draft promotion remains in TimelineCaptureLifecycle; no visible-row lookup
  replaces the retained source.
- Shared Evidence picker consumers and browser fixtures now activate the actual
  chooser before completing it. Existing Evidence-surface behavior and source
  ownership remain supported. No HTTP, schema, data or migration compatibility
  layer was needed. Source guides document the current boundary. The new handoff is explicitly included in `.markdownlint-cli2.jsonc`, a documentation-only check.

## Verification and workstream exits

Commands ran from repository root through public Make targets. Execution used
`PATH="$PWD/tmp/node-runtime/bin:$PATH"` so harness subprocesses could locate
Node. Identifiers below are run roots under `.cartulary/test-results/`; each
retains its run summary, row results and relevant browser/unit artifacts.

The selected guides were:

```sh
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=module.timeline
make task-guide ROLE=module-author OWNER=module.evidence
```

The initial retained-file/shared-creation baseline slice
`20260923T160848Z-p12019` passed. It is baseline owner evidence, not targeting
verification or an eligible full warm check. Baseline failure artifacts above
were preserved before production edits. Characterization then implementation,
production verification, and handoff were performed in that dependency order.

| Public command/selection | Result and artifact root |
| --- | --- |
| `make generate` after initial routing; again after added scenarios | PASS `20260923T161457Z-p16843`, `20260923T163849Z-p3779`, `20260923T165425Z-p29288` |
| `make test-slice OWNER=module.evidence ROWS=module.evidence.frontend_unit.file_picker_targeting,module.evidence.frontend_unit.retained_file_recovery,module.evidence.frontend_unit.timeline_file_draft_recovery` | PASS, 4/4 units, `20260923T162948Z-p25343` |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_evidence_attachment_plan_a111000002` | PASS `20260923T162949Z-p25573` |
| `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser.file_row_targeting,module.evidence.browser.file_draft_targeting,module.evidence.browser.file_background_targeting,module.evidence.browser.file_picker_targeting` | Initial implementation PASS, 11/11 units, `20260923T163116Z-p27148`; expanded scenarios recorded below |
| `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser.file_draft_orderings,module.evidence.browser.file_original_source_review,module.evidence.browser.file_stage_recovery,module.evidence.browser.existing_file_recovery,module.evidence.accessibility.file_recovery` | PASS, 13/13 units, `20260923T163305Z-p65717` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.evidence_access_request_lifetime_ea01,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e,web.workbook.regression.timeline_keyboard_event_and_focus_ownership_97dc4e9a17` | PASS, 4/4 units, `20260923T163306Z-p65957` |
| `make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_active_view_schema_id_selects_inspector_c_82f9499272` | PASS, 2/2 units, `20260923T164501Z-p46010` |
| `make service-backed-test-slice OWNER=module.evidence` selecting the four existing visual rows below | PASS, 11/11 units, `20260923T163948Z-p41588`; 4 browser tests, unchanged golden manifest, renderer attestation and visual reconciliation PASS |
| `make format` | PASS; latest pre-finalize run `20260923T165444Z-p32593` |
| Early `make frontend-typecheck` | PASS `20260923T163154Z-p57535`; repeated in final verification |

The four visual row suffixes are
`capture_evidence_count_affordance_available_requ_cfada809e4`,
`the_visual_harness_captures_blocked_evidence_acc_779473e830`,
`the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc`, and
`the_visual_harness_captures_requested_evidence_a_1eb50235af`, all under
`module.evidence.visual`. Reviewed capture intents/reconciliation and matching
Evidence access and Timeline badge goldens at the declared crop, density,
viewport and renderer. These unchanged layout fixtures are implementation support,
not targeting proof or a Core 05 publication claim.

### Failure classification

- Initial browser routing required generation before new rows were selectable.
- `20260923T161517Z-p19926` and `20260923T161613Z-p52615` failed with
  `infra service_start_error`: startup services were ready, but the harness child
  could not spawn Node. Supplying the existing local Node runtime PATH fixed this;
  no harness or dependency change was made.
- Early production fixture runs (`20260923T161735Z-p54391`,
  `20260923T161923Z-p88115`, `20260923T162101Z-p22769`) tried unmounted/hidden
  cells or did not activate the actual semantic gridcell. Fixtures were corrected
  to production grid focus/column controls; these are not product defects.
- Initial new unit assertions exposed the baseline defects above; early typecheck
  caught unchecked fixture array elements and mock call indexing, corrected in
  tests before the passing check.
- Extended run `20260923T163923Z-p11639` passed background rejection/active draft
  and keyboard cancellation/source removal. It failed one incorrect expected
  whitespace in native paste and an expectation of automatic linking after
  detachment. `20260923T164138Z-p77368` passed filtering/original review but exposed
  the same existing review requirement after ending a native editor. Tests now
  honor Core 03 §8.1's retained review rule; production recovery was not weakened.

- `20260923T164439Z-p16183` failed extended browser fixtures: Tags needed
  explicit scrolling after column reordering; a larger fixture with equal sort
  values put the target outside the mounted window; native authoring could settle
  before capture and legitimately avoid review. Fixtures now use semantic scroll
  helpers and explicit timestamp ordering, await ordinary autosave, and allow
  either existing source-review outcome while asserting the same association and
  exact file request counts. The expanded three-row selection passed at
  `20260923T165035Z-p56100` (11/11 units), including actual virtual row unmount.
- `20260923T165243Z-p93804` failed with `artifact_error` because the final draft
  chooser scenario was added to authored routing before its test title was present
  while a combined run was still starting. No product outcome was measured in
  that attempt. The test and routing were completed together and regenerated
  before rerunning. Future runs should keep source/routing fixed during execution.

### Advisory query

Manual, offline, outside product checks:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py "keyboard focus error feedback" --domain ux -n 3 --json
```

Three focus results were classified through the maintained rules. R002 ADOPT:
existing focus tokens/ownership, no extra AAA claim. R012 ADAPT: invocation and
operation lifetimes follow current React/source owners, not a second component
store. R006/R008 ADOPT: local instruction and existing recovery distinctions.
R034 REJECT: incidental selectors; fixtures use view/record/field identities and
semantic roles. No upstream design-system generation, persistence, palette,
framework or dependency recommendation was adopted.

## Final verification and acceptance

Final combined production selection: PASS, 13/13 units, at
`20260923T165508Z-p37133`. The exact command was:

```sh
make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser.file_row_targeting,module.evidence.browser.file_draft_targeting,module.evidence.browser.file_background_targeting,module.evidence.browser.file_picker_targeting,module.evidence.browser.file_editor_continuity,module.evidence.browser.file_delayed_target,module.evidence.browser.file_picker_lifetime,module.evidence.browser.file_picker_promotion,module.evidence.browser.file_draft_orderings,module.evidence.browser.file_original_source_review,module.evidence.browser.file_stage_recovery,module.evidence.browser.existing_file_recovery,module.evidence.accessibility.file_recovery
```

Its production-grid report includes eight targeting scenarios and four retained
functional recovery scenarios; the accompanying accessibility group passes.
The association/request-count attachments prove B-only association for drop and
scalar/collection paste, all four draft gestures, both shared creation completion
orders, native text/selection/composition and chooser focus borrowing, virtual
unmount/filter/detachment retention, background rejection/active draft, keyboard
cancellation, source deletion, and chooser completion after draft promotion with
zero-byte upload. An accepted original retains its identity; existing review after
detachment never selects a replacement. Recovery rows check exact request bytes
and IDs and reads without repeating acknowledged writes.

Finalization passed before broader static verification. No adopted-owner contradiction remained; existing review after detachment was preserved.
Retained-run maintenance is skipped with RESULTS_DIR unset: focused slices do
not qualify as full warm runs.

| Acceptance | Assessment | Scope and evidence |
| --- | --- | --- |
| A001 Authority | PASS | Exact Core/design clauses and independent source/verification map above; authorized clarification stays in §8.1. |
| A002 Scope | PASS | One admission decision; full rubric, future reuse path, concrete retirement/retention and no speculative framework. |
| A003 Repository state | PASS | Clean baseline and final branch/commit recorded; current manifests, stack, grid import boundary, guides and generated registry inspected. Digest date/commit drift recorded. |
| A004 Tokens | PASS | No color, geometry, typography, density or token registry additions; existing control styles retained. |
| A005 Theme | PASS | No theme change; existing dark_graphite visual capture/reconciliation passes. |
| A006 Density | N/A | No density selection, geometry or CSS changed; this admission slice does not revise the shared density algorithm. |
| A007 Creation | PASS | Captured original draft and accepted alias; production scalar/collection gestures, both completion orders, chooser-before-promotion and exactly one creation. Existing capability/authority gates preserved. |
| A008 Responsive | N/A | No viewport thresholds, Inspector clamp, resize or responsive chrome changes. |
| A009 Overflow | N/A | No shell/grid/Inspector overflow or navigation ownership changes. |
| A010 Inspector | PASS | Stable original source across Inspector replacement and detachment; existing review invalidation retained, no forced Inspector opening. Dispatcher behavior unchanged. |
| A011 Continuity | PASS | Production source identity after picker replacement/filter, draft promotion, virtual unmount and detachment; native focus/text/selection/composition snapshots. |
| A012 Transactions | PASS | Request counts and retained exact replay rows; Web Crypto ID creation unchanged; one gesture admits one file operation. |
| A013 Acknowledgement/recovery | PASS | Existing stage-recovery/existing-file rows and retained units preserve exact uncertain attempts, accepted receipts, fresh-slot rules and read-only refresh. No generic queue retry redesign. |
| A014 Editing | PASS | Native text paste, mixed clipboard file precedence, composition continuation, caret/selection, chooser focus borrowing/cancel, and existing Timeline keyboard regressions. |
| A015 Conflict | N/A | Same-field conflict presentation and saved/local-value comparison were not changed; file original-source review remains covered separately. |
| A016 Data/interaction states | PASS | Grid editable-mode admission remains separate from rows; original-source availability/read failure never supplies authority. Existing file-owner authority rechecks and recovery units pass. General matrix unchanged. |
| A017 Authorization scope | PASS | Retained-file units cover session replacement/suspension, authority denial/recovery, access fencing and account replacement; initial gate checks current generation/authority after reads. No generic clear-all policy introduced. |
| A018 Evidence | PASS | Existing distinct Evidence lifecycle/overlay/access fixtures pass; byte attachment does not alter custody or lifecycle. |
| A019 Accessibility | PASS | Owner-selected file-recovery accessibility row, keyboard Enter/Space/cancellation, existing keyboard ownership tests and visible native focus checks; no extra conformance claim. |
| A020 Components | PASS | Shared compact/full chooser units and production grid/Inspector paths pass; current visual fixtures unchanged. No new variants, typography or overflow geometry. |
| A021 Virtualization | PASS | Production 63-row fixture unmounts the source row while transfer is held, then filters it out; original B alone receives the association after existing review. No fake rows or performance claim. |
| A022 Visual fixtures | PASS | Four existing visual rows and renderer/golden reconciliation pass; capture intents and matching declared-crop goldens reviewed. These are layout support, not targeting authority. |
| A023 Selectors | PASS | Existing semantic view/record/field IDs, roles and capability states; no visual row numbers or vendor coordinates. UI contract verification recorded below. |
| A024 Test authority | PASS | Added tests/imports/routing audited: no docs/Markdown reads, hashes, stat calls or assertions. Human handoff/owner edits remain outside product inputs. |
| A025 Generated artifacts | PASS | Authored routing/source assignment changed first; only Make generators changed registered outputs; final policy/drift results below. |
| A026 Compatibility | PASS | All internal callers migrated, no public/data changes or indefinite aliases; no migration, explicit rollback and retained owner contracts. |
| A027 Handoff | PASS | This record contains scope, owners, rubric, observed defects, cutover, commands/artifacts, failure classification, limits and rollback. Final checks are recorded below; no historical pass substitutes for fresh evidence. |

### Final repository checks

| Command | Result and run root |
| --- | --- |
| `make agent-finalize` without RESULTS_DIR | PASS `20260923T165639Z-p72850`; `unit-artifacts/finalize-summary.json` reports `results_dir: null`, `results_dir_status: skipped`. Retained evidence validation/maintenance was skipped because no eligible successful full warm run was supplied. |
| `make frontend-typecheck` after fixture cleanup | PASS `20260923T165947Z-p88136` |
| `make lint-markdown` | PASS `20260923T165947Z-p88163`; after including the final handoff in the explicit documentation-lint globs, PASS `20260923T170207Z-p91703` (`adhoc/lint-markdown/tool-run-summary.json`). |
| `make frontend-import-boundary-check` | PASS `20260923T165803Z-p77285` |
| `make test-catalog-check` | PASS `20260923T165803Z-p77656` |
| `make generated-artifact-policy-check` | PASS `20260923T165803Z-p77097` |
| `make generate-drift` | PASS `20260923T165803Z-p77087` |
| `make test-slice OWNER=package.ui` | PASS, 10/10 units, `20260923T165803Z-p77240` |
| `make lint-biome` | Initial FAIL `20260923T165803Z-p77342`: five non-null assertion warnings in new tests. Replaced with explicit fixture checks/literal fixture IDs; PASS `20260923T165947Z-p88155`. |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_evidence_attachment_plan_a111000002` after test-guard cleanup | PASS `20260923T165947Z-p88078` |

No production code changed after the final combined production run. All applicable acceptance rows are PASS; N/A rows have the scope rationale above. The final
lint repair only tightened fixture null checks. `git diff --check` passes.
Broad backend/release/measurement suites were not selected: there is no backend,
protocol, persistence, runtime topology, layout algorithm or performance change.
All scoped product exits are complete; no additional implementation work is deferred.

## Compatibility, limits and rollback

Only internal Timeline callback signatures changed. Shared Evidence chooser
behavior remains supported and its consumers/tests were migrated together.
No data migration, reload recovery, persistent browser storage or API change.
The existing retained-file review after presentation detachment remains intentional;
new tests distinguish it from retargeting or duplicate upload.

Tests use Chromium's actual chooser activation and file completion, DataTransfer
file events, native clipboard text and CDP composition. They do not claim manual
OS file-dialog behavior, OS IME coverage or cross-browser conformance. Ambiguous
identity rejection is unit evidence; production cannot legitimately materialize
duplicate authoritative row identities. No performance or speed claim was made.
Visual/accessibility evidence is implementation support only.

Rollback is a coherent revert of the resolver/admission callbacks, chooser
capture, initial availability gate, owner clarification, migrated tests/source
guides and authored routing, followed by repository generation. There is no data
rollback or migration. Preserve the existing retained operation owner, accepted
receipts and source-review obligations in either version. No deployment was made.
