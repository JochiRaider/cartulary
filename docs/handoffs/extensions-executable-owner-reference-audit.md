# Extensions executable owner-reference audit

This manual source review supplements the controlling Extensions tracker. It is
not executable metadata or evidence consumed by the product or test harness.
The audit began at E-327 and closes at E-398 after the separately checkpointed
corrections through E-395 and final source-restatement review. This establishes
executable owner-reference closure, not complete module readiness.

| Boundary | Actual implementation | Disposition and remaining witness |
| --- | --- | --- |
| Core human and attributed identities | `internal/platform/jobs/jobcontract/identity.go` | Current v1 and v2 have different continuing production consumers. Retain both. Closed decoding and canonical UUID checks are implemented. Core §17 owns the closed field sets, actor/scope relations and canonical UUIDs. Route-specific request limits remain with their actual owners; no Incidents-only transaction limit is imported as a global Core rule. |
| Common Job success and resource identities | `internal/platform/jobs/jobcontract/terminal.go`, Jobs transaction completion and Extensions reconciliation | WS-57/E-331 closes resource semantics; WS-60/E-352 closes selected-owner codes and minima. F-066/WS-67 closes at E-395: terminal-success v2 and commit-proof v2 bind exact RFC-8785 bytes through the existing shared primitive, with fixed vectors and real finalizer/reconciliation/restore evidence. No current-resource lookup or inactive semantics are introduced. |
| Transaction operations and byte contracts | `internal/app/extensionassembly/validation_policy.go`, actual Incident Bundles and Network Flow owner contracts | WS-29/44 already bind the actual descriptor, operation validators and exact canonical byte contracts. The concrete operation policy resolves the selected owner descriptor and invokes typed owner input validation; prior coherent-rebinding rejection evidence remains required in the final portfolio. No replacement generic callback registry is warranted. |
| Reporting frozen-model admission | `internal/modules/reporting/render_model_admission.go`, `reporting_job_worker.go` | WS-65/E-382 removes the redundant self-participant, generic emitter contract and timeout. Reporting privately admits exact frozen model bytes, schema/identity/digest, limits, copy isolation and redaction fields under its own deadline. Current rendering/lifetime/process checks pass within their scopes. |
| Network Flow state admission | `internal/app/server/extension_admission.go`, Network Flow saved-graph and state validators | Concrete typed callbacks perform current cutover and final validation. WS-21 established actual inactive restore and subsequent claim boundaries. WS-58/62/64 bind exact algorithm/configuration references and isolate inactive construction. E-395 adds current proof-preserving restore/claim process evidence; inactive semantics never execute. |
| Graph engine and current restore binding | `internal/modules/graphprojection`, `internal/modules/graphprojection/restore/contract.go`, `tools/contractgen/recovery_validation.go` | E-336 adopts the eight-member binding v7; E-338 implements actual typed algorithm/projection support and strict byte admission, removes unused stage/provenance/database declarations, and records passing generator, Graph and current Recovery selections. E-339 completes actual current capture/restore and later-claim process evidence and the WS-58 checkpoint. |
| Extensions supporting references | `internal/modules/extensions/registry_schema.go`, `coordinator.go`, authored implementation bindings | WS-28/29/61 close required and extra identities, participant joins, exact operational references and coherent-rebinding rejection. WS-62 binds configuration schemas to the actual typed catalog; WS-63/65 use native source admission and private render validation. These are executable owner consumers, not declaration-only registrations. |
| Authored contract byte admission | `tools/contractgen/main.go`, `party_policy.go`; shared `canonicaljson.DecodeStrict` | F-056 / WS-59 closed E-344: common and Party readers now use shared strict byte admission, with malformed-byte and valid Unicode/number vectors. E-343 records 347 unchanged generated files and the justified real generator-source integrity update; E-344 proves actual coordinator admission and immutability. |
| Job success code registry | Core 01 §3.3.9.1/§17, NF-REQ-170c; `jobs/definition_catalog.go`, `jobcontract/terminal.go`, `extensions/jobs.go` | F-057/F-058 / WS-60 closed E-352: current job-kind v3 and base catalog v2 project exact codes and required resource minima into immutable typed definitions. Live completion, public and owner retained reads, and inactive proofs share selected admission. E-350/351 records passing unit/service/static evidence. Two composite collaborator rows still fail at unfenced Recovery fixtures (F-017/WS-16); they receive no passing portfolio credit. |

## Correction sequence

The WS-30 audit checkpoint precedes WS-57 (Core Job resource contracts), then
WS-58 (Graph executable definitions and binding cleanup), WS-59 (strict
authored contract byte admission), WS-60 (selected-owner Job success codes),
then final WS-30 closure. Each implementation slice has its own tracker checkpoint before its
successor begins. All four corrections block WS-11. WS-36, WS-16, WS-18, WS-11
and WS-12 remain required.

WS-57 repairs implementation against existing Core 01 §3.3.9.1 and REQ-634;
it does not invent global scalar bounds or remove live v1 identity consumers.
Use typed, pure structural validation shared by current execution and retained
proof admission. Validate canonical IDs, exact resource read routes, declared
kind/count limits, duplicate rejection and byte-preserving terminal proofs.
Current invalid values receive no compatibility alias or historical reader.

WS-58 must review the exact Graph owner amendment before implementation and
record adoption under the user's standing authorization. The current-only
cutover must update authored contracts, generated bindings and dependent
generation identities together. Its validation must reject unknown algorithms,
unresolved definitions, changed definition digests and retired binding shapes;
current graph execution, reconstruction and current Recovery must still pass.
Do not hash source files into backup compatibility merely to replace placeholders.
Build attestation and release qualification are distinct from logical format
compatibility and remain separately reported.

## Risks and exits

Stricter resource validation may expose invalid test fixtures or an actual owner
writer defect. Diagnose the precise caller before changing a contract; preserve
all valid current production kinds and routes. Graph definition changes can
affect current generation admission even without changing database tables.
Regenerate through Make and validate a coherent unpublished cutover.

No correction is complete on declaration-only tests. WS-30 closes only
after each operational reference has an actual typed consumer and appropriate
negative evidence, and after all correction workstreams pass their exits.
Full module readiness still requires the complete compatible portfolio and the
final validation/handoff workstreams.

## Final-reference audit corrections

E-353 registers three separately sequenced corrections before implementation:
WS-61 (F-059) closed exact supporting/dependency reference inventories at E-359; WS-62
(F-060) binds configuration value schemas to the existing actual typed catalog;
WS-64 (F-062/F-063) closes at E-371, preserving recognized-unclaimable
policies and isolating inactive materialization. WS-63 (F-061) then admits the
Network Flow Reporting source at actual composition. E-374 adopts a dedicated
typed read contract because the former emitter declaration contradicted the
source owners. E-378 closes WS-63 with current registry, composition, byte-bound, real
source/lifetime and shipped-process evidence.
Transaction adapters remain. E-382 closes F-064 / WS-65: remove the render
self-participant, retain private frozen-model admission, and retire orphan
protocol/configuration metadata. Current unit, application/service, process and
production-build evidence passes within the recorded scopes. Each slice
checkpoints before its successor; required outstanding corrections still block
WS-30 closure and WS-11.

E-359 establishes the reference-inventory correction with actual coordinator,
generator, empty-registry and immutability evidence. Import declarations do not
confer executable support. WS-62 closes configuration binding at E-364: the actual typed catalog checks
exact key/schema sets before deployment input, with immutable references and no
callback execution. WS-64 then removes inactive owner defaults/views and
operational Network Flow construction, with current restore/later-claim process
evidence. Static transaction contract matching remains independent of the live
owner required for input execution. WS-63 binds Reporting source metadata to its concrete typed owner; WS-65 removes the redundant self-emitter. Final storage review also
registers F-065 / WS-66: Graph publication validates digest shape without
recomputing content, and exact reads do not prove envelope/row fidelity. This
violates Graph §§6.2/8/8.0 independently of the source metadata binding. Correct
the Graph-owned ports with real tamper and current-producer evidence. The next
sequence is WS-66 → final WS-30 → WS-36 → WS-16 → WS-18 → WS-11 → WS-12.

## Final audit outcome

WS-30 closes at E-398. Required operational definitions resolve to actual typed
owners and consumers, with the scoped positive and negative evidence referenced
above. Exact byte admission, immutable catalogs and current-only identity joins
remain distinct from selected owner semantics. No Markdown reader, generic
callback facade, inactive profile executor or compatibility alias is introduced.

WS-66/E-387 proves Graph content/row integrity at publication and every content
read while retaining metadata-lock ordering. WS-67/E-395 defines current Job
bytes, verifies actual finalization/reconciliation and proves full Job result
and proof preservation through real process capture/restore and subsequent
claims. These extend earlier owner/consumer witnesses without conferring release
or full module qualification.

The final source pass repairs superseded positive Graph v4/four-table/lease
restatements in Core 00, Core 04 AC-568, Appendix I, the Graph boundary decision
and rollout guide. Network Flow and Reporting criteria now describe current
major-8/state-5 and frozen-input behavior. Stable requirement/acceptance IDs,
current live Core identity versions and original adoption history remain.
Those corrections project already adopted owners and change no product source.

Remaining work is explicit: WS-36 registry owner-input/collision diagnostics,
WS-16 current Recovery fixture replacement and legacy test retirement, WS-18
actual browser/process obligations, WS-11 the complete compatible portfolio,
and WS-12 final validation and handoff. The known unfenced Recovery fixture
failures remain required corrections and receive no passing credit. Module
readiness is not yet established.

## Subsequent diagnostic-source discovery

E-400 records F-067 / WS-68 during WS-36 preparation. The diagnostic clauses
refer to absent normalization and fact-identity definitions (EXT-REQ-177/205).
This is a required new source/primitive correction before WS-36 implementation,
not passing evidence or an exclusion from readiness. It leaves the earlier
closed typed-consumer evidence scoped as recorded. Sequence now begins WS-68,
then resumes WS-36 before the remaining Recovery, process/browser and complete
validation/handoff workstreams.

The subsequently discovered registry definition gap is now closed by E-407:
Extensions EXT-REQ-177/205 and Table 5-B define normalization and current logical
fact identity v2. A pure shared primitive, closed authored schema, independent
vectors and actual generator/coordinator validation pass at E-406's source.
This closes the new source prerequisite; complete registry diagnostic collection
remains WS-36 and full module readiness remains a separate validation exit.

## Current Recovery consumer evidence

The earlier composite failures at unfenced restore calls are historical.
WS-16 replacement and post-removal evidence is retained in tracker E-435–438.
The current Reference Data canonical-persistence row and Report Composition
preview/private-Reporting row pass against actual capture, fenced restore,
independent Core rebuilds and durable target-local completion. Recovery's 14-row
selection also passes. This closes their fixture defect, not the full owner
portfolio or actual browser/process qualification. WS-75 source cleanup,
WS-74 static analysis, WS-18 boundary witnesses and WS-11/12 remain required.
