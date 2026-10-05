# Reference Pack cutover owner and artifact inventory

This inventories the S01–S08 pre-production remediation on baseline `c9f5b366fe0d295057a30aaf6e488af9c587fdcb`. The [controlling tracker](reference-pack-remediation.md) owns execution status and final S10 acceptance. The Reference Pack owner remains **draft 0.2.0**, profile major **2**. Existing adopted headers do not adopt the coordinated candidate amendments. Human content references are separate from executable contracts; no test, generator, runtime or release evidence may read this document. Formal adoption, validated cryptographic build, target security approval and deployment rehearsal remain production gates.

## Owner revisions and projections

The S08 Reference Pack 0.2.0 source has editorial SHA-256 `68e16e49d820662796af41054bf2cab221b024a4de9c0e753076f491941874d8`. Its eleven exact companion content revisions are in Table 4-C. These are editorial revision references; runtime, generation and test evidence remain derived exclusively from authored machine contracts.

The historical cutover record reports a complete release gate, 1259/1259 units, at `.cartulary/test-results/20261004T063630Z-p80139`. The run manifest records source digest `sha256:a0078d33960af5195218b0d56d7e856c5fe739e7034126596b2e24330efc2785`. Equivalence of that dirty source to `54f4219` was not established. S06 freshly passed all 23 Reference Data execution units at `20261005T024203Z-p93991`; S10 qualified the final candidate with release-check 1261/1261 at `20261005T030730Z-p45700` and test-fast 739/739 at `20261005T030116Z-p50484`, both on source `sha256:b5d0f1c25905104352a03faae5c571510902d31ac89b4b16d1632dabe671d6ae`. The [implementation ledger](reference-pack-remediation.md) records superseding results; the statuses below continue to distinguish content revision identity from formal production adoption.

| Owner / gate | Candidate amendment | Authored machine projection | Adoption disposition |
| --- | --- | --- | --- |
| Reference Pack / corpus | `docs/reference-pack-subsystem-nlspec.md`, document 0.2.0, profile 2 | `contracts/reference-packs/`; `contracts/extensions/fragments/core01.profile-classifications.json` | Draft; pre-production implementation acceptance uses §32. Formal production adoption remains separate. |
| Core 00 / RP-GATE-001 | Recognition, authority and conformance dependency | `contracts/extensions/fragments/core00.recognition.json`; `contracts/verification/owners/module.reference_data.json` | Major-2 recognition and bounded subsystem ownership reconciled for pre-production; formal adoption unclaimed. |
| Core 01 / RP-GATE-002, 003 | Required Base independently of operator routes; health/disablement; removal; exact-set binding; closed errors | `contracts/openapi-source/owners/module.reference_data/openapi.json`; Reference Pack administrative and consumer projections | Pre-production amendment; exact revision in Table 4-C; formal adoption unclaimed. |
| Core 02 / RP-GATE-004 | Immutable content/envelopes/sets/provenance; no mutable overrides | Reference Pack manifest, envelope, operation and set schemas; migrations 46–63; authored owner SQL | Pre-production amendment; exact revision in Table 4-C; formal adoption unclaimed. |
| Core 04 / RP-GATE-005 | Offline trust, clock assertion, inclusive timeout, backup integrity | `contracts/reference-packs/configuration.v2.schema.json`; `limits.v1.json`; Extensions configuration | Pre-production amendment; exact revision in Table 4-C; formal adoption unclaimed. |
| Extensions / RP-GATE-013, 016 | Document 0.12.0 over adopted 0.11.0; configuration namespace; attributed Job identity; shared finalization | `contracts/extensions/profiles/reference_pack/configuration.json`; `fragments/core01.profile-jobs.json`; `specification/route-identity.v2.schema.json` | Pre-production cutover edits; exact revision in Table 4-C; formal adoption unclaimed. |
| Reporting / RP-GATE-006 | Proposed document 1.4.0 over adopted 1.3.0; exact snapshot binding; frozen render identities/times | `contracts/reporting/`; Reporting owner codecs and migration 63 | Explicit `proposed-amendment`; exact revision in Table 4-C; formal adoption unclaimed. |
| Report Composition / companion to RP-GATE-006 | Proposed document 1.3.0 over adopted 1.2.0; real preview identity and admission time | Reporting render identity projection; companion persisted preview owner | Explicit `proposed-amendment`; exact revision in Table 4-C; formal adoption unclaimed. |
| Incident portability, Core 01 §12.3 / RP-GATE-014 | Incident-bundle format 5; closed references; destination verification; inactive import | `contracts/reference-packs/reference_pack_refs.v1.schema.json`; portability schemas; `tools/schemas/cartulary.incident_bundle_source_catalog.v5.schema.json` | Pre-production amendment; exact revision in Table 4-C; formal adoption unclaimed. |
| Recovery, Core 01/Core 04 / RP-GATE-015 | Retained-reference inventory; historical-time verification; required-loss readiness | `contracts/recovery/`; Reference Data and Reporting recovery contributions | Pre-production amendment; exact revision in Table 4-C; formal adoption unclaimed. |
| OpenTelemetry / RP-GATE-007 | Closed operations, results, spans and histogram attributes | `internal/platform/telemetry/`; `internal/app/referenceassembly/telemetry.go`; telemetry owner tests | Pre-production amendment; exact revision in Table 4-C; formal adoption unclaimed. |
| Testing Harness / RP-GATE-008 | Closed fixture manifests, rooted inputs, separate projection consistency and executed evidence | `tools/schemas/cartulary.reference_pack_fixture_manifest.v1.schema.json`; `tools/harness_schema_attachments.json`; `tools/test_families/module.reference_data.json` | Pre-production amendment; exact revision in Table 4-C; formal adoption unclaimed. |
| Domain / RP-GATE-009 | Candidate, envelope, logical version, set, pin and immutable provenance vocabulary | No runtime projection is derived from vocabulary prose | Vocabulary reviewed for consistent candidate/envelope/set/pin distinctions; exact revision in Table 4-C. |
| Security / RP-GATE-010 | Ed25519-only offline profile under TUF 1.0.35 | TUF role/signature schemas, check registry and independent signed vectors | Implementing-agent review exists; target production compliance approval is not asserted. Pre-production verification does not claim FIPS validation. |
| Licensing / RP-GATE-011 | Original Base assets and synthetic optional fixtures only | `contracts/reference-packs/builtins/release.v1.json`; `builtins/source-profile.v1.json`; synthetic profile fixtures | Distribution inventory below; no external dataset redistribution approval is inferred. |

## S08 gate dispositions and responsible authorities

“Reviewed” below means candidate text, typed projection and the behavioral evidence index agree within the stated scope. It is not an approving authority signature. S10 supplies passing integrated execution acceptance; it cannot close production adoption.

| Gate | Candidate disposition | Responsible authority and remaining production action |
| --- | --- | --- |
| RP-GATE-001 | Reviewed bounded Core recognition; no executable document provenance. | Core 00 owner coordinates formal recognition/adoption. |
| RP-GATE-002 | Reviewed Base, import/renewal, disablement, fallback and immutable sets. | Core 01 lifecycle owner adopts coordinated semantics. |
| RP-GATE-003 | Reviewed routes, complete signer sets, closed errors and removal. Stale generic/error and acceptance copies corrected in S08. | Core 01 transport owner approves the existing major-2 contract. |
| RP-GATE-004 | Reviewed immutable envelopes, attempts, sets, pins and no mutable overrides. | Core 02 persistence owner adopts minima; applied migrations remain unchanged. |
| RP-GATE-005 | Reviewed roots, limits, trusted-clock assertion, no-egress and historical restore. | Core 04 owner adopts policy/configuration; deployment protects storage. |
| RP-GATE-006 | Reviewed exact snapshot/set binding and distinct release/preview identity, bytes and admission times. | Reporting and Composition owners adopt proposed 1.4.0/1.3.0 together. |
| RP-GATE-007 | Reviewed eleven operations, closed result classes, span/histogram shape and privacy. | OpenTelemetry owner adopts companion registration. |
| RP-GATE-008 | Reviewed machine-only generation, 27 fixture families, routing and executable evidence separation. | Harness owner adopts companion obligations; S10 records the fresh release. |
| RP-GATE-009 | Reviewed every Reference Data vocabulary occurrence, mandatory Base, optional administration and Reporting-owned templates. | Domain owner accepts editorial vocabulary navigation. |
| RP-GATE-010 | Implementer review refreshed; ordinary strict Ed25519 supported, FIPS unsupported. | Target security authority selects permitted posture. S07 remains DEFERRED. |
| RP-GATE-011 | Current distribution is three original Apache-2.0 Base packs; optional profile data is synthetic test material. | Distribution/licensing authority approves any future external corpus and target packaging. |
| RP-GATE-012 | All 264 requirement bodies and 54 criteria reviewed, including sub-obligations and unreachable-limit arguments; machine counts alone are insufficient. | Reference Pack/corpus owner approves candidate fidelity; S10 integrated evidence passed; production adoption remains distinct. |
| RP-GATE-013 | Reviewed profile major 2, configuration namespace, empty capability arrays and common finality. | Extensions owner adopts coordinated amendment over its retained authority. |
| RP-GATE-014 | Reviewed format 5 exact catalogs, independent destination trust, optional/required outcomes and pins. | Core 01 portability owner adopts format and degradation rules. |
| RP-GATE-015 | Reviewed retained history, successful-time verification, independent roots and restore readiness. | Core 01/Core 04 recovery owners approve catalog and deployment rehearsal prerequisites. |
| RP-GATE-016 | Reviewed typed local operator actor, shared Jobs admission/worker, no listener or second executor. | Core Jobs/Extensions owners approve attribution and operator contract. |

S08 clarifies producer-only cross-version framework/enrichment continuity: runtime never certifies source subject meaning or infers a global subject registry. Type-registry replacement remains enforced. Exact version-bound objects, generic unknown framework types and pinned history preserve safe consumer behavior. No executable identifier, canonical bytes, database schema or public response changes follow from this clarification.

## Version changes

Application compatibility major, profile major, document versions, schema IDs and route versions have separate meanings. `/api/v1` remains the route namespace. Public breaking changes are recorded in `contracts/openapi-releases/2.0.0.change-set.json`; the historical 1.0.0 baseline is retained.

| Contract | Cutover representation |
| --- | --- |
| Reference Pack extension profile | Major 2; capability arrays remain empty. |
| Reference Pack configuration | `reference_pack.configuration.v2`, `cartulary.reference_pack.configuration_namespace.v2`; shared configuration contract v4 includes explicit `configuration_namespace`. |
| Persisted Reference Pack lifecycle | Import/reverify/refresh Job kinds v2; frozen operation input v2; no retired Job-codec fallback. |
| Administrative resource | `cartulary.reference_pack_administrative_version.v2`; successful-only members are explicit nulls before success. |
| Consumer operations | Five request/result schema pairs and `cartulary.reference_pack_consumer_operations.v1`; closed error/evaluation/item unions with explicit null result arms and ordered request checks. These are corrected, never-adopted projections. |
| Canonical packs, algorithms, sets and trust | Corrected never-adopted `.v1` identifiers retained. TUF dependency is pinned to 1.0.35. |
| Incident bundle | Format 5 only; `reference_pack_refs.v1` contains required sets and versions arrays. |
| Reporting snapshots | `cartulary.reporting_snapshot_model.v1`; snapshot ID exists before nested source references are constructed. |
| Reporting rendered models | `cartulary.export_model.v2`, `cartulary.redacted_export_model.v2`, `cartulary.render_bundle_manifest.v2`; exact snapshot/set binding and admitted render identity. |
| Changed render artifacts | Redaction manifest, Reporting token manifest, Reporting token reveal map and slide deck each advance independently to v2. Preview ID domains are distinct from release ID domains. |
| Recovery | Current recovery catalog v2, PostgreSQL snapshot artifact v2, backup integrity manifest v3, target marker v3, restore verification v3, journal payload v4 and graph restore binding v5; exact current generation selected by `contracts/recovery/fixtures/recovery-generation-registry.v1.json`. |

## Distribution and licensing inventory

Only three packs are packaged as Base. Their source is original project vocabulary, declared `project_owned` in source profile `cartulary.base_registries.source.v1`, source version 1, built at `2026-10-02T00:00:00Z`, under Apache-2.0. The release input includes canonical manifests, the license notice, complete payload members and exact bindings. Host and Evidence each supply the required `unknown` rough-capture entry. Their current source owners expose no assigned custom registry-token field. Indicator supplies exactly the nine adopted Core identities; no optional upstream dataset is included.

The sixteen profile acceptance datasets are synthetic test inputs. Names such as ATT&CK, KEV and LOLBAS identify the implemented content profiles; they do not mean an upstream corpus was copied into the release. Cryptographic vectors contain public test roots and signatures and are marked synthetic. Test keys must never be installed as deployment trust.

Runtime acceptance of an operator pack does not grant distribution rights. Every future project-distributed external corpus still requires stable source identity, source artifact digests, approved redistribution classification and notices before packaging. That approval cannot be inferred from an SPDX parser accepting an expression.

## Persistence, reset and recovery

Appended migrations 46–63 add immutable Reference Data state, snapshot bindings, identity preflight, audit and diagnostic evidence, actor attribution, dependency captures, portability preparation/retention, source Reporting artifacts, fallback attribution and admitted Reporting model identity. Migration 58 removes only empty retired mutable tables. Applied historical migration bytes are unchanged. The source catalog hash at migration 63 is `9052cea15514ace7ede6cbc4d214c585c27c3ced742595b7b9bfd5269e6176f3`.

Follow the [development cutover procedure](../guides/reference-pack-development-cutover.md). Preflight inventories incompatible retained packs, Jobs, incident identities and historical artifacts before mutation and reports category counts. There is no trust reconstruction, snapshot backfill or identity rewrite. The procedure was rehearsed twice only in the isolated tmpfs-backed database recorded in the [implementation ledger](reference-pack-remediation.md); migration head 63 and cleanup were verified. Preserve incompatible backups with their matching historical application. Resetting does not convert them.

Retired behavior includes the checksum-labelled signature verifier, legacy pack reader, mutable version upserts, empty built-in generation, independent indicator canonicalizers, unversioned portability references, retired lifecycle codecs and render-time reconstruction of snapshot/release identity. None is promised compatibility by this prerelease cutover.

## Resuming the deferred production slices

S07 requires a named target deployment and a recorded security-authority decision about validated cryptography. Keep the current unsupported-FIPS posture until that decision exists. If it requires a different posture, coordinate the Reference Pack and Core 04 amendment, select the exact verification-method identity and pinned module/build, test the actual verifier and whole process, and enforce the selected posture at startup. Record the target authority's disposition and its environment limits. Ordinary Ed25519 tests and this implementer review cannot substitute for that decision.

S09 requires the coordinated owner adoptions, deployment security and distribution approvals, and an identified disposable target with explicit authorization for destructive initialization. Preserve any incompatible retained deployment with its matching application and backup; preflight must reject it without mutation. Record immutable image/build/configuration identities, admitted root capabilities and explicit operator trust/clock inputs before rehearsal. Exercise fresh initialization, explicit import and activation, restart, root rotation, exact snapshot/report binding, destination-trusted portability, historical restore and collection. Confirm source roots remain untouched and account for all target cleanup. The deployment owner and recovery/security authorities accept that environment-specific evidence; a passing local release run does not prove it.

Resume each slice by changing its own controlling-tracker row from DEFERRED to IN_PROGRESS only after those prerequisites are supplied. Record concrete commands, identities, outcomes, unresolved findings and approval references. Do not infer historical trust, rewrite canonical identities, add a converter or introduce dual writers as a shortcut to either gate.
