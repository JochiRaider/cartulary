# Timeline relationship mention focus reveal

Date: 2026-09-30. Status: implementation and review complete; final verification recorded below.

## Boundary and authority

The checkout started clean on `main`, aligned with `origin/main`, at
`1f23eb98178ea755b2a75538bb6c7d74643dc321`. HEAD and branch remain unchanged.
All changes in this slice are uncommitted; no unrelated work was present.

Source ownership is `web.workbook`, rechecked against
`tools/frontend_source_ownership.json` and `tools/frontend_import_boundaries.json`.
Verification is independently routed through `module.timeline`, `module.entities`,
`web.workbook` and `module.workbook`, using their authored family manifests.
The current public Make surface and all four owner guides were inspected.

The governing owner is
[Core 03](../../spec/03_workbook_interaction_collaboration_and_workflows.md):
§2.3A governs current Inspector subjects, review validity and operation lifetimes;
§9 REQ-03-129–134 governs inspection and explicit resolve/dismiss/restore;
REQ-03-219 requires independent retained collection authoring and focus borrowing
without submission. REQ-03-283 and REQ-03-299/100 retain their continuity and
scoped authority boundaries. [Design](../../design.md) §§7.1, 7.3, 9.2–9.3,
12.7 and 14 supply independent scrolling, Inspector presentation, chip semantics,
keyboard navigation and visible focus. [Domain](../../domain.md) supplies vocabulary
and navigation. Research, embedded prompts, the digest and historical handoffs
were consulted as source material. No owner contradiction or specification change
was found.

Both `cartulary-ui-ux-refactor` and `cartulary-ui-review` were applied. The digest
README, START_HERE, LOCAL_AGENT_PROMPT, REPO_MAP, OWNER_MAP, QUERY_RECIPES, rules and
acceptance material, applicable AGENTS.md and corresponding source guides were
read. The narrow offline UX query for keyboard focus visibility used suppressed
Python caches; relevant focus advice was ADOPT under R002, without importing an
additional upstream conformance profile.

## Reproduction and correction

The current-HEAD seeded review reproduced the original defect at 1440×900 and
1024×720 before implementation. With a saved Timeline row, two unresolved Host
mentions and the first mention's details expanded, ArrowRight gave the second
button DOM focus at y=1024.421875, height=26.796875, with no visible rectangle.
The images were hash-verified and inspected. That review was observation evidence,
not a failed product test. Its exact session was stopped, cleanup completed, and
the foreground process exited successfully.

Before editing production code, a new production-browser characterization was
added and routed. Its pre-key fixture proved the origin contained and the next
button clipped. The first run failed the destination containment assertion:
focused=true, contained=false, y=1330.421875–1357.21875 versus usable clip
170–871, with a computed 4px ring. Test assertions after the arrow only observed;
no helper repaired focus or scroll.

Arrows now keep the panel's original group-local neighbor calculation and borrow
focus through the live semantic registry. The Inspector navigation boundary
accepts optional `start` or `nearest` placement, defaulting to `start`. Only
mention-list arrows use `nearest`; explicit inspection and mutation restoration
retain their default positioning. The navigator validates current scope, admitted
section membership and availability, focuses with `preventScroll`, checks focus
acknowledgement, and minimally reveals that target and its ring in the existing
body. Disconnected, hidden, inert, disabled, CSS-concealed and closed-disclosure
targets are rejected. This path has no deferred request, timer, observer, polling
or persisted focus state.

The common decision is bounded Inspector-body reveal: client clipping, viewport
intersection, clipping ancestors, computed outline allowance, scale conversion,
clamped scroll and boundary correction. The existing geometry now exposes a
control-only entry alongside ordinary field label/value context reveal. The
persistent header stays outside the body clip, and navigation retains its
programmatic-scroll tracking. No second scroll container or theme was introduced.

Selection, candidate discovery, collection authoring and mutation admission remain
with their existing owners. Tests verify Host/Identity and active/dismissed
boundaries; raw independent grid/Inspector drafts and carets; unchanged candidate,
pressed state, record and expanded details; no arrow-induced API requests or
collection submission; unchanged grid/document scroll; unavailable and replaced
subjects; viewer and closed read-only inspection. Enter and Space still explicitly
select the focused semantic mention. Existing Tab/Escape and native-control event
ownership are unchanged and covered by the retained accessibility integration.

## Changed paths and retirement

Paths below are relative to the repository root.

- `apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx`: presentation
  callback replaces direct arrow focus. The local button map and `useRef` import
  were retired with all their callers. Mention and collection registrations remain
  because their semantic consumers still need them.
- `apps/web/src/workbook/timeline/components/TimelineWorkbookInspector.tsx`: supplies
  the callback with the current live subject and registry.
- `apps/web/src/workbook/timeline/focus/timelineInspectorElementRegistry.ts`: forwards
  optional placement, retaining subject/source fencing and synchronous mention focus.
- `apps/web/src/workbook/layout/workbookInspectorNavigation.ts`: owns availability,
  focus acknowledgement and nearest reveal; start navigation remains the default.
- `apps/web/src/workbook/inspector/presentation/workbookInspectorFieldReveal.ts`:
  shares existing geometry while preserving ordinary field context selection.
- `TimelineCollectionCell.test.tsx`, `timelineInspectorElementRegistry.test.ts` and
  `WorkbookInspectorPresentation.test.tsx` in those source directories: extend
  component, semantic lifetime and real-shell navigation evidence.
- `apps/web/e2e/workbook.a11y.spec.ts` and `apps/web/e2e/support/entities/mentions.ts`:
  production browser regression and separate fixture/observation helpers.
- `tools/test_families/module.entities.json`, `module.timeline.json` and
  `web.workbook.json`: authored routing. `make generate` projected only
  `tools/browser_e2e_batch_manifest.json` and `tools/execution_topology_render_index.json`.
- Inspector presentation and Timeline focus source READMEs document the capability.
  This handoff is the completion record; no separate controlling tracker exists.

The browser observation helper intentionally measures the result independently of
production reveal. It is not a competing production navigator. Ordinary field
context and target-only reveal share geometry but keep their different context
choices. No duplicate selected-mention state remains or was added.

## Verification

All commands ran from the repository root through public Make targets. Run roots
are under `.cartulary/test-results/`; each selected slice's run summary retains
its exact owner/row selection. `make explain-run` identified the initial browser
failure as a product assertion failure.

| Command or selection | Result | Run ID |
| --- | --- | --- |
| New entity accessibility row, before production changes | Expected FAIL, destination containment | `20260930T130525Z-p53134` |
| `make generate`, initial browser routing | PASS | `20260930T130506Z-p50207` |
| `make generate`, all authored routing | PASS | `20260930T132256Z-p94753` |
| `make format`, final source/test formatting; diff inspected | PASS | `20260930T133957Z-p24427` |
| `make agent-finalize`, no RESULTS_DIR | PASS | `20260930T132324Z-p2469` |
| Timeline collection and registry slices | PASS | `20260930T132407Z-p7073` |
| Shared Inspector presentation and edit-focus guard slices | PASS | `20260930T132407Z-p7084` |
| New mention-arrow row, existing entity accessibility, resolve/create and dismiss/restore | PASS | `20260930T132420Z-p8797` |
| Timeline/Entity ordinary-edit and Evidence shortcut reveal | PASS | `20260930T132515Z-p45964` |
| `make frontend-typecheck` | PASS | `20260930T132516Z-p46621` |
| `make frontend-import-boundary-check` | PASS | `20260930T132516Z-p46634` |
| `make lint-biome` | PASS | `20260930T132516Z-p46651` |
| `make generate-drift` | PASS | `20260930T132626Z-p99986` |
| `make generated-artifact-policy-check` | PASS | `20260930T132626Z-p306` |
| `make json-shape-check` | PASS | `20260930T132626Z-p327` |
| `make agent-finalize`, final rerun without RESULTS_DIR | PASS | `20260930T134326Z-p35111` |
| Final mention-arrow browser case, including Enter and Space | PASS | `20260930T134812Z-p40299` |
| Final Timeline collection and registry slices | PASS | `20260930T134812Z-p40309` |
| Final `make lint-biome` | PASS | `20260930T134812Z-p40513` |
| Final `make frontend-typecheck` | PASS | `20260930T134828Z-p63140` |
| `make lint-markdown`, including this handoff and source guides | PASS | `20260930T134812Z-p40517` |

The narrow command selections were:

```sh
make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.collection_inspection_preserves_draft_text_and_s_f5c8c4c211,module.timeline.frontend.collection_inspection_focuses_only_live_semantic_40e7d15084
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_persistent_context,web.workbook.regression.inspector_edit_focus_request_guards_887cb69659
make service-backed-test-slice OWNER=module.entities ROWS=module.entities.accessibility.mention_arrow_navigation_reveals_focused_inspect_524ff6af49,module.entities.accessibility.verify_mention_chip_states_and_manual_resolution_e5964739d3,module.entities.browser.the_browser_inspector_dismisses_a_mention_and_re_6150fd43cd,module.entities.browser.the_browser_inspector_resolves_existing_entities_325548131d
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.inspector_timeline_ordinary_edit_reveal_f4bd40fdfb,module.workbook.accessibility.inspector_entity_ordinary_edit_reveal_c435d77f80,module.workbook.accessibility.inspector_evidence_reference_shortcut_reveal_afdc31949c
```

Initial formatting/lint failed on an implicit-any variable in the new shell test
(`20260930T132130Z-p89134`, `20260930T132223Z-p94042`); its type was supplied.
An attempted diagnostic flag override was rejected as an unsupported public input;
normal `make lint-biome` supplied the needed summary. Initial typecheck failed on
missing null relationship editor slots in the component fixture
(`20260930T132407Z-p7198`); the fixture was corrected and typecheck passed.
These failures were related to new test code and are resolved.

Final `git diff --check` passed. The working-tree review contains only the five
production files, focused tests/helpers, two source guides, authored routing,
their two generated outputs and this handoff. There is no applicable blocked
verification and no unrelated change.

Retained-run maintenance was skipped because RESULTS_DIR was unset: there was no
qualifying successful full warm check. Full CI/release, unrelated backend suites,
performance runs, canonical visual suites and golden refresh were skipped because
this is a bounded navigation correction with no registered appearance change.
No dependency, API, authorization, storage, query ordering or data migration changed.

## Geometry and rendered review

The product regression uses stable record, field and item identities at 1440×900,
1024×720 and 1280×720 with the repository's 200% CSS-zoom fixture approach.
Containment includes the computed ring against body client bounds, ancestor clips
and viewport, with a 1px fractional-layout allowance. Focus/visible assertions
alone are not the acceptance criterion. The successful browser measurements were:

| Layout | Direction | Button vertical bounds | Usable vertical clip | Ring | Result |
| --- | --- | --- | --- | --- | --- |
| 1440×900 | Forward | 840.015625–866.8125 | 170–871 | 4px | Contained |
| 1440×900 | Reverse | 174.015625–200.8125 | 170–871 | 4px | Contained |
| 1024×720 | Forward | 660.015625–686.8125 | 170–691 | 4px | Contained |
| 1024×720 | Reverse | 174.015625–200.8125 | 170–691 | 4px | Contained |
| 1280×720, 200% | Forward | 600.53125–654.125 | 324–662 | 8px | Contained |
| 1280×720, 200% | Reverse | 332.46875–386.0625 | 324–662 | 8px | Contained |

Corrected seeded review used isolated synthetic data against a sealed build.
Its structural run root is `.cartulary/test-results/20260930T132601Z-p79086`;
the exact session was `uireview-61c8d556805da06b8246159f95a6b2e8`.
Workspace and served source digests agreed. Live product controls added two short
unresolved Host tokens, selected the first and expanded Mention details. After
fixture positioning, arrows used the active keyboard with no repair action before
capture. Forward and reverse images were inspected at 1440×900 and 1024×720, plus
1024×720 at the harness-supported 125% zoom. Focus remained on the intended button,
the ring was visibly inside the body and selection/details remained on the first.

Reviewed bundle IDs were 25/27/31 (1440 precondition/forward/reverse),
37/39/41 (1024), and 44/46/48 (125%). Forward button y values were 839.71875,
659.71875 and 646.375 respectively; reverse values were 187.3125, 187.3125 and
215.75. The Inspector header and body edge were inspected in the images.
The harness's visible rectangle is viewport evidence and was not substituted for
the product test's ancestor/body containment calculation.

The exact 200% browser attachment was extracted without changing its bytes,
imported through `make ui-capture` as bundle 49, and inspected. Its SHA-256 was
`7713d6a8a9175bb50118932d08218039c8cf9a7c3c45d776dedecf51fc98fc68`
(67,804 bytes). Every consumed original and observation component was checked
against its manifest's SHA-256 and byte length before viewing. The PNG import has
no DOM/Axe channel; the browser test supplies its geometry. No private paths,
raw screenshots or raw observation dumps are committed here.

Axe completed on the baseline/forward states and reported heading-one and, at
1440, empty-table-header advisories, plus incomplete color-contrast assessment.
Those are outside this focus correction and are not product-test failures or a
claim of complete accessibility conformance. Repeated reverse/fixture captures
explicitly disabled redundant scans; their geometry and images were still consumed.
Malformed role/fill request keys were rejected before admission and corrected
using the schema. There were no failed product requests in the corrected review.

`make ui-review-stop` used the exact session locator. The terminal receipt reports
state=closed, status=ok, cleanup=complete, with no failures, console errors or failed
requests. The foreground process exited 0. Caller-owned request/image scratch was
removed only after inspection and cleanup. Private review links expired on stop;
the retained structural terminal receipt is the durable cleanup evidence.

## Digest acceptance

PASS below is bounded to this slice, not certification of unrelated workflows.

| Row | Assessment | Evidence or scope rationale |
| --- | --- | --- |
| A001 | PASS | Exact Core/design mapping above; source and verification ownership kept separate. |
| A002 | PASS | Common reveal geometry, bounded arrow correction, retired map and unchanged default callers reviewed. Future control reveal can use the existing target entry without a focus framework. |
| A003 | PASS | Clean baseline, current authored maps, generated policy and local source guides inspected; no direct grid-vendor imports added. |
| A004 | PASS | No production design literal, token, theme or density registry introduced. |
| A005 | PASS | Existing dark graphite renderer inspected at all review layouts; no theme change. |
| A006 | N/A | Density selection and grid/full-cell editor geometry were not changed. |
| A007 | N/A | Creation capabilities/payloads are unchanged; existing resolve/create row passed as integration evidence. |
| A008 | PASS | Body/viewport/ancestor geometry and scaling verified at required layouts; responsive chrome selection is unchanged. Broader responsive audit is outside scope. |
| A009 | PASS | Browser asserts unchanged grid/document scroll; review shows persistent context and bounded body scrolling. |
| A010 | PASS | Registry/shell reject unavailable and replaced destinations; actual-owner component covers viewer/closed states without changing admission. |
| A011 | PASS | Focus-only groups, raw drafts/carets, selected candidate/record/details and failed-request lifetime tests pass. |
| A012 | N/A | Captured request identities/bytes and duplicate mutation admission were untouched. |
| A013 | N/A | Acknowledgement, uncertainty and refresh recovery algorithms were untouched. |
| A014 | PASS | Independent draft/caret assertions and ordinary-edit reveal guards pass; Enter/Space remain explicit activation. |
| A015 | N/A | No conflict locus or representation changed. |
| A016 | PASS | Read-only/closed inspection remains independent of write eligibility in actual-owner component tests; query-state presentation is unchanged. |
| A017 | N/A | Incident/account-session recovery and protected-data clearing are unchanged. Local focus authority replacement is covered by A010/A011. |
| A018 | N/A | Evidence lifecycle/overlay/preview presentation is unchanged; shared reveal shortcut row passes. |
| A019 | PASS | Product keyboard/containment and existing accessibility integration pass; ring, names and non-color markers inspected. Axe advisories are separately reported. |
| A020 | PASS | All mention groups, expanded content, 200% browser and 125% rendered zoom, boundaries and visible-target behavior covered. |
| A021 | N/A | Virtualization and result-size behavior were not changed; semantic record identity and grid scroll are preserved in the focused case. |
| A022 | PASS | Exact production screenshots/geometry consumed through Make review. No registered golden appearance changed or refresh occurred. |
| A023 | PASS | Existing UI-contract semantic record/field/item selectors used; no selector contract change requires a package.ui slice. |
| A024 | PASS | New runtime, tests and routing consume no Markdown; docs remain human review material. |
| A025 | PASS | Authored routing changed before generation; generation drift, generated policy and JSON checks pass. |
| A026 | PASS | Optional placement retains start consumers; operation/API/data boundaries unchanged; rollback and retirement explicit. |
| A027 | PASS | This handoff records evidence, resolved failures, limitations, skipped checks and cleanup; no applicable blocked check remains. |

## Compatibility and rollback

The externally visible correction is only arrow-driven focus reveal. Default
explicit navigation, mutation-driven restoration and collection capacity are
preserved. There is no API, dependency, schema or data migration and no data
rollback is needed.

Rollback reverts the source, tests and authored family routing together, then runs
`make generate` to regenerate affected projections. Re-run the same narrow owner
slices and static checks. Review and publication remain separate: this slice is
ready for user review and has not been committed, published or deployed.
