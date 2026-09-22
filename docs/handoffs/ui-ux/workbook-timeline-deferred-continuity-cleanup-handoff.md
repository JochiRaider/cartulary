# Timeline deferred continuity cleanup

## State and scope

Branch `main`; baseline HEAD `83c1f7b9f10cf50d7064d10f3bcdf50ad484eb45`.
The working tree was clean at planning and implementation entry. Preserve later
user changes. This slice changes deferred Timeline presentation after acceptance
and refresh; operation semantics, persistence and public HTTP contracts remain
with their owners. No commit, push or deployment is authorized.

The approved policy is document-wide cancellation for newer pointer, keyboard,
wheel, native input and external focus, with narrow recognition of restoration's
own effects and passive editor detachment. Timeline owns appropriateness;
Grid Adapter owns semantic readiness/reveal/focus; mutation owners retain writes,
exact replay, acknowledgement and read recovery.

## Authority and navigation

Read AGENTS.md and the Cartulary UI/UX skill, digest START_HERE, local prompt,
maps, rules, acceptance and query recipes. Revalidated current source guides,
manifests, generated-artifact policy and verification routing. The digest is
September 13 navigation at `59fa79e04035a4f29251fcb2337376f29c40f1f4`, not current
behavior authority. The completed row-menu handoff is a regression reference.

| Behavior | Governing owner | Source / verification |
| --- | --- | --- |
| Accepted versions and follow-up continuity | Core 03 REQ-03-035/283, §§3–4 | Timeline controller/coordinator/query owners; web.workbook and module.timeline |
| Editor departure, retained authoring, selection and Find | Core 03 REQ-03-217/218/219/298/300, §§13.4–13.5 | Timeline editing and Grid Adapter; web.workbook, module.timeline, package.grid_adapter |
| Loaded-window continuity and no editor resurrection | Core 03 §14.9 | Timeline query/presentation and Grid Adapter |
| Focus fallback and native keyboard ownership | design §§8.4–8.6, §10, §14 | Timeline continuity; existing menu/Inspector/Find owners |
| Authority invalidation independent of retained work | Core 03 REQ-03-299/100; Core 04 §§1–2 | Existing collaboration and mutation authority; Timeline presentation |

No adopted-owner contradiction was found. Research nlspec-spec.md is explanatory
material. Product checks do not consume Markdown.

Offline advisory query (Python bytecode disabled):
`keyboard focus async restoration --domain ux -n 3 --json` returned three results,
without fallback. ADOPT R002 for owner-required visible focus; ADAPT R018 to
preserve legitimate horizontal grid scrolling; REJECT R033/R034 as new behavior
authority or incidental selectors. No upstream design-system operations ran.

## Sequential exits

1. **Characterize — exited.** Current source had separate generation-only input
   handling and explicit abort handling. Row-target external focus was unchecked.
   New focused tests establish a pending semantic request before interruption,
   then mount the old target: both wheel/native-input and external-focus scenarios
   fail because the old cell takes focus. Red run:
   `.cartulary/test-results/20260922T110305Z-p7010`, `test-slice`, 1/2 graph units.
   Initial attempt `20260922T110235Z-p6077` failed in fixture setup because the
   repository's DOM geometry shim is read-only; corrected by defining per-element
   test geometry. It is not defect evidence. Acceptance-time fresh-draft direct
   reveal/focus is a further structural bypass to close.
2. **Implement — exited.** One Timeline cancellation lifetime now owns adapter
   requests, geometry, frames, retries and fallback. Accepted fresh-draft/input
   continuity delegates to it; active editor promotion and operation settlement
   retain their owners. Composition supplies accepted scope and live authority.
   Authored source/test routing preceded generator run
   `20260922T111939Z-p18263` (PASS). Focused controller, continuity model,
   keyboard and menu rows pass `20260922T112040Z-p21796` (7/7 units);
   coordinator passes `20260922T111244Z-p13688` (2/2); typecheck passes
   `20260922T112409Z-p24748` (2/2). Intermediate typecheck failures were
   nullable test handles, missing test-ID arguments and a browser type import
   pulling CSS into Node; corrected in test support. Production verification
   starts after this recorded exit.
3. **Production verification — exited.** Production pending-request,
   compatibility, Find, menu/Inspector accessibility, collection promotion and
   read-only recovery checks pass. The final pending-request rerun after geometry
   focus reuse passes 20260922T113937Z-p95416 (11/11). Existing AC-043 measurement
   passes 20260922T114149Z-p28932 (14/14), with the current budget unchanged.
4. **Finalize and acceptance — exited.** agent-finalize passes before the final
   typecheck, import-boundary, JSON-shape, generated-policy, lint and composition
   checks. All applicable acceptance rows are PASS; scoped N/A rows give reasons.
   RESULTS_DIR is unset because these narrow runs are not eligible full warm-check
   evidence; retained-run maintenance is explicitly skipped. Markdown maintenance
   also passes with this handoff included in its authored scope.

Planning baseline checks passed: continuity model/menu `20260922T105447Z-p439`
(5/5 units), coordinator `20260922T105554Z-p1695` (2/2), adapter semantic focus
`20260922T105554Z-p1702` (2/2). They do not cover the new race.

## Selection rubric and cutover

| Decision | Assessment |
| --- | --- |
| Observed weakness | Confirmed controller race: after a request is already pending, wheel/native input or an external focus destination does not retire it; mounting its old target takes focus. Split scheduling and acceptance-time focus are structural bypasses. No prior production performance defect is claimed. |
| Remediation/change areas | Authorized behavior correction: newer destination wins across every asynchronous boundary. Structural cutover: one controller lifetime, a scope observation port, and one accepted-continuity command. Changes are implementation, focused/browser tests, authored routing, generated routing projections and local guides; adopted specifications are unchanged. |
| Responsible owner/boundary | Core/design clauses above own behavior. Timeline decides appropriateness; Grid Adapter decides semantic readiness/reveal/focus; mutation owners decide admission/settlement/replay. Frontend source owner is web.workbook. Independent tests route to web.workbook, module.timeline, package.grid_adapter and relevant module.workbook consumers. |
| Rationale/benefit | Synchronous retirement makes every continuation share the same eligibility decision. Live authority checks fence effects even before React cleanup. This avoids callers independently deciding whether a late acceptance may focus. No timing or speed improvement is inferred. |
| Future extension | A later Timeline follow-up can advance/settle its existing named requirement under the same token. Another surface can use the existing Adapter AbortSignal capability with its own appropriateness owner; no cross-surface navigation framework is added. |
| Capability value | Keep semantic identities, row-version floors and named follow-ups for authoritative projection; native input registration for detached-editor recovery; Adapter readiness notifications for virtualization; owned geometry for visible focus and both scroll axes. These are required continuity capabilities, not retained solely for historical compatibility. |
| Retirement | Remove generation-only interruption, input-only external-focus cancellation, unconditional window.focus, separate repeated frame/reveal/focus scheduling, acceptance-time direct reveal/input focus, and unused mutation-editor focusInput/reveal plus exposed snapshot/input helpers. Migrate callers together. |
| Compatibility/migration | Existing requestFocus(signal, preserveSelection) stays unchanged, including Adapter pointer/key cancellation. No data, protocol, dependency, persistence or deployment migration. Only private Timeline ports change; coordinator and runtime-binding consumers are updated. No alias/dual implementation remains. |
| Risk of leaving gap | Demonstrated stale focus takeover in pending-target controller tests; structural risk of later geometry/retry/fallback moving focus or scroll after a newer destination. No speculative security breach or measured performance cost is asserted. |
| Validation/delivery | Characterize before correction; focused implementation exit before production checks; handoff/finalization after browser checks. Deterministic gates remain in E2E support. Rollback the slice's source/tests/authored routing/generated projections together, preserving later user changes. |

### Retained readiness and effect policy

Each restoration captures its token, accepted incident/surface/query scope,
interaction generation and root abort controller. Each projection pass has a
linked abort controller. Replacement, clearing, document-wide interruption,
scope/authority invalidation and unmount synchronously retire pending Adapter
requests and clear timers/frames. Late advances and named settlements cannot
recreate a retired token.

Pointerdown, keydown, wheel and native input are observed in document capture
without consuming events. Focusin yields to external controls, including nested
controls within a requested cell. Only the exact registered input or current
requested semantic grid target is exempt. Passive editor removal does not itself
mean the user chose another destination. No blanket asynchronous focus exemption
or global window focus remains.

Every effect checks current ownership: Adapter invocation and its awaited result,
direct input focus, both scroll assignments, geometry, stabilization frame and
fallback. A successful focus result is reused during transient geometry retries
while its element remains connected. Actual detachment can recover. Adapter
registration notifications own pending virtualized targets; terminal cancellation
is final. Unavailable membership is not polled; explicit pending named projection
requirements await their owner callbacks.

The existing bounded 50 ms / 60-retry fallback remains for registered input mounts
and transient geometry, which lack a dedicated readiness notification. It stops
immediately on cancellation. Exhausted editor readiness uses eligible same-row
navigation, then grid root and eligible shell destinations; it never activates an
editor. Existing §14.9 loaded-window/query fallback remains with its owner.

Fresh-draft continuation delegates the accepted semantic result to the controller
and drops pre-create scroll. Active draft-to-record editor transfer stays with the
coordinator to preserve newer text, selection and composition. Accepted writes,
exact replay, acknowledgement, retained drafts and required refresh reads do not
share the presentation abort signal. Pending/failed query replacements retain
the accepted scope; an accepted replacement or changed authority retires focus.

### Repository and file review

Revalidated authored source/import manifests, verification registry/catalog and
family manifests, generated policy/topology, and local Timeline/Adapter guides.
The current stack is React 19.2.5, TypeScript 6.0.2, Vite 8.0.8, Vitest 4.1.4,
Playwright 1.59.1, Node 24.15.0 and pnpm 10.33.0. The sole direct react-data-grid
integration remains packages/grid-adapter. The dated digest omits some current
source/test paths and the generated view-contracts root; current manifests govern.

Changed production files under apps/web/src/workbook/timeline:

- hooks/useTimelineViewportContinuityController.ts: lifetime, policy, readiness,
  guarded effects and fallback.
- composition/useTimelineWorkbookComposition.ts and useTimelineGridEnvironment.ts:
  accepted scope/live authority observation and controller wiring.
- composition/useTimelineMutationComposition.ts and
  mutations/useTimelineRowMutationCoordinator.ts: accepted-continuity delegation.
- adapters/createTimelineRowMutationEditorAdapter.ts and
  models/timelineControllerPorts.ts: retire the superseded mutation-editor ports.

Focused tests are useTimelineViewportContinuityController.test.tsx (13 cases),
useTimelineCompositionLifecycle.test.tsx, useTimelineMutationRuntimeBindings.test.tsx,
and mutations/useTimelineRowMutationCoordinator.test.tsx. The browser test is
apps/web/e2e/timeline-deferred-continuity.spec.ts with choreography isolated in
apps/web/e2e/support/timeline/deferredContinuity.ts. Local root/adapters/composition/
hooks/mutations README guides describe the resulting ownership.

Authored tools/frontend_source_ownership.json and the web.workbook/module.timeline
family manifests route the new cases and the previously unrouted existing
fresh-draft coordinator assertion. make generate produces only the changed
browser_e2e_batch_manifest.json and execution_topology_render_index.json. The authored .markdownlint-cli2.jsonc list adds this handoff for documentation
maintenance. No product/check input reads Markdown; no dependency/lockfile or generated product
contract was hand-edited.

## Verification record

All commands run at the repository root with
`PATH="$PWD/tmp/node-runtime/bin:$PATH"`. Artifacts below are beneath
`.cartulary/test-results/`; each run has run-summary.json and owner row results,
or the target's tool-run-summary.json. Browser groups retain Playwright reports
and traces for failures; the new passing race also attaches its observations.

Guidance: `make task-guide ROLE=module-author OWNER=web.workbook`,
`OWNER=module.timeline`, `OWNER=package.grid_adapter`, and `OWNER=module.workbook`.
Authored routing was reviewed independently from behavioral authority.

| Command/selection | Result and artifact run |
| --- | --- |
| make test-slice OWNER=web.workbook; new controller + three continuity-model rows + menu + keyboard | PASS 7/7 graph units, 20260922T112040Z-p21796 |
| make test-slice OWNER=web.workbook; controller + runtime bindings after unused-port removal | PASS 3/3, 20260922T113029Z-p55601 |
| make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_deferred_continuity; geometry reuse added | PASS 2/2, 20260922T113835Z-p90881 |
| make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_row_mutation_coordinator_1a7e2c9b44 | PASS 2/2, 20260922T113029Z-p55624 |
| make test-slice OWNER=package.grid_adapter; semantic_focus_requests, overlay_focus_preserves_selection, range_keyboard_entry | PASS 4/4, 20260922T112717Z-p88527 |
| make service-backed-test-slice OWNER=module.timeline; 17 compatibility rows listed below | PASS 15/15 graph units, 20260922T112954Z-p25498 |
| make service-backed-test-slice OWNER=module.workbook; find_edit_departure, find_source_editors, find_virtualized | PASS 11/11, 20260922T113311Z-p93646 |
| make service-backed-test-slice OWNER=module.workbook; accessibility.timeline_row_action_menu, accessibility.inspector_edit_recovery | PASS 13/13, 20260922T113422Z-p56995 |
| make service-backed-test-slice OWNER=module.timeline; browser.collection_input_promotion | Row PASS in 20260922T113421Z-p56770; companion recovery worker killed, see failures |
| make service-backed-test-slice OWNER=module.timeline; browser.deferred_continuity, browser.related_evidence_projection_recovery | PASS 13/13, 20260922T113725Z-p57898; pending race includes current row-version, range, bulk, native text/selection and one accepted browser write |
| make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.deferred_continuity; final geometry reuse | PASS 11/11, 20260922T113937Z-p95416 |
| make generate after authored routing, repeated after geometry case | PASS 20260922T111939Z-p18263 and 20260922T113905Z-p92068 |
| make frontend-typecheck during implementation | PASS 2/2, 20260922T112409Z-p24748; final check PASS 20260922T114648Z-p70310 below |

Exact 17-row compatibility command:

```sh
make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.row_action_menu,module.timeline.browser.spreadsheet_keyboard,module.timeline.browser.spreadsheet_creation,module.timeline.browser.spreadsheet_uncertain_create,module.timeline.browser.spreadsheet_refresh,module.timeline.browser.spreadsheet_rejection,module.timeline.browser.spreadsheet_virtualization,module.timeline.browser.spreadsheet_batch_replay,module.timeline.browser.spreadsheet_bulk,module.timeline.browser.collection_input_authoring,module.timeline.browser.range_entry_cycles,module.timeline.browser.range_entry_lifecycle,module.timeline.browser.range_entry_native,module.timeline.browser.range_entry_settlement,module.timeline.browser.range_departure,module.timeline.browser.range_cancellation,module.timeline.browser.range_rejection_query
```

Routing detail: the retained measurement row ID contains `summary_enter_focus`,
but its current selector is “measures paint-qualified Timeline summary pointer
focus within AC-043” with predicate `perf.timeline_summary_focus_edit.v1`.
The current selector/budget is preserved; Enter/Tab behavior is covered by the
separate spreadsheet keyboard/departure regressions.

The browser probe observes the public semantic request and its real abort signal.
Its test-only connected-readiness gate delays the exact semantic cell without
changing production runtime. Every case proves a request is pending before new
intent; release is followed by a real accepted row-version update. Postconditions
observe active-element identity, both scroll axes, completed range, bulk checkbox,
exact native draft/caret and PATCH count without refocusing or scrolling.
Uninterrupted restoration must focus the original semantic cell. This gate is
controlled readiness evidence, not a reproduction frequency or performance claim.

### Failures and corrections

- Red characterization and fixture setup failure are distinguished in the first
  workstream exit. The implementation's first focused pass was
  20260922T110928Z-p10596. An expanded fixture assumed exactly one fake timer;
  20260922T111219Z-p12805 failed that assumption, corrected to assert the pending
  retry and cancellation behavior; 20260922T111343Z-p14542 passed.
- Intermediate typechecks 20260922T110909Z-p9912 and
  20260922T112040Z-p21854 failed nullable handling and new E2E type/selector setup;
  corrected before the implementation exit.
- Browser gate attempts 20260922T112442Z-p25340 and
  20260922T112639Z-p58593 detached React-owned cells and caused renderer teardown.
  Replaced physical DOM removal with the isolated readiness observation gate.
- 20260922T112824Z-p91885 reached pending focus and passed wheel interruption but
  timed out waiting for an accepted, correctly deactivated grid input. The native
  input case now uses the persistent Inspector input. 20260922T113225Z-p61399 used
  a nonexistent overflow control for one tag; corrected to its semantic Inspect
  action. These are test-choreography failures, not product defect evidence.
- 20260922T113421Z-p56770's related-evidence browser worker exited via SIGKILL
  without a row assertion result; target summary then lacked that evidence.
  Cause was not established. The exact recovery row passed on the isolated
  rerun in 20260922T113725Z-p57898. No test or requirement was weakened.

- Final static attempts 20260922T114552Z-p66930 (typecheck) and
  20260922T114552Z-p67009 (Biome) caught a nullable captured test-gate variable and
  four unchecked fixture indexes. Both test-only issues were corrected; final
  typecheck and Biome pass in the runs below.

## Acceptance assessment

Assessment is scoped to this cancellation slice; unchanged themes, rendering
variants and unrelated workflows are not claimed freshly certified.

| Row | Status | Evidence or scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Exact Core/design map above; advisory dispositions separate source placement and verification. |
| A002 Scope | PASS | Complete selection rubric and retired paths above; one presentation decision, no second navigation/draft/selection owner. |
| A003 Repository state | PASS | Initial clean main/HEAD recorded; current source/import/generated/routing/stack and sole vendor integration inspected; localization drift noted. |
| A004 Tokens | PASS | Diff adds no design literals or token/theme/density registry. |
| A005 Theme | N/A | No theme/rendering behavior changes. |
| A006 Density | N/A | No density, typography or editor dimensions changed; geometry uses existing rectangles. |
| A007 Creation | PASS | Fresh-draft and promotion controller/coordinator coverage; production creation, uncertain creation and collection-promotion rows pass; no capability/payload changes. |
| A008 Responsive | N/A | No shell thresholds, clamps or responsive control selection changed. Existing focused accessibility cases exercise constrained layouts. |
| A009 Overflow | PASS | Owned-axis checks and browser wheel postconditions; menu/Inspector accessibility remains reachable within existing shell ownership. |
| A010 Inspector | PASS | Persistent Inspector native input survives cancelled restoration; explicit Find/Inspector borrowing and Inspector recovery accessibility pass; dispatch/review owners unchanged. |
| A011 Continuity | PASS | Pending-target controller/browser matrix, accepted scope/authority, detached editor, named follow-ups, range/bulk, fresh draft and both-axis checks. |
| A012 Transactions | PASS | Browser pending case observes one accepted write; uncertain creation and batch replay compatibility pass. No transaction bytes/ID logic changes. |
| A013 Acknowledgement/recovery | PASS | Exact replay regressions and accepted-write/failed-projection read-only recovery pass; operation ownership remains separate from abort signals. |
| A014 Editing | PASS | Enter/Tab acceptance/rejection, collection authoring, native composition, range departure/cancellation and exact drafts pass. |
| A015 Conflict | N/A | Conflict presentation/admission is unchanged; rejection and exact replay consumers verified. |
| A016 Query/interaction states | PASS | Accepted query scope retained until replacement; scope/authority focused cases plus refresh/rejection/Find browser cases; data-state producers unchanged. |
| A017 Refresh/authorization scope | PASS | Live authority and accepted-scope fences, unmount and late-remount tests; refresh/virtualization pass. Existing account/session retention and clearing owners untouched. |
| A018 Evidence | N/A | No Evidence access, lifecycle, overlays or previews changed; shared follow-up recovery consumer verified separately. |
| A019 Accessibility | PASS | Document events are observed without consumption; narrow owned focus and nested-control cancellation tests; menu/Inspector accessibility and keyboard/Find browser checks pass. |
| A020 Components | N/A | No component visual variants, typography, density or compound-state precedence changed. |
| A021 Virtualization | PASS | Production semantic pending target, virtualization and final geometry rerun pass; retained AC-043 measurement row passes its unchanged budget in 20260922T114149Z-p28932. |
| A022 Visual fixtures | N/A | No renderer styles/layout or visual golden change; semantic focus behavior uses functional/accessibility evidence. |
| A023 Selectors | PASS | Tests use view/record/field IDs, semantic controls and public focus capability; no component-name or vendor-coordinate gates. |
| A024 Test authority | PASS | Tests and runtime do not consume Markdown; documentation lint is maintenance only. |
| A025 Generated artifacts | PASS | Authored source/family routing precedes Make generation; generated diff is routing only. agent-finalize also passes catalog/shape/generate-drift with zero further generated edits. |
| A026 Compatibility | PASS | Private port cutover removes superseded paths; production shared API unchanged; no data migration, new endpoint, persistence or dependency. |
| A027 Handoff | PASS | Sequential exits, complete rubric/acceptance, results/failures/artifacts and rollback recorded; all selected product checks pass. Documentation maintenance result recorded below. |

## Limitations and rollback

The bounded input/geometry fallback remains because those mounts/layouts have no
sufficient existing readiness notification. Browser gates deliberately control
readiness; coverage does not measure natural race frequency. No speed improvement,
new conformance profile or Core 05 publication claim is made. No shared Adapter
implementation changed; its relevant consumers and existing cancellation are
covered. No visual goldens or unrelated service/backend suites are rerun because
no corresponding behavior changed.

Rollback reverts this slice's Timeline source/private ports, tests, authored source
and family routing, generated routing projections and guides/handoff together.
There is no data migration. Preserve subsequent user edits; do not reset the tree.
No commit, push or deployment was performed.

## Finalization and final checks

`make agent-finalize` passes in 20260922T114517Z-p62842 before broader final
verification. Its `unit-artifacts/finalize-summary.json` records PASS for shape,
catalog, tier coverage and generate-drift; zero generated files were changed.
Retained canonical-run/scheduler/performance maintenance was skipped because
RESULTS_DIR was unset. Narrow successful runs are not full warm-check evidence.

| Final command | Result / run |
| --- | --- |
| make frontend-typecheck | PASS 2/2, 20260922T114648Z-p70310 |
| make frontend-import-boundary-check | PASS 2/2, 20260922T114552Z-p66967 |
| make json-shape-check | PASS 3/3, 20260922T114552Z-p66776 |
| make generated-artifact-policy-check | PASS 3/3, 20260922T114552Z-p66770 |
| make lint-biome | PASS 2/2, 20260922T114648Z-p70344 |
| make test-slice OWNER=web.workbook; exact grid-environment, Inspector lifecycle and public-root/presentation architecture rows | PASS 5/5, 20260922T114603Z-p68179 |
| git diff --check | PASS; branch/HEAD unchanged |
| make lint-markdown | PASS, 20260922T114928Z-p73506, adhoc/lint-markdown/tool-run-summary.json; includes this handoff through its authored lint glob |

Exact retained measurement command:

```sh
make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95
```

PASS 14/14 graph units in 20260922T114149Z-p28932. The retained
`browser-e2e-measurement/browser-groups/measurement-measurement-timeline-grid-d03cf54e95/frontend-measurement-summary.v3.json`
reports p95 30.1 ms against the unchanged 100 ms threshold for
`perf.timeline_summary_focus_edit.v1`. This is a budget check for the current
fixture, not an improvement comparison or publication claim.

Final state: only this slice's source, tests, guides, handoff and authored/generated
routing are dirty on the original main/HEAD. No unrelated user edits were present
or removed. No unresolved product failure remains. The next action is review;
no commit, push or deployment was performed.
