---
title: Cartulary Reference Pack Subsystem NLSpec
status: draft
document_class: nlspec
document_version: 0.1.0
contract_major: 1
profile_id: reference_pack
schema_id: cartulary.reference_pack_subsystem_nlspec.v1
---

# 1. Status, scope, authority, and promotion conditions

This NLSpec defines the Cartulary Reference Pack subsystem. It specializes the existing `reference_pack` extension profile and defines the limited Base Profile contract for immutable built-in type registries. It does not create a second extension identity.[^7]

The document remains `status: draft` until every promotion condition in Table 1-A is adopted by its named owner. A draft implementation may use this document for development, but it must not claim Reference Pack Extension Profile conformance while any required promotion condition remains open.

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

**Table 1-A. Promotion conditions**

| Gate ID | Owner | Required adopted state |
| --- | --- | --- |
| `RP-GATE-001` | Core 00 | Core 00 lists this NLSpec as adopted for the exact boundary in Table 1-B and makes the `reference_pack` claim depend on it. |
| `RP-GATE-002` | Core 01 | Core 01 distinguishes packaged Base registries from optional pack administration and imports this NLSpec's lifecycle specialization. |
| `RP-GATE-003` | Core 01 | Core 01 adds the removal route, required resource fields, public reason-code additions, refresh semantics, and reference-pack-set snapshot/release bindings named by this NLSpec. |
| `RP-GATE-004` | Core 02 | Core 02 replaces undefined mutable local overrides with immutable signed replacement packs and adopts the persistence minima in §24. |
| `RP-GATE-005` | Core 04 | Core 04 adopts the trust-bootstrap key, verification timeout key, security rules, and conformance mapping in this NLSpec. |
| `RP-GATE-006` | Reporting Subsystem NLSpec | Reporting binds every pack-sensitive derivation to the snapshot's exact `reference_pack_set_id`. |
| `RP-GATE-007` | OpenTelemetry NLSpec | The Reference Pack operation and safe-attribute registry in §27 is adopted by the telemetry owner. |
| `RP-GATE-008` | Testing Harness NLSpec | Generated schemas, canonical fixtures, malicious-container fixtures, cryptographic fixtures, and drift checks in §29 are adopted by the harness owner. |
| `RP-GATE-009` | `domain.md` | The concepts in Table 5-A are added without redefining this NLSpec's behavior. |
| `RP-GATE-010` | Security authority | The security authority confirms that the Ed25519-only v1 profile is permitted for the target compliance posture. |
| `RP-GATE-011` | Licensing authority | Every project-distributed external-data pack has an approved redistribution classification and required notices. |
| `RP-GATE-012` | Corpus maintenance | Every `RP-REQ-*` maps to at least one `RP-AC-*` or canonical fixture, and no adopted normative placeholder remains. |

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
A canonical or public object defined as closed MUST reject unknown members. An object may permit extension members only through an explicit `extensions` member defined by this NLSpec.

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
The dependencies in Table 4-B are exact. A newer upstream version MUST NOT be substituted without revising this NLSpec or the named owner artifact. Before promotion to `adopted/current`, every Core and adopted-subsystem row MUST name the exact adopted repository revision identifier or content digest that includes the companion amendments in Table 28-A. The current section locators are draft discovery locators and do not satisfy adoption by themselves.

**Table 4-B. Normative dependency registry**

| Dependency | Imported contract | Required baseline |
| --- | --- | --- |
| Core 00 | Profile, precedence, owner matrix | Current adopted Core 00, §§4.2 and 5.1. |
| Core 01 | Routes, jobs, outer lifecycle, snapshots, portability, backup | Current adopted Core 01, §§11, 12, and 17.4. |
| Core 02 | Type registries, indicator identity, persistence minima | Current adopted Core 02, §§10.2, 11, and 14.1. |
| Core 03 | Workbook hot-path and no pack-defined surface boundary | Current adopted Core 03, §2. |
| Core 04 | Authorization, roots, security, limits, conformance | Current adopted Core 04, §§2, 4.1, 9.4, and 12.3. |
| Reporting Subsystem NLSpec | Pack-sensitive reporting consumer boundary | Current adopted Reporting NLSpec after `RP-GATE-006`. |
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
Every JSON document and NDJSON line governed by this NLSpec MUST be UTF-8 without BOM. Invalid UTF-8, duplicate object members at any depth, invalid JSON syntax, unpaired surrogates, and a top-level value of the wrong JSON type MUST fail before semantic validation.

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
An archive member path MUST be ASCII, use `/` as the only separator, contain no empty segment, no `.` segment, no `..` segment, no NUL, and no leading `/` or `./`. The path MUST be at most 1024 bytes and each segment MUST be at most 255 bytes.

**RP-REQ-037**
Path comparison is exact ASCII byte comparison. The archive MUST additionally reject two paths that differ only by ASCII letter case. Directory member names MUST end with `/`; regular-file member names MUST not end with `/`.

**RP-REQ-038**
Every `extensions` object is closed at its immediate parent but open inside `extensions`. Extension keys MUST use reverse-domain form with at least three dot-separated segments and satisfy the `pack_key` segment grammar. Extension values may contain only the JSON value classes permitted by RP-REQ-031. Unknown extension keys MUST be preserved in canonical bytes and ignored for v1 behavior. An extension MUST NOT change identity, trust, dependency, compatibility, content semantics, or consumer results in v1.

# 7. Pack container and logical layout

**RP-REQ-039**
The accepted physical container formats are exactly:

- ZIP;
- uncompressed TAR;
- GZIP containing exactly one TAR stream.

The implementation MUST identify format from bytes. Media type and filename are advisory after Core upload-envelope admission.

**RP-REQ-040**
ZIP input MUST use the single-disk non-ZIP64 format and only compression methods `stored` or `deflate`. The first ZIP local-file header MUST begin at byte offset `0`, and the end-of-central-directory record, including any declared ZIP comment, MUST end at the final container byte. TAR input MUST use POSIX ustar headers without GNU, PAX, sparse, link, device, FIFO, or socket extensions; its byte length MUST be a multiple of 512, it MUST end with at least two consecutive 512-byte zero blocks, and bytes after the first terminating pair may contain only additional complete zero blocks. GZIP input MUST begin at byte offset `0`, contain exactly one CRC-valid GZIP member whose decompressed bytes are one conforming ustar archive, and end immediately after that member. The implementation MUST reject encrypted ZIP, prepended ZIP data, multi-disk ZIP, ZIP64, any other ZIP compression method, arbitrary non-TAR GZIP, concatenated or trailing-data GZIP, non-ustar TAR extensions, sparse members, and any other trailing bytes. Archive comments, timestamps, ownership, and permission bits are non-semantic and MUST NOT affect logical identity or execution behavior.

**RP-REQ-041**
An operator-imported container MUST have exactly this logical top-level layout:

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
`manifest.json` MUST be a closed canonical object conforming to `cartulary.reference_pack_manifest.v1`. Every top-level member in Table 8-A is required. An empty collection MUST be serialized as `[]`; unused extensions MUST be `{}`.

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

A `license_ref_bindings[]` item contains exactly `license_ref` and `notice_path`. `license_ref` MUST be one exact `LicenseRef-*` token present in `license.expression`; `DocumentRef-*` forms are forbidden in v1. `notice_path` MUST appear in both `notice_paths[]` and `manifest.files[]` with role `notice`. Bindings sort by `license_ref`, then `notice_path`; duplicates are invalid.

**RP-REQ-051**
Every `LicenseRef-*` token in `license.expression` MUST have exactly one `license_ref_bindings[]` item. A binding for a token absent from the expression is invalid. Every bound notice file MUST contain the corresponding license or notice text. An expression containing only SPDX License List identifiers and exceptions requires `license_ref_bindings=[]`. `notice_paths[]` MUST equal the complete set of `manifest.files[]` paths whose role is `notice`; if the set is empty, `notices/` MUST be absent.

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

Missing or mismatched binding data fails with `metadata_mix_and_match_detected`.

**RP-REQ-073**
Threshold satisfaction MUST count only distinct valid signatures from key IDs authorized for the applicable role in the currently trusted root. An untrusted signer, invalid signature, duplicate signer, or insufficient distinct threshold fails with `signature_threshold_not_met`.

**RP-REQ-074**
The successful verification attestation MUST retain every valid signer key ID used to satisfy each role threshold. For public summary compatibility, `signer_key_id` MUST equal the lexicographically least targets-role signer key ID. Before this NLSpec is adopted, Core 01 MUST add `verified_signer_key_ids[]` as required by RP-REQ-215. Until that Core amendment is adopted, the complete signer set remains attestation-only and the Reference Pack conformance claim remains blocked.

## 10.5 Expiry and historical use

**RP-REQ-075**
Import, reverify, refresh, and activation MUST evaluate trust freshness at their fixed operation start time. Passage of time beyond `trust_valid_until` MUST NOT by itself mutate an already published active set.

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
| `disabled` | `verified_available` | Successful reverify. No separate re-enable operation exists in v1. |
| `verified_available`, `disabled`, or active | `failed` | Later reverify or integrity check fails. |
| `verified_available`, `disabled`, or active | `missing` | Required retained payload becomes unavailable. |
| `failed` or `missing` | `staged` | New exact reimport of the same logical version. |
| `failed` or `missing` | `verified_available` | Successful reverify only when all retained bytes are present. |

Any unlisted transition is invalid.

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
Mutable deployment-local registry overrides are not supported in v1. Organization-specific registry behavior MUST be delivered as a complete immutable operator-imported pack version. Core 02's existing optional-local-override wording is a promotion blocker until amended.

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
The phase order in RP-REQ-089 is also the public failure-precedence order. The first fatal or error issue under deterministic issue ordering controls the terminal result. An issue whose `code` is a Core-recognized public verification reason maps to `error.code='reference_pack_verification_failed'` with `reason_code=code`. `job_canceled` maps to the common canceled-job outcome. `index_construction_failed` and `publication_failed` map to the common internal job-failure outcome and MUST NOT be represented as an invented public validation reason.

**RP-REQ-091**
After manifest identity is safely decoded and the `staged` identity is committed, the Core-owned `staged` metadata resource MUST be queryable while verification continues. A failure before safe identity attribution MUST create no pack-version resource. Before phase 21 commits, the implementation MUST NOT expose extracted content, derived indexes, verified provenance, or a candidate as usable pack data. Successful publication commits the verified condition and all success effects atomically. A terminal verification failure after staged identity allocation MUST atomically publish only the `failed` or `missing` condition, validation summary, failure attestation when identity is attributable, and job result; it MUST NOT publish indexes, active-pointer changes, pack sets, trust-root updates, or highest-trusted metadata-version updates. A crash, cancellation, timeout, or injected failure MUST leave no partially visible success effect, and failed verification MUST leave the pre-operation trust state unchanged.

**RP-REQ-092**
A verification issue MUST conform to closed `reference_pack_issue.v1`:

| Member | Type | Rule |
| --- | --- | --- |
| `issue_id` | string | Deterministic ID from RP-REQ-095. |
| `severity` | string | `fatal`, `error`, `warning`, or `info`. |
| `phase` | string | Exact phase token from Table 12-A. |
| `code` | string | One public verification reason recognized by Core 01, or `job_canceled`, `index_construction_failed`, or `publication_failed`. |
| `reason_code` | string or null | Equal to `code` when `code` is public; otherwise `null`. |
| `path` | string | Exact pack path or canonical JSON diagnostic path. |
| `entry_id` | string or null | Exact content entry ID when attributable. |
| `safe_details` | object | Closed object defined below. |

`safe_details` MUST contain exactly `requirement_id`, `limit_id`, `expected_token`, `actual_token`, `related_pack_key`, and `related_pack_version`. Each member is a string or `null`; values not applicable to the issue are `null`. Every non-null string is limited to 256 Unicode scalar values, MUST satisfy RP-REQ-034, and MUST contain no payload-derived value. `requirement_id`, when non-null, MUST be one exact `RP-REQ-*` identifier. `related_pack_key` and `related_pack_version` MUST satisfy RP-REQ-022 and RP-REQ-023. Count and byte-limit tokens MUST use unsigned base-10 with no leading zero except `0`. A canonical JSON diagnostic path starts at `$`, uses `.member` for an ASCII identifier member, `[<zero-based integer>]` for an array index, and `[` plus an RFC 8785 JSON string plus `]` for any other member name; `path` is limited to 4096 UTF-8 bytes.

Severity `fatal` or `error` makes verification fail. Severity `warning` or `info` does not make verification fail by itself.

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

**RP-REQ-093**
Issue ordering MUST be:

1. severity rank `fatal`, `error`, `warning`, `info`;
2. phase rank from Table 12-A;
3. `code` ascending exact bytes;
4. `path` ascending UTF-8 bytes;
5. `entry_id`, with `null` before strings;
6. `issue_id` ascending.

**RP-REQ-094**
A validation summary MUST conform to closed `cartulary.reference_pack_validation_summary.v1` and contain exactly `schema_id`, `result`, `issues_truncated`, `total_issue_count`, `retained_issue_count`, and `issues`. `schema_id` is exact; `result` is `succeeded` or `failed`; `issues` is ordered by RP-REQ-093. At most 1000 issues are retained. If more exist, the implementation MUST retain the first 1000 and set `issues_truncated=true`, `total_issue_count=<actual count>`, and `retained_issue_count=1000`. If no truncation occurs, `issues_truncated=false` and both counts equal the retained count.

**RP-REQ-095**
Before issue counts are calculated, byte-identical issue objects MUST coalesce. `issue_id` is unique only within one validation summary and MUST equal:

```text
"rpi_" + lowercase_hex(
  SHA256(
    RFC8785_JCS({severity, phase, code, path, entry_id, safe_details})
  )
)
```

**RP-REQ-096**
Safe issue details, job summaries, logs, telemetry, and administrative audit MUST NOT contain payload values, entry descriptions, raw signatures, private keys, source credentials, filesystem paths, object-store keys, or incident data.

**RP-REQ-097**
Cancellation MUST be checked at least between every phase in RP-REQ-089 and during bounded streaming loops. A cancellation before phase 21 publishes no new candidate result. If candidate identity was already committed as `staged`, terminal cancellation sets it to `failed` with internal code `job_canceled`.

**RP-REQ-098**
Verification timeout MUST use `limits.reference_packs.max_verification_seconds`. Timeout before publication produces `reference_pack_verification_failed` with `reason_code='verification_timeout'`, publishes no active-set or index change, and cleans the temporary root.

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
Aliases MUST be sorted by the profile's alias-normalization output, then exact alias bytes. Two aliases in one pack that normalize to the same lookup key but resolve to different entries are invalid.

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
A framework `external_refs[]` item contains exactly `source_name`, `external_id`, and nullable `url`. `source_name` and `external_id` are non-empty single-line strings. `url` is either `null` or the canonical output of RP-REQ-124 with scheme `https`, no userinfo, and no fragment. Items sort by `source_name`, then `external_id`, then `url` with `null` first. Duplicates are invalid.

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
`cartulary.indicator.normalize.ipv6.v1` and `validate.ipv6.v1` MUST reject zone IDs, brackets, ports, CIDR suffixes, dotted-quad suffixes, and IPv4-mapped or IPv4-compatible forms. Output MUST use RFC 5952 lowercase compression, longest zero run, leftmost tie, and no leading hextet zeroes.

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
2. require ASCII URI input and reject every raw space, control, or code point outside RFC 3986 URI syntax;
3. reject userinfo;
4. lowercase the scheme;
5. canonicalize the host through the applicable domain, IPv4, or IPv6 algorithm;
6. enclose a canonical IPv6 host in `[` and `]` in URL output;
7. remove port `80` for HTTP and `443` for HTTPS;
8. preserve any other decimal port in `1..65535` without leading zero;
9. replace an empty path with `/`;
10. remove RFC 3986 dot segments from the path;
11. uppercase hexadecimal digits in percent encodings;
12. decode percent-encoded unreserved ASCII characters, where the unreserved set is exactly `A-Z`, `a-z`, `0-9`, `-`, `.`, `_`, and `~`;
13. preserve query member order, duplicate query members, and fragment text;
14. preserve the presence of an explicitly supplied empty query delimiter `?` or empty fragment delimiter `#`;
15. apply the percent-encoding rule independently to path, query, and fragment;
16. emit `scheme://host[:port]/path[?query][#fragment]`.

A malformed percent encoding or forbidden component is invalid. Canonical URL output MUST contain at most 8192 ASCII bytes.

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
4. remove one trailing separator unless the value is a root token;
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
A framework pack MAY remove an object or relationship in a later pack version. It MUST NOT reuse the removed ID for a different subject or relationship meaning. Pinned historical pack sets preserve the old object.

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
| `cve_id` | string | `^CVE-[0-9]{4}-[0-9]{4,}$`. |
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

`entry_id` MUST equal `ascii_lower_v1(provider_name) + ":" + base10(event_id)`. `provider_event_id` lookup input is one closed object containing `provider_name`, `event_id`, and optional `event_version`; omitted `event_version` matches every version, while explicit `null` is invalid. Results sort by `event_version` with `null` first, then `entry_id`. `event_id` lookup returns every matching provider sorted by `ascii_lower_v1(provider_name)`, exact provider bytes, nullable version, then `entry_id`.

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

`reference_pack_windows_path_lookup_v1` applies RP-REQ-118 trimming and NFC, replaces `/` with `\`, collapses consecutive separators, applies `ascii_lower_v1`, and performs no environment expansion, filesystem access, path existence check, or dot-segment resolution. Candidate paths sort by that output, then exact bytes and reject duplicate normalized values. HijackLibs `library_name` uses ASCII case-insensitive basename lookup; `candidate_path` uses the named path algorithm. SID grammar is `S-1-<identifier-authority>-<subauthority>...`; numeric components are unsigned base-10 with no leading zero except `0`, identifier authority is `0..281474976710655`, each subauthority is `0..4294967295`, and at least one subauthority is required. SID lookup is exact; `name` and `alias` use ASCII case-insensitive comparison. Results sort by `entry_id`.

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
During verification, every declared dependency tuple MUST resolve to one retained, non-removed version with the exact `payload_sha256`, complete available canonical bytes, and durable condition `verified_available` or `disabled`, or to the exact active version of that tuple. An absent, digest-mismatched, failed, missing, or administratively removed dependency MUST fail verification with `dependency_unsatisfied`. The candidate plus every reachable retained dependency MUST form an acyclic graph; a cycle MUST fail with `dependency_cycle`. A candidate that conflicts with itself or with one of its declared dependency tuples MUST fail with `pack_conflict`. The presence of an unrelated active conflict is not a verification failure and is evaluated during activation.

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

**RP-REQ-159**
If an active dependency becomes disabled, failed, missing, removed, or replaced by another version, every transitive dependent whose exact tuple is no longer satisfied MUST lose active status in the same publication transaction. No dependent is automatically reactivated later.

**Table 18-A. Compatibility classification**

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
| Add an informational namespaced extension | Compatible. |
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

**RP-REQ-162**
The application MUST maintain one `algorithm_behavior_digest` for every application-owned indicator algorithm ID. The digest MUST be computed as lowercase SHA-256 of RFC 8785 canonical bytes of the harness-owned ordered conformance-vector array `{algorithm_id, value_kind, raw_value, valid, validation_code, normalized_value, defanged_value, dedupe_key}`. Vectors sort by `algorithm_id`, `value_kind`, then raw-value UTF-8 bytes. Every conforming implementation of one application contract major MUST produce the same digest from the same canonical vector set. A changed behavior digest under the same algorithm ID is `type_registry_incompatible`; source-code or executable-file hashes MUST NOT be used as cross-implementation behavior digests.

# 19. Immutable reference pack set

**RP-REQ-163**
Every Base Profile runtime MUST maintain one current immutable `cartulary.reference_pack_set.v1`, even when the Reference Pack Extension Profile is not claimed. The set MUST contain exactly one effective member for each required Base registry key.

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

**RP-REQ-169**
Activation, rollback, safety fallback, startup profile reconciliation, disablement of an active version, failed or missing active-version detection, dependency invalidation, and removal-related pointer changes MUST publish exactly one replacement pack set in the same atomic transaction as active-pointer and attestation effects. When the canonical member array is unchanged, the operation MUST reuse the existing `pack_set_id` and MUST NOT create a byte-distinct duplicate set object.

**RP-REQ-170**
A reader MUST observe either the complete prior set or the complete replacement set. It MUST NOT observe a mixture of versions from two sets.

**RP-REQ-171**
Every pack-dependent operation MUST capture one `pack_set_id` at admission and MUST use that set for its complete execution. A later active-set change MUST NOT change lookups, validation, report derivation, snapshot derivation, or index selection for the admitted operation.

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
All mutating semantic operations MUST serialize conflicting changes by affected `pack_key` in ascending exact order. Multi-key operations MUST acquire or otherwise enforce their mutation boundary in ascending `pack_key` order to prevent nondeterministic deadlock resolution.

**RP-REQ-176**
The operation start time, current trusted roots, highest metadata versions, selected key set, current pack set, and relevant retained bytes MUST be fixed once at operation admission. An operation MUST NOT silently switch to later trust state, a later active set, or a changed selector during execution.

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

## 20.4 Reverify

**RP-REQ-183**
`ReverifyPack` is valid only for an operator-imported version in `verified_available`, `disabled`, `failed`, or `missing`. It is invalid for `staged` or `packaged_builtin`.

**RP-REQ-184**
Reverify MUST use the exact current successful verification envelope, including its retained operator container, TUF metadata, manifest, payload, and notice bytes. It MUST NOT synthesize missing metadata from canonical payload bytes, import new bytes, scan the incoming directory, infer a newer version, or perform network access.

**RP-REQ-185**
If any required retained container, metadata, manifest, payload, or notice byte is absent, reverify fails with `payload_missing` and MUST NOT attempt partial reconstruction. If every required byte is present, reverify runs the complete verification pipeline under current trust and runtime state.

**RP-REQ-186**
Successful reverify changes `disabled`, `failed`, or `missing` to `verified_available`. If the version was active and remains valid, it remains active. Failed reverify of an active version invokes `InvalidateUnavailablePack` and applies optional invalidation or required-registry safety fallback.

## 20.5 Refresh

**RP-REQ-187**
At Core route admission, omitted `pack_keys[]` MUST resolve once to every `pack_key` that has at least one retained operator-imported version not tombstoned by administrative removal. Packaged-built-in-only keys and keys whose only imported versions are administrative-removal tombstones do not enter the omitted set. Explicit selected keys use Core's canonical exact-token set semantics.

**RP-REQ-188**
`RefreshPacks` MUST process every retained operator-imported version of each selected key except a version with `missing_reason='administrative_removal'`, sorted by `pack_key`, then exact `pack_version`. Administrative-removal tombstones are skipped and unchanged. It MUST NOT:

- scan `incoming/`;
- discover a new container;
- import new bytes;
- download content;
- infer a latest version;
- auto-activate a different version.

**RP-REQ-189**
Refresh verification and index construction MUST occur in temporary state. All selected version outcomes, active-pointer removals, required-registry fallbacks, dependency invalidations, attestations, compatible trust-root updates, trusted metadata versions, and the replacement pack set MUST publish in one transaction. Root-update chains from different selected containers MUST start from the admission root, MUST have byte-identical metadata for every shared root version, and MUST form one prefix-compatible chain; otherwise refresh fails with `tuf_root_rotation_invalid` and publishes no refresh outcomes.

**RP-REQ-190**
For each selected version, successful refresh preserves `verified_available`, preserves `disabled`, preserves an active valid version, and changes `failed` or `missing` to `verified_available` only when every retained byte is present and verification succeeds. A failed refresh changes the addressed version to `failed` or `missing` under RP-REQ-202 and applies active invalidation when required. If any selected version fails, the job terminal result is failed. The first failure under pack ordering and validation issue ordering controls the public summary. All selected version outcomes MUST nevertheless be published together at the one publication boundary. Cancellation or timeout before that boundary publishes none of the refresh outcomes.

**RP-REQ-191**
A refresh selection containing zero retained operator-imported versions is a successful deterministic no-op job. It produces no version mutation and no new pack set.

## 20.6 Activation and rollback

**RP-REQ-192**
`ActivatePack` is legal only when the addressed version is `verified_available`, all trust metadata required for fresh activation is unexpired, dependencies are active, conflicts are absent, and compatibility passes.

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
`DisablePack` is valid only for an operator-imported version in `verified_available`, whether active or inactive. An already disabled version fails through Core's state-conflict contract. A packaged built-in is `not_disableable`.

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

**RP-REQ-201**
An administratively removed version may return only through exact reimport of the same logical bytes. Reverify alone cannot recreate absent bytes. Removal history remains append-only after reimport.

## 20.9 Missing-payload invalidation

**RP-REQ-202**
When a consumer or integrity scanner detects unavailable retained state, `InvalidateUnavailablePack` MUST serialize with ordinary pack mutations and use this closed classification: absent or unreadable authoritative bytes become `missing` with `missing_reason='storage_loss'`; absent staged bytes become `missing` with `missing_reason='staging_loss'`; present bytes whose digest, schema, semantics, or runtime compatibility fails become `failed` with `missing_reason=null`. The operation MUST preserve the first detection attestation and atomically apply active-pointer, dependency, safety-fallback, and pack-set effects before another dependent operation succeeds.

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
Every operation MUST first resolve the exact retained pack set. `pack_set_id` MUST match `^rpset_[0-9a-f]{64}$`; malformed input fails with `invalid_pack_request`. Unknown or garbage-collected `pack_set_id` fails with `pack_set_not_found`. A pinned historical set may be read from its exact previously verified bytes even when the current version condition is `failed` solely because a later freshness reverify reported `metadata_expired`; missing bytes, digest failure, or content incompatibility still fails with `pack_unavailable`. A set member whose required bytes or index cannot be read triggers RP-REQ-202 when current state has not already recorded the failure.

**RP-REQ-206**
`GetPackEntry` uses exact `entry_id` or `object_id` comparison. The requested identity MUST be a non-empty string of at most 192 UTF-8 bytes. A pack key not present in the set returns `pack_unavailable`. A present pack with no matching entry returns `entry_not_found`. Success returns one closed object containing exactly `item` and `provenance`; `item` is the canonical decoded entry or object and `provenance` satisfies RP-REQ-210.

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
| `last_verified_at` | non-null | Successful verification timestamp. |
| `trust_valid_until` | nullable | `null` for packaged built-ins. |
| `missing_reason` | nullable | `administrative_removal`, `storage_loss`, `staging_loss`, or `null`. |
| `verified_signer_key_ids` | non-null array | Empty for packaged built-ins; sorted valid signer IDs for imported packs. |

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

The resulting registry is closed. No other `reference_pack_verification_failed` reason is valid in contract major 1. Table 22-A is exhaustive. Within one validation phase, the first matching row in Table 22-A controls; across phases, RP-REQ-090 controls.

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
| `target_length_mismatch` | `target_inventory` | A TUF target descriptor length differs from the exact target byte length. |
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
| `payload_missing` | `payload_integrity` | Reverify, refresh, restore-time pack verification, or integrity validation requires retained authoritative bytes that are absent or unreadable; initial import uses `required_member_missing`. |
| `verification_timeout` | current performed phase | Elapsed verification time exceeds `limits.reference_packs.max_verification_seconds` before publication. |
| `archive_extracted_bytes_exceeded` | `archive_preflight` or `archive_structure` | Extracted bytes exceed the effective `max_extracted_bytes`. |
| `archive_compression_ratio_exceeded` | `archive_preflight` | Compression ratio exceeds the effective `max_compression_ratio`. |
| `archive_member_count_exceeded` | `archive_preflight` | Regular-file member count exceeds the effective `max_members`. |

**RP-REQ-217**
Core 01 MUST add activation-rejection reasons:

```text
already_active
not_verified_available
trust_expired
contract_incompatible
dependency_not_active
active_conflict
type_registry_incompatible
required_registry_gap
```

Table 22-B is exhaustive. When multiple conditions apply, the first matching row controls.

**Table 22-B. Activation rejection mapping**

| Precedence | Reason | Exact trigger |
| ---: | --- | --- |
| 1 | `already_active` | The addressed exact version is already the effective active version for its key. |
| 2 | `not_verified_available` | The target is not a packaged built-in with a valid release binding and is not in durable condition `verified_available`. |
| 3 | `trust_expired` | The target is operator-imported and its required metadata is not fresh at the fixed activation start time. |
| 4 | `contract_incompatible` | The target or proposed set violates pack contract, content-profile, runtime-capability, or pack-set-size compatibility. |
| 5 | `dependency_not_active` | A declared exact dependency is not active in the proposed set. |
| 6 | `active_conflict` | The proposed set contains a conflict declared by the target or another active member. |
| 7 | `type_registry_incompatible` | Type-registry compatibility fails under §18. |
| 8 | `required_registry_gap` | The proposed set cannot contain exactly one effective version for each required Base registry key after required fallback processing. |

Only reasons applicable to the activation operation may be returned.

**RP-REQ-218**
Core 01 MUST add `error.code='reference_pack_removal_rejected'` with exactly:

```text
packaged_builtin
already_removed
active_version
reproducibility_pinned
verification_pending
```

Table 22-C is exhaustive. When multiple conditions apply, the first matching row controls.

**Table 22-C. Removal rejection mapping**

| Precedence | Reason | Exact trigger |
| ---: | --- | --- |
| 1 | `packaged_builtin` | The addressed version has `distribution_kind='packaged_builtin'`. |
| 2 | `already_removed` | The addressed version is `missing` with `missing_reason='administrative_removal'`. |
| 3 | `active_version` | The addressed version is the effective active version for its key. |
| 4 | `reproducibility_pinned` | A retained snapshot, release, report, pack set, or other reproducibility object pins the version or a set containing it. |
| 5 | `verification_pending` | Import, reverify, refresh, activation, disablement, removal, or integrity publication work for the version is admitted and nonterminal. |

**RP-REQ-219**
Internal validation codes MUST map deterministically to one public reason. More specific internal detail may remain in the retained issue summary, but the public error MUST NOT expose raw payload content, paths, signatures, or keys.

**RP-REQ-220**
Until `RP-GATE-003` is adopted, implementations may expose experimental removal or expanded verification diagnostics only outside a conformance claim. Public validation MUST report the missing Core dependency rather than silently using a private public contract.

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
`reference_pack_refs.json` MUST identify every referenced member with exactly:

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
An embedded pack in an incident bundle MUST receive no trust from the incident-bundle signature itself, remain inactive after import, import no source-deployment activation or attestation state, reuse an existing exact logical version when digests match, and reject a same-key/version digest collision. An embedded `operator_imported` pack MUST pass the ordinary TUF pipeline using target deployment trust roots. An embedded `packaged_builtin` pack may be reused only when the target deployment already retains the exact logical version or has a trusted application-release binding for the exact digests; otherwise it is unavailable and MUST NOT be converted into an operator-imported or target-trusted built-in implicitly.

**RP-REQ-227**
An embedded optional pack failure MUST NOT block core incident import unless the incident portability manifest explicitly classifies that exact pack set member as required for the imported artifact. Required-member behavior remains owned by the portability contract.

**RP-REQ-228**
Only `redistribution='allowed'` pack payloads may be embedded. Restricted or prohibited packs must be represented by exact references only.

## 23.3 Operational backup and restore

**RP-REQ-229**
When the Reference Pack Extension Profile is claimed, operational backup MUST include:

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

# 24. Persistence minima and invariants

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
| Trust-bootstrap file bytes | `8388608` |
| Manifest-declared files | `66` |
| `bundle.json` bytes | `16384` |
| `manifest.json` bytes | `1048576` |
| One TUF metadata file | `2097152` |
| Total TUF metadata bytes | `8388608` |
| One payload or notice member | `268435456` |
| Archive path bytes | `1024` |
| Path-segment bytes | `255` |
| Source artifacts | `64` |
| Dependencies | `64` |
| Conflicts | `64` |
| TUF keys in one root | `64` |
| Signatures in one metadata file | `64` |
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
| `max_trust_bootstrap_bytes` | Table 26-B | Invalid deployment configuration before workers start. |
| `max_manifest_declared_files` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_bundle_json_bytes` | Table 26-B | `reference_pack_verification_failed/bundle_hint_invalid` |
| `max_manifest_json_bytes` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_tuf_metadata_file_bytes` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_tuf_metadata_total_bytes` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_payload_member_bytes` | Table 26-B | `reference_pack_verification_failed/archive_structure_invalid` |
| `max_archive_path_bytes` | Table 26-B | `reference_pack_verification_failed/archive_structure_invalid` |
| `max_path_segment_bytes` | Table 26-B | `reference_pack_verification_failed/archive_structure_invalid` |
| `max_source_artifacts` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_dependencies` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_conflicts` | Table 26-B | `reference_pack_verification_failed/manifest_schema_invalid` |
| `max_tuf_keys` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_tuf_signatures` | Table 26-B | `reference_pack_verification_failed/tuf_metadata_invalid` |
| `max_entries` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_framework_objects` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_framework_relationships` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_aliases` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_source_refs` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_profile_array_items` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_pack_set_members` | Table 26-B | `reference_pack_activation_rejected/contract_incompatible` |
| `max_ndjson_line_bytes` | Table 26-B | `reference_pack_verification_failed/content_schema_invalid` |
| `max_retained_validation_issues` | Table 26-B | Truncate under RP-REQ-094; no failure. |
| `max_verification_seconds` | RP-REQ-245 | `reference_pack_verification_failed/verification_timeout` |

**RP-REQ-245**
Core 04 MUST add:

```text
limits.reference_packs.max_verification_seconds
```

The default is `1800`; valid values are integer seconds in `60..86400`. Omitted configuration resolves to `1800`.

**RP-REQ-246**
A limit MUST be checked at the earliest phase where it can be determined safely. A limit breach MUST stop further untrusted parsing that is not required to identify the failure, clean temporary state, and produce no candidate publication.

**RP-REQ-247**
Every limit breach MUST use the exact outcome in Table 26-C. When the outcome is a verification issue, `code` and `reason_code` equal the reason after `/`; `safe_details.limit_id` equals the Table 26-C token; and `expected_token` and `actual_token` contain only the unsigned base-10 maximum and observed count, byte length, ratio numerator, ratio denominator, or elapsed seconds required to explain the breach. A limit MUST NOT be remapped according to implementation phase or parser choice.

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
| `container_sha256` | Non-null only when one operator container was evaluated. |
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

It MUST display content profile, source version, source-as-of date, logical digests, last verification, trust-valid-until when applicable, previous active version, missing reason, dependencies, and reproducibility-pin removal blockers. These UI fields are projections over owner state and do not authorize actions.

# 28. Core companion amendments and adoption status

**RP-REQ-255**
The companion amendments in Table 28-A are normative adoption dependencies. This NLSpec MUST remain draft while any row is open.

**Table 28-A. Required owner amendments**

| Owner artifact | Required amendment |
| --- | --- |
| Core 00 | Adopt this NLSpec for its bounded Base and Reference Pack scopes; update the owner matrix and claim dependency. |
| Core 01 §11 | Distinguish packaged Base registries, imported replacement behavior, safety fallback, no separate re-enable action, exact refresh behavior, and immutable pack sets. |
| Core 01 §12 | Bind snapshots, releases, portability, backup, and restore to exact pack sets and logical digests. |
| Core 01 §17.4 | Add removal, resource fields, expanded error reasons, and this NLSpec's semantic bindings. |
| Core 02 §11 | Remove mutable local overrides; import content-profile schemas and stable algorithm IDs. |
| Core 02 §14.1 | Adopt logical persistence minima in §24. |
| Core 04 §4.1 | Adopt mandatory signed trust, no-egress, and hostile-content rules. |
| Core 04 §9.4 | Map Reference Pack conformance to §31. |
| Core 04 §12.3 | Add trust-bootstrap path, `max_container_bytes`, and verification timeout; retain the imported archive-limit domains in §26. |
| Reporting Subsystem NLSpec | Consume exact snapshot pack set and fail on missing pinned packs. |
| OpenTelemetry NLSpec | Adopt operation and attribute registry. |
| Testing Harness NLSpec | Adopt schema generation, fixture execution, and drift/accounting obligations. |
| `domain.md` | Add vocabulary from Table 5-A. |

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
| `spec_traceability` | Every contiguous requirement range in Table 31-B maps to a live criterion and fixture family where required. |
| `telemetry_privacy` | No forbidden data leakage. |

**RP-REQ-258**
Every Table 13-A content profile MUST have at least one canonical valid fixture, one malformed-structure fixture, and one semantic-error fixture. The three packaged built-in registry fixtures MUST be byte-identical release inputs and MUST publish their exact digests.

**RP-REQ-259**
The harness MUST verify the exhaustive range mapping in Table 31-B against the live requirement and criterion registries and MUST report `unmapped=0`, `unknown_requirement=0`, and `unknown_acceptance_criterion=0`. Research reports, implementation guides, screenshots, or manual demonstrations MUST NOT substitute for canonical fixture evidence.

**RP-REQ-260**
Reference Pack performance measurements are engineering information unless Core 05 claim-publication requirements are separately satisfied. This NLSpec creates no timed public claim.

# 30. Assumptions, blockers, and future-only areas

**RP-REQ-261**
The current v1 trust profile assumes Ed25519 is permitted. If a FIPS-validated cryptographic module requirement applies, `RP-GATE-010` fails and this document MUST be revised with a new verification-method ID and exact cryptographic profile. An implementation MUST NOT silently substitute ECDSA or RSA under `tuf_1_0_35_offline_bundle_v1`.

**RP-REQ-262**
The deployment clock is assumed sufficiently trustworthy for metadata-expiry validation. If that assumption is false, activation, import, reverify, and refresh MUST fail closed until a later trusted-time contract is adopted.

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
| `RP-AC-041` | Embedded incident-bundle packs are independently reverified, do not import source activation or attestation state, and remain inactive. |
| `RP-AC-042` | Restricted and prohibited packs are references-only in incident portability. |
| `RP-AC-043` | Restore verifies every retained artifact and cannot enter ready state with a missing, invalid, or historically unverifiable active or pinned pack; current-time metadata expiry is not retroactively applied to a valid retained attestation. |
| `RP-AC-044` | Import, verify, activate, refresh, remove, and lookup perform no outbound network access. |
| `RP-AC-045` | Cancellation, timeout, verification failure, and injected failure leave no partial verified content, active set, visible index, or trust-state update; only closed staged or terminal metadata may be published. |
| `RP-AC-046` | Every fixed and configurable limit has exact at-limit success and one-over failure evidence. |
| `RP-AC-047` | Every distributed pack carries valid license, source, builder, immutable source-profile ID and digest, transformation, and notice provenance. |
| `RP-AC-048` | Logs, telemetry, job summaries, issues, and audit contain no payload values, private keys, raw signatures, secrets, uncontrolled paths, or incident data. |
| `RP-AC-049` | Every `RP-REQ-*` maps to at least one acceptance criterion or canonical fixture with `unmapped=0`. |
| `RP-AC-050` | No current normative behavior is defined only in a guide, appendix, generated artifact, research report, or implementation-local convention. |

**Table 31-B. Exhaustive requirement-to-acceptance mapping**

Every integer requirement ID in each inclusive range maps to every listed acceptance criterion. No range overlaps another range, and the ranges cover `RP-REQ-001` through `RP-REQ-264` without omission.

| Requirement range | Acceptance criteria |
| --- | --- |
| `RP-REQ-001..005` | `RP-AC-001`, `RP-AC-050` |
| `RP-REQ-006..013` | `RP-AC-001`, `RP-AC-049`, `RP-AC-050` |
| `RP-REQ-014..016` | `RP-AC-037`, `RP-AC-038`, `RP-AC-044` |
| `RP-REQ-017..021` | `RP-AC-002`, `RP-AC-049` |
| `RP-REQ-022..028` | `RP-AC-015`, `RP-AC-016` |
| `RP-REQ-029..038` | `RP-AC-009` |
| `RP-REQ-039..047` | `RP-AC-006`, `RP-AC-007`, `RP-AC-008`, `RP-AC-011`, `RP-AC-046` |
| `RP-REQ-048..052` | `RP-AC-009`, `RP-AC-010`, `RP-AC-011`, `RP-AC-047` |
| `RP-REQ-053..057` | `RP-AC-006`, `RP-AC-010` |
| `RP-REQ-058..076` | `RP-AC-012`, `RP-AC-013`, `RP-AC-014` |
| `RP-REQ-077..088` | `RP-AC-003`, `RP-AC-004`, `RP-AC-005`, `RP-AC-015`, `RP-AC-016`, `RP-AC-024` |
| `RP-REQ-089..098` | `RP-AC-007`, `RP-AC-008`, `RP-AC-009`, `RP-AC-010`, `RP-AC-011`, `RP-AC-012`, `RP-AC-013`, `RP-AC-045`, `RP-AC-046`, `RP-AC-048` |
| `RP-REQ-099..111` | `RP-AC-017`, `RP-AC-018` |
| `RP-REQ-112..131` | `RP-AC-019`, `RP-AC-020`, `RP-AC-021`, `RP-AC-036` |
| `RP-REQ-132..137` | `RP-AC-017`, `RP-AC-037` |
| `RP-REQ-138..148` | `RP-AC-017`, `RP-AC-037` |
| `RP-REQ-149..155` | `RP-AC-006`, `RP-AC-047` |
| `RP-REQ-156..162` | `RP-AC-020`, `RP-AC-021`, `RP-AC-022` |
| `RP-REQ-163..173` | `RP-AC-023`, `RP-AC-024`, `RP-AC-025`, `RP-AC-026`, `RP-AC-027`, `RP-AC-039`, `RP-AC-040` |
| `RP-REQ-174..203` | `RP-AC-015`, `RP-AC-016`, `RP-AC-023`, `RP-AC-024`, `RP-AC-027`, `RP-AC-028`, `RP-AC-029`, `RP-AC-030`, `RP-AC-031`, `RP-AC-032`, `RP-AC-033`, `RP-AC-044`, `RP-AC-045` |
| `RP-REQ-204..213` | `RP-AC-034`, `RP-AC-035`, `RP-AC-036`, `RP-AC-037`, `RP-AC-038` |
| `RP-REQ-214..220` | `RP-AC-001`, `RP-AC-007`, `RP-AC-008`, `RP-AC-009`, `RP-AC-010`, `RP-AC-011`, `RP-AC-012`, `RP-AC-013`, `RP-AC-031` |
| `RP-REQ-221..231` | `RP-AC-039`, `RP-AC-040`, `RP-AC-041`, `RP-AC-042`, `RP-AC-043` |
| `RP-REQ-232..235` | `RP-AC-015`, `RP-AC-023`, `RP-AC-032`, `RP-AC-043` |
| `RP-REQ-236..242` | `RP-AC-018`, `RP-AC-044`, `RP-AC-048` |
| `RP-REQ-243..248` | `RP-AC-046` |
| `RP-REQ-249..254` | `RP-AC-048` |
| `RP-REQ-255` | `RP-AC-001`, `RP-AC-050` |
| `RP-REQ-256..260` | `RP-AC-049` |
| `RP-REQ-261..264` | `RP-AC-002`, `RP-AC-050` |

# 32. Definition of done

This NLSpec is complete for adoption only when all of the following are true:

- every promotion gate passes;
- every schema ID and algorithm ID resolves exactly once;
- every object member has type, requiredness, nullability, default, and unknown-member behavior;
- every collection has bounds, duplicate behavior, and ordering;
- every operation has deterministic admission, cancellation, timeout, publication, and error behavior;
- every current pack key has a valid fixture, malformed fixture, and semantic-error fixture;
- every limit has boundary fixtures;
- every public reason mapping is adopted by Core 01;
- every Core companion amendment in Table 28-A is adopted;
- the recreatability test passes;
- two independent implementations produce interchangeable observable results for the same canonical inputs;
- no normative `TODO:`, open delegation, or contradictory owner statement remains.

## Sources

[^1]: `R05-responsive-interface-design-report.cr.md`, thesis and staged-interface findings, lines 5-19 and 133-147. Source limit: research supports nonblocking semantic progress presentation but does not define Reference Pack runtime contracts.

[^2]: JSON Schema, `Draft 2020-12`, published 16 June 2022, https://json-schema.org/draft/2020-12. Accessed 16 July 2026.

[^3]: RFC 8785, *JSON Canonicalization Scheme (JCS)*, especially canonical representation, deterministic property sorting, UTF-8 generation, and invalid-Unicode failure, https://www.rfc-editor.org/rfc/rfc8785.html.

[^4]: The Update Framework Specification 1.0.35, last modified 15 July 2026, especially §§1.5, 2.1, 2.3, 4, and 5, https://theupdateframework.github.io/specification/latest/.

[^5]: SPDX Specification 3.0.1, https://spdx.github.io/spdx-spec/v3.0.1/; SPDX License List 3.28.0, dated 20 February 2026, https://spdx.org/licenses. Accessed 16 July 2026.

[^6]: `R03-Kanvas_technical_research_report.md`, locally downloaded reference-data families and missing signature/checksum verification, lines 911-915 and 1036-1040.

[^7]: `00_document_set_status_and_precedence.md`, §§4.2 and 5.1, lines 77-109 and 146-193; `01_architecture_storage_and_view_contracts.md`, §11 and §17.4, lines 5790-6012 and 7347-7497; `02_domain_model_schema_and_history.md`, §11 and §14.1, lines 1484-1499 and 1987-2026; `03_workbook_interaction_collaboration_and_workflows.md`, §2; `04_security_deployment_and_conformance.md`, §4.1, §9.4, and §12.3.1, lines 470-490, 1130-1150, and 2072-2135. Source limit: these files define the current outer profile and owner boundaries; the companion amendments required by this draft have not been applied by this artifact.

[^8]: `nlspec-spec.md`, “Behavioral Completeness,” “Unambiguous Interfaces,” “Explicit Defaults and Boundaries,” “Mapping Tables for Translation,” “Testable Acceptance Criteria,” and “Spec Economy,” lines 11-122 and 142-216.

[^9]: `R06-spreadsheet_of_doom_dfir_research_report.md`, reference-vocabulary separation and optional nonblocking enrichment, lines 454-463 and 490-500. Source limit: research evidence supports the product boundary but does not define runtime pack semantics.

[^10]: `R01-aurora_incident_response_report.md`, direct renderer-side external integrations, remote runtime dependency, global invalid-certificate acceptance, and the resulting expanded trust boundary, lines 94-98 and 1243-1274. Source limit: the report supports the no-egress and strict trust-boundary rationale; it does not define this subsystem's protocol.
