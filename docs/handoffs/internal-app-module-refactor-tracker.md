# internal-app Module Refactoring Tracker and Handoff

## 1. Current Scope, Authority, and Source Hierarchy

| Field | Value |
| --- | --- |
| Target path | `internal/app` |
| Target label | `internal-app` |
| Output path | `docs/handoffs/internal-app-module-refactor-tracker.md` |
| Current tracker status | Current-state planning tracker after the completed 2026-07-06 recovery remediation work. |
| Allowed changes for this iteration | This tracker only. No production code, generated files, phase owner inputs, migrations, tests, or harness manifests are changed by this pass. |
| Implementation posture | Future implementation requires separate authorization. This tracker may recommend removal or movement of behavior; it does not authorize the code changes. |
| Current boundary thesis | `internal/app` should be an application assembly and CLI facade boundary. It should not own recovery semantics, migration-history evidence semantics, restore workbook probe semantics, object-store migration semantics, view/workbook semantics, or platform adapter policy. |

Authority and source hierarchy:

- Core 00 through Core 04 own current Base Profile product behavior. Core 05 applies only to claim-bearing timed, benchmark, fixture-sensitive, or publication evidence.
- Core 01 section 12.2.1 owns operator recovery CLI logical command grammar, result schema, progress schema, timeout defaults, exit-code mapping, backup selection rules, and operator recovery error vocabulary.
- Core 04 owns local-operator recovery authority, the no-listener/no-browser boundary, recovery-operation exclusion, restore-target preflight, encrypted journal requirements, safe administrative-audit summaries, and redaction expectations.
- `docs/testing-harness-nlspec.md` owns harness mechanics and evidence routing. It classifies legacy `deployment_admin` recovery tests as negative-only and migration-history evidence as database-contract or migration-evidence evidence.
- `docs/domain.md` owns vocabulary and concept-boundary interpretation. Backup/restore remains deployment-local operator-facing behavior, not workbook or public route behavior.
- `internal/modules/*` owns domain and application behavior. `internal/platform/*` owns runtime plumbing, storage/config adapters, auth primitives, and database support. `contracts/*` is the derived contract layer.

Repository state inspected for this iteration:

- `internal/app/*`
- `internal/modules/recovery/operatorcli/*`
- `internal/modules/recovery/workbook_probe.go`
- `internal/platform/postgres/migrationevidence/*`
- `cmd/operator/operator_phase10_test.go`
- `cmd/server/main_phase10_recovery_sentinel_test.go`
- `tools/phase0_test_map.json`
- `tools/phase10_test_map.json`
- `docs/testing/phase10_coverage_ledger.md`
- `docs/testing-harness-nlspec.md`
- Core 00, Core 01, Core 04, and `docs/domain.md`

## 2. Current Inventory

### `internal/app`

| Path | Current responsibility | Boundary diagnosis | Tests and evidence posture | Next action |
| --- | --- | --- | --- | --- |
| `internal/app/runtime.go` | Builds the server runtime: validates config, bootstraps telemetry, opens Postgres/object store, runs bootstrap preflight, seeds reference data, wires jobs, WebSocket hub, auth keys, pagination codec, module route registrars, readiness, and HTTP handler. | Legitimate `internal/app` assembly facade. Keep thin and reject domain behavior. | `runtime_phase0_test.go`, `runtime_phase0_integration_test.go`, server/process tests, and testutil callers. | Keep. Add assembly characterization only before any future runtime wiring movement. |
| `internal/app/migrate.go` | CLI facade for migration command parsing, config load, SQL open, embedded SQL source selection, `postgres.Migrate`, and remediation report emission. | Acceptable CLI facade over `platform/postgres`; optional future simplification only. | `migrate_test.go`; Phase 0 accounting uses app package tests. | Defer. Do not move unless app facade thinning needs it. |
| `internal/app/operator.go` | Shared operator runner, object-store init, support-only object-store migration, legacy recovery command parser/DTOs, legacy `backup capture` and restore handlers, restore target preflight helpers, advisory-lock helpers, JSON output helpers, and storage/config opening helpers. | Mixed responsibility. Some pieces are facade support; legacy recovery handlers and support-only object-store migration should not become product conformance. | `operator_test.go`; blocked/negative-only Phase 10 operator process rows in `cmd/operator/operator_phase10_test.go`. | Split/remove legacy recovery burden; decide whether object-store migration should be removed or extracted as support-only. |
| `internal/app/operator_recovery.go` | Bridges canonical recovery commands into `internal/modules/recovery/operatorcli` and currently performs canonical backup/restore/verification orchestration using recovery and platform adapters. | Partially improved facade, but still owns too much recovery orchestration. | `operator_recovery_test.go`; recovery `operatorcli` tests; Phase 10 target planning. | Move orchestration into a recovery-owned service/package with injected app/platform adapters. |
| `internal/app/operator_migration_evidence.go` | Thin CLI wrapper around `internal/platform/postgres/migrationevidence.Build`; re-exports migration-evidence result types for app tests/callers. | Improved after remediation. App is facade only, but names still carry `Operator*` terminology from the old location. | `operator_migration_evidence_test.go`; `operator_migration_evidence_integration_test.go`; Phase 0 migration-evidence rows. | Plan neutral naming cleanup under the platform owner. |
| `internal/app/recovery_probe.go` | Removed. Restore workbook probe moved to `internal/modules/recovery/workbook_probe.go`. | Closed remediation item. Not current inventory. | `internal/modules/recovery/workbook_probe_test.go`; server/browser restore callers now use recovery owner. | No app work. Preserve owner routing in future changes. |

### Adjacent Extracted Owners and Accounting Inputs

| Path | Current responsibility | Owner | Notes |
| --- | --- | --- | --- |
| `internal/modules/recovery/operatorcli/cli.go` | Canonical recovery CLI parsing, result/progress schemas, timeout bounds, negative-only legacy rejection, error mapping, and final stdout/stderr behavior. | Recovery module, subordinate to Core 01/Core 04. | App should call this package rather than owning CLI grammar. |
| `internal/modules/recovery/operatorcli/journal.go` | Encrypted recovery journal append and safe administrative-audit summary append. | Recovery module, subordinate to Core 04. | App currently supplies DB/key-loading adapter. |
| `internal/modules/recovery/workbook_probe.go` | Restore-verification workbook probe using a deterministic incident lookup and owner-registered timeline workbook query. | Recovery module, subordinate to Core 01 probe contract. | The old app probe is gone. |
| `internal/platform/postgres/migrationevidence/evidence.go` | Migration manifest/source/goose ledger evidence construction with neutral schema `cartulary.migration_history_evidence.v1`. | Platform Postgres / migration-evidence support. | Should not close operator recovery conformance. |
| `db/migrations/00045_operator_recovery_journal.sql` | Authored SQL input for encrypted operator recovery journal storage. | Migration SQL input; generated SQL models are downstream. | Do not hand-edit generated SQL model output. |
| `tools/schema_object_ownership_manifest.json` | Owner input assigning the recovery journal table/index ownership. | Harness/schema ownership input. | Edit this owner input before regenerated or drift outputs when schema ownership changes. |
| `tools/phase0_test_map.json` | Phase 0 owner input for runtime/bootstrap and migration-evidence rows. | Harness owner input. | Update before generated ledgers/schedules if tests move. |
| `tools/phase10_test_map.json` | Phase 10 owner input. Current E-10-01 is blocked and legacy `deployment_admin` recovery tests are negative-only. | Harness owner input. | Update before generated ledgers/schedules when canonical operator recovery evidence replaces legacy rows. |

## 3. Boundary Diagnosis by Responsibility

| Responsibility | Current location | Intended owner | Diagnosis | Keep / move / remove |
| --- | --- | --- | --- | --- |
| Server application assembly | `runtime.go` | `internal/app` | Correct. App composes platform services and module route registrars. | Keep. |
| Runtime route semantics | Owning modules registered by `runtime.go` | `internal/modules/*` and `platform/httpapi` | App should only wire registrars and dependency set. | Keep assembly only. |
| Migration CLI facade | `migrate.go` | `internal/app` facade over `platform/postgres` | Acceptable. Migration mechanics belong to Postgres platform support. | Keep or optionally thin later. |
| Canonical recovery CLI grammar/result/progress | `internal/modules/recovery/operatorcli`; app delegates through `operator_recovery.go` | Core 01/Core 04 via recovery module | Correct owner after remediation. | Keep outside app. |
| Canonical recovery operation orchestration | `operator_recovery.go` | `internal/modules/recovery` service/package with app/platform adapter ports | Still too much recovery behavior in app. | Move in a future slice. |
| Legacy recovery commands and DTOs | `operator.go` | No current conformance owner; negative-only if retained | Compatibility burden. These commands must not become a second public contract. | Remove by default unless an owner gives continuing support value. |
| Backup metadata inspection legacy support | `operator.go` | Recovery support or remove | Still `deployment_admin` gated and legacy-shaped. Phase 10 treats it as blocked/negative-only for conformance. | Replace with canonical `backup inspect latest` evidence, then remove or extract support-only code. |
| Object-store init | `operator.go` | `internal/app` CLI facade over `platform/objectstore` | Deployment-local support command; not recovery conformance. | Keep thin or move to platform support if it grows. |
| Object-store migration | `operator.go` plus recovery helpers | Support-only recovery/platform migration tooling if retained | Large support-only behavior in app. Future value is unclear after recovery remediation. | Decide remove vs extract; removal is the default unless value is documented. |
| Migration-history evidence | `operator_migration_evidence.go` facade; `platform/postgres/migrationevidence` owner | `internal/platform/postgres/migrationevidence` | Correct owner after remediation, but names still say `Operator*`. | Keep owner; plan neutral rename. |
| Restore workbook probe | `internal/modules/recovery/workbook_probe.go` | `internal/modules/recovery` | Closed. No longer app-owned. | Keep outside app. |
| Recovery journal and safe audit | `operatorcli/journal.go`; app supplies adapter | Recovery/Core 04 | Correct owner, but adapter integration is still in app orchestration. | Move adapter wiring with orchestration if it clarifies the facade. |
| Phase/accounting evidence | Phase owner inputs and generated ledgers | `docs/testing-harness-nlspec.md` and Make-owned generation | Evidence maps are not architecture. | Owner-input first; no generated hand edits. |

## 4. Gap Register

| Gap | Remediation | Affected area | Rationale | Expected long-term benefit | Compatibility or migration impact | Risk if left unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Stale tracker sections contradicted current repo. | Rewrite sections 1-12 from current repository state; archive W0-W8 remediation history. | Documentation | Prevent rediscovery and avoid reopening completed recovery work. | Another agent can resume from current facts. | No runtime impact. | Future agents may follow obsolete `recovery_probe.go` or old `deployment_admin` recovery guidance. | `make lint-markdown`; record result in this tracker. |
| Legacy recovery handlers and DTOs remain in `operator.go`. | Plan future removal of `backup capture`, `backup-metadata latest`, legacy restore, and legacy restore-verify parsing/DTOs unless an owner explicitly promotes them as support-only. | Implementation, tests, harness accounting | These commands are negative-only and should not become compatibility burden. | Smaller app facade and less accidental conformance drift. | Canonical commands stay unchanged; legacy tests must be deleted, rewritten, or kept negative-only. | Accidental reactivation of non-conformant recovery behavior. | `make build-operator`; `make backend-unit`; `make backend-process`; `make phase-slice PHASE=phase10`; phase map owner-input first. |
| Canonical recovery orchestration still executes inside `operator_recovery.go`. | Move operation orchestration into a recovery-owned package or service with injected app/platform adapters. | Implementation, tests | App should compose the CLI facade, not own backup/restore algorithms. | Recovery behavior becomes easier to test, extend, and reason about for later phases. | Internal movement only; preserve canonical stdout, progress, exit codes, and errors. | App remains a mixed domain/platform coordinator. | `make build-operator`; `make backend-unit`; `make backend-store`; `make backend-process`; `make service-backed-slice PHASE=phase10`. |
| Phase 10 E-10-01 remains blocked around legacy operator evidence. | Replace blocked legacy `cmd/operator/operator_phase10_test.go` recovery evidence with canonical Core 01/Core 04 process evidence. | Harness accounting, tests, documentation | Conformance evidence must match owner docs, not old command aliases. | AC-402, AC-427, and AC-428 can close on current commands. | Update `tools/phase10_test_map.json` first; regenerate ledgers/schedules through Make only. | Recovery conformance remains blocked or misleading. | `make json-shape-check`; `make phase-ledgers`; `make phase-schedules`; `make phase-ledger-drift`; `make phase-schedule-drift`; `make phase-slice PHASE=phase10`. |
| Migration-evidence names still carry `Operator*` terminology under the platform owner. | Plan neutral internal rename in `migrationevidence`; keep app aliases only if needed by CLI facade tests. | Implementation, tests, documentation | Database-contract evidence should not look like operator recovery behavior. | Cleaner ownership and easier evidence classification. | Internal API/test churn only. | Future evidence may be misclassified as recovery conformance. | `make backend-unit`; `make backend-integration`; `make phase-slice PHASE=phase0`; `make json-shape-check`. |
| Object-store migration remains support-only but large in `operator.go`. | Mark as legacy/support-only. Default future remediation is removal unless a current owner states continuing value; if retained, move orchestration/artifact writing to recovery support code. | Implementation, tests, documentation | Avoid carrying migration-era behavior that does not support future product architecture. | Less unsupported surface in app. | Potential removal affects local support workflows only, not product conformance. | App keeps indefinite support behavior that complicates later expansion. | `make build-operator`; `make backend-unit`; `make backend-process`; support-only phase rows stay non-conformance. |

## 5. Recommended Next Workstreams

| Workstream | Order | Dependencies | Risk | Exit criteria |
| --- | ---: | --- | --- | --- |
| WS-01 Tracker reconciliation | 1 | None | Low | Current tracker reflects current repo; validation result recorded. |
| WS-02 Canonical Phase 10 evidence cleanup | 2 | WS-01 | High | E-10-01 no longer depends on negative-only legacy recovery tests; owner inputs and generated ledgers/schedules agree. |
| WS-03 Legacy recovery code removal | 3 | WS-02 | High | Old recovery aliases/DTOs/handlers are absent or isolated as explicitly support-only; canonical commands unchanged. |
| WS-04 Recovery orchestration extraction | 4 | WS-03 | High | `internal/app/operator_recovery.go` becomes facade wiring; recovery module owns operation bodies and tests. |
| WS-05 Migration-evidence terminology cleanup | 5 | WS-01 | Medium | Platform migration-evidence package exposes neutral type names; app aliases are minimized. |
| WS-06 Object-store migration support decision | 6 | WS-01, preferably WS-03 | Medium | Support command is removed or extracted with documented non-conformance status and validation. |

## 6. Workstream Sequencing Detail

| Workstream | Implementation notes | Generated-file handling | Rollback scope |
| --- | --- | --- | --- |
| WS-01 | Documentation-only tracker edit. Do not touch code, owner inputs, generated outputs, or tests. | Not applicable. | Revert this tracker file only. |
| WS-02 | Add canonical process tests for `operator backup inspect latest`, `backup create`, `restore latest`, `restore-verify latest`, and `restore-verify due`; update Phase 10 owner input before generated companions. | Edit `tools/phase10_test_map.json`; run Make generators and drift checks. Do not hand-edit generated ledgers/schedules. | Revert tests and owner-input/generated companion changes together. |
| WS-03 | Delete or isolate old recovery parser branches, DTOs, result constants, and legacy process tests after canonical coverage exists. | Update phase owner input first if test paths or row evidence change. | Revert removal as one app/operator slice. |
| WS-04 | Introduce a recovery-owned operation service with ports for config loading, Postgres/object-store opening, backup storage creation, projection rebuilder, clock, locking, journal, and audit. App supplies adapter values only. | No generated files unless new SQL/schema ownership is added. | Revert service extraction and app wiring together. |
| WS-05 | Rename migration-evidence types and helpers to remove `Operator` prefix in platform package; keep JSON schema `cartulary.migration_history_evidence.v1`. | No generated files expected. Update Phase 0 owner input only if test identifiers or paths change. | Revert rename and app aliases together. |
| WS-06 | First decide whether support-only object-store migration still has continuing value. If not, remove command, tests, and support accounting. If yes, extract from app and mark non-conformance. | Owner-input first if phase accounting changes; generated companions through Make only. | Revert command/test/accounting changes together. |

## 7. Validation Plan

| Scenario | Make-owned target | When to run | Notes |
| --- | --- | --- | --- |
| Tracker-only edit | `make lint-markdown` | Required for this pass. | Result recorded in Section 10. |
| App/operator build | `make build-operator` | Any operator facade or recovery command change. | Confirms binary composition after app changes. |
| Narrow unit coverage | `make backend-unit` | Any app, recovery CLI, migration-evidence, object-store support, or parser change. | Includes app and recovery unit coverage selected by Make. |
| Service-backed process evidence | `make backend-process` | Any operator process or server process evidence change. | Baseline before this edit passed at `.cartulary/test-results/20260706T113639Z-p2681164`. |
| Recovery phase slice | `make phase-slice PHASE=phase10` | Recovery/operator evidence, phase map, restore, journal, or support-only cleanup. | Required before claiming E-10-01 progress. |
| Recovery service-backed slice | `make service-backed-slice PHASE=phase10` | Restore/verify behavior or app-to-recovery orchestration movement. | Use after WS-04 or high-risk recovery changes. |
| Phase 0 slice | `make phase-slice PHASE=phase0` | Migration-evidence movement, naming, or accounting changes. | Migration evidence is Phase 0/database-contract evidence. |
| Generated/artifact policy | `make generated-artifact-policy-check`; `make json-shape-check` | Any owner-input, manifest, schema, or generated-policy-adjacent change. | Must pass before generated drift claims. |
| Phase generated companions | `make phase-ledgers`; `make phase-schedules`; `make phase-ledger-drift`; `make phase-schedule-drift` | When phase owner inputs change. | Generated ledgers/schedules are downstream. |
| Boundary guardrails | `make backend-module-boundary-check` | Package movement, new imports, or ownership extraction. | Use especially after WS-03, WS-04, or WS-06. |

## 8. Generated-File and Owner-Input Handling Rules

- Do not hand-edit generated roots declared by `tools/generated_artifact_policy.json`: `internal/gen/**`, `packages/protocol-ts/src/generated/**`, and `packages/ui-contracts/src/generated/**`.
- Do not hand-edit generated harness/topology outputs, generated phase ledgers, generated schedules, or generated topology/render indexes.
- For phase-map or accounting changes, edit owner inputs first, such as `tools/phase0_test_map.json` or `tools/phase10_test_map.json`, then run the Make-owned generation and drift targets.
- For schema ownership changes, edit owner manifests such as `tools/schema_object_ownership_manifest.json` before running `make json-shape-check`.
- For SQL changes, author migrations under `db/migrations`, update migration-history owner inputs as required, and regenerate downstream artifacts only through `make generate` or the relevant Make target.
- Do not hand-edit `go.sum`, `pnpm-lock.yaml`, or tool-managed dependency/install artifacts.

## 9. Handoff Log Format

Future sessions must append one row per meaningful work slice:

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `<UTC timestamp>` | `<agent/session>` | `<what changed or what was learned>` | `<paths>` | `<Make targets and focused reads>` | `<pass/fail plus run root when available>` | `<none or concrete blocker>` | `<next dependency-aware action>` |

Final reports should state planning summary, files inspected or changed, substantive edits, verification commands and results, and skipped checks with reasons.

## 10. Current Handoff Log

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-06T11:41:55Z | Codex tracker reconciliation | Started current-state tracker update after recovery remediation. | Inspected this tracker, `internal/app`, recovery `operatorcli`, recovery workbook probe, platform migration evidence, Phase 10 ledger/map, harness NLSpec, Core 01/Core 04, and domain vocabulary. Touched only this tracker. | `git status --short`; `sed`; `find`; `rg`; baseline `make lint-markdown`; baseline `make backend-process`. | Baseline `make lint-markdown` passed; baseline `make backend-process` passed at `.cartulary/test-results/20260706T113639Z-p2681164`. | None. | Run final `make lint-markdown` after tracker edit and record the result. |
| 2026-07-06T11:44:21Z | Codex tracker reconciliation | Current-state tracker rewrite completed. | Touched only `docs/handoffs/internal-app-module-refactor-tracker.md`. | `make lint-markdown`; final rerun of `make lint-markdown`. | Both passed. | None. | Handoff complete for tracker-only iteration. |

## 11. Completion Criteria for This Tracker Iteration

| Criterion | Status | Evidence |
| --- | --- | --- |
| Current `internal/app` inventory matches the repository. | PASS | Section 2 lists `operator_recovery.go` and removes `recovery_probe.go` from current inventory. |
| Adjacent extracted packages are inventoried. | PASS | Section 2 lists `operatorcli`, recovery workbook probe, migration evidence, journal migration, and owner inputs. |
| Each remaining app responsibility has a boundary diagnosis. | PASS | Section 3 classifies keep, move, remove, and support-only responsibilities. |
| Gaps include remediation, affected area, rationale, benefit, compatibility impact, risk, and validation. | PASS | Section 4 gap register. |
| Workstreams are ordered by dependency and risk. | PASS | Sections 5 and 6. |
| Validation plan uses Make-owned targets. | PASS | Section 7. |
| Generated-file and owner-input rules are explicit. | PASS | Section 8. |
| Handoff log format is explicit. | PASS | Section 9. |
| Required tracker validation is recorded. | PASS | `make lint-markdown` passed after the tracker rewrite and passed again after recording the validation row. |

## 12. Completed Recovery Remediation Log

The following archive preserves the completed W0-W8 recovery remediation history. It is historical evidence, not the current work queue.

| Workstream | Status | Started | Completed | Evidence |
| --- | --- | --- | --- | --- |
| W0 Tracker bootstrap | DONE | 2026-07-06T02:32:27Z | 2026-07-06T02:34:22Z | Tracker scope changed for the remediation run; `make lint-markdown` passed. |
| W1 Spec/docs/accounting cleanup | DONE | 2026-07-06T02:34:22Z | 2026-07-06T02:37:02Z | Updated Core 01 probe/alias text, dev guide alias guidance, Phase 0/10 owner inputs, and Make-regenerated ledgers/schedules. Validation passed: `make phase-ledgers`, `make phase-schedules`, `make json-shape-check`, `make lint-markdown`, `make phase-ledger-drift`, `make phase-schedule-drift`. |
| W2 Canonical recovery CLI contract | DONE | 2026-07-06T02:37:02Z | 2026-07-06T02:44:43Z | Added canonical recovery CLI interception for the five Core logical commands, typed `cartulary.operator_recovery_result.v1` and `cartulary.operator_recovery_progress.v1` emitters, timeout/default parsing, closed exit/error mapping, and negative-only legacy rejection. Validation passed: `make build-operator`, `make backend-unit`. |
| W3 Recovery orchestration extraction | DONE | 2026-07-06T02:44:43Z | 2026-07-06T02:51:58Z | Added `internal/modules/recovery/operatorcli` as recovery-owned contract/dispatch; app delegates canonical parsing/progress/result/timeout/error mapping while supplying process adapters. Validation passed: `make build-operator`, `make backend-unit`, `make backend-module-boundary-check`. |
| W4 Security, preflight, lock, journal, audit | DONE | 2026-07-06T02:51:58Z | 2026-07-06T03:00:18Z | Added shared recovery-operation advisory locking, encrypted operator recovery journal persistence, safe administrative-audit summaries, journal/audit error mapping, and migration `00045_operator_recovery_journal.sql` with generated SQL model refresh. Validation passed: `make build-operator`, `make backend-unit`, `make generate`, `make migration-drift`, `make backend-store`, `make backend-process`. |
| W5 Restore probe ownership move | DONE | 2026-07-06T03:00:18Z | 2026-07-06T03:07:45Z | Moved `RestoreVerificationWorkbookProbe` from `internal/app` into `internal/modules/recovery`, updated operator/server/browser restore call sites, and changed empty workbook-query results from probe failure to success. Validation passed: `make build-operator`, `make backend-unit`, `make backend-integration`, `make build-server`, `make service-backed-slice PHASE=phase10`. |
| W6 Migration-evidence owner move | DONE | 2026-07-06T03:07:45Z | 2026-07-06T03:17:58Z | Added `internal/platform/postgres/migrationevidence` as migration-history evidence owner with schema `cartulary.migration_history_evidence.v1`; reduced app to a CLI facade; removed deployment-admin authorization from the command path and real PostgreSQL test. Validation passed: `make build-operator`, `make backend-unit`, `make backend-integration`, `make phase-slice PHASE=phase0`, `make backend-module-boundary-check`, `make lint-markdown`. |
| W7 Support-command split and boundary guardrails | DONE | 2026-07-06T03:17:58Z | 2026-07-06T03:23:46Z | `object-store init` remains local support tooling; `object-store-migration run` remains support-only, no longer requires or authorizes a deployment-admin credential, records fixed `local_os_execution`, and retains safe-output process coverage. Validation passed: `make build-operator`, `make backend-unit`, `make backend-process`, `make backend-module-boundary-check`. |
| W8 Final validation and handoff | DONE | 2026-07-06T03:23:46Z | 2026-07-06T03:30:28Z | Final validation passed: `make build-operator`, `make generated-artifact-policy-check`, `make json-shape-check`, `make backend-module-boundary-check`, `make lint-markdown`, `make phase-slice PHASE=phase0`, `make phase-slice PHASE=phase10`, `make service-backed-slice PHASE=phase10`, `make agent-finalize`, and `make test-fast`. Initial final `make json-shape-check` failed because `tools/schema_object_ownership_manifest.json` lacked the new recovery journal table/index ownership; fixed by adding a recovery-owned manifest entry. `make agent-finalize` reported retained-run maintenance skipped because `RESULTS_DIR` was unset. |
