# Entra application IDs v2 delivery

The current release is **2026.11.07.2**, release sequence **2**, with 139 entries.
The content profile and entry schema are v2. Shared pack contract v2, manifest
v1, complete sets v2, incident format 7 and `reference_pack_refs.v2` are unchanged.
There is no Entra v1 compatibility reader, conversion, dual writing or reset.

No v1 Reference Pack deployment exists or has ever existed. The only existing
reference packs are the artifacts currently retained under `fixtures/`,
including the ATT&CK, Windows Event IDs and Entra fixture families. Predecessor
schema files and synthetic incompatible-state test cases are design and test
evidence. There is no v1 deployment data to migrate, and this delivery requires
no migration authorization or environment reset.

The initial delivery described below implements the maintainer-approved plan against starting commit
`34e307cf55a22a1722de2c5ac0e39c1c8c3e55e0`. The supplied research plan is source
context; current adopted owner contracts and the maintainer's subsequent v2 and
factual/original-text decisions govern implementation. No external deployment
was changed. The implementation remains in the working tree for review.

## Public release and retained inputs

The [current release descriptor](/home/jochi/code/cartulary/fixtures/entra-app-ids/release.v2.json)
binds the exact canonical files and renewed envelope. Product tests read this
JSON and canonical pack bytes, never this document or upstream Markdown.

- [Signed pack](/home/jochi/code/cartulary/fixtures/entra-app-ids/2026.11.07.2/envelopes/5c45b5078da238c2c274782e90136895335927f53e98b5c8a69080957bee0db6/container.pack).
- [Public bootstrap](/home/jochi/code/cartulary/fixtures/entra-app-ids/2026.11.07.2/envelopes/5c45b5078da238c2c274782e90136895335927f53e98b5c8a69080957bee0db6/bootstrap.json).
- [Exact signed manifest](/home/jochi/code/cartulary/fixtures/entra-app-ids/2026.11.07.2/logical/e26b14a616c643b95ccbda7e2a224f9a3f89612f506b4335590bb68730c13629/manifest.json).
- [Renewal record](/home/jochi/code/cartulary/fixtures/entra-app-ids/2026.11.07.2/envelopes/5c45b5078da238c2c274782e90136895335927f53e98b5c8a69080957bee0db6/envelope-report.json).

| Identity | SHA-256 |
| --- | --- |
| Producer executable | `744126a524f8213f2d81368d2d4e903d4cb88b2c0279115d349f81ab3f28a503` |
| Final input inventory | `cd77ef64c1133a2780e1c8a7b8fff0172a2623c6ef7a681b7335c8a685a4f8f8` |
| Accepted source lock | `5cf55e78f58cdb4ed501bb6f4673ccd73d98987dd208ab3a447b6c79a09ced1c` |
| Source profile | `6266ef7000f784475bb6ab1b5a2028ba7872efce0ff1e25d095530434add3126` |
| Reconciliation | `9dd211916a0025fa3479094806d23908cbe756e44f3d2391f51f80064825938d` |
| Bound distribution decision | `fca16efd399b4536c92a44496ec3be9bd0044c905fd009e403d3adbe8c1ee106` |
| Canonical entries | `8e4be86e1b452c6883032227a2a7c32659e9288d0485ba20a6c66355fef8e802` |
| Manifest | `e26b14a616c643b95ccbda7e2a224f9a3f89612f506b4335590bb68730c13629` |
| Framed payload | `1b0c84e5c95d14b2f7889e5ad5f5ebfdf55b88dc395635dc43439d787efdc3d5` |
| Renewed container | `5c45b5078da238c2c274782e90136895335927f53e98b5c8a69080957bee0db6` |
| Public bootstrap | `f3f71533c3e1edd3a19e8d348b41044c4e4112575d2dc45e5b4fee196c8769ba` |

Source acceptance was **2026-11-07T07:39:00Z**; build time was
**2026-11-07T07:39:47Z**. The renewed metadata version is **4**, issued at
**2026-11-07T07:40:58Z**, with online expiry **2026-12-07T07:40:58Z** and root
expiry **2027-11-07T07:15:58Z**. These are actual acquisition/build/envelope
parameters, independent of the research cutoff of 2026-11-06.

The publisher is `cartulary.entra_app_ids`. Continuous tests generate separate
`fixture.entra_app_ids.v2` keys in memory and fresh envelopes against the trusted
wall clock. The real key is at
`/home/jochi/.local/share/cartulary/reference-pack-signing/entra-app-ids/v2/signing-context.json`;
its directory is mode `0700`, the file is mode `0600`, and neither is in the
repository or public archive. The bootstrap establishes this local authority;
an operator must explicitly configure destination trust before deployment import.

Research retention is under
`.cartulary/reference-packs/entra-v2/sources-final/`: exact source/license bytes,
actual acquisition receipts, immutable lock, audits, candidate payload and
`publications/<manifest-sha>/` containing field accounting, exclusions,
editorial inputs, distribution decision and build report. The immutable builder
and two-build evidence are under
`.cartulary/reference-packs/builders/<producer-sha>/`. Reproduction is under
`.cartulary/reference-packs/entra-v2/reproduction-final/`.
Logical reservations and envelope reservations are separate:
`.cartulary/reference-packs/entra-v2/publication-ledger/` and
`.cartulary/reference-packs/entra-v2/envelope-ledger/`.
The final production-verifier receipt is
`envelope-ledger/verification/046a6750877355ba500521386419fba8db81bdc339e653bd25712b6a6a38497b/verification.json`.
Preserve these ignored local records with the private publishing custody backup;
Git alone does not retain them. The earlier 2026.11.07.1 v2 build, original lock,
envelopes and evidence remain immutable historical records.

The 118,272-byte USTAR has exactly eight regular members:

```text
bundle.json
manifest.json
metadata/snapshot.json
metadata/targets.json
metadata/timestamp.json
notices/Apache-2.0.txt
notices/attribution.txt
payload/entries.ndjson
```

All four logical members were compared byte for byte between independent builds.
The archive manifest equals its sidecar. Independent Node witnesses verify
canonical framing, all source revisions/digests, baseline identities, aliases,
original descriptions, exact locators, archive integrity and Ed25519 signatures.
The production verifier separately establishes trust freshness and registered
content validation. No private key, raw source, upstream description, upstream
license evidence, research report or reconciliation record is an archive member.

## Corpus and editorial accounting

| Measure | Count |
| --- | ---: |
| Selected physical rows | 145 |
| Candidate assertions | 147 |
| Quarantined assertions | 7 |
| Admitted assertions | 140 |
| SharePoint coalescences | 1 |
| Entries | 139 |
| Current-source-supported entries | 20 |
| Historical-only entries | 119 |
| Entries with aliases | 135 |
| Aliases | 139 |
| SPNs | 0 |

The [immutable inventory](/home/jochi/code/cartulary/contracts/reference-pack-producers/entra-app-ids/inventory.v2.json)
retains all three source revisions, all three license revisions, byte lengths,
digests, commit times, front-matter dates and section/header selection. Exact
commit-pinned URLs carry upstream repository paths; there is no ambiguous
project-relative `path` field. S1 is commit
`ad2392b54e54af312a83c1095d1532609ac46a05`; S2 is
`6052ca412335c69ac07a3eaca9ee6cc9d1a097b1`; S3 is
`4d3e713353ecd290d8615487c486df4b5649501c`.

Quarantine matches the pinned digest, line locator and original cells for S1
lines 36, 37, 42, 43, 49 and 50, and S3 line 117. No malformed ID is repaired.
Both multi-ID rows retain their separate IDs and duplicate labels, with ambiguous
aliases removed from all affected entries. The documented Defender/Radius and
Viva Engage/Yammer name relationships retain all four additional aliases.
The source inventory, full baseline and disposition report make every admitted,
excluded, coalesced and unselected row inspectable.

The 16 S1 summaries were checked against their retained source rows for factual
support and original wording. Four S2-only entries have null descriptions. All
119 S3-only entries retain the exact producer-authored historical qualification.
Microsoft publisher attribution means documentation attribution, not Graph
publisher verification, tenant observation, ownership proof or authorization.
The maintainer's Apache-2.0 instruction is recorded in a machine-readable policy
and bound to the actual accepted lock. It covers the factual catalog and original
text; retained upstream research keeps its original licensing.

## Workstreams, gaps and exit criteria

| Phase and dependency | Remediation and area | Rationale and long-term benefit | Compatibility impact and unresolved risk if omitted | Validation exit |
| --- | --- | --- | --- | --- |
| 1. Owner adoption, first | Specification, contracts, implementation and tests: admit Entra v2 only; clarify pinned upstream data versus executable prose. | One current profile removes compatibility branches; explicit source authority keeps expectations independent. | Pre-deployment schema change with no deployed v1 data or migration prerequisite. Accepting unsupported schema inputs would contaminate retained sets. | Current manifests/sets/references/preparations succeed; synthetic v1 inputs fail; startup and operator rejection preserves test state. Database migration applies schema changes without interpreting pack content. |
| 2. Immutable producer, after 1 | Producer inputs/code/tests: pinned URLs, bounded acquisition, exact locks, section parser, quarantine, aliases, complete dispositions and field accounting. | A small source-specific extractor shares canonicalization and publication infrastructure; strict inventory eliminates silent corpus erosion. | Updates require new locks/releases. Conflicting claims cannot be repaired automatically. | Exact six artifacts, baseline counts, original locators, offline checking and no public output on failure. |
| 3. Factual projection, after 2 | Specification, editorial inputs, notices and tests: factual names/IDs, 16 original summaries, 119 historical notes, explicit Apache-2.0 decision. | A visible distribution boundary preserves historical qualifications without copying source descriptions. | Deliberate description/digest change; dropping qualifications or copying prose would violate the selected policy. | Baseline and editorial witnesses pass; archive inventory contains only approved logical/public trust artifacts. |
| 4. Publishing, after 1–3 | Shared implementation and documentation: explicit real/test publisher policies, neutral release verification, separate sequence/envelope ledgers and immutable renewal. | Reusing JCS, USTAR, TUF and scoped reservations avoids a second cryptographic subsystem. | Independent Entra trust must be configured explicitly; custody loss prevents renewal. Logical and envelope ledgers must remain separate. | Reproduced builder and logical files, accepted production verification, independent signatures, replay preservation and collision rejection. |
| 5. Owner integration, after 4 | Tests and narrow helper reuse: actual HTTP/Jobs admission, activation, all consumer interfaces, exact-set reporting, format-7 destination import and recovery. | Full-corpus scenarios test existing boundaries without adding a parallel importer, route or analyst schema. | No public route/UI/incident-field change. Missing coverage could hide rebasing, unintended activation or retention loss. | All entries and aliases round-trip; Base+Entra sets are exact; A survives B; missing A fails; destination import stays inactive; active/pinned recovery preserves provenance and trust. |
| 6. Qualification and handoff, alongside 2–5 | Tests, authored routing, generated projections and documentation: full corpus, signed negatives, SPN bounds, retained-state cases, fresh test authority and explicit readiness claims. | Closed fixture bindings and focused owner rows prevent unexecuted fixtures and expiry-dependent tests. | Test trust is never deployment trust. Broader repository acceptance is not inferred from selected rows. | Final owner portfolio, static/generated checks and documented renewal with failed/skipped evidence preserved. |

The release-evidence fixture producer now reads the active profile registry and
split fixtures, removing its stale aggregate-v1 input and hardcoded profile
version. The shared real-envelope verifier derives receipt identity from its
publisher policy. A pre-existing Sysmon error-string capitalization lint finding
was repaired without changing its rejection behavior.

The subsequent review identified unresolved reconciliation atomicity, strict
retained admission, startup ordering and continuous corpus-oracle gaps. The
remediation record below supersedes the earlier unqualified completion claim.
Pending opaque uploads are unverified inputs and can be admitted as Jobs; their
content is not admitted runtime state. Current verification rejects retired
entries/profiles. Retained-admission tests seed synthetic incompatible pending
operations, typed bindings, portable preparations, prepared results, snapshots,
releases and compositions. Those fixtures exercise rejection before mutation;
they are not records of a v1 deployment.

## Maintenance

From the repository root, the executed final production commands use retained
JSON parameters in `.cartulary/reference-packs/entra-v2/inputs/`:

```sh
make reference-pack-producer-build
make reference-pack-source ACTION=acquire SOURCE_LOCK=contracts/reference-pack-producers/entra-app-ids/inventory.v2.json SOURCE_DIR=.cartulary/reference-packs/entra-v2/sources-final
make reference-pack-source ACTION=check SOURCE_LOCK=contracts/reference-pack-producers/entra-app-ids/inventory.v2.json SOURCE_DIR=.cartulary/reference-packs/entra-v2/sources-final
make reference-pack-source ACTION=audit SOURCE_LOCK=contracts/reference-pack-producers/entra-app-ids/inventory.v2.json SOURCE_DIR=.cartulary/reference-packs/entra-v2/sources-final
make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/entra-v2/inputs/assemble-final.json
make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/entra-v2/inputs/reproduce-final.json
make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/entra-v2/inputs/envelope-final.json
make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/entra-v2/inputs/renew-final.json
```

A content update requires explicit new source pins/profile decisions, new
acquisition receipts, reconciliation/editorial review, a new version
`YYYY.MM.DD.N`, and the next logical reservation. Do not edit old locks or release
bytes. Build inputs bind the exact producer executable; preserve that executable
and its reproducibility record when preserving a release.

Envelope renewal keeps the logical root, pack version, manifest and payload
unchanged. Create a new closed input using the current envelope directory as
`previous_envelope`, the existing private context, the separate envelope ledger,
`create_signing_context: false`, a new actual UTC issuance time, fresh online
expiry within owner limits and the next metadata version. Run the public `renew`
target, inspect production verification and the independent witness, then update
the public descriptor to the new immutable envelope. Do not extend root validity
or change keys implicitly; root maintenance uses the existing trust policy.

The real-release integration row is explicit-only because its retained envelope
expires. Continuous tests re-sign the same canonical corpus with independent
fresh test keys and current wall-clock expiry. No Network Flow clock-control
route is used. A new sign-in adapter, analyst UI and report enrichment content
remain outside this delivery.

## Initial delivery verification

These runs record the initial implementation before the subsequent machinery
audit below. In particular, the earlier migration rejection row records a
then-existing scan that the audit removes; it is not a current acceptance
requirement. Every invocation below used `CARTULARY_TEST_RUN_ID=<run-id>` from
`/home/jochi/code/cartulary`. Its exact run root is
`/home/jochi/code/cartulary/.cartulary/test-results/<run-id>/`. Tool invocations
publish `<target>/tool-run-summary.json`; graph invocations publish
`run-summary.json`, `target-summaries/`, `rows/`, `unit-results/` and
`cleanup-results.json`. These are evidence locations, not executable inputs.

The final implementation was formatted and generated before `agent-finalize`.
After finalization, the final builder and release were produced; only public
release artifacts, their descriptor and this handoff were added. Final owner
tests consume that descriptor and the final corpus. Earlier focused runs below
are retained diagnostics, not substitutes for this final portfolio.
All eight final product test invocations record the same harness source digest:
`sha256:b36383def6da00dee98ac473dceb8b5b716ef2ce3c05c48a2bbb5991d2ef5aea`.

| Run ID | Command / selection | Result and observed outer exit | Added obligation |
| --- | --- | --- | --- |
| `entra-v2-final-format` | `make format` | Pass, 0 | Authored source formatting. |
| `entra-v2-generate-3` | `make generate` | Pass, 0 | Authored contracts/routing regenerated through Make. |
| `entra-v2-finalize-2` | `make agent-finalize` | Pass, 0; generated files unchanged | Schema, catalog, tiers, generated structure and executable source boundaries. |
| `entra-v2-final-builder` | `make reference-pack-producer-build` | Pass, 0 | Two independent identical producer binaries. |
| `entra-v2-final-acquire` | Source `ACTION=acquire`, exact inventory and `sources-final` paths above | Pass, 0 | Actual acquisition and six exact retained bytes/receipts. |
| `entra-v2-final-source-check` | Source `ACTION=check`, same paths | Pass, 0 | Offline lock, length, digest, UTF-8 and regular-file admission. |
| `entra-v2-final-audit` | Source `ACTION=audit`, same paths | Pass, 0 | Complete accounting and independent JSON witness. |
| `entra-v2-final-assemble-2` | Package `ACTION=assemble INPUT=.cartulary/reference-packs/entra-v2/inputs/assemble-final.json` | Pass, 0 | Validated canonical v2 content and sequence 2. |
| `entra-v2-final-reproduce-2` | Package `ACTION=assemble INPUT=.cartulary/reference-packs/entra-v2/inputs/reproduce-final.json` | Pass, 0 | Second independent logical build, all four members byte-identical. |
| `entra-v2-final-envelope` | Package `ACTION=envelope INPUT=.cartulary/reference-packs/entra-v2/inputs/envelope-final.json` | Pass, 0 | Real publisher, production verification and independent signature/archive witness. |
| `entra-v2-final-renew` | Package `ACTION=renew INPUT=.cartulary/reference-packs/entra-v2/inputs/renew-final.json` | Pass, 0 | Immutable logical identity, fresh metadata version 4 and retained renewal record. |
| `entra-v2-final-replay` | Package `ACTION=assemble INPUT=.cartulary/reference-packs/entra-v2/inputs/assemble-final.json` | Pass, 0 | Exact publication replay preserves bytes. |
| `entra-v2-final-collision` | Package `ACTION=assemble INPUT=.cartulary/reference-packs/entra-v2/inputs/collision-final.json` | Expected rejection; Make 2, producer 1 | Same release identity with different build input fails; rejected output root remains absent. |
| `entra-v2-final-reference` | `make test-slice OWNER=module.reference_data ROWS=<reference rows below>` | Pass, 0 | Full corpus, real release, consumers, signed negatives, source parser, SPN boundaries and fixture closure. |
| `entra-v2-final-reporting` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.entra_snapshot_binding` | Pass, 0 | Snapshot A remains exact through B activation, rendering and rerendering; missing A fails. |
| `entra-v2-final-portability` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.entra_bundle_round_trip` | Pass, 0 | Format-7 destination trust and exact content/provenance; newly imported content stays inactive. |
| `entra-v2-final-recovery` | `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.entra_recovery` | Pass, 0 | Active B and snapshot-pinned A, roots/metadata/envelopes/sets/provenance restored; missing and corrupted A each prevent readiness. |
| `entra-v2-final-startup` | `make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission` | Pass, 0 | Startup rejects synthetic v1 test records before database/filesystem mutation. |
| `entra-v2-final-migrate` | `make test-slice OWNER=app.migrate ROWS=app.migrate.integration.reference_pack_retained_admission` | Pass, 0 | Database migration entrypoint rejects synthetic v1 test records before database/filesystem mutation. |
| `entra-v2-final-operator` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission` | Pass, 0 | Operator rejects synthetic v1 test records before database/filesystem mutation. |
| `entra-v2-final-shared` | `make test-slice OWNER=module.reference_data ROWS=<shared rows below>` | Pass, 0 | Targeted ATT&CK and Windows producer/lifecycle regressions for shared packaging changes. |
| `entra-v2-final-go` | `make lint CHECK=go` | Pass, 0 | Format, vet and staticcheck. |
| `entra-v2-final-gosec` | `make go-gosec-targeted` | Pass, 0 | Targeted Go security checks. |
| `entra-v2-final-schema` | `make source-check CHECK=schema` | Pass, 0 | Closed machine contracts. |
| `entra-v2-final-generated` | `make source-check CHECK=generated` | Pass, 0 | Generated-artifact policy and drift. |
| `entra-v2-final-scripts` | `make lint CHECK=scripts` | Pass, 0 | Producer witness and release-evidence JavaScript. |
| `entra-v2-final-boundary` | `make lint CHECK=backend-boundary` | Pass, 0 | Exact backend module dependencies. |
| `entra-v2-final-markdown` | `make lint-markdown` | Pass, 0 | Specification, repository guidance and completed handoff Markdown. |

Markdown lint uses the legacy wrapper summary at
`adhoc/lint-markdown/tool-run-summary.json` within its run root. Ordinary
`git diff --check` also passed. No test or publishing process remains running.

The seven reference rows, passed as one comma-separated `ROWS` value, were:

```text
module.reference_data.integration.entra_canonical_lifecycle
module.reference_data.integration.entra_real_release_import
module.reference_data.unit.entra_source_producer
module.reference_data.unit.retained_binding_admission
module.reference_data.unit.enrichment_entra_app_ids_independently_validates_02047fa20a
module.reference_data.unit.canonical_consumers
module.reference_data.unit.canonical_fixture_manifests
```

The four shared regression rows were:

```text
module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c
module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930
module.reference_data.integration.attack_producer_artifacts_preserve_independent_a_67e9ef139a
module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e
```

Owning terminal summaries, selected row results and same-run cleanup records
were inspected. Product invocations executed with no skipped or cached product
units; fixture/service/runtime cleanup completed without reported failures.
Real-release and continuous-test lifecycle overlap is intentional: one proves
the shipped signature/bootstrap, the other proves durable testing under fresh
separate authority. Reporting, portability and recovery add different owner
boundaries. This selected portfolio establishes the requested feature scope;
it is not a complete-owner evidence audit or repository-wide release acceptance.

### Preserved failures and repairs

No failed run was deleted, overwritten or relabeled as passing. These roots
remain under the same result parent. Earlier failures relate to implementation
development or fixture construction, with successful final replacement evidence
above:

| Run ID / target | Observed failure and classification | Repair and final evidence |
| --- | --- | --- |
| `entra-v2-focused-unit` / `test-slice` | Make 2; product/test_assertion_failure: synthetic SPN ordering violated ASCII comparison. | Corrected fixture order; final reference profile row passes. |
| `entra-v2-focused-lifecycle` / `test-slice` | Make 2; product/test_assertion_failure: provenance index 9999 reached the schema limit before the intended semantic check. | Used an in-schema index equal to artifact count; signed negative now reaches `content_semantics`. The real release already passed in the failed aggregate. |
| `entra-v2-focused-contracts` / `test-slice` | Make 2; product/test_assertion_failure: owner fixture `input_refs` not sorted. | Sorted authored refs and regenerated; final fixture closure passes. Prepared admission passed in the failed aggregate. |
| `entra-v2-focused-startup` / `test-slice` | Make 2; product/test_assertion_failure: test setup attempted to update immutable history. | Seeded the incompatible history directly, preserving production immutability; final startup/migrate/operator rows pass. |
| `entra-v2-focused-recovery-2` / `test-slice` | Make 2; product/test_assertion_failure: test queried nonexistent `reporting_snapshot` pin kind. | Asserted the actual `reporting_snapshot_job` pin and job identity; final recovery passes. |
| `entra-v2-lint-go` / `lint` | Make 2; source formatting, ST1005 error strings and an overlapping constant rename while the check was running. | Final source frozen, formatted and linted successfully; no concurrent final source edits. |
| `entra-v2-finalize` / `agent-finalize` | Make 2; harness/child_target_failure: lexical boundary identified redundant upstream `path: docs/...` as a project-prose dependency. | Removed the redundant field; exact immutable URL retains the upstream path. Reacquired a new immutable lock and rebuilt. No boundary exception or bypass; `finalize-2` passes. |
| `entra-v2-final-assemble`, `entra-v2-final-reproduce` / `reference-pack-package` | Make 2, producer 1; harness/unknown_failure: logical allocator encountered envelope directories in a shared ledger root. | Corrected input layout to separate ledgers. Only task-created envelope record directories moved byte-unchanged; logical reservations retained. Correction recorded in `inputs/ledger-layout-correction.json`. Final assemble/reproduce/replay pass. |
| `entra-v2-final-collision` / `reference-pack-package` | Intentional Make 2, producer 1; harness/unknown_failure, diagnostic `fixture publication identity collision`. | Expected safety result. Changed build time under an already reserved version; verified no output publication. |

Initial task-guide preflight also identified unsorted test selectors; authored
selectors were sorted before successful guide compilation. Those preflight and
diagnostic argument errors are not product tests or acceptance evidence. Failed
product-run cleanup records reported no cleanup failures.

### Unperformed checks and limits

- `agent-finalize` did not run retained-run canonical evidence and scheduler
  timing maintenance: no eligible full warm-check `RESULTS_DIR` was supplied.
  Its `unit-artifacts/finalize-summary.json` records `results-dir-not-provided`;
  source schema/catalog/generated checks did execute.
- Full `make check`, CI/release qualification and complete-owner evidence
  audits were not selected. The task required targeted end-to-end coverage;
  generic authorization, replay, archive, interruption, concurrency and removal
  infrastructure was retained rather than cloned. The focused portfolio does
  not claim those broader suites executed in this delivery.
- Browser/UI review was not run because no presentation, adapter or analyst
  workflow was added. Trust was configured only in isolated test deployments.
- Human projection-fidelity review remains the repository's review step;
  passing generated contracts/tests does not prove specification completeness.

## Readiness and handoff

| State | Completion evidence and limits |
| --- | --- |
| Source-complete | Complete: six exact retained artifacts, actual receipts, immutable lock, seven exclusions, full reconciliation and field accounting. |
| Canonical-content complete | Complete: 139 v2 entries; registered schema/semantic validation and independent baseline/editorial witness pass. |
| Fixture-verified | Complete for the selected delivery: full corpus, executed owner fixtures, signed content negatives, synthetic SPN/parser boundaries and fresh independent test trust. |
| Real release produced | Complete: release 2026.11.07.2, reproducible builder/logical bytes, real signed archive/public bootstrap, production acceptance and successful renewal. |
| Deployment-importable | Demonstrated in an isolated deployment with explicit real bootstrap and trusted wall clock. Any destination must configure explicit trust and verify a fresh envelope; trust was configured only in isolated tests. |
| Distribution-cleared | Maintainer-authorized Apache-2.0 factual/original-text policy recorded and bound to accepted source lock; exact public member inventory inspected. This does not relicense retained upstream research. |
| Reporting integration complete | Complete: exact-set admission/render/rerender and missing-content failure under existing interfaces. |
| Portability/recovery complete | Complete: executed destination import and backup/restore of active/pinned content, trust history, envelopes, sets and provenance. |
| Analyst-workflow complete | Not claimed: new analyst presentation, sign-in adapter and report enrichment were explicitly excluded. |

Before using this authority beyond the isolated deployment, retain the private
key and ignored publishing records in operator custody, configure the exact
public bootstrap explicitly, and renew online metadata before the recorded
expiry. These are initial deployment configuration and publishing-maintenance
steps; no v1 migration or replacement environment is required.


## Reference-pack implementation audit — 2026-11-07

The maintainer confirmed that no v1 Reference Pack deployment has ever existed.
The audit therefore treats the concrete artifacts under `fixtures/` as the
existing pack inventory. It found no v1 compatibility reader, automatic pack
converter, dual writer, or data-reset implementation. It did find unnecessary
cutover scaffolding in migration composition, pending-input admission and the
compiled contract bundle.

| Finding | Remediation and ownership | Benefit and validation criterion |
| --- | --- | --- |
| Database migration composed a Reference Data/Reporting content scan for a nonexistent deployment transition. | Remove the migration hook, SQL-reader adapters, duplicate scan loops and migration binding adapter. Move Reporting runtime binding inspection into its own `referenceadmission` package. Update the NLSpec/Core scope and qualification routing. Replace migration rejection fixtures with an isolated test that current schema application neither inspects nor rewrites synthetic invalid pack content. | Schema lifecycle remains independent of content admission; current server/operator and recovery checks still enforce validity before use. Migration facade tests must pass without a Reference Pack content scan. |
| Pending-input admission recursively blacklisted selected old schema strings, including fictitious nested Entra bindings. | Delete that scanner. Share the existing closed current operation decoder between startup admission and execution, and use the current closed portability-input shape. Remove the redundant Entra-pending fixture. | Unknown future identifiers and unknown members fail without maintaining a blacklist. Current input, missing/extra/null fields, unsupported IDs, canonical encoding and worker execution are exercised. |
| Fifteen superseded contract inputs were still compiled despite having no active readers. | Move their exact bytes into `fixtures/reference-packs/historical-v1/contracts/`; regenerate projections through Make. They include old profile/consumer registries, old operation/portability/set/reference/result shapes, superseded ATT&CK/Windows/Entra schemas and the old built-in release binding. | Runtime embeds only the applicable contract inputs. Historical design fixtures are inert. Current manifest/profile/consumer tests and generated checks must pass. |
| Runtime scans silently skipped absent current tables, a concession to pre-migration inspection. | Remove table-existence branches; propagate database errors from required current queries. Add server/operator missing-table scenarios with unchanged-state assertions. | Missing current schema cannot silently admit incomplete content. Failures remain operational errors, distinct from incompatible content. |
| A storage-schema test and its routing still described a reset/cutover and loaded an obsolete portable catalog. | Keep the current opaque-storage-column assertion; remove duplicate predecessor rejection coverage, the historical fixture dependency and the migration/reset naming. Existing runtime rejection scenarios continue to cover unsupported inputs. | Current DDL is tested on its own terms without an invented upgrade history. |
| An unused positive profile-vector bundle still contained predecessor Entra entries and incomplete Windows v2 entries. | Move its exact bytes to `fixtures/reference-packs/historical-v1/fixtures/profiles.pre-entra-v2.json`. Regenerate the fixture projection; retain current individual profile fixtures and synthetic rejection cases. | Obsolete positive examples are no longer compiled alongside current test inputs. |
| Built-in validation compared the current catalog against an obsolete release. | Validate current member lengths/digests against the manifest and the independent current Base witness. | Current identity is sufficient; no predecessor release is required to establish correctness. |

The retained runtime inspection, immutable-set/provenance checks, signature
verification, actual content-profile validation and backup/restore integrity
checks remain necessary for current v2 artifacts. Unchanged manifest/TUF and
other active profile identifiers ending in `.v1` remain their current contracts;
they do not imply a v1 deployment. Entra admits only profile/entry v2. The
existing signed Entra release remains byte-immutable with its original builder
identity; this runtime cleanup does not manufacture a new content release.

Owner-text review also corrected the profile registry table’s stale Entra v1
row to v2 and removed obsolete subsystem-wide “in v1” wording. Active manifest,
TUF and unchanged content-profile v1 identifiers retain their specific meaning.

Focused verification also exposed stale independent complete-set and portability
expectations from the earlier Entra v2 change. Their expected identities were
recomputed with an independent Python standard-library recipe; the retained
derivation is `.cartulary/reference-packs/v2-audit/independent-vectors.json`.
Product tests continue to consume canonical JSON and fixed expectations.

Compatibility is intentionally limited to current contracts. No deployed v1
content needs migration, conversion, reset, or special authorization. The
ordinary database migration mechanism and unrelated owners’ admission checks
remain. Invalid pending shapes now fail earlier at runtime admission; missing
current tables produce operational errors. Leaving the removed scaffolding in
place would retain duplicate validation paths, accidentally compile obsolete
formats, accept undeclared pending fields, and conceal incomplete current
schemas. No public route, current wire format, content identity, or signing
policy changes in this audit.

### Audit verification and handoff

The final portfolio ran after `make agent-finalize` completed. All final product
and static runs below bind repository commit
`34e307cf55a22a1722de2c5ac0e39c1c8c3e55e0`, dirty-source digest
`sha256:38c1072233ce67581e3d40a38169373bc60e18b73a6a37cc22848ee4be585e94`.
No executable inputs changed afterward. Human documentation is outside the
product evidence inputs. Source finalization reported unchanged generated
outputs.

Each command below was prefixed with `CARTULARY_TEST_RUN_ID=<run-id>`.
Each run root is `.cartulary/test-results/<run-id>/`. For graph runs, the owning
summary is `run-summary.json`; `cleanup-results.json` records cleanup. All listed
commands have observed outer exit **0**, passing owning summaries, and no failed
or blocked cleanup. Counts below are harness work units, not individual tests.

| Run ID | Exact Make selection | Passed work units | Completed cleanup steps |
| --- | --- | --- | --- |
| `reference-v2-audit-finalize-4` | `make agent-finalize` | 1 | 10 |
| `reference-v2-audit-final-reference` | Reference Data command below | 8 | 23 |
| `reference-v2-audit-final-server` | `make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission` | 3 | 14 |
| `reference-v2-audit-final-operator` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission` | 3 | 14 |
| `reference-v2-audit-final-migrate` | `make test-slice OWNER=app.migrate` | 4 | 15 |
| `reference-v2-audit-final-reporting` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.entra_snapshot_binding` | 3 | 14 |
| `reference-v2-audit-final-portability` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.entra_bundle_round_trip` | 3 | 14 |
| `reference-v2-audit-final-recovery` | `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.entra_recovery` | 3 | 14 |
| `reference-v2-audit-final-go` | `make lint CHECK=go` | 6 | 15 |
| `reference-v2-audit-final-boundary` | `make lint CHECK=backend-boundary` | 4 | 13 |
| `reference-v2-audit-final-schema` | `make source-check CHECK=schema` | 3 | 12 |
| `reference-v2-audit-final-generated` | `make source-check CHECK=generated` | 5 | 14 |
| `reference-v2-audit-final-gosec` | `make go-gosec-targeted` | 4 | 13 |

The exact Reference Data selection was:

```sh
CARTULARY_TEST_RUN_ID=reference-v2-audit-final-reference make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.retained_binding_admission,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_profiles,module.reference_data.integration.execution_admission_scope,module.reference_data.integration.reference_pack_storage_schema,module.reference_data.integration.entra_canonical_lifecycle,module.reference_data.integration.entra_real_release_import
```

This selection exercises the current Base/catalog identities, closed pending
inputs, all active profile fixture contracts, consumer/prepared content,
worker admission, opaque storage schema, full Entra lifecycle, and real signed
release import. Server/operator scenarios cover synthetic unsupported contracts,
query failures and absent required Reference Data/Reporting tables, asserting
unchanged database/filesystem state. The migration facade applies current schema
without parsing or rewriting pack bytes. Reporting, incident portability and
recovery add their own exact-set/rendering, destination-trust and restore checks.
This is targeted executed coverage, not a full-owner evidence audit or full CI
claim. Initial delivery evidence above remains tied to its original source.

Earlier failures are retained:

| Run ID / command | Observed result | Diagnosis and resolution |
| --- | --- | --- |
| `reference-v2-audit-focused`: `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.retained_binding_admission,module.reference_data.unit.canonical_pack_manifest,module.reference_data.integration.execution_admission_scope` | Outer 2; `product/test_assertion_failure`; 16 cleanup steps, none failed or blocked. | Stale complete-set and portable-catalog expected hashes from the earlier Entra v2 change. Closed-input admission and execution passed. Independent vector recalculation fixed the expectations; final Reference Data slice passed. |
| `reference-v2-audit-migrate`: `make test-slice OWNER=app.migrate` | Outer 2; `product/test_assertion_failure`; 15 cleanup steps, none failed or blocked. | New test setup omitted the candidate parent row and hit SQLSTATE `23503`. Added the required parent; repaired and final migration slices passed. |
| `reference-v2-audit-finalize`: `make agent-finalize` | Outer 2; `config/configuration_error` at `runtime_plan`; no graph run manifest. | Renamed storage-schema row still had its old transaction-policy approval identity. Updated the authored policy entry. Only runtime-build tool evidence exists in this root. |
| `reference-v2-audit-finalize-2`: `make agent-finalize` | Outer 2; `harness/child_target_failure`; 10 cleanup steps, none failed or blocked. | JSON shape validation reported stale generated topology after the routing edit. `make generate` refreshed it. Finalization then passed. |

Focused repairs also passed under `reference-v2-audit-repaired`,
`reference-v2-audit-migrate-repaired`, and
`reference-v2-audit-startup-focused` (each outer 0). The final generation was
`CARTULARY_TEST_RUN_ID=reference-v2-audit-generate-5 make generate` (outer 0,
`generate/tool-run-summary.json` passed). Final Markdown verification uses
`CARTULARY_TEST_RUN_ID=reference-v2-audit-final-markdown make lint-markdown`;
its outer result is reported with completion.

Retained warm-check/evidence maintenance in `agent-finalize` was not selected:
there was no compatible successful full warm-check `RESULTS_DIR`. Full CI,
browser/analyst workflows, new source acquisition and release re-signing were
not run for this runtime audit. The affected contracts and existing signed pack
were covered by the selected owner scenarios; the immutable release does not
need reissuance. No deployment was migrated or reset, and trust was configured
only by the existing isolated test fixtures. Distribution approval and the
excluded analyst-workflow completion claim remain as recorded above.

## Atomic reconciliation and strict-admission remediation

This section records the subsequent four-gap remediation and supersedes earlier
completion claims for those gaps. The earlier runs remain evidence for their
own source states. No deployed V1 pack, migration, reset, compatibility reader,
or conversion is assumed or introduced.

### Owner decisions and implemented changes

The Reference Pack NLSpec now defines the atomic reconciliation boundary,
strict current retained-artifact admission, application admission ordering, and
the distinction between extensible runtime profile validation and this release's
exact factual oracle. The current `cartulary.reference_pack_reconciliation.v1`
identifier describes a terminal internal receipt, not a V1 pack. Pending-input
admission remains limited to its existing resumable operation contracts.

| Gap | Remediation and ownership | Benefit and compatibility impact | Executed exit evidence |
| --- | --- | --- | --- |
| Interrupted Base reconciliation | Specification, Reference Data implementation, application assembly and integration tests: prepare leased objects outside the publication transaction; insert the operation, construct transaction-scoped indexes, publish versions/envelopes/bindings/sets/attestations, and terminalize the receipt in one transaction. Require the existing `ActionFinalizer` at every caller. Its proof reads the exact operation identity, frozen input and terminal outcome. | Removes the durable unfinished synchronous operation and reuses existing commit-proof infrastructure. No DDL or public format changes. `BaseReleaseOptions` now requires a finalizer. Uncertain commits preserve potentially retained objects and block readiness when proof is unreadable. | The canonical-persistence row exercises failures after preparation, operation insertion, connection loss, index construction and before commit; lost acknowledgement and unreadable proof; unchanged database effects on rollback; fresh admission/reconciliation/readiness; guarded collection; existing claimed/unclaimed, stale-selection and pinned-set cases. |
| Permissive retained artifacts | Specification, shared owner decoders, Reporting binding admission and tests: full manifest decoding uses retained candidate distribution kind; standalone sets share strict shape/canonical/identity/order validation; portable catalogs reuse set semantics. Reporting fallback reads and validates original catalog bytes instead of PostgreSQL JSON reserialization. | One admission contract serves startup and consumers. Valid artifacts retain bytes and identities; malformed current representations fail earlier without rewriting them. Valid unavailable portable history remains admissible. Database failures retain their operational classification. | Closed-admission unit cases cover unknown/missing/null/type errors, duplicate JSON keys, trailing/noncanonical bytes, identities, ordering, duplicate members and required registries. Server/operator test malformed current retained artifacts before mutation. Consumer/recovery tests reject malformed retained sets; Reporting binding resolves both local and portable-only history. |
| Startup dependency ordering | Specification clarification, server/operator assembly and integration tests: cryptographic admission, then schema readiness, then shared read-only Reference Data/Reporting inspection, then leases and mutable startup. | Restores schema diagnostics without missing-table skips or content-aware database migration. Operator transport still redacts internal failures as `internal_error`; public formats are unchanged. | Both entrypoints exercise invalid crypto, behind-head schema with absent pack tables, ahead-head and wrong lineage, head-stamped missing tables, incompatible current content and valid content. They assert earliest failure and unchanged state. The migration owner still applies schema without interpreting pack content. |
| Incomplete continuous corpus oracle | Test support, a dedicated authored verification row and documentation: compare all entries against baseline/editorial/inventory JSON, including exact descriptions/defaults and source URLs, revisions, digests, indices and locators. Keep the independent production witness separate. | Protects release facts without freezing one corpus into runtime V2 validation. No runtime schema, signed artifact or canonical payload changes. Synthetic lifecycle B is mutated only after validating original A. | `entra_corpus_expectations` passes the real corpus and rejects altered/missing historical notes, changed summaries/null policy, wrong in-range indices, source identity changes, missing/duplicate provenance and incorrect defaults. Full lifecycle and downstream owners use the strengthened oracle. |

The oracle enforces 139 entries, 119 exact historical qualifiers, 16 original
summaries, four null descriptions, 135 entries with aliases and 139 total aliases.
Product checks read canonical artifacts and authored JSON; they do not read
upstream Markdown. Only the explicit offline producer operations below inspect
retained source inputs. Retained portable catalogs use their owner-defined
canonical bytes without the transport encoder's terminal newline.

### Final source and immutable inputs

The final implementation portfolio below ran after
`CARTULARY_TEST_RUN_ID=reference-remediation-finalize make agent-finalize`
completed with outer exit 0. Its runs bind commit
`34e307cf55a22a1722de2c5ac0e39c1c8c3e55e0` and dirty-source digest
`sha256:4ece7546abfbab9ed432fc9119f91ebfef5272636919165ff415d7f3c0411152`.
The toolchain digest is
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
No executable inputs changed after finalization. Subsequent handoff edits are
human documentation, excluded from executable evidence inputs.

The before/after comparison of all 36 files beneath `fixtures/entra-app-ids/`
and `contracts/reference-pack-producers/entra-app-ids/` reports no changes in
`.cartulary/reference-packs/remediation/immutable-comparison.json`; its baseline
is `immutable-before.json` in the same directory. This preserves the signed
release, bootstrap, manifest, corpus, source inventory, quarantine/editorial
inputs, notices and distribution decision. The existing release remains
`2026.11.07.2`, sequence 2, with online expiry `2026-12-07T07:40:58Z` and root
expiry `2027-11-07T07:15:58Z`. Original release reproduction and renewal evidence
above retain their original builder identity; this remediation did not reissue
the release.

### Final executed portfolio

Each table command has the prefix `CARTULARY_TEST_RUN_ID=<run-id>` and root
`.cartulary/test-results/<run-id>/`. Every listed invocation completed with
observed outer exit **0**, a passing owning summary, and no failed or blocked
cleanup. Graph summaries are `run-summary.json`, with target projections under
`target-summaries/` and cleanup in `cleanup-results.json`. Work-unit counts are
not test counts. No selected test was skipped.

| Run ID | Exact Make command/selection | Passed work units | Completed cleanup steps |
| --- | --- | --- | --- |
| `reference-remediation-finalize` | `make agent-finalize` | 1 | 10 |
| `reference-remediation-final-reference` | Reference Data command below | 9 | 21 |
| `reference-remediation-final-server` | `make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission` | 3 | 14 |
| `reference-remediation-final-operator` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission` | 3 | 14 |
| `reference-remediation-final-migrate` | `make test-slice OWNER=app.migrate` | 4 | 15 |
| `reference-remediation-final-reporting` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.entra_snapshot_binding` | 3 | 14 |
| `reference-remediation-final-portability` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.entra_bundle_round_trip` | 3 | 14 |
| `reference-remediation-final-recovery` | `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.entra_recovery` | 3 | 14 |
| `reference-remediation-final-go` | `make lint CHECK=go` | 6 | 15 |
| `reference-remediation-final-boundary` | `make lint CHECK=backend-boundary` | 4 | 13 |
| `reference-remediation-final-schema` | `make source-check CHECK=schema` | 3 | 12 |
| `reference-remediation-final-generated` | `make source-check CHECK=generated` | 5 | 14 |
| `reference-remediation-final-gosec` | `make go-gosec-targeted` | 4 | 13 |

The exact 16-row Reference Data command was:

```sh
CARTULARY_TEST_RUN_ID=reference-remediation-final-reference make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.attack_producer_artifacts_preserve_independent_a_67e9ef139a,module.reference_data.integration.canonical_persistence,module.reference_data.integration.entra_canonical_lifecycle,module.reference_data.integration.entra_real_release_import,module.reference_data.integration.execution_admission_scope,module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e,module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.enrichment_entra_app_ids_independently_validates_02047fa20a,module.reference_data.unit.entra_corpus_expectations,module.reference_data.unit.entra_source_producer,module.reference_data.unit.retained_binding_admission,module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930
```

Selection was confirmed by
`make target-plan TARGET=test-slice INPUTS=.cartulary/reference-packs/remediation/reference-selection.json DETAIL=rows`
(outer 0). The canonical-persistence row includes both new integration tests,
`TestBaseReconciliationAtomicPublication_Integration` and
`TestCurrentRetainedStructureAcrossConsumersAndRecovery_Integration`. Its
published row result is passed; the Go runner requires a passing terminal result
for every selected test and rejects missing/skipped selectors. Passing runs
need not retain verbose per-test console output.

The Reference Data portfolio adds format/profile validation, fixture closure,
execution admission, producer/oracle, full Entra lifecycle, existing real-release
import, and all four ATT&CK/Windows producer/lifecycle regression rows against
this final source. Downstream rows independently add exact-set Reporting
render/rerender/missing-content behavior, format-7 destination import, and
active/pinned recovery. The static checks cover the shared decoder and assembly
changes. This is a targeted portfolio, not full-owner or repository-wide
acceptance and not proof of specification completeness.

The current producer was rebuilt and the retained pinned inputs checked offline:

```sh
CARTULARY_TEST_RUN_ID=reference-remediation-final-builder make reference-pack-producer-build
CARTULARY_TEST_RUN_ID=reference-remediation-final-source-check make reference-pack-source ACTION=check SOURCE_LOCK=contracts/reference-pack-producers/entra-app-ids/inventory.v2.json SOURCE_DIR=.cartulary/reference-packs/entra-v2/sources-final
CARTULARY_TEST_RUN_ID=reference-remediation-final-source-audit make reference-pack-source ACTION=audit SOURCE_LOCK=contracts/reference-pack-producers/entra-app-ids/inventory.v2.json SOURCE_DIR=.cartulary/reference-packs/entra-v2/sources-final
```

All three outer exits are 0 and their corresponding target
`tool-run-summary.json` files report pass. The rebuilt executable digest is
`8146aab34af6bcf3d7eac330b734fbf123acd96cdf2c2ca983cc8ec818b06938`;
its builder directory under `.cartulary/reference-packs/builders/` contains
`reproducibility.json` recording two byte-identical builds. The audit report
digest is `19cc383ea67bb8f94a6dd8da2a79c8c26cfc80dd61431b85046209797391769d`.
Its witness under `sources-final/witnesses/` is
`audit-19cc383ea67bb8f94a6dd8da2a79c8c26cfc80dd61431b85046209797391769d.json`,
SHA-256 `09002a8bba156c7d68ff7f429aa177c735f7f2072a7d54e1a78d09f12e36345a`.
The independent witness reports exact baseline agreement and the unchanged
entries digest. It declares `upstream_source_verification: false`; pinned-byte
verification is supplied separately by the successful source `check` operation.
Neither operation acquires new sources or reassembles the signed release.

Final documentation verification uses
`CARTULARY_TEST_RUN_ID=reference-remediation-final-markdown make lint-markdown`;
its outer result is reported with completion. This documentation check is
separate from the frozen product-source portfolio.

### Preserved diagnostic evidence

The following failures remain in their original run roots. They led to repairs
before source finalization and are not substituted for final-source acceptance.

| Run ID and selection | Observed result | Relationship and resolution |
| --- | --- | --- |
| `reference-remediation-format`: `make format` | Outer 2, `config/configuration_error` at `runtime_plan`; no graph manifest. | New selector list was not ASCII-sorted. Fixed the authored list; `reference-remediation-format-2` passed and `reference-remediation-generate` regenerated routing (both outer 0). Runtime-build tool evidence is the only successful child evidence in the failed root. |
| `reference-remediation-focused`: `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.retained_binding_admission,module.reference_data.unit.entra_corpus_expectations,module.reference_data.integration.canonical_persistence` | Outer 2, `product/test_assertion_failure`; 12 cleanup steps, none failed/blocked. | Decoder consolidation left an unused import; a retained-catalog unit fixture used transport bytes with a terminal newline. Removed the import and made the fixture use canonical retained bytes; retained decoding stayed strict. |
| `reference-remediation-focused-2`: same three rows | Outer 2, `product/test_assertion_failure`; 4/5 work units passed; 16 cleanup steps, none failed/blocked. | Both unit rows passed. The new corruption fixture changed bytes without updating its storage checksum, triggering SQLSTATE `23514` before the intended reader. Updated that fixture's checksum so the decoder rejection is actually reached. |
| `reference-remediation-server`: final server selection above; `reference-remediation-operator`: final operator selection above | Each outer 2, `infra/preflight_error`; 2/3 work units passed; 14 cleanup steps each, none failed/blocked. | Shared portable-fixture canonicalization needed its import. Fixed the test helper; both repaired runs and the final runs passed. |

`reference-remediation-persistence-repaired` ran the canonical-persistence row,
and `reference-remediation-server-repaired` / `reference-remediation-operator-repaired`
ran their respective admission rows; all completed with outer 0.
`reference-remediation-format-final make format` also completed with outer 0
before finalization (the run ID was supplied through `CARTULARY_TEST_RUN_ID`).
The diagnostic runs have earlier source identities. The final portfolio reused
one cached unit for `module.reference_data.unit.retained_binding_admission`:
`go:d3970936c8381682:d9af86728dfb`, with zero execution time and no fresh process
exit. Its passing result is harness-admitted reused evidence, not a fresh rerun.
The other affected checks executed in that portfolio. The subsequent
identity-binding remediation below records separate fresh validation with result
reuse disabled; it does not rewrite the earlier run's outcome.

### Completion claims and limits

| Claim | Final disposition |
| --- | --- |
| Specification reconciled | The four owner contradictions are resolved in the adopted NLSpec and its implementation projections. Passing tests do not establish completeness of every owner requirement. |
| Implementation remediated | Atomic publication, shared strict readers, application ordering and the independent corpus oracle passed their targeted final-source exits. |
| Fixture verified | Full corpus, focused negatives and lifecycle scenarios executed under fresh isolated test authority; the new oracle row is routed into the Entra verification obligation. |
| Real release verified | The unchanged signed release passed production import under its explicit isolated bootstrap and current trusted wall clock. Expiry remains as recorded above. |
| Deployment trust installed | Not claimed. No deployment trust configuration changed. |
| Distribution policy satisfied | The existing approved factual/original-text policy and its unchanged notices and decision remain in force; this remediation does not establish a new approval. |
| Reporting/portability/recovery complete | The corresponding final-source owner scenarios passed, including exact retained provenance and required loss/corruption failures. |
| Analyst workflow complete | Not claimed; new analyst-facing enrichment remains excluded. |

No required targeted check was intentionally skipped. Full CI, full-owner
evidence-union auditing, browser workflows, new source acquisition, release
reissuance and real deployment trust installation were not selected because
they are outside this remediation's scope. Retained warm-check maintenance was
not selected by finalization because no compatible successful full warm-check
`RESULTS_DIR` was supplied. Continue using the earlier release update/renewal
procedure; runtime remediation alone is not a reason to change immutable release
bytes or source locks.

## Stored identity binding and interruption with pinned state

The follow-up review found that an internally valid set B could pass admission
when its canonical bytes were stored under set A's database identity. The owner
specification now explicitly requires the decoded ID and digest to match the
selected row and requested identity before startup mutation or recovery index
rebuilding. `packformat.DecodeBoundSet` applies that check after the existing
strict canonical decoder. Startup admission, current/historical consumer reads,
provenance, pin capture, reconciliation inspection, selection loading, Reporting
bindings and historical validation use it. Operational database errors retain
their existing classification; a misbound set maps to incompatible retained
state or consumer `pack_unavailable` at the appropriate boundary. No format,
database schema, compatibility reader, migration or rewrite was introduced.

The existing canonical-persistence row now includes a valid-B-under-A fixture.
It verifies rejection by current and historical reads, provenance, pin capture,
selection loading, Reporting binding admission, Base reconciliation, readiness
validation and recovery rebuilding. It compares complete retained rows before
and after rejection. Server and operator admission fixtures independently cover
the same identity substitution and retain their database/filesystem/lease/Jobs
non-mutation assertions. Unit coverage distinguishes a mismatched ID from a
mismatched digest and admits the correctly bound original.

Interruption coverage now starts with two pinned sets: the previous Base set
and a subsequently activated imported registry set. Five fault cases run with
the profile claimed and again unclaimed: failure after operation insertion,
connection loss after insertion, failure before commit, lost acknowledgement
after commit, and unreadable commit proof. The unclaimed case exercises an
actual selection change back to Base. Assertions compare complete row values,
including current pointer/revision and immutable provenance, plus retained
physical bytes. Aborted transactions preserve the entire prior state; committed
effects remain authoritative; unreadable proof signals fatal integrity. Fresh
owner admission/reconciliation, guarded collection, readiness and reads of both
pinned sets complete each scenario. These extend the seven initial-publication
fault cases without introducing another top-level routing row.

The earlier cached-admission claim was corrected in place above. Historical
passing and failed runs remain retained under their original identities. The
new validation uses `CARTULARY_HARNESS_CACHE_MODE=off` explicitly, which disables
execution-result reuse; ordinary compiler/toolchain caches are separate.

### Follow-up source and verification scope

`CARTULARY_TEST_RUN_ID=reference-binding-format make format` completed with outer
exit 0. Focused checks used the following commands, both with observed outer
exit 0 and passing owning summaries:

```sh
CARTULARY_TEST_RUN_ID=reference-binding-focused make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence,module.reference_data.unit.retained_binding_admission CARTULARY_HARNESS_CACHE_MODE=off
CARTULARY_TEST_RUN_ID=reference-binding-server-focused make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission CARTULARY_HARNESS_CACHE_MODE=off
CARTULARY_TEST_RUN_ID=reference-binding-finalize make agent-finalize CARTULARY_HARNESS_CACHE_MODE=off
```

Finalization completed with outer exit 0, one passing work unit and ten completed
cleanup steps. The focused Reference Data run passed four work units with 15
completed cleanup steps; the focused server run passed three with 14. Neither
used a cached result or had failed/blocked cleanup.

The frozen source is commit `34e307cf55a22a1722de2c5ac0e39c1c8c3e55e0`,
dirty-source digest
`sha256:24d05a38d8e8cdef7af2d6c79210025382ee9acc736662ff9b379fe309a166b9`,
toolchain digest
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
No executable inputs changed after finalization. Existing routing already
executes the extended tests; no authored routing or schema projection needed
changing. Finalization and generated-artifact validation cover projection drift.

The final Reference Data selection is retained in
`.cartulary/reference-packs/identity-binding/reference-selection.json`.
`make target-plan TARGET=test-slice INPUTS=.cartulary/reference-packs/identity-binding/reference-selection.json DETAIL=rows`
completed with outer 0 and selected ten rows. The exact execution command is:

```sh
CARTULARY_TEST_RUN_ID=reference-binding-final-reference make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.attack_producer_artifacts_preserve_independent_a_67e9ef139a,module.reference_data.integration.canonical_persistence,module.reference_data.integration.entra_canonical_lifecycle,module.reference_data.integration.entra_real_release_import,module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.entra_corpus_expectations,module.reference_data.unit.retained_binding_admission CARTULARY_HARNESS_CACHE_MODE=off
```

All final graph invocations below completed with observed outer exit **0** and
passing owning summaries. Each used the prefix
`CARTULARY_TEST_RUN_ID=<run-id>` and the final argument
`CARTULARY_HARNESS_CACHE_MODE=off`. Every run reports **zero cache hits**, zero
skipped/failed/cancelled work units, completed cleanup with no failed or blocked
steps, and a passing retained-secret scan. Every unit has an observed zero
process exit in its published result. All manifests bind the frozen source
digest above.

Run roots are `.cartulary/test-results/<run-id>/`; owning summaries are
`run-summary.json`, target summaries are under `target-summaries/`, row results
under `rows/`, and cleanup is recorded in `cleanup-results.json`.

| Run ID | Exact Make command/selection, before the common final argument | Passed work units | Completed cleanup steps |
| --- | --- | --- | --- |
| `reference-binding-final-reference` | Ten-row Reference Data command above | 8 | 21 |
| `reference-binding-final-server` | `make test-slice OWNER=app.server ROWS=app.server.integration.reference_pack_retained_admission` | 3 | 14 |
| `reference-binding-final-operator` | `make test-slice OWNER=app.operator ROWS=app.operator.integration.reference_pack_retained_admission` | 3 | 14 |
| `reference-binding-final-reporting` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.entra_snapshot_binding` | 3 | 14 |
| `reference-binding-final-portability` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.entra_bundle_round_trip` | 3 | 14 |
| `reference-binding-final-recovery` | `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.entra_recovery` | 3 | 14 |
| `reference-binding-final-go` | `make lint CHECK=go` | 6 | 15 |
| `reference-binding-final-boundary` | `make lint CHECK=backend-boundary` | 4 | 13 |
| `reference-binding-final-schema` | `make source-check CHECK=schema` | 3 | 12 |
| `reference-binding-final-generated` | `make source-check CHECK=generated` | 5 | 14 |
| `reference-binding-final-gosec` | `make go-gosec-targeted` | 4 | 13 |

Reference Data adds the newly extended integrity/interruption cases, closed
admission, consumer and profile contracts, fixture closure, independent Entra
oracle, full lifecycle, real signed import and both ATT&CK/Windows lifecycle
regressions. Server/operator cover the pre-mutation application boundary.
Reporting, portability and recovery exercise the changed readers through their
existing owner paths. The previously cached admission unit
`go:d3970936c8381682:d9af86728dfb` executed freshly in this final run with outer
unit exit 0; its prior cached result remains historical evidence.

The final documentation command is
`CARTULARY_TEST_RUN_ID=reference-binding-final-markdown make lint-markdown`.
Its owning artifact is
`.cartulary/test-results/reference-binding-final-markdown/adhoc/lint-markdown/tool-run-summary.json`;
the outer result is reported with completion. Documentation edits do not alter
the frozen executable source identity. `git diff --check` also passed.

There were no failed validation invocations in this follow-up. Earlier failures
listed in this handoff remain retained and are not erased by these passes.
The 36 immutable files were compared again against
`.cartulary/reference-packs/remediation/immutable-before.json`; none changed.
Source pins, quarantine decisions, editorial inputs, the 139-entry corpus,
notices, distribution decision and signed release bytes remain unchanged.

| Completion claim | Follow-up evidence and limit |
| --- | --- |
| Identity binding remediated | Shared bound decoding and valid-B-under-A rejection pass at the selected startup, consumer, pinning, Reporting and recovery boundaries. |
| Pinned-state interruption covered | Ten claimed/unclaimed scenarios pass with exact retained rows, bytes, historical provenance and subsequent readiness checked. |
| Fresh fixture/runtime verification | The selected final portfolio passes with execution-result reuse disabled and zero cache hits. This is targeted coverage, not full-owner or repository-wide acceptance. |
| Real release importable under configured trust | The existing immutable release passes the production import scenario with explicit isolated trust and freshness. No deployment trust is installed. |
| Distribution policy | The approved factual/original-text policy remains applied to unchanged content; no new distribution approval is inferred. |
| Reporting/portability/recovery | Their selected final-source owner scenarios pass. |
| Analyst workflow | Not claimed; new analyst-facing enrichment remains excluded. |

No required check for this follow-up was skipped. Full CI/full-owner audits,
the migration owner, standalone producer acquisition/check/audit/reproduction,
and producer-only ATT&CK/Windows rows were not repeated: this change concerns
stored-set reads and interruption tests, with no migration or producer-input
change. Their earlier evidence retains its original source identity and is not
claimed as fresh execution here. Release reissuance, environment reset and trust
installation were neither needed nor performed. Finalization again skipped
retained warm-check maintenance because no compatible full warm-check
`RESULTS_DIR` was supplied.
