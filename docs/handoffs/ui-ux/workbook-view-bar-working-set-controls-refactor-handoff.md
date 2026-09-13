# Workbook View-Bar Working-Set Controls Refactor Handoff

## Baseline

- Branch: `main`
- Commit: `9b9ac3c238287d8ab3a2b0607bc0883b3733920f`
- Initial worktree: clean
- Toolchain: Node `24.15.0`, pnpm `10.33.0`, Go launcher/effective `1.26.4/1.27.1`, Python `3.14.4`
- Focused green baseline: `.cartulary/test-results/20260904T001305Z-p86253`
- Commit creation: not authorized

## Authority

- Adopted owners: `docs/design.md`, Core 01 view-query and field-capability contracts, and Core 03 Workbook saved-view/query/layout/lifecycle requirements.
- Advisory inputs: the UI/UX digest and bundled upstream search package. They do not define product behavior.
- Machine projections: authored view contracts, `contracts/design/tokens.v1.json`, `packages/ui-contracts`, and verification-owner manifests.
- No adopted-owner contradiction was found.

## Allowed Paths

- `docs/design.md` and this handoff.
- Authored design-token inputs and generator-produced design-token projections.
- Workbook view-bar models, components, hooks, shell/surface composition, and their tests under `apps/web`.
- Authored stable selectors under `packages/ui-contracts` and their tests.
- Applicable browser fixtures, visual goldens, verification-family inputs, and generator-produced topology/golden manifests.

Generated roots from `tools/generated_artifact_policy.json` will not be hand-edited.

## Workstreams

| Workstream | Status | Outcome |
| --- | --- | --- |
| VB-01 — Authority, current-state inventory, and characterization | DONE | Authority, baseline, current behavior, and demonstrated mismatches recorded before product edits. |
| VB-02 — Presentation contract and pure working-set model | DONE | Narrow design correction, token-backed allocation, pure projection, and transient model. |
| VB-03 — Saved-view and query-control migration | DONE | Shared semantic projection, grouped saved-view actions, complete query editors, and subject-scoped transient controls. |
| VB-04 — Keyboard, responsive, accessibility, and visual evidence | DONE | Stable-selector browser/component evidence, accessibility matrix, zoom/text-spacing coverage, and reviewed goldens. |
| VB-05 — Final validation, scope audit, and handoff | DONE | Terminal verification, scope audit, compatibility, and completion record. |

## Current-State Characterization

- Control order is saved-view selector/actions, Sort, Group, Filters, Columns, canonical query chips, Inspector, and Add Row; Timeline may insert its owner-backed bulk controls before the protected right rail.
- Before this refactor, the responsive chip rule was `8 / 6 / 0 / 0`. The 1440 by 900 saved-view/query golden preserved eight inline chips only by clipping the saved-view identity, modified state, and repeated sort labels into ambiguous fragments.
- Query chip order is group, applied sorts, then normalized filters. Chip commands currently remove immediately, every chip is tabbable, and the full text is exposed only through `title`.
- The Filters overlay repeats all applied query entries as removal buttons. Its `Clear all` shell port resets sort, group, and filters even though the owning design rule defines filter-only clear behavior.
- Sort currently interleaves applied rows and the complete sortable-field inventory. Columns remain operable when all data columns are hidden.
- Saved-view resource, selection, dirty comparison, permissions, action execution, versioning, invalid-selection fallback, and stale asynchronous guards already reside in cohesive saved-view owners and must remain there.
- Query and filter drafts are surface-keyed, but not saved-view/version-keyed. Surface transitions close query panels; saved-view transitions can leave presentation drafts or panel ownership stale.
- Existing overlay navigation uses stable registered keys with Arrow, Home/End, Escape, and connected-trigger fallback; active query chips do not use that roving model.
- Grid query/layout changes use existing semantic controllers and must not cause an unrelated grid, selection, scroll, draft, or Inspector remount.

## Demonstrated Owner Mismatches

- Readability: the current golden visibly contains repeated `Sort…` controls and clipped saved-view state.
- Keyboard: `docs/design.md` requires one Tab entry for the chip region and filter-chip activation into the editor; current chips are independent removal buttons.
- Filter clear: `docs/design.md` defines clearing all filters; the shared shell callback resets the complete query.
- Sort: `docs/design.md` requires a complete ordered applied list and explicit Add-sort lifecycle; the current panel mixes selected and unused fields.

## Advisory Classification

- ADOPT: R001–R005, R007–R008, and R010–R015.
- ADAPT: R018 keeps horizontal scrolling inside the grid owner; R021 introduces no decorative motion; R022 applies one primary action per local decision locus; R023 retains the current semantic icon family; R035 uses one-row semantic disclosure rather than wrapping.
- REJECT: R026–R034, including the bundled Cybersecurity Platform styling, marketing/dashboard structure, invented behavior, and incidental selectors.
- Offline recommendations: ADAPT generic wrapping/`+n` advice into the owner-required one-row Filters overflow; ADOPT semantic controls, focus management, surfaced async errors, derived projection, and reducer guidance; ADAPT lifted state to transient presentation only; REJECT React Actions, `startTransition`, or broad state rewrites without an owner-backed need.

## Execution Log

- VB-01: Reconfirmed branch, commit, clean state, and toolchain. Created this tracker before product edits.
- VB-01: Retained focused green baseline for saved-view selector, grid controls/model, and responsive layout: `.cartulary/test-results/20260904T001305Z-p86253`.
- VB-01: Added an owner-focused filter-chip activation characterization. The expected-red slice failed at `.cartulary/test-results/20260904T002431Z-p90485` because the current chip dispatched removal instead of opening the editor.
- VB-02: Corrected the presentation owner, added four view-bar allocation tokens, generated their projection, and introduced a pure semantic working-set/query-entry model. Maximum-pressure tests prove `22 / 23 / 25` hidden entries for base, narrow, and compact placement.
- VB-02: `make frontend-typecheck` passed at `.cartulary/test-results/20260904T004400Z-p98424` after resolving projection-boundary errors found at `.cartulary/test-results/20260904T003928Z-p95347` and `.cartulary/test-results/20260904T004306Z-p97143`.
- VB-03: Migrated Sort to applied priority rows plus unused Add-sort fields; Filters to discriminated exact-operator drafts and filter-only clear; active chips to semantic activation/Delete; and saved-view actions to grouped primary/secondary/destructive loci.
- VB-03: Removed the superseded filter-chip selector after migrating application, browser, and shared test-utility consumers. Focused workbook regression passed at `.cartulary/test-results/20260904T005337Z-p70694`.
- VB-03: Added saved-view/version subject scoping and connected semantic focus return for filter, group, and sort chip editors. Roving chip coverage passed at `.cartulary/test-results/20260904T010126Z-p49881`.
- VB-04: Ran both required offline advisory searches with bytecode disabled. Their one-row overflow, semantic control, focus, async-error, and derived-state guidance was classified as recorded above.
- VB-04: Replaced separately projected React nodes with one data-backed working-set binding across shell and surface composition; Timeline's inline fallback now supplies that binding instead of rendering a duplicate control path.
- VB-04: `make browser-e2e-a11y` passed all 12 units at `.cartulary/test-results/20260904T020604Z-p26418`; `make browser-e2e-support` passed all 19 units at `.cartulary/test-results/20260904T014015Z-p63998`.
- VB-04: Promoted dedicated clean/modified identity, saved-view action, maximum Sort, filter-editing overflow, long Columns, base/narrow/compact pressure, 200% zoom, and text-spacing goldens through `make browser-e2e-visual-update` at `.cartulary/test-results/20260904T020328Z-p78090`. Every changed PNG was manually reviewed; narrow/compact collision and zoom allocation defects found during review were corrected before acceptance.
- VB-04: The ordinary visual suite passed twice after promotion at `.cartulary/test-results/20260904T020738Z-p73802` and `.cartulary/test-results/20260904T020935Z-p20692`.
- VB-05: Opened only after VB-04 completed. Owner-routed slices, service-backed slices, terminal checks, final scope audit, and this completion record are complete.
- VB-05: Added direct characterization of the full closure-free working-set projector across saved-view action groups, query/layout descriptors, active filter editing, stable focus targets, fixed control order, and subject identity. The focused row passed at `.cartulary/test-results/20260904T024704Z-p97857`.
- Generation: the final `make generate` passed at `.cartulary/test-results/20260904T024711Z-p98398`; generated changes are limited to the design-token projection and the execution-topology input digest. Visual hashes changed only through the Make-owned visual promotion workflow.

## Design Correction Record

Before: §7.5 admitted eight base and six narrow chips solely by count. The current 1440 by 900 golden visibly compressed saved-view state and multiple sort chips into repeated fragments.

After: §7.5 retains canonical first-N overflow while using three base, two narrow, and zero compact/below-minimum semantic slots with non-truncating `G`, `S<n>`, and `F<n>` identity tokens and full operable descriptions. §8.3 clarifies activation, Delete, roving focus, and the operable overflow path. Query meaning, identity, ordering, and persistence are unchanged.

Compatibility impact: fewer applied-query details remain inline at base and narrow widths, but each visible entry is now distinguishable and every hidden entry remains in the existing Filters overflow. No saved-view or query data changes and no migration or compatibility adapter are required.

Validation evidence: pure-model tests cover zero/one/several and maximum pressure, including canonical tokens and hidden counts of `22 / 23 / 25`; browser geometry and reviewed goldens cover base, narrow, compact, below-minimum, zoom, text spacing, and long content.

## Resulting Behavior

- The view bar remains one non-wrapping, fixed-height row in the owner sequence: saved view, Sort, Group, Filters, Columns, applied-query detail/overflow, Inspector, and Add Row. The right rail protects Inspector and create reachability while saved identity and query detail yield first.
- Saved-view identity, complete `Modified` state, and incident save state are distinct. The saved-view dialog groups create, update/reset, duplicate, startup references, and separated deletion, with Update primary for a mutable selection and Save as new primary otherwise.
- The pure working-set seam contains semantic saved-view/action, sort, group, filter-editor/applied-filter, column, overflow, and focus descriptors. It contains no callbacks, React nodes, selectors, CSS measurements, coordinates, route labels, persistence, or authorization inference. Existing execution ports remain in their query, layout, and saved-view owners.
- Sort exposes ordered priority rows, direction, earlier/later movement, removal, and an unused-field Add-sort locus with the owner maximum and disabled explanation. Group remains a single owner-declared selection. Columns remain usable and resettable with every data column hidden.
- Filter drafts round-trip `eq`, `range`, `contains_any`, `contains_all`, `prefix`, and `full_text` argument shapes. Add/Edit, local validation, Apply/Cancel, single removal, and filter-only clearing are distinct; hidden group/sort entries redirect to their owning controls.
- Query chips have stable kind-plus-`field_key` identities and canonical group/sort/filter order. Exactly one visible chip is tabbable; arrows and Home/End rove, Enter/Space/click opens the owning editor, and Delete clears the focused entry. Focus returns to the invoking chip or the deterministic owning-control fallback.
- Query panels share one subject-scoped transient reducer. Incident, surface, selected-view, or saved-view-version changes close stale panels and discard local drafts. Saved-view async completions retain their existing semantic action identity and cannot update a new subject.
- Timeline uses the same data-backed working-set binding as the generic, entity, and assessment surfaces; the duplicate inline query renderer was removed. Network Flow remains outside this seam.
- Chrome placement depends only on the inline-size band. Root CSS zoom is reflected in the effective inline size, while vertical-only resizing preserves the selected band. Compact/below-minimum actions retain stable accessible names and pointer disclosure.

## Changed Path Groups

- Presentation/model: Workbook view-bar, saved-view, query-control, surface-composition, Timeline-presentation, responsive-layout, focus-navigation, query-model, and focused tests under `apps/web/src/workbook/**` plus the Timeline runtime fixture.
- Contracts: authored allocation tokens in `contracts/design/tokens.v1.json`; authored semantic selectors and exports/tests in `packages/ui-contracts`; generator-produced `packages/ui-contracts/src/generated/design-tokens.ts`.
- Browser evidence: Workbook keyboard, accessibility, state/geometry sentinel, visual fixtures, shared test actions, 21 updated existing PNGs, and 11 new dedicated PNGs.
- Verification inputs: authored frontend ownership and `web.workbook` family rows; generator-produced execution-topology digest; Make-promoted visual-golden manifest.
- Documentation: the narrow `docs/design.md` correction and this handoff. No Core owner was edited.

## Acceptance Evaluation

| IDs | Evaluation and evidence |
| --- | --- |
| A001–A003 | ADOPT. The owner/change map above separates the design correction, pure model, execution ports, and tests. Baseline, repository map, generated policy, direct dependency cone, and final Git scope were inspected. |
| A004–A005 | ADOPT. Allocation uses four authored `--ct-layout-viewBar*` tokens and existing spacing, focus, typography, component, motion, and dark-graphite tokens. No second registry, local theme, or Cybersecurity Platform palette was introduced. |
| A008–A009 | ADOPT. Responsive unit/browser evidence covers visual-viewport fallback, base/narrow/compact/below-minimum, vertical-only resize, 200% zoom, text spacing, protected right-rail actions, safe top navigation, and no shell/document horizontal overflow. Grid scrolling remains grid-owned. |
| A011 | ADOPT. Stateful and webserver-backed suites retain semantic grid selection, active cell, scroll, drafts, latest-query, and Inspector continuity while controls change. |
| A014 | ADAPT. The global grid editing contract remains unchanged; this slice adds deterministic chip keyboard activation/Delete, roving navigation, overlay Escape/focus return, filter Apply/Cancel, and draft retention/discard tests. |
| A016–A017 | ADOPT as preserved behavior. No loading/empty/unauthorized/unavailable/refresh owner was changed; resource, pending, error, stale completion, and prior-row behavior remain covered by owner-routed and webserver-backed suites. |
| A019–A020 | ADOPT. Stable names, non-color `Modified`, semantic busy/disabled states, focus visibility/return, reduced motion, action-group variants, filter operators, maximum Sort, overflow, and all-hidden Columns are covered by unit, accessibility, and visual matrices. |
| A021 | ADOPT as preserved boundary. Virtualization/vendor integration was untouched; stateful, measurement, and webserver-backed continuity suites pass. |
| A022 | ADOPT. Exact registered visual fixtures ran through the ordinary suite and Make-owned promotion. Every changed PNG was manually inspected after promotion. |
| A023–A024 | ADOPT. Runtime/test identity uses fixture IDs, incident, `view_schema_id`, saved-view identity/version, `field_key`, panel/action IDs, and query-entry kind. Runtime/tests do not read Markdown or bind behavior to document text, DOM hierarchy, CSS class, component name, or vendor coordinate. |
| A025 | ADOPT. Authored token/family/ownership inputs produced generated design-token/topology outputs through `make generate`; drift and generated-artifact policy checks pass. |
| A026 | ADOPT. No route, payload, schema, field capability, query grammar/operator, write, authorization, lifecycle, storage, evidence, compatibility alias, migration, or dependency was invented. |
| A027 | ADOPT. This record contains authority, before/after behavior, paths, commands, results, failure resolution, compatibility, rollback, risks, deferral, and the next safe seam. |

## Verification and Failures

- Expected-red characterization: `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.grid_controls_component_a104000002`; failed as expected at `.cartulary/test-results/20260904T002431Z-p90485`.
- `make format`: initial product-lint failures at `.cartulary/test-results/20260904T005002Z-p49463` and `.cartulary/test-results/20260904T005024Z-p53841` were corrected; passed at `.cartulary/test-results/20260904T005043Z-p58176`, `.cartulary/test-results/20260904T005145Z-p63685`, and `.cartulary/test-results/20260904T010050Z-p44950`.
- `make frontend-unit`: intermediate product failures at `.cartulary/test-results/20260904T004454Z-p99895` and `.cartulary/test-results/20260904T005644Z-p7475` identified stale selector/capacity assertions and two layout/ownership-policy violations; all were corrected without weakening coverage. The final suite passed 432/432 at `.cartulary/test-results/20260904T022508Z-p52649`.

### Owner-routed verification

- Focused slices: `web.workbook` 160/160 at `.cartulary/test-results/20260904T021238Z-p71764`; `web.design` 15/15 at `.cartulary/test-results/20260904T021323Z-p89112`; `package.ui` 10/10 at `.cartulary/test-results/20260904T021425Z-p37781`; `package.view_contracts` 5/5 at `.cartulary/test-results/20260904T021428Z-p38177`; `module.workbook` 69/69 at `.cartulary/test-results/20260904T021430Z-p38574`; `module.savedviews` 26/26 at `.cartulary/test-results/20260904T021642Z-p98387`; `harness.browser` 28/28 at `.cartulary/test-results/20260904T021748Z-p50126`.
- Service-backed slices: `web.design` 15/15 at `.cartulary/test-results/20260904T021848Z-p69456`; `module.workbook` 39/39 at `.cartulary/test-results/20260904T021950Z-p18168`; `module.savedviews` 24/24 at `.cartulary/test-results/20260904T022157Z-p75314`; `harness.browser` 6/6 at `.cartulary/test-results/20260904T022302Z-p26865`.
- The final full-working-set model row passed 2/2 at `.cartulary/test-results/20260904T024704Z-p97857` after the broader owner slice; final frontend typecheck passed at `.cartulary/test-results/20260904T024638Z-p96999`.

### Terminal verification

- `make agent-finalize` passed against the final implementation at `.cartulary/test-results/20260904T024859Z-p2436`. `RESULTS_DIR` was intentionally unset because no retained successful full warm run from the exact source was supplied; retained-run maintenance was skipped.
- Generation/policy: `make generate` passed at `.cartulary/test-results/20260904T024711Z-p98398`; `make generate-drift` at `.cartulary/test-results/20260904T024918Z-p5455`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260904T024932Z-p8589`; `make json-shape-check` at `.cartulary/test-results/20260904T024936Z-p9040`.
- Frontend: `make frontend-typecheck` at `.cartulary/test-results/20260904T024638Z-p96999`; final `make frontend-unit` 432/432 at `.cartulary/test-results/20260904T024950Z-p10070`; `make frontend-import-boundary-check` at `.cartulary/test-results/20260904T022457Z-p51751`; final `make lint-biome` at `.cartulary/test-results/20260904T024942Z-p9569`.
- Browser: `make browser-e2e-a11y` 12/12 at `.cartulary/test-results/20260904T020604Z-p26418`; `make browser-e2e-measurement` 22/22 at `.cartulary/test-results/20260904T022606Z-p72215`; `make browser-e2e-stateful` 34/34 at `.cartulary/test-results/20260904T023034Z-p31829`; `make browser-e2e-support` 19/19 at `.cartulary/test-results/20260904T023253Z-p83088`; `make browser-e2e-webserver-backed` 60/60 at `.cartulary/test-results/20260904T023416Z-p29099`.
- Visual: final Make-owned update passed at `.cartulary/test-results/20260904T020328Z-p78090`; ordinary `make browser-e2e-visual` passed twice at `.cartulary/test-results/20260904T020738Z-p73802` and `.cartulary/test-results/20260904T020935Z-p20692`.
- Broad/static: final `make test-fast` passed 484/484 at `.cartulary/test-results/20260904T025142Z-p51110`; `make lint-markdown` passed at `.cartulary/test-results/20260904T025159Z-p51967`; `git diff --check` passed.

### Failure classification and resolution

- Expected product mismatch: filter-chip activation removed the filter instead of opening its exact editor at `.cartulary/test-results/20260904T002431Z-p90485`; remediated with semantic activation and Delete-only clearing.
- Implementation/type failures: projection boundaries at `.cartulary/test-results/20260904T003928Z-p95347` and `.cartulary/test-results/20260904T004306Z-p97143`; a missing React type at `.cartulary/test-results/20260904T013416Z-p72270`; and a test DOM type at `.cartulary/test-results/20260904T015534Z-p72775`. All were product-code corrections; final typecheck is green.
- Focused unit failures: stale selectors/capacity assertions and layout ownership at `.cartulary/test-results/20260904T004454Z-p99895` and `.cartulary/test-results/20260904T005644Z-p7475`; the initial zoom-width calculation disturbed normal jsdom layout at `.cartulary/test-results/20260904T015953Z-p31524`. Assertions were migrated to semantic identities and the responsive fix was limited to explicit root zoom.
- Browser harness/expectation failures: a parallel support-build artifact collision at `.cartulary/test-results/20260904T011102Z-p79426` was rerun serially; stale accessibility assertions at `.cartulary/test-results/20260904T013701Z-p15397` were updated to the owner behavior. Final support and accessibility targets pass.
- Intentional visual diffs: the pre-promotion ordinary suite at `.cartulary/test-results/20260904T011740Z-p24257` detected the expected presentation changes. An intermediate visual run at `.cartulary/test-results/20260904T014607Z-p66766` exposed a filter Add-to-Edit subject-key defect; `.cartulary/test-results/20260904T015623Z-p78302` exposed unreachable zoom actions. Both product defects were corrected before the final promotion and two ordinary green runs.
- A temporary local contact-sheet output overwrote tracked PNG paths during manual inspection; the Make-owned update at `.cartulary/test-results/20260904T012313Z-p18003` restored canonical captures before final promotion. No contact-sheet artifact remains.

## Final Scope Audit

- The final diff is confined to the allowed Workbook frontend dependency cone, authored design/UI/test contracts, their generated projections/manifests, browser fixtures/goldens, `docs/design.md`, and this handoff.
- No Network Flow runtime or golden changed. No route, API/HTTP payload, database/storage/migration, authorization, saved-view lifecycle/version semantics, query grammar, field capability, grid vendor/virtualization, Inspector internals, row-creation behavior, mutation recovery, dependency, theme, or lockfile changed.
- No runtime or test reads Markdown. The pure model contains semantic data only; renderers derive selectors from stable semantic IDs. No temporary adapter, compatibility alias, selector-driven runtime identity, duplicate Timeline control path, feature flag, or TODO was introduced.
- Generated differences are the design-token projection and execution-topology digest from authored inputs; the golden manifest was written by the Make-owned visual workflow. All 21 changed existing goldens and 11 new dedicated goldens were manually reviewed.
- No Git commit was created.

## Compatibility, Rollback, Risks, and Deferrals

- Compatibility: no route, payload, query grammar, saved-view persistence, authorization, version, startup-pointer, grid-vendor, dependency, or stored-data boundary changed.
- Rollback: revert the presentation/model/selectors, the narrow design/token correction, generated projections, browser assertions, and reviewed goldens as one change. No data rollback is needed.
- Risk: the reduced inline detail changes screenshots by design; full semantics remain in accessible names, hover/focus disclosure, editors, and overflow.
- Deferred next seam: grid empty-state presentation remains outside this refactor.

“Workbook view-bar presentation, local interaction, and internal component architecture were refactored. Core saved-view, query, layout, route, authorization, persistence, versioning, and stored-data behavior remain unchanged. Any design-contract correction is limited to deterministic readable overflow presentation.”
