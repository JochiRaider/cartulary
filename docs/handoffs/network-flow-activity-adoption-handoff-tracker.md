# Network Flow Activity Adoption Handoff Tracker

## 1. Purpose, scope, and non-scope

This document is a resumable implementation-control tracker for moving
`docs/network-flow-activity-nlspec.md` from `draft` toward an adoption-ready
state. It coordinates owner decisions, derived contracts, generated artifacts,
implementation work, fixture bytes, executable conformance evidence, validation,
and the final status transition. It is not normative authority. Every behavior
decision remains with the cited owner document.

The tracker follows this authority order:

1. Adopted subsystem NLSpecs within their owned boundaries.
2. Core 00 through Core 04 for current implementation-conformance behavior.
3. The draft Network Flow Activity NLSpec for the proposed extension boundary.
4. Core 05 only for separately scoped claim-bearing publication.
5. `docs/domain.md` and implementation/testing guidance for vocabulary and mechanics.
6. Current code and tests as implementation evidence, not behavioral authority.
7. Research and earlier planning artifacts as supporting evidence only.

If an adopted owner contradicts the draft, record `BLOCKED: owner contradiction`,
cite both seams, and obtain an owner decision. If code differs from an owner,
record drift without treating the code as authority. Behavior-affecting closure
must proceed in this order: owner amendment; derived contract; generator and
generated outputs; implementation; fixtures and executable evidence; drift
validation; adoption/status transition.

This tracker does not authorize:

- silently changing Network Flow behavior or treating the draft as adopted;
- implementing unrelated Network Flow features;
- converting extension resources into Core records, system views, or saved views;
- adding time buckets, raw IPFIX, live collection, binding-read routes, or other
  future-only behavior;
- using research, this tracker, or implementation state as normative authority;
- making performance-publication claims unless Core 05 is separately activated.

Only this tracker was in scope for the session that created it. Subsequent
execution sessions MUST use this file as the controlling ledger and MUST commit
the completed workstream artifact plus a tracker checkpoint before beginning the
next workstream.

## 2. Session and repository snapshot

### 2.1 Snapshot

| Field | Observed value |
| --- | --- |
| Date/time | `2026-07-10T00:03:46-04:00` |
| Repository root | `/home/jochi/code/cartulary` |
| Applicable instructions | `AGENTS.md`; no deeper `AGENTS.md` exists under `docs/handoffs` |
| Branch | `main` tracking `origin/main` |
| Commit | `65d7461a5e70fb3e13515f224612f8ba470b2c6a` |
| Dirty-tree summary before execution | Clean; `git status --short --branch` returned only `## main...origin/main [ahead 1]` |
| Tracker path | `docs/handoffs/network-flow-activity-adoption-handoff-tracker.md`; committed by `65d7461a` |
| Network Flow owner proposal | `docs/network-flow-activity-nlspec.md`; `status: draft`; `document_version: 1.0.0-draft.1`; `contract_major: 1` |
| Network Flow inspection checksum | SHA-256 `e6d9b0120bf0f1c76b5469572e367ccd04249cce9c0513c833586a85218b6260` |
| Framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md`; SHA-256 `bc40b9f161a7fb3bf49cc696ed491351054ba2a3994883785d1f2dda66205e20` |
| Domain vocabulary | `docs/domain.md`; SHA-256 `6d57edca2cd10cfe1b2d503aff53cf4eb00edf87febb773449029cb0d9823b2d` |
| Reconciled owner counts | 7 Table 1-B dependencies; 12 gates; 17 blockers; 28 fixtures; 107 acceptance criteria |
| Current Network Flow code/tests | `TODO: source not found`; no repository path matched Network Flow implementation, test, fixture, contract, schema, manifest, or generated-artifact names |
| Current timezone source | Official IANA `tzdata2026c` release identified; no vendored, signed, policy-owned artifact exists yet |

The commit and checksums above freeze what was inspected. They are not adopted
document versions or immutable dependency locators for Table 1-B. Every Table
1-B locator remains an adoption blocker until the owner documents and the draft
name the adopted versions and exact sections or schemas.

### 2.2 Owner-source snapshot

| Source | Current status/version | SHA-256 | Relevant seam |
| --- | --- | --- | --- |
| `docs/spec/00_document_set_status_and_precedence.md` | Authoritative current Core; no `document_version` field observed | `f173b6ed43030e79812047288529d52d144bb46c2dd49391e90013b807b6b432` | §§4.2, 4.3, 5, 5.1 |
| `docs/spec/01_architecture_storage_and_view_contracts.md` | Authoritative current Core; no `document_version` field observed | `2e25ba55283120785dd1493a5bdc98c61c27194a2298182069a30ebafce5eb67` | §§3.3.3.1, 3.3.6, 3.3.7, 3.3.9.1, 17.1, 17.2 |
| `docs/spec/02_domain_model_schema_and_history.md` | Authoritative current Core; no `document_version` field observed | `9a868704fc6c28a72db4d83629efd42a00c600cde6eac4bb1babe872a0b1faf7` | §§10.2, 14, 15, 18 |
| `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Authoritative current Core; no `document_version` field observed | `aa652fa8abf4c8411944de2d5b7d73d125c24c91d74ad919b67b3ea85041c4e2` | §§2, 4.3.1 |
| `docs/spec/04_security_deployment_and_conformance.md` | Authoritative current Core; no `document_version` field observed | `f2251886bd165b299ff82b971b3e683c0fe03ad6ede6b0fe8df02b085acba768` | §§2, 3, 9, 12 |
| `docs/graph_projection_nlspec.md` | `adopted/current`; no `document_version` field observed | `50873a3cbd6774f854fac1b822aa50a7b7c3a2844310423adae90491cac6223e` | §§4–5, 10, 13–14 |
| `docs/testing-harness-nlspec.md` | `adopted/current`; `cartulary.testing_harness.current.v1` | `1026095055720fc658a4303f91da607c2fba58217f52a48803520f15a2e68e52` | §§4, 8, 11, 12, 16, 17 |

### 2.3 Current adjacent implementation and contract evidence

| Area | Observed paths | Current evidence limit |
| --- | --- | --- |
| Extension discovery | `internal/platform/httpapi/extensions.go`; `internal/modules/extensions/api.go`; `internal/modules/extensions/routes.go`; `contracts/extensions/index.json` | Closed to five current profiles; no Network Flow entry |
| Public contracts | `contracts/openapi/cartulary.openapi.yaml`; `contracts/errors/index.json`; `contracts/ws/index.schema.json` | No Network Flow route, DTO, error, or invalidation contract |
| Import owner boundary | `internal/modules/imports/targets.go`; `internal/modules/imports/owner_apply.go`; `internal/modules/imports/ownerfacade/finalize.go`; `internal/modules/imports/tabularingest/` | Current result is record and `view_row_v1` shaped |
| Indicator behavior | `internal/modules/indicators/store.go`; `internal/modules/indicators/import_create.go` | Implementation accepts `ipv4_addr`; Core does not designate the required canonical IP-literal token |
| Graph Projection | `internal/modules/graphprojection/`; `contracts/graph-projection/conformance_matrix.v1.json`; `contracts/graph-projection/fixtures/corpus.v1.json` | Retained lifecycle implementation only; no ephemeral adapter boundary |
| Harness controls | `internal/platform/httpapi/testclock.go`; `internal/testutil/testruntime/public_error_fault.go`; `internal/testutil/testruntime/reset.go` | Test clock and one-shot public error exist; required commit/worker/randomness controls do not |
| Contract generator | `contracts/index.json`; `tools/contractgen/main.go`; `tools/contractgen/families.go`; `tools/contractgen/validation.go`; `tools/harness/generated-artifacts/generate-artifacts.sh` | Generator families are registry-owned; Network Flow is declared as a planned family and is not emitted until authored contracts activate it |
| Generated outputs | `internal/gen/contracts/`; `packages/protocol-ts/src/generated/`; `packages/ui-contracts/src/generated/` | Generated roots; never hand-edit |
| Generated policy and drift | `tools/generated_artifact_policy.json`; `tools/generate_drift_scratch_inputs.json`; `tools/harness/generated-artifacts/check-generate-drift.sh` | Existing Make-owned generation and drift mechanics |

### 2.4 Discovery commands and limits

Exact discovery commands run from the repository root included:

```text
rg --files -g 'AGENTS.md' -g '!node_modules' -g '!vendor' .
rg --files docs/spec docs | rg '(^docs/spec/|graph_projection_nlspec|testing-harness-nlspec|network-flow|handoffs)'
rg -n --hidden -g '!node_modules' -g '!vendor' '(Network Flow|network[-_ ]flow|NF-(GATE|BLOCK|FIX|AC)-|tzdb-2026c)' .
rg --files . | rg -i '(network[-_]?flow|netflow|ipfix|fixture|contract|schema|generator|conformance)'
rg -n -i '(extension profile|import owner|target_kind|terminal result|indicator|purge|invalidation|cursor|audit|retention)' docs/spec/0[0-4]_*.md
rg -n -i '(ephemeral|adapter|arbitrary-precision|dependency_error)' docs/graph_projection_nlspec.md
rg -n -i '(fixture manifest|failure inject|fake clock|authorization transition|audit.count|generated contract|drift)' docs/testing-harness-nlspec.md
rg -n -i '(tzdb|tzdata|zoneinfo|time/tzdata|Unicode 17)' go.mod go.sum internal apps packages tools contracts configs docs
make help-all
```

The scan did not exhaust every unrelated module, every migration, research file,
browser scenario, or external timezone-distribution source. External tzdb
provenance and licensing were not researched in this session. Existing paths in
the tables were directly observed; proposed Network Flow paths remain explicit
`TODO:` values.

## 3. Authority and source map

| Dependency | Owner file and exact section/schema | Current version/status | Immutable locator present? | Network Flow interface imported | Conflict/drift | Tracker work items |
| --- | --- | --- | --- | --- | --- | --- |
| Core 00 | `docs/spec/00_document_set_status_and_precedence.md` §§4.2, 4.3, 5, 5.1; `REQ-00-003`, `REQ-00-064` | Adopted current-profile Core 00 revision; owner artifact `155b5f64` | Yes; Network Flow Table 1-B locator closed by `a5dc9847` | Extension ownership, precedence, adopted-document registry | Final claimability/status flip remains coordinated under `NFA-ADOPT-001`; v1 no-private-purge posture is explicit | `NFA-C00-001`, `NFA-C00-002`, `NFA-LOC-001` |
| Core 01 | `docs/spec/01_architecture_storage_and_view_contracts.md` §§3.3.3.1, 3.3.6.1, 3.3.6.2, 3.3.7, 3.3.9.1, 3.3.10.1, 17, 17.2; `REQ-01-542..548`; `REQ-01-618..620d`; `REQ-01-240..242`; `REQ-01-151.1` | Adopted current-profile Core 01 revisions; owner artifacts `89580f0c`, `90401fb2`, `08fa716e` | Yes; Network Flow Table 1-B locator closed by `a5dc9847` | Discovery, analytical import target/result, owner facade, unit of work, envelopes, terminal result | Owner seams are adopted; generated contracts, implementation, and fixtures remain later | `NFA-C01-001..005`, `NFA-LOC-001` |
| Core 02 | `docs/spec/02_domain_model_schema_and_history.md` §§10.2, 14.4, 18; `REQ-02-074A..074C`, `REQ-02-210`, registry token row for `indicator.indicator_type` | Adopted current-profile Core 02 revision; owner artifact `344486e7` | Yes; Network Flow Table 1-B locator closed by `a5dc9847` | Canonical IP indicator identity, create/dedupe transaction participation, and no-private-purge boundary | Exact IP tokens and participant exist; Network Flow binding implementation and fixtures remain later | `NFA-C02-001..003`, `NFA-LOC-001` |
| Core 03 | `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` §§2, 4.3.1; `REQ-03-004`, `REQ-03-011A`, `REQ-03-030`, `REQ-03-067..072` | Adopted current-profile Core 03 revision; owner artifact `08fa716e` | Yes; Network Flow Table 1-B locator closed by `a5dc9847` | Extension tab and current-authorization resource invalidation | Owner artifact defines claimed extension workspace identity and generic extension-resource invalidation; generated contracts, UI implementation, and browser evidence remain later | `NFA-C03-001`, `NFA-C03-002`, `NFA-LOC-001` |
| Core 04 | `docs/spec/04_security_deployment_and_conformance.md` §§2, 3, 9.1B, 12.3; `REQ-04-123..142`, `AC-475..477`; Core 01 §3.3.7 for current cursor wire ownership | Adopted current-profile Core 04 revisions; owner artifacts `3b942fe0`, `90401fb2`, `cd645750`, `71258589`, `663c8684` | Yes; Network Flow Table 1-B locator closed by `a5dc9847` | Route authorization, cursor protection, safe digest, audit, key lifecycle, retention | Owner contracts define authorization, cursor security, digest lifecycle, audit occurrence, and retention/no-purge semantics; implementation and fixtures remain later | `NFA-C04-001..005`, `NFA-LOC-001` |
| Graph Projection NLSpec | `docs/graph_projection_nlspec.md` §§4, 5.1.1, 10.0, 10.9, 12, 13, 14; `GP-AC-033`, `GP-AC-053`, `GP-AC-069` | `adopted/current`; owner artifacts `4e446354`, `f177fb6b`, `81941bba` | Yes; Network Flow Table 1-B locator closed by `a5dc9847` | Ephemeral request, mapping, result, dependency outcome | `project_ephemeral`, no private direct-copy members, exact adapter mapping/outcome boundary, and 69/36 evidence alignment are closed | `NFA-GP-001..003`, `NFA-LOC-001` |
| Testing Harness NLSpec | `docs/testing-harness-nlspec.md` §§8, 11, 12, 16, 17; `TH-HARNESS-REQ-657..663`; `TH-HARNESS-AC-049..055`; Network Flow schemas | `adopted/current`; profile v1 | Yes; Network Flow Table 1-B locator closed by `a5dc9847` | Generated contracts, fixture manifests/execution, fault controls, evidence accounting | Harness capabilities and structural accounting exist; fixture bytes, generated outputs, Phase 12 maps, and executable product rows remain later | `NFA-TH-001..007`, `NFA-LOC-001` |

## 4. Top-level adoption work tracker

Allowed statuses are `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and
`DROPPED`. `DONE` requires a committed named artifact and retained validation
evidence. The original control artifact is committed; later rows become `DONE`
only through the artifact-plus-checkpoint protocol in §6.1.

| ID | Work item | Workstream | Status | Depends on | Owner | Owner IDs | Repository paths/symbols | Required artifact or evidence | Validation command | Exit condition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `NFA-INV-001` | Repository and source inventory | discovery | `DONE` | none | tracker maintainer | `NF-AC-106` | Sources in §2 | Snapshot committed by `1bb6fdbd`; validation in §15.1 | `git status --short --branch` | Snapshot is committed and attributable |
| `NFA-AUTH-001` | Authority, dependency, and contradiction crosswalk | contracts | `DONE` | `NFA-INV-001` | owner reviewers | all gates and blockers | §§3, 5 | Crosswalk committed by `1bb6fdbd`; exact 7/12/17 accounting | `make lint-markdown` plus §15 checks | Every dependency and owner ID maps to tasks |
| `NFA-WS-000` | Normalize the controlling execution ledger | `WS-00` | `DONE` | `NFA-INV-001`, `NFA-AUTH-001` | tracker maintainer | tracker ACs | This file | Artifact `1bb6fdbd`; retained Make results in §15.1 | Commands in §15 | Stale creation state is corrected and all execution workstreams are explicit |
| `NFA-C00-001` | Decide extension ownership and adoption model | Core 00 | `DONE` | `NFA-AUTH-001` | Core 00 | `NF-GATE-001`, `NF-BLOCK-001`, `NF-BLOCK-010` | Core 00 §§4.2, 4.3, 5, 5.1 | Owner amendment `155b5f64`; §15.2 evidence | `make lint-markdown`; structural owner checks | Core recognizes the bounded profile without premature adoption |
| `NFA-C00-002` | Record future-only incident purge and coordinate the final registry flip | Core 00 | `BLOCKED` | `NFA-C00-001`, `NFA-ADOPT-001` | Core 00 | `NF-GATE-001`, `NF-BLOCK-001` | Core 00 §§4.3, 5.1 | Explicit v1 non-purge decision plus coordinated final Core/NLSpec amendment | `TODO: owner-specific target not found` | Current v1 makes no purge claim and final adoption is atomic |
| `NFA-C01-001` | Extension discovery resource and reserved route root | Core 01 | `DONE` | `NFA-C00-001` | Core 01 | `NF-GATE-002`, `NF-BLOCK-002`, `NF-BLOCK-010` | Core 01 §3.3.3.1; `contracts/extensions/index.json` | Owner amendment `89580f0c`; §15.3 evidence; generated/fixture evidence remains in later rows | `make lint-markdown`; `make json-shape-check` | Claimed and unclaimed behavior is owner-defined; generated output remains gated by `NFA-GEN-*` |
| `NFA-C01-002` | Analytical import target/result union | Core 01 | `DONE` | `NFA-C00-001` | Core 01 | `NF-GATE-003`, `NF-GATE-004`, `NF-BLOCK-003`, `NF-BLOCK-011` | Core 01 §17.2; `internal/modules/imports/targets.go` | Owner amendment `89580f0c`; §15.3 evidence; implementation remains in later rows | `make lint-markdown`; `make json-shape-check` | `network_flow_table` is owner-admitted and not coerced into a Core record/view |
| `NFA-C01-003` | Opaque stream, preview/apply owner facade, and source-change check | Core 01 | `DONE` | `NFA-C01-002` | Core 01/imports | `NF-GATE-003`, `NF-BLOCK-011` | `REQ-01-618..620`; imports module | Owner amendment `89580f0c`; §15.3 evidence; implementation remains in later rows | `make lint-markdown`; `make json-shape-check` | No path/URL leakage and source changes are owner-defined to fail closed |
| `NFA-C01-004` | Cross-owner atomic unit of work | Core 01 | `DONE` | `NFA-C01-003`, `NFA-C02-002`, `NFA-C04-004` | Core 01 | `NF-GATE-007`, `NF-BLOCK-011`, `NF-BLOCK-012`, `NF-BLOCK-014` | Import/indicator/idempotency/audit owners | Core 01 owner side in `89580f0c`; C02 participant in `344486e7`; C04 audit participant in `71258589`; harness faults remain in later rows | `make lint-markdown`; `make json-shape-check`; `make backend-store` | Core 01 transaction boundary, Core 02 indicator participant, and C04 audit occurrence participant are defined; implementation and harness faults remain later |
| `NFA-C01-005` | Terminal result publication and cancellation recovery | Core 01 | `DONE` | `NFA-C01-002`, `NFA-C01-004` | Core 01/jobs | `NF-GATE-004`, `NF-BLOCK-003`, `NF-BLOCK-011`, `NF-AC-107` | Core 01 §3.3.9.1, §17.2 | Owner amendment `89580f0c`; worker/recovery implementation remains in later rows | `make lint-markdown`; `make json-shape-check` | Exactly-once terminal publication is owner-defined; worker-fault evidence remains gated by `NFA-TH-*` and `NFA-IMPL-*` |
| `NFA-C02-001` | Canonical IP indicator token and canonicalization | Core 02 | `DONE` | `NFA-C00-001` | Core 02/indicators | `NF-GATE-008`, `NF-BLOCK-009`, `NF-BLOCK-012` | Core 02 §§10.2, 18; indicators module | Artifact `344486e7`; §15.4 evidence | `make lint-markdown`; `make backend-unit`; `make backend-store` | `ipv4_addr` remains valid, `ipv6_addr` is adopted, and canonical IP vectors pass |
| `NFA-C02-002` | Indicator find/create/dedupe transaction participation | Core 02 | `DONE` | `NFA-C02-001` | Core 02/indicators | `NF-GATE-007`, `NF-BLOCK-012` | `internal/modules/indicators/` | Artifact `344486e7`; §15.4 evidence | `make backend-store` | Indicator owner participant creates/reuses and rolls back inside the caller transaction; Network Flow binding routes remain later |
| `NFA-C02-003` | Network Flow-specific incident purge cascade | Core 02 | `DROPPED` | Core future incident-removal profile | Core 00/Core 02 | future generic cascade obligation | Core 00 §4.3; Core 02 §§14–15 | Recorded decision that v1 does not invent a private purge boundary | `rg -n -e 'future-only' -e 'purge' docs/spec/00_document_set_status_and_precedence.md docs/network-flow-activity-nlspec.md` | A future generic Core cascade can admit Network Flow without a v1 compatibility promise |
| `NFA-C03-001` | Extension-contributed top-level tab | Core 03 | `DONE` | `NFA-C00-001`, `NFA-C01-001` | Core 03 | `NF-GATE-005`, `NF-BLOCK-004` | Core 03 §2; workbook shell | Artifact `08fa716e`; §15.5 evidence; generated contract and browser implementation remain later | `make lint-markdown`; `make json-shape-check` | Base built-in list remains unchanged and extension workspace identity is owner-defined |
| `NFA-C03-002` | Extension-resource invalidation topics and UI consequences | Core 03 | `DONE` | `NFA-C03-001`, `NFA-C01-001` | Core 03; Core 01 wire owner | `NF-GATE-009`, `NF-BLOCK-013` | Core 03 §4.3.1; Core 01 §3.3.10.1 | Artifact `08fa716e`; §15.5 evidence; C04 route authorization, generated WS contracts, UI, and fixtures remain later | `make lint-markdown`; `make json-shape-check` | Rename/delete/auth loss invalidation semantics are owner-defined |
| `NFA-C04-001` | Network Flow route-family authorization | Core 04 | `DONE` | `NFA-C00-001`, `NFA-C01-001` | Core 04 | `NF-GATE-006`, `NF-BLOCK-005` | Core 04 §2 | Artifact `3b942fe0`; §15.6 evidence; route hooks and fixtures remain later | `make lint-markdown`; `make json-shape-check` | Current membership/role and no-`deployment_admin` bypass are owner-defined |
| `NFA-C04-002` | Cursor confidentiality, integrity, TTL, and key rotation | Core 01/Core 04 | `DONE` | `NFA-C04-001` | Core 01 wire; Core 04 security | `NF-GATE-010`, `NF-BLOCK-014` | Core 01 §3.3.7; Core 04 §§2, 12 | Artifact `90401fb2`; §15.7 evidence; implementation and fixtures remain later | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `git diff --check` | Common envelope, opaque sealed token, 15-minute TTL, and rotation lifecycle are owner-defined |
| `NFA-C04-003` | Safe-digest secret and key-ID lifecycle | Core 04 | `DONE` | `NFA-C04-001` | Core 04 | `NF-GATE-010`, `NF-BLOCK-014` | Core 04 §12 `secret_ref_v1` | Artifact `cd645750`; §15.8 evidence; implementation and fixtures remain later | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `git diff --check` | Every digest carries key ID without secret disclosure |
| `NFA-C04-004` | Transactional audit occurrence semantics | Core 04 | `DONE` | `NFA-C04-001` | Core 04 | `NF-GATE-010`, `NF-BLOCK-014` | Core 04 §3 | Artifact `71258589`; §15.9 evidence; implementation and fixtures remain later | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `git diff --check` | Counts and no-audit replay behavior are exact |
| `NFA-C04-005` | Network Flow soft-delete and raw-source retention boundary | Core 04 | `DONE` | `NFA-C04-004` | Core 04 | `NF-GATE-010`, `NF-BLOCK-014` | Core 04 §§9.1B, 12.3 | Owner amendment `663c8684`; §15.10 evidence; implementation and fixtures remain later | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `git diff --check` | Soft-deleted data, import-source cleanup, retained counts, and no current purge claim are owner-defined |
| `NFA-C05-001` | Optional claim-publication boundary for large-limit timing | Core 05 | `DEFERRED` | explicit separate claim scope | Core 05 | `NF-AC-050` | Core 05; `NF-FIX-008` | Only required if timing becomes publication evidence | `make benchmark-claim-check` | Remains engineering-only unless separately activated |
| `NFA-GP-001` | Ephemeral Graph Projection invocation | Graph Projection | `DONE` | `NFA-AUTH-001` | Graph Projection | `NF-GATE-011`, `NF-BLOCK-015` | Graph Projection §§4–5, 10, 12–13 | Owner amendment `4e446354`; §15.11 evidence; adapter mapping and evidence repair remain later | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `git diff --check` | Invocation allocates no retained view/run |
| `NFA-GP-002` | Exact adapter mapping and outcome contract | Graph Projection | `DONE` | `NFA-GP-001` | Graph Projection/Network Flow | `NF-GATE-011`, `NF-BLOCK-015` | Graph Projection §§4–5; Network Flow §14.4 | Owner/downstream amendment `f177fb6b`; §15.12 evidence; evidence repair remains later | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `git diff --check` | Exact metadata and counter strings validate without leakage |
| `NFA-GP-003` | Close pre-existing Graph Projection evidence drift | Graph Projection | `DONE` | `NFA-GP-002` | Graph Projection/harness | `NF-GATE-011`, `NF-BLOCK-015` | GP spec, matrix, corpus, JSON shape checker | Authored evidence/checker artifact `81941bba`; §15.13 evidence | `make json-shape-check`; `make lint-scripts`; `make generated-artifact-policy-check`; `git diff --check` | Owner spec, matrix, corpus, and validator agree |
| `NFA-TH-001` | Immutable fixture-manifest schema and execution | Testing Harness | `DONE` | `NFA-AUTH-001` | Testing Harness | `NF-GATE-012`, `NF-BLOCK-016` | Harness §§8, 11, 16–17 | Artifact `b3f46bfd`; §15.14 evidence; fault/clock/randomness/auth/audit/drift controls remain separate rows | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `git diff --check` | Manifest schema, owner refs, sorted paths, per-file hashes, bundle hashes, run-local materialization, and unlisted-file rejection are enforced |
| `NFA-TH-002` | Final-commit and worker fault injection | Testing Harness | `DONE` | `NFA-TH-001`, `NFA-C01-004` | Testing Harness | `NF-GATE-012`, `NF-BLOCK-016` | Harness §12 test controls | Artifact `ecd4edd5`; §15.15 evidence; fake clock/randomness/auth/audit/drift controls remain separate rows | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `make backend-unit`; `git diff --check` | Named import/worker boundaries, one-shot faults, correlation scoping, conflict handling, and reset clearing are harness-owned |
| `NFA-TH-003` | Fake-clock coverage | Testing Harness | `DONE` | `NFA-TH-001` | Testing Harness | `NF-GATE-012`, `NF-BLOCK-016` | `internal/platform/httpapi/testclock.go` | Artifact `30880992`; §15.16 evidence; randomness/auth/audit/drift controls remain separate rows | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `make backend-unit`; `git diff --check` | Clock set/reset/state responses are schema-owned, runtime reset clears registered clock state, and Network Flow time evidence is owner-routed |
| `NFA-TH-004` | Deterministic CSPRNG and collision injection | Testing Harness | `DONE` | `NFA-TH-001` | Testing Harness | `NF-GATE-012`, `NF-BLOCK-016` | Harness §§8, 12, 16–17 | Artifact `84eb7fb5`; §15.17 evidence; auth/audit/drift controls remain separate rows | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `make backend-unit`; `make backend-integration`; `git diff --check` | IDs/collisions are deterministic without production weakening |
| `NFA-TH-005` | Authorization-transition and hidden-resource controls | Testing Harness | `DONE` | `NFA-TH-001`, `NFA-C04-001` | Testing Harness | `NF-GATE-012`, `NF-BLOCK-016` | Harness §§8, 12, 16–17; Core 04 route authorization owner | Artifact `43b0fe36`; §15.18 evidence; audit/drift controls remain separate rows | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `make backend-unit`; `make backend-integration`; `git diff --check` | Auth loss and hidden resources are tested at route time |
| `NFA-TH-006` | Audit-count and no-audit replay assertions | Testing Harness | `DONE` | `NFA-TH-001`, `NFA-C04-004` | Testing Harness | `NF-GATE-012`, `NF-BLOCK-016` | Harness §§8, 12, 16-17; Core 04 audit occurrence owner | Artifact `d96f9ad2`; §15.19 evidence; generated-contract and drift accounting remain separate rows | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `make backend-unit`; `make backend-integration`; `git diff --check` | Exact occurrences and zero replay increments are proven |
| `NFA-TH-007` | Generated-contract, structural-lint, and drift accounting | Testing Harness | `DONE` | `NFA-TH-001`, `NFA-GEN-001` | Testing Harness | `NF-GATE-012`, `NF-BLOCK-007`, `NF-BLOCK-008`, `NF-BLOCK-016` | `tools/network_flow_activity_accounting.json`; `tools/schemas/cartulary.network_flow_activity_accounting.v1.schema.json`; Harness §§8, 16–17 | Artifact `0eaa8db3`; §15.22 evidence; generated outputs remain planned and fixture/Phase 12 rows remain later | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `make generate-drift`; `git diff --check` | Network Flow structural accounting fails closed through public validation targets |
| `NFA-LOC-001` | Adopted versions and immutable locators for Table 1-B | dependency registry | `DONE` | non-final owner amendment tasks; final adoption flip remains `NFA-ADOPT-001` | all dependency owners | `NF-BLOCK-010`, `NF-AC-106` | Network Flow Table 1-B; `tools/network_flow_activity_accounting.json` | Artifact `a5dc9847`; §15.23 evidence; final adoption flip remains later | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `make generate-drift`; `git diff --check` | No Table 1-B `TODO:` dependency cell remains and structural accounting enforces that state |
| `NFA-TZ-001` | `tzdb-2026c` provenance, license, revision, and digest | timezone | `DONE` | `NFA-AUTH-001` | Network Flow/harness | `NF-BLOCK-017`, `NF-FIX-022`, `NF-AC-087` | `contracts/network-flow/timezone/tzdb-2026c.provenance.json`; Network Flow §9.7 timestamp ruleset; CP-27 | Artifact `42338a1d`; §15.20 evidence; timestamp transition fixture bytes remain `NFA-FIX-022` | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `git diff --check` | Fold/gap expectations have one immutable source ruleset; fixture transitions remain later |
| `NFA-GEN-001` | Adopt Network Flow authored contract/generator ownership | generated contracts | `DONE` | all non-final owner amendments | contract owners | `NF-BLOCK-007` | `contracts/index.json`; `tools/contractgen/families.go`; `contracts/network-flow/` | Artifact `d8dbaf3f`; §15.21 evidence; Network Flow family is `planned` and generated outputs remain byte-stable | `make lint-markdown`; `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-scripts`; `make harness-contract`; `make backend-unit`; `make generate-drift`; `git diff --check` | Generator family and source path are explicit |
| `NFA-GEN-002` | Author derived schemas/contracts | generated contracts | `DONE` | `NFA-GEN-001`, `NFA-LOC-001` | contract owners | `NF-BLOCK-007` | `contracts/network-flow/index.json`; route/error/schema bundle; JSON shape checker | Artifact `9bb6711f`; §15.24 evidence; generated outputs remain blocked until `NFA-GEN-003` | `make json-shape-check`; `make harness-contract`; `make lint-scripts`; `make generated-artifact-policy-check`; `make generate-drift` | Every public object is closed and owner-derived |
| `NFA-GEN-003` | Regenerate Go and TypeScript outputs | generated contracts | `DONE` | `NFA-GEN-002` | generator | `NF-BLOCK-007` | `contracts/index.json`; `internal/gen/contracts/contracts_gen.go`; `packages/protocol-ts/src/generated/contracts.ts` | Artifact `540921da`; §15.25 evidence; generated drift enforcement remains `NFA-GEN-004` | `make generate`; `make json-shape-check`; `make generate-drift`; `make backend-unit`; `make frontend-typecheck` | Outputs contain generated markers and no hand edit |
| `NFA-GEN-004` | Enforce generated and schema drift | generated contracts | `TODO` | `NFA-GEN-003`, `NFA-TH-007` | harness | `NF-BLOCK-007` | Drift scripts and manifests | Passing retained drift artifacts | `make generate-drift` | Clean regeneration is byte-identical |
| `NFA-FIX-001` | Author and freeze `NF-FIX-001-cisco-sna-minimal` | fixtures | `BLOCKED` | owner contracts, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, `NF-AC-006`, `NF-AC-052` | `TODO: fixtures/network-flow/cisco-sna-minimal.csv` unobserved | Bytes, SHA-256, mapping, transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-002` | Author and freeze `NF-FIX-002-cisco-sna-interface-fields` | fixtures | `BLOCKED` | owner contracts, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, `NF-AC-070` | `TODO: fixtures/network-flow/cisco-sna-interface-fields.csv` unobserved | Bytes, SHA-256, mapping, transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-003` | Author and freeze `NF-FIX-003-duplicate-headers` | fixtures | `BLOCKED` | owner contracts, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, `NF-AC-009` | `TODO: fixtures/network-flow/duplicate-headers.csv` unobserved | Bytes, SHA-256, mapping, transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-004` | Author and freeze `NF-FIX-004-rejected-rows` | fixtures | `BLOCKED` | owner contracts, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, `NF-AC-014..016` | `TODO: fixtures/network-flow/rejected-rows.csv` unobserved | Bytes, SHA-256, diagnostics transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-005` | Author and freeze `NF-FIX-005-csv-parser-edges` | fixtures | `BLOCKED` | owner contracts, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, `NF-AC-012` | `TODO: fixtures/network-flow/csv-parser-edges.csv` unobserved | Bytes, SHA-256, parser transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-006` | Author and freeze `NF-FIX-006-cross-table-graph` | fixtures | `BLOCKED` | `NFA-GP-002`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, graph ACs | `TODO: fixtures/network-flow/cross-table-graph/` unobserved | Directory manifest, mappings, transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-007` | Author and freeze `NF-FIX-007-indicator-linking` | fixtures | `BLOCKED` | `NFA-C02-002`, `NFA-C04-004`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, link ACs | `TODO: fixtures/network-flow/indicator-linking.csv` unobserved | Bytes, SHA-256, link/audit transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-008` | Author and freeze `NF-FIX-008-large-limits` | fixtures | `BLOCKED` | `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, `NF-AC-050` | `TODO: fixtures/network-flow/large-limits/` unobserved | Directory manifest and engineering evidence | `TODO: Network Flow target not found` | Evidence remains non-claim-bearing |
| `NFA-FIX-009` | Author and freeze `NF-FIX-009-soft-delete-stale-graph` | fixtures | `BLOCKED` | `NFA-C03-002`, `NFA-C04-002`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, lifecycle ACs | `TODO: fixtures/network-flow/soft-delete-stale-graph.csv` unobserved | Bytes, SHA-256, stale-result transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-010` | Author and freeze `NF-FIX-010-json-admission` | fixtures | `BLOCKED` | `NFA-GEN-002`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, `NF-AC-003..005` | `TODO: fixtures/network-flow/json-admission.jsonl` unobserved | Bytes, SHA-256, error transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-011` | Author and freeze `NF-FIX-011-alias-collision` | fixtures | `BLOCKED` | `NFA-C01-003`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, mapping ACs | `TODO: fixtures/network-flow/alias-collision.csv` unobserved | Bytes, SHA-256, approved mapping | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-012` | Author and freeze `NF-FIX-012-sys-uptime-timestamps` | fixtures | `BLOCKED` | `NFA-TZ-001`, `NFA-TH-003` | Network Flow | `NF-BLOCK-006`, timestamp ACs | `TODO: fixtures/network-flow/sys-uptime-timestamps.csv` unobserved | Bytes, SHA-256, time transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-013` | Author and freeze `NF-FIX-013-filename-display` | fixtures | `BLOCKED` | `NFA-C01-003`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, `NF-AC-080` | `TODO: fixtures/network-flow/filename-display.jsonl` unobserved | Bytes, SHA-256, display transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-014` | Author and freeze `NF-FIX-014-cursor-pagination` | fixtures | `BLOCKED` | `NFA-C04-002`, `NFA-TH-003` | Network Flow | `NF-BLOCK-006`, cursor ACs | `TODO: fixtures/network-flow/cursor-pagination.csv` unobserved | Bytes, SHA-256, cursor transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-015` | Author and freeze `NF-FIX-015-graph-adapter-input` | fixtures | `BLOCKED` | `NFA-GP-002`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, adapter ACs | `TODO: fixtures/network-flow/graph-adapter-input/` unobserved | Directory manifest and exact adapter transcript | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-016` | Author and freeze `NF-FIX-016-redaction` | fixtures | `BLOCKED` | `NFA-C04-003`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, redaction ACs | `TODO: fixtures/network-flow/redaction.csv` unobserved | Bytes, SHA-256, safe-digest transcript | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-017` | Author and freeze `NF-FIX-017-indicator-link-mismatch` | fixtures | `BLOCKED` | `NFA-C02-001`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, link-rejection ACs | `TODO: fixtures/network-flow/indicator-link-mismatch.csv` unobserved | Bytes, SHA-256, error transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-018` | Author and freeze `NF-FIX-018-resource-limits` | fixtures | `BLOCKED` | `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, limit ACs | `TODO: fixtures/network-flow/resource-limits/` unobserved | Directory manifest and limit transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-019` | Author and freeze `NF-FIX-019-canonical-json-unicode` | fixtures | `BLOCKED` | `NFA-GEN-002`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, canonicalization ACs | `TODO: fixtures/network-flow/canonical-json-unicode.jsonl` unobserved | Bytes, SHA-256, canonical-byte transcript | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-020` | Author and freeze `NF-FIX-020-atomic-import-commit` | fixtures | `BLOCKED` | `NFA-C01-004`, `NFA-TH-002` | Network Flow | `NF-BLOCK-006`, `NF-AC-081`, `NF-AC-107` | `TODO: fixtures/network-flow/atomic-import-commit/` unobserved | Directory manifest and failure matrix | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-021` | Author and freeze `NF-FIX-021-preview-boundaries` | fixtures | `BLOCKED` | `NFA-C01-003`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, preview ACs | `TODO: fixtures/network-flow/preview-boundaries/` unobserved | Directory manifest and parser transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-022` | Author and freeze `NF-FIX-022-timestamp-rulesets` | fixtures | `BLOCKED` | `NFA-TZ-001`, `NFA-TH-003` | Network Flow | `NF-BLOCK-006`, `NF-BLOCK-017`, time ACs | `TODO: fixtures/network-flow/timestamp-rulesets/` unobserved | Directory manifest, tzdb provenance, transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-023` | Author and freeze `NF-FIX-023-import-facade-source-change` | fixtures | `BLOCKED` | `NFA-C01-003`, `NFA-TH-002` | Network Flow | `NF-BLOCK-006`, `NF-BLOCK-011` | `TODO: fixtures/network-flow/import-facade-source-change/` unobserved | Directory manifest and facade transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-024` | Author and freeze `NF-FIX-024-query-normalization-cursors` | fixtures | `BLOCKED` | `NFA-C04-002`, `NFA-TH-003` | Network Flow | `NF-BLOCK-006`, query/cursor ACs | `TODO: fixtures/network-flow/query-normalization-cursors.jsonl` unobserved | Bytes, SHA-256, cursor transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-025` | Author and freeze `NF-FIX-025-graph-contributors` | fixtures | `BLOCKED` | `NFA-GP-002`, `NFA-C04-001`, `NFA-TH-005` | Network Flow | `NF-BLOCK-006`, contributor ACs | `TODO: fixtures/network-flow/graph-contributors/` unobserved | Directory manifest and page transcripts | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-026` | Author and freeze `NF-FIX-026-audit-and-replay` | fixtures | `BLOCKED` | `NFA-C04-003`, `NFA-C04-004`, `NFA-TH-006` | Network Flow | `NF-BLOCK-006`, audit ACs | `TODO: fixtures/network-flow/audit-and-replay/` unobserved | Directory manifest and exact count transcript | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-FIX-027` | Author and freeze `NF-FIX-027-retention-soft-delete` | fixtures | `BLOCKED` | `NFA-C04-005`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, lifecycle ACs | `TODO: fixtures/network-flow/retention-soft-delete/` unobserved | Directory manifest and retained-state transcript | `TODO: Network Flow target not found` | Soft delete, incident closure retention, non-queryability, and no purge claim are evidenced |
| `NFA-FIX-028` | Author and freeze `NF-FIX-028-graph-aggregate-bounds` | fixtures | `BLOCKED` | `NFA-GP-002`, `NFA-TH-001` | Network Flow | `NF-BLOCK-006`, aggregate ACs | `TODO: fixtures/network-flow/graph-aggregate-bounds/` unobserved | Directory manifest and no-partial-output transcript | `TODO: Network Flow target not found` | §8 row is fully evidenced |
| `NFA-TRANSCRIPT-001` | Canonical expected-output transcript production | fixtures | `BLOCKED` | `NFA-FIX-001..028`, `NFA-GEN-002` | Network Flow/harness | `NF-BLOCK-006`, `NF-AC-052` | `TODO: transcript path convention not found` | All §22 applicable outputs | `TODO: Network Flow target not found` | Every fixture names immutable expected outputs |
| `NFA-TEST-001` | Parser, canonicalization, digest, and mapping tests | conformance | `TODO` | contracts and fixtures | Network Flow | `NF-BLOCK-008` | `TODO: test selectors not found` | Executed AC evidence | `TODO: Network Flow target not found` | Mapped AC rows have retained results |
| `NFA-TEST-002` | Route, authorization, and hidden-resource tests | conformance | `TODO` | Core 01/Core 04, fixtures | Network Flow | `NF-BLOCK-008` | `TODO: test selectors not found` | Executed route evidence | `TODO: Network Flow target not found` | Mapped AC rows have retained results |
| `NFA-TEST-003` | Graph adapter and contributor tests | conformance | `TODO` | Graph Projection, fixtures | Network Flow | `NF-BLOCK-008`, `NF-BLOCK-015` | `TODO: test selectors not found` | Executed graph evidence | `TODO: Network Flow target not found` | Mapped AC rows have retained results |
| `NFA-TEST-004` | Indicator link and atomicity tests | conformance | `TODO` | Core 02, Core 04, fixtures | Network Flow | `NF-BLOCK-008` | `TODO: test selectors not found` | Executed link/failure evidence | `TODO: Network Flow target not found` | Mapped AC rows have retained results |
| `NFA-TEST-005` | Clock, cursor, rotation, audit, soft-delete, and retention tests | conformance | `TODO` | harness capabilities, fixtures | Network Flow | `NF-BLOCK-008`, `NF-BLOCK-016`, `NF-BLOCK-017` | `TODO: test selectors not found` | Executed security/lifecycle evidence | `TODO: Network Flow target not found` | Mapped AC rows have retained results |
| `NFA-TEST-006` | Browser workspace and invalidation tests | conformance | `TODO` | Core 03, production UI | Network Flow | `NF-BLOCK-008` | `TODO: phase/map/test selectors not found` | Browser row accounting | `TODO: Network Flow target not found` | UI AC rows close through public boundaries |
| `NFA-TEST-007` | Structural references, unknown members, and drift tests | conformance | `TODO` | generated contracts, harness | Network Flow | `NF-BLOCK-007`, `NF-BLOCK-008` | `TODO: Network Flow structural target not found` | Structural and drift summaries | `make generate-drift` | All references resolve and unknown members reject |
| `NFA-DOM-001` | Define Network Flow extension-workspace vocabulary and boundaries | `WS-01` | `DONE` | `NFA-C00-001` | domain vocabulary | `NF-BLOCK-001`, `NF-BLOCK-004` | `docs/domain.md` | Domain amendment `155b5f64`; §15.2 evidence | `make lint-markdown`; domain structural checks | Analytical resources remain distinct from records, views, and import units |
| `NFA-DESIGN-001` | Define Network Analysis design-direction application | `WS-17` | `BLOCKED` | `NFA-C03-001` | frontend design | UI ACs | `docs/design.md` | Narrow token/layout/interaction guidance | `make lint-markdown` | UI uses existing design tokens without claiming product conformance |
| `NFA-MIG-001` | Expand/contract schema and rollout compatibility | `WS-11`–`WS-13` | `BLOCKED` | `NFA-C01-003..005`, `NFA-C04-005` | storage/platform owners | `NF-BLOCK-003`, `NF-BLOCK-011`, `NF-BLOCK-014` | `db/migrations`; deployment guidance | Migration, drain, rollback, and cleanup evidence | `make migration-drift` | Existing completed state remains readable and unsafe nonterminal state fails closed |
| `NFA-PHASE12-001` | Add Phase 12 test maps, ledgers, schedules, and public targets | `WS-09`, `WS-18` | `BLOCKED` | `NFA-TH-007`, `NFA-GEN-004` | Testing Harness | `NF-BLOCK-007`, `NF-BLOCK-008`, `NF-BLOCK-016` | `tools/phase12_test_map.json`; phase registries | Generated topology plus retained phase slices | `make phase-map-check`; `make phase-ledger-drift`; `make phase-schedule-drift` | Phase 12 owns every Network Flow test row without changing earlier phase accounting |
| `NFA-IMPL-010` | Replace closure-only jobs with durable named handlers | `WS-10` | `BLOCKED` | `NFA-GEN-004`, `NFA-C01-005` | platform/jobs and async module owners | `NF-BLOCK-003`, `NF-BLOCK-011`, `NF-BLOCK-016` | `internal/platform/jobs`; async modules | Leases, attempts, recovery, cancellation, drain, and migration tests | `make backend-unit`; `make backend-integration` | Every async family recovers after restart and duplicate delivery |
| `NFA-IMPL-011` | Add import source streams, mapping v2, and transaction coordination | `WS-11` | `BLOCKED` | `NFA-IMPL-010`, `NFA-C01-002..005` | imports/platform | `NF-BLOCK-003`, `NF-BLOCK-011`, `NF-BLOCK-012` | imports, object store, migrations | Opaque source lifecycle and atomic extension-target stub | `make backend-unit`; `make backend-integration`; `make migration-drift` | v1 remains valid and v2 commits or fails without partial publication |
| `NFA-IMPL-012` | Register discovery, configuration, and claim gating | `WS-12` | `BLOCKED` | `NFA-GEN-004`, `NFA-IMPL-011`, `NFA-C04-003` | extensions/config/app assembly | `NF-BLOCK-001`, `NF-BLOCK-002`, `NF-BLOCK-005` | extension/config/runtime packages | Default-unclaimed and fail-closed startup evidence | `make backend-unit`; `make backend-integration` | Unclaimed deployments expose no Network Flow route or workspace |
| `NFA-IMPL-013` | Implement Network Flow storage and table lifecycle | `WS-13` | `BLOCKED` | `NFA-IMPL-011`, `NFA-IMPL-012` | `internal/modules/networkflow` | lifecycle ACs | module, migrations, authored queries | Immutable row, table-limit, rename, soft-delete, and metric evidence | `make backend-unit`; `make backend-store`; `make migration-drift` | Only active tables are queryable and retained counts are observable |
| `NFA-IMPL-014` | Implement streaming preview and atomic apply | `WS-14` | `BLOCKED` | `NFA-IMPL-013`, `NFA-TZ-001` | Network Flow/imports | ingest ACs | Network Flow parser/import facade | Parser, staging, source-hash, bulk-load, fault, and recovery evidence | `make backend-unit`; `make backend-integration`; `make backend-process` | No staging object is public and final publication is atomic |
| `NFA-IMPL-015` | Implement queries, secure cursors, digests, and audit | `WS-15` | `BLOCKED` | `NFA-IMPL-013`, `NFA-IMPL-014`, `NFA-C04-002..005` | Network Flow/security | query/security ACs | Network Flow routes and security primitives | Authorization, pagination, rotation, digest, and audit-count evidence | `make backend-unit`; `make backend-integration`; `make go-gosec-targeted` | Every continuation reauthorizes and no secret/raw value leaks |
| `NFA-IMPL-016` | Implement ephemeral graph and atomic indicator binding | `WS-16` | `BLOCKED` | `NFA-GP-003`, `NFA-C02-002`, `NFA-IMPL-015` | Network Flow/Graph Projection/indicators | graph/link ACs | owning modules and transaction ports | No-retention graph plus concurrent binding/fault evidence | `make backend-unit`; `make backend-integration` | Projection creates no retained run and binding is all-or-nothing |
| `NFA-IMPL-017` | Implement invalidation and Network Analysis workspace | `WS-17` | `BLOCKED` | `NFA-C03-002`, `NFA-IMPL-012..016`, `NFA-DESIGN-001` | WebSocket/frontend | UI ACs | WebSocket contracts and `apps/web` | Functional, stateful, accessibility, and visual evidence | `make frontend-typecheck`; `make frontend-unit`; `make browser-e2e-stateful`; `make browser-e2e-a11y`; `make browser-e2e-visual` | Claimed workspace remains current-authorization safe and Base surfaces do not change |
| `NFA-IMPL-018` | Freeze fixtures and execute Phase 12 conformance | `WS-18` | `BLOCKED` | `NFA-IMPL-014..017`, `NFA-PHASE12-001`, all fixture tasks | Network Flow/Testing Harness | `NF-BLOCK-006..008`, `NF-BLOCK-016..017` | fixture corpus, transcripts, test maps | Exact 28/107 accounting and retained phase runs | `make phase-slice PHASE=phase12`; `make service-backed-slice PHASE=phase12` | Every acceptance row has one attributable passing result |
| `NFA-VAL-001` | Structural lint and tracker matrix validation | validation | `DONE` | tracker creation | tracker maintainer | tracker ACs | This file | Artifact `1bb6fdbd`; §15.1 retained validation | Commands in §15 | All tracker-level criteria pass |
| `NFA-VAL-002` | Full Network Flow conformance run | validation | `BLOCKED` | all tests and fixtures | Testing Harness | all `NF-AC-*` | `TODO: public target not found` | Retained run root and row accounting | `TODO: Network Flow target not found` | Every AC executes against intended artifact |
| `NFA-VAL-003` | Security, fault, and drift evidence bundle | validation | `BLOCKED` | `NFA-VAL-002`, `NFA-GEN-004` | owners/harness | blockers 007, 008, 014, 016, 017 | `TODO: evidence bundle path not found` | Attributable immutable bundle | `TODO: Network Flow target not found` | No unresolved product, harness, or drift failure |
| `NFA-ADOPT-001` | Final owner review, version, locator, and status transition | adoption | `BLOCKED` | every gate, blocker, fixture, generated task, AC row, validation task | all owners | `NF-AC-106` and all adoption IDs | Core 00 and Network Flow status headers/registries | Coordinated adopted/current changes and evidence | structural/status target `TODO:` | Nothing required remains open |
| `NFA-HANDOFF-001` | Session handoff and next-slice bootstrap | handoff | `IN_PROGRESS` | current session | tracker maintainer | tracker ACs | §§14–17 | Current handoff record | `git diff --name-only` | Another agent can resume without discovery |

## 5. Gate and blocker crosswalk

Each owner ID appears once in this crosswalk. Overlap between a gate and a
blocker maps to the same tasks rather than creating duplicate product behavior.

| Owner ID | Requirement summary | Owning document/module | Tracker task IDs | Required repository changes | Required evidence | Current status | Blocking dependency | Closure condition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `NF-GATE-001` | Add Network Flow to extension ownership and adopted map | Core 00 | `NFA-C00-001`, `NFA-C00-002`, `NFA-LOC-001` | Owner amendment; final registry/status coordination | Adopted Core text and immutable locator | `BLOCKED` | Core 00 extension decision | Core lists the adopted bounded profile only at final close |
| `NF-GATE-002` | Discover the claimed route family | Core 01/extensions | `NFA-C01-001`, `NFA-GEN-002` | Discovery owner contract and derived schemas | Claimed/unclaimed discovery fixture | `BLOCKED` | `NFA-C00-001`; discovery shape contradiction | Exact adopted discovery behavior passes |
| `NF-GATE-003` | Admit extension analytical import targets | Core 01/imports | `NFA-C01-002`, `NFA-C01-003` | Import target/result and facade amendments | Import mapping/apply fixtures | `BLOCKED` | Current record/view-only facade | Exact `network_flow_table` target is adopted |
| `NF-GATE-004` | Publish terminal Network Flow resource results | Core 01/jobs/imports | `NFA-C01-002`, `NFA-C01-005` | Result union and publication amendment | Replay/cancel/recovery fixture | `BLOCKED` | Current closed `resource_refs[].kind` | Terminal result references an extension resource safely |
| `NF-GATE-005` | Admit an extension top-level incident tab | Core 03 | `NFA-C03-001`, `NFA-TEST-006` | Extension surface contract and UI implementation | Claimed/unclaimed browser fixture | `BLOCKED` | Owner contract exists in `08fa716e`; UI implementation and browser fixture remain later | Base built-in list stays unchanged |
| `NF-GATE-006` | Add Network Flow route authorization | Core 04 | `NFA-C04-001`, `NFA-TEST-002` | Authorization matrix and route hooks | Membership/role/admin fixtures | `BLOCKED` | Owner matrix exists in `3b942fe0`; generated contracts, route hooks, and fixtures remain later | Route-time reauthorization passes |
| `NF-GATE-007` | Provide one cross-owner unit of work | Core 01 with Core 02/Core 04 participants | `NFA-C01-004`, `NFA-C02-002`, `NFA-C04-004`, `NFA-TH-002` | Transaction capability and participants | Failure at every final-commit step | `BLOCKED` | Owner participants exist in `89580f0c`, `344486e7`, and `71258589`; harness faults and implementation evidence absent | All-or-nothing state is proven |
| `NF-GATE-008` | Designate IP identity and binding transaction interface | Core 02 | `NFA-C02-001`, `NFA-C02-002`, `NFA-C01-004` | Indicator registry/canonicalization and transaction-participant amendments | Canonical IP, concurrent binding, and fault fixtures | `BLOCKED` | Core 02 token and participant artifact exists; Network Flow binding implementation, concurrent binding, and fault fixtures remain later | Adopted family-specific tokens and atomic binding are exact |
| `NF-GATE-009` | Provide extension-resource invalidation | Core 03 with Core 01 wire owner | `NFA-C03-002`, `NFA-TH-005` | Event/wire and UI consequence amendments | Rename/delete/auth-loss UI fixtures | `BLOCKED` | Owner/wire contract exists in `08fa716e`; generated WS contracts, C04 authorization, UI implementation, and harness fixtures remain later | Current authorization invalidates resources deterministically |
| `NF-GATE-010` | Provide cursor, audit, secret, and soft-delete/source-retention hooks | Core 04 with Core 01 cursor wire owner | `NFA-C04-002..005`, `NFA-TH-003`, `NFA-TH-006` | Security lifecycle and conformance amendments | Cursor/rotation/audit/source-expiry/soft-delete fixtures | `BLOCKED` | Cursor, safe-digest, audit, and retention owner contracts exist in `90401fb2`, `cd645750`, `71258589`, and `663c8684`; harness controls, implementation hooks, and fixtures remain absent | Every named security fixture passes |
| `NF-GATE-011` | Adopt ephemeral Graph Projection adapter boundary | Graph Projection | `NFA-GP-001`, `NFA-GP-002`, `NFA-GP-003` | Subsystem amendment and aligned evidence | Exact input/success/failure fixtures | `BLOCKED` | Ephemeral operation owner exists in `4e446354`; adapter mapping/outcome owner text exists in `f177fb6b`; 69/36 evidence drift closed in `81941bba`; Network Flow adapter fixtures remain | Ephemeral invocation and mappings are adopted |
| `NF-GATE-012` | Execute Network Flow contracts, fixtures, lint, and drift | Testing Harness | `NFA-TH-001..007`, `NFA-VAL-002` | Harness amendment, manifests, targets, accounting | Retained full conformance run | `BLOCKED` | Harness manifest, fault, clock, randomness, auth, audit, and structural accounting primitives exist; generated outputs, fixtures, Phase 12 targets, and full retained conformance run remain later | All §22 and §23 rows execute |
| `NF-BLOCK-001` | Core 00 recognizes the profile | Core 00 | `NFA-C00-001`, `NFA-C00-002` | Same owner amendment as gate 001 | Final adopted registry evidence | `BLOCKED` | `NFA-C00-001` | Core registry and NLSpec status agree |
| `NF-BLOCK-002` | Discovery lists the route only as claimed behavior | Core 01 | `NFA-C01-001` | Discovery contract/implementation | Claimed/unclaimed fixture | `BLOCKED` | `NFA-C00-001` | Discovery is exact and generated |
| `NF-BLOCK-003` | Import terminal refs admit Network Flow tables | Core 01 | `NFA-C01-002`, `NFA-C01-005` | Result union amendment | Terminal/replay fixture | `BLOCKED` | Closed Core result vocabulary | Extension ref is owner-adopted |
| `NF-BLOCK-004` | Extension tab without base-tab expansion | Core 03 | `NFA-C03-001` | Extension-surface contract | Browser surface fixture | `BLOCKED` | Owner seam exists in `08fa716e`; browser surface fixture remains later | Claimed-only tab behavior passes |
| `NF-BLOCK-005` | Authorization/conformance hooks exist | Core 04 | `NFA-C04-001` | Route-family matrix and criteria | Route authorization results | `BLOCKED` | Owner matrix exists in `3b942fe0`; implementation and route authorization results remain later | No membership/admin bypass drift |
| `NF-BLOCK-006` | Every fixture has path, hash, and transcript | Network Flow/Testing Harness | `NFA-FIX-001..028`, `NFA-TRANSCRIPT-001` | Author normative bytes and manifests | 28 immutable fixture rows | `BLOCKED` | Owner contracts and manifest schema | No fixture `TODO:` remains |
| `NF-BLOCK-007` | Generated contracts exist and do not drift | Contract owners/harness | `NFA-GEN-001..004`, `NFA-TH-007` | Authored inputs, generator, outputs, drift | Passing retained generation artifacts | `BLOCKED` | Contract-family ownership exists in `d8dbaf3f`; authored Network Flow contracts, generated outputs, drift-accounting target, and retained generation artifacts remain absent | Clean regeneration is byte-identical |
| `NF-BLOCK-008` | All acceptance families have executable tests | Network Flow/Testing Harness | `NFA-TEST-001..007`, `NFA-VAL-002` | Production/test implementation and accounting | 107 retained AC results | `BLOCKED` | No Network Flow tests or target | Every matrix row names a passing artifact |
| `NF-BLOCK-009` | Exact IP indicator type is designated | Core 02 | `NFA-C02-001` | Indicator registry amendment | Canonical IPv4/IPv6 fixture | `DONE` | Artifact `344486e7` designates `ipv4_addr` and `ipv6_addr` with byte-exact canonicalization | Adopted token and algorithm are cited |
| `NF-BLOCK-010` | Every dependency has version and locator | All dependency owners | `NFA-LOC-001` | Fill Table 1-B after owner adoption | Structural reference validation | `DONE` | Artifact `a5dc9847` fills all seven Table 1-B rows and adds structural accounting for unresolved locator cells | Seven exact immutable locators resolve |
| `NF-BLOCK-011` | Import facade and atomic publication are adopted | Core 01 | `NFA-C01-003..005`, `NFA-TH-002` | Owner facade/stream/source-change/publication amendments | Facade, fault, cancel, and replay fixtures | `BLOCKED` | Current per-row record facade | Two-operation boundary passes |
| `NF-BLOCK-012` | IP identity and binding unit-of-work behavior is adopted | Core 02 | `NFA-C02-001..002`, `NFA-C01-004` | Indicator and transaction-participant amendments | IP/link/fault fixtures | `BLOCKED` | Core 02 identity and participant artifact exists; Network Flow binding implementation and link/fault fixtures remain later | Exact owner behavior passes |
| `NF-BLOCK-013` | Invalidation topics and consequences are adopted | Core 03 | `NFA-C03-002`, `NFA-TH-005` | Interaction/wire amendments | Rename/delete/auth-loss fixtures | `BLOCKED` | Owner/wire event exists in `08fa716e`; generated contract, route authorization, UI, and fixture evidence remain later | Every consequence is current-authorization safe |
| `NF-BLOCK-014` | Cursor, digest, audit, and retention behavior is adopted | Core 04/Core 01 | `NFA-C04-002..005`, `NFA-TH-003`, `NFA-TH-006` | Security owner amendments | Rotation/count/source-expiry/soft-delete/cursor fixtures | `BLOCKED` | Cursor, safe-digest, audit, and retention owners resolved in `90401fb2`, `cd645750`, `71258589`, and `663c8684`; harness controls, implementation, and fixtures unresolved | Exact security lifecycle passes without a current purge claim |
| `NF-BLOCK-015` | Graph Projection accepts the exact adapter boundary | Graph Projection | `NFA-GP-001..003`, `NFA-FIX-015`, `NFA-FIX-028` | Subsystem amendment and evidence alignment | Input/outcome/aggregate fixtures | `BLOCKED` | Ephemeral operation, adapter mapping/outcome contracts, and Graph Projection 69/36 evidence alignment exist; Network Flow adapter fixtures remain | Adapter inputs and outcomes pass |
| `NF-BLOCK-016` | Harness can execute required immutable and injected scenarios | Testing Harness | `NFA-TH-001..007` | Harness amendment and implementation | Harness contract results | `DONE` | Artifacts `b3f46bfd`, `ecd4edd5`, `30880992`, `84eb7fb5`, `43b0fe36`, `d96f9ad2`, and `0eaa8db3` cover manifests, faults, clock, randomness, auth transitions, audit assertions, and structural accounting | Manifest/fault/clock/auth/audit capabilities pass |
| `NF-BLOCK-017` | `tzdb-2026c` data is immutable and tested | Network Flow/harness | `NFA-TZ-001`, `NFA-FIX-022` | Provenance artifact and fixture transitions | License, revision, digest, fold/gap results | `BLOCKED` | Provenance artifact `42338a1d` exists and is schema-checked; timestamp transition fixture bytes and transcripts remain absent | One immutable ruleset backs exact expectations |

## 6. Dependency-ordered workflow map

| Workflow | Objective | Prerequisites | Parallelizable peers | Output | Acceptance | Stop condition |
| --- | --- | --- | --- | --- | --- | --- |
| `WF-00` | Session/source bootstrap | Repository access | none | Snapshot, constraints, source list | Branch, commit, dirty state, paths, limits recorded | Stop if instructions or owner sources are missing |
| `WF-01` | Current repository inventory | `WF-00` | none | Authored/generated/code/test/fixture inventory | Existing and absent paths are distinguished | Stop on unexplained dirty overlap |
| `WF-02` | Owner-contract and contradiction map | `WF-01` | none | Table 1-B and owner seam map | Every contradiction has owner and resolution condition | Stop before choosing either side of a conflict |
| `WF-03` | Core 00–04 amendment slices | `WF-02` | none under the controlling-ledger protocol | Owner-approved Core amendments | Each slice closes one named seam and has conformance criteria | Stop if profile ownership or the explicit v1 non-purge boundary is unresolved |
| `WF-04` | Graph Projection and Testing Harness amendments | `WF-02` | Graph Projection and Harness can proceed in parallel | Adopted ephemeral and harness capability contracts | Exact interfaces and evidence mechanics are closed | Stop on mismatch with adopted Core ownership |
| `WF-05` | Generated-contract and generator plan | Relevant `WF-03`, `WF-04` owners adopted | None for each dependent contract family | Authored schemas, generator ownership, output inventory | No generated file hand edit and drift target exists | Stop if public shape still has owner ambiguity |
| `WF-06` | Fixture design and byte freeze | `WF-02`; design may begin before owners close | Fixture families may be designed in parallel | Bytes, mappings, manifest entries, hashes | Freeze only after relevant owner seams close | Stop before hash if any referenced owner is unresolved |
| `WF-07` | Expected-output transcript production | `WF-05`, relevant `WF-06` bytes frozen | Independent fixture families | Canonical expected outputs | Inputs, time, IDs, randomness, ordering are frozen | Stop on owner, generator, or byte drift |
| `WF-08` | Executable AC coverage | `WF-03`, `WF-04`, `WF-05`; relevant fixtures/transcripts | Test layers may proceed by owner family | Tests, row maps, retained artifacts | Every AC executes against intended artifacts | Stop if harness cannot account for a row |
| `WF-09` | Validation, drift, security, and injected failure evidence | `WF-07`, `WF-08` | Drift/security suites may run in parallel | Full retained evidence bundle | Product, harness, infra, and drift results classified | Stop on any unresolved required failure |
| `WF-10` | Adoption close-out and status transition | `WF-09`, `NFA-LOC-001` | none | Owner review, versions, locators, status changes | Every gate/blocker/fixture/AC is closed | Stop if any prerequisite is not evidenced |
| `WF-11` | Handoff and next-slice bootstrap | Every session | none | Updated tracker and one next action | Another agent can resume without discovery | Stop before ending a session without handoff |

### 6.1 Artifact and tracker checkpoint protocol

Execution is strictly serial even when two workstreams are technically
independent. For each workstream:

1. refresh the snapshot and mark exactly one workstream `IN_PROGRESS`;
2. commit the owned artifact without marking the workstream `DONE`;
3. run the narrowest Make-owned validation and retain its run root;
4. update this tracker with the artifact commit, validation result, gate/blocker
   movement, migration notes, and the next safe workstream;
5. commit that tracker checkpoint; and
6. only then mark the next workstream `IN_PROGRESS`.

A failure leaves the row `IN_PROGRESS` or `BLOCKED` with one classified failure
and a safe restart command. A `DROPPED` row records an intentional owner decision,
not omitted work. No two implementation workstreams may be active concurrently.

## 7. Core and subsystem amendment workstreams

### 7.1 Core 00 extension ownership and adoption registry

- **Owner seams:** Core 00 §§4.2, 4.3, 5, and 5.1.
- **Smallest slice:** define how `network_flow_activity` enters the current
  extension model while remaining draft until final adoption and record that v1
  does not invent whole-incident purge before a generic Core profile exists.
- **Interfaces/tokens:** profile registry and adopted-subsystem map only; no route
  or storage schema is chosen here.
- **Consumers:** Core 01 discovery, Core 04 claims, derived extension registry,
  and the final Network Flow status header.
- **Tests/fixtures:** structural registry checks and final claimed/unclaimed
  discovery evidence.
- **Compatibility/security:** must not expand Base Profile or imply current
  purge; final status and registry change must be coordinated.
- **Stop point:** no downstream behavior freeze until the owner contradiction is
  resolved. Roll back by withholding the final adopted registry/status change.
- **Done:** committed owner amendment, immutable locator, and retained structural evidence.

### 7.2 Core 01 discovery, import facade, transaction, and publication

- **Owner seams:** §§3.3.3.1, 3.3.6, 3.3.7, 3.3.9.1, 17.1, and
  17.2; `REQ-01-542..548`; `REQ-01-618..620`.
- **Smallest slices:** discovery item/route family; analytical target/result
  union; two-operation stream facade/source-change check; common unit of work;
  terminal publication and cancellation recovery.
- **Interfaces/tokens:** exact extension discovery schema, import target/result
  discriminators, opaque capability, transaction participant contract, and
  terminal resource reference. The tracker does not choose their final shapes.
- **Generated consumers:** extensions, OpenAPI, errors, generated Go/TypeScript,
  and any owner-adopted result registry.
- **Implementation callers:** extensions, imports, jobs, indicators, audit, and a
  future Network Flow module path still marked `TODO:`.
- **Tests/fixtures:** `NF-FIX-001`, `003`, `011`, `020`, `021`, `023`, and `026`.
- **Compatibility/security:** must not reinterpret a flow table as `record_id`,
  `view_schema_id`, or `view_row_v1`; no path/URL or raw stream leakage.
- **Stop point:** stop at each owner seam before contract generation. Roll back
  derived work to the last adopted owner commit.
- **Done:** all five Core 01 tasks have owner text, generated contracts, implementation, and retained fault/replay evidence.

### 7.3 Core 02 indicator identity and atomic dedupe participation

- **Owner seams:** Core 02 §§10.2 and 18 plus Core 00 §4.3 for the explicit
  future-only incident-removal boundary.
- **Smallest slices:** exact family-specific IP tokens/canonicalization and
  find/create/dedupe transaction participation. A Network Flow-specific purge
  cascade is intentionally dropped from v1.
- **Interfaces/tokens:** canonical IP-literal type and comparison behavior,
  and indicator transaction participant. Existing `ipv4_addr` remains valuable
  compatibility behavior; `ipv6_addr` is additive.
- **Generated consumers:** indicator registry/contracts if adopted; no current
  generated Network Flow consumer exists.
- **Implementation callers:** indicators and the import/binding unit of work.
- **Tests/fixtures:** `NF-FIX-007`, `016`, `017`, `020`, and `026`.
- **Compatibility/security:** keep flow resources outside record envelopes and
  avoid creating observations in v1; preserve atomicity and retained audit.
- **Stop point:** stop if IP identity or transaction ownership remains unspecified.
- **Done:** owner contract plus exact canonicalization and transaction evidence;
  the future generic cascade obligation remains named but is not a v1 gate.

### 7.4 Core 03 extension workspace and invalidation

- **Owner seams:** Core 03 §§2 and 4.3.1; Core 01 §3.3.10.1 if a wire event is used.
- **Smallest slices:** claimed-only top-level extension workspace; generic
  current-authorization invalidation with rename/delete/auth-loss consequences.
- **Interfaces/tokens:** extension surface identity and owner-approved event/topic;
  do not infer identity from the `Network Analysis` label.
- **Generated consumers:** WebSocket contract only if the wire owner adopts an event.
- **Implementation callers:** workbook shell, collaboration transport, Network Flow UI.
- **Tests/fixtures:** `NF-FIX-009`, `014`, `024`, `025`, and `027`, plus browser scenarios.
- **Compatibility/security:** preserve the five Base Profile built-in tabs and
  clear hidden/unauthorized resources without disclosure.
- **Stop point:** stop before UI implementation if surface/event identity is not owner-defined.
- **Done:** owner contract, generated wire if applicable, and browser invalidation evidence.

### 7.5 Core 04 authorization, cursors, secrets, audit, and retention

- **Owner seams:** Core 04 §§2, 3, 9, 12; Core 01 §3.3.7 for cursor wire behavior.
- **Smallest slices:** route authorization; cursor security/key lifecycle;
  safe-digest key IDs; audit occurrences/transactionality; retention hooks.
- **Interfaces/tokens:** no final token or namespace is chosen by this tracker.
  Core 01/Core 04 must explicitly divide cursor wire and security ownership.
- **Generated consumers:** OpenAPI/errors/config contracts only after owner adoption.
- **Implementation callers:** route middleware, cursor codec, config/secret loader,
  immutable incident-audit ledger and source/soft-delete retention jobs.
- **Tests/fixtures:** `NF-FIX-009`, `014`, `016`, `020`, `024`, `026`, and `027`.
- **Compatibility/security:** no `deployment_admin` incident bypass; no raw secret,
  digest key, cursor content, incident data, or raw fixture values in diagnostics.
- **Stop point:** stop any fixture-secret or cursor freeze until rotation and
  retention behavior is adopted.
- **Done:** five owner seams close with route, rotation, audit-count, source-expiry,
  and soft-delete evidence without a current incident-purge claim.

### 7.6 Graph Projection ephemeral adapter

- **Owner seams:** Graph Projection §§4–5, 10, 13–14 and Network Flow §14.4.
- **Smallest slices:** non-retained invocation/result; exact compatible property
  and metadata mapping; closed outcome mapping; evidence-matrix repair.
- **Interfaces/tokens:** the adopted contract now has `project_ephemeral` and
  `ephemeral_projection_id`; exact Network Flow metadata/property mapping and any
  direct-copy-compatible wording remain separate.
- **Generated consumers:** Graph Projection conformance matrix/corpus and any
  future owner-adopted schema; current matrix/corpus are authored evidence, not
  generated roots.
- **Implementation callers:** Graph Projection facade and future Network Flow graph adapter only.
- **Tests/fixtures:** `NF-FIX-006`, `015`, `025`, and `028` plus owner GP fixtures.
- **Compatibility/security:** no retained view/run, internal provider leakage, or partial graph output.
- **Stop point:** stop before adapter implementation if input or outcome shape is unresolved.
- **Done:** adopted ephemeral contract, aligned 69-plus acceptance evidence, and exact adapter results.

### 7.7 Testing Harness execution and evidence accounting

- **Owner seams:** Harness §§4, 8, 11, 12, 16, and 17.
- **Smallest slices:** manifest schema/runner; commit and worker faults; Network
  Flow clock use; deterministic randomness; authorization transitions; audit
  counts; generated/structural/drift row accounting.
- **Interfaces/tokens:** all test controls remain test-only, harness-owned,
  token-protected, resettable, and unavailable in production.
- **Generated consumers:** harness schemas/manifests and generated ledgers only
  after their authored owners are adopted.
- **Implementation callers:** harness routes/helpers and future Network Flow tests;
  no production module may import harness internals.
- **Tests/fixtures:** every `NF-FIX-*` and `NF-AC-*` row.
- **Compatibility/security:** fixture-only secret/randomness controls must never
  weaken production configuration or expose secrets in retained artifacts.
- **Stop point:** stop before inventing a manifest format or public Make target.
- **Done:** adopted harness amendment, public target, complete row accounting, and retained full run.

### 7.8 Immutable dependency locators and timezone ruleset

- **Owner seams:** Network Flow Table 1-B, §9.7, §22, §24; all dependency owners.
- **Smallest slices:** fill seven adopted version/section locators only after
  adoption; separately select and freeze one `tzdb-2026c` source with provenance,
  revision, license, and SHA-256.
- **Interfaces/tokens:** `timezone_ruleset_id='tzdb-2026c'` is proposed by the
  draft; this tracker does not choose a vendor/source path or manifest format.
- **Consumers:** mapping fingerprint, `NF-FIX-022`, timestamp tests, structural lint.
- **Tests/fixtures:** `NF-FIX-012`, `019`, `022`; timestamp AC rows.
- **Compatibility/security:** no host timezone/locale fallback and no unlicensed
  or mutable download as conformance input.
- **Stop point:** stop byte freeze if any locator or tzdb provenance field is missing.
- **Done:** all seven locators resolve and timestamp transitions reproduce immutable expectations.

## 8. Normative fixture-byte control ledger

This ledger transcribes Table 22-A into independently closable work. It does not
create fixture authority. The 16 non-directory proposals are single-file SHA-256
inputs. The 12 directory proposals (`006`, `008`, `015`, `018`, `020`, `021`,
`022`, `023`, `025`, `026`, `027`, and `028`) are manifest-backed bundles whose
manifest schema and canonical entry ordering remain harness-owner decisions.

Every row inherits these unresolved byte-affecting inputs: exact UTF-8 and BOM
treatment; chosen CRLF, LF, and CR bytes; file and manifest-entry ordering; fixed
timestamps and identifiers; deterministic randomness injection; a fixture-only
safe-digest key and key ID; fake time; locale; Unicode 17.0.0 inputs; and the
adopted timezone ruleset. A row cannot freeze until the applicable values are
written into harness-owned evidence. Mapping JSON and transcripts are separate
immutable artifacts even when embedded in a future bundle. No proposed path below
was observed in the repository.

| Fixture ID | Required behavior/output | Proposed or existing path | Authored source or generator | Byte-freeze status | SHA-256 or manifest status | Approved mapping JSON | Expected transcript paths | Covered REQ/AC IDs | Dependency tasks | Validation command | Owner | Status | Notes/blocker |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `NF-FIX-001-cisco-sna-minimal` | One table; fingerprint; row IDs; zero diagnostics; graph result. | `TODO: source not found; NLSpec proposes fixtures/network-flow/cisco-sna-minimal.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-006`, `017`, `052` | `NFA-C01-004`; `NFA-TH-001`; `NFA-GEN-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; lock minimal accepted rows, IDs, time, and randomness. |
| `NF-FIX-002-cisco-sna-interface-fields` | Interface mappings; row IDs; graph result. | `TODO: source not found; NLSpec proposes fixtures/network-flow/cisco-sna-interface-fields.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-070`, `085` | `NFA-C01-003`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; interface text and code-point order must be exact. |
| `NF-FIX-003-duplicate-headers` | Ordinal-disambiguated source descriptors. | `TODO: source not found; NLSpec proposes fixtures/network-flow/duplicate-headers.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-009`, `089` | `NFA-C01-003`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; duplicate header bytes and contiguous ordinals are byte-critical. |
| `NF-FIX-004-rejected-rows` | Ordered IP, port, protocol, timestamp, counter, field-count, and end-before-start diagnostics. | `TODO: source not found; NLSpec proposes fixtures/network-flow/rejected-rows.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-014..016`, `063`, `091` | `NFA-C02-001`; `NFA-TZ-001`; `NFA-TH-004` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; diagnostic order, reason keys, and truncation controls must be fixed. |
| `NF-FIX-005-csv-parser-edges` | Terminal newline, blank line, quoted newline, quote escaping, and malformed-quote outcomes. | `TODO: source not found; NLSpec proposes fixtures/network-flow/csv-parser-edges.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-012`, `013`, `083`, `084` | `NFA-TH-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; selected CRLF/LF/CR and quote bytes are the behavior under test. |
| `NF-FIX-006-cross-table-graph` | Two tables; shared vertex; aggregate edge; query digest; snapshot ID; edge IDs. | `TODO: source not found; NLSpec proposes fixtures/network-flow/cross-table-graph/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-032..040`, `056`, `059` | `NFA-GP-001..003`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Graph Projection + Harness | `BLOCKED` | Common inputs; bundle order, table IDs, contributor order, and graph limits must be fixed. |
| `NF-FIX-007-indicator-linking` | Existing/create links and duplicate binding result. | `TODO: source not found; NLSpec proposes fixtures/network-flow/indicator-linking.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-041..043`, `066`, `067`, `101` | `NFA-C02-001..002`; `NFA-C01-004` | `TODO: source not found` | Network Flow + Core 02 + Harness | `BLOCKED` | Common inputs; indicator IDs, canonical IP token, transaction, and duplicate result remain owner-blocked. |
| `NF-FIX-008-large-limits` | Graph/table limit failures and timing measurements. | `TODO: source not found; NLSpec proposes fixtures/network-flow/large-limits/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-038`, `048`, `050`, `076` | `NFA-TH-001`; `NFA-TH-007`; `NFA-C05-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; engineering-only unless Core 05 is separately activated for publication claims. |
| `NF-FIX-009-soft-delete-stale-graph` | Soft-delete invalidates active graph and cursors. | `TODO: source not found; NLSpec proposes fixtures/network-flow/soft-delete-stale-graph.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-023`, `028`, `104` | `NFA-C03-002`; `NFA-C04-002`; `NFA-C02-003` | `TODO: source not found` | Network Flow + Core owners + Harness | `BLOCKED` | Common inputs; fake time and delete/auth transition sequence must be exact. |
| `NF-FIX-010-json-admission` | Duplicate member, invalid null, unknown member, malformed JSON, and non-object failures. | `TODO: source not found; NLSpec proposes fixtures/network-flow/json-admission.jsonl` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-003..005`, `057`, `071` | `NFA-GEN-001`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; raw JSON member order, duplicate bytes, UTF-8 failures, and line endings must freeze. |
| `NF-FIX-011-alias-collision` | Alias keys; duplicate warning; approved mapping; source-reuse conflict. | `TODO: source not found; NLSpec proposes fixtures/network-flow/alias-collision.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-073`, `074`, `089` | `NFA-C01-003`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; alias normalization, warning order, and approved mapping serialization must freeze together. |
| `NF-FIX-012-sys-uptime-timestamps` | Export time and uptime-derived timestamps; wrap-ambiguous rejection. | `TODO: source not found; NLSpec proposes fixtures/network-flow/sys-uptime-timestamps.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-018`, `060`, `087` | `NFA-TZ-001`; `NFA-TH-003` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; exact fake time, unsigned values, event ordinals, and wrap boundary are required. |
| `NF-FIX-013-filename-display` | Path stripping, hidden/trailing-dot stems, override failure, suffixing, and reuse. | `TODO: source not found; NLSpec proposes fixtures/network-flow/filename-display.jsonl` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-007`, `061`, `080`, `082` | `NFA-C01-003`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; Unicode filenames, separators, collision sequence, and commit ordering must be fixed. |
| `NF-FIX-014-cursor-pagination` | Sort tail; keyset continuation; null terminal cursor; actor mismatch; rename survival. | `TODO: source not found; NLSpec proposes fixtures/network-flow/cursor-pagination.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-027`, `028`, `061`, `094`, `095` | `NFA-C04-002`; `NFA-C01-005`; `NFA-TH-003` | `TODO: source not found` | Network Flow + Core 01/Core 04 + Harness | `BLOCKED` | Common inputs; cursor implementation, fake time, actor, and exact sort tuples remain absent after owner split. |
| `NF-FIX-015-graph-adapter-input` | Exact adapter input, view key, snapshot, property definitions, and safe metadata. | `TODO: source not found; NLSpec proposes fixtures/network-flow/graph-adapter-input/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-040`, `069`, `097` | `NFA-GP-001..003` | `TODO: source not found` | Network Flow + Graph Projection + Harness | `BLOCKED` | Common inputs; adapter and metadata schema contradiction must close before transcript freeze. |
| `NF-FIX-016-redaction` | Safe samples, numeric samples, raw SHA-256, safe digests, and no leakage. | `TODO: source not found; NLSpec proposes fixtures/network-flow/redaction.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-047`, `077`, `102` | `NFA-C04-003`; `NFA-TH-004` | `TODO: source not found` | Network Flow + Core 04 + Harness | `BLOCKED` | Core 04 fixture-vs-production key boundary exists; fixture bytes, manifest, and harness deterministic secret control remain absent. |
| `NF-FIX-017-indicator-link-mismatch` | Non-IP/ambiguous selectors, confirmation mismatch, indicator mismatch, and Core failure. | `TODO: source not found; NLSpec proposes fixtures/network-flow/indicator-link-mismatch.csv` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-044`, `068`, `079`, `100` | `NFA-C02-001..002`; `NFA-C01-004` | `TODO: source not found` | Network Flow + Core 02 + Harness | `BLOCKED` | Common inputs; canonical IP identity and closed selector/target owners must settle first. |
| `NF-FIX-018-resource-limits` | Parser/import/graph limits, diagnostic truncation, counter limit, invalid config. | `TODO: source not found; NLSpec proposes fixtures/network-flow/resource-limits/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-038`, `048`, `064`, `076`, `078` | `NFA-TH-001`; `NFA-GP-001`; `NFA-GEN-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; configuration values, limit order, truncation, and bundle entry order must freeze. |
| `NF-FIX-019-canonical-json-unicode` | Exact escapes, scalar ordering, Unicode whitespace/NFC, digest framing, surrogate rejection. | `TODO: source not found; NLSpec proposes fixtures/network-flow/canonical-json-unicode.jsonl` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-017`, `086`, `093` | `NFA-LOC-001`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; Unicode 17.0.0 data, NFC, null framing, raw scalars, and UTF-8 bytes must be immutable. |
| `NF-FIX-020-atomic-import-commit` | Fault at each final commit step leaves no partial table, row, diagnostic, binding, or audit. | `TODO: source not found; NLSpec proposes fixtures/network-flow/atomic-import-commit/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-008`, `072`, `081`, `107` | `NFA-C01-004..005`; `NFA-TH-002` | `TODO: source not found` | Network Flow + Core 01 + Harness | `BLOCKED` | Common inputs; named commit boundaries, worker sequence, fixed IDs, and recovery schedule require harness ownership. |
| `NF-FIX-021-preview-boundaries` | Header semantics, controls, exact 50 records, later error, blank/mismatch count, and `limit+1`. | `TODO: source not found; NLSpec proposes fixtures/network-flow/preview-boundaries/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-012`, `065`, `083`, `084` | `NFA-C01-003`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; record boundaries, CRLF/LF/CR choices, ordering, and exact post-boundary corruption must freeze. |
| `NF-FIX-022-timestamp-rulesets` | Closed timestamp variants, RFC3339, fold/gap, epochs, uptime bounds, ordinals. | `TODO: source not found; NLSpec proposes fixtures/network-flow/timestamp-rulesets/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-018`, `060`, `087` | `NFA-TZ-001`; `NFA-LOC-001`; `NFA-TH-003` | `TODO: source not found` | Network Flow + Harness | `BLOCKED` | Common inputs; `tzdb-2026c` provenance artifact exists in `42338a1d`, but zones, transition fixture bytes, fake time use, locale controls, and transcripts remain unresolved. |
| `NF-FIX-023-import-facade-source-change` | Exact preview/apply envelopes, server descriptors, no path leak, and every source-change reason. | `TODO: source not found; NLSpec proposes fixtures/network-flow/import-facade-source-change/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-088`, `089` | `NFA-C01-003..005`; `NFA-TH-002` | `TODO: source not found` | Network Flow + Core 01 + Harness | `BLOCKED` | Common inputs; opaque-source revision/change sequence and closed result union require Core 01 adoption. |
| `NF-FIX-024-query-normalization-cursors` | Closed scopes, normalized duplicates, request variants, byte bound, expiry, independent tuples. | `TODO: source not found; NLSpec proposes fixtures/network-flow/query-normalization-cursors.jsonl` | `TODO: source not found` | `TODO: bytes not frozen` | `TODO: single-file SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-024..031`, `092..095` | `NFA-C04-002`; `NFA-C01-005`; `NFA-TH-003`; `NFA-TH-005` | `TODO: source not found` | Network Flow + Core 01/Core 04 + Harness | `BLOCKED` | Common inputs; canonical JSON, cursor implementation, fake time, actor/auth transitions, and query spellings must freeze. |
| `NF-FIX-025-graph-contributors` | Exact graph response and contributor pages; current auth; stale digest; no rejected rows. | `TODO: source not found; NLSpec proposes fixtures/network-flow/graph-contributors/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-036..039`, `095`, `098`, `099` | `NFA-GP-001..003`; `NFA-C03-002`; `NFA-TH-005` | `TODO: source not found` | Network Flow + Graph Projection + Harness | `BLOCKED` | Common inputs; contributor sort, page boundaries, authorization schedule, and graph identifiers must freeze. |
| `NF-FIX-026-audit-and-replay` | Binding create/reuse; key IDs; replay silence; graph-success audit; truncation count. | `TODO: source not found; NLSpec proposes fixtures/network-flow/audit-and-replay/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-072`, `101..103` | `NFA-C04-003..004`; `NFA-C01-005`; `NFA-TH-006` | `TODO: source not found` | Network Flow + Core 04 + Harness | `BLOCKED` | Core 04 secret/key-ID and audit-count owners exist; fixture bytes, harness audit-count helper, implementation, replay transcript, and fake time remain absent. |
| `NF-FIX-027-retention-soft-delete` | Soft-delete, incident-closure retention, non-queryability, retained counts, and explicit absence of a v1 purge claim. | `TODO: source not found; NLSpec will own fixtures/network-flow/retention-soft-delete/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-023`, `062`, `104` | `NFA-C04-005`; `NFA-TH-006` | `TODO: source not found` | Network Flow + Core 04 + Harness | `BLOCKED` | Core 04 owner rules exist; fixture bytes, manifest schema, harness retention/audit controls, and NLSpec fixture-name cleanup remain absent. |
| `NF-FIX-028-graph-aggregate-bounds` | Arbitrary-precision sums, digit bound, fixed failure order, and no partial output. | `TODO: source not found; NLSpec proposes fixtures/network-flow/graph-aggregate-bounds/` | `TODO: source not found` | `TODO: bundle bytes not frozen` | `TODO: manifest schema, ordering, and SHA-256` | `TODO: source not found` | `TODO: source not found` | `NF-REQ-177..178`; `NF-AC-038`, `096` | `NFA-GP-001..003`; `NFA-TH-001` | `TODO: source not found` | Network Flow + Graph Projection + Harness | `BLOCKED` | Common inputs; decimal digit strings, aggregate order, limits, IDs, and adapter failure transcript must freeze. |

Fixture freeze means all columns in the applicable row are concrete and the
future harness validates both bytes and expected outputs. `NF-FIX-008` remains
engineering-only evidence unless Core 05 is explicitly activated; its presence
does not create a publication claim.

## 9. Embedded executable acceptance-coverage matrix

No generated Network Flow conformance matrix exists. This authored tracker row
set is therefore the only current planning map, not conformance evidence and not
a generated artifact. Each future test selector must be replaced by an observed
repository selector. “Structural” means document or contract structure, not a
waiver from executable validation.

| AC ID | Behavior | Test level | Test path/node | Fixture IDs | Dependency tasks | Expected artifact | Last command/result | Status | Gap |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `NF-AC-001` | Unclaimed tab and routes unavailable. | browser + route integration | `TODO: source not found` | n/a: claim-state scenario | `NFA-C00-001`; `NFA-C01-001`; `NFA-C03-001`; `NFA-C04-001` | claim-state route/UI transcript | not run; target absent | `TODO` | No adopted discovery, route, UI, implementation, or selector. |
| `NF-AC-002` | Claimed empty state and authorized import action. | browser | `TODO: source not found` | n/a: claim/role scenario | `NFA-C00-001`; `NFA-C03-001`; `NFA-C04-001` | browser trace and assertions | not run; target absent | `TODO` | No extension surface or selector. |
| `NF-AC-003` | Core envelopes and unknown-member rejection. | contract + route integration | `TODO: source not found` | `NF-FIX-010-json-admission` | `NFA-C01-004`; `NFA-C04-001`; `NFA-GEN-001` | request/response transcript | not run; target absent | `TODO` | No route contract or implementation. |
| `NF-AC-004` | Duplicate JSON members rejected before mutation/idempotency. | parser + integration | `TODO: source not found` | `NF-FIX-010-json-admission` | `NFA-C01-005`; `NFA-TH-002`; `NFA-GEN-001` | rejection and zero-state transcript | not run; target absent | `TODO` | No admission parser or fault/state evidence. |
| `NF-AC-005` | UTF-8, JSON, object, null, and member admission is deterministic. | parser + route integration | `TODO: source not found` | `NF-FIX-010-json-admission` | `NFA-GEN-001`; `NFA-TH-001` | ordered admission-error transcript | not run; target absent | `TODO` | Fixture and selector absent. |
| `NF-AC-006` | Minimal import creates one table and inner tab. | integration + browser | `TODO: source not found` | `NF-FIX-001-cisco-sna-minimal` | `NFA-C01-003..005`; `NFA-C03-001`; `NFA-TH-001` | durable table and UI transcript | not run; target absent | `TODO` | Import facade, table implementation, fixture, and selector absent. |
| `NF-AC-007` | Filename-derived display naming and suffixing. | unit + integration | `TODO: source not found` | `NF-FIX-013-filename-display` | `NFA-C01-003`; `NFA-TH-001` | naming vector transcript | not run; target absent | `TODO` | No implementation or fixture bytes. |
| `NF-AC-008` | Apply replay returns same table without duplicate. | transactional integration | `TODO: source not found` | `NF-FIX-020-atomic-import-commit` | `NFA-C01-004..005`; `NFA-TH-002` | replay/state-count transcript | not run; target absent | `TODO` | Core result/publication and commit controls absent. |
| `NF-AC-009` | Duplicate headers use source ordinals. | parser + integration | `TODO: source not found` | `NF-FIX-003-duplicate-headers` | `NFA-C01-003`; `NFA-TH-001` | descriptor/mapping transcript | not run; target absent | `TODO` | Descriptor contract and bytes absent. |
| `NF-AC-010` | Cisco SNA required fields enforced. | parser conformance | `TODO: source not found` | `NF-FIX-001-cisco-sna-minimal`; `NF-FIX-004-rejected-rows` | `NFA-TH-001` | accepted/rejected field transcript | not run; target absent | `TODO` | Parser and fixtures absent. |
| `NF-AC-011` | Reserved profiles cannot be claimed. | contract + integration | `TODO: source not found` | n/a: source-profile vectors | `NFA-C00-001`; `NFA-GEN-001` | closed-enum error transcript | not run; target absent | `TODO` | Owner transition and contract absent. |
| `NF-AC-012` | CSV edge cases follow exact parser outcomes. | parser conformance | `TODO: source not found` | `NF-FIX-005-csv-parser-edges`; `NF-FIX-021-preview-boundaries` | `NFA-TH-001` | byte-to-outcome corpus report | not run; target absent | `TODO` | Bytes, parser, and selector absent. |
| `NF-AC-013` | Formula-looking values remain inert. | parser + security | `TODO: source not found` | `NF-FIX-005-csv-parser-edges` | `NFA-TH-001` | inert-value transcript | not run; target absent | `TODO` | Fixture bytes and implementation absent. |
| `NF-AC-014` | Invalid rows yield deterministic diagnostics and no query/graph rows. | integration | `TODO: source not found` | `NF-FIX-004-rejected-rows` | `NFA-GP-001`; `NFA-TH-004` | diagnostics/query/graph transcript | not run; target absent | `TODO` | Diagnostic ordering control and implementation absent. |
| `NF-AC-015` | Partly valid import retains accepted rows and rejected count. | integration | `TODO: source not found` | `NF-FIX-004-rejected-rows` | `NFA-C01-004`; `NFA-TH-001` | table/count transcript | not run; target absent | `TODO` | Atomic import and fixture absent. |
| `NF-AC-016` | All-invalid import creates no table. | transactional integration | `TODO: source not found` | `NF-FIX-004-rejected-rows` | `NFA-C01-004`; `NFA-TH-002` | error and zero-table transcript | not run; target absent | `TODO` | Commit boundary and executable evidence absent. |
| `NF-AC-017` | Five digest/ID algorithms match exact vectors. | unit + fixture conformance | `TODO: source not found` | `NF-FIX-001-cisco-sna-minimal`; `NF-FIX-006-cross-table-graph`; `NF-FIX-019-canonical-json-unicode` | `NFA-LOC-001`; `NFA-TH-004` | digest vector report | not run; target absent | `TODO` | Immutable inputs and implementation absent. |
| `NF-AC-018` | Timestamp folds, gaps, leap seconds, inference, and reversed intervals reject. | unit + fixture conformance | `TODO: source not found` | `NF-FIX-004-rejected-rows`; `NF-FIX-012-sys-uptime-timestamps`; `NF-FIX-022-timestamp-rulesets` | `NFA-TZ-001`; `NFA-TH-003` | timestamp vector report | not run; target absent | `TODO` | Immutable `tzdb-2026c` source exists; clocked tests and fixture bytes remain absent. |
| `NF-AC-019` | Canonical IP output and invalid forms. | unit conformance | `TODO: source not found` | `NF-FIX-004-rejected-rows`; `NF-FIX-017-indicator-link-mismatch` | `NFA-C02-001` | IP canonicalization vectors | not run; target absent | `TODO` | Core canonical IP token absent. |
| `NF-AC-020` | Unsigned-decimal grammar and max bound. | unit conformance | `TODO: source not found` | `NF-FIX-004-rejected-rows`; `NF-FIX-028-graph-aggregate-bounds` | `NFA-TH-001` | decimal parser vectors | not run; target absent | `TODO` | Implementation and fixture bytes absent. |
| `NF-AC-021` | Table lifecycle excludes renamed state. | structural + state-machine unit | `TODO: source not found` | n/a: structural | `NFA-GEN-001`; `NFA-VAL-002` | schema/state-machine assertion | not run; target absent | `TODO` | No schema or state machine exists. |
| `NF-AC-022` | Rename changes metadata/version only and preserves identities/cursors/bindings. | transactional integration | `TODO: source not found` | `NF-FIX-013-filename-display`; `NF-FIX-014-cursor-pagination` | `NFA-C01-004`; `NFA-C04-002`; `NFA-TH-002` | before/after invariant transcript | not run; target absent | `TODO` | Cursor owner split exists; table/cursor implementation and fixtures absent. |
| `NF-AC-023` | Soft delete hides table and invalidates graph/cursors terminally. | integration + browser | `TODO: source not found` | `NF-FIX-009-soft-delete-stale-graph`; `NF-FIX-027-retention-soft-delete` | `NFA-C03-002`; `NFA-C04-002`; `NFA-C04-005` | invalidation/state transcript | not run; target absent | `TODO` | Invalidation, cursor, and soft-delete owners exist; implementation, browser state, and fixtures absent. |
| `NF-AC-024` | Filters use field keys and reject visible labels. | contract + integration | `TODO: source not found` | `NF-FIX-024-query-normalization-cursors` | `NFA-GEN-001` | query error/success transcript | not run; target absent | `TODO` | Query contract and selector absent. |
| `NF-AC-025` | Filter `in` rejects duplicate/empty values. | unit + route integration | `TODO: source not found` | `NF-FIX-024-query-normalization-cursors` | `NFA-GEN-001` | normalization/error vectors | not run; target absent | `TODO` | Query implementation absent. |
| `NF-AC-026` | CIDR rejects family mismatch and mapped-family coercion. | unit conformance | `TODO: source not found` | `NF-FIX-024-query-normalization-cursors` | `NFA-C02-001` | IP/CIDR vectors | not run; target absent | `TODO` | Canonical IP owner and implementation absent. |
| `NF-AC-027` | Sort semantics and default tail are exact. | integration | `TODO: source not found` | `NF-FIX-014-cursor-pagination`; `NF-FIX-024-query-normalization-cursors` | `NFA-C04-002`; `NFA-TH-001` | ordered page transcript | not run; target absent | `TODO` | Cursor/query implementation and bytes absent. |
| `NF-AC-028` | Cursor invalidation covers TTL, actor, route, query, delete, and auth loss. | security + integration | `TODO: source not found` | `NF-FIX-009-soft-delete-stale-graph`; `NF-FIX-014-cursor-pagination`; `NF-FIX-024-query-normalization-cursors` | `NFA-C04-002`; `NFA-C03-002`; `NFA-TH-003`; `NFA-TH-005` | invalidation matrix | not run; target absent | `TODO` | Cursor security owner contract exists; harness transitions and executable fixtures absent. |
| `NF-AC-029` | Graph filter selector exposes table-selection controls. | browser | `TODO: source not found` | `NF-FIX-006-cross-table-graph` | `NFA-C03-001`; `NFA-GP-001` | browser assertion trace | not run; target absent | `TODO` | UI and graph interface absent. |
| `NF-AC-030` | Table-tab graph defaults to active table scope. | browser + integration | `TODO: source not found` | `NF-FIX-006-cross-table-graph` | `NFA-C03-001`; `NFA-GP-001` | request/UI transcript | not run; target absent | `TODO` | UI and graph implementation absent. |
| `NF-AC-031` | Selected tables reject duplicate IDs. | route integration | `TODO: source not found` | `NF-FIX-024-query-normalization-cursors` | `NFA-GEN-001`; `NFA-GP-001` | admission-error transcript | not run; target absent | `TODO` | Contract and implementation absent. |
| `NF-AC-032` | Multiple selected tables compose one filtered graph. | integration | `TODO: source not found` | `NF-FIX-006-cross-table-graph` | `NFA-GP-001..003` | graph result transcript | not run; target absent | `TODO` | Ephemeral projection contract and implementation absent. |
| `NF-AC-033` | Canonical endpoint merges across selected tables with provenance. | integration | `TODO: source not found` | `NF-FIX-006-cross-table-graph` | `NFA-C02-001`; `NFA-GP-001..003` | vertex/provenance transcript | not run; target absent | `TODO` | IP token and graph adapter unresolved. |
| `NF-AC-034` | Default edge aggregation and table provenance are exact. | integration | `TODO: source not found` | `NF-FIX-006-cross-table-graph`; `NF-FIX-028-graph-aggregate-bounds` | `NFA-GP-001..003` | edge/provenance transcript | not run; target absent | `TODO` | Adapter interface and implementation absent. |
| `NF-AC-035` | Time overlap and zero-duration handling are exact. | unit + integration | `TODO: source not found` | `NF-FIX-006-cross-table-graph`; `NF-FIX-022-timestamp-rulesets` | `NFA-TZ-001`; `NFA-GP-001` | interval vectors and graph transcript | not run; target absent | `TODO` | Ruleset and graph implementation absent. |
| `NF-AC-036` | Vertex selection pivots by stable identity. | browser + integration | `TODO: source not found` | `NF-FIX-025-graph-contributors` | `NFA-GP-001..003`; `NFA-C03-001` | selection/request transcript | not run; target absent | `TODO` | Graph/UI implementation absent. |
| `NF-AC-037` | Edge selection opens deterministically grouped contributors. | browser + integration | `TODO: source not found` | `NF-FIX-025-graph-contributors` | `NFA-GP-001..003`; `NFA-C03-001` | drawer/order transcript | not run; target absent | `TODO` | Graph/UI implementation absent. |
| `NF-AC-038` | Graph over-limit errors are deterministic with no partial graph. | integration + fault | `TODO: source not found` | `NF-FIX-008-large-limits`; `NF-FIX-018-resource-limits`; `NF-FIX-028-graph-aggregate-bounds` | `NFA-GP-003`; `NFA-TH-002` | error and zero-output transcript | not run; target absent | `TODO` | Adapter error contract and fault controls absent. |
| `NF-AC-039` | Example refs and truncation fields are exact. | integration | `TODO: source not found` | `NF-FIX-006-cross-table-graph`; `NF-FIX-025-graph-contributors` | `NFA-GP-002`; `NFA-TH-001` | graph resource transcript | not run; target absent | `TODO` | Property/result contract and fixtures absent. |
| `NF-AC-040` | Adapter input fields and safe metadata are exact. | contract + integration | `TODO: source not found` | `NF-FIX-015-graph-adapter-input` | `NFA-GP-001..003` | adapter-input transcript | not run; target absent | `TODO` | Owner contract exists; generated schema, implementation, and fixture transcript absent. |
| `NF-AC-041` | Existing-indicator link binds source without rewriting row. | transactional integration | `TODO: source not found` | `NF-FIX-007-indicator-linking` | `NFA-C02-001..002`; `NFA-C01-005` | row/binding invariant transcript | not run; target absent | `TODO` | Core indicator participation and implementation absent. |
| `NF-AC-042` | Create-indicator uses Core owner and enforces role. | authorization + integration | `TODO: source not found` | `NF-FIX-007-indicator-linking`; `NF-FIX-017-indicator-link-mismatch` | `NFA-C02-001..002`; `NFA-C04-001` | authorization/transaction transcript | not run; target absent | `TODO` | Canonical IP create/dedupe participation unresolved. |
| `NF-AC-043` | Duplicate link returns existing binding. | transactional integration | `TODO: source not found` | `NF-FIX-007-indicator-linking` | `NFA-C02-002`; `NFA-C01-005` | binding-count/replay transcript | not run; target absent | `TODO` | Atomic dedupe owner and implementation absent. |
| `NF-AC-044` | Rejected/deleted/stale/unmapped selectors fail. | integration | `TODO: source not found` | `NF-FIX-009-soft-delete-stale-graph`; `NF-FIX-017-indicator-link-mismatch`; `NF-FIX-025-graph-contributors` | `NFA-C02-002`; `NFA-GP-003`; `NFA-C04-001` | closed-selector error transcript | not run; target absent | `TODO` | Selector contracts and implementation absent. |
| `NF-AC-045` | Deployment admin without membership has no incident access. | authorization integration | `TODO: source not found` | `NF-FIX-025-graph-contributors` | `NFA-C04-001`; `NFA-TH-005` | route authorization matrix | not run; target absent | `TODO` | Network Flow hooks and auth-transition controls absent. |
| `NF-AC-046` | Network Flow actions perform no third-party egress. | security integration | `TODO: source not found` | n/a: egress-observation scenario | `NFA-C04-001`; `NFA-TH-001` | egress assertion report | not run; target absent | `TODO` | Module and harness observation absent. |
| `NF-AC-047` | Logs, telemetry, audit, and diagnostics obey raw-value policy. | security + integration | `TODO: source not found` | `NF-FIX-016-redaction` | `NFA-C04-003..004`; `NFA-TH-004` | redaction/leakage report | not run; target absent | `TODO` | Safe-digest lifecycle exists; audit occurrence owner, implementation, deterministic harness key, and fixture evidence absent. |
| `NF-AC-048` | Limits are discoverable, lowerable within bounds, and phase-enforced. | contract + integration | `TODO: source not found` | `NF-FIX-008-large-limits`; `NF-FIX-018-resource-limits` | `NFA-GEN-001`; `NFA-TH-001` | discovery and limit transcript | not run; target absent | `TODO` | Contracts, implementation, and fixtures absent. |
| `NF-AC-049` | Route errors include required details. | contract conformance | `TODO: source not found` | `NF-FIX-010-json-admission`; `NF-FIX-018-resource-limits` | `NFA-GEN-001` | exhaustive error-shape report | not run; target absent | `TODO` | Error contract family and selector absent. |
| `NF-AC-050` | Large timing remains engineering-only absent Core 05. | structural + harness accounting | `TODO: source not found` | `NF-FIX-008-large-limits` | `NFA-C05-001`; `NFA-TH-007` | evidence-classification assertion | not run; target absent | `TODO` | Fixture and accounting row absent; Core 05 not activated. |
| `NF-AC-051` | Unmapped raw remains inert provenance. | integration + security | `TODO: source not found` | `NF-FIX-011-alias-collision`; `NF-FIX-016-redaction` | `NFA-GEN-001`; `NFA-GP-002` | public-row/query/link transcript | not run; target absent | `TODO` | Row schema and implementation absent. |
| `NF-AC-052` | Every fixture has immutable bytes and expected artifacts before adoption. | structural + fixture drift | `TODO: source not found` | all `NF-FIX-001` through `NF-FIX-028` full IDs in §8 | `NFA-FIX-001..028`; `NFA-TH-001`; `NFA-VAL-003` | fixture completeness report | not run; target absent | `BLOCKED` | All 28 fixture rows retain `TODO:` fields. |
| `NF-AC-053` | Every MAY has omission behavior. | structural lint | `TODO: source not found` | n/a: owner document | `NFA-VAL-002`; `NFA-ADOPT-001` | normative-keyword report | not run; target absent | `TODO` | No Network Flow structural selector exists. |
| `NF-AC-054` | Normative behavior stays in adopted owner documents. | authority review + structural lint | `TODO: source not found` | n/a: document set | `NFA-AUTH-001`; `NFA-VAL-002` | authority-boundary review | not run; target absent | `TODO` | Contradictions are open and no selector exists. |
| `NF-AC-055` | Internal references resolve. | structural lint | `TODO: source not found` | n/a: owner document | `NFA-VAL-002` | link/reference report | not run; target absent | `TODO` | No Network Flow document lint selector exists. |
| `NF-AC-056` | Edge IDs and null-port aggregation match exact vectors. | unit + integration | `TODO: source not found` | `NF-FIX-006-cross-table-graph`; `NF-FIX-028-graph-aggregate-bounds` | `NFA-GP-001..003`; `NFA-TH-001` | edge-ID vector report | not run; target absent | `TODO` | Adapter contract, fixtures, and implementation absent. |
| `NF-AC-057` | Time-bucket request members remain unknown. | contract + route integration | `TODO: source not found` | `NF-FIX-010-json-admission` | `NFA-GEN-001` | unknown-member/error-catalog report | not run; target absent | `TODO` | Contract family and implementation absent. |
| `NF-AC-058` | Only binding observation mode; created observation refs empty. | contract + integration | `TODO: source not found` | `NF-FIX-007-indicator-linking` | `NFA-C02-002`; `NFA-GEN-001` | binding response transcript | not run; target absent | `TODO` | Core participation and response contract absent. |
| `NF-AC-059` | Graph digests ignore deployment/caller limit changes. | integration | `TODO: source not found` | `NF-FIX-006-cross-table-graph`; `NF-FIX-018-resource-limits` | `NFA-GP-001..003`; `NFA-TH-001` | paired digest transcript | not run; target absent | `TODO` | Graph adapter and immutable fixtures absent. |
| `NF-AC-060` | Timestamp precision, epochs, and uptime derivation are exact. | unit + fixture conformance | `TODO: source not found` | `NF-FIX-012-sys-uptime-timestamps`; `NF-FIX-022-timestamp-rulesets` | `NFA-TZ-001`; `NFA-TH-003` | timestamp vector report | not run; target absent | `TODO` | Immutable ruleset source exists; clocked fixtures and implementation remain absent. |
| `NF-AC-061` | Duplicate rename error and cursor survival are exact. | transactional integration | `TODO: source not found` | `NF-FIX-013-filename-display`; `NF-FIX-014-cursor-pagination` | `NFA-C04-002`; `NFA-C01-004`; `NFA-TH-002` | rename/cursor transcript | not run; target absent | `TODO` | Cursor ownership exists; table implementation and fixture transcript absent. |
| `NF-AC-062` | Only active/soft-deleted tables and exact limit accounting. | structural + integration | `TODO: source not found` | `NF-FIX-027-retention-soft-delete` | `NFA-C04-005`; `NFA-GEN-001` | lifecycle/count report | not run; target absent | `TODO` | Soft-delete lifecycle owner exists; generated schema, implementation, and fixtures absent. |
| `NF-AC-063` | All-rejected error has ordered safe diagnostics and no table. | transactional integration | `TODO: source not found` | `NF-FIX-004-rejected-rows` | `NFA-C01-004`; `NFA-TH-002`; `NFA-TH-004` | error and zero-state transcript | not run; target absent | `TODO` | Commit/diagnostic controls and implementation absent. |
| `NF-AC-064` | Query defaults and graph override bounds are exact. | contract + integration | `TODO: source not found` | `NF-FIX-018-resource-limits`; `NF-FIX-024-query-normalization-cursors` | `NFA-GEN-001`; `NFA-GP-001` | limit normalization transcript | not run; target absent | `TODO` | Contracts and implementation absent. |
| `NF-AC-065` | Preview slice counts and apply outcomes are independent. | integration | `TODO: source not found` | `NF-FIX-021-preview-boundaries` | `NFA-C01-002..005`; `NFA-TH-001` | preview/apply paired transcript | not run; target absent | `TODO` | Analytical import result union and fixture absent. |
| `NF-AC-066` | Binding source refs/truncation and duplicate selector handling are exact. | integration | `TODO: source not found` | `NF-FIX-007-indicator-linking` | `NFA-C02-002`; `NFA-GEN-001` | binding resource transcript | not run; target absent | `TODO` | Binding contract and implementation absent. |
| `NF-AC-067` | Dedupe uses resolved binding identity tuple. | transactional integration | `TODO: source not found` | `NF-FIX-007-indicator-linking` | `NFA-C02-002`; `NFA-C01-005` | binding identity/count transcript | not run; target absent | `TODO` | Atomic dedupe participation absent. |
| `NF-AC-068` | Exact-value mismatch fails before mutation with safe details. | transactional integration | `TODO: source not found` | `NF-FIX-017-indicator-link-mismatch` | `NFA-C02-001..002`; `NFA-TH-002` | error and zero-mutation transcript | not run; target absent | `TODO` | Canonical IP and transaction controls absent. |
| `NF-AC-069` | Adapter uses endpoint/edge IDs and no invalid retention token. | contract + integration | `TODO: source not found` | `NF-FIX-015-graph-adapter-input` | `NFA-GP-001..003` | adapter-input schema transcript | not run; target absent | `TODO` | Owner contract exists; generated schema, implementation, and fixture transcript absent. |
| `NF-AC-070` | Interface fields are bounded text/null and code-point sorted. | unit + integration | `TODO: source not found` | `NF-FIX-002-cisco-sna-interface-fields` | `NFA-GEN-001`; `NFA-TH-001` | row/order transcript | not run; target absent | `TODO` | Row contract, bytes, and implementation absent. |
| `NF-AC-071` | Aggregation and mapping combinability are closed to v1 values. | contract + route integration | `TODO: source not found` | `NF-FIX-010-json-admission`; `NF-FIX-011-alias-collision` | `NFA-GEN-001` | admission-error transcript | not run; target absent | `TODO` | Closed contract family absent. |
| `NF-AC-072` | Rename/delete/link/apply idempotency follows exact comparison/replay points. | transactional integration | `TODO: source not found` | `NF-FIX-020-atomic-import-commit`; `NF-FIX-026-audit-and-replay` | `NFA-C01-004..005`; `NFA-C02-002`; `NFA-TH-002` | idempotency matrix | not run; target absent | `TODO` | Generic terminal publication and transaction evidence absent. |
| `NF-AC-073` | Mapping accepts only three variants and rejects extras/sentinels. | contract + integration | `TODO: source not found` | `NF-FIX-011-alias-collision` | `NFA-C01-003`; `NFA-GEN-001` | mapping admission report | not run; target absent | `TODO` | Analytical import mapping contract absent. |
| `NF-AC-074` | Alias keys/warnings require explicit complete approval. | parser + browser + integration | `TODO: source not found` | `NF-FIX-011-alias-collision` | `NFA-C01-003`; `NFA-C03-001` | suggestion/warning/approval transcript | not run; target absent | `TODO` | Import facade/UI and fixture absent. |
| `NF-AC-075` | Every success data object matches exact closed schema. | contract conformance | `TODO: source not found` | `NF-FIX-001-cisco-sna-minimal`; `NF-FIX-023-import-facade-source-change`; `NF-FIX-025-graph-contributors` | `NFA-C01-004`; `NFA-GEN-001` | exhaustive success-schema report | not run; target absent | `TODO` | Network Flow contract family and generated types absent. |
| `NF-AC-076` | All resource limits match defaults, phases, failures, and invalid config. | integration + configuration | `TODO: source not found` | `NF-FIX-008-large-limits`; `NF-FIX-018-resource-limits` | `NFA-GEN-001`; `NFA-TH-001`; `NFA-TH-007` | limit matrix report | not run; target absent | `TODO` | Configuration contract, fixtures, and executor absent. |
| `NF-AC-077` | Safe/source samples follow exact raw/numeric/null rules. | unit + security | `TODO: source not found` | `NF-FIX-016-redaction` | `NFA-C04-003`; `NFA-TH-004` | sample/redaction vectors | not run; target absent | `TODO` | Safe-digest key lifecycle exists; implementation, fixture secret control, and vectors absent. |
| `NF-AC-078` | Route errors have exact details and deterministic precedence. | contract + integration | `TODO: source not found` | `NF-FIX-004-rejected-rows`; `NF-FIX-018-resource-limits` | `NFA-GEN-001`; `NFA-TH-004` | exhaustive error-order report | not run; target absent | `TODO` | Error family and deterministic scheduling absent. |
| `NF-AC-079` | Linking accepts only IP endpoints and validates Core identity. | integration | `TODO: source not found` | `NF-FIX-007-indicator-linking`; `NF-FIX-017-indicator-link-mismatch` | `NFA-C02-001..002` | selector/target vector report | not run; target absent | `TODO` | Core canonical IP participation unresolved. |
| `NF-AC-080` | Filename/display naming cases match exact rules. | unit + integration | `TODO: source not found` | `NF-FIX-013-filename-display` | `NFA-C01-003`; `NFA-TH-001` | naming vector transcript | not run; target absent | `TODO` | Implementation and immutable fixture absent. |
| `NF-AC-081` | Final commit faults expose no partial state/audit. | fault + transactional integration | `TODO: source not found` | `NF-FIX-020-atomic-import-commit` | `NFA-C01-004`; `NFA-TH-002` | commit-step fault matrix | not run; target absent | `TODO` | Atomic UoW and named fault boundaries absent. |
| `NF-AC-082` | Explicit duplicate names fail; omitted names suffix under commit lock. | concurrency + transactional integration | `TODO: source not found` | `NF-FIX-013-filename-display`; `NF-FIX-020-atomic-import-commit` | `NFA-C01-004`; `NFA-TH-002`; `NFA-TH-004` | concurrent naming transcript | not run; target absent | `TODO` | Unit of work and deterministic scheduler absent. |
| `NF-AC-083` | Preview stops at 50 complete data records with exact counting. | parser conformance | `TODO: source not found` | `NF-FIX-021-preview-boundaries` | `NFA-C01-003`; `NFA-TH-001` | preview-boundary report | not run; target absent | `TODO` | Fixture manifest, parser, and selector absent. |
| `NF-AC-084` | Row limit counts logical records and stops at `limit+1`. | parser + integration | `TODO: source not found` | `NF-FIX-021-preview-boundaries` | `NFA-TH-001` | limit-boundary report | not run; target absent | `TODO` | Immutable fixture and implementation absent. |
| `NF-AC-085` | Cisco profile allows only required plus two interface targets. | parser + contract | `TODO: source not found` | `NF-FIX-002-cisco-sna-interface-fields`; `NF-FIX-011-alias-collision` | `NFA-C01-003`; `NFA-GEN-001` | target-mapping report | not run; target absent | `TODO` | Mapping contract and implementation absent. |
| `NF-AC-086` | ASCII-space trim and empty policy are exact. | unit conformance | `TODO: source not found` | `NF-FIX-019-canonical-json-unicode` | `NFA-TH-001` | transform vectors | not run; target absent | `TODO` | Unicode/byte fixture and implementation absent. |
| `NF-AC-087` | Timestamp variants, grammar, ruleset, uptime, and ordinals are closed. | contract + unit conformance | `TODO: source not found` | `NF-FIX-012-sys-uptime-timestamps`; `NF-FIX-022-timestamp-rulesets` | `NFA-TZ-001`; `NFA-GEN-001`; `NFA-TH-003` | timestamp contract/vector report | not run; target absent | `TODO` | Immutable tzdb source exists; generated contracts, timestamp vectors, and implementation remain absent. |
| `NF-AC-088` | Preview is side-effect-free; apply uses opaque source and fails closed on change. | integration + fault | `TODO: source not found` | `NF-FIX-023-import-facade-source-change` | `NFA-C01-003..005`; `NFA-TH-002` | preview/apply/source-change transcript | not run; target absent | `TODO` | Core 01 import shapes are incompatible and closed. |
| `NF-AC-089` | Descriptors/dispositions/mappings and provenance are exact. | contract + integration | `TODO: source not found` | `NF-FIX-003-duplicate-headers`; `NF-FIX-011-alias-collision`; `NF-FIX-023-import-facade-source-change` | `NFA-C01-002..005`; `NFA-GEN-001` | descriptor/mapping/provenance report | not run; target absent | `TODO` | Analytical import target/result union absent. |
| `NF-AC-090` | Public rows include nullable fields, unmapped values, and observation ref. | contract + integration | `TODO: source not found` | `NF-FIX-001-cisco-sna-minimal`; `NF-FIX-011-alias-collision` | `NFA-GEN-001` | public-row schema transcript | not run; target absent | `TODO` | Network Flow row contract and implementation absent. |
| `NF-AC-091` | Parallel diagnostic discovery is byte-identical and ordered. | concurrency + integration | `TODO: source not found` | `NF-FIX-004-rejected-rows` | `NFA-TH-004` | repeated-run byte comparison | not run; target absent | `TODO` | Deterministic scheduling/randomness and implementation absent. |
| `NF-AC-092` | Table scopes reject invalid variants without resource disclosure. | authorization + contract integration | `TODO: source not found` | `NF-FIX-024-query-normalization-cursors` | `NFA-C04-001`; `NFA-GEN-001`; `NFA-TH-005` | scope/admission/auth transcript | not run; target absent | `TODO` | Route hooks, contract, and auth controls absent. |
| `NF-AC-093` | Filter normalization canonicalizes and rejects canonical duplicates. | unit + integration | `TODO: source not found` | `NF-FIX-019-canonical-json-unicode`; `NF-FIX-024-query-normalization-cursors` | `NFA-C02-001`; `NFA-GEN-001` | canonical query vectors | not run; target absent | `TODO` | Canonical IP token and query implementation absent. |
| `NF-AC-094` | Initial/continuation variants and token byte/expiry bounds are exact. | security + integration | `TODO: source not found` | `NF-FIX-014-cursor-pagination`; `NF-FIX-024-query-normalization-cursors` | `NFA-C04-002`; `NFA-TH-003`; `NFA-GEN-001` | cursor boundary report | not run; target absent | `TODO` | Cursor ownership/key lifecycle exists; generated contract, implementation, and fake-clock coverage absent. |
| `NF-AC-095` | Three cursor families use independent full keysets without gaps/duplicates. | integration | `TODO: source not found` | `NF-FIX-014-cursor-pagination`; `NF-FIX-024-query-normalization-cursors`; `NF-FIX-025-graph-contributors` | `NFA-C04-002`; `NFA-GP-003`; `NFA-TH-001` | full pagination transcript | not run; target absent | `TODO` | Cursor and graph interfaces plus fixtures absent. |
| `NF-AC-096` | Arbitrary-precision aggregation and fixed failure order have no partial output. | unit + integration | `TODO: source not found` | `NF-FIX-028-graph-aggregate-bounds` | `NFA-GP-003`; `NFA-TH-001` | aggregate vectors/error transcript | not run; target absent | `TODO` | Adapter result/error contract and bytes absent. |
| `NF-AC-097` | Projection metadata and outcome mapping are exact and safe. | contract + integration | `TODO: source not found` | `NF-FIX-015-graph-adapter-input` | `NFA-GP-001..003` | metadata/outcome transcript | not run; target absent | `TODO` | Owner mapping/outcome text exists; generated schema, implementation, and fixture transcript absent. |
| `NF-AC-098` | Graph success schema is exact and closed. | contract + integration | `TODO: source not found` | `NF-FIX-025-graph-contributors` | `NFA-GP-001..003`; `NFA-GEN-001` | graph success-schema report | not run; target absent | `TODO` | Adapter and Network Flow contract families absent. |
| `NF-AC-099` | Contributors recompute composition/auth and paginate without fallback. | authorization + integration | `TODO: source not found` | `NF-FIX-025-graph-contributors` | `NFA-GP-001..003`; `NFA-C03-002`; `NFA-TH-005` | contributor/auth transcript | not run; target absent | `TODO` | Ephemeral interface and auth-transition control absent. |
| `NF-AC-100` | Indicator variants and exact confirmation are closed; create input is constrained. | contract + integration | `TODO: source not found` | `NF-FIX-017-indicator-link-mismatch` | `NFA-C02-001..002`; `NFA-GEN-001` | selector/target contract report | not run; target absent | `TODO` | Core canonical IP token and Network Flow contract absent. |
| `NF-AC-101` | Indicator create/dedupe plus binding is atomic with exact statuses. | transactional integration | `TODO: source not found` | `NF-FIX-007-indicator-linking`; `NF-FIX-026-audit-and-replay` | `NFA-C02-002`; `NFA-C01-004`; `NFA-TH-002` | transaction/status transcript | not run; target absent | `BLOCKED` | Core 02 atomic participation is not adopted. |
| `NF-AC-102` | Safe digests carry key IDs and compare only within a key ID. | security + unit/integration | `TODO: source not found` | `NF-FIX-016-redaction`; `NF-FIX-026-audit-and-replay` | `NFA-C04-003`; `NFA-TH-004` | rotation/comparison vectors | not run; target absent | `BLOCKED` | Safe-digest key lifecycle is owner-defined; implementation, harness key control, and vectors absent. |
| `NF-AC-103` | Audit occurrences/replay/truncation counts are exact. | transactional + audit integration | `TODO: source not found` | `NF-FIX-026-audit-and-replay` | `NFA-C04-004`; `NFA-C01-005`; `NFA-TH-006` | audit-count/replay transcript | not run; target absent | `BLOCKED` | Exact occurrence owner contract exists; implementation, harness count assertions, and terminal replay transcript absent. |
| `NF-AC-104` | Soft delete and incident closure retain but make Network Flow data non-queryable, and v1 exposes no incident-purge claim. | retention + integration | `TODO: source not found` | `NF-FIX-009-soft-delete-stale-graph`; `NF-FIX-027-retention-soft-delete` | `NFA-C00-002`; `NFA-C04-005`; `NFA-TH-005`; `NFA-TH-006` | soft-delete/closure retention report | not run; target absent | `BLOCKED` | Core 03 invalidation and Core 04 retention owners exist; closure/auth/audit harness controls, implementation, and fixtures absent. |
| `NF-AC-105` | Every route status/schema/error/reason/detail/retry is exact. | generated-contract + integration | `TODO: source not found` | all applicable route fixtures in §8 | `NFA-GEN-001`; `NFA-C01-004`; `NFA-C04-001` | exhaustive route conformance report | not run; target absent | `TODO` | No Network Flow contract family, generated types, routes, or selector. |
| `NF-AC-106` | All locators, blockers, and immutable fixtures close before adoption. | structural + adoption review | `TODO: source not found` | all `NF-FIX-001` through `NF-FIX-028` full IDs in §8 | `NFA-LOC-001`; `NFA-FIX-001..028`; `NFA-ADOPT-001` | zero-open-dependency adoption report | not run; target absent | `BLOCKED` | Seven locators, 17 blockers, and 28 byte freezes are open. |
| `NF-AC-107` | Pre-commit cancellation leaves nothing; post-commit recovery publishes once. | worker fault + transactional integration | `TODO: source not found` | `NF-FIX-020-atomic-import-commit`; `NF-FIX-023-import-facade-source-change` | `NFA-C01-004..005`; `NFA-TH-004`; `NFA-TH-007` | cancellation/recovery transcript | not run; target absent | `BLOCKED` | Terminal publication and worker-fault controls are absent. |

The adoption task consumes each row individually; a passing broad target cannot
mask a missing selector, fixture, expected artifact, or dependency task.

## 10. Generated artifacts and anti-drift inventory

`contracts/index.json` declares generated contract families. The five active
families remain `openapi`, `ws`, `view-schemas`, `errors`, and `extensions`.
Network Flow is declared as a planned family under `contracts/network-flow/`
and is not emitted until authored route, error, schema, and discovery inputs
activate it through `NFA-GEN-002..004`. Generated roots must never be
hand-edited.

| Generator/target | Authored inputs | Generated outputs | Consumers | Current drift | Regeneration command | Verification command | Owner task | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `tools/contractgen` OpenAPI family | `contracts/openapi/cartulary.openapi.yaml` | `internal/gen/contracts/contracts_gen.go`; `packages/protocol-ts/src/generated/contracts.ts`; generated TS index | HTTP server, Go callers, web clients | Current input/output exists; no Network Flow routes | `make generate` | `make generate-drift` | `NFA-GEN-001..004` | `BLOCKED` |
| `tools/contractgen` WebSocket family | `contracts/ws/index.schema.json` | Same generated Go/TS contract roots | collaboration transport and web client | No Network Flow invalidation event | `make generate` | `make generate-drift` | `NFA-C03-002`; `NFA-GEN-001..004` | `BLOCKED` |
| `tools/contractgen` view-schema family | `contracts/view-schemas/index.json`; observed `contracts/view-schemas/*.json` | Same generated Go/TS contract roots | workbook/view-schema consumers | Existing family is not authority to model extension tables as saved views | `make generate` | `make generate-drift` | `NFA-GEN-001`; `NFA-AUTH-001` | `BLOCKED` |
| `tools/contractgen` error family | `contracts/errors/index.json` | Same generated Go/TS contract roots | server error registry and clients | No Network Flow error family/members | `make generate` | `make generate-drift`; `make json-shape-check` | `NFA-GEN-001..004` | `BLOCKED` |
| `tools/contractgen` extension family | `contracts/extensions/index.json` | Same generated Go/TS contract roots | extension discovery and clients | Closed to current profiles; no Network Flow profile | `make generate` | `make generate-drift`; `make json-shape-check` | `NFA-C00-001`; `NFA-C01-001`; `NFA-GEN-001..004` | `BLOCKED` |
| Network Flow authored contract ownership | `contracts/index.json`; `contracts/network-flow/index.json`; `contracts/network-flow/routes.v1.json`; `contracts/network-flow/errors.v1.json`; `contracts/network-flow/schemas.v1.json`; `contracts/network-flow/timezone/tzdb-2026c.provenance.json` | `internal/gen/contracts/contracts_gen.go`; `packages/protocol-ts/src/generated/contracts.ts`; generated TS index remains unchanged | Future module, routes, web UI, and harness | Family is active and emitted; generated outputs contain Network Flow markers; final generated-drift checkpoint remains `NFA-GEN-004` | `make generate` | `make json-shape-check`; `make generate-drift` | `NFA-GEN-001..004` | `DONE for ownership, authored inputs, and generated outputs; drift checkpoint pending` |
| Graph Projection conformance evidence | `contracts/graph-projection/conformance_matrix.v1.json`; `contracts/graph-projection/fixtures/corpus.v1.json`; no generator observed | n/a: observed files are authored evidence | `tools/harness/generated-artifacts/check-json-shapes.mjs` and reviewers | Artifact `81941bba`: adopted NLSpec has 69 `GP-AC-*` and 36 `GP-FIX-*`; matrix has 69 AC and 36 fixtures; corpus has 36 fixtures; checker enforces both matrix and corpus ID ranges | n/a: no generator observed | `make json-shape-check` `.cartulary/test-results/20260710T061106Z-p35273` | `NFA-GP-003` | `DONE` |
| Network Flow fixture manifests and transcripts | `TODO: source not found`; no harness-owned schema, bytes, mappings, or transcript convention observed | `TODO: source not found`; no generated corpus or matrix exists | Future Network Flow conformance target | All 28 fixture and 107 AC evidence families absent | `TODO: source not found` | `TODO: source not found` | `NFA-TH-001`; `NFA-TRANSCRIPT-001`; `NFA-VAL-002` | `BLOCKED` |
| Generated-artifact policy | `tools/generated_artifact_policy.json`; `tools/generate_drift_scratch_inputs.json`; `contracts/index.json` | Policy-defined generated roots and scratch comparison | repository generation/drift harness | Family registry is schema-checked; policy roots unchanged; Network Flow output drift enforcement remains later | `make generate` | `make generated-artifact-policy-check`; `make json-shape-check`; `make generate-drift` | `NFA-GEN-004`; `NFA-TH-007` | `BLOCKED` |

The former Graph Projection mismatch is closed by `NFA-GP-003`: the matrix,
corpus, and JSON-shape checker now align to 69 acceptance rows and 36 fixture
rows.

## 11. Execution checkpoints

Each checkpoint is one reviewable seam. A downstream checkpoint may prepare a
design, but it cannot freeze behavior or bytes before its owner preconditions.

| Checkpoint | Owner/task IDs | Edit scope | Preconditions | Validation | Expected diff/artifact | Rollback/stop point | Handoff-ready result |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `CP-00` | `NFA-INV-001`; `NFA-AUTH-001` | Tracker snapshot and crosswalk only | Clean ownership of local diff | §15 tracker checks | One tracker diff | Stop on unexplained dirty overlap | Refreshed attributable baseline |
| `CP-01` | `NFA-C00-001` | Core 00 extension-profile model | Owner review; no status flip | Owner structural target `TODO:` | One Core owner amendment | Withhold downstream contract generation | Bounded profile model decided |
| `CP-02` | `NFA-C00-002`; `NFA-C02-003` | Core 00 future-only purge posture | `CP-01`; Core 02/Core 04 review | Owner structural target `TODO:` | Explicit v1 non-purge decision, not final registry flip | Stop if any current purge claim remains | Future generic cascade obligation without a v1 compatibility promise |
| `CP-03` | `NFA-C01-001` | Core 01 discovery resource | `CP-01` | Owner contract target `TODO:` | Discovery schema/criteria amendment | Stop before contracts if claim behavior unresolved | Exact claimed/unclaimed discovery contract |
| `CP-04` | `NFA-C01-002` | Core 01 analytical target/result union | `CP-01`; domain-boundary review | Owner contract target `TODO:` | Extension resource union amendment | Stop if table is coerced into record/view | Exact non-record result contract |
| `CP-05` | `NFA-C01-003` | Opaque stream, preview/apply, source-change facade | `CP-04` | Owner contract target `TODO:` | Facade capability amendment | Stop on path/URL/raw stream exposure | Closed preview/apply contract |
| `CP-06` | `NFA-C01-004`; `NFA-C02-002`; `NFA-C04-004` | Cross-owner atomic unit of work | `CP-05`; indicator and audit decisions | Fault target `TODO:` | Transaction participant contract | Stop before implementation without rollback semantics | Named atomic commit boundaries |
| `CP-07` | `NFA-C01-005` | Terminal publication/recovery | `CP-04`; `CP-06` | Worker-fault target `TODO:` | Result and publication contract | Stop on duplicate/ghost publication ambiguity | Exactly-once terminal result semantics |
| `CP-08` | `NFA-C02-001` | Canonical IP token/algorithm | `CP-01` | Canonical-vector target `TODO:` | Core 02 registry/amendment | Stop before fixture IP bytes freeze | Exact IPv4/IPv6 identity contract |
| `CP-09` | `NFA-C02-002` | Indicator create/dedupe participation | `CP-06`; `CP-08` | Fault/link target `TODO:` | Transaction participant amendment | Stop on observation or record expansion | Atomic binding participation contract |
| `CP-10` | `NFA-C02-003` | Drop private incident-purge cascade | `CP-02`; retention review | Structural owner target `TODO:` | Recorded future-only decision | Stop if Network Flow still claims current incident purge | No v1 purge dependency; future generic cascade remains named |
| `CP-11` | `NFA-C03-001` | Claimed extension workspace/tab | `CP-01`; `CP-03` | Browser selector `TODO:` | Core 03 extension surface amendment | Stop if base tab list would expand | Claimed-only extension surface contract |
| `CP-12` | `NFA-C03-002` | Extension-resource invalidation | `CP-11`; route authorization | Browser/wire selector `TODO:` | Core 03 and optional Core 01 wire amendment | Stop on event identity/hidden-resource ambiguity | Exact rename/delete/auth-loss consequences |
| `CP-13` | `NFA-C04-001` | Network Flow route authorization | `CP-03` | Route matrix target `TODO:` | Core 04 authorization amendment | Stop on deployment-admin bypass | Exact current-membership route matrix |
| `CP-14` | `NFA-C04-002` | Cursor wire/security ownership | `CP-13`; Core 01/Core 04 joint review | Artifact `90401fb2`; §15.7 evidence | Explicit owner split and lifecycle | Stop before token/schema freeze | Confidential, protected, rotating cursor contract |
| `CP-15` | `NFA-C04-003` | Safe-digest key lifecycle | `CP-13` | Artifact `cd645750`; §15.8 evidence | Secret namespace/key-ID amendment | Stop before fixture key selection | Production/fixture key boundary defined |
| `CP-16` | `NFA-C04-004` | Transactional audit occurrences | `CP-06`; `CP-15` | Artifact `71258589`; §15.9 evidence | Audit/outbox occurrence amendment | Stop if exact replay count is unresolved | Exact transactional audit contract |
| `CP-17` | `NFA-C04-005` | Soft-delete and import-source retention hooks | `CP-10`; `CP-16` | Artifact `663c8684`; §15.10 evidence | Retention amendment and purge omission | Stop on conflict with Core-retained audit or source expiry | Exact v1 retention boundary without incident purge |
| `CP-18` | `NFA-GP-001` | Non-retained projection invocation/result | `WF-02` complete | Graph Projection owner target `TODO:` | Ephemeral owner amendment | Stop before adapter code | Exact no-retained-state operation |
| `CP-19` | `NFA-GP-002` | Adapter properties, metadata, results, errors | `CP-18` | Adapter contract target `TODO:` | Compatible schema and outcome map | Stop on direct-copy schema contradiction | Exact safe adapter boundary |
| `CP-20` | `NFA-GP-003` | Graph Projection matrix/corpus/checker alignment | Owner-approved GP criteria | `make json-shape-check` | 69-AC/36-fixture aligned evidence | Revert derived evidence if owner text changes | No owner/evidence count drift |
| `CP-21` | `NFA-TH-001` | Fixture manifest and runner contract | `WF-02` complete | `make harness-contract` after adoption | Harness schema/runner amendment | Stop before inventing manifest format | Ordered per-file and aggregate hash mechanics |
| `CP-22` | `NFA-TH-003` | Network Flow fake-clock usage | `CP-21` | Harness selector `TODO:` | Clock reset/boundary evidence | Stop if production clock can be changed | TTL/timezone/retention boundaries controllable |
| `CP-23` | `NFA-TH-004` | Deterministic fixture randomness/collisions | `CP-21`; `CP-15` | Harness selector `TODO:` | Test-only random source contract | Stop on production import or secret leakage | Repeatable IDs/collisions with reset |
| `CP-24` | `NFA-TH-002`; `NFA-C01-004` | Every final-commit fault boundary | `CP-06`; `CP-21` | Fault selector `TODO:` | Named one-shot commit fault controls | Stop before claiming atomicity | Complete commit failure matrix |
| `CP-25` | `NFA-TH-002`; `NFA-C01-005` | Worker crash/cancellation timing | `CP-07`; `CP-21` | Worker selector `TODO:` | Crash/cancel/recovery controls | Stop before claiming exactly-once publication | Complete worker recovery matrix |
| `CP-26` | `NFA-TH-005`; `NFA-TH-006` | Auth transitions and audit-count helpers | `CP-12`; `CP-13`; `CP-16`; `CP-21` | Harness selectors `TODO:` | Executable helpers and reset evidence | Stop on hidden-resource or count leakage | Exact route-time auth and audit assertions |
| `CP-27` | `NFA-TZ-001` | `tzdb-2026c` provenance and immutable source | Owner-approved source selection | `make json-shape-check` | Provenance, revision, license, digest; artifact `42338a1d` | Stop on mutable/unlicensed/host fallback | One attributable ruleset input |
| `CP-28` | `NFA-LOC-001` | Seven dependency versions/locators | Relevant owner amendments adopted | Structural reference target `TODO:` | Complete Table 1-B | Stop if any locator is mutable or unresolved | Seven resolving immutable locators |
| `CP-29` | `NFA-GEN-001..004` | Authored contracts, generator, generated outputs, drift | All affected owner checkpoints | `make generate`; `make generate-drift` | Owner-derived authored/generated diff | Stop before generator on public-shape ambiguity; never hand-edit outputs | Clean deterministic regeneration |
| `CP-30` | `NFA-FIX-001..005` | Minimal/parser/rejection fixture family | Relevant Core and `CP-21` | Network Flow selector `TODO:` | Five single-file byte/hash/mapping rows | Stop before hash on owner/byte ambiguity | Immutable ingestion/parser family |
| `CP-31` | `NFA-FIX-006..009` | Graph/link/limit/lifecycle fixture family | GP, indicator, security, `CP-21` | Network Flow selector `TODO:` | Four fixture rows, including two manifests | Stop before hash on dependency ambiguity | Immutable graph/link/lifecycle family |
| `CP-32` | `NFA-FIX-010..014` | Admission/mapping/time/name/cursor family | Contract, tzdb, cursor, `CP-21` | Network Flow selector `TODO:` | Five immutable rows | Stop before hash on owner/clock ambiguity | Immutable admission/query family |
| `CP-33` | `NFA-FIX-015..019` | Adapter/redaction/link/limits/Unicode family | GP, Core 02/Core 04, `CP-21` | Network Flow selector `TODO:` | Five immutable rows, two manifests | Stop before hash on schema/key/Unicode ambiguity | Immutable adapter/security family |
| `CP-34` | `NFA-FIX-020..024` | Atomic/preview/time/source/cursor family | Core 01, cursor, tzdb, fault controls | Network Flow selector `TODO:` | Five immutable manifest/file rows | Stop before hash on schedule/time ambiguity | Immutable transaction/time family |
| `CP-35` | `NFA-FIX-025..028` | Contributor/audit/soft-delete/aggregate family | GP, audit, retention, auth controls | Network Flow selector `TODO:` | Four manifest-backed rows | Stop before hash on owner/count ambiguity | Immutable final graph/lifecycle family |
| `CP-36` | `NFA-TRANSCRIPT-001` | Expected outputs for all fixture rows | `CP-29..35` applicable outputs fixed | Network Flow transcript target `TODO:` | Canonical mapping and expected-output corpus | Stop on any input, contract, or generator drift | Every §22 output is immutable and referenced |
| `CP-37` | `NFA-TEST-001..007`; all AC rows | Executable coverage by test level | Owner, implementation, fixture, and harness dependencies | Network Flow target `TODO:` | 107 selector/result rows | Stop if any row is grouped away or unevidenced | One attributable result per AC |
| `CP-38` | `NFA-VAL-002`; `NFA-VAL-003` | Full validation/security/fault/drift bundle | `CP-20`; `CP-29`; `CP-36`; `CP-37` | Public Network Flow target `TODO:` | Retained classified run root | Stop on any product, harness, infra, fixture, or drift failure | Complete adoption evidence bundle |
| `CP-39` | `NFA-ADOPT-001`; `NFA-C00-002` | Coordinated Core 00 registry and NLSpec version/status | Every gate, blocker, locator, fixture, generated check, AC, and validation closed | Final status target `TODO:` | One reviewed status/registry transition | Withhold or revert both sides together | Network Flow adoption-ready/current transition |
| `CP-40` | `NFA-HANDOFF-001` | Tracker notes and next slice | Every session | `git diff --name-only` plus §15 checks | Current handoff record | Stop before session end if stale | Safe restart at one named checkpoint |

## 12. Validation and evidence accounting

Commands are repository-root, Make-owned unless explicitly described as
read-only tracker checks. Current-session outcomes are filled after this file is
validated; no future Network Flow target is claimed to exist.

| Target/command | Purpose | Required at which checkpoint | Expected artifact | Current baseline | Failure classification | Retention location |
| --- | --- | --- | --- | --- | --- | --- |
| `make agent-finalize` with `RESULTS_DIR` unset | Refresh/validate harness maintenance before end-run checks | `CP-00`, `CP-40` | retained phase summary; explicit retained-run skip | `TODO: run after tracker draft` | harness defect or infrastructure/configuration failure | `.cartulary/test-results/<run>/` |
| `make generated-artifact-policy-check` | Enforce generated-root ownership | `CP-00`, `CP-29`, `CP-38` | retained passing summary | Baseline pass before creation at `.cartulary/test-results/20260710T030255Z-p60440` | generated-artifact drift or harness defect | named run root |
| `make json-shape-check` | Validate current JSON/contract shapes | `CP-00`, `CP-20`, `CP-29`, `CP-38` | retained passing/failing summary | Baseline pass before creation at `.cartulary/test-results/20260710T030256Z-p60474`; checker still enforces GP 68/23 | owner-contract contradiction, schema drift, or harness defect | named run root |
| `make lint-markdown` | Run configured Markdown lint | `CP-00`, `CP-40` | target summary | Baseline pass before creation; configured globs omit `docs/handoffs/**/*.md` | documentation defect or configuration gap | target summary/cache artifact |
| `make generate-drift` | Prove generated outputs reproduce | `CP-29`, `CP-38` | retained drift summary | Not run; no generated source changed | generated-artifact drift | named run root |
| `make harness-contract` | Prove adopted harness mechanics | `CP-21..26`, `CP-38` | retained harness contract run | Not run; no harness change | harness defect | named run root |
| Future Network Flow public target `TODO: source not found` | Execute fixtures and 107 AC rows | `CP-30..38` | row-accounted retained run | Unavailable dependency | product failure, fixture mismatch, harness defect, or infrastructure failure | owner-selected retained run root `TODO:` |
| `git diff --check -- docs/handoffs/network-flow-activity-adoption-handoff-tracker.md` | Whitespace integrity beyond configured Markdown globs | `CP-00`, `CP-40` | zero exit | `TODO: run after tracker draft` | documentation defect | terminal output |
| Read-only table-column consistency check in §15 | Ensure every Markdown table row matches its header width | `CP-00`, `CP-40` | zero inconsistent tables | `TODO: run after tracker draft` | tracker structural defect | terminal output |
| Read-only ID/count check in §15 | Prove 7/12/17/28/107 and contiguous owner IDs | `CP-00`, `CP-40` | exact counts and no missing/duplicate IDs | `TODO: run after tracker draft` | tracker structural defect | terminal output |
| Existing/absent path check in §15 | Prove observed paths exist and proposals are marked absent | `CP-00`, `CP-40` | all existing paths resolve; absent references include `TODO: source not found` | `TODO: run after tracker draft` | tracker inventory defect | terminal output |
| `git diff --name-only` | Preserve one-file task scope | `CP-00`, `CP-40` | only tracker path | `TODO: run after tracker draft` | scope violation | terminal output |

Failure triage must choose one primary class: product failure; owner-contract
contradiction; generated-artifact drift; fixture-byte/transcript mismatch;
harness defect; infrastructure/configuration failure; or unavailable dependency.
Do not rewrite owner text or expectations merely to make a failing command pass.

## 13. Risks, assumptions, and blockers

| ID | Risk/assumption/blocker | Owner | Affected IDs/tasks | Security/compatibility impact | Resolution condition | Status |
| --- | --- | --- | --- | --- | --- | --- |
| `NFA-RISK-001` | Draft behavior contradicts closed adopted Core shapes. | Core 00–04 | `NFA-C00-*` through `NFA-C04-*` | Private extension choices could break Core compatibility. | Each contradiction receives an adopted owner decision. | `BLOCKED` |
| `NFA-RISK-002` | Core 00 omits Network Flow and keeps whole-incident purge future-only. | Core 00 | `NFA-C00-001..002`; `NFA-ADOPT-001` | Premature discovery/adoption or an unsupported deletion promise. | Adopt the profile boundary while preserving the future-only generic incident-removal decision. | `BLOCKED` |
| `NFA-RISK-003` | Flow resources could be coerced into Core record/view semantics. | Core 01; Network Flow | `NFA-C01-002..005` | Identity, storage, envelope, and compatibility corruption. | Adopt an analytical extension resource/result boundary. | `BLOCKED` |
| `NFA-RISK-004` | Cross-owner import/indicator/binding/audit commit lacks one unit of work. | Core 01/Core 02/Core 04 | `NFA-C01-004`; `NFA-C02-002`; `NFA-C04-004`; `NFA-TH-002` | Ghost or partial state and mismatched audit. | Owner participants exist; implementation and injected-fault proof for every boundary remain required. | `BLOCKED` |
| `NFA-RISK-005` | Cursor wire ownership and security lifecycle are split. | Core 01/Core 04 | `NFA-C04-002`; `NFA-IMPL-015`; `NFA-TH-003` | Disclosure, forgery, stale access, incompatible rotation. | Owner split is assigned in `90401fb2`; executable rotation/expiry evidence remains with implementation and harness rows. | `DONE` |
| `NFA-RISK-006` | Secrets or raw flow values could enter fixtures/logs/manifests/transcripts. | Core 04/Harness | `NFA-C04-003`; `NFA-FIX-016`; `NFA-FIX-026` | Credential or incident-data disclosure. | Core 04 owner boundary exists in `cd645750`; deterministic harness material and zero-leak fixture validation remain required. | `BLOCKED` |
| `NFA-RISK-007` | Soft-delete retention, idempotency, terminal publication, raw-source expiry, and audit retention may disagree. | Core 01/Core 04 | `NFA-C01-003..005`; `NFA-C04-004..005` | Replay resurrection, duplicate audit, source over-retention, or inaccessible retained-table saturation. | Audit and retention owners exist; implementation, retained-count instrumentation, source cleanup, and exact fixture evidence remain required. | `BLOCKED` |
| `NFA-RISK-008` | Draft Graph Projection metadata/type wording is incompatible. | Graph Projection/Network Flow | `NFA-GP-002`, `NFA-GP-003` | Adapter rejection, leakage, or private type fork. | Owner/downstream text now rejects private direct-copy members and defines exact property, metadata, result, and outcome mapping; Graph Projection evidence drift is closed. | `MITIGATED` |
| `NFA-RISK-009` | Graph Projection evidence is already stale at 68/23 versus 69/36. | Graph Projection/Harness | `NFA-GP-003` | False-positive JSON-shape validation. | Matrix, corpus, and checker now align to 69 AC and 36 fixtures in artifact `81941bba`. | `MITIGATED` |
| `NFA-RISK-010` | Future fixture generators/manifests may be non-reproducible. | Testing Harness | `NFA-TH-001`; `NFA-FIX-001..028` | Hash and transcript instability. | Owner schema freezes entry order, per-file/aggregate hashes, revision, and outputs. | `BLOCKED` |
| `NFA-RISK-011` | Host timezone, locale, collation, Unicode, or line endings may affect bytes. | Network Flow/Harness | `NFA-TZ-001`; `NFA-FIX-005`; `NFA-FIX-019`; `NFA-FIX-022` | Cross-platform digest and behavior drift. | Freeze all environmental inputs and use immutable `tzdb-2026c`. | `BLOCKED` |
| `NFA-RISK-012` | IDs, time, CSPRNG, scheduling, or ordering may destabilize transcripts. | Testing Harness | `NFA-TH-002..006`; all fixtures | Non-repeatable acceptance evidence. | Fake clock/randomness/fault/scheduler/auth/audit controls are resettable. | `BLOCKED` |
| `NFA-RISK-013` | Generated outputs may be hand-edited or a family may bypass generator ownership. | Contract owners/Harness | `NFA-GEN-001..004` | Go/TS/client drift and hidden compatibility changes. | Contract-family registry `d8dbaf3f` makes family ownership explicit; authored/generated/drift rows remain. | `BLOCKED` |
| `NFA-RISK-014` | Acceptance criteria lack executable harness primitives/selectors. | Testing Harness/Network Flow | `NFA-TH-001..007`; `NFA-TEST-001..007` | Prose-only or misclassified conformance. | All 107 rows name selectors and retained intended-artifact results. | `BLOCKED` |
| `NFA-RISK-015` | Status/version may flip before all dependencies and bytes close. | All owners | `NFA-ADOPT-001`; `NF-AC-106` | False adoption claim and unresolved behavior. | Coordinated Core 00/NLSpec review proves zero open prerequisites. | `BLOCKED` |
| `NFA-RISK-016` | Large-limit timing may be presented as claim evidence. | Core 05/Network Flow | `NFA-FIX-008`; `NF-AC-050` | Unsupported performance publication. | Keep engineering-only or separately activate and satisfy Core 05. | `TODO` |
| `NFA-RISK-017` | Assumption: current clean `main` snapshot is the intended planning base. | Tracker maintainer | `NFA-INV-001` | Rebase could invalidate checksums and inventories. | Refresh snapshot/checksums before every resumed edit. | `IN_PROGRESS` |

## 14. Append-only workstream notes and session handoff

Do not rewrite earlier log entries to make later state look cleaner. Append a
timestamped correction and cite the superseded entry. Decisions below are
inventory/control decisions only; none resolves product behavior.

### 14.1 Scope and evidence log

- `2026-07-09T23:06:49-04:00` — selected `docs/handoffs/` because it already
  contains the governing framework and resumable trackers. Refreshed clean-main
  snapshot, owner checksums, exact owner-ID counts, adjacent code/contracts, and
  absent Network Flow surfaces. Current commit/checksums are inspection evidence,
  not adopted dependency locators.

### 14.2 Contracts and owner-decision log

- `2026-07-09T23:06:49-04:00` — recorded Core 00 omission/future-only purge as
  the first decision; kept Core 01 discovery/import incompatibilities, Core 02
  IP/purge seams, Core 03 extension UI seams, Core 04 security seams, and the
  cursor ownership split blocked. No shape, token, namespace, or version chosen.
- `2026-07-10T00:16:57-04:00` — `WS-01` owner artifact
  `155b5f64b57a2ee0fd6cb5beb4ff4ea5b26a7d1e` recognizes stable profile identity
  `network_flow_activity` only as unclaimed, preserves five claimable extension
  profiles, reserves final coordinated adoption, and keeps whole-incident removal
  future-only. Core discovery, wire, security, and runtime shapes remain unchosen.
- `2026-07-10T00:30:34-04:00` — `WS-02` owner artifact
  `89580f0c9f8e72c0ebf44da2c09975d6075d2f16` updates Core 01 discovery,
  common job resource refs, import target selection, analytical extension
  preview/apply facades, source-change failure, cross-owner final commit, and
  terminal publication. It reserves `network_flow_activity` only as unclaimed
  until coordinated adoption. It does not generate contracts, implement routes,
  author fixtures, or claim executable conformance.

### 14.3 Core and subsystem workstream log

- `2026-07-09T23:06:49-04:00` — split Core 00–04, Graph Projection, and Testing
  Harness work into owner-sized tasks/checkpoints. Recorded Graph Projection
  retained-only behavior and current 69/36 owner versus 68/23 evidence drift.
- `2026-07-10T00:16:57-04:00` — `docs/domain.md` now classifies Network Flow
  Analysis as a supporting extension context and maps its table, row, workspace,
  and indicator-binding terms without changing the exhaustive Base surface
  registry. Raw external telemetry remains behind an anti-corruption boundary.

### 14.4 Implementation-module log

- `2026-07-09T23:06:49-04:00` — found adjacent extension, import, indicator,
  graph-projection, clock, and public-error-fault seams. Found no Network Flow
  module, route, migration, query, frontend surface, test, or selector. No module
  path was proposed.

### 14.5 Fixtures and transcript log

- `2026-07-09T23:06:49-04:00` — transcribed all 28 full fixture IDs; classified
  16 as single-file SHA-256 inputs and 12 as manifest-backed bundles. Left bytes,
  hashes, manifest schema, mapping paths, and transcript paths `TODO:`. Kept
  `NF-FIX-008` engineering-only and `NF-FIX-022` provenance-blocked.

### 14.6 Tests and harness log

- `2026-07-09T23:06:49-04:00` — mapped all 107 acceptance IDs to intended level,
  fixtures, dependency tasks, and expected artifacts. Every selector is
  `TODO: source not found`; no executable Network Flow result is claimed.

### 14.7 Generated-artifact log

- `2026-07-09T23:06:49-04:00` — observed the five hard-coded contractgen
  families and three generated roots. No Network Flow family or fixture/transcript
  generator owner exists. Generated outputs remain regeneration-only.

### 14.8 Risk and blocker log

- `2026-07-09T23:06:49-04:00` — adoption is blocked on all 12 gates, all 17
  blockers, all seven locators, all 28 fixture freezes, generated ownership,
  and executable evidence. No adoption prerequisite is deferred; only optional
  Core 05 publication activation is `DEFERRED`.
- `2026-07-10T00:03:46-04:00` — correction to the preceding entry: the approved
  execution plan drops only the Network Flow-specific incident-purge slice from
  v1 because Core keeps generic incident removal future-only. `NFA-C02-003` is
  `DROPPED`, not deferred; fixture 027 and AC 104 now prove soft-delete/closure
  retention and no premature purge claim. The 12/17/28/107 accounting is unchanged.

### 14.8.1 Execution-control log

- `2026-07-10T00:03:46-04:00` — began `WS-00` on clean committed tracker base
  `65d7461a`. Added the serial artifact/checkpoint protocol, explicit domain,
  design, migration, Phase 12, and `NFA-IMPL-010..018` workstreams. No owner,
  contract, generated, production, fixture, or test artifact changed in this slice.
- `2026-07-10T00:10:19-04:00` — completed `WS-00` artifact commit
  `1bb6fdbd6ced0fa5835142d205f82f2da75289e7`. `make agent-finalize`, generated
  policy, JSON shape, Markdown lint, table/count/path, fixture-rename, whitespace,
  and one-file-scope checks passed. The next workstream is `WS-01` only.
- `2026-07-10T00:10:19-04:00` — began `WS-01` from clean checkpoint
  `46731b5b`. Only Core 00 bounded-profile ownership and `docs/domain.md`
  vocabulary are active; discovery, wire, storage, and implementation shapes
  remain blocked for their later owner workstreams.
- `2026-07-10T00:23:29-04:00` — committed the `WS-01` checkpoint as
  `58e57ea966ce4ab0d8822bccb2d731b204ea5745`; began `WS-02` from a clean
  worktree. `NFA-C01-001` is the single active Core 01 tracker row. The planned
  artifact is an owner-only Core 01 amendment for discovery, analytical import
  target/result, opaque source/facade, transaction boundary, and terminal
  publication. Generated contracts, implementation, fixtures, and tests remain
  blocked until the owner text is committed and checkpointed.
- `2026-07-10T00:30:34-04:00` — completed the `WS-02` owner artifact
  `89580f0c9f8e72c0ebf44da2c09975d6075d2f16` and validation in §15.3. The next
  safe workstream is the Core 02 owner slice for canonical IP identity and
  indicator transaction participation; no generated, production, fixture, or
  test artifact may start before its controlling tracker checkpoint is committed.
- `2026-07-10T00:34:50-04:00` — committed the `WS-02` checkpoint as
  `537b7068cdfa83f8b8af6313699a2264b8a80197`; began `WS-03` from a clean
  worktree. `NFA-C02-001` is the single active Core 02 tracker row. The planned
  artifact is an owner-only Core 02 amendment for canonical IPv4/IPv6 indicator
  identity and indicator find/create/dedupe transaction participation. Generated
  contracts, production implementation, fixtures, and executable Phase 12
  evidence remain blocked until this owner text is committed and checkpointed.
- `2026-07-10T00:49:02-04:00` — completed the `WS-03` Core 02/indicator owner
  artifact `344486e7cf269f148dd011741434cbbfb7402a23` and validation in §15.4.
  `NFA-C02-001..002` are `DONE`, `NFA-C02-003` remains `DROPPED`, and the next
  safe workstream is `WS-04` / Core 03 claimed extension workspace and
  invalidation semantics. Generated Network Flow contracts, fixtures, C04
  security hooks, C03 browser behavior, and Phase 12 evidence remain blocked.
- `2026-07-10T00:51:38-04:00` — committed the `WS-03` checkpoint as
  `2869c850d820f407dffd9d76d7d768020c8e3e95`; began `WS-04` from a clean
  worktree. `NFA-C03-001` is the single active Core 03 tracker row. The planned
  artifact is a Core 03 owner amendment for claimed-only extension workspace
  identity, with extension-resource invalidation prepared for `NFA-C03-002`.
  Browser implementation, generated WebSocket contracts, fixtures, and Phase 12
  evidence remain blocked until the owner semantics are committed and
  checkpointed.
- `2026-07-10T01:00:29-04:00` — completed the `WS-04` Core 03/Core 01 owner
  artifact `08fa716e50c0978edd02fb3b4637f00f3730bc62` and validation in §15.5.
  `NFA-C03-001..002` are `DONE`. Broader `NF-GATE-005` and `NF-GATE-009` remain
  blocked by generated WebSocket contracts, C04 authorization rules, UI/browser
  implementation, harness auth-transition controls, fixtures, and Phase 12
  evidence. The next safe workstream is `WS-05` / Core 04 security lifecycle.
- `2026-07-10T01:02:44-04:00` — committed the `WS-04` checkpoint as
  `2d997ae9bd0ae1290d1e7b3100b767201a2fe396`; began `WS-05` from a clean
  worktree. `NFA-C04-001` is the single active Core 04 tracker row. The planned
  first artifact is an owner-only Core 04 route-family authorization amendment
  for claimed Network Flow routes. Cursor, safe digest, audit, retention,
  generated contracts, implementation, and fixtures remain blocked until their
  own rows are explicitly activated.
- `2026-07-10T01:07:01-04:00` — completed the `WS-05` route-authorization
  owner artifact `3b942fe02297ce9b0ff548d8984e492e94d878ac` and validation in
  §15.6. `NFA-C04-001` is `DONE`. `NFA-C04-002..005` remain blocked until their
  own rows are activated. Generated contracts, route implementation, membership
  and role fixtures, auth-transition controls, and Phase 12 evidence remain
  later work.
- `2026-07-10T01:08:41-04:00` — committed the `WS-05` route-authorization
  checkpoint as `864ba5caafca564e6d0d305594099f8b669f95b2`; began the Core
  04 cursor-security slice from a clean worktree. `NFA-C04-002` is the single
  active tracker row. The planned artifact is a Core 01/Core 04 owner amendment
  for cursor wire/security ownership, confidentiality, integrity, TTL,
  invalidation, and key rotation. Safe digest, audit, and retention rows remain
  blocked until separately activated.
- `2026-07-10T01:15:12-04:00` — completed the Core 01/Core 04 cursor-security
  owner artifact `90401fb2e85a1730f4096bbd873d8a88f3831184` and validation in
  §15.7. `NFA-C04-002` is `DONE`. The owner contract now fixes the common
  `cursor_token`/`meta.paging.next_cursor` envelope for Network Flow, opaque
  authenticated encryption, key IDs, 15-minute expiry, exact rotation behavior,
  rename survival, and delete/auth invalidation. Safe digest, audit, retention,
  generated contracts, Network Flow implementation, fixtures, and Phase 12
  evidence remain later work.
- `2026-07-10T01:19:31-04:00` — committed the cursor-security checkpoint as
  `9b1fcbab09424bcce4c91c97db4ebfbd59099b39`; began the Core 04 safe-digest
  key lifecycle slice from a clean worktree. `NFA-C04-003` is the single active
  Core 04 security owner row. The planned artifact is an owner-only amendment
  for safe-digest secret namespace, key IDs, rotation/comparison rules, and
  production-versus-fixture secret boundaries. Audit occurrence semantics,
  retention, generated contracts, implementation, and fixtures remain blocked
  until their own rows are activated.
- `2026-07-10T01:25:17-04:00` — completed the Core 04 safe-digest owner
  artifact `cd6457505218f88099c8c2496d0b14088f487c5f` and validation in
  §15.8. `NFA-C04-003` is `DONE`. The owner contract defines the Network Flow
  safe-digest key ring, `safe_digest_key_id`, exact HMAC-SHA-256 boundary,
  comparison-only-within-key-ID rule, rotation epoch semantics, fixture-only
  key prohibition outside harness runtime, and startup failure reason codes.
  Audit occurrence semantics, retention, generated contracts, implementation,
  fixtures, and Phase 12 evidence remain later work.
- `2026-07-10T01:28:23-04:00` — committed the safe-digest checkpoint as
  `785f4a989b601f665796a3160927348ade730816`; began the Core 04 audit
  occurrence slice from a clean worktree. `NFA-C04-004` is the single active
  Core 04 security owner row. The planned artifact is an owner-only amendment
  for transactional domain audit occurrence counts, outbox participation,
  retry/idempotency replay behavior, and no-audit failure boundaries. Retention,
  generated contracts, implementation, fixtures, and harness audit-count helpers
  remain blocked until their own rows are activated.
- `2026-07-10T01:31:56-04:00` — completed the Core 04 audit occurrence owner
  artifact `712585899634f79bde369b5e6d6d8d81d117f918` and validation in §15.9.
  `NFA-C04-004` is `DONE`. The owner contract defines incident-scoped Network
  Flow domain audit occurrences, exact event counts, no-audit replay behavior,
  retry/recovery constraints, safe audit payload boundaries, and graph truncated
  example count semantics. Retention, generated contracts, implementation,
  fixtures, harness audit-count helpers, and Phase 12 evidence remain later
  work.
- `2026-07-10T01:35:07-04:00` — committed the audit occurrence checkpoint as
  `7ba84b02711f5da10cffdc276ea3aa0ee7d49824`; began the Core 04 retention
  slice from a clean worktree. `NFA-C04-005` is the single active Core 04
  security owner row. The planned artifact is an owner-only amendment for
  Network Flow soft-delete retention, raw-source/staging expiry, retained-count
  semantics, cursor/graph invalidation after delete, and the explicit v1
  omission of an incident purge claim. Generated contracts, implementation,
  fixtures, and harness retention/fake-clock evidence remain later work.
- `2026-07-10T01:40:14-04:00` — completed the Core 04 retention owner artifact
  `663c8684616d142851e335ce579894437f899968` and validation in §15.10.
  `NFA-C04-005` is `DONE`. The owner contract defines terminal soft delete,
  non-queryable retained state, active versus retained table counts, cursor and
  ephemeral graph invalidation, Core import-source/staging cleanup ownership,
  no raw-source locator disclosure, exact replay boundaries, and the explicit
  absence of a current whole-incident purge claim. Generated contracts,
  implementation, fixtures, harness retention controls, and NLSpec fixture-name
  cleanup remain later work.
- `2026-07-10T01:43:31-04:00` — committed the retention checkpoint as
  `ea35923d820c9fabb720f9957b9aa605a3cfca88`; began the Graph Projection
  ephemeral invocation slice from a clean worktree. `NFA-GP-001` is the single
  active Graph Projection owner row. The planned artifact is an owner-only
  amendment that adds a non-retained invocation/result boundary without changing
  retained graph-view lifecycle behavior. Exact Network Flow adapter mapping,
  69/36 evidence repair, generated contracts, implementation, and fixtures
  remain later workstreams.
- `2026-07-10T01:50:45-04:00` — completed the Graph Projection ephemeral
  invocation owner artifact `4e446354bb00afe03047be43e26677d74f035bac` and
  validation in §15.11. `NFA-GP-001` is `DONE`. The owner contract defines
  `project_ephemeral`, `ephemeral_projection_result`, `ephemeral_projection_id`,
  admitted ephemeral validation issue identity, direct failure semantics, no
  retained graph-view/run/query/idempotency/cache state, and exact invisibility
  to retained Graph Projection queries. Adapter metadata/property mapping,
  outcome mapping, 69/36 evidence repair, fixtures, implementation, and generated
  contracts remain later workstreams.
- `2026-07-10T01:53:49-04:00` — committed the Graph Projection ephemeral
  checkpoint as `3ca750b534bb40b4eb28a219719d332bb3dce254`; began the exact
  adapter mapping and outcome-contract slice from a clean worktree. `NFA-GP-002`
  is the single active row. The planned artifact is an owner/downstream
  amendment that reconciles Network Flow §14.4 with adopted Graph Projection
  property, metadata, result, and error semantics. Matrix/corpus/checker repair,
  generated contracts, implementation, fixtures, and Phase 12 evidence remain
  later workstreams.
- `2026-07-10T01:55:10-04:00` — validated the `NFA-GP-002` start checkpoint.
  The only workspace change is the tracker start edit. Owner document edits have
  not started; commit this checkpoint before amending Graph Projection or Network
  Flow text.
- `2026-07-10T02:02:31-04:00` — completed the exact Graph Projection adapter
  mapping and outcome artifact `f177fb6b25affeb6b0e7c6ef7846d3fa06f4f6a2` and
  validation in §15.12. `NFA-GP-002` is `DONE`. Graph Projection now explicitly
  rejects private direct-copy/transform members for property and metadata
  mappings; Network Flow §14.4 now invokes `project_ephemeral`, declares six
  exact metadata mapping rules, fixes `dst_port` null emission, tightens ID
  projected types, requires a zero-issue `ephemeral_projection_result` for
  success, and maps all adapter failures without leaking Graph Projection
  internals. Matrix/corpus/checker repair, fixtures, generated contracts,
  implementation, and Phase 12 evidence remain later workstreams.
- `2026-07-10T02:05:10-04:00` — committed the Graph Projection adapter
  checkpoint as `31dd003a434cbd2271eae0d36dff0e7216806ef7`; began the
  Graph Projection matrix/corpus/checker evidence-drift slice from a clean
  worktree. `NFA-GP-003` is the single active row. The planned artifact is an
  authored evidence/checker update that aligns observed Graph Projection owner
  text with 69 `GP-AC-*` rows and 36 `GP-FIX-*` rows. No generated roots are
  edited by hand.
- `2026-07-10T02:06:19-04:00` — validated the `NFA-GP-003` start checkpoint.
  The only workspace change is the tracker start edit. Evidence/checker edits
  have not started; commit this checkpoint before amending authored evidence
  files.
- `2026-07-10T02:12:03-04:00` — completed the Graph Projection evidence
  alignment artifact `81941bbafe47dd66f30b0473e951218c0c9fc919` and validation
  in §15.13. `NFA-GP-003` is `DONE`. The authored conformance matrix now lists
  69 `GP-AC-*` rows and 36 `GP-FIX-*` registry rows, the corpus lists the same
  36 fixture IDs, and `make json-shape-check` now enforces both matrix and corpus
  ranges. Network Flow-specific adapter fixtures and implementation evidence
  remain later workstreams.
- `2026-07-10T02:14:53-04:00` — committed the Graph Projection evidence
  checkpoint as `53c156e6d5452ac7c8660f704e9f6400ca1e0f6c`; began the Testing
  Harness immutable fixture-manifest slice from a clean worktree. `NFA-TH-001`
  is the single active row. The planned artifact is a harness-owner amendment
  for immutable Network Flow fixture manifests and execution mechanics. Fault,
  clock, randomness, auth-transition, audit-count, generated-contract, and
  Phase 12 accounting controls remain later harness slices.
- `2026-07-10T02:16:07-04:00` — validated the `NFA-TH-001` start checkpoint.
  The only workspace change is the tracker start edit. Harness owner edits have
  not started; commit this checkpoint before amending Testing Harness text.
- `2026-07-10T02:43:45-04:00` — completed the Testing Harness final-commit
  and worker-fault artifact `ecd4edd594489a0e8857ffd8e3cf31834a9bfb40`.
  Harness §12 now owns the Network Flow fault-control route, closed boundary
  tokens, one-shot/correlation/reset semantics, and product-owner routing
  guardrails. The implementation adds a guarded in-memory registry, closed
  response schema, schema contract fixture, and route tests for authorization,
  invalid requests, conflict, exact consumption, and reset clearing.
- `2026-07-10T02:47:29-04:00` — committed the Testing Harness fault-control
  checkpoint as `d9080de9ff6451416fc3d5fcb4383d43b262b758`; began the
  fake-clock coverage slice from a clean worktree. `NFA-TH-003` is the single
  active row. The planned artifact is limited to harness-owned clock controls
  and reset evidence for Network Flow time-sensitive fixtures; deterministic
  randomness, auth transitions, audit counts, and drift accounting remain
  separate rows.
- `2026-07-10T02:57:28-04:00` — completed the Testing Harness fake-clock
  artifact `30880992ec1e23a5dddb6a4cfe2de0336bc6c78b`. Harness §12 now owns
  clock set/reset/state routes, the `cartulary.test.clock_control.v1` response,
  runtime-reset clearing for registered clocks, and Network Flow owner-routing
  limits for time-sensitive fixtures. The implementation adds reset/state
  handlers, schema validation, route tests, and reset integration evidence.
- `2026-07-10T02:59:29-04:00` — committed the Testing Harness fake-clock
  checkpoint as `6a19aa6c03341f305d31665aff75c2fb9af4f117`; began the
  deterministic randomness/collision slice from a clean worktree. `NFA-TH-004`
  is the single active row. The planned artifact is limited to fixture-only
  deterministic random streams and collision injection controls; auth
  transitions, audit counts, and drift accounting remain separate rows.
- `2026-07-10T03:11:21-04:00` — committed the Testing Harness deterministic
  randomness artifact as `84eb7fb5d4a8ba0ae0476ce5562126f2355f45f0`.
  The artifact adds the closed `cartulary.test.network_flow_randomness_control.v1`
  schema, the guarded Network Flow randomness route, an in-memory fixture-only
  stream registry, schema tests, route/reset integration tests, and harness
  owner text for fail-closed collision and deterministic ID/nonce fixtures.
- `2026-07-10T03:13:21-04:00` — committed the Testing Harness deterministic
  randomness checkpoint as `ce510906a8363c2e406b4c290c13e441f5662ee4`;
  began the authorization-transition and hidden-resource harness slice from a
  clean worktree. `NFA-TH-005` is the single active row. The planned artifact is
  limited to fixture-only transition controls and owner-routed hidden-resource
  assertions; audit-count helpers and drift accounting remain separate rows.
- `2026-07-10T03:23:11-04:00` — committed the Testing Harness
  authorization-transition artifact as
  `43b0fe36e2c35022c13a4b33e9d226b6303ae379`. The artifact adds the closed
  `cartulary.test.network_flow_auth_transition_control.v1` schema, the guarded
  Network Flow auth-transition route, an in-memory fixture-only transition
  registry keyed by boundary/actor/incident/resource kind/resource ref, schema
  tests, route/reset integration tests, and harness owner text for hidden-resource
  non-disclosure assertions.
- `2026-07-10T03:25:05-04:00` — committed the Testing Harness
  authorization-transition checkpoint as
  `cc11c4f30ee51ad8ba3fa00ca558c0bf0d5fbf97`; began the audit-count and
  no-audit replay harness slice from a clean worktree. `NFA-TH-006` is the
  single active row. The planned artifact is limited to fixture-only exact-count
  and replay-silence assertions; generated/drift accounting remains separate.
- `2026-07-10T03:35:15-04:00` — committed the Testing Harness audit-count
  assertion artifact as `d96f9ad2ee5cf2c4caaa4f46b39eca69d39e2f50`. The
  artifact adds the closed
  `cartulary.test.network_flow_audit_assertion_control.v1` schema, the guarded
  Network Flow audit-assertion route, an in-memory fixture-only assertion
  registry keyed by assertion/event/operation/resource and optional
  correlation, schema tests, route/reset integration tests, and harness owner
  text for exact-count, zero-occurrence, and no-audit replay checks.
- `2026-07-10T03:38:04-04:00` — committed the Testing Harness audit-count
  checkpoint as `d8856caa08c55708edc01fb08b65b54b4928b454`; began the
  `tzdb-2026c` provenance slice from a clean worktree. `NFA-TZ-001` is the
  single active row. `NFA-TH-007`, `NFA-LOC-001`, and `NFA-GEN-001` remain
  blocked until their prerequisites close.
- `2026-07-10T03:46:40-04:00` — committed the `tzdb-2026c` provenance artifact
  as `42338a1d92359fab5f1140d89363e22ebf4be865`. The artifact adds the
  schema-checked
  `contracts/network-flow/timezone/tzdb-2026c.provenance.json` record for the
  IANA data-only archive `tzdata2026c.tar.gz`, including release version,
  release timestamp, source and signature SHA-256 values, OpenPGP issuer
  fingerprint, license digest, owner references, and no-host-timezone
  conformance policy.
- `2026-07-10T03:50:32-04:00` — committed the `tzdb-2026c` provenance
  checkpoint as `0504f532a67ec761402a6bdb2f2cfbc123002398`; began the
  generated-contract ownership slice from a clean worktree. `NFA-GEN-001` is
  the single active row. The planned change is to make contract family ownership
  explicit and manifest-driven while keeping Network Flow generated outputs
  inactive until authored contracts exist.
- `2026-07-10T04:01:51-04:00` — committed the contract-family ownership
  artifact as `d8dbaf3f6ab9fd5365e67846b6592f0ba3a0688a`. The artifact adds
  the `contracts/index.json` family registry, registry schema, registry-driven
  `tools/contractgen` loader, Go validation tests, JSON-shape checker coverage,
  and Harness contract tests. The five existing generated families remain
  active, Network Flow is declared as a planned family, and `make
  generate-drift` proves generated outputs remain byte-stable.
- `2026-07-10T04:05:31-04:00` — committed the contract-generator ownership
  checkpoint as `93c3fc47763f2c3533795ae5d94a41a0030a0817`; began the
  Testing Harness generated-contract, structural-lint, and drift-accounting
  slice from a clean worktree. `NFA-TH-007` is the single active implementation
  row. The planned change is a fail-closed Network Flow structural-accounting
  check behind existing public Make targets, not a private raw command path.
- `2026-07-10T04:15:01-04:00` — committed the Network Flow structural
  accounting harness artifact as `0eaa8db333ae841c376603bc3aadeb0fcfa9e84b`.
  The artifact adds `cartulary.network_flow_activity_accounting.v1`, the
  authored accounting manifest, JSON-shape validation, Harness contract tests,
  and generated-drift scratch input coverage. The checker verifies 28 fixture
  IDs, 107 acceptance IDs, tracker mapping, planned/active contract-family
  state, generated-output marker consistency, required scratch inputs, and
  required public Make target presence.
- `2026-07-10T04:19:05-04:00` — committed the Testing Harness structural
  accounting checkpoint as `74555f60887a78d7a45ca5bab65eed1b01d50f41`;
  began the dependency-locator and soft-delete fixture-name reconciliation slice
  from a clean worktree. `NFA-LOC-001` is the single active row. The planned
  artifact fills the seven Network Flow Table 1-B owner locators without
  performing the final adoption flip, adds structural lint coverage for those
  locators, and reconciles `NF-FIX-027` to the no-private-purge v1 decision.
- `2026-07-10T04:30:07-04:00` — committed the dependency-locator artifact as
  `a5dc9847c32b490d988a74b6330d5017d04809be`. The artifact fills all seven
  Network Flow Table 1-B locators with owner file/section/requirement/schema
  references, reconciles `NF-FIX-027` and `NF-AC-104` to soft-delete,
  incident-closure retention, non-queryability, and no v1 whole-incident purge
  claim, and extends Network Flow structural accounting so `json-shape-check`
  fails on missing dependency rows, locator `TODO:` cells, or missing required
  locator fragments.
- `2026-07-10T04:33:42-04:00` — committed the dependency-locator completion
  checkpoint as `54138e413bec85ca471ca05a4f08f6c1b72daa2d`; began the
  Network Flow authored-contract slice from a clean worktree. `NFA-GEN-002` is
  the single active generated-contract row. The planned artifact is limited to
  authored, owner-derived Network Flow contract inputs under `contracts/network-flow/`;
  generated outputs, Phase 12 maps, implementation, fixtures, and transcripts
  remain blocked until later rows activate them.
- `2026-07-10T05:01:07-04:00` — completed the `NFA-GEN-002` authored-contract
  artifact `9bb6711fa79e964646784d5b924b3965f731f3b2`. The artifact adds
  `contracts/network-flow/index.json`, route, error, and public schema bundles,
  plus JSON-shape and harness-contract checks proving exact route inventory,
  exact error inventory, closed public object schemas, contract-file references,
  and immutable timezone provenance. Generated Go/TypeScript outputs, Phase 12
  maps, implementation, fixtures, and transcripts remain blocked until their
  later rows. The next safe workstream is `NFA-GEN-003`, but only after this
  tracker completion checkpoint is validated and committed.
- `2026-07-10T05:04:12-04:00` — began `NFA-GEN-003` from clean checkpoint
  `fb458479f452f1a43f0ee72daa13859c4e468631`. `NFA-GEN-003` is the single
  active generated-contract row. The planned artifact is limited to activating
  the Network Flow contract family through `contracts/index.json`, running the
  repository generator, and accepting only generator-produced Go/TypeScript
  output changes. Implementation, Phase 12 maps, fixtures, transcripts, and
  final adoption remain blocked.
- `2026-07-10T05:12:20-04:00` — completed the `NFA-GEN-003` generated-contract
  artifact `540921da6f93c4c2997a9f53abb565d4dff117dd`. The artifact activates
  the Network Flow contract family, regenerates Go and TypeScript contract
  artifact indexes through `make generate`, corrects public schema
  `schema_id` constants discovered during generated-output review, and adds
  harness coverage so public `schema_id` constants must match their
  `x_schema_id`. `NFA-GEN-004` is the next safe generated-contract row for the
  final drift checkpoint before implementation rows can start.

### 14.9 Current session handoff

| Field | Value |
| --- | --- |
| Date/time | `2026-07-10T05:12:20-04:00` |
| Branch/commit | `main`; artifact snapshot `540921da6f93c4c2997a9f53abb565d4dff117dd` |
| Dirty-tree state | Tracker-only completion checkpoint edit for `NFA-GEN-003`; generated contract artifact committed |
| Current workflow/task | Network Flow generated-output completion checkpoint; `NFA-GEN-003` is complete and `NFA-GEN-004` is the next safe generated-contract row |
| Completed tasks | `WS-00` artifact/checkpoint `1bb6fdbd`/`46731b5b`; `WS-01` artifact/checkpoint `155b5f64`/`58e57ea`; `WS-02` owner/checkpoint `89580f0c`/`537b7068`; `WS-03` artifact/checkpoint `344486e7`/`2869c850`; `WS-04` owner/checkpoint `08fa716e`/`2d997ae9`; `WS-05` route authorization owner/checkpoint `3b942fe0`/`864ba5ca`; cursor-security owner/checkpoint `90401fb2`/`9b1fcbab`; safe-digest owner/checkpoint `cd645750`/`785f4a98`; audit occurrence owner/checkpoint `71258589`/`7ba84b02`; retention owner/checkpoint `663c8684`/`ea35923d`; Graph Projection ephemeral owner/checkpoint `4e446354`/`3ca750b`; Graph Projection adapter start/owner/checkpoint `252f1235`/`f177fb6b`/`31dd003a`; Graph Projection evidence start/artifact/checkpoint `e09828da`/`81941bba`/`53c156e6`; Testing Harness manifest start/artifact/checkpoint `32b2bec5`/`b3f46bfd`/`a7ce696e`; Testing Harness fault-control start/artifact/checkpoint `0b8dc777`/`ecd4edd5`/`d9080de9`; Testing Harness fake-clock start/artifact/checkpoint `d65308ff`/`30880992`/`6a19aa6c`; Testing Harness deterministic randomness start/artifact/checkpoint `9b2aac15`/`84eb7fb5`/`ce510906`; Testing Harness auth-transition start/artifact/checkpoint `fca79909`/`43b0fe36`/`cc11c4f3`; Testing Harness audit-count start/artifact/checkpoint `a2f60652`/`d96f9ad2`/`d8856caa`; timezone provenance start/artifact/checkpoint `2a466ab1`/`42338a1d`/`0504f532`; contract-generator ownership start/artifact/checkpoint `e7b7a349`/`d8dbaf3f`/`93c3fc47`; Testing Harness structural-accounting start/artifact/checkpoint `2cc44891`/`0eaa8db3`/`74555f60`; dependency-locator start/artifact/checkpoint `2bcc06bc`/`a5dc9847`/`54138e41`; authored-contract start/artifact/checkpoint `8b65d79e`/`9bb6711f`/`fb458479`; generated-contract start/artifact `411421c4`/`540921da` |
| Tracker file changed | `docs/handoffs/network-flow-activity-adoption-handoff-tracker.md` |
| Other changed files | none expected after the artifact commit; tracker completion checkpoint only |
| Commands run | `git status --short --branch`; `git rev-parse HEAD`; `date -Iseconds`; tracker/source `rg`; `sed`; `node --check`; direct `network-flow-contract-index` checker; `make generate`; `make json-shape-check`; `make generated-artifact-policy-check`; `make generate-drift`; `make harness-contract`; `make lint-scripts`; `make backend-unit`; `make frontend-typecheck`; `git diff --check` |
| Passing validation | `NFA-GEN-002` start, artifact, and completion checkpoint validation are recorded in §15.24; `NFA-GEN-003` start and artifact validation are recorded in §15.25 |
| Failing validation | `make json-shape-check` `.cartulary/test-results/20260710T082738Z-p15417` failed on the new locator-fragment check because the manifest expected lowercase `adopted` while the table used `Adopted`; fixed and reran pass. One broad ad hoc multi-file table-column check failed on escaped-pipe examples outside changed tables; targeted tracker/Table 1-B/Table 8-D checks passed. Prior unrelated `make harness-contract` failure `.cartulary/test-results/20260710T081325Z-p2192` remains documented in §15.22. During `NFA-GEN-003`, `make harness-contract` `.cartulary/test-results/20260710T090918Z-p50578` failed because the accounting test still expected active-without-generated-marker failure; the test was updated for the active family state and reran pass. |
| Decisions recorded | `network-flow` is now an active generated contract family; generated roots were changed only by `make generate`; generated-drift enforcement remains a separate checkpoint |
| Open questions | none for starting `NFA-GEN-004`; implementation rows remain blocked until `NFA-GEN-004` is checkpointed |
| Blockers | Phase 12 maps, implementation, immutable fixtures, executable evidence, and final coordinated adoption remain blocked |
| Next recommended task/workflow | Validate and commit this completion checkpoint, then start `NFA-GEN-004` from a clean worktree |
| Safe restart command | `rg -n -e 'NFA-GEN-004' -e 'NetworkFlowArtifacts' -e 'networkFlowArtifacts' -e 'network-flow' docs/handoffs/network-flow-activity-adoption-handoff-tracker.md contracts tools/harness internal/gen packages/protocol-ts` |

## 15. Tracker validation procedure and current accounting

Run from repository root. The first four commands are the required Make-owned
documentation/harness checks:

```text
make agent-finalize
make generated-artifact-policy-check
make json-shape-check
make lint-markdown
```

`RESULTS_DIR` must be unset for this tracker-creation run. Record the explicit
retained-run maintenance skip from `agent-finalize`. Because
`.markdownlint-cli2.jsonc` does not include `docs/handoffs/**/*.md` in its globs,
supplement the passing target with these read-only checks:

```text
git diff --check -- docs/handoffs/network-flow-activity-adoption-handoff-tracker.md
awk '/^\|/ { count = gsub(/\|/, "&"); if (!inside) { expected = count; inside = 1 } if (count != expected) { print FNR ": inconsistent table columns"; bad = 1 } next } { inside = 0 } END { exit bad }' docs/handoffs/network-flow-activity-adoption-handoff-tracker.md
```

The mechanical count/contiguity check used for this file is:

```text
awk -F'`' '/^## 3\./{s=3} /^## 4\./{s=4} /^## 5\./{s=5} /^## 6\./{s=6} /^## 8\./{s=8} /^## 9\./{s=9} /^## 10\./{s=10} s==3 && /^\| (Core 0[0-4]|Graph Projection|Testing Harness)/{dep++} s==5 && /^\| `NF-GATE-/{gate++; seen[$2]++} s==5 && /^\| `NF-BLOCK-/{block++; seen[$2]++} s==8 && /^\| `NF-FIX-/{fix++; seen[$2]++} s==9 && /^\| `NF-AC-/{ac++; seen[$2]++} END { bad=(dep!=7 || gate!=12 || block!=17 || fix!=28 || ac!=107); for(i=1;i<=12;i++){id=sprintf("NF-GATE-%03d",i); if(seen[id]!=1)bad=1} for(i=1;i<=17;i++){id=sprintf("NF-BLOCK-%03d",i); if(seen[id]!=1)bad=1} for(i=1;i<=107;i++){id=sprintf("NF-AC-%03d",i); if(seen[id]!=1)bad=1} printf("dependencies=%d gates=%d blockers=%d fixtures=%d acceptance=%d\n",dep,gate,block,fix,ac); exit bad }' docs/handoffs/network-flow-activity-adoption-handoff-tracker.md
```

After `NFA-LOC-001`, exact fixture identifiers compare directly between the
owner NLSpec and tracker. No registered fixture-ID rename shim remains:

```text
diff <(sed -n '/^## 22\./,/^## 23\./p' docs/network-flow-activity-nlspec.md | awk -F'`' '/^\| `NF-FIX-/{print $2}') <(sed -n '/^## 8\./,/^## 9\./p' docs/handoffs/network-flow-activity-adoption-handoff-tracker.md | awk -F'`' '/^\| `NF-FIX-/{print $2}')
```

Path and scope verification uses observed-file checks plus explicit absence
markers; it does not attempt to resolve API route strings as filesystem paths:

```text
for path in AGENTS.md docs/domain.md docs/network-flow-activity-nlspec.md docs/handoffs/cartulary_modular_refactor_planning_framework.md docs/spec/00_document_set_status_and_precedence.md docs/spec/01_architecture_storage_and_view_contracts.md docs/spec/02_domain_model_schema_and_history.md docs/spec/03_workbook_interaction_collaboration_and_workflows.md docs/spec/04_security_deployment_and_conformance.md docs/graph_projection_nlspec.md docs/testing-harness-nlspec.md contracts/openapi/cartulary.openapi.yaml contracts/ws/index.schema.json contracts/errors/index.json contracts/extensions/index.json contracts/graph-projection/conformance_matrix.v1.json contracts/graph-projection/fixtures/corpus.v1.json tools/contractgen/main.go tools/contractgen/validation.go tools/generated_artifact_policy.json tools/generate_drift_scratch_inputs.json tools/harness/generated-artifacts/check-generate-drift.sh tools/harness/generated-artifacts/check-json-shapes.mjs internal/gen/contracts packages/protocol-ts/src/generated packages/ui-contracts/src/generated; do test -e "$path" || exit 1; done
rg -n 'TODO: source not found' docs/handoffs/network-flow-activity-adoption-handoff-tracker.md
git diff --name-only
git diff -- docs/handoffs/network-flow-activity-adoption-handoff-tracker.md
```

### 15.1 `WS-00` normalization results

| Check | Result |
| --- | --- |
| `make agent-finalize` with `RESULTS_DIR` unset | Pass; generated unchanged, run checks skipped as specified; `.cartulary/test-results/20260710T040943Z-p79223` |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T040943Z-p79194` |
| `make json-shape-check` | Pass against the recorded pre-existing 68/23 GP validator; `.cartulary/test-results/20260710T040943Z-p79200` |
| `make lint-markdown` | Pass; target coverage caveat above remains |
| `git diff --check` | Pass |
| Table-column check | Pass; zero inconsistent tables |
| Count/contiguity and registered fixture-027 rename checks | Pass; 7 dependencies, 12 gates, 17 blockers, 28 fixtures, 107 acceptance rows |
| Existing/absent path review | Pass |
| Diff authority/invention review | Pass; execution decisions are explicitly user-approved and owner work remains blocked |
| One-file scope | Pass; artifact commit `1bb6fdbd` changes only this tracker |
| Broad `make check` and runtime conformance | Skipped by design: this is one documentation-control artifact with no implementation changes and no Network Flow target |

### 15.2 `WS-01` ownership results

| Check | Result |
| --- | --- |
| Core 00/domain owner artifact | Pass; artifact `155b5f64b57a2ee0fd6cb5beb4ff4ea5b26a7d1e` |
| Tracker checkpoint | Pass; checkpoint `58e57ea966ce4ab0d8822bccb2d731b204ea5745` |
| Validation summary | Pass as recorded in §14.9 before this checkpoint; no failing validation was reported |
| Scope | Owner text and domain vocabulary only; discovery, contracts, generated outputs, implementation, fixtures, and tests remained blocked |

### 15.3 `WS-02` Core 01 owner results

| Check | Result |
| --- | --- |
| Core 01 owner artifact | Pass; artifact `89580f0c9f8e72c0ebf44da2c09975d6075d2f16` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T043028Z-p92650` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T043034Z-p92838` |
| `git diff --check` on Core 01 and tracker | Pass |
| Table-column check on Core 01 and tracker | Pass; zero inconsistent tables |
| Targeted Core 00/Core 01/tracker owner-boundary review | Pass; `network_flow_activity`, `network_flow_table`, `import_source_capability_v1`, `source_changed`, and `owner_apply_contract_unavailable` resolve to owner text |
| Generated contracts and drift | Skipped by design; `NFA-GEN-001..004` remain blocked until all owner amendments and locators close |
| Runtime, fixture, and browser conformance | Skipped by design; `NFA-IMPL-*`, `NFA-FIX-*`, and `NFA-TEST-*` remain blocked |

### 15.4 `WS-03` Core 02/indicator owner results

| Check | Result |
| --- | --- |
| Core 02/indicator owner artifact | Pass; artifact `344486e7cf269f148dd011741434cbbfb7402a23` |
| `make format-go` | Pass |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T044538Z-p15040` |
| `make json-shape-check` | Pass after correcting a Phase 9-looking test name; `.cartulary/test-results/20260710T044707Z-p24651` |
| `make backend-unit` | Pass; `.cartulary/test-results/20260710T044714Z-p25013`; 96 tests, 0 failed |
| `make backend-store` | Pass; `.cartulary/test-results/20260710T044738Z-p27550`; 130 tests, 0 failed |
| `git diff --check` | Pass |
| Table-column check on Core 02 and tracker | Pass; zero inconsistent tables |
| Targeted Core 02/indicator owner-boundary review | Pass; `ipv6_addr`, `indicator_find_or_create_participant_v1`, canonical IPv4/IPv6 helpers, and rollback participant evidence resolve to owner text and code |
| Generated contracts and drift | Skipped by design; `NFA-GEN-001..004` remain blocked until all owner amendments and locators close |
| Network Flow binding routes and fault fixtures | Skipped by design; `NFA-IMPL-016`, `NFA-FIX-007`, `NFA-FIX-017`, and `NFA-TH-002` remain later workstreams |

### 15.5 `WS-04` Core 03/Core 01 workspace and invalidation owner results

| Check | Result |
| --- | --- |
| Core 03/Core 01 owner artifact | Pass; artifact `08fa716e50c0978edd02fb3b4637f00f3730bc62` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T045938Z-p45970` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T045938Z-p46002` |
| `git diff --check` | Pass |
| Table-column check on Core 01, Core 03, and tracker | Pass; zero inconsistent tables |
| Targeted Core 03/Core 01 owner-boundary review | Pass; `extension_workspace`, `network_analysis`, `extension_resource_changed`, `renamed`, `soft_deleted`, and `authorization_lost` resolve to owner text |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T050159Z-p48884`; `make json-shape-check` `.cartulary/test-results/20260710T050159Z-p48857`; `git diff --check`; tracker table-column check |
| Generated WebSocket contracts and drift | Skipped by design; `NFA-GEN-001..004` remain blocked until owner amendments, locators, and generator ownership close |
| UI, browser, auth-transition, and fixture conformance | Skipped by design; `NFA-IMPL-017`, `NFA-TH-005`, `NFA-FIX-*`, and `NFA-TEST-006` remain later workstreams |

### 15.6 `WS-05` Core 04 route-authorization owner results

| Check | Result |
| --- | --- |
| Core 04 route-authorization owner artifact | Pass; artifact `3b942fe02297ce9b0ff548d8984e492e94d878ac` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T050626Z-p52911` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T050626Z-p52912` |
| `git diff --check` | Pass |
| Table-column check on Core 04 | Pass; zero inconsistent tables |
| Targeted Core 04 owner-boundary review | Pass; `REQ-04-105A`, `network_flow_activity`, `/api/v1/incidents/{incident_id}/network-flow`, `deployment_admin`, route roles, hidden-resource behavior, and third-party egress prohibition resolve to owner text |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T050803Z-p55540`; `make json-shape-check` `.cartulary/test-results/20260710T050803Z-p55561`; `git diff --check`; tracker table-column check |
| Route implementation and fixtures | Skipped by design; `NFA-IMPL-012`, `NFA-IMPL-015`, `NFA-TH-005`, `NFA-TEST-002`, and generated contracts remain later workstreams |
| Cursor, digest, audit, and retention | Skipped by design at route-authorization checkpoint; `NFA-C04-002..005` required separate activation |

### 15.7 `WS-05` Core 01/Core 04 cursor-security owner results

| Check | Result |
| --- | --- |
| Core 01/Core 04 cursor owner artifact | Pass; artifact `90401fb2e85a1730f4096bbd873d8a88f3831184` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T051420Z-p60082` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T051420Z-p60090` |
| `git diff --check` | Pass |
| Targeted Core 01/Core 04 owner-boundary review | Pass; `REQ-01-559A`, `REQ-04-128..REQ-04-130`, `cursor_token`, `meta.paging.next_cursor`, `cursor_key_id`, `secret_ref_v1`, `network_flow_cursor_invalid`, and the new deployment-config reason codes resolve to owner text |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T051857Z-p67570`; `make json-shape-check` `.cartulary/test-results/20260710T051857Z-p67574`; `make lint-markdown`; `git diff --check` |
| Downstream Network Flow draft alignment | Skipped by design; `docs/network-flow-activity-nlspec.md` still has downstream `next_cursor_token` draft text to reconcile during contract/NLSpec cleanup before generated schemas freeze |
| Cursor implementation and fixtures | Skipped by design; `NFA-IMPL-015`, `NFA-FIX-009`, `NFA-FIX-014`, `NFA-FIX-024`, `NFA-TH-003`, and Phase 12 evidence remain later workstreams |
| Digest, audit, and retention | Skipped by design at cursor checkpoint; `NFA-C04-003..005` required separate activation |

### 15.8 `WS-05` Core 04 safe-digest owner results

| Check | Result |
| --- | --- |
| Core 04 safe-digest owner artifact | Pass; artifact `cd6457505218f88099c8c2496d0b14088f487c5f` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T052448Z-p73520` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T052448Z-p73545` |
| `git diff --check` | Pass |
| Targeted Core 04 owner-boundary review | Pass; `REQ-04-131..REQ-04-134`, `AC-475`, `safe_digest_key_id`, `secret_ref_v1`, `HMAC-SHA-256`, `network_flow_safe_digest_*` reason codes, and `network_flow_fixture_safe_digest_key_forbidden` resolve to owner text |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T052745Z-p78415`; `make json-shape-check` `.cartulary/test-results/20260710T052745Z-p78445`; `make lint-markdown`; `git diff --check` |
| Downstream Network Flow draft alignment | Skipped by design; `docs/network-flow-activity-nlspec.md` still uses the draft term `deployment_audit_secret` and will need cleanup before generated schemas and fixture manifests freeze |
| Safe-digest implementation and fixtures | Skipped by design; `NFA-IMPL-015`, `NFA-FIX-016`, `NFA-FIX-026`, `NFA-TH-004`, and Phase 12 evidence remain later workstreams |
| Audit and retention | Skipped by design at safe-digest checkpoint; `NFA-C04-004..005` required separate activation |

### 15.9 `WS-05` Core 04 audit occurrence owner results

| Check | Result |
| --- | --- |
| Core 04 audit occurrence owner artifact | Pass; artifact `712585899634f79bde369b5e6d6d8d81d117f918` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T053128Z-p83788` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T053128Z-p83789` |
| `git diff --check` | Pass |
| Targeted Core 04 owner-boundary review | Pass; `REQ-04-135..REQ-04-138`, `AC-476`, Network Flow event codes, exact replay silence, graph truncated-example counts, and no-audit failure boundaries resolve to owner text |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T053431Z-p88657`; `make json-shape-check` `.cartulary/test-results/20260710T053431Z-p88668`; `make lint-markdown`; `git diff --check` |
| Audit implementation and fixtures | Skipped by design; `NFA-IMPL-015`, `NFA-FIX-026`, `NFA-TH-006`, and Phase 12 evidence remain later workstreams |
| Retention | Skipped by design; `NFA-C04-005` remains blocked until activated separately |

### 15.10 `WS-05` Core 04 retention owner results

| Check | Result |
| --- | --- |
| Core 04 retention owner artifact | Pass; artifact `663c8684616d142851e335ce579894437f899968` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T053945Z-p94486` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T053945Z-p94504` |
| `git diff --check` | Pass |
| Targeted Core 04 owner-boundary review | Pass; `REQ-04-139..REQ-04-142`, `AC-477`, terminal soft delete, retained/non-queryable state, active/retained counts, cursor and ephemeral graph invalidation, Core import-source cleanup ownership, no raw-source leakage, exact replay, and no current whole-incident purge claim resolve to owner text |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T054236Z-p97371`; `make json-shape-check` `.cartulary/test-results/20260710T054236Z-p97396`; `make lint-markdown`; `git diff --check` |
| Downstream Network Flow draft alignment | Skipped by design; `docs/network-flow-activity-nlspec.md` still has downstream `NF-FIX-027-retention-purge` and incident-purge wording to reconcile during locator/NLSpec cleanup before fixture bytes freeze |
| Retention implementation and fixtures | Skipped by design; `NFA-IMPL-013`, `NFA-IMPL-014`, `NFA-IMPL-015`, `NFA-FIX-009`, `NFA-FIX-027`, `NFA-TH-003`, `NFA-TH-006`, and Phase 12 evidence remain later workstreams |

### 15.11 `WF-04` Graph Projection ephemeral invocation owner results

| Check | Result |
| --- | --- |
| Graph Projection owner artifact | Pass; artifact `4e446354bb00afe03047be43e26677d74f035bac` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T055023Z-p9331` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T055023Z-p9372` |
| `git diff --check` | Pass |
| Graph Projection AC/fixture count guard | Pass; `gp_ac=69 gp_fix=36` |
| Targeted Graph Projection owner-boundary review | Pass; `project_ephemeral`, `ephemeral_projection_result`, `ephemeral_projection_id`, `ephemeral_projection_failed`, non-retained validation issue identity, query invisibility, and `ephemeral_response_only` rejection resolve to owner text |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T055204Z-p11884`; `make json-shape-check` `.cartulary/test-results/20260710T055204Z-p11896`; `make lint-markdown`; `git diff --check` |
| Adapter mapping and evidence repair | Skipped by design; `NFA-GP-002`, `NFA-GP-003`, `NFA-FIX-015`, `NFA-FIX-028`, and Network Flow §14.4 alignment remain later workstreams |

### 15.12 `WF-04` Graph Projection adapter mapping and outcome results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `3ca750b534bb40b4eb28a219719d332bb3dce254`; `NFA-GP-002` is the single active row |
| Start checkpoint validation | Pass; `make lint-markdown`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T055556Z-p19709`; `make json-shape-check` `.cartulary/test-results/20260710T055601Z-p19894`; `git diff --check` |
| Owner artifact | Pass; artifact `f177fb6b25affeb6b0e7c6ef7846d3fa06f4f6a2` |
| `make lint-markdown` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T060136Z-p25029` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T060143Z-p25221` |
| `git diff --check` | Pass |
| Graph Projection AC/fixture count guard | Pass; `gp_ac=69 gp_fix=36` |
| Stale adapter-token guard | Pass; no `direct-copy mode`, no `source_null_behavior='preserve'`, and no invalid Network Flow retention-token wording remains |
| Targeted Graph Projection/Network Flow adapter review | Pass; Graph Projection closes private property/metadata copy members, Network Flow uses `project_ephemeral`, six fixed metadata mappings, `dst_port` `emit_null`, identifier projected types for IDs, zero-issue ephemeral success, closed failure mapping, and no retained lifecycle selectors |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T060423Z-p27771`; `make json-shape-check` `.cartulary/test-results/20260710T060427Z-p27956`; `make lint-markdown`; `git diff --check` |
| Evidence drift repair | Skipped by design for this slice; `NFA-GP-003` remains responsible for matrix, corpus, and checker alignment |

### 15.13 `WF-04` Graph Projection evidence alignment results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `31dd003a434cbd2271eae0d36dff0e7216806ef7`; `NFA-GP-003` is the single active row |
| Start checkpoint validation | Pass; `make lint-markdown`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T060615Z-p31708`; `make json-shape-check` `.cartulary/test-results/20260710T060619Z-p31891`; `git diff --check` |
| Evidence/checker artifact | Pass; artifact `81941bbafe47dd66f30b0473e951218c0c9fc919` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T061106Z-p35273` |
| `make lint-scripts` | Pass |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T061120Z-p36004` |
| JSON parse/count checks | Pass; matrix `ac_count=69 fixture_count=36`; corpus `fixture_count=36` |
| Graph Projection AC/fixture count guard | Pass; owner text `gp_ac=69 gp_fix=36` |
| `git diff --check` | Pass |
| Targeted evidence drift review | Pass; matrix, corpus, and checker enforce the adopted 69 AC and 36 fixture ID ranges; `GP-AC-069` is marked `planned` pending executable operation-admissibility evidence |
| Tracker checkpoint validation | Pass; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T061359Z-p38307`; `make json-shape-check` `.cartulary/test-results/20260710T061403Z-p38492`; `make lint-markdown`; `git diff --check` |

### 15.14 `WF-05` Testing Harness immutable fixture-manifest results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `53c156e6d5452ac7c8660f704e9f6400ca1e0f6c`; `NFA-TH-001` is the single active row |
| Start checkpoint validation | Pass; `make lint-markdown`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T061559Z-p42267`; `make json-shape-check` `.cartulary/test-results/20260710T061607Z-p42458`; `git diff --check` |
| Harness owner artifact | Pass; artifact commit `b3f46bfd3a1ef7c528ee38c8ed6b27f428978aa8` adds `cartulary.network_flow_fixture_manifest.v1`, Harness §§8/11/16/17 owner text, JSON-shape manifest validator, and harness-contract schema/semantic coverage |
| `make lint-markdown` | Pass; `.cartulary/test-results/20260710T062929Z-p56754` |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T062929Z-p56748` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T062914Z-p55542` |
| `make lint-scripts` | Pass; `.cartulary/test-results/20260710T062914Z-p55564` |
| `make harness-contract` | Pass; `.cartulary/test-results/20260710T062914Z-p55527` |
| `git diff --check` | Pass |
| Targeted fixture-manifest review | Pass; schema and checker enforce canonical directory identity, owner refs, sorted source/expected/transcript paths, exact byte hashes, source/expected bundle hashes, symlink/path traversal rejection, unlisted-file rejection, and run-local materialization semantics; later fault/clock/randomness/auth/audit/drift primitives remain `NFA-TH-002..007` |

### 15.15 `WF-06` Testing Harness final-commit and worker-fault results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `a7ce696ea6a9d80a1849d547524d18baa79e716d`; `NFA-TH-002` is the single active row; prerequisites `NFA-TH-001` and `NFA-C01-004` are `DONE` |
| Start checkpoint validation | Pass; `make lint-markdown`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T063312Z-p61561`; `make json-shape-check` `.cartulary/test-results/20260710T063312Z-p61572`; `git diff --check` |
| Harness fault-control artifact | Pass; artifact `ecd4edd594489a0e8857ffd8e3cf31834a9bfb40` adds Harness §12 `POST /api/v1/test/runtime/network-flow-faults`, closed commit/worker boundary tokens, `cartulary.test.network_flow_fault_control.v1`, `NetworkFlowFaultRegistry`, authorization/validation/consume/conflict/reset tests, and schema contract fixtures |
| `make lint-markdown` | Pass; `.cartulary/test-results/20260710T064345Z-p68913` |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T064345Z-p68929` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T064345Z-p68909` |
| `make lint-scripts` | Pass; `.cartulary/test-results/20260710T064426Z-p74036` |
| `make harness-contract` | Pass; `.cartulary/test-results/20260710T064345Z-p68910` |
| `make backend-unit` | Pass; `.cartulary/test-results/20260710T064345Z-p68977`; 96 tests, 0 failed |
| `git diff --check` | Pass |
| Targeted fault-control review | Pass; route is test-only and guarded, response schema is closed, boundary and fault-kind vocabularies match code/schema/spec, only one fault can be pending, exact boundary and optional correlation are required for consumption, consumed faults are one-shot, reset clears registered state through the existing reset hook, and product-owner routing is explicit in TH-HARNESS-REQ-658 |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T064644Z-p75216`; `git diff --check` |

### 15.16 `WF-06` Testing Harness fake-clock results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `d9080de9ff6451416fc3d5fcb4383d43b262b758`; `NFA-TH-003` is the single active row; prerequisite `NFA-TH-001` is `DONE` |
| Source review | Pass; `internal/platform/httpapi/testclock.go` already provides guarded fixed/offset test-clock mutation and `internal/testutil/testruntime/reset_integration_test.go` injects that clock into harness runtimes; reset currently has no explicit clock-reset evidence and Network Flow fixture selector ownership remains pending |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T064852Z-p77501`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T064852Z-p77503`; `make json-shape-check` `.cartulary/test-results/20260710T064852Z-p77542`; `git diff --check` |
| Harness clock-control artifact | Pass; artifact `30880992ec1e23a5dddb6a4cfe2de0336bc6c78b` adds Harness §12 clock set/reset/state ownership, `cartulary.test.clock_control.v1`, `TestClock.Reset`, `TestClock.Snapshot`, guarded reset/state routes, reset hook clearing through test-runtime `ModuleOverrides`, route tests, reset integration evidence, and schema contract fixtures |
| `make lint-markdown` | Pass; `.cartulary/test-results/20260710T065556Z-p81934` |
| `make generated-artifact-policy-check` | Pass; `.cartulary/test-results/20260710T065556Z-p81956` |
| `make json-shape-check` | Pass; `.cartulary/test-results/20260710T065556Z-p81976` |
| `make lint-scripts` | Pass; `.cartulary/test-results/20260710T065556Z-p81996` |
| `make harness-contract` | Pass; `.cartulary/test-results/20260710T065556Z-p82010` |
| `make backend-unit` | Pass; `.cartulary/test-results/20260710T065556Z-p82041`; 96 tests, 0 failed |
| `git diff --check` | Pass |
| Targeted fake-clock review | Pass; clock routes are test-only and guarded, set/reset/state responses validate against a closed schema, fixed/offset/wall modes are explicit, state reads are non-mutating, reset clears fixed and offset state, test-runtime reset clears a registered clock before success, and Network Flow clock evidence is owner-routed through TH-HARNESS-REQ-659 |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T065841Z-p94306`; `git diff --check` |

### 15.17 `WF-06` Testing Harness deterministic-randomness results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `6a19aa6c03341f305d31665aff75c2fb9af4f117`; `NFA-TH-004` is the single active row; prerequisite `NFA-TH-001` is `DONE` |
| Source review | Pass; no harness-owned deterministic random stream or collision-injection route exists; current production/runtime random call sites use UUID generation or `crypto/rand` directly and must not be globally weakened by this fixture-only slice |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T070039Z-p96549`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T070039Z-p96550`; `make json-shape-check` `.cartulary/test-results/20260710T070039Z-p96567`; `git diff --check` |
| Harness deterministic-randomness artifact | Pass; artifact commit `84eb7fb5d4a8ba0ae0476ce5562126f2355f45f0` adds `docs/testing-harness-nlspec.md` owner text, `tools/schemas/cartulary.test.network_flow_randomness_control.v1.schema.json`, schema contract tests, `internal/testutil/testruntime/network_flow_randomness.go`, and route/reset integration tests |
| Artifact validation | Pass; `make backend-unit` `.cartulary/test-results/20260710T070844Z-p1972`; `make backend-integration` `.cartulary/test-results/20260710T070922Z-p4691`; `make lint-markdown` `.cartulary/test-results/20260710T071026Z-p18042`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T071026Z-p18031`; `make json-shape-check` `.cartulary/test-results/20260710T071026Z-p18060`; `make lint-scripts` `.cartulary/test-results/20260710T071026Z-p18096`; `make harness-contract` `.cartulary/test-results/20260710T071026Z-p18125`; `git diff --check` |
| Targeted randomness review | Pass; route registration is test-only and guarded, stream/value-kind vocabularies are closed, deterministic values are never echoed in control responses, duplicate values are preserved for collision fixtures, wrong stream is a no-op, wrong kind and exhaustion fail closed, same-stream rearm is rejected until reset, runtime reset clears registered streams, and no production random or secret source is weakened |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T071238Z-p21622`; `git diff --check` |

### 15.18 `WF-06` Testing Harness auth-transition results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `ce510906a8363c2e406b4c290c13e441f5662ee4`; `NFA-TH-005` is the single active row; prerequisites `NFA-TH-001` and `NFA-C04-001` are `DONE` |
| Source review | Pass; Core 04 route-authorization owner text exists, Core 03 hidden-resource invalidation owner text exists, and ordinary incident membership/session routes exist; no harness-owned Network Flow auth-transition or hidden-resource assertion control exists yet |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T071424Z-p23800`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T071424Z-p23799`; `make json-shape-check` `.cartulary/test-results/20260710T071424Z-p23825`; `git diff --check` |
| Harness auth-transition artifact | Pass; artifact commit `43b0fe36e2c35022c13a4b33e9d226b6303ae379` adds `docs/testing-harness-nlspec.md` owner text, `tools/schemas/cartulary.test.network_flow_auth_transition_control.v1.schema.json`, schema contract tests, `internal/testutil/testruntime/network_flow_auth_transition.go`, and route/reset integration tests |
| Artifact validation | Pass; `make backend-integration` `.cartulary/test-results/20260710T072118Z-p27929`; `make backend-unit` `.cartulary/test-results/20260710T072220Z-p39138`; `make lint-markdown` `.cartulary/test-results/20260710T072220Z-p39066`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T072220Z-p39077`; `make json-shape-check` `.cartulary/test-results/20260710T072220Z-p39099`; `make lint-scripts` `.cartulary/test-results/20260710T072220Z-p39159`; `make harness-contract` `.cartulary/test-results/20260710T072221Z-p39227`; `git diff --check` |
| Targeted auth-transition review | Pass; route registration is test-only and guarded, request validation is closed, response schema omits product membership/session internals and hidden-resource details, controls consume once by exact boundary/actor/incident/resource kind/resource ref plus optional correlation key, duplicate exact tuples conflict without replacement, independent tuples can coexist, runtime reset clears registered transitions, and control semantics are owner-routed rather than product-defining |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T072419Z-p44970`; `git diff --check` |

### 15.19 `WF-06` Testing Harness audit-count results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `cc11c4f30ee51ad8ba3fa00ca558c0bf0d5fbf97`; `NFA-TH-006` is the single active row; prerequisites `NFA-TH-001` and `NFA-C04-004` are `DONE` |
| Source review | Pass; Core 04 audit occurrence owner text exists and defines exact counts/replay silence; no harness-owned Network Flow audit-count or no-audit replay assertion control exists yet |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T072608Z-p47124`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T072608Z-p47134`; `make json-shape-check` `.cartulary/test-results/20260710T072608Z-p47136`; `git diff --check` |
| Harness audit-count artifact | Pass; artifact commit `d96f9ad2ee5cf2c4caaa4f46b39eca69d39e2f50` adds Testing Harness owner text, the closed `cartulary.test.network_flow_audit_assertion_control.v1` schema, contract tests, guarded Go route/registry/reset support, and route/reset integration tests |
| Artifact validation | Pass; `make backend-integration` `.cartulary/test-results/20260710T073222Z-p51008`; `make backend-unit` `.cartulary/test-results/20260710T073325Z-p62184`; `make lint-markdown` `.cartulary/test-results/20260710T073325Z-p62226`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T073325Z-p62270`; `make json-shape-check` `.cartulary/test-results/20260710T073325Z-p62245`; `make lint-scripts` `.cartulary/test-results/20260710T073325Z-p62257`; `make harness-contract` `.cartulary/test-results/20260710T073325Z-p62409`; `git diff --check` |
| Targeted audit-count review | Pass; route registration is test-only and guarded, request validation is closed, response schema omits raw audit payloads and secret material, assertion/event/resource vocabularies are closed, count rules enforce exact-count, zero-occurrence, and no-audit replay semantics, controls consume by exact tuple plus optional correlation key, duplicate exact tuples conflict without replacement, independent tuples can coexist, runtime reset clears registered assertions, and control semantics are owner-routed rather than product audit storage defining |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T073643Z-p68520`; `git diff --check` |

### 15.20 `WF-06` timezone provenance results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `d8856caa08c55708edc01fb08b65b54b4928b454`; `NFA-TZ-001` is the single active row; prerequisite `NFA-AUTH-001` is `DONE` |
| Source review | Pass; Network Flow §9.7 already names `timezone_ruleset_id='tzdb-2026c'`, but no repo-owned immutable source/provenance artifact, license locator, release digest, or timezone validation hook exists yet; host timezone and locale data remain invalid as conformance inputs |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T073915Z-p71162`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T073915Z-p71183`; `make json-shape-check` `.cartulary/test-results/20260710T073915Z-p71175`; `git diff --check` |
| Timezone provenance artifact | Pass; artifact commit `42338a1d92359fab5f1140d89363e22ebf4be865` adds the schema-checked `contracts/network-flow/timezone/tzdb-2026c.provenance.json` record, the closed provenance schema, a `json-shape-check` exact-field validator, Harness contract schema tests, and Network Flow owner text binding v1 timestamp fixtures to the provenance artifact |
| Artifact validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T074547Z-p75171`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T074547Z-p75097`; `make json-shape-check` `.cartulary/test-results/20260710T074547Z-p75082`; `make lint-scripts` `.cartulary/test-results/20260710T074547Z-p75139`; `make harness-contract` `.cartulary/test-results/20260710T074547Z-p75112`; `git diff --check` |
| Targeted timezone review | Pass; IANA data-only archive URL is exact and versioned, source SHA-256 is pinned, detached signature SHA-256 and OpenPGP issuer metadata are recorded, license digest is pinned, `latest` URLs are rejected, host timezone and locale sources are non-authoritative, and fixture transition bytes remain explicitly deferred to `NFA-FIX-022` |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T074833Z-p78801`; `git diff --check` |

### 15.21 `WF-05` generated-contract ownership results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `0504f532a67ec761402a6bdb2f2cfbc123002398`; `NFA-GEN-001` is the single active row; non-final owner amendments for Core 01-04, Graph Projection, Testing Harness primitives, and timezone provenance are complete |
| Source review | Pass; `tools/contractgen/main.go` owns a hard-coded five-family list and no Network Flow family or owner-approved input/output inventory exists; generated roots remain policy-owned and must not be hand-edited |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T075149Z-p81513`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T075149Z-p81530`; `make json-shape-check` `.cartulary/test-results/20260710T075149Z-p81544`; `git diff --check` |
| Contract-family ownership artifact | Pass; artifact commit `d8dbaf3f6ab9fd5365e67846b6592f0ba3a0688a` adds `contracts/index.json`, the closed registry schema, registry-driven contract-family loading, Go validation tests, JSON-shape checker coverage, Harness contract tests, and the schema registry row; Network Flow is declared as a planned family and generated outputs remain unchanged |
| Artifact validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T075824Z-p85733`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T075824Z-p85744`; `make json-shape-check` `.cartulary/test-results/20260710T075824Z-p85768`; `make lint-scripts` `.cartulary/test-results/20260710T075900Z-p92169`; `make harness-contract` `.cartulary/test-results/20260710T075900Z-p92150`; `make backend-unit` `.cartulary/test-results/20260710T075825Z-p85873`; `make generate-drift` `.cartulary/test-results/20260710T075825Z-p85890`; `git diff --check` |
| Targeted generator ownership review | Pass; active families are read from `contracts/index.json`, the Network Flow planned family is not emitted, generated family order remains stable, generated roots were not hand-edited, `make generate-drift` is byte-identical, and planned activation depends on `NFA-GEN-002..004` |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T080327Z-p94581`; tracker table-column check; `git diff --check` |

### 15.22 `WF-06` Testing Harness generated-contract accounting results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `93c3fc47763f2c3533795ae5d94a41a0030a0817`; `NFA-TH-007` is the single active implementation row; prerequisites `NFA-TH-001` and `NFA-GEN-001` are `DONE` |
| Source review | Pass; fixture manifest schemas, contract-family registry validation, and generic generated drift checks exist, but no Network Flow-specific structural accounting check ties planned/generated contract state, fixture/AC row counts, owner routing, and drift closure together behind Make-owned validation |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T080641Z-p97724`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T080641Z-p97722`; `make json-shape-check` `.cartulary/test-results/20260710T080641Z-p97728`; `git diff --check` |
| Structural-accounting artifact | Pass; artifact commit `0eaa8db333ae841c376603bc3aadeb0fcfa9e84b` adds Harness owner text, `cartulary.network_flow_activity_accounting.v1`, `tools/network_flow_activity_accounting.json`, JSON-shape validation, Harness contract tests, generated-drift scratch coverage, and a forward-compatible contract-family registry check |
| Artifact validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T081325Z-p2141`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T081325Z-p2096`; `make json-shape-check` `.cartulary/test-results/20260710T081325Z-p2094`; `make lint-scripts` `.cartulary/test-results/20260710T081325Z-p2105`; `make harness-contract` `.cartulary/test-results/20260710T081354Z-p5062`; `make generate-drift` `.cartulary/test-results/20260710T081354Z-p5270`; `git diff --check` |
| Targeted accounting review | Pass; checker proves 28 contiguous `NF-FIX-*` bases, 107 contiguous `NF-AC-*` rows, tracker row coverage, planned Network Flow contract-family activation dependencies, absence of generated Network Flow markers while planned, generated-drift scratch input coverage, and required public validation target presence |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T081622Z-p7529`; tracker table-column check; `git diff --check` |

### 15.23 `WF-07` dependency locator results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `74555f60887a78d7a45ca5bab65eed1b01d50f41`; `NFA-LOC-001` is the single active row; non-final owner amendments and structural accounting are complete, while the final adoption flip remains deferred to `NFA-ADOPT-001` |
| Source review | Pass; Network Flow Table 1-B still contains seven `TODO:` locator cells, the tracker already records the intended `NF-FIX-027-retention-soft-delete` row, and the owner NLSpec still needs the matching no-private-purge fixture name and acceptance wording |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T082021Z-p10726`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T082021Z-p10729`; `make json-shape-check` `.cartulary/test-results/20260710T082021Z-p10739`; tracker table-column check; `git diff --check` |
| Dependency-locator artifact | Pass; artifact commit `a5dc9847c32b490d988a74b6330d5017d04809be` fills all seven Table 1-B locators, updates Core 02/gate/blocker wording away from private purge semantics, renames `NF-FIX-027-retention-soft-delete`, updates `NF-AC-104`, and extends `cartulary.network_flow_activity_accounting.v1` plus its validator/test coverage for dependency-locator accounting |
| Artifact validation | Pass after one manifest-fragment casing fix; `make lint-markdown` `.cartulary/test-results/20260710T082837Z-p19883`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T082837Z-p19905`; `make json-shape-check` `.cartulary/test-results/20260710T082837Z-p19871`; `make lint-scripts` `.cartulary/test-results/20260710T082837Z-p19922`; `make harness-contract` `.cartulary/test-results/20260710T082837Z-p19939`; `make generate-drift` `.cartulary/test-results/20260710T082837Z-p20013`; `git diff --check` |
| Targeted locator review | Pass; Table 1-B has seven dependency rows and no `TODO:` locator cells, `json-shape-check` enforces required dependency names and locator fragments, the old `NF-FIX-027-retention-purge` slug is absent, direct fixture-ID comparison between NLSpec §22 and tracker §8 passes without a rename shim, and Table 8-D uses retention/no-current-purge language |
| Completion checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T083248Z-p24895`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T083248Z-p24901`; `make json-shape-check` `.cartulary/test-results/20260710T083248Z-p24917`; tracker table-column check; direct fixture-ID diff; targeted Table 1-B check; `git diff --check` |

### 15.24 `WF-08` authored Network Flow contract results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `54138e413bec85ca471ca05a4f08f6c1b72daa2d`; `NFA-GEN-002` is the single active generated-contract row; prerequisites `NFA-GEN-001` and `NFA-LOC-001` are `DONE` |
| Source review | Pass; `contracts/index.json` declares `network-flow` as a planned family, `contracts/network-flow/` contains only timezone provenance, and no authored Network Flow route/error/schema inputs or generated outputs exist yet |
| Start checkpoint validation | Pass; `make lint-markdown` `.cartulary/test-results/20260710T083712Z-p28527`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T083712Z-p28531`; `make json-shape-check` `.cartulary/test-results/20260710T083712Z-p28545`; tracker table-column check; `git diff --check` |
| Authored contract artifact | Pass; artifact `9bb6711f` adds `contracts/network-flow/index.json`, `routes.v1.json`, `errors.v1.json`, `schemas.v1.json`, and `tools/schemas/cartulary.network_flow_contract_index.v1.schema.json`; no generated roots were hand-edited |
| Contract-shape review | Pass; index references route, error, schema, and timezone provenance files; route validator enforces 11 v1 routes and preview/apply facade schemas; error validator enforces 42 Table 21-A error codes and required reason families; public schema bundle exposes 39 index-listed schema IDs and rejects open object schemas |
| Artifact validation | Pass; direct JSON parse; Node syntax checks for `tools/harness/generated-artifacts/check-json-shapes.mjs` and `tools/harness/tests/test-harness-contracts.mjs`; direct `network-flow-contract-index` checker pass; `make json-shape-check` `.cartulary/test-results/20260710T085921Z-p37209`; `make harness-contract`; `make lint-scripts`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T085946Z-p38624`; `make generate-drift` `.cartulary/test-results/20260710T085952Z-p38815`; `git diff --check` |
| Targeted generated-root review | Pass; generated outputs remain blocked for `NFA-GEN-003..004`, `contracts/index.json` still controls family activation, and `make generate-drift` was byte-identical after adding the authored inputs |
| Completion checkpoint validation | Pass; `make lint-markdown`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T090259Z-p42425`; `make json-shape-check` `.cartulary/test-results/20260710T090259Z-p42447`; tracker table-column check; count/contiguity check `dependencies=7 gates=12 blockers=17 fixtures=28 acceptance=107`; `git diff --check` |

### 15.25 `WF-08` generated Network Flow contract results

| Check | Result |
| --- | --- |
| Start snapshot | Pass; clean worktree at `fb458479f452f1a43f0ee72daa13859c4e468631`; `NFA-GEN-003` is the single active generated-contract row; prerequisite `NFA-GEN-002` is `DONE` |
| Source review | Pass; `contracts/index.json` still marks `network-flow` as `planned`, generated output roots contain no Network Flow markers, and `tools/contractgen` emits only active families from the registry |
| Start checkpoint validation | Pass; `make lint-markdown`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T090526Z-p43911`; `make json-shape-check` `.cartulary/test-results/20260710T090527Z-p43938`; tracker table-column check; count/contiguity check `dependencies=7 gates=12 blockers=17 fixtures=28 acceptance=107`; `git diff --check` |
| Generated contract artifact | Pass; artifact `540921da` activates `network-flow` in `contracts/index.json`, updates `internal/gen/contracts/contracts_gen.go`, updates `packages/protocol-ts/src/generated/contracts.ts`, leaves `packages/protocol-ts/src/generated/index.ts` unchanged, and contains Network Flow generated markers |
| Public schema consistency repair | Pass; generation review found nine public result/list schemas whose `schema_id.const` used underscore IDs while `x_schema_id` and route contracts used dotted IDs; `contracts/network-flow/schemas.v1.json` now aligns those constants and `json-shape-check` enforces future consistency |
| Artifact validation | Pass; `make generate` `.cartulary/test-results/20260710T090837Z-p48051`; `make json-shape-check` `.cartulary/test-results/20260710T090901Z-p49044`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T090910Z-p49467`; `make generate-drift` `.cartulary/test-results/20260710T091154Z-p59956`; `make harness-contract` after fixing the active-family test; `make lint-scripts`; `make backend-unit` `.cartulary/test-results/20260710T091038Z-p52932`; `make frontend-typecheck`; `git diff --check` |
| Targeted generated-root review | Pass; generated roots were changed by `make generate`, generated Go contains `NetworkFlowArtifacts` and `NetworkFlowArtifactsIndex`, generated TypeScript contains `networkFlowArtifacts` and `networkFlowArtifactsIndex`, and generated artifact hashes include route/error/schema/index/timezone inputs |
| Completion checkpoint validation | Pass; `make lint-markdown`; `make generated-artifact-policy-check` `.cartulary/test-results/20260710T091438Z-p61879`; `make json-shape-check` `.cartulary/test-results/20260710T091438Z-p61957`; tracker table-column check; count/contiguity check `dependencies=7 gates=12 blockers=17 fixtures=28 acceptance=107`; `git diff --check` |

## 16. Top-level adoption checklist

### Tracker-control checklist

- [x] Authority posture, non-scope, snapshot, sources, commands, and search limits recorded.
- [x] Seven dependencies and every exact gate/blocker ID crosswalked.
- [x] `WF-00` through `WF-11` ordered with prerequisites and stop conditions.
- [x] Core/subsystem seams and generated-artifact ownership mapped.
- [x] All 28 full fixture IDs have independent byte/hash/manifest/transcript rows.
- [x] All 107 acceptance IDs have independent executable-evidence rows.
- [x] Risks, append-only notes, checkpoints, validation accounting, and handoff included.
- [x] `WS-00` validation and diff review recorded in §15.1.
- [x] Normalized tracker artifact committed as `1bb6fdbd` with attributable evidence.

### Adoption-prerequisite checklist

- [ ] Core 00 extension-profile and explicit v1 non-purge decision adopted.
- [x] Core 01 discovery/import/facade/UoW/publication seams adopted.
- [x] Core 02 IPv4/IPv6 identity and create/dedupe participation adopted; private purge slice remains `DROPPED`.
- [x] Core 03 extension workspace and invalidation seams adopted.
- [x] Core 04 authorization/cursor/digest/audit/retention seams adopted.
- [x] Graph Projection ephemeral interface adopted and 69/36 evidence drift closed.
- [x] Testing Harness manifests/clock/randomness/fault/auth/audit/drift capabilities adopted.
- [x] Seven dependency versions and immutable locators resolve.
- [x] `tzdb-2026c` provenance, revision, license, and digest are immutable.
- [ ] All 28 fixture byte/hash/manifest/mapping/transcript rows are frozen.
- [ ] Authored Network Flow contracts, generator ownership, generated outputs, and drift checks pass.
- [ ] All 107 acceptance selectors run against intended artifacts with retained evidence.
- [ ] Full security/fault/retention/drift evidence bundle passes.
- [ ] Core 00 registry and Network Flow status/version transition pass one coordinated final review.

## 17. Tracker-level binary acceptance criteria

| ID | Criterion | Current result |
| --- | --- | --- |
| `NF-HT-AC-001` | The tracker records branch, commit, dirty-tree state, sources, and search limits. | `PASS: WS-00` |
| `NF-HT-AC-002` | Every Table 1-B dependency maps to owner sections, tracker tasks, and a closure condition. | `PASS: WS-00` |
| `NF-HT-AC-003` | Every `NF-GATE-*` and `NF-BLOCK-*` ID appears exactly once in the crosswalk and maps to actionable tasks. | `PASS: WS-00` |
| `NF-HT-AC-004` | Every `NF-FIX-*` ID has its own byte/hash/manifest/transcript/evidence row. | `PASS: WS-00` |
| `NF-HT-AC-005` | Every `NF-AC-*` ID has one executable-evidence row; no generated matrix is claimed. | `PASS: WS-00` |
| `NF-HT-AC-006` | No dependency, fixture, acceptance criterion, or status transition is `DONE` without named evidence. | `PASS: WS-00` |
| `NF-HT-AC-007` | Owner amendments precede derived contracts, generated outputs, implementation, and frozen transcripts. | `PASS: WS-00` |
| `NF-HT-AC-008` | Generated files are identified and no hand edit is planned. | `PASS: WS-00` |
| `NF-HT-AC-009` | Security fixtures use isolated fixture key material and reveal no production secret. | `PASS: WS-00` |
| `NF-HT-AC-010` | Time, randomness, locale, Unicode, timezone, ordering, and line endings freeze wherever they affect bytes/transcripts. | `PASS: WS-00` |
| `NF-HT-AC-011` | Every checkpoint names prerequisites, validation, expected artifacts, and rollback/stop point. | `PASS: WS-00` |
| `NF-HT-AC-012` | Existing repository drift and owner contradictions remain visible and unresolved. | `PASS: WS-00` |
| `NF-HT-AC-013` | Final transition cannot complete with any gate, blocker, fixture, locator, generated check, or AC open. | `PASS: WS-00` |
| `NF-HT-AC-014` | The handoff permits resumption without repeating repository discovery. | `PASS: WS-00` |
| `NF-HT-AC-015` | Only this tracker is newly changed; unrelated dirty state is preserved and recorded. | `PASS: WS-00` |
