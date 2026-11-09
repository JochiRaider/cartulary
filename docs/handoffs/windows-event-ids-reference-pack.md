# Windows Event IDs V2 reference pack

The active delivery is the user-approved CSV/OOXML V2 cutover. The older online
V1 execution record below is historical evidence, not current acceptance or an
active contract. Original inputs and earlier locks/releases remain untouched.

## Created real release: 2026.11.07.1

The actual Windows reference pack is now assembled, signed and verified. This
section supersedes earlier statements below that only a candidate and fixtures
exist or that publisher inputs are missing. The user authorized creation of this
local release and its independent publisher authority.

- Pack key: `enrichment.windows_event_ids`; version: `2026.11.07.1`; release sequence: 1.
- Publisher: `cartulary.windows_event_ids`, distinct from Windows and ATT&CK fixture authorities.
- [Signed reference pack](/home/jochi/code/cartulary/fixtures/windows-event-ids/2026.11.07.1/container.pack).
- [Public trust bootstrap](/home/jochi/code/cartulary/fixtures/windows-event-ids/2026.11.07.1/bootstrap.json).
- [Release manifest](/home/jochi/code/cartulary/fixtures/windows-event-ids/2026.11.07.1/manifest.json).

The distributable release is in its own family/version directory,
`fixtures/windows-event-ids/2026.11.07.1/`, following the existing fixture layout
for end-to-end testing and workflow implementation. These are the full real release
artifacts, not synthetic substitutes. Their signed bytes, publisher identity,
version, hashes and expiry are unchanged. `.gitattributes` preserves exact pack
and bootstrap bytes across checkouts. The pack, public bootstrap and exact signed-manifest sidecar are visible to Git and have
ordinary public-fixture permissions (`0644`); no private key is included.

The original canonical envelope directory was relocated from ignored `dist/` to
`.cartulary/reference-packs/windows-v2/releases/retained-envelopes/b866d7b97128526cf7a431b4225e63e10c94bd9861c82d7409f5158009487b47/`.
It preserves `container.pack`, `bootstrap.json` and `envelope-report.json` together
for the existing renewal interface. Use that directory as `previous_envelope`
when renewing this release. The versioned fixture pack and bootstrap are byte-identical copies
of this retained envelope; `manifest.json` is the exact manifest also inside the pack. Producer logical outputs and historical run receipts
keep their original locations. Future workflow checks must consume the pack and
bootstrap directly; this Markdown handoff is not an executable input.

The USTAR container is **958,976 bytes** and contains **503 real entries**, including
73 Sysmon identities. The 255 unresolved event/version rows remain explicitly
excluded. Provider/ID/version identity and the accepted payload bytes are unchanged.
The actual archive has exactly these seven regular members:

```text
bundle.json
manifest.json
metadata/snapshot.json
metadata/targets.json
metadata/timestamp.json
notices/Apache-2.0.txt
payload/entries.ndjson
```

The pack contains factual definitions and original summaries under Apache-2.0.
It contains no CSV, Office files, Microsoft pages, Sysmon source material, research
archive, copied message templates or XML samples. The public bootstrap accompanies
the pack; it contains public keys only. The private signing context was created
at `/home/jochi/.local/share/cartulary/reference-pack-signing/windows-event-ids/v2/signing-context.json`
with mode `0600`, outside the repository and public artifacts. Preserve it for
renewal; do not include it in distribution.

### Verification and reproducibility

Both logical assembly and signed-envelope replay completed successfully and
reproduced identical manifest, payload, container and bootstrap identities. The
actual completed container passed the production archive parser, manifest/inventory
checks, Windows content validator, TUF signatures, bindings and freshness checks
using the real UTC wall clock. The independent Node witness separately confirmed
canonical bytes, USTAR framing and Ed25519 signatures. Its `freshness_verified=false`
means that independent witness does not perform clock verification; the production
verification receipt does.

The first production receipt is at
`.cartulary/reference-packs/windows-v2/releases/envelope-ledger/verification/3e6a0ca60bc8aa32f8aed7ca6e5cd021cd29251612078cc8d0a7754e1666b050/verification.json`.
It records `status=accepted`, verification at `2026-11-07T06:34:13.944918303Z`, and
trust validity until `2026-12-07T06:34:13Z` (**December 7, 2026, 1:34:13 a.m. EST**).
Root metadata expires `2027-11-07T06:34:13Z`. Replaying verification retains a new
clock receipt while preserving the signed artifact bytes.

### Actual release identities

All values in this table are SHA-256 digests of the named subject.

| Subject | SHA-256 |
| --- | --- |
| Manifest | `895d36e75685fd1f1dac7de84b63471e4778fc2f07fe17b24b0583f91489b64c` |
| Logical payload, including project notice | `2233f6b31ec2d65b71e28e8f3eaa8a3054f3ca15796351d0dec5c0ab5de103b3` |
| Entries file | `e05a31ee0af578c2db6ec4a0b5e045f3f760d174e9a7d319b6ca4ad59c4ea8b1` |
| Signed container | `b866d7b97128526cf7a431b4225e63e10c94bd9861c82d7409f5158009487b47` |
| Public bootstrap | `36ffc425499acda4ffdf9a4632eddd26f45846cd3cbaed0921ef8394bdf93f2b` |
| Actual producer executable | `0a06af3d23ebd453960fa8c84e2615f6d82a47d81b4bde6d316d03088874493d` |
| Source lock | `40a5e272b0fc2483526f02f5a7eb98707524f6ca7489919d9b88abd754c43fca` |
| Source profile | `25e26e9ac0d855184338b92dcaf943b6ae69d0ea51720bdea00774e06e2080b0` |
| Local research archive; not shipped | `cb950856be207cb79a68b130b58b44ca5dfe849806d4d3bfadb3e6d54be60009` |
| Real assembly reconciliation | `b6366b7a8b1f6f172a8543748dabab3cd891aa94806e53d661543cd202e9f1dc` |
| Recorded license decision with canonical final LF | `5bb22892493e71b2f2c3c748e6ca43c0b4e5a0049c90c6adb6777f3025ab9cca` |
| Independent logical witness | `2e22630839de22add57a36f94881e286c38d880560e6e2fe87eddaa2aab4a6c8` |
| Independent envelope witness | `7ff05075cb351878356befc16bf7ede113c7e6b2270ebe5c23bf66914c211c20` |

The producer is `cartulary.reference-pack.producer`, version `references-v2`, Go
`go1.27.3`, 15,626,015 bytes, reproduced by two build passes. Build time was explicitly
recorded as `2026-11-07T06:30:26Z`; envelope issuance was `2026-11-07T06:34:13Z`.
The pack contract/content/source profile identifiers are unchanged. No real
application pack-set identity is claimed because this release has not been
activated in a target application.

### Changes and executed commands for real packaging

Shared envelope mechanics now accept an explicit publication policy. Existing
fixture adapters keep their identities and ATT&CK restrictions. Windows releases
have their own repository, private-context schema, ledger namespace and closed
envelope input. Real envelopes pass runtime verification before publication.
This does not weaken runtime trust admission or install a bootstrap implicitly.

Changed authored paths:

- `internal/modules/reference_data/producer/envelope.go`, `windows_package.go`,
  `windows_contract.go`, and new `windows_release.go`.
- `contracts/reference-pack-producers/windows-event-ids/contracts.v2.schema.json`
  and `license.project.v2.json` (added the notice's required terminal LF).
- `tools/reference-packs/witness.mjs` and `tools/task_surface_owner.json`.
- `docs/reference-pack-subsystem-nlspec.md` and this handoff.

Task-surface projections were regenerated through Make. No tests were created.
Release inputs are retained under
`.cartulary/reference-packs/windows-v2/releases/2026.11.07.1/inputs/`;
`build.corrected.json` and `envelope.json` are the successful inputs. The earlier
`build.json` remains retained as the failed pre-correction input. Initial public build output
was under `dist/reference-packs/windows-event-ids/2026.11.07.1/`; the pack and
bootstrap now live in `fixtures/windows-event-ids/2026.11.07.1/` as linked above. Publication and
metadata ledgers remain under `.cartulary/reference-packs/windows-v2/releases/`.

All commands ran from the repository root with `CARTULARY_TEST_RUN_ID=<run ID>`.
Run roots are `.cartulary/test-results/<run ID>`. Graph summaries are
`run-summary.json`; package/build tool summaries are `<target>/tool-run-summary.json`.
The final Windows production source digest is
`sha256:fba4311ffc7371bb21a070394149fd0800a262b961024a9fceffa2c03143e7ec`.

```sh
make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/windows-v2/releases/2026.11.07.1/inputs/build.corrected.json
make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/windows-v2/releases/2026.11.07.1/inputs/envelope.json
make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.unit.enrichment_windows_event_ids_independently_valid_fc66475a5f,module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e
```

| Run ID | Command/selection | Outer exit and result |
| --- | --- | --- |
| `windows-real-release-format` | `make format` | 0 |
| `windows-real-release-generate` | `make generate` | 0 |
| `windows-real-release-focused` | Windows producer unit row only | 0 |
| `windows-real-release-builder`, `windows-real-release-final-builder`, `windows-real-release-notice-builder` | `make reference-pack-producer-build` | 0 each; last is the shipped builder |
| `windows-real-release-finalize`, `windows-real-release-notice-finalize` | `make agent-finalize` | 0 each; zero generated changes |
| `windows-real-release-production`, `windows-real-release-final-production` | Exact three-row selection above, cache off | 0 each; final run after the notice correction; all 16 cleanup operations completed |
| `attack-release-signing-regression` | Separate cache-off invocation selecting only `module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c` | 0; shared signing-mechanics regression, separate from Windows qualification |
| `windows-real-release-schema` | `make source-check CHECK=schema` | 0 |
| `windows-real-release-generated` | `make source-check CHECK=generated` | 0 |
| `windows-real-release-assemble` | Assemble with original `build.json` | 2; wrapper `harness/unknown_failure`; producer `content_semantic_invalid` because project notice lacked final LF; no release reserved or published |
| `windows-real-release-assemble-corrected`, `windows-real-release-rebuild` | Assemble command above | 0 each; identical logical identities |
| `windows-real-release-envelope`, `windows-real-release-envelope-replay` | Envelope command above | 0 each; identical signed container/bootstrap; production verification accepted |

The notice correction changes only canonical framing of the project's notice;
it does not alter event content or the user's licensing decision. Schema/generated
checks and the separate ATT&CK regression preceded that notice-only correction;
real assembly and final Windows production qualification validate the corrected
input. Finalizers skipped retained warm-check maintenance because no compatible
full check was supplied. Full `make check`, extra Reporting/Recovery, and browser/UI
suites were not run for this scoped producer/signing change. They are not claimed
as current full-owner or repository acceptance.

Final documentation validation command:
`CARTULARY_TEST_RUN_ID=windows-real-release-handoff-lint make lint-markdown`.
Its retained result is
`.cartulary/test-results/windows-real-release-handoff-lint/adhoc/lint-markdown/tool-run-summary.json`.

### Import and maintenance

To admit this locally signed pack, configure the public `bootstrap.json` through
`CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH` in the target deployment, using
that deployment's ordinary trust/clock configuration. Preserve other trusted
repositories when configuring a deployment that already has a bootstrap. Then
upload `fixtures/windows-event-ids/2026.11.07.1/container.pack` through existing Reference Pack administration and
explicitly activate version `2026.11.07.1`. No target trust configuration, import,
activation or incident snapshot was changed by this work.

Before December 7, 2026, renew online metadata using the existing `renew` action,
the same logical root and private signing context, the original root-expiry
instant, a new explicit issuance/online-expiry instant, a larger metadata version,
and the retained canonical envelope directory above as `previous_envelope`.
Renewal preserves manifest and logical
payload identities. Root expiry/rotation requires a deliberate trust-root update;
this initial release does not implement an automatic root rotation policy.

Producer, corpus, distribution artifact, local publisher authority and runtime
verification are complete for this release. Deployment still requires the target
operator to configure trust and import/activate it. Historical sources, locks,
fixtures and earlier candidates remain unchanged; no destructive reset occurred.

## Historical pre-release distribution implementation

The user's latest direction supersedes the earlier rich-content distribution
policy: ship factual definitions and original summaries only. CSV, DOCX, XLSX,
Microsoft documentation, Sysmon sources, message templates, XML examples, tables
and copied explanatory text remain local research materials. They are not pack
payloads. Historical source locks, evidence and candidates remain untouched;
older rich-content candidates below are not eligible for this distribution policy.

The producer now projects an explicit allowlist of supported facts: provider,
event ID, explicit or unspecified version, numeric native descriptors, GUID,
technical manifest identifiers, classifications, version history and recognized
applicability facts. It uses 89 authored event summaries with a factual identity
fallback for other entries. These are advisory reference descriptions, never
claims of observed activity. Identical factual blocks coalesce their citations
instead of repeating text. Trailing empty paragraphs remain stripped from public
presentation. Existing lookup operations return the complete resulting entry.

Research and distribution roots must be disjoint. A completed logical directory
contains only `manifest.json` and its declared payload/project-notice files.
Evidence archives, source profiles, licensing records, reconciliation and build
reports remain under the source root. Source hashes and locators identify retained
research without embedding it or requiring a website for consumer operation.
The existing envelope and portability paths carry the final pack's admitted
members and trust metadata.

The project user's licensing direction is recorded in
`contracts/reference-pack-producers/windows-event-ids/license.project.v2.json`:
Apache-2.0, redistribution allowed, exact frozen source-lock binding and the
project's license notice. This classification covers the factual/project-authored
output only. It does not relicense upstream research material. Real assembly
rejects substitution of an acquired upstream notice or a different decision.
Synthetic fixture licensing stays independent. No deployment or activation is
implied by the recorded license.

Changed authored paths for this follow-up:

- `internal/modules/reference_data/producer/windows_distribution.go` (new),
  `windows_reconcile.go`, `windows_package.go`, `windows_contract.go` and
  `windows_provider.go`.
- `contracts/reference-pack-producers/windows-event-ids/original-summaries.v2.json`
  and `license.project.v2.json` (new), `contracts.v2.schema.json` and
  `source-profile.v2.json` and `editorial.references.v2.json`.
- `docs/reference-pack-subsystem-nlspec.md`, this handoff, the gap CSV (now
  explicitly labeled excluded), and `tools/reference-packs/windows-office-witness.mjs`.
- Existing `windows_online_test.go` and `windows_online_integration_test.go`:
  removed assertions that required upstream templates in the public payload;
  adapted the existing synthetic notice initializer. No tests were added.

### Accepted corpus and immutable output identities

The final reconciliation is **accepted**, with **503 entries and zero issues**.
The user-approved exclusions cover **255 event/version rows (237 distinct IDs)**;
unsupported versions are skipped without discarding independently supported
identities. The gap CSV now labels every row `excluded` with the user's decision
reference and preserves the exact evidence needed for any future reconsideration.
Future sources and unresolved rows do not inherit these exclusions. A changed,
duplicate or no-longer-applicable exclusion fails review rather than silently
changing coverage.

All 448 CSV records are accounted for: 192 admitted event/version rows, 255
approved provider exclusions, and the separate non-event boundary marker 8191.
The independent local audit also accounts for all 443 worksheet message rows,
411 catalog rows and 410 Word sections. All 323 retained sources were used
without visiting their websites. Exclusion is not a claim that the provider was
resolved, and Word/Excel overlap is not required for admission.

The 503 entries cover 445 provider/event pairs, 436 distinct numeric event IDs
and 22 providers. Sysmon contributes 73 identities. The six original families
remain; EMET and Windows Error Reporting still have zero finite event IDs in
the selected references. Versions are 0:179, 1:210, 2:17, 3:11, 4:6, 5:10 and
unspecified:70. The factual payload is 937,379 bytes; its largest entry is 5,861
bytes including LF and its longest common description is 158 Unicode scalars.

Source profile remains `cartulary.windows_event_ids.microsoft_references.v2`;
Windows content profile remains `cartulary.reference_pack.enrichment.windows_event_ids.v2`.
All values below are SHA-256 digests of their separately named subjects.

| Subject | SHA-256 |
| --- | --- |
| Source lock | `40a5e272b0fc2483526f02f5a7eb98707524f6ca7489919d9b88abd754c43fca` |
| Source profile | `25e26e9ac0d855184338b92dcaf943b6ae69d0ea51720bdea00774e06e2080b0` |
| Reviewed editorial, including the 255 exclusions | `b5d7d27e61863a8115423677371e363679dee3fab86fc3981d44b08b74f1be66` |
| Authored original summaries | `156e29539134e3d754387c61a74e499372df24899a9cce4deab89c04f50a2d77` |
| Recorded project-license decision | `8167e2ce1a4b3e9060fd9cc1c1ac4c386d38f13219f9999025e38d320943d893` |
| Actual producer executable | `a53d237c2cc4137c3ea55f95bb11efd7f3d51cf74e778c223da4d5cb6a6b2135` |
| Accepted entries file | `e05a31ee0af578c2db6ec4a0b5e045f3f760d174e9a7d319b6ca4ad59c4ea8b1` |
| Local research archive; not distributable | `cb950856be207cb79a68b130b58b44ca5dfe849806d4d3bfadb3e6d54be60009` |
| Accepted reconciliation | `c775fbfaf6038cbd734ab2f40cafd624ef4da509f35ba980831f863d93d99b8b` |
| Explicit independent source audit | `a34516ca81cc7065cd89504897f39fe2a2a87251d755a40f1895ed493eb6ecb6` |

The CLI builder is `cartulary.reference-pack.producer`, version `references-v2`,
Go `go1.27.3`, 15,322,324 bytes. Its two build passes reproduced identical bytes.
The accepted entries are retained as `entries.candidate.ndjson` beneath
`.cartulary/reference-packs/windows-v2/channels/acquisition/audits/c775fbfaf6038cbd734ab2f40cafd624ef4da509f35ba980831f863d93d99b8b/`.
That directory is research retention, not a distribution directory. No real-corpus
manifest, logical payload digest, signed container, bootstrap or pack-set identity
has been produced; those require the actual publication inputs below.

### Executed verification

Every command ran from the repository root with
`CARTULARY_TEST_RUN_ID=<run ID>`. Roots are `.cartulary/test-results/<run ID>`.
The final executable-source digest is
`sha256:bb969bad93da212aa654bd87bb690eb7fdfa3aa9a38a19deb95240d59a154fc2`.
No tests were added. The final production invocation disabled execution-result
reuse with `CARTULARY_HARNESS_CACHE_MODE=off` and selected exactly:

```sh
make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.unit.enrichment_windows_event_ids_independently_valid_fc66475a5f,module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e
```

The explicit research audit used:

```sh
make reference-pack-source ACTION=audit SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.channels.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/channels/acquisition EDITORIAL=contracts/reference-pack-producers/windows-event-ids/editorial.references.v2.json
```

| Final run ID | Command | Outer exit and outcome |
| --- | --- | --- |
| `windows-distribution-exclusions-format` | `make format` | 0; passed |
| `windows-distribution-exclusions-generate` | `make generate` | 0; passed |
| `windows-distribution-exclusions-finalize` | `make agent-finalize` | 0; passed; zero generated changes |
| `windows-distribution-exclusions-builder` | `make reference-pack-producer-build` | 0; reproduced executable |
| `windows-distribution-exclusions-schema` | `make source-check CHECK=schema` | 0; passed |
| `windows-distribution-exclusions-generated` | `make source-check CHECK=generated` | 0; passed |
| `windows-distribution-exclusions-production` | Exact three-row selection above | 0; all three passed; all 16 cleanup operations completed |
| `windows-distribution-exclusions-audit` | Audit command above | 0; accepted, zero issues; independent accounting and runtime content validation passed |

Graph summaries are `run-summary.json`; tool summaries are
`<target>/tool-run-summary.json`. The source audit is a separately requested
research diagnostic, not a normal build gate. The production rows exercise input
handling, deterministic publication, verifier trust, import, explicit activation,
lookup/pagination, retained A after B, and envelope renewal.

Earlier runs in this follow-up remain separate evidence. `windows-distribution-focused`
and `windows-distribution-final-production` passed the same existing three rows
before the approved exclusions. Builder, generation, finalizer and schema/generated
checks with the corresponding `windows-distribution-` prefixes also passed.
`windows-distribution-audit` and `windows-distribution-final-audit` exited 2,
reported by the wrapper as `harness/unknown_failure`, with the producer's
`projection_conflict` explained by exactly 255 missing-provider issues. The final
approved-exclusion run supersedes those blocked candidates. Initial
`windows-distribution-format` has a passing summary but its outer exit was not
captured; `windows-distribution-format-final` completed with exit 0.

The finalizer skipped retained warm-check maintenance because no compatible full
check root was supplied. Full `make check`, additional Reporting/Recovery and
browser/UI rows were not rerun for this Windows producer/presentation change.
MITRE rows were not selected: its adapter, inputs and qualification remain
independent. The unrelated database/harness work in the workspace was preserved.
Whole-repository or full-owner qualification is not claimed.

### Retained synthetic integration artifacts

These are test-only identities from the final production run, not a real release.
They reside beneath that run's
`_shared/test-services/bd773176ca264f138e1b0ee6/` directory. Both logical trees
contain exactly `manifest.json`, `payload/entries.ndjson` and
`notices/synthetic.txt`; container inspection found only those files plus
`bundle.json` and the three trust metadata members. No research material is in
either distribution tree or any of the three containers.

| Subject | Identity |
| --- | --- |
| A manifest | `394a236cb1bb4f36fbc71284214be0c75a8cad6f0ba4be227103d130907da646` |
| B manifest | `7882ec5da145155bd0b8ca0f49accd07056b74af0beb7c6e7772bdc60b65a318` |
| A/B logical payload | `1ba8962031240c85a5e77373413317852b89662e3af82abeec19b7f2a59d841f` |
| A container | `2e428f6b81047b749cda4f921fe30c1bacc7a762697315680a45fa37472dfc8c` |
| B container | `490808987a7ceced91d078115e8b42422039533e4bbb4b21a61a96b7f9464db5` |
| Renewed B container | `ef88bbcbb9d90d320cd5a45acfd4ec9e81f0092206cc67bf7557bc0c4d6d1c9a` |
| Fixture bootstrap | `c348495889b1999e73ce925d609e626fca10fc3061c852bf36921a9b428c9697` |
| A plus three required Base registries | `rpset_0fdca40fa6dd626b79a1506f7183598eec990cf3ddf5df5ee9982638e892c954` |
| B plus three required Base registries | `rpset_c9f8832c31c8d7091e5e456b584d5d143c43a6445925bde2d2591127d7b51963` |

Hashes above were read or calculated from the retained artifacts. Set identities
were calculated from those exact fixture manifests/payloads and authored Base
bindings; the existing lifecycle test verifies the same four-member calculation
against runtime sets. A and B intentionally have different release identities;
their payload bytes are equal. Renewal changes the envelope while retaining B's
logical identity.

### Pre-release readiness and remaining inputs at that time

| Category | Verdict and closure |
| --- | --- |
| Producer | Implemented; deterministic builder/publication and existing Windows production checks pass. |
| Corpus | Accepted for the frozen scope with 255 explicit exclusions and zero reconciliation issues. The gap CSV records optional future recovery evidence, not release blockers. |
| Test trust | Windows fixture authority and ordinary verification pass; no other optional pack is required. |
| Existing-interface workflow | Selected Windows import, activation, rich lookup, retained release and renewal cases pass. Additional snapshot/portability/recovery evidence remains separately historical. |
| Distribution | Factual-only projection and project Apache-2.0 decision implemented. Original materials are retained locally and excluded from products. Real assembly still awaits publication allocation; no missing upstream-material distribution decision is imposed. |
| Deployment | Supply the actual publisher `trust_repository_id`, release version/allocation and ledger location, production signing process/key ownership, expiry/renewal policy, and target trusted bootstrap configuration. Fixture authority cannot substitute for these inputs. |
| Repository-wide qualification | Not claimed; full check and additional owners were not rerun for this scoped change. |

There are no remaining corpus identity blockers for this selected release.
Publishing/signing and real-pack application acceptance await the actual publisher
inputs above. No deployment, implicit activation, destructive reset, historical
conversion or new incident-enrichment UI/importer/report derivation occurred.

## Historical Sysmon expansion before distribution narrowing

The user-selected expansion covers Sysmon and providers documented for Security,
System, Application and Setup. The current frozen inventory is
`contracts/reference-pack-producers/windows-event-ids/inventory.channels.v2.json`;
its retained root is `.cartulary/reference-packs/windows-v2/channels/acquisition`.
It retains the previous 312 inputs unchanged and adds 11 first-party references:
eight rendered Microsoft articles plus the pinned Sysmon XML manifest, inert
publisher template and license notice. Acquisition happened once. Checking and
building do not revisit these websites.

The authored provider catalog now supplies producer names without a second
hard-coded Go inventory. New catalog adapters bind provider and ID within a
single log block, a finite provider-qualified subscription selector, or a rendered
event-list row. A provider elsewhere on the same page cannot qualify unrelated
IDs. Rich query and event-list blocks contain only the selected assertion, avoiding
unrelated events in entry presentation while retaining the complete source.
Product versions such as 6.1 and symbolic levels such as Critical do not
become event versions or numeric levels. Full explanatory text remains available
through existing rich entries and lookup operations; trailing empty paragraphs
are stripped only from canonical presentation.

Sysmon is read from Microsoft's shared manifest at commit
`6678ccf265e13ea7b849f3ff479e1da5b42cb6e3`. The matching inert template explicitly
names `Microsoft-Windows-Sysmon` and its GUID; it is never executed. The adapter
accounts for all 498 manifest assertions: 497 Windows-applicable definitions and
one explicitly excluded Linux-only definition. It preserves all explicit event
versions and documentary schema context. Thirty Sysinternals event sections add
version-unspecified identities. Historical manifest IDs 61441 and 61442 remain
separate from later IDs; no aliases or installation claims are invented.

The candidate contains **503 entries**, 105 more than the preceding 398-entry
candidate. Sysmon contributes **73 identities** (43 numeric-version identities,
30 unspecified-version identities). The six original provider families remain.
The complete per-provider coverage is recorded in the current reconciliation.
EMET and Windows Error Reporting are named by retained selectors, but those
selectors do not enumerate event IDs; their explicit coverage remains zero.

One authoritative source is sufficient. Word definitions need no Excel or online
confirmation. Reviewed provider associations no longer require title, message
or field-label corroboration from another source. The gap CSV has Excel message
locators for **all 255 rows**; 28 lack Word sections. These coverage observations
are not rejection reasons. The remaining missing fact is the canonical provider
name bound to each definition; source IDs or Security channel labels cannot
invent it. Exact rows remain in `windows-event-provider-evidence-gaps.csv`.

No tests or fixtures were added. Removed provenance-only cases include the
provider-recovery research matrix, duplicate citation/source-substitution checks,
frozen-source replay and source-accounting-only assertions. Normal Windows
assembly no longer replays the source-accounting witness. Existing production
input, security-boundary, rich-entry, lookup, packaging, trust and lifecycle cases
remain. The independent source audit is an explicitly requested diagnostic,
separate from the normal build cycle. MITRE inputs, authority, ledger and test
selection remain independent.

Changed authored paths for this follow-up:

- `docs/reference-pack-subsystem-nlspec.md`, this handoff and
  `docs/handoffs/windows-event-provider-evidence-gaps.csv`.
- `contracts/reference-pack-producers/windows-event-ids/contracts.v2.schema.json`,
  `source-profile.v2.json`, and new `inventory.channels.v2.json`.
- `internal/modules/reference_data/producer/windows_contract.go`,
  `windows_acquire.go`, `windows_extract.go`, `windows_reconcile.go`,
  `windows_provider.go`, `windows_csv.go`, and new `windows_catalog.go` and
  `windows_sysmon.go`.
- Existing `windows_csv_test.go` and `windows_online_test.go`: provenance-only
  cases removed; no new cases.
- `tools/reference-packs/witness.mjs`, `windows-witness.mjs` and
  `windows-office-witness.mjs`.

### Historical candidate identities and coverage

The candidate uses source profile
`cartulary.windows_event_ids.microsoft_references.v2` and Windows content profile
`cartulary.reference_pack.enrichment.windows_event_ids.v2`. Pack contract V2,
provider/ID/version lookup identity and incident-bundle format 7 are unchanged.
All digests below are SHA-256 of the named subject, not interchangeable identities.

| Subject | SHA-256 |
| --- | --- |
| Source lock | `40a5e272b0fc2483526f02f5a7eb98707524f6ca7489919d9b88abd754c43fca` |
| Source profile | `0eb63924252ab91925af72a6b550166bf08f1cdbd2e0394e01dae6a3c1242334` |
| Reviewed editorial | `fc8e3470f6f979337f93dec05469524ec8b24b4d76d157de441c7abab38f875f` |
| Actual producer executable | `ed3de28f3a19a40ec97b3812b7403606b1b7851e748b6de97e20d8d5d03b2c9a` |
| Candidate entries file | `b1ab148a95dc99b0991d29fbcb4177728ab8f783f2e332e73b78914bcf0bfaf0` |
| Evidence archive | `46f3e81c23fbaad333befe72388118a93cfbb363331dd354d39f140fff29c67d` |
| Reconciliation report | `13b0e9b828dba0d3bea5e3dc53e503c998641b8a419bc74318b4afbddbcb3cad` |
| Explicit independent source audit | `6adf00911b4b624359c20a583c8c8c87278479d611d4d75bb12d90bf2f650e52` |

The CLI builder is `cartulary.reference-pack.producer`, version `references-v2`,
Go `go1.27.3`, 15,208,635 bytes. Both Make build passes produced identical
executable bytes, retained under `.cartulary/reference-packs/builders/<digest>/`.
The candidate, evidence archive and reconciliation are at
`audits/<reconciliation digest>/` below the retained source root. The witness is
at `witnesses/audit-<reconciliation digest>.json`. Audit and replay reproduced
identical report, payload, archive and witness identities. Both report and witness
explicitly retain `candidate_status=blocked` / `status=blocked`.

| Provider | Entries | Numeric versions | Unspecified-version entries |
| --- | ---: | --- | ---: |
| Application Error | 1 | none | 1 |
| Application Hang | 1 | none | 1 |
| EMET | 0 | none | 0 |
| Microsoft Antimalware | 4 | none | 4 |
| Microsoft-Windows-Eventlog | 5 | 0 | 0 |
| Microsoft-Windows-GroupPolicy | 1 | none | 1 |
| Microsoft-Windows-Kernel-General | 2 | none | 2 |
| Microsoft-Windows-Kernel-Power | 1 | none | 1 |
| Microsoft-Windows-PowerShell | 1 | none | 1 |
| Microsoft-Windows-Security-Auditing | 189 | 0, 1, 2 | 0 |
| Microsoft-Windows-Servicing | 2 | none | 2 |
| Microsoft-Windows-Sysmon | 73 | 1, 2, 3, 4, 5 | 30 |
| Microsoft-Windows-TaskScheduler | 8 | none | 8 |
| Microsoft-Windows-Time-Service | 2 | none | 2 |
| Microsoft-Windows-User Profiles Service | 2 | none | 2 |
| Microsoft-Windows-WMI-Activity | 3 | 0 | 1 |
| Microsoft-Windows-WUSA | 1 | none | 1 |
| NETLOGON | 1 | none | 1 |
| PowerShellCore | 195 | 1 | 1 |
| RemoteAccess | 3 | none | 3 |
| Service Control Manager | 5 | none | 5 |
| USER32 | 1 | none | 1 |
| VSS | 2 | none | 2 |
| Windows Error Reporting | 0 | none | 0 |

There are 445 provider/event pairs, 436 distinct numeric IDs and 22 providers
with admitted entries. Version counts are 0:179, 1:210, 2:17, 3:11, 4:6, 5:10
and null:70. The maximum canonical entry line is 132,096 bytes; the maximum common
description is 256 Unicode scalars. Rich content is not truncated.

The CSV audit still accounts for 448 records, 443 worksheet message rows, 411
catalog rows and 410 Word sections; 192 event/version rows are admitted, 255 lack
provider identity, and 8191 is explicitly a non-event boundary. None of the new
sources resolves those 255 remaining provider facts. The gap CSV retains its
original source-lock identity because those exact source bytes are unchanged.

No real-corpus manifest, logical payload digest, container, bootstrap or pack-set
identity was produced. Those identities remain absent. Earlier synthetic fixture
identities below are test evidence and must not be substituted for a real release.

### Executed commands and production qualification

Every invocation used `CARTULARY_TEST_RUN_ID=<run ID>` and the repository root.
Roots are `.cartulary/test-results/<run ID>`. Tool targets publish
`<target>/tool-run-summary.json`; graph targets publish `run-summary.json` and
individual `rows/<row ID>.json` results. The final source digest is
`sha256:9766cc240b5c4a1251a79e7fc81c91f8f71f1b150ff141069ba26f7aa77dbf7b`.

The exact final production selection (with `CARTULARY_HARNESS_CACHE_MODE=off`) was:

```sh
make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.unit.enrichment_windows_event_ids_independently_valid_fc66475a5f,module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e
```

Source operations used:

```sh
make reference-pack-source ACTION=acquire SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.channels.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/channels/acquisition SOURCE_INPUT=.cartulary/reference-packs/windows-v2/local-inputs.json
make reference-pack-source ACTION=check SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.channels.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/channels/acquisition
make reference-pack-source ACTION=audit SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.channels.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/channels/acquisition EDITORIAL=contracts/reference-pack-producers/windows-event-ids/editorial.references.v2.json
```

| Run ID | Command or selection | Outer exit and result |
| --- | --- | --- |
| `windows-channels-acquire-builder`, `windows-channels-builder`, `windows-channels-final-builder`, `windows-channels-presentation-builder` | `make reference-pack-producer-build` | 0; two-pass executable reproduction on each build |
| `windows-channels-acquire` | Acquire command above | 0; 323 members, including 11 new acquisitions and 312 reused members |
| `windows-channels-format` | `make format` | 0 |
| `windows-channels-focused` | Existing Windows producer row only | 0 |
| `windows-channels-candidate` | Audit command above | 2; old independent diagnostic lacked new catalog support; producer separately reported the 255 existing provider gaps |
| `windows-channels-source-audit` | Audit command above | 2; independent diagnostic initially rejected historical hexadecimal Sysmon IDs; corrected to preserve them |
| `windows-channels-audit-corrected`, `windows-channels-final-audit` | Audit command above | 2; independent diagnostic passes; exactly 255 provider gaps, no new corpus issue |
| `windows-channels-generate` | `make generate` | 0 |
| `windows-channels-finalize`, `windows-channels-finalize-presentation` | `make agent-finalize` | 0; zero generated changes; warm-run maintenance skipped without a compatible successful full check |
| `windows-channels-schema` | `make source-check CHECK=schema` | 0 |
| `windows-channels-generated` | `make source-check CHECK=generated` | 0 |
| `windows-channels-lock-check` | Check command above | 0; retained bytes only |
| `windows-channels-final-production` | Three-row production selection above | 0; fresh execution before final presentation narrowing |
| `windows-channels-presentation-production` | Three-row production selection above | 0; all three rows passed, fresh execution after final source freeze |
| `windows-channels-presentation-audit`, `windows-channels-presentation-replay` | Audit command above | 2; exactly 255 unresolved provider identities; independent diagnostic passes and identities reproduce |
| `windows-channels-markdown` | `make lint-markdown` | 0 |

The final documentation-only check is `CARTULARY_TEST_RUN_ID=windows-channels-handoff-lint make lint-markdown`; its retained summary is
`.cartulary/test-results/windows-channels-handoff-lint/adhoc/lint-markdown/tool-run-summary.json`.

The lifecycle row exercised synthetic Windows import, explicit activation, rich
lookup, a second release, retained earlier-release consumption and ordinary trust
verification. These are reused production scenarios; no Sysmon-specific fixture
or new test was authored. The real 503-entry candidate also passed the registered
content validator during reconciliation. No live website was a test dependency.

`make check`, browser/UI, additional Reporting/Recovery rows and MITRE regression
rows were not rerun for this scoped producer expansion. Consumer contracts and
those owners were not changed by the expansion; broader historical outcomes below
remain separate and are not claimed as current whole-repository acceptance.
The workspace also contains unrelated database/harness work that was preserved.

### Historical readiness before the project-license decision

| Category | Verdict and closure |
| --- | --- |
| Producer | Implemented; reproduced builder/candidate and existing production checks pass. No new tests were created. |
| Corpus | Frozen and accounted for; exactly 255 CSV provider identities remain unresolved. One authoritative definition identifying the publisher for each row closes its gap; Word/Excel overlap or a second confirmation is not required. EMET and Windows Error Reporting have zero defined IDs in the selected sources. |
| Test trust | Existing independent Windows fixture authority and ordinary verifier pass; no MITRE authority or inputs used. |
| Existing-interface workflow | Selected Windows import/activation/lookup/retained-release scenario passes; real-corpus signed import awaits publication inputs and corpus closure. Earlier snapshot/portability/recovery coverage remains separately recorded. |
| Distribution | Missing a real `cartulary.windows_event_license_decision.v2` bound to the new exact source-lock digest. It must include `decision_ref`, `approved_by`, license expression/list version, redistribution class (`allowed`, `restricted` or `prohibited`), license-reference bindings and required notice files. Scope includes the CSV derivative, both Office originals, retained Microsoft pages, Sysmon manifest/template and notice material. Retaining a license text does not itself provide this decision. |
| Deployment | Missing the actual publisher `trust_repository_id`, release/version allocation and ledger location, production signing process/key ownership, metadata expiry/renewal policy and target trusted bootstrap configuration. The fixture signer is test-only. No activation is implicit. |
| Repository-wide qualification | Not claimed for this expansion; full `make check` and additional owners were not rerun. |

Licensing and deployment inputs are required for publishing a real signed release,
not for inspecting this local candidate. Corpus identity gaps remain a separate
blocker. No production deployment, activation, destructive reset, historical
conversion or incident-enrichment UI/report work was performed.

Earlier execution records below are historical and do not establish qualification
of the expanded candidate.

## Earlier original-source admission execution record

An event does not need to appear in both Word and Excel. The earlier provider
recovery rule incorrectly required a matching Word section as well as a worksheet
message; that conjunction has been removed from the owner, producer profile,
implementation and independent witness. Either original can independently
establish the definition. A Word-only event can establish its provider/version
through its own matching System fragment; a reviewed retained provider-qualified
sample independently supplies its explicit identity facts. No second original is required.
The referenced sample's explicit numeric version is retained, and native fields
are still restricted to their supported identities.

The CSV adapter permits either retained original rather than requiring both
artifact kinds. Actual citations and source claims still require their referenced
bytes and fragment checks; no event is rejected merely because the other original
does not list it. The gap report now explicitly distinguishes missing provider
identity from non-blocking source coverage. No new test was created. The subsequent change removed provenance-only cases
and the extra source-accounting witness from the assembly path. Existing
production-use cases remain.

Changed paths in this correction: `docs/reference-pack-subsystem-nlspec.md`,
`contracts/reference-pack-producers/windows-event-ids/source-profile.v2.json`,
`internal/modules/reference_data/producer/windows_provider.go`,
`internal/modules/reference_data/producer/windows_csv.go`,
`tools/reference-packs/windows-office-witness.mjs`, this handoff, and
`docs/handoffs/windows-event-provider-evidence-gaps.csv`.

The corrected real-corpus audit still accounts for all 448 CSV records, 443
worksheet message rows, 411 catalog rows and 410 Word event sections. It admits
192 CSV records into 398 candidate entries; 255 rows remain blocked solely by
unestablished provider identity. Their version evidence is already available.
The missing Word sections noted in the report are coverage observations, not
rejection reasons. Source bytes, lock identity and reviewed associations remain
unchanged. New builder/profile/payload/evidence identities appear in the current
candidate table below; no real release was assembled.

Execution roots are `.cartulary/test-results/<run ID>`. The existing Windows
producer selection is
`make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930`.
The frozen-corpus command is
`make reference-pack-source ACTION=audit SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.references.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/acquisition EDITORIAL=contracts/reference-pack-producers/windows-event-ids/editorial.references.v2.json`.
All commands use `CARTULARY_TEST_RUN_ID=<run ID>` from the repository root.

| Run ID | Command / selection | Outer exit / result |
| --- | --- | --- |
| `windows-independent-originals-format` | `make format` | 0 |
| `windows-independent-originals-build` | `make reference-pack-producer-build` | 0 |
| `windows-independent-originals-focused` | Existing Windows producer selection above | 0 |
| `windows-independent-originals-audit` | Frozen-corpus command above | 2; exactly 255 provider blockers, independent witness passes |
| `windows-independent-originals-finalize` | `make agent-finalize` | 0; generated changes 0; retained warm-run maintenance skipped without a compatible successful full check |
| `windows-independent-originals-qualified` | Existing Windows producer selection above | 0; compatible passing result reused by Harness |
| `windows-independent-originals-audit-replay` | Frozen-corpus command above | 2; same report, payload, evidence and independent witness identities |

Finalizer source digest: `sha256:4876836297b44d70f55caac0e3a68b514d213ad8af1b365e9e9775510a20cc5d`.

The public `make task-guide OWNER=module.reference_data DETAIL=rows` and
`make help ALL=1` discovery commands also returned 0. No additional tests or
fixtures were authored. Service-backed lifecycle, other optional packs and full
repository verification were not selected for this focused admission correction;
prior outcomes remain separately recorded below.

## Earlier provider evidence recovery follow-up

The retained-source review resolved **2 of the 257 blocked CSV rows**. The
candidate now admits 192 of 447 event/version records; **255 remain blocked**.
The 8191 marker retains its non-event disposition. The candidate remains 398
entries across all six families: these two definitions were already present from
retained HTML, and now also carry their full CSV/Office documentation. No source
was downloaded, refreshed, modified or discarded.

| CSV data record | Event/version | Recovered provider | Independent original correspondence | Retained provider evidence |
| --- | --- | --- | --- | --- |
| 351 | 5155 / 0 | Microsoft-Windows-Security-Auditing | Complete Event Messages row 347; Security Audit Events row 328; Word body children 8405–8430 | `discovered-74cf37b71819ef1f`, `html/pre/0/text()` |
| 357 | 5159 / 0 | Microsoft-Windows-Security-Auditing | Complete Event Messages row 353; Security Audit Events row 332; Word body children 8683–8709 | `discovered-b2a021911e194f41`, `html/pre/0/text()` |

Both retained samples explicitly name GUID
`54849625-5478-4994-a5ba-3e3b0328c30d`, the matching event ID and version 0.
The original Word heading and complete worksheet message agree with the retained
definition. All eight ordered message labels agree: Process ID, Application
Name, Source Address, Source Port, Protocol, Filter Run-Time ID, Layer Name and
Layer Run-Time ID. Provider recovery does not copy native descriptors onto the
CSV extraction record; the existing matching HTML assertion remains their source.

The authored `editorial.references.v2.json` binds exact CSV fragments, full HTML
artifacts and XML sample fragments. Its exact hashes and locators are retained
inside the evidence archive alongside the original worksheet/Word assertion
index. The active editorial contract requires `provider_bindings`; synthetic
inputs explicitly use an empty array. Real recovery is selected explicitly with
`EDITORIAL=contracts/reference-pack-producers/windows-event-ids/editorial.references.v2.json`.
An empty/default editorial input does not implicitly import these associations.

### Exact remaining evidence gaps

[The complete 255-row evidence-gap report](windows-event-provider-evidence-gaps.csv)
records each CSV record number, event ID, version, exact missing evidence,
original worksheet row locators and hashes, Word section ranges, existing sample
IDs, retained page URLs/hashes, CSV fragment hash and immutable source-lock hash.
The report's record set was checked against the final reconciliation issues.
It is a handoff artifact, never an executable producer input or test oracle.

The remaining rows cover 237 distinct IDs, with versions 0:237, 1:15 and 2:3.
Their missing identity fact is an explicit canonical provider name bound to the
original definition. Existing worksheet evidence already supports the numeric
versions. A missing optional GUID does not itself block admission.

- 225 rows have original Word sections without a matching provider-qualified
  System fragment.
- 28 rows have no original Word event section.
- 4866/0 and 4867/0 have original and retained HTML XML naming **4865**. Those
  samples cannot establish the other events' providers. Direct inspection of
  `word/media/image160.png` and `word/media/image161.png` found only truncated
  friendly provider display labels; neither image exposes a canonical Provider
  Name, GUID or numeric event version. Image SHA-256 values are respectively
  `05c810cb64f621f886b8f3024139b205b0e86e4534d11670d9dc18bb4fd30509`
  and `40faa27b9db4555edd4cfb94444c8584f1d2659c08b385300de5ee569e695b8d`.

Among these unresolved rows, 189 have no retained event-document page; 66 map to
65 retained pages. Of those pages, 63 contain no canonical provider name or
provider-qualified XML, and the remaining two are the mismatched 4866/4867 pages.
The original workbook contains classifications and message/version cells, not
publisher bindings. The retained audit indexes and monitoring tables do not
supply an explicit provider binding for these missing definitions. Security
channel/category labels, coincident IDs and truncated provider labels remain
insufficient. Closing a row needs an independently reviewable provider-qualified
original definition or corroborating retained evidence meeting the owner rule;
there is no blanket provider assignment.

### Change rationale and scope

| Gap | Remediation / areas | Benefit and compatibility | Risk without fix / validation |
| --- | --- | --- | --- |
| Two corroborated definitions could not contribute rich CSV documentation | Closed reviewed associations, original message/title/field checks, immutable assertion links; specification, producer contracts, implementation, independent witness and Windows tests | Explicit evidence chain without ID-only inference; runtime Windows V2 and lookup interfaces unchanged; editorial authoring now requires a bounded recovery list | False provider assignment or lost documentation; exact ID/version, full original correspondence and fragment hashes must agree |
| Remaining provider blockers were difficult to review individually | Complete row-level evidence-gap report and more specific producer issue text; documentation and implementation | Missing identity facts and closure conditions are explicit; no fabricated provider, version or GUID | Hidden source losses or misleading coverage; report rows exactly equal blocking reconciliation rows |
| A Windows service fixture constructed the earlier editorial shape | Load the authored empty Windows editorial input; fixture support | One maintained Windows fixture input; no dependency on another optional pack | Service-backed assembly rejects a null recovery array; affected Windows lifecycle scenario must pass |

Follow-up paths changed:

- `contracts/reference-pack-producers/windows-event-ids/`: `contracts.v2.schema.json`,
  `editorial.v2.json`, new `editorial.references.v2.json`, `source-profile.v2.json`.
- `internal/modules/reference_data/producer/`: `windows_contract.go`,
  `windows_csv.go`, `windows_csv_test.go`, new `windows_provider.go`,
  `windows_reconcile.go`.
- `internal/modules/reference_data/testsupport/windows_fixture.go`.
- `tools/reference-packs/windows-office-witness.mjs`.
- `docs/reference-pack-subsystem-nlspec.md`, this handoff and
  `docs/handoffs/windows-event-provider-evidence-gaps.csv`.

The surrounding worktree already contained the earlier Windows implementation.
This follow-up does not attribute those earlier changes or test outcomes to the
provider recovery. Generated projections were refreshed through Make; finalizer
inspection found no additional generated changes.

### Follow-up execution evidence

All run roots below are `.cartulary/test-results/<run ID>`; each command uses
`CARTULARY_TEST_RUN_ID=<run ID>` from the repository root. Exact command selections
and source digests are retained in the corresponding run manifests.

| Run ID | Command / selection | Outer exit / result |
| --- | --- | --- |
| `windows-provider-format-01`, `windows-provider-format-02`, `windows-provider-format-03` | `make format` | 0 |
| `windows-provider-generate-01` | `make generate` | 0 |
| `windows-provider-unit-01`, `windows-provider-unit-02` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930` | 0; synthetic recovery and rejection cases pass |
| `windows-provider-build-01`, `windows-provider-build-02`, `windows-provider-build-final`, `windows-provider-build-replay` | `make reference-pack-producer-build` | 0; final/replay executable identity agrees |
| `windows-provider-audit-01` | Source audit command below | 2; page heading was outside the body content block; producer refused recovery and independent witness detected absent corroboration |
| `windows-provider-audit-02`, `windows-provider-audit-final`, `windows-provider-audit-replay` | Source audit command below | 2; exactly 255 remaining provider blockers; independent witness passes; final/replay report, payload and evidence identities agree |
| `windows-provider-finalize-01` | `make agent-finalize` | 0; generated changes 0; warm-run maintenance skipped without a compatible successful full check |
| `windows-provider-unit-final` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.unit.enrichment_windows_event_ids_independently_valid_fc66475a5f,module.reference_data.unit.canonical_consumers` | 0; three selected rows |
| `windows-provider-generated-final` | `make source-check CHECK=generated` | 0 |
| `windows-provider-schema-final` | `make source-check CHECK=schema` | 0 |
| `windows-provider-lifecycle-final` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e` | 2; fixture supplied null instead of required empty provider-binding array; diagnosed and repaired |
| `windows-provider-lifecycle-repair` | Same Windows lifecycle selection | 0; authored empty Windows editorial input accepted |
| `windows-provider-unit-order` | Same Windows producer unit selection | 0; explicit field-order rejection case |

`windows-provider-finalize-02` passed with no generated changes and froze
`sha256:b012ac11d52f80467415dfe90fbfaa428013c27668bac29649e9fe13d7c4e74b`.
The final qualification manifests bind this same source digest.

| Run ID | Command after the run-ID assignment | Outer exit |
| --- | --- | ---: |
| `windows-provider-finalize-02` | `make agent-finalize` | 0 |
| `windows-provider-qualified` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.unit.enrichment_windows_event_ids_independently_valid_fc66475a5f,module.reference_data.unit.canonical_consumers,module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e` | 0; all four rows passed |
| `windows-provider-schema-qualified` | `make source-check CHECK=schema` | 0 |
| `windows-provider-generated-qualified` | `make source-check CHECK=generated` | 0 |

The fixture repair is limited to the Windows test-support input. The final CLI
builder remains `175fc72cc9a91a80c47751c9f58bc342f7c3100a7cefa4c3317a046012da9fd7`.
The final selected service scenario exercises the producer, ordinary verifier,
staged import, activation, rich lookup/pagination, exact four-member sets, retained
A after B, and renewal. Other-owner portability/recovery scenarios and the full
`make check` were not rerun for this producer-only follow-up; their prior results
and the existing unrelated database failures remain below. MITRE tests were not
selected: no MITRE policy or shared packaging mechanics changed. Finalizer
retained-run maintenance remains skipped because no compatible successful full
check was supplied. Diagnostic `explain-run` calls first returned usage errors
without a target and with the internal `backend-integration` projection; the
correct public `TARGET=test-slice DETAIL=logs` call returned 0 and identified the
fixture failure above. Diagnostic success is not passing test evidence.

The final synthetic service artifacts are retained under
`.cartulary/test-results/windows-provider-qualified/_shared/test-services/feb20aecc303418d83f36aed/`.
Their actual test-executable builder digest is
`d46084ace7797c04038642aad6fd053b3e2e957b5e7418406d647c7650e9ba01`.
Both use the synthetic logical payload digest
`ee86057988143cfa2f53b13dbaf9ed78afb4288f4068d12cad25f21fdab9d9a7`,
payload-file digest `aae6df8120f981b8cd904f2fc942642ae55731eb260a98074218fc3346621f80`
and evidence archive `b968108b83389da19a55ad4dc8fa89d6bdfcfa3c6998bb14de91e9cc6ead2ccf`.
The exact set IDs below are reconstructed from the retained bindings using the
same independent expectation that the passing application scenario compares to
its returned set IDs. Each includes all three required Base registries.

| Synthetic release | Manifest SHA-256 | Exact pack-set ID |
| --- | --- | --- |
| A | `fe8d2c81f606c16d20ad6ce619de8b2e78845df16fcf1a6647ebebc3adf9bca7` | `rpset_2a9f32bd1659bb7eef69b968c78b935a5f3ab2051f111227de14d13710043e3d` |
| B | `c1bce2664e812fd87468e5a75ae1a055de65989e23c580a61dd3b05f04f98cc3` | `rpset_5fc3ad604631c18512f0b68ed82f1e5a28eab4d870b65889ccde26bedba487ab` |

The retained A container is
`74a4b7b4c2faa4f84e2a0ca7f6010908bbc070d3ca8436c4c5b43cbd85830030`.
The two B envelopes (including renewal) are
`8bef51ef28eff1d7acd33cb03b9a7c43bcbcd6bb06a6e9d801a48c4f55609d7d` and
`a9e3c13b5b4ec400f7d8fdad9a117c79cfbfe5ff8136942d5e84d184de64d51a`;
they preserve the same logical B manifest. All three use bootstrap digest
`1b707c1a181f3b058b8b02f1cc0f8ba380da1495f668336f5964fd88f9504579`.
These identities establish synthetic test trust only.

The first audit issue was fixed by checking the page's unique heading while
preserving body-fragment locators. A split heading/body synthetic fixture now
exercises this layout. Negative tests reject wrong IDs/versions, missing or
misordered fields, unrelated titles, unknown providers, substituted bytes and
stale/duplicate associations. The independent witness reads retained CSV, OOXML
and HTML without calling producer extraction/projection code.

The exact real-corpus audit command is:

```sh
CARTULARY_TEST_RUN_ID=<fresh-run-id> make reference-pack-source ACTION=audit SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.references.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/acquisition EDITORIAL=contracts/reference-pack-producers/windows-event-ids/editorial.references.v2.json
```

The audit's nonzero outer exit is intentional while unresolved identity issues
remain; passing source accounting is not publication acceptance. No real-corpus
logical payload digest, manifest, container, bootstrap or pack-set identity was
produced. Current candidate identities are recorded below. Licensing and actual
publisher/deployment inputs remain absent. Prior synthetic workflow outcomes
remain identified separately; no deployment, activation, conversion or reset was
performed during this recovery.

## Current scope and phase gates

Windows content/entry profiles are V2; the source profile is
`cartulary.windows_event_ids.microsoft_references.v2`. Pack contract V2,
Reference Pack major 3, incident-bundle format 7, provider/ID/nullable-version
identity and the existing public operations remain unchanged. Windows V1 is
rejected rather than read, converted, dual-written or automatically reset.
Other packs retain independent profiles, tests, sources, trust and ledgers.

Canonical documentation removes trailing Unicode White_Space from text leaves,
including nested table cells. Leading indentation, internal paragraphs and
message placeholders remain intact. Exact CSV and Office bytes, raw extraction
records and fragment hashes retain the original whitespace. No UI feature or
incident enrichment derivation was added; existing entry and lookup operations
return the complete V2 documentation.

| Phase / gap | Remediation and areas | Dependency and rationale / long-term benefit | Compatibility and risk if unresolved | Exit / current evidence |
| --- | --- | --- | --- | --- |
| 1: active Windows V1 cannot carry rich reference details | Adopt closed typed documentation blocks and V2 registry; specification, contracts, implementation, fixtures | Owner inspection; reusable evidence kinds and scopes avoid a CSV-shaped runtime contract | Intentional pre-production Windows V1 rejection; no conversion/reset; otherwise silent content loss and legacy burden | Current registry, content validator and fixtures agree; V1 rejection is tested |
| 2: derivative CSV citations need original evidence | Immutable local admission and bounded ZIP/XML parsing; producer contracts, implementation, tests | Phase 1; exact bytes and source claims are separate from local admission time | New lock only; no website availability dependency; otherwise unverifiable extraction or unsafe Office parsing | 312 retained inventory members; 448 CSV records, 443 messages, 411 catalog rows and 410 Word sections independently accounted |
| 3: incomplete providers and disputed source assertions | Recover original worksheet/body locators; qualify samples and descriptive disputes; producer and independent witness | Phase 2; canonical native fields require matching original System evidence; no invented provider/version | 255 rows remain blocking; 8191 is excluded as a boundary marker; otherwise false identities or version leakage | 398 candidate identities across all six families; 192 admitted CSV rows; publication remains blocked |
| 3: trailing empty paragraphs waste presentation space | Trim only canonical documentation leaves; owner, producer, consumer validation, independent witness, tests | Raw evidence remains immutable; avoids whitespace-only tails without collapsing meaningful layout | New V2 output bytes; originals unchanged; leaving the gap retains unnecessary UI space | Published message and nested text checks preserve indentation and internal paragraphs |
| 4: pack-specific policy must remain independent | V2 Windows fixture authority and ledger namespace over shared neutral mechanics; implementation and tests | Accepted synthetic corpus; isolated policy prevents cross-pack trust or selection coupling | ATT&CK keeps its own adapter and separate regression invocation | Windows lifecycle and separate ATT&CK regression evidence recorded below |
| 5: release claims need real approval inputs | Separate producer/corpus/test-trust/workflow/distribution/deployment verdicts; handoff | Applicable preceding gates; maintainers can update by new immutable acquisition/release | No implicit activation or publication; otherwise fixture success is mistaken for approval | Exact remaining inputs and actual produced identities recorded below |

## Supplied evidence and source accounting

All sources are frozen, one-time references. No test depends on a live website,
its continuing availability, or its future contents. The missing supplementary
`research/windows_event_csv/docx/document.txt` is not required: original OOXML
part/body-child and worksheet-row locators are used.

| Input | Exact bytes | SHA-256 |
| --- | ---: | --- |
| Combined CSV | 9,780,933 | `a1bc89260b3d37c9f6010d0aff04c012829ef67ad638e03d1b4f1ab0367b4206` |
| XLSX | 70,075 | `67550232363e5321ad12339a1572fe41f860bdfb4539faed99714a5148d0094b` |
| DOCX | 9,859,839 | `73c896e28dcc7adc33ee2712a41cb541bc44c36c867d103a8d6486680baa9ae5` |

CSV admission retains 448 records and all 44 columns: 447 event/version rows
covering 414 event IDs, plus the 8191 boundary marker. Version counts are
0:414, 1:28 and 2:5. The original 443 message rows include the boundary marker;
411 catalog rows and 410 non-TOC Word event sections are accounted independently.
Repeated original assertions have a single digest-bound locator with associated
CSV record numbers. Recovered associations do not rewrite the supplied CSV.

The 4866/4867 samples naming 4865 remain mismatched context. Declared/sample
version differences for 5632/5140/5156/5157 do not transfer native descriptors.
The 4668/4817 descriptive disagreements remain qualified source claims. Fourteen
unique partially parsed samples are inspected through their retained System
fragments; their malformed payload text is not a canonical payload schema.
Sample users, hosts and times are documentation, never incident observations.

### Complete column destination map

Every column is retained verbatim in the archived CSV. The table identifies its
additional canonical or provenance destination. Empty values are not fabricated.
Flattened summaries remain provenance when structured associations preserve the
same content more accurately.

| CSV columns | Canonical / provenance destination |
| --- | --- |
| `record_kind` | Admission disposition; 8191 is explicitly a non-event |
| `event_id`, `event_version` | Canonical identity after independent original-source verification |
| `version_evidence`, `source_coverage`, `docx_context_scope` | Typed extraction context, qualified as source context |
| `provider_names_from_docx_examples`, `provider_guids_from_docx_examples` | Typed extraction context; canonical provider/GUID facts are independently recovered from matching original samples |
| `xlsx_message_summaries`, `docx_event_titles` | Event-level title variants |
| `xlsx_audit_categories`, `xlsx_audit_subcategories`, `docx_audit_categories`, `docx_audit_subcategories`, `docx_audit_outcomes` | Original flattened summaries in CSV provenance; structured classifications retain the source associations below |
| `xlsx_message_template`, `docx_message_template` | Separate message blocks, respectively event/version and event scope |
| `xlsx_minimum_os_for_message_version`, `docx_minimum_os_as_documented`, `docx_required_server_roles` | Separate applicability blocks with version/event scope |
| `docx_event_description`, `docx_other_event_context` | Event-level explanations |
| `docx_version_history` | Version-history block; declarations do not imply version-specific field inheritance |
| `docx_event_field_reference`, `docx_field_definitions_json` | Field reference and structured definitions, annotations, relationships and table indices |
| `docx_security_monitoring_recommendations`, `docx_subcategory_guidance_json`, `docx_shared_monitoring_guidance` | Distinct event, subcategory and shared monitoring scopes |
| `docx_documented_versions`, `docx_example_versions_for_this_event` | Original flattened version summaries in provenance; structured declarations/examples remain below |
| `docx_sample_metadata_for_row_version_json`, `docx_other_sample_metadata_json` | Qualified examples, with matching version or source-context scope |
| `xlsx_catalog_rows_json`, `docx_section_contexts_json` | Source-specific classifications; Word contexts also support version history and original section accounting |
| `docx_reference_tables_json` | Typed tables with cell text, spans, merge metadata and owner relationships |
| `docx_links_json` | Source links; never fetched or executed during checking/building |
| `docx_untranscribed_image_count`, `docx_image_references_json` | Original count in provenance; typed untranscribed-image references and retained image members |
| `quality_flags`, `quality_notes_json` | Qualified quality-issue blocks, including source disagreements |
| `source_filenames`, `source_download_urls` | Typed extraction context; publication/download claims are not invented retrieval times |
| `source_locators_json` | Reviewed original associations and digest-bound assertion index; supplementary text line numbers remain provenance |
| `source_artifact_metadata_json` | Typed source metadata plus exact retained originals and hashes |

## V2 qualification and release evidence

Execution roots are `.cartulary/test-results/<run ID>`. Each command below used
`CARTULARY_TEST_RUN_ID=<run ID>` from the repository root. Outer Make exits are
recorded separately from inner summaries. No selected or skipped check is
claimed as passing.

The first finalized implementation source digest was
`sha256:b67ead0b8ef496cc93e903f49b67f1e2b5c038486badcec1ff01fd40203493a9`.
`windows-v2-finalize-01` passed with no generated changes; retained warm-run
maintenance was skipped because no compatible successful full check was supplied.
The broad check completed before the shared fixture correction below. The
correction changes test vectors and generated fixture projection only, with no
producer or application behavior change.

### Executed source and packaging commands

| Run ID | Command after the run-ID assignment | Outer exit / result |
| --- | --- | --- |
| `windows-v2-acquire-01` | `make reference-pack-source ACTION=acquire SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.references.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/acquisition SOURCE_INPUT=.cartulary/reference-packs/windows-v2/local-inputs.json` | 0; 312 retained sources |
| `windows-v2-build-final` | `make reference-pack-producer-build` | 0; twice-built executable identity recorded below |
| `windows-v2-source-check-final` | `make reference-pack-source ACTION=check SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.references.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/acquisition` | 0; retained digests valid |
| `windows-v2-audit-final` | `make reference-pack-source ACTION=audit SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.references.v2.json SOURCE_DIR=.cartulary/reference-packs/windows-v2/acquisition` | 2; producer reports 257 provider blockers; independent witness passed |
| `windows-v2-assemble-final` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/windows-v2/qualification/build-input.json` | 0; synthetic logical witness passed |
| `windows-v2-envelope-final` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/windows-v2/qualification/envelope-input.json` | 0; synthetic signature/container witness passed |
| `windows-v2-renew-final` | `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/windows-v2/qualification/renew-input.json` | 0; logical identity preserved |
| `windows-v2-rebuild-final` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/windows-v2/qualification/build-input.json` | 2; setup incorrectly reused one directory for publication and envelope ledgers |
| `windows-v2-rebuild-isolated` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/windows-v2/qualification/rebuild-input.json` | 0; separate publication ledger and output root reproduced identical logical bytes |
| `windows-v2-replay-final` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/windows-v2/qualification/rebuild-input.json` | 0; immutable same-root replay passed |

The mixed-ledger setup was preserved for diagnostics, not silently repaired. The
rebuild input uses a fresh publication ledger and preserves all content-affecting
inputs. Private signing context bytes are outside public release and evidence
roots and are not included in this handoff.

### Focused verification and diagnosed failures

Producer/profile validation passed in `windows-v2-producer-02`,
`windows-v2-producer-03`, `windows-v2-producer-profile-04`,
`windows-v2-profile-01`, `windows-v2-presentation-01` and
`windows-v2-sample-01` (outer exits 0). These runs cover the Windows producer row,
Windows profile row and selected common consumer/profile rows. Exact selections
and commands are retained in each run manifest. The presentation checks prove
that published values lose trailing empty paragraphs while raw evidence,
indentation, internal paragraphs and placeholders remain intact.

Initial compile failures (`windows-v2-build-01`, `windows-v2-build-03`) were an
unused variable and missing import, both repaired. The first producer run exposed
old description-limit expectations after rich content moved into documentation;
they were corrected to exercise the rich-field bounds. The first mixed lifecycle
selection (`windows-v2-lifecycle-01`) passed the produced Windows lifecycle but
failed the generic consumer's hard-coded V1 manifest metadata; the registry-based
helper repair passed in `windows-v2-lifecycle-02`.

The first Go lint (`windows-v2-static-01`) failed on two capitalized error strings;
`windows-v2-static-02` passed after correction. `windows-v2-schema-01` passed.
`make format` passed in `windows-v2-format-01`, `-02` and `-03`. Authored projections
were refreshed with successful `make generate` runs `windows-v2-generate-02`,
`-03`, `-04` and `windows-v2-generate-repair`; the earlier auto-ID generation root
is `20261107T023417Z-p58522`.

Exploratory audits `windows-v2-audit-01` through `-06` returned outer exit 2 while
closed CSV nullability and independent XML/table/heading handling were corrected.
Audits `-07`, `-08` and `windows-v2-audit-final` completed the independent source
witness; their outer exit 2 reflects the explicitly blocked corpus. The final two
produce identical report and witness hashes. None is reported as an accepted
real corpus or a successful release build.

`CARTULARY_TEST_RUN_ID=windows-v2-check-final make check` returned outer exit 2:
335 of 338 work units passed, three failed, none skipped. Its summary is
`.cartulary/test-results/windows-v2-check-final/run-summary.json`. The failures
separate as follows:

- The V2 cutover left stale complete-set expectations and V1 assumptions in two
  generic portability boundary fixtures. Their independent Python encoder/hash
  reproduced the old vectors before calculating the V2 expectations. The fixture
  constructors now use registered profile metadata; active complete-set and
  expected hashes were updated and regenerated. The exact affected rows
  `module.reference_data.unit.canonical_fixture_manifests` and
  `module.reference_data.unit.canonical_pack_manifest` passed with outer exit 0 in
  `windows-v2-fixture-repair`.
- Database migration assertions reproduced the previously recorded failures:
  rejected downgrade leaves head 47, and manifest coverage misses
  `import_committed_source_records_source_fk`. Owning rows are
  `module.database_migrations.integration.production_preflight_state_matrix` and
  `module.database_migrations.integration.production_ddl_v3_recurrence`.
- The pre-existing PostgreSQL ACL assertion still expects DELETE permission for
  `public.import_committed_source_records` but observes false. Owning row:
  `platform.postgres.integration.purpose_identity_roles_and_acl`.

Those database/ACL assertions require owner corrections and reruns before a green
repository gate can be claimed. Their schema/permissions were not changed by this
pack work. The prior historical Mermaid timeout did not recur in this check.


### Changed paths and maintenance procedure

The V2 slice changes the following authored owners and implementations; generated
projections were refreshed only through Make:

- `docs/reference-pack-subsystem-nlspec.md` and this handoff.
- `contracts/reference-packs/profiles.v2.json`, `consumer_item.v2.schema.json`,
  and `profiles/enrichment.windows_event_ids.entry.v2.schema.json`.
- `contracts/reference-pack-producers/windows-event-ids/`: closed V2 contracts,
  source profile, seeds, frozen inventory, editorial inputs and synthetic fixtures.
- `internal/modules/reference_data/producer/windows_contract.go`,
  `windows_local.go`, `windows_acquire.go`, `windows_ooxml.go`, `windows_csv.go`,
  `windows_csv_documentation.go`, `windows_extract.go`, `windows_reconcile.go`,
  `windows_package.go`, `windows_csv_test.go` and `windows_online_test.go`.
- `internal/modules/reference_data/producer/envelope.go` and `identity.go`:
  explicit Windows V2 fixture policy and truthful executable identities.
- `internal/modules/reference_data/internal/packformat/content.go` and
  `content_test.go`: nested rich-reference admission and presentation rules.
  `portability_retention_test.go`, `portable_content_test.go`,
  `contracts/reference-pack-fixtures/fixtures/complete-set.v1.json` and
  `fixture-manifests/complete_set.v1.json` align generic boundary expectations
  with the current profile registry.
- `internal/modules/reference_data/testsupport/windows_fixture.go`,
  `reference_pack_test_helpers_test.go`, `windows_online_integration_test.go`
  and Windows-specific fixtures under `contracts/reference-pack-fixtures/`.
- `tools/reference-packs/main.go`, `source.sh`, `witness.mjs`,
  `windows-witness.mjs`, `windows-office-witness.mjs` and the authored
  `tools/task_surface_owner.json` source-input declaration.
- Generated `internal/gen/contractreferencepacks/artifacts_gen.go`,
  `internal/gen/contractreferencepackfixtures/artifacts_gen.go` and task-surface
  projections. Earlier independent-pack routing and fixture splits remain
  documented in the historical record below.

An update admits new exact source bytes into a new immutable acquisition. Audit
and assembly read the retained lock and originals; they never refresh websites.
A reviewed correction is a new digest-bound input, never an in-place CSV edit.
Use distinct roots for publication reservations and envelope allocations, and
keep each pack's roots and fixture authority separate. Rebuilds of the same
logical release bind the same complete input identity; changed inputs require a
new release allocation. Renewal preserves logical bytes and changes envelopes.
Application activation is an explicit administration action. Historical packs,
locks, snapshots and evidence are not converted or reset by this delivery.

### Historical candidate identities before channel expansion

Retained source root: `.cartulary/reference-packs/windows-v2/acquisition`.
The authored inventory is
`contracts/reference-pack-producers/windows-event-ids/inventory.references.v2.json`.
The local-path acquisition recipe remains outside portable identities. The 308
previous online members were copied from retained bytes; three supplied files and
the prior receipt bring the closed inventory to 312. No website was refreshed.

| Subject | Identity |
| --- | --- |
| Source profile | `cartulary.windows_event_ids.microsoft_references.v2` |
| Source profile SHA-256 | `24515dc7d30f5216dc5627171bd9c9dfdf4ecc42926a7ad1ff5a11c9b904f2c5` |
| Source lock SHA-256 | `211348d7afa2e09ca6a07c5b8b54250114ccf2f1dc8c8a98e55c01800feb6a16` |
| Reviewed editorial SHA-256 | `fc8e3470f6f979337f93dec05469524ec8b24b4d76d157de441c7abab38f875f` |
| Actual CLI builder SHA-256 | `ad547cfcd8f0204e4f36fd522dcf6c5fa1d07e7068df8ca585e7dfa5eb05947b` |
| Builder | `cartulary.reference-pack.producer`, `references-v2`, `go1.27.3`, 14,939,270 bytes |
| Candidate payload-file SHA-256 | `cbe639afcaff9bf1f6c53c7b98feade7ad496a2119fc7c80373e9698814ccb1c` |
| Evidence archive SHA-256 | `9553b953e42d3c9cefb73afc1ed2ac1162e5be471e983be6b5365c8e895ddc28` |
| Reconciliation SHA-256 | `a6305be6e27a1cd8a7e8af1baecfa68bf579e8d699f882fecbfd650ee83d15f9` |
| Independent witness SHA-256 | `b1b9d3ab99ed3e59b1deb04e2da4a10489ed01604aa811a10a05d9a75641dcec` |

The builder is retained under `.cartulary/reference-packs/builders/<SHA-256>/`.
Candidate payload, archive and report are retained under the source root at
`audits/<reconciliation SHA-256>/`. The independent witness is at
`witnesses/audit-<reconciliation SHA-256>.json` and explicitly says
`candidate_status=blocked`. Repeated audit runs produced identical identities.
Maximum candidate line size is 132,092 bytes; maximum common description is 193
Unicode scalars. Rich content was not truncated.

| Provider | Candidate entries | Numeric version set | Unspecified-version entries |
| --- | ---: | --- | ---: |
| Microsoft-Windows-Eventlog | 5 | 0 | 0 |
| Microsoft-Windows-PowerShell | 1 | none | 1 |
| Microsoft-Windows-Security-Auditing | 189 | 0, 1, 2 | 0 |
| Microsoft-Windows-TaskScheduler | 5 | none | 5 |
| Microsoft-Windows-WMI-Activity | 3 | 0 | 1 |
| PowerShellCore | 195 | 1 | 1 |

The 398 entries represent 380 provider/event pairs and 379 distinct numeric IDs.
Output versions are 0:179, 1:207, 2:4 and null:8.

These are candidate definitions from admitted evidence, not a claim of complete
Windows event coverage. The exact 255 unresolved CSV record numbers, event IDs
and versions are in the report's `issues` with code `identity_conflict`; each has
an explicit `provider_evidence_missing` blocking disposition. Closing them needs
provider-qualified evidence tied to those original definitions. Coincident IDs,
Security channel labels and worksheet titles do not close the gap.

No real-corpus manifest, logical payload digest, container, bootstrap or pack-set
identity was produced. They remain absent rather than borrowed from fixtures.

### Earlier synthetic packaging and workflow identities

The public Make packaging exercise uses only synthetic sources and the explicit
synthetic license decision. Its repository is
`fixture.windows.events.references.v2`; it is not deployment authority.
The reproduced logical root is
`.cartulary/reference-packs/windows-v2/qualification/reproduced-release/logical/5571b67b40477b0ff1ee6df2aa583bdb327051b28f75cb8565921e81e64bc749`.

| Synthetic subject | Identity |
| --- | --- |
| Manifest SHA-256 | `5571b67b40477b0ff1ee6df2aa583bdb327051b28f75cb8565921e81e64bc749` |
| Logical payload digest | `ee86057988143cfa2f53b13dbaf9ed78afb4288f4068d12cad25f21fdab9d9a7` |
| Payload-file SHA-256 | `aae6df8120f981b8cd904f2fc942642ae55731eb260a98074218fc3346621f80` |
| Evidence archive SHA-256 | `35ccb85507579f8ac19c9bb72a864ea5b73534967909662ac498f16a81122617` |
| Initial container SHA-256 | `e9f13d1cccd59fdd120abb2358e6bd2539a87ee9dabf77a6f3197e548daafc73` |
| Renewed container SHA-256 | `b4dfa9ca3d09e33c8578023fdcfc1e716fa608fc442a54d690858217822ec750` |
| Bootstrap SHA-256, unchanged on renewal | `4a869f06bb50ca0250c028fd06032a07b8084a78b3680a34932dd3aea47467de` |

Both containers are 18,944-byte deterministic USTAR archives. Renewal advanced
metadata from 1 to 2 while preserving logical bytes and the bootstrap. Independent
logical and signature witnesses passed; freshness is verified separately by the
ordinary application lifecycle tests using the trusted wall clock.

The service fixture uses its actual test executable identity
`a849ec2f4e36e0d754c08a126fe38010afb185be66abcfc244da6ef4c8a89a73`.
Its A and B releases deliberately have different release labels, with the same
four-entry synthetic payload. The passing lifecycle test asserts the following
exact sets, each containing Windows plus the three required Base registries:

| Release | Manifest SHA-256 | Exact set ID |
| --- | --- | --- |
| A | `b5a1848f393dd037b6732538520ee953fb100ff624a512274b2a80a2b021c3b9` | `rpset_a3ee7318b0744f8779be4e4645c5a36d9194da02f1c19b65a37799203e00fdf3` |
| B | `75e87e3ae176e35f391a731b8d46fa2779c2803c86647dcb66a6cd24dec1d2ce` | `rpset_436d21319471fd00e275a80ca1247483415ec1823f616107f7bba0059655f927` |

Set values were reconstructed for this handoff from the same retained manifest
bindings and authored Base release bindings that the passing test compares to
application results. No application data or active deployment was modified.

### Final source qualification

`windows-v2-finalize-02` (`make agent-finalize`) passed, changed no generated
files and froze source digest
`sha256:a0783c35f1d838bd43c029d5da8b8a70f4ac341ffb86ecc755ce0f8b07c4fcbb`.
Its retained warm-run maintenance was deliberately skipped because the full check
was not successful. The following fresh selections validate the repaired source;
Windows and MITRE-specific rows remain separate invocations.

| Run ID | Exact command after the run-ID assignment | Outer exit |
| --- | --- | ---: |
| `windows-v2-unit-qualified` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930,module.reference_data.unit.enrichment_windows_event_ids_independently_valid_fc66475a5f` | 0 |
| `windows-v2-reference-qualified` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.enrichment_windows_event_ids_independently_impor_2bc83a2b47,module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e` | 0 |
| `windows-v2-reporting-qualified` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.windows_event_ids_snapshots_retain_their_exact_i_7312056cfb` | 0 |
| `windows-v2-portability-qualified` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.windows_event_ids_pack_portability_preserves_ind_d31bf11266` | 0 |
| `windows-v2-recovery-qualified` | `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.windows_event_ids_recovery_restores_independentl_0983448cb7` | 0 |
| `windows-v2-generated-qualified` | `make source-check CHECK=generated` | 0 |
| `windows-v2-schema-qualified` | `make source-check CHECK=schema` | 0 |
| `windows-v2-build-qualified` | `make reference-pack-producer-build` | 0; same retained executable hash |
| `attack-regression-v2-qualified` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c,module.reference_data.integration.framework_attack_independently_imports_activates_cd0be06ccc,module.reference_data.integration.attack_producer_artifacts_preserve_independent_a_67e9ef139a` | 0 |
| `windows-v2-markdown-qualified` | `make lint-markdown` | 0 |

The Windows selections contain 11 passing rows; the separate MITRE selection
contains three passing rows. The earlier `attack-regression-v2-final` invocation
also passed with the same three MITRE-only selections. Markdown lint retains
`adhoc/lint-markdown/tool-run-summary.json`.

Every graph-backed selection retains `run-summary.json` and exact row results
under `rows/`. The earlier `windows-v2-reference-final`,
`windows-v2-reporting-final`, `windows-v2-portability-final` and
`windows-v2-recovery-final` also passed on the first source freeze. These final
reruns account for the changed compiled fixture projection and truthful test
executable identities. The broad check was not repeated after a test-only repair;
its identified affected rows and generated/schema boundaries were rerun, and its
unrelated database/ACL failures remain an explicit repository-gate limitation.

Additional Windows lifecycle set identities, each with all three Base registries:

| Selection / release | Manifest SHA-256 | Exact set ID |
| --- | --- | --- |
| Reporting A | `06c68199400f4630e32bd8e27cd7c3b53b78098ada895001d93a54438852cf46` | `rpset_03f39ce1e824e517ea683351c24484e4cf250a64da3de82b0b6fa44c36e8d90f` |
| Reporting B | `fb1f852dffb552580925c05c8ca4067435ed32306486cfd39f4f7f78b9b49e49` | `rpset_33ffb5818e504bd252febe02176a1696495195b5b08e73d7ef08cf2409c1dde4` |
| Portability A | `07dcf3f69aec9fb69f6e01e1b7b1c1ad53711af280bc101c5d3d910c2ceec627` | `rpset_f7cca9408beda0fd899d4f627c0edbca7ea7587ecb804989a37f9b228441e54b` |
| Recovery A | `bb94c5fd41dfc682db66f2edab8a94399bdfa9720859403dd23fdd12446e8078` | `rpset_1455fd0ee464754c5ac183e7060a34d826386207a8b1784e736a0a39b2eaf586` |

Their exact build reports, actual test executable hashes and retained containers
are below each matching qualified run's `_shared/test-services/` root in
`windows-producer/<release>/logical/<manifest SHA-256>/build-report.json` and
`windows-envelopes/<container SHA-256>/`. Different test executables legitimately
produce different manifest identities; those identities are never reused as the
real corpus's identity.

### Earlier readiness record before channel expansion

| Category | Verdict and closure condition |
| --- | --- |
| Producer | V2 implementation and focused qualification pass; deterministic candidate, assembly, replay, witnesses and immutable publication are exercised. No real release is claimed. |
| Corpus | Complete source accounting and independent identity/original-fragment checks pass; publication is blocked by the exact 255 provider-unresolved CSV rows in the reconciliation report. Supply digest-bound provider-qualified evidence for those definitions and rerun audit to obtain zero unresolved canonical issues. |
| Test trust | Isolated Windows V2 authority and ordinary production verification pass. Synthetic licensing and test signing do not establish distribution or deployment trust. |
| Existing-interface workflow | Windows-backed import, explicit activation, rich lookup, exact-set provenance, retained A after B, snapshot pins, renewal, portability and restored availability pass. |
| Distribution | Blocked: no recorded source-specific licensing determination covering the derivative CSV, DOCX/XLSX originals, frozen online material and required notice bytes. Supply an approved exact-lock decision, expression/classification and notices; do not substitute the synthetic fixture decision. |
| Deployment | Blocked: actual publisher repository, release allocation, signing process and target trust configuration have not been supplied. Actual activation remains explicit. |
| Repository-wide gate | Not green: the three pre-existing database/ACL assertions above remain. Correct their owners and execute their affected checks before claiming broad acceptance. |

The new incident enrichment UI, incident importer and report derivation remain
outside delivery. No deployment, live activation, historical conversion or reset
occurred. Updates are new immutable acquisitions and releases; originals and
historical provenance remain byte-exact.

## Historical online V1 execution record

The remainder records the superseded online-only implementation. Its identities,
counts and passing runs describe V1 artifacts only and are not V2 acceptance.


Status: implementation and scoped qualification complete; real release and broad repository
qualification remain blocked by the explicit inputs/findings below. The controlling scope is the user-approved
online-source plan. The historical research attachment in `temp/` is preserved.
Initial checkout: `3bae6b5fc0720dd7d35d61747f3c366e9d85f650` (clean).

## Scope and owner reconciliation

The current owners agree on Reference Pack 0.3.0 / major 3, pack contract v2,
Windows entry profile v1 and incident-bundle format 7. No adoption cutover,
runtime profile expansion, migration, reset or historical rewrite is required.

Completeness means every admissible definition in the frozen first-party online
inventory, with complete source accounting. It does not mean complete installed
Windows provider metadata. The original four-build/24-cell capture requirement,
Windows collector, DLL/MUI acquisition, MSI installation and machine preparation
are superseded. No live logs, Windows APIs or incident content are collected.

All six provider families remain in scope. Explicit numeric event versions are
preserved; documented version-unspecified definitions use null, never a wildcard.
Known-version queries exclude null-version entries. Optional native metadata is
null unless supported for that exact definition. Documentation samples do not
prove occurrence or exact-build applicability. Current public schemas and consumer
operations remain authoritative.

The delivery includes offline production and existing pack administration,
lookup, retained binding, portability and recovery interfaces. A new incident
importer, enrichment UI or report derivation is outside the agreed scope.

## Remediation and execution phases

| Gap | Remediation and area | Rationale and long-term benefit | Compatibility and unresolved risk | Exit |
| --- | --- | --- | --- | --- |
| Capture scope conflicts with available inputs | Replace producer scope and gates; specification/documentation | Honest, reproducible documentary coverage | New source-profile identity; retaining old claims would misrepresent coverage | Online scope and every exclusion recorded |
| Mutable and incomplete online evidence | Closed inventory, bounded acquisition and immutable lock; implementation/tests | Reviewable updates and detectable omissions | No runtime effect; otherwise sources drift silently | Exact bytes, URLs, times and digests retained |
| Mixed documentary and manifest fidelity | Field-level evidence and conservative reconciliation; producer specification/implementation/tests | Preserve uncertainty without losing known versions | Existing nullable profile; otherwise invented precision or version collapse | Every assertion accounted for and conflicts block |
| ATT&CK-specific publication mechanics | Explicit source adapters over shared packaging and fixture trust; implementation/tests | One maintained security boundary | Retained artifacts unchanged; otherwise wrong identity or trust attribution | Independent witnesses and regression checks |
| Unexecuted Windows workflow | Existing consumer/lifecycle Harness integration; tests | Validate production boundaries without duplicate verifier | No new incident workflow; otherwise provenance drift can escape | Executed expectation bindings |
| Cross-pack test setup and implicit source defaults | Independent profile rows, explicit source inputs, pack-neutral helpers and isolated fixture namespaces; specification/implementation/tests/documentation | Each pack can qualify without another optional corpus or release | Old combined row and implicit ATT&CK source defaults are removed; callers select their own inputs/rows | Every profile has separate unit/consumer selection; Windows and ATT&CK lifecycle selections execute independently |
| Runtime/test/producer import coupling | Separate HTTP fixture helpers from fixture data; enforce that runtime cannot import offline producer/fixture packages; implementation/tests | Avoid application import cycles and keep deployment storage checks scoped to runtime | No public runtime change; otherwise unrelated test composition fails or offline tooling is mistaken for deployment storage | Composition and rooted filesystem boundary rows pass |
| Stale incident-bundle test generation | Align two existing test expectations with the adopted format 7; tests | Preserve meaningful rejection and defensive-copy checks | No data migration; stale tests incorrectly reject the current contract | Source catalog and codec rows pass |
| Publication inputs absent | Separate readiness and input gates; documentation/release evidence | Prevent approval claims from test success | No deployment or distribution inferred | Actual licensing and trust decisions supplied |

Acquisition and projection implementation can proceed before publication inputs.
Real assembly is blocked until a recorded source-specific license decision exists.
Synthetic fixtures never stand in for Microsoft-derived corpus acceptance.

### Historical requirement disposition

The original attachment is research history, not an additional execution mandate.
Its sections map to this delivery as follows:

| Original sections | Disposition |
| --- | --- |
| 1, 3.1, 3.5, 4.1, 6.1–6.4 | Replaced by a finite documentary inventory, HTTPS acquisition and immutable receipt; no exact-build completeness claim. |
| 2 | Adoption mismatch resolved against live owners; retain all current identities. |
| 3.2–3.4, 3.6–3.7, 4.3–4.4 | Retain provider/ID/version identity, provider-qualified collisions, text fidelity and six-family scope; no installation/resource inference. |
| 4.2 | Replace installer acquisition with the release-pinned XML manifest and first-party release metadata. |
| 5, 6.5 | Retain separate source, logical and transport artifacts; reuse existing source/package Make interfaces with closed-schema dispatch. |
| 7 | Retain common entry/profile mapping and strict normalization; absent evidence remains null, including undocumented versions. |
| 8 | Replace build union with documentary assertion reconciliation; retain explicit dispositions, deterministic formatting, lifecycle neutrality and bounds. |
| 9–10 | Retain immutable source locators, canonical bytes, trust renewal and separate licensing gates; use existing USTAR. |
| 11.1–11.2, 11.4 | Retain current lookups and historical set bindings, with actual corpus coverage instead of assumed installed definitions. |
| 11.3 | Exclude a new incident consumer, importer, UI and report derivation; samples remain documentation. |
| 12 | Retain producer/consumer/lifecycle verification and independent expectations; use ordinary trusted wall-clock verification with renewed fixture envelopes. |
| 13–15 | Replace capture-dependent sequencing and completion gates with the phases and readiness categories below; never infer publication approval. |

### Sequencing and phase exit criteria

| Phase / workstream | Dependencies | Principal risk | Exit criteria |
| --- | --- | --- | --- |
| 0: owner and scope cleanup | Inspect current normative owners and implementations | Obsolete capture gates survive | Current identities, scope mapping, explicit exclusions and handoff recorded. |
| 1: acquisition | Phase 0 and closed producer schemas | Mutable or missing source bytes; unsafe redirects | Complete immutable lock, or individual failures with no complete-lock claim; index link dispositions retained. |
| 2: extraction/reconciliation | Phase 1 schemas; retained bytes for corpus acceptance | Version leakage, silent omissions, ambiguous provider binding | Every source assertion has a disposition, all locators resolve, conflicts/bounds block, independent identity sets agree. |
| 3: packaging/trust | Phase 1 interfaces and Phase 2 accepted output | Cross-family authority or partial publication | Scoped reservations, atomic logical publication, deterministic USTAR, isolated fixture trust, actual builder identity and ATT&CK regression. Real release additionally needs licensing and publisher inputs. |
| 4: current-interface acceptance | Phase 3 synthetic fixtures; licensing for a real pack | Producer success conceals consumer or historical-binding defects | Production import, explicit activation, version lookups, exact sets including Base, retained A across B, renewal, snapshot pins, portability and restored bytes execute. |
| 5: handoff | All supported preceding exits | Test trust mistaken for deployability | Actual identities and coverage, exact test evidence, separate readiness verdicts and missing-input closure conditions recorded. |

These workstreams do not add public selectors, aliases, extension fields or routes.
They keep the runtime profile stable while making source-specific producer policy
explicit and reusable publication mechanics independent of ATT&CK projection.

## Execution evidence

Execution evidence is recorded below. Candidate corpus acceptance,
synthetic application acceptance, real release assembly and deployment are
separate claims.

### Source addition requested during implementation

The user replaced the proposed Ultimate Windows Security encyclopedia with
[Microsoft Appendix L: Events to monitor](https://learn.microsoft.com/en-us/windows-server/identity/ad-ds/plan/appendix-l--events-to-monitor).
The encyclopedia proposal is superseded and is not a remaining dependency.
No encyclopedia content was bulk acquired or incorporated. The new Microsoft
source is included in the maintained seeds and a separate immutable acquisition;
the original 307-source acquisition remains unchanged.

Appendix L supplies monitoring recommendations and current/legacy event IDs.
It does not provide an explicit provider or numeric event version for each row.
Each row therefore receives an explicit exclusion disposition rather than a
fabricated identity. Current and legacy Windows generations are not event
versions, and monitoring criticality is not native event level. The table and
its applicability remain available as retained documentary evidence. The
independent witness checks every unqualified table row as well as admitted
source identity sets.

### Frozen corpus and identities

The accepted inventory has **308 retained sources**: 238 event pages, 60 selected
audit indexes, five provider articles, two licensing sources, two GitHub release
records and one pinned PowerShell manifest. Acquisition retained 13,447,983 bytes
between `2026-11-07T01:17:20Z` and `2026-11-07T01:18:40Z`.

Discovery records 401 candidate-link dispositions: 263 included event-document
links, 133 included index links and five excluded navigation links. Repeated
links coalesce into the closed source inventory; their dispositions remain.

There are 569 qualified definition assertions: 391 emitted identities and 178 compatible
coalescences. Another 63 event documents lack unambiguous provider evidence and
are explicitly excluded; their documented IDs/versions are not assigned a provider
from a folder or channel. Appendix L adds 381 explicitly excluded table-row
assertions without provider qualification. Its retained HTML has 99,826 bytes
and SHA-256 `91cdfc5960ffb3e7cfc6fcc43908073ce139a96064c0e169dfe8755a0ba4c237`.
Supporting dispositions cover 60 indexes, two licensing
sources, two release records and one non-event performance-counter provider.
There are no unresolved conflicts. The maximum complete description is 1,696
scalars and the maximum NDJSON row including LF is 2,516 bytes. The maximum
source-reference count is six per entry, represented through one indexed
evidence archive; eight entries have a genuinely unspecified version.

| Provider | Assertions | Entries | Numeric versions | Null-version entries |
| --- | ---: | ---: | --- | ---: |
| `Microsoft-Windows-Eventlog` | 2 | 1 | 0 | 0 |
| `Microsoft-Windows-PowerShell` | 1 | 1 | none | 1 |
| `Microsoft-Windows-Security-Auditing` | 358 | 186 | 0, 1, 2 | 0 |
| `Microsoft-Windows-TaskScheduler` | 10 | 5 | none | 5 |
| `Microsoft-Windows-WMI-Activity` | 3 | 3 | 0 | 1 |
| `PowerShellCore` | 195 | 195 | 1 | 1 |

For example, event 4624 retains versions 0, 1 and 2; only version 2 receives
the numeric fields present in its XML sample. PowerShellCore 4104 retains both
version 1 and a separate null-version documentary identity.

All hashes below are actual SHA-256 values. The payload-file hash identifies the
candidate NDJSON, not the manifest-framed logical payload digest.

| Subject | Identity |
| --- | --- |
| Source profile | `cartulary.windows_event_ids.microsoft_online.v1` |
| Frozen inventory | `0dea611eb43496cf626322855d6c092e056981643d317f0f23475412d4c5a149` |
| Source profile bytes | `6af24017d53942dd852d40085014aeccfaf2550a3d0f07523f9c49958bb30fb8` |
| Accepted source lock | `f8ab668f98e707a7e3cc87c9918369912901bbdb234ff18849bf3521daa0c1d8` |
| Reviewed editorial input (empty notes) | `28926782cd80239bac78890a2ec7a7824f9c7218dfb733fe9568c64aa50119cf` |
| Candidate payload file | `916720d0073aa473596357d349dbc4d69979a1b1eb80968b62055c8c4aa1215c` |
| Source evidence USTAR | `f6243bbc81177f8f5fa5b22c7388c97cdd5e835eaedc9a087763b92d77868e30` |
| Reconciliation report | `07075debab330fe7a19dea57e496f3f100e8db7c06d612282042edfc5e114c80` |
| Independent source witness | `9b343998a161e7ad88f5370801b7f7605f6cb79933a3ffd55886b1015c22dfca` |
| Actual producer executable | `c8b41f2d1e65885b16cc1ed9a48c38f1fb2a105d6eae917a57b1e9c4074ed0d4` (14,120,425 bytes; Go 1.27.3) |

The final builder ID is `cartulary.reference-pack.producer`, version `online-v1`.
Two builds with identical inputs produced identical bytes; the executable,
identity and reproduction receipt are retained under
`.cartulary/reference-packs/builders/c8b41f2d1e65885b16cc1ed9a48c38f1fb2a105d6eae917a57b1e9c4074ed0d4/`.
The accepted lock and exact sources are under
`.cartulary/reference-packs/windows-online/acquisition-appendix/`.
The candidate, evidence archive and report are under that root's
`audits/07075debab330fe7a19dea57e496f3f100e8db7c06d612282042edfc5e114c80/`;
the witness is `witnesses/audit-07075debab330fe7a19dea57e496f3f100e8db7c06d612282042edfc5e114c80.json`.
These are local retained artifacts in ignored storage, not committed distribution assets.

The PowerShell manifest is bound to commit
`bf5fc434120666e5f1a8d1dfd6ecac7efc17a4be` through retained annotated tag
`a0edd6969993766081edc6bf0306696d1adef88a` and release `v7.6.9`. A mutable
`target_commitish` is not accepted as that binding. Learn's page revision metadata
remains in the exact retained HTML; no undocumented event version is inferred.

**Real-pack logical payload digest, manifest digest, container identity,
bootstrap identity and pack-set identity are absent.** No real license decision,
release allocation or deployment authority was fabricated to create them.

### Operation and update procedure

Run public targets from the repository root. The current accepted inventory is
`.cartulary/reference-packs/windows-online/discovery-appendix/inventories/0dea611eb43496cf626322855d6c092e056981643d317f0f23475412d4c5a149.json`.

1. `make reference-pack-producer-build` retains the actual deterministic builder.
2. `make reference-pack-source ACTION=discover SOURCE_LOCK=<seed-inventory.json> SOURCE_DIR=<new-discovery-root>` enumerates the bounded official indexes and records every candidate disposition. Review the resulting inventory and exclusions.
3. `make reference-pack-source ACTION=acquire SOURCE_LOCK=<frozen-inventory.json> SOURCE_DIR=<new-source-root>` retains exact bytes and publishes a complete lock only on success. Failed reports are retained separately.
4. `make reference-pack-source ACTION=check SOURCE_LOCK=<frozen-inventory.json> SOURCE_DIR=<retained-source-root>` verifies existing bytes and release pins without fetching.
5. `make reference-pack-source ACTION=audit SOURCE_LOCK=<frozen-inventory.json> SOURCE_DIR=<retained-source-root> EDITORIAL=<reviewed-editorial.json>` produces a candidate and independent source witness. Omitted `EDITORIAL` selects the embedded empty-notes input.
6. After an actual source-specific licensing decision and publication allocation, use `make reference-pack-package ACTION=assemble INPUT=<windows-build-input.json>`. The closed build and license schemas define required inputs; no default real classification exists.
7. Isolated fixture envelopes use `make reference-pack-package ACTION=envelope INPUT=<windows-envelope-input.json>` and `ACTION=renew` with the previous envelope path. This signer cannot represent a deployment repository. Activation remains an explicit administrator action.

A failed assembly may occupy its immutable reservation. Replay requires exactly
the same output-affecting inputs. Changes require a new version/allocation; never
reset the ledger or edit retained evidence. A new acquisition uses a new root;
accepted locks are never refreshed. Source-profile semantic changes require a new
profile identity, including changes to the frozen Unicode 17.0.0 normalization.

Websites are one-time references, as explicitly requested by the user. Ordinary
tests use synthetic inputs; retained-corpus checks and builds use exact local
bytes. No test, build, release acceptance or ongoing operation requires a source
website to continue to exist or remain unchanged. The URLs preserve citations.
Optional future acquisition is a deliberate new research input, not a refresh
gate for any retained artifact. Live source availability is not a readiness
category or a post-acquisition blocker.

New logical and envelope trees publish atomically. Independent witnesses live in
sibling `witnesses/` directories. Historical output trees containing an in-tree
witness remain untouched; replay those legacy artifacts with their retained
builder, or choose a fresh output root for a new builder. No historical pack,
ledger, Base registry or trust bootstrap was rewritten.

### Current synthetic integration identities

These identities belong only to the four-entry project-authored fixture in
`pack-enrichment-windows-event-ids-release`; they are not a Microsoft-derived release.
The Windows lifecycle row passed and checked both exact sets independently,
including all three required Base registries. Its actual test executable is
truthfully identified as builder. A and separately labeled derivative B have
identical definition bytes; their versions produce distinct logical manifests.

| Subject | SHA-256 or exact identity |
| --- | --- |
| Actual test executable | `b8f0bfcd7687d9bfd349703bb32635181b5db9f00b57aa8ab469858b0ab5662d` |
| Synthetic source lock | `d4283b5a7269e5aeeb07363b6cf9760ad52e509027fedb1087402263ae822fd4` |
| Synthetic evidence archive | `7f2da3a6b2ce733ea54a167b669253162211a74688692ff8f2245486ffd3ba66` |
| Payload file | `6eeefaa9d64aacdd81b131f2c465c5d7dda00c4a264bbcc60caab9d5cd704dbe` |
| Logical payload, A and B | `436364d48356ceeca8c8408e599e5bd378502b8cf5db487351ac6077027d3917` |
| A manifest | `0c797da0d7bd229717a81b9a848c9b174c84394e67b3e36f810e5307099f0a7b` |
| B manifest | `2a72294088a1e2704a4a332b30cf4664b8c71aea9713166f63f679963238689f` |
| A exact set | `rpset_cb770f890c0558549fb3845af885d44d42a6d3f3cc2122a9647390429c343b5e` |
| B exact set | `rpset_fce3bce6deebc917e1bcab0ae02615a64677b6522d14872ab829410714d96bc3` |
| Imported A container | `c2107972431cfaf00bbd94751237d7e809f0d9bf6860b36e1ee486343257644d` |
| Imported B container, initial | `71c5c1667cc3ccddb74318e7dc6d447dca6db84abe9cfffb47f56b1b36c87bad` |
| Imported B container, renewed | `849c704645596ef8528903613721e507e13663fa4cd638471e3c3413a0d64ab7` |
| Integration bootstrap | `867096ecf4f0d7c920fce4f0356447ec60f753f20539e7b4db9a85491a1ae93a` |
| Independently witnessed A transport, metadata 1 | `28eadfd556de7e0c0fa8709047850c7a51c154b30aa06877d94651c1d248c8ff` |
| Independently witnessed A renewal, metadata 2 | `ac70e59dea5cb564b50e57b18684d2f1c4ceb3aab7c0690e1a59e30ce78e8189` |
| Independent transport bootstrap | `b814ba78ef2824b2b8db8152259fe848f86d861509f9f4ec0aa6ade599bfce86` |

Suite artifacts live under
`.cartulary/test-results/pack-enrichment-windows-event-ids-release/_shared/test-services/c8d50ff160a04cf8ab8f7089/`.
Independent envelope reports and witnesses live under
`.cartulary/reference-packs/windows-online/release-fixture-transport/`.
The independent Node witness establishes byte and signature integrity; its
`freshness_verified:false` is intentional. Ordinary production verification in
the service tests establishes freshness using trusted current wall-clock time.
No Network Flow clock control or system-clock change was used. Earlier
synthetic identities remain in their historical run roots and were not rewritten.

### Earlier executed verification portfolio

All commands ran from the repository root through public Make targets. For each
named run below the command was prefixed with `CARTULARY_TEST_RUN_ID=<run>` and
the retained root is `.cartulary/test-results/<run>/`. Graph summaries are
`target-summaries/<target>.json` plus `run-summary.json`; producer tool summaries
are `<target>/tool-run-summary.json`. Process exits are recorded separately from
diagnostic-reader exits. Selected rows establish their stated obligations, not
complete-owner or release acceptance.

Exact row selections:

- `W`: `module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930`
- `A`: `module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c`
- `C`: `module.reference_data.unit.canonical_consumers`
- `P`: `module.reference_data.unit.canonical_pack_profiles`
- `R`: `module.reference_data.integration.all_profiles_consumer_boundary`
- `S`: `module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c`
- `I`: `module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8`
- `B`: `module.recovery.integration.real_postgres_backed_runtime_metadata_persistenc_d652f0998b`

The letter notation below expands literally to those row IDs; it is not Make syntax.

| Run | Command after run-ID prefix | Process exit / outcome |
| --- | --- | --- |
| `windows-online-unit-03` | `make test-slice OWNER=module.reference_data ROWS=W,A,C,P` | 0, pass |
| `windows-online-unit-04` | `make test-slice OWNER=module.reference_data ROWS=W,A,P` | 0, pass; inclusive text bounds and immutable publication |
| `windows-online-unit-05` | `make test-slice OWNER=module.reference_data ROWS=W,A` | 0, pass; release/tag pin checks |
| `windows-online-unit-06` | `make test-slice OWNER=module.reference_data ROWS=W,A` | 0, pass; final producer changes |
| `windows-online-finalize-01`, `windows-online-finalize-02` | `make agent-finalize` | 0, pass; generated changes reviewed; second run froze the earlier pre-Appendix implementation |
| `windows-online-builder-final` | `make reference-pack-producer-build` | 0, pass; two byte-identical executable builds |
| `windows-online-ref-final` | `make test-slice OWNER=module.reference_data ROWS=R SCOPE=services` | 0, pass |
| `windows-online-report-final` | `make test-slice OWNER=module.reporting ROWS=S SCOPE=services` | 0, pass |
| `windows-online-portable-final` | `make test-slice OWNER=module.incidentbundles ROWS=I SCOPE=services` | 0, pass |
| `windows-online-recovery-final` | `make test-slice OWNER=module.recovery ROWS=B SCOPE=services` | 0, pass |
| `windows-online-generated-final` | `make source-check CHECK=generated` | 0, pass |
| `windows-online-schema-final` | `make source-check CHECK=schema` | 0, pass |
| `windows-online-envelope-final` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/windows-online/fixture-transport-inputs/envelope.json` | 0, pass including independent witness |
| `windows-online-renew-final`, `windows-online-renew-replay-final` | `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/windows-online/fixture-transport-inputs/renew.json` | 0, pass; exact renewal replay |
| `windows-online-attack-assemble-final`, `windows-online-attack-replay-final` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/windows-online/attack-regression-input.json` | 0, pass including independent ATT&CK witness and atomic replay |

Earlier source operations used the original 307-source inventory and acquisition root:

| Run | Command | Process exit / outcome |
| --- | --- | --- |
| `windows-online-discovery-02` | `make reference-pack-source ACTION=discover SOURCE_LOCK=contracts/reference-pack-producers/windows-event-ids/inventory.seeds.v1.json SOURCE_DIR=.cartulary/reference-packs/windows-online/discovery-02` | 0, 307 sources frozen |
| `windows-online-acquire-01` | `make reference-pack-source ACTION=acquire SOURCE_LOCK=<frozen-inventory> SOURCE_DIR=.cartulary/reference-packs/windows-online/acquisition-01` | 0, complete acquisition |
| `windows-online-source-check-final` | `make reference-pack-source ACTION=check SOURCE_LOCK=<frozen-inventory> SOURCE_DIR=.cartulary/reference-packs/windows-online/acquisition-01` | 0, offline integrity and release pin verification |
| `windows-online-audit-final`, `windows-online-audit-replay-final` | `make reference-pack-source ACTION=audit SOURCE_LOCK=<frozen-inventory> SOURCE_DIR=.cartulary/reference-packs/windows-online/acquisition-01` | 0, accepted candidate, independent accounting, identical replay |

`<frozen-inventory>` in this historical table is
`.cartulary/reference-packs/windows-online/discovery-02/inventories/185d8cd000b82bd5fe6cd299682e1ffd499def2e78b6b4206408acd11e634b56.json`. No acquisition, lock, candidate or test fixture is itself a licensing
decision. The ATT&CK regression used its retained source lock/profile and a new
output root/ledger; it did not replace any historical artifact.

The Windows producer rows cover source substitution and incomplete receipts,
redirect/URL and UTF-8 boundaries, strict closed JSON, DTD/entity rejection,
documented versus unspecified versions, zero descriptors, scoped manifest
opcodes, conflicting source facts, input reordering, immutable locators, final
text bounds and atomic/colliding publication. Existing canonical consumer/profile
rows supply provider-colon identity parsing and broader malformed-request cases.
The new production scenario adds case folding, ID-only ambiguity, pagination,
numeric/null version distinction, retained A provenance after B and envelope
renewal. Reporting executes actual snapshot admission/pins and retained A
rendering. Portability embeds and restores exact allowed synthetic pack bytes;
Recovery restores actual metadata, logical and container availability.

### Resolved development failures and evidence limits

Earlier failures are retained; later passes do not erase them:

- `windows-online-unit-01`: process exit 2, missing brace in the new authored producer schema; corrected and the selected rows passed from `unit-02` onward.
- Generation roots `20261107T003813Z-p4436`, `20261107T003844Z-p7883`, and `20261107T005010Z-p29981`: process exit 2. Causes were the platform-reserved `windows_test.go` filename, missing nullable default for authored `EDITORIAL`, and an invalid fixture-root declaration. Renamed the test, repaired authored routing, removed the unnecessary root, then regenerated through Make. No policy was weakened.
- Initial discovery root `20261107T004132Z-p22537`: process exit 2 before the canonical input/schema correction; `windows-online-discovery-02` is the accepted discovery.
- `windows-online-audit-01`: process exit 2; documentation samples contained occurrence placeholders/malformed payload text. Extraction was narrowed to the exact inert System metadata rather than repairing source text; non-event manifest providers gained explicit dispositions.
- `windows-online-audit-02`: process exit 2 in independent accounting; version-list `0.` punctuation and documentary table accounting were corrected. Subsequent independent audits passed. Final identities above supersede earlier audit identities, which remain retained.
- `windows-online-ref-01`: process exit 2 because the new test passed Go integers to the canonical consumer instead of `json.Number`. The fixture request was corrected; `windows-online-ref-02` and `windows-online-ref-final` passed. No consumer contract was relaxed.
- A multi-owner `test-slice` attempt was rejected with usage exit 2 before test execution; individual owner commands above are the executed evidence. A diagnostic request for undeclared `TARGET=backend-integration` was similarly rejected; the retained target projection was inspected directly.

The pre-final service roots `windows-online-report-01`, `windows-online-portable-01`
and `windows-online-recovery-01` also passed with process exit 0. They are historical
development evidence, not substituted for the final frozen-source runs.
Retained-run maintenance requiring a compatible passing full warm-check root was
not requested with an incompatible or partial root. No visual golden, UI review,
deployment, real-pack activation or new incident enrichment workflow was run.

### Changed paths

All 76 paths are repository-relative. Generated projections were produced through
Make from authored inputs. The historical research attachment and domain model
were read and left unchanged. `testsupport/attack_http.go` was removed; its
neutral HTTP helper now lives in `testsupport/fixturehttp/http.go`.

- `contracts/reference-pack-fixtures/fixture-manifests/owner_windows_online.v1.json`
- `contracts/reference-pack-fixtures/fixture-manifests/owner_windows_snapshot_binding.v1.json`
- `contracts/reference-pack-fixtures/fixtures/owner_snapshot_binding.v1.json`
- `contracts/reference-pack-fixtures/fixtures/owner_windows_online.v1.json`
- `contracts/reference-pack-fixtures/fixtures/owner_windows_snapshot_binding.v1.json`
- `contracts/reference-pack-fixtures/owner_fixture.v1.schema.json`
- `contracts/reference-pack-producers/windows-event-ids/contracts.v1.schema.json`
- `contracts/reference-pack-producers/windows-event-ids/editorial.v1.json`
- `contracts/reference-pack-producers/windows-event-ids/fixture-sources.v1.json`
- `contracts/reference-pack-producers/windows-event-ids/inputs.go`
- `contracts/reference-pack-producers/windows-event-ids/inventory.seeds.v1.json`
- `contracts/reference-pack-producers/windows-event-ids/source-profile.v1.json`
- `contracts/verification/release-qualification.v1.json`
- `docs/handoffs/windows-event-ids-reference-pack.md`
- `docs/reference-pack-subsystem-nlspec.md`
- `internal/gen/contractreferencepackfixtures/artifacts_gen.go`
- `internal/modules/incidentbundles/reference_attack_round_trip_integration_test.go`
- `internal/modules/incidentbundles/reference_windows_round_trip_integration_test.go`
- `internal/modules/incidentbundles/routes_round_trip_integration_test.go`
- `internal/modules/incidents/incident_bundle_source_codec_test.go`
- `internal/modules/incidents/internal/sourcestate/manifest_test.go`
- `internal/modules/recovery/backup_metadata_integration_test.go`
- `internal/modules/recovery/reference_windows_integration_test.go`
- `internal/modules/reference_data/internal/packformat/content_admission_test.go`
- `internal/modules/reference_data/internal/packformat/content_diagnostics_test.go`
- `internal/modules/reference_data/internal/packformat/content_limit_test.go`
- `internal/modules/reference_data/internal/packformat/content_test.go`
- `internal/modules/reference_data/internal/packformat/framework_lookup_test.go`
- `internal/modules/reference_data/internal/packformat/manifest_limit_test.go`
- `internal/modules/reference_data/internal/packformat/manifest_test.go`
- `internal/modules/reference_data/producer/attack_projection_test.go`
- `internal/modules/reference_data/producer/base.go`
- `internal/modules/reference_data/producer/envelope.go`
- `internal/modules/reference_data/producer/identity.go`
- `internal/modules/reference_data/producer/ledger.go`
- `internal/modules/reference_data/producer/package.go`
- `internal/modules/reference_data/producer/publication.go`
- `internal/modules/reference_data/producer/source.go`
- `internal/modules/reference_data/producer/source_test.go`
- `internal/modules/reference_data/producer/test_helpers_test.go`
- `internal/modules/reference_data/producer/windows_acquire.go`
- `internal/modules/reference_data/producer/windows_contract.go`
- `internal/modules/reference_data/producer/windows_extract.go`
- `internal/modules/reference_data/producer/windows_online_test.go`
- `internal/modules/reference_data/producer/windows_package.go`
- `internal/modules/reference_data/producer/windows_reconcile.go`
- `internal/modules/reference_data/profile_consumers_integration_test.go`
- `internal/modules/reference_data/reference_pack_test_helpers_test.go`
- `internal/modules/reference_data/testsupport/attack_http.go`
- `internal/modules/reference_data/testsupport/fixturehttp/http.go`
- `internal/modules/reference_data/testsupport/owner_fixture.go`
- `internal/modules/reference_data/testsupport/windows_fixture.go`
- `internal/modules/reference_data/windows_online_integration_test.go`
- `internal/modules/reporting/composition_preview_integration_test.go`
- `internal/modules/reporting/migrationadmission/reference.go`
- `internal/modules/reporting/reference_binding_attack_integration_test.go`
- `internal/modules/reporting/reference_binding_integration_test.go`
- `internal/modules/reporting/reference_binding_windows_integration_test.go`
- `internal/platform/rootedfs/production_boundary_test.go`
- `tools/backend_module_boundaries.json`
- `tools/execution_topology_render_index.json`
- `tools/reference-packs/build.sh`
- `tools/reference-packs/main.go`
- `tools/reference-packs/source.sh`
- `tools/reference-packs/windows-witness.mjs`
- `tools/reference-packs/witness.mjs`
- `tools/task_surface.generated.mk`
- `tools/task_surface.help-all.generated.json`
- `tools/task_surface.runtime.generated.mk`
- `tools/task_surface_manifest.json`
- `tools/task_surface_owner.json`
- `tools/test_families/module.incidentbundles.json`
- `tools/test_families/module.recovery.json`
- `tools/test_families/module.reference_data.json`
- `tools/test_families/module.reporting.json`
- `tools/test_support_inventory.json`

### Independent pack qualification

The user's later instruction requires separate pack qualification. The previous
combined consumer row has been removed. Each of the 16 profiles now has its own
content-contract row and isolated production-consumer row, reading only its own
authored fixture. Each consumer scenario starts fresh application state and
checks that only the selected optional pack plus the three required Base
registries is active. Indicator normalization and maximum-identity checks run
only in their owning profile scenarios.

Windows and ATT&CK lifecycle, snapshot and portability scenarios have separate
rows and fixture-construction files. Shared lifecycle helpers receive explicit
pack data and license expectations; they no longer choose between pack families.
Windows recovery has its own scenario, while the original generic metadata
recovery scenario remains independently runnable. Windows source/trust tests
use an unrelated synthetic repository for isolation checks, not ATT&CK fixtures.
Common canonical encoding, filesystem and HTTP test helpers are pack-neutral.
HTTP support lives in `testsupport/fixturehttp`, so pure fixture data no longer
imports application/server composition. The runtime filesystem boundary test
rejects runtime imports of offline producer and fixture packages.
The source command requires explicit `SOURCE_LOCK` and `SOURCE_DIR`; it no
longer silently selects ATT&CK. Existing callers must supply those two inputs.
The release qualification union names the independent rows explicitly; it does
not make one optional pack a setup dependency of another.

Current pack-specific selections (pass evidence is recorded separately):

| Pack/scenario | Row ID |
| --- | --- |
| `profile-unit:type_registry.host` | `module.reference_data.unit.type_registry_host_independently_validates_canon_2d41993e71` |
| `profile-services:type_registry.host` | `module.reference_data.integration.type_registry_host_independently_imports_activat_a3707c8ace` |
| `profile-unit:type_registry.evidence` | `module.reference_data.unit.type_registry_evidence_independently_validates_c_ac918d851a` |
| `profile-services:type_registry.evidence` | `module.reference_data.integration.type_registry_evidence_independently_imports_act_046a55d962` |
| `profile-unit:type_registry.indicator` | `module.reference_data.unit.type_registry_indicator_independently_validates_bc40be154e` |
| `profile-services:type_registry.indicator` | `module.reference_data.integration.type_registry_indicator_independently_imports_ac_6c2cb894b3` |
| `profile-unit:framework.attack` | `module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f` |
| `profile-services:framework.attack` | `module.reference_data.integration.framework_attack_independently_imports_activates_cd0be06ccc` |
| `profile-unit:framework.d3fend` | `module.reference_data.unit.framework_d3fend_independently_validates_canonic_420fac0262` |
| `profile-services:framework.d3fend` | `module.reference_data.integration.framework_d3fend_independently_imports_activates_f8ef62f99f` |
| `profile-unit:framework.veris` | `module.reference_data.unit.framework_veris_independently_validates_canonica_bb0bacb53d` |
| `profile-services:framework.veris` | `module.reference_data.integration.framework_veris_independently_imports_activates_8ed7f795eb` |
| `profile-unit:enrichment.tor` | `module.reference_data.unit.enrichment_tor_independently_validates_canonical_e4e41a33af` |
| `profile-services:enrichment.tor` | `module.reference_data.integration.enrichment_tor_independently_imports_activates_a_69bd8896bd` |
| `profile-unit:enrichment.cisa_kev` | `module.reference_data.unit.enrichment_cisa_kev_independently_validates_cano_9bc6dc025b` |
| `profile-services:enrichment.cisa_kev` | `module.reference_data.integration.enrichment_cisa_kev_independently_imports_activa_c6cc146f1c` |
| `profile-unit:enrichment.ms_portals` | `module.reference_data.unit.enrichment_ms_portals_independently_validates_ca_57e34f2a64` |
| `profile-services:enrichment.ms_portals` | `module.reference_data.integration.enrichment_ms_portals_independently_imports_acti_860490d5c2` |
| `profile-unit:enrichment.windows_event_ids` | `module.reference_data.unit.enrichment_windows_event_ids_independently_valid_fc66475a5f` |
| `profile-services:enrichment.windows_event_ids` | `module.reference_data.integration.enrichment_windows_event_ids_independently_impor_2bc83a2b47` |
| `profile-unit:enrichment.entra_app_ids` | `module.reference_data.unit.enrichment_entra_app_ids_independently_validates_02047fa20a` |
| `profile-services:enrichment.entra_app_ids` | `module.reference_data.integration.enrichment_entra_app_ids_independently_imports_a_07faf90a4a` |
| `profile-unit:enrichment.lolbas` | `module.reference_data.unit.enrichment_lolbas_independently_validates_canoni_397137c3ee` |
| `profile-services:enrichment.lolbas` | `module.reference_data.integration.enrichment_lolbas_independently_imports_activate_f14dd266ff` |
| `profile-unit:enrichment.loldrivers` | `module.reference_data.unit.enrichment_loldrivers_independently_validates_ca_8d4da9f6a9` |
| `profile-services:enrichment.loldrivers` | `module.reference_data.integration.enrichment_loldrivers_independently_imports_acti_a068f18ffc` |
| `profile-unit:enrichment.lolesxi` | `module.reference_data.unit.enrichment_lolesxi_independently_validates_canon_5a1260a9d7` |
| `profile-services:enrichment.lolesxi` | `module.reference_data.integration.enrichment_lolesxi_independently_imports_activat_ebb18a1387` |
| `profile-unit:enrichment.hijacklibs` | `module.reference_data.unit.enrichment_hijacklibs_independently_validates_ca_4c78461ed3` |
| `profile-services:enrichment.hijacklibs` | `module.reference_data.integration.enrichment_hijacklibs_independently_imports_acti_42ba3a94e9` |
| `profile-unit:enrichment.windows_sids` | `module.reference_data.unit.enrichment_windows_sids_independently_validates_6f7b5c9cc9` |
| `profile-services:enrichment.windows_sids` | `module.reference_data.integration.enrichment_windows_sids_independently_imports_ac_0a6c5f8b7f` |
| `windows-lifecycle` | `module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e` |
| `attack-lifecycle` | `module.reference_data.integration.attack_producer_artifacts_preserve_independent_a_67e9ef139a` |
| `windows-snapshot` | `module.reporting.integration.windows_event_ids_snapshots_retain_their_exact_i_7312056cfb` |
| `attack-snapshot` | `module.reporting.integration.attack_snapshots_retain_their_exact_independentl_a1740e14f8` |
| `windows-portable` | `module.incidentbundles.integration.windows_event_ids_pack_portability_preserves_ind_d31bf11266` |
| `attack-portable` | `module.incidentbundles.integration.attack_pack_portability_preserves_independent_re_a273637b25` |
| `windows-recovery` | `module.recovery.integration.windows_event_ids_recovery_restores_independentl_0983448cb7` |

This is a test-routing migration: callers of the former
`all_profiles_consumer_boundary` row must choose the specific profile and/or
producer lifecycle row. There is no compatibility alias that silently selects
other packs. Runtime profiles, public lookup contracts, historical pack bytes,
Base registry requirements and explicit activation semantics are unchanged.

### Isolation verification and repaired boundary failures

The following focused invocations exited 0. They preceded the final executable
source freeze; final-source qualification is recorded separately below. All use
`make test-slice OWNER=<owner> ROWS=<exact rows>` with the run-ID prefix and
retained-root convention above. Service rows additionally use `SCOPE=services`.
Rows named by scenario expand to the exact IDs in the independent selection table.

| Run | Owner / selected scenarios | Result |
| --- | --- | --- |
| `windows-independent-unit-01`, `windows-independent-unit-02` | `module.reference_data`: W and Windows Event IDs profile unit | Pass; no ATT&CK tests selected |
| `attack-independent-unit-01` | `module.reference_data`: A and ATT&CK profile unit | Pass; separate invocation |
| `windows-independent-ref-01` | `module.reference_data`: Windows profile consumer and lifecycle | Pass |
| `windows-independent-report-01` | `module.reporting`: Windows snapshot | Pass |
| `windows-independent-portable-01` | `module.incidentbundles`: Windows portability | Pass |
| `windows-independent-recovery-01` | `module.recovery`: Windows recovery | Pass |
| `attack-independent-ref-01` | `module.reference_data`: ATT&CK profile consumer and lifecycle | Pass |
| `attack-independent-report-01` | `module.reporting`: ATT&CK snapshot | Pass |
| `attack-independent-portable-01` | `module.incidentbundles`: ATT&CK portability | Pass |
| `windows-boundary-rootedfs-01` | `platform.rootedfs`: `platform.rootedfs.unit.operation_containment` | Pass |
| `windows-boundary-composition-01` | `module.incidentbundles`: source catalog, assembly and atomic-storage unit rows | Pass; HTTP fixture import cycle removed |
| `repository-check-catalog-repair-01` | `module.incidents`: `module.incidents.unit.incident_16_source_catalog_i3` | Pass; current format 7 expectation |

All 14 remaining profile unit rows ran separately and passed (exit 0), each
selecting only the matching `profile-unit` row above:

- `pack-type-registry-host-unit-01`
- `pack-type-registry-evidence-unit-01`
- `pack-type-registry-indicator-unit-01`
- `pack-framework-d3fend-unit-01`
- `pack-framework-veris-unit-01`
- `pack-enrichment-tor-unit-01`
- `pack-enrichment-cisa-kev-unit-01`
- `pack-enrichment-ms-portals-unit-01`
- `pack-enrichment-entra-app-ids-unit-01`
- `pack-enrichment-lolbas-unit-01`
- `pack-enrichment-loldrivers-unit-01`
- `pack-enrichment-lolesxi-unit-01`
- `pack-enrichment-hijacklibs-unit-01`
- `pack-enrichment-windows-sids-unit-01`

The old broad run `windows-online-check-final` exited 2 and did not publish its
final aggregate. It also overlapped later source edits and cannot qualify the
final implementation. Retained units exposed the fixture/application import
cycle, the offline-versus-runtime filesystem scanner scope, two stale bundle-6
test expectations, and database-migration failures. The first three categories
were repaired in this delivery. The isolated catalog diagnosis
`repository-check-catalog-diagnosis-01` also exited 2 before that repair.
Migration failures are recorded with the final broad run's outcome below.

`windows-independent-finalize-02` exited 2 because task-surface inputs required
regeneration before its schema check. Public `make generate` refreshed the
projections; subsequent finalization passed. No generated file was hand-edited.
The first Appendix audit (`windows-online-audit-appendix`) exited 0 but manual
review rejected its projection: an incidental provider name in a footer command
had incorrectly qualified event 4719 from a mixed current/legacy table. The
extractor and independent witness now exclude each unqualified monitoring row;
a synthetic regression includes the same incidental-name shape. That retained
intermediate audit is superseded and is not corpus acceptance evidence.

### Pre-repair executable-source evidence

Executable source was frozen after `windows-independent-finalize-04`
(`make agent-finalize`, exit 0). Its generated changes were reviewed. The broad
check binds source digest
`sha256:574d098486a546d481eb09f640d115a7a5f6b0326ab91bc56e0da16a604c178f`.
Later handoff prose edits are outside executable evidence inputs.

| Run | Exact command after run-ID prefix | Process exit / result |
| --- | --- | --- |
| `windows-independent-generate-03`, `windows-independent-generate-04` | `make generate` | 0, generated projections refreshed |
| `windows-independent-finalize-01`, `windows-independent-finalize-03`, `windows-independent-finalize-04` | `make agent-finalize` | 0; fourth is the final freeze |
| `windows-independent-generated-final` | `make source-check CHECK=generated` | 0, pass |
| `windows-independent-schema-final` | `make source-check CHECK=schema` | 0, pass |
| `windows-independent-builder-final` | `make reference-pack-producer-build` | 0, two byte-identical actual executable builds |
| `windows-independent-source-check-final` | `make reference-pack-source ACTION=check SOURCE_LOCK=<current-inventory> SOURCE_DIR=.cartulary/reference-packs/windows-online/acquisition-appendix` | 0, retained bytes and pins verified offline |
| `windows-independent-audit-final`, `windows-independent-audit-replay-final` | `make reference-pack-source ACTION=audit SOURCE_LOCK=<current-inventory> SOURCE_DIR=.cartulary/reference-packs/windows-online/acquisition-appendix` | 0, accepted; independent witness passed; byte-identical replay |
| `windows-independent-envelope-final` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/windows-online/independent-transport-inputs/envelope.json` | 0, independent byte/signature witness passed |
| `windows-independent-renew-final`, `windows-independent-renew-replay-final` | `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/windows-online/independent-transport-inputs/renew.json` | 0, independent witness and immutable renewal replay |
| `attack-independent-assemble-final`, `attack-independent-replay-final` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/attack/independent-regression-input.json` | 0, separate ATT&CK regression, independent witness and immutable replay |

`<current-inventory>` is
`.cartulary/reference-packs/windows-online/discovery-appendix/inventories/0dea611eb43496cf626322855d6c092e056981643d317f0f23475412d4c5a149.json`.
The original and intermediate acquisitions/audits remain immutable. The accepted
report and witness identities are in the current corpus table; the incorrect
intermediate Appendix projection is explicitly superseded.

`make target-plan TARGET=check DETAIL=rows JSON=1` exited 0 and selected all 16
independent profile unit/consumer pairs plus separate Windows and ATT&CK
lifecycle, snapshot and portability scenarios and Windows recovery. Planning
only establishes selection; the broad command outcome is recorded below.

### Readiness and exact remaining inputs

| Category | Verdict and closure condition |
| --- | --- |
| Producer | Implemented: bounded one-time acquisition, immutable offline checks, conservative reconciliation, canonical assembly, atomic publication and isolated fixture renewal. Final repository-wide outcome is separate below. |
| Corpus | Accepted documentary candidate: 308 retained sources; 391 identities; 569 qualified assertions; all other assertions explicitly accounted for; zero unresolved conflicts; independent whole-set witness and deterministic replay passed. This is not installed-Windows coverage. |
| Test trust | Windows has repository `fixture.windows.events.online.v1`, separate keys/reservations and its own qualification rows. Synthetic fixtures exercise ordinary trusted-wall-clock verification. Fixture authority is not deployment authority. |
| Deployment | Blocked: supply the actual publisher repository identity, release/version allocation and build timestamp, authorized signing process/key custody, and target trust-bootstrap configuration. The existing helper signs isolated fixtures only; it cannot approve or establish deployment authority. Then verify the real envelope through the existing production importer and choose activation explicitly. |
| Distribution | Blocked: supply a recorded source-specific license decision for exact lock `f8ab668f98e707a7e3cc87c9918369912901bbdb234ff18849bf3521daa0c1d8`, including decision reference/approver, license expression, redistribution classification and all required retained notice mappings. No default Microsoft-wide or third-party classification was assigned. |
| Existing-interface workflow | Implemented and exercised with explicitly synthetic Windows data: administration, exact-set lookup/provenance, snapshot binding, retained A after B, renewal, removal pins, portability and recovery. A real-corpus instance additionally needs the preceding licensing/publication inputs. Final-source test results are recorded below. |
| Incident enrichment workflow | Outside this delivery. No new incident importer, UI, report derivation or decoder was claimed. |

Real-pack assembly is deliberately unperformed until licensing is supplied;
therefore its manifest, manifest-framed logical payload digest, container,
bootstrap and exact pack-set identities remain absent. The candidate NDJSON
hash is not a substitute for any of those identities. No deployment, activation,
reset, historical conversion, website-liveness gate or Windows capture gate is
pending implicitly.

The local retained acquisition and builder directories must be preserved with
this handoff to reproduce the candidate without websites. They are ignored local
evidence, not a committed or approved distribution. An update is a new acquisition
and immutable release. Envelope renewal preserves logical bytes and provenance.

### Repository qualification blockers outside the pack change

The final broad run exposed these database-migration and PostgreSQL ACL failures.
`db/`, `internal/modules/database_migrations/` and `internal/platform/postgres/`
were not changed by this delivery.
The normalized unit result is `product/test_assertion_failure`, runner exit 10;
the outer check exit and final evidence completeness are recorded separately.

| Gap | Remediation / area | Rationale, benefit and unresolved risk | Compatibility impact / validation exit |
| --- | --- | --- | --- |
| `TestCatalogAppliesEveryVersionAndRejectsOrientationDowngrade_Integration` reports head 47 after a rejected downgrade | Investigate the disposable migration-scratch rollback path and reject a forbidden target before any downgrade effects; implementation/tests in the migrations owner | A refusal should preserve the starting ledger and managed schema; partial downgrade undermines the asserted cutover protection | Confirm current scratch-only rollback contract; no Windows-pack migration is needed. Exit: rejected downgrade preserves the initial head and exact catalog, and the owning integration row passes. |
| `TestProductionDDLCatalogManifestParity_Integration` reports one uncovered FK: `import_committed_source_records.import_committed_source_records_source_fk` (actual/manifest counts both 345) | Reconcile the FK's authored invariant coverage with its schema owner and regenerate downstream manifest projections; authored contracts/tests, implementation only if the owner reveals a schema defect | Exact counts alone do not prove invariant coverage; the missing binding prevents schema qualification | Preserve existing data semantics unless owner review establishes a defect. Exit: every actual FK has the required invariant coverage, manifest parity and the owning integration row pass. |
| `TestPostgresRolesOwnershipAndPrivileges_Integration` reports `public.import_committed_source_records` DELETE privilege false, expected true for `table_immutable` | Reconcile the table's normative access class, authored ACL projection and actual grants; specification/contracts/tests or migration as the owning security contract requires | Avoid changing privileges merely to satisfy a stale projection; unresolved mismatch blocks role/ACL qualification | Outside the Windows-pack change; do not grant DELETE without owner justification. Exit: expected access class and actual permissions agree, and `platform.postgres.integration.purpose_identity_roles_and_acl` passes. |

Exact evidence under `.cartulary/test-results/windows-independent-check-final/`:

- `unit-results/go-8d52ff9598edcaa3-2ad52c8a20d0.json`
- `unit-logs/go-8d52ff9598edcaa3-2ad52c8a20d0/stderr.log`
- `rows/module.database_migrations.integration.production_ddl_v3_recurrence.json`
- `unit-results/go-b1ba64269845e2ac-571637860284.json`
- `unit-logs/go-b1ba64269845e2ac-571637860284/stderr.log`
- `rows/platform.postgres.integration.purpose_identity_roles_and_acl.json`

These failures are not reclassified as source/licensing, Windows-provider or
pack-trust failures. Closing repository-wide qualification requires repairing
these owner defects and rerunning the applicable final portfolio on the resulting
frozen source. Test selection or an independently passing pack row cannot replace
that result.

### Scoped acceptance before static repairs

Under `windows-independent-check-final`, all 39 independent rows listed above
published `terminal_state:passed`, exit 0: 32 profile unit/consumer rows and seven
Windows/ATT&CK lifecycle, snapshot, portability and Windows recovery rows. The
Windows producer row W and separate ATT&CK producer row A also passed, exit 0.
Thus Windows Recovery is executed, not merely selected. Each profile consumer
starts isolated state; the repository-wide union does not add cross-pack setup.

The same final source passed the rooted-filesystem boundary row and both repaired
incident-source catalog/codec rows. Their results are retained under the run's
`rows/<exact-row-id>.json`; executable unit results and logs remain in
`unit-results/` and `unit-logs/`. These row results are scoped acceptance, not a
claim that the broad repository check passed.

Generic existing consumer validation, canonical profile rejection and portability
boundary coverage were reused. Windows-specific qualification adds the produced
fixture and exact identities; no parallel verifier, public route, incident field,
new UI or report derivation was introduced.

### Static repair and final requalification

`windows-independent-check-final` completed with outer Make exit 2 and a complete
failed summary: 338 work units, 332 passed, six failed, zero skipped/cancelled;
all 573 cleanup steps passed. This full-source result is retained, not replaced
by the subsequent focused passes. Its six failure groups were database migrations,
PostgreSQL ACL, Go security lint, Go staticcheck, backend architecture boundary,
and Mermaid rendering timeout. The pack rows themselves passed.

The producer/security and boundary findings were repaired after that run:

- Builder identity now reads Linux's kernel-bound `/proc/self/exe`, preventing pathname substitution from changing the claimed running executable; no broad filesystem suppression was added.
- Base producer staging uses private directory permissions. Dead acquisition/extraction assignments were removed while retaining original provider spelling; producer and one reporting-admission error prefix now meet existing lint policy.
- The authored boundary projection explicitly authorizes the exact new integration test files, distinguishes offline producer I/O from deployment-root consumers, and prohibits production imports of producer code or fixture subpackages. Test-only allowances do not widen runtime imports.
- ATT&CK's runtime-projection assertion was separated from source acquisition's temporary-file mutation test, keeping the generated-root write guard intact and keeping ATT&CK policy out of Windows tests.

Focused repair commands all exited 0:

| Run | Command after run-ID prefix | Outcome |
| --- | --- | --- |
| `windows-static-repair-format` | `make format` | Pass |
| `windows-static-repair-lint` | `make lint CHECK=go` | Pass, including staticcheck |
| `windows-static-repair-security` | `make go-gosec-targeted` | Pass |
| `windows-static-repair-boundary` | `make lint CHECK=backend-boundary` | Pass |
| `windows-static-repair-unit` | `make test-slice OWNER=module.reference_data ROWS=W` | Pass |
| `attack-static-repair-unit` | `make test-slice OWNER=module.reference_data ROWS=A` | Pass, separate invocation |

The renderer finding in unchanged renderer code remains `timing/timeout_failure`: the test
“Mermaid renders real graph geometry arrowheads and sequence messages as
repeatable passive SVG” reached its 30,000 ms limit. No timeout, renderer behavior
or golden was changed. The exact diagnostic is
`.cartulary/test-results/windows-independent-check-final/unit-logs/row-module.reporting.integration.renderer_mermaid/vitest-failure-details.json`.
Closure requires diagnosis of that renderer invocation and a passing owning row
under its existing budget; a timeout alone does not establish its cause.

The final portfolio requalifies the repaired boundaries and every independent
pack selection after finalization. It does not rerun the remaining migration/ACL or
renderer scenarios merely to obtain another broad failing result. Full-repository
qualification remains blocked by those separately described findings.

### Final repaired-source results

The final executable-source digest is
`sha256:fd5bd77533a9f9babe0a7113bfcd11afc54966a9658df045ae94e58a5416fd39`.
`windows-release-finalize` (`make agent-finalize`) exited 0 and reported no
additional generated changes. The successful Go lint, targeted security and
backend-boundary runs above have this exact digest and remain applicable.
All final pack invocations below also have this digest. Each exited 0.

The following table gives exact public Make commands. As throughout this handoff,
each command is prefixed with `CARTULARY_TEST_RUN_ID=<run>`; summaries and row
results reside under `.cartulary/test-results/<run>/`.

| Run | Command | Process exit |
| --- | --- | ---: |
| `pack-type-registry-host-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.type_registry_host_independently_validates_canon_2d41993e71,module.reference_data.integration.type_registry_host_independently_imports_activat_a3707c8ace` | 0 |
| `pack-type-registry-evidence-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.type_registry_evidence_independently_validates_c_ac918d851a,module.reference_data.integration.type_registry_evidence_independently_imports_act_046a55d962` | 0 |
| `pack-type-registry-indicator-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.type_registry_indicator_independently_validates_bc40be154e,module.reference_data.integration.type_registry_indicator_independently_imports_ac_6c2cb894b3` | 0 |
| `pack-framework-attack-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.framework_attack_independently_validates_canonic_147df00f1f,module.reference_data.integration.framework_attack_independently_imports_activates_cd0be06ccc,module.reference_data.integration.attack_producer_artifacts_preserve_independent_a_67e9ef139a,module.reference_data.unit.attack_producer_validates_pinned_all_domain_sour_79f71e137c` | 0 |
| `pack-framework-d3fend-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.framework_d3fend_independently_validates_canonic_420fac0262,module.reference_data.integration.framework_d3fend_independently_imports_activates_f8ef62f99f` | 0 |
| `pack-framework-veris-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.framework_veris_independently_validates_canonica_bb0bacb53d,module.reference_data.integration.framework_veris_independently_imports_activates_8ed7f795eb` | 0 |
| `pack-enrichment-tor-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_tor_independently_validates_canonical_e4e41a33af,module.reference_data.integration.enrichment_tor_independently_imports_activates_a_69bd8896bd` | 0 |
| `pack-enrichment-cisa-kev-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_cisa_kev_independently_validates_cano_9bc6dc025b,module.reference_data.integration.enrichment_cisa_kev_independently_imports_activa_c6cc146f1c` | 0 |
| `pack-enrichment-ms-portals-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_ms_portals_independently_validates_ca_57e34f2a64,module.reference_data.integration.enrichment_ms_portals_independently_imports_acti_860490d5c2` | 0 |
| `pack-enrichment-windows-event-ids-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_windows_event_ids_independently_valid_fc66475a5f,module.reference_data.integration.enrichment_windows_event_ids_independently_impor_2bc83a2b47,module.reference_data.integration.windows_event_ids_producer_artifacts_preserve_in_466944f93e,module.reference_data.unit.windows_online_producer_preserves_documented_ide_30fc80a930` | 0 |
| `pack-enrichment-entra-app-ids-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_entra_app_ids_independently_validates_02047fa20a,module.reference_data.integration.enrichment_entra_app_ids_independently_imports_a_07faf90a4a` | 0 |
| `pack-enrichment-lolbas-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_lolbas_independently_validates_canoni_397137c3ee,module.reference_data.integration.enrichment_lolbas_independently_imports_activate_f14dd266ff` | 0 |
| `pack-enrichment-loldrivers-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_loldrivers_independently_validates_ca_8d4da9f6a9,module.reference_data.integration.enrichment_loldrivers_independently_imports_acti_a068f18ffc` | 0 |
| `pack-enrichment-lolesxi-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_lolesxi_independently_validates_canon_5a1260a9d7,module.reference_data.integration.enrichment_lolesxi_independently_imports_activat_ebb18a1387` | 0 |
| `pack-enrichment-hijacklibs-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_hijacklibs_independently_validates_ca_4c78461ed3,module.reference_data.integration.enrichment_hijacklibs_independently_imports_acti_42ba3a94e9` | 0 |
| `pack-enrichment-windows-sids-release` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.enrichment_windows_sids_independently_validates_6f7b5c9cc9,module.reference_data.integration.enrichment_windows_sids_independently_imports_ac_0a6c5f8b7f` | 0 |
| `windows-snapshot-release` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.windows_event_ids_snapshots_retain_their_exact_i_7312056cfb` | 0 |
| `attack-snapshot-release` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.attack_snapshots_retain_their_exact_independentl_a1740e14f8` | 0 |
| `windows-portable-release` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.windows_event_ids_pack_portability_preserves_ind_d31bf11266` | 0 |
| `attack-portable-release` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.attack_pack_portability_preserves_independent_re_a273637b25` | 0 |
| `windows-recovery-release` | `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.windows_event_ids_recovery_restores_independentl_0983448cb7` | 0 |

This portfolio executed all 39 independent rows plus both producer rows, with
Windows and ATT&CK in separate invocations. It closes the scoped pack obligations
on the repaired source; it does not turn the earlier failed full check into a pass.

| Run | Command | Process exit / result |
| --- | --- | --- |
| `windows-release-common` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_pack_profiles` | 0, shared consumer/profile validation |
| `windows-release-root-boundary` | `make test-slice OWNER=platform.rootedfs ROWS=platform.rootedfs.unit.operation_containment` | 0, containment/runtime boundary |
| `windows-release-incident-source` | `make test-slice OWNER=module.incidents ROWS=module.incidents.unit.incident_16_source_catalog_i3,module.incidents.unit.incident_16_source_codec_i3` | 0, adopted bundle generation and source contracts |
| `windows-release-generated` | `make source-check CHECK=generated` | 0 |
| `windows-release-schema` | `make source-check CHECK=schema` | 0 |
| `windows-release-builder` | `make reference-pack-producer-build` | 0, reproduced executable retained |
| `windows-release-source-check` | `make reference-pack-source ACTION=check SOURCE_LOCK=<current-inventory> SOURCE_DIR=.cartulary/reference-packs/windows-online/acquisition-appendix` | 0, offline |
| `windows-release-audit`, `windows-release-audit-replay` | `make reference-pack-source ACTION=audit SOURCE_LOCK=<current-inventory> SOURCE_DIR=.cartulary/reference-packs/windows-online/acquisition-appendix` | 0, same accepted report and independent witness |
| `windows-release-envelope` | `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/windows-online/release-transport-inputs/envelope.json` | 0, independent witness |
| `windows-release-renew`, `windows-release-renew-replay` | `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/windows-online/release-transport-inputs/renew.json` | 0, independent witness and immutable replay |
| `attack-release-assemble`, `attack-release-replay` | `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/attack/release-regression-input.json` | 0, separate ATT&CK witness and immutable replay |

`<current-inventory>` expands to the full path given in the operation procedure.
The exact final report is
`acquisition-appendix/audits/07075debab330fe7a19dea57e496f3f100e8db7c06d612282042edfc5e114c80/`
under `.cartulary/reference-packs/windows-online/`; its retained bytes, hash and
builder identities are in the current corpus table. No remote source was queried
during final checks, audits, builds, witnesses or application tests.

A further full `make check` was not repeated after the static repairs: the changed
producer, boundaries and all independent pack obligations were requalified as
above, while the migration/ACL assertions and renderer timeout remain unresolved.
The timeout's relationship to workload or resource contention was not established.
Thus there is no passing final full-check or release-qualification claim.
Retained-run maintenance requiring a compatible passing full warm-check was
skipped because no such root exists. UI review, visual golden updates, deployment,
real-pack activation and new incident enrichment were outside this delivery.

Closing prose validation: `CARTULARY_TEST_RUN_ID=windows-release-markdown make lint-markdown`
exited 0, with summary at
`.cartulary/test-results/windows-release-markdown/adhoc/lint-markdown/tool-run-summary.json`.
The final wording is checked with the same public target under run ID
`windows-release-markdown-final`; its summary uses the same relative location.
`git diff --check` also passed. No commit, push, publication or activation was performed.

## Remediation closure follow-up — 2026-11-08

The [RC closure section](reference_data-module-refactor-tracker.md#rc--remediation-closure-2026-11-08)
records the corrected authority and resource rules, Linux filesystem-isolation
prerequisites, final producer pins and scoped application qualification. Assembly
now retains research under the explicit output root, leaving its pinned source
corpus read-only. Existing signed release descriptors, notices, distribution
decisions and signing custody remain unchanged. The new unsigned corpus outputs
are qualification candidates, not replacement releases.
