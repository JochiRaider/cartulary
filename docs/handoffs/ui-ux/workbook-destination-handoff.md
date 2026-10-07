# Workbook destination handoff

Status: remediation complete on 2026-10-07. The fresh full check passed 1,014/1,014
units, all 21 original failures are resolved, and final test-only maintenance
verification passed. The historical root-cause analysis and failed run remain
below, followed by current remediation evidence. This is a
bounded implementation record, not behavioral authority or a general UI audit.

## Scope and authority

Before the original handoff implementation, `main`, `HEAD` and `origin/main` were revalidated at
`69184ac184db412a2401a3e09c79f5f7554180c3`; the working tree was clean. No commit,
deployment, specification change, dependency change or unrelated cleanup was made
in that original pass. Remediation started from its uncommitted changes and
preserves that work. The specification amendments made during remediation are
identified separately below.

| Decision | Adopted behavior owner | Source boundary | Verification owner |
| --- | --- | --- | --- |
| Work detaches when an admitted destination takes over | Core 03 REQ-03-311, REQ-03-314; Design §§7–8 | Work controls, auxiliary dock and surface layout | `web.workbook`, `module.workbook` |
| Record selection, visible-field/root focus and explicit Inspect | Core 03 REQ-03-314, §13; Core 04 AC-589 | Existing browsing registry, source selection callback and GridHandle; inspector presentation | `web.workbook`, `package.grid_adapter`, `module.workbook` |
| Pending reads retain origin; accepted navigation and focus have separate lifetimes | Core 03 REQ-03-314–315, §14.9; Core 04 AC-581–586 | Workbench admission and session navigation | `web.workbook`, `module.workbook` |
| Pins, captured query/layout and Return bounds remain unchanged | Core 03 REQ-03-313–315; Core 04 AC-583, AC-585–588 | Existing session descriptors, query browser and saved-view owner | `web.workbook` |
| Retained authoring, authority and presentation are independent | Core 03 REQ-03-298/299/100; Core 04 REQ-04-169/170, AC-572/584/592/596 | Existing authoring/recovery owners; navigation cancellation and shell interaction capture | `web.workbook`, `module.workbook` |
| Bounded authorized location remains source-owned | Core 01 REQ-01-680–684; Core 04 REQ-04-170 | Locator port, adapter and Workbook route, inspected without changes | Existing locator ownership unchanged |

The maintained digest, domain vocabulary, source guides and historical workbench
handoffs were consulted as navigation and supporting evidence. The digest's
September localization predates this baseline; current manifests, import
boundaries, source files and routing were rechecked. React/TypeScript and the
existing Grid Adapter remain the implementation stack. No direct grid-vendor
import, second theme/token registry or dependency was introduced.

The common decision is when an admitted destination can attach selection and
focus after auxiliary presentation has detached. It belongs between the existing
navigation, layout and GridHandle owners. An explicit transient handoff in the
existing browsing registry removes competing callbacks without creating another
navigation store or checkpoint cache. A future existing-owner destination can
use the same readiness and acknowledgment seam; no speculative entry API was
implemented. The authorized behavior correction is visible destination handoff,
including the previously missing semantic root fallback.

## Characterization and implementation

A fresh seeded `make ui-review` at
`.cartulary/test-results/20261007T104620Z-p80015` reproduced the baseline at
1024×768 before implementation. A committed Timeline pin opened from Notes
selected a committed cell, but Work still covered the destination and actual
focus ended on the Work heading. With all data columns hidden, Timeline loaded
without grid focus. Work coordination entry and Return paths also retained the
obscuring panel. The baseline admission slice passed; it did not prove rendered
handoff. The baseline review was stopped and cleaned up.

Source inspection identified both competing mechanisms: overlay presentation
made the grid inert, so a semantic focus request could be rejected; Work's portal
remount unconditionally focused its heading. The all-hidden-field early return
also skipped source selection and semantic root focus. Focused tests now establish
inert rejection versus actual root acknowledgment, and the ordered handoff below;
a selected-cell style is never used as proof of focus.

1. Record/Return admission validates the bounded accepted read while retaining the
   eligible origin. Failure, outside-query and unavailable results keep local
   recovery and do not detach Work.
2. On admission, the existing dock/layout owner detaches presentation and the
   source editor detaches without submitting or discarding retained authoring.
   Ordinary explicit surface entry keeps its existing loading/recovery owner.
3. The registry waits for the matching accepted producing request, committed
   presentation readiness and a registered semantic grid handle. Target-group
   reveal still uses the existing source callback.
4. Source-owned record selection commits before focus. Focus tries the requested
   visible field, then eligible visible fields, then the existing semantic root.
   Completion requires GridHandle's actual focus acknowledgment. Exhausted or
   disappeared targets get generic unavailable feedback.
5. Only explicit Inspect opens Record after grid acknowledgment. It then waits for
   the matching inspector registration and acknowledges actual focus through the
   inspector's existing Close control ref. Return never reopens an editor.

An admitted session intent remains cancellable after its read has settled.
New deliberate interaction, supersession, unmount and authority/account changes
invalidate delayed selection, mounting, focus and inspection. Accepted Return
push/pop effects remain independent: later legitimate focus cancellation does
not roll back accepted navigation or invent new trail entries.

Work focuses its heading once per opening, including across portal remounts.
Navigation closure suppresses origin-focus restoration. Ordinary Escape closes
the dock through its existing presentation callback; a child that already
consumed Escape still wins. Adjacent Work also closes on destination activation,
which avoids retaining a competing attachment at every tested width.

## Changed paths and compatibility

All paths in this section are repository-relative.

| Changed files | Responsibility |
| --- | --- |
| `apps/web/src/workbook/navigation/WorkbookSessionNavigation.ts`, `useWorkbookWorkbench.ts`, `WorkbookWorkbenchContext.tsx` | Admission-scoped cancellation, presentation detachment, explicit inspector opening/focus registration |
| `apps/web/src/workbook/query/WorkbookQueryBrowsingContext.tsx` | Existing registry's transient handoff, accepted-observation gate, selection commit, cancellable semantic field/root focus |
| `apps/web/src/workbook/hooks/useWorkbookSemanticGridFocus.ts` | Grid registration, source selection acknowledgment and ordinary entry readiness |
| `apps/web/src/workbook/layout/WorkbookAuxiliaryDock.tsx`, `WorkbookSurfaceLayout.tsx` | Presentation detachment/readiness and unconsumed Escape |
| `apps/web/src/workbook/navigation/WorkbookWorkPanel.tsx`, `apps/web/src/workbook/WorkbookShell.tsx` | Once-per-opening Work focus, coordinated closure and precise deliberate-interaction cancellation |
| `apps/web/src/workbook/components/WorkbookViewBar.tsx`, `apps/web/src/workbook/inspector/presentation/WorkbookInspectorShell.tsx` | Migrated inspector registrations and actual inspector focus acknowledgment |
| `apps/web/src/workbook/navigation/useWorkbookWorkbench.test.tsx`, `WorkbookSessionNavigation.test.ts`; `apps/web/src/workbook/hooks/useWorkbookSemanticGridFocus.test.tsx`; `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`; `packages/grid-adapter/src/semanticFocusRequest.test.ts` | Admission, ordering, root fallback, lifecycle, rendered shell and compatibility regressions |
| `apps/web/e2e/workbook-destination-handoff.spec.ts` | Six service-backed rendered destination regressions |
| `tools/test_families/web.workbook.json`, `tools/test_families/module.workbook.json` | Authored test registration |
| `tools/browser_e2e_batch_manifest.json`, `tools/execution_topology_render_index.json` | Generated routing, produced by `make generate` |

Retired paths are the workbench's separate inspector-close registry, unconditional
Work remount focus, the no-field early return, and broad container-level
navigation exemptions that could preserve obsolete focus work. Affected callers
now use the existing presentation and inspector owners. Candidate discovery
controllers, Recovery execution and source mutation owners were not merged or
rewritten.

The query browser, locator port/adapter and `internal/modules/workbook/routes_locate.go`
were traced and left unchanged. There is no page scanning, new cache, copied
editor state, DOM-query focus system or timer-based success assumption. Public
GridHandle, transport and persisted formats remain unchanged. The 20-pin and
32-origin limits, separate bounded query checkpoints, saved-view version handling,
record-pin base defaults, target-group reveal, generic unavailable feedback,
authorization/high-water admission and read-recovery budgets remain with their
existing owners. Presentation closure does not cancel admitted writes,
acknowledge conflicts, dismiss receipts or change uncertain-write recovery.

No persisted migration is needed. Rollback reverts this frontend handoff slice,
its tests and authored/generated routing together; stored data, adopted specs
and backend contracts need no change.

## Verification

Routing was rediscovered with `make help`, `make help-all`, owner `task-guide`
and `explain-test-owner`, then checked against `contracts/verification`, the
authored catalog/families and frontend source/import ownership. The relevant
owners were `web.workbook`, `module.workbook` and `package.grid_adapter`.
New browser row IDs were obtained through `make author-test-row-id`; generated
routing was not edited by hand.

The final focused commands were:

```sh
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbench_navigation_admission,web.workbook.regression.session_navigation,web.workbook.regression.explicit_built_in_grid_entry_92c0a52a3d,web.workbook.regression.useworkbookstartupcontroller_zsemantic_grid_focus_6f31bf7e92
make test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.regression.semantic_focus_requests
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.destination_handoff_1440_920aaf64d5,module.workbook.browser.destination_handoff_1024_963c4de116,module.workbook.browser.destination_handoff_768_8f85fa718d,module.workbook.browser.destination_handoff_hidden_a9564ecda1,module.workbook.browser.destination_handoff_supersession_62c776e360,module.workbook.browser.destination_handoff_viewer_64fc8d19e2,module.workbook.browser.built_in_grid_entry_editor_59f8176b4a,module.workbook.browser.built_in_grid_entry_supersession_5f9a05b19f,module.workbook.browser.built_in_grid_entry_viewer_c3bbf9714f
```

Retained run roots below are under `.cartulary/test-results/`.

| Command/check | Final result | Run root |
| --- | --- | --- |
| `make generate` | PASS | `20261007T112045Z-p48084` |
| `make format` | PASS, 2/2 units | `20261007T114047Z-p84460` |
| `make agent-finalize` | PASS, before final broader checks | `20261007T114211Z-p90420` |
| `make frontend-typecheck` | PASS, 2/2 units | `20261007T114323Z-p95295` |
| `make lint-biome` | PASS, 2/2 units | `20261007T114323Z-p95327` |
| `make frontend-import-boundary-check` | PASS, 2/2 units | `20261007T114323Z-p95307` |
| Web slice above | PASS, 5/5 units | `20261007T114323Z-p95177` |
| Grid Adapter slice above | PASS, 2/2 units, five focus tests | `20261007T114447Z-p67298` |
| Service-backed slice above | PASS, 12/12 units; all nine browser scenarios passed | `20261007T114323Z-p95183` |
| `make generated-artifact-policy-check` | PASS, 3/3 units | `20261007T115708Z-p81176` |
| `make lint-markdown` | PASS | `20261007T115956Z-p82677` |
| `git diff --check` | PASS | Working-tree diff |

The browser matrix covers 1440×900 adjacent presentation, 1024×768 right overlay
and 768×768 full overlay. It checks keyboard pin activation, same-destination
opening, actual committed-cell focus and selection, Arrow navigation, Work Escape,
coordination entry, Return and Return and inspect. Additional scenarios cover
hidden-field/root fallback, unsubmitted Notes and exact invalid Evidence draft
retention, failed and held locator reads superseded by newer interaction,
recordless Return, and viewer entry without enabling writes. Navigation causes
no source mutation requests. Existing creation-first, delayed-entry supersession
and viewer rows were run as compatibility checks, alongside the new scenarios.
Unit/component tests additionally cover unavailable/outside-query recovery,
delayed mounting and acknowledgment, unmount, authority/account replacement,
and Return bounds/coalescing. Group reveal and saved-view version/compatibility
branches were source-reviewed with their existing owners preserved; their full
matrices were not rerun.

Exploratory failures were corrected and are not final pass evidence:

- Red root-fallback and rendered-shell regressions:
  `20261007T110338Z-p32143`, `20261007T110357Z-p33027`.
- `make generate` rejected dynamic browser titles at
  `20261007T111935Z-p44410`; literal catalog-resolvable titles replaced them.
- Service-backed attempts at `20261007T112415Z-p92120` and
  `20261007T113124Z-p84283` exposed missing portal Escape handling and inspector
  focus remaining on the now-inert grid. Both production defects were repaired.
- Other intermediate browser failures involved virtualized offscreen targets,
  responsive menu timing and a Timeline free-text value incorrectly treated by
  the test as invalid. Tests now use the existing semantic scroll helper,
  settled-focus assertions and an actually invalid Evidence timestamp. Earlier
  type/lint errors were also repaired. The final checks above supersede them.

Retained-run maintenance was skipped because `RESULTS_DIR` was unset; no eligible
successful full warm check was supplied. Full backend, release, performance and
all-surface suites were not run: the changed production boundary is frontend
presentation/focus, with unchanged locator and mutation contracts. No golden
was changed or refreshed.

## Fresh rendered review

A new seeded `make ui-review` sealed the final source at
`.cartulary/test-results/20261007T114342Z-p12607`. At 1024×768, keyboard opening
of the Timeline pin from Notes ended on a visible committed cell with Work
closed; ArrowRight moved actual focus. Work → Task Requests closed Work and
focused the ordinary creation input. Work-open Return and inspect restored the
intended record, opened Record and focused its Close control. The three captures
were resolved through their manifests, digest/length verified and visually
inspected before stopping.

Review observations are advisory and separate from the product-test results.
The terminal receipt reports three images, zero axe violations, two incomplete
axe observations, zero console errors and ten failed network observations during
navigation. An aborted query was observed; these network counts are not a
product-pass metric. The final inspector capture had no incomplete axe findings.
This was not an assistive-technology or full accessibility-conformance audit.

The exact session was stopped with `make ui-review-stop`; its terminal receipt
reports `status=ok`, `state=closed`, `cleanup=complete`, and the foreground process
exited successfully. Only caller-owned request scratch was removed. Baseline
and final private captures have expired; no private image paths or raw
observations are retained here.

## Scoped acceptance and remaining limits

This assessment applies the digest's acceptance rows only to the changed seam.
It does not certify unchanged portions of a row or import upstream requirements.

| Rows | Result | Evidence or scope rationale |
| --- | --- | --- |
| A001–A003 | PASS | Owner map, clean baseline, current source/routing inspection and bounded selection rationale above |
| A004–A005 | PASS | Diff adds no theme, token, density or palette authority |
| A006 | N/A | Density geometry is unchanged; no density redesign or measurement claim |
| A007 | PASS | Creation-first and viewer compatibility rows; record/Return paths remain navigation mode |
| A008–A009 | PASS for changed handoff | Three-width overlay/adjacency matrix, visible destinations and keyboard usability; no claim to the full responsive boundary/zoom matrix |
| A010–A011 | PASS for changed handoff | One attachment, source selection before focus, explicit inspector acknowledgment, retained work and cancellation tests |
| A012–A013 | N/A for transaction/recovery algorithms | No write/retry/receipt implementation changed; navigation's no-write and presentation-only obligations are checked under A010–A011 |
| A014 | PASS for affected transitions | Exact invalid raw draft and unsubmitted creation survive detachment; Return stays in navigation mode; Escape tested |
| A015 | N/A | Conflict rendering and resolution owners are unchanged |
| A016–A017 | PASS for affected admission | Pending/failed/outside/unavailable origin retention; viewer and authority/account cancellation tests; no generic clear-all policy |
| A018 | N/A | Evidence lifecycle/preview states are unchanged; invalid draft is used only as retention evidence |
| A019 | PASS for affected keyboard/focus | Actual DOM focus, Enter, Arrow, Tab and Escape checks; fresh advisory axe review; no broader conformance claim |
| A020 | N/A | Component variant, typography, zoom and text-spacing design are unchanged |
| A021 | PASS for affected semantic focus | Existing virtualized semantic handles and scroll helper; requested/hidden/no-field fallback and delayed registration tests; no performance claim |
| A022 | PASS for scoped rendered evidence | Fresh seeded renderer inspection; canonical golden fixtures and visual-maintenance inputs unchanged |
| A023–A025 | PASS | Existing semantic UI identifiers, no executable Markdown dependency, authored catalog then generated routing and policy check |
| A026–A027 | PASS | Public/data compatibility, retired callbacks, no-migration decision, rollback, results and limits recorded here |

The executed browser scenarios use Chromium and the representative viewport
matrix. OS-native IME, assistive technology, all-surface/extension behavior and
the complete visual matrix were not audited. The original scoped readiness
conclusion is superseded by the following full-check and baseline comparison.

## Subsequent full check and root-cause analysis

### Remediation implementation status

The explicit-composition remediation is DONE. The implementation keeps
reusable layout independent of query navigation, requires a Network Flow frame,
retires the legacy overlay stack, reconciles selector and source ownership
policy, and adopts the production locator boundary additions. Public wire and
persisted formats are unchanged. No internal compatibility aliases were added.

| Phase | Status | Exit |
| --- | --- | --- |
| Contract cleanup | DONE | Harness focus-assertion policy and Entities/Projections locator boundaries describe the intended capabilities. |
| Layout composition | DONE | Provider-free frame, connected query surfaces and all eight affected Workbook rows pass; the Network Flow exploration row passes in its owner slice. |
| Dock migration and retirement | DONE | Required Network Flow frame, named fixture compositions, legacy deletion, exact source ownership and Fallow pass. |
| Policy and locator validation | DONE | Exact inventories, consistent-read behavior and extended selector alias/spread negatives pass. |
| Integrated validation and handoff | DONE | Focused browser/accessibility/visual checks, 1,014-unit full check and final locator test refinement pass; cleanup has zero failed or blocked steps. |

The historical results below remain diagnostic evidence. They do not certify
the remediated candidate; current run references are recorded separately below.

The user-requested `make check` finished in 1,158.87 seconds with 991/1,012 units
passing and 21 failing, exit code 2. Its retained root is
`.cartulary/test-results/20261007T124649Z-p98672`; `run-summary.json` and
`target-summaries/check.json` contain the aggregate result. Failing child targets
were `frontend-unit`, `frontend-fallow-static`, `backend-unit` and
`backend-integration`. `make explain-run` confirmed 1,131 cleanup steps with zero
failed or blocked steps. The full check is not a successful warm run eligible for
retained-run maintenance.

Analysis compared the current diff with an isolated clean checkout of
`69184ac184db412a2401a3e09c79f5f7554180c3`, reran the failing rows through public
Make targets, and tested one temporary fixture-composition change. There are
six failure groups, not 21 independent defects:

| Root cause | Failed units | Attribution and causal evidence |
| --- | --- | --- |
| Shared layout acquired a mandatory navigation-provider dependency | 9 | New handoff regression. `WorkbookSurfaceLayout.tsx:85` unconditionally calls `useWorkbookBrowsingRegistry`, which throws without its provider. Previously the layout used optional workbench registration. `WorkbookRecoveryFixture` mounts this layout for isolated Recovery and Network Flow consumers without a browsing provider. Eight workbook rows and the Network Flow exploration-focus row pass on clean baseline and fail at this exact new call in the current run. |
| Network Flow fixture has no auxiliary-dock host | 6 | Pre-existing baseline defect. `NetworkAnalysisWorkspace.test.tsx:140` selects `standalone={false}`, suppressing the fixture's `WorkbookSurfaceLayout`, and supplies no `workAreaFrame`. The workspace's fallback installs the old `WorkbookWorkAreaOverlayHost`. Recovery now renders through `WorkbookAuxiliaryDock`, which returns null while its separate host is absent. Thus Recovery, rename and mapping state can open without any corresponding DOM. All six fail on baseline; enabling the fixture's existing standalone host makes all six pass. |
| Frontend source-owner inventory was not reconciled | 1 | Pre-existing baseline drift. `tools/frontend_source_ownership.json` omits 16 live paths, all already tracked in the baseline, including Commands, navigation, auxiliary-dock and record-locator files. The manifest covers `apps/web/src`; the new e2e file is not the cause. The exact same ownership failure reproduces on baseline. |
| Selector policy conflates focus assertions with readiness checks | 1 | Mixed attribution. `selectorContractPolicy.test.ts:104` matches every named `getByRole("heading", ...)` query without considering the assertion. The flagged uses check actual Work-heading focus. Baseline already fails on `workbook.a11y.spec.ts:10740`; the handoff adds two matching assertions in the shell and new e2e tests. This is a policy/test-intent mismatch, not evidence that those focus assertions failed at runtime. |
| Old overlay export was left after the dock migration | 1 | Pre-existing baseline retirement gap. All production panel callers have moved to the auxiliary dock, leaving the `WorkbookWorkAreaOverlay` component unused. Its provider/host wiring remains, including the obsolete Network Flow fallback. Baseline Fallow reproduces the same single blocking web-surface finding. Removing the unused export alone would not repair the missing dock host. |
| Locator implementation and exact Go boundary inventories diverged | 3 | Pre-existing baseline contract-projection drift. Entities added `Store.LocateHostRows` and `Store.LocateIdentityRows`; Projections added `Store.LocateRows` and transactional Host/Identity reader methods. Export allowlists and the reflected `projectionports.QueryReader` method expectation still describe the earlier surface. All three failures reproduce on baseline. These tests fail on API shape, not returned locator rows or transaction execution. |

The arithmetic is nine newly failing units plus twelve already-failing baseline
units. The two new heading assertions add violations to a selector-policy unit
that was already failing. This attribution does not excuse the new provider
regression or establish full baseline health beyond the selected comparisons.

The main architectural issue is incomplete migration of composition contracts.
The full workbook shell supplies both the browsing provider and work-area frame,
so the original focused browser scenarios passed. Shared layouts and isolated
owner fixtures have other supported compositions; the new unconditional hook
changed their runtime requirement without migrating those consumers. Separately,
the earlier overlay-to-dock migration left one test composition connected to the
retired host. The source inventory, unused export and backend method inventories
show corresponding migration/retirement work left incomplete in the baseline.

The verification gap was selection by the visible destination journey without
fully following shared-layout consumers and independent ownership/selector
guards. Typechecking and import-boundary checks cannot establish React provider
availability. Passing navigation admission tests cannot establish fixture host
composition or reconcile authored inventories. No timing, resource exhaustion or
browser flakiness is needed to explain the observed failures.

### Controlled verification and evidence

Copied baseline and experiment evidence is retained under
`.cartulary/rca-destination-handoff/`. Each selected test slice used explicit
`ROWS`, preserved in its `run-manifest.json`. These are baseline/diagnostic runs,
not acceptance passes for the current implementation.

| Public target and selection | Result | Retained child directory |
| --- | --- | --- |
| `make test-slice OWNER=web.architecture`, selector and source-owner rows | Same two baseline failures | `20261007T132147Z-p33344` |
| `make test-slice OWNER=web.workbook`, eight affected Recovery/history/authoring rows | All eight rows pass; 9/9 units | `20261007T132147Z-p33346` |
| `make test-slice OWNER=web.networkflow`, six affected rows | Exploration-focus passes; five missing-panel rows fail | `20261007T132147Z-p33345` |
| `make test-slice OWNER=module.networkflow`, production-grid integration row | Same missing Recovery region on baseline | `20261007T132314Z-p37679` |
| `make test-slice OWNER=module.entities`, export and contribution-contract rows | Same two baseline failures | `20261007T132314Z-p37680` |
| `make test-slice OWNER=module.projections`, boundary/export row | Same baseline export drift | `20261007T132314Z-p37681` |
| `make frontend-fallow-static` | Same unused overlay export | `20261007T132430Z-p67389` |
| Network Flow five-row slice after temporary host-composition change | All five rows pass; 6/6 units | `20261007T132546Z-p68604` |
| Network Flow production-grid row after the same temporary change | Pass; 2/2 units | `20261007T132546Z-p68607` |

The temporary change only removed `standalone={false}` from the isolated
Network Flow test wrapper, thereby installing the existing shared frame/host.
It was reverted byte-for-byte after the experiment. The baseline checkout was
clean before cleanup and its retained reports were copied before archival.
An initial baseline slice could not start before dependencies were installed;
`make frontend-install` resolved that setup prerequisite
(`20261007T132034Z-p32158`). No production repair was applied during this analysis.

### Required follow-up

1. Restore the shared layout's supported composition contract. Integrate the
   existing handoff owner only where available, or inject its registration at
   the workbook composition boundary; do not create fallback navigation stores.
   Add the affected shared-consumer rows to verification of this seam.
2. Complete the Network Flow fixture/frame migration to the auxiliary dock,
   preserving production-like geometry, focus and retained-owner behavior.
3. Reconcile the selector rule with intentional focus assertions through the
   authored selector-policy owner, retaining the ban on copy-based readiness
   checks. Preserve the actual-focus regression assertions.
4. Assign the 16 existing source paths to their current owners and retire the
   unused overlay component together with any confirmed obsolete host wiring.
5. Review the locator-related Go surface additions against their adopted owners,
   then reconcile exact boundary projections or reduce unnecessary exports.
   Do not mechanically suppress the guards or weaken locator semantics.

At the end of the root-cause analysis, the earlier PASS assessment was blocked at A003 (current source ownership),
A010–A011 (shared-consumer composition), A023 (selector-policy reconciliation)
and A026 (supported-consumer compatibility). Those findings required repairs and
fresh verification; the original focused passes remain valid evidence of only
the journeys they executed. The full check was not rerun during that analysis.

## Explicit-composition remediation evidence

The remediation builds on the pre-existing uncommitted destination handoff. It
separates the reusable frame from `WorkbookQuerySurfaceLayout`, which binds the
required query registry and supplies browsing controls. Timeline, Entity,
Assessment and Generic surfaces use that adapter. Standalone Recovery uses the
frame without query providers. The frame publishes committed readiness through
a narrow mount-token binding; retirement cannot remove a newer registration.

Network Flow's `workAreaFrame` is required throughout the extension prop chain.
Its isolated tests inject the same extension frame as production. Recovery
fixtures now expose named framed and providers-only compositions; Timeline uses
the latter. The obsolete overlay component/provider/host and all executable
references were deleted. Responsive dock overlays and live contextual overlays
remain owner-required presentation. An auxiliary destination without its dock
provider fails explicitly; a host mounting during attachment remains valid.

All sixteen baseline source omissions and both new Workbook files are assigned
to `web.workbook`; the deleted shared overlay entry is removed. Source ownership
still covers exactly `apps/web/src` and is independent of test routing.

The Testing Harness owner now distinguishes direct heading focus assertions
from copy-based readiness. AST enforcement accepts actual-focus predicates and
rejects ambiguous uses. The stronger scan also exposed pre-existing async
heading waits and identity assertions hidden by the old regex's spelling/order
limitations. Network Flow title observations now use the existing stable heading
selector; incident identity tests use the existing identity selector and the
single semantic heading; Import Assistant waits on typed unit identity before
asserting its rendering. The original Work-focus assertions remain intact.

Entities and Projections boundary decisions now explicitly describe the five
retained locator/query-engine exports and the four-method directional
`QueryReader`. Exact export and interface inventories are reconciled. No
backend production algorithm, public wire contract, authorization rule, cursor
format or persisted schema changed. New production-catalog integration evidence
covers Host, Identity and Note snapshots across a committed concurrent update,
read-only repeatable-read options, borrowed read lifetime, begin/read/commit
failure and cancellation. No time-based sleeps or test-only production hooks
were added.

### Current focused runs

All roots below are under `.cartulary/test-results/`. Commands used public Make
entrypoints; exact row selections remain in their run manifests.

| Command / selection | Result | Run root |
| --- | --- | --- |
| `make generate`, initial new catalog | FAIL: new titles were not ASCII sorted; corrected in authored input | `20261007T161401Z-p18239` |
| `make generate`, corrected catalog | PASS | `20261007T161506Z-p21758` |
| `make generate`, added locator snapshot row | PASS | `20261007T161822Z-p34649` |
| `make format`, final authored-source pass | PASS, 2/2 units | `20261007T162401Z-p5672` |
| `make frontend-typecheck`, composition changes | PASS, 2/2 units | `20261007T161648Z-p30742` |
| `make test-slice OWNER=web.workbook`, affected Recovery/navigation, viewport and new composition rows | PASS, 13/13 units | `20261007T161647Z-p30635` |
| `make test-slice OWNER=web.architecture`, initial stronger selector scan | FAIL: identified older heading readiness/ambiguous uses; migrated without suppressions | `20261007T161647Z-p30634` |
| `make test-slice OWNER=web.architecture` | PASS, 13/13 units before added alias/spread negatives | `20261007T162105Z-p50147` |
| `make test-slice OWNER=web.networkflow` | PASS, 66/66 units | `20261007T162049Z-p43519` |
| `make test-slice OWNER=module.networkflow`, production-grid row | PASS, 2/2 units | `20261007T162250Z-p3607` |
| `make test-slice OWNER=module.entities`, exact export/contribution rows | PASS, 2/2 units | `20261007T162105Z-p50143` |
| `make test-slice OWNER=module.projections`, boundary row | PASS, 3/3 units | `20261007T162105Z-p50222` |
| `make service-backed-test-slice OWNER=module.workbook`, locator validation/public contract/consistent read | PASS, 3/3 units | `20261007T162050Z-p43797` |
| `make frontend-import-boundary-check` | PASS, 2/2 units | `20261007T162249Z-p3408` |
| `make frontend-fallow-static` | PASS, 2/2 units | `20261007T162300Z-p4583` |

The first format/typecheck attempts were rejected before child execution while
new catalog routing was not yet regenerated (`20261007T161330Z-p17662` and
`20261007T161341Z-p17986`); those empty roots carry no product result. Successful
generation resolved admission. Final integrated evidence follows below.

Rollback must revert the explicit frame/adapter migration with its consumers,
fixtures, source inventory, authored rows and generated routing. The locator
boundary amendments and tests are independently revertible; stored data needs
no migration. Do not restore only the obsolete host or only an optional prop.

### Workstream decisions and durable exits

| Gap / workstream | Areas and remediation | Rationale and long-term benefit | Migration, unresolved risk and observable exit |
| --- | --- | --- | --- |
| 1. Shared composition | Implementation, tests, source guides: independent frame plus connected query-surface adapter with an explicit mount-token binding. | The frame owns geometry and committed presentation; query navigation owns registration. New isolated consumers need no unrelated runtime. | Internal built-in consumers migrate together. No public change. Leaving the implicit provider dependency would keep isolated consumers vulnerable to render failures. Exit: raw frame renders unaided, connected surface rejects a missing provider, old cleanup cannot retire a replacement, and affected consumers pass. |
| 2. Dock host | Implementation, tests, guides: required extension frame and named framed/providers-only Recovery fixtures. | Production and tests share the same host composition; optional boolean configuration cannot silently remove Recovery presentation. | Internal callers must supply the frame. Leaving the fallback would allow accepted panel state with no DOM. Exit: Recovery, rename, mapping, responsive inertness, ordinary Escape and focus restoration pass with one panel. |
| 3. Selector intent | Testing Harness specification, AST policy and tests: recognize direct actual-focus predicates; migrate readiness to stable identity. | Preserves meaningful accessibility evidence while removing copy and formatting dependence from readiness. New cases extend one structural policy rather than regex exceptions. | Test authors must use the declared direct forms or stable selectors; aliases/spreads are conservatively checked. Leaving the gap either rejects valid focus evidence or encourages weakening that evidence. Exit: positive/negative policy fixtures, repository scan, unit and browser focus checks pass. |
| 4. Source ownership | Authored inventory and exact guard: add sixteen existing Workbook paths and two new files; remove the retired path. | Keeps one complete source-owner inventory distinct from verification routing. | No runtime impact or scope expansion. Leaving omissions hides ownership and prevents trustworthy boundary checks. Exit: exact coverage, unique assignment and import-boundary checks pass. |
| 5. Overlay retirement | Implementation, fixtures, source inventory and guides: delete obsolete component/provider/host after migrating the consumers. | One auxiliary presentation mechanism reduces lifecycle, focus and compatibility burden. Owner-required contextual overlays remain separate. | Internal removal only; revert the whole composition migration if rollback is needed. Leaving the old host invites split presentation ownership and dead wiring. Exit: no executable legacy references, Fallow clean, Recovery/browser checks pass. |
| 6. Locator boundaries | Adopted Entities/Projections boundaries, exact inventories and integration tests: retain the five justified exports and four-method query-reader port. | Source owners retain bounded location and one consistent snapshot; borrowed transaction methods do not acquire lifecycle authority. Future location work extends the existing directional ports. | No database, wire, cursor or authorization migration. Leaving drift either defeats exact boundary enforcement or invites removal of required consistent-read capabilities. Exit: exact guards, public admission/validation, all three source snapshots, cancellation and transaction-failure tests pass. |

The dependency order was contract review, layout separation, complete dock-host
cutover, legacy retirement with ownership reconciliation, then integrated
verification. Selector policy and backend boundary work were independent of
the UI host changes and converged at the validation gate. Generated routing was
refreshed only after authored rows were valid. The main phase risks were a
half-migrated host, stale registration cleanup, copy-dependent waits hidden by
the old regex, and mechanically expanding API allowlists. The composition tests,
stronger repository scan and production-catalog snapshot evidence address those
risks before handoff.

### Integrated remediation validation

| Command / selection | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make test-slice OWNER=web.workbook`, shell surfaces and incident identity | PASS, 3/3 units | `20261007T162418Z-p11034` |
| `make test-slice OWNER=module.imports`, operator assistant | PASS, 2/2 units | `20261007T162438Z-p43639` |
| `make frontend-typecheck`, final source candidate | PASS, 2/2 units | `20261007T162429Z-p29110` |
| `make test-slice OWNER=web.architecture`, extended focus-policy fixtures | PASS, 2/2 units | `20261007T162834Z-p5114` |
| `make service-backed-test-slice OWNER=module.workbook`, destination handoff, grid entry and workbench accessibility | PASS, 14/14 units | `20261007T162418Z-p11029` |
| `make service-backed-test-slice OWNER=module.networkflow`, keyboard Recovery, claimed accessibility and visual states | PASS, 13/13 units | `20261007T162418Z-p11028` |
| `make agent-finalize` | PASS, 1/1 unit; generated structure unchanged | `20261007T162903Z-p6595` |
| `make check` | PASS, 1,014/1,014 units; 1,038.38 seconds | `20261007T162949Z-p11436` |
| `make format`, final test-only refinement | PASS, 2/2 units | `20261007T164729Z-p14082` |
| `make service-backed-test-slice OWNER=module.workbook`, final consistent-read regression | PASS, 3/3 units | `20261007T164755Z-p19628` |
| `make lint-go`, final test-only refinement | PASS, format/vet/staticcheck | `20261007T164757Z-p20061`, `20261007T164804Z-p29006`, `20261007T164806Z-p30353` |
| `make lint-markdown`, completed handoff and owner/source guides | PASS | `20261007T164941Z-p47017` |

`agent-finalize` passed JSON shape, catalog/tier coverage and generated drift
checks. Retained-run evidence/performance maintenance was skipped because
`RESULTS_DIR` was unset; the earlier failed full check was not presented as
successful warm evidence.

The fresh full check has zero failed, skipped or cancelled units. All 21 units
that failed in the historical run pass in this run. `make explain-run` confirms
1,029 cleanup steps with zero failed or blocked steps; the exact aggregate and
cleanup artifacts are `run-summary.json`, `target-summaries/check.json` and
`cleanup-results.json` under that run root.

After that aggregate pass, only the new locator regression test was refined:
it no longer requires a minimum SQL query count, and injected read failure is
independent of read ordinal. This permits later query consolidation while
retaining the snapshot, failure-propagation and transaction-lifetime assertions.
No production source, specification, catalog or generated routing changed after
the aggregate run. Final focused execution and Go lint cover this test-only
refinement; the full suite is not redundantly rerun for it. The focused run's
six cleanup steps also pass without failed or blocked cleanup. Final
`git diff --check` passes.

Artifact-mode UI review imported the exact mapping Recovery and narrow query
captures from the successful Network Flow visual run. Both imports contained
only expected images (`no_actual`, `no_dom`, `no_axe`, `no_trace`); byte lengths
and SHA-256 digests were checked before inspection. The reviewed expected states
show one right-side Recovery region and compact query controls above the grid.
Current rendering correspondence is established by the passing canonical visual
assertions, not by treating those expected images as new captures. No golden was
changed. No interactive HTML report was inspected. The private session stopped
successfully, its terminal receipt records `cleanup: complete`, and caller-owned
request files were removed. The structural review receipt is
`20261007T162844Z-p5926/ui-review/terminal.json`; private images are intentionally
not retained as handoff links.

### Acceptance after remediation

This assessment supersedes the original scoped assessment and its subsequent
RCA block; the integrated phase is DONE. The acceptance rows
remain advisory; adopted owners and the executable checks retain their separate
roles.

| Rows | Result | Current evidence / boundary |
| --- | --- | --- |
| A001–A003 | PASS | Owner amendments and workstream decision table; existing dirty work preserved; exact source/import ownership checks pass. |
| A004–A005 | PASS | No theme, token, density, icon or grid-vendor authority added; existing graphite rendering retained. |
| A006 | N/A | Density metrics and selection are unchanged. |
| A007 | PASS for affected composition | Built-in creation-first and viewer entry browser rows, Import Assistant and retained Recovery tests pass. |
| A008–A009 | PASS for affected composition | Three viewport handoff rows, viewport-layout fixture, adjacent/compact dock inertness and Escape tests pass. Full zoom/below-minimum design audit is outside this repair. |
| A010–A011 | PASS | Provider-free isolated frame, required connected composition, one dock attachment, stale mount retirement, retained authoring and browser destination focus pass. |
| A012–A013 | PASS for presentation obligations | Affected Recovery/history and Network Flow owner suites pass. Request construction, retries and receipts remain with their existing owners; no write/replay algorithm changed. |
| A014 | PASS for affected transitions | Retained drafts, navigation closure, ordinary Escape and invoker focus restoration pass. |
| A015 | N/A | No conflict rendering or resolution algorithm changed. |
| A016–A017 | PASS for affected admission | Existing pending/failure/unavailable, supersession, viewer and authority-loss rows pass; explicit composition creates no new authorization or query state owner. |
| A018 | N/A | Evidence lifecycle and preview combinations are unchanged. |
| A019 | PASS for affected keyboard/focus | Actual focus assertions remain enforced; Workbook and Network Flow accessibility slices pass. Artifact-only review supplies no additional axe or DOM evidence. |
| A020 | N/A | No component variant, density, typography, zoom or text-spacing redesign. Relevant long-content visual comparison remains under A022. |
| A021 | PASS for affected continuity | Existing semantic grid adapter and delayed-registration/handoff checks pass; no performance claim or new row model. |
| A022 | PASS for scoped visual evidence | Canonical Network Flow visual row passes with unchanged goldens; exact expected artifacts inspected with the limitations recorded above. |
| A023 | PASS | Structural focus-policy positives/negatives and full repository scan pass; migrated readiness uses stable selectors or typed owner state. |
| A024–A025 | PASS | No executable Markdown dependency added; authored rows precede generated routing; finalizer's shape, catalog and generated-drift checks pass. |
| A026 | PASS | Internal migration is complete; obsolete host removed; justified locator ports retained. No public schema, persisted data or authorization migration. |
| A027 | PASS | All workstream exits, exact aggregate/focused results, clean resource teardown, compatibility, rollback and skipped-scope reasons are recorded. No implementation blocker remains. |

Release publication, a new performance baseline, the complete visual/zoom matrix,
OS-native IME and assistive-technology certification were not selected: this is a
composition and contract-reconciliation repair, not a release or general UI audit.
No deployment, commit, dependency installation change or data migration is part
of this handoff.
