# Reporting coordinated adoption record

## Decision and applicability

Status: **reviewed candidate; promotion pending integrated validation**.

The user authorized coordinated Reporting 1.4.0 and Report Composition 1.3.0
adoption, a clean pre-production cutover and removal of unnecessary compatibility
behavior. Codex performed the behavioral and execution/evidence reviews under
that instruction. This is not a claim of independent human approval. These
records are supporting material, never executable validation inputs.

The immutable comparison texts are Reporting 1.3.0 and Composition 1.2.0 at
`78a4effc23aac4b46ca414e11b91e17892f790c7`, at their current document paths.
Their unchanged obligations remain the review baseline. Amended expectations
below are candidate coverage until coordinated promotion; an old test conflicting
with a proposal was not automatically classified as obsolete. The implementation
started from clean `00d24aa5354284dfcf6533b8f650e9eb0291ad44`.

The candidate applies to the `snapshot_reporting` profile, ordinary composition
authoring and internal previews, and the current Slidev PDF/Mermaid SVG release
paths. It does not claim a report-builder UI, PNG/PPTX rendering, legacy readers,
historical identity reconstruction, or Core 05 benchmark publication. Native
packaged evidence currently qualifies Linux amd64 on WSL2 with the declared
no-LSM confinement profile. It does not qualify other host/LSM configurations.

The [historical implementation review](reporting-cutover-adoption-review.md)
retains failed roots, individual repairs and their source applicability. Its
initial partial-status table is historical. The following register is the current
review of the complete amended clause set and its assurance families.

## Changed-clause dispositions

Clause identifiers in this table have the `REQ-RPT-` prefix unless marked `RC`.
All listed clauses are reviewed candidate amendments; no clause is deleted.
An entry includes the tables attached to its clause and editorial transitions
before the next requirement. The right column names the assurance family in the
next table, not an executable requirement-to-test contract.

| Changed clauses | Final decision and compatibility impact | Assurance family |
| --- | --- | --- |
| 005, 006, 014, 018, 019, 019a, 020, 021 | Coordinate authority, owner interfaces and current identities. Closed identifiers replace retired formats without aliases. | Authority; owner boundaries; admission |
| 025, 026d, new 026e–026g | Accept an asynchronous snapshot Job promptly; capture immutable source once, retain exact compact capture and independently bound recipe, and validate before publication outside the terminal lock. No live materialization during render replay. | Snapshot lifecycle; frozen inputs |
| 027, 027a–027e, 028a, 057a | Resolve exact derivation and authored profile identity once; internal omission materializes the explicit passthrough selector. Worker preparation checks recipients against retained policy and source, preserving specific terminal reasons without registry substitution. External release still requires the exact pair and permitted recipients. | Admission; frozen inputs; redaction |
| 029, 031 | Current output scope is Slidev/Mermaid. PDF and internal source-only options are explicit; unsupported rendered outputs fail. | Output options; native rendering |
| 038, 039, new 039a, 040, 040b, 041, 042a, 042b | Canonical retained model, separate snapshot/release/preview identities and immutable time; closed field states and structural identities. Never infer missing historical facts. | Canonical identity; retained content |
| 043a, new 043b, 046b, new 046c/046d, 046a | Closed source records, support, asset, nested validation and sensitive-adjunct boundaries. | Retained content; assets; portability |
| new 046rp, new 046pa | Exact Reference Pack set capture/pins and portable historical artifacts. Reveal maps are private adjuncts outside shareable inventories. | Reference binding; portability; recovery |
| 048, 050b | Literal canonical UTF-8 and domain-separated Reporting identities; Composition keeps its distinct plain digest preimage. | Canonical identity |
| 053, 054 | Assemble only admitted owner exports and frozen recipe inputs. Validate the entire model rather than trusting selected fields or a matching digest. | Owner boundaries; frozen inputs |
| 059, 059a, 059b, 060, 060a–060c, 061 | Preserve contributor restrictions and hierarchy. Remove inactive aggregate declassification and bucketing contracts; reject even null/empty retired members. | Redaction; retained content |
| 062, 062a–062c, 063a, 064, 067–069, 069a | Exact rule resolution, stable subjects/tokens, value-free removal evidence, private reveal maps and recovery binding. | Redaction; sensitive artifacts |
| 070, 072a, 073, 075 | Use Core materialized Timeline ordering and preserve display text; no Reporting rebucketing to broaden disclosure. | Timeline; retained content |
| 079b, 079d, 080, 080a, 080b | Closed graph/timeline selection and deterministic manual layout without changing source topology. | Graph and diagrams |
| 083–086, 086a, 086c | Canonical Mermaid grammar and bounds; actual safe SVG output. | Graph and diagrams; native rendering |
| 087, 087b, 087e–087h | Exact deck/composition effects, semantic anchors, authored text and pre-chunk operation order. | Composition effects; Slidev |
| 091, 091a, 091b, 092, 093, 096 | Closed source escaping, local asset references, fixed source options and click/page-count semantics. | Slidev; assets |
| 097, 098, 098a, new 098b, 099, new 099a/099g | Actual pinned toolchain, independently measured confinement and deterministic admitted environment. | Native rendering; confinement |
| 100, 101, new 102a, 103a–103c, new 103d/103e | Closed template vocabulary, declared local assets and bounded local CSS/font themes. No remote assets or unsupported capabilities. | Templates; assets; confinement |
| 107, 108, 108a | Exact bundle member inventory, child bytes, identities and root digest. Wrapper metadata is not output identity. | Bundle integrity; portability |
| 113a, 115 | Deterministic safe diagnostics and closed reason tokens; no raw URLs, renderer payloads or sensitive values. | Failure handling |
| 117b, new 117c, 118, 119, 120–122, 122b, 123 | Supervised bounded rendering, cancellation, secure SVG admission, cleanup and atomic output visibility. | Confinement; failure handling; limits |
| 125, 127, 128, 129 | Reconcile behavior/acceptance/completion tables with the final supported scope; PNG/PPTX and declassification remain excluded. No title-based assurance. | All families; scope review |
| RC-017, RC-019a, RC-040 | Strict canonical composition admission and separate digest preimages. | Composition identity |
| RC-025, RC-031, RC-033a, RC-035a, RC-035c | Ordinary authoring lifecycle and transactional preview admission through the Reporting port. No synthetic release identity. | Composition lifecycle; preview |
| RC-050 | Semantic targets only; local grammar and execution-time resolution have distinct owners. | Composition effects |
| RC-072, RC-073, RC-073c, RC-074, RC-075a, RC-076a | Explicit local validation v2. Retire silently ignored context v1; Reporting Jobs retain complete contextual checks. Sort filtered findings by normative provenance, revalidate immutable bodies, and enforce binding before digest before semantic checks. | Composition validation |
| RC-080, RC-084 | Builder guidance and behavioral scenarios use explicit local scope. Builder UI is optional under unchanged RC-079; no builder conformance is claimed. | Composition validation; scope review |

## Observable obligations and assurance

Exact operational selection, fixture and runner facts remain in
`tools/test_families/module.reporting.json` and
`tools/test_families/module.reportcomposition.json`. Service cases use the
catalog's isolated PostgreSQL/object-store fixtures and ordinary public authority;
unit cases use controlled collaborators. Installed-renderer cases execute the
actual pinned tools and confinement process. Browser progress has its separate
large-grid measurement context. A helper named below is reached by its listed
top-level test; helpers are not additional executable identities.

| Family | Implementation boundary and actual exercising assertions | Evidence posture |
| --- | --- | --- |
| Authority | Core 00 version/dependency rule; the companion table below identifies governing owners and remaining promotion action. | Human review only; tests cannot adopt a document. |
| Owner boundaries | Reporting materializer invokes typed owner contributions; `TestReportingMaterializerUsesOwnerContracts`, registration rejection and production import/table-boundary tests reject direct foreign ownership. Participant malformed/oversized output publishes no result. | Existing full Reporting qualification; integrated applicability pending. |
| Admission | `TestBoundaryReplayDefaultsAndActionIdempotency_Integration`, `TestExactShapesAndRouteScopedVisibility_Integration`, `TestDecoderNormalizationAndRegisteredReasons_Unit` check ordinary auth/roles, exact defaults, unsupported values, replay, recipient/profile and tuple rejection. | Existing full Reporting qualification; integrated applicability pending. |
| Snapshot lifecycle | `TestSnapshotAdmissionAndCancellationDoNotWaitForSourceCapture_Integration` proves acceptance/cancel responsiveness with blocked source work; `TestSnapshotCaptureRetainsSourceOnce` proves frozen replay; publication tests reject altered compact captures without holding the terminal lock across reconstruction. | Existing full Reporting qualification; browser snapshot progress 173.7 ms at the retained recipe checkpoint. |
| Frozen inputs | Snapshot/release payload, execution-boundary and render-model assembly tests reject altered identity, clocks, recipe and unavailable execution; `TestSnapshotReplayAndReleaseProvenanceAreStable_Integration` exercises retained release/replay behavior. | Existing full Reporting qualification; current decoder changes require integrated retest. |
| Canonical identity | `TestReportingAdmittedModelIdentityAndTime`, timestamp-precision and canonical-byte vectors check exact preimages, microsecond admission and distinct snapshot/release/preview hashes; altered admitted models reject. | Existing full Reporting qualification. |
| Retained content | `TestRedactionProfilePrecedenceActionsAndManifest_Unit` reaches retained sections, records, model, fields, structural removals and containers. Source facts/refs, deleted lifecycle and display-state helpers assert actual projections and fail-closed mappings. | Existing full Reporting qualification. |
| Redaction | Profile precedence/conflict and disclosure tests cover exact allow/drop/token decisions, recipient subsets, hierarchy, unsafe actions and default passthrough. Token/subject helpers assert identity, order and no reversible public values. Retired aggregate flags reject at admission. | Existing full Reporting qualification. |
| Sensitive artifacts | `TestRenderBundleSeparatesSensitiveRedactionArtifacts` and ordinary release/preview integration reach private authorization, missing/altered reveal-map denial and atomic binding. External bundles exclude reveal maps. | Existing full Reporting and portability qualification. |
| Reference binding | `TestSnapshotAdmissionAndRerenderRetainSetAcrossActivation_Integration` captures one exact set, changes active data, rerenders from retained pins and rejects invalid binding. | Existing full Reporting qualification; full Reference Data owner also passed. |
| Timeline | Timeline sort/retained-timeline helpers check parsed/partial/ambiguous/unparseable times, stable ties, display strings, overflow and source-state filtering. | Existing full Reporting qualification. |
| Graph and diagrams | Exact result/owner registry, graph selection, retention and diagram-layout tests check every admitted selection rule, duplicate/missing refs, preserved topology, stale/unfinished projections and closed manual geometry. Native tests render manual and automatic diagrams together. | Existing full Reporting/native qualification. |
| Composition effects | `TestCompositionRenderBinding_Unit` and its deck/record-anchor/text helpers check all-null versus exact tuple, semantic ambiguity, operation order, authored roles/partitions, profile default-deny, placeholders and label restrictions. Native authored rendering uses retained redaction. | Existing full Reporting qualification; integrated applicability pending. |
| Slidev | Canonical source/block/text/click tests assert exact escaped bytes, fixed headmatter, chunking, titles, notes and page-state count. `TestReportingInstalledSandboxRenders` checks actual PDF pages and repeated admitted-output identity. | Existing full Reporting/native/package qualification. |
| Output options | Reporting and Composition decoder tests reject PNG/PPTX, malformed/null/unknown members and conflicting source-only options. Source-only mode emits no rendered-output role. | Reporting retained pass and current Composition owner pass. |
| Templates | Template contract/pack/derivation/binding/reachability helpers validate complete closed manifests, feature admission and exact semantic references. | Existing full Reporting qualification. |
| Assets | Asset integration and native image/theme helpers bind retained image/font/CSS bytes, reject changed/undeclared/remote material, and validate portable restored content. | Existing full Reporting/native/portability qualification. |
| Bundle integrity | Bundle/member/portable-catalog tests enforce exact role/path/media inventory, file hashes, no self-listed root manifest, root digest and archive-wrapper independence. | Existing full Reporting and portability qualification. |
| Native rendering | Actual Slidev/Mermaid and pinned Chromium execute through the installed sandbox. Native PDF/SVG structure and page counts are checked; fabricated observation/toolchain/output bindings reject. | Linux amd64/no-LSM package and owner evidence only. |
| Confinement | Native sandbox tests attempt filesystem/network/process escapes; preflight checks declared roots and mount boundaries. Rendered SVG rejects active content and unsafe URLs without echoing raw input. Environment credentials/evidence roots are not mounted. | Existing native/package qualification; broader platform support not inferred. |
| Failure handling | Worker execution, supervision, deadlines, participant failure and cancellation tests require safe deterministic diagnostics, no partial publishable output and final cleanup. Exact replay preserves identities; a new attempt uses new admission. | Existing full Reporting qualification. |
| Limits | `TestReportingDeclaredLimits`, `TestReportingStageEnforcesDeclaredLimits` and chunking/native page checks enforce declared defaults, overrides, hard bounds and mapped failures. | Existing full Reporting qualification. |
| Composition identity | `TestCompositionCanonicalDigests_Unit` and strict render-source cases prove literal vectors, separate preimages, malformed/rehashed-body rejection and metadata binding. | Current focused and full Composition owner passes below. |
| Composition lifecycle | `TestCompositionPublicLifecycleAndAuthority_Integration` checks empty/default state, read roles, CSRF, create/update/freeze/retire, exact replay, conflicts without mutation, immutable/gapless versions and revoked-member replay denial. | Current full Composition owner pass. |
| Composition validation | Same public lifecycle test covers draft/version/inline, binding-before-digest precedence, retired-context rejection, explicit scope and unchanged resource/execution counts. `TestCompositionLocalValidationScopeAndOrdering` checks schema-first precedence, source ordering and safe details. | Current focused and full Composition owner passes. |
| Preview | `TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration` and receipt/recovery helpers prove exact preview identity/time, internal-only execution, replay integrity, no synthetic release and failed-context denial. | Current full Composition owner pass; shared decoder integrated review pending. |
| Portability | Incident Bundles' native-snapshot round trip checks exact Reference Pack requirements, child artifact bytes and tampering before publication. Historical import creates no native Job, approval or destination trust. | Retained selected portability pass; current integrated applicability pending. |
| Recovery | The preview integration reaches `assertReportingPrivateBackupRestore`: capture real release/preview state into encrypted backup, restore into a named isolated database and separate roots, validate retained models/pins, compare private bytes/time/digests, remove adjuncts and require readiness denial. | Current full Composition owner pass reaches actual backup/restore. |
| Clean cutover | `TestReportingIdentityCutoverDoesNotReconstructHistory_Integration` tests each retired migration boundary with retained history, safe preflight diagnostics, unchanged history and migration head; package smoke initializes a fresh isolated deployment. | Existing full Reporting/package qualification; final integrated migration applicability pending. |
| Scope review | Tables 26-A, 27-A and 29-A were reviewed against these families and the implementation checkpoints. Unsupported PNG/PPTX, aggregate declassification and ignored contextual validation were explicitly retired, not silently counted as coverage. | Candidate review complete; promotion still awaits integrated validation and exact source record. |

## Companion decisions

The final portability review found stale companion projections: Core 01 still
named portable catalog v2, Reporting's built-in paragraphs still named template
3 and profile 4/internal 2, and Core 04 AC-266/267 still described eager
request-time capture/profile materialization. Those clauses now match the
already approved v3 portable capture, template 4/profile 5 and cancelable
preparation contracts. AC-056 binds determinism to the same admitted recipe and
sealed preparation, including release identity/time; identical snapshot and
selector strings do not make independent releases the same attempt. This
reconciliation is part of the pending coordinated amendment, not promotion.

Implementation review also found that recipient validation resolved the current
profile selector again during preparation. It must consume the retained authored
policy, with no transport-layer dependency or registry lookup. Add a negative
case in the existing decoder/reason row where the retained policy's allowed
partitions differ from the current built-in policy, then qualify the repaired
owner and packaged recovery path. Until that evidence exists, this entry is
an open implementation obligation.

The implementation now validates recipients against the retained authored policy
through a Reporting-only validator; it no longer resolves selectors or depends
on HTTP errors during preparation. The three specific recipient failures remain
closed Reporting reason tokens, now under the admitted Job's
`release_render_failed` / `redaction_manifest_invalid` diagnostic. Unit checks
cover the retained-policy mismatch and terminal-diagnostic admission. The
existing route idempotency integration also checks a missing recipient through
ordinary admission, failed preparation, absent release row and exact replay.
Focused unit validation passed in
`remediation-retained-recipient-focused-20261016-01`, native exit 0 in 4.011 s.
Full owner and package evidence remains pending. Finalization 01 passed in
21.937 s before the additional public integration assertion; its successor
must qualify the complete source. Retained-run maintenance remains skipped
because `RESULTS_DIR` is unset.

The complete repair passed finalization 02 (`22.162 s`) and
`remediation-retained-recipient-owner-20261016-01` (`make test-slice
OWNER=module.reporting`, `220.430 s`), both native exit 0. The owner run passed
39/39 work units and all 60 cleanup operations. Snapshot progress qualified at
201.3 ms and render progress at 211.8 ms, one observation each, threshold
1,000 ms and zero scheduler overlap. The public failed-preparation/replay
assertions and installed sandbox rendering also passed in this owner selection.
Source: `4d519ca948da4d9daa88f7d6c14e0fa8bb66f07e5a2fc8199a769bbe58399708`.
Graph: `e302a869878b7ac6c043a2199a597f8b43e413184f128040a82b7e0f24704014`.
The explicit manifest
`.cartulary/remediation-evidence/retained-recipient-reporting-20261016.json`
selects only that full-owner root. Its `make test-evidence-audit` passed in
`remediation-retained-recipient-audit-20261016-01`, native exit 0, 2.848 s.
The companion specification reconciliation passed `make lint-markdown` in
`remediation-companion-reconciliation-markdown-20261016-01`, native exit 0,
10.320 s, before this result entry. Final integrated and package qualification
remain separate; this closes the identified retained-policy repair.

The packaged Reference Pack path now also passes:
`remediation-reference-package-repair-20261016-03`, public
`make standup-reference-pack-smoke`, native exit 0, 522.683 s, 15/15 work units
and all 24 cleanup operations complete. It binds the same source above and graph
`3fdadeab1fbb0e72d04295c483b4e2a139f79d55c8d5c074fe8ad9ddcff7b9d6`.
Its retained `standup-reference-pack-smoke/artifacts/scenarios.json` records all
14 scenarios, including actual historical review/external PDFs, pin-protected
removal, destination-trusted portability without source activation, exact PDF
bytes, no reveal-map transport, complete encrypted export, physical collection,
required-content loss and source-loss recovery. The recovery artifact confirms
that original services and source configuration were removed, exact retry is
stable, backup freshness is not fabricated, the prior session is rejected, and
a fresh MFA session succeeds. All destructive steps target only that invocation's
explicit disposable installations. No user deployment or database was reset.
This current package evidence supersedes the incomplete portability observers
in repair roots 01/02; those failed roots remain retained. Broader integrated
release acceptance still requires its own successful invocation.

| Owner | Governing decision and review | Promotion state |
| --- | --- | --- |
| Core 00 | Immutable adopted text remains governing while a successor is proposed; an `adopted_base_version` alone is insufficient. Update the profile registry and this record together after validation. | Pending coordinated status update. |
| Core 01 | Only Slidev/Mermaid; exact snapshot/release/preview tuples, fixed admission time, complete output options, composition binding and bundle-root digest; no invented failed-release facts. | Candidate owner text reconciled. |
| Core 02 | Host/Identity six party-assignment collections and versioned state are source-owner data, captured at the snapshot boundary. | Adopted owner; ordinary mutation/browser evidence retained. |
| Core 03 | Party controls use ordinary inspector authority. REQ-03-287b defines the bounded Snapshots and reports drawer, uncertainty/replay/cancel behavior and exact Job results. | Remove proposed label only at coordinated adoption. |
| Core 04 | Private reveal maps, rendering confinement, default-false authored text permission, exact product progress predicates. | Candidate owner text reconciled; all timed families have retained observations. |
| Reporting / Composition | Promote 1.4.0 and 1.3.0 together; local-authoring validation v2 does not authorize contextual execution. | Pending integrated validation. |
| Reference Data | Adopted Reference Pack 0.2.0 / contract major 2 owns exact set identity, trust, pinning, retention and payload verification. | Current full owner passed; no authority moved into Reporting. |
| Extensions | Adopted 0.12.0 / contract major 2 owns typed participant/condition boundaries, with no active capability grant. | Reporting uses bounded typed owner contributions; integrated dependency checks pending. |
| Graph Projection | Adopted 2.2.0 owns exact immutable result refs and leases. Reporting consumes typed projections and cannot regenerate from live source. | Current interface review complete; integrated owner checks pending. |
| Incident Bundles | Core 01 owns historical portability; exact optional-section inventory and no destination trust/native identity synthesis. | Retained portability passes; integrated owner checks pending. |
| Recovery | Core owner and typed state catalog govern complete relational/object closure, private artifact encryption and readiness validation. | Actual release/preview backup/restore passes; integrated owner checks pending. |

## Current Composition validation evidence

All roots are beneath `.cartulary/test-results/`; native exits were observed
separately. No failed root is reused as performance or aggregate acceptance.

| Run ID | Command and result |
| --- | --- |
| `remediation-composition-local-generate-20261016-01` | `make generate`, exit 2: OpenAPI gate correctly rejected nine undeclared API changes. The immutable 1.0.0 release was untouched; exact changes were added to the authored 2.0.0 change set. |
| `remediation-composition-local-generate-20261016-02` | `make generate`, exit 0, 10.010 s. |
| `remediation-composition-local-focused-20261016-01` | Two exact lifecycle/canonical rows, exit 0, 16.751 s, before the additional precedence/revocation cases. |
| `remediation-composition-local-focused-20261016-02` | Same rows, exit 2; `infra/preflight_error`: test helper used `database/sql.QueryRow` with a context argument. No tests exercised in those failed build units; all 15 cleanup operations completed. |
| `remediation-composition-local-focused-20261016-03` | Same rows after `QueryRowContext` correction, exit 0, 12.242 s. |
| `remediation-composition-local-format-20261016-02` | `make format`, exit 0, 6.230 s. |
| `remediation-composition-local-finalize-20261016-01` | `make agent-finalize`, exit 0, 20.829 s. Retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-composition-local-owner-20261016-01` | `make test-slice OWNER=module.reportcomposition`, exit 0, 33.919 s, 6/6 work units, zero cache hits. |

The full Composition owner root binds executable source
`sha256:4251c8ba3ca5c3c9ae43ac7f530bc18f8eb2053da558fc0c47b200b78e69bbad`
and graph `sha256:01d3626be4486d4043c3f46e578ff015b1400d53583beace40bbcb5e54783ad4`.

## Integrated validation repair register

`remediation-coordinated-integrated-check-20261016-01` ran `make check` with
the physical-fixture scratch root outside `/tmp`. Native Make exited 2 after
676.302 s: 784 of 799 work units passed and all 1,076 cleanup operations
completed. The root retains the failures; it is not aggregate acceptance.

The fifteen failed units resolve to these reviewed causes:

- Reporting's short-lived snapshot publication candidate embedded a generated
  persistence parameter type. It now carries owner-private native metadata;
  conversion remains at the persistence call. Tuple fencing and independent
  admission validation remain required.
- Database inventory tests copied obsolete cardinalities. Their replacement
  requires nonempty, duplicate-free, exact bidirectional catalog membership and
  per-table owner/recovery classification. Three constraints removed implicitly
  by column drops are now dropped explicitly in migrations 93 and 95. Both
  migration catalog projections are declared generated artifacts; hand editing
  them cannot repair the owning migration semantics.
- Server assembly tests constrained private field names/counts. The replacement
  rejects exposed fields and retained generic persistence/construction
  capabilities, preserves the public lifecycle/capability signatures, and
  demonstrates benign private renaming/extraction. This is an assertion-level
  consolidation within the same test and catalog row, with unchanged runner,
  fixture and evidence context. Codex reviewed behavioral and execution
  equivalence under the authorized plan; it does not prove arbitrary runtime
  behavior or recursively inspect all possible wrappers.
- Extension publication, Network Flow's restored Reporting fixture, and the
  client profile catalog retained superseded exact identities. They use the
  current snapshot Job, redaction profile and Reporting profile declarations.
- The telemetry classification registry lacked the new ordinary
  `evidence_blob_verification_failed` reason. It maps to `policy_rejected`,
  consistently with Reference Pack verification failure; normative telemetry
  vocabulary is updated with the authored registry.
- Two unused Reporting frontend exports are private implementation details and
  are no longer exported. The maintained feature entrypoint remains public.
- Workbook and view-contract tests copied field/capability totals or supplied
  incomplete Host/Identity rows after party-reference expansion. Complete
  contract-shaped fixtures preserve explicit domain values, while assertions
  still check capability denial, unique fields, ownership, and reference policy.
  The no-reference inspector fixture now uses Parties. Evidence upload tests
  model the ordinary verification Job and assert verification precedes attach.

Focused native-exit evidence (all beneath `.cartulary/test-results/`):

| Run ID | Result and applicability |
| --- | --- |
| `remediation-integrated-repairs-workbook-20261016-01` | Exit 2, 28.617 s; four selected regression rows passed; five surface cases still had incomplete Host/Identity fixtures. No retry or deadline policy changed. |
| `remediation-integrated-repairs-workbook-20261016-02` | Exit 0, 14.260 s; the same surface row passed after completing those fixtures. |
| `remediation-integrated-repairs-server-20261016-01` | Exit 0, 3.465 s; exact runtime-boundary and publication-catalog rows. |
| `remediation-integrated-repairs-generated-boundary-20261016-01` | Exit 0, 0.790 s; generated SQL types remain behind persistence adapters. |
| `remediation-integrated-repairs-view-contracts-20261016-01` | Exit 0, 2.340 s; reference ownership, uniqueness and policy assertions. |
| `remediation-integrated-repairs-database-20261016-01` | Exit 2, 23.071 s; unit recurrence passed; live catalog parity identified the three implicitly removed constraints. Explicit migration repairs follow this root. |
| `remediation-integrated-repairs-ddl-generate-20261016-01` | Exit 0, 10.010 s; projections regenerated from explicit migration drops. |
| `remediation-integrated-repairs-database-20261016-02` | Exit 0, 23.785 s, 4/4 work units; live database/catalog parity and recurrence checks pass. |
| `remediation-integrated-repairs-extensions-20261016-01` | Exit 0, 13.893 s, 3/3 work units; exact client availability/profile catalog. |
| `remediation-integrated-repairs-fallow-20261016-01` | Exit 0, 20.283 s; unused exported Reporting implementation details removed. The existing advisory complexity report is not recast as a blocking acceptance gate. |
| `remediation-integrated-repairs-otel-20261016-01` | Exit 0, 7.209 s, 6/6 work units; ordinary verification failure has the declared telemetry classification. |
| `remediation-integrated-repairs-networkflow-20261016-01` | Exit 2, 67.281 s; the fixture still selected retired profile version 2. It now resolves `DefaultRedactionProfileVersion` from the Reporting owner. No restored history is migrated or inferred. |
| `remediation-integrated-repairs-networkflow-20261016-02` | Exit 2, 63.942 s; after selector repair, the partial hand-built job payload failed the retained graph-input constraint. The final fixture now admits its snapshot and release through ordinary authenticated public routes and holds only the injected render protocol. It does not manufacture retained profiles, hashes, timestamps, recipes or execution-attempt identity. |
| `remediation-integrated-repairs-snapshot-20261016-01` | Exit 0, 76.167 s; exact snapshot/provenance row, including publication candidate fencing and clean-cutover cases. |
| `remediation-integrated-repairs-format-20261016-02` | Exit 0, 6.183 s; current source formatting after fixture and generated-policy repairs. |
| `remediation-integrated-repairs-networkflow-20261016-03` | Exit 0, 67.882 s; ordinary snapshot/release admission reaches the controlled render boundary, and the existing atomic restore rollback, exact result identity and job/lease reconciliation assertions pass. This fixture does not qualify native rendering or confinement. |
| `remediation-integrated-repairs-finalize-20261016-01` | Exit 0, 20.539 s; finalization before integrated verification. Retained-run maintenance skipped because `RESULTS_DIR` was unset; no compatible successful full warm check root was available. |

`remediation-coordinated-integrated-check-20261016-02` reran the complete
`make check` after these repairs: native exit 0, 676.800 s, 799/799 work units,
zero cache hits, and all 1,080 cleanup operations completed. It binds source
`sha256:87f7aaba072327d300f18c9755abf6b6fe9fd0522d413e250b56da22da17b3cc`
and graph
`sha256:01038448d002f6e15ad1cd371db15e61e5686391bb7e86ae14ff9b5b53918fab`.
This closes the integrated repair register at that source checkpoint. Subsequent
harness execution-context changes require their own validation.

Release qualification, exact adoption source identity, final owner audits and
final handoff remain pending. The separate cost portfolio and browser
consolidation also remain required for overall nine-finding closure.
