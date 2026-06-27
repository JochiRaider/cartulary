# Operational Recovery Stand-Up Handoff Tracker

Status: implementation-support live handoff tracker. This document is binding only for sequencing, sprint status, validation recording, and handoff interpretation for operational recovery stand-up work. It is not a Core product specification and does not create, widen, narrow, or replace Base Profile or extension-profile conformance behavior.

Core 00 through Core 04 remain the implementation-conformance authority. Core 05 applies only at claim-bearing publication boundaries. `docs/testing-harness-nlspec.md` owns command invocation, target selection, scheduling, fixture lifecycle, artifact emission, cleanup, and verification gates. `docs/opentelemetry-instrumentation-nlspec.md` owns telemetry configuration containment, privacy, exporter behavior, and log correlation. `docs/domain.md` remains the domain vocabulary and concept-boundary reference.

Default decisions locked for this tracker:

- Tracker location: `docs/handoffs/operational-recovery-standup-handoff.md`.
- Package profile: `deployment_profile = "on_prem"` for the MVP stand-up package.
- Scheduler shape: systemd timers invoking package-local one-shot scripts.
- Backup cadence: every 6 hours.
- Restore-verification cadence: daily due-runner execution, satisfying the Core 01 at-least-every-7-days floor.
- Backup storage realization: current filesystem-backed `roots.backup_storage` with authenticated encrypted artifact envelopes and `CARTULARY_RECOVERY_MASTER_KEY`.
- Core updates: no owner-spec edits are planned unless Sprint 0 identifies a concrete conformance ambiguity or functionality-preservation need.
- Public surface boundary: no public `/api/v1/backups*`, `/api/v1/restores*`, `/api/v1/restore-verifications*`, `/ws/v1/backups*`, `/ws/v1/restores*`, or `/ws/v1/restore-verifications*` family is authorized by this tracker.

Normative terms in this tracker have the following local meanings:

| Term | Meaning inside this tracker |
| --- | --- |
| `MUST` | Required for this handoff tracker to consider the named sprint or work unit complete. |
| `MUST NOT` | Forbidden for work performed under this tracker. |
| `SHOULD` | Required default unless this tracker records a specific exception and validation effect. |
| `MAY` | Optional tracker behavior only when omission behavior is stated in the same row or paragraph. |
| `default` | Required value, interpretation, action, or status when a more specific value is omitted. |

Statement classes:

| Statement class | Meaning | Required handling |
| --- | --- | --- |
| Tracker-owned requirement | Sequencing, status, validation, or handoff interpretation rule owned by this document. | Binding for work performed under this tracker. |
| Core-owner restatement | Compact pointer to behavior owned by Core 00 through Core 04. | MUST cite the owner; MUST NOT be treated as independent product authority. |
| Implementation guidance | Preferred implementation organization that does not change observable product behavior. | Follow unless a stronger local code boundary requires a recorded exception. |
| Validation requirement | Evidence required before a sprint or unit is complete. | Blocking unless recorded as skipped with reason and owner. |
| Research support | Non-normative rationale from research reports R01 through R09. | MAY explain a decision, but MUST NOT override owner documents. |
| Out-of-scope exclusion | Named work this tracker does not authorize. | MUST be handled by a separate tracker, owner-spec change, or product task. |

## Pinned Inputs And Source Snapshot

Default drift behavior: the tracker is bound to the repository snapshot at the time the file is added until this file is explicitly updated. If a source file changes later, the changed source informs future tracker revision but does not silently alter this tracker.

| Source path | Snapshot identity | Authority role | Required use | Drift handling |
| --- | --- | --- | --- | --- |
| `docs/spec/00_document_set_status_and_precedence.md` through `docs/spec/04_security_deployment_and_conformance.md` | Repo snapshot at tracker creation time. | Product-conformance owner corpus. | Resolve backup-set, restore, verification, runtime-root, deployment-admin, trust-boundary, and acceptance-criteria behavior. | If this tracker and Core differ, Core governs and this tracker needs repair. |
| `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` | Repo snapshot at tracker creation time. | Claim-publication owner only. | Use only if operational recovery evidence is later published as timed, benchmark, fixture-sensitive, or claim-bearing evidence. | Do not cite this tracker as Core 05 evidence. |
| `docs/testing-harness-nlspec.md` | Repo snapshot at tracker creation time. | Harness mechanics and Make-command owner. | Resolve command invocation, public target addition, generated task-surface handling, retained artifacts, summaries, cleanup, and finalization. | Update this tracker before relying on changed command mechanics. |
| `docs/opentelemetry-instrumentation-nlspec.md` | Repo snapshot at tracker creation time. | Telemetry subsystem owner. | Preserve telemetry containment, privacy, exporter behavior, and log correlation in recovery jobs and diagnostics. | Runtime diagnostics must be reconciled with this owner before completion. |
| `docs/domain.md` | Repo snapshot at tracker creation time. | Vocabulary and concept-boundary reference. | Resolve backup/restore, deployment administration, evidence, object blob, and operator-facing wording. | If vocabulary differs from a Core owner, owner section governs and this tracker needs repair. |
| `docs/guides/*.md` | Repo snapshot at tracker creation time. | Implementation-support guidance. | Use for existing build, package, deployment, and development conventions. | Guides do not override Core or harness owners. |
| `docs/research/R01-*` through `docs/research/R09-*` | Repo snapshot at tracker creation time. | Supporting research only. | Use where materially helpful to justify recovery-package decisions. | Research never creates conformance behavior. |

## Applicability And Omission Semantics

This tracker covers planning and implementation of operational recovery for the existing MVP on-prem stand-up package. The intended closure is scheduled backup capture, retained backup evidence, scheduled restore verification against isolated restored state, and an operator runbook. Work outside that package is out of scope unless required by Core owner conformance or by tests touched by this tracker.

Omission behavior: a file, route, command, scheduler unit, package artifact, deployment profile, or verification target not named in this tracker is out of scope unless a listed owner document or touched implementation boundary requires it. Out-of-scope items MUST NOT be added by implication from research reports, guides, or convenience scripts.

The MVP package remains an on-prem local stand-up package. Completing this tracker MUST NOT by itself represent the package as disconnected-profile conformance.

## Sprint Governing Table

Sprint status defaults to `not-started`. A sprint can move to `complete` only when the validation status names passing evidence or a recorded skip with owner and impact.

| Sprint | Objectives | Implementation status | Work requirements | Validation status | Validation requirements |
| --- | --- | --- | --- | --- | --- |
| Sprint 0 | Characterize current recovery behavior and owner sufficiency. | `complete` | Inspect Core 01/Core 04, Phase 10 map, operator commands, `deploy/mvp`, recovery module, existing smoke targets, and relevant R01-R09 materiality. | `complete` | Record gaps; decide whether Core edits are required; confirm no production movement occurred before characterization. |
| Sprint 1 | Add and pin backup capture behavior. | `complete` | Add tests first, then implement deployment-local backup capture using existing recovery helpers and encrypted backup storage. | `complete` | Prove latest retained backup under 24 hours old, 30-day retention, artifact durability, redaction, deployment-admin authorization, and fail-closed missing recovery key behavior. |
| Sprint 2 | Add scheduled MVP recovery package wiring. | `complete` | Add systemd unit and timer templates, one-shot package scripts or services, restore target config example, restore target marker, and environment placeholders. | `complete` | Static and package smoke prove timers and scripts are present, non-secret, and use package-local configs without broadening the TOML schema. |
| Sprint 3 | Add scheduled restore-verification proof. | `complete` | Wire `operator restore-verify due` against an isolated target and record pass/fail artifacts plus metadata. | `complete` | Prove the due runner verifies backups due by age or basis change, writes restore-verification artifacts, and rejects unsafe or unmarked targets. |
| Sprint 4 | Add operational recovery smoke gate. | `complete` | Add Make-owned `standup-operational-recovery-smoke` through topology owner inputs; regenerate generated task-surface outputs through owner tooling only. | `complete` | Smoke captures a backup, verifies latest metadata, runs due restore verification, proves no public backup/restore route family, and retains summary artifacts. |
| Sprint 5 | Finalize docs and closeout evidence. | `complete` | Update package runbook, tracker live rows, skipped-check notes, residual risks, and retained-run maintenance. | `complete` | `make lint-markdown`, recovery smoke, `make phase-slice PHASE=phase10`, broader closeout as risk requires, and `make agent-finalize` with a retained successful `make check` run root when available. |

## Implementation Plans By Sprint

Implementation order for every sprint: characterize current behavior, add or pin tests if needed, extract pure utilities only when needed, extract hooks after state/effect boundaries are clear, extract presentational components after props stabilize, then extract surface-specific components. This operational recovery tracker is backend/runtime/package focused; frontend hook and component extraction is expected to remain out of scope.

### Sprint 0 - Characterization

Plan order: characterize behavior first; add tests only after gaps are confirmed; no package or implementation edits before current behavior is recorded.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Tracker setup | `docs/handoffs/operational-recovery-standup-handoff.md` | Local sprint/status authority, assumptions, evidence notes, and handoff interpretation. | Product conformance behavior or implementation claims. | Inputs: owner docs and repo inspection; output: pinned facts and open gaps. |
| Recovery fact map | tracker section | Current Phase 10 and MVP package state. | New implementation. | Inputs: recovery module, operator, package docs, phase maps; output: current-state table. |
| Owner sufficiency review | Core/spec refs | Whether Core edits are needed. | Spec churn by default. | Inputs: Core 01 §12 and Core 04 §9.14/§12.3.3; output: `no Core edits` unless a concrete ambiguity is found. |

Sprint 0 required characterization rows:

| Area | Characterization requirement | Completion evidence |
| --- | --- | --- |
| Core backup contract | Confirm the current Core defines `backup_set`, `backup_attestation`, 24-hour freshness, 30-day retention, coherent restore, restore verification, and no-public-route behavior. | Owner-doc inspection recorded in this tracker. |
| Current operator surface | Confirm existing `cartulary-operator` commands include metadata inspection, restore latest, restore-verify latest, restore-verify due, object-store init, and object-store migration, and identify whether backup capture is missing. | Operator usage inspection and Phase 10 evidence map. |
| Recovery module substrate | Confirm existing recovery helpers can capture Postgres and object-store artifacts, write encrypted backup storage artifacts, verify durability, restore selected backup sets, and write restore-verification artifacts. | Code/test inspection and Phase 10 ledger rows. |
| MVP package gap | Confirm `deploy/mvp` currently allocates a persistent backup root but does not claim backup/restore conformance because backup capture, retention, and restore-verification scheduling are separate. | README and smoke target inspection. |
| Owner update decision | Decide whether Core 01/Core 04 already own enough behavior for implementation or need a narrow owner update to preserve functionality. | Recorded Sprint 0 decision. |
| Research support | Confirm R03, R06, and R07 are material supporting rationale and that other R01-R09 reports do not alter owner behavior. | Research-support table updated only as rationale. |

### Sprint 1 - Backup Capture

Plan order: add/pin tests first, then implement the narrow operator surface and reusable capture utilities. No scheduler or runbook work is complete until Sprint 2.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Capture tests | `cmd/operator`, `internal/modules/recovery` | Capture contract evidence for operator authorization, backup metadata shape, retention, encryption, durability, and redaction. | Scheduler behavior or package install docs. | Inputs: service-backed fixtures; output: failing tests before implementation and passing evidence after implementation. |
| Operator capture | `internal/app/operator.go` | Deployment-local `backup capture` command or equivalent operator surface. | Public API routes, workbook workflow, browser controls, or incident-scoped jobs. | Inputs: source config, active deployment-admin email, quiescence proof; output: retained `backup_set` and structured JSON result. |
| Quiescence proof | recovery/operator support | App-stopped or write-quiesced capture evidence for the MVP package mechanism. | General maintenance mode or arbitrary PITR. | Input: stopped app container/proof file or equivalent process proof; output: artifact-bound proof metadata. |

Sprint 1 capture rules:

- The operator command MUST require an active `deployment_admin`.
- The command MUST be deployment-local and MUST NOT add public HTTP or WebSocket routes.
- Capture MUST produce one retained `backup_set` with exactly one declared `consistency_point_at`.
- Capture MUST include Postgres restore artifact/anchor, object-store restore artifact/anchor, integrity proof, and backup metadata sufficient for AC-398.
- `retained_until`, Postgres anchor retention, and object-store anchor retention MUST be at least `created_at + 30 days`.
- Filesystem-backed incident-bearing artifacts MUST use encrypted backup storage and fail closed without `CARTULARY_RECOVERY_MASTER_KEY`.
- The MVP package mechanism MAY stop the app container briefly; direct out-of-band writes during capture are unsupported.

### Sprint 2 - Package Scheduling

Plan order: use Sprint 1 capture command as the stable interface, then add package-local scheduling templates and scripts. Do not add deployment-config schedule keys unless Core 04 is explicitly updated.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Systemd templates | `deploy/mvp/systemd/` | Backup and restore-verify timer/service examples. | Runtime TOML schema, product API, or secrets. | Output: `cartulary-backup` and `cartulary-restore-verify` service/timer templates. |
| Recovery scripts | `deploy/mvp/scripts/` | Compose-safe one-shot orchestration. | Long-running scheduler sidecar or hidden app startup side effect. | Inputs: `.env`, source config, target config, compose file; output: backup capture and due verification runs. |
| Target config | `deploy/mvp/` examples | Isolated restore-verification target binding and marker. | Production restore target or disconnected-profile claim. | Output: target TOML example and `cartulary.restore_verification_target.v1` marker JSON. |

Sprint 2 scheduling rules:

- Backup capture timer default MUST be every 6 hours.
- Restore-verification timer default MUST be daily.
- Systemd files MUST be examples/templates only and MUST NOT contain real secrets.
- Recovery scripts MUST use explicit config paths and service binding environment variables.
- The production app service MAY be stopped during capture only through explicit package script behavior.
- The restore-verification target MUST use a different config file, different Postgres binding, different object-store binding or bucket, and a restore-verification target marker.

### Sprint 3 - Restore Verification

Plan order: pin target-safety and artifact tests, then wire package due-runner execution against the isolated target.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Due-runner packaging | `deploy/mvp/docker-compose.yml` or `deploy/mvp/scripts/` | Package invocation of existing `restore-verify due`. | Manual-only verification or public API routes. | Inputs: source and target configs; output: verification summary. |
| Target safety tests | `cmd/operator` | Unsafe-target rejection for same config, same DSN, same object store, and missing or invalid marker. | Broad restore UI or arbitrary timestamp restore. | Output: fail-closed evidence. |
| Artifact checks | recovery tests | Restore-verification artifact proof and metadata transitions. | Publication evidence under Core 05. | Output: pass/fail state, verification basis digest, artifact key, SHA-256, and size. |

Sprint 3 restore-verification rules:

- Due-runner execution MUST select backups due by verification age or verification-basis change.
- A successful run MUST set `verification_state='verified'`, update `last_verified_restore_at`, and record the verification basis.
- A failed run MUST set `verification_state='failed'`, update `last_verified_restore_at`, and MUST NOT be represented as verified or ready.
- Verification MUST restore Postgres and object-store content from the same selected `backup_set`, rebuild projections, check evidence/blob invariants, and when incident data exists, open at least one incident and execute at least one built-in workbook query.
- The target environment MUST be isolated and MUST fail closed before mutation when the marker or storage separation checks fail.

### Sprint 4 - Smoke Gate

Plan order: add a public Make target only after the package recovery workflow is runnable. Update owner inputs, then run the repository generator; do not hand-edit generated task-surface outputs.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Public Make target | topology owner inputs | `standup-operational-recovery-smoke` command surface. | Hand-edited generated files. | Output: regenerated generated files through Make/generator-owned tooling. |
| Smoke script | `scripts/ci/` | End-to-end package recovery proof. | Broad `make check` or release evidence. | Output: retained summary and artifacts. |
| Route absence check | smoke/process tests | No public backup/restore API family. | Operator command removal. | Output: static and running-route absence evidence. |

Sprint 4 smoke assertions:

- The package can run backup capture and produce a successful retained `backup_set`.
- Latest metadata inspection sees a durable backup whose `consistency_point_at` is no older than 24 hours.
- The backup shows `retained_until >= created_at + 30 days`.
- The smoke can run `restore-verify due` against an isolated target and observe a passing verification artifact.
- The public route inventory still exposes no backup, restore, or restore-verification route family.
- `standup-package-smoke` remains package smoke only and MUST NOT be reclassified as backup/restore conformance evidence.

### Sprint 5 - Final Handoff

Plan order: update runbook and live tracker evidence, then run the narrowest sufficient validation; broaden only when runtime or harness changes require it.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Runbook update | `deploy/mvp/README.md` | Operator steps and troubleshooting for schedule install, capture, inspect, restore verification, and failure handling. | Normative config schema or broad deployment policy. | Output: actionable operational recovery instructions. |
| Tracker closure | this handoff file | Live sprint evidence, skipped checks, residual risks, and closure notes. | Prose-only completion claims. | Output: files changed, substantive edits, validation commands, run roots, skipped checks, and residual risk. |
| Finalizer evidence | harness artifacts | Retained-run maintenance. | Product verification. | Input: successful `make check` root when available; output: finalizer summary. |

Sprint 5 closure rules:

- Every completed sprint row MUST name validation evidence or a recorded skipped check with owner and impact.
- The final handoff update MUST list changed files, substantive edits, validation commands, results, skipped checks, and residual risks.
- If broad `make check` passes, `make agent-finalize RESULTS_DIR=<successful-check-run-root>` SHOULD be run and recorded.
- If broad `make check` is skipped, this tracker MUST record why narrower validation was sufficient.

## Planned Interfaces And Public Additions

The following additions are planned work, not current behavior claims:

| Interface or artifact | Required behavior | Owner |
| --- | --- | --- |
| `cartulary-operator backup capture` or equivalent | Deployment-local command that captures one retained `backup_set` under the current backup contract, requires active `deployment_admin`, emits structured JSON, and exposes no public route. | Core 01 backup contract, Core 04 deployment administration, and `cmd/operator`. |
| `cartulary.operator.backup_capture_result.v1` | Structured operator result with backup identity, consistency point, retention, restore anchors, artifact proof summary, and safe failure state. | Supporting operator contract; exact shape finalized in Sprint 1 tests. |
| `deploy/mvp/systemd/cartulary-backup.*` | Example service/timer that runs backup capture every 6 hours. | Supporting deployment package. |
| `deploy/mvp/systemd/cartulary-restore-verify.*` | Example service/timer that runs due restore verification daily. | Supporting deployment package. |
| `deploy/mvp/restore-verification-target.toml.example` | Example isolated target config using different source/target bindings. | Supporting deployment package under Core 04 root-binding rules. |
| `standup-operational-recovery-smoke` | Make-owned smoke target proving backup capture, retained metadata, due restore verification, isolated target safety, and no public route family. | `docs/testing-harness-nlspec.md` and task-surface owner inputs. |

## Definition Of Done

| Acceptance row | Required evidence | Blocking |
| --- | --- | --- |
| Scheduled backup capture | Systemd timer and package script run `backup capture` every 6 hours. | yes |
| Latest retained backup | Metadata inspection proves at least one successful retained `backup_set` is less than 24 hours old. | yes |
| Retention floor | Successful backup metadata and artifact proofs show at least 30 days retention. | yes |
| Encrypted backup storage | Filesystem-backed artifacts require `CARTULARY_RECOVERY_MASTER_KEY` and fail closed when missing or wrong. | yes |
| Restore verification cadence | Due runner is scheduled daily and proves backups are verified at least every 7 days or after basis changes. | yes |
| Isolated restored state | Restore verification uses a separated target and rejects unsafe targets before mutation. | yes |
| Workbook proof | Verification opens an incident and runs a built-in workbook query when restored incident data exists. | yes |
| Public route absence | Static and process evidence prove no public backup, restore, or restore-verification route family. | yes |
| Runbook | Package docs explain install, inspect, failure handling, and recovery verification. | yes |

## Commands

Planning and inspection:

```sh
make task-guide ROLE=feature-dev PHASE=phase10
make explain-phase PHASE=phase10
make explain-target TARGET=standup-package-smoke DETAIL=summary
```

Docs and generated-artifact validation:

```sh
make generated-artifact-policy-check
make json-shape-check
make lint-markdown
```

Phase 10 recovery validation:

```sh
make backend-unit
make backend-store
make backend-integration
make backend-process
make phase-slice PHASE=phase10
make service-backed-slice PHASE=phase10
```

Package validation:

```sh
make standup-package-smoke
make standup-operational-recovery-smoke
```

Broader closeout when runtime, harness, or package behavior changes:

```sh
make check
make agent-finalize RESULTS_DIR=<successful-check-run-root>
```

`standup-operational-recovery-smoke` is a planned new command surface. Add it through harness/topology owner inputs and generated outputs only.

## Research Support

Research reports are supporting inputs only. They do not create owner requirements.

| Report | Material use for this tracker |
| --- | --- |
| R01 Aurora incident response report | Not material to operational recovery mechanics beyond background on local-app incident response tooling. |
| R02 CRM/TEM DFIR report | Not material to operational recovery mechanics. |
| R03 Kanvas technical report | Supports treating lack of backup/versioning as a data-integrity risk for workbook-centered incident tooling. |
| R04 responsive browser spreadsheet UI memo | Not material to recovery mechanics beyond preserving workbook usability after restored-state proof. |
| R05 responsive interface design report | Not material to recovery mechanics. |
| R06 spreadsheet-of-doom DFIR report | Supports explicit audit, evidence, sensitivity, backup, and portability boundaries. |
| R07 spreadsheet-of-doom report | Supports least privilege, durable snapshots, and sanitized paths for sensitive cases. |
| R08 Handsontable React report | Not material to operational recovery mechanics. |
| R09 React Data Grid report | Not material to operational recovery mechanics beyond preserving workbook query behavior in restored-state checks. |

## Assumptions And Dependencies

- The MVP package remains an on-prem local stand-up package.
- Current filesystem-backed backup storage remains the first supported operational recovery target.
- `CARTULARY_RECOVERY_MASTER_KEY` is required for backup capture and restore-verification artifact access.
- Backup capture may stop the app container briefly through explicit package script behavior.
- Direct out-of-band DB or object-store writes during capture are unsupported.
- Restore verification uses a dedicated isolated target, not the production deployment.
- Managed-service backup storage remains future work unless Sprint 0 identifies a current owner requirement that already makes it necessary.
- Arbitrary operator-supplied timestamp PITR remains out of scope for the current profile.
- Generated files such as `tools/task_surface.generated.mk` MUST NOT be hand-edited.

## Current Characterization Snapshot

This snapshot is initial orientation for future Sprint 0 work. It does not mark Sprint 0 complete.

| Area | Current evidence | Gap or follow-up |
| --- | --- | --- |
| Core backup contract | Core 01 §12.1-§12.2 defines retained `backup_set`, same-point Postgres and object-store restore anchors, 24-hour latest-success freshness, 30-day retention, restore ordering, fail-closed missing artifacts, due restore verification, and SeaweedFS restore-verification artifact evidence. | Sprint 0 must confirm whether any owner text update is required after detailed implementation review. |
| Core security/deployment boundary | Core 04 §9.14 and §12.3.3 define AC-398 through AC-403, `roots.backup_storage`, encrypted filesystem-backed artifact envelopes, deployment-local operator-facing controls, and no public backup/restore route families. | Implementation must preserve the deployment-local boundary and avoid incident-scoped workflow leakage. |
| Domain boundary | `docs/domain.md` classifies Backup and Restore as deployment-local operator-facing recovery and forbids workbook route-family or incident workflow promotion. | Tracker wording and implementation names must preserve this boundary. |
| Current operator surface | Existing usage includes `backup-metadata latest`, `restore latest`, `restore-verify latest`, `restore-verify due`, `object-store init`, and object-store migration. | A package-ready backup capture command or equivalent operator surface is the central missing runtime interface. |
| Current recovery substrate | Existing recovery code includes Postgres/object-store artifact capture helpers, encrypted backup storage, durability checks, restore runner, restore-verification service, and restore-verification artifacts. | Need package-facing command and schedule wiring, not a new public recovery model. |
| MVP package | `deploy/mvp` persists `cartulary-backups` and includes `cartulary-operator`, but README says backup/restore conformance is not claimed because backup capture, retention, and restore-verification scheduling are separate requirements. | Sprints 1-4 must close the operational wiring and evidence gap. |
| Existing package smoke | `make standup-package-smoke` validates package runtime shape and explicitly is not backup/restore conformance evidence. | Add separate `standup-operational-recovery-smoke` target. |

## Live Execution Tracker

This section records implementation-support progress after the tracker is adopted. Rows start as `not-started`.

| Sprint | Status | Blocking work before completion | Evidence and notes |
| --- | --- | --- | --- |
| Sprint 0 | `complete` | None. | Characterization completed before implementation. Files changed: this tracker only. Evidence: Core 00 identifies Core 01 §12.1-§12.2 as owner for operational backup, coherent restore, retention floor, equivalence, and restore verification; Core 01 REQ-01-571 through REQ-01-578 define one retained `backup_set`, one `consistency_point_at`, `backup_attestation`, readable artifact/integrity proof, 24-hour freshness, 30-day retention, coherent same-set Postgres/object-store restore, projection rebuild, due restore verification, basis-change selection, workbook-query proof when incident data exists, and deployment-local separation from incident portability; Core 04 REQ-04-106 and AC-398 through AC-403 define `deployment_admin`, deployment-local operator-facing backup/restore controls, unsafe target preflight, no public `/api/v1/` or `/ws/v1/` backup/restore/restore-verification route families, `roots.backup_storage`, encrypted filesystem-backed artifact envelopes, and fail-closed missing-key behavior. `make task-guide ROLE=feature-dev PHASE=phase10`, `make explain-phase PHASE=phase10`, and `make explain-target TARGET=standup-package-smoke DETAIL=summary` passed; Phase 10 map reports complete coverage with public evidence in backend-unit, backend-store, backend-integration, backend-process, and browser-e2e-webserver-backed, and `standup-package-smoke` is helper-only/package-shape evidence with no phase coverage. Operator inspection confirmed existing `backup-metadata latest`, `restore latest`, `restore-verify latest`, `restore-verify due`, `object-store init`, and `object-store-migration run`; backup capture is the missing operator/package interface. Recovery module inspection confirmed existing encrypted filesystem backup storage, Postgres/object-store snapshot capture helpers, backup metadata store, durability catalog, restore runner, restore-verification due support, restore-verification artifacts, workbook probe, target marker, and no-public-route sentinel. MVP package inspection confirmed `cartulary-backups` persistent volume and README non-claim boundary. Research support remains rationale only: R03 supports backup/versioning risk, R06 supports audit/evidence/backup/portability boundaries, and R07 supports least privilege, durable snapshots, and sanitized paths. Core edit decision: no Core owner edits are required because Core 01/Core 04 already own the implementation behavior needed for this tracker. Validation skipped: no code validation beyond planning/inspection targets because Sprint 0 made no production implementation changes. Residual risk: implementation still needed for backup capture, scheduling, smoke target, and runbook. |
| Sprint 1 | `complete` | None. | Files changed: `internal/app/operator.go`, `cmd/operator/operator_phase10_test.go`, `tools/phase10_test_map.json`, generated `docs/testing/phase10_coverage_ledger.md`, generated `tools/scheduler_manifest.json`, generated `tools/execution_topology_render_index.json`, `scripts/test-print-target-plan.sh`, and this tracker. Substantive edits: added deployment-local `cartulary-operator backup capture` with required active `deployment_admin`, required backup capture quiescence proof, no public route surface, encrypted filesystem-backed backup storage, Postgres snapshot artifact capture, SeaweedFS/object-store snapshot plus operator-private manifest and redacted summary capture, one retained `backup_set` with one `consistency_point_at`, 30-day metadata and anchor retention, integrity/durability verification before result emission, structured `cartulary.operator.backup_capture_result.v1` output, redacted object-store metadata, and missing-key/wrong-key fail-closed coverage. Added selected E-10-01 process test `TestPhase10_E_10_01_DeploymentLocalOperatorCaptureBackupSet`; updated Phase 10 owner input and regenerated derived phase ledger/schedule artifacts through Make, not by hand. Validation passed: `make format` at `.cartulary/test-results/20260614T215201Z-p70688`; `make phase-ledgers` at `.cartulary/test-results/20260614T215213Z-p72120`; `make phase-schedules` at `.cartulary/test-results/20260614T215213Z-p72143`; `make backend-process` at `.cartulary/test-results/20260614T215223Z-p72647` with 36 tests passed; `make phase-ledger-drift` at `.cartulary/test-results/20260614T215332Z-p82047`; `make phase-schedule-drift` at `.cartulary/test-results/20260614T215332Z-p82050`; `make json-shape-check` at `.cartulary/test-results/20260614T215332Z-p82083`. Skipped checks: package scheduling and smoke were not run for Sprint 1 because Sprint 2/Sprint 4 own package timers/scripts and the operational-recovery smoke target; impact is package automation remains unproven until those sprints. Residual risk: first-time capture cannot detect a wrong recovery key before any encrypted artifact exists; wrong-key fail-closed behavior is proven against existing encrypted backup artifacts through metadata/durability access. |
| Sprint 2 | `complete` | None. | Files changed: added `deploy/mvp/scripts/backup-capture.sh`, `deploy/mvp/scripts/restore-verify-due.sh`, `deploy/mvp/systemd/cartulary-backup.service`, `deploy/mvp/systemd/cartulary-backup.timer`, `deploy/mvp/systemd/cartulary-restore-verify.service`, `deploy/mvp/systemd/cartulary-restore-verify.timer`, `deploy/mvp/restore-verification-target.toml.example`, `deploy/mvp/restore-verification-target.marker.json.example`; updated `deploy/mvp/.env.example` and this tracker. Substantive edits: backup script stops the app service explicitly, writes a `cartulary.backup_capture_quiescence_proof.v1` proof, runs `cartulary-operator backup capture` through the package image with explicit config and env-file bindings, and restarts the app; restore-verification script uses a separate target config, target Postgres DSN/service ref, target object-store service ref/bucket, target filesystem root, and `cartulary.restore_verification_target.v1` marker before running target migration, target object-store init, and `restore-verify due`; systemd examples are non-secret templates with backup every 6 hours and restore verification daily; no TOML schema broadening or public recovery routes were added. Validation passed: `make lint-shell` (exit 0; no run root emitted) and `make deployable-shape` (exit 0; no run root emitted). Skipped/deferred checks: package runtime proof of the scheduled recovery workflow is deferred to Sprint 4 `standup-operational-recovery-smoke` because that Make target does not exist until Sprint 4; impact is timers/scripts are static-package evidence only until the operational smoke runs them. Residual risk: restore-verification target database/bucket lifecycle is script-managed and still needs Sprint 3/Sprint 4 runtime proof against a live package topology. |
| Sprint 3 | `complete` | None. | Files changed: `cmd/operator/operator_phase10_test.go` and this tracker. Substantive edits: extended selected E-10-01 `restore-verify due` process evidence to reject same source/target config, same Postgres DSN, same object store, missing target marker, and invalid `cartulary.restore_verification_target.v1` marker before target mutation; retained existing due-runner proof that selects backups due by never-verified/stale-age/basis-change state, writes verified state with non-null `last_verified_restore_at`, records verification basis digest and restore-verification artifact key/SHA-256/size, restores Postgres and object-store content from the same `backup_set`, rebuilds projections, checks evidence/blob invariants, and runs the workbook probe when restored incident data exists. Validation passed: `make format` at `.cartulary/test-results/20260614T215912Z-p91849`; `make backend-process` at `.cartulary/test-results/20260614T215918Z-p93233` with 36 tests passed. Skipped/deferred checks: live package invocation of `deploy/mvp/scripts/restore-verify-due.sh` is deferred to Sprint 4 `standup-operational-recovery-smoke`; impact is operator/runtime due-runner behavior is proven now, while package-script orchestration remains smoke-gated. Residual risk: no unresolved Sprint 3 code blocker; package topology proof remains open for Sprint 4. |
| Sprint 4 | `complete` | None. | Files changed: `docs/testing-harness-nlspec.md`, `tools/execution_topology_manifest.json`, generated `tools/task_surface_manifest.json`, generated `tools/task_surface.generated.mk`, generated `tools/scheduler_manifest.json`, generated `tools/execution_topology_render_index.json`, added `scripts/ci/check-standup-operational-recovery-smoke.sh`, updated `deploy/mvp/scripts/backup-capture.sh`, updated `deploy/mvp/scripts/restore-verify-due.sh`, updated `deploy/mvp/.env.example`, and this tracker. Substantive edits: added Make-owned public `standup-operational-recovery-smoke` through harness/topology owner inputs, regenerated generated task-surface/scheduler outputs through `make generate`, added smoke coverage that builds/runs the MVP Compose package, captures one backup, inspects latest metadata, proves 24-hour latest freshness and 30-day retention, runs due restore verification against the isolated target config and marker, proves no public `/api/v1/` or `/ws/v1/` backup/restore/restore-verification route family, and retains JSON summary artifacts. Kept `standup-package-smoke` as package-shape evidence only. Package-script fixes discovered by the smoke: first failing run `.cartulary/test-results/20260614T220444Z-p6310` showed package scripts were joining a different Compose project; added explicit `CARTULARY_MVP_COMPOSE_PROJECT_NAME` support and smoke binding. Second failing run `.cartulary/test-results/20260614T220856Z-p16076` showed the bind-mounted quiescence proof was unreadable by the non-root container user; changed proof mode to read-only world-readable because it contains only local stop-state facts. Third failing run `.cartulary/test-results/20260614T220944Z-p23924` showed the restore target marker was unreadable by the non-root container user and the target database probe was noisy; changed marker mode to read-only world-readable, constrained `RESTORE_VERIFY_POSTGRES_DB` to a simple identifier, and fixed the existence query. Validation passed: `make generate` at `.cartulary/test-results/20260614T220359Z-p3703`; `make phase-schedules` at `.cartulary/test-results/20260614T220429Z-p5381`; `make task-surface-report TASK_SURFACE_REPORT_ARGS=--check` passed; `make json-shape-check` at `.cartulary/test-results/20260614T221118Z-p40043`; `make phase-schedule-drift` at `.cartulary/test-results/20260614T221118Z-p40002`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260614T221118Z-p40028`; `make lint-shell` passed after script fixes; final `make standup-operational-recovery-smoke` passed at `.cartulary/test-results/20260614T221042Z-p32092` with retained `backup-capture.json`, `latest-backup.json`, `restore-verify-due.json`, `public-route-absence.json`, and `standup-operational-recovery-summary.json`. Skipped checks: broad `make check` remains deferred to Sprint 5 closeout because Sprint 4 owner validation is limited to public target generation, shell safety, JSON shape/drift, and the new operational smoke. Residual risk: no unresolved Sprint 4 blocker; final runbook and phase/broader validation remain Sprint 5 work. |
| Sprint 5 | `complete` | None. | Files changed for Sprint 5: `deploy/mvp/README.md`, `docs/handoffs/operational-recovery-standup-handoff.md`, generated `tools/go_test_duration_baselines.json`, generated `tools/browser_e2e_duration_baselines.json`, generated `tools/service_backed_make_target_duration_baselines.json`, generated `tools/harness_smoke_duration_baselines.json`, and regenerated schedule/ledger artifacts touched by `make phase-schedules`, `make phase-ledgers`, and `make agent-finalize`. Substantive edits: updated the MVP runbook with deployment-local operational recovery prerequisites, manual backup capture, latest metadata inspection, manual due restore verification, systemd timer installation, package validation, and failure handling; kept the no-disconnected-profile and no-public-route-family boundaries explicit. Validation passed: `make lint-markdown` (exit 0, rerun after final tracker edits); fresh `make standup-operational-recovery-smoke` at `.cartulary/test-results/20260614T221315Z-p41712`; `make phase-slice PHASE=phase10` at `.cartulary/test-results/20260614T221345Z-p49333` with 39 tests passed; `make test-service-backed` at `.cartulary/test-results/20260614T221609Z-p96817`; `make go-test-duration-baselines RESULTS_DIR=.cartulary/test-results/20260614T221609Z-p96817 PRUNE_OBSERVED_PACKAGES=1` at `.cartulary/test-results/20260614T221859Z-p37843`; `make go-test-duration-baseline-coverage` at `.cartulary/test-results/20260614T221906Z-p37937`; `make phase-schedules` at `.cartulary/test-results/20260614T221946Z-p62951`; `make json-shape-check` at `.cartulary/test-results/20260614T221957Z-p63476` and rerun at `.cartulary/test-results/20260614T222345Z-p24597`; `make phase-schedule-drift` at `.cartulary/test-results/20260614T221957Z-p63483` and rerun at `.cartulary/test-results/20260614T222345Z-p24625`; `make go-test-duration-baseline-drift` at `.cartulary/test-results/20260614T221957Z-p63514`; `make generated-artifact-policy-check` at `.cartulary/test-results/20260614T222345Z-p24583`; final `make check` at `.cartulary/test-results/20260614T222001Z-p64017` with 161 work units, 855 tests, zero failures, zero missing; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260614T222001Z-p64017` at `.cartulary/test-results/20260614T222200Z-p21989`. Earlier broad-check failures were resolved through owner targets: `.cartulary/test-results/20260614T221448Z-p61562` failed because the new selected operator process test lacked a Go duration baseline, and `.cartulary/test-results/20260614T221911Z-p38072` failed because the baseline update made phase schedules stale. Skipped checks: none; broad `make check` and retained-run maintenance both passed. Residual risk: first-time backup capture still cannot distinguish a wrong recovery key from a new key before any encrypted backup artifact exists; wrong-key fail-closed behavior is covered once encrypted artifacts exist. This tracker does not claim disconnected-profile conformance. |

## Final Closure Evidence

Planning summary: Sprint 0 confirmed Core 01/Core 04 owner sufficiency and no Core edits were required. Sprints 1 through 4 implemented deployment-local backup capture, package-local scheduling, restore-verification target safety, and the Make-owned operational recovery smoke. Sprint 5 updated the MVP package runbook and closed validation with a successful broad check and finalizer run.

Changed files: `internal/app/operator.go`; `cmd/operator/operator_phase10_test.go`; `deploy/mvp/.env.example`; `deploy/mvp/README.md`; `deploy/mvp/scripts/backup-capture.sh`; `deploy/mvp/scripts/restore-verify-due.sh`; `deploy/mvp/systemd/cartulary-backup.service`; `deploy/mvp/systemd/cartulary-backup.timer`; `deploy/mvp/systemd/cartulary-restore-verify.service`; `deploy/mvp/systemd/cartulary-restore-verify.timer`; `deploy/mvp/restore-verification-target.toml.example`; `deploy/mvp/restore-verification-target.marker.json.example`; `scripts/ci/check-standup-operational-recovery-smoke.sh`; `scripts/test-print-target-plan.sh`; `docs/handoffs/operational-recovery-standup-handoff.md`; `docs/testing-harness-nlspec.md`; generated `docs/testing/phase10_coverage_ledger.md`; owner/generated harness files under `tools/phase10_test_map.json`, `tools/execution_topology_manifest.json`, `tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, `tools/go_test_duration_baselines.json`, `tools/browser_e2e_duration_baselines.json`, `tools/service_backed_make_target_duration_baselines.json`, and `tools/harness_smoke_duration_baselines.json`.

Substantive edits: added `cartulary-operator backup capture` with active `deployment_admin` authorization, quiescence proof, encrypted filesystem-backed backup storage, Postgres and object-store restore anchors, integrity/durability proof, redacted metadata, 30-day retention, latest-backup freshness support, and missing/wrong-key fail-closed coverage; added package backup and restore-verification scripts, non-secret systemd timers, isolated restore-verification target config and marker, target safety tests, public operational recovery smoke, and MVP runbook instructions. No public `/api/v1/` or `/ws/v1/` backup, restore, or restore-verification route family was added.

Validation commands and results: all required Sprint 0 through Sprint 5 validation is recorded in the live tracker rows above. Final closeout validation passed with `make lint-markdown`, `make standup-operational-recovery-smoke`, `make phase-slice PHASE=phase10`, `make test-service-backed`, duration-baseline refresh/coverage/drift targets, `make generated-artifact-policy-check`, `make json-shape-check`, `make phase-schedule-drift`, `make check`, and `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260614T222001Z-p64017`.

Skipped checks: none. The initial broad `make check` failures were not skipped; they were resolved by refreshing Go test duration baselines from a successful full service-backed run and regenerating phase schedules through Make-owned targets.

Blockers and residual risks: no blockers remain. Residual risk is limited to first-time wrong-key ambiguity before any encrypted backup artifact exists; missing-key and wrong-key fail-closed behavior after encrypted artifacts exist is covered. Operational recovery remains MVP on-prem package evidence only and is not disconnected-profile conformance.

Retained-run maintenance: `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260614T222001Z-p64017` passed at `.cartulary/test-results/20260614T222200Z-p21989`, validated the successful `make check` retained run, refreshed duration baselines, and reported `run_checks=pass`.
