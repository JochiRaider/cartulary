# Extensions Subsystem Implementation Tracker

## 1. Scope and Source Posture

| Field | Recorded value |
| --- | --- |
| Target label | `extensions-subsystem` |
| Current implementation root | `internal/modules/extensions` |
| Proposed target contract | `docs/extension-subsystem-nlspec.md` |
| Planning framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md` |
| Boundary-completeness input | `temp/analysis-notes.md` (informative and non-normative) |
| Tracker output | `docs/handoffs/extensions-subsystem-implementation-tracker.md` |
| Planning baseline | Branch `revision/grid-adapter`, commit `200f631152b76cf102bb6e3f81953de820978075` |
| Implementation baseline | Branch `revision/grid-adapter`, commit `bf8427a55fdb5f2b8e919a9567baf48329b0da5b`; clean worktree before the ES-00 tracker checkpoint. |
| Draft baseline | SHA-256 `18fbd7f8c83e4a92ceec5bb0913a5443f454d8be4f96a2397df8627ec0915ee8`; unchanged between the recorded planning input and ES-00 start. |
| Execution status | ES-00 through ES-14 are `DONE`. Atomic promotion, post-promotion owner evidence, independent evidence classes, explicit-root audits, finalization, and broad/release gates are complete. |
| Boundary status | Decision state `SELECTED`; normative adoption state `DONE`; the Extensions NLSpec and every required companion owner are `adopted/current`, and runtime adoption is complete. |
| Authorized change scope | Tracker, tests, owner documents, contracts, generators and generated outputs, migrations, configuration, Harness v2 inputs, backend, frontend, and coordinated adoption required by ES-00 through ES-14. |
| Continuing authority constraint | Core 00 through Core 04 and the adopted named owners remain authoritative. Runtime-downloadable extensions, marketplaces, capability activation, separate extension hosts, and any Core 05 claim boundary remain future-only. |

The Extensions Subsystem NLSpec has `status: adopted/current`. Its coordinated
adoption completed with Core 00 through Core 04, Network Flow major 2, the named
shared owners, domain vocabulary, OpenTelemetry, and Testing Harness inputs on one
validated promotion. Current behavior remains allocated to those primary owners;
the Extensions document supplies the generic subsystem contract without absorbing
their behavior.
The local planning framework supplies planning doctrine and table structure only; it
is not evidence that a package, contract, test, or command exists.

`temp/analysis-notes.md` is informative source material for this revision. The
boundary rules selected from it are recorded canonically in Section 6. Tracker
decision selection and target-contract boundary closure are complete. Uppercase
normative language in this tracker governs execution and handoff only; product
conformance authority resides in the adopted Core and subsystem documents.

### Implementation baseline owner digests

| Owner input | ES-00 SHA-256 |
| --- | --- |
| Extensions draft | `18fbd7f8c83e4a92ceec5bb0913a5443f454d8be4f96a2397df8627ec0915ee8` |
| Core 00 | `021e8e495a91e5be9e808050d6e891a8d118c5025d4a4207a0cad9c084c0d5cd` |
| Core 01 | `a2cfc13348f6f23ea8fcd33da973fd847a867e476d7ed632c6d2e564957debf5` |
| Core 02 | `7ac20d0c3b55b46b46b6e5bfa961c916868b3c1b0574f1b6ed620b10dd84c1c0` |
| Core 03 | `bf196f685fc32804c593427842ad24f78666946c31e6dd1938c5286da3636c9e` |
| Core 04 | `ec35f9f1fccbc4292b95e6cebbe8a5040f6fbd9832a2880c44f0a911ddf671e5` |
| Domain vocabulary | `9b683eff8ec0c5a89216502390e65fd11caa3116549ea2a81154cf272b9bb2cd` |
| Testing Harness v2 | `61d120b42fac884034721510612af6fb6121662db67230292de693104fce1e89` |

The ES-01A Extensions draft digest is
`796e119fe4d74386027074ddf9954d1089cac72c7bb3a2fe79c231e24ffc3c62`.
It contains 236 existing requirement IDs, continuous acceptance criteria through
`EXT-AC-158`, 28 existing gate IDs, and remains `status: draft`.

These digests are the characterization inputs for ES-00. Later owner amendments
MUST update the applicable digest and locator records rather than treating this table
as current authority.

### ES-06 pre-work input snapshot

ES-06 started on branch `revision/grid-adapter` at commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 31 staged-worktree status
entries accumulated only from ES-01A through ES-05. Its dependency state is exact:
ES-01A and ES-05 are `DONE`; ES-07 remains unstarted. Current owner-document digests
are Extensions `796e119f...`, Core 00–04 `3358cc76...`, `9ad58e13...`,
`79803b2d...`, `79afa553...`, `67583e75...`, Network Flow `40c447d5...`,
Reporting `72f8948f...`, Composition `489a839d...`, OpenTelemetry `b31baa1e...`,
and domain vocabulary `ed7c51b0...`.

### ES-06 completion snapshot

ES-06 replaced `cartulary.extensions.phase2.v1` with the closed
`cartulary.extension_authored_input_catalog.v1` and admitted 40 explicitly indexed
owner-authored inputs. Ten dependency declarations bind exact owner documents,
manifests, anchors, and the Core 01 Base-reservation artifact. Six recognized
profiles now have complete recognition, route/workspace, claim-configuration,
classification, configuration, implementation-binding, and applicable state/client
inputs. The catalog also carries closed protocol limits, distinct portability result
contracts, inactive-value schemas, validation surfaces, closure mappings,
canonicalization/safe-reference vectors, and exact mappings for EXT-AC-142 through
EXT-AC-158.

Contract generation now validates the draft digest, catalog completeness, exact
owner-document bytes and anchor ranges, manifest and fragment digests, fragment
adoption, dependency/manifest parity, required per-profile fact multiplicity,
configuration presence, capability absence, validation-condition uniqueness, and
traceability ranges without scanning implementation code or owner prose. Extension
canonical inputs use the required final LF in their canonical digest. Isolated drift
generation now copies the exact owner documents required by those checks.

The passing completion roots are `make generate` at
`20260720T054424Z-p199375`, `make generate-drift` at
`20260720T054434Z-p200906`, `make backend-unit` at
`20260720T054707Z-p228024` (243 tests), `make json-shape-check` at
`20260720T054803Z-p232788`, `make generated-artifact-policy-check` at
`20260720T054805Z-p233131`, and `make harness-contract` at
`20260720T054807Z-p233309`. The first backend-unit attempt at
`20260720T054458Z-p203864` exposed a stale phase-index test reader; it was replaced
by exact owner-fragment characterization and the rerun passed. ES-07 owns only the
generated normalized/runtime/accounting projections from these validated inputs.

### ES-07 pre-work input snapshot

ES-07 started only after the ES-06 tracker checkpoint passed. The branch remains
`revision/grid-adapter` at commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 47 accumulated status entries.
The exact authored catalog digest is `4c46bfa7...`; the pre-ES-07 generated Go and
TypeScript contract projection digests are `16bb304a...` and `a31ff344f...`.
ES-07 may create only deterministic projections from the exact ES-06 inputs; it may
not add an owner fact, infer behavior from code/prose, or reintroduce the retired
phase registry.

### ES-07 completion snapshot

The pre-generation completeness audit expanded the catalog from the 40 ES-06
foundation inputs to 48 exact authored inputs. The eight additions are three typed
participant contracts, four Network Flow physical-backup codecs, and one generated
schema-source set. The audit also completed contribution ownership for Core 01,
Network Flow, and Reporting and replaced the single phase-shaped binding input with
an exact six-profile binding source set. All additions remain owner-authored facts;
none are inferred from packages, routes, or prose.

Contract generation now emits 46 deterministic `contracts/extensions/generated/**`
artifacts: the dependency snapshot; normalized owner inputs; six descriptors; the
registry and integrity objects; six bindings; validation conditions; six closure
catalogs; six conformance manifests and their index; accounting and clause
traceability; client support; and thirteen closed top-level generated JSON Schemas.
All 48 authored and 46 generated artifacts are packaged in the generated Go and
TypeScript projections. Registry/binding parity includes participant algorithms,
the exact four Network Flow authoritative state families, and four digest-bound
backup codecs. No phase-shaped identity, compatibility alias, or old-format reader
survives.

Passing roots are `make generate` at `20260720T062455Z-p270992`,
`make generate-drift` at `20260720T062500Z-p272501`, `make backend-unit` at
`20260720T062519Z-p275470` (243 tests), `make json-shape-check` at
`20260720T062638Z-p281673`, `make generated-artifact-policy-check` at
`20260720T062638Z-p281670`, `make harness-contract` at
`20260720T062638Z-p281659`, `make test-slice OWNER=module.extensions` at
`20260720T063205Z-p288085` (18 tests), and the service-backed owner slice at
`20260720T063217Z-p288606` (11 tests). The exact BC-004 through BC-008 selectors
validate the packaged projections, not tracker text alone. An initial generate at
`20260720T061607Z-p250556` rejected a legitimate empty participant-binding list;
the over-constraint was removed and the rerun passed. An initial strengthened owner
slice at `20260720T063135Z-p287358` exposed and corrected the clause-ID length
assertion; generated IDs and their source digests were unchanged.

Clause traceability and conformance acceptance rows deliberately cover the newly
allocated EXT-AC-142 through EXT-AC-158 boundary set at this checkpoint. ES-13 and
ES-14 retain the explicit obligation to audit all 158 criteria and every normative
clause before promotion. The standard client support row is packaged; its final
web-build asset-set digest is bound during ES-12 rather than fabricated here.

### ES-08 pre-work input snapshot

ES-08 started only after the ES-07 tracker and documentation checkpoint passed. The
branch remains `revision/grid-adapter` at commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 48 accumulated status entries.
Dependencies ES-00, ES-02, ES-06, and ES-07 are `DONE`; ES-09 remains unstarted.
The exact draft/Core 00–04 digests remain `796e119f...`, `3358cc76...`,
`9ad58e13...`, `79803b2d...`, `79afa553...`, and `67583e75...`. The 48-input
catalog digest is `7208e58c...`, and the generated Go package digest is
`02e903d1...`.

This slice may add only the generic coordination facade, immutable generated-input
reader, claim/dependency/binding admission, collision checks, and six-stage
publication-plan construction. Transport routing, storage adapters, profile
behavior, browser policy, telemetry vocabulary, configuration/lease lifecycle, and
actual component publication remain with their named later slices and owners.

### ES-08 completion snapshot

`internal/modules/extensions` now exposes one immutable generated-artifact port and
a cohesive coordinator for registry queries, explicit claim resolution, dependency
ordering, collision admission, implementation-binding parity, and process-local
publication-plan construction. It validates exact packaged byte digests against the
registry-integrity object, including the Base reservation and client-support
artifacts. It never scans packages, routes, owner prose, the filesystem, or runtime
reflection, and it invokes no profile, transport, storage, browser, reporting,
backup, portability, or telemetry behavior.

Claim resolution rejects unknown, unclaimable, duplicate, missing-dependency, and
incompatible requests without auto-claiming. The canonical topological order is
dependency-first with lexical tie-breaking. Publication construction derives the
contribution, all-route reservation/dispatch, claimed-workspace, worker, fixed
listener-gate, client-support, and binding component digests from that single
resolution. Its DTO and byte accessors return copies. Application assembly admits
and retains the coordinator before constructing public HTTP behavior; ES-09 still
owns the actual discovery/dispatch/publication switch.

Five exact Harness rows increased the Extensions family to 23 rows and execute the
registry, claim/dependency, collision, binding, and publication-plan selectors.
Passing evidence includes `make generate` root `20260720T064356Z-p300514`,
`make generate-drift` root `20260720T064711Z-p309208`, local owner root
`20260720T064727Z-p312162` (23 tests), service-backed owner root
`20260720T064727Z-p312171` (11 tests), app-server owner root
`20260720T064727Z-p312190` (20 tests), module-boundary root
`20260720T064904Z-p325448`, backend-unit root
`20260720T064904Z-p325460` (248 tests), JSON-shape root
`20260720T064904Z-p325462`, generated-policy root
`20260720T064904Z-p325472`, and a passing `make harness-contract` invocation.
Initial owner root `20260720T064123Z-p296675` exposed a missing local scalar helper;
root `20260720T064405Z-p302043` then exposed an incorrect treatment of exact Base
reservations as descendant captures. Both defects were corrected. Initial Harness
root `20260720T064727Z-p312170` detected the expected five-row catalog-count drift;
the exact aggregate assertions were updated and the rerun passed.

### ES-09 pre-work input snapshot

ES-09 started only after the ES-08 tracker/document checkpoint passed. The branch is
still `revision/grid-adapter` at commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 51 accumulated status entries.
ES-08 is `DONE`; ES-10 remains unstarted. The draft and authored catalog digests are
unchanged at `796e119f...` and `7208e58c...`; the generated Go package, coordinator,
and 23-row Extensions family digests are `1ae48cfa...`, `3c6bf5d7...`, and
`d34dfa38...`.

This slice owns the one coordinated Core 01 switch: resolve one process claim set,
derive discovery/reservations/dispatch from the same coordinator epoch, publish the
seven-member discovery resource, update OpenAPI/generated clients, and ensure no
listener/worker becomes visible before plan installation. It removes the hard-coded
platform profile registry and any profile-local discovery producer; configuration,
lease/deadline/fatal lifecycle details remain ES-10.

### ES-09 completion snapshot

Core 01 discovery, reservations, and dispatch now derive from the same admitted
generated registry and resolved process claim set. Application startup constructs
the coordinator, resolves claims, and freezes the six-component publication plan
before telemetry, stores, modules, listeners, workers, or the HTTP handler are
created. The runtime retains that exact plan and claim set; the injected HTTP
profile view is derived from the same descriptor/claim epoch.

`GET /api/v1/extensions` now has one strict seven-member producer. Its authored
OpenAPI schema and generated Go/TypeScript contracts expose exactly `profile_id`,
`claimable`, `claimed`, nullable `contract_major`, `route_families[]`,
`workspace_keys[]`, and empty `capabilities[]`. The TypeScript decoder reads only
those members, ignores unknown additive members, and never returns or executes
them. Runtime and test contract readers consume the generated profile registry;
the former platform literal registry and owner-fragment reconstruction are gone.
Network Flow contract artifacts are version `2.0.0`/major `2`, and its authored
profile-local discovery object and top-level singular route root are removed.

Passing roots are: generation `20260720T070407Z-p345701`, drift
`20260720T070422Z-p347244`, JSON shape `20260720T070422Z-p347282`, generated policy
`20260720T070422Z-p347249`, Extensions owner `20260720T070437Z-p350769`, Protocol TS
`20260720T070737Z-p378113`, web Network Flow `20260720T070737Z-p378102`, exact
incident discovery unit/integration `20260720T071444Z-p443228`, frontend unit
`20260720T071500Z-p443642`, backend unit `20260720T071559Z-p445589`, frontend import
boundary `20260720T071714Z-p451030`, and app-server process publication
`20260720T071714Z-p451044`. `make frontend-typecheck` also passed.

Initial frontend/backend roots `20260720T070437Z-p350749` and
`20260720T070437Z-p350751` identified stale three-member fixtures and a stale
test-contract projection. Incident owner root `20260720T070737Z-p378085` identified
the remaining three-member integration assertion. All were structurally replaced
and exact reruns passed. Network Flow owner root `20260720T070737Z-p378090` failed
only in browser-stack startup with a shared PostgreSQL container resource conflict;
its affected product selectors passed separately and the full evidence-class retry
remains assigned to ES-13 rather than being misreported as ES-09 product evidence.

### ES-10 pre-work input snapshot

ES-10 started only after the ES-09 tracker checkpoint passed `git diff --check`,
`make lint-markdown`, and `make harness-contract`. The branch remains
`revision/grid-adapter` at commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 76 accumulated status entries.
ES-09 is `DONE`; ES-11 remains unstarted. The Extensions draft, staged Core 01,
staged Core 04, coordinator, and 23-row Extensions family digests are
`796e119f...`, `9ad58e13...`, `67583e75...`, `3c6bf5d7...`, and `d34dfa38...`.

This slice is restricted to BC-002, BC-006, BC-009, BC-010, the runtime half of
BC-011, and the process half of BC-017: validation-result precedence and registered
conditions, inert inactive configuration, the application-process lease, checked
deadlines/cancellation, readiness/admission closure, fatal component-loss/drain
semantics, safe diagnostics, and exit `2`/`70`. It does not add durable state,
participant, portability, backup, staged-object, or browser lifecycle behavior;
those remain ES-11/12.

### ES-10 completion snapshot

ES-10 is `DONE`. The Extensions runtime now enforces the generated condition
registry and the exact validation-result precedence, performs only inert structural
checking of unclaimed `syntax_only` values, calculates checked/saturating deadlines,
and exposes no secret resolver, egress, retained-value, profile-callback, or inactive
configuration-view path. Core 04 process composition now owns a dedicated-session
application lease with closed uncertainty/loss behavior, admission/readiness closure,
listener-bound Stage 6 activation, bounded drain, and typed startup/fatal exits `2`
and `70`. Fatal publication loss is irreversible in-process and cannot restart the
published epoch.

The runtime implementation digests at checkpoint are coordinator `3984a380...`,
validation `bb08995f...`, process lease `cd42bade...`, and lifecycle `08a76d91...`.
The staged Extensions/Core 04 owner documents remain `796e119f...` and
`67583e75...`; no owner text was activated. The worktree contains 100 accumulated
ES-01A-through-ES-10 status entries on `revision/grid-adapter` at baseline commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b`.

Exact passing evidence includes Extensions local `20260720T080718Z-p592834` and
service-backed `20260720T080737Z-p593448`, platform configuration
`20260720T075138Z-p502999`, HTTP runtime `20260720T075148Z-p503254`, app-server
lifecycle `20260720T075305Z-p505479`, readiness `20260720T075352Z-p509938`,
PostgreSQL lease integration `20260720T075403Z-p510201`, backend unit
`20260720T075433Z-p510741`, backend process `20260720T080627Z-p574091`, targeted
Gosec `20260720T075614Z-p522474`, OTel `20260720T080105Z-p558426`, web build
`20260720T080053Z-p555104`, Protocol TS `20260720T080130Z-p561439`, generation
`20260720T080531Z-p571618`, drift `20260720T080814Z-p594313`, generated policy
`20260720T080822Z-p597201`, JSON shape `20260720T080824Z-p597380`, and Harness
`20260720T080536Z-p573086`.

Diagnosed initial failures were not accepted as evidence. Generate roots
`20260720T073513Z-p471237` and `20260720T073623Z-p472350` exposed stale Core 04
fragment digests; Extensions `20260720T073945Z-p477877` exposed a compile-time
inactive-validator helper error; backend unit `20260720T074103Z-p481753` and
readiness `20260720T075318Z-p507319` exposed stale test expectations/imports. OTel
roots `20260720T075632Z-p543547` and `20260720T075750Z-p547410` found that the
complete generated owner-evidence family was incorrectly entering the browser
runtime bundle. The durable fix adds explicit, validated TypeScript runtime-artifact
prefixes per contract family: Go retains the full immutable Extensions artifact set,
while browser packaging receives only the client-support, descriptors, and registry
projections. JSON-shape root `20260720T080216Z-p561982` then found the checker did
not yet admit that authored key, and Harness roots `20260720T080347Z-p568038` and
`20260720T080453Z-p570612` found the expected four-row/eight-selector catalog growth;
the checkers and exact aggregate expectations were updated and all listed reruns
passed.

At the ES-10 checkpoint, BC-011 remained `IN_PROGRESS` only for the ES-11
transaction coordinator's proven-commit/cancellation matrix, and BC-017 remained
`IN_PROGRESS` only for ES-11 dequeue-gate/worker proof and committed-state
preservation. That checkpoint claimed no durable protocol, state migration,
backup/restore, staged-object, or browser-lifecycle outcome.

### ES-11 pre-work input snapshot

ES-11 started only after the ES-10 tracker checkpoint passed `git diff --check`,
`make lint-markdown` at `20260720T081236Z-p598843`, and `make harness-contract` at
`20260720T081301Z-p600365`; the final tracker-only Markdown rerun also passed. The
branch remains `revision/grid-adapter` at commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 100 accumulated status entries.
ES-10 is `DONE`; ES-12 remains unstarted.

The staged Core 01, Core 02, Network Flow, shared-protocol, physical state-binding,
implementation-binding, and 23-row Extensions family digests are `9ad58e13...`,
`79803b2d...`, `40c447d5...`, `d6d2bd9e...`, `4ad4df1c...`, `3953aadd...`, and
`d34dfa38...`. Authored migrations currently end at
`db/migrations/00033_network_flow_keyset_indexes.sql`. No generic extension metadata,
migration ledger, staged-object ledger, state coordinator, participant transaction
coordinator, backup-codec runtime, or extension job-proof component exists.

Existing `internal/platform/jobs`, `internal/platform/objectstore`,
`internal/modules/recovery`, Incident Bundle/Reporting participant paths, and
Network Flow authoritative tables remain their named owners. ES-11 may add narrow
ports and adapters around them, but may not move their behavior or storage authority
into `internal/modules/extensions`. The slice is restricted to BC-001, BC-003, the
transaction half of BC-011, BC-012 through BC-014, and the durable worker half of
BC-017. Browser availability/capability behavior remains ES-12.

### ES-11 completion snapshot

ES-11 is `DONE`. It remained on branch `revision/grid-adapter` at baseline commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b`; the accumulated worktree has 114
status entries. The Extensions draft remains `status: draft` at digest
`796e119f...`. No owner contradiction was encountered, and ES-12 did not start
before this tracker update.

The slice adds forward-only migration `00034_extension_coordination.sql` at digest
`097909a2...`. It owns only generic state metadata, the immutable migration ledger,
staged-object/reference lifecycle, and extension job commit/cancellation proof. The
migration establishes Network Flow's known v1 lineage for existing and empty
deployments; authoritative Network Flow tables remain owned by `module.networkflow`.
The schema ownership manifest names the new coordination objects without absorbing
profile tables or the shared jobs table.

Generated runtime catalogs now cover state plans, physical backup/codec plans, and
typed participant contracts. Runtime admission verifies each packaged digest before
exposing immutable DTOs. The Network Flow physical restore groups are the exact
owner order `100/200/300/400`, and its import-apply and indicator-link transaction
participants are owner-authored and binding-complete. Raw final digests for the
physical input, implementation input, fragment, manifest, dependency set, and
generated Go package are `2e8397a1...`, `e2bf5268...`, `b36b3330...`,
`c2ce8c90...`, `bbea6381...`, and `91e04947...`.

`internal/modules/extensions` now coordinates authoritative-family-only state
presence, empty initialization, lineage/version admission, immutable migration
steps, exact-once final validation, ordered bounded shared transactions, distinct
portability import preparation, operation-scoped staged output, serialized/coalesced
cleanup, and stopped-empty sequential restore. Platform PostgreSQL and object-store
adapters own physical I/O. Named profile validators/counters remain in Network Flow;
transport, profile behavior, storage authority, Incident Bundle publication, and
Reporting behavior were not moved into Extensions. The periodic cleanup worker is
published only after Stage 6 and unexpected worker termination emits
`published_component_lost`; ordinary deletion failures remain retryable operations.

Passing evidence roots are: generation `20260720T083609Z-p617811`, generation drift
`20260720T085124Z-p649577`, Extensions local
`20260720T084722Z-p630618`, Extensions service-backed
`20260720T084918Z-p646306`, backend unit `20260720T084745Z-p633176`, backend store
`20260720T085151Z-p654620`, backend integration
`20260720T090022Z-p697387`, backend process `20260720T085940Z-p678300`, migration
drift `20260720T085133Z-p652470`, module boundary
`20260720T085140Z-p654395`, targeted Gosec
`20260720T090113Z-p702968`, JSON shape `20260720T090206Z-p723472`, generated policy
`20260720T090207Z-p723806`, and Harness `20260720T090209Z-p723994`.

Diagnosed failing roots were not accepted as evidence. Generate roots
`20260720T083312Z-p612275`, `20260720T083424Z-p613426`, and
`20260720T083523Z-p614552` exposed authored-path order, transaction-key order, and
participant-parity omissions and were fixed structurally. Backend integration root
`20260720T085226Z-p663787` exposed a stale readiness test that skipped Stage 6;
root `20260720T085539Z-p671888` exposed an evidence test attempting overlapping
runtimes on one database under the adopted process lease. Both tests now follow the
published lifecycle and the passing rerun is recorded above. JSON-shape root
`20260720T090125Z-p722900` exposed the missing schema-object ownership declaration;
the exact generic ownership row was added and the rerun passed.

### ES-12 pre-work input snapshot

ES-12 started only after the ES-11 tracker checkpoint passed `git diff --check` and
`make lint-markdown`; the final rerun exited zero after the recorded Markdown root
`20260720T090818Z-p726470`. The branch remains `revision/grid-adapter` at baseline
commit `bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 114 accumulated status
entries. ES-11 is `DONE`; ES-13 remains unstarted.

The Extensions draft, staged Core 03 owner document/manifest, authored client-support
source, generated Protocol TS package, current App shell, Workbook shell, backend
workspace registry, and workbook-startup resource digests are `796e119f...`,
`79afa553...`, `5830c97a...`, `2cf623bd...`, `c1ec9450...`, `080fef29...`,
`0d69d9df...`, `c63b4057...`, and `4b5ea290...`. The generated client-support
registry is packaged at digest `00cd1fb6...`, declares the sole build class
`standard`, and supports exactly Network Flow contract major `2`, workspace
`network_analysis`, and no capabilities. Its current asset-set digest
`a3461154...` is still source-bound rather than bound to the final browser asset
set.

The backend currently resolves workbook extension pointers from claim declaration,
workspace declaration, and role only. The startup response has no no-store
`extension_workspace_availability` member, epoch, or generation. The web shell
loads discovery but does not intersect it with the generated client-support
registry or a current authorization/availability generation before rendering.
Network Flow controllers contain operation-local stale-result guards, but there is
no profile/workspace lifecycle owner that disposes all extension state on
authorization loss or epoch rollover while preserving Base workbook state.

ES-12 is restricted to BC-015 and the browser-facing negative half of BC-016. It may
add the immutable standard client-support projection, no-store availability result,
linearized epoch/generation controller, lazy workspace loader/fallback, and exact
extension-only disposal. It may not move discovery authority from Core 01, browser
lifecycle authority from Core 03, authorization from the workbook/backend owner,
or Network Flow behavior into Extensions. Base caches, pending requests and queue,
optimistic state, drafts, client identity, and WebSocket resume identity must remain
stable across extension transitions.

### ES-12 completion snapshot

ES-12 completed on branch `revision/grid-adapter` at baseline commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 154 accumulated status entries.
The web build now emits an exact, sorted asset-set manifest from the packaged bytes
and a final `standard` client-support registry bound to that manifest digest. Server
startup validates the archive, manifest, support registry, profile major/workspace
row, and empty capability arrays before binding the immutable publication plan. The
dynamic backend root injects that final registry outside the immutable asset mapping;
direct Vite development retains only the generated source projection.

Workbook startup now returns a no-store
`cartulary.extension_workspace_availability.v1` member derived from the resolved
claim set, workspace declaration, and current authorization. The browser computes
the exact discovery/support/authorization/availability intersection and protects
extension-originated requests with one locally generated epoch and linearized
generation reservation. Overflow rolls the epoch, random-source failure closes the
extension surface, stale completions cannot render, and capability activation always
returns `extension_capability_not_supported`.

The Workbook shell lazy-loads the Network Flow feature only after current eligibility
is established and falls back to Timeline when eligibility is lost. Extension state
is keyed to the availability epoch/revision, while Base caches, pending requests,
queue, optimistic state, drafts, the session-scoped client identity, and WebSocket
resume identity remain outside that disposal boundary. Network Flow request adapters
consume the same availability reservation rather than maintaining competing guards.

The exact BC-015/BC-016 Go rows and the additional stateful browser row
`module.extensions.browser_stateful.bc015_availability_continuity_d538000c38` are
active. Full Extensions evidence passed at `20260720T100506Z-p886531` (24 local
tests) and `20260720T100600Z-p902219` (12 service-backed tests), with no missing or
unmapped rows. The focused stateful scenario passed at
`20260720T100241Z-p865949`. Frontend typecheck, unit, import-boundary, web build,
backend unit, and module-boundary roots are respectively
`20260720T100409Z-p883885`, `20260720T100437Z-p884808`,
`20260720T100431Z-p884407`, `20260720T100656Z-p917170`,
`20260720T100817Z-p924637`, and `20260720T100800Z-p924374`. Generation drift,
JSON shape, generated policy, and Harness passed at
`20260720T100703Z-p919999`, `20260720T100715Z-p922923`,
`20260720T100724Z-p923304`, and `20260720T100735Z-p923531`.

Diagnosed intermediate failures were not accepted as evidence. Format roots
`20260720T092112Z-p735273`, `20260720T092143Z-p737849`,
`20260720T092219Z-p740465`, and `20260720T092241Z-p742999` exposed exact Biome
non-null, label, and hook-dependency violations. The backend-unit root
`20260720T092456Z-p748376`, frontend type roots
`20260720T092456Z-p748396`/`20260720T092652Z-p759521`, import-boundary roots
`20260720T092456Z-p748383`/`20260720T092735Z-p760183`, and frontend-unit roots
`20260720T092844Z-p763941`/`20260720T093149Z-p769444` exposed incomplete API,
facade-import, and lifecycle-test integration; all were fixed structurally. Harness
root `20260720T095424Z-p837515` exposed stale aggregate catalog counts after the
stateful row was added. The first focused browser root
`20260720T095834Z-p848185` timed out because it searched the Vite preview for a
production-backend bootstrap; the test now verifies the packaged bootstrap from the
backend root and lifecycle behavior from the preview root. Final reruns above pass.

### ES-13 pre-work input snapshot

ES-13 started only after the ES-12 tracker checkpoint passed `git diff --check` and
`make lint-markdown`; the post-update Markdown rerun exited zero. The branch remains
`revision/grid-adapter` at baseline commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 154 accumulated status entries.
ES-09 through ES-12 are `DONE`; ES-14 remains unstarted.

The exact staged owner-document digests remain Extensions `796e119f...`, Core 00–04
`3358cc76...`, `9ad58e13...`, `79803b2d...`, `79afa553...`, and `67583e75...`,
Network Flow `40c447d5...`, Reporting `72f8948f...`, Composition `489a839d...`,
OpenTelemetry `b31baa1e...`, and domain vocabulary `ed7c51b0...`. The Extensions
verification contract, five-family test manifest, and owner catalog digests are
`4af4ea2c...`, `2aa10f73...`, and `a49e684c...`.

This slice may execute tests and create explicit evidence-root manifests only. It
must resolve the complete affected-owner/target partition through public task-guide,
owner-explanation, and plan artifacts; run every active exact selector once; retain
paired accounting and owner-summary shards; and reject mixed snapshots, newest-run
selection, historical fallback, unexpected execution, unauthorized skips, missing
rows, and unmapped rows. It may repair product, test, or Harness defects discovered
by those exact runs, but it may not promote an owner document or mark an adoption
gate complete. The successful ES-12 roots are inputs for diagnosis, not substitutes
for ES-13 explicit-root accounting.

### ES-13 completion snapshot

ES-13 resolved the affected execution set through the public owner guides and owner
explanations, then executed every current row for 14 owners on the same source and
owner-document snapshot. The local and service-backed partitions are compatible:
`module.extensions` 24/12, `app.server` 23/19, `module.networkflow` 118/20,
`module.incidentbundles` 17/8, `module.reporting` 10/4, `module.workbook` 83/54,
`module.incidents` 36/25, and `module.evidence` 34/29. The local-only partitions are
`platform.config` 8, `platform.httpruntime` 1, `module.reportcomposition` 1,
`package.protocol_ts` 4, `web.networkflow` 31, and `web.workbook` 98. Every partition
passed with no failed, missing, skipped, unexpected, or unmapped row.

The final local roots, in the same order, are
`20260720T105259Z-p1225912`, `20260720T105138Z-p1207745`,
`20260720T105446Z-p1255871`, `20260720T105047Z-p1205232`,
`20260720T105100Z-p1206523`, `20260720T110310Z-p1316979`,
`20260720T111507Z-p1394454`, `20260720T111823Z-p1432645`,
`20260720T104924Z-p1199331`, `20260720T104925Z-p1199478`,
`20260720T104927Z-p1199621`, `20260720T104929Z-p1199924`,
`20260720T104932Z-p1200091`, and `20260720T104953Z-p1201638`. The eight
service-backed roots are `20260720T105351Z-p1240936`,
`20260720T105221Z-p1217143`, `20260720T105903Z-p1286714`,
`20260720T105114Z-p1207250`, `20260720T105126Z-p1207480`,
`20260720T110905Z-p1356986`, `20260720T111642Z-p1414352`, and
`20260720T112124Z-p1458096`.

The final evidence-class roots are backend unit
`20260720T112436Z-p1482685`, integration support
`20260720T112533Z-p1487724`, integration `20260720T112605Z-p1490074`, store
`20260720T112704Z-p1495495`, process `20260720T112737Z-p1498993`, frontend unit
`20260720T112436Z-p1482691`, browser support `20260720T104302Z-p1140896`,
accessibility `20260720T104410Z-p1144944`, visual
`20260720T104612Z-p1163634`, measurement `20260720T104829Z-p1182605`,
webserver-backed `20260720T112814Z-p1516700`, and stateful
`20260720T113319Z-p1543062`. Module-boundary, frontend-type, frontend-import,
vulnerability, targeted Gosec, OTel, and Harness roots are respectively
`20260720T114027Z-p1593166`, `20260720T114027Z-p1593179`,
`20260720T114027Z-p1593193`, `20260720T114027Z-p1593169`,
`20260720T113841Z-p1566027`, `20260720T113840Z-p1566006`, and
`20260720T113841Z-p1566049`.

Canonical explicit manifests under the ignored
`.cartulary/evidence-manifests/es13/` bind each owner to only those compatible exact
roots. Fourteen `make test-evidence-audit` invocations passed: Extensions
`20260720T113743Z-p1565100`; app server `20260720T113757Z-p1565235`; Evidence
`20260720T113759Z-p1565291`; Incident Bundles `20260720T113801Z-p1565339`;
Incidents `20260720T113802Z-p1565380`; Network Flow
`20260720T113804Z-p1565437`; Report Composition
`20260720T113807Z-p1565491`; Reporting `20260720T113808Z-p1565532`; Workbook
`20260720T113809Z-p1565589`; Protocol TS `20260720T113812Z-p1565636`; platform
configuration `20260720T113813Z-p1565683`; HTTP runtime
`20260720T113815Z-p1565734`; web Network Flow
`20260720T113816Z-p1565782`; and web Workbook
`20260720T113817Z-p1565829`. The audits reject newest-run lookup, mixed snapshots,
subset inference, stale roots, and historical fallback.

Two related failures were retained and repaired before the final roots were created.
The first webserver-backed root `20260720T102403Z-p1013633` found that the recovery
restore helper announced readiness before the already-bound listener activated the
publication epoch; the helper now binds, activates, and only then serves or reports
ready. Browser-support roots `20260720T103855Z-p1103154` and
`20260720T104120Z-p1122210` found an incident-support test reading the retired
phase-shaped authored index and expecting the old three-member discovery object; it
now consumes the generated registry and asserts the exact seven-member producer.
No compatibility reader or production relaxation was added. ES-13 is complete; the
158-criterion/all-clause static promotion audit and every adoption state remain
exclusively assigned to ES-14.

### ES-14 pre-work input snapshot

ES-14 started only after the ES-13 completion update passed `git diff --check` and
`make lint-markdown`; the final post-update rerun exited zero. The branch remains
`revision/grid-adapter` at baseline commit
`bf8427a55fdb5f2b8e919a9567baf48329b0da5b` with 156 accumulated status entries.
ES-00 through ES-13 are `DONE`, T-012 is `DONE`, and every remaining blocker is an
atomic-adoption or complete static-accounting obligation owned by this slice.

The exact pre-promotion owner-document digests remain Extensions `796e119f...`,
Core 00–04 `3358cc76...`, `9ad58e13...`, `79803b2d...`, `79afa553...`, and
`67583e75...`, Network Flow `40c447d5...`, Reporting `72f8948f...`, Composition
`489a839d...`, OpenTelemetry `b31baa1e...`, and domain vocabulary `ed7c51b0...`.
The Extensions verification contract, five-family test manifest, Harness catalog,
and authored input catalog digests are `4af4ea2c...`, `2aa10f73...`,
`ff7ac649...`, and `31ee760f...`. ES-13 exact-root evidence remains valid only for
this pre-promotion snapshot and will not be reused as post-promotion evidence.

This slice may change adoption/status, exact owner digests and manifest locators,
traceability/accounting sources, and downstream generated projections only as one
reviewable change set. It must first close all 158 acceptance-criterion and
all-normative-clause mappings; then regenerate; then promote the Extensions NLSpec
and every required companion together; then rerun exact affected-owner evidence,
static/source/drift/security/Harness gates, finalization, and broad release gates.
Any failure leaves the promotion unclaimed and is recorded with its exact root.

### ES-14 completion snapshot

ES-14 is `DONE`. The candidate atomic promotion is the adopted source state: the
Extensions NLSpec is `adopted/current` at raw SHA-256
`1d10578b2d10df2bfa17e3ab48d0cd3ad0cf36a1f0c285b8875f5df4109e2440`; Core 00
through Core 04 are
`106451bfa440fea1a1913e000ad1ac2d52cf58f112323371a8456337a21c19b9`,
`ae34ee625b8e4a12d2596554b779785c68bdeb6e264540b0d74efe23f7221df2`,
`85d0f16ceca81d0222ccaa96c55c0e8c4eab39e190e9e1801058bf7d6989651e`,
`7ec2ddf522549eaf5faa9c42fa1dcc2293bae2a541248731732685f30bc8bfa9`, and
`da7ff2d94e96125b5bf117e9eb35103fb2d96ea8cf12275c05db8447d832074d`.
Network Flow Activity is `adopted/current` at document `2.0.0`, contract major `2`,
and raw SHA-256
`14f8fdb92491ab60d567bae02fe8d5e9220a1edceb1699a644b8c9e91991973e`.
Reporting, Report Composition, OpenTelemetry, domain vocabulary, and Testing Harness
are `72f8948f5492264a198a6ffdfe3ec9c7bce820917361d728d400bafc9a139bc1`,
`489a839dbf57969111b1f633fc11833a3b2c4adccde847da32c85243c251f8ea`,
`b31baa1e2d5c2843f359041f6acf64dbd7c8f84b31ee102b2025ddcb72015b99`,
`62c8cb0f1b9dcf9e2f0ef3796e14675d4c2fa30c7e47e31d491771e3a3e5b701`, and
`61d120b42fac884034721510612af6fb6121662db67230292de693104fce1e89`.

The final authored Extensions catalog, verification contract, five-family manifest,
Harness owner catalog, and traceability mapping source are
`6ad82f7ca28d77065f064bad91ae61ac94d9005b4283feca617fde1c1a565d30`,
`4af4ea2cef53553e3445891e9c1c9f999157291dbaf11076df54327f1c5d20cb`,
`6f2f91b761c2ec9c7acf10215c4f0c437665e94d4e8e9b4e93d757549094985c`,
`a49e684c1e26522f8b08f1593c29fa8e099062912d78f3b386c0a137c67732ef`, and
`29bf7a0e49caf4fc01d569bef3d9b1530f18dfa45ffb65e153ffa5e1b1fe4399`.
The final Core 00-04, Network Flow, Reporting, Report Composition, OpenTelemetry, and
Testing Harness owner-manifest raw digests are, in that order,
`e032d688d8932a3a631b433de19a33a0c54c88e687256c1442128cf21f1013ae`,
`996805124a0186868894fb0c3e606f43121bc3642d59f3119ecc6478ff7be559`,
`5b82f7742f61b72fe6f0a638ca5b12811a8d88ce56c1506cc707446c4132cbf1`,
`94f8b03406a641533d35879121a82c85b1ae2a64e86e5cb7e6707dd005dbd794`,
`4cde78a565eba67a975cd10559a60750fc3bff5f5415f7fe3fbb95ad1dd57b7d`,
`50a291fa0c57db1db3cc9307f60e7d5716e1ce3845a6f244a7d53eb6616cc97c`,
`1a74ce98b652ad1a5b82e6ee59e78251a7924a607875a59483f7a8ccfa080d48`,
`6c3bc1949de842d17162f3f4bf403f61a39a2f2a94747fd8eef86e0d5096e8cb`,
`4fc1bb2defa5ba92b38f5b04bbc8f57e8b697b43163a6de90026721b0c2a4fa7`, and
`c86a42de1d0b4844f7cf98d1e74186d70ca45c09a0de4ad35e1bae0b79429902`.

Final owner evidence uses one exact compatibility identity: source snapshot
`sha256:d7acd6787c6cd4a2d92d5f1d0c9cd81ddee45dfbfa72487fed78fd83247dd4d6`,
catalog semantic digest
`sha256:09768162803b93e9d9148130f3386b1b7c333b8ef8fb8fbecb7008719a557d9a`,
and verification semantic digest
`sha256:c5268101b92eb89be1b9bc4e0565bddce7932175268cb499a0b94bedb6daee36`.
The 14 full-owner roots are, in owner order `module.extensions`, `app.server`,
`module.networkflow`, `module.incidentbundles`, `module.reporting`,
`module.workbook`, `module.incidents`, `module.evidence`, `platform.config`,
`platform.httpruntime`, `module.reportcomposition`, `package.protocol_ts`,
`web.networkflow`, and `web.workbook`: `20260720T131225Z-p2339362`,
`20260720T131315Z-p2354379`, `20260720T131350Z-p2363611`,
`20260720T131755Z-p2394096`, `20260720T131808Z-p2394701`,
`20260720T131817Z-p2395284`, `20260720T132415Z-p2434668`,
`20260720T132548Z-p2454200`, `20260720T132837Z-p2479318`,
`20260720T132838Z-p2479462`, `20260720T132840Z-p2479600`,
`20260720T132842Z-p2479754`, `20260720T132845Z-p2479932`, and
`20260720T132906Z-p2481454`. Exact selected/passed counts are 24, 23, 118, 17,
10, 83, 36, 34, 8, 1, 1, 4, 31, and 98, with zero failed, infrastructure-failed,
skipped, cancelled, unexpected, missing, or unmapped rows and paired accounting and
owner-summary shards.

The required service-backed roots for the first eight owners are
`20260720T133011Z-p2485017`, `20260720T133100Z-p2499836`,
`20260720T133133Z-p2508552`, `20260720T133543Z-p2538429`,
`20260720T133554Z-p2538642`, `20260720T133602Z-p2538860`,
`20260720T134155Z-p2576056`, and `20260720T134327Z-p2594196`, with exact
selected/passed counts 12, 19, 20, 8, 4, 54, 25, and 29. The independent backend
unit, integration-support, integration, store, and process roots are
`20260720T134832Z-p2619951`, `20260720T134916Z-p2623264`,
`20260720T134945Z-p2625611`, `20260720T135033Z-p2630630`, and
`20260720T135104Z-p2634028`; frontend unit is
`20260720T135151Z-p2651723`; browser support, accessibility, visual, measurement,
webserver-backed, and stateful are `20260720T135239Z-p2665772`,
`20260720T135340Z-p2669725`, `20260720T135533Z-p2688377`,
`20260720T135749Z-p2707238`, `20260720T135831Z-p2723836`, and
`20260720T140340Z-p2749815`. Boundary, type, frontend import, vulnerability,
targeted Gosec, and OTel roots are `20260720T134830Z-p2619781`,
`20260720T135151Z-p2651697`, `20260720T135151Z-p2651714`,
`20260720T140642Z-p2771217`, `20260720T140642Z-p2771213`, and
`20260720T140642Z-p2771212`.

Explicit manifests under `.cartulary/evidence-manifests/es14/` name only the exact
compatible target roots. The 14 audits passed with zero unused or rejected inputs and
exact required-target closure: app server `20260720T140822Z-p2793411`; Evidence
`20260720T140824Z-p2793468`; Extensions `20260720T140826Z-p2793515`; Incident
Bundles `20260720T140828Z-p2793556`; Incidents `20260720T140829Z-p2793614`;
Network Flow `20260720T140831Z-p2793661`; Report Composition
`20260720T140834Z-p2793711`; Reporting `20260720T140835Z-p2793759`; Workbook
`20260720T140837Z-p2793800`; Protocol TS `20260720T140839Z-p2793858`;
platform configuration `20260720T140841Z-p2793905`; HTTP runtime
`20260720T140842Z-p2793946`; web Network Flow `20260720T140843Z-p2793993`; and
web Workbook `20260720T140845Z-p2794044`.

Recovery retained every rejected or diagnostic root. `agent-finalize` without
`RESULTS_DIR` failed at `20260720T121738Z-p1655494` because new Go duration keys had
no baseline; the first preliminary `check` at `20260720T124703Z-p1859898` confirmed
the same coverage gap. Repository-owned aggregation and generation refreshed the Go
baseline; later preliminary checks exposed generated drift and three staticcheck
findings, which were repaired structurally. The successful preliminary check root is
`20260720T125846Z-p2061894`. Finalization rejected that root at
`20260720T130113Z-p2153247` for warm timing only. A hotter check passed at
`20260720T130242Z-p2153853`, but finalization root
`20260720T130445Z-p2238676` found a one-millisecond nested scheduler event-envelope
regression. The nested envelope now covers both completed work and every projected
event, with a deterministic regression test. The corrected full check passed 127/127
at `20260720T130806Z-p2242217`; `agent-finalize` then passed at
`20260720T131008Z-p2326085` and refreshed only the browser, Harness-smoke, and
service-backed duration baselines through repository tooling.

The four concurrent heavy-owner roots `20260720T121938Z-p1686491`,
`20260720T121938Z-p1686497`, `20260720T121938Z-p1686506`, and
`20260720T121938Z-p1686514` are rejected Harness evidence with
`scheduler_accounting_error`, not product failures. The earlier serial Incidents
root `20260720T122716Z-p1806950` is diagnostic only, and the interrupted Evidence
root `20260720T122856Z-p1825503` remains invalid because its paired authoritative
summary is absent; no summary was synthesized.

Post-finalizer static roots are generation drift `20260720T131102Z-p2331494`,
generated-artifact policy `20260720T131102Z-p2331533`, JSON shape
`20260720T131102Z-p2331499`, migration drift `20260720T131102Z-p2331547`,
Harness contract `20260720T131130Z-p2336673`, Markdown
`20260720T131130Z-p2336700`, and toolchain drift
`20260720T131130Z-p2336687`. Broad gates passed serially: `test-fast` with 651 tests
at `20260720T140910Z-p2794227`, `check` with 130/130 units and 510 tests at
`20260720T141130Z-p2836303`, `test` with 651 tests at
`20260720T141333Z-p2921954`, and `release-check` with 14/14 units at
`20260720T141740Z-p2992081`.

Migration `00034_extension_coordination.sql` remains forward-only. No down migration,
durable-state reinterpretation, compatibility alias, old discovery reader, partial
promotion, historical evidence fallback, or Core 05 claim boundary was introduced.

Authority is applied in this order:

1. adopted named-subsystem NLSpecs within their allocated behavior;
2. Core 00 through Core 04 for current implementation conformance;
3. Core 05 only if a genuinely claim-bearing timed, benchmark, fixture-sensitive,
   or publication boundary is later introduced;
4. `docs/domain.md`, the Testing Harness NLSpec, and implementation-support guides
   for terminology and mechanics;
5. current code, authored contracts, tests, and generated outputs for current state;
6. this tracker and the planning framework as non-authoritative handoff material.

No adopted-owner contradiction was found during implementation, coordinated
promotion, or final evidence execution. The former differences between the draft and
the Network Flow/Core discovery contracts were resolved by the atomic adoption. Any
future contradiction discovered during companion amendment review must be recorded
exactly as `BLOCKED: owner contradiction` and must not be resolved by
implementation convention.

### Owner documents inspected

- `AGENTS.md`; the `refactor-tracker` skill, its tracker-format reference, and the
  modular refactor planning framework.
- `docs/extension-subsystem-nlspec.md`, including the current 236 `EXT-REQ-*` IDs,
  continuous `EXT-AC-001` through `EXT-AC-158`, Section 27 artifacts, and
  `EXT-GATE-001` through `EXT-GATE-028`. ES-01A added exactly the selected 17
  draft criteria without allocating a requirement or gate ID.
- Core 00 through Core 04: `docs/spec/00_document_set_status_and_precedence.md`,
  `01_architecture_storage_and_view_contracts.md`,
  `02_domain_model_schema_and_history.md`,
  `03_workbook_interaction_collaboration_and_workflows.md`, and
  `04_security_deployment_and_conformance.md`.
- `docs/testing-harness-nlspec.md`, `docs/domain.md`,
  `docs/network-flow-activity-nlspec.md`,
  `docs/reporting-subsystem-nlspec.md`,
  `docs/report-composition-nlspec.md`, and
  `docs/opentelemetry-instrumentation-nlspec.md`.
- Core 00/01/02/03/04 Incident Portability and physical-backup sections. There is no
  separate adopted Incident Portability NLSpec in the inspected repository; its
  current behavior is distributed across Core owners and `internal/modules/incidentbundles`.

### Repository evidence inspected

- `temp/analysis-notes.md` as a non-normative boundary-completeness input; it is not
  an owner document, contract, or implementation-conformance source.

- Every file under `internal/modules/extensions`, including its testsupport package.
- `internal/platform/httpapi/extensions.go`, `httpapi.go`, and their tests;
  `internal/app/server/runtime.go`, `runtime_routes.go`, and route-composition tests;
  platform configuration, jobs, object-store, PostgreSQL, and telemetry boundaries.
- Workbook startup registry/routes/tests; `apps/web/src/app/App.tsx`,
  `app/api/appShellClient.ts`, the incident-directory debug harness, relevant browser
  support tests, and Network Flow browser/workbook consumers.
- `contracts/extensions/index.json`, `contracts/index.json`, the OpenAPI extension
  route/schema, `tools/contractgen`, generated Go/TypeScript projections, and
  `tools/generated_artifact_policy.json`.
- Network Flow module facade, routes, stores, import facade, transaction participants,
  configuration, migrations `00028`, `00029`, `00030`, and `00033`, and its current
  verification owner/family rows.
- Reporting, Report Composition, Incident Bundle, jobs, backup/configuration, and
  telemetry implementation surfaces named by searches and opened before conclusions.
- Harness v2 verification registry, test-owner registry, family manifests, runner
  registry, runtime/resource/fixture profiles, execution topology, render index,
  schema attachments, and evidence-audit command surface.

### Fixed planning assumptions and dependencies

- BC-001 through BC-017 are selected and are not reopened by this tracker revision.
- Network Flow's empty authoritative state is valid and therefore selects
  `empty_state_policy=allowed`.
- Portability import mutates only through the shared transaction protocol.
- Restore v1 targets a stopped empty deployment.
- Browser contract v1 has exactly one `standard` build class.
- Capability advertisement remains entirely disabled in v1.
- PostgreSQL advisory locking is permitted supporting implementation guidance, not a
  normative mechanism or required storage coupling.
- The system remains statically packaged; runtime-downloadable executable extensions
  remain deferred.
- Existing requirements are amended in later normative work. No new `EXT-REQ` or
  `EXT-GATE` ID is allocated; exactly 17 acceptance IDs are planned.

### Explicit non-goals

- No production, test, owner-specification, contract, schema, generated-artifact,
  dependency, migration, configuration, harness, fixture, or lockfile change.
- No runtime package installation, marketplace, arbitrary callback bus, separate
  extension host, or independently distributed executable extension format.
- No transfer of Core 00 recognition/claimability, Core 01 public discovery/dispatch,
  Reporting, Report Composition, Incident Portability, backup, or OpenTelemetry
  ownership to `internal/modules/extensions`.
- No phase identity, delivery phase, test row, fixture family, or adoption gate used
  as runtime architecture.
- No `EXT-FIX-*`, extension fixture-result schema, v1 alias, compatibility reader,
  newest-run lookup, or historical retained-run fallback.

## 2. Current-State Repository Inventory

All three target files are accounted for below. Adjacent rows are included because
the current responsibility is materially distributed outside the target path.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `internal/modules/extensions/api.go` | Constructs the current three-member discovery item and `{extensions: [...]}` response data. | `BuildResource`, `BuildResponseData` | `routes.go`; response behavior is exercised through server and browser tests. | `internal/platform/httpapi.ExtensionProfile` | Incident HTTP/integration/request suites and `apps/web/e2e/incident.support.spec.ts` through the route. | Current OpenAPI resource and generated Go/TS types. | Split: Extensions normalization facade plus Core 01 discovery producer. | high | Current shape is `profile_id`, `claimed`, `route_families`; draft target is the coordinated generic seven-member item. |
| `internal/modules/extensions/routes.go` | Registers authenticated `GET /api/v1/extensions`, validates singleton query shape, slides the session, and writes common envelopes/errors. | `Service`, `RegisterRoutes` | `internal/app/server/runtime.go` built-in route contribution. | `authn`, `httpauth`, `httpapi`, PostgreSQL handle, master keys, clock. | Process/integration HTTP suites, browser support route checks, and common auth/envelope tests. | OpenAPI path and generated clients/types. | Split: Core 01 route contract; platform auth/session adapter; small Extensions query facade. | high | The package directly composes platform authentication and transport concerns; no registry generation or claim-resolution logic lives here. |
| `internal/modules/extensions/testsupport/routetest/routes.go` | Contributes the discovery route to shared route-inventory assertions. | `PublicDiscovery` | `internal/modules/incidents/http_conformance_test.go` | `internal/testutil/routeinventory` | Shared HTTP conformance suite. | None. | Keep as owner-aligned testsupport only if `module.extensions` owns the row; otherwise move to Core 01/app-server testsupport. | medium | Current entry freezes method/path/envelope only, not the target seven-member response. |
| `internal/platform/httpapi/extensions.go` | Holds the process-global recognized profile list, claim flags, route families, workspace metadata, reservation matching, cloning, and testing override. | `ExtensionProfile`, `ExtensionWorkspace`, profile resolution/claim/reservation functions | Server assembly, route wrapper, Network Flow, workbook startup, tests. | Standard library only. | `internal/platform/httpapi/httpapi_test.go`, workbook and module tests. | Parity-tested against `contracts/extensions/index.json`. | Split: Core 00 registry facts, Core 01 route reservation, Core 03 workspace declarations; generated Extensions consumers. | critical | Current code creates recognition from implementation data, which the draft forbids after adoption. Core ownership must not be transferred. |
| `internal/platform/httpapi/httpapi.go` | Builds the HTTP mux and wraps it with unclaimed reserved-family dispatch before common handling. | `DependencySet`, `RouteRegistrar`, `NewHandler`, envelope/error APIs | Server runtime and test harness composition. | Web assets, config, jobs, object store, PostgreSQL, telemetry, WebSocket. | HTTP API tests cover reserved-unclaimed outcomes. | OpenAPI/common error contracts. | Core 01/platform HTTP runtime. | critical | Current reserved-unclaimed wrapper returns `404 extension_profile_not_claimed` before the inner handler. Exact precedence must be characterized before replacement. |
| `internal/app/server/runtime.go` and `runtime_routes.go` | Resolves profile claims, feeds telemetry, constructs Network Flow, registers the Extensions route in fixed built-in order, and injects profiles into HTTP/workbook dependencies. | `NewRuntime`, `Runtime`, route composition helpers | `cmd/server`, server harness/tests. | Modules and platform runtime dependencies. | Runtime route-order, profile, process, and integration tests. | None directly. | Application assembly; future Extensions coordinator is injected here, not implemented here. | critical | `network_flow_activity` and enterprise-auth claims are applied from config; current assembly is not the draft's six-stage atomic publication coordinator. |
| `contracts/extensions/index.json` | Authored derived registry input with recognized profile IDs and route families. | `registry_id=cartulary.extensions.phase2.v1` | Contract generator and parity tests. | Core 00 owner sections declared in `contracts/index.json`. | Contractgen validation and `httpapi` parity test. | Generates `internal/gen/contracts/**` and `packages/protocol-ts/src/generated/**`. | Replace through owner inputs plus Extensions generator after owner adoption. | critical | Historical phase-shaped identity must not survive adoption or become a compatibility alias. |
| `contracts/openapi/cartulary.openapi.yaml` | Owns the current discovery route and three-member resource schema. | `/api/v1/extensions` and generated public types | Server/client generated consumers. | Core 01 contract ownership. | OpenAPI, backend route, browser support, reporting boundary tests. | Go/TS contracts and validators. | Core 01. | critical | Target producer and decoder changes require coordinated contract-major adoption. |
| `tools/contractgen/**`, `contracts/index.json`, generated roots | Validates the single current extension index and emits Go/TS embedded contracts. | Contract-family `extensions` | `make generate`, drift checks, builds. | Authored contracts and family registry. | Contractgen and generated-artifact checks. | `internal/gen`, protocol TS generated files. | Extensions generator plus contract-generation platform. | critical | Current validation admits only `contracts/extensions/index.json`; Section 27 requires a much larger generated artifact family. Generated roots must never be hand-edited. |
| `apps/web/src/app/api/appShellClient.ts`, `apps/web/src/app/App.tsx`, debug harness | Fetches and consumes current discovery during app startup and uses claim state for feature availability. | `ExtensionProfileResource`, `fetchExtensions` and application state | Web application startup and debug surface. | Generated protocol types and HTTP client helpers. | App landing/unit tests and browser support. | Generated TS discovery type. | Core 03/web application, consuming a generated client support registry. | critical | Current browser does not gate on digest-bound support registry plus availability epoch/generation. |
| Workbook startup backend and web models | Validates `extension_workspace` sheet refs, declared workspace, claim state, and minimum role; renders Network Analysis when eligible. | Startup DTOs/registry and URL state | Workbook routes/controllers/shell. | `httpapi.ExtensionProfile`, auth/membership, generated startup contracts. | Backend startup tests, workbook unit/browser tests. | OpenAPI startup schemas. | Core 01/03 and web application. | high | Must preserve Base startup, drafts, pending queue, WebSocket identity, and fallback semantics. |
| `internal/modules/networkflow/**` | Owns Network Flow resources, routes, import facade, stores, transaction participants, graph adapter, security, and claimed-route registration. | `Module`, `NewModule`, `ImportOwner`, `RegisterRoutes`, profile constants | Server assembly, imports, browser workspace. | PostgreSQL, Imports port, Incidents/Indicators participants, key rings. | Unit/store/integration/browser/accessibility/visual/measurement rows. | Network Flow contracts and generated consumers. | `module.networkflow` named profile owner. | critical | Adopted v1.2.0/major 1 currently competes with the draft's proposed generic discovery and major-2 action. Profile behavior stays owned here. |
| `db/migrations/00028`, `00029`, `00030`, `00033` and Network Flow queries | Current Network Flow import linkage, authoritative tables/rows/diagnostics/bindings, and indexes. | SQL schema | Network Flow/Imports stores. | PostgreSQL. | Store and integration tests; migration drift. | sqlc outputs where applicable. | Network Flow state owner plus migration application facade. | critical | No generic extension metadata, migration ledger, state-presence, or physical-binding schema was found. Future authored migrations require later authorization. |
| `internal/modules/incidentbundles/**` and Core Incident Portability sections | Whole-incident export/import, bundle files, attribution, jobs, and publication. | Route/store/worker facades | Server assembly and revision assembly. | Object store, jobs, incidents, revisions. | API/integration/worker tests. | Incident-bundle OpenAPI/contracts. | Incident Portability/Core 01 shared owner. | high | Draft participation must be typed; inactive blocking must be declarative and must not execute profile code. |
| `internal/modules/reporting/**` and `internal/modules/reportcomposition/**` | Snapshot/report render/release and authored composition behavior. | Reporting and composition routes/services/providers | Server, jobs, browser reporting surfaces. | PostgreSQL, graph projection, object store, jobs. | Unit/integration/traceability/OpenAPI tests. | Reporting/composition contracts and generated outputs. | Existing named owners. | high | Generic participation may be imported, but ownership cannot move to Extensions. Record no-change parity when interfaces do not change. |
| Platform config, jobs, object store, PostgreSQL, telemetry | Current claim keys, job shells, storage adapters, transactions, and claimed-profile telemetry serialization. | Platform facades | Server and modules. | External services and standard libraries. | Platform unit/integration/conformance tests. | Telemetry contracts and config examples. | Core 01/Core 04 and platform owners. | critical | `cartulary.profile.claims` currently uses a hard-coded known-profile set; target derives it from the canonical resolved claim set and digest. |
| Harness v2 registries and topology | Routes exact test evidence by owner/family/verification/runner/profile. | Verification registry, owner registry, family manifests, runner registry, execution topology. | Public Make targets and owner-slice scheduler. | Authored JSON plus generated topology projections. | Harness contract/json-shape/generated-drift suites. | Generated task surface, render index, browser batch manifests. | Testing Harness v2. | critical | No `module.extensions` owner exists. `task-guide` and `explain-test-owner` fail for it today. |

## 3. Module Boundary Diagnosis

The current target is a shallow transport facade, not a durable subsystem boundary.
The proposed subsystem needs a small coordination facade around generated immutable
inputs while leaving fact ownership, transport/runtime plumbing, named-profile state,
and shared-owner behavior with their existing authorities.

### Current-state versus proposed-target boundary map

| Concern | Current state | Proposed target boundary | Decision |
| --- | --- | --- | --- |
| Recognition and claimability | Hard-coded in `internal/platform/httpapi/extensions.go`. | Core 00 remains the sole fact owner; generated owner inputs are consumed by Extensions. | split |
| Discovery | Target package builds three fields; Core OpenAPI owns the route. | Core 01 owns producer/decoder/route; Extensions supplies validated descriptor-derived data. | split |
| Claim resolution | Server applies selected config claims; modules self-check claim flags. | Extensions coordinator executes the closed dependency/admission algorithm using Core 04 configuration inputs. | move |
| Route dispatch | HTTP wrapper matches hard-coded reserved families; claimed modules register directly. | Core 01 owns Base reservation registry, overlap rules, and public dispatch precedence; Extensions supplies registry facts. | split |
| Registry generation | One phase-shaped authored index plus narrow contractgen validator. | Extensions generator consumes digest-bound owner manifests/fragments and emits canonical registry/integrity artifacts. | move |
| Implementation bindings | Go assembly and constants imply availability. | Build-owned packaged bindings must match descriptors and integrity digests; runtime consumes bindings only. | split |
| State ownership | Network Flow owns its SQL; no generic state metadata/ledger exists. | Named profiles own authoritative families; Extensions owns generic metadata/ledger contracts and scoped coordination. | split |
| Migration | Ordinary DB migrations plus profile code; no generic extension migration coordinator. | Profile owners author definitions; Extensions coordinates locks, scoped contexts, pending validation, ledger, and exact-once final validation. | split |
| Jobs | Platform common manager; modules own job behavior. | Core 01 retains common shell; profile owners define job-kind contracts; Extensions coordinates proof/reconciliation rules. | split |
| Transactions | Network Flow composes PostgreSQL/incident/audit/indicator participants directly. | Core 01 owns bounded cross-owner protocol/final commit; participants remain with semantic owners. | split |
| Staged objects | Object-store adapters and job/module code. | Core 01 object-storage owner controls access cutoff/cleanup; Extensions supplies typed staged-object contract and fatal conditions. | split |
| Backup/restore | Core configuration and incident-bundle paths; profile SQL is physically included by deployment backup. | Physical backup owner orchestrates codecs/bindings; profile owners declare state; Extensions validates parity and ordering. | split |
| Portability | Incident Bundle owns public export/import. | Incident Portability owns operation; Extensions supplies closed participant interfaces and declarative inactive blocker. | split |
| Reporting | Reporting/Composition modules own behavior. | Existing owners retain behavior and import generic participant interfaces when needed. | keep |
| Browser | App startup consumes discovery; workspace shell uses claim/workspace facts. | Core 03/web app intersects discovery, client-support registry, authorization, and no-store availability epoch/generation. | split |
| Security/configuration | Core config/platform secret handling plus module-specific checks. | Core 04 owns claim keys, syntax-only inactive policy, timeouts, lease, readiness, fatal lifecycle; profiles own local schemas. | split |
| Diagnostics | Common API errors plus module diagnostics; no canonical extension condition registry. | Extensions generator emits validation-condition registry; Core 04 owns startup/fatal presentation and exit behavior. | split |
| Observability | Telemetry receives a server-derived claimed-profile string. | OpenTelemetry owner derives the signal from the canonical resolved claim set/digest; Extensions exposes no secrets. | split |
| Conformance accounting | Existing owners only; no `module.extensions`. | Static Extensions accounting plus Harness v2 owner/selector/evidence accounting, without runtime coupling. | move |

### Responsibility map

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Authenticated collection route | `extensions/routes.go` | Core 01 route plus platform auth adapter | split | Core 01 owns the public route/envelope; current code imports auth/platform directly. | Preserve method, query rejection, auth, session sliding, envelopes until coordinated change. |
| Discovery data normalization | `extensions/api.go` | Extensions facade | keep | Target draft allocates descriptor consumption to Extensions. | Input must become generated descriptors, not hard-coded profile structs. |
| Recognized profile vocabulary | `httpapi/extensions.go` | Core 00 | move | Core 00 current owner and draft Table 1-A. | Remove implementation-created recognition only after generated parity exists. |
| Route reservation matching | `httpapi/extensions.go` and HTTP wrapper | Core 01/platform HTTP | move | Current public dispatch behavior and Core 01 ownership. | Must consume generated Base and extension registries. |
| Workspace declarations | `httpapi.ExtensionProfile.Workspaces` | Core 01 identity plus Core 03 behavior | split | Workbook registry consumes the field; draft separates identity and rendering. | Browser support registry adds build compatibility, not ownership. |
| Claim config application | `server.applyConfigExtensionClaims` | Core 04 plus Extensions coordinator | move | Only two profiles currently read explicit config here. | No partial serving during transition. |
| Named-profile implementation | Network Flow, Imports, Incident Bundle, Reporting, etc. | Existing named module owners | keep | Current adopted owners and repository module boundaries. | Extensions coordinates closed interfaces only. |
| Application composition | `internal/app/server` | Application facade | keep | Repository procedure designates exact composition root. | Inject the coordinator and bindings; do not add product logic. |
| State metadata/migration ledger | absent | Extensions module with profile-owned definitions | split | Draft Sections 19/21/27; no current generic schema found. | Authored SQL/migrations are later behavior-changing work. |
| Cross-owner transaction engine | PostgreSQL/module-specific composition | Core 01/platform transaction coordinator | move | Draft explicitly leaves final-commit protocol with Core 01. | Extensions supplies participant contracts, not a second transaction engine. |
| Backup codec/physical binding | deployment backup and module stores | Backup platform plus durable profile owners | split | Draft Table 27-A and Core physical-backup ownership. | Filesystem remains derived-only for extension state. |
| Client lifecycle state | `apps/web` controllers/shell | Core 03/web application | keep | Browser state is application behavior. | Extensions provides generated support facts only. |
| Telemetry claim serialization | `internal/platform/telemetry` | OpenTelemetry/platform telemetry | keep | Adopted OTel NLSpec owns signal shape. | Replace hard-coded known set only in coordinated adoption. |
| Contract generation | `tools/contractgen` | Contract tooling plus Extensions generator | split | Existing family generation and Section 27 target artifacts. | Owner inputs authored; projections generated. |
| Harness evidence routing | Harness v2 registries/topology | Harness owners and `module.extensions` test owner | split | Current registries lack the owner; draft defines two verification IDs. | Test rows are evidence accounting, never runtime architecture. |
| Independent executable package support | absent | Future NLSpec | defer | Draft Section 30 explicitly makes it future-only. | Do not reserve a format or compatibility reader. |

## 4. Public Contract and Behavior Freeze Map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `GET /api/v1/extensions` method/path | Core 01; target route facade | OpenAPI and `RegisterRoutes` | Shared route inventory; browser support request | Exact GET success, non-GET 405, unknown query rejection, request ID, auth/session sliding | high | Behavior-preserving until the coordinated discovery-major change. |
| Current discovery envelope | Core 01 | `{data:{extensions:[...]},meta}` with three fields | Incident/browser support tests | Snapshot every recognized/claimable/claimed ordering and omitted forbidden fields | critical | Target seven-member shape is `requires later authorization`. |
| Tolerant discovery decoder | Core 01/web | Current generated type and app casts | App landing/unit tests | Known-member malformation rejection; additive unknown-member ignore; no unknown execution | critical | Producer stays strict while decoder becomes tolerant only with owner amendment. |
| Reserved-unclaimed routes | Core 01/platform HTTP | `withUnclaimedReservedExtensionFamilies` | `httpapi_test.go` and Network Flow route tests | Every exact/ancestor/descendant route; verify precedence before auth, incident, payload, cursor, job, resource lookup | critical | Preserve `404 extension_profile_not_claimed` until adopted owner changes exact token/status. |
| Claimed-family dispatch | Core 01 plus named profile | Network Flow registers only when claimed | Network Flow integration/browser tests | Authorization denial and empty-resource outcomes never collapse to unclaimed | critical | Profile-local errors begin after Core dispatch. |
| Recognition/claimability | Core 00 | Hard-coded six-profile list plus Core registry | Contract parity and config tests | Recognized-unclaimable, zero-profile generation, no implementation-created profile | critical | Core 00 remains sole owner. |
| Claim configuration and resolution | Core 04/Extensions target | Server config application | Core config and startup tests | Omitted/false/null/type cases, dependency order/cycles, preflight, no listener/worker on failure | critical | Characterize exit `2` and no partial profile set. |
| Startup publication and fatal lifecycle | Core 04 | Current runtime construction/cleanup | Process/bootstrap tests | Second process, crash release, lease loss at three stages, Stage 6 atomic serving, repeated fatal exit `70` | critical | No current complete implementation was found. |
| Implementation binding admission | Build/Extensions target | Current Go composition only | Module construction tests | Missing/extra/major/digest/capability/state/job/participant mismatch before migration | critical | Packaged binding artifact is new. |
| Migration state and results | Profile owners/Extensions target | Network Flow SQL migrations only | Migration/store tests | Missing path, too new/old, immutable digest, lock/step/profile timeout, resumability, exact-once final validation | critical | Closed scoped contexts; no cross-owner access. |
| Cross-owner transaction result | Core 01 | Current Network Flow transaction participants | Store/integration tests | Every ordered failure/cancel position, bounds, lock order, conflict/timeout/commit/no-partial-effects | critical | No automatic retry. |
| Staged-object lifecycle | Core 01 object storage | Object store/job paths | Object-store and job tests | Access cutoff independent of cleanup, batching, every deletion outcome, retry, readiness degradation, fatal contradiction | critical | Publication must be atomic with final DB commit. |
| Job proof/reconciliation | Core 01 jobs plus profile owner | Current common manager/runner | Jobs and module tests | Proof required/prohibited, replay, cancellation precedence, contradictory proof, inactive reconciliation | high | Internal owner profile identity is mandatory. |
| Backup/restore | Physical backup owner/profile owner | Config backup tests and module storage | Backup/config/store tests | Claimed/unclaimed state, codec ID/digest, empty binding, historical codec, unsupported codec, restore ordering | critical | Inactive restore executes no profile code. |
| Incident portability | Core 01/Incident Bundle | Incident-bundle module | Bundle API/integration/worker tests | Every claim/state matrix, declarative inactive blocker, pre-publication failure, no profile execution inactive | high | Preserve public bundle ownership. |
| Reporting participation | Reporting/Report Composition | Adopted NLSpecs/modules | Reporting/composition tests | No-participation omission and typed participant matrices | high | Record no-change parity if generic interfaces do not change their owner contracts. |
| Workbook extension workspace | Core 03/web | Startup registry and shell | Backend/web workbook tests | Discovery/support/auth intersection, lazy load, fallback, unsupported major, stale generation, authorization loss | critical | Preserve Base tabs, caches, requests, queue, optimistic state, drafts, and stable client/WebSocket identity. |
| WebSocket behavior | Core 03/platform WS | Base collaboration implementation | Collaboration/workbook tests | Unknown extension values, authorization/session loss, epoch rollover without Base identity reset | high | Draft adds no arbitrary extension event bus. |
| Security and egress | Core 04/profile owners | Config/secret/Network Flow security | Config/security/telemetry tests | Inactive syntax-only no resolution/DNS/connect/profile code; egress-none across all call sites; secret-negative assertions | critical | Diagnostics and telemetry must remain content-free. |
| Telemetry claim identity | OpenTelemetry | `SerializeProfileClaims` and OTel NLSpec | Telemetry resource/privacy/conformance tests | Canonical resolved set/digest, unknown profile rejection, ordering, no secret or incident content | high | OTel remains signal owner. |
| Generated contracts | Owner specs/contracts/tooling | Contract family and generated policy | Generate/json-shape/drift tests | Full Section 27 schema/identity/digest/parity/limit vectors | critical | Never hand-edit generated roots. |
| Harness owner accounting | Testing Harness v2 | Current registries/topology; absent owner | Harness contract tests | Exact row resolution, full-owner versus subset, paired shards, evidence-class gates, exact-root audit, stale-root rejection | critical | Aggregate `check` success is insufficient. |
| Core 05 publication | Core 05 | No draft publication behavior | Existing benchmark claim check | None unless a later owner introduces a claim-bearing timed/fixture-sensitive boundary | low | Out of scope for current adoption. |

## 5. Coupling and Boundary Findings

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| Profile recognition, claim state, reservations, and workspaces share one platform struct. | `internal/platform/httpapi/extensions.go` | Ownership drift and implementation-created facts. | must_fix | Core 00/Core 01/Core 03 plus generated Extensions inputs | Split facts by owner and consume an immutable generated registry. |
| The target route facade imports auth store, master keys, HTTP auth, and transport response helpers directly. | `internal/modules/extensions/routes.go` | A future deep module could become transport/platform coupled. | should_fix | Platform HTTP/auth adapter plus Extensions query facade | Keep authentication/session behavior in route adapter and expose a narrow discovery query port. |
| Server assembly performs partial claim application and telemetry derivation before any generic coordinator exists. | `internal/app/server/runtime.go` | Partial publication and inconsistent claim consumers. | must_fix | Application assembly injecting Core 04/Extensions coordinator | Introduce one resolved claim-set result and atomic publication plan in a later slice. |
| Contract identity is phase-shaped. | `registry_id=cartulary.extensions.phase2.v1` | Historical delivery identity could leak into runtime and compatibility. | must_fix | Extensions generator/Core 00 contract family | Replace after owner inputs are adopted; do not alias or compatibility-read it. |
| Generated contract validation permits only one narrow extension artifact. | `tools/contractgen/validation.go` | Section 27 artifacts cannot be represented or drift-checked. | must_fix | Contract tooling/Extensions generator | Add authored schemas/owner inputs first, then generator outputs and drift checks. |
| Network Flow currently owns a competing discovery shape/major 1 while the selected target requires major 2. | Adopted Network Flow NLSpec and current contracts | Premature target implementation would break adopted behavior. | must_fix | Core 00/01/03/04 plus Network Flow owner | Follow the ten-step coordinated amendment and promotion order; no intermediate adoption. |
| Current browser trusts discovery and local code support without a digest-bound support registry/availability generation. | App startup and workbook shell | Stale or unsupported UI can render. | must_fix | Core 03/web application/client build | Add generated support registry and no-store availability epoch/generation atomically. |
| Generic extension state metadata, migration ledger, presence manifest, physical binding, and codec artifacts are absent. | SQL/config/contract searches | State adoption cannot be proven by code alone. | must_fix | Extensions generator, profile owners, backup owner | Author schemas/migrations only after companion owner authorization. |
| Cross-owner transaction semantics are implemented profile-by-profile. | Network Flow module/store participant composition | Inconsistent deadlines, ordering, replay, and final commit. | must_fix | Core 01 transaction coordinator | Characterize current behavior, then adopt one bounded protocol with typed participants. |
| Incident portability/reporting/backup interfaces are not generic Extensions participant contracts. | Current modules and owner docs | Ownership transfer or implicit participation risk. | should_fix | Existing shared owners | Amend only imported interfaces; otherwise record exact no-change parity. |
| Telemetry validates a hard-coded profile vocabulary. | `internal/platform/telemetry/resource.go` and privacy registry | Drift from Core 00 registry. | must_fix | OpenTelemetry/platform telemetry | Derive from canonical claim set and digest after adoption. |
| `module.extensions` is absent from every owner-first Harness input. | Verification/owner/family registries and topology; Make failures | No owner slice or evidence audit can close adoption. | must_fix | Harness v2 plus `module.extensions` | Add both verification IDs, nonempty exact-selector family, topology, and projections together. |
| Framework catalog does not list an `extensions` module. | Planning framework module table versus live target path | Treating framework as repository truth would erase the target seam. | intentional/no_action | Tracker | Record the mismatch and use live repository evidence. |
| No dynamic executable extension system exists. | Code/config/route searches and draft non-goals | Accidental scope expansion. | intentional/no_action | Future owner | Preserve absence; defer any package format to a future NLSpec. |
| No direct vendor-grid import was found in the Extensions backend target. | Target file inspection | None for this seam. | intentional/no_action | Grid adapter/web owners | Keep vendor semantics out of Extensions work. |

## 6. Refactor Workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Authority and baseline freeze | root | none | WF-01 | Freeze exact draft, adopted owners, worktree, public contracts, and source digests. | Owner docs, tracker | Markdown/source lint; recorded digests | No source ambiguity and no owner contradiction selected. |
| WF-01 | Characterization closure | chain | WF-00 | WF-02, WF-03, WF-04 | Add exact pre-change evidence for every observable current boundary. | Existing backend/web tests and future owner rows | Narrow current owner slices | Every risky move has pre-change evidence. |
| WF-02 | Companion-owner coordination | chain | WF-00, WF-01 | WF-02A | Amend Core/profile/shared owners or record exact permitted no-change parity. | Core 00-04, Network Flow, Reporting, Composition, OTel, domain, Harness | Normative-source and documentation checks | All owner versions/digests/anchors are adoption-ready; draft remains draft. |
| WF-02A | Normative boundary closure | chain | WF-02 | WF-03, WF-04, WF-05, WF-06, WF-07, WF-08 | Apply BC-001 through BC-017 to the draft and companion-owner amendment set; assign exact owner anchors, schemas, algorithms, acceptance criteria, and selectors while the Extensions NLSpec remains draft. | Draft, Core/profile/shared owner documents, traceability inputs | Normative-source, traceability, and documentation checks | Every BC row has exact target anchors and an active acceptance/verification mapping; no generation or implementation starts from the pre-closure draft. |
| WF-03 | Owner manifests and canonical generation | chain | WF-02A | WF-04, WF-08 | Produce digest-bound inputs, descriptors, registry, integrity, closure, schemas, and projections. | `contracts/**`, generator inputs, generated roots | `make generate-drift`, JSON shape, artifact policy | All artifacts are generator-owned and byte-stable. |
| WF-04 | Runtime coordination and publication | chain | WF-01, WF-02A, WF-03 | WF-05, WF-06, WF-07 | Implement claim resolution, admission, bindings, dependencies, lease, publication, diagnostics, and fatal lifecycle. | Extensions facade, app server, config/platform adapters | Backend unit/process/integration; security | No listener/worker/route is partially published. |
| WF-05 | Core discovery and dispatch | parallel | WF-02A, WF-03, WF-04 | WF-07, WF-09 | Move producer/decoder/reservation/dispatch to generated generic contracts without transferring Core ownership. | Core OpenAPI, HTTP runtime, Extensions query facade | Go integration/browser support | One seven-member producer and one dispatch precedence. |
| WF-06 | State, migration, jobs, transactions, and storage | parallel | WF-02A, WF-03, WF-04 | WF-08, WF-09 | Add scoped generic coordination while preserving profile-owned state and Core shared protocols. | Extensions, Core 01 transaction/storage/jobs, profile stores/migrations | Service-backed owner slices, migration drift | State and commit outcomes satisfy every closed matrix. |
| WF-07 | Browser support and lifecycle | parallel | WF-02A, WF-03, WF-04, WF-05 | WF-09 | Add support registry, availability generation, eligibility intersection, fallback, and cleanup. | Web app, generated protocol/UI contracts, workbook startup | Typecheck, unit, browser/stateful/a11y/visual as allocated | Base state remains stable and stale extension state cannot render. |
| WF-08 | Named-profile and shared-owner adoption | parallel | WF-02A, WF-03, WF-06 | WF-09 | Adopt Network Flow major 2 and typed portability/reporting/backup participation without ownership transfer. | Network Flow, Incident Bundle, Reporting, Composition, backup | Affected full-owner slices | Every participant and parity row resolves to its primary owner. |
| WF-09 | Harness v2 onboarding and traceability | chain | WF-03, WF-05, WF-06, WF-07, WF-08 | WF-10 | Register owner-first contracts, exact selectors/profiles, clause traceability, paired shards, and evidence audit. | Verification/owner/family/topology authored inputs and generated projections | Harness contract, task guide, full-owner slices, evidence audit | Both Extensions verification IDs, all planned boundary acceptance criteria, and every imported-owner obligation exist and resolve. |
| WF-10 | Atomic adoption and final handoff | chain | WF-09 | none | Execute all gates, audit retained roots, and promote all companions together. | Owner docs and status metadata only after evidence | Full drift/check/release gates plus exact-root audit | All 28 gates are `DONE`; no intermediate artifact claimed adoption. |

No schema generation, implementation, browser conversion, or participant work may
begin from the pre-closure draft. WF-03 through WF-08 MUST consume the WF-02A
closure set, and WF-09 MUST reject any planned boundary criterion that is absent,
unresolved, or lacks an exact selector.

### Canonical boundary-closure ledger

This table is the sole tracker definition of the selected boundary rules. Other
sections reference BC IDs rather than restating the rules. `SELECTED` records a
planning decision, not normative adoption. Owner requirement anchors become exact
under ES-01A through ES-04; verification IDs and selectors become exact under ES-05.

| BC ID | Planned acceptance | Required target rule | Normative owners | Main slices | Decision | Normative adoption | Implementation | Required future bindings |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| BC-001 | EXT-AC-142 | Add `empty_state_policy=allowed\|forbidden`; metadata never makes state present; empty initialization is permitted only when allowed; metadata-present/empty state is valid only under `allowed`; Network Flow selects `allowed`. | Extensions, Core 02, Network Flow | ES-01A, ES-02, ES-03, ES-11 | SELECTED | DONE | DONE | Draft `EXT-REQ-188`, `EXT-REQ-234`; Core `REQ-02-261`; profile `NF-REQ-182`, `NF-AC-109..110`; generated state plan, authoritative owner counters, metadata bootstrap, allowed/forbidden matrix, empty initialization, and final validation pass the exact BC-001 selector. |
| BC-002 | EXT-AC-143 | Establish one validation precedence: invocation failure, structural invalidity, overflow, remaining schema defects, valid findings, valid empty result. Counts `257..4096` violate 256-bound schemas; `4097+` selects overflow. | Extensions, Core 04 | ES-01A, ES-02, ES-10 | SELECTED | DONE | DONE | Draft `EXT-REQ-225`; Core `REQ-04-144`; the runtime exact precedence and counts `0`, `256`, `257`, `4096`, and `4097` pass in Extensions root `20260720T080718Z-p592834`. |
| BC-003 | EXT-AC-144 | Split portability export/import results; make import side-effect-free preparation followed by shared transaction participation; add scoped staged-output capability; set participant and aggregate import input ceilings to 64 MiB. | Extensions, Core 01, Incident Portability, profile owners | ES-01A, ES-02, ES-04, ES-11 | SELECTED | DONE | DONE | Draft `EXT-REQ-221`, `EXT-REQ-232`; Incident Portability/Core `REQ-01-631`; Reporting `REQ-RPT-019a`; Composition no-change `REQ-RC-076a`; distinct result types, pre-invocation 64 MiB bounds, scoped staging denial/abandonment, and transaction handoff pass BC-003 without a combined reader. |
| BC-004 | EXT-AC-145 | Add authored `cartulary.extension_dependency_declaration_set.v1`; arrays are always present; null is invalid; manifests supply versions/digests; generation emits the snapshot. | Extensions and dependency owners | ES-01A, ES-06, ES-07 | SELECTED | DONE | DONE | Adopted anchor `EXT-REQ-174`; exact ten-row snapshot and present arrays are packaged and validated by `TestExtensionBC004DependencyDeclarations_Static`. |
| BC-005 | EXT-AC-146 | Add `recognized_profile.primary_owner_contract_ref`; define the sole source of every descriptor member; reject missing/multiple scalar sources and duplicate set members; prohibit prose or code inference. | Core 00, Extensions generator | ES-01A, ES-02, ES-06, ES-07 | SELECTED | DONE | DONE | Six packaged descriptors each have exactly one recognition source, scalar provenance, and an empty capability set; `TestExtensionBC005DescriptorProvenance_Unit` validates the projection. |
| BC-006 | EXT-AC-147 | Add condition annotations for generated schema rules and closed decision tables for procedural validators; every admitted validation surface must have a complete condition inventory; unregistered emitted conditions fail conformance. | Extensions plus every validation owner | ES-01A, ES-02 through ES-04, ES-06, ES-07, ES-10 | SELECTED | DONE | DONE | The packaged registry closes the authored condition set; coordinator integrity admission and runtime emission reject missing, stale, duplicate, and invented conditions in Extensions root `20260720T080718Z-p592834`. |
| BC-007 | EXT-AC-148 | Replace every “at least” closure-category derivation with an exact subject/contribution-kind mapping; generated subject rows permit no owner-authored not-applicable reason; only fixed baseline rows retain their enumerated reasons. | Extensions generator | ES-01A, ES-06, ES-07 | SELECTED | DONE | DONE | Six packaged catalogs derive every contribution row from the exact twelve-kind mapping and permit no generated `not_applicable`; validated by `TestExtensionBC007ClosureMapping_Unit`. |
| BC-008 | EXT-AC-149 | Close clause kinds, parent kinds, zero-based ordinals, half-open byte ranges, parent scope, clause-ID digest input, and document-clause mapping; add authored `cartulary.extension_traceability_mapping_source.v1`. | Extensions documentation tooling, Harness | ES-01A, ES-05, ES-06, ES-07, ES-09 | SELECTED | DONE | DONE | Exact digest-bound half-open mappings for EXT-AC-142 through EXT-AC-158 are packaged and validated by `TestExtensionBC008ClauseTraceability_Static`; ES-13/14 retain complete 158-criterion and all-clause accounting. |
| BC-009 | EXT-AC-150 | Add non-null inactive schema reference exactly for `syntax_only`; restrict its vocabulary to inert structural validation; while inactive apply no required/default omission policy, create no configuration view, perform no resolution, and discard accepted values. | Extensions, Core 04, profile owners | ES-01A, ES-02, ES-03, ES-10 | SELECTED | DONE | DONE | Draft `EXT-REQ-207`, `EXT-REQ-208`; Core `REQ-04-143`; inert present/omitted/null/reference/invalid matrices and negative effect/secret paths pass at `20260720T080718Z-p592834` and `20260720T075614Z-p522474`. |
| BC-010 | EXT-AC-151 | Define `unacquired -> acquiring -> held -> uncertain -> held/lost`, plus release; uncertainty closes admission immediately; proof must come from the original lease session; loss is irreversible and exits 70; initial acquisition timeout exits 2. | Core 04, platform lease adapter | ES-01A, ES-02, ES-10 | SELECTED | DONE | DONE | Draft `EXT-REQ-213`; Core `REQ-04-145`; closed lifecycle, original-session recovery, PostgreSQL contention/crash release, readiness closure, and process exits pass at the ES-10 roots recorded above. |
| BC-011 | EXT-AC-152 | Define checked/saturating local deadline calculation, inherited deadline minimum, `now >= deadline` expiry, commit/cancellation/timeout precedence, equal-deadline tiebreak, and zero-grace behavior. | Core 04, Core 01 | ES-01A, ES-02, ES-10, ES-11 | SELECTED | DONE | DONE | Draft `EXT-REQ-215`; Core `REQ-01-629`, `REQ-04-144`; checked/saturating deadlines and the bounded transaction coordinator implement cancellation/timeout sampling, proven-commit precedence, and indeterminate-commit fatality without retry or rollback claims. |
| BC-012 | EXT-AC-153 | Bound participants to `1..16384`; bound per-participant and aggregate input to 64 MiB; bound aggregate prepare results to 64 MiB; stop at the first invalid validator in participant order; sample cancellation around every step and invocation. | Core 01, Extensions | ES-01A, ES-02, ES-04, ES-11 | SELECTED | DONE | DONE | Draft `EXT-REQ-219`; Core `REQ-01-629`; exact count/byte boundaries, participant-order first-invalid behavior, cancellation checkpoints, commit proof, and indeterminate outcomes pass BC-012. |
| BC-013 | EXT-AC-154 | Define every allocated staged-object default; define the exact expiry/retry eligibility predicate and ordering; abandon upload failures immediately; prohibit holding a database transaction during deletion; serialize sweeps and coalesce missed intervals. | Core 01 object storage | ES-01A, ES-02, ES-11 | SELECTED | DONE | DONE | Draft `EXT-REQ-192`; Core `REQ-01-630`; the closed ledger constraints, atomic publication/reference boundary, immediate access denial, committed abandonment before I/O, saturated retry order, and serialized/coalesced janitor pass BC-013 and real-PostgreSQL evidence. |
| BC-014 | EXT-AC-155 | Restore v1 only into a stopped empty target; process groups numerically and bindings sequentially; validate before advancing; failed targets never serve; no inactive profile code; rebuild derived state only after successful claim. | Core 01 backup, profile owners | ES-01A, ES-02 through ES-04, ES-11 | SELECTED | DONE | DONE | Draft `EXT-REQ-235`; physical-backup/Core `REQ-01-632`; Network Flow `NF-REQ-184`, `NF-AC-112`; digest-bound codecs and exact `100/200/300/400` stopped-empty sequential restore validate each binding before advancement and permanently deny failed-target serving. |
| BC-015 | EXT-AC-156 | Add `client_build_class='standard'`; require one support row for every claimable profile with workspaces; require Network Flow major 2 and `network_analysis`; linearize generation reservation and epoch rollover. | Core 01, Core 03, Network Flow, web | ES-01A, ES-02, ES-03, ES-12 | SELECTED | DONE | DONE | Draft `EXT-REQ-211`, `EXT-REQ-212`; Core `REQ-01-542`, `REQ-01-151.1`, `REQ-03-011A`, `REQ-03-303`; Network Flow `NF-REQ-006a`, `NF-REQ-181`, `NF-REQ-187`, `NF-AC-108`, `NF-AC-114`; exact support/asset binding, no-store availability, intersection, stale/concurrency/rollover, lazy-load, fallback, and Base identity matrices pass in the ES-12 roots. |
| BC-016 | EXT-AC-157 | Prohibit capability facts and all nonempty capability arrays in v1; use `extension_capability_not_supported`; retain empty wire arrays for future compatibility; require a later capability contract before activation. | Core 00, Extensions | ES-01A, ES-02, ES-06, ES-09, ES-12 | SELECTED | DONE | DONE | Draft `EXT-REQ-077`; Core owner anchors; every descriptor, binding, discovery row, client-support row, and availability input has an empty capability array; generation and runtime reject nonempty values, and browser activation returns exact `extension_capability_not_supported`. |
| BC-017 | EXT-AC-158 | Add fatal `published_component_lost` for unexpected termination of a publication-plan listener, dequeue gate, or worker; distinguish individual operation failures; close readiness/admission, drain, preserve committed state, and exit 70; no in-process restart. | Core 04, Core 01 jobs/runtime | ES-01A, ES-02, ES-10, ES-11 | SELECTED | DONE | DONE | Draft `EXT-REQ-134`, `EXT-REQ-193`; Core `REQ-01-633`, `REQ-04-145`, `REQ-04-146`; listener and published cleanup-worker loss use the fatal no-restart path, while individual deletion/job-operation failures retain durable proof/retry state. |

ES-01A through ES-05 replaced every boundary's specification and primary-selector
future binding. Runtime evidence remains incomplete until the implementation slices
replace the current `informative` selector posture with owner-correct implementation
rows and terminal evidence. The adopted closure MUST contain no `TODO`,
“appropriate,” “as needed,” “at least,” or implementer-selected fallback for any BC
rule.

**Table 6-A. Boundary-to-primary-selector routing**

All rows use verification
`module.extensions.verification.behavior_contract`. The separate exact row
`module.extensions.static.contract_accounting_e80c9e3dc7` uses
`module.extensions.verification.contract_accounting` and verifies one-to-one routing.

| BC | Acceptance criterion | Exact primary row | Family |
| --- | --- | --- | --- |
| BC-001 | EXT-AC-142 | `module.extensions.unit.bc001_empty_state_7ca75ba0bc` | `module.extensions.unit` |
| BC-002 | EXT-AC-143 | `module.extensions.unit.bc002_validation_precedence_af944dca6e` | `module.extensions.unit` |
| BC-003 | EXT-AC-144 | `module.extensions.integration.bc003_portability_separation_e4b6d361d2` | `module.extensions.integration` |
| BC-004 | EXT-AC-145 | `module.extensions.static.bc004_dependency_declarations_7dd570e1e4` | `module.extensions.static` |
| BC-005 | EXT-AC-146 | `module.extensions.unit.bc005_descriptor_provenance_1e0ea91df8` | `module.extensions.unit` |
| BC-006 | EXT-AC-147 | `module.extensions.unit.bc006_validation_inventory_6e85895643` | `module.extensions.unit` |
| BC-007 | EXT-AC-148 | `module.extensions.unit.bc007_closure_mapping_08c4e88841` | `module.extensions.unit` |
| BC-008 | EXT-AC-149 | `module.extensions.static.bc008_clause_traceability_1991b482d2` | `module.extensions.static` |
| BC-009 | EXT-AC-150 | `module.extensions.unit.bc009_inactive_syntax_42f66761e3` | `module.extensions.unit` |
| BC-010 | EXT-AC-151 | `module.extensions.process.bc010_lease_lifecycle_4be7ab1e5d` | `module.extensions.process` |
| BC-011 | EXT-AC-152 | `module.extensions.integration.bc011_deadline_precedence_ef23af86ac` | `module.extensions.integration` |
| BC-012 | EXT-AC-153 | `module.extensions.integration.bc012_participant_limits_2c63f740c8` | `module.extensions.integration` |
| BC-013 | EXT-AC-154 | `module.extensions.integration.bc013_staged_objects_2b44f1267c` | `module.extensions.integration` |
| BC-014 | EXT-AC-155 | `module.extensions.integration.bc014_restore_ordering_bc05082e06` | `module.extensions.integration` |
| BC-015 | EXT-AC-156 | `module.extensions.integration.bc015_browser_availability_e0a71bee5d` | `module.extensions.integration` |
| BC-016 | EXT-AC-157 | `module.extensions.unit.bc016_capabilities_disabled_77bb995602` | `module.extensions.unit` |
| BC-017 | EXT-AC-158 | `module.extensions.process.bc017_component_loss_755919c8d7` | `module.extensions.process` |

### Companion-owner amendment plan

| Owner | Required amendment or parity decision | Boundary closures | Ownership guardrail | Gate dependencies | Completion evidence | Status |
| --- | --- | --- | --- | --- | --- | --- |
| Core 00 | Adopt the shared subsystem, manifest association, current majors, `network_flow_activity@2`, and `network_flow_activity -> import@1`. | BC-005, BC-016 | Recognition, claimability, retirement, and current major remain Core 00 only. | EXT-GATE-001, 013, 017 | `REQ-00-065`; document `106451bf...`; manifest `e032d688...`; exact digests in the ES-14 completion snapshot | ADOPTED |
| Core 01 | Adopt strict seven-member producer/tolerant decoder, Base reservation registry, availability/no-store member, dispatch precedence, bounded transactions, staged objects, jobs, backup/participant interfaces, final commit, and recovery. | BC-003, BC-011 through BC-015, BC-017 | Public routes, envelopes, errors, transaction shell, object storage, and physical backup orchestration remain Core 01. | EXT-GATE-002, 003, 017, 022, 025, 026 | `REQ-01-151.1`, `542..548`, `629..633`; document `ae34ee62...`; manifest `99680512...`; exact digests in the ES-14 completion snapshot | ADOPTED |
| Core 02 | Adopt or confirm extension-resource, authoritative/derived family, state-presence exclusion, and no cross-owner authoritative-write boundaries. | BC-001 | No implicit Core record/view/saved-view promotion. | EXT-GATE-004 | `REQ-02-261`; document `85d0f16c...`; manifest `5b82f774...`; exact digests in the ES-14 completion snapshot | ADOPTED |
| Core 03 | Adopt epoch/generation, stable client/WebSocket identity, eligibility intersection, lazy loading, fallback, unsupported-major and authorization-loss consequences. | BC-015 | Browser/workbook behavior remains Core 03/web owned. | EXT-GATE-005, 017, 021 | `REQ-03-011A`, `REQ-03-303`; document `7ec2ddf5...`; manifest `94f8b034...`; exact digests in the ES-14 completion snapshot | ADOPTED |
| Core 04 | Adopt inactive syntax-only processing, all deadlines, process lease, Stage 6, readiness degradation, diagnostics, fatal lifecycle, and exit codes. | BC-002, BC-009 through BC-011, BC-017 | Deployment config, authorization, readiness, and process exit remain Core 04. | EXT-GATE-006, 015, 023 | `REQ-04-143..146`; document `da7ff2d9...`; manifest `4cde78a5...`; exact digests in the ES-14 completion snapshot | ADOPTED |
| Network Flow Activity | Publish owner manifest/fragments; remove competing discovery; adopt major 2; declare dependency, state, initialization, migrations, bindings/codecs, jobs, participants, rebuilds, and blocker. | BC-001, BC-009, BC-014, BC-015 | Network Flow resources/routes/algorithms remain `module.networkflow`. | EXT-GATE-007, 016, 017, 024 | `NF-REQ-006a`, `NF-REQ-172a`, `NF-REQ-181..187`; document `14f8fdb9...`; manifest `50a291fa...`; fragment `224bcf2c...`; final owner root `20260720T131350Z-p2363611` | ADOPTED |
| Incident Portability | Import closed export/import results, scoped staged-output capability, participant context/result/finding, logical refs, declarative inactive blocker, and invocation matrix when the interface changes. | BC-003, BC-012 | Public incident-bundle workflow remains Core/Incident Bundle owned. | EXT-GATE-018 | Amended Core `REQ-01-631`; resolution set `97f39cf2...`; owner slice `20260720T044918Z-p135306` | ADOPTED |
| Physical backup owner | Import binding, codec, ordering, restore-target, and invocation contracts when changed. | BC-014 | Physical backup orchestration remains platform/Core owned. | EXT-GATE-003, 016, 018 | Amended Core `REQ-01-632`; resolution set `97f39cf2...`; runtime/codec evidence remains ES-11 | ADOPTED |
| Reporting | Import generic descriptor/claim/compatibility/state-presence/participant lifecycle only if its interface changes. | BC-003, BC-012 | Snapshot, render, release, and report model remain Reporting owned. | EXT-GATE-008, 018 | `REQ-RPT-019a..019b`; document `72f8948f...`; manifest `1a74ce98...`; fragment `364e0dbb...`; final owner root `20260720T131808Z-p2394701` | ADOPTED |
| Report Composition | Import only the generic interfaces it actually consumes. | BC-003, BC-012 | Authoring, validation, preview, and composition schema remain Composition owned. | EXT-GATE-008, 018 | No generic participation at `REQ-RC-076a`; validation surface amended there; document `489a839d...`; manifest `6c3bc194...`; resolution set `97f39cf2...` | ADOPTED |
| OpenTelemetry | Derive `cartulary.profile.claims` from canonical resolved claim set and digest; prohibit profile secrets/content. | none | Telemetry signal shape remains OpenTelemetry owned. | EXT-GATE-010 | OTel document `b31baa1e...`; manifest `4fc1bb2d...`; hard-coded vocabulary removed; final OTel root `20260720T140642Z-p2771212` | ADOPTED |
| Domain vocabulary | Add adopted extension terms and remove stale discovery/unclaim/migration/multiprocess/client-support language after owners adopt. | BC-001 through BC-017 (terminology only) | Domain document remains terminology reference, never behavior owner. | EXT-GATE-011 | Document `62c8cb0f...`; adopted recognition/claim, state/metadata, discovery/support/availability, capability, and stable-identifier distinctions | ADOPTED |
| Testing Harness v2 | Add authored owner contracts/rows/profiles/topology; amend NLSpec only if a new public command, runner, schema family, or execution-profile contract is needed. | BC-006, BC-008 | Harness owns mechanics only and cannot derive product behavior from the Extensions document. | EXT-GATE-009, 020, 027, 028 | Harness document `61d120b4...`; manifest `c86a42de...`; verification contract `4af4ea2c...`; 24 exact rows across five nonempty families; final Harness root `20260720T131130Z-p2336673` | ADOPTED |

### Section 27 artifact and generation plan

Authored owner manifests, fragments, schemas, contracts, catalog rows, and topology
inputs are the only hand-authored sources. All canonical registries, descriptors,
digests, indexes, embedded constants, TypeScript projections, task surfaces, schedules,
and topology render outputs are generated through repository-owned generators.

| Artifact or schema family | Authored owner/input | Generated or runtime consumers | Dependencies and validation | Status |
| --- | --- | --- | --- | --- |
| `cartulary.extension_dependency_declaration_set.v1` | Extensions specification owner; dependency owners supply manifest versions and digests | Exact input to dependency snapshot generation | All ten exact rows, present empty arrays, owner bytes, anchors, manifests, and artifact digest validate; BC-004 | DONE |
| `cartulary.extension_dependency_snapshot.v1` | Adopted dependency identities/versions/digests | Extensions generator and adoption accounting | Exact ten-row projection is packaged and drift-tested; BC-004 | DONE |
| `cartulary.extension_owner_contract_manifest.v1` | One manifest per dependency owner document | Locator validator, snapshot, integrity, runtime admission | All ten dependency manifests pass exact document, length, anchor-range, fragment-association, and digest validation | DONE |
| `cartulary.extension_owner_fragment.v1` | Contributing named owner documents | Owner-input registry and descriptor derivation | Core 00/01/04, Network Flow, and Reporting fragments supply the complete recognized-profile fact set and are manifest-bound | DONE |
| `cartulary.extension_owner_input_registry.v1` | Generator input from validated manifests/fragments | Descriptor and registry generation | Exact dependency, manifest, and fragment projections are packaged and ordered | DONE |
| `cartulary.extension_owner_fact_identity.v1` | Extensions schema and derivation vectors | Generator ordering/collision/accounting | Exact fact-source inventory and derived identities are generator-tested | DONE |
| Profile configuration contract/view | Extensions shape; named profile content; Core 04 view | Claim resolution and diagnostics | All six configuration contracts and their claim-fact canonical digests validate; claimed views and unclaimed inert processing use the one resolved claim set | IMPLEMENTED |
| Inactive-value schema family | Profile configuration owners under the closed Extensions vocabulary | Inert validation of unclaimed `syntax_only` values | Closed structural vocabulary, forbidden active annotations, effect-free contract, and non-null syntax-only validation exist; BC-009 | DONE |
| Descriptor-source schema | Extensions schema; ephemeral construction only | Generator in-memory normalization | Sole scalar sources and multiplicity validate without retaining a source artifact; BC-005 | DONE |
| Descriptor schema and per-profile descriptors | Generated from owner input, including `recognized_profile.primary_owner_contract_ref` | Discovery, bindings, client support, accounting | Six exact descriptor projections, generated schema, provenance/parity tests; BC-005 | DONE |
| Profile registry schema and canonical registry | Extensions generator | Runtime admission, discovery input, build package | Six-profile canonical registry and generated schema are packaged and drift-tested | DONE |
| Registry integrity object | Extensions generator/build packaging | Runtime admission and static accounting | Coordinator admits exact registry/descriptor/binding/Base-reservation/client-support package digests before exposing queries | DONE |
| Base route reservation registry | Core 01 authored route ownership, generated registry | HTTP dispatch overlap validator | Canonical Core 01 reservation input is dependency-bound; runtime admission validates it before the one descriptor/claim epoch is injected into dispatch | DONE |
| Client support registry and client asset-set manifest | Web build inputs and generated asset manifest | Browser eligibility intersection | Source support facts are owner-authored; the web build hashes every packaged asset, emits a canonical manifest, binds the final standard registry to its digest, and server startup validates the exact archive/manifest/registry set | DONE |
| Workspace availability | Core 01 plus Extensions result | Workbook startup and web controller | Exact no-store claim/workspace/authorization rows feed a locally generated epoch and linearized generation intersection; stale or unauthorized work cannot render | IMPLEMENTED |
| Publication plan and six component schemas | Extensions/Core 04 | Application startup Stage 6 coordinator | The exact plan and resolved claim set are frozen before runtime side effects; the listener-bound publication callback opens admission only after bind under a held lease; readiness/loss/drain/fatal lifecycle is closed; BC-010, BC-011, BC-017 | IMPLEMENTED |
| Implementation binding schema and packaged bindings | Build plus profile implementations | Runtime admission | Six explicit packaged bindings pass runtime identity/major/descriptor/capability/order/integrity admission before claim resolution | DONE |
| Admission validation/context/result schemas | Extensions | Profile preflight/post-migration calls | Exact runtime precedence and condition-registry enforcement apply to initial, migration-step, and final owner validation | IMPLEMENTED |
| Migration context/apply/validation/final-state schemas | Extensions shape; profile definitions | Migration coordinator | The owner-locked coordinator admits immutable step definitions, records the forward ledger, validates each step, and never down-migrates committed state | IMPLEMENTED |
| State-presence manifest/digest/vectors | Each versioned profile plus generator | Migration, backup, portability, accounting | Generated immutable catalog plus named-owner counters implement metadata-independent presence and the complete `allowed`/`forbidden` matrix; BC-001 | IMPLEMENTED |
| State-initialization definition/context/result | Extensions shape; each versioned profile | Fresh-state coordinator | Network Flow's digest-bound empty initialization runs only under `allowed`, then final owner validation and metadata commit complete the state epoch; BC-001 | IMPLEMENTED |
| Physical state binding | Build plus durable profile owner | Backup/restore and integrity | Generated Network Flow binding names exactly four authoritative families in numeric restore-group order and admits the packaged catalog by digest | DONE |
| Backup binding codec and vectors | Build, backup owner, durable profile | Physical backup/restore | Digest-bound codecs, 64 MiB member limits, stopped-empty admission, sequential restore/validation, and failed-serving denial pass BC-014 | IMPLEMENTED |
| State metadata and migration-ledger logical schemas | Extensions | PostgreSQL/profile state coordinator | Forward migration `00034` creates generic metadata and immutable ledger tables, backfills Network Flow lineage/version, and is exercised through the PostgreSQL store | IMPLEMENTED |
| Job-kind contracts | Extensions shape; profile content | Core jobs and profile workers | No claimed Network Flow worker kind exists; future kinds must be generated and admitted before publication, while the cleanup worker uses the generic published-component lifecycle | DONE |
| Job commit proof and cancellation observation | Core 01 jobs plus Extensions | Reconciliation/replay | Forward schemas and transaction-scoped platform APIs record commit proof and cancellation observation without absorbing named-owner job behavior | IMPLEMENTED |
| Transaction participant contract/context/result/finding | Core 01 plus Extensions | Cross-owner coordinator and participants | Ordered coordinator enforces `1..16384`, all 64 MiB limits, cancellation checkpoints, first-invalid validation, and proven/absent/indeterminate commit precedence | IMPLEMENTED |
| Participant specialization and portability/reporting/backup contexts/results | Applicable shared and profile owners | Incident Bundle, Reporting, Composition, backup | Generated specialized and transaction participants drive side-effect-free preparation, scoped staging, shared transaction mutation, and ordered backup restore; BC-003, BC-012, BC-014 | IMPLEMENTED |
| Operation-specific portability export/import results | Extensions plus Incident Portability | Side-effect-free import preparation, shared transaction participation, and export publication | Distinct closed export/import identities, side-effect-free preparation, scoped staging, shared-transaction mutation, and 64 MiB ceilings are authored | DONE |
| State blocking predicate | Extensions plus applicable profile | Inactive portability checks | The generated Network Flow predicate names only its authoritative tables family; state presence remains a named-owner fact and generic metadata never blocks portability | IMPLEMENTED |
| Staged-object logical schema | Core 01 object storage plus Extensions | Transaction publication and cleanup | Forward schema and platform store implement closed defaults, atomic publication, immediate abandonment, transaction-free deletion, retry order, serialized sweeps, and coalescing | IMPLEMENTED |
| `cartulary.extension_validation_surface_declaration.v1` | Each schema or procedural-algorithm owner | Complete input to validation-condition registry generation | Extensions, Core 01/04, Network Flow, Reporting, and Composition declarations have present arrays, unique conditions, closed decision rows, and required shared conditions | DONE |
| Validation-condition registry | Generator plus every validation owner | Startup/runtime diagnostics and accounting | Exact authored-set parity, ordering, redacted-formatter safety, coordinator integrity admission, and unregistered-emission rejection are implemented; BC-002, BC-006 | IMPLEMENTED |
| Startup finding schema | Core 04 plus Extensions | Startup diagnostics/readiness | Closed safe findings cover configuration, process, state, migration, transaction, staging, and restore admission without secret-bearing adapter text | IMPLEMENTED |
| Fatal-condition and process-lifecycle registries | Core 04 plus Core 01 jobs/runtime | Readiness, admission, drain, process supervision, and exit handling | Lease/listener and published cleanup-worker loss use readiness/admission closure, drain, exit `70`, and no in-process restart; individual operations remain retryable | IMPLEMENTED |
| Per-profile contract closure catalog | Extensions generator | Conformance manifest/accounting | Six exact catalogs include fixed baseline and exhaustive contribution rows; no owner reduction | DONE |
| Conformance manifest and index | Named profile owners plus generator | Static adoption accounting | Six manifests and one index are packaged; full 158-criterion adoption accounting remains ES-13/14 | DONE |
| Registry accounting object | Extensions generator | Adoption gate only | Named deterministic predicates and current source digests are packaged without run-result inputs; final evidence accounting remains ES-13/14 | DONE |
| `cartulary.extension_traceability_mapping_source.v1` | Extensions specification owner | Exact input to acceptance/verification and document/clause mappings | Exact draft digest and nonoverlapping half-open ranges bind EXT-AC-142 through EXT-AC-158 to owner requirements and the active verification ID | DONE |
| Clause traceability object | Specification owner/document tooling | Static adoption accounting | Exact EXT-AC-142 through EXT-AC-158 source ranges, ordinals, IDs, and digests are packaged; complete all-clause audit remains ES-13/14; BC-008 | DONE |
| Canonicalization and normative-source-lint vectors | Extensions/document tooling | Generator/linter tests | Canonical final-LF and digest vectors plus the full source/promotion lint pass on the adopted snapshot | DONE |
| `extension_safe_logical_ref_v1` and vectors | Extensions | Diagnostics, indexes, findings | Valid logical identities and path, endpoint, whitespace, and traversal rejection vectors are authored | DONE |
| Verification registry/owner contracts | Harness verification owners | Catalog validator and evidence accounting | Both immutable `module.extensions` IDs are active; later runtime slices replace informative posture and add owner-correct evidence | DONE |
| Test-owner registry/family manifest | Harness test owners | Owner-slice planner/scheduler | 24 active exact-selector rows; unit/static/integration/process/browser-stateful families are nonempty; Report Composition has one active owner row | DONE |
| Runner registry and runtime/resource/fixture profiles | Testing Harness v2 | Runner adapters/scheduler | Existing Go runner and `none`/`default`, `go_balanced`, and `none` fixture profiles validate; no new public Harness mechanic or profile was required | DONE |
| Slice plan and scheduler summary | Harness-generated per invocation | Owner execution/finalization | Exact owner plans and paired current-snapshot shards were resolved and audited for all 14 affected owners in ES-13 | DONE |
| Evidence accounting and owner summary | Harness target finalizers | Evidence auditor/adoption gate | Paired shard for every owner/target partition | DONE |
| Evidence-root manifest and audit summary | Evidence-audit caller/auditor | Final adoption gate | Explicit compatible roots only; no newest/historical fallback | DONE |
| Generated Go/TS/schema/topology/task projections | Generators from the authored inputs above | Backend, web, Make, scheduler | `make generate-drift`, artifact policy, JSON shape, harness contract | DONE |

### Harness v2 onboarding plan

The Extensions owner must be added atomically to
`contracts/verification/owners/module.extensions.json`,
`contracts/verification/registry.json`, `tools/test_catalog_owner.json`, and
`tools/test_families/module.extensions.json`, with authored execution-topology/profile
inputs and generator-owned projections. The two immutable product verification IDs are:

- `module.extensions.verification.behavior_contract` for shared Base Profile
  extension runtime postconditions;
- `module.extensions.verification.contract_accounting` for owner input, generated
  contracts, registry integrity, closure, diagnostics, traceability, and static routing.

Imported Core, web, platform, application, and named-profile behavior remains routed
through those primary owners' verification IDs; `module.extensions` must not duplicate
it. The initial exact selectors below are selected planned identities. If an
adopted owner later reallocates a postcondition, change the owning row and traceability
together; do not duplicate the selector.

| Planned owner/family | Verification ID | Exact selector | Profiles | Evidence class | Required role |
| --- | --- | --- | --- | --- | --- |
| `module.extensions.unit` | behavior contract | Go package `./internal/modules/extensions`; tests `TestExtensionsClaimResolutionContract`, `TestExtensionsDependencyAndCollisionContract`, `TestExtensionsBindingAdmissionContract` | runtime/resource/fixture `none` | unit | Default owner row |
| `module.extensions.integration` | behavior contract | Go package `./internal/modules/extensions`; tests `TestExtensionsDiscoveryAndDispatchContract_Integration`, `TestExtensionsStateInitializationAndMigrationContract_Integration`, `TestExtensionsTransactionAndJobContract_Integration`, `TestExtensionsBackupPortabilityReportingContract_Integration` | `default` / `go_transaction_heavy` / `postgres_transaction` | integration | Service-backed full-owner row |
| `module.extensions.process` | behavior contract | Go package `./internal/app/serverprocess`; tests `TestExtensionsApplicationLeaseAndFatalLifecycle_Process`, `TestExtensionsAtomicPublication_Process`, `TestExtensionsStagedObjectExpiryCleanup_Process` | authored unclaimed and claimed profiles / `go_process` / service stack | process | Process/evidence-class gate |
| `module.extensions.static` | contract accounting | Go package `./internal/modules/extensions`; tests `TestExtensionsGeneratedArtifactIntegrity`, `TestExtensionsValidationConditionRegistry`, `TestExtensionsContractClosureAndAccounting`, `TestExtensionsClauseTraceabilityAndLimits` | `none` / `none` / `none` | static | Default owner row |
| `web.application.extensions` | existing/new primary web verification | Vitest file `apps/web/src/app/extensions/extensionLifecycle.test.tsx`; title `Verify extension discovery, client support, authorization, availability generation, fallback, and Base state preservation.` | `none` / `none` / `none` | unit | Affected web full-owner row |
| `web.application.extensions_browser` | existing/new primary web verification | Playwright file `apps/web/e2e/extensions.spec.ts`, project `chromium`, stage `webserver_backed`; title `Extensions eligibility uses discovery, client support, authorization, and current availability generation.` | claimed authored runtime / `browser_exclusive` / `service_stack` | browser | Browser evidence gate |
| `web.application.extensions_stateful` | existing/new primary web verification | Playwright file `apps/web/e2e/extensions.stateful.spec.ts`, project `chromium`, stage `stateful`; title `Extension authorization loss and generation rollover dispose extension state while preserving Base state and client identity.` | claimed authored runtime / `browser_exclusive` / `service_stack` | browser | Stateful evidence gate |
| `module.networkflow` and `web.networkflow` | existing Network Flow IDs plus any owner-amended ID | Existing exact selectors updated for generic seven-member discovery and major 2; add exact title `Network Flow major 2 uses only generic extension discovery and reserved dispatch.` if no existing selector proves it | `network_flow_claimed` and `default` as applicable | unit/integration/browser | Affected full-owner closure |
| Harness generated-artifact owner | Harness-owned verification | Existing shell command IDs for generation drift, artifact policy, JSON shape, and catalog/topology checks | `none` | static | Evidence-class gate, not product behavior |

Exact row IDs are generated from these owner/family/selector identities by the existing
catalog tooling; the generated row ID is not authored as product identity. A family
name, filename, target exit, title prefix, aggregate `make check`, or support-only row
cannot substitute for the exact selectors. Full adoption requires compatible
successful full-owner slices, every required evidence-class target, paired
`cartulary.test_evidence_accounting.v1` and `cartulary.test_owner_summary.v1` shards,
one terminal record per row, and one exact-root evidence audit.

### Compact requirement traceability ledger

Every requirement ID appears exactly once in the section-grouped rows below. The
grouping references IDs rather than reproducing normative prose. Primary ownership is
the draft allocation to be confirmed by companion adoption; current adopted ownership
continues until then.

| Requirement IDs | Workstream | Primary owner | Affected artifacts/code | Dependencies | Validation posture | Completion state |
| --- | --- | --- | --- | --- | --- | --- | --- |
| EXT-REQ-001, 002, 003, 004, 005, 174 | WF-00, WF-02, WF-02A | Extensions spec plus imported owners | Dependency declarations/snapshot/manifests and source posture; BC-004 | Core 00-04 and named owners | Source/anchor/dependency validation | DONE |
| EXT-REQ-006 through 014 | WF-00, WF-02A, WF-09 | Extensions spec/document tooling | Normative source and identifier traceability; BC-008 | Exact draft bytes and Harness IDs | Linter, uniqueness, bidirectional mapping | DONE |
| EXT-REQ-015 through 019 | WF-00, WF-04 | Extensions spec/Core boundaries | Runtime non-goals and subsystem facade | Owner allocation | Negative architecture/security tests | DONE |
| EXT-REQ-020, 021, 022, 023, 175, 203 | WF-02A, WF-03 | Extensions generator | Scalars, locators, manifests, safe refs; BC-004 | Owner document digests | Grammar, traversal, symlink, digest vectors | DONE |
| EXT-REQ-024, 025, 026, 027, 176, 177, 204, 205, 206 | WF-02, WF-02A, WF-03 | Named owners plus Extensions generator | Owner fragments/input/fact identities; BC-005, BC-006 | Adopted manifests | Determinism, omission, identity/collision tests | DONE |
| EXT-REQ-028 through 033 | WF-02, WF-02A | Core 00 | Recognized profile facts and adoption state; BC-005, BC-016 | Companion owner revisions | Core/registry/discovery parity | DONE |
| EXT-REQ-034 through 040, 178, 209 | WF-02A, WF-03 | Extensions generator/profile owners | Descriptor-source/configuration/descriptor schemas; BC-005, BC-009 | Owner input registry | Closed-shape/default/ephemeral/digest tests | DONE |
| EXT-REQ-041 through 046, 179, 180 | WF-02A, WF-03 | Extensions generator/build | Registry, dependency snapshot, canonical JSON, integrity and package roots; BC-004, BC-005 | Descriptors/bindings/static support | Canonical, bound, integrity, zero-profile vectors | DONE |
| EXT-REQ-047 through 053, 184, 207, 208, 213, 214, 215 | WF-02, WF-02A, WF-04 | Core 04 plus Extensions | Claim config/view, lease, publication, deadlines; BC-009 through BC-011, BC-017 | Registry, bindings, platform config | Config/process/readiness/fatal tests | DONE |
| EXT-REQ-054 through 058, 187 | WF-04 | Extensions coordinator | Claim/dependency/admission algorithm | Config, descriptors, bindings | Order, preflight, no-side-effect, no-listener tests | DONE |
| EXT-REQ-059 through 064, 185 | WF-04 | Extensions coordinator/profile owners | Runtime dependency graph/probes | Resolved claims and owner contracts | Cycle/order/probe/timeout tests | DONE |
| EXT-REQ-065 through 067, 181, 182 | WF-02A, WF-03, WF-04 | Build plus Extensions | Implementation binding and parity; BC-012, BC-016 | Registry/integrity/profile implementation | Missing/extra/mismatch admission tests | DONE |
| EXT-REQ-068 through 072, 210 | WF-03 | Extensions generator/Core 01 | Collision and Base reservation registries | Canonical facts/routes/dependencies | Every collision/multiplicity/route-overlap case | DONE |
| EXT-REQ-073 through 079, 196 | WF-02A, WF-03, WF-04, WF-07 | Extensions/Core 03/profile owners | Compatibility matrices, capability prohibition, and support registry; BC-015, BC-016 | Descriptor/binding/state/schema versions | Matrix and unsupported-value tests | DONE |
| EXT-REQ-080 through 086, 194, 195, 231 | WF-02, WF-02A, WF-05 | Core 01 | OpenAPI discovery producer/decoder; BC-005, BC-016 | Core 00 facts and registry | Strict seven-member producer/tolerant decoder/parity | DONE |
| EXT-REQ-087 through 094 | WF-02, WF-02A, WF-03, WF-08 | Named owners plus generator | Contribution registry/participant bindings; BC-003, BC-007, BC-012 | Adopted fragments/descriptors | Closed-kind/duplicate/parity tests | DONE |
| EXT-REQ-095 through 099 | WF-05 | Core 01/platform HTTP | Reservation and dispatch | Base/extension registries | Exact precedence and claimed/unclaimed outcomes | DONE |
| EXT-REQ-100 through 109, 197, 201, 211, 212, 220 | WF-02A, WF-07 | Core 03/web application | Client support, availability, workspace lifecycle; BC-015, BC-016 | Discovery/auth/browser assets | Unit/browser/stateful selectors and Base preservation | DONE |
| EXT-REQ-110 through 114, 200, 226 | WF-02A, WF-06 | Core 02/profile state owners/Extensions | Resource/state ownership, closure catalog; BC-001, BC-007 | Profile descriptors/state families | No-promotion/no-cross-write/closure tests | DONE |
| EXT-REQ-115 through 119, 192, 219 | WF-02A, WF-06 | Core 01 transaction owner | Participant protocol and staged publication; BC-011 through BC-013 | Jobs/storage/owner participants | Ordered failure/cancel/deadline/commit tests | DONE |
| EXT-REQ-120 through 129, 188, 189, 190, 216, 217, 234 | WF-02A, WF-06 | Extensions plus profile state owners | State metadata, initialization, migration, ledger; BC-001 | Bindings/state presence/locks | Fresh/current/migrated/restored and resumability tests | DONE |
| EXT-REQ-130 through 135, 191, 193, 218 | WF-02A, WF-06 | Core 01 jobs plus profile owners | Job contracts/proof/reconciliation/failure isolation; BC-017 | Resolved claim and transaction results | Proof/cancel/replay/fatal tests | DONE |
| EXT-REQ-136 through 145, 198, 199, 221, 222, 223, 232, 235 | WF-02A, WF-06, WF-08 | Backup, Incident Portability, Reporting, profile owners | Presence/bindings/codecs/participant specializations; BC-003, BC-012, BC-014 | State metadata and owner amendments/parity | Backup/restore/portability/reporting matrices | DONE |
| EXT-REQ-146 through 153 | WF-02, WF-02A, WF-04 | Core 04 plus profile security owners | Config, secret refs, authorization, egress; BC-009 | Owner configuration contracts | Syntax-only/egress/secret-negative/security tests | DONE |
| EXT-REQ-154 through 158 | WF-02, WF-04 | OpenTelemetry/audit owners | Claim-set telemetry and audit fields | Canonical resolved claims | OTel conformance/privacy and audit tests | DONE |
| EXT-REQ-159 through 163, 186, 224, 225, 233 | WF-02A, WF-03, WF-04 | Extensions generator plus Core 04 | Validation registry, startup findings, messages, paths, formatters, exits; BC-002, BC-006, BC-011, BC-017 | All validation owners | Exact precedence/path/order/overflow/fatal tests | DONE |
| EXT-REQ-164 through 167, 183, 202, 227, 228, 229, 230, 236 | WF-02A, WF-03, WF-09 | Extensions spec/generator plus Harness v2 | Section 27 artifacts, accounting, traceability, selectors/evidence; BC-004, BC-006 through BC-008 | All owners and generated inputs | Static accounting, limits, full-owner shards, evidence audit | DONE |
| EXT-REQ-168 through 171 | WF-10 | Coordinated document owners | Adoption statuses and companion revisions | All gates/evidence | Atomic promotion audit | DONE |
| EXT-REQ-172, 173 | WF-00 | Future owner | No current executable package surface | Future NLSpec | Negative upload/installation/execution tests | DEFERRED |

### Acceptance-criterion traceability ledger

The first eleven rows inventory the 141 acceptance criteria currently present in the
draft. The final row reserves 17 planned criteria. The BC ledger supplies their exact
one-to-one allocation; they become normative criteria only when ES-01A adds them to
the draft. After ES-01A, the fixed acceptance count is 158.

| Acceptance IDs | Workstream | Primary owner set | Affected artifacts | Dependencies | Validation posture | Completion state |
| --- | --- | --- | --- | --- | --- | --- |
| EXT-AC-001 through 014 | WF-00, WF-03, WF-04 | Extensions/Core 00/Core 04 | Owner inputs, descriptors, registry, claims, bindings | Adopted owners/config | Static plus startup unit/process | DONE |
| EXT-AC-015 through 023 | WF-03, WF-04, WF-05 | Extensions/Core 00/Core 01 | Compatibility, dependency graph, collisions/reservations | Registry and bindings | Exact matrix/collision/route tests | DONE |
| EXT-AC-024 through 033 | WF-05 | Core 01 and Network Flow | Discovery and dispatch | Seven-member contract and major action | HTTP integration/browser support | DONE |
| EXT-AC-034 through 040 | WF-07 | Core 03/web application | Workspace, Base fallback, identity, lazy load | Support/availability registries | Unit/browser/stateful | DONE |
| EXT-AC-041 through 050 | WF-06 | Extensions/profile state owners/Core 00 | Unclaim/reclaim/retirement/migration | State metadata and ledger | Service-backed migration/process | DONE |
| EXT-AC-051 through 060 | WF-06 | Core 01 jobs/transactions plus profile/Core 02 | Jobs, commit proof, participants, ownership | Typed contracts | Failure-injection integration | DONE |
| EXT-AC-061 through 071 | WF-06, WF-08 | Backup/Portability/Reporting/Core 04/OTel | Codec, state, participants, security, telemetry | Owner amendments/parity | Restore/matrix/security/OTel | DONE |
| EXT-AC-072 through 084 | WF-02, WF-03, WF-09 | All owners/generator/Harness | Drift, manifests, limits, diagnostics, accounting | Current digests and catalog | Static drift/shape/limit/diagnostic | DONE |
| EXT-AC-085 through 096 | WF-04, WF-06, WF-07, WF-08 | Core 04/Extensions/shared owners/web | Preflight, state, jobs, objects, discovery, browser outcomes | Publication and participant contracts | Process/service-backed/browser | DONE |
| EXT-AC-097 through 110 | WF-03, WF-04, WF-05, WF-07, WF-09 | Spec/generator/Core 01/03/04/Network Flow | Traceability, owner facts, config, lease, client support | Manifests and exact selectors | Static/process/browser/full-owner | DONE |
| EXT-AC-111 through 121 | WF-03, WF-06, WF-08 | Core 01/Extensions/profile/shared owners | Transactions, staged objects, migrations, jobs, codecs, closure | State/participant contracts | Failure injection and canonical vectors | DONE |
| EXT-AC-122 through 141 | WF-03, WF-04, WF-07, WF-09, WF-10 | Generator/Harness/Core 04/web/all affected owners | Accounting, selectors, lint, artifacts, fatal/browser/evidence | All current digests and retained roots | Static/full-owner/evidence audit/docs | DONE |
| EXT-AC-142 through 158 | WF-02A, WF-03 through WF-09 | Owners identified by BC-001 through BC-017 | Boundary schemas, algorithms, tables, registries, and runtime matrices | ES-01A owner anchors; ES-05 verification IDs and exact selector families | Every criterion has one informative exact primary row; implementation slices must replace posture with behavior evidence before adoption | DONE |

### Adoption-gate traceability ledger

| Gate IDs | Workstream | Primary owner set | Affected artifacts | BC prerequisites | Dependencies | Validation posture | Completion state |
| --- | --- | --- | --- | --- | --- | --- | --- |
| EXT-GATE-001 through 006 | WF-02A, WF-04, WF-05, WF-06, WF-07 | Core 00 through Core 04 | Core revisions, registry facts, discovery, transactions, client lifecycle, config/process | BC-001 through BC-003, BC-005, BC-009 through BC-017 | Characterization and boundary-closed draft retained as draft | Core owner evidence and parity | DONE |
| EXT-GATE-007 through 011 | WF-02A, WF-07, WF-08 | Network Flow, Reporting, Composition, Harness, OTel, domain | Owner manifests/fragments, major action, participants, telemetry, vocabulary | BC-001, BC-003, BC-008, BC-014, BC-015 | Core companion amendments | Affected full-owner/parity evidence | DONE |
| EXT-GATE-012 through 016 | WF-02A, WF-03, WF-04, WF-06, WF-09 | Generator/build/Core 04/state owners | Complete artifacts, manifests, integrity, diagnostics, state/migration/codec | BC-001, BC-004 through BC-007, BC-009, BC-013, BC-014 | All authored owners | Generation/drift/limit/process/service-backed | DONE |
| EXT-GATE-017 through 022 | WF-02A, WF-03, WF-05, WF-07, WF-08, WF-09 | Core 01, Network Flow, shared owners, web, generator | Generic discovery, participants, closure, traceability, client support, Base reservations | BC-003, BC-005, BC-007, BC-008, BC-015, BC-016 | Companion adoption | Static parity plus exact owner rows | DONE |
| EXT-GATE-023 through 026 | WF-02A, WF-04, WF-06, WF-09 | Core 04/Core 01/Extensions | Lease/fatal, initialization/migration, transaction, staged-object evidence | BC-001, BC-010 through BC-014, BC-017 | Implemented failure injection | Exact v2 selectors and terminal row records | DONE |
| EXT-GATE-027, EXT-GATE-028 | WF-02A, WF-09, WF-10 | Harness v2/document tooling/all owners | Full-owner shards, evidence audit, source lint, acceptance continuity, limit selectors | BC-001 through BC-017 | Every prior gate and all 158 acceptance criteria | Explicit-root audit and documentation/static gates | DONE |

## 7. Completed Refactor Slice Plan

Every slice below completed in dependency order. The Extensions NLSpec and dependent
owners remained non-current until the ES-14 atomic promotion. The rollback notes are
retained as historical execution controls; committed profile migration steps are
never rolled back or reinterpreted by a down migration.

### Execution status

| Slice | Status | Started or completed at | Evidence checkpoint |
| --- | --- | --- | --- |
| ES-00 | DONE | 2026-07-19 | Baseline refreshed; exact current-owner discovery, reservation, claim/config, workbook, process, Network Flow, browser, HTTP-runtime, and OTel evidence passed. |
| ES-01 | DONE | 2026-07-19 | Existing draft foundation at digest `18fbd7f...` was verified: 236 unique requirements, continuous 141 acceptance rows, 28 gates, closed owner/dependency manifest model, and draft status retained. |
| ES-01A | DONE | 2026-07-19 | Draft `0.6.0` at SHA-256 `796e119f...` closes BC-001 through BC-017 through existing owner requirements and exactly EXT-AC-142 through EXT-AC-158; status remains draft. |
| ES-02 | DONE | 2026-07-19 | Core 00-04 amended as one staged companion set with exact anchors and digest-bound `cartulary.extension_owner_contract_manifest.v1` files; runtime activation remained gated until ES-14. |
| ES-03 | DONE | 2026-07-19 | Network Flow is staged at document `2.0.0`/contract major `2`; generic discovery is sole, the owner declarations close state/configuration/backup/participant/browser boundaries, and canonical owner/profile inputs are manifest-bound. Exact owner slice passed 118 tests at `20260720T043710Z-p90854`. |
| ES-04 | DONE | 2026-07-19 | Incident Portability and physical backup resolve to amended Core `REQ-01-631..632`; Reporting imports one typed participant at `REQ-RPT-019a`; Composition explicitly consumes no generic participant at `REQ-RC-076a`; validation-condition obligations are owner-anchored. Resolution set `97f39cf2...`; active owner slices passed. |
| ES-05 | DONE | 2026-07-19 | Both Extensions verification IDs and 18 exact rows validate across four nonempty families; every BC maps once in Table 6-A; Report Composition now has an active owner row; telemetry consumes a verified immutable claim-set digest without a local profile vocabulary. Harness, Extensions local/service-backed, Report Composition, OTel, and app-server evidence passed. |
| ES-06 | DONE | 2026-07-19 | Forty foundation inputs established the closed catalog; the ES-07 completeness audit brought it to 48 exact owner inputs. Dependency/manifests/fragments/profile/configuration/protocol/validation/traceability inputs validate without code/prose inference. |
| ES-07 | DONE | 2026-07-20 | Forty-six deterministic generated artifacts cover the normalized owner, dependency, descriptor, registry/integrity, binding, validation, closure/conformance/accounting/traceability, client, schema, state/codec/participant projections. Go/TS packaging, generation/drift, policy/shape/Harness, and exact BC-004–008 owner selectors pass at the recorded roots. |
| ES-08 | DONE | 2026-07-20 | Immutable artifact port, registry queries, explicit claims, dependency order, Base/extension collision admission, binding parity, Stage 6 component/plan construction, copy-safe DTOs, and app assembly admission pass 23 exact owner selectors plus boundary/unit/app-server gates. |
| ES-09 | DONE | 2026-07-20 | One generated-registry/resolved-claim epoch now drives the strict seven-member producer, tolerant inert decoder, reservations, dispatch, telemetry identity, and retained pre-side-effect publication plan; Network Flow is major 2 with no competing discovery. Exact passing and diagnosed failure roots are recorded above. |
| ES-10 | DONE | 2026-07-20 | Validation/configuration, lease/deadline, listener-bound publication, readiness/admission, fatal drain/no-restart, safe diagnostics, and exit `2`/`70` matrices pass at the exact ES-10 roots recorded above; transaction and worker halves remain explicitly assigned to ES-11. |
| ES-11 | DONE | 2026-07-20 | Forward coordination schema, generated runtime catalogs, named-owner state admission, bounded transactions, staged publication/cleanup, scoped portability preparation, digest-bound stopped-empty restore, job proofs, and published-worker loss pass the exact roots in the ES-11 completion snapshot. |
| ES-12 | DONE | 2026-07-20 | Build-bound standard support, no-store availability, exact browser intersection, epoch/generation reservation, lazy load/fallback, capability rejection, and extension-only disposal pass 24 local and 12 service-backed Extensions rows plus the recorded frontend/backend/generation/Harness gates. |
| ES-13 | DONE | 2026-07-20 | All 14 affected owner partitions, required evidence-class targets, security/OTel/Harness gates, paired shards, explicit manifests, and exact-root audits pass on one compatible snapshot; retained related failures were repaired without compatibility paths. |
| ES-14 | DONE | 2026-07-20 | Atomic owner promotion, post-finalizer exact owner evidence, independent evidence classes, 14 explicit-root audits, all-clause/158-criterion accounting, and serial broad/release gates passed on the recorded final snapshot. |

Each slice status update, evidence record, affected ledger update, and handoff-log
entry was completed before the next slice started.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| ES-00 | none | Freeze current route, envelopes, reservations, claim/config, workbook, telemetry, Network Flow, state, jobs, and accounting behavior. `requires later authorization` | Existing owner tests only | Missing characterization can hide drift. | Exact tests named in Section 4 before movement. | `make task-guide ROLE=module-author OWNER=<current-owner>` then narrow owner slices | Remove only new characterization tests if invalid; no production change. | Every critical freeze row has passing owner-aligned evidence or `BLOCKED: missing characterization`. |
| ES-01 | ES-00 | Revise the draft while retaining `status: draft`; close source lint, IDs, anchors, imports, and manifest model. `requires later authorization` | Draft and document tooling | Draft changes could invalidate all locators/digests. | Normative-source golden vectors and traceability extraction. | `make lint-markdown` plus future normative-source target | Revert the draft revision; preserve current owners. | Exact draft digest is accepted and contains no open delegation or required placeholder. |
| ES-01A | ES-01 | Apply the exact BC-001 through BC-017 rules to the draft; amend existing owning requirements; add EXT-AC-142 through EXT-AC-158; update Section 27 artifacts, clause traceability, canonical limits, and gate mappings; retain `status: draft`. `requires later authorization` | Draft and document tooling | A rule could remain split, contradictory, unbounded, or allocated to the wrong owner. | Every BC boundary vector and source-lint/traceability check. | `make lint-markdown` plus future normative-source and traceability targets | Revert the complete boundary-closure edit; preserve ES-01 source foundations. | Every BC row has exact draft anchors, one acceptance criterion, complete omission/bound/error behavior, and no conflicting normative sentence. |
| ES-02 | ES-01A | Amend Core 00, 01, 03, and 04 together; Core 02 confirms/adopts BC-001; allocate Core anchors for BC-001 through BC-003, BC-005, and BC-009 through BC-017. `requires later authorization` | Core owner documents | Accidental ownership transfer or partial generic discovery. | Owner-section traceability and current contract characterization. | Documentation/owner contract validation targets | Revert the complete companion-doc set together. | Core manifests/anchors are digest-bound; every allocated BC has an exact Core anchor; draft still not adopted. |
| ES-03 | ES-02 | Amend Network Flow and state-owning profile owners for BC-001, BC-009, BC-014, and BC-015; require Network Flow major 2; declare dependencies/state/init/migrations/bindings/codecs/jobs/participants/rebuilds. `requires later authorization` | Network Flow owner and profile owner inputs | Major/version mismatch; competing discovery remains. | Existing Network Flow full-owner rows plus planned major-2 selectors. | `make test-slice OWNER=module.networkflow`; affected browser slices | Revert owner inputs as a unit; retain current major 1 behavior. | No competing discovery item remains in proposed owner set; every allocated BC has an exact profile-owner anchor; no runtime change yet. |
| ES-04 | ES-02, ES-03 | Amend Incident Portability, physical backup, Reporting, and Composition for BC-003, BC-006, BC-012, and BC-014 where imported interfaces change; otherwise emit exact no-change parity. `requires later authorization` | Core shared sections and named NLSpecs | Generic interface could usurp shared-owner behavior. | Participant matrix and no-participation/no-change assertions. | Affected owner task guides/slices | Revert each owner amendment and its parity row together. | Every allocated BC resolves to an amended shared-owner anchor or exact digest-bound no-change parity result. |
| ES-05 | ES-02, ES-03, ES-04 | Register Harness v2 contracts/owner/families/selectors/topology, including exact selectors for EXT-AC-142 through EXT-AC-158, and revise OTel; amend Harness NLSpec only for a new public mechanic. `requires later authorization` | Verification registry, test-owner/family manifests, topology profiles, OTel owner | Zero-row owner, duplicate ownership, unregistered runner/profile, hard-coded OTel set. | Exact planned selectors, every BC scenario family, and OTel golden/privacy tests. | `make harness-contract`; owner task guide; `make otel-conformance` | Revert all `module.extensions` authored rows/profiles and OTel change together. | Both IDs resolve, family is nonempty, every planned BC criterion has a primary-owner verification ID and exact selector family, profiles validate, and OTel consumes canonical claim identity. |
| ES-06 | ES-01A, ES-05 | Author owner manifests/fragments and every Section 27 schema/input, including dependency declarations, validation surfaces, traceability mappings, inactive-value schemas, and operation-specific portability results; extend generator and generation manifests. `requires later authorization` | `contracts/**`, generator code, authored schemas/manifests | Hand-editing outputs, stale locators, incomplete condition inventories, inconsistent byte limits. | Canonicalization, locator, identity, declaration, validation-condition, traceability, bounds, zero-profile and overflow vectors. | `make generate`; `make generate-drift`; JSON shape and artifact policy | Revert authored inputs/generator; regenerate old outputs. | Every required authored input is owner-bound and generator consumes no prose search or implementation-derived fact. |
| ES-07 | ES-06 | Generate dependency snapshot, owner input, descriptors, registry/integrity, validation-condition registry, closure/manifests/accounting/traceability, bindings, state, codec, client, portability, and Base reservation projections. `requires later authorization` | Generated roots and packaged assets via generators only | Partial or stale generated set; phase-shaped alias survives. | Artifact identity/digest/parity and package-root tests, including BC-004 through BC-008. | `make generate-drift`; `make harness-contract` | Regenerate from last passing authored input commit; never edit outputs. | Full artifact set is byte-stable, current, packaged, and contains no phase/v1 compatibility alias. |
| ES-08 | ES-00, ES-02, ES-06, ES-07 | Introduce a deep Extensions coordination facade for immutable registry queries, claim resolution, bindings, dependencies, admission, and publication-plan output. `requires later authorization` | `internal/modules/extensions`, app/server ports, platform adapters | Facade could absorb Core/platform/profile ownership. | Unit selectors for claims/dependencies/collisions/bindings. | `make test-slice OWNER=module.extensions` | Keep current profile structs/route path behind adapter until parity passes. | Callers depend on narrow DTOs/ports; coordinator owns no profile behavior or transport/storage adapter. |
| ES-09 | ES-08 | Switch Core 01 discovery/reservation/dispatch and application assembly to generated registry facts and atomic Stage 6 publication. `requires later authorization` | HTTP runtime, server assembly, OpenAPI, Extensions query facade | Public response major change and partial routing. | Discovery/dispatch integration, process atomic publication, Base reservation parity. | Module/app/platform owner slices and `make backend-process` | Feature-branch rollback to characterized three-field/current dispatch implementation; no mixed producer. | One generic producer/decoder, one reservation registry, no listener/worker before serving, and no competing Network Flow discovery. |
| ES-10 | ES-08, ES-09 | Implement the BC-002, BC-006, BC-009, BC-010, BC-011, and BC-017 runtime matrices: Core 04 config views, validation precedence, process lease, deadlines, readiness/fatal lifecycle, validation registry diagnostics, and safe errors. `requires later authorization` | Config, app server, Extensions validation, telemetry/log adapters | Secret disclosure, wrong exit, partial startup, lease split brain. | Config, process lease/fatal, path/message/details/overflow, secret-negative tests. | Owner slices; `make go-gosec-targeted`; OTel conformance | Disable new coordinator before serving and fall back only to the fully characterized current runtime on the branch. | Exit 2/70, lease/readiness/timeout, component-loss, and diagnostic matrices all close. |
| ES-11 | ES-08, ES-10 | Implement the BC-001, BC-003, BC-011, BC-012, BC-013, BC-014, and BC-017 runtime matrices: state presence, metadata/ledger, initialization/migration, jobs/proof, Core 01 transactions, staged objects, backup codecs, portability/reporting participation. `requires later authorization` | Extensions, Core platform, profile modules, authored migrations | Durable partial effects, cross-owner access, data loss, unsupported restore. | All fresh/current/migrated/restored, ordered failure, cleanup and matrix selectors. | Service-backed full-owner slices; `make migration-drift` | Stop before schema publication where possible; never down-migrate committed profile state. | Every state/transaction/job/storage/participant result is closed, resumable, and owner-correct. |
| ES-12 | ES-07, ES-09, ES-10 | Implement BC-015 and BC-016: move browser to the standard support-registry/availability/auth intersection, prohibit capability activation, add lazy loading, fallback, stable identity, and exact cleanup. `requires later authorization` | `apps/web`, generated TS/UI contracts, workbook startup | Base cache/draft/queue loss, stale UI render, or unsupported capability activation. | Planned Vitest/Playwright browser/stateful selectors; every nonempty capability rejection; existing a11y/visual allocations. | `make frontend-typecheck`; unit and browser targets | Restore current discovery consumer as one complete version; never dual-read. | Unsupported/stale/unauthorized extensions and nonempty capabilities cannot render; Base state survives every transition. |
| ES-13 | ES-09, ES-10, ES-11, ES-12 | Execute full-owner slices and all required evidence-class gates; finalize paired shards and exact-root audit. `requires later authorization` | Harness artifacts only | Selected subset or broad target falsely claimed as closure. | Every active exact selector once; no unmapped/skipped/unexpected result. | Full-owner `test-slice`, service-backed slices, browser gates, `test-evidence-audit` | Retain failed roots for diagnosis; do not promote. | Compatible successful roots close every affected owner/target/row partition. |
| ES-14 | ES-13 | Run static registry/traceability/drift accounting and promote the draft and all companions together. `requires later authorization` | Owner statuses/manifests and generated projections | Intermediate adopted/current state. | All 158 acceptance criteria and 28 gates. | `make check`, `make harness-contract`, release-required gates, exact-root audit | Revert the entire promotion-status commit if any post-promotion gate fails; preserve implementation/data. | All 158 criteria and all gates are `DONE`, all digests/evidence current, and no older competing contract remains current. |

### Mandatory atomic-adoption sequence

The following order is mandatory and maps directly to ES-00 through ES-14. It may be
implemented as an ordered change series, but no intermediate change may mark the
Extensions NLSpec adopted/current or expose partially coordinated behavior:

1. characterize current behavior;
2. revise source/lint foundations and retain `status: draft`;
3. apply BC-001 through BC-017 to the draft;
4. amend Core owners;
5. amend profile and shared owners, recording exact no-change parity only where the
   imported interface does not change;
6. register Harness/OpenTelemetry inputs and exact boundary selectors;
7. author contracts and generate every Section 27 extension artifact,
   clause-traceability object, and v2 projection;
8. implement runtime, participant, and browser behavior;
9. execute owner evidence, explicit-root audit, and static registry,
   clause-traceability, and drift accounting;
10. promote the Extensions NLSpec and every required companion revision atomically.

## 8. Validation Plan

Commands are public Make targets discovered through `make help`, `make help-all`, and
`make explain-target`. Direct `go`, `pnpm`, Vitest, Playwright, and raw scripts are not
conformance commands. Run `make task-guide` before owner slices and
`make agent-finalize` before broad end-of-run verification in the later authorized
implementation task. `make agent-finalize` is intentionally not run for this
tracker-only task because it may refresh harness-maintenance artifacts.

### Boundary acceptance scenario inventory

Each planned criterion MUST cover every applicable enum token and state transition,
plus minimum, maximum, maximum-plus-one, empty, omitted, and explicit-null inputs.
The cases below are the minimum executable inventory; owner decision tables MAY add
cases but MUST NOT remove or merge outcomes that have different normative effects.

| BC ID | Minimum executable scenarios for its planned acceptance criterion |
| --- | --- |
| BC-001 | Every metadata/state-presence combination under `allowed` and `forbidden`; empty initialization; metadata-only input; omitted and explicit-null policy. |
| BC-002 | Invocation failure, structural invalidity, overflow, remaining schema defect, valid findings, and valid empty result; counts 0, 256, 257, 4096, and 4097. |
| BC-003 | Absent, malformed, incompatible, minimum, 64 MiB, and 64 MiB-plus-one payloads; prepared import, rollback, committed publication, and indeterminate commit; scoped staged-output denial. |
| BC-004 | Present empty arrays; omitted arrays; explicit null; missing, duplicate, extra, and stale dependency declarations; manifest version/digest mismatch. |
| BC-005 | Exactly one scalar source; missing and multiple scalar sources; duplicate set members; stale owner reference; attempted prose/code inference. |
| BC-006 | Complete schema annotations and procedural decision tables; missing, duplicate, stale, extra, and unregistered emitted conditions. |
| BC-007 | Every subject/contribution-kind mapping; generated subject with a not-applicable reason; every permitted and unrecognized fixed-baseline reason. |
| BC-008 | Every clause/parent kind; root and child scope; ordinal zero and maximum-plus-one; empty and adjacent half-open ranges; invalid overlap/out-of-bounds; digest and document-clause mismatch. |
| BC-009 | Syntax-only present, omitted, defaulted, explicit-null, reference-shaped, and structurally invalid values; prohibited view creation, resolution, retention, and egress attempts. |
| BC-010 | Initial acquisition success/timeout; every lease transition; uncertainty recovery on the original session; different-session proof; session loss; detection timeout; release; forbidden reacquisition after loss. |
| BC-011 | Checked and saturated local deadlines; inherited earlier/equal/later deadlines; cancellation before, at, and after deadline; cancellation before, during, and after proven commit; equal-deadline tiebreak and zero grace. |
| BC-012 | Participant counts 1 and 16384; count 0 and 16385; per-participant and aggregate bytes at 0, 64 MiB, and 64 MiB-plus-one; aggregate results at the same boundary; first-invalid ordering; malformed result; cancellation around every step/invocation. |
| BC-013 | Every staged-object initial default; noncandidate, expiry candidate, and retry candidate; upload failure; each deletion result; transaction-held deletion attempt; sweep overlap, missed interval coalescing, startup failure, and serving-dependency failure. |
| BC-014 | Stopped-empty, running, and nonempty targets; numeric group order and sequential binding order; validation failure isolation; inactive restore attempt; failed target serving attempt; derived rebuild before and after successful claim. |
| BC-015 | Required standard support row; omitted API-only profile; Network Flow major/workspace mismatch; stale/current generation; concurrent reservation; epoch rollover; minimum and overflow generation values. |
| BC-016 | Empty wire arrays; every nonempty capability fact/array surface; capability activation attempt; exact `extension_capability_not_supported` outcome. |
| BC-017 | Individual operation failure versus termination of each publication-plan listener, dequeue gate, and worker; readiness/admission closure, drain completion/timeout, committed-state preservation, exit 70, and prohibited restart. |

For this tracker-only revision, validation is limited to `git diff --check`,
`make lint-markdown`, `make json-shape-check`, and
`make generated-artifact-policy-check`, plus structural searches for all BC and
planned acceptance IDs, their owner/slice/verification/completion references, and
stale completion claims. Generation, formatting, implementation suites, and
`make agent-finalize` are out of scope because no production, contract, generated,
or Harness input changes in this revision.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| Current owner discovery | `make task-guide ROLE=module-author OWNER=<owner-id>` | Select narrow authoritative rows for each existing owner. | yes | `OWNER=module.extensions` currently fails until ES-05. |
| Owner explanation | `make explain-test-owner OWNER=<owner-id>` | Inspect exact owner rows/targets/profiles. | yes | Current Extensions owner failure is recorded evidence. |
| Extensions unit/full owner | `make test-slice OWNER=module.extensions` | Full default-owner unit/static rows. | no | Required after ES-05/ES-08; selected `ROWS=` is focus evidence only. |
| Extensions service-backed | `make service-backed-test-slice OWNER=module.extensions` | Integration/process/state/transaction partitions. | no | Must finish as full-owner closure, not a subset. |
| Backend unit | `make backend-unit` | Pure Go behavior and exact Go selectors. | no | Characterization may use current-owner projections before Extensions onboarding. |
| Backend store/integration/process | `make backend-store`, `make backend-integration`, `make backend-process` | PostgreSQL, HTTP, process lease/publication/fatal boundaries. | no | Required according to row evidence classes. |
| Migration | `make migration-drift` | Authored SQL against scratch database. | no | Required for state metadata/ledger or profile schema changes. |
| Frontend types/unit | `make frontend-typecheck`, `make frontend-unit` | Generated types, app controller, workbook state. | no | Both required for browser client changes. |
| Frontend import boundary | `make frontend-import-boundary-check` | Web/package dependency ownership. | no | Prevent generated/vendor/platform leakage. |
| Backend module boundary | `make backend-module-boundary-check` | Module/platform/application imports. | no | Enforce coordinator ports and prevent peer-internal imports. |
| Browser ordinary | `make browser-e2e-webserver-backed` | Discovery/support/auth and claimed/unclaimed routes. | no | Broad browser target closes only its exact selected rows. |
| Browser stateful | `make browser-e2e-stateful` | Authorization loss, epoch/generation rollover, state preservation. | no | Required for Core 03 cleanup/identity behavior. |
| Browser accessibility/visual/measurement | `make browser-e2e-a11y`, `make browser-e2e-visual`, `make browser-e2e-measurement` | Only rows whose primary owners require these evidence classes. | no | Do not promote design/support evidence into Base/extension conformance. |
| OpenTelemetry | `make otel-conformance` | Canonical claim signal, privacy, golden corpus. | no | Required after OTel amendment. |
| Security | `make go-gosec-targeted`, `make go-vulncheck` | New coordinator, config, storage, secret and egress surfaces. | no | Gosec audit remains advisory unless owner gates say otherwise. |
| Generation | `make generate`, then `make generate-drift` | Authored inputs to generated Go/TS/topology projections. | no | Run generation only in an authorized implementation task. |
| Generated policy | `make generated-artifact-policy-check` | Markers, roots, and lint-scope protection. | yes | Also required for this tracker handoff. |
| JSON/bootstrap shapes | `make json-shape-check` | Contracts, registries, manifests, topology inputs. | yes | Also required for this tracker handoff. |
| Harness contracts | `make harness-contract` | Verification, owner, family, runner, selector, topology, artifact schemas. | no | Required after every ES-05/ES-06/ES-07/ES-09 change. |
| Evidence audit | `make test-evidence-audit OWNER=<owner-id> EVIDENCE_ROOTS_FILE=<path>` | Explicit compatible retained roots for one owner. | no | No absent roots, newest-run lookup, or historical fallback. |
| Documentation | `make lint-markdown` | Authored Markdown structure. | yes | Required for this tracker and every owner-document slice. |
| Narrow developer gate | `make test-fast` | Fast combined loop after coherent code slices. | no | Does not replace service/browser/evidence-class gates. |
| Full developer gate | `make check` | Default full correctness projection. | no | Run after `make agent-finalize`; broad success does not replace omitted evidence classes. |
| Full/release | `make test`, `make release-check` | Full corpus and release aggregation when promotion requires it. | no | Core 05 remains excluded unless separately activated by a claim owner. |

## 9. Top-Level Work Tracker

### Planning and implementation work

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T-001 | Define target, authority, and exclusive write scope | WF-00 | DONE | none | Sections 1 and 2 | Target and non-goals are explicit. |
| T-002 | Inventory target and adjacent current state | WF-00 | DONE | T-001 | Sections 2 through 5 | All target files and material adjacent owners are accounted for. |
| T-003 | Map proposed requirements, criteria, gates, artifacts, and owners | WF-00 | DONE | T-002 | Section 6 ledgers | Every EXT ID and Section 27 artifact is mapped. |
| T-004 | Freeze current observable behavior | WF-01 | DONE | T-003 | Section 4 characterization matrix and ES-00 retained roots | Every critical current boundary has pre-change owner evidence. |
| T-005 | Coordinate draft and companion owners | WF-02 | DONE | T-004 | ES-01 through ES-05 | Extensions draft, Core, profile, shared-owner, Harness, telemetry, and domain amendments are staged without activation. |
| T-005A | Close selected boundary decisions in normative sources | WF-02A | DONE | T-005 | BC-001 through BC-017; ES-01A through ES-05 | Every BC has exact draft/owner anchors, one criterion, one verification ID, and one exact primary selector row. |
| T-006 | Build owner-manifest and canonical generation pipeline | WF-03 | DONE | T-005A | Forty-eight authored inputs, 46 generated projections, exact source/manifests/fragments, contractgen validation, generation/drift roots | All required authored inputs and schemas validate, generated packages are byte-stable, and no fact comes from implementation/prose inference. |
| T-007 | Implement runtime coordinator and atomic publication | WF-04 | DONE | T-006 | Immutable Extensions facade, one resolved claim/plan, listener-bound activation, application lease, readiness/admission and fatal lifecycle, exact coordinator/process rows | Callers receive narrow copy-safe DTOs and Stage 6 opens one public epoch only after successful bind under the held lease; loss closes admission and exits `70` without in-process restart. |
| T-008 | Adopt Core discovery/reservation/dispatch | WF-05 | DONE | T-007 | OpenAPI/route parity, generated clients, Network Flow major 2, and one retained pre-side-effect coordinator plan | One seven-member producer and inert tolerant decoder consume the same generated-registry/claim epoch as reservation/dispatch; no profile-local producer or dual reader remains. |
| T-009 | Implement state/migration/jobs/transaction/storage coordination | WF-06 | DONE | T-007 | Generated state/backup/participant catalogs, forward migration `00034`, platform stores, owner-bounded coordinators, and exact ES-11 roots | State, transaction, staging, job-proof, codec, restore, and participant matrices are owner-correct and resumable without down migration. |
| T-010 | Adopt browser support/lifecycle behavior | WF-07 | DONE | T-006, T-008 | Final standard support/asset binding, no-store availability, browser epoch/generation controller, lazy workspace boundary, and exact BC-015/016 rows | Only the current discovery/support/authorization/availability intersection renders; stale extension state is disposed without discarding Base state or stable identities. |
| T-011 | Adopt Network Flow and shared-owner participation | WF-08 | DONE | T-005A, T-009 | Network Flow owner counters/validator, exact transaction participants, portability preparation, and physical binding/codecs | Runtime adapters consume packaged owner facts without moving Network Flow, portability, Reporting, or backup behavior into Extensions. |
| T-012 | Onboard Harness v2 owner and evidence | WF-09 | DONE | T-006, T-008, T-009, T-010, T-011 | Two active verification IDs, exact authored routing, compatible full-owner/evidence-class roots, paired shards, and 14 explicit-root audits | Every current affected owner/target partition passes once with no missing, skipped, unexpected, or unmapped result. |
| T-013 | Audit and atomically promote | WF-10 | DONE | T-012 | All 28 gates, all-clause/158-criterion accounting, atomic companion promotion, and ES-14 final evidence | Every binary criterion is `DONE`, and all companions were promoted together. |
| T-014 | Independently distributed executable extensions | future | DEFERRED | none | Draft Section 30 | A later adopted NLSpec explicitly authorizes the surface. |

### Gate closure tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| G-000 | Establish normative boundary prerequisites | WF-02A | DONE | T-005A | Adopted owner anchors plus active EXT-AC-142 through EXT-AC-158 verification mappings | Every BC row has adopted owner anchors and active acceptance/verification mappings. |
| G-001 | Close EXT-GATE-001 through EXT-GATE-006 | WF-02A through WF-07 | DONE | G-000 | Adopted Core owner revisions and final owner evidence | All six Core gates pass. |
| G-002 | Close EXT-GATE-007 through EXT-GATE-011 | WF-02, WF-07, WF-08, WF-09 | DONE | G-001 | Adopted profile/shared/Harness/OTel/domain amendments or parity | All five companion gates pass. |
| G-003 | Close EXT-GATE-012 through EXT-GATE-016 | WF-03, WF-04, WF-06, WF-09 | DONE | G-001, G-002 | Generated/runtime/state artifacts and exact-root evidence | All five artifact, runtime, and state gates pass on the adopted snapshot. |
| G-004 | Close EXT-GATE-017 through EXT-GATE-022 | WF-03, WF-05, WF-07, WF-08, WF-09 | DONE | G-003 | Discovery, participant, closure, reservation, build-bound client support, availability, lifecycle, and exact-root evidence | All six discovery, client, and lifecycle gates pass on the adopted snapshot. |
| G-005 | Close EXT-GATE-023 through EXT-GATE-026 | WF-04, WF-06, WF-09 | DONE | G-003 | Process, migration, transaction, cleanup, failure-injection, and exact-root evidence | All four process and durable-coordination gates pass on the adopted snapshot. |
| G-006 | Close EXT-GATE-027 and EXT-GATE-028 | WF-09, WF-10 | DONE | G-004, G-005 | Full-owner paired shards, 14 explicit-root audits, source lint, and complete limit/traceability accounting | Both evidence/accounting gates and the complete static promotion audit pass. |

## 10. Session Handoff Log

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Planning complete; draft remains non-authoritative and pre-existing edit preserved. | `AGENTS.md`, skill/reference, framework, draft, Core 00-04, domain and named owners; touched tracker only. | `git status`, `git rev-parse`, `sha256sum`, `sed`, `rg`, `awk`, `wc` | Authority, 236 requirements, 141 criteria, and 28 gates indexed. | Later authorization for every implementation/adoption slice. | Begin ES-00 characterization in a separately authorized task. |
| 2026-07-19 | Codex implementation ES-00 | Characterization complete; draft remains non-authoritative. | Tracker, current owner manifests, discovery/dispatch/workbook/process/config/telemetry tests; tracker changed only. | Owner task guides and exact `test-slice`/`service-backed-test-slice` rows; `make otel-conformance` | All selected current-owner rows passed. Retained roots: `20260720T035330Z-p26902`, `p26905`, `p26926`, `p26939`, `p26958`, `p26975`; `20260720T035340Z-p27993`; `20260720T035350Z-p28275`; `20260720T035417Z-p35854`; `20260720T035515Z-p53386`; `20260720T035524Z-p53640`. | none | Validate the tracker checkpoint, then start ES-01. |
| 2026-07-19 | Codex implementation ES-01 | Draft source foundations verified complete; `status: draft` retained. | Extensions draft and tracker; tracker changed only. | Structural byte/ID/continuity counts, exact SHA-256, `make lint-markdown` | Draft remains digest `18fbd7f8...`; owner/dependency manifest model and source discipline are closed. | none | Validate the tracker checkpoint, then start ES-01A. |
| 2026-07-19 | Codex implementation ES-01A | Boundary closure complete in the non-authoritative Extensions draft. | Extensions draft and tracker. | Targeted source inspection and stale-contract searches; `git diff --check`; `make lint-markdown`; SHA-256 and AC continuity checks. | Draft `0.6.0` digest `796e119f...`; all BC-001 through BC-017 rules have exact draft anchors and exactly EXT-AC-142 through EXT-AC-158; no new requirement or gate ID; status remains draft. | none | Validate the tracker checkpoint, then start ES-02 companion-owner amendments. |
| 2026-07-19 | Codex implementation ES-02 | Core companion clauses and manifests staged; atomic-adoption guard retained. | Core 00-04, five Core owner manifests, Extensions draft, tracker. | Owner-section inventory; exact anchor/digest validator; `git diff --check`; `make lint-markdown`; `make json-shape-check`; `make generated-artifact-policy-check`. | Core ownership remains separated; every allocated Core anchor and half-open manifest range validates. JSON root `20260720T042625Z-p80431`; policy root `20260720T042628Z-p80773`. | no runtime activated | Validate this checkpoint, then begin ES-03 profile owners. |
| 2026-07-19 | Codex implementation ES-03 | Network Flow v2 profile-owner amendment staged; runtime remains characterized and unchanged. | Network Flow NLSpec, owner manifest/fragment, canonical state-presence, initialization, configuration, and portability-predicate inputs, tracker. | Exact canonical JSON/locator/range/digest validator; stale discovery/major search; `make task-guide ROLE=module-author OWNER=module.networkflow`; `make test-slice OWNER=module.networkflow`; documentation and artifact-policy gates. | Owner document `40c447d5...`, manifest `a2d23813...`, fragment `da957b30...`; sole generic discovery and exact profile facts declared; owner slice passed 118 tests at `20260720T043710Z-p90854`; JSON root `20260720T043659Z-p90211`; policy root `20260720T043701Z-p90553`. | no runtime activated; generic schemas/generation remain ES-06/07 | Validate this checkpoint, then begin ES-04 shared owners. |
| 2026-07-19 | Codex implementation ES-04 | Shared-owner amendment/parity decisions staged; ownership remains with Core 01, Reporting, and Composition. | Reporting and Composition NLSpecs/manifests, Snapshot/Reporting fragment, shared-owner resolution set, Core 01 imported anchors, tracker. | Exact canonical JSON/locator/range/digest validator; owner task guides; Incident Bundle and Reporting full owner slices; documentation and artifact-policy gates. | Incident Bundle passed 8/8 units and 17 tests at `20260720T044918Z-p135306`; Reporting passed 4/4 units and 10 tests at `20260720T044938Z-p135723`; JSON root `20260720T044847Z-p134454`; policy root `20260720T044848Z-p134795`. | `make task-guide ROLE=module-author OWNER=module.reportcomposition` failed `unknown active test owner`; this is the pre-existing missing Harness family assigned to ES-05, not a product-owner failure. | Validate this checkpoint, then begin ES-05 Harness/telemetry onboarding. |
| 2026-07-19 | Codex implementation ES-05 | Harness and telemetry companion work staged; Extensions draft remains non-authoritative. | Verification registry/contract, owner registry, Extensions and Report Composition family manifests, Extensions claim identity/tests, telemetry resource/bootstrap/tests, server assembly, OTel NLSpec, domain vocabulary, tracker. | Harness contract; owner guides/explanations; local and service-backed owner slices; `make otel-conformance`; app-server owner slice; `make format`; generation/shape diagnostics. | Harness root `20260720T050022Z-p143861`; Extensions roots `20260720T050109Z-p145059` and `20260720T050841Z-p165516`; Report Composition `20260720T050119Z-p145549`; OTel `20260720T050626Z-p149641`; app server `20260720T050653Z-p152679`. | `make generate` root `20260720T050808Z-p164484` and `make json-shape-check` root `20260720T050857Z-p165971` fail because ES-03/04 staged extension inputs are intentionally unknown to the pre-ES-06 contract generator; ES-06 owns that generator expansion and regenerated topology. | Validate this checkpoint, then start ES-06 by extending the generator before rerunning generation/shape gates. |
| 2026-07-19 | Codex implementation ES-06 pre-work | Authored-contract/generator work started after the ES-05 checkpoint passed. | Tracker plus exact branch/worktree/digest inventory. | `git branch --show-current`; `git rev-parse HEAD`; `git status --short`; owner `sha256sum`; `git diff --check`; `make lint-markdown`. | Dependencies ES-01A/ES-05 are done; source digests and dirty scope are recorded above. | Current generator intentionally rejects the new families until this slice implements their schemas and dispatch. | Extend the authored contract family and generator, then rerun `make generate` and drift/shape gates. |
| 2026-07-19 | Codex implementation ES-06 | Authored Section 27 inputs and their closed generator admission boundary are complete. | Extension input catalog, dependency/manifests/fragments, six profile configurations, build/state/client inputs, protocol/schema/validation/closure/traceability vectors, contractgen and tests, drift scratch inputs, generated Go/TS/topology outputs, characterization reader, tracker. | `make format`; `make generate`; `make generate-drift`; `make backend-unit`; `make json-shape-check`; `make generated-artifact-policy-check`; `make harness-contract`; checkpoint `git diff --check` and `make lint-markdown`. | Passing roots: generate `20260720T054424Z-p199375`, drift `20260720T054434Z-p200906`, backend unit `20260720T054707Z-p228024`, shape `20260720T054803Z-p232788`, policy `20260720T054805Z-p233131`, Harness `20260720T054807Z-p233309`. | Initial backend-unit root `20260720T054458Z-p203864` found a stale phase-index characterization reader; exact owner-fragment derivation replaced it and the rerun passed. | Validate this tracker checkpoint, then mark ES-07 in progress and generate normalized/runtime/accounting projections only from the admitted inputs. |
| 2026-07-20 | Codex implementation ES-07 | Generated Extensions artifact family and exact BC-004–008 selectors complete. | Eight completeness-audit inputs, contribution facts, six binding sources, Network Flow physical binding/four codecs, generated schema sources, closure baselines, contractgen projection/validation/tests, generated Go/TS packages, Extensions contract selectors, tracker. | `make format`; `make generate`; `make generate-drift`; `make backend-unit`; shape/policy/Harness gates; local and service-backed Extensions owner slices. | Forty-eight authored inputs produce 46 byte-stable generated artifacts. Passing roots: generate `20260720T062455Z-p270992`, drift `20260720T062500Z-p272501`, backend unit `20260720T062519Z-p275470`, shape `20260720T062638Z-p281673`, policy `20260720T062638Z-p281670`, Harness `20260720T062638Z-p281659`, owner `20260720T063205Z-p288085`, service-backed owner `20260720T063217Z-p288606`. | Generate root `20260720T061607Z-p250556` found an over-constrained empty participant list and owner root `20260720T063135Z-p287358` found a test-only clause-ID length error; both were corrected and reruns passed. Full 158-criterion traceability/accounting remains ES-13/14; final client asset digest remains ES-12. | Run tracker/doc checkpoint gates, then start ES-08 by recording its actual pre-work state before adding the coordinator. |
| 2026-07-20 | Codex implementation ES-08 pre-work | Registry-backed coordinator work started after ES-07 passed. | Tracker plus exact branch/worktree/draft/Core/catalog/generated-package digest inventory. | `git branch --show-current`; `git rev-parse HEAD`; `git status --short`; exact `sha256sum`; `git diff --check`; `make lint-markdown`. | ES-00/02/06/07 are done; ES-09 remains unstarted; scope is restricted to the generic coordinator and immutable output DTOs. | No owner contradiction; HTTP publication and lifecycle adoption intentionally remain later work. | Add narrow registry/claim/dependency/binding/publication-plan ports and exact unit selectors, then update this tracker before ES-09. |
| 2026-07-20 | Codex implementation ES-08 | Immutable Extensions coordination facade and app composition admission complete. | Coordinator/artifact port, safe findings, descriptors/claims/dependency order, collisions, binding admission, Stage 6 plan DTOs, five exact tests/family rows, Harness aggregate assertions, app-server runtime and test, generated topology/packages, tracker. | Format/generate/drift; Extensions local/service-backed slices; app-server owner slice; backend unit/boundary; Harness; shape/policy; diff/Markdown checkpoint. | Exact passing roots are recorded in the ES-08 completion snapshot; 23 Extensions selectors and 248 backend unit tests pass. | Initial roots `20260720T064123Z-p296675`, `20260720T064405Z-p302043`, and `20260720T064727Z-p312170` found a compile helper omission, Base exact-scope error, and expected Harness count drift; each rerun passed after the structural fix. | Validate this checkpoint, then record ES-09 pre-work before switching Core 01 discovery/reservation/dispatch to the one coordinator plan. |
| 2026-07-20 | Codex implementation ES-09 pre-work | Coordinated Core 01 publication switch started after ES-08 passed. | Tracker plus exact branch/worktree/draft/catalog/generated-package/coordinator/family digest inventory. | Branch/commit/status/digest inspection; `git diff --check`; `make lint-markdown`. | ES-08/T-007 are done; ES-10 is unstarted; scope is one discovery/reservation/dispatch/public-contract epoch. | No owner contradiction; lifecycle failure semantics remain ES-10. | Replace the platform hard-coded registry and three-member producer atomically, update contracts/clients, and close exact route/OpenAPI/browser-facing parity before ES-10. |
| 2026-07-20 | Codex implementation ES-09 | Coordinated Core 01 discovery/reservation/dispatch switch complete; draft remains gated. | Runtime claim/plan assembly, generated-registry HTTP projection, seven-member producer/OpenAPI/generated clients, tolerant decoder, Network Flow v2 contract family, contract tests, browser fixtures/support, tracker. | Format/generate/drift; shape/policy; Extensions, Protocol TS, web Network Flow, incident discovery, app-server process selectors; backend/frontend unit; typecheck/import boundary; Harness and tracker checkpoint. | Exact passing roots and diagnosed initial failures are recorded in the ES-09 completion snapshot; no literal fact registry, fragment reconstruction, competing Network Flow discovery, or dual client reader remains. | Network Flow full-owner retry root `20260720T070737Z-p378090` encountered only a shared-container startup conflict; ES-13 owns the isolated full evidence-class rerun. Lease/deadline/component-loss behavior remains ES-10. | Validate this checkpoint, then record ES-10 pre-work before changing configuration or process lifecycle behavior. |
| 2026-07-20 | Codex implementation ES-10 pre-work | Configuration/process lifecycle work started after ES-09 passed. | Tracker plus exact branch, commit, worktree count, Extensions/Core owner, coordinator, and family digests. | Branch/commit/status/digest inspection; prior ES-09 `git diff --check`, Markdown, and Harness checkpoints. | ES-09 is done; ES-11 is unstarted; scope is only BC-002/006/009/010, runtime BC-011, and process BC-017. | No owner contradiction; durable protocols and browser lifecycle intentionally remain later slices. | Implement closed validation/configuration/lease/deadline/fatal-lifecycle ports and exact selectors, then update this tracker before ES-11. |
| 2026-07-20 | Codex implementation ES-10 | Configuration and process-lifecycle adoption complete; staged owner documents remain unactivated. | Extensions validation/deadline/inactive configuration, platform config/process lease/process lifecycle/HTTP runtime, server assembly/process runner, exact tests/families, contract-family runtime packaging boundary, generated outputs, tracker. | Format/generate/drift; Extensions local/service-backed; config/HTTP runtime/app-server exact rows; backend unit/process; Gosec; OTel; web build; Protocol TS; shape/policy/Harness. | Exact passing roots and diagnosed initial failures are recorded in the ES-10 completion snapshot; BC-002/006/009/010 are implementation-complete and runtime BC-011/process BC-017 are closed. | Transaction BC-011 and dequeue-gate/worker BC-017 remain ES-11 by design; no owner contradiction. | Validate this checkpoint, then record ES-11 pre-work before adding migrations, transactions, storage, backup, or participants. |
| 2026-07-20 | Codex implementation ES-11 pre-work | Durable-state/shared-protocol work started after ES-10 passed. | Tracker plus exact branch, commit, worktree count, Core 01/02, Network Flow, shared-protocol, state/implementation binding, migration tip, and Extensions family inventory. | Owner task guide/explanation; source/artifact/package/migration inventory; exact SHA-256 and status inspection; prior ES-10 diff/Markdown/Harness checkpoint. | ES-10 is done; no generic durable coordinator or ledger exists; named-owner jobs/object/recovery/profile behavior remains outside Extensions. | No owner contradiction; browser lifecycle intentionally remains ES-12. | Add forward schemas and narrow coordination ports, then close BC-001/003/011-014 and durable BC-017 with exact owner evidence before ES-12. |
| 2026-07-20 | Codex implementation ES-11 | Durable-state and shared-protocol coordination complete; staged owner documents remain unactivated. | Authored state/backup/participant inputs; generated runtime catalogs; migration `00034`; extension platform store; Extensions state/transaction/portability/staging/backup coordinators; Network Flow owner adapter; server assembly; tests and tracker. | Generate/drift; Extensions local/service-backed; backend unit/store/integration/process; migration/boundary/security/shape/policy/Harness gates. | Exact passing roots and all diagnosed failures are recorded in the ES-11 completion snapshot; BC-001/003/011-014 and durable BC-017 are implementation-complete. | Atomic adoption and full-owner evidence remain ES-13/14; no owner contradiction. | Validate this checkpoint, then record ES-12 pre-work before changing browser lifecycle behavior. |
| 2026-07-20 | Codex implementation ES-12 | Browser lifecycle and build-bound client-support implementation complete; staged owner documents remain unactivated. | Web asset manifest/support packaging, server bootstrap validation, workbook availability, browser controller/context, Workbook/Network Flow request lifecycle, exact Go/Vitest/Playwright rows, tracker. | Format; frontend type/unit/import boundary; web build; backend unit/module boundary; Extensions local/service-backed slices; focused stateful row; generation drift; shape/policy/Harness. | BC-015/016 and T-010 are implementation-complete; exact passing and diagnosed failure roots are recorded in the ES-12 completion snapshot. | Atomic adoption and explicit-root full-owner accounting remain ES-13/14; no owner contradiction. | Validate this checkpoint, then record ES-13 pre-work before executing full-owner evidence partitions. |
| 2026-07-20 | Codex implementation ES-13 pre-work | Full-owner evidence execution started after ES-12 passed. | Tracker plus branch, commit, worktree count, staged owner-document, verification-contract, Extensions family, and owner-catalog digest inventory. | Branch/commit/status/digest inspection; prior ES-12 diff/Markdown checkpoint. | ES-09 through ES-12 are done; ES-14 is unstarted; this slice is restricted to current-snapshot evidence execution, defect repair, explicit-root manifests, and audits. | No owner contradiction; compatible affected-owner/target partitions must still be resolved and executed. | Run public owner guides/explanations, build the exact affected-owner matrix, then execute and audit every partition before updating this tracker. |
| 2026-07-20 | Codex implementation ES-13 | Full-owner and evidence-class accounting complete; no owner status promoted. | Fourteen affected owners, backend/frontend/browser/process/security/OTel/Harness targets, recovery browser helper, incident-support browser test, explicit ignored evidence manifests, tracker. | Full local/service-backed owner slices; all required evidence classes; 14 explicit `test-evidence-audit` invocations; boundary/type/security/OTel/Harness gates. | Every current exact row passed on one compatible snapshot with paired accounting/summary shards; related retained failures were structurally repaired and rerun. | Complete 158-criterion/all-clause static accounting and atomic adoption remain ES-14 only. | Validate this tracker checkpoint, then record ES-14 pre-work state before changing any adoption status. |
| 2026-07-20 | Codex implementation ES-14 pre-work | Atomic validation/promotion slice started after ES-13 passed. | Tracker plus branch, baseline, worktree count, owner-document, verification, Harness, and authored-catalog digests. | Branch/commit/status/digest inspection; ES-13 checkpoint gates. | Dependencies and exact pre-promotion snapshot are recorded; prior evidence is explicitly invalid after promotion changes the source snapshot. | Complete all-158/all-clause accounting and one coordinated status/digest transition before claiming adoption. | Audit status/digest mechanics and close static traceability before editing adoption state. |
| 2026-07-20 | Codex implementation ES-14 | Atomic promotion and final handoff are complete. | Extensions/Core/profile/shared-owner documents and manifests; contracts/generated projections; runtime/frontend/Harness repairs; duration baselines; migration `00034`; final evidence manifests; tracker. | Finalization; static/drift gates; 14 full-owner and eight service-backed slices; independent backend/frontend/browser/security/OTel evidence; 14 explicit-root audits; `test-fast`, `check`, `test`, and `release-check`. | All 158 criteria, 28 gates, affected owner partitions, evidence classes, audits, and broad gates pass on the recorded final identity; rejected scheduler and interrupted roots remain classified and excluded. | none | Preserve the staged atomic change set for review; do not commit or push without separate authorization. |

### Backend boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Current target is a three-file thin HTTP facade; profile/reservation/claim composition is adjacent. | Target files, `httpapi/extensions.go`, `httpapi.go`, server runtime/routes, Network Flow and shared-owner modules. | `find`, `rg`, `sed` | Boundary split and current call paths recorded. | No generic coordinator/state/ledger/binding implementation exists. | Add owner-aligned characterization before moving responsibilities. |
| 2026-07-19 | Codex implementation ES-05 | Immutable claim identity foundation exists without moving transport/storage/profile behavior. | `internal/modules/extensions/claims.go`, contract tests, server telemetry composition. | Extensions local/service-backed slices; app-server owner slice. | Claim IDs canonicalize once, are immutable to callers, and bind to the exact canonical digest consumed by telemetry. | Full registry-backed claim resolution remains ES-08. | ES-06/07 generate registry facts before ES-08 replaces the current HTTP-profile source. |
| 2026-07-20 | Codex implementation ES-10 | Core 04 runtime boundary is closed without moving transport, storage, or profile behavior into Extensions. | Extensions validation/configuration/deadline ports; platform config, lease, lifecycle, HTTP runtime; app-server composition. | Exact Extensions/config/HTTP-runtime/app-server rows; backend unit/process; targeted Gosec. | Immutable coordinator inputs drive inert config admission and process publication; dedicated platform adapters own PostgreSQL lease and listener/process mechanics. | Durable transaction/storage/job adapters remain ES-11. | Validate ES-10, then inventory the existing Core 01 durable protocol owners before ES-11 edits. |
| 2026-07-20 | Codex implementation ES-11 | Generic durable coordination is narrow and named-owner behavior stays outside Extensions. | `internal/modules/extensions` coordinators, `internal/platform/extensionstore`, Network Flow state adapter, server composition, migration `00034`. | Extensions owner slices; backend unit/store/integration/process; module boundary and migration drift. | Extensions consumes immutable catalogs and narrow storage/owner ports; Network Flow retains authoritative table knowledge and validation; platform owns PostgreSQL, object, and job-proof mechanics. | Browser lifecycle remains ES-12; owner promotion remains ES-14. | Preserve these import/ownership boundaries while adding client availability. |
| 2026-07-20 | Codex implementation ES-12 | Core 01 availability and packaged-browser admission remain outside the Extensions facade. | Workbook startup registry/routes, webassets validator/bootstrap, server publication binding, narrow coordinator digest rebinding. | Backend unit/module boundary; Extensions local/service-backed slices; focused stateful row. | Authorization stays workbook-owned, asset verification stays platform-owned, and Extensions receives only an immutable final support digest before Stage 6. | Owner promotion remains ES-14. | Preserve these ports during ES-13 evidence execution. |

### Frontend boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Web consumes current discovery and claim state; no support-registry/availability generation exists. | App shell client, App, debug harness, workbook startup/shell tests, browser support and Network Flow surfaces. | `rg`, `sed` | Core 03/web ownership and browser characterization plan recorded. | Target behavior requires coordinated client/server contract changes. | Implement only after support and availability artifacts generate. |
| 2026-07-20 | Codex implementation ES-12 | Browser eligibility and lifecycle have one explicit owner boundary. | Extension availability controller/context/tests, Workbook shell/runtime, Network Flow feature/controllers, standalone identity constants, stateful Playwright evidence. | Frontend typecheck/unit/import boundary; web build; focused and full Extensions service-backed slices. | Only the exact four-way current intersection can lazy-load; request completions are generation-guarded; extension disposal leaves Base state and stable identities intact. | Full cross-owner evidence accounting remains ES-13. | Run existing affected browser partitions from explicit ES-13 roots without changing lifecycle semantics. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | One phase-shaped authored extension index feeds current Go/TS projections. | Extension/OpenAPI contracts, family registry, contractgen validator, generated policy and generated consumers. | `rg`, `sed` | Full Section 27 authored/generated ownership plan recorded. | Owner manifests/fragments and generator do not exist. | ES-06 authors inputs/generator, then ES-07 generates outputs. |
| 2026-07-19 | Codex implementation ES-05 | Authored Harness topology changed; generated topology remains deliberately pending on ES-06. | Verification and family inputs; current contractgen validation path. | `make generate`; `make json-shape-check`; failure-artifact inspection. | Failures are deterministic: current contractgen rejects the staged `contracts/extensions/**` families before topology rendering. No generated root was hand-edited. | Pre-ES-06 generator recognizes only `contracts/extensions/index.json`. | ES-06 expands the contract family and reruns `make generate`; ES-07 verifies byte-stable output. |
| 2026-07-20 | Codex implementation ES-10 | Browser runtime packaging is an explicit subset of each generated contract family. | Contract-family registry/schema, contractgen filtering/validation/tests, generated Go/TS packages, OTel source/bundle scanner. | Generate/drift, JSON shape, generated policy, Protocol TS, web build, OTel conformance. | Extensions owner/accounting evidence remains packaged for Go/runtime admission but cannot leak server-only authority vocabulary into browser bundles; only client-support, descriptors, and registry projections enter TS runtime. | Final client asset-set digest remains ES-12. | Preserve this closed packaging allowlist while adding ES-11 server-only durable artifacts. |
| 2026-07-20 | Codex implementation ES-11 | Durable catalogs are generated from exact owner inputs and packaged only for server admission. | Two Network Flow transaction participants, implementation/physical bindings, generated schema sources and state/backup/participant registries, contractgen validator/tests. | Format/generate/drift; JSON shape; generated policy; Harness contract. | Byte-stable catalogs bind state, codecs, restore order, and participants without implementation/prose inference; TS runtime packaging remains unchanged. | Client asset-set digest and availability projection remain ES-12. | Add only browser-owned projections in ES-12. |
| 2026-07-20 | Codex implementation ES-12 | Final browser support identity is derived from actual packaged bytes. | Authored support source, embed-web-assets generator/script/tests, asset archive/sidecars, platform webassets validator, Make/task-surface inputs. | Web build and embedded-assets generation; generation drift; JSON shape; generated policy; backend unit. | Every archive entry is length/digest-bound, extras and omissions fail startup, and the final standard registry binds exactly Network Flow major 2/network-analysis with empty capabilities. | Adoption accounting remains ES-13/14. | Use final build artifacts, not the source projection, for publication evidence. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Harness v2 is adopted/current; `module.extensions` is absent. | Verification registry/owners, test-owner registry, Network Flow family, runner registry, execution topology/profiles/schemas. | `make help`; `make help-all`; `make task-guide ROLE=module-author OWNER=module.extensions`; `make explain-test-owner OWNER=module.extensions`; `make explain-target` for required doc/policy targets | Help/explain-target commands succeeded; both Extensions owner commands failed with `unknown active test owner module.extensions`. | Owner/family/topology cannot close until ES-05 and executable selectors exist. | Add both verification IDs and nonempty exact-selector rows atomically. |
| 2026-07-19 | Codex implementation ES-05 | Extensions and Report Composition are active Harness owners. | Both verification IDs, 18 Extensions rows, one Composition row, catalog integrity tests. | `make harness-contract`; both owner task guides/explanations; Extensions local/service-backed slices; Composition local slice. | Catalog summary is 48 owners, 174 families, 852 rows, and 1396 selectors; every BC maps once; all selected rows passed without missing or unmapped results. | Runtime rows remain `informative` until their implementation slices provide behavioral matrices. | Preserve row identities and replace posture/evidence in ES-08 through ES-13. |
| 2026-07-20 | Codex implementation ES-10 | Four exact config/process selectors and the BC-002/006/009/010/011/017 matrices execute under active owners. | Extensions, app-server, platform-config and HTTP-runtime family inputs; Harness aggregate expectations; exact Go/process tests. | Harness contract; Extensions local/service-backed; exact app-server rows; backend unit/process. | Harness validates 861 rows/1409 selectors; ES-10 exact rows pass without missing/unmapped results, including PostgreSQL lease and fatal-process cases. | ES-11 must replace the remaining durable-protocol informative matrices; ES-13 reruns full partitions. | Validate checkpoint and then update ES-11 to `IN_PROGRESS` before any durable implementation. |
| 2026-07-20 | Codex implementation ES-11 | Exact durable protocol matrices execute under the active Extensions and platform owners. | BC-001/003/011-014/017 tests, service-backed state/store coverage, app-server lifecycle integration, Harness family inputs. | Extensions local/service-backed; backend unit/store/integration/process; Harness contract. | Final roots pass including 255 backend unit, 104 store, 180 integration, and 36 process tests; exact participant, byte, state, restore, cleanup, and worker-loss boundaries execute. | ES-13 must rerun full owner/target partitions and build explicit evidence-root manifests. | Validate checkpoint, then start ES-12 only after its tracker pre-work update. |
| 2026-07-20 | Codex implementation ES-12 | Exact browser lifecycle evidence is active under the Extensions owner. | BC-015/016 Go rows, additional browser-stateful row, five-family Extensions manifest, generated schedules/topology, aggregate Harness assertions. | Full local/service-backed Extensions slices; focused stateful selector; Harness contract. | Twenty-four local and twelve service-backed tests pass with no missing/unmapped rows; the stateful row verifies no-store availability, backend bootstrap, lazy loading, and Base client identity continuity. | ES-13 must create compatible explicit-root manifests for every affected owner/target partition. | Validate checkpoint, then begin ES-13 with actual source/evidence snapshots. |
| 2026-07-20 | Codex implementation ES-13 | Exact implementation evidence and accounting are complete for every affected owner. | Fourteen owner plans/manifests; retained local/service-backed/evidence-class summaries; ignored explicit root manifests. | Public owner guides/explanations; full owner slices; service-backed slices; evidence classes; `make test-evidence-audit` per owner; Harness contract. | All owner counts and exact roots are recorded in the ES-13 completion snapshot; audits report no missing, skipped, unexpected, unmapped, stale, or mixed-snapshot evidence. | Static all-clause/158-criterion promotion accounting remains ES-14. | Preserve these roots for the pre-promotion snapshot; rerun affected partitions after ES-14 changes the source snapshot. |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Current auth/session handling is in route adapter; config and OTel hard-code parts of profile handling. | Extensions route, platform auth/config/telemetry, Network Flow security/key-ring code, Core 04 and OTel owner. | `rg`, `sed` | Secret, egress, syntax-only inactive, lease/fatal and telemetry risks mapped. | Core 04/OTel companion amendments and implementation authorization. | Preserve current outcomes until ES-10 passes negative/security evidence. |
| 2026-07-19 | Codex implementation ES-05 | Telemetry no longer owns or filters the profile vocabulary. | OTel resource/bootstrap, privacy/log/resource tests, server input adapter, OTel/domain docs. | `make otel-conformance`; app-server owner slice. | Bootstrap requires a digest-bound canonical identity, rejects malformed/unsorted/duplicate/base/mismatched input, and admits a future syntactically valid owner-resolved profile without telemetry code change. | Canonical source remains the current server profile resolver until ES-08. | Feed the same type from the generated-registry coordinator in ES-08. |
| 2026-07-20 | Codex implementation ES-10 | Inactive configuration and fatal diagnostics are closed negative-security surfaces. | Inert validator, configuration validation, lease/lifecycle safe reasons, server JSON fatal reporting, OTel contract-family scanner. | Targeted Gosec, OTel conformance, exact config/process/Extensions rows. | No inactive secret/reference resolution, retained value, profile callback, egress, or unredacted fatal payload exists; server-only owner evidence is excluded from browser runtime bundles. | ES-11 must preserve these boundaries for participant findings, staged objects, and backup/restore. | Apply the same generated-condition and safe-finding enforcement to ES-11 durable protocols. |
| 2026-07-20 | Codex implementation ES-11 | Durable operations expose bounded capabilities and safe closed outcomes. | Scoped portability staging, ordered transaction results, staged-object store/janitor, stopped-empty restore, state findings. | Targeted Gosec; exact Extensions and service-backed matrices; backend store/integration/process. | Import preparation cannot publish/redeem or mutate outside the shared transaction; payloads and participant counts are bounded; cleanup performs no object I/O in a database transaction; restore cannot serve a failed target. | Full vulnerability scan remains an ES-13/14 aggregate gate. | Preserve these negative surfaces during browser adoption. |
| 2026-07-20 | Codex implementation ES-12 | Unsupported or stale browser execution closes before extension behavior. | Strict support/startup decoders, capability rejection, availability reservation, asset-contract startup validation, Network Flow guarded requests. | Frontend unit; backend unit; exact BC-015/016 and stateful rows. | Malformed bootstrap, nonempty capability arrays, unsupported major/workspace, authorization loss, stale generation, epoch overflow, or RNG failure cannot render or activate extension behavior. | Aggregate security/vulnerability gates remain ES-13/14. | Preserve these fail-closed surfaces during broad verification. |
| 2026-07-20 | Codex implementation ES-13 | Aggregate negative-security evidence passes on the final implementation snapshot. | Backend/frontend/process/browser partitions and generated/runtime security boundaries. | `make go-gosec-targeted`; `make go-vulncheck`; OTel conformance; stateful/webserver-backed browser gates; exact owner audits. | Targeted Gosec root `20260720T113841Z-p1566027` and vulnerability root `20260720T114027Z-p1593169` pass; no secret, inactive egress, unsupported capability, stale render, or unsafe fatal behavior was introduced. | Promotion changes the evidence snapshot; ES-14 must rerun release-required security gates. | Preserve fail-closed behavior through atomic owner promotion. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Tracker decision selection is complete; target-contract boundary closure remains blocked. | Tracker plus recorded inspected corpus and informative `temp/analysis-notes.md`. | Required validation commands and final worktree/hash audit. | Selected BC-001 through BC-017 recorded; no normative or production implementation performed. | Normative adoption, implementation, and evidence remain blocked on later authority. | Next authorized session starts at ES-00 and proceeds through ES-01A before generation or production edits. |
| 2026-07-19 | Codex implementation ES-05 | Specification/owner/Harness coordination is complete; generation is next. | ES-05 changes and failure summaries. | Exact owner tests and failure inspection. | No owner contradiction or secret-bearing diagnostic was found. | Expected generator capability gap blocks `make generate`/shape until ES-06; this is not `BLOCKED: owner contradiction`. | ES-06 first admits all staged authored contract families, then updates topology and shape evidence. |
| 2026-07-20 | Codex implementation ES-12 | Browser implementation is complete; only evidence consolidation and adoption remain. | ES-12 source, contract, generated, test, and retained-root inventory. | Slice-specific gates and failure-artifact inspection recorded above. | No owner contradiction remains; production bootstrap and Vite preview are deliberately distinct, and no dual contract reader was introduced. | RB-002 atomic owner promotion and RB-006 explicit compatible evidence roots remain blocked on ES-13/14. | Validate this tracker checkpoint, then start ES-13 pre-work and build explicit evidence-root manifests. |
| 2026-07-20 | Codex implementation ES-13 | Evidence consolidation is complete; adoption is the sole remaining workstream. | ES-13 exact roots, failure artifacts, explicit manifests, owner/harness summaries, and tracker. | Complete owner/evidence/security/OTel/Harness matrix and explicit-root audits. | RB-003 through RB-007 are resolved; no owner contradiction, missing partition, unsafe compatibility path, or unexplained failure remains. | RB-001, RB-002, and RB-008 remain until ES-14 atomically promotes the complete owner set. | Validate the ES-13 tracker checkpoint, then begin ES-14 only after its pre-work update. |

### Tracker validation

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-07-19 | Codex planning session | Final tracker validation | Tracker; retained summaries under `.cartulary/test-results` are ignored runtime evidence, not authored changes. | `git diff --check`; `make lint-markdown`; `make json-shape-check` | All passed. JSON-shape run root: `.cartulary/test-results/20260719T223232Z-p9`. | none | Preserve results in handoff. |
| 2026-07-19 | Codex planning session | Generated policy validation | Generated-artifact policy and retained target summary | `make generated-artifact-policy-check`; `make explain-run RESULTS_DIR=.cartulary/test-results/20260719T223232Z-p9 TARGET=generated-artifact-policy-check DETAIL=summary`; escalated rerun of `make generated-artifact-policy-check` | Initial sandboxed run failed as harness `unknown_failure` because `spawnSync git` returned `EPERM`; retained summary identified the sandbox cause. Escalated exact rerun passed at `.cartulary/test-results/20260719T223316Z-p21698`. | none | Cite the successful rerun and the diagnosed initial failure. |
| 2026-07-19 | Codex boundary-closure revision | Tracker-only boundary validation | Tracker only; no normative, contract, generated, Harness, or production input changed. | `git diff --check`; `make lint-markdown`; `make json-shape-check`; `make generated-artifact-policy-check`; structural BC/AC and stale-claim searches | All passed. JSON-shape root: `.cartulary/test-results/20260720T034056Z-p18996`; generated-policy root: `.cartulary/test-results/20260720T034110Z-p19402`. | Normative adoption remains blocked. | Skip generation, formatting, implementation suites, and `make agent-finalize` because this is a tracker-only revision. |
| 2026-07-19 | Codex implementation ES-00 | Characterization checkpoint validation | Tracker plus retained exact current-owner evidence; no production or normative source changed. | `git diff --check`; `make lint-markdown`; `make json-shape-check`; `make generated-artifact-policy-check` | All passed. JSON-shape root: `.cartulary/test-results/20260720T035608Z-p56760`; generated-policy root: `.cartulary/test-results/20260720T035608Z-p56773`. | none | ES-00 is complete; ES-01 may start. |
| 2026-07-19 | Codex implementation ES-01 | Draft-foundation checkpoint validation | Tracker and unchanged draft. | `git diff --check`; `make lint-markdown`; structural source/ID/continuity checks | All passed; draft digest remains `18fbd7f8c83e4a92ceec5bb0913a5443f454d8be4f96a2397df8627ec0915ee8`. | none | ES-01 is complete; ES-01A may start. |
| 2026-07-19 | Codex implementation ES-01A | Boundary-closure checkpoint validation | Draft and tracker. | `git diff --check`; `make lint-markdown`; exact requirement/acceptance/gate and stale-contract searches | All passed; 236 requirement IDs, continuous acceptance IDs through 158, 28 gates, draft digest `796e119fe4d74386027074ddf9954d1089cac72c7bb3a2fe79c231e24ffc3c62`. | none | ES-01A is complete; ES-02 may start only after its pre-work tracker update. |
| 2026-07-19 | Codex implementation ES-02 | Core-companion checkpoint validation | Core 00-04, five Core owner manifests, and tracker. | `git diff --check`; `make lint-markdown`; exact document/anchor range and SHA-256 validator; `make json-shape-check`; `make generated-artifact-policy-check` | All passed. Core documents: `3358cc76...`, `9ad58e13...`, `79803b2d...`, `79afa553...`, `67583e75...`; manifest digests recorded in the companion table. | none | ES-02 is complete; ES-03 may start after its pre-work tracker update. |
| 2026-07-19 | Codex implementation ES-03 | Profile-owner checkpoint validation | Network Flow NLSpec, manifest/fragment, four canonical profile inputs, tracker. | Exact canonical JSON/locator/range/digest validator; `git diff --check`; `make lint-markdown`; `make json-shape-check`; `make generated-artifact-policy-check`; full `make test-slice OWNER=module.networkflow` | All passed. Document/manifest/fragment digests are `40c447d5...`, `a2d23813...`, `da957b30...`; owner slice passed 4/4 work units and 118 tests at `20260720T043710Z-p90854`. | none | ES-03 is complete; ES-04 may start only after its pre-work tracker update. |
| 2026-07-19 | Codex implementation ES-04 | Shared-owner checkpoint validation | Reporting/Composition owner docs and manifests, Snapshot/Reporting fragment, resolution set, tracker. | Exact canonical JSON/locator/range/digest validator; `git diff --check`; `make lint-markdown`; `make json-shape-check`; `make generated-artifact-policy-check`; active Incident Bundle and Reporting owner slices | All applicable gates passed. Reporting/Composition documents `72f8948f...`/`489a839d...`; manifests `3deed271...`/`6c3bc194...`; resolution set `97f39cf2...`. | Composition owner-guide failure `unknown active test owner module.reportcomposition` is assigned to ES-05 Harness onboarding. | ES-04 is complete; ES-05 may start only after its pre-work tracker update. |
| 2026-07-19 | Codex implementation ES-05 | Harness/telemetry checkpoint validation | Verification/family inputs, claim identity, telemetry/server changes, OTel/domain docs, tracker. | `git diff --check`; `make lint-markdown`; `make harness-contract`; `make generated-artifact-policy-check`; exact owner and OTel tests listed above. | All ES-05-owned gates passed. Markdown root `20260720T051412Z-p167856`; Harness root `20260720T051422Z-p169287`; policy root `20260720T051444Z-p170139`. | `make generate`/`make json-shape-check` failures are recorded at roots `20260720T050808Z-p164484`/`20260720T050857Z-p165971` and assigned to the predeclared ES-06 generator expansion; no generated output was accepted. | ES-05 is complete; ES-06 may start only after its pre-work tracker update. |
| 2026-07-20 | Codex implementation ES-10 | Configuration/process checkpoint validation | ES-10 implementation, exact evidence, and completed tracker state. | `git diff --check`; `make lint-markdown`; `make harness-contract`; prior slice-specific generation, owner, process, security, OTel, and web gates recorded above. | Diff passed; Markdown root `20260720T081236Z-p598843` and Harness root `20260720T081301Z-p600365` passed with no missing, skipped, failed, or unmapped work. | none | ES-10 is complete; ES-11 may start only after its actual pre-work tracker update. |
| 2026-07-20 | Codex implementation ES-11 | Durable-state/shared-protocol checkpoint validation | ES-11 implementation, exact evidence, artifact/BC/T/G states, and completed handoff rows. | `git diff --check`; `make lint-markdown`; prior slice-specific generation, owner, migration, backend, security, shape, policy, and Harness gates recorded above. | Diff passed; Markdown root `20260720T090818Z-p726470` passed after the ES-11 tracker completion update. | none | ES-11 is complete; ES-12 may start only after its actual pre-work tracker update. |
| 2026-07-20 | Codex implementation ES-12 | Browser-lifecycle checkpoint validation | ES-12 implementation, exact evidence, artifact/BC/T/G states, and completed handoff rows. | `git diff --check`; `make lint-markdown`; prior slice-specific browser, frontend, backend, generation, shape, policy, and Harness gates recorded above. | Diff passed; Markdown root `20260720T101336Z-p937550` passed after the ES-12 tracker completion update. | none | ES-12 is complete; ES-13 may start only after its actual pre-work tracker update. |
| 2026-07-20 | Codex implementation ES-13 | Full-owner evidence checkpoint validation | ES-13 implementation repairs, exact roots/audits, T/G/blocker/checklist states, and handoff rows. | `git diff --check`; `make lint-markdown`; all slice-specific owner, evidence-class, security, OTel, Harness, and audit gates recorded above. | Diff passed; Markdown root `20260720T114424Z-p1596330` passed after the ES-13 completion update. | none | ES-13 is complete; ES-14 may start only after its actual pre-work tracker update. |
| 2026-07-20 | Codex implementation ES-14 | Post-adoption tracker and document validation | Completed tracker, adopted owner set, Harness catalogs/topology, and staged implementation. | `git diff --check`; `make lint-markdown`; `make harness-contract`; prior finalization, drift, owner, evidence-class, audit, broad, and release roots recorded in the ES-14 completion snapshot. | Diff passed; Markdown root `20260720T144102Z-p3199338` and Harness root `20260720T144118Z-p3200785` passed. | none | Repeat the final diff/Markdown/Harness trio after recording this row, then hand off the staged change set without commit or push. |

## 11. Open Questions and Blockers

No blocker to safe implementation or adoption remains. The table retains the
historical blockers and their exact resolution workstreams. No design question
remains open for BC-001 through BC-017; reopening an adopted rule requires a
separately authorized normative change and corresponding tracker revision.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | The target NLSpec was draft and its coordinated adoption gates could not be treated as current behavior. | Implementing the target over adopted owners would have created competing contracts. | Coordinated owner-document promotion plus all gate evidence. | RESOLVED by ES-14 |
| RB-002 | Network Flow was adopted at document version 1.2.0/contract major 1 while BC-015 required major 2. | Discovery and client compatibility could not be changed piecemeal. | Core 00 status/major action plus Network Flow/Core 01/03/04 companion adoption and major-2 owner evidence. | RESOLVED by ES-14 |
| RB-003 | No `module.extensions` verification contract, owner row, family manifest, or topology profile exists. | Owner slices, paired shards, and evidence audit cannot close. | ES-05 Harness-authored inputs and implemented exact selectors. | RESOLVED by ES-05 and ES-13 |
| RB-004 | Generic extension owner manifests, generator, registry integrity, bindings, state metadata/ledger, codecs, closure, and accounting artifacts do not exist. | Runtime admission or adoption cannot be inferred from code/routes/database. | ES-06/ES-07 contract/generator work plus ES-11 durable projections. | RESOLVED by ES-06, ES-07, and ES-11 |
| RB-005 | Current code lacks the process lease, Stage 6 atomic publication, generic migration/final-validation coordinator, and complete transaction/staged-object protocol. | Failure cases could publish partial behavior or corrupt durable state. | ES-08 through ES-11 implementation plus exact process/service-backed evidence. | RESOLVED by ES-08 through ES-13 |
| RB-006 | No compatible successful retained roots exist for future Extensions rows. | Static registry success or broad target success cannot prove execution. | ES-13 full-owner/evidence-class runs and explicit evidence-root manifests. | RESOLVED by ES-13 |
| RB-007 | Every production, test, owner, contract, migration, config, generated, and harness change is outside this tracker-only task. | The planning write boundary must not be mistaken for implementation authority. | Explicit implementation authorization. | RESOLVED by the current implementation request |
| RB-008 | The 17 selected boundary rules were not present in adopted owner documents. | Generation or implementation against the pre-closure draft would have preserved ambiguity and could have created divergent contracts. | ES-01A through ES-04 owner amendments, ES-05 active verification mappings, G-000 closure, and ES-14 atomic promotion. | RESOLVED by ES-14 |

If companion review reveals two adopted primary owners claiming incompatible behavior,
add a new blocker with the exact text `BLOCKED: owner contradiction`, cite both exact
owner anchors/digests, and stop the affected workflow. No such contradiction was
established during this implementation and promotion effort.

## 12. Binary Completion Criteria

Planning is complete only when every tracker criterion below is true; implementation
is complete only when every implementation/adoption criterion is `DONE`. Code existing
or a broad test target passing is never sufficient by itself.

### Tracker handoff criteria

- [x] Every file under `internal/modules/extensions`, including testsupport, is
  inventoried.
- [x] Current callers, assembly, platform dependencies, route registration, tests,
  frontend consumers, contracts, migrations, generated outputs, configuration, and
  Harness v2 mappings are represented by opened repository evidence.
- [x] Authored inputs, generated outputs, runtime state, retained evidence, and
  normative owner documents are separated.
- [x] Current and target boundaries cover discovery, claims, dispatch, generation,
  bindings, state, migrations, jobs, transactions, backup/restore, portability,
  reporting, browser, security, diagnostics, and conformance accounting.
- [x] Every current/adjoining responsibility has a keep/move/split/defer decision.
- [x] All 236 `EXT-REQ-*` IDs, all 158 `EXT-AC-*` IDs, and all 28
  `EXT-GATE-*` IDs are present in traceability ledgers with workstream, owner,
  artifacts, dependencies, validation, and state.
- [x] BC-001 through BC-017 occur exactly once as target-rule definitions in the
  canonical ledger, preserve their selected decision and terminal adoption/
  implementation states, identify owners and slices, and map one-to-one to
  EXT-AC-142 through EXT-AC-158.
- [x] Every planned boundary criterion has a required owner-anchor, verification-ID,
  exact-selector, scenario-inventory, gate-prerequisite, and completion path through
  ES-01A through ES-05 and G-000.
- [x] Every Section 27 artifact/schema family has an authored owner, generated/runtime
  consumer, dependency, and validation posture.
- [x] Every workflow and slice has dependencies, validation, rollback, and an exact
  exit condition; behavior-changing slices say `requires later authorization`.
- [x] Public validation commands were discovered through Make; current owner-command
  failures are recorded rather than hidden.
- [x] No generated-file hand edit, phase-shaped runtime boundary, `EXT-FIX-*`, fixture
  result schema, v1 alias, compatibility reader, or historical evidence fallback is
  planned.

### Implementation and coordinated-adoption criteria

- [x] ES-00 characterization closes every route, envelope, claim, reserved-route,
  startup/fatal, migration, transaction, browser, and accounting boundary.
- [x] ES-01A incorporates every selected BC rule into the draft, amends existing
  requirements, creates EXT-AC-142 through EXT-AC-158, and leaves no conflicting,
  omitted, unbounded, or implementer-selected behavior.
- [x] All required Core, profile, shared-owner, OpenTelemetry, domain, and Harness
  companion amendments or exact permitted no-change parity records are adopted.
- [x] Every BC row has exact adopted owner requirement anchors plus active
  primary-owner verification IDs and exact selector families; G-000 is `DONE`.
- [x] Every owner locator/manifest/fragment resolves against exact current digests;
  no required placeholder or open delegation remains.
- [x] Every ES-06/07 Section 27 authored input exists, including dependency declarations,
  validation-surface declarations, traceability mappings, inactive-value schemas,
  and operation-specific portability results; every generated artifact/projection
  regenerates byte-identically with no drift and generated roots were not hand-edited.
- [x] Core 00 alone controls recognition, claimability, retirement, and current major;
  Core 01 alone owns public discovery/reservation/dispatch and shared protocols.
- [x] The Extensions coordinator exposes a small facade and owns only generic
  validation/coordination; profile, transport, storage, browser, reporting,
  portability, backup, and telemetry behavior stays with primary owners.
- [x] No route, workspace, job dequeue, WebSocket subscription, or readiness success is
  visible before atomic Stage 6 serving; admission exits `2` and fatal integrity exits
  `70` under the adopted Core 04 lifecycle.
- [x] Every state-owning profile has exact state presence, initialization, binding,
  codec, migration, ledger, lock, rebuild, validation, and backup/restore parity.
- [x] Every cross-owner transaction, job, staged-object, portability, reporting, and
  backup matrix passes exact failure/cancellation/boundary selectors with no partial
  effects or inactive profile-code execution.
- [x] Browser eligibility is the exact discovery/support/authorization/availability
  intersection; stale responses cannot render; Base cache/request/queue/optimistic
  state/drafts and stable client/WebSocket identity survive required transitions.
- [x] Security, egress, diagnostics, audit, telemetry, readiness, jobs, browser state,
  and WebSocket payloads contain no prohibited secret or incident content.
- [x] Both `module.extensions` verification IDs, every affected primary-owner
  verification, every active exact selector, runner, and execution profile validate
  with no unmapped, duplicate, unauthorized-skipped, or unexpected executed case.
- [x] Every affected owner has compatible successful full-owner evidence and every
  required evidence-class target has paired accounting/owner-summary shards.
- [x] Explicit-root Harness v2 evidence audits close every required owner/target/row
  partition; subset, broad-target, stale-root, newest-run, and historical fallback are
  rejected.
- [x] Static registry accounting, clause traceability, source lint, acceptance
  continuity, canonical limits, generated drift, owner evidence, documentation checks,
  all 158 acceptance criteria, and all `EXT-GATE-001` through `EXT-GATE-028` are
  `DONE`.
- [x] The mandatory ten-step sequence completed and the Extensions NLSpec plus every
  required companion revision was promoted atomically; no intermediate artifact ever
  claimed the generic subsystem current.
- [x] Core 05 remains outside implementation conformance unless a separately adopted,
  genuinely claim-bearing timed or fixture-sensitive publication boundary explicitly
  activates it.

Implementation and coordinated adoption are complete. The staged change set remains
uncommitted and unpushed pending separate authorization.
