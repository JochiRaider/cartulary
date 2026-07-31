# Cartulary OpenAPI and Public HTTP Contract Refactor Tracker

## Baseline

- Created: 2026-07-25
- Baseline commit: `53f4553150117d09535794e59a8f500485bdb94c`
- Historical evidence:
  `docs/handoffs/cartulary-openapi-module-refactor-tracker.md`
- Canonical artifact: `contracts/openapi/cartulary.openapi.yaml`
- Canonical version: `1.0.0`
- Canonical SHA-256:
  `de976721d98d6f6e73747dfffa3a0a3af21f6592037abb3d6ee8ad1b86195f0f`
- Canonical size: 441,211 bytes
- Inventory: 93 paths, 110 operations, 276 schemas, 23 operation tags,
  353 unique internal references, and five used security schemes.
- Source topology: 70 manifest-selected fragments across 18 owners.
- Authority posture: executable requirements and typed machine contracts are
  upstream; generated artifacts are downstream; Markdown is explanatory only.
- Last released public contract: `1.0.0`.

The historical remediation tracker is immutable. This tracker records the
successor effort only.

The current release candidate is OpenAPI `2.0.0`, SHA-256
`48581e497a570b99180a41e5df69d55d07ec6c45520a9625d362e29259c464cf`,
449,196 bytes, with 94 paths, 111 operations, 282 schemas, 23 operation tags,
359 unique internal references, and five used security schemes. The released
pointer remains `1.0.0` until publication succeeds.

## Status vocabulary

- `TODO`: not started.
- `IN_PROGRESS`: implementation or required validation is active.
- `BLOCKED`: a required dependency, decision, or validation failed.
- `DONE`: implementation and all required checkpoint validation passed.

A workstream with an unresolved required failure is `BLOCKED`, never `DONE`.

## Legacy decision register

| ID | Decision | Continuing value or rationale | Completion criterion | Owner |
| --- | --- | --- | --- | --- |
| L-01 | replace | Released predecessor comparison is valuable; mutable current-state freezing is not | Immutable release history and semantic compatibility gate replace `baseline.v1.json` | `platform.openapi` |
| L-02 | remove | Empty waiver machinery has no consumer | No waiver file, schema, code, requirement, or test remains | `platform.openapi` |
| L-03 | simplify | Semantic owner decomposition remains valuable | At most one order-independent unit per owner | `platform.openapi` |
| L-04 | replace | Domain schemas remain valuable under their semantic owners | `platform.openapi` owns transport-wide components only | Semantic owners |
| L-05 | retain and simplify | Safe deterministic aggregation prevents ambiguous sources | Migration-only branches removed; safety and atomicity retained | `platform.openapi` |
| L-06 | retain | Shared HTTP envelopes, errors, parameters, and security reduce drift | Only reachable transport-wide components remain shared | `platform.openapi` |
| L-07 | replace | Runtime parity is valuable; authored duplicate inventories are not | Generated catalog replaces every manual operation list | `platform.openapi` |
| L-08 | replace | Active route diagnostics have value | Real binder populates the only active registry | `platform.openapi` |
| L-09 | replace | A real registration boundary has value | No no-op wrapper or AST-only boundary test remains | `platform.openapi` |
| L-10 | simplify | Exact route behavior is valuable | Every canonical route uses method-qualified binding | Semantic owners |
| L-11 | replace | Assessment creation is a canonical row-create implementation | Workbook dispatcher owns the route; assessments supplies a provider | `module.assessments` |
| L-12 | replace | Network Flow remains a separate contract family | Its own generated binder replaces a string exclusion | `network_flow_activity` |
| L-13 | simplify | Semantic and live-route tests retain value | No hardcoded operation count or migration-only tree walker remains | Verification owners |
| L-14 | replace | Generated Go contracts remain valuable | Runtime imports are family-scoped | Generated-artifact owners |
| L-15 | replace | Typed TS contracts remain valuable | Raw OpenAPI is absent from browser runtime exports | `package.protocol_ts` |
| L-16 | replace | Network Flow should own only its family | Protocol package owns cross-family selections | `package.protocol_ts` |
| L-17 | replace | HTTP clients need one wire authority | Selected consumers use generated bindings and validators | Protocol and web owners |
| L-18 | replace | Explicit public projection remains valuable | One closed source schema feeds generated Go and TS types | View-schema owners |
| L-19 | remove | Silent Timeline substitution masks invalid state | Unknown schema lookup returns an explicit failure | Workbook owners |
| L-20 | retain | Startup recovery and constrained layout evolution protect persisted user data | Reason-coded recovery and evolution remain executable | Workbook and saved-view owners |
| L-21 | remove | Generic empty-change audit projections are ambiguous | No legacy public projection, vocabulary, filter, UI, or fixture remains | Audit owners |
| L-22 | retain | Raw journal and safe projection have audit and recovery value | Atomic immutable persistence and backup/restore remain covered | `platform.audit` |
| L-23 | retain and complete | Lifecycle and membership audit have active product and security value | Ordinary admin UI and browser evidence exist | `module.incidents` |
| L-24 | replace | Durable requirements and accurate human guidance have continuing value | No active migration-only requirement or stale authority prose remains | Contract and harness owners |

## Workstreams

| ID | Phase | Status | Dependencies | Compatibility | Owners | Started | Completed | Migration impact | Rollback boundary | Commands and run roots | Failure | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| OHTTP-00 | Baseline | DONE | none | none | `platform.openapi` | 2026-07-25 | 2026-07-25 | none | tracker only | `lint-markdown`: `20260725T234114Z-p3523336` | none | Complete |
| OHTTP-01 | Authority | DONE | OHTTP-00 | tooling only | `platform.openapi`, generated artifacts | 2026-07-25 | 2026-07-25 | none | release authority slice | owner slice: `20260725T234312Z-p3525954` | Initial catalog-order failure corrected and rerun | Complete |
| OHTTP-02 | Assembly | DONE | OHTTP-01 | wire-neutral | `platform.openapi`, semantic owners | 2026-07-25 | 2026-07-25 | none | source/assembler slice | owner slice: `20260725T235022Z-p3535212`; generation: `20260725T235009Z-p3532844`; drift: `20260725T235059Z-p3538720`; JSON shape: `20260725T235110Z-p3542236`; artifact policy: `20260725T235113Z-p3542729` | Scratch-copy omission caused `20260725T235027Z-p3535808`; fixed and narrow drift passed | Complete |
| OHTTP-03 | Runtime metadata | DONE | OHTTP-02 | wire-neutral, `1.0.1` metadata | `platform.openapi` | 2026-07-25 | 2026-07-25 | none | generated catalog slice | generation: `20260725T235828Z-p3563824`; owner slices: `20260725T235834Z-p3566054`, `20260725T235839Z-p3566821`; drift: `20260725T235853Z-p3570831`; JSON shape: `20260725T235905Z-p3574581` | Initial requirement ordering failure corrected before final runs | Complete |
| OHTTP-04 | Routing | DONE | OHTTP-03 | strict routing | route owners | 2026-07-25 | 2026-07-26 | none | owner registration groups | OpenAPI: `20260726T014735Z-p4147807`; server: `20260726T010257Z-p3876359`; Network Flow: `20260726T014302Z-p4113659`; auth service: `20260726T013807Z-p4079181` | Live SAML completion route was absent from OpenAPI; declared, bound, and exact-diff approved | Complete |
| OHTTP-05 | Generation | DONE | OHTTP-02 | workspace package | generated-artifact owners | 2026-07-25 | 2026-07-26 | none | family generation slice | generation: `20260726T015527Z-p4182580`; protocol: `20260726T015731Z-p4517`; view contracts: `20260726T015731Z-p4504`; drift: `20260726T015829Z-p33783`; policy: `20260726T020000Z-p60359` | Evidence response inheritance was unsatisfiable; owner schema was corrected before generation | Complete |
| OHTTP-06 | Frontend bindings | DONE | OHTTP-05 | wire-neutral | protocol, web, semantic owners | 2026-07-25 | 2026-07-26 | none | operation family | frontend unit: `20260726T015703Z-p1592`; boundary: `20260726T015703Z-p1581`; protocol: `20260726T015731Z-p4517` | Strict validators exposed stale fixtures and a direct UI protocol import; fixtures and adapter boundary were corrected | Complete |
| OHTTP-07 | Audit | DONE | OHTTP-01, OHTTP-06 | breaking, `2.0.0` | audit owners | 2026-07-25 | 2026-07-26 | migration `00040` | old app reads cleaned DB; snapshot required to restore removed projections | audit unit: `20260726T010402Z-p3905665`; audit service: `20260726T010415Z-p3907599`; migration: `20260726T015546Z-p4187769`; compatibility: `20260726T015024Z-p4173011` | Migration history attachment was initially missing and was added with the exact migration digest | Complete |
| OHTTP-08 | Incident UI | DONE | OHTTP-06, OHTTP-07 | additive UI | incidents and web | 2026-07-25 | 2026-07-26 | none | frontend slice | incidents: `20260726T014830Z-p4150682`; incident service: `20260726T005955Z-p3851879`; frontend unit: `20260726T015703Z-p1592`; stateful: `20260726T021713Z-p333836`; accessibility: `20260726T022007Z-p357614` | none unresolved | Complete |
| OHTTP-09 | View schema | DONE | OHTTP-05 | strict internal source | view/workbook owners | 2026-07-25 | 2026-07-26 | none | schema/generator slice | view schema: `20260726T015731Z-p4491`; view contracts: `20260726T015731Z-p4504`; workbook: `20260726T013139Z-p4035176`; JSON shape: `20260726T020000Z-p60399` | Registered source schema initially lacked required `schema_id`; all owner documents and generated types now carry it | Complete |
| OHTTP-10 | Verification | DONE | OHTTP-04 through OHTTP-09 | internal | verification and harness owners | 2026-07-26 | 2026-07-26 | none | owner catalogs and generated topology | harness contract passed; generation drift: `20260726T020147Z-p71424`; JSON shape: `20260726T020201Z-p75587`; check: `20260726T020933Z-p208470`; Markdown: `20260726T020107Z-p63560` | Managed owner slices raced stale service images; one readiness work unit now rebuilds them before dependent rows | Complete |
| OHTTP-11 | Final validation | DONE | all | release review | all affected owners | 2026-07-26 | 2026-07-26 | migration and rollback boundary verified | last completed workstream | finalize: `20260726T020118Z-p65030`; webserver: `20260726T021159Z-p305113`; stateful: `20260726T021713Z-p333836`; accessibility: `20260726T022007Z-p357614`; visual: `20260726T022216Z-p378134`; release: `20260726T022432Z-p399390` | Initial check `20260726T020256Z-p81103` found one stale auth operation set; corrected and narrow/full reruns passed | Complete; promote release pointer only after publication |

## Checkpoint template

For every workstream append:

- Timestamp and status transition.
- Exact authored files and generated outputs changed.
- Substantive decisions and compatibility classification.
- Migration preconditions, before/after counts or digests, and rollback boundary.
- Commands, result/run roots, and relevant summary artifacts.
- Failure classification and exact next action.
- `make lint-markdown` result.

## Completed checkpoints

### OHTTP-00 — Baseline

- Created this successor tracker without modifying the historical tracker.
- Recorded commit `53f4553150117d09535794e59a8f500485bdb94c`, the
  released `1.0.0` document hash, byte length, and structural inventory.
- Confirmed no named external consumer or concurrent `/api/v2` dependency is
  registered in repository authority.
- `make lint-markdown` passed at
  `.cartulary/test-results/20260725T234114Z-p3523336`.

### OHTTP-01 — Released-contract compatibility

- Captured the exact released `1.0.0` document under
  `contracts/openapi-releases/` with a closed immutable release registry.
- Added a conservative semantic comparator, exact fingerprinted change-set
  contract, SemVer enforcement, and generation-pipeline verification.
- Added durable requirement and verification ownership under
  `platform.openapi`.
- The initial owner slice rejected unsorted `contract_ids`; the machine
  catalog was corrected and the same slice passed at
  `.cartulary/test-results/20260725T234312Z-p3525954`.
- No migration impact. Rollback is the release-authority slice as a unit.

### OHTTP-02 — Semantic owner units

- Replaced 70 numbered fragments with 20 order-independent semantic owner
  units, including new explicit `module.assessments` and
  `platform.viewquery` ownership.
- Moved job, saved-view, report-composition, workbook, assessment, and
  view-query schemas out of `platform.openapi`.
- Removed the empty waiver registry, its schema and code paths, the mutable
  current-state baseline, v1 manifest schema, and associated tests and active
  requirements.
- Retained collision, reference, response, security, path-parameter,
  resource-limit, safe-path, and atomic-write validation. Canonical member
  ordering is now independent of manifest order.
- Official generation passed at
  `.cartulary/test-results/20260725T235009Z-p3532844`.
  `platform.openapi` passed at
  `.cartulary/test-results/20260725T235022Z-p3535212`; generation drift,
  JSON shape, and artifact policy passed at the run roots recorded above.
- The first generated-artifact slice exposed a missing scratch-copy
  registration for the new comparator, an in-scope harness integration
  failure. The registry was corrected and the narrow drift check passed.
- No wire-semantic or database migration impact. Rollback must restore the
  manifest, policy, owner units, assembler, catalogs, and generated outputs
  together.

### OHTTP-03 — Generated operation catalog

- Added a deterministic generated Go catalog containing method, path,
  `operationId`, semantic owner, availability, successful statuses, security,
  and state-changing metadata for every core OpenAPI operation.
- Added explicit availability metadata for Enterprise Authentication, Import,
  Incident Portability, Reference Pack, and Snapshot Reporting operations.
  All availability IDs resolve through the extension profile registry.
- Added an owner-scoped route-registry foundation that derives `ServeMux`
  patterns from the catalog, rejects unknown or wrong-owner bindings, detects
  duplicates, records active routes, validates claimed-profile parity, and
  attaches non-sensitive owner/operation/profile metadata to request context.
- `platform.httpapi` is the only runtime importer of the generated catalog.
- Recorded the metadata-only contract diff as an exact approved `1.0.1`
  change set. It contains no wire, security, request, or response change.
- Official generation, both affected owner slices, generation drift, and JSON
  shape passed at the run roots recorded in the workstream table.
- No database migration. Rollback removes the generated catalog, route
  foundation, availability metadata, and `1.0.1` change set together.

### OHTTP-04 — Contract-backed route binding

- Migrated every core public route to the owner-scoped generated binder and
  validated the active base and claimed-profile sets before serving.
- Removed all authored `public_operations.go` inventories, application
  aggregation, no-op wrappers, exclusion enums, and the AST-only route test.
- Replaced broad prefix registration with generated method-qualified patterns.
  Assessment creation now participates in the canonical workbook dispatcher;
  Network Flow uses its own generated route catalog and profile binder.
- Live verification found the production SAML completion route was missing
  from the former parallel inventory. It is now the conditional
  `finishEnterpriseSAML` operation, not an undocumented active route.
- Owner slices and service-backed route evidence passed at the run roots in
  the table. Route owner groups remain independently revertible.

### OHTTP-05 — Family-scoped generation

- Replaced the monolithic Go and TypeScript artifact dump with family-scoped
  packages and explicit package entrypoints. Protocol package selection is now
  owned by `package.protocol_ts`; Network Flow owns only its family selection.
- Removed the raw OpenAPI artifact from browser runtime exports and added
  generated-artifact and frontend import policy coverage.
- Generation validation exposed an unsatisfiable closed-schema inheritance
  between evidence attachment data and workbook mutation data. The semantic
  owner contract now uses shared open fields with closed concrete schemas; the
  exact compatibility fingerprints are approved in the `2.0.0` change set.
- Generation is staged and validated before tracked outputs are replaced.

### OHTTP-06 — Generated HTTP operation consumers

- Added a closed operation-selection contract and generated request/response
  types, validators, path builders, and query encoders for repository web
  consumers while retaining the existing transport policy.
- Migrated account/session/user, audit, incident lifecycle and membership
  audit, evidence, discovery, and view-schema consumers. Removed duplicate
  audit decoding, schema anchors, and handwritten path construction in those
  families.
- Strict response validation exposed stale mock envelopes, invalid UUIDs, and
  incomplete resources. Fixtures were corrected to the public contract;
  validators were not weakened.
- UI components now consume generated protocol knowledge only through approved
  service or `app/api` adapters. The initial boundary failure at
  `20260726T015547Z-p4188015` was resolved and the exact check passed.

### OHTTP-07 — Audit cleanup and `2.0.0`

- Removed active `legacy_backfill` and `legacy_administrative_event`
  vocabulary, filters, UI labels, constants, and manufactured safe
  projections. The immutable raw journal and safe transactional projection
  boundary remain.
- Generated audit vocabulary and redaction bindings from
  `contracts/audit/index.json`.
- Added transactional migration
  `00040_remove_legacy_administrative_audit_projections.sql`. It deletes only
  ambiguous projected rows, tightens checks, recreates immutability, and
  asserts completion. Raw history is never rewritten.
- Registered the exact migration hash in machine history. Migration drift,
  unit and service-backed audit evidence, and semantic `1.0.0` to `2.0.0`
  compatibility checks passed.

### OHTTP-08 — Incident product closure

- Replaced membership-audit placeholder content with an admin-only generated
  client surface supporting action, actor, and target filters plus keyset
  pagination, empty/error/loading states, redaction, and UTC-safe display.
- Added admin-only close/reopen controls with normalized reasons, fresh client
  transaction IDs, version-conflict refresh without silent replay, returned
  resource replacement, focus handling, and live status.
- Owner, service-backed, frontend unit, stateful, and accessibility evidence
  passed at the run roots recorded in the table.

### OHTTP-09 — Closed view-schema source

- Added and registered one closed source schema, generated family-scoped Go and
  TypeScript types, and explicit `schema_id` identity on every owner document.
- Retained cross-artifact semantic checks and the deliberate public projection
  boundary. Unknown workbook schema IDs now fail explicitly instead of
  selecting Timeline.
- Retained reason-coded startup recovery and constrained additive
  hidden/read-only saved-layout evolution. No persisted layout rows were
  rewritten.

### OHTTP-10 — Durable verification and harness ownership

- Replaced migration-characterization, waiver, mutable-baseline,
  hardcoded-count, and correction-specific rows with active evidence for
  deterministic assembly, release compatibility, runtime-profile parity,
  generated clients, audit safety, and public view projection.
- Regenerated authored verification, test-family, task-surface, and topology
  ownership. Human guidance now states that machine contracts are upstream.
- Fixed an owner-slice correctness race: managed and Playwright rows now depend
  on one owner-scoped service-runtime readiness unit that runs
  `make test-service-images`. Harness regression coverage locks the dependency.
- Targeted security initially reported repository file reads in the new
  comparator at `20260726T015731Z-p4850`. The comparator now uses `os.Root`;
  the rerun passed at `20260726T015829Z-p34068`.

### OHTTP-11 — Production validation and handoff

- `make agent-finalize` passed at
  `.cartulary/test-results/20260726T020118Z-p65030`. Retained-run maintenance
  was skipped because `RESULTS_DIR` was unset.
- Final generation drift, artifact policy, JSON shape, migration drift,
  harness contract, frontend type, and import-boundary gates passed. The
  durable run roots include `20260726T020147Z-p71424`,
  `20260726T020159Z-p75219`, `20260726T020201Z-p75587`,
  `20260726T020204Z-p76111`, and `20260726T020249Z-p80489`.
- The first full check at `20260726T020256Z-p81103` found an in-scope
  migration-era auth descriptor assertion that omitted the live SAML
  completion operation. The semantic operation/status set was corrected,
  the hardcoded inventory count was removed, and the complete auth owner
  slice passed 10/10 work units at `20260726T020448Z-p173209`.
- The rerun of `make check` passed 154/154 work units and 683 tests at
  `.cartulary/test-results/20260726T020933Z-p208470`.
- Standalone webserver-backed, stateful, accessibility, and visual browser
  gates passed for the default and claimed Network Flow profiles at the run
  roots recorded in the table.
- `make release-check` passed 12/12 release work units and its embedded
  683-test check at
  `.cartulary/test-results/20260726T022432Z-p399390`.
- The exact `1.0.0` to `2.0.0` semantic change set is approved with no
  unclassified or unexpected security changes. The immutable latest-released
  pointer remains `1.0.0`; promotion is a post-publication action.
- Application rollback remains compatible with the cleaned audit projection
  table. Recreating deleted ambiguous public projections requires an explicit
  pre-migration database snapshot and is not an automatic down migration.

## Residual behavior register

- `/api/v1` remains the public path namespace.
- The five credential schemes and semantic-owner authorization remain.
- Enterprise authentication remains conditional on its claimed profile.
- Network Flow remains outside the core OpenAPI contract.
- Raw audit history and sanitized atomic projections remain.
- Audit resource vocabularies remain forward tolerant where explicitly
  additive; query vocabularies remain closed.
- Workbook startup recovery and constrained saved-layout evolution remain.
- `docs/domain.md` and historical trackers remain non-executable guidance.

## Final validation and handoff

OHTTP-11 is complete. It records `make agent-finalize`, generation and
JSON-shape drift, migration drift, harness contract, frontend boundaries,
`make check`, webserver-backed/stateful/accessibility/visual browser evidence,
and `make release-check`. No required failure remains. The `2.0.0` release
snapshot must not become the latest released baseline until publication
succeeds.
