# Workbook Conflict and Queued-Edit Recovery Surface Refactor Handoff

## 1. Baseline and authority

| Item | Recorded value |
| --- | --- |
| Started | `2026-08-31T08:53:11-04:00` |
| Branch | `main` |
| Commit | `e894fbfefff1fcf172ce14b6f8f290f6e216a0dc` (`Imports Specification and Implementation`) |
| Initial Git status | Clean |
| Tool versions | Git `2.53.0`; Node `v24.15.0`; pnpm `10.33.0`; Go `1.26.6`; Python `3.14.4`; jq `1.8.1` |
| Digest localization | Digest last changed at `1a580e5afca1c6e7f85de88592992773ff7f6c98`; current repository commit differs. Live owners and repository inputs were revalidated. |
| Advisory drift | `START_HERE.md` points to absent `docs/handoffs/cartulary-ui-ux-remediation-handoff.md`. This is advisory drift, not product authority or an owner contradiction. |
| Existing user changes | None at start. Any later unrelated changes remain user-owned and will be preserved. |

Authority was read in the requested order. Core 03 owns pending-queue and
mutation behavior. `docs/design.md` owns observable presentation. The only
identified owner reconciliation is the explicitly authorized disagreement for
the client transaction recovery live priority: §§10.5 and 14.2 require polite,
while §10.8 and its machine projection currently say assertive.

Authorized paths are limited to:

- `docs/design.md` and this handoff;
- `contracts/design/presentation.v1.json` and its Make-generated UI projection;
- workbook shell, recovery, same-field resolver, status, Timeline notice,
  pending-replay, and narrowly related presentation sources under
  `apps/web/src/workbook/**`;
- authored semantic selector sources/tests under `packages/ui-contracts/src`;
- `tools/frontend_source_ownership.json`, because the structural split must be
  accounted for by the authored documentation-free source-ownership manifest;
- relevant routed frontend, accessibility, keyboard, measurement, and visual
  tests plus reconciled workbook visual goldens;
- authored verification/catalog inputs only if an existing routed test cannot
  absorb required evidence, followed by Make-owned generated topology.

`contracts/design/tokens.v1.json` is inspected but not authorized for change:
the required opaque surfaces, borders, elevation, focus, secondary action, and
destructive action tokens already exist. Generated roots will not be hand
edited. Runtime, tests, generators, conformance, and release evidence will not
read or depend on this handoff or another Markdown file.

## 2. Workstream ledger

| Workstream | Status | Exit condition |
| --- | --- | --- |
| CR-01 — Owner and projection reconciliation | DONE | Owner text, authored projection, generated projection, and package test agree on one polite announcement. |
| CR-02 — Characterization | DONE | Every claimed token, opacity, duplication, safe-copy, focus, action-state, and responsive gap is reproducible. |
| CR-03 — Semantic component split | DONE | Queued-edit recovery and same-field resolution are distinct shell-owned components with only visual framing shared. |
| CR-04 — Presentation and focus correction | DONE | Safe copy, valid opaque tokens, single recovery locus, focus routing/fallback, action state, and supported layout bands satisfy owners. |
| CR-05 — Validation and handoff | BLOCKED | Task-owned checks and goldens pass, but repository-wide visual and webserver-backed targets retain unrelated pre-existing failures described below. |

Only one row may be `IN_PROGRESS`. Before starting a row, its predecessor must
be `DONE`; after each row, record paths, decisions, commands, results, retained
run roots, compatibility, risks, rollback, and the next action. If another
adopted-owner contradiction is found, set the affected row to `BLOCKED`, record
`BLOCKED: owner contradiction`, and stop that workstream.

## 3. Characterization baseline

The current implementation was inspected across the combined conflict
resolver, status strips, Timeline notices/composition, shell mutation runtime,
pending replay model, public-error presentation, design contracts/tokens, and
the routed unit, accessibility, keyboard, and visual scenarios.

Confirmed gaps:

1. Recovery/resolver styles reference five undefined variables:
   `--ct-colors-surface-raised`, `--ct-colors-surface-muted`,
   `--ct-colors-border-subtle`, `--ct-colors-border-strong`, and
   `--ct-shadow-lg`.
2. The current recovery and resolver goldens visibly show grid, toolbar, and
   notice content through their panels.
3. `WorkbookConflictResolver` owns both queued-edit recovery and same-field
   value resolution.
4. Timeline renders a second queued-edit recovery notice beneath the
   action-bearing shell panel.
5. Conflict activation focuses that Timeline notice instead of the real panel.
6. The panel heading is `Pending edit recovery`, not `Queued edits`.
7. Halt presentation is derived through transport-provided message text and can
   expose `client_txn_conflict` or other implementation terminology.
8. Retry does not share the immediate transition guard used by discard.
9. Recovery-specific narrow, compact, zoom, text-spacing, long-text, and
   multiple-conflict evidence is incomplete.

## 4. Advisory query dispositions

The two required offline searches ran with `PYTHONDONTWRITEBYTECODE=1` and
`python3 -B`; `--design-system`, `--persist`, and `--force` were not used. The
digest snapshot was not mutated.

| Result | Disposition | Cartulary application |
| --- | --- | --- |
| Clear recovery next steps | ADOPT | Use only the owner-permitted retry and discard actions. |
| Visible focus | ADOPT | Preserve the owned focus ring and unobscured keyboard focus. |
| Toast auto-dismiss | REJECT | Unresolved recovery persists until its state changes. |
| Error near the problem | ADAPT | Use one shell-owned same-surface recovery locus. |
| Confirmation dialog | REJECT | No owner requires a modal; direct discard remains non-modal. |
| Error live announcement | ADAPT | Use exactly one safe polite recovery live region, not a generic alert. |
| Disabled state | ADOPT | Use native disabled semantics with owned component tokens. |
| Field-level error placement | REJECT | Queue recovery is not field validation; same-field conflicts remain cell anchored. |
| Modal focus trap/return | ADAPT | Do not trap or steal focus; return focus deterministically only when a focused panel disappears. |
| Semantic HTML | ADOPT | Use an `aside`, headings, labels, and native buttons. |
| Form labels | ADOPT | Preserve explicit same-field editor labels. |
| Dynamic announcements | ADAPT | Announce typed safe recovery copy once with polite priority. |
| Lift state up | ADAPT | Lift only global recovery targeting/focus to the shell; keep workflow state separate. |
| Local state | ADOPT | Keep only transition, message, and active-conflict UI state locally. |
| Avoid unnecessary state | ADOPT | Derive copy and action availability from typed snapshots. |
| New reducer | REJECT | Existing runtimes remain the state-machine owners. |

## 5. Checkpoint log

| Time | Workstream | Paths and decisions | Commands and results | Compatibility, risk, rollback, next action |
| --- | --- | --- | --- | --- |
| `2026-08-31T08:53:11-04:00` | Baseline / CR-01 start | Created this handoff; recorded clean branch/commit/toolchain, authority order, authorized paths, digest drift, characterization, and advisory dispositions. Existing token contracts cover the implementation; no token alias or token-contract edit is planned. | Baseline Git/tool commands passed. Both required offline advisory searches returned eight results and made no file changes. | No product interface or stored-state change. Risk is confined to the authorized workbook presentation seam. Rollback this handoff with the slice. Next: reconcile §10.8 and the authored presentation projection. |
| `2026-08-31T08:56:17-04:00` | CR-01 completion / CR-02 start | Reconciled `docs/design.md` §10.8 and `contracts/design/presentation.v1.json` to polite; added an exact package projection assertion; generated only `packages/ui-contracts/src/generated/design-presentation.ts`. Recorded current routing for `web.workbook`, `web.design`, `package.ui`, `module.workbook`, and `module.collaboration`. | `make generate` passed at `.cartulary/test-results/20260831T125512Z-p36940`. `make test-slice OWNER=package.ui` passed 10/10 at `.cartulary/test-results/20260831T125558Z-p40950`. Generated diff inspection found only the expected input digest and `live: "polite"` change. | Owner and projection now agree; no token/API/storage change. Rollback CR-01 as the owner, authored projection, generated TypeScript, and package assertion together. Next: add characterization assertions before structural movement. |
| `2026-08-31T09:00:31-04:00` | CR-02 completion / CR-03 start | Extended existing routed Workbook/Collaboration tests before moving code. Assertions cover owned opaque tokens, absence of all five invalid variables, secondary same-field actions, exact safe copy, terminal discard-only behavior, one polite region, duplicate-notice removal, and focus on the real panel. Existing 1440×900 recovery and 1280×720/1440×900 resolver goldens were inspected and show underlying chrome/content through both panels. | The expected-red Workbook row failed at `.cartulary/test-results/20260831T125934Z-p44063` because both recovery states still render `Pending edit recovery`; the expected-red Collaboration row failed at `.cartulary/test-results/20260831T130008Z-p44749` because the resolver still uses `--ct-colors-surface-raised`. An earlier harness run `.cartulary/test-results/20260831T125828Z-p42986` exposed unsupported jest-dom matchers in the new assertions; those assertions were corrected before the product-gap rerun. | These failures are characterization evidence, not accepted final results. No owner contradiction was found. Rollback the added assertions only with the complete slice. Next: split visual framing, queued-edit recovery, overflow, and same-field resolution without changing state-machine semantics. |
| `2026-08-31T09:22:59-04:00` | CR-03 completion / CR-04 start | Replaced the combined resolver with `RecoverySurface`, `WorkbookEditRecoveryPanel`, `WorkbookQueueOverflowNotice`, and `WorkbookSameFieldConflictResolver`; moved snapshot selection and priority to the shell; removed Timeline recovery refs/focus commands and blocked/overflow notices; made all Base status strips use the shell activation port. The Timeline test fixture now composes the same distinct panels. | `make format` passed at `.cartulary/test-results/20260831T131943Z-p58904`; `make frontend-typecheck` passed at `.cartulary/test-results/20260831T132208Z-p68418`. The characterized same-field row passed at `.cartulary/test-results/20260831T132227Z-p68891`; the characterized queued-edit row passed at `.cartulary/test-results/20260831T132243Z-p69486`. Earlier format roots `.cartulary/test-results/20260831T131735Z-p49909` and `.cartulary/test-results/20260831T131846Z-p54428` identified and led to removal of incomplete focus/ref shapes; typecheck root `.cartulary/test-results/20260831T132001Z-p63125` identified the remaining fixture/composition cutover. | Workflow state/action semantics are no longer shared; only visual framing is shared. No invalid recovery token name remains in component or Timeline sources. No route, storage, or queue algorithm changed. Rollback the four components, shell/facade/status wiring, Timeline composition cleanup, runtime presentation boundary, and fixture together. Next: complete focus fallback, action-state, safe-copy, selector, responsive/a11y, and visual evidence. |
| `2026-08-31T09:50:20-04:00` | CR-04 completion / CR-05 start | Added typed safe client/terminal presentation, synchronous shared action disabling and duplicate guard, one polite atomic recovery region, shell status activation priority, removal fallback, retry-to-resolver focus, the active-surface fallback selector, existing-token opaque framing, secondary same-field actions, wrapping, containment, and narrow/compact/zoom/text-spacing coverage. Updated the existing stateful and queue assertions to target the shell panel rather than the removed Timeline duplicate. | `make frontend-typecheck` passed at `.cartulary/test-results/20260831T134004Z-p40692`; focused queue/recovery rows passed 3/3 at `.cartulary/test-results/20260831T134154Z-p46398`; `make browser-e2e-a11y` passed 12/12 at `.cartulary/test-results/20260831T134842Z-p47908`. The first full `module.workbook` slice root `.cartulary/test-results/20260831T132737Z-p71218` exposed one relevant raw-message queue expectation, the removed notice in stateful/a11y tests, expected recovery golden drift, plus unrelated incident-administration and non-workbook golden drift. Follow-up a11y roots `.cartulary/test-results/20260831T134244Z-p47397` and `.cartulary/test-results/20260831T134557Z-p97803` respectively identified synthetic CSS-zoom block measurement and unmount blur clearing focus ownership; both were corrected without weakening base/narrow/compact/text-spacing assertions. | Presentation contains no transport code, identifier, route, token, payload, or stack text. Pending admission, secure re-key, FIFO, discard, and same-field semantics remain unchanged; the action result union is internal. The CSS 200% emulation retains panel/action/horizontal/overlap checks while document block extent is asserted at native viewport and text-spacing layouts because CSS `zoom` doubles synthetic `100vh`. Rollback CR-04 through the shell/status selector, typed presentation/runtime boundary, component styles/actions, and related tests. Next: run the full requested generation, routed frontend/browser matrix, reconcile and inspect every changed golden, then finalize scope and handoff. |
| `2026-08-31T14:54:00-04:00` | CR-05 terminal checkpoint | Reconciled all visual captures through the Make-owned updater, restored every unrelated command-produced PNG, and retained exactly eight owned recovery/resolver goldens. Manual review found and led to removal of a duplicate Timeline stale-error banner, content-sized recovery framing, and more compact resolver comparisons/editing at narrow bands. The source-ownership manifest now accounts for all four split components and the typed presentation module. Final scope contains only the owner/projection/generated presentation, workbook source/tests, semantic selector/package tests, source-ownership projection, eight workbook goldens, and this handoff. | Generation passed: `make generate` `.cartulary/test-results/20260831T135135Z-p93599`; final `make generate-drift` `.cartulary/test-results/20260831T145348Z-p42254`; generated policy `.cartulary/test-results/20260831T135205Z-p99402`; JSON shape `.cartulary/test-results/20260831T135212Z-p99858`; catalog `.cartulary/test-results/20260831T135222Z-p791`; semantic identity `.cartulary/test-results/20260831T135236Z-p1327`. Frontend passed: typecheck `.cartulary/test-results/20260831T145333Z-p41775`; unit 391/391 `.cartulary/test-results/20260831T144704Z-p28210`; import boundary `.cartulary/test-results/20260831T142141Z-p34713`; Biome `.cartulary/test-results/20260831T142154Z-p35158`. Focused owners passed: `package.ui` 10/10 `.cartulary/test-results/20260831T142207Z-p35626`; `web.workbook` 139/139 `.cartulary/test-results/20260831T142221Z-p36077`; `web.design` 15/15 `.cartulary/test-results/20260831T142307Z-p51131`. Browser task evidence passed: final a11y 12/12 `.cartulary/test-results/20260831T145024Z-p93041`; measurement 22/22 `.cartulary/test-results/20260831T143415Z-p63067`; stateful 34/34 `.cartulary/test-results/20260831T143847Z-p20100`; visual update 12/12 `.cartulary/test-results/20260831T141332Z-p44911`. Finalization passed `.cartulary/test-results/20260831T145202Z-p37658`; `test-fast` passed 443/443 `.cartulary/test-results/20260831T145220Z-p40577`. | Compatibility is source-only and additive at the internal selector boundary. `CR-05` is `BLOCKED`, not falsely complete: final full visual validation `.cartulary/test-results/20260831T144805Z-p47185` fails only untouched non-slice goldens; `module.collaboration` `.cartulary/test-results/20260831T142422Z-p98002` and its service-backed run `.cartulary/test-results/20260831T142731Z-p95250` fail only untouched presence goldens; `module.workbook` `.cartulary/test-results/20260831T142922Z-p45422`, its service-backed run `.cartulary/test-results/20260831T143158Z-p5787`, and standalone webserver-backed `.cartulary/test-results/20260831T144112Z-p68949` additionally retain the unrelated incident-administration compact-padding assertion (`expected 2px 5px`, computed `0px 5px 2px`). No owner contradiction was found. Next action is owner-authorized reconciliation of those unrelated baseline failures; this slice does not accept or mutate them. |

## 6. Visual refresh record

Accepted trigger: the adopted opaque recovery framing, component-boundary,
responsive-layout, and same-field secondary/destructive-action corrections.
The successful update target used Chromium, device scale factor `1`, dark
graphite, default density, `100%` capture zoom, declared masks, and the existing
top-left/right-edge scroll normalization. Screenshot scope did not change.

| Owner row | Capture identities retained | Viewport |
| --- | --- | --- |
| `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | `timeline-mutation-transaction-recovery-panel`; `timeline-mutation-transaction-recovery-panel-narrow`; `timeline-mutation-transaction-recovery-panel-compact` | `1440x900`; `1024x720`; `768x640` |
| `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `collaboration-conflict-resolver` (registered `D-VFIX-003`) | `1280x720` |
| `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | `collaboration-grid-conflict-resolver`; `collaboration-grid-conflict-resolver-narrow`; `collaboration-grid-conflict-resolver-compact` | `1440x900`; `1024x720`; `768x640` |
| `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc` | `collaboration-grid-blocked-conflict` | `1280x720` |

All eight retained PNGs were inspected after the final successful update.
They have opaque backgrounds, valid borders/elevation, no content bleed, no
duplicate blocker banner, deterministic long-token wrapping, secondary
resolution controls, destructive draft-discard styling, and no panel/status
overlap. Compact resolver actions remain in the panel's owned vertical scroll
region; the final a11y run scrolls the last action fully into view and proves no
horizontal or document-level page scrolling. Thirty-one unrelated PNG changes
produced by each full update were restored to their clean baseline and are not
accepted by this slice.

## 7. Final compatibility and rollback

This is a workbook presentation,
accessibility, and internal component boundary correction. It introduces no
public API, route, authorization, storage, persisted-data, or migration change.

Rollback is source-only: revert the owner/projection/generated presentation,
frontend component/runtime/selector/test changes, reconciled goldens, and this
handoff together. No data rollback, migration reversal, or external cleanup is
required.

Post-check evidence: `make lint-markdown` passed at
`.cartulary/test-results/20260831T145559Z-p45755`; `git diff --check` passed
with no output. The commands were rerun after recording this evidence so the
final handoff bytes are covered.
