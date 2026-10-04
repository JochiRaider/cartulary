# Reference Pack limit evidence

This is a human review index for candidate NLSpec 0.2.0, profile 2, and its typed `contracts/reference-packs/limits.v1.json` projection. It is not consumed by tests or generators. Execution results and superseding failures are in the [implementation ledger](reference-pack-remediation.md). The [obligation index](reference-pack-obligation-evidence.md) gives public Make routes. A component fixture establishes that component's admission boundary; it does not establish destination trust, persistence or adoption by itself.

`format/` below means `internal/modules/reference_data/internal/packformat/`. “Complete” means all required data at the named boundary is present. A complete content row is not represented as a separately signed container. Cross-language identity/signature expectations remain the independently authored checked-in vectors; the compact boundary recipes do not replace them.

## Configured boundaries

| Value | Equality and excess evidence | Scope |
| --- | --- | --- |
| Container bytes, default 536870912 | `verification_engine_test.go`, `TestVerifierRejectsArchiveLimits_Unit`, complete equality subtests | An independently signed complete container succeeds with the effective bound equal to its exact length; decreasing that bound by one rejects the same container. |
| Extracted bytes, default 536870912 | Same complete live verifier subtests | Exact sum of all regular members; a one-byte smaller allowance fails before publication. |
| Regular members, default 10000 | Same complete live verifier subtests; archive admission tests | Exact regular-member inventory and one fewer permitted member; directory markers have an independent allowance. |
| Compression ratio, default 100, configured range 1–1000 | `format/archive_ratio_test.go`, reached from manifest unit row; archive overflow/regression tests | A complete manifest, profile payloads and notices reach exactly 100 extracted bytes per complete ZIP byte. Increasing the extracted numerator by one fails. An admitted ZIP comment fixes only the denominator; it does not enter logical digests. Overflow and TAR multi-member denominator guards are separately tested. |
| Verification seconds, default 1800, range 60–86400 | `internal/app/configassembly/configuration_test.go`; `format/portable_context_test.go`; coordinator/finalizer deadline tests | Both endpoints accepted; 59, 86401, null and fractional input rejected. Equality with the elapsed deadline is timeout unless commit is proven. |
| Lookup limit, default 50, range 1–200 | `consumer_test.go`, cursor test | Complete first-page results at 1 and 200, default 50, 0/201 rejection, continuation omission inheritance, explicit mismatch and null rejection. |
| Cursor lifetime, 900 seconds | Same cursor test | Valid at 900 seconds minus one nanosecond; invalid at equality. |
| Clock assertion, default false | Configuration assembly, coordinator and startup tests | Boolean rather than a numeric limit. False rejects fresh trust operations; historical reads, Base and empty refresh remain usable. |

## Fixed boundaries

| Value | Maximum | Evidence and reachability |
| --- | ---: | --- |
| JSON object/array nesting | 10000 | `format/tuf_limit_test.go`: complete authenticated metadata with a permitted signed extension reaches exactly 10000 open containers including the envelope and signed object. Opening the 10001st container fails original JSON admission before signature processing. |
| Trust bootstrap bytes | 8388608 | `format/tuf_limit_test.go`: complete canonical four-repository document at equality and one byte over, with fresh valid test-only signatures and individually bounded roots. |
| Declared manifest files | 66 | `format/manifest_limit_test.go`: two framework payloads plus 64 notices, valid content and exact archive inventory; 67-item schema rejection. |
| Bundle hint bytes | 16384 | Unreachable for the closed canonical two-field object; isolated `admittedHintSize` equality/one-over guard in the prepared-contract test. Argument below. |
| Manifest bytes | 1048576 | Unreachable under the other current canonical field bounds; manifest prerequisite tests exercise the independent size gate. Argument below. |
| One metadata file | 2097152 | `format/tuf_limit_test.go`: authenticated canonical metadata at equality, one-byte excess rejected before canonical/signature work. |
| Aggregate metadata bytes | 8388608 | Same test: four complete authenticated files at equality; individually valid bounded files total exactly 8388609 for aggregate rejection. |
| Successful envelope bytes | 67108864 | Isolated `admittedSuccessfulEnvelopeSize` 67108864/67108865 guard; complete retained-envelope tests establish shape and historical trust separately. Conservative reachability argument below. |
| Prepared result bytes | 71303168 | `format/prepared_test.go`, isolated equality/one-over; complete prepared objects execute in persistence and portability integration. Bound below is larger than any valid nested object. |
| Portable references bytes | 67108864 | `format/portability_retention_test.go`: complete admitted empty references object plus permitted unsigned whitespace at equality; one extra byte rejected. |
| Portable sets / versions | 1024 / 16384 | Same test: complete independently specified catalog with sixteen keys per set, no version reuse; isolated one-over collection guards. |
| One payload or notice | 268435456 | `format/archive_limit_test.go`: streamed complete manifest/registry/license/notice tree with independently known notice digest at equality; oversized archive rejected before destination writes. |
| Original archive path / segment | 1024 / 255 | `format/manifest_limit_test.go`: a declared notice's complete 1024-byte path contains three 255-byte segments; exact archive/inventory accepted; 1025/256 rejected. |
| Source artifacts | 64 | Same complete manifest fixture; 65 rejected by schema admission. |
| Dependencies / conflicts | 64 / 64 | Same complete candidate manifest has 64 exact version dependencies and 64 nonoverlapping conflicts. Pure dependency verification receives explicit authenticated-retention inputs; each 65-item manifest fails. This does not claim sixty-four independent database imports. |
| Root keys / ordinary signatures | 64 / 64 | `format/tuf_limit_test.go`: complete authorized 64-key/signature metadata; 65-key or 65-signature shape rejection. |
| Root-update signatures | 128 | Independent `rotation_64` vector verifies two disjoint 64-key roots; the byte-boundary test rejects a 129-signature envelope. |
| Entry / object counts | 2000000 / 2000000 | Unreachable under the per-file ceiling; isolated equality/one-over counters and executable schema lower-bound proof in `format/manifest_test.go`. |
| Relationship count | 5000000 | Same proof and isolated 5000000/5000001 guard. |
| Aliases / source refs | 64 / 64 | Complete longest-SID profile fixture, independent 65-item rejection manifests, and actual Get/lookup integration. The event and CVE boundary fixtures also carry 64 source refs. |
| Profile-specific array | 64 | `format/content_limit_test.go`: complete LOLBAS row has 64 usage examples and 64 HTTPS references; each 65-item array is tested separately with a small row. Other declared array shapes remain covered by their own profile admission tests. |
| Set members | 64 | Closed catalog admits only sixteen unique keys: complete independent sixteen-member vector, isolated 64/65 guard. |
| NDJSON line including LF | 1048576 | `format/content_limit_test.go`: complete canonical, semantically valid, indexed row at equality; one extra byte fails the bounded reader. |
| Retained issues | 1000 | `format/diagnostics_test.go`: complete check summaries for 1000 and 1001 unique findings; exact truncation boundary, plus 40002 deduplicated findings and traversal permutations. |
| Validation summary bytes | 16777216 | Isolated equality/one-over guard and closed complete summary tests. Valid 1000-issue output cannot reach this ceiling; argument below. |
| Generic Get entry ID bytes | 512 | Consumer request precedence test accepts 512 bytes through scalar admission and rejects 513 before set resolution. No current profile produces a valid 512-byte identity; longest current profile round trips are separate. |
| Windows event ID bytes | 266 | Independent longest-provider/version fixture through production Get and lookup, including a colon in the provider. Numeric version 255 and omitted-version lookup select the declared rows. |
| CVE ID bytes | 192 | Independent complete profile and production Get/lookup boundary fixture. |
| SID bytes / subauthorities | 184 / 15 | Independent maximum authority and fifteen maximum subauthorities; complete profile and production Get/lookup; pure identity tests reject the excess subauthority and invalid numeric forms. |
| Empty directory markers | 10000 | Complete TAR framing at 10000/10001 in archive tests; markers are semantically empty and excluded from regular-file count and digest inventory. This component test does not claim a signed 10000-directory container. |
| TAR terminating zero blocks | 20 | Complete TAR stream framing at 20/21; independent of extracted regular bytes. |
| Administrative request bytes | 65536 | `api_test.go`: complete request with admitted whitespace at equality and one-byte excess; upload-envelope owner independently bounds metadata before staging. |
| Portable content manifest bytes | 16777216 | `format/portable_content_test.go`: complete empty member plus admitted unsigned whitespace at equality, one-over rejection. |
| Portable containers | 16384 | Complete 1024-set × sixteen-profile catalog and content declarations; isolated 16385-item guard. |
| Portable required pairs | 65536 | Current catalog maximum is 16384 distinct pairs; complete reachable catalog plus isolated 65536/65537 guard. |
| Portable verification context bytes | 4096 | Closed canonical scalar object is smaller; `format/portable_context_test.go` tests isolated equality/one-over, all fields, timeout endpoints and canonical encoding. |
| Attempt result bytes | 4096 | Closed canonical scalar object is smaller; `format/attempt_result_test.go` tests isolated equality/one-over and all allowed result branches. |

## Constraint arguments

The bundle hint has two fixed names, one fixed schema ID and an ASCII repository ID no longer than 128 bytes. Its complete canonical form is below 512 bytes. Whitespace cannot manufacture equality because canonical bytes are required.

A conservative valid-manifest upper bound is 897044 canonical bytes: source artifacts at most `64 × 5300`; files `66 × 2250`; notice paths `64 × 2060`; license-ref bindings `64 × 3150`; dependencies `64 × 400`; conflicts `64 × 310`; capability IDs `64 × 200`; compatibility majors `64 × 20`; and 16384 bytes for all remaining names, values and framing. These deliberately loose per-item bounds include delimiters. Single-line source text excludes control characters, so four bytes per scalar bounds its canonical encoding. Original paths are ASCII and at most double in JSON escaping. Each bound license-ref token appears in the 1024-character ASCII license expression. Pack keys, versions, digests and capability identifiers use bounded ASCII grammars. Changing those schemas requires reviewing this argument; a newly reachable equality needs a complete fixture.

Entry, object and relationship lower bounds are derived conservatively from their typed required fields in `testContentCountReachability`. With LF they require at least 271, 278 and 247 bytes respectively. Their count ceilings therefore require at least 542000000, 556000000 and 1235000000 bytes, each exceeding the 268435456-byte single payload-file limit. Each profile has exactly one file for that row kind.

A set has at most one member per key and only sixteen keys are admitted. A portable catalog has at most 1024 sets, so no more than 16384 distinct required set/member pairs can exist. The generic 64-member and 65536-pair ceilings remain defensive guards for future catalog growth; they do not authorize unknown keys or duplicate members.

The successful envelope duplicates bounded evidence intentionally. Its two current-root fields, two histories and six role files have at most 36 MiB decoded bytes under the per-field and 10 MiB history limits. Base64 requires at most 48 MiB plus bounded padding. Each root transition's two signer lists encode at most two key IDs per supplied valid signature; their combined canonical size is bounded by the corresponding signed root file, including its larger signature values and mandatory root structure. Thus all transition objects are bounded by the 10 MiB proposal-history allowance. Map-key/framing, role and binding overhead stays below 2 MiB under the 10000-root and 64-signer bounds. This is below the 64 MiB envelope ceiling. The prepared result adds at most a Base64 one-MiB manifest and 68 bounded descriptors; even starting from the full 64 MiB envelope allowance, this remains below 71303168 bytes.

An admitted diagnostic path uses at most twice 4096 bytes after canonical escaping; a single-line entry ID at most twice 512 bytes. Registered tokens, bounded related identities, fixed codes and object framing keep a complete issue below 12 KiB. A 1000-issue summary is therefore below 16 MiB. Frozen portability context and attempt-result objects contain only fixed names/tokens, hashes, booleans and bounded integers, so neither can approach 4096 canonical bytes.

## Evidence scope and final validation

This inventory makes component scope explicit. The final production implementation passed `make test-fast` 737/737 at `20261004T044530Z-p34338` and `make check` 990/990 at `20261004T045139Z-p26860`; the complete final `make release-check` passed 1259/1259 at `20261004T063630Z-p80139`. The pre-production implementation ledger is complete for its requested scope.

Execution and schema/semantic review remain distinct from production owner adoption and the target deployment's security-authority disposition. The project owner explicitly excludes recreation/interchangeability reports from this pre-production cutover. No passing component guard or fixture count establishes production adoption. In particular, structural directory/TAR framing evidence must not be represented as a fully authenticated lifecycle fixture. Reachability arguments above preserve production limits and identify where complete at-limit artifacts would contradict another admitted constraint.
