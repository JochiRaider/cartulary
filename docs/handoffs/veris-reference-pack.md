# VERIS Community reference pack handoff

The real VERIS Community 1.4.2 pack is delivered under
[`fixtures/veris/`](../../fixtures/veris/). The user's later instruction replaced
local-store installation with fixture delivery. No development deployment,
existing catalog, trust configuration, database, or object-store volume was
changed. Installation, live catalog inspection, and development-stack restart
are therefore outside this delivery. Isolated integration tests exercise the
existing import, activation, lookup, reporting, portability, and recovery owners.

The entry point is [`release.v1.json`](../../fixtures/veris/release.v1.json).
It selects the current logical tree and renewed signed envelope. The adjacent
previous envelope is retained as renewal evidence. Intermediate implementation
candidates are outside the fixture tree, under
`.cartulary/reference-packs/veris/interim-candidates/`.

## Decisions and ownership

The explicit instruction to implement the supplied plan authorized the concrete
native-profile amendment. Reference Pack RP-REQ-270–274, its RP-REQ-018/019
versioning exception, Core 00 adoption, the Extensions companion statement, and
domain vocabulary now describe that amendment. The supplied research document
was evidence, not an independent source of implementation authority.

| Workstream / phase | Change and gate |
| --- | --- |
| A / P0–P2 | Reconciled current owners; adopted native VERIS v2; regenerated schemas, registry, consumer union, and fixtures. Predecessor VERIS v1 fails incoming and retained admission. |
| B / P1–P3 | Pinned eleven sources; reviewed source-unit dispositions; reconciled schema and vocabulary; implemented closed acquisition, offline projection, audit, and assembly. Independent JSON expectations reconstruct the full schema and definitions. |
| B/C / P4 | Production semantic validation precedes sequence allocation. Existing archive, publication, indexing, and TUF implementations are reused. Dedicated release and disposable test authorities remain separate. |
| C / P5 | Qualified real-corpus lifecycle, consumer provenance, replacement/history, reporting, portable embedding and rejection, recovery, and server/operator predecessor admission. |
| D / P6, revised | Delivered canonical content, signed archives, public bootstrap, and witnesses in the root fixture directory, as requested. Local deployment and UI-review work were removed from scope. |
| A–D / P7 | This handoff records commands, custody, evidence, limitations, update and rollback procedures. The initial baseline failures and their subsequent repairs are recorded below. |

The current profile is `cartulary.reference_pack.framework.veris.v2`, version
`"2"`, with object and relationship schema v2. Shared pack contract v2,
administrative major 3, manifest v1, immutable-set/reference catalog v2, and
incident-bundle format 7 remain unchanged. Unrelated profile identities and Base
content are unchanged. Old VERIS schema bytes exist only under historical test
fixtures; there is no conversion, migration, reset, or dual reader.

The Reference Data owner maintains the producer and semantic admission.
Reporting, incident bundles, Recovery, server, and operator retain their existing
consumer and lifecycle responsibilities. The local publisher custodian controls
the signing context and monotonic ledgers described below.

## Source and corpus accounting

The source revision is `vz-risk/veris@dba30b035039f3d002c6eb06dc149dbed7476059`,
VERIS Community 1.4.2, source time `2026-10-14T19:34:35Z`.
The immutable source profile is `cartulary.veris.community.native.v1`.
[`inventory.v1.json`](../../contracts/reference-pack-producers/veris/inventory.v1.json)
records exact immutable URLs, lengths, media types, SHA-256 values and roles for
the base, enum, labels and merged schemas; changelog; license; validation, merge
and enrichment scripts; short guide; and complete 172-page coding guide.

| Record / relationship kind | Count |
| --- | ---: |
| Schema root | 1 |
| Schema nodes below the root | 260 |
| Enum locations | 68 |
| Contextual enum values and exact definitions | 2,168 |
| Inert semantic rules | 62 |
| Total objects | 2,491 |
| `contains` | 200 |
| `has-item-schema` | 60 |
| `allows-value` | 2,168 |
| `governed-by-rule` | 105 |
| Total relationships | 2,533 |

Source admission reconciles base versus merged schema without enumerations,
exact enum arrays and order, and all contextual label definitions. Native
keyword presence and values reconstruct the parsed upstream merged schema
exactly. The independent Node witness compares retained machine-readable
baseline JSON, not the Go projection or source prose. It also checks every
contextual definition, curated rule, relationship tuple, and manifest source pin.

[`dispositions.v1.json`](../../contracts/reference-pack-producers/veris/dispositions.v1.json)
accounts for all eleven sources and 198 source units, including every PDF page
and relevant script functions. Applicable coding, validation and enrichment
guidance becomes categorized inert statements. The full guide's extracted text
was reviewed; quoted incident cases, illustrations, WebApp instructions, and
VCDB-only material have explicit exclusions. Research scripts and the guide are
retained as evidence, not redistributed as runtime members or executed.

Preserved discrepancies and interpretations include:

- The upstream `schema_version` default and description still name 1.4.0; release
  provenance identifies 1.4.2.
- The observed closed schema vocabulary contains **18** keywords, correcting the
  plan's count of 19. The nonstandard `definition` annotation is retained.
- Unknown-actor prose does not add machine-required fields. Missing, Unknown,
  Other, and context-specific NA remain distinct; Namibia is not Not Applicable.
- Guide v3.0 targets schema 1.4.1 and uses VCDB examples. It cannot override the
  Community 1.4.2 schema. Open `plus`, object arrays, multiple action/actor
  branches, and agentic-system guidance remain source-backed.
- Region-code disagreement, misuse prose versus executable behavior, the
  defacement availability issue, and stale enrichment field references are
  qualified in the rules and dispositions rather than silently reconciled.

The native limits are 8,192 Unicode scalars per native string, 1,024 per pointer,
512 enum members, 64 child properties/required names, structural depth 32 with
root at zero, and the existing 1 MiB complete record limit. Stricter common-field
limits still apply. Unknown keywords and overflow fail without truncation.

IDs bind the Community namespace, kind, exact source pointer and, for enum
values, exact raw string value. Rules have curated stable identifiers. Labels and
release versions never define identity. Qualified aliases exclude all colliding
labels; exact object IDs remain available. Future incompatible meaning changes
require the explicit semantic-key decision described in
[`continuity.v1.json`](../../contracts/reference-pack-producers/veris/continuity.v1.json).

## Licensing and distribution

The upstream license is CC BY-SA 4.0. The exact text and attribution/adaptation
notice are included as declared, hashed runtime notice files. The explicit
project disposition is
[`license.project.v1.json`](../../contracts/reference-pack-producers/veris/license.project.v1.json),
recording the user's implementation authorization and adapted-content scope.
Its reference to the original installation plan records that authorization's
history; the later fixture-only instruction controls delivery scope.

Assembly requires the disposition and includes its digest-bound decision as
source provenance. Redistribution is `allowed` for the adapted Community schema,
vocabulary/definitions, and original descriptive rule summaries under CC BY-SA
4.0. It does not inherit Entra's Apache policy. PDF images, quoted incident cases,
scripts, and incident data are not embedded. Keep the exact notices with every
redistributed pack and portable embedding. An unresolved disposition for a
future source blocks distributable assembly and must not be bypassed by changing
only the manifest's redistribution flag.

## Release identities and retained evidence

Release `1.4.2-20261107.3` has publication sequence **3** and publisher
`cartulary.veris.community`. Build time is `2026-11-08T00:02:54Z`.

| Identity | SHA-256 |
| --- | --- |
| Manifest | `6c0091603e7fb0144ce0ae4a4984250f00c989f129d1d0cf641363baab514876` |
| Canonical payload | `7350c4ef1329e40060a79cb1ae05f63513c11406309de22a109cdda2ae2386d9` |
| Current signed container | `ae2a3c6f4a8dba62016829723e5da17c97fcc6d1682cb48b45de5baa4729518e` |
| Public bootstrap | `16c8d7bd756eb863ab6eddf2c70e884ae8da9d64cb5b434832b9fbb870e52065` |
| Producer executable | `098a3f392e18be3506ec7d29ca0b3fb3e885cbf044ea2e6b5cb4366e221b2217` |
| Source profile | `a14b13314a69ff4739debb2f9827c95d3374f1c78ecd2d98d07b0c8cf225fa72` |
| Source inventory | `ec762619cb6f9ef639971736bd05361f8eb6b4a485e46d21a3ec5a6258fbb089` |
| Source lock | `abdfc2f87646cd225e6bd905bff61a555b2a1f5f066be719d1a1cfd1eaca0322` |
| Source audit | `f6c847171ee4e317ab9849403c65d2b6daf5c2e6f2b18924ce9a9b985bf201b6` |
| Reconstructed merged schema, JCS | `5062a1e2a8b449438b8053f02985e9cf0878bfcabde9d519897f9bc248fbdc1d` |

The archive is 3,544,576 bytes. Canonical objects occupy 1,978,402 bytes;
relationships occupy 1,535,090 bytes. Detailed member hashes and lengths are in
the logical witness and manifest. The envelope witness independently checks
archive contents, public signatures and metadata bindings.

Current envelope metadata version **5** was issued `2026-11-08T00:03:43Z`.
Online metadata expires `2026-12-09T00:03:43Z`; the root expires
`2027-11-07T23:46:01Z`. Version 4 is retained under container digest
`47bae1cdc8f090fa4f4006df72bbffdd2f680919d1bafc2aad54529c96207fe7`.
Renewal preserved manifest, payload, and bootstrap bytes.

Acquired sources, exact receipts, audit, license decision, source profile, rule
inventory and continuity evidence are retained locally under
`.cartulary/reference-packs/veris/source/`, with release evidence in
`publications/<manifest-sha256>/`. Original research and guide extraction remain
under `/tmp/cartulary-veris-research-20261107/`; that temporary directory is not
required by tests, runtime, witnesses, or rebuilding from retained sources.

## Reproduce and refresh

Run commands from the repository root. All managed producer and verification
operations use Make. Closed source/build/envelope input types are implemented in
`internal/modules/reference_data/producer/`; do not add arbitrary input members.
No product test reads Cartulary documentation or upstream prose.

```sh
make task-guide OWNER=module.reference_data
make reference-pack-producer-build
make reference-pack-source ACTION=acquire \
  SOURCE_LOCK=contracts/reference-pack-producers/veris/inventory.v1.json \
  SOURCE_DIR=.cartulary/reference-packs/veris/source
make reference-pack-source ACTION=check \
  SOURCE_LOCK=contracts/reference-pack-producers/veris/inventory.v1.json \
  SOURCE_DIR=.cartulary/reference-packs/veris/source
make reference-pack-source ACTION=audit \
  SOURCE_LOCK=contracts/reference-pack-producers/veris/inventory.v1.json \
  SOURCE_DIR=.cartulary/reference-packs/veris/source \
  BUILT_AT=2026-11-08T00:02:54Z
make reference-pack-package ACTION=assemble \
  INPUT=.cartulary/reference-packs/veris/build-release.json
make reference-pack-package ACTION=assemble \
  INPUT=.cartulary/reference-packs/veris/build-release-reproduction.json
```

Acquisition is the only network operation. Existing retained source bytes can be
checked and built offline. The two build inputs differ only in output root;
their five canonical files, including the manifest, compared byte-identical.
The reproduction output is
`.cartulary/reference-packs/veris/reproduction-release/logical/<manifest-sha256>/`.

The release build input is reproducible from this closed example; substitute
absolute paths for another checkout. A later build with changed builder, source
or build-time identity needs a new release name rather than overwriting the
existing ledger allocation.

```json
{
  "schema_id": "cartulary.veris_build_input.v1",
  "source_root": "/home/jochi/code/cartulary/.cartulary/reference-packs/veris/source",
  "source_inventory": "/home/jochi/code/cartulary/contracts/reference-pack-producers/veris/inventory.v1.json",
  "output_root": "/home/jochi/code/cartulary/fixtures/veris/1.4.2-20261107.3",
  "ledger_root": "/home/jochi/code/cartulary/.cartulary/reference-packs/veris/ledger",
  "built_at": "2026-11-08T00:02:54Z",
  "pack_version": "1.4.2-20261107.3",
  "builder_sha256": "098a3f392e18be3506ec7d29ca0b3fb3e885cbf044ea2e6b5cb4366e221b2217",
  "trust_repository_id": "cartulary.veris.community"
}
```

For a source refresh, pin the new commit and every exact source hash; review
schema differences, all guidance units, definitions, licensing and semantic
continuity. Update the frozen profile, independent baseline, curated rules and
dispositions deliberately. Rebuild through Make, rerun the affected qualification
portfolio, publish a new release, and update `release.v1.json` only to verified
artifacts. Do not silently retain an ID for a different meaning.

## Signing custody and renewal

The dedicated private context is outside tracked/public artifacts at
`/home/jochi/.local/share/cartulary/reference-pack-signing/veris/v2/signing-context.json`.
Its directory is mode 0700 and file mode 0600. Preserve it in the custodian's
protected backup together with publication and envelope ledgers. No secret bytes
belong in this handoff, fixtures, logs, or commits. The public bootstrap can be
distributed with the archive. The separate `fixture.veris.community.v2` authority
is generated in memory for isolated tests and is never deployment trust.

The nonsecret envelope input is retained at
`.cartulary/reference-packs/veris/envelope-release.json`; the executed renewal
input is `.cartulary/reference-packs/veris/renewal-release.json`. Both have exactly
these fields: `schema_id`, `logical_root`, `output_root`, `signing_context`,
`create_signing_context`, `issued_at`, `root_expires`, `online_expires`,
`metadata_version`, `ledger_root`, and `previous_envelope`. The schema ID is
`cartulary.veris_release_envelope_input.v1`; creation is false for existing keys.

For the next renewal, copy the renewal input to a new local file, keep the exact
logical root and signing context, use a fresh issue time and future online
expiry within root validity, increase metadata version above 5, and point
`previous_envelope` at the currently selected envelope. Keep
`ledger_root=.cartulary/reference-packs/veris/envelope-ledger` as an absolute path.

```sh
make reference-pack-package ACTION=renew \
  INPUT=.cartulary/reference-packs/veris/renewal-next.json
```

Inspect the new envelope report and witness before updating the fixture
descriptor. Renewal must preserve logical content, release sequence, and public
root; it changes only the signed envelope. Expired metadata must be renewed
before a new import. Do not recreate keys to bypass expiry or replay protection.
Root rotation requires the existing shared trust procedure.

## Verification and limitations

The baseline was clean commit `ef2048497e836dd58b112c2d0d003ec703ee865a`.
Evidence below records working-tree implementation runs, not a committed release
or a complete-owner evidence audit. Every listed pass includes observed outer
Make exit 0. Failures had outer Make exit 2. Run IDs below are exact children of
`/home/jochi/code/cartulary/.cartulary/test-results/`; each graph run's
`target-summaries/test-slice.json` and `rows/` hold its selection and results.

| Command / scope | Run ID | Result |
| --- | --- | --- |
| `make reference-pack-source ACTION=acquire` with the inventory/source paths above | `20261107T234359Z-p46441` | Pass; exact eleven-source acquisition. |
| `make reference-pack-producer-build` | `20261108T000200Z-p68313` | Pass; immutable builder reproduced. |
| `make reference-pack-source ACTION=audit` with the release build time | `20261108T000254Z-p69330` | Pass; reconciliation and independent witness. |
| `make reference-pack-package ACTION=assemble INPUT=.cartulary/reference-packs/veris/build-release.json` | `20261108T000255Z-p69547` | Pass; production validation and publication. |
| Same assembly with `build-release-reproduction.json` | `20261108T000322Z-p69845` | Pass; second offline canonical build. |
| `make reference-pack-package ACTION=envelope INPUT=.cartulary/reference-packs/veris/envelope-release.json` | `20261108T000333Z-p70128` | Pass; envelope 4. |
| `make reference-pack-package ACTION=renew INPUT=.cartulary/reference-packs/veris/renewal-release.json` | `20261108T000343Z-p70367` | Pass; envelope 5, unchanged logical identities. |
| `make test-slice OWNER=module.reference_data ROWS=<all 42 unit rows>` | `20261108T000417Z-p84848` | 40 rows passed. Two failed: stale v2 composite hash expectations, subsequently repaired; unrelated existing assembly construction failure, described below. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest` | `20261108T000910Z-p98943` | Pass after independently recomputing complete-set and portable-catalog expectations with Python JSON/SHA-256. |
| Reference Data real release and synthetic successor lifecycle rows | `20261108T000436Z-p88320` | Pass, both rows. |
| Reporting `veris_snapshot_binding` | `20261108T000436Z-p88319` | Pass. |
| Incident bundles `veris_bundle_round_trip` | `20261108T000436Z-p88331` | Pass; allowed and restricted derivatives. |
| Recovery `veris_recovery` | `20261108T000436Z-p88342` | Pass; exact active and historical content recovery. |
| Server retained admission | `20261108T000436Z-p88373` | Pass; predecessor fixtures rejected before reconciliation. |
| Operator retained admission | `20261108T000436Z-p88374` | Pass. |
| `make source-check CHECK=all` | `20261108T000401Z-p70752` | Pass before final expectation-only repair; final source check recorded below. |
| `make lint CHECK=shell` | `20261108T000401Z-p70768` | Pass. |
| `make lint CHECK=scripts` | `20261107T235833Z-p90883` | Pass. |
| `make lint CHECK=go` | `20261108T000401Z-p70755` | Format passed; vet/staticcheck failed on baseline defects below. |
| `make go-gosec-targeted` | `20261107T235833Z-p90885` | Failed on baseline compile defect; general profile reported zero issues, later profiles did not complete. |

The final expectation-only repair does not change production code or pack bytes.
Earlier-source run identities remain visible; final freeze and selected reruns
are recorded below rather than represented as a full-owner acceptance audit.

Final `make agent-finalize` passed at `20261108T001149Z-p1074`. The following
selected portfolio then passed with source digest
`sha256:af4528982a5fb14a17e2521b6b9ffe3db5891fb964e0497cdf7694dd2b21bee2`.
All seven test invocations exited 0; their retained cleanup had zero failed or
blocked steps. The selection contains 41 unit rows and seven integration rows.
The already diagnosed baseline construction row was deliberately excluded from
that repeat. The follow-up below repairs it and records its passing execution.

| Final command / selection | Run ID | Result |
| --- | --- | --- |
| Reference Data unit family, excluding `application_construction` | `20261108T001246Z-p6711` | All 41 selected rows passed. |
| Reference Data real release and synthetic lifecycle | `20261108T001246Z-p6655` | Both rows passed. |
| Reporting VERIS snapshot | `20261108T001246Z-p6665` | Pass. |
| Incident bundles VERIS round trip | `20261108T001246Z-p6676` | Pass. |
| Recovery VERIS | `20261108T001246Z-p6701` | Pass. |
| Server retained admission | `20261108T001246Z-p6709` | Pass. |
| Operator retained admission | `20261108T001246Z-p6751` | Pass. |
| `make source-check CHECK=all` | `20261108T001246Z-p6785` | Pass, including generated/schema/catalog boundaries. |
| `make reference-pack-producer-build` | `20261108T001246Z-p6789` | Pass; executable digest still matches the signed release. |
| `make lint-markdown` | `20261108T001246Z-p6808` | Pass for the handoff and owner amendments. |

To repeat the focused integration selection:

```sh
make test-slice OWNER=module.reference_data \
  ROWS=module.reference_data.integration.framework_veris_independently_imports_activates_8ed7f795eb,module.reference_data.integration.veris_real_release_import
make test-slice OWNER=module.reporting \
  ROWS=module.reporting.integration.veris_snapshot_binding
make test-slice OWNER=module.incidentbundles \
  ROWS=module.incidentbundles.integration.veris_bundle_round_trip
make test-slice OWNER=module.recovery \
  ROWS=module.recovery.integration.veris_recovery
make test-slice OWNER=app.server \
  ROWS=app.server.integration.reference_pack_retained_admission
make test-slice OWNER=app.operator \
  ROWS=app.operator.integration.reference_pack_retained_admission
make source-check CHECK=all
make lint-markdown
```

Use `make task-guide OWNER=module.reference_data DETAIL=rows` for the current
unit row IDs. The exact executed unit list is retained in the final unit run's
`run-manifest.json` under `declared_inputs.ROWS`; it selects the unit family
except the diagnosed baseline row. Finalize before running an expanded portfolio.
Full warm-run evidence was not supplied to `agent-finalize`, so its optional
retained-run maintenance checks were skipped as reported by that target.

Five isolated source-admission operations additionally exercised the current
producer against damaged copies of the retained source tree. These intentionally
failed `make reference-pack-source ACTION=check` with outer exit 2 (producer
exit 1), classified by that tool wrapper as `harness/unknown_failure`. Their
specific producer diagnostics establish the expected rejection; they are not
successful harness runs or product tests reading upstream guidance.

| Negative source input | Run ID | Observed rejection |
| --- | --- | --- |
| Changed file length | `20261108T001411Z-p20192` | Expected regular-file/length evidence mismatch. |
| Missing pinned file | `20261108T001411Z-p20195` | Missing source file. |
| Extra undeclared file | `20261108T001411Z-p20193` | Undeclared VERIS source member. |
| Wrong version/inventory | `20261108T001411Z-p20198` | Inventory differs from frozen profile. |
| Same-length byte substitution | `20261108T001446Z-p21268` | Source hash or length mismatch. |

These inputs are retained under
`.cartulary/reference-packs/veris/admission-negatives-20261108/` as `changed`,
`missing`, `extra`, `digest`, and `wrong-inventory.json`. Use the normal inventory
as `SOURCE_LOCK` and the corresponding copied directory as `SOURCE_DIR`; for the
wrong-inventory case use that JSON as `SOURCE_LOCK` and the original source root.
The original retained source tree was not modified.

The initial baseline defects were confirmed against `git show HEAD:<path>`.
Their original runs remain retained with `product/test_assertion_failure`;
cleanup completed without failed or blocked steps. The subsequent user request
authorized repairing them. All of those blockers are now resolved:

- `tools/packagerecovery/main.go` constructs the required verification lifetime
  using configured cancellation grace and supplies it to recovery storage. Its
  fatal sink emits the safe integrity diagnostic and exits 70 when grace is
  exhausted, preserving the existing owner contract for active cleanup.
- `TestReferencePackAdministrationConstruction_Unit` supplies a real test
  lifetime to both the coordinator and admission-only constructors. Both paths
  explicitly reject an omitted lifetime; successful construction still exercises
  worker registration and the admission-only dependency boundary.
- The Extensions test uses a named context-key type (resolving SA1029). The
  incident-bundle test infers the already-returned finalizer interface type
  (resolving ST1023). Neither change alters the tested cancellation behavior.

### Follow-up verification — 2026-11-07

These repairs change four Go files and this handoff. The VERIS source profile,
canonical corpus, signed archive, bootstrap, release selection, and signing keys
are unchanged. No development-store operation was performed.

`make format` passed at `20261108T022305Z-p50100`. Focused admission and portable
tests first passed at `veris-followup-admission-20261107` and
`veris-followup-portable-20261107`. The first assembly retry,
`veris-followup-assembly-20261107`, exposed the second missing lifetime in its
admission-only constructor (outer exit 2, `product/test_assertion_failure`).
After correcting that call, fresh assembly execution passed with outer exit 0
at `veris-followup-assembly-fixed-20261107`. Both the failed attempt and the
successful repair evidence are retained.

Final `make agent-finalize` passed without changing generated files. Its optional
retained-run checks remained unselected because no full warm-check root was
supplied. The final portfolio below shares source digest
`sha256:ba81aa1e8ccca8bb2d956bade0f1146b8e54761e697cbc1ba3ce840eb22c47d3`.
Every listed command completed with observed outer exit 0.

| Command / selection | Exact run ID under `.cartulary/test-results/` | Result |
| --- | --- | --- |
| `make agent-finalize` | `veris-followup-finalize-20261107` | Pass; generated files unchanged. |
| `make lint CHECK=go` | `veris-followup-go-lint-20261107` | Pass, including formatting, vet, and staticcheck. |
| `make go-gosec-targeted` | `veris-followup-security-20261107` | Pass; the previously blocked target completed. |
| `make source-check CHECK=all` | `veris-followup-source-20261107` | Pass, including generated/schema/catalog checks. |
| Reference Data `application_construction` | `veris-followup-final-assembly-20261107` | Pass; final invocation retained the compatible result with no test process launched. Fresh execution is recorded above. |
| Extensions `admission_invocation_v6` | `veris-followup-final-admission-20261107` | Pass. |
| Incident bundles `reference_pack_fatal_completion` | `veris-followup-final-portable-20261107` | Pass. |

The Go lint, targeted security, and portable integration runs' retained cleanup
has zero failed or blocked steps. Their summaries are respectively
`target-summaries/lint.json`, `target-summaries/go-gosec-targeted.json`, and
`target-summaries/test-slice.json` within the roots above. Finalizer details are
in `unit-artifacts/finalize-summary.json`.

Reproduce the affected checks with fresh run IDs:

```sh
make test-slice OWNER=module.reference_data \
  ROWS=module.reference_data.unit.application_construction
make test-slice OWNER=module.extensions \
  ROWS=module.extensions.unit.admission_invocation_v6
make test-slice OWNER=module.incidentbundles \
  ROWS=module.incidentbundles.integration.reference_pack_fatal_completion
make agent-finalize
make lint CHECK=go
make go-gosec-targeted
make source-check CHECK=all
make lint-markdown
```

The final handoff Markdown check uses run ID
`veris-followup-markdown-20261107`; its result is retained at
`.cartulary/test-results/veris-followup-markdown-20261107/adhoc/lint-markdown/tool-run-summary.json`.
The earlier failed lint, security, and construction results are historical;
they no longer describe current blockers. This follow-up does not claim a full
repository release gate or complete-owner audit. Packaged deployment recovery
qualification was not repeated: the fixture tool's dependency wiring is covered
by Go compilation/static analysis, while the affected owner tests were selected
explicitly. The prior corpus and consumer qualification remains identified by
its original source and run roots above.

Focused coverage includes exact reconstruction, all definitions, contextual
Unknown/Other/NA, graph identities/reachability/direction/provenance, hostile
records, bounds, cancellation, current-profile admission, archive and trust
regressions, staged-only import, explicit activation/disablement, qualified
lookup, no-hit/cursor errors, historical bindings, allowed/restricted embedding,
destination inactivity, and active/pinned recovery. Existing shared-owner cases
cover trust expiry/signature failures, storage interruption, and policy behavior.
The lifecycle checks sample entries through the consumer; the corpus witness
checks every canonical record. A clearly labeled fixture successor changes one
definition only to exercise replacement; it is not an upstream VERIS release.

Development-store installation, live catalog UI/non-admin observation, and
development restart were not performed because the user explicitly removed
local-store work. No frontend behavior was changed. A future deployment must
register the public bootstrap while preserving other trust, enable the existing
profile, import through administration, wait for verification, then explicitly
activate. Record actual job/set identities and restart evidence at that time;
the package alone is not evidence of an installed pack.

## Rollback and recovery

For a future deployment's first VERIS release, disable it through the existing
administration lifecycle. If compatible releases have been retained, explicitly
reactivate the selected prior release. Preserve historical set pins, signed
containers, provenance, and notices. Never rewrite historical sets or reset data
to bypass incompatible VERIS v1 admission. Recovery must retain the exact active
and pinned content or fail through the existing owner behavior.

For fixture selection, retain old immutable release trees and change only the
descriptor after qualifying the selected release and its unexpired envelope.
Do not delete signing ledgers or regenerate publisher keys as a rollback step.

## Remediation closure follow-up — 2026-11-08

The [RC closure section](reference_data-module-refactor-tracker.md#rc--remediation-closure-2026-11-08)
records the corrected authority and resource rules, Linux filesystem-isolation
prerequisites, final producer pins and scoped application qualification. Assembly
now retains research under the explicit output root, leaving its pinned source
corpus read-only. Existing signed release descriptors, notices, distribution
decisions and signing custody remain unchanged. The new unsigned corpus outputs
are qualification candidates, not replacement releases.
