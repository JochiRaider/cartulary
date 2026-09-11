# Workbook Decision supersession refactor

Delivery status: DS-01–04 complete. DS-05 is blocked by a reproducible shared
harness frontend-artifact publication race. The final full check passed 833/834
units; no failing gate is waived and no complete full-check claim is made.

## Authorization and baseline

One bounded Decision workflow: review, admission, retained exact recovery, and
receipt reconciliation. No Task, Timeline, approval, general relationship,
storage, dependency, migration, or public-interface redesign is authorized.
No commit, reset, push, deployment, or analyst-data cleanup will be performed.

Baseline and implementation recheck: branch `main`, HEAD
`8f299b8aff548a6059fbf50ef3d764a60e39ec7d`, clean index and working tree.
Only root `AGENTS.md` applies. The digest localized read order and relevant
owner sections were read during planning and revalidated against source.
Historical handoffs supply evidence, not additional instructions or scope.
The latest merge handoff records full `make check` PASS 826/826 at
`.cartulary/test-results/20260910T231454Z-p87451`; no older waiver applies.

`make doctor`: PASS at `.cartulary/test-results/20260911T002128Z-p51850`.
Actual Go 1.27.1, Node 24.15.0, pnpm 10.33.0, ShellCheck 0.11.0;
React 19.2.5, TypeScript 6.0.2, Vite 8.0.8, Vitest 4.1.4,
Playwright 1.59.1. Doctor reports only incomplete inspection of inotify usage.

Both requested task guides passed during planning and again before edits:

```bash
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=module.tasksdecisions
```

Collaborating verification owners: `module.workbook`, `module.revisions`,
`platform.httpruntime`, and `package.ui`; affected design/browser rows retain
their authored owners. A planning lookup of `platform.httpapi` failed because
that owner does not exist; the corrected `platform.httpruntime` guide passed.

Permitted authored paths are Decision coordination features and necessary
workbook inspector/runtime/transport/query/composition, focused tests and browser
support, UI selectors, verification contracts/catalog/topology inputs, affected
reviewed visual artifacts, and this handoff. Generated derivatives are produced
only through Make. Existing digest and completed handoffs remain untouched.

## Governing owners

- Core 01 REQ-01-086–088, especially 087A: request, replay, Decision receipt,
  attribution, versions, and replayable events; REQ-01-615–617: semantic action
  placement/roles/confirmation; REQ-01-339–341: Decisions projection and writes.
- Core 02 §10.4.2/§10.4.2.1, REQ-02-118/119 and typed-link/history owners:
  machine consistency, directed supersession, executed-target distinction.
- Core 03 REQ-03-291/292, committed-version and continuity requirements,
  REQ-03-301, session recovery and §16.4: selected subject, immutable review,
  secure identity, dependent writes, workbook-native placement.
- Core 04 current role/visibility, CSRF, concealment, and session lifecycle.
- Design §§7.3, 10, 12, 14: local feedback, dense inspector, explicit destructive
  review, keyboard operation, focus, and non-color state. Digest advice is
  ADOPT for local recovery/accessibility and ADAPT for destructive confirmation.

## Gap ledger

| Gap | Owner | Areas/remediation | Rationale | Long-term benefit | Compatibility | Unresolved risk | Binary completion |
| --- | --- | --- | --- | --- | --- | --- | --- |
| G1 Private Workflow form and independently chosen target | Core 01 615–617; Core 03 291/292 | Inspector semantic dispatch; remove private Decision form | Current action bypasses declared History review | One entry and selected-subject identity | Task unchanged | Accidental duplicate composition | Exactly one History action; no target picker |
| G2 Grid-count gate and lossy reference options | Core 01 Decisions/query; Core 02 10.4.2.1 | Decision paged query reader and stateful candidates | One-row filters conceal valid replacements; broker drops paging/failures | Trustworthy discovery independent of grid | Existing public query; broker unchanged | Hidden or incomplete reads mistaken for empty | Outside candidates, paging, disabled reasons, retry and concealment pass |
| G3 Mutable inputs with no admission guard | Core 03 review/committed versions | Captured review and runtime sequencing | Async/disposable state permits stale or duplicate dispatch | Auditable immutable intent | Only target base version on wire | Earlier writes or authority change during preparation | Stale confirmation sends zero; double confirm sends once |
| G4 Per-call identity and component lifetime | Core 01 087; Core 03 301 | Dedicated retained owner and captured transport | Lost replies cannot be safely recovered | Explicit exact recovery survives inspector closure | Memory only; no lock/FIFO protocol | Uncertainty mistaken for rejection | Replay uses identical path/body/ID after target changes |
| G5 Discarded receipt and refresh-owned completion | Core 01 087A/088; Core 02 118/119 | Complete receipt validation, high-water versions, reconciliation | Accepted change can appear failed and replacement stays stale | Reliable acknowledgement and independent refresh recovery | Existing history/rollback unchanged | Older query or late callback corrupts current selection | Receipt precedes refresh; both rows reconcile; refresh retry sends zero |

## Tracker

| Row | Status | Exit |
| --- | --- | --- |
| DS-01 | DONE | Baseline, owners, ledger and failing characterization recorded |
| DS-02 | DONE | Declared action, candidate states and immutable review tests pass |
| DS-03 | DONE | Admission, retained transport and lifecycle recovery tests pass |
| DS-04 | DONE | Receipt, deterministic/service/browser and affected visual evidence pass |
| DS-05 | BLOCKED | Shared frontend artifact publication prevents a passing full check; separate harness scope required |

## DS-01 log

Inspected the requested coordination, generic inspector, command-port, Decisions
contract, admission and supersession source files; query broker/adapter, runtime,
history, entity-merge retained owner, HTTP projections, owner routing and visual
maintenance. Confirmed all five suspected gaps. Backend already supplies the
Decision variant and executed-target status behavior; no backend change planned.
The query route already supports cursor pagination. UI candidate reads will retain
that metadata in a Decision-specific adapter rather than changing the broker.
Visible ineligible candidates remain listed with reasons, as selected by user.

Baseline narrow command PASS 3/3 at
`.cartulary/test-results/20260911T003642Z-p55802`:

```bash
make test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.frontend_unit.coordination_workflow_lifecycle_7a8d13c2f0,module.tasksdecisions.source_mutations.task_and_decision_mutation_admission_and_replay_6d4ea5900a
```

Characterization added at
`apps/web/src/workbook/features/coordination/decisionSupersession.characterization.test.tsx`
and routed in `tools/test_families/web.workbook.json`.

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.decision_supersession_characterization
```

Expected FAIL at `.cartulary/test-results/20260911T004230Z-p59488` (1/2 graph
units). The assertions require exactly one declared History action and retention
of normalized receipt reason. Both fail against the baseline. Failure is directly
related to the seam; assertions will move with their production boundaries rather
than be weakened. DS-01 exits are met. No product code changed in DS-01.

## DS-02 log

Added the dedicated Decision model, paged read adapter, candidate hook, editor,
and owner/context interfaces under workbook coordination and adapters. Generic
inspector composition dispatches the exact canonical History identity and hosts
the editor. Removed the private Decision Workflow form and controller branch.
The new UI waits for the runtime provider activated in DS-03; it cannot fall back
to the old unsafe submission. Task behavior remains separate and unchanged.

Focused placement and review PASS 3/3 units at
`.cartulary/test-results/20260911T005303Z-p64907`:

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.decision_supersession_characterization,web.workbook.regression.decision_supersession_review
```

This covers outside-grid replacement discovery, duplicate-label identity and
disabled reasons, partial-page failure/retry, review cancellation and version/
reason invalidation, detached reads, machine-state and Unicode reason handling.
Earlier runs exposed test-helper mistakes (unsupported jest-dom matchers and
missing cleanup) and an unsorted authored selector list; corrected without
weakening behavior assertions. Typecheck identified two mock Promise type
widenings; explicit port types correct them. Receipt characterization remains a
separately routed expected failure until DS-03 transport replacement.

Compatibility: no public interface, dependency, stored-data, or Task change.
Remaining risks belong to DS-03 admission/lifetime and DS-04 reconciliation.
DS-02 exits met. Next: activate dedicated retained owner and exact transport.

## DS-03 log

Source inspection demonstrated a bounded owner mismatch: Decision supersession
admission calls `fieldnorm.NormalizeNote`, which has no length bound and rejects
Unicode format characters. REQ-01-491/496 requires 4096 normalized scalars and
only C0/C1 rejection except LF/TAB. Correct only Decision supersession admission,
with focused admission evidence; do not change shared normalization or other
routes. This is an existing-contract repair within the authorized Decision seam,
not a public-interface or backend subsystem expansion.

DS-03 implementation is connected through `WorkbookMutationRuntime`, shell
infrastructure/authority, the dedicated Decision context and recovery surface,
captured-request adapter, and Decision direct-write admission. Task command
ownership remains unchanged; the obsolete Decision command was removed.
Clipboard inspection confirmed Decisions have no paste route, so that adapter
remains unchanged. History/FIFO/direct Decision writes coordinate locally;
no server lock or queue ownership was added.

```bash
make frontend-typecheck
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.decision_supersession_characterization,web.workbook.regression.decision_supersession_receipt_characterization,web.workbook.regression.decision_supersession_review,web.workbook.regression.decision_supersession_owner
make test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.source_mutations.task_and_decision_mutation_admission_and_replay_6d4ea5900a,module.tasksdecisions.frontend_unit.coordination_workflow_lifecycle_7a8d13c2f0
```

PASS: typecheck 2/2 `.cartulary/test-results/20260911T010439Z-p69318`;
Decision slice 5/5 `.cartulary/test-results/20260911T010644Z-p71097`;
admission/Task slice 3/3 `.cartulary/test-results/20260911T010644Z-p71095`.
The added reason characterization first failed at
`.cartulary/test-results/20260911T010506Z-p69922` (0/1), confirming the mismatch;
the bounded admission repair passes with existing request-hash goldens intact.
Earlier typecheck exposed an unsupported hypothetical Decision paste boundary
and unnormalized conflict-row typing; removed that unnecessary paste work and
used the conflict receipt's validated version boundary.

Deterministic evidence covers same-turn reservations, changed replacement
version after coordination, review/selection/closure/session fences, exact
replay after supersession, replay rejection preserving uncertainty, late
acknowledgement concealed during access loss, account retirement, and
acknowledgement preceding refresh-only recovery. Compatibility is unchanged
except correcting the existing Decision reason contract. Remaining risk is
actual projection reconciliation and browser/service behavior. DS-03 exits met.
Next: implement both-row monotonic reconciliation and validate real service flow.

## DS-04 log

Reconciliation now reads both affected Decisions across opaque pages, refuses
older rows, refreshes both histories, and refreshes the current Decision surface
through its continuity boundary. Reference composition consumes the refreshed
Decision rows without changing the general broker. Receipt high-water marks also
fence full query and live patch regression, and incoming Decision notifications
invalidate stale reviews before they can dispatch.

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.decision_supersession_runtime,web.workbook.regression.decision_supersession_reconciliation,web.workbook.regression.decision_supersession_review,web.workbook.regression.decision_supersession_owner,web.workbook.regression.decision_supersession_transport
make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.browser.the_browser_workbook_opens_task_requests_and_dec_8048a75ceb,module.tasksdecisions.browser.decision_supersession_exact_recovery
make service-backed-test-slice OWNER=module.tasksdecisions
```

PASS 6/6 `.cartulary/test-results/20260911T011940Z-p11104`;
PASS 11/11 `.cartulary/test-results/20260911T011940Z-p11130`;
PASS 15/15 `.cartulary/test-results/20260911T012531Z-p53219`.
The real browser scenarios attach `decision-recovery-evidence` containing exact
requests, receipts, both row projections, and both complete histories before and
after replay. Executed status is preserved; target/replacement each advance once;
committed-but-lost replay leaves histories and versions unchanged. The routed
store row independently counts one authoritative replacement-to-target link and
unchanged durable effects after replay, including change sets and revisions.
Existing Task and public coordination browser regressions pass.

The new accessibility scenario passed at
`.cartulary/test-results/20260911T012531Z-p53208`, including cancellation focus,
390px width, 200% zoom, text spacing, named controls, and recovery Escape return.
The accompanying visual row reached its capture intents but failed because new
scenario IDs did not follow the visual schema's 12-hex format. Corrected authored
IDs and regenerated projections; no schema/assertion relaxation. A preliminary
browser build at `.cartulary/test-results/20260911T011258Z-p76052` was invalidated
by concurrent formatting, so its evidence is not used. Type checks found only
new test fixture typing issues (nonempty affected-view tuple and DOM Node
removal), now corrected. `make lint-biome BIOME_CHECK_FLAGS=--max-diagnostics=200`
was rejected as an unsupported public input; use the ordinary Make summary.

Full ordinary visual baseline at `.cartulary/test-results/20260911T013346Z-p87935`
failed only the new Decision row: 213 capture intents, all 210 existing goldens
active/passing, zero orphans, zero ambiguous mappings, 26 registered fixtures,
and three missing new goldens. No existing visual behavior requires a refresh.
The accepted trigger is implementation of the adopted Decision History action
and its explicit review/recovery states. New nonregistry captures are owned by
`module.workbook.visual.decision_supersession_review_recovery`, scenario
`scenario_d87078727b34`: `decision-supersession-review-linux.png`,
`decision-supersession-review-narrow-linux.png`, and
`decision-supersession-accepted-linux.png`. Their capture profiles retain the
pinned renderer, dark graphite theme, default Decision density, 100% zoom,
viewport scope, existing UUID/time masks, and explicit review scroll placement.
Desktop is 1280×720; narrow review is 768×640. No old masks or tolerances change.

Broader affected regressions: web.workbook PASS 192/192 at
`.cartulary/test-results/20260911T013346Z-p87728`; module.revisions PASS 31/31 at
`.cartulary/test-results/20260911T013346Z-p87751`; platform.httpruntime PASS 1/1
at `.cartulary/test-results/20260911T013623Z-p80253`. These include retained
History/Timeline behavior. `make agent-finalize` passed at
`.cartulary/test-results/20260911T013308Z-p84192` and again after the final
History-version admission guard at `.cartulary/test-results/20260911T013736Z-p87920`.
`RESULTS_DIR` was unset; retained-run maintenance was skipped. Its first attempt
at `.cartulary/test-results/20260911T012901Z-p30755` stopped because catalog inputs
needed `make generate` before shape validation; regeneration resolved it.

Final code review added synchronous History high-water admission, a second
version check after coordinated writes, and reconciliation debt for a newer
history version than the loaded Decision row. It also preserves focus on local
submission feedback, restores the action trigger on explicit editor closure,
and clears history presentation callbacks on retirement.

`make browser-e2e-visual-update` PASS 12/12 at
`.cartulary/test-results/20260911T013950Z-p99817`. Inspected every added PNG:
desktop/narrow review retains readable participant identities and consequences,
visible confirmation/cancel controls and focus rings; accepted recovery presents
the executed target status independently from refresh completion. Exactly three
new PNGs and their Make-generated manifest entries changed; no existing golden
changed. No image edits or threshold changes were needed.

Latest focused runtime/reconciliation/transport/review PASS 5/5 at
`.cartulary/test-results/20260911T013950Z-p99615`; typecheck PASS 2/2 at
`.cartulary/test-results/20260911T013950Z-p99728`; import boundaries PASS 2/2 at
`.cartulary/test-results/20260911T014007Z-p30789`; Biome PASS 2/2 at
`.cartulary/test-results/20260911T014011Z-p31200`. Standalone Decision accessibility
PASS 11/11 at `.cartulary/test-results/20260911T014752Z-p70758`:

```bash
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.decision_supersession_review_recovery
```

Final review found that rejected attempts had correct detail but an inaccurate
shell summary ("in progress"). Correcting the Decision summary and adding an
assertion before final visual acceptance.

The summary assertion failed as expected at
`.cartulary/test-results/20260911T015134Z-p5791` (1/2), then passed after repair
at `.cartulary/test-results/20260911T015217Z-p10851` (2/2). The summary reports
pending, unknown, accepted, and rejected counts separately. The earlier ordinary
visual pass also passed 12/12 at `.cartulary/test-results/20260911T014712Z-p37956`;
two fresh passes are required after this last source correction.

`make agent-finalize` PASS 1/1 at
`.cartulary/test-results/20260911T015218Z-p11064`, with `RESULTS_DIR` unset.
Final-source typecheck, import boundaries, and Biome each PASS 2/2 at
`.cartulary/test-results/20260911T015248Z-p15112`,
`.cartulary/test-results/20260911T015248Z-p15120`, and
`.cartulary/test-results/20260911T015248Z-p15132`. The `package.ui` task guide
recommended its narrow slice; `make test-slice OWNER=package.ui` PASS 10/10 at
`.cartulary/test-results/20260911T015303Z-p44970`.

First final-source ordinary visual pass PASS 12/12 at
`.cartulary/test-results/20260911T015248Z-p15165`. The second run failed 10/12 at
`.cartulary/test-results/20260911T015648Z-p52870`, solely on the new narrow review
capture. Inspected actual/expected images: the target summary had scrolled above
the viewport. The scenario scrolled before responsive layout and text
normalization settled. Replaced its manual early scroll with the existing
visual-anchor helper, which positions and verifies the review after
normalization. This repairs capture geometry without changing goldens, masks,
tolerances, or product behavior. DS-04 remains open until two ordinary passes
succeed with the corrected capture.

The initial anchor adaptation failed its explicit scrollport assertion at
`.cartulary/test-results/20260911T020221Z-p95938` (9/11): the existing helper
defaults to the incident drawer. Added an optional authored scrollport selector
in `apps/web/e2e/support/visual/capture.ts`, retaining the drawer default and
all geometry assertions. Only the Decision scenario selects its existing
inspector container. No product layout or capture tolerance changed.

The inspector anchor passed geometry verification in the focused run at
`.cartulary/test-results/20260911T020454Z-p33213`; screenshot comparisons failed
9/11 because all three original captures had a two-pixel document scroll offset.
Inspected every actual image against its golden. Canonical document origin zero
and the declared inspector anchor retain all intended review/recovery content,
focus and density. This is a capture-profile correction, not a product change.
Regenerate these three new goldens through Make; existing goldens remain intact.

Corrected `make browser-e2e-visual-update` PASS 12/12 at
`.cartulary/test-results/20260911T020706Z-p68774`. Reinspected all three promoted
images; target/replacement identity, consequences, reason and confirmation remain
visible, and accepted recovery is accurate. Geometry now records the actual
inspector anchor and document origin zero. Only the three new Decision PNGs and
their generated manifest entries differ from baseline; no existing golden or
tolerance changed. `make agent-finalize` passed before this broader run at
`.cartulary/test-results/20260911T020635Z-p65264` with `RESULTS_DIR` unset.

First ordinary visual pass against corrected captures PASS 12/12 at
`.cartulary/test-results/20260911T021129Z-p4246`.

Second ordinary visual pass PASS 12/12 at
`.cartulary/test-results/20260911T021504Z-p39196`. Both fresh ordinary runs checked
all 213 captures against the reviewed final goldens with no source changes
between them. All DS-04 exits are met: deterministic state/admission/transport/
reconciliation, service-backed replay, Task/Timeline/History regressions,
accessibility and affected visual evidence pass. Compatibility remains the
existing public/stored contract with the bounded Decision reason repair. No
unresolved implementation or capture failure remains. The normal limits are
memory-only recovery and current server validation of authoritative facts.
Next: DS-05 maintenance, final gates and coordinated full check.

## DS-05 log

Final source is held stable. `make agent-finalize` PASS 1/1 at
`.cartulary/test-results/20260911T021903Z-p73949`. `RESULTS_DIR` was unset:
retained-run maintenance was skipped because qualifying exact-source full warm
evidence did not yet exist. No earlier waiver or baseline success substitutes
for the coordinated check.

| Exact command | Result | Evidence root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make frontend-typecheck` | PASS 2/2 | `20260911T021935Z-p77725` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260911T021935Z-p77755` |
| `make lint-biome` | PASS 2/2 | `20260911T021935Z-p77782` |
| `make json-shape-check` | PASS 3/3 | `20260911T021935Z-p77521` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260911T021935Z-p77514` |
| `make generate-drift` | PASS 4/4 | `20260911T021935Z-p77505` |
| `make service-backed-test-slice OWNER=module.tasksdecisions` | PASS 15/15 | `20260911T021935Z-p77621` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.decision_supersession_review_recovery` | PASS 11/11 | `20260911T021935Z-p77633` |

The final service run contains both real replay scenarios and retained Task
behavior. Its browser report at
`browser-e2e-webserver-backed/browser-groups/functional-support-default-sentinel/playwright-report.json`
embeds `decision-recovery-evidence` JSON attachments. Each full run root also
contains `run-summary.json`, row results and owner-routed unit logs.

Pre-completion documentation/scope audit: `make lint-markdown` PASS at
`.cartulary/test-results/20260911T022125Z-p80471`; `git diff --check` PASS.
`git branch --show-current`, `git rev-parse HEAD`, `git diff --cached --name-only`,
and the working-tree/inventory comparison confirm `main` at the original HEAD,
an empty index and exactly 60 inventoried changed paths. Only this new handoff
is Markdown; no public contracts, migrations, generated protocol/source-owner
SQL, digest or completed handoffs changed.

The first coordinated `make check` at
`.cartulary/test-results/20260911T022108Z-p47185` exposed two architecture policy
failures in new verification code. `sentinel.spec.ts` used a raw dynamic test-ID
selector; use its existing shared builder through `getByTestId`. Two new feature
tests embedded wire-format fixtures; move those unchanged fixtures into the
existing `decisionSupersessionTestSupport.ts` test-support owner. Preserve the
exact request/payload assertions and both architecture policies unchanged. The
`web.architecture` task guide recommends its narrow owner slice and import
boundaries. These repairs stay within the authorized tests/selectors seam.

The complete first check result was FAIL 832/834 units (1220/1222 rows), solely
those two new verification placement errors. Corrected the selector and moved
unchanged request/live-patch fixtures to test support; neither policy nor payload
assertion was weakened. No production or visual-capture behavior changed.

```bash
make test-slice OWNER=web.architecture
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.decision_supersession_owner,web.workbook.regression.decision_supersession_reconciliation
make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.browser.decision_supersession_exact_recovery
```

PASS 12/12 at `.cartulary/test-results/20260911T022856Z-p16034`;
PASS 3/3 at `.cartulary/test-results/20260911T022856Z-p16043`;
PASS 11/11 at `.cartulary/test-results/20260911T022856Z-p16059`.
The architecture verification collaborator is `web.architecture`; product
ownership remains as recorded above.

After the verification-only corrections, `make agent-finalize` PASS 1/1 at
`.cartulary/test-results/20260911T023002Z-p49808`, again with `RESULTS_DIR` unset
and retained-run maintenance skipped. Repeated final gates all pass:

| Exact command | Result | Evidence root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make frontend-typecheck` | PASS 2/2 | `20260911T023037Z-p53351` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260911T023037Z-p53359` |
| `make lint-biome` | PASS 2/2 | `20260911T023037Z-p53371` |
| `make json-shape-check` | PASS 3/3 | `20260911T023037Z-p53246` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260911T023037Z-p53241` |
| `make generate-drift` | PASS 4/4 | `20260911T023037Z-p53238` |

The second check at `.cartulary/test-results/20260911T023143Z-p58319` reached an
existing harness artifact-publication race before the protocol browser boundary
assertion: two producers entered `frontend-artifact.mjs` for the same private
`frontend-production` directory, and the second `renameSync` failed `ENOTEMPTY`.
The shell row and `target:build-web` both have `cache=none`; the shell consumer
has no graph prerequisite for that frontend artifact. The adopted production
bundle boundary itself passes independently:

```bash
make task-guide ROLE=module-author OWNER=package.protocol_ts
make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b
```

Guide PASS; narrow row PASS 1/1 at
`.cartulary/test-results/20260911T023530Z-p93991`. No test assertion, source,
scheduler capacity, or harness engine changed to obtain that evidence. Preserve
the publication failure as a harness limitation; an engine/graph redesign is
outside this seam. Repeat the unchanged full gate in a fresh private run rather
than waiving the row.

The second run finished FAIL 832/834. Its last unit also encountered an object
store capability-readiness timeout in the unchanged fixture test
`TestResetBucketPreservesNamespaceAndProvesMutation` (28 attempts; cleanup
failed), classified `service_readiness_timeout`, not a Decision assertion.
The first full run had passed that unit. Preserve both infrastructure failures
and verify the fixture through its routed owner before repeating the full gate.

The `harness.browser` task guide passed. Isolated fixture verification:

```bash
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_fixture_admission
```

PASS 3/3 at `.cartulary/test-results/20260911T024132Z-p2125`. Both failed
infrastructure paths now have independent passing evidence. The third full
`make check` uses unchanged source, default capacity and all ordinary assertions;
there is no waiver or scheduler override.

The third default-capacity check at
`.cartulary/test-results/20260911T024227Z-p19508` reproduced the same production
frontend publication collision before the bundle assertion. Repeating the gate
has established a shared-harness blocker, not passing full-check evidence.
DS-05 is BLOCKED pending separate authorization for a focused artifact ownership/
publication repair and concurrency regression. The user explicitly required a
stop for necessary material expansion beyond this Decision seam. No harness
engine, capacity policy, or failing assertion was changed. An authorization
question was presented while the current full run finished; no approval has been
received and no repair outside the authorized seam has been applied.

Final third-run result: `make check` FAIL 833/834 units at
`.cartulary/test-results/20260911T024227Z-p19508`. The only failed unit is
`package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b`;
the object-store fixture passed in this run. Its failed log is
`unit-logs/row-package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b/stderr.log`.
The repeated publication error occurs in unchanged
`tools/harness/readiness/frontend-artifact.mjs`, before the boundary assertion.

Decision implementation and DS-01–04 are complete. DS-05's binary exit is not
met: there is no passing coordinated full check. No implementation, service,
accessibility or visual failure remains in the Decision seam. A separate focused
harness repair must coordinate the production artifact producer and its bundle
consumer, preserve provenance/receipt validation, add concurrency regression
coverage, and rerun the unchanged full gate. No such next refactor is authorized
by this handoff.

Final post-tracker audit commands are `make lint-markdown`, `git diff --check`,
`git branch --show-current`, `git rev-parse HEAD`, `git diff --cached --stat`,
`git status --short`, and comparison of the complete working-tree path set to
the inventory below. These are repeated after this final tracker edit; the
pre-completion audits already passed as recorded above. The final response
reports their post-edit outcomes. No commit, reset, push or deployment occurred.
Next dependent action: obtain the separately requested harness-repair scope;
then resume DS-05 and require a passing full check before marking it DONE.

## Compatibility and rollback

Existing HTTP interfaces, schemas, stored data, dependencies and migrations
remain unchanged. Decision-only admission now follows the already adopted
`reason_note_v1` bound and character policy. Task lifecycle, Timeline lifecycle,
Decision approval, existing History/rollback and completed workbook seams retain
their owners and behavior. Supersession is not in REQ-01-104's destructive-lock
family; no client/server lock protocol or autosave FIFO ownership was added.

Recovery is explicit and memory-only: inspector closure preserves an admitted
attempt, while account/incident/runtime retirement discards recovery memory.
A browser reload cannot restore it. Access loss conceals protected contents;
same-account recovery within the same runtime can reveal retained attempts again.
Transport timeout does not imply cancellation. Replay keeps the exact request and
transaction identity; refresh-only recovery sends no mutation. The replacement
version is checked locally against loaded/accepted state, and is not an atomic
public precondition. Server validation remains final for concurrent changes,
visibility/deletion and authoritative link facts unavailable in the projection.
Paged or failed reads can leave explicit refresh debt. No automatic undo,
unsupersede route, re-key, rebase or replay was introduced.
Implementation rollback restores source/verification inputs without deleting
Decisions, links, history, revisions, receipts, or analyst data. Owner
contradiction or necessary material scope expansion blocks dependent work.

## Changed path inventory

- `apps/web/e2e/sentinel.spec.ts`
- `apps/web/e2e/support/visual/capture.ts`
- `apps/web/e2e/support/workbook/decisionSupersession.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/decision-supersession-accepted-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/decision-supersession-review-linux.png`
- `apps/web/e2e/workbook.visual.spec.ts-snapshots/decision-supersession-review-narrow-linux.png`
- `apps/web/src/testing/decisionSupersessionTestSupport.ts`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/adapters/createDecisionCandidateReader.ts`
- `apps/web/src/workbook/adapters/createWorkbookDecisionSupersessionAdapter.test.ts`
- `apps/web/src/workbook/adapters/createWorkbookDecisionSupersessionAdapter.ts`
- `apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts`
- `apps/web/src/workbook/features/coordination/CoordinationWorkflowBindings.tsx`
- `apps/web/src/workbook/features/coordination/DecisionSupersessionContext.ts`
- `apps/web/src/workbook/features/coordination/DecisionSupersessionEditor.test.tsx`
- `apps/web/src/workbook/features/coordination/DecisionSupersessionEditor.tsx`
- `apps/web/src/workbook/features/coordination/WorkbookDecisionSupersessionOwner.test.ts`
- `apps/web/src/workbook/features/coordination/WorkbookDecisionSupersessionOwner.ts`
- `apps/web/src/workbook/features/coordination/WorkbookDecisionSupersessionRecovery.tsx`
- `apps/web/src/workbook/features/coordination/decisionSupersession.characterization.test.tsx`
- `apps/web/src/workbook/features/coordination/decisionSupersessionModel.ts`
- `apps/web/src/workbook/features/coordination/decisionSupersessionOperation.ts`
- `apps/web/src/workbook/features/coordination/decisionSupersessionReconciliation.test.tsx`
- `apps/web/src/workbook/features/coordination/decisionSupersessionRuntime.test.ts`
- `apps/web/src/workbook/features/coordination/reconcileDecisionReceipt.ts`
- `apps/web/src/workbook/features/coordination/useCoordinationWorkflowController.test.tsx`
- `apps/web/src/workbook/features/coordination/useCoordinationWorkflowController.ts`
- `apps/web/src/workbook/features/coordination/useDecisionCandidates.ts`
- `apps/web/src/workbook/features/generic/GenericWorkbookInspector.tsx`
- `apps/web/src/workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`
- `apps/web/src/workbook/features/generic/createGenericMutationCommandPort.ts`
- `apps/web/src/workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`
- `apps/web/src/workbook/history/WorkbookRecordHistoryOwner.ts`
- `apps/web/src/workbook/hooks/useOwnerReferenceOptions.ts`
- `apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts`
- `apps/web/src/workbook/hooks/useWorkbookSurfaceQueries.ts`
- `apps/web/src/workbook/inspector/WorkbookInspectorContextualActions.tsx`
- `apps/web/src/workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx`
- `apps/web/src/workbook/inspector/inspectorCapabilityResolver.test.ts`
- `apps/web/src/workbook/inspector/inspectorCapabilityResolver.ts`
- `apps/web/src/workbook/inspector/presentation/WorkbookInspectorActions.tsx`
- `apps/web/src/workbook/inspector/useWorkbookRecordHistoryController.ts`
- `apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.test.ts`
- `apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.ts`
- `apps/web/src/workbook/mutations/workbookMutationCommandPorts.ts`
- `apps/web/src/workbook/query/useGenericSurfaceQuery.ts`
- `apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts`
- `docs/handoffs/workbook-decision-supersession-refactor-handoff.md`
- `internal/modules/tasksdecisions/admission_supersession.go`
- `internal/modules/tasksdecisions/mutation_admission_test.go`
- `packages/ui-contracts/src/index.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/test_families/module.tasksdecisions.json`
- `tools/test_families/module.workbook.json`
- `tools/test_families/web.workbook.json`
