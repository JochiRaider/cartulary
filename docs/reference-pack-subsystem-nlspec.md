---
title: Cartulary Reference Pack Subsystem NLSpec
status: draft
document_class: nlspec
document_version: 0.2.0
contract_major: 2
profile_id: reference_pack
schema_id: cartulary.reference_pack_subsystem_nlspec.v1
---

# 1. Status, scope, authority, and promotion conditions

This NLSpec defines the Cartulary Reference Pack subsystem. It specializes the existing `reference_pack` extension profile and defines the limited Base Profile contract for immutable built-in type registries. It does not create a second extension identity.[^7]

The document remains `status: draft` until the coordinated owner requirements in Table 1-A are satisfied through the existing document-status process. A draft implementation may use this document for development, but it must not claim Reference Pack Extension Profile conformance while any required promotion condition remains open. Pre-production implementation acceptance is distinct from that promotion: it requires the behavioral, projection and executable evidence in §32, without a second implementation or an independent recreation/interchangeability report. This distinction reflects the project owner's explicit pre-production scope decision on 2026-10-04.

Contract major 2 replaces the adopted Reference Pack administration contract. The canonical pack, content-profile, and algorithm identifiers introduced here have never been adopted and retain their `.v1` identifiers for first adoption. Application compatibility majors and the `/api/v1` route namespace are separate contracts and are not incremented by this document. Retired pack bytes are rejected; no compatibility reader or dual-write format is defined. Incident portability's changed reference representation requires incident-bundle format 5. A migration MUST reject incompatible retained state before mutation rather than invent trust, silently delete incident data, or rewrite indicator identities.

**RP-REQ-001**
This NLSpec MUST own only these behavior families:

- canonical reference-pack container and logical-member behavior;
- canonical pack manifest, canonical JSON, digest, and signature behavior;
- packaged built-in registry snapshots;
- operator-imported pack verification through the Cartulary offline TUF profile;
- content-profile and entry schemas;
- pack-version immutability, compatibility, dependencies, conflicts, and deterministic validation;
- immutable active reference-pack-set construction and publication;
- language-neutral pack consumer operations;
- pack-side snapshot, reporting, portability, backup, and restore consequences;
- reference-pack-specific conformance fixtures and acceptance criteria.

**RP-REQ-002**
This NLSpec MUST NOT redefine:

- extension-profile recognition or claimability;
- public HTTP success and error envelopes;
- common job-resource behavior;
- route-scoped idempotency mechanics;
- `deployment_admin` derivation;
- incident authorization;
- workbook source-record mutation;
- Base Profile workbook surface identity;
- report rendering or release approval;
- OpenTelemetry signal mechanics;
- Testing Harness command mechanics.

Those behaviors remain owned by Core 00 through Core 04 and their adopted subsystem NLSpecs.

**RP-REQ-003**
When this NLSpec conflicts with Core 00 through Core 04 outside the Reference Pack subsystem, the conflict is a defect in this NLSpec. When it conflicts with a non-normative appendix, research report, design guide, implementation guide, or example, this NLSpec governs only after adoption and only within the boundary in Table 1-B.

**RP-REQ-004**
The Base Profile imports only these parts of this NLSpec:

- content schemas for `type_registry.host`, `type_registry.evidence`, and `type_registry.indicator`;
- packaged built-in verification;
- required built-in registry availability;
- the internal registry resolver;
- immutable reference-pack-set identity for Base registry consumption.

External import, activation, disablement, refresh, removal, trust-repository processing, and deployment administration are available only when the `reference_pack` extension profile is claimed.

**RP-REQ-005**
The Reference Pack Extension Profile MUST remain a deployment-scoped administrative subsystem. Reference packs MUST remain incident-external state. A reference pack MUST NOT become an incident record, record envelope, saved view, `view_schema`, workbook tab, extension workspace, or incident-specific ACL object.

**Table 1-A. Coordinated owner requirements**

| Gate ID | Owner | Required adopted state |
| --- | --- | --- |
| `RP-GATE-001` | Core 00 | Core 00 lists this NLSpec as adopted for the exact boundary in Table 1-B and makes the `reference_pack` claim depend on it. |
| `RP-GATE-002` | Core 01 | Core 01 distinguishes packaged Base registries from optional pack administration and imports this NLSpec's lifecycle specialization. |
| `RP-GATE-003` | Core 01 | Core 01 adds the removal route, required resource fields, public reason-code additions, refresh semantics, and reference-pack-set snapshot/release bindings named by this NLSpec. |
| `RP-GATE-004` | Core 02 | Core 02 replaces undefined mutable local overrides with immutable signed replacement packs and adopts the persistence minima in §24. |
| `RP-GATE-005` | Core 04 | Core 04 adopts the trust-bootstrap key, verification timeout key, security rules, and conformance mapping in this NLSpec. |
| `RP-GATE-006` | Reporting and Report Composition Subsystem NLSpecs | Reporting binds every pack-sensitive derivation to the snapshot's exact `reference_pack_set_id`; Report Composition supplies the exact admitted preview identity and time. |
| `RP-GATE-007` | OpenTelemetry NLSpec | The Reference Pack operation and safe-attribute registry in §27 is adopted by the telemetry owner. |
| `RP-GATE-008` | Testing Harness NLSpec | Generated schemas, canonical fixtures, malicious-container fixtures, cryptographic fixtures, and drift checks in §29 are adopted by the harness owner. |
| `RP-GATE-009` | `domain.md` | The concepts in Table 5-A are added without redefining this NLSpec's behavior. |
| `RP-GATE-010` | Security authority | The security authority confirms that the Ed25519-only v1 profile is permitted for the target compliance posture. |
| `RP-GATE-011` | Licensing authority | Every project-distributed external-data pack has an approved redistribution classification and required notices. |
| `RP-GATE-012` | Corpus maintenance | Required behavioral families and every supported profile have executed fixture coverage under §29; ordinary owner review resolves normative placeholders and contradictions. Counts and identifier mappings do not establish completeness. |
| `RP-GATE-013` | Extensions Subsystem NLSpec | Profile major 2, configuration v2 and lifecycle Jobs v2 are adopted with shared transaction/deadline finalization and empty capability arrays. |
| `RP-GATE-014` | Core 01 incident portability owner | Incident-bundle format 5, exact references, destination trust, inactive import, required-member failure and licensing rules are adopted together. |
| `RP-GATE-015` | Core 01/Core 04 recovery owners | Current recovery contributions retain content, envelopes, trust history, immutable provenance and pins; historical-time verification and required-loss readiness are adopted. |
| `RP-GATE-016` | Core 01 Jobs and Extensions owners | Local operator admission has an explicit non-user actor and operation identity in the common Job, durable admission evidence, terminal finalization and recovery contracts; no fabricated user or second verification pipeline is permitted. The operator waits on the same admitted lifecycle Job without creating a listener. |

**Table 1-B. Owner boundary**

| Contract family | Primary owner | This NLSpec's rule |
| --- | --- | --- |
| Profile recognition, claimability, precedence | Core 00 | Imported, not redefined. |
| Public route inventory, HTTP status, common envelopes, idempotency, jobs, public resource projection | Core 01 | This NLSpec defines semantic operation effects and required Core amendments only. |
| Pack version conditions and outer activation shell | Core 01 | Imports the existing condition vocabulary and specializes legal effects. |
| Pack bytes, manifest, canonicalization, digests, trust, content profiles, verification, dependencies, active set | This NLSpec | Defined completely here. |
| Pack persistence minima | Core 02 | This NLSpec defines logical facts that Core 02 must require; physical schema remains implementation latitude. |
| Authorization, deployment configuration, roots, hostile-content boundary | Core 04 | Imported or amended through the promotion gates. |
| Workbook surfaces | Core 01 and Core 03 | No new surface is created by this NLSpec. |
| Snapshot and release tuple fields | Core 01 | This NLSpec defines the pack-set object that those tuples reference. |
| Reporting consumption | Reporting Subsystem NLSpec | Reporting consumes, but does not redefine, the exact pack set. |
| Telemetry | OpenTelemetry NLSpec | This NLSpec supplies operation semantics only. |
| Harness mechanics | Testing Harness NLSpec | This NLSpec supplies fixture and acceptance obligations only. |

# 2. Normative language and document discipline

**RP-REQ-006**
The key words **MUST**, **MUST NOT**, and **MAY** are normative in this NLSpec. **MUST** and **MUST NOT** define conformance requirements. **MAY** defines optional behavior only when omission behavior is stated in the same requirement, table row, or immediately following paragraph.

**RP-REQ-007**
The word `default` defines the required value or behavior when omission is valid. A default is an observable conformance requirement.

**RP-REQ-008**
Object member names, schema identifiers, algorithm identifiers, profile identifiers, pack keys, route path values, error codes, reason codes, lifecycle tokens, and closed-vocabulary values MUST be compared as exact Unicode code point sequences after decoding unless this NLSpec names an exact normalization algorithm.

**RP-REQ-009**
A canonical or public object MUST reject unknown members unless this NLSpec explicitly admits them in the TUF `signed` extension policy in Table 10-A. The presence of an `extensions` member does not permit undeclared members inside it.

Unless a member table explicitly says otherwise, every named member at every nesting level is required and non-null. A required nullable member MUST be present and use explicit `null` when absent semantically. Producer defaults MUST be materialized before signing; consumers MUST NOT repair omitted signed members. The TUF `signed` extension policy in Table 10-A is the only unknown-member exception.

**RP-REQ-010**
For every optional member, this NLSpec MUST define omission behavior and explicit JSON `null` behavior separately. An omitted member and an explicit `null` MUST NOT be treated as equivalent unless the owning requirement says so.

**RP-REQ-011**
Every set-like array MUST define duplicate handling and canonical ordering. Every bounded scalar or collection MUST define its minimum, maximum, equality-at-boundary behavior, and failure result.

**RP-REQ-012**
A requirement or table MUST NOT delegate observable behavior through phrases such as `when available`, `as appropriate`, `equivalent`, `latest`, `implementation-defined`, or `trusted source` unless the same requirement names the exact owner, schema, algorithm, comparison rule, or future-only omission behavior that makes the statement decidable.

**RP-REQ-013**
An acceptance criterion MUST verify behavior already defined by a requirement. An acceptance criterion MUST NOT introduce a new default, interface member, state transition, or algorithm. This document applies the NLSpec completeness, interface, default, mapping, and binary-acceptance discipline defined by the repository's NLSpec grounding artifact.[^8]

# 3. Purpose and non-goals

**RP-REQ-014**
The Reference Pack subsystem MUST provide deterministic, locally verifiable, versioned registry, framework, and enrichment data without requiring internet access and without changing incident source authority.

**RP-REQ-015**
Reference-pack administration MUST remain outside the workbook capture hot path. Import, verification, reverify, and refresh MUST execute as background jobs. Their UI state MUST distinguish staged, verified, active, disabled, failed, and missing conditions without blocking timeline capture, entity resolution, evidence attachment, or core editing.[^1]

**RP-REQ-016**
Framework and enrichment data MUST be advisory or reference-only. It MUST NOT create, update, delete, merge, supersede, or change lifecycle state on incident records automatically.[^9]

**Table 3-A. Non-goals and required omission behavior**

| Non-goal | Required omission behavior |
| --- | --- |
| Executable plugin system | Pack-provided code, scripts, native libraries, macros, install hooks, and executable transforms are rejected. |
| Live update client | No operation downloads, discovers, or polls remote pack content. |
| Pack-defined workbook surface | No `view_schema`, system view, built-in tab, saved view, or extension workspace is created. |
| Forms-first capture dependency | Pack absence or verification work does not block ordinary workbook capture. |
| Automatic authoritative enrichment | Hits remain advisory and do not mutate incident source records. |
| Mutable deployment-local override layer | No override resource, precedence stack, or unversioned local mutation exists in v1. |
| Version-range dependency | Only exact dependency tuples are valid. |
| Generic template-pack behavior | Reporting template packs remain outside this NLSpec. |
| Browser trust-root administration | No browser route creates, replaces, or removes a trust root. |
| Raw upstream fetch during verification | Verification reads only admitted local bytes and configured local trust state. |
| Arbitrary source transform execution | Runtime pack consumption never executes a source-profile transform. |
| Cross-incident state | Pack state is deployment-scoped, not incident-scoped. |
| Pack hard purge without tombstone | Removal preserves identity, digests, provenance, and attestations. |

# 4. Versioning and normative dependencies

**RP-REQ-017**
This document uses semantic document versioning only for the NLSpec artifact. `document_version` MUST NOT be used as a pack version, content-profile version, source-profile version, or public API version.

**Table 4-A. Version dimensions**

| Dimension | Meaning | Ordering semantics |
| --- | --- | --- |
| `document_version` | Version of this NLSpec | Semantic document version. |
| `contract_major` | Breaking compatibility family of this NLSpec | Integer equality only at runtime discovery. |
| `pack_contract_version` | Pack envelope, manifest, trust, and generic lifecycle-facing contract | Exact identifier equality. |
| `content_profile_version` | One content family's canonical schema and algorithms | Exact identifier equality. |
| `source_profile_id` | Producer-side transformation identity | Opaque exact identifier. |
| `pack_version` | Publisher-selected release identity | Opaque; no ordering. |
| `pack_release_sequence` | Per-repository, per-pack-key rollback-protection counter | Integer ordering only for import admission. |

**RP-REQ-018**
The exact v1 pack contract identifier MUST be `cartulary.reference_pack_contract.v1`. A change to canonical bytes, identity, digest, trust, dependency, activation, pack-set, or consumer semantics MUST use a new pack contract identifier and a new `contract_major`.

**RP-REQ-019**
Editorial corrections that do not change observable behavior MAY increment only the patch component of `document_version`. Additive content-profile registrations MAY increment the minor component only when existing canonical bytes, defaults, errors, and consumers remain unchanged.

**RP-REQ-020**
The dependencies in Table 4-B identify the imported owners and external contract versions. A newer external contract version MUST NOT be substituted without revising this NLSpec or its named owner. Coordinated owner amendments are reviewed and adopted through the existing repository document-status process. Repository history records the reviewed changes; companion-document hashes, repeated revision bindings and separate signoff dossiers are not required. Tests, generators, runtime metadata and release evidence MUST NOT read, stat or hash Markdown files.

**Table 4-B. Normative dependency registry**

| Dependency | Imported contract | Required baseline |
| --- | --- | --- |
| Core 00 | Profile, precedence, owner matrix | Current adopted Core 00, §§4.2 and 5.1. |
| Core 01 | Routes, jobs, outer lifecycle, snapshots, portability, backup | Current adopted Core 01, §§11, 12, and 17.4. |
| Core 02 | Type registries, indicator identity, persistence minima | Current adopted Core 02, §§10.2, 11, and 14.1. |
| Core 03 | Workbook hot-path and no pack-defined surface boundary | Current adopted Core 03, §2. |
| Core 04 | Authorization, roots, security, limits, conformance | Current adopted Core 04, §§2, 4.1, 9.4, and 12.3. |
| Reporting Subsystem NLSpec | Pack-sensitive reporting consumer boundary | Reporting 1.4.0 companion amendment after `RP-GATE-006`. |
| Report Composition Subsystem NLSpec | Exact admitted preview identity and time | Report Composition 1.3.0 companion amendment after `RP-GATE-006`. |
| Extensions Subsystem NLSpec | Profile/configuration/Job version selection; transaction and deadline protocol | Coordinated companion amendment under `RP-GATE-013`. |
| Core 01 incident portability owner | Format 5 references, embedding and destination reuse | Coordinated §12.3 companion amendment under `RP-GATE-014`. |
| Core 01/Core 04 recovery owners | Current recovery catalog and historical integrity | Coordinated backup/restore companion amendments under `RP-GATE-015`. |
| OpenTelemetry NLSpec | Telemetry mechanics | Current adopted OpenTelemetry NLSpec after `RP-GATE-007`. |
| Testing Harness NLSpec | Fixture and generated-artifact mechanics | Current adopted Testing Harness NLSpec after `RP-GATE-008`. |
| JSON Schema | Structural schema dialect | Draft 2020-12, published 16 June 2022.[^2] |
| RFC 8785 | JSON canonicalization | RFC 8785, June 2020.[^3] |
| TUF | Offline trust and metadata workflow | TUF Specification 1.0.35, last modified 15 July 2026.[^4] |
| SPDX | License expression model | SPDX Specification 3.0.1 and SPDX License List 3.28.0.[^5] |


**RP-REQ-021**
An external dependency supplies only the interface imported in Table 4-B. This NLSpec MUST NOT import an upstream implementation library, repository layout, programming language, or network protocol implicitly.

# 5. Concepts and identifiers

**Table 5-A. Concepts**

| Term | Definition |
| --- | --- |
| `reference pack` | One immutable logical bundle identified by pack key, exact pack version, manifest digest, and payload digest. |
| `pack container` | Exact ZIP, TAR, or GZIP-TAR bytes admitted for import. |
| `pack version` | One immutable retained candidate for one `pack_key` and `pack_version`. |
| `pack manifest` | Canonical signed metadata that declares identity, compatibility, provenance, dependencies, licensing, and payload members. |
| `content profile` | Closed schema and semantic contract for canonical pack payload data. |
| `source profile` | Producer-side, versioned description of transformation from upstream source artifacts into canonical content. |
| `trust repository` | One configured TUF root-of-trust namespace. |
| `verification attestation` | Durable structured evidence for import, verification, activation, disablement, fallback, removal, or trust-root processing. |
| `reference pack set` | Immutable sorted set of active logical pack versions used by one consumer operation. |
| `reproducibility pin` | Durable reference from a snapshot, release, or other retained object to one exact reference pack set. |
| `packaged built-in` | Application-distributed pack version whose bytes are bound to the application release manifest. |
| `operator-imported pack` | Pack version admitted from an upload or deployment-local operator import and verified through the offline TUF profile. |

**RP-REQ-022**
`pack_key` MUST be ASCII and match:

```text
^[a-z][a-z0-9_]{0,31}(\.[a-z][a-z0-9_]{0,31}){1,7}$
```

Its UTF-8 length MUST be in `3..128` bytes. No normalization is applied.

**RP-REQ-023**
`pack_version` MUST be ASCII and match:

```text
^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$
```

It is opaque. The implementation MUST NOT infer semantic-version order, `latest`, `current`, release time, or compatibility from its characters.

**RP-REQ-024**
`pack_release_sequence` MUST be a JSON integer in `1..9007199254740991`. Its comparison scope is the tuple `(trust_repository_id, pack_key)`. Packaged built-ins use an application-release-owned sequence and do not participate in operator-repository rollback comparison.

**RP-REQ-025**
`trust_repository_id` MUST satisfy the `pack_key` lexical contract. `content_profile_id`, `content_profile_version`, `source_profile_id`, algorithm IDs, and schema IDs MUST be non-empty ASCII strings of at most 192 bytes and MUST contain only letters, digits, `.`, `_`, `-`, and `:`.

**RP-REQ-026**
A SHA-256 text value MUST be exactly 64 lowercase hexadecimal characters with no prefix. A raw SHA-256 value used inside an algorithm is exactly 32 bytes.

**RP-REQ-027**
The tuple `(pack_key, pack_version)` identifies immutable logical content deployment-wide. Reimport of the same tuple with the same `manifest_sha256` and `payload_sha256` is exact replay. The same tuple with either digest changed MUST fail with `pack_version_collision` and MUST NOT modify retained bytes or metadata.

**RP-REQ-028**
A `pack_release_sequence` already accepted for the same `(trust_repository_id, pack_key)` MAY be replayed only with the same logical digests. A different logical digest at the same sequence MUST fail with `pack_release_sequence_collision`. A previously unseen lower sequence MUST fail with `pack_release_sequence_rollback`.

# 6. Common JSON, scalar, timestamp, and path contracts

**RP-REQ-029**
Every JSON document and NDJSON line governed by this NLSpec MUST be UTF-8 without BOM. The maximum nesting depth is 10000 open object or array containers, counting the top-level object or array as depth 1 and counting no additional depth for a scalar. Equality is valid; opening the 10001st container fails original JSON admission before schema or canonical-byte checks, using the existing JSON-admission reason for that file family. This bound includes otherwise permitted TUF signed-extension values and applies to a complete bootstrap document including its enclosing objects. Invalid UTF-8, duplicate object members at any depth, invalid JSON syntax, unpaired surrogates, and a top-level value of the wrong JSON type MUST fail before semantic validation.

**RP-REQ-030**
Every canonical JSON document and NDJSON line MUST be byte-for-byte equal to its RFC 8785 representation. Whitespace outside strings is therefore absent, object members use RFC 8785 order, and invalid Unicode terminates validation.[^3]

**RP-REQ-031**
Canonical signed and digest-bearing JSON objects MUST use only:

- `null` where explicitly permitted;
- booleans;
- strings;
- arrays;
- objects;
- JSON integers in `-9007199254740991..9007199254740991`.

Floating-point values, exponent-form numbers, negative zero, NaN, and Infinity are invalid.

**RP-REQ-032**
A JSON integer token MUST use base-10 syntax with no fraction, exponent, leading plus sign, or leading zero except the token `0`. A mathematically integral number written with a fraction or exponent is not an integer under this NLSpec.

**RP-REQ-033**
A timestamp MUST use the Core-owned UTC instant contract. For pack-controlled files, the canonical lexical form is:

```text
YYYY-MM-DDTHH:MM:SSZ
```

Fractional seconds and non-zero offsets are invalid in pack bytes. A calendar date MUST use exact lexical form `YYYY-MM-DD`, MUST identify a valid proleptic-Gregorian date in years `0001..9999`, and MUST contain no timezone or time-of-day component.

**RP-REQ-034**
A single-line human-readable string MUST be Unicode NFC, contain no C0 or C1 control, CR, LF, U+2028, or U+2029, and have no leading or trailing Unicode whitespace. Unless another field table narrows it, the default length domain is `1..256` Unicode scalar values. `ascii_lower_v1` maps only ASCII `A..Z` to `a..z` and leaves every other code point unchanged. A requirement that says ASCII case-insensitive comparison means comparison of `ascii_lower_v1(value)`, with exact UTF-8 bytes as the secondary ordering key when an order is required.

**RP-REQ-035**
A multiline description string MUST be Unicode NFC, use LF only, contain no C0 or C1 control other than LF and horizontal tab, contain no leading or trailing Unicode whitespace, and contain at most 8192 Unicode scalar values. Empty description values MUST be represented by JSON `null`, not `""`.

**RP-REQ-036**
An archive member path MUST be ASCII, use `/` as the only separator, contain no empty segment, no `.` segment, no `..` segment, no NUL, and no leading `/` or `./`. A directory marker has exactly one terminal `/`; remove only that terminal marker before applying segment checks. The complete path including that marker MUST be at most 1024 bytes and each segment MUST be at most 255 bytes. Implementations MUST validate the original name and MUST NOT clean or normalize an invalid name into an accepted path.

**RP-REQ-037**
Empty directory markers remain excluded from the logical inventory and regular-file count, but at most 10000 directory markers may occur in one container. This independent structural bound limits collision-index memory; exceeding it fails with `archive_structure_invalid` before recording the excess marker.

A bounded framing scan MUST precede named extraction. It MUST retain no more than the effective regular-member limit plus 10000 inert header descriptors; exceeding this derived structural ceiling fails `archive_structure_invalid`. This ceiling also bounds inventories of prohibited headers and is compared without integer overflow. Original names and record types remain inert until all applicable path, collision and type checks pass. Extraction MUST be confined to that exact admitted inventory; a changed header or byte stream between preflight and extraction is an operational storage failure.

The terminating TAR zero-block run contains 2–20 blocks inclusive, including the mandatory first pair. The same bound applies after GZIP decompression. Exceeding it is `archive_structure_invalid`; zero padding does not count as regular-file payload or enter its compression ratio. This independent bound prevents compressed structural padding from bypassing extraction budgets. ZIP metadata MUST be processed with bounded working storage before any per-member allocation: names obey the path bound, comments are skipped, extra fields are consumed one member at a time, and inventories retain at most the admitted regular files and directory markers.

Path comparison is exact ASCII byte comparison. The archive MUST additionally reject two paths that differ only by ASCII letter case. Directory member names MUST end with `/`; regular-file member names MUST not end with `/`.

**RP-REQ-038**
Every `extensions` member is required, non-null, and a closed empty object. Its only admitted value in the canonical v1 schemas is `{}`. Omission, null, another JSON type, and any member (including a namespaced member) MUST be rejected. Adding extension semantics requires an adopted schema and profile revision; consumers MUST NOT silently ignore or repair such content. This rule does not change the explicitly admitted TUF `signed` extension policy in Table 10-A.

# 7. Pack container and logical layout

**RP-REQ-039**
The accepted physical container formats are exactly:

- ZIP;
- uncompressed TAR;
- GZIP containing exactly one TAR stream.

The implementation MUST identify format from bytes. Media type and filename are advisory after Core upload-envelope admission.

**RP-REQ-040**
ZIP input MUST use the single-disk non-ZIP64 format and only compression methods `stored` or `deflate`. The first ZIP local-file header MUST begin at byte offset `0`, and the end-of-central-directory record, including any declared ZIP comment, MUST end at the final container byte. TAR input MUST use POSIX ustar headers without GNU, PAX, sparse, link, device, FIFO, or socket extensions; every header MUST have a valid unsigned ustar checksum before type dispatch, and bytes after the first NUL in either fixed-width name or prefix field MUST all be NUL padding; its byte length MUST be a multiple of 512, it MUST end with at least two consecutive 512-byte zero blocks, and bytes after the first terminating pair may contain only additional complete zero blocks. GZIP input MUST begin at byte offset `0`, contain exactly one CRC-valid GZIP member whose decompressed bytes are one conforming ustar archive, and end immediately after that member. The implementation MUST reject encrypted ZIP, prepended ZIP data, multi-disk ZIP, ZIP64, any other ZIP compression method, arbitrary non-TAR GZIP, concatenated or trailing-data GZIP, non-ustar TAR extensions, sparse members, and any other trailing bytes. Archive comments, timestamps, ownership, and permission bits are non-semantic and MUST NOT affect logical identity or execution behavior.

**RP-REQ-041**
The regular files of an operator-imported container MUST have exactly this logical top-level layout. Validated empty directory markers do not add logical files or alter this layout:

```text
bundle.json
manifest.json
metadata/
payload/
```

A packaged built-in logical member tree MUST have exactly:

```text
manifest.json
payload/
```

For either distribution kind, the optional top-level directory `notices/` MAY be present. Operator-imported content MUST contain `bundle.json` and `metadata/`; packaged built-in content MUST contain neither. No other top-level member is valid.

**RP-REQ-042**
For an operator-imported container, `metadata/` MUST contain:

```text
timestamp.json
snapshot.json
targets.json
```

It MAY contain zero or more sequential root-update files named `<version>.root.json`, where `<version>` is a base-10 positive integer with no leading zero. Delegated target-role metadata is forbidden in v1. Packaged built-ins do not contain or process TUF metadata.

**RP-REQ-043**
The selected content profile MUST require exactly one payload shape:

```text
payload/entries.ndjson
```

or:

```text
payload/objects.ndjson
payload/relationships.ndjson
```

A pack containing both shapes, neither shape, or an additional payload member fails unless the registered content profile explicitly declares the additional member. No v1 profile declares an additional payload member.

**RP-REQ-044**
`notices/` MAY contain one or more UTF-8 text files whose paths are declared in the manifest. A notice file MUST use media type `text/plain`, UTF-8 without BOM, Unicode NFC, LF line endings, no NUL, and no C0 or C1 control except LF and horizontal tab. It MUST be non-empty and MUST end in exactly one LF. An omitted `notices/` directory means no notice file is present.

**RP-REQ-045**
Archive validation MUST reject:

- absolute paths;
- traversal;
- path collisions under RP-REQ-037;
- duplicate normalized paths;
- a regular file whose path is the parent of another member;
- symlinks;
- hard links;
- device nodes;
- FIFOs;
- sockets;
- sparse-file encodings;
- member data outside the declared archive stream;
- extraction outside the newly created temporary root.

This validation MUST complete before any untrusted regular-file content is opened by a content parser.

**RP-REQ-046**
Directory entries MAY be present. They do not count as regular-file members, do not satisfy a required logical member, and do not contribute to any digest. Empty undeclared directories are ignored after path and type validation.

**RP-REQ-047**
For an operator-imported container, the complete regular-file set MUST equal:

- `bundle.json`;
- `manifest.json`;
- required TUF metadata files;
- zero or more sequential root-update files;
- every file declared by `manifest.files[]`.

For a packaged built-in logical member tree, the complete regular-file set MUST equal `manifest.json` plus every file declared by `manifest.files[]`. An undeclared regular file or a missing declared file is invalid for either distribution kind.

# 8. Bundle hint and manifest contract

## 8.1 Bundle hint

**RP-REQ-048**
Every operator-imported container MUST contain `bundle.json`. It MUST be a closed canonical object conforming to `cartulary.reference_pack_bundle_hint.v1` with exactly:

| Member | Type | Required | Rule |
| --- | --- | --- | --- |
| `schema_id` | string | Yes | Exactly `cartulary.reference_pack_bundle_hint.v1`. |
| `trust_repository_id` | string | Yes | Satisfies RP-REQ-025. |

The object is untrusted until TUF target verification completes. Invalid UTF-8, BOM, duplicate object members, invalid JSON syntax, a non-object top level, a missing member, an unknown member, a type mismatch, or an invalid scalar MUST fail with `bundle_hint_invalid`. Bytes that otherwise decode to a valid object but are not byte-for-byte equal to their RFC 8785 representation MUST fail with `bundle_hint_noncanonical`.

**RP-REQ-049**
For an operator-imported container, the implementation MUST use the untrusted `trust_repository_id` only to select an already configured trust repository. An unknown repository MUST fail with `tuf_root_untrusted`. After TUF target verification, the verified `bundle.json` bytes MUST equal the exact bytes used for repository selection; a mismatch MUST fail with `metadata_mix_and_match_detected`. Packaged built-ins perform no repository selection.

## 8.2 Manifest

**RP-REQ-050**
`manifest.json` MUST be a closed canonical object conforming to `cartulary.reference_pack_manifest.v1`. Every top-level member in Table 8-A is required. An empty collection MUST be serialized as `[]`; `extensions` MUST be `{}`.

**Table 8-A. Manifest top-level members**

| Member | Type | Null allowed | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | No | Exactly `cartulary.reference_pack_manifest.v1`. |
| `pack_contract_version` | string | No | Exactly `cartulary.reference_pack_contract.v1`. |
| `pack_key` | string | No | RP-REQ-022. |
| `pack_kind` | string | No | Exact registered value for the selected content profile. |
| `pack_version` | string | No | RP-REQ-023. |
| `pack_release_sequence` | integer | No | RP-REQ-024. |
| `content_profile_id` | string | No | Registered in Table 13-A. |
| `content_profile_version` | string | No | Exact value registered for the profile. |
| `source_profile_id` | string | No | RP-REQ-025; producer provenance only. |
| `source_profile_sha256` | string | No | SHA-256 of the exact immutable source-profile artifact under RP-REQ-150. |
| `trust_repository_id` | string or null | Yes | Non-null and equal to verified `bundle.json` for operator-imported packs; exactly `null` for packaged built-ins. |
| `source_identifier` | string | No | Single-line string, maximum 512 scalars. |
| `source_version` | string | No | Single-line string, maximum 256 scalars. |
| `source_as_of` | timestamp or null | Yes | Operator-imported packs require a non-null timestamp; packaged built-ins may use `null` when the application release has no distinct upstream snapshot time. |
| `built_at` | timestamp | No | Pack build completion time. |
| `builder` | object | No | Table 8-B. |
| `source_artifacts` | array | No | Table 8-C; `1..64` items. |
| `compatibility` | object | No | Table 8-D. |
| `dependencies` | array | No | Table 8-E; `0..64` items. |
| `conflicts` | array | No | Table 8-F; `0..64` items. |
| `files` | array | No | Table 8-G; exactly one or two payload items plus `0..64` notice items, for a total of `1..66`. |
| `content_summary` | object | No | Table 8-H. |
| `license` | object | No | Table 8-I. |
| `extensions` | object | No | RP-REQ-038. |

**Table 8-B. `builder` object**

| Member | Type | Rule |
| --- | --- | --- |
| `builder_id` | string | Stable single-line identifier, maximum 192 bytes. |
| `builder_version` | string | Opaque single-line version, maximum 128 bytes. |
| `builder_source_sha256` | string | SHA-256 of the exact immutable builder executable, package, or archive bytes used. A mutable source-tree directory is not a valid digest subject unless a separately adopted tree-digest algorithm first serializes it into one immutable artifact. |

**Table 8-C. `source_artifacts[]` item**

| Member | Type | Rule |
| --- | --- | --- |
| `source_ref` | string | Non-empty single-line source reference, maximum 1024 scalars. |
| `source_version` | string | Non-empty single-line source version or snapshot identity. |
| `sha256` | string | Exact source artifact SHA-256. |

Items MUST sort by `source_ref`, then `source_version`, then `sha256`. Exact duplicate items are invalid.

**Table 8-D. `compatibility` object**

| Member | Type | Required v1 value |
| --- | --- | --- |
| `cartulary_contract_majors` | integer array | Exactly `[1]`. |
| `required_capabilities` | string array | Default and required v1 value `[]`. |

A non-empty `required_capabilities` array is contract-incompatible in v1.

**Table 8-E. `dependencies[]` item**

| Member | Type | Rule |
| --- | --- | --- |
| `pack_key` | string | RP-REQ-022. |
| `pack_version` | string | RP-REQ-023. |
| `payload_sha256` | string | Exact required logical payload digest. |

Dependencies MUST sort by `pack_key`, then `pack_version`, then `payload_sha256`. Identical duplicates coalesce during manifest normalization only in producer tooling; admitted canonical bytes containing duplicates are invalid.

**Table 8-F. `conflicts[]` item**

| Member | Type | Null allowed | Rule |
| --- | --- | ---: | --- |
| `pack_key` | string | No | RP-REQ-022. |
| `pack_version` | string or null | Yes | `null` means every version of the key conflicts. |

Conflicts MUST sort by `pack_key`, then `pack_version`, with `null` before strings. Duplicate items are invalid.

**Table 8-G. `files[]` item**

| Member | Type | Rule |
| --- | --- | --- |
| `path` | string | Exact payload or notice path under RP-REQ-036. |
| `role` | string | Exactly `payload` or `notice`. |
| `media_type` | string | Exactly `application/x-ndjson` for payload and `text/plain` for notice. |
| `size_bytes` | integer | `0..268435456`; equality at the maximum is valid. |
| `sha256` | string | SHA-256 of exact raw file bytes. |

Items MUST sort by `path` as ascending ASCII bytes. Paths MUST be unique. `files[]` MUST NOT list `bundle.json`, `manifest.json`, or metadata files.

**Table 8-H. `content_summary` variants**

| `kind` | Required members | Forbidden members |
| --- | --- | --- |
| `entries` | `kind`, `entry_count` | `object_count`, `relationship_count` |
| `objects_relationships` | `kind`, `object_count`, `relationship_count` | `entry_count` |

Counts MUST be integers in `0..9007199254740991` and MUST equal parsed payload counts.

**Table 8-I. `license` object**

| Member | Type | Rule |
| --- | --- | --- |
| `expression` | string | ASCII SPDX expression, `1..1024` bytes, valid under SPDX 3.0.1. |
| `license_list_version` | string | Exactly `3.28.0`. |
| `redistribution` | string | Exactly `allowed`, `restricted`, or `prohibited`. |
| `notice_paths` | string array | `0..64` sorted exact paths under `notices/`; no duplicates. |
| `license_ref_bindings` | array | `0..64` sorted closed bindings from every `LicenseRef-*` token to one notice path; `[]` when none. |

A `license_ref_bindings[]` item contains exactly `license_ref` and `notice_path`. `license_ref` MUST be one exact `LicenseRef-*` token present in `license.expression`; `DocumentRef-*` and `AdditionRef-*` forms are forbidden in v1. SPDX License List identifiers and exception identifiers compare ASCII case-insensitively; operators are exactly all-uppercase or all-lowercase SPDX spellings. Distinct spellings of one case-insensitive LicenseRef suffix in the same expression are invalid, so each user-defined identity has one exact binding token. `notice_path` MUST appear in both `notice_paths[]` and `manifest.files[]` with role `notice`. Bindings sort by `license_ref`, then `notice_path`; duplicates are invalid.

**RP-REQ-051**
Every `LicenseRef-*` token in `license.expression` MUST have exactly one `license_ref_bindings[]` item. A binding for a token absent from the expression is invalid. Every bound notice file MUST contain the corresponding license or notice text. An expression containing only SPDX License List identifiers and exceptions requires `license_ref_bindings=[]`. `notice_paths[]` MUST equal the complete set of `manifest.files[]` paths whose role is `notice`; if the set is empty, no regular file may appear under `notices/`; a valid empty directory marker has no logical meaning under RP-REQ-036.

`built_at` is an explicit reproducible build input. Producer tooling MUST NOT default it from the current clock. Given identical manifest inputs other than TUF metadata, identical declared payload and notice bytes, the same source profile, and the same builder artifact, producer tooling MUST emit byte-identical `manifest.json`, payload files, notice files, `manifest_sha256`, and `payload_sha256`. When `source_as_of` is non-null, it MUST be less than or equal to `built_at`.

**RP-REQ-052**
`manifest.files[]` is the only authoritative logical payload inventory. Archive order, compression metadata, local filesystem order, and parser enumeration order MUST NOT affect logical identity.

# 9. Canonical digest algorithms

**RP-REQ-053**
`container_sha256` MUST equal lowercase hexadecimal SHA-256 of the exact admitted container bytes. It is an audit and upload-idempotency value. It is not logical pack identity.

**RP-REQ-054**
`manifest_sha256` MUST be computed as:

```text
lowercase_hex(SHA256(RFC8785_JCS(manifest.json)))
```

The canonical bytes MUST equal the admitted `manifest.json` bytes.

**RP-REQ-055**
`payload_sha256` MUST use `reference_pack_payload_digest_v1` exactly:

```text
input = ASCII("cartulary.reference_pack.payload.v1") || NUL

for each manifest.files[] item in manifest order:
    path_bytes = ASCII(item.path)
    input += uint32_be(length(path_bytes))
    input += path_bytes
    input += uint64_be(item.size_bytes)
    input += raw_32_byte_sha256(item.sha256)

payload_sha256 = lowercase_hex(SHA256(input))
```

The algorithm includes payload and notice members. It excludes TUF metadata, `bundle.json`, `manifest.json`, archive headers, and compression bytes.

**RP-REQ-056**
Two containers with different physical formats or archive metadata but identical canonical manifest and declared file bytes MUST produce the same `manifest_sha256` and `payload_sha256`. Their `container_sha256` values MAY differ.

**RP-REQ-057**
A digest comparison MUST use constant-time comparison for equal-length raw digest bytes. Hex decoding failure is a schema failure, not a digest mismatch.

# 10. Verification methods and Cartulary TUF profile

## 10.1 Verification-method registry

**RP-REQ-058**
The v1 verification-method registry is closed:

| `verification_method` | `distribution_kind` | Permitted input |
| --- | --- | --- |
| `packaged_release_manifest_v1` | `packaged_builtin` | Application-packaged built-in registry snapshot. |
| `tuf_1_0_35_offline_bundle_v1` | `operator_imported` | Upload or deployment-local operator import. |

No `unsigned`, `checksum_only`, `tls_only`, `trusted_source`, or `user_accepted` method is conformant. `distribution_kind` and `verification_method` are server-derived from the application release binding or the admitted operator route; a caller or manifest MUST NOT select them.

## 10.2 Packaged built-ins

**RP-REQ-059**
The application release manifest MUST contain one closed `cartulary.reference_pack_builtin_release_binding.v1` item for each packaged built-in. Each item MUST contain exactly `schema_id`, `pack_key`, `pack_version`, `pack_release_sequence`, `manifest_sha256`, `payload_sha256`, `content_profile_id`, and `content_profile_version`; `schema_id` MUST equal `cartulary.reference_pack_builtin_release_binding.v1`. Items MUST sort by `pack_key`, then `pack_version`, and duplicate `pack_key` values are invalid within one application release. Verification MUST recompute the manifest and payload digests from the packaged logical member tree and compare every binding member before ready state. The release binding is trusted only as part of the same application-release integrity boundary that admits the running executable; an operator-created runtime file cannot claim `packaged_release_manifest_v1`. The release binding and its exact logical member bytes MUST be retained with the admitted built-in version for later backup, restore, and reproducibility checks.

**RP-REQ-060**
Packaged built-in bytes MUST be verified before application readiness. Failure to verify any required Base registry MUST block ready state. Packaged built-ins do not expire and do not use TUF metadata.

## 10.3 Trust bootstrap

**RP-REQ-061**
When the Reference Pack Extension Profile is claimed, Core 04 MUST require `reference_packs.trust_bootstrap_path`. The value MUST be an absolute path to one regular non-symlink file. The file MUST contain at most `8388608` bytes, MUST be read before Reference Pack workers start, and MUST conform to `cartulary.reference_pack_trust_bootstrap.v1`. Equality at the byte maximum is valid; a larger file is invalid deployment configuration.

**RP-REQ-062**
`cartulary.reference_pack_trust_bootstrap.v1` is a closed canonical JSON object:

| Member | Type | Rule |
| --- | --- | --- |
| `schema_id` | string | Exactly `cartulary.reference_pack_trust_bootstrap.v1`. |
| `repositories` | array | `1..64` items sorted by `repository_id`. |

Each repository item contains exactly:

| Member | Type | Rule |
| --- | --- | --- |
| `repository_id` | string | `trust_repository_id` contract. |
| `trusted_root` | object | Complete canonical TUF root metadata object. |
| `trusted_root_sha256` | string | SHA-256 of canonical `trusted_root` bytes. |

Duplicate `repository_id` values are invalid. Every `trusted_root.signed` object MUST contain a top-level closed `cartulary` member with exactly `schema_id='cartulary.reference_pack_tuf_root_binding.v1'` and `trust_repository_id=repository_id`. Every sequential root update MUST preserve that binding exactly.

**RP-REQ-063**
The bootstrap file is deployment configuration, not a secret, not incident portability content, and not a browser-editable object. Current trusted-root state derived from it is authoritative retained deployment state and MUST participate in operational backup and restore.

## 10.4 TUF POUF

**RP-REQ-064**
`tuf_1_0_35_offline_bundle_v1` MUST implement the closed POUF in Table 10-A. TUF supplies the update-security workflow, but interoperability requires this application-specific format profile.[^4]

**Table 10-A. Cartulary offline TUF POUF**

| Property | Required value |
| --- | --- |
| TUF specification | Exactly `1.0.35`; every metadata `signed.spec_version` MUST equal `1.0.35`. |
| Metadata format | RFC 8785 canonical JSON. |
| Top-level roles | `root`, `targets`, `snapshot`, `timestamp`. |
| Delegated roles | Forbidden. |
| Consistent snapshots | `false`. |
| Required target hash | Exactly one `hashes.sha256` value per target or metadata descriptor; every other hash key is invalid. |
| Metadata version domain | Every root, targets, snapshot, and timestamp `signed.version` is an integer in `1..9007199254740991`. |
| Permitted key type and scheme | `keytype='ed25519'`, `scheme='ed25519'` only. |
| Root key count | At least 3 distinct root-role key IDs. |
| Root threshold | At least 2 and not greater than root key count. |
| Targets threshold | At least 1. |
| Snapshot threshold | At least 1. |
| Timestamp threshold | At least 1. |
| Target inventory | `bundle.json`, `manifest.json`, every `payload/**` file, every `notices/**` file. |
| Verification time | Server UTC captured once at verification-job start. |
| Minimum remaining validity | Every final root, targets, snapshot, and timestamp expiry MUST be strictly later than the fixed verification start time. No additional minimum remaining-validity interval applies in v1. |
| Maximum remaining timestamp validity | `2678400` seconds. |
| Maximum remaining snapshot validity | `8035200` seconds. |
| Maximum remaining targets validity | `31622400` seconds. |
| Maximum remaining final-root validity | `63072000` seconds. |
| Expiry ordering | `timestamp.expires <= snapshot.expires <= targets.expires <= final_root.expires`. |
| Root updates | Sequential `<version>.root.json` files only. |
| Mirror access | None. All metadata and targets are local archive members. |
| Metadata compression | Forbidden. |
| Unknown signed fields | Preserved and signed; no v1 semantic effect unless this NLSpec names them. |

**RP-REQ-065**
Every TUF metadata file MUST be canonical JSON and have the standard outer object with exactly `signed` and `signatures`. Invalid UTF-8, BOM, duplicate object members, invalid JSON syntax, a non-object top level, an invalid TUF role schema, or an unknown outer member MUST fail with `tuf_metadata_invalid`. Bytes that otherwise decode to valid TUF metadata but are not byte-for-byte equal to their RFC 8785 representation MUST fail with `metadata_noncanonical`.

Each signature item MUST contain exactly `keyid` and `sig`. `keyid` MUST be 64 lowercase hexadecimal characters. `sig` MUST be 128 lowercase hexadecimal characters encoding one Ed25519 signature over `RFC8785_JCS(signed)`. Signature items MUST sort by `keyid`, MUST have unique key IDs, and MUST NOT contain an unknown member. Multiple signatures by the same key ID are invalid and MUST NOT count more than once.

The final root `signed.roles` object MUST contain exactly `root`, `targets`, `snapshot`, and `timestamp`. Each role object MUST contain exactly `keyids` and `threshold`. Every `keyids` array MUST contain unique key IDs sorted by ascending ASCII bytes. Every referenced key ID MUST exist in the same root `signed.keys` object. A duplicate, unsorted, absent, or unknown role-key reference MUST fail with `tuf_root_untrusted` for the bootstrap root or `tuf_root_rotation_invalid` for a root update.

**RP-REQ-066**
Every admitted TUF key object MUST contain exactly `keytype`, `scheme`, and `keyval`; `keyval` MUST contain exactly `public`. `keytype` and `scheme` MUST both equal `ed25519`. `keyval.public` MUST be exactly 64 lowercase hexadecimal characters encoding the 32-byte Ed25519 public key. The key ID MUST equal `lowercase_hex(SHA256(RFC8785_JCS(key_object)))`. A key object or role reference whose supplied key ID differs from that result is invalid and fails with `tuf_root_untrusted` for the bootstrap root or `tuf_root_rotation_invalid` for a root update.

**RP-REQ-067**
Root update processing MUST begin from the currently trusted root for the selected repository. Every next root version MUST be exactly predecessor version plus one and MUST satisfy both:

- the predecessor root's root-role threshold over the new root's signed bytes;
- the new root's root-role threshold over the same signed bytes.

A gap, version regression, invalid old threshold, invalid new threshold, or changed repository identity fails with `tuf_root_rotation_invalid`.

Root-update signature authorization uses the union of the predecessor and successor root-role key sets. Each signature MUST verify with its matching key; the predecessor and successor thresholds MUST then be counted independently over their respective authorized subsets. A key shared by both roles may count once toward each threshold. Root-update metadata permits at most 128 signatures; other metadata permits at most 64. Unknown, duplicate, or invalid extra signatures fail even when a threshold is otherwise satisfied.

A packaged root file at or below the current trusted version is consumed history, not a new update. It MUST be byte-identical to the retained canonical root metadata of that version; an absent historical record or byte difference fails with `tuf_root_rotation_invalid`. Exact consumed roots do not advance trust or replace the current root. The next unconsumed root MUST continue from the current trusted root with no version gap.

**RP-REQ-068**
The implementation MUST persist the highest trusted root version per `trust_repository_id`. It MUST persist the highest trusted `timestamp`, `snapshot`, and `targets` metadata version per `(trust_repository_id, pack_key, pack_version, role)`. Metadata below the applicable persisted version fails with `metadata_rollback_detected`; exact same-version replay is valid only when canonical bytes are identical. Cross-release rollback is governed by `pack_release_sequence`, not by sharing timestamp, snapshot, or targets version counters between different `pack_version` values.

**RP-REQ-069**
The final trusted root, targets, snapshot, and timestamp metadata MUST each have an expiry strictly later than the fixed verification start time. No additional minimum remaining-validity interval applies. A predecessor root in a sequential update chain MAY be expired only for verifying the immediately succeeding root; omission behavior: an expired predecessor root cannot serve as the final trusted root. Each final-role expiry MUST be no more than its Table 10-A maximum remaining validity after the verification start time, and the four expiries MUST satisfy the Table 10-A ordering. An expired final role fails with `metadata_expired`; an excessive horizon or invalid ordering fails with `metadata_expiry_policy_invalid`. The earliest expiry among the successfully trusted final timestamp, snapshot, targets, and root metadata MUST be stored as `trust_valid_until`.

**RP-REQ-070**
Snapshot metadata MUST contain exactly one metadata descriptor, for `targets.json`, and bind its exact version, length, and SHA-256. Timestamp metadata MUST contain exactly one metadata descriptor, for `snapshot.json`, and bind its exact version, length, and SHA-256. Any additional descriptor or mismatch fails with `metadata_mix_and_match_detected`.

**RP-REQ-071**
Targets metadata MUST declare exactly `bundle.json`, `manifest.json`, every payload file, and every notice file, each exactly once. Each target descriptor MUST contain exactly `length` and `hashes`; `length` MUST equal the exact target byte length, and `hashes` MUST contain exactly one member named `sha256` whose value is the exact lowercase digest. A required file absent from targets metadata fails with `target_not_declared`; an additional target fails with `unexpected_target`; a length mismatch fails with `target_length_mismatch`; a hash mismatch fails with `checksum_mismatch`.

**RP-REQ-072**
The signed targets metadata MUST contain a top-level extension member named `cartulary` with exactly:

| Member | Rule |
| --- | --- |
| `schema_id` | `cartulary.reference_pack_tuf_targets_binding.v1` |
| `trust_repository_id` | Equals verified `bundle.json`. |
| `pack_key` | Equals manifest. |
| `pack_version` | Equals manifest. |
| `pack_release_sequence` | Equals manifest. |
| `manifest_sha256` | Equals computed manifest digest. |
| `payload_sha256` | Equals computed payload digest. |

Missing binding data and mismatches between structurally admitted binding values and verified identities fail with `metadata_mix_and_match_detected`. A present binding must be an object; present members must have their declared scalar types and domains, and unknown nested members are invalid during `tuf_schema`. Omission is not repaired: required binding completeness is checked at `metadata_links` before expiry or target comparisons.

**RP-REQ-073**
Threshold satisfaction MUST count only distinct valid signatures from key IDs authorized for the applicable role in the currently trusted root, except for the explicitly defined old/new root authorization union in RP-REQ-067. An untrusted signer, invalid signature, duplicate signer, or insufficient distinct threshold fails with `signature_threshold_not_met`, or `tuf_root_rotation_invalid` when evaluating a root update. Verification MUST examine every supplied signature and MUST NOT stop after reaching a threshold.

**RP-REQ-074**
The successful verification attestation MUST retain the complete sorted set of valid authorized supplied signer key IDs for each role, not an implementation-selected threshold subset. Root transitions retain predecessor-authorized and successor-authorized signer sets separately. Profile major 2 exposes the complete targets-role set through `verified_signer_key_ids[]` under RP-REQ-215; the retired scalar `signer_key_id` is not an alternate projection. Coordinated adoption of that Core amendment remains required before a Reference Pack conformance claim.

## 10.5 Expiry and historical use

**RP-REQ-075**
Import, reverify, and refresh MUST evaluate trust freshness at one UTC instant captured when an execution attempt begins verification, after queueing. Every member of a refresh cohort uses that same instant. Activation captures its instant when its semantic operation starts. `last_verified_at` is the successful verification instant, not admission or completion time. Admission, start, and completion timestamps remain separately attributable. An attempt that starts at or after metadata expiry fails even if it was admitted before expiry. Completion after expiry does not invalidate an otherwise successful attempt. A recovered Jobs attempt retains the admitted byte references and cohort but captures a new verification instant. Passage of time beyond `trust_valid_until` MUST NOT by itself mutate an already published active set.

Core 04 MUST define non-null boolean `reference_packs.clock_trusted`, with omission resolving to `false`. Only deployment configuration authority controls this operator assertion; no implicit network-time heuristic exists. False rejects fresh operator TUF verification and activation with `reference_pack_operation_rejected/clock_untrusted` without invalidating retained content. Built-in verification, pinned historical consumption, and zero-cohort refresh do not require the assertion.

**RP-REQ-076**
A later activation, reverify, or refresh of an operator-imported version whose required metadata is expired MUST fail with `metadata_expired`. A retained historical snapshot or report MUST continue to consume its exact previously verified pinned pack when the required bytes, digests, content-profile implementation, and successful attestation remain available. Historical consumption MUST NOT re-run update-freshness admission and MUST fail with `required_reference_pack_unavailable` only when the exact pinned set cannot be reconstructed or consumed.

# 11. Built-in registries, distribution kinds, and lifecycle

**RP-REQ-077**
Every Base Profile release MUST provide packaged built-in versions for exactly these required pack keys:

- `type_registry.host`;
- `type_registry.evidence`;
- `type_registry.indicator`.

The exact canonical built-in pack bytes and digests MUST be retained as release fixtures. They are not hard-coded label maps.

**RP-REQ-078**
On startup of an application release, every current-release packaged built-in version absent from retained pack state MUST be admitted after release-manifest verification. Its exact release binding and logical member bytes MUST be retained. Every previously admitted packaged built-in version MUST remain retained across application upgrades so a pinned prior pack set remains reconstructable.

Before ready state, startup reconciliation MUST execute in one publication boundary: verify all current-release built-ins; re-evaluate the byte availability, digests, pack contract, content profile, dependencies, and runtime compatibility of every retained active member without reapplying metadata-expiry freshness; invalidate unavailable or incompatible optional members; and publish one complete replacement set when the effective member list changes. When the Reference Pack Extension Profile is unclaimed, reconciliation MUST remove every operator-imported active pointer, invalidate every optional dependent, select the current-release packaged built-in for each required key with `activation_mode='application_release'`, and retain imported versions only as inactive historical state. When the profile is claimed, a compatible active operator-imported required registry remains active and the current-release built-in becomes its fallback; a required key without such an imported active version MUST use the current-release built-in. Startup reconciliation MUST record a `profile_reconciliation` attestation whenever it changes an effective pointer or invalidates an optional member.

**RP-REQ-079**
When the Reference Pack Extension Profile is not claimed, the internal registry resolver MUST use exactly the three current-release packaged built-ins selected by startup reconciliation, imported versions MUST remain inactive, and the public Reference Pack route family MUST remain unavailable. When the profile is claimed, all retained packaged built-ins MUST become visible as `distribution_kind='packaged_builtin'`, and every effective required or optional version MUST participate in the active reference pack set.

**RP-REQ-080**
A packaged built-in version MUST NOT be disabled, removed, or reverified through the operator-imported TUF path. An imported verified version of the same required key MAY be explicitly activated as a replacement.

**RP-REQ-081**
The durable version-condition vocabulary imported from Core 01 is exactly:

```text
staged
verified_available
disabled
failed
missing
```

`active` is a derived boolean from the active pointer and is not another condition token.

**RP-REQ-082**
The legal verification-condition transitions are:

| From | To | Required cause |
| --- | --- | --- |
| absent | `staged` | New admitted operator import or exact reimport of a missing version. |
| `staged` | `verified_available` | Successful complete verification and publication. |
| `staged` | `failed` | Verification failure or cancellation after candidate identity became durable. |
| `staged` | `missing` | Staged container or payload became unavailable before verification completed. |
| `verified_available` | `disabled` | Explicit disable of an operator-imported version. |
| `disabled` | `disabled` | Successful reverify preserves administrative disablement. |
| `disabled` | `verified_available` | Explicit successful activation clears administrative disablement. No separate re-enable operation exists in v1. |
| `verified_available`, `disabled`, or active | `failed` | Later reverify or integrity check fails. |
| `verified_available`, `disabled`, or active | `missing` | Required retained payload becomes unavailable. |
| `failed` or `missing` | `staged` | New exact reimport of the same logical version. |
| `failed` or `missing` | `verified_available` | Successful reverify only when all retained bytes are present. |

These public conditions project independent verification health and administrative disablement. Health is `staged`, `verified_available`, `failed`, or `missing`; administrative disablement is a separate boolean. The public condition is `disabled` exactly when health is `verified_available` and administrative disablement is true; otherwise it equals health. Verification never clears that boolean, including after a failed or missing disabled version recovers. Only explicit successful activation clears it. An initial import with unavailable staging bytes becomes `missing` with `missing_reason='staging_loss'`; an operational abort with no content verdict becomes `failed`. Any unlisted transition is invalid.

**RP-REQ-083**
At most one version per `pack_key` may be active. Activation of one version atomically moves the previous active operator-imported or packaged-built-in version to inactive `verified_available` without changing that version's condition.

**RP-REQ-084**
A failed or missing active optional pack MUST lose its active pointer and MUST NOT trigger automatic activation of another optional version. A future activation requires explicit administrator action.

**RP-REQ-085**
A failed or missing active imported required Base registry MUST trigger `safety_fallback` to the packaged built-in version in the same publication transaction. If the packaged built-in cannot be verified, the deployment MUST become not-ready.

**RP-REQ-086**
The safety-fallback transaction MUST:

1. remove the invalid imported active pointer;
2. activate the packaged built-in for the same key;
3. invalidate every exact dependent no longer satisfied;
4. create a `safety_fallback` attestation;
5. publish one replacement reference pack set.

No other key receives automatic fallback.

**RP-REQ-087**
The closed activation-mode registry is `normal`, `rollback`, `safety_fallback`, and `application_release`. Explicit activation of a retained verified version that has a prior successful activation attestation, or of an operator-imported target whose sequence is lower than the replaced version in the same sequence scope, MUST use `rollback`. Startup selection of the current application release built-in uses `application_release`; RP-REQ-085 uses `safety_fallback`; every other explicit activation uses `normal`. Every activation preserves sequence history and records the replaced version. Import rollback remains forbidden under RP-REQ-028.

**RP-REQ-088**
Mutable deployment-local registry overrides are not supported in v1. Organization-specific registry behavior MUST be delivered as a complete immutable operator-imported pack version. Core 02 §11 carries the corresponding replacement-pack rule; adoption still requires the exact amended owner revision to be recorded.

# 12. Verification pipeline, issues, and publication boundary

**RP-REQ-089**
For an operator-imported container, `verify_reference_pack_v1` MUST execute phases in this exact order:

```text
1. confirm Core upload or local-operator admission;
2. identify container format from bytes;
3. enforce source-byte and archive preflight limits;
4. validate archive member paths and types;
5. extract into a new private temporary root;
6. decode and validate bundle.json;
7. select configured trust repository;
8. decode and verify TUF root, timestamp, snapshot, and targets metadata;
9. verify the TUF target inventory, lengths, and hashes;
10. decode and validate manifest.json;
11. verify manifest identity syntax, pack-key/profile registration, and release-sequence ordering that does not depend on logical digests;
12. compare the manifest file set to extracted regular files;
13. verify every declared file length and SHA-256;
14. compute `manifest_sha256` and `payload_sha256`, verify the signed Cartulary targets binding in RP-REQ-072, and enforce immutable-tuple, same-sequence-digest, and exact-replay rules;
15. validate content-profile structure;
16. validate content-profile semantics;
17. validate runtime compatibility;
18. validate dependencies and conflicts;
19. validate type-registry compatibility when applicable;
20. build deterministic derived indexes;
21. atomically publish candidate state, indexes, successful trust-root and trusted-metadata-version updates, attestation, and validation summary;
22. remove the temporary extraction root.
```

**RP-REQ-090**
The ordered check registry in Table 12-B is the sole content failure-precedence authority. The first failing check controls the terminal content result. Its issues map to `error.code='reference_pack_verification_failed'` with `reason_code=code`. Cancellation maps to the common canceled-job outcome. Timeout retains its Core-owned `verification_timeout` public mapping but is an execution abort, not a content verdict. Indexing and publication failures map to the common internal job-failure outcome. Abort classification and proven commit precedence follow the adopted Extensions deadline and owner-finalization contracts.

**RP-REQ-091**
After manifest identity is safely decoded and the `staged` identity is committed, the Core-owned `staged` metadata resource MUST be queryable while verification continues. A failure before safe identity attribution MUST create no pack-version resource. Before phase 21 commits, the implementation MUST NOT expose extracted content, derived indexes, verified provenance, or a candidate as usable pack data. Success-only metadata is explicit `null` until first successful publication; signer arrays are empty. Operation effects are closed by this table:

| Outcome | Atomic effects |
| --- | --- |
| First import content rejection | Attributable failed or missing candidate, validation summary, failure attestation, and terminal job; no usable data or trust advancement. |
| Established-version renewal rejection | Failed-attempt evidence only; preserve the successful envelope, health, administrative disablement, activation, and trust state. |
| Reverify content rejection or missing bytes | Publish the verdict and required active invalidation, dependency pruning, and safety fallback together. |
| Mixed refresh | Publish every cohort verdict, successful indexes, compatible successful trust proposals, attestations, invalidations, and one resulting set in one transaction; the job fails when any member fails. |
| Cancellation, timeout, stale state, indexing failure, or proven absent publication | No partial semantic delta and no invalid-content verdict for established versions. |

An initial staged candidate whose attempt aborts MUST become terminal `failed` with no successful metadata and no content-validation verdict. Attempt outcome records the abort; it MUST NOT remain staged indefinitely. Jobs-owned terminal failure or cancellation, including exhausted execution retries and inactive-profile reconciliation, MUST close the admitted Reference Data operation and any existing unfinished attempt in the same terminal transaction. Established health, successful envelopes, disablement, trust, indexes, activation and historical provenance remain unchanged. A never-started verification MUST NOT receive an invented attempt row or verification instant. Its operation records the operational abort, and attributable staged candidates become failed without a content verdict. Repeated terminal reconciliation is a mutation-free no-op after the owner outcome is terminal. Before readiness, Reference Data MUST reconcile retained nonterminal operations against the Jobs-owned terminal lifecycle projection, including after public Job expiry and with the administration profile unclaimed. A nonterminal Job never authorizes that repair. Reconciliation uses the common finalizer and its exact immutable owner receipt to classify uncertain commit; it does not infer outcomes from process age or lost leases. A common Job success without terminal owner publication MUST fail closed. A crash with an indeterminate terminal mutation follows the imported Jobs/Extensions recovery and integrity protocol. Conflicting refresh root proposals abort the whole cohort without publishing member verdicts. Proven committed results remain authoritative despite lost acknowledgement. Failed members never contribute trust advancement.

**RP-REQ-092**
A verification issue MUST conform to closed `reference_pack_issue.v1`:

| Member | Type | Rule |
| --- | --- | --- |
| `issue_id` | string | Deterministic ID from RP-REQ-095. |
| `check_id` | string | Exact stable check token from Table 12-B; never a specification requirement ID. |
| `severity` | string | Exactly `error`; execution aborts are operation outcomes, not content issues. |
| `phase` | string | Exact phase token from Table 12-A. |
| `code` | string | Exact public verification reason assigned to the check by Table 12-B. |
| `reason_code` | string or null | Equal to `code` for every emitted content issue; the member remains explicitly nullable in the wire schema. |
| `path` | string | Exact pack path or canonical JSON diagnostic path. |
| `entry_id` | string or null | Exact content entry ID when attributable. |
| `safe_details` | object | Closed object defined below. |

`safe_details` MUST contain exactly `limit_id`, `expected_token`, `actual_token`, `related_pack_key`, and `related_pack_version`. Each member is a string or `null`; values not applicable to the issue are `null`. Every non-null string is limited to 256 Unicode scalar values, MUST satisfy RP-REQ-034, and MUST contain no payload-derived value. `related_pack_key` and `related_pack_version` MUST satisfy RP-REQ-022 and RP-REQ-023. Count and byte-limit tokens MUST use unsigned base-10 with no leading zero except `0`. A canonical JSON diagnostic path starts at `$`, uses `.member` for an ASCII identifier member, `[<zero-based integer>]` for an array index, and `[` plus an RFC 8785 JSON string plus `]` for any other member name; `path` is limited to 4096 UTF-8 bytes. Validation issues MUST NOT expose specification requirement identifiers.

Every emitted verification issue makes content verification fail. No warning or informational verification issue is emitted in this contract. Diagnostic paths may contain only admitted archive paths, schema-known member names, and array indexes; hostile unknown names use the fixed segment `["<unknown>"]` and MUST NOT be echoed.

Lexical manifest attribution does not select a manifest verdict before trust selection. Missing or oversized `bundle.json` fails `bundle_shape`. Repository selection completes before TUF metadata admission. An absent required top-level role file fails `tuf_schema`. Descriptor members that are present must have their declared types, integer domains, lowercase digest encoding, and closed member sets during `tuf_schema`; explicit null and unknown descriptor or hash members are invalid. Omitted integrity members instead fail the separately ranked `metadata_integrity` check. That check covers every metadata and target descriptor before any link, expiry, or inventory comparison. Malformed root-version filenames are held for `root_rotation` after metadata schema, canonical bytes and current-root trust checks. These rules do not permit reading beyond an applicable resource guard. Archive format identification precedes the source-byte guard. Framing and bounded decompression precede path and collision checks. A framing failure prevents interpreting later headers. A count, byte or ratio guard stops before processing the exceeding member body; hypothetical faults beyond that boundary are inapplicable. A prohibited extension or special-member body is never decoded to discover later headers. Plain zero-body special headers remain independently enumerable. Among the safely enumerated original names, paths sort by exact bytes for virtual `$.archive_members[index]` diagnostics; every member participating in a case, duplicate or file/descendant collision receives a finding. No rejected archive name reaches named extraction. Manifest and bundle-hint shape checks enumerate all admitted structural nodes. Manifest cross-field invariants execute only after the complete typed structure passes shape admission; they never reconstruct a partial typed object from malformed members. With an admitted typed manifest, all independently evaluable cross-field failures are collected; a comparison requiring a valid timestamp is inapplicable when that timestamp is invalid. Canonical-byte comparison follows those schema checks.

For reverify and refresh, retained logical members participate at the same declared ranks as admitted container members. `retained_payload` checks the presence of every required retained object before container or trust checks; it does not hash those objects. Retained length and digest comparisons execute at `member_length` and `member_hash`, respectively, after all earlier checks pass. A present altered member therefore does not mask an earlier trust failure. A missing member precedes an altered member regardless of path traversal order. Consumer-detected integrity checks use presence, length and digest order without performing fresh trust verification. These checks never reconstruct a missing member from a retained container. Each retained-member check enumerates the entire immutable logical inventory, at most one manifest plus 66 declared files, and uses `$.objects[index]` in logical-path order. Unexpected or duplicate logical binding rows are retained-state integrity failures; missing bindings, physical absence and member byte failures are content findings. Storage execution failures abort the attempt even when earlier members produced findings. Container absence is combined with retained logical absence in that same check. Container and retained length/hash findings are combined before proceeding to the next rank; neither source can hide the other's enumerable findings. Manifest inventory comparisons use `$.files[index]` and its declared `size_bytes` or `sha256` member; extra regular files use `$.archive_members[index]` in full inventory path order. Each independent signed target-binding mismatch uses `$.binding.member`, without echoing its value.

Root metadata's required nested object members and JSON types are admitted during `tuf_schema`, including `keys`, each key and `keyval`, `roles`, each role and its key-ID array, and `cartulary`. Omission, explicit null, a wrong member type, an unknown nested member, or a role threshold outside the positive integer domain fails that check before canonical-byte or rotation comparisons. The signed-extension exception applies only to additional members of `signed`. Derived key-ID equality, supported cryptographic algorithms, sorted unique role references, threshold sufficiency and repository identity remain root-trust or root-rotation policy checks. Bootstrap admission maps rejection to `tuf_root_untrusted` as specified in §10.3. Typed TUF role projections close the outer envelope, signature objects, root keys and roles, and metadata/target descriptors. Only `signed` admits extension members; `targets.signed.delegations` remains forbidden. Metadata schema diagnostics enumerate original files in lexicographic path order under the virtual `$.metadata[index]` array. Missing required role files use their fixed logical path. Dynamic map keys use `["<unknown>"]`; unknown signed extensions never become diagnostic vocabulary. Signature ordering is checked after the complete typed metadata object is structurally admitted. After shape admission, descriptor maps are diagnosed as virtual arrays in ascending key order: `signed.meta[index]` and `signed.targets[index]`. Missing inventory declarations use `$.archive_members[index]`, ordered by non-metadata regular-file path. These indexes identify a bounded location without echoing an arbitrary key. Every invalid supplied signature is diagnosed at its signature index; an insufficient threshold adds the role envelope’s `.signatures` finding. Identical threshold findings coalesce. All ordinary roles are examined during the winning check. Root rotation compares consumed historical roots independently. After the first broken unconsumed transition or gap, later transition thresholds are inapplicable because their authenticated predecessor is unavailable; they MUST NOT be evaluated against a guessed root.

**Table 12-A. Validation phase tokens and order**

| Rank | Token |
| ---: | --- |
| 1 | `container_admission` |
| 2 | `archive_preflight` |
| 3 | `archive_structure` |
| 4 | `trust_selection` |
| 5 | `tuf_metadata` |
| 6 | `target_inventory` |
| 7 | `manifest_schema` |
| 8 | `logical_identity` |
| 9 | `payload_integrity` |
| 10 | `content_schema` |
| 11 | `content_semantics` |
| 12 | `runtime_compatibility` |
| 13 | `dependency_validation` |
| 14 | `type_registry_compatibility` |
| 15 | `index_construction` |
| 16 | `publication` |

**Table 12-B. Ordered content check registry**

Every check has severity `error` and public family `reference_pack_verification_failed`. The `code` is also its public reason. Applicability follows distribution kind and operation; an inapplicable check is skipped. Every issue uses the closed safe-details shape in RP-REQ-092. Resource guards run during bounded streaming and stop unsafe parsing immediately under RP-REQ-246; simultaneous breached guards use their registry rank. A resource abort does not authorize continued hostile parsing to discover other issues.

The typed registry MUST declare one applicability token per check. `attempt_bytes` applies to the admitted attempt's referenced bytes, including an absent source; `available_container` requires available source bytes; `admitted_archive` requires the bounded, safely extracted archive; `present_bounded_manifest` requires a present manifest at or below 1048576 bytes; `present_manifest` requires its presence; `admitted_manifest` requires the complete manifest schema and canonical bytes; `admitted_type_registry` additionally requires a type-registry profile. These tokens do not permit caller-controlled disabling. The live verification program MUST bind every registry check before execution, reject missing, nil, duplicate or unknown implementations, execute only in registry order, and stop after the winning check's complete safely enumerable findings. Dependent programs may initialize lazily when their first rank is reached, but must implement their whole declared range.

If `manifest.json` is absent, manifest-dependent ranks 28–35 are inapplicable and `member_missing` reports its fixed logical path, unless an earlier trust or inventory check fails. In particular, a declared TUF target with absent bytes still fails `target_extra` first. A manifest exceeding its byte ceiling is never decoded: ranks 28–30 are inapplicable and `manifest_schema` reports the byte guard. A bounded present manifest's raw UTF-8, BOM and escaped-surrogate checks belong to `manifest_encoding`. Its whole-document JSON syntax and top-level object requirement belong to `manifest_json`, before any duplicate verdict. No partial or repaired typed manifest is constructed.

For the current sixteen closed data profiles, active role/media declarations are already excluded by manifest schema and unrecognized logical members by inventory admission. Therefore a complete admitted container cannot first fail `disallowed_content` in this release. The check MUST still assert the complete data-only role/media allowlist: payload `application/x-ndjson` and notice `text/plain`. Inert command/query examples, literal markup, and other profile text are never interpreted or guessed to be executable programs. A future profile adding another media role must amend this applicability argument, the typed allowlist and its acceptance evidence together. Producer prohibitions on actual secrets remain distinct from guessed content classification.


| Rank | Check ID | Phase | Code |
| ---: | --- | --- | --- |
| 1 | `retained_payload` | `container_admission` | `payload_missing` |
| 2 | `container_format` | `container_admission` | `unsupported_container_format` |
| 3 | `container_bytes` | `archive_preflight` | `container_bytes_exceeded` |
| 4 | `archive_structure` | `archive_preflight` | `archive_structure_invalid` |
| 5 | `archive_member_count` | `archive_preflight` | `archive_member_count_exceeded` |
| 6 | `archive_extracted_bytes` | `archive_preflight` | `archive_extracted_bytes_exceeded` |
| 7 | `archive_compression_ratio` | `archive_preflight` | `archive_compression_ratio_exceeded` |
| 8 | `archive_path` | `archive_structure` | `path_traversal` |
| 9 | `archive_collision` | `archive_structure` | `path_collision` |
| 10 | `archive_member_type` | `archive_structure` | `disallowed_member_type` |
| 11 | `bundle_shape` | `trust_selection` | `bundle_hint_invalid` |
| 12 | `bundle_canonical` | `trust_selection` | `bundle_hint_noncanonical` |
| 13 | `repository_selection` | `trust_selection` | `tuf_root_untrusted` |
| 14 | `tuf_schema` | `tuf_metadata` | `tuf_metadata_invalid` |
| 15 | `tuf_canonical` | `tuf_metadata` | `metadata_noncanonical` |
| 16 | `root_trust` | `tuf_metadata` | `tuf_root_untrusted` |
| 17 | `root_rotation` | `tuf_metadata` | `tuf_root_rotation_invalid` |
| 18 | `role_signature` | `tuf_metadata` | `signature_threshold_not_met` |
| 19 | `metadata_rollback` | `tuf_metadata` | `metadata_rollback_detected` |
| 20 | `metadata_integrity` | `tuf_metadata` | `missing_integrity_metadata` |
| 21 | `metadata_links` | `tuf_metadata` | `metadata_mix_and_match_detected` |
| 22 | `metadata_expiry` | `tuf_metadata` | `metadata_expired` |
| 23 | `metadata_expiry_policy` | `tuf_metadata` | `metadata_expiry_policy_invalid` |
| 24 | `target_missing` | `target_inventory` | `target_not_declared` |
| 25 | `target_extra` | `target_inventory` | `unexpected_target` |
| 26 | `target_length` | `target_inventory` | `target_length_mismatch` |
| 27 | `target_hash` | `target_inventory` | `checksum_mismatch` |
| 28 | `manifest_encoding` | `manifest_schema` | `manifest_encoding_invalid` |
| 29 | `manifest_json` | `manifest_schema` | `manifest_json_invalid` |
| 30 | `manifest_duplicate` | `manifest_schema` | `duplicate_object_member` |
| 31 | `manifest_schema` | `manifest_schema` | `manifest_schema_invalid` |
| 32 | `manifest_canonical` | `manifest_schema` | `manifest_noncanonical` |
| 33 | `contract` | `logical_identity` | `contract_incompatible` |
| 34 | `sequence_rollback` | `logical_identity` | `pack_release_sequence_rollback` |
| 35 | `sequence_collision` | `logical_identity` | `pack_release_sequence_collision` |
| 36 | `member_missing` | `payload_integrity` | `required_member_missing` |
| 37 | `member_extra` | `payload_integrity` | `undeclared_member` |
| 38 | `member_length` | `payload_integrity` | `target_length_mismatch` |
| 39 | `member_hash` | `payload_integrity` | `checksum_mismatch` |
| 40 | `logical_collision` | `payload_integrity` | `pack_version_collision` |
| 41 | `target_binding` | `payload_integrity` | `metadata_mix_and_match_detected` |
| 42 | `content_schema` | `content_schema` | `content_schema_invalid` |
| 43 | `disallowed_content` | `content_schema` | `disallowed_content` |
| 44 | `content_semantics` | `content_semantics` | `content_semantic_invalid` |
| 45 | `runtime_compatibility` | `runtime_compatibility` | `contract_incompatible` |
| 46 | `dependencies` | `dependency_validation` | `dependency_unsatisfied` |
| 47 | `dependency_cycle` | `dependency_validation` | `dependency_cycle` |
| 48 | `conflicts` | `dependency_validation` | `pack_conflict` |
| 49 | `type_registry` | `type_registry_compatibility` | `type_registry_incompatible` |

**RP-REQ-093**
Issue ordering MUST be:

1. check rank from Table 12-B;
2. `path` ascending UTF-8 bytes;
3. `entry_id`, with `null` before strings;
4. `issue_id` ascending.

Checks MUST execute in registry order. The first failing check controls the result. Only that check's deterministically enumerable findings are collected, coalesced, counted, sorted, and retained; later checks do not execute. Phase or reason-table row order MUST NOT provide a competing precedence rule.

Within `manifest_json`, the complete bounded document's syntax and Unicode MUST be admitted before `manifest_duplicate` can select a verdict. Duplicate members are compared by decoded member name. Duplicate findings use the declared member's schema path; an undeclared subtree collapses to its first fixed unknown-member segment, including duplicates nested below that segment. Repeated occurrences at an identical diagnostic path coalesce. No repaired object becomes an admitted manifest.

`content_semantics` enumerates independent predicates across every schema-admitted row, every notice, and the manifest license. Row paths use the zero-based `entries`, `objects`, or `relationships` collections and declared field names; array findings name the offending element. Sortedness and uniqueness findings identify occurrences after the first declaration. A row with another semantic failure still contributes its schema-admitted identity and aliases to cross-row checks. Replacement-graph findings use the virtual collection path and the source entry identifier; duplicate source declarations use the first declared edge for that graph check and independently fail row identity. One source finding covers an absent target, cycle, or chain longer than 32 nodes. Relationship endpoint checks remain independently enumerable even when the relationship digest is invalid. An invalid license expression has one expression finding; binding order and notice membership remain independently enumerable, while expression-to-binding completeness is evaluated only for a valid expression. Notice validity has one finding per notice file. Multiple predicate failures producing an identical issue coalesce.

Runtime and manifest compatibility findings identify every independently mismatching field. Registry compatibility enumerates every schema-admitted row's known policy fields. An unknown indicator type reports its identifier and independently checkable deprecation/replacement fields; it MUST NOT fabricate expected algorithms. Missing required registry identities are reported using the collection path and the known missing identifier. None of these passes writes a consumer index or emits payload-derived field values.


**RP-REQ-094**
A validation summary MUST conform to closed `cartulary.reference_pack_validation_summary.v1` and contain exactly `schema_id`, `result`, `primary_issue_id`, `issues_truncated`, `total_issue_count`, `retained_issue_count`, and `issues`. `schema_id` is exact; `result` is `succeeded` or `failed`; `issues` is ordered by RP-REQ-093. Success has no issues, zero counts, `issues_truncated=false`, and `primary_issue_id=null`. Failure names the first retained issue as `primary_issue_id`. At most 1000 issues are retained. If more exist, the implementation MUST retain the first 1000 and set `issues_truncated=true`, `total_issue_count=<deduplicated count for the failed check>`, and `retained_issue_count=1000`. If no truncation occurs, `issues_truncated=false` and both counts equal the retained count. Execution aborts do not manufacture a content-validation summary.

Retained summaries use canonical JSON without a trailing line feed and are bounded to 16,777,216 bytes. Their reference is `rpvs_` followed by lowercase SHA-256 of those canonical bytes. An attempt binds the exact summary to its verdict before semantic finalization; only a committed content verdict is publicly readable. The successful prefix and exact deduplicated count MUST be independent of finding traversal order. Implementations MUST bound memory while counting all enumerable findings; any scratch storage remains private to the attempt and is removed on completion or abort. A scratch read, write, close or cleanup failure is an execution failure, not evidence that content is invalid.

Diagnostic member names come only from declared schemas and the virtual collection members `archive_members`, `entries`, `metadata`, `objects`, and `relationships`. Limit IDs come from the typed limit registry. Non-numeric detail tokens come from the closed diagnostic vocabulary projection; count and byte-measurement tokens use unsigned 64-bit decimal strings (`0..18446744073709551615`), without leading zeroes. These strings do not widen the JSON-number or diagnostic-path-index safe-integer domain. Entry identifiers are limited to 512 UTF-8 bytes and must satisfy the profile's identifier grammar before attribution. Unknown names coalesce at the fixed unknown-member segment when their resulting issue objects are identical.

**RP-REQ-095**
Before issue counts are calculated, byte-identical issue objects MUST coalesce. `issue_id` is unique only within one validation summary and MUST equal:

```text
"rpi_" + lowercase_hex(
  SHA256(
    RFC8785_JCS({check_id, severity, phase, code, path, entry_id, safe_details})
  )
)
```

**RP-REQ-096**
Safe issue details, job summaries, logs, telemetry, and administrative audit MUST NOT contain payload values, entry descriptions, raw signatures, private keys, source credentials, filesystem paths, object-store keys, or incident data.

**RP-REQ-097**
Cancellation MUST be checked at least between every phase in RP-REQ-089 and during bounded streaming loops. Before proven semantic commit, cancellation follows the adopted Extensions outcome classification and RP-REQ-091. An initially staged candidate becomes terminal failed without a content verdict; an established version retains its prior health. A proven committed result remains authoritative.

**RP-REQ-098**
Verification timeout MUST use `limits.reference_packs.max_verification_seconds`. The attempt's monotonic budget starts immediately before verification begins and includes verification, indexing, publication-lock waiting, and commit classification. Jobs queueing precedes this boundary and does not consume this budget. Refresh has one budget for the whole cohort. Expiry is `now >= deadline`. Parent deadlines, checked arithmetic, cancellation grace, proven-commit precedence, and indeterminate-commit fatal handling follow the adopted Extensions `extension_deadline_v1` contract. A proven absent publication on timeout produces `reference_pack_verification_failed/verification_timeout`, publishes no active-set or index change, and cleans temporary state. A proven committed result remains authoritative.

# 13. Content-profile registry and common payload contracts

**RP-REQ-099**
A pack is content-compatible only when the exact pair `(pack_key, content_profile_id)` appears in Table 13-A and `content_profile_version` equals the registered value. A later adopted NLSpec revision MAY add rows without changing the pack contract major only when existing rows, bytes, algorithms, and consumer behavior remain unchanged. Omission behavior: an unregistered pair fails with `contract_incompatible`.

**Table 13-A. Current content-profile registry**

| Pack key | Pack kind | Content profile ID | Version | Payload shape | Authority class |
| --- | --- | --- | --- | --- | --- |
| `type_registry.host` | `type_registry` | `cartulary.reference_pack.type_registry.host.v1` | `1` | entries | `registry` |
| `type_registry.evidence` | `type_registry` | `cartulary.reference_pack.type_registry.evidence.v1` | `1` | entries | `registry` |
| `type_registry.indicator` | `type_registry` | `cartulary.reference_pack.type_registry.indicator.v1` | `1` | entries | `registry` |
| `framework.attack` | `framework` | `cartulary.reference_pack.framework.attack.v1` | `1` | objects and relationships | `framework_reference` |
| `framework.d3fend` | `framework` | `cartulary.reference_pack.framework.d3fend.v1` | `1` | objects and relationships | `framework_reference` |
| `framework.veris` | `framework` | `cartulary.reference_pack.framework.veris.v1` | `1` | objects and relationships | `framework_reference` |
| `enrichment.tor` | `enrichment` | `cartulary.reference_pack.enrichment.tor.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.cisa_kev` | `enrichment` | `cartulary.reference_pack.enrichment.cisa_kev.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.ms_portals` | `enrichment` | `cartulary.reference_pack.enrichment.ms_portals.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.windows_event_ids` | `enrichment` | `cartulary.reference_pack.enrichment.windows_event_ids.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.entra_app_ids` | `enrichment` | `cartulary.reference_pack.enrichment.entra_app_ids.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.lolbas` | `enrichment` | `cartulary.reference_pack.enrichment.lolbas.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.loldrivers` | `enrichment` | `cartulary.reference_pack.enrichment.loldrivers.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.lolesxi` | `enrichment` | `cartulary.reference_pack.enrichment.lolesxi.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.hijacklibs` | `enrichment` | `cartulary.reference_pack.enrichment.hijacklibs.v1` | `1` | entries | `advisory_enrichment` |
| `enrichment.windows_sids` | `enrichment` | `cartulary.reference_pack.enrichment.windows_sids.v1` | `1` | entries | `advisory_enrichment` |

The registry reflects the current Core pack keys and the locally staged dataset families observed in the Kanvas research. That research also identifies missing checksum and signature verification as a supply-chain weakness, which this NLSpec closes through mandatory verification.[^6]

**RP-REQ-100**
A profile's `pack_kind` MUST equal the exact value in Table 13-A. The public Core resource may continue to serialize `pack_kind` as an open string, but this v1 registry is closed for conformant content profiles.

## 13.1 Common NDJSON

**RP-REQ-101**
Every NDJSON file MUST:

- use UTF-8 without BOM;
- contain one RFC 8785 canonical JSON object per line;
- use LF as the only line separator;
- terminate every line, including the last, with LF;
- contain no blank line or comment;
- contain no line longer than 1048576 bytes including LF;
- contain exactly the count declared by `content_summary`.

**RP-REQ-102**
`entries.ndjson` objects MUST sort by the profile's canonical entry identity as ascending UTF-8 bytes. `objects.ndjson` objects MUST sort by `object_id`. `relationships.ndjson` objects MUST sort by `relationship_type`, `source_object_id`, `target_object_id`, then `relationship_id`.

**RP-REQ-103**
Duplicate entry IDs, object IDs, or relationship IDs are invalid. Two different IDs that normalize to the same profile lookup identity are also invalid unless the profile explicitly classifies one as a non-identity alias and alias resolution remains unambiguous.

## 13.2 Common entry members

**RP-REQ-104**
Every entries-profile object MUST include the common members in Table 13-B plus the profile-specific members in §§14 and 16. Unknown members are invalid.

**Table 13-B. Common entries-profile members**

| Member | Type | Null allowed | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | No | Exact profile entry schema ID. |
| `entry_id` | string | No | Profile-owned canonical identity. |
| `display_label` | string | No | Single-line string, maximum 256 scalars. |
| `description` | string or null | Yes | RP-REQ-035. |
| `aliases` | string array | No | `0..64` items, profile-normalized, sorted, unambiguous. |
| `deprecated` | boolean | No | Default materialized value `false`. |
| `replacement_entry_id` | string or null | Yes | Non-null only when `deprecated=true`. |
| `source_refs` | array | No | `1..64` source-ref objects. |
| `extensions` | object | No | RP-REQ-038. |

**RP-REQ-105**
A source-ref object MUST contain exactly:

| Member | Type | Rule |
| --- | --- | --- |
| `artifact_index` | integer | Index in `manifest.source_artifacts[]`, domain `0..63`. |
| `locator` | string | Non-empty single-line source locator, maximum 1024 scalars. |

`artifact_index` MUST be less than the actual length of `manifest.source_artifacts[]`. Source refs MUST sort by `artifact_index`, then `locator`. Duplicate source refs are invalid.

**RP-REQ-106**
If `replacement_entry_id` is non-null, the replacement must exist in the same pack, must not equal the current entry, and must not resolve through a replacement cycle. A replacement chain longer than 32 entries is invalid.

**RP-REQ-107**
Aliases MUST be sorted by the profile's alias-normalization output, then exact alias bytes. Duplicate normalized alias keys are invalid both within one entry and across entries, even if they would resolve to the same entry. The indicator registry uses `registry_alias_v1`: RP-REQ-118 trimming, then NFC, then `ascii_lower_v1`; empty output is invalid. Other profiles retain their explicitly named alias algorithms. Default `deprecated=false` is a producer materialization rule; an omitted `deprecated` member is invalid canonical content.

## 13.3 Common framework object and relationship members

**RP-REQ-108**
Every framework object MUST be a closed object containing exactly:

| Member | Type | Null allowed | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | No | Profile-specific object schema ID. |
| `object_id` | string | No | Source-stable ASCII identity matching `^[A-Za-z0-9][A-Za-z0-9._:-]{0,191}$`. |
| `object_type` | string | No | Source-stable token matching `^[A-Za-z][A-Za-z0-9._-]{0,63}$`. |
| `display_label` | string | No | Common display label. |
| `description` | string or null | Yes | Common description. |
| `aliases` | string array | No | `0..64`, sorted by ASCII case-folded alias then exact bytes. |
| `deprecated` | boolean | No | Materialized default `false`. |
| `replacement_object_id` | string or null | Yes | Same rules as replacement entries. |
| `external_refs` | array | No | `0..64` objects under RP-REQ-109. |
| `source_refs` | array | No | `1..64` common source refs. |
| `extensions` | object | No | Common extensions. |

**RP-REQ-109**
A framework `external_refs[]` item contains exactly `source_name`, `external_id`, and nullable `url`. `source_name` and `external_id` are non-empty single-line strings. `url` is either `null` or the canonical output of RP-REQ-124 with scheme `https`, no userinfo, and no fragment. Items sort by `source_name`, then `external_id`, then `url` with `null` first. Duplicates are invalid. Reference uniqueness uses the complete tuple. Different references MAY share an external ID. Derived lookup keys form a set per object and exact `(lookup_kind, lookup_value)`; repeated keys coalesce in canonical first-occurrence order without changing canonical content or provenance. Lookup returns each matching object once, while different objects sharing an external ID remain separate results.

**RP-REQ-110**
Every framework relationship MUST be a closed object containing exactly:

| Member | Type | Null allowed | Rule |
| --- | --- | ---: | --- |
| `schema_id` | string | No | Profile-specific relationship schema ID. |
| `relationship_id` | string | No | Output of `reference_pack_relationship_id_v1`. |
| `relationship_type` | string | No | Token matching `^[A-Za-z][A-Za-z0-9._-]{0,63}$`. |
| `source_object_id` | string | No | Existing object ID. |
| `target_object_id` | string | No | Existing object ID. |
| `description` | string or null | Yes | Common description. |
| `source_refs` | array | No | `1..64` common source refs. |
| `extensions` | object | No | Common extensions. |

**RP-REQ-111**
`reference_pack_relationship_id_v1` MUST compute:

```text
"rpr_" + lowercase_hex(
  SHA256(
    ASCII(content_profile_id) || NUL ||
    UTF8(relationship_type) || NUL ||
    UTF8(source_object_id) || NUL ||
    UTF8(target_object_id) || NUL ||
    RFC8785_JCS(source_refs)
  )
)
```

The declared `relationship_id` must equal the computed value.

# 14. Type-registry content profiles and indicator algorithms

## 14.1 Host and evidence registries

**RP-REQ-112**
A `type_registry.host` or `type_registry.evidence` entry MUST include the common members and exactly:

| Member | Type | Rule |
| --- | --- | --- |
| `category` | string | Token matching `^[a-z][a-z0-9_]{0,63}$`. |
| `icon_key` | string | Token matching `^[a-z][a-z0-9_.-]{0,127}$`. |

The schema IDs are `cartulary.reference_pack.type_registry.host.entry.v1` and `cartulary.reference_pack.type_registry.evidence.entry.v1` respectively.

**RP-REQ-113**
Host and evidence registry `entry_id` values MUST match `^[a-z][a-z0-9_]{0,63}$`. Identity and alias lookup use exact ASCII lowercase. A supplied alias containing an uppercase ASCII letter is invalid rather than silently case-folded in canonical pack bytes.

**RP-REQ-114**
Every packaged host and evidence registry MUST contain a non-deprecated `entry_id='unknown'`. `unknown` MUST have no replacement. This entry permits rough capture when no narrower type has been selected.

**RP-REQ-115**
The exact additional built-in host and evidence entries are release fixtures, not hard-coded implementation constants. Two conforming implementations of the same application release MUST consume byte-identical packaged built-in pack fixtures and therefore expose the same registry entries and digests.

## 14.2 Indicator registry schema

**RP-REQ-116**
A `type_registry.indicator` entry MUST include the common members and exactly:

| Member | Type | Rule |
| --- | --- | --- |
| `allowed_value_kinds` | string array | Non-empty subset of `atomic`, `pattern`, `reference`, in that order. |
| `normalization_algorithm_id` | string | Exact ID from Table 14-A. |
| `validation_algorithm_id` | string | Exact ID from Table 14-A. |
| `defang_algorithm_id` | string | Exact ID from Table 14-A. |
| `dedupe_algorithm_id` | string | Exact ID from Table 14-A. |
| `stix_mapping` | object or null | Table 14-B. |

The schema ID is `cartulary.reference_pack.type_registry.indicator.entry.v1`.

**RP-REQ-117**
The indicator registry MUST contain exactly the nine current Core indicator IDs in Table 14-A. Every one MUST have `deprecated=false`, `replacement_entry_id=null`, the exact allowed-value-kind list, and the exact four algorithm IDs shown in that table. An additional indicator type requires a Core 02 token-registry amendment before this NLSpec may register it.

**Table 14-A. Indicator type and algorithm registry**

| `entry_id` | Allowed value kinds | Normalization | Validation | Defang | Dedupe |
| --- | --- | --- | --- | --- | --- |
| `ipv4_addr` | `atomic` | `cartulary.indicator.normalize.ipv4.v1` | `cartulary.indicator.validate.ipv4.v1` | `cartulary.indicator.defang.ipv4.v1` | `cartulary.indicator.dedupe.normalized.v1` |
| `ipv6_addr` | `atomic` | `cartulary.indicator.normalize.ipv6.v1` | `cartulary.indicator.validate.ipv6.v1` | `cartulary.indicator.defang.ipv6.v1` | `cartulary.indicator.dedupe.normalized.v1` |
| `domain_name` | `atomic` | `cartulary.indicator.normalize.domain_ascii.v1` | `cartulary.indicator.validate.domain_ascii.v1` | `cartulary.indicator.defang.domain.v1` | `cartulary.indicator.dedupe.normalized.v1` |
| `url` | `atomic`, `reference` | `cartulary.indicator.normalize.http_url.v1` | `cartulary.indicator.validate.http_url.v1` | `cartulary.indicator.defang.http_url.v1` | `cartulary.indicator.dedupe.normalized.v1` |
| `sha256` | `atomic` | `cartulary.indicator.normalize.sha256.v1` | `cartulary.indicator.validate.sha256.v1` | `cartulary.indicator.defang.identity.v1` | `cartulary.indicator.dedupe.sha256.v1` |
| `email_addr` | `atomic` | `cartulary.indicator.normalize.email_ascii.v1` | `cartulary.indicator.validate.email_ascii.v1` | `cartulary.indicator.defang.email.v1` | `cartulary.indicator.dedupe.normalized.v1` |
| `registry_key` | `atomic`, `reference` | `cartulary.indicator.normalize.windows_registry_key.v1` | `cartulary.indicator.validate.windows_registry_key.v1` | `cartulary.indicator.defang.identity.v1` | `cartulary.indicator.dedupe.normalized.v1` |
| `process_name` | `atomic`, `reference` | `cartulary.indicator.normalize.process_name.v1` | `cartulary.indicator.validate.process_name.v1` | `cartulary.indicator.defang.identity.v1` | `cartulary.indicator.dedupe.normalized.v1` |
| `text` | `atomic`, `pattern`, `reference` | `cartulary.indicator.normalize.text.v1` | `cartulary.indicator.validate.text.v1` | `cartulary.indicator.defang.identity.v1` | `cartulary.indicator.dedupe.text_hash.v1` |

**Table 14-B. `stix_mapping` object**

| Member | Type | Rule |
| --- | --- | --- |
| `stix_object_type` | string | Exact STIX object type token. |
| `pattern_property` | string | Exact STIX pattern property path. |

The required mappings are:

| Indicator type | Mapping |
| --- | --- |
| `ipv4_addr` | `ipv4-addr`, `value` |
| `ipv6_addr` | `ipv6-addr`, `value` |
| `domain_name` | `domain-name`, `value` |
| `url` | `url`, `value` |
| `sha256` | `file`, `hashes.'SHA-256'` |
| `email_addr` | `email-addr`, `value` |
| `registry_key` | `windows-registry-key`, `key` |
| `process_name` | `process`, `name` |
| `text` | `null` |

`stix_mapping` MUST NOT affect canonical identity.

## 14.3 Indicator algorithm common contract

**RP-REQ-118**
`indicator_input_trim_v1` MUST remove leading and trailing code points from this exact set and no others:

```text
U+0009 U+000A U+000B U+000C U+000D U+0020 U+0085 U+00A0 U+1680
U+2000..U+200A U+2028 U+2029 U+202F U+205F U+3000
```

After trimming, algorithms that permit Unicode MUST apply NFC.

**RP-REQ-119**
Every indicator normalization algorithm accepts one JSON string of at most 8192 Unicode scalar values and returns either one canonical string or one validation code from this closed registry:

| Family | Validation code |
| --- | --- |
| Common input bound | `indicator_input_too_long` |
| IPv4 | `invalid_ipv4` |
| IPv6 | `invalid_ipv6` |
| Domain | `invalid_domain_name` |
| URL | `invalid_http_url` |
| SHA-256 | `invalid_sha256` |
| Email | `invalid_email_addr` |
| Registry key | `invalid_registry_key` |
| Process name | `invalid_process_name` |
| Text | `invalid_text` |

It MUST NOT perform network access, filesystem access, locale-sensitive comparison, DNS resolution, or source-record mutation.

**RP-REQ-120**
`EvaluateIndicatorValue` MUST run normalization, then validation, then defanging, then dedupe. On success, `display_value` and `normalized_value` both equal the canonical normalization output. If normalization or validation fails, `display_value`, `normalized_value`, defanging output, and dedupe output are `null`.

## 14.4 Family-specific indicator algorithms

**RP-REQ-121**
`cartulary.indicator.normalize.ipv4.v1` and `validate.ipv4.v1` MUST implement Core 02's exact dotted-decimal IPv4 rules: four octets, range `0..255`, no sign, no empty octet, and no leading zero except `0`. Output is shortest dotted-decimal text.

**RP-REQ-122**
`cartulary.indicator.normalize.ipv6.v1` and `validate.ipv6.v1` MUST reject zone IDs, brackets, ports, CIDR suffixes, dotted-quad suffixes, and IPv4-mapped or IPv4-compatible forms. IPv4-compatible means the `::/96` range other than the unspecified `::` and loopback `::1` addresses; those two addresses remain valid. Output MUST use RFC 5952 lowercase compression, longest zero run, leftmost tie, and no leading hextet zeroes.

**RP-REQ-123**
`cartulary.indicator.normalize.domain_ascii.v1` MUST:

1. apply `indicator_input_trim_v1`;
2. require ASCII only;
3. remove exactly one trailing `.` when present;
4. lowercase ASCII letters;
5. require total length `1..253` bytes;
6. require labels of `1..63` bytes separated by `.`;
7. permit only letters, digits, and `-` in each label;
8. reject a label beginning or ending with `-`;
9. accept `xn--` labels as ordinary ASCII labels and perform no IDNA conversion.

Validation applies the same rules to canonical output.

**RP-REQ-124**
`cartulary.indicator.normalize.http_url.v1` MUST accept only absolute `http` or `https` URIs and MUST:

1. apply `indicator_input_trim_v1`;
2. before splitting or normalizing components, require ASCII URI input and reject every remaining raw space, C0 control, DEL, or code point outside RFC 3986 URI syntax; component parsing MUST NOT trim internal whitespace, including inside IPv6 brackets;
3. reject userinfo;
4. lowercase the scheme;
5. canonicalize bracketed hosts as IPv6; otherwise remove exactly one trailing dot, classify four non-empty decimal labels as IPv4 (reject invalid octets rather than falling back to domain), and classify all other hosts as domains;
6. enclose a canonical IPv6 host in `[` and `]` in URL output;
7. remove port `80` for HTTP and `443` for HTTPS;
8. preserve any other decimal port in `1..65535` without leading zero;
9. replace an empty path with `/`;
10. decode percent-encoded unreserved ASCII characters in the path before removing RFC 3986 dot segments;
11. uppercase hexadecimal digits in percent encodings;
12. decode percent-encoded unreserved ASCII characters in query and fragment, where the unreserved set is exactly `A-Z`, `a-z`, `0-9`, `-`, `.`, `_`, and `~`; encoded reserved delimiters remain encoded;
13. preserve query member order, duplicate query members, and fragment text;
14. preserve the presence of an explicitly supplied empty query delimiter `?` or empty fragment delimiter `#`;
15. apply the percent-encoding rule independently to path, query, and fragment;
16. emit `scheme://host[:port]/path[?query][#fragment]`.

A malformed percent encoding or forbidden component is invalid. Canonical URL output MUST contain at most 8192 ASCII bytes. Normalization MUST be idempotent, including paths containing encoded dot segments.

**RP-REQ-125**
`cartulary.indicator.normalize.sha256.v1` MUST apply `indicator_input_trim_v1`, require exactly 64 hexadecimal digits, and lowercase ASCII `A..F`. Validation requires exactly 64 lowercase hexadecimal digits.

**RP-REQ-126**
`cartulary.indicator.normalize.email_ascii.v1` MUST accept only the ASCII dot-atom subset:

1. apply `indicator_input_trim_v1`;
2. require exactly one `@`;
3. preserve local-part letter case;
4. require local length `1..64` bytes;
5. permit ASCII letters, ASCII digits, and exactly these punctuation code points: `U+0021`, `U+0023..U+0027`, `U+002A`, `U+002B`, `U+002D..U+002F`, `U+003D`, `U+003F`, `U+005E`, `U+005F`, `U+0060`, `U+007B..U+007E`;
6. reject leading, trailing, or consecutive `.` in the local part;
7. canonicalize the domain with RP-REQ-123;
8. require total output length at most 254 bytes.

Quoted local parts, comments, display names, and internationalized local parts are invalid.

**RP-REQ-127**
`cartulary.indicator.normalize.windows_registry_key.v1` MUST:

1. apply `indicator_input_trim_v1` and NFC;
2. replace `/` with `\`;
3. collapse consecutive `\` separators;
4. remove one trailing separator, including after a root token; canonical roots contain no trailing separator;
5. map `HKLM`, `HKCU`, `HKCR`, `HKU`, and `HKCC` to `HKEY_LOCAL_MACHINE`, `HKEY_CURRENT_USER`, `HKEY_CLASSES_ROOT`, `HKEY_USERS`, and `HKEY_CURRENT_CONFIG`;
6. require one recognized root;
7. reject NUL and C0/C1 controls;
8. uppercase ASCII letters in the complete canonical output;
9. require output length `1..1024` scalars.

**RP-REQ-128**
`cartulary.indicator.normalize.process_name.v1` MUST apply `indicator_input_trim_v1` and NFC, require `1..260` scalars, reject C0/C1 controls and `/` and `\`, and preserve case. Canonical identity is case-sensitive.

**RP-REQ-129**
`cartulary.indicator.normalize.text.v1` MUST apply `indicator_input_trim_v1`, NFC, and CRLF-or-CR to LF conversion. It permits LF and horizontal tab, rejects other C0/C1 controls, and requires `1..8192` scalars.

**RP-REQ-130**
Defang algorithms MUST produce:

| Algorithm | Output rule |
| --- | --- |
| `defang.ipv4.v1` | Replace every `.` with `[.]`. |
| `defang.ipv6.v1` | Replace every `:` with `[:]`. |
| `defang.domain.v1` | Replace every `.` with `[.]`. |
| `defang.http_url.v1` | Replace leading `http` with `hxxp` or `https` with `hxxps`; defang domain and IPv4 host dots, or IPv6 host colons while retaining brackets; preserve all other canonical components. |
| `defang.email.v1` | Replace `@` with `[@]` and defang domain dots. |
| `defang.identity.v1` | Return the canonical value unchanged. |

**RP-REQ-131**
Dedupe algorithms MUST produce:

| Algorithm | Output rule |
| --- | --- |
| `dedupe.normalized.v1` | `indicator_type_id + ":" + normalized_value`. |
| `dedupe.sha256.v1` | `"sha256:" + normalized_value`. |
| `dedupe.text_hash.v1` | `"text:" + lowercase_hex(SHA256(UTF8(normalized_value)))`. |

# 15. Framework content profiles

**RP-REQ-132**
The three framework profiles use the common object and relationship contracts with these schema IDs:

| Pack key | Object schema ID | Relationship schema ID |
| --- | --- | --- |
| `framework.attack` | `cartulary.reference_pack.framework.attack.object.v1` | `cartulary.reference_pack.framework.attack.relationship.v1` |
| `framework.d3fend` | `cartulary.reference_pack.framework.d3fend.object.v1` | `cartulary.reference_pack.framework.d3fend.relationship.v1` |
| `framework.veris` | `cartulary.reference_pack.framework.veris.object.v1` | `cartulary.reference_pack.framework.veris.relationship.v1` |

**RP-REQ-133**
Framework object IDs, object types, relationship types, and external references are source-stable data. Consumers MUST treat unrecognized object or relationship types as generic reference data rather than executing type-specific behavior. Display labels MUST NOT define identity.

**RP-REQ-134**
Every relationship endpoint MUST resolve to an object in the same pack. Self-relationships are valid only when the source framework explicitly represents them and the relationship has at least one source ref.

**RP-REQ-135**
A framework producer MAY remove an object or relationship in a later pack version. It MUST NOT reuse the removed ID for a different subject or relationship meaning. This is a source-publication obligation: the producer reviews continuity against its source history. Runtime consumers identify an object by its exact pack set, member and object ID; they MUST NOT infer cross-version subject equivalence from an unqualified ID. Pinned historical pack sets preserve the old object.

**RP-REQ-136**
Framework lookup kinds are exactly:

```text
object_id
alias
external_id
```

`object_id` and `external_id` use exact comparison. `alias` uses ASCII case-insensitive comparison after NFC and rejects ambiguous matches.

**RP-REQ-137**
Framework results have authority class `framework_reference`. They may support report, diagram, suggestion, and analyst-pivot behavior only. They MUST NOT create a Base Profile workbook surface or mutate incident state.

# 16. Enrichment content profiles

## 16.1 Common enrichment behavior

**RP-REQ-138**
Every enrichment entry uses the common entry members. All enrichment results have authority class `advisory_enrichment`, include exact pack and source provenance, and have no automatic incident mutation effect.

**RP-REQ-139**
A no-hit lookup is successful and returns an empty result array. Pack unavailable, lookup kind unsupported, malformed lookup input, and no-hit are distinct outcomes.

## 16.2 Profile schemas and lookup normalization

**Table 16-A. Enrichment-specific members and lookup kinds**

| Pack key | Entry schema ID | Additional required members | Lookup kinds and normalization |
| --- | --- | --- | --- |
| `enrichment.tor` | `cartulary.reference_pack.enrichment.tor.entry.v1` | `network`, `address_family`, nullable `first_observed_at`, nullable `last_observed_at` | `ip_literal`: canonicalize as IPv4 or IPv6 and test membership; `network`: exact canonical CIDR. |
| `enrichment.cisa_kev` | `cartulary.reference_pack.enrichment.cisa_kev.entry.v1` | `cve_id`, `vendor_project`, `product`, `vulnerability_name`, `date_added`, `short_description`, `required_action`, `due_date`, `known_ransomware_campaign_use`, nullable `notes` | `cve_id`: ASCII uppercase `CVE-YYYY-NNNN...`. |
| `enrichment.ms_portals` | `cartulary.reference_pack.enrichment.ms_portals.entry.v1` | `service_id`, `portal_urls`, `audiences`, `clouds` | `service_id`: exact; `url_host`: canonical ASCII domain or IP host; `alias`: ASCII case-insensitive. |
| `enrichment.windows_event_ids` | `cartulary.reference_pack.enrichment.windows_event_ids.entry.v1` | `provider_name`, `event_id`, nullable `event_version`, nullable `level`, nullable `task`, nullable `opcode` | `provider_event_id`: ASCII-lower provider name, NUL, decimal event ID; `event_id`: exact integer. |
| `enrichment.entra_app_ids` | `cartulary.reference_pack.enrichment.entra_app_ids.entry.v1` | `app_id`, nullable `publisher`, `service_principal_names` | `app_id`: lowercase canonical UUID; `service_principal_name`: ASCII case-insensitive; `alias`: ASCII case-insensitive. |
| `enrichment.lolbas` | `cartulary.reference_pack.enrichment.lolbas.entry.v1` | `name`, `platforms`, `categories`, `usage_examples`, `references` | `entry_id`: exact; `name`: ASCII case-insensitive; `alias`: ASCII case-insensitive. |
| `enrichment.loldrivers` | `cartulary.reference_pack.enrichment.loldrivers.entry.v1` | `driver_id`, `filenames`, `hashes`, `signer_names`, `categories`, `references` | `driver_id`: exact; `filename`: ASCII case-insensitive basename; `hash`: lowercase exact by algorithm; `alias`: ASCII case-insensitive. |
| `enrichment.lolesxi` | `cartulary.reference_pack.enrichment.lolesxi.entry.v1` | `name`, `platforms`, `categories`, `usage_examples`, `references` | Same as LOLBAS. |
| `enrichment.hijacklibs` | `cartulary.reference_pack.enrichment.hijacklibs.entry.v1` | `library_name`, `candidate_paths`, `categories`, `usage_examples`, `references` | `library_name`: ASCII case-insensitive basename; `candidate_path`: Windows-path normalization without filesystem access. |
| `enrichment.windows_sids` | `cartulary.reference_pack.enrichment.windows_sids.entry.v1` | `sid`, `name`, `category`, `scope` | `sid`: canonical SID; `name`: ASCII case-insensitive; `alias`: ASCII case-insensitive. |

**RP-REQ-140**
Every enrichment-profile member named in Table 16-A is required. An array member MUST be a non-null array and MUST use `[]` when the profile allows no values; a nullable scalar member MUST use explicit JSON `null` when absent. Unless a profile rule narrows the bound:

- a profile-specific string array contains `0..64` non-null strings satisfying RP-REQ-034, sorts by `ascii_lower_v1(value)` then exact UTF-8 bytes, and rejects duplicate normalized values;
- a `references[]` array contains `1..64` strings equal to the canonical output of RP-REQ-124 with scheme `https`, no userinfo, and no fragment, sorts by canonical URI bytes, and rejects duplicates;
- a `usage_examples[]` array contains `0..64` non-empty inert multiline strings, each satisfying RP-REQ-035 and limited to 8192 scalars;
- an explicit JSON `null` is invalid for an array;
- an unknown additional member is invalid.

For a Table 16-A row whose lookup-kind cell names `alias`, `aliases[]` uses ASCII case-insensitive normalization under RP-REQ-034. For every other enrichment profile, `aliases` MUST equal `[]`. This rule is the complete enrichment alias-normalization registry.

**RP-REQ-141**
An `enrichment.tor` entry contains exactly the common entry members plus:

| Member | Type | Rule |
| --- | --- | --- |
| `network` | string | Canonical IPv4 or IPv6 CIDR with all host bits zero. |
| `address_family` | string | `ipv4` or `ipv6`, consistent with `network`. |
| `first_observed_at` | date or null | RP-REQ-033 calendar date. |
| `last_observed_at` | date or null | RP-REQ-033 calendar date. |

`entry_id` MUST equal `network`. IPv4 prefix length is `0..32`; IPv6 prefix length is `0..128`. If both dates are non-null, `first_observed_at` MUST be less than or equal to `last_observed_at`. `ip_literal` lookup canonicalizes the supplied address and returns every containing network sorted by prefix length descending, then `entry_id`; `network` lookup requires exact canonical CIDR and returns zero or one entry.

**RP-REQ-142**
An `enrichment.cisa_kev` entry contains exactly the common entry members plus:

| Member | Type | Rule |
| --- | --- | --- |
| `cve_id` | string | `^CVE-[0-9]{4}-[0-9]{4,}$`, at most 192 ASCII bytes. |
| `vendor_project` | string | Single-line, maximum 256 scalars. |
| `product` | string | Single-line, maximum 256 scalars. |
| `vulnerability_name` | string | Single-line, maximum 512 scalars. |
| `date_added` | date | RP-REQ-033 calendar date. |
| `short_description` | string | Non-empty multiline string, maximum 8192 scalars. |
| `required_action` | string | Non-empty multiline string, maximum 8192 scalars. |
| `due_date` | date | RP-REQ-033 calendar date; not earlier than `date_added`. |
| `known_ransomware_campaign_use` | string | `known`, `unknown`, or `not_reported`. |
| `notes` | string or null | RP-REQ-035. |

`entry_id` MUST equal `cve_id`. `cve_id` lookup uppercases ASCII letters before validation and returns zero or one entry.

**RP-REQ-143**
An `enrichment.ms_portals` entry contains exactly the common entry members plus:

| Member | Type | Rule |
| --- | --- | --- |
| `service_id` | string | `^[a-z][a-z0-9_.-]{0,127}$`; equals `entry_id`. |
| `portal_urls` | string array | `1..64` values equal to the RP-REQ-124 canonical output, with scheme `https`, no userinfo, and no fragment. |
| `audiences` | string array | `0..64` tokens matching `^[a-z][a-z0-9_.-]{0,63}$`. |
| `clouds` | string array | `0..64` tokens matching `^[a-z][a-z0-9_.-]{0,63}$`. |

`portal_urls` sort by canonical URI bytes. `audiences` and `clouds` sort by exact bytes. `url_host` lookup canonicalizes the supplied domain, IPv4, or IPv6 host and returns matching entries sorted by `service_id`; `service_id` is exact; `alias` uses ASCII case-insensitive comparison.

**RP-REQ-144**
An `enrichment.windows_event_ids` entry contains exactly the common entry members plus:

| Member | Type | Rule |
| --- | --- | --- |
| `provider_name` | string | ASCII single-line value, `1..256` bytes. |
| `event_id` | integer | `0..65535`. |
| `event_version` | integer or null | `0..255` when non-null. |
| `level` | string or null | Single-line, maximum 128 scalars. |
| `task` | string or null | Single-line, maximum 256 scalars. |
| `opcode` | string or null | Single-line, maximum 128 scalars. |

`entry_id` MUST equal `ascii_lower_v1(provider_name) + ":" + base10(event_id) + ":" + version_token`, where `version_token` is `*` for null or the shortest base-10 event version. Parse the last two colons from the right; provider names may contain colons. The maximum identity length is 266 UTF-8 bytes. `provider_event_id` lookup input is one closed object containing `provider_name`, `event_id`, and optional `event_version`; omitted `event_version` matches every version, explicit numeric version matches only that version and excludes null-version entries, and explicit `null` is invalid. Results sort by `event_version` with `null` first, then `entry_id`. `event_id` lookup returns every matching provider sorted by `ascii_lower_v1(provider_name)`, exact provider bytes, nullable version, then `entry_id`.

**RP-REQ-145**
An `enrichment.entra_app_ids` entry contains exactly the common entry members plus:

| Member | Type | Rule |
| --- | --- | --- |
| `app_id` | string | Canonical lowercase UUID with hyphens; equals `entry_id`. |
| `publisher` | string or null | Single-line, maximum 256 scalars. |
| `service_principal_names` | string array | `0..64` ASCII single-line values, each at most 1024 bytes. |

Service-principal names sort by `ascii_lower_v1(value)`, then exact bytes; duplicate normalized values are invalid. `app_id` lookup lowercases ASCII UUID hex before canonical validation. `service_principal_name` and `alias` lookup use ASCII case-insensitive comparison. Results sort by `entry_id`.

**RP-REQ-146**
LOLBAS and LOLESXi entries contain exactly the common entry members plus `name`, `platforms`, `categories`, `usage_examples`, and `references`. Their rules are:

| Member | Type | Rule |
| --- | --- | --- |
| `entry_id` | string | Source-stable ASCII ID matching `^[A-Za-z0-9][A-Za-z0-9._:-]{0,191}$`. |
| `name` | string | Single-line, maximum 256 scalars. |
| `platforms` | string array | `1..64` values under RP-REQ-140. |
| `categories` | string array | `0..64` values under RP-REQ-140. |
| `usage_examples` | string array | RP-REQ-140 inert-text contract. |
| `references` | string array | RP-REQ-140 HTTPS-reference contract. |

`entry_id` lookup is exact. `name` and `alias` use ASCII case-insensitive comparison. Usage examples are inert data and MUST NOT be executed, shell-expanded, rendered as active HTML, or passed to an interpreter.

**RP-REQ-147**
A LOLDrivers entry contains exactly the common entry members plus:

| Member | Type | Rule |
| --- | --- | --- |
| `driver_id` | string | Source-stable ASCII ID under the RP-REQ-146 `entry_id` grammar; equals `entry_id`. |
| `filenames` | string array | `1..64` ASCII basenames, each `1..260` bytes, containing no `/` or `\`. |
| `hashes` | object array | `1..64` hash objects. |
| `signer_names` | string array | `0..64` values under RP-REQ-140. |
| `categories` | string array | `0..64` values under RP-REQ-140. |
| `references` | string array | RP-REQ-140 HTTPS-reference contract. |

A hash object contains exactly `algorithm` and `value`. `algorithm` is `md5`, `sha1`, or `sha256`; `value` is lowercase hexadecimal of length 32, 40, or 64 respectively. Hashes sort by algorithm rank `md5`, `sha1`, `sha256`, then value and reject duplicates. Filenames sort by `ascii_lower_v1(value)`, then exact bytes and reject duplicate normalized values. `driver_id` lookup is exact; `filename` and `alias` are ASCII case-insensitive. `hash` lookup requires one closed object containing exact members `algorithm` and `value`, normalizes and validates them under the hash-object rule above, and returns entries sorted by `driver_id`.

**RP-REQ-148**
HijackLibs and Windows SID entries use these exact additional contracts:

| Pack | Additional member | Type and rule |
| --- | --- | --- |
| `enrichment.hijacklibs` | `entry_id` | Source-stable ASCII ID under the RP-REQ-146 grammar. |
| `enrichment.hijacklibs` | `library_name` | ASCII basename, `1..260` bytes, no `/` or `\`. |
| `enrichment.hijacklibs` | `candidate_paths` | `1..64` non-empty single-line paths, each at most 1024 scalars. |
| `enrichment.hijacklibs` | `categories` | `0..64` values under RP-REQ-140. |
| `enrichment.hijacklibs` | `usage_examples` | RP-REQ-140 inert-text contract. |
| `enrichment.hijacklibs` | `references` | RP-REQ-140 HTTPS-reference contract. |
| `enrichment.windows_sids` | `sid` | Canonical SID text; equals `entry_id`. |
| `enrichment.windows_sids` | `name` | Single-line, maximum 256 scalars. |
| `enrichment.windows_sids` | `category` | Token matching `^[a-z][a-z0-9_]{0,63}$`. |
| `enrichment.windows_sids` | `scope` | Token matching `^[a-z][a-z0-9_]{0,63}$`. |

`reference_pack_windows_path_lookup_v1` applies RP-REQ-118 trimming and NFC, replaces `/` with `\`, collapses consecutive separators, applies `ascii_lower_v1`, and performs no environment expansion, filesystem access, path existence check, or dot-segment resolution. Candidate paths sort by that output, then exact bytes and reject duplicate normalized values. HijackLibs `library_name` uses ASCII case-insensitive basename lookup; `candidate_path` uses the named path algorithm. SID grammar is `S-1-<identifier-authority>-<subauthority>...`; numeric components are unsigned base-10 with no leading zero except `0`, identifier authority is `0..281474976710655`, each subauthority is `0..4294967295`, and exactly 1..15 subauthorities are required. A SID contains at most 184 ASCII bytes. SID lookup is exact; `name` and `alias` use ASCII case-insensitive comparison. Results sort by `entry_id`.

# 17. Source profiles, provenance, and licensing

**RP-REQ-149**
`source_profile_id` is producer provenance. Runtime pack verification MUST NOT execute a source profile and MUST NOT require a source-specific parser after canonical pack bytes exist.

**RP-REQ-150**
Every project-produced or project-distributed pack MUST retain one separate immutable regular-file source-profile artifact whose exact bytes are identified by both `source_profile_id` and `source_profile_sha256`. The digest MUST equal lowercase SHA-256 of the exact artifact bytes. The artifact MUST define:

- accepted upstream artifact shape and media types;
- source version extraction;
- source date extraction;
- source artifact digest calculation;
- field mapping;
- Unicode and identifier normalization;
- duplicate handling;
- null and omission handling;
- relationship construction;
- canonical output ordering;
- licensing inputs;
- valid, malformed, and semantic-error fixtures.

The physical source-profile artifact location and media type are repository-support choices and MUST NOT affect pack runtime behavior. The harness MUST index the exact artifact bytes by `(source_profile_id, source_profile_sha256)` and MUST reject two byte-distinct artifacts that claim the same pair.

**RP-REQ-151**
A source-profile behavior change that can alter canonical pack output MUST use a new `source_profile_id`. Any byte change to the retained source-profile artifact MUST change `source_profile_sha256`; an output-affecting change MUST also change `source_profile_id`. A source-profile artifact MUST NOT be edited in place while retaining both identifiers.

**RP-REQ-152**
An operator-imported pack MAY use a `(source_profile_id, source_profile_sha256)` pair unknown to the running application. Omission behavior: both values are retained as opaque provenance, while runtime content validation depends only on the registered content profile. The runtime MUST NOT fetch, execute, or require the source-profile artifact to admit canonical pack bytes.

**RP-REQ-153**
The implementation MUST validate the SPDX expression and notice references. It MUST NOT decide whether a particular upstream license legally permits transformation or redistribution. Project distribution of a pack remains blocked until the licensing authority records that decision under `RP-GATE-011`.

**RP-REQ-154**
Only packs with `license.redistribution='allowed'` may be embedded in an incident portability bundle. `restricted` and `prohibited` packs are references-only. Operational backup is not redistribution under this technical contract and MUST retain every required local byte.

**RP-REQ-155**
A pack MUST NOT embed raw upstream source artifacts unless each embedded artifact is declared as canonical payload by the content profile and redistribution is `allowed`. No v1 content profile permits raw upstream source artifacts as payload.

# 18. Compatibility, dependencies, and conflicts

**RP-REQ-156**
Version 1 dependencies are exact. Version ranges, wildcards, `latest`, minimum versions, semantic-version comparisons, and dependency auto-resolution are invalid.

**RP-REQ-157**
During verification, every declared dependency tuple MUST resolve to one retained, non-removed version with the exact `payload_sha256`, complete available canonical bytes, and durable condition `verified_available` or `disabled`, or to the exact active version of that tuple. An absent, digest-mismatched, failed, missing, or administratively removed dependency MUST fail verification with `dependency_unsatisfied`. The candidate plus every reachable retained dependency MUST form an acyclic graph; a cycle MUST fail with `dependency_cycle`. Dependency resolution enumerates all independently reachable unavailable exact tuples during `dependencies`; an unavailable tuple is not traversed. Unknown profile keys are unavailable under the captured application contract. Cycles are evaluated only after resolution succeeds, with one diagnostic per tuple belonging to a cyclic strongly connected component. Dependency findings use `$.dependencies` with the bounded related key and version; identical findings coalesce. Operational byte-read failures abort verification. A disabled healthy dependency remains usable. Historical restore resolves retained identities and bytes without consulting current disablement, removal or health, because later administration cannot rewrite successful historical verification. A candidate that conflicts with itself or with one of its declared dependency tuples MUST fail with `pack_conflict`. The presence of an unrelated active conflict is not a verification failure and is evaluated during activation.

Activation requires every declared dependency tuple to be active in the proposed complete pack set. A verified but inactive dependency does not satisfy the dependency. Activation MUST NOT auto-activate a dependency.

**RP-REQ-158**
Before activation, the implementation MUST:

1. construct the proposed complete active set;
2. verify every exact dependency;
3. reject dependency cycles;
4. reject any declared conflict in either direction;
5. evaluate runtime and content-profile compatibility;
6. evaluate type-registry compatibility when applicable;
7. compute the replacement pack set;
8. commit all effects atomically.

Graph traversal order is `pack_key`, then `pack_version`, then `payload_sha256` ascending exact bytes.

Before the final compatibility checks, apply the initiating pointer mutation, restore any absent required registry from the current application release built-in, and remove optional members with unsatisfied exact dependencies. Repeat pruning in graph order until no member changes. Reject cycles or conflicts in the stable proposed set. Required registry packs MUST have empty dependency and conflict arrays; optional packs MUST NOT declare conflicts against required registry keys. No optional member is automatically reactivated.

**RP-REQ-159**
If an active dependency becomes disabled, failed, missing, removed, or replaced by another version, every transitive dependent whose exact tuple is no longer satisfied MUST lose active status in the same publication transaction. No dependent is automatically reactivated later.

**Table 18-A. Compatibility classification**

Registry rows define runtime replacement checks under RP-REQ-160 and RP-REQ-161. Framework/enrichment ID continuity and framework `object_type` stability are producer compatibility obligations against the producer's source history; they are not deployment-history admission checks. A deployment may first encounter any valid version and does not maintain an inferred cross-version subject registry. Runtime validates each version's closed schema, intra-pack identities and references, exact dependencies and immutable logical tuple. Consumers retain version-qualified provenance and treat unknown framework types generically under RP-REQ-133. A producer changing subject or object type must assign a new source-stable ID. Runtime acceptance does not certify that producer obligation.

| Change | Required result |
| --- | --- |
| Add a host or evidence registry entry | Compatible. |
| Change registry `display_label`, `description`, or `icon_key` | Compatible. |
| Change registry `category` for an existing key | Incompatible; use a new key. |
| Mark an entry deprecated while retaining exact resolution | Compatible. |
| Change a non-null `replacement_entry_id` to another target | Incompatible. |
| Remove a key present in the current application release built-in or referenced by retained incident data | Incompatible. |
| Reuse a registry key for a different concept | Incompatible. |
| Add or remove an unambiguous alias | Compatible. |
| Add an alias that creates ambiguous resolution | Incompatible. |
| Add an allowed indicator value kind while all algorithm IDs remain unchanged | Compatible. |
| Remove an allowed indicator value kind | Incompatible. |
| Change an indicator algorithm ID or its behavior digest | Incompatible. |
| Change only `stix_mapping` | Compatible. |
| Add a framework or enrichment entry | Compatible. |
| Remove a framework or enrichment entry without ID reuse | Compatible; pinned old sets preserve old data. |
| Change `object_type` for an existing framework object ID | Incompatible. |
| Reuse a framework or enrichment ID for another subject | Incompatible. |
| Add a member inside `extensions` | Rejected by the current closed schema; requires an adopted schema and profile revision. |
| Change manifest, digest, identity, trust, dependency, ordering, or pack-set algorithm | Requires a new pack contract major. |

**RP-REQ-160**
Type-registry compatibility MUST be evaluated against:

- the currently active version;
- the current application release packaged built-in for the same key;
- active incident values referencing registry keys;
- retained snapshots and releases pinned to the candidate predecessor;
- exact pack dependencies;
- stable algorithm IDs and their adopted implementation digests.

**RP-REQ-161**
An active registry replacement MUST retain every key in the current application release packaged built-in and every key currently referenced by retained incident data. A deprecated retained key remains resolvable. A replacement entry does not rewrite existing incident records.

The only exception is required-registry safety fallback after active content becomes unusable. That fallback MUST occur even when a custom incident token is absent from the built-in. Preserve the exact token and expose an unresolved-reference projection with degraded registry status; do not substitute `unknown` or reuse descriptions from failed content. Unrelated edits remain legal. New type assignments MUST resolve in the effective registry. Restoring a compatible registry resolves the retained token without rewriting the incident record.

Current-profile reachability is explicit: Host and Evidence have no assigned registry-token field in their source-owner record contracts; their effective rough-capture entry is `unknown`. The Indicator registry admits exactly the nine mandatory Core IDs, all present in Base. Consequently a retained custom incident token cannot arise through a current admitted assignment. Host/Evidence custom pack entries remain valid and addressable through the Reference Data consumer interface, but do not create record fields or implicit assignments. Evidence for the custom-token branch must record this constraint, exercise the fixed-point fallback and owner usage guards, and must not fabricate a source-record field solely for a test. Any owner amendment introducing custom assignments must, in the same revision, project unresolved-reference degradation, exact-token preservation, unrelated-edit eligibility and effective-registry assignment validation through that owner's schema, implementation and UI. The general safety-fallback rule remains authoritative for that expanded profile.


**RP-REQ-162**
The application MUST maintain one `algorithm_behavior_digest` for every application-owned indicator algorithm ID. The digest MUST be computed as lowercase SHA-256 of RFC 8785 canonical bytes of the harness-owned ordered conformance-vector array `{algorithm_id, value_kind, raw_value, valid, validation_code, normalized_value, defanged_value, dedupe_key}`. Vectors sort by `algorithm_id`, `value_kind`, then raw-value UTF-8 bytes. Every conforming implementation of one application contract major MUST produce the same digest from the same canonical vector set. A changed behavior digest under the same algorithm ID is `type_registry_incompatible`; source-code or executable-file hashes MUST NOT be used as cross-implementation behavior digests.

# 19. Immutable reference pack set

**RP-REQ-163**
Every ready Base Profile runtime MUST maintain one current immutable `cartulary.reference_pack_set.v1`, even when the Reference Pack Extension Profile is not claimed. The set MUST contain exactly one effective member for each required Base registry key.

**RP-REQ-164**
`cartulary.reference_pack_set.v1` is a closed object:

| Member | Type | Rule |
| --- | --- | --- |
| `schema_id` | string | Exactly `cartulary.reference_pack_set.v1`. |
| `pack_set_id` | string | Derived by RP-REQ-167. |
| `pack_set_sha256` | string | Derived by RP-REQ-166. |
| `members` | array | `3..64` sorted pack-set members; at most one per `pack_key`. |

**RP-REQ-165**
A pack-set member contains exactly:

| Member | Type | Rule |
| --- | --- | --- |
| `pack_key` | string | Exact active key. |
| `pack_version` | string | Exact active version. |
| `manifest_sha256` | string | Exact manifest digest. |
| `payload_sha256` | string | Exact payload digest. |
| `pack_contract_version` | string | Exact pack contract. |
| `content_profile_id` | string | Exact content profile. |
| `content_profile_version` | string | Exact content-profile version. |

Members MUST sort by `pack_key`, then `pack_version`. A duplicate `pack_key` is invalid.

**RP-REQ-166**
`pack_set_sha256` MUST equal:

```text
lowercase_hex(
  SHA256(
    RFC8785_JCS({
      "schema_id": "cartulary.reference_pack_set.v1",
      "members": members
    })
  )
)
```

**RP-REQ-167**
`pack_set_id` MUST equal `"rpset_" + pack_set_sha256`.

**RP-REQ-168**
The set MUST include the effective active versions of:

- `type_registry.host`;
- `type_registry.evidence`;
- `type_registry.indicator`.

A running Base deployment with any required member absent is not ready. Optional keys appear only when active.

If definitive content loss leaves a required key without a healthy selected version or healthy current-release built-in, the invalidation transaction MUST retain the loss and its attestation, clear the current set pointer, and advance its revision. It MUST NOT manufacture an incomplete set, erase immutable history, or roll back the loss because fallback is unavailable. Its attestation has `resulting_pack_set_id=null`. Live readiness MUST remain false until recovery and Base reconciliation establish a complete usable current set. Historical reads may still use exact retained sets whose content remains available. Loss of a required pinned version also blocks readiness; administrative disablement and metadata expiry alone do not invalidate historical pins.

While the current selection is absent, current-set resolution and new consumer captures fail with `pack_unavailable`; new administrative operations fail with `reference_pack_operation_rejected/required_registry_gap` after their prior request and eligibility checks. System integrity operations remain admissible to record further historical losses, freezing an explicit `pack_set_id=null` and the current selection revision. Only those system operations permit this null in the frozen input; omission is invalid. A relevant revision change still rejects pending publication with `stale_admission_state`.

**RP-REQ-169**
Activation, rollback, safety fallback, startup profile reconciliation, disablement of an active version, failed or missing active-version detection, dependency invalidation, and removal-related pointer changes MUST publish exactly one replacement pack set (or clear the current pointer under RP-REQ-168 when no complete usable set exists) in the same atomic transaction as active-pointer and attestation effects. When the canonical member array is unchanged, the operation MUST reuse the existing `pack_set_id` and MUST NOT create a byte-distinct duplicate set object.

**RP-REQ-170**
A reader MUST observe either the complete prior set or the complete replacement set. It MUST NOT observe a mixture of versions from two sets.

**RP-REQ-171**
Every pack-dependent operation MUST capture one `pack_set_id` at admission and MUST use that set for its complete execution. A later active-set change MUST NOT change lookups, validation, report derivation, snapshot derivation, or index selection for the admitted operation.

The first successful publication of a retained logical set MUST fix an immutable provenance anchor for every member. Consumer provenance for that set always uses those anchors. Later verification may update administrative health, successful-envelope selection, and current verification timestamps, but MUST NOT rewrite the set's provenance. Local provenance anchors are not part of the content-derived pack-set digest and are not imported as destination trust.

**RP-REQ-172**
Every retained pack set referenced by a reproducibility pin MUST remain queryable and reconstructable. An application release MUST retain decoders, validators, and consumer behavior required to read every content-profile version referenced by a retained pin; an upgrade that cannot do so MUST fail readiness or require an owner-defined migration before the upgrade is admitted. An unpinned historical pack set MAY be garbage-collected only when none of its members or attestations are needed for rollback, audit, backup retention, or other retained state.

**RP-REQ-173**
Derived caches and indexes MUST bind to at least:

- `pack_set_id`;
- `pack_key`;
- `pack_version`;
- `payload_sha256`;
- `content_profile_id`;
- `content_profile_version`.

A cache entry without that binding MUST NOT satisfy a pack-dependent read.

# 20. Semantic operation contracts

## 20.1 Common operation rules

**RP-REQ-174**
Core 01 owns public route admission. After admission, every route MUST invoke exactly one semantic operation from Table 20-A.

**Table 20-A. Semantic operations**

| Operation | Required effect |
| --- | --- |
| `ImportPack` | Admit one operator container, verify it, retain one candidate, never auto-activate. |
| `ImportPackFromRoot` | Admit one named regular file under the configured incoming root through the same pipeline. |
| `ReverifyPack` | Verify one retained operator-imported version against retained bytes and current trust/runtime state. |
| `RefreshPacks` | Reverify all retained operator-imported versions of the frozen selected keys and rebuild indexes. |
| `ActivatePack` | Activate one exact verified version after all compatibility and closure checks. |
| `DisablePack` | Disable one exact operator-imported version and publish required invalidations or safety fallback. |
| `RemovePack` | Tombstone one eligible operator-imported version and defer physical garbage collection. |
| `ResolveCurrentPackSet` | Return the complete current immutable pack set. |
| `InvalidateUnavailablePack` | Mark a previously usable payload failed or missing and atomically update active-set consequences. |

**RP-REQ-175**
All mutating semantic operations MUST serialize conflicting changes at publication. Lock order is trust repositories by exact ID, affected pack keys by exact key, current-set state, then registry-reference-usage guards. Verification occurs outside long-lived publication locks. Publication MUST compare every relevant captured revision before applying any semantic delta; a mismatch fails with `reference_pack_operation_rejected/stale_admission_state`. No automatic retry, merge, or rebase is permitted. A revision unrelated to the operation's read or write dependencies MUST NOT cause rejection. Verification admission additionally freezes revisions for every supported key reachable through exact usable dependency tuples, including missing supported targets. These read guards are distinct from selected keys and do not assert pending mutation work. Admission checks the closure again after acquiring the ordered key guards; a changed closure or revision rejects with `stale_admission_state`. Verification and publication check those revisions without rebasing. The database retains the recursive closure while the application retains only the current profile catalog's key set.

**RP-REQ-176**
Admission MUST freeze current trusted roots and trust revisions, highest metadata versions, the exact ordered version/envelope cohort, relevant immutable byte references, current pack-set revision, dependency inputs, referenced-type usage revision, and relevant configuration/profile revisions. Later versions never join a refresh cohort. Verification time is captured separately under RP-REQ-075. An operation MUST NOT silently switch to later trust state, a later active set, different bytes, or a changed selector during execution.

Referenced-type usage is a dependency of normal registry activation, which checks existing assignments. Import, reverify, refresh and emergency fallback do not consult assignments or rewrite existing tokens; their key captures use `usage_revision=null` and must not reject solely because record assignments changed. A non-null captured usage revision requires the corresponding guard and exact revision comparison at publication. This distinction does not relax pack-key, trust, configuration or current-set guards.

## 20.2 Import

**RP-REQ-177**
`ImportPack` accepts exactly one admitted container and one semantic context containing:

| Member | Rule |
| --- | --- |
| `container_sha256` | Digest of exact admitted bytes. |
| `activation_policy` | Exactly `staged_only`. Omission at the Core route materializes this value. |
| `actor_kind` | `user` or `local_operator`. |
| `actor_user_id` | Non-null only for `user`. |
| `operator_operation_id` | Non-null only for `local_operator`. |
| `admitted_at` | Fixed operation timestamp. |

Auto-activation is invalid.

**RP-REQ-178**
Import MUST allocate a new `staged` candidate identity only after canonical manifest decoding and lexical validation make `(pack_key, pack_version)` safe to retain. A failure before that boundary creates only the common job result and administrative audit. Replay under the same Core-owned route-scoped idempotency key MUST return the original accepted job or terminal result and MUST NOT start another verification. Every separately admitted import request, including one whose container bytes equal previously admitted bytes, MUST execute full current verification.

When a separately admitted container resolves to an already retained `(pack_key, pack_version, manifest_sha256, payload_sha256)` tuple, the operation is a verification-envelope renewal. Renewal MUST use metadata versions no lower than the retained versions for that logical tuple. On success, it MUST atomically retain the newly verified container and trust metadata as the current reverify envelope while preserving every prior container digest and attestation as history. A failed renewal MUST preserve the prior current envelope, durable condition, active status, `last_verified_at`, and `trust_valid_until`; it records only the failed job, issue summary, and attributable attestation. Renewal MUST NOT replace manifest or payload bytes or create another logical version.

**RP-REQ-179**
Successful first import terminates with condition `verified_available`. Successful exact replay or verification-envelope renewal has these closed effects: `verified_available` remains `verified_available`; `disabled` remains `disabled`; an active valid version remains active with its prior durable condition; and `failed` or `missing` becomes `verified_available` when complete exact bytes were supplied and verification succeeded. Every success updates `last_verified_at` and `trust_valid_until` and records a new verification attestation. Import MUST NOT move an active pointer.

## 20.3 Deployment-local root import

**RP-REQ-180**
Core 01 MUST bind the deployment-local command:

```text
operator reference-pack import <bundle_name>
```

`bundle_name` MUST match `^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`, MUST be exactly one filename segment, and MUST resolve only under:

```text
<roots.reference_pack_storage>/incoming/<bundle_name>
```

**RP-REQ-181**
The local operator command MUST reject `/`, `\`, NUL, empty input, `.`, `..`, symlinks, non-regular files, and root escape. It MUST create no network listener and MUST use the same `ImportPack` pipeline as HTTP upload.

**RP-REQ-182**
The command MUST wait for the admitted background job to reach a terminal state and write exactly one RFC 8785 canonical UTF-8 JSON object followed by LF to stdout. It MUST write no other stdout bytes; successful and ordinary failed execution MUST leave stderr empty. Invalid command syntax MAY write bounded usage text to stderr, but stdout MUST still contain one failed result object. The object conforms to closed `cartulary.reference_pack_operator_result.v1`:

| Member | Type | Rule |
| --- | --- | --- |
| `schema_id` | string | Exactly `cartulary.reference_pack_operator_result.v1`. |
| `operation_id` | string | Stable local operator operation ID. |
| `result` | string | `succeeded` or `failed`. |
| `container_sha256` | string or null | Non-null after the file is read. |
| `pack_key` | string or null | Non-null when identity was decoded safely. |
| `pack_version` | string or null | Non-null when identity was decoded safely. |
| `job_id` | string or null | Non-null after job admission. |
| `error_code` | string or null | Non-null only on failure. |
| `reason_code` | string or null | Non-null only when the error family defines one. |

Exit code is `0` for `succeeded`, `2` for invalid command use, and `3` for an admitted or pre-admission operation failure.

The command loads the deployment configuration through the normal configuration mechanism and requires the Reference Pack profile to be admitted. The deployment's admitted lifecycle worker owns execution; the command MUST NOT create another worker or listener. If that worker is stopped, the Job remains queued until normal worker recovery starts execution. Interruption of command observation MUST NOT cancel or replay an admitted Job. The failed command result retains any known `job_id`; the operator inspects that Job before deliberately admitting another import. Configuration, storage, schema-readiness, observation, and process-lease failures map to `internal_error` without raw error text. A canceled terminal Job maps to `job_canceled`. Typed admission rejections and terminal verification failures retain their declared code and reason. Invalid syntax maps to `invalid_operator_request`.

The command allocates one nonzero operation UUID before input admission. The Reference Data operation, Jobs' durable local submitter identity and the `.v2` private route identity MUST agree on that UUID. No human submitter or human route receipt is invented. Jobs retain the local UUID after ordinary private-request compaction. The public `submitted_by_user_id` is explicitly null for this branch. A new command invocation allocates a new operation; it is an explicit new verification request.

## 20.4 Reverify

**RP-REQ-183**
`ReverifyPack` is valid only for an operator-imported version in `verified_available`, `disabled`, `failed`, or `missing`. It is invalid for `staged` or `packaged_builtin`.

**RP-REQ-184**
Reverify MUST use the exact current successful verification envelope, including its retained operator container, TUF metadata, manifest, payload, and notice bytes. A version without prior successful verification is rejected with `reference_pack_operation_rejected/no_successful_verification` and requires explicit reimport. No retry-envelope resource exists. Reverify MUST NOT synthesize missing metadata from canonical payload bytes, import new bytes, scan the incoming directory, infer a newer version, or perform network access.

**RP-REQ-185**
If any required retained container, metadata, manifest, payload, or notice byte is absent, reverify fails with `payload_missing` and MUST NOT attempt partial reconstruction. If every required byte is present, reverify runs the complete verification pipeline under current trust and runtime state.

**RP-REQ-186**
Successful reverify restores verification health but preserves administrative disablement. The public condition is `disabled` when that flag is set, otherwise `verified_available`. If the version was active and remains valid, it remains active. Failed reverify of an active version invokes `InvalidateUnavailablePack` and applies optional invalidation or required-registry safety fallback. Only explicit successful activation clears administrative disablement.

## 20.5 Refresh

**RP-REQ-187**
At Core route admission, omitted `pack_keys[]` MUST resolve once to every `pack_key` that has at least one retained operator-imported version with a prior successful verification envelope and is not tombstoned by administrative removal. Packaged-built-in-only keys, never-successful candidates, and administrative-removal tombstones do not enter the eligible cohort. Explicit selected keys use Core's canonical exact-token set semantics. Conflicting nonterminal work under a selected key rejects the whole admission with `reference_pack_operation_rejected/verification_pending`. A known explicitly selected key with no eligible version contributes no cohort member.

**RP-REQ-188**
`RefreshPacks` MUST process exactly the eligible version/envelope cohort frozen at admission, sorted by `pack_key`, then exact `pack_version`. Never-successful candidates and administrative-removal tombstones are skipped and unchanged. It MUST NOT:

- scan `incoming/`;
- discover a new container;
- import new bytes;
- download content;
- infer a latest version;
- auto-activate a different version.

**RP-REQ-189**
Refresh verification and index construction MUST occur in temporary state. All selected version outcomes, active-pointer removals, required-registry fallbacks, dependency invalidations, attestations, compatible trust-root updates, trusted metadata versions, and the replacement pack set MUST publish in one transaction. Root-update chains from different selected containers MUST start from the admission root, MUST have byte-identical metadata for every shared root version, and MUST form one prefix-compatible chain; otherwise refresh fails with `tuf_root_rotation_invalid` and publishes no refresh outcomes.

**RP-REQ-190**
For each selected version, successful refresh restores verification health, preserves administrative disablement and an active valid version, and clears storage-related `missing_reason` only when every retained byte is present and verification succeeds. Its public condition is `disabled` when the administrative flag is set, otherwise `verified_available`. A failed refresh changes the addressed version to `failed` or `missing` under RP-REQ-202 without clearing disablement and applies active invalidation when required. If any selected version fails, the job terminal result is failed. The first failure under pack ordering and validation issue ordering controls the public summary. All selected version outcomes MUST nevertheless be published together at the one publication boundary. Cancellation or timeout with proven absent publication publishes none of the refresh outcomes.

**RP-REQ-191**
A refresh selection containing zero retained operator-imported versions is a successful deterministic no-op job. It produces no version mutation and no new pack set.

## 20.6 Activation and rollback

**RP-REQ-192**
`ActivatePack` is legal only when the addressed version has healthy successful verification, its public condition is `verified_available` or `disabled`, all trust metadata required for fresh activation is unexpired, dependencies are active, conflicts are absent, and compatibility passes. Success clears administrative disablement in the publication transaction.

**RP-REQ-193**
Version 1 supports one target version per activation request. Multi-pack activation requests are unsupported. Dependencies must already be active.

**RP-REQ-194**
Activation MUST atomically write:

- the target active pointer;
- the prior active version reference;
- dependency invalidations caused by the replacement;
- activation attestation and mode;
- cache/index invalidation publication;
- the replacement reference pack set;
- administrative audit outbox state.

**RP-REQ-195**
Activation of an older retained version uses the same operation and public route as normal activation. `activation_mode` is derived under RP-REQ-087. For an operator-imported target, trust freshness applies; packaged built-ins have no TUF freshness boundary but MUST remain release-binding-valid and runtime-compatible.

## 20.7 Disablement

**RP-REQ-196**
`DisablePack` is valid only for an operator-imported version in `verified_available`, whether active or inactive. A packaged built-in fails with `reference_pack_operation_rejected/packaged_builtin`. An already disabled or nonhealthy operator version fails with `reference_pack_operation_rejected/not_disableable`.

**RP-REQ-197**
Disabling an inactive version changes only that version's condition and attestation state. Disabling an active optional version removes its active pointer and invalidates dependents. Disabling an active imported required registry performs safety fallback under RP-REQ-085 and then marks the imported version disabled.

## 20.8 Removal

**RP-REQ-198**
Core 01 MUST add:

```text
POST /api/v1/reference-packs/{pack_key}/{pack_version}/remove
```

The request object MUST contain required `client_txn_id` and required `reason`. `reason` MUST normalize under Core's `reason_note_v1` and MUST remain non-empty after normalization.

**RP-REQ-199**
Removal MUST reject:

- an active version;
- a packaged built-in version;
- a version pinned by a retained snapshot, release, report, or other reproducibility object;
- a version with pending import, reverify, refresh, activation, disablement, or removal work;
- an already administratively removed version.

**RP-REQ-200**
Successful removal MUST atomically:

1. preserve identity, contract, content profile, provenance, manifest digest, payload digest, and attestations;
2. set condition `missing`;
3. set `missing_reason='administrative_removal'`;
4. make derived indexes unavailable;
5. write a removal attestation and administrative audit event;
6. schedule physical byte garbage collection after commit;
7. change no unrelated version and activate no fallback.

Before deleting bytes, the garbage collector MUST re-read the logical version and prove that it remains `missing` with `missing_reason='administrative_removal'`, remains unpinned, has no active pointer, has not been restored by exact reimport, and shares no candidate, verification-envelope, backup-retention, or other retained byte reference. A failed proof makes collection a no-op.

The removal tombstone is durable collection work. The collector runs before readiness after terminal-Job reconciliation, then retries once per minute with a 30-second per-sweep budget. A failed sweep preserves bytes and cannot change an acknowledged administrative outcome. Every successful envelope retains its exact signed container for historical authentication, including after removal; removal releases extracted version copies, not this independent envelope reference. Terminal operations release preparation byte ownership while retaining frozen descriptors, attribution, attempts and their outcomes. Nonterminal operations retain every input and prepared object needed to resume. Release, backup, version, envelope and pin retention all independently prevent physical deletion.

A publisher holds a shared confined-storage namespace lease from immutable-byte creation until database registration resolves, and registers a shared byte-retention guard in the mutation transaction. Collection acquires the exclusive namespace lease before the exclusive database retention guard. This order also covers a lost registration commit acknowledgement: physical deletion waits until the registering database transaction actually finishes. Unregistered physical objects left by process death can then be removed without an age or PID heuristic. Collection marks unreferenced objects unavailable before unlinking; unknown commit outcomes retain bytes until a later authoritative read. Each new publication has a distinct physical identity, and collection rechecks current references immediately before unlinking. Removed extracted copies are neither required nor recreated by restore; their retained signed containers still undergo full verification at the recorded successful instant.

Backup capture acquires its shared byte-retention guard **before** opening the repeatable-read database snapshot, and holds it through completion of object streaming. Acquiring this guard as a statement inside that snapshot is insufficient: the snapshot might have been established before a waiting collector finished. A backup may therefore retain pre-removal copies even when removal completes while capture is streaming.


**RP-REQ-201**
An administratively removed version may return only through exact reimport of the same logical bytes. Reverify alone cannot recreate absent bytes. Removal history remains append-only after reimport.

## 20.9 Missing-payload invalidation

**RP-REQ-202**
When a consumer or integrity scanner detects unavailable retained state, `InvalidateUnavailablePack` MUST serialize with ordinary pack mutations and use this closed classification: definitively absent authoritative bytes or replacement by a prohibited filesystem object become `missing` with `missing_reason='storage_loss'`; the same loss of staged bytes becomes `missing` with `missing_reason='staging_loss'`; present bytes whose digest, schema, semantics, or runtime compatibility fails become `failed` with `missing_reason=null`. A symlink, non-regular object, or multiply linked file is a prohibited replacement. Permission errors, device or transport read failures, loss of the root capability, and cancellation are operational aborts. They fail the attempted read without declaring content invalid, changing established health, publishing fallback, or emitting an invalidity attestation. Archive parsers MUST NOT translate an underlying operational read failure into a malformed-content verdict. The invalidation operation MUST preserve the first detection attestation and atomically apply active-pointer, dependency, safety-fallback, and pack-set effects before another dependent operation succeeds.

A consumer operating inside another owner's transaction MUST abort that entire transaction before attempting the independent integrity mutation. The participating owner MUST treat `pack_unavailable` as failure of that source operation; it cannot commit source changes after authoritative content loss. Integrity detection may supersede pending verification work; advancing the affected revisions makes that work fail its publication guard. The invalidator rechecks bytes outside publication locks and compares the observed key revision at publication. A stale observation cannot condemn concurrently restored content. Repeated reads of an already invalidated version do not replace its first detection attestation. Loss of a derived index alone is an operational availability failure and MUST NOT fabricate a content rejection.

Reverify and refresh that first invalidate established content MUST retain both the ordinary verification attestation and a distinct `payload_invalidation` attestation. Both name the complete resulting set and commit with the same attempt, health transition and fallback. An attempt member records whether it performed this first invalidation; preparation and operational aborts MUST leave that fact false. Repeating a content rejection while the version remains invalid creates verification evidence only. A later successful verification begins a new health cycle and may be followed by a new first invalidation. Failed import renewal never asserts this fact because it preserves the established version's health.

## 20.10 Job progress

**RP-REQ-203**
Reference Pack verification jobs MUST use these semantic progress phases in order:

```text
admitted
archive_preflight
extracting
trust_validation
manifest_validation
payload_verification
content_validation
compatibility_validation
index_build
publishing
```

A job MAY omit a phase that has zero work, but it MUST NOT report a later phase before an earlier performed phase. These are progress tokens, not durable pack conditions.

# 21. Consumer interface

**RP-REQ-204**
Pack consumers MUST use the operations in Table 21-A. They MUST NOT read extracted pack files, reference-pack persistence tables, object-store objects, or mutable active pointers directly.

**Table 21-A. Consumer operations**

| Operation | Request | Result |
| --- | --- | --- |
| `ResolveCurrentPackSet` | none | Complete current `cartulary.reference_pack_set.v1`. |
| `GetPackEntry` | `pack_set_id`, `pack_key`, `entry_id` | One entry and provenance or `entry_not_found`. |
| `LookupPackEntries` | `pack_set_id`, `pack_key`, `lookup_kind`, `lookup_value`, optional `limit`, optional `cursor` | Ordered page and provenance. |
| `EvaluateIndicatorValue` | `pack_set_id`, `indicator_type_id`, `value_kind`, `raw_value` | Canonical indicator evaluation. |
| `GetPackProvenance` | `pack_set_id`, `pack_key` | Manifest, source, license, and verification summary. |

**RP-REQ-205**
Every operation MUST validate its closed request shape and scalar bounds, then resolve the exact retained pack set, then resolve the selected key/profile/type, then validate operation-specific semantics and cursors. Within request validation, fields follow the request table order. `pack_set_id` MUST match `^rpset_[0-9a-f]{64}$`; malformed input fails with `invalid_pack_request`. Unknown or garbage-collected `pack_set_id` fails with `pack_set_not_found`. A pinned historical set may be read from its exact previously verified bytes even when the current version condition is `failed` solely because a later freshness reverify reported `metadata_expired`; missing bytes, digest failure, or content incompatibility still fails with `pack_unavailable`. A set member whose required bytes or index cannot be read triggers RP-REQ-202 when current state has not already recorded the failure.

Consumer request objects are closed: unknown members, omitted required members, explicit nulls and wrong types are invalid. `pack_key` uses the canonical pack-key grammar and at most 128 ASCII bytes. `lookup_kind`, `indicator_type_id` and `value_kind` are non-empty strings of at most 64 UTF-8 bytes; supported-token selection occurs only after set resolution. A supplied cursor is a non-empty string of at most 65536 UTF-8 bytes. The structured lookup union admits exactly `{provider_name,event_id}` with optional non-null `event_version`, or exactly `{algorithm,value}`. Provider names have at most 256 Unicode scalars, event IDs are integers in `0..65535`, versions in `0..255`, algorithm names are non-empty strings of at most 64 scalars, and hash lookup values are non-empty strings of at most 8192 scalars. These are request-shape bounds; profile-specific syntax and supported algorithms remain later checks. `ResolveCurrentPackSet` has no request fields; its empty object projection does not introduce a public route.

Every serialized consumer result contains exactly required `value` and `error`. Exactly one is non-null. An error contains exactly required `code` and nullable `reason_code`; the latter is `indicator_input_too_long` only with `invalid_pack_request`, and otherwise null. The closed error-code registry is `invalid_pack_request`, `pack_set_not_found`, `pack_unavailable`, `entry_not_found`, `lookup_kind_unsupported`, `cursor_invalid`, `cursor_query_mismatch`, `indicator_type_unsupported`, `indicator_value_kind_unsupported` and `indicator_algorithm_unsupported`.

Every consumer returns a discriminated success-or-error result. An error contains its stable code and bounded safe details; it does not contain a fabricated success payload. An overlong indicator raw value fails request admission with `invalid_pack_request/indicator_input_too_long` without echoing the raw value. Unknown type, unsupported algorithm, or disallowed value kind is an operation error under RP-REQ-212. Only an admitted input for a resolved supported type may return the RP-REQ-211 evaluation object with `valid=false`.

**RP-REQ-206**
`GetPackEntry` uses exact `entry_id` or `object_id` comparison. The requested identity MUST be a non-empty string of at most 512 UTF-8 bytes, with the addressed profile's narrower identity syntax and bound enforced after profile resolution. A pack key not present in the set returns `pack_unavailable`. A present pack with no matching entry returns `entry_not_found`. Success returns one closed object containing exactly `item` and `provenance`; `item` is the canonical decoded entry or object and `provenance` satisfies RP-REQ-210.

**RP-REQ-207**
`LookupPackEntries` requires the profile-declared `lookup_value` type: a non-empty string of at most 8192 Unicode scalar values, an exact JSON integer in the profile domain, or the closed structured lookup object named by the profile. On a first-page request, omitted `limit` resolves to `50`; the minimum is `1` and maximum is `200`. On a continuation request, omitted `limit` resolves to the cursor-bound limit, while an explicit limit MUST equal that bound value or fail with `cursor_query_mismatch`. Omitted cursor means the first page; explicit JSON `null` for `limit` or `cursor` is invalid. Results sort by the profile-specific order, or otherwise by canonical lookup key then canonical entry or object ID. Success returns a closed object containing exactly `items`, `next_cursor`, and `provenance`; `items` is an array of canonical decoded entries or objects, and `next_cursor` is `null` on the final page.

**RP-REQ-208**
A lookup cursor MUST be opaque, confidential, and integrity-protected. It MUST bind to:

- `pack_set_id`;
- `pack_key`;
- `lookup_kind`;
- normalized lookup value;
- effective limit;
- continuation identity.

The cursor MUST bind an issuance timestamp and an expiry timestamp exactly 900 seconds later. At continuation admission, `current_time >= expires_at` is expired. Malformed, expired, undecryptable, or tampered cursor input fails with `cursor_invalid`. Reuse with a different bound value fails with `cursor_query_mismatch`. Neither error may disclose cursor plaintext or continuation state. Cursor state is deployment-local and is not portability content.

**RP-REQ-209**
Type-registry lookup kinds are exactly `entry_id` and `alias`. Framework lookup kinds are defined by RP-REQ-136. Enrichment lookup kinds are defined by Table 16-A. An unsupported lookup kind fails with `lookup_kind_unsupported`.

**RP-REQ-210**
Every successful entry, lookup, and provenance result MUST include one closed provenance object containing:

```text
pack_set_id
pack_key
pack_version
manifest_sha256
payload_sha256
pack_contract_version
content_profile_id
content_profile_version
authority_class
source_profile_id
source_profile_sha256
source_identifier
source_version
source_as_of
source_artifacts
license
verification_method
last_verified_at
trust_valid_until
verified_signer_key_ids
```

`trust_valid_until` is `null` and `verified_signer_key_ids=[]` for packaged built-ins. The provenance object MUST NOT include raw signatures, private keys, full TUF metadata, or storage paths. `GetPackProvenance` returns exactly this object. Entry and object results additionally expose their exact entry or object ID through the canonical `item`.

All members of this provenance object, including successful verification time and signer sets, come from the immutable anchor fixed by RP-REQ-171. Current administrative health and subsequent verification times are available only through the Core administrative version resource; they MUST NOT change a pinned consumer response.

**RP-REQ-211**
`EvaluateIndicatorValue` MUST return a closed object containing:

| Member | Type | Rule |
| --- | --- | --- |
| `indicator_type_id` | string | Exact requested registry key. |
| `value_kind` | string | Exact requested Core token. |
| `raw_value` | string | Exact caller input, maximum 8192 Unicode scalar values; MUST NOT be logged or emitted as telemetry. |
| `valid` | boolean | Overall result. |
| `validation_code` | string or null | Stable failure code or `null`. |
| `display_value` | string or null | Canonical display value. |
| `normalized_value` | string or null | Canonical normalized value. |
| `defanged_value` | string or null | Algorithm result. |
| `dedupe_key` | string or null | Algorithm result. |
| `normalization_algorithm_id` | string | Applied registry value. |
| `validation_algorithm_id` | string | Applied registry value. |
| `defang_algorithm_id` | string | Applied registry value. |
| `dedupe_algorithm_id` | string | Applied registry value. |
| `provenance` | object | Exact pack provenance under RP-REQ-210. |

The closed invalid-value code registry is `invalid_ipv4`, `invalid_ipv6`, `invalid_domain_name`, `invalid_http_url`, `invalid_sha256`, `invalid_email_addr`, `invalid_registry_key`, `invalid_process_name` and `invalid_text`; each family returns its corresponding code. Unsupported algorithms are operation errors and never invalid-value evaluations.

When `valid=true`, `validation_code` MUST be `null` and all four derived value members MUST be non-null. When `valid=false`, `validation_code` MUST be non-null and all four derived value members MUST be `null`.

**RP-REQ-212**
`EvaluateIndicatorValue` MUST reject an unknown type with `indicator_type_unsupported`, a disallowed `value_kind` with `indicator_value_kind_unsupported`, and a registry algorithm ID unsupported by the running application with `indicator_algorithm_unsupported`. It MUST NOT infer an algorithm from display label, icon, or STIX mapping.

**RP-REQ-213**
Framework and enrichment consumer results MUST NOT be represented as authoritative evidence. The consuming UI or report model MUST retain their `authority_class` and provenance.

# 22. Core public route and error binding

**RP-REQ-214**
Core 01 public routes map to semantic operations as follows:

| Core route | Semantic operation |
| --- | --- |
| `POST /api/v1/reference-packs/import` | `ImportPack` |
| `POST .../activate` | `ActivatePack` |
| `POST .../disable` | `DisablePack` |
| `POST .../reverify` | `ReverifyPack` |
| `POST /api/v1/reference-packs/refresh` | `RefreshPacks` |
| Proposed `POST .../remove` | `RemovePack` |
| `GET` list and singleton | Metadata projection over retained version state. |

This table does not redefine Core envelopes, authorization, paging, idempotency, or status codes.

**RP-REQ-215**
Core 01 MUST extend `reference_pack_version` with these members:

| Member | Nullability | Rule |
| --- | --- | --- |
| `distribution_kind` | non-null | `packaged_builtin` or `operator_imported`. |
| `pack_release_sequence` | non-null | RP-REQ-024. |
| `content_profile_id` | non-null | Exact registered ID. |
| `content_profile_version` | non-null | Exact registered version. |
| `source_profile_id` | non-null | Exact opaque producer-profile ID. |
| `source_profile_sha256` | non-null | Exact source-profile artifact digest. |
| `source_version` | non-null | Exact manifest value. |
| `source_as_of` | nullable | Exact manifest value. |
| `license_expression` | non-null | Exact SPDX expression. |
| `redistribution` | non-null | `allowed`, `restricted`, or `prohibited`. |
| `trust_repository_id` | nullable | `null` for packaged built-ins. |
| `last_verified_at` | nullable | Successful verification-start instant; null before first success. |
| `trust_valid_until` | nullable | `null` for packaged built-ins. |
| `missing_reason` | nullable | `administrative_removal`, `storage_loss`, `staging_loss`, or `null`. |
| `verified_signer_key_ids` | non-null array | Empty for packaged built-ins; sorted valid signer IDs for imported packs. |

Before a first successful verification, every success-only field above other than `distribution_kind` and `missing_reason` is nullable and MUST be `null`; `verified_signer_key_ids` is instead `[]`. After success, the table's ordinary nullability applies, including when subsequent attempts fail. `trust_valid_until` remains null for built-ins. A `missing_reason` is non-null exactly when the public condition is `missing`; `administrative_removal` can be cleared only by exact reimport. Storage or staging loss clears after complete successful verification; failure with present but invalid bytes sets condition `failed` and clears the missing reason. Administrative disablement is retained independently of verification health.

Core 01 MUST add `reference_pack_operation_rejected` with HTTP status `409` and the closed reason registry `clock_untrusted`, `no_successful_verification`, `verification_pending`, and `stale_admission_state`. These are admission or execution outcomes, not proof that established content is invalid. Common authentication, request-shape, and idempotent-replay admission precede these checks. Within this family, pending-work rejection precedes successful-envelope eligibility; clock assertion applies when fresh trust evaluation begins; stale-state rejection applies only at publication.

**RP-REQ-216**
Core 01 MUST add these `reference_pack_verification_failed` reasons:

```text
unsupported_container_format
container_bytes_exceeded
archive_structure_invalid
path_traversal
path_collision
disallowed_member_type
undeclared_member
required_member_missing
bundle_hint_invalid
bundle_hint_noncanonical
manifest_encoding_invalid
manifest_json_invalid
duplicate_object_member
manifest_schema_invalid
manifest_noncanonical
tuf_metadata_invalid
metadata_noncanonical
tuf_root_untrusted
tuf_root_rotation_invalid
signature_threshold_not_met
missing_integrity_metadata
metadata_expired
metadata_expiry_policy_invalid
metadata_rollback_detected
metadata_mix_and_match_detected
target_not_declared
unexpected_target
target_length_mismatch
checksum_mismatch
pack_version_collision
pack_release_sequence_rollback
pack_release_sequence_collision
contract_incompatible
content_schema_invalid
content_semantic_invalid
dependency_unsatisfied
dependency_cycle
pack_conflict
type_registry_incompatible
disallowed_content
payload_missing
verification_timeout
archive_extracted_bytes_exceeded
archive_compression_ratio_exceeded
archive_member_count_exceeded
```

The resulting registry is closed. No other `reference_pack_verification_failed` reason is valid in contract major 2. Table 22-A defines reason meanings, not precedence. The check registry in Table 12-B exclusively owns content-check precedence. Timeout is an execution-abort mapping and never a content-invalidity verdict.

**Table 22-A. Reference Pack verification reason mapping**

| Reason | Phase | Exact trigger |
| --- | --- | --- |
| `unsupported_container_format` | `container_admission` | The admitted bytes identify neither a permitted ZIP nor a permitted TAR, or identify GZIP whose single decompressed member is not TAR. |
| `container_bytes_exceeded` | `container_admission` | Exact admitted container bytes exceed the effective `max_container_bytes`. |
| `archive_structure_invalid` | `archive_preflight` or `archive_structure` | A recognized archive violates RP-REQ-040, a path violates RP-REQ-036 or RP-REQ-037 without matching `path_traversal` or `path_collision`, a fixed archive-structure limit maps here under Table 26-C, or the logical layout violates RP-REQ-041 through RP-REQ-047 without a more specific row below. |
| `path_traversal` | `archive_structure` | A member path is absolute, begins with `./`, contains `.` or `..` segments, or extraction would escape the private temporary root. |
| `path_collision` | `archive_structure` | Two members have the same normalized path, differ only by ASCII case, or create a regular-file parent and descendant collision. |
| `disallowed_member_type` | `archive_structure` | A member is a symlink, hard link, device, FIFO, socket, sparse member, or another non-directory, non-regular-file type. |
| `undeclared_member` | `target_inventory` | A regular file exists outside the exact distribution-kind inventory in RP-REQ-047 or is not declared by `manifest.files[]` where declaration is required. |
| `required_member_missing` | `target_inventory` | A required structural, metadata, payload, notice, or manifest-declared regular file is absent. |
| `bundle_hint_invalid` | `trust_selection` | `bundle.json` fails RP-REQ-048 for any reason other than canonical byte inequality, including its fixed byte limit. |
| `bundle_hint_noncanonical` | `trust_selection` | A structurally valid `bundle.json` is not byte-for-byte equal to its RFC 8785 representation. |
| `manifest_encoding_invalid` | `manifest_schema` | `manifest.json` has invalid UTF-8, a BOM, invalid Unicode, or an unpaired surrogate. |
| `manifest_json_invalid` | `manifest_schema` | `manifest.json` has invalid JSON syntax or its decoded top level is not an object. |
| `duplicate_object_member` | `manifest_schema` | `manifest.json` contains a duplicate object member at any depth. Duplicate members in another file family use that family's schema reason. |
| `manifest_schema_invalid` | `manifest_schema` | `manifest.json` has a missing or unknown member, type or nullability mismatch, invalid scalar, invalid ordering or duplicate collection item, count mismatch, or manifest-specific limit breach, and no earlier manifest row applies. |
| `manifest_noncanonical` | `manifest_schema` | A structurally valid `manifest.json` is not byte-for-byte equal to its RFC 8785 representation. |
| `tuf_metadata_invalid` | `tuf_metadata` | TUF metadata has invalid encoding, JSON, outer shape, role schema, key or signature lexical shape, threshold domain, role membership, fixed metadata limit, or another POUF violation, except a canonical-byte mismatch, root-trust failure, root-rotation failure, or missing integrity binding. |
| `metadata_noncanonical` | `tuf_metadata` | Structurally valid TUF metadata is not byte-for-byte equal to its RFC 8785 representation. |
| `tuf_root_untrusted` | `trust_selection` or `tuf_metadata` | The repository is unknown, the bootstrap-root digest or self-consistency check fails, or the configured bootstrap root cannot authorize the selected repository. |
| `tuf_root_rotation_invalid` | `tuf_metadata` | A sequential root update violates version continuity, predecessor threshold, successor threshold, repository identity, key reference, or the refresh prefix-compatibility rule. |
| `signature_threshold_not_met` | `tuf_metadata` | An applicable role lacks the required count of distinct valid authorized signatures, or contains an untrusted, duplicate, or invalid signature under RP-REQ-073. |
| `missing_integrity_metadata` | `tuf_metadata` or `target_inventory` | A required packaged-built-in release binding, TUF metadata descriptor version, target length, or SHA-256 integrity member is absent from an otherwise attributable object. |
| `metadata_expired` | `tuf_metadata` | A required final role expiry is not strictly later than the fixed operation start time. |
| `metadata_expiry_policy_invalid` | `tuf_metadata` | A final-role expiry exceeds its maximum remaining-validity horizon or violates required expiry ordering. |
| `metadata_rollback_detected` | `tuf_metadata` | A metadata version is below its persisted highest trusted version in the RP-REQ-068 scope, or same-version canonical bytes differ. |
| `metadata_mix_and_match_detected` | `tuf_metadata` or `target_inventory` | Timestamp, snapshot, targets, bundle hint, manifest identity, or signed Cartulary target binding refers to mutually inconsistent versions, bytes, identities, or digests. |
| `target_not_declared` | `target_inventory` | A required `bundle.json`, `manifest.json`, payload, or notice file is absent from targets metadata. |
| `unexpected_target` | `target_inventory` | Targets metadata declares a target outside the exact required inventory. |
| `target_length_mismatch` | `target_inventory` or `payload_integrity` | A TUF target descriptor or manifest member length differs from the exact target or retained member byte length. |
| `checksum_mismatch` | `target_inventory` or `payload_integrity` | A TUF target SHA-256, manifest-declared member SHA-256, packaged-built-in release digest, or retained-byte digest differs from recomputed bytes. |
| `pack_version_collision` | `logical_identity` | A retained `(pack_key, pack_version)` has different logical manifest or payload digests. |
| `pack_release_sequence_rollback` | `logical_identity` | A previously unseen release sequence is below the highest accepted sequence in its comparison scope. |
| `pack_release_sequence_collision` | `logical_identity` | An accepted release sequence is reused with different logical digests. |
| `contract_incompatible` | `runtime_compatibility` | Pack contract, content-profile registration or version, runtime capability, application contract major, or another declared compatibility constraint is unsupported. |
| `content_schema_invalid` | `content_schema` | NDJSON framing, canonical line shape, field schema, identity grammar, sort order, duplicate rule, declared count, cross-file structural rule, or content-profile limit fails. |
| `content_semantic_invalid` | `content_semantics` | A content-profile semantic invariant, cross-reference, alias, relationship, normalization-independent value rule, or license/notice semantic rule fails and no specialized reason applies. |
| `dependency_unsatisfied` | `dependency_validation` | A declared exact dependency does not resolve under RP-REQ-157. |
| `dependency_cycle` | `dependency_validation` | The candidate and reachable retained dependency graph contains a cycle. |
| `pack_conflict` | `dependency_validation` | The candidate conflicts with itself or one of its declared dependency tuples. |
| `type_registry_incompatible` | `type_registry_compatibility` | A registry replacement violates Table 18-A, retained-key requirements, alias uniqueness, or indicator algorithm behavior compatibility. |
| `disallowed_content` | `content_schema` or `content_semantics` | Logical members, media roles, or profile values contain or declare prohibited executable or active content under RP-REQ-237 rather than permitted inert text. |
| `payload_missing` | `payload_integrity` | Reverify, refresh, restore-time pack verification, or integrity validation requires retained authoritative bytes that are definitively absent or replaced by prohibited filesystem objects; permission, transport and device failures remain operational; initial import uses `required_member_missing`. |
| `verification_timeout` | current performed phase | The monotonic attempt budget reaches or exceeds `limits.reference_packs.max_verification_seconds` before publication; equality is timeout. |
| `archive_extracted_bytes_exceeded` | `archive_preflight` or `archive_structure` | Extracted bytes exceed the effective `max_extracted_bytes`. |
| `archive_compression_ratio_exceeded` | `archive_preflight` | Compression ratio exceeds the effective `max_compression_ratio`. |
| `archive_member_count_exceeded` | `archive_preflight` | Regular-file member count exceeds the effective `max_members`. |

**RP-REQ-217**
Core 01's `reference_pack_activation_rejected` family has exactly `already_active`, `not_verified_available`, and `metadata_expired`. Shared operation admission and compatibility failures use `reference_pack_operation_rejected`, with the closed reason registry in Core 01 REQ-01-482. Retired reason aliases are invalid.

Exact idempotent replay precedes new eligibility checks. Authentication, closed request admission, target resolution, frozen-input capture and publication guards apply before the operation-specific table. Pending work is evaluated at the selected pack-key boundary, including work on another version, before missing-envelope eligibility. Relevant revision changes are operational `stale_admission_state` rejections and do not condemn content.

**Table 22-B. Activation rejection mapping after common admission**

| Order | Public family | Reason | Exact trigger |
| ---: | --- | --- | --- |
| 1 | `reference_pack_operation_rejected` | `removed` | The exact version is an administrative removal tombstone. |
| 2 | `reference_pack_activation_rejected` | `already_active` | The exact version is already effective for its key. |
| 3 | `reference_pack_activation_rejected` | `not_verified_available` | Healthy successful verification or present immutable bytes are absent. Disablement alone does not fail this check. |
| 4 | `reference_pack_operation_rejected` | `clock_untrusted` | Operator-imported content requires a fresh activation check and the deployment clock assertion is false. |
| 5 | `reference_pack_activation_rejected` | `metadata_expired` | Operator metadata expires at or before the fixed activation start instant. |
| 6 | `reference_pack_operation_rejected` | `type_registry_incompatible` | Normal registry replacement violates §18's referenced-record or retained-key requirements. |
| 7 | `reference_pack_operation_rejected` | `dependency_unsatisfied` | An exact declared dependency is absent from the proposed active selection. |
| 8 | `reference_pack_operation_rejected` | `required_registry_gap`, `contract_incompatible`, `pack_conflict`, or `dependency_cycle` | The deterministic effective-set algorithm in §19 rejects the proposal. Its check order controls this stage; table order is not a second verifier precedence rule. |

**RP-REQ-218**
Removal uses `reference_pack_operation_rejected`; it has no separate removal-error family. After common admission and exact replay, Table 22-C is exhaustive for removal eligibility. Admission rejects `verification_pending` for any conflicting nonterminal operation on the selected pack key before these checks. Integrity loss and execution failure use their existing operational outcomes rather than a removal content verdict.

**Table 22-C. Removal rejection mapping after common admission**

| Order | Reason | Exact trigger |
| ---: | --- | --- |
| 1 | `removed` | The exact version is already an administrative removal tombstone. |
| 2 | `packaged_builtin` | The version is packaged Base content. |
| 3 | `active` | The exact version is effective for its key. |
| 4 | `pinned` | A retained exact-version pin or a pin on a containing set requires the version. |

The administrative resource's `pending_work` and `reproducibility_pinned` booleans expose these current blockers without disclosing pin-owner or operation details. A client MUST still handle a transactional rejection after observing an eligible resource.

**RP-REQ-219**
Internal validation codes MUST map deterministically to one public reason. More specific internal detail may remain in the retained issue summary, but the public error MUST NOT expose raw payload content, paths, signatures, or keys.

**RP-REQ-220**
Until `RP-GATE-003` is adopted, implementations may expose experimental removal or expanded verification diagnostics only outside a conformance claim. Conformance validation MUST report the missing Core adoption dependency; experimental routes still use the explicitly projected public contract and MUST NOT invent a runtime error alias for an adoption gate.

# 23. Snapshots, reporting, portability, backup, and restore

## 23.1 Snapshot and reporting binding

**RP-REQ-221**
Every snapshot created under the Snapshot and Reporting Extension Profile MUST retain:

```text
reference_pack_set_id
reference_pack_set_sha256
```

The full immutable member list MUST remain durably resolvable from that ID.

**RP-REQ-222**
Every report release tuple and export-model materialization that can depend on pack data MUST bind to the snapshot's exact pack set. A renderer MUST NOT consult the current active set after admission.

**RP-REQ-223**
A rerender or re-derivation whose pinned pack set cannot be reconstructed MUST fail with `required_reference_pack_unavailable`. It MUST NOT substitute another active, newer, older, or built-in version silently.

**RP-REQ-224**
Every pack-derived report or snapshot value MUST retain entry-level provenance sufficient to recover the exact pack key, version, payload digest, and entry or object ID.

## 23.2 Incident portability

**RP-REQ-225**
`reference_pack_refs.json` MUST be a closed object with required `schema_id='reference_pack_refs.v1'`, `sets`, and `versions`. Both collections are required non-null arrays; no references is exactly `sets=[]` and `versions=[]`. This representation is carried only by incident-bundle format 5. Each `sets` element is the complete RP-REQ-164 pack-set object, with digest and ID recomputed under RP-REQ-166 and RP-REQ-167. Sets sort by `pack_set_id` and reject duplicate IDs. Each version tuple occurs exactly once in `versions`, sorted by `pack_key`, `pack_version`, `manifest_sha256`, then `payload_sha256`. Every set member MUST resolve to exactly one matching version item, and every version item MUST be referenced by a set. A same-key/version digest collision is invalid. Each version item contains exactly:

```text
pack_key
pack_version
manifest_sha256
payload_sha256
pack_contract_version
content_profile_id
content_profile_version
distribution_kind
verification_method
source_profile_id
source_profile_sha256
```

**RP-REQ-226**
An embedded pack in an incident bundle MUST receive no trust from the incident-bundle signature itself, import no source-deployment activation or trust state, reuse an existing exact logical version when digests match, and reject a same-key/version digest collision. Newly imported versions remain inactive. Reusing an existing version MUST neither activate nor deactivate it; an exact active local version remains active. Source attestations may be retained as historical artifact evidence only and never confer destination trust. An embedded `operator_imported` pack MUST pass the ordinary TUF pipeline using target deployment trust roots. An embedded `packaged_builtin` pack may be reused only when the target deployment already retains the exact logical version or has a trusted application-release binding for the exact digests; otherwise it is unavailable and MUST NOT be converted into an operator-imported or target-trusted built-in implicitly.

**RP-REQ-227**
An embedded optional pack failure MUST NOT block core incident import unless the incident portability manifest explicitly classifies that exact pack set member as required for the imported artifact. Required-member behavior remains owned by the portability contract.

**RP-REQ-228**
Only `redistribution='allowed'` pack payloads may be embedded. Restricted or prohibited packs must be represented by exact references only.

### Embedded-content inventory and exact requirements

When an incident bundle embeds operator packs or classifies pack members as required, it MUST include `ext/reference_packs/content.json`. This is a closed `reference_pack_content.v1` object with required non-null `schema_id`, `containers`, and `required_members`. It is bounded to 16777216 UTF-8 bytes before JSON admission, including whitespace. Original JSON MUST pass strict duplicate-member and Unicode admission; producer output is canonical JSON followed by one LF. An absent content manifest means no embedded containers and no required members. A present empty or null document is invalid. Empty collections are explicit arrays.

`containers` contains at most 16384 closed objects, each with required non-null `manifest_sha256`, `container_sha256`, and `size_bytes`. Both digests are lowercase 64-digit hexadecimal; size is an integer in the inclusive range 1–9007199254740991 and is additionally subject to the effective deployment's ordinary pack/container and incident-bundle limits. Rows sort by `manifest_sha256` and reject duplicates. Every row MUST resolve to exactly one `operator_imported` version in `reference_pack_refs.v1`. The only container path for that row is `ext/reference_packs/containers/<manifest_sha256>.pack`; the digest token is the exact row value. No alias, supplied path, archive suffix, duplicate version encoding, unreferenced container, or embedded built-in container is admitted by this format. The container format is detected from its bytes. A catalog row may name unavailable optional bytes; it does not prove their availability or trust.

`required_members` contains at most 65536 closed objects, each with required non-null `pack_set_id` and `pack_key`. Each pair MUST resolve to the exact member of an admitted reference set. Rows sort by set ID then key, reject duplicates, and never refer to a floating current selection. Requirement of one member does not imply requirement of other members or sets. If several sets select the same version, any required pair selecting it makes that exact version necessary for import. Required membership is meaningful with references-only transport and does not override redistribution restrictions. Source-artifact owners select these exact requirements; no server infers them from the destination's active set.

The 16384-container ceiling is reachable with 1024 sets, sixteen distinct operator-imported profile versions per set, and no version reuse between sets. With the sixteen declared keys, the maximum distinct required pairs is 16384 (1024 × 16). The generic 65536 required-pair ceiling is therefore unreachable in this profile: a sixty-fifth member or an unsupported seventeenth key fails earlier. Acceptance MUST exercise a complete 16384-container/16384-required-pair catalog, the isolated 65536/65537 cardinality guard, and the independently reachable document-byte equality. The JSON safe-integer size ceiling is an isolated schema guard; effective container limits reject that many actual bytes before extraction. No production limit is increased to manufacture a reachable fixture.

Structural or identity failure of a recognized content manifest fails the reference source family's `reference_pack_refs.exact_shape` or `reference_pack_refs.identity_exact` invariant. Unavailable, inadmissible or unsuccessfully verified optional bytes remain unresolved source references. An unresolved required pair fails `reference_pack_refs.degradation_bounded`. Execution failures, untrusted destination clock for new trust work, stale state and indeterminate publication are operational failures, never optional-content absence. Reuse of an already usable exact local version does not require fresh trust work or inspect an optional redundant embedded container.

An embedded export selects only exact referenced operator-imported versions with `redistribution='allowed'`. It copies an envelope-bound retained container through the Reference Data streaming interface; the transport MUST NOT interpret storage references. The exporter verifies its recorded size and digest before copying and verifies the copied stream again before publishing the enclosing bundle. Missing or invalid optional container bytes leave exact references without an embedded inventory row. Operational read, close or sink failures abort export. Export never fabricates a container or re-signs historical content. The content manifest lists only containers actually included by that export, and preserves exact required pairs independently of embedding. The complete enclosing archive, including generated manifest and checksum members, remains subject to the deployment's incident-bundle byte and member ceilings.

### Destination retention of source reference catalogs

The complete `reference_pack_refs.v1` member is limited to 67108864 UTF-8 bytes, including admitted JSON whitespace; admission checks the bound before JSON parsing. It contains at most 1024 sets and 16384 version rows. Equality is admitted; excess is `reference_pack_refs.exact_shape`. These limits bound portability admission independently from a pack container's limits. The producer emits canonical JSON followed by one LF. Consumer admission requires strict original JSON and canonical array ordering, but does not require canonical object-member order or whitespace.

The destination MUST retain an admitted source reference catalog as immutable historical evidence, including exact optional tuples that are unavailable locally. Re-export MUST preserve those references; it MUST NOT substitute the current active set or erase unresolved members. Source reference metadata never authorizes local content or replaces local first-success provenance.

An incident export MUST union that retained source catalog with the exact sets referenced by the incident's locally retained Reporting snapshots. Reporting supplies immutable bindings through its owner interface in the export transaction; Incident Portability MUST NOT read Reporting or Reference Data tables. The Reference Data owner validates each native binding, including its immutable provenance, and materializes the exact set/member graph. Identical references coalesce. Conflicting facts for one set ID or logical version tuple fail export. No current-set lookup, optional-section selection, or current profile claim may replace or erase a retained snapshot binding. The union remains subject to the same catalog ceilings. Missing required native bytes or binding evidence fails export; unavailable imported optional references remain historical evidence.

Reference Data supplies an opaque preparation and a transaction-bound retention operation to Incident Portability. Preparation validates every local tuple collision before treating availability as optional, compares the declared profile, distribution and source-profile identity with the retained manifest and successful envelope, and checks retained bytes outside publication locks. Publication acquires selected pack-key guards in ascending key order, compares captured revisions, and rejects relevant changes with `stale_admission_state`. It MUST share the incident import's final transaction. A committed identical operation replays before fresh availability or revision checks; a conflicting identity fails.

Reuse leaves current health, administrative disablement, activation, and trust revisions unchanged. Every reusable member receives a destination-envelope pin, including members of an otherwise unavailable set. A complete reusable set is retained with its exact identity and destination first-success provenance and receives a set pin. An incomplete set remains an unresolved exact source reference; no partial member list receives its identity. Removal MUST consider both set pins and individual version pins under the same pack-key guard. Backup includes source catalogs, destination resolution evidence, and both pin forms. Restore validates their identities and linkage before readiness.

The local retention operation uses closed canonical `cartulary.reference_pack_portability_input.v1` with required non-null `schema_id`, `incident_id`, `source_operation_id`, `catalog_sha256`, and `content_manifest`. The latter is the complete admitted `reference_pack_content.v1` object; absence at transport admission materializes its empty object with both arrays present. Required-member classification is retained through re-export even in references-only mode. A prior retained catalog without that evidence requires explicit cutover; no classification is inferred. UUIDs are nonzero canonical lowercase UUIDs. The digest is SHA-256 of the canonical `reference_pack_refs.v1` object without a trailing LF. Its closed result is `cartulary.reference_pack_portability_resolution.v1`, containing exactly required `schema_id` and `versions`. The array has one item per source version, in the same canonical order. Each item contains exactly required `pack_key`, `pack_version`, `available`, and `reason_code`. `available=true` requires explicit null reason. Otherwise the reason is exactly one of `not_retained` (no retained successful logical version), `removed` (an exact removed version), `not_usable` (unusable health or incomplete index), or `content_unavailable` (missing or invalid retained bytes). Execution failures are not optional-content absence and abort retention. These destination facts do not change the source catalog, confer trust on embedded bytes, or weaken required-artifact behavior under RP-REQ-227.

The Reference Data retention operation UUID is UUIDv5 with the source operation UUID as namespace and the exact UTF-8 name `reference_pack:incident_retention`. This separates owner operation identities while preserving attribution and idempotent replay. Successful retention becomes terminal when its catalog, resolution, pins, and incident import commit together. Failed preparation becomes terminal with its attributable rejection or operational-abort evidence and the parent failed or canceled Job; it creates no catalog.

### Frozen destination verification preparation

Transport eligibility precedes fresh pack verification. After validating the closed catalog and inventory, the destination first selects exact reusable local versions. Their redundant embedded bytes are not inspected. For an unavailable version, an embedded descriptor above the effective container-byte limit is ineligible without opening its bytes. Otherwise the destination streams the bytes under that limit and verifies the descriptor's exact size and container digest; disagreement is `reference_pack_refs.identity_exact` and fails the whole import.

Bounded lexical archive/manifest inspection may establish attribution and transport eligibility, never authenticity. A container without a safely attributable manifest tuple is ineligible and creates no pack candidate, verification attempt or content verdict. A safely decoded tuple that contradicts its source reference fails `reference_pack_refs.identity_exact`. When the closed manifest schema is decodable, its manifest digest and declared contract, profile, distribution and source-profile projection must match the catalog; contradiction is the same whole-import identity failure. A decoded redistribution classification other than `allowed` makes those bytes ineligible. Other malformed pack fields remain the ordered verifier's responsibility. Ineligible optional bytes remain unresolved historical references; an ineligible required member fails `reference_pack_refs.degradation_bounded`. The enclosing admitted bundle remains the exact retry input, and no transport denial is represented as a cryptographic verdict.

Only eligible, safely attributable new containers enter the fresh-verification cohort. The cohort retains their completed immutable byte references before executing the complete verification registry. Every fresh candidate still performs all required signature, schema, canonicalization, profile, compatibility and dependency checks; lexical inspection cannot satisfy any of those checks. Existing successful local content is never condemned by transport ineligibility or a failed renewal.

A destination import that needs fresh verification MUST retain one immutable preparation cohort under its Incident Portability Job identity before verification starts. It records the complete canonical source catalog, each original local availability/envelope selection, every new immutable container object and expected transport size/digest, and the relevant key and trust revisions. Later local imports do not join that cohort. Each source version has one fixed ordinal; a selected new container also names its exact operation-member ordinal and tuple. A reused member never receives a new verification attempt merely because redundant bytes are embedded.

Its canonical `cartulary.reference_pack_portable_verification_context.v1` object contains exactly required non-null `schema_id`, `configuration_sha256`, `clock_trusted`, and `timeout_seconds`. The digest is lowercase SHA-256. The boolean is the deployment assertion captured at admission; it has no omission default in persisted input. Timeout is an integer in 60–86400 seconds, captured from the effective configuration (whose default is 1800). The context is bounded to 4096 bytes; its closed canonical form cannot reach that guard under the other scalar constraints. Unknown members, duplicate keys, malformed Unicode, omission, null and noncanonical persisted encoding are invalid. A subsequent attempt retains that context and exact selection but receives a new verification-start instant and monotonic execution budget. The process-local deadline is never persisted.

Verification ordinals use a deterministic topological schedule of exact embedded dependency tuples, with canonical source-version order breaking ties. Bounded lexical declarations establish that schedule only. A previously verified cohort member may satisfy an exact dependency; an unverified, rejected or cyclic predecessor cannot. Remaining cyclic members and their dependants retain canonical source order and receive the ordinary ordered dependency checks. The schedule is retained at admission and is not reconstructed from later local state.

Every member compares its release sequence with the captured destination high-water mark. Successful predecessors do not raise that mark during verification. Two fresh members assigning different identities to the same repository/key/sequence cannot both succeed: the later verification ordinal fails `sequence_collision`. Compatible successes publish in repository/key/numeric-sequence order. Compatible root proposals are merged before any member publication; conflict rolls back the whole parent transaction. A failed optional member receives the retained `not_usable` resolution and attributable failed-attempt evidence without changing an established successful envelope or its health. A failed required member aborts incident publication and publishes no successful pack or trust delta. Its content rejection, diagnostic evidence, failed initial-candidate projection and complete attempt result commit with the failed parent Job. Successful siblings remain unpublished; never-successful sibling candidates close as failed without a content-verdict claim. Existing successful envelopes and health remain unchanged. Frozen revision guards still apply to rejection publication. Changed relevant state, cancellation or timeout converts the uncommitted outcome to the corresponding operational abort without publishing content verdicts.

One process-local execution scope carries the Reference Pack budget and cancellation observation through container preparation, verification, remaining source-owner preparation, publication-lock waiting and the parent finalizer's commit classification. The elapsed budget begins at parent attempt execution; all cohort members share one verification-start UTC instant. Retried parent attempts retain the immutable cohort and receive a new start instant and execution budget. Proven commit remains authoritative. Only proven absence permits cancellation/timeout classification under the shared Extensions policy; equality is timeout, and process shutdown or a lost lease leaves recoverable work. The parent Job, owner abort evidence and failed initial-candidate projections terminate in the same transaction. Incident Portability projects registered Reference Pack operation rejections and verification timeout directly through the common Job error contract.

Restore validates canonical preparation references and context, the source-operation identity, exact selection and verification ordinals, original destination-envelope bindings, descriptor/object size and digest agreement, input object ownership and captured key guards. An operationally aborted cohort retains a closed operational result without an invented catalog. A required-content rejection retains its closed `content_rejected` result without a catalog and must match an exactly equal completed attempt result; successful results without a catalog are invalid. A published cohort requires a completed matching verification attempt and its complete catalog and pins. Neither current freshness nor current availability can rewrite those historical facts.

These preparation facts remain private and participate in backup/recovery. They cannot confer destination trust, publish indexes, change activation, substitute content, or count as successful incident import. Candidate outcomes, compatible trust advancement, local envelope/set pins and the incident's final result require the shared owner-finalization boundary. A references-only reuse operation does not manufacture an empty fresh-verification cohort.

## 23.3 Operational backup and restore

**RP-REQ-229**
In every deployment mode, including when the Reference Pack Extension Profile is unclaimed, operational backup MUST include:

- every non-removed retained pack's logical bytes and every retained current or historical verification envelope required by audit or reverify;
- all reproducibility-pinned versions;
- packaged built-in retained versions and their trusted application-release bindings;
- canonical manifests and applicable TUF metadata;
- complete trusted-root history required to verify retained attestations, the current trusted-root pointer, and highest trusted metadata versions;
- pack attestations and validation summaries;
- active pointers and retained pack sets;
- reproducibility pin relations;
- integrity proofs for every retained byte artifact.

**RP-REQ-230**
Restore MUST verify the integrity and retained-state linkage of every restored pack artifact. An integrity failure in any restored retained artifact fails restore rather than silently discarding that state. For an already verified active or pinned operator pack, restore MUST recheck signatures, metadata linkage, and the recorded invariant `last_verified_at < trust_valid_until` using the retained attestation time; it MUST NOT require the metadata to remain unexpired at the later restore time. Ready state MUST fail when an active or pinned member is absent, its digest differs, its historical trust state cannot be reconstructed, or an active pointer targets `disabled`, `failed`, `missing`, or removed state. Any later activation, reverify, or refresh still applies current-time freshness under RP-REQ-075.

**RP-REQ-231**
Derived search and lookup indexes MAY be omitted from backup. When omitted, restore MUST rebuild them from verified canonical pack bytes before pack-dependent reads are admitted.

Reference Pack objects MUST be read and restored through their own confined storage capability. The generic object-store capability MUST NOT interpret Reference Pack storage references. Restore preserves exact opaque object references, verifies streaming sizes and digests before publication, and rejects occupied destinations. The restore-target marker and verification basis MUST bind the Reference Pack root independently from the database and generic object store. Disposable restore-verification cleanup MUST operate only under the already acquired exclusive target lease and MUST leave source roots unchanged.

# 24. Persistence minima and invariants

## 24.1 Successful verification envelope

The immutable `cartulary.reference_pack_successful_envelope.v1` object MUST be RFC 8785 canonical bytes with no trailing LF. Every member below is required; unknown members are invalid at every envelope and trust-state object boundary. Invalid retained evidence is an integrity failure, never a new content-verification verdict. This persistence representation is separate from the original signed metadata bytes; admitting its shape does not establish destination trust.

| Member | Rule |
| --- | --- |
| `schema_id` | Exact envelope schema ID above. |
| `operation_id` | Nonzero canonical lowercase UUID of the publishing Reference Data operation. |
| `pack_key`, `pack_version` | Exact logical identity under RP-REQ-022 and RP-REQ-023. |
| `distribution_kind` | `operator_imported` or `packaged_builtin`. |
| `manifest_sha256`, `payload_sha256` | Exact lowercase 64-hex logical digests. |
| `container_sha256`, `container_ref` | Lowercase 64-hex container digest and confined opaque storage reference for an operator pack; both explicitly null for a built-in. A storage reference is nonempty, at most 4096 UTF-8 bytes, and follows the confined storage reference grammar. |
| `verified_at` | Nonzero verification-start instant in canonical UTC RFC3339Nano form. |
| `trust_valid_until` | Operator trust expiry, strictly later than `verified_at`; explicitly null for a built-in. |
| `trust_snapshot` | Frozen trust object below for an operator pack; explicitly null for a built-in. |
| `trust_proposal` | Verified trust object below for an operator pack; explicitly null for a built-in. |

`trust_snapshot` contains exactly `root`, `root_history`, and `highest`. `trust_proposal` contains exactly `root`, `root_history`, `metadata`, `signers`, `root_transitions`, `valid_until`, and `binding`. All of these members are non-null. Byte strings are strict standard padded Base64 of the original metadata, at most 2097152 decoded bytes each; alternate encodings are invalid. A `root_history` is a map from canonical positive decimal root version to exact root bytes; each key MUST equal the contained version. A frozen history includes its exact admitted root and never includes a later root. Its aggregate decoded root bytes are at most 10485760, comprising the admitted container metadata allowance plus one current root. A proposal history contains exactly its newly consumed roots. The complete successful envelope is bounded by 67108864 bytes, including Base64 expansion and duplicated bounded metadata. This is an independent retained-evidence bound, not permission to enlarge archive or TUF limits.

`highest` admits only `timestamp`, `snapshot`, and `targets`; an absent role means there is no previously accepted metadata for that role and logical tuple. `metadata` requires all three roles. Each present role value contains exactly positive safe-integer `version` and nonempty `bytes`. `signers` requires exactly `root`, `timestamp`, `snapshot`, and `targets`; each value is a nonempty, strictly ascending list of at most 64 lowercase 64-hex key IDs. `root_transitions` is a required array, empty when no rotation occurred. Each transition contains exactly `version`, `previous_signers`, and `next_signers`; versions advance consecutively from the admitted root, and each signer list follows the same complete sorted-set contract. The proposal root is the last transition root, or the exact admitted root when the list is empty. `valid_until` MUST equal the outer `trust_valid_until`. `binding` is exactly the closed signed Cartulary targets binding, and its key, version, manifest digest, and payload digest MUST equal the envelope. Only the explicitly permitted extension members inside original signed TUF metadata retain their TUF exception; no such exception applies to these persistence wrappers.

## 24.2 Authoritative state

The closed canonical `cartulary.reference_pack_attempt_result.v1` persistence object records operation or attempt completion, not proof that verification started. An interrupted prior attempt contains exactly `schema_id` and `outcome='interrupted'`. Other outcomes require exactly those members plus `member_count` and `failed_count`, both nonnegative JSON-safe integers; failed count cannot exceed member count. `content_rejected` requires positive failed count; `succeeded`, `canceled`, `timed_out`, `stale_state` and `execution_failed` require zero failed count. Unknown outcomes, extra members, omissions, null, duplicate keys and noncanonical persisted bytes are invalid. A 4096-byte admission guard is unreachable for this closed canonical object; acceptance tests its isolated guard without weakening the scalar bounds. Only an actual attempt row establishes a verification-start instant. An operation aborted before verification may record this terminal result without creating an attempt.

Restore and readiness validation MUST validate retained attempt result shapes, the matching stored outcome, exact admitted cohort count, complete member counts for published content verdicts, and successful operation linkage to a completed attempt with the identical result. A terminal operation cannot retain an unfinished attempt. Interrupted private preparation remains operational history; it cannot supply a content verdict. Ordinary import, reverify and refresh inputs retain their closed frozen-input contract through recovery.

Durable verification preparation uses closed canonical `cartulary.reference_pack_prepared.v1` with exactly required, non-null `schema_id`, `manifest`, `envelope`, `objects`, and `index_id`. The schema ID is exact. `manifest` is strict standard padded Base64 of the admitted canonical manifest, at most 1048576 decoded bytes. `envelope` is the §24.1 operator-imported envelope. `index_id` is a nonzero canonical lowercase UUID identifying the completed index generation owned by the same operation.

Each `objects` element contains exactly non-null `id`, `path`, `sha256`, `size`, and `reference`. Object IDs are nonzero canonical lowercase UUIDs; digests are lowercase 64-hex strings; sizes are integers in `0..9007199254740991`; references follow the confined storage grammar and are at most 4096 UTF-8 bytes. Objects sort by logical path as UTF-8 bytes and contain exactly one manifest, every declared payload and notice member, and one `container` descriptor, with no duplicate or undeclared path. There are at most 68 descriptors. Logical member digests and lengths equal the manifest inventory; the manifest and container descriptors equal their envelope bindings. Container size is positive. Repeated object IDs may share identical content bindings only; one physical reference cannot name two different object IDs. The complete prepared object is at most 71303168 bytes. This bound covers the envelope, Base64 manifest and bounded descriptor overhead without increasing any archive limit.

Preparation is private execution evidence and MUST NOT confer successful verification or consumer availability. Publication MUST prove that each operator object was completed and retained by that operation, or, for a renewed container, by the exact successful envelope captured in its frozen cohort. It MUST NOT create an unprepared object's storage record merely because a prepared descriptor names it. The operation ID, cohort key and version, verification-start instant and index operation ownership MUST all match before publication. An inconsistent preparation is an execution integrity failure, not a content rejection.

**RP-REQ-232**
Core 02 MUST require logical persistence for at least:

- immutable pack identity and distribution kind;
- nullable container digest (`null` for packaged built-ins), plus exact manifest and payload digests;
- pack contract, content profile, source profile ID, and source-profile artifact digest;
- source and license provenance;
- retained canonical container or logical member bytes;
- condition, verification result, missing reason, and active pointer;
- pack release sequence and highest accepted sequence;
- trust repository, complete trusted-root history needed by retained attestations, current trusted-root pointer, and highest trusted role versions in the exact scopes defined by RP-REQ-068;
- current successful verification envelope plus all prior container digests and verification attestations retained as history;
- complete verification, activation, fallback, disablement, removal, and integrity attestations;
- dependency and conflict declarations;
- deterministic index identity;
- immutable pack sets;
- reproducibility pins;
- validation summaries;
- physical garbage-collection eligibility.

**RP-REQ-233**
Persistence MUST enforce immutable `(pack_key, pack_version)` logical digests. An update operation MUST NOT replace retained manifest or payload bytes in place.

**RP-REQ-234**
Attestation and administrative audit history MUST be append-only. A later event may supersede the current operational condition but MUST NOT erase the earlier event.

**RP-REQ-235**
Physical table names, index types, object-store key layouts, database extensions, and package directories remain implementation latitude when they preserve the observable contract.

# 25. Security, privacy, and hostile-content boundary

**RP-REQ-236**
Pack verification and consumption MUST perform no outbound network request. This prohibition applies to HTTP, DNS, package resolution, schema resolution, URL dereference, external font or script loading, and any other remote access.[^10]

**RP-REQ-237**
Reference packs MUST NOT contain credentials, secret references, access tokens, or private keys. Their logical member roles and media types MUST NOT declare scripts, executable binaries, native libraries, active HTML, SVG, JavaScript, CSS, macros, installer hooks, runtime templates, executable regular-expression programs, dynamic query programs, symbolic links, or hard links. Every admitted string is inert data. A UI MUST contextually escape it; the pack subsystem MUST never execute, shell-expand, import as code, render as active markup, or pass it to an interpreter. Command and query examples remain inert text.

**RP-REQ-238**
Temporary extraction MUST occur under a newly created private directory beneath `roots.temporary_work`. The directory MUST not be shared between jobs. It MUST be removed after success, failure, cancellation, or timeout.

**RP-REQ-239**
Startup crash recovery MUST identify and remove abandoned temporary pack directories that are not referenced by a live admitted job. It MUST NOT remove retained authoritative pack bytes or an active index.

Private verification workspaces are disposable execution scratch, never frozen inputs or authoritative publication references. A recovered Job uses its frozen container/envelope input and allocates a new workspace. Workspace ownership MUST extend across processes, including operator admission and workers. Creation and collection serialize at the scratch namespace; an attempt holds an exclusive directory-inode lease until it seals and closes its workspace. A collector removes only unleased workspace directories within that namespace, using confined, non-recursive regular-file deletion and bounded directory enumeration. An unleased interval during creation MUST NOT be observable to collection. Process death releases the ownership lease; wall-clock age and PID reuse MUST NOT decide liveness. Live work is skipped. Unexpected names, links or file types fail closed. The sweep never addresses incoming uploads, durable operation inputs, published objects, or indexes. Startup reconciliation runs before readiness, and deployment-local operator admission uses the same ownership boundary. Internal file descriptors MUST be close-on-exec to prevent subprocesses from inheriting storage authority or retaining workspace leases.

**RP-REQ-240**
Persistent and temporary pack bytes MUST receive the encryption-at-rest and filesystem protection required by Core 04 for deployment state. Private signing keys MUST never be stored in a pack or trust-bootstrap file.

**RP-REQ-241**
A trust-root update is valid only through signed sequential TUF root processing. No routine administrator, browser action, environment override, or pack manifest field may bypass threshold validation.

**RP-REQ-242**
The browser administration surface MUST display semantic pack state and provenance, not storage internals. It MUST NOT display raw signatures, private keys, object-store keys, staging paths, temporary paths, or complete payload entries by default.

# 26. Resource limits and failure behavior

**RP-REQ-243**
Core 04 MUST expose the effective limits below. Omission resolves to the stated default. Byte-count and count values have valid configured domain `1..9223372036854775807`; compression ratio has domain `1..1000`.

**Table 26-A. Configurable archive and container limits**

| Configuration key | Default | Applicability |
| --- | ---: | --- |
| `limits.reference_packs.max_container_bytes` | `536870912` | Exact admitted operator container bytes. |
| `limits.reference_packs.max_extracted_bytes` | `536870912` | Extracted operator bytes and packaged built-in logical member bytes. |
| `limits.archives.max_compression_ratio` | `100` | Operator-imported compressed containers. |
| `limits.archives.max_members` | `10000` | Regular files for either distribution kind. |

Equality at the effective configured value is valid. Exceeding `max_container_bytes` fails with `container_bytes_exceeded`; the other three limits use their existing exact archive-limit reasons. Packaged built-ins do not apply compression-ratio validation.

**RP-REQ-244**
The fixed v1 subsystem limits are:

**Table 26-B. Fixed subsystem limits**

| Value | Maximum |
| --- | ---: |
| JSON object/array nesting depth | `10000` |
| Trust-bootstrap file bytes | `8388608` |
| Manifest-declared files | `66` |
| `bundle.json` bytes | `16384` |
| `manifest.json` bytes | `1048576` |
| One TUF metadata file | `2097152` |
| Total TUF metadata bytes | `8388608` |
| Complete successful verification envelope bytes | `67108864` |
| One payload or notice member | `268435456` |
| Empty archive directory markers | `10000` |
| TAR terminating zero blocks, including the mandatory pair | `20` |
| Archive path bytes | `1024` |
| Path-segment bytes | `255` |
| Source artifacts | `64` |
| Dependencies | `64` |
| Conflicts | `64` |
| TUF keys in one root | `64` |
| Signatures in one ordinary metadata file | `64` |
| Signatures in one root-update metadata file | `128` |
| Entries in one entries profile | `2000000` |
| Framework objects | `2000000` |
| Framework relationships | `5000000` |
| Aliases per entry or object | `64` |
| Source refs per entry, object, or relationship | `64` |
| Any profile-specific array per entry or object | `64` |
| Pack-set members | `64` |
| NDJSON line bytes including LF | `1048576` |
| Retained validation issues | `1000` |

Equality at a maximum is valid. A value greater than a maximum is invalid.

**Table 26-C. Exact limit failure mapping**

| `limit_id` | Limit source | Required outcome |
| --- | --- | --- |
| `max_container_bytes` | Table 26-A | `reference_pack_verification_failed/container_bytes_exceeded` |
| `max_extracted_bytes` | Table 26-A | `reference_pack_verification_failed/archive_extracted_bytes_exceeded` |
| `max_compression_ratio` | Table 26-A | `reference_pack_verification_failed/archive_compression_ratio_exceeded` |
| `max_members` | Table 26-A | `reference_pack_verification_failed/archive_member_count_exceeded` |
| `max_json_nesting_depth` | Table 26-B | The original-JSON admission error for the enclosing document: invalid bootstrap configuration, `manifest_json_invalid`, `bundle_hint_invalid`, `tuf_metadata_invalid`, or `content_schema_invalid`. No later schema or signature check runs. |
| `max_trust_bootstrap_bytes` | Table 26-B | Invalid deployment configuration before workers start. |
| `max_manifest_files` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_bundle_json_bytes` | Table 26-B | `reference_pack_verification_failed/bundle_hint_invalid` |
| `max_manifest_json_bytes` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_metadata_file_bytes` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_metadata_total_bytes` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_member_bytes` | Table 26-B | `reference_pack_verification_failed/archive_structure_invalid` |
| `max_directory_markers` | Table 26-B | `reference_pack_verification_failed/archive_structure_invalid` |
| `max_tar_terminating_zero_blocks` | Table 26-B | `reference_pack_verification_failed/archive_structure_invalid` |
| `max_path_bytes` | Table 26-B | `reference_pack_verification_failed/archive_structure_invalid` |
| `max_segment_bytes` | Table 26-B | `reference_pack_verification_failed/archive_structure_invalid` |
| `max_source_artifacts` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_dependencies` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_conflicts` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_root_keys` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_signatures` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_root_update_signatures` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_entries` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_objects` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_relationships` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_aliases` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_source_refs` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_profile_array` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_set_members` | Table 26-B | `reference_pack_operation_rejected/contract_incompatible` for an activation proposal; current sixteen-key profile cannot reach this guard. |
| `max_line_bytes` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_issues` | Table 26-B | Truncate under RP-REQ-094; no failure. |
| `max_verification_seconds` | RP-REQ-245 | `reference_pack_verification_failed/verification_timeout` |

**RP-REQ-245**
Core 04 MUST add:

```text
limits.reference_packs.max_verification_seconds
```

The default is `1800`; valid values are integer seconds in `60..86400`, inclusive. Omitted configuration resolves to `1800`; explicit null, fractional values, and values outside the range are invalid configuration. The deadline boundary and outcome classification are defined only by RP-REQ-098.

**RP-REQ-246**
A limit MUST be checked at the earliest phase where it can be determined safely. A limit breach MUST stop further untrusted parsing that is not required to identify the failure and clean temporary state. It MUST publish no usable content or trust advancement. An attributable initial import still retains failed candidate and attempt evidence under the operation outcome matrix; an established version follows that matrix's renewal or reverify rules.

**RP-REQ-247**
Every limit breach MUST use the exact outcome in Table 26-C. When the outcome is a verification issue, `code` and `reason_code` equal the reason after `/`; `safe_details.limit_id` equals the Table 26-C token; and `expected_token` and `actual_token` contain the unsigned base-10 maximum and observed count or byte length. The compression-ratio guard compares allowed extracted bytes (complete admitted container bytes multiplied by the ratio) with observed extracted bytes. A streaming guard reports the first observed lower bound proving the breach, without reading further to count the final total. Original-JSON depth admission is an exception: it returns the enclosing document's malformed-JSON finding with null limit details, before a parsed object exists. Invalid declared manifest summary counts and file-size claims are manifest schema failures; the payload row and actual member guards use their Table 26-C outcomes. Timeout is the operational outcome in RP-REQ-098 and has no content-validation issue. A limit MUST NOT be remapped according to implementation phase or parser choice. The container streaming admission guard stops on the first byte proving excess and returns the Core 01 synchronous import rejection before Job admission. This guard cannot establish complete-container verification precedence: the ordered check registry still governs every fully admitted container. The early rejection has no safely attributed candidate, Job, attestation, persisted summary, trust update or published bytes. Its inline summary uses the existing `container_bytes` check and a single finding at `$`; its observed count is the first proven lower bound. Operational read, write, cancellation or cleanup failures remain operational errors.

**RP-REQ-248**
Before adoption, a valid canonical fixture for every current pack key MUST fit within every applicable limit. A fixture that exceeds a limit blocks adoption; implementations MUST NOT carry private wider limits.

# 27. Administrative audit, attestation, observability, and UI state

**RP-REQ-249**
For every pack-attributable event after `(pack_key, pack_version)` is safely decoded, the subsystem MUST create an append-only attestation whose `event_kind` is exactly one of:

```text
import_verification
reverification
refresh_verification
activation
rollback_activation
safety_fallback
profile_reconciliation
dependency_invalidation
disablement
removal
exact_reimport
payload_invalidation
trust_root_update
```

An operation rejected before pack identity is safely attributable MUST create no pack attestation; its common job result and administrative audit event remain required. `trust_root_rejection` is an administrative-audit event, not a pack-attestation token.

**RP-REQ-250**
A verification attestation MUST be a closed `cartulary.reference_pack_attestation.v1` object containing exactly:

| Member | Rule |
| --- | --- |
| `schema_id` | Exactly `cartulary.reference_pack_attestation.v1`. |
| `attestation_id` | Derived as specified below. |
| `event_kind` | One RP-REQ-249 token. |
| `pack_key`, `pack_version` | Exact logical identity. |
| `distribution_kind`, `verification_method` | Exact applicable registry values. |
| `result` | `succeeded`, `failed`, or `rejected`. |
| `occurred_at` | Fixed Core timestamp. |
| `actor_kind` | `user`, `local_operator`, `system`, or `application_release`. |
| `actor_user_id` | Non-null only for `user`. |
| `operator_operation_id` | Non-null only for `local_operator`. |
| `container_sha256` | Non-null only for `import_verification`, `reverification`, or `refresh_verification` when one operator container was evaluated; null for associated root-update, exact-reimport and other administrative events. |
| `manifest_sha256`, `payload_sha256` | Non-null when logical identity was verified or previously retained. |
| `source_profile_id`, `source_profile_sha256` | Non-null when manifest identity was verified or previously retained. |
| `trust_repository_id` | Non-null only for operator-imported packs. |
| `trusted_metadata_versions` | Closed object with `root`, `targets`, `snapshot`, and `timestamp` integer-or-null members. |
| `verified_signer_key_ids` | Sorted unique signer IDs; `[]` for packaged built-ins or pre-signature rejection. |
| `trust_valid_until` | Non-null only after successful operator verification. |
| `validation_summary_ref` | Non-null only when a validation summary exists. |
| `prior_active_version` | Exact version string or `null`. |
| `resulting_pack_set_id` | Exact pack-set ID or `null` when no set was published. |

`attestation_id` MUST equal `"rpa_" + lowercase_hex(SHA256(RFC8785_JCS(attestation_without_attestation_id)))`. Route-idempotent replay MUST reuse the original attestation rather than create a second object with the same ID.

**RP-REQ-251**
Administrative audit MUST cover import admission, verification completion, activation, rollback, safety fallback, disablement, reverify, refresh, root import, removal, exact reimport, trust-root update, trust-root rejection, sequence rejection, and dependency invalidation.

**RP-REQ-252**
Before this NLSpec is adopted, the OpenTelemetry NLSpec MUST register exactly these Reference Pack operation names:

```text
reference_pack.import
reference_pack.verify
reference_pack.reverify
reference_pack.activate
reference_pack.disable
reference_pack.refresh
reference_pack.reconcile
reference_pack.remove
reference_pack.invalidate
reference_pack.collection
reference_pack.lookup
```

The OpenTelemetry NLSpec remains the signal-shape and attribute owner; absence of this companion registration blocks `RP-GATE-007`.

**RP-REQ-253**
Permitted low-cardinality telemetry attributes are limited to operation kind, pack kind, content-profile ID, distribution kind, verification method, result class, failure phase, and bounded counts or duration. Telemetry MUST NOT emit pack payload values, lookup values, raw pack keys when the deployment classifies them as sensitive, source URLs containing credentials, signatures, keys, paths, or incident data.

**RP-REQ-254**
The administration UI MUST distinguish:

```text
staged
verified_available
active
disabled
failed
missing
```

It MUST display content profile, source version, source-as-of date, logical digests, last verification, trust-valid-until when applicable, previous active version, missing reason, dependencies, and reproducibility-pin removal blockers. These UI fields are projections over owner state and do not authorize actions. The required nullable Core `fallback_from_version` projection identifies the displaced version only on the active packaged registry selected by safety fallback. It is retained through unchanged claimed-profile startup, cleared by a different selection or unclaimed reconciliation, and published in the same transaction as the selection and attestation. It MUST NOT be reconstructed by sorting event wall-clock timestamps.

# 28. Core companion amendments and adoption status

**RP-REQ-255**
The coordinated owner requirements in Table 1-A MUST be satisfied before this NLSpec is adopted through the existing document-status process. This section adds no second gate ledger or recurring approval record. Core owns recognition, public routes and errors, source persistence, security, configuration, portability and recovery; participating subsystem owners adopt their named contracts together. Implementation verification does not itself establish adoption.

# 29. Conformance fixtures and harness obligations

**RP-REQ-256**
The Testing Harness owner MUST register one closed `cartulary.reference_pack_fixture_manifest.v1` object per fixture:

| Member | Type | Rule |
| --- | --- | --- |
| `schema_id` | string | Exactly `cartulary.reference_pack_fixture_manifest.v1`. |
| `fixture_id` | string | Matches `^rpfx_[a-z0-9_]{1,96}$`. |
| `fixture_version` | integer | `1..9007199254740991`. |
| `fixture_family` | string | Exact token from RP-REQ-257. |
| `input_refs` | string array | `1..64` unique normalized POSIX repo-relative paths sorted by UTF-8 bytes; absolute paths, empty segments, `.`, `..`, backslashes, and symlinks are invalid. |
| `expected_container_sha256` | string or null | Exact digest or `null` when the fixture has no operator container. |
| `expected_manifest_sha256` | string or null | Exact digest or `null` when manifest decoding cannot succeed. |
| `expected_payload_sha256` | string or null | Exact digest or `null` when payload identity cannot be computed. |
| `expected_pack_set_sha256` | string or null | Exact digest or `null` when no set publication is expected. |
| `expected_condition` | string or null | One RP-REQ-081 condition or `null` when no version resource exists. |
| `expected_issues` | array | Exact ordered `reference_pack_issue.v1` objects; `[]` when none. |
| `expected_public_error` | object or null | `null` on success; otherwise exactly `code` and nullable `reason_code`. |
| `expected_side_effects` | string array | Sorted unique subset of the side-effect registry below. |
| `forbidden_side_effects` | string array | Sorted unique subset of the side-effect registry below and disjoint from `expected_side_effects`. |

The fixture side-effect registry is closed:

```text
job_admitted
candidate_created
condition_changed
attestation_appended
trust_state_changed
index_published
active_pointer_changed
pack_set_published
audit_event_appended
payload_bytes_deleted
temporary_state_removed
network_request_attempted
```

A digest not applicable to a fixture is explicit JSON `null`. Unknown object members or side-effect tokens are invalid.

**RP-REQ-257**
The required fixture families are exhaustive:

| Family | Required coverage |
| --- | --- |
| `container_equivalence` | Same logical pack in ZIP, TAR, and GZIP-TAR. |
| `archive_attack` | Traversal, links, devices, sparse entries, ZIP64, non-ustar TAR, duplicate paths, case collision, parent-child collision. |
| `json_admission` | BOM, invalid UTF-8, duplicate members, invalid Unicode, noncanonical JSON, unknown members. |
| `manifest_identity` | Key/version mismatch, inventory mismatch, count mismatch, digest mismatch. |
| `tuf_success` | Valid threshold, metadata chain, and root update. |
| `tuf_failure` | Untrusted root, invalid root rotation, threshold failure, expiry, expiry-policy violation, rollback, mix-and-match, missing target, and unexpected target. |
| `version_identity` | Exact replay, version collision, sequence collision, lower unseen sequence. |
| `content_profile` | Valid, malformed, and semantic-error fixtures for every Table 13-A row. |
| `type_compatibility` | Removed key, reused key, ambiguous alias, changed algorithm digest. |
| `dependency` | Missing exact dependency, cycle, conflict, transitive invalidation. |
| `atomic_publication` | Old/new reader consistency and injected failure. |
| `base_fallback` | Required imported registry fails and packaged built-in becomes effective atomically. |
| `builtin_reconciliation` | Claimed and unclaimed startup branches, application-release replacement, optional invalidation, and retained imported history. |
| `refresh` | No discovery, no import, no network, no-op, state-transition matrix, and all-outcome publication. |
| `verification_envelope_renewal` | Different container, identical logical bytes, monotonic metadata, retained prior envelope history, and failed renewal rollback. |
| `removal` | Active, built-in, pinned, pending, successful tombstone, exact reimport. |
| `consumer` | Ordering, cursor binding, no-hit, unavailable pack, indicator algorithm provenance. |
| `snapshot_reporting` | Pinned set survives later activation and rerender uses the pin. |
| `portability` | Embedded pack reverified and inactive. |
| `backup_restore` | Missing active or pinned pack blocks readiness. |
| `resource_limit` | At-limit success and one-over failure for every limit. |
| `cancellation_timeout` | No partial publication and temporary cleanup. |
| `no_egress` | Network access unavailable during every operation. |
| `licensing` | Valid SPDX, invalid expression, missing LicenseRef notice, embedding restriction. |
| `source_profile` | Stable ID and digest, byte change, output-affecting ID change, unknown runtime pair, and canonical producer fixtures. |
| `telemetry_privacy` | No forbidden data leakage. |

**RP-REQ-258**
Every Table 13-A content profile MUST have at least one canonical valid fixture, one malformed-structure fixture, and one semantic-error fixture. The three packaged built-in registry fixtures MUST be byte-identical release inputs and MUST publish their exact digests.

Research reports, implementation guides, screenshots, manual demonstrations and fixture-routing mappings do not substitute for executed canonical fixture evidence. Every discovered fixture MUST have an expectation runner or a participating-owner execution binding; all required families and all three case classes for every supported profile MUST be covered. Fixture count totals do not establish completeness.

**RP-REQ-259 — retired**
The accounting-only requirement-range projection is removed. This identifier is reserved and MUST NOT be reassigned.

**RP-REQ-260**
Reference Pack performance measurements are engineering information unless Core 05 claim-publication requirements are separately satisfied. This NLSpec creates no timed public claim.

# 30. Assumptions, blockers, and future-only areas

**RP-REQ-261**
The current v1 trust profile assumes Ed25519 is permitted. If a FIPS-validated cryptographic module requirement applies, `RP-GATE-010` fails and this document MUST be revised with a new verification-method ID and exact cryptographic profile. An implementation MUST NOT silently substitute ECDSA or RSA under `tuf_1_0_35_offline_bundle_v1`.

**RP-REQ-262**
The deployment clock assertion is the current contract in RP-REQ-075: `reference_packs.clock_trusted` defaults to `false`. While false, fresh operator TUF verification and activation MUST reject with `reference_pack_operation_rejected/clock_untrusted`. Packaged Base verification, historical consumption and a zero-cohort refresh remain available; false does not condemn retained content. No time-service health inference or later trusted-time contract is required to configure this explicit assertion.

**RP-REQ-263**
Every project-distributed external-data pack depends on a stable source version or snapshot, source artifact digests, and an approved license classification. An unavailable source identity or unresolved licensing decision blocks distribution of that pack but does not weaken runtime validation of other profiles.

**RP-REQ-264**
The following areas are future-only. Their current omission behavior is closed:

| Future area | Current omission behavior |
| --- | --- |
| Version ranges | Rejected. |
| Live mirrors or network update | No operation exists. |
| Mutable local overlays | No resource exists. |
| Delegated TUF roles | Metadata is rejected. |
| Browser trust-root management | No route exists. |
| Emergency out-of-band root replacement | No operation exists. |
| Trust-repository removal | No operation exists; configured repositories remain retained. |
| Pack-provided algorithms or scripts | Content is rejected. |
| Pack-defined workbook surfaces | No surface is created. |
| Template packs | Remain reporting-owned. |
| Cross-repository dependency resolution | Dependencies resolve only against the one active deployment set. |
| Multi-pack activation request | One target per activation request. |
| Algorithm negotiation | Exact v1 algorithms only. |
| FIPS verification profile | Unsupported until a new method is adopted. |
| Raw source-profile execution | Runtime never executes it. |
| Remote provider-managed repositories | Unsupported. |
| Automatic optional-pack fallback | No fallback occurs. |

# 31. Acceptance criteria

A conforming implementation and an adopted document set MUST satisfy every criterion below.

**Table 31-A. Binary acceptance criteria**

| ID | Binary acceptance criterion |
| --- | --- |
| `RP-AC-001` | Core and subsystem ownership are non-overlapping and exhaustive under Tables 1-A and 1-B. |
| `RP-AC-002` | The dependency registry contains no unresolved adopted dependency or use of `latest`. |
| `RP-AC-003` | Base resolves all three required registries from packaged immutable pack bytes using the same content schemas as imported replacements. |
| `RP-AC-004` | Packaged built-ins are verified before readiness, retained with release bindings across upgrades, reconciled deterministically for claimed and unclaimed startup branches, non-disableable, and non-removable. |
| `RP-AC-005` | Mutable local registry overrides are unavailable. |
| `RP-AC-006` | ZIP, TAR, and GZIP-TAR representations of one operator logical pack produce identical logical digests, and repeated canonical production from the same declared inputs produces byte-identical logical members. |
| `RP-AC-007` | Encrypted, multi-disk, ZIP64, concatenated, arbitrary GZIP, non-ustar TAR, and structurally invalid containers fail before publication. |
| `RP-AC-008` | Absolute paths, traversal, links, devices, sparse files, duplicates, case collisions, and parent-child collisions fail before content parsing. |
| `RP-AC-009` | Invalid UTF-8, BOM, duplicate JSON members, invalid Unicode, noncanonical JSON, and unknown closed-object members fail deterministically. |
| `RP-AC-010` | Manifest and payload digests reproduce RP-REQ-054 and RP-REQ-055 exactly. |
| `RP-AC-011` | Every operator-imported non-metadata regular file is signed as a TUF target and is either declared by the manifest or rejected; every packaged built-in member is release-manifest-bound or rejected. |
| `RP-AC-012` | An operator-imported pack without a valid TUF threshold cannot become `verified_available`. |
| `RP-AC-013` | Untrusted root, invalid root rotation, insufficient threshold, expiry, expiry-policy violation, rollback, mix-and-match, missing-target, and unexpected-target fixtures return their exact reason. |
| `RP-AC-014` | Passive time passage does not mutate an existing active set; later activation or reverify applies current expiry rules. |
| `RP-AC-015` | Same key/version with different logical bytes fails and leaves retained state unchanged. |
| `RP-AC-016` | A lower unseen release sequence fails, while explicit activation rollback to a retained version succeeds and is attested. |
| `RP-AC-017` | Every content profile in Table 13-A has a closed schema and all required fixture families. |
| `RP-AC-018` | Pack-provided executable, active, template, or regex-program content is rejected or treated only as inert profile text where explicitly allowed. |
| `RP-AC-019` | The indicator registry contains exactly the nine current Core types and applies the exact algorithms in §14. |
| `RP-AC-020` | Removing or reusing a referenced registry key is rejected as incompatible. |
| `RP-AC-021` | Indicator identity-affecting behavior cannot change under the same algorithm ID or behavior digest. |
| `RP-AC-022` | Missing dependencies, dependency cycles, and active conflicts reject activation. |
| `RP-AC-023` | Activation publishes pointer, attestation, dependency effects, index invalidation, audit outbox, and pack set atomically. |
| `RP-AC-024` | Failure of an imported required Base registry atomically activates the packaged built-in fallback. |
| `RP-AC-025` | Concurrent readers observe either the complete old pack set or the complete new pack set. |
| `RP-AC-026` | A job admitted under one pack set continues to use that set after later activation. |
| `RP-AC-027` | Loss or replacement of an active dependency invalidates every unsatisfied transitive dependent before another dependent operation succeeds. |
| `RP-AC-028` | Refresh performs no discovery, import, network fetch, latest selection, or automatic activation. |
| `RP-AC-029` | A zero-version refresh selection completes as a deterministic successful no-op. |
| `RP-AC-030` | Local operator import admits exactly one root-relative regular file through the ordinary verification pipeline and never activates it. |
| `RP-AC-031` | Removal rejects active, built-in, pinned, pending, and already removed versions. |
| `RP-AC-032` | Successful removal preserves a queryable tombstone and changes no unrelated version. |
| `RP-AC-033` | Exact reimport can restore a removed version; reverify alone cannot. |
| `RP-AC-034` | Pack consumers cannot bypass the service boundary by reading extracted files, storage objects, or tables directly. |
| `RP-AC-035` | Lookup ordering, limits, confidential cursor binding, 900-second expiry, no-hit, unavailable-pack, and missing-entry behavior match §21. |
| `RP-AC-036` | Indicator evaluation returns exact applied algorithm IDs and pack provenance and never logs raw input. |
| `RP-AC-037` | Framework and enrichment results carry exact provenance and never automatically mutate incident records. |
| `RP-AC-038` | Optional pack absence never blocks timeline capture, entity resolution, evidence attachment, or core editing. |
| `RP-AC-039` | Snapshots and report operations retain exact pack-set ID and digest. |
| `RP-AC-040` | Rerender fails when a pinned pack cannot be reconstructed rather than substituting the current active version. |
| `RP-AC-041` | Newly embedded incident-bundle packs are independently reverified and remain inactive. Exact local reuse preserves its existing active or inactive selection; source activation and source attestations never establish destination trust. |
| `RP-AC-042` | Restricted and prohibited packs are references-only in incident portability. |
| `RP-AC-043` | Restore verifies every retained artifact and cannot enter ready state with a missing, invalid, or historically unverifiable active or pinned pack; current-time metadata expiry is not retroactively applied to a valid retained attestation. |
| `RP-AC-044` | Import, verify, activate, refresh, remove, and lookup perform no outbound network access. |
| `RP-AC-045` | Operational aborts leave no partial semantic delta. Content rejection follows the operation-specific RP-REQ-091 table, including atomic reverify invalidation and complete mixed-refresh outcomes. Lost commit acknowledgement replays the proven committed result. |
| `RP-AC-046` | Every fixed and configurable limit has equality and one-over guard evidence. When equality is reachable under all other constraints, a complete accepted fixture reaches it. Otherwise a recorded constraint argument explains unreachability and an isolated guard fixture verifies the counter without widening production limits. |
| `RP-AC-047` | Every distributed pack carries valid license, source, builder, immutable source-profile ID and digest, transformation, and notice provenance. |
| `RP-AC-048` | Logs, telemetry, job summaries, issues, and audit contain no payload values, private keys, raw signatures, secrets, uncontrolled paths, or incident data. |
| `RP-AC-049` | Retired accounting-only criterion; identifier reserved and not reassigned. |
| `RP-AC-050` | No current normative behavior is defined only in a guide, appendix, generated artifact, research report, or implementation-local convention. |
| `RP-AC-051` | Every attributable event creates the exact closed append-only attestation; pre-attribution rejection creates none. Operation replay reuses its original attestation and preserves complete authorized signer sets. |
| `RP-AC-052` | Every required administrative audit event is emitted with its initiating actor and operation identity; replay creates no duplicate semantic event. |
| `RP-AC-053` | The telemetry owner registers every required operation name, and emitted events use only the permitted attributes and owner-defined signal shapes. |
| `RP-AC-054` | The administration UI represents every durable and derived state, disabled health, historical provenance versus current verification, fallback degradation, pending work, removal blockers, mixed refresh, uncertain outcomes, and replay without offering ineligible actions. |

Count-ceiling reachability under the current closed payload schemas is constrained by the 268435456-byte per-file ceiling. Every entry requires at least 270 JSON bytes, every framework object at least 277, and every relationship at least 246, before its required NDJSON LF. Consequently 2000000 entries require at least 542000000 bytes, 2000000 objects at least 556000000 bytes, and 5000000 relationships at least 1235000000 bytes. Each profile uses one corresponding payload file, so no complete admissible container can reach these count ceilings. Acceptance retains isolated equality/one-over counter tests and an executable conservative lower-bound proof over the typed row schemas. Schema expansion that makes equality reachable requires a complete accepted fixture; it does not justify raising file limits.


# 32. Definition of done

The pre-production implementation is complete when the behavioral and evidence obligations below are satisfied. Promotion to `adopted/current` additionally requires coordinated owner adoption under Table 1-A; passing a release-check target does not publish that claim. Separate recreation, interchangeability, recurring security-disposition and per-requirement dossiers are not required. Ordinary review of changed behavior, trust/clock admission, applicable redistribution approval and target cryptographic suitability remain required. Independent canonical byte, identity, normalization and signature vectors remain required evidence.

Implementation acceptance requires the following behavioral evidence. Coordinated adoption is governed once, by Table 1-A and the existing document-status process:

- every schema ID and algorithm ID resolves exactly once;
- every object member has type, requiredness, nullability, default, and unknown-member behavior;
- every collection has bounds, duplicate behavior, and ordering;
- every operation has deterministic admission, cancellation, timeout, publication, and error behavior;
- every current pack key has a valid fixture, malformed fixture, and semantic-error fixture;
- every limit has boundary fixtures;
- every public reason mapping is adopted by Core 01;
- independent vectors and the coordinated implementation acceptance suites pass;
- no normative `TODO:`, open delegation, or contradictory owner statement remains.

## Sources

[^1]: `R05-responsive-interface-design-report.cr.md`, thesis and staged-interface findings, lines 5-19 and 133-147. Source limit: research supports nonblocking semantic progress presentation but does not define Reference Pack runtime contracts.

[^2]: JSON Schema, `Draft 2020-12`, published 16 June 2022, https://json-schema.org/draft/2020-12. Accessed 16 July 2026.

[^3]: RFC 8785, *JSON Canonicalization Scheme (JCS)*, especially canonical representation, deterministic property sorting, UTF-8 generation, and invalid-Unicode failure, https://www.rfc-editor.org/rfc/rfc8785.html.

[^4]: The Update Framework Specification 1.0.35, last modified 15 July 2026, especially §§1.5, 2.1, 2.3, 4, and 5, https://theupdateframework.github.io/specification/v1.0.35/.

[^5]: SPDX Specification 3.0.1, https://spdx.github.io/spdx-spec/v3.0.1/; SPDX License List 3.28.0, dated 20 February 2026, https://spdx.org/licenses. Accessed 16 July 2026.

[^6]: `R03-Kanvas_technical_research_report.md`, locally downloaded reference-data families and missing signature/checksum verification, lines 911-915 and 1036-1040.

[^7]: `00_document_set_status_and_precedence.md`, §§4.2 and 5.1, lines 77-109 and 146-193; `01_architecture_storage_and_view_contracts.md`, §11 and §17.4, lines 5790-6012 and 7347-7497; `02_domain_model_schema_and_history.md`, §11 and §14.1, lines 1484-1499 and 1987-2026; `03_workbook_interaction_collaboration_and_workflows.md`, §2; `04_security_deployment_and_conformance.md`, §4.1, §9.4, and §12.3.1, lines 470-490, 1130-1150, and 2072-2135. Source limit: these files define the current outer profile and owner boundaries; the companion amendments required by this draft have not been applied by this artifact.

[^8]: `nlspec-spec.md`, “Behavioral Completeness,” “Unambiguous Interfaces,” “Explicit Defaults and Boundaries,” “Mapping Tables for Translation,” “Testable Acceptance Criteria,” and “Spec Economy,” lines 11-122 and 142-216.

[^9]: `R06-spreadsheet_of_doom_dfir_research_report.md`, reference-vocabulary separation and optional nonblocking enrichment, lines 454-463 and 490-500. Source limit: research evidence supports the product boundary but does not define runtime pack semantics.

[^10]: `R01-aurora_incident_response_report.md`, direct renderer-side external integrations, remote runtime dependency, global invalid-certificate acceptance, and the resulting expanded trust boundary, lines 94-98 and 1243-1274. Source limit: the report supports the no-egress and strict trust-boundary rationale; it does not define this subsystem's protocol.
