# Reference Pack cutover owner and artifact inventory

This is an implementation handoff for the uncommitted remediation based on `78a4effc23aac4b46ca414e11b91e17892f790c7`. It is not an adoption record. The Reference Pack owner is draft version **0.2.0**, profile major **2**. Companion edits in this working tree require coordinated owner review; an existing document's adopted header does not establish adoption of these edits. Exact companion file revisions are recorded in Reference Pack Table 4-C. These references identify this reviewed pre-production amendment set; they do not assert separate approving authorities or production adoption. The project owner explicitly excluded independent recreation and interchangeability reports from pre-production acceptance on 2026-10-04. Tests and release tooling must not consume this document or derive facts from Markdown.

## Owner revisions and projections

The Reference Pack 0.2.0 source for this pre-production cutover has SHA-256 `619f36747cd86fa438d8f0e5c054ba2776b0ee20e7822f1106adfa08dfd8fa26`. Its eleven exact companion content revisions are in Table 4-C. These are editorial revision references; runtime, generation and test evidence remain derived exclusively from authored machine contracts.

The coordinated pre-production implementation passed the complete release gate, 1259/1259 units, at `.cartulary/test-results/20261004T063630Z-p80139`. The run manifest records source digest `sha256:a0078d33960af5195218b0d56d7e856c5fe739e7034126596b2e24330efc2785`. The [implementation ledger](reference-pack-remediation.md) records the final validation, phase exits and historical failures; the statuses below continue to distinguish content revision identity from formal production adoption.

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
