# UI review preparation remediation handoff

This change makes seeded review prepare the selected checkout after explicit tool
setup and preserves structural preparation diagnostics across child processes.
Both seeded profiles now start directly after setup. Rendered keyboard/focus
inspection and repository qualification are complete: the final `make check`
passed all 982 units. This document separates harness evidence from product
acceptance and advisory findings.

## Authority and historical evidence

The Testing Harness NLSpec owns lifecycle, diagnostic, privacy and migration
requirements. Domain vocabulary and the NLSpec research document remain unchanged.
Executable checks consume authored machine projections, never Markdown.

Work began at `b179d87fa` with no registered active review sessions. The original
receipt at `.cartulary/test-results/20260930T002033Z-p28086/ui-review/terminal.json`
remains historical v1 evidence: startup failure, exit 3, complete cleanup. The
successful smoke and later retry support a prerequisite hypothesis but do not
identify the lost exception. No historical receipt was rewritten.
The cited smoke receipt is
`.cartulary/test-results/design-review-1790729068254-631e9129/review-session.json`;
the later successful retry is
`.cartulary/test-results/20260930T004551Z-p5499/ui-review/terminal.json` (exit 0,
cleanup complete). The implementation baseline's full commit is
`b179d87fa8260db90d31b64fcecf0aee84c7a628`.

## Changes and ownership

| Gap | Owner and change | Qualification route |
| --- | --- | --- |
| G1 | Testing Harness lifecycle now separates installed readiness, source preparation, service acquisition/readiness, seeding and browser startup. | UI review contract and lifecycle rows, AC-130/131. |
| G2 | Readiness owner publishes installation proof only from explicit installation; doctor and review share read-only predicates. Cache eviction does not invalidate installation. | Isolated prerequisite fixtures in the UI review lifecycle row; doctor; cold seeded workflows. |
| G3 | Shared preparation uses `ensure` or `installed_only`, builds all five source producers in dependency order, and validates the source snapshot before ready. Go execution is offline and image acquisition is blocked at the Docker client boundary. | Both seeded profile rows; installed-only image row; producer and frontend artifact checks. |
| G4 | Invocation-bound private failures survive Make and preparation IPC. V2 public records retain closed context and recovery IDs; cleanup remains secondary. | Real Make children, malformed/conflicting/cross-invocation envelopes, IPC death, lifecycle and terminal publication fixtures. |
| G5 | Test-owned installation and reuse-cache fixtures cover cold state without deleting shared tools or caches. | `harness.browser` and `harness.command_surface` routes. |
| G6 | Operating guide and skill references explain direct startup, v2 cutover and exact-session recovery. Investigation output lists retained UI receipts and mode-dependent services. | Registry parity; `explain-target`; documentation lint. |
| G7 | Workbook rendered keyboard/focus inspection completed for editor/viewer roles on both profiles. | Selector, extension entry, accessibility rows and private seeded inspection below. |

Authored changes are in readiness, shared review preparation, the UI lifecycle,
private command diagnostics, service fixture adapters, JSON schemas, task metadata,
verification routing, the owner specification and operating guidance. Generated
task/topology projections were regenerated through `make generate`. Lockfiles and
generated outputs were not hand-edited.

## Setup and migration

Run `make doctor` to inspect installed prerequisites. `make bootstrap` installs
pinned tools and project Go modules; `make frontend-install` establishes or repairs
the frontend installation proof. Browser and service-image recovery use the exact
supported targets printed by the closed diagnostic mapping. Then start
`make ui-review UI_MODE=seeded REVIEW_PROFILE=default`, or select
`REVIEW_PROFILE=network_flow_claimed`. No smoke run is required.

Before upgrade or rollback, stop active sessions with their exact `UI_SESSION`
locators using the running version and confirm cleanup. Update result producers,
validators, control readers and guidance together. Start fresh sessions; v1
receipts remain historical and are not translated into v2. Do not delete unresolved
ownership proof. Capture/analysis requests, bundles and locators remain unchanged. Browser action
requests use v2: a bounded CSS zoom action (100 or 125 percent) is necessary to
perform the requested rendered zoom scenario through the supported interface.
It adds no arbitrary evaluation or private browser attachment escape hatch.

## Verification evidence

Completed commands and exact run roots:

| Command | Outcome | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make generate` | Pass after coordinated contract updates and final preparation changes. | `20260930T023641Z-p14021` |
| `make bootstrap` | Pass; explicit dependency setup. | `20260930T012908Z-p69449` |
| `make frontend-install` | Pass; installation proof published. | `20260930T014308Z-p99878` |
| `make doctor` | Pass after narrowing the Go dependency predicate. | `20260930T013446Z-p12847` |
| `make format` | Pass for touched Go sources. | `20260930T012646Z-p39754` |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_contract,harness.browser.boundary_support.ui_review_lifecycle CARTULARY_HARNESS_CACHE_MODE=off` | Pass, 2/2 units. | `20260930T023512Z-p9094` |
| `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity CARTULARY_HARNESS_CACHE_MODE=off` | Pass, 1/1 unit. | `20260930T012957Z-p71500` |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.ui_review_seeded_default,harness.browser.integration.ui_review_seeded_network_flow_claimed CARTULARY_HARNESS_CACHE_MODE=off` | Pass, 4/4 units; isolated empty preparation caches, editor/viewer and artifact workflows. | `20260930T021418Z-p94550` |
| `make agent-finalize` | Pass, 1/1 unit; retained-run maintenance skipped because `RESULTS_DIR` was unset. | `20260930T023655Z-p17324` |
| `make generate-drift` | Pass, 4/4 units. | `20260930T023728Z-p21639` |
| `make json-shape-check` | Pass, 3/3 units. | `20260930T023728Z-p21649` |
| `make lint-scripts` | Pass, 2/2 units. | `20260930T023511Z-p8875` |
| `make lint-shell` | Pass, 4/4 units. | `20260930T021513Z-p59428` |
| `make lint-markdown` | Pass; ad hoc documentation summary. | `20260930T024923Z-p339` |
| `make harness-contract` | Pass, 2/2 units. | `20260930T023728Z-p21865` |
| `make check` | Pass, 982/982 units; includes both seeded profile workflows. | `20260930T023727Z-p21438` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.desktop_built_in_surface_selector_keyboard_navig_ebad53c0ea CARTULARY_HARNESS_CACHE_MODE=off` | Pass, 2/2 units. | `20260930T014925Z-p28255` |
| `make explain-target TARGET=ui-review DETAIL=artifacts` | Pass; locator and terminal receipt, no generic run summary. | Read-only; no run root. |

The routed workbook command passed 14/14 units at
`20260930T014925Z-p28257` (counts include fixture/setup units):

```bash
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.desktop_built_in_surface_selector_keyboard_journ_3475c2c6d2,module.workbook.browser.built_in_selector_entry_from_network_analysis_8f5bebcd07,module.workbook.browser.built_in_grid_entry_editor_59f8176b4a,module.workbook.browser.built_in_grid_entry_supersession_5f9a05b19f,module.workbook.browser.built_in_grid_entry_viewer_c3bbf9714f,module.workbook.browser.verify_system_views_switcher_keyboard_entry_rovi_90a2f62956,module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159 CARTULARY_HARNESS_CACHE_MODE=off
```

Final focused qualification additionally passed:

```bash
make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_contract,harness.browser.boundary_support.ui_review_lifecycle,harness.browser.boundary_support.installed_only_images,harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131 CARTULARY_HARNESS_CACHE_MODE=off
make service-backed-test-slice OWNER=package.grid_adapter ROWS=package.grid_adapter.accessibility.verify_grid_cells_editors_group_rows_active_cell_3156bd379d CARTULARY_HARNESS_CACHE_MODE=off
```

Their roots are respectively `20260930T021205Z-p50814` (5/5 units) and
`20260930T021305Z-p54266` (11/11 units). The lifecycle route now includes one
composed fault fixture: a real failing Make child crosses private diagnostic and
preparation IPC boundaries into the session terminal receipt and human recovery
guidance, followed by private cleanup and byte-identical repeated stop. Registry
parity passed again at `20260930T021306Z-p54483`; doctor passed at
`20260930T021239Z-p52580`; `make agent-finalize` passed again at
`20260930T021412Z-p94074`, with retained-run maintenance still skipped because
`RESULTS_DIR` was unset.

Implementation failures were related to this change and were repaired: optional
absent `.npmrc` handling in proof publication; missing summary-schema metadata;
bootstrap GO forwarding; inconsistent test-owned lease identity; the image-pull
API's interface return type; and an overbroad Go dependency query. Their failures
were not attributed to the historical incident. The Go prerequisite failure in
`20260930T013201Z-p91207` reached human output with configuration class, prerequisite
phase, Go subject and bootstrap recovery guidance; both profile fixtures completed
cleanup. The earlier compiler failure is recorded in `20260930T012957Z-p71505`.
Later lifecycle fixtures caught raw release errors being treated as operational
failures (`20260930T014014Z-p95527`) and duplicate stop IPC
(`20260930T014330Z-p801`). Both were repaired before the passing lifecycle run.
The installed-only image row passed in both of those runs; their overall failures
are not counted as passing suites. A schema draft mismatch was also corrected to
the repository's supported 2020-12 dialect before generation passed.

The first broad `make check` at `20260930T020302Z-p45672` found a pre-existing raw
vendor selector in `workbook.a11y.spec.ts`, confirmed in the starting commit. The
run was gracefully cancelled after that failure (864/982 units completed), then
the focused architecture and grid accessibility rows above passed. The test now
observes the focused semantic group row, preserves its focus/toggle assertions,
and names that observation accurately. No grid implementation changed. The shared
doctor's remaining pnpm version probe also explicitly disables Corepack network
acquisition, matching the read-only readiness contract.

The next full `make check` at `20260930T021618Z-p81808` completed 981/982 units
successfully. Its only failure was a pre-existing selector consumer in
`WorkbookShell.surfaces.test.tsx` that still queried built-ins as buttons. The
query now uses the named built-in tablist and tab roles, retaining the exact
registry-order assertion without prefix filtering. No product behavior changed.
The complete affected suite passed at `20260930T022836Z-p53827` (2/2 units):

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e CARTULARY_HARNESS_CACHE_MODE=off
```

`make lint-biome` also passed at `20260930T022837Z-p54240` (2/2 units).

Final code review found cleanup-only errors in shared preparation could inherit
the last seeding phase's classification. Cleanup now has an explicit owner outcome
and retains its context across stop/IPC; an earlier operational failure remains
primary. Fixture classification is applied only at the seed operation, so later
publication or IPC exceptions remain unclassified. Added fixtures cover both
precedence and cleanup-only cancellation. Their first focused run passed at
`20260930T023301Z-p3777` (2/2 units). The in-flight broad run
`20260930T022940Z-p59679` was gracefully cancelled at 273/982 units to make this
correction; it is not counted as completed qualification.
The final lifecycle rerun passed at `20260930T023512Z-p9094` (2/2 units), and
script lint passed at `20260930T023511Z-p8875`. Finalizer run
`20260930T023509Z-p8476` failed because the changed preparation owners made topology
inputs stale; `make json-shape-check` confirmed this at
`20260930T023607Z-p11831`. Regeneration from authored inputs through Make passed,
followed by the finalizer and drift/shape checks recorded in the table above.

## Rendered review and product follow-ups

Used the repository UI-review skill and public Make interface only. Each capture's
original bytes and SHA-256 were checked before viewing. Eight private screenshots
were inspected before stopping; none were retained or copied into documentation.
No golden or product implementation changed in this harness slice.

| Session | Inspected state and outcome |
| --- | --- |
| `20260930T015036Z-p66667`, default | Editor and viewer; desktop 1440×900, narrow 1024×720 and 125% CSS zoom. Tab entered the selected tab; arrows wrapped; Home/End moved focus without selection; Enter/Space activated. The visible selector focus ring stayed within its strip. Evidence creation draft survived switching away/back. Compact replacement, selected menu entry, Escape dismissal and return to desktop preserved focus. A focused Recovery button retained focus across resizing. System views roved from Indicators to Assessments and explicit activation entered its grid. A newer Recovery focus remained after destination activation. Viewer activation entered an existing read-only Hosts cell. |
| `20260930T015955Z-p10941`, network_flow_claimed | Editor and viewer; desktop, narrow and 125% CSS zoom. Network Analysis left all built-ins unselected. Tab entered Timeline; arrow/Home/End navigation preserved the extension; explicit Space/Enter entered Hosts. Editor destination was the Display Name creation control; viewer destination was an existing read-only cell. Compact replacement and Escape dismissal preserved focus. |

Fresh axe scans completed on desktop default editor/viewer, claimed Network
Analysis editor and narrow claimed viewer. Layout-only follow-up captures disabled
axe after the fresh scan; disabled results are not accessibility passes. Every scan
had one incomplete result. Raw advisory findings require these separate follow-ups:

| ID | Owner | Finding and completion criterion |
| --- | --- | --- |
| UI-REVIEW-A11Y-01 | `package.grid_adapter`, with `module.workbook` integration evidence | Closed by the bounded correction: ordinary absent/empty gutter content renders Row; explicit Source row and grouping labels retain their meaning. Fresh targeted-rule execution and production/test-binding parity pass. Full refreshed visual catalog passes two ordinary validations. |
| UI-REVIEW-A11Y-02 | `web.workbook`, with `module.workbook` accessibility evidence | Closed by the bounded correction: the shared identity disclosure owns one native incident h1, with its region outside the heading and truthful unavailable states. Fresh default/claimed targeted-rule, identity-update, disclosure/navigation and two ordinary visual validations pass. |

The [bounded structural correction handoff](ui-ux/workbook-structural-accessibility-correction-handoff.md)
records the fresh evidence, owner decisions, full-catalog golden refresh authorized
by the user, incomplete contrast results and separate constrained Network Analysis
scrollability advisory. The historical observations and receipts below remain
unchanged; current implementation evidence does not retroactively change them.

These findings are advisory, not proven selector regressions. The routed
accessibility check passed. Rendered observations cover the selected scenarios,
not all assistive technologies, browser zoom mechanisms, contrast measurements or
product acceptance. Delayed destination-mount cancellation and empty-grid fallback
are established by the routed deterministic tests; the manual review did not
inject a held mount. Default-session receipts count five failed requests; those
counters alone do not identify a product or infrastructure defect. No console
errors were observed in either terminal count. One ambiguous Open target in the
claimed-profile directory was rejected without a click; a fresh snapshot supplied
the exact reference before continuing.

Exact-session stop succeeded for both locators under their run root's
`ui-review/session.json`. Each terminal receipt records exit 0 and cleanup complete;
repeated stop returned the same digest:

- Default: `4e6fb1f16ec34d26b6b0261ee68be861f9f735b01e8784c7888de5cb7962548c`.
- Claimed: `54d889f76fe1bd2f9b79d34d944713a323fcc56627edc9bda7c49f05d2df1e7a`.

Both foreground processes exited, both private runtime directories were absent,
and caller-owned request scratch was removed. Retained UI output is structural
only. The original selector handoff now points here for closure of its formerly
blocked rendered gate; its historical evidence remains intact.

## Completion and handoff

All seven workstreams have implementation or review evidence above. The final
full check completed successfully in 668.455 seconds, with zero failed, skipped
or cancelled units. Both seeded profile workflows passed again within that run.
Executable inputs remained unchanged during the final check; Markdown updates
record human handoff evidence. Its source digest is
`sha256:7c07a90c17585b885592b42127d89542ad02401810564bbbd615fc49d24fd924`.
The run recorded zero result-cache hits (805 misses and 177 bypasses).
Generated projections came from Make, and lockfiles and
visual goldens were not changed. The two stale product-test consumers were repaired
without changing application behavior.

Retained-run finalizer maintenance was skipped because `RESULTS_DIR` was unset;
no successful full warm result was supplied to those earlier finalizer invocations.
No commit, deployment or release publication was performed. Historical root-cause
uncertainty remains; no speculative diagnosis replaces it. Product owners should
triage UI-REVIEW-A11Y-01 and UI-REVIEW-A11Y-02 separately from this completed harness
remediation. Those two product findings were subsequently corrected and closed
with the fresh evidence in the bounded correction handoff linked above. That
closure does not dispose of the separately recorded Network Analysis constrained
scrollability advisory or certify screen-reader/WCAG conformance.
