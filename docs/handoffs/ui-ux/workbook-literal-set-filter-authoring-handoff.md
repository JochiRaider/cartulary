# Workbook literal-set filter authoring handoff

## Baseline, scope and authority

Implemented on `main` at `adc8cafe86bdda668bf6eab476c4e52db4fb8178`, still one commit ahead of `origin/main`. The checkout was clean before this slice; HEAD was rechecked after implementation and did not advance. No unrelated work was changed, branch created, or commit made. Root `AGENTS.md` was the applicable repository procedure.

Behavior owners are Core 01, `docs/spec/01_architecture_storage_and_view_contracts.md`, REQ039–046 and REQ494; Core 03, `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, applicable native keyboard/query clauses including REQ220, REQ223–224 and Assessment support REQ304; and design §8.3 for exact reopening, accepted summaries, keyboard access and continuity. `docs/domain.md` supplies vocabulary and navigation. The refactor digest, bundled prompts and historical handoffs were navigation/source material. Specifications were not changed.

Source ownership stays `web.workbook`, recorded independently in `tools/frontend_source_ownership.json`. Verification is routed by authored `tools/test_families/{web.workbook,module.workbook,module.savedviews}.json` and existing verification owners. New row IDs were derived with public `make author-test-row-id`; generated routing was refreshed through Make. No executable consumer was added for Markdown.

The selected boundary is the literal-set draft and its two editor presentations. It hides the common decision about membership, validation, disclosure and deliberate row manipulation. Stable member rows repair actual operand loss and avoid separate delimiter parsers in each editor. A future set operand can reuse these rows without acquiring query state, selection, authorization or discovery ownership. No general form framework was introduced.

## Representation and compatibility

`FilterDraft` collection operands and non-boolean equality sets now contain readonly `{ id, value }` member records. IDs identify draft slots; only member values enter requests and saved JSON. Each restored JSON array element becomes one record directly. Equality scalar, one-member set and null shapes remain distinct; scalar and set drafts remain independently retained on mode changes. Numeric array restoration retains numeric wire values. Boolean operands retain their dedicated typed choices.

Both editors use `WorkbookLiteralSetFilterOperand.tsx`: native one-row textareas, Add value and deliberate Remove actions. A comma, quote or pasted line break never creates another member. Raw editing text survives validation; whitespace-only, zero-member and unsupported control/format-character operands block admission with associated correction guidance. Interior line breaks are invalid under the existing query text normalization; permitted interior whitespace and Unicode remain literal content. Query operands acquire no tag-creation length limit. Admission checks each raw member before request copies are trimmed or coalesced; server NFC/case normalization, sorting and deduplication remain authoritative.

Add focuses the appended textarea. Remove focuses the next surviving member, then the preceding member, then Add value. Stable slot keys retain input identity during editing/removal. Main popover reconciliation uses the same semantic successor; its registered native form navigation, dismissal and return-focus owner remain intact. Candidate controls retain native tab order.

Main Apply still uses existing query admission and canonical acceptance. Candidate Add filter stages locally; Apply candidate query performs discovery. Typing, Add value, Remove and staging perform no discovery. Selected record IDs remain owned by existing candidate discovery/picker and Assessment controllers, including filtering, paging, failed reads and cancellation.

Accepted chips, edit listings and candidate summaries share quoted bracketed set disclosure: `["review,priority"]` versus `["priority", "review"]`. Existing truncation and complete accessible disclosures are retained; accepted and requested projections remain separate. Scalar, boolean, enum, date and timestamp behavior is retained, with set summaries now disclosing array shape explicitly.

Retired paths: comma/newline splitting in `buildFilterFromDraft`; array-to-delimiter joining in `filterDraftFromFilter`; enum string compatibility parsing; delimiter text inputs/instructions in both editors; comma flattening of set summaries; enum member inputs keyed by positional index. No implicit bulk-entry convenience remains. Existing saved `["review", "priority"]` keeps its present meaning and is never combined automatically.

No backend, route, query field, operator, contract schema, saved-view schema or storage migration changed. Inspected `internal/platform/viewquery/query.go` and `internal/modules/savedviews/policy.go` confirm existing array admission/normalization/persistence compatibility. `fieldnorm.NormalizeLine` is the text-normalization reference; writable tag limits do not define query limits.

## Inspected and changed consumers

Production changes are confined to:

- `apps/web/src/workbook/models/workbookQuery.ts`: structured restoration, serialization, member validation/control keys and set formatter.
- `WorkbookEnumFilterOperand.tsx` / `workbookEnumFilterOperand.ts`: declared choices retained, custom set rows reuse the shared treatment.
- `WorkbookFiltersControl.tsx`: main editor, associated feedback and dynamic focus reconciliation.
- `WorkbookGridControls.tsx`, `useWorkbookQueryController.ts` and `useTimelineWorkbookPresentation.tsx`: existing callers pass their current contract to draft clearing, preserving enum choice reset without a phantom blank custom member.
- `WorkbookCandidateQueryControl.tsx`: candidate editor and staged summaries.
- `workbookViewBarWorkingSet.ts`: truthful accepted/requested disclosure.
- New `WorkbookLiteralSetFilterOperand.tsx`, registered under `web.workbook`.

Inspected the downstream grid query projection, query controller, Timeline presentation, `WorkbookQueryBrowser`, authoring reference picker, candidate discovery and Assessment discovery. Their query/selection/lifetime ownership was retained. Tag mutation controllers/intents and Timeline/Notes view contracts were inspected for compatibility and left unchanged.

Updated the query model, grid model/component, query controller and shell query test fixtures to the structured draft shape. Added focused model/component regressions, browser scenarios in `workbook-query-browsing.spec.ts` / `authoring-candidates.spec.ts`, and synthetic `e2e/support/workbook/literalSetFilters.ts`. Catalog changes generated only `tools/browser_e2e_batch_manifest.json` and `tools/execution_topology_render_index.json`; these outputs were not edited manually.

## Before and after evidence

Characterization ran before production edits:

| Command selection | Result / run root |
| --- | --- |
| New frontend rows (literal round trips and explicit editors) | Expected failures: literal split to separate operands in model/main/candidate. `.cartulary/test-results/20261007T003048Z-p35512` |
| New Workbook browser tag/organization/contextual rows | Expected failures waiting for explicit member controls absent in the old editor. `.cartulary/test-results/20261007T003048Z-p35517` |
| Saved exact-array browser row | Expected product failure: exact saved predicate initially/restored A; unchanged editor reapplication lost A and matched component tags. `.cartulary/test-results/20261007T003048Z-p35518` |

The first two failing selections were run exactly as follows; the saved-array command is the same single-row command listed under final browser selections.

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.lossless_literal_set_draft_round_trips_0dd48635bf,web.workbook.regression.explicit_literal_member_editors_5902c7ecdd
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.exact_literal_tag_set_membership_fcb35d55a3,module.workbook.browser_stateful.literal_organization_equality_membership_b30b827b5d,module.workbook.browser.literal_contextual_candidate_staging_1c1261f5df
```

The first two browser authoring failures characterize missing explicit boundaries, rather than independently reproducing the supplied initial UI split. The saved-view round trip independently exercises actual membership loss on unchanged reapplication.

Final populated browser fixtures distinguish A (only `review,priority`), B (only `priority`) and C (separate `review` and `priority`). Timeline A's tag is assigned through the UI. Timeline and Notes assertions inspect exact response and visible record IDs:

| Operands | Contains any | Contains all |
| --- | --- | --- |
| `["review,priority"]` | A | A |
| `["review", "priority"]` | B, C | C |

Party organization equality returns the comma-containing organization alone; explicitly separate `Northwind` and `Inc.` operands return the two component organizations. Accepted operands are case-folded by existing server normalization. Saved exact arrays restore, reload, reopen and reapply unchanged; UI-created saved resources also retain the one-member array across reload and reopening.

Contextual candidate filtering stages without a read, explicitly discovers A and retains selected B through filtering, a failed read and cancellation. Assessment support checks both operators and both memberships, retained B, explicit staging/application and cancellation without promoting staged selections. Existing candidate/browser rows additionally cover page eviction and boolean/enum/timestamp behavior. Component coverage verifies raw correction, paste, composition, native keys, stable input identity, Add/Remove focus, ordinary equality sets and no admission callback on invalid members.

## Verification commands and run roots

All repository verification used public Make targets. Initial orientation used `make help`, `make help-all` and `make task-guide ROLE=module-author OWNER=web.workbook`, `OWNER=module.workbook`, and `OWNER=module.savedviews`. Final routing was inspected with `make explain-test-owner` for those three owners, `make explain-target TARGET=test-slice DETAIL=summary`, `TARGET=service-backed-test-slice DETAIL=summary`, and `TARGET=frontend-typecheck` / `TARGET=frontend-import-boundary-check DETAIL=rows`. Successful `make target-plan TARGET=frontend-typecheck`, `TARGET=frontend-import-boundary-check` and `TARGET=lint-biome` covered actual static targets. `target-plan` does not accept OWNER/ROWS; the rejected attempted slice plan did not execute tests.

Functional browser cases use the isolated synthetic worker-admin actor from `apps/web/e2e/fixtures.ts`. Existing constrained keyboard/enum scenarios explicitly exercise 768×640; the separately seeded editor review below uses both requested viewports. No browser identity in this handoff is a real user or live record.

The exact final frontend selection is:

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookquery_suite_904073db6c,web.workbook.regression.grid_controls_component_a104000002,web.workbook.regression.authoring_candidate_presentation,web.workbook.regression.grid_controls_model_a104000001,web.workbook.regression.useworkbookquerycontroller_suite_203a4e98cc,web.workbook.regression.assessment_discovery,web.workbook.regression.lossless_literal_set_draft_round_trips_0dd48635bf,web.workbook.regression.explicit_literal_member_editors_5902c7ecdd,web.workbook.regression.workbook_query_controls_preserve_ordered_chip_ca_0fc544e3c5,web.workbook.regression.workbookshell__query_normalizes_saved_view_query_22724faeb7,web.workbook.regression.workbookshell__query_notes_full_text_controls_su_889a129c8b
```

Final browser selections:

```bash
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.exact_literal_tag_set_membership_fcb35d55a3,module.workbook.browser_stateful.literal_organization_equality_membership_b30b827b5d,module.workbook.browser.literal_contextual_candidate_staging_1c1261f5df,module.workbook.browser.literal_assessment_support_membership_d541f44d18,module.workbook.browser_stateful.enum_equality_choices_remain_explicit_and_preser_28bc32bc06
make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.browser.literal_saved_array_unchanged_reapplication_f2ec5c3af8
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.filter_editor_keyboard_focus,module.workbook.browser.boolean_support_filters_stage_typed_queries_and_7785af35c2,module.workbook.browser_stateful.boolean_filters_preserve_typed_saved_and_accepte_221721fc9a,module.workbook.browser_stateful.date_filter_drafts_remain_locally_correctable_be_15ce5a6e7c,module.workbook.browser_stateful.enum_equality_choices_remain_explicit_and_preser_28bc32bc06,module.workbook.browser.assessment_timeline_support_enum_filtering_prese_9700791fe7,module.workbook.browser.party_timestamp_candidate_filters_stage_explicit_45110487e8
make service-backed-test-slice OWNER=module.savedviews ROWS=module.savedviews.browser_stateful.browser_saved_view_query_layout_state_user_home_cb9c681674,module.savedviews.browser.timestamp_saved_queries_reload_and_reopen_withou_6a98a26dd7
```

| Verification | Result / run root |
| --- | --- |
| Final frontend selection | PASS, 12/12 graph units (11 rows); `.cartulary/test-results/20261007T010438Z-p20371` |
| Final four new Workbook browser rows plus enum regression | PASS, 13/13; `.cartulary/test-results/20261007T010441Z-p20908` |
| Final new saved-array browser row | PASS, 11/11; `.cartulary/test-results/20261007T010442Z-p21412` |
| Existing affected Workbook browser rows | PASS, 14/14 graph units; `.cartulary/test-results/20261007T004613Z-p76260` |
| Existing saved-view replay/timestamp browser rows | PASS, 13/13; `.cartulary/test-results/20261007T004614Z-p76535` |
| `make format`, authored-source diff inspected | PASS; `.cartulary/test-results/20261007T010359Z-p15013` |
| `CARTULARY_GENERATE_DRIFT_REFRESH=1 make generate-drift` | PASS; `.cartulary/test-results/20261007T004309Z-p41245` |
| `make agent-finalize` before final static verification | PASS; `.cartulary/test-results/20261007T010439Z-p20646` |
| `make generate-drift` | PASS, 4/4; `.cartulary/test-results/20261007T005950Z-p62995` |
| `make generated-artifact-policy-check` | PASS, 3/3; `.cartulary/test-results/20261007T005950Z-p62987` |
| `make frontend-typecheck` | PASS, 2/2; `.cartulary/test-results/20261007T010600Z-p8173` |
| `make frontend-import-boundary-check` | PASS, 2/2; `.cartulary/test-results/20261007T010600Z-p8185` |
| `make lint-biome` | PASS, 2/2; `.cartulary/test-results/20261007T010600Z-p8225` |
| `make lint-markdown` | PASS; `.cartulary/test-results/20261007T010832Z-p12002` |
| `git diff --check`, complete working-tree review | PASS; only the cohesive slice, no whitespace errors or unrelated formatter changes |

Intermediate corrective runs are retained honestly: typecheck test-fixture typing failure at `20261007T003645Z-p18635`; member focus failure at `20261007T004049Z-p31085`; old summary assertions at `20261007T004155Z-p33545`; an ambiguous component text selector at `20261007T004414Z-p15841`; Assessment fixture's undeclared reset field at `20261007T004339Z-p52249`; Party canonical-case expectation at `20261007T004718Z-p65099`. Format initially rejected new non-null assertions (`20261007T004022Z-p21771`, `20261007T004156Z-p33887`); catalog title order was corrected before generation. The final review also characterized the enum reset regression at `20261007T010326Z-p14168`, then corrected all reset callers and added a component check that another declared choice remains admissible. These were repaired and are not substituted for final passes. Bare run IDs in this paragraph are beneath `.cartulary/test-results/`.

`RESULTS_DIR` was unset: retained-run maintenance was skipped because no qualifying successful full warm-check run was supplied. Broad backend/release/performance suites and visual-golden refresh were skipped because this slice changes neither backend/storage nor the grid renderer/measurement contracts. No goldens or lockfiles were edited.

## Rendered review and limits

Used `cartulary-ui-review` with `make ui-review UI_MODE=seeded`, actor `editor`, current production renderer, `dark_graphite`, compact density, and viewports 1440×900 and 768×640. Finite `ui-browser` / `ui-capture` commands used schema-valid private scratch requests, exact returned session/epochs and observed targets. Seven image captures were byte-length/SHA-256 verified and inspected before exact-session stop. Scratch requests were removed.

Reviewed Timeline main member rows and Assessment Timeline support candidate rows with a comma-containing value, a long Unicode/quoted value and an invalid blank row. Checked visible focus, guidance, vertical scrolling, native Tab traversal to candidate Apply, correction by deliberate removal and quoted staged summary. Main footer is reachable by native focus scrolling at constrained height; candidate footer and summary remain reachable inside inspector scrolling. Browser functional rows also cover contextual candidates and text spacing for the shared enum treatment.

Logical captures were main desktop/constrained/focused footer (bundles 12, 14, 16) and candidate desktop/constrained/Tab-focused Apply/staged summary (37, 40, 44, 48). Review root `.cartulary/test-results/20261007T004720Z-p66090` retains structural receipts only; private screenshots/observations expired on stop and are not handoff links. Terminal receipt reports `cleanup: complete`, seven images, no console errors or failed requests. Four axe observations completed with zero violations and six aggregate incomplete observations; those incomplete checks and rendered-node-only/no-trace limits remain. This is supporting design evidence, not a product-test, accessibility-conformance or golden/publication pass.

An earlier review (`20261007T004401Z-p88762`) failed source sealing while source was still being formatted; exact-session stop confirmed cleanup complete. In the successful review, one attempt to click a grid cell obscured by the constrained inspector timed out; a fresh snapshot followed, and candidate inspection proceeded through native Assessment support navigation. That failed operation is not membership evidence. After review, the identical raw-member checks were consolidated and format-character validation was tightened for BOM correction; member-row rendering and focus behavior did not change. The later enum reset correction affects reopening a cleared enum draft, covered by component/browser regression, rather than the inspected literal rows. Final product checks cover these changes.

## Acceptance assessment

Assessments apply to this literal-set slice, rather than asserting a fresh complete subsystem audit.

| Row | Assessment | Evidence / scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Exact behavior owners above; source and independent verification catalogs distinguished. |
| A002 Scope | PASS | One structured member boundary, two migrated editors, explicit retirement, future reuse and risk/rollback decisions. |
| A003 Repository state | PASS | Clean starting main/HEAD rechecked; root AGENTS, stack, current ownership/import manifests and generated policy inspected. |
| A004 Tokens | PASS | Existing text-input/button styles and spacing tokens; no token/theme/density registry added. |
| A005 Theme | PASS | Existing dark_graphite renderer reviewed; no theme choices changed. |
| A006 Density | N/A | No row/header/cell density or geometry owner changed; member controls reuse current styles. |
| A007 Creation | N/A | No creation capabilities, minima, routes or writes changed; candidate filtering remains within existing authoring owners. |
| A008 Responsive | PASS | New member controls checked at both requested sizes; native vertical resizing/scrolling retains controls. Responsive accessors/clamp policies unchanged. |
| A009 Overflow | PASS | Main popover and inspector scrolling retain actions, shell navigation and status strip in inspected states. |
| A010 Inspector | N/A | No dispatch, feature routing, confirmation or late-operation ownership changed. |
| A011 Continuity | PASS | Stable input nodes/raw values, deliberate focus successors, retained selected IDs, failure/cancel and saved reload/reopen checks. |
| A012 Transactions | N/A | No transaction identity, replay, mutation admission or duplicate activation behavior changed. |
| A013 Recovery | N/A | No write queue/retry/acknowledgement ownership changed; candidate read failure selection retention is covered under A011/A016. |
| A014 Editing | PASS | Native textareas/paste/composition/keys, invalid correction, explicit removal, dismissal and return-focus checks. |
| A015 Conflict | N/A | No cell conflict presentation or write behavior changed. |
| A016 Query states | PASS | Main accepted/requested ownership retained; invalid drafts do not admit; candidate staging, explicit reads and failed reads retain identities. |
| A017 Authorization scope | N/A | No authorization, incident/account clearing, protected-data or refresh lifetime owner changed; no new reads from editing. |
| A018 Evidence | N/A | No evidence lifecycle, upload, preview or overlay changes. |
| A019 Accessibility | PASS | Selected native-keyboard/focus browser and component checks; associated invalid feedback and complete disclosure; supporting review with stated axe limits. No broad conformance claim. |
| A020 Components | PASS | Shared main/candidate/enum member variants, long Unicode content, constrained scrolling and existing text-spacing browser scenario. Density/zoom owners unchanged. |
| A021 Virtualization | N/A | No grid adapter, virtual row identity, virtualization or performance behavior changed. |
| A022 Visual fixtures | PASS | Scoped seeded production renderer/captures inspected with declared viewport/theme/density; diagnostic status retained, no golden refresh. |
| A023 Selectors | PASS | Exact view/record/field IDs and semantic role/name selectors; no label promoted to a candidate identity. |
| A024 Test authority | PASS | Executable changes use typed contracts/source and authored JSON routing; no Markdown dependency added. |
| A025 Generated artifacts | PASS | Authored catalogs/source manifest updated; public generation/finalize/drift and generated-policy checks; no manual generated edits. |
| A026 Compatibility | PASS | Existing arrays preserve meaning; server authoritative; no route/schema/storage migration; obsolete delimiter paths removed. |
| A027 Handoff | PASS | Commands, failures, evidence/limits, all acceptance rows and rollback recorded here; no applicable criterion blocked. |

## Remaining risk and rollback

Comma-entry semantics intentionally change: existing users must explicitly Add value for another member. Very large sets increase editor height and use the existing popover/inspector scrolling; no query-size limit was invented. Native IME evidence is composition-event regression coverage, not a physical IME certification. Narrow rendered review does not establish unrelated layout, zoom, authorization, performance or accessibility conformance.

Rollback reverts the cohesive frontend, test, source-manifest and authored-catalog slice, then regenerates derived routing through public Make generation/drift workflow. Stored arrays are already compatible JSON; no data rollback is expected. No automatic rewrite of old saved predicates is permitted.
