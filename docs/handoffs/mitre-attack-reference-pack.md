# MITRE ATT&CK reference pack: live implementation handoff

Status: pre-adoption source preparation implemented; native runtime contract proposal awaiting review.
Research cutoff: 2026-11-06. Initial checkout: clean `main`, `ba4fdc06d6fb353c5f52ffe08e8a008d55311dd3`.
No commit, push, deployment, publisher sequence allocation, database reset, or deployment-root change was performed.

## Scope and authority

The user approved excluding every ambiguous lookup alias from every affected object, preserving original aliases and collision evidence, requiring native structured metadata/citations/URLs/lifecycle, and withholding an interim lossy v1 pack. Native contract adoption precedes assembly, signing, and positive runtime import. Relationship retrieval remains a separate owner gate. These instructions authorize the producer and proposal; they do not adopt new runtime identifiers.

The supplied `temp/cartulary-mitre-attack-reference-pack-findings.md` remains unchanged. Its snapshot observations were rechecked against the live checkout. This handoff is supporting evidence, not normative adoption or executable product semantics. Tests and producer commands consume JSON and retained source bytes, never this file or other Markdown.

### Live owner reconciliation

| Finding | Current owner and implementation | Disposition |
| --- | --- | --- |
| Closed v1 framework rows cannot preserve native ATT&CK fields | `docs/reference-pack-subsystem-nlspec.md:1101` RP-REQ-108–111; profile registry at line 1026; schema IDs at line 1373; `contracts/reference-packs/profiles/framework.attack.*.v1.schema.json` | Runtime schemas unchanged. Producer-only candidates propose new identifiers. |
| Canonical-byte/consumer changes need coordinated versioning | Reference Pack RP-REQ-018/019/020 at lines 176–183; `docs/spec/00_document_set_status_and_precedence.md:30` | Propose pack contract v2, ATT&CK profile v2/string `"2"`, object/relationship v2, subsystem major 3. None admitted yet. |
| Four stale format-5 references | Reference Pack lines 15, 82, 197, 2297 versus Core 01 REQ-01-635 at `docs/spec/01_architecture_storage_and_view_contracts.md:9047` | Corrected all four to format 7. This is editorial synchronization with an already adopted owner, without changing v1 pack identities, document major, or implementation. |
| Snapshot/live Core 00 difference | `docs/spec/00_document_set_status_and_precedence.md:116` names bundle v7 | The live file has no “Remediation adoption — 2026-11-06” section described by the snapshot. Core 01 explicitly admits only v7 and rejects 1–6; that live rule supplies authority. |
| No relationship/adjacency consumer operation | Reference Pack §21 at line 1983, RP-REQ-204; `internal/modules/reference_data/consumer.go`; `contracts/reference-packs/consumer_operations.v1.json` | Existing five operations remain closed. No direct consumer storage reads or invented relationship API. Relationship-pivot E2E remains blocked separately. |
| Clock evidence restricted to Network Flow | `docs/testing-harness-nlspec.md:10414` TH-HARNESS-REQ-659 | No Reference Data frozen-clock control used. Producer `BUILT_AT` is reproducible evidence input, not verification time. Future imports require fresh envelopes and trusted wall-clock admission. |
| Immutable source profile and exact builder bytes | Reference Pack lines 468, 1557–1578, RP-REQ-150/151 | Exact executable retained by SHA-256; candidate index binds exact proposal files. Final runtime source-profile freezing remains blocked. |
| Reference-only scope | `docs/domain.md:421`, `:478`, `:935`; Core 03 at `docs/spec/03_workbook_interaction_collaboration_and_workflows.md:97`; RP-REQ-137 | No tabs, views, incident records, ACLs, activation, or automatic incident mutations created. |
| Current trust/security and persistence owners | Core 02 Reference Data persistence section; Core 04 lines 245, 452, 1199, 1457; Extensions lines 28, 610, 621 | No successful verification rows fabricated, trust thresholds weakened, roots replaced, or packaged Base identities changed. |
| Reporting/Graph boundaries | Reporting §5 and snapshot materialization/retained-pack rules; Report Composition 1.3.0 line 14; Core 00 Graph owner boundary lines 65–69; `docs/graph_projection_nlspec.md` | Existing snapshot and exact-result boundaries remain authoritative. Source graph audit is not consumer Graph integration. |

The runtime verifier to reuse after adoption is `internal/modules/reference_data/verification_engine.go:84` (`verifyCanonicalContainer`), with `internal/modules/reference_data/internal/packformat/` for canonical content, identity, digest, inventory, archive, and offline TUF checks. Producer source preparation currently reuses `internal/platform/canonicaljson`; it does not invoke a replacement import verifier. There is no canonical pack to pass to the production verifier yet.

## Pinned upstream inputs

Official [v20.1 release](https://github.com/mitre-attack/attack-stix-data/releases/tag/v20.1), published **2026-11-05T16:35:43Z**, resolved commit **f59b03ecf828579d5f9e93683ce25b7e2e5fd9d7**. Annotated tag object: `fe094cfc533102f8eed3acefd89bb38a96fdf2c1`. The tag is unsigned; no claim of MITRE signing a Cartulary container is made. The [version history](https://attack.mitre.org/resources/versions/) associates v20.1 with the underlying major release date **2026-10-13**; that date does not replace the exact release publication timestamp.

| Input | Exact bytes | Independently checked SHA-256 |
| --- | ---: | --- |
| Enterprise `enterprise-attack.json` | 55,277,605 | `539a843a7b53ed9749740b2123dddae3726b1a4831841c32aca740b011b910b5` |
| Mobile `mobile-attack.json` | 6,144,331 | `632e2ec40848394f9873a7a3485fc32f92b861a4e7271f7e88e0cb02f11bb6af` |
| ICS `ics-attack.json` | 5,291,092 | `3f6448369f9135ade84ff6437479441530d0b262f910dca991879a7146b524ff` |
| Selected-commit `LICENSE.txt` | 1,354 | `738144f7fb054722a4ef9d3367c51710341dc12fc574c6ac3a41daaaecd8bf5e` |

The source lock records exact official URLs, release/tag API locators, commit, and the three collection identities. Acquisition checks the live tag reference, annotated object, release timestamp, asset names/sizes/published digests, and independently hashes downloaded bytes. A mismatch aborts admission; no fallback version or moving branch is used. Existing successful acquisition replays by offline checking retained bytes; it does not silently refresh upstream metadata.

Original inputs are retained under `.cartulary/reference-packs/attack/20.1/sources/<sha256>/<name>`. `acquisition.json` binds the retained source lock and exact release/tag/ref evidence. Partial acquisitions retain only exact immutable evidence; they do not publish a successful receipt. Differing existing bytes, path traversal, and symlinked evidence are rejected.

## Measured all-domain audit

All 31,975 input records were checked against the producer-only native source schema. The audit reconciles **31,796 unique records**, **155 shared IDs**, no unresolved same-ID content differences, **426 type/property paths**, and **54 colliding normalized aliases**. Shared IDs can occur in more than two domains; 155 is not a count of redundant records. These are source counts, not canonical pack counts.

| Unique category | Count | Proposed disposition |
| --- | ---: | --- |
| Techniques/sub-techniques (`attack-pattern`) | 1,184 | Native object and exact embedded tactic membership |
| Tactics | 41 | Native object, short name and domains |
| Matrices | 4 | Native object, ordered tactic references |
| Groups | 197 | Native object, original and unambiguous lookup aliases separated |
| Malware | 886 | Native object |
| Tools | 97 | Native object |
| Campaigns | 62 | Native object, first/last seen and citations |
| Mitigations | 335 | Native object |
| Detection strategies | 939 | Native object, direct analytic references |
| Analytics | 2,288 | Native object, mutable elements and component/log-source structures |
| Data components | 123 | Native object, ordered log sources |
| Historical data sources | 42 | Native object, independent lifecycle flags |
| Assets | 18 | Native object, sectors and structured related assets |
| Collections | 3 | Native provenance object including versioned contents |
| Identity | 1 | Native provenance object |
| Marking definition | 1 | Native provenance object |
| Direct STIX relationships | 25,575 | Native SRO metadata in canonical relationship proposal; original direction, identity, procedure and citations |

| Domain | Input records | Direct SROs | Active sub-techniques | Maximum description scalars |
| --- | ---: | ---: | ---: | ---: |
| Enterprise | 26,699 | 21,822 | 479 | 4,668 |
| Mobile | 2,706 | 1,954 | 47 | 3,390 |
| ICS | 2,570 | 1,832 | 26 | 2,720 |

No description exceeds 8,192 scalars. No encountered SRO has either retirement flag true. SRO lifecycle is nevertheless evaluated independently of endpoints. Replacement cycles, missing endpoints, conflicting replacements, chains over 32, and collection-version disagreement fail auditing. All direct source references resolve; 1,464 named tactic memberships resolve uniquely by domain and short name.

The source contains 2,284 strategy/analytic references, 5,268 analytic/component references, 3,838 component log-source entries, 41 ordered matrix tactic references, and 43 related-asset entries. Provenance references (creator, modifier, markings and collection inventory) are separately identified in the audit. Related assets named without STIX IDs remain structured data; the proposal does not invent graph identities.

Measured maxima: 16 aliases per object, 49 external references per record, 395 log sources on one component, 26,697 collection content entries, and 3,033,999 bytes for the largest JCS source record. The current NDJSON line ceiling is 1,048,576 bytes including LF (Reference Pack lines 1055, 2545). A native collection row therefore needs an explicitly adopted larger bound or a different owner-approved representation. The concrete candidate proposes 4,194,304 line bytes including LF, 32,768 collection entries, 512 component log-source entries, and the existing 64 bound for ordinary arrays. These are proposed bounds, not runtime changes. Full augmented rows and containers remain unmeasured until assembly.

There are 37,608 citations without external IDs, 2,684 URLs containing fragments, and 1,241 non-HTTPS URLs. The recursive audit found no non-NFC strings and no prohibited control strings, but 4,197 string occurrences with outer whitespace. Original native values remain exact; canonical display fields require explicit normalization without truncation.

### Fidelity decisions and evidence

- All 54 colliding NFC/ASCII-folded alias keys are excluded from every proposed lookup alias array, including collisions involving retired records. Original `aliases` and `x_mitre_aliases` remain separate source fields. StrongPity identifies both G0056 PROMETHIUM and S0491 StrongPity; Corona Updates and Concipit1248 also collide. No alias winner is selected.
- AN1789 retains `https://attack.mitre.org/detectionstrategies/DET0682#AN1789` in native `source.external_references`. Its real canonical external ID can have `url=null`. Other scheme/fragment URLs and citations without IDs are preserved natively, never assigned invented IDs or rewritten URLs. URL strings are inert data, not a fetch instruction.
- Separate optional `revoked` and `x_mitre_deprecated` flags preserve absence, false and true. The proposal defines generic `deprecated` as their OR alongside the originals; it introduces no current-only filter. Retired endpoints and valid direct replacement assertions remain available.
- Direct SROs keep their direction, procedure description, citations and `source.id`. Proposed embedded predicates are `has-tactic`, `member-of-tactic`, `has-analytic`, and `uses-data-component`, each with exact source-item JSON pointers. Transitive analytical paths are not MITRE assertions.
- The [official relationship guide](https://mitre-attack.github.io/attack-data-model/schemas/relationship-types/) still says sub-techniques are Enterprise/Mobile only. Selected ICS bytes contain 26 active sub-techniques and 26 `subtechnique-of` SROs. The producer retains all of them and records the documentation discrepancy.
- Every encountered source record category and recursive property has an explicit proposed native disposition. Bundle framing/IDs and GitHub metadata remain source-only evidence. The exact license becomes a declared notice after approval. There are no excluded source object categories. This is a native-preserving projection proposal, not a byte-lossless STIX archive or complete native ATT&CK application.

## Concrete adoption proposal

Producer-only files live under `contracts/reference-pack-producers/attack/`, outside `contracts/index.json` runtime projections. `candidate-index.v1.json` binds exact SHA-256 and length for the proposal, field policy, schemas, vocabulary and source-validation fixtures. The native schema is closed per source type, including nested citations, log sources, mutable elements, assets, collections and provenance. Object/relationship wrapper candidates retain current common members and add the typed native `source`; embedded relationship source is explicitly null. Current `.v1` schemas and runtime identifiers are untouched.

Recommended adoption: `cartulary.reference_pack_contract.v2`; `cartulary.reference_pack.framework.attack.v2`, profile version string `"2"`; object/relationship `.v2`; subsystem contract major 3. Preserve `framework.attack`, `framework`, `framework_reference`, application contract majors `[1]`, and empty required capabilities. Proposal-only schema identifiers are deliberately different and must never be admitted as runtime identifiers.

The final source profile is not frozen. The candidate defines accepted inputs, dates, digest subjects, all common projections, exact alias policy, lifecycle mapping, source coalescing, direct embedded-reference predicates, source locators, ordering, license/notices, limits and fixtures. It proposes retaining the existing RP-REQ-111 relationship hash algorithm and RP-REQ-102 row ordering. Runtime source artifacts are the three bundle descriptors, sorted by `source_ref`, `source_version`, `sha256`; license bytes are retained separately and bound as a notice. A behavior-changing revision requires a new source-profile ID once adopted; retained candidate revisions and prior audits must not be overwritten.

The owner review must resolve these exact effects before runtime work:

1. Reference Pack/Core 00/Extensions: adopt the new version tuple, typed native fields, line/array limits, generic OR lifecycle projection, unambiguous aliases and embedded predicates. Register machine inputs and regenerate only after adoption. This blocks canonical assembly, fixture registration and positive import.
2. Reference Pack/Core 01/Core 02/Core 04: explicitly bind the successor contract to immutable sets, consumer item unions, persistence, portability v7 and recovery. Decide the successor Base registry release identities and whether the approved registry permits v1 members alongside ATT&CK v2. Do not infer mixed-version support or add a legacy reader. Historical sets and first-success evidence must remain exact.
3. Reference Pack §21/Graph/Reporting: define a relationship/adjacency operation if relationship pivots are wanted. Options are a separately adopted paged operation with exact-set provenance, or leaving pivots unsupported. Recommend a separate proposal; neither source validation nor direct storage reads closes this gate.
4. Security and licensing authorities: provide the actual RP-GATE-010/011 dispositions, publisher trust repository ID, recorded release sequence/version, signing credentials and required notices. The selected-commit license is retained; [MITRE terms](https://attack.mitre.org/resources/legal-and-branding/terms-of-use/) are not project approval.

## Operator commands and retention

Run from the repository root with the repository-managed toolchain:

```sh
make reference-pack-producer-build
make reference-pack-source ACTION=acquire
make reference-pack-source ACTION=check
make reference-pack-source ACTION=audit BUILT_AT=2026-11-06T19:00:00Z
```

Build runs twice with the pinned Go toolchain, `CGO_ENABLED=0`, `-mod=readonly`, `-trimpath`, and `-buildvcs=false`, compares executable bytes, and retains executable, self-hashed identity, and reproduction witness under `.cartulary/reference-packs/builders/<sha256>/`. `build/bin/reference-pack-producer` is the replaceable local composition entrypoint. Rebuild it after producer source changes. No private signing material or deployment credentials are loaded.

Only `ACTION=acquire` supports network access. It accepts only the locked official HTTPS locations and approved GitHub asset redirects. `check` and `audit` read retained evidence and use an offline schema loader. An existing acquisition never refreshes itself from network. To independently reacquire later without replacing old evidence, choose a fresh `SOURCE_DIR` and retain both roots. Source/schema/hash failures stop that operation; prior evidence is preserved.

`ACTION=audit` requires explicit `BUILT_AT` at or after release publication. It retains exact candidate files under `candidates/<candidate-index-sha256>/` and a deterministic JCS audit under `audits/<sha256>.json`. Running twice with the same retained source, candidate, builder and timestamp must produce the same audit bytes and path. The report includes original-source locators, field distributions, complete collisions, proposed lookup aliases, all embedded references, source-version values, and candidate-schema validation. These evidence files are intentionally outside any runtime pack.

There is no pack assembly, import, signing or renewal command in this producer. After adoption, assembly must use existing packformat primitives and the production verifier; then operator import uses the existing deployment-worker route and default staged-only admission. Future isolated fixture trust must be explicit, keep keys outside container/bootstrap, preserve required TUF thresholds, and use fresh envelopes with wall-clock admission. Envelope renewal must preserve canonical manifest/payload bytes and prior envelopes. Intended deployment trust is a separate package identity when `trust_repository_id` differs. No current source audit is a trust attestation.

## Tests and next workflow integration

Source-validation fixtures are explicitly synthetic parser/schema vectors, not a positive ATT&CK runtime pack. Every case has an expectation runner in `producer/source_test.go` or `producer/candidate_test.go`, bound through the existing `module.reference_data` unit family. Cases include valid source coalescing, alias exclusion, immutable replay/no fetch, malformed JSON, hash mismatch, moved tag, wrong release, unknown properties/types, conflicting IDs, missing endpoints, replacement cycles, typed native fields, omitted members, and closed extensions. Retention tests cover tampering, collisions, traversal and symlinks. Source counts/hashes are also checked against independently supplied release witnesses; no Markdown is an executable input.

The new row is `module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c`. `contracts/reference-pack-producers/attack/inputs.go` embeds only JSON machine inputs for tests, so the existing Go transitive-import dependency closure binds every producer contract and fixture byte into both harness and Go cache identities. Runtime assembly and the source-producer executable do not import this test-input package. No fixture-root schema or harness dependency algorithm was widened. Existing canonical fixture closure is checked with `module.reference_data.unit.canonical_fixture_manifests`. No new harness layer or runtime fixture exception was added.

After adoption, register the actual complete signed pack plus separate malformed and semantic-error derivatives through `contracts/reference-pack-fixtures/`, with real owner bindings. Preserve the existing small synthetic profile fixture as unit evidence only. Exercise production import, verification, staged-only default, current admin authorization, activation, object/external-ID/unambiguous-alias lookup, provenance, replay/collision, no egress, failure cleanup, immutable sets, renewal, disablement and removal. Do not insert verified database rows.

Pin a second authentic ATT&CK source release and all three Base registry identities before activation-transition/set-hash assertions. Extend the existing Reporting `TestSnapshotAdmissionAndRerenderRetainSetAcrossActivation_Integration` binding in `owner_snapshot_binding.v1.json`, incident-bundle v7 `TestIncidentReferenceCatalogSurvivesUnavailablePackRoundTrip_Integration` binding in `owner_portable_destination.v1.json`, and affected Recovery owner scenarios. These were inspected, not executed as ATT&CK evidence. Relationship pivots remain outside them until the separate API gate closes.

## Retained execution evidence

All paths below are repository-relative. Final implementation source digest in the harness: `sha256:6ff671be49a5415eeefeebfac738b70fe7536df733609288b8c9f4f4e36d5020`. Documentation-only evidence additions do not supply executable semantics.

| Artifact | Bytes | SHA-256 |
| --- | ---: | --- |
| Source lock | 2,143 | `171183b6b4c7e2f8fb0b4e2c08cc0f36d411602c13746fe755f052110c3ddeaa` |
| Candidate source profile (not final) | 9,966 | `9cc1bd346824b2ebb7457e49aecf3cfa1e11b39876a160e2ba32277c13191e40` |
| Candidate index | 1,868 | `6773005fc8129017a6102a6c00aca13c6675b495cee7a742b86973a6605d9998` |
| Field policy | 35,375 | `d8515b522ff84ea1f71677fed4b82389cfe41d5c7c4db323a8334fefc6f1920e` |
| Native source schema | 72,250 | `461c5f13bf9760fb86f4559ffa88030865f8f9705e9fb6be4ea51aecdc88ebd8` |
| Acquisition inventory | 1,559 | `2749ab561df2c674fe04d1b5181beb5a28bc7b9de7a5dc1fa65106af806f6901` |
| Reproduced source audit | 53,141,779 | `7f3b9f539cb70d43d4b20f3c8a71eccfe0eca20f492fb41172bc67f78bfcd7b4` |
| Exact producer executable | 11,568,767 | `e91269af0e7c68a9d125b3b662fb6855850b932a1a93ecc04e82ffb44477b612` |

Final audit: `.cartulary/reference-packs/attack/20.1/audits/7f3b9f539cb70d43d4b20f3c8a71eccfe0eca20f492fb41172bc67f78bfcd7b4.json`. The native proposal shape check validated **31,975** input records with no issues. `attack-source-audit-20261106-09` and `-10` produced identical report bytes: the second run performed immutable no-replace retention and exact byte comparison. Built-at input was `2026-11-06T19:00:00Z`. Builder `candidate-1` used Go `go1.27.3`; `attack-builder-20261106-06` independently built twice and compared bytes. Reproduction is established in this pinned checkout/toolchain environment, not as a cross-platform build claim.

| Executed command | Final retained run under `.cartulary/test-results/` | Outcome |
| --- | --- | --- |
| `make task-guide OWNER=module.reference_data` | selection only | Selected Reference Data owner slice |
| `make generate` | `attack-generate-20261106-06` | Pass; authored owners rendered |
| `make reference-pack-producer-build` | `attack-builder-20261106-06` | Pass; two byte-identical executables |
| `make reference-pack-source ACTION=acquire` | `attack-source-acquire-20261106-02` | Pass; official metadata and all four byte hashes |
| `make reference-pack-source ACTION=check` | `attack-source-check-20261106-01` | Pass; retained-byte verification, no network |
| `make reference-pack-source ACTION=audit BUILT_AT=2026-11-06T19:00:00Z` | `attack-source-audit-20261106-09`, `attack-source-audit-20261106-10` | Both pass; identical JCS audit |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.canonical_fixture_manifests` | `attack-unit-20261106-06` | Pass; two selected work units, zero failed/skipped |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c CARTULARY_HARNESS_CACHE_MODE=off` | `attack-unit-20261106-07` | Pass; final producer including early-failure verdict regressions, fresh execution |
| `make format` | `attack-format-20261106-05` | Pass |
| `make agent-finalize` | `attack-finalize-20261106-04` | Pass; required finalization before final checks |
| `make source-check CHECK=generated` | `attack-generated-check-20261106-03` | Pass |
| `make source-check CHECK=schema` | `attack-schema-check-20261106-03` | Pass |

`make lint-markdown` passed in `attack-markdown-20261106-02` before this final evidence annotation. A final lint invocation is recorded separately by Make.

A separate fresh execution of both selected rows with `CARTULARY_HARNESS_CACHE_MODE=off` passed in `attack-unit-20261106-05`. After the final candidate citation/index update, `attack-unit-20261106-06` executed the producer row again (exit 0) and legitimately reused the unchanged canonical-fixture row. The current proposal fixtures include an actual retained acquisition-receipt shape witness alongside explicitly synthetic source/schema vectors.

Final review also corrected early parse/schema-compilation failure reports to retain failed verdicts; the source and candidate tests explicitly reject false passing labels. `attack-unit-20261106-07` is the fresh final producer evidence after this correction.

Every invocation used the explicit run ID shown. The harness retained source/toolchain/system/graph identities and terminal summaries; all final command exits were zero. Full-owner/CI, service/browser, Reporting, incident-bundle, Recovery and runtime trust tests were not run: no native runtime pack or adopted consumer tuple exists, and this change does not alter those implementations. No benchmark was run. Retained warm-check maintenance was not requested because no compatible full warm-check result was supplied.

Earlier repaired harness setup failures remain available:

- `20261106T193506Z-p89830`: generation rejected missing observability exclusions and an empty help usage field for the new helper targets; corrected authored task-surface input.
- `attack-source-acquire-20261106-01`: recursive public builder prerequisite inherited undeclared `ACTION`; replaced with explicit build prerequisite instructions so offline source operations do not trigger builds.
- `attack-format-20261106-01` and `attack-generate-20261106-03`: newly authored row test selectors were not ASCII-sorted. `make target-plan TARGET=format JSON=1` identified the same catalog error; corrected the row and regenerated.

- `attack-unit-20261106-04` reused an earlier unit result before the producer JSON directory was in its dependency closure; it is not evidence for the later acquisition-schema fixtures. The attempted `fixture_input_roots` binding in `attack-generate-20261106-05` was correctly rejected because TH §4 restricts that field to `fixtures/`. The final implementation uses the ordinary test-only Go import/embedded-input closure, preserving that restriction. Fresh `attack-unit-20261106-05` and the correctly invalidated producer execution in `-06` replace this incomplete evidence.

These were task-introduced harness configuration/dependency defects, not upstream hash mismatches or passing product evidence. Their successors passed; no failed run was overwritten.


## Completion boundary

| Claim | Status |
| --- | --- |
| Source acquisition/audit and builder reproducibly buildable | Implemented; exact retained witnesses below |
| Final native source profile frozen | Blocked on coordinated adoption |
| Canonically conformant runtime pack | Not assembled; object/relationship counts, NDJSON sizes, manifest/payload/container hashes are unavailable |
| Verified under isolated fixture trust | Not performed; no signed envelope |
| Importable under intended deployment trust | Not established; intended identity, sequence and credentials missing |
| Approved for distribution | Not established; actual security/licensing dispositions missing |
| Complete ATT&CK E2E workflow | Blocked on native adoption, real pack and authentic second release |
| Relationship-pivot E2E | Separately blocked on consumer contract |

The reviewable next action is native-contract adoption with the coordinated owner and Base-registry decisions above. Passing source/schema tests do not substitute for that decision, signed import verification, distribution approval, or complete Reference Pack conformance.

## Exact changed-file inventory


### Authored producer contracts

- `contracts/reference-pack-producers/attack/acquisition.v1.schema.json`
- `contracts/reference-pack-producers/attack/candidate-index.v1.json`
- `contracts/reference-pack-producers/attack/dispositions.v1.json`
- `contracts/reference-pack-producers/attack/field-policy.candidate.v1.json`
- `contracts/reference-pack-producers/attack/fixtures/cases.v1.json`
- `contracts/reference-pack-producers/attack/fixtures/inputs.v1.json`
- `contracts/reference-pack-producers/attack/fixtures/native-cases.v1.json`
- `contracts/reference-pack-producers/attack/inputs.go`
- `contracts/reference-pack-producers/attack/native-source.candidate.v1.schema.json`
- `contracts/reference-pack-producers/attack/object.candidate.v1.schema.json`
- `contracts/reference-pack-producers/attack/relationship.candidate.v1.schema.json`
- `contracts/reference-pack-producers/attack/source-lock.v20.1.json`
- `contracts/reference-pack-producers/attack/source-profile.candidate.v1.json`

### Producer and composition implementation

- `internal/modules/reference_data/producer/audit.go`
- `internal/modules/reference_data/producer/candidate.go`
- `internal/modules/reference_data/producer/candidate_test.go`
- `internal/modules/reference_data/producer/identity.go`
- `internal/modules/reference_data/producer/source.go`
- `internal/modules/reference_data/producer/source_test.go`
- `tools/reference-packs/build.sh`
- `tools/reference-packs/main.go`
- `tools/reference-packs/source.sh`

### Authored routing and documentation

- `docs/handoffs/mitre-attack-reference-pack.md`
- `docs/reference-pack-subsystem-nlspec.md`
- `tools/harness_work_graph_owner.json`
- `tools/task_surface_owner.json`
- `tools/test_families/module.reference_data.json`

### Generated through Make

- `internal/testutil/harnessruntime/policy_generated.go`
- `tools/execution_topology_render_index.json`
- `tools/task_surface.generated.mk`
- `tools/task_surface.help-all.generated.json`
- `tools/task_surface.runtime.generated.mk`
- `tools/task_surface_manifest.json`

Retained build/source/audit artifacts under `.cartulary/reference-packs/` and Make evidence under `.cartulary/test-results/` are local ignored artifacts, not distributable runtime pack content. The existing research handoff was preserved.


## Native v2 implementation outcome — 2026-11-06

This appendix supersedes the earlier completion boundary; the preceding research, acquisition, candidate, failed-run, and executable history remains intact. The authorized cutover is implemented. The native v20.1 and independently qualified v20.0 packs contain Enterprise, Mobile, and ICS together. No deployment publication, deployment sequence allocation, root replacement, database reset, commit, or push was performed.

### Adopted behavior and ownership

Reference Pack subsystem major 3 now binds pack contract v2 and ATT&CK profile/object/relationship v2, with profile version `"2"`. The current profile registry supplies explicit schema and line-limit bindings. Trusted compiled owner artifacts resolve local definitions and references; cycles, missing definitions, traversal, unsupported constructs, and external resolution fail. The inclusive ATT&CK row ceiling is 4,194,304 bytes, collection inventory limit 32,768, component log-source limit 512, and ordinary array limit 64. Other profiles retain their identifiers and limits.

The five consumer operations retain their request contracts. New pack-set, portable-reference, consumer-item, and affected result schemas carry native v2. Manifest, TUF, provenance, indicator-result, Jobs, and configuration identities remain unchanged. Incident-bundle stays at outer format 7 and its source catalog explicitly binds `reference_pack_refs.v2`. Core, Extensions, Reporting, Composition, and Recovery bindings were coordinated. Frozen internal operation input advances to v3 and portability input to v2 so pending predecessor work can be rejected before execution; their predecessor schemas remain historical evidence.

Reference Data owns read-only admission of retained manifests, immutable sets, portable catalogs/preparations, prepared attempts, and pending work. Reporting owns inspection of its retained admissions, render requests, source captures, render bundles, delegated Composition previews, and chunked imported control artifacts. Application assembly composes their checks before migration/startup writes and readiness. Worker admission and Base reconciliation guard incompatible Reference Data state. Public binding callbacks let consumers check retained set/provenance identities without storage access. Recovery adopts generation `recovery.reference_pack_native.v3` and rejects the retired catalog before target mutation. No legacy reader, conversion, or reset is provided.

The producer preserves every audited native record and property, optional presence, array order, citation, original URL, structured detection field, and independent lifecycle flag. Common displays alone use the existing trimming, NFC, and LF normalization. Absent or empty descriptions become null. Ambiguous normalized aliases are removed from all lookup arrays; original aliases and complete collision evidence remain native/in the retained audit. Generic retirement is the OR of the original flags. Direct SRO lifecycle remains independent, with original identity, direction, procedure text, and citations. Embedded edges preserve every declared property occurrence and coalesced locator for exactly the four predicates; no transitive assertions or consumer adjacency API was added.

`make reference-pack-package ACTION=assemble|envelope|renew INPUT=<typed-json>` now uses shared production canonicalization, validation, framing, and deterministic ustar primitives. Its closed inputs bind retained sources, frozen source profile, executable digest, publication identity, build timestamp, append-only ledger, signing context, metadata version, and expiries. Source acquisition alone may use the network. The public target is registered through authored task-surface and work-graph owners, with generated projections refreshed through Make. The separate Node witness independently recomputes JCS, SHA-256 framing, native preservation, archive structure, metadata links, and Ed25519 signatures without importing the Go producer.

### Frozen inputs and reproducibility

The primary baseline remains v20.1, published `2026-11-05T16:35:43Z`, commit `f59b03ecf828579d5f9e93683ce25b7e2e5fd9d7`. The transition baseline is v20.0, published `2026-11-05T16:13:00Z`, commit `5bbae6aa5bf4197aa7abd44b4864d8043d4fed98`; the annotated tag object is `db3ec60461f42b3edf437a1a5d1cf77b202d655e`. Separate retained roots are `.cartulary/reference-packs/attack/20.0` and `/20.1`. Official metadata, all bundle bytes, collection identities, and selected-commit license bytes were checked before freezing the new profile. Existing candidates and the prematurely frozen native v1 profile remain historical inputs.

| Input | SHA-256 |
| --- | --- |
| v20.0 exact source lock | `ba44e42a1c1d4d000b69055bba8214b264e83130baa9aa6292374cbeb78202fc` |
| v20.1 exact source lock | `171183b6b4c7e2f8fb0b4e2c08cc0f36d411602c13746fe755f052110c3ddeaa` |
| `cartulary.attack.all-domains.native.v2` source profile | `80d546d402318c2866b999dc2c80ae8e1c663f002c9187b71dfc025f6f1484db` |
| Native field policy | `32b69aa7463c302f8b1be56d857aff3527638517054850654801a371fd4fceeb` |
| Exact final executable | `5bbc51fcc92494ab0f6e8406f38c422eff2583815f12aae442c541cc42bd7f9a` |
| Final v20.0 audit | `d089608c20911262d0093962a8c7480b43e00eb14d62e0fb3af7e8f1a7fc0a2b` |
| Final v20.1 audit | `6b6f9e4c423ad7333fabfe5e7ceacd79a55f9d29b6e9e206dd4bd6f58062f279` |

| v20.0 retained source | Bytes | SHA-256 | Collection identity |
| --- | ---: | --- | --- |
| `LICENSE.txt` | 1,354 | `738144f7fb054722a4ef9d3367c51710341dc12fc574c6ac3a41daaaecd8bf5e` | License |
| `enterprise-attack.json` | 55,271,014 | `2339a3c36353d5ee2c07e98af3fafbf00761e0e3979ede06402657445b7b4278` | `x-mitre-collection--1f5f1533-f617-4ca8-9ab4-6a02367fa019` |
| `ics-attack.json` | 5,266,683 | `7bffdff9853f04117a607bb5de652140a5724a3a06afc88bdc3faee37103d622` | `x-mitre-collection--90c00720-636b-4485-b342-8751d232bf09` |
| `mobile-attack.json` | 6,095,145 | `c7d6c923c1f8462ab1c5db0195d81e2585db9949d574128391c682f229c6695a` | `x-mitre-collection--dac0d2d7-8653-445c-9bff-82f934c1e858` |

The executable is 13,098,315 bytes, builder version `native-v2`, built with pinned `go1.27.3`. Exact bytes, identity, and reproduction report are retained under `.cartulary/reference-packs/builders/5bbc51fcc92494ab0f6e8406f38c422eff2583815f12aae442c541cc42bd7f9a/`. Base reissuance preceded this freeze. Build invocations `attack-v2-builder-20261106-07` and `-08` each built twice and compared bytes; both produced that executable. This establishes reproducibility in this checkout and pinned toolchain, not across arbitrary platforms.

Assembly uses explicit `built_at=2026-11-06T19:00:00Z`. The two final assembly invocations for each release replay identical explicit inputs and compare every retained logical member. The two envelope invocations for each release also replay identical timestamps, metadata versions, and signing context, reproducing the same archive bytes. Renewal is compared separately: it changes envelope metadata/container identity while preserving manifest, payload, pack version, publication sequence, and prior envelopes.

### Exact canonical fixture identities

The current immutable fixture catalog is `fixtures/reference-packs/attack-native-v2/catalog.json`. Each release root is `<version>/logical/<manifest-sha256>/` beneath that directory. It contains the JCS manifest, LF-terminated NDJSON, exact notice, build report, independent witness, and exact source/profile/policy/publication-input evidence. Previous `.2` logical outputs and catalog snapshots remain retained; final `.3` identities bind the final executable.

| Canonical quantity | v20.0 | v20.1 |
| --- | ---: | ---: |
| Objects | 6,211 | 6,221 |
| Direct SROs | 25,554 | 25,575 |
| Embedded relationships | 9,050 | 9,057 |
| Total relationships | 34,604 | 34,632 |
| Native records preserved | 31,765 | 31,796 |
| Manifest JCS bytes | 2,569 | 2,569 |
| Objects NDJSON bytes | 23,037,177 | 23,061,726 |
| Relationships NDJSON bytes | 51,947,988 | 51,986,185 |
| Maximum object row bytes, including LF | 3,033,785 | 3,034,554 |
| Maximum relationship row bytes, including LF | 10,697 | 10,697 |
| MITRE notice bytes | 1,354 | 1,354 |

| Digest | v20.0 | v20.1 |
| --- | --- | --- |
| Manifest | `c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90` | `45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f` |
| Framed payload | `af29e228df3453600b4388969c12c5e4be5f1e5c4f6a45c798207f25cb8e936a` | `c1f6ddf1a73e2544bb2b568ac0135a355f975087c6dac37291ebbad15aa18e4e` |
| Set with required Base members | `910f7a489f461fd42da75cff66e312771f44b4db842fc9c1bb3cf9949d424b66` | `416b935f9215eaa32f0d199c0a1213a8906e66b2601ae4cf30d4c1d11f25311c` |
| Objects member | `78bfdda846326020b0e9e20c5b240d693596ccdb235f7f1f4b15ae28eaa39d07` | `22893a502098db91e8aa33618cb48b39ed3c16345651066a9ce065166674bad9` |
| Relationships member | `7bc3ccbfe63e9425c0f35c0cb05f05d45d13dfaa424f4867dcd8842a10b3e483` | `ee8a5ecacf9094d087ddcc20cabb3ac7196ad04ac41497e58e477bf4ea9b6274` |
| MITRE notice | `738144f7fb054722a4ef9d3367c51710341dc12fc574c6ac3a41daaaecd8bf5e` | `738144f7fb054722a4ef9d3367c51710341dc12fc574c6ac3a41daaaecd8bf5e` |

The independent logical witnesses cover all source categories and exact parsed native values, including omissions, ordered arrays, and original noncanonical URLs. The production validator additionally checks common/native consistency, collisions, endpoint and provenance integrity, embedded locators, replacement chains, and limits. Separate malformed and semantic derivatives are registered through the existing fixture manifests and owner runners; their bytes and the complete positive fixtures are bound through `fixtures/` dependency roots. Validation does not rewrite those canonical fixtures.

### Required Base reissuance

All three required Base manifests use release `cartulary.base_registries.2026-11-06.v2`, pack version `base-2026-11-06.1`, sequence 2, and pack contract v2. Their registry payloads, notices, content-profile IDs, and profile version `"1"` remain unchanged. `contracts/reference-packs/builtins/witness.v2.json` freezes complete tuples independently; the Base set SHA-256 is `2a3420dcc6036316fbb68d5c2a1d42e28cbd1003fe5207a93912ca45e9b993df`, and release bytes SHA-256 is `9eeb17c67436155136af26e60a5865ad55607dc3197481c980f5b7b5dcb657f3`.

| Pack key | Manifest SHA-256 | Payload SHA-256 | Content profile |
| --- | --- | --- | --- |
| `type_registry.evidence` | `d446139e2a9f83ff4baf5597a9ec97723a440dc3d138f135836164aca211dd12` | `9a24ec390bf4e9f51ea299f6e08c53f7f8186a66ed2a7decb213a2ad9e901d39` | `cartulary.reference_pack.type_registry.evidence.v1` |
| `type_registry.host` | `f0c19e9d426077f3ddc16eeba20680a41b80509083d9c59bc9c56305ea7bbca5` | `afaae18a960bf0d7f75365494966403333706ee4b7f680ca789c5cad8a237b8f` | `cartulary.reference_pack.type_registry.host.v1` |
| `type_registry.indicator` | `a121e107724bd4f50734f075c997c727fe4b1cf2dd379f06b387f436f8f7c74f` | `1c559a7fd18803b4eb0fbcc53f39be43ff48404d7253dfa0ffd5c62cdb3fecb8` | `cartulary.reference_pack.type_registry.indicator.v1` |

### Fixture publication and public trust

Repository identity is `fixture.attack.native.v2`; redistribution is `restricted` as local fixture publisher policy. This does not decide MITRE redistribution permission. Only this isolated repository has allocated publication sequences. The append-only ledger is `.cartulary/reference-packs/fixture.attack.native.v2/publications/`; allocation is serialized, exact replay is stable, conflicting identities fail, and abandoned reservations remain. Metadata allocations are retained separately under the same repository's `envelopes/<manifest>/` tree.

| Sequence | Publication | Input SHA-256 | Disposition |
| ---: | --- | --- | --- |
| 1 | `attack-20.0.1` | `0c2f6ec6a07ab4c7bce5bcedfed9b0f037fa207d51d5f64f8b1f42d6a3a167ec` | Abandoned reservation from the earlier rejected assembly; preserved |
| 2 | `attack-20.0.2` | `17d199c7ae1cf21b75aec6ef59241b31d0f497e4b9b1cdeb9224a885a5e12401` | Earlier retained logical fixture; preserved |
| 3 | `attack-20.1.2` | `a376f9ee2c48d218541377a503b59a7c4f74f1e33b83defe01fe6208a3978963` | Earlier retained logical fixture; preserved |
| 4 | `attack-20.0.3` | `3173935dcd7e447a775d8061e5c0504ff47d1747717581cc24e8bf84d53d8472` | Final transition fixture |
| 5 | `attack-20.1.3` | `1e2ea1267715d1ed6ef7e56bfd330c7c3351070bf64ebc04b715cf331114e42d` | Final primary fixture |

The root authorizes three distinct root keys and requires two signatures. Each online role has its own key and one required signature. Public bootstrap SHA-256 is `eeed15d1d99bce09f92745c0f601167122a01fdd0ee57fe9b08a8e9b68d3c7e7`. Private signing material is host-local with restrictive permissions, outside containers, bootstrap files, and public evidence; no private bytes are recorded here. Isolated runtimes receive only their matching public bootstrap. Fresh workflow envelopes use wall-clock timestamps and the normal production verifier, without Reference Data frozen-clock controls.

Public package outputs are retained under `.cartulary/reference-packs/native-v2/public/envelopes/<container-sha256>/`, each with `container.pack`, `bootstrap.json`, `envelope-report.json`, and `independent-witness.json`. Initial metadata version 1 was issued `2026-11-06T21:51:53Z` and expires `2026-11-13T21:51:53Z`. Renewal metadata version 2 was issued `2026-11-06T21:52:59Z` and expires `2026-11-20T21:52:59Z`. Root expiry remains `2027-11-06T21:51:53Z`.

| Release / metadata version | Container bytes | Container SHA-256 |
| --- | ---: | --- |
| 20.1 / 1 | 75,061,248 | `a17d981c5ae28f5ae7c02b309ea3cd92d3f19e0cb6fc8ad5a59c71641ba66eb6` |
| 20.1 / 2 | 75,061,248 | `109bd29b43ceb0de502252faf91c1d3379a8b74a42e17c1721ad575490748d8e` |
| 20.0 / 1 | 74,998,272 | `4d2ed787b8d8c50e8c8f3c2d35471a02343a109a76875f9657a140eb3c5a301e` |
| 20.0 / 2 | 74,998,272 | `e8a8b6d3e95949db10ac63d91af2d6f11fd993b2d77e33f961bda72dd2e2e801` |

The independent envelope witnesses prove signature and byte integrity and explicitly report `freshness_verified=false`; production workflow runs supply fresh wall-clock trust verification. Fresh test envelopes and matching public bootstraps are retained under each consuming run's existing `_shared/test-services/<suite>/reference-pack-envelopes/<container-sha256>/` tree. Canonical fixture bytes remain frozen; private test keys remain only in caller memory.

### Supported workflow evidence and separate gates

Production routes and workers exercised authentic import with staged-only default, explicit activation, object/external-ID/unambiguous-alias lookup, provenance, immutable selection, identical replay, renewed metadata, disablement, and removal. The retained broader Reference Data rows exercise authorization, collisions, threshold/expiry failures, no egress, staging cleanup, and other-profile admission. Reporting activates authentic v20.0, admits a snapshot pin, activates authentic v20.1, and renders/rerenders using the original pin; removal of pinned content is refused. Incident-bundle v7 preserves native references through unavailable-content round trips. Restricted ATT&CK containers are excluded from embedding while references remain exportable, as required by the existing export contract. Composition previews use the admitted snapshot binding through Reporting. Historical-integrity/readiness and Recovery pre-target rejection remain covered by their owner boundaries.

| Required outcome | Status |
| --- | --- |
| Build reproducibility | Passed: exact executable, repeated logical assemblies, and repeated identical envelope inputs |
| Canonical conformance | Passed for both authentic all-domain baselines and the selected canonical/negative/limit boundaries |
| Fixture-trust verification | Passed through production import/workers using fresh isolated trust; independent byte/signature witnesses retained |
| Deployment-trust importability | Not established; fixture identities and keys do not authorize intended deployment import |
| Distribution approval | Not established; fixture publisher policy remains restricted and actual redistribution disposition is outstanding |
| Supported E2E completion | Backend production workflow scenarios passed; this is not full CI, full-owner, browser, deployment, or release acceptance |

Positive ATT&CK embedding requires an actual redistribution disposition. Relationship-pivot E2E still requires a separately adopted relationship consumer API. Neither gate was bypassed. Full CI, browser/visual suites, unrelated owners, benchmarks, deployment-trust imports, and release publication were not run because they are outside this change's selected obligations or explicitly separate gates. Retained warm-check maintenance was skipped because no compatible successful full warm-check root was supplied; no redundant warm-check was launched.

### Retained failures and repairs

All earlier failed roots remain evidence of their actual outcomes. They are not passing qualification rows. Initial generation failures rejected a missing typed `INPUT` default and migration-owner allocation; authored owners were corrected before regeneration. The first v20.0 acquisition invocation rejected a lock-byte mismatch; the exact separately retained lock was preserved, then offline source check/audit succeeded. Early builder/assembly failures exposed schema-document map merging, authored-versus-compiled schema identity binding, and native timestamp spelling preservation. The abandoned publication reservation remains sequence 1.

Early fixture runs exposed stale v1 expectations, signed synthetic fixtures, canonical entry hashes, and a missing collection-version field in a limit vector. The native timestamp annotation preserves original fractional spelling while other profile timestamps remain strict. Temporary diagnostic instrumentation was removed. A retained-secret scanner rejected diagnostic output containing token-shaped field names; the test now reports only safe check/path/code information, and the scanner was not weakened.

`attack-v2-authentic-20261106-01` failed on the obsolete Reporting imported-content column; the owner check now reads ordered chunks. `-02` exposed the root's three-authorized-key requirement and the frozen operation decoder's new member count; both were corrected. `-03` passed the authentic lifecycle itself but failed the other-profile consumer result because the new set schema still constrained every profile version to `"1"`; v2 set/reference schemas now admit `"1"` and `"2"`, with semantic admission enforcing the exact registry version for each key. Recovery's first selected run failed preflight because it preceded generation of the new portability projection; the generated successor passed. Extensions' first run exposed a stale major-2 assertion; its focused successor passed.

`attack-v2-reference-final-20261106-01` passed 10 selected rows and failed two: the compiler confinement test used the predecessor document-only cycle key, and lifecycle fixtures retained an old activated-set witness. The test now uses the document/fragment identity. Independently recomputing the complete Base-plus-synthetic-host tuple produced `d61e98deed7daa3b568a3f784b031956efb81d1cc5e17eda2d319b8d0064fc7f`; active fixture bindings now reference Base release v2. `attack-v2-reference-repair-20261106-01` passed all three selected rows: canonical fixture manifests, trust, and persistence. An additional final review added retention of fresh public workflow envelopes in the existing harness artifact tree; its consuming workflows were rerun separately.

The following command and file inventories document this implementation workspace; they are not executable contract inputs or release evidence. The changed-file inventory includes preserved pre-existing uncommitted acquisition/candidate work, rather than attributing all dirty paths to this cutover.

### Native v2 exact workspace file inventory

This is the full dirty/untracked file inventory at implementation handoff, including preserved acquisition work and historical copies. Generated paths were regenerated through Make. Ignored source/build/audit/run roots are documented separately above.

```text
contracts/extensions/build/implementation-bindings.json
contracts/extensions/fragments/core00.recognition.json
contracts/incident-bundles/source_catalog.json
contracts/recovery/fixtures/graph-projection-restore-implementation-binding.v9.json
contracts/recovery/fixtures/graph-projection-restore-rebuild-result.v6.json
contracts/recovery/fixtures/operator-recovery-journal-payload.v7.json
contracts/recovery/fixtures/recovery-generation-registry.v1.json
contracts/recovery/fixtures/recovery-state-catalog.v2.json
contracts/reference-pack-fixtures/fixture-manifests/admission/equivalent_gzip.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/equivalent_tar.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/equivalent_zip.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/license_prohibited.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/license_restricted.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/license_spdx_exception.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/license_spdx_valid.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/source_0b221a86ed33.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/source_f0c7012f41ea.v1.json
contracts/reference-pack-fixtures/fixture-manifests/admission/source_fe60f8ccc99b.v1.json
contracts/reference-pack-fixtures/fixture-manifests/complete_set.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/activate.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/cancel.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/disable.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/empty_refresh.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/import.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/reactivate.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/reconcile_claimed.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/reconcile_unclaimed.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/refresh.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/reimport.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/remove.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/renew.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/replay.v1.json
contracts/reference-pack-fixtures/fixture-manifests/lifecycle/reverify_active.v1.json
contracts/reference-pack-fixtures/fixture-manifests/manifest_identity.v1.json
contracts/reference-pack-fixtures/fixture-manifests/native_malformed.v1.json
contracts/reference-pack-fixtures/fixture-manifests/native_semantic.v1.json
contracts/reference-pack-fixtures/fixture-manifests/native_valid.v1.json
contracts/reference-pack-fixtures/fixture-manifests/owner_historical_readiness.v1.json
contracts/reference-pack-fixtures/fixture-manifests/owner_portable_destination.v1.json
contracts/reference-pack-fixtures/fixture-manifests/owner_snapshot_binding.v1.json
contracts/reference-pack-fixtures/fixture-manifests/profiles/framework_attack_malformed.v1.json
contracts/reference-pack-fixtures/fixture-manifests/profiles/framework_attack_semantic.v1.json
contracts/reference-pack-fixtures/fixture-manifests/signed_container.v1.json
contracts/reference-pack-fixtures/fixtures/admission/equivalent_gzip.v1.json
contracts/reference-pack-fixtures/fixtures/admission/equivalent_tar.v1.json
contracts/reference-pack-fixtures/fixtures/admission/equivalent_zip.v1.json
contracts/reference-pack-fixtures/fixtures/admission/json_bom.v1.json
contracts/reference-pack-fixtures/fixtures/admission/json_duplicate.v1.json
contracts/reference-pack-fixtures/fixtures/admission/json_invalid_utf8.v1.json
contracts/reference-pack-fixtures/fixtures/admission/json_noncanonical.v1.json
contracts/reference-pack-fixtures/fixtures/admission/json_unknown.v1.json
contracts/reference-pack-fixtures/fixtures/admission/license_document_ref.v1.json
contracts/reference-pack-fixtures/fixtures/admission/license_prohibited.v1.json
contracts/reference-pack-fixtures/fixtures/admission/license_restricted.v1.json
contracts/reference-pack-fixtures/fixtures/admission/license_spdx_exception.v1.json
contracts/reference-pack-fixtures/fixtures/admission/license_spdx_invalid.v1.json
contracts/reference-pack-fixtures/fixtures/admission/license_spdx_valid.v1.json
contracts/reference-pack-fixtures/fixtures/admission/license_unbound_ref.v1.json
contracts/reference-pack-fixtures/fixtures/admission/license_unregistered.v1.json
contracts/reference-pack-fixtures/fixtures/admission/source_0b221a86ed33.v1.json
contracts/reference-pack-fixtures/fixtures/admission/source_f0c7012f41ea.v1.json
contracts/reference-pack-fixtures/fixtures/admission/source_fe60f8ccc99b.v1.json
contracts/reference-pack-fixtures/fixtures/attestation.v1.json
contracts/reference-pack-fixtures/fixtures/complete-set.v1.json
contracts/reference-pack-fixtures/fixtures/manifest.v1.json
contracts/reference-pack-fixtures/fixtures/native_malformed.v1.json
contracts/reference-pack-fixtures/fixtures/native_semantic.v1.json
contracts/reference-pack-fixtures/fixtures/native_valid.v1.json
contracts/reference-pack-fixtures/fixtures/portable-content.v1.json
contracts/reference-pack-fixtures/fixtures/portable-input.v1.json
contracts/reference-pack-fixtures/fixtures/portable-references.v1.json
contracts/reference-pack-fixtures/fixtures/profiles.v1.json
contracts/reference-pack-fixtures/fixtures/profiles/framework_attack_malformed.v1.json
contracts/reference-pack-fixtures/fixtures/profiles/framework_attack_semantic.v1.json
contracts/reference-pack-fixtures/fixtures/profiles/framework_attack_valid.v1.json
contracts/reference-pack-fixtures/fixtures/signed-container.v1.json
contracts/reference-pack-producers/attack/acquisition.v1.schema.json
contracts/reference-pack-producers/attack/candidate-index.v1.json
contracts/reference-pack-producers/attack/dispositions.v1.json
contracts/reference-pack-producers/attack/field-policy.candidate.v1.json
contracts/reference-pack-producers/attack/field-policy.native.v1.json
contracts/reference-pack-producers/attack/fixtures/cases.v1.json
contracts/reference-pack-producers/attack/fixtures/inputs.v1.json
contracts/reference-pack-producers/attack/fixtures/native-cases.v1.json
contracts/reference-pack-producers/attack/inputs.go
contracts/reference-pack-producers/attack/native-source.candidate.v1.schema.json
contracts/reference-pack-producers/attack/object.candidate.v1.schema.json
contracts/reference-pack-producers/attack/relationship.candidate.v1.schema.json
contracts/reference-pack-producers/attack/source-lock.v20.0.json
contracts/reference-pack-producers/attack/source-lock.v20.1.json
contracts/reference-pack-producers/attack/source-profile.candidate.v1.json
contracts/reference-pack-producers/attack/source-profile.native.v1.json
contracts/reference-pack-producers/attack/source-profile.native.v2.json
contracts/reference-packs/builtins/release.v2.json
contracts/reference-packs/builtins/witness.v2.json
contracts/reference-packs/consumer_item.v2.schema.json
contracts/reference-packs/consumer_operations.v2.json
contracts/reference-packs/get_pack_entry_result.v2.schema.json
contracts/reference-packs/lookup_pack_entries_result.v2.schema.json
contracts/reference-packs/operation_input.v3.schema.json
contracts/reference-packs/pack_set.v2.schema.json
contracts/reference-packs/portability_input.v2.schema.json
contracts/reference-packs/profiles.v2.json
contracts/reference-packs/profiles/framework.attack.native_source.v2.schema.json
contracts/reference-packs/profiles/framework.attack.object.v2.schema.json
contracts/reference-packs/profiles/framework.attack.relationship.v2.schema.json
contracts/reference-packs/reference_pack_refs.v2.schema.json
contracts/reference-packs/resolve_current_pack_set_result.v2.schema.json
db/migrations/00048_reference_pack_native_index_bound.sql
docs/extension-subsystem-nlspec.md
docs/handoffs/mitre-attack-reference-pack.md
docs/reference-pack-subsystem-nlspec.md
docs/report-composition-nlspec.md
docs/reporting-subsystem-nlspec.md
docs/spec/00_document_set_status_and_precedence.md
docs/spec/01_architecture_storage_and_view_contracts.md
docs/spec/02_domain_model_schema_and_history.md
docs/spec/03_workbook_interaction_collaboration_and_workflows.md
docs/spec/04_security_deployment_and_conformance.md
docs/testing-harness-nlspec.md
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/build-report.json
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/evidence/field-policy.json
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/evidence/publication-input.json
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/evidence/source-lock.json
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/evidence/source-profile.json
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/manifest.json
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/notices/MITRE-ATTACK.txt
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/payload/objects.ndjson
fixtures/reference-packs/attack-native-v2/20.0/logical/b71707064c23db3ba8ead6155ea95e9973126d83a9b975028080bed8fa0e0c7a/payload/relationships.ndjson
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/build-report.json
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/evidence/field-policy.json
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/evidence/publication-input.json
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/evidence/source-lock.json
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/evidence/source-profile.json
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/independent-witness.json
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/manifest.json
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/notices/MITRE-ATTACK.txt
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/payload/objects.ndjson
fixtures/reference-packs/attack-native-v2/20.0/logical/c2ae7651ccd2799c1836f09e67173a00140030a60b267c9cf180e12b25e3dd90/payload/relationships.ndjson
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/build-report.json
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/evidence/field-policy.json
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/evidence/publication-input.json
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/evidence/source-lock.json
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/evidence/source-profile.json
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/independent-witness.json
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/manifest.json
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/notices/MITRE-ATTACK.txt
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/payload/objects.ndjson
fixtures/reference-packs/attack-native-v2/20.1/logical/45c74160814ef850777aad3d2a50207dafdafe112aa36bc9ea80a3ec81e8703f/payload/relationships.ndjson
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/build-report.json
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/evidence/field-policy.json
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/evidence/publication-input.json
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/evidence/source-lock.json
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/evidence/source-profile.json
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/independent-witness.json
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/manifest.json
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/notices/MITRE-ATTACK.txt
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/payload/objects.ndjson
fixtures/reference-packs/attack-native-v2/20.1/logical/b3c82d559952d77252528db857976458d430c61bc9b889db51b0b78f6c56470b/payload/relationships.ndjson
fixtures/reference-packs/attack-native-v2/catalog.json
fixtures/reference-packs/attack-native-v2/catalogs/93e1982e2b0768df1b06cd36ca5585b25955b3a1ebb1a28aef9c92cb390e9199.json
fixtures/reference-packs/attack-native-v2/derivatives/malformed/objects.ndjson
fixtures/reference-packs/attack-native-v2/derivatives/semantic/objects.ndjson
fixtures/reference-packs/historical-v1/fixture-manifests/admission/equivalent_gzip.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/equivalent_tar.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/equivalent_zip.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_case_collision.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_concatenated.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_device.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_directory_double_slash.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_directory_traversal.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_fifo.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_hardlink.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_parent_collision.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_parent_collision_reverse.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_pax.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_sparse.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_symlink.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/gzip_traversal.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/json_bom.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/json_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/json_invalid_utf8.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/json_noncanonical.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/json_nonobject.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/json_surrogate.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/json_unknown.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/json_unknown_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/license_document_ref.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/license_prohibited.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/license_restricted.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/license_spdx_exception.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/license_spdx_invalid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/license_spdx_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/license_unbound_ref.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/license_unregistered.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/non_ustar.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/source_0b221a86ed33.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/source_f0c7012f41ea.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/source_fe60f8ccc99b.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_case_collision.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_device.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_directory_double_slash.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_directory_traversal.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_fifo.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_hardlink.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_parent_collision.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_parent_collision_reverse.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_pax.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_sparse.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_symlink.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_traversal.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/tar_truncated.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip64.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip_case_collision.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip_directory_double_slash.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip_directory_traversal.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip_parent_collision.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip_parent_collision_reverse.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip_symlink.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/admission/zip_traversal.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/complete_set.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/dependency/conflict.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/dependency/cycle.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/dependency/missing.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/dependency/transitive_missing.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/dependency/valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/activate.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/cancel.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/disable.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/empty_refresh.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/import.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/reactivate.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/reconcile_claimed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/reconcile_unclaimed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/refresh.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/reimport.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/remove.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/renew.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/replay.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/lifecycle/reverify_active.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/manifest_identity.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/owner_historical_readiness.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/owner_portable_destination.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/owner_snapshot_binding.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_cisa_kev_boundary.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_cisa_kev_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_cisa_kev_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_cisa_kev_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_entra_app_ids_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_entra_app_ids_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_entra_app_ids_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_hijacklibs_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_hijacklibs_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_hijacklibs_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_lolbas_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_lolbas_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_lolbas_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_loldrivers_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_loldrivers_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_loldrivers_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_lolesxi_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_lolesxi_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_lolesxi_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_ms_portals_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_ms_portals_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_ms_portals_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_tor_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_tor_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_tor_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_event_ids_boundary.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_event_ids_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_event_ids_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_event_ids_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_sids_aliases_over.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_sids_boundary.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_sids_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_sids_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_sids_source_refs_over.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/enrichment_windows_sids_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_attack_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_attack_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_attack_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_d3fend_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_d3fend_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_d3fend_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_veris_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_veris_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/framework_veris_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_evidence_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_evidence_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_evidence_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_host_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_host_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_host_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_indicator_malformed.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_indicator_semantic.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/profiles/type_registry_indicator_valid.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/set_cardinality_64.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/set_cardinality_65.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/signed_container.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/all_roles_bad_signatures.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/all_roles_expired.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/all_roles_rollback.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/canonical_before_root.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/expiry_equality.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/length_before_hash.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/missing_declarations.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/missing_payloads.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/ordinary_3.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/ordinary_3_replay.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/root_history_missing.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/rotation_3.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/rotation_3_replay.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/rotation_64.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/rotation_64_replay.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/rotation_gap.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/rotation_predecessor_only.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/rotation_successor_only.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/trust/target_substitution.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/type_compatibility/ambiguous_alias.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/type_compatibility/changed_algorithm.v1.json
fixtures/reference-packs/historical-v1/fixture-manifests/type_compatibility/removed_key.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/equivalent_gzip.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/equivalent_tar.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/equivalent_zip.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_case_collision.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_concatenated.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_device.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_directory_double_slash.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_directory_traversal.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_fifo.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_hardlink.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_parent_collision.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_parent_collision_reverse.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_pax.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_sparse.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_symlink.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/gzip_traversal.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/json_bom.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/json_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/json_invalid_utf8.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/json_noncanonical.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/json_nonobject.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/json_surrogate.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/json_unknown.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/json_unknown_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/license_document_ref.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/license_prohibited.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/license_restricted.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/license_spdx_exception.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/license_spdx_invalid.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/license_spdx_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/license_unbound_ref.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/license_unregistered.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/non_ustar.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/source_0b221a86ed33.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/source_f0c7012f41ea.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/source_fe60f8ccc99b.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_case_collision.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_device.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_directory_double_slash.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_directory_traversal.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_fifo.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_hardlink.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_parent_collision.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_parent_collision_reverse.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_pax.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_sparse.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_symlink.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_traversal.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/tar_truncated.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip64.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip_case_collision.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip_directory_double_slash.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip_directory_traversal.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip_duplicate.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip_parent_collision.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip_parent_collision_reverse.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip_symlink.v1.json
fixtures/reference-packs/historical-v1/fixtures/admission/zip_traversal.v1.json
fixtures/reference-packs/historical-v1/fixtures/attestation.v1.json
fixtures/reference-packs/historical-v1/fixtures/complete-set.v1.json
fixtures/reference-packs/historical-v1/fixtures/dependency/conflict.v1.json
fixtures/reference-packs/historical-v1/fixtures/dependency/cycle.v1.json
fixtures/reference-packs/historical-v1/fixtures/dependency/missing.v1.json
fixtures/reference-packs/historical-v1/fixtures/dependency/transitive_missing.v1.json
fixtures/reference-packs/historical-v1/fixtures/dependency/valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/lifecycle.v1.json
fixtures/reference-packs/historical-v1/fixtures/manifest.v1.json
fixtures/reference-packs/historical-v1/fixtures/owner_historical_readiness.v1.json
fixtures/reference-packs/historical-v1/fixtures/owner_portable_destination.v1.json
fixtures/reference-packs/historical-v1/fixtures/owner_snapshot_binding.v1.json
fixtures/reference-packs/historical-v1/fixtures/portable-content.v1.json
fixtures/reference-packs/historical-v1/fixtures/portable-input.v1.json
fixtures/reference-packs/historical-v1/fixtures/portable-references.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_cisa_kev_boundary.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_cisa_kev_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_cisa_kev_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_cisa_kev_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_entra_app_ids_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_entra_app_ids_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_entra_app_ids_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_hijacklibs_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_hijacklibs_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_hijacklibs_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_lolbas_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_lolbas_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_lolbas_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_loldrivers_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_loldrivers_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_loldrivers_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_lolesxi_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_lolesxi_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_lolesxi_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_ms_portals_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_ms_portals_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_ms_portals_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_tor_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_tor_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_tor_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_event_ids_boundary.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_event_ids_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_event_ids_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_event_ids_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_sids_aliases_over.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_sids_boundary.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_sids_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_sids_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_sids_source_refs_over.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/enrichment_windows_sids_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_attack_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_attack_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_attack_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_d3fend_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_d3fend_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_d3fend_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_veris_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_veris_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/framework_veris_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_evidence_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_evidence_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_evidence_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_host_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_host_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_host_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_indicator_malformed.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_indicator_semantic.v1.json
fixtures/reference-packs/historical-v1/fixtures/profiles/type_registry_indicator_valid.v1.json
fixtures/reference-packs/historical-v1/fixtures/set_cardinality_64.v1.json
fixtures/reference-packs/historical-v1/fixtures/set_cardinality_65.v1.json
fixtures/reference-packs/historical-v1/fixtures/signed-container.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/all_roles_bad_signatures.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/all_roles_expired.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/all_roles_rollback.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/canonical_before_root.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/expiry_equality.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/length_before_hash.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/missing_declarations.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/missing_payloads.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/ordinary_3.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/ordinary_3_replay.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/root_history_missing.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/rotation_3.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/rotation_3_replay.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/rotation_64.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/rotation_64_replay.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/rotation_gap.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/rotation_predecessor_only.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/rotation_successor_only.v1.json
fixtures/reference-packs/historical-v1/fixtures/trust/target_substitution.v1.json
fixtures/reference-packs/historical-v1/fixtures/tuf.v1.json
fixtures/reference-packs/historical-v1/fixtures/type_compatibility/ambiguous_alias.v1.json
fixtures/reference-packs/historical-v1/fixtures/type_compatibility/changed_algorithm.v1.json
fixtures/reference-packs/historical-v1/fixtures/type_compatibility/removed_key.v1.json
fixtures/reference-packs/historical-v1/recovery/graph-projection-restore-implementation-binding.v9.json
fixtures/reference-packs/historical-v1/recovery/graph-projection-restore-rebuild-result.v6.json
fixtures/reference-packs/historical-v1/recovery/operator-recovery-journal-payload.v7.json
fixtures/reference-packs/historical-v1/recovery/recovery-generation-registry.v1.json
fixtures/reference-packs/historical-v1/recovery/recovery-state-catalog.v2.json
internal/app/migrate/migrate.go
internal/app/server/runtime_assembly.go
internal/gen/contractextensions/artifacts_gen.go
internal/gen/contractrecovery/artifacts_gen.go
internal/gen/contractrecovery/graph_projection_restore_binding_gen.go
internal/gen/contractrecovery/recovery_generation_registry_gen.go
internal/gen/contractreferencepackfixtures/artifacts_gen.go
internal/gen/contractreferencepacks/artifacts_gen.go
internal/modules/extensions/coordinator_test.go
internal/modules/incidentbundles/api_test.go
internal/modules/incidentbundles/routes_helpers_integration_test.go
internal/modules/incidentbundles/routes_round_trip_integration_test.go
internal/modules/incidentbundles/source_catalog_test.go
internal/modules/recovery/vnext_graph_restore_artifacts_test.go
internal/modules/reference_data/attack_native_integration_test.go
internal/modules/reference_data/base_release.go
internal/modules/reference_data/contractadmission/admission.go
internal/modules/reference_data/contractadmission/binding.go
internal/modules/reference_data/coordinator_admission.go
internal/modules/reference_data/coordinator_execution.go
internal/modules/reference_data/coordinator_jobs_integration_test.go
internal/modules/reference_data/historical_validation.go
internal/modules/reference_data/incident_portability_test.go
internal/modules/reference_data/incident_retention.go
internal/modules/reference_data/internal/packformat/admission_fixture_test.go
internal/modules/reference_data/internal/packformat/archive_limit_test.go
internal/modules/reference_data/internal/packformat/archive_writer.go
internal/modules/reference_data/internal/packformat/attack.go
internal/modules/reference_data/internal/packformat/attack_native_test.go
internal/modules/reference_data/internal/packformat/builtin_binding_test.go
internal/modules/reference_data/internal/packformat/consumer_contract.go
internal/modules/reference_data/internal/packformat/content.go
internal/modules/reference_data/internal/packformat/content_admission_test.go
internal/modules/reference_data/internal/packformat/content_diagnostics.go
internal/modules/reference_data/internal/packformat/content_test.go
internal/modules/reference_data/internal/packformat/fixture_contract_test.go
internal/modules/reference_data/internal/packformat/framework_lookup_test.go
internal/modules/reference_data/internal/packformat/identity.go
internal/modules/reference_data/internal/packformat/limit_diagnostics_test.go
internal/modules/reference_data/internal/packformat/manifest.go
internal/modules/reference_data/internal/packformat/native_fixture_test.go
internal/modules/reference_data/internal/packformat/portability.go
internal/modules/reference_data/internal/packformat/portability_retention.go
internal/modules/reference_data/internal/packformat/portability_retention_test.go
internal/modules/reference_data/internal/packformat/portable_content_test.go
internal/modules/reference_data/internal/packformat/prepared_test.go
internal/modules/reference_data/internal/packformat/shape.go
internal/modules/reference_data/internal/packformat/shape_diagnostics.go
internal/modules/reference_data/portable_admission_integration_test.go
internal/modules/reference_data/producer/audit.go
internal/modules/reference_data/producer/base.go
internal/modules/reference_data/producer/candidate.go
internal/modules/reference_data/producer/candidate_test.go
internal/modules/reference_data/producer/envelope.go
internal/modules/reference_data/producer/identity.go
internal/modules/reference_data/producer/ledger.go
internal/modules/reference_data/producer/ledger_test.go
internal/modules/reference_data/producer/package.go
internal/modules/reference_data/producer/source.go
internal/modules/reference_data/producer/source_test.go
internal/modules/reference_data/profile_consumers_integration_test.go
internal/modules/reference_data/recoverycontribution/contribution.go
internal/modules/reference_data/recoverycontribution/inventory.go
internal/modules/reference_data/reference_pack_integration_test.go
internal/modules/reference_data/reference_pack_storage_reference_migration_test.go
internal/modules/reference_data/reference_pack_test_helpers_test.go
internal/modules/reference_data/testsupport/attack_fixture.go
internal/modules/reference_data/testsupport/attack_http.go
internal/modules/reporting/composition_preview_integration_test.go
internal/modules/reporting/migrationadmission/reference.go
internal/modules/reporting/reference_binding_integration_test.go
internal/modules/reporting/reporting_integration_test.go
internal/testutil/harnessruntime/policy_generated.go
packages/protocol-ts/src/generated/extension-profile-registry.ts
tools/database-migrations/generate-catalog-projections.mjs
tools/database-migrations/owner-allocation.mjs
tools/execution_topology_render_index.json
tools/harness/browser/reference-pack-fixture.mjs
tools/harness/contract/generated/schema-validators/validateCartularyIncidentBundleSourceCatalogV6.cjs
tools/harness/fixtures/incident-bundle.mjs
tools/harness_work_graph_owner.json
tools/migration_history_manifest.json
tools/reference-packs/build.sh
tools/reference-packs/main.go
tools/reference-packs/package.sh
tools/reference-packs/source.sh
tools/reference-packs/witness.mjs
tools/release-evidence/reference-pack-fixtures.mjs
tools/schema_object_ownership_manifest.json
tools/schemas/cartulary.incident_bundle_source_catalog.v6.schema.json
tools/task_surface.generated.mk
tools/task_surface.help-all.generated.json
tools/task_surface.runtime.generated.mk
tools/task_surface_manifest.json
tools/task_surface_owner.json
tools/test_families/module.incidentbundles.json
tools/test_families/module.reference_data.json
tools/test_families/module.reportcomposition.json
tools/test_families/module.reporting.json
```

### Native v2 executed command ledger

All run roots below are under `.cartulary/test-results/`. Graph targets retain `run-manifest.json`, `run-summary.json`, row/unit results, and cleanup results; tool targets retain `<target>/tool-run-summary.json` and step logs. Commands below preserve the actual declared selections; normalized argument ordering is immaterial. Prefix each command with `CARTULARY_TEST_RUN_ID=<run-id>`. Passing commands completed with outer exit 0. Failed Make invocations returned nonzero (normally 2); their summaries retain the more specific child/normalized code and failure classification.

| Run ID | Public command | Result |
| --- | --- | --- |
| `attack-v2-assemble200-20261106-01` | `make reference-pack-package` | fail; unknown_failure |
| `attack-v2-assemble200-20261106-02` | `make reference-pack-package` | fail; unknown_failure |
| `attack-v2-assemble200-20261106-03` | `make reference-pack-package` | fail; unknown_failure |
| `attack-v2-assemble200-20261106-04` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2/inputs/assemble-20.0-04.json` | pass |
| `attack-v2-assemble201-20261106-01` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2/inputs/assemble-20.1-04.json` | pass |
| `attack-v2-authentic-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary HARNESS_DIAGNOSTICS=basic` | fail; test_assertion_failure |
| `attack-v2-authentic-20261106-02` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary,module.reference_data.integration.reference_pack_storage_reference_cutover_is_reset_dc28423946 HARNESS_DIAGNOSTICS=basic` | fail; test_assertion_failure |
| `attack-v2-authentic-20261106-03` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary,module.reference_data.integration.reference_pack_storage_reference_cutover_is_reset_dc28423946,module.reference_data.unit.canonical_pack_profiles HARNESS_DIAGNOSTICS=basic` | fail; test_assertion_failure |
| `attack-v2-base-reissue-20261106-01` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2/inputs/base.json` | pass |
| `attack-v2-builder-01` | `make reference-pack-producer-build` | pass |
| `attack-v2-builder-20261106-01` | `make reference-pack-producer-build` | fail; unknown_failure |
| `attack-v2-builder-20261106-02` | `make reference-pack-producer-build` | pass |
| `attack-v2-builder-20261106-03` | `make reference-pack-producer-build` | pass |
| `attack-v2-builder-20261106-04` | `make reference-pack-producer-build` | pass |
| `attack-v2-builder-20261106-05` | `make reference-pack-producer-build` | pass |
| `attack-v2-builder-20261106-06` | `make reference-pack-producer-build` | pass |
| `attack-v2-builder-20261106-07` | `make reference-pack-producer-build` | pass |
| `attack-v2-builder-20261106-08` | `make reference-pack-producer-build` | pass |
| `attack-v2-builder-20261106-09` | `make reference-pack-producer-build` | pass |
| `attack-v2-bundles-20261106-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8 HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-bundles-final-20261106-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8 HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-composition-20261106-01` | `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.preview_reporting_delegation_ecdd2bc64a HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-composition-final-20261106-01` | `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.preview_reporting_delegation_ecdd2bc64a HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-envelope-retention-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-envelope200-20261106-01` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/native-v2/inputs/envelope-20.0-final.json` | pass |
| `attack-v2-envelope200-20261106-02` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/native-v2/inputs/envelope-20.0-final.json` | pass |
| `attack-v2-envelope201-20261106-01` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/native-v2/inputs/envelope-20.1-final.json` | pass |
| `attack-v2-envelope201-20261106-02` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/native-v2/inputs/envelope-20.1-final.json` | pass |
| `attack-v2-extensions-20261106-01` | `make test-slice OWNER=module.extensions ROWS=module.extensions.unit.coordinator_registry_1dce9a539a,module.extensions.unit.generated_recovery_composition_preserves_admitte_1e77c4dab8,module.extensions.unit.publication_catalog_exact_profile_contribution_s_15f926e0bd HARNESS_DIAGNOSTICS=basic` | fail; test_assertion_failure |
| `attack-v2-extensions-20261106-02` | `make test-slice OWNER=module.extensions ROWS=module.extensions.unit.coordinator_registry_1dce9a539a HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-final-assemble200-20261106-01` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2/inputs/assemble-20.0-final.json` | pass |
| `attack-v2-final-assemble200-20261106-02` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2/inputs/assemble-20.0-final.json` | pass |
| `attack-v2-final-assemble201-20261106-01` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2/inputs/assemble-20.1-final.json` | pass |
| `attack-v2-final-assemble201-20261106-02` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2/inputs/assemble-20.1-final.json` | pass |
| `attack-v2-finalize-20261106-01` | `make agent-finalize` | pass |
| `attack-v2-finalize-20261106-02` | `make agent-finalize` | pass |
| `attack-v2-finalize-20261106-03` | `make agent-finalize` | pass |
| `attack-v2-finalize-20261106-04` | `make agent-finalize` | pass |
| `attack-v2-format-20261106-01` | `make format` | pass |
| `attack-v2-format-20261106-02` | `make format` | pass |
| `attack-v2-format-20261106-03` | `make format` | pass |
| `attack-v2-format-20261106-04` | `make format` | pass |
| `attack-v2-format-20261106-05` | `make format` | pass |
| `attack-v2-format-20261106-06` | `make format` | pass |
| `attack-v2-format-20261106-07` | `make format` | pass |
| `attack-v2-format-20261106-08` | `make format` | pass |
| `attack-v2-format-20261106-09` | `make format` | pass |
| `attack-v2-generate-20261106-01` | `make generate` | fail; unknown_failure |
| `attack-v2-generate-20261106-02` | `make generate` | fail; unknown_failure |
| `attack-v2-generate-20261106-03` | `make generate` | pass |
| `attack-v2-generate-20261106-04` | `make generate` | pass |
| `attack-v2-generate-20261106-05` | `make generate` | pass |
| `attack-v2-generate-20261106-06` | `make generate` | pass |
| `attack-v2-generate-20261106-07` | `make generate` | pass |
| `attack-v2-generate-20261106-08` | `make generate` | pass |
| `attack-v2-generate-20261106-09` | `make generate` | pass |
| `attack-v2-generate-20261106-10` | `make generate` | pass |
| `attack-v2-generate-20261106-11` | `make generate` | pass |
| `attack-v2-generate-20261106-12` | `make generate` | pass |
| `attack-v2-generate-20261106-13` | `make generate` | pass |
| `attack-v2-generate-20261106-14` | `make generate` | pass |
| `attack-v2-generate-20261106-15` | `make generate` | pass |
| `attack-v2-generate-20261106-16` | `make generate` | pass |
| `attack-v2-generate-20261106-17` | `make generate` | pass |
| `attack-v2-generated-final-20261106-01` | `make source-check CHECK=generated` | pass |
| `attack-v2-generated-final-20261106-02` | `make source-check CHECK=generated` | pass |
| `attack-v2-migrate-20261106-01` | `make test-slice OWNER=app.migrate ROWS=app.migrate.unit.migrate_facade_contract HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-recovery-20261106-01` | `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c HARNESS_DIAGNOSTICS=basic` | fail; preflight_error |
| `attack-v2-recovery-20261106-02` | `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-reference-final-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary,module.reference_data.integration.canonical_persistence,module.reference_data.integration.disconnected_reference_pack_failure_modes_produc_d1bffa4cca,module.reference_data.integration.reference_pack_import_jobs_re_derive_route_famil_dd456a1690,module.reference_data.integration.reference_pack_storage_reference_cutover_is_reset_dc28423946,module.reference_data.integration.reference_pack_upload_envelope_failures_create_n_710dd1d70f,module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_trust HARNESS_DIAGNOSTICS=basic` | fail; test_assertion_failure |
| `attack-v2-reference-repair-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_trust HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-renew200-20261106-01` | `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/native-v2/inputs/renew-20.0-final.json` | pass |
| `attack-v2-renew201-20261106-01` | `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/native-v2/inputs/renew-20.1-final.json` | pass |
| `attack-v2-reporting-20261106-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-reporting-final-20261106-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-reporting-retained-20261106-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-schema-20261106-01` | `make source-check CHECK=schema` | pass |
| `attack-v2-schema-final-20261106-01` | `make source-check CHECK=schema` | pass |
| `attack-v2-schema-final-20261106-02` | `make source-check CHECK=schema` | pass |
| `attack-v2-server-20261106-01` | `make test-slice OWNER=app.server ROWS=app.server.integration.borrowed_application_fixtures_and_deployment_byt_f2fa662581,app.server.unit.extension_policy_catalog_7f6ad80724 HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-source-acquire-20261106-01` | `make reference-pack-source` | fail; unknown_failure |
| `attack-v2-source-audit-20261106-01` | `make reference-pack-source ACTION=audit BUILT_AT=2026-11-06T19:00:00Z SOURCE_DIR=.cartulary/reference-packs/attack/20.0 SOURCE_LOCK=contracts/reference-pack-producers/attack/source-lock.v20.0.json` | pass |
| `attack-v2-source-check-20261106-01` | `make reference-pack-source ACTION=check SOURCE_DIR=.cartulary/reference-packs/attack/20.0 SOURCE_LOCK=contracts/reference-pack-producers/attack/source-lock.v20.0.json` | pass |
| `attack-v2-tooling-20261106-01` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.public_registry_parity HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-unitfinal-focus-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_pack_manifest HARNESS_DIAGNOSTICS=basic` | pass |
| `attack-v2-unitfixtures-20261106-03` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests` | fail; test_assertion_failure |
| `attack-v2-unitfixtures-20261106-04` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests HARNESS_DIAGNOSTICS=basic` | fail; test_assertion_failure |
| `attack-v2-unitfixtures-20261106-06` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests` | fail; test_assertion_failure |
| `attack-v2-unitfixtures-20261106-07` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests` | fail; test_assertion_failure |
| `attack-v2-unitfixtures-20261106-08` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests` | fail; test_assertion_failure |
| `attack-v2-unitfixtures-20261106-09` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles` | fail; test_assertion_failure |
| `attack-v2-unitlimits-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_profiles HARNESS_DIAGNOSTICS=basic` | fail; test_assertion_failure |
| `attack-v2-unitprofiles-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles` | fail; test_assertion_failure |
| `attack-v2-unitprofiles-20261106-02` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles` | fail; test_assertion_failure |
| `attack-v2-witness-probe-20261106-01` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2/inputs/assemble-20.1-04.json` | pass |

Selection used `make task-guide` for `module.reference_data`, `module.reporting`, `module.reportcomposition`, `module.incidentbundles`, `module.recovery`, `module.extensions`, `app.migrate`, `app.server`, and `harness.command_surface`. Scope was broadened only for their affected contract, startup, fixture, and task-surface boundaries. `git diff --check` also passed. Final `make agent-finalize` was followed by generated/schema checks and a fresh executable reproduction; documentation lint is recorded in the closing verification entry.

### Closing verification and retained workflow envelopes

After the final artifact-retention change, `attack-v2-format-20261106-09` and `attack-v2-finalize-20261106-04` passed. `attack-v2-generated-final-20261106-02` and `attack-v2-schema-final-20261106-02` passed with outer exit 0. `attack-v2-builder-20261106-09` again built two identical executables with the frozen SHA-256 `5bbc51fcc92494ab0f6e8406f38c422eff2583815f12aae442c541cc42bd7f9a`; the final test-only changes did not alter the retained producer. No final canonical reissuance or new ledger allocation was needed.

The following consuming workflows all completed with outer exit 0, passed row results, and completed cleanup actions. Their public envelopes are retained; no private signing material is present in these artifacts. They supplement the repaired canonical fixture/trust/persistence rows and earlier affected-owner portfolio rather than claiming full-owner or release acceptance.

| Run ID | Exact selected public command | Duration ms | Source digest |
| --- | --- | ---: | --- |
| `attack-v2-envelope-retention-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary HARNESS_DIAGNOSTICS=basic` | 77,434 | `sha256:4182711abf03ce60f5092f9be9d27a4c72a2e0c23178861cfaa1ee593dbfb9ef` |
| `attack-v2-reporting-retained-20261106-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c HARNESS_DIAGNOSTICS=basic` | 73,046 | `sha256:4182711abf03ce60f5092f9be9d27a4c72a2e0c23178861cfaa1ee593dbfb9ef` |
| `attack-v2-bundles-retained-20261106-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8 HARNESS_DIAGNOSTICS=basic` | 94,584 | `sha256:4182711abf03ce60f5092f9be9d27a4c72a2e0c23178861cfaa1ee593dbfb9ef` |
| `attack-v2-composition-retained-20261106-01` | `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.preview_reporting_delegation_ecdd2bc64a HARNESS_DIAGNOSTICS=basic` | 112,694 | `sha256:4182711abf03ce60f5092f9be9d27a4c72a2e0c23178861cfaa1ee593dbfb9ef` |

| Retained workflow artifact root, relative to `.cartulary/test-results/` | Container bytes | Bootstrap SHA-256 |
| --- | ---: | --- |
| `attack-v2-envelope-retention-20261106-01/_shared/test-services/81bdafb744814c7185fdd0c7/reference-pack-envelopes/1ae7712fba6ab665df026ab4b0c402c30f4c94087d847892c38d518660ae425c` | 75,061,248 | `e7e76b8962c9c89425d52a0e506b1fc37c118867343af855cba8961706719738` |
| `attack-v2-envelope-retention-20261106-01/_shared/test-services/81bdafb744814c7185fdd0c7/reference-pack-envelopes/845b76f74bf733e4c7f05a8f8713b6af67c7769e57efc1e25aee2474270974e0` | 75,061,248 | `e7e76b8962c9c89425d52a0e506b1fc37c118867343af855cba8961706719738` |
| `attack-v2-reporting-retained-20261106-01/_shared/test-services/bba490546c8c47e0bad4923a/reference-pack-envelopes/7dc5e35d6462e5d5d09b750fab5880869db2b7f06269970838ae4e30b5c2830c` | 75,061,248 | `ac9e7a69b4d33eaee449365202512344764599991ae972a66421eabb249b8231` |
| `attack-v2-reporting-retained-20261106-01/_shared/test-services/bba490546c8c47e0bad4923a/reference-pack-envelopes/cf96c539e85f79d7d4869e58382e415ab57f43d78a1031dd17f2dd3c68eb10c0` | 74,998,272 | `ac9e7a69b4d33eaee449365202512344764599991ae972a66421eabb249b8231` |
| `attack-v2-bundles-retained-20261106-01/_shared/test-services/44f8227bd69c418aa001dd8f/reference-pack-envelopes/ec7b38453d0337557c2aec40fe4bb5adf3b96ad0c72e2dc048530cc33025f057` | 75,061,248 | `997ca0e89bb76923fcfd074baa9cf6f86a1441d836e8ac00a0bcfa4ece827524` |
| `attack-v2-composition-retained-20261106-01/_shared/test-services/2754176ac3aa42268e6263dd/reference-pack-envelopes/fad313dd2efd813b7098156e34f5927a749c82a4b4a293f7b9f9f791a254a879` | 75,061,248 | `adaddfa18afb90c85c3c68221c9fd71bc310e1640cf75e04332ee4763eeee7d9` |

Canonical byte comparisons, independent witnesses, and the selected production boundary runs establish the six outcomes above within their stated scope. The unchanged deployment-trust, distribution, positive-embedding, and relationship-API gates remain outstanding.

Markdown lint run `attack-v2-markdown-20261106-01` returned outer exit 2 (child exit 1; harness/unknown_failure) because bare regular expressions in the newly adopted native-property tables were parsed as Markdown reference links (MD052). The repair wraps those literals in code spans without changing their requirements. The final documentation invocation is `CARTULARY_TEST_RUN_ID=attack-v2-markdown-20261106-02 make lint-markdown`; its exact result is retained under `.cartulary/test-results/attack-v2-markdown-20261106-02/adhoc/lint-markdown/`. No product regeneration or rerun is required for this prose-formatting repair.


## Native ATT&CK remediation r2 — 2026-11-06

This entry supplements the preceding history. The four review findings are closed in the current pre-production implementation. Historical source acquisitions, locks, audits, profiles, executables, canonical members, envelopes and reservations remain evidence; there is no legacy reader, conversion, fallback catalog, dual-format admission or database reset. Subsystem major 3, pack contract v2, native ATT&CK v2 and incident-bundle outer format 7 remain unchanged.

### Findings and implementation disposition

| Finding | Remediation and final evidence |
| --- | --- |
| Operator mutation before retained-state admission | `referenceassembly.ValidateRetainedRuntimeState` coordinates read-only Reference Data and Reporting inspectors in one repeatable-read, read-only transaction. Server and operator call it before readiness, leases and mutable storage; migration calls the matching composition under migration exclusion. Actual entrypoint tests cover manifests, sets, portable catalogs/preparations, pending/prepared work, snapshots, render bindings and Composition preview bindings. Every rejection preserves all public database rows/sequences and storage fingerprints, including collectible published objects, abandoned verification scratch and incoming bytes. Operator rejection is canonical JSON plus LF, exit 3, `reference_pack_operation_rejected/contract_incompatible`, with null Job/pack/container attribution. Current state and unavailable current portable references succeed; cancellation/query failures remain operational. |
| Incorrect embedded endpoint category | A closed four-rule native derivation table binds source field, source/target category, predicate and locator extraction. Independent expectations cover all four mappings, retirement, repeated occurrences and cross-domain locators. Eight immutable endpoint recipes include valid mappings and wrong-category/ambiguous-tactic derivatives. Fresh signatures and correct inventories allow negative production verification to reach `content_semantics`, reject with `content_semantic_invalid`, avoid index creation and close scratch. Public test envelopes are retained. |
| Opaque resource-limit diagnostics | Active limits v2 replaces the predecessor catalog, which is retained outside compiled inputs. Native arrays bind 64, 512 and 32768 to exact limit tokens. Line limits come from the profile registry. Trusted compiler validation rejects inconsistent/unknown annotations and unsupported types. Resolved union dispatch uses disjoint kinds or required constant/enum tags; overlapping unions retain exact cardinality semantics. Selected branches retain nested findings; resource excess stops descendants. Object traversal is deterministic, and operational cleanup errors cannot become content verdicts. Tests assert inclusive limits, first excess, safe paths, code/reason, limit IDs, decimal counts and repeatable issue identities. Current other-profile fixtures pass. |
| Duplicate normative declarations | Only the five native additions became RP-REQ-265 through RP-REQ-269. Original RP-REQ-240 through RP-REQ-244 retain their meanings. Tables 26-B/C agree with the native registry; endpoint and admission ordering are explicit. Core 00 now points to NLSpec 0.3.0. Markdownlint RP001 checks top-level declaration syntax only. Its deliberate negative probe produced exactly one duplicate error while ignoring a reference, block quotation and fenced example. The probe was removed from the source tree and retained as documentation evidence. |

### Changed files and retained history

The remediation's authored implementation and test files are:

```text
internal/app/referenceassembly/retained_admission.go
internal/app/operator/operator_reference_pack.go
internal/app/operator/operator_reference_admission_integration_test.go
internal/app/server/runtime_assembly.go
internal/app/server/reference_admission_integration_test.go
internal/app/migrate/migrate.go
internal/app/migrate/reference_admission_integration_test.go
internal/modules/reference_data/contractadmission/admission.go
internal/modules/reference_data/contractadmission/binding.go
internal/modules/reporting/migrationadmission/reference.go
internal/modules/reference_data/internal/packformat/attack.go
internal/modules/reference_data/internal/packformat/attack_native_test.go
internal/modules/reference_data/internal/packformat/attack_remediation_test.go
internal/modules/reference_data/internal/packformat/content.go
internal/modules/reference_data/internal/packformat/content_diagnostics.go
internal/modules/reference_data/internal/packformat/diagnostics.go
internal/modules/reference_data/internal/packformat/diagnostic_safety.go
internal/modules/reference_data/internal/packformat/shape.go
internal/modules/reference_data/internal/packformat/shape_diagnostics.go
internal/modules/reference_data/internal/packformat/shape_limits.go
internal/modules/reference_data/internal/packformat/shape_union.go
internal/modules/reference_data/internal/packformat/tuf_limit_test.go
internal/modules/reference_data/internal/packformat/native_fixture_test.go
internal/modules/reference_data/attack_endpoint_verification_test.go
internal/modules/reference_data/producer/package.go
internal/testutil/referenceadmissionfixture/fixture.go
```

Authored contracts and routing include `contracts/reference-packs/limits.v2.json`, the annotations in `contracts/reference-packs/profiles/framework.attack.native_source.v2.schema.json`, `contracts/reference-pack-producers/attack/source-profile.native.v2.r2.json`, current native fixture descriptors/manifests, and `tools/test_families/{app.operator,app.server,app.migrate,module.reference_data}.json`. `contracts/reference-packs/limits.v1.json` moved byte-for-byte to `fixtures/reference-packs/historical-v1/contracts/limits.v1.json`. Current fixture registration now includes authentic 20.0 and 20.1, reissued malformed/semantic inventories, and eight endpoint fixtures under `fixtures/reference-packs/attack-native-v2/endpoint-types/`. Earlier descriptors and the earlier catalog are retained under `fixtures/reference-packs/attack-native-v2/history/`; no runtime reads them.

Generated contract embeddings and harness projections were refreshed through Make, including `internal/gen/contractreferencepacks/artifacts_gen.go`, `internal/gen/contractreferencepackfixtures/artifacts_gen.go` and applicable harness policy/topology projections. No generated file was hand-edited. Documentation changes are in this handoff, the Reference Pack NLSpec and Core 00. Documentation lint changes are `.markdownlint-cli2.jsonc`, `tools/harness/static-analysis/markdownlint.sh` and `tools/harness/static-analysis/requirement-declarations.cjs`.

Exact full-worktree path inventories are retained at `.cartulary/reference-packs/native-v2-r2/tracked-diff-files.txt` and `untracked-files.txt`. They include the pre-existing uncommitted cutover/acquisition work, not only this remediation. Authored fixture manifests enumerate every exact immutable input path. The newly generated downstream changes were reviewed after finalization.

### Frozen identities and independent witnesses

The active producer accepts only source profile `cartulary.attack.all-domains.native.v2.r2`, SHA-256 `bea57e402fab7d5281e159652f19ffb74a0d4e162be92bf9e8bfdad70d2bbf3c`. That file binds the unchanged v20.0/v20.1 source locks, unchanged field policy, corrected endpoint policy and exact authored/compiled native schema digests. The former native v2 profile remains unchanged at `80d546d402318c2866b999dc2c80ae8e1c663f002c9187b71dfc025f6f1484db`.

Both executable builds in each producer-build invocation were byte-identical. The retained executable is `.cartulary/reference-packs/builders/d369415b683ebb88cd538e0bb2c7ce4e534d501b9125a70190bee80ca691f145/reference-pack-producer`: SHA-256 `d369415b683ebb88cd538e0bb2c7ce4e534d501b9125a70190bee80ca691f145`, 13,126,244 bytes, Go 1.27.3. Its neighboring `identity.json` and `reproducibility.json` retain the independent build results. The second producer-build invocation after final fixture registration reproduced the same bytes.

Assembly inputs are `.cartulary/reference-packs/native-v2-r2/inputs/assemble-20.0.json` and `assemble-20.1.json`, with explicit build time `2026-11-06T23:02:00Z`. Each input was assembled twice unchanged. Immutable retention compares freshly computed members with prior bytes; the independent Node witness separately recomputes JCS, LF framing, digests, exact source-value preservation, notices and counts. `comparison-20.0.json` and `comparison-20.1.json` retain old/new payload comparisons and independent set preimages. All notice/object/relationship bytes are identical to the previous authentic fixtures.

| Release | Fixture version | Fixture sequence | Objects | Relationships | Embedded occurrences | Largest object row including LF | Manifest bytes |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 20.0 | `attack-20.0.4` | 6 | 6,211 | 34,604 | 9,050 | 3,033,785 | 2,572 |
| 20.1 | `attack-20.1.4` | 7 | 6,221 | 34,632 | 9,057 | 3,034,554 | 2,572 |

| Release | Manifest SHA-256 | Payload SHA-256 | Set SHA-256 including unchanged Base |
| --- | --- | --- | --- |
| 20.0 | `6e94d63980e3816b0edb6303da63a6f54387dfd90764b5f6d58efa51caec876f` | `af29e228df3453600b4388969c12c5e4be5f1e5c4f6a45c798207f25cb8e936a` | `7e339b1228cf502ca8fc5b80a741aeb54c9447acd7d19b49193ebab8871fa1c7` |
| 20.1 | `797de8a4133f5bd4b7ff79a26ce57efcdeebc43556033ea3f7ff685004252865` | `c1f6ddf1a73e2544bb2b568ac0135a355f975087c6dac37291ebbad15aa18e4e` | `7a1a24e24b3a2631067806720dacb1255bb21b627b5b6f3c3d52bbbad1643b7e` |

| Release/member | Bytes | SHA-256 |
| --- | ---: | --- |
| 20.0 `notices/MITRE-ATTACK.txt` | 1,354 | `738144f7fb054722a4ef9d3367c51710341dc12fc574c6ac3a41daaaecd8bf5e` |
| 20.0 `payload/objects.ndjson` | 23,037,177 | `78bfdda846326020b0e9e20c5b240d693596ccdb235f7f1f4b15ae28eaa39d07` |
| 20.0 `payload/relationships.ndjson` | 51,947,988 | `7bc3ccbfe63e9425c0f35c0cb05f05d45d13dfaa424f4867dcd8842a10b3e483` |
| 20.1 `notices/MITRE-ATTACK.txt` | 1,354 | `738144f7fb054722a4ef9d3367c51710341dc12fc574c6ac3a41daaaecd8bf5e` |
| 20.1 `payload/objects.ndjson` | 23,061,726 | `22893a502098db91e8aa33618cb48b39ed3c16345651066a9ce065166674bad9` |
| 20.1 `payload/relationships.ndjson` | 51,986,185 | `ee8a5ecacf9094d087ddcc20cabb3ac7196ad04ac41497e58e477bf4ea9b6274` |

The append-only fixture publication ledger retained reservations 1–5 and allocated only 6 and 7 for the two `.4` versions. Replays reused those allocations. Intended-deployment sequences remain unallocated. The primary source baseline remains v20.1, published `2026-11-05T16:35:43Z`, commit `f59b03ecf828579d5f9e93683ce25b7e2e5fd9d7`; the independently pinned v20.0 lock remains unchanged.

Base release `cartulary.base_registries.2026-11-06.v2`, pack version `base-2026-11-06.1`, sequence 2 and Base-only set SHA-256 `2a3420dcc6036316fbb68d5c2a1d42e28cbd1003fe5207a93912ca45e9b993df` are unchanged. All three tuples retain pack contract v2 and content-profile version `1`:

| Pack key | Content-profile ID | Manifest SHA-256 | Payload SHA-256 |
| --- | --- | --- | --- |
| `type_registry.evidence` | `cartulary.reference_pack.type_registry.evidence.v1` | `d446139e2a9f83ff4baf5597a9ec97723a440dc3d138f135836164aca211dd12` | `9a24ec390bf4e9f51ea299f6e08c53f7f8186a66ed2a7decb213a2ad9e901d39` |
| `type_registry.host` | `cartulary.reference_pack.type_registry.host.v1` | `f0c19e9d426077f3ddc16eeba20680a41b80509083d9c59bc9c56305ea7bbca5` | `afaae18a960bf0d7f75365494966403333706ee4b7f680ca789c5cad8a237b8f` |
| `type_registry.indicator` | `cartulary.reference_pack.type_registry.indicator.v1` | `a121e107724bd4f50734f075c997c727fe4b1cf2dd379f06b387f436f8f7c74f` | `1c559a7fd18803b4eb0fbcc53f39be43ff48404d7253dfa0ffd5c62cdb3fecb8` |

### Isolated envelopes and trust

Repository identity remains `fixture.attack.native.v2`, with publisher redistribution policy `restricted`. Private signing material remains host-local, outside fixture containers, bootstrap files and retained public evidence. Public bootstrap SHA-256 is `eeed15d1d99bce09f92745c0f601167122a01fdd0ee57fe9b08a8e9b68d3c7e7`; root threshold/signatures are 2, with one signature for targets, snapshot and timestamp. Root expiry remains `2027-11-06T21:51:53Z`. Explicit envelope inputs record metadata version 1 at `2026-11-06T23:05:00Z`, online expiry `2026-11-13T23:05:00Z`; renewal inputs record version 2 at `2026-11-06T23:12:00Z`, online expiry `2026-11-20T23:12:00Z`. These are explicit publication inputs, not substituted verifier clocks.

Each initial envelope was generated twice from identical inputs and reproduced the same container. Renewals retain the prior envelope, manifest, payload, version and sequence while changing metadata/container identity. Each directory below is under `.cartulary/reference-packs/native-v2-r2/public/envelopes/` and contains `container.pack`, `bootstrap.json`, `envelope-report.json` and an independent signature/ustar witness. Offline witnesses intentionally report `freshness_verified=false`; wall-clock production verification is established separately by the final harness runs.

| Release | Metadata version | Container bytes | Envelope directory / container SHA-256 |
| --- | ---: | ---: | --- |
| 20.1 | 2 | 75,061,248 | `2023cf29744711a6c05ec9010901d5a8e9837ac4692372be87aef46900aefb20` |
| 20.0 | 1 | 74,998,272 | `2f03a3cca0acf0145e214ca55777c115a4c63d55009b1aea4930aa40f065e470` |
| 20.0 | 2 | 74,998,272 | `851def9e569bc1e1c76fcd254ac3bd847d6f5f2032c51f3d259f967d529889f9` |
| 20.1 | 1 | 75,061,248 | `d8bf685356fe4b19feab8846153730883e34b0fe59987e36b7b21203ef2de9e0` |

The final harness runs additionally retain 14 fresh public envelopes, including eight signed endpoint fixtures. Exact roots, container sizes/hashes and bootstrap hashes are recorded in `.cartulary/reference-packs/native-v2-r2/workflow-public-envelopes.json`. Harness-created keys remained in memory. Production verification used current wall-clock time; no Reference Data frozen-clock control was added.

### Verification commands and results

Commands below ran from the repository root with `CARTULARY_TEST_RUN_ID=<Run ID>`. Every final selected invocation exited 0; every declared cleanup step completed; no work unit was skipped. Their shared source digest is `sha256:f06490a46bfe30a33d5640a65df3af6d7cccfbf6184e8f6d33744bb6def0671c`. The source and graph identities, row selection and cleanup results are retained in each `.cartulary/test-results/<Run ID>/` directory. `final-validation-inventory.json` and `final-commands.json` under the r2 evidence root preserve the exact selected commands and summaries. Work-unit counts are not test counts.

| Run ID | Exact command | Outer exit | Duration ms | Cleanup steps |
| --- | --- | ---: | ---: | ---: |
| `attack-r2-final-units-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_diagnostics,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_trust,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c HARNESS_DIAGNOSTICS=basic` | 0 | 42,342 | 12 completed |
| `attack-r2-final-lifecycle-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary,module.reference_data.integration.canonical_persistence HARNESS_DIAGNOSTICS=basic` | 0 | 179,168 | 17 completed |
| `attack-r2-final-reporting-20261106-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c HARNESS_DIAGNOSTICS=basic` | 0 | 220,083 | 14 completed |
| `attack-r2-final-bundles-20261106-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8 HARNESS_DIAGNOSTICS=basic` | 0 | 156,366 | 14 completed |
| `attack-r2-final-composition-20261106-01` | `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.preview_reporting_delegation_ecdd2bc64a HARNESS_DIAGNOSTICS=basic` | 0 | 176,896 | 14 completed |
| `attack-r2-final-operator-20261106-01` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 0 | 40,358 | 14 completed |
| `attack-r2-final-migrate-20261106-01` | `make test-slice OWNER=app.migrate ROWS=app.migrate.integration.reference_pack_retained_admission,app.migrate.unit.migrate_facade_contract HARNESS_DIAGNOSTICS=basic` | 0 | 36,779 | 15 completed |
| `attack-r2-final-server-20261106-01` | `make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission,app.server.support_unit.server_runner_writes_diagnostics_and_returns_fai_1f2addc1fd HARNESS_DIAGNOSTICS=basic` | 0 | 161,718 | 15 completed |
| `attack-r2-final-recovery-20261106-01` | `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.graph_restore_v4_contract_projection,module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c,module.recovery.integration.selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a,module.recovery.integration.restore_readiness_selects_the_latest_retained_su_e71d710085 HARNESS_DIAGNOSTICS=basic` | 0 | 218,978 | 15 completed |
| `attack-r2-final-extensions-20261106-01` | `make test-slice OWNER=module.extensions ROWS=module.extensions.unit.coordinator_registry_1dce9a539a,module.extensions.unit.generated_recovery_composition_preserves_admitte_1e77c4dab8,module.extensions.unit.publication_catalog_exact_profile_contribution_s_15f926e0bd HARNESS_DIAGNOSTICS=basic` | 0 | 161,413 | 12 completed |
| `attack-r2-final-tooling-20261106-01` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.behavior.public_registry_parity HARNESS_DIAGNOSTICS=basic` | 0 | 224,439 | 14 completed |

This portfolio covers canonical profile/schema diagnostics, manifest/trust, producer, fixture registration, consumers and persistence; actual operator/server/migration admission; authentic import, staged-only default, activation, lookups, provenance, replay, renewal, disablement and removal; v20.0→v20.1 snapshot pin/rerender; incident-bundle v7 references-only round trips and restricted embedding refusal; Composition preview; Recovery historical integrity and pre-mutation/readiness rejection; Extensions and changed tooling boundaries. It is selected-boundary evidence, not a complete-owner, CI, deployment or release acceptance claim. No full-owner evidence audit was performed or claimed.

`make task-guide` was consulted for `module.reference_data`, `module.reporting`, `module.incidentbundles`, `module.reportcomposition`, `module.recovery`, `module.extensions`, `app.operator`, `app.server`, `app.migrate` and `harness.command_surface`, all with exit 0. Broader CI/release/browser suites were not run because this portfolio covers the affected backend and contract boundaries. No compatible full warm-check root was supplied to `agent-finalize`, so optional retained-run maintenance was not claimed.

| Run IDs | Exact command | Outer exit / result |
| --- | --- | --- |
| `attack-r2-format-20261106-01` | `make format` | 0 / passed |
| `attack-r2-format-20261106-02` | `make format` | 0 / passed |
| `attack-r2-format-20261106-03` | `make format` | Outer exit not recovered from the native session / passed summary; later format runs have observed exit 0 |
| `attack-r2-format-20261106-04` | `make format` | 0 / passed |
| `attack-r2-format-20261106-05` | `make format` | 0 / passed |
| `attack-r2-format-20261106-06` | `make format` | 0 / passed |
| `attack-r2-format-20261106-07` | `make format` | 0 / passed |
| `attack-r2-generate-20261106-01` | `make generate` | 2 / failed; selector ordering repaired |
| `attack-r2-generate-20261106-02` | `make generate` | 0 / passed |
| `attack-r2-generate-20261106-03` | `make generate` | 0 / passed |
| `attack-r2-generate-20261106-04` | `make generate` | 0 / passed |
| `attack-r2-generate-20261106-05` | `make generate` | 0 / passed |
| `attack-r2-finalize-20261106-01` | `make agent-finalize` | 0 / passed |
| `attack-r2-finalize-20261106-02` | `make agent-finalize` | 0 / passed |
| `attack-r2-builder-20261106-01` | `make reference-pack-producer-build` | 0 / passed |
| `attack-r2-builder-20261106-02` | `make reference-pack-producer-build` | 0 / passed |
| `attack-r2-assemble200-20261106-01` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2-r2/inputs/assemble-20.0.json` | 0 / passed |
| `attack-r2-assemble200-20261106-02` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2-r2/inputs/assemble-20.0.json` | 0 / passed |
| `attack-r2-envelope200-20261106-01` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/native-v2-r2/inputs/envelope-20.0.json` | 0 / passed |
| `attack-r2-envelope200-20261106-02` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/native-v2-r2/inputs/envelope-20.0.json` | 0 / passed |
| `attack-r2-renew200-20261106-01` | `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/native-v2-r2/inputs/renew-20.0.json` | 0 / passed |
| `attack-r2-assemble201-20261106-01` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2-r2/inputs/assemble-20.1.json` | 0 / passed |
| `attack-r2-assemble201-20261106-02` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/native-v2-r2/inputs/assemble-20.1.json` | 0 / passed |
| `attack-r2-envelope201-20261106-01` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/native-v2-r2/inputs/envelope-20.1.json` | 0 / passed |
| `attack-r2-envelope201-20261106-02` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/native-v2-r2/inputs/envelope-20.1.json` | 0 / passed |
| `attack-r2-renew201-20261106-01` | `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/native-v2-r2/inputs/renew-20.1.json` | 0 / passed |
| `attack-r2-final-generated-20261106-01` | `make source-check CHECK=generated` | 0 / passed |
| `attack-r2-final-schema-20261106-01` | `make source-check CHECK=schema` | 0 / passed |

Finalization ran before executable freeze and again after final fixture registration. The final generated/schema checks passed on the shared final source digest. `git diff --check` passed. Final Markdown lint is recorded in the closing entry below.

Earlier focused commands and failures remain retained:

| Run IDs | Exact command | Outer exit / disposition |
| --- | --- | --- |
| `attack-r2-units-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_diagnostics,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_trust HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-r2-units-20261106-02` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_diagnostics,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2 HARNESS_DIAGNOSTICS=basic` | 2 / test_assertion_failure; repaired below |
| `attack-r2-units-20261106-03` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_diagnostics,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2 HARNESS_DIAGNOSTICS=basic` | 2 / test_assertion_failure; repaired below |
| `attack-r2-units-20261106-04` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_profiles HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-r2-operator-20261106-01` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 2 / fixture seed constraints; repaired below |
| `attack-r2-operator-20261106-02` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-r2-operator-20261106-03` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-r2-migrate-20261106-01` | `make test-slice OWNER=app.migrate ROWS=app.migrate.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 2 / fixture seed constraints; repaired below |
| `attack-r2-migrate-20261106-02` | `make test-slice OWNER=app.migrate ROWS=app.migrate.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-r2-server-20261106-01` | `make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 2 / fixture seed constraints; repaired below |
| `attack-r2-server-20261106-02` | `make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-r2-markdown-20261106-01` | `make lint-markdown` | 0 / passed |
| `attack-r2-markdown-negative-20261106-01` | `make lint-markdown` | 2 / intentional negative probe; child 1, harness/unknown_failure classification |

The first generation failure rejected an unsorted test selector; sorting authored routing repaired it. Initial operator/server/migration failures exposed three seed errors: a missing pack-key parent, an attempted update of an immutable completed operation, and absent prepared-member identity. Those test recipes were corrected without weakening database constraints. Unit run 02 exposed a numeric dispatch regression in the bounded compiler and the signed test passing a complete bootstrap catalog to the root-only decoder; dispatch was corrected and the test now uses `DecodeBootstrap`. The remaining run-03 failure used Go `float64` where the production decoder supplies `json.Number`; the test now supplies the actual parsed representation. Current non-ATT&CK content, native diagnostics and signed semantic boundaries all passed afterward and in the final portfolio. These failures remain recorded with their original classifications and cleanup results.

The Markdown negative probe used a temporary document under `docs/research/rp-declaration-probe/`, with exactly two declarations plus a reference, quotation and fenced example. The only finding was RP001 at line 11, referring to line 1. Its retained input is `.cartulary/reference-packs/native-v2-r2/requirement-declaration-probe.md`; the temporary source directory was removed. This check belongs solely to documentation lint. Product code, tests, generators and release calculations do not consume Markdown.

### Acceptance and outstanding gates

| Outcome | Remediation disposition |
| --- | --- |
| Build reproducibility | Passed: repeated executable builds, two canonical assemblies for each release, repeated identical envelope inputs, and unchanged producer after registration. |
| Canonical conformance | Passed for the reissued authentic fixtures and focused negative/boundary portfolio; native payload bytes, notices, counts and digest witnesses are unchanged. |
| Fixture-trust verification | Passed through production verifier/routes/workers with fresh isolated public bootstraps and thresholds 2/1/1/1. |
| Deployment-trust importability | Not established; deployment trust and identities were not changed or allocated. |
| Distribution approval | Not established; fixture policy remains restricted and positive ATT&CK embedding remains gated. |
| Supported E2E completion | Final selected supported workflows passed with completed cleanup and outer exit 0. Relationship-pivot E2E remains separately gated because no relationship consumer API was introduced. |

No remaining implementation blocker was found for the four remediations. The changes are ready for human contract-projection review and commit review within this scope. Human projection fidelity review remains a repository responsibility; these tests do not prove specification completeness. No publication, deployment, commit, push, deployment-root replacement or database reset was performed.

Closing documentation verification: `CARTULARY_TEST_RUN_ID=attack-r2-final-markdown-20261106-01 make lint-markdown` passed with observed outer exit 0. Its summary is `.cartulary/test-results/attack-r2-final-markdown-20261106-01/adhoc/lint-markdown/tool-run-summary.json`. The final read-only `git diff --check` also passed. The executable and both former/current frozen profile hashes were rechecked against the identities above.

## W2 follow-up — binding errors and execution scope — 2026-11-06

The subsequent review found two remaining W2 defects despite the preceding passing selections: binding lookup converted operational database errors into contract incompatibility, and every verification Job repeated the complete retained-state scan before starting its verification budget. This entry supersedes the preceding readiness assessment for those findings and preserves the earlier evidence and conclusions as history.

### Corrections and regression coverage

`contractadmission.validateBinding` now returns loader errors unchanged. Missing or invalid retained set bytes still produce the owner's typed incompatibility error. No error text enters operator output: operational failure remains `internal_error`, exit 3, with null reason, Job, pack and container attribution, encoded as one canonical JSON object followed by LF.

The worker no longer runs global retained-state inspection. Server, operator and migration entrypoints retain their coordinated admission gates. Worker execution starts its verification budget immediately, validates the selected operation's frozen input before recording an attempt, and then processes its admitted cohort through the existing owner pipeline. No admission bypass, compatibility reader, fallback, conversion or new public API was introduced.

New focused tests cover:

- Database failure, cancellation and deadline propagation from binding lookup, while missing and incompatible set content remains a typed rejection.
- A real binding-query failure in operator, server and migration admission. The fixture preserves the columns used by earlier global scans and removes only the column needed to resolve a Reporting binding; assertions require PostgreSQL `42703`, proving the intended boundary was reached. Database rows, migration ledger, lease records, Jobs and storage sentinels remain unchanged. Server rejection preserves the borrowed pool. Operator output is checked for exact sanitized canonical JSON and attribution.
- A claimed import completing through the production worker while unrelated portability catalog/preparation tables are locked. The ten-second context is a deadlock bound: the previous global scan cannot finish while those locks are held. It is not a performance threshold or a substituted verification clock.
- Rejection of incompatible schema and contract identities by the selected-operation attempt path, with no attempt written.

Authored files changed in this follow-up:

```text
internal/modules/reference_data/contractadmission/binding.go
internal/modules/reference_data/contractadmission/binding_test.go
internal/modules/reference_data/coordinator_execution.go
internal/modules/reference_data/coordinator_execution_scope_integration_test.go
internal/testutil/referenceadmissionfixture/fixture.go
internal/app/operator/operator_reference_admission_integration_test.go
internal/app/migrate/reference_admission_integration_test.go
internal/app/server/reference_admission_integration_test.go
tools/test_families/module.reference_data.json
docs/handoffs/mitre-attack-reference-pack.md
```

The two new Reference Data rows are `module.reference_data.unit.retained_binding_admission` and `module.reference_data.integration.execution_admission_scope`. Existing facade rows now include the binding-query case. No normative contract, schema, frozen source profile, fixture registration or producer source changed. Make generation completed; finalization reported `generated.status=unchanged`, zero updated files. Existing generated cutover changes remain part of the uncommitted worktree and were not hand-edited.

### Retained commands and results

Selection used `make task-guide OWNER=module.reference_data`, `OWNER=app.operator`, `OWNER=app.migrate`, `OWNER=app.server`, and `OWNER=harness.command_surface`. Commands below ran from the repository root with `CARTULARY_TEST_RUN_ID` set to the exact listed ID. Run roots are `.cartulary/test-results/<run-id>`.

| Run ID | Exact Make command | Observed outer exit / result |
| --- | --- | --- |
| `attack-w2-format-20261106-01` | `make format` | 0 / passed |
| `attack-w2-generate-20261106-01` | `make generate` | 0 / passed |
| `attack-w2-regression-red-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.retained_binding_admission,module.reference_data.integration.execution_admission_scope HARNESS_DIAGNOSTICS=basic` | 2 / expected pre-fix product failure |
| `attack-w2-format-20261106-02` | `make format` | 0 / passed |
| `attack-w2-focused-reference-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.retained_binding_admission,module.reference_data.integration.execution_admission_scope HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-w2-focused-operator-20261106-01` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-w2-focused-migrate-20261106-01` | `make test-slice OWNER=app.migrate ROWS=app.migrate.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-w2-focused-server-20261106-01` | `make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-w2-finalize-20261106-01` | `make agent-finalize` | 0 / passed |
| `attack-w2-final-lifecycle-20261106-01` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.integration.all_profiles_consumer_boundary HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-w2-final-tooling-20261106-01` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.owner_boundaries,harness.command_surface.behavior.public_registry_parity HARNESS_DIAGNOSTICS=basic` | 0 / passed |
| `attack-w2-final-generated-20261106-01` | `make source-check CHECK=generated` | 0 / passed |
| `attack-w2-final-schema-20261106-01` | `make source-check CHECK=schema` | 0 / passed |
| `attack-w2-builder-20261106-01` | `make reference-pack-producer-build` | 0 / passed; both builds reproduce the frozen executable |

The intentional pre-fix run retained two failed rows classified `product/test_assertion_failure`: lookup errors lost their original identity, and the global scan blocked on unrelated catalog locks until context expiry. Its cleanup completed all 15 steps. Both rows passed after the two production corrections. Two diagnostic `explain-run` attempts exited 2 for a missing or invalid `TARGET`; the corrected `make explain-run RESULTS_DIR=.cartulary/test-results/attack-w2-regression-red-20261106-01 TARGET=test-slice DETAIL=logs` exited 0. These were observation-command usage errors, not additional test failures or reruns.

All passing focused and final graph selections share source digest `sha256:b23a8c056c9af9cca82d5f557d072c7504d0e9a6ea265b14951057bcb386cff7`; the intentional red run used `sha256:e74037fe0eb0c17d668171f70880c7725a26cc5779790eac04daeaaa89f4de1c`. Finalization preserved the passing source digest, so repeating the focused selections afterward was unnecessary. Final selections had zero failed, skipped or cancelled units; all reported cleanup steps completed. Lifecycle coverage adds authentic import/renewal and canonical persistence to the new regressions; tooling coverage checks ownership and public routing for the new rows.

Exact commands, observed outer exits, summary paths, source identities and cleanup dispositions are indexed in `.cartulary/reference-packs/w2-followup/validation-inventory.json`. Graph summaries remain the owning `run-summary.json`; non-graph generation/build and closing Markdown commands retain their tool summaries. The index is an inspection aid, not a replacement harness event reducer or complete-owner evidence audit.

### Preserved artifact identities and trust

The producer's two builds remain byte-identical to the prior retained executable: SHA-256 `d369415b683ebb88cd538e0bb2c7ce4e534d501b9125a70190bee80ca691f145`, 13,126,244 bytes. The current source profile remains `cartulary.attack.all-domains.native.v2.r2`, SHA-256 `bea57e402fab7d5281e159652f19ffb74a0d4e162be92bf9e8bfdad70d2bbf3c`. Consequently this application-admission correction requires no source-profile, canonical pack or publication reissue.

Read-only file inspection confirmed the current manifests and every notice/payload member match the retained catalog. The exact hashes and sizes are retained in `.cartulary/reference-packs/w2-followup/artifact-identities.json`, alongside source-lock/profile, Base witness, ledger and prior public-envelope identities.

| Release | Fixture version / sequence | Objects | Relationships | Largest object row including LF |
| --- | --- | ---: | ---: | ---: |
| 20.0 | `attack-20.0.4` / 6 | 6,211 | 34,604 | 3,033,785 |
| 20.1 | `attack-20.1.4` / 7 | 6,221 | 34,632 | 3,034,554 |

All manifest, payload and set hashes in the preceding r2 tables are unchanged. The three Base v2 tuples and Base-only set `2a3420dcc6036316fbb68d5c2a1d42e28cbd1003fe5207a93912ca45e9b993df` remain unchanged. The fixture ledger still contains exactly reservations 1–7; no publication sequence was allocated. Both authentic source acquisitions, audits, previous executables, canonical fixtures and four prior r2 public envelopes remain retained.

The final lifecycle run retained two fresh public envelopes and matching public bootstraps under its own suite artifact root. Exact paths, container sizes/hashes and bootstrap hashes are listed in `.cartulary/reference-packs/w2-followup/workflow-public-envelopes.json`. Fixture trust remains isolated under `fixture.attack.native.v2`, redistribution remains `restricted`, and thresholds remain two root signatures plus one signature for each online role. The existing production lifecycle checks use fresh wall-clock verification; private fixture keys remain in memory and no signing material was copied into evidence.

### Commit-readiness review and limits

Both W2 findings are closed by the production changes and passing regressions. Review confirmed the global inspector remains at server/operator/migration admission, operational loader errors retain their identity, invalid retained content still fails closed, and execution preserves selected-input validation. No further specification or implementation blocker was found in this follow-up scope. The changes are ready to commit after the closing documentation check; no commit or push was performed.

| Outcome | Follow-up assessment |
| --- | --- |
| Build reproducibility | Passed again: two fresh builds reproduce the exact frozen executable. Prior logical/container reproduction evidence remains retained for unchanged producer and canonical bytes. |
| Canonical conformance | Canonical identities remain unchanged; current lifecycle and persistence selections passed. Prior native semantic/limit evidence remains identified with its original source digest. |
| Fixture-trust verification | Current authentic production lifecycle passed with fresh isolated envelopes and completed cleanup. |
| Deployment-trust importability | Not established; deployment identities and trust remain untouched. |
| Distribution approval | Not established; restricted policy and positive ATT&CK embedding gate remain. |
| Supported E2E completion | The affected admission and Reference Data lifecycle selections passed. Prior Reporting, Composition, incident-bundle and Recovery evidence remains historical; this follow-up does not claim a new complete-owner or complete-release audit. Relationship-pivot E2E remains separately gated. |

Full CI, browser suites, release checks and a complete-owner evidence audit were not run: this correction changes application admission and worker scope, covered by the selected facade/owner regressions, lifecycle paths and routing checks. The earlier consumer workflow runs were preserved rather than presented as fresh executions. `agent-finalize` retained-run maintenance was skipped because no compatible full warm-check root was supplied; no redundant full check was launched to create one. Canonical reassembly and envelope reissuance were unnecessary because the producer, profile, canonical bytes and publication identities are unchanged.

Closing documentation command: `CARTULARY_TEST_RUN_ID=attack-w2-final-markdown-20261106-02 make lint-markdown`. Its owning summary is `.cartulary/test-results/attack-w2-final-markdown-20261106-02/adhoc/lint-markdown/tool-run-summary.json`; the observed result is recorded in the follow-up validation inventory. Run 01 also passed; run 02 checks the final wording distinguishing frozen-input validation before attempt creation from subsequent cohort processing. Final review also uses `git diff --check`. No publishing, deployment, legacy support, deployment-root replacement, database reset, commit or push is part of this follow-up.
