# operator-app Module Refactoring Tracker and Handoff

## 1. Scope and Source Posture

| Field | Normative value |
| --- | --- |
| Target path | `internal/app/operator` |
| Target label | `operator-app` |
| Output path | `docs/handoffs/operator-app-module-refactor-tracker.md` |
| Baseline | `main` at `0bd8bd63bd0c19eed8b9f48fc8a9819134667401`; the tracked worktree was initially clean |
| Tracker status | The SL-00 through SL-07 remediation and PC-00 through PC-05 production-hardening iteration are complete. |
| Next-iteration planning baseline | Clean HEAD `846d024b900be4cf8019dcb6cf0c1d938386ff71`; the document update MUST preserve this baseline as the inspected implementation state. |
| Permitted change in this task | The PC-00 through PC-05 write sets in section 13, with this tracker checkpointed after every completed or blocked slice. |
| Prohibited changes in this task | Unrelated product behavior, owner behavior, frontend/browser/network surfaces, data migrations, dependencies, hand edits to generated artifacts, compatibility aliases, and work from a dependent slice before its gate closes. |
| Implementation authority | The 2026-08-06 directive authorized and completed SL-00 through SL-07. The current directive authorizes PC-00 through PC-05; adopted owners remain authoritative over behavior throughout execution. |

`MUST`, `MUST NOT`, `SHOULD`, and `MAY` are normative for execution of this
refactor plan. They do not amend product behavior owned by Core or an adopted
subsystem NLSpec. A proposed owner amendment becomes authoritative only after
the owning document adopts it. Current-state statements describe the inspected
repository and MUST NOT be mistaken for normative owner text.

The document classes used below are closed:

| Class | Meaning | May authorize product behavior? |
| --- | --- | --- |
| Adopted owner | An adopted subsystem NLSpec or the applicable Core owner section | yes, within its named scope |
| Typed projection | A schema, registry, fixture, or generated artifact downstream of an adopted owner | no |
| Current implementation | Live source and test behavior at the recorded baseline | no |
| Refactor requirement | An `OA-REQ-*` rule governing future refactor execution | no product behavior; yes for refactor sequencing and evidence |
| Proposed owner amendment | Exact content that a later owner-authorized task must adopt before dependent implementation | no, until adopted |
| Supporting material | Guides, research, analysis notes, prior trackers, and handoffs | no |

The source hierarchy is:

1. Adopted subsystem NLSpecs, within their named scopes.
2. Core 00 through Core 04 for implementation-conformance behavior.
3. Core 05 only for claim-bearing timed or fixture-sensitive publication. No
   such publication is present in this target.
4. Domain vocabulary and implementation-support guides.
5. Current repository code and tests.
6. Prior plans, handoffs, framework files, research, and analysis notes.

Owner and editorial material inspected:

- `docs/extension-subsystem-nlspec.md`, including `EXT-REQ-236`;
- `docs/spec/00_document_set_status_and_precedence.md`;
- `docs/spec/01_architecture_storage_and_view_contracts.md`, including
  `REQ-01-647` and the Recovery CLI contract;
- `docs/spec/02_domain_model_schema_and_history.md`;
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`;
- `docs/spec/04_security_deployment_and_conformance.md`;
- `docs/testing-harness-nlspec.md`, including `TH-HARNESS-REQ-012`,
  `TH-HARNESS-REQ-667`, and `TH-HARNESS-REQ-673`;
- `docs/domain.md` and `docs/guides/cartulary-dev-guide.md`;
- `docs/handoffs/cartulary_modular_refactor_planning_framework.md`;
- `docs/research/nlspec-spec.md` as an editorial model only;
- `temp/analysis-notes.md` as advisory analysis only.

Repository evidence inspected includes every file under `internal/app/operator`,
`cmd/operator/main.go`, direct assembly/module/platform dependencies, Recovery
and Collaboration process evidence, Recovery contracts and generated projections,
authored test-family and verification-owner inputs, deployment consumers, and
Make-owned command discovery. The non-target files that materially determine
this revision include:

- `internal/modules/extensions/recoverycontribution/contribution.go`;
- `internal/modules/collaboration/recovery.go` and
  `internal/modules/collaboration/stream_integration_test.go`;
- `internal/platform/recoverystate/catalog.go`;
- `contracts/recovery/index.json`,
  `contracts/recovery/recovery-state-catalog.v1.schema.json`, and
  `contracts/recovery/fixtures/recovery-state-catalog.v1.json`;
- `internal/gen/contractrecovery/artifacts_gen.go` and
  `tools/contractgen/recovery_validation.go`;
- the `app.operator`, `module.recovery`, `module.collaboration`,
  `platform.postgres`, and `platform.objectstore` test-family and verification
  owner inputs;
- `tools/test_catalog_owner.json`, `tools/backend_module_boundaries.json`,
  `tools/generated_artifact_policy.json`, and
  `tools/schema_object_ownership_manifest.json`.

### Refactor requirements

| Requirement ID | Normative requirement |
| --- | --- |
| OA-REQ-001 | SL-00 MUST modify only this tracker. Every later slice MUST stay within its section 7 write set and MUST checkpoint this tracker before the next slice starts. |
| OA-REQ-002 | Refactor planning MUST apply the authority hierarchy above. It MUST label conflicting adopted owners `BLOCKED: owner contradiction` and MUST NOT choose a side in implementation. |
| OA-REQ-003 | `cmd/operator` MUST continue to depend exclusively on `internal/app/operator`, and `RunOperatorCLIContext` MUST remain the stable binary-facing entry point unless an adopted owner explicitly authorizes an interface change. |
| OA-REQ-004 | Every file under `internal/app/operator` MUST remain inventoried with its callers, dependencies, tests, contract exposure, target owner, and risk. |
| OA-REQ-005 | Recovery catalog work MUST reconcile exact table identities, owner contributions, classifications, restore actions, and all count fields before it changes catalog composition. Integer equality alone MUST NOT close RB-001. |
| OA-REQ-006 | Collaboration requeue adoption MUST allocate semantics to Core 03, CLI grammar and envelopes to Core 01, local authorization and redaction to Core 04, and cross-document ownership to Core 00. |
| OA-REQ-007 | A behavior-preserving change touching Collaboration requeue MUST preserve the live interface in section 4 until adopted owner text explicitly changes a named behavior. |
| OA-REQ-008 | Collaboration owner adoption MUST close every decision listed in the owner-adoption matrix in section 4 and MUST be backed by the complete characterization matrix. |
| OA-REQ-009 | Harness ownership MUST follow the normative postcondition. Owner changes MUST allocate new owner-qualified identities and a temporary crosswalk; runtime aliases are forbidden. |
| OA-REQ-010 | Characterization MUST precede owner adoption and structural source movement for every behavior not already closed by an adopted owner. |
| OA-REQ-011 | Generated files MUST be changed only through authorized owner inputs and Make-owned generation. No slice may hand-edit a generated artifact. |
| OA-REQ-012 | Work MUST follow the dependency order in sections 6 and 7. A blocked prerequisite MUST prevent every dependent slice from starting. |
| OA-REQ-013 | Validation MUST use the exact Make targets in section 8, record real results and run roots, and distinguish current coverage from required future coverage. |
| OA-REQ-014 | Every session MUST preserve prior handoff history and append current evidence, blockers, changed files, commands, results, and the next authorized action. |
| OA-REQ-015 | A change to command grammar, accepted identifiers, defaults, authorization, mutation semantics, output members, error classification, exit codes, timeout behavior, or exported symbols requires prior owner authorization. |
| OA-REQ-016 | `internal/app/operator` MUST remain a composition and transport facade. It MUST NOT acquire Recovery, Collaboration, Postgres migration-evidence, Object Store, Timeline, Evidence, or generated-catalog semantic ownership. |

## 2. Current-State Repository Inventory

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Required target owner | Risk | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/app/operator/operator.go` | Root CLI facade, dependency defaults, exact registry composition, and binary entry point | `RunOperatorCLIContext` | `cmd/operator/main.go`; target tests; deployment/release scripts through the binary | Four private executors plus shared default constructors | Root target tests and Recovery operator process tests | Generated Recovery inputs through the Recovery executor | `app.operator` facade; semantic owners remain separate | critical | The root runner holds writers and the four command-specific executors; it has no command-specific service field or generic dependency container. |
| `internal/app/operator/operator_collaboration.go` | Strict Collaboration v2 parser, envelope/exit projection, timeout, pool lifecycle, and narrow service delegation | `OperatorCollaborationRequeueResultSchemaID`; `OperatorCollaborationRequeueResult` | Exact registry descriptor and package/process tests | `configassembly`, Collaboration service port, Postgres, UUID | Strict grammar, byte envelope, failure/redaction, timeout, delivery, closure, and process evidence | `contracts/collaboration` v2 schema and closed registry | `app.operator` transport with `module.collaboration` semantics | critical | v1 and parser aliases are removed; transition SQL remains in `module.collaboration`. |
| `internal/app/operator/operator_migration_evidence.go` | Parses migration-evidence arguments, loads source configuration, opens/closes Postgres, delegates evidence construction, and emits JSON | No exported symbol | Operator registry | `configassembly`, Postgres, `migrationevidence`, embedded migrations | Split transport and semantic migration-evidence test symbols | `cartulary.migration_history_evidence.v1`; manifest and SQL inputs | `app.operator` transport; `platform.postgres` semantics | high | Evidence-only; no rewrite authority; SL-04 rows separate each postcondition. |
| `internal/app/operator/operator_migration_evidence_integration_test.go` | Separately exercises operator transport/resource lifecycle and real Postgres migration-evidence semantics | Test surface only | Exact `app.operator` and `platform.postgres` Harness rows | Postgres harness, migration runner, operator wrapper | This file | Migration evidence schema and manifest | Split between `app.operator` and `platform.postgres` rows | high | SL-04 uses two nonoverlapping top-level selectors and new owner-qualified IDs. |
| `internal/app/operator/operator_migration_evidence_test.go` | Separately exercises parser/JSON/redaction/resource transport and migration-evidence projection semantics | Test surface only | Exact `app.operator` and `platform.postgres` Harness rows | Operator adapter, migration-evidence types, fake Postgres | This file | Migration evidence result contract | Split transport and platform semantic test symbols | high | Shared fixtures do not merge the independent normative postconditions. |
| `internal/app/operator/operator_object_store.go` | Object Store init parser, v1 success encoding, typed failure projection, and narrow platform delegation | `OperatorObjectStoreInitResultSchemaID`; `OperatorObjectStoreInitResult` | Exact registry descriptor and package/process tests | `configassembly`, typed config diagnostics, Object Store bucket port | Object Store transport/redaction and durable-effect selectors | Local Object Store v1 result schema literal | `app.operator` transport with `platform.objectstore` semantics | high | Only known typed diagnostics receive specific reasons; every untyped or unknown typed error becomes secret-safe `dependency_unavailable`. |
| `internal/app/operator/operator_recovery.go` | Recovery executor builds source/target dependencies, generated catalogs, projection rebuilders, workbook probe, Evidence hooks, and the Recovery application service | No exported symbol | Operator registry through `recoveryExecutor` | Extension, Recovery, Timeline, Evidence, config, storage, workbook-probe assemblies | Recovery target tests and Recovery process tests | Generated Extension/Recovery catalogs and Recovery contracts | `app.operator` Recovery composition | critical | Injected contributions remain owner-owned; downstream projections and runtime agree on 110/82 with Goose separate. |
| `internal/app/operator/operator_recovery_test.go` | Characterizes Recovery command parsing and failure projection | Test surface only | Harness, partially cataloged | `recoverycli`, Recovery application failures | This file | Recovery result/failure behavior | `app.operator` with `module.recovery` collaborator | critical | Every live test MUST receive exact catalog accounting before source movement. |
| `internal/app/operator/operator_registry.go` | Validates descriptor uniqueness/prefix safety and performs exact routing | No exported symbol | Root runner and registry tests | Standard library and private handlers | Registry test file | Eight-command grammar | `app.operator` | high | Routing, usage, invalid namespace behavior, and exit 2 are frozen. |
| `internal/app/operator/operator_registry_test.go` | Characterizes duplicate/prefix rejection and exact routing | Test surface only | Go test runner; not independently cataloged | Private registry | This file | CLI grammar | `app.operator` | medium | It MUST receive an exact catalog selector before registry restructuring. |
| `internal/app/operator/operator_test.go` | Characterizes Object Store output/redaction, Collaboration arguments, and retired commands | Test surface only | Harness, partially cataloged | Root runner, Object Store, Collaboration | This file | Operator result schema IDs | `app.operator` transport with semantic-owner collaborators | high | Success, failure, cancellation, closure, and exact-byte gaps remain. |
| `internal/app/operator/operator_transport.go` | Shared writer normalization, JSON encoding, error logger, common CLI fields, and closeable Postgres port | No exported symbol | All four private executors and package tests | Standard JSON/I/O/logging/time plus Postgres port | All package adapter tests | None | `app.operator` transport | medium | Shared code is limited to transport mechanics and contains no command-specific policy. |
| `internal/app/operator/recoverycli/cli.go` | Recovery parsing, target-path and timeout validation, JSON/JSONL transport, progress, failure projection, and exit mapping | `Result`, `ArtifactRef`, `Error`, `Progress`, `Command`, `Runner`; parser/timeout/path/failure helpers | Recovery assembly, target tests, Recovery process tests | Recovery application interface and standard library | Recovery CLI, adapter, and process tests | Recovery transport contracts | `app.operator/recoverycli` | critical | It owns operator transport only; exported types and functions remain frozen. |
| `internal/app/operator/recoverycli/cli_test.go` | Characterizes due-verification context behavior | Test surface only | Current `app.operator` row | Recovery CLI/application request types | This file | Recovery command behavior | `app.operator` with `module.recovery` collaborator | high | Exact output and terminal-path coverage remains incomplete. |

All 14 target files are in scope after SL-05: eight production files and six
test files. The three added production files are fully inventoried above.

## 3. Module Boundary Diagnosis

The target is the exact binary-facing application facade and a mixed command-
orchestration package. It is not a domain module. It is not a public HTTP or
WebSocket adapter. It is not a frontend controller or grid integration.

### Authority placement

| Contract portion | Required owner | Tracker treatment | Forbidden substitution |
| --- | --- | --- | --- |
| Recovery catalog and Recovery CLI semantics | Core 01 plus adopted source-owner subsystem NLSpecs | Freeze existing behavior; block contradictory catalog work | Runtime constants, fixtures, or this tracker |
| Collaboration quarantine/requeue transition | Core 03 | Adopted REQ-03-307 and module-owned implementation | `app.operator`, guide prose, or facade SQL |
| Collaboration command grammar, result/error envelopes, output streams, and exit mapping | Core 01 | Adopted REQ-01-655 and v2 typed/runtime projection | Core 03 or implementation-defined behavior |
| Collaboration local-process authorization, no-listener boundary, cancellation safety, and redaction | Core 04 | Adopted REQ-04-151/152 and conformance evidence | Browser/session/admin authorization or raw implementation errors |
| Cross-Core Collaboration ownership allocation | Core 00 §5.1 | Adopted REQ-00-068 separate owner-matrix rows | One ambiguous jointly owned family |
| Harness row ownership and immutable identity | Testing Harness NLSpec | Adopted rule; RB-003 planning decision is closed | Filename, package, runner, binary, or maintainer identity |
| Typed Collaboration CLI projection | Downstream contract chosen by adopted Core owner | MUST follow adopted literal values | A schema invented from current Go structs before adoption |
| Parsing, rendering, resources, and composition | `app.operator` implementation | Retained application-facade responsibility | Mutation policy or domain semantics |
| Requeue mutation | `module.collaboration` implementation | Narrow service port beneath operator executor | SQL in `app.operator` |
| Operational guidance | Development/deployment guides | Examples and runbook only | Product grammar, authorization, schemas, or error registry |

### Responsibility disposition

| Responsibility | Current location | Required owner | Disposition | Evidence | Normative note |
| --- | --- | --- | --- | --- | --- |
| Stable operator entry point | `operator.go` | `app.operator` | keep | `cmd/operator/main.go` | OA-REQ-003 applies. |
| Eight-command routing | Root and registry files | `app.operator` | keep | Live registry | Five Recovery paths plus migration evidence, Object Store init, and Collaboration requeue MUST remain routable. |
| Recovery semantics | Recovery application module | `module.recovery` | keep | Facade/service types | Operator MUST delegate. |
| Recovery transport | `recoverycli` | `app.operator` | keep | Live exported surface | Parser, JSON/JSONL, failure, and exit behavior remain transport concerns. |
| Recovery dependency construction | `operator_recovery.go` | Private Recovery executor in `app.operator` | split after gates | Assembly imports | Timeline, Evidence, catalogs, probes, and storage remain injected contributions. |
| Migration evidence | Migration adapter | `platform.postgres` semantics; operator transport | split after gates | `migrationevidence.Build` | No semantic duplication is permitted. |
| Object Store initialization | `operator_object_store.go` | `platform.objectstore` semantics; operator transport | mechanically split; executor pending | `objectstore.EnsureBucket` | Safe result/error projection remains observable. |
| Collaboration requeue | `operator_collaboration.go` and Collaboration service | Core 03/module semantics; Core 01/app transport; Core 04 local security | complete with narrow executor and typed service port | Live handler/service | Owner, v2 projection, Harness, runtime, and process evidence agree. |
| Broad private runner | `operator.go` | Command-specific private executors | split after gates | Unrelated injected dependencies | No generic command plug-in bus may replace it. |
| Frontend, grid, HTTP/WS, saved views, ordinary entity rows, test-util runtime | Not present | Existing external owners | no action | Direct target/caller search | No workstream may invent these surfaces. |

### Private executor port map

SL-06 implements the source-movement map established in SL-05. Each executor
is package-private and receives only the named ports; the root runner retains
registry composition, writers, and executor selection, not command-specific
services.

| Executor | Required ports | Explicitly excluded |
| --- | --- | --- |
| Recovery | config loader, Postgres opener, Object Store opener, backup-storage constructor, UTC clock, stdout/stderr transport | Collaboration service, migration-evidence builder, Object Store bucket initializer |
| Migration evidence | config loader, Postgres opener, UTC clock, shared JSON/error transport | Recovery catalogs/storage, Collaboration service, Object Store opener/initializer |
| Object Store | config loader, bucket initializer, shared JSON/error transport | Postgres, Recovery, Collaboration, Object Store durable implementation details beyond the platform port |
| Collaboration | config loader, Postgres opener, UTC clock/operation-ID source, shared JSON/error transport | Recovery, Object Store, migration-evidence builder, raw SQL in `internal/app/operator` |

No generic dependency container, plug-in registry, service locator, or command
bus is authorized. Collaboration transition logic remains behind its module
service; the facade may parse and project only the adopted local CLI contract.

## 4. Public Contract and Behavior Freeze Map

### Cross-command freeze

| Contract | Current owner | Frozen behavior | Test posture | Risk |
| --- | --- | --- | --- | --- |
| Binary facade | `app.operator` | `cmd/operator` imports only `internal/app/operator` and invokes `RunOperatorCLIContext` | Boundary and operator-build gates pass | critical |
| Command registry | `app.operator` | Exactly eight command paths; unknown global command prints usage and exits 2; only Collaboration's usage line changed | Exact registry/usage selector passes | critical |
| Recovery commands | Core 01/Recovery plus operator transport | Exact five logical paths, Recovery JSON/JSONL, local authorization, exits 0/2/3/4, no browser/HTTP/WS surface | Focused, service-backed, catalog-negative, and process evidence passes | critical |
| Migration evidence | Postgres semantics plus operator transport | Evidence-only capture, manifest/source audit, safe output, resource closure | Transport and semantic rows are split by owner and pass | high |
| Object Store init | Object Store semantics plus operator transport | Exact command and success schema retained; only known typed failures receive specific safe reasons | Transport, redaction, typed/untyped, and durable-effect rows pass | high |
| Collaboration requeue | Core 01/Core 03/Core 04 through app/module implementations | Strict v2 transport and atomic repaired-state transaction below | Grammar, exact bytes, process, concurrency, rollback, and closure evidence passes | critical |
| Generated contracts/catalogs | Adopted owners and generators | Generated outputs are downstream and tool-owned; Recovery is 110/82 with Goose separate; Collaboration is v2-only | Shape, generation, drift, and generated policy pass | critical |
| Harness accounting | Testing Harness | Normative postcondition ownership and immutable owner-qualified IDs | Owner migration, exact selectors, and Harness contract pass | high |

### Adopted and implemented Collaboration CLI interface

Core 01 REQ-01-655 owns this interface. The table records the matching SL-06
implementation and executable projection; it does not create independent
requirements.

| Interface element | Exact current behavior | Default or omission behavior | Owner-adoption status |
| --- | --- | --- | --- |
| Logical command | `operator collaboration requeue --incident-id <canonical-uuid> [--config <absolute-path>] [--timeout-seconds <seconds>]` | Positional arguments and `--` are rejected | Adopted and implemented. |
| Flag spellings | Only double-dash long names; `--name value` and `--name=value` in any order | Duplicate, single-dash, unknown, missing/empty-value, and positional forms reject through the closed envelope | Adopted and implemented. |
| Help | `--help` is accepted only as the sole command argument | Usage to stderr, exit 0, and no operation envelope | Adopted and exact-byte tested. |
| Unknown/missing flags | Closed `invalid_operator_request` reason mapping | One v2 failure envelope plus LF, ordinary stderr empty, exit 2 | Adopted and grammar-matrix tested. |
| Incident identifier | Exactly non-zero lowercase hyphenated UUID text | Missing produces `missing_required_flag`; every other invalid lexical/value form produces `invalid_flag_value` | Adopted and tested against compact, URN, braced, uppercase, and zero forms. |
| Incident output form | Canonical lowercase hyphenated UUID | Null only when a valid incident identifier was not admitted | Adopted and schema tested. |
| Config flag | Explicit value must be a literal absolute path without NUL, `~`, shell-variable syntax, or lexical `.`/`..` segments | Omission delegates to `CARTULARY_CONFIG_FILE`, then `/etc/cartulary/config.toml` | Adopted and parser/config-tested. |
| Timeout | Inclusive `1..300` seconds and applied to Postgres setup plus the transaction | Default 30 seconds; caller cancellation remains distinct | Adopted and bounded-context tested. |
| Result output | Every matched non-help invocation attempts exactly one v2 JSON object plus LF on stdout; ordinary stderr is empty | Success and failure use the same closed top-level member set | Exact bytes/order/LF and process output pass. |
| Runtime failures | Closed typed code/reason/exit projection with fixed secret-safe messages | Rejected actions exit 3; dependency/transaction/timeout/cancellation failures exit 4 | Every closed family, redaction, and pool-closure case passes. |
| Delivery failure | A stdout failure after commit exits 4 and emits only operation ID plus `result_delivery_failed` to stderr | No rollback is claimed | Adopted and writer-failure tested. |
| Resource closure | Every acquired Postgres pool closes on every terminal path | No close is attempted before successful setup | Adopted and tested. |

The implemented v2 object is closed and ordered:

| JSON member | Type | Exact value rule | Required |
| --- | --- | --- | --- |
| `schema_id` | string | `cartulary.operator.collaboration_requeue_result.v2` | yes |
| `operation_id` | string | Canonical non-zero UUID | yes |
| `operation` | string | `collaboration_requeue` | yes |
| `result` | string | `succeeded` or `failed` | yes |
| `started_at` | string | UTC RFC3339 timestamp | yes |
| `completed_at` | string | UTC RFC3339 timestamp not before `started_at` | yes |
| `incident_id` | string or null | Canonical UUID after identifier admission | yes |
| `requeued_intent_count` | integer or null | Exact nonnegative count on success; null on failure | yes |
| `error` | object or null | Null on success; exact `code`, `reason_code`, `message` on failure | yes |

The encoder emits only those members in that order and appends LF. There is no
v1 decoder, parser alias, dual-output mode, warning member, or unknown member.

### Adopted and implemented Collaboration mutation mapping

| Step | Current implementation effect | Atomicity/failure behavior | Required owner decision |
| --- | --- | --- | --- |
| Admission | Require non-zero operation/incident IDs and one UTC mutation timestamp | Invalid internal construction fails before mutation | Typed request boundary. |
| Cursor lock | Lock only a cursor whose `quarantined_at` is non-null | Missing and non-quarantined states collapse to `incident_not_quarantined` | Single locked admission point. |
| Pending-intent lock and repair proof | Lock all pending intents in deterministic `intent_key` order and validate each canonical payload | Any invalid payload returns `repair_not_verified` before mutation | Fail-closed repair proof. |
| Cursor reset | Clear only `failure_count`, `quarantined_at`, and `quarantine_reason`; update its timestamp | Any failure rolls back the transaction | Exact owner-declared state change. |
| Pending-intent reset | Reset only `attempt_count`, `next_attempt_at`, `last_error_code`, and `updated_at` | Payload, intent/event identity, dispatch state, and sequencing identity remain unchanged | Exact affected-row count must match the locked set. |
| Private journal | Append one raw non-public operator event with operation/incident attribution, prior quarantine summary, safe final summary, and exact count | Journal failure rolls back state and journal together | No public projection or incident revision is added. |
| Commit | Commit once after every state and journal effect | Commit failure maps to `commit_outcome_unknown`; no rollback claim is delivered | Operator guidance requires state/journal inspection before retry. |
| Concurrency and repeat | The quarantined cursor lock admits one winner | Concurrent loser and every post-success invocation return `incident_not_quarantined`; no duplicate journal/effect | Single-winner and deliberately non-idempotent. |
| Forbidden effects | No event selector, deletion, public action, revision, HTTP/WS/browser/common-job, or canonical payload rewrite | Process and integration evidence prove the bounded local surface | Core 03/Core 04 negative boundaries preserved. |

### Collaboration owner-adoption matrix

RB-002's owner-adoption gate closed in SL-02, projection completed in SL-03,
and the matching runtime behavior completed in SL-06.

| Decision ID | Owner | Decision that MUST be adopted | Current evidence | Gate |
| --- | --- | --- | --- | --- |
| CRQ-001 | Core 00 | Separate owner-matrix rows for transition, CLI, and local security | REQ-00-068 | Adopted and implemented |
| CRQ-002 | Core 03 | Incident existence, exact quarantined state, repaired-state precondition, and non-idempotent rejection | Guide plus current SQL | Characterization cases 1, 2, 8, 14 |
| CRQ-003 | Core 03 | Exact cursor and intent effects, failed-event preservation, atomic rollback, and forbidden effects | Current SQL and guide | Characterization cases 1, 9-11, 18 |
| CRQ-004 | Core 03 | Concurrent-attempt ordering, stale/non-quarantined result, and duplicate-effect prohibition | Conditional update only | Characterization cases 13-14 |
| CRQ-005 | Core 03 | Exact system attribution and revision/audit posture | One private raw operator journal in the transaction | Adopted and integration-tested |
| CRQ-006 | Core 01 | Literal token order, flag aliases/forms, duplicates, help, extra args, UUID forms, config flag, and canonical output ID | Live parser | Characterization cases 3-5, 16 |
| CRQ-007 | Core 01 | Success schema ID, complete closed member set, ordering requirement, stdout/stderr policy, and versioning | Live result struct/encoder | Exact-byte snapshot |
| CRQ-008 | Core 01 | Closed failure/reason registry and exit mapping for invocation, config, not found, not quarantined, storage, transaction, encoding, cancellation, and timeout/no-timeout | Current exits are 0/1/2 and failures are untyped | Negative characterization plus owner adoption |
| CRQ-009 | Core 04 | Local OS/config/secret authority; no listener, browser, session, CSRF, bearer, admin, common-job, or WebSocket authority | Recovery pattern and local binary | Positive and negative security conformance |
| CRQ-010 | Core 04 | Secret-safe stdout, stderr, logs, cancellation, resource closure, and no partial success | Raw current logging is not a closed safe contract | Redaction/closure matrix |
| CRQ-011 | Adopted contract projection | Closed v2 schema and registries matching Core literals; no v1 compatibility reader | Core 01 REQ-01-655 plus `contracts/collaboration` | Projection/runtime complete |

Required characterization cases are closed and numbered:

1. quarantined incident success;
2. non-quarantined incident;
3. malformed, zero, noncanonical, and canonical UUID forms;
4. missing UUID, unknown flag, duplicate flags, help, and extra arguments;
5. config omission, environment override, explicit path, and config failure;
6. Postgres settings failure and pool-open failure;
7. exact success stdout bytes, LF, empty stderr, and exit 0;
8. incident not found;
9. storage read/update failure;
10. failure after quarantine update and before intent update completion;
11. commit failure and indeterminate commit posture;
12. caller cancellation before setup, during transaction, and before encoding;
13. two concurrent attempts against one quarantine instance;
14. second invocation after a committed success;
15. pool and transaction closure on every terminal path;
16. exact stderr and exit for every invocation/runtime failure family;
17. redaction of DSNs, credentials, hosts, SQL, event payloads, record content,
    object keys, stack traces, and storage constraint names;
18. proof that the failed event is neither skipped nor deleted.

SL-06 replaced the transitional current-state assertions with the following
adopted v2 and transactional conformance selectors. Historical SL-01 evidence
remains in the session log and retained roots only.

| Cases | Exact selector or selector set | Adopted result |
| --- | --- | --- |
| 1, 18 | `TestDurableIncidentStream_Integration/deterministic_payload_failures_quarantine_only_their_incident_and_requeue_explicitly`; `.../requeue_is_single_winner,_attributed,_and_preserves_pending_event_identity`; `TestOperatorCollaborationRequeueV2_Process` | Successful requeue resets only declared retry/quarantine fields while preserving payload, pending dispatch state, and event identity. |
| 2, 8 | `TestDurableIncidentStream_Integration/requeue_preconditions_collapse_missing_and_non-quarantined_incidents` | Both states collapse to `incident_not_quarantined`. |
| 3, 4 | `TestOperatorCollaborationRequeueArgs_U_StrictCanonicalGrammar`; `TestOperatorCollaborationRequeueArgs_U_RejectsClosedGrammarMatrix` | Only the closed long-flag grammar and canonical non-zero UUID are admitted. |
| 5 | The two argument selectors above; `TestOperatorCollaborationRequeueCommand_U_V2DeliveryAndClosure/configuration_and_Postgres_failures_are_secret_safe` | Config omission delegates to standard discovery; explicit paths are lexical-safe; config failures are typed and redacted. |
| 6 | `TestOperatorCollaborationRequeueCommand_U_V2DeliveryAndClosure/configuration_and_Postgres_failures_are_secret_safe` | Postgres setup failure is secret-safe `postgres_unavailable`; no pool is closed before acquisition. |
| 7 | `TestOperatorCollaborationRequeueCommand_U_V2DeliveryAndClosure/exact_success_envelope_and_acquired_pool_closure`; `TestOperatorCollaborationRequeueV2_Process` | Exact v2 order/bytes/LF, empty stderr, exit 0, count, request attribution, and pool closure pass. |
| 9, 10 | `TestDurableIncidentStream_Integration/requeue_rolls_back_cursor_release_when_intent_reset_fails`; `.../requeue_journal_failure_rolls_back_cursor_and_intent_mutation` | Storage or journal failure leaves cursor, intent, and journal state atomic. |
| 11 | `TestDurableIncidentStream_Integration/requeue_commit_failure_reports_unknown_outcome_and_rolls_back_the_uncommitted_transaction`; v2 failure-mapping subtest `commit_unknown` | The service returns typed outcome-unknown and the transport emits its closed code/reason/exit without a rollback claim. |
| 12 | `TestOperatorCollaborationRequeueCommand_U_V2DeliveryAndClosure/caller_cancellation_and_timeout_remain_distinct`; `.../timeout_flag_bounds_Postgres_and_transactional_work`; `.../stdout_failure_after_success_reports_only_the_delivery_exception`; `TestDurableIncidentStream_Integration/requeue_honors_a_cancelled_caller_before_transaction_admission` | Caller cancellation, operation timeout, and post-commit delivery failure remain distinct and resource-safe. |
| 13, 14 | `TestDurableIncidentStream_Integration/requeue_is_single_winner,_attributed,_and_preserves_pending_event_identity` | Exactly one concurrent attempt commits and journals; the loser and every later invocation reject. |
| 15 | `TestOperatorCollaborationRequeueCommand_U_V2DeliveryAndClosure`; the durable requeue subtests above | Every admitted terminal path has explicit pool/transaction closure or rollback evidence. |
| 16 | `TestOperatorCollaborationRequeueArgs_U_RejectsClosedGrammarMatrix`; `TestOperatorCollaborationRequeueCommand_U_V2DeliveryAndClosure` | Help and every closed invocation/runtime family have exact stderr/envelope and exit evidence. |
| 17 | `TestOperatorCollaborationRequeueCommand_U_V2DeliveryAndClosure/configuration_and_Postgres_failures_are_secret_safe`; Object Store typed/untyped redaction selectors | Injected forbidden values never reach stdout or stderr. |
| 18 | `TestDurableIncidentStream_Integration/requeue_rejects_an_unrepaired_pending_payload_without_changing_state`; `.../requeue_is_single_winner,_attributed,_and_preserves_pending_event_identity` | Invalid payloads reject before mutation, and successful requeue never skips, deletes, or rewrites the pending event. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Required owner | Required action |
| --- | --- | --- | --- | --- | --- |
| The exact facade remains valid with four narrow private executors | Binary root and runner | Future broad-field regression | resolved/guard | `app.operator` | Preserve executor-specific ports and boundary tests; do not add a generic container. |
| Recovery semantics already sit behind a Recovery application facade | `recoverycli` and Recovery service | Ownership inversion if moved outward | intentional/no_action | `module.recovery` | Preserve delegation. |
| Timeline, Evidence, generated catalogs, workbook probes, and storage are injected contributions | Recovery assembly | Catch-all ownership if mislabeled | intentional/no_action | Named source owners | Preserve injection and resource lifetimes. |
| Platform imports occur at the application composition edge | Target imports and boundary policy | Broad private seams | intentional/no_action | `app.operator` composition | Narrow private executors without moving semantic policy. |
| Core 01, Extensions, projections, fixture, validator, and runtime now agree on 110 authored tables including `extension_migration_ledger` | `REQ-01-647`, `EXT-REQ-236`, contracts, exact negative tests | Future drift could omit owned state | resolved/guard | Recovery/Core owner | Preserve exact-set and synthetic-Goose admission tests. |
| `goose_db_version` is not in catalog `tables[]`; runtime database coverage adds it separately | Fixture and `ValidateDatabaseTableNames` | Incorrect synthetic classification proposal | intentional/no_action | Recovery/Core owner | Do not add `entry_origin` or move Goose into `tables[]` unless an adopted owner explicitly changes the model and versions compatibility. |
| Collaboration runtime, guide, contracts, and adopted owners agree on v2 and the repaired-state transaction | Core 00 REQ-00-068, Core 01 REQ-01-655, Core 03 REQ-03-307, Core 04 REQ-04-151 | Future projection/runtime drift | resolved/guard | Core 00/01/03/04 owners | Preserve v2-only generation, strict transport, and module-owned transaction evidence. |
| Collaboration runtime uses closed fixed diagnostics and a single delivery-exception line | Handler and transport tests | Future raw-error disclosure | resolved/guard | Core 01/Core 04 | Preserve forbidden-value and exact stderr tests. |
| Object Store error projection uses only typed configuration and known typed adapter failures | Sanitizer and tests | Future message-based reclassification | resolved/guard | Object Store plus Core/CLI owner | Unknown typed and all untyped errors remain generic. |
| Migration and Object Store rows are allocated by normative postcondition | Split test symbols and family manifests | Future owner drift | resolved/guard | Testing Harness with platform owners | Preserve immutable IDs and exact selector uniqueness. |
| Every target test has exact owner-qualified catalog selection | Harness contract and five owner slices | Future lost evidence | resolved/guard | `app.operator` and semantic owners | Preserve exact, nonoverlapping selectors and immutable row identities. |
| No direct SQL exists in operator | Collaboration handler delegates to service | SQL leakage if refactor moves it | intentional/no_action | `module.collaboration` | Retain service boundary. |
| No frontend, grid, HTTP/WS, saved-view, entity-row, or test-util runtime surface exists in target | Direct search | Scope invention | intentional/no_action | Existing owners | No such work is authorized. |

### RB-001 verified reconciliation state

| Evidence source | Authored catalog count | `authoritative_required` | `extension_migration_ledger` | `goose_db_version` | Authority class |
| --- | --- | --- | --- | --- | --- |
| Core 01 `REQ-01-647` | 110 | 82 | Included through the complete authored owner set | Described as synthetic exclusion | adopted owner |
| Extensions `EXT-REQ-236` | No global count | Requires the Extensions table as authoritative | required | not owned | adopted subsystem owner |
| Recovery schema/index/generated projection | 110 | Registry limit 82 | Schema validates the complete fixture | not in `tables[]` | downstream projection |
| Canonical Recovery fixture | 110 | 82 | present once, authoritative/restore | absent | downstream fixture |
| Generator validator | requires 110 and exact equality with authored migration `CREATE TABLE` identities | requires 82 | included through authored SQL | excluded from authored set | implementation support |
| Runtime catalog | `AuthoredTableCount=110` | `RequiredTableCount=82` | admitted through contribution | appended only for database coverage | implementation |
| Database coverage expectation | 110 catalog identities plus one synthetic identity | not a separate required-count change | present | separate synthetic `goose_db_version` | implementation |

Core 01 REQ-01-647 now adopts 110 authored catalog entries, exactly 82
`authoritative_required`, plus one separate synthetic Goose table used for
database coverage. The owner contradiction is closed. Downstream registry,
schema, generated projection, and negative-fixture repair completed in SL-03.

RB-001 closure includes negative fixtures for extra, missing, duplicate,
misowned, misclassified, wrong-action, wrong-codec, contribution/catalog
mismatch, missing Goose, and unexpected second synthetic identities. Admission
fail before backup publication or restore mutation.

## 6. Refactor Workstreams

| Workflow ID | Name | Class | Required previous workflows | Required subsequent workflows | Goal | Required output | Validation/handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Source and tracker bootstrap | root | None | WF-01 | Freeze baseline, authority, write scope, and prior evidence | Current tracker and clean scope audit | OA-AC-001, OA-AC-002 |
| WF-01 | Exact inventory and reconciliation | chain | WF-00 | WF-03, WF-04 | Reconcile Recovery identities/counts and capture the complete live Collaboration interface | Identity report, interface report, complete inventory | OA-AC-003 through OA-AC-006 |
| WF-03 | Characterization test completion | parallel | WF-01 | WF-02, WF-07 | Turn every unowned observable behavior into exact current evidence | Numbered Collaboration matrix plus command/resource tests | OA-AC-009, OA-AC-010 |
| WF-04 | Boundary and coupling scan | parallel | WF-01 | WF-02, WF-05 | Confirm the thin-facade boundary and reject semantic leakage | Updated boundary findings and import evidence | OA-AC-007, OA-AC-016 |
| WF-02 | Owner adoption and typed projection plan | chain | WF-03, WF-04 | WF-07 | Resolve RB-001 owner contradiction and adopt CRQ-001 through CRQ-011 before product change | Adopted owner amendments and authorized typed-input plan | OA-AC-005, OA-AC-008 |
| WF-07 | Harness ownership migration | chain | WF-02, WF-03 | WF-05 | Split mixed tests, allocate correct owners and identities, regenerate accounting | Authored inputs, temporary crosswalk, generated accounting | OA-AC-011, OA-AC-012 |
| WF-05 | Private facade/executor redesign | chain | WF-02, WF-03, WF-04, WF-07 | WF-06 | Design command-specific private executors with stable public facade | Approved source-movement map and ports | OA-AC-013 |
| WF-06 | Adopted behavior and structural execution | chain | WF-05 | WF-08 | Split files, narrow ports, implement the adopted Collaboration v2 behavior, isolate Recovery assembly, and remove obsolete helpers | Authorized implementation slices with per-slice evidence | OA-AC-013 through OA-AC-015 |
| WF-08 | Validation and final handoff | chain | WF-06 | None | Run exact focused, service-backed, generated, build, and broader gates | Run roots, diff, residual blockers, final handoff | OA-AC-015, OA-AC-016 |

WF-03 and WF-04 MAY run concurrently. No other dependency may be reordered.

## 7. Authorized Remediation Slice Record

| Slice | Status | Depends on | Exact intended change | Authorized files/packages | Contract risk | Required tests/evidence | Validation | Rollback | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| SL-00 | DONE | WF-01 | Revalidate Recovery identity/classification reconciliation, Collaboration interface/exit reports, baseline, and boundaries | No product file change; tracker/evidence only | None | Exact set diff and exact CLI observations | Read-only inspection; boundary check; Markdown and diff integrity | Discard report; no runtime state changes | Reports name every identity, default, output, exit, and unresolved owner decision |
| SL-01 | DONE | SL-00, WF-03 authority | Added exact characterization tests for registry, Collaboration cases 1-18, Object Store failures, migration transport/semantics, closure, cancellation, and output bytes | Four authorized test files plus this mandatory tracker checkpoint; no product source changed | Unsafe v1 parser, raw-error, and unrepaired-payload behavior is labeled legacy/current evidence only | Section 4 exact selector matrix; eight-path registry assertion; migration pool closure; Object Store typed/untyped redaction baseline | `app.operator` focused/service-backed and `module.collaboration` service-backed slices; format, Markdown, and diff integrity | Revert only SL-01 tests; no runtime state changed | Every observed branch has an exact selector, current owner slices pass, and SL-02 is the only authorized next slice |
| SL-02 | DONE | SL-01, owner authorization | Amended Core 01 to 110; adopted Core 00/01/03/04 Collaboration ownership, v2 transport, transaction, local security/journal rules, and typed-only Object Store classification; clarified domain quarantine vocabulary | Core 00, Core 01, Core 03, Core 04, `docs/domain.md`, and this tracker checkpoint only | Product behavior and conformance authority changed intentionally before projections/runtime | REQ-00-068, REQ-01-655/656, REQ-03-307, REQ-04-151/152, AC-535/536; domain glossary distinctions | Requirement-ID uniqueness scan; stale-owner search; `make lint-markdown`; diff/status integrity | Revert coordinated owner change; implementation remains legacy until downstream slices | RB-001/RB-002 owner gates are closed; every CRQ decision has one primary owner; SL-03 is the sole next slice |
| SL-03 | DONE | SL-02, contract authorization | Updated Recovery authored limits/schema to 110; added the v2-only Collaboration family, schema, registry, fixtures, validators/tests, generated Go embedding, and family registration | `contracts/index.json`, `contracts/recovery` inputs, new `contracts/collaboration`, contract generator/shape-checker sources and tests, Recovery exact-set tests, and Make-generated outputs | v1 is deliberately unreadable; generated extension integrity digest and topology render index changed because generator/checker sources changed | Recovery extra/missing/duplicate/misowned/misclassified/wrong-action/wrong-codec/contribution/Goose negatives; Collaboration v1/alias/result negatives; exact v2 registry/fixture order | JSON shape, generate, backend unit, generation drift, generated-artifact policy, Markdown/diff/status checkpoint | Revert authored inputs and regenerate; generated files remain tool-owned | Recovery is exactly 110/82 with Goose absent from `tables[]`; Collaboration is v2-only; all generation gates pass; SL-04 is next |
| SL-04 | DONE | SL-01, SL-02, SL-03, WF-07 | Split migration-evidence transport from Postgres semantics and Object Store process transport from durable effects; added exact Collaboration, registry, Object Store, and Recovery catalog selectors; allocated all changed identities through the Make-owned authoring command | Four authored family manifests, split migration/Object Store test entry points, prior characterization tests, Recovery catalog negatives, generated accounting, and this tracker | Legacy Collaboration selectors remain current-state evidence until SL-06; full-tier process rows require explicit row selection | Temporary old-to-new mapping in section 11; exact row selectors; Harness contract | Five focused and service-backed owner slices, explicit full-tier process rows, generate/shape/drift/policy, Harness contract | Revert authored inputs/test splits and regenerate; runtime aliases remain forbidden | Every target selector is unique, owner postconditions are split, all required owner gates pass, and SL-05 is next |
| SL-05 | DONE | SL-04, WF-05 | Moved Object Store and Collaboration adapters into dedicated same-package files, moved migration parsing beside its adapter, isolated shared transport mechanics, and recorded the narrow executor port map | `operator.go`, `operator_migration_evidence.go`, new `operator_collaboration.go`, `operator_object_store.go`, `operator_transport.go`, complete section 2 inventory, and this tracker | No semantics changed; legacy Collaboration v1 and Object Store string matching deliberately remain until SL-06 | SL-01 exact characterization plus eight-path registry and resource-lifecycle selectors | App owner slice, backend boundary, operator build, format, Markdown/diff/status checkpoint | Restore prior same-package file layout | Package/exports/registry/grammar/bytes/exits are unchanged; all 14 target files are inventoried; SL-06 is next |
| SL-06 | DONE | SL-05 | Introduced four private executors and narrow ports; implemented Collaboration v2 plus repaired-state transaction/private journal; removed Object Store string matching; updated guide and Harness evidence | Operator facade/source/tests, Collaboration service/integration tests, Recovery operator process evidence, app family manifest, generated accounting, guide, and this tracker | Intentional v2 break is complete; closure and owner boundaries are guarded by exact tests | Strict grammar/envelope/exit/redaction/timeout/delivery tests; cases 1-18; real process; repair, concurrency, attribution, rollback, commit-unknown, and typed Object Store evidence | All five focused/service-backed owners; shape/generation/drift/policy; Harness contract; boundary; operator build | Restore previous private wiring and v1 runtime behavior; no data migration exists | Each executor depends only on its command needs, v2 is exclusive, semantic SQL stays in Collaboration, and SL-07 is the only next slice |
| SL-07 | DONE | SL-06 | Removed obsolete private test helpers, completed the final Harness reconciliation, removed the temporary crosswalk, and retained broad validation/handoff evidence | `internal/app/operator/operator_test.go`, `tools/contractgen/collaboration_validation_test.go`, and this tracker | Cleanup was limited to proven-unused private test support plus one compile-correct validation wrapper | Final active-row reconciliation; exact final worktree `test-fast`; full `check`; retained successful-run maintenance | Pre-broad and retained-run `agent-finalize`; `test-fast`; `backend-unit`; `check`; Markdown/diff/status integrity | Restore the deleted private test helpers only if a future test demonstrates continuing value | No obsolete helper remains; every required gate has current successful evidence; no blocker or follow-on workstream remains |

The 2026-08-06 implementation directive authorized these historical slices.
Owner adoption in SL-02 and typed projection in SL-03 satisfied the hard
prerequisites for the Harness and behavioral changes completed in SL-04 through
SL-07. Their completion did not freeze private implementation details against
the separately planned cleanup in section 13.

## 8. Validation Plan

### Tracker-only validation

| Command | Required result | Scope |
| --- | --- | --- |
| `make lint-markdown` | exit 0 with a recorded run root | This tracker revision |
| `git diff --check` | exit 0 | Whitespace integrity |
| `git status --short` | Every changed path is within the authorized SL-00 through SL-07 write sets; the pre-existing staged tracker remains staged | Write-scope integrity |
| Structural `rg` checks | Sections 1-12, WF-00 through WF-08, RB-001 through RB-003, OA requirements/criteria, and seven appended session rows are present | Document completeness |

### Implementation validation matrix

| Layer | Exact commands | Required gate |
| --- | --- | --- |
| Owner discovery | `make task-guide ROLE=module-author OWNER=<owner>` and `make explain-test-owner OWNER=<owner>` for `app.operator`, `platform.postgres`, `platform.objectstore`, `module.collaboration`, and `module.recovery` | Before modifying each owner's tests/accounting |
| Focused unit/integration | `make test-slice OWNER=app.operator`; repeat with the four owners above | Before source movement and after every affected slice |
| Service-backed | `make service-backed-test-slice OWNER=app.operator`; repeat with the four owners above | Before source movement and after SL-04/SL-06 |
| Identifier authoring | `make author-test-row-id FAMILY_ID=<owner-qualified-family> CLAIM=<semantic-claim> SELECTOR_KEY=<exact-selector-key>` | Every newly allocated row ID |
| Generated/accounting | `make json-shape-check`; `make generate`; `make generate-drift`; `make generated-artifact-policy-check` | SL-03 and SL-04 |
| Static/build | `make backend-module-boundary-check`; `make build-operator` | SL-05 through SL-07 |
| Browser/E2E | Not applicable to this local CLI facade | No browser, HTTP, WS, frontend, or grid surface may be introduced |
| Broader completion | `make agent-finalize`; `make test-fast`; `make check` | SL-07 only after all narrow gates pass |

### Retained execution evidence

| Command | Result | Run root | Coverage limit |
| --- | --- | --- | --- |
| `make test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T153158Z-p1322227` | Five currently cataloged rows only |
| `make backend-module-boundary-check` | passed | `.cartulary/test-results/20260806T153158Z-p1322353` | Current declared boundary policy only |
| `make generate-drift` | passed | `.cartulary/test-results/20260806T153158Z-p1322195` | Did not detect RB-001 |
| `make json-shape-check` | passed | `.cartulary/test-results/20260806T153439Z-p1327815` | Did not detect RB-001 |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T155153Z-p1334181` | Previous tracker revision only |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T164411Z-p1348214` | NLSpec-style tracker revision; no product validation |
| `make backend-module-boundary-check` | passed | `.cartulary/test-results/20260806T170006Z-p1356399` | SL-00 current-baseline boundary evidence |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T170007Z-p1356769` | SL-00 pre-checkpoint tracker evidence |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T170222Z-p1358966` | SL-00 post-checkpoint integrity evidence |
| `make format` | passed | `.cartulary/test-results/20260806T171636Z-p1458038` | SL-01 authored Go test formatting |
| `make test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T171639Z-p1461024` | Seven current app-owner units including legacy v1 adapter, registry, Object Store, and migration evidence characterization |
| `make service-backed-test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T171554Z-p1456287` | Three current app-owner service-backed units |
| `make service-backed-test-slice OWNER=module.collaboration` | failed | `.cartulary/test-results/20260806T170752Z-p1389440` | Initial SL-01 fixture used pgx-prepared multi-statement SQL; 32/33 passed and the failure was related only to the new fixture setup |
| `make service-backed-test-slice OWNER=module.collaboration` | passed | `.cartulary/test-results/20260806T171132Z-p1420785` | All 33 current Collaboration units passed after splitting fixture statements and scoping the injected trigger |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T171811Z-p1463467` | SL-01 tracker checkpoint; subsequent diff/status integrity also passed |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T172312Z-p1468450` | SL-02 coordinated owner/domain adoption before tracker checkpoint |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T172503Z-p1470596` | SL-02 post-tracker checkpoint |
| `make json-shape-check` | failed | `.cartulary/test-results/20260806T173141Z-p1477580` | New active family was not yet recognized by the authored repository-wide shape checker; related SL-03 failure |
| `make json-shape-check` | failed | `.cartulary/test-results/20260806T173208Z-p1478425` | Authored checker was corrected and the gate then correctly required regeneration of stale topology inputs |
| `make generate` | failed | `.cartulary/test-results/20260806T173224Z-p1478888` | New validator duplicated an existing package-local helper name; related compile defect, corrected before regeneration |
| `make generate` | passed | `.cartulary/test-results/20260806T173257Z-p1483566` | Generated Recovery, Collaboration, extension-integrity, and topology outputs from authored inputs |
| `make json-shape-check` | passed | `.cartulary/test-results/20260806T173319Z-p1486076` | Three repository shape units pass against generated state |
| `make backend-unit` | passed | `.cartulary/test-results/20260806T173327Z-p1486539` | 59/59 packages, including new exact-set and v2 compatibility-negative tests |
| `make generate-drift` | passed | `.cartulary/test-results/20260806T173405Z-p1507666` | Four generation drift units pass |
| `make generated-artifact-policy-check` | passed | `.cartulary/test-results/20260806T173416Z-p1510374` | Three policy units prove generated roots were not hand-edited |
| `make format` | passed | `.cartulary/test-results/20260806T174300Z-p1518329` | SL-04 split Go entry points and authored Harness JSON are formatted and admitted |
| `make generate` | passed | `.cartulary/test-results/20260806T174306Z-p1521332` | Authored Harness ownership inputs and prior contract inputs generate without drift |
| `make test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T174325Z-p1523907` | Ten exact app transport/registry/resource units |
| `make service-backed-test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T174341Z-p1526352` | Four standard-tier app service-backed units |
| `make test-slice OWNER=module.collaboration` | passed | `.cartulary/test-results/20260806T174359Z-p1527898` | All 44 focused Collaboration-owned units |
| `make test-slice OWNER=module.recovery` | passed | `.cartulary/test-results/20260806T174531Z-p1556563` | All 35 focused Recovery-owned units, including exact catalog admission negatives |
| `make test-slice OWNER=platform.postgres` | passed | `.cartulary/test-results/20260806T174619Z-p1591647` | Nine focused Postgres-owned units, including migration-evidence semantics |
| `make test-slice OWNER=platform.objectstore` | passed | `.cartulary/test-results/20260806T174632Z-p1593757` | Five focused Object Store-owned units |
| `make service-backed-test-slice OWNER=module.collaboration` | passed | `.cartulary/test-results/20260806T174646Z-p1595492` | All 33 service-backed Collaboration units |
| `make service-backed-test-slice OWNER=module.recovery` | passed | `.cartulary/test-results/20260806T174813Z-p1621883` | All 24 standard-tier service-backed Recovery units |
| `make service-backed-test-slice OWNER=platform.postgres` | passed | `.cartulary/test-results/20260806T174850Z-p1654315` | Five service-backed Postgres units |
| `make service-backed-test-slice OWNER=platform.objectstore` | passed | `.cartulary/test-results/20260806T174907Z-p1656047` | Five standard-tier service-backed Object Store units |
| Explicit app/Object Store process rows | passed | `.cartulary/test-results/20260806T174916Z-p1657468`; `.cartulary/test-results/20260806T174919Z-p1658845` | Full-tier operator transport and durable bucket-effect selectors each pass under their primary owner |
| `make harness-contract` | passed | `.cartulary/test-results/20260806T174924Z-p1660285` | Authored selectors, identities, ownership, fixtures, and accounting satisfy the extended Harness contract |
| `make json-shape-check` | passed | `.cartulary/test-results/20260806T174946Z-p1660816` | All authored/generated JSON shapes pass after ownership migration |
| `make generate-drift` | passed | `.cartulary/test-results/20260806T174948Z-p1661233` | Four generation-drift units pass after Harness migration |
| `make generated-artifact-policy-check` | passed | `.cartulary/test-results/20260806T174956Z-p1663910` | Three generated-artifact policy units pass after Harness migration |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T175138Z-p1664874` | SL-04 tracker checkpoint after all Harness and generated gates |
| `make format` | passed | `.cartulary/test-results/20260806T175440Z-p1669225` | SL-05 authored Go source movement and documentation formatting |
| `make test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T175446Z-p1672258` | All ten exact app-owned rows remain stable after same-package movement |
| `make backend-module-boundary-check` | passed | `.cartulary/test-results/20260806T175458Z-p1674742` | Three boundary units preserve the exclusive binary edge and allowed imports |
| `make build-operator` | passed | `.cartulary/test-results/20260806T175459Z-p1675210` | Operator binary builds after the mechanical split |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T175716Z-p1686684` | SL-05 tracker checkpoint after source inventory and executor-map updates |
| `make service-backed-test-slice OWNER=module.collaboration` | failed | `.cartulary/test-results/20260806T181055Z-p1706632` | Initial SL-06 compile attempt found one unused test import introduced while replacing legacy fixtures; related and corrected before semantic execution |
| Durable Collaboration row | passed | `.cartulary/test-results/20260806T181247Z-p1741828` | Initial repaired-state, concurrency, rollback, journal, and cancellation transaction evidence passed |
| `make test-slice OWNER=app.operator` | failed | `.cartulary/test-results/20260806T181851Z-p1749300` | New real-process row correctly exposed that the retained root `operator` binary predated SL-06; source/unit rows passed and the binary was rebuilt before rerun |
| `make build-operator` | passed | `.cartulary/test-results/20260806T181922Z-p1752485` | Rebuilt the injected operator binary against strict v2 before process evidence |
| `make test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T181936Z-p1763322` | All 11 app units, including strict v2 and the real-process transaction, pass with the rebuilt binary |
| `make generate` | passed | `.cartulary/test-results/20260806T181949Z-p1764940` | Regenerated Harness accounting after final v2 process and row-identity additions |
| `make service-backed-test-slice OWNER=module.collaboration` | passed | `.cartulary/test-results/20260806T182004Z-p1767326` | All 33 service-backed Collaboration units pass against the adopted transaction |
| `make test-slice OWNER=module.collaboration` | passed | `.cartulary/test-results/20260806T182136Z-p1795081` | All 44 focused Collaboration units pass |
| `make service-backed-test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T182305Z-p1822751` | All five service-backed app units pass, including process/resource behavior |
| `make test-slice OWNER=module.recovery` | passed | `.cartulary/test-results/20260806T182320Z-p1824283` | All 35 focused Recovery units pass with the reconciled catalog |
| `make service-backed-test-slice OWNER=module.recovery` | passed | `.cartulary/test-results/20260806T182407Z-p1862661` | All 24 service-backed Recovery units pass |
| `make test-slice OWNER=platform.postgres` | passed | `.cartulary/test-results/20260806T182443Z-p1895116` | All nine focused Postgres units pass |
| `make service-backed-test-slice OWNER=platform.postgres` | passed | `.cartulary/test-results/20260806T182501Z-p1899330` | All five service-backed Postgres units pass |
| `make test-slice OWNER=platform.objectstore` | passed | `.cartulary/test-results/20260806T182516Z-p1901076` | All five focused Object Store units pass with typed-only classification |
| `make service-backed-test-slice OWNER=platform.objectstore` | passed | `.cartulary/test-results/20260806T182530Z-p1902807` | All five service-backed Object Store units pass |
| `make json-shape-check` | passed | `.cartulary/test-results/20260806T182538Z-p1904184` | All three shape units pass against v2 and final owner inputs |
| `make generate-drift` | passed | `.cartulary/test-results/20260806T182543Z-p1904602` | All four generation-drift units pass |
| `make generated-artifact-policy-check` | passed | `.cartulary/test-results/20260806T182553Z-p1907305` | All three generated-artifact policy units pass |
| `make backend-module-boundary-check` | passed | `.cartulary/test-results/20260806T182556Z-p1907796` | Three boundary units preserve the binary edge and semantic-owner imports |
| `make build-operator` | passed | `.cartulary/test-results/20260806T182601Z-p1908271` | Strict-v2 operator binary is current and buildable |
| `make harness-contract` | passed | `.cartulary/test-results/20260806T182614Z-p1918843` | Final SL-06 selectors, ownership, identities, and accounting are valid |
| Durable Collaboration row after commit-failure hardening | passed | `.cartulary/test-results/20260806T182812Z-p1923080` | Commit failure maps outcome unknown and leaves the deliberately uncommitted transaction rolled back |
| v2 app transport row after timeout hardening | passed | `.cartulary/test-results/20260806T182901Z-p1927979` | Strict timeout bounds Postgres setup and transactional work; exact v2 transport still passes |
| `make test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T183451Z-p1930355` | Final post-hardening app rerun passes all 11 units |
| `make test-slice OWNER=module.collaboration` | failed | `.cartulary/test-results/20260806T183505Z-p1932573` | All backend evidence passed; one unrelated pre-existing browser conflict-resolution test timed out waiting 25 seconds for a local value, while four sibling browser tests passed |
| `make service-backed-test-slice OWNER=module.collaboration` | passed | `.cartulary/test-results/20260806T183714Z-p1961437` | Final post-hardening service-backed rerun passes all 33 units |
| `make test-slice OWNER=module.collaboration` | passed | `.cartulary/test-results/20260806T183840Z-p1987749` | Immediate focused retry passes all 44 units, classifying the prior unchanged browser failure as transient |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T184015Z-p2014967` | SL-06 tracker content passes Markdown lint before the final checkpoint integrity rerun |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T184040Z-p2016660` | SL-06 post-checkpoint Markdown evidence; diff and status integrity also passed |
| `make format` | passed | `.cartulary/test-results/20260806T184135Z-p2018557` | SL-07 private test-helper cleanup and authored source formatting |
| `make test-slice OWNER=app.operator` | passed | `.cartulary/test-results/20260806T184141Z-p2021584` | All 11 app-owner units pass after obsolete helper removal |
| `make agent-finalize` | passed | `.cartulary/test-results/20260806T184155Z-p2024072` | Required pre-broad finalization completed before `test-fast` and `check` |
| `make test-fast` | failed | `.cartulary/test-results/20260806T184209Z-p2026729` | Product units reached 333/348 before the shared `/tmp` filesystem exhausted space; this was an environment failure, not a test assertion |
| `make test-fast` with task-local Go caches | passed | `.cartulary/test-results/20260806T184525Z-p2090874` | All 348 units pass after moving only this task's Go caches out of the saturated shared `/tmp` cache |
| `make check` with task-local Go caches | failed | `.cartulary/test-results/20260806T184557Z-p2097601` | 689/716 units completed before Go work directories still allocated under saturated `/tmp`; no product assertion failed |
| `make check` with a repo-local `TMPDIR` | failed | `.cartulary/test-results/20260806T184956Z-p2209841` | 713/716 units passed; the Harness smoke correctly rejected a scratch root inside the repository, and vet exposed a related validator test function-signature mismatch |
| `make format` | passed | `.cartulary/test-results/20260806T185156Z-p2304980` | Formatted the compile-correct validator wrapper used by the generic negative-test helper |
| `make backend-unit` with long external `TMPDIR` | failed | `.cartulary/test-results/20260806T185207Z-p2308025` | 57/59 packages passed; the remaining service fixtures rejected an overlong Unix socket path, so the external scratch root was shortened |
| `make backend-unit` with short external `TMPDIR` | passed | `.cartulary/test-results/20260806T185259Z-p2329275` | All 59 backend packages pass after using `/var/tmp/ct-op` for environment scratch space |
| `make check` with short external `TMPDIR` and task-local Go caches | passed | `.cartulary/test-results/20260806T185312Z-p2330898` | All 716 units pass against the final implementation, including static, generation, boundary, build, security, and Harness gates |
| `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260806T185312Z-p2330898` | passed | `.cartulary/test-results/20260806T185433Z-p2411499` | Successful full-check evidence is retained and final-run maintenance passes |
| `make test-fast` with short external `TMPDIR` and task-local Go caches | passed | `.cartulary/test-results/20260806T185634Z-p2414620` | Exact final worktree rerun passes all 348 units after the validator-wrapper fix |
| `make lint-markdown` | passed | `.cartulary/test-results/20260806T190000Z-p2416212` | Completed SL-07 tracker content passes the mandatory Markdown checkpoint |

The original planning baseline did not run the service-backed operator slice,
Recovery process rows, generated-artifact policy, or build/broad gates. The
SL-04 through SL-07 evidence above supersedes those gaps. No plan-required gate
was skipped. `make ci` and `make release-check` were not separately requested;
the required 716-unit `make check` gate passed. Browser/visual, frontend, and
migration-specific rollout commands were not run because this change adds no
such surface or migration.

## 9. Top-Level Work Tracker

The controlled status vocabulary is unchanged; every tracked work item below
is now `DONE`.

| ID | Work item | Workstream | Status | Depends on | Evidence | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| OT-001 | Freeze scope, authority, baseline, and normative vocabulary | WF-00 | DONE | None | Section 1 | Write and authority boundaries are explicit |
| OT-002 | Inventory all target files | WF-01 | DONE | OT-001 | Section 2 | All 14 post-SL-05 files remain accounted for |
| OT-003 | Reconcile live Recovery counts and identities | WF-01 | DONE | OT-002 | Section 5 | Live 110-authored plus separate-Goose state is explicit |
| OT-004 | Resolve Recovery owner contradiction | WF-02 | DONE | Owner adoption | Core 01 REQ-01-647 now states 110/82/separate Goose | Adopted owners state one exact coherent identity set |
| OT-005 | Freeze exact live Collaboration interface | WF-01 | DONE | OT-002 | Section 4 | Grammar, defaults, output, exits, and mutation are explicit |
| OT-006 | Complete Collaboration characterization matrix | WF-03 | DONE | OT-005 | Section 4 SL-01 exact selector matrix and retained roots | Every current-state case has exact evidence; legacy assertions are marked for v2 replacement |
| OT-007 | Adopt Collaboration owners and typed projection | WF-02 | DONE | OT-006, owner authorization | Owners adopted in SL-02; v2-only family generated in SL-03 | Core owners and typed projections close every row |
| OT-008 | Complete boundary/coupling scan | WF-04 | DONE | OT-002 | Sections 3 and 5 | No semantic owner leaks into the facade plan |
| OT-009 | Correct Harness owner accounting and identities | WF-07 | DONE | OT-006, OT-007 | SL-04 family manifests, temporary crosswalk, author-ID runs, and Harness contract | Correct owner partitions, new IDs/crosswalk, no aliases |
| OT-010 | Design private command executors | WF-05 | DONE | OT-004, OT-007, OT-009 | Section 3 private executor port map and SL-05 file split | Approved ports and source-movement map exist |
| OT-011 | Execute behavior-preserving structural split and adopted behavior cutover | WF-06 | DONE | OT-010 | SL-05 source split; SL-06 four executors, Collaboration v2 transaction/journal, typed-only Object Store mapping, and passing narrow gates | Stable facade with narrow executors and passing focused gates |
| OT-012 | Run final validation and handoff | WF-08 | DONE | OT-011 | Section 8 final broad roots and section 11 reconciliation | Every required run has successful retained evidence |
| OT-013 | Record non-applicable frontend/grid/network surfaces | WF-04 | DONE | OT-002 | Sections 3-5 | No unsupported workstream is present |
| OT-014 | Revise tracker in NLSpec voice | WF-00, WF-08 | DONE | OT-001..OT-008 | This document | OA requirements, criteria, mappings, and history are complete |

## 10. Session Handoff Log

Existing history is preserved. The `2026-08-06T16:37:45Z` rows record this
NLSpec-style documentation revision.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T15:34:42Z | Codex / operator-app planning | Framework read first; baseline and source hierarchy frozen; Plan Mode ended without a write | Touched in implementation turn: this tracker only. Inspected: framework, Core 00-04, domain, Harness NLSpec, dev guide | `git status --short`; `git branch --show-current`; `git rev-parse HEAD` | Baseline `main`/`0bd8bd6...` initially clean; planning-only scope confirmed | None for tracker creation | Keep implementation disabled until separately authorized |
| 2026-08-06T16:37:45Z | Codex / NLSpec tracker revision | Advisory notes converted into refactor requirements without promoting them to owner authority | Touched: this tracker only. Inspected: analysis notes, NLSpec editorial guide, adopted owners | `git status --short`; owner/reference searches | Normative refactor posture and authority classes added | RB-001, RB-002 | Obtain owner authorization before any non-tracker change |
| 2026-08-06T17:00:34Z | Codex / SL-00 execution | User-authorized remediation execution started at the unchanged baseline; only the staged tracker pre-existed | Touched: this tracker only. Inspected: all operator files, Recovery inputs/runtime/validator, boundary policy | Branch/HEAD/status inventory; exact `jq`/`rg` reconciliation; `git diff --check` | `main` at `0bd8bd6...`; tracker remains the only changed path; SL-00 is complete | RB-001 and RB-002 remain gated on SL-02 adoption | Start SL-01 characterization only |
| 2026-08-06T17:13:44Z | Codex / SL-01 execution | Test-only characterization is complete; no product, contract, owner, or Harness input changed | Touched: this tracker and four test files named in the test/harness row | Five owner task guides/explanations; format; focused/service-backed owner slices; integrity checks | Cases 1-18 and the eight registry paths have exact current-state selectors; SL-01 is complete | RB-001 and RB-002 now await the authorized SL-02 owner adoption | Start SL-02 owner/vocabulary adoption only |
| 2026-08-06T17:23:46Z | Codex / SL-02 execution | Coordinated owner adoption is complete; downstream projections and runtime remain unchanged | Touched: Core 00/01/03/04, `docs/domain.md`, and this tracker | Requirement-ID uniqueness and stale-owner searches; `make lint-markdown`; diff/status integrity | REQ-00-068, REQ-01-655/656, REQ-03-307, REQ-04-151/152, and AC-535/536 are adopted | No owner blocker remains; downstream 109/v1 projection drift is explicit SL-03 work | Start SL-03 typed contracts/generation only |
| 2026-08-06T17:34:36Z | Codex / SL-03 execution | Recovery and Collaboration typed projections now match the adopted owners | Touched: authored contract/generator/shape-checker/test inputs, Make-generated outputs, and this tracker | JSON shape; generate; backend unit; drift; generated policy; exact `jq` reconciliation | Recovery 110/82/one Extension ledger/no catalog Goose; Collaboration major 2/no historical reader/no compatibility flags | No contract blocker remains; Harness ownership is still intentionally unmigrated | Start SL-04 Harness ownership migration only |
| 2026-08-06T17:50:02Z | Codex / SL-04 execution | Harness ownership now follows each normative postcondition; temporary crosswalk records every retired mixed identity | Touched: four family manifests, split migration/Object Store tests, generated accounting, and this tracker | Ten author-ID runs; five focused and five service-backed owner slices; explicit full-tier rows; Harness/generation gates | OT-009 and RB-003 implementation are complete; no aliases or duplicate selectors exist | None | Start SL-05 mechanical same-package split only |
| 2026-08-06T17:56:16Z | Codex / SL-05 execution | Mechanical source movement is complete and the four private executor port sets are closed for SL-06 | Touched: five operator production files, complete 14-file inventory, port map, and this tracker | Format; app owner slice; backend boundary; operator build | No command behavior changed; OT-010 is DONE | None | Start SL-06 adopted behavior and executor implementation only |
| 2026-08-06T18:29:24Z | Codex / SL-06 execution | Adopted runtime cutover is complete; strict Collaboration v2 is exclusive and Object Store specificity is typed-only | Touched: operator facade/executors/tests, Collaboration service/integration tests, process evidence, app Harness manifest/generated accounting, dev guide, and this tracker | Five focused and five service-backed owners; v2/process rows; shape/generation/drift/policy; Harness; boundary; build | OT-011 is DONE; every narrow SL-06 gate passes at section 8 roots | None | Start SL-07 cleanup, broad validation, and final handoff only |
| 2026-08-06T18:57:24Z | Codex / SL-07 completion | All authorized remediation slices and acceptance criteria are complete | Touched: final private-test cleanup, validator test wrapper, and this tracker; inspected the complete authorized diff | Pre-broad and retained-run finalization; final `test-fast`; full `check`; Markdown/diff/status integrity | SL-00 through SL-07 and OT-009 through OT-012 are `DONE`; the user's staged tracker history remains preserved | None | Handoff complete; no follow-on workstream is authorized or required |
| 2026-08-06T20:59:24Z | Codex / PC planning | The completed remediation remains historical; a separate legacy-cleanup and production-hardening iteration is decision-complete but unstarted | Touched: this tracker only. Inspected: current operator source/exports/callers, target tests, owner selectors, boundary policy, and command-retirement rules | Read-only caller/export/test/selector inventory; owner task guidance; `make lint-markdown`; diff/status integrity | Section 13 fixes the PC-00 through PC-04 scope, deletion evidence, immutable-ID posture, gates, and acceptance criteria at clean HEAD `846d024...`; Markdown lint passed at `.cartulary/test-results/20260806T205911Z-p2678277` | None for planning; implementation remains separately gated | Await an explicit instruction to start PC-00; do not mutate product or Harness files during this documentation step |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T15:34:42Z | Codex / operator-app planning | Exact application facade is legitimate; private root file mixes four command-owner areas | Touched: this tracker only. Inspected: all `internal/app/operator/**`, `cmd/operator/main.go`, direct dependencies, boundary policy | `make backend-module-boundary-check`; relevant `make explain-target` | Passed; run root `.cartulary/test-results/20260806T153158Z-p1322353` | RB-001, RB-002 | Characterize, then implement private executors in a later task |
| 2026-08-06T16:37:45Z | Codex / NLSpec tracker revision | Thin-facade boundary retained; structural movement moved behind owner, Harness, and characterization gates | Touched: this tracker only. Inspected: operator root/registry/Recovery adapter and direct services | Exact-file inspection and import search | OA-REQ-003 and OA-REQ-016 define the stable boundary | RB-001, RB-002 | Complete WF-03, WF-02, and WF-07 before WF-05 |
| 2026-08-06T17:00:34Z | Codex / SL-00 execution | Eleven target files and the exclusive `cmd/operator` edge remain unchanged; no new surface is present | Touched: this tracker only. Inspected: target tree, caller, and boundary manifest | `make backend-module-boundary-check` | Passed at `.cartulary/test-results/20260806T170006Z-p1356399` | None for characterization | Preserve this boundary through SL-01 through SL-04 |
| 2026-08-06T17:13:44Z | Codex / SL-01 execution | Product imports and the exclusive binary edge remain unchanged | Touched: test files only; inspected the runner, registry, and Collaboration service boundary | Exact registry test; current owner slices | Exactly eight canonical paths are asserted; no production boundary moved | None for SL-02 documentation | Preserve the frozen boundary while adopting owners |
| 2026-08-06T17:23:46Z | Codex / SL-02 execution | Thin-facade boundary is now normative: semantic transition stays in Collaboration and transport stays in the app facade | Touched: owner/docs only | Core 00 owner-matrix review | No SQL, plugin bus, dependency container, or new surface is authorized in `internal/app/operator` | None | Project owners without moving implementation in SL-03 |
| 2026-08-06T17:34:36Z | Codex / SL-03 execution | Contract embedding introduces no product import or command dependency | Touched: generated contract packages and validation tests only | Generated-artifact policy and status review | `internal/gen/contractcollaboration` is data-only; facade source remains unchanged | None | Migrate verification ownership before structural source movement |
| 2026-08-06T17:50:02Z | Codex / SL-04 execution | Product imports remain unchanged; test selectors now distinguish facade transport from Postgres/Object Store semantics | Touched: operator and Recovery process tests plus authored family manifests | Five focused slices; Harness contract | Runtime-binary use no longer implies semantic ownership; the root facade is still unsplit | None | Move command adapters mechanically in SL-05 without behavior change |
| 2026-08-06T17:56:16Z | Codex / SL-05 execution | Root facade now contains only entry/default composition/registry concerns; command bodies occupy dedicated same-package files | Touched: `operator.go`, migration adapter, and three new private source files | App slice; backend boundary; operator build | Exclusive `cmd/operator` edge and eight paths remain intact; broad dependency fields remain only until SL-06 | None | Replace broad fields with four narrow executors in SL-06 |
| 2026-08-06T18:29:24Z | Codex / SL-06 execution | Root runner now holds four narrow executors and no command-specific services; Collaboration SQL remains module-owned | Touched: operator production files and Collaboration service | Exact registry, five owner slices, boundary, build | Exclusive binary edge, `RunOperatorCLIContext`, five Recovery paths, and eight total paths pass; no bus/container was added | None | Remove only proven-unused private helpers in SL-07 |
| 2026-08-06T18:57:24Z | Codex / SL-07 completion | Final facade retains four private executors and the frozen binary-facing surface | Touched: removed only unused private test helpers; no production boundary changed | App owner rerun; 716-unit `make check`; final diff inspection | Boundary, build, resource closure, five Recovery commands, and exactly eight paths pass | None | Preserve the executor/import boundary in future command work |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T15:34:42Z | Codex / operator-app planning | No frontend shell, browser route, HTTP/WS Recovery surface, grid adapter, or selector contract is present | Touched: this tracker only. Inspected: target imports/callers and Core 04 Recovery rules | Repository search; no browser command was run | Frontend/browser workstream is not applicable | None | Re-scan only if a future diff introduces frontend coupling |
| 2026-08-06T16:37:45Z | Codex / NLSpec tracker revision | Non-applicability retained as an explicit boundary requirement | Touched: this tracker only. Inspected: target/caller searches and Core local-operator rules | Structural `rg` searches | No frontend, grid, HTTP, WS, saved-view, entity-row, or test-util runtime work is authorized | None | Reject any future slice that introduces such a surface without owner adoption |
| 2026-08-06T17:13:44Z | Codex / SL-01 execution | Characterization introduced no frontend, browser, HTTP, WebSocket, grid, saved-view, or public action surface | Touched: backend tests and tracker only | Diff/status inspection | Local operator scope remains exact | None | Keep these surfaces explicitly negative in SL-02/Core 04 |
| 2026-08-06T17:23:46Z | Codex / SL-02 execution | Core 03/Core 04 explicitly forbid public, browser, workbook, HTTP, WebSocket, common-job, incident-action, and public-audit additions | Touched: Core 03/Core 04 and domain vocabulary | Owner/negative-surface review | Collaboration requeue remains local and private | None | Preserve negative surfaces in projection and runtime tests |
| 2026-08-06T17:34:36Z | Codex / SL-03 execution | New contracts describe only the local operator surface and semantic transaction | Touched: `contracts/collaboration` and generated Go embedding | Contract path and generated-package inspection | No frontend/TypeScript projection, public route, browser, grid, or WebSocket artifact was added | None | Keep Harness rows backend-only |
| 2026-08-06T17:50:02Z | Codex / SL-04 execution | Harness migration remains backend-only and declares no browser/runtime frontend evidence | Touched: backend Go tests and test-family manifests only | Diff and selector inspection | No frontend, HTTP, WebSocket, grid, saved-view, or public action surface was introduced | None | Preserve the negative surface during SL-05 source movement |
| 2026-08-06T17:56:16Z | Codex / SL-05 execution | Same-package movement added no network, browser, frontend, or public projection surface | Touched: backend operator Go files and tracker only | Boundary/diff inspection | Local CLI scope remains exact | None | Keep SL-06 executor and behavior work local/private |
| 2026-08-06T18:29:24Z | Codex / SL-06 execution | Runtime cutover remains deployment-local and writes only the private raw operator journal | Touched: backend CLI/module/tests/guide only | Boundary, negative-surface search, owner slices | No frontend, browser, HTTP, WS, grid, saved-view, public action/audit, or common-job surface exists | None | Preserve this negative result through final diff review |
| 2026-08-06T18:57:24Z | Codex / SL-07 completion | Final diff remains local CLI/backend-only | Inspected: complete authorized diff and generated inventory; no frontend file touched | Full `make check`; status/diff review | No frontend, browser, HTTP, WS, grid, saved-view, public action/audit, or common-job surface was introduced | None | No frontend handoff or rollout is required |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T15:34:42Z | Codex / operator-app planning | Recovery/Extension projections are consumed, not owned; 109/110 drift identified | Touched: this tracker only. Inspected: Recovery schema/fixture/generated catalog/runtime/validator | `make generate-drift`; `make json-shape-check` | Both passed at recorded roots; neither detected RB-001 | RB-001 | Obtain owner/contracts authorization and never hand-edit generated files |
| 2026-08-06T16:37:45Z | Codex / NLSpec tracker revision | Exact diff disproved the advisory Goose-in-catalog hypothesis; adopted Extensions contribution conflicts with Core count | Touched: this tracker only. Inspected: `REQ-01-647`, `EXT-REQ-236`, schema, fixture, validator, runtime, synthetic manifest | `jq`; `rg`; exact source inspection | 110 authored entries include `extension_migration_ledger`; Goose is separate; RB-001 relabeled owner contradiction | RB-001 | Coordinate owner adoption before schema, fixture, generator, or runtime changes |
| 2026-08-06T17:00:34Z | Codex / SL-00 execution | Current reconciliation repeated against HEAD | Touched: this tracker only. Inspected: Recovery registry, schema, canonical fixture, runtime constants, generator validator | Exact `jq` counts and `rg` constants | Registry/schema remain 109; fixture/runtime/validator are 110; required count is 82; Extension ledger occurs once; Goose occurs zero times in `tables[]` | RB-001 | Characterize first, then adopt 110 in SL-02 |
| 2026-08-06T17:13:44Z | Codex / SL-01 execution | No contract or generated artifact changed; legacy v1 output and unsafe failure behavior are now exact replacement evidence | Touched: tests and tracker only | Exact-byte v1 assertion; current owner slices | SL-03 remains gated; v1 has no compatibility claim | RB-001 and RB-002 | Adopt 110 and Collaboration v2 owners in SL-02 before authoring projections |
| 2026-08-06T17:23:46Z | Codex / SL-02 execution | Recovery 110 and Collaboration v2 literals are authoritative; generated inputs remain intentionally stale until the next slice | Touched: owner/docs only | Exact requirement/reference search | Core owner contradiction and unowned-v1 gap are closed | Contracts still say Recovery 109 and have no Collaboration family | Repair authored inputs and generate outputs in SL-03; add no v1 reader |
| 2026-08-06T17:34:36Z | Codex / SL-03 execution | Contract family registry now has active Collaboration order 9; later families shifted contiguously; v2 schema/registry/fixtures are embedded | Touched: `contracts/index.json`, `contracts/recovery`, new `contracts/collaboration`, generator validation, generated packages/integrity/topology | `make generate`; shape/drift/policy gates | All gates pass; initial three related failures and their fixes are retained in section 8 | None | Do not add v1 runtime support during Harness/implementation slices |
| 2026-08-06T17:50:02Z | Codex / SL-04 execution | Contract projections are unchanged semantically; generation reconciles the new Harness ownership inputs | Touched: authored family manifests and Make-owned generated accounting | `make generate`; JSON shape; generation drift; generated policy | All gates pass at the SL-04 roots in section 8; no generated root was hand-edited | None | Keep SL-05 product behavior byte-stable against these projections |
| 2026-08-06T17:56:16Z | Codex / SL-05 execution | No contract input or generated artifact changed during mechanical source movement | Touched: operator authored Go files and tracker only | Status/diff review; operator build | Generated Collaboration v2 remains ready but unconsumed until SL-06 | None | Consume only the adopted v2 projection during SL-06 behavior work |
| 2026-08-06T18:29:24Z | Codex / SL-06 execution | Runtime emits only the generated-family v2 contract; Recovery remains 110/82 with Goose separate | Touched: operator runtime/tests, app Harness input, Make-generated accounting | Generate; shape; drift; generated policy | v1 exists only in compatibility-negative fixtures/tests; generated roots pass policy and drift | None | Final generation/policy confirmation in SL-07 |
| 2026-08-06T18:57:24Z | Codex / SL-07 completion | Authored owners, contracts, runtime, and generated projections are reconciled | Inspected: Recovery and Collaboration inputs plus all generated changes | Full `make check`; retained-run finalization; generated-artifact policy within the broad gate | Recovery is 110/82 with separate Goose; Collaboration is v2-only; generated drift/policy pass with no hand edit | None | Future contract changes must continue through authored inputs and Make generation |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T15:34:42Z | Codex / operator-app planning | Five current `app.operator` rows pass; accounting is incomplete/misaligned | Touched: this tracker only. Inspected: target tests, Recovery process tests, family/owner inputs | Task guide; owner explanation; `make test-slice OWNER=app.operator` | Focused slice passed at recorded root; service-backed/process rows not run | RB-003 | Add characterization, then correct authored accounting |
| 2026-08-06T16:37:45Z | Codex / NLSpec tracker revision | Harness authority now determines owners and immutable identity actions; RB-003 planning ambiguity is closed | Touched: this tracker only. Inspected: Harness identifier/ownership/crosswalk rules and exact current rows/tests | `make explain-target TARGET=author-test-row-id DETAIL=summary`; `jq`; `rg` | Owner changes require new IDs; mixed tests require exact splitting; runtime binary remains independent | RB-002 blocks semantic adoption, not the Harness rule | Implement section 11 mapping after characterization and owner adoption |
| 2026-08-06T17:13:44Z | Codex / SL-01 execution | Current behavior has explicit selectors before owner migration | Touched: `internal/app/operator/operator_test.go`, `operator_registry_test.go`, `operator_migration_evidence_test.go`, `internal/modules/collaboration/stream_integration_test.go`, and this tracker | `make format`; `make test-slice OWNER=app.operator`; service-backed slices for `app.operator` and `module.collaboration` | App focused 7/7, app service-backed 3/3, Collaboration service-backed 33/33; initial 32/33 failure was a related fixture defect and is retained above | Harness row ownership remains intentionally unchanged until SL-04 | Adopt owners in SL-02, generate contracts in SL-03, then migrate Harness accounting |
| 2026-08-06T17:23:46Z | Codex / SL-02 execution | No test-family or generated Harness input changed | Touched: owner/docs only | Owner acceptance trace to AC-535/536 | Existing legacy characterization remains evidence, not conformance | Harness migration is gated on SL-03 projections | Generate contracts next; allocate row IDs only in SL-04 |
| 2026-08-06T17:34:36Z | Codex / SL-03 execution | Contract-generator and Recovery admission negatives are executable; Harness owner rows are still unchanged by design | Touched: `tools/contractgen/collaboration_validation_test.go`, `internal/platform/recoverystate/catalog_test.go`, prior SL-01 tests | `make backend-unit` | 59/59 pass, covering compatibility rejection and all requested Recovery mutation families | Exact catalog selection/owner allocation remains SL-04 | Allocate immutable owner-qualified IDs and temporary crosswalk next |
| 2026-08-06T17:50:02Z | Codex / SL-04 execution | Mixed migration/Object Store rows are retired; registry, resource, Collaboration, and Recovery catalog tests have exact owner-qualified selectors | Touched: `operator_migration_evidence*_test.go`, `operator_process_test.go`, four family manifests, and this tracker | Make-owned ID allocation; five focused/service-backed slices; two explicit full-tier process rows; Harness contract | Every target selector is selected once; all required slices pass at recorded roots | Legacy Collaboration row intentionally remains current-state until SL-06 | Perform only the byte-stable SL-05 source split next |
| 2026-08-06T17:56:16Z | Codex / SL-05 execution | Existing exact selectors remain byte-stable across the file split | Touched: production operator files; no test selector or family row changed | `make test-slice OWNER=app.operator` | All 10/10 app units pass; registry, v1 bytes, closure, Object Store, migration, and Recovery transport remain characterized | Legacy behavior remains intentional until SL-06 | Replace transitional selectors only when adopted assertions change |
| 2026-08-06T18:29:24Z | Codex / SL-06 execution | Transitional Collaboration/Object Store assertions are replaced by adopted conformance and one real v2 process row | Touched: app/module/process tests and `app.operator` family manifest; generated accounting | Make-owned row IDs; five focused/service-backed slices; Harness contract | Cases 1-18, typed-only Object Store, exact registry usage, closure, timeout, concurrency, journal attribution, and commit-unknown pass | None | Remove the temporary crosswalk only after final reconciliation is recorded |
| 2026-08-06T18:57:24Z | Codex / SL-07 completion | Final active rows have one primary owner and exact nonoverlapping selectors; the temporary crosswalk is removed | Touched: obsolete app test helpers, validator test wrapper, final reconciliation report, and tracker | App slice 11/11; backend unit 59/59; final `test-fast` 348/348; `check` 716/716 | Cases 1-18 and every target owner/postcondition remain covered; no runtime alias or mixed selector remains | None | Use section 11 active-row report for future owner discovery |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T15:34:42Z | Codex / operator-app planning | Recovery remains local/config/OS-authorized; Object Store and Collaboration preserve current errors | Touched: this tracker only. Inspected: Core 04, Recovery CLI, migration, Object Store, Collaboration | Repository search and exact-file inspection | No new authorization surface found | RB-002; failure characterization gaps | Add characterization; require owner authorization for behavior changes |
| 2026-08-06T16:37:45Z | Codex / NLSpec tracker revision | Collaboration local authority is allocated to proposed Core 04 ownership; current raw logger output is explicitly non-contractual | Touched: this tracker only. Inspected: live handler/logger/config defaults/service resource closure | Exact source inspection | CRQ-009/010 close the required owner decisions and negative proof | RB-002 | Adopt safe output and authorization before changing failures or exits |
| 2026-08-06T17:13:44Z | Codex / SL-01 execution | Legacy raw-error disclosure is proven with injected config, Postgres, transaction, and commit markers; no marker is treated as acceptable v2 behavior | Touched: Collaboration adapter/service tests and tracker | Focused app slice and service-backed Collaboration slice | Resource closure/rollback is explicit; unsafe stderr assertions are labeled for replacement | RB-002 | Adopt the closed v2 error/redaction/exit registry in Core 01/Core 04 |
| 2026-08-06T17:23:46Z | Codex / SL-02 execution | Core 04 now owns local authority, literal path trust, resource closure, secret-safe fixed diagnostics, raw non-public journaling, and negative surfaces | Touched: Core 04 plus cross-owner rows | AC-535/536 trace review | Private journal limits and forbidden-value set are closed | Runtime remains legacy until SL-06 | Project the closed registries in SL-03, then implement only after Harness migration |
| 2026-08-06T17:34:36Z | Codex / SL-03 execution | v2 registry closes every safe code/reason/exit pair and explicitly disables v1 reader, aliases, and dual output | Touched: Collaboration schema/registry/fixtures/validators | Backend unit plus generator validation | Invalid code/reason, extra member, v1 schema, parser alias, and inconsistent terminal results reject | Runtime remains legacy until SL-06 | Ensure Harness conformance rows reference v2 only |
| 2026-08-06T17:50:02Z | Codex / SL-04 execution | Ownership changed without normalizing unsafe legacy behavior into a requirement | Touched: exact Collaboration/Object Store transport selectors and semantic-owner rows | Focused/service-backed slices and crosswalk audit | Current raw-error/legacy parser assertions remain explicitly transitional; durable semantic evidence is owner-scoped | SL-06 must replace the transitional Collaboration row and Object Store classifier assertions | Preserve behavior exactly in SL-05; adopt safe behavior only in SL-06 |
| 2026-08-06T17:56:16Z | Codex / SL-05 execution | Legacy v1 and string matching were moved without alteration and remain visibly isolated | Touched: `operator_collaboration.go`, `operator_object_store.go`, shared transport file | App slice and exact diff review | No accidental early security/grammar change occurred | SL-06 still owns the intentional v2 break and typed-only mapping | Implement adopted redaction, grammar, timeout, and journaling next |
| 2026-08-06T18:29:24Z | Codex / SL-06 execution | Strict lexical admission, fixed safe failures, bounded cancellation/timeout, atomic private journaling, and typed-only Object Store mapping are live | Touched: Collaboration/Object Store executors, Collaboration service, tests, and guide | Forbidden-value fixtures; writer failure; timeout; rollback; five owner slices | Ordinary Collaboration stderr is empty; raw dependencies never classify Object Store failures; journal/state failure is atomic | None | Run broad security/static gates in SL-07 |
| 2026-08-06T18:57:24Z | Codex / SL-07 completion | Final broad security/static evidence confirms the adopted local-only and secret-safe behavior | Inspected: executor failure maps, transaction/journal service, negative surfaces, and full diff | `make check` 716/716; retained-run finalization | Typed-only Object Store mapping, fixed Collaboration failures, rollback, concurrency, cancellation, and redaction all pass | None | Treat any future public or remote operator surface as new owner work |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T15:34:42Z | Codex / operator-app planning | Initial tracker ready; no production refactor performed | Touched: this tracker only. Inspected evidence summarized in sections 1-5 | Help/task/owner discovery and four planning checks | Workstreams and behavior-preserving slices sequenced | RB-001, RB-002, RB-003 | Obtain Recovery authority and add characterization |
| 2026-08-06T16:37:45Z | Codex / NLSpec tracker revision | Planning gaps now have exact authority, interface, identity, sequencing, and acceptance rules | Touched: this tracker only. Inspected all material named above | Read-only reconciliation and command discovery; tracker validation follows the write | RB-003 is decision-complete; structural refactor remains gated | RB-001, RB-002 | Start only SL-00/SL-01 under explicit test authority; do not start source split |
| 2026-08-06T17:00:34Z | Codex / SL-00 execution | SL-00 completed and the tracker now records active user authority and per-slice status | Touched: this tracker only | `make lint-markdown`; `git diff --check`; status and structural inspection | Pre-checkpoint lint passed at `.cartulary/test-results/20260806T170007Z-p1356769`; post-checkpoint lint follows | RB-001, RB-002 | Start SL-01 only after the post-checkpoint integrity gate passes |
| 2026-08-06T17:13:44Z | Codex / SL-01 execution | SL-01 completed with product behavior unchanged and unsafe legacy evidence isolated by naming/comments | Touched: four test files and this tracker | Owner discovery; format; focused/service-backed slices; `git diff --check`; status; Markdown checkpoint follows | Current selectors, run roots, failure diagnosis, closure expectations, and next slice are recorded | RB-001 and RB-002 must close in SL-02; Harness ownership remains SL-04 work | Run the SL-01 Markdown/diff/status checkpoint, then start SL-02 only if it passes |
| 2026-08-06T17:23:46Z | Codex / SL-02 execution | SL-02 closes both owner blockers without changing downstream behavior prematurely | Touched: five owner/vocabulary documents and this tracker | `make lint-markdown`; requirement/acceptance trace search; `git diff --check`; status | Owner adoption passes; tracker records expected downstream drift | SL-03 projection gates, then SL-04 Harness gates, remain | Run the SL-02 tracker integrity checkpoint, then start SL-03 only if it passes |
| 2026-08-06T17:34:36Z | Codex / SL-03 execution | SL-03 closes typed projection drift without hand-editing generated files | Touched: complete authored/generated set described above and this tracker | Required generation gates, backend unit, exact reconciliation, diff/status; Markdown checkpoint follows | Recovery and Collaboration projection exits are met | Harness ownership/crosswalk is the next dependency | Run SL-03 Markdown/diff/status checkpoint, then start SL-04 only if it passes |
| 2026-08-06T17:50:02Z | Codex / SL-04 execution | SL-04 closes Harness ownership and selector-accounting risk with an explicit temporary crosswalk | Touched: split tests, four authored family manifests, generated accounting, and this tracker | Format; generate; five focused and five service-backed owner slices; explicit full-tier rows; Harness/shape/drift/policy gates | OT-009 is DONE; SL-05 is unblocked; all recorded gates pass | Legacy runtime behavior remains by design until SL-06 | Run the SL-04 Markdown/diff/status checkpoint, then start SL-05 only if it passes |
| 2026-08-06T17:56:16Z | Codex / SL-05 execution | SL-05 completes the behavior-neutral file split and records every added file and planned executor port | Touched: operator source layout, section 2 inventory, executor map, and this tracker | Format; focused app slice; boundary; build; Markdown checkpoint follows | All source-movement gates pass; SL-06 is unblocked | Adopted behavior is not yet implemented | Run the SL-05 Markdown/diff/status checkpoint, then start SL-06 only if it passes |
| 2026-08-06T18:29:24Z | Codex / SL-06 execution | SL-06 completes the intentional v2 behavior cutover and narrow-executor implementation | Touched: complete authorized SL-06 source/test/guide/Harness set and this tracker | All required narrow owner, generation, Harness, boundary, build, diff/status gates; Markdown checkpoint follows | No product or owner blocker remains; two related failed roots and their corrections are retained in section 8 | Broad `test-fast`/`check` and temporary cleanup remain | Run the SL-06 Markdown/diff/status checkpoint, then start SL-07 only if it passes |
| 2026-08-06T18:57:24Z | Codex / SL-07 completion | Remediation and handoff are complete with retained current-worktree evidence | Touched: final cleanup and this controlling tracker; inspected all authorized files | Pre-broad `agent-finalize`; final `test-fast`; `check`; retained-run `agent-finalize`; integrity checkpoint | All OA-AC criteria pass; environmental failed roots and successful corrections are retained in section 8 | None; shared `/tmp` remains externally space-constrained, so final runs used `/var/tmp/ct-op` and task-local caches | No next remediation slice; retain the successful run roots for review |
| 2026-08-06T20:59:24Z | Codex / PC planning | Section 13 identifies removable private residue without reopening completed product behavior | Touched: this tracker only; inspected the live operator package and active Harness ownership | Read-only reachability and stale-surface scans; `make task-guide ROLE=module-author OWNER=app.operator`; `make explain-test-owner OWNER=app.operator`; tracker validation | Owner-required commands are retained; five uncataloged tests, stale ordinal coverage, dead fields, test-only exports, and the broad Recovery CLI path have closed dispositions | PC execution has not been authorized in this document-only step | Begin with read-only PC-00 reconciliation only after a later explicit implementation directive |

## 11. Resolved Questions and Blockers

### Resolved owner blockers and downstream gates

| ID | Blocker | Exact conflict or missing authority | Required closure evidence | Status |
| --- | --- | --- | --- | --- |
| RB-001 | Recovery authored-table identity/count | Historical conflict: Core 01 said 109 while Extensions and the live exact set required 110; Core 01 REQ-01-647 now adopts 110/82 with Goose separate | Registry/schema/generation and exact-set/Goose negative evidence now agree; runtime already used 110/82 | **CLOSED: owner and projection reconciled** |
| RB-002 | Collaboration requeue normative authority | Historical gap: no owner contained the full transition, transport, or local-security contract; REQ-00-068, REQ-01-655, REQ-03-307, and REQ-04-151 now allocate and define it | v2-only schema/registry/fixtures are generated; cases 1-18, process transport, transaction/journal, concurrency, and redaction pass against adopted behavior | **CLOSED: owner, projection, and runtime reconciled** |

### Final Harness reconciliation report: RB-003

RB-003 is closed. The Testing Harness NLSpec supplied the ownership and
immutable-identity rule; SL-04 allocated the changed owner-qualified IDs,
SL-06 replaced transitional behavior rows, and SL-07 removed the temporary
old-to-new crosswalk after recording this active-state report. No retired ID is
a selector source, runtime alias, compatibility reader, or catalog dependency.

| Primary owner | Active row ID | Exact selector or postcondition | Runtime-binary posture |
| --- | --- | --- | --- |
| `app.operator` | `app.operator.integration.collaboration_requeue_v2_process_contract_c1e6d7ba96` | `TestOperatorCollaborationRequeueV2_Process` proves the real v2 process envelope and transaction | Injected `operator`; binary use does not transfer Collaboration semantics |
| `app.operator` | `app.operator.unit.collaboration_requeue_v2_transport_contract_11562cd85f` | Three exact strict-grammar, canonical-admission, delivery, and closure selectors | none |
| `module.collaboration` | `module.collaboration.integration.durable_incident_stream_commits_intents_atomically_s_80128fe153` | `TestDurableIncidentStream_Integration` owns repair validation, reset preservation, journal atomicity, concurrency, cancellation, and commit outcome | none |
| `app.operator` | `app.operator.integration.migration_evidence_transport_and_resource_lifecy_6f6f81a872` | `TestMigrationEvidenceTransport_Integration` owns CLI transport and pool lifecycle | none |
| `app.operator` | `app.operator.unit.migration_evidence_cli_transport_and_redaction_906768574c` | `TestMigrationEvidenceTransport_Unit` owns grammar, output, and redaction | none |
| `platform.postgres` | `platform.postgres.integration.migration_evidence_database_semantics_86249968df` | `TestMigrationEvidenceSemantics_Integration` owns database-backed evidence meaning | none |
| `platform.postgres` | `platform.postgres.support_unit.migration_evidence_projection_semantics_130d030c9a` | `TestMigrationEvidenceSemantics_Unit` owns projected manifest, Goose, and source-audit semantics | none |
| `app.operator` | `app.operator.integration.object_store_init_process_transport_c19d4d6cf0` | `TestMVPObjectStoreInitOperatorTransport` owns process routing/output | Injected `operator`; binary use does not transfer durable Object Store ownership |
| `app.operator` | `app.operator.unit.object_store_init_typed_failure_transport_5cee6a5190` | Three exact deployment-local, redaction, and typed-only failure selectors | none |
| `platform.objectstore` | `platform.objectstore.support_integration.object_store_init_durable_effect_578bf55fdf` | `TestMVPObjectStoreInitOperatorCreatesConfiguredBucket` owns the durable bucket effect | Injected `operator`; owner remains the durable adapter boundary |
| `platform.objectstore` | `platform.objectstore.integration.disconnected_filesystem_root_object_store_setup_0f6ea3351e` | `TestObjectStoreInitialization_Integration` preserves adapter lifecycle evidence | none |
| `app.operator` | `app.operator.unit.operator_registry_eight_paths_and_collaboration_01217009b8` | Four exact registry selectors own eight paths, v2 usage, duplicates, routing, and negative admission | none |
| `module.recovery` | `module.recovery.unit.exact_recovery_state_catalog_and_synthetic_goose_774b28fd25` | Two exact catalog/set/classification and separate-Goose selectors | none |
| `module.recovery` | `module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9` | Five exact Recovery command process selectors; Object Store is excluded | Injected `operator` for the five preserved Recovery paths |
| `app.operator` | `app.operator.unit.recovery_failure_mapping_is_exhaustive_465f2ab0de` | Exact closed Recovery transport failure mapping | none |
| `app.operator` | `app.operator.unit.the_due_verification_timeout_is_passed_to_recove_cbf25983a5` | Exact Recovery due-verification timeout delegation | none |

The SL-06 `make harness-contract` root and the final 716-unit `make check` root
prove that every active selector is admitted exactly once with one primary
owner. Future identities MUST continue to use `make author-test-row-id` with an
owner-qualified family, semantic claim, and exact selector key.

## 12. Binary Completion Criteria

### Acceptance criteria

| Acceptance ID | Binary acceptance condition |
| --- | --- |
| OA-AC-001 | The repository change set stays within the authorized SL-00 through SL-07 owner, contract, implementation, test, guide, Harness, generated, and tracker write sets; the user's pre-existing staged tracker history is preserved. |
| OA-AC-002 | Every normative statement distinguishes adopted owner, current implementation, proposed amendment, refactor rule, and support material. |
| OA-AC-003 | All 14 post-SL-05 target files appear exactly once in section 2 with complete ownership and risk posture. |
| OA-AC-004 | `RunOperatorCLIContext`, the exclusive `cmd/operator` dependency, and all eight command paths are frozen unless later owner authorization names a change. |
| OA-AC-005 | Core 01, Recovery projections, fixture, validator, runtime, and exact-set negative tests all agree on 110 authored catalog identities, 82 required entries, and separate Goose database coverage. |
| OA-AC-006 | The Collaboration interface tables and executable v2 family agree on strict grammar, identifiers, defaults, members/bytes, exits, timeout/cancellation, delivery failure, resources, and atomic mutation effects. |
| OA-AC-007 | Authority placement assigns each semantic, transport, security, Harness, typed-projection, implementation, and guide responsibility exactly once. |
| OA-AC-008 | CRQ-001 through CRQ-011 form a closed owner-adoption checklist and no Collaboration behavior change may precede their adoption. |
| OA-AC-009 | Characterization cases 1-18 each resolve to at least one exact nonoverlapping test selector. |
| OA-AC-010 | Every target test is selected exactly once after test splitting; no mixed postcondition remains under a catch-all owner. |
| OA-AC-011 | The final RB-003 reconciliation records every active target row's primary owner, exact selector/postcondition, and runtime-binary posture. |
| OA-AC-012 | Every owner change used a new immutable owner-qualified ID; the temporary crosswalk is removed and no runtime alias exists. |
| OA-AC-013 | Structural slices retain the exact facade, required exports, eight paths, resource lifetimes, and semantic-owner delegation; only the separately adopted Collaboration v2 and typed Object Store behavior changes. |
| OA-AC-014 | Generated outputs change only through authorized inputs and Make-owned generation; generated-artifact policy passes. |
| OA-AC-015 | Every required focused, service-backed, static, build, generated, and broader validation has an actual successful result and run root before implementation completion. |
| OA-AC-016 | All seven handoff tables preserve prior rows and contain a current row with touched files, commands, result, blockers, and next action. |

### Requirement traceability

| Requirement | Owner artifact or evidence | Workflow | Slice | Validation | Acceptance |
| --- | --- | --- | --- | --- | --- |
| OA-REQ-001 | Tracker/write-scope audit | WF-00 | Documentation task | `git status --short` | OA-AC-001 |
| OA-REQ-002 | Core/Extensions/Harness authority | WF-00, WF-02 | SL-02 | Owner review; Markdown lint | OA-AC-002, OA-AC-005 |
| OA-REQ-003 | `cmd/operator/main.go`, boundary policy | WF-04, WF-05 | SL-05, SL-06 | Boundary check; operator build | OA-AC-004, OA-AC-013 |
| OA-REQ-004 | Target inventory | WF-01 | SL-00 | Structural inventory check | OA-AC-003 |
| OA-REQ-005 | Recovery owners/contracts/runtime | WF-01, WF-02 | SL-00, SL-02, SL-03 | JSON shape; generation/drift; Recovery slices | OA-AC-005, OA-AC-014 |
| OA-REQ-006 | Proposed Core 00/01/03/04 amendments | WF-02 | SL-02 | Owner review | OA-AC-007, OA-AC-008 |
| OA-REQ-007 | Live handler/parser/service | WF-01, WF-03 | SL-01, SL-05, SL-06 | App operator and Collaboration slices | OA-AC-006, OA-AC-013 |
| OA-REQ-008 | CRQ-001..011 | WF-02, WF-03 | SL-01, SL-02, SL-03 | Characterization matrix; owner review | OA-AC-008, OA-AC-009 |
| OA-REQ-009 | Testing Harness owner/ID/crosswalk rules | WF-07 | SL-04 | Five owner and service-backed slices | OA-AC-010 through OA-AC-012 |
| OA-REQ-010 | Characterization cases and target test inventory | WF-03 | SL-01 | Exact selected tests | OA-AC-009, OA-AC-010 |
| OA-REQ-011 | Generated-artifact policy | WF-02, WF-07 | SL-03, SL-04 | Generate, drift, policy | OA-AC-014 |
| OA-REQ-012 | Workflow/slice dependency tables | WF-00 through WF-08 | SL-00 through SL-07 | Handoff dependency audit | OA-AC-013, OA-AC-015 |
| OA-REQ-013 | Validation matrix | WF-08 | SL-07 | Section 8 exact commands | OA-AC-015 |
| OA-REQ-014 | Session log | WF-08 | Every slice | Handoff row audit | OA-AC-016 |
| OA-REQ-015 | Adopted owner changes | WF-02 | SL-02 before SL-03..SL-07 | Owner review and contract validation | OA-AC-008, OA-AC-013 |
| OA-REQ-016 | Module boundary and narrow ports | WF-04, WF-05, WF-06 | SL-05, SL-06 | Boundary/build/focused slices | OA-AC-007, OA-AC-013 |

OA-AC-001 through OA-AC-016 all pass. RB-001 through RB-003 are closed,
OT-009 through OT-012 are `DONE`, and section 8 contains successful retained
evidence for every required focused, service-backed, generated, boundary,
build, static, broad, and handoff gate. Sections 1 through 12 remain the final
handoff for that remediation. Section 13 supersedes only the historical claim
that no later private cleanup iteration is planned; it does not reopen or
weaken any completed acceptance result.

## 13. Planned Operator Legacy Cleanup and Production Hardening

### 13.1 Status, authority, and behavior freeze

This section is a refactor execution ledger, not an adopted product owner. Its
status is `DONE`: PC-00 through PC-05 are complete, including the explicitly
authorized Harness blocker and owner-tier reachability remediation.
The user-authorized implementation MUST execute PC-00 through PC-05 in order
at the baseline recorded in section 1. A slice is not complete until this
tracker is updated and its Markdown, diff, and scope checkpoint passes.

The next iteration has one structural objective: make `internal/app/operator`
a production-ready binary facade with the smallest justified private surface,
no live legacy accommodation, and complete active verification. It MUST preserve
all of the following observable contracts:

- `RunOperatorCLIContext` remains the sole binary-facing Go entry point;
- the five Recovery commands, `migration-evidence capture`, `object-store init`,
  and `collaboration requeue` remain the exact eight command paths;
- command grammar, usage, defaults, output member order and bytes, schema IDs,
  exit codes, timeout and cancellation behavior, redaction, mutation semantics,
  and acquired-resource closure remain unchanged;
- `cmd/operator` continues to import only `internal/app/operator`;
- semantic behavior remains delegated to Recovery, Collaboration, Postgres, and
  Object Store owners; and
- no schema version, migration, command alias, compatibility reader, forwarding
  package, or dual behavior is introduced.

The supported commands are not cleanup candidates. The Testing Harness NLSpec
retains the five Recovery commands while Core 01 requires them, retains
Migration Evidence while its owner consumes the evidence, and retains Object
Store initialization while stand-up packaging requires configured-bucket
initialization. Collaboration requeue remains adopted current-profile behavior.
None of those retirement predicates is satisfied at the execution baseline.
Core 01, Core 03, Core 04, and the Testing Harness NLSpec already close the
behavior and runtime-binary requirements. `docs/domain.md` classifies package
paths, Go symbols, and operator wiring as implementation detail. PC-00 found no
owner contradiction and authorizes no owner or domain amendment.

### 13.2 Inspected cleanup inventory and final disposition

The implementation baseline is HEAD
`846d024b900be4cf8019dcb6cf0c1d938386ff71` with only this user-owned tracker
revision staged. The index MUST remain untouched. The baseline contains 21 Go
test entry points beneath `internal/app/operator`, nine active `app.operator`
Harness rows, and five target test entry points absent from every authored
selector. The following dispositions are closed:

| Candidate | Baseline evidence | Final disposition | Continuing contract |
| --- | --- | --- | --- |
| `operatorRunner.stdout` | Stored during construction and never read | Delete in PC-04 | Writers remain owned by each executor's `operatorTransport`. |
| `operatorCommandDescriptor.Owner` | Checked only for nonempty text and never used for routing, diagnostics, ownership, or evidence | Delete in PC-04 | Harness ownership remains in authored owner manifests; descriptors retain tokens, usage, handlers, and invalid-namespace routing. |
| `operatorCLIResult.command` | Written by two parsers and read only by one test assertion | Delete in PC-04 | Exact command identity remains in the registry and command-local executor. |
| `operatorCLIResult` | Union of Migration Evidence fields and Object Store fields, coupling unrelated parsers | Replace with command-local argument types in PC-04 | Parsers return `(commandArgs, stop, exitCode)`; no generic parser container is added. |
| `defaultMigrationEvidenceManifestPath` | Private alias used only by implementation and tests | Delete in PC-04 | Use `migrationevidence.DefaultManifestPath` directly. |
| Root operator result structs and schema constants | Exported only for same-repository tests; not production composition APIs | Make private in PC-03 | JSON schema IDs and wire shapes remain byte-stable; process tests use strict test-local decoders. |
| `internal/app/operator/recoverycli` import path | Imported by its operator parent and Recovery process tests | Move beneath `internal/app/operator/internal` in PC-03 and delete the old path | No forwarding package, import alias, or duplicate implementation remains. |
| Recovery CLI helper exports | Parsing, timeout, path, error, progress, and mapping helpers exceed production need | Make all private except `Run` and `FailureEvidenceFields` in PC-03 | The parent facade retains exactly its two required composition seams. |
| `TestOperatorCommandRegistry_U_RejectsUnregisteredSixthCommand` | Refers to a stale ordinal and duplicates global unknown-command coverage | Delete in PC-02 and remove its selector | The registry row retains its immutable identity and current postcondition. |
| Five uncataloged Recovery tests in `operator_recovery_test.go` | Current parser/path/envelope value is mixed with legacy spellings | Replace with one current-contract test in PC-02 | Canonical grammar and generic failure behavior become active evidence without historical compatibility vocabulary. |
| Retired top-level-token boundary rule | Machine policy rejects production reintroduction of `backup-metadata` | Keep | A structural absence guard has continuing value and adds no runtime compatibility path. |
| Black-box process rows lack the `operator` runtime-binary edge | Baseline owner slice failed because `CARTULARY_OPERATOR_BIN` named a missing file | Move both to `app.operator.process` and map that family to `operator` in PC-01 | Scheduler-produced binary provenance replaces workspace-order dependence. |
| Object Store process row claims a Postgres fixture | The test mutates only an isolated Object Store bucket | Use `object_store_namespace` in PC-01 | Fixture ownership matches the durable effect and avoids a false database dependency. |

The five currently uncataloged entry points are
`TestOperatorRecoveryParserAcceptsCanonicalCommands`,
`TestOperatorRecoveryParserRejectsLegacyRecoverySurface`,
`TestOperatorRecoveryParserLeavesRetiredTopLevelNamesForRegistryUsage`,
`TestOperatorRecoveryParserRejectsUnsafeTargetPaths`, and
`TestOperatorRecoveryCLIEmitsSingleFailureEnvelopeForLegacyCommand`.
Their names MUST disappear. PC-02 MUST preserve only their current-contract
assertions: five canonical operations, literal absolute target-path validation,
generic unknown Recovery subcommand projection, and the single JSON failure
envelope. The invalid-path matrix includes relative, `~`, shell-variable, NUL,
and lexical `.` and `..` forms. Unknown flags MUST be covered generically, not
by enumerating retired flag spellings.

The final intended Go exposure is:

| Package | Permitted exported surface after PC-03 |
| --- | --- |
| `internal/app/operator` | `RunOperatorCLIContext` only |
| `internal/app/operator/internal/recoverycli` | `Run` and `FailureEvidenceFields` only; wire DTOs and parser helpers remain private |
| retired `internal/app/operator/recoverycli` | Package absent; zero imports and no forwarding alias |

The baseline `make test-slice OWNER=app.operator` run at
`.cartulary/test-results/20260806T210808Z-p2685286` failed only the two process
rows because `/home/jochi/code/cartulary/operator` did not exist. The failure is
related to the PC-01 Harness projection gap and is retained as diagnostic
evidence, not accepted as a product failure.

### 13.3 Ordered implementation slices

The status vocabulary is `PLANNED`, `READY`, `IN_PROGRESS`, `BLOCKED`, and
`DONE`. Slices execute in exact order and this tracker MUST be checkpointed
after each completed or blocked slice.

| Slice | Status | Depends on | Exact change | Authorized write set | Validation and completion gate |
| --- | --- | --- | --- | --- | --- |
| PC-00 | DONE | Explicit implementation directive | Revalidated HEAD/index posture, adopted owners, callers, exports, 21 tests, nine rows, five uncataloged tests, retirement predicates, and the missing operator-binary edge; adopted this six-slice ledger. | This tracker only; all other inspection was read-only. | Baseline and owner reports agree; the staged tracker is preserved; PC-01 is `READY`. |
| PC-01 | DONE | PC-00 | Replaced the two process row identities, added `app.operator.process -> operator`, corrected the Object Store fixture capability, repaired graph-runner machine-state propagation under the explicitly expanded write set, and regenerated topology/task-surface outputs. | `tools/test_families/app.operator.json`, `tools/execution_topology_manifest.json`, their Make-generated outputs, the task-surface renderer, its focused regression test and generated Make output, and this tracker. | Exact and omitted owner selections build the missing operator artifact and pass both process rows; generated-artifact owner, Harness, JSON, generation-drift, and policy gates pass. |
| PC-02 | DONE | PC-01 | Deleted the stale ordinal and five legacy-named entry points; added one current parser/path/invalid-envelope test; moved exhaustive failure mapping beside Recovery CLI; updated selectors/accounting. | Operator registry/Recovery tests, Recovery CLI tests, app operator family manifest, generated accounting, and this tracker. | The intermediate inventory is 16 selected entry points and projects to the required final 17 after PC-03 adds the export guard; registry/failure/due IDs remain stable and the new parser row is active. |
| PC-03 | DONE | PC-02 | Moved Recovery CLI beneath nested `internal`, exposed only `Run` and `FailureEvidenceFields`, privatized root wire types, introduced strict test-local process DTOs, updated boundary policy, and added the export-surface guard. | Operator/Recovery CLI source and tests, process tests, backend boundary input, app manifest, generated accounting, and this tracker. | Old path/imports are absent; the exact 17-test and 11-row inventory reconciles; black-box bytes, boundary, build, process, owner, Harness, JSON, generation-drift, and generated-policy evidence pass. |
| PC-04 | DONE | PC-03 | Removed dead runner/descriptor/parser fields and the manifest alias; introduced command-local Migration Evidence and Object Store args with explicit stop/exit returns; strengthened incomplete-descriptor tests. | Operator facade, registry, transport, Migration Evidence, Object Store, directly affected tests, and this tracker. | Dead symbols are absent; only the two command-local parser result types remain; help, invalid, and success stop/exit semantics pass; affected app, Postgres, Object Store, and operator-build evidence passes. |
| PC-05 | DONE | PC-04 | Reconciled selectors, generated outputs, boundaries, the 11-row and 17-test inventory, exact acceptance rows, all focused/service-backed owners, and every required broad gate. The resumed slice asserted the resource-conflict diagnostics, retained and activated current run-step coverage while removing only its retired Go-helper cases, activated the current Vitest and Playwright wrappers with external scratch, removed the redundant semantic-identity wrapper, and added structural orphan-check rejection. | Authorized authored/generated inputs changed by PC-01 through PC-04; Harness execution-wrapper tests and scratch support; the task-surface owner, validator, focused validator test, and Make-generated projections; obsolete run-go cases and redundant semantic-wrapper deletion; and this tracker. | The current owner and generated projections agree; all retained Harness checks are tier-reachable; focused gates, `lint-shell`, preflight finalization, `test-fast`, all 716 `check` units, retained finalization, and the handoff checkpoint pass. |

PC-02 retains the existing failure-mapping and due-timeout row IDs because
their owner and semantic postconditions do not change. The current registry row
also retains its ID after the redundant selector is removed because it still
proves the same eight-route and routing postcondition. The new active
parser/transport postcondition is
`app.operator.unit.recovery_cli_parser_and_invalid_invocation_trans_b3caf1f612`.
The export guard is
`app.operator.unit.operator_facade_and_nested_recovery_cli_expose_o_0b8f75c21c`.
Neither has a predecessor.

The two process-family changes require new row IDs because `row_id` is prefixed
by immutable `family_id`. The documentation-only crosswalk is:

| Previous row | Replacement row |
| --- | --- |
| `app.operator.integration.collaboration_requeue_v2_process_contract_c1e6d7ba96` | `app.operator.process.collaboration_requeue_v2_process_contract_4ce8fef4eb` |
| `app.operator.integration.object_store_init_process_transport_c19d4d6cf0` | `app.operator.process.object_store_init_process_transport_400591b90d` |

The old IDs are removed from executable inputs. This table is review evidence,
not a runtime alias, selector source, or compatibility reader.

### 13.4 Verification and evidence plan

Before editing each owner input, run its Make-owned task guidance. The later
implementation MUST use these gates in order and record actual run roots:

| Stage | Exact commands | Required result |
| --- | --- | --- |
| Owner and selector discovery | `make task-guide ROLE=module-author OWNER=<owner>` and `make explain-test-owner OWNER=<owner>` for `app.operator`, `module.recovery`, `module.collaboration`, `platform.postgres`, and `platform.objectstore` | Current rows and service-backed posture recorded before selector changes. |
| Identifier allocation | `make author-test-row-id` with the approved family, claim, and final selector keys recorded above | Four exact new IDs reproduce; no ID is allocated for deleted legacy-only tests. |
| Focused owner evidence | `make test-slice OWNER=<owner>` for all five owners above | Every affected unit/integration selector passes after PC-01 through PC-03. |
| Service-backed evidence | `make service-backed-test-slice OWNER=<owner>` for all five owners above | Resource, process, durable-effect, and semantic-owner evidence remains unchanged. |
| Static and build | `make backend-module-boundary-check`; `make build-operator` | Nested internal boundary and exclusive binary composition pass; production operator builds. |
| Harness and projections | `make harness-contract`; `make json-shape-check`; `make generate`; `make generate-drift`; `make generated-artifact-policy-check` | Authored selector and boundary changes project cleanly; generated files are never hand-edited. |
| Broad completion | `make agent-finalize`; `make test-fast`; `make check` | Final current-worktree evidence succeeds before PC-05 becomes `DONE`; retained-run maintenance uses `RESULTS_DIR` for the successful full run. |
| Tracker integrity | `make lint-markdown`; `git diff --check`; `git status --short` | Markdown, whitespace, and write-scope integrity pass after every checkpoint. |

No browser, frontend, HTTP, WebSocket, data-migration, or schema-version test is
added. A failure outside the authorized write set MUST be recorded with its run
root and relationship assessment; it MUST NOT broaden the iteration silently.

### 13.5 Binary acceptance and next action

| Acceptance ID | Required terminal condition |
| --- | --- |
| PC-AC-001 | `RunOperatorCLIContext` is the only export from `internal/app/operator`; `cmd/operator` remains its only production composition caller. |
| PC-AC-002 | The old Recovery CLI package and import path are absent, and no forwarding package, type alias, compatibility reader, or duplicate implementation remains. |
| PC-AC-003 | All eight commands retain exact grammar, usage, output bytes and schemas, failure/exit mapping, cancellation/timeouts, redaction, mutation semantics, and resource closure. |
| PC-AC-004 | `operatorRunner.stdout`, descriptor owner metadata, the mixed parser result and its dead command field, and the manifest-path alias are absent. |
| PC-AC-005 | Migration Evidence and Object Store parsers use separate command-local argument types; no generic plugin framework, dependency container, or shared union parser replaces the deleted structure. |
| PC-AC-006 | Every retained Go test beneath the target is selected exactly once by an active owner row, and no stale ordinal or legacy-specific test name remains. |
| PC-AC-007 | Current Recovery parser, absolute target-path, generic unknown-command/envelope, exhaustive failure mapping, and due-timeout behavior have active exact selectors. |
| PC-AC-008 | Continuing owner/postcondition rows retain their immutable IDs; four exact new IDs cover the process-family replacements and new parser/export postconditions; the documentation crosswalk is not executable and no runtime alias exists. |
| PC-AC-009 | The retired-token production boundary guard remains active while historical command identifiers are absent from live production code and active test selectors. |
| PC-AC-010 | Focused, service-backed, boundary, build, Harness, generation, broad, Markdown, diff, and status gates all have recorded successful evidence before PC-05 is marked `DONE`. |

Additional terminal conditions are:

- both black-box process rows receive a scheduler-produced `operator` binary;
- `app.operator` contains 11 active rows and the target contains 17 Go test
  entry points, each selected exactly once;
- the Object Store process row uses `object_store_namespace` and no false
  Postgres fixture dependency; and
- the export-surface AST guard admits only the root `RunOperatorCLIContext` and
  nested `Run` and `FailureEvidenceFields` seams.

### 13.6 Production-hardening execution log

| Timestamp | Slice | Status | Files and substantive work | Commands and evidence | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-08-06T21:18:38Z | PC-00 | DONE | Updated this tracker only; preserved staged baseline; reconciled owners, callers, exports, tests, selectors, runtime-binary topology, exact new IDs, and the process-row crosswalk | `git status --porcelain=v2`; owner task guides/explanations; exact `rg` inventories; baseline failed owner slice `.cartulary/test-results/20260806T210808Z-p2685286`; checkpoint validation follows | None; the failed owner slice is the confirmed PC-01 Harness gap | Execute PC-01 only |
| 2026-08-06T21:23:34Z | PC-01 | BLOCKED | Replaced both process row IDs and family IDs, mapped `app.operator.process` to the authored `operator` runtime binary, changed the Object Store fixture from `postgres_dedicated` to `object_store_namespace`, and regenerated the topology render index. The staged tracker baseline remains untouched. | First `make generate` failed on row ordering at `.cartulary/test-results/20260806T211931Z-p2701475`; the corrected rerun passed at `.cartulary/test-results/20260806T212027Z-p2704296`. With no pre-existing `operator` file, the exact two-row slice built the binary and passed Collaboration, but failed Object Store fixture acquisition at `.cartulary/test-results/20260806T212047Z-p2706713`. Checkpoint `make lint-markdown` passed at `.cartulary/test-results/20260806T212400Z-p2729930`; staged and unstaged `git diff --check` and scoped status inspection passed. | Related Harness defect: graph-runner recipes strip the three Make-owned Go machine-state variables, while `object_store_namespace` requires them to compile its run-scoped proxy. Repairing the task-surface renderer/generated Make output and adding Harness regression evidence is outside the PC-01 authorized write set. | Obtain authority to expand PC-01 to the task-surface renderer, its tests, and generated task-surface output; repair propagation, rerun PC-01 gates, and do not begin PC-02 beforehand. |
| 2026-08-06T23:46:00Z | PC-01 | IN_PROGRESS | User explicitly expanded the slice write set to the task-surface renderer, focused renderer regression tests, and generated task-surface Make output; no PC-02 mutation began. | Authorization recorded in this ledger before the resumed implementation. | None | Repair graph-runner machine-state propagation and complete all PC-01 gates. |
| 2026-08-06T23:48:42Z | PC-01 | DONE | Added Make-owned machine-state forwarding to every graph-runner recipe after public-input stripping; added a renderer regression assertion; regenerated task-surface and topology outputs; retained new process identities, operator binary mapping, and correct Object Store fixture. | `make generate` passed at `.cartulary/test-results/20260806T234654Z-p2763340`. From an absent `operator` artifact, the exact two-row slice passed at `.cartulary/test-results/20260806T234723Z-p2765734`; omitted-row `app.operator` passed at `.cartulary/test-results/20260806T234753Z-p2785430`. `harness.generated_artifacts` passed at `.cartulary/test-results/20260806T234822Z-p2805493`; `harness-contract`, `json-shape-check`, and `generate-drift` passed respectively at `.cartulary/test-results/20260806T234822Z-p2805726`, `.cartulary/test-results/20260806T234822Z-p2805428`, and `.cartulary/test-results/20260806T234822Z-p2805388`; generated-artifact policy passed at `.cartulary/test-results/20260806T234834Z-p2811464`. | None; the earlier related fixture failure remains recorded above. | Run the PC-01 tracker checkpoint, then execute PC-02 only. |
| 2026-08-06T23:53:18Z | PC-02 | DONE | Removed the stale sixth-command test and the five uncataloged legacy-oriented Recovery entry points; replaced their continuing value with `TestRecoveryCLIParserAndInvalidInvocationContract`; moved the exhaustive failure mapping unchanged beside Recovery CLI; converted registry negatives to generic unknown inputs; retained registry, failure, and due row IDs; added the new parser row and regenerated accounting. | `make format` passed at `.cartulary/test-results/20260806T235158Z-p2817848`; `make generate` passed at `.cartulary/test-results/20260806T235201Z-p2820813`; exact parser/registry/failure/due rows passed at `.cartulary/test-results/20260806T235213Z-p2823146`; omitted `app.operator` and `module.recovery` slices passed at `.cartulary/test-results/20260806T235227Z-p2824158` and `.cartulary/test-results/20260806T235227Z-p2824175`; Harness contract and generation drift passed at `.cartulary/test-results/20260806T235227Z-p2824367` and `.cartulary/test-results/20260806T235227Z-p2824077`. Inventory inspection found 16 intermediate entry points with no stale ordinal, legacy-named test, or historical negative identifier in active operator tests/selectors. | None | Run the PC-02 tracker checkpoint, then execute PC-03 only. |
| 2026-08-07T00:03:25Z | PC-03 | DONE | Moved Recovery CLI to `internal/app/operator/internal/recoverycli`; replaced its exported runner and DTO/helper surface with free `Run`, `FailureEvidenceFields`, and private implementation types; privatized root Object Store and Collaboration wire types; removed process-test imports of both production transport packages in favor of closed local DTOs, literal schema IDs, and strict decoders; added retired-import boundary policy and updated runtime scan paths; added the AST export allowlist with a negative fixture and its immutable row. | Initial `make build-operator` exposed a local type-shadowing compile error at `.cartulary/test-results/20260806T235931Z-p2890145`; the repaired build passed at `.cartulary/test-results/20260807T000004Z-p2904141`. `make generate` passed at `.cartulary/test-results/20260807T000028Z-p2915028`; exact export/parser/failure/due rows passed at `.cartulary/test-results/20260807T000042Z-p2917385`. The first exact process rerun exposed a test-local Object Store schema literal typo at `.cartulary/test-results/20260807T000050Z-p2918263`; the corrected rerun passed at `.cartulary/test-results/20260807T000140Z-p2941339`. Boundary passed at `.cartulary/test-results/20260807T000210Z-p2961164`; omitted `app.operator` and `module.recovery` passed at `.cartulary/test-results/20260807T000229Z-p2961823` and `.cartulary/test-results/20260807T000229Z-p2961822`; Harness and generation drift passed at `.cartulary/test-results/20260807T000229Z-p2962046` and `.cartulary/test-results/20260807T000229Z-p2961744`; JSON and generated-policy gates passed at `.cartulary/test-results/20260807T000317Z-p3019825` and `.cartulary/test-results/20260807T000317Z-p3019816`. | None; both related local implementation/test failures were repaired without owner or wire changes. | Run the PC-03 tracker checkpoint, then execute PC-04 only. |
| 2026-08-07T00:07:36Z | PC-04 | DONE | Removed `operatorRunner.stdout`, `operatorCommandDescriptor.Owner`, `operatorCLIResult` and its dead command field, and the manifest-path alias; introduced `migrationEvidenceCaptureArgs` and `objectStoreInitArgs` with explicit `(args, stop, exitCode)` parse returns; used `migrationevidence.DefaultManifestPath` directly; expanded the registry test across empty tokens, invalid tokens, blank usage, and nil handlers; added direct help/invalid/success parser assertions. | `make format` passed at `.cartulary/test-results/20260807T000609Z-p3023431`. The first exact three-row run found a test-local variable redeclaration at `.cartulary/test-results/20260807T000621Z-p3026479`; after repair, exact Migration Evidence, Object Store, and registry rows passed at `.cartulary/test-results/20260807T000648Z-p3030162`. Omitted `app.operator`, `platform.postgres`, and `platform.objectstore` passed at `.cartulary/test-results/20260807T000706Z-p3031109`, `.cartulary/test-results/20260807T000706Z-p3031113`, and `.cartulary/test-results/20260807T000706Z-p3031108`; `make build-operator` passed at `.cartulary/test-results/20260807T000706Z-p3031529`. Dead-symbol searches were empty and found only the two command-local parser types/signatures. | None; the related test compile failure was repaired without product behavior changes. | Run the PC-04 tracker checkpoint, then execute PC-05 only. |
| 2026-08-07T00:19:36Z | PC-05 | BLOCKED | Reconciled all 11 active `app.operator` rows and 17 target tests; reran the seven exact process/parser/export/registry/failure/due rows; completed all five focused and service-backed owner selections; completed static, build, Harness, JSON, generation, drift, policy, preflight-finalization, and fast-suite gates. | Exact acceptance rows passed at `.cartulary/test-results/20260807T000833Z-p3067976`. Focused `app.operator`, `module.recovery`, `module.collaboration`, `platform.postgres`, and `platform.objectstore` passed respectively at `.cartulary/test-results/20260807T000856Z-p3087712`, `.cartulary/test-results/20260807T000856Z-p3087723`, `.cartulary/test-results/20260807T000856Z-p3087739`, `.cartulary/test-results/20260807T000856Z-p3087758`, and `.cartulary/test-results/20260807T000856Z-p3087774`; their service-backed selections passed at `.cartulary/test-results/20260807T001040Z-p3173353`, `.cartulary/test-results/20260807T001040Z-p3173339`, `.cartulary/test-results/20260807T001040Z-p3173340`, `.cartulary/test-results/20260807T001040Z-p3173369`, and `.cartulary/test-results/20260807T001040Z-p3173362`. `generate` passed at `.cartulary/test-results/20260807T001210Z-p3254168`; boundary, build, Harness, JSON, drift, and policy passed at `.cartulary/test-results/20260807T001227Z-p3257012`, `.cartulary/test-results/20260807T001227Z-p3257382`, `.cartulary/test-results/20260807T001227Z-p3257099`, `.cartulary/test-results/20260807T001226Z-p3256702`, `.cartulary/test-results/20260807T001226Z-p3256698`, and `.cartulary/test-results/20260807T001226Z-p3256729`. `agent-finalize` and all 348 `test-fast` units passed at `.cartulary/test-results/20260807T001243Z-p3271839` and `.cartulary/test-results/20260807T001259Z-p3274508`. Full `check` failed one of 716 units at `.cartulary/test-results/20260807T001451Z-p3323661`; independent `lint-shell` reproduced it at `.cartulary/test-results/20260807T001912Z-p3478277`. | Unrelated pre-existing `SC2034`: unchanged `tools/harness/tests/test-run-step.sh` assigns `capacity_target_output` but never consumes it. Retained-evidence finalization was not run because no successful full-check root exists. Browser suites remain skipped because this target has no browser surface. | Obtain authority for the narrow Harness test repair; do not mark PC-05 done or publish final handoff before the broad check and retained finalization pass. |
| 2026-08-07T00:31:49Z | PC-05 | IN_PROGRESS | The user authorized the capacity-output assertion, external scratch migration for the three unique execution-wrapper tests, activation of those tests in owner-controlled tiers, deletion of the redundant semantic-identity smoke wrapper, and an authored task-surface orphan-check guard with generated projections. | Reconfirmed clean HEAD `d2af46097d7bce69554b988bd1bd666f28985bb6`, empty `git status --short`, and the unchanged SC2034 site before mutation. `docs/domain.md` and the adopted Testing Harness NLSpec remain unchanged because their authority and behavior are already complete. | None | Implement only the expanded Harness repair, generate from authored inputs, and run the focused-to-broad validation ladder before completing PC-05. |
| 2026-08-07T01:01:57Z | PC-05 | DONE | Asserted `failure_class=infra` and `reason=resource_conflict` on the captured quiet capacity failure while retaining the writer-status and public exit-4 checks. Retained the run-step monolith's current execution, output, taxonomy, summary, timing, fixture, and diagnostic coverage; removed only cases for the already-removed Go-step helper; updated helper evidence to the canonical target-summary profile; and contained every fixture through shared Harness scratch. Activated run-step, Vitest, and Playwright wrappers in `execution`, `extended`, and `full`, removed the duplicate semantic-identity wrapper, added exact orphan-check validation and regression evidence, and regenerated task-surface/topology projections through Make. No adopted owner, domain, product, dependency, database, frontend, or public interface changed. | Final `make generate` passed at `.cartulary/test-results/20260807T005314Z-p3627526`. All three execution wrappers passed with default and allowed external scratch at `.cartulary/test-results/20260807T005741Z-p3654433` and `.cartulary/test-results/20260807T005806Z-p3661623`. `lint-shell`, Harness contract, JSON shape, generation drift, and generated policy passed at `.cartulary/test-results/20260807T005909Z-p3685981`, `.cartulary/test-results/20260807T005909Z-p3685980`, `.cartulary/test-results/20260807T005909Z-p3685640`, `.cartulary/test-results/20260807T005909Z-p3685645`, and `.cartulary/test-results/20260807T005909Z-p3685661`. Preflight finalization made zero generated updates at `.cartulary/test-results/20260807T005926Z-p3690440`; `test-fast` passed 348/348 at `.cartulary/test-results/20260807T005943Z-p3693108`; `check` passed 716/716 at `.cartulary/test-results/20260807T010006Z-p3693708`; retained finalization validated that latest root and made zero generated updates at `.cartulary/test-results/20260807T010118Z-p3772083`; the final Markdown checkpoint passed at `.cartulary/test-results/20260807T010236Z-p3775154`. | The first extended run exposed and then passed after repair of a related multiline generated-recipe test expectation at `.cartulary/test-results/20260807T004238Z-p3513327`. The final extended run at `.cartulary/test-results/20260807T005829Z-p3668925` proves all three activated wrappers and the orphan regression pass, then records unrelated pre-existing release-task-surface and browser work-graph fixture assertions. Those defects did not expand this remediation, and standalone browser suites remain skipped because the operator target has no browser surface. | No further PC slice is authorized or required; preserve the current-profile Harness tiers and orphan guard in future task-surface changes. |

Sections 1 through 12 remain closed historical evidence. PC-00 through PC-05
are complete, the successful full-check root is retained, and no further
production-hardening slice is authorized by this ledger.
