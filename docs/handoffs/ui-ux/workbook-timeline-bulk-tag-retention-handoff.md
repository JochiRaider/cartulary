# Retained Timeline bulk-tag authoring

## Control, baseline and authority

Baseline: clean `main` at `ab658d413fb6e62d7719b2e40b53fbe7f5649016`,
one commit ahead of `origin/main`. Root AGENTS.md applies. This slice adopts
and implements only retained bulk-tag authoring within one account, incident and
client-instance runtime. The landed date-filter slice is outside this work.

Core 03 REQ-03-297 now explicitly adopts exact raw-text/revision retention across
presentation detachment, zero-selection visibility, explicit clearing, no implicit
submission, authority concealment/resumption and terminal retirement. REQ-03-299/100,
§13.3 REQ-03-221/222, Core 01 §3.3.5 and Core 04 §§1–2 retain their existing owners.
REQ-03-298 is ordinary-cell precedent, not the source of this extension. No owner
contradiction was identified. Domain owns vocabulary; design supplies bounded
presentation direction. Research, bundled prompts, digest maps and historical
handoffs were consulted as source material, not additional authorization.

The older bulk-tag interaction-cleanup handoff's mounted raw-authoring lifetime
is historical and superseded by this explicit REQ-03-297 amendment. Its independent
selection, batch, native-input and render-isolation decisions remain relevant.

| Workstream | State | Exit |
| --- | --- | --- |
| TR-01 Characterization/authority | COMPLETE | Seeded navigation loss observed, narrow owner amendment adopted, failing focused regression recorded. |
| TR-02 Retained authoring | COMPLETE | Focused unit/shell/security checks and browser proof. |
| TR-03 Rendered/terminal verification | COMPLETE | Current-build review, cleanup and terminal checks. |
| TR-04 Handoff | COMPLETE | Evidence ledger and complete digest assessment. |

## Boundary and behavior

Before: select records, enter a tag, switch to Hosts, return; selection and the
control disappear, Recovery stays empty, and reselection reveals empty text.
After: the exact text returns with zero selection and an explanation to select
records. Only a fresh explicit selection and submission can assign it. Returning
never focuses the input or creates a request.

| Concern | Owner |
| --- | --- |
| Exact raw text/revision | TimelineBulkTagAuthoring, retained by WorkbookTimelineMutationOwner. |
| Current query selection/readiness | Existing Timeline controller, planner and prerequisite readers. |
| Focus/caret/DOM/local feedback | Mounted TimelineBulkTagControl; feedback matches revision, authority generation and selection. |
| Admitted operation | WorkbookBatchOperationOwner: captured identity/targets/bytes, prerequisites, replay, outcomes and read recovery. |

Foundation, interaction composition, root composition and presentation bindings
carry one stable authoring capability. The leaf subscribes directly. Its mounted
raw-text/revision owner is removed; there are no synchronized copies. The source
slot retains no rows, field identities or target versions. Authoring publications
never notify aggregate mutation status. Selection pruning is unchanged.

The small Timeline-specific binding owns the lifetime decision: presentation
commands require a current attachment token and authority generation. Suspension
conceals without destruction. Readable role loss/closure prevents assignment;
account replacement, incident change and disposal permanently retire the value.
Existing source ownership is web.workbook, independent of module.timeline and
web.workbook verification routing. No grid-vendor integration was added.

Clear advances the authoring revision, clears local feedback and explicitly
returns focus to the mounted input. It preserves valid selection, cell drafts
and admitted batches. Admission/acknowledgement do not clear text. Newer or cleared
work is insulated from late outcomes; the originating operation retains its independent batch results and recovery.

The extension is deliberately one source-owned value and narrow bindings. Future
presentations could borrow this same Timeline capability; no general form or draft
registry is introduced. External request/receipt contracts and stored data remain
unchanged. **No persisted-data migration is required.** No reload, crash, cross-tab,
cross-incident or durable retention guarantee is added.

Advisory treatment: R002/R006/R008/R013/R014 ADOPT for owner-required focus,
local explanations, semantic controls and continuity; R012 ADAPT state placement
through retained ownership; R018 ADAPT overflow through workbook scroll ownership.
A targeted offline React query returned three results without fallback. Modal
focus trapping and framework Actions examples do not apply to this native form.

## Evidence ledger

Paths below are repository-relative retained synthetic evidence, not private
session credentials, raw observations or transient screenshot links.

| Target/action | Result | Run root / disposition |
| --- | --- | --- |
| Planning controller/planner slice | PASS 3/3 harness units | `.cartulary/test-results/20261001T205157Z-p86820`; does not prove navigation. |
| Seeded baseline navigation | Reproduced loss; cleanup PASS | `.cartulary/test-results/20261001T210835Z-p92151`; closed receipt and foreground exit observed. |
| Focused detach regression before production edits | Expected FAIL, 5/6 assertions pass | `.cartulary/test-results/20261001T211230Z-p21479`; returning textbox absent. |
| Initial frontend-typecheck | PASS 2/2 | `.cartulary/test-results/20261001T211505Z-p23238`. |
| Initial generation | PASS | `.cartulary/test-results/20261001T211748Z-p24559`. |
| Expanded Timeline controller/planner | PASS 3/3 harness units | `.cartulary/test-results/20261001T211807Z-p27517`. |
| Shell/runtime continuity/responsibility/batch rows | PASS 5/5 harness units | `.cartulary/test-results/20261001T211807Z-p27519`. |
| Browser routing generation | Corrected authoring failure | `.cartulary/test-results/20261001T211955Z-p29445`; added titles lacked paired scenario IDs, repaired in authored family inputs. |

| Browser routing generation, second attempt | Corrected authoring failure | `.cartulary/test-results/20261001T212124Z-p32826`; titles required ASCII sorting with paired scenario IDs. |
| Corrected generation | PASS | `.cartulary/test-results/20261001T212232Z-p35947`; generated render index only. |
| Format and initial finalize | PASS | `.cartulary/test-results/20261001T212343Z-p39084`, `.cartulary/test-results/20261001T212447Z-p43779`. |
| Fresh Timeline and shell/runtime slices | PASS 3/3 and 5/5 | `.cartulary/test-results/20261001T212753Z-p48578`, `.cartulary/test-results/20261001T212753Z-p48587`. |
| Frontend typecheck correction | FAIL then PASS | `.cartulary/test-results/20261001T212753Z-p48697`; browser test passed an unsupported argument to saveStateTestId; corrected, `.cartulary/test-results/20261001T212858Z-p83234`. |
| Browser slice, first execution | FAIL 9/11 harness units | `.cartulary/test-results/20261001T212802Z-p50518`; seven scenarios passed, delayed acceptance reached retained B then expected unsorted targets. Assertion repaired to match existing planner ordering. |
| Corrected six browser rows | PASS 11/11 harness units, eight scenarios | `.cartulary/test-results/20261001T213119Z-p27318`; real shell navigation, prerequisites/rejection, membership, replay/read recovery and layouts. |
| Biome correction | FAIL then PASS | `.cartulary/test-results/20261001T212907Z-p84429`, `20261001T213119Z-p27443`, `20261001T213151Z-p79206`; 14 test non-null assertions replaced with explicit guards; PASS `.cartulary/test-results/20261001T213323Z-p90086`. |
| Diagnostic invocation | Rejected input, no product execution | `make lint-biome VERBOSE=1` was rejected as an undeclared Make input. Environment output mode obtained diagnostics instead. |
| Frontend import boundaries | PASS 2/2 | `.cartulary/test-results/20261001T212907Z-p84409`. |
| Generated-artifact policy | PASS 3/3 | `.cartulary/test-results/20261001T212907Z-p84229`. |
| Generation drift and JSON shapes | PASS 4/4 and 3/3 | `.cartulary/test-results/20261001T212935Z-p6131`, `.cartulary/test-results/20261001T212935Z-p6148`. |
| Formatting and finalize after corrections | PASS | `.cartulary/test-results/20261001T213244Z-p84906`, `.cartulary/test-results/20261001T213313Z-p89404`; scoped diff inspected. |
| Guarded Timeline slice and frontend typecheck | PASS 3/3 and 2/2 | `.cartulary/test-results/20261001T213324Z-p90882`, `.cartulary/test-results/20261001T213338Z-p94523`. |
| Review startup during formatting | FAIL, cleanup complete | `.cartulary/test-results/20261001T212906Z-p84083`; build/source_snapshot/invalid_artifact. Finished changes before restarting; exact stop retained the failure and foreground exited. |
| Review startup interrupted for test repairs | FAIL, cleanup complete | `.cartulary/test-results/20261001T213121Z-p27943`; seeding/fixture_seed/child_failed. Exact stop and foreground exit confirmed; subsequent serialized startup succeeded. |
| Seeded production review | Reviewed four images, cleanup complete | `.cartulary/test-results/20261001T213343Z-p95068`; operations 12/21/30/33; SHA/byte-verified original images inspected at both requested sizes. |

| Final Clear/cell-draft assertion slice | PASS 3/3 harness units | `.cartulary/test-results/20261001T214322Z-p41008`; Clear retains independent cell revision/value and admitted A while late outcomes settle. |
| Final format / finalize | PASS 2/2 and 1/1 | `.cartulary/test-results/20261001T214017Z-p31824`, `.cartulary/test-results/20261001T214137Z-p36547`; RESULTS_DIR unset. |
| Final frontend-typecheck / lint-biome | PASS 2/2 each | `.cartulary/test-results/20261001T214322Z-p41160`, `.cartulary/test-results/20261001T214322Z-p41217`. |
| Markdown documentation check | PASS | `.cartulary/test-results/20261001T214322Z-p41246`, final handoff `.cartulary/test-results/20261001T214911Z-p73537`. |
| Final sealed navigation/console review | Reviewed two images; cleanup PASS | `.cartulary/test-results/20261001T214324Z-p42970`; operations 16/19, exact original bytes/SHA checked and images inspected. |
| Final scope review / git diff --check | PASS | Same main/HEAD; 22 tracked edits and three new files, all within this slice; no unrelated changes, commits or persisted-data changes. |

No verification remains blocked. Earlier failures and corrections above are retained
rather than reclassified as passes. Product browser proof comes from executable
assertions, separately from rendered observations.
Captures and Axe are advisory review evidence, not accessibility certification,
product-test passes or Core 05 publication evidence.


## Changed and inspected paths

The affected production path is:
`timeline/ports/TimelineBulkTagAuthoringPort.ts` →
`timeline/bulk/TimelineBulkTagAuthoring.ts` →
`timeline/mutations/WorkbookTimelineMutationOwner.ts` → runtime lifecycle forwarding.
The existing foundation/interaction/workbook compositions carry the capability to
`timeline/components/TimelineBulkTagControl.tsx` through the existing presentation.
These paths are under `apps/web/src/workbook`. The controller still owns selection;
readiness, planner, command adapter/port, batch owner and query browser were
revalidated and retain their existing contracts.

Changed regressions are `timeline/timelineBulkTagInteraction.test.tsx`,
`WorkbookShell.surfaces.test.tsx`, `runtime/WorkbookRuntimeResponsibilities.test.ts`
and `apps/web/e2e/timeline-bulk-tag.spec.ts`. Existing planner/runtime continuity/
batch retention suites were exercised without broadening their product contracts.
Supporting edits are the applicable Timeline/runtime READMEs, Core 03 owner,
`docs/design.md`, source ownership and the two test families plus generated index.
Domain, Core 01/04, digest and historical cleanup handoff were inspected, not edited.

## Reproduction, rendered assessment and execution

Discovery used `make help`, `make help-all`, and
`make task-guide ROLE=module-author OWNER=module.timeline` / `OWNER=web.workbook`.
Confirmed existing row IDs were reused; new test titles/scenario pairs extend
those semantic selectors through authored family inputs, then `make generate`.

Focused commands executed from the repository root:

```bash
make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_bulk_tag_controller_8a15df23b4,module.timeline.frontend.timeline_bulk_tag_plan_a111000001
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e,web.workbook.regression.workbookz_mutation_runtime_responsibilities_a106000001,web.workbook.regression.workbookz_mutation_runtime_surface_continuity_97808fe8ae,web.workbook.regression.batch_operation_retention
make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.bulk_tag_characterization,module.timeline.browser.bulk_tag_membership,module.timeline.browser.bulk_tag_prerequisites,module.timeline.browser.bulk_tag_rejection,module.timeline.browser.bulk_tag_replay,module.timeline.browser.bulk_tag_layouts
```

Rendered review used `make ui-review UI_MODE=seeded`, repository synthetic editor
and BROWSER-REVIEW incident, and only public `ui-browser`, `ui-capture`, status and
exact-locator stop commands. At 1440×900 and 768×640, selected and zero-selection
controls stay compact and readable. The raw native input can scroll its long text;
its focus outline and both buttons remain reachable. The zero-selection message
occupies its own row without hiding the grid or status strip. Accepted-result
feedback remains readable at the smaller size. Tab/Shift+Tab reaches Assign;
Enter explicitly submits. At zero selection Tab skips disabled Assign and reaches
Clear; Enter clears and returns input focus. Small-screen Surfaces menu navigation
also returns without recreating the cleared control; reselection finds empty text.

All four reviewed captures completed Axe with zero violations and one incomplete
color-contrast rule each. This is an advisory scan of rendered nodes in the main
document, not contrast certification or a broad accessibility pass. Captures were
consumed before stop; private links expired and are intentionally absent here.
The session's terminal receipt confirms closed/ok/cleanup complete, four images,
zero failed requests and foreground exit 0. Caller-owned scratch requests removed.
Final sealed review after the added cell-draft assertion repeated navigation at
both sizes. Its initial unauthenticated document recorded HTTP 401 and HTTP 404
resource errors; authenticated directory, workbook actions, navigation and captures
recorded no console errors. The initial resource observations are outside this
bulk-tag change and were left untouched. The final receipt records two images,
two incomplete contrast observations, zero failed requests, closed/ok and complete
cleanup; foreground exit 0 and caller scratch removal confirmed.

Quality/generation commands actually executed were `make generate`,
`make generated-artifact-policy-check`, `make generate-drift`, `make json-shape-check`,
`make frontend-typecheck`, `make lint-biome`, `make frontend-import-boundary-check`,
`make lint-markdown` and `git diff --check`; their evidence is in the ledger.

`make format` preceded diff inspection and `make agent-finalize` before broader
terminal checks. No qualifying successful full warm-check run exists:
**retained-run maintenance skipped because RESULTS_DIR was unset**. Broad backend,
release, full warm checks, unrelated goldens, date filters and package.ui were not
rerun: the slice changes no backend/data contracts, grid/selector contracts or
those independent owners. No new dependencies or persisted migration.

## Acceptance and rollback

Applicable digest rows are assessed below against this bounded slice. N/A does
not assert compliance for untouched workflows. No historical pass substitutes for
fresh product evidence.

| Row | Assessment | Evidence / scope |
| --- | --- | --- |
| A001 | PASS | REQ-03-297 owner amendment; source ownership web.workbook versus module.timeline/web.workbook test routing. Core 01/04 and REQ-03-299/100 preserved. |
| A002 | PASS | One retained Timeline value owns the detachment decision; authoritative raw state moved, all callers migrated, mounted copy retired. Stable leaf subscription permits future Timeline presentations without a generic registry. Runtime-only limitation and rollback recorded. |
| A003 | PASS | Clean baseline/HEAD and final scoped diff; local runtime/Timeline guides, authored source ownership/import rules, generated policy and independent family routing inspected. No new direct vendor imports or localization strings outside the existing source layer. |
| A004 | PASS | Control styles unchanged; existing ct tokens continue to supply spacing, color, border, rounded controls and responsive inline sizing. |
| A005 | PASS | Current dark_graphite seeded renderer inspected; no palette/theme behavior or fixture changed. |
| A006 | PASS | Existing bulk_tag_layouts row freshly exercises all three densities; shared grid/density ownership unchanged, control remains usable. No claim about untouched measurement fixtures. |
| A007 | N/A | No record creation, contextual authoring/source replacement or create capability change. |
| A008 | PASS | Focused layout row and manual 1440×900/768×640 production review; existing responsive chrome retained. Boundary/fallback/inspector clamp algorithms untouched, so their broader fixtures were not rerun. |
| A009 | PASS | Both reviewed sizes keep grid scrolling inside shell and retain status, Recovery, surface and account navigation. Layout row checks document overflow. |
| A010 | N/A | No inspector dispatcher, feature review or confirmation changes; authoring retention belongs to the Timeline command, not inspector identity. |
| A011 | PASS | Unit/shell/browser detach-return and late-outcome scenarios retain exact text while selection resets/prunes; no auto-selection, focus or retargeting. |
| A012 | PASS | Existing random-ID/captured operation assembly unchanged; controller delivery fencing, fresh current versions and exact replay row pass. Delayed acceptance preserves A request/targets while B survives. |
| A013 | PASS | Batch retention and replay/read-recovery rows pass; Clear does not cancel admitted work. No queue Retry/Discard behavior changed. |
| A014 | PASS | Exact whitespace/Unicode, failed/unsubmitted blockers, native caret/input continuity, Clear, detachment and authority-lifetime tests; no durable persistence. |
| A015 | N/A | No cell conflict locus or saved/local distinction changed. Failed-selected-edit browser row verifies existing conflict still blocks the command. |
| A016 | PASS | Current accepted membership/versions and authoring authority remain independent. Readable viewer/closure blocks edits and assignment; suspended content is concealed. No query-state projection change. |
| A017 | PASS | Runtime tests cover suspension/resumption, role/closure, replacement, incident change, disposal, instances and obsolete bindings; DOM suspension and shell return covered. |
| A018 | N/A | No evidence lifecycle, overlay or preview change. |
| A019 | PASS | Native input, semantic form/name, local status/alert, keyboard Assign/Clear and explicit Clear focus; no autofocus on return/settlement. Axe observations advisory, contrast incomplete; no conformance certification asserted. |
| A020 | PASS | Fresh layouts/rejection rows cover density, long raw content, zoom/text spacing and overflow; reviewed production zero/selected states and long accepted-result feedback. |
| A021 | N/A | No virtualization, row identity, focus/scroll algorithm or fake rows. Current-query membership/pruning browser row remains fresh evidence for the affected selection boundary. |
| A022 | PASS | Production renderer manually reviewed with sealed source, dark_graphite/compact and explicit viewports. No relevant golden changed or refreshed; artifacts remain implementation support. |
| A023 | PASS | Existing ui-contract selectors and semantic roles/names used; stable record_id/view_schema_id assertions. No selector contract change requiring package.ui tests. |
| A024 | PASS | Source/test diff audit: no executable dependency on docs/digest/Markdown. Runtime and verification use typed source/manifests only. |
| A025 | PASS | Authored families/source ownership updated; make generate produced render-index hashes. Policy, shape and drift checks pass; no generated file hand edits. |
| A026 | PASS | Adopted amendment bounds the behavior; no external request/receipt/storage/schema change. One mounted state retired; explicit no persisted-data migration and revert rollback. |
| A027 | PASS | Owner/scope/compatibility, failure ledger, fresh unit/runtime/browser evidence, rendered assessment/cleanup, skipped checks, rollback and completed workstreams recorded. |

Rollback reverts this slice's owner/design amendments, source bindings and slot,
tests, authored routing/ownership inputs and their generated projections together.
It needs no persisted-data conversion and must not undo accepted incident records.
No new storage, dependencies, routes, schema, backend tag semantics, selection
restoration or broad Recovery workflow was added. Remaining limitation: retention
ends with this browser runtime and its account/incident authority lifetime.

No remaining implementation action is required. Review this scoped diff and handoff;
retention remains memory-only within the adopted runtime/security lifetime.
