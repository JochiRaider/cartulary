# Check harness and Network Flow remediation handoff

## Baseline and authority

Entry: clean `main` at `1d25114774e91d73febae44069669451c774178e`.
Root `AGENTS.md` is the only applicable instruction file. Existing planning
reads and focused reruns used this exact product source. Toolchain pins specify
Go launcher 1.27.1 (module Go 1.27.0), Node 24.15.0 and pnpm 10.33.0.
No commit, reset, push, deployment, dependency or migration is authorized.

Testing Harness REQ-013 owns catalog closure and prohibits separately authored
repository totals; REQ-020 leaves private implementation structure unspecified.
REQ-003/004 own Make invocation and generated projections. REQ-017/804 own
fixture classification, service dependencies and isolation. Network Flow §13.1
owns initial/continuation request variants; §17.8 owns paging metadata.
Current owners require no behavioral amendment. Markdown remains human evidence
and is never a test, runtime or generation input. Completed history handoffs and
the history implementation are preserved.

Permitted changes: private harness test support, focused test/selector support,
frontend ownership input, minimal saved-result identity metadata, removal of the
unused Network Flow hash helper, necessary authored verification inputs and
Make-produced derivatives, and this dedicated handoff.

## Tracker

Only the current row is IN_PROGRESS; close it before advancing.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| RM-01 Authority and characterization | DONE | Baseline, owners and seven failure signatures recorded. |
| RM-02 Harness structure | DONE | Three harness checks pass; closure and import isolation covered. |
| RM-03 Frontend verification integrity | DONE | Policy, integration, selector and recovery rows pass. |
| RM-04 Backend and integrated validation | DONE | Staticcheck, receipt integrity and four real browser rows pass. |
| RM-05 Final verification and handoff | DONE | Full check 820/820; evidence, compatibility and final audit complete. |

## Gap ledger

| Gap / failure | Owner and remedy | Compatibility / unresolved risk | Binary completion |
| --- | --- | --- | --- |
| G1: three harness rows expect 258 dedicated rows, actual 262 | Harness catalog; reconcile live inventory and prove counts on synthetic fixtures | Test-only; fixed mirrors break ordinary growth | Valid growth passes, invalid closure fails |
| G2: shared import aborts unrelated suites | Private harness test support; lazy per-case context and named assertions | Public command/result contracts stable; masked failures otherwise persist | Import has no repository I/O; independent cases execute |
| G3: six browser selector violations | UI selectors/browser support; shared family builders and identity-based readiness | Existing IDs and visual presentation stable; stale/wrong graph risk | Policy and affected browser rows pass |
| G4: 24 unowned TypeScript paths | Authored frontend source ownership; 22 web.network_flow and two web.services entries | Metadata only; hidden ownership gaps | Exact-once coverage and explicit negative diagnostics |
| G5: contributor mock limit 50 versus requested 500 | Network Flow test fixture; capture context, valid paging, readiness before action | Decoder and public wire unchanged; invalid fixtures mask behavior | Next/Previous exact bodies pass, wrong limits still reject |
| G6: unused networkFlowRequestHash | Network Flow private backend; remove unused helper/import | Active hashes and receipts unchanged; redundant unsafe alternative | Staticcheck and receipt integrity pass |

## Retained characterization

Original `make check`: FAIL 813/820 units, seven failures, at
`.cartulary/test-results/20260910T192529Z-p62009`. Its earlier source was
`e675950eed39620de84439e82b0ec405dab0428d` plus the now-committed history changes.
Both that HEAD and current authored catalogs have 262 dedicated rows; browser
rows use browser_stack and do not contribute. The count mismatch predates the
history browsing changes.

Exact current-source focused reproductions from planning:

- `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.selectorcontractpolicy_keeps_cross_boundary_sele_61868c7039,web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`:
  FAIL 1/3 at `20260910T194014Z-p20403`.
- `make test-slice OWNER=module.networkflow ROWS=module.networkflow.frontend_integration.verify_production_network_flow_grids_read_only_b_f662be335e`:
  FAIL 1/2 at `20260910T194014Z-p20410`; contributor continuation never admitted
  because generated decoding rejects the fixture's limit mismatch.
- `make harness-command-surface-contract`: FAIL at `20260910T194035Z-p22015`,
  shared import count assertion. The same import accounts for evidence and
  catalog-suite failures in the full run.
- `make lint-go-staticcheck`: FAIL at `20260910T194035Z-p22076`,
  `internal/modules/networkflow/routes.go:751`, unused `networkFlowRequestHash`.

Run IDs resolve under `.cartulary/test-results/`; graph runs retain
`run-summary.json`, row results and unit logs; leaf wrappers retain the target's
`tool-run-summary.json`. Task guides for harness.test_catalog, web.architecture
and module.networkflow passed during planning and again on implementation entry.

## Compatibility and rollback

HTTP schemas, stored data, receipt semantics, fixture isolation and public Make
identities remain unchanged. Rollback reverts coordinated private support,
implementation, tests, selectors and authored/generated verification artifacts.
It deletes no history, revisions, receipts or analyst data. No subsequent
refactor is implied.

## RM-01 exit

Entry audit confirms clean main/HEAD as above, one applicable AGENTS.md and no
user changes. All three requested current task guides passed again. Existing
exact-source characterization is retained above; no failed checks are called
passing. Only this handoff changed in RM-01. Owner review confirms that changing
258 to 262 would preserve the catalog-total violation; the fix must remove the
live mirror. Next: G1/G2, with import isolation and retained semantic assertions.

## RM-02 exit

Removed fixed live catalog/target/profile totals and exponential live subset
coverage. Named cases reconcile owner inventories, exercise synthetic growth and
invalid isolation approvals, and verify one builder per resolved fixture identity.
Shared imports define helpers/cases only. Every case receives a lazy private
context; policy fixtures are owned and cleaned by their tests. Child-process
coverage proves import purity and independent execution after policy failure.
Existing case IDs, Make identities and retained result contracts are preserved.

Changed: contract-suite-support.mjs and three private sibling modules
(contract-test-context.mjs, contract-catalog-cases.mjs,
contract-initialization-cases.mjs); tools/task_surface_owner.json declares their
backing inputs. `make generate` passed at 20260910T195348Z-p27656 and updated
only task_surface_manifest.json and execution_topology_render_index.json.

Verification: command-surface and evidence checks initially passed at
20260910T195203Z-p26855/p26856. Full harness attempts at
20260910T195348Z-p28079 and 20260910T195505Z-p31721 exposed test implementation
issues: inherited Node test reporting, undefined command IDs on private targets,
and explicit TAP selection. These were corrected. The first attempt also saw a
cache-closure fallback while generation was running concurrently; subsequent
serial runs passed that case without a cache implementation change. Generation
and validation are now sequenced. Full harness passed at
20260910T195558Z-p32558. Final combined command
`make harness-command-surface-contract harness-evidence-contract harness-contract-tests`
passed after the final import-probe adjustment at 20260910T195701Z-p33433,
20260910T195701Z-p33577 and 20260910T195705Z-p33767 respectively.

Exit: all three original harness failures are resolved, and invalid policy
fixtures remain rejected in named tests. No production, fixture classification,
schema or public command change. Remaining risk: frontend and backend failures
await their ordered workstreams. Next: G3-G5.

## RM-03 exit

Added escaped edge/vertex family selectors in ui-contracts and consumed them in
bounded browser DOM observation. Saved-result containers expose their admitted
graph and projection IDs; browser readiness now checks selected graph identity
and both result identities. Display names remain content assertions, including
duplicate names. No visible presentation change is intended.

Added the 24 explicitly assigned ownership paths. The private ownership test now
reports missing, stale, duplicate and unsafe paths individually, with synthetic
negative cases (including sortedness and duplicate owners). No directory-based
ownership inference was added. The authored architecture row includes the new
case; generation updates its projections.

Contributor fixtures capture request limit, digest and selector and bind them to
opaque fixture cursors. A controlled delayed response proves the grid shell is
insufficient readiness; Next/Previous wait for accepted pages and enabled
controls, preserving exact two-field continuation requests. The separate page
recovery negative test now includes mismatched limits. Production decoding is
unchanged.

Changed: apps/web/e2e/network-flow-navigation.spec.ts and network-flow.spec.ts;
apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx,
NetworkFlowSavedGraphPanel.tsx and networkFlowPageRecovery.test.tsx;
apps/web/src/testing/sourceOwnershipPolicy.test.ts; packages/ui-contracts/src/
index.ts, networkFlowSelectors.ts and network-flow-selectors.test.ts;
tools/frontend_source_ownership.json and tools/test_families/web.architecture.json;
Make-generated verification artifacts listed in the final inventory.

Commands/results:

- `make format`: PASS, 2/2, 20260910T200345Z-p36018.
- `make generate`: PASS, 20260910T200404Z-p40327.
- `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.selectorcontractpolicy_keeps_cross_boundary_sele_61868c7039,web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`:
  PASS, 3/3, 20260910T200433Z-p43465.
- `make test-slice OWNER=module.networkflow ROWS=module.networkflow.frontend_integration.verify_production_network_flow_grids_read_only_b_f662be335e`:
  PASS, 2/2, 20260910T200433Z-p43472.
- `make test-slice OWNER=package.ui ROWS=package.ui.frontend_unit.network_flow_selector_contracts_b95d925bdf`:
  PASS, 2/2, 20260910T200433Z-p43485.
- `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.page_recovery_boundaries`:
  PASS, 2/2, 20260910T200433Z-p43500.

Exit: original frontend failures resolved and negative checks retained. Only
additive DOM metadata affects production; HTTP and paging behavior are stable.
Remaining risk: real browser timing and backend lint await RM-04. Next: G6 and
serial browser/service verification.

## RM-04 exit

Removed only networkFlowRequestHash and its unused crypto/sha256 import from
internal/modules/networkflow/routes.go. Active route/resource/version-bound
hashes are unchanged; a repository source reference search finds no remaining
networkFlowRequestHash. Review also moved two fixture-policy source reads before
temporary directory allocation, so loader failures cannot leak those directories.
This touches contract-catalog-cases.mjs and contract-suite-support.mjs only.

Commands/results:

- `make lint-go-staticcheck`: PASS, 20260910T200548Z-p46555.
- `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.saved_graph_receipt_integrity`:
  PASS, 1/1, 20260910T200547Z-p46025.
- `make harness-command-surface-contract harness-evidence-contract harness-contract-tests`:
  PASS after cleanup review, respectively 20260910T200757Z-p14149,
  20260910T200757Z-p14534 and 20260910T200801Z-p16062.
- `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.measurement.exploration_graph_dom_ceiling`:
  PASS, 12/12, 20260910T200547Z-p46044; one Chromium scenario passed in 8.6s.
- `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.saved_graph_read_recovery`:
  PASS, 11/11, 20260910T200723Z-p83690; one Chromium scenario passed in 6.1s.
- `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.saved_graph_deferred_navigation`:
  PASS, 11/11, 20260910T200834Z-p17259; one Chromium scenario passed in 8.7s.
- `make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519`:
  PASS, 11/11, 20260910T200949Z-p49274.
- `git diff --check`: PASS. `make help-all` and
  `make explain-target TARGET=check DETAIL=summary`: PASS; check currently
  resolves 820 execution units, covering 1208 catalog rows.

Browser jobs ran serially against Make-owned isolated real PostgreSQL/object
store/server stacks. Retained roots contain run-summary.json, unit logs and
browser-group reports; measurement evidence preserves DOM ceilings and bucket
membership assertions. Recovery retains exact replay equality, authorization
withdrawal, duplicate-name selection and late-response fencing assertions.
No screenshot promotion or visual change was needed.

Exit: every original failing row now passes in focused verification, with real
browser validation complete. Compatibility remains as above; no receipt or
stored-data changes. Remaining risk: broad checks may expose additional failures.
Next: finalizer with RESULTS_DIR unset, broader checks, fresh make check and
handoff/scope audit.

## RM-05 verification

`env -u RESULTS_DIR make agent-finalize`: PASS, 1/1,
20260910T201114Z-p81166. Its unit-artifacts/finalize-summary.json records schema,
tier and generated-structure checks passing. Canonical retained-evidence and
scheduler drift validation were SKIPPED with results-dir-not-provided; no
qualifying exact-source full successful run existed before finalization. Skipped
maintenance is not a passing check and no duration-baseline refresh was claimed.

The following combined command passed, sequenced before the full check:

```bash
make frontend-typecheck frontend-import-boundary-check backend-module-boundary-check lint-biome lint-go lint-scripts test-catalog-check generated-artifact-policy-check json-shape-check generate-drift toolchain-drift
```

| Component | Result | Retained run ID |
| --- | --- | --- |
| frontend-typecheck | PASS 2/2 | 20260910T201149Z-p84575 |
| frontend-import-boundary-check | PASS 2/2 | 20260910T201149Z-p84579 |
| backend-module-boundary-check | PASS 3/3 | 20260910T201149Z-p84581 |
| lint-biome | PASS 2/2 | 20260910T201149Z-p84585 |
| lint-go-format | PASS | 20260910T201210Z-p85771 |
| lint-go-vet | PASS | 20260910T201212Z-p89735 |
| lint-go-staticcheck | PASS | 20260910T201216Z-p93562 |
| lint-scripts | PASS 2/2 | 20260910T201149Z-p84587 |
| test-catalog-check | PASS | 20260910T201222Z-p94221 |
| generated-artifact-policy-check | PASS 3/3 | 20260910T201149Z-p84519 |
| json-shape-check | PASS 3/3 | 20260910T201149Z-p84521 |
| generate-drift | PASS 4/4 | 20260910T201149Z-p84517 |
| toolchain-drift | PASS 2/2 | 20260910T201149Z-p84523 |

Fresh `make check`: PASS **820/820**, zero failed/skipped/cancelled units,
389670ms (6m29.670s), at **20260910T201249Z-p98482**. Every original failing
unit passed in that full run. It includes the routed frontend regression corpus;
the four separately retained real browser runs cover the changed browser seams.
No source, test or verification-input edits occurred during the full run.
No unrelated baseline failure remains in this check result.

Primary evidence is `.cartulary/test-results/20260910T201249Z-p98482/`:
run-manifest.json, run-summary.json, unit-results/, unit-logs/ and target-summaries/.
The manifest records main HEAD 1d25114774e91d73febae44069669451c774178e with a
modified working tree, normal cache mode and source digest
sha256:2326d125bb935ad77a06b610c09f64a87226f7323393a680797020f1782040fd.
Use `make explain-run RESULTS_DIR=<run-root>` and retained JSON/log artifacts to
inspect evidence. All run IDs above are relative to .cartulary/test-results/.

## Complete changed-path inventory

Authored source, tests and selectors:

- apps/web/e2e/network-flow-navigation.spec.ts
- apps/web/e2e/network-flow.spec.ts
- apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx
- apps/web/src/networkFlow/NetworkFlowSavedGraphPanel.tsx
- apps/web/src/networkFlow/networkFlowPageRecovery.test.tsx
- apps/web/src/testing/sourceOwnershipPolicy.test.ts
- internal/modules/networkflow/routes.go
- packages/ui-contracts/src/index.ts
- packages/ui-contracts/src/network-flow-selectors.test.ts
- packages/ui-contracts/src/networkFlowSelectors.ts
- tools/harness/tests/contract-suite-support.mjs
- tools/harness/tests/contract-catalog-cases.mjs (new)
- tools/harness/tests/contract-initialization-cases.mjs (new)
- tools/harness/tests/contract-test-context.mjs (new)

Authored verification inputs:

- tools/frontend_source_ownership.json
- tools/task_surface_owner.json
- tools/test_families/web.architecture.json

Make-produced derivatives:

- tools/task_surface_manifest.json
- tools/execution_topology_render_index.json

Human-only implementation evidence:

- docs/handoffs/check-harness-and-network-flow-remediation-handoff.md (new)

## Maintenance, limitations and rollback

For future catalog additions, change the authored owner inventory and applicable
references, then use Make-owned generation; do not add expected live totals.
For new frontend source paths, make an explicit exactly-once owner assignment and
route new test titles through the authored catalog. Synthetic fixture totals
remain useful assertions. Presentation readiness should verify admitted stable
identity, while display names remain content checks.

No adopted specification amendment was needed. No public command identity,
HTTP schema, fixture isolation classification, transaction approval, stored
hash, receipt or data migration changed. Existing history work and completed
handoffs are untouched. The only production changes are additive saved-result
identity attributes and removal of unreferenced private Go code.

Rollback reverts the coordinated 20 paths above, including Make projections.
It requires no data operation and deletes no history, revisions, receipts or
analyst data. No commits, staging, resets, pushes or deployment were performed.
The working tree remains on the original main/HEAD with only these authorized
changes. No broader Network Flow or subsequent refactor is authorized.

Limitations/skips: browser evidence is the four affected Chromium scenarios,
not the full browser/a11y/visual/release corpus. No intentional visible change,
visual golden update or promotion occurred. Those broader browser and release
checks were not needed for this bounded metadata/test/private-code repair.
Retained-run maintenance skips are itemized above; all requested code checks
and the full developer gate passed. Final Markdown and scope checks follow.

## RM-05 exit

`make lint-markdown`: PASS, 67590ms, at 20260910T202042Z-p56176;
the current wrapper retains adhoc/lint-markdown/tool-run-summary.json.
`git diff --check`: PASS. `git branch --show-current`, `git rev-parse HEAD`,
`git status --short`, `git diff --cached --name-only` and the changed-path audit
confirm main at the original HEAD, exactly the 20 listed authorized paths and an
empty index. No additional scope or unresolved code-check failure was found.

All five workstreams are DONE. The implementation, owner decisions, exact
verification results, evidence, compatibility, limitations and rollback are
recorded above. After this final tracker edit, repeat `make lint-markdown`,
`git diff --check` and the same branch/status/scope audit. No implementation
work or subsequent refactor is pending or implied.
