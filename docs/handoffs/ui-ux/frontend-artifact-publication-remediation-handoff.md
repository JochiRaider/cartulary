# Frontend artifact publication remediation

Delivery status: FP-01–FP-05 complete. Two consecutive ordinary full checks
passed 834/834 units and all 1222 selected rows on the same final source.
The separately blocked Decision DS-05 is reconciled in its existing handoff.

## Baseline and authorization

The user authorized FP-01–FP-05: repair shared frontend artifact ownership,
publication, and diagnostics, then complete Decision DS-05 verification. No
commit, index rewrite, reset, push, deployment, dependency or data migration,
visual refresh, generic artifact engine, or unrelated refactor is authorized.

Actual baseline: clean `main`, HEAD
`c36eef239a7e93e490dda5cea670eb3a5f6035fc`. The user committed the previously
staged Decision work before implementation; it is preserved in this baseline.
The index and worktree were both clean. Root `AGENTS.md` is the only applicable
repository instruction file. Historical handoffs are evidence, not instructions.

`make doctor` PASS at `.cartulary/test-results/20260911T121837Z-p27237`.
Pins: Go language 1.27.0/toolchain 1.27.1, Node 24.15.0, pnpm 10.33.0,
ShellCheck 0.11.0. All four `make task-guide ROLE=module-author OWNER=...`
commands passed for `harness.command_surface`, `harness.browser`,
`harness.generated_artifacts`, and `package.protocol_ts`.

Retained failure: `.cartulary/test-results/20260911T024227Z-p19508`, 833/834.
Its unit events show `target:build-web` starting at 2040 ms and the protocol
bundle-security shell row starting at 2074 ms with no dependencies. Its nested
Make build failed publishing `frontend-production` with `ENOTEMPTY`, before the
security assertion. The shell adapter incorrectly reported a product assertion.
The separate object-store readiness timeout passed subsequent evidence and is
not attributed to this repair.

## Owners, boundaries, and consumer inventory

Testing Harness REQ-060/801/802/807 own producer lifetime, graph equivalence,
admission, and deduplication; Section 9 owns normalized failure attribution.
`harness.command_surface` owns graph and Make invocation mechanics;
`harness.browser` collaborates on frontend/runtime lifetime;
`harness.generated_artifacts` owns derived projections; `package.protocol_ts`
owns the bundle-security assertion. No product owner contradiction was found.

Permitted implementation paths: `Makefile`; authored task/graph/topology and
schema/verification inputs under `tools/` and `contracts/verification/`;
necessary harness compiler, Make renderer, readiness, runtime, execution,
security analyzer, and focused tests; Make-generated derivatives; the harness
NLSpec, this handoff, and the final Decision handoff reconciliation.

Production consumers: browser stages selecting production; bundle-security
analysis; embedded-web-assets, then server/server-harness/operator compilation.
Measurement consumers: measurement/visual stages selecting measurement.
Conventional dist outputs remain packaging projections, not scheduled readers'
artifact authority. Existing browser receipt references and frontend receipt v1
are preserved. Tests and runtime must not depend on Markdown.

## Gap ledger

| Gap | Remediation / areas | Rationale | Long-term benefit | Compatibility | Unresolved risk | Binary completion |
| --- | --- | --- | --- | --- | --- | --- |
| G1 | Extend normative frontend contract to all consumers; spec, projections, tests, docs | Browser-only detail obscures existing shared ownership | One lifecycle for future consumers | Receipt/public commands unchanged | Another consumer bypasses admission | Every consumer follows declared producer and readiness |
| G2 | Shared dependency expansion; graph owner, compiler, tests | Aggregate-only edges leave other selectors inconsistent | One declaration and semantic identity | Graph v5 retained; dependency digests change | Hidden work and concurrent builds | Equivalent selections and one producer pass |
| G3 | Graph entry and resolve-only consumers; Make renderer, authored inputs, tests | Edges alone cannot prevent nested production | Visible bounded production | Missing artifacts fail instead of rebuilding | Correct graph still hides work | Consumers build zero times; producer once |
| G4 | Exclusive claim, sealed payload, receipt-last readiness; runtime, security tests | Check-then-act permits duplicate publication | Deterministic crash/cancel behavior | No same-run producer reuse; no data impact | Partial readiness or overwritten artifacts | Concurrency, partial states, and isolation tests pass |
| G5 | Private scoped typed failure envelope; spec/schema, adapters, tests | Make status loses preparation cause | Reliable causal evidence, one result writer | Result identities/fields retained; row enum aligned to existing taxonomy | Harness failure misreported as product | Classified failures survive all canonical layers |
| G6 | Deterministic and service/browser tests, routing, handoff | Isolated success does not prove shared-run behavior | Reproducible regression evidence | Assertions and visuals preserved | Lucky full run conceals defect | Focused matrix and two ordinary full checks pass |

## Tracker

| Phase | Status | Exit criteria / next action |
| --- | --- | --- |
| FP-01 | DONE | Baseline, owner decisions, bounded scope, and normative clarification recorded |
| FP-02 | DONE | Graph/Make characterization and repaired selection parity pass |
| FP-03 | DONE | Exclusive admission, crash boundaries, scoped diagnostics, and bounded process termination pass |
| FP-04 | DONE | Combined service/browser consumers and focused matrices pass |
| FP-05 | DONE | Final gates and two consecutive full checks pass; Decision completion and scope are reconciled |

## FP-01 log

Read root instructions, adopted harness execution/artifact/failure sections,
current compiler/Make/rendering/artifact/runner boundaries, receipt schema,
authored graph/task inputs, and retained failing events/results. Commands:
`git status --short`, `git branch --show-current`, `git rev-parse HEAD`,
`git log -1 --format=fuller`, four task guides, and `make doctor`; all passed.
Changed `docs/testing-harness-nlspec.md` first to close shared consumer admission,
exclusive claims, receipt-last readiness, and private command failure semantics.
Created this handoff. Public interfaces and application state remain unchanged.
Risk: compiler unification can expose inconsistent selection paths; characterize
them before repair. FP-01 exit satisfied; next action is FP-02 characterization.

## Compatibility and rollback

Rollback only this repair's authored sources and generated derivatives, while
preserving the Decision baseline, analyst data, Decisions, links, history,
revisions, receipts, and retained evidence. No action here deletes those data.
No full-suite serialization, capacity override, automatic retry, assertion
weakening, or visual golden refresh is part of the repair.

## FP-02 log

Characterization `make test-slice OWNER=harness.command_surface` failed as
expected at `.cartulary/test-results/20260911T122126Z-p28973`: the bundle-security
row had `needs=[]` instead of `target:build-web`. An earlier test-wiring run
at `20260911T122058Z-p28393` failed importing the not-yet-created test module;
it is not characterization evidence. A Make fixture initially omitted the real
Makefile's secondary expansion (`20260911T122431Z-p34490`); corrected the fixture,
without weakening its assertions.

Implemented shared `compileDependencies` for rows/policies and aggregate union;
authored the protocol consumer edge; made standalone frontend/binary/embedded
targets graph entries; suppressed transitive producer Make prerequisites in
graph children. Added synthetic profile consumer, cycle/missing producer,
selection parity, and instrumented generated/real Make-header tests.

`make generate` PASS at `20260911T122257Z-p29742` regenerated task-surface and
topology derivatives. `make test-slice OWNER=harness.command_surface` PASS 1/1
at `20260911T122444Z-p34999`, including the complete graph/Make matrix.
`make test-slice OWNER=package.protocol_ts
ROWS=package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b`
PASS 3/3 at `20260911T122319Z-p33227` (now includes both producer prerequisites).
All evidence roots in this log are beneath `.cartulary/test-results/`.

Changed paths: `Makefile`, authored task/graph owners, graph compiler,
`test-frontend-producer-graph.mjs`, command-surface suite entry, and generated
task/topology projections. Graph v5, public command identities, receipt v1,
security assertions, and stored data remain unchanged. Risks remaining: private
producer admission and diagnostic attribution are FP-03. FP-02 exit satisfied;
next action is lifecycle/failure characterization and implementation.

## FP-03 log

Duplicate-publication characterization failed as expected at
`20260911T122642Z-p35928`: the old rename error lacked the required normalized
producer-violation classification. Implemented exclusive current-run/profile
claims before compiler work, removed completed-directory producer reuse, sealed
immutable payloads with receipt-last readiness, rejected existing destinations,
and strengthened receipt ownership/mode/link validation. Current browser receipt
identity, fields, and retained paths remain unchanged.

Added private `cartulary.harness_command_failure.v1`, its schema attachment,
runtime channel, producer/consumer reporting, and executor/shell adaptation.
Canonical row/unit writers remain unchanged. Invocation identities and paths
are private; no logs or Make numeric status determine a classified envelope.
Timeout/cancellation retain parent precedence and terminate unresponsive owned
process groups within the existing boundary (two-second SIGKILL escalation).

The real shell-row test uncovered a demonstrated projection mismatch: row-result
v2 accepted only the Go setup subset of Section 9's failure classes, rejecting
`artifact` before writing evidence. Corrected that enum to the existing complete
taxonomy, with no schema identity, field, or new failure-class change. This is
the bounded additive projection correction needed for G5; earlier planning's
assumption that every result schema already admitted those values was wrong.
Existing consumers and historical valid v2 rows remain valid.

Tests cover two IPC-coordinated producer processes (one admitted, one rejected),
independent profiles/runs, unchanged receipt after a duplicate, killed producer,
reconstructed sealed-payload/absent-receipt crash state, compiler failure,
malformed/conflicting/mis-scoped/unsafe diagnostic envelopes, success/failure
inconsistency, actual Make exit-code loss, canonical shell-row attribution,
timeouts and cancellation. Existing artifact/security assertions remain intact.

`make generate` PASS `20260911T123359Z-p38433`.
`make test-slice OWNER=harness.command_surface` PASS 1/1 with the full matrix at
`20260911T123850Z-p46015`. Earlier wiring tests at `123645Z-p43858`,
`123727Z-p44599`, and `123800Z-p45270` exposed the row enum mismatch; the final
assertion retains proof that a canonical artifact-failure row was written.
`make lint-scripts` PASS 2/2 `20260911T123414Z-p41604`;
`make lint-shell` PASS 4/4 `20260911T123417Z-p42137` (before final test additions).
All roots are under `.cartulary/test-results/`.

Changed areas: frontend artifact owner, private runtime diagnostic channel,
executor, shell/row adapters, protocol artifact admission wrapper, schemas and
attachments, focused fixture/test modules, and task backing-input declarations.
Remaining risk is integrated service/browser admission and final generated
closure. FP-03 exit satisfied; next action is FP-04 combined-consumer evidence.

## FP-04 log

Added the internal Make-owned `frontend-artifact-consumer-check` fixed-row graph
binding. It selects the existing protocol boundary and browser artifact-lifetime
rows, adds no product row or aggregate policy, and takes no caller row override.
This is the narrow combined-consumer verification required by the plan; it
uses the existing compiler, scheduler, schemas, and lifecycle. Extended authored
Make recipe validation/rendering for exact nonempty sorted unique active row
selection, with invalid/unknown-row and graph-equivalence tests. The lifetime
row remains full-tier; aggregate tiers and default fast smoke are unchanged.

`make generate` PASS `20260911T124021Z-p46909`; generated task-surface and topology
projections only. `make explain-target TARGET=frontend-artifact-consumer-check
DETAIL=summary` confirmed 12 units/2 rows, owners `harness.browser` and
`package.protocol_ts`, and declared browser/Postgres/object-store services.

`make frontend-artifact-consumer-check` PASS 12/12 at
`20260911T124145Z-p51027`. Exactly one `target:build-web` started at 241 ms and
completed at 2418 ms; the security row started at 2422 ms with
`needs=[target:build-web]`. The real browser retained loaded lazy asset hashes
through independently published production and measurement builds and repeated
reloads. `make service-backed-test-slice OWNER=harness.browser
ROWS=harness.browser.boundary_support.frontend_artifact_lifetime` PASS 11/11 at
`20260911T124244Z-p83421`, proving the separate owner-selection entry path.

Final artifact review added strict duplicate-member receipt parsing, moved all
owned-runtime work beneath cleanup, and expanded every receipt-identity/digest
negative. The combined service scenario passed again, 12/12 at
`20260911T124458Z-p17188`. The focused command-surface matrix passed at
`20260911T124426Z-p16406` before a final executor review exposed the cooperative
zero-exit timeout edge; that now retains the parent timeout classification and
normalized exit, with a dedicated assertion. This does not change ordinary
service/browser success behavior. No visual golden or product source changed.

All evidence roots are beneath `.cartulary/test-results/`. Compatibility:
fixed-row verification is additive and internal; public command identities and
artifact/result field layouts remain intact. The documented row failure-enum
projection correction remains the only result-contract adjustment. Remaining
risk is coordinated final verification. Next action is FP-05 generation,
finalizer, explicit gates, and two ordinary full checks.

## FP-05 log

Final focused matrix: `make test-slice OWNER=harness.command_surface` PASS 1/1
at `.cartulary/test-results/20260911T124619Z-p49497`, including unresponsive and
cooperative timeout, cancellation, and canonical shell failure evidence.
`make generate` PASS at `20260911T124714Z-p50250`.
`make agent-finalize` PASS 1/1 at `20260911T124751Z-p53259`, before broader
verification. `RESULTS_DIR` was unset: retained-run maintenance was skipped;
no failed or mismatched-source full run was reused. Its canonical graph-mode
finalizer report is `unit-artifacts/finalize-summary.json`.

| Preliminary broad command | Result | Evidence root below `.cartulary/test-results/` |
| --- | --- | --- |
| `make frontend-typecheck` | PASS 2/2 | `20260911T124831Z-p57403` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260911T124831Z-p57452` |
| `make lint-biome` | PASS 2/2 | `20260911T124831Z-p57533` |
| `make lint-scripts` | PASS 2/2 | `20260911T124831Z-p57571` |
| `make lint-shell` | PASS 4/4 | `20260911T124831Z-p57668` |
| `make json-shape-check` | PASS 3/3 | `20260911T124831Z-p57057` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260911T124831Z-p57065` |
| `make generate-drift` | PASS 4/4 | `20260911T124831Z-p57098` |
| `make harness-contract` | PASS 2/2 | `20260911T124831Z-p57795` |

Two ordinary full `make check` runs passed 834/834 units at
`20260911T124901Z-p64376` and `20260911T130004Z-p19948`, with matching executable
source, graph, and toolchain identities. Each selected row also passed. These
runs precede the final test strengthening described below and do not substitute
for the required final-source pair.

The final evidence audit identified a remaining test detail: publication crash
states were reconstructed rather than interrupted at both actual filesystem
boundaries. FP-05 remains open while controlled process barriers strengthen that
coverage. Any full evidence preceding the verification-source change remains
historical; two ordinary checks must pass again on the final coordinated source.

Added simultaneous producer release with exactly one build-entry witness and
actual SIGKILL barriers immediately before receipt rename and during conventional
output replacement. The tests prove unavailable receipt-less payloads, no
producer success after interruption, no same-run takeover, and continued access
to an independent run after cleanup. No production source or injection hook was
added. The initial focused run at `20260911T130539Z-p9288` and direct diagnostic
run at `20260911T130619Z-p10342` rejected a fixture path passed as one unsafe
component; corrected the fixture to use the runtime's existing component API.
`make test-slice OWNER=harness.command_surface` PASS 1/1 at
`20260911T130657Z-p10891`. All prior assertions remain intact.

Final-source generation: `make generate` PASS `20260911T130731Z-p11578`.
Combined service/browser validation: `make frontend-artifact-consumer-check`
PASS 12/12 `20260911T130744Z-p14626`. `make agent-finalize` PASS 1/1
`20260911T130839Z-p46844`, before broader verification; generated outputs were
unchanged (zero updated files). `RESULTS_DIR` remained unset because the previous
full runs preceded these verification inputs, so retained-run maintenance was
skipped again. No retained failed or mismatched-source evidence qualified.

| Final-source command | Result | Evidence root below `.cartulary/test-results/` |
| --- | --- | --- |
| `make lint-scripts` | PASS 2/2 | `20260911T130902Z-p51007` |
| `make lint-shell` | PASS 4/4 | `20260911T130902Z-p51067` |
| `make lint-biome` | PASS 2/2 | `20260911T130902Z-p51009` |
| `make frontend-typecheck` | PASS 2/2 | `20260911T130902Z-p50945` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260911T130902Z-p51040` |
| `make json-shape-check` | PASS 3/3 | `20260911T130902Z-p50573` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260911T130902Z-p50523` |
| `make generate-drift` | PASS 4/4 | `20260911T130902Z-p50554` |
| `make harness-contract` | PASS 2/2 | `20260911T130902Z-p51277` |

The next ordinary full check, `20260911T130931Z-p57827`, finished FAIL 833/834
and reported an unchanged
Network Analysis test failure in `web.networkflow.regression.exploration_focus`.
Its Vitest result stopped at 15148.77 ms with `STACK_TRACE_ERROR` at the test
declaration, consistent with the existing 15000 ms limit in
`apps/web/vite.config.ts`; it provided no more specific failed assertion.
This is a timing inference, not a demonstrated frontend-publication defect.
The `web.networkflow` task guide passed, followed by
`make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.exploration_focus`
PASS 2/2 at `20260911T131515Z-p9047`. No application code, assertion, timeout,
capacity, or cache policy was changed. This failed full run cannot qualify;
two consecutive ordinary successful checks are still required.

The next ordinary `make check` PASS 834/834 at `20260911T131633Z-p12992`;
all 1222 selected rows passed. Normal cache mode reused 665 eligible units and
executed the 169 bypass units. The production artifact producer remains
uncached: it started once at 2116 ms and completed at 15338 ms. Embedded assets
started at 15340 ms and bundle security at 15346 ms, each with the declared
`target:build-web` dependency. The second consecutive run uses unchanged source.

Pre-completion documentation audit: `make lint-markdown` PASS at
`20260911T131740Z-p95213` (`adhoc/lint-markdown/tool-run-summary.json`).
`git diff --check`, `git diff --cached --check`, and `git diff --cached --quiet`
all passed. Branch remains `main` and HEAD remains
`c36eef239a7e93e490dda5cea670eb3a5f6035fc`. The changed paths are restricted to
the inventory below; the index is empty and unchanged. These audits will be
repeated after the final handoff edit.

## FP-05 completion

Required consecutive ordinary `make check` evidence, without executable edits
between runs or caller capacity/cache overrides:

| Run root below `.cartulary/test-results/` | Units | Selected rows | Duration |
| --- | --- | --- | --- |
| `20260911T131633Z-p12992` | PASS 834/834 | PASS 1222/1222 | 273640 ms |
| `20260911T132122Z-p3018` | PASS 834/834 | PASS 1222/1222 | 264392 ms |

Both manifests have `cache_mode=normal`, empty declared inputs, 665 eligible
cache hits and 169 bypass units, with no failed, skipped, or cancelled units.
Their common exact executable identity is:

- Source: `sha256:ecffc356e71a3c08c4848b14233307338d6dd2a4efb9e8cbc7cf4ff32053b121`
- Graph: `sha256:f97fb2dd13813f7b462fd415583aba690c799f04447cc71af35ea4396e30e671`
- Toolchain: `sha256:b9e2facb1e05340bde25cd0d93fa0fb207518a67f1f5ca371dab848c49b012a8`
- HEAD: `c36eef239a7e93e490dda5cea670eb3a5f6035fc`

Audited canonical manifests, summaries, all unit/row records, unit events, and
each retained `build-web/frontend-artifact.json`. Each run selected production
and executed exactly one uncached producer. In the second run it started at
2107 ms and completed at 15035 ms; embedded assets started at 15039 ms and bundle
security at 15049 ms, both with the declared producer dependency. Receipt run,
profile, source, and toolchain identities match their manifests. Both fresh
builds produced content digest
`sha256:93dd88171c2d439b5621a3a501829adb0d2bd3b9e22017d714dd798b1a3843a6`.
Separate production/measurement concurrency and independent publication remain
covered by the focused process matrix and actual browser lifetime scenario.

All G1–G6 exits are met. The original publication failure and the later unrelated
Network Analysis timing observation remain retained diagnostic evidence; neither
is substituted for successful full evidence or waived. No owner contradiction
remains. The row enum alignment is the only existing result-contract adjustment;
the new envelope is private and canonical results remain parent-owned.

Updated [Decision DS-05](workbook-decision-supersession-refactor-handoff.md)
with this successful coordinated evidence, preserving its original blocked-run
history and application implementation. Final post-tracker commands are
`make lint-markdown`, `git diff --check`, `git diff --cached --check`,
`git diff --cached --quiet`, `git branch --show-current`, `git rev-parse HEAD`,
and `git status --short`. The recorded pre-completion audits passed; these are
repeated after this final documentation edit before delivery. No next refactor
is authorized or required to complete this remediation.

## Changed path inventory

Specification and handoffs:

- `docs/testing-harness-nlspec.md`
- `docs/handoffs/frontend-artifact-publication-remediation-handoff.md` (new)
- `docs/handoffs/workbook-decision-supersession-refactor-handoff.md` (completion only)

Authored composition, contracts, and runtime:

- `Makefile`
- `tools/task_surface_owner.json`
- `tools/harness_work_graph_owner.json`
- `tools/harness_schema_attachments.json`
- `tools/schemas/cartulary.harness_command_failure.v1.schema.json` (new)
- `tools/schemas/cartulary.harness_row_result.v2.schema.json`
- `tools/harness/generated-artifacts/task-surface/make-renderer.mjs`
- `tools/harness/generated-artifacts/task-surface/recipe-validation.mjs`
- `tools/harness/scheduler/work-graph/compiler.mjs`
- `tools/harness/scheduler/work-graph/executor.mjs`
- `tools/harness/readiness/frontend-artifact.mjs`
- `tools/harness/runtime/command-failure.mjs` (new)
- `tools/harness/execution/runners/row-runner-cli.mjs`
- `tools/harness/execution/runners/shell.mjs`
- `tools/harness/static-analysis/protocol-ts-browser-artifact-reachability.mjs`

Focused verification:

- `tools/harness/tests/test-harness-command-surface-contracts.mjs`
- `tools/harness/tests/test-frontend-producer-graph.mjs` (new)
- `tools/harness/tests/test-frontend-producer-lifecycle.mjs` (new)
- `tools/harness/tests/frontend-producer-fixture.mjs` (new)
- `tools/harness/tests/test-command-failure.mjs` (new)

Make-generated changes, reported separately from authored work:

- `tools/task_surface.generated.mk`
- `tools/task_surface_manifest.json`
- `tools/execution_topology_render_index.json`

## Operational limits

Private artifacts remain available only during their owning suite runtime.
An interrupted producer cannot be retried or taken over in that run; a new run
builds independently. Conventional output publication retains its brief lock;
it does not become a shared cache or reader authority. Interrupted conventional
publication can leave an incomplete conventional output, but cannot invalidate
another run's private artifact or count as successful producer completion.

Typed diagnostics are supplied by the affected frontend producer/consumer paths;
unrelated shell commands retain their existing fallback. Consumers validating
row-result v2 against an older four-class projection need the corrected enum;
existing valid records and schema identities remain compatible. The independent
object-store timeout observation is retained, not claimed repaired. Application
code, APIs, stored data, Decision semantics, Task/Timeline behavior, dependencies,
migrations, and reviewed visuals are unchanged by this harness remediation.
