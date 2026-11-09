# reference_data Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

Current invocation: user-authorized remediation of the final LC review, 2026-11-09. **LC-F4 and LC-F5 are RESOLVED; LC-R1–LC-R5 are DONE for scoped engineering acceptance. Human specification projection-fidelity review remains PENDING, as confirmed by the user.** The [review remediation checkpoints](#lc-review-remediation-checkpoints) and final LC-R5 handoff supersede earlier unconditional closure claims. The complete-owner audit accepts all 101 required rows on final executable source `sha256:4e1cb8113d7b0f36e9d96ea95c9610d67a28bb2807c922f39fd68fab722b7f1f`. Earlier LC-P0–P5 and superseded review-freeze results remain historical evidence. Incoming staged work and immutable releases remain preserved. Earlier document-only authorization statements describe their historical invocations, not this implementation request.

RI-P0 through RI-P6 and RC-WS1 through RC-WS7 remain historical completions on their recorded source and scoped evidence. The MITRE-license decision remains closed and the four production artifacts remain delivered. RC's 130-row portfolio is not complete-owner acceptance or qualification of subsequently changed source. LC does not reopen that licensing decision or rewrite earlier releases, provenance or validation records.

This is the controlling tracker for Reference Data refactoring. Historical I8: the user explicitly authorized implementation of I8's F-33/34 remediation through S-37–40 on 2026-11-08. S-37–40 are DONE and F-33/F-34 are closed by E-I8-S40; sequential execution checkpoints and final qualification are recorded in §22. I6 / S-28–32 and I7 / S-33–36 remain historical completions on §§20–21's evidence; their passes are not current I8 qualification. Execution is sequential, with each completion checkpoint recorded before its successor starts.

- Supplied output and admitted artifact: `/home/jochi/code/cartulary/docs/handoffs/reference_data-module-refactor-tracker.md`; permitted location `docs/handoffs`; no replacement tracker.
- Target established from this tracker's existing identity: `internal/modules/reference_data`, a directory under canonical repository root `/home/jochi/code/cartulary`; normalized label `reference-data`. Preserve the established underscore filename.
- Artifact mode: `normal` for the bounded LC plan; inspection depth and future characterization gates are explicit in the LC inventory. Planning completeness does not mean an exhaustive subsystem correctness/dead-code audit or production-readiness certification. Sections 22, 23 and RC retain their historical decisions, evidence and acceptance limits; LC controls current planning.
- Planning baseline I8-00: clean `main`, HEAD `e2291271a502033b85c5e2d6aace2609bc8c8210`. Target/output and parent components are not symlinks; the output is a single-link regular file with the expected canonical target identity. E-I8-01 records the content/identity baseline and final comparison locations.
- Historical I8 scope: implement F-33/34 through S-37–40 and V-46–48, with tracker updates at each exit; preserve current release bytes, immutable historical artifacts, producer fixture signing, public HTTP/storage formats and Base readiness. The user chose retirement of the one-shot Base reissue command; a general release builder is deferred until a concrete release requires it. Source acquisition, schema/data migration, deployment and release certification remain outside scope. V-49 belongs to the completed planning-document invocation.
- Historical S-24–27 completed fatal-finalization remediation (§19); F-21/F-24/F-25 remain closed on that evidence, with F-22/F-23 also closed. S-23's premature closure remains superseded. Earlier execution permissions and test passes are historical, not authorization or qualification for this iteration.
- Framework `docs/handoffs/cartulary_modular_refactor_planning_framework.md` and the refactor-tracker format guided the historical planning update, whose write boundary was this tracker only. The later user-authorized LC implementation and product qualification are recorded in the execution amendment.

Implementation baseline I8-03: `main` at `e2291271a502033b85c5e2d6aace2609bc8c8210`, with only this tracker already modified. `/tmp/reference-i8-implementation-xt30lmqv/baseline.json` records content/identity for all 8,942 tracked/nonignored files, the index and 497 immutable Base/producer/fixture artifacts. `tracker-before.md`, `preexisting.diff` and `index-entries.txt` preserve the incoming edits and staged state. The incoming tracker SHA-256 is `2ce9d34c7433dbdea6dbf29ebd31178d85f7bcaff9e21ecaf547c30bc9ee3c5e`. Ignored build/cache/results outputs are managed execution outputs, not part of the immutable preservation claim. The following table retains the earlier document-session baseline.

| Baseline item | I8 starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Branch and commit | Clean `main` at `e2291271a502033b85c5e2d6aace2609bc8c8210`; `/tmp/reference-i8-docs-e5p53yna/baseline.json` | §22 document checkpoint and final preservation comparison | This commit contains the prior I6/I7 implementation; its historical run artifacts were not requalified. |
| Staged/index state | Clean staged entries; raw index hash and file identity captured in the same baseline | Must match baseline, excluding no staged entry | Read-only Git uses `--no-optional-locks`. |
| Unstaged/untracked content | Clean; all 8,942 tracked/nonignored entries fingerprinted by content and identity | Only this tracker may differ | No untracked entry existed at baseline. |
| Relevant ignored outputs | `ignored-baseline.json`: 2,525 entries covering frontend installation/toolchain, Corepack, installed lockfile and existing static cache; workspace lock captured separately | Compare after Markdown validation; all results/new static cache redirected outside the repository | Not a hash of every unrelated ignored run, build or dependency tree. |
| Tracker identity/content | SHA-256 `ab41f0a7ea4d9e12ad6b71eef088e2eab93aa42940947816ccf90f2c5c899d04`; original text in `tracker-before.md`; device/inode/mode/link count in baseline | Recheck before writing; reconcile rather than overwrite concurrent edits | Retained comparison directory is `/tmp/reference-i8-docs-e5p53yna/`. |

The following table records the earlier document-session baseline. The implementation baseline I6-03 and final preservation comparison are in §20; the tracker was already dirty when implementation began.

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Branch and commit | `main` / `a8f2fe2d84eecf88ced265b34327e0a8c7d4c8df`; I6-00 | Final session comparison in §20 | Prior sessions used `3cc655473…`; their passes are not current-source evidence. |
| Staged/index state | Clean; index content, mode and file identity in `/tmp/reference-i6-docs-iigz5fsh/baseline.json` | Final session comparison in §20 | Read-only Git uses `--no-optional-locks`. |
| Unstaged and untracked content | Clean; all 8,936 tracked/nonignored entries fingerprinted, with regular-file hashes and file identities in the same baseline | Only the admitted tracker may differ | Enumeration includes nonignored untracked files; none existed at baseline. |
| Relevant ignored outputs | `/tmp/reference-i6-docs-iigz5fsh/ignored-baseline.json`: frontend installation/toolchain, Corepack, static-analysis cache, installed lockfile and workspace lock | Final session comparison in §20 | Not a snapshot of every ignored product run or dependency file. Lint results/static cache are redirected outside the repository; installation is prohibited. |
| Tracker content and identity | Initial SHA-256 `a9c17bf05d5c050db071ebeca9cd9b7bd84825aaa61927e8e900efe429092966`; device/inode/mode/link count in baseline; initial text in `tracker-before.md` beside it | Identity/content rechecked before writing; final comparison in §20 | Re-read and reconcile on any concurrent change; never overwrite unseen edits. |

The baseline records hashes and metadata, not repository contents or secrets. The retained original tracker supports historical reconciliation. Historical baseline and authority records follow; their session-specific statements do not override the current scope above.

Baseline E-00: branch `main`, commit `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`; initially clean staged, unstaged and untracked state. The tracker was absent. The target and output-parent path components were inspected for symlink/identity hazards. There were 240 tracked regular Go target files in eight directories, including 93 test files; none were excluded. There was no prior tracker handoff history to preserve. The requested directory is distinct from the normalized label: no rename is planned.

At historical S-00, the index SHA-256 baseline was `18ea56f64fc1a96e7b7d1ca64403cb9017dc13ada3651cdb73e6949225fdff20`. A read-only fingerprint of 8,865 tracked/nonignored files included path, mode, device, inode, link count, size and content hash; its aggregate was `ab7967c029b7a4c47453a83acbe89ff7e11e2fe7a193c19dd4763976d2d76f41`. That historical final comparison excluded only the newly created tracker. This is scope accounting, not runtime evidence; ignored caches are not covered by that aggregate.

Authority order is: adopted subsystem NLSpecs for their named subsystem; Core 00–04 for implementation conformance; Core 05 only for claim-bearing timed or fixture-sensitive publication; domain vocabulary and implementation-support guides; live implementation/tests; historical plans/frameworks as evidence only. Typed projections remain downstream. No applicable-owner contradiction was established. If a later inspection establishes one, record `BLOCKED: owner contradiction` and stop the affected design decision rather than choosing an owner.

| Evidence ID | Inspected authority or support source | Scope and posture |
| --- | --- | --- |
| O-01 | `docs/reference-pack-subsystem-nlspec.md`, status/scope, current adoption, canonical formats, verification/lifecycle, consumer, route delegation and portability/recovery passages (§§1, 4, 9–12, 14, 19–25, 29) | Revalidated for I6: adopted Reference Pack 0.4.0 / profile major 3; pack contract v2, current immutable v2 sets, current ATTACK/Windows/Entra/VERIS content profiles. Shared manifest/schema namespaces do not all become v2. Adoption is authority evidence, separate from implementation evidence. |
| O-02 | `docs/spec/00_document_set_status_and_precedence.md`, precedence and coordinated adoption entries | Revalidated adoption includes native ATTACK 2026-11-06 and Entra/VERIS 2026-11-07. No authority inferred from the target directory name. |
| O-03 | `docs/spec/01_architecture_storage_and_view_contracts.md`, §2, Reference Pack public contract §17.4 and coordinated amendment | Logical module boundaries; common HTTP, paging, idempotency, Jobs and route-family authorization. REQ-01-480/481 distinguish current family authorization order and administrative resource shape. |
| O-04 | `docs/spec/02_domain_model_schema_and_history.md`, indicator identity/participant and §11 registry passages | Reference Data owns registry algorithms; Indicators owns incident identity and mutations; caller acquires registry guards before incident/record locks. |
| O-05 | `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`, coordinated Reference Pack amendment | Pack activation is deployment scoped and does not automatically mutate incident records. Workbook interaction remains outside this target. |
| O-06 | `docs/spec/04_security_deployment_and_conformance.md`, §4.1 and Reference Pack adoption/configuration passages | Offline hostile-pack verification, attribution, current authorization, trusted-clock/configuration requirements. Core 05 publication is not invoked by this planning task. |
| O-07 | `docs/domain.md`, §§5, 7, 10, 10.2, 11, 14 and owner-navigation passages | Supporting Reference Data context; anti-corruption boundary with incident entities; Base readiness versus optional enrichment; vocabulary guidance below adopted owners. |
| E-01 | `AGENTS.md`, `internal/AGENTS.md`, `tools/AGENTS.md`; `docs/handoffs/cartulary_modular_refactor_planning_framework.md` | Procedure and planning template. Applicable nested-instruction enumeration found no additional target/docs subtree instructions. Framework workflow names are adapted to live seams. |
| E-02 | Historical E-00 target files in the original §2 table | Historical direct reads of declarations/imports and relevant bodies/assertions; critical seams also read in full. This was a file-level planning inventory, not a correctness audit. Current I6 depth is separately accounted above; E-02 does not establish current direct inspection of all 259 files. |
| E-03 | `internal/app/referenceassembly/administration.go`, `administration_test.go`, `storage.go`, `artifact_storage.go`, `verification_storage.go`, `operator_admission.go`, `retained_admission.go`, `registry_transactions.go`, `registry_usage.go`, `recovery_storage.go`, `telemetry.go`; `internal/app/server/runtime_assembly.go`, `runtime_routes.go`, `module_settings.go`; `internal/app/configassembly/reference_pack.go` | Assembly/constructor and selected concrete adapter bodies. Worker registration, physical roots, source-owned usage and startup orchestration remain in application assembly. |
| E-04 | `internal/app/extensionassembly/reference_pack_jobs.go`; `internal/app/operator/operator_reference_pack.go`; `internal/platform/httpapi/operation_catalog.go`, `httpapi.go`; `internal/platform/httpauth/httpauth.go` | Exact finalization adapter, route binding/dispatch and relevant auth/operator paths. Supports F-01; unknown-family/method responses are not assumed protected by a nonexistent outer admin check. |
| E-05 | `internal/modules/indicators/application.go`, `internal/identity/identity.go`; `internal/app/revisionassembly/revisions.go`; `internal/modules/reporting/snapshot_capture.go`, `store.go`, `portable_references.go` | Exact caller portions: guarded registry evaluation, immutable reporting binding and revision provider composition. These callers do not transfer their owning domain to Reference Data. |
| E-06 | `internal/modules/incidentbundles/source.go`, `worker_service.go`; `internal/app/incidentportabilityassembly/catalog.go`; `internal/app/recoveryassembly/state_catalog.go`, `vnext.go`; `internal/testutil/appsupport/workbook.go`, `current_restore.go` | Exact transaction/import/export and backup-retention caller portions. Broader downstream implementation audit is not claimed. |
| E-07 | `apps/web/src/services/referencePacks.ts`, `referencePacks.test.ts`; `apps/web/src/app/referencePackAdminModel.ts`, `referencePackAdminModel.node.test.ts`, `referencePackAdminController.ts`, `referencePackAdminController.node.test.ts`, `useReferencePackAdmin.ts`, `ReferencePackAdminPanel.tsx`, `ReferencePackAdminPanel.test.tsx`; `apps/web/e2e/reference-pack.spec.ts`, `measurement/reference-pack-progress.spec.ts` | Transport bodies, controller/model/presentation seams and test declarations/selected assertions. No rendered UI review or browser execution. |
| E-08 | `contracts/openapi/cartulary.openapi.yaml` (nine operation objects); `tools/generated_artifact_policy.json`; `tools/execution_topology_manifest.json`; `tools/backend_module_boundaries.json`; `tools/test_support_inventory.json` | Downstream generated contract and authored policies, read only. OpenAPI is explicitly generated, not an editable source of truth. |
| E-09 | `tools/test_families/module.reference_data.json`; `contracts/verification/owners/module.reference_data.json`; `Makefile`; `tools/task_surface.generated.mk`, `task_surface.runtime.generated.mk`, `task_surface.help-all.generated.json`; relevant harness diagnostics/execution/Markdown scripts | Row ownership, command inputs, output/side-effect policy and invocation mechanics. Catalog accounting is not runtime architecture or passing test evidence. |

Skills used as execution guidance: `/mnt/c/Users/jochi/.codex/skills/refactor-tracker/SKILL.md` and its tracker-format reference; `.agents/skills/cartulary-test-ops/SKILL.md` and command recipes. Historical table layouts and IDs are retained; current planning additions follow the skill's evidence and readiness requirements. No external browsing was needed for repository inspection.

## 2. Current-State Repository Inventory

The RI implementation supersedes the I8 layout below: source producers now live
in four independent `referencepacks/` modules, the neutral format is a fifth
module, and the application retains generic consumption and lifecycle ownership.
Sections 23 and RC record the adopted contracts, artifact selections and prior qualification; LC records the current inventory and future cleanup gates.
The following enumerations remain historical snapshots of their named slices.

### I8 enumeration and bounded inspection

E-I8-01 enumerates all **265 tracked regular Go files, including 109 test files, in eight directories** under `internal/modules/reference_data`; no nonignored untracked, symlink, submodule or generated target entry was found. Directory counts are root 121, `contractadmission` 6, `internal/packformat` 81, `internal/packstate` 2, `producer` 43, `recoverycontribution` 2, `testsupport` 9 and `testsupport/fixturehttp` 1. Ignored caches/build/results, vendor trees and external paths are excluded from the target universe.

S-38 removes only `producer/base.go`: the implementation inventory is now **264 Go files, including 109 test files**, with 42 files under `producer`. All other directory counts remain unchanged. The removed source remains accounted for in the pre-removal inventory and I8-03 baseline. The following inspection table describes the planning baseline; E-I8-S38/39 records its exact implementation delta.

The complete per-path accounting is `/tmp/reference-i8-docs-e5p53yna/target-inventory.json`: **7 directly inspected and 258 metadata-only** entries for this bounded I8 change surface. Direct inspection means relevant bodies/assertions were opened, not that every path in the file was audited. Other earlier exploratory reads do not expand this claim. Historical inventories are preserved below; their inspection depth is not promoted to current evidence.

| Path / kind, relative to target | Disposition / responsibility | Callers, dependencies and tests | Contract / owner / risk / evidence |
| --- | --- | --- | --- |
| `producer/base.go` | Directly inspected; authored one-shot previous-release reader and fixed-identity reissue | `producer/command.go`; private format validators, canonical JSON and filesystem; no dedicated reissue test found | C-18; offline Reference Data producer; medium; F-33 / E-I8-03 |
| `producer/command.go`, `producer/command_test.go` | Directly inspected; authored CLI dispatch and bounded subprocess assertions | Thin tools executable; current assembly/audit implementations; existing CLI row | C-18/19; offline producer/test owner; medium; E-I8-03/05 |
| `retired_format_fixture_test.go`, `verification_engine_test.go` | Directly inspected; authored predecessor fixture options and live verifier rejection assertions | Only one fixture call site, looping unsigned/signed; `addZipFile` has other callers | C-13/19; owner-local test support; low; F-34 / E-I8-04 |
| `base_release.go`, `consumer_test.go` | Directly inspected; current release loading and Base-consumer assertions | Generated current release artifact; closed bindings, manifest/content validation and nine indicator types | C-12/15; runtime owner retained; high preservation value; E-I8-02/04 |
| Every other current target path in `target-inventory.json` | Metadata-only; no new movement or correctness conclusion | Historical responsibility labels remain navigation only | No required I8 movement source is pending; broader audit is not claimed |

Direct expansion outside the target covers `tools/reference-packs/witness.mjs`, `package.sh`, the CLI executable, owner test routing, Make/Markdown invocation files and selected owner clauses. Repository-wide symbol searches found the reissue function, command branch and witness allowlist entry; they do not prove absence of external CLI users. The deliberate removal has no compatibility alias. S-37 revalidated the affected sources/callers and executable baseline. No module-wide audit is claimed.

The original inventory below is the E-00 snapshot: all 240 files, with no exclusions. Paths in its first column are relative to `internal/modules/reference_data/`. Current execution additions are recorded in the following delta table; unchanged historical line counts are not current source locations. Other filenames matching an inventory path use the same target root (including `internal/packformat/...`). Other paths beginning `internal/modules/`, `internal/app/`, `internal/platform/`, `internal/testutil/`, `internal/gen/`, `tools/` or `contracts/` are repository relative.

### I6 planning enumeration and inspection depth

The I6 planning universe was every tracked and nonignored untracked entry under the literal target: **259 regular Go files, 105 test files, eight directories**; no symlink, submodule or generated target entry was found. Ignored caches/build/results, vendor content and external trees were outside this universe. The complete historical path-by-path disposition and identity record is `/tmp/reference-i6-docs-iigz5fsh/target-inventory.json`.

There are **15 directly inspected files and 244 metadata-only files** for this bounded planning pass; no enumerated entry is silently omitted. Direct inspection covers the files listed below. Every other current file, including new VERIS implementation, is metadata-only in this session even when historical E-02 recorded an earlier direct read. No whole-module behavioral revalidation is claimed. Required producer movement sources have been inspected; S-28 requires future executable characterization, not another unbounded inventory.

| Path / kind, relative to target | Accounting / responsibility | Callers, dependencies and tests | Contract / owner / risk / evidence |
| --- | --- | --- | --- |
| `producer/audit.go`, `candidate.go`, `package.go` | Directly inspected; active proposal audit and adopted packaging paths | Producer CLI, source/candidate tests; pinned source acquisition, canonical JSON and current schema projections | C-18; offline Reference Data producer; high; E-I6-02/03 |
| `producer/windows_contract.go` | Directly inspected; live Windows contracts use `CompileCandidate` | Windows decode/schema initialization and online producer tests; producer-local schema library | C-18; offline producer; medium; E-I6-04 |
| `producer/candidate_test.go`, `source_test.go`, `attack_projection_test.go`, `ledger_test.go` | Directly inspected; proposal expectations, source integrity, runtime exclusion and ledger assertions | Embedded producer inputs; exact catalog selectors in §8 | C-18/19; test owner; medium; E-I6-02/05 |
| `producer/envelope.go` | Directly inspected; fixture/release signing policies and shared envelope primitives | Multiple current content profiles and cross-owner fixtures | C-18; keep; E-I6-04; test-only callers do not imply dead code |
| `internal/packformat/shape.go`, `content.go`, `attack_native_test.go` | Directly inspected; current schema/content validation and native assertions | Generated current contract projection, canonical content and existing native-profile selector | C-13/18/19; private format owner; high; E-I6-03/05 |
| `portable_execution.go`, `verification_lifetime.go`, `attack_native_integration_test.go` | Directly inspected; historical lifecycle boundary and authentic canonical-artifact consumption | Existing lifetime owners and actual HTTP/consumer fixture path | C-10/17/18; runtime unchanged; E-I6-01/05 |
| Every other current path in the retained inventory | Metadata-only; existing responsibility labels remain historical | No new movement or correctness conclusion for these files | Outside the bounded change surface, not excluded from enumeration |

Relative to the 240-file historical table, these 20 paths now exist:

- `contractadmission/content_profile_test.go`
- `http_admission_integration_test.go`
- `http_query.go`
- `http_security_test.go`
- `internal/packformat/diagnostic_helpers_test.go`
- `internal/packformat/veris.go`
- `internal/packformat/veris_test.go`
- `object_collector_test.go`
- `portable_cleanup_integration_test.go`
- `producer/veris_package.go`
- `producer/veris_projection.go`
- `producer/veris_source.go`
- `producer/veris_source_test.go`
- `refresh_admission_matrix_integration_test.go`
- `testsupport/veris_fixture.go`
- `verification_lifetime.go`
- `verification_lifetime_test.go`
- `veris_lifecycle_integration_test.go`
- `workspace_lifetime_integration_test.go`
- `workspace_recovery_integration_test.go`

Historical `contractadmission/entra_test.go` is absent; its current replacement is `contractadmission/content_profile_test.go`. The 20 additions minus this one historical path account for the 259-file total. These are changes since the historical inventory, not files created by this document session.

Current direct caller expansion includes `tools/reference-packs/main.go`, `source.sh`, `witness.mjs`, the ATT&CK producer input embed and selected immutable JSON inputs, and authored test routing. Text searches identified references; exact file reads established the relevant CLI and Windows execution paths. External consumers of the internal producer CLI/report are not exhaustively known; the chosen clean cutover deliberately offers no proposal compatibility aliases.

### Execution inventory delta

I6 adds `producer/attack_inputs.go`, `schema.go`, `command.go` and their focused tests, `producer/audit_report_test.go`, and `internal/packformat/attack_source_contract_test.go`; removes `producer/candidate.go` and `candidate_test.go`; and moves tools command composition into the existing offline producer package. The temporary sibling `producercli` package was removed during S-32 boundary qualification. Final inventory and preservation are recorded in §20; the earlier inventory below remains historical.

| Path (repository relative) | Role / boundary | Validation / owner |
| --- | --- | --- |
| `internal/modules/reference_data/http_query.go` | Owner-local closed non-list query contract; no global singleton-policy change. | S-06; existing request unit row and new query integration row. |
| `internal/modules/reference_data/http_admission_integration_test.go` | Server-composed family authorization, exact routing and query regressions. | Three cataloged Reference Data integration rows; canonical fixtures only. |
| `internal/modules/reference_data/http_security_test.go` | Admission body/application/session-effect probes under existing request unit selector. | Reference Data unit row; no independent uncataloged top-level test. |
| `internal/platform/httpapi/family_routes.go` | Catalog-derived family routing and owner-supplied policy composition; no Reference Data import. | Shared runtime/catalog/parity checks. |
| `internal/platform/httpapi/family_routes_test.go` | Registration rejection/atomicity, boundary, method and metadata probes under existing registry selector. | `platform.openapi.unit.runtime_profile_catalog`. |

Historical target inventory after the preceding transport implementation: **243 Go files, 95 test files**, including all 240 original files; no moves or deletions. S-10 added `internal/packformat/diagnostic_helpers_test.go`, making the S-13 inventory **244 Go files, 96 test files**; existing files are retained. S-14–18 adds six module files, making that historical inventory **250 Go files, 101 test files**. The new file contains only relocated diagnostic composition helpers and is directly inspected/tested. The two platform additions are outside that target total. The preceding transport iteration changed owner transport/API comments/tests, shared HTTP/auth plumbing/tests, authored contracts/catalog, their Make-generated projections, owner documentation and this tracker. S-09–13 deltas are in §16; that iteration's deltas and preservation evidence are in §17. It required no producer, SQL, stored-format, UI implementation or dependency-lockfile edit; current planning scope is separately defined in §20.

### Historical file inventory

Depth `D` means the file was directly opened and its imports, declarations, relevant implementation and/or assertions inspected; it does not mean every execution path was audited. Public-symbol cells are representative declarations, not a promised stable exported API: exported methods on private receivers are private implementation surface. Go's `internal/packformat` and `internal/packstate` exports are owner-private.

Caller/test cells marked `lexical` are conservative identifier/import cross-references from exact files, not proven call graphs or coverage. Two examples are shown where available; interfaces and callback dispatch also reach the contract groups in §4. A test file's own assertions are its primary evidence. Test helpers with no top-level `Test` entry are executed through parent tests, not separate harness rows. Dependency prefixes abbreviate the repository module prefix only; `std` means Go standard-library imports. Contract cells name affected/consumed surfaces, never generated files to edit.

| Directory | Files | Test files | Disposition |
| --- | --- | --- | --- |
| target root | 110 | 44 | Application, consumer, HTTP, persistence and integration seams; keep ownership pending F-01 correction. |
| `contractadmission` | 6 | 4 | Read-only current-contract admission; keep. |
| `internal/packformat` | 77 | 35 | Private canonical verification/format semantics; keep. |
| `internal/packstate` | 2 | 1 | Pure lifecycle state transitions; keep. |
| `producer` | 34 | 8 | Offline source producers; preserve runtime exclusion. |
| `recoverycontribution` | 2 | 0 | Owner recovery declarations; keep. |
| `testsupport` | 8 | 1 | Owner-local fixture support; keep. |
| `testsupport/fixturehttp` | 1 | 0 | Test-only import adapter; keep. |
| **Total** | **240** | **93** | **147 other Go files, including test-support sources; no generated target files identified.** |

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `administration.go` | Administrative version DTO and transport-facing application interface; separates current health from successful metadata | `AdministrativeVersion`, `AdministrativeDependency`, `AdministrativeApplication` | lexical: `internal/app/referenceassembly/administration.go`, `internal/app/server/runtime_assembly.go`; 6 total candidate files | std only | lexical: `api_test.go`, `lifecycle_fixture_integration_test.go` | C-01–09, C-10: HTTP/Jobs projections | Reference Data | high | D; 69 lines; keep owner; §3 seam diagnosis |
| `administration_repository.go` | Owner SQL reads and scans administrative version resources | `ListVersions`, `GetVersion` | lexical: `administration.go`, `administrative_actions.go`; 3 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/modules/reference_data/internal/packstate`, `pgx/v5`; std | lexical: `base_reconciliation_integration_test.go`, `coordinator_admission_integration_test.go`; 4 candidate test files | C-01–09, C-10: HTTP/Jobs projections | Reference Data | high | D; 101 lines; keep owner; §3 seam diagnosis |
| `administrative_actions.go` | Exact activate, disable, remove and source-owner registry compatibility checks in publication transaction | `RegistryUsageReader`, `Activate`, `Disable`, `Remove` | lexical: `internal/app/referenceassembly/object_collection.go`, `internal/app/referenceassembly/verification_storage.go`; 9 total candidate files | `internal/modules/reference_data/internal/packstate`, `internal/platform/canonicaljson`, `uuid`, `pgx/v5`; std | lexical: `base_atomicity_integration_test.go`, `base_reconciliation_integration_test.go`; 10 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 316 lines; keep owner; §3 seam diagnosis |
| `api_test.go` | Assertions: Request Validation Normalization And Closed Registries; Support Reference Pack List Query Uses Shared List Query Scope; Support Reference Pack List Filter Applies Search And Exact Filters | `TestRequestValidationNormalizationAndClosedRegistries_Unit`, `TestSupportReferencePackListQueryUsesSharedListQueryScope`, `TestSupportReferencePackListFilterAppliesSearchAndExactFilters` | Go test runner; helper calls within owner tests | `internal/platform/httpapi`; std | This file's assertions/helpers; C-19 | C-01–09, C-10: HTTP/Jobs projections | Reference Data | medium | D; 198 lines; test/support, not runtime architecture |
| `attack_endpoint_verification_test.go` | Assertions: Native ATTACKSigned Endpoint Semantics | `TestNativeATTACKSignedEndpointSemantics_Unit` | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`, `internal/modules/reference_data/producer`, `internal/testutil/suiteservices`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 100 lines; test/support, not runtime architecture |
| `attack_native_integration_test.go` | Assertions: Authentic ATTACKProduction Lifecycle | `TestAuthenticATTACKProductionLifecycle_Integration` | Go test runner; helper calls within owner tests | `internal/app/referenceassembly`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/modules/reference_data/producer`, `internal/modules/reference_data/testsupport`, `internal/platform/pagination`, `internal/testutil/appsupport`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 100 lines; test/support, not runtime architecture |
| `attestation.go` | Retains attributable verification evidence and successful/failed attempt facts | Package-private: `appendPackAttestationTx` | lexical: `administrative_actions.go`, `base_release.go`; 6 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `uuid`, `pgx/v5`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 149 lines; keep owner; §3 seam diagnosis |
| `audit.go` | Appends pack administrative audit facts through shared audit substrate | Package-private: `appendPackAuditTx` | lexical: `attestation.go`, `coordinator.go`; 6 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/administrativeaudit`, `uuid`, `pgx/v5`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 66 lines; keep owner; §3 seam diagnosis |
| `base_atomicity_integration_test.go` | Assertions: Base Reconciliation Atomic Publication; Current Retained Structure Across Consumers And Recovery | `TestBaseReconciliationAtomicPublication_Integration`, `BeginTx`, `QueryRow`, `Scan`, `Exec`, `Commit`; +1 declarations | Go test runner; helper calls within owner tests | `internal/modules/reference_data/contractadmission`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `internal/platform/extensionstore`, `internal/platform/jobs`, `internal/platform/postgres`, `internal/testutil/pgtest`, `uuid`, `pgx/v5`, `pgx/v5/pgconn`, `pgx/v5/pgxpool`; std | This file's assertions/helpers; C-19 | C-15, C-17: current contract/readiness | Reference Data | medium | D; 498 lines; test/support, not runtime architecture |
| `base_reconciliation.go` | Inspects release reconciliation before selecting guarded changes | Package-private: `inspectReconciliation`, `guard` | lexical: `administrative_actions.go`, `base_release.go`; 8 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/modules/reference_data/internal/packstate`, `internal/platform/postgres`, `pgx/v5`; std | lexical: `canonical_persistence_integration_test.go`, `concurrent_publication_integration_test.go`; 9 candidate test files | C-15, C-17: current contract/readiness | Reference Data | high | D; 131 lines; keep owner; §3 seam diagnosis |
| `base_reconciliation_integration_test.go` | Assertions: Base Reconciliation Rechecks Selected Bytes Before Publication | `TestBaseReconciliationRechecksSelectedBytesBeforePublication_Integration`, `OpenPublished` | Go test runner; helper calls within owner tests | `internal/platform/jobs`, `uuid`; std | This file's assertions/helpers; C-19 | C-15, C-17: current contract/readiness | Reference Data | medium | D; 140 lines; test/support, not runtime architecture |
| `base_release.go` | Loads release-bound canonical Base registries and atomically reconciles their retained objects and set | `BaseReleaseOptions`, `Open`, `Append`, `ReconcileBaseRelease` | lexical: `internal/app/server/runtime_assembly.go`, `administrative_actions.go`; 4 total candidate files | `internal/gen/contractreferencepacks`, `internal/modules/reference_data/contractadmission`, `internal/modules/reference_data/internal/packformat`, `internal/modules/reference_data/internal/packstate`, `internal/platform/canonicaljson`, `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `attack_endpoint_verification_test.go`, `base_atomicity_integration_test.go`; 14 candidate test files | C-15, C-17: current contract/readiness; reads `internal/gen/contractreferencepacks` | Reference Data | high | D; 444 lines; keep owner; §3 seam diagnosis |
| `byte_retention.go` | Coordinates backup retention leases and publication/collection byte locks | `AcquireBackupRetention`, `Close` | lexical: `internal/app/recoveryassembly/vnext.go`, `base_release.go`; 6 total candidate files | `internal/platform/postgres`, `pgx/v5`; std | lexical: `object_collection_integration_test.go` | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 66 lines; keep owner; §3 seam diagnosis |
| `canonical_index.go` | Batches canonical entry and lookup-index persistence for verified content | `Append`, `Complete` | lexical: `base_release.go`, `coordinator_preparation.go`; 3 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `internal/platform/postgres`, `uuid`, `pgx/v5`, `pgx/v5/pgconn`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 148 lines; keep owner; §3 seam diagnosis |
| `canonical_persistence_integration_test.go` | Assertions: Canonical Base Persistence Pins And Immutable Provenance | `TestCanonicalBasePersistencePinsAndImmutableProvenance_Integration` | Go test runner; helper calls within owner tests | `internal/testutil/recoverytest`, `internal/app/extensionassembly`, `internal/app/projectionassembly`, `internal/app/recoveryassembly`, `internal/app/referenceassembly`, `internal/modules/evidence`, `internal/modules/recovery`, `internal/modules/reference_data`, `internal/modules/reference_data/testsupport`, `internal/platform/objectstore`, `internal/platform/pagination`, `internal/platform/postgres`, `internal/testutil/appsupport`, `internal/testutil/pgtest`, `uuid`, `pgx/v5`, `pgx/v5/pgconn`, `pgx/v5/pgxpool`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 426 lines; test/support, not runtime architecture |
| `canonical_publication.go` | Publishes versions, immutable sets, retained objects and provenance anchors transactionally | Package-private: `releasePublication`, `insertCandidateTx`, `publishVersionTx` | lexical: `administration_repository.go`, `administrative_actions.go`; 15 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `uuid`, `pgx/v5`; std | lexical: `consumer_test.go`, `coordinator_admission_integration_test.go`; 6 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 342 lines; keep owner; §3 seam diagnosis |
| `canonical_repository.go` | Private SQL adapter for current/historical sets, provenance, exact entries, paged lookup and pins | `NewConsumer`, `CurrentSet`, `RetainedSet`, `Provenance`, `Entry`, `Lookup`; +1 declarations | lexical: `internal/app/revisionassembly/revisions.go`, `internal/app/server/runtime_assembly.go`; 17 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/pagination`, `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `attack_native_integration_test.go`, `base_atomicity_integration_test.go`; 16 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 243 lines; keep owner; §3 seam diagnosis |
| `concurrent_publication_integration_test.go` | Assertions: Concurrent Root Publication Rejects Frozen Other Key | `TestConcurrentRootPublicationRejectsFrozenOtherKey_Integration` | Go test runner; helper calls within owner tests | `internal/testutil/httptestx`, `internal/app/referenceassembly`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/testutil/appsupport`, `uuid`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 178 lines; test/support, not runtime architecture |
| `configuration.go` | Owner configuration validation, overlay admission and trust-bootstrap decoding | `Configuration`, `TrustBootstrap`, `ValidateReferencePackConfiguration`, `ApplyConfigurationOverlay`, `AdmitTrustBootstrap` | lexical: `internal/app/configassembly/deployment.go`, `internal/app/configassembly/reference_pack.go`; 8 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `coordinator_admission_integration_test.go`, `coordinator_jobs_integration_test.go`; 3 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 52 lines; keep owner; §3 seam diagnosis |
| `consumer.go` | Five closed consumer operations, request validation order, typed errors and confidential lookup cursors | `PackProvenance`, `ConsumerError`, `ConsumerResult`, `GetPackEntryRequest`, `LookupPackEntriesRequest`, `EvaluateIndicatorRequest`; +12 declarations | lexical: `internal/app/incidentportabilityassembly/catalog.go`, `internal/modules/indicators/incident_bundle_contribution.go`; 20 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/pagination`; std | lexical: `attack_native_integration_test.go`, `base_atomicity_integration_test.go`; 12 candidate test files | C-11, C-12: consumer/set/provenance projections | Reference Data | high | D; 345 lines; keep owner; §3 seam diagnosis |
| `consumer_contract_test.go` | Shared test scenarios/fixtures: testConsumerMachineContracts | Package-private: `testConsumerMachineContracts` | Parent owner test scenarios (lexical): `consumer_test.go` | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-11, C-12: consumer/set/provenance projections | Reference Data | medium | D; 159 lines; test/support, not runtime architecture |
| `consumer_integrity_integration_test.go` | Assertions: Consumer Loss Publishes Invalidation And Aborts Source Transaction; Required Base Loss Clears Selection And Blocks Readiness; Unavailable Base Preserves Healthy Replacement | `OpenPublished`, `ReadAt`, `TestConsumerLossPublishesInvalidationAndAbortsSourceTransaction_Integration`, `TestRequiredBaseLossClearsSelectionAndBlocksReadiness_Integration`, `TestUnavailableBasePreservesHealthyReplacement_Integration` | Go test runner; helper calls within owner tests | `internal/testutil/httptestx`, `internal/app/referenceassembly`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/platform/pagination`, `internal/testutil/appsupport`, `pgx/v5`; std | This file's assertions/helpers; C-19 | C-11, C-12: consumer/set/provenance projections | Reference Data | medium | D; 321 lines; test/support, not runtime architecture |
| `consumer_test.go` | Assertions: Canonical Base Release Bindings And Consumer Evaluation; Canonical Consumer Cursor Binding Omission And Exact Expiry | `CurrentSet`, `RetainedSet`, `Provenance`, `Entry`, `Lookup`, `TestCanonicalBaseReleaseBindingsAndConsumerEvaluation_Unit`; +1 declarations | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`, `internal/platform/pagination`; std | This file's assertions/helpers; C-19 | C-11, C-12: consumer/set/provenance projections | Reference Data | medium | D; 267 lines; test/support, not runtime architecture |
| `contractadmission/admission.go` | Read-only startup/restore admission of retained current contract artifacts | `Reader`, `IsIncompatible`, `Error`, `ReasonCode`, `Validate` | lexical: `internal/app/operator/operator_reference_pack.go`, `historical_validation.go`; 3 total candidate files | `pgx/v5`, `internal/modules/reference_data/internal/packformat`; std | lexical: `contractadmission/binding_test.go`, `contractadmission/entra_test.go`; 7 candidate test files | C-15, C-17: current contract/readiness | Reference Data admission | high | D; 105 lines; keep owner; §3 seam diagnosis |
| `contractadmission/binding.go` | Validates dependent set bindings against exact retained canonical identity | `ValidateBinding` | lexical: `internal/app/referenceassembly/retained_admission.go` | `pgx/v5`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | lexical: `contractadmission/binding_test.go`, `base_atomicity_integration_test.go` | C-15, C-17: current contract/readiness | Reference Data admission | high | D; 84 lines; keep owner; §3 seam diagnosis |
| `contractadmission/binding_test.go` | Assertions: Reference Binding Load Failures | `TestReferenceBindingLoadFailures` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-15, C-17: current contract/readiness | Reference Data admission | medium | D; 33 lines; test/support, not runtime architecture |
| `contractadmission/entra_test.go` | Assertions: Entra Current Contract Admission | `TestEntraCurrentContractAdmission_Unit` | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-15, C-17: current contract/readiness | Reference Data admission | medium | D; 82 lines; test/support, not runtime architecture |
| `contractadmission/pending_test.go` | Assertions: Pending Current Contract Admission | `TestPendingCurrentContractAdmission_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-15, C-17: current contract/readiness | Reference Data admission | medium | D; 95 lines; test/support, not runtime architecture |
| `contractadmission/strict_test.go` | Assertions: Retained Artifacts Closed Admission | `TestRetainedArtifactsClosedAdmission_Unit` | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-15, C-17: current contract/readiness | Reference Data admission | medium | D; 136 lines; test/support, not runtime architecture |
| `coordinator.go` | Lifecycle application construction and staged import/retained verification admission | `Coordinator`, `CoordinatorOptions`, `ImportAdmission`, `ImportAdmissionOptions`, `VerificationRequest`, `PendingImport`; +10 declarations | lexical: `internal/app/operator/operator_reference_pack.go`, `internal/app/referenceassembly/administration.go`; 46 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `internal/platform/jobs`, `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `attack_endpoint_verification_test.go`, `base_atomicity_integration_test.go`; 22 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 380 lines; keep owner; §3 seam diagnosis |
| `coordinator_admission.go` | Freezes operation cohort, identity, trust and relevant revisions | Package-private: `previousSetID`, `probeContainerIdentity`, `admitOperationTx` | lexical: `administrative_actions.go`, `coordinator.go`; 11 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `uuid`, `pgx/v5`; std | lexical: `coordinator_admission_integration_test.go`, `verification_engine_test.go` | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 331 lines; keep owner; §3 seam diagnosis |
| `coordinator_admission_integration_test.go` | Assertions: Canonical Admission Freezes Cohort And Rejects Relevant Stale State; Canonical Coordinator Publishes Signed Content And Atomic Invalidation | `PublishStream`, `OpenPublished`, `StageStream`, `OpenStaged`, `RemoveStaged`, `RemovePublished`; +5 declarations | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`, `internal/testutil/pgtest`, `uuid`, `pgx/v5/pgxpool`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 598 lines; test/support, not runtime architecture |
| `coordinator_artifacts.go` | Retains attempt objects and verified artifact preparations | Package-private: `retainAttemptObject`, `prepareVerifiedObjects` | lexical: `coordinator_preparation.go` | `internal/platform/postgres`, `uuid`, `pgx/v5`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 90 lines; keep owner; §3 seam diagnosis |
| `coordinator_execution.go` | Runs lifecycle worker verification and finalizes success, rejection or execution abort | `Execute` | lexical: `internal/app/referenceassembly/administration.go`, `internal/modules/incidentbundles/worker_service.go` | `internal/modules/reference_data/internal/packformat`, `internal/modules/extensions/deadline`, `internal/platform/jobs`, `uuid`, `pgx/v5`; std | lexical: `base_atomicity_integration_test.go`, `base_reconciliation_integration_test.go`; 6 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 208 lines; keep owner; §3 seam diagnosis |
| `coordinator_execution_scope_integration_test.go` | Assertions: Canonical Coordinator Execution Scope | `TestCanonicalCoordinatorExecutionScope_Integration` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`, `internal/platform/jobs`, `uuid`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 82 lines; test/support, not runtime architecture |
| `coordinator_guards.go` | Ordered publication locks and captured dependency comparison | Package-private: `lockPublicationTx`, `lockPublicationDependenciesTx` | lexical: `coordinator_publication.go`, `portable_publication.go` | `uuid`, `pgx/v5`; std | lexical: `coordinator_admission_integration_test.go`, `portable_admission_integration_test.go` | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 127 lines; keep owner; §3 seam diagnosis |
| `coordinator_identity.go` | Live verifier strategy for retained bytes, release sequence and logical-version identity | Package-private: `checkRetainedPresence`, `checkRetainedLength`, `checkRetainedIntegrity` | lexical: `coordinator_preparation.go`, `historical_identity.go`; 5 total candidate files | `internal/modules/reference_data/internal/packformat`, `uuid`, `pgx/v5`; std | lexical: `verification_engine_test.go` | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 118 lines; keep owner; §3 seam diagnosis |
| `coordinator_integrity.go` | Inspects retained content integrity without silently treating operational failures as missing content | Package-private: `checkRetainedMembers`, `checkRetainedVersion`, `inspectRetainedVersion` | lexical: `administrative_actions.go`, `base_reconciliation.go`; 8 total candidate files | `internal/modules/reference_data/internal/packformat`, `pgx/v5`; std | lexical: `coordinator_admission_integration_test.go`, `coordinator_jobs_integration_test.go` | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 169 lines; keep owner; §3 seam diagnosis |
| `coordinator_jobs_integration_test.go` | Assertions: Canonical Coordinator Leased Jobs And Terminal Replay | `FinalizeReferencePackAction`, `FinalizeReferencePackJobTimeout`, `FinalizeReferencePackJobSuccess`, `FinalizeReferencePackJobFailure`, `FinalizeReferencePackJobCancellation`, `TestCanonicalCoordinatorLeasedJobsAndTerminalReplay_Integration` | Go test runner; helper calls within owner tests | `internal/platform/jobs/jobcontract`, `internal/modules/reference_data/internal/packformat`, `internal/platform/extensionstore`, `internal/platform/jobs`, `internal/testutil/collaborationsupport`, `internal/testutil/pgtest`, `uuid`, `pgx/v5`, `pgx/v5/pgconn`, `pgx/v5/pgxpool`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 529 lines; test/support, not runtime architecture |
| `coordinator_observation.go` | Observes cancellation/execution while preparation runs | Package-private: `observeDuringPreparation`, `finish` | lexical: `coordinator_execution.go`, `portable_execution.go` | `internal/platform/jobs`; std | lexical: `coordinator_execution_scope_integration_test.go`, `job_terminal_effects_integration_test.go`; 4 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 64 lines; keep owner; §3 seam diagnosis |
| `coordinator_preparation.go` | Canonical private preparation encoding and attributable attempt start | Package-private: `encodePrepared`, `decodePrepared`, `beginAttempt` | lexical: `administrative_actions.go`, `coordinator_execution.go`; 10 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `internal/platform/jobs`, `uuid`, `pgx/v5`; std | lexical: `coordinator_admission_integration_test.go`, `coordinator_execution_scope_integration_test.go`; 5 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 259 lines; keep owner; §3 seam diagnosis |
| `coordinator_publication.go` | Atomic attempt outcomes, fallback selection and terminal owner state | Package-private: `publishAttemptTx`, `loadSelectionTx`, `publishFallbackTx` | lexical: `administrative_actions.go`, `base_release.go`; 5 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/modules/reference_data/internal/packstate`, `uuid`, `pgx/v5`; std | lexical: `base_atomicity_integration_test.go`, `coordinator_admission_integration_test.go` | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 320 lines; keep owner; §3 seam diagnosis |
| `dependency_admission.go` | Captures recursive dependency key/revision guards without enlarging selected mutation cohort | Package-private: `admissionDependencyRevisions`, `admissionGuardKeys` | lexical: `coordinator_admission.go`, `portable_admission.go` | `internal/modules/reference_data/internal/packformat`, `pgx/v5`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 75 lines; keep owner; §3 seam diagnosis |
| `entra_lifecycle_integration_test.go` | Assertions: Entra Canonical Production Lifecycle; Entra Real Release Import | `TestEntraCanonicalProductionLifecycle_Integration`, `TestEntraRealReleaseImport_Integration` | Go test runner; helper calls within owner tests | `internal/app/referenceassembly`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/modules/reference_data/producer`, `internal/modules/reference_data/testsupport`, `internal/platform/pagination`, `internal/testutil/appsupport`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 213 lines; test/support, not runtime architecture |
| `historical_attempts.go` | Validates retained historical attempt/outcome linkage | Package-private: `validateHistoricalAttempts` | lexical: `historical_validation.go` | `internal/modules/reference_data/internal/packformat`, `uuid`, `pgx/v5`; std | lexical: `job_terminal_effects_integration_test.go`, `portable_admission_integration_test.go` | C-15, C-17: current contract/readiness | Reference Data | high | D; 113 lines; keep owner; §3 seam diagnosis |
| `historical_identity.go` | Historical verifier identity strategy independent of live admission | Package-private: `checkRetainedPresence`, `checkRetainedLength`, `checkRetainedIntegrity` | lexical: `coordinator_identity.go`, `historical_validation.go`; 5 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `verification_engine_test.go` | C-15, C-17: current contract/readiness | Reference Data | high | D; 42 lines; keep owner; §3 seam diagnosis |
| `historical_validation.go` | Validates required/historical state and reconstructs authorized retained state on restore | `ValidateRequiredState`, `RestoreHistoricalState`, `Append`, `Open` | lexical: `internal/app/referenceassembly/recovery_storage.go`, `internal/app/server/runtime_assembly.go`; 3 total candidate files | `internal/modules/reference_data/contractadmission`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `base_atomicity_integration_test.go`, `base_reconciliation_integration_test.go`; 11 candidate test files | C-15, C-17: current contract/readiness | Reference Data | high | D; 447 lines; keep owner; §3 seam diagnosis |
| `http_adapter.go` | Binds nine administrative operations, maps errors/envelopes/paging and dispatches fresh accepted Jobs | `RegisterRoutes`, `Resource` | lexical: `internal/app/extensionassembly/reference_pack_jobs.go`, `internal/app/operator/operator_reference_pack.go`; 17 total candidate files | `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/httpauth`, `internal/platform/listquery`, `internal/platform/pagination`, `uuid`; std | lexical: `api_test.go`, `coordinator_jobs_integration_test.go`; 6 candidate test files | C-01–09, C-10: HTTP/Jobs projections | Reference Data | high | D; 559 lines; F-01: auth ordering; no move decided |
| `http_requests.go` | Closed HTTP/upload request admission and normalized action/refresh identity | `DecodeImportMetadata`, `DecodeActionRequest`, `DecodeRemovalRequest`, `DecodeRefreshRequest` | lexical: `http_adapter.go` | `internal/platform/canonicaljson`, `internal/platform/fieldnorm`, `internal/platform/httpapi`; std | lexical: `api_test.go`, `base_atomicity_integration_test.go`; 6 candidate test files | C-01–09, C-10: HTTP/Jobs projections | Reference Data | high | D; 315 lines; keep owner; §3 seam diagnosis |
| `idempotency_receipts.go` | Transaction-bound exact action receipt locks, reads and writes | Package-private: `lockReceiptTx`, `readReceipt`, `writeReceiptTx` | lexical: `administrative_actions.go`, `coordinator.go`; 3 total candidate files | `internal/platform/authn`, `uuid`, `pgx/v5`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 74 lines; keep owner; §3 seam diagnosis |
| `incident_export.go` | Streams exact redistribution-eligible retained pack containers into incident export | `IncidentReferenceExportRequest`, `ExportContentTx` | lexical: `internal/modules/incidentbundles/source.go`, `incident_retention.go`; 4 total candidate files | `internal/modules/reference_data/internal/packformat`, `uuid`, `pgx/v5`; std | lexical: `incident_export_integration_test.go`, `incident_retention_integration_test.go`; 3 candidate test files | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 169 lines; keep owner; §3 seam diagnosis |
| `incident_export_integration_test.go` | Shared test scenarios/fixtures: testPortableContainerExport | Package-private: `testPortableContainerExport` | Parent owner test scenarios (lexical): `coordinator_admission_integration_test.go` | `internal/modules/reference_data/internal/packformat`, `uuid`, `pgx/v5`, `pgx/v5/pgxpool`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 112 lines; test/support, not runtime architecture |
| `incident_portability.go` | Public exact catalog validation, encoding and portable binding facade | `IncidentBundleReferenceValidationError`, `Error`, `IncidentBundleReferenceInvariant`, `DecodeIncidentBundleReferences`, `ValidateIncidentBundleReferences`, `EncodeIncidentBundleReferences`; +2 declarations | lexical: `internal/modules/incidentbundles/source.go`, `testsupport/portable_fixture.go`; 10 total candidate files | `internal/platform/canonicaljson`, `internal/modules/reference_data/internal/packformat`, `pgx/v5`; std | lexical: `coordinator_jobs_integration_test.go`, `incident_export_integration_test.go`; 8 candidate test files | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 137 lines; keep owner; §3 seam diagnosis |
| `incident_portability_test.go` | Assertions: Incident Bundle Reference Catalog Validation | `TestIncidentBundleReferenceCatalogValidation_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 129 lines; test/support, not runtime architecture |
| `incident_retention.go` | Opaque pack-side incident preparation, exact reuse, resolution, pinning and same-transaction apply | `IncidentReferences`, `IncidentReferenceImportRequest`, `IncidentReferenceOptions`, `PreparedReferenceImport`, `NewIncidentReferences`, `PrepareImport`; +2 declarations | lexical: `internal/app/operator/operator_reference_pack.go`, `internal/app/server/runtime_assembly.go`; 14 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `internal/platform/jobs`, `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `coordinator_jobs_integration_test.go`, `incident_retention_integration_test.go`; 7 candidate test files | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 565 lines; keep owner; §3 seam diagnosis |
| `incident_retention_integration_test.go` | Assertions: Incident References Retain Local Versions And Unavailable History | `TestIncidentReferencesRetainLocalVersionsAndUnavailableHistory_Integration` | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`, `internal/platform/jobs`, `internal/testutil/collaborationsupport`, `internal/testutil/pgtest`, `uuid`, `pgx/v5/pgconn`, `pgx/v5/pgxpool`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 322 lines; test/support, not runtime architecture |
| `incident_retention_integrity.go` | Validates retained source catalog and resolution/pin linkage | Package-private: `validatePortableCatalogs` | lexical: `historical_validation.go` | `internal/modules/reference_data/internal/packformat`, `uuid`, `pgx/v5`; std | lexical: `incident_retention_integration_test.go`, `portable_admission_integration_test.go` | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 105 lines; keep owner; §3 seam diagnosis |
| `integrity.go` | Publishes unavailable-pack consequences; supports invalidation after a source transaction rolls back | `IntegrityOptions`, `InvalidateUnavailablePack` | lexical: `internal/app/server/runtime_assembly.go`, `canonical_repository.go`; 4 total candidate files | `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `coordinator_jobs_integration_test.go`, `reference_pack_test_helpers_test.go`; 4 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 129 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/admission_fixture_test.go` | Shared test scenarios/fixtures: runAdmissionManifest, assertManifestVerdict | Package-private: `runAdmissionManifest`, `assertManifestVerdict` | Parent owner test scenarios (lexical): `internal/packformat/dependency_fixture_test.go`, `internal/packformat/fixture_contract_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 118 lines; test/support, not runtime architecture |
| `internal/packformat/archive.go` | Hostile archive inventory, paths, members, limits and bounded IO | `Failure`, `ArchiveLimits`, `Destination`, `Member`, `Inventory`, `Error`; +4 declarations | lexical: `base_reconciliation.go`, `base_release.go`; 43 total candidate files | std only | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/archive_limit_test.go`; 27 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 567 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/archive_limit_test.go` | Shared test scenarios/fixtures: testCompleteArchiveFileByteBoundary | `Read`, `Write` | Parent owner test scenarios (lexical): `internal/packformat/archive_ratio_test.go`, `internal/packformat/archive_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 163 lines; test/support, not runtime architecture |
| `internal/packformat/archive_program.go` | Ordered archive checks and diagnostics before extraction | `Extract`, `ExtractWithDiagnostics`, `ArchiveChecks` | lexical: `coordinator_admission.go`, `producer/release_verification.go`; 12 total candidate files | std only | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/archive_limit_test.go`; 15 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 261 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/archive_ratio_test.go` | Shared test scenarios/fixtures: testCompleteCompressionRatio | Package-private: `testCompleteCompressionRatio` | Parent owner test scenarios (lexical): `internal/packformat/manifest_limit_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 106 lines; test/support, not runtime architecture |
| `internal/packformat/archive_test.go` | Assertions: Archive ZIPMetadata Has Bounded Working Reads; Archive ZIPChecks Decoded CRCSize And Compressed End; Archive TARPadding Has Independent Structural Bound; 8 additional test entries | `ReadAt`, `TestArchiveZIPMetadataHasBoundedWorkingReads_Unit`, `TestArchiveZIPChecksDecodedCRCSizeAndCompressedEnd_Unit`, `TestArchiveTARPaddingHasIndependentStructuralBound_Unit`, `Write`, `TestArchiveDirectoryMarkersHaveAnIndependentStructuralBound_Unit`; +7 declarations | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 477 lines; test/support, not runtime architecture |
| `internal/packformat/archive_writer.go` | Deterministic USTAR output for canonical producer/fixture containers | `WriteUSTAR` | lexical: `producer/envelope.go`, `producer/windows_reconcile.go` | std only | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 37 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/attack.go` | Native ATTACK object semantics, references and canonical projection support | `ATTACKText`, `ATTACKDescription`, `ATTACKLabel`, `ATTACKDeprecated`, `ATTACKAliases`, `ATTACKExternalRefs`; +3 declarations | lexical: `producer/package.go`, `internal/packformat/attestation.go`; 22 total candidate files | `internal/platform/canonicaljson`, `uuid`, `golang.org/x/text/unicode/norm`; std | lexical: `internal/packformat/archive_limit_test.go`, `internal/packformat/archive_test.go`; 19 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 373 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/attack_native_test.go` | Assertions: Native ATTACKProjection And Local References | `TestNativeATTACKProjectionAndLocalReferences_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 148 lines; test/support, not runtime architecture |
| `internal/packformat/attack_remediation_test.go` | Assertions: Native ATTACKTyped Embedded Endpoints; Native ATTACKLimit Summary; Schema Union Dispatch And Limit Bindings | `TestNativeATTACKTypedEmbeddedEndpoints_Unit`, `TestNativeATTACKLimitSummary_Unit`, `TestSchemaUnionDispatchAndLimitBindings_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 219 lines; test/support, not runtime architecture |
| `internal/packformat/attempt_result.go` | Closed canonical attempt-result validation and encoding | `ValidateAttemptResult`, `EncodeAttemptResult` | lexical: `coordinator_preparation.go`, `coordinator_publication.go`; 6 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/attempt_result_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 60 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/attempt_result_test.go` | Shared test scenarios/fixtures: testAttemptResultBoundary | Package-private: `testAttemptResultBoundary` | Parent owner test scenarios (lexical): `internal/packformat/portability_retention_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 74 lines; test/support, not runtime architecture |
| `internal/packformat/attestation.go` | Closed attestation shape and identity validation | `ValidateAttestation` | lexical: `attestation.go`, `historical_validation.go` | `internal/platform/canonicaljson`, `uuid`; std | lexical: `internal/packformat/attestation_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 84 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/attestation_test.go` | Assertions: Attestation Independent Identity And Closed Shape | `TestAttestationIndependentIdentityAndClosedShape_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 93 lines; test/support, not runtime architecture |
| `internal/packformat/bootstrap.go` | Strict trust-bootstrap decoding and root version extraction | `DecodeBootstrap`, `RootVersion` | lexical: `attestation.go`, `canonical_publication.go`; 8 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/trust_fixture_test.go`, `internal/packformat/tuf_limit_test.go`; 5 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 66 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/builtin_binding.go` | Decodes and matches application-release built-in identity bindings | `BuiltinReleaseBinding`, `DecodeBuiltinReleaseBinding`, `Matches` | lexical: `base_release.go`, `historical_validation.go` | std only | lexical: `internal/packformat/builtin_binding_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 37 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/builtin_binding_test.go` | Assertions: Builtin Binding Closes Identity And Shape | `TestBuiltinBindingClosesIdentityAndShape_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 106 lines; test/support, not runtime architecture |
| `internal/packformat/consumer_contract.go` | Admits consumer request/result shapes using generated machine contracts | `ValidateConsumerRequest`, `AdmitConsumerRequest`, `ValidateConsumerResult` | lexical: `consumer.go` | `internal/platform/canonicaljson`; std | lexical: `consumer_contract_test.go`, `profile_consumers_integration_test.go` | C-11, C-12: consumer/set/provenance projections | Reference Data private formats | high | D; 153 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/content.go` | Profile content schema/semantics, canonical rows, registry compatibility and exact identities | `ContentSource`, `LookupKey`, `ContentRow`, `ContentSink`, `ValidateContent`, `ValidateContentSchema`; +6 declarations | lexical: `base_reconciliation.go`, `base_release.go`; 21 total candidate files | `internal/platform/canonicaljson`, `golang.org/x/text/unicode/norm`; std | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/archive_limit_test.go`; 20 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 552 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/content_admission_test.go` | Shared test scenarios/fixtures: testNestedContentAdmission | Package-private: `testNestedContentAdmission`, `wrongContentType`, `nullableContentMember` | Parent owner test scenarios (lexical): `internal/packformat/content_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 101 lines; test/support, not runtime architecture |
| `internal/packformat/content_diagnostics.go` | Enumerates content/schema/registry-compatibility findings | `ContentSemanticsFindings`, `RegistryCompatibilityFindings`, `ContentSchemaSummary`, `ContentSchemaFindings` | lexical: `producer/package.go`, `verification_program.go` | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/attack_native_test.go`, `internal/packformat/attack_remediation_test.go`; 6 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 158 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/content_diagnostics_test.go` | Shared test scenarios/fixtures: testContentSemanticFindings, testRegistrySemanticFindings, testContentSchemaFindings | Package-private: `testContentSemanticFindings`, `testRegistrySemanticFindings`, `testContentSchemaFindings` | Parent owner test scenarios (lexical): `internal/packformat/content_test.go`, `internal/packformat/diagnostics_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 272 lines; test/support, not runtime architecture |
| `internal/packformat/content_limit_test.go` | Shared test scenarios/fixtures: testCompleteContentLineBoundary | Package-private: `testCompleteContentLineBoundary` | Parent owner test scenarios (lexical): `internal/packformat/content_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 76 lines; test/support, not runtime architecture |
| `internal/packformat/content_test.go` | Assertions: Content Schema Failure Precedence; Content Rejects Malformed Streams And Semantic Ambiguity; Windows Event Identity And Lookup Boundaries; 18 additional test entries | `Open`, `Append`, `TestContentSchemaFailurePrecedence_Unit`, `TestContentRejectsMalformedStreamsAndSemanticAmbiguity_Unit`, `TestWindowsEventIdentityAndLookupBoundaries_Unit`, `TestWindowsSIDIdentityBoundaries_Unit`; +17 declarations | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 435 lines; test/support, not runtime architecture |
| `internal/packformat/dependencies.go` | Exact dependency closure, supported key guards and conflict checks | `DependencyResolver`, `VerifyDependencies`, `DependencyChecks`, `SupportedDependencyKeys` | lexical: `dependency_admission.go`, `verification_program.go` | std only | lexical: `internal/packformat/dependencies_test.go`, `internal/packformat/dependency_fixture_test.go`; 3 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 202 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/dependencies_test.go` | Assertions: Dependency Verification Graph And Precedence | `TestDependencyVerificationGraphAndPrecedence_Unit` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 67 lines; test/support, not runtime architecture |
| `internal/packformat/dependency_fixture_test.go` | Shared test scenarios/fixtures: runDependencyManifest | Package-private: `runDependencyManifest` | Parent owner test scenarios (lexical): `internal/packformat/fixture_contract_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 37 lines; test/support, not runtime architecture |
| `internal/packformat/diagnostic_collector.go` | Bounded external sorting/deduplication of verification issues through scratch capability | `DiagnosticScratch`, `Len`, `Less`, `Swap`, `Push`, `Pop` | lexical: `internal/packformat/archive_program.go`, `internal/packformat/attack.go`; 13 total candidate files | std only | lexical: `internal/packformat/attack_native_test.go`, `internal/packformat/diagnostics_test.go`; 3 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 277 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/diagnostic_safety.go` | Safe diagnostic vocabulary, paths and bounded decimal measurements | Package-private: `loadDiagnosticVocabulary`, `diagnosticDecimal`, `diagnosticMeasurement` | lexical: `internal/packformat/diagnostics.go` | `internal/gen/contractreferencepacks`; std | lexical: `internal/packformat/limit_diagnostics_test.go` | C-12–14: canonical pack/check projections; reads `internal/gen/contractreferencepacks` | Reference Data private formats | high | D; 183 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/diagnostics.go` | Ranked verification check program and deterministic summary/issue identities | `Check`, `SafeDetails`, `Finding`, `Issue`, `ValidationSummary`, `FindingSink`; +7 declarations | lexical: `audit.go`, `coordinator_execution.go`; 33 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/attack_native_test.go`; 14 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 328 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/diagnostics_test.go` | Assertions: Diagnostic Registry Selects One Check And Stable Deduplicated Prefix | `TestDiagnosticRegistrySelectsOneCheckAndStableDeduplicatedPrefix_Unit`, `Write`, `Open`, `Remove`, `Close` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 261 lines; test/support, not runtime architecture |
| `internal/packformat/duplicate_diagnostics.go` | Reports duplicate original JSON members without losing their location | Package-private: `duplicateManifestFindings`, `diagnosticProperty`, `duplicateValue` | lexical: `internal/packformat/manifest_program.go` | std only | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 94 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/entra_content_test.go` | Shared test scenarios/fixtures: testEntraBoundaries | Package-private: `testEntraBoundaries` | Parent owner test scenarios (lexical): `internal/packformat/content_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 78 lines; test/support, not runtime architecture |
| `internal/packformat/envelope.go` | Admits closed successful verification envelopes and their exact retained bytes | `ValidateSuccessfulEnvelope` | lexical: `successful_envelope.go`, `internal/packformat/prepared.go` | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/prepared_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 164 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/fixture_contract_test.go` | Assertions: Executable Fixture Contracts | `TestExecutableFixtureContracts_Unit` | Go test runner; helper calls within owner tests | `internal/gen/contractreferencepackfixtures`, `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections; reads `internal/gen/contractreferencepackfixtures` | Reference Data private formats | medium | D; 414 lines; test/support, not runtime architecture |
| `internal/packformat/fixture_projection_test.go` | Shared test scenarios/fixtures: testFixtureProjectionIsolation | Package-private: `fixtureProjection`, `compileFixtureProjection`, `testFixtureProjectionIsolation` | Parent owner test scenarios (lexical): `internal/packformat/admission_fixture_test.go`, `internal/packformat/attestation_test.go` | `internal/gen/contractreferencepackfixtures`, `internal/gen/contractreferencepacks`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections; reads `internal/gen/contractreferencepackfixtures`, `internal/gen/contractreferencepacks` | Reference Data private formats | medium | D; 44 lines; test/support, not runtime architecture |
| `internal/packformat/fixture_scratch_test.go` | Shared test scenarios/fixtures: fixtureRelationScratch | Package-private: `fixtureRelationScratch` | Parent owner test scenarios (lexical): `internal/packformat/fixture_contract_test.go` | std only | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 24 lines; test/support, not runtime architecture |
| `internal/packformat/framework_lookup_test.go` | Shared test scenarios/fixtures: testFrameworkExternalLookupKeys | Package-private: `testFrameworkExternalLookupKeys` | Parent owner test scenarios (lexical): `internal/packformat/content_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 73 lines; test/support, not runtime architecture |
| `internal/packformat/identity.go` | Canonical set/member digest and identity construction and bound-set decoding | `SetMember`, `Set`, `Digest`, `ValidateInventory`, `ValidateInventoryLengths`, `ValidateInventoryHashes`; +5 declarations | lexical: `attestation.go`, `base_release.go`; 48 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/archive_limit_test.go`; 28 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 142 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/indicator.go` | Registry-selected normalization, validation, defanging and dedupe algorithms | `IndicatorPolicy`, `Evaluation`, `IndicatorPolicies`, `Evaluate`, `TrimIndicatorInput`, `NormalizeURL` | lexical: `consumer.go`, `public_types.go`; 6 total candidate files | `golang.org/x/text/unicode/norm`; std | lexical: `internal/packformat/indicator_test.go`, `public_types_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 492 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/indicator_test.go` | Assertions: Indicator Algorithms Canonical Vectors; URLNormalization Preserves Delimiters And Is Idempotent | `TestIndicatorAlgorithmsCanonicalVectors_Unit`, `TestURLNormalizationPreservesDelimitersAndIsIdempotent_Unit` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 102 lines; test/support, not runtime architecture |
| `internal/packformat/inventory_diagnostics.go` | Inventory length/hash and trust-binding diagnostic program | `InventoryFindings`, `TrustBindingFindings` | lexical: `producer/release_verification.go`, `verification_program.go`; 3 total candidate files | std only | lexical: `internal/packformat/inventory_diagnostics_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 119 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/inventory_diagnostics_test.go` | Assertions: Inventory And Binding Diagnostics Enumerate Every Member | `TestInventoryAndBindingDiagnosticsEnumerateEveryMember_Unit` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 46 lines; test/support, not runtime architecture |
| `internal/packformat/license.go` | SPDX license expression and manifest license findings | `ManifestLicenseFindings` | lexical: `verification_program.go`, `internal/packformat/diagnostic_collector.go`; 4 total candidate files | std only | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/manifest_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 192 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/limit_diagnostics_test.go` | Shared test scenarios/fixtures: testLimitDiagnostics | Package-private: `testLimitDiagnostics`, `checkProgramRange` | Parent owner test scenarios (lexical): `internal/packformat/diagnostics_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 166 lines; test/support, not runtime architecture |
| `internal/packformat/manifest.go` | Manifest/hint decoding, profile registry, canonical payload digest and runtime compatibility | `Profile`, `Manifest`, `SourceArtifact`, `Dependency`, `Conflict`, `File`; +14 declarations | lexical: `administration_repository.go`, `attestation.go`; 54 total candidate files | `internal/platform/canonicaljson`, `golang.org/x/text/unicode/norm`; std | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/archive_limit_test.go`; 27 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 293 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/manifest_diagnostics.go` | Closed manifest/hint and compatibility issue enumeration | `ManifestCompatibilityFindings`, `DecodeManifestWithDiagnostics`, `DecodeHintWithDiagnostics` | lexical: `verification_program.go`, `internal/packformat/manifest.go`; 3 total candidate files | std only | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/manifest_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 191 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/manifest_limit_test.go` | Assertions: ATTACKManifest And Path Limits | `TestATTACKManifestAndPathLimits_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 126 lines; test/support, not runtime architecture |
| `internal/packformat/manifest_program.go` | Ordered hint/manifest verification and disallowed-content findings | `ManifestChecks`, `HintChecks`, `DisallowedContentFindings` | lexical: `verification_program.go`, `internal/packformat/manifest.go`; 3 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/limit_diagnostics_test.go`, `internal/packformat/manifest_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 143 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/manifest_test.go` | Assertions: Manifest Independent Identity Vector; Manifest Closes Nested Omissions And Nulls; License Grammar And Notice Bindings; 2 additional test entries | `TestManifestIndependentIdentityVector_Unit`, `TestManifestClosesNestedOmissionsAndNulls_Unit`, `TestLicenseGrammarAndNoticeBindings_Unit`, `TestPackSetIdentityIsContentOnlyAndRequiresBase_Unit`, `TestManifestDiagnosticEnumerationAndPrecedence_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 372 lines; test/support, not runtime architecture |
| `internal/packformat/native_fixture_test.go` | Shared test scenarios/fixtures: Open, runNativeFixture | `Open` | Parent owner test scenarios (lexical): `internal/packformat/fixture_contract_test.go` | std only | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 80 lines; test/support, not runtime architecture |
| `internal/packformat/notice.go` | Bounded streamed NOTICE text validation | `ValidateNoticeStream` | lexical: `verification_program.go` | `golang.org/x/text/transform`, `golang.org/x/text/unicode/norm`; std | lexical: `internal/packformat/archive_limit_test.go`, `internal/packformat/manifest_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 74 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/operation_input.go` | Closed canonical lifecycle input and pending-input validation | `OperationInput`, `DecodeOperationInput`, `ValidatePendingInput` | lexical: `contractadmission/admission.go`, `coordinator_admission.go` | `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 103 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/portability.go` | Exact v2 set/version reference catalog decoding, canonical ordering and encoding | `PortableVersion`, `PortableReferences`, `DecodePortableReferences`, `DecodeRetainedPortableReferences`, `EncodePortableReferences` | lexical: `contractadmission/admission.go`, `contractadmission/binding.go`; 7 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/portability_retention_test.go`, `internal/packformat/portable_content_test.go`; 5 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 124 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/portability_retention.go` | Validates retained portability request/result binding and resolution shape | `PortabilityInput`, `DecodePortabilityInput`, `ValidatePortabilityRetention` | lexical: `incident_retention.go`, `incident_retention_integrity.go`; 4 total candidate files | `internal/platform/canonicaljson`, `uuid`; std | lexical: `internal/packformat/portability_retention_test.go` | C-16, C-17: retention/portability artifacts | Reference Data private formats | high | D; 85 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/portability_retention_test.go` | Assertions: Portable Retention Closed Shape And Catalog Binding; Portable Catalog Reachable Set And Version Limits | `TestPortableRetentionClosedShapeAndCatalogBinding_Unit`, `TestPortableCatalogReachableSetAndVersionLimits_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data private formats | medium | D; 174 lines; test/support, not runtime architecture |
| `internal/packformat/portable_content.go` | Embedded container descriptors, exact required members and allowed inventory paths | `PortableContainer`, `PortableRequiredMember`, `PortableContent`, `EmptyPortableContent`, `DecodePortableContent`, `PortableContainerPath`; +3 declarations | lexical: `incident_export.go`, `incident_retention.go`; 7 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/portability_retention_test.go`, `internal/packformat/portable_content_test.go`; 6 candidate test files | C-16, C-17: retention/portability artifacts | Reference Data private formats | high | D; 150 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/portable_content_test.go` | Shared test scenarios/fixtures: testPortableContentManifest, testPortableContentReachableLimits | Package-private: `testPortableContentManifest`, `testPortableContentReachableLimits` | Parent owner test scenarios (lexical): `internal/packformat/portability_retention_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data private formats | medium | D; 262 lines; test/support, not runtime architecture |
| `internal/packformat/portable_context.go` | Closed frozen destination verification context | `PortableVerificationContext`, `DecodePortableVerificationContext` | lexical: `portable_admission.go`, `portable_history.go` | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/portable_context_test.go` | C-16, C-17: retention/portability artifacts | Reference Data private formats | high | D; 40 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/portable_context_test.go` | Shared test scenarios/fixtures: testPortableVerificationContext | Package-private: `testPortableVerificationContext` | Parent owner test scenarios (lexical): `internal/packformat/portability_retention_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data private formats | medium | D; 58 lines; test/support, not runtime architecture |
| `internal/packformat/prepared.go` | Private preparation artifact shape, size and binding validation | `ValidatePrepared` | lexical: `contractadmission/admission.go`, `coordinator_preparation.go` | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/prepared_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 105 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/prepared_test.go` | Assertions: Prepared Bound And Schema Reference Confinement | `TestPreparedBoundAndSchemaReferenceConfinement_Unit` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 43 lines; test/support, not runtime architecture |
| `internal/packformat/profile.go` | Profile-specific entry identities, lookup unions and normalized lookup keys | `LookupInput`, `ValidateEntryIdentity`, `NormalizeLookup` | lexical: `canonical_repository.go`, `consumer.go`; 3 total candidate files | `golang.org/x/text/unicode/norm`; std | lexical: `internal/packformat/content_test.go`, `internal/packformat/entra_content_test.go`; 3 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 494 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/provenance.go` | Closed immutable provenance validation | `ValidateProvenance` | lexical: `canonical_publication.go`, `contractadmission/binding.go`; 4 total candidate files | `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-11, C-12: consumer/set/provenance projections | Reference Data private formats | high | D; 85 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/semantic_findings.go` | Owner-local deterministic semantic finding helpers | Package-private: `require`, `finish`, `ordered` | lexical: `internal/packformat/archive.go`, `internal/packformat/archive_program.go`; 12 total candidate files | std only | lexical: `internal/packformat/diagnostics_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 53 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/shape.go` | Compiles machine-schema projections into bounded shape validation | `ProjectionDigest` | lexical: `producer/package.go`, `internal/packformat/archive_program.go`; 32 total candidate files | `internal/gen/contractreferencepacks`, `internal/platform/canonicaljson`; std | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/attack_native_test.go`; 16 candidate test files | C-12–14: canonical pack/check projections; reads `internal/gen/contractreferencepacks` | Reference Data private formats | high | D; 244 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/shape_diagnostics.go` | Enumerates safe schema-path shape diagnostics | Package-private: `diagnosticMember`, `diagnosticKind`, `walk` | lexical: `internal/packformat/content_diagnostics.go`, `internal/packformat/diagnostic_safety.go`; 7 total candidate files | `internal/platform/canonicaljson`, `uuid`; std | lexical: `internal/packformat/attack_remediation_test.go` | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 279 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/shape_limits.go` | Loads fixed shape limits from machine contracts | Package-private: `loadFixedShapeLimits` | Owner package/interface dispatch; no external lexical caller identified | std only | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 16 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/shape_union.go` | Compiles and selects discriminated schema union branches | Package-private: `literalKey`, `finiteValues`, `compileUnion` | lexical: `internal/packformat/shape.go`, `internal/packformat/shape_diagnostics.go` | `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 119 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/trust_fixture_test.go` | Shared test scenarios/fixtures: runTrustManifest | Package-private: `runTrustManifest` | Parent owner test scenarios (lexical): `internal/packformat/fixture_contract_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 78 lines; test/support, not runtime architecture |
| `internal/packformat/tuf.go` | Offline TUF structures, cryptographic checks, bootstrap and historical root verification | `TrustSnapshot`, `RetainedMetadata`, `TrustProposal`, `RootTransition`, `AdmitBootstrap`, `VerifyHistoricalRoot` | lexical: `attestation.go`, `canonical_publication.go`; 41 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/archive_limit_test.go`, `internal/packformat/archive_ratio_test.go`; 27 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 345 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/tuf_diagnostics.go` | TUF metadata shape issue enumeration | `TrustSchemaSummary` | lexical: `internal/packformat/tuf.go`, `internal/packformat/tuf_program.go` | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/limit_diagnostics_test.go`, `internal/packformat/tuf_limit_test.go`; 3 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 145 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/tuf_limit_test.go` | Shared test scenarios/fixtures: testCompleteTUFByteBoundaries | Package-private: `testCompleteTUFByteBoundaries` | Parent owner test scenarios (lexical): `internal/packformat/tuf_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 225 lines; test/support, not runtime architecture |
| `internal/packformat/tuf_program.go` | Ordered offline trust verification, rollback/expiry and binding checks | `VerifyTrust`, `VerifyTrustWithDiagnostics`, `TrustChecks` | lexical: `producer/release_verification.go`, `verification_program.go`; 40 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/admission_fixture_test.go`, `internal/packformat/archive_limit_test.go`; 30 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 414 lines; keep owner; §3 seam diagnosis |
| `internal/packformat/tuf_root_shape_test.go` | Assertions: Offline TUFRoot Nested Admission Precedes Canonical And Rotation | `TestOfflineTUFRootNestedAdmissionPrecedesCanonicalAndRotation_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 103 lines; test/support, not runtime architecture |
| `internal/packformat/tuf_test.go` | Assertions: Offline TUFRequires Both Root Thresholds And Rejects Extra Signatures; Offline TUFIndependent Vectors And Root Replay; Offline TUFFails Closed On Tampering And Expiry; 4 additional test entries | `TestOfflineTUFRequiresBothRootThresholdsAndRejectsExtraSignatures_Unit`, `TestOfflineTUFIndependentVectorsAndRootReplay_Unit`, `TestOfflineTUFFailsClosedOnTamperingAndExpiry_Unit`, `TestOfflineTUFDescriptorAdmission_Unit`, `TestOfflineTUFCombinedFaultPrecedence_Unit`, `TestOfflineTUFCompleteSchemaDiagnostics_Unit`; +1 declarations | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`, `internal/platform/cryptography`; std | This file's assertions/helpers; C-19 | C-12–14: canonical pack/check projections | Reference Data private formats | medium | D; 664 lines; test/support, not runtime architecture |
| `internal/packformat/validation_summary.go` | Closed bounded validation summary decoding | `DecodeValidationSummary` | lexical: `coordinator_execution.go`, `coordinator_preparation.go`; 3 total candidate files | `internal/platform/canonicaljson`; std | lexical: `internal/packformat/archive_test.go`, `internal/packformat/content_diagnostics_test.go`; 5 candidate test files | C-12–14: canonical pack/check projections | Reference Data private formats | high | D; 68 lines; keep owner; §3 seam diagnosis |
| `internal/packstate/state.go` | Pure linked lifecycle transitions, refresh cohort, dependency fallback and revision comparison | `Health`, `Version`, `Envelope`, `Operation`, `Verdict`, `Rejection`; +8 declarations | lexical: `administration_repository.go`, `administrative_actions.go`; 5 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `internal/packstate/state_test.go` | C-12–14: lifecycle/identity/verification | Reference Data pure state | high | D; 403 lines; keep owner; §3 seam diagnosis |
| `internal/packstate/state_test.go` | Assertions: Verification Transitions Preserve Disablement And Successful History; Refresh Admission Freezes Exact Successful Cohort; Dependency Fallback Prunes To Fixed Point And Rejects Cycles; 1 additional test entries | `TestVerificationTransitionsPreserveDisablementAndSuccessfulHistory_Unit`, `TestRefreshAdmissionFreezesExactSuccessfulCohort_Unit`, `TestDependencyFallbackPrunesToFixedPointAndRejectsCycles_Unit`, `TestPublicationRevisionComparisonIgnoresUnrelatedState_Unit` | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data pure state | medium | D; 175 lines; test/support, not runtime architecture |
| `job_finalization.go` | Owner mutation/finalization ports for shared Jobs and Extensions commit proof | `JobSuccessMutation`, `JobSuccessFinalization`, `JobFailureFinalization`, `JobCancellationFinalization`, `ActionFinalizer`, `JobSuccessFinalizer` | lexical: `internal/app/extensionassembly/reference_pack_jobs.go`, `internal/app/server/runtime_assembly.go`; 11 total candidate files | `pgx/v5`, `internal/platform/jobs`; std | lexical: `coordinator_jobs_integration_test.go`, `internal/app/referenceassembly/administration_test.go`; 5 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 48 lines; keep owner; §3 seam diagnosis |
| `job_ports.go` | Private narrow Jobs admission/execution operations and failure-completion construction | Package-private: `failedCompletion` | lexical: `coordinator.go`, `coordinator_execution.go`; 3 total candidate files | `internal/platform/jobs`, `uuid`, `pgx/v5`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 33 lines; keep owner; §3 seam diagnosis |
| `job_reconciliation.go` | Repairs owner terminal state against retained terminal Jobs | `TerminalJobReader`, `ReconcileTerminalJobs` | lexical: `internal/app/server/runtime_assembly.go` | `internal/platform/jobs`, `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `job_terminal_effects_integration_test.go` | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 83 lines; keep owner; §3 seam diagnosis |
| `job_terminal_effects.go` | Same-transaction owner abort/candidate closure on terminal Jobs outcomes | `JobTerminalEffects`, `ApplyJobTerminalEffectsTx` | lexical: `internal/app/referenceassembly/operator_admission.go`, `internal/app/server/runtime_assembly.go`; 6 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/jobs`, `uuid`, `pgx/v5`; std | lexical: `coordinator_jobs_integration_test.go`, `incident_retention_integration_test.go`; 4 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 119 lines; keep owner; §3 seam diagnosis |
| `job_terminal_effects_integration_test.go` | Shared test scenarios/fixtures: testJobTerminalOwnerRecovery, testRetainedTerminalJobRepair, testHistoricalAttemptCorruption | `ApplyJobTerminalEffectsTx` | Parent owner test scenarios (lexical): `coordinator_jobs_integration_test.go` | `internal/platform/jobs`, `internal/testutil/collaborationsupport`, `uuid`, `pgx/v5`, `pgx/v5/pgxpool`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 281 lines; test/support, not runtime architecture |
| `lifecycle_fixture_integration_test.go` | Assertions: Canonical Lifecycle Fixture Manifests | `TestCanonicalLifecycleFixtureManifests_Integration`, `ReferencedRegistryEntriesTx`, `RoundTrip`, `BeginReferenceOperation` | Go test runner; helper calls within owner tests | `internal/gen/contractreferencepackfixtures`, `internal/platform/jobs`, `uuid`, `pgx/v5`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification; reads `internal/gen/contractreferencepackfixtures` | Reference Data | medium | D; 306 lines; test/support, not runtime architecture |
| `limits.go` | Owner archive/reference limit projection and defaults | `ArchiveLimits`, `ReferenceLimits`, `Limits`, `DefaultLimits` | lexical: `internal/app/configassembly/deployment.go`, `internal/app/operator/operator_recovery.go`; 23 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/archivepolicy` | lexical: `attack_endpoint_verification_test.go`, `base_atomicity_integration_test.go`; 28 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 37 lines; keep owner; §3 seam diagnosis |
| `mixed_refresh_integration_test.go` | Assertions: Mixed Refresh Commits Cohort Fallback And Replay | `TestMixedRefreshCommitsCohortFallbackAndReplay_Integration` | Go test runner; helper calls within owner tests | `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/platform/telemetry/testsupport`, `internal/testutil/appsupport`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 183 lines; test/support, not runtime architecture |
| `object_collection.go` | Retires unneeded object references then collects unreferenced published objects safely | `CollectUnreferencedObjects` | lexical: `internal/app/operator/operator_reference_pack.go`, `object_collector.go` | `internal/platform/postgres`, `pgx/v5`; std | lexical: `base_atomicity_integration_test.go`, `coordinator_jobs_integration_test.go`; 4 candidate test files | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 101 lines; keep owner; §3 seam diagnosis |
| `object_collection_integration_test.go` | Shared test scenarios/fixtures: testRemovedObjectCollection | Package-private: `testRemovedObjectCollection` | Parent owner test scenarios (lexical): `coordinator_jobs_integration_test.go` | std only | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 81 lines; test/support, not runtime architecture |
| `object_collection_storage_integration_test.go` | Shared test scenarios/fixtures: testPublicationCollectionAndBackupGuards | `BeginTx`, `Commit` | Parent owner test scenarios (lexical): `base_atomicity_integration_test.go`, `canonical_persistence_integration_test.go` | `internal/app/recoveryassembly`, `internal/app/referenceassembly`, `internal/modules/recovery`, `internal/modules/reference_data`, `internal/platform/postgres`, `pgx/v5`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 116 lines; test/support, not runtime architecture |
| `object_collector.go` | Application-managed maintenance lifecycle for collection with readiness/serving guards | `ObjectCollector`, `NewObjectCollector`, `Sweep`, `Prepare`, `Close` | lexical: `internal/app/server/runtime_assembly.go`, `internal/modules/indicators/internal/providers/incidentbundle/source_port.go` | `internal/platform/postgres`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 95 lines; keep owner; §3 seam diagnosis |
| `observation.go` | Bounded semantic operation observation and consumer outcome classification | `OperationObserver`, `ReferenceOperationNames`, `ResolveCurrentPackSet`, `GetPackEntry`, `LookupPackEntries`, `EvaluateIndicatorValue`; +1 declarations | lexical: `internal/app/referenceassembly/telemetry.go`, `internal/modules/indicators/internal/identity/identity.go`; 14 total candidate files | `internal/platform/jobs`; std | lexical: `attack_native_integration_test.go`, `canonical_persistence_integration_test.go`; 9 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 97 lines; keep owner; §3 seam diagnosis |
| `openapi_contract_test.go` | Assertions: Open APIAnd Error Registries Expose Closed Reference Pack Contract | `TestOpenAPIAndErrorRegistriesExposeClosedReferencePackContract_Unit` | Go test runner; helper calls within owner tests | `internal/gen/contractreferencepacks`, `internal/platform/contracttest`; std | This file's assertions/helpers; C-19 | C-01–09, C-10: HTTP/Jobs projections; reads `internal/gen/contractreferencepacks` | Reference Data | medium | D; 230 lines; test/support, not runtime architecture |
| `operation_types.go` | Public action/accepted-Job results and owner catalog Job identity | `JobAcceptedResult`, `ActionParams`, `ActionResult` | lexical: `administration.go`, `administrative_actions.go`; 6 total candidate files | `internal/platform/jobs`, `uuid`; std | lexical: `base_atomicity_integration_test.go`, `base_reconciliation_integration_test.go`; 6 candidate test files | C-01–09, C-10: HTTP/Jobs projections | Reference Data | high | D; 43 lines; keep owner; §3 seam diagnosis |
| `operator_contract.go` | Validates a single deployment-local operator bundle filename | `ValidOperatorBundleName` | lexical: `internal/app/operator/operator_reference_pack.go`, `internal/app/referenceassembly/storage.go` | std only | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 17 lines; keep owner; §3 seam diagnosis |
| `operator_import_integration_test.go` | Assertions: Local Operator Import Uses Deployment Worker | `TestLocalOperatorImportUsesDeploymentWorker_Integration` | Go test runner; helper calls within owner tests | `internal/app/operator`, `internal/modules/reference_data`, `internal/testutil/appsupport`, `internal/testutil/configtest`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 87 lines; test/support, not runtime architecture |
| `portable_admission.go` | Frozen destination cohort admission and deterministic dependency scheduling | `Len`, `Less`, `Swap`, `Push`, `Pop` | lexical: `incident_retention.go`, `portable_publication.go`; 3 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `uuid`, `pgx/v5`; std | lexical: `portable_admission_integration_test.go`, `portable_admission_test.go` | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 369 lines; keep owner; §3 seam diagnosis |
| `portable_admission_integration_test.go` | Shared test scenarios/fixtures: testPortablePreparationAdmission | Package-private: `testPortablePreparationAdmission` | Parent owner test scenarios (lexical): `incident_retention_integration_test.go` | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `internal/platform/jobs`, `internal/platform/jobs/jobcontract`, `internal/testutil/collaborationsupport`, `uuid`, `pgx/v5/pgxpool`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 477 lines; test/support, not runtime architecture |
| `portable_admission_test.go` | Shared test scenarios/fixtures: testPortableVerificationOrder | Package-private: `testPortableVerificationOrder` | Parent owner test scenarios (lexical): `portable_input_test.go` | `internal/modules/reference_data/internal/packformat`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 46 lines; test/support, not runtime architecture |
| `portable_execution.go` | Parent incident Job execution budget, cancellation/absence classification and abort transaction | `IncidentReferenceExecution`, `BeginImportExecution`, `Context`, `Close`, `ClassifyAbsent`, `FinalizationContext`; +1 declarations | lexical: `internal/modules/incidentbundles/import_execution_failure.go`, `internal/modules/incidentbundles/source.go`; 8 total candidate files | `internal/modules/extensions/deadline`, `internal/platform/jobs`, `uuid`, `pgx/v5`; std | lexical: `api_test.go`, `attack_native_integration_test.go`; 17 candidate test files | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 163 lines; keep owner; §3 seam diagnosis |
| `portable_execution_test.go` | Shared test scenarios/fixtures: testPortableExecutionDeadline | Package-private: `testPortableExecutionDeadline` | Parent owner test scenarios (lexical): `portable_input_test.go` | `internal/platform/jobs`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 57 lines; test/support, not runtime architecture |
| `portable_history.go` | Validates persisted portable preparation history | Package-private: `validatePortablePreparations` | lexical: `historical_validation.go` | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `uuid`, `pgx/v5`; std | lexical: `portable_admission_integration_test.go` | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 145 lines; keep owner; §3 seam diagnosis |
| `portable_input.go` | Bounded embedded container admission into opaque private storage | `Read` | lexical: `incident_retention.go`, `portable_admission.go`; 3 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `portable_admission_integration_test.go`, `portable_admission_test.go`; 3 candidate test files | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 127 lines; keep owner; §3 seam diagnosis |
| `portable_input_test.go` | Shared test scenarios/fixtures: testPortableInputAdmission | `Close`, `Read`, `RemoveStaged`, `RemovePublished` | Parent owner test scenarios (lexical): `coordinator_admission_integration_test.go`, `portable_admission_integration_test.go` | `internal/modules/reference_data/internal/packformat`; std | This file's assertions/helpers; C-19 | C-16, C-17: retention/portability artifacts | Reference Data | medium | D; 184 lines; test/support, not runtime architecture |
| `portable_publication.go` | Publishes prepared portable verification/rejection facts under ordered guards | Package-private: `publishPortableAttemptTx`, `publishPortableRejectionsTx`, `publishRequiredPortableRejectionTx` | lexical: `incident_retention.go`, `portable_execution.go` | `internal/modules/reference_data/internal/packformat`, `pgx/v5`; std | lexical: `portable_admission_integration_test.go` | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 237 lines; keep owner; §3 seam diagnosis |
| `portable_verification.go` | Runs shared verifier with destination portable identity/dependency strategy | Package-private: `resolveDependency`, `checkReleaseSequence`, `verifyPortablePreparation` | lexical: `coordinator_identity.go`, `historical_identity.go`; 5 total candidate files | `internal/modules/reference_data/internal/packformat`, `uuid`, `pgx/v5`; std | lexical: `portable_admission_integration_test.go`, `verification_engine_test.go` | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 133 lines; keep owner; §3 seam diagnosis |
| `prepared_test.go` | Assertions: Prepared Result Closed Shape And Bindings | `TestPreparedResultClosedShapeAndBindings_Unit` | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `uuid`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 193 lines; test/support, not runtime architecture |
| `producer/attack_projection_test.go` | Shared test scenarios/fixtures: assertATTACKProducerExcludedFromRuntime | Package-private: `assertATTACKProducerExcludedFromRuntime` | Parent owner test scenarios (lexical): `producer/source_test.go` | `internal/gen/contractreferencepacks`; std | This file's assertions/helpers; C-19 | C-18: producer artifacts; no runtime-generated output; reads `internal/gen/contractreferencepacks` | Reference Data offline producer | medium | D; 17 lines; test/support, not runtime architecture |
| `producer/audit.go` | Offline ATTACK source field accounting, relationship/alias diagnostics and audit reports | `FieldRule`, `Policy`, `Locator`, `FieldStats`, `DomainStats`, `AliasCollision`; +4 declarations | lexical: `tools/reference-packs/main.go`, `producer/candidate.go`; 21 total candidate files | `internal/platform/canonicaljson`, `golang.org/x/text/unicode/norm`; std | lexical: `producer/entra_test.go`, `producer/source_test.go`; 4 candidate test files | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 581 lines; offline boundary, F-06 |
| `producer/base.go` | Offline reissue of release-bound Base registry artifacts | `BaseReissueInput`, `ReissueBase` | lexical: `tools/reference-packs/main.go` | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 114 lines; offline boundary, F-06 |
| `producer/candidate.go` | Offline native ATTACK candidate schema compilation and validation | `CandidateIndex`, `CandidateValidation`, `ReadCandidate`, `CompileCandidate`, `ValidateNativeCandidate` | lexical: `tools/reference-packs/main.go`, `producer/audit.go`; 3 total candidate files | `internal/platform/canonicaljson`, `github.com/santhosh-tekuri/jsonschema/v6`; std | lexical: `producer/candidate_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 136 lines; offline boundary, F-06 |
| `producer/candidate_test.go` | Assertions: ATTACKNative Candidate | `TestATTACKNativeCandidate` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | medium | D; 64 lines; test/support, not runtime architecture |
| `producer/entra_package.go` | Assembles Entra canonical payload, manifest and package report | `EntraBuildInput`, `EntraPackageReport`, `AssembleEntra` | lexical: `tools/reference-packs/main.go` | `contracts/reference-pack-producers/entra-app-ids`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 160 lines; offline boundary, F-06 |
| `producer/entra_projection.go` | Parses pinned Entra source tables and reconciles per-field provenance | `EntraAssertion`, `EntraFieldAccounting`, `EntraAudit`, `EntraProjection`, `EntraReconcile`, `EntraAuditSources` | lexical: `tools/reference-packs/main.go`, `producer/entra_package.go` | `contracts/reference-pack-producers/entra-app-ids`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `golang.org/x/text/unicode/norm`; std | lexical: `producer/entra_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 504 lines; offline boundary, F-06 |
| `producer/entra_source.go` | Acquires/checks immutable Entra upstream source identity and digest | `EntraSource`, `EntraInventory`, `EntraAcquired`, `EntraLock`, `EntraDownload`, `EntraFetcher`; +4 declarations | lexical: `tools/reference-packs/main.go`, `producer/entra_package.go`; 3 total candidate files | `contracts/reference-pack-producers/entra-app-ids`, `internal/platform/canonicaljson`; std | lexical: `producer/entra_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 287 lines; offline boundary, F-06 |
| `producer/entra_test.go` | Assertions: Entra Source Boundaries; Entra Source Integrity | `TestEntraSourceBoundaries_Unit`, `FetchEntra`, `TestEntraSourceIntegrity_Unit` | Go test runner; helper calls within owner tests | `contracts/reference-pack-producers/entra-app-ids`; std | This file's assertions/helpers; C-19 | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | medium | D; 107 lines; test/support, not runtime architecture |
| `producer/envelope.go` | Offline signing contexts, fixture separation and TUF envelope publication | `SigningContext`, `EnvelopeInput`, `EnvelopeReport`, `NewFixtureSigningContext`, `NewWindowsFixtureSigningContext`, `NewEntraFixtureSigningContext`; +2 declarations | lexical: `testsupport/attack_fixture.go`, `testsupport/entra_fixture.go`; 7 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | lexical: `producer/windows_online_test.go`, `internal/modules/incidentbundles/reference_attack_round_trip_integration_test.go`; 14 candidate test files | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 442 lines; offline boundary, F-06 |
| `producer/identity.go` | Build identity and isolated publication suffix creation | `Builder`, `BuilderIdentity` | lexical: `testsupport/windows_fixture.go`, `tools/reference-packs/main.go`; 14 total candidate files | std only | lexical: `producer/source_test.go`, `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 44 lines; offline boundary, F-06 |
| `producer/ledger.go` | Locked publication allocation, continuity and collision detection | `Publication` | lexical: `producer/base.go`, `producer/entra_package.go`; 5 total candidate files | `internal/platform/canonicaljson`, `golang.org/x/sys/unix`; std | lexical: `producer/ledger_test.go`, `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 119 lines; offline boundary, F-06 |
| `producer/ledger_test.go` | Assertions: ATTACKPublication Ledger | `TestATTACKPublicationLedger` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | medium | D; 37 lines; test/support, not runtime architecture |
| `producer/package.go` | Offline native ATTACK canonical package assembly | `PackageInput`, `PackageReport`, `Open`, `Assemble` | lexical: `tools/reference-packs/main.go`, `producer/base.go`; 9 total candidate files | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 316 lines; offline boundary, F-06 |
| `producer/publication.go` | Atomically publishes offline producer tree and logical member inventory | Package-private: `publishTree`, `logicalMemberInventory` | lexical: `producer/entra_package.go`, `producer/envelope.go`; 6 total candidate files | `internal/modules/reference_data/internal/packformat`, `golang.org/x/sys/unix`; std | lexical: `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 108 lines; offline boundary, F-06 |
| `producer/release_verification.go` | Verifies assembled release envelope through runtime-compatible owner verifier | `Write` | lexical: `producer/envelope.go` | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 71 lines; offline boundary, F-06 |
| `producer/source.go` | Pinned ATTACK acquisition, retained source evidence and digest/URL/path admission | `SourceLock`, `SourceArtifact`, `EvidenceFile`, `Acquisition`, `Fetcher`, `HTTPFetcher`; +8 declarations | lexical: `testsupport/attack_fixture.go`, `testsupport/entra_fixture.go`; 26 total candidate files | `internal/platform/canonicaljson`; std | lexical: `producer/candidate_test.go`, `producer/ledger_test.go`; 6 candidate test files | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 470 lines; offline boundary, F-06 |
| `producer/source_test.go` | Assertions: ATTACKSource Fixtures; ATTACKSource Retention Boundaries; ATTACKSource Contract Closure | `Fetch`, `TestATTACKSourceFixtures`, `TestATTACKSourceRetentionBoundaries`, `TestATTACKSourceContractClosure` | Go test runner; helper calls within owner tests | `contracts/reference-pack-producers/attack`; std | This file's assertions/helpers; C-19 | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | medium | D; 241 lines; test/support, not runtime architecture |
| `producer/test_helpers_test.go` | Shared test scenarios/fixtures: testRoot | Package-private: `encoded`, `testRoot` | Parent owner test scenarios (lexical): `producer/candidate_test.go`, `producer/source_test.go` | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | medium | D; 25 lines; test/support, not runtime architecture |
| `producer/windows_acquire.go` | Bounded Windows source discovery/acquisition and release pin verification | `WindowsDownload`, `WindowsFetcher`, `WindowsHTTPFetcher`, `FetchWindows`, `WindowsAcquire`, `WindowsCheck`; +1 declarations | lexical: `testsupport/windows_fixture.go`, `tools/reference-packs/main.go`; 6 total candidate files | `internal/platform/canonicaljson`, `golang.org/x/net/html`; std | lexical: `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 425 lines; offline boundary, F-06 |
| `producer/windows_catalog.go` | Extracts Windows event catalog records and source section/query identities | Package-private: `catalogRecord`, `extractWindowsCatalog`, `catalogQueryIDs` | lexical: `producer/windows_reconcile.go`, `producer/windows_sysmon.go` | `golang.org/x/net/html`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 186 lines; offline boundary, F-06 |
| `producer/windows_contract.go` | Windows producer input, provenance, coverage, diagnostics and closed contracts | `WindowsError`, `WindowsSource`, `WindowsDiscovery`, `WindowsInventory`, `WindowsAcquired`, `WindowsLock`; +12 declarations | lexical: `testsupport/windows_fixture.go`, `tools/reference-packs/main.go`; 14 total candidate files | `contracts/reference-pack-producers/windows-event-ids`, `internal/platform/canonicaljson`, `github.com/santhosh-tekuri/jsonschema/v6`; std | lexical: `producer/windows_csv_test.go`, `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 274 lines; offline boundary, F-06 |
| `producer/windows_csv.go` | Extracts owner-specific Windows CSV assertions and exact source locators | `WindowsAssertion` | lexical: `producer/windows_contract.go`, `producer/windows_csv_documentation.go`; 5 total candidate files | `internal/platform/canonicaljson`; std | lexical: `producer/windows_csv_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 600 lines; offline boundary, F-06 |
| `producer/windows_csv_documentation.go` | Projects CSV documentation references and optional fields | Package-private: `csvDocumentation`, `selectCSV`, `csvNullable` | lexical: `producer/windows_csv.go` | `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 141 lines; offline boundary, F-06 |
| `producer/windows_csv_test.go` | Shared test scenarios/fixtures: testWindowsCSVReferences | Package-private: `csvFixtureSources`, `rewriteCSV`, `testWindowsCSVReferences` | Parent owner test scenarios (lexical): `producer/windows_online_test.go` | std only | This file's assertions/helpers; C-19 | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | medium | D; 174 lines; test/support, not runtime architecture |
| `producer/windows_distribution.go` | Separates Windows public facts from original documentation evidence | Package-private: `windowsPublicFacts`, `windowsOriginalDescription`, `windowsCoalesceDocumentation` | lexical: `producer/windows_reconcile.go` | `internal/platform/canonicaljson`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 150 lines; offline boundary, F-06 |
| `producer/windows_extract.go` | Extracts event records and documentation from pinned XML/HTML inputs | Package-private: `parseWindowsXML`, `xmlChildren`, `xmlDescendants` | lexical: `producer/windows_catalog.go`, `producer/windows_csv.go`; 7 total candidate files | `golang.org/x/net/html`, `golang.org/x/text/unicode/norm`; std | lexical: `producer/windows_csv_test.go`, `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 678 lines; offline boundary, F-06 |
| `producer/windows_local.go` | Admits local immutable Windows source files through producer fetch port | `WindowsLocalInput`, `WindowsLocalFetcher`, `FetchWindows` | lexical: `testsupport/windows_fixture.go`, `tools/reference-packs/main.go`; 5 total candidate files | `internal/platform/canonicaljson`; std | lexical: `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 203 lines; offline boundary, F-06 |
| `producer/windows_online_test.go` | Assertions: Windows Online Sources; Windows Online Projection; Windows Online Publication; 1 additional test entries | `RoundTrip`, `FetchWindows`, `TestWindowsOnlineSources`, `TestWindowsOnlineProjection`, `TestWindowsOnlinePublication`, `TestWindowsOnlineFixtureTrust` | Go test runner; helper calls within owner tests | `contracts/reference-pack-producers/windows-event-ids`, `internal/modules/reference_data/internal/packformat`; std | This file's assertions/helpers; C-19 | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | medium | D; 361 lines; test/support, not runtime architecture |
| `producer/windows_ooxml.go` | Reads bounded Office archive members and source cell/range locators | Package-private: `windowsOfficeMembers`, `officeText`, `readWindowsOffice` | lexical: `producer/candidate.go`, `producer/windows_catalog.go`; 8 total candidate files | std only | lexical: `producer/windows_csv_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 237 lines; offline boundary, F-06 |
| `producer/windows_package.go` | Assembles Windows event package with notice/licensing decisions | `WindowsBuildInput`, `WindowsNotice`, `WindowsLicenseDecision`, `WindowsPackageReport`, `AssembleWindows` | lexical: `testsupport/windows_fixture.go`, `tools/reference-packs/main.go` | `contracts/reference-pack-producers/windows-event-ids`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | lexical: `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 223 lines; offline boundary, F-06 |
| `producer/windows_provider.go` | Applies declared provider exclusions and validates source-provider bindings | Package-private: `applyWindowsProviderExclusions`, `validateWindowsProviderBindings`, `providerFromWindowsBinding` | lexical: `producer/windows_csv.go`, `producer/windows_reconcile.go` | std only | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 124 lines; offline boundary, F-06 |
| `producer/windows_reconcile.go` | Reconciles Windows assertions, presentation and provenance accounting | `WindowsProjection`, `WindowsReconcile`, `WindowsAudit` | lexical: `tools/reference-packs/main.go`, `producer/windows_package.go` | `contracts/reference-pack-producers/windows-event-ids`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `golang.org/x/text/unicode/norm`; std | lexical: `producer/windows_csv_test.go`, `producer/windows_online_test.go` | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 391 lines; offline boundary, F-06 |
| `producer/windows_sysmon.go` | Extracts Sysmon publisher, manifest and document facts | Package-private: `windowsSysmonPublisher`, `extractWindowsSysmonManifest`, `extractWindowsSysmonDocument` | lexical: `producer/windows_extract.go`, `producer/windows_reconcile.go` | `golang.org/x/net/html`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-18: producer artifacts; no runtime-generated output | Reference Data offline producer | high | D; 146 lines; offline boundary, F-06 |
| `profile_consumers_integration_test.go` | Assertions: Host Registry Profile Consumer; Evidence Registry Profile Consumer; Indicator Registry Profile Consumer; 13 additional test entries | `TestHostRegistryProfileConsumer_Integration`, `TestEvidenceRegistryProfileConsumer_Integration`, `TestIndicatorRegistryProfileConsumer_Integration`, `TestATTACKProfileConsumer_Integration`, `TestD3FENDProfileConsumer_Integration`, `TestVERISProfileConsumer_Integration`; +10 declarations | Go test runner; helper calls within owner tests | `internal/testutil/httptestx`, `internal/app/referenceassembly`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/modules/reference_data/internal/packformat`, `internal/platform/pagination`, `internal/testutil/appsupport`; std | This file's assertions/helpers; C-19 | C-11, C-12: consumer/set/provenance projections | Reference Data | medium | D; 362 lines; test/support, not runtime architecture |
| `provenance.go` | Decodes canonical retained provenance bound to exact immutable set | Package-private: `decodeRetainedProvenance` | lexical: `canonical_repository.go`, `historical_validation.go`; 3 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `provenance_test.go` | C-11, C-12: consumer/set/provenance projections | Reference Data | high | D; 18 lines; keep owner; §3 seam diagnosis |
| `provenance_test.go` | Assertions: Retained Provenance Closed And Bound To Exact Set | `TestRetainedProvenanceClosedAndBoundToExactSet_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-11, C-12: consumer/set/provenance projections | Reference Data | medium | D; 159 lines; test/support, not runtime architecture |
| `public_types.go` | Owner DTOs and explicit translations shielding private format representations | `PackSetMember`, `PackSet`, `PackSourceArtifact`, `PackLicense`, `PackLicenseBinding`, `Evaluation`; +5 declarations | lexical: `internal/modules/indicators/internal/identity/identity.go`, `testsupport/indicators.go`; 27 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `base_atomicity_integration_test.go`, `consumer_test.go`; 12 candidate test files | C-11, C-12: consumer/set/provenance projections | Reference Data | high | D; 193 lines; keep owner; §3 seam diagnosis |
| `public_types_test.go` | Shared test scenarios/fixtures: testOwnerDTOCanonicalRoundTrips | Package-private: `testOwnerDTOCanonicalRoundTrips` | Parent owner test scenarios (lexical): `consumer_test.go` | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-11, C-12: consumer/set/provenance projections | Reference Data | medium | D; 74 lines; test/support, not runtime architecture |
| `readiness.go` | Checks retained required Base state for serving readiness | `CheckReadyState` | lexical: `internal/app/server/runtime_assembly.go`, `historical_validation.go` | `internal/platform/postgres`; std | lexical: `consumer_integrity_integration_test.go` | C-15, C-17: current contract/readiness | Reference Data | high | D; 44 lines; keep owner; §3 seam diagnosis |
| `recoverycontribution/contribution.go` | Declares Reference Data authoritative recovery table contribution | `RecoveryStateContribution` | lexical: `internal/app/recoveryassembly/state_catalog.go` | `internal/platform/recoverystate` | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-17: recovery contribution/catalog projection | Reference Data contribution; Recovery coordinates | high | D; 47 lines; keep owner; §3 seam diagnosis |
| `recoverycontribution/inventory.go` | Declares pack member-object inventory and restore family | `VNextRecoveryObjectInventory` | lexical: `internal/app/recoveryassembly/state_catalog.go` | `internal/modules/recovery`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-17: recovery contribution/catalog projection | Reference Data contribution; Recovery coordinates | high | D; 54 lines; keep owner; §3 seam diagnosis |
| `reference_pack_integration_test.go` | Assertions: Import List Read Replay And Job Summary; Activation Disable Reverify And Refresh Lifecycle; Failures Remain Inactive And No Network Is Needed; 6 additional test entries | `TestImportListReadReplayAndJobSummary_Integration`, `TestActivationDisableReverifyAndRefreshLifecycle_Integration`, `TestFailuresRemainInactiveAndNoNetworkIsNeeded_Integration`, `TestAdmissionQueuesBeforeVerificationAndCancelPreventsCommit_Integration`, `TestMinimumDisconnectedBundleSeededExactly_Integration`, `TestRefreshOmittedSelectorReplayUsesAdmittedSet_Integration`; +3 declarations | Go test runner; helper calls within owner tests | `internal/modules/evidence/testsupport/flowtest`, `internal/app/referenceassembly`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/entities/entitycontract`, `internal/modules/evidence`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/reference_data`, `internal/modules/timeline`, `internal/modules/timeline/testsupport/routetest`, `internal/platform/authn`, `internal/testutil/appsupport`, `internal/testutil/authcookietest`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 973 lines; test/support, not runtime architecture |
| `reference_pack_storage_schema_test.go` | Assertions: Reference Pack Storage Reference Head Schema Contract | `TestReferencePackStorageReferenceHeadSchemaContract_Integration` | Go test runner; helper calls within owner tests | `internal/modules/reference_data/contractadmission`, `internal/testutil/pgtest`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 25 lines; test/support, not runtime architecture |
| `reference_pack_test_helpers_test.go` | Shared test scenarios/fixtures: consumerIntegrityOptions, integrationCanonical, integrationDigest | Package-private: `consumerIntegrityOptions`, `integrationCanonical`, `integrationDigest` | Parent owner test scenarios (lexical): `attack_native_integration_test.go`, `canonical_persistence_integration_test.go` | `internal/app/extensionassembly`, `internal/app/server`, `internal/modules/reference_data`, `internal/modules/reference_data/internal/packformat`, `internal/platform/authn`, `internal/platform/canonicaljson`, `internal/testutil/appsupport`, `internal/testutil/authcookietest`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 407 lines; test/support, not runtime architecture |
| `registry_assignment.go` | Caller-transaction registry guards, exact evaluation and usage revision recording | `RegistryAssignments`, `RegistryAssignment`, `NewRegistryAssignments`, `BeginTx`, `RecordUsage` | lexical: `internal/app/incidentportabilityassembly/catalog.go`, `internal/app/referenceassembly/registry_transactions.go`; 10 total candidate files | `internal/platform/pagination`, `internal/platform/postgres`, `pgx/v5`; std | lexical: `canonical_persistence_integration_test.go`, `consumer_integrity_integration_test.go`; 3 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 104 lines; keep owner; §3 seam diagnosis |
| `requests.go` | Semantic action/import/refresh DTOs, rejection types and normalized route/hash helpers | `ImportMetadataRequest`, `ActionRequest`, `RefreshRequest`, `RequestRejection`, `ActivationRejection`, `Error` | lexical: `administrative_actions.go`, `coordinator.go`; 6 total candidate files | std only | lexical: `api_test.go` | C-01–09, C-10: HTTP/Jobs projections | Reference Data | high | D; 85 lines; keep owner; §3 seam diagnosis |
| `retention.go` | Snapshot/reporting current or historical exact set capture, validation and pins | `SetBinding`, `Retention`, `RetainedBindingValidator`, `NewRetention`, `CaptureCurrentTx`, `CaptureRetainedTx`; +2 declarations | lexical: `internal/app/server/runtime_assembly.go`, `internal/modules/incidentbundles/source.go`; 23 total candidate files | `internal/platform/canonicaljson`, `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `attack_native_integration_test.go`, `base_atomicity_integration_test.go`; 20 candidate test files | C-16, C-17: retention/portability artifacts | Reference Data | high | D; 147 lines; keep owner; §3 seam diagnosis |
| `retired_format_fixture_test.go` | Shared test scenarios/fixtures: retiredReferencePackBundle, addZipFile | Package-private: `retiredReferencePackBundle`, `addZipFile` | Parent owner test scenarios (lexical): `reference_pack_test_helpers_test.go`, `verification_engine_test.go` | std only | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 90 lines; test/support, not runtime architecture |
| `staging_rejection.go` | Distinguishes streamed container limit rejection from operational storage faults | `ImportContentRejection` | lexical: `internal/app/operator/operator_reference_pack.go`, `coordinator.go`; 3 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `staging_rejection_test.go` | C-01–09, C-10: HTTP/Jobs projections | Reference Data | high | D; 46 lines; keep owner; §3 seam diagnosis |
| `staging_rejection_test.go` | Shared test scenarios/fixtures: testStagingRejectionClassification | Package-private: `testStagingRejectionClassification` | Parent owner test scenarios (lexical): `api_test.go` | std only | This file's assertions/helpers; C-19 | C-01–09, C-10: HTTP/Jobs projections | Reference Data | medium | D; 44 lines; test/support, not runtime architecture |
| `storage_port.go` | Opaque root-free storage references and narrow artifact/workspace/collection capabilities | `StagingLimitError`, `StorageRef`, `StagingRef`, `VerificationStorage`, `ArtifactStorage`, `ContainerReader`; +8 declarations | lexical: `internal/app/referenceassembly/artifact_storage.go`, `internal/app/referenceassembly/object_collection.go`; 32 total candidate files | `golang.org/x/text/unicode/norm`; std | lexical: `attack_endpoint_verification_test.go`, `base_atomicity_integration_test.go`; 15 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 151 lines; keep owner; §3 seam diagnosis |
| `storage_port_test.go` | Assertions: Reference Pack Storage References Are Strict And Root Free | `TestReferencePackStorageReferencesAreStrictAndRootFree_Unit` | Go test runner; helper calls within owner tests | std only | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 40 lines; test/support, not runtime architecture |
| `successful_envelope.go` | Decodes successful canonical envelopes into retained owner state | Package-private: `decodeSuccessfulEnvelope` | lexical: `administration_repository.go`, `attestation.go`; 9 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `successful_envelope_test.go` | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 29 lines; keep owner; §3 seam diagnosis |
| `successful_envelope_test.go` | Assertions: Successful Envelope Closed Persistence | `TestSuccessfulEnvelopeClosedPersistence_Unit` | Go test runner; helper calls within owner tests | `internal/platform/canonicaljson`, `uuid`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 132 lines; test/support, not runtime architecture |
| `testsupport/attack_fixture.go` | Constructs authentic ATTACK pack fixture from canonical owner artifacts | `AuthenticATTACKPack`, `AuthenticATTACKFixture` | Owner package/interface dispatch; no external lexical caller identified | `internal/modules/reference_data/internal/packformat`, `internal/modules/reference_data/producer`, `internal/testutil/suiteservices`; std | lexical: `internal/modules/incidentbundles/reference_attack_round_trip_integration_test.go`, `attack_native_integration_test.go`; 4 candidate test files | C-19: canonical fixture inputs; no generated output | Reference Data test support | medium | D; 93 lines; test-only boundary, F-07 |
| `testsupport/entra_expectations.go` | Checks independent Entra corpus baseline expectations | Package-private: `checkEntraBaseline` | lexical: `testsupport/entra_fixture.go` | `contracts/reference-pack-producers/entra-app-ids`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | lexical: `testsupport/entra_expectations_test.go` | C-19: canonical fixture inputs; no generated output | Reference Data test support | medium | D; 148 lines; test-only boundary, F-07 |
| `testsupport/entra_expectations_test.go` | Assertions: Entra Corpus Expectations | `TestEntraCorpusExpectations_Unit` | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | This file's assertions/helpers; C-19 | C-19: canonical fixture inputs; no generated output | Reference Data test support | medium | D; 88 lines; test/support, not runtime architecture |
| `testsupport/entra_fixture.go` | Builds deterministic Entra fixture release and expected set digests | `ProducedEntraPack`, `EntraSetDigest`, `EntraRelease`, `EntraFixture` | Owner package/interface dispatch; no external lexical caller identified | `internal/modules/reference_data/internal/packformat`, `internal/modules/reference_data/producer`, `internal/platform/canonicaljson`, `internal/testutil/suiteservices`; std | lexical: `testsupport/entra_expectations_test.go`, `internal/modules/incidentbundles/reference_entra_round_trip_integration_test.go`; 5 candidate test files | C-19: canonical fixture inputs; no generated output | Reference Data test support | medium | D; 166 lines; test-only boundary, F-07 |
| `testsupport/fixturehttp/http.go` | Test-only authenticated Reference Pack fixture import over HTTP | `ImportPackFixture` | Owner package/interface dispatch; no external lexical caller identified | `internal/modules/reference_data/testsupport`, `internal/platform/authn`, `internal/testutil/authcookietest`, `internal/testutil/httptestx`; std | lexical: `internal/modules/incidentbundles/routes_round_trip_integration_test.go`, `internal/modules/recovery/backup_metadata_integration_test.go`; 3 candidate test files | C-19: canonical fixture inputs; no generated output | Reference Data test support | medium | D; 64 lines; test-only boundary, F-07 |
| `testsupport/indicators.go` | Test-only registry/consumer capability for isolated indicator tests | `IndicatorRegistry`, `EvaluateIndicator`, `BeginTx`, `RecordUsage`, `ResolveCurrentPackSet`, `EvaluateIndicatorValue`; +3 declarations | lexical: `internal/modules/indicators/testsupport/fixtures.go`, `internal/testutil/appsupport/performancefixture/owners.go`; 4 total candidate files | `internal/modules/reference_data`, `internal/modules/reference_data/internal/packformat`, `pgx/v5`; std | lexical: `internal/app/importassembly/owner_registry_test.go`, `internal/app/importassembly/tasksdecisions_integration_test.go`; 29 candidate test files | C-19: canonical fixture inputs; no generated output | Reference Data test support | medium | D; 64 lines; test-only boundary, F-07 |
| `testsupport/owner_fixture.go` | Reads machine-readable owner fixture expectations | `OwnerFixtureExpectation`, `OwnerFixture` | Owner package/interface dispatch; no external lexical caller identified | `internal/gen/contractreferencepackfixtures`; std | lexical: `internal/modules/incidentbundles/reference_entra_round_trip_integration_test.go`, `internal/modules/incidentbundles/routes_reference_verification_integration_test.go`; 10 candidate test files | C-19: canonical fixture inputs; no generated output; reads `internal/gen/contractreferencepackfixtures` | Reference Data test support | medium | D; 38 lines; test-only boundary, F-07 |
| `testsupport/portable_fixture.go` | Builds exact portable pack/catalog/cohort fixtures | `PortablePackFixture`, `PortableFixture`, `PortableCohortFixture` | lexical: `testsupport/fixturehttp/http.go`, `testsupport/attack_fixture.go`; 4 total candidate files | `internal/modules/reference_data`, `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`; std | lexical: `internal/modules/incidentbundles/reference_attack_round_trip_integration_test.go`, `internal/modules/incidentbundles/reference_entra_round_trip_integration_test.go`; 13 candidate test files | C-19: canonical fixture inputs; no generated output | Reference Data test support | medium | D; 307 lines; test-only boundary, F-07 |
| `testsupport/windows_fixture.go` | Builds deterministic Windows fixture package and set expectations | `ProducedWindowsPack`, `WindowsSetDigest`, `FetchWindows`, `WindowsFixture` | lexical: `testsupport/entra_fixture.go` | `contracts/reference-pack-producers/windows-event-ids`, `internal/modules/reference_data/internal/packformat`, `internal/modules/reference_data/producer`, `internal/platform/canonicaljson`, `internal/testutil/suiteservices`; std | lexical: `internal/modules/incidentbundles/reference_windows_round_trip_integration_test.go`, `internal/modules/recovery/reference_windows_integration_test.go`; 4 candidate test files | C-19: canonical fixture inputs; no generated output | Reference Data test support | medium | D; 225 lines; test-only boundary, F-07 |
| `trust_repository.go` | Reconciles admitted trust bootstrap with retained repository state | `ReconcileTrustBootstrap` | lexical: `internal/app/server/runtime_assembly.go` | `internal/modules/reference_data/internal/packformat`, `internal/platform/canonicaljson`, `internal/platform/postgres`, `uuid`, `pgx/v5`; std | lexical: `coordinator_admission_integration_test.go`, `coordinator_jobs_integration_test.go`; 3 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 81 lines; keep owner; §3 seam diagnosis |
| `upload_limit_integration_test.go` | Shared test scenarios/fixtures: testOversizedUploadAdmission | Package-private: `testOversizedUploadAdmission` | Parent owner test scenarios (lexical): `reference_pack_integration_test.go` | `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/testutil/appsupport`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 75 lines; test/support, not runtime architecture |
| `validation_summaries.go` | Reads immutable diagnostic resources and validates historical diagnostic linkage | `GetValidationSummary` | lexical: `administration.go`, `historical_validation.go`; 3 total candidate files | `internal/modules/reference_data/internal/packformat`, `uuid`, `pgx/v5`; std | No direct named test cross-reference; indirect contract tests in §4; TODO: confirm before changing | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 75 lines; keep owner; §3 seam diagnosis |
| `verification_diagnostics.go` | Adapts workspace scratch and retained findings into shared diagnostic collection | `Write`, `Open`, `Remove` | lexical: `internal/app/referenceassembly/object_collection.go`, `internal/app/referenceassembly/verification_storage.go`; 11 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `concurrent_publication_integration_test.go`, `coordinator_admission_integration_test.go`; 16 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 75 lines; keep owner; §3 seam diagnosis |
| `verification_engine.go` | Single canonical verifier entry, strategy/capability orchestration and typed verdicts | `OperationRejection`, `ContentRejection`, `Close`, `Error`, `Read` | lexical: `internal/app/operator/operator_reference_pack.go`, `internal/modules/incidentbundles/import_execution_failure.go`; 29 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `api_test.go`, `attack_endpoint_verification_test.go`; 9 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 197 lines; keep owner; §3 seam diagnosis |
| `verification_engine_test.go` | Assertions: Canonical Verification Engine Independent Signed Container; Retained Content Checks Follow Verification Precedence; Verifier Never Turns Operational Reads Into Content Verdicts; 3 additional test entries | `Close`, `OpenStaged`, `OpenPublished`, `NewWorkspace`, `Write`, `Open`; +8 declarations | Go test runner; helper calls within owner tests | `internal/modules/reference_data/internal/packformat`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 702 lines; test/support, not runtime architecture |
| `verification_precedence_integration_test.go` | Assertions: Live Verification Identity Precedes Content And Runtime Checks; Live Verification Resolves Exact Dependencies Before Indexing | `TestLiveVerificationIdentityPrecedesContentAndRuntimeChecks_Integration`, `TestLiveVerificationResolvesExactDependenciesBeforeIndexing_Integration` | Go test runner; helper calls within owner tests | `internal/modules/auth/testsupport/flowtest`, `internal/testutil/appsupport`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 185 lines; test/support, not runtime architecture |
| `verification_program.go` | Builds ordered live verification checks over archive, identity, trust and content | Package-private: `checks`, `retained`, `archiveChecks` | lexical: `administrative_actions.go`, `base_reconciliation.go`; 31 total candidate files | `internal/modules/reference_data/internal/packformat`; std | lexical: `attack_endpoint_verification_test.go`, `base_atomicity_integration_test.go`; 28 candidate test files | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 367 lines; keep owner; §3 seam diagnosis |
| `verification_repository.go` | Loads frozen member/trust inputs for execution | Package-private: `frozenMember`, `frozenTrust` | lexical: `administrative_actions.go`, `coordinator_execution.go`; 7 total candidate files | `internal/modules/reference_data/internal/packformat`, `pgx/v5`; std | lexical: `coordinator_admission_integration_test.go`, `job_terminal_effects_integration_test.go` | C-12–14: lifecycle/identity/verification | Reference Data | high | D; 84 lines; keep owner; §3 seam diagnosis |
| `windows_online_integration_test.go` | Assertions: Windows Online Production Lifecycle | `TestWindowsOnlineProductionLifecycle_Integration` | Go test runner; helper calls within owner tests | `internal/app/referenceassembly`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/modules/reference_data/producer`, `internal/modules/reference_data/testsupport`, `internal/platform/pagination`, `internal/testutil/appsupport`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 138 lines; test/support, not runtime architecture |
| `worker_isolation_integration_test.go` | Assertions: Worker Database Barriers Are Instance Scoped | `TestWorkerDatabaseBarriersAreInstanceScoped_Integration` | Go test runner; helper calls within owner tests | `internal/app/server`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/reference_data`, `internal/testutil/appsupport`, `internal/testutil/httptestx`; std | This file's assertions/helpers; C-19 | C-12–14: lifecycle/identity/verification | Reference Data | medium | D; 55 lines; test/support, not runtime architecture |

## 3. Module Boundary Diagnosis

LC retains the runtime/neutral-library/producer boundary. Its current responsibility decisions and intentional no-action findings are in the LC contract and finding tables; the earlier diagnoses below retain their original scope.

I8 keeps the existing runtime/private-format/offline-producer boundaries. Split obsolete Base cutover execution from maintained producer responsibilities by retiring it, not by moving it into a compatibility package. Keep current release loading and all owner-required Base verification. Simplify the owner-local rejection fixture without creating another fixture framework. A future general Base builder is deferred; no new public facade, package subdivision or platform abstraction is justified by this slice. E-I8-02–04 support F-33/34 and C-13/15/18/19.

Historical I6 diagnosis follows:

I6 retains the subsystem boundary. Split the live producer schema utility from proposal-specific execution, consolidate adopted input admission, and retire the obsolete proposal path. The offline producer remains Reference Data-owned; no platform helper package or generic plugin abstraction is warranted. F-26–29 supersede the earlier blanket no-action conclusion only for the named ATT&CK producer seams. VERIS joins the existing source-specific producer responsibilities; its implementation is not refactored here.

The target is a legitimate subsystem containing an application/mutation coordinator, persistence adapters, a transport-adjacent adapter and private format/state implementations. It is more than a thin service facade, but the inspected responsibilities have coherent Reference Pack semantics. Its size alone does not establish an accidental catch-all. No frontend shell, grid vendor implementation, generic workbook row/view-schema implementation or generic projection-refresh coordinator was found in the target.

This is an architectural finding, not a decision to preserve every file forever. Keep/move/split below are planning dispositions. The narrow authorization seam in F-01 is corrected by S-02; the inspection does not justify a mass extraction to `platform`, `runtime` or another domain module.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Canonical pack identity, hostile-content verification, diagnostic ordering and TUF | `internal/packformat`, `verification_engine.go`, `verification_program.go` | Reference Data private verification/format boundary | keep | O-01 §§6–18; E-02; C-12/13 | Shared canonical JSON and generated schema access are technical dependencies, not ownership of incident data. Verification check ranks express owner semantics, not harness phases. |
| Condition projection and effective-set resolution | `internal/packstate/state.go` | Reference Data private state support | split | O-01 §§11, 19–20; E-02; F-19 | Keep live projection/fallback and immutable copies; retire unused parallel lifecycle algorithms after transferring their assertions to live coordinator tests (S-11). |
| Lifecycle admission/execution/publication | `coordinator*.go`, `dependency_admission.go`, `job_*.go` | Reference Data application service; common Jobs/Extensions provide execution/finalization | keep | O-01 RP-REQ-175/176; E-03/04; C-14 | SQL transactions and Jobs DTOs in orchestration are not evidence that pure state logic moved into platform. |
| Administrative SQL and canonical indexes | `administration_repository.go`, `canonical_repository.go`, `canonical_index.go`, `canonical_publication.go` | Reference Data private persistence adapters | keep | E-02; C-11/12/14 | Preserve owner table writes and transaction boundaries. No new generic repository facade is justified merely by direct SQL. |
| Administrative HTTP admission and mapping | `http_adapter.go`, `http_requests.go` | Reference Data transport adapter composed by application; shared auth/HTTP primitives | split | E-04; O-03; F-01; C-22 | Candidate separation is family-wide authorization from route-specific handling. Implemented as owner-supplied family policy through shared HTTP routing; no whole-module relocation. |
| Consumer lookup, algorithms and immutable provenance | `consumer.go`, `public_types.go`, `provenance.go`, private format/profile files | Reference Data consumer API | keep | O-01 §21; O-04; E-05 | Indicator algorithms belong here; Indicators retains source identity, records, revisions and user-facing mutation outcomes. |
| Registry assignment and source-owner usage | `registry_assignment.go`, `administrative_actions.go`; application `RegistryUsage` and `RegistryTransactions` | Reference Data guards/evaluation; Entities/Evidence/Indicators supply owned usage | keep | E-03/05; C-21 | Preserve guard order and rollback-before-invalidation. No permission to read/write peer source tables directly. |
| Retention, incident reference catalogs and recovery contribution | `retention.go`, `incident_*.go`, `portable_*.go`, `byte_retention.go`, `recoverycontribution` | Reference Data owns pack facts; Reporting, Incident Bundles and Recovery coordinate their workflows | keep | O-01 §23; E-06; C-16/17 | Pack-side same-transaction participation is legitimate. Incident projection rebuilding stays in Incident Bundles' supplied port. |
| Physical storage, worker registration, configuration and telemetry wiring | Owner capability interfaces in target; concrete adapters under `internal/app/referenceassembly` and composition under `internal/app/server` | Application/platform adapters, with owner-specific capability contracts | keep | E-02/03/04; F-03 | Opaque references already hide roots. No filesystem adapter extraction is currently missing. |
| ATTACK, Windows and Entra source production | `producer` | Reference Data offline producer | keep | O-01 source-profile/current amendments; E-02/08; C-18 | Source-specific CSV/XML/OOXML/Markdown parsing is not incident tabular ingest. Production imports of this package are prohibited by boundary policy. |
| Canonical fixtures and indicator test capability | `testsupport`, `testsupport/fixturehttp`, target test files | Reference Data owner-local support; service startup remains harness-owned | keep | E-02/06/08; C-19 | Do not move to generic test-util by filename. Keep runtime import exclusion and canonical machine-artifact inputs. |
| Administration controller/model/presentation | `apps/web/src/app/referencePackAdmin*`, `ReferencePackAdminPanel.tsx`, `useReferencePackAdmin.ts`; service adapter | Web application presentation/transport; Reference Data protocol owner | keep | E-07; C-20 | Already outside backend target; no UI-shell or vendor move proposed. |
| Generic rows, queries, saved views, view schemas, links, timeline, collaboration and grid vendor | Not implemented by the target; limited downstream effects through registry assignments/Jobs/portability | Existing source owners, workbook/projections, collaboration and grid adapter | defer | O-01 scope exclusions; E-05/06/07; C-21 | No direct target-owned implementation found to move. Do not infer ownership from framework examples or similar labels. Reopen only if a later slice traces an actual affected contract. |

## 4. Public Contract and Behavior Freeze Map

LC-C1 through LC-C4 define the current freeze. No new public API, profile retirement, compatibility reader, state conversion or recovery format is proposed. Earlier contract rows remain historical evidence and do not reinstate predecessor formats.

I8 preserves runtime contracts and implements one separately authorized offline behavior removal in S-38. NLSpec §1 (adopted/current 0.4.0, major 3) excludes compatibility readers and implicit conversion; RP-REQ-059/060 require release-bound Base verification. These clauses do not require the one-shot producer command. Removing it is the user's selected maintenance direction, not a claim that its offline existence violates the owner.

| Contract | Observed behavior / evidence | Owner-required behavior and I8 disposition | Characterization / slice / validation |
| --- | --- | --- | --- |
| C-18: offline producer | E-I8-03 records the retired fixed-release operation; E-I8-S38 removes its function/type, dispatch and witness admission | `cartulary.base_registry_reissue_input.v1` now fails unsupported-schema admission with no output or alias/conversion. Other producer inputs and outputs are preserved. | S-37/38; V-46; current CLI assertions plus new removed-input subprocess case |
| C-12/15: current Base bytes/readiness | Runtime reads embedded `builtins/release.v2.json` independently of the producer; E-I8-02/04 | Preserve canonical bytes, exact bindings, verification before readiness and retained provenance; NLSpec §1, RP-REQ-059/060 | S-37/40; V-48; current Base-consumer row and immutable-file comparison |
| C-13/19: rejection/test accounting | Live verifier rejects both checksum-labelled fixture variants with `bundle_hint_invalid`; unused helper options are removed; E-I8-04 and E-I8-S39 | Preserve both negative cases, error precedence and useful ZIP callers; no new reader or parallel verifier. Tests consume machine data, not prose. | S-37/39; V-47; existing verifier row |
| C-01–11, C-14/16/17/20–22 | No I8 change to HTTP, Jobs, consumers, storage/lifecycle, portability, UI or shared authorization | Preserve; no public HTTP/schema/data migration, rendered inspection or cross-owner restructuring | Not applicable to additional I8 checks: these surfaces are outside the bounded edit set |

Historical contract maps and earlier observations follow; their authority and execution claims retain their original scope.

Historical I6 contract delta (all rows retain their original evidence limits; I7 changes only the internal command return type and validation coverage in §21):

| Contract / surface | Observed behavior and evidence | Adopted owner / required behavior | Disposition and prerequisites | Trace |
| --- | --- | --- | --- | --- |
| C-18: source audit and producer interface | CLI audit loads proposal index; packaging binds current `native.v2.r2`; Audit emits proposal-era status/fields; E-I6-02/03 | RP-REQ-001/002 and §§1, 17; RP-REQ-149–152 preserve offline separation and exact immutable source provenance; producer interface details are repository-support choices | Authorized producer behavior change: current-only audit inputs and report v2; no change to runtime pack admission or source-profile bytes | F-26/27; S-28/30/31; V-36/38/40 |
| C-18: live Windows schema compilation and canonical packaging | Windows uses compiler in candidate file; assembly/report hashing and ledger remain live; E-I6-03/04 | RP-REQ-150/151 and deterministic-build requirements: preserve exact policy/profile binding, canonical artifacts, original source and signing distinctions | Preserve; characterize before moving; compare output for identical explicit inputs and builder identity, not distinct binaries | F-28; S-28/29; V-36/37/40 |
| C-19: expectations and embedded inputs | Candidate test consumes 13 schema-shape cases; producer input embed includes proposal files; E-I6-05 | Machine expectations remain downstream of adopted requirements; product verification must not consume prose or live websites | Migrate useful assertions to current contracts before deleting obsolete selectors/dependencies; retain immutable historical evidence without an active proposal reader | F-29; S-28/31; V-36/39 |

No HTTP, Job, saved-view, consumer result, persisted format, recovery or UI contract change is proposed. No new primary-owner contradiction was found. A stale proposal label in an immutable artifact is not permission to rewrite its bytes; preserve it as provenance.

Preserve owner-aligned behavior; correct the explicitly authorized gaps under §§5–7. Historical observations below describe the pre-remediation baseline. Owner-required behavior and observed implementation are distinguished explicitly; F-01 is not an approved exception. No tests ran during historical S-00. “Existing tests” in that baseline means inspected source; current execution evidence is in §10. Before a slice, add characterization only for the affected uncovered behavior; avoid duplicating broad existing tests.

All nine operations below are bound by `RegisterRoutes` through generated operation IDs. `P` abbreviates `/api/v1/reference-packs`; `V` abbreviates `P/{pack_key}/{pack_version}`. Shared HTTP errors, authentication/session/CSRF behavior and authorization order belong to Core. The collection uses common `meta.paging`; consumer lookup cursors are a separate contract. OpenAPI allows asynchronous activate/disable responses, while this backend currently returns inline 200.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| C-01: `GET P` / `listReferencePacks` | Reference Data adapter and repository; Core HTTP/paging | E-02 `http_adapter.go:75`, `administration_repository.go`; E-08; O-03 | `api_test.go` list-query/filter tests; `TestImportListReadReplayAndJobSummary_Integration` | Preserve exact ordering, empty array, nullable members, scope-bound continuation and actor binding. Add mixed-fault auth cases under S-01/02. | high | 200 `data.pack_versions` plus `meta.paging`; accepted query keys: limit, cursor_token, search, pack_version_state, active. |
| C-02: `GET V` / `getReferencePackVersion` | Reference Data adapter; Core envelope | E-02 `http_adapter.go:208`; O-03 | Import/list/read integration test; OpenAPI contract unit test | S-01 captures query-before-auth observation; S-02 tests required auth-before-query behavior. Preserve exact key/version and explicit nulls. | high | 200 `data=<version>`; singleton rejects paging. Current order conflicts with owner requirement; do not freeze that discrepancy as normative. |
| C-03: `POST P/import` / `importReferencePack` | Reference Data admission; Core upload/idempotency and Jobs | E-02 HTTP/coordinator/request sources; E-03/04; O-03 | Import replay, upload-envelope/admin, upload-limit and cancellation integration tests | Preserve byte-identical replay identity, staged-only default, no auto-activation, fresh-job notification versus replay, early-limit no durable state. | high | 202 common Job; metadata client_txn_id; exact admitted bytes enter normalized identity. Streaming limit is 409 content rejection without Job. |
| C-04: `POST P/refresh` / `refreshReferencePacks` | Reference Data lifecycle; Core admission/Jobs | `http_requests.go`, `coordinator_admission.go`, `coordinator_execution.go`; O-01 §20 | Refresh omitted-selector replay, frozen-cohort, mixed-refresh and request tests | Preserve omitted selector versus null/empty, deduped sorted explicit keys, frozen cohort, no-op and stale-revision behavior. | high | 202 common Job; newly imported versions never join an admitted cohort. |
| C-05: `POST V/activate` / `activateReferencePackVersion` | Reference Data semantic activation; Core transport | `administrative_actions.go`, `idempotency_receipts.go`; E-08 | Lifecycle and coordinator admission/publication tests; frontend transport 200/202 tests | Preserve reason normalization, exact route-scoped receipt, source usage compatibility, disablement clearing and atomic selected set. | high | Current 200 `data.pack_version`; permitted asynchronous projection does not authorize changing this implementation during refactor. |
| C-06: `POST V/disable` / `disableReferencePackVersion` | Reference Data lifecycle; Core transport | Same adapter/action/receipt sources; private state transitions | Lifecycle, pure state and frontend transport tests | Preserve independent disablement versus health, fallback/dependency consequences and replay before fresh eligibility. | high | Current 200 `data.pack_version`; no incident mutation. |
| C-07: `POST V/remove` / `removeReferencePackVersion` | Reference Data retention/lifecycle; Core transport | `administrative_actions.go`, collection/retention sources | Incident retention, collection, lifecycle/request tests | Preserve required nonempty bounded reason, built-in/active/pinned/pending-work rejections, tombstone/history and deferred byte deletion. | high | 200 `data.pack_version`; exact reimport and history are distinct from physical collection. |
| C-08: `POST V/reverify` / `reverifyReferencePackVersion` | Reference Data verification; Core Jobs | HTTP/coordinator/frozen verification sources | Lifecycle, retained-integrity, expiry and precedence tests | Preserve prior-success eligibility, current trust input, disabled state and failed-renewal versus invalidating-reverify outcomes. | high | 202 common Job; replay notifies no new execution. |
| C-09: `GET P/validation-summaries/{summary_id}` / `getReferencePackValidationSummary` | Reference Data immutable diagnostics; Core singleton envelope | `http_adapter.go:246`, `validation_summaries.go`; E-08 | Verification precedence, diagnostic and OpenAPI tests; frontend diagnostic-read tests | Preserve admin-first query/lookup, immutable `rpvs_` digest identity, exact bounded summary, absent/uncommitted rejection and safe errors. | high | 200 `data=<summary>`; no query members, private preparation or host paths. |
| C-10: accepted Jobs, progress, cancel/recovery and terminal effects | Reference Data owner operation; platform Jobs/jobapi and Extensions finalization | `job_ports.go`, `job_finalization.go`, `job_terminal_effects.go`, `coordinator_execution.go`; E-04 | Jobs/coordinator execution-scope/terminal-effect/worker-isolation tests; `TestJobsRequireDeploymentAdminAtPollAndCancelTime_Integration` | Preserve fresh-admission Notify, no replay duplication, cancellation before commit, one attempt budget, commit proof and indeterminate-fatal path. | high | No Reference Data WebSocket route/event found. UI observes common Jobs through HTTP. Do not invent pack events; any later common Jobs/collaboration change requires its own owner trace. |
| C-11: five consumer operations, errors, lookup cursors and indicator evaluation | Reference Data; pagination codec is shared substrate | `consumer.go`, `canonical_repository.go`, `internal/packformat/profile.go`, `indicator.go`; O-01 §21 | `consumer_test.go`, `consumer_contract_test.go`, profile-consumer, framework-lookup and indicator tests | Preserve shape→exact set→profile/type→semantics/cursor order; exactly one value/error arm; default 50/max 200; 900-second expiry including equality; bound normalized lookup/limit; no silent set fallback. | high | Canonical evaluation belongs here; raw source mutation, authorization and historical identity remain with Indicators. |
| C-12: canonical digests, current v2 set identity and immutable first-success provenance | Reference Data | `identity.go`, `public_types.go`, `provenance.go`, canonical publication; O-01 §§9,19 | Set/manifest/provenance/public DTO/persistence tests | Preserve bytes, member order, null versus empty, exact digest binding and unchanged pinned provenance across later verification. | high | Administrative current metadata must never replace snapshot provenance. No v1 compatibility migration or reset is planned. |
| C-13: verification order, offline TUF, safe diagnostics and resource limits | Reference Data private verifier | Format/archive/manifest/TUF programs, verification engine; O-01 §§7–18,26 | Canonical verifier, combined-fault/precedence, TUF, archive/content/manifest limit and fixture tests | Retain ordered winning check, duplicate-member rejection, equality/one-over bounds, bounded issue sorting/truncation, safe tokens, no runtime fetch and operational-fault distinction. | high | Generated check registry is downstream of owner semantics. Test phase maps do not define execution structure. |
| C-14: admission, revision guards, atomic publication and failure/cancellation | Reference Data application + shared finalization | Coordinator admission/guards/publication, packstate, integrity; O-01 RP-REQ-175/176 | Admission, concurrent publication, mixed refresh, consumer-integrity, terminal-effect, execution-scope tests | Preserve ordered trust/key/current-set/usage locks, frozen relevant revisions, no auto-rebase, verification outside long locks, all-or-nothing state/index/audit/Job effects. | high | Consumer reads can publish integrity loss; source transaction rollback precedes separate invalidation. This side effect must not disappear behind a “read-only” rename. |
| C-15: Base readiness, startup compatibility and historical restoration | Reference Data; application readiness/recovery coordinate | Base reconciliation, readiness, contractadmission and historical validation; E-03/06; O-01 | Base atomicity/reconciliation and required-Base-loss tests; contractadmission tests | Preserve three required registries, optional-only degradation, exact current-contract admission before mutation/readiness, fatal required-table failures and explicit restoration path. | high | Read-only admission is not schema migration, freshness verification or a reset procedure. |
| C-16: snapshot/reporting binding, pins and collection | Reference Data retention; Reporting owns capture/release, Recovery owns snapshot | `retention.go`, byte retention/object collection; E-05/06 | Canonical persistence, incident retention/export and collection-storage tests | Preserve exact pinned set/provenance, no current substitution, selected-object lease before repeatable-read backup snapshot, no deletion on uncertain commit. | high | No direct saved-view or view-schema changes. Owner pin semantics can block removal. |
| C-17: incident catalogs/embedded content, destination verification and recovery contribution | Reference Data pack side; Incident Bundles/Recovery coordinators | Incident/portable sources, recoverycontribution; E-06; O-01 §23 | Portability unit/retention/export/admission tests; object-storage backup integration assertions | Preserve refs v2 under outer bundle 7, exact collisions, inactive new imports, destination trust, unresolved optional references, required-member failure, same final transaction and restored linkage. | high | Parent incident execution owns projection rebuild and incident authorization. Pack contribution does not become a generic portability engine. |
| C-18: offline producer source identity, licensing and publication | Reference Data producer; offline tool entrypoint | Producer files; `tools/reference-packs/main.go` import trace; E-08 runtime exclusion; O-01 current source profiles | ATTACK source/candidate/ledger, Windows acquisition/projection/publication, Entra source integrity/boundary tests | Preserve pinned acquisition, immutable source profiles, deterministic canonical artifacts, fixture signing separation and no runtime producer dependency. | high | Explicitly pinned upstream Markdown parsing is confined to the offline Entra producer; product tests use canonical artifacts/expectations, not repository specification prose/live sites. |
| C-19: harness accounting, canonical fixture inputs and test-util isolation | Reference Data test owner + harness/test-util | E-02/08/09 | 93 test files; 84 catalog rows; testsupport corpus expectations | Preserve executed selectors/parent subtests, owner verification IDs, service profiles, production import exclusions and fixture-only fake registry behavior. | medium | Counts prove neither execution nor completeness. No new harness runtime phase/module is proposed. |
| C-20: admin UI transport, controller lifetime, focus/selection and selectors | Web application; protocol/UI contract owners | E-07, generated imports in panel/service | Six frontend and three browser owner rows; three measurement rows (§8) | Preserve uncertain-request replay bytes/transaction, independent observation recovery, latest query generation, access-loss cleanup, hidden-panel lifetime and existing generated selectors. | high | No grid vendor import in the inspected admin seam. Timing publication would additionally require Core 05; none is claimed here. |
| C-21: entities/indicator mutations, revision history, workbook/test-util rows, saved views and projections | Source owners, Revisions, Workbook/Projections and test-util | E-05/06; registry transaction and source-port composition; O-04/05 | Consumer-integrity and optional-degradation tests; isolated test registry support; downstream rows require selection if touched | Before changing registry/portability ports, characterize caller rollback, revision/projection consequences and row/query/save/conflict/inspector behavior only along the actual changed caller path. | high if those ports change | No direct target implementation of timeline, links, saved views, view schemas or grid adapter was found. They are not automatic refactor slices. TODO: exact downstream row selection only when a future slice crosses those ports. |
| C-22: claimed-family authorization ordering and denial side effects | Core 01/04; Reference Data transport and application routing | `http_adapter.go:164,208`; `httpapi.go:185–225`; `operation_catalog.go:126–134`; O-03 REQ-01-480 | Upload-admin test checks a valid import; singleton pagination test uses authorized caller; shared unclaimed-family test | S-01/02: unauthenticated, ordinary authenticated and admin callers × claimed/unclaimed profile × malformed query/path, unsupported method, unknown descendant and normal nine operations. Assert no Job, receipt or pack mutation on denial. | high | Static mismatch: singleton query validation precedes auth; generated method routing/unknown paths can terminate before handler auth. No runtime reproduction run. This is implementation-versus-owner drift, not an owner contradiction. |

## 5. Coupling and Boundary Findings

LC-F1 through LC-F6 are the current bounded cleanup findings. Earlier closed findings below remain closed on their recorded evidence; LC does not infer dead code from a filename, old version suffix or prior completion claim.

Current I8 findings are F-33/34. Their removal/simplification is implemented and qualified by E-I8-S38/39/40; both findings are closed. The table retains each finding's baseline diagnosis and required acceptance:

| Finding | Evidence / contracts | Risk / classification | Owner / required action | Slice / validation / completion evidence |
| --- | --- | --- | --- | --- |
| F-33: obsolete Base cutover remains executable | E-I8-03; fixed `base-2026-11-06.1`, sequence 2 and release ID; C-18, preserved C-12/15 | medium / must_fix for S-38 retirement acceptance | Offline Reference Data producer; remove function/input type, CLI branch and witness admission together. Defer a general builder until a concrete release needs one. | S-37/38/40; V-46/48; E-I8-S38/40; DONE |
| F-34: predecessor fixture exposes unused configuration branches | E-I8-04; only `Signed` varies at the sole caller; C-13/19 | low / should_fix | Owner-local tests; keep a signed/unsigned helper and actual rejection assertions; remove unused options and branches while retaining other ZIP callers. | S-37/39/40; V-47/48; E-I8-S39/40; DONE |

I8 explicitly supersedes I2's §16 and I6's blanket decision to retain the producer reissue path, only for `ReissueBase`. Historical rationale/results remain intact. Fixture signing helpers, consumer-result validation, live current-content readers, negative compatibility fixtures and effective-set fallback remain intentional/no_action; a test-only caller or a `v1` schema suffix alone is not evidence of dead code.

I6 findings at baseline, scoped to C-18/C-19 and S-28–32. All four findings are resolved by E-I6-S32; the baseline evidence and remediation rationale remain below:

| Finding | Evidence | Risk / classification | Required action and durable benefit | Slice / validation |
| --- | --- | --- | --- | --- |
| F-26: active obsolete ATT&CK proposal execution | E-I6-02/03; CLI `--candidate`, proposal index, loader and validator still execute beside adopted packaging | high / must_fix for the current-only producer acceptance criterion | Remove the alternate proposal route; share current profile/policy admission with packaging so there is one supported producer contract | S-28/30; V-36/38/40; RESOLVED E-I6-S32 |
| F-27: source audit claims obsolete adoption blockage | E-I6-02; `pack_status`, `native_candidate`, `proposed_lookup_aliases` | medium / must_fix for truthful current audit output | Report v2 describes source-audit outcome and exact profile provenance; it must not imply runtime conformance, deployment trust or distribution approval | S-30/31; V-38/40; RESOLVED E-I6-S32 |
| F-28: live Windows schema utility lives in candidate implementation | E-I6-04; two Windows schema loops call `CompileCandidate` | medium / should_fix; required before deleting that file | Keep one private producer-local compiler under a current neutral name, preserving its behavior and Windows callers; no compatibility alias | S-29; V-37; RESOLVED E-I6-S32 |
| F-29: candidate tests/embedded input closure retain obsolete workflow | E-I6-05; 13 cases, source fixtures, broad JSON embed and catalog selector | medium / must_fix for removal without lost assertions | Migrate useful tests, remove proposal-only executable dependencies and update authored routing; retain immutable artifacts needed by historical provenance | S-28/31; V-36/39; RESOLVED E-I6-S32 |

These were active legacy responsibilities at the I6 baseline, not proven unreachable functions. The bounded lexical scan did not establish further dead runtime functions. I6 retained fixture signing helpers, consumer-result validation, effective-set fallback, retained-state rejection and producer reissue paths; I8 supersedes only the last decision for the named Base cutover command. F-06/F-07's owner separation remains valid; their earlier blanket no-action wording is narrowed by F-26–29 and F-33/34. F-12 remains deferred: file count alone still does not justify package subdivision.

The next-iteration findings F-18–20, their remediation/compatibility decisions and final dispositions are recorded in §16. Earlier findings below retain their original scope and evidence.

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| F-01: claimed-family authorization does not uniformly precede route-specific handling | C-22; `handleRead` calls `ValidateSingletonReadQuery` before authentication; route registry binds method patterns without auth; outer HTTP wrapper only handles unclaimed families | high: changed error/denial ordering and incomplete family protection | must_fix | Reference Data transport composed through application HTTP; Core owns required ordering | S-01 establishes runtime matrix; S-02 corrects this under the current implementation authorization. Preserve unclaimed-profile precedence. RB-001/002. |
| F-02: public facade and pure-state seams already exist | AdministrativeApplication, Consumer, Retention, RegistryAssignments, IncidentReferences, private packstate; construction test | high if duplicated/replaced | intentional/no_action | Reference Data | Keep these boundaries; C-11/14/16/17 and V-01/02/04. Do not introduce another facade solely to match a framework example. |
| F-03: platform/SQL dependencies are mostly at owner adapters/orchestration, with concrete filesystem roots in assembly | E-02/03; storage_port; coordinator Jobs ports; pure packstate imports; canonicaljson used in formats | high if locks, bytes or finalization move | intentional/no_action | Reference Data adapters; application/platform substrate | Preserve opaque capabilities, same-transaction ports and commit proof. No SQL extraction slice warranted. C-12/14/17; V-04. |
| F-04: consumer integrity detection can mutate authoritative pack health/selection | `canonical_repository.requireAvailable`, `integrity.go`, `registry_assignment.go`; consumer-integrity tests | high: hidden rollback/fallback consequence | should_fix | Reference Data consumer/integrity API | S-07 makes this obligation explicit on the exported consumer APIs and preserves existing tests; C-14/21. No behavior change or new pure-read promise is proposed. |
| F-05: pack-local portability and recovery are intentional; generic workflow stays elsewhere | IncidentReferences preparation/apply; Reporting supplies bindings; Incident Bundles applies in its final transaction; Recovery acquires retention lease | high | intentional/no_action | Reference Data contribution; Reporting/Incident Bundles/Recovery orchestrators | C-16/17/21; keep current owner split; reselect collaborator tests only if a later slice changes a port. |
| F-06: producer parsing/network/filesystem code is offline and source-profile specific | E-02 producer reads; E-08 `production-no-reference-pack-producer` and raw-filesystem exception | high if production imports appear | intentional/no_action | Reference Data offline producer | C-18; V-04/05/06. Do not move Windows CSV/OOXML into incident ingest or allow upstream Markdown into runtime/test authority. |
| F-07: test helpers and test-only registry are owner-local, not production orchestration | testsupport files, appsupport workbook composition; E-08 support inventory and `production-no-reference-pack-fixtures` | medium | intentional/no_action | Reference Data tests; test-util/harness assembly | C-19; V-04. Preserve test-only exception; do not infer production algorithms from the fixture fake. |
| F-08: no duplicated target-owned workbook row/view-schema or grid implementation was found | Target imports/declarations, source-owner caller trace and E-07 UI imports | medium if hypothetical moves are invented | intentional/no_action | Existing Workbook/Views/Projections/grid and source owners | C-20/21; no timeline, saved-view, links, collaboration or grid relocation slice. Reopen on concrete evidence only. |
| F-09: canonical/generated surfaces are broad despite no target-generated files | E-08 generated policy, OpenAPI operations; format schemas and public DTO conversions | high: downstream drift/hand-edit risk | intentional/no_action | Contract authors/generators and Reference Data projections | C-01–20; V-05/06. S-08 edits authored inputs and regenerates through Make for the targeted transport amendment. |
| F-10: existing test presence does not establish family-wide denial precedence | Target auth/pagination tests and shared dispatch tests inspected; no combined matrix found in those sources | high | must_fix | Reference Data test owner; shared HTTP owner if dispatcher changes | S-01/02; C-22; V-02/07. Add focused evidence, not a copy of existing lifecycle tests. |
| F-11: framework examples and workflow catalog exceed this target's live seams | Framework versus E-02/03/07: facade/storage/worker seams already exist; no generic grid/workbook catch-all | low | intentional/no_action | This tracker | Record mismatch rather than manufacture extraction work. Use task-local WF-00–08 in §6. |
| F-12: broad private package subdivision lacks an evidenced payoff | Root has 110 files; formats/state/producers already separated; no measured bottleneck or demonstrated ownership violation from file count | medium: unnecessary churn across canonical/transaction boundaries | defer | Reference Data maintainers | No production extraction slice. Reopen only with concrete coupling/change-locality evidence and the relevant C/V map. RB-002 does not authorize a general reorganization. |
| F-13: UI administration already separates service/model/controller/hook/panel | E-07 and frontend rows | high if access/replay/lifetime state changes | intentional/no_action | Web application | C-20; V-03. Keep outside backend target. No rendered UI or selector refactor authorized. |

F-01 and F-10 are the same correction chain at different evidence layers, not two independent runtime changes. F-04 is an exported API documentation obligation completed by S-07 without a signature or behavior change. No finding establishes an unresolved contradiction between applicable owners.

### Targeted audit remediation decisions — E-10

| Finding | Areas and remediation | Rationale and long-term benefit | Compatibility impact and unresolved risk | Validation / slices |
| --- | --- | --- | --- | --- |
| F-01 / F-10 | Implementation/tests; family authorization before mux dispatch; current session principal propagated once; operation-declared CSRF after admin gate. Core 01 clarifies ordering. | One enforcement point protects future descendants and prevents handler omissions. | Unauthorized malformed requests change to authentication/authorization errors. Leaving it unresolved permits inconsistent disclosure and owner drift. No data migration. | C-22; complete claimed/unclaimed and actor matrix, no operation effects; S-01/S-02, V-02/V-07. |
| F-14 | Documentation: repair Reference Data primary-owner navigation, Base-and-extension applicability and optional-enrichment wording. | Correct authority and availability distinctions guide future changes. | No runtime change; stale guidance could weaken required registry availability. | Human owner review and Markdown lint; S-04, V-00. |
| F-15 | Specification/implementation/contracts/tests: direct generated operation binding; exact once-decoded opaque versions; retire missing-version heuristic and `pack_version_required`; reject implicit HEAD/OPTIONS. | One routing definition eliminates ambiguous interpretation and accidental methods. | Clients relying on retired errors or incidental HEAD must update; valid action-named versions become addressable. No aliases/migration. Unresolved ambiguity makes valid versions inaccessible. | C-02–09/C-22; exact-version/method matrix; S-04/S-08/S-01/S-05. |
| F-16 | Specification/implementation/contracts/tests: empty queries on all eight non-list operations; parse raw query/UTF-8 before body/replay; malformed_query before pagination_not_supported before unknown_query_member. Family errors use field=query and echo no input. | Closed inputs avoid accidental compatibility promises and silent partial parsing. | Previously ignored parameters fail, including replay requests; clients remove them. Unresolved actions violate paging requirements and summaries accept prohibited input. | All non-list routes, aliases/repetition/encoding/mixed faults and no side effects; S-04/S-08/S-01/S-06. |
| F-17 | Specification/contracts/generated/tests: include summary operation in compact inventory; project 401/403/inactive 404 and revised query/method/version behavior. | Complete reviewable transport and generated-client contract. | Keep unreleased HTTP API 2.0.0 against immutable 1.0.0; exact compatibility accounting. Pack v2/profile major 3 unchanged. Missing projections hide security outcomes. | Nine-operation parity, error registry, generated/schema/compatibility checks; S-04/S-08. |
| F-04 | Exported API documentation: disclose owner integrity mutations, operational failure distinction, rollback-before-independent invalidation and transaction lifetime. | Callers can reason about locks and source transaction consequences without a misleading pure-read facade. | No signatures or behavior change. Undocumented effects risk deadlocks and committing after content loss. | Existing rollback/fallback/readiness/operational-failure/replacement tests; S-07, V-02. |

F-02/03/05–09/11/13 retain their evidence-backed owner boundaries and safeguards. F-12 remains deferred without an evidenced extraction benefit; it is not an unfinished required remediation. No consumer, producer, retention or UI architecture rewrite is justified.

## 6. Refactor Workstreams

LC-W1 through LC-W5 are the next workstreams. The LC phase table is their authoritative dependency graph; the historical workflow tables below do not authorize current execution.

The following is the sole authoritative I8 prerequisite graph. Implementation is explicitly authorized; record each exit checkpoint before its successor. No sub-agent or parallel execution is prescribed.

| Workflow ID / name | Class | Authoritative prerequisites | Goal / likely files | Validation | Handoff checkpoint / binary exit |
| --- | --- | --- | --- | --- | --- |
| WF-I8-00: characterization / S-37 | root | Current user implementation authorization and E-I8-S37 revalidation | Recheck callers/assertions; fresh CLI, verifier and Base-consumer baseline; immutable fingerprints | V-46–48 baseline portions | E-I8-S37: three rows pass on the implementation baseline; assertion/preservation map complete |
| WF-I8-01: retirement / S-38 | chain | WF-I8-00 | `producer/base.go`, CLI branch/test and witness allowlist | V-46 | E-I8-S38: command rejected before release input/output work; executable path absent; checkpoint recorded |
| WF-I8-02: fixture simplification / S-39 | chain | WF-I8-01 | Retired fixture helper and verifier caller | V-47 | E-I8-S39: unused options absent, both live rejection cases and other ZIP callers preserved |
| WF-I8-03: qualification / S-40 | chain | WF-I8-02 | Finalization, source review, scoped checks and tracker | V-48; implementation Markdown/whitespace checks | E-I8-S40: compatible final evidence, immutable preservation and complete handoff |

The graph has no conditional peers, cycles or dangling prerequisites. T-034's document update is separate and does not complete S-37. The I6 table below and §21's I7 graph record historical implementation.

| Workflow ID / name | Class | Authoritative prerequisites | Goal / likely files | Validation | Handoff checkpoint / binary exit |
| --- | --- | --- | --- | --- | --- |
| WF-I6-00: characterization / S-28 | root | Revalidated I6 baseline and later implementation authorization | Current producer/CLI assertions and all 13 candidate-case destinations | V-36 | Complete assertion map and passing focused baseline; no lost behavior silently accepted |
| WF-I6-01: schema utility / S-29 | chain | WF-I6-00 | `producer/candidate.go`, `windows_contract.go`, private producer schema helper | V-37 | Live Windows compiler callers migrated with unchanged semantics |
| WF-I6-02: current audit / S-30 | chain | WF-I6-01 | Producer audit/assembly input admission and CLI | V-38 | Proposal CLI/loader/validator removed; current source audit works; checkpoint explicitly records report cleanup remaining |
| WF-I6-03: reports and inputs / S-31 | chain | WF-I6-02 | Report DTO, assertions, input embed and authored test routing | V-39/40 | Report v2 and complete test migration; no active proposal input/alias remains |
| WF-I6-04: qualification / S-32 | chain | WF-I6-03 | Final source, generated routing and tracker handoff | V-40/41 | All scoped obligations pass on frozen source, findings dispositioned and preservation evidence recorded |

Every dependency resolves in this table; there are no conditional peers or cycles. Documentation publication is separate from S-28–32 and does not satisfy their implementation gates.

Historical transport workstreams:

| Workstream | Phase | Predecessor | Risk | Binary exit criterion |
| --- | --- | --- | --- | --- |
| S-04 specification/tracker | A | historical S-00; fresh E-10 | Authority duplication or adoption mistaken for validation | Owner amendment, cross-references and tracker agree; Markdown lint passes. |
| S-08 contract projections | A | S-04 | Generated drift or unreviewed security changes | Authored contracts regenerated; schema/generated/compatibility checks and contract assertions pass. |
| S-01 regression evidence | B | S-08 | Freezing a defect as expected behavior | Reproduction and intended assertions exist; expected failures attributed to S-02/05/06; selectors accounted. |
| S-02 family authorization | B | S-01 | Missing/misordered gate or auth regression | Full family matrix and registration/composition checks pass; no denial operation effects. |
| S-05 operation routing | B | S-02 | Parameter decoding or method ambiguity | Direct operations and exact-version/method matrix pass; old parser/reason emission removed. |
| S-06 query admission | B | S-05 | Partial parsing, input echo or late validation | Eight non-list operations reject forbidden input before operation work; list behavior unchanged. |
| S-07 consumer contract | B | S-06 | Misleading purity or weakened rollback | Exported contract documents existing semantics; integrity regression and Markdown lint pass. |
| S-03 validation/handoff | C | S-07 | Stale/partial evidence presented as completion | Targeted portfolio passes; all findings dispositioned; final handoff and preservation review recorded. CI is not a completion requirement. |

## 7. Proposed Refactor Slice Plan

LC-P0 through LC-P5 are the next slices. LC-P0 is ready for inspection in a later authorized task; deletion and movement remain blocked on its caller/assertion characterization and the stated predecessor exits.

Current I8 slices use §6's WF-I8 graph under the user's explicit implementation authorization. Readiness and work status are independent; current execution checkpoints are in §22.

| Slice / kind | Findings / contracts | Intended change | Entry evidence / readiness now | Tests / validation | Rollback / migration | Binary completion / eventual evidence |
| --- | --- | --- | --- | --- | --- | --- |
| S-37 / characterization | F-33/34; C-12/13/15/18/19 | Confirm caller/assertion map, capture three baseline rows and immutable fingerprints | DONE; E-I8-S37 confirms sources, assertions, routing and passing baseline | V-46–48 pre-move portions | Inspection/test evidence only; no immutable output regeneration | Passing compatible baseline and complete destinations; E-I8-S37 |
| S-38 / authorized offline behavior removal | F-33; C-18, preserve C-12/15 | Remove `ReissueBase`, `BaseReissueInput`, CLI branch and witness accepted-input entry; add removed-input subprocess assertion | DONE; E-I8-S37 resolved RB-005 and E-I8-S38 records retirement | V-46; existing CLI row, no new selector | Revert function, dispatch, witness and tests together; no data conversion or compatibility alias | Retired schema fails unsupported-schema admission, with no success report or requested output; no executable reissue path; E-I8-S38 |
| S-39 / structural test simplification | F-34; C-13/19 | Replace options struct with signed/unsigned helper; remove unused branches, preserve ZIP helper callers | DONE; E-I8-S38 preceded edits; E-I8-S39 records preserved assertions | V-47; existing live verifier rejection row | Revert helper and caller together; no fixture artifact rewrite | Both predecessor variants still reject at `bundle_hint_invalid`; unused options absent; E-I8-S39 |
| S-40 / qualification | F-33/34; C-12/13/15/18/19 | Finalize, review, freeze source and run scoped portfolio; publish checkpoint | DONE; E-I8-S39 preceded finalization; E-I8-S40 records final qualification and preservation | V-48 and documentation checks | Repair invalidates affected qualification; no deployment/reset | Required checks and preservation pass on compatible final source; E-I8-S40 |

Completed I6 slices follow; their authorization and DONE statuses are historical. I7's completed supplemental slices remain in §21.

| Slice / kind | Findings / contracts | Intended change | Entry evidence / readiness | Tests / validation | Rollback / migration | Binary completion / future evidence |
| --- | --- | --- | --- | --- | --- | --- |
| S-28 / characterization | F-26–29; C-18/19 | Establish current-input, output and ledger evidence; map every candidate case before retirement | DONE; E-I6-S28 and V-36 characterization passed | V-36; §20 assertion map | Test-only rollback; no live-source fetching or schema relaxation | All 13 destinations implemented or explicitly justified as obsolete; focused baseline passes; retain E-I6-S28 |
| S-29 / structural preservation | F-28; C-18 | Move/rename compiler to a private producer-local helper, migrating Windows and remaining current callers | DONE; E-I6-S29 compiler and Windows assertions passed | V-37 | Revert helper and callers together; no alias, library change or wire migration | No live caller depends on candidate compiler placement/name; Windows row passes; E-I6-S29 |
| S-30 / producer cutover | F-26/27; C-18 | Cut CLI audit to adopted inputs; share admission with assembly; remove candidate flag, loader, validator and execution-only types | DONE; E-I6-S30 current admission and CLI passed | V-38; current input/strict rejection and CLI execution | Revert producer/CLI/tests together; S-30/31 form one releasable cutover; no input conversion | Current audit path passes; no proposal execution; remaining report-only types removed in S-31; E-I6-S30 |
| S-31 / report cutover and cleanup | F-27/29; C-18/19 | Emit report v2, remove obsolete report DTO/types, retarget assertions and narrow embedded inputs; regenerate changed routing through Make | DONE; E-I6-S31 report/input cutover passed | V-39/40 | Revert report consumers and producer changes coherently; preserve old retained reports as historical bytes | All report destinations match canonical bytes; no active proposal dependency/selector; immutable evidence unchanged; E-I6-S31 |
| S-32 / qualification | F-26–29; C-18/19 | Finalize, review, freeze, run scoped portfolio and publish final handoff | DONE; E-I6-S32 final qualification and handoff passed | V-40/41 | No deployment or data reset; repair invalidates affected qualification | Compatible source/commands/native exits/artifacts close each obligation; E-I6-S32 |

The following table is the historical transport implementation plan, retained for traceability:

| Slice | Implementation / interface decision | Validation | Rollback / migration |
| --- | --- | --- | --- |
| S-04 | Core 01 owns auth-before-routing, declared-only methods, exact opaque versions, complete route inventory and empty non-list queries. Add dated adoption in Core 00; Reference Pack/Extensions import rather than duplicate. Repair domain owner navigation. | V-00; human owner review | Revert coherent owner amendment only with corresponding downstream projections; no data conversion. |
| S-08 | Edit authored Reference Data OpenAPI/error registry; add common 401/403/inactive 404 responses; remove retired reason, add malformed_query/unknown_query_member; update contract assertions and exact 2.0.0 compatibility accounting. Make regenerates protected outputs. | V-05/06/09 and OpenAPI contract row | Keep immutable released 1.0.0 snapshot unchanged. |
| S-01 | Add complete server-composed authorization/method/version/query regressions plus focused no-body/no-call checks. Use canonical fixtures and machine expectations, never prose. | V-01/02/07; known red tests are reproduction only | Tests remain requirement assertions; no permanently failing completed implementation. |
| S-02 | Owner-bound family middleware registered with existing HTTP registry; validate binding against admitted projection. Generic HTTP invokes policy without importing Reference Data. Authenticate/admin once, private principal context, operation-declared CSRF against principal. Preserve unrelated auth ordering. | V-02/07 and current-role/CSRF cases | Revert gate, plumbing and tests together; no durable-state edits. |
| S-05 | Direct generated operation handlers, router PathValue once, explicit operation decoder selection; delete manual parser/suffix dispatch and method aliases. | V-01/02/07; action-named and encoded versions | No compatibility parser retained. |
| S-06 | Local empty-query helper parses RawQuery, validates UTF-8, applies malformed > pagination > unknown precedence. Run after security and method gates, before body/resource/replay. No global singleton behavior change. | V-01/02; all eight non-list routes | Clients remove ignored parameters; no receipt rewrite. |
| S-07 | Document Consumer, constructor and transaction-bound API integrity effects and lifetime. Keep existing service and immediate rollback/invalidation design. | V-02 integrity cases; V-00 | No behavior/signature changes. |
| S-03 | Finalize, review generated diff, freeze implementation, run targeted Reference Data/shared HTTP/client and source checks; complete tracker and handoff. CI removed by user direction. | V-00–09 selected portfolio | Preserve exact run artifacts and unrelated user work. |

I8 readiness is determined by its own predecessor exits, not historical DONE rows or earlier implementation permissions. A genuine primary-owner contradiction, if discovered, blocks only the affected decision; none is established for I8.

## 8. Validation Plan

LC-V0 is this invocation's documentation check. LC-V1 through LC-V5 are future implementation obligations, including complete Reference Data owner acceptance and affected cross-owner coverage. Historical passes below do not satisfy those obligations on changed source.

Current implementation obligations V-46–48 and their exact portfolio/results are in §22. Baseline, post-slice checks, finalization, final three-row portfolio, boundary/generated/schema checks and reproducible producer build pass. S-40 documentation and final preservation checks also pass. V-49 records historical planning-document validation. Scoped qualification does not claim whole-owner or production-release acceptance. All commands use repository root `/home/jochi/code/cartulary` and public Make entrypoints.

The following I6 plan and document-session discovery are historical; completed execution is recorded in §20 and I7 supplemental evidence in §21. I6 planning discovery found 92 owner rows before the CLI row was added. The authored inventory authority remains `tools/test_families/module.reference_data.json`. Counts are routing facts, not passes.

| ID / layer | Findings / slices | Stage / exact public command | Environment / expected evidence | Failure gate |
| --- | --- | --- | --- | --- |
| V-36 / characterization | F-26–29; S-28 | Pre-move: `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f` | Fresh run ID, cache off; machine fixtures/current canonical schemas; baseline assertions and 13-case map | No deletion before current-path coverage is demonstrated |
| V-37 / compiler preservation | F-28; S-29 | Post-slice: `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930` | No services/live website; Windows schema and producer assertions | Block compiler removal or semantic drift |
| V-38 / CLI/current audit | F-26/27; S-30 | Post-slice: ATT&CK producer row from V-36, extended with real current entry-path cases; `make reference-pack-producer-build` | Retained source or explicitly synthetic unit inputs; production rejects unbound sources; later tests must cover actual CLI dispatch/report destination rather than only isolated helpers | Wrong profile, proposal input, altered digest, malformed source or command dispatch must not yield success |
| V-39 / test/input accounting | F-29; S-31 | Post-slice: `make generate` for authored selector/input changes, then `make source-check CHECK=generated` and `make source-check CHECK=schema` | Authored test-family inputs first; generated outputs only through Make; source/caller review confirms no stale proposal execution | Missing selector or active proposal dependency blocks completion |
| V-40 / final behavior | F-26–29; S-31/32 | Final regression: exact five-row selection recorded in §20 after `make agent-finalize`, diff inspection and source freeze | Cache off; authentic lifecycle row needs PostgreSQL/object store; report/output parity cases stay in the producer row | Historical fixture import alone cannot prove new audit/assembly behavior; new pipeline assertions are required |
| V-41 / boundaries and docs | F-26–29; S-32 | Final regression: `make lint CHECK=backend-boundary`, generated/schema checks above, `make lint-markdown`, `git diff --check` | Current source, exact retained roots/native exits; no added runtime producer dependency | Do not close findings on stale or failed evidence |
| V-I6-DOC / document update | Tracker only | Planning-session check: `make lint-markdown` with the environment recorded in §20; `git diff --check`; baseline/index comparison | Installed-only prerequisites; external temporary results/static cache; fix disabled | Documentation task incomplete if lint or tracker-only preservation fails |

The historical I6 plan called for current-input binding, current native schema rejection, deterministic audit/report output and identical-input assembly-output checks. Its completed implementation rejects `--candidate` without a compatibility alias and retains acquisition/ledger and runtime-exclusion assertions. Review found that native rejection and failed-report delivery needed stronger isolated witnesses; I7 adds those under the existing rows. Product tests use canonical artifacts and machine expectations, never specification prose or live websites. No I7 routing expansion is planned.

| Historical I6 document-session invocation / side-effect evidence | Admission and result at that time | Evidence limits |
| --- | --- | --- |
| `make task-guide OWNER=module.reference_data DETAIL=rows JSON=1`; generated Make recipe and read-only diagnostic wrapper inspected | Plan-mode discovery passed, native exit 0; E-I6-05. Retained current catalog facts supply exact IDs above. | Selection only; no product pass. |
| `make lint-markdown`; Make prerequisites, installed-readiness probe, step wrapper, static cache and runner inspected | Admitted with `CARTULARY_PREPARATION_POLICY=installed_only`, external results/static cache, network-disabled Corepack; result recorded in §20 | `fix:false`; readiness checks must fail instead of installing. |
| Format, generation, agent-finalize, producer build/source/package and product suites | Not run in this document session | Future obligations only; no implementation, acquisition, release evidence, service mutation or dependency change is authorized here. |

Historical validation plans and accounting follow. Their references to 84 rows, unrun checks or later passing runs belong to those earlier sessions and are superseded for current selection by this section.

The following original selection is historical discovery, not passing evidence. Current execution is authorized; earlier write-boundary skip reasons no longer apply. S-03 uses the targeted portfolio below after `make agent-finalize`. The user explicitly removed CI from the requirements and directed cancellation of its in-flight invocation; neither CI nor `make check` is required. V-09 is `make openapi-compatibility-check`. V-07 owners are `platform.httpruntime`, `platform.openapi`, `app.server` and the existing shared-auth row owner.

The table below preserves historical command discovery. Its historical “not run” and tracker-only restrictions are superseded by the exact execution records in §10; they are not current skip claims. Row presence and command discovery do not constitute a pass.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| documentation | `make lint-markdown` | Repository Markdown, including this tracker | no; required for this documentation task | V-00: **skipped** under the explicit tracker-only write boundary. E-09 help declares retained tool-run artifacts; `markdownlint.sh` uses a cache and runner writes an output stamp; Make can prepare frontend dependencies. Do not bypass Make or run then restore outputs. |
| unit | `make test-slice OWNER=module.reference_data` | Owner selection; catalog includes Go, frontend, browser and measurement rows, not a unit-only promise | yes, narrowed to affected rows | V-01: prefer exact `ROWS` selection below for the first baseline. Owner-wide command is the discovered public recommendation, not a claim that all rows are fast/unit. |
| integration | `make test-slice OWNER=module.reference_data SCOPE=services` | 38 service-backed catalog rows, including browser/measurement classes | yes, affected rows | V-02: targeted auth/import/lifecycle rows before broad owner execution. Service permissions/resources are a later-task obligation. |
| e2e/browser | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.browser.administration_real_lifecycle` | Real admin lifecycle browser scenario | no for planning; conditional baseline for changed UI/API behavior | V-03a: browser owner is Reference Data with `web.application` and `harness.browser` collaborators. Do not use guessed generic frontend owner. |
| frontend | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.frontend.administration_transport` | Status-specific envelopes, upload/selector replay and diagnostic transport | conditional on adapter/response changes | V-03b: other real rows are `administration_query_ownership`, `administration_recovery`, panel rows and `progress_paint` below. |
| generated drift | `make source-check CHECK=generated` | Repository generated-artifact drift | yes for changes reaching projections; otherwise after relevant implementation | V-05: OpenAPI, Go/TypeScript and topology outputs are protected. No direct generated-file edits. Not run. |
| schema | `make source-check CHECK=schema` | Machine-schema validation | conditional on contracts/catalog/schema implications | V-06: no contract/schema modification needed by this documentation task or presumed by S-02. Not run. |
| import-boundary/static | `make lint CHECK=backend-boundary` | Backend facade/import/root/fixture/producer boundaries | yes before boundary-affecting edits | V-04: authored boundary policy inspected; check itself not run. Preserve source-owned SQL and runtime exclusions. |
| shared HTTP / downstream owners | Exact `make test-slice` selections for `platform.openapi`, `platform.httpruntime`, `app.server`, `module.auth` in E-22/23 | Family registry, runtime parity, unclaimed dispatch, readiness/publication, existing CSRF/session behavior | required for shared plumbing changes | V-07: eight targeted rows; exact ROWS inputs retained in final run manifests. |
| final implementation portfolio | Targeted test/source checks in E-22/23 after `make agent-finalize` | Only changed owner contracts and affected shared/client behavior | required | V-08: replaces full CI/check at the user's explicit direction. No full-suite acceptance claim. |

Verified catalog accounting from `make task-guide OWNER=module.reference_data DETAIL=coverage`: 84 rows; 40 unit-family, 32 integration-family, six frontend, three browser and three measurement; 38 service-backed, 46 runtime-profile `none`; runners: 72 Go, six Vitest, six Playwright. “Family” and “evidence class” are not identical: the six frontend rows have unit evidence class. Aggregate placements are 46 fast, 31 standard, three full, three release and one explicit-only. None of these labels defines runtime module ownership.

| Selection purpose | Exact existing row or row group | Evidence/test posture |
| --- | --- | --- |
| Construction | `module.reference_data.unit.application_construction` | `TestReferencePackAdministrationConstruction_Unit` in application assembly; no execution performed. |
| Closed requests | `module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda` | Request and incident catalog unit selectors; not a family-authorization matrix. |
| OpenAPI/error projection | `module.reference_data.unit.generated_open_api_and_error_registry_artifacts_0843af138f` | Exact generated surface test; projections do not override owners. |
| Existing admin denial | `module.reference_data.integration.reference_pack_upload_envelope_failures_create_n_710dd1d70f` | Valid import by non-admin and bad upload by admin; expand or add explicit selectors for C-22. |
| Import/list/replay/operator | `module.reference_data.integration.reference_pack_import_uses_the_shared_upload_env_95574f69b8` | Two top-level integration selectors. |
| Jobs poll/cancel authorization | `module.reference_data.integration.reference_pack_import_jobs_re_derive_route_famil_dd456a1690` | Preserve current admin recheck at read/cancel time. |
| Cohort/publication/fallback | `module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` | Multiple integration selectors, including integrity, precedence, mixed refresh and worker isolation. |
| UI state/transport | `module.reference_data.frontend.administration_query_ownership`, `module.reference_data.frontend.administration_recovery`, `module.reference_data.frontend.administration_transport` | Node-unit model/controller and browser-unit service assertions; selected titles retained in manifest. |
| UI panel | `module.reference_data.frontend.the_deployment_admin_reference_pack_panel_starts_1cf94e4f79`; `module.reference_data.frontend.the_landing_surface_reference_pack_admin_panel_i_819d1aade7` | Panel behavior and non-admin hiding; no backend permission inferred from hiding controls. |
| Browser continuity | `module.reference_data.browser.administration_async_recovery`; `module.reference_data.browser.administration_real_lifecycle`; `module.reference_data.browser.deployment_admin_browser_sessions_expose_referen_8c10ae0a9c` | Exact three browser rows in `apps/web/e2e/reference-pack.spec.ts`. |
| Progress timing | `module.reference_data.measurement.import_progress`, `refresh_progress`, `reverify_progress` (same prefix); `module.reference_data.frontend.progress_paint` | Three timed browser rows plus harness-node paint assertions. Core 05 applies only if publishing claim-bearing results; ordinary test presence is not timing evidence. |

Additional-source command discovery inspected `tools/harness/diagnostics/owner-diagnostics.mjs`, `tools/harness/diagnostics/owner-diagnostics-cli.mjs`, `tools/harness/execution/run-make-node-tool.sh`, `tools/harness/execution/run-make-node-tool-cli.mjs`, Markdown scripts, generated public help and Make includes. Only public Make targets are proposed for managed verification.

## 9. Top-Level Work Tracker

| ID | Current LC work item | Status | Dependencies / trace | Exit condition |
| --- | --- | --- | --- | --- |
| LC-DOC | Adopt the bounded LC plan in this tracker | DONE | Approved document plan; LC-E0; LC-V0; final LC document checkpoint | Current summary reconciled; Markdown and preservation checks pass |
| LC-P0 | Baseline and removal ledger | TODO | LC-F1–F6; current source revalidation | Caller/input/assertion map and immutable identities complete |
| LC-P1 | Dead-code removal / LC-W1 | BLOCKED | LC-P0; LC-RB1 | Focused checks and affected independent module builds pass |
| LC-P2 | Transitional-tool retirement / LC-W2–W3 | BLOCKED | LC-P0; LC-RB1 | Stable generation, preserved releases and current witnesses pass |
| LC-P3 | Boundary qualification / LC-W4 | BLOCKED | LC-P1 and LC-P2; LC-RB2 | Affected generation/build closure and negative probes pass in isolation |
| LC-P4 | Final qualification / LC-W5 | BLOCKED | LC-P3; LC-RB3 | Frozen-source portfolio and required owner evidence audit pass |
| LC-P5 | Handoff and closure | BLOCKED | LC-P4 | Findings, evidence, operator guidance and remaining limits reconciled |

The following work items are historical I8 and earlier completion records.

| ID | Current work item / workflow | Status | Dependencies / trace | Exit condition |
| --- | --- | --- | --- | --- |
| T-034 | I8 document update | DONE | E-I8-01–06; user-approved document plan; V-49 | Reconciled tracker, Markdown/whitespace pass and tracker-only preservation; final transcription check recorded in §22 |
| T-035 | S-37 / WF-I8-00 | DONE | E-I8-S37; F-33/34; C-12/13/15/18/19 | Caller/assertion map and all three baseline rows pass; 497 immutable fingerprints retained |
| T-036 | S-38 / WF-I8-01 | DONE | E-I8-S38/40; §6; F-33; V-46 | Entire reissue path retired; unsupported-input/no-output assertion and final qualification pass |
| T-037 | S-39 / WF-I8-02 | DONE | E-I8-S39/40; §6; F-34; V-47 | Minimal helper retains both live rejection cases and other ZIP callers; final qualification passes |
| T-038 | S-40 / WF-I8-03 | DONE | E-I8-S40; §6; V-48 | Final compatible scoped qualification, documentation and preservation checks pass; handoff complete |

Completed I6 work is recorded below; completed I7 work remains in §21. I8 implementation rows become DONE only through their separate execution checkpoints.

| ID | Work item / workflow | Status | Dependencies / trace | Exit condition |
| --- | --- | --- | --- | --- |
| T-028 | I6 document planning/update | DONE | I6-00; E-I6-01–06; F-26–29; I6-02 | Current tracker reconciled, Markdown passed and tracker-only write boundary verified; final transcription check recorded below |
| T-029 | S-28 / WF-I6-00 | DONE | User implementation authorization; V-36 | Complete current-contract characterization and assertion map |
| T-030 | S-29 / WF-I6-01 | DONE | RB-004; V-37; prerequisites in §6 | Compiler moved with live Windows behavior preserved |
| T-031 | S-30 / WF-I6-02 | DONE | RB-004; V-38; prerequisites in §6 | Proposal execution retired; current audit entry path validated |
| T-032 | S-31 / WF-I6-03 | DONE | RB-004; V-39/40; prerequisites in §6 | Report v2 and test/input cleanup complete |
| T-033 | S-32 / WF-I6-04 | DONE | Earlier slice completion; V-40/41 | Frozen-source qualification and final handoff |

Historical T-001–009 completed initial planning; T-010/011/013/014 were superseded by the historical execution chain below. T-012 remains F-12's intentional deferral. None of its DONE statuses completes S-28–32.

| ID | Workstream | Status | Evidence / remaining gate |
| --- | --- | --- | --- |
| T-015 | S-04 specification/tracker | DONE | E-11: owner amendment and navigation aligned; Markdown lint passed. |
| T-016 | S-08 contract projections | DONE | E-13: regenerated contracts, schema, generated drift, compatibility and owner contract assertion pass. |
| T-017 | S-01 regression evidence | DONE | E-15: all three intended regression rows fail before fixes; failure is product/test_assertion_failure, cleanup passed. |
| T-018 | S-02 family authorization | DONE | E-16: owner-bound pre-mux gate, one principal, declared CSRF; authorization and shared composition checks pass. |
| T-019 | S-05 operation routing | DONE | E-17: direct operations, once-decoded PathValue, explicit removal decoder; routing/upload/lifecycle checks pass. |
| T-020 | S-06 query admission | DONE | E-18: eight non-list operations reject malformed/pagination/unknown input before operation work; list preserved. |
| T-021 | S-07 consumer contract | DONE | E-19: public consumer/constructor/assignment obligations documented; integrity and Markdown checks pass. |
| T-022 | S-03 validation/handoff | DONE | E-24/25: final targeted tests/source checks, compatibility, Markdown and preservation review complete. CI removed from requirements. |
| T-023 | S-09 contract/evidence reconciliation | DONE | §16; current authorization, bounded findings and validation mapping. |
| T-024 | S-10 dead-helper cleanup | DONE | After S-09; remove unused APIs without a parallel test verifier. |
| T-025 | S-11 lifecycle consolidation | DONE | After S-10; live-path assertions precede removal of model-only algorithms. |
| T-026 | S-12 cancellation-aware cleanup | DONE | After S-11; context-bearing close and recoverable terminal ownership. |
| T-027 | S-13 targeted qualification/handoff | DONE | I2-05: fourteen fresh product rows, source checks, Markdown and preservation review complete. |

## 10. Session Handoff Log

Current document session `reference-lc-docs`, 2026-11-08 America/New_York: the approved LC plan is adopted into this tracker only. LC-E0 records the incoming dirty source/index and file identities; LC-V0 and the final LC checkpoint record passing documentation validation and tracker-only preservation. Future LC implementation and production-readiness acceptance remain unexecuted.

Historical implementation session `reference-i8-implementation-xt30lmqv`, 2026-11-08 America/New_York: S-37–40 are DONE with sequential checkpoints; finalization, executable checks, documentation and preservation pass. F-33/F-34 are closed by E-I8-S40. §22 retains source identities, exact commands, native sessions, summaries and compatibility limits. The initial tracker edits are preserved; the only production change is offline Base-command retirement, accompanied by regression/fixture changes and witness admission removal. No frontend, runtime schema, authorization or lifecycle contract changed.

Historical document session `reference-i8-docs`, 2026-11-08 America/New_York: the user approved implementation of the document-update plan only. E-I8-01–05 and §22 record revalidated source, scope, tests/harness discovery and remaining future gates. Scope/authority and backend changes are planning-only; frontend/rendered review is not applicable because no UI change is planned; contract/codegen changes are not planned; security/rejection guarantees are preserved. Only the tracker is touched. Final documentation/preservation results are recorded at the end of §22.

Historical document session: `reference-i6-docs`, 2026-11-07 in the supplied America/New_York context. I6-00 was the clean `a8f2fe2d…` baseline. That session admitted only tracker writes; §§20–21 record subsequent I6/I7 implementation. The tables below preserve the original `reference-data-planning-2026-11-07` session, when no tracker existed, and its later execution history.

### Scope and authority

I6 handoff — S-28–32 implementation authorized; Reference Pack 0.4.0/major 3 remains governing. I6-03 onward records execution; the rows below preserve earlier planning history.

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-07, planning prelude | Codex / prior Plan-mode evidence | Framework-first orientation; tracker absent | Inspected framework, instructions, owner passages and representative target sources; touched none | `make help ALL=1`; `make task-guide OWNER=module.reference_data DETAIL=coverage`; read-only Git/search/source inspection | Both Make discovery commands reported successful; no product validation | Exhaustive file inspection and tracker creation remained | Completed in execution session below; preserve this history as supplied prior-session evidence. |
| 2026-11-07, execution session | Codex / reference-data-planning | Planning artifact complete; implementation not authorized | Inspected E-00–09/O-01–07 and all §2 files; touched tracker only | `git status --porcelain=v1 --untracked-files=all`, `git branch --show-current`, `git rev-parse HEAD`, `git ls-files`; Python metadata/content reads; `rg`, `sed`, `cat` | Target/label/scope and baseline reconciled; final source/index comparison recorded in §12 | No blocker to planning completion; future RB-001–003 | Revalidate HEAD/target/tracker before a later authorized task; do not infer production permission. |

### Backend module boundary

I6 handoff — Current-only offline producer consolidation is implemented; runtime boundaries and Windows utility behavior are preserved. Final boundary qualification is recorded in §20.

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-07 | Codex / reference-data-planning | Legitimate owner with private format/state, SQL/transport adapters, opaque storage and application composition | Inspected all target files; E-03–06 exact caller portions; touched tracker only | Read-only source opens, symbol/import extraction and targeted caller searches | Keep current broad boundary; F-01 auth seam requires follow-up; F-04 integrity side effect documented | RB-001/002/003 for implementation | Execute S-01 only in a later test-authorized task; preserve C-11–17/21. |

### Frontend module boundary

I6 handoff — Not applicable to I6: no presentation, transport or selector change; no UI inspection or browser validation required.

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-07 | Codex / reference-data-planning | Admin service/model/controller/hook/panel outside backend target | Inspected E-07 exact source portions and tests; touched tracker only | `rg --files`, source reads and test declaration/selector inspection | C-20 and real owner rows identified; no grid-vendor or workbook-shell relocation proposed | No rendered/behavioral pass evidence | If backend error-order changes affect UI, select V-03 transport/continuity rows; do not redesign UI under this tracker. |

### Contract and codegen

I6 handoff — S-30/31 introduce current admission and report v2. Immutable profiles/policies and runtime contracts are preserved; authored routing projections are regenerated through Make.

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-07 | Codex / reference-data-planning | Nine HTTP operations and current canonical identities mapped | Inspected E-08, target public DTO/format/contractadmission sources and owner passages; touched tracker only | JSON/source reads of generated policy/topology and OpenAPI operation objects | Generated roots/files protected; current major-3/v2 adoption distinct from code observations | None for planning; drift checks not run | Preserve outputs. If authored contracts change later, discover producer Make target and regenerate; never hand-edit OpenAPI or generated Go/TS. |

### Tests and harness

I6 handoff — Candidate assertions have current destinations and focused tests pass. Final-source tests, checks and offline qualification are recorded in §20; earlier counts remain historical.

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-07 | Codex / reference-data-planning | 93 target test files inventoried; 84 catalog rows/38 service-backed confirmed | Inspected E-09, tests/support, frontend rows, Make help inputs and Markdown side-effect scripts; touched tracker only | `make task-guide OWNER=module.reference_data DETAIL=coverage` (exit 0); read-only catalog/source parsing | Command reports row accounting only. No product tests, lint, generation or finalization ran | V-00 skipped; RB-003 no behavioral baseline | S-01/03 select tests through Make, retain run/summary references and diagnose before rerunning. Do not use phase labels as package boundaries. |

### Security and authorization

I6 handoff — No runtime authorization change. Preserve offline acquisition boundary, pinned bytes, strict input validation and fixture/deployment trust separation; C-18.

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-07 | Codex / reference-data-planning | Static family-order mismatch identified; valid-operation checks and common Jobs auth already present | Inspected HTTP adapter/request code, exact shared mux/dispatcher, HTTP auth, upload/pagination/Jobs tests and Core 01 §17.4; touched tracker only | Targeted searches then exact source reads | F-01 is implementation drift, not owner contradiction; no exploit or runtime test performed | RB-001/002/003 | Confirm matrix without changing requirements; authorize narrow error-order correction separately. |

### Open risks and next session

I6 handoff — RB-004 and F-26–29 are resolved. S-28–32, final qualification and handoff are complete; see E-I6-S32.

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-07 | Codex / reference-data-planning | File-level planning complete; no production refactor warranted beyond separately scoped auth correction | Inspected inventory/contract/workflow/slice/row maps; touched tracker only | Read-only Markdown structure/path-set/dependency/scope review; final Git/index/content comparison | S-00 complete; S-01/02 not execution-ready; canonical lint explicitly skipped | RB-001–003; V-00 scope exception | Start from F-01/C-22 and S-01; revalidate live identity, resolve gate placement, select exact rows. No rediscovery of the full target needed unless baseline changed. |

Read-only exploratory commands included unsuccessful searches of guessed Core/route-file paths and a nonexistent Make subdirectory, followed by discovery of the actual files. Broad output and two metadata dumps were truncated; smaller exact-file reads/compact metadata extraction replaced them before relying on their contents. These were inspection issues, not failed validation runs. No files were restored or reverted to hide command outputs.

### I6 evidence register

E-I6-01–06 are pre-implementation inspection snapshots. E-I6-S28–32 and the implementation checkpoints below supersede their current-execution observations.

| Evidence ID | Source / exact inspected surface | Baseline / observation | Applicability |
| --- | --- | --- | --- |
| E-I6-01 | Git baseline; tracker/framework; root/backend/tooling instructions; Reference Pack §§1, 17, VERIS adoption and domain owner navigation | `a8f2fe2d…`, 2026-11-07; immutable baseline under `/tmp/reference-i6-docs-iigz5fsh` | Current scope/admission; broad historical claims remain historical |
| E-I6-02 | `producer/audit.go`, `candidate.go`, `candidate_test.go`; `tools/reference-packs/main.go` audit dispatch | Proposal index/loader/validator and stale report members execute today; no runtime dead-function claim | F-26/27; removal needs future current-path assertions |
| E-I6-03 | `producer/package.go`; ATT&CK `source-profile.native.v2.r2.json`, native/candidate policies, candidate index; `internal/packformat/shape.go`, `content.go` | Packaging checks frozen profile/lock/policy/schema digests; native and candidate policy maps differ only in `candidate_id`/`status` | Preserve exact frozen bytes; share current admission without editing identities |
| E-I6-04 | `producer/windows_contract.go`, `envelope.go`; compiler caller search | Windows schema initialization is a live compiler consumer; signing helpers have active fixture callers | F-28; retain useful code rather than deleting by filename/test-only use |
| E-I6-05 | `source_test.go`, `attack_projection_test.go`, `ledger_test.go`, current native/ATT&CK lifecycle tests; embedded inputs; `tools/test_families/module.reference_data.json`; Make task-guide | 13 proposal schema cases, broad embed, exact current row selectors; task-guide exit 0; no tests run | F-29; existing lifecycle test imports canonical fixtures, so it cannot alone qualify new source audit/assembly |
| E-I6-06 | Make/Markdown invocation chain, protected-output policy and topology; final tracker diff and baseline comparison | Planned documentation check only; results/checkpoint in §20 | Proves document validity and bounded writes, never product readiness |

### Execution checkpoint log

| Checkpoint | Changes / decisions | Commands / results / retained artifacts | Next action |
| --- | --- | --- | --- |
| E-10 / S-04 start, 2026-11-07 | User implementation authorization supersedes historical tracker-only scope. Baseline unchanged since planning; original tracker preserved. All four product choices recorded in §§1/5–7. | Git status/HEAD and source inspection; `make help ALL=1` exit 0. No product pass claimed. | Complete S-04 lint and checkpoint before S-08. |
| E-11 / S-04 complete | Core 00/01 amendment and Reference Pack/Extensions/domain navigation updated under approved plan. No runtime or stored-data change. | `CARTULARY_TEST_RUN_ID=reference-remediation-s04-docs make lint-markdown` exit 0; `.cartulary/test-results/reference-remediation-s04-docs/adhoc/lint-markdown/tool-run-summary.json`. | S-08 starts after this checkpoint. |
| E-12 / S-08 generation diagnosis | Authored errors/OpenAPI and contract assertions updated. Initial generation stopped at expected stale release accounting; diagnostic diff reviewed: only Reference Pack response/description/path entries. Exact fingerprints reconciled without altering immutable history. | `make generate`, run `reference-remediation-s08-generate`, outer exit 2; `generate/tool-run-summary.json`, compatibility stdout report. Harness classification `harness/unknown_failure`; cause stale release fingerprint. `make explain-run` exit 0 is diagnosis only. | Regenerate after accounting repair; retain initial failure. |
| E-13 / S-08 complete | Authored errors and nine OpenAPI operations projected; exact release accounting repaired. Immutable release history unchanged. | `make generate` run `reference-remediation-s08-generate-reviewed` exit 0; `make source-check CHECK=generated` run `reference-remediation-s08-generated`, `CHECK=schema` run `reference-remediation-s08-schema`, `make openapi-compatibility-check` run `reference-remediation-s08-compatibility`, and `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.generated_open_api_and_error_registry_artifacts_0843af138f` run `reference-remediation-s08-contracts`: all exit 0. Run roots under `.cartulary/test-results/`; graph summaries and tool summary retained. | S-01 starts after this checkpoint. |
| E-14 / S-01 selection | Added three server-composed regressions and Make-derived catalog rows: `reference_pack_family_authorization_precedes_dis_d77fa37efb`, `reference_pack_operations_use_exact_versions_and_677e505495`, `reference_pack_non_list_queries_fail_before_body_fd74d4ef0e` (all under `module.reference_data.integration`). Tests inspect canonical fixtures/state only. | `make author-test-row-id` for each semantic claim/selector: exit 0. Baseline command selects exactly these three rows; known failures belong to S-02/05/06. | Format/regenerate row projections, then run pre-fix baseline. |
| E-15 / S-01 complete | Three new rows executed against pre-fix production. All fail as expected; retained query diagnostics show body error winning over required query errors. Per-branch auth/routing details exceed the bounded retained log; their row-level failures are retained. No failed row is treated as product acceptance. | `make format` run `reference-remediation-s01-format` and `make generate` run `reference-remediation-s01-generate`: exit 0. Exact three-row `make test-slice` run `reference-remediation-s01-baseline`: outer exit 2, runner exit 10, product/test_assertion_failure; `run-summary.json`, `rows/`, `unit-logs/` under that root. Cleanup 14 steps, zero failures; `make explain-run ... DETAIL=accounting` exit 0. | S-02 corrects authorization; S-05 routing; S-06 queries. |
| E-16 / S-02 complete | Added catalog-validated owner family registration and pre-mux policy; authenticate/current-admin once, propagate private principal, apply declared cookie CSRF. Generic family dispatch also supplies exact methods required for gate ordering; S-05 removes owner secondary routing. Denial probes assert no body reads/downstream calls/session slide; integration snapshots cover pack/admission state. | Focused `make test-slice` run `reference-remediation-s02-authorization` reported product/test_assertion_failure (outer exit unavailable after session compaction): claimed matrix and four existing rows passed; unclaimed fixture incorrectly supplied inactive-profile configuration. Fixed fixture only. `reference-remediation-s02-authorization-fixed`, `reference-remediation-s02-registry`, `reference-remediation-s02-dispatch`, `reference-remediation-s02-composition` all outer exit 0, `run-summary.json` under each exact root. Format exit 0. Registry rows cover binding rejection/atomicity/metadata; runtime/server rows cover unclaimed dispatch, parity, publication and readiness. Failed run cleanup: 16 steps, zero failures. | S-05 starts after this checkpoint. |
| E-17 / S-05 complete | Nine operations now bind directly to handlers. Removed manual URL parser, second unescape, missing-version heuristic/reason emission, redundant collection method dispatch and suffix-selected removal decoder. Exact methods come from generated catalog via S-02 family router. | `make format` run `reference-remediation-s05-format` and `make test-slice OWNER=module.reference_data` with exact routing, shared-upload and activation/lifecycle rows, run `reference-remediation-s05-routing`: outer exit 0. Retained `run-summary.json` and row results; includes encoded/action-named versions, HEAD/OPTIONS/PATCH, import/list/replay and integrity/lifecycle selectors. | S-06 starts after this checkpoint. |
| E-18 / S-06 complete | Added owner-local complete RawQuery parser and UTF-8 validation. Precedence is malformed, pagination, unknown; diagnostics use the safe query field. One operation gate applies it to every non-list operation after security, before application/body/replay. Removed redundant singleton validation; list contract unchanged. | `make format` run `reference-remediation-s06-format` and exact four-row `make test-slice` run `reference-remediation-s06-query`: outer exit 0. Rows: non-list query matrix, authorization matrix, request/denial probes and existing import/list/replay. Retained `run-summary.json` and row results; probes establish no body reads/application invocation/session slide, integration snapshots establish no rejected mutation or replay effects. | S-07 starts after this checkpoint. |
| E-19 / S-07 complete | Documented Consumer/NewConsumer/RegistryAssignments/RegistryAssignment integrity side effects, operational-failure distinction, independent publication, full source rollback and transaction lifetime. No signatures or behavior changed. | `make lint-markdown` run `reference-remediation-s07-docs` and exact integrity/lifecycle row via `make test-slice` run `reference-remediation-s07-integrity`: outer exit 0. Existing assertions cover rollback, fallback, operational failures, Base readiness and healthy replacement. | S-03 starts after this checkpoint. |
| E-20 / S-03 original selection (superseded by E-22) | Final CI plan selects 2,487 rows / 444 work units, including 83 Reference Data rows, all three new regressions, six frontend and three administration browser rows, shared runtime/parity/auth checks, generated/schema/catalog/boundary obligations. Only Reference Data release timing rows and explicit upstream Entra release import are excluded; none is required by this transport change. | `make target-plan TARGET=ci JSON=1` exit 0; plan inspected by exact row/evidence membership. `harness.source_schema_selection` is `harness.json_shape_check`, already in CI. Additional final obligations: OpenAPI compatibility and Markdown lint. No redundant `make check` or complete owner rerun. | Run `make agent-finalize`, review generated diff, freeze source, then CI. Retained-run maintenance skipped because no compatible successful full warm-check root exists; do not create one solely for maintenance. |
| E-21 / S-03 freeze | `agent-finalize` passed; generated semantic review found only 30 changed Reference Pack OpenAPI JSON pointers (descriptions/security responses), the intended error reasons, derived digests and test-catalog input digest. No unrelated generated changes. Product source is frozen for final validation. | `make agent-finalize`, run `reference-remediation-s03-finalize`: outer exit 0. `git diff --check`: exit 0. `make openapi-compatibility-check`, run `reference-remediation-s03-compatibility`: outer exit 0. CI launched with native session 87028, run `reference-remediation-s03-ci`; manifest source digest `sha256:a5e43983d576ff2ee70beaba0b4cf8cf38dc4f3f639300092a30405a0d0771b8`, graph digest `sha256:91d809892828225351e44f5d8acd1be59cee3f227693cac8f03493de9a70da89`. | Superseded by E-22: cancel CI as requested, preserve its outcome, and complete targeted validation. |
| E-22 / S-03 user-directed validation change | The user explicitly requested stopping the current CI run and removing it from plan requirements, using targeted tests instead. Sent Ctrl-C through the verified native session 87028; retained root is `reference-remediation-s03-ci`. Its partial/cancelled results will not establish acceptance. Finalization and final OpenAPI compatibility already passed. | Final targeted portfolio: Reference Data authorization/exact-route/query matrices, existing upload/list/replay/Jobs/lifecycle-integrity, construction/request/projection/list-query units, frontend transport/query/recovery and three admin browser rows; shared HTTP registry/parity/dispatch/server admission and existing auth CSRF/session rows. Source checks: generated, schema, catalog and backend boundary. Final Markdown lint. | Wait for CI cleanup and owning command exit; launch each final target with a fresh explicit run ID. No CI/check/full-owner/release-measurement requirement remains. |
| E-23 / S-03 cancellation complete | CI stopped at the user's request; no test failures were reported before cancellation. Its 42 passed / 402 cancelled work units are partial evidence only and are excluded from final acceptance. | Native session 87028 returned outer exit 130. `reference-remediation-s03-ci/run-summary.json`: interrupted/cancelled_or_interrupted; `cleanup-results.json`: all 92 outcomes completed; `retained-secret-scan.json`: pass. Final targeted test runs launched under `reference-remediation-s03-targeted-{reference,registry,dispatch,composition,auth}`: 17 + 2 + 2 + 2 + 2 selected rows. Source checks launched under corresponding `targeted-{generated,schema,catalog,boundary}` roots. | Observe these exact roots and native sessions; final handoff reports targeted acceptance only. |
| E-24 / S-03 targeted validation passed | All **25 selected rows passed**: 17 Reference Data and eight shared HTTP/server/auth rows. The Reference Data run includes seven integration, four unit, three frontend and three browser rows; it is not a full-owner claim. All final test/source/compatibility manifests use the frozen source digest recorded in E-21. | Every listed final command returned outer exit 0. Reference Data root `reference-remediation-s03-targeted-reference`: 18 work units passed, 17 rows passed, 33 cleanup outcomes completed and retained scan passed. Shared roots `targeted-registry`, `targeted-dispatch`, `targeted-composition`, `targeted-auth` each passed two rows. `source-check CHECK=generated`, `CHECK=schema`, `CHECK=catalog`, and `lint CHECK=backend-boundary` passed in their corresponding `reference-remediation-s03-targeted-*` roots. Final compatibility root `reference-remediation-s03-compatibility` passed. All retained scans passed. | Complete final Markdown lint and handoff checkpoint. |
| E-25 / S-03 complete | Final handoff records all required remediation dispositions, exact retained roots, source identity, client migration, intentionally deferred/no-action boundaries and user-directed CI cancellation. All required targeted validation passed; no remaining implementation blocker. | `CARTULARY_TEST_RUN_ID=reference-remediation-s03-handoff-markdown make lint-markdown`: outer exit 0; `adhoc/lint-markdown/tool-run-summary.json`. `git diff --check`: exit 0. Final result-record edits reviewed for Markdown structure. | Ready for human review of specification/projection fidelity and the working-tree change. No deployment or commit performed. |

### Final validation handoff

All root names below are relative to `.cartulary/test-results/` and prefixed `reference-remediation-s03-`. Each graph root retains `run-manifest.json` (exact command/ROWS and source identity), `run-summary.json`, selected `rows/`, cleanup and scan results. These are execution records, not an evidence audit or a claim that specifications are complete.

| Make command / scope | Run suffix | Result |
| --- | --- | --- |
| `make agent-finalize` | `finalize` | PASS; generated changes reviewed before source freeze. |
| `make test-slice OWNER=module.reference_data ROWS=...` — the exact 17-row E-22/24 selection | `targeted-reference` | PASS; all selected backend, frontend and browser behavior. |
| `make test-slice OWNER=platform.openapi ROWS=...` — runtime catalog and runtime contract parity | `targeted-registry` | PASS; two selected rows. |
| `make test-slice OWNER=platform.httpruntime ROWS=...` — reserved extension dispatch and handler construction | `targeted-dispatch` | PASS; two selected rows. |
| `make test-slice OWNER=app.server ROWS=...` — extension publication and readiness admission | `targeted-composition` | PASS; two selected rows. |
| `make test-slice OWNER=module.auth ROWS=...` — existing cookie CSRF and session inspection | `targeted-auth` | PASS; two selected rows. |
| `make source-check CHECK=generated` | `targeted-generated` | PASS. |
| `make source-check CHECK=schema` | `targeted-schema` | PASS. |
| `make source-check CHECK=catalog` | `targeted-catalog` | PASS. |
| `make lint CHECK=backend-boundary` | `targeted-boundary` | PASS. |
| `make openapi-compatibility-check` | `compatibility` | PASS. |
| `make lint-markdown` | `handoff-markdown` | PASS; tool summary under `adhoc/lint-markdown/`. |

`make ci` is **removed from requirements**, cancelled at the user's direction, and supplies no final acceptance evidence. `make check`, full-owner reruns, release-only measurements, explicit upstream Entra release import, visual goldens and deployment qualification are not required by this scope. Retained warm-check maintenance was skipped because no compatible successful full warm-check root exists; no extra check was run just to create one. Initial expected red regressions and the corrected inactive-profile fixture failure remain retained in E-12/15/16.

### Current finding dispositions and client handoff

| Findings | Disposition / maintenance boundary |
| --- | --- |
| F-01 / F-10 | Implemented by S-01/S-02; R-03 additionally closes structural path admission before redirects. P-01/P-02 resolve cross-family precedence. Current targeted evidence is in §§14–15; §13 preserves the earlier broader portfolio. |
| F-04 | S-07 exported consumer/transaction contract documents existing effects; no consumer redesign. |
| F-14 | S-04 plus R-01 align the canonical glossary, closed applicability vocabulary, owner navigation and required-Base/optional-enrichment distinction. |
| F-15 | S-04/S-08/S-05 exact operations and versions; R-02 proves successful encoded/action-named mutations and literal replays, exact Job scope and removal-reason rejection. |
| F-16 | S-04/S-08/S-06 closed non-list queries with deterministic safe errors and early rejection. |
| F-17 | S-04/S-08 complete inventory/security-response/error projections; generated outputs refreshed through Make. |
| F-02/03/05–09/11/13 | Intentional no action on existing facade, SQL/transaction, portability/recovery, offline producer, fixture, workbook and UI boundaries; the transport amendment does not justify expanding those surfaces. |
| F-12 | Intentionally deferred package subdivision; reopen only for demonstrated coupling/change-locality benefit. No required work remains hidden in this deferral. |

Client migration is limited to transport: remove query members from non-list calls, use the declared GET/POST operations, include exact versions in action URLs, and stop expecting `pack_version_required` or implicit HEAD/OPTIONS. Malformed unauthorized requests now receive the family authentication/authorization outcome first; ordinary cookie sessions without admin rights receive authorization denial before CSRF. Cookie mutation callers still supply CSRF. Successful resources, receipts and Jobs retain their contracts. No data migration, compatibility reader, receipt rewrite, profile-major change or pack-format change is needed.

The shared router accepts a complete generated owner family and owner-supplied policy; adding a declared operation requires binding it at startup. Reference Data owns its admission rules; unrelated route families retain their current authentication flow. New non-list operations inherit the closed query gate unless an explicit owner amendment changes their contract. Future consumer callers must respect the documented guard/rollback lifetime rather than assuming a pure read.

## 11. Open Questions and Blockers

LC has no unresolved product-scope choice: retain adopted features and remove only proven obsolete paths. LC-RB1–3 below are future characterization, integration and qualification gates, not new licensing or deployment approval requests. LC-DOC completion does not resolve them. Historical blockers follow.

I8 has no unresolved design/authority question. The user chose retirement rather than a new general builder and explicitly authorized implementation. E-I8-S37 resolves the characterization gate; E-I8-S38/39 satisfy the sequential implementation exits in §6:

| ID | Question or blocker | Affected scope | Needed evidence | Status / resolution |
| --- | --- | --- | --- | --- |
| RB-005 | Current executable pre-removal characterization | S-38; F-33; C-18 and preserved C-12/15 | S-37 caller/assertion map, three fresh passing baseline rows and immutable fingerprints under V-46–48 | DONE; E-I8-S37 resolves the missing-characterization gate |

S-37–40 are DONE; E-I8-S40 closes the bounded effort. No unresolved I8 blocker remains. Historical I6/I7 blockers and completed validation follow and do not satisfy I8 gates:

| ID | Question or blocker | Affected scope | Needed evidence | Status / resolution |
| --- | --- | --- | --- | --- |
| RB-004 | Missing current-path pre-move characterization, including replacement of candidate-only assertions and CLI/report/assembly coverage | S-29–31; C-18/19; F-26–29 | Passing S-28 assertion destinations and focused baseline under V-36; later slices follow §6 | RESOLVED; E-I6-S28 demonstrates all replacement assertions before removal; E-I6-S30/31 cover current admission and reports |

Historical blockers below remain closed for their original scopes; RB-001 is not authorization for I6 production work.

| ID | Current disposition | Resolution / remaining condition |
| --- | --- | --- |
| RB-001 | DONE | User explicitly authorized implementation of the proposed plan. |
| RB-002 | DONE | Owner-bound registry family middleware before mux, after unclaimed dispatch; private authenticated-principal context; exact operation binding. |
| RB-003 | DONE | S-01 retained expected pre-fix failures; S-02–07 and final targeted S-03 evidence pass. CI cancellation is user-directed and is not a blocker. |

## 12. Binary Completion Criteria

Current LC outcomes are separate: the bounded design is specified; LC-P0 is ready for later inspection while implementation slices await their entry gates; production readiness is unproven until LC-P4/P5. This invocation authorizes only the tracker update. LC-V0 and the final checkpoint establish document completion and preservation, not product acceptance. The following table preserves historical I8 acceptance.

| Outcome | Historical I8 statement |
| --- | --- |
| Planning completeness | complete for the bounded F-33/34 retirement/simplification design; normal artifact, no pending required source inspection or owner decision; not an exhaustive audit or production-readiness certification |
| Per-slice readiness | S-37–40 DONE; all predecessor gates and scoped acceptance criteria satisfied |
| Implementation authorization | Explicit user request to implement the complete F-33/F-34 plan, including S-37–40 and sequential tracker checkpoints |
| Current validation | V-46/47 focused and final rows pass; V-48 finalization, final portfolio, boundary/generated/schema and producer-build checks pass on the frozen source. S-40 documentation and preservation checks pass. V-49 is historical document-only validation. |
| Session preservation | E-I8-S40 compares against I8-03: only the six authorized changed files and one deletion differ; all 497 immutable artifacts, historical §§13–21, HEAD, branch, index content/identity and staged entries are preserved. Earlier tracker-only comparisons remain historical. |

Planning criteria are met: canonical target/output admitted; all 265 paths accounted at declared depth; required removal sources directly inspected; owner-required versus observed behavior separated; F-33/34 trace through contracts, slices, V-46–48 and evidence; §6 prerequisites are acyclic; each slice has readiness/exit/rollback terms; offline removal was separate from test simplification; exact validation selectors are discovered; historical passes remain historical. V-49 and content/index/identity preservation passed at E-I8-06; the completion transcription receives the final check specified in §22. The final handoff in §22 closes this bounded effort; future work starts from a newly revalidated baseline.

The table below preserves I6 acceptance. I7's separate completion is recorded in §21; neither historical portfolio closes I8 gates.

| Outcome | I6 statement at completion |
| --- | --- |
| Planning completeness | Complete for the bounded producer-retirement design in §20; no exhaustive module correctness/dead-code audit or production-readiness certification. Inventory accounts for every current path at its declared depth. |
| Per-slice readiness | S-28–32 DONE; E-I6-S28–32 records all gates. Every workstream was checkpointed before its successor. |
| Implementation authorization | User approved the remediation plan and requested implementation, including isolated retained-source qualification. |
| Current validation | Five final selected rows, generated/schema/boundary checks, reproducible build, v20.0/v20.1 offline audit/assembly/replay, preservation and Markdown checks pass; exact evidence in §20. |
| Session write-boundary verification | I6-02 verifies the historical document-only session. I6-03 establishes the separate implementation baseline; final source, immutable evidence and index preservation are checked in S-32. |
| Historical completion | Original transport work and S-09–27 remain historical completions; §19 is the latest scoped fatal-finalization closure. Its 15-row portfolio and earlier 21-row §17 portfolio retain original source applicability only. |

I6 planning acceptance: target/output admitted; inventory depth explicit; F-26–29 trace through C-18/19, S-28–32, V-36–41 and E-I6-01–06; workflow prerequisites resolve without cycles; every slice has entry/exit/rollback criteria; behavior changes are separate from compiler movement; immutable source artifacts remain unchanged; current and historical evidence are distinguished; documentation validation and tracker-only preservation pass. I6 implementation acceptance was S-32's separate gate. No deployment, data conversion, UI, timed publication, CI/full-owner or release claim was part of that document task.

The previous §12 current-state summary described S-18 rather than the later §19 completion. Its historical results remain in §§17–19; this replacement makes the current planning posture unambiguous without revoking those results.

## 13. Post-review remediation and handoff

The user authorized all three review findings on 2026-11-07. These workstreams supplement the original slices; earlier completion statements do not establish their completion. Preserve prior evidence and use only targeted validation. Do not restart CI.

| Workstream | Scope, rationale and compatibility | Dependencies / risk | Exit criteria | Status |
| --- | --- | --- | --- | --- |
| R-01 / S-04 glossary | Align the canonical Reference pack glossary's behavior owner and applicability with the adopted owner and required-Base/optional-extension distinction. Documentation only; no client migration. | First. Stale owner navigation can misdirect future changes. | Human cross-reference review and Markdown lint; tracker checkpoint before R-02. | DONE |
| R-02 / S-05 route evidence | Exercise successful mutations against action-named and encoded versions, exact affected identities and removal-reason validation; remove vacuous claimed-admin GET matrix branches. Tests only. | R-01 complete. Negative-only coverage could conceal wrong-target dispatch. | Targeted exact-route and authorization rows pass; tracker checkpoint before R-03. | DONE |
| R-03 / S-02 path normalization | Clarify the Core transport contract for paths normalized into a reserved family; align dispatch, projections and focused tests while preserving once-decoded parameter identities. | R-02 complete. Avoid normalization that changes encoded parameter boundaries or affects unrelated families. | Claimed/unclaimed and redirect-order cases pass with exact-method/identity regressions; tracker checkpoint before handoff. | DONE |
| R-04 / S-03 validation and handoff | Finalize generated artifacts, inspect final diff, freeze implementation and run the affected targeted portfolio; record evidence and remaining limits. | R-01–03 complete. Earlier source evidence must remain historical. | Targeted tests/checks and Markdown lint pass; exact roots, source identity, migration and preservation review recorded. No CI or full-owner claim. | DONE |

R-01 review also found that the existing `base-and-extension` rows lacked a declaration in the glossary's closed applicability vocabulary. The vocabulary now explicitly defines that value without weakening either Base readiness or extension claim requirements. `current-required` remains appropriate for the canonical term because §1.2 includes vocabulary required for domain interpretation.

R-01 checkpoint: canonical glossary, context owner and applicability matrix now agree. `CARTULARY_TEST_RUN_ID=reference-review-r01-docs-final make lint-markdown` returned outer exit 0; `.cartulary/test-results/reference-review-r01-docs-final/adhoc/lint-markdown/tool-run-summary.json` reports pass. Documentation-only work ran no product tests or generators. R-02 may now begin.

R-02 checkpoint: all four action words are imported as opaque versions. Each receives encoded activate/disable/reverify/remove calls, literal-spelling idempotent replays and exact resource-state assertions; reverify also checks the durable Job member identity. Missing, null and blank removal reasons fail without pack/admission changes. Claimed-admin collection/member/summary GETs now execute assertions for cookie and bearer sessions. `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.reference_pack_family_authorization_precedes_dis_d77fa37efb,module.reference_data.integration.reference_pack_operations_use_exact_versions_and_677e505495`, run `reference-review-r02-routes-fixed`, returned outer exit 0 with both rows passed. Retained root: `.cartulary/test-results/reference-review-r02-routes-fixed`.

The first R-02 run, `reference-review-r02-routes`, returned outer exit 2 with `product/test_assertion_failure`: two new fixture imports violated the existing anti-rollback sequence rule. Authorization and the first two mutation-version cases passed. The correction supplies monotonically increasing signed fixture sequences through the existing manifest transform; no production behavior changed. Both roots retain completed cleanup (14 steps, zero failures). `reference-review-r02-format` passed. R-03 may now begin.

R-03 decision: Core 01 REQ-01-547 now defines admission matching against the original and structurally cleaned escaped path, splitting at literal slashes before decoding segments once. Unclaimed matches precede redirects for all reserved families. Core 01 §17.4 requires Reference Pack family authentication/current-admin authorization first, then rejects structural noncanonicality with ordinary HTTP 404, no Location, no URL rewrite, no operation or session extension. Encoded slashes/dots remain data. Paths traversing into or out of the family receive this protection; literal Base paths and unrelated paths retain ordinary dispatch. Core 00 adoption, Extensions import and all nine OpenAPI descriptions agree. Canonical clients and stored identities are unchanged; callers relying on incidental normalization redirects must construct canonical paths. This intentionally removes an unnecessary alias/identity compatibility burden.

R-03 checkpoint: `reference-review-r03-reference` passed the two R-02 rows with claimed/unclaimed × anonymous/non-admin/admin × cookie/bearer × method coverage for noncanonical paths and exact identity/mutation regressions. `reference-review-r03-registry` passed `platform.openapi.unit.runtime_profile_catalog`, including encoded slash/dot/double-escape preservation, no redirect/dispatch on rejected paths, and family-boundary exclusions. `reference-review-r03-dispatch` passed `platform.httpruntime.unit.new_handler_serves_embedded_root_and_assets_41dd7a2589` and `platform.httpruntime.unit.reserved_extension_dispatch_preserves_base_and_o_4409626a59`, including all reserved unclaimed roots and their normalized spellings. All three commands returned outer exit 0; cleanup completed (14 / 10 / 10 steps, zero failures). Roots are `.cartulary/test-results/<run-id>`, each with `run-summary.json`; source digest `sha256:6a5d63631d25df53ccbd407c811415f98011d25ba9f68ad281d9527c10c13c06`.

`reference-review-r03-format` and `reference-review-r03-generate-fixed` passed. Initial `reference-review-r03-generate` stopped with outer exit 2 and the harness classification `harness/unknown_failure`; its typed tool summary identifies stale OpenAPI release-change fingerprints after the description edits. Reviewed exactly nine changed Reference Pack fingerprints (seven description additions and two already-added operation subtrees against immutable 1.0.0), updated only their unreleased 2.0.0 authored accounting, then regenerated successfully through Make. No generated file was hand-edited. R-04 may now begin.

### R-04 final targeted portfolio and handoff

`CARTULARY_TEST_RUN_ID=reference-review-r04-finalize make agent-finalize` passed with outer exit 0. Reviewed the generated diff and froze implementation. R-03 and every R-04 manifest share source digest `sha256:6a5d63631d25df53ccbd407c811415f98011d25ba9f68ad281d9527c10c13c06`; finalization introduced no source changes, so the five passing R-03 rows remain compatible and were not needlessly rerun. R-04 adds six product rows, for **11 passing selected product rows** on this source, plus source and compatibility checks. This is selected-scope evidence, not a complete-owner audit.

All commands below use `CARTULARY_TEST_RUN_ID=<run-id>` and retain `.cartulary/test-results/<run-id>/run-summary.json`, `run-manifest.json`, `cleanup-results.json` and `retained-secret-scan.json`. Every owning command returned outer exit 0. All cleanup outcomes completed and all retained scans passed.

| Command / selected obligation | Run ID | Result |
| --- | --- | --- |
| R-03 Reference Data exact-route and family-authorization rows listed above | `reference-review-r03-reference` | PASS, two rows |
| R-03 runtime family registration and metadata | `reference-review-r03-registry` | PASS, one row |
| R-03 reserved dispatch and handler construction | `reference-review-r03-dispatch` | PASS, two rows |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.reference_pack_non_list_queries_fail_before_body_fd74d4ef0e,module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda,module.reference_data.unit.generated_open_api_and_error_registry_artifacts_0843af138f` | `reference-review-r04-reference` | PASS, three rows; closed-query precedence, denial effects, request decoding and contract projections |
| `make test-slice OWNER=platform.openapi ROWS=platform.openapi.unit.runtime_contract_parity_4fc98c2b71` | `reference-review-r04-parity` | PASS, one row; generated/runtime parity |
| `make test-slice OWNER=app.server ROWS=app.server.integration.unclaimed_reserved_extension_families_return_can_57db0f8a45,app.server.unit.extension_readiness_admission_gate_05faff0b60` | `reference-review-r04-composition` | PASS, two rows; actual unclaimed dispatch and outer process admission |
| `make source-check CHECK=generated` | `reference-review-r04-generated` | PASS |
| `make source-check CHECK=schema` | `reference-review-r04-schema` | PASS |
| `make source-check CHECK=catalog` | `reference-review-r04-catalog` | PASS |
| `make lint CHECK=backend-boundary` | `reference-review-r04-boundary` | PASS |
| `make openapi-compatibility-check` | `reference-review-r04-compatibility` | PASS |

The three review findings are resolved in documentation, tests and transport/specification respectively. The final diff preserves the earlier remediation and adds only the documented scope. No pack data migration, profile-major change, environment reset, deployment or commit is involved. For claimed Reference Pack URLs, the additional client impact concerns noncanonical paths: clients must use structurally canonical paths rather than relying on redirects; valid percent-encoded parameter identities still work. Unclaimed reserved families recognize normalized spellings before redirects. No compatibility alias remains to maintain.

CI remains cancelled and excluded at the user's direction. No `make ci`, `make check`, full-owner run, release measurement or browser/UI suite was launched for this follow-up. Browser/UI code and successful canonical transport shapes are unchanged; the targeted backend/contract portfolio covers the changed behavior. Retained warm-check maintenance remains skipped because no compatible successful full warm-check root exists. Human review of specification/projection fidelity remains the normal review boundary; no unresolved implementation blocker or review finding remains.

R-04 completion checkpoint: `CARTULARY_TEST_RUN_ID=reference-review-r04-handoff-markdown make lint-markdown` returned outer exit 0; `.cartulary/test-results/reference-review-r04-handoff-markdown/adhoc/lint-markdown/tool-run-summary.json` reports pass. `git diff --check` passed. Final status-only tracker edits preserve the linted Markdown structure. All four follow-up workstreams are DONE, each predecessor checkpoint was written before its successor began, and the effort ends with this validation/handoff slice.

## 14. Cross-family precedence follow-up

The next review identified a remaining contradiction: original and cleaned path forms can match different families, but Core 01 REQ-01-548 and Extensions EXT-REQ-096 still listed claimed-family handling before inactive-family handling. The user authorized this follow-up, targeted validation only, and a final handoff. Earlier R-03/R-04 completion statements do not close this newly identified gap.

| Workstream | Remediation and rationale | Dependencies / risks | Exit criteria | Status |
| --- | --- | --- | --- | --- |
| P-01 / S-04 precedence contract | Align Core 01 REQ-01-548, Extensions EXT-REQ-096/098 and Core 00 adoption: direct Base matches first; any inactive match in either path form before claimed handling; only then ordinary claimed/unknown handling. A claimed match cannot cancel an inactive boundary. Existing runtime behavior and canonical clients remain unchanged; no migration or alias. | First. Conflicting normative order could cause future authentication/redirect drift. | Both traversal examples and reciprocal claim states are explicit; Markdown lint passes; tracker checkpoint before P-02. | DONE |
| P-02 / S-02 traversal evidence | Add focused shared-dispatch and real-server cases for both traversal directions and reciprocal claim states, exact inactive error identity, no redirect/body consumption/downstream effects, and credential/method independence. Preserve the existing runtime rule if it satisfies the clarified contract. | P-01 complete. Status-only 404 assertions could hide wrong-family or fallback behavior. | Targeted dispatch and family-authorization rows pass; tracker checkpoint before P-03. | DONE |
| P-03 / S-03 validation and handoff | Finalize, freeze source, review diff and retained evidence, complete affected targeted checks and document limits and final disposition. | P-02 complete. Do not reuse incompatible earlier-source evidence or restart CI. | Compatible selected evidence, final Markdown check, no remaining precedence contradiction, and completed handoff. | DONE |

The chosen rule intentionally preserves original Base-route precedence while making an inactive match decisive across both escaped-path forms. Response diagnostics must name a matching inactive family. No primary owner contradiction remains between the revised Core and Extensions lists. This workstream changes specification/adoption prose only; executable behavior and tests must not read it.

P-01 checkpoint: Core and Extensions now place inactive matching before claimed handling and explicitly cover both traversal directions, reciprocal claim states, diagnostic ownership and structural cleaning toward Base. Core 00 records their coordinated adoption. `CARTULARY_TEST_RUN_ID=reference-precedence-p01-docs make lint-markdown` returned outer exit 0; `.cartulary/test-results/reference-precedence-p01-docs/adhoc/lint-markdown/tool-run-summary.json` reports pass. No product test or generator ran for this documentation-only workstream. P-02 may begin.

P-02 checkpoint: shared dispatch now tests both traversal directions with Reference Pack and Import claim states swapped, plus traversal toward Base and direct-route positive controls. It asserts the exact inactive error identity, no redirect, no claimed policy/operation invocation and no body reads for GET/POST/HEAD/OPTIONS/PATCH. The real-server family-authorization row exercises both directions for anonymous, ordinary-user and admin callers with cookie/bearer transport, malformed query/body and absent cookie CSRF; existing durable pack/job/idempotency snapshots remain unchanged. HEAD envelope identity is checked at handler level because the HTTP server suppresses its body. The existing runtime passes this clarified contract without a production change.

`make format`, run `reference-precedence-p02-format`, returned outer exit 0. Targeted `make test-slice` runs `reference-precedence-p02-dispatch` (the reserved-extension dispatch row) and `reference-precedence-p02-reference` (the Reference Pack family-authorization row) both report pass at `.cartulary/test-results/<run-id>/run-summary.json`; their source digest is `sha256:beb4b641d44ff9f0a994bbc3ce9efb2e525038f2667ee0daf8959419fe1ed2c8`. Cleanup completed with 10/14 steps and zero failures; both retained-secret scans passed. The final native-session exits were lost in a truncated response during context handoff and are not claimed. P-03 will independently observe these two commands after finalization to close that evidence gap, rather than infer outer success from an observer. P-03 may begin.

### P-03 final targeted portfolio and handoff

`CARTULARY_TEST_RUN_ID=reference-precedence-p03-finalize make agent-finalize` returned outer exit 0. Finalization and all four final validation manifests share source digest `sha256:beb4b641d44ff9f0a994bbc3ce9efb2e525038f2667ee0daf8959419fe1ed2c8` at dirty HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`. Finalization introduced no source changes. The affected tests and the existing encoded-segment/normalization registry test provide three selected product rows on this source; no complete-owner acceptance is claimed.

Every command below uses `CARTULARY_TEST_RUN_ID=<run-id>`. Exact roots are `.cartulary/test-results/<run-id>`, with `run-manifest.json`, `run-summary.json`, `cleanup-results.json` and `retained-secret-scan.json`. All four owning commands returned outer exit 0, all summaries report pass, cleanup completed with zero failures, and retained scans passed. Finalization also completed cleanup and passed its retained scan.

| Command / selected obligation | Run ID | Result / evidence scope |
| --- | --- | --- |
| `make test-slice OWNER=platform.httpruntime ROWS=platform.httpruntime.unit.reserved_extension_dispatch_preserves_base_and_o_4409626a59` | `reference-precedence-p03-dispatch` | PASS, one row; compatible harness cache hit from P-02, not fresh execution. Independently observed final command exit and retained evidence close P-02's missing outer-exit observation. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.reference_pack_family_authorization_precedes_dis_d77fa37efb` | `reference-precedence-p03-reference` | PASS, one freshly executed service-backed row; both traversal directions, reciprocal claims, credentials/methods, error identity and durable denial effects. |
| `make test-slice OWNER=platform.openapi ROWS=platform.openapi.unit.runtime_profile_catalog` | `reference-precedence-p03-registry` | PASS, one freshly executed row; distinct encoded slash/dot/double-escape identity, canonicality, family boundary and registration controls. |
| `make source-check CHECK=catalog` | `reference-precedence-p03-catalog` | PASS; changed tests remain covered by the existing catalog selectors, with no new row or harness change. |

Final review confirms that Core 01 REQ-01-547/548, Core 00 adoption and Extensions EXT-REQ-096–098 now agree. The inactive-family wrapper precedes the claimed owner-family wrapper; both examine original and structurally cleaned escaped paths without rewriting request identity. Tests demonstrate both directions and reciprocal claim states, name the inactive family, and distinguish inactive rejection from authentication, fallback 404 or redirects. Original Base routing has a positive control, while traversal cleaning into Base remains subject to inactive-family admission. No remaining precedence contradiction or implementation gap was identified within this follow-up scope.

Compatibility and handoff: this clarification matches existing runtime behavior, adds no alias and requires no data migration, profile-major change or client change beyond the earlier canonical-path requirements in §13. Its lasting benefit is one explicit transport precedence shared by Core and Extensions, protected against future wrapper-order regressions. Earlier OpenAPI projections remain unchanged because this follow-up clarifies admission before operation selection, not any declared operation's canonical request/response contract. No generated artifact was hand-edited, and no deployment or commit was performed.

CI remains cancelled and excluded by user instruction. No `make ci`, `make check`, full-owner run, release measurement or browser/UI suite ran. The earlier broader §13 evidence remains historical rather than being relabeled as current-source acceptance. Retained warm-check maintenance is skipped because no compatible successful full warm-check root exists. Human review of specification/projection fidelity remains the normal handoff boundary.

P-03 completion checkpoint: `CARTULARY_TEST_RUN_ID=reference-precedence-p03-handoff-markdown make lint-markdown` returned outer exit 0; `.cartulary/test-results/reference-precedence-p03-handoff-markdown/adhoc/lint-markdown/tool-run-summary.json` reports pass. `git diff --check` passed. Final status-only tracker edits preserve the linted Markdown structure. All three workstreams are DONE; each predecessor checkpoint was written before its successor began. The effort ends with this validation and handoff slice, with no unresolved blocker in the authorized follow-up scope.

## 15. Final-source targeted qualification

The user requested fresh execution of the remaining routing, query, CSRF, parity and compatibility checks against the final source, followed by an updated handoff. This closes the evidence limitation identified in the subsequent read-only review. Production, test, contract and generated source remain frozen; only this tracker is edited. The expected source digest is `sha256:beb4b641d44ff9f0a994bbc3ce9efb2e525038f2667ee0daf8959419fe1ed2c8`, already finalized successfully by `reference-precedence-p03-finalize`. No repeat finalization or generation is needed for a validation/documentation-only follow-up.

| Workstream | Scope and dependency | Exit criteria | Status |
| --- | --- | --- | --- |
| Q-01 / S-03 targeted qualification | Execute seven exact rows plus OpenAPI compatibility with harness result caching disabled. Reuse the compatible §14 precedence/registration evidence without rerunning it. | Every new manifest matches the frozen final source; owning exits, rows, cleanup and retained scans pass; checkpoint recorded before Q-02. | DONE |
| Q-02 / S-03 handoff completion | After Q-01, reconcile the current evidence portfolio, exact commands/results and remaining limits; lint the updated tracker. | Markdown lint and whitespace review pass; current handoff complete, with CI still excluded. | DONE |

Q-01 selection: `make task-guide OWNER=module.reference_data DETAIL=summary`, `OWNER=module.auth` and `OWNER=platform.openapi` all returned exit 0. Exact rows were reconciled with the prior retained manifests and current catalogs. Executed invocations use `CARTULARY_TEST_RUN_ID=<run-id>` from the table below, `CARTULARY_HARNESS_CACHE_MODE=off`, and the default parent `.cartulary/test-results`.

| Command / distinct obligation | Run ID | Result |
| --- | --- | --- |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.reference_pack_operations_use_exact_versions_and_677e505495,module.reference_data.integration.reference_pack_non_list_queries_fail_before_body_fd74d4ef0e,module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda,module.reference_data.unit.generated_open_api_and_error_registry_artifacts_0843af138f` — exact mutation/version/method routing, closed-query precedence and effects, owner CSRF/denial probes, generated contract projection | `reference-final-source-reference` | PASS; four rows freshly executed, outer exit 0 |
| `make test-slice OWNER=module.auth ROWS=module.auth.support_unit.session_inspection_helpers_da94c8032f,module.auth.unit.cookie_mutations_reject_missing_or_invalid_csrf_dfaa827393` — shared session/CSRF behavior | `reference-final-source-auth` | PASS; two rows freshly executed, outer exit 0 |
| `make test-slice OWNER=platform.openapi ROWS=platform.openapi.unit.runtime_contract_parity_4fc98c2b71` — runtime/generated operation parity | `reference-final-source-parity` | PASS; one row freshly executed, outer exit 0 |
| `make openapi-compatibility-check` — unreleased 2.0.0 change accounting against immutable 1.0.0 | `reference-final-source-compatibility` | PASS; outer exit 0 |

CI, `make check`, complete-owner runs, release measurements and browser/UI reruns remain outside this requested qualification. A new source mismatch or failure must be investigated before any rerun or scope change; older-source passes will not be relabeled as current evidence.

Q-01 completion checkpoint: all four manifests report the expected final source digest above, dirty HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`, and cache mode `off`. All seven row results are `passed` with exit 0 and no failure classification. Every owning command completed with independently observed outer exit 0 (native sessions 90683, 28698, 44674 and 9341 in table order). All four run summaries report pass and zero cache hits. Cleanup completed with 15/10/10/13 steps respectively and zero failures or blocked steps; every retained-secret scan passed. Inspected `run-manifest.json`, `run-summary.json`, `rows/*.json`, `cleanup-results.json` through `make explain-run ... DETAIL=accounting`, and `retained-secret-scan.json` under each exact `.cartulary/test-results/<run-id>` root. No failure, repair or rerun was needed. Q-02 may now begin.

### Q-02 current validation handoff

The requested evidence gap is closed: routing, closed-query admission, owner and shared CSRF/session behavior, generated contract projections, runtime parity and OpenAPI compatibility now have successful fresh evidence on final source `sha256:beb4b641d44ff9f0a994bbc3ce9efb2e525038f2667ee0daf8959419fe1ed2c8`. Combined with the three compatible §14 product rows, the current targeted portfolio contains **ten distinct passing product rows**, plus successful current catalog consistency, finalization and OpenAPI compatibility. The §14 dispatch result remains explicitly identified as a compatible cache hit; every new Q-01 result bypassed the result cache.

Earlier-source outcomes retain their original identities and are not promoted into current-source evidence. The selected portfolio is sufficient for the specifically requested checks; no complete-owner audit, full suite or release acceptance is claimed. Unselected broader checks remain historical. CI remains cancelled and was not restarted. No product, test, authored contract, generated artifact, dependency lockfile or runtime configuration was edited for this follow-up. The existing working-tree changes remain intact; only this tracker/handoff was updated. No commit or deployment was performed, and no additional migration or client compatibility impact was introduced.

The existing `reference-precedence-p03-finalize` evidence remains compatible, so finalization was not repeated merely to update documentation. Retained warm-check maintenance remains skipped because no compatible successful full warm-check root exists; no broad check was launched to create one. Final tracker validation uses `CARTULARY_TEST_RUN_ID=reference-final-source-handoff-markdown make lint-markdown`, followed by a whitespace review and completion checkpoint.

Q-02 completion checkpoint: `reference-final-source-handoff-markdown` returned outer exit 0; `.cartulary/test-results/reference-final-source-handoff-markdown/adhoc/lint-markdown/tool-run-summary.json` reports pass and exit 0. `git diff --check` passed. Final status-only edits preserve the linted Markdown structure. Both workstreams are DONE, with Q-01's checkpoint recorded before Q-02 began. All requested checks now have passing final-source evidence, and the updated tracker/handoff closes this follow-up without restarting CI.

## 16. Legacy removal and production hardening

The user authorized this iteration after reviewing the proposed plan. Its scope is evidenced dead code plus bounded reliability fixes, not new features or persisted-contract redesign. S-09 through S-13 execute sequentially; record each completion checkpoint before starting its successor. The refactor-tracker skill supplies artifact structure; this implementation authorization supersedes its earlier planning-only boundary. No CI restart, full-owner run, reset, migration, or release-readiness claim is part of this iteration.

### Baseline, authority and findings

Baseline I2-00: `main`, HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`, 29 tracked modifications and six untracked files, including this tracker. Tracker starting SHA-256 is `7857fa06cdd404a6e40f063ae2f692aa52f5759bbd191eb9d8fa2388731caa19`. Nonignored Git-enumerated file content hashes and index hash are retained outside the repository in `/tmp/reference-iteration2-1dyy1jx4`; ignored build/test outputs are governed by the harness. Preserve all preceding work. The prior focused planning inspection and current caller rechecks cover the named candidates, not a new exhaustive audit of the historical 243-file module inventory.

The adopted Reference Pack owner governs lifecycle, immutable verification history and fallback; RP-REQ-093/094 govern registry precedence and operational scratch failures; RP-REQ-097/098 govern cancellation and the whole-attempt budget. Extensions EXT-REQ-215 owns inherited deadlines, cancellation grace and proven-commit precedence. These requirements support the fixes without a normative amendment. Tests exercise canonical artifacts and machine expectations, never specification prose. Core HTTP/path-normalization/security contracts remain unchanged.

| Finding | Remediation / areas | Rationale and durable benefit | Compatibility / unresolved risk | Validation |
| --- | --- | --- | --- | --- |
| F-18 / should_fix | Remove unreferenced `SameContent` and exported test-only request wrappers. Move same-package diagnostic composition and inline shape helpers into test files; keep real check builders. Implementation/tests/docs. | Reduce misleading production APIs without duplicating validation rules. | Internal Go callers change together; no wire/data migration. Unresolved helpers create unnecessary maintenance surface. | V-I2-01: caller/build-tag review, affected request/archive/manifest/diagnostics/profile/trust/engine rows. |
| F-19 / should_fix | Remove unused `AdmitReverify`, `FreezeRefresh`, `ApplyVerification`, `Activate`, `CheckCapturedRevisions` and exclusively supporting declarations. Move meaningful assertions to live admission/publication/activation/revision tests. Implementation/tests/catalog/docs. | One lifecycle implementation; regression evidence tests actual execution. | No intended observable change. Leaving duplicate models permits tests to pass while runtime diverges. | V-I2-02: assertion mapping, live canonical persistence/lifecycle and retained fallback coverage, catalog consistency. |
| F-20 / must_fix for S-12 | Replace workspace `Close()` with `Close(context.Context)` throughout owner/adapter/fakes. Seal on cleanup, release leases on every return, retain terminal outcome after relinquishing ownership. Implementation/interface/tests/docs. | Bound cooperative lock waiting and preserve safe recovery without another timeout policy. | Internal signature change only. Current background-context parent-lock waiting can stall workers; incorrect cleanup could race collection or conceal operational failure. | V-I2-03: lock contention, cancellation/deadline, partial failure, repeated close, cross-workspace isolation and recovery; engine fault precedence. |

Intentional/no_action: keep consumer-result validation used by real-output integration evidence, producer fixture support, live retained-content readers, negative compatibility fixtures, effective-set fallback and the current producer reissue path. A test-only caller or older version identifier alone is not dead-code evidence. Public HTTP contracts, successful envelopes, stored formats, trust profiles and recovery admission remain intact.

### Workstreams and binary gates

| Slice | Dependency / work | Principal risk | Exit criterion | Status |
| --- | --- | --- | --- | --- |
| S-09 contract/evidence reconciliation | Current authorization; reconcile owner rules, candidate callers, baseline and live assertion destinations. Correct architectural description. | Treating model tests as normative requirements. | Findings have owners and validation mappings; tracker lint passes; no unresolved owner contradiction. | DONE |
| S-10 helper cleanup | S-09; delete unused helper/wrappers, move diagnostic composition to test-only source using existing production primitives. | A second verifier in test support or missed caller. | Targeted helper/request/format/engine rows pass; no removed API callers. | DONE |
| S-11 lifecycle consolidation | S-10; establish live-path assertions before removal; retain condition/fallback and immutable-copy tests. Update authored catalog and regenerate through Make. | Losing meaningful behavioral evidence. | Assertion map complete; live lifecycle/fallback rows and catalog consistency pass. | DONE |
| S-12 cancellation-aware cleanup | S-11; propagate operation context, keep namespace lock order, observe cancellation between deletions, cache terminal close result, recover residual private scratch through reconciliation. | Lease/collector race, error misclassification or reset cancellation grace. | Targeted storage/engine/lifecycle checks pass, including failure/recovery cases. | DONE |
| S-13 qualification/handoff | S-12; required finalization, review generated changes, freeze source, run exact affected portfolio and document limits. | Historical evidence mislabeled current. | All required targeted obligations pass on compatible final source; docs/whitespace checks and preservation review complete. | DONE |

S-12 uses the existing operation lifetime; no new timeout setting, detached cleanup goroutine, or fresh grace interval. After the first cleanup attempt relinquishes the workspace lease, repeated close returns the recorded result rather than deleting through an unowned handle. Cleanup errors are operational, never content rejection or successful verification. Proven committed success remains authoritative. Residual failed-cleanup scratch stays private and is reclaimed by existing reconciliation; authoritative objects and other live workspaces are excluded.

### Validation and handoff requirements

Discover selections with `make task-guide OWNER=module.reference_data`; execute exact rows with `make test-slice OWNER=module.reference_data ROWS=<recorded IDs>`. V-I2-01 includes canonical archive/manifest/diagnostics/profiles/trust, ATTACK schema-summary consumers, request decoding and canonical consumers. V-I2-02 includes `canonical_pack_state`, `canonical_persistence` and the activation/disable/reverify/refresh integration row. V-I2-03 includes storage containment/cleanup, canonical verification and operational-read classification plus affected live lifecycle tests. New fault cases belong under these existing semantic selectors; retiring model tests requires explicit catalog updates, not hidden lost coverage.

Run `make generate` for authored routing changes and the required generated/schema/catalog checks. Before the final targeted portfolio, run `make agent-finalize`, review its diff and freeze source. Preserve exact native command exits, source digests, selected row results, run roots, cleanup and retained-secret scans. Recompute affected obligations after repairs; never promote old-source passes. Use `make lint-markdown` and `git diff --check` for final documentation. No broad suite is required merely to create a warm-check root. Previous HTTP/security qualification stays historical unless independently compatible or rerun. Roll back each code slice coherently with its tests/catalog; no data conversion or compatibility reader is involved.

### Execution checkpoints

I2-00 / S-09 start: repository instructions, skill guidance, current source, owner clauses and candidate callers rechecked. `make help ALL=1` and `make task-guide OWNER=module.reference_data DETAIL=summary` exited 0. Baseline retained outside the repository. No product evidence claimed; S-09 tracker validation is next.

I2-01 / S-09 complete: `CARTULARY_TEST_RUN_ID=reference-i2-s09-docs make lint-markdown` exited 0; `.cartulary/test-results/reference-i2-s09-docs/adhoc/lint-markdown/tool-run-summary.json` reports pass. F-18–20 map to existing owner requirements; no amendment or compatibility layer is needed. S-10 may begin.

I2-02 / S-10 complete: removed `SameContent`, both exported request wrappers and eight production convenience helpers (now private test composition). Live validators and retained-reader entry points remain. `reference-i2-s10-format` passed. Initial nine-row `reference-i2-s10-targeted` failed at compilation (harness `infra/preflight_error`, outer exit 2): `identity.go` still required `bytes`; the import was restored after inspecting retained logs. `make explain-run ... TARGET=test-slice DETAIL=logs` diagnosed it; earlier diagnostic requests without a public target were rejected and supplied no validation. Exact same nine-row selection in `reference-i2-s10-targeted-fixed`, cache off, passed with outer exit 0; all rows passed and cleanup/retained scan passed. Source digest `sha256:ebaf5c426a669952fddd6600217736583dd9806ddc5e547ee14a93b40516be8e`. Exact ROWS and runtime identity are retained in that run manifest. S-11 may begin.

I2-03 / S-11 complete: removed the five unused lifecycle algorithms, Operation/Verdict enums, unused Staged constant and cohort-only key helper. Condition and effective-set resolution remain; condition projection and nested immutable-copy assertions now exercise those live functions. `reference-i2-s11-format`, `reference-i2-s11-generate`, `reference-i2-s11-catalog` and the three-row `reference-i2-s11-targeted` all exited 0. An assertion-completeness review added initial content/staging failure, disabled missing/recovered reverify and missing usage-guard checks to real persistence tests. `reference-i2-s11-assertions` failed (outer 2, product/test_assertion_failure): the expanded scenario retries a missing candidate, so preparation correctly preserves its prior missing verdict rather than the fresh-candidate staged state assumed by the old assertion. After correcting that test expectation, the same canonical-persistence row passed in `reference-i2-s11-assertions-fixed`, outer 0, cache off. The failed run and diagnosis remain retained; cleanup and scans passed. S-12 may begin.

| Retired model assertions | Live evidence destination / disposition |
| --- | --- |
| Renewal failures preserve history; successful reverify preserves disablement and updates its envelope | Canonical coordinator publication test now compares the entire retained candidate across failed renewal; canonical Jobs recovery checks aborted established operations; local-operator renewal and canonical lifecycle fixtures cover successful renewal. |
| Initial abort/content/missing outcomes; never-successful reverify | Canonical leased-Jobs test covers initial cancellation and explicit no-successful-verification rejection; publication test now executes initial metadata rejection and missing-input staging_loss. |
| Missing reverify retains history/disablement and successful recovery clears missing reason | Added actual missing-member deletion/restoration in canonical publication test; mixed refresh also checks retained envelopes, storage_loss and disabled successful members. |
| Explicit activation clears disablement; condition preserves unhealthy states | Existing real activation/reactivation and canonical lifecycle fixtures; retained live Condition matrix in fallback unit selector. |
| Exact refresh eligibility, pending precedence and frozen inputs | Corrected by F-22/S-17: `TestRefreshAdmissionEligibilityMatrix_Integration` creates available, disabled, failed/missing retained successes, never-successful candidates and removal tombstones before admission; compares exact ordered identities/envelopes, explicit and omitted key resolution, empty no-op cohorts and pending-ineligible-key rejection with no partial writes. `TestCanonicalAdmissionFreezesCohortAndRejectsRelevantStaleState_Integration` inserts a never-successful candidate after admission and proves only frozen membership, not eligibility exclusion. The matrix also checks later successful imports stay outside both frozen cohorts. Old in-memory input-copy assertions are redundant with immutable operation rows and live ResolveEffectiveSet copy assertions. |
| Relevant/unrelated/missing revision guards | Canonical admission guards, concurrent root publication and source-usage tests; added actual missing activation-usage guard. |
| Fallback/cycles and immutable copies | Existing fallback selector retained and expanded; no algorithm relocation or alternate lifecycle validator. |

I2-04 / S-12 complete: context-bearing workspace close now reaches verification, diagnostic scratch, admission, coordinator preparation and historical validation. The rooted adapter seals once, records the final outcome after lease closure, honors cancellation before parent-lock acquisition and between deletions, and relinquishes failed cleanup to reconciliation. Early preparation/result cleanup faults are propagated. The fake implements the same terminal close contract. New existing-selector cases cover contended explicit cancellation/deadline, no reopening/retry after failed close, partial failure and containment, recovery preserving live workspaces, operation-context propagation, and a successful verification whose cleanup fault produces an internal-error Job without a successful prepared member or public content.

`reference-i2-s12-format` passed. `reference-i2-s12-targeted` failed at compile preflight (outer 2, infra/preflight_error) because historical validation and three test call sites still used zero-argument Close. After inspecting its retained logs and updating those sites, identical four-row `reference-i2-s12-targeted-fixed` passed, outer 0, cache off, digest `sha256:c091bd1c6d6286fa2336d063eb914824aad139a07050e307d8104e62a76ba7a7`; all row, cleanup and scan results passed. Historical restoration is exercised by canonical persistence. No change to commit classification, persisted data, public routes, or cancellation-grace settings. S-13 may begin.

### S-13 final targeted qualification and handoff

Final source is `sha256:c091bd1c6d6286fa2336d063eb914824aad139a07050e307d8104e62a76ba7a7`, dirty HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`. `reference-i2-s13-finalize` exited 0, reported generated structure unchanged, and passed source/schema/catalog/tier validation. Its `unit-artifacts/finalize-summary.json` records no updated files. Retained warm-check maintenance was skipped because no compatible successful full warm-check root was supplied; no broad check was launched to manufacture one. The generated render-index delta from S-11 is limited to the Reference Data catalog hash and aggregate input digest.

All following runs report that same final source digest and independently observed outer exit 0. Run roots are `.cartulary/test-results/<run-id>`; each contains `run-manifest.json`, `run-summary.json`, applicable `rows/*.json`, `cleanup-results.json`, and `retained-secret-scan.json`.

| Invocation | Run ID | Result / retained accounting |
| --- | --- | --- |
| `make agent-finalize` | `reference-i2-s13-finalize` | PASS; generated unchanged, 10 cleanup steps completed, scan pass. |
| Exact 14-row `make test-slice OWNER=module.reference_data` below, cache off | `reference-i2-s13-targeted` | PASS; every selected row passed, zero cache hits, 21 cleanup steps completed, scan pass. |
| `make source-check CHECK=generated` | `reference-i2-s13-generated` | PASS; 14 cleanup steps completed, scan pass. |
| `make source-check CHECK=schema` | `reference-i2-s13-schema` | PASS; 12 cleanup steps completed, scan pass. |
| `make lint CHECK=backend-boundary` | `reference-i2-s13-boundary` | PASS; 13 cleanup steps completed, scan pass. |

Exact final product invocation:

```sh
CARTULARY_TEST_RUN_ID=reference-i2-s13-targeted CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_archive_admission,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_diagnostics,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_trust,module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f,module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_state,module.reference_data.integration.canonical_persistence,module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a,module.reference_data.unit.reference_pack_storage_is_atomic_and_contained_14f3d4892a,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2
```

Final review: F-18's unused exports are absent from production; test composition delegates to existing production checks. F-19's five shadow algorithms and exclusively supporting declarations are gone; live projection/fallback remains and the assertion map above reaches production paths. F-20's internal Close signature is migrated through every compiler-reachable adapter, verifier, admission/preparation, historical-reader and test caller; cancellation, terminal ownership, containment, recovery and operational classification have targeted evidence. No compatibility shim, new timeout policy, detached cleanup worker, public-format change or normative amendment was introduced.

Compatibility/handoff: internal VerificationWorkspace implementers must supply `Close(context.Context) error` with terminal recorded outcome and lease release even on failure; all repository implementations are updated. Failed cleanup can leave private scratch for existing reconciliation, and the closed handle cannot retry it. No database conversion, reset, client migration, deployment or producer update is required. Broader producer, frontend/browser, HTTP/security, OpenAPI compatibility, performance, full-owner and CI checks were not repeated because this iteration changes none of their external contracts; earlier runs retain their own source identities. This is bounded production hardening, not full production/release certification.

Preservation comparison uses `/tmp/reference-iteration2-1dyy1jx4/baseline.json`: 47 intended changed/added paths, including one new test-helper file, and no deleted baseline files. Original unrelated content is unchanged. The Git index byte hash differs from the initial hash, so byte-for-byte index preservation is not claimed; the staged tree remains unchanged and `git diff --cached --exit-code` passes. No staged changes or commit were made. Final handoff lint and whitespace checks follow; source remains frozen while only this tracker is updated.

I2-05 / S-13 complete: fourteen of fourteen selected product rows passed with cache off on the final source above; all owning commands exited 0. Source/generated/schema/boundary/finalization checks passed with no cleanup failures, blocked cleanup or retained-scan findings. `CARTULARY_TEST_RUN_ID=reference-i2-s13-handoff-docs make lint-markdown` exited 0; its `adhoc/lint-markdown/tool-run-summary.json` reports pass. `git diff --check` and `git diff --cached --exit-code` passed. These final status-only edits preserve the linted Markdown structure. F-18–20 are resolved, S-09–13 are DONE, and every successor began after its predecessor checkpoint. No remaining implementation blocker was identified within this bounded iteration; the stated broader audit/release limits remain. CI was not restarted. No commit or deployment was performed.

## 17. Cancellation cleanup and live refresh admission

The user explicitly authorized implementation of S-14–18 on 2026-11-07. This section supersedes §16's claim that F-19/F-20 have no remaining gaps. Historical runs remain tied to their original source. Baseline I3-00: dirty `main`, HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`; nonignored file content/mode and index fingerprints are retained in `/tmp/reference-i3-ihtg3ccs/baseline.json`. Preserve all preceding work. No CI restart or full-owner run is authorized or required.

| Finding | Remediation / areas | Rationale / durable benefit | Compatibility / unresolved risk | Validation |
| --- | --- | --- | --- | --- |
| F-21 / must_fix, resolved S-18; follows F-20 | Supervised operation cleanup under the original Extensions cancellation grace, plus automatic serving reclamation. Specification, implementation, assembly, tests, documentation. | Reference Data owns completion, Extensions owns lifetime enforcement, storage owns confined deletion. Prevent normal cancellation from accumulating private bytes without duplicating timeout policy. | Internal lifetime dependencies only; no data/client migration. Leaving it open risks disk exhaustion and violates RP-REQ-094/098/238/239. | V-21 shared lifetime; V-22 real operation cleanup; V-23 automatic recovery; V-24 isolation/shutdown; V-26 fatal process consequence. |
| F-22 / should_fix, resolved S-18; follows F-19 | Seed eligibility states before real refresh admission; assert exact ordered identities, envelope bindings and pending precedence. Tests/documentation. | Protect the actual SQL/lifecycle instead of a shadow model. | No intended product change. Missing evidence permits cohort regressions and overstates closure. | V-25 live admission matrix under RP-REQ-187/188/191. |

| Slice | Dependency | Work / risk | Binary exit criterion | Status |
| --- | --- | --- | --- | --- |
| S-14 | Review | Reconcile owner clauses and adoption; reopen current claims; preserve baseline. Risk: weakening cleanup requirements. | Owner wording, requirements and validation map reconciled; Markdown lint passes. | DONE |
| S-15 | S-14 | Shared supervised lifetime for admission, verification, diagnostics, preparation, refresh, portability and historical readers; configure grace/fatal handling. Risk: renewed grace, masked fatal outcome or detached work. | Targeted lifetime/storage/verifier/composed operation tests pass; all close paths accounted for. | DONE |
| S-16 | S-15 | Add narrow scratch reconciliation capability to existing serving collector, scratch first under the same sweep budget. Risk: live collection, namespace crossing, overlapping sweeps. | Automatic recovery, containment, retry and shutdown tests pass. | DONE |
| S-17 | S-16 | Live pre-admission cohort matrix and corrected retired assertion mapping. Risk: count-only checks or bypassing real lifecycle. | Exact membership/order, pending rejection atomicity and frozen cohort tests pass. | DONE |
| S-18 | S-17 | Finalize, freeze source, targeted portfolio, specification/diff review and handoff. Risk: promoting stale passes. | Every required obligation has compatible passing evidence; final tracker and handoff complete. | DONE |

Update this tracker after each slice and before beginning its successor. Preserve terminal workspace close semantics, lease ordering, safe diagnostics, commit precedence and original timeout settings. No compatibility aliases, schema conversion, detached cleanup worker or new public readiness state. Reconciliation remains active without the optional administration claim.

Validation: V-21 includes cancellation before/during cleanup, earlier parent deadline, owner timeout, shared grace across resources/cohort members, zero grace and fatal late results. V-22 uses rooted storage for success, content rejection, cancellation, timeout and preparation failure, including probing and historical validation. V-23 observes the actual serving loop after a real cleanup fault without a manual sweep. V-24 covers live cross-instance leases, creation/collection ordering, unsafe entries, namespace preservation, terminal close and joined shutdown. V-25 covers available/disabled/failed/missing retained successes, never-successful candidates, tombstones, Base, empty/omitted selections, exact ordering/envelopes and pending explicitly selected ineligible keys. V-26 proves the real server/operator fatal consequence with exit 70 and secret-safe diagnostics.

Use public Make discovery and exact cache-disabled selections. Retain S-13's fourteen Reference Data rows and add local operator plus narrowly cataloged cleanup, Extensions lifetime and server/operator consequence coverage. Regenerate authored catalog changes through Make. Run agent-finalize before final qualification, then generated/schema/boundary/Markdown/whitespace checks. Record native exits, source identity, rows, cleanup and retained scans. No test reads specification prose. S-18 alone closes these findings.

I3-00: baseline captured and tracker reopened before production changes. S-14 owner clarification and discovery are in progress.

I3-01 / S-14 complete: owner clarification adopted under the user authorization; RP-REQ-238/239 explicitly connect existing cleanup obligations, original Extensions grace, operator fatal consequence and serving collection. Core 00 records the clarification. `reference-i3-s14-docs` / `make lint-markdown` exited 0; retained tool summary reports pass. No product source changed. S-15 begins next.


I3-02 / S-15 complete: `VerificationLifetime` imports Extensions completion and explicitly requires configured grace plus a fatal sink. Inherited tracking spans extraction, diagnostics, nested preparation and refresh members, portable preparation, import probing and historical readers. Workspace deletion detaches cancellation only inside the joined completion; storage retains terminal close caching and releases its lease after deletion finishes or fails. Server, local operator and recovery constructors supply the policy. A fatal cleanup error bypasses ordinary Job failure finalization. No compatibility constructor remains.

S-15 evidence uses `CARTULARY_TEST_RUN_ID=<run>` and `CARTULARY_HARNESS_CACHE_MODE=off` with the following exact `make test-slice` selections. Each listed run root is `.cartulary/test-results/<run>`; native outer exits were 0, cleanup outcomes completed and retained-secret scans passed. These are development checkpoints; S-18 requalifies the final source.

| Run | Owner / exact rows | Source digest |
| --- | --- | --- |
| `reference-i3-s15-qualified-fixed` | `module.reference_data`: `module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.unit.reference_pack_storage_is_atomic_and_contained_14f3d4892a,module.reference_data.integration.canonical_persistence,module.reference_data.integration.workspace_cleanup_lifecycle` | `sha256:a093d2edf4627f71a04283f7f45778e28d51955aa9aa4978121de3ea5fa9347b` |
| `reference-i3-s15-lifetime-final` | `module.reference_data`: the two unit rows above, after adding shared-grace and cooperative lease-wait cases | `sha256:ca65dd4a505d440c4322a0eb9a8f5d10c698ccdc774de80f38d2a43407c72c2b` |
| `reference-i3-s15-extensions` | `module.extensions`: `module.extensions.unit.admission_invocation_v6` | `sha256:ffa3b35b817e381361e8293033d40136f571f09263519f94e9151fbba11f385b` |
| `reference-i3-s15-server` | `app.server`: `app.server.process.cancellation_grace_terminates_blocked_startup_se_b4b985d465` | same `ffa3b35b…385b` |
| `reference-i3-s15-operator` | `app.operator`: `app.operator.process.reference_cleanup_grace,app.operator.integration.reference_pack_retained_admission,app.operator.unit.operator_registry_eight_paths_and_collaboration_01217009b8` | same `ffa3b35b…385b` |

V-21/22/26 evidence now includes real 60-second owner timeout, cancellation at private indexing, content rejection, partial preparation failure, import probing and historical interruption with rooted scratch removed before completion; cancellation during a directory-lease wait; shared nested grace; zero grace; late callback fatal precedence; and server/operator child processes exiting 70 with the registered safe diagnostic when real rooted cleanup cannot acquire its lease. Tests do not manually reconcile to establish cooperative cleanup.

Retained failed development invocations (all outer exit 2): `reference-i3-s15-compile` and `-compile-fixed` were preflight compilation failures from migrated test signatures; repaired all call sites. `reference-i3-s15-format2` and `-generate` rejected unsorted/stale catalog input; sorted authored rows and regenerated. `reference-i3-s15-targeted` exposed two direct test probe calls without the required lifetime; corrected their composition. `reference-i3-s15-qualified` exposed a duplicated timeout assertion in the historical test; removed it. Their run summaries/logs remain retained. `reference-i3-s15-targeted-fixed` passed all four rows before the final timeout/fatal additions; it is superseded by the evidence above. `make generate` (`reference-i3-s15-generate-fixed`, `-generate2`) and `make format` (latest `reference-i3-s15-format5`) exited 0. S-16 begins next; F-21 remains open pending serving recovery and final qualification.


I3-03 / S-16 complete: `ObjectCollector` now requires a separate `WorkspaceReconciler` capability. It reconciles scratch first, then collects published bytes while the same sweep context remains live, joining both errors. Server composition supplies rooted storage for each capability; startup/operator recovery remains. The serving gate, one-minute interval, 30-second total budget, serial execution, safe `reference_pack.collection` observation and joined close remain. A cancellation check prevents a queued tick from starting another sweep during shutdown. RP-REQ-239's introductory sentence now explicitly uses leases, removing the old Job-liveness ambiguity.

`reference-i3-s16-generate` / `make generate` and `reference-i3-s16-format` / `make format` exited 0. Cache-disabled `reference-i3-s16-targeted` selected `module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.unit.reference_pack_storage_is_atomic_and_contained_14f3d4892a,module.reference_data.integration.workspace_cleanup_lifecycle` via `make test-slice OWNER=module.reference_data ROWS=...`. Both unit rows passed; integration failed (outer 2, product/test_assertion_failure) because the new test expected an OS error from the deliberately safe storage error boundary. The physical-deletion assertion was corrected without changing production behavior. `reference-i3-s16-recovery-fixed` reran the affected `module.reference_data.integration.workspace_cleanup_lifecycle` row, cache off, outer 0; source `sha256:2c99458a7bf6bc690d0d7d98f69840ba536bc76326f9b75a401af3442ba20748`. The unit checkpoint source was `sha256:b3c9cda177ecc53eec65fa0f62d332e21ad15905b778705513590913fe38ace3`. Run roots follow `.cartulary/test-results/<run>`; all cleanup completed and retained scans passed, including the preserved failed run.

V-23 evidence uses the real collector `Prepare` serving loop and rooted deletion: a file-type fault makes terminal close fail; one maintenance tick preserves unsafe residue while collecting an unrelated published orphan; repair plus the next tick removes residue. The test never calls `Sweep` or `ReconcileWorkspaces`. A test-only constructor supplies tick delivery, asserting the unchanged production cadence. Separate live storage instances, concurrent creation/collection, retained Base validation, frozen-input preservation and cached close outcomes pass. Unit tests establish phase error aggregation, common budget, serving gate, no overlap and shutdown joining. S-17 begins next.


I3-04 / S-17 in progress, S-15 dependency repair: the new pre-admission matrix's held configuration read exposed a cancellation-context omission in execution. That read still used `budgetCtx`, so cooperative Job cancellation could not release it through `preparationCtx`; supervision then correctly exhausted grace. `reference-i3-s17-targeted` failed (outer 2, product/test_assertion_failure, lifecycle row); canonical persistence passed. The read now receives `preparationCtx`. The barrier remains held until cancellation, so the regression case must demonstrate real propagation rather than an artificial release. Admission SQL is unchanged. The affected S-15/S-17 evidence will be rerun before this workstream closes.


I3-05 / S-17 complete: the live matrix creates all eligibility states through canonical imports, disable/remove actions and physical retained-byte faults followed by real reverify. It checks exact `(pack_key, pack_version, envelope_id)` membership in ordinal order (`10` before `2`), Base/never-successful/tombstone exclusion, explicitly selected known empty keys, omitted-key resolution, deterministic empty no-op jobs, pending precedence on an ineligible key alone and mixed with an eligible key, rejected-admission atomicity across Jobs/receipts/operation tables, and subsequent real imports remaining outside frozen cohorts. The retired-assertion map above is corrected. No admission SQL or normative eligibility rule changed.

`reference-i3-s17-generate` / `make generate` and `reference-i3-s17-format` / `make format` exited 0. After the configuration-read cancellation repair recorded in I3-04, `reference-i3-s17-cancellation-fixed` ran cache-disabled `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a,module.reference_data.integration.workspace_cleanup_lifecycle`; all three rows passed, native outer exit 0, source `sha256:638e7409d9b1c3d9ce63ac72daf89d464f3ac5dbc23907e084b09574866107a0`. All cleanup completed and retained-secret scan passed. Exact run root: `.cartulary/test-results/reference-i3-s17-cancellation-fixed`. The earlier failed run remains retained. S-15's affected composed-operation evidence is refreshed by this run. S-18 begins next; findings remain open until final source qualification and review.


I3-06 / S-18 qualification repair: final review removed the private `verifiedContent.Close` context argument because the resource now closes under its bound operation lifetime; it also made admission cleanup preserve fatal errors consistently. The first finalization passed with no generated changes (`reference-i3-s18-finalize`, outer 0). First qualification on `sha256:6998869b27838ca71de8fbfcb64e0630fd30e45147232f18529af87229fee0a8` retained passing Extensions/server/operator and generated/schema checks, but `reference-i3-s18-targeted` failed compile preflight (outer 2) on two remaining private-close test calls, and `reference-i3-s18-boundary` failed (outer 2) on five newly required exact owner-port importer paths. The calls and authored boundary allowlist were corrected; no wildcard widening or boundary-check suppression was added. These first qualification roots remain historical. Re-finalize and repeat the targeted portfolio on the repaired final digest; do not combine old-source passes into closure.


I3-07 / S-18 qualification repair: the complete portfolio on `44398fd6…2895` exposed `TestRetainedContentChecksFollowVerificationPrecedence_Unit/cleanup_inherits_cancellation_and_operation_values` in the canonical-consumers row. Index construction replaced semantic cancellation with a generic operational error; detached cooperative cleanup no longer happened to overwrite it. `reference-i3-s18-targeted-fixed` exited 2 (product/test_assertion_failure); independent work drained normally. Preserve the original interruption directly at the index-construction boundary, while retaining generic operational treatment for actual index faults. Keep the assertion and requalify; do not weaken cancellation expectations. Other rows and all companion checks remain historical until the repaired digest is qualified.


### S-18 final qualification and handoff

Final source: `sha256:5b40ffb972a84635f529b4d55d5734decf11afe1d1e399d4f843e64647819948`, dirty `main` at HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`. The targeted cancellation repair passed in `reference-i3-s18-cancellation-unit` (canonical-consumers and verifier rows, cache disabled, outer 0). `reference-i3-s18-finalize-qualified` then passed with no generated changes. Source was frozen after finalization and its non-document inputs were rechecked against `/tmp/reference-i3-ihtg3ccs/final-source.json`; no product or tooling source changed during qualification. Documentation does not participate in executable source/evidence identity.

Every final invocation below has an independently observed native outer exit **0**, a passing owning summary, completed cleanup with zero failures, and a passing retained-secret scan. All manifests have the final digest above. Product test invocations used `CARTULARY_HARNESS_CACHE_MODE=off`; zero cache hits. Run roots are `.cartulary/test-results/<run-id>` with `run-manifest.json`, `run-summary.json`, selected `rows/*.json`, `cleanup-results.json` and `retained-secret-scan.json`. Commands set `CARTULARY_TEST_RUN_ID=<run-id>`.

| Run ID | Exact Make command / selection | Result and cleanup |
| --- | --- | --- |
| `reference-i3-s18-finalize-qualified` | `make agent-finalize` | PASS; generated unchanged; 10 cleanup steps |
| `reference-i3-s18-targeted-qualified` | `make test-slice OWNER=module.reference_data ROWS=<sixteen rows below, comma-separated>` | 16/16 rows PASS; 21 cleanup steps |
| `reference-i3-s18-extensions-qualified` | `make test-slice OWNER=module.extensions ROWS=module.extensions.unit.admission_invocation_v6` | 1/1 PASS; 10 cleanup steps |
| `reference-i3-s18-server-qualified` | `make test-slice OWNER=app.server ROWS=app.server.process.cancellation_grace_terminates_blocked_startup_se_b4b985d465` | 1/1 PASS; 10 cleanup steps |
| `reference-i3-s18-operator-qualified` | `make test-slice OWNER=app.operator ROWS=app.operator.process.reference_cleanup_grace,app.operator.integration.reference_pack_retained_admission,app.operator.unit.operator_registry_eight_paths_and_collaboration_01217009b8` | 3/3 PASS; 16 cleanup steps |
| `reference-i3-s18-generated-qualified` | `make source-check CHECK=generated` | PASS; 14 cleanup steps |
| `reference-i3-s18-schema-qualified` | `make source-check CHECK=schema` | PASS; 12 cleanup steps |
| `reference-i3-s18-boundary-qualified` | `make lint CHECK=backend-boundary` | PASS; 13 cleanup steps |

Exact final Reference Data rows (the original fourteen plus local-operator and cleanup-lifecycle coverage):

```text
module.reference_data.integration.canonical_persistence
module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a
module.reference_data.integration.reference_pack_import_uses_the_shared_upload_env_95574f69b8
module.reference_data.integration.workspace_cleanup_lifecycle
module.reference_data.unit.canonical_consumers
module.reference_data.unit.canonical_fixture_manifests
module.reference_data.unit.canonical_pack_archive_admission
module.reference_data.unit.canonical_pack_diagnostics
module.reference_data.unit.canonical_pack_manifest
module.reference_data.unit.canonical_pack_profiles
module.reference_data.unit.canonical_pack_state
module.reference_data.unit.canonical_pack_trust
module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f
module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda
module.reference_data.unit.reference_pack_storage_is_atomic_and_contained_14f3d4892a
module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2
```

No earlier-source results are used for this closure. Earlier qualification failures and their diagnoses remain in I3-06/I3-07 and their exact retained roots. In particular, `reference-i3-s18-targeted-fixed` passed 15 rows but failed canonical-consumers; its cleanup and retained scan passed. The final fresh portfolio replaces that incomplete evidence. `make task-guide OWNER=module.reference_data`, `OWNER=module.extensions`, `OWNER=app.server` and `OWNER=app.operator` supplied owner routing, each observed with exit 0; the exact authored rows above narrow those owners. No full-owner suite or CI ran.

| Obligation / owner clauses | Final evidence and review conclusion |
| --- | --- |
| V-21 / EXT-REQ-215; RP-REQ-098/238 | `TestVerificationLifetimeCleanup_Unit`, Extensions admission invocation tests and `TestImportProbeSupervisedRootedCleanup_Unit` cover prior cancellation, inherited deadlines, shared resource/nested-member tracking, zero grace, lease waiting, fatal late completion and retained context values. `TestReferenceWorkspaceOperationCleanup_Integration` exercises the real 60-second owner timeout. Review confirms one tracked outer refresh scope and inherited tracking in every member; the mixed-refresh lifecycle remains covered. |
| V-22 / RP-REQ-094/098/238 | Rooted operation, import-probe and historical-validation cases cover success, content rejection, cancellation, timeout and partial preparation failure. Extraction/diagnostic scratch is absent before cooperative completion without manual reconciliation. Canonical coordinator cleanup-failure cases prove no successful preparation/content verdict; fatal failure cannot become ordinary Job finalization. The cancellation boundary remains semantic even when cleanup succeeds. |
| V-23 / RP-REQ-200/239 | `TestReferenceWorkspaceAutomaticRecovery_Integration` drives the actual `Prepare` loop through test-only ticks: real close failure, retained unsafe residue, independent published-object collection, dependency repair and successful automatic scratch reclamation. It calls neither `Sweep` nor `ReconcileWorkspaces`. Maintenance works with the administration claim absent. Production remains one minute / 30 seconds. |
| V-24 / RP-REQ-238/239 | Rooted storage isolation/cleanup tests cover cross-instance live leases, unsafe entries/symlinks, terminal close and retained namespaces. Automatic recovery adds concurrent workspace creation. `TestObjectCollectorLifecycle_Unit` covers phase order, shared budget, aggregated errors, serving gate, serial execution, cancellation of queued ticks and joined shutdown. |
| V-25 / RP-REQ-187/188/191 | `TestRefreshAdmissionEligibilityMatrix_Integration` passes the complete pre-admission state matrix, exact lexical membership/envelopes, explicit/omitted keys, empty no-op jobs, pending-ineligible-key precedence and rejection atomicity. Real later imports leave frozen cohorts unchanged. The older post-admission candidate test remains specifically immutability evidence; the retired assertion map is corrected above. |
| V-26 / EXT-REQ-215; Core 04 REQ-04-144/146 | The server cancellation-grace process selector's `reference_cleanup` case and `TestReferenceCleanupGrace_Process` block real rooted deletion. Actual server/operator process sinks terminate child processes with exit 70 and the exact registered safe diagnostic. Mock sinks are used only for deterministic joined-lifetime unit assertions. |

Separate final diff/evidence review checked these owner clauses against lifetime acquisition and reuse, all production workspace-close callers, fatal precedence and commit handling, server/operator/recovery composition, lease confinement and terminal caching, collector ordering/gating/shutdown, admission SQL and fixture lifecycle transitions. No remaining scoped implementation/specification blocker was found. Direct cross-references now connect RP-REQ-094/098 to the adopted RP-REQ-238/239 clarification; Core 00 already records all four clauses. Tests consume canonical artifacts and machine expectations; none acquires specification/adoption prose. This is a scoped review and targeted qualification, not complete subsystem, complete-owner or release certification.

Preservation: relative to I3-00, 54 existing paths changed and 11 files were added; none was removed. The six module additions are `verification_lifetime.go`, `verification_lifetime_test.go`, `object_collector_test.go`, `workspace_lifetime_integration_test.go`, `workspace_recovery_integration_test.go`, and `refresh_admission_matrix_integration_test.go`. Other additions provide operator lifetime/process composition, real-storage admission tests and shared test fixtures. Constructor call-site migrations account for changes in recovery/reporting tests and application test composition. Authored catalog inputs and five exact boundary import paths were updated; the generated render-index delta reflects the two owner catalogs. Previous unrelated dirty-tree work is preserved. No schema, pack-format, SQL admission, public HTTP API, dependency-lockfile, UI, source-producer or compatibility-reader change was introduced. The index's stat fingerprint can change during Git inspection; staged content remains unchanged.

Deployment and rollback: startup recovers existing abandoned scratch. Ordinary cancellation/timeout completes cleanup synchronously within the original remaining grace. Genuine storage cleanup failures stay private and are retried while serving; a persistent storage fault or unsafe inventory still requires repair before a sweep can succeed. Noncooperative teardown follows fatal exit 70. No reset, data conversion or new timeout setting is required. Roll back the lifetime/constructor callers, collector capability, tests, catalog and boundary entries as coherent units. Do not silently reintroduce indefinite cleanup or disable serving recovery.

Finalizer retained warm-check/evidence-maintenance actions were skipped because no compatible successful full warm-check root was supplied. No broad run was created to satisfy that optional maintenance path. Current-source targeted summaries, rows, cleanup and scans were inspected directly. Historical HTTP/security evidence stays historical; the final scope is the portfolio above. Final Markdown validation and the S-18 completion checkpoint follow.


Final Reference Data reproduction command (normalized row order from its retained manifest):

```sh
CARTULARY_TEST_RUN_ID=reference-i3-s18-targeted-qualified CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a,module.reference_data.integration.reference_pack_import_uses_the_shared_upload_env_95574f69b8,module.reference_data.integration.workspace_cleanup_lifecycle,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_archive_admission,module.reference_data.unit.canonical_pack_diagnostics,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_state,module.reference_data.unit.canonical_pack_trust,module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f,module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda,module.reference_data.unit.reference_pack_storage_is_atomic_and_contained_14f3d4892a,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2
```

Use a fresh run ID for any later execution; the retained ID above must not be reused.

I3-08 / S-18 complete: all V-21–26 obligations have the passing final-source evidence above. `CARTULARY_TEST_RUN_ID=reference-i3-s18-handoff make lint-markdown` exited 0; `adhoc/lint-markdown/tool-run-summary.json` reports pass. `git diff --check` and `git diff --cached --exit-code` each exited 0. Final source comparison found no added, removed or modified non-document input after freeze. F-21 and F-22 are resolved; S-14–18 are DONE. The final status/command transcription is checked once more by `reference-i3-s18-handoff-final` / `make lint-markdown`. No scoped blocker remains. CI was not restarted, and no full-owner suite, deployment, reset or data conversion was performed.

## 18. Portable execution lifetime and fatal precedence

The user authorized S-19–23 implementation on 2026-11-07. F-21 is reopened for the portable paths below; F-22 stays closed. S-18's 21 passing rows on `sha256:5b40ffb972a84635f529b4d55d5734decf11afe1d1e399d4f843e64647819948` remain historical evidence, not qualification of this remediation. This section supersedes its complete F-21/S-18 closure claim without erasing prior execution history.

I4-00 baseline: dirty `main`, HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`; content/mode and index fingerprints are retained in `/tmp/reference-i4-obq_3abd/baseline.json`. Preserve preceding changes. The tracker reconciliation is the first repository mutation. Each slice is a separate sequential workstream; update this artifact after completing it and before starting its successor. Targeted validation only; no CI or full-owner suite.

| Finding | Remediation / areas | Rationale and long-term benefit | Compatibility / unresolved risk | Validation |
| --- | --- | --- | --- | --- |
| F-23 / must_fix, resolved in S-23; child of F-21 | One callback-owned portable execution, joined resource completion and inherited cancellation tracking. Implementation, internal interfaces, tests, owner clarification and documentation. | Explicit ownership prevents unsupervised defers and renewed grace across containers, source preparation and parent completion. | Internal call-site migration only; no persisted/public contract change. Unresolved cleanup can block beyond grace. | V-27 shared lifetime; V-28 rooted cleanup; V-31 process consequence; V-32 compatibility. |
| F-24 / must_fix, resolved in S-23; child of F-21 | Preserve fatal errors through cleanup and classification; exclude ordinary finalization after exhaustion. Implementation, internal interfaces, tests and documentation. | Fatal policy no longer depends on process exit winning a goroutine race. | No data/client migration. Unresolved late results can publish misleading terminal state and defeat recovery. | V-29 fatal precedence; V-30 parent mutation exclusion; V-31 process consequence. |

| Slice | Depends on | Work / principal risk | Binary exit criterion | Status |
| --- | --- | --- | --- | --- |
| S-19 | Review | Reconcile tracker, owner clauses and scope. Risk: weakening adopted requirements. | Requirements, interface and validation map recorded; Markdown passes. | DONE |
| S-20 | S-19 | Callback-owned portable lifetime and supervised cleanup; migrate callers. Risk: grace reset or escaping teardown. | Focused lifetime and portable success/replay tests pass; resources have explicit completion ownership. | DONE |
| S-21 | S-20 | Fatal-preserving cleanup/classification and guarded parent finalization. Risk: ordinary terminal writes or deletion after fatal. | Late fatal results cannot become ordinary finalization; cooperative contracts pass. | DONE |
| S-22 | S-21 | Rooted portable, live persistence and actual fatal-process evidence; authored catalog updates. Risk: helper-only tests. | Production paths prove cleanup, retained recovery inputs and exit 70. | DONE |
| S-23 | S-22 | Finalize, freeze, targeted qualification, independent review and handoff. Risk: stale evidence or overclaimed closure. | Every scoped obligation has compatible final passing evidence and reviewed handoff. | DONE |

Authority: RP-REQ-094/098/228/238/239 already require bounded cleanup and one portable execution through parent commit classification; EXT-REQ-215 owns grace and fatal lifetime precedence; Core 04 REQ-04-144/146 imports process termination. RP-REQ-228's clarification links these existing rules without changing deadlines, public outcomes or recovery. Reference Data owns verification and its resources, Incident Bundles supplies the orchestration callback, Extensions enforces lifetime, and storage retains confined deletion/lease policy.

Replace manual `BeginImportExecution`/`Context`/`Close` with `RunImportExecution(ctx, execution, started, callback) error`. Preserve the parent start instant; activate the verification budget only for fresh work. Supervise complete Reference Data preparation and parent orchestration/teardown; nested calls inherit tracking. A scope-bound cleanup method detaches cancellation only inside joined completion. Remove the narrow verifier wrapper and superseded caller paths. References-only reuse creates no fresh cohort. Add a read-only Extensions exhaustion query and explicit fatal classification error; no ordinary abort, terminal Job mutation or input disposal may follow fatal selection. Preserve committed/indeterminate bytes.

| Obligation | Required evidence |
| --- | --- |
| V-27 | Prior/during cancellation, inherited/owner deadlines, zero grace, shared original grace across containers/resources/later stages, normal teardown without manufactured interruption. |
| V-28 | Rooted portable success, optional/required rejection, cancellation, timeout, partial failure; source/scratch/object/lease completion before return; retained authoritative inputs protected. |
| V-29 | Fatal survives late success, close/disposal faults and observed cancellation/timeout; no successful preparation or content verdict. |
| V-30 | Late callback after fatal cannot produce parent/owner terminal writes, catalog/pins/success or ordinary finalizer calls; admitted recovery inputs remain. |
| V-31 | Production portable boundary plus actual server fatal sink exits 70 with registered safe diagnostic for blocked source closure and rooted cleanup. Preserve existing operator consequence coverage. |
| V-32 | References-only/redundant inputs, optional/required semantics, frozen retry, committed replay, proven commit/cancellation, lost lease, indeterminate publication, ordinary cleanup failure and serving recovery. |

Validation procedure: discover exact rows using `make task-guide OWNER=...`; retain S-18's 21-row final portfolio plus Incident Bundles worker-transition, publication recheck, portability matrix and reference round-trip rows, and narrowly cataloged new lifecycle/process coverage. Update authored inputs and `make generate`. Before final qualification run `make agent-finalize`, inspect changes and freeze source; run exact `make test-slice` rows with fresh IDs and cache off. Complete generated/schema source checks, backend-boundary lint, Markdown lint and `git diff --check`. Tests consume canonical artifacts and machine expectations, never specification prose. Preserve failed runs/diagnoses; repairs invalidate affected evidence.

Handoff records exact commands/native exits, rows, source digests, roots, cleanup/secret scans and independent diff review. No reset/conversion/configuration change is planned. Roll back interfaces/callers, error policy, tests and catalog entries coherently; no unsupervised fallback. Only S-23 may close F-21/F-23/F-24.

I4-01 / S-19 complete: tracker/owner reconciliation precedes code changes. `CARTULARY_TEST_RUN_ID=reference-i4-s19-docs make lint-markdown` exited 0, retained tool summary passed. `make task-guide OWNER=module.reference_data`, `module.incidentbundles`, `module.extensions`, `app.server`, and `app.operator` each exited 0. Existing normative RP-REQ-228 now explicitly links resource closure and fatal precedence to RP-REQ-238/EXT-REQ-215. S-20 begins next.

I4-02 / S-20 complete: callback-owned `RunImportExecution` replaces manual construction/context/close; entire owner preparation and parent orchestration/teardown are supervised. Fresh work requires its active matching scope. Nested completion inherits tracking; normal final cancellation follows completion classification. Scope cleanup preserves context values and refuses use after return. Parent-owned prepared resources, portability resources, member scope and archive now finish before ordinary failure classification. The narrower verifier wrapper is removed. Fatal downgrade guards remain the explicitly open S-21 obligation.

`reference-i4-s20-characterization` passed all four Incident Bundles rows before mutation, source `5b40ffb9…9948`, outer exit 0. `reference-i4-s20-format` failed (outer 2): migrated test callback had a missing newline before `return`; retained formatter stderr identified line 380. After correction, `reference-i4-s20-format-fixed` passed, outer 0. `reference-i4-s20-reference` passed exact rows `module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.integration.canonical_persistence`; `reference-i4-s20-incident` passed exact rows `module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e,module.incidentbundles.integration.incident_bundle_import_final_publication_recheck_b7b07b3e93,module.incidentbundles.integration.extension_portability_matrix_and_atomic_import_1b27fc2d91,module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8`. Both commands used `CARTULARY_TEST_RUN_ID=<root name> CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=<matching owner> ROWS=<above>`; native exits 0, cache hits 0, source `sha256:b80dbeb417125afb77c7d965fd4e529b89f648b315c63c8c9c58f8a079f1d4f5`. Roots are `.cartulary/test-results/<run>`; cleanup and retained scans completed successfully. S-21 begins next.

I4-03 / S-21 complete: Extensions exposes a read-only inherited exhaustion query without observation/reset. Portable classification now returns an explicit fatal error before ordinary classification; abort refuses exhausted/inactive/unclassified execution. Cleanup preserves fatal errors across source close, private disposal and lease release. Parent teardown preserves fatal causes, skips ordinary finalization and retains frozen input on fatal exit. Deterministic late-result cases assert independently of the outer fatal return, including nil transaction/finalizer dependencies proving no ordinary mutation call. Live parent-persistence and process evidence remain S-22 obligations.

`reference-i4-s21-format` / `make format` passed, outer 0. Cache-disabled `make test-slice` runs `reference-i4-s21-reference` (same two Reference Data rows as S-20), `reference-i4-s21-incident` (worker-transition row from S-20), and `reference-i4-s21-extensions` (`module.extensions.unit.admission_invocation_v6`) all passed, native exits 0. Exact selections are retained in each manifest under `.cartulary/test-results/<run>`; all cleanup outcomes completed and retained-secret scans passed. These are development evidence, to be replaced by final qualification. S-22 begins next.


I4-04 / S-22 complete: rooted portable success, optional/required rejection, cancellation, parent/owner timeout, partial write failure and source-close failure now run through the public owner boundary. A two-container test blocks real unpublished-object disposal, releases it after fatal selection, completes both disposals and proves the next parent completion cannot renew grace. The live parent test enters the production Incident Bundles failure path with no ordinary mutation capabilities after fatal selection; the Job stays running, catalog/pin/version/event/owner-terminal counts stay unchanged and frozen input objects remain available. Server children block source closure or rooted workspace deletion through public portable preparation and the real fatal sink: both exit 70 with the exact registered safe diagnostic. Test-only composition uses canonical fixtures and actual persistence/storage; no environment-controlled production faults or manual reconciliation establish cleanup.

The separate S-22 path review found and repaired observer shutdown ordering: the verification timer now stays active until the observer joins. Ordinary parent failure publication also uses supervised completion, retaining the original tracker through its local timeout. Focused regression cases cover blocked observer shutdown, inherited exhausted grace, subsequent source/parent stages and normal zero-grace teardown. These repairs invalidate the earlier development snapshots for final qualification.

`reference-i4-s22-generate` / `make generate`, `reference-i4-s22-format` / `make format`, and `reference-i4-s22-review-format` / `make format` all exited 0. Authored catalogs add exactly three rows: `module.reference_data.integration.portable_completion_lifecycle`, `module.incidentbundles.integration.reference_pack_fatal_completion`, and `app.server.process.portable_reference_cleanup_grace`; generated render-index inputs were refreshed through Make. Rediscovery through `make task-guide OWNER=module.reference_data`, `module.incidentbundles`, and `app.server` exited 0.

Initial development runs `reference-i4-s22-reference`, `reference-i4-s22-incident`, and `reference-i4-s22-server` reported pass on `sha256:c56e031cf5a172352ea67b5bb826a9196f3f00fc69a7f042162dbdccc59fec03`; cleanup and scans passed, but their native outer exits were not captured. They are superseded development evidence, not qualification. Following the review repairs, these cache-disabled commands each passed with independently observed native outer exit 0 on `sha256:bf9cb8db06a9c48acd5750e9c1ec62eaf7c23559001f99318ca2f231f956a08e`:

| Run ID | Exact `make test-slice` arguments | Rows / cleanup |
| --- | --- | --- |
| `reference-i4-s22-reviewed-reference` | `OWNER=module.reference_data ROWS=module.reference_data.integration.portable_completion_lifecycle,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2` | 2 PASS; 15 completed cleanup steps |
| `reference-i4-s22-reviewed-incident` | `OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.reference_pack_fatal_completion,module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e,module.incidentbundles.integration.incident_bundle_import_final_publication_recheck_b7b07b3e93,module.incidentbundles.integration.extension_portability_matrix_and_atomic_import_1b27fc2d91,module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | 5 PASS; 18 completed cleanup steps |
| `reference-i4-s22-reviewed-server` | `OWNER=app.server ROWS=app.server.process.portable_reference_cleanup_grace` | 1 PASS; 14 completed cleanup steps |

Each command sets `CARTULARY_TEST_RUN_ID=<run ID> CARTULARY_HARNESS_CACHE_MODE=off`; roots are `.cartulary/test-results/<run ID>`. Cache hits 0 and retained-secret scans PASS for all three. S-23 begins next; findings remain open until final-source qualification and handoff.

I4-05 / S-23 qualification review (not a completion checkpoint): `reference-i4-s23-finalize` / `make agent-finalize` passed, native exit 0, on `bf9cb8db…56a08e`; the generated digest remained unchanged. The source snapshot is `/tmp/reference-i4-obq_3abd/final-source.json`. Added the exact cross-owner protocol row after `make task-guide OWNER=module.crossownertransaction` (exit 0) to qualify V-32 commit/cancellation/indeterminate precedence, bringing the planned final portfolio to 29 rows. Embedded destination verification is already invoked by the reference-catalog round-trip row; no duplicate selector is needed.

The first S-23 backend-boundary run `reference-i4-s23-boundary` failed (native outer 2, normalized product/test_assertion_failure). `make explain-run RESULTS_DIR=.cartulary/test-results/reference-i4-s23-boundary TARGET=lint DETAIL=logs` exited 0 and identified six newly added exact test-file imports missing from the owner-port allowlist. No production ownership violation or boundary broadening is required. Cleanup completed (13 steps), and the failed run remains retained. The independent path review also identified blocking observer and abort-lock returns that need a fatal recheck before ordinary classification/mutation. The first portfolio is allowed to drain unchanged; the boundary declaration and regression repairs will precede a new finalization/freeze and complete targeted requalification. F-21/F-23/F-24 remain open.

I4-06 / S-23 review repairs: all 29 first-portfolio product rows ultimately passed, native exits 0, on `bf9cb8db…56a08e`; generated/schema checks also passed. They are retained historical evidence after the following source repairs. The six exact test import declarations were added to the existing owner-port allowlists (no directory-wide exemption). Fatal identity is rechecked after observer/lease sampling, after the abort lock returns and after abort work, so a blocked call cannot resume into ordinary classification/publication. A real-persistence late-lock regression commits its otherwise empty fixture transaction to ensure rollback cannot conceal an unauthorized owner mutation. A further review case makes scope cleanup failure clear unpublished required-content verdict eligibility and stale ordinary classification; the rooted rejection-plus-cleanup-failure case must classify as operational failure.

`reference-i4-s23-review-format` and `reference-i4-s23-verdict-format` / `make format` passed, native exits 0. On intermediate digest `sha256:3e65a06cd4834543cdbe74411dc58c891b5c6d28e5aad615c662532b941d21be`, cache-disabled `reference-i4-s23-review-reference` (verifier row), `reference-i4-s23-review-incident` (new parent fatal-completion row), and `reference-i4-s23-review-boundary` (`make lint CHECK=backend-boundary`) passed with native exits 0, respectively 10/14/13 completed cleanup steps and passing retained scans. These review repairs are part of S-23's qualification repair loop; no finding closes until the replacement final-source evidence passes.

I4-07 / S-23 repair validation: `CARTULARY_TEST_RUN_ID=reference-i4-s23-verdict-reference CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.portable_completion_lifecycle,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2` passed, native exit 0, on `sha256:6b4bee4aded56e5462c3fc05e64d1e30a2651c5bb0e93b312aa9eefcc6324dc1`; 15 cleanup steps completed and retained-secret scan passed. Both the rooted required-rejection/cleanup-failure case and the late observer fatal-precedence case pass. A replacement `reference-i4-s23-finalize-qualified` precedes final freeze and fresh qualification; no earlier pass will be promoted as final evidence.

I4-08 / S-23 evidence-strength repair (not final completion): the replacement 29-row portfolio and generated/schema/boundary checks all passed on `6b4bee4a…324dc1`, native exits 0, completed cleanup and passing scans. Before closing V-30, the evidence review replaced nil finalizer dependencies with explicit failing finalizer methods: lifetime panic recovery could otherwise hide an attempted forbidden call. The worker unit now directly exercises classification rather than stopping at an inactive completion scope; the composed live-parent test still enters the public owner and production failure path. This is a test-only strengthening, with no production behavior change. Earlier results remain retained; final qualification will use the strengthened source snapshot and fresh IDs ending `-final`.

I4-09 / S-23 final freeze: `reference-i4-s23-finalizer-assertion-format` / `make format` and `reference-i4-s23-finalizer-assertion` / cache-disabled Incident Bundles fatal-completion plus worker-transition rows passed, native exits 0; the latter completed 15 cleanup steps and passed its retained scan. `CARTULARY_TEST_RUN_ID=reference-i4-s23-finalize-final make agent-finalize` then passed, native exit 0, with no generated-source change. Final source is `sha256:8d43665d18d847d817427e264c5f911847f87f5f36401266532657b43d892217`; `/tmp/reference-i4-obq_3abd/final-source-complete.json` freezes 8,506 non-document paths. Final invocations use fresh `reference-i4-s23-<selection>-final` run IDs. Completion remains gated on their terminal evidence and the documentation checkpoint.

I4-10 / S-23 timeout-evidence refinement: the rooted owner-timeout scenario originally waited for the full 60-second budget while its isolated fixture did not renew the parent Job lease; an earlier lease-loss cancellation could satisfy its generic error assertion. This does not invalidate the deterministic owner-deadline unit case or the explicit parent-deadline case, but it is insufficient rooted owner-timeout evidence. Refine the fixture's parent start instant to account for already elapsed parent work and assert `context.DeadlineExceeded` as the cancellation cause; also assert the explicit cancellation cause. Production budgets and runtime policy remain unchanged. Results on `8d43665d…892217` are retained, but the final handoff will bind to the refined test source.

I4-11 / S-23 exact-timeout proof: `reference-i4-s23-timeout-format` / `make format` and `CARTULARY_TEST_RUN_ID=reference-i4-s23-timeout-proof CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.portable_completion_lifecycle` passed, native exits 0. The rooted owner and parent timeout cases now assert `context.DeadlineExceeded`, and explicit cancellation asserts `context.Canceled`. Source is `sha256:136fd04132da784ccebea89ab6d70caa21e37c4dbf9746f4cc616c65ca3dc760`; no production timeout, Job policy or constructor change accompanied this fixture refinement. The 29-row portfolio on `8d43665d…892217` and its required source checks passed with native exits 0, but remain historical. Final qualification uses the exact-cause fixture source and `-complete` run IDs after `reference-i4-s23-finalize-complete`.


### S-23 final qualification and handoff

Final source: `sha256:136fd04132da784ccebea89ab6d70caa21e37c4dbf9746f4cc616c65ca3dc760`, dirty `main` at HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`. `reference-i4-s23-finalize-complete` passed with native outer exit 0 and no generated-source delta. Source was then frozen in `/tmp/reference-i4-obq_3abd/final-source-handoff.json`; all 8,506 non-document path/content/mode entries still match after qualification. Documentation is excluded from executable evidence identity. The native command/exit ledger is `/tmp/reference-i4-obq_3abd/final-native-exits.json`.

Every invocation below has an independently observed native outer exit **0**, a passing owning summary, completed cleanup with zero failures and a passing retained-secret scan. Every manifest carries the final digest above. The six product selections use `CARTULARY_HARNESS_CACHE_MODE=off` and have zero cache hits: **29/29 selected product rows pass**. Roots are `.cartulary/test-results/<run ID>`; inspect `run-manifest.json`, `run-summary.json`, `rows/*.json`, `cleanup-results.json` and `retained-secret-scan.json`. The failed formatter and first boundary run, earlier passing snapshots, diagnoses and test-strengthening decisions remain preserved above; none substitutes for these final results.

| Run ID | Public Make target / selection | Result / completed cleanup steps |
| --- | --- | --- |
| `reference-i4-s23-finalize-complete` | `make agent-finalize` | PASS / 10 |
| `reference-i4-s23-reference-complete` | `make test-slice OWNER=module.reference_data ROWS=<exact command below>` | 17/17 PASS / 24 |
| `reference-i4-s23-incident-complete` | `make test-slice OWNER=module.incidentbundles ROWS=<exact command below>` | 5/5 PASS / 18 |
| `reference-i4-s23-extensions-complete` | `make test-slice OWNER=module.extensions ROWS=<exact command below>` | 1/1 PASS / 10 |
| `reference-i4-s23-server-complete` | `make test-slice OWNER=app.server ROWS=<exact command below>` | 2/2 PASS / 15 |
| `reference-i4-s23-operator-complete` | `make test-slice OWNER=app.operator ROWS=<exact command below>` | 3/3 PASS / 16 |
| `reference-i4-s23-commit-complete` | `make test-slice OWNER=module.crossownertransaction ROWS=<exact command below>` | 1/1 PASS / 10 |
| `reference-i4-s23-generated-complete` | `make source-check CHECK=generated` | PASS / 14 |
| `reference-i4-s23-schema-complete` | `make source-check CHECK=schema` | PASS / 12 |
| `reference-i4-s23-boundary-complete` | `make lint CHECK=backend-boundary` | PASS / 13 |

Each command sets `CARTULARY_TEST_RUN_ID=<run ID>`. Exact product commands from the final manifests follow. Use fresh IDs for any future execution; never reuse these retained run directories.

```sh
CARTULARY_TEST_RUN_ID=reference-i4-s23-reference-complete CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.integration.portable_completion_lifecycle,module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a,module.reference_data.integration.reference_pack_import_uses_the_shared_upload_env_95574f69b8,module.reference_data.integration.workspace_cleanup_lifecycle,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_archive_admission,module.reference_data.unit.canonical_pack_diagnostics,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_state,module.reference_data.unit.canonical_pack_trust,module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f,module.reference_data.unit.reference_pack_request_decoders_enforce_activati_7074ebbbda,module.reference_data.unit.reference_pack_storage_is_atomic_and_contained_14f3d4892a,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2
CARTULARY_TEST_RUN_ID=reference-i4-s23-incident-complete CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.extension_portability_matrix_and_atomic_import_1b27fc2d91,module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8,module.incidentbundles.integration.incident_bundle_import_final_publication_recheck_b7b07b3e93,module.incidentbundles.integration.reference_pack_fatal_completion,module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e
CARTULARY_TEST_RUN_ID=reference-i4-s23-extensions-complete CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.extensions ROWS=module.extensions.unit.admission_invocation_v6
CARTULARY_TEST_RUN_ID=reference-i4-s23-server-complete CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=app.server ROWS=app.server.process.cancellation_grace_terminates_blocked_startup_se_b4b985d465,app.server.process.portable_reference_cleanup_grace
CARTULARY_TEST_RUN_ID=reference-i4-s23-operator-complete CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission,app.operator.process.reference_cleanup_grace,app.operator.unit.operator_registry_eight_paths_and_collaboration_01217009b8
CARTULARY_TEST_RUN_ID=reference-i4-s23-commit-complete CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.crossownertransaction ROWS=module.crossownertransaction.unit.protocol_boundaries_7f6ad80724
```

| Obligation / authority | Final evidence and boundary |
| --- | --- |
| V-27 / RP-REQ-228/238; EXT-REQ-215 | Verifier lifetime selection exercises prior cancellation, earlier parent and owner deadlines, zero grace, context values, inherited exhaustion, shared cleanup, observer joining, late results and rejection of completion after scope return. Rooted two-container disposal and subsequent completion retain one fatal selection. Exact-cause rooted timeout assertions distinguish deadline expiry from Job lease loss. The cross-owner row retains commit/rollback completion bounds; the parent failure path now carries the inherited tracker through its local timeout. |
| V-28 / RP-REQ-094/098/228/238 | `TestPortableRootedCompletion_Integration` exercises success, optional/required rejection, cancellation, parent/owner timeout, partial write and source-close failure through the public owner boundary with actual storage. Scratch is absent before cooperative return; actual two-container disposals and publication leases complete. Required rejection followed by parent cleanup failure is operational, never a content verdict. Canonical persistence and parent tests preserve admitted frozen inputs and exact replay. No test calls reconciliation to establish cooperative cleanup. |
| V-29 / RP-REQ-228/238; EXT-REQ-215 | Fatal precedence tests cover late nil/success, close failure, cancellation, timeout and content errors; fatal source preparation survives source-close failure. Exhaustion selected during observer shutdown precedes late lease-loss observation. Cleanup failure invalidates unpublished content-verdict eligibility and stale ordinary classification. Extensions' read-only query neither starts nor resets tracking. |
| V-30 / RP-REQ-228; Core fatal lifecycle | `TestPortableFatalCompletionPreservesParent_Integration` uses actual owner preparation and persistence, delays callback completion after fatal selection, and invokes the production Incident Bundles failure path with explicit forbidden-finalizer assertions. The Job remains running; catalog, pin, version, event and owner-terminal state is unchanged; frozen objects remain available. Its late abort-lock case commits the test transaction so a rollback cannot hide unauthorized abort writes. Late lock/observer results receive another fatal check. |
| V-31 / EXT-REQ-215; Core 04 REQ-04-144/146 | `TestPortableReferenceCleanupGrace_Process` enters the production portable boundary with blocked source closure or rooted workspace deletion. Children run the actual server fatal sink and exit 70 with exact registered `extension_integrity_failure` / `operation_cancellation_grace_exhausted` JSON diagnostic. Composition seams and faults are test-only; this is not a deployed listener test. Existing server lifecycle and operator cleanup consequence rows pass on the same final source. |
| V-32 / RP-REQ-228/239 | Canonical persistence and reference-catalog round-trip cover references-only reuse, redundant ineligible bytes, optional/required semantics, immutable frozen retries, committed replay, multi-container cohorts, publication rollback and lost commit acknowledgement. The cross-owner protocol row supplies proven/absent/indeterminate commit precedence against cancellation and late completion; the portable unit retains recoverable lease-loss classification. Rooted failure and live-parent tests cover operational cleanup and retained input ownership. The workspace-lifecycle row retains serving reconciliation, confinement, live-lease isolation and joined maintenance-shutdown evidence. |

Separate final diff-to-obligation review traced the production worker's `RunImportExecution` callback through fresh input probing, complete Reference Data preparation, deferred source close/disposal/lease release, later source-owner preparation, publication and parent teardown/failure handling. It checked observer/timer ordering, exhaustion queries before and after blocking calls, ordinary abort guards, fatal-preserving error merging and committed/indeterminate ownership flags. All superseded manual entry points and the narrow verifier wrapper are absent. The final assertions detect forbidden finalizer calls directly rather than relying on swallowed panics, and timeout cases assert their exact causes. No remaining scoped specification or implementation blocker was identified after the recorded repairs. This is a targeted conformance review, not complete-owner or release certification; tests do not establish specification completeness.

Preservation relative to I4-00: 23 existing files changed and five files were added; none was removed. The changes comprise the portable owner/worker boundary and tests, read-only Extensions query, RP-REQ-228 clarification, tracker, three authored selector additions, their generated render index and six exact test import allowances. Existing unrelated dirty-tree work and staged content remain preserved. No persistence schema, public API, pack format, client, dependency-lockfile or configuration migration is introduced. No CI, full-owner suite, browser expansion or release-certification run was performed.

Deployment: no reset or data conversion is required. Existing cohorts resume using retained inputs. Cooperative portable cancellation removes scratch synchronously within the original remaining grace; genuine residual storage failures remain eligible for startup and serving recovery. Non-cooperative completion uses fatal exit 70 and leaves admitted work available to existing recovery. Proven committed and indeterminate bytes retain their existing ownership protections.

Rollback the callback interface and callers, completion/classification policy, tests, owner clarification, catalogs/generated index and exact boundary declarations together. Do not restore a manual unsupervised fallback, renew cancellation grace, erase retained inputs or silently disable fatal enforcement/reconciliation.

Finalizer canonical-evidence and scheduler-drift maintenance actions were skipped with `results-dir-not-provided`: no compatible full warm-check root was supplied, and none was created to expand this targeted task. Current-source selected rows, owning summaries, cleanup and retained scans were inspected directly; this is not a claim that the skipped full-run audits executed. The final documentation checkpoint below completes S-23.


I4-12 / S-23 validation and handoff COMPLETE: all V-27–32 obligations have the final-source targeted evidence above. `CARTULARY_TEST_RUN_ID=reference-i4-s23-handoff make lint-markdown` passed with native outer exit 0 and a passing `adhoc/lint-markdown/tool-run-summary.json`. `git diff --check` and `git diff --cached --exit-code` exited 0. The final non-document snapshot comparison found no added, removed or modified input after freeze. F-21/F-23/F-24 are resolved; F-22 remains closed; S-19–23 are DONE. No scoped blocker remains. The completion-status transcription receives one final documentation-only check under `reference-i4-s23-handoff-final`; CI and full-owner suites remain excluded.

## 19. Fatal-finalization race remediation

The user authorized this implementation on 2026-11-07. F-25 (P1, resolved in S-27), linked to F-24/F-21, records that an ordinary terminal transaction can pass `AbortTx`, block in a later owner/Jobs/receipt write, then resume after fatal exhaustion and commit. S-23's evidence did not cover this interval. Its 29 passing rows on `sha256:136fd04132da784ccebea89ab6d70caa21e37c4dbf9746f4cc616c65ca3dc760` remain historical; its complete F-21/F-24 closure and V-30 claim are superseded. F-22/F-23 remain closed.

I5-00 baseline: dirty `main`, HEAD `3cc655473f872f796fd0dd07d1b8bb99e7947b2b`. `/tmp/reference-i5-b58exaks/baseline.json` records tracked/nonignored file content/mode and index fingerprints. This tracker reconciliation is the first repository mutation. Preserve preceding changes and retained runs. No CI restart or full-owner suite is authorized or required.

Authority: RP-REQ-228/238 and EXT-REQ-215 already require one bounded portable scope through parent completion and fatal precedence; Core 04 REQ-04-144/146 supplies the process consequence. This is an implementation/projection repair, not a relaxation or behavioral amendment. Proven committed state remains authoritative; fatal lifetime selection cannot authorize a replacement terminal outcome.

Remediation (implementation, tests, documentation): give ordinary terminal publication a joined completion context that survives semantic cancellation only until the inherited original grace exhausts. Revoke it synchronously when fatal exhaustion is selected, before signaling the process sink. Carry it through owner, Job, receipt and commit work; reject dispatch of a terminal commit with a revoked context while preserving the existing proof of an already-dispatched commit. Keep physical teardown separate from publication authority. No new timer setting, cancellation tracker, detached worker, schema, API or compatibility alias is introduced.

Rationale and benefit: one lifetime decision controls both process termination and permission to publish, eliminating reliance on the asynchronous process sink winning a race. Unresolved, late failed/canceled Job evidence can hide retained recovery work. Compatibility impact is internal only; committed and indeterminate ownership, cooperative cancellation and original grace remain intact. Roll back the completion capability, callers, finalizer guard and regression tests together.

| Slice | Dependency | Work / principal risk | Exit criterion | Status |
| --- | --- | --- | --- | --- |
| S-24 | Review | Record baseline, reopen scoped findings, map authority and discover exact tests. Risk: overstating old evidence. | Tracker reconciliation and Markdown validation pass. | DONE |
| S-25 | S-24 | Add revocable joined publication completion and propagate it through terminal commit. Risk: renewed grace or overturning proven commit. | Focused lifetime/publication tests pass; no unsupervised publication fallback. | DONE |
| S-26 | S-25 | Add actual persistence/common-finalizer regression blocking after `AbortTx`; prove rollback and retained recovery inputs, plus commit precedence. Risk: helper-only assertions masking durable writes. | Targeted real-persistence and commit proof cases pass. | DONE |
| S-27 | S-26 | Finalize, freeze, qualify affected exact rows, review diff and finish handoff. Risk: stale-source evidence or premature closure. | Compatible final targeted evidence and explicit obligation mapping; final tracker checkpoint. | DONE |

Validation obligations: V-33 proves inherited grace revokes publication before the fatal sink while retaining values and allowing cooperative cancellation; V-34 blocks after successful `AbortTx` with real persistence and the common finalizer, delays the process sink, then asserts no terminal Job/owner/receipt publication and preserved recovery inputs; V-35 preserves proven commit, including lost acknowledgement and cancellation races. Retain earlier portable cleanup and process consequences with affected targeted regressions. Discover rows through `make task-guide`; extend existing selectors where appropriate. Run `make agent-finalize`, inspect its changes and freeze before final cache-disabled selections. Complete generated/schema checks, backend-boundary lint, Markdown lint and `git diff --check`. Record commands, native exits, source digest, roots, cleanup/scans and any failure diagnosis. Update this tracker after each workstream and before its successor.

I5-01 / S-24 complete: `CARTULARY_TEST_RUN_ID=reference-i5-s24-docs make lint-markdown` passed, native exit 0 (session 22500), retained tool summary under `.cartulary/test-results/reference-i5-s24-docs`. `make task-guide OWNER=module.extensions`, `module.incidentbundles`, and `module.reference_data` each exited 0 (sessions 58649/12882/25316); exact discovery outputs are retained in `/tmp/reference-i5-b58exaks/guide-{extensions,incident,reference}.txt`. Existing owner requirements suffice; no normative amendment is needed. S-25 begins next.

I5-02 / S-25 complete: Extensions `CompleteWithinGrace` joins publication and synchronously revokes all registered publication contexts at inherited fatal exhaustion before the external sink. It retains context values, refuses already exhausted/expired grace and removes registrations at completion. Physical cleanup keeps its existing separate completion path. Portable `CompleteFinalization` now owns failure publication; worker completion errors propagate fatal/indeterminate retention. The common terminal finalizer checks context cause before dispatching commit; existing proof classification remains unchanged after dispatch. `reference-i5-s25-format` / `make format` passed (native 0, session 56066). Cache-disabled `make test-slice OWNER=module.extensions ROWS=module.extensions.unit.admission_invocation_v6` (`reference-i5-s25-lifetime`, session 68419) and `OWNER=module.reference_data ROWS=module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2` (`reference-i5-s25-reference`, session 65377) each passed, native 0. These focused rows cover shared/zero/expired grace, revocation before a delayed sink, late-result fatal precedence, cooperative completion and existing verification lifetime; real durable post-AbortTx exclusion remains S-26.

I5-03 / S-26 complete: the existing portable fatal-completion selector now includes real production import admission, claimed extension Job, canonical rooted pack preparation, production parent failure orchestration and the actual common finalizer. One barrier intercepts the first owner payload write after `AbortTx` succeeds; the other holds the completed receipt write immediately before commit. Both verify the owner abort exists inside the transaction, then exhaust grace while the test sink delays process termination. Publication is already revoked at the fatal signal; late callbacks remain joined, no Job/payload/owner/receipt/catalog/pin/proof changes persist, and frozen object bytes remain readable. A cooperative case retains ordinary failed publication. The common-finalizer selector adds a canceled-context precommit guard assertion (zero commit dispatches), alongside its existing postcommit/lost-acknowledgement matrix for success/failure/timeout/cancellation.

`reference-i5-s26-format` / `make format` passed (session 53391, native 0). Cache-disabled `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.reference_pack_fatal_completion` (`reference-i5-s26-persistence`, session 80259) and `OWNER=module.extensions ROWS=module.extensions.integration.job_finalization_cancellation_and_reconciliation_2c712826c2` (`reference-i5-s26-commit`, session 16741) passed, native 0, source `sha256:50d9889c9f283f2b81d46bdbe33479d25cd4b45f3c0f33bcf293dc386f9ef227`; all 14 cleanup entries per run completed and retained scans passed. `reference-i5-s26-boundary` / `make lint CHECK=backend-boundary` passed (session 2695, native 0; 13 completed cleanup entries; scan pass). No boundary exception was added. The composed fixture then gained deterministic barrier release on assertion failure and an explicit joined-return assertion; `reference-i5-s26-format-final` passed (session 18359, native 0) and the same exact portable selector passed again under `reference-i5-s26-persistence-final` (session 51916, native 0). No failed runs occurred. Existing catalog selectors include the new subcases; no authored selector or generated routing change is required.

S-27 begins next. Additional `make task-guide OWNER=module.crossownertransaction`, `app.server`, and `app.operator` passed (sessions 97449/83107/75586, native 0). Final qualification is recomputed for this repair: portable owner/parent behavior, shared Extensions lifetime and terminal finalization, cross-owner commit precedence, and affected server/operator process consequences. Historical unrelated format/profile/HTTP/UI rows will not be repeated or promoted as current-source evidence. Full-owner and CI remain excluded.

I5-04 / S-27 source freeze: `CARTULARY_TEST_RUN_ID=reference-i5-s27-finalize make agent-finalize` passed (session 86921, native 0). Its changes were inspected: no generated, schema, lockfile or unrelated source delta was introduced relative to I5-00. Ten existing source/test files and this tracker are the entire iteration delta. `make task-guide DETAIL=coverage OWNER=<owner> JSON=1` passed for Incident Bundles, Reference Data, Extensions, server, operator and cross-owner transaction (sessions 33342/8883/96759/17835/65775/28124, native 0); every selected row was checked against these discoveries. `/tmp/reference-i5-b58exaks/portfolio.json` records the exact 15-row final selection. `/tmp/reference-i5-b58exaks/final-source.json` freezes all 8,887 tracked/nonignored regular inputs except this controlling handoff; subsequent changes are documentation only. Final runs use fresh IDs, cache disabled, and no CI/full-owner invocation.

I5-05 / S-27 final targeted qualification: **15/15 exact product rows passed**, cache disabled, on `sha256:98eecb244fdb833be1e58e52b182405e2aa199035f8385e18b497dec29a49249`. All native outer exits are 0. Every selected row result is `passed`, with no cache hits. The final manifest, selected row set, cleanup records and retained-secret scan were inspected directly; no full-owner evidence audit or release claim is implied. Native exits/commands and the inspection record are retained in `/tmp/reference-i5-b58exaks/final-native-exits.json` and `final-evidence-inspection.json`.

All run roots below are beneath `.cartulary/test-results/`. Every listed cleanup entry completed without failures, and every retained-secret scan passed.

| Run ID | Scope | Native session / outer exit | Rows | Completed cleanup entries |
| --- | --- | --- | --- | --- |
| `reference-i5-s27-reference` | `module.reference_data` | 21226 / 0 | 4 | 21 |
| `reference-i5-s27-incident` | `module.incidentbundles` | 60854 / 0 | 5 | 18 |
| `reference-i5-s27-extensions` | `module.extensions` | 22867 / 0 | 2 | 15 |
| `reference-i5-s27-server` | `app.server` | 74186 / 0 | 2 | 15 |
| `reference-i5-s27-operator` | `app.operator` | 23657 / 0 | 1 | 10 |
| `reference-i5-s27-commit` | `module.crossownertransaction` | 75876 / 0 | 1 | 10 |
| `reference-i5-s27-finalize` | `make agent-finalize` | 86921 / 0 | — | 10 |
| `reference-i5-s27-generated` | `make source-check CHECK=generated` | 65048 / 0 | — | 14 |
| `reference-i5-s27-schema` | `make source-check CHECK=schema` | 4251 / 0 | — | 12 |
| `reference-i5-s27-boundary` | `make lint CHECK=backend-boundary` | 50189 / 0 | — | 13 |

Exact final product commands follow. Use fresh run IDs for any future execution; these retained identities must not be reused.

```sh
CARTULARY_TEST_RUN_ID=reference-i5-s27-reference CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.integration.canonical_persistence,module.reference_data.integration.portable_completion_lifecycle,module.reference_data.integration.workspace_cleanup_lifecycle
CARTULARY_TEST_RUN_ID=reference-i5-s27-incident CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.reference_pack_fatal_completion,module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e,module.incidentbundles.integration.incident_bundle_import_final_publication_recheck_b7b07b3e93,module.incidentbundles.integration.extension_portability_matrix_and_atomic_import_1b27fc2d91,module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8
CARTULARY_TEST_RUN_ID=reference-i5-s27-extensions CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.extensions ROWS=module.extensions.unit.admission_invocation_v6,module.extensions.integration.job_finalization_cancellation_and_reconciliation_2c712826c2
CARTULARY_TEST_RUN_ID=reference-i5-s27-server CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=app.server ROWS=app.server.process.cancellation_grace_terminates_blocked_startup_se_b4b985d465,app.server.process.portable_reference_cleanup_grace
CARTULARY_TEST_RUN_ID=reference-i5-s27-operator CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=app.operator ROWS=app.operator.process.reference_cleanup_grace
CARTULARY_TEST_RUN_ID=reference-i5-s27-commit CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.crossownertransaction ROWS=module.crossownertransaction.unit.protocol_boundaries_7f6ad80724
```

| Obligation | Final evidence and scope |
| --- | --- |
| V-33 / EXT-REQ-215; RP-REQ-228/238 | Extensions lifetime row exercises shared revocation before a deliberately delayed sink, inherited tracking, zero grace, expired inherited deadlines, context values, cooperative cancellation and normal scope release. Late nil results and later completion cannot replace fatal selection or renew grace. Reference Data lifetime and rooted portable rows retain joined resource completion and existing timeout behavior. |
| V-34 / RP-REQ-228; Core fatal lifecycle | `TestPortableFatalCompletionPreservesParent_Integration/after_abort` blocks the real owner payload write after successful `AbortTx`; `after_receipt` blocks after real owner, Job and receipt writes. Both use actual common finalization and production parent failure orchestration. Fatal selection revokes the publication context before the test sink returns; no operation returns while the transaction callback is held. After release, snapshots prove unchanged Job, owner payload/operation/selections and receipt, unchanged catalog/pin/version/event/proof counts, and readable frozen input bytes. Existing pre-finalizer and abort-lock cases remain. The common-finalizer precommit case separately asserts zero COMMIT dispatches after revocation. |
| V-35 / RP-REQ-228; EXT-REQ-215 commit precedence | Common-finalizer integration retains proven success/failure/timeout/cancellation after cancellation and lost commit acknowledgement. The new cooperative parent case commits ordinary failure within grace. Cross-owner protocol, parent publication recheck and reference-catalog round-trip rows retain proven/absent/indeterminate publication, replay and retained-input ownership. A commit already dispatched is classified by proof; no postcommit cancellation check replaces proven state. |
| Shared process/recovery regressions | Server cancellation and portable cleanup process rows and operator cleanup process row retain actual fatal exit 70 and safe diagnostic assertions. Portable/rooted and workspace lifecycle rows retain cleanup, serving reconciliation, lease safety and shutdown joining. No test manually reconciles scratch to claim cooperative cleanup. |

A separate final diff review traced `RunImportExecution` through `CompleteFinalization`, classification, `AbortTx`, payload/Jobs/receipt writes and the physical commit boundary. Publication registration and exhaustion share the tracker mutex; all registered contexts are canceled before the sink is called or exhaustion is observed through the tracker. Late callbacks cannot dispatch a new terminal commit with a revoked context. Already-dispatched commits retain their existing proof path; fatal/indeterminate completion does not authorize replacement finalization or input disposal. Physical teardown remains joined under its original lifetime and never uses publication permission as a deletion policy. No remaining scoped implementation/specification blocker was identified. The new cases exercise persisted transactions rather than only guarded helper entry points.

The final-source snapshot comparison reports no changed, added or removed frozen input after qualification. Relative to I5-00, only ten existing implementation/test/fixture files and this tracker changed; no files were added or removed. Prior dirty-tree history and staged content are preserved. No specification amendment, schema, public API, pack format, configuration, dependency lockfile, boundary exception or catalog selection change was required. Existing selectors include the new subcases; finalization refreshed and checked generated projections without a generated delta.

Deployment requires no reset or data conversion. Cooperative portable cancellation can complete ordinary terminal publication within the original remaining grace. Fatal exhaustion revokes further publication before signaling the asynchronous process sink; admitted recovery inputs remain available. Proven committed bytes remain authoritative, and indeterminate bytes remain protected. Existing startup/serving scratch recovery is unchanged. Roll back the Extensions completion capability, portable/worker callers, common-finalizer guard and tests together; do not restore an unsupervised fallback or silently suppress fatal enforcement.

Finalizer canonical-evidence and scheduler-drift maintenance were skipped with `results-dir-not-provided` because no compatible full warm-check root was supplied. No full run was created for those optional maintenance paths. Current-source targeted rows, native exits, cleanup and scans were inspected directly. No new failure occurred during this iteration; earlier retained failures and their diagnoses remain intact. CI, full-owner suites, UI/browser expansion and release certification were not run. The documentation validation and final S-27 handoff checkpoint follow.

I5-06 / S-27 validation and handoff COMPLETE: V-33–35 pass on the frozen final source above, with the affected portable/recovery/process regressions passing as well. `CARTULARY_TEST_RUN_ID=reference-i5-s27-handoff make lint-markdown` passed (session 30298, native outer exit 0), with a passing `adhoc/lint-markdown/tool-run-summary.json`. `git diff --check` and `git diff --cached --exit-code` passed; the frozen-input comparison found no source drift. F-21/F-24/F-25 are resolved; F-22/F-23 remain closed. S-24–27 are DONE and no scoped blocker remains. The completion transcription receives the final documentation-only check `CARTULARY_TEST_RUN_ID=reference-i5-s27-handoff-final make lint-markdown`; no implementation source changes follow the freeze. No CI restart or full-owner suite was performed or required.

## 20. Current producer consolidation and proposal retirement

### I6 design decisions and authority

The selected next iteration removes the active ATT&CK proposal workflow and simplifies current producer execution. It does not manufacture another runtime refactor after the completed helper/lifecycle removals. F-26–29 are evidenced legacy responsibilities; no additional dead runtime function was proven. S-28–32 execute under the single prerequisite graph in §6; completion evidence follows.

Reference Pack 0.4.0/major 3 is adopted/current. RP-REQ-001/002 allocate source/pack behavior and imported outer interfaces; RP-REQ-149–152 require offline source interpretation, immutable source-profile provenance, and current runtime admission independent of source parsers. The current native ATT&CK profile and Core 00's current-only cutover support the chosen producer simplification without an amendment to runtime requirements. The producer CLI/report cutover is explicitly authorized; obsolete compatibility paths are removed. VERIS adoption RP-REQ-270–274 is reflected in inventory/authority only.

| Decision | Concrete implementation / acceptance condition |
| --- | --- |
| Current producer input | CLI audit selects the retained `source-profile.native.v2.r2.json` and its exact native field policy; remove `--candidate` entirely. Validate admitted lock, profile ID/status, exact policy digest and current schema digest bindings with the same owner-local admission used by assembly. No fallback to proposal inputs, prose, a live website or a newer profile. |
| One audit path | Replace proposal-index loading and native-candidate validation with current schema validation and existing source semantic auditing. Preserve strict source bytes, duplicate/collision/reference/lifecycle handling, deterministic ordering, cancellation and source-retention boundaries. Current native schemas come from the existing typed contract projection; do not copy another schema set into producer code. |
| Compiler boundary | Retain the current compiler's behavior in a private producer-local helper with a neutral name. Migrate Windows and current producer callers; remove the exported candidate name without an alias. Do not turn this into a generic cross-module JSON Schema service. |
| Report v2 | Emit `cartulary.attack_source_audit.v2`. Preserve source-audit pass/fail meaning and remaining factual fields; add `source_profile_id` and `source_profile_sha256`; replace `proposed_lookup_aliases` with `lookup_aliases`; remove `pack_status` and `native_candidate`. No v1 writer/reader adapter or synthetic adoption-success status. Audit success is source evidence, not runtime conformance, production trust or distribution approval. |
| Report retention / destination | Continue content-addressed audit retention. When `--report-file` is supplied, write exactly the canonical bytes retained for that audit, including attributable failed-audit reports; no successful report when admission fails before a report exists. Preserve nonzero failure exit and existing independent witness dispatch for other producers; do not add a new release claim or witness system. |
| Immutable provenance | Leave source profiles, native policy bytes and their IDs/digests unchanged, including historical `candidate_id` spelling and prose inside the pinned policy/profile JSON. They are immutable provenance, not compatibility aliases. Do not regenerate retained packs, signed envelopes, ledgers or historical reports. New report hashes may differ; for identical explicit builder and pack inputs, manifest/payload/notice bytes and publication identity must not. |
| Test/input closure | Replace broad ATT&CK producer input embedding with the explicit set still consumed by current producer/tests. Remove proposal index/schema dependencies from current tests. Keep immutable historical artifacts at their existing paths when referenced by retained provenance; their presence is not an active reader. Preserve current acquisition schema, useful source fixtures, current/frozen profiles and source-lock bytes as needed by their actual consumers. |
| Coherent cutover | S-30 removes execution-only candidate types; report-only DTO types remain only until S-31 removes their last use. S-30/31 are one releasable change and must not be independently deployed. This avoids uncompilable intermediate slices without creating a supported legacy mode. Roll back producer, CLI, report consumers, tests and routing together. |

The generic fixture signing functions and the current Base reissue path materially support current artifacts and stay. Historical v1 rejection fixtures, live runtime fallback, retained-state validation, portable lifetime/finalization and cleanup safeguards remain. The names `v1`, `candidate` or `legacy` alone are insufficient deletion evidence.

### Gap remediation accounting

| Gap / areas | Remediation and rationale | Long-term benefit | Compatibility / migration | Risk if unresolved | Completion criteria |
| --- | --- | --- | --- | --- | --- |
| F-26 — implementation, tests, documentation; existing adopted specification suffices | Share exact frozen profile, lock, policy and compiled-schema admission across audit and assembly; remove proposal flag, loader and validator. One source contract eliminates divergent authority. | Smaller supported surface, consistent fail-closed admission, one owner-local extension point for a future deliberately versioned profile. | Remove `--candidate`; callers supply current pinned inputs. No compatibility alias, runtime or data migration. | Proposal validation could diverge from current packaging and accept unsupported provenance. | V-36/38/40: all rejected input variants fail without publication; both retained releases audit/assemble; old execution symbols and dependencies absent. |
| F-27 — implementation, tests, documentation; producer report is a support contract | Introduce report v2 with exact profile provenance and truthful audit status; retain and deliver one canonical serialization, including failed partial audits. | Reports are deterministic, attributable and independent of adoption/trust claims. | Report readers select v2, rename `proposed_lookup_aliases` to `lookup_aliases`, consume profile ID/hash and drop `pack_status` / `native_candidate`. Old reports remain historical; no dual writer. | Misleading adoption blockage and incompatible report destinations undermine diagnosis and downstream automation. | V-38/40: exact success/failure delivery bytes, nonzero failure, no fabricated pre-admission report, repeatable partial reports and real audit-to-build linkage. |
| F-28 — implementation and tests | Move the unchanged schema compiler into private producer-local `schema.go`; preserve strict decode, format checking and disabled external loaders. | Windows remains independent of retired proposal machinery without creating a shared platform abstraction. | Internal symbol rename only; no exported alias or Windows schema change. | Deleting the candidate file could break live Windows validation or preserve unnecessary coupling indefinitely. | V-37: Windows row and malformed/duplicate/schema/remote/file/date-time compiler assertions pass. |
| F-29 — tests, authored routing/generated projection, documentation and embedded inputs | Migrate all 13 useful assertions, retain current source/ledger tests, explicitly embed consumed inputs and remove obsolete selectors. Preserve original historical files. | Verification follows adopted contracts; build dependencies expose only active inputs while provenance remains recoverable. | No fixture reissue, historical digest rewrite or product schema migration. New CLI row exercises the actual command/delivery boundary. | Lost assertions conceal regressions; broad embedding carries proposal dependencies forward and obscures what production consumes. | V-36/39/41: assertion map complete, source checks and boundary lint pass, no active proposal reader, immutable artifacts unchanged. |

### Candidate assertion destinations

This is the S-28 assertion mapping. All entries now execute against current contracts; E-I6-S28 records the pre-removal pass. Retarget synthetic values to current schema IDs/shapes without weakening their semantic assertions. Preserve existing current-schema assertions when they already cover a case; do not keep a parallel proposal validator solely to run its old test.

| Existing case | Current destination / obligation |
| --- | --- |
| `native_optional_ids_original_urls_and_independent_flags` | Current native-source validation/projection: optional presence, unchanged original URLs and independent lifecycle flags |
| `closed_citation` | Current native-source schema rejects unknown citation members |
| `lifecycle_boolean` | Current native-source schema rejects wrongly typed lifecycle flags |
| `typed_component_reference` | Current native-source schema rejects wrongly typed component references |
| `object_native_and_canonical_projection` | Current object wrapper accepts matching native source and canonical projection |
| `native_source_required` | Current object wrapper rejects absent source |
| `extensions_closed` | Current object/relationship contract retains closed extensions |
| `sro_citations_and_upstream_identity` | Current relationship wrapper preserves original SRO identity and citations |
| `sro_original_required` | Current direct-SRO relationship rejects absent original source |
| `embedded_has_no_fabricated_sro` | Current embedded relationship accepts explicit null source with real source locators |
| `embedded_cannot_claim_sro` | Current embedded relationship rejects invented SRO source |
| `retained_acquisition_inventory_shape` | Live acquisition reader/current acquisition schema accepts exact inventory shape |
| `acquisition_requires_tag_reference` | Live acquisition reader/current acquisition schema rejects missing pinned tag evidence |

Also retain the independent SHA-256 `abc` witness, source-audit determinism, alias-collision accounting, source retention/replay, wrong-property/type and missing-endpoint rejection, runtime producer exclusion, publication-ledger concurrency/collision/sequence/tamper assertions, and compiler-failure rejection. Move the malformed-schema compiler case into the private current-schema helper tests rather than preserving proposal-only types.

New current-path cases must exercise exact profile/policy/schema binding, rejection of proposal or unbound inputs, CLI `--candidate` rejection, identical retained/report-file bytes on success and attributable failure, and current audit-to-assembly output preservation. Synthetic unit inputs may exercise private admission/semantic seams; the production entry point must never accept synthetic unpinned sources merely to simplify tests. Product tests consume retained canonical/machine fixtures, never live upstream data or Cartulary prose. S-32 qualifies named and byte-validated retained sources offline; no upstream acquisition is included.

### Exact qualification portfolio

These five exact rows define the final portfolio, including the relocated CLI selector. Focused checkpoints run first. After implementation, run `make agent-finalize`, inspect every resulting delta and freeze source before this final portfolio:

```sh
CARTULARY_TEST_RUN_ID=reference-i6-s32-final-fixed CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f,module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.integration.attack_producer_artifacts_preserve_independent_a_67e9ef139a,module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_9107ea959a
```

The unit rows need installed pinned tools; the lifecycle row needs harness-managed PostgreSQL and object store. Fresh final run IDs must not reuse an existing retained root. Extend the ATT&CK producer selector with new current-path cases and remove `TestATTACKNativeCandidate` only after its assertions are accounted for. Canonical-artifact lifecycle success cannot substitute for exercising the changed audit/assembly entry path. No producer source acquisition, fixture regeneration, full-owner/CI, browser, visual or release suite is required merely to close this bounded iteration.

Generated/schema checks and backend-boundary lint accompany the final portfolio where specified in §8. If source changes after qualification, diagnose and rerun the affected obligations on compatible source rather than combining stale passes. Retain native exits, exact selections, run roots, cleanup/secret-scan outcomes and any failed-run diagnoses. Optional finalizer maintenance without a compatible full warm-check root remains explicitly skipped; do not manufacture a broad run solely for maintenance.

### I6 document-session checkpoints

I6-00 / E-I6-01: clean baseline captured before mutation. `/tmp/reference-i6-docs-iigz5fsh/baseline.json` records 8,936 tracked/nonignored file identities/content hashes plus index identity/hash; `tracker-before.md` retains the original tracker; `target-inventory.json` accounts for all 259 target paths; `ignored-baseline.json` covers relevant ignored prerequisite/cache locations. Current tracker identity/content rechecked immediately before this update. No production refactor, test/catalog change, dependency installation, source acquisition or fixture mutation was performed.

I6-01 / T-028: controlling §§1–12 reconciled; prior permissions/results marked historical; F-26–29, S-28–32, V-36–41, RB-004 and the exact assertion map added. All historical IDs/results remain traceable. The selected clean producer cutover preserves frozen input bytes and separates live utility movement from changed CLI/report behavior. Documentation validation and final preservation comparison are the remaining current-session gates.

The admitted documentation command uses existing pinned tools, installed-only prerequisite verification, network-disabled Corepack, external result/cache directories and Markdown `fix:false`:

```sh
CARTULARY_PREPARATION_POLICY=installed_only COREPACK_ENABLE_NETWORK=0 CARTULARY_TEST_RESULTS_DIR=/tmp/reference-i6-docs-iigz5fsh/results CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/reference-i6-docs-iigz5fsh/static-cache CARTULARY_TEST_RUN_ID=reference-i6-docs make lint-markdown
```

This is the only managed validation kind used for this document session. `git diff --check` and the content/index/identity comparisons complete preservation review. Format, generation, agent-finalize, product suites, producer builds and source/package commands are intentionally skipped because this is documentation-only work. No expected or historical product check is treated as a current pass.


I6-02 / T-028 document update COMPLETE: the exact documentation command above passed with native outer exit 0 (session 49853). Its retained summary is `/tmp/reference-i6-docs-iigz5fsh/results/reference-i6-docs/adhoc/lint-markdown/tool-run-summary.json`; it reports pass, zero failures and a 24,120 ms tool duration. `git diff --check` passed. `/tmp/reference-i6-docs-iigz5fsh/preservation-checkpoint.json` confirms that only this tracker changed among all 8,936 baseline inputs; HEAD/index content and identity are unchanged, with no changes in the inspected ignored prerequisite/cache locations. No production refactor or product verification ran. Planning is complete for this bounded iteration; S-28 remains ready for a later authorized implementation task, S-29–31 remain blocked by RB-004/prerequisites, and S-32 awaits their completion.

Completion transcription is checked once more with the same command/environment and fresh run ID `reference-i6-docs-final`; its summary path is `/tmp/reference-i6-docs-iigz5fsh/results/reference-i6-docs-final/adhoc/lint-markdown/tool-run-summary.json`. That check and the final source/index/identity comparison are recorded in `/tmp/reference-i6-docs-iigz5fsh/final-checks.json` and `preservation-final.json`; this sentence specifies the artifact locations and does not preclaim the check's result. No repository edit follows the final successful check. Relevant ignored locations are covered by the named snapshot; no claim is made to hash every unrelated ignored historical run or dependency tree.


### I6 implementation checkpoints

I6-03 / S-28 IN_PROGRESS: implementation authorized on 2026-11-07. Baseline remains `main` at `a8f2fe2d84eecf88ced265b34327e0a8c7d4c8df`; only this tracker was dirty. `/tmp/reference-i6-exec-4688lrew/baseline.json` records content/identity for 8,936 tracked/nonignored entries, `tracker-before.md` preserves the prior tracker, `index-before` preserves index bytes, and `immutable-baseline.json` fingerprints 153 retained source, metadata, input and historical ledger files. Root/backend/tooling instructions and current source were revalidated. `make task-guide OWNER=module.reference_data` and `make help ALL=1` passed. No owner amendment is needed: RP-REQ-149–152 and the adopted native contracts support the current-only path.

The approved plan extends V-40 with mandatory offline v20.0/v20.1 audit and assembly qualification using existing retained sources, isolated copied source roots, isolated outputs and fresh fixture-only ledgers. Never execute historical assembly input files unchanged: they address tracked fixtures and the historical ledger. No source acquisition, signing, fixture reissue, runtime migration or deployment is included. CLI tests receive an authored owner row. Exact profile bytes are bound in shared admission; synthetic unpinned sources are confined to private test seams. S-30 and S-31 form one releasable cutover. Report/builder identity changes are accounted separately from canonical payload/notice preservation.

I6-04 / S-28 COMPLETE (E-I6-S28): all 13 replacement assertions now execute against current owner schemas/semantics or the live acquisition reader. Eleven are under `TestNativeATTACKProjectionAndLocalReferences_Unit/current_source_contracts`; two are under `TestATTACKSourceRetentionBoundaries/retained_acquisition_inventory_shape`. Original source/ledger tests remain active. Exact retained canonical fixture bytes and their explicit builder/manifest inputs are frozen by the baseline; later qualification compares payload/notices independently of the new executable identity. The old ATT&CK handoff now identifies its proposal-era opening as historical.

`CARTULARY_TEST_RUN_ID=reference-i6-s28-characterization-final CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f` passed with native outer exit 0 (session 38441), two rows passed and retained-secret scan passed. Run root: `.cartulary/test-results/reference-i6-s28-characterization-final`; summary `run-summary.json`. Earlier roots `reference-i6-s28-characterization` and `reference-i6-s28-characterization-fixed` failed with product/test_assertion_failure, outer exit 2, because new synthetic witnesses used float64 instead of the runtime decoder's json.Number and an embedded predicate in a direct SRO. Both test-only defects were corrected; no production schema was weakened. Cleanup reported no failures. `reference-i6-s28-format` passed. RB-004's pre-move assertion gate is satisfied; current CLI/admission/report cases remain obligations of S-30/31. S-29 begins only after this checkpoint.

I6-05 / S-29 COMPLETE (E-I6-S29): the compiler is now `compileProducerSchema` in producer-local `schema.go`; both Windows schema loops and temporary candidate callers use it. Strict JSON, Draft 2020, format assertion and empty external loader behavior are unchanged. New helper assertions reject malformed/duplicate JSON, invalid schemas, HTTPS/file references and bad date-time values; the independent SHA-256 witness now has a current home. `CARTULARY_TEST_RUN_ID=reference-i6-s29-compiler CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c` passed with native outer exit 0 (session 4417), run root `.cartulary/test-results/reference-i6-s29-compiler`, two passing rows and retained-secret scan pass. F-28 is resolved subject to final-source regression. S-30 begins after this checkpoint; report DTO cleanup remains assigned to S-31.

I6-06 / S-30 COMPLETE (E-I6-S30): public audit and assembly share `admitATTACKInputs`, binding exact frozen profile bytes, admitted lock, current policy and compiled native schemas. Public entry points reject synthetic/unbound sources; the private semantic seam is used only by tests. Native records use owner-local `packformat.ValidateATTACKNativeSource`; no new schema set, runtime producer import or general schema service was added. Cancellation is observed at audit entry and within the source-record loop. Proposal CLI/loading/validation execution is removed; the report-only DTO remains until S-31. Candidate tests were removed after their S-28 replacements passed.

The first new CLI row targeted `tools/reference-packs`, which the harness schema intentionally excludes. `reference-i6-s30-format` failed before graph execution with config/configuration_error, and `reference-i6-s30-generate` failed at task-surface generation with the precise selector/package pattern error in `generate/generate-task-surface-and-execution-topology/stderr.log` (outer exit 2). The command was moved to the owner-local offline `producercli` package; the executable is composition-only. No harness allowlist/schema was widened. `make author-test-row-id` derived `module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_2640e6b813` for `./internal/modules/reference_data/producercli:TestATTACKProducerCLI`; authored routing and its generated topology digest now include that row. `reference-i6-s30-generate-fixed` and `reference-i6-s30-format-fixed` passed, outer exit 0.

`CARTULARY_TEST_RUN_ID=reference-i6-s30-current CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f,module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_2640e6b813` passed, outer exit 0 (session 85230), three rows and retained-secret scan passed. `CARTULARY_TEST_RUN_ID=reference-i6-s30-build make reference-pack-producer-build` passed, outer exit 0 (session 83622). Current CLI execution was then exercised on a byte-validated isolated copy of retained v20.1 with `CARTULARY_TEST_RUN_ID=reference-i6-s30-audit make reference-pack-source ACTION=audit SOURCE_DIR=.cartulary/reference-packs/i6-qualification-20261108/20.1/source SOURCE_LOCK=contracts/reference-pack-producers/attack/source-lock.v20.1.json BUILT_AT=2026-11-06T23:02:00Z`; outer exit 0 (session 75026), 31,796 unique records and zero issues. Its intermediate report remains v1 pending S-31 and is not a releasable outcome. Run roots are `.cartulary/test-results/<run-id>` with `run-summary.json` for tests and `<target>/tool-run-summary.json` for standalone tools. S-31 starts after this checkpoint.


I6-07 / S-31 COMPLETE (E-I6-S31): audit now emits `cartulary.attack_source_audit.v2`, binds the exact profile ID/hash, uses `lookup_aliases` and removes the obsolete pack/candidate fields and DTO. The command retains one canonical byte sequence and delivers those same bytes through `--report-file` on success or attributable failure; failure remains unsuccessful. Partial audit counts and ordering are deterministic. All 13 migrated assertions remain active. The ATT&CK embed now explicitly names only current consumed inputs; proposal JSON files remain unchanged as historical evidence. The candidate selector is removed and the new owner-local CLI selector is active. The supporting ATT&CK handoff documents the current contract and historical-input replay restriction.

`CARTULARY_TEST_RUN_ID=reference-i6-s31-reports-final CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_2640e6b813` passed, native outer exit 0 (session 62941), two rows, cleanup without failures and retained-secret scan pass. `reference-i6-s31-format-final` / `make format` passed, native 0 (29893). `reference-i6-s31-generated-check` / `make source-check CHECK=generated` and `reference-i6-s31-schema-check` / `make source-check CHECK=schema` passed, native 0 (38090/96594). Prior `reference-i6-s31-generate` and `reference-i6-s31-reports` summaries report pass; their native session exits were not retained across context continuation, so final qualifying checks above supply independently captured exits. These roots are under `.cartulary/test-results/`. No normative amendment or immutable input rewrite was required. S-32 begins after this checkpoint with finalization, frozen-source validation and mandatory isolated offline qualification.


I6-08 / S-32 qualification repair: `reference-i6-s32-finalize` passed (native 0, session 21931), followed by review/freeze on harness source digest `sha256:4aab6a17085437567e81a3132f3b19597689cfd25e2b243826c70ae88a411a95`. The five-row `reference-i6-s32-final` passed (native 0, session 70952), as did `reference-i6-s32-generated`, `reference-i6-s32-schema`, `reference-i6-s32-build` and both real-source audit commands `reference-i6-s32-audit-20-0` / `20-1`. These passes precede the boundary repair below and are not final closure.

`reference-i6-s32-boundary` failed, native outer 2 (session 95809), product/test_assertion_failure; `unit-logs/row-module.records.architecture.boundary/stderr.log` identifies three violations introduced by the new sibling `producercli` package and six pre-existing import violations in three VERIS fixture-only integration files. Cleanup completed without failure. The producer import/raw-filesystem restrictions are retained unchanged: command composition and its tests move into the existing offline `producer` package, with `RunCommand` called only from the tools entrypoint. The CLI selector is regenerated as `module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_9107ea959a` for `./internal/modules/reference_data/producer:TestATTACKProducerCLI`. This removes the unnecessary sibling package and keeps one enforceable offline boundary.

The exact existing VERIS test files in Incident Bundles, Recovery and Reporting consume fixture signing/testsupport just like already-listed ATT&CK/Entra/Windows tests. Only their three explicit paths are added to the Reference Data facade test allowlist; no production exception, wildcard, test behavior change or owner weakening is introduced. This narrow authored boundary projection repair is necessary for the required check. Finalization, source freeze, the five-row portfolio, boundary/generated/schema checks, final build and offline qualification will be repeated on repaired source. Earlier run roots and builder evidence remain intact.


I6-09 / S-32 repair checkpoint: `reference-i6-s32-repair-format` and `reference-i6-s32-repair-generate` passed, native 0 (sessions 77752/69468). `reference-i6-s32-repair-boundary` passed, native 0 (43851), confirming the existing production import and filesystem restrictions remain satisfied. The first relocated CLI row (`reference-i6-s32-repair-cli`, 19155) failed at compilation, outer 2, normalized infra/preflight_error with missing_selector_result at the row: the standard-library `flag` import collided with the producer's existing `flag` helper. The bounded log is `unit-logs/go-dea23afc276bb8f5-f0ba9d0ab6ad/stderr.log`. Aliasing the standard-library import fixes this move-only defect; no tests executed in that failed invocation and cleanup completed without failure.


I6-10 / S-32 final qualification PASS; handoff documentation check pending. The relocated CLI regression `reference-i6-s32-repair-cli-fixed` passed, native 0 (36786). `reference-i6-s32-finalize-fixed` passed, native 0 (1934), with no unexpected generated changes. Final executable source is frozen at harness digest `sha256:ba7640b08e64c42b4677b92b2a6049acd5f451806fbb50be32a3697ed30fa4b3`; `/tmp/reference-i6-exec-4688lrew/final-source-fixed.json` records the independently inspected path hashes (aggregate `ebc9ce16c8155f08a0a3227170b3a8d799a35f44b3790e6d8b024073193571c4`, a distinct scope/encoding from the harness digest). Only documentation changes follow this freeze.

All commands below ran from the repository root. Each run root is `.cartulary/test-results/<run-id>`. Graph commands retain `run-summary.json`, `run-manifest.json`, exact row results, `cleanup-results.json` and `retained-secret-scan.json`; standalone producer tools retain `<target>/tool-run-summary.json` and referenced logs. All final graph roots have the frozen digest above, no cleanup failures and passing retained-secret scans. Each final producer command's owning summary reports pass and its native outer exit was independently observed as 0.

| Run ID | Exact selection / command (prefix `CARTULARY_TEST_RUN_ID=<run-id>`) | Native session / exit | Purpose / outcome |
| --- | --- | --- | --- |
| `reference-i6-s32-finalize-fixed` | `make agent-finalize` | 1934 / 0 | Schema, tier coverage and generated refresh pass; freeze follows review |
| `reference-i6-s32-final-fixed` | Exact five-row, cache-disabled `make test-slice` command in the qualification portfolio above | 64800 / 0 | All five selected rows pass: source/ledger/CLI, native format, Windows and canonical-artifact lifecycle |
| `reference-i6-s32-boundary-fixed` | `make lint CHECK=backend-boundary` | 82770 / 0 | Runtime cannot import offline producer; exact fixture test exceptions accepted |
| `reference-i6-s32-generated-fixed` | `make source-check CHECK=generated` | 58966 / 0 | Authored routing and generated projection agree |
| `reference-i6-s32-schema-fixed` | `make source-check CHECK=schema` | 31729 / 0 | Machine contracts/routing validate |
| `reference-i6-s32-build-fixed` | `make reference-pack-producer-build` | 17722 / 0 | Independent two-build byte comparison passes |
| `reference-i6-s32-audit-20-0-fixed` | `make reference-pack-source ACTION=audit SOURCE_DIR=.cartulary/reference-packs/i6-qualification-20261108/20.0/source SOURCE_LOCK=contracts/reference-pack-producers/attack/source-lock.v20.0.json BUILT_AT=2026-11-06T23:02:00Z` | 77750 / 0 | Current CLI, report v2 and retained exact v20.0 source |
| `reference-i6-s32-audit-20-1-fixed` | Same public source command with `20.1` in `SOURCE_DIR` and `SOURCE_LOCK` | 49332 / 0 | Current CLI, report v2 and retained exact v20.1 source |
| `reference-i6-s32-assemble-20-0` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/i6-qualification-20261108/20.0/assemble-final.json` | 34099 / 0 | Current audit-to-assembly path; independent Node byte/native witness passes |
| `reference-i6-s32-assemble-20-1` | Same public package command with `20.1` in `INPUT` | 12049 / 0 | Same v20.1 qualification |
| `reference-i6-s32-replay-20-0` | Identical v20.0 package command and inputs | 86432 / 0 | Exact replay; no extra publication allocation or changed output |
| `reference-i6-s32-replay-20-1` | Identical v20.1 package command and inputs | 91780 / 0 | Exact replay; no extra publication allocation or changed output |

Final builder: `61a8c6a45b75da84611eeab2cc2c73dc097cddcf800994394bac56962d9dab09`, 16,352,487 bytes, `cartulary.reference-pack.producer` / `references-v2`, Go `go1.27.3`. Exact executable, identity and two-build equality evidence are retained under `.cartulary/reference-packs/builders/<builder-digest>/`. The independently reproduced executable is the one used by every final source/package command.

### Final offline qualification evidence

The source roots are isolated byte-validated copies of retained v20.0/v20.1 acquisition, lock, source and metadata evidence. There was no network acquisition. Each version's `assemble-final.json` explicitly points to its own `output-final` and fresh `ledger-final`, uses the exact final builder and retains the historical explicit `built_at` and pack version. No historical assembly input was invoked or rewritten. Reports are under `<version>/source/audits/<audit-digest>.json`; logical packages are under `<version>/output-final/logical/<manifest-digest>/`, relative to `.cartulary/reference-packs/i6-qualification-20261108/`.

| Release | Source audit | Canonical output | Exact identities |
| --- | --- | --- | --- |
| v20.0 | 31,765 unique records, 54 collision groups, zero issues | 6,211 objects; 34,604 relationships | Audit `4b7fd0cd79621752bca252f637c06a809ab69e42b97dd8e948515d51cd78a164`; manifest `19572a6b810547b1343c5df2a6148ce6e120377ccbcd327dc2f14fa5157d60d1`; payload `af29e228df3453600b4388969c12c5e4be5f1e5c4f6a45c798207f25cb8e936a` |
| v20.1 | 31,796 unique records, 54 collision groups, zero issues | 6,221 objects; 34,632 relationships | Audit `57464f1e1b9319930a89c1cb3ac8fbf1d1a55ef9916dd8e4b71905e7ae618e1d`; manifest `1bf0aa33740e37c49a539f0cd2402951e1a7422f73d5e547822d36728c2b02c3`; payload `c1f6ddf1a73e2544bb2b568ac0135a355f975087c6dac37291ebbad15aa18e4e` |

Both object/relationship files and the MITRE notice are byte-identical to retained fixtures `20.0/logical/6e94d63980e3816b0edb6303da63a6f54387dfd90764b5f6d58efa51caec876f` and `20.1/logical/797de8a4133f5bd4b7ff79a26ce57efcdeebc43556033ea3f7ff685004252865`, under `fixtures/reference-packs/attack-native-v2/`. The framed payload identities above are unchanged. Manifest field comparison finds exactly two differences: `builder` and `pack_release_sequence` (historical 6/7 becomes 1 in each isolated fresh ledger). The historical builder has a different ID/version/digest; these differences are explicitly preserved, never forged away. All other manifest fields match, including profile digest, source artifacts, license and canonical file inventory. Current build reports bind the exact v2 audit hash shown above.

The existing public package target's independent Node witness verifies canonical framing, member hashes, original native source, relationships and exact notice bytes. Replay leaves all 24 output/ledger files unchanged, including build reports, manifests, witnesses and publication records. File comparisons are retained in `/tmp/reference-i6-exec-4688lrew/qualification-comparison.json`, `qualification-before-replay.json` and `qualification-replay.json`; audit inspection is in `final-audits.json`. This is isolated producer qualification, not signature renewal, deployment trust or a release claim. No signed artifact was reissued.

### Final obligation mapping and handoff

| Obligation | Closing evidence |
| --- | --- |
| V-36 | E-I6-S28 maps and passes all 13 assertions before retirement; final native/source rows retain those assertions |
| V-37 | Final Windows row plus private strict-schema/loader/date-time cases pass |
| V-38 | Final producer and CLI rows reject removed/unbound inputs, test success/failure byte delivery and cancellation; final real CLI audits and reproducible build pass |
| V-39 | Explicit embed, absent proposal executable symbols/selector, generated/schema checks; immutable input/evidence comparison |
| V-40 | Five final rows plus both current-source audit/assembly/replay paths, independent package witnesses, byte equality and accounted manifest provenance differences |
| V-41 | Final boundary/generated/schema checks, preserved offline import prohibition, frozen-source and diff review; final Markdown check completes documentation gate |

F-26–29 are implemented and pass executable qualification. No adopted specification amendment was needed; current owner requirements already justify exact source admission and current-only packaging. The supporting ATT&CK handoff distinguishes its historical sections from the current producer contract. No runtime schema/API/database conversion or deployment change is required. Consumers of the offline report must switch to v2 and the new fields; callers must stop using `--candidate`. Roll back producer command, admission/report changes, tests and routing together if necessary; do not reissue immutable historical artifacts or restore only one half of S-30/31.

Preservation inspection found exactly 23 changed/added/deleted repository paths, all within the approved producer/test/document/routing work plus the explained three-entry boundary repair. Only the topology render index is regenerated; immutable contract JSON, canonical fixtures, dependency locks, source profiles and policies are unchanged. The final target contains 265 Go files (109 tests) in the same eight directories. All 153 retained historical files and their path set are unchanged. Final source path hashes still match the freeze. HEAD and staged entries are unchanged; raw index bytes changed through stat-cache refresh, so byte-identical index preservation is explicitly not claimed. `git diff --cached --exit-code` and `git diff --check` pass. Details: `/tmp/reference-i6-exec-4688lrew/source-review-fixed.json`, `index-comparison.json` and `preservation-final.json`.

Optional finalizer canonical-evidence and scheduler-drift maintenance report `results-dir-not-provided`; no compatible full warm-check root existed and no broad run was created solely for maintenance. No required scoped executable check is skipped. Full-owner/CI, UI/browser, source acquisition, signing and release certification were outside this bounded effort. All failures, repairs and superseded source evidence remain recorded above. Remaining S-32 work is Markdown validation and the completion checkpoint.


I6-11 / S-32 validation and handoff COMPLETE (E-I6-S32): `CARTULARY_TEST_RUN_ID=reference-i6-s32-handoff make lint-markdown` passed, native outer exit 0 (session 27027), with summary `.cartulary/test-results/reference-i6-s32-handoff/adhoc/lint-markdown/tool-run-summary.json`. V-36–41 are closed on the final source and qualification evidence above. F-26–29 and RB-004 are resolved; T-029–033 / S-28–32 are DONE. Every workstream has its completion checkpoint before the next began. The boundary and compilation failures were diagnosed, repaired and requalified; none remains unresolved. No required scoped check is outstanding.

The I6 final completion transcription used `CARTULARY_TEST_RUN_ID=reference-i6-s32-handoff-final make lint-markdown`; its result is retained at `.cartulary/test-results/reference-i6-s32-handoff-final/adhoc/lint-markdown/tool-run-summary.json`. Final source/index-entry/immutable-evidence and whitespace comparisons are retained in `/tmp/reference-i6-exec-4688lrew/final-checks.json`. Within I6, no implementation edit followed the final source freeze, and no repository edit followed the successful transcription check. I7 below supplements that completed work. Historical proposal artifacts confer no executable compatibility requirement.

## 21. I7 validation-gap remediation and handoff

### Authority, baseline and scope

I7-00 / S-33 IN_PROGRESS (2026-11-08 UTC): the user authorized implementation of the complete supplemental plan. RP-REQ-149–152 and RP-REQ-265–269 already govern provenance, exact source preservation and native field validation; no normative amendment, projection change or new source profile is required. Review found two validation gaps and one handoff inconsistency, with no confirmed production-behavior regression. I6's original passes and completion checkpoints remain historical evidence for their stated coverage; V-38's stronger failure-path claims require the new witnesses below before supplemental closure.

The dirty-tree baseline is `main` at `a8f2fe2d84eecf88ced265b34327e0a8c7d4c8df`. `/tmp/reference-i7-exec-otf0_hif/` contains `baseline.json` (8,942 tracked/nonignored regular-file hashes), `status-before`, `head-before`, `staged-before`, `tracker-before.md` and `immutable-baseline.json` (364 historical evidence files, including all 153 previously preserved files, I6 qualification and existing producer builds). Preserve existing work and staged entries. Root/backend/tooling instructions and test-operations guidance were read; `make task-guide OWNER=module.reference_data` and `make help ALL=1` passed (native session 68564, exit 0). On this date the authored catalog `tools/test_families/module.reference_data.json` contains 93 rows, including 43 unit-family rows; the catalog, not repeated prose counts, remains the inventory authority.

### Findings and durable remediation

| Finding / areas | Remediation and rationale / long-term benefit | Compatibility and unresolved risk | Completion gate / status |
| --- | --- | --- | --- |
| F-30 / tests, documentation | Add valid/invalid `created` and 64/65-contributor fixtures on the enterprise-only tool before hashes/acquisition. Audit through `auditATTACKSources`, with otherwise identical identity/version/reference fields. Isolates the audit's native-schema guard from field-policy and semantic checks, protecting the connection rather than only the validator. | No adopted input/schema/report change. Without this, native validation can be bypassed while unrelated negative fixtures still pass. | V-42 / PASS on final source |
| F-31 / implementation, tests, documentation | Make owner-local `RunCommand` return exit status; private dispatcher takes one explicit audit function and public entry always binds real `Audit`. Executable only calls `os.Exit`. Test-only child mode runs real synthetic audits through shared dispatch/delivery/status policy, with bounded joined subprocesses and parent-owned directories. One command owner prevents duplicated exit policy and helper-only false assurance. | Atomically migrate the internal Go return type from `error` to `int`; no alias, external flag/output/report/exit change, public injection option or mutable override. Without this, early audit-error returns can silently discard failed reports or return success. | V-43 / PASS on final source |
| F-32 / documentation | Reconcile §1/§8 and qualify historical I6 summaries; retain all commands, failures and hashes. Record this iteration's dependencies and evidence separately. Consistent current instructions make future selection and handoff reliable. | No runtime impact or executable prose dependency. Leaving stale future-work/count/completeness statements invites repeated work or unsupported claims. | V-44 / PASS; E-I7-S36 |

### Workstreams, gates and sequencing

| Slice / dependency | Work, principal risk and rollback boundary | Exit criterion / status |
| --- | --- | --- |
| S-33 / current checkout | Baseline preservation, authority reconciliation and F-32 plan. Risk: overwrite history or overstate evidence. | Updated tracker, explicit gates and Markdown pass / DONE, E-I7-S33 |
| S-34 / S-33 | F-30 audit regressions. Risk: independent rejection masks the native guard. | V-42 focused source row passes; checkpoint before S-35 / DONE, E-I7-S34 |
| S-35 / S-34 | F-31 dispatcher/status and subprocess witnesses. Risk: injection escapes test code or duplicates production logic. Roll back command and callers together. | V-43 focused CLI row passes; checkpoint before S-36 / DONE, E-I7-S35 |
| S-36 / S-35 | Finalize, inspect, freeze, qualify both real releases and hand off. Risk: incompatible-source evidence or historical artifact rewrites. | V-42–45 closed, preservation and final transcription pass / DONE, E-I7-S36 |

| Gate | Required evidence |
| --- | --- |
| V-42 | Valid timestamp and 64 contributors pass without issues; invalid timestamp and 65 contributors error with `source_audit_failed` and only the exact tool/enterprise `/objects/2` `native_source_invalid` issue. Repeated canonical report bytes match. Existing source-contract row includes all cases; public admission remains strict. |
| V-43 | Shared dispatcher child exits 0 on valid input, 1 on native-invalid input; retained/requested canonical reports match with expected provenance. Failed replay retains identical identity/bytes. Invalid destination still retains the audit and fails. Public synthetic/unbound locks create no report; removed flags fail and help succeeds. All under the existing CLI row. Test-only controls are absent from production. Controlled process failure evidence is distinct from fixed-auditor real executable qualification. |
| V-44 | Current summaries/statuses/row accounting/evidence agree; historical statements are identified; Markdown passes. No product test or generator consumes this prose. |
| V-45 | Final evidence binds the frozen source after `make agent-finalize`: fresh `reference-i7-final`, cache off, the exact five rows from §20; backend-boundary lint, generated/schema checks, reproducible producer build; isolated v20.0/v20.1 audit/assembly/replay with new builder and fresh fixture ledgers; payload/notice equality, only accounted manifest builder/sequence differences; historical bytes and staged state preserved; cleanup/secret scans, native exits, final Markdown/whitespace/transcription checks recorded. |

Focused checks follow each implementation workstream; diagnose any failure before retrying. Final qualification uses public Make targets and retained sources only, with no acquisition or signing. Full CI/full-owner suites, UI/browser, deployment and runtime/data migration are outside scope. Optional retained-run maintenance is skipped with an explicit reason when no compatible full warm-check root exists; no broad run is created just to obtain one. Update this tracker after each completed workstream before starting its successor. No supplemental gate is closed by I6 evidence alone.

I7-01 / S-33 COMPLETE (E-I7-S33): the current scope and §8 now distinguish completed I6 execution from historical planning and supplemental acceptance. F-32's stale wording is reconciled; V-44 remains subject to final handoff consistency. `CARTULARY_TEST_RUN_ID=reference-i7-s33-markdown make lint-markdown` passed, native exit 0 (session 29441), summary `.cartulary/test-results/reference-i7-s33-markdown/adhoc/lint-markdown/tool-run-summary.json`. Baseline preservation and owner reconciliation above are complete. S-34 starts after this checkpoint.

I7-02 / S-34 COMPLETE (E-I7-S34): all four native-boundary cases now run under `TestATTACKSourceContractClosure/native_audit_enforcement`. The enterprise-only tool is mutated before acquisition hashing; unrelated identity/version/reference fields are unchanged. Three audits per case assert canonical byte equality; valid controls have no issues, invalid controls have exactly `native_source_invalid:tool--22222222-2222-4222-8222-222222222222:enterprise-attack.json:/objects/2` and fail. `reference-i7-s34-format` (`make format`, session 51876) and `CARTULARY_TEST_RUN_ID=reference-i7-s34-source CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c` (session 55282) passed, native exits 0. Evidence root `.cartulary/test-results/reference-i7-s34-source` has a passing run summary, one passing row, completed cleanup and passing retained-secret scan. F-30 / V-42 passes focused validation, subject to final-source confirmation. S-35 starts after this checkpoint.

I7-03 / S-35 COMPLETE (E-I7-S35): `RunCommand` now returns an integer and owns diagnostics/exit mapping; the thin executable calls `os.Exit`. Public dispatch always binds `Audit`; only a private typed function parameter enables package tests. The test-only `TestMain` runs bounded joined children with actual synthetic audits, proving successful/failed delivery, canonical retained/requested bytes, repeat identity, public synthetic/unbound rejection, removed-flag rejection and successful help. Destination failures use fresh retained roots for both successful and failed audits, proving retention happens before the write error and preserving the original audit diagnostic. All child controls and synthetic admission remain in `_test.go`. `reference-i7-s35-format` (`make format`, session 8418) and `CARTULARY_TEST_RUN_ID=reference-i7-s35-cli CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_9107ea959a` (session 84205) passed, native exits 0. `.cartulary/test-results/reference-i7-s35-cli` contains the passing row/run summary, completed cleanup and passing retained-secret scan. F-31 / V-43 passes focused validation, subject to final-source confirmation. No routing or boundary exception changed. Additional historical summary labels in §§2/4/6/7/9/10/11/12 were reconciled under F-32. S-36 starts after this checkpoint.

### Final source and validation evidence

I7-04 / S-36 executable qualification PASS; final documentation gate pending. `make agent-finalize` passed with no unexpected generated changes. The I7 diff from its dirty-tree baseline contains exactly six paths: this tracker, `producer/audit_report_test.go`, `producer/command.go`, `producer/command_test.go`, `producer/source_test.go` and `tools/reference-packs/main.go` (producer paths under the target). All earlier work remains intact; no authored routing, boundary exception, normative specification, frozen input, schema projection or dependency lock changed during I7. The final source freeze is harness digest `sha256:73e284201f7e19b768cdd496e940f2794789edba88e1ca14b6f3babc826086e0`. `/tmp/reference-i7-exec-otf0_hif/final-source.json` records 8,559 non-Markdown/non-doc source hashes; its separate inspection aggregate is `54d07296e6caa79f40dd70e4ddcc00a4f6bf1af68cb4027b130a7ef114bab24d`. This is preservation inspection, not a replacement harness identity or an executable dependency on prose. Only tracker edits follow the freeze.

All commands below ran from the repository root, with prefix `CARTULARY_TEST_RUN_ID=<run-id>`. Each root is `.cartulary/test-results/<run-id>`. Graph roots retain `run-manifest.json`, `run-summary.json`, exact `rows/` results where selected, `cleanup-results.json` and `retained-secret-scan.json`. All final graph manifests carry the frozen source digest; all summaries pass, all cleanup outcomes are completed with no failures, and all retained-secret scans pass. Standalone producer roots retain `<target>/tool-run-summary.json` and referenced command logs; their native outer exits and summaries agree. No I7 validation command failed or required a repair/rerun.

| Run ID | Command / exact selection | Native session / exit | Result and obligation |
| --- | --- | --- | --- |
| `reference-i7-s36-finalize` | `make agent-finalize` | 66690 / 0 | Schema, coverage and generated refresh pass; source freeze follows inspection |
| `reference-i7-final` | Cache-disabled exact five-row `make test-slice` command below | 54853 / 0 | All five rows passed; five cache bypasses, no hits; V-42/43 and regression portfolio |
| `reference-i7-s36-boundary` | `make lint CHECK=backend-boundary` | 5613 / 0 | Existing owner/import/filesystem boundaries pass |
| `reference-i7-s36-generated` | `make source-check CHECK=generated` | 16279 / 0 | Generated projections agree without I7 routing edits |
| `reference-i7-s36-schema` | `make source-check CHECK=schema` | 30228 / 0 | Machine contracts pass |
| `reference-i7-s36-build` | `make reference-pack-producer-build` | 91797 / 0 | Two builds are byte identical; test child marker absent from executable |
| `reference-i7-s36-audit-20-0` | `make reference-pack-source ACTION=audit SOURCE_DIR=.cartulary/reference-packs/i7-qualification-20261108/20.0/source SOURCE_LOCK=contracts/reference-pack-producers/attack/source-lock.v20.0.json BUILT_AT=2026-11-06T23:02:00Z` | 69808 / 0 | Fixed production auditor, retained v20.0 source |
| `reference-i7-s36-audit-20-1` | Same source command with `20.1` in directory and lock | 49199 / 0 | Fixed production auditor, retained v20.1 source |
| `reference-i7-s36-assemble-20-0` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/i7-qualification-20261108/20.0/assemble.json` | 42239 / 0 | Assembly and independent Node native/byte witness pass |
| `reference-i7-s36-assemble-20-1` | Same package command with `20.1` in input | 47300 / 0 | Same v20.1 qualification |
| `reference-i7-s36-replay-20-0` | Identical v20.0 package command/input | 38904 / 0 | Exact replay preserves bytes and ledger allocation |
| `reference-i7-s36-replay-20-1` | Identical v20.1 package command/input | 72557 / 0 | Exact replay preserves bytes and ledger allocation |

The exact final portfolio was:

```sh
CARTULARY_TEST_RUN_ID=reference-i7-final CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_9107ea959a,module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f,module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.integration.attack_producer_artifacts_preserve_independent_a_67e9ef139a
```

The reproduced builder is `406646707d1310fe6f1dc2830392f550cc871d2bb7725b0fa6d546a1842de94a`, 16,356,788 bytes, ID `cartulary.reference-pack.producer`, version `references-v2`, Go `go1.27.3`. Its executable, identity and reproduction witness are under `.cartulary/reference-packs/builders/<digest>/`; all final offline inputs bind that digest. The compiled production executable does not contain the test child-mode marker. This supplements the source review showing the child controls and synthetic admission exist only in `_test.go`.

### Offline release comparisons and preservation

Qualification root: `.cartulary/reference-packs/i7-qualification-20261108/`, with separate `20.0` and `20.1` source/output/ledger roots. Eighteen source/acquisition/metadata files were copied from retained I6 roots and checked byte-for-byte; no old audits were copied, no acquisition occurred, and no historical ledger was used. Each input uses timestamp `2026-11-06T23:02:00Z`, unchanged pack version `attack-20.0.4` or `attack-20.1.4`, fixture repository `fixture.attack.native.v2`, restricted redistribution and a fresh sequence-1 ledger. No signature was reissued.

| Release | Audit SHA-256 | Manifest SHA-256 | Payload SHA-256 | Counts / status |
| --- | --- | --- | --- | --- |
| 20.0 | `1a11c0501191da1f8541721b25915c4f96f198efcaa707658780e10be0c9383a` | `45ea4eff8ca4a7b2667140663798fe0c5dbe8b28f9924d80c185887999c39a9a` | `af29e228df3453600b4388969c12c5e4be5f1e5c4f6a45c798207f25cb8e936a` | 31,765 unique records; 54 collisions; 0 issues; 6,211 objects / 34,604 relationships |
| 20.1 | `4e6cab489ae2be2075354254048ccd7eaa05fbdf43d4a1d4911138ed1ac6f946` | `baa8400c9d694c3d12d80cb69fed32c815933331b9873d8a496cc0bd269a3afa` | `c1f6ddf1a73e2544bb2b568ac0135a355f975087c6dac37291ebbad15aa18e4e` | 31,796 unique records; 54 collisions; 0 issues; 6,221 objects / 34,632 relationships |

Payload and notice paths/bytes exactly equal the frozen fixtures identified in §20. Manifests differ only in the accounted builder identity and historical sequence 6/7 versus fresh sequence 1. Audit reports differ from I6 only in builder identity. Independent witness digests are `9f78c5432d42e90e60db649875856615109aca071c8bb00f462fb120e80eac10` (20.0) and `bbab8727ecddda49663e88143809500e83eacecaf49d624b7c58738c703b0a16` (20.1). Both identical-input replays preserve all 24 output/ledger files plus both retained audit reports, including their path sets.

Human inspection records in `/tmp/reference-i7-exec-otf0_hif/`: `qualification-source-copies.json`, `builder-review.json`, `qualification-comparison.json`, `qualification-before-replay.json`, `qualification-replay.json`, `standalone-evidence.json`, `source-freeze-review.json` and `preservation-final.json`. All 364 baseline historical evidence files are unchanged; the original 153-file historical path set is also unchanged. Frozen contracts/profiles/policies/locks, canonical fixtures, signed artifacts and dependency locks have no I7 delta. HEAD and staged entries are unchanged. Raw index byte equality is not claimed. `git diff --check` passes; final transcription will repeat source/preservation/whitespace inspection.

### Acceptance limits and handoff

V-42 now isolates native timestamp/array-bound rejection through the audit. V-43 now observes actual process exits and report delivery through the shared command code, including failed replay, destination failures and public admission rejection. These controlled process witnesses use the compiled Go test executable and a private synthetic auditor; they are not black-box failure qualification of the two valid accepted release corpora. The separately reproduced production executable always uses real `Audit` and passed both full offline release paths. No public override was added.

The only interface migration is the internal Go `RunCommand` result from `error` to `int`, with all repository callers migrated atomically. External flags, diagnostics, report v2 and exit behavior are unchanged. No normative specification or runtime/data migration is required. Roll back the command owner, executable caller and command tests together if necessary; preserve immutable historical artifacts.

Optional finalizer `canonical_evidence_validation` and `scheduler_drift_validation` were skipped with `results-dir-not-provided`: no compatible full warm-check root was available, and no broad run was launched only for maintenance. No required scoped executable check is skipped. Full CI/full-owner suites, UI/browser, deployment and release certification remain outside scope. At I7-04 only V-44's Markdown/transcription gate and the S-36 completion checkpoint remained; their completion is recorded below.

I7-05 / S-36 COMPLETE (E-I7-S36): `CARTULARY_TEST_RUN_ID=reference-i7-s36-handoff make lint-markdown` passed, native exit 0 (session 33058), summary `.cartulary/test-results/reference-i7-s36-handoff/adhoc/lint-markdown/tool-run-summary.json`. Current-state summaries and statuses were reconciled with the preserved historical record and final evidence. V-42–45 and F-30–32 are closed; S-33–36 are DONE, with a completion checkpoint after each workstream and before its successor. No required scoped validation or implementation uncertainty remains unresolved.

The final completion-transcription command is `CARTULARY_TEST_RUN_ID=reference-i7-s36-handoff-final make lint-markdown`; its terminal result is retained at `.cartulary/test-results/reference-i7-s36-handoff-final/adhoc/lint-markdown/tool-run-summary.json`. The handoff requires that command's native exit 0 and the final source/immutable-evidence/staged-entry/whitespace comparison in `/tmp/reference-i7-exec-otf0_hif/final-checks.json`. No source edit follows the freeze and no repository edit follows the successful final transcription check. The controlled process witness's limitation, unchanged external contract, internal signature migration, exact qualification roots and optional maintenance skips above are part of this closure.

## 22. I8 obsolete Base tooling and dead fixture branches

### Current authority, scope and baseline

I8 began as a document-only plan dated 2026-11-08 America/New_York. The user subsequently explicitly requested implementation of S-37–40, including sequential tracker checkpoints and final qualification. The execution record below supersedes planning-only authorization/readiness statements for this current invocation. The goal remains bounded production hardening through removal of obsolete execution and unused test branches, not whole-subsystem certification.

The adopted/current Reference Pack NLSpec is 0.4.0 / major 3. Section 1 admits only current pack contract v2, distinguishes retained predecessor evidence from deployed state, and prohibits compatibility readers, implicit reset and data conversion. RP-REQ-059/060 require exact release bindings and verified Base bytes before readiness. Core 00 records the native ATT&CK/Entra/VERIS cutovers and their owner boundaries. These requirements remain intact; they do not require a reusable historical reissue command. No owner amendment, new schema, HTTP change or migration is needed for the planned retirement. A `v1` suffix on a still-current manifest, source profile or binding does not make it obsolete.

E-I8-01 records clean `main` at `e2291271a502033b85c5e2d6aace2609bc8c8210`, 8,942 tracked/nonignored entries and the 265-file target inventory. The tracker began at SHA-256 `ab41f0a7ea4d9e12ad6b71eef088e2eab93aa42940947816ccf90f2c5c899d04`. Its supplied and canonical identities match; no target/output path component is a symlink and the tracker is a single-link regular file. The framework and applicable root/backend/tooling instructions remain available. The target's historical I6/I7 counts now match the current enumeration, but historical execution artifacts are not requalified here.

### Removal decisions and preservation map

The decision/gate tables below retain the approved plan. Current execution status and completion evidence are in §9 and the implementation checkpoints at the end of this section; V-49 and the planning-document session remain historical.

| Finding / decision | Current evidence and future action | Compatibility / future-state benefit | Acceptance |
| --- | --- | --- | --- |
| F-33 / retire one-shot Base cutover | `producer/base.go` reads `previous_release`, fixes pack version `base-2026-11-06.1`, sequence 2 and release ID `cartulary.base_registries.2026-11-06.v2`; `producer/command.go` dispatches its schema and the independent witness admits it. Remove this entire executable path and the input type. | Removes a dated release operation and its prior-release parsing/unchecked shape assertions without extending a compatibility contract. External users of this offline input must stop invoking it; no alias, conversion or replacement builder. | S-38 / V-46: no executable reissue path; existing unsupported-schema diagnostic and status; no requested output or success report |
| F-34 / simplify obsolete-format fixture | Only the signed/unsigned boolean varies at the sole `retiredReferencePackBundle` call. Replace `retiredBundleOptions` with `retiredReferencePackBundle(t testing.TB, signed bool)`; make the existing identity/payload defaults fixed helper data. Remove unused checksum/signature corruption, alternate paths, omitted-payload and extra-member options. | Keeps the meaningful security regression while shrinking test support and eliminating unexercised configuration branches. This helper stays test-only. | S-39 / V-47: both cases reach the live verifier and still fail with `bundle_hint_invalid`; `addZipFile` and its other callers remain |
| Current Base release / keep | Runtime `loadBuiltinRelease` reads generated `builtins/release.v2.json`, independently of `ReissueBase`; current consumer tests verify exact bindings and the required registry vocabulary. | Keeps required readiness and exact historical identities without preserving the old producer operation. | C-12/15; V-48: checked-in current release, witness/source-profile and historical artifacts retain exact bytes |
| Other current producers and meaningful negative evidence / keep | Current ATT&CK/Windows/Entra/VERIS packaging, fixture signing, schema validators, retained readers and fallback retain live callers or owner requirements. | Avoids treating a test-only caller, a version suffix or old fixture date as dead code. | No unrelated producer/package/format/UI/lifecycle change; existing selectors remain |

This decision supersedes only the Base-reissue portion of I2's §16 intentional/no_action conclusion and I6's corresponding retention wording in §5. Their source snapshots, rationale, commands, failures and completion evidence remain historical. The current independent witness lists Base reissue as admitted but has no dedicated witness-producing branch for it; remove that admission entry together with the producer route rather than adding a new witness for the retired operation. Preserve all other witness input schemas and branches.

The source search boundary was tracked/nonignored repository content, including Go declarations/callers and tool schema dispatch. `ReissueBase`, `BaseReissueInput` and `base_registry_reissue_input` lead to the producer file, command branch and witness allowlist. Textual absence is not an external-consumer guarantee. Current release and historical artifact bytes remain evidence, not code scheduled for deletion. No archived command, signed fixture, historical report or original source profile is rewritten to imply it came from the new executable.

### Workstream gates and assertion destinations

Section 6 is the authoritative dependency graph; §7 records per-slice readiness and rollback. No dependency is satisfied by completing this document. Each future workstream records its checkpoint before the successor begins.

| Slice | Required entry / work | Binary exit and evidence |
| --- | --- | --- |
| S-37 | Explicit user implementation authorization; revalidate current checkout and E-I8-01–05. Recheck callers/build tags and each fixture option; inspect current selectors; capture immutable current/historical artifact fingerprints. Run the three-row baseline below. | E-I8-S37: baseline rows pass with compatible source, all useful assertions have named destinations, no unresolved caller or ownership question; resolves RB-005 |
| S-38 | S-37 exit. Remove the producer function/input type, its CLI branch and witness allowlist entry together. Extend the existing CLI test using the real production dispatch child. | E-I8-S38 / V-46: retired input exits 1, reports `unsupported package input schema`, writes no success report and creates no requested release/output parent; CLI row passes and removal search has no executable references |
| S-39 | S-38 exit. Simplify the helper to its signed boolean; preserve the two live rejection cases and unrelated ZIP-helper callers. | E-I8-S39 / V-47: both variants still reject with `bundle_hint_invalid`; unused options/branches absent; verifier row passes |
| S-40 | S-39 exit. Run `make agent-finalize`, inspect any generated diff, freeze executable source, qualify the exact scoped portfolio and record final preservation/handoff. | E-I8-S40 / V-48: all required checks pass on compatible final source; immutable bytes and existing selectors retained; findings close only here |

| Assertion / preservation obligation | Destination / constraint |
| --- | --- |
| Retired offline input rejected at dispatch | Add a subtest to existing `TestATTACKProducerCLI`, using `runProducerCLIChild` in production mode, parent-owned temporary input/output paths and the existing bounded joined child. Use the retired schema with a nonexistent previous-release canary and absent output directory; require the unsupported-schema diagnostic, native exit 1, empty stdout and unchanged/absent requested output. No public test override or new selector. |
| Existing CLI help, invalid arguments and audit report delivery | Preserve current `TestATTACKProducerCLI` cases and their real dispatch/status policy; no duplicated command implementation. |
| Signed and unsigned checksum-labelled formats stay untrusted | Preserve both iterations in `TestVerifierAcceptsLocalBundleAndRejectsFailures_Unit`; the `Signed` flag describes the obsolete checksum label, not a valid cryptographic signature. Keep the existing winning rejection. |
| ZIP construction used by current malformed/archive-limit cases | Retain `addZipFile`; recheck its callers before simplifying its containing helper file. |
| Current Base release binding and vocabulary | Preserve `TestCanonicalBaseReleaseBindingsAndConsumerEvaluation_Unit`, including exact current release loading and all nine required indicator types. |
| Artifact immutability and offline/runtime separation | Compare current built-in contract artifacts, canonical/signed fixtures and explicitly selected retained evidence against S-37 fingerprints; preserve producer import exclusion. Human preservation inspection is not a new executable dependency on specification prose. |

### Validation obligations and exact selections

| ID / layer | Findings, contracts and slices | Stage / command | Environment / expected result | Failure consequence |
| --- | --- | --- | --- | --- |
| V-46 / offline retirement | F-33; C-18, preserve C-12/15; S-37/38 | Baseline then post-slice: existing CLI row from the portfolio below; inspect function/dispatch/witness removal | Machine JSON input, no services/network; bounded production-dispatch child; exact unsupported-schema/no-output assertion | Block S-38 completion on a missed route, wrong diagnostic, success status or output effect |
| V-47 / test preservation | F-34; C-13/19; S-37/39 | Baseline then post-slice: existing verifier row below | Canonical/local fixtures and the two minimal predecessor variants; no live website or prose | Block S-39 if a useful assertion or ZIP caller is lost or rejection precedence changes |
| V-48 / scoped qualification | F-33/34; C-12/13/15/18/19; S-37/40 | S-37: portfolio baseline. S-40: `make agent-finalize`, source review/freeze, final portfolio, `make lint CHECK=backend-boundary`, `make source-check CHECK=generated`, `make source-check CHECK=schema`, `make reference-pack-producer-build`, `make lint-markdown`, `git diff --check` | Fresh explicit run IDs, cache off for rows, installed pinned tools; no product services needed by selected rows. Retain native exits, owning summaries/manifests, exact source identity and available cleanup/secret-scan results. | Any required failure or immutable change blocks closure; investigate before retrying or broadening |
| V-49 / current document integrity | T-034; this tracker only | Current session: admitted `make lint-markdown` command below, `git diff --check`, current-section review and content/index/identity comparison | Installed-only prerequisites, network-disabled Corepack, external result/cache paths, `fix:false`; only tracker may change | Document update remains incomplete if lint, consistency or preservation fails |

At the start of a later implementation task, use `make task-guide OWNER=module.reference_data` to revalidate routing. Existing authored rows select `TestATTACKProducerCLI`, the verifier group containing both predecessor assertions, and the Base-consumer group respectively. Their presence is discovery evidence, not a pass. Preserve these selectors; the plan adds no harness framework or routing row. If a later necessary routing change is established, update authored inputs and regenerate through Make; never hand-edit generated projections.

The S-37 baseline and S-40 final portfolio use the following command with distinct fresh run IDs (`reference-i8-baseline` and `reference-i8-final` are proposed names; choose a fresh suffix if already occupied):

```sh
CARTULARY_TEST_RUN_ID=reference-i8-baseline CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_9107ea959a,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.unit.canonical_consumers
```

Post-S-38 uses only the first row; post-S-39 uses only the second. The final portfolio follows finalization and source freeze. The producer build proves reproducibility of the changed executable; it does not reissue existing Base bytes or qualify a new external-data release. Integration, browser/UI, full-owner/full-CI suites, source acquisition, real-corpus assembly/replay, signing, deployment and release certification are not required for this bounded deletion/test-helper change. Broaden only if actual edits or failures expose additional ownership/risk. No full warm-check run is created solely for optional retained-run maintenance.

### I8 planning-document evidence and session handoff

| Evidence ID | Source / exact inspection | Baseline / time | Observation / retained material | Applicability |
| --- | --- | --- | --- | --- |
| E-I8-01 | Read-only Git status/HEAD/staged enumeration; file hash/identity and output-path admission | Clean `main`, `e2291271a502033b85c5e2d6aace2609bc8c8210`; 2026-11-08 | `/tmp/reference-i8-docs-e5p53yna/`: `baseline.json`, `ignored-baseline.json`, `target-inventory.json`, `tracker-before.md`, `inspection.py`; 8,942 repository entries, 265 target entries | Preservation/inspection only, not harness conformance or runtime evidence; relevant ignored scope is explicit in §1 |
| E-I8-02 | Reference Pack NLSpec front matter/§1/RP-REQ-059/060; Core 00 adoption/precedence; domain owner navigation; root/backend/tooling instructions and framework | Same source baseline; planning reads revalidated before update | Current-only pack admission and exact release verification; no requirement for the one-shot builder; target/output identity established | Owner authority separate from implementation success; historical producer retention decision superseded only as stated |
| E-I8-03 | `producer/base.go`, `producer/command.go`, `producer/command_test.go`; `tools/reference-packs/main.go`, `package.sh`, `witness.mjs`; repository symbol/schema searches | Same baseline | Fixed release identity and sequence, only CLI call path, witness admission, existing subprocess harness and unsupported-schema branch | Sufficient bounded movement/removal evidence; external callers not exhaustively known; no command executed |
| E-I8-04 | `retired_format_fixture_test.go`, relevant verifier assertions/ZIP callers, `base_release.go`, `consumer_test.go` | Same baseline | One fixture caller with two variants; unused options; runtime current release read independent of producer; Base binding/vocabulary assertions | Supports F-34 and preservation destinations; test presence is not execution |
| E-I8-05 | Authored Reference Data test-family rows; Make/generated recipes, workspace admission, installed-readiness, step/static-cache and Markdown runners/configuration | Same baseline | Exact three-row future portfolio; Markdown `fix:false`; installed-only readiness, redirected results/cache and existing workspace lock admit the document check | Discovery only; product checks/builds/finalization intentionally skipped in this invocation |
| E-I8-06 | Admitted `make lint-markdown`, native session 62440 / exit 0; `git diff --check`; historical text and full baseline comparison | Same executable baseline; tracker-only update; 2026-11-08 | Run `/tmp/reference-i8-docs-e5p53yna/results/reference-i8-docs`; `document-review.json` and `preservation-checkpoint.json` in the comparison directory; final transcription artifacts named below | Documentation/preservation pass only; no product check executed; original §§13–21 preserved byte-for-byte |

| Area / session | Current state / files inspected or touched | Revalidated or superseded conclusion | Next action / boundary |
| --- | --- | --- | --- |
| Scope and authority / I8-00 | E-I8-01/02; only tracker write admitted | Previous implementation permissions and passes remain historical | Complete V-49; future production work requires later authorization |
| Backend boundary / I8-01 | E-I8-03/04; producer removal and owner-local test simplification planned | Supersedes the named one-shot reissue no-action decision; keeps current release loader and subsystem boundaries | S-37 caller/assertion/immutable baseline; no general package extraction |
| Contract/codegen / I8-01 | C-12/13/15/18/19; no authored schema/projection edit | Only offline reissue input is scheduled for removal; no runtime wire/data change | Preserve artifact bytes; future finalization review cannot silently accept an identity rewrite |
| Tests/harness / I8-01 | E-I8-05; existing three rows and V-46–49 | I6/I7 passes do not satisfy I8 baseline; no test or harness reads this tracker as requirements | Run only V-49 now; future rows use machine fixtures/current production paths |
| Security/authorization / I8-01 | Exact Base binding and unsigned/signed predecessor rejection assertions inspected | Historical formats remain untrusted; fixture signing and real current validators stay | Preserve winning rejection, no output on retired command and offline producer exclusion |
| Risks/next session / I8-01 | RB-005; S-38–40 predecessor gates; no unresolved owner decision | Planning complete for bounded scope; implementation incomplete/unexecuted | Revalidate affected evidence and run S-37 after later authorization |

Frontend/rendered review is not applicable: neither the current document edit nor the planned production slices change UI, selectors or browser behavior. No new public HTTP contract or codegen ownership is introduced.

I8-00 baseline and pre-write admission: the last read-only capture before the first tracker patch retained the original tracker contents and identity under E-I8-01. It matched the inspected canonical target and initial SHA-256; no pre-existing repository edit was present. The update preserves stable historical IDs and all original execution records while changing current summaries and adding this plan.

I8-01 / T-034 IN_PROGRESS: §§1–12 now distinguish the document invocation from future implementation; §22 supplies decisions, assertion destinations, validation and evidence. F-33/34, S-37–40, V-46–49 and RB-005 are linked. No production refactor, test edit, artifact regeneration or dependency change has been performed. Markdown validation and final preservation are the remaining current-session gates.

The only admitted managed validation kind for this document session is:

```sh
CARTULARY_PREPARATION_POLICY=installed_only COREPACK_ENABLE_NETWORK=0 CARTULARY_TEST_RESULTS_DIR=/tmp/reference-i8-docs-e5p53yna/results CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/reference-i8-docs-e5p53yna/static-cache CARTULARY_TEST_RUN_ID=reference-i8-docs make lint-markdown
```

The command uses the existing pinned toolchain, cannot install prerequisites under this policy, and has no Markdown fix mode. Make's workspace lock already exists; its identity/content is included in preservation checks. Results and new static-cache material stay under the named external directory. `git diff --check` and full tracked/nonignored content/index/identity comparison complete V-49. Format, generation, `agent-finalize`, product tests, producer build/source/package commands and service operations are skipped because the authorized change is documentation only. Future V-46–48 and historical passes must not be reported as current execution.

I8-02 / T-034 document update COMPLETE (E-I8-06): the exact documentation command above passed with native outer exit 0 (session 62440). Its owning summary at `/tmp/reference-i8-docs-e5p53yna/results/reference-i8-docs/adhoc/lint-markdown/tool-run-summary.json` reports pass, zero failures and 25,170 ms. `git diff --check` passed. `document-review.json` confirms original §§13–21 are unchanged and verifies the exact three future row IDs and their selectors. `preservation-checkpoint.json` confirms that only this tracker differs among all 8,942 baseline entries; there are no added/removed repository entries. HEAD, branch, staged entries, raw index content/identity and all 2,525 inspected ignored entries plus the workspace lock are unchanged. No unexpected concurrent edit was observed.

Planning completeness is complete for the bounded I8 design. S-37 is ready for a later authorized task; S-38 remains blocked by RB-005, and S-39/40 await their §6 predecessors. F-33/34 are not implemented or closed. No production refactor or product verification occurred; the current built-in artifacts, fixture signing and historical evidence are preserved. The remaining operation in this document session is validation of this completion transcription, not implementation of S-37.

The final transcription uses the same admitted environment and `make lint-markdown` with fresh run ID `reference-i8-docs-final`. Its owning summary is `/tmp/reference-i8-docs-e5p53yna/results/reference-i8-docs-final/adhoc/lint-markdown/tool-run-summary.json`; native exit and final review are recorded in `/tmp/reference-i8-docs-e5p53yna/final-checks.json`, with full comparison in `preservation-final.json`. These paths specify the final gate without preclaiming its result. No repository edit follows the successful final check. Ignored-location preservation remains bounded to §1's explicit scope; unrelated ignored dependency and historical-run trees are not claimed as fully fingerprinted.

### I8 implementation checkpoints

**I8-03 / S-37 DONE / E-I8-S37 — 2026-11-08.** The current user explicitly authorized the complete implementation plan. This supersedes the earlier document-only scope for this invocation, without rewriting its evidence. The content/index baseline is `/tmp/reference-i8-implementation-xt30lmqv/baseline.json`; the incoming tracker edit is retained in `tracker-before.md` and `preexisting.diff`. Baseline HEAD is `e2291271a502033b85c5e2d6aace2609bc8c8210`; the index was clean. All 8,942 tracked/nonignored files were fingerprinted, including 497 immutable files under `contracts/reference-packs/builtins`, `contracts/reference-pack-producers` and `fixtures/reference-packs` (including historical, canonical and signed evidence). Ignored execution outputs are outside this preservation inventory.

Revalidated sources: root/backend/tooling instructions; Reference Pack §1 and RP-REQ-059/060; Core 00 and domain owner navigation from the immediately preceding plan; producer function, dispatcher, bounded CLI subprocess helper, witness admission/wrapper; retired helper and all ZIP/fixture callers; current Base loader and its consumer assertions; generated-artifact policy, topology outputs, pin registry and authored three-row routing. There are no additional nested instructions in the affected subtrees. Searches confirm one executable reissue caller and one retired-fixture caller; the latter varies only the signed flag. No normative amendment, compatibility alias, schema migration or replacement builder is required. The current assertion destinations remain those listed above.

`make help ALL=1`, `make task-guide OWNER=module.reference_data` and `make task-guide OWNER=module.reference_data DETAIL=rows JSON=1` all exited 0; detailed routing is retained in `/tmp/reference-i8-implementation-xt30lmqv/task-guide.json`. The exact baseline command was:

```sh
CARTULARY_TEST_RUN_ID=reference-i8-s37-xt30lmqv CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_9107ea959a,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.unit.canonical_consumers
```

Native session `69374` exited 0. Run root `/home/jochi/code/cartulary/.cartulary/test-results/reference-i8-s37-xt30lmqv` contains `run-manifest.json`, owning `run-summary.json`, all three passing row results, cleanup results and retained-secret scan. The summary reports pass, two passed execution units, zero failures/skips, two cache bypasses and 5,634 ms. This is pre-removal characterization, not final qualification. RB-005 is resolved. S-37's checkpoint is recorded before any S-38 implementation edit; S-38 may now begin, while S-39 and S-40 retain their sequential gates.

**I8-04 / S-38 DONE / E-I8-S38 — 2026-11-08.** After the S-37 checkpoint, deleted authored `producer/base.go` (`ReissueBase` and `BaseReissueInput`), its `producer/command.go` dispatch branch and the corresponding `tools/reference-packs/witness.mjs` allowlist entry. Current producer routes and witness branches remain intact. Added a subtest to `TestATTACKProducerCLI` using the existing production-mode bounded/joined child, matching executable identity, valid legacy fields, absent previous-release canary and parent-owned temporary output/report paths. It requires native exit 1, exactly `reference-pack-source: unsupported package input schema` on stderr, empty stdout and absent output parent, release and report. The changed test has no public test hook or new selector.

```sh
CARTULARY_TEST_RUN_ID=reference-i8-s38-xt30lmqv CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_9107ea959a
```

Native session `62540` exited 0; `/home/jochi/code/cartulary/.cartulary/test-results/reference-i8-s38-xt30lmqv/run-summary.json` reports pass, one passed unit, zero failures/skips, one cache bypass and 4,006 ms. The selected row, cleanup and retained-secret scan pass. Repository searches find the retired input schema only in its negative CLI assertion outside documentation; no executable function/type/dispatch/witness reference remains. `/tmp/reference-i8-implementation-xt30lmqv/preservation-s38.json` confirms all 497 immutable artifact hashes are unchanged.

Compatibility: offline callers of the retired schema must stop invoking it; no alias, replacement builder, runtime contract change or data migration is introduced. The removal eliminates the dated transition and unchecked predecessor-shape parsing. Rollback would restore the function, dispatch, witness entry and test change together without rewriting artifacts. F-33 remains pending S-40 final qualification. This checkpoint precedes all S-39 edits; S-39 may now begin.

**I8-05 / S-39 DONE / E-I8-S39 — 2026-11-08.** After the S-38 checkpoint, narrowed `retiredReferencePackBundle` to `(testing.TB, bool)` and removed `retiredBundleOptions` plus unused corruption/path/omission/extra-member branches. The helper fixes the same host-registry identity, contract, payload and member path previously supplied/defaulted. The sole verifier caller still loops over both signed/unsigned variants and requires `bundle_hint_invalid`. Its current-format success assertion and unrelated ZIP-helper callers are unchanged. The helper comment explicitly distinguishes the checksum-labelled signed variant from a cryptographic signature. No persisted fixture bytes, runtime code or public contract changed in this slice.

```sh
CARTULARY_TEST_RUN_ID=reference-i8-s39-xt30lmqv CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2
```

Native session `73503` exited 0; `/home/jochi/code/cartulary/.cartulary/test-results/reference-i8-s39-xt30lmqv/run-summary.json` reports pass, one passed unit, zero failures/skips, one cache bypass and 20,106 ms. The selected row, cleanup and retained-secret scan pass. `git diff --check` passes; source review confirms the useful fixture behavior and ZIP helper remain while the unused options disappear. Rollback would restore the helper and caller together. F-34 remains pending S-40 qualification. This checkpoint is recorded before S-40 finalization; all predecessor gates for S-40 are now satisfied.

**I8-06 / S-40 IN_PROGRESS — executable qualification.** After E-I8-S39, `make agent-finalize` passed and reported zero generated updates. Executable source was then frozen in `/tmp/reference-i8-implementation-xt30lmqv/source-freeze.json`. All final graph manifests identify source commit `e2291271a502033b85c5e2d6aace2609bc8c8210` plus dirty-source digest `sha256:960d8a14b70c784cda35d19eded9306a0b1ad333887b56384b16be2a2c3354ab`. This is also S-39's source digest; S-37 and S-38 remain separate earlier-source evidence. Only this documentation artifact changes after the freeze.

Each table command ran from `/home/jochi/code/cartulary` with `CARTULARY_TEST_RUN_ID` equal to its listed run ID. All graph commands also used `CARTULARY_HARNESS_CACHE_MODE=off`; the producer build used its normal two-pass reproduction workflow. Native exits were observed independently from summaries. Run roots are `/home/jochi/code/cartulary/.cartulary/test-results/<run-id>`; graph results use the owning `run-summary.json`, `run-manifest.json`, row/unit results, `cleanup-results.json` and `retained-secret-scan.json`.

| Command / acceptance purpose | Run ID / native session | Native exit / result | Owning evidence / limits |
| --- | --- | --- | --- |
| `make agent-finalize` / required pre-qualification refresh | `reference-i8-finalize-xt30lmqv` / `13172` | 0 / pass; 30,471 ms | `unit-artifacts/finalize-summary.json`: generated unchanged, zero updated files; retained-run maintenance skipped because no compatible full warm-check root was supplied |
| Three-row `make test-slice` below / final V-46–48 behavior | `reference-i8-final-xt30lmqv` / `14045` | 0 / pass; 3,319 ms | All three row results passed; two execution units passed, zero failures/skips, two cache bypasses |
| `make lint CHECK=backend-boundary` / offline producer exclusion and owner boundaries | `reference-i8-boundary-xt30lmqv` / `79277` | 0 / pass; 3,730 ms | Four execution units passed; no boundary-rule change |
| `make source-check CHECK=generated` / generated projection drift | `reference-i8-generated-xt30lmqv` / `56280` | 0 / pass; 16,358 ms | Five execution units passed; generated output unchanged |
| `make source-check CHECK=schema` / machine schema integrity | `reference-i8-schema-xt30lmqv` / `53456` | 0 / pass; 8,393 ms | Three execution units passed; no schema edit |
| `make reference-pack-producer-build` / changed executable reproducibility | `reference-i8-build-xt30lmqv` / `85051` | 0 / pass; 1,710 ms | `reference-pack-producer-build/tool-run-summary.json`; two byte-identical builds; no pack production, signing or release certification |

The exact final behavior command was:

```sh
CARTULARY_TEST_RUN_ID=reference-i8-final-xt30lmqv CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_cli_enforces_current_inputs_and_9107ea959a,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.unit.canonical_consumers
```

Every graph invocation in the table has completed cleanup with zero reported cleanup failures and a passing retained-secret scan. Unit counts describe harness work, not individual test counts. The reproduced producer SHA-256 is `0e3b60629b8624d55bfa0f8cc781788a032071531b23a55712ec75a24b2a7c03`; its immutable executable, identity and reproduction record are under `.cartulary/reference-packs/builders/0e3b60629b8624d55bfa0f8cc781788a032071531b23a55712ec75a24b2a7c03`. Existing release manifests are not rewritten to identify this new builder.

The current three rows preserve the retired-input negative assertion, both predecessor verifier rejections, current-format acceptance and the Base release/three-registry/nine-indicator obligations. Source inspection covers witness admission removal and unchanged shared ZIP callers. These checks close the selected behavioral and boundary obligations; they do not establish complete-owner acceptance or qualify an external-data release. Integration, browser/UI, full-owner/CI, source acquisition, corpus replay, signing, deployment and release certification were not run because this bounded change does not touch those behaviors. No full warm-check was run solely to obtain optional maintenance input.

Remaining S-40 gates: `make lint-markdown`, `git diff --check`, final content/index/immutable comparison, and completion transcription. F-33/F-34 remain open until those gates pass.

**I8-07 / S-40 DONE / E-I8-S40 — 2026-11-08.** All scoped implementation, qualification and handoff gates pass. F-33/F-34 are closed; T-035–038 and S-37–40 are DONE. Each predecessor completion was written to this controlling artifact before the next workstream began. No owner amendment or migration was needed, no compatibility shim or general replacement builder was added, and current Base releases retain their original identities/provenance.

`CARTULARY_TEST_RUN_ID=reference-i8-docs-xt30lmqv make lint-markdown` passed with native session `98828`, exit 0, and owning summary `/home/jochi/code/cartulary/.cartulary/test-results/reference-i8-docs-xt30lmqv/adhoc/lint-markdown/tool-run-summary.json` reporting pass, zero failures and 23,410 ms. `git diff --check` also passed. `/tmp/reference-i8-implementation-xt30lmqv/preservation-s40.json` confirms the exact change set: the tracker, CLI implementation/test, retired helper, verifier caller and witness script changed; only `producer/base.go` was deleted; no tracked/nonignored files were added. The 497 immutable artifact hashes, frozen executable source, original §§13–21, HEAD, branch, raw index content/identity and staged entries are unchanged. There is no unexpected tracked/nonignored change. Incoming tracker content was retained before reconciliation; current summaries were deliberately updated without erasing historical execution records. Ignored cache/build/run output preservation is not claimed.

| Obligation | Final completion evidence |
| --- | --- |
| F-33 / C-18 / V-46 | E-I8-S38 removal and exact production-dispatch rejection; final CLI row passes in `reference-i8-final-xt30lmqv` |
| F-34 / C-13/19 / V-47 | E-I8-S39 minimal helper; both signed/unsigned predecessor cases and current-format acceptance pass in the final verifier row |
| C-12/15 / V-48 preservation | Final Base-consumer row passes with three registries and nine indicator types; all 497 immutable artifacts retain their original hashes |
| V-48 source/build boundaries | Final boundary, generated/schema checks and two-pass producer reproduction pass on frozen source; finalization changed no generated output |
| Tracker integrity and handoff | Markdown/whitespace and content/index/historical-record comparison pass; current statuses, compatibility impact, exact evidence and excluded qualification scopes are reconciled |

Final completion-transcription validation uses `CARTULARY_TEST_RUN_ID=reference-i8-docs-final-xt30lmqv make lint-markdown`, `git diff --check` and the same preservation comparison. Its owning summary is `/home/jochi/code/cartulary/.cartulary/test-results/reference-i8-docs-final-xt30lmqv/adhoc/lint-markdown/tool-run-summary.json`; exact native exit and final results are retained in `/tmp/reference-i8-implementation-xt30lmqv/completion-validation.json`, with final comparison in `preservation-final.json`. No repository edit follows the successful transcription check. No additional implementation work remains in the approved F-33/F-34 scope; a future general Base builder remains intentionally deferred until a concrete release needs it.

## 23. Reference-pack independence and generic-format remediation

### Authorization and baseline

The user approved implementation of the complete six-gap plan on 2026-11-08.
The starting checkout is clean `main` at
`70ba53b47269d1733f1bdbbf310cfd4b94a7b455`. The adopted direction is one
application-independent format module and four producer modules in this repository,
a generic dataset profile, pack contract v3 / extension major 4, exact-set/catalog
v3 and incident-bundle 8. Source-specific correctness moves to producers; runtime
retains trust, structural safety, integrity, references, bounded indexing and
lookup validation. Optional Entra SPNs remain supported with zero actual baseline
coverage. No permanent predecessor reader, automatic conversion, deployment
installation or hosted publication is authorized by this implementation scope.

The initial ATT&CK licensing gate and blocked-release checkpoints below are historical. The user subsequently directed adoption of MITRE's issued license; the production release and gate-resolution checkpoint at the end of this section controls current distribution status.

### Phase gates

| ID | Work | Status | Dependency | Exit |
| --- | --- | --- | --- | --- |
| RI-P0 | Baseline, retained Windows import and assertion destinations | DONE | Authorized plan | Four corpus baselines, Windows real-artifact pass, boundary inventory |
| RI-P1 | Coordinated owner amendments and typed contract freeze | DONE | RI-P0 | Generic contracts, safety limits and assurance boundary adopted |
| RI-P2 | Standalone format library and four producers | DONE | RI-P1 | Isolated builds and both dependency-direction guards pass |
| RI-P3 | Generic projections and production publication | DONE | RI-P2 | Four qualified candidates; explicit ATT&CK distribution disposition |
| RI-P4 | Generic runtime and coordinated current-only cutover | DONE | RI-P2, RI-P3 | Lifecycle, consumer, Base, portability and recovery qualification |
| RI-P5 | Exact signed-artifact qualification | DONE | RI-P3, RI-P4 | Four independent real-artifact scenarios plus security/extensibility portfolio |
| RI-P6 | Cleanup and final handoff | DONE | RI-P5 | Legacy paths removed, final source qualified, per-pack handoffs complete |

### Evidence and checkpoints

- RI-E0: the initial source identity and clean status were inspected directly.
  Immutable corpus/member fingerprints are retained in
  `.cartulary/reference-packs/independence-v3/p0-artifact-baseline.json`.
  This is an implementation preservation inventory, not a prose-derived product
  contract or qualification result.
- RI-E1: `make task-guide OWNER=module.reference_data` and `make help ALL=1`
  discover the current Make surface. No previous audit is current qualification.
- RI-P0 adds an independently selectable Windows retained-release import test.
  It reads the exact published archive/bootstrap, validates their recorded hashes,
  imports without re-signing, retrieves all 503 complete canonical definitions,
  exercises provider/version/ambiguous/no-hit queries, and proves disablement
  preserves historical reads. The synthetic lifecycle remains separately selected.

### Finding and assertion destinations

| Gap | Owning remediation | Required destination |
| --- | --- | --- |
| RI-R1 | Independent production and consumption | Module build closure, transitive import/embed/input guards, standalone vectors |
| RI-R2 | Real ATT&CK publication | Explicit release authority, license decision, append-only ledger and exact-artifact import |
| RI-R3 | Truthful optional SPN support | Descriptor zero coverage; real no-hit and synthetic positive/duplicate/case tests |
| RI-R4 | Windows real-artifact coverage | `TestWindowsRealReleaseImport_Integration`, followed by its generic release successor |
| RI-R5 | Generic extensibility | Generic safety tests, producer native semantics, fifth unknown-key dataset |
| RI-R6 | Coherent owner and evidence boundary | Coordinated contracts and final source-bound handoff; no executable prose dependencies |

Engineering and four-pack artifact delivery are complete on the scoped evidence below. RI-GATE-ATTACK was subsequently resolved through the user-directed MITRE license adoption and exact production publication/qualification recorded at the end of this section. Historical blocked statuses are not current-state claims.

### RI-P0 completion — 2026-11-08

`reference-independence-v3-p0-r1` passed with caching disabled, native exit 0,
and all six selected rows passing. The exact row selection is recorded in its
run manifest under `.cartulary/test-results/reference-independence-v3-p0-r1/`.
It covers the new Windows retained-release import, VERIS and Entra real-release
imports, authentic ATT&CK producer artifacts with fixture authority, optional-pack
absence/failure and Base canonical consumers. All 15 cleanup operations completed;
the retained-secret scan passed over 50 files. This is baseline characterization,
not v3 qualification or production ATT&CK publication evidence. Source identity is
commit `70ba53b47269d1733f1bdbbf310cfd4b94a7b455` plus dirty-source digest
`sha256:d6d4a89659066e2645cd0aa3845c71d707cc7608f5b839da99885c4b60263345`.
`make format` and `make generate` passed before that run; generation added the
new test selection's topology projection.

The preservation inventory records 543 files. The separate corpus inspection
`p0-corpus-inventory.json` records ATT&CK 6,221 objects/34,632 relationships,
VERIS 2,491/2,533, Windows 503 definitions and Entra 139 application identities,
zero SPNs. The largest native row is 3,034,554 bytes and maximum native nesting is
five. Final generic bytes, lookup counts and aggregate key bytes still require
measurement before the new projections are qualified.

### Assertion ownership inventory

These are explicit transfer obligations; moving a source file alone does not close
them. The listed suites remain active until destination qualification passes.

| Existing assertions | Destination after cutover | Preservation or intentional change |
| --- | --- | --- |
| `attack_native_test`, `attack_source_contract_test`, native members of `attack_remediation_test` | ATT&CK producer native-model validation | Source shapes, native absence/ordering/lifecycle, display projection, citations, domains, embedded edge cardinality/direction, occurrence provenance, arrays and source bounds retained |
| `veris_test` | VERIS producer native-model validation | Schema reconstruction, contextual identities, rules, semantic edges, provenance, cycle/source-keyword and native bound checks retained |
| Windows semantic assertions in `content_test` and producer Windows suites | Windows producer | Stable provider/event/version identities, native scalar bounds, full documentation, source reconciliations and editorial dispositions retained; optional-version query replaced by two named composite lookups |
| `entra_content_test` and producer Entra suites | Entra producer plus generic normalization vectors | 139 IDs/zero real SPNs, UUID/source correspondence, qualifiers/aliases, 64 SPNs and 1024-byte native bound, normalized shared-SPN rejection retained |
| `framework_lookup_test` | Producers for citation tuples; generic engine for postings | Full citation tuples remain distinct; producer coalesces repeated factual lookup values before emission; generic engine rejects duplicate normalized postings |
| Archive, manifest, inventory, notice, canonical JSON and TUF suites | Neutral format library | Preserve exact canonical bytes, framing, trust, threshold, rollback, rotation, expiry, inventory and verification precedence; new contract vectors are separately versioned |
| Schema compiler, depth, line, duplicate, endpoint, replacement-chain, cancellation and safe-diagnostic assertions | Neutral generic admission; application integration for cancellation/publication | Native shape checks leave runtime; closed envelopes, bounded inert JSON, local references and complete private indexing remain runtime obligations |
| Set, provenance, consumer, indicator, lifecycle, retention, portability, recovery and attestation suites | Reference Data and participating application owners | Exact-set provenance, Base algorithms, atomic publication, Jobs and authorization retained under the coordinated version changes |
| Fixture-only repository, ledger, key confinement, reservations and renewal suites | Neutral publication package plus each producer authority adapter | Fixture/release inputs remain distinct; ATT&CK gains a separate production path only with an exact-output license decision |

### Boundary and interface inventory

Current coupling enters through `producer/release_verification.go`, producer
canonicalization imports, generated application schemas and the omnibus
`tools/reference-packs/main.go`. Source profiles embed producer contracts from
`contracts/reference-pack-producers`. Test support also imports the producer.
The format extraction must separate canonical/container/trust primitives from
application attestations, operations, sets, indicator policies and recovery.

Public affected projections include manifest, profile registry, lookup request,
consumer item/results, immutable sets/reference catalogs, portability bindings,
incident-bundle version, release-bound Base manifests and recovery codec bindings.
The administrative routes, common Jobs and deployment authorization remain with
their current owners. Build/test routing currently assumes the root Go module;
standalone module archives need explicit task-surface, execution-topology and
transitive import/embed/input-discovery checks. RI-P1 begins only after this
checkpoint; no producer assertion is retired by this inventory.

### RI-P1 contract preparation checkpoint

The Reference Pack owner now specifies generic dataset v1 under RP-REQ-275–283,
with explicit native-producer responsibility, bounded inert data, injective
normalized key framing, exact coverage counts, separate Windows composite lookups
and descriptor consumption. Companion owner amendments bind major 4 / contract v3,
sets/catalogs v3 and incident-bundle 8. Current machine schemas are being prepared
alongside predecessors; the runtime switch and predecessor removal remain RI-P4/P6.

Prospective generic envelope sizing over every retained canonical row is recorded
in `p1-generic-sizing-inspection.json`. Maximum encoded rows including LF are
3,035,129 bytes (ATT&CK), 4,102 (VERIS), 7,696 (Windows) and 1,588 (Entra); maximum
nesting is six. Maximum object postings are 25/2/3/4 respectively. Aggregate key
bytes are 448,241/296,079/35,137/9,460, well below 256 MiB; the maximum single key
is 88 bytes. Counts remain the P0 counts. This is sizing inspection of the defined
projection, not a producer witness or release qualification. Final descriptors,
canonical outputs and independent expectations still require RI-P3 validation.

RI-P1 validation: `make generate` passed in `20261108T160605Z-p87822`;
`make source-check CHECK=schema` passed in `reference-independence-v3-p1-schema-r1`
with native exit 0 and caching disabled. `make lint-markdown` passed in
`reference-independence-v3-p1-docs-r1`, native exit 0. Subsequent owner-version
cleanup and operation-input v4 preparation will be included in the next refresh.
The contract decisions and assertion allocation are fixed; executable adoption,
Base reissue, participating-owner projections and final-source qualification remain
explicit RI-P4/P5 work. RI-P2 begins with the neutral library and module-aware
build/test routing; no legacy producer or validator has yet been deleted.

### RI-P2 foundation checkpoint

A standalone `referencepacks/format` module now contains canonical JSON and a
bounded generic dataset engine. It has its own manifest/checksums, no application
imports, and focused rows under `module.reference_data.unit.independent_format_*`.
`reference-independence-v3-format-r1` passed both initial canonical/dataset rows
with cache off and native exit 0; cleanup completed and the retained-secret scan
passed. These are foundation tests, not complete format or release qualification.
Application runtime, producers and retained archives remain on the predecessor
implementation until their coordinated cutover.

Module-aware Go routing uses an authored registry, `GOWORK=off`, `-mod=readonly`
and explicit module working directories. Module rows do not load application
authentication fixtures. Preparation uses `make bootstrap PART=go`; it generated
the module checksum file and also corrected the root's existing `x/net` direct
import classification without changing its version.

Intermediate generation/format attempts exposed both closed runner-schema root
allowlists. Generation failures are retained in `20261108T162326Z-p98858` and
`20261108T162427Z-p2745`; their failing task-surface generator logs identify the
registry-root and exact-selector patterns. The authored schemas were updated;
`make generate` then passed in `20261108T162531Z-p6866`, and `make format` passed
in `20261108T162616Z-p10302`. The earlier format attempts failed plan admission
before formatter execution, not product assertions.

`reference-independence-v3-runner-r1` passed the scheduler contract but failed the
public command-surface row: a minimal selected-source fixture lacked the new
module registry. Investigation also removed eager reads of every sibling module
manifest from build-context discovery. Metadata loading now reads the registry;
execution validates only the selected module's manifest. The fixture includes the
new metadata. `reference-independence-v3-runner-r2` reran only the affected
command-surface row and passed, native exit 0 with cache off. No complete harness
acceptance is claimed for these focused runs.

Container, manifest, notice and offline-TUF primitives are now being extracted
into the module's `envelope` package. Its machine inputs are embedded within that
module; application profile/Base policy, sets, Jobs, attestations and diagnostic
retention remain internal. Existing application assertions stay active while their
standalone counterparts are qualified. Remaining RI-P2 obligations include full
neutral primitive qualification, publication extraction, four producer modules,
exported-archive builds, transitive/embed/generator/input boundary negatives and
complete generic diagnostic enumeration. RI-P2 is not complete.

#### RI-P2 standalone envelope and publication checkpoint

The format module now owns neutral archive admission/writing, manifest v2 and
contract v3 admission, offline TUF, inventory hashing, SPDX expression checks,
and notice validation. Application registries, activation, jobs, retained
attestations, and dependency policy remain outside it. `verification.Dataset`
combines these primitives with generic admission without application imports.
A separate `publication` package owns typed release/fixture entry points, a
private-key-free signer interface, scoped private signing contexts, append-only
logical and metadata reservations, immutable publication, and renewal. Release
publication requires an exact approved-output binding represented in manifest
provenance; this mechanism does not constitute ATT&CK licensing approval.

- `make bootstrap PART=go`: native exit 0; run
  `20261108T164800Z-p57809` updated the new module's managed dependency checksums.
- `make format`: native exit 0; run `20261108T164907Z-p58329` passed.
- `make generate`: native exit 0; run `20261108T164925Z-p65597` passed.
- Cache-disabled `make test-slice OWNER=module.reference_data
  ROWS=module.reference_data.unit.independent_format_envelope,module.reference_data.unit.independent_format_dataset`:
  run `reference-independence-v3-envelope-r1`, native exit 0, pass 1,356 ms,
  all 11 cleanup actions completed, retained-secret scan passed (16 files).
  This covers independent TUF vectors, root replay/thresholds, expiry/tampering,
  archive traversal/ratio/collision/cancellation, changed admitted sources, and
  the generic dataset structural, lookup, reference and resource boundaries.
- Cache-disabled `make test-slice OWNER=module.reference_data
  ROWS=module.reference_data.unit.independent_format_publication`:
  run `reference-independence-v3-publication-r1`, native exit 0, pass 1,543 ms.
  Authority separation, exact output approval binding, renewal, tampered prior
  envelopes, expiry, concurrent logical reservations, metadata rollback and
  immutable output collisions passed using synthetic authority only.

These are narrow development results. The source archive qualification,
producer modules, runtime cutover, final-source qualification, and release
artifacts remain outstanding. Existing private implementations remain until
replacement obligations pass. No claim of four-pack readiness is made.

#### RI-P2/RI-P3 independent producers and candidate checkpoint

The repository now contains five independently versioned Go modules under
`referencepacks/`: `format`, `attack`, `veris`, `windows`, and `entra`. Each
producer has its own source inputs, native validation, command composition,
executable, and dependency manifest. The public producer targets require
`PRODUCER=attack|veris|windows|entra`; implicit omnibus selection is retired from
these targets. The old internal dispatcher remains pending caller/test migration
and removal in RI-P6. Runtime consumption is still the predecessor implementation.

`make reference-pack-module-check MODULE=<module>` exports only the selected
module, extracts it outside the application checkout, and builds with
`GOWORK=off`, `GOENV=off`, `GOPROXY=off`, and `-mod=readonly`. It inspects the
transitive package/test/embed closure. Snapshot module versions bind the complete
exported source digest; producer manifests pin an immutable format module archive,
with no local `replace`. Explicit `make bootstrap PART=go` owns archive dependency
preparation and managed checksum updates. `reference-pack-producer-build` also
builds each selected exported module twice and retains byte-identical executables.

The four producer projections use the same descriptor/object/relationship engine.
Each object or relationship retains its complete predecessor source-model record
as inert `data`; producer-owned native validation runs before serialization.
Windows declares separate `provider_event_id` and `provider_event_version`
lookups. Entra declares SPN lookup with measured zero real coverage. Production
and fixture signing have distinct typed input contracts. ATT&CK production uses
`cartulary.attack`; its candidate output policy records pending external licensing
review. No ATT&CK distribution approval has been inferred or recorded.

Narrow development evidence, with native exit 0 unless explicitly identified:

- `reference-independence-v3-producer-deps-r2` and `-r3`: public Go bootstrap passed;
  current producer manifests were updated by the dependency tool.
- Isolated module checks: Entra and VERIS `*-build-r1` passed; ATT&CK and Windows
  `*-build-r2` passed after fixing an extracted type-name collision and a missing
  source commit validator. Failed `*-build-r1` roots preserve those compiler
  errors. These checks predate subsequent command/test additions.
- `reference-independence-v3-generic-encode-r1`: cache-disabled dataset and
  publication rows passed (1,369 ms), including producer encoding and typed
  approval binding.
- `reference-independence-v3-producer-command-generate-r1`: generation passed.
  `reference-independence-v3-producer-format-r1` failed plan admission because the
  newly relocated tests were not yet registered; `-r2` passed after authored
  routing was added. No formatter assertion failed in the first attempt.
- Reproduced independent binaries: ATT&CK, Windows and Entra `*-binary-r2`, and
  VERIS `*-binary-r3` passed. ATT&CK `*-binary-r1` failed compilation because a
  flag import collided with a producer helper; that import was explicitly aliased.
  VERIS retains its existing production repository `cartulary.veris.community`.
- `reference-independence-v3-producer-tests-r1`: all four selected producer rows
  passed with cache disabled (2,967 ms). This includes relocated acquisition,
  integrity, projection, publication and fixture-authority cases, plus positive
  synthetic Entra SPNs, ASCII matching, within/across-object duplicate rejection,
  exact provenance and zero-coverage behavior.
- `reference-independence-v3-native-tests-r1`: Entra and VERIS source-model rows
  passed; ATT&CK failed test compilation on an extracted import, native outer exit
  2, classified `infra/preflight_error`. All 12 cleanup actions completed.
  `reference-independence-v3-native-tests-r2` reran the repaired ATT&CK row only
  and passed with cache disabled (973 ms). Native ATT&CK typed endpoint/source
  limits and VERIS corpus reconstruction, semantic rejection and graph bounds
  now have producer-owned destinations.

The independent Node witness consumes canonical artifacts and explicit retained
predecessor expectations, never the producer implementation or upstream Markdown.
It compares every complete native object/relationship, common factual fields,
lookup mapping and notice, and independently measures normalized coverage.
The following unsigned candidate invocations all passed:

| Pack | Explicit run | Objects | Relationships | Postings | Encoded key bytes |
| --- | --- | ---: | ---: | ---: | ---: |
| ATT&CK | `reference-independence-v3-attack-candidate-r1` | 6,221 | 34,632 | 14,490 | 448,241 |
| VERIS | `reference-independence-v3-veris-candidate-r1` | 2,491 | 2,533 | 4,982 | 296,079 |
| Windows | `reference-independence-v3-windows-candidate-r1` | 503 | 0 | 1,439 | 35,137 |
| Entra | `reference-independence-v3-entra-candidate-r1` | 139 | 0 | 278 | 9,460 |

Each exact candidate report and witness is retained below
`.cartulary/reference-packs/independence-v3/candidates/<producer>/`; the invocation's
published stdout identifies its immutable manifest, payload, and witness hashes.
Windows has 433 version-specific postings; Entra has zero SPN objects, postings,
and distinct keys. Every candidate passed the inclusive generic limits.
These results establish candidate preservation and feasibility, not final-source
or signed-release qualification. Later source changes require new builder/release
identities; retained candidate bytes and historical releases remain unchanged.

RI-P2 remains open for complete executable/generation input enforcement, actual
application dependency closure checks, neutral schema ownership, and complete
safe generic diagnostics. RI-P3 remains open for the remaining assertion
allocation, production signing/renewal qualification, and release dossiers.
RI-P4–RI-P6 have not exited. The application cutover, participating-owner contract
bindings, current Base releases, final exact-artifact import tests, finalization,
and handoff are outstanding. Four-pack distribution readiness is not claimed.


#### RI-P2/RI-P4 shared primitives and consumer cutover checkpoint

This checkpoint supersedes the earlier statement that runtime consumption is
still entirely the predecessor implementation. The generic runtime is now
implemented and has narrow passing evidence; RI-P2–RI-P6 have not exited.

The application delegates container admission/writing, manifest parsing and
inventory hashes, notices/licensing syntax, bootstrap decoding and offline TUF
to the standalone format module. Application diagnostics, registry policy,
transactions, Jobs, retention and attestations remain internal. Private native
validators and the duplicate private schema compiler remain pending removal;
this checkpoint does not claim that extraction is complete.

Consumer request v2, generic item/results v3 and exact retained descriptor reads
are implemented. Generic lookups preserve exact-set binding, confidential cursors,
zero-hit success and separate invalid/unsupported/unavailable outcomes. Base
release v3 preserves its payloads and algorithms. Incident-bundle format 8,
reference catalogs/sets v3, Reference Pack extension major 4, recovery codecs v4,
state catalog v3 and participating owner bindings are generated together.
Terminal operation admission now also covers closed bootstrap/reconciliation
inputs, rather than treating those records as ordinary pending operations.
Predecessor recovery catalog bytes were moved intact to historical fixtures.

Assertion destinations retained in independent producers include:

| Original obligation | Current destination |
| --- | --- |
| ATT&CK source discriminator, recursive closure and exact optional fields | `referencepacks/attack/internal/native/source_envelope_test.go`, `source_contract_test.go` |
| ATT&CK disjoint schema dispatch, exact/one-over native arrays and bounded traversal | `referencepacks/attack/internal/native/schema_boundaries_test.go`, `source_limits_test.go` |
| ATT&CK native typed endpoints and embedded relationships | Independent ATT&CK native source tests; runtime replacement is generic signed local-reference rejection |
| Windows identity, nullable/numeric versions and documentation normalization | `referencepacks/windows/internal/native/source_policy_test.go`, `source_envelope_test.go` |
| Entra recursive closed source records and SPN correctness | `referencepacks/entra/internal/native/source_envelope_test.go` and independent source-model tests |
| Canonical archive/TUF parsing, descriptor admission and thresholds | `referencepacks/format/envelope`; application adapter checks preserve diagnostic/operational distinctions |

Application lifecycle and participating-owner tests now use a test-only in-memory
signer instead of importing producers. The neutral publication package exposes a
fixture-authority-only envelope path for negative payloads; production publication
still requires typed release approval, reservations and standalone verification.
The Windows synthetic lifecycle reads authored canonical generic bytes and never
acquires or parses source documents. Generic corpus fixture catalogs bind the
actual producer outputs; independent factual expectations remain separate.

Development evidence (not final-source qualification):

- `reference-independence-v3-owner-bindings-generate-r5`: `make generate`, native
  exit 0. Earlier `-r2` through `-r4` failed stale recovery generator/binding
  projections; those exact roots are retained. `-r1` has a failed structured
  summary but its native exit was not observed.
- `reference-independence-v3-neutral-runtime-r3`: trust adapters, generic runtime,
  descriptor consumer, Windows and Entra source-model rows passed; ATT&CK test
  compilation failed on a stale helper call, native exit 2. The corrected
  ATT&CK `reference-independence-v3-native-assertions-r2` has a passing structured
  result and completed cleanup; its native exit was not observed. Prior
  `native-assertions-r1` failed because a test supplied a Go integer where a
  canonical decoded JSON number was required.
- `reference-independence-v3-recovery-bindings-r1` and
  `reference-independence-v3-extension-bindings-r2`: selected recovery codec/
  catalog and extension registry checks passed, native exits 0. The first
  extension run failed its stale major-3 expectation.
- `reference-independence-v3-bundle-bindings-r4`: selected cross-owner source-port
  checks passed, native exit 0. Earlier runs exposed remaining version-7 owner
  declarations and an unsupported-version test that incorrectly rejected 8.
- `reference-independence-v3-format-deps-r9`: public Go dependency bootstrap
  passed, native exit 0. It owns the updated immutable format pin/checksums.
- `reference-independence-v3-{attack,veris,windows,entra}-builder-r2`: each public
  independent producer build passed, native exit 0, with repeated byte-identical
  builds from its exported module source.
- `reference-independence-v3-{veris,windows,entra}-candidate-r2` and
  `reference-independence-v3-attack-fixture-20-{0,1}-r1`: producer assembly and
  independent complete-corpus witnesses passed, native exits 0. Canonical
  artifacts are retained under `fixtures/reference-packs/dataset-v1/`; the authored
  `catalog.json` binds manifests, payloads and independently established counts.
  ATT&CK uses fixture authority here; no production approval is implied.
- `reference-independence-v3-retained-input-generate-r1`: generation passed,
  native exit 0, including current bootstrap/reconciliation/portability inputs.
- `reference-independence-v3-windows-generic-lifecycle-r1` and `-r2` failed with
  native exits 2: first a missing synthetic source timestamp, then current
  bootstrap input rejected as an ordinary operation. Both were repaired.
- `reference-independence-v3-generic-fixtures-r2`: Windows generic lifecycle
  passed; the other two selected rows failed on fixture artifact-root selection
  and independent alias ordering. All 16 cleanup actions completed; outer
  native exit 2. `-r1` stopped at routing admission because a new authored row
  was not ASCII-sorted; no product test ran.
- `reference-independence-v3-generic-fixtures-r3`: generic signed reference
  validation, Entra complete factual oracle, and retained binding/operation
  admission passed with caching disabled, native exit 0 (13,053 ms). The generic
  signed test uses an unknown advisory key and proves native data is inert while
  dangling endpoints, duplicate relationships and replacement cycles cannot
  reach indexing.

Remaining work includes complete input-discovery and transitive boundary guards,
legacy runtime/dispatcher removal, current fixture refresh, all four real signed
artifact scenarios, descriptor/storage integration, full resource-bound evidence,
Reporting/Composition and portability/restore integration, finalization, frozen
source qualification and operator dossiers. The ATT&CK licensing decision remains
an external gate. No installed history has been converted or reset, and no
four-pack distribution readiness is claimed.

#### RI-P2/RI-P4 runtime and fixture separation checkpoint

This checkpoint supersedes the preceding remaining-work list where the work below
is now implemented. It is development evidence, not P5/P6 acceptance. Executable
source is still changing; no production envelope or deployment trust was created.

- Removed the omnibus `internal/modules/reference_data/producer` and its command
  composition. Each standalone producer preserves its command rejection, help,
  identity and report-delivery checks. ATT&CK also preserves native child exits,
  frozen-profile selection, attributable audit failures, canonical replay and
  retention before report-delivery failure. The command tests exposed and fixed
  empty report delivery after rejected publication/source inputs.
- Removed ATT&CK/VERIS semantic dispatch, Windows event identity/documentation
  rules, Entra SPN rules and the optional-version Windows query from application
  runtime code. Their source correctness assertions run in producer-owned native
  model packages. The unchanged D3FEND relationship identity algorithm remains
  local to Reference Data. Authored row migrations name replacement obligations.
- Replaced the duplicate application schema compiler, union evaluator and duplicate
  member walker with the neutral format evaluator. Application request ordering,
  exact consumer contracts, findings retention and operational error classification
  remain local. No format publication dependency enters production runtime.
- Reissued current manifest, complete-set, portability and signed synthetic
  vectors through `make generate` using an independent Node implementation of
  canonicalization, SHA-256, Ed25519 and ZIP/TAR/GZIP. The complete set now contains
  64 keys, including unknown advisory datasets. Predecessor input bytes are pinned
  under `fixtures/reference-packs/predecessor-contract-v2-vectors/`. Synthetic
  authority and clock-controlled vectors are not production release qualification.
- Replaced current four-profile structural fixtures with generic envelopes.
  Native endpoint fixtures and their registrations are retained as historical
  inputs outside runtime contract embedding; producer endpoint obligations and
  generic signed-reference rejection have separate active destinations.
- Extended the complete portability boundary recipe to 1,024 sets, 16,384 distinct
  versions and 65,536 required memberships through version reuse over 64 keys.
  Independent Python expectations are 35,317,052 bytes for each reference catalog
  and 11,186,253 bytes for embedded-content declarations. This recipe still needs
  its focused validation run.
- Hardened publication: relative reads use confined open handles and bounded
  regular-file admission; FIFO substitutions cannot block opening; immutable
  records and tree publication sync parent directories; no-replace publication
  uses pinned directory descriptors. Metadata replay validates the entire retained
  reservation chain, including later records. Fixture signing explicitly rejects
  release authority. Focused publication tests passed before the new library pin.

Observed native exits (all successful test invocations disabled result caching):

| Run suffix after `reference-independence-v3-` | Outcome |
| --- | --- |
| `independent-cli-r1` | Exit 2; four commands emitted zero-valued reports after rejected input. |
| `independent-cli-r3` | Exit 0; four independent producer command/source rows passed. |
| `attack-cli-preservation-r1` | Exit 0; preserved ATT&CK command and source assertions passed. |
| `native-removal-r1` | Exit 0; four producer native models and generic runtime passed. |
| `schema-extraction-r1` | Exit 2; stale v2 manifest vector failed one manifest-limit test; trust, retained admission and generic checks passed. |
| `schema-extraction-r2` | Exit 0; refreshed canonical profile/manifest-limit and generic runtime checks passed. |
| `contract-vector-reissue-r1` | `make generate`, exit 0. |
| `generic-fixture-projections-r1` | `make generate`, exit 0. |
| `current-fixture-closure-r1` | Exit 2; admission vectors still embedded predecessor manifest bytes. |
| `admission-vector-reissue-r1` | `make generate`, exit 0; explicit malformed encoding/spacing/duplicate mutations preserved. |
| `current-fixture-closure-r2` | Exit 0; complete current fixture registration/admission closure passed. |
| `publication-durability-r1` | Exit 0; standalone publication authority, renewal, immutable retention and damaged-ledger replay checks passed. |

Preflight-only invocations rejected an unsupported run-root argument, a run ID
supplied as a Make argument instead of environment, a missing owner selection,
unsorted test selectors and an already nonempty run identity. They supply no
product verification; no retained run root was cleared or reused.

Remaining engineering: qualify transitive and executable-input boundaries with
exported module/corpus builds; remove unused predecessor runtime projections;
finish full-bound and descriptor/database tests; issue new signed real artifacts
and independently import each exact envelope; validate Reporting/Composition,
portability, recovery, readiness and core-workflow absence. Then finalize, freeze
source, qualify the affected owner union and deliver release/operator dossiers.
ATT&CK redistribution remains externally blocked; signing-path implementation
must not be reported as licensing approval or four-pack distribution readiness.

#### RI-P2–RI-P5 artifact and database checkpoint

This checkpoint supersedes prior current-state statements about missing production
envelopes and runtime projection cleanup. It does not close the final qualification
or external ATT&CK gate. Source continues to change during the repairs below.

- Retired unused native runtime projections and predecessor Base release bindings
  into byte-preserving historical fixture directories. Current Base manifests and
  independently calculated witnesses bind release 3 without changing Base payloads.
- Qualified all five exported module archives with workspace resolution disabled,
  read-only pinned dependency resolution and no network fallback. Closure inspection
  includes test dependencies and embedded files. Negative guards cover application
  imports, sibling producers, publication imports in runtime, local replacements,
  escaped/prose/corpus/key embeddings, generation directives and implicit executable
  input discovery. Each producer was built twice with identical executable bytes.
- Added empty-directory, empty-environment invocation of copied producer binaries.
  The adapter resolves only explicitly supplied paths. It retains closed-input
  rejection for duplicate JSON members, invalid UTF-8/surrogates, unsafe integer
  encodings, oversized inputs and symlink substitution.
- Qualified complete generic corpus preservation with an independent Node witness:
  native fields, identities, relationships, lookups, notices, provenance and resource
  bounds. This uses retained machine artifacts, never upstream Markdown or owner
  prose. Counts remain ATT&CK 6,221/34,632, VERIS 2,491/2,533, Windows 503/0 and
  Entra 139/0 with zero real SPN postings.
- Issued three production-authority envelopes and one explicitly fixture-authority
  ATT&CK envelope. Their immutable paths are retained beneath
  `fixtures/reference-packs/dataset-v1/`. Per-pack `release.v3.json` selectors bind
  exact archives, bootstraps, manifests, payloads, source profiles and builders.
  Existing output policies authorize the three production outputs. ATT&CK has an
  unsigned `cartulary.attack` candidate and a pending exact-output licensing gate;
  its fixture result is not production distribution readiness.
- Added independently selected exact-artifact tests for all four packs. These read
  delivered archives and bootstraps unchanged, use the current trusted clock, and
  never sign or renew their subject. The three production releases passed. ATT&CK
  initially failed because migration 48 allowed large rows only for retired native
  ATT&CK schema IDs. New migration 50 installs the generic 4 MiB-minus-LF storage
  bound without interpreting inert data, and blocks in-place downgrade. The retained
  ATT&CK scenario then passed. Historical migration 48 remains unchanged.
- Added unknown advisory-key lifecycle coverage, exact descriptor/provenance reads,
  nullable Base descriptors, scalar/composite inputs, distinct errors, pagination,
  cursor expiry, historical-set reads and corruption-triggered atomic invalidation.
  The database-bound test admits equality and rejects one byte over; migration
  traversal, installed catalog parity and downgrade-barrier tests pass.
- Empty-directory corpus replay exposed shared-ledger contamination after signing:
  release reservations rejected envelope subdirectories. The neutral publication
  library now owns separate `logical` and `envelopes` namespaces. A complete
  assembly/signing/replay/renewal/next-release test passes. Flat predecessor ledgers
  fail closed. Existing unpublished development reservations were copied byte for
  byte into new namespaced roots, preserving both original roots and sequence
  history; the explicit mapping is
  `.cartulary/reference-packs/independence-v3/ledger-namespace-transition.v1.json`.
  No deployment state, installed history or signing key was converted.
- The current format pin is
  `v0.1.0-dev.d829afb9b953015a5ba6cc2e0fe5556a41aa3e4d1002f1a76f4137a74284b45f`.
  Four producer rebuilds passed and new `.dataset.4` candidates are being qualified.
  Prior `.dataset.3` production and ATT&CK `.dataset.2` fixture artifacts remain
  immutable historical outputs. Final evidence must follow the selected successor.

Additional observed evidence, with run roots under `.cartulary/test-results/` and
prefix `reference-independence-v3-` (test runs disable result caching):

| Run suffix | Native result and scope |
| --- | --- |
| `generic-db-r1` | 0; unknown-key/ATT&CK generic consumers and complete portability-bound vectors. |
| `input-discovery-boundaries-r2` | 0; module archive and closure negative guards. |
| `backend-boundary-r1` | 0; backend and transitive production dependency direction. |
| `{format,attack,veris,windows,entra}-isolation-r1` | All 0; exported module build and dependency/input closure. |
| `{attack,veris,windows,entra}-release-builder-r3` | All 0; independently reproduced binaries. |
| `{attack,veris,windows,entra}-logical-r3` | All 0; standalone assembly and complete independent preservation witness. |
| `{attack,veris,windows,entra}-envelope-r3` | All 0; exact signed bytes and independent current-clock trust witness. |
| `attack-production-candidate-r3` | 0; unsigned production candidate, licensing still pending. |
| `engineering-finalize-r2` | 0; finalizer before later database/publication repairs. |
| `delivered-{entra_real_release_import,veris_real_release_import,windows_real_release_import}-r1` | All 0; exact delivered production-artifact imports. |
| `delivered-attack_retained_artifact-r1` | 2; database native-schema size dispatch rejected a 3,035,100-byte generic row; all cleanup completed. |
| `generic-storage-tests-r1` | 0; repaired ATT&CK retained import and descriptor-corruption lifecycle. |
| `generic-storage-migrations-r2` | 0; exact storage bounds, installed catalog, migration traversal and downgrade barriers. |
| `empty-cwd-boundaries-r1` | 0; strict explicit-input and dependency boundary tests. |
| `cross-owner-finalize-r2` | 0; generated finalization before the ledger repair. |
| `{attack,veris,windows,entra}-empty-cwd-corpus-r1` | All 2; correctly rejected contaminated flat publication ledgers after signing. |
| `ledger-namespace-tests-r1` | 0; publication authority, immutable reservations, renewal and shared-ledger lifecycle. |
| `format-deps-r11` | 0; Make-owned module pin and checksum update. |
| `{attack,veris,windows,entra}-release-builder-r4` | All 0; reproduced binaries using the corrected publication library. |

Preflight failures also identified an invalid migration row selector and stale
routing/topology projections after authored fixture changes. Corrected public
selections and `make generate` resolved them; no failed root was reused. Finalizer
retained-run maintenance remains skipped because no compatible full warm-check
root was supplied. These are selected engineering results, not complete-owner or
final-source acceptance.

Remaining work: finish successor artifact signing and exact descriptors; qualify
all affected Reference Data backend rows and participating Reporting/Composition,
portability, recovery, server/operator admission and Base/core-absence obligations;
rerun finalization and required generated/schema/boundary/documentation checks on
frozen source; complete per-pack operator/build/renewal/custody handoffs. ATT&CK's
exact-output licensing authority remains the named external distribution gate.

#### RI-P5 artifact and operator handoff before MITRE license adoption

The selected `.dataset.4` releases below supersede the earlier current-artifact
selections. Immutable logical trees, envelope trees and descriptors remain
append-only; `release.v3.json` is only the current selection. Each descriptor
records the full archive, bootstrap, envelope-report, manifest and payload hashes,
source profile and builder identity. Its byte-identical immutable copy is
`descriptors/<container_sha256>.json` beside the selector. The application tests
require that equality before importing the delivered archive unchanged.

| Pack | Selected descriptor | Authority and distribution | Factual corpus |
| --- | --- | --- | --- |
| veris | [2026.11.08.dataset.4](../../fixtures/reference-packs/dataset-v1/veris/descriptors/00d3740ee5f3ffd8cde9048cb34f5997f633d746f5ec925d6a4c06f705ba5e76.json) | `cartulary.veris.community`; approved existing output decision, CC-BY-SA-4.0 | 2,491 objects / 2,533 relationships |
| windows | [2026.11.08.dataset.4](../../fixtures/reference-packs/dataset-v1/windows/descriptors/ba65ca6288fbb96720afc1c95104a581c68bbcce7545171547b20d8a87776427.json) | `cartulary.windows_event_ids`; approved existing project-output decision, Apache-2.0 | 503 objects / no relationships |
| entra | [2026.11.08.dataset.4](../../fixtures/reference-packs/dataset-v1/entra/descriptors/ea7109980a2fd0641921268ae89c069ed08236f684e181a37f09735d526e6d15.json) | `cartulary.entra_app_ids`; approved existing output decision, Apache-2.0 | 139 applications / no relationships / zero SPN postings |
| attack | [20.1.dataset.4](../../fixtures/reference-packs/dataset-v1/attack/20.1/descriptors/de8ce9573ea446e3ba142d7672a2c56f76af8c71fe70b858656f32b6a18c9ec9.json) | `fixture.attack.dataset.v1`; restricted fixture only, production distribution blocked | 6,221 objects / 34,632 relationships |

For each row, `logical_root` contains the canonical `manifest.json`, three generic
payload members and the exact notices. `envelope_root` contains `container.pack`,
public `bootstrap.json` and `envelope-report.json`. The current online metadata
expires at **2026-12-08T19:52:00Z**; root metadata expires at
**2027-11-08T19:20:00Z**. Expiry is a verification failure, not permission to alter
an archived qualification subject. Renewal creates another envelope and descriptor.

Each production approval is retained at
`fixtures/reference-packs/dataset-v1/<pack>/approvals/<manifest_sha256>.json`.
These records preserve previously established output decisions and bind the exact
successor output. They are not new legal opinions or approvals inferred from a
signature. A signature authenticates its publisher, not the factual correctness
of source interpretation. The producer owns that correctness.

ATT&CK's unsigned production candidate is
`fixtures/reference-packs/dataset-v1/attack/release-candidate/logical/eaa1d842f23c6dd5be9d2a758ab3fe030f7fe70f1d2c068ec9ef3979bd9d8403/`.
It binds repository `cartulary.attack`, version `20.1.dataset.4`, and the actual
20.1 upstream corpus. `output-policy.pending.v1.json` remains pending. The earlier
`9dac…` candidate and fixture artifacts remain historical. To close **RI-GATE-ATTACK**,
the licensing authority must record its decision reference, exact permitted output
and notices, redistribution/embedding classification and source-profile scope.
An output-changing decision requires a new immutable profile/policy and release;
otherwise it must approve the exact current manifest. A fixture key or fixture
qualification cannot satisfy production publication. No production ATT&CK key,
bootstrap installation or distribution-readiness claim has been created.

##### Build and update procedure

The five independent source modules are `referencepacks/format`, `attack`,
`veris`, `windows` and `entra`, each with its own Go dependency manifest. The shared
format dependency is pinned at
`v0.1.0-dev.d829afb9b953015a5ba6cc2e0fe5556a41aa3e4d1002f1a76f4137a74284b45f`.
Public Make commands below run from the repository root. Discovery remains
`make help ALL=1`; substitute one explicit producer name for `<pack>`.

```sh
make bootstrap PART=go
make reference-pack-module-check MODULE=format
make reference-pack-module-check MODULE=<pack>
make reference-pack-producer-build PRODUCER=<pack>
make reference-pack-source PRODUCER=<pack> ACTION=check SOURCE_LOCK=<pinned-json> SOURCE_DIR=<explicit-source-root>
make reference-pack-package PRODUCER=<pack> ACTION=assemble INPUT=<closed-build-input.json>
make reference-pack-package PRODUCER=<pack> ACTION=envelope INPUT=<closed-release-envelope-input.json>
```

Bootstrap is the dependency-maintenance operation; review changed manifests and
checksums. It completes cache population and tidies checksums before exporting the
format module, so the pinned archive describes the final source tree. Qualification
exports each module, builds from its isolated archive with `GOWORK=off`,
`GOENV=off`, `GOPROXY=off` and `-mod=readonly`, and checks transitive/test/embed/input
closure. Builder reproduction compiles twice and compares executable bytes.
Qualified archives and `.mod`/`.info` companions are retained under
`.cartulary/reference-packs/modules/proxy/`; corresponding source hashes, archive
hashes and closure evidence live under `modules/qualified/<source_sha256>/`.
Deliver the selected source archive and its pinned dependency archives together
for offline builds. No application source, sibling producer, workspace or local
`replace` is part of a producer build. Production executables and reproduction
records are in `.cartulary/reference-packs/builders/<builder_source_sha256>/`.

The exact completed recipes are
`.cartulary/reference-packs/independence-v3/inputs/<pack>-candidate-r4.json` and
`<pack>-envelope-r4.json`; ATT&CK production uses
`attack-release-candidate-r4.json`. ATT&CK fixtures require **assemble-fixture**
and **envelope-fixture**, with fixture input types. Passing a fixture input to
`assemble` correctly fails; there is no authority fallback. Production ATT&CK
uses `assemble` and, only after RI-GATE-ATTACK, `envelope` with release approval.
All input paths are explicit. The package wrapper invokes a copied binary from
an empty directory and environment, then runs the independent machine witness.

For a source update, acquire only through the selected producer and its reviewed,
immutable lock/inventory. Use `reference-pack-source ACTION=acquire` with that
producer's explicit acquisition input; `check` validates the retained inputs and
`audit` applies source policy. Source-specific contracts and machine expectations
reside in the producer's `inputs/` and tests. Upstream Markdown is permitted only
inside explicitly pinned source acquisition/interpretation; application tests and
release witnesses consume machine artifacts. Output-affecting source/policy changes
receive new source-profile identity, pack version and monotonically reserved
release sequence. Do not edit a retained release or reset its reservation ledger.
Update complete independent factual expectations and qualify the new artifact
before changing a selector. Corpus counts alone do not establish preservation.

##### Signing custody, renewal and distribution

The three production signing contexts and the separate ATT&CK fixture context
are outside the checkout, under
`/home/jochi/.local/share/cartulary/reference-pack-signing/dataset-v1/`.
Context files are private regular files with mode 0600 in private 0700 directories.
They are local file-signer custody, not a managed HSM or externally escrowed key
service. Private bytes are absent from module archives, application inputs,
release outputs and retained test logs. Public bootstrap files are distributable.
A custodian must preserve the private contexts and their monotonically increasing
reservation history; copying public artifacts alone does not transfer signing
capability. The optional library signer interface permits a future custody adapter.

Development release ledgers now use separate `logical` and `envelopes` namespaces.
The explicit byte-preserving development transition recorded earlier does not
supply a migration for installed application history. An unrecognized ledger
layout fails closed; never erase it to force publication.

For renewal, retain the same logical root and exact-output approval, use the same
signer with `create_signing_context=false`, provide `previous_envelope`, increase
`metadata_version`, and supply explicit issue/expiry timestamps valid against the
trusted current clock. Invoke `make reference-pack-package PRODUCER=<pack>
ACTION=renew INPUT=<renewal-input.json>`; fixture ATT&CK uses `renew-fixture`.
Retain the new immutable envelope/report/bootstrap and create its immutable
release descriptor before advancing the current selector. Manifest, payload and
notice bytes must remain identical. Requalify that exact new archive; neither a
test nor a consumer silently renews expired metadata. Root rotation follows the
offline trust chain and threshold rules and needs its own review/qualification.

##### Destination trust, import, activation and rollback

Publication does not install deployment trust. An operator must authenticate the
selected public bootstrap independently, explicitly configure the destination's
trust bootstrap path and establish trusted UTC. The configuration example is
`deploy/mvp/reference-pack-administration.toml.example`; its false clock assertion
must not be treated as a default trust grant. Install the approved **production**
bootstrap for the corresponding repository. Never install the ATT&CK fixture root
as production trust. The three approved packs are independently installable; none
requires another optional pack or a producer checkout.

Validate this contract cutover in fresh isolated state. Server, operator and
Recovery admission reject incompatible retained manifests, sets, operations,
prepared imports, catalogs, snapshots and participating owner bindings before
mutation/readiness. Required Base release 3 readiness remains mandatory. There is
no predecessor reader, implicit reset or automatic historical conversion.

Copy the exact archive into the configured Reference Pack incoming directory,
verify its descriptor hashes, and use the existing administrative import route or
`operator reference-pack import <bundle_name>`. Observe the returned job/operation
to terminal success. Import leaves the pack inactive. Explicitly activate its
version through the existing administration route, resolve the current set, and
confirm the descriptor, coverage, provenance and representative lookups. Lookups
bind the returned exact set; provider/event and provider/event/version are separate
Windows keys. Entra's supported SPN lookup successfully returns no hits for this
real release; it is not evidence that an arbitrary SPN is invalid.

Disablement removes current optional availability while retaining admissible
historical pinned sets. Removal remains subject to retention/pinning rules.
Incident-bundle 8 respects the recorded distribution classification, retains
historical provenance and does not silently activate embedded packs. Backup/restore
uses the coordinated current Recovery codec. Before cutover, rollback restores
the previous application and its preserved compatible environment together. After
v3 state is accepted, an older application may run only against its preserved old
environment; it must never reinterpret new state. Database migration 50 explicitly
blocks an in-place downgrade of generic index storage. No installation, deployment,
hosted publication, trust grant or production state reset ran in this task.

##### Qualified module archive identities

These exact exports passed the independent archive/closure check after dependency
maintenance. All four producer binaries reproduced the selected `.dataset.4`
builder hashes unchanged. The format export exactly matches its pinned archive,
including maintained checksums. The version is `v0.1.0-dev.<source_sha256>`.

| Module | Source SHA-256 | Archive SHA-256 |
| --- | --- | --- |
| format | `d829afb9b953015a5ba6cc2e0fe5556a41aa3e4d1002f1a76f4137a74284b45f` | `20081930328e649d54ca6988c3d43e6709a7efd114e9a299ac405e059c69cd6e` |
| attack | `db9a03909929d336db2f2629ab9811c2132185e184cd797f99722ceb6bdc4a73` | `f58068168c6dc04d7b105e85e31410d396df519aba6d164a7f8ba1ed847502f5` |
| veris | `260fd3afd982f8348ec5e598063702bf35f80a63d9d2530a6ed48b3c2d65edc5` | `e19baa51126d003d9de31039d0b13bff94e7b75cd2fed11eca6eec9cd8865b82` |
| windows | `9a697c487ef6cd40ea6f2dc7f8ce01ec056e5ea18c42c7e789fca66641bdaf74` | `4d2d73537c686e673d2e28bbe7b6dc25f3930b349fa7be3245b268a4cd315b28` |
| entra | `8f447de6a70ead3a2d807498bfe3cbeea719c699a600732df5c06f20bda6acdb` | `e341012a9f07dd476c0089f2133e3053f45a23aad9a74b0e1d110af5b2035843` |

P0 preservation review confirms 540 of 543 original artifact paths unchanged and
the three retired Base bindings preserved byte-for-byte under
`fixtures/reference-packs/predecessor-contract-v2-projections/builtins/`. This is
a bounded immutable-artifact claim; mutable current projections and generated
clients were intentionally updated. Historical audit/run manifests were not
rewritten or promoted to final-source evidence.

#### RI-P4–RI-P6 repair and qualification accounting

The broader portfolio exposed stale cutover fixtures and one production binding:

- Base-set, manifest and contract expectations still named predecessor identities.
  They now name reviewed current machine artifacts. Retired Entra profile tests
  require rejection; generic success is proved by the real-artifact scenarios.
- Multiple Entra negative fixtures reused a release sequence, so reservation
  rejection masked content rejection. Each now has a distinct sequence.
- Shared server/operator admission fixtures retained `{}` as a terminal operation
  input. A valid current reconciliation input now permits each test to reach its
  intended incompatible record. Both owners preserve state and classify SQL
  failures separately from contract incompatibility.
- Committed-source preparation passed bundle version 7 to its version-8 descriptor.
  It now uses 8, coherently with the other source owners. All four bundle round
  trips pass, including permitted embedding and restricted references. Remaining
  version-bound source-port expectations were refreshed and their affected rows
  passed for Imports, Indicators, Links, Artifacts, Revisions and Saved Views.
- Bootstrap exported the format module before `mod download all` appended unused
  dependency-test checksums. It now tidies after cache population and exports the
  final source. The existing pin and all executable/artifact hashes are unchanged;
  isolated archive qualification confirms exact equality, including `go.sum`.

Failed roots are retained, never reused: `backend-unit-0-r1`,
`backend-integration-1-r1`, `operator-admission-r1`, `server-admission-r1`,
`portability-portfolio-r1` and the one-case `portability-diagnosis-r1`, all under
`.cartulary/test-results/reference-independence-v3-<suffix>`. The respective
`make explain-run RESULTS_DIR=<root> TARGET=test-slice DETAIL=summary|logs`
records distinguish assertion failures, the missing selector caused by a test
compile mismatch, native exits and completed cleanup. Temporary diagnostic output
used to locate the source-port failure was removed before finalization.

The `attack-replay-r4` package invocation supplied a fixture input to production
`assemble` and failed closed before publication. `attack-fixture-replay-r4` used
`assemble-fixture` and passed. All four post-signing logical replays passed the
independent corpus witness without changing manifests, payloads, notices or
reservation history. Invalid selector, oversized ROWS and stale-generated
preflight invocations were corrected through discovery/generation; they do not
count as product qualification. Full details of earlier database/publication
repairs remain in the preceding checkpoints.

Final executable freeze follows `reference-independence-v3-finalize-r2`, native
exit 0. Final graph manifests bind commit
`70ba53b47269d1733f1bdbbf310cfd4b94a7b455` plus dirty-source digest
`sha256:1c5d4a3effb22a18d003b31a3528175e24b7ccb7148165cce27b8b0dc1cedc73`.
Only this documentation handoff changes after that freeze. Earlier results with
different source identities are characterization or repair evidence, not members
of the final-source portfolio. Repeating the selected portfolio after the last
dependency-maintenance change establishes a coherent source-bound handoff.

#### RI final-source portfolio and disposition before MITRE license adoption

The final twelve `test-slice` invocations cover **109 distinct selected rows**.
Every invocation below used a fresh explicit run ID, disabled harness caching,
completed with observed native exit 0, reported pass, completed cleanup with no
failed/blocked step, and passed its retained-secret scan. These are scoped
engineering results; no complete-owner evidence-union audit or full CI/release
certification is claimed. Exact ROWS, command IDs and source/toolchain/graph
identities are in each `run-manifest.json`, with outcomes in `run-summary.json`,
row/unit results, `cleanup-results.json` and `retained-secret-scan.json`.

All roots in this table have prefix
`.cartulary/test-results/reference-independence-v3-qualified-` and suffix `-r1`.
Commands use `CARTULARY_TEST_RUN_ID=<root-name> CARTULARY_HARNESS_CACHE_MODE=off
make test-slice OWNER=<owner> ROWS=<manifest-selection>`.

| Root stem / owner | Selected obligation | Native result |
| --- | --- | --- |
| `backend-unit` / `module.reference_data` | 48 rows: neutral format/trust/security/limits, producer native correctness and publication, consumer contracts, current fixtures and source-independent expectations | 0; 19/19 work units passed |
| `backend-integration` / `module.reference_data` | 37 rows: Base readiness, absent optional packs/core behavior, generic admission/indexing, unknown-key extensibility, lifecycle, history, bounded operations, portability and failure atomicity | 0; 6/6 work units passed |
| `attack_retained_artifact` / `module.reference_data` | Exact `.dataset.4` ATT&CK archive/bootstrap, real corpus; fixture authority only | 0; separate one-row invocation |
| `windows_real_release_import` / `module.reference_data` | Exact `.dataset.4` production Windows archive/bootstrap and all 503 objects | 0; separate one-row invocation |
| `entra_real_release_import` / `module.reference_data` | Exact `.dataset.4` production Entra archive/bootstrap, 139 applications and zero real SPNs | 0; separate one-row invocation |
| `veris_real_release_import` / `module.reference_data` | Exact `.dataset.4` production VERIS archive/bootstrap and complete object/relationship counts | 0; separate one-row invocation |
| `reporting` / `module.reporting` | Four pack snapshot bindings, exact-set admission and frozen release/snapshot payloads | 0; 4/4 work units passed |
| `composition` / `module.reportcomposition` | Delegation, release-binding reason matrix and canonical digests | 0; 5/5 work units passed |
| `recovery` / `module.recovery` | Entra, VERIS and Windows backing-storage/provenance restore through the shared current codec | 0; 3/3 work units passed |
| `operator-admission` / `app.operator` | Incompatible retained state fails before mutation; current unavailable catalogs and operational SQL failures remain distinct | 0; 3/3 work units passed |
| `server-admission` / `app.server` | Startup retained-state gate, state preservation and operational failure classification | 0; 3/3 work units passed |
| `portability` / `module.incidentbundles` | All four round trips, permitted/restricted embedding, historical provenance and predecessor-bundle rejection | 0; 3/3 work units passed |

Module qualification and twice-reproduced binaries passed in
`reference-independence-v3-{format,attack,veris,windows,entra}-isolation-r3` and
`reference-independence-v3-{attack,veris,windows,entra}-release-builder-confirm-r4`.
Each uses its named public `reference-pack-module-check` or
`reference-pack-producer-build` target and retains `tool-run-summary.json` beneath
that target's directory. Complete independent factual witnesses and current-clock
standalone signature verification passed for the selected r4 outputs; four
post-signing replays passed as recorded above. Native exits were all 0.

Final `make source-check CHECK=generated`, `make source-check CHECK=schema` and
`make lint CHECK=backend-boundary` passed with native exit 0 in the respective
`reference-independence-v3-freeze-{generated,schema,boundary}-r1` roots, all on the
same frozen source digest. `make agent-finalize` passed in `finalize-r2`.
Migration 50's exact/equality/one-over storage bounds and downgrade barrier passed
in `generic-storage-migrations-r2`; those assertions and migration bytes did not
change afterward. The affected source-port rows passed in `final-imports-r1`,
`final-indicators-r1`, `final-links-r1`, `final-artifacts-r1`, `final-revisions-r1`
and `final-savedviews-r1` before the checksum-only dependency repair; they are
supporting evidence, not promoted to the final-source row union.

The finalizer's optional retained-run maintenance remained skipped because no
compatible full warm-check root was supplied. Full CI, browser/UI visual review,
deployment installation and hosted publication were not required by this backend
cutover and were not run. There is no dedicated ATT&CK backup/restore scenario in
this portfolio; its exact artifact import and bundle round trip passed, while
Recovery's shared generic codec is covered by the three named restore scenarios.
No required scoped engineering check is intentionally omitted. Human review of
specification-to-projection fidelity remains a repository responsibility.

| Remediation | Final disposition | Remaining action |
| --- | --- | --- |
| RI-R1 | CLOSED: five independent modules, four binaries, application-independent publication, isolated dependency/input closure and both-direction guards | Preserve these boundaries for future modules/releases |
| RI-R2 | ENGINEERING_DONE; production distribution BLOCKED by RI-GATE-ATTACK | Licensing authority approves exact output/notices; issue production envelope and bootstrap; independently qualify that actual signed artifact before distribution readiness |
| RI-R3 | CLOSED: optional SPNs with truthful zero real coverage and positive synthetic coverage | A populated SPN release needs a new reviewed immutable source profile and explicit shared-SPN policy |
| RI-R4 | CLOSED: historical 503-definition baseline and new exact generic release independently exercised | Future selectors require their own exact-artifact qualification |
| RI-R5 | CLOSED: one generic runtime engine; source-specific semantics and useful negative assertions retained in producers | New normalizers require reviewed engine changes; new supported declarative datasets do not |
| RI-R6 | CLOSED: coordinated major/contract cutover, projections, owner bindings, historical evidence separation and operator handoff | Preserve single-owner authority and final-source evidence distinctions |

P0/P1/P2/P4 exits are complete. P3/P5/P6 engineering, cleanup and handoff exits are
complete; their four-production-release distribution exit remains explicitly
blocked by RI-GATE-ATTACK. The three approved production packs are delivered with
exact bytes and qualification. ATT&CK is delivered as a qualified restricted
fixture plus unsigned production candidate and a tested, approval-gated publisher.
The next action is the external licensing decision, followed by production
publication/qualification for ATT&CK. No compatibility reader, implicit state reset,
undocumented data inference or fixture-to-production fallback is a workaround.

The additional final boundary row
`harness.command_surface.behavior.reference_pack_modules` passed with native exit
0, all ten cleanup steps complete and a passing retained-secret scan in
`reference-independence-v3-qualified-module-boundaries-r1`. Its archive/dependency,
embed/generation/input-discovery and strict explicit-input negative tests share
the same frozen source identity. Including this row, final selected coverage is
**110 rows across thirteen independent test-slice invocations**.

The final documentation gate is
`CARTULARY_TEST_RUN_ID=reference-independence-v3-handoff-final-r1 make lint-markdown`;
its result is retained at
`.cartulary/test-results/reference-independence-v3-handoff-final-r1/adhoc/lint-markdown/tool-run-summary.json`.
Completion reporting requires its observed native exit 0 and a clean
`git diff --check`. No executable source, release bytes or dependency manifest
changes follow the recorded freeze.

#### RI-GATE-ATTACK — MITRE-issued license and production release

The user answered the pending licensing question: **“Use the license issued by
MITRE.”** This is project authorization to apply MITRE's published license to the
retained corpus, not a claim that MITRE separately reviewed or endorsed Cartulary.
The earlier question referenced manifest `9dac8399…`; subsequent engineering had
already produced `eaa1d842…`. Neither immutable candidate was rewritten. A new
release binds the adopted license decision and current source profile.

On 2026-11-08 the official [MITRE ATT&CK® terms of use](https://attack.mitre.org/resources/legal-and-branding/terms-of-use/)
were checked against the already pinned upstream `LICENSE.txt`. MITRE grants
royalty-free research, development and commercial use on the condition that
copies reproduce its copyright designation and license. The
[branding guidance](https://attack.mitre.org/resources/legal-and-branding/)
also prohibits implying affiliation, sponsorship or endorsement. The retained
notice includes the complete copyright statement, license and disclaimers;
it is not replaced with the project's own license.

The immutable licensing source is the [selected upstream license](https://raw.githubusercontent.com/mitre-attack/attack-stix-data/f59b03ecf828579d5f9e93683ce25b7e2e5fd9d7/LICENSE.txt),
SHA-256 `738144f7fb054722a4ef9d3367c51710341dc12fc574c6ac3a41daaaecd8bf5e`.
The project decision and exact-output approval are authored machine inputs under
`fixtures/reference-packs/dataset-v1/attack/release/`, separate from the human
handoff. Product tests and release witnesses do not access either website or
this document. The approved classification is `allowed`, including distribution
of complete packs and incident-bundle embedding that retains their notices.

The successor source profile is
`referencepacks/attack/inputs/dataset-profile.v2.json`, ID
`cartulary.attack.all_domains.dataset.v2`, SHA-256
`782a47c25857afd720fb46b5e2e09143bd3f54b9d64a7a2c0a798ac2ffd37e9e`.
It records the license conditions and predecessor identity. The generic content
profile remains v1; no consumer contract or data normalization changed. The
producer admits this exact current profile. The previous source profile, fixture
releases, pending policy and unsigned candidates remain historical bytes.

The selected production artifact is
[20.1.dataset.5](../../fixtures/reference-packs/dataset-v1/attack/release/descriptors/83ec1d97ae95f08dea534ababead88cece253f88ae3433feeabf8b771390e02f.json),
using repository `cartulary.attack`, release authority, logical release sequence
3 and initial metadata version 1. Its manifest is
`d517c015e1de62ac0317f8ab16648d88077165f65d74663b343f3008f5abafa9`;
container `83ec1d97ae95f08dea534ababead88cece253f88ae3433feeabf8b771390e02f`;
payload `7e7a75265d8f8d56ed4d4aa4052e8c0d2a28316e926e7be7eac38f4a41b958e0`.
The payload, complete notice, 6,221 objects and 34,632 relationships are unchanged.
Only release/provenance/approval identities and the permitted redistribution
classification change. The independent complete-corpus witness passed.

The new independently reproduced builder is
`674d1490ab93ba58b573bb6d3df8bcc80f43c46b020579bcf7b850487d6b9b69`.
It passed isolated module/archive qualification and producer/source-model/publication
unit tests. Production signing uses a distinct private context
`/home/jochi/.local/share/cartulary/reference-pack-signing/dataset-v1/attack-release.json`;
it does not reuse the fixture authority. Custody and renewal procedures above
apply. Online metadata expires `2026-12-08T20:40:00Z`; root metadata expires
`2027-11-08T20:40:00Z`. The public bootstrap is beside the selected container.
No deployment bootstrap was installed and no hosted publication occurred.

Reproduction inputs are
`.cartulary/reference-packs/independence-v3/inputs/attack-mitre-release-r5.json`
and `attack-mitre-envelope-r5.json`. Use production actions `assemble` and
`envelope`; a renewal requires its own explicit predecessor and higher metadata
version. The application retained-artifact selector now reads this production
release directly. Its positive import test cannot substitute fixture signing or
renew the subject. Portability qualification additionally imports and embeds these
exact bytes, while preserving the separate restricted-fixture regression.

##### MITRE license gate closure and final handoff

**RI-GATE-ATTACK CLOSED; RI-R2 CLOSED; RI-P3/P5/P6 DONE.** The user-directed
license decision, complete notice preservation, distinct production signing and
exact production-artifact qualification satisfy the remaining distribution gate.
This supersedes all earlier current-state claims that ATT&CK has fixture-only
publication or pending licensing. Those earlier artifacts and evidence remain
historical. All four packs now have independently signed production artifacts
and their own import qualification; no fixture authority is promoted to production.

| Current production pack | Selected descriptor | Distribution classification |
| --- | --- | --- |
| MITRE ATT&CK® | [20.1.dataset.5](../../fixtures/reference-packs/dataset-v1/attack/release/release.v3.json) | Allowed under the retained MITRE-issued license and notice conditions |
| VERIS | [2026.11.08.dataset.4](../../fixtures/reference-packs/dataset-v1/veris/release.v3.json) | Allowed under its retained CC-BY-SA-4.0 decision/notices |
| Windows Event IDs | [2026.11.08.dataset.4](../../fixtures/reference-packs/dataset-v1/windows/release.v3.json) | Allowed under its retained project-output decision/notices |
| Entra application IDs | [2026.11.08.dataset.4](../../fixtures/reference-packs/dataset-v1/entra/release.v3.json) | Allowed under its retained decision/notices; zero real SPNs |

The ATT&CK module archive has source hash
`0a2d635363d1373bd687442252e50a2837e8876e74b3fc42d571fd8fb5d0558d` and archive
hash `a7674dc8ef853cd5de195d1af85024cf1c25b6194aa5c950e2d4f9321741f955`,
with its qualification under
`.cartulary/reference-packs/modules/qualified/<source_hash>/`. Source profile v1
is historical; the producer archive embeds only the current v2 source profile.
The three other producer modules and selected release artifacts did not change.

All following roots are under `.cartulary/test-results/` and have prefix
`attack-mitre-license-`. Every listed command completed with observed native exit
0. Final test invocations disabled caching, completed cleanup, and passed retained
secret scans. The final source identity is commit
`70ba53b47269d1733f1bdbbf310cfd4b94a7b455` plus dirty digest
`sha256:176bb8b6a5a78e1c3ebdd6ca9283c758952d1f14cd93020f1489c7c73f78addd`.

| Run suffix | Public command / qualification | Result |
| --- | --- | --- |
| `builder-r1` | `make reference-pack-producer-build PRODUCER=attack` | Two byte-identical production binaries |
| `isolation-r1` | `make reference-pack-module-check MODULE=attack` | Exported source builds independently; closure checks pass |
| `logical-r1` | `make reference-pack-package PRODUCER=attack ACTION=assemble INPUT=.cartulary/reference-packs/independence-v3/inputs/attack-mitre-release-r5.json` | Complete independent factual/notice preservation witness passes |
| `envelope-r1` | `make reference-pack-package PRODUCER=attack ACTION=envelope INPUT=.cartulary/reference-packs/independence-v3/inputs/attack-mitre-envelope-r5.json` | Production signing, standalone verification and independent trust witness pass |
| `replay-r1` | Repeat the exact logical assembly input after signing | Identical logical bytes and reservation replay |
| `finalize-r1` | `make agent-finalize` | Generated finalization passes; executable source frozen afterward |
| `final-producer-r1` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.independent_attack_producer,module.reference_data.unit.independent_attack_source_model,module.reference_data.unit.independent_format_publication` | Three selected rows / three work units pass; 12 cleanup steps complete |
| `final-import-r1` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.attack_retained_artifact` | Actual production archive/bootstrap, no re-signing, trusted current clock; 14 cleanup steps complete |
| `final-portability-r1` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.attack_pack_portability_preserves_independent_re_a273637b25` | Production bytes embed/reimport exactly; restricted fixture remains references-only; cleanup and secret scan pass |
| `generated-r1` | `make source-check CHECK=generated` | Pass |
| `schema-r1` | `make source-check CHECK=schema` | Pass |
| `boundary-r1` | `make lint CHECK=backend-boundary` | Pass |

The manifest of each test run retains the complete ROWS selection and command;
`run-summary.json`, row/unit results, `cleanup-results.json` and
`retained-secret-scan.json` record terminal evidence. Package/build/module targets
retain their named `tool-run-summary.json`. The prior 110-row remediation portfolio
remains earlier-source engineering evidence; this follow-up requalified the changed
producer/profile, delivered artifact and affected application scenarios. It does
not claim a new full-owner union or full CI run. The finalizer's optional retained-run
maintenance was skipped because no compatible full warm-check root was supplied.
No required scoped check was skipped, and no deployment or hosted publication ran.

Final documentation validation is
`CARTULARY_TEST_RUN_ID=attack-mitre-license-handoff-r1 make lint-markdown`, with its
result at
`.cartulary/test-results/attack-mitre-license-handoff-r1/adhoc/lint-markdown/tool-run-summary.json`.
Completion reporting requires its observed native exit 0 and clean
`git diff --check`. The approval, notices, selected immutable descriptor and signing
custody are concrete; there is no remaining licensing approval question for this
release under the user's adopted MITRE license decision.

## RC — Remediation closure, 2026-11-08

This section supersedes earlier current-state closure claims while preserving
all prior source identities, releases and validation records as history. The
user confirmed this is pre-deployment and authorized WS1–WS7. Pack contract v3,
Reference Pack major 4 / NLSpec 0.5.0, dataset v1, sets/reference catalogs v3 and
incident-bundle 8 remain the adopted cutover. No compatibility reader, reset,
historical-anchor rewrite, deployment, trust installation or signing-custody
change is introduced.

### RC checkpoints and assertion destinations

| ID | Dependency | Implementation and evidence destination | State |
| --- | --- | --- | --- |
| RC-P0 | Existing staged implementation | Revalidated manifest-policy callers, resource passes, producer inputs, retained selectors and prior portfolios. Index digest `595000ca0cb52f563d622bc6fa5fba8d9107e0c15151a5956cf7a655cfdbecb9` records the preserved staged baseline. | Done |
| RC-WS1 / P1 | RC-P0 | Core 00 current major/NLSpec paragraph; Core 01 REQ-01-635 and current path/source-catalog bindings; Core 04 acceptance clauses; Reporting current bundle amendment; RP-REQ-099/230/275/279. Historical adoption records retain their original context. | Closed |
| RC-WS2 / P2 | RC-P1 | `packformat.ManifestAuthority`, exact local manifest/provenance checks in `contractadmission`, consumer and retained-binding paths. `TestUnregisteredGenericDatasetLifecycle`, its two recovery cases, and `TestBaseReconciliationAtomicPublication_Integration` own unknown framework/enrichment, wrong-authority, portable-only and registry-escalation assertions. | Closed; focused and final checks pass |
| RC-WS3 / P2 | RC-P1 | Neutral `canonicaljson.ScanTokens`, bounded dataset row projection, shared semantic posting/key budgets and producer encoding. Authored object/descriptor schemas remove structural posting/coverage ceilings; semantic bounds remain enforced. | Closed; focused and final checks pass |
| RC-WS4 / P3 | RC-WS3 | `referencepacks/format/dataset/limit_matrix.json` and `TestGenericLimitMatrix`, plus existing replacement/cancellation tests. Independent token vectors live in `TestBoundedTokenViews`. | Closed; focused and final checks pass |
| RC-WS5 / P2–P3 | RC-P1 | `tools/reference-pack-modules/{isolation,isolated-go,application-isolation}.mjs`, module/build/package Make workflows and negative probes. Producer research outputs moved out of read-only source corpora. | Closed; five modules, four builders and four real corpora pass isolation |
| RC-WS6 / P3 | RC-WS2 | `TestATTACKProductionPackRestoresPinnedHistory_Integration`; shared fresh-target recovery probes; Reference Data `AdmitRestoredState` before Recovery commits authoritative rows. | Closed; exact production artifact and three rejection probes pass |
| RC-WS7 / P4–P5 | RC-P3 | Frozen-source 130-row scoped portfolio, exact production selectors, independent build/corpus qualification and final static checks below. | Closed; 130 final-source rows pass |

RC-P0 → RC-P1 → RC-P2 → RC-P3 → RC-P4 → RC-P5 remains the closure sequence.
All seven workstreams and RC-P0 through RC-P5 are complete with the final
evidence below. Earlier development runs are diagnostic history, not substitutes
for final-source rows.

### Corrected runtime and retained-history behavior

An unknown key resolves a generic content model without inventing a kind or
source authority. The admitted manifest resolves `framework` to
`framework_reference` and `enrichment` to `advisory_enrichment`. Standardized
keys and the Base registry namespace retain their policy restrictions. Local
anchors bind exact manifest and payload identities to their set member before
comparing authority. Consumers, startup, retention and restore use that check.
Incorrect local anchors fail closed; no bytes or pins are repaired.

Reference-only portable assertions may lack matching manifests. Their closed
shape and permitted advisory authority are checked, and the historical assertion
is preserved; it supplies neither destination trust nor content availability.
A locally present matching logical identity requires exact manifest agreement.
Both unknown kinds have lifecycle/recovery coverage, with portable-only binding
and escalation coverage in the canonical-persistence row.

Recovery now checks authoritative bindings and non-removed retained logical
members within its uncommitted transaction. Missing/damaged members and wrong
authority cannot commit pack rows or indexes. A failed target stays quarantined
in `restoring`; successfully staged immutable files may remain confined there.
This is not a usable partial restore or an implicit cleanup/reset. Existing
recorded-time trust checks and index rebuilding still precede readiness. The
ATT&CK test uses the actual production archive, a Reporting snapshot pin,
disabled historical content, exact descriptor/anchor/inventory comparisons,
6,221 objects and 34,632 relationships, and a removal rejection while pinned.

### Resource admission and allocation evidence

Structural admission owns canonical syntax, Unicode, duplicate members, safe
integers, closed shape, depth, descriptor/row size and component count. Semantic
admission owns normalized keys, posting budgets, actual coverage, uniqueness and
references. One lookup with 129 inputs and two lookups totaling 129 now fail
`content_semantic_invalid` with `max_dataset_object_postings`.

The token reader retains inert data as bounded raw JSON. Row materialization
retains at most 128 lookup inputs across names. It scans excess inputs without
retaining their collections so that a later structural defect still wins.
Encoded-key sizes and aggregate budgets are checked before encoded-key
allocation or posting insertion. Cancellation closes an opened stream and
leaves no published index. The private-index boundary remains application-owned.

The authored matrix checks equality/one-over for 64 KiB descriptors, 32 lookups,
four components, 128 object postings, 64 KiB keys, eight million pack postings,
256 MiB encoded keys, depth 64 and 4 MiB rows. Descriptor byte admission is
qualified separately from its closed semantic shape; no padding field was added.
Exact aggregate arithmetic uses controlled near-limit state; streaming public
admission cases exercise the shared guards, failure precedence and zero index
writes. Coverage declarations, NFC expansion, JSON escaping, normalized
collisions, replacement chains and cancellation have companion assertions.

### Isolation prerequisites and operator workflow

Qualification requires Linux namespaces and the exact bubblewrap pin in
`tools/toolchain_pins.json`. An unavailable namespace, mount capability or pinned
tool fails qualification visibly. There is no unrestricted fallback or host
reconfiguration. Static transitive import/embedding checks remain supplementary.
Negative probes cover absolute/relative reads, environment discovery, generator
inputs, symlink escape, sibling corpora, signing dependencies and writes to
read-only inputs.

Module qualification exports the selected module, mounts only its pinned Go
runtime/dependencies and creates private writable caches. `GOWORK=off`,
`-mod=readonly`, local toolchain selection and disabled network resolution apply.
No application checkout or sibling producer is visible. Application qualification
builds the `cmd` closure and regenerates neutral-format contract projections
without producer sources/corpora/private signing inputs. A completed operator
binary exercises its usage path with the documented exit 2; qualification checks
both that exit and the expected usage text. This probe is command admission, not
a claim that all operational application paths execute in the namespace.

Producer package execution receives only its binary, explicit input files,
selected corpus and private output/ledger directories. Assembly writes research
under `<output_root>/research/<manifest_sha256>/`; ATT&CK includes its source audit
there. Source acquisition/audit commands retain their separate source-management
purpose. Assembly no longer needs to modify source corpora. New builder hashes
change future candidate identities; already-valid signed releases are untouched.

Use the public workflows from the repository root:

```sh
make reference-pack-module-check MODULE=format
make reference-pack-module-check MODULE=<attack|veris|windows|entra>
make reference-pack-producer-build PRODUCER=<attack|veris|windows|entra>
make reference-pack-package PRODUCER=<pack> ACTION=assemble INPUT=<explicit-input.json>
make lint CHECK=backend-boundary
```

Successful namespace manifests are retained under
`.cartulary/reference-packs/isolation/{module,producer,application,corpus}/`.
The owning command's stderr links its exact content-addressed evidence record.
Module archive qualification also retains selected dependency pins and a recipe
hash under `.cartulary/reference-packs/modules/qualified-isolated-v1/`.
No record includes signing key contents. The common neutral dependency pin is
`v0.1.0-dev.45295465495d9a17284932765904e16b3e788239215d6d977242ff4b825cba46`,
managed through `make bootstrap PART=go`, with no local-path replacements.

The final corpus inputs are
`.cartulary/reference-packs/closure-qualification/<pack>/input-final-r1.json`.
Their output directories contain unsigned qualification candidates and independent
complete-corpus witnesses, not replacement distribution releases. All four
corpora preserve their identities, factual fields, relationships, documentation,
notices and lookup coverage against the existing machine expectations. Entra
remains 139 application objects and zero populated SPNs.

### Frozen source and selected releases

`reference-closure-finalize-r1` completed `make agent-finalize` with native exit 0.
Generated changes were inspected before freezing executable source. The final
source is commit `70ba53b47269d1733f1bdbbf310cfd4b94a7b455` plus dirty digest
`sha256:ad28e6d13bb312a7b68ea51e2da3aa33bc07b163cead8f8db2656fbb7dbe5646`.
The preexisting staged baseline remains unchanged. Documentation-only handoff
updates do not enter executable source or evidence inputs.

Production selectors remain the four `dataset-v1` descriptors listed in the
MITRE-license closure above. Their signed bytes, source profiles, distribution
decisions and custody remain unchanged. ATT&CK online metadata expires
`2026-12-08T20:40:00Z` (root `2027-11-08T20:40:00Z`); the other three online
envelopes expire `2026-12-08T19:52:00Z` (roots `2027-11-08T19:52:00Z`).
Fresh import uses the trusted current clock and must reject expired metadata.
Renewal is a separate explicit `ACTION=renew` operation with the preserved
logical root, approval, production authority, predecessor envelope and advancing
metadata versions. Retain its new immutable descriptor before selecting it for
future qualification. Tests never renew or re-sign their subject. Installation,
activation and rollback procedures above continue to apply to the selected
production releases; this work installs no deployment trust.

### Final scoped portfolio

The authored selection and pre-execution plans are retained at
`.cartulary/reference-packs/closure-qualification/portfolio.v1.json` and
`portfolio-inputs/*.plan.json`. The portfolio deduplicates the earlier 110-row
remediation coverage, ATT&CK follow-up, affected source-owner supporting rows and
new recovery obligation into 130 rows. Every invocation below uses a fresh run
identity and cache mode `off`. Exact row selections are retained in its
`run-manifest.json`. No earlier-source root supplies missing final-source coverage.

Every selected row passed with observed native exit 0, completed cleanup and a
passing retained-secret scan. All 20 run manifests bind the one frozen digest
above and match the predeclared selections exactly. This is a scoped
remediation portfolio; no full-owner evidence audit or full CI claim is made.
Optional finalizer retained-run maintenance is skipped because no compatible full
warm-check root was supplied. Hosted publication, installation and signing-key
changes remain outside scope.

All roots in this table are under `.cartulary/test-results/` with prefix
`reference-closure-final-` and suffix `-r1`. Each has `run-manifest.json`,
`run-summary.json`, `cleanup-results.json` and `retained-secret-scan.json`.

| Run stem | Selected rows | Native exit | Cleanup steps, all completed | Secret scan |
| --- | ---: | ---: | ---: | --- |
| `app-operator-1` | 1 | 0 | 14 | Pass |
| `app-server-1` | 1 | 0 | 14 | Pass |
| `harness-command_surface-1` | 1 | 0 | 10 | Pass |
| `module-artifacts-1` | 1 | 0 | 10 | Pass |
| `module-database_migrations-1` | 2 | 0 | 14 | Pass |
| `module-imports-1` | 1 | 0 | 10 | Pass |
| `module-incidentbundles-1` | 5 | 0 | 14 | Pass |
| `module-indicators-1` | 4 | 0 | 16 | Pass |
| `module-links-1` | 5 | 0 | 19 | Pass |
| `module-recovery-1` | 4 | 0 | 14 | Pass |
| `import-attack` | 1 | 0 | 14 | Pass |
| `import-windows` | 1 | 0 | 14 | Pass |
| `import-entra` | 1 | 0 | 14 | Pass |
| `import-veris` | 1 | 0 | 14 | Pass |
| `reference-unit-1` | 48 | 0 | 28 | Pass |
| `reference-integration-1` | 37 | 0 | 23 | Pass |
| `module-reportcomposition-1` | 3 | 0 | 18 | Pass |
| `module-reporting-1` | 7 | 0 | 15 | Pass |
| `module-revisions-1` | 2 | 0 | 14 | Pass |
| `module-savedviews-1` | 4 | 0 | 15 | Pass |

The final `boundary-r1`, `generated-r1` and `schema-r1` roots under the same
prefix each passed their named Make check with native exit 0. Five
`module-<module>-r1` runs passed independent exported-module qualification. Four
`builder-<pack>-r1` runs produced byte-identical binaries in two isolated builds.
Four `corpus-<pack>-r1` and four `replay-<pack>-r1` runs passed assembly and its
independent factual/notice witness with unchanged logical payload bytes. These
non-graph workflows retain their named `tool-run-summary.json`, exact commands,
stdout/stderr references and immutable namespace/archive/builder evidence. They
are separate build/corpus witnesses, not additional rows in the 130-row count.

| Producer | Final module source SHA-256 | Qualified binary SHA-256 |
| --- | --- | --- |
| ATT&CK | `c33620e8007de7567a08172815b890be038459558ddfc78a808887543b1fad24` | `d58423612ea68cfc302fd9a83f876267e67088d987d1caac535848b68c473688` |
| VERIS | `8886eb88877b20ee7beceb46ef2f3a2dddd5998b806f815c94595d64a64f7e1f` | `e71f90c8303821559975a09bf4ebbfca50d2503591dd3726cdec24d5ded96785` |
| Windows | `5ef95e728632880d289138b781aff80163a5908e2e408514cf80971d64dcbe86` | `044ba3a67f95fa7eb3cb7d184064cda7ec2c70afc410677e4b4b5eec4ae06e11` |
| Entra | `ba2291cfcafa1f650ac3fee02c24daef06ee28955eab800c97da8550869baab0` | `e5df7bb7cac158ded18eb4b50ca8a6466eaf85ce022ac5c8db4164c45f8ea938` |

The neutral module source identity is the hash in its dependency version above.
Archive descriptors and their pinned dependency qualifications are retained by
source identity and qualification recipe, so repeated qualification does not
rewrite an earlier record.

### Development failures and their disposition

All entries below are repaired development evidence. None supplies final-source
acceptance. Paths are under `.cartulary/test-results/`; graph roots retain
`run-summary.json`, and package/module roots retain the target's
`tool-run-summary.json`.

| Root / target | Finding and disposition |
| --- | --- |
| `reference-closure-format-dev-r1`, `reference-closure-format-dev-r2` / `test-slice` | Fixed a test compilation error and moved deliberate immutable-provenance corruption into an explicit test-superuser transaction. Runtime permissions and immutable triggers remain enforced. |
| `reference-closure-isolation-format-dev-r1`, `reference-closure-isolation-format-dev-r2` / `reference-pack-module-check` | Resolved the exact pinned toolchain rather than its launcher and included selected lazy-module archive metadata. No shared-cache or network fallback was added. |
| `reference-closure-application-isolation-dev-r1` through `-r3` / `lint CHECK=backend-boundary` | Added the pinned cryptographic build environment, asserted the operator's real usage exit 2, and registered the new cross-owner recovery test in the authored boundary inventory. Final boundary qualification passes. |
| `reference-closure-corpus-<pack>-dev-r1` / `reference-pack-package` | Read-only source mounts exposed research/audit writes in all four assemblers. Research moved to explicit outputs; final corpus and replay passes retain read-only source mounts. |
| `reference-closure-attack-recovery-dev-r1`, `reference-closure-attack-recovery-dev-r2` / `test-slice` | Fixed test cleanup compilation and compared retained version/envelope/release references rather than transient operation references swept by recovery. Exact descriptor, canonical content, anchors and pins remain asserted. |

The final documentation command is
`CARTULARY_TEST_RUN_ID=reference-closure-handoff-final-r1 make lint-markdown`, with
summary at
`.cartulary/test-results/reference-closure-handoff-final-r1/adhoc/lint-markdown/tool-run-summary.json`.
Completion reporting requires its observed native exit 0 and clean
`git diff --check`. No executable edit follows the recorded freeze. No required
scoped obligation or external prerequisite remains blocked. This conclusion
closes the specified remediation; it does not assert an exhaustive subsystem,
full-owner, full-CI, hosted-publication or deployment qualification.

## LC — Legacy removal and production readiness

### LC authority, scope and source posture

The user approved this document amendment after choosing **current-contract
cleanup**. This section is the next refactoring plan, not execution of its
production changes. Preserve adopted capabilities while removing proven dead
code, completed migration machinery and unnecessary input/dependency breadth.
Prefer a smaller set of clear owners over compatibility aliases or another
producer framework. No source acquisition, hosted publication, deployment,
trust installation or signing-custody redesign is included.

The preceding Plan-mode invocation performed read-only exploration and wrote
nothing. This document session supersedes that pending-write status. Existing
staged, unstaged and untracked work must be preserved; do not stage this update
or reset the working tree. Implementation authorization is **not granted by this
document-update invocation**. A later authorized implementation task must
revalidate LC-E0 and satisfy each phase's entry conditions.

| Evidence ID | Baseline / source identity | Scope and limits |
| --- | --- | --- |
| LC-E0 | `main`, HEAD `70ba53b47269d1733f1bdbbf310cfd4b94a7b455`; baseline and comparisons under `/tmp/reference-lc-docs-u5jejg2_/` | Dirty source is preserved, not described as a clean release baseline. `baseline.json` fingerprints 9,473 tracked/nonignored entries by content and identity; `index-entries.bin` and `status-before.bin` retain staged/working-tree posture. |
| LC-E0-index | Index-entry SHA-256 `9f71114aa9c7ad0a4b18047abb83171f241477a10863a9b4252d764c35534c88` | Compare exact entries after the update; this is not a claim about every Git internal file. |
| LC-E0-tracker | Incoming SHA-256 `e4fe7aadbdb0690a3f55a8a4e1acc74d2b13ccd0a26a591fadc7123759b5a830`; inode `708148`, one link; `tracker-before.md` retains the incoming document | Canonical target remains `internal/modules/reference_data`; admitted output remains this existing regular file. Output/parent components are not symlinks. |
| LC-E1 | Candidate source reads and reference searches from the preceding LC exploration; current per-path hashes and dispositions in `lc-inventory.json` | Textual absence is a removal candidate, not proof that dynamic, generated or external consumers are absent. LC-P0 supplies executable characterization. |
| LC-E2 | `tools/harness/generated-artifacts/generate-artifacts.sh`, Base/fixture conversion scripts and input declarations; source/package witness callers; application isolation launcher | Establishes active transitional paths and the generation-boundary expansion required below; not a failed product test. |
| LC-E3 | RC final portfolio and `.cartulary/test-results/reference-closure-final-reference-unit-1-r1/{run-manifest.json,run-summary.json,cleanup-results.json,retained-secret-scan.json}` | Selected retained records were inspected, not rerun. RC binds commit above plus dirty digest `sha256:ad28e6d13bb312a7b68ea51e2da3aa33bc07b163cead8f8db2656fbb7dbe5646`; no LC final-source or complete-owner claim follows. |
| LC-E4 | `ignored-baseline.json`, public Make Markdown route, installed-readiness checks, non-fixing Markdown configuration, external result/cache paths | Preservation covers relevant frontend installation/toolchain/Corepack/static-cache files and selected lock metadata. Unrelated ignored build/service/result trees are not exhaustively fingerprinted. |

The current owner remains adopted Reference Pack NLSpec 0.5.0 / extension major
4 under Core 00 precedence. RP-REQ-018 governs compatibility, RP-REQ-099 generic
authority, and RP-REQ-275 through RP-REQ-279 the neutral boundary, generic
contract and resource admission. Core 01's current bundle/consumer contracts,
Core 04 security requirements, and the participating Reporting, portability and
Recovery owners continue to apply. This plan adopts no normative amendment.
An actual owner conflict or behavior defect receives a separate finding and
correction slice; it must not be hidden inside structural cleanup.

### LC inventory and responsibility decisions

The universe is tracked and nonignored untracked files under the established
target, expanded to the four producers, neutral library and named tooling.
Ignored caches/results, external paths and unrelated repository subtrees are
outside this bounded inventory. All 532 entries in these four roots have a
per-path disposition in `lc-inventory.json`. Direct inspection means relevant
bodies were read during LC exploration, not exhaustive correctness review.

| Root | Entries | Directly inspected | Metadata-only | Next inspection boundary |
| --- | ---: | ---: | ---: | --- |
| `internal/modules/reference_data` | 226 | 9 | 217 | Candidate helpers, complete caller/generated closures and surviving assertion destinations in LC-P0 |
| `referencepacks` | 276 | 8 | 268 | Current profile/input reachability before removing historical-only exported material |
| `tools/reference-pack-modules` | 18 | 6 | 12 | Generation input/output declarations and affected isolation execution closure |
| `tools/reference-packs` | 12 | 3 | 9 | Source-audit helpers and current witness assertions before dispatch reduction |

Additional direct evidence includes the old ATT&CK contract-input embedding
file, generated-artifact policy, execution topology, current owner clauses and
selected retained RC results. Metadata-only accounting is not deletion approval.
Any file reached by a proposed removal becomes required direct inspection in
LC-P0. The plan is complete for the bounded workstreams and deterministic entry
gates below; it is not an exhaustive dead-code or correctness audit of all 532
entries or the whole subsystem.

| Responsibility | Owner / disposition | Rationale and preservation rule |
| --- | --- | --- |
| Canonical encoding, envelope/trust primitives, generic validation and budgets | Neutral format module; keep | Remove abandoned local copies without exporting application lifecycle policy. |
| Authorization, Jobs, transaction publication, sets, retention, Base policy and local provenance | Reference Data/application adapters; keep | These are live application responsibilities, not accidental duplication of producer logic. |
| Source interpretation, factual correctness, licensing decisions and authority selection | Each independent producer; keep | Do not merge small policy adapters into an omnibus dispatcher or shared producer framework. |
| One-time predecessor conversion | Remove from active generation; retain necessary historical evidence | Current releases must not be reissued as a side effect of ordinary generation. |
| Current-format fixture construction and independent witnesses | Tooling; simplify and retain | Independent expectations must not be manufactured by the production encoder under test. |
| D3FEND relationships and other adopted non-generic profiles | Existing runtime owners; intentional/no_action | These are live supported features. Their retirement or redesign is outside LC. |

### LC contract freeze and findings

| Contract ID | Required preservation | Evidence / affected phases |
| --- | --- | --- |
| LC-C1 | Pack contract v3, subsystem major 4, generic dataset v1, sets/reference catalogs v3, incident-bundle 8; current consumer types, routes, authorization and recovery formats | Adopted owners above; LC-P1–P5. No compatibility reader, automatic conversion, reset or historical-anchor rewrite. |
| LC-C2 | Exact existing Base and optional production release bytes, identities, notices, approved distribution decisions and historical results | RC release selectors; LC-P0 fingerprints, LC-P2 and LC-P4 comparisons. Code cleanup changes future builder/source identities, not retained releases. |
| LC-C3 | Five independently buildable modules, four independent producers, finite generic runtime and isolated declared inputs | RP-REQ-275; LC-P1–P4. Runtime must not acquire publication, producer, corpus or private-signing dependencies. |
| LC-C4 | Verification precedence, bounded diagnostics/resources, cancellation and atomicity, exact-set provenance, portable-history limits and source correctness | Current owner requirements and RC assertion destinations; LC-P0 characterization and LC-P1–P4 regression. A signature does not establish factual truth. |

| Finding ID / classification | Observed evidence and uncertainty | Remediation / areas | Trace |
| --- | --- | --- | --- |
| LC-F1 / must_fix for LC-P1 | App declarations `packTime`, `equalHex`, `equalDigest`, `duplicateManifestFindings`, `ProjectionDigest`, `WriteUSTAR` and neutral dataset `sourceRefs` have no callers in the inspected searches. Same-named library/native functions can be live. | Prove exact declaration ownership and remove only unused copies; implementation, tests and managed pins. | LC-C1/C3/C4; LC-W1; LC-P0/P1; LC-V1 |
| LC-F2 / must_fix for LC-P2 | Normal generation invokes `base-release.mjs`, which derives a current release and builder identity from predecessor bytes and script content. | Seal existing v3 Base bytes as an immutable input; remove active cutover registration; tooling/declarations/tests. Preserve any source snapshot needed to explain the retained builder identity. | LC-C2/C3; LC-W2; LC-P2; LC-V2 |
| LC-F3 / must_fix for LC-P2 | `current-fixtures.mjs` still constructs current vectors by reading predecessor-contract-v2 vectors. | Replace conversion inputs with reviewed current-format recipes and an independent encoder; preserve explicit predecessor rejection fixtures; tooling/machine inputs/tests. | LC-C1/C3/C4; LC-W2; LC-P2; LC-V2 |
| LC-F4 / must_fix for LC-P2 | `source.sh` still calls the old witness for source audits; `package.sh` uses the dataset witness. Old native assembly/envelope/renew branches remain. The old ATT&CK Go embed has no discovered consumer; producer archive input reachability is not yet fully classified. | Reduce the old entrypoint to supported source-audit work; remove confirmed redundant executable copies and historical-only export inputs after characterization; tooling/producer packaging/tests/docs. | LC-C2/C3/C4; LC-W3; LC-P0/P2; LC-V3 |
| LC-F5 / must_fix for LC-P3 | Application isolation executes the neutral projection generator, application build and operator usage probe. It does not yet qualify the entire affected product-generation closure. | Extend existing isolation using explicit current inputs and negative probes; tooling/execution declarations/tests. This is an evidence gap, not proof of a present hidden application dependency. | LC-C3; LC-W4; LC-P3; LC-V4 |
| LC-F6 / must_fix for LC-P4/P5 | RC explicitly records scoped 130-row acceptance without a complete-owner evidence audit. Cleanup will produce new source identities. | Predeclare complete Reference Data owner plus affected cross-owner obligations; finalize/freeze/requalify and reconcile handoff; validation/release accounting/docs. | LC-C1–C4; LC-W5; LC-P4/P5; LC-V5 |

### LC workstreams and durable outcomes

**LC-W1 — Remove extraction leftovers.** Delete proven unused declarations and
their orphaned imports/files. Remove the old ATT&CK contract-input Go embed only
after its complete consumer check; independent witness inputs remain live even
if they share that directory. Retain application policy/error translation and
all adopted specialized profiles. Refresh changed format-module dependencies
through managed Make workflows, never by hand-editing dependency lockfiles.
The benefit is fewer competing implementations and clearer ownership; the risk
of leaving this gap is future fixes landing in inactive copies. No public
behavior change is intended. Exit requires no active references to deleted
symbols, passing focused assertions and independent affected-module builds.

**LC-W2 — End predecessor-driven normal generation.** Remove the one-time Base
cutover operation from ordinary generation and retain the current v3 Base release
byte-for-byte as an immutable machine input. Preserve required historical
builder/source evidence outside active generation. A general Base publisher is
deferred until a concrete release requires one. Replace predecessor-derived
fixture recipes with reviewed current-format inputs and an independent encoder;
do not use the production format implementation to manufacture its own answers.
Retain predecessor bytes only for historical evidence or explicit rejection
tests. Update authored generation ownership and execution declarations before
regenerating. This reduces compatibility burden and prevents incidental release
reissuance. Existing production bytes do not change; fixture changes require
explicit expectation review. Exit requires repeatable generation, identical Base
bytes and successful current generation with retired conversion inputs hidden.

**LC-W3 — Retire obsolete witness and producer-input paths.** Keep the old witness
only for its supported source audits; delete obsolete native publication actions
without aliases. Generic publication keeps the dataset witness. Characterize
ATT&CK's empty/nonempty audit-report behavior and the meaning of successful
source audit before restructuring dispatch: unsupported witness actions must
not silently count as evidence. Any required observable correction is a separate
finding, not an incidental refactor. Classify each producer input through current
profiles, embeddings, tests and independent witnesses. Exclude historical-only
material from current exported trees without deleting immutable evidence or
inputs reachable through older-named current profiles. Producer authority policy
stays local. This reduces unused inputs and duplicated qualification paths;
unresolved ambiguity risks false witness success or lost factual checks. Exit
requires independently passing current audits/witnesses, no obsolete callers and
complete inputs for all four exported producers.

**LC-W4 — Qualify the complete affected generation boundary.** Reuse the existing
Linux filesystem-isolation launcher and evidence model. Cover the affected
product-generation closure, keeping harness catalog discovery of producer test
metadata separate. Expose only declared current inputs, pinned tools and private
outputs; hide producer source, upstream corpora, private signing material and
retired conversion inputs. Preserve offline execution, workspace-disabled/read-only
module resolution, negative probes and visible failure when prerequisites are
unavailable. Do not add an unrestricted fallback or reconfigure the host. This
makes future filesystem dependencies detectable without a second framework.
Qualification changes; application interfaces do not. Exit requires isolated
generation/build success and intended rejection of undeclared reads, symlink
escapes and forbidden generator inputs.

**LC-W5 — Establish production-readiness evidence.** Before execution, materialize
the required portfolio from current routing: complete Reference Data owner
coverage, affected cross-owner obligations, five standalone modules, four
producer qualifications and four independent exact production-artifact
lifecycles. Preserve existing selectors and approved distribution decisions.
Final evidence must bind one frozen executable-source state rather than combine
incompatible earlier roots. The benefit is traceable readiness; the unresolved
risk is overstating accumulated scoped passes as current acceptance. Exit
requires every required obligation mapped to compatible evidence, observed
native exits, cleanup and secret-scan results, plus the complete-owner audit.

### LC slices, dependencies and entry gates

The dependency column below is authoritative. LC-P1 and LC-P2 may proceed
independently after LC-P0, but LC-P3 requires both exits. All code and tooling work
is future implementation; the document update does not start LC-P0.

| Phase / status | Depends on | Work and readiness | Principal risk / control | Binary exit / rollback |
| --- | --- | --- | --- | --- |
| LC-P0 / TODO | Current source and LC-E0 revalidation | Ready for inspection in a later authorized task. Complete per-declaration callers, generated/file-input reachability, assertion destinations, current tool selections and immutable artifact hashes. | False dead-code conclusions; inspect exact declarations and compiled/executable closures rather than names alone. | A complete bounded removal ledger, retained characterization and concrete portfolio inputs. No production mutation to roll back. |
| LC-P1 / BLOCKED | LC-P0 | LC-W1; blocked on LC-RB1 characterization. | Deleting live policy or error mapping; preserve assertion owners and phase/diagnostic behavior. | LC-V1 passes, dead references absent, affected modules independently build. Revert only this slice's changes, preserving incoming work and artifacts. |
| LC-P2 / BLOCKED | LC-P0 | LC-W2 and LC-W3; blocked on LC-RB1, including witness/input characterization. | Fixture drift, lost independent checks or Base reissuance; compare canonical bytes and hashes before accepting generation. | LC-V2/V3 pass, current recipes stable, existing releases identical and old publication paths unregistered. Restore authored recipe/tooling changes if needed, never rewrite historical outputs. |
| LC-P3 / BLOCKED | LC-P1 and LC-P2 | LC-W4; blocked on LC-RB2 integration. | Nominal isolation or accidental reliance on harness discovery; exercise the declared product closure with forbidden inputs inaccessible. | LC-V4 passes positive and negative probes. Roll back launcher/declaration changes as a unit; no unrestricted bypass. |
| LC-P4 / BLOCKED | LC-P3 | LC-W5; blocked on LC-RB3 final qualification. Finalize, inspect generated changes, freeze executable source, then run the predeclared portfolio. | Mixed source evidence or fresh substitute artifacts; enforce source compatibility and exact selectors. | LC-V5 and complete-owner evidence audit pass. Repairs invalidate affected evidence; re-finalize/freeze and requalify before closure. |
| LC-P5 / BLOCKED | LC-P4 | Reconcile findings, release links, operator guidance and coverage limitations. | A DONE label outrunning evidence; require every named gate and evidence edge. | No unresolved required row or unrecorded uncertainty. If a gate fails, retain explicit BLOCKED status and exact evidence; no readiness claim. |

No runtime/storage migration is planned. Existing incompatible retained state
continues to fail closed under the adopted cutover. Rollback restores the
appropriate application/source environment without reinterpreting new state,
rewriting historical anchors or resetting a deployment.

### LC validation and evidence obligations

All managed commands run from the repository root through public Make targets.
LC-P0 must resolve exact row/module selections from `make task-guide
OWNER=module.reference_data`, affected-owner guides and `make target-plan` before
execution. Record the concrete commands and required rows in machine-owned
portfolio inputs. Those tools, tests and release witnesses must not read or hash
this tracker, specification/adoption prose, README or other `docs/` content.
Documentation lint is a separate documentation-maintenance check.

| Validation ID / stage | Required coverage and workflow | Expected evidence / failure consequence |
| --- | --- | --- |
| LC-V0 / document update | `make lint-markdown` with installed-only prerequisites, network-disabled Corepack and external result/static-cache roots; `git --no-optional-locks diff --check -- docs/handoffs/reference_data-module-refactor-tracker.md`; final content/index/identity comparison | Native exit 0, retained tool summary, no whitespace errors and only this tracker changed. Failure keeps LC-DOC open. No product checks, formatting, generation or finalization in this session. |
| LC-V1 / characterization and post-LC-P1 | Narrow discovered `make test-slice` selections for affected format/adapter behavior; `make reference-pack-module-check` for affected modules; retain live-profile and diagnostic precedence assertions | Named surviving assertions, independent build evidence and absence of removed callers. Pure deletion does not justify tests that merely mirror removed implementation. |
| LC-V2 / post-LC-P2 generation | `make generate`, repeat generation, `make source-check CHECK=generated`, `make source-check CHECK=schema`; current recipe and predecessor-rejection assertions | Stable output, reviewed fixture expectations, unchanged Base/production hashes and no predecessor conversion dependency in the active closure. |
| LC-V3 / post-LC-P2 producers | Current source-audit, logical assembly, publication and renewal witness selections through public producer workflows; all four exported modules and their pinned corpora | Current factual/notice assertions survive; obsolete action paths have no callers; unsupported witness actions cannot provide successful evidence. ATT&CK audit-report characterization is mandatory before removal. |
| LC-V4 / post-LC-P3 boundary | `make lint CHECK=backend-boundary` and the discovered isolation/command-surface rows; application generation/build, five module builds, four producer input closures and negative probes | Offline isolated success; undeclared absolute/relative reads, environment discovery, symlink escapes, sibling corpora and signing/generator dependencies rejected. Missing prerequisites fail visibly. |
| LC-V5 / final regression | `make agent-finalize`, generated-change review and executable freeze; fresh cache-disabled `make test-slice` selections; required build/corpus qualification; boundary/generated/schema/Markdown checks; `make test-evidence-audit OWNER=module.reference_data EVIDENCE_ROOTS_FILE=<exact compatible roots file>` | One compatible final-source portfolio, complete required Reference Data owner union and separately identified affected cross-owner coverage. Replace the roots-file placeholder with the predeclared concrete artifact. Missing rows block LC-P4; full CI is not implied. |

The LC-V5 portfolio includes canonical/trust vectors; generic limits and
verification precedence; cancellation/cleanup and atomic publication; Base
readiness; adopted specialized profiles; optional-pack absence; unknown generic
framework/enrichment keys; all four exact production imports and recovery
scenarios; Reporting/Composition; permitted/restricted portability; and
server/operator retained-state admission. UI/browser obligations follow current
complete-owner routing; no visual redesign or golden refresh is introduced.
Scoped cross-owner success is not complete acceptance of those other owners.

Reuse the existing production selectors recorded in RI/RC. Fresh imports use the
trusted current clock. ATT&CK online metadata expires `2026-12-08T20:40:00Z`
(root `2027-11-08T20:40:00Z`); the other three online envelopes expire
`2026-12-08T19:52:00Z` (roots `2027-11-08T19:52:00Z`). Revalidate these identities
and dates before LC-P4. If renewal is necessary, perform a separately authorized
publication operation before qualification, retain its new immutable descriptor
and preserve logical bytes. Never renew or re-sign inside the test. Historical
restore retains its adopted recorded-time verification semantics.

### LC blockers, completion criteria and durable handoff

| Gate ID | Status | Affected phases | Exact resolution |
| --- | --- | --- | --- |
| LC-RB1 | TODO | LC-P1/P2 | LC-P0 proves callers/input reachability and surviving assertions, including ATT&CK source-audit behavior and immutable hashes. Missing characterization blocks deletion. |
| LC-RB2 | TODO | LC-P3 | LC-P1/P2 exits and declared generation input/output closure are available; isolation prerequisites pass without installation or fallback. |
| LC-RB3 | TODO | LC-P4/P5 | Predeclared required portfolio, completed finalization/freeze, compatible passing roots and complete-owner evidence audit. Existing RC passes do not resolve this gate. |

There is no unresolved LC product-scope choice, new license decision or requested
compatibility exception. A newly discovered behavior defect gets a separate
finding, owner allocation and correction slice; a true primary-owner conflict
blocks only its dependent work. Do not silently broaden LC into feature removal.

The implementation handoff must record each LC finding's disposition, actual
deleted/moved paths, surviving assertion destinations, refreshed dependency pins,
qualified module/build identities, exact production selectors, expiry/renewal
instructions, isolation inputs/prerequisites, retained evidence roots and any
skipped required checks. Preserve earlier history while replacing current-state
claims. Entra remains 139 application identities with zero real SPN postings
unless a separately adopted source release changes that factual coverage.

| Outcome | LC completion rule |
| --- | --- |
| Planning completeness | Complete for the specified bounded workstreams, dependency graph and characterization gates; metadata-only entries and broader audit limits remain explicit. |
| Slice readiness | LC-P0 ready for later inspection; LC-P1–P5 blocked until named predecessor/characterization exits. A planning DONE never means code or validation DONE. |
| Implementation authorization | Not granted by this document-update invocation. Follow the approved future plan only in a later authorized implementation task. |
| Document completion | LC-DOC becomes DONE only after LC-V0 passes and tracker-only preservation is verified. |
| Production readiness | Requires LC-P4/P5 with every required obligation and owner audit passing on compatible final source. A blocked gate remains blocked; no full-CI, hosted-publication or deployment certification follows. |

### LC document-session checkpoint

LC-DOC is DONE. The approved plan, current-state summary, finding/contract links,
slice dependencies, validation portfolio and future readiness gates are recorded.
Only this tracker has been edited. The following command passed with observed
native exit 0:

```sh
CARTULARY_PREPARATION_POLICY=installed_only COREPACK_ENABLE_NETWORK=0 CARTULARY_TEST_RESULTS_DIR=/tmp/reference-lc-docs-u5jejg2_/results CARTULARY_STATIC_ANALYSIS_CACHE_DIR=/tmp/reference-lc-docs-u5jejg2_/static-cache CARTULARY_TEST_RUN_ID=reference-lc-docs-r1 make lint-markdown
```

The tool summary is
`/tmp/reference-lc-docs-u5jejg2_/results/reference-lc-docs-r1/adhoc/lint-markdown/tool-run-summary.json`.
`git --no-optional-locks diff --check -- docs/handoffs/reference_data-module-refactor-tracker.md`
also passed. `preservation-result.json` in the LC-E0 directory confirms that only
this tracker differs among all 9,473 baseline entries; staged entries and tracker
file identity are preserved, and none of the 6,815 inspected relevant ignored
entries changed. Historical sections 13 through RC remain byte-for-byte intact.
The LC-DOC outcome does not close LC-P0–P5 or LC-RB1–3.

The completion transcription receives the same Markdown command with fresh run
ID `reference-lc-docs-final-r1`; its tool summary is under that exact run root.
Final reporting requires its observed native exit 0 and a repeated preservation
comparison after the final write. Comparison scripts and records stay outside
the repository and are not product qualification inputs.

Product tests, generation, dependency updates and `make agent-finalize` are
intentionally skipped for this documentation-only step. This does not waive any
future LC-V1–V5 obligation. No production refactor, release mutation or deployment
action was performed. The next action is a separately authorized LC-P0 baseline
and removal-ledger task, starting with source/evidence revalidation.

### LC execution amendment and sequential checkpoints

The user explicitly authorized implementation of the complete remediation plan.
The scope is current-contract cleanup, preserving adopted capabilities and
immutable releases while removing obsolete behavior and dependencies. This
amendment supersedes the earlier LC dependency table and readiness summary;
LC-F1–F6, LC-C1–C4 and their detailed rationale remain applicable.

Baseline LC-EX0: `main` at `70ba53b47269d1733f1bdbbf310cfd4b94a7b455`;
`/tmp/reference-lc-implementation-2wfeic7g/baseline.json` records content and
identity for 9,473 tracked/nonignored entries. The same directory retains exact
index entries, initial status, unstaged diff and the incoming tracker. The index
is preserved; no staging, reset or release rewrite is authorized. Managed build,
cache and result outputs are not immutable source artifacts.

| Finding | Decision / areas | Rationale, impact and completion |
| --- | --- | --- |
| LC-F1 | Remove proven unused application helpers and neutral `sourceRefs`; implementation/tests/dependencies | One active owner per responsibility; no public behavior change. Exact callers, surviving assertions and independent affected-module builds must pass. |
| LC-F2 | Seal the existing Base v3 release as an immutable input; tooling/declarations/tests/docs | Ordinary generation must not reissue releases. Preserve exact bytes and historical builder evidence; current generation must work without conversion inputs. No new Base publisher. |
| LC-F3 | Replace predecessor conversion with current-format recipes and independent encoding; tooling/fixtures/tests | Remove obsolete-format coupling without making production code its own oracle. Preserve fixture IDs and meaningful malformed cases; repeat generation must be stable. |
| LC-F4 | Correct unsupported ATT&CK source-witness success separately, then retire old publication witness paths and unused export inputs; implementation/tooling/tests/docs | Success must represent an actual check. ATT&CK retains its producer audit and independent canonical-artifact witness. Unsupported internal witness actions fail without aliases; current producer audits and all independent factual checks survive. |
| LC-F5 | Isolate the full affected product-generation chain, including downstream embeddings/projections; tooling/declarations/tests | Start with authored inputs, not copied generated answers. Reuse existing isolation; no producer/corpus/private-signing/prose/conversion dependencies, network fallback or host reconfiguration. |
| LC-F6 | Complete final-source portfolio and owner audit; tests/evidence/docs | Historical scoped passes do not establish current complete-owner acceptance. Native exits, cleanup, scan results and compatible evidence are required. |
| LC-F7 | Correct the NLSpec introduction's reconciliation v1 restatement through RP-REQ-283's v2 contract and mark Core 00's unresolved ATT&CK license statement historical; specification/docs | Editorial owner alignment prevents retired-format restoration and stale distribution blockers. Preserve adoption history; no API/storage/contract-major change or reopened license decision. Markdown validation and owner review are required. |

LC-P2 is an umbrella, not an executable combined slice. The following prerequisite
column is authoritative. No successor starts until the preceding exit and its
tracker checkpoint are complete. Failures remain BLOCKED with exact evidence;
repairs are within the authorized workstream and invalidate affected evidence.

| Workstream | Depends on | Status | Work / binary exit |
| --- | --- | --- | --- |
| LC-P0 | none | DONE | Revalidate source/owners/releases, complete removal and input ledger, assertion map, executable characterization and concrete portfolio inputs. |
| LC-P0a | LC-P0 | DONE | LC-F7 owner alignment and Markdown lint; record editorial-only scope. |
| LC-P1 | LC-P0a | DONE | LC-F1 removal, managed pins and focused assertions/builds; no remaining deleted-symbol callers. |
| LC-P2a | LC-P1 | DONE | LC-F2 immutable Base ownership and historical builder retention; unchanged release hashes and no active conversion. |
| LC-P2b | LC-P2a | DONE | LC-F3 current recipes; independent deterministic generation and canonical/trust/admission assertions. |
| LC-P2c | LC-P2b | DONE | LC-F4 audit correctness; valid nonempty success reports, preserved failures, unsupported witness rejection, actual supported witnesses. |
| LC-P2d | LC-P2c | DONE | LC-F4 obsolete branch/embed/export-input removal; current producer closures and factual checks preserved. |
| LC-P3 | LC-P2d | DONE | LC-F5 isolated complete affected generation/build, exact output comparison and negative boundary probes. |
| LC-P4 | LC-P3 | DONE | Finalize/review/freeze; fresh compatible full required portfolio, standalone modules/builds/corpora and independent exact production lifecycles. |
| LC-P5 | LC-P4 | DONE | Complete-owner evidence audit, finding reconciliation, preservation, operator guidance and validation/handoff completion. |

Current routing identifies 101 Reference Data rows: 48 backend unit, 41 backend
integration, six frontend, three browser and three measurement. Materialize five
complete target partitions, explicitly including all three explicit-only import
rows. The final evidence manifest names their child targets, not a filtered
`test-slice` as an unfiltered owner result. Separate four exact production-pack
runs establish independence beyond the combined partition. Add affected
cross-owner obligations, five standalone module qualifications and four producer
build/corpus qualifications. Recompute exact routing before the final freeze.

All managed checks use public Make targets. Narrow characterization precedes
movement; authored generation changes precede `make generate`, repeat generation,
`make source-check CHECK=generated` and `make source-check CHECK=schema`.
`make lint CHECK=backend-boundary` verifies isolation/boundaries.
`make agent-finalize` precedes final executable-source freeze and cache-disabled
qualification. LC-P5 requires `make test-evidence-audit OWNER=module.reference_data
EVIDENCE_ROOTS_FILE=<recorded-manifest>` with the concrete compatible roots file,
then Markdown validation. Tests, generators, runtime metadata and release
evidence never consume this tracker or other Cartulary prose.

Selected releases remain ATT&CK `20.1.dataset.5` and VERIS/Windows/Entra
`2026.11.08.dataset.4`. Revalidate exact descriptors/hashes and freshness before
qualification. Tests never renew or re-sign their subjects. Expiry blocks that
qualification until a separately authorized renewal; no historical bytes or
anchors are rewritten. No deployment, trust installation, source acquisition,
compatibility alias, automatic conversion/reset or signing-custody change is
included. Rollback restores only this effort's authored changes and regenerated
projections while preserving incoming work and immutable artifacts.

Planning, sequential implementation and final validation/handoff are complete
for this bounded scope. Each workstream's exit is recorded below. Completion
means scoped engineering readiness, not full CI, deployment or formal
cryptographic certification. The handoff identifies removed paths, surviving
assertions, module/dependency identities, generation inputs, exact releases/expiry,
commands, source/run identities, cleanup/scans, preservation and acceptance limits.

#### LC-P0 completion

LC-P0 is DONE. LC-EX0 preserved the incoming source/index and 733 immutable
Reference Pack artifact entries. Exact helper searches confirm only the named
application copies and neutral `sourceRefs` lack callers; live neutral/native
namesakes remain. `producer-input-ledger.json` beside the baseline accounts for
54 producer input files, with 11 historical-only export copies eligible for
removal. ATT&CK `fixtures/native-cases.v1.json` remains live through the current
profile and is retained even though it is not directly embedded. Other producer
policy, baseline, notice and native-schema inputs remain owned by their modules.
The unused old ATT&CK Go embedding has no import consumer; its JSON files remain
available to independent witnesses. Base builder source/input and existing v3
release hashes are included in the baseline.

The concrete portfolio is
`.cartulary/reference-packs/lc-qualification/portfolio.v1.json`, with exact JSON
selections under `selections/` and current owner discovery in `owner-guide.json`.
It includes all 101 Reference Data rows, independent production selections,
affected cross-owner selections and generation/module boundary obligations.
Discovery used `make help ALL=1` and `make task-guide OWNER=module.reference_data
JSON=1`, both native exit 0. Target plans will be refreshed for the final source.

`reference-lc-p0-characterization-r1` ran `make test-slice
OWNER=module.reference_data ROWS=<eight exact rows>` with cache mode `off` and
native exit 0. Its run manifest records the exact eight rows: independent ATT&CK
producer, independent format dataset/envelope, canonical manifest/trust/
diagnostics/archive and fixture manifests. The retained root is
`.cartulary/test-results/reference-lc-p0-characterization-r1`; run summary and
retained-secret scan pass and cleanup completed. `TestATTACKProducerCLI` covers
successful nonempty canonical reports, attributable failure reports, empty/no
report on unbound input and delivery failure. The old witness's ATT&CK branch
produces no independent witness for a nonempty report; LC-P2c removes that false
acceptance explicitly. No production source changed during characterization.
Surviving assertion destinations are the eight rows above plus the later
producer/native and boundary rows recorded in the portfolio.

LC-P0a now begins with the two editorial owner corrections; no runtime contract
or distribution policy is being amended.

#### LC-P0a completion

LC-P0a is DONE. The NLSpec introduction now delegates retained operation-input
versions to RP-REQ-283 instead of restating retired reconciliation v1. Core 00
marks the prior ATT&CK licensing blocker historical and links the existing MITRE
gate closure. This implements the user's editorial decision without new product
requirements, version changes or distribution approval. `reference-lc-p0a-markdown-r1`
completed `make lint-markdown` with native exit 0; summary:
`.cartulary/test-results/reference-lc-p0a-markdown-r1/adhoc/lint-markdown/tool-run-summary.json`.
Human projection-fidelity review remains the repository's review responsibility;
these checks do not claim to perform it.

The LC-P0 input ledger was refined through transitive JSON references before any
removal: the current ATT&CK profile names native fixture cases, which in turn
name three candidate-named schemas. All four files remain; the eligible count
is 11, not the preliminary 14. This resolves the naming/reachability risk without
changing the current profile. LC-P1 now starts the seven proven-unused helper
removals; the surviving neutral/native implementations remain intact.

#### LC-P1 completion

LC-P1 is DONE and LC-F1 is resolved. Removed the six named application helpers
(two complete wrapper files) and neutral dataset `sourceRefs`; live neutral
encoding/trust and producer-native helpers remain. No public behavior changed.
`make bootstrap PART=go` (`reference-lc-p1-dependencies-r1`) refreshed application
and producer pins through the managed workflow; current format source version is
`v0.1.0-dev.54251c5b6c647d7a17fe9cc98d31c5a6da334d01be9a164a0ec33b994aae7aad`.
`reference-lc-p1-focused-r1` passed six exact existing format/verifier rows with
cache disabled and native exit 0. All five `reference-lc-p1-module-<module>-r1`
invocations passed `make reference-pack-module-check MODULE=<module>` with native
exit 0. Their roots are under `.cartulary/test-results/`; graph evidence retains
selection/cleanup/scan records and module runs retain tool summaries and immutable
isolation/archive evidence. Deleted declarations have no remaining local caller.
LC-P2a now begins immutable Base ownership; retained release bytes stay unchanged.

#### LC-P2a completion

LC-P2a is DONE and LC-F2 is resolved. Base `release.v3.json` is an immutable
contract input, removed from generated-file ownership and scratch deletion.
Ordinary generation no longer invokes the historical builder. Its exact script
and input moved to `fixtures/reference-packs/predecessor-contract-v2-projections/builtins/`,
with `builder-evidence.v3.json` recording their hashes and the retained release
SHA-256 `96a42a19520de0fe32c9a88d93ebf4041cc2ce6a2daa6b5ebb5fe1337889abf3`.
Historical builder inputs are absent from scratch generation. No release was
reissued or publisher added.

`make generate` passed twice (`reference-lc-p2a-generate-r1`, `-r2`);
`make source-check CHECK=generated` and `CHECK=schema` passed in
`reference-lc-p2a-generated-r1` and `reference-lc-p2a-schema-r1`. All native exits
were 0; roots are under `.cartulary/test-results/`. The retained release matches
LC-EX0 byte-for-byte. LC-P2b starts current-format fixture recipes with independent
encoding and unchanged assertion destinations.

#### LC-P2b completion

LC-P2b is DONE and LC-F3 is resolved. `current-fixtures-recipes.v1.json` replaces
the predecessor hash/conversion input. It contains current manifest/set inputs,
unsigned metadata templates, public synthetic-key recipes, pinned expectations,
and literal deliberately malformed bytes. The independent Node encoder produces
ZIP/TAR/GZIP, canonical JSON, digests and Ed25519 signatures without production
encoding or predecessor reads. It fails on changed pinned identities instead of
silently refreshing expectations. Existing IDs and assertion destinations remain.
All current fixture files match the LC-P2b baseline byte-for-byte.

Generation passed in `reference-lc-p2b-generate-r1`, `-r2`, `-r3`; the final two
runs establish stability after preserving readable JSON field order. Scratch
generation no longer copies predecessor vectors; `reference-lc-p2b-generated-r1`
and `reference-lc-p2b-schema-r1` passed generated/schema checks. Five existing
canonical manifest, trust, diagnostic, archive-admission and fixture-manifest rows
passed in `reference-lc-p2b-focused-r1`. All native exits were 0; exact manifests
and summaries are under `.cartulary/test-results/`. The complete owner portfolio
still owns portability and broader integration acceptance. LC-P2c now begins the
source-audit evidence correction separately from obsolete tooling retirement.

#### LC-P2c completion

LC-P2c is DONE. Successful source audits now require a regular, nonempty,
canonical report with the expected producer schema, success status where defined,
no reported issues and bound source/builder digests. Producer failures preserve
their original native status and diagnostics without running a success witness.
ATT&CK source-witness admission is removed; its producer audit and independent
dataset artifact witness remain. Other source witnesses must actually complete
before publishing success evidence.

New wrapper command tests cover successful delivery, empty/missing/malformed and
wrong reports, failed status, preserved failure exit/diagnostics, unsupported
ATT&CK witness input and missing supported witness inputs. The first boundary run
`reference-lc-p2c-boundaries-r1` failed (native exit 2): a symlinked CLI skipped its
entry-point guard. After resolving the actual executable path,
`reference-lc-p2c-boundaries-r2` passed with native exit 0, completed cleanup and a
passing retained-secret scan. Historical failure evidence remains intact.

Four `reference-lc-p2c-build-<pack>-r1` producer builds and four
`reference-lc-p2c-audit-<pack>-r1` public offline source audits passed, all native
exits 0. Audit inputs are private copies of existing pinned corpora under
`.cartulary/reference-packs/lc-qualification/audits/<pack>/source/`; no acquisition
or signing occurred. Windows, VERIS and Entra retained actual independent reports
under `witnesses/`; ATT&CK retained its canonical producer audit under `audits/`.
Generation passed twice, and generated/schema checks passed in
`reference-lc-p2c-generated-r1` and `reference-lc-p2c-schema-r1`. LC-P2d now starts
obsolete publication-witness removal and reachability-proven export pruning.

#### LC-P2d completion

LC-P2d is DONE and LC-F4 is resolved. `witness.mjs` was replaced by the explicitly
bounded `source-witness.mjs`: only Windows, Entra and VERIS source audits are
supported. Native assembly/envelope/renew branches and their crypto/framing
copies are removed, with unsupported actions tested to fail without publishing
evidence. Candidate bytes must match report digests before factual witnessing.
`dataset-witness.mjs` and its independent historical canonical artifacts remain.
The unused `contracts/reference-pack-producers/attack/inputs.go` embedding is gone.

Ten exact historical input copies were removed from producer exports; their
originals remain under `contracts/reference-pack-producers/`. The eleventh,
ATT&CK `dataset-profile.v1.json`, had no identical historical original and moved
to `fixtures/reference-packs/predecessor-producer-inputs/attack/`. The preservation
ledger records exact old/new paths and hashes. Current profiles, native cases,
transitive candidate-named schemas, notices and factual expectations remain.
`reference-lc-p2d-dependencies-r1` refreshed module pins through Make bootstrap.

All five `reference-lc-p2d-module-<module>-r1` checks, four producer unit rows in
`reference-lc-p2d-focused-r1`, both command-boundary rows in
`reference-lc-p2d-boundaries-r1`, and the three supported source audits in
`reference-lc-p2d-audit-<veris|windows|entra>-r1` passed with native exit 0. These
checks build each exported closure and retain real witness evidence; final
fresh-source producer/corpus qualification remains LC-P4. LC-P3 now extends the
isolated product-generation chain through downstream embeddings/projections.

#### LC-P3 completion

LC-P3 is DONE and LC-F5 is resolved. `product-generation.json` declares the shared
ordinary/isolated sequence: neutral schemas, current fixtures, contract-family
Go/TypeScript generation, then frontend protocol projections. Existing harness
operation ownership supplies the profile; catalog discovery stays outside this
product boundary. The isolated tree starts without all 119 declared outputs,
including Reference Pack and fixture Go embeddings. It admits compiler-discovered
application/generator source, explicit machine contract inputs and 20 locally
resolved pinned Node dependency packages. No producer module, corpus, signing
context, predecessor vectors or Cartulary prose is exposed.

Output directories are private, with read-only overlays preserving authored
inputs such as the immutable Base release. The fixture encoder receives only
its script, current recipes and output directories; negative probes verify that
production encoder files and predecessor vectors cannot be read. Existing probes
cover absolute/relative/environment discovery and symlink escapes; new probes
also reject overwriting, deleting or replacing an overlaid authored input.
Module and runtime closure tests retain both dependency-direction checks.

`reference-lc-p3-generate-r1`, `-r2`, `-r3` passed generation with unchanged outputs.
`reference-lc-p3-boundary-r1` and `-r2` passed `make lint CHECK=backend-boundary`;
the second includes the stricter fixture capability scope. The complete isolated
generation reproduced all 119 outputs, built `./cmd/...` offline and passed the
binary-only operator help probe. Exact capability/output manifests are retained
under `.cartulary/reference-packs/isolation/application/` and linked from each
run's logs. `reference-lc-p3-probes-r1`, `reference-lc-p3-generated-r1` and
`reference-lc-p3-schema-r1` passed, all native exits 0. LC-EX0 comparison confirms
all 733 immutable artifact entries and the incoming Git index remain unchanged.
LC-P4 now begins finalization, source freeze and the complete fresh portfolio.

#### LC-P4 source freeze and acceptance inputs

`reference-lc-p4-finalize-r1` passed `make agent-finalize` with native exit 0.
Generated changes were reviewed; no executable source changes follow this
freeze. Final source is `70ba53b47269d1733f1bdbbf310cfd4b94a7b455` plus
`sha256:9b3503d81be4236c9425e70a0aa5997062537d3d4c16e0b6dfc456b594ca9f44`.
The concrete portfolio contains 24 invocations / 149 row executions, including
101 complete-owner rows, affected cross-owner coverage and four intentionally
separate production imports. The duplicate command-boundary selection was
removed. `portfolio.v1.json`, `plans/`, `final-source.json` and the five-partition
`evidence-roots-final.json` are retained under
`.cartulary/reference-packs/lc-qualification/`. All final executions disable the
execution cache and use fresh `reference-lc-final-…-r1` identities.

`selected-releases-final.json` records exact descriptors, logical/container/
bootstrap hashes and expiry for the selected ATT&CK `20.1.dataset.5` and three
`2026.11.08.dataset.4` releases. Hashes matched and all were fresh at
`2026-11-09T04:16:11Z`. No test subject was renewed or re-signed. Optional retained
warm-run maintenance was not selected: no compatible full warm-check root was
created or supplied merely for maintenance. This does not waive final coverage.

LC-P4 measurement gate is BLOCKED pending diagnosis:
`reference-lc-final-reference-browser-e2e-measurement-r1` returned native exit 2
with a failing test-slice summary after 295,565 ms. Other 23 final test invocations
and all 13 standalone module/build/corpus qualifications passed. The measurement
run had no concurrent task verification. LC-P5 has not started; the failed root
is preserved and no acceptance audit or complete-owner claim has been made.

#### LC-P4 measurement diagnosis and new freeze cycle

All three failed measurement rows stopped at the workload identity assertion,
before timing: the authored performance contract still named
`base-2026-10-02.1`, whereas the preserved Base and browser fixture identify
`base-2026-11-08.1`. The payload hash, pack key, browser version, repository and
activation policy agree. Correct only the stale `base_version` expectation in
`contracts/performance/reference-pack-progress.v1.json`; keep the 1,000 ms
threshold and all behavioral assertions. The failed run completed cleanup (36
steps, none failed/blocked) and passed its retained-secret scan (194 files).

The first repair generation (`reference-lc-p4-repair-generate-r1`, native exit 2)
identified the same stale constant in the authored performance schema. Align
that constant as well and regenerate its validator through Make; do not relax
the exact workload admission rule. Both failures remain retained.

LC-P4 resumes IN_PROGRESS. This authored repair invalidates the first executable
source freeze for final acceptance. Preserve all `-r1` results as historical
evidence; regenerate, finalize and freeze again, then qualify the complete
portfolio under fresh `-r2` identities. Run the measurement partition first and
alone to resolve this gate before the remaining final portfolio. LC-P5 remains
TODO until all required final-source qualifications pass.

Repair generation passed twice (`reference-lc-p4-repair-generate-r2`, `-r3`),
then `reference-lc-p4-finalize-r2` passed `make agent-finalize`; native exits 0.
Reviewed generated changes are the workload value and its bound digest in Go
and TypeScript projections plus the regenerated schema validator. The second
freeze is commit `70ba53b47269d1733f1bdbbf310cfd4b94a7b455` plus source digest
`sha256:ba3965050f504fceecd78f4908986e6d2810b5cfaf4ba7b869e1743fc2fa8845`.
All 24 refreshed target plans pass with unchanged row selection and cache off.
The three measurements passed in `reference-lc-final-reference-browser-e2e-measurement-r2`
(native exit 0, 301,712 ms), with all 36 cleanup steps completed and a passing
187-file retained-secret scan. Remaining `-r2` portfolio execution now proceeds
against that freeze. The prior freeze's native observations and selections are
preserved under `lc-qualification/*-r1.json` and their original result roots.

#### LC-P4 completion

LC-P4 is DONE. All 24 final `test-slice` invocations passed with independently
observed native exit 0: 149 row executions comprise the complete 101-row
Reference Data partition, 44 affected cross-owner rows, and four separately
executed production-artifact lifecycles. All use fresh `reference-lc-final-…-r2`
IDs, cache mode `off`, and the second frozen source digest above. Exact commands,
selections, native exits and terminal output are in
`.cartulary/reference-packs/lc-qualification/native-observations-final.json`;
the concrete portfolio remains `portfolio.v1.json` in that directory.

Five standalone module checks, four producer builds and four current-corpus
assemblies passed with native exit 0. Module/archive/dependency identities are
recorded in `module-identities-final.json`; reproducible executable identities
are in `builder-identities-final.json`. Corpus qualification writes only private
`corpora/<pack>/output-r2` and `ledger-r2` outputs. Every independent dataset
witness reports complete native preservation; payload hashes match the selected
immutable releases. Newly qualified builder identities change the private
assembly manifests, not the retained production envelopes or descriptors.

`reference-lc-final-generated-r2`, `reference-lc-final-schema-r2`, and
`reference-lc-final-boundary-r2` passed the required generated, schema and backend
boundary checks with native exit 0. All 27 final graph roots share the exact
source, toolchain and system identities; all 434 cleanup steps completed, and
every retained-secret scan passed. `lifecycle-results-final.json` records those
per-root results. Standalone module/build/corpus helpers retain their declared
tool summaries and immutable qualification/isolation records, not graph cleanup
or secret-scan reports; they supplement the fully scanned product test roots.

This checkpoint is persisted before LC-P5 begins. LC-P5 now runs the complete-owner
audit against the five exact child-target partitions in `evidence-roots-final.json`,
then reconciles preservation, operator guidance, findings and Markdown validation.

#### LC-P5 audit, disposition and durable handoff

The complete-owner audit passed in `reference-lc-final-owner-audit-r2`, native
exit 0, using the exact five entries in `evidence-roots-final.json`. The report is
`.cartulary/test-results/reference-lc-final-owner-audit-r2/test-evidence-audit/audit.json`.
It accepts all 101 required rows on the second frozen source; no fragmented
selection is represented as an unfiltered owner run. The audit's ten cleanup
steps completed and its 11-file retained-secret scan passed. Final handoff
Markdown validation passed as recorded in the LC-P5 completion checkpoint below.

| Finding / gate | Final disposition | Evidence-backed exit |
| --- | --- | --- |
| LC-F1 | RESOLVED | Seven unreachable declarations removed; live format/native owners and application policy retained. Focused tests, 48 final owner unit rows and five independent module qualifications pass. |
| LC-F2 | RESOLVED | Immutable Base is an authored machine input. Historical builder source/input preserved exactly; ordinary and isolated generation do not access them. Baseline release hash is unchanged. |
| LC-F3 | RESOLVED | Current recipes and independent encoding replace conversion. All 342 current fixture files remain byte-identical. Canonical, trust, malformed-input, predecessor rejection and portability coverage passes. |
| LC-F4 | RESOLVED | Audits require truthful nonempty reports and preserve failures. Supported source witnesses and all four independent dataset witnesses remain; obsolete publication branches, embedding and 11 export inputs are retired. |
| LC-F5 | RESOLVED | Four declared generation stages reproduce 119 outputs from authored inputs under filesystem isolation; application build, operator probe and negative access/dependency probes pass. |
| LC-F6 | RESOLVED | Fresh final-source portfolio, native exits, complete-owner audit and retained lifecycle results pass. Historical evidence is preserved separately. |
| LC-F7 | RESOLVED in authorized editorial scope | Introduction delegates to RP-REQ-283; historical Core 00 licensing status links the recorded closure. No new requirement or distribution decision. Human projection-fidelity review remains normal repository review, not a claim supplied by automated checks. |
| LC-RB1 | PASS | Exact declaration/input ledger and source-audit characterization completed before removal; the three transitively referenced candidate-named schemas and native fixture cases were retained. |
| LC-RB2 | PASS | Shared generation declaration, isolated positive build and negative boundaries qualified on final source. |
| LC-RB3 | PASS | Finalize/freeze, complete compatible portfolio and complete-owner evidence audit passed. |

The five owner partitions below all passed with native exit 0. Their run IDs are
`reference-lc-final-reference-<partition>-r2` under `.cartulary/test-results/`.
Each includes its exact child-target projection, completed cleanup and passing
retained-secret scan.

| Partition | Accepted rows |
| --- | ---: |
| `backend-unit` | 48 |
| `backend-integration` | 41 |
| `frontend-unit` | 6 |
| `browser-e2e-webserver-backed` | 3 |
| `browser-e2e-measurement` | 3 |

The other 19 test invocations contain 44 affected cross-owner rows and four
independent production imports. Cross-owner coverage includes operator/server
retained admission, Recovery, Reporting, Composition, Artifacts, Indicators,
Links, Revisions, database migrations, Imports, Incident Bundles, Saved Views,
generation and command boundaries. Independent roots are
`reference-lc-final-import-<attack|veris|windows|entra>-r2`. The exact row lists,
commands and native exits are recorded in `portfolio.v1.json`, `selections/`,
`plans/` and `native-observations-final.json` in the qualification directory;
`lifecycle-results-final.json` records every final product/check root's cleanup
and scan outcome. This is complete Reference Data acceptance plus the named
cross-owner scope, not complete acceptance of each neighboring owner.

The final source identity is the second LC-P4 freeze. Subsequent changes are
handoff prose and private bookkeeping only. No executable source repair follows
that freeze. `final-source.json` also pins system/toolchain/catalog/routing
identities. The first measurement failure and incomplete first repair generation
remain recorded above; their corrected `-r2`/`-r3` successors supply acceptance.
The earlier source-audit symlink failure remains historical characterization.

##### Removed paths, surviving assertions and module identities

`source-change-ledger.json` records the exact delta from LC-EX0, distinguishing
this effort from incoming staged/unstaged work. Removed application files are
`internal/modules/reference_data/internal/packformat/archive_writer.go` and
`duplicate_diagnostics.go`. Removed declarations in surviving files are
`packTime` (`manifest.go`), `equalHex` (`identity.go`), `equalDigest`
(`json_values.go`), `ProjectionDigest` (`shape.go`) and neutral `sourceRefs`
(`referencepacks/format/dataset/objects.go`). The two deleted wrapper files owned
application `WriteUSTAR` and `duplicateManifestFindings`. Live neutral functions
with matching names were not removed.

The old `contracts/reference-pack-producers/attack/inputs.go` embedding,
`tools/reference-packs/witness.mjs`, and
`tools/reference-pack-modules/current-fixtures-input.v3.json` are removed.
`source-witness.mjs`, `audit-report.mjs`, `source-audit.test.mjs` and
`current-fixtures-recipes.v1.json` supply the supported current behavior.
`base-release.mjs` and `base-release-input.v3.json` moved unchanged from
`tools/reference-pack-modules/` to
`fixtures/reference-packs/predecessor-contract-v2-projections/builtins/` with
`builder-evidence.v3.json`; these are historical evidence, not active commands.

The pruned export paths are ATT&CK `inputs/candidate-index.v1.json`,
`dataset-profile.v1.json`, `dispositions.v1.json`, `field-policy.candidate.v1.json`,
`source-profile.candidate.v1.json`, `source-profile.native.v1.json`,
`source-profile.native.v2.json`, and `source-profile.native.v2.r2.json`; VERIS
`inputs/source-profile.v1.json`; Windows and Entra `inputs/source-profile.v2.json`.
Ten originals remain in `contracts/reference-pack-producers/`; the unique ATT&CK
dataset profile moved to `fixtures/reference-packs/predecessor-producer-inputs/attack/`.
`preservation/producer-input-ledger.json` binds every old and retained path by
hash, and `preservation-final.json` confirms all 11 retained copies match.

Surviving assertion owners include neutral `canonicaljson`, `dataset`,
`envelope` and `publication` tests; application `internal/packformat` manifest,
archive/limit, trust, fixture and diagnostic tests; `retired_format_fixture_test.go`;
Base reconciliation/atomicity and admission/cleanup integration tests; each
producer's command/source/native tests; and the existing canonical-artifact
`dataset-witness.mjs`. New wrapper failure/report tests live in
`tools/reference-packs/source-audit.test.mjs`; isolation assertions live in
`tools/reference-pack-modules/isolation.test.mjs`. They run through the existing
command-boundary rows. Browser lifecycle and progress assertions remain in
`apps/web/e2e/reference-pack.spec.ts` and
`apps/web/e2e/measurement/reference-pack-progress.spec.ts`. No assertion was
deleted merely because its old implementation helper was removed.

| Independently qualified module | Source SHA-256 |
| --- | --- |
| format | `54251c5b6c647d7a17fe9cc98d31c5a6da334d01be9a164a0ec33b994aae7aad` |
| attack | `6feee7c809ff50f51259a8518ed6593aa2e7aa1ee916cd9eebb3e0d702b12f54` |
| veris | `3009639d02e0131ab139845344a1377917a58eb741421bdc231c6cd3b8905a31` |
| windows | `ea6392dfc6ee099151e781a04d7f5b3afb4917b5847aedbe4d853db0bc3b232d` |
| entra | `886f18e6d303309b307d6a55edcbea7d96ffb64ba02d4e6d38c2f0d1657960a9` |

Managed versions use `v0.1.0-dev.<source SHA-256>`. Application and producer
dependency manifests/sums were refreshed through Make. Exact archive hashes,
dependency lists and qualification records are linked by
`module-identities-final.json`; current qualification recipe is
`38fd591e26fe87dce3fed7a1a68e4389d64c5fbe7444cd50b5ee54e9fcc2be7c`.
All five exported modules build independently with workspace/network resolution
off and readonly dependency admission. Four reproducible builder hashes are in
`builder-identities-final.json` and their immutable builder directories. These
new builder identities do not supersede historical release-builder identities.

##### Generation and operator handoff

`tools/reference-pack-modules/product-generation.json` is the authored shared
sequence consumed by ordinary generation and isolated qualification. Its four
stages project neutral schemas, encode current fixtures, generate contract
embeddings/projections, then generate frontend protocol artifacts. Inputs come
from the active contract registry, projection/scratch declarations, current
recipes, compiler-discovered source and explicitly pinned local dependencies.
The 119 outputs are absent at isolated startup. The fixture stage sees only its
script, recipe and output roots. Immutable Base bytes are protected by readonly
overlays even where authored inputs share a generated output directory.

Final exact capability/output evidence is
`.cartulary/reference-packs/isolation/application/b2fe5be515adbd52173aadfc7653024d35ea944059e2f7537a21e8d10f5e6fd9.json`,
indexed by `generation-isolation-final.json`. It records four generation stages,
20 Node dependency packages, 119 matching output hashes and the operator probe.
The boundary check exercises the existing offline launcher and negative probes;
missing prerequisites fail instead of acquiring dependencies or reconfiguring
the host. Harness catalog discovery remains outside the product-generation
boundary. Product verification never reads Cartulary specification/adoption prose.

Ordinary maintenance remains `make generate`, repeated generation when authored
inputs change, `make source-check CHECK=generated`, `make source-check CHECK=schema`
and `make lint CHECK=backend-boundary`, with selection through current task-guide
and target-plan surfaces. Base `contracts/reference-packs/builtins/release.v3.json`
has SHA-256 `96a42a19520de0fe32c9a88d93ebf4041cc2ce6a2daa6b5ebb5fe1337889abf3`
and version `base-2026-11-08.1`. Generation consumes it; a future Base publication
requires a separately scoped publisher/release task. Do not restore conversion
or invoke the inert historical builder as an ordinary maintenance path.

Use public `make reference-pack-source PRODUCER=<pack> ACTION=audit` with its
declared source inputs. Successful producer reports must be valid and nonempty;
failed producer native statuses remain failures. The source witness supports
Windows, VERIS and Entra only. ATT&CK uses its producer audit report and the
independent dataset witness during current-corpus assembly. Removed native
assembly/envelope/renew witness actions fail explicitly; there are no aliases.
Supported producer publication operations remain available through their public
Make workflows. This cleanup adds no publisher, trust installation or signing
custody change.

##### Exact retained production selectors and preservation

The final tests use these exact descriptors, verified unchanged and fresh before
the second freeze. `selected-releases-final.json` records complete logical,
payload, container, bootstrap, report and source-profile identities and expiry.

| Pack / selected version | Descriptor path | Descriptor SHA-256 | Online metadata expiry (UTC) |
| --- | --- | --- | --- |
| ATT&CK / `20.1.dataset.5` | `fixtures/reference-packs/dataset-v1/attack/release/release.v3.json` | `70166c6c5a1671269a324781737594400cbb1ad6701841599788bdc595783a05` | 2026-12-08 20:40 |
| VERIS / `2026.11.08.dataset.4` | `fixtures/reference-packs/dataset-v1/veris/release.v3.json` | `7fea2e94bc7bd3431ebfb5d64ad209cab4298da56fc013f73dd283030093daa1` | 2026-12-08 19:52 |
| Windows / `2026.11.08.dataset.4` | `fixtures/reference-packs/dataset-v1/windows/release.v3.json` | `f45c91c7418774657ae9a31f5c2e9a445f7fd716616ddabbe2899dcacf880469` | 2026-12-08 19:52 |
| Entra / `2026.11.08.dataset.4` | `fixtures/reference-packs/dataset-v1/entra/release.v3.json` | `e01e3e25f692704a46fcb0db7d78f074cbebe1cbd6c6d3b0f44f6aaba8aeba0a` | 2026-12-08 19:52 |

Trusted roots expire on 2027-11-08 at 20:40 UTC for ATT&CK and 19:20 UTC for the
other three. At expiry, stop affected qualification and retain the failure;
request a separately authorized envelope renewal with unchanged logical content,
preserved historical envelopes/anchors and independent exact-artifact reruns.
Tests must never renew or re-sign their subjects. No renewal occurred here.
Entra still qualifies 139 objects and zero service-principal-name postings; this
is the pinned source's current coverage, not a newly inferred completeness claim.

Final comparison at `2026-11-09T04:49:57Z` found zero mismatches among 733 immutable
artifact entries and 342 current fixture files. All 11 historical export-input
copies are preserved. Incoming staged entries are unchanged with SHA-256
`9f71114aa9c7ad0a4b18047abb83171f241477a10863a9b4252d764c35534c88`.
`preservation-final.json` retains those results, and the original LC-EX0 baseline
remains available. No API, persisted format, pack v3, subsystem major 4 or generic
dataset v1 migration is required. Internal obsolete witness callers must use
the supported operations; unsupported actions intentionally receive no alias.

No required product validation is skipped. Optional warm-run maintenance was
not selected because no compatible full warm-check root was supplied or created
for that purpose. Normal human review of projection fidelity remains explicit.
Acceptance is scoped engineering readiness for this remediation: no deployment,
hosted publication, full CI or formal cryptographic certification is claimed.
Future expiry and changes to declared generation inputs require fresh scoped
qualification; they are maintenance obligations, not unresolved LC defects.

#### LC-P5 completion

LC-P5 is DONE. The complete-owner audit, finding reconciliation, exact evidence
index, operator handoff and preservation comparison are complete.
`CARTULARY_TEST_RUN_ID=reference-lc-p5-markdown-r1 make lint-markdown` passed with
native exit 0; its summary is
`.cartulary/test-results/reference-lc-p5-markdown-r1/adhoc/lint-markdown/tool-run-summary.json`.
`git diff --check` also passed. This final checkpoint closes the last workstream;
no required engineering obligation or unresolved LC blocker remains. Human
projection-fidelity review and the explicitly excluded release/deployment claims
remain outside the automated acceptance record. The tracker is the sole
controlling plan artifact, with private machine evidence indexed beside the
recorded final roots rather than promoted into requirements.

### LC review remediation checkpoints

The user authorizes reopening LC-F4 and LC-F5, durable implementation repairs,
acceptance reconciliation, final-source requalification and a completed handoff.
The original LC release/compatibility constraints remain: no production artifact
rewrite, signing/renewal, deployment, alias, migration or reset. The new baseline
is `.cartulary/reference-packs/lc-review-qualification/source-baseline.json`, with
the exact incoming index and tracker alongside it. Prior result roots remain
unchanged; their passing status does not qualify executable repairs below.

| Workstream | Dependency | Status | Work and binary exit |
| --- | --- | --- | --- |
| LC-R1 / LC-F4 | reviewed baseline | DONE | Producer-owned closed audit-report shapes; exact typed digest and invocation bindings; preserve failure diagnostics/status; reject incomplete, mistyped and mismatched reports. Focused positive/negative command tests and supported real audits must pass. |
| LC-R2 / LC-F5 | LC-R1 | DONE | Reject linked/nonregular/escaping generation outputs and compare the complete produced inventory before host reads. Add positive and negative boundary coverage, including parent symlinks and undeclared outputs; isolated generation/build must pass. |
| LC-R3 / acceptance | LC-R2 | DONE | Reconcile graph versus helper evidence under the adopted harness contract; record the actual human projection-review status separately from engineering completion. No invented scans, source bindings or human approval. |
| LC-R4 / final qualification | LC-R3 | DONE | Generate/finalize, refresh exact selections, freeze source, run the complete compatible required portfolio with cache off and fresh IDs; preserve selected production releases and independently qualify all four. |
| LC-R5 / validation and handoff | LC-R4 | DONE | Complete-owner audit, final preservation comparison, native/lifecycle evidence index, updated findings and operator handoff, Markdown validation. Close engineering work only when all applicable gates pass. |

Each workstream begins IN_PROGRESS and exits with a persisted checkpoint before
the next starts. A failed required gate is BLOCKED until repaired and qualified.
Repairs after final source freeze require new compatible acceptance evidence.
Human projection-fidelity review is not performed by an automated test; its
status must remain explicit even when engineering handoff is complete.

#### LC-R1 completion — LC-F4 resolved

Binary exit: **PASS**. Complete producer-owned schemas now govern successful
audit delivery. Windows reuses its existing reconciliation definition; ATT&CK,
VERIS and Entra own closed schemas in their exported `inputs/` trees and validate
successful command output before delivery. Make projects standalone JavaScript
validators for the source helper; audit execution needs no installed AJV package.
These tooling projections are separate from application product generation.

The wrapper snapshots the exact selected executable, inventory/lock, policy,
profile, applicable editorial input and ATT&CK timestamp before invocation.
Inventory-based producers bind their retained acquisition lock back to the
selected inventory; VERIS uses its existing canonical lock digest. Strict report
types reject digest arrays, missing members, unknown fields and malformed nested
values. Schema-conformant reports with mismatched invocation identities fail.
Producer failures retain their original exit and diagnostics, including when the
secondary context check also fails. Supported factual witnesses remain required;
ATT&CK source witness and retired publication actions remain unsupported.

Changed paths and before/after hashes are recorded in
`.cartulary/reference-packs/lc-review-qualification/r1-source-changes.json`.
No public report format, producer action, release or runtime contract changed.
`make bootstrap PART=go` completed twice without dependency-lock changes.
`make format`, two `make generate` invocations, schema validation, ATT&CK's
exported-module qualification, four producer builds and four actual source audits
passed. Native observations are retained in this evidence directory.

Focused graph acceptance: `reference-lc-review-r1-boundaries-4` passed with native
exit 0, completed cleanup and retained-secret scan pass. This covers complete
report shapes, identity mismatches, failed-producer precedence, missing witness
inputs and unsupported actions. `reference-lc-review-r1-generated-2` and
`reference-lc-review-r1-schema` passed with native exits 0. Actual audit roots are
`reference-lc-review-r1-audit-attack` and
`reference-lc-review-r1-audit-{veris,windows,entra}-2`, all native exits 0;
independent reports remain under the corresponding private `audits/` trees.

Historical failures remain retained: the first inventory-based audit attempts
rejected an incorrect inventory-as-lock expectation; the corrected implementation
follows the producer's existing lock semantics. Boundary attempt 2 exposed a test
variable-scope error, repaired before attempts 3/4 passed. The first generated
check detected missing scratch declarations for the new tooling generator; exact
schema/script inputs were added before attempt 2 passed. No failed result is used
for acceptance. Final frozen-source qualification remains LC-R4's obligation.

The user confirmed that human review of the earlier specification edits is still
pending. This is an explicit handoff limitation, not an automated acceptance claim.
LC-R2 has not begun at this checkpoint.

#### LC-R2 completion — LC-F5 resolved

Binary exit: **PASS**. `generation-outputs.mjs` validates the complete writable
product-output inventory before reading content. Declared files and every parent
must be contained regular paths; symbolic links, hard links, special files,
missing outputs, extra files and extra directories fail. Content reads use
no-follow opens and descriptor identity checks. Inputs sharing writable output
directories must remain byte-identical. Generation no longer receives the binary
output directory as a writable capability.

`application-isolation.mjs` now invokes this trusted comparison in a separate
read-only sandbox with only the authored/current comparison files and private
produced tree. Generator stages still cannot access checked-in answers. Host
qualification code receives validated hashes from stdout rather than reading
paths created by a generator. The retained isolation record includes the exact
comparison capability. Paths/hashes are in
`.cartulary/reference-packs/lc-review-qualification/r2-source-changes.json`.

`reference-lc-review-r2-boundaries-2` passed the focused command row, including
sandbox-generated dangling links to byte-identical host answers, linked parents,
hard links, FIFO rejection, missing/extra outputs, byte drift, authored-input
mutation and path traversal. `reference-lc-review-r2-isolation-2` passed
`make lint CHECK=backend-boundary`: complete isolated product generation, exact
output comparison, application build, operator probe and dependency boundaries.
Both native exits were 0, cleanup completed and retained-secret scans passed.
The earlier boundary attempt failed because a negative test removed an empty
directory without the required recursive option; that fixture cleanup error was
fixed before the successful run. Both historical and successful roots remain.
No host reconfiguration, network fallback or release rewrite occurred.

LC-R3 has not begun at this checkpoint. Final-source qualification remains open.

#### LC-R3 acceptance reconciliation

The adopted Testing Harness §5.3 and its independent-module qualification
section distinguish graph-backed product evidence from helper-only producer
operations. The final acceptance ledger will apply those actual contracts:

| Obligation | Required retained evidence | Acceptance limit |
| --- | --- | --- |
| Complete Reference Data owner, cross-owner rows, four independent exact-release lifecycles, generated/schema/boundary checks | Compatible final graph source/selection; native exit 0; passing row/target results; completed cleanup; passing retained-secret scan | Only exact selected roots participate in the owner audit; fragments do not masquerade as an unfiltered owner invocation. |
| Five exported-module checks, four producer builds, four current-corpus assemblies | Native exit 0 and successful public tool summary; exact exported source/archive/dependency identities; byte-identical builder evidence or independent corpus witness as applicable | Helper summaries are not graph run manifests, do not supply graph cleanup receipts, and do not publish retained-secret scans. No such fields or results are fabricated. These supplement the graph-backed release evidence. |
| Four successful source audits added for this repair | Native exit 0 and public tool summary; complete invocation-bound audit report; the three supported independent source witnesses | ATT&CK's source report is producer evidence; its factual artifact witness is exercised separately. Source auditing does not grant publication or trust authority. |
| Human projection-fidelity review | A human's recorded disposition of the existing RP-REQ-283 cross-reference and Core 00 licensing clarification | **PENDING**, explicitly confirmed by the user. Automated checks cannot close this item. |

The earlier plan's blanket cleanup/scan wording is thus narrowed to the
life-cycle evidence that the adopted graph harness actually provides. It is not
a claim that helper-retained archives or reports were separately secret-scanned.
Helpers retain their declared source/artifact evidence and ordinary private-work
cleanup behavior; no new helper lifecycle protocol is introduced by this repair.
All exact-production-artifact acceptance runs still require the full graph
cleanup and retained-secret-scan gates. Historical results remain intact.

Final execution must freeze source after `make agent-finalize`, refresh routing
and selections, retain the selected immutable descriptors and expiry checks, and
record helper source/artifact identities alongside the final graph identity.
A repair after freezing invalidates affected acceptance and requires a new
finalize/freeze cycle. Human review remains a handoff obligation even after all
engineering gates pass; no deployment, hosted release or cryptographic
certification is implied.

LC-R3 binary exit: **PASS** for evidence-contract reconciliation. The human review
item remains explicitly pending. `reference-lc-review-r3-markdown` completed
`make lint-markdown` with native exit 0. The applicable evidence categories,
limits and fresh final obligations are recorded above; LC-R4 has not begun at
this checkpoint.

#### LC-R4 first freeze — repair required

`reference-lc-review-r4-finalize` passed with native exit 0. The first frozen
source was `sha256:73e5e363298031604b79ab51fa0168366ffc7a2c83d7f7d2a0a7e498ab0148ec`.
Complete-owner discovery still contains exactly 101 rows, with no missing or
extra selections. All 24 target plans passed. Selected production descriptors,
container/bootstrap/report/manifest hashes and expiries matched their retained
identities; no renewal is needed.

The final boundary gate failed with native exit 2: the new output validator also
rejected a legitimate hard-linked comparison input installed by the frontend
artifact cache (`client-asset-set-manifest.json`). Private generated outputs must
remain single-link regular files; read-only reference inputs may share an inode
with the immutable build cache. LC-R4 is BLOCKED while that distinction is repaired
and tested. Earlier focused output-isolation passes remain historical. Any graph
runs under this first freeze will not be used for acceptance after the repair;
a new finalize/freeze cycle and compatible qualification are required.

The LC-F5 repair supplement passed: `reference-lc-review-r2-boundaries-3`
explicitly accepts hard-linked read-only comparison inputs while rejecting
hard-linked generated outputs. `reference-lc-review-r2-isolation-3` passed the
complete isolated generation/build/operator and backend-boundary checks. Both
native exits were 0. LC-R4 resumes with a new finalize/freeze cycle; the first
freeze is retained as `first-freeze-source.json` and `first-freeze-content.json`
and is excluded from final acceptance.

#### LC-R4 second freeze — deterministic deadline test repair

The second frozen source was
`sha256:278359eb3a2a3f8212f6dc0c2fb709ea94e3da82cf26b0b90f778f6fc6c83d16`.
Standalone modules/builds, all four corpus assemblies/source audits and final
schema/generated/backend-boundary checks passed. However, the complete backend
unit partition failed (`reference-lc-review-final-reference-backend-unit-r2`,
native exit 2) in the existing supervised-cleanup deadline test. Its 40 ms wall
clock starts before workspace creation; under host scheduling delay it expires
before the directory exists, while the assertion requires an existing empty
directory. This is an untested early-expiry path mistaken for a cleanup failure,
not evidence of retained scratch or an LC-F4/LC-F5 production defect.

LC-R4 is BLOCKED while the test is made deterministic using Go's virtual test
clock for that deadline case, with an explicit assertion that a workspace write
occurred. Production timeouts and cleanup behavior will not change. This source
repair again requires finalize/freeze and compatible final graph evidence.

`reference-lc-review-cleanup-deadline` passed the exact affected storage/cleanup
row with native exit 0. The deadline now expires under `testing/synctest` only
after the workspace-write hook blocks; the test proves both a write and
`context.DeadlineExceeded`, then checks complete scratch removal. Other cleanup
cases and production code are unchanged. LC-R4 resumes with the third
finalize/freeze cycle. All superseded run roots remain historical.

#### LC-R4 completion — final-source qualification

Binary exit: **PASS**. `make agent-finalize` completed as
`reference-lc-review-r4-finalize-3` with native exit 0. Its changes were inspected
before freezing executable source:
`sha256:4e1cb8113d7b0f36e9d96ea95c9610d67a28bb2807c922f39fd68fab722b7f1f`
(commit `70ba53b47269d1733f1bdbbf310cfd4b94a7b455`, preserved dirty checkout).
No executable source changed afterward. The only additional acceptance repair
was the deterministic cleanup deadline test documented above.

The final portfolio has 43 observed successful native command exits: 26
compatible graph roots and 17 helper qualifications. Every graph root uses the
frozen source and execution cache `off`, has a passing summary, completed cleanup
(420 completed actions in total), and a passing retained-secret scan. Helpers
have their adopted tool/artifact evidence and the explicit LC-R3 limits.

All 101 Reference Data rows passed in the five exact partitions: backend unit 48,
backend integration 41, frontend 6, browser administration 3 and measurement 3.
The three explicit-only production-import rows are included. Each of ATT&CK,
VERIS, Windows and Entra also passed a separate exact-production-artifact run.
Affected cross-owner selections, five standalone modules, four reproducible
producer builds, four current-corpus assemblies and four source audits passed.
Final generated, schema and backend-boundary checks passed. The generated check
also supplies the exact two cross-owner generation rows: refreshed plans proved
identical membership, so no redundant `harness-generation` invocation was added.

Final runs use `reference-lc-review-final-<name>-r3`. The exact commands, native
sessions/exits and all refreshed target plans are retained in
`.cartulary/reference-packs/lc-review-qualification/native-observations-final.json`.
`portfolio.v1.json`, `plans/`, `selections/`, `evidence-roots-final.json`,
`lifecycle-results-final.json`, `helper-results-final.json`,
`module-identities-final.json`, `builder-identities-final.json`,
`source-audit-identities-final.json` and `final-source.json` provide the bounded
qualification index. Module source/archive identities match current exported
inputs; all builder pairs reproduce byte-for-byte. Final audit reports match the
selected builders and retain the three supported independent witnesses.

All first/second-freeze roots and their failures remain historical and are
excluded from acceptance. No required execution was skipped; the duplicate
generation selection is satisfied by compatible exact-row evidence, as recorded
in the portfolio. LC-R5 has not begun at this checkpoint: the complete-owner
accounting audit, final preservation comparison and final handoff are still due.
Human projection-fidelity review remains pending.

#### LC-R5 validation and handoff

The complete-owner command
`make test-evidence-audit OWNER=module.reference_data EVIDENCE_ROOTS_FILE=.cartulary/reference-packs/lc-review-qualification/evidence-roots-final.json`
passed as `reference-lc-review-r5-owner-audit`, native exit 0. Its authoritative
result is
`.cartulary/test-results/reference-lc-review-r5-owner-audit/test-evidence-audit/audit.json`.
It accepts all five child partitions and all 101 required rows, with no rejected
or unused roots, against the third final-source identity. The audit itself also
has completed cleanup and a passing retained-secret scan.

| Finding / uncertainty | Current disposition | Final acceptance evidence |
| --- | --- | --- |
| LC-F4 — source-audit report acceptance | RESOLVED | Producer-owned complete schemas, exact invocation bindings, preserved failure precedence; final command-boundary row and all four current source audits pass. Three supported factual source witnesses remain; ATT&CK retains producer and independent artifact evidence separately. |
| LC-F5 — isolated generation output acceptance | RESOLVED | Complete contained regular output inventory and authored-input preservation are checked inside a read-only comparison sandbox; linked/missing/extra/special outputs fail. Final boundary qualification reproduces outputs, builds the application and runs the operator probe. Read-only comparison files may use immutable-cache hard links. |
| Deadline cleanup validation uncertainty found during qualification | RESOLVED | Virtual-clock deadline case proves a workspace write, deadline expiry and scratch cleanup; the complete final backend unit partition passes. Production behavior/timeouts are unchanged. |
| Helper versus graph evidence | RECONCILED | LC-R3 records the adopted distinct contracts. Seventeen helper qualifications have observed native success and their tool/artifact evidence; they are not claimed to publish graph cleanup receipts or retained-secret scans. All 26 product/check graph roots do publish passing lifecycle evidence. |
| Human specification projection-fidelity review | PENDING | The user explicitly confirmed that review has not occurred. No automated result or engineering completion represents human acceptance of the earlier specification edits. |

`preservation-final.json` confirms all 733 immutable release/evidence files, all
342 current fixture files and all 11 preserved historical producer inputs match
the original LC baselines. The incoming staged index remains byte-identical
(SHA-256 `9f71114aa9c7ad0a4b18047abb83171f241477a10863a9b4252d764c35534c88`).
The review repair changed 29 paths, recorded individually with before/after hashes
in `source-change-ledger-final.json`; no path was removed by this repair.
Earlier LC removals/moves and surviving assertion destinations remain documented
in the historical handoff above. Managed dependency files and retained production
artifacts are unchanged by this review repair.

Selected releases remain ATT&CK `20.1.dataset.5` and VERIS/Windows/Entra
`2026.11.08.dataset.4`. `selected-releases-final.json` records exact descriptor,
manifest, container, bootstrap and envelope-report hashes. Online metadata expires
2026-12-08 at 20:40Z for ATT&CK and 19:52Z for the other three; root metadata expires
2027-11-08 at 20:40Z and 19:20Z respectively. No renewal/re-signing occurred.
Expiry must block future qualification until separately authorized renewal with
unchanged logical content; tests must not renew their subjects. Base release bytes
and readiness expectations remain unchanged.

Operator and maintainer handoff:

- Rebuild a selected producer through `make reference-pack-producer-build
  PRODUCER=<pack>` after changing its owned inputs. Update audit schemas under
  the producer's `inputs/` tree (Windows uses the existing reconciliation
  definition), then run `make generate` for tooling validators. Never hand-edit
  generated validator code or dependency lockfiles.
- `make reference-pack-source PRODUCER=<pack> ACTION=audit SOURCE_DIR=<retained-root>
  SOURCE_LOCK=<selected-input>` must receive an explicit selection. ATT&CK selects
  its source lock and explicit `BUILT_AT`; the other producers select inventories
  whose retained acquisition locks must bind those exact bytes. Windows may
  receive the explicit `EDITORIAL` used for the audited corpus. Mismatched
  report/input/builder identities fail and must be investigated, not bypassed.
- Source-audit success requires the complete report and every supported factual
  witness. Retired witness actions and ATT&CK source-witness admission have no
  compatibility alias. The independent canonical-artifact witness remains part
  of current corpus qualification.
- Expand product generation through its existing authored machine declarations,
  including the complete output inventory. Regenerate through Make and qualify
  with `make lint CHECK=backend-boundary`; extend meaningful negative cases when
  adding new capabilities. Generators never receive checked-in answers; the
  separate trusted comparison receives read-only reference files afterward.
- Preserve `native-observations-final.json`, the five-root owner manifest, final
  source/content inventories, module/builder/audit identities, lifecycle/helper
  results, and the complete-owner audit artifact when handing this work onward.
  Superseded and failed roots remain historical. Future executable repairs require
  a new finalize/freeze and compatible evidence; do not combine incompatible roots.
- A human still needs to review the RP-REQ-283 cross-reference and Core 00's
  historical ATT&CK licensing clarification, and record that disposition here.
  This is documentation review, not a new product requirement or licensing decision.

No required engineering execution was skipped. Full CI, deployment, hosted
publication, trust installation, signing-custody redesign and formal cryptographic
certification are outside this scoped engineering handoff.

LC-R5 binary exit: **PASS** for scoped engineering acceptance and handoff.
`reference-lc-review-r5-markdown` passed `make lint-markdown` with native exit 0;
its summary is
`.cartulary/test-results/reference-lc-review-r5-markdown/adhoc/lint-markdown/tool-run-summary.json`.
`git diff --check` also passed. The final checkpoint is followed by a final
Markdown validation recorded as `reference-lc-review-r5-markdown-final` in
`handoff-validation.json`. No required engineering gate or implementation blocker
remains. Human review remains explicitly pending and is the next external
handoff action; it does not authorize changing retained releases or reopening
the existing licensing decision. No deployment/publication claim is made.
