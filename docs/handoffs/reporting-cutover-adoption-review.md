# Reporting cutover clause and adoption review

This is human review material, never runtime, generator, test, release, or conformance input. Baseline: adopted Reporting 1.3.0 and Composition 1.2.0 at `78a4effc23aac4b46ca414e11b91e17892f790c7`. Candidates: Reporting 1.4.0 and Composition 1.3.0. Review: Codex under the approved remediation instruction; incomplete rows remain open for accountable owner review.

Current clause, companion and assurance review: [coordinated adoption register](reporting-coordinated-adoption.md). This document retains historical implementation checkpoints and failed roots. Earlier partial-status tables are historical; they do not describe current closure unless explicitly revalidated.

The following rows are the initial changed-clause review against that immutable baseline, including changed tables attached to a requirement. Later implementation checkpoints add and refine obligations; this initial table is not a complete current adoption checklist. Artifact-version-only edits preserve the underlying behavioral obligation; new identity, Reference Pack, portability, and canonical-byte cases are candidate-amendment coverage. No adopted-base test is retired merely because the candidate changes an expectation. Newly enrolled stale fixtures were corrected against current typed admission boundaries without deleting their assertions.

Reporting tests below reside in `internal/modules/reporting/`, unless the name identifies Composition or Incident Bundles. Exact operational rows and fixture/isolation routing remain in existing owner catalogs; source reconciliation records enrollment. `remediation-reporting-unit-20261013-01` passed all 20 Reporting rows, including six service rows. `remediation-composition-cutover-20261013-02` passed all three Composition rows. `remediation-reporting-portability-20261013-01` passed its selected Incident Bundles integration row. A listed test establishes only its actual assertions, not every clause named beside it.

| Clause | Observable changed obligation | Adoption dependency | Exercising assertion/test | Disposition |
| --- | --- | --- | --- | --- |
| REQ-RPT-006 | Promotion requires coordinated owners and production-facing evidence | Core 00–04 and all companions | Human promotion checklist below | Open; no test can adopt a document |
| REQ-RPT-014 | Updated deck artifact identity; questions answered from frozen content | Core 01, Graph Projection | redaction_test.go; graph_result_render_test.go | Partial; rendering assertions pass, full scenario review pending |
| REQ-RPT-018 | Companion blockers retain fail-closed conformance consequence | Core 00 | Human dependency review | Open; document status is never executable metadata |
| REQ-RPT-019 | Companion closure table imports actual owner decisions | Core 00–04, Composition, Reference Data, Extensions, Graph Projection, Incident Bundles, Recovery | Human immutable-owner comparison | Open; a table label is not evidence |
| REQ-RPT-019a | Participant emits current export schema under exact bounded context | Extensions | TestRenderExportParticipantValidatesClosedOutputAndBounds; TestReportingJobWorkerParticipantFailurePublishesNoRenderOutput | Demonstrated within Reporting owner; companion admission review pending |
| REQ-RPT-021 | Closed current identifiers; retire replaced v1 artifacts without aliases | Core 01, Composition | TestReportingAdmittedModelIdentityAndTime; TestReportingMaterializerUsesOwnerContracts; route decoder rejection tests | Partial; behavioral successors replace the retired private-name guard; complete nested schema admission remains open |
| REQ-RPT-025 | No live source reads after snapshot materialization | Core 01, source owners | TestSnapshotAdmissionAndRerenderRetainSetAcrossActivation_Integration | Demonstrated for pack activation and rerender; full source-family review pending |
| REQ-RPT-027 | Frozen render tuple including admission timestamp | Core 01 | TestSnapshotReplayAndReleaseProvenanceAreStable_Integration; TestReportingAdmittedModelIdentityAndTime | Demonstrated for exercised tuples |
| REQ-RPT-027a | Current derivation profile resolution and unsupported rejection | Core 01, Composition | TestDecoderNormalizationAndRegisteredReasons_Unit | Partial; complete unsupported-profile matrix pending review |
| REQ-RPT-027c | Derivation profile v2 closed shape and tokens | Core 01, Composition | TestDecoderNormalizationAndRegisteredReasons_Unit | Partial; every profile member needs assertion review |
| REQ-RPT-027e | Deterministic immutable content-selection algorithms | Composition, Graph Projection | TestGraphResultSelectionRulesPreserveExactProjectionOrder_Unit | Partial; exact graph ordering covered, full derivation table pending |
| REQ-RPT-038 | Only current export model supplies canonical case content | Core 01 | TestBuildExportModelUsesReportingOwnedMetadataSnapshotStableHash; TestRedactionProfilePrecedenceActionsAndManifest_Unit | Structured retained source projection implemented below; lifecycle eligibility is implemented in the latest checkpoint; full container-rule census and remaining diagram/asset derivation still require closure |
| REQ-RPT-039 | Current export shape, required nullable identity pair and pack binding | Reference Data, Composition | TestReportingAdmittedModelIdentityAndTime | Demonstrated missing/null/type/unknown-member rejection |
| REQ-RPT-039a | Separate snapshot, release and preview identities; frozen microsecond time; reject tampering | Core 01, Composition | TestReportingAdmittedModelIdentityAndTime; TestReportingTimestampPrecisionIsFixedAtAdmission; TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration | Demonstrated for tested identity/time/replay cases; retained-job tamper matrix review pending |
| REQ-RPT-042a | Current redaction manifest; no raw value digests or dropped fields retained | Core 04 | TestRedactionProfilePrecedenceActionsAndManifest_Unit; TestExternalValidationRejectsOpaqueBytesAndWorkingMaterial_Unit | Demonstrated for exercised actions; identity propagation review pending |
| REQ-RPT-042b | Explicit source-state mapping under derivation v2 | Source owners | TestDisclosurePartitionsAndCuratedSupportRefsFailClosed_Unit | Partial; state mapping table review pending |
| REQ-RPT-046a | Closed nested validation/content objects and exact nullable members | Core 01 | TestBuildExportModelUsesReportingOwnedMetadataSnapshotStableHash | Partial; persistence round-trip passes, all unknown-member cases pending |
| REQ-RPT-046rp | Atomic exact Reference Pack capture/pin, retention and replay; no current-set substitution | Reference Data, Incident Bundles | TestSnapshotAdmissionAndRerenderRetainSetAcrossActivation_Integration; TestNativeSnapshotReferencesSurviveBundleRoundTrip_Integration | Demonstrated activation/rerender/round-trip contexts; recovery remains pending |
| REQ-RPT-046pa | Portable canonical catalog, exact artifacts and explicit sensitive omissions | Incident Bundles, Recovery | TestPortableArtifactCatalogClosure; TestNativeSnapshotReferencesSurviveBundleRoundTrip_Integration | Partial; catalog and bundle transport pass, full recovery pending |
| REQ-RPT-048 | Literal UTF-8 canonical serialization; Composition retains separate hash preimages | Composition | canonical_json_test.go; TestCompositionCanonicalDigests_Unit | Demonstrated independent literal canonical/digest vectors; shared canonical owner verification pending |
| REQ-RPT-050b | Domain-separated snapshot, release, preview, deck and token hash IDs | Composition | TestReportingAdmittedModelIdentityAndTime | Demonstrated independent fixed vectors and release/preview collision denial |
| REQ-RPT-053 | Current materialization stages preserve immutable admitted input | Core 01, Composition, source owners | TestSnapshotReplayAndReleaseProvenanceAreStable_Integration; TestReportingJobWorkerRenderFailureSummaryIsDeterministic | Partial; complete stage review pending |
| REQ-RPT-062a | Redaction manifest v4 omits source and replacement values | Core 04 | TestRedactionProfilePrecedenceActionsAndManifest_Unit | Current fixture actions and structural causes are value-free; complete container rule census remains open |
| REQ-RPT-067 | Token manifest v3: one subject identity/display digest, exact recipient set, complete rule-manifest binding, no reversible display material | Core 04 | TestRenderBundleSeparatesSensitiveRedactionArtifacts; TestRedactionProfilePrecedenceActionsAndManifest_Unit; current release/preview integration | Subject/use shape and binding demonstrated in the canonical-token checkpoint; full structured-model census remains open |
| REQ-RPT-068 | Reveal map v3: exactly one frozen display value per public token, private storage and atomic binding | Core 04, Incident Bundles, Recovery | TestRenderBundleSeparatesSensitiveRedactionArtifacts; release/preview recovery helpers; current native-snapshot portability row | Demonstrated bounded contexts and rehashed tampering rejection; current packaged deployment/recovery qualification remains pending |
| REQ-RPT-069a | Redaction manifest v4 binds actual rule decisions and separate causal structural removals | Core 01, Core 04 | TestRenderBundleManifestBindsOutputHash; TestRedactionProfilePrecedenceActionsAndManifest_Unit | Bounded structural accounting implemented; full canonical retained-model derivation remains open |
| REQ-RPT-087 | Deck v2 identity follows release/preview admission | Composition | TestReportingAdmittedModelIdentityAndTime; TestRenderBundleManifestBindsOutputHash | Partial; ID vector covered, complete deck/preview assertion review pending |
| REQ-RPT-107 | Bundle v2 binds exact model, pack set, timestamp, digests and identity | Core 01, Reference Data, Incident Bundles | TestRenderBundleManifestBindsOutputHash; TestRenderBundleExternalReleaseRecordsDeterminismDigest | Demonstrated deterministic bundle binding for exercised inputs; restored artifacts pending |
| REQ-RPT-125 | Updated scenario table remains behavioral, no one-test-per-row requirement | All reporting companions | Reporting owner tests plus scenario-level review | Open; aggregate pass does not prove all scenarios |
| REQ-RPT-127 | Current acceptance criteria require exact artifact and admission semantics | All reporting companions | Obligation review, not routing identity | Open; unresolved assurance remains visible |
| REQ-RPT-129 | Completion checklist requires aligned owners, schemas, deterministic time and secure rendering | Core 00–04, Composition | Promotion checklist below | Open; no adoption until every condition closes |
| REQ-RC-019a | Strict duplicate/Unicode/integer admission; fixed canonical byte profile | Reporting §§8/10, Core 01 route admission | `TestCompositionCanonicalDigests_Unit`; Composition routes | Canonical vectors demonstrated; route side-effect denial review pending |
| REQ-RC-040 | Plain Composition digest preimages remain distinct from Reporting schema domains | Reporting §10 | `TestCompositionCanonicalDigests_Unit` | Demonstrated independent literal vectors; no domain-prefix compatibility reader |
| Composition status/preview introduction | Exact admitted preview-attempt identity and time, no fabricated release | Reporting 1.4.0, Core 01 | `TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration` | Preview/release separation and replay demonstrated; coordinated adoption pending |

## Current Composition adoption review — 2026-10-16

The render-recipe checkpoint closes the production renderer implementation and
its retained owner evidence; the historical initial table above is not the
current implementation inventory. Direct review then found that Composition's
validation route ignored `validation_context`, and immutable-version validation
checked a digest without validating the body. These are adoption blockers.

The selected structural remediation explicitly retires contextual validation
from this synchronous authoring route. Composition owns local schema, defaults,
resource/digest binding, grammar and local references. Reporting preview and
release Jobs remain the single owner of full snapshot/template/redaction/Graph
and rendering validation. A second synchronous analysis pipeline would duplicate
expensive materialization, depend on live graph-retention machinery, and risk
claiming render admissibility without the exact admitted execution. No existing
working contextual behavior is preserved by accepting its ignored parameter.

Composition REQ-RC-072–074/080 now define
`validate_report_composition_local_v2`, request v2 and summary v2. Every summary
contains `validation_scope=local_authoring`. Retired `validation_context`, even
null, and render-context members reject with `composition_schema_invalid`.
The removed context-v1 schema has no compatibility reader. Clients consume local
scope and use ordinary preview/release for contextual assessment. The immutable
OpenAPI release remains unchanged as comparison evidence; authored projection
components advance to V2 and are regenerated. This interface cutover is part of
the user-authorized clean pre-production amendment, not a claim of compatibility.

Immutable validation now uses the strict retained-body decoder and compares
stored composition ID, version and digest. Rehashed malformed content rejects.
Typed internal findings retain source-array/index provenance and sort by stage,
source order, local identity, code and canonical safe details. Invalid identity
values are excluded from diagnostic details. These are implementation changes;
fresh V2 route and whole-owner evidence are still required.

The first new ordinary HTTP lifecycle row passed in
`remediation-composition-lifecycle-focused-20261016-01` (native 0, 16.673 s),
together with canonical/retained-source checks. It covers create/update/freeze,
read and retirement, exact replay after lifecycle changes, conflicts with no
state change, viewer/editor/reviewer/admin roles, deployment-admin-only denial
and CSRF. That root precedes the V2 validation change and is not V2 evidence.
Formatting 01/02 passed in 6.198/6.360 s; generation passed in 9.810 s. Local-V2
formatting 01 passed in 7.420 s. No test was deleted for this change.

Builder UI conformance is not claimed: REQ-RC-079 explicitly permits route-family
implementation without a builder. This disposition does not remove authoring,
preview or rendering obligations. Reviewer: Codex under the user-authorized
implementation and coordinated-adoption plan; no independent human approval is
invented. Coordinated promotion remains withheld pending remaining review and
validation.

## Promotion checklist

| Condition | Owning text / validation boundary | Current disposition |
| --- | --- | --- |
| Core 00 registry and precedence | Core 00 immutable-baseline rule and coordinated registry entry | Governance clarified; registry promotion withheld |
| Core 01 output vocabulary | Only Slidev/Mermaid current; invalid-output route tests | Reporting tests pass; companion clause review pending |
| Core 01 release tuple | Admission time, output options, root bundle digest | Reporting route/model tests pass; companion clause review pending |
| Core 02 party source state | Host/Identity assignment and partition provenance | Existing adopted owner; production-facing evidence review pending |
| Core 03 party editing | Ordinary editing authorization; no recipient-visibility hot path | Existing adopted owner; UI evidence review pending |
| Core 04 reveal-map boundary | Internal sensitive authorization, retention and export omission | Partial bundle/portable evidence; full authorization/recovery pending |
| Core 04 sandbox | Local rendering and capability isolation | Installed native PDF/SVG and adversarial checks passed; exact-image package qualification passed under the retained native-amd64/no-LSM environment below. Other LSM policies are not qualified. |
| Core 01 composition tuple | Exact version/digest and freeze binding | Composition facade/preview tests pass; companion review pending |
| Core 04 authored text gate | Default false and denial before output | Default-deny job path demonstrated below; opt-in partition/redaction remains open |
| Composition companion | Coordinated 1.3.0 adoption with immutable baseline record | Candidate remains proposed |
| Reporting derivation v2 | Closed profile/token semantics and unsupported rejection | Full clause-by-clause evidence review pending |
| Every current behavior | Reporting Tables 26-A, 27-A and 29-A | No aggregate-pass shortcut; unresolved rows above block closure |
| Reference Data / Extensions / Graph Projection | Exact pack binding and current participant/artifact contracts | Reporting integration evidence available; companion-owner slices pending |
| Incident Bundles / Recovery | Exact portable catalog, sensitive omissions, restored identity and retained pins | Selected portability passes; Reporting-specific recovery pending |
| Clean preproduction cutover | Fresh initialization and incompatible-history rejection before any mutation | Existing migration fixture must be bound to exact retained run; new operator rehearsal pending |

## Cutover and rollback

No compatibility reader, dual writer, historical identity reconstruction, or in-place downgrade is admitted. Rehearsal uses a named disposable database and separately provisioned storage; never a guessed development or production destination. Preserve a matching application, data, configuration, keys, object roots and verified backup as one recovery set. Rollback restores that matching set into isolation or initializes a fresh compatible deployment. See the [cutover procedure](../guides/reference-pack-development-cutover.md); older undocumented rehearsal claims are not reused as current proof.

Coordinated promotion remains withheld while any required dependency or assurance is open. Once closed, the adoption record must name the exact candidate source revision, governing scope, companion versions, per-condition evidence, and review disposition, and update both subsystem status headers and Core 00 together. Executable metadata never reads these records.


## Confirmed renderer adoption blocker

Direct implementation inspection found that `renderDeterministicPDF` in `internal/modules/reporting/render_bundle.go` emits a PDF header, two hashes and EOF, without a PDF object graph or rendered slide pages. `defaultReportingToolchainSnapshot` records synthetic in-process Slidev/Mermaid version names and `/usr/bin/chromium` without launching those tools. The sandbox observation is constructed with `network_access_observed=false` and `in_process_deterministic_renderer`; it is not measured process-isolation evidence.

Consequently the passing bundle tests demonstrate deterministic serialization and hashing of the existing bytes, not conformant PDF rendering, pinned rendering tools, sandbox isolation, or publication readiness. This is an implementation and assurance defect under the existing adopted rendering requirements, not a reason to weaken them or promote the candidate. Replace the placeholder path with a typed renderer boundary and actual pinned local rendering under the adopted sandbox policy; validate actual pages, hostile input denial, filesystem/network isolation, deterministic output and truthful toolchain/sandbox evidence. Until then, Reporting Tables 1-A, 26-A, 27-A and 29-A remain open. The cutover identity work can proceed independently, but coordinated adoption cannot.

## Rendered SVG boundary repair

The bundle admission path now parses rendered SVG as XML and validates the closed
Reporting Table 24-E subset. It rejects directives/entities, processing
instructions, comments, wrong namespaces, duplicate attributes, active elements,
URL-bearing/style attributes, invalid numeric geometry and malformed path or
transform data. The candidate text makes its renderer output grammar explicit.
No rejected SVG bytes are included in diagnostics. This is a validation boundary,
not a sanitizer that silently rewrites approvable output.

The former `TestRenderBundleRejectsUnsafeSVGLabelInput` asserted that an inert
label containing `url(...)` must fail. That prohibition has no behavioral owner:
the CSS restriction applies to active attributes, while redacted text remains
text. Its maintained successors are `TestRenderBundlePreservesInertSVGLabelText`,
`TestRenderedSVGClosedOutputBoundary`, and
`TestRenderedSVGRejectsMalformedDocuments`, in the existing Reporting source
contract row and unchanged unit runner/fixture context. Behavioral review:
Reporting REQ-RPT-122b/Table 24-E. Execution/evidence review: existing Reporting
Go owner catalog, no lost browser or sandbox assertion. Both reviews performed
under the approved remediation instruction. No real-rendering or sandbox
assurance is inferred from these unit checks; the renderer adoption blocker above
remains open.

Parser review used the [SVG path grammar](https://www.w3.org/TR/SVG11/paths.html#PathDataBNF)
and [Go XML decoder contract](https://pkg.go.dev/encoding/xml#Decoder.Token).
These references inform implementation review only; executable verification
does not read this document or the external sources.

The first SVG row run, `remediation-reporting-svg-unit-20261014-01`, failed
(3.546 s) because the old bundle fixture omitted `/incident/status` and
`/incident/title`; its previous rejection therefore never established SVG
security. The next attempt (2.177 s) exposed the fixture's assumption that the
current placeholder diagram used values as labels; it uses field paths. The
fixture now includes required fields and an explicit inert path label.
`remediation-reporting-svg-unit-20261014-03` passed (2.329 s, outer exit 0), with
all 10 cleanup operations completed. Source:
`sha256:61e8c7208fcecf36b8c8f2b8defb1eabfefd77fdc1cad31058d8fa2227769c5b`.
The XML adversarial checks passed in all three runs. Additional split-text
control-character, reserved namespace and ASCII numeric-whitespace cases now
pass in the complete Reporting owner slice recorded below.

Generation passed in `remediation-reporting-svg-generate-20261014-01` (10.430 s).
Finalization passed most recently in
`remediation-reporting-svg-finalize-20261014-03` (24.191 s); retained-run
maintenance was skipped because `RESULTS_DIR` was unset. Markdown lint passed in
`remediation-audit-svg-markdown-20261014-01` (21.400 s), before this retained-run
annotation.

The integrated check found an authored-Go formatting failure in
`rendered_svg.go` after the final split-text test change. `agent-finalize` does
not replace `make format`; formatting must be rerun before lint qualification.
The retained root is `remediation-audit-svg-integrated-check-20261014-01`.

That integrated run finished with 767 passing units and one formatting failure
in 559.799 s (outer exit 2); all 1,043 cleanup operations completed. It is a
failed aggregate, despite the passing test units. `make format` subsequently
corrected the authored Go formatting.

The Mermaid source serializer now assigns renderer IDs in retained selection
order, resolves every edge endpoint, emits the specified flowchart edge-label
syntax, and serializes sequence participants and messages separately. Labels
use NFC normalization, whitespace collapse, bounds and escaped literals. Exact
byte vectors and negative bounds/endpoint cases replace the former permissive
edge assertion. Template-specific direction/truncation and real rendering remain
open; these source checks establish neither browser execution nor sandboxing.

`remediation-reporting-source-unit-20261014-01` passed the two affected rows in
2.841 s. Formatting and generation passed in the corresponding source roots;
`remediation-reporting-source-finalize-20261014-01` passed in 24.411 s, with
retained-run maintenance skipped because `RESULTS_DIR` was unset. The complete
Reporting owner slice, `remediation-reporting-source-owner-20261014-01`, then
passed in 28.407 s (outer exit 0).

Manual Slidev diagram support, structured deck derivation and the real sandboxed
renderer remain open implementation work.

## Companion input admission repair

Composition now exposes one `DecodeRenderSource` facade used before Reporting
projects render fields. It checks the complete top-level member set, canonical
byte equality, required/null distinctions, identity scalars, body validation and
the appropriate plain SHA-256 preimage. Version-preview descriptors additionally
reconstruct and verify the immutable composition digest. Retained version and
preview loaders compare their independently stored binding columns with the
validated document; a matching digest column alone cannot admit altered bytes.
Reporting rejects preview-source documents outside an internal-draft preview.

The Reporting-owned `diagramcontract` leaf provides the shared closed selection
rule decoder to both owners. Composition materializes declared defaults and
authored-body NFC before admission/canonicalization; immutable inputs needing
repair are rejected. Source refs and anchor identifiers remain opaque; renderer
name substrings and presumed prefixes are no longer treated as proof that a
valid semantic reference is generated source. Explicit raw-source members remain
invalid. Actual anchor/diagram membership and complete operation semantics still
need the renderer work above.

`remediation-composition-admission-owner-20261014-02` passed the complete
Composition slice (20.000 s), including immutable/preview digest separation,
unknown fields, nested selection errors, default equivalence, normalization,
opaque references, and persisted-byte tampering. The affected Reporting row
passed (6.317 s); finalization passed (24.450 s), with retained-run maintenance
skipped because `RESULTS_DIR` was unset. Broader Reporting owner verification
passed in `remediation-composition-admission-full-reporting-20261014-01`
(29.228 s), backend boundary verification passed in the corresponding `boundary`
root (2.280 s), and standalone Go lint passed in the corresponding `lint` root,
all outer exit 0. These results do not close the real-renderer adoption blocker.

## Snapshot aggregation and frozen output options

Snapshot materialization now copies input before sorting, rejects empty or
duplicate source paths, and constructs each Timeline event from its complete
aggregated record. The event carries all fields, support references and
disclosure partitions, with contiguous field ordinals and empty collections
preserved as arrays. The current Timeline provider normally emits one object
per record; the regression exercises the supported multi-field aggregation
boundary rather than claiming a reproduced production duplicate-event incident.
Typed provider-fact consumption and the complete Timeline event schema remain
open work.

`TestSnapshotRecordMaterialization`, selected by
`module.reporting.unit.snapshot_record_materialization`, checks aggregation,
input immutability, traversal-order independence and invalid source identities.
The first two runs exposed an empty-array/null mismatch in copied event support
references. Both retained the assertion in their terminal Go diagnostics; there
was no missing-diagnostic defect. After preserving empty arrays,
`remediation-reporting-records-focused-20261014-02` passed in 2.211 s, outer exit
0. The failed roots `remediation-reporting-records-unit-20261014-01` and
`remediation-reporting-records-focused-20261014-01` remain retained with
`product/test_assertion_failure`, outer exit 2 and completed cleanup.

Reporting now admits defaults only at request admission. Frozen render consumers
require the complete output-options shape and reject missing, null, unknown,
duplicated, mistyped or conflicting members, retired schemas and invalid UTF-8.
Public admission also rejects explicit null instead of interpreting it as a
false boolean. Retained release loading validates those options before use.
This projects REQ-RPT-027/Table 7-A, REQ-RPT-031–033 and Table 7-D; it does not
add a compatibility reader or change the options identity. JSONB whitespace and
member order are immaterial, but missing semantic content is never repaired.
`TestReportingFrozenOutputOptions` owns the new exact catalog row
`module.reporting.unit.frozen_output_options` under the existing source-contract
verification. Render fixtures now supply admitted options explicitly.

The interrupted generation root
`remediation-reporting-options-generate-20261014-01` retained a pass summary
(10.190 s, tool exit 0); its original outer Make exit was unavailable after
interruption. Formatting passed in
`remediation-reporting-options-format-20261014-01` (4.700 s, outer exit 0).
The four affected rows passed in
`remediation-reporting-options-focused-20261014-01` (4.262 s, outer exit 0).
`remediation-reporting-options-finalize-20261014-01` passed in 18.165 s, outer
exit 0; retained-run maintenance was skipped because `RESULTS_DIR` was unset.
These are implementation checks, not coordinated adoption or real-rendering
qualification.

The complete Reporting slice subsequently passed in
`remediation-reporting-options-owner-20261014-01` (24.300 s, outer exit 0).

## Retained release tuple and preview admission

Both retained release loaders now use one closed private-job decoder. It admits
strict JSON before decoding, compares the complete reencoded shape, and verifies
the independently stored actor, incident and admission timestamp. Frozen source
metadata must agree with the model. Output kind, options, graph references,
recipients, template and composition identities must agree with the exact
normalized admission request. The template must resolve to its declared local
version; a bound composition must pass the companion decoder and match the
retained tuple. Unknown, missing, null or changed inputs fail before rendering;
loading no longer invents request defaults. This establishes consistency across
retained bindings, not protection against an attacker replacing every stored
binding and digest together.

`TestReportingFrozenReleasePayload`, in
`module.reporting.unit.frozen_release_payload`, exercises omitted members,
unknown members at multiple depths, duplicate keys, changed source metadata,
actor/incident/job/time mismatches, template changes, option changes, recipient
changes, graph nulls, and unbound composition inputs. Normal JSONB formatting
remains accepted. The test and snapshot/release integration row passed in
`remediation-reporting-payload-focused-20261014-01` (17.183 s, outer exit 0).

Preview requests previously retained an empty options object and materialized
defaults in the worker. Reporting's `outputcontract` leaf now owns admission
and immutable consumption for both routes. Composition materializes exact
Reporting defaults before normalization, idempotency comparison and persistence;
the worker validates complete options without repair. This implements
Composition §9 preview defaults and Reporting §7.5. The candidate's clean
pre-production cutover remains required; no historical preview-default reader
was added. The new Composition admission test proves omitted, empty and explicit
defaults have identical normalized requests, source-only outputs are frozen
disabled, and malformed/conflicting values reject. The integration test checks
stored options and consumes the same frozen composition version in a release
with its own identity.

Extraction initially left an unused Go import. The exact roots
`remediation-reporting-preview-options-unit-20261014-01` (0.657 s,
`infra/preflight_error`) and
`remediation-reporting-preview-options-composition-20261014-01` (1.656 s,
`harness/child_target_failure`) retained that build failure; both outer exits
were 2. After removing the import, focused Reporting checks passed in
`remediation-reporting-preview-options-unit-20261014-02` (2.700 s, outer exit 0).
The complete Composition slice passed in
`remediation-reporting-preview-options-composition-20261014-02` (13.854 s,
outer exit 0), including the persisted preview and bound-release integration.

The broader Reporting run `remediation-reporting-payload-owner-20261014-01`
failed only its import guard (20.371 s, 7 of 8 units passed, all 21 cleanup
operations complete, outer exit 2). The guard had tied approved package
interfaces to private filenames. Its successor keeps the same test identity,
row, unit runner and fixture, but constrains permitted owner package interfaces.
AST cases reject internal or unadmitted source imports, including raw quoted
paths, while renamed aliases and extracted consumers pass. Per-file placement
assertions are retired as incidental realization. Behavioral review: Reporting
Table 5-A/§6.1 consumer boundaries and §7 source ownership; execution review:
the unchanged `module.reporting.unit.owner_boundaries` context and retained
direct-table guards. Reviews: Codex under the approved remediation instruction,
without claiming independent adoption review. The survivor passed in
`remediation-reporting-scope-boundary-reporting-20261014-01` (2.675 s, outer
exit 0).

After the import-boundary repair, the complete Reporting slice passed in
`remediation-reporting-payload-owner-20261014-02` (21.023 s, outer exit 0),
with all 20 cleanup operations complete. The backend boundary check passed in
`remediation-reporting-payload-boundary-20261014-02` (2.050 s, outer exit 0),
with all 12 cleanup operations complete. Both use source
`sha256:8ddfb5f3dc8944ed4c45161e5e2442be5c3372da2b50baf66590ad07cc328a8e`.
Finalization passed in
`remediation-reporting-scope-boundary-finalize-20261014-01` (18.003 s,
outer exit 0); `RESULTS_DIR` was unset, so retained-run maintenance was skipped.

Generation drift passed in the `remediation-reporting-payload-generated-20261014-01`
target summary (12.530 s). That multi-target Make command nevertheless exited
2: the next target correctly rejected reuse of the nonempty explicit run root
as `config/configuration_error`. No output root was cleared or reused. The
remaining targets ran separately with fresh identities:
`remediation-reporting-payload-policy-20261014-01` passed generated-artifact
policy (0.843 s), and `remediation-reporting-payload-shapes-20261014-01` passed
JSON shape checks (4.367 s), both outer exit 0.

The next retained-input review adds two independent comparisons. Release source
content must match the separately retained snapshot, including its row binding;
changing the job model and recomputing its own digest is insufficient. The
normalized request must also match the request digest and route identity in the
Jobs-owned retained admission, accessed through its existing typed reader.
Checks run during loading and transactional completion. Expired or incompatible
worker admission is not reconstructed. Regression cases distinguish a
self-consistent altered job from the original snapshot and Jobs admission.
These checks concern retained-input integrity; the actual renderer and full
promotion criteria remain open.

The independent-binding focused run passed in
`remediation-reporting-independent-binding-focused-20261014-01` (17.098 s,
outer exit 0). Finalization passed in
`remediation-reporting-independent-binding-finalize-20261014-01` (18.066 s,
outer exit 0); retained-run maintenance was skipped because `RESULTS_DIR` was
unset. At source
`sha256:05094636d4eb08370688fc5ec1c6119bb08c171bb262347658c8cd2fa42a664c`,
complete Reporting and Composition owner slices passed, respectively:

- `remediation-reporting-independent-binding-owner-20261014-01`: 20.657 s,
  outer exit 0, 8 passing units, one permitted cache hit, all 20 cleanup
  operations completed. Graph
  `sha256:09e4d5adaf96068f1a0f2970e9f28cfc91b23ba906ad26221dc3cac9f93aea38`.
- `remediation-reporting-independent-binding-composition-20261014-01`:
  11.621 s, outer exit 0, 5 passing units, no cache hits, all 18 cleanup
  operations completed. Graph
  `sha256:b6300c55248984912830d09882a417bed8cd6fed266931cf900d21ba061e1466`.

Markdown lint passed before this final run annotation in
`remediation-reporting-payload-markdown-20261014-01` (21.380 s, outer exit 0).
None of these runs is a clean-source performance portfolio or publication claim.

The subsequent lint failure and explicit measurement-guard repairs are recorded
in [the timing review](timed-product-assurance-review.md#measurement-guard-validation-after-restart).
The unused Composition default was removed; the repaired full lint invocation
`remediation-measurement-guards-lint-20261014-01` passed all 11 units (25.093 s,
outer exit 0). This does not close the remaining renderer, model, recovery or
coordinated-adoption obligations above.

## Selected-rule and subject-token corrections after restart

The next Reporting slice addresses existing Core 01 REQ-386–390 and Reporting
REQ-RPT-059a, 062c and 063/063a/065 obligations. It does not promote Reporting or
claim complete redaction-model conformance.

- An external field retaining a disallowed partition now fails with
  `redaction_manifest_invalid / disclosure_partition_unresolved`. The former
  implicit conversion to `drop` and rewritten selected-rule trace are removed.
  Explicit selected drops retain their original rule identity and trace.
- Rule selection and partition qualification complete before token assignment
  and output serialization. Duplicate source paths reject before output.
- Token-backed mask/stub actions are explicit. Ordinary masks remain literal;
  other token-backed actions, ambiguous literal/token parameters, duplicate rule
  IDs and mismatched profile digests reject.
- The built-in tokenized review profile binds exact path rules to the retained
  snapshot's subject/record references. The empty registry wrapper and unused
  incident argument are removed; profile resolution requires snapshot content.
  Rule and profile hashes include the explicit token parameters.
- Subject binding uses an index over explicit immutable record/source references,
  rather than manufacturing a subject from a field's path spelling. Canonical
  host/identity/party and typed mention references use REQ-RPT-065 identities.
  Display tokens follow sorted distinct subject references with per-family
  ordinals, or a single `SUBJECT` sequence for an explicitly neutral profile.
  Duplicate appearances reuse a token; dropped subjects consume no ordinal.
  Release/preview identity remains part of token IDs, separate from display labels.
- Closed render-validation reasons survive the render-candidate boundary.

The existing redaction-precedence catalog row now additionally exercises literal
versus token-backed rules, family/neutral ordering, repeated subjects, dropped
subjects, array-order independence, preview/release separation, missing/ambiguous
and cross-snapshot references, invalid stable identities and safe token evidence.
The recipient-partition row covers all four retaining actions at all three rule
precedences, plus explicit-drop identity and unsupported curated narratives.
No test was removed. Both reviews are implementation review by Codex under the
user's instruction; accountable owner adoption review remains open.

`remediation-reporting-partition-focused-20261014-01` passed its three explicit
redaction rows (3.920 s, outer exit 0). The first token-focused invocation
`remediation-reporting-token-focused-20261014-01` rejected an invalid requested
row (`module.reporting.unit.canonical_digests`) with `config / configuration_error`
and outer exit 2, before a run summary was published; no product test ran.
`explain-run` therefore had no summary to inspect. The corrected exact selection
in `remediation-reporting-token-focused-20261014-02` passed four rows (4.021 s,
outer exit 0). Later staged-selection changes require the next validation run.

Remaining redaction blockers include the flattened emitted model, complete
Table 13 profile/manifest shapes, canonical path/content-class projections,
scope-derived effective partition sets and all-element partition filtering,
source-owner fact consumption, aggregate-public proof, structural pruning and
ordinal recomputation. The old source-partition vocabulary is not reclassified
by this slice. No retained artifact from the former token algorithm establishes
current token assurance; final source-specific integrated qualification remains
required.

### Token slice validation and retained failure dispositions

`remediation-reporting-token-finalize-20261014-01` passed (18.589 s, exit 0).
The first full Reporting owner invocation,
`remediation-reporting-token-owner-20261014-01`, failed at Go compilation because
one new regression assertion assigned three values from a two-result function.
The normalized class was `infra / preflight_error`, outer exit 2; all 21 cleanup
steps completed. This was an implementation-edit defect, not evidence of
intermittent product behavior. After correcting it, the two redaction rows passed
in `remediation-reporting-token-focused-20261014-03` (2.612 s, exit 0).

`remediation-reporting-token-finalize-20261014-02` passed (18.012 s, exit 0).
Both finalizers skipped retained-run maintenance because `RESULTS_DIR` was unset.
The complete Reporting owner passed in
`remediation-reporting-token-owner-20261014-02` (20.199 s, exit 0): eight units,
all 20 cleanup steps completed. The source digest was
`sha256:afd79744dc5ad94407398a24208e762aa4802ccbe82dc5819716cca659d742dd`
and graph digest
`sha256:09e4d5adaf96068f1a0f2970e9f28cfc91b23ba906ad26221dc3cac9f93aea38`.

The companion checks were launched concurrently at that same source:

- `remediation-reporting-token-portability-20261014-01` passed the exact
  Incident Bundles round-trip row (39.688 s, exit 0), including retained
  historical Reporting artifacts and sensitive reveal-map exclusion.
- `remediation-reporting-token-composition-20261014-01` passed all five work
  units, then failed overall with `harness / cleanup_error` (17.295 s, outer
  exit 2). Seventeen cleanup steps completed; `runtime_close` failed with no
  retained substep diagnostic. The exact `cleanup-results.json` and run summary
  preserve this result. No test assertion failed, and the runtime-close cause is
  unresolved. Concurrent companion/lint execution is context, not an established
  cause. This is one observed cleanup failure, not a confirmed intermittent test.
- `remediation-reporting-token-lint-20261014-01` failed `lint-go-staticcheck`
  (`SA4006`) for the initial unused profile-digest assignment in the recipient
  test; all 20 cleanup steps completed. The redundant assignment was removed.

All roots above are under `.cartulary/test-results/`. Failure evidence is retained;
no root was cleared, no services were reset, and no automatic retry policy was
introduced. Fresh validation after the lint correction and a separate Composition
invocation are required. A later pass will not attribute or erase the cleanup
incident. Historical results remain source-specific, not frozen-source release
qualification.

After the lint correction, `remediation-reporting-token-finalize-20261014-03`
passed (17.884 s, outer exit 0; retained-run maintenance skipped because
`RESULTS_DIR` was unset). The isolated full Composition owner invocation
`remediation-reporting-token-composition-20261014-02` passed all five units and all
18 cleanup steps (13.988 s, outer exit 0), at source
`sha256:2c7ba840e52edadda19d4725b2ac5840556702054526561da92b0036ffaa3074`.
`remediation-reporting-token-lint-20261014-02` passed (21.038 s, outer exit 0).
The earlier portability invocation completed all 14 cleanup steps. The source
change between the companion invocations removed only the redundant assignment
in the Reporting regression test; execution overlap also differed. These are
separate observations, not a measured failure rate or proof of cleanup stability.
The earlier cleanup cause remains unresolved; no harness or timeout change was
made to obtain the isolated pass.

## Declared source identity through materialization

Candidate REQ-RPT-026e makes the provider identity/consistency boundary explicit.
The materializer now consumes validated bound fields from declared RecordFacts;
it no longer discards record bindings and reconstructs IDs from path text.
Missing, duplicate, overlapping, dangling or foreign bindings reject the complete
contribution. Record partition/support unions and declared subject, Timeline,
partition and support references are checked against the field inventory.
Cross-provider support targets remain valid for aggregate resolution, and the
Links adapter's repeated-source-target contract is preserved.

The production `exportRecordID` path parser was removed. Snapshot record and
relationship IDs now use the declared owner identity. Incident metadata no
longer produces fabricated records named `title`, `status`, and similar member
names. An adversarial fixture changes a selector's spelling while retaining the
owner ID; materialization must still produce one record with all of its fields.
A missing owner identity fails before model bytes. Existing test-only fixture
shorthand explicitly prepares source identities before calling production code;
there is no production compatibility inference.

`TestProviderOutputValidateRejectsFlatOrIncompleteFacts` now checks incomplete,
conflicting and foreign fact bindings through the existing provider catalog row.
`TestSnapshotRecordMaterialization` additionally checks source identity independent
of path spelling. Both remain on their original owners/runners/fixtures; no test
was removed. `remediation-reporting-owner-facts-focused-20261014-01` passed the
four selected provider/materialization/redaction rows (3.011 s, outer exit 0).
Full owner and companion validation are still required for this slice.

This fixes declared record identity consumption; it does not claim that every
provider is fully typed or that all Subject/Timeline/provider metadata is already
consumed in the canonical model. Template-dependent section derivation and the
flat post-redaction model remain open structural work. The source-owner query
helper's own path-based fact construction also remains to be replaced by explicit
owner fact construction; central materialization no longer depends on that choice.

The first full owner run for declared bindings,
`remediation-reporting-owner-facts-owner-20261014-01`, passed (23.418 s,
outer exit 0) after finalizer `remediation-reporting-owner-facts-finalize-20261014-01`
(17.869 s, exit 0). A subsequent tightening preserves the declared family/type
through the redaction projection too, rejects cross-snapshot/orphaned source
bindings there, and removes both the unused `CompatibilityFields` method and
production population of the noncanonical field cache. The in-memory legacy
fixture-input member still exists; this is not completion of the structured
post-redaction-model cutover.

`remediation-reporting-owner-facts-focused-20261014-02` then failed (2.482 s,
`product / test_assertion_failure`, outer exit 2, all 10 cleanup steps completed):
two token tests expected token-stage failures, while the new source validator
correctly rejected the fixtures earlier with `snapshot_ref_mismatch` and
`dangling_source_ref`. Exact owning failure-code assertions replaced those stale
expectations; no acceptance predicate was weakened. The next four-row selection,
`remediation-reporting-owner-facts-focused-20261014-03`, passed (2.251 s, exit 0).
Finalizer `remediation-reporting-owner-facts-finalize-20261014-02` also passed
(18.247 s, exit 0). Both finalizers skipped retained-run maintenance because
`RESULTS_DIR` was unset.

The current slice additionally rejects conflicting record types for one source
identity and rejects fabricated incident-metadata records at render identity
validation. No reader repairs those historical invalid shapes. The existing
identity tests cover missing type, mismatched source ID and the retired fabricated
metadata record posture. Final full-owner, companion and lint results follow
below when available.

Final source-binding validation completed after the interrupted session:
`remediation-reporting-owner-facts-focused-20261014-04` passed (2.676 s),
`remediation-reporting-owner-facts-finalize-20261014-03` passed (17.897 s),
and the complete Reporting owner `remediation-reporting-owner-facts-owner-20261014-02`
passed (23.272 s). The Composition owner passed in
`remediation-reporting-owner-facts-composition-20261014-01` (11.036 s); the
selected Incident Bundles portability row passed in
`remediation-reporting-owner-facts-portability-20261014-01` (42.700 s). Lint
passed in `remediation-reporting-owner-facts-lint-20261014-01` (29.535 s).
All owning Make exits were 0. Finalization again skipped retained-run maintenance
because `RESULTS_DIR` was unset. The retained summary for format invocation
`remediation-reporting-owner-facts-format-20261014-06` reports pass (4.762 s);
its native exit was not recovered after the interrupted tool response.
These are source-specific checks, not coordinated adoption or full canonical
model/rendering assurance.

## Source-owner construction without selector inference

The provider constructor now accepts explicit bound fields; every SQL and typed
reader adapter supplies the source identity and declared source type. The query
contract has five columns (ID, family, type, class, value). Selectors remain
opaque identifiers. `recordIDFromPath`, the unused `CollectQueryFieldsTx` and
Records/Artifacts `CollectFieldsTx` wrappers, the `Field` alias and the lossy
`ProviderOutput.Fields` accessor were removed. There is no compatibility adapter.
The operational provider schemas retain their existing members; this is a
compile-time port cutover with stricter admission, not a historical JSON reader.

Provider construction and validation preserve complete subject/Timeline field
membership and metadata unions, canonical subject-reference grammar and exact
source identity. Missing declarations and contradictory support targets reject.
Construction returns a zero contribution on failure, even after reading earlier
valid fields. Query decoding uses the shared strict JSON boundary and preserves
integer tokens instead of passing source values through `float64`. These checks
do not assert that auxiliary source facts (record envelopes, tags and mentions)
already satisfy canonical Core record-summary eligibility; that mapping and the
complete canonical post-redaction model remain open.

`TestProviderOutputValidateAndBindings` succeeds
`TestProviderOutputValidateAndFields` in the same catalog row, package, unit
runner and fixture context. It retains the valid-contribution and copy-isolation
assertions while exercising the validated bound accessor. Added cases cover
renamed selectors, one-record/multiple-field grouping, source-order independence,
repeated Links support targets, missing/conflicting owner identity, canonical
mention identity, missing or incomplete derived declarations, exact large integer
decoding, and query/scan/iteration failure without partial publication. The
removed wrappers had no callers. Behavior review: candidate REQ-RPT-026e and
existing source-boundary/identity rules; execution review: unchanged owner row
`module.reporting.unit.provider_output_validate_and_fields_1d0b913cab`. No test
assurance or integration context was removed.

Formatting and generation passed in the owner-construction roots (4.881 s and
9.030 s). The four-row provider/materialization/boundary/composition selection
`remediation-reporting-owner-construction-focused-20261014-01` passed (3.226 s,
outer exit 0). Full owner, companions and generated-artifact checks follow.

The full Reporting owner passed at
`remediation-reporting-owner-construction-owner-20261014-01` (24.304 s),
Composition at `remediation-reporting-owner-construction-composition-20261014-01`
(16.543 s), and the selected Incident Bundles portability row at
`remediation-reporting-owner-construction-portability-20261014-01` (33.902 s).
`make lint`, `make generate-drift`, `make generated-artifact-policy-check` and
`make json-shape-check` passed in the corresponding owner-construction roots
(30.129 s, 12.997 s, 0.879 s and 5.099 s). Markdown lint passed in
`remediation-reporting-owner-construction-markdown-20261014-01` (12.960 s).
All owning Make exits were 0. The required finalizer passed in
`remediation-reporting-owner-construction-finalize-20261014-01` (17.949 s);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

## Structured redaction input only

The hidden `SnapshotContent.Fields` cache and its read fallback are removed.
Redaction consumes structured blocks and declared record bindings. Missing source
references and missing, padded or unknown source classifications reject; they
are never repaired from field keys or path spelling. Snapshot materialization
rejects raw-blob, opaque-binary and generated-presentation contributions before
projection could discard those markers. The source value serializer independently
rejects binary byte values. Core 01 REQ-378 prohibits raw blob bytes in every
release scope; generated presentation is a later stage, not snapshot case data.

The existing redaction unit cases now use the actual structured materializer.
The prior cache-precedence assertion becomes structured-field consumption plus
rejection of a hostile legacy `fields` wire member through the portable codec.
Raw/opaque cases now exercise snapshot admission directly, and retained working
material exercises external redaction separately. This corrects a test-context
gap: a flag on an in-memory flat fixture did not establish rejection through
the production structured projection. The same existing catalog rows and unit
runner retain these assertions; no scenario or evidence context was deleted.

Focused root `remediation-reporting-structured-input-focused-20261014-01` failed
(2.428 s, outer exit 2, `product / test_assertion_failure`, all 10 cleanup steps
completed). The curated-support fixture removed support from the first block,
but structured materialization sorts blocks and that was the restricted field.
The fixture now selects the intended summary by exact source reference and
recomputes its source digest. Root
`remediation-reporting-structured-input-focused-20261014-02` passed (2.561 s,
outer exit 0).

A further specification conflict remains explicit: Core 01 REQ-378–380 persists
source disclosure classes (`curated_narrative`, `derived_analytic`,
`source_evidence`, `working_material`), while Reporting Table 9-C uses a different
six-token presentation vocabulary for the same `content_class` member. The
current materializer carries Core source classifications; removing heuristic
fallback does not reconcile those canonical schema vocabularies. The full model
cutover must resolve that ownership conflict with the companion owners, preserve
source eligibility without reclassification, and revise affected canonical
identities before adoption. Neither the current typed struct nor this passing
selection overrides an adopted requirement.

The first full structured-input owner run,
`remediation-reporting-structured-input-owner-20261014-01`, failed (29.907 s,
outer exit 2, `product / test_assertion_failure`; 7/8 units passed and all 20
cleanup steps completed). The reason-mapping fixture expected a retained working
field to fail post-redaction, but the selected external profile correctly drops
working material. It now uses retained curated narrative without required
support to exercise that failure stage. Raw-source rejection remains covered
at snapshot admission. Lint passed (28.159 s) and Markdown lint passed
(12.750 s) in the corresponding structured-input roots.

The final structured-input source is
`sha256:e11f1dc0aceac7db43d85742db22c7d37a737be6a286582eba626978e6f16b19`.
Missing classification now uses the existing Table 11-B
`content_class_missing / content_class_missing` pair; malformed tokens retain
`export_model_invalid / invalid_scalar_value`, and missing source references use
`dangling_source_ref`. The negative projection cases assert those exact codes.
The focused root `remediation-reporting-structured-input-focused-20261014-03`
passed (2.906 s, fresh execution, 10 completed cleanup steps). Finalizer
`remediation-reporting-structured-input-finalize-20261014-02` passed (18.184 s);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The complete Reporting owner passed at
`remediation-reporting-structured-input-owner-20261014-02` (30.455 s, 8/8 units,
one compatible cached unit, 20 completed cleanup steps). Lint passed at
`remediation-reporting-structured-input-lint-20261014-02` (28.696 s). Composition
passed at `remediation-reporting-structured-input-composition-20261014-01`
(11.011 s). Markdown lint passed at
`remediation-reporting-structured-input-markdown-20261014-02` (14.340 s).
All owning exits were 0. The earlier failed roots remain retained and do not
enter a fresh-pass denominator.

The final selected Incident Bundles portability row passed in
`remediation-reporting-structured-input-portability-20261014-01` (32.169 s,
outer exit 0). It exercises preserved Reporting artifacts and sensitive reveal-map
omission through real bundle transport. This is the selected row's assurance;
full Recovery and coordinated adoption remain open.

The handoff-only Markdown run interrupted at the tool boundary,
`remediation-reporting-restart-handoff-markdown-20261014-01`, retains a passing
tool summary (8.600 s, tool exit 0). Its outer Make process exit was not recovered
after the Windows restart and is not inferred from that summary.

## Complete subject record bindings

The new multi-field subject regression reproduced a materialization defect:
`subjectsByRef` was overwritten for every field, so the final subject kept only
the last field's disclosure labels, source-ref selector and display name.
`remediation-reporting-subject-binding-regression-20261014-01` failed as expected
(2.453 s, outer exit 2, `product / test_assertion_failure`, all 10 cleanup steps
completed). Its source was
`sha256:ff2b0ddc4ed86c19189328e148d5f484e8aead8ac6d60d241c5e8ce9cc2c037e`.
The failure is an attributed deterministic product defect, not an instability
allegation.

Materialization now derives subjects after complete record aggregation, with one
subject per declared source identity, the record's source binding/display-name
selection and the full disclosure-label union. Host and Identity records sharing
an opaque ID remain distinct. Entity-mention fields must agree on one supported
subject family; conflicting families reject before any snapshot is returned.
Record reference sets are ordered consistently before the Timeline and subject
projections consume them. The candidate's REQ-RPT-026e now states this obligation
explicitly. No schema member or historical reader is introduced.

The existing `TestSnapshotRecordMaterialization` row adds multi-field labels,
source-reference/display selection, traversal reversal, cross-family identity,
and valid/invalid/conflicting mention cases. It retains the same Go unit runner
and fixture context; no test was removed. This repairs the source aggregation
stage, not the open classification-vocabulary conflict, full canonical export
model or all-element disclosure filtering.

The first repaired selection,
`remediation-reporting-subject-binding-focused-20261014-01`, exposed an empty
array regression in the new ordering step (2.375 s, outer exit 2,
`product / test_assertion_failure`, all 10 cleanup steps completed). Reusing
`cloneStrings` converted an empty support array to null, so Timeline and record
projections differed. Sorting the newly owned aggregate slices in place preserves
their array shape without mutating source inputs. The same three-row selection
then passed in `remediation-reporting-subject-binding-focused-20261014-02`
(2.295 s, outer exit 0). The original product regression and this repair regression
remain retained; neither is an intermittent-outcome claim.

The complete Reporting owner then passed in
`remediation-reporting-subject-binding-owner-20261014-01` (29.182 s), Composition
in `remediation-reporting-subject-binding-composition-20261014-01` (20.752 s),
selected Incident Bundles portability in
`remediation-reporting-subject-binding-portability-20261014-01` (32.599 s), and
lint in `remediation-reporting-subject-binding-lint-20261014-01` (27.675 s).
All owning exits were 0. Source:
`sha256:abce6737d9a4edeb005a3a05c9c080463b527b2051a4b78007b0c9d69849eb5f`.
Finalization passed in the corresponding `finalize-20261014-01` root (18.143 s);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The Reporting evidence audit at
`remediation-reporting-subject-binding-audit-20261014-01` failed (2.395 s, outer
exit 2, `artifact / artifact_error`, all 10 cleanup steps completed), despite
successful source/graph/row compatibility. A reused unit correctly has no child
exit code, but the canonical event reducer omitted its recorded cache outcome.
The audit therefore rejected it as an unsuccessful unit. Composition's audit at
`remediation-composition-subject-binding-audit-20261014-01` passed (2.545 s,
outer exit 0). These results are retained; the Reporting failure is an attributed
evidence-consumer defect, not a product or intermittent test failure.


## Source scalar and speaker-note admission

Reporting now has a separate `field_value_to_text_v1` boundary that uses its
canonical JSON admission for integer/Unicode validity, accepts only scalars and
flat scalar arrays, preserves scalar-array order, and emits null as empty text.
Deck derivation uses this boundary instead of the diagnostic formatter. Objects,
nested arrays, malformed Unicode, controls, noncanonical number tokens and
nonintegral numbers reject with `slidev_source_invalid /
invalid_field_value_for_text`, without including input values in diagnostics.

Speaker notes are joined in declared order into one final comment. A forbidden
comment terminator rejects before escaping and before any source or bundle
bytes are returned; the serializer no longer repairs it to `--&gt;`. REQ-RPT-091b
explicitly checks the pre-escape text as well, preventing an escape pass from
laundering the delimiter. `TestSlidevFieldTextCanonicalValues`,
`TestSlidevSpeakerNotesRejectBeforeOutput`, and
`TestSlidevInvalidFieldPublishesNoBundle` are enrolled in the existing source
contract row. They establish the scalar boundary and exercised bundle failure
path, not the full deck grammar, composition anchor semantics or sandbox.

The exact row passed in `remediation-reporting-text-unit-20261014-01`
(4.077 s, native exit 0). Format passed in `remediation-reporting-text-format-20261014-01`
(4.844 s) and generation passed in `remediation-reporting-text-generate-20261014-01`
(9.160 s). Broader owner validation remains pending.

Further source review found that the current flat deck projection, incomplete
headmatter and Markdown escaping do not establish Tables 17/18 conformance.
In particular, scalar conversion does not prove literal text survives the
Markdown-to-Vue rendering pipeline. The [Slidev syntax contract](https://sli.dev/guide/syntax)
and [CommonMark escaping rules](https://spec.commonmark.org/0.31.2/#backslash-escapes)
are supporting research only. Full parser/render validation remains open and
no renderer or coordinated-adoption condition is closed by this slice.


The full Reporting owner run `remediation-reporting-text-owner-20261014-01`
failed (23.068 s, native exit 2): seven work units passed, one failed, and all
21 cleanup operations completed. Two integration rows exposed whole-record
objects in canonical fields. This was a source projection defect hidden by the
old JSON-to-display formatter, not evidence for weakening scalar admission.
The exact `explain-run` diagnosis and bounded Go runner output are retained.

REQ-RPT-026f now distinguishes owner-declared source containers from canonical
scalar fields. Materialization projects immediate named attributes into ordered
fields before redaction, retaining explicit source identities, support and
partition metadata. It rejects nested containers and cross-contribution selector
collisions; it no longer derives display labels from storage names or selectors.
Replay/recovery admission checks scalar shapes in section, record and event
copies. Source-owner field allowlists remain independently required: this is not
authority to expose arbitrary storage columns. The existing whole-row queries
outside Evidence remain an open source-boundary defect requiring closed owner
projections. Complete field-copy equality, template derivation and canonical
nested-schema admission are also not claimed by this scalar slice.

The existing snapshot materialization row now exercises scalar projection,
identity preservation, metadata, shared subject tokens, collisions and invalid
containers. Integration assertions inspect actual scalar fields and every
affected note/party manifest entry. The test-only legacy flat-field reader and
path-inferred source/classification helpers were removed; no production
compatibility reader is introduced. Focused source/redaction/materialization
validation passed in `remediation-reporting-field-projection-unit-20261014-01`
(2.795 s, native exit 0). Both previously failing service-backed integration
rows passed in `remediation-reporting-field-projection-integration-20261014-01`
(20.021 s, native exit 0). Additional frozen-model admission cases were added
after these observations and require fresh verification.

The earlier source finalizer passed in
`remediation-reporting-text-finalize-20261014-01` (18.079 s); retained-run
maintenance was skipped because `RESULTS_DIR` was unset. Markdown lint passed
in `remediation-reporting-text-markdown-20261014-01` (9.270 s) before the new
projection amendment. The latest projection format `-03` passed (4.858 s)
and generation `-01` passed (9.160 s), before the frozen-model reader change.

The completed scalar-projection slice subsequently passed full Reporting in
`remediation-reporting-field-projection-owner-20261014-01` (33.479 s), full
Composition in `remediation-reporting-field-projection-composition-20261014-01`
(22.720 s), lint in `remediation-reporting-field-projection-lint-20261014-01`
(30.939 s), and Markdown lint in
`remediation-reporting-field-projection-markdown-20261014-01` (11.010 s).
All native exits were 0. Finalization passed in
`remediation-reporting-field-projection-finalize-20261014-01` (18.514 s);
retained-run maintenance was skipped because `RESULTS_DIR` was unset.
These results cover scalar admission including the frozen-model cases, before
the following closed-query amendment.

## Closed source-query projection

REQ-RPT-026g now specifies the exact source-container members. Timeline, Records,
Parties, Mentions, Host/Identity, Task/Decision and Artifact queries select them
explicitly. Artifact projection selects only the declared family's members;
the existing excluded investigative-query/forensic-keyword families remain
excluded. Shared projection readers preserve integer tokens using strict JSON
decoding before Reporting canonical admission.

The pre-production cutover removes storage-only incident copies, redundant active
record-envelope tombstone nulls, projection day/search columns and unrelated
artifact-family nulls. It adds no legacy reader or automatic schema discovery.
Existing immutable content is not rewritten; fresh initialization and the
coordinated derivation/adoption review remain required. These changes do not
close the full canonical model or renderer promotion conditions.

The existing snapshot/release provenance integration path now alters nine tables
in its dedicated disposable database, adding an unpublished column with a
synthetic sentinel value. It checks exclusion from the snapshot and release and
the exact member set for every affected source and all six admitted artifact
families. Evidence retains its separate future-column regression. Task/Decision
contribution fixtures now use valid source member names instead of view-qualified
field keys. Verification of this closed-query slice is pending.

The closed-query slice passed its exact snapshot/replay integration row in
`remediation-reporting-closed-source-integration-20261014-01` (22.758 s), full
Reporting in `remediation-reporting-closed-source-owner-20261014-01` (41.119 s),
full Projections in `remediation-reporting-closed-source-projections-20261014-01`
(32.867 s), full Tasks/Decisions in
`remediation-reporting-closed-source-tasks-20261014-01` (58.644 s), lint in
`remediation-reporting-closed-source-lint-20261014-01` (39.745 s), and Markdown
lint in `remediation-reporting-closed-source-markdown-20261014-01` (9.240 s).
All native exits were 0. Format passed (4.957 s), generation passed (9.310 s),
and finalization passed (18.833 s) in the matching `format`, `generate`, and
`finalize` roots ending `-20261014-01`. Retained-run maintenance was skipped
because `RESULTS_DIR` was unset. These results precede frozen-copy validation.

## Frozen field-copy admission

The same frozen source has section, record and Timeline lookup views. Admission
now checks their field values and metadata together instead of validating only
their scalar shapes. Field schema, stable member name, ordinal, state/value
pairing, null raw digest and ordered non-null reference arrays are checked before
use. Repeated fields compare canonically with only their list-local ordinal
excluded; record reference unions and full Timeline copies must agree.
Duplicate ownership, omitted copies and independently altered values, support,
partitions or source bindings fail before redaction, import or recovery.

The existing materialization row now includes hostile-copy cases using freshly
computed matching digests, so rejection does not rely only on stale checksums.
Producers emit empty canonical reference arrays as `[]`, without a repair reader
for stored null arrays. This closes the exercised field boundary, not all nested
schemas, classification mapping, template derivation or rendering. Verification
of this slice is pending.

The focused frozen-field row passed in
`remediation-reporting-frozen-fields-unit-20261014-01` (2.764 s, native exit 0).
The full Reporting owner then failed in
`remediation-reporting-frozen-fields-owner-20261014-01` (42.433 s, native exit 2,
`product / test_assertion_failure`, six units passed and two failed; all 20
cleanup operations completed). Its exact `explain-run` and bounded failure
details identify two mechanisms. The curated-support test had corrupted a
required array to null; it now materializes a valid unsupported narrative to
exercise the intended redaction failure. The Evidence integration exposed
repeated Links support targets reaching canonical fields. REQ-RPT-026d requires
those repetitions in source output; Reporting materialization now coalesces
them into ordered unique arrays before frozen admission. Source contributions
remain unchanged, and stored duplicates are rejected instead of repaired.

Composition passed in
`remediation-reporting-frozen-fields-composition-20261014-01` (28.780 s), selected
portability passed in `remediation-reporting-frozen-fields-portability-20261014-01`
(55.170 s), and lint passed in `remediation-reporting-frozen-fields-lint-20261014-01`
(30.700 s), all with native exit 0. Finalization passed (18.878 s) in the matching
`finalize-20261014-01` root; maintenance was skipped because `RESULTS_DIR` was
unset. These observations precede reference coalescing; repaired verification
is pending. Neither failure is classified as intermittent instability.

The repaired three-row selection passed in
`remediation-reporting-frozen-fields-focused-20261014-02` (16.017 s, native exit 0).
Additional record-reference union and backing-array isolation cases passed in
`remediation-reporting-frozen-fields-unit-20261014-03` (2.407 s). Finalization
then passed in `remediation-reporting-frozen-fields-finalize-20261014-02`
(19.062 s); retained-run maintenance was skipped because `RESULTS_DIR` was unset.

Final slice validation used source
`sha256:af76b0af7874e14801895eccd7797fa58a70ece999a68b43b6ebf5f2e396e5e5`:

| Public command / selection | Run ID under `.cartulary/test-results/` | Result |
| --- | --- | --- |
| `make test-slice OWNER=module.reporting` | `remediation-reporting-frozen-fields-owner-20261014-02` | Pass, 45.132 s; 8/8 units, all 20 cleanup operations completed |
| `make test-slice OWNER=module.reportcomposition` | `remediation-reporting-frozen-fields-composition-20261014-02` | Pass, 29.933 s |
| `make service-backed-test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | `remediation-reporting-frozen-fields-portability-20261014-02` | Pass, 58.132 s; 3/3 units, all 14 cleanup operations completed |
| `make lint` | `remediation-reporting-frozen-fields-lint-20261014-02` | Pass, 30.967 s |
| `make lint-markdown` | `remediation-reporting-frozen-fields-markdown-20261014-01` | Pass, 12.240 s, before this evidence append |
| `make generate-drift` | `remediation-reporting-frozen-fields-drift-20261014-01` | Pass, 13.503 s |
| `make generated-artifact-policy-check` | `remediation-reporting-frozen-fields-generated-policy-20261014-01` | Pass, 1.002 s |
| `make json-shape-check` | `remediation-reporting-frozen-fields-json-shape-20261014-01` | Pass, 5.263 s |
| `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/owner-frozen-fields-20261014-02.roots.json` | `remediation-reporting-frozen-fields-audit-20261014-02` | Pass, 2.653 s |
| `make test-evidence-audit OWNER=module.reportcomposition EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/composition-frozen-fields-20261014-02.roots.json` | `remediation-composition-frozen-fields-audit-20261014-02` | Pass, 2.661 s |

Every owning native exit was 0. Every command used its displayed run ID as
`CARTULARY_TEST_RUN_ID`. The explicit manifests select only the corresponding
complete owner root above. The Reporting graph digest is
`sha256:b887f5775023b7df59daadc47dac069491f9663f1df99cfacc2eae6e8845b188`;
the selected portability graph digest is
`sha256:07752d119e29ef4c62999af0ddf22ba94a3d84564953b4994d2ef0ca325b49f9`.
These are dirty-source implementation observations and owner evidence audits,
not a clean performance portfolio, full Recovery qualification or coordinated
adoption. The earlier two failed rows remain retained with their causes above.

## Timeline source ordering and retained event admission

Core 02 REQ-02-143a.1 now specifies source-owned ordering facts at the immutable
snapshot boundary. Reporting REQ-RPT-070–073 separates these facts from redacted
display text and defines the serialized sort tuple. These amendments are part
of the coordinated review candidate; this implementation record is not adoption.

Timeline reads its closed source members and immutable `records.created_at` in
one caller-owned transaction. Its own parser supplies UTC/offset interpretation,
minute/second precision, unresolved classification and the date-entered key.
No viewer timezone, mutable projection or snapshot timestamp supplies missing
source facts. More than six fractional digits are unresolved, not rounded.
Provider output and Timeline fact contracts advance to v2; missing, duplicate,
foreign and retired contributions fail without partial materialization.

Reporting preserves the facts through `SnapshotSource`, materializes one event
per Timeline record and uses the specified typed sort tuple. Descending changes
only the activity timestamp comparison; unresolved-first changes only the bucket.
The stored key is a canonical JSON array, not a lexical ordering codec. Retained
events use the closed Table 14-B shape (`record_id`, not `timeline_event_id`),
with complete null display members until template/redaction derivation. Readers
reject missing nullable members, duplicate keys, unknown/retired members,
inconsistent sort keys, invalid facts and incorrect frozen ordering. The
redundant event-level `source_record_ref` is removed; exact record/field/reference
binding remains enforced against the snapshot's canonical record summaries.

Exact surviving coverage includes `TestTimelineReportingSortFacts` at catalog
row `module.timeline.support_unit.reporting_sort_facts_v2`,
`TestProviderOutputValidateAndBindings`, and the Timeline subtests of
`TestSnapshotRecordMaterialization`. The existing real snapshot/replay integration
now asserts immutable creation time, UTC precedence, microsecond preservation,
UTC date rollover, the exact event member set, and null snapshot display values.
The boundary manifest admits only the Timeline time parser and strict JSON
decoder needed by this source-owner contribution; broader imports remain denied.

All roots below are under `.cartulary/test-results/`; each command used the
listed run ID as `CARTULARY_TEST_RUN_ID`. Successful native exits were 0.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make generate` | `remediation-reporting-timeline-sort-generate-20261014-01` | Pass, 9.260 s |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.support_unit.reporting_sort_facts_v2` | `remediation-timeline-sort-source-unit-20261014-01` | Pass, 0.551 s |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.snapshot_record_materialization,module.reporting.unit.provider_output_validate_and_fields_1d0b913cab` | `remediation-reporting-timeline-sort-unit-20261014-03` | Pass, 2.849 s |
| `make agent-finalize` | `remediation-reporting-timeline-sort-finalize-20261014-01` | Pass, 18.808 s |
| `make test-slice OWNER=module.reporting` | `remediation-reporting-timeline-sort-owner-20261014-01` | Pass, 27.115 s; 8/8 units, 21 cleanup operations completed |
| `make test-slice OWNER=module.reportcomposition` | `remediation-reporting-timeline-sort-composition-20261014-01` | Pass, 16.052 s |
| `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | `remediation-reporting-timeline-sort-portability-20261014-01` | Pass, 39.674 s |
| `make generate-drift` | `remediation-reporting-timeline-sort-drift-20261014-01` | Pass, 22.082 s |
| `make json-shape-check` | `remediation-reporting-timeline-sort-json-shape-20261014-01` | Pass, 6.024 s |
| `make generated-artifact-policy-check` | `remediation-reporting-timeline-sort-policy-20261014-01` | Pass, 1.223 s |
| `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/reporting-timeline-sort-20261014-01.roots.json` | `remediation-reporting-timeline-sort-audit-20261014-01` | Pass, 2.639 s |
| `make test-evidence-audit OWNER=module.reportcomposition EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/composition-timeline-sort-20261014-01.roots.json` | `remediation-composition-timeline-sort-audit-20261014-01` | Pass, 2.638 s |
| `make backend-module-boundary-check` | `remediation-reporting-timeline-sort-boundary-20261014-01` | Pass, 1.890 s |
| `make agent-finalize` | `remediation-reporting-timeline-sort-finalize-20261014-02` | Pass, 18.998 s |
| `make lint` | `remediation-reporting-timeline-sort-lint-20261014-02` | Pass, 21.440 s |
| `make lint-markdown` | `remediation-reporting-timeline-sort-markdown-20261014-01` | Pass, 10.390 s before this evidence append |

The owner/portability/audit observations bind source
`sha256:c1783b464fb8a9f60d1a63e527172b18ecbd08da1599e6363064a0a7e0c1b155`.
The subsequent change admits the two exact imports in the architecture manifest
and clarifies prose ownership of display text; production Go source is unchanged. Final lint binds source
`sha256:8469fe698dcccb4787a36ebfcee6fa6a6aa0b9003828c2d0a41b601445bb988f`.
Both finalizations omitted `RESULTS_DIR`; retained-run maintenance was skipped.
These are dirty-source implementation observations, not qualified cost windows,
Core 05 publication evidence or coordinated adoption.

Two failures remain retained. `remediation-reporting-timeline-sort-unit-20261014-01`
failed in 1.209 s (native exit 2, `infra/preflight_error`) because two stale Go
references did not compile after removing the redundant event source member and
renaming the materializer result. Exact `explain-run`, typed unit evidence and
bounded compiler logs identified both; the repaired focused run `unit-20261014-02`
passed in 2.984 s. `remediation-reporting-timeline-sort-lint-20261014-01` failed in
31.462 s (native exit 2, `product/test_assertion_failure`, 10/11 units passed,
20 cleanup operations completed). The architecture row identified the two
undeclared imports, now explicitly admitted and rechecked above. Neither outcome
is evidence of intermittent instability.

This slice does not implement template-specific Timeline selection/chunking,
redacted time display, the complete canonical render model, actual Slidev/PDF
rendering or its sandbox. Those promotion conditions remain open. Existing
artifacts with the retired event shape fail admission; no compatibility reader
or inferred historical time is introduced for the pre-production cutover.

## Snapshot, render and participant memory isolation

Render binding now copies canonical source content and identity pointers instead
of sharing nested slices, values, source references, Timeline timestamps and
Reference Pack provenance with the immutable snapshot or another render.
Participant invocation receives an independent copy; admission compares the
returned bytes against the caller's original frozen model and digest. A failed
participant cannot mutate the caller's payload. This enforces existing immutable
input ownership; no public schema, stored identity or compatibility reader is
added. Exact integer tokens and null/empty distinctions survive the copy.

`TestReportingAdmittedModelIdentityAndTime` now runs a lifecycle-isolation
regression under existing catalog row
`module.reporting.unit.admitted_render_export_participant_c0da40986c`. It mutates
nested fields, arrays, source bindings, timestamps, provenance, release/preview
identity pointers and recipient lists, checks snapshot and sibling bytes, and
exercises a participant that mutates its input before returning an error.

Every root below is under `.cartulary/test-results/`, with its explicit run ID
passed as `CARTULARY_TEST_RUN_ID`; successful owning native exits were 0.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-reporting-model-isolation-format-20261014-01` | Pass, 4.898 s |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.admitted_render_export_participant_c0da40986c` | `remediation-reporting-model-isolation-unit-20261014-01` | Pass, 2.845 s |
| `make agent-finalize` | `remediation-reporting-model-isolation-finalize-20261014-01` | Pass, 19.080 s |
| `make test-slice OWNER=module.reporting` | `remediation-reporting-model-isolation-owner-20261014-01` | Pass, 37.521 s |
| `make test-slice OWNER=module.reportcomposition` | `remediation-reporting-model-isolation-composition-20261014-01` | Pass, 27.176 s |
| `make agent-finalize` | `remediation-reporting-model-isolation-finalize-20261014-02` | Pass, 18.683 s |
| `make lint` | `remediation-reporting-model-isolation-lint-20261014-02` | Pass, 21.619 s |
| `make json-shape-check` | `remediation-reporting-model-isolation-json-20261014-01` | Pass, 5.779 s |
| `make test-slice OWNER=module.reporting` | `remediation-reporting-model-isolation-owner-20261014-02` | Pass, 27.891 s |
| `make test-slice OWNER=module.reportcomposition` | `remediation-reporting-model-isolation-composition-20261014-02` | Pass, 17.595 s |
| `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/reporting-model-isolation-20261014-02.roots.json` | `remediation-reporting-model-isolation-audit-20261014-02` | Pass, 2.647 s |
| `make test-evidence-audit OWNER=module.reportcomposition EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/composition-model-isolation-20261014-02.roots.json` | `remediation-composition-model-isolation-audit-20261014-02` | Pass, 2.656 s |

The final source digest is
`sha256:01d29869032a60f27ca36ae7f1dda7279b6310bbb65bb52d0ef82225ad84784f`.
Both finalizations omitted `RESULTS_DIR`, so retained-run maintenance was skipped.
The earlier owner runs used source
`sha256:9c7e3930f8ad5b25854a5f6d04f1b9ad093870eaf63d61448904d3861d57e228`.
Their subsequent audits (`remediation-reporting-model-isolation-audit-20261014-01`
and `remediation-composition-model-isolation-audit-20261014-01`) correctly failed
with native exit 2 and `artifact/artifact_error`: the retained source differed
from the audited executable snapshot after an architecture-manifest edit. Exact
`explain-run`, unit results and bounded diagnostics established the mismatch.
Fresh owner runs and new explicit manifests above close that evidence defect.

The retained first lint root `remediation-reporting-model-isolation-lint-20261014-01`
failed in 29.649 s (native exit 2, `product/test_assertion_failure`, 10/11 units
passed, all 20 cleanup operations completed). Its architecture diagnostic named
the new test's undeclared Reference Data facade import. The exact test file is
now admitted for its typed provenance fixture; production dependency scope is
unchanged. This was a deterministic configuration omission, not instability.
Full renderer, canonical derivation, security-companion and adoption work remain
open. These observations do not qualify a clean-source performance portfolio.

## External authored-text permission boundary

The bundle boundary now receives the actual selected typed redaction profile.
After companion identity/digest admission, every external composition containing
authored text fails with `composition_invalid / authored_text_not_permitted`
unless `allow_authored_presentation_text` is true. This includes unused nodes
and all three authored roles; internal scopes and empty collections retain their
existing applicability. No files or bundle manifest are returned on denial.

`TestCompositionRenderBinding_Unit` exercises the role/default/explicit-denial
matrix. `TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration`
uses a Party captured in the snapshot as a valid external recipient: the same
immutable composition succeeds internally, then its external job fails under
the default profile and retains no output digest, redaction manifest digest or
media type. These use existing owner rows; no duplicate execution is added.

All roots below are under `.cartulary/test-results/`, selected with the literal
run ID as `CARTULARY_TEST_RUN_ID`. Owning native exits are 0 except where stated.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-reporting-authored-permission-format-20261014-03` | Pass, 5.149 s |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393` | `remediation-reporting-authored-permission-unit-20261014-01` | Pass, 2.734 s |
| `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.preview_reporting_delegation_ecdd2bc64a` | `remediation-reporting-authored-permission-integration-20261014-01` | Pass, 16.080 s |
| `make agent-finalize` | `remediation-reporting-authored-permission-finalize-20261014-02` | Retained summary and secret scan pass, 18.345 s; native exit unavailable after interruption |
| `make test-slice OWNER=module.reporting` | `remediation-reporting-authored-permission-owner-20261014-01` | Pass, 28.374 s |
| `make test-slice OWNER=module.reportcomposition` | `remediation-reporting-authored-permission-composition-20261014-01` | Pass, 18.280 s |
| `make lint` | `remediation-reporting-authored-permission-lint-20261014-01` | Pass, 25.675 s |
| `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/reporting-authored-permission-20261014-01.roots.json` | `remediation-reporting-authored-permission-audit-20261014-01` | Pass, 2.692 s |
| `make test-evidence-audit OWNER=module.reportcomposition EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/composition-authored-permission-20261014-01.roots.json` | `remediation-composition-authored-permission-audit-20261014-01` | Pass, 2.706 s |

Final owner/lint/audit evidence binds executable source
`sha256:d31e5505d9cf22af6167e6ff120811049b88fc4efd05d33fa38e59a5a2f01980`.
Finalization omitted `RESULTS_DIR`; retained-run maintenance was skipped. The
permission check does not establish correct partition filtering, redaction,
placeholder substitution or publication when permission is enabled. Those
conditions, actual rendering/sandboxing and coordinated adoption remain open.

## Source classification and section summary reconciliation

Core 01 REQ-01-377/378, Core 04 REQ-04-046a and candidate Reporting REQ-RPT-041a
now distinguish the four immutable source classifications from Reporting's
four generated-text categories. `case_fact` and `derived_summary` are retired
without aliases. Section and block schemas advance independently to v2. This
is an explicit coordinated amendment, not a claim that the former adopted
Reporting text already matched Core or implementation. Core source eligibility,
support and disclosure rules continue to govern every source-derived value.

Snapshot blocks retain source classification verbatim, summaries count actual
top-level/nested blocks, and admission rejects a generated category supplied as
source content, retired schemas, unknown classes or inconsistent counts. Profile
class selectors use the closed eight-token vocabulary. Composition prose uses
`paragraph` / `presentation_text`; it no longer claims curation or invented
support merely because it was inserted into a deck. Full authored-text redaction
and template derivation remain open.

`TestSnapshotRecordMaterialization` now exercises source-class preservation
across scopes, actual block counts with multiple fields, nested counts, external
working-material exclusion, complete summary shape and hostile retained objects
whose attacker recomputed their digests. Existing profile-admission and
composition-binding tests cover selector rejection and authored paragraph
classification. No test deletion or new execution identity was needed.

Roots are under `.cartulary/test-results/`; every command received its listed
literal `CARTULARY_TEST_RUN_ID`. All owning native exits were 0.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-reporting-content-classes-format-20261014-02` | Pass, 5.107 s |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.snapshot_record_materialization,module.reporting.unit.snapshot_reporting_rejects_duplicate_same_preced_7daa9a6a07` | `remediation-reporting-content-classes-unit-20261014-01` | Pass, 2.825 s before final test additions |
| `make agent-finalize` | `remediation-reporting-content-classes-finalize-20261014-01` | Pass, 18.797 s |
| `make test-slice OWNER=module.reporting` | `remediation-reporting-content-classes-owner-20261014-01` | Pass, 32.232 s; 8/8 units |
| `make test-slice OWNER=module.reportcomposition` | `remediation-reporting-content-classes-composition-20261014-01` | Pass, 20.146 s; 5/5 units |
| `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | `remediation-reporting-content-classes-portability-20261014-01` | Pass, 38.889 s |
| `make lint` | `remediation-reporting-content-classes-lint-20261014-01` | Pass, 31.483 s |
| `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/reporting-content-classes-owner-20261014-01.roots.json` | `remediation-reporting-content-classes-audit-20261014-01` | Pass, 2.610 s |
| `make test-evidence-audit OWNER=module.reportcomposition EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/reporting-content-classes-composition-20261014-01.roots.json` | `remediation-composition-content-classes-audit-20261014-01` | Pass, 2.615 s |

Final owner evidence binds executable source
`sha256:958efdf982a337424055952bde4389737dc85cd10bc39d216d6562b1b4cc7807`.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset. These
observations close the exercised source-classification boundary; they do not
establish complete post-redaction canonical-model derivation, renderer/sandbox
conformance, independent owner review or coordinated adoption.

## Native snapshot readiness validation

Reporting now validates retained native snapshots before server readiness in a
repeatable-read, read-only transaction. Each model must satisfy its closed shape,
digest, semantic constraints, independent row identity and original snapshot
admission payload. A recomputed digest cannot authorize changed admitted content.
Reference Data supplies a transaction-bound retained-binding validator that
checks the exact historical set and provenance without activation, pin creation,
freshness substitution or integrity-state mutation. Keyset traversal holds one
snapshot at a time and does not impose incident-portability catalog limits on
deployment history. Candidate REQ-RPT-046rv and Reference Data RP-REQ-231 describe
these owner boundaries.

The existing snapshot/reference integration case accepts historical content
after current-set activation, rejects changed content with a recomputed digest,
incorrect summaries and retired section schemas, and rejects an unavailable
pinned manifest despite an intact active replacement. The Reference Data
canonical-persistence case validates the retained binding in a read-only
transaction before and after actual backup/restore, rejects an altered set
digest and rejects a missing transaction.

Roots are under `.cartulary/test-results/`; commands use their listed literal
`CARTULARY_TEST_RUN_ID`. Native exits are 0 except where stated.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-reporting-native-readiness-format-20261014-01` | Pass, 4.928 s |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` | `remediation-reporting-native-readiness-integration-20261014-01` | Pass, 32.373 s |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | `remediation-reference-native-readiness-integration-20261014-01` | Pass, 55.480 s |
| `make agent-finalize` | `remediation-reporting-native-readiness-finalize-20261014-01` | Retained summary and secret scan pass, 18.905 s; native exit unavailable after interruption |
| `make test-slice OWNER=module.reporting` | `remediation-reporting-native-readiness-owner-20261014-01` | Pass, 37.116 s |
| `make lint` | `remediation-reporting-native-readiness-lint-20261014-01` | Pass, 38.579 s |
| `make test-slice OWNER=app.server` | `remediation-reporting-native-readiness-server-20261014-01` | Failed, 136.803 s; native Make exit 2; `infra / service_readiness_timeout` |
| `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/reporting-native-readiness-owner-20261014-01.roots.json` | `remediation-reporting-native-readiness-audit-20261014-01` | Pass, 2.507 s |

Final owner/lint/audit evidence binds executable source
`sha256:dc4fcb02f8017732c863fe84bc28ca74748f6a7475cc3d3c18377cc0fe4e293d`.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The server run completed 17/26 units successfully. Eight units failed during
shared object-store fixture acquisition, before their selected tests executed;
browser collation additionally reported `artifact / artifact_error` for missing
group results. Its exact `service-scope.json` beneath
`_shared/test-services/7afc7f24e71bbff791d54c85` reports initial-lane readiness
expiration after 241 probes (240 transport-unreachable, one operation timeout).
The service failure reports cleanup failed; the owning run separately reports
35 cleanup steps with none failed or blocked. Neither observation is suppressed.
`make explain-run` succeeded for this exact root. The retained evidence does not
attribute the infrastructure failure to Reporting, establish intermittency, or
establish a passing server owner. That assurance remains open.

This slice does not validate native release bundles, preview outputs or all
pending jobs, and does not rehearse a tampered full server restart or complete
Reporting backup/restore. Renderer, canonical derivation, security-companion and
coordinated adoption work remain open.

## Closed snapshot Job admission

`snapshot_payload.go` now validates the entire retained input under the current
snapshot Job kind, including nested model shape, digest, generated identity,
capture time, independent row bindings and independent Jobs admission. Unknown,
missing or malformed members fail before worker publication or replay. Revisions
owns `sourceboundary.ValidateRetained`, which recovers the producer's exact
ordered bytes from JSONB formatting only after checking the complete closed
value and original token. Reporting does not duplicate source-boundary hashing.
Completion of an existing snapshot now validates its original admission binding.

Successful private Jobs tombstones remain compatible with native snapshot
history. The reader validates immutable retained identity and owner input while
leaving erased public response/idempotency fields erased. A live retained Job
also requires its exact normalized-request digest and idempotency identity.
Opaque transaction text is preserved exactly. Core 02 REQ-02-143b and candidate
Reporting REQ-RPT-046rv clarify these storage and retention boundaries.

`TestReportingFrozenSnapshotPayload` has its own owner-catalog row. Its cases
cover required members, unknown/duplicate nested input, retired models, digest
and row contradictions, Jobs admission contradictions and tombstones. Revisions'
existing source-boundary row covers producer vectors, storage formatting,
required members, duplicates, unknowns, foreign incidents and altered tokens.
The snapshot/reference integration case additionally mutates retained admission
input and verifies rejection, then uses a database-valid compacted tombstone
fixture to verify that public Job expiry does not invalidate snapshot history.
This models the compaction result; it does not test the Jobs expiry scheduler.

All roots below are under `.cartulary/test-results/` with literal
`CARTULARY_TEST_RUN_ID`. Native exits are 0 except for the stated failures.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-snapshot-payload-format-20261014-02` | Pass, 4.937 s |
| `make generate` | `remediation-snapshot-payload-generate-20261014-01` | Pass, 9.540 s |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.frozen_snapshot_payload` | `remediation-snapshot-payload-unit-20261014-01` | Failed, 6.684 s; Make exit 2, `product / test_assertion_failure`; incorrect padded-identity expectation |
| `make test-slice OWNER=module.revisions ROWS=module.revisions.unit.source_boundary_v1_bytes_hash_and_invalid_inputs_4a71df3ac8` | `remediation-snapshot-boundary-unit-20261014-01` | Pass, 7.754 s |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` | `remediation-snapshot-payload-integration-20261014-01` | Failed, 22.809 s; Make exit 2, `product / test_assertion_failure`; invalid tombstone fixture |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.frozen_snapshot_payload` | `remediation-snapshot-payload-unit-20261014-02` | Pass after fixture/expectation corrections, 4.023 s |
| `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` | `remediation-snapshot-payload-integration-20261014-02` | Pass, 18.997 s |
| `make agent-finalize` | `remediation-snapshot-payload-finalize-20261014-01` | Pass, 18.704 s; retained-run maintenance skipped, `RESULTS_DIR` unset |
| `make test-slice OWNER=module.reporting` | `remediation-snapshot-payload-owner-20261014-01` | Pass, 33.229 s |
| `make lint` | `remediation-snapshot-payload-lint-20261014-01` | Pass, 31.435 s |
| `make test-slice OWNER=app.server ROWS=app.server.integration.http_harness_restart_storage` | `remediation-snapshot-payload-server-reproduction-20261014-01` | Pass, 25.320 s; narrowed selection, not full server closure |
| `make generate-drift` | `remediation-snapshot-payload-drift-20261014-01` | Pass, 22.158 s |
| `make json-shape-check` | `remediation-snapshot-payload-json-20261014-01` | Pass, 4.936 s |
| `make generated-artifact-policy-check` | `remediation-snapshot-payload-policy-20261014-01` | Pass, 0.928 s |
| `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/reporting-snapshot-payload-owner-20261014-01.roots.json` | `remediation-snapshot-payload-audit-20261014-01` | Pass, 2.750 s |
| `make test-slice OWNER=module.reportcomposition` | `remediation-snapshot-payload-composition-20261014-01` | Failed, 134.530 s; Make exit 2, `infra / service_readiness_timeout` |

Final executable source is
`sha256:596327e9e3ad2d09fe19e7c0d76f2a47bacf9414337a1e8896368595d82f744e`.
The first two failures were diagnosed through their exact retained runs and
bounded logs before correction: opaque padded identifiers must be preserved,
and `jobs_expiry_tombstone_ck` requires all attempt fields to be cleared. No
product validation was weakened to make those fixtures pass.

Composition completed 3/5 units; its two service-backed units failed before test
execution. Its exact service scope
`_shared/test-services/51a2501256bd36c80485628f/service-scope.json` reports the same
initial object-store readiness symptom as the earlier full server run: 241
probes, 240 transport-unreachable and one timeout, with the container reported
running and not OOM-killed. Service cleanup reports failed; owning run cleanup
reports 14 steps with none failed/blocked. `make explain-run` succeeded for the
exact root. The narrow server pass has a different source/selection context from
the earlier failure and is not proof of intermittency or full server closure.
Composition owner assurance and infrastructure cause attribution remain open.

This closes the exercised snapshot-input boundary, not all pending-job recovery,
native release/preview validation, renderer conformance or coordinated adoption.

## Private reveal-map separation and recovery

The renderer no longer receives reveal-map bytes. Release and preview maps are
stored atomically with their parent output in separate Reporting-owned tables,
`reporting_release_reveal_maps` and `reporting_preview_reveal_maps`. Runtime
access is SELECT/INSERT; full recovery has restore privileges. Parent retention
cascades through explicit foreign keys. No public sensitive-artifact read route
was added. Ordinary manifests, file storage and portable import reject sensitive
roles, media types and the internal namespace; unowned file roles fail closed.

The private admission check validates exact canonical bytes and schema-domain
digests, render identity, snapshot, admission time, token-manifest binding,
complete token/path coverage and original-value digests. The shareable bundle
check independently validates its manifest projection, child bytes and token
manifest digest. Native readiness traverses retained release and preview parents
in a repeatable-read, read-only transaction, without portable-catalog cardinality
limits. Missing or altered required maps fail readiness without reconstruction.
This does not replace complete native release, preview or pending-Job validation.

Reporting REQ-RPT-069, REQ-RPT-046pa and the Core 01/04 companion clauses now
state the separation. The portable catalog advances to
`cartulary.reporting_portable_artifacts.v2`; the retired v1 catalog and
`omitted_files` member reject. Bundle v2 already excluded the sensitive role,
so this correction retains its identity. Migration 69 and the preflight reject
retained Reporting history before mutation; no legacy reader, row conversion,
synthetic identity or dual writer was introduced. Both 62→63 and 68→69 rejection
paths preserve data and migration state. No developer database was reset.

Recovery's authored catalog includes both private tables and has digest
`b31d0d790236eab6776cd75c3f868b72a5cfd28fb6caf0097a93e5959fee8f35`.
Its required table inventory and dependent Graph restore binding were updated;
Go, SQL and ownership projections were regenerated through Make. The new private
storage does not add object-store families or change incident-portability scope.

The existing renderer artifact entrypoint is now
`TestRenderBundleSeparatesSensitiveRedactionArtifacts`, replacing
`TestRenderBundleCarriesRedactionArtifactsWithSensitiveRevealRole` in catalog row
`module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393`.
It preserves the redaction/profile/token artifact assertions and replaces the
incorrect reveal-member assertion with exclusion and adversarial private-map
admission checks. Release integration verifies the separately retained canonical
map. The preview integration path now includes a tokenized preview and release,
then a full encrypted backup and restore into the explicitly disposable
`reporting-private-artifact-restore` harness database and temporary object roots.
It uses the production database purpose-role initializer, compares restored
release/preview bytes, digests and timestamps, validates retained snapshot
admission and Reference Pack bindings, and demonstrates that deleting either
required map fails readiness until its original bytes are restored. This is a
fresh harness rehearsal, not an operator deployment rehearsal or renderer proof.
The expanded preview workload is not comparable to its previous timing samples.

Implementation/equivalence review: Codex, this authorized slice. Independent
behavior-owner and harness-owner adoption review remains open. No original
security or sensitive-retention obligation was retired. The broader renderer,
subject-schema, model-derivation and adoption blockers above remain open.

Verification source is dirty HEAD
`00d24aa5354284df6533b8f650e9eb0291ad44`, final execution digest
`sha256:e09864a795a0e52639da53821e6689308cbdc71b16f0d3df88574df19f7553e5`.
Run IDs below are exact children of `.cartulary/test-results/`; all successful
commands also completed with independently observed native exit 0. The owning
summaries have null failure classification. No evidence here qualifies a clean
source performance window or Core 05 publication claim.

| Command / selected scope | Exact run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-private-artifacts-format-20261014-05` | Pass, 4.941 s |
| `make generate` | `remediation-private-artifacts-generate-20261014-03` | Pass, 9.310 s |
| Reporting artifact and portable-catalog unit rows | `remediation-private-artifacts-unit-20261014-01` | Pass, 5.515 s; earlier slice source |
| Reporting snapshot/cutover/release integration row | `remediation-private-artifacts-integration-20261014-02` | Pass, 20.069 s; earlier slice source |
| Preview and full private-artifact restore row | `remediation-private-artifacts-restore-20261014-02` | Pass, 14.225 s |
| Recovery catalog and parallel codec rows | `remediation-private-artifacts-recovery-20261014-01` | Pass, 4.187 s |
| `make agent-finalize` | `remediation-private-artifacts-finalize-20261014-01` | Pass, 18.510 s; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make test-slice OWNER=module.reporting` | `remediation-private-artifacts-reporting-owner-20261014-01` | Pass, 36.745 s; no cache hits; 21 cleanup steps completed |
| `make test-slice OWNER=module.reportcomposition` | `remediation-private-artifacts-composition-owner-20261014-01` | Pass, 15.116 s; no cache hits; 18 cleanup steps completed |
| `make test-slice OWNER=app.server` | `remediation-private-artifacts-server-owner-20261014-01` | Pass, 55.482 s; no cache hits; 51 cleanup steps completed |
| DDL recurrence/catalog row | `remediation-private-artifacts-ddl-20261014-01` | Pass, 19.337 s; 14 cleanup steps completed |
| Incident-bundle reference/artifact round-trip row | `remediation-private-artifacts-portability-20261014-01` | Pass, 35.151 s; 14 cleanup steps completed |
| `make lint` | `remediation-private-artifacts-lint-20261014-01` | Pass, 36.491 s |
| `make generate-drift` | `remediation-private-artifacts-drift-20261014-01` | Pass, 12.952 s |
| `make json-shape-check` | `remediation-private-artifacts-json-20261014-01` | Pass, 5.085 s |
| `make generated-artifact-policy-check` | `remediation-private-artifacts-policy-20261014-01` | Pass, 0.897 s |
| Reporting owner evidence audit | `remediation-private-artifacts-reporting-audit-20261014-01` | Pass, 2.546 s |
| Composition owner evidence audit | `remediation-private-artifacts-composition-audit-20261014-01` | Pass, 2.788 s |
| Server owner evidence audit | `remediation-private-artifacts-server-audit-20261014-01` | Pass, 2.616 s |

Owner audits used the explicit root manifests in `.cartulary/remediation-evidence/`:
`reporting-private-artifacts-owner-20261014-01.roots.json`,
`composition-private-artifacts-owner-20261014-01.roots.json`, and
`server-private-artifacts-owner-20261014-01.roots.json`. Targeted rows were selected
through the public owner catalogs; the canonical manifests retain exact commands,
selection, toolchain, execution contexts and artifacts.

Retained failures and corrections:

- `remediation-private-artifacts-format-20261014-01`: native Make exit 2,
  `config/configuration_error` before root publication. `explain-run` confirmed no
  summary; `explain-target` identified the renamed test's unsorted selector list.
  Corrected the authored list. No formatting ran in this failed attempt.
- `remediation-private-artifacts-generate-20261014-01`: native Make exit 2,
  tool summary `harness/unknown_failure`, 1.625 s. The DDL policy rejected unnamed
  constraints and missing explicit foreign-key update actions. Corrected the
  authored migration; successful generate roots `-02` and `-03` retain the result.
- `remediation-private-artifacts-integration-20261014-01`: native Make exit 2,
  `product/test_assertion_failure`, 19.616 s. Both cutover rejection cases passed;
  application startup failed because the new private tables lacked role grants.
  Added only required runtime/recovery privileges and aligned the ownership
  projection. All 14 owning cleanup steps completed.
- `remediation-private-artifacts-preview-20261014-01`: native Make exit 2,
  `product/test_assertion_failure`, 11.982 s. The new fixture incorrectly resolved
  a subject-dependent tokenized profile against an empty snapshot. It now uses
  the retained admitted model. The targeted `-02` passed in 12.654 s. The failed
  attempt's 14 cleanup steps completed; no production check was weakened.
- `remediation-private-artifacts-restore-20261014-01`: native Make exit 2,
  `product/test_assertion_failure`, 12.972 s. The fixture opened a NOINHERIT
  recovery login without the production role initializer; capture failed while
  streaming its first table. The reported rooted-filesystem error hid its
  callback cause. Replaced the raw pool with `postgres.Setup` for both source
  and target recovery connections. The complete rehearsal then passed. The
  failed attempt's 14 cleanup steps completed.

### Earlier object-store investigation remains unattributed

The prior full-server and Composition readiness failures remain separate
infrastructure incidents; these new-source owner passes do not erase them or
establish intermittent behavior under identical conditions. The diagnostic
`remediation-object-store-preview-diagnosis-20261014-01` used only the preview
row with `HARNESS_DIAGNOSTICS=basic` and passed in 11.562 s, native exit 0. Its
selection and policy differ from the failed full owner.

A bounded, read-only Docker inspection of the failed Composition fixture found
that its S3 process logged HTTPS listener startup while the harness probes
reported transport-unreachable outcomes. The container had no OOM indication;
shutdown later exceeded its close interval. A diagnostic TCP probe occurred
after the successful diagnostic run's teardown and is not readiness-failure
evidence. No containers were manually removed, services reset or retry policy
changed. Host transport and cleanup cause attribution remain open.

## Native release completion and recovery

Reporting now validates release completion against the original closed admission
before writing output or accepting an existing completion. The comparison binds
the derived release identity, actor/incident/transaction, immutable snapshot and
its original admission, source digest/boundary, fixed admission time, template,
profile, output options, graph selection, composition tuple and recipients.
A recomputed standalone manifest digest cannot replace those independent facts.
Release payload decoding preserves exact JSON integers with `UseNumber`.

The new native release reader uses one repeatable-read, read-only transaction
and bounded release traversal. It validates the matching terminal Jobs outcome,
root and member inventories, actual bytes, redaction/profile artifacts and
private reveal map, and checks the exact Reference Pack binding through its
owner. A failed release must have its retained failure reason and no successful
output. Existing completion cannot change outcome, bytes, or failure reason.
Legitimate expired success/failure tombstones remain usable without recreating
public responses or idempotency state. Approval, publication and invalidation
remain independent lifecycle transitions over immutable admission and output.
The reader is installed before server readiness and exercised after the actual
encrypted backup/restore rehearsal.

Composition now checks its own retained release binding inside the caller's
transaction. A release has exactly the declared tuple/scope or no binding when
none was declared. Duplicate inserts no longer hide changed digests or scope.
Reporting validates frozen composition bytes against the exact retained version
through that owner, without reading Composition tables itself. Missing pins are
rejected, never recreated by validation. No schema migration, compatibility
reader, new public response format or artifact identity was introduced by this
slice. The changed behavior corrects admission/replay integrity of the current
candidate formats.

Authority: proposed Reporting REQ-RPT-046rr and Composition REQ-RC-025 clarify
these boundaries. Implementation review by Codex is not independent human
adoption approval. Reporting 1.4.0 and Composition 1.3.0 remain proposed.

### Obligation evidence and applicability

| Obligation | Exact exercising tests and context | Status and limit |
| --- | --- | --- |
| Immutable release input, model, output and lifecycle binding | `TestReportingFrozenReleasePayload`, including `assertFrozenReleaseBindings`; pure owner tests with self-consistently rehashed substitutions | Demonstrated for current typed input and stored output bindings; no claim of complete renderer or redaction semantics |
| Repeated successful completion rejects altered retained data | `TestSnapshotReplayAndReleaseProvenanceAreStable_Integration` through `assertNativeReleaseReadiness`; actual authenticated snapshot/release creation, PostgreSQL, stored token/private artifacts | Altered transaction, profile/output digest, unknown admission member, stored bytes and Jobs result reject; identical replay and expired successful history pass |
| Failed completion preserves failure identity | `TestBoundaryReplayDefaultsAndActionIdempotency_Integration` through `assertFailedReleaseReadiness` | Identical failure replay passes, replacement reason rejects, expired failed history passes; fixture models the Jobs tombstone, not its expiry scheduler |
| Composition owns historical binding | `TestBindReleaseTupleTxRecordsReleaseBoundVersion` and `TestResolveReleaseTupleTxReasonMatrix` | Missing/unclaimed bindings and changed duplicate scope/digest reject; retained version bytes are independently validated |
| Full recovery retains native bindings | `TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration`, including the encrypted backup/restore helper | Actual release/snapshot/private-artifact history passes native readers after restore into explicitly disposable `reporting-private-artifact-restore` database and isolated roots |
| Startup integration | Full `app.server` owner | Owner suite passes with the reader installed; not full release qualification |

No test or scenario was removed in this slice. Existing catalog rows retain
these assertions. Integration workloads expanded; their old duration windows
are not equivalent cost evidence. Preview completion, complete preview/pending
Job recovery, actual renderer/sandbox execution, remaining source-model fidelity,
qualified performance windows and coordinated adoption are still open. The
existing all-zero failed-profile sentinel also remains a pre-production design
debt: it is now accepted only when profile resolution itself fails, and is never
accepted as a successful profile digest. This historical limitation is resolved
by the frozen-profile admission slice recorded below.

### Retained execution evidence

All roots below are under `.cartulary/test-results/`. Native Make exit 0 was
observed for each successful command. Full-owner evidence shares cleanly bound
**dirty-worktree** source digest
`sha256:f5420061d9bfee0e4b8451663d760de19126a7ae55e0e7509c66fec72f67fb1a`
at HEAD `00d24aa5354284df6533b8f650e9eb0291ad44`; this is implementation evidence,
not a clean publication/performance window.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-release-binding-format-20261014-03` | Pass, 5.098 s; earlier format roots `-01` and `-02` also passed |
| Exact frozen-release-payload row | `remediation-release-binding-unit-20261014-02` | Pass, 4.299 s |
| Exact provenance and route-idempotency rows | `remediation-release-binding-integration-20261014-01` | Pass, 25.569 s; earlier slice source |
| Exact Composition tuple and preview/delegation rows | `remediation-release-binding-composition-20261014-01` | Pass, 19.539 s; includes restore rehearsal |
| `make agent-finalize` | `remediation-release-binding-finalize-20261014-01` | Pass, 19.011 s; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `make test-slice OWNER=module.reporting` | `remediation-release-binding-reporting-owner-20261014-01` | Pass, 36.067 s; 8 units, one compatible cache hit; 20 cleanup steps completed |
| `make test-slice OWNER=module.reportcomposition` | `remediation-release-binding-composition-owner-20261014-01` | Pass, 13.337 s; 5 units, no cache hits; 18 cleanup steps completed |
| `make test-slice OWNER=app.server` | `remediation-release-binding-server-owner-20261014-01` | Pass, 54.913 s; 26 units, no cache hits; 51 cleanup steps completed |
| `make lint` | `remediation-release-binding-lint-20261014-01` | Pass, 29.759 s |
| `make generate-drift` | `remediation-release-binding-drift-20261014-01` | Pass, 12.957 s |
| `make json-shape-check` | `remediation-release-binding-json-20261014-01` | Pass, 4.900 s |
| `make generated-artifact-policy-check` | `remediation-release-binding-policy-20261014-01` | Pass, 0.864 s |
| Reporting owner evidence audit | `remediation-release-binding-reporting-audit-20261014-01` | Pass, 2.684 s |
| Composition owner evidence audit | `remediation-release-binding-composition-audit-20261014-01` | Pass, 2.689 s |
| Server owner evidence audit | `remediation-release-binding-server-audit-20261014-01` | Pass, 2.738 s |

Audit root manifests are
`.cartulary/remediation-evidence/reporting-release-binding-owner-20261014-01.roots.json`,
`composition-release-binding-owner-20261014-01.roots.json`, and
`server-release-binding-owner-20261014-01.roots.json`. They identify only the exact
full-owner roots above. Reused cache evidence is not another fresh execution.

Retained failure: `remediation-release-binding-unit-20261014-01` exited Make 2
with `infra/preflight_error` (row `missing_selector_result`), 0.844 s. Structured
results and its bounded referenced stderr showed a Go compile error: this slice
had redeclared the existing `sameReportingValue` helper. Removed the duplicate
and reused the existing canonical comparison. No test assertion executed in that
attempt; all ten owning cleanup steps completed. The successful `-02` and full
owner supersede that implementation error without erasing its evidence.

## Frozen release profile admission

Release admission now resolves the exact typed redaction profile against the
retained snapshot, before admitting a Job. The closed private payload retains
that profile and its schema-domain digest; a separate Reporting column binds
the admitted digest independently of the payload. Execution, completion and
native recovery use those frozen bytes instead of resolving today's profile
factory. Unknown members, missing profiles, zero/sentinel digests, mismatched
selectors and self-consistently rehashed replacement profiles reject. Failed
renders retain the admitted profile just as successful renders do. No test was
removed, and no new executable catalog or document-dependent check was added.

The durable release Job identity is now `snapshot_reporting.release_create_v3`.
Snapshot/preview kinds, worker identity, public route normalization and shared
Jobs/Auth receipt comparison remain unchanged. Authored extension/telemetry
projections and generated contracts carry the new kind. Migration 70 rejects
prior Reporting history under locks before mutation; the application preflight
also rejects it. Neither path backfills profiles or creates historical facts.
Fresh initialization and the production DDL checks pass. The cutover regression
now covers prior migration states 62, 68 and 69, asserting unchanged history and
migration state after rejection. Rollback still requires a matching application
and backup, or explicit disposable initialization; no live environment was reset.

The first implementation incorrectly added a private admission envelope to the
Jobs request hash. The integration run
`remediation-profile-admission-integration-20261014-01` failed (native exit 2,
`product` / `test_assertion_failure`, 37.565 s) because the shared finalizer
correctly rejected a digest different from the public idempotency receipt.
Three release cases remained running while their assertions waited. All 14
cleanup steps completed. The corrected design uses the separate Reporting
column and preserves the shared request digest. This is a reproduced design
error, not evidence of intermittent instability. Finalization-error swallowing
in the Reporting worker was separately identified and remains follow-up work.

### Frozen-profile obligation evidence

| Obligation | Exact exercising test / context | Result and limits |
| --- | --- | --- |
| Closed profile bytes and independent admission digest | `TestReportingFrozenReleasePayload`, including profile substitutions, unknown nested members, missing independent digest and retired release Job kind | Demonstrated by unit assertions; no claim of complete profile/redaction model fidelity |
| Successful and failed releases preserve admitted profile | `TestSnapshotReplayAndReleaseProvenanceAreStable_Integration`, `TestBoundaryReplayDefaultsAndActionIdempotency_Integration`, and native release readers | Authenticated routes, real database, repeat completion and expired history pass |
| Reject incompatible history before mutation | `TestReportingIdentityCutoverDoesNotReconstructHistory_Integration` at 62, 68 and 69 | All three histories reject without data or migration-state changes |
| Restore preserves new admission column | `TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration` and encrypted backup/restore helper | Native release validation passes after actual restore into the named disposable database |
| New Job kind is coherent with owned projections | Exact contract-generator and extension registry rows, full server owner, generated drift/policy/shape checks | Current projection and startup checks pass; human adoption fidelity review remains separate |

All run roots are under `.cartulary/test-results/`. Successful commands below
have observed native exit 0 unless explicitly qualified. Current owner evidence
binds dirty-worktree source
`sha256:e2906d130587935b796ffb7374bd169fa5de0f835fe6eadc8d4803bbdea8f497`
at HEAD `00d24aa5354284df6533b8f650e9eb0291ad44`.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format`, then `make generate` | `remediation-profile-admission-format-20261014-02`, `remediation-profile-admission-generate-20261014-02` | Pass, 4.937 s / 9.490 s |
| Exact frozen-payload, export-model and worker rows | `remediation-profile-admission-unit-20261014-02` | Pass, 3.509 s |
| Exact provenance/admission and route-idempotency rows | `remediation-profile-admission-integration-20261014-02` | Pass, 29.618 s; 14 cleanup steps completed |
| Exact preview/delegation row, including restore | `remediation-profile-admission-preview-20261014-01` | Summary reports pass, 14.920 s; 14 cleanup steps completed; original native exit unavailable after restart |
| Contract-generator validation row | `remediation-profile-admission-contractgen-20261014-01` | Pass, 0.851 s |
| `make agent-finalize` | `remediation-profile-admission-finalize-20261014-01` | Pass, 18.819 s; retained maintenance skipped because `RESULTS_DIR` was unset |
| Full `module.reporting` owner | `remediation-profile-admission-reporting-owner-20261014-01` | Pass, 32.483 s; 8 units, one compatible cache hit; 20 cleanup steps completed |
| Full `module.reportcomposition` owner | `remediation-profile-admission-composition-owner-20261014-01` | Pass, 14.714 s; 5 units, no cache hits; 18 cleanup steps completed |
| Full `app.server` owner | `remediation-profile-admission-server-owner-20261014-01` | Pass, 69.383 s; 26 units, no cache hits; 51 cleanup steps completed |
| Exact production DDL unit/integration rows | `remediation-profile-admission-ddl-20261014-01` | Pass, 15.987 s; 15 cleanup steps completed |
| Exact extension coordinator registry row | `remediation-profile-admission-extensions-20261014-01` | Pass, 2.292 s |
| `make lint` | `remediation-profile-admission-lint-20261014-01` | Pass, 37.216 s |
| `make generate-drift` | `remediation-profile-admission-drift-20261014-01` | Pass, 20.204 s |
| `make json-shape-check` | `remediation-profile-admission-json-20261014-01` | Pass, 4.888 s |
| `make generated-artifact-policy-check` | `remediation-profile-admission-policy-20261014-01` | Pass, 0.875 s |
| Reporting / Composition / server evidence audits | `remediation-profile-admission-reporting-audit-20261014-01`, `remediation-profile-admission-composition-audit-20261014-01`, `remediation-profile-admission-server-audit-20261014-01` | Pass, 2.804 / 2.785 / 2.846 s, with explicit owner root manifests |

The successful full Composition owner supplies fresh restore coverage after the
interrupted narrow run. These are implementation observations, not qualified
cost or Core 05 publication evidence. Preview still needs its own persisted
admission/recovery closure. Actual renderer/sandbox execution, remaining model
fidelity, worker failure handling, timed Jobs UI evidence, independent review,
qualified performance windows and coordinated adoption remain open. Reporting
1.4.0 and Composition 1.3.0 remain proposed.

## Reporting worker execution boundary

The Reporting worker now has one error-returning handler. Snapshot, release and
preview finalization errors reach the shared Jobs supervisor unchanged. A
finalization error never triggers a second owner completion, inferred rollback,
local transaction retry or invented success. The Extensions finalizer retains
ownership of proof-based commit classification and its fatal integrity path;
Jobs retains the adopted attempt budget, retry delays and shutdown release.

Execution observation and progress errors stop work. A fresh checkpoint after
rendering observes cancellation before publishing either success or render
failure. Parent-context shutdown returns to the supervisor without writing a
new failed Job. The independent Reporting operation deadline still has its
existing closed timeout outcome. Persistence and participant error text no
longer becomes a public Jobs message; closed safe summaries replace it. These
are corrections to Core 01 REQ-01-453, not a new retry policy or schema epoch.

`TestReportingJobWorkerExecutionBoundary` is enrolled in the existing Reporting
worker unit row. It injects success/failure-finalizer errors for the affected
operations, mutation/observation/progress/completion errors, shutdown before and
during rendering, cancellation after rendering, and unsafe diagnostic content.
Assertions establish exactly one finalizer call, preservation of the returned
error, no invented terminal state/output, and safe public summaries. Existing
participant, timeout, cancellation and integration assertions remain active.
No test was removed. The real shared-supervisor integration row separately
exercises its existing failure-budget and shutdown semantics; this does not
claim a new physical indeterminate-commit fault rehearsal in Reporting.

All successful roots below are under `.cartulary/test-results/`, with native
exit 0 observed. Current owner evidence binds dirty-worktree source
`sha256:f19e113bf1b07dfa8295a87fe9ce6fc266339729ea70f46babab9a22954d02b2`.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` / `make generate` | `remediation-worker-errors-format-20261014-01` / `remediation-worker-errors-generate-20261014-01` | Pass, 4.981 / 9.690 s |
| Exact worker and participant unit rows | `remediation-worker-errors-unit-20261014-01` | Pass, 2.956 s |
| `make agent-finalize` | `remediation-worker-errors-finalize-20261014-01` | Pass, 19.049 s; retained maintenance skipped because `RESULTS_DIR` was unset |
| Full Reporting owner | `remediation-worker-errors-reporting-owner-20261014-01` | Pass, 48.889 s; 8 units, no cache hits; 21 cleanup steps completed |
| Full Composition owner | `remediation-worker-errors-composition-owner-20261014-01` | Pass, 13.521 s; 5 units, no cache hits; 18 cleanup steps completed |
| `platform.jobs.integration.supervised_runner` | `remediation-worker-errors-supervision-20261014-01` | Pass, 41.041 s |
| `make lint` | `remediation-worker-errors-lint-20261014-01` | Pass, 31.094 s |
| `make generate-drift` | `remediation-worker-errors-drift-20261014-01` | Pass, 13.394 s |
| `make json-shape-check` | `remediation-worker-errors-json-20261014-01` | Pass, 4.977 s |
| `make generated-artifact-policy-check` | `remediation-worker-errors-policy-20261014-01` | Pass, 0.887 s |
| Reporting / Composition owner evidence audits | `remediation-worker-errors-reporting-audit-20261014-01` / `remediation-worker-errors-composition-audit-20261014-01` | Pass, 2.638 / 2.601 s; explicit root manifests |

The preceding profile-admission documentation also passed `make lint-markdown`
in `remediation-profile-admission-markdown-20261014-01` (8.860 s, native exit 0).
Worker failure swallowing is closed for these exercised paths. The remaining
preview-admission/recovery, model, renderer, performance and coordinated-adoption
work remains open. No new cost comparison or stability-rate claim is made.


## Frozen preview admission and recovery

Composition now allocates its preview-attempt identity and fixed admission time
before handing the exact selected source to Reporting. Reporting freezes a
preview-specific render input, including the full profile and independent digest,
within the same transaction as Composition source retention, Jobs admission,
Graph leases and the authenticated request receipt. Preview inputs no longer
embed a synthetic release request or acquire a release identity. Common immutable
render data has one private type shared by the two operations.

Only admission resolves the selected profile. Worker reads, completion and startup
recovery validate retained preview input against the original snapshot, exact
Reference Pack binding, Composition source and Jobs admission. Repeat completion
validates retained output before returning success. Recovery enumerates Composition
admissions so missing Reporting input cannot disappear from the scan, rejects
orphaned input/output, and validates successful output members and the Jobs result.
Legitimate expired terminal Jobs retain their immutable admission binding without
reconstructing erased public idempotency fields. The encrypted restore scenario
runs this reader too. Explicit pending-state recovery characterization remains
open; the implementation's non-success branch forbids published output.

The durable preview Job identity is now
`snapshot_reporting.composition_preview_v3`. Migration 71 rejects incompatible
history before mutation, extends the existing private input table and independent
profile constraint, and does not backfill historical facts or add a legacy reader.
No recovery-family membership changed. Composition owns the exact private request
normalization reader: absent draft version is explicitly null in that private
representation, while public explicit null remains invalid. Invalid snapshot,
profile/digest, environment and graph selections produce the closed public
`composition_preview_input_invalid` denial without committing admission.

The first narrow preview run failed because the private normalized draft request
was incorrectly passed to the public decoder. Exact root
`remediation-preview-admission-integration-20261014-01` retains native exit 2,
`product/test_assertion_failure`, the unexpected HTTP 500, and 14 completed cleanup
steps. The owner-specific normalization reader fixes that mismatch without changing
public null semantics. Subsequent narrow runs exercised real admission, completion,
tamper rejection and encrypted restore. Run `...integration-20261014-04` reports
pass (21.545 s), 14 completed cleanup steps and a passing retained-secret scan;
its native exit is unavailable after the desktop restart. The fresh full
Composition owner below supplies the complete invocation result.

The following exact roots are under `.cartulary/test-results/`; native exit 0 was
observed for each. Full owner evidence binds dirty-worktree source
`sha256:7134677cc5821a0a3d840212eb15c6eb34e7333bf289b6a3e715a8dd59d1f4fd`.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make agent-finalize` | `remediation-preview-admission-finalize-20261014-01` | Pass, 18.963 s; retained maintenance skipped because `RESULTS_DIR` was unset |
| Full Reporting owner | `remediation-preview-admission-reporting-owner-20261014-01` | Pass, 38.849 s; 8 units, one compatible cache hit; 20 cleanup steps completed |
| Full Composition owner | `remediation-preview-admission-composition-owner-20261014-01` | Pass, 17.145 s; 5 units, no cache hits; 18 cleanup steps completed |
| Full server owner | `remediation-preview-admission-server-owner-20261014-01` | Pass, 64.286 s; 26 units, no cache hits; 51 cleanup steps completed |
| Exact contractgen row | `remediation-preview-admission-contractgen-20261014-01` | Pass, 38.179 s |
| Exact extension coordinator registry row | `remediation-preview-admission-extensions-20261014-01` | Pass, 39.312 s |
| Exact production DDL unit/integration rows | `remediation-preview-admission-ddl-20261014-01` | Pass, 64.703 s |
| `make lint` | `remediation-preview-admission-lint-20261014-01` | Pass, 52.774 s |
| `make generate-drift` | `remediation-preview-admission-drift-20261014-01` | Pass, 43.427 s |
| `make json-shape-check` | `remediation-preview-admission-json-20261014-01` | Pass, 27.599 s |
| `make generated-artifact-policy-check` | `remediation-preview-admission-policy-20261014-01` | Pass, 22.525 s |
| Reporting / Composition / server evidence audits | `remediation-preview-admission-reporting-audit-20261014-01`, `remediation-preview-admission-composition-audit-20261014-01`, `remediation-preview-admission-server-audit-20261014-01` | Pass, 2.721 / 2.726 / 2.756 s; explicit root manifests |

Adversarial assertions cover unknown private members, normalized request changes,
Composition admission-time changes, output/profile substitutions, corrupted member
bytes, wrong Jobs outcomes, missing input and expired response retention. The
public-denial subcases assert no new Jobs or Composition admission rows. No test
was removed. Sources were unchanged during these runs. Concurrent verification
makes these durations engineering observations, not qualified cost evidence.

Reporting 1.4.0 and Composition 1.3.0 remain proposed. Live-execution fencing of
handler payload reads, actual renderer/sandbox execution, model fidelity, timed
Jobs UI evidence, independent review and coordinated adoption remain open. No
full current `check`, release qualification or clean-source performance portfolio
is claimed. The earlier worker documentation passed `make lint-markdown` in
`remediation-worker-errors-markdown-20261014-01` (8.530 s, native exit 0).


## Live execution fencing and unfinished-preview recovery

Reporting's handler-facing kind, release-input and preview-input readers now
accept the opaque Jobs execution, validate it under the Jobs transition lock,
and revalidate before returning copied input. Reference binding and Graph-result
reads/renewals use that same transaction. Rendering occurs after commit. There
is no UUID-only handler read fallback. The Network Flow Reporting adapter now
borrows the transaction for exact Graph reads and renewals; failed reads roll
back renewal. This is a stricter Reporting consumer policy permitted by Graph
Projection §8's separate operations, not a change to the reusable lease API.

The shared Jobs transaction guard now checks `clock_timestamp()` rather than
PostgreSQL's fixed transaction-start `now()`. The old expression could admit a
lease that expired after the transaction began. Lost execution is returned to
the supervisor without a new owner completion. Cancellation discovered between
an earlier checkpoint and payload read produces cancellation, not a failed Job.
Graph-read failures retain their closed render-failure classification.

`TestReportingPayloadReadsRequireLiveExecution_Integration`, enrolled as
`module.reporting.integration.execution_payload_boundary`, uses real Jobs claims,
replacement and terminal transitions. With the Reporting input table exclusively
locked, expired, replaced, cancellation-requested and terminal executions must
return their Jobs guard error rather than touch payload storage. A current
replacement reads successfully. A deterministic transaction fixture places lease
expiry between transaction start and current time, proving that the guard uses
current liveness. Worker tests exercise loss/cancellation between checkpoint and
read. The existing exact Graph lifecycle scenario now proves renewal rollback
when exact-result validation fails. No assertion was simply deleted.

The real preview integration additionally stops its own harness runner after
render/restore verification, admits genuine queued previews, and uses Jobs-owned
claims and terminal transitions to exercise queued, running, failed and canceled
retained inputs. Each legitimate state passes recovery, changed private input
rejects, and no output is invented. Failed/canceled compacted tombstones also
pass. The helper explicitly models compaction; it does not claim to exercise the
Jobs expiry scheduler. These observations close the pending-state characterization
left open in the preceding preview section.

Four unsuccessful development invocations remain under `.cartulary/test-results/`:

| Run ID | Native exit / owning classification | Cause and disposition |
| --- | --- | --- |
| `remediation-execution-input-narrow-20261014-01` | 2; aggregate `infra/preflight_error`, build child `harness/child_target_failure`; 11 cleanup steps complete | Graph interface edit accidentally referenced a transaction in the separate release method; compile error repaired. Three dependent units did not execute. |
| `remediation-execution-input-narrow-20261014-02` | 2; `product/test_assertion_failure`; 18 cleanup steps complete | New test fixture used an invalid progress-unit identity; fixture corrected, production constraint retained. Worker and Graph rows passed. |
| `remediation-execution-input-fence-20261014-01` | 2; `product/test_assertion_failure`; 14 cleanup steps complete | Test fixture attempted an incomplete cancellation-state update. Replaced it with Jobs-owned cancellation. Expired, replaced, terminal and transaction-time assertions passed. |
| `remediation-execution-input-preview-20261014-01` | 2; `product/test_assertion_failure`; 14 cleanup steps complete | Explicit server fixture omitted required route mode; restored the original disabled test-route mode. Product scenario did not execute. |

The following roots all have observed native exit 0. Final full-owner evidence
binds dirty-worktree source
`sha256:94436b76791692c1d80f1d760398a35f3eb56ab9c48228b47940631336d15080`.
No source changed during the final qualification and audit invocations.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-execution-input-format-20261014-03` | Pass, 5.016 s |
| `make generate` | `remediation-execution-input-generate-20261014-01` | Pass, 9.370 s |
| Exact execution-fencing integration row | `remediation-execution-input-fence-20261014-02` | Pass, 13.717 s |
| Exact preview integration row | `remediation-execution-input-preview-20261014-02` | Pass, 16.458 s |
| Network Flow query-authoring / production-facade unit row | `remediation-execution-input-network-boundary-20261014-01` | Pass, 3.054 s |
| `make agent-finalize` | `remediation-execution-input-finalize-20261014-01` | Pass, 18.797 s; retained maintenance skipped because `RESULTS_DIR` was unset |
| Full Reporting owner | `remediation-execution-input-reporting-owner-20261014-01` | Pass, 43.213 s; 9 fresh units; 24 cleanup steps complete |
| Full Composition owner | `remediation-execution-input-composition-owner-20261014-01` | Pass, 21.401 s; 5 fresh units; 18 cleanup steps complete |
| Full Jobs owner | `remediation-execution-input-jobs-owner-20261014-01` | Pass, 38.203 s; 6 fresh units; 21 cleanup steps complete |
| Full server owner | `remediation-execution-input-server-owner-20261014-01` | Pass, 79.745 s; 26 fresh units; 51 cleanup steps complete |
| `make lint` | `remediation-execution-input-lint-20261014-01` | Pass, 32.997 s |
| `make generate-drift` | `remediation-execution-input-drift-20261014-01` | Pass, 23.398 s |
| `make json-shape-check` | `remediation-execution-input-json-20261014-01` | Pass, 6.625 s |
| `make generated-artifact-policy-check` | `remediation-execution-input-policy-20261014-01` | Pass, 1.462 s |
| Reporting / Composition / Jobs / server evidence audits | `remediation-execution-input-reporting-audit-20261014-01`, `remediation-execution-input-composition-audit-20261014-01`, `remediation-execution-input-jobs-audit-20261014-01`, `remediation-execution-input-server-audit-20261014-01` | Pass, 2.735 / 2.744 / 2.762 / 2.772 s; explicit owner-root manifests |

The previous preview handoff passed `make lint-markdown` in
`remediation-preview-admission-markdown-20261014-01` (8.530 s, native exit 0).
These are owner-level engineering results, not cost-window or release acceptance.
The exact Graph adapter remains covered through Reporting; unrelated Network Flow
browser scenarios were not rerun for the transaction interface change. Actual
renderer/sandbox output, model fidelity, timed Jobs UI evidence, qualified cost
windows, independent review and coordinated adoption remain open. No full current
`check` or release qualification is claimed.


## Standalone Slidev runtime characterization

The renderer package now exercises the pinned Slidev engine through its public
server API and pinned Playwright Chromium. Its PDF adapter parses actual output,
checks expanded slide count and parsed PDF page count, rejects malformed output,
and normalizes PDF timestamps and document IDs from explicit admission inputs.
It never substitutes a PDF header or source digest for rendered content.
Two fresh working directories and browser contexts under the same browser process
produced byte-identical two-page PDFs for the synthetic fixture. The test also
asserts rendered headings and nonempty page-content streams. A separate fixture
exercises click expansion and rejects an incorrect expected page count.

`packages/report-renderer` pins Slidev, its default theme, Markdown-it, Mermaid,
PDF-lib and Playwright Chromium. The lockfile was updated through Make. A direct
Markdown-it 14.3.2 dependency resolves the installed Comark plugin's declared
14.x peer requirement; the initial resolver selected incompatible 15.x.
Toolchain drift checks reject mismatched or non-exact renderer runtime pins.
Mermaid is installed but has no real rendering adapter yet.

Browser interception admits only the declared loopback endpoint, blocks external
HTTP and undeclared loopback before connection, and retains bounded observation
categories without request URLs. Tests exercise the actual Playwright request
interceptor, not only a URL classifier. **These are browser-defense observations,
not OS sandbox evidence.** Native/Node network attempts, filesystem confinement,
environment confinement, deadline escalation, packaged fonts and runtime, and
production launch remain unimplemented. The Go Reporting path still uses the
previous placeholder renderer; the new package has not been wired into release or
preview execution. No Reporting adoption or production rendering claim follows
from this checkpoint.

The `renderer-node` execution context is explicitly admitted under Reporting.
Pure unit rows require frontend installation. Browser integration rows require
pinned Playwright preparation, exclusive test processes and the new
`renderer_browser` resource profile. They acquire no product services or
fixtures. Metadata rejects missing capacity, unrelated runtime profiles and
unadmitted evidence contexts. Graph tests verify the preparation dependencies and
absence of PostgreSQL/application fixture acquisition. No runtime consumer reads
repository documentation; generated temporary presentation source is product
input, not documentation evidence.

The adapter uses the pinned deck's canvas dimensions and Slidev's rounded-up
canvas height. An explicit block print layout removes the screen grid's trailing
page break without dropping PDF pages. The exact PDF page-count assertion remains
mandatory. This geometry is renderer correctness evidence, not a product timing
measurement. Pinned Slidev currently emits a caught FloatingVue/Twoslash patch
console diagnostic even with Twoslash disabled; the synthetic test records it on
failure. This does not establish that all disabled renderer features are excluded
from the production module graph. Production packaging/security review remains
open.

### Retained unsuccessful development runs

All roots below are under `.cartulary/test-results/`; native exits were observed.
No failed root was overwritten or counted as passing evidence.

| Run ID suffix after `remediation-render-runtime-` | Native exit / classification | Diagnosis and disposition |
| --- | --- | --- |
| `format-20261014-01` | 2; `config/configuration_error` before graph execution | Existing browser resource profile required PostgreSQL. Added a standalone renderer profile. Only runtime-build tool evidence was published. |
| `format-20261014-02`, `format-20261014-03` | 2; `config/configuration_error` before graph execution | New profile required schema and generated-validator refresh. No canonical run summary was published; `target-plan-json` identified the rejected enum. `make generate` repaired projections. |
| `unit-20261014-01` | 2; `product/test_assertion_failure`; 12 cleanup steps complete | PDF-lib deferred an invalid-document exception until page access. Enclosed parsing, page access and serialization in the safe error boundary. |
| `slidev-20261014-01` | 2; `infra/preflight_error`; 12 cleanup steps complete | Comark could not load Markdown-it 15's retired private export. Installed the explicitly pinned compatible peer. No render assertion executed. |
| `slidev-20261014-02`, `slidev-20261014-03` | 2; `product/test_assertion_failure`; 12 cleanup steps each | Disabling the Vite client left Slidev compile-time globals unavailable. Synthetic diagnostics identified `__SLIDEV_HASH_ROUTE__`; retained the supported local client transport and bounded its endpoint. |
| `slidev-20261014-04` through `slidev-20261014-09` | 2; `product/test_assertion_failure`; 12 cleanup steps each | Two DOM slides produced three PDF pages. Added bounded count/geometry diagnostics, matched rounded canvas geometry and corrected screen-grid pagination. No page-count tolerance or output-page truncation was introduced. |
| `finalize-20261014-01`, `finalize-20261014-03` | 2; `harness/child_target_failure`; 10 cleanup steps each | Finalizer shape validation rejected stale generated topology bindings after catalog/source changes. Regenerated through Make before finalizing. No finalizer mutation rollback was required. |

### Current owner evidence

The final owner runs and audits bind dirty-worktree source
`sha256:e624bafa95fab61b3fee178690fc04167523d4e433f6226f6164c7be668841d2`.
No authored source changed during final verification. The following roots have
observed native exit 0; durations are invocation observations, not qualified cost
windows.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make format` | `remediation-render-runtime-format-20261014-05` | Pass, 5.049 s |
| Four exact renderer unit/integration rows | `remediation-render-runtime-narrow-20261014-01` | Pass, 10.266 s; before final metadata guards |
| `make generate` | `remediation-render-runtime-generate-20261014-03` | Pass, 9.310 s |
| `make agent-finalize` | `remediation-render-runtime-finalize-20261014-04` | Pass, 18.368 s; retained maintenance skipped because `RESULTS_DIR` was unset |
| Full Reporting owner | `remediation-render-runtime-reporting-owner-20261014-01` | Pass, 30.771 s; 15 fresh units, 30 cleanup steps complete |
| Full test-catalog owner | `remediation-render-runtime-catalog-20261014-01` | Pass, 27.694 s; 2 fresh units, 11 cleanup steps complete |
| `make harness-contract` | `remediation-render-runtime-harness-20261014-01` | Pass, 45.446 s; 9 fresh units, 18 cleanup steps complete |
| `make lint` | `remediation-render-runtime-lint-20261014-02` | Pass, 21.789 s |
| `make generate-drift` | `remediation-render-runtime-drift-20261014-01` | Pass, 12.505 s |
| `make json-shape-check` | `remediation-render-runtime-json-20261014-01` | Pass, 4.533 s |
| `make generated-artifact-policy-check` | `remediation-render-runtime-policy-20261014-01` | Pass, 0.839 s |
| Reporting / test-catalog evidence audits | `remediation-render-runtime-reporting-audit-20261014-01`, `remediation-render-runtime-catalog-audit-20261014-01` | Pass, 2.516 / 2.478 s; explicit owner-root manifests |

Owner manifests are
`.cartulary/remediation-evidence/render-runtime-reporting-20261014.json` and
`.cartulary/remediation-evidence/render-runtime-catalog-20261014.json`.
The preceding execution-fencing handoff passed `make lint-markdown` in
`remediation-execution-input-markdown-20261014-01` (9.050 s, native exit 0).
Current full `check`, release acceptance, clean-source performance portfolios,
independent fidelity review and coordinated Reporting adoption remain open.

## Inert generated renderer text

Candidate Reporting now defines `slidev_markdown_escape_v2`: entity-looking text
retains its code points, ordered-list prefixes escape their separator, and the
post-Markdown text renderer encodes braces before Vue compilation. Automatic
linkification and typography are disabled. The real Slidev test demonstrates
literal expression-like text, HTML, entities and URLs without executing an
expression or creating links/scripts; existing click expansion still passes.
Go vectors cover source bytes, Unicode and line prefixes. The Go render bundle
implementation version advances to `2`; v1 escaping evidence is retired.

These changes close the exercised text boundary, not the complete deck model or
production rendering pipeline. Browser rendering remains a standalone adapter;
native sandboxing and its evidence, production integration and the other
adoption conditions remain open.

Verification binds source
`sha256:cb70ecc22b60cac1ba17c8853602d14818870bcfbf54b0ee96b6e8f54c991544`.
Roots are under `.cartulary/test-results/`. Native exits are 0 unless noted.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| `make generate` | `remediation-render-text-generate-20261014-01` | Pass, 9.350 s |
| `make agent-finalize` | `remediation-render-text-finalize-20261014-02` | Pass, 18.446 s; retained maintenance skipped because `RESULTS_DIR` was unset |
| Full Reporting owner | `remediation-render-text-reporting-owner-20261014-01` | Pass, 33.415 s; 15 fresh units; 30 cleanup steps complete |
| Full Composition owner | `remediation-render-text-composition-owner-20261014-01` | Pass, 16.468 s; 5 fresh units; 18 cleanup steps complete |
| `make lint` | `remediation-render-text-lint-20261014-03` | Pass, 22.196 s |
| `make generate-drift` | `remediation-render-text-drift-20261014-01` | Pass, 12.491 s |
| `make json-shape-check` | `remediation-render-text-json-20261014-01` | Pass, 4.596 s |
| `make generated-artifact-policy-check` | `remediation-render-text-policy-20261014-01` | Pass, 0.893 s |
| Reporting / Composition evidence audits | `remediation-render-text-reporting-audit-20261014-01`, `remediation-render-text-composition-audit-20261014-01` | Pass, 2.753 / 2.588 s; explicit manifests under `.cartulary/remediation-evidence/render-text-{reporting,composition}-20261014.json` |

The earlier `remediation-render-text-lint-20261014-01` failed with native exit 2,
`harness/child_target_failure` and 20 completed cleanup steps. TypeScript rejected
the Slidev-specific Vite configuration extension and implicit callback types.
The repair uses Slidev's exported plugin-options type, without `any`.
`remediation-render-text-lint-20261014-02` retained a passing summary (21.670 s),
secret scan and all cleanup results after restart, but its native exit could not
be recovered. It is not used as complete invocation evidence; `-03` supplies a
fresh observed success. No failed root was overwritten.

The preceding standalone-runtime handoff passed `make lint-markdown` in
`remediation-render-runtime-markdown-20261014-01` (22.280 s, native exit 0).

## Canonical source framing and template metadata cutover

`slidev_markdown_serialize_v2` removes the uncounted synthetic cover, emits complete
headmatter with quoted YAML strings, uses the first model slide's layout in that
headmatter, and emits later layout blocks separately. Titles use heading level
one. Paragraph blocks remain paragraphs; unsupported block kinds reject before
returning bytes. The existing private field-summary model is still partial and
does not establish the full Table 18-D block mapping or section derivation.

An authored synthetic JSON vector under `contracts/reporting/fixtures/` is shared
by `TestSlidevCanonicalSourceV2` and the pinned Slidev parser test. It checks exact
Go-produced bytes, metadata, two model slides without extra blank/cover slides,
layout, quoted strings and speaker notes. The parser test is source-framing
evidence; it is not a new end-to-end production PDF test. Existing real PDF,
click and inert-text characterization tests also pass.

The first-slide boundary in the previous candidate wording would create an extra
slide under the pinned parser. Reporting §18 now states the applicable boundary.
Template manifest v2 gives `author` and `theme` explicit ownership; the built-in
template is `cartulary.report.default@2`, and renderer implementation version is
`3`. Template `@1` is rejected, not aliased. Core 01 REQ-01-383 now has an explicitly
proposed companion amendment; the adopted `@1` requirement remains visible until
coordinated promotion. Candidate implementation therefore does not claim adopted
base conformance merely from these passing tests.

Migration 72 and Reporting preflight reject retained snapshots, Jobs, imported
artifacts and previews before changing history or the migration ledger. The
migration repeats the inventory under locks and requires explicit request-bound
template metadata. Supported template membership stays in Reporting's registry,
not a duplicated SQL roster. Fresh initialization and historical rejection at
schema versions 62, 68, 69, 70 and 71 passed against harness-owned disposable
databases. No user's development or production data was reset. Recovery still
requires a matching application/data set or fresh initialization; no in-place
downgrade is implemented.

| Obligation / owner | Exercising evidence | Remaining scope |
| --- | --- | --- |
| Reporting REQ-RPT-091 / 091a, source framing and text boundary | Shared byte vector; actual Slidev parser; existing real renderer text/click/PDF rows | Full block/deck derivation and production renderer remain open |
| Reporting REQ-RPT-100 / 101, explicit manifest metadata | Template registry rejection; frozen release input tampering/missing-author checks; full Reporting and Composition suites | Complete template-pack schema fidelity remains open |
| Core 01 REQ-01-383 companion | Explicit proposed `@2` amendment and this immutable-source evidence record | Coordinated adoption and accountable fidelity review remain open |
| Clean cutover / Reporting and database lifecycle | Extended `TestReportingIdentityCutoverDoesNotReconstructHistory_Integration`; full migration and server owners | Packaged deployment rehearsal remains open |

### Structural assertion replacement

The migration catalog test previously copied the current whole-catalog hash.
Adding migration 72 correctly changed that hash and caused a test failure. Its
survivor remains `TestCanonicalEmbeddedMigrationCatalogCharacterization` in the
same owner row and execution context. It still checks every embedded migration's
name and SQL digest against the generated history manifest, validates ordering
and immutable-source behavior, and now checks production hash repeatability.
An explicit fixed synthetic preimage/vector preserves byte-level hash framing;
changed runner identity, SQL bytes and filename must change the hash. No scenario
or security boundary was removed. This avoids changing a copied total hash for
every legitimate migration while retaining independent detection of protocol
and catalog-binding regressions. Accountable owner review remains part of the
final assurance handoff.

### Retained results

Final source: `sha256:afe57b34a0975542ae9de9f6104791f22ec5e841106df34a48a00a55714415c7`.
Roots below are under `.cartulary/test-results/`; listed successful invocations
have observed native exit 0. Timing is diagnostic, not qualified portfolio cost.

| Command / selection | Run ID | Result |
| --- | --- | --- |
| Exact source/model and Slidev rows | `remediation-render-source-narrow-20261014-01` | Pass, 14.847 s; before migration 72 |
| `make generate` | `remediation-render-source-generate-20261014-04` | Pass, 9.360 s |
| `make format` | `remediation-render-source-format-20261014-03` | Pass, 5.127 s |
| `make agent-finalize` | `remediation-render-source-finalize-20261014-02` | Pass, 18.642 s; retained maintenance skipped because `RESULTS_DIR` was unset |
| Full Reporting owner | `remediation-render-source-reporting-owner-20261014-02` | Pass, 47.401 s; 15 units, 5 compatible cache hits; 25 cleanup steps complete |
| Full Composition owner | `remediation-render-source-composition-owner-20261014-02` | Pass, 27.026 s; 5 fresh units; 18 cleanup steps complete |
| Full database-migrations owner | `remediation-render-source-migrations-owner-20261014-02` | Pass, 58.529 s; 9 units, 1 compatible cache hit; 21 cleanup steps complete |
| Full server owner | `remediation-render-source-server-owner-20261014-01` | Pass, 71.356 s; 26 fresh units; 51 cleanup steps complete |
| `make lint` | `remediation-render-source-lint-20261014-02` | Pass, 28.560 s |
| `make generate-drift` | `remediation-render-source-drift-20261014-01` | Pass, 16.986 s |
| `make json-shape-check` | `remediation-render-source-json-20261014-01` | Pass, 6.212 s |
| `make generated-artifact-policy-check` | `remediation-render-source-policy-20261014-01` | Pass, 1.107 s |
| Four owner evidence audits | `remediation-render-source-{reporting,composition,migrations,server}-audit-20261014-01` | Pass, 2.777 / 2.805 / 2.777 / 2.829 s; explicit manifests `.cartulary/remediation-evidence/render-source-{reporting,composition,migrations,server}-20261014.json` |

Unsuccessful roots are retained: generation `-01` rejected unsorted catalog
titles (8.390 s), and generation `-03` rejected missing migration-owner allocation
(1.680 s). Both exited 2, reported `harness/unknown_failure` in tool-summary-only
evidence, and were repaired at their authored inputs before regeneration.
`remediation-render-source-migrations-owner-20261014-01` exited 2 with
`product/test_assertion_failure` (50.631 s; all 22 cleanup steps complete). Its
sole failing row was `module.database_migrations.unit.public_surface_and_reader_capability`:
the catalog hash assertion described above. Earlier Reporting, Composition and
lint runs passed before that test-only repair; final roots supersede them for
current-source audits. Cache hits are reused evidence, not fresh stability runs.

The preceding text-safety handoff passed Markdown lint in
`remediation-render-text-markdown-20261014-01` (8.810 s, native exit 0).
Actual production rendering, OS sandboxing, full model/manifest fidelity,
remaining timed UI predicates, qualified clean-source cost windows, final review
and coordinated adoption remain open. No full current `check` or release
qualification is claimed.

## Strict logical support references

Reporting REQ-RPT-046b / Table 9-H2 now has one typed selector resolver for
support kind and target identity. Malformed or unadmitted selectors reject with
`export_model_invalid / invalid_support_ref` before any snapshot output. Empty,
ambiguous, trailing-separator, control-character, overlong, storage-shaped and
unknown-root selectors are not repaired. The former `record:unknown:<hash>`
fallback is removed. Incident field selectors identify the actual incident,
never the field name. Retained models reject malformed targets, mismatched kinds,
retired unknown-target hashes and incident support bound to a different incident.

The existing `TestReportingEvidenceLogicalSupportIdentity` entrypoint owns the
positive logical mappings, malformed-input/no-partial-output cases and retained
model tampering regressions. No test was removed and no new catalog was added.
This closes syntax and incident binding only. General target reachability,
Table 9-H source-index storage, support-ID links from retained objects and full
source/support schema fidelity remain open. Several synthetic fixtures still
reference unselected support targets; those fixtures are not reachability proof.

The interrupted narrow run `remediation-support-ref-narrow-20261014-01` has a
passing summary (14.250 s), 15 completed cleanup operations and a passing retained
secret scan; its native process exit could not be recovered after restart. It
predates the additional retained incident-binding case. The following runs have
observed native exit 0 and use source
`sha256:d6e0272ed562e31664de7abad20ba2bed3559daccd969fb63166eccdf7e2a527`.
All roots are beneath `.cartulary/test-results/`.

| Command | Exact run ID | Result |
| --- | --- | --- |
| `make agent-finalize` | `remediation-support-ref-finalize-20261015-01` | Pass, 20.850 s; maintenance skipped because `RESULTS_DIR` was unset |
| `make test-slice OWNER=module.reporting` | `remediation-support-ref-reporting-owner-20261015-01` | Pass, 36.582 s; 15 units, five compatible cache hits, 25 cleanup operations complete |
| `make test-slice OWNER=module.reportcomposition` | `remediation-support-ref-composition-owner-20261015-01` | Pass, 21.679 s; five units, 18 cleanup operations complete |
| `make lint` | `remediation-support-ref-lint-20261015-01` | Pass, 29.828 s |
| `make generate-drift` | `remediation-support-ref-drift-20261015-01` | Pass, 20.278 s |
| `make json-shape-check` | `remediation-support-ref-json-20261015-01` | Pass, 4.882 s |
| `make generated-artifact-policy-check` | `remediation-support-ref-policy-20261015-01` | Pass, 0.917 s |
| Reporting owner evidence audit | `remediation-support-ref-reporting-audit-20261015-01` | Pass, 2.612 s |
| Composition owner evidence audit | `remediation-support-ref-composition-audit-20261015-01` | Pass, 2.607 s |

Audit manifests are `.cartulary/remediation-evidence/support-ref-reporting-20261015.json`
and `support-ref-composition-20261015.json`. The broader model, production renderer,
sandbox, timed UI, qualified cost-window and coordinated adoption obligations
remain open. These runs are engineering evidence on dirty sources, not a clean
performance portfolio or release qualification.

## Exact token-use and frozen redaction binding

Reporting REQ-RPT-069 now states the cross-artifact and frozen-input integrity
obligations explicitly. Completion and sensitive-state validation bind every
private token/path use to its exact selected-rule entry, action and disclosure
partitions in the retained redaction manifest. Token inventory must cover all and
only tokenized paths; aggregate actions, classes, source families and partitions
must equal their contributing entries. Another path's action cannot authorize a
substitution. Native sensitive recovery reads the actual manifest and token file
bytes through their declared descriptors and validates hashes and cardinality.

Release/preview completion and their native recovery checks also reuse pure
Reporting redaction against the exact frozen admission model and profile. Exact
manifest, token and private-map bytes must match that deterministic result.
This replaces partial output-field comparisons with the owning computation;
it does not invoke Slidev/Mermaid, query live case state, select current profiles
or repair artifacts. A forged original value with a recomputed value digest and
private-map digest is therefore rejected even when its artifact relationships
remain internally consistent. Verification adds deterministic redaction work at
completion/recovery; its production cost is not yet qualified.

The existing `TestRenderBundleSeparatesSensitiveRedactionArtifacts` entrypoint
now exercises two selected rules and both mask/stub on one subject. Its negative
cases cover changed rule/action/disclosure, missing/extra token use, profile/scope,
class/family aggregation and altered selected-rule traces with recomputed
artifact digests. `TestReportingFrozenReleasePayload` checks rejection of invented
private values against frozen admission. Existing real release and preview
integration entrypoints exercise altered persisted rules/actions/disclosure and
rehashed original values, including restored valid state. The complete Reporting
suite includes its encrypted recovery rehearsal. No test identity was removed;
existing catalog contexts and owner routing remain in use.

This closes the exercised artifact-integrity mechanisms. It does not close the
Table 13-B/D/E schema discrepancy: the current subject/token projections still
use private shapes, and the normative one-token-per-subject identity versus
singular rule/action entry representation needs an explicit coordinated shape
and migration decision. No per-rule token-manifest conformance is claimed.
Source/support reachability, full template/deck derivation, actual production
rendering, OS sandboxing, remaining timed UI workloads, qualified cost windows,
independent fidelity review and coordinated adoption remain open.

All following commands completed with observed native exit 0. Roots are under
`.cartulary/test-results/`. Final source:
`sha256:2d9845c4cafa2a994315e3c0c4bb6cc0bc97bb9f569dc5312651876505a15569`.

| Command / selection | Exact run ID | Result |
| --- | --- | --- |
| Narrow source-contract row | `remediation-token-binding-narrow-20261015-01` | Pass, 2.557 s, before frozen-input replay addition |
| Source-contract and frozen-release-payload rows | `remediation-token-binding-narrow-20261015-02` | Pass, 3.029 s |
| `make format` | `remediation-token-binding-format-20261015-03` | Pass, 5.080 s; earlier format 01/02 also passed |
| `make agent-finalize` | `remediation-token-binding-finalize-20261015-01` | Pass, 18.869 s; maintenance skipped because `RESULTS_DIR` was unset |
| `make test-slice OWNER=module.reporting` | `remediation-token-binding-reporting-owner-20261015-01` | Pass, 44.089 s; 15 units, five compatible cache hits, 25 cleanup operations complete |
| `make test-slice OWNER=module.reportcomposition` | `remediation-token-binding-composition-owner-20261015-01` | Pass, 28.977 s; five fresh units, 18 cleanup operations complete |
| `make test-slice OWNER=app.server` | `remediation-token-binding-server-owner-20261015-01` | Pass, 54.240 s; 26 units, 51 cleanup operations complete |
| `make lint` | `remediation-token-binding-lint-20261015-01` | Pass, 33.792 s |
| `make generate-drift` | `remediation-token-binding-drift-20261015-01` | Pass, 15.343 s |
| `make json-shape-check` | `remediation-token-binding-json-20261015-01` | Pass, 6.638 s |
| `make generated-artifact-policy-check` | `remediation-token-binding-policy-20261015-01` | Pass, 1.650 s |
| Reporting owner evidence audit | `remediation-token-binding-reporting-audit-20261015-01` | Pass, 2.885 s |
| Composition owner evidence audit | `remediation-token-binding-composition-audit-20261015-01` | Pass, 2.881 s |
| Server owner evidence audit | `remediation-token-binding-server-audit-20261015-01` | Pass; exact summary retained |

Explicit owner manifests are
`.cartulary/remediation-evidence/token-binding-{reporting,composition,server}-20261015.json`.
These results do not establish a clean-source performance or release pass.

## Frozen support target closure

Reporting REQ-RPT-046c and Table 10-B now define support identities over the
snapshot, target kind and exact logical target. Source-owner selectors stay at
the provider boundary. Every retained support-bearing object carries indexed
identities; equivalent selectors coalesce without changing caller-owned input.
The former unregistered 24-hex selector hashes and retained path references are
rejected. The full schema-domain digest has an independent fixed byte vector,
and release/preview binding preserves the snapshot's support identity.

Snapshot and render admission, redaction, replay and imported-model validation
check the same frozen inventory. Missing targets, ambiguous targets, incorrect
families, changed identity bindings, duplicate or orphaned entries, and dangling
references in sections, nested blocks, records, relationships and Timeline events
fail closed. These checks never query live records or invent missing historical
provenance. The existing clean preproduction initialization remains required;
there is no legacy reader, in-place history repair or new reset operation.

This replaces incidental selector-spelling expectations in snapshot assertions
with indexed target/identity assertions. Their named surviving tests are
`TestSnapshotRecordMaterialization`,
`TestReportingEvidenceLogicalSupportIdentity`, the existing full-family snapshot
integration assertion `requireExportModelCoverage`, and
`TestReportingEvidenceProviderRedaction_Integration`, in their unchanged owner
rows and execution contexts. The canonical round-trip fixture now references its
selected Timeline event; the isolated classification and copy tests likewise
reference admitted incident/Timeline targets. No test or assurance context was
deleted. Reviewer: Codex under the authorized remediation scope.

The first focused run, `remediation-support-closure-focused-20261015-01`, failed
with outer exit 2, `product / test_assertion_failure`, in 14.221 s; all 15 cleanup
steps completed. One edited fixture accidentally replaced its disclosure list;
that was corrected. The integration path also exposed a relationship constructor
serializing an absent support list as `null`; it now emits the required empty
array. Source was
`sha256:d00ff0ce593595bbe1ed2624cf7d4968cfd0631bdf5695e4465e3739abe743b5`.
The failed root remains retained and is not classified as an intermittent outcome.

All following runs completed with observed native exit 0, under
`.cartulary/test-results/`. Final executable source:
`sha256:8308cc4507fe0349261a46fad153c36a447f73c60c1147aa4dc0199c1c04d79b`.
IDs have prefix `remediation-support-closure-`, suffix `-20261015-01`, except
the focused and final format runs use suffix `-20261015-02`.

| Run middle | Command / scope | Result |
| --- | --- | --- |
| `format` | `make format` | Pass, 5.078 s; earlier format 01 also passed |
| `generate` | `make generate` | Pass, 9.330 s |
| `focused` | Five exact Reporting rows listed below | Pass, 13.919 s; four fresh units, all 15 cleanup steps completed |
| `finalize` | `make agent-finalize` | Pass, 19.070 s; retained-run maintenance skipped because `RESULTS_DIR` was unset |
| `reporting-owner` | `make test-slice OWNER=module.reporting` | Pass, 43.074 s; 15 units, five compatible cache hits, all 25 cleanup steps completed |
| `composition-owner` | `make test-slice OWNER=module.reportcomposition` | Pass, 25.126 s; five fresh units, all 18 cleanup steps completed |
| `server-owner` | `make test-slice OWNER=app.server` | Pass, 53.914 s; 26 fresh units, all 51 cleanup steps completed |
| `lint` | `make lint` | Pass, 31.327 s |
| `drift` | `make generate-drift` | Pass, 21.123 s |
| `json` | `make json-shape-check` | Pass, 4.745 s |
| `policy` | `make generated-artifact-policy-check` | Pass, 0.915 s |
| `reporting-audit` | Reporting owner evidence audit | Pass, 3.322 s |
| `composition-audit` | Composition owner evidence audit | Pass, 3.683 s |
| `server-audit` | Server owner evidence audit | Pass, 2.590 s |

Focused rows: `module.reporting.integration.evidence_provider_redaction_4b6d27e1e9`,
`module.reporting.unit.snapshot_record_materialization`,
`module.reporting.unit.snapshot_reporting_recipient_redaction_profiles_75c820718c`,
`module.reporting.unit.frozen_release_payload`, and
`module.reporting.unit.admitted_render_export_participant_c0da40986c`.
Audit commands use `make test-evidence-audit OWNER=<owner>` and explicit manifests
`.cartulary/remediation-evidence/support-closure-{reporting,composition,server}-20261015.json`.

Target closure and indexed support links are demonstrated. Table 9-H source-index
storage, required source-backed `source_ref_id`, disclosure-label propagation,
and complete source/subject/token schema fidelity remain open. This checkpoint
does not close production rendering, OS sandboxing, remaining timed job workloads,
qualified cost windows, independent fidelity review or coordinated adoption.


## Renderer process supervision and live cancellation

The production Reporting worker now observes its exact Jobs execution while a
renderer is active. Cancellation requests, lost execution ownership, observation
failures and parent-context cancellation cancel the renderer and drain it before
returning. A renderer that returns successful bytes after cancellation cannot
contribute those bytes to finalization. Renderer panics remain safe handler
failures; raw panic values do not become validation details. Both release and
preview paths use this supervision. The existing post-render execution check
and transactional finalization guards remain necessary.

`internal/platform/renderprocess` now provides a Linux process supervisor with
explicit user, PID, network, IPC, UTS and cgroup namespaces, dropped capabilities,
disabled nested user namespaces, a read-only runtime root, private writable work
and diagnostic directories, and an explicit environment. It captures a bounded
protocol, withholds child stderr, terminates the owned group on cancellation and
waits up to two seconds before killing it. Output admission uses `os.Root`, rejects
symlinks, hard links, nonregular files, missing outputs, traversal and oversized
files, and returns no artifacts when cleanup fails. The runtime installation
capabilities are trusted assembly inputs, not template or release-request fields.

The native dependency is pinned as Bubblewrap 0.11.1 in the pin registry; `make
doctor` reports version availability. Actual namespace capability is exercised
by the Reporting owner tests. No unconfined fallback or skip was added. The
production distribution still needs to supply this dependency and its verified
renderer runtime; a development-machine installation is not deployment evidence.

The Node stage adapter has a closed, data-only request identity
`cartulary.reporting_render_stage_request.v1` and a result identity
`cartulary.reporting_render_stage_result.v1`. It rejects duplicate keys, invalid
UTF-8, malformed identity/time/page bindings, and caller-supplied runtime
capabilities. Its real Slidev test checks PDF parsing, page count, immutable
admission metadata, artifact bytes and digest, and actual browser observations.
That test runs the adapter directly. It does not yet demonstrate a packaged
Slidev render inside the Go process supervisor.

### Scope and remaining integration

This checkpoint establishes process confinement primitives, a tested renderer
stage interface, and production worker cancellation. It does **not** establish
production renderer/sandbox conformance. In particular:

- `defaultCandidateRenderer` still reaches the in-process placeholder PDF and
  Mermaid implementation. The new process/stage components are not yet wired
  into that rendering path.
- A pinned, distributable Node/Chromium/package/font runtime, measured complete
  toolchain snapshot, and production assembly are still required.
- Network namespaces block non-loopback egress and isolate host loopback, but
  do not alone enforce the exact same-attempt port policy or account for all
  blocked Node/native network attempts. Browser interception is not that proof.
- Complete staged deadlines, actual Mermaid and passive SVG qualification,
  clean-directory external rerendering, and atomic artifact publication remain
  required integration work. The worker's existing overall timeout has not
  been presented as the specified stage-timeout implementation.
- Source/model/template/deck/token fidelity, remaining timed Jobs UI workloads,
  qualified clean-source cost windows, independent fidelity review and
  coordinated adoption remain open. Reporting 1.4.0 and Composition 1.3.0 remain
  proposed. No deployment reset or adoption transition was performed.

### Retained failures and repairs

All names below are under `.cartulary/test-results/` and start with
`remediation-render-process-`. Unsuccessful roots remain retained.

| Runs | Outcome and disposition |
| --- | --- |
| `focused-20261015-01` through `focused-20261015-04` | Outer exit 2; product/test_assertion_failure during the new native test's setup. Bubblewrap required explicit `--unshare-user` before `--disable-userns`; an `--unshare-all` argument was insufficient. The launcher now names the required namespaces explicitly. These were deterministic setup failures, not instability evidence. |
| `format-20261015-02` and `format-20261015-03` | Outer exit 2; formatter rejected unnecessary control-character ranges in the JSON tokenizer. JSON syntax is already checked before tokenization, so the redundant ranges were removed. Formatter diagnostics now suppress informational messages so they cannot displace blocking diagnostics. |
| `lint-20261015-01` | Outer exit 2; the PDF metadata test needed to handle the parser's optional creation-date type. The assertion still requires the exact admission date. |
| `security-20261015-01` | Outer exit 2; targeted Gosec reported G304 for opening an output via a variable host path. Output access was structurally confined with `os.Root`; no suppression was added. |

The temporary synthetic-only stderr diagnostic used to identify the Bubblewrap
setup error was removed. The production supervisor exposes only a fixed safe
classification and exit status, never child stderr.

### Verification before final output-root hardening

The following successful runs bind source
`sha256:5fc9b62b5342751526f7ab72322d543c9287ede61ff029812068b222df2b0f56`:

| Run suffix | Command | Result |
| --- | --- | --- |
| `owner-20261015-01` | `make test-slice OWNER=module.reporting` | 18/18 units; no cache hits; 46.796 s |
| `composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | 5/5 units; no cache hits; 20.400 s; 18 cleanup steps completed |
| `server-20261015-01` | `make test-slice OWNER=app.server` | 26/26 units; no cache hits; 65.037 s; 51 cleanup steps completed |
| `lint-20261015-02` | `make lint` | Pass; 33.519 s |
| `drift-20261015-01` | `make generate-drift` | Pass; 15.049 s |
| `json-20261015-01` | `make json-shape-check` | Pass; 10.697 s |
| `policy-20261015-01` | `make generated-artifact-policy-check` | Pass; 0.960 s |

The focused `output-root-20261015-01` run passed after the `os.Root` repair
(0.594 s, one fresh unit, ten cleanup steps completed), on source
`sha256:66981f2b6c4ab2ada3f50f55aa678059b833a7a7a8de287e783a87637dc05329`.
All `agent-finalize` invocations in this slice left `RESULTS_DIR` unset; retained-run
maintenance was therefore skipped. These results are dirty-source engineering
evidence, not a clean-source performance window or Core 05 publication evidence.


## 2026-10-15 continuation: real Mermaid stages and native network mediation

This checkpoint supersedes the earlier stage-v1 and missing-Mermaid statements
above. It does not adopt Reporting or Composition and does not claim that the
production candidate-rendering path is connected to the new runtime.

### Implemented boundaries

- Real pinned Mermaid rendering now produces passive SVG geometry, labels and
  arrowheads. The adapter constructs a new SVG tree, materializes permitted
  paint/text/dash/marker geometry and rejects active content or unsupported
  effects. Reporting independently parses and validates the resulting SVG.
- Mermaid source serialization is `mermaid_source_serialize_v2`. Shared authored
  Go/Node vectors cover quotes, backslashes, entity-like text and Unicode.
  Sequence aliases are unquoted; reserved characters use a closed, one-pass
  decimal-entity encoding. The old backslash-quote spelling was rejected by the
  actual pinned renderer. No compatibility parser was retained.
- The authored `mermaid@12.2.0` pnpm patch decodes only that closed entity set into
  inert text after tokenization and before SVG text measurement. It corrects
  the pinned engine's plain-text label path; it does not interpret HTML. The
  exact patch is lockfile-bound. Remove it only after an upstream version passes
  the shared literal/geometry/determinism vectors through both owners.
- The private stage protocol is v2. It binds operation, source digest, immutable
  identity, artifact path/media/size/digest and actual browser observations. Go
  rejects malformed, unknown, duplicate, unbound, unsafe or missing results.
  Version-1 results are rejected. This is an internal process interface, not a
  release API or a toolchain/sandbox attestation by itself.
- Renderer module routes admit only pinned entry modules and the named Mermaid
  chunk directories. Prohibited roots are rejected before filesystem access;
  links, traversal and maps are rejected. Negative tests use the operational
  restricted-root policy and an absent root, without reading human documents.
- Catalog v8 explicitly declares finite preparation targets for supported Go
  and Vitest rows. Browser-backed Go interop requires frontend/browser setup,
  browser capacity and exclusive process isolation. Shards include preparation
  dependencies in their compatibility key. Resource claims do not implicitly
  select preparation. Retired v7 input fails; generated projections came from
  Make. No additional runner or execution catalog was introduced.
- The Linux amd64 process supervisor now owns a seccomp notification listener
  transferred by a launch shim before renderer execution. Connect/bind use
  copied addresses and pidfd-duplicated socket descriptions; inspected syscalls
  are never resumed with mutable pointers. Exact supervisor-allocated loopback
  ports are permitted. UDP/raw/alternate protocols, Fast Open, io_uring and
  filter replacement cannot bypass that boundary. Nonblocking descriptor mode
  cannot be cleared behind the supervisor's check. Unsupported architecture or
  observation capabilities fail closed. No raw destination enters evidence.
- Native tests exercise actual Bubblewrap namespaces, allowed loopback traffic,
  denied TCP/UDP/undeclared loopback, inherited child filters, alternate native
  paths, descriptor-mode protection, cancellation and output cleanup. These
  tests use a small Go child; they are not packaged Chromium qualification.
- The stage receives its allocated port from the supervisor instead of probing
  for one. Runtime source data cannot select the port or executable/environment.
- The Slidev adapter applies local-only font families and a fixed 1280 by 720
  canvas/viewport before rebuilding derived HTML and defines. The actual Go
  serializer's source is now exercised through the Node process and real PDF
  renderer twice in separate work directories. Core 04 REQ-04-046b owns these
  environment controls. The deployed font set and full toolchain still require
  measurement and packaging qualification.

### Evidence retained at this checkpoint

All roots below are under `.cartulary/test-results/`. Outer exit was zero for
these successful commands; failures listed afterward remain retained.

| Root | Command/scope | Result |
| --- | --- | --- |
| `remediation-mermaid-routing-20261015-02` | `make test-slice OWNER=harness.test_catalog` | 2/2 fresh units; 41.326 s; 11 cleanup outcomes completed. |
| `remediation-mermaid-assets-20261015-01` | Exact module-assets, Mermaid and stage-boundary rows | Pass; 16.979 s. |
| `remediation-network-process-20261015-03` | Native process row | Pass; 0.709 s; initial real loopback enforcement. |
| `remediation-network-process-20261015-04` | Native process row | Pass; 0.767 s; additional bypass/descriptor/descendant cases. |
| `remediation-network-finalize-20261015-02` | `make agent-finalize` | Pass; 18.988 s. |
| `remediation-network-owner-20261015-01` | `make test-slice OWNER=module.reporting` | 22/22 units; no cache hits; 46.386 s; 37 cleanup outcomes completed. |
| `remediation-network-security-20261015-01` | `make go-gosec-targeted` | 4/4 units; 17.076 s; 13 cleanup outcomes completed. |
| `remediation-render-profile-20261015-01` | Exact stage-boundary and Slidev rows | 4/4 units; no cache hits; 27.512 s; 13 cleanup outcomes completed. |

The full Reporting and security runs bind
`sha256:89ccfd5629c8c5cbabe8c480c1225699b183b6383fec8306e74d5baab1ea396e`.
The later profile run binds
`sha256:7b2c17678844697fdd7e3d7f85779cb88e3d6de191bf29c99682386d4a5a7bfb`.
Catalog routing ran on
`sha256:0b62fa46433214865295754537285c9cae1efe2b179426a371f49b0339378f35`.
These are dirty-source engineering results. They do not combine into a current
full-owner audit, clean-source cost window or publication claim. All finalize
runs left `RESULTS_DIR` unset, so retained-run maintenance was skipped.

### Diagnosed unsuccessful runs

- The earlier Composition/server evidence audits at
  `remediation-render-process-composition-audit-20261015-01` and
  `remediation-render-process-server-audit-20261015-01` rejected with
  `retained source differs from the audited executable snapshot`. They are not
  product failures and were not relabeled as successful owner evidence.
- `remediation-mermaid-focused-20261015-01` through `-06` exposed unsupported
  dash/marker/paint/effect output. Repairs materialized exact passive geometry
  and selected the engine's classic look. `focused-20261015-07` passed.
- `remediation-mermaid-boundary-20261015-01` and `-02` exposed `em` text lengths
  rejected by the independent Go SVG contract. Numeric SVG text geometry fixed
  that mismatch; `boundary-20261015-03` passed.
- `remediation-mermaid-literals-20261015-01` through `-03` exposed the invalid old
  source escaping, literal entity output and the test's mistaken assumption
  that positioned SVG word nodes contain prose spaces. The v2 source, pinned
  correction and semantic text assertion passed in `literals-20261015-04`.
- `remediation-mermaid-lint-20261015-01` rejected a nullable SVG root type. The
  checked root is now explicitly narrowed before nested geometry functions.
- `remediation-network-process-20261015-01` used an unsupported Bubblewrap option;
  the pinned launcher already transfers ordinary inherited descriptors. `-02`
  exposed the Go convenience listener's unrelated port-zero capability probe.
  The native fixture now exercises explicit sockets and the exact admitted port.
- `remediation-network-finalize-20261015-01` rejected a literal restricted-root
  path in the negative test. The test now covers every prohibited root from the
  operational policy without opening any of them. No exemption was added.
- `remediation-network-lint-20261015-01` rejected formatting of the new negative
  test. This was a deterministic formatting failure, not a renderer failure.

No failure here establishes intermittent behavior. No timing threshold, retry
policy or skip was widened to make these tests pass.

### Required next work

The production candidate still reaches placeholder rendering. A dedicated
installed launch shim, a distributable pinned Node/Chromium/package/font runtime,
complete measured toolchain evidence, real Chromium under the native filter,
staged deadlines, clean-directory external rerendering and atomic publication
must be integrated and validated before adoption. In particular, host native
capability tests and unsandboxed real-renderer tests are separate evidence; they
must not be presented as their untested composition. Source/model/template/deck
and token fidelity, companion adoption dependencies, timed Jobs workloads,
qualified cost windows and final human fidelity review remain open as described
in the parent remediation tracker. Reporting 1.4.0 and Composition 1.3.0 remain
proposed. No reset, deployment, compatibility fallback or adoption was performed.


### Updated profile verification

On source `sha256:4f850b6e6da3e5deb0dbfefd9a4b8db718fa950a9a41a72b6f851266c6aa3f8f`,
`remediation-render-profile-owner-20261015-01` passed the full Reporting slice
(22/22 units, no cache hits, 38.488 s),
`remediation-render-profile-lint-20261015-01` passed lint (11/11 units, 28.484 s),
and `remediation-render-profile-drift-20261015-01` passed generation drift
(4/4 units, 18.987 s). Outer exits were zero. This followed successful
`remediation-render-profile-finalize-20261015-01` (18.942 s; `RESULTS_DIR` unset).
`remediation-render-profile-markdown-20261015-01` passed Markdown lint (9.090 s).

Work after this source freeze starts a dedicated `cmd/render-sandbox` application
facade and further native-filter hardening. Those subsequent edits require new
applicable evidence; the preceding results do not cover them. Runtime packaging
and production rendering integration are still in progress.

## Installed renderer continuation: packaging and native integration

The private `cmd/render-sandbox` facade now assembles only the native confinement
primitive. `make build-report-renderer` builds a content-addressed installation
containing that launcher, Bubblewrap, the pinned Node binary, frozen renderer
packages, Playwright's pinned Chromium headless shell, the native library
closure, and a closed Liberation font set. The installation receipt and complete
file manifest are independently admitted by Go. Missing, additional, altered,
linked, special, and human-documentation files reject; rendering never repairs
an installation or resolves packages from the network. Build success is not
render or adoption evidence.

The build-input enumerator now shares the executable-input boundary and excludes
human material before traversal. Its existing security test is reachable through
the command-surface owner. The public command is outside the existing measured
public-target portfolio; its presence does not silently change that roster.
Catalog v8 preparation explicitly admits `build-report-renderer` only for an
exclusive Go row with the renderer browser resource context. Generated command
recipes share their repeated graph-runtime arguments without raising the
existing generated-Make size budget.

Native startup investigation exposed and repaired distinct integration problems:

- Node's pinned libuv probes `io_uring` even with `UV_USE_IO_URING=0`. Ring
  creation now returns `ENOSYS`; submission and registration remain unavailable.
  No ring descriptor is inherited. Capability probing is not outbound traffic.
- Node restores blocking pipe flags on startup. The supervisor now permits that
  operation on duplicated non-network descriptors while preventing Internet
  sockets from clearing nonblocking mode. It never resumes an inspected syscall
  against the child's mutable descriptor table.
- Full Chromium requires a filesystem profile socket. Packaging now selects the
  pinned headless-shell implementation used by the renderer profile. The native
  denial of named Unix socket operations remains in place.
- Unsupported socket creation is denied without inventing a destination-bearing
  observation. Actual undeclared-loopback and non-loopback operations still fail
  the attempt, including descendant operations.
- Vite probes wildcard addresses even with a fixed port. Slidev now runs behind
  an explicitly bound middleware endpoint. Its authoring inspector is disabled;
  its client websocket shares the same supervisor-allocated endpoint. HTML goes
  through Vite's environment transform before the print application starts.
- Teardown closes print clients before the shared HTTP endpoint, avoiding a wait
  on upgraded websocket connections that would otherwise prevent their cleanup.

### Retained investigation evidence

All roots below are under `.cartulary/test-results/`. They are dirty-source
engineering results, not release qualification or performance observations.

`remediation-build-boundary-owner-20261015-01` passed the exact
`harness.command_surface.behavior.public_registry_parity` row, outer exit 0,
42.284 s. `remediation-build-boundary-format-20261015-01` passed, outer exit 0.
The build-boundary generation 02 summary reported pass, but its native outer
exit was lost across context recovery and is unobserved.

Renderer installation generation attempts 01–03 failed on missing command
backing metadata and generated-Make size. Attempts 04–08 passed with observed
outer exit 0 after those corrections. Build attempts 01–08 separately exposed
nested public-command identity, staging manifest permissions, an unavailable
frozen dependency, and accidental removal of an upstream production directory
named `test`. Build 07 passed, outer exit 0; build 08 correctly rejected the
incomplete installed module tree. The final copier does not classify upstream
production directories as tests by their names.

`remediation-render-installed-sandbox-20261015-01` through `-08` failed during
native startup; retained causes include the libuv probe, blocking stdio setup,
and missing deployed modules. Run 09 reported `stage_browser_launch_failed`;
its native session was no longer available after context recovery, so its outer
exit remains unobserved. Runs 10–22 have observed outer exit 2:

| Runs | Diagnosed outcome |
| --- | --- |
| 10 | Full Chromium profile socket denied. Temporary diagnostics used synthetic fixtures and were removed. |
| 11 | Build rejected an incorrect call into the pinned Playwright registry export; the selected test did not execute. |
| 12 | Headless-shell startup exposed unsupported socket creation being misclassified as outbound traffic. |
| 13–14 | Real Mermaid passed; Slidev's wildcard port probe was denied. |
| 15–16 | The authoring inspector's separate loopback endpoint was denied. |
| 17–18 | Slidev loaded without its required client environment defines. |
| 19–21 | The client websocket attempted an undeclared loopback endpoint. |
| 22 | The stage exceeded its unchanged 60-second test deadline; cleanup waited on upgraded websocket connections before closing their owners. |

Run 23 validates the teardown correction. Full current owner, harness routing,
installation-tampering, formatting, finalization and broader verification remain
required. Production candidate rendering still uses the old placeholder path;
these installed-stage results do not close that integration, toolchain snapshot
fidelity, staged product deadlines, deployment, companion review, or coordinated
adoption. Reporting 1.4.0 and Composition 1.3.0 remain proposed.

### Installed-stage validation checkpoint

Run `remediation-render-installed-sandbox-20261015-23` passed, outer exit 0,
26.814 s: actual packaged PDF and SVG stages completed under native confinement
and cleaned their attempt directories. The next exact two-row run,
`remediation-render-installed-qualification-20261015-01`, passed, outer exit 0,
38.099 s, five work units, zero cache hits. It compared bytes from independent
fresh PDF and SVG attempts, checked native traffic and per-attempt cleanup, and
exercised the native denial/cancellation/output and installation-admission tests.
This is scoped engineering evidence; it is not Core 05 publication qualification.

Installation metadata admission now enforces bounded version labels and exact
package version syntax. Packaging reads and checks the actual deployed direct
package identities against the pin registry before publishing the manifest.
Synthetic admission fixtures contain synthetic versions instead of copied pins.

Following successful format (5.108 s), generate (9.380 s), and `agent-finalize`
(18.776 s), the following current-source checks passed. All outer exits were 0.
`RESULTS_DIR` was unset for finalization, so retained-run maintenance was skipped.

| Run suffix under `remediation-render-installed-` | Command | Outcome |
| --- | --- | --- |
| `owner-20261015-01` | `make test-slice OWNER=module.reporting` | 24/24 non-service work units, zero cache hits, 46.933 s. |
| `lint-20261015-01` | `make lint` | Pass, 32.405 s. |
| `catalog-20261015-01` | `make test-slice OWNER=harness.test_catalog` | Pass, 37.976 s. |
| `command-20261015-01` | `make test-slice OWNER=harness.command_surface` | Pass, 57.977 s. |
| `drift-20261015-01` | `make generate-drift` | Pass, 21.156 s. |
| `security-20261015-01` | `make go-gosec-targeted` | Pass, 15.896 s. |

The Reporting run source is
`sha256:13caa06ac27a8d66bbd1e5c41805dfa8ef2d8f34af439dd2286a34d784a1e969`
at dirty commit `00d24aa5354284dfcf6533b8f650e9eb0291ad44`. Ordinary checks were
allowed to run concurrently; their durations are not performance observations.
The next implementation slice must inject this runtime into the production
candidate pipeline, remove fabricated rendering/toolchain/observation claims,
bind exact launch arguments and environment, and preserve the separately stated
model, deployment, timing and adoption blockers.

## Production renderer admission and packaging

This checkpoint supersedes the earlier statements that production candidate
rendering still uses placeholder bytes. Reporting 1.4.0 and Composition 1.3.0
remain proposed. It does not close the complete model, template, Composition,
resource-limit, deployment, measurement or adoption obligations.

### Implementation and authority

The application now injects an explicit Reporting renderer. Claimed startup
admits the complete installed file set before profile publication, object-store
construction or worker registration. There is no production placeholder renderer
or host-binary fallback. Missing/incompatible installations return a name-only
`invalid_deployment_config` diagnostic at
`snapshot_reporting.runtime_receipt_path`, reason `reporting_runtime_unavailable`.
The configuration projection preserves the existing `claimed` key; the proposed
configuration contract is `snapshot_reporting.configuration.v2`. Core 04
REQ-04-046b/c and Reporting REQ-RPT-099a own the installed profile and admission.

The worker's candidate path now runs the installed Slidev PDF and Mermaid SVG
stages through the native supervisor. Reporting independently admits protocol
bindings, artifact bytes and browser observations, and separately requires
native network observations. Cleanup, cancellation, timeout, unavailable runtime
and forbidden connections return no publishable stage output. Complete runtime
bytes are verified before and after each candidate attempt. Startup avoids an
immediate duplicate scan after `OpenInstallation` has already verified them.

The toolchain snapshot is built from the admitted installation, actual package
identities, browser executable, font set, environment value digests and exact
browser arguments. The driver disables automatic arguments and uses the bound
browser-profile file. Attempt observation counts no longer enter bundle content,
metadata or determinism preimages. A regression varies native and browser counts
while requiring identical approval digests.

Real Mermaid qualification exposed a serializer defect: an empty quoted edge
label was invalid source. `mermaid_source_serialize_v3` now emits the canonical
unlabelled edge form; v2 is retired without a compatibility parser. The shared
cross-language vector includes inert `/incident/url(evil)` label text and real
SVG rendering. Attribute-aware assertions permit that text while continuing to
reject active elements and reference attributes.

The outer namespace launcher now uses the installed ELF interpreter and library
closure, with host linker-cache and RPATH selection disabled. This follows the
[glibc loader interface](https://sourceware.org/glibc/manual/latest/html_node/Dynamic-Linker-Invocation.html)
and the inspected [bubblewrap 0.11.1 implementation](https://github.com/containers/bubblewrap/blob/v0.11.1/bubblewrap.c).
Actual installed PDF/SVG execution passed through this path. That host execution
does not establish compatibility with Docker's outer seccomp or LSM policy.
[Docker's documented default restrictions](https://docs.docker.com/engine/security/seccomp/)
make explicit container qualification necessary; no unconfined or privileged
deployment workaround has been introduced.

The image producer now assembles a closed build context from the three public
binaries, explicit empty owned storage directories, the Containerfile and exactly
one receipt-selected renderer installation. It does not send the checkout or
retired runtime roots to Docker. The runtime is root-owned inside the image.
Staging validates the Reporting-owned schema at build time, exact file membership
and digests, regular single-link files, canonical paths and absence of human
material before traversal. The package requires `build-report-renderer`; developer
and server/browser fixtures declare the same preparation dependency and receipt.
The reference-pack package scenario preserves the shipped Reporting configuration
instead of appending a duplicate TOML table.

### Obligation and replacement review

Reviewer: Codex under the authorized implementation plan. These are scoped
behavior and execution reviews, not coordinated specification adoption.

| Obligation | Surviving evidence and context | Review disposition |
| --- | --- | --- |
| Actual render bytes and isolated repeated attempts | `TestReportingInstalledSandboxRenders`, row `module.reporting.integration.renderer_installed_sandbox`; installed Linux runtime and fresh private directories | Actual PDF/SVG and candidate pipeline exercised; container environment still open |
| Protocol and native observations are independently required | `TestReportingRenderStageRejectsUnboundOutputs`, row `module.reporting.unit.renderer_stage_result_admission`; synthetic adversarial runtime | Rejects altered artifacts and missing/invalid native evidence; not OS qualification |
| Attempt counts cannot change approval identity | `TestRenderBundleExternalReleaseRecordsDeterminismDigest`; explicit bundle fixture | Named survivor for removed sandbox-observation hash assertions; observation validity remains independently checked |
| Inert labels survive actual SVG layout | Shared `mermaid-source.v3.json` vector, real Mermaid renderer row and `TestReportingRenderStageAdmitsRealSVG` | Replaces synthetic SVG text assertion with real renderer evidence; `TestRenderBundlePreservesInertSourceLabelText` retains source-serialization assertion |
| Missing runtime rejects before downstream publication | `TestFailClosedStartup_Unit`, row `app.server.unit.invalid_deployment_config_and_bootstrap_prefligh_a87fd921ec` | New preflight assertion; bootstrap/ownership tests receive explicit fixture runtime to reach their own failure mechanisms |
| Application transaction/auth tests do not fabricate production rendering proof | Explicit `reportingfixture.Runtime` in test assembly | Preserves transaction/auth assertions; marked synthetic and excluded from render/sandbox claims |
| Package contains only the selected installation | `test-package-release.mjs`, row `harness.release.behavior.current_owner_evidence` | Changed/missing/extra files, links, duplicate receipt members, escape paths and forbidden-root fixtures reject; retired roots excluded |

The removed `escapeSVGText` helper belonged only to the retired synthetic
renderer. Actual SVG safety remains covered by the independent Go admission and
the real renderer tests. No product-contract identity is removed with this helper.

### Retained validation and failures

All roots below are under `.cartulary/test-results/`. These are ordinary checks
on dirty sources, not qualified timing windows. Observed outer exits are recorded
separately from summaries. `RESULTS_DIR` was unset for finalization; retained-run
maintenance was skipped.

| Exact run ID | Command/scope | Result |
| --- | --- | --- |
| `remediation-prod-config-projection-20261015-02` | `make test-slice OWNER=platform.config` | Outer 0; full owner, 2/2 units, 3.050 s |
| `remediation-prod-render-owner-20261015-02` | Full Reporting owner | Outer 2; 23/24 units passed; real candidate pipeline passed; inert SVG text failed the overbroad regex, since replaced |
| `remediation-prod-render-server-20261015-02` | Full server owner | Outer 2; 26/27 units passed, including process/browser units; startup fixture failures corrected in the next exact row |
| `remediation-prod-server-admission-20261015-01` | Exact startup row above | Outer 0; new renderer denial and existing bootstrap/ownership assertions passed |
| `remediation-prod-render-owner-20261015-03` | Full Reporting owner | Outer 2; 21/24 units passed, including installed loader/candidate rendering and all renderer assertions; three groups could not acquire the common object-store fixture |
| `remediation-prod-package-contract-direct-20261015-02` | `make release-evidence-contract` | Outer 0; package-copy negative cases and release graph checks passed |
| `remediation-prod-package-contract-20261015-02` | Exact release owner row above | Outer 0; 23.133 s, fresh execution |
| `remediation-prod-render-finalize-20261015-05` | `make agent-finalize` | Outer 0; 18.567 s |
| `remediation-prod-render-lint-20261015-02` | `make lint` | Outer 0; 11/11 units, 40.296 s |
| `remediation-prod-render-drift-20261015-01` | `make generate-drift` | Outer 0; 30.097 s |
| `remediation-prod-render-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0; 1.235 s |
| `remediation-prod-render-security-20261015-01` | `make go-gosec-targeted` | Outer 0; 22.992 s |
| `remediation-prod-render-shape-20261015-01` | `make deployable-shape` | Outer 0; owning step 0.800 s, build prerequisites additional |

The finalization-05/package-contract-02/lint-02 source identity is
`sha256:d2b0fefce68c8e23ca5388ab3ec6fc8937b0ad7c121a5ff86d4478cf5ba8e969`.
The installed-renderer success inside owner-03 and startup row use
`sha256:44362843e94bd7a758d147382a791f05f13b0748a9b6304fa08ad39c2a9d9027`.
Both are dirty descendants of the original baseline; neither is a clean-source
performance window or complete release acceptance.

Earlier failed development runs remain retained. `remediation-prod-render-finalize-20261015-03`
rejected a literal forbidden-root path in a negative fixture; the fixture now
derives forbidden roots from the existing executable-input policy, without an
exemption. `remediation-prod-render-lint-20261015-01` found capitalized errors and
the unused synthetic helper, now corrected. Package-contract-01 reported
`artifact/artifact_error` with `private child capture failed`; the narrower direct
run exposed an unregistered product schema in the harness validator. The package
producer now compiles its product-owned schema at build time. The cause of the
capture wrapper's loss of the underlying diagnostic is not established by that
direct run and remains an evidence-quality issue.

The owner-03 service scope
`_shared/test-services/ffd35f7f46adab4038d83fc2/service-scope.json` reports
`infra/service_readiness_timeout`, service `object_store`, initial list phase:
241 attempts, 240 `transport_unreachable`, one `operation_timeout`, 120.001 s,
container running, no OOM, and failed readiness cleanup. This is one shared
infrastructure failure, not three independent product defects. No intermittent
test-defect attribution or failure-rate claim is made. The scoped
`remediation-prod-render-services-20261015-01` reports all five units passed in
38.567 seconds, with fresh execution and every published cleanup operation
completed. The native process handle was lost across the interruption; its outer
exit is unavailable. This later pass does not establish the cause of the earlier
object-store startup/cleanup failure or erase that incident.

### Remaining adoption work

Staged template
deadlines/resource limits, complete manifest/deck/diagram/source/token fidelity,
the remaining Composition operations, full production portability/recovery,
remaining timed criteria, qualified cost windows and coordinated companion
review remain open. In particular, the legacy default diagram still derives an
ad hoc graph from fields; successful SVG rendering does not make that model
conformant. No current full `check`, release acceptance, Core 05 publication or
coordinated adoption claim is made here.

## Packaged renderer qualification

The existing package gate now builds `tools/renderprobe` as test support, mounts
only that binary and the shared synthetic Slidev fixture into a disposable
container from the exact application image, and runs two fresh native PDF and SVG
attempts. It checks output repeatability, native loopback observations, installation
integrity and attempt cleanup. This does not replace Reporting's independent
artifact/protocol admission or establish full source/model/template fidelity.
The probe uses the deployed seccomp settings and LSM domain; it cannot select a
privileged/unconfined alternative. Its container has no application credentials.

The first run, `remediation-container-render-20261015-01`, exited 2 because the
package's old global Node/Vite prohibition rejected the required runtime before
native execution. The replacement checks the exact receipt-bound image file set
and keeps tooling outside that installation forbidden. Negative cases for extra,
missing, duplicate and retired-root entries passed in
`remediation-container-layout-contract-20261015-01` (outer 0). This replacement's
behavior/execution review is in the source reconciliation document.

The second run, `remediation-container-render-20261015-02`, exited 2 after package
startup succeeded. `standup-package-smoke/artifacts/renderer-container.json`
records `mermaid_svg:sandbox_namespace_unavailable` under Docker's built-in
seccomp policy. Image `sha256:804a13cb29f36e39a8c0f4d2c15353334e4d860766e777769cb6383e32e088c5`
passed the exact-file layout check, with runtime digest
`748f91ba1d43c7423d23282528094b5c773effa8c0aa3f67240cdceb827254a9`.
Both failed runs completed owned workspace/container/volume/network cleanup.
They are retained development failures, not successful package qualification.

The candidate package now uses the authored `application-seccomp.v1.json`, drops
all outer capabilities and enables `no-new-privileges`. Its ordinary syscall set
was reviewed against [Moby's immutable default profile](https://github.com/moby/profiles/blob/6fe7deb1b9fb7c0397a4593480d7d22b9ee8caef/seccomp/default.json).
Only native amd64 is supported. Namespace construction permits the pinned
[bubblewrap 0.11.1](https://github.com/containers/bubblewrap/blob/v0.11.1/bubblewrap.c)
all-private-namespace clone flags and user-only unshare, plus mount/pivot/unmount
and the supervisor's descriptor duplication. Kernel ownership checks still apply;
no outer administration or tracing capability is added. Alternative ABIs,
arbitrary namespace joins, BPF, keyrings, `io_uring` submission and unrelated
socket families remain denied. The inner renderer policy continues to mediate
connections and prevent replacement of its namespace/filter restrictions.
Core 04 REQ-04-046c now states this deployment boundary and required proof.

Policy contract assertions passed in
`remediation-container-policy-contract-20261015-01` (outer 0). Native qualification
additionally checks non-root execution, capability absence, no-new-privileges and
representative kernel denials before rendering. Container execution of this new
profile is still pending at this checkpoint. Unsupported LSM environments remain
unsupported; there is no unconfined fallback.

The subsequent `remediation-container-render-20261015-03` still failed namespace
setup (outer 2). Safe setup classifications were refined; the focused native
row passed in `remediation-container-diagnostics-native-20261015-01` (outer 0).
`remediation-container-render-20261015-04` then identified
`sandbox_namespace_lock_unavailable`: initial namespace creation succeeded, but
bubblewrap's sysctl lock could not open Docker's protected `/proc/sys/user`
control. Both runs completed owned cleanup. No raw subprocess stderr or host
path is exposed by the new closed reason codes.

The final candidate replaces that sysctl mechanism with the child filter's
irreversible rejection of namespace-creating clone flags, unknown high clone
flags, unshare and setns. The outer profile no longer allows unshare. Bubblewrap's
broad `/dev` setup also requested a PTY namespace; the headless runtime now binds
only null/zero/random/urandom and stream descriptor links into a private read-only
device directory. It exposes no devpts, tty, shared host device tree or shared
memory directory. Docker's protected system paths are unchanged.

`remediation-namespace-lock-native-20261015-01` (outer 0, 5/5 units, 40.657 s)
passed the new namespace adversarial cases and actual installed render pipeline.
After narrowing devices, `remediation-headless-devices-native-20261015-01`
(outer 0, 5/5 units, 39.571 s) additionally verified device absence/read-only
behavior and repeated actual PDF/SVG output. `remediation-namespace-lock-contract-20261015-01`
passed with outer 0. These host-side results still require the next exact-image
container qualification; they do not replace it.

Full lint passed in `remediation-container-probe-lint-20261015-01` (outer 0,
11/11 units, 24.359 seconds) before the explicit outer profile was introduced.
Generation and finalization are ordinary dirty-source validation, not timing
windows. `RESULTS_DIR` remains unset and retained-run maintenance is skipped.

`remediation-container-render-20261015-05` completed with native outer exit 0,
15/15 work units passed, and 365.168 seconds invocation wall time. Its exact
application image is `sha256:80cdd9dcde08681207c30966ee7a9f67d573dcf3595b36ababe592ad0af98558`.
The retained `standup-package-smoke/artifacts/renderer-container.json` records
two fresh attempts each for real SVG (6,272 bytes) and PDF (11,170 bytes), with
matching bytes, native loopback observations, verified attempt cleanup and
outer-policy denials. The container exited 0 without OOM. The qualified environment
has no AppArmor or SELinux profile; this result does not qualify other LSM policies.
Certificate replacement and owned cleanup completed. Full lint also passed in
`remediation-headless-devices-lint-20261015-01` (outer 0, 11/11 units, 35.536 s),
and targeted security scanning passed in
`remediation-headless-devices-security-20261015-01` (outer 0, 4/4 units, 16.237 s).
These are development validation results, not a qualified cost window or a
Reporting adoption decision. Application-worker, portability and recovery
qualification of actual rendered artifacts remains a separate obligation.

## Production worker artifacts and deadline continuation

The package scenario now creates its recipient through the ordinary authenticated
workbook API before snapshot capture. It admits two independent releases from the
historical snapshot after activation changes: tokenized internal review and
external release. Both use the installed native PDF renderer; external rendering
also performs the required determinism rerender. Per-release evidence checks the
retained PDF digest, exact byte count, media role and header. Portability and
source-loss recovery must preserve those exact artifacts and exclude sensitive
adjuncts.

`remediation-worker-artifacts-package-20261015-01` failed with native exit 2
because the external-release fixture omitted its required Party recipient
(`recipient_partition_profile_mismatch`). The fixture was corrected through the
public API; no production authorization rule was weakened.
`remediation-worker-artifacts-package-20261015-02` then produced both native PDFs
and passed historical pack binding and pin protection, but import failed with
`incident_bundle_import_rejected/source_family_invalid`. Its source identity is
`sha256:a3953ebb371632b830b4efe0898b157f9eb970fd00f39a5d23ac164f3e71d014`.
The retained `standup-reference-pack-smoke/artifacts/rendered-releases.json`
records 13,250-byte review and 13,385-byte external PDFs with their exact digests.
Each failed run completed all 24 published cleanup steps. The import cause is
unresolved at this checkpoint; the next scenario retains only closed family and
invariant tokens to identify the failing owner, without retaining raw errors.

Template-owned declared limits now have one typed resolver for defaults, hard
bounds, scalar units and failure identities. PDF output admission, diagram count,
slide count, blocks per slide and speaker notes consume the resolved limits.
Deck blocks are chunked deterministically after composition: identities and
ordering survive, notes occur once, and the next slide uses its final ordinal.
The full post-redaction model/resource census is still open; resolving every key
does not demonstrate enforcement at every consumer.

Native stages, redaction, materialization and determinism use owned deadline
contexts. Release timeout and participant failures retain a failed release bound
to the frozen admission, with no output. Interrupted render supervision drains
the renderer and preserves cleanup failures above timeout classification while
discarding output. Job ownership checks still govern finalization. Preview
lifecycle parity, durable failure-validation summaries and the total-deadline
publication boundary remain open.

Verification roots are under `.cartulary/test-results/`:

| Run | Command / scope | Outcome |
| --- | --- | --- |
| `remediation-render-limits-owner-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 2; 23/24 units passed. Two integration assertions exposed unchunked slides above the new block limit. All 39 cleanup steps passed. |
| `remediation-render-chunk-unit-20261015-01` | Stage-admission, external-release and snapshot/replay rows | Outer 0; 4/4 units, 30.943 s after deterministic chunking. |
| `remediation-render-deadlines-unit-20261015-01` | Stage-result admission row | Outer 0; 1/1 unit, 3.028 s, including expired-context and cleanup-precedence cases. |
| `remediation-render-lifecycle-format-20261015-01` | `make format` | Terminal summary reports pass; native session was lost, so outer exit is unavailable. |
| `remediation-render-lifecycle-unit-20261015-01` | Worker and stage-result rows | Outer 0; 1/1 unit, 3.084 s. |
| `remediation-render-lifecycle-finalize-20261015-01` | `make agent-finalize` | Outer 2; `harness/child_target_failure`: renamed catalog selector required topology regeneration. No test execution was claimed. |
| `remediation-render-lifecycle-generate-20261015-01` | `make generate` | Outer 0; refreshed authored-input projections. |
| `remediation-render-lifecycle-finalize-20261015-02` | `make agent-finalize` | Outer 0; 19.263 s. |
| `remediation-render-lifecycle-owner-20261015-01` | Full Reporting owner | Outer 0; 24/24 units, 68.446 s, including native renderer, worker lifecycle and integration cases. |

The last owner run is bound to
`sha256:2601c9853eedbf24b586c5e0f67944fa9d23b3087b1108e9a4eb4c9f61454535`.
All are dirty-source development evidence; durations are not qualified cost
windows. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
Coordinated adoption and overall nine-finding closure remain open.

## Party clock authority and validation-summary continuation

`remediation-worker-artifacts-package-20261015-03` exited 2 after both native
PDF releases and historical pack/pin assertions passed. The new bounded
diagnostic identified `source_family_id=revisions` and
`invariant_id=revisions.history_reconstruction`. All 24 cleanup steps passed.
The cause was duplicate Party source timestamps: canonical Revisions snapshots
included `parties.created_at/updated_at`, but the portable Party source contract
retains ten business/identity fields and gives Records envelope ownership.
Import generated new source clocks, making reconstructed source history differ.

The correction removes the duplicate source clocks. Party projections and
Reporting now read Records clocks; create, update and rollback no longer maintain
parallel timestamps. Party canonical source snapshots use the new
`cartulary.revisions.snapshot.party.v2` identity. Migration 73 and the application
preflight reject retained Party state or v1 history before changing schema or
the migration head. The locked migration repeats the inventory check; no path
converts historical provenance. Downgrade requires a matching application/data
backup or fresh initialization. Core 02 owns the clock/snapshot decision;
Core 01's Party portable-source clause now names the current bundle version.

The historical-artifact integration fixture now creates its Party through the
ordinary authenticated workbook API. Its previous SQL-only setup omitted the
canonical create history and therefore missed the defect. The repaired test
also checks the ten-member source snapshot before round trip. The Revisions
history comparison remains strict.

Passing render summaries now bind release (or null preview), snapshot and fixed
admission time, with closed top-level and diagnostic shapes. Publication,
portable import and recovery reject rehashed summaries with missing/substituted
identity, failure results, undeclared details or inconsistent issue state.
Renderer version 6 identifies the changed bytes. Durable failed-attempt
summaries remain separate open work.

Verification roots are under `.cartulary/test-results/`:

| Run | Scope | Outcome |
| --- | --- | --- |
| `remediation-render-summary-unit-20261015-01` | Bundle summary row | Outer 2; synthetic fixture omitted required incident status. Fixture corrected without relaxing production validation. |
| `remediation-render-summary-unit-20261015-02` | Same row, fresh execution | Outer 0; positive release/preview and adversarial rehashed-summary cases passed. |
| `remediation-party-history-reproduction-20261015-01` | Historical artifact round trip through public Party creation | Outer 2; reproduced the same reconstruction defect; all 14 cleanup steps passed. |
| `remediation-party-clock-generate-20261015-01` | Generation | Outer 2; new migration lacked its explicit owner allocation. Authored allocation corrected. |
| `remediation-party-clock-generate-20261015-02` | Generation | Outer 0. |
| `remediation-party-clock-unit-20261015-01` | Party claims/cutover row | Outer 0; 3/3 units, 34.815 s. Includes empty initialization, incompatible-history preservation and rejected downgrade. |
| `remediation-party-clock-roundtrip-20261015-01` | Historical artifact round trip | Outer 2; new diagnostic assertion omitted its SQL argument. Test call corrected. |
| `remediation-party-clock-roundtrip-20261015-02` | Historical artifact round trip | Outer 0; 3/3 units, 35.817 s. |
| `remediation-party-clock-finalize-20261015-01` | Finalization | Outer 0; generated artifacts unchanged, zero rewritten files. Maintenance skipped because `RESULTS_DIR` was unset. |

The finalizer overlapped the last round-trip invocation, but its retained summary
confirms zero source rewrites. Future mutation/finalization and dependent
verification remain sequential. These are development runs, not qualified cost
windows. Actual packaged portability and recovery still require their own result.

`remediation-worker-artifacts-package-20261015-04` subsequently passed with native
outer exit 0, 15/15 units and 468.598 s wall time. The immutable application image
is `sha256:a71b88727be8cdedfd2ee488c52df7d562a2cd33a82a852a26b3f13ddaff6535`,
built from `sha256:95c0fc9a8bc6440935f19251fe55591cd50194a84fe3d12ce8748b92f461e31b`.
Its two real PDFs (13,236 and 13,373 bytes) retained their exact digests through
destination-trusted portability and encrypted source-loss restore. The source
services and configuration were destroyed by the disposable scenario before
restore; fresh authentication, historical pins, workbook state and exact retry
passed. Cleanup reported zero remaining resource groups. The newer publication
deadline changes began only after this immutable image was built and its
scenario running; this package result does not qualify those later changes.

`remediation-party-clock-service-20261015-01` passed with outer 0, 19/19 units,
97.354 s on the same source identity, including Party browser recovery. Both
are ordinary development validation, not clean-source performance windows.

Release and preview attempts now carry the total deadline through participant
materialization and successful publication. The success finalizer retains its
independent commit-proof semantics: proven commit wins over a late deadline;
only a proven-uncommitted context error permits recording a timeout; unknown
acknowledgement remains with Jobs reconciliation. Preview failures use the owner
transaction for terminal state and graph-lease release. Cleanup failure retains
its security reason while a deadline still selects the timeout finalization
mechanism, including a concurrent cancel request.

The new `TestReportingRenderDeadlineOwnsMaterializationAndPublication` is enrolled
in the existing worker row. It exercises both render kinds across materialization
timeout and proven-absent/proven-committed/indeterminate publication outcomes.
`remediation-render-publication-unit-20261015-01` passed with outer 0, 1/1 unit,
5.577 s. Generation, format and `remediation-render-publication-finalize-20261015-01`
passed; maintenance was skipped because `RESULTS_DIR` was unset. Durable failure
summaries, cancellation outcome retention and full model/Composition fidelity
remain open; this evidence is not coordinated adoption.

## Durable abort diagnostics and recovery

Reporting now owns typed, bounded abort diagnostics and stores their canonical
Table 23-B bytes and domain-separated digest with immutable job admission.
Release failure, preview failure, cancellation and graph-lease release use the
shared owner finalization transaction. Failure diagnostics never become bundle
members or approvable output. Preview diagnostics keep a null release identity.
Migration 74 and preflight reject incompatible retained state before mutation;
no historical summary is inferred from a reason token. REQ-RPT-117c records
this durability and lifecycle contract in the proposed amendment.

Cancellation winning the terminal transaction supersedes ordinary render failure,
while partial-output cleanup defects retain their security identity. A typed
timeout joined with cleanup failure cannot hide cleanup. Shutdown and lost
execution still return control to Jobs without an unauthorized terminal write.
Unknown commit acknowledgement retains its existing reconciliation semantics.
The worker tests exercise both preview and release cancellation races, including
success publication, ordinary failure, cleanup failure and materialization.

Recovery now enumerates admitted release inputs, so a missing terminal release
cannot disappear from verification. Pending admission requires no output or
failure diagnostics. Completed and canceled outcomes require the matching retained
state. Exact terminal replay rejects missing diagnostics and missing release
rows; it does not repair them. Recovery tests reject changed digests and correctly
rehashed substitutions of identity, admission time, first failure and raw details.
They also exercise public Jobs expiry, cancellation requested before worker claim,
and canceled cleanup failure through production owner transaction adapters.

One new integration scenario exposed an old assumption: release admission allowed
only succeeded/failed Jobs tombstones after public expiry. Canceled admission is
now valid, with the separate release-outcome validator requiring `render_canceled`
or `partial_output_cleanup_failed` and no publishable bytes. Running tombstones
remain invalid. The failure and subsequent fresh validation are retained below.

| Run under `.cartulary/test-results/` | Scope / outcome |
| --- | --- |
| `remediation-render-publication-owner-20261015-01` | Preceding publication slice; outer 0, 24/24 units, 59.080 s. |
| `remediation-party-clock-ddl-20261015-01` | Preceding Party migration DDL recurrence; outer 0, 3/3 units, 35.182 s. |
| `remediation-party-clock-history-20261015-01` | Revisions integration selectors passed; catalog-guard fixture still synthesized Party v1 and failed. Outer 2, 3/4 units, 36.297 s; all 15 cleanup steps passed. |
| `remediation-party-clock-catalog-20261015-02` | Corrected fixture uses Party v2 and explicitly rejects v1; outer 0, 1/1 unit, 3.910 s. |
| `remediation-render-diagnostics-unit-20261015-01` | Initial abort diagnostics worker row; outer 0, 1/1 unit, 7.256 s. |
| `remediation-render-diagnostics-focused-20261015-02` | Worker and failed-release/replay integration; outer 0, 4/4 units, 21.919 s. |
| `remediation-render-diagnostics-preview-20261015-01` | Preview failure/expiry and rehashed-corruption scenarios; outer 0, 3/3 units, 27.354 s. |
| `remediation-render-diagnostics-focused-20261015-03` | Expanded recovery enumeration and missing terminal outcome checks; outer 0, 4/4 units, 14.372 s. |
| `remediation-render-diagnostics-preview-20261015-02` | New canceled-release expiry scenario exposed the admission status defect; outer 2, `product/test_assertion_failure`, 23.214 s; all 14 cleanup steps passed. |
| `remediation-render-diagnostics-preview-20261015-03` | Corrected status rule, queued cancellation and cleanup failure; outer 0, 3/3 units, 29.036 s. |
| `remediation-render-diagnostics-finalize-20261015-01` | Outer 0, 20.430 s. Retained-run maintenance skipped because `RESULTS_DIR` was unset. |

These dirty-source development runs are not qualified performance windows.
Table 23-B structure and durable single-failure retention do not establish
REQ-RPT-113a exhaustive evaluation. Full structured post-redaction model fidelity,
resource census, complete Composition operations and remaining product timing
criteria still block coordinated adoption. Broader owner/DDL/finalization/lint
results follow once their exact invocations finish.

Broader abort-diagnostics validation completed with native outer exits:
`remediation-render-diagnostics-owner-20261015-01` passed all 24 Reporting units
(82.371 s); `remediation-render-diagnostics-ddl-20261015-01` passed the DDL
recurrence (3/3, 31.050 s); and
`remediation-render-diagnostics-finalization-20261015-01` passed shared Jobs
cancellation/reconciliation (3/3, 19.320 s). Finalization's generated transaction
reported unchanged files, so its overlap with the last focused preview run did
not mutate source. Broader validation began after finalization completed.

`remediation-render-diagnostics-lint-20261015-01` exited 2 (10/11 units passed,
41.705 s): the Records SQL boundary had not enrolled `parties/cutover.go`.
Review confirms the preflight only checks existence of Party envelopes before
migration; it neither changes envelopes nor makes their clocks Party-owned.
The authored boundary policy now admits that exact file for reads only, like
other owner source readers. Records writes remain exclusive to Records. This is
an explicit consumer declaration, not removal of the structural assertion.

## Integrated validation after abort diagnostics

The shared failure finalizer also owns generic Reporting failure completion and
its graph-lease cleanup in one transaction. The worker no longer calls separate
Jobs failure/cancellation writes. `remediation-render-terminal-unit-20261015-01`
passed (outer 0, 1/1 unit, 5.342 s), followed by
`remediation-render-terminal-finalize-20261015-01` (outer 0, 19.309 s).
Retained-run maintenance was skipped because `RESULTS_DIR` was unset.

The earlier Records boundary correction passed in
`remediation-render-diagnostics-boundary-20261015-02` (outer 0, 3/3, 1.878 s).
`remediation-render-diagnostics-gosec-20261015-01` passed (outer 0, 4/4,
12.951 s), and `remediation-render-diagnostics-markdown-20261015-01` passed
(outer 0, 30.810 s) before these later handoff additions.

The full `make check` invocation
`remediation-render-terminal-integrated-20261015-01` completed with outer exit
2, `product/test_assertion_failure`, 785/790 units passed and 616.342 s graph
wall duration. Its source digest was
`sha256:ca020819746fe9f414d4cecada8f1ffa1c7b04b3395e79aa8e19333fb7a260d6`;
its graph digest was
`sha256:d71abb550c04de395841bad6662d598987218a1bbeb29062c4e9e122ae1a0066`.
The 11 cache hits are reused evidence, not fresh executions. Five failing units
were diagnosed individually; none was retried without a repair:

| Failed unit / exact test | Cause and repair | Fresh successor run under `.cartulary/test-results/` |
| --- | --- | --- |
| `go:2daa34425febee46:4caf5349fd95`, `TestCoordinatorGeneratedRegistry_Unit` | The inactive configuration roster omitted the new Reporting runtime receipt capability. The exact sixth policy must belong to `snapshot_reporting` and forbid `snapshot_reporting.runtime_receipt_path` when inactive. | `remediation-integrated-repairs-extensions-20261015-01`: outer 0, 4/4, 13.487 s, alongside the routing check. |
| `go:73f2332e16dad54c:5cf4eb2b7060`, `TestExtensionBehaviorRouting_Static` | The assertion still named retired family schema v7. It now requires v8, the current Harness owner contract, preserving exact boundary selectors. | Same Extensions successor. |
| `go:13dfa120e787c9cb:d29730e9bf0c`, `TestRemovedJobsSurfacesRemainAbsent_Architecture` | Repository-wide call-name matching mistook Reporting's configuration validation for a retired Jobs method. The replacement inspects Jobs-owned declarations and immutable constructor signatures. Compilation rejects calls to absent APIs; unrelated owner spelling is allowed. | `remediation-integrated-repairs-jobs-20261015-01`: outer 0, 1/1, 2.908 s. |
| `go:4cf21a365b2172f3:67fb9edbf782`, `TestPartyExportedSurfaceLock_Unit` | The public API roster omitted the authorized read-only clock-cutover preflight and its typed remediation error. The roster now names those exact declarations. | `remediation-integrated-repairs-parties-20261015-01`: outer 0, 1/1, 6.489 s. |
| `go:a74a1797b1d4dc5f:a7a44286b8e0`, `TestPartyProjectionSourceEnumerationIsDeterministicCompleteAndActiveOnly` | The fixture still wrote the retired Party clock. It now seeds Records' timestamp and asserts the projection returns that exact UTC instant. | `remediation-integrated-repairs-projections-20261015-01`: outer 0, 3/3, 15.020 s. |

The Jobs guard replacement retains the existing catalog row and fixture context.
Its detector characterization covers removed types and aliases, Manager mutation,
Runner dispatch, function aliases, and invalid constructor signatures. Harmless
parameter renaming, extraction through another owner's calls, and Store creation
pass. Behavior review: the prohibited capability is post-construction Jobs
configuration/legacy dispatch, not the spelling of unrelated APIs. Harness
review: the same exact architecture selector still executes in the unit context;
no unique behavioral or service evidence was deleted. Both reviews were performed
by Codex under the implementation instruction; they do not adopt specifications.

The broader source audit additionally confirms that Core 02 REQ-02-143a's
Host/Identity assignment projection is not yet present in executable source or
migrations. Current Host/Identity provider output has no assignment facts, and
most contributed fields have no disclosure refs. Reporting's internal profile
also still uses legacy partition labels. These are implementation gaps, not
satisfied companion promotion conditions. The next structural rendering slice
must close assignment availability, scope-derived partition eligibility, complete
post-redaction model structure and its resource census before claiming external
release fidelity. The flat emitted redaction model and incomplete Composition
operations remain adoption blockers.

## Bounded stage diagnostics and admission

The second integrated run, `remediation-integrated-repairs-check-20261015-01`,
finished with outer exit 2, 789/790 units passed, 592.014 s, and
`product/test_assertion_failure`. All 959 cleanup outcomes completed. Its
108 cache hits are reused evidence. Source:
`sha256:20abea626f3b0daae7ce8d472c073e6fdeded25e1af29307b8d539962aff89b4`;
graph: `sha256:d71abb550c04de395841bad6662d598987218a1bbeb29062c4e9e122ae1a0066`.
The remaining failure was `row:module.records.architecture.boundary`: a raw
source-token rule matched `ConfigureTelemetry(` inside a negative Jobs fixture.
It was unrelated to the Records timestamp fixture, which passed.

The boundary input is now v4, with explicit `source_text` or `go_identifier`
matching on every source-token rule; v3 has no compatibility reader. The
retired mutable-lifecycle rule checks complete Go identifiers, excluding
strings/comments while retaining executable test code. Characterization covers
real declarations/calls, whitespace and intervening comments, quoted/raw
fixtures, escaped quotes, longer Unicode identifiers, and missing/unknown modes.
Other literal checks retain their explicit source-text semantics. The Jobs
declaration guard also rejects mutable constructor aliases. Behavioral review:
the prohibited lifecycle capability remains detected; inert fixture content
does not create it. Harness review: existing boundary and Jobs selectors remain
active, and no assurance context was removed. Generated validators were rebuilt
through Make; no documentation is an executable input.

Redaction admission now rejects parameters that do not belong to the selected
action, control-bearing or invalid-Unicode replacement/stub literals, and
truncation bounds outside 1–240 Unicode scalars. Empty literals remain explicit
values, and inert markup is escaped at rendering rather than rewritten during
profile admission. Default, path, and class actions use the same validation.
Core 01 REQ-01-387 owns these closed parameter semantics.

Failed render evidence now retains a deterministic bounded issue set with its
total evaluated count. Slidev source/limits and Mermaid source validation finish
the stage before returning; no Mermaid renderer starts before all selected
diagram sources validate. Retention is independent of traversal order. Above
2,000 evaluated issues, the first 499 errors and the explicit truncation warning
remain in the 500-item record. REQ-RPT-113a now resolves its former incompatible
prefix-retention and mandatory-warning instructions. Preview/release identity,
fixed admission time, closed safe details, canonical bytes, first-failure
consistency, and replay validation apply to the complete retained set. Tests
cover 1/2/499/500/501/2000/2001/2300 issues, reversed traversal, nested collectors,
multiple defective slides/diagrams, omitted or reordered issues, and unsafe
diagnostic fields. This closes collection for these stages; complete model,
template, Composition and bundle validation remain separate work.

Commands used the repository root, `CARTULARY_OUTPUT_MODE=machine`, and each
listed fresh `CARTULARY_TEST_RUN_ID`. Roots are under `.cartulary/test-results/`.

| Run | Command / selection | Observed result |
| --- | --- | --- |
| `remediation-render-admission-generate-20261015-01` | `make generate` | Outer 0. |
| `remediation-render-admission-unit-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.snapshot_reporting_rejects_duplicate_same_preced_7daa9a6a07,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | Outer 0, 1/1, 3.334 s. |
| `remediation-render-admission-boundary-20261015-01` | `make backend-module-boundary-check` | Outer 0, 3/3, 2.531 s. |
| `remediation-render-admission-jobs-20261015-01` | `make test-slice OWNER=platform.jobs ROWS=platform.jobs.unit.architecture_guard` | Outer 0, 1/1, 2.325 s. |
| `remediation-render-issues-unit-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.renderer_stage_result_admission,module.reporting.unit.reporting_job_worker_cancels_before_work_68d5024f75` | Outer 0, 1/1, 6.226 s. Preceded the final Slidev aggregation change. |
| `remediation-render-issues-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.259 s. Retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-render-issues-owner-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 24/24, 54.904 s; no cache hits, all 39 cleanup outcomes completed. |
| `remediation-render-issues-shape-20261015-01` | `make json-shape-check` | Outer 0, 3/3, 6.454 s. |
| `remediation-render-issues-drift-20261015-01` | `make generate-drift generated-artifact-policy-check` | Drift graph passed 4/4, 13.219 s; outer 2 because the second target correctly rejected reuse of the nonempty explicit run root. Policy needs its own invocation; this was command misuse, not artifact drift. |
| `remediation-render-issues-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 3/3, 0.891 s. |
| `remediation-render-issues-package-20261015-01` | `make standup-reference-pack-smoke` | Outer 0, 15/15, 449.146 s, all 24 cleanup outcomes completed. |
| `remediation-render-issues-fixed-protocol-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.renderer_stage_result_admission` | Outer 0, 1/1, 2.865 s; rejects template overrides of the fixed diagnostic protocol. |
| `remediation-render-issues-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.405 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-render-issues-lint-20261015-01` | `make lint` | Outer 2, 10/11, 21.264 s; `harness/child_target_failure`. `lint-go` identified required formatting in `slidev_source.go`. |
| `remediation-render-issues-format-20261015-03` | `make format` | Outer 0, 5.467 s; formatted the source without behavioral changes. |
| `remediation-render-issues-finalize-20261015-03` | `make agent-finalize` | Outer 0, 19.334 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-render-issues-lint-20261015-02` | `make lint` | Outer 0, 11/11, 28.053 s. |
| `remediation-render-issues-security-20261015-01` | `make go-gosec-targeted` | Outer 0, 4/4, 12.312 s. |
| `remediation-render-issues-markdown-20261015-02` | `make lint-markdown` | Outer 0, 15.950 s. |

The complete Reporting slice bound source
`sha256:b6a30468cd7816ca47c5cafcac4e9e9e86002753e9c86e184f0ed2eb6a8fcdcd`
and graph `sha256:67312714ff527160f6bffdfa04ddbe3bf7686fe52d52d4f535c159ae375e696c`.
Source remains dirty against the planning commit; these durations do not qualify
a clean-source performance portfolio. Coordinated adoption and overall
nine-finding closure remain withheld while the preceding source/model,
Composition, product-timing, portfolio and consolidation obligations are open.

The package run binds the same source digest as the 24-unit owner run and graph
`sha256:3fdadeab1fbb0e72d04295c483b4e2a139f79d55c8d5c074fe8ad9ddcff7b9d6`.
Its actual application image is
`sha256:bf9f7f2d5c23e8fc1de51589b49306844fa2e2954e285eda6e530ee1a9040921`.
The retained renderer-container artifact records two fresh attempts each for
SVG (6,272 bytes) and PDF (11,170 bytes), identical bytes within each operation,
native loopback observation, cleanup verification and outer-policy denial.
The container exited 0 without OOM, with all capabilities dropped,
no-new-privileges and the pinned seccomp policy. AppArmor/SELinux qualification
is not inferred: both labels are absent in this environment. The worker-created
release PDFs were 13,097 and 13,227 bytes and passed the package's portability
and encrypted source-loss recovery flow.

After that package run, admission was tightened to reject
`declared_limits.validation_issues.count`. Its two distinct fixed quantities
(500 retained / 2,000 evaluated before stage-end stopping) are a diagnostic
protocol, not a template override. REQ-RPT-123 now makes that exception explicit;
the fixed-protocol row above is fresh evidence for this subsequent change.
Table 13-A2 was also reconciled with Core truncation semantics: source-dependent
truncation output is not a literal profile parameter, and the fixed marker is
additional to the source-prefix bound. The full normalized Core profile-view
projection remains part of the open model-contract work.

## Shared model validation boundary

The preceding integrated run
`remediation-render-issues-integrated-20261015-01` published a passing summary:
790/790 work units, 608.020 seconds, 11 cache hits, 605 misses and 174 bypasses.
All 1,056 retained cleanup outcomes are `completed`. Its source digest is
`sha256:2c3a36cce00dd4b9c2657b6c64d76f57ec1cee451f5d1e3bd0cd978e24d43d6a`
and graph digest is
`sha256:d71abb550c04de395841bad6662d598987218a1bbeb29062c4e9e122ae1a0066`.
The native execution handle was lost during context recovery. The outer command
exit is therefore unobserved; the summary alone is not a claim that final
command acceptance, including retained-secret scanning, completed successfully.
No source changes were made until that terminal summary was observed.

Snapshot admission, redaction, replay and recovery now share one content
validation boundary. Existing support-reference, content-class, scalar-value,
duplicate-binding, Timeline-copy and snapshot-reference checks evaluate their
complete invariant sets before returning bounded model-stage diagnostics.
Ambiguous duplicate sources and records cannot select an arbitrary value for
follow-on comparisons. A bad content class does not produce an invented summary
mismatch when the summary could not be calculated. Diagnostic ordering, total
evaluated count and bounded retention use the same stage protocol as renderer
source validation. No raw source value or unadmitted identity enters diagnostics,
and a failed model returns no partial render view.

The existing `renderer_stage_result_admission` row includes regression cases for
five independent defects, multiple record-value mismatches, reversed traversal
with conflicting duplicate sources, nested invariant collectors, and 2,100 bad
fields with a 500-entry retained summary. A renderer-boundary assertion verifies
that invalid content never reaches native rendering or publishes bundle bytes.
These are additional assertions under the existing operational owner, not a new
catalog or a claim that all normative model rules have been implemented.

| Run | Command / selection | Observed result |
| --- | --- | --- |
| `remediation-model-validation-format-20261015-01` | `make format` | Outer 0, 5.468 s. |
| `remediation-model-validation-unit-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.renderer_stage_result_admission,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | Outer 0, 1/1, 3.108 s; before the additional invariant-set cases. |
| `remediation-model-validation-format-20261015-02` | `make format` | Outer 0, 5.536 s. |
| `remediation-model-validation-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.482 s. Retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-model-validation-owner-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 2, 23/24, 64.557 s; `product/test_assertion_failure` in the support-reference row. Its multiple-defect fixtures assumed support validation always returned first. |
| `remediation-model-validation-lint-20261015-01` | `make lint` | Outer 0, 11/11, 33.197 s. |
| `remediation-model-validation-format-20261015-03` | `make format` | Outer 0, 5.596 s. |
| `remediation-model-validation-support-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.evidence_provider_redaction_4b6d27e1e9` | Outer 0, 3/3, 15.665 s. Retains support-owner rejection, requires the support diagnostic in the complete stage set, and still rejects partial redaction output. |
| `remediation-model-validation-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.817 s; `RESULTS_DIR` unset, maintenance skipped. |
| `remediation-model-validation-owner-20261015-02` | `make test-slice OWNER=module.reporting` | Outer 0, 24/24, 52.093 s; 11 cache hits, 5 misses and 8 bypasses. |
| `remediation-model-validation-markdown-20261015-01` | `make lint-markdown` | Outer 0, 9.880 s; precedes the assignment-source specification addition. |

The next model-contract work remains structural: current profile path overrides
address source pointers rather than the stable structured export-model paths;
the profile view and selected-rule traces retain private shapes; the final
redacted model is still flattened. Correcting only their schema labels would
misstate conformance. The source partition facts, full normalized profile and
template contracts, structured post-redaction model, Composition operations,
complete resource census and remaining evidence obligations stay open.
Reporting and Composition have not been promoted, and all nine findings are
not yet closed.

## Frozen assignment and Party source facts

Core 02 REQ-02-143a.2 now identifies the six Host/Identity assignment
collections, actual Links provenance, active endpoint rules, and the
equality-only `rpa_` state token. Its independent fixed hash vector is covered
by the Links source test. Assignment membership is explicitly excluded from
Reporting evidence-support references. Collection writes reject mismatched
source types, target policy and link kinds before mutation; scalar replacement
cannot silently replace a multi-valued assignment collection.

The Links test moved from a transaction fixture to its own
`postgres_dedicated` row:
`module.links.integration.party_assignment_source_facts_preserve_snapshot_963c86bbbb`.
The original isolation failure was a fixture mismatch: nested savepoints in
one rollback transaction cannot demonstrate concurrent repeatable-read
isolation. The new test checks its actual isolation level, observes a concurrent
committed source change, and proves that the frozen transaction preserves the
original version. This is an attributed test-context defect, not evidence of
intermittent product behavior. Both runs are retained.

Reporting now retains a closed `cartulary.reporting_source_facts.v1` object in
`cartulary.reporting_snapshot_model.v2`. Owner contributions use
`cartulary.reporting_owner_provider_output.v3`. Party identity and default-false
directory eligibility come from the Party owner; assignment facts come from
Links under the same borrowed transaction. Snapshot construction clones the
facts. Release and preview admission preserve them separately from the export
model, retain their schema-domain digest, and compare them to the independently
stored snapshot on replay. Native renderer requests do not receive this private
object. The current canonical serializer admits the exact UUID text type while
continuing to reject raw byte arrays and buffers.

Migration 75 rejects retained Reporting history before mutation and installs
the v2 snapshot constraint. No legacy reader or provenance reconstruction is
provided. The existing cutover test now includes migration 74 to 75 and requires
unchanged data and migration position after rejection. Current-head tests use
the current schema; historical cutover fixtures intentionally retain v1.
No user database was reset during this work.

| Run | Selection | Observed result |
| --- | --- | --- |
| `remediation-assignment-source-links-20261015-01` | Original active-links row | Terminal failure `product/test_assertion_failure`; native exit was unavailable after context recovery. Snapshot assertion used a rollback/savepoint fixture. |
| `remediation-assignment-source-generate-20261015-01` | `make generate` | Outer 0, 10.090 s. |
| `remediation-assignment-source-links-20261015-02` | Dedicated assignment-source row | Outer 0, 3/3, 20.284 s; fresh database execution. |
| `remediation-frozen-source-owner-20261015-01` | Full Reporting owner | Outer 2, 22/24, 62.843 s. Snapshot construction omitted the new source object while adding incident metadata; fixed by preserving the typed contribution. Malformed-schema diagnostics again reject the outer schema first. |
| `remediation-frozen-source-unit-20261015-01` | Materialization, provider, frozen-release, renderer-admission and redaction rows | Outer 2, 1/2, 3.406 s. UUID storage was rejected as binary by the serializer; a tokenization fixture also failed to carry its newly added Party facts. Both corrected without relaxing source admission. |
| `remediation-frozen-source-unit-20261015-02` | Same five exact rows | Outer 0, 2/2, 3.843 s; one cached unit and one fresh unit. Includes source omission, duplicate, endpoint, provenance, tampering, renderer rejection and snapshot-substitution cases. |
| `remediation-frozen-source-support-20261015-01` | Reporting evidence-provider row | Outer 0, 3/3, 14.173 s. |
| `remediation-frozen-source-finalize-20261015-01` | `make agent-finalize` | Outer 2, 4.718 s; new migration absent from authored history manifest. |
| `remediation-frozen-source-finalize-20261015-02` | `make agent-finalize` | Outer 2, 4.786 s; new migration absent from owner allocation. Both authored registries corrected. |
| `remediation-frozen-source-finalize-20261015-03` | `make agent-finalize` | Outer 0, 19.453 s. Retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-frozen-source-owner-20261015-02` | Full Reporting owner | Outer 2, 23/24, 68.149 s. The new assignment fixture exposed a false endpoint ambiguity between generic record envelopes and their domain records. |
| `remediation-frozen-source-composition-20261015-01` | Full Composition owner | Outer 0, 5/5, 35.443 s. |
| `remediation-frozen-source-lint-20261015-01` | `make lint` | Outer 0, 11/11, 37.622 s. |
| `remediation-frozen-source-provenance-20261015-01` and `-02` | Snapshot creation/provenance row | Outer 2, 2/3, 22.862 s and 22.252 s. Temporary provider probes narrowed the deterministic failure to snapshot endpoint matching; probes removed after diagnosis. |
| `remediation-frozen-source-provenance-20261015-03` | Snapshot creation/provenance and record-materialization rows | Outer 0, 4/4, 31.274 s. Domain endpoints are matched separately from generic envelopes; the regression includes both projections. Real snapshot admission retains assignment identity, provenance and Party privacy defaults, excludes assignment links from support evidence, and preserves old facts after live source mutation. |
| `remediation-frozen-source-finalize-20261015-04` | `make agent-finalize` | Outer 0, 19.769 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-frozen-source-owner-20261015-03` | Full Reporting owner | Outer 0, 24/24, 61.263 s; 11 hits, 5 misses, 8 bypasses. |
| `remediation-frozen-source-links-20261015-01` | Full Links owner | Outer 0, 21/21, 88.518 s; 0 hits, 5 misses, 16 bypasses. |
| `remediation-frozen-source-migration-20261015-01` | `make migration-drift` | Outer 0, 6/6, 19.971 s. |
| `remediation-frozen-source-markdown-20261015-01` | `make lint-markdown` | Outer 0. |

All run roots are under `.cartulary/test-results/<run-id>`. Native exits were
collected independently of summaries for these four final checks.

This closes source retention plumbing only. Partition derivation still needs to
consume these facts; full Host/Identity assignment authoring and restoration
validation, normalized profiles, structured post-redaction output, complete
Composition operations and resource census remain adoption blockers. The
existing native-renderer/package evidence predates this schema change and must
not be represented as a current package or recovery qualification.

## Explicit profile policy and normalized rule selection

Authored profiles now use `cartulary.redaction_profile.v2`, with explicit
lifecycle status, release scopes, superseded-record policy and nullable time
bucketing. Built-in profiles advance to version `2`; selectors and frozen
payload readers reject version `1`. Internal and tokenized-review profiles
permit internal draft/review; the external profile permits external release.
External admission requires `adopted_current`. This is a proposed coordinated
Core companion change, not a completed adoption decision.

The renderer selects actions through the normalized
`cartulary.reporting_redaction_profile_view.v1` interface. Separate path and
content-class arrays retain authored order; the profile default is explicit.
Non-token mask/stub defaults materialize as literals, including an explicitly
empty literal. The `selected_rule_trace.v1` object contains only rule identity,
precedence, selector kind and nullable selector value. It carries no action
parameters, source values or display tokens. Normalized views own copies of
policy values. Duplicate selectors remain invalid under Core 01; an ordering
tie never authorizes competing rules.

Profile partition declarations now use the closed normative vocabulary and
ordering. `public_summary` and `working_material` are no longer admitted as
partition names; the latter remains a valid content class. This vocabulary
correction does not establish that every model element has been assigned or
filtered correctly. Full structured model paths and frozen-source partition
derivation remain required work.

Migration 76 rejects retained Reporting history before replacing the render
profile admission constraint. It admits only explicit profile v2 policy and
checks the chosen scope and external lifecycle status. Preflight uses the new
constraint marker. The migration 75-to-76 rehearsal retains a valid v2 snapshot
and proves that rejection leaves data and migration position unchanged. No
user database was reset; no profile policy or historical digest is inferred.

| Run | Selection | Observed result |
| --- | --- | --- |
| `remediation-profile-view-format-20261015-01` | `make format` | Outer 0, 5.638 s. |
| `remediation-profile-view-unit-20261015-01` | Redaction, frozen-release and renderer-admission rows | Outer 2, 0/1, 4.189 s; tokenized fixture retained version `1`. API selector also still admitted the retired version; both corrected. |
| `remediation-profile-view-unit-20261015-02` | Same rows plus normalized-profile validation | Outer 2, 0/1, 3.889 s; new test miscounted the 14-member profile view as 15. Corrected the test expectation. |
| `remediation-profile-view-unit-20261015-03` | Same four rows | Outer 0, 1/1, 4.402 s. Includes unknown/retired shape rejection, explicit nulls and arrays, default/empty literal handling, scope/lifecycle policy, trace integrity and non-aliasing. |
| `remediation-profile-view-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.417 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-profile-view-owner-20261015-01` | Full Reporting owner | Outer 2, 22/24, 68.259 s. Existing database constraint still required profile v1. Partial unit fixtures also omitted scope or used an internal profile to exercise a later external-render error. |
| `remediation-profile-view-lint-20261015-01` | `make lint` | Outer 2, 10/11, 28.175 s; `harness/child_target_failure` in `lint-go` because two newly authored Go files required formatting. |
| `remediation-profile-view-markdown-20261015-01` | `make lint-markdown` | Outer 0; before this new handoff section. |
| `remediation-profile-view-format-20261015-02` | `make format` | Outer 0, 5.610 s. |
| `remediation-profile-view-cutover-20261015-01` | Snapshot admission/cutover, admitted participant, Job deadline/cancellation, and decoder rows | Outer 0, 4/4, 32.086 s; fresh database evidence. Migration 76 fixes the constraint, and complete fixtures again reach their intended failure mechanisms. |

All roots are `.cartulary/test-results/<run-id>`. These failures were
attributed deterministic implementation/fixture defects, not intermittent
outcomes. Broader verification of the latest migration and profile state is
still required. Reporting and Composition remain proposed. The overall remediation is not
complete until all nine findings meet their separate design, implementation
and assurance exits.

| Run | Selection | Observed result |
| --- | --- | --- |
| `remediation-profile-view-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.013 s; `RESULTS_DIR` unset, maintenance skipped. |
| `remediation-profile-view-owner-20261015-02` | Full Reporting owner | Outer 0, 24/24, 71.325 s; 11 hits, 5 misses, 8 bypasses. |
| `remediation-profile-view-composition-20261015-01` | Full Composition owner | Outer 0, 5/5, 36.942 s. |
| `remediation-profile-view-migration-20261015-01` | `make migration-drift` | Outer 0, 6/6, 21.371 s. |
| `remediation-profile-view-lint-20261015-02` | `make lint` | Outer 2, 10/11, 38.922 s; `harness/child_target_failure`. Staticcheck identified the now-unused trimming/string-set helper. Removed it; partition identifiers use exact-token ordering. |

| Run | Selection | Observed result |
| --- | --- | --- |
| `remediation-profile-view-finalize-20261015-03` | `make agent-finalize` | Outer 0, 20.142 s; `RESULTS_DIR` unset, maintenance skipped. |
| `remediation-profile-view-lint-20261015-03` | `make lint` | Outer 2, 10/11, 22.447 s; helper removal left trailing whitespace requiring `make format`. |
| `remediation-profile-view-markdown-20261015-02` | `make lint-markdown` | Outer 0, before the subsequent exact-path clarification. |
| `remediation-profile-view-format-20261015-03` | `make format` | Outer 0, 5.638 s. |
| `remediation-profile-view-finalize-20261015-04` | `make agent-finalize` | Outer 0, 19.553 s; `RESULTS_DIR` unset, maintenance skipped. |
| `remediation-profile-view-format-20261015-04` | `make format` | Outer 0, 5.968 s; includes identifier admission changes. |
| `remediation-profile-view-identifiers-20261015-01` | Profile validation and frozen-release rows | Outer 0, 1/1, 3.088 s. |

## Exact structured-model rule addresses

Profile and rule identifiers now enforce Table 8-A's exact Unicode scalar
contract, including its 128-scalar bound, UTF-8 validity, control and delimiter
exclusions, and edge whitespace rules. Values are never trimmed, normalized,
case-folded or encoded into substitute identities. Graph source registry owner
admission shares that same identifier contract.

Exact path rules now resolve against structured field-object addresses in the
admitted model. Source provenance paths remain separate. The closed parser
rejects source pointers, array indexes, wildcards and query expressions.
Snapshot-bound tokenized profiles freeze the resolved addresses before hashing;
rendering does not consult live source state. Nested block paths preserve
ancestry and escape quoted IDs. Duplicate source bindings or model addresses
fail rather than overwrite a candidate.

Independent fixed path vectors and existing behavioral redaction assertions
cover scalar and multi-attribute source contributions, nested blocks, quoted
identifiers, rule precedence, dropped token subjects, literal versus token
substitution and retained admission binding. This implements exact field rule
selection, not the entire structured output or manifest contract: distributable
output remains flattened, manifest entries still need their normative model
paths/shape, and section/block rule handling and post-redaction resource census
remain open. Frozen-source partition derivation and filtering for every scope
and model element also remain required. No adoption is claimed.

| Run | Selection | Observed result |
| --- | --- | --- |
| `remediation-model-path-format-20261015-01` | `make format` | Outer 0, 5.919 s. |
| `remediation-model-path-unit-20261015-01` | Precedence, closed profile validation, frozen release, renderer admission and exact graph selection rows | Outer 0, 1/1, 3.249 s; fresh execution. |

| Run | Selection | Observed result |
| --- | --- | --- |
| `remediation-model-path-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.530 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-model-path-owner-20261015-01` | Full Reporting owner | Outer 0, 24/24, 64.217 s; 11 hits, 5 misses, 8 bypasses. |
| `remediation-model-path-lint-20261015-01` | `make lint` | Outer 0, 11/11, 33.418 s. |
| `remediation-model-path-markdown-20261015-01` | `make lint-markdown` | Outer 0, before adding these result rows. |

## Effective disclosure scope and internal request defaults

The production package was requalified after migration 76, profile v2 and exact
field selectors. Its smoke fixture had still requested retired profile version
1; it now requests version 2. The current package run completed all 15 work
units and all 24 cleanup steps, with native exit 0. This covers its real renderer
and confinement scenario, historical review/external release generation,
Reference Pack pins, cross-installation portability and source-loss recovery.
Its source digest is
`sha256:b8bb1c5e95177c5a6784d6c961ccbdbefc3e53b5fdfad11b64555cfebc072495`;
it precedes the subsequent scope/default changes below and is not evidence
for those later changes.

Field redaction now intersects the declared profile permissions with the
effective set computed from the release scope, selected recipients and exact
frozen Party source-record identities. Internal scopes include public,
internal-only and all snapshot Party partitions. External scope includes public
and only selected snapshot Parties. No scope admits blocked content. Labels,
field partition assertions and an overbroad profile cannot invent a Party or
recipient permission. The intersection is computed once per selection pass.
Selected rule identity and trace remain unchanged; an unresolved retaining
action fails rather than silently changing to a drop or declassifying content.
Partition arrays use the normative order, and invalid retained field arrays
are rejected rather than normalized during reading.

The Core companion identifies `internal_passthrough` as the existing
`cartulary.redaction.internal@2` identity, with no selector alias. Both profile
request members may be omitted together for internal scopes. The normalized
request is identical to the explicit default; admission subsequently freezes
the complete snapshot-bound profile and digest. External requests, partial or
null tuples, and retained payloads receive no fallback. Internal passthrough
and tokenized review profiles bind all frozen Party permissions. External
recipient admission also checks the profile's scope, even when its Party set
would otherwise match.

The OpenAPI projection and authored release change set record the conditional
pair. Generation initially rejected the two unrecorded changes (405 of 407
changes covered); the authored records were added before successful generation.
The first focused run exposed the missing external profile-scope check; it was
fixed without weakening the recipient rule. These are deterministic defects,
not evidence of intermittent outcomes. Unit coverage includes all scopes and
retaining actions, overbroad profiles, unknown/duplicate recipients, snapshot
identity, exact defaults, partial/null tuples and retained-payload rejection.
The existing idempotency integration scenario now omits the internal profile
and replays the same request with the explicit pair.

These changes cover the current field-selection boundary. They do not close
source-fact partition derivation, structured post-redaction model/manifest
materialization, or section/block/record/asset filtering. Those obligations,
Composition operations, remaining product timing evidence and the qualified
cost portfolio still block coordinated adoption and overall nine-finding closure.

Commands below ran from the repository root with the exact run ID passed as
`CARTULARY_TEST_RUN_ID`. Roots are under `.cartulary/test-results/`.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-profile-package-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.994 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-profile-package-20261015-01` | `make standup-reference-pack-smoke` | Outer 0, 15/15, 469.287 s; all 24 cleanup steps completed. |
| `remediation-disclosure-scope-format-20261015-01` | `make format` | Outer 0, 6.064 s. |
| `remediation-disclosure-scope-generate-20261015-01` | `make generate` | Outer 2, compatibility change-set coverage failure; normalized `harness/unknown_failure`, 3.402 s. |
| `remediation-disclosure-scope-generate-20261015-02` | `make generate` | Outer 0, 10.370 s. |
| `remediation-disclosure-scope-unit-20261015-01` | Redaction precedence, recipient profiles, request decoding and frozen-release rows | Outer 2, 0/1, 4.935 s; `product/test_assertion_failure`, missing external profile-scope rejection; all 10 cleanup steps completed. |
| `remediation-disclosure-scope-unit-20261015-02` | Same four rows | Outer 0, 1/1, 2.935 s. |
| `remediation-disclosure-scope-format-20261015-02` | `make format` | Outer 0, 5.869 s. |
| `remediation-disclosure-scope-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.836 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-disclosure-scope-owner-20261015-01` | Full Reporting owner | Outer 2, 23/24, 72.590 s; `product/test_assertion_failure`. Private-adjunct fixture omitted its explicit internal-only permission; all 28 cleanup steps completed. |
| `remediation-disclosure-scope-shape-20261015-01` | `make json-shape-check` | Outer 0, 7.375 s. |
| `remediation-disclosure-scope-drift-20261015-01` | `make generate-drift` | Outer 0, 29.197 s. |
| `remediation-disclosure-scope-lint-20261015-01` | `make lint` | Outer 0, 40.874 s. |
| `remediation-disclosure-scope-format-20261015-03` | `make format` | Outer 0, 5.823 s; explicit permission added to the sensitive-artifact fixture. |
| `remediation-disclosure-scope-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.487 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-disclosure-scope-owner-20261015-02` | Full Reporting owner | Outer 0, 24/24, 54.475 s; includes default-profile admission and explicit-pair idempotent replay. |
| `remediation-disclosure-scope-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.865 s. |
| `remediation-disclosure-scope-markdown-20261015-01` | `make lint-markdown` | Outer 0, 31.120 s; before these final result rows. |

### Partition-order validation across retained copies

The structured-model review found that generic lexical reference ordering still
appeared in record finalization, scalar-field materialization, retained field/
record validation and private token-use union validation. This conflicts with
the normative order public, Party partitions, internal-only, blocked. All those
Reporting canonical boundaries now use the same vocabulary-aware ordering and
validation. Source/support references retain their independent lexical contract.
The provider port's intermediate record list remains a source transport
projection; Reporting canonical materialization does not reuse its ordering.

Mixed-partition vectors exercise scalar projection, source/record consistency,
token manifests and sensitive-map validation. Frozen nil, duplicate, unknown,
nested and lexically sorted-but-normatively-wrong arrays reject. An additional
scope case proves that a broader effective set cannot override a profile's
narrower permission set. The first two runs exposed the remaining lexical sorts;
the third passes after correcting the canonical materializer. These failures
are retained as deterministic regression evidence.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-partition-order-format-20261015-01` | `make format` | Outer 0, 5.699 s. |
| `remediation-partition-order-format-20261015-02` | `make format` | Outer 0, 5.868 s. |
| `remediation-partition-order-unit-20261015-01` | Snapshot records, renderer artifact-boundary and recipient-profile rows | Outer 2, 0/1, 3.902 s; `product/test_assertion_failure`; both mixed-partition paths rejected; all 10 cleanup steps completed. |
| `remediation-partition-order-unit-20261015-02` | Same rows after record-sort correction | Outer 2, 0/1, 2.585 s; same classification; scalar materializer still sorted lexically; all 10 cleanup steps completed. |
| `remediation-partition-order-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.663 s; before the scalar materializer correction; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-partition-order-unit-20261015-03` | Same three rows | Outer 0, 1/1, 2.821 s. |
| `remediation-partition-order-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.497 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-partition-order-owner-20261015-01` | Full Reporting owner | Outer 0, 24/24, 71.290 s. |
| `remediation-partition-order-lint-20261015-01` | `make lint` | Outer 0, 32.042 s. |
| `remediation-partition-order-markdown-20261015-01` | `make lint-markdown` | Outer 0, 35.480 s; before these result rows. |

### Frozen subject assignments in release redaction

The production redaction entrypoint now requires the independently retained
SourceFacts input. Active `subject_party` assignments determine Host and
Identity field partitions; responsible-party and observer-party assignments
grant no disclosure permission. Unassigned subjects receive `internal_only`
for internal scopes and `blocked` for external release. Direct Party fields
remain Party-partitioned. A public-directory flag alone cannot authorize public
labels, and existing restrictions survive assignment. Derivation copies its
input; it neither mutates the snapshot nor consults current source state.

The snapshot-record scenario exercises these rules and the production redaction
entrypoint, including missing assignments, multiple required Parties, duplicate
active assignments and existing blocked content. The provenance integration
fixture now assigns both its Host and Identity through real source links. It
then changes the live Host assignment to an observer relation after snapshot
admission, testing that release rendering uses frozen permissions.

This slice covers direct subject fields and generic envelopes bound to those
subjects by exact frozen source ID. Indirect subject references, free-text
derivation and complete structured output remain
separate open obligations; this evidence does not authorize adoption.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-subject-partitions-format-20261015-01` | `make format` | Published pass, 2/2, 5.700 s; 11 cleanup steps completed. Native session exit was lost across continuation and is not asserted. |
| `remediation-subject-partitions-unit-20261015-01` | Snapshot records, recipient-profile and deterministic-redaction rows | Outer 0, 4.385 s; before adding the production-entrypoint regression. |
| `remediation-subject-partitions-format-20261015-02` | `make format` | Outer 0, 5.883 s. |
| `remediation-subject-partitions-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.021 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-subject-partitions-owner-20261015-01` | Full Reporting owner | Outer 2, 23/24, 66.148 s; `product/test_assertion_failure`. Publication fixture lacked required subject assignments; all 28 cleanup steps completed. |
| `remediation-subject-partitions-format-20261015-03` | `make format` | Outer 0, 5.918 s. |
| `remediation-subject-partitions-unit-20261015-02` | Snapshot-record and external publication rows | Outer 0, 14.506 s; explicit source assignments added to the positive publication fixture, with generic-envelope disclosure checks. |
| `remediation-subject-partitions-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.552 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-subject-partitions-owner-20261015-02` | Full Reporting owner | Outer 0, 24/24, 50.209 s. |

### Canonical subject identity and display digest

`TokenizableSubject` now projects Table 13-B's ten-member closed object. The
source-record and unresolved-mention arms are explicit and mutually exclusive;
the former `subject_kind`, embedded source-reference and raw `display_name`
members are rejected. A subject carries a canonical display-value digest, and
snapshot/replay validation checks that digest and every reference/partition
against the independently frozen record. Duplicate records never select an
arbitrary display value for validation. Token selection resolves the identity
union against frozen records rather than parsing display labels or source paths.

The specification now pins the display preimage's schema domain and explicit
null for unavailable source labels. That clarification does not authorize
inventing a label or evidentiary support, nor does it close later token/reveal
stages whose contract requires an available display value. The full token and
reveal schemas, post-redaction structured model and final bundle digest remain
open. Ordinary snapshots may retain raw source record content; this slice
removes the second raw subject copy, not the admitted private source itself.

Migration 77 rejects incompatible retained history before mutation. Its marker
is checked by preflight, and its snapshot constraint rejects the retired
subject shape. Migration ownership/history inputs are authored; projections
are regenerated through Make. No historical reader, digest backfill or actual
user-environment reset was introduced. The migration rehearsal now includes
the state immediately after migration 76.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-subject-model-format-20261015-01` | `make format` | Outer 0, 5.672 s. |
| `remediation-subject-model-generate-20261015-01` | `make generate` | Outer 2, 1.722 s; `harness/unknown_failure`, missing migration source-owner allocation. Added the Reporting allocation before retrying. |
| `remediation-subject-model-format-20261015-02` | `make format` | Outer 0, 5.781 s. |
| `remediation-subject-model-generate-20261015-02` | `make generate` | Outer 0, 10.110 s. |
| `remediation-subject-model-unit-20261015-01` | Snapshot-record, recipient-profile and frozen-release rows | Outer 0, 3.516 s; includes fixed canonical digest vector, identity/null/omission/tampering cases and recovered-model validation. |
| `remediation-subject-model-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.062 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-subject-model-owner-20261015-01` | Full Reporting owner | Outer 0, 24/24, 67.255 s. |
| `remediation-subject-model-lint-20261015-01` | `make lint` | Outer 0, 35.252 s. |
| `remediation-subject-model-migrations-20261015-01` | `make migration-drift` | Outer 0, 26.715 s. |
| `remediation-subject-model-composition-20261015-01` | Full Composition owner | Outer 2, 4/5, 11.637 s; `product/test_assertion_failure`; all 18 cleanup steps completed. Preview fixture computed an internal profile digest against an empty snapshot despite an admitted Party. |
| `remediation-subject-model-drift-20261015-01` | `make generate-drift` | Outer 0, 14.507 s. |
| `remediation-subject-model-shape-20261015-01` | `make json-shape-check` | Outer 0, 5.387 s. |
| `remediation-subject-model-markdown-20261015-01` | `make lint-markdown` | Outer 0, 26.510 s; before these result rows. |
| `remediation-subject-model-format-20261015-03` | `make format` | Outer 0, 5.832 s; preview fixture now binds both profiles to the exact retained snapshot. |
| `remediation-subject-model-composition-20261015-02` | Exact preview/delegation row | Outer 0, 22.905 s. |
| `remediation-subject-model-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.949 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-subject-model-composition-20261015-03` | Full Composition owner | Outer 0, 5/5, 23.035 s. |
| `remediation-subject-model-portability-20261015-01` | Exact Incident Bundles reference-catalog/native-snapshot round-trip row | Outer 0, 35.998 s. |

### Model addresses and rendered labels

Field outputs, redaction entries and sensitive-map uses now serialize
`export_model_path` with the exact keyed address in the admitted pre-removal
model. Token-use arrays carry those same addresses. Source selectors remain
private inputs for subject resolution; they no longer serve as artifact
addresses. Final entries sort by model path, and redaction/adjunct validation
rejects non-model addresses. The source/profile remain independently bound by
retained admission and deterministic recomputation.

Slide text now contains the redacted value rather than a source-selector
prefix. Default diagram labels likewise use the redacted scalar/array text
conversion, so a token appears as its display token rather than the source
record path. Invalid values still fail before bundle publication. The existing
inert-label test previously asserted that a source path appeared as a diagram
label; it now asserts the actual inert value survives and source selectors do
not become visible text. No rendering or sandbox assertion was deleted.

These are existing REQ-RPT-062c/069a address and case-content boundary repairs.
The complete Table 13-F member set, section/block entries, structured
post-redaction output, and exact template-driven diagram derivation remain open.
The current flat redacted model is not presented as canonical Table 9-A closure.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-artifact-paths-format-20261015-01` | `make format` | Outer 0, 5.869 s. |
| `remediation-artifact-paths-unit-20261015-01` | Four Reporting redaction/subject/artifact rows | Outer 2, 0/1, 3.055 s; `product/test_assertion_failure`; the old inert-label assertion expected a source selector to be rendered. All 10 cleanup steps completed. |
| `remediation-artifact-paths-format-20261015-02` | `make format` | Outer 0, 5.836 s. |
| `remediation-artifact-paths-unit-20261015-02` | Same four rows with corrected value-preservation assertion | Outer 0, 3.115 s. |
| `remediation-artifact-paths-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.091 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-artifact-paths-lint-20261015-01` | `make lint` | Outer 0, 33.360 s. |
| `remediation-artifact-paths-owner-20261015-01` | Full Reporting owner | Outer 2, 21/24, 59.274 s; required template bindings incorrectly compared private source selectors with canonical output addresses. All 28 cleanup steps completed. |
| `remediation-artifact-paths-format-20261015-03` | `make format` | Outer 0, 5.947 s. |
| `remediation-artifact-paths-binding-20261015-01` | Redaction and snapshot/release integration rows | Outer 0, 34.409 s; required-field validation now resolves frozen source bindings at the Reporting boundary before native rendering. |
| `remediation-artifact-paths-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.730 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-artifact-paths-owner-20261015-02` | Full Reporting owner | Outer 2, 23/24, 51.812 s; unsafe-label fixture still put its angle brackets in the source selector instead of the visible value. All 28 cleanup steps completed. |
| `remediation-artifact-paths-labels-20261015-01` | Exact decoder/registered-reason row | Outer 0, 2.929 s; unsafe text now exercises the rendered value boundary. |

### Canonical token artifacts

The candidate specification and implementation now use redaction-manifest,
token-manifest and reveal-map version 3. A field's redaction entry owns its
selected rule, precedence, action, partition handling and nullable token ID.
Token manifests contain one exact subject identity, display digest and admitted
recipient set per token. They bind the complete redaction-manifest digest.
Private reveal maps contain one canonical display value per token and bind the
token-manifest digest. This acyclic structure removes conflicting subject-level
rule unions and arbitrary per-field original values. The bundle independently
binds both public manifests; native completion and recovery independently compare
all three artifacts against deterministic redaction of retained admission data.

Canonical Host/Identity/Party display values come only from frozen `display_name`.
Unresolved mentions use frozen `raw_text`, never normalized text or an inferred
identifier. Absence rejects tokenization rather than inventing a private value.
Public artifacts contain the digest, never that display value. Private maps remain
outside renderer inputs, ordinary bundle/download inventories and portability.

The complete v3 entry member sets are closed, with explicit nullable identity
arms and arrays. Tests remove every required top-level and entry member, inject
retired fields, alter identities, rules, token ordinals and recipients, and
recompute hashes. The strongest native regression consistently rewrites the
private value, public subject digest, member bytes and enclosing bundle hash;
public integrity validation passes but the frozen-admission comparison rejects it.
Portable import/export now validates current public artifact shapes and profile /
selected-rule bindings instead of accepting arbitrary rehashed artifact bytes.

Migration 78 rejects retained snapshots, jobs, imported artifacts and previews
before mutation; its marker replaces migration 77 as the preflight boundary.
Release and preview rows require v3 redaction manifests. The disposable migration
rehearsal includes version 77 and proves rejection preserves data and migration
state. No user database was reset. Rollback remains a matching application/data
pair or fresh initialization; retired v2 data is not converted.

This closes the token/reveal shape and binding slice, subject to the qualification
results below. It does not close section/block census, structured post-redaction
output, template-driven derivation, all indirect disclosure rules or overall
Reporting adoption. Historical v2 entries above describe prior evidence only.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-token-artifacts-format-20261015-01` | `make format` | Outer 0, 5.692 s. |
| `remediation-token-artifacts-unit-20261015-01` | Focused Reporting selection | Outer 2; configuration failure, nonexistent row ID; no execution assurance. Corrected to catalog row `module.reporting.unit.frozen_release_payload`. |
| `remediation-token-artifacts-unit-20261015-02` | Redaction, recipients and frozen payload | Outer 2, 0/1, 3.210 s; positive scalar fixtures had no canonical display value. Explicit structured display fixtures replaced them; missing-display rejection remains tested. All 10 cleanup steps completed. |
| `remediation-token-artifacts-format-20261015-02` | `make format` | Outer 0, 5.722 s. |
| `remediation-token-artifacts-generate-20261015-01` | `make generate` | Outer 0, 9.300 s; authored migration/history/ownership projections regenerated. |
| `remediation-token-artifacts-unit-20261015-03` | Same three rows | Outer 0, 2.747 s. |
| `remediation-token-artifacts-boundary-20261015-01` | Artifact and snapshot/release rows | Outer 2, 3/4, 29.382 s; unresolved-mention display selection was absent and migration-77 fixture omitted its required subject array. Both corrected at their owning boundary; all 15 cleanup steps completed. |
| `remediation-token-artifacts-format-20261015-03` | `make format` | Outer 0, 5.671 s. |
| `remediation-token-artifacts-boundary-20261015-02` | Artifact, snapshot/release and record rows | Outer 2, 3/4, 35.072 s; the new integration assertion compared a redaction digest with the helper's rendered-primary return value. Corrected to independent release and bundle digest fields; all 15 cleanup steps completed. |
| `remediation-token-artifacts-format-20261015-04` | `make format` | Outer 0, 5.749 s. |
| `remediation-token-artifacts-unit-20261015-04` | Four artifact/redaction/admission rows | Outer 2, 0/1, 3.072 s; the self-consistent scope-tamper fixture now failed earlier against profile scope. It now switches between two profile-supported internal scopes to preserve the independent admission-binding assertion. All 10 cleanup steps completed. |
| `remediation-token-artifacts-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.772 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-token-artifacts-unit-20261015-05` | Four artifact/redaction/admission rows | Outer 0, 3.028 s. |
| `remediation-token-artifacts-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.495 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-token-artifacts-migrations-20261015-01` | `make migration-drift` | Outer 0, 22.550 s. |
| `remediation-token-artifacts-lint-20261015-01` | `make lint` | Outer 2, 10/11, 26.812 s; `harness/child_target_failure`, two Go files needed formatting. All 20 cleanup steps completed. |
| `remediation-token-artifacts-owner-20261015-01` | Full Reporting owner | Outer 2, 23/24, 55.566 s; `infra/preflight_error` aggregate. A stale test accessed the removed top-level recipient array and panicked, leaving later selectors without results. This is an attributed fixture defect, not intermittent evidence. All 28 cleanup steps completed. |
| `remediation-token-artifacts-format-20261015-05` | `make format` | Outer 0, 5.729 s. |
| `remediation-token-artifacts-integration-20261015-01` | Exact snapshot/release/replay/cutover row | Outer 0, 40.746 s; validates migration 78, current release/token/reveal provenance, fresh recipient assertions and retained private recovery. |
| `remediation-token-artifacts-format-20261015-06` | `make format` | Outer 0, 5.579 s. |
| `remediation-token-artifacts-finalize-20261015-03` | `make agent-finalize` | Outer 0, 19.663 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-token-artifacts-owner-20261015-02` | Full Reporting owner, including installed native rendering and sandbox cases | Outer 0, 70.335 s. |
| `remediation-token-artifacts-lint-20261015-02` | `make lint` | Outer 0, 36.609 s. |
| `remediation-token-artifacts-portability-20261015-01` | Exact Incident Bundles reference-catalog/native-snapshot round-trip row | Outer 0, 53.163 s; current public artifact admission and private-map exclusion. |
| `remediation-token-artifacts-composition-20261015-01` | Full Composition owner | Outer 2, 4/5, 33.988 s; `product/test_assertion_failure`, minimal relational facade fixture supplied an empty manifest contrary to the new v3 database constraint. All 18 cleanup steps completed. |
| `remediation-token-artifacts-composition-20261015-02` | Exact release-tuple facade row | Outer 0, 10.207 s; fixture supplies the current schema marker and explicitly remains relational-only evidence, not rendering qualification. |
| `remediation-token-artifacts-markdown-20261015-01` | `make lint-markdown` | Outer 0, 9.290 s; before these final result rows. |
| `remediation-token-artifacts-finalize-20261015-04` | `make agent-finalize` | Outer 0, 19.865 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-token-artifacts-composition-20261015-03` | Full Composition owner | Outer 0, 22.267 s. |
| `remediation-token-artifacts-drift-20261015-01` | `make generate-drift` | Outer 0, 14.235 s. |
| `remediation-token-artifacts-shapes-20261015-01` | `make json-shape-check` | Outer 0, 5.184 s. |
| `remediation-token-artifacts-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.881 s. |


## Reference disclosure closure

The pre-render disclosure pass now computes restrictions over the frozen
reference graph. A field inherits each support target's restrictions, and a
relationship field inherits both exact endpoint records. Record targets include
the union of their domain and Records-envelope projections. An unresolved
Host/Identity mention is internal-only internally and blocked externally.
Propagation is monotone and terminates on cycles; input order cannot change the
result. Missing, foreign or wrong-family endpoints fail without partial output.
Cancellation is observed during propagation, with no usable partial result.

The graph contains identifiers and labels only. It does not inspect prose,
query current source state, reinterpret recipient permissions, erase earlier
restrictions, or infer an aggregate-public proof. A selected drop of a referenced
target cannot make an unchanged citation public. Existing action selection and
its rejection of unresolved retaining outcomes remain authoritative. The
candidate's Table 12-C now states the reference inheritance and fixed-point
rules explicitly. No artifact schema or adopted version changed in this slice.

The existing positive partition fixture incorrectly described a summary as
public while citing the entire incident, including its restricted field. It now
cites a separate, explicitly public Evidence record. Its selected-action and
manifest-trace assertions remain. New cases exercise transitive references,
cycles, multi-Party permissions, an unresolved mention, earlier blocked labels,
traversal-order independence, cancellation, wrong-family/foreign/missing
endpoints, and a production render denial that remains after the source is
dropped. The independent public-field survivor passes after the dependent
consumer is also explicitly dropped.

This is bounded implementation closure for explicit source/support references.
Unlinked free-text derivation, complete container census, canonical structured
post-redaction output, template/Composition derivation, current packaged
qualification and coordinated adoption remain open. It does not promote the
proposed specification set or close all nine findings.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-reference-disclosure-format-20261015-01` | `make format` | Outer 0, 6.103 s. |
| `remediation-reference-disclosure-unit-20261015-01` | Exact snapshot-record, redaction and recipient rows | Outer 2, 0/1, 4.203 s; `product/test_assertion_failure`. New reference tests passed; the prior positive summary fixture cited a restricted incident. All 10 cleanup steps completed. |
| `remediation-reference-disclosure-format-20261015-02` | `make format` | Outer 0, 5.850 s. |
| `remediation-reference-disclosure-unit-20261015-02` | Same three rows, corrected public-source fixture and multi-Party citations | Outer 0, 3.195 s. |
| `remediation-reference-disclosure-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.145 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-reference-disclosure-owner-20261015-01` | Full Reporting owner | Outer 2, 23/24, 73.115 s; `product/test_assertion_failure`. Two positive external-lifecycle fixtures still contained an unresolved Host mention. The disclosure denial was correct. All 28 cleanup steps completed. |
| `remediation-reference-disclosure-lint-20261015-01` | `make lint` | Outer 0, 33.710 s; before the mention-binding extension. |
| `remediation-reference-disclosure-format-20261015-03` | `make format` | Outer 0, 6.108 s. |
| `remediation-reference-disclosure-unit-20261015-03` | Three exact unit rows | Outer 2, 0/1, 3.083 s; `product/test_assertion_failure`. Missing-subject admission now fails earlier with `token_subject_not_stable`; the old test expected later token-selection ambiguity. The exact expected owner reason was corrected; all 10 cleanup steps completed. |
| `remediation-reference-disclosure-format-20261015-04` | `make format` | Outer 0, 5.839 s. |
| `remediation-reference-disclosure-unit-20261015-04` | Three exact unit rows | Outer 0, 2.862 s; includes resolved/unresolved metadata, canonical token/display binding and missing census rejection. |

The reference slice also repairs mention resolution. Frozen mention origin,
semantic field key, family, status and nullable canonical target are mandatory.
Resolved mentions bind the selected canonical subject and reuse its token and
private display value; raw aliases do not become canonical reveal values.
Unresolved mentions retain their own identity and restrict the originating
record. The current projection has no finer admitted semantic split, so origin
and mention use the combined restriction set. Admission and replay reject a
missing mandatory subject, missing origin, wrong target family, mixed resolution
arms and unknown/missing status instead of reconstructing a census.

The positive external-lifecycle fixtures now explicitly resolve their Host
mention to the Party-authorized Host before snapshot admission. The lifecycle
entrypoint additionally runs the original unresolved state through authenticated
snapshot and release Jobs, requires the exact disclosure denial, verifies failed
release recovery readiness and requires no successful output identities.
Synthetic unit mention fixtures now provide complete source metadata rather
than rely on an inferred unresolved state. No entrypoint was removed or skipped.


| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-reference-disclosure-integration-20261015-01` | Exact snapshot/replay/cutover and external-lifecycle rows | Outer 0, 41.414 s; includes actual unresolved denial, no successful output identities, resolved publication, native recovery and replay. |
| `remediation-reference-disclosure-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.443 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-reference-disclosure-owner-20261015-02` | Full Reporting owner, including installed native rendering and sandbox cases | Outer 0, 67.616 s. |
| `remediation-reference-disclosure-portability-20261015-01` | Existing Incident Bundles native-snapshot/reference round-trip row | Outer 0, 52.725 s. |
| `remediation-reference-disclosure-lint-20261015-02` | `make lint` | Outer 0, 34.277 s. |
| `remediation-reference-disclosure-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/evidence-inputs/remediation-reference-disclosure-reporting-20261015-01.json` | Outer 0, 2.720 s; explicit full-owner source-compatible retained evidence, not specification completeness or adoption. |
| `remediation-reference-disclosure-markdown-20261015-01` | `make lint-markdown` | Outer 0, 23.380 s; before these final result rows. |

The owner audit's manifest names only
`.cartulary/test-results/remediation-reference-disclosure-owner-20261015-02`
as its `test-slice` root. All measurements in this checkpoint are dirty-source
execution observations, not qualified cost windows or comparative speed claims.


## Structural removal accounting

The candidate now distinguishes a selected Core redaction action from the
consequence of emptying a structural container. Earlier REQ-RPT-059b required a
`dropped` rule entry for every empty ancestor, even when no Core drop rule had
been selected for that ancestor. Inventing that rule or borrowing a child's
rule would make the evidence false. Redaction manifest v4 adds a required
`structural_removals` array while retaining exact selected-rule entries.
Token and reveal shapes remain v3; their existing digest edges bind the complete
new manifest without adding cycles or compatibility readers.

Structural accounting walks the admitted block tree bottom-up. A block becomes
empty only when no field or child block survives; its causal list contains the
exact immediate dropped fields and removed child blocks. Non-validation
sections record removal only when all blocks are gone. An empty validation
section remains eligible. Paths retain pre-removal identities, including escaped
keys, and arrays are canonical and strictly ordered. Causal evidence rejects
self/cross-parent edges, skipped ancestors, missing or retained causes,
duplicates, absent arrays, wrong container kinds and unknown reasons. Native
admission/recovery recomputes the inventory from retained input; portable
admission validates the closed artifact and causal links without claiming to
reconstruct an unseen original inventory.

Migration 79 rejects retained snapshots, Jobs, imports and previews before
mutation and replaces the two v3 manifest constraints with v4 constraints.
The preflight marker follows the current constraint. The clean-cutover rehearsal
includes predecessor 78. No user data is reset or converted; rollback remains a
matching application/data backup or fresh initialization. The relational-only
Composition fixture uses the current marker but does not claim rendered-artifact
validation. Authored migration history/ownership inputs were regenerated through
Make.

This supplies structural removal accounting for the current field selection.
It does not yet replace the flat redacted model with the canonical retained
section/record/event graph, evaluate every container rule, recompute all retained
structural IDs or implement complete template/Composition derivation. Those
remain required work before adoption. Both candidate documents remain proposed.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-structural-removal-format-20261015-01` | `make format` | Outer 0, 5.851 s. |
| `remediation-structural-removal-generate-20261015-01` | `make generate` | Outer 0, 10.360 s. |
| `remediation-structural-removal-unit-20261015-01` | Exact redaction, frozen-release and model/artifact rows | Outer 0, 6.208 s; includes nested causes, escaped keys, retained-child preservation, validation sections, incomplete decisions and adversarial closed artifact shapes. |
| `remediation-structural-removal-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.136 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-structural-removal-owner-20261015-01` | Full Reporting owner, installed rendering/sandbox and migration-79 cutover | Outer 0, 75.080 s. |
| `remediation-structural-removal-composition-20261015-01` | Full Composition owner | Outer 0, 51.112 s. |
| `remediation-structural-removal-migrations-20261015-01` | `make migration-drift` | Outer 0, 39.564 s. |
| `remediation-structural-removal-lint-20261015-01` | `make lint` | Outer 0, 35.983 s. |
| `remediation-structural-removal-drift-20261015-01` | `make generate-drift` | Outer 0, 13.985 s. |
| `remediation-structural-removal-shapes-20261015-01` | `make json-shape-check` | Outer 0, 5.133 s. |
| `remediation-structural-removal-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.896 s. |
| `remediation-structural-removal-portability-20261015-01` | Exact Incident Bundles native-snapshot/reference round-trip row | Outer 0, 33.034 s. |
| `remediation-structural-removal-audit-20261015-01` | Explicit full Reporting owner evidence audit | Outer 0, 2.690 s; before the subsequent known-child completeness tightening. |
| `remediation-structural-removal-markdown-20261015-01` | `make lint-markdown` | Outer 0, 9.380 s; before subsequent prose/results. |
| `remediation-structural-removal-format-20261015-02` | `make format` | Outer 0, 5.812 s. |
| `remediation-structural-removal-unit-20261015-02` | Same three exact artifact/redaction rows | Outer 0, 3.252 s; adds omitted-known-cause denial and coexistence of an actual retaining rule with later structural removal. |

The causal validator also requires every published immediate child of a removed
container to appear in its causal inventory and be removed. This prevents a
forged empty cause list from contradicting another decision in the same public
manifest. Native frozen-source comparison remains necessary for unpublished or
omitted original inventory. A container's actual retaining rule may coexist
with later structural removal: these describe distinct stages. A selected Core
drop may not be double-counted as consequential empty removal. Typed path parsing
handles ancestry by parsed components, including escaped identifiers, without
prefix guesses. These refinements are part of the still-proposed v4 contract.


The final structural-validation runs all completed with native exit 0:

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-structural-removal-finalize-20261015-02` | `make agent-finalize` | 19.489 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-structural-removal-owner-20261015-02` | Full Reporting owner | 67.492 s. |
| `remediation-structural-removal-portability-20261015-02` | Exact Incident Bundles round-trip row | 59.600 s. |
| `remediation-structural-removal-lint-20261015-02` | `make lint` | 34.592 s. |
| `remediation-structural-removal-audit-20261015-02` | Full Reporting evidence audit using explicit `remediation-structural-removal-reporting-20261015-02.json` manifest | 2.676 s; only owner-02 is named. |

These observations predate the subsequent retained-field-state changes. They do
not qualify a clean performance window or close specification adoption.

## Retained field-state policy

Redaction now validates each retained value against its independently frozen
source state and the actual selected outcome before returning any model or
artifact bytes. Explicit allow records `allowed`; profile-default allow records
`unchanged`. Null remains null; unavailable and withheld states retain their
separate allowed actions. No action rewrites the source state to make a value
renderable. Invalid transformations return `redaction_manifest_invalid` /
`redaction_action_unresolved` and no partial result.

Built-in profile revision 3 freezes exact null-preserving path rules in the
external and tokenized-review profiles. External working material still drops.
Null fields do not acquire subject tokens; an explicit null allow remains
subject to ordinary disclosure qualification. Revision 2 is rejected, including
at route selection. The profile schema stays v2 because its shape is unchanged.
The internal profile retains its passthrough semantics within the revised
built-in catalog. Migration 80 rejects historical artifacts before mutation and
requires the current profile revision in new render admissions; the current
preflight marker and predecessor-79 rehearsal follow that contract. No user
history was reset or converted.

The field-state matrix, production null-action denial, explicit frozen null
rules and old-version rejection use existing owner entrypoints. The reference
closure fixture now replaces an existing exact-path rule when deliberately
constructing its custom drop profile; appending a competing rule correctly
failed duplicate-selector admission. Its original source/citation denial remains
required. This change does not yet close complete canonical model, template,
Composition, or coordinated adoption requirements.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-retained-fields-format-20261015-01` | `make format` | Outer 0, 5.848 s. |
| `remediation-retained-fields-generate-20261015-01` | `make generate` | Outer 0, 10.040 s. |
| `remediation-retained-fields-unit-20261015-01` | Exact redaction, frozen-release, recipient-profile and route-decoder rows | Outer 0, 3.494 s. |
| `remediation-retained-fields-integration-20261015-01` | Exact snapshot/cutover/replay and external-lifecycle rows | Outer 0, 48.378 s. |
| `remediation-retained-fields-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.339 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-retained-fields-owner-20261015-01` | Full Reporting owner | Outer 2, 23/24 work units, 68.455 s; `product` / `test_assertion_failure`, duplicate null-policy/custom-drop selector in reference-closure fixture; all 28 cleanup steps complete. |
| `remediation-retained-fields-lint-20261015-01` | `make lint` | Outer 0, 36.495 s. |
| `remediation-retained-fields-migrations-20261015-01` | `make migration-drift` | Outer 0, 22.903 s. |
| `remediation-retained-fields-drift-20261015-01` | `make generate-drift` | Outer 0, 26.716 s. |
| `remediation-retained-fields-format-20261015-02` | `make format` | Outer 0, 5.813 s. |
| `remediation-retained-fields-unit-20261015-02` | Exact snapshot-record materialization row after fixture correction | Outer 0, 2.901 s. |
| `remediation-retained-fields-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.712 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-retained-fields-owner-20261015-02` | Full Reporting owner | Outer 0, 61.004 s. |
| `remediation-retained-fields-portability-20261015-01` | Exact Incident Bundles native snapshot/reference round-trip | Outer 0, 49.092 s. |
| `remediation-retained-fields-shapes-20261015-01` | `make json-shape-check` | Outer 0, 5.764 s. |
| `remediation-retained-fields-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.944 s. |
| `remediation-retained-fields-audit-20261015-01` | Full Reporting evidence audit with explicit retained-fields manifest naming owner-02 | Outer 0, 2.655 s; predates retained-section work. |

## Retained section reconstruction

The rendering path now reconstructs its section tree after field selection.
Selected rule evidence preserves pre-removal addresses. Retained fields keep
frozen source state and exact selected values; empty blocks and non-validation
sections disappear; field/block ordinals, structural IDs, parent links,
container references and content-class counts are rebuilt. Values and mutable
reference arrays are detached from admitted source. Deck derivation uses this
retained tree rather than the old field decision index, retaining section titles
and block order. The unused `field_summary` serializer branch is removed and
its retired token is explicitly rejected. Renderer implementation version 7
identifies this changed derivation.

Frozen-source admission now validates section/container hierarchy, canonical
identities/order, complete arrays and exact reference unions. It rejects source
claims of generated aggregate proof, split identity or excluded fields.
Materialization records the actual section disclosure union. Reconstructing a
retained model cannot be used to repair contradictory admitted history.
Reference unions normalize once per container rather than sorting an expanding
section array after every source field.

This is a private rendering-stage projection. Full canonical public ExportModel
publication, duplicate record-view resolution, complete template manifests and
bindings, non-paragraph serializers, container-action selection, timeline and
diagram derivation, and the remaining Composition operations are still open.
The flat decision index remains temporarily for evidence and the existing graph
adapter; it is not proof of the complete canonical model contract. Neither
candidate specification is adopted by this checkpoint.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-retained-sections-format-20261015-01` | `make format` | Outer 0, 6.021 s. |
| `remediation-retained-sections-unit-20261015-01` | Exact redaction, post-redaction/serializer and renderer-result rows | Outer 0, 3.274 s; retained tree, nested pruning, state preservation and index-injection resistance. |
| `remediation-retained-sections-format-20261015-02` | `make format` | Outer 0, 6.067 s. |
| `remediation-retained-sections-unit-20261015-02` | Exact redaction, snapshot-record, post-redaction and metadata/artifact rows | Outer 0, 3.373 s; adds frozen hierarchy, hostile rehash and reference-union denial. |
| `remediation-retained-sections-format-20261015-03` | `make format` | Outer 0, 5.935 s. |
| `remediation-retained-sections-unit-20261015-03` | Same four exact model/redaction rows, with normalized container references | Outer 0, 3.378 s. |
| `remediation-retained-sections-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.641 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-retained-sections-owner-20261015-01` | Full Reporting owner | Outer 2, 23/24 units, 69.965 s; harness `infra` / `preflight_error`. The worker fixture's nil section inventory caused fixture admission to panic before its deadline scenario and left other selectors unexecuted; all 28 cleanup steps completed. |
| `remediation-retained-sections-portability-20261015-01` | Exact Incident Bundles round-trip | Outer 0, 59.814 s. |
| `remediation-retained-sections-composition-20261015-01` | Full Composition owner | Outer 0, 41.308 s. |
| `remediation-retained-sections-lint-20261015-01` | `make lint` | Outer 0, 35.751 s. |
| `remediation-retained-sections-format-20261015-04` | `make format` | Outer 0, 5.939 s. |

The worker fixture now uses normal snapshot materialization and render binding
for its synthetic incident. No production admission rule is bypassed, and the
existing deadline, cancellation, failed-release and cleanup assertions remain.
The failed owner root retains its original harness classification; the cause is
a deterministic test fixture defect, not evidence of intermittent infrastructure.


| Follow-up run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-retained-sections-unit-20261015-04` | Exact worker deadline/cancellation, renderer-result and render-export-participant rows | Outer 0, 6.271 s after normal-materialization fixture repair. |
| `remediation-retained-sections-finalize-20261015-02` | `make agent-finalize` | Published pass, 19.658 s; retained-secret scan passed. Native session was unavailable after context recovery, so outer exit is unavailable. `RESULTS_DIR` unset; retained-run maintenance skipped. |
| `remediation-retained-sections-owner-20261015-02` | Full Reporting owner | Outer 0, 52.605 s, 24/24 units; all 28 cleanup steps complete. |
| `remediation-retained-sections-audit-20261015-01` | Reporting owner evidence audit with explicit retained-sections manifest naming owner-02 | Outer 0, 2.595 s; valid for the retained-section checkpoint, before subsequent diagram changes. |

## Selected Timeline diagrams

Renderer version 8 removes the fabricated field-to-star-graph fallback. Mermaid
with no retained diagram declaration fails with `no_diagrams_selected` before
renderer execution. Timeline declarations resolve exact IDs against the retained
Timeline projection, reuse Core ordering facts, emit one vertex per selected
record and only adjacent directed edges. Removed events cannot re-enter through
a composition reference. Retained field values and reference unions come from
redaction; source sort timestamps never become display labels. Without an
explicit template time-display binding, time-display members remain null and
record IDs supply the specified fallback.

Flowchart and sequence vertex limits apply independently. Internal summarization
keeps the ordered prefix and exact omitted vertex/edge counts. Composition alone
cannot authorize the external template summarization exception. Timeline edge
references now have a documented canonical, directional, schema-domain identity
preimage, independent of renderer ordinals. Exact opaque label targets resolve
against selected vertices/edges; generated Mermaid IDs cannot target them.

The former Mermaid fixtures now name actual Timeline selection. The installed
runtime candidate test likewise carries an explicit Timeline source and sealed
composition. The first focused run rejected two fixtures with an empty incident
identity; binding validation was preserved and those fixtures received their
explicit synthetic incident identity. Graph kind selection now honors sequence
limits rather than silently rendering every declaration as a flowchart. A
replacement declaration with no template target fails explicitly. Graph label
overrides fail closed while retained subject-aware graph mapping remains open;
this checkpoint does not claim that required graph redaction is complete.

The complete public ExportModel/template binding stages, full template-owned
diagram declarations, graph disclosure/redaction materialization, manual diagram
rendering in Slidev, remaining Composition effects, and coordinated adoption
remain open. Source-only and synthetic runtime results do not establish native
rendering or packaged deployment qualification.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-timeline-diagrams-format-20261015-01` | `make format` | Outer 0, 5.800 s. |
| `remediation-timeline-diagrams-unit-20261015-01` | Exact metadata/artifact and redaction rows | Outer 2, 3.073 s, `product` / `test_assertion_failure`; two Mermaid fixtures lacked an incident identity for composition binding. All 10 cleanup steps completed. |
| `remediation-timeline-diagrams-unit-20261015-02` | Exact metadata/artifact, redaction and graph-selection rows | Outer 0, 3.112 s after fixture correction; exact Timeline ordering, removed/missing/duplicate selection, bounds, labels, no invented output and unchanged graph selection assertions. |
| `remediation-timeline-diagrams-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.840 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-timeline-diagrams-owner-20261015-01` | Full Reporting owner | Outer 2, 66.113 s, 23/24 units; `product` / `test_assertion_failure`. The route/renderer denial fixture still relied on the removed field-to-diagram fallback. Installed rendering and sandbox cases passed; all 28 cleanup steps completed. |
| `remediation-timeline-diagrams-lint-20261015-01` | `make lint` | Outer 2, 34.067 s; staticcheck found the now-unused flat-field sorting helper. It was removed. |
| `remediation-timeline-diagrams-markdown-20261015-01` | `make lint-markdown` | Outer 0, 11.380 s before these evidence rows. |
| `remediation-timeline-diagrams-format-20261015-02` | `make format` | Outer 0, 5.860 s. |
| `remediation-timeline-diagrams-unit-20261015-03` | Exact route/renderer denial row | Outer 0, 3.563 s. The unsafe-label case now selects a real Timeline event and an exact composition label override; it retains the renderer-syntax rejection assertion. |
| `remediation-timeline-diagrams-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.037 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-timeline-diagrams-lint-20261015-02` | `make lint` | Outer 0, 34.096 s. |
| `remediation-timeline-diagrams-owner-20261015-02` | Full Reporting owner | Outer 0, 65.189 s, 24/24 units; all 28 cleanup steps completed. Source `sha256:de605e14927b5c6e6fc9970b3062c412963128cdc3408d9e9bbf1f76202265bb`, graph `sha256:67312714ff527160f6bffdfa04ddbe3bf7686fe52d52d4f535c159ae375e696c`. |
| `remediation-timeline-diagrams-audit-20261015-01` | Reporting owner evidence audit with explicit Timeline-diagrams manifest naming owner-02 | Outer 0, 2.825 s; source-scoped to the diagram checkpoint before later profile-binding changes. |

## Exact redaction selector binding

Profile shape validation no longer suffices to admit a frozen selection. Every
exact path rule must resolve to an executable candidate in the admitted model.
Stale fields, missing fields, root paths, unsupported containers and unsupported
duplicate-view addresses fail with `redaction_action_unresolved` before any
redaction artifact is returned. Frozen payload admission/replay uses the same
binding check. No fallback to a class/default rule may hide an unmatched exact
rule. Valid field drops still retain their selected-rule evidence.

This closes silent acceptance of unsupported exact selectors. It does not
implement container action semantics, prove element-level disclosure closure,
or qualify Reporting adoption. Container selection remains required work;
rejection is explicitly a withheld capability, not successful conformance.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-redaction-bindings-format-20261015-01` | `make format` | Outer 0, 6.071 s. |
| `remediation-redaction-bindings-unit-20261015-01` | Exact redaction, frozen-release and route-decoder rows | Outer 0, 3.297 s; supported exact drop, missing/unsupported target rejection, frozen binding and no partial artifacts. |
| `remediation-redaction-bindings-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.541 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-redaction-bindings-lint-20261015-01` | `make lint` | Outer 0, 31.997 s. |
| `remediation-redaction-bindings-owner-20261015-01` | Full Reporting owner | Outer 0, 62.776 s, 24/24 units. Eleven compatible cached units were reused; thirteen units executed. No cached pass is a new execution observation. |
| `remediation-redaction-bindings-audit-20261015-01` | Explicit Reporting owner evidence manifest naming bindings owner-01 | Outer 0, 2.623 s; validates source-scoped owner evidence before subsequent edits. |

## Unique canonical field identities

Every canonical field list now rejects duplicate keys before snapshot/model
bytes are returned, including section blocks, record summaries and Timeline
events. The derivation specification now requires rejection rather than
traversal-dependent omission of the later value. This preserves complete source
accounting and prevents keyed references from selecting an ambiguous field.
Replay and recovery use the same admission boundary; rehashing a contradictory
model does not make it valid.

Subject-family resolution now runs once over the complete assembled source
record. Disjoint mention contributions can supply separate required metadata
fields without repeating identity facts. The existing frozen mention validator
still owns the complete binding and contradictory families still fail. This
removes the per-contribution inference that forced fixtures to repeat fields.

The token alias fixture uses a distinct `dns_name` field on the same Host,
preserving the same-subject token and single canonical reveal-value assertions.
The multi-field Timeline fixture uses distinct metadata/first keys and checks
that its renamed source binding is actually encountered. Negative cases cover
both source traversal orders and hostile duplicate keys in each frozen view.
No source value is silently dropped, merged by spelling or assigned an ordinal
identity. Cross-family record-envelope normalization and the complete public
model remain separate open work.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-field-identity-format-20261015-01` | `make format` | Outer 0, 5.893 s. |
| `remediation-field-identity-unit-20261015-01` | Exact snapshot-record, redaction and metadata/artifact rows | Outer 2, 2.679 s; `product` / `test_assertion_failure`. Existing token-alias and multi-contribution mention fixtures repeated canonical keys. All 10 cleanup steps completed. |
| `remediation-field-identity-format-20261015-02` | `make format` | Outer 0, 5.882 s after record-level subject-family resolution and fixture corrections. |
| `remediation-field-identity-unit-20261015-02` | Exact snapshot-record, redaction and metadata/artifact rows | Outer 2, 3.088 s; `product` / `test_assertion_failure`. The alias drop rule still addressed the old field key; all 10 cleanup steps completed. |
| `remediation-field-identity-unit-20261015-03` | Same exact rows after correcting the alias selector | Outer 0, 2.874 s; duplicate-key rejection, complete-record subject binding and same-subject token behavior pass. |
| `remediation-field-identity-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.595 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-field-identity-lint-20261015-01` | `make lint` | Outer 0, 33.421 s. |
| `remediation-field-identity-owner-20261015-01` | Full Reporting owner | Outer 0, 64.045 s; source `sha256:f7bec5f4007403f24972cb4f4a2fd4d858e71b7b22da2f3ad446363b0f88752d`, graph `sha256:67312714ff527160f6bffdfa04ddbe3bf7686fe52d52d4f535c159ae375e696c`. |
| `remediation-field-identity-audit-20261015-01` | Explicit Reporting owner evidence manifest naming field-identity owner-01 | Outer 0, 2.678 s; source-scoped to that checkpoint. |

## Typed deck model and block serialization

Deck derivation carries retained block and field schemas through chunking,
instead of replacing each block with an implementation-specific text member.
The deck records materialized output options and source/support unions. Slides
use `slide_ordinal`, `layout_id`, explicit empty click steps, nullable speaker
notes, and reference/partition unions. Removed private derivation/composition
diagnostics are no longer serialized as canonical deck members. Renderer
revision 9 identifies the changed output semantics. No legacy reader was added.

Paragraph, list, table, metric and speaker-note serializers consume typed
retained values. Structure is serializer-owned and text is escaped. Malformed
tables cannot shift columns, nested lists are bounded, and note blocks join
slide-level notes in one final delimiter-checked comment. Chunking detaches
field arrays and reference lists, recomputes local ordinals and references, and
keeps whole-slide notes only on the first chunk. The shared parser vector now
uses typed blocks and the declared layout/note members.

Template expansion, semantic composition targeting, authored-text redaction,
click generation, timeline/diagram/asset/overflow bindings and the full public
ExportModel remain open. Unsupported targets do not acquire a text fallback.
This checkpoint alone does not satisfy coordinated adoption.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-typed-deck-format-20261015-01` | `make format` | Outer 0, 5.947 s. |
| `remediation-typed-deck-unit-20261015-01` | Exact metadata/artifact and redaction rows | Outer 0, 3.270 s; typed serialization, retained values, immutable copies and slide wire members. |
| `remediation-typed-deck-limits-20261015-01` | Exact stage-admission and worker/diagnostic rows | Outer 0, 4.687 s; chunking, note limits and exhaustive failure diagnostics. |
| `remediation-typed-deck-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.762 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-typed-deck-markdown-20261015-01` | `make lint-markdown` | Outer 0, 12.980 s before these final evidence rows. |
| `remediation-typed-deck-lint-20261015-01` | `make lint` | Outer 2, 38.552 s; `harness` / `child_target_failure`. Biome requested formatting of the added shared-vector parser assertion; Go lint and frontend type checking passed. All 20 cleanup steps completed. |
| `remediation-typed-deck-owner-20261015-01` | Full Reporting owner | Outer 0, 69.543 s, including installed rendering of the expanded typed-block vector and candidate sandbox pipeline. |
| `remediation-typed-deck-format-20261015-02` | `make format` | Outer 0, 5.989 s. |
| `remediation-typed-deck-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.668 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-typed-deck-lint-20261015-02` | `make lint` | Outer 0, 33.355 s after formatting. |
| `remediation-typed-deck-owner-20261015-02` | Full Reporting owner | Outer 0, 64.135 s, with the unbound-click rejection included. Source `sha256:654ee7f15ad4f64c898cc33128c6aa42ac1ed89ae5ba1242029f8e85e237a278`, graph `sha256:67312714ff527160f6bffdfa04ddbe3bf7686fe52d52d4f535c159ae375e696c`. |
| `remediation-typed-deck-audit-20261015-01` | Explicit Reporting owner manifest naming typed-deck owner-02 | Outer 0, 2.774 s before the following source-binding changes. |

## Frozen presentation binding

Production Slidev rendering and native retained-output validation now share one
pure preparation function. Completion, replay and recovery compare exact deck
and source bytes and descriptors against the independently frozen input, rather
than trusting a mutually consistent bundle hash graph. Required deck/source
members cannot be missing, renamed, replaced or supplemented. PDF membership
must match admitted output options; source-only output cannot carry rendered
content. Unbound diagrams cannot be smuggled into the current Slidev bundle.
Native candidate admission also rejects a different or retired renderer contract.

The comparison does not execute a renderer or inspect current host toolchains.
It is separate from rendered-byte determinism and portable historical-artifact
integrity. Mermaid graph preparation still needs retained source-owned projection
material for the analogous recovery comparison; the current transient leased
projection reader cannot establish that requirement after lease expiry.
Toolchain artifact closure, full template expansion, authored-text redaction and
the remaining adoption obligations are still open.

Adversarial tests recompute member sizes/digests and the bundle digest before
asserting rejection. The database recovery test also updates the retained root
and release hash atomically, then verifies rejection at the presentation-binding
boundary after public Job response expiry. It restores the original bytes and
verifies recovery again. No production history was reset or rewritten.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-presentation-binding-format-20261015-01` | `make format` | Outer 0, 5.717 s. |
| `remediation-presentation-binding-unit-20261015-01` | Exact frozen-release and metadata/artifact rows | Outer 0, 3.399 s; self-consistent substitutions, retired shape, missing/extra artifacts and output-option denial. |
| `remediation-presentation-binding-integration-20261015-01` | Exact snapshot/reference-binding service row | Outer 0, 39.103 s; native completion, recovery, Job expiry and coherent deck/source substitutions. |
| `remediation-presentation-binding-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.433 s; `RESULTS_DIR` unset, retained-run maintenance skipped. Subsequent toolchain changes are qualified separately below. |

## Retained toolchain artifact closure

Both native and portable bundle admission now require the single named toolchain
artifact, its exact role/media/required descriptor, canonical closed bytes,
schema-domain digest, admission timestamp, and validation roster. Required
digest and environment fields are checked. Renderer/browser presence follows
the actual source/rendered roles; source-only bundles cannot claim a browser
execution. Arrays, environment-name ordering and argv shapes are explicit.
No validation reads the current installation or changes historical bytes.
This checks retained evidence integrity, not a fresh execution attestation.

Synthetic unit and application fixtures now populate the required toolchain
shape with explicit synthetic identities. They remain protocol/transaction
fixtures and do not establish native rendering or confinement. One old output-
kind substitution assertion needed updating because the new shared boundary
rejects its contradictory toolchain before native admission. The candidate
still must fail; no production validation was relaxed.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-toolchain-binding-format-20261015-01` | `make format` | Outer 0, 5.818 s. |
| `remediation-toolchain-binding-unit-20261015-01` | Exact frozen-release and metadata/artifact rows | Outer 2, 3.180 s; `product` / `test_assertion_failure`. The output-kind negative fixture is now rejected at shared toolchain admission. All 10 cleanup steps completed. |
| `remediation-toolchain-binding-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.770 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-toolchain-binding-unit-20261015-02` | Exact frozen-release and metadata/artifact rows | Outer 0, 3.044 s after correcting the negative fixture's expected rejection boundary. |
| `remediation-toolchain-binding-finalize-20261015-02` | `make agent-finalize` | Outer 0, 19.535 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-toolchain-binding-lint-20261015-01` | `make lint` | Outer 0, 34.828 s. |
| `remediation-toolchain-binding-markdown-20261015-01` | `make lint-markdown` | Outer 0, 11.510 s before these final evidence rows. |
| `remediation-toolchain-binding-portability-20261015-01` | Exact Incident Bundle native-snapshot round-trip service row | Outer 0, 55.225 s; current shared toolchain admission survives export/import. |
| `remediation-toolchain-binding-owner-20261015-01` | Full Reporting owner | Outer 0, 70.019 s; includes presentation recomputation, native/preview recovery and installed sandbox rendering. Source `sha256:0c7c83430323a8535d2161c149fd168bb9d01d4d2954a58471aea142bea6a1b1`, graph `sha256:67312714ff527160f6bffdfa04ddbe3bf7686fe52d52d4f535c159ae375e696c`. |
| `remediation-toolchain-binding-audit-20261015-01` | Explicit Reporting owner manifest naming toolchain-binding owner-01 | Outer 0, 2.627 s; evidence applicability confirmed before further implementation. |

## Retained graph sources and Mermaid presentation binding

Release and preview admission now capture exact source-owner Graph Projection
contributions in their admission transaction. The private versioned object owns
its nested values and excludes actual publication time. A separate database
digest binds this input. The worker compares renewed contributions to the
frozen input; it cannot substitute a new result. Completed-output validation
uses frozen graph sources without renewing a lease or reading current graph
state. These private inputs do not enter shareable artifacts.

Mermaid source preparation is shared by rendering and native candidate,
completion, replay and recovery validation. It recomputes diagram/source bytes
and exact membership before accepting output. SVG membership must match
admitted options. Tests substitute model/source bytes, update their sizes and
hashes plus the bundle root, and require rejection. Independent graph admission
digests reject changed labels/topology. Canonical rendering determinism remains
a separate obligation from source fidelity.

Migration 81 and preflight reject incompatible retained history before mutation;
no graph values are reconstructed. The migration includes a closed graph-input
schema check and nullable-only-for-snapshot independent digest. The cutover
rehearsal now includes predecessor 80. Existing pending-job Graph lease restore
requirements remain. Full template expansion, public ExportModel normalization,
authored-text redaction and coordinated adoption remain open.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-graph-retention-generate-20261015-01` | `make generate` | Outer 2, 1.780 s; `harness` / `unknown_failure`: migration 81 omitted from the authored source-owner allocation. No verification claim. |
| `remediation-graph-retention-format-20261015-01` | `make format` | Outer 0, 6.489 s. |
| `remediation-graph-retention-generate-20261015-02` | `make generate` | Outer 0, 10.150 s after assigning migration 81 to Reporting. |
| `remediation-graph-retention-unit-20261015-01` | Exact graph-selection/source and frozen-release rows | Outer 0, 3.897 s; frozen source identity, ownership and independent digest. |
| `remediation-graph-retention-format-20261015-02` | `make format` | Outer 0, 5.877 s. |
| `remediation-graph-retention-unit-20261015-02` | Exact graph-selection/source and frozen-release rows | Outer 0, 3.313 s; includes coherent Mermaid substitutions and missing/extra SVG membership. |
| `remediation-graph-retention-integration-20261015-01` | Exact snapshot/reference-binding and Graph source lease rows | Outer 0, 41.119 s; fresh initialization, predecessor-80 rejection and source-owner lease lifecycle. The subsequent independent-digest readiness assertion is included in the next owner run. |
| `remediation-graph-retention-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.910 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-graph-retention-lint-20261015-01` | `make lint` | Outer 0, 38.558 s. |
| `remediation-graph-retention-markdown-20261015-01` | `make lint-markdown` | Outer 0, 11.880 s before final evidence rows. |
| `remediation-graph-retention-drift-20261015-01` | `make generate-drift` | Outer 0, 29.268 s. |
| `remediation-graph-retention-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.959 s. |
| `remediation-graph-retention-shapes-20261015-01` | `make json-shape-check` | Outer 0, 5.057 s. |
| `remediation-graph-retention-owner-20261015-01` | Full Reporting owner | Outer 0, 71.157 s; includes independent graph-digest recovery denial, preview admission, cutover, installed rendering and sandboxing. Source `sha256:fa21da25fd976af53499f6baf577bbb87bee20643a42810f6113fbd0c46c1741`, graph `sha256:67312714ff527160f6bffdfa04ddbe3bf7686fe52d52d4f535c159ae375e696c`. |
| `remediation-graph-retention-audit-20261015-01` | Explicit Reporting owner manifest naming graph-retention owner-01 | Outer 0, 2.606 s before template implementation changes. |

## Declared template manifest and selection

The built-in version 3 template now carries the Table 21 manifest itself. The
private parallel binding lists and unused CSS asset were removed. Selection
resolves declared bindings against original admitted addresses and reads only
retained field values. Redaction cannot shift a binding onto another field by
renumbering blocks, nor can a required dropped field be refilled from the raw
snapshot. Sections use declaration titles and layout decisions. Output-option
support is checked per output kind. Migration 82 rejects incompatible history
before mutation; predecessor 81 is included in the disposable cutover rehearsal.

The new helper runs inside `TestSlidevCanonicalSourceV2` through the existing
metadata row. It covers manifest membership, retired versions, binding grammar
and containment, declaration identity, required empty sections, layout/options
validation, and source immutability. General expansion, narrative slots, asset
resolution, semantic Composition effects and canonical public ExportModel
materialization remain adoption obligations. The built-in manifest definition
does not remove those requirements.

All roots below are under `.cartulary/test-results/`. The restart lost native
handles for unit-03 and format-02; their published failures are retained, and
their owning outer exits are unavailable. No observer exit substitutes for them.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-template-manifest-generate-20261015-01` | `make generate` | Outer 0, 10.260 s. |
| `remediation-template-manifest-format-20261015-01` | `make format` | Outer 0, 5.847 s. |
| `remediation-template-manifest-unit-20261015-01` | Metadata, frozen-release and redaction rows | Outer 2, 0.699 s; `infra` / `preflight_error`: unused import prevented compilation. |
| `remediation-template-manifest-unit-20261015-02` | Same rows | Outer 2, 3.406 s; `product` / `test_assertion_failure`: synthetic retained trees lacked source bindings. |
| `remediation-template-manifest-unit-20261015-03` | Same rows | Published failure, 3.290 s; `product` / `test_assertion_failure`: text fixture omitted content classes. All 10 cleanup steps completed; outer exit unavailable. |
| `remediation-template-manifest-format-20261015-02` | `make format` | Published failure, 5.990 s; `harness` / `child_target_failure`: invalid Go escape in the new binding fixture. All 11 cleanup steps completed; outer exit unavailable. |
| `remediation-template-manifest-format-20261015-03` | `make format` | Outer 0, 5.873 s. |
| `remediation-template-manifest-unit-20261015-04` | Same three rows | Outer 2, 3.241 s; `product` / `test_assertion_failure`: expectation-only correction did not fix the missing fixture content classes. |
| `remediation-template-manifest-unit-20261015-05` | Same three rows | Outer 0, 2.949 s after fixing the fixture and preserving the invalid-text assertion. |
| `remediation-template-manifest-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.755 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-template-manifest-lint-20261015-01` | `make lint` | Outer 0, 34.679 s. |
| `remediation-template-manifest-markdown-20261015-01` | `make lint-markdown` | Outer 0, 9.510 s before these final evidence rows. |
| `remediation-template-manifest-owner-20261015-01` | Full Reporting owner | Outer 2, 67.554 s; `product` / `test_assertion_failure`: stale Mermaid fixtures requested Slidev PDF options and one unsupported-kind classification expectation changed. Other 23 work units passed; all 28 cleanup steps completed. |
| `remediation-template-manifest-decoder-20261015-01` | Exact route-decoder row | Outer 0, 3.528 s after giving Mermaid fixtures the corresponding output options. |
| `remediation-template-manifest-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.196 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-template-manifest-owner-20261015-02` | Full Reporting owner | Outer 0, 56.405 s. Source `sha256:63cdd69abf7c79a6cca804be3cebc934c2304764721002c7cb491d7cacbc25fa`, graph `sha256:67312714ff527160f6bffdfa04ddbe3bf7686fe52d52d4f535c159ae375e696c`. |
| `remediation-template-manifest-audit-20261015-01` | Explicit Reporting owner manifest naming template-manifest owner-02 | Outer 0, 2.687 s before semantic Composition changes. |

## Semantic composition and reveal steps

Reporting now retains the companion's typed anchors and applies section and
unqualified block operations to current retained objects before chunking.
Section removal, ordered reordering, title/note/layout/click overrides, nested
block removal and relative authored-block insertion have explicit effects.
Zero-match and ambiguous targets differ; internal drop applies only to a
zero-match operation. External drop is rejected at the common output-kind
boundary. Removed objects are not retained for later target resolution.
Reorders resolve atomically before mutation, and repeated inserts preserve
operation order. Structural block IDs, ordinals and descendant reference unions
are rebuilt after operations. Authored note partitions follow only the first
chunk; authored title partitions follow each titled chunk.

Layout tokens remain template identities in the deck. Serialization maps them
to the declared Slidev layout, and block-kind checks include nested blocks after
overrides. The incorrect `overflow_notice` template spelling was replaced by
the existing `overflow_summary` vocabulary, shared with section validation.
Migration 82 checks manifest shape rather than pinning a second registry of
template versions in SQL. The packaged release rehearsal now requests template
version 3; its previous version-2 evidence is historical until rerun.

The new reveal-only profile v2 resolves an actual specification inconsistency:
the pinned Slidev `VClicks` component reveals each direct list item, whereas
the former text counted one step for a whole list. The corrected profile counts
items, excludes non-body speaker notes from block reveals, and generates steps
only after chunk identities. The serializer recomputes target sets, component,
order, visibility and every state hash before emitting wrappers. Renderer
contract version 10 retires the previous native interpretation without a
fallback. No-click source retains its meaning. Native sandbox checks render
both block and list profiles, verify exact PDF page counts, repeatability,
observed loopback traffic and cleanup.

Record-qualified anchors still require canonical retained record summaries;
diagram operations, general template expansion and authored-text
partition/redaction/placeholder processing are not closed by these changes.
Those and public ExportModel normalization remain required adoption work.
No document has been promoted.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-composition-effects-format-20261015-01` | `make format` | Outer 0, 6.096 s. |
| `remediation-composition-effects-unit-20261015-01` | Metadata, frozen-release and renderer-stage admission rows | Outer 2, 3.338 s; `product` / `test_assertion_failure`: the new immutable composition fixture omitted the materialized `on_unresolved` default. All 10 cleanup steps completed. |
| `remediation-composition-effects-generate-20261015-01` | `make generate` | Outer 0, 10.300 s. |
| `remediation-composition-effects-format-20261015-02` | `make format` | Outer 0, 5.945 s. |
| `remediation-composition-effects-unit-20261015-02` | Same three rows | Outer 0, 2.952 s after fixing the canonical fixture. |
| `remediation-composition-effects-native-20261015-01` | Exact installed-renderer sandbox row | Outer 0, 49.995 s; real list/block reveals, page counts and repeatable bytes. Later note-partition changes require the next owner run. |
| `remediation-composition-effects-format-20261015-03` | `make format` | Outer 0, 5.855 s. |
| `remediation-composition-effects-unit-20261015-03` | Same three rows | Outer 0, 3.237 s; adds nested removal, note partition/chunk and common external-drop assertions. |
| `remediation-composition-effects-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.633 s; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-composition-effects-format-20261015-04` | `make format` | Outer 0, 5.750 s. |
| `remediation-composition-effects-lint-20261015-01` | `make lint` | Outer 0, 37.302 s. |
| `remediation-composition-effects-drift-20261015-01` | `make generate-drift` | Outer 0, 27.187 s. |
| `remediation-composition-effects-markdown-20261015-01` | `make lint-markdown` | Outer 0, 12.640 s before these final evidence rows. |
| `remediation-composition-effects-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.960 s. |
| `remediation-composition-effects-shapes-20261015-01` | `make json-shape-check` | Outer 0, 5.184 s. |
| `remediation-composition-effects-owner-20261015-01` | Full Reporting owner | Outer 2, 76.704 s; `product` / `test_assertion_failure`: layout rejection incorrectly stopped source-stage diagnostic collection. Other 23 work units passed; all 28 cleanup steps completed. |
| `remediation-composition-effects-diagnostics-20261015-01` | Worker/diagnostics and metadata rows | Outer 2, 6.067 s; after restoring exhaustive validation, the synthetic fixture's undeclared layout added an eighth diagnostic. All 10 cleanup steps completed. The fixture now uses the admitted built-in template. |
| `remediation-composition-effects-diagnostics-20261015-02` | Same two rows | Published pass; restart lost the native handle, so outer exit is unavailable. |
| `remediation-composition-effects-finalize-20261015-02` | `make agent-finalize` | Published pass; native outer exit unavailable after restart. `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `remediation-composition-effects-owner-20261015-02` | Full Reporting owner | Outer 0, 57.363 s. Source `sha256:bea7a7d051247a0dbabae81e34f4d2ea40578588dda8bb70cabe80ad38c0e9c1`, graph `sha256:67312714ff527160f6bffdfa04ddbe3bf7686fe52d52d4f535c159ae375e696c`. |
| `remediation-composition-effects-audit-20261015-01` | Explicit Reporting owner manifest naming composition-effects owner-02 | Outer 0, 2.692 s before further diagram implementation. |

## Native manual and embedded diagrams

Reporting now keeps the companion's closed layout data, resolves opaque selected
refs, and serializes manual flowcharts with exact boxes, clipped endpoints,
ordered routes, direction arrows and passive escaped text. The specification
pins the previously unspecified SVG byte presentation. Unsupported manual
Mermaid output still fails; it cannot silently switch to automatic layout.
Template declarations and composition replacements are reconciled before ID
uniqueness, and diagram exclusion/label operations execute against the retained
diagram set. Composition diagram slides use semantic section anchors, preserve
insertion order, reject duplicate insertion, and receive final structural IDs
after operations. General template expansion remains separate work.

Slidev preparation now includes every retained diagram model and auto Mermaid
source, plus deterministic manual SVG. Manual diagrams use local image paths;
auto diagrams use the specified Mermaid fence. The renderer checks local image
decoding before PDF capture. Stage protocol v3 binds a constrained, ordered SVG
asset list by exact UTF-8 bytes; Go admits passive SVG before launch, the stage
rejects escaping/duplicate paths and hashes, and file creation cannot overwrite
an existing file. Stage v1/v2 have no fallback. Renderer contract 11 retires the
former native interpretation. Completion and recovery recompute manual SVG,
diagram models and source from frozen inputs and check exact output membership.
Toolchain evidence distinguishes passive manual SVG from browser-produced SVG.

No application database or retained history was reset. This slice changes
rendered artifact semantics, not the retained request schema or public routes.
The clean pre-production cutover remains required. Public ExportModel
normalization, record-qualified anchors, authored-text redaction, graph
subject/disclosure handling, narrative slots and general template expansion
remain adoption obligations. In particular, graph label overrides and external
raw graph values still reject because their required redaction bindings are
not materialized. The successful manual path does not waive those obligations.

| Observable obligation | Owner / exact exercising test | Assertion and context | Coverage status / reviewer |
| --- | --- | --- | --- |
| RPT-086b/c exact manual geometry and bytes; RC-063b/d layout modes, targets and bounds | `TestSlidevCanonicalSourceV2/manual_and_embedded_diagrams`, existing metadata row | Independent byte vector, side/tie cases, waypoint order, bounds, unknown/duplicate/missing refs, zero-vector rejection, no input mutation, unsupported manual Mermaid | Unit demonstrated for the retained Timeline input; Codex implementation review |
| RPT-079e replacement and RPT-087g diagram effects | Same entrypoint, declaration ownership/operations subtest | Replacement before uniqueness, owner restrictions, later label wins, removed-target failure/drop, insertion order and duplicate insertion | Bounded unit coverage; graph token/partition semantics remain open |
| RPT-126 mixed diagram bundle | Same entrypoint plus `TestReportingInstalledSandboxRenders/mixed_manual_and_auto_diagram_candidate` | Exact manual/auto membership, real PDF with local SVG and Mermaid, independent native attempts, byte equality and attempt cleanup | Native run recorded below; no Core 05 publication claim |
| Native completion fidelity | Same metadata entrypoint | Rehash substituted SVG, source and model bytes; completion must reject | Unit demonstrated; broader recovery uses the shared binding boundary |
| Local asset confinement and byte binding | `TestReportingRenderStageRejectsUnboundOutputs/local_diagram_assets`; existing stage-admission Vitest title | Escaping/duplicate paths, active SVG, wrong digest, unsupported operation, nested duplicate keys and independent Unicode framing vector | Unit demonstrated; native installed test exercises actual transport and cleanup |

All roots below are under `.cartulary/test-results/`.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-diagram-rendering-format-20261015-01` | `make format` | Outer 0, 6.051 s. |
| `remediation-diagram-rendering-format-20261015-02` | `make format` | Outer 0, 5.903 s. |
| `remediation-diagram-rendering-unit-20261015-01` | Metadata, frozen-release and Go stage-admission rows | Outer 2, 3.243 s; `product` / `test_assertion_failure`: the new fixture rendered but lacked top-level snapshot/template identity for completion validation. All 10 cleanup steps completed. |
| `remediation-diagram-rendering-unit-20261015-02` | Metadata row | Outer 2, 2.798 s; retained the explicit completion rejection before fixture repair. All 10 cleanup steps completed. |
| `remediation-diagram-rendering-format-20261015-03` | `make format` | Outer 0, 5.959 s. |
| `remediation-diagram-rendering-unit-20261015-03` | Metadata, frozen-release, Go and TS stage-admission rows | Outer 0, 5.412 s after completing fixture identities. |
| `remediation-diagram-rendering-native-20261015-01` | Installed sandbox, stage execution and Slidev browser rows | Outer 0, 58.470 s; mixed diagram PDF, repeated native bytes and cleanup. The later direction-arrow addition is covered by the next owner run. |
| `remediation-diagram-rendering-format-20261015-04` | `make format` | Outer 0, 5.934 s. |
| `remediation-diagram-rendering-unit-20261015-04` | Same four unit rows | Outer 0, 5.611 s; includes direction arrows and operation-order/ownership assertions. |
| `remediation-diagram-rendering-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.126 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-diagram-rendering-lint-20261015-01` | `make lint` | Outer 0, 38.609 s. |
| `remediation-diagram-rendering-markdown-20261015-01` | `make lint-markdown` | Outer 0, 13.080 s; preceding documentation revision. |
| `remediation-diagram-rendering-owner-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 83.164 s; full owner including installed sandbox, mixed diagrams, service-backed completion and recovery. |
| `remediation-diagram-rendering-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/evidence-inputs/remediation-diagram-rendering-reporting-20261015-01.json` | Outer 0, 2.560 s; explicit manifest contains the immediately preceding full owner root. |

## Composition presentation-text redaction

The next slice closes the permission-only bypass for Composition prose. The
Reporting redaction owner admits each title, note and paragraph as a typed
presentation body, selects ordinary path/class/default rules, validates the
author's singleton partition, resolves explicit subject placeholders from
selected frozen values, and allocates shared subject tokens before emission.
Substitution and retained replacements obey role limits. Decisions cover every
declared body, including unused and dropped bodies. Dropped text has no deck
effect, but its semantic anchor must still resolve under the operation policy.
The renderer cannot reuse an admitted decision for a changed body.

The public export-model contract advances to v3 with a required empty
`presentation_texts` array at admission; snapshots remain v2 and exclude that
render-derived collection. The typed retained presentation nodes are generated
only from the independent Composition input. Native renderer contract 12 and
the private redacted projection v3 replace their prior meanings. Companion
participant declarations/digests and generated projections follow the new
schema. Migration 83 and the preflight marker require a clean pre-production
cutover and reject retained history before mutation. No reset of an existing
user environment is part of the rehearsal.

Composition authoring now rejects non-contractual partition strings, including
the previously accepted `internal` spelling. The normative token is
`internal_only`; no alias reader was added. Corrected fixtures preserve their
original NFC and special-character assertions under the valid partition.

This slice does not claim complete canonical ExportModel normalization. The
remaining record/source-index representation, full post-redaction public
model, template expansion/narratives/assets, graph subject/disclosure behavior,
timed workloads, browser consolidation, qualified cost portfolio and final
human fidelity review remain required. Reporting 1.4.0 and Composition 1.3.0
remain proposed pending coordinated adoption evidence.

| Obligation / owner | Exercising identity | Assertion / context |
| --- | --- | --- |
| RPT-087f redaction, partitions and roles / Reporting | `TestCompositionRenderBinding_Unit/presentation_text_redaction` in existing metadata row | Actual render pipeline and pure replay binding; retained mask/drop/allow, path precedence, token and display placeholders, forbidden partitions, substitution/replacement bounds, immutable input and changed-body rejection |
| RPT-039 admission shape / Reporting and Extensions | `TestReportingAdmittedModelIdentityAndTime` and existing participant row | Retired schema, absent/null pool and prepopulated pool reject; participant output remains exact admitted bytes |
| RC-052 partition grammar / Composition | `TestCompositionRenderSourceAdmission_Unit` in `module.reportcomposition.unit.canonical_digests` | Closed tokens, Unicode scalar bounds and forbidden delimiters/whitespace; no inferred Party existence |
| Clean cutover / Reporting and database lifecycle | `TestReportingIdentityCutoverDoesNotReconstructHistory_Integration/82` | Reject the new migration with retained source history; data and migration head remain unchanged |

All following roots are under `.cartulary/test-results/`.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-authored-redaction-format-20261015-01` | `make format` | Outer 0, 5.846 s. |
| `remediation-authored-redaction-generate-20261015-01` | `make generate` | Outer 2, `harness` / `unknown_failure`; migration 83 lacked its authored source-owner assignment. No acceptance claimed. |
| `remediation-authored-redaction-generate-20261015-02` | `make generate` | Outer 2, `harness` / `unknown_failure`; changed participant specialization had a stale authored contribution digest. |
| `remediation-authored-redaction-generate-20261015-03` | `make generate` | Outer 0, 10.170 s after both authored inputs were corrected. |
| `remediation-authored-redaction-format-20261015-02` | `make format` | Outer 0, 5.813 s. |
| `remediation-authored-redaction-unit-20261015-01` | Metadata, frozen-release and admitted participant rows | Outer 0, 4.704 s; initial authored redaction and replay cases. |
| `remediation-authored-redaction-format-20261015-03` | `make format` | Outer 0, 5.929 s; additional boundary cases and strict Composition grammar. |
| `remediation-authored-redaction-unit-20261015-02` | Same three Reporting unit rows | Outer 0, 4.722 s. |
| `remediation-authored-redaction-composition-20261015-01` | Composition canonical-digest row | Outer 0, 1.400 s; strict authored-partition grammar. |
| `remediation-authored-redaction-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.991 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-authored-redaction-drift-20261015-01` | `make generate-drift` | Outer 0, 16.980 s. |
| `remediation-authored-redaction-reporting-20261015-01` | Full Reporting owner | Outer 0, 89.312 s; native authored-title and mixed diagrams, migration 83, completion and recovery. This precedes the bounded-substitution and frozen-path follow-up. |
| `remediation-authored-redaction-composition-20261015-02` | Full Composition owner | Outer 2, 24.012 s; `product` / `test_assertion_failure`. The missing-input recovery fixture attempted to restore a row without its independently retained graph-input digest. PostgreSQL rejected it with `reporting_graph_input_v1_ck`; all 18 cleanup steps completed. The constraint was preserved and the fixture now captures/restores that exact digest. |
| `remediation-authored-redaction-format-20261015-04` | `make format` | Outer 0, 6.005 s. |
| `remediation-authored-redaction-unit-20261015-03` | Same three Reporting unit rows | Outer 0, 7.092 s; bounded placeholder allocation and frozen-profile authored-path validation. |
| `remediation-authored-redaction-preview-20261015-01` | Exact Composition preview-delegation row | Outer 0, 28.899 s after restoring the complete retained input. No failure was reclassified as intermittent. |
| `remediation-authored-redaction-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.363 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-authored-redaction-reporting-20261015-02` | Full Reporting owner | Outer 0, 110.045 s; includes bounded substitution, exact authored paths, native PDF and recovery. |
| `remediation-authored-redaction-composition-20261015-03` | Full Composition owner | Outer 0, 35.258 s. |
| `remediation-authored-redaction-contracts-20261015-01` | Attempted mixed-owner selection | Outer 2 before execution, `config` / `usage_error`: `OWNER` is required. Rerouted to the two owner commands below. |
| `remediation-authored-redaction-extensions-20261015-01` | Extensions coordinator-registry and cross-owner descriptor rows | Outer 0, 50.482 s. |
| `remediation-authored-redaction-contractgen-20261015-01` | Contract decoder row under `harness.generated_artifacts` | Outer 0, 50.391 s. |
| `remediation-authored-redaction-lint-20261015-01` | `make lint` | Outer 2, 37.247 s; `harness` / `child_target_failure`, 10/11 units passed and all 20 cleanup steps completed. Go lint reported unused `assignDisplayTokens` wrapper (`U1000`). Removed the wrapper and retained one typed function. |
| `remediation-authored-redaction-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.928 s. |
| `remediation-authored-redaction-shapes-20261015-01` | `make json-shape-check` | Outer 0, 5.331 s. |
| `remediation-authored-redaction-reporting-audit-20261015-01` | Reporting evidence audit, explicit `.cartulary/evidence-inputs/remediation-authored-redaction-reporting-20261015-01.json` | Outer 0, 2.780 s; full Reporting owner root above. |
| `remediation-authored-redaction-composition-audit-20261015-01` | Composition evidence audit, explicit `.cartulary/evidence-inputs/remediation-authored-redaction-composition-20261015-01.json` | Outer 0, 2.764 s; full Composition owner root above. |
| `remediation-authored-redaction-database-20261015-01` | Full database-migrations owner | Outer 0, 43.104 s. |
| `remediation-authored-redaction-finalize-20261015-03` | `make agent-finalize` after wrapper cleanup | Outer 0, 20.356 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-authored-redaction-lint-20261015-02` | `make lint` after wrapper cleanup | Outer 0, 29.304 s. |

The two owner audits bind the pre-wrapper-cleanup source checkpoint, not later
source-index work. They establish the authored-text slice's evidence and do not
qualify a later model or renderer contract automatically.

## Canonical source-reference admission

Reporting now materializes the Table 9-H source index from explicit source-owner
bindings. Transient selectors never survive in public reference arrays. Field
references use a snapshot-bound generated ID and the exact keyed address in the
admitted snapshot. Record-level references use an explicit null address.
Sections, blocks, records, Timeline events, subjects, relationship endpoints and
source-backed support references share those identities. Support partitions
preserve the target's ordering and restrictions. Incident provenance uses the
supplied incident identity, including when a source selector has no incident
prefix. Readers validate this index; they do not reconstruct or upgrade it.

Snapshot schema v3, ExportModel v4, renderer contract 13 and migration 84 retire
selector-based retained inputs. The migration repeats the no-history guard,
preserves source-fact and presentation-admission constraints, and advances the
preflight marker. The participant budget includes the source index. No existing
environment was reset. The public record-envelope/type normalization and full
post-redaction ExportModel remain open; this slice does not misclassify the
existing generic envelope contributions as fully normalized Core records.

| Obligation | Exercising identity | Assertion and evidence context |
| --- | --- | --- |
| RPT-046d / 050 source identity | `TestSnapshotRecordMaterialization/canonical_source_index`, existing materialization row | Independent SHA-256 vector; private selector exclusion; explicit incident binding; complete field/record/support references; malformed, cross-snapshot, orphaned, altered and unknown-member rejection |
| RPT-044a / 046b source-backed support | Same entrypoint and existing support/reference-disclosure tests | Exact target binding, preserved disclosure order, no source-path inference or raw summary fallback |
| Frozen admission and replay | Existing admitted-participant and frozen-release rows | Changed source index rejects; participant bounds include indexed source objects; snapshot/render identities remain distinct |
| Clean cutover | Existing `TestReportingIdentityCutoverDoesNotReconstructHistory_Integration/83` | Retained history rejects migration 84 before data or migration-state changes |

Test-only selector aliases are supplied by fixture authors to check their
expected values; no production reader retains those aliases. The diagnostic
traversal fixture explicitly re-materializes its re-authored addresses, keeping
its intended defects constant. This is distinct from replay, which rejects any
changed admitted address.

All following roots are under `.cartulary/test-results/`.

| Run | Command / scope | Observed result |
| --- | --- | --- |
| `remediation-source-index-format-20261015-01` | `make format` | Outer 0, 6.077 s. |
| `remediation-source-index-generate-20261015-01` | `make generate` | Outer 0, 10.080 s; schema declarations and migration ownership refreshed. |
| `remediation-source-index-unit-20261015-01` | Materialization, metadata and redaction rows | Outer 2, 3.498 s; `product` / `test_assertion_failure`, all 10 cleanup steps completed. Old selector assertions and omitted subject-reference conversion were exposed. |
| `remediation-source-index-format-20261015-02` | `make format` | Outer 0, 5.932 s. |
| `remediation-source-index-format-20261015-03` | `make format` | Outer 0, 6.005 s. |
| `remediation-source-index-unit-20261015-02` | Same three rows | Outer 2, 3.456 s; same classification and complete cleanup. Fixed partition copying to preserve normative order and limited incident fallback to absent incident fields. |
| `remediation-source-index-format-20261015-04` | `make format` | Outer 0, 5.948 s. |
| `remediation-source-index-unit-20261015-03` | Same three rows | Outer 2, 3.036 s; same classification and complete cleanup. Cross-snapshot references incorrectly received a generated-ID mismatch first; validation now separates source identity from containing-snapshot equality. |
| `remediation-source-index-format-20261015-05` | `make format` | Outer 0, 6.002 s. |
| `remediation-source-index-unit-20261015-04` | Three preceding rows plus admitted-participant and frozen-release rows | Outer 0, 3.550 s; all targeted source-index and existing redaction/composition checks passed. |
| `remediation-source-index-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.160 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |


### Source-index integration and native PDF investigation

The first complete source-index owner run exposed three independent failures:
a source-backed support error had lost its existing `invalid_support_ref`
classification; integration fixtures still looked up retired selector strings;
and fresh native mixed-diagram renders produced unequal PDFs. Composition's
relational-only fixture also retained the retired snapshot schema. The support
classification and Composition fixture are corrected, with focused passes.
Integration fixture labels now resolve through explicit source-index ownership;
their authored incident scalar list is fixture input, not a production reader.

The PDF discrepancy is not closed by a later pass. Identical frozen mixed manual
and automatic diagram input produced equal bundle members except `deck.pdf`.
The repeated differing hashes were
`00a06b1b1a57883f54da51343798a595ace27ad1e7f739865a7a2063dcd4152f` and
`bd6794faeb344a1a5f5efb02c0713ae62ccf5606e56f4172e7d222695ed7a5a6`.
The first difference is a PDF page-parent object reference, so timestamp
normalization alone does not explain it. A missing Mermaid completion barrier
was repaired and its real browser regression passed, but native differences
persisted afterward. That repair is not claimed as the incident's cause or
closure. Bounded diagnostics inspect only this fixed synthetic fixture.
No action retry, deadline slack, output rebaselining or disabled assertion was
introduced. All attempts use the installed sandbox and clean attempt directories.

The following run IDs remain under `.cartulary/test-results/`. Commands use the
owner and exact row names already recorded in this section. No failed run is
qualified evidence for adoption.

| Run | Command / selection | Result |
| --- | --- | --- |
| `remediation-source-index-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.240 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-source-index-reporting-20261015-01` | Full Reporting owner | Outer 2, 84.598 s; `product` / `test_assertion_failure`, 22/24 work units passed, all 28 cleanup steps completed. Support classification, integration fixture selectors and native mixed PDF failed. |
| `remediation-source-index-composition-20261015-01` | Full Composition owner | Outer 2, 35.720 s; same classification, 4/5 units passed, all 18 cleanup steps completed. Retired snapshot schema in a relational fixture. |
| `remediation-source-index-lint-20261015-01` | `make lint` | Outer 2, 30.226 s; `harness` / `child_target_failure`, 10/11 units passed, all 20 cleanup steps completed. Two Go test files required formatting. |
| `remediation-source-index-markdown-20261015-01` | `make lint-markdown` | Outer 0, 39.460 s; earlier documentation revision. |
| `remediation-source-index-format-20261015-06` | `make format` | Outer 0, 5.925 s. |
| `remediation-source-index-native-diagnosis-20261015-01` | Installed native renderer row | Outer 2, 70.398 s; product assertion, 3/4 units passed, all 13 cleanup steps completed. Isolated unequal PDF. |
| `remediation-source-index-format-20261015-07` | `make format` | Outer 0, 5.918 s. |
| `remediation-source-index-render-repair-20261015-01` | Installed renderer, Slidev, Evidence-provider and snapshot integration rows | Outer 2, 67.625 s; product assertion, 6/8 units passed, all 19 cleanup steps completed. Slidev completion and support classification passed; native PDF and integration selectors still failed. |
| `remediation-source-index-composition-repair-20261015-01` | `module.reportcomposition.integration.resolve_release_tuple_tx_reason_matrix_ad31a02455` | Outer 0, 16.198 s. |
| `remediation-source-index-format-20261015-08` | `make format` | Outer 0, 6.017 s. |
| `remediation-source-index-native-diagnosis-20261015-02` | Installed native renderer row | Terminal failed summary, 69.730 s; product assertion, 3/4 units passed, all 13 cleanup steps completed. Native session exit unavailable after context restoration; no successful exit inferred. First differing PDF object reference retained. |
| `remediation-source-index-format-20261015-09` | `make format` | Outer 0, 6.048 s. |
| `remediation-source-index-format-20261015-10` | `make format` | Outer 0, 5.986 s. |
| `remediation-source-index-native-diagnosis-20261015-03` | Installed renderer and snapshot integration rows | Outer 2, 71.664 s; product assertion, 6/7 units passed, all 18 cleanup steps completed. Native repeated PDFs passed; integration fixture incorrectly expected six incident fields instead of the five it authored. |
| `remediation-source-index-format-20261015-11` | `make format` | Outer 0, 5.937 s; corrects the fixture and adds four incident-specific native observations, with immediate failure on unequal bytes. |

The original full owner source was
`sha256:ea8d1bdf7f212089f71354093b14774d537a912834bd28a45ac5dbc16a2309bb`.
Diagnostic run 02 used
`sha256:cc9c29999d2d137e40d995ce553d1fa247d57fbe6124bf1c27ffaebd59d8cd53`;
run 03 used
`sha256:645bded52ba0536335046816fb92662a43061aeab032f4470abf68878ec95e53`.
These are distinct dirty source identities. Their diagnostics distinguish actual
fresh render outcomes from test-source changes; no aggregate failure rate or
whole-owner current closure is inferred.


The source-index checkpoint completed on
`sha256:dd21aad9c787407a0b9ffc6672ac53b16af0d40a11cbe85cdfdca798e017e3a5`.
All run roots below remain under `.cartulary/test-results/`.

| Run | Command / scope | Result |
| --- | --- | --- |
| `remediation-source-index-native-diagnosis-20261015-04` | Installed renderer plus snapshot integration rows | Outer 0, 83.210 s; corrected five-field fixture and four fresh native mixed-diagram attempts. |
| `remediation-source-index-native-diagnosis-20261015-05` | Installed renderer row alone | Outer 0, 81.809 s; four fresh mixed-diagram attempts. The source is unchanged from run 04, with a different selection context. |
| `remediation-source-index-finalize-20261015-03` | `make agent-finalize` | Outer 0, 20.366 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-source-index-reporting-20261015-02` | Full Reporting owner | Outer 0, 109.706 s; 24/24 units, cache split 8 bypass / 3 hits / 13 misses. Cached units are not fresh executions. |
| `remediation-source-index-composition-20261015-02` | Full Composition owner | Outer 0, 41.125 s. |
| `remediation-source-index-lint-20261015-02` | `make lint` | Outer 0, 40.250 s. |
| `remediation-source-index-drift-20261015-01` | `make generate-drift` | Outer 0, 29.621 s. |
| `remediation-source-index-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.953 s. |
| `remediation-source-index-shapes-20261015-01` | `make json-shape-check` | Outer 0, 5.326 s. |
| `remediation-source-index-reporting-audit-20261015-01` | Reporting evidence audit | Outer 0, 2.710 s; explicit `.cartulary/evidence-inputs/remediation-source-index-module.reporting-20261015-01.json`, containing the full owner root above. |
| `remediation-source-index-composition-audit-20261015-01` | Composition evidence audit | Outer 0, 2.697 s; explicit `.cartulary/evidence-inputs/remediation-source-index-module.reportcomposition-20261015-01.json`, containing its full owner root above. |
| `remediation-source-index-database-20261015-01` | Full database-migrations owner | Outer 0, 43.051 s. |
| `remediation-source-index-extensions-20261015-01` | Extensions coordinator-registry and cross-owner descriptor rows | Outer 0, 2.116 s. |
| `remediation-source-index-contractgen-20261015-01` | Contract decoder row under generated-artifact owner | Outer 0, 0.856 s. |

These passes validate the exercised source-index slice. They do not explain or
repair the earlier PDF intermittency, qualify later renderer-profile changes,
or close coordinated adoption. Current Core record normalization, public
post-redaction model fidelity, template semantics, timed assurance, browser
consolidation and qualified performance portfolio remain required.


## Shared native Mermaid render profile

Renderer contract 14 gives standalone SVG and Slidev-embedded Mermaid one
explicit profile. Both use strict security, SVG text labels, local sans-serif
fonts, the same layout choices, deterministic IDs and the fixed renderer root
identity. Slidev's supported setup boundary receives only renderer-authored code
and installed module identity; no case text becomes JavaScript or overrides
configuration. The setup uses the pinned browser ESM bundle, not the package's
general entrypoint. Source Markdown and Mermaid serializer identities remain
unchanged. Earlier rendered-byte evidence does not qualify this new profile.

The print barrier awaits completed diagram SVG before font/image readiness and
two paint frames. A renderer-owned failure signal handles Slidev's caught
component failures: a failed diagram rejects the entire export before PDF
capture and closes its browser context. Existing browser interception and native
network mediation remain independent enforcement layers.

The same existing Slidev characterization now observes actual SVG text,
identifier, absence of foreign objects, completed headings, real PDF page
contents and byte equality. A deliberately invalid diagram verifies failure
before the PDF spy can run. Test assertions execute outside the adapter's
production error-normalization boundary; their original assertion details are
preserved. No executable test identity was removed or added.

This is a structural correction to inconsistent renderer configuration. The
PDF discrepancy remains unattributed; passing checks neither explain the cause
nor prove universal stability. The complete Reporting owner subsequently
reproduced the same two PDF digests under renderer contract 14. Inspection of
the installed runtime identity and its setup source confirmed the new profile,
ruling out stale packaging for that run. All failed roots remain retained. A
bounded PDF-object diagnostic probe is temporarily enabled only for the fixed
synthetic native fixture to distinguish resource allocation from incomplete
rendered content. This incident remains open.

| Run under `.cartulary/test-results/` | Command / scope | Result |
| --- | --- | --- |
| `remediation-mermaid-profile-format-20261015-01` | `make format` | Outer 0, 6.126 s. |
| `remediation-mermaid-profile-browser-20261015-01` | Slidev and Mermaid browser rows | Outer 2, 11.440 s; product assertion, 3/4 units passed, all 13 cleanup steps complete. General Mermaid entrypoint exposed a CommonJS dependency to the browser. |
| `remediation-mermaid-profile-format-20261015-02` | `make format` | Outer 0, 5.724 s. |
| `remediation-mermaid-profile-browser-20261015-02` | Same two rows | Outer 2, 10.715 s; product assertion, 3/4 units passed, all 13 cleanup steps complete. ESM loading repaired; the old HTML-label selector did not describe SVG text. |
| `remediation-mermaid-profile-format-20261015-03` | `make format` | Outer 0, 5.862 s. |
| `remediation-mermaid-profile-browser-20261015-03` | Slidev row | Outer 2, 13.178 s; product assertion, 2/3 units passed, all 12 cleanup steps complete. Preserved diagram observations for diagnosis outside PDF capture. |
| `remediation-mermaid-profile-format-20261015-04` | `make format` | Outer 0, 5.796 s. |
| `remediation-mermaid-profile-browser-20261015-04` | Slidev row | Outer 0, 13.256 s; standard SVG text assertions and original assertion diagnostics. |
| `remediation-mermaid-profile-format-20261015-05` | `make format` | Outer 0, 5.752 s; adds component-failure rejection. |
| `remediation-mermaid-profile-native-20261015-01` | Installed native, Go stage execution, Slidev, Mermaid and browser-policy rows | Outer 0, 79.598 s; four fresh mixed-diagram native renders and the failed-diagram browser case. |
| `remediation-mermaid-profile-format-20261015-06` | `make format` | Outer 0, 6.218 s. |
| `remediation-mermaid-profile-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.154 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-mermaid-profile-composition-20261015-01` | Full Composition owner | Outer 0, 40.516 s. |
| `remediation-mermaid-profile-lint-20261015-01` | `make lint` | Outer 0, 32.866 s. |
| `remediation-mermaid-profile-reporting-20261015-01` | Full Reporting owner | Outer 2, 101.976 s; product assertion, 23/24 units passed, all 33 cleanup steps complete. Mixed native diagrams reproduced the same unequal PDF digest pair. |
| `remediation-mermaid-profile-format-20261015-07` | `make format` | Outer 0, 6.001 s; restored bounded synthetic diagnostics and eight incident-specific fresh attempts, failing on the first unequal pair. |
| `remediation-mermaid-profile-native-diagnosis-20261015-01` | Installed native renderer row | Terminal passed summary, 108.340 s, all four units passed and all 13 cleanup steps complete. Native session exit unavailable after context restoration; no outer successful exit inferred. |
| `remediation-mermaid-profile-reporting-diagnosis-20261015-01` | Full Reporting owner, unchanged executable source from the preceding diagnostic | Outer 2, 88.727 s; product assertion, 23/24 units passed, all 28 cleanup steps complete. PDF-object diagnostics captured a Liberation Mono Bold resource in one PDF and different Liberation Sans usage in the other. |

## Declared local asset rendering and failed-slide rejection

Targeted reproduction added a manual local SVG to the existing real-browser
Slidev characterization. The generated relative image reference was interpreted
as a module import relative to a virtual slide, while the admitted asset lived
in the attempt's public asset directory. The browser received a 404 and Slidev
caught its failed slide import, displaying its error layout. The adapter's
`pageerror` listener did not observe that caught failure. Page count, metadata,
and even repeated PDF equality therefore did not establish successful slide
content for this input. The error layout explicitly uses a monospace bold font,
consistent with the additional font resource in the native incident.

Renderer contract 15 keeps canonical bundle paths unchanged and maps admitted
image references to the same attempt's local HTTP resource. Vue asset rewriting
is disabled: the first repair demonstrated that Vite could otherwise inline the
same asset as a `data:` URL in one render and leave it as a public URL in another.
The image renderer checks exact declared membership before emitting the local
URL. Failed HTTP resources and the pinned Slidev loader's caught-error signal
reject before capture. Native diagnostics retain only closed classes, without
URLs, input bytes or vendor messages. Ordinary browser policy and native
confinement remain independent checks.

The existing browser characterization now checks decoded local image dimensions,
completed headings, the finished Mermaid SVG and complete independently rendered
PDF equality. An injected 404 must fail before the PDF spy runs. Generic console
errors are not treated as slide failures: an unrelated pinned Twoslash authoring
plugin emits a caught setup error even when Twoslash source is disabled. The
specific loader diagnostic and actual failed resources define the render failure
boundary. No retry, deadline relaxation or PDF-byte normalization was added to
hide the incident. Native qualification and integrated verification remain
required before closing it.

| Run under `.cartulary/test-results/` | Command / scope | Result |
| --- | --- | --- |
| `remediation-slide-import-format-20261015-01` | `make format` | Outer 0, 5.773 s. |
| `remediation-slide-import-native-20261015-01` | Installed renderer and Slidev browser rows | Outer 2, 47.753 s; product assertion, 3/5 units passed, all 14 cleanup steps complete. Exposed missing native resources and the overly broad console-error check. |
| `remediation-slide-import-format-20261015-02` | `make format` | Outer 0, 5.790 s. |
| `remediation-slide-import-browser-20261015-01` | Slidev browser row | Outer 2, 9.783 s; product assertion, all 12 cleanup steps complete. Local image deterministically reproduced a 404 and caught slide-import failure. |
| `remediation-slide-import-format-20261015-03` | `make format` | Outer 0, 5.922 s. |
| `remediation-slide-import-browser-20261015-02` | Slidev browser row | Outer 2, 13.163 s; product assertion, all 12 cleanup steps complete. Decoded image present but exact URL assertion failed. |
| `remediation-slide-import-browser-20261015-03` | Same row, bounded image diagnostics | Outer 2, 13.711 s; product assertion, all 12 cleanup steps complete. Exposed `data:` inlining versus the local URL across independent renders. |
| `remediation-slide-import-format-20261015-04` | `make format` | Outer 0, 5.967 s. |
| `remediation-slide-import-browser-20261015-04` | Slidev browser row | Outer 0, 16.525 s; decoded assets, completed diagram/headings, repeatable bytes, invalid diagram and injected missing asset all passed. |
| `remediation-slide-import-native-20261015-02` | Installed renderer, process isolation, stage execution and PDF admission rows | Outer 2, 57.036 s; product assertion, 6/7 units passed, all 16 cleanup steps complete. Only the native mixed-diagram candidate failed, now with a caught browser import error rather than an accepted PDF. |
| `remediation-slide-import-format-20261015-05` | `make format` | Outer 2; configuration preflight, no run summary. The existing catalog incorrectly restricted installed preparation to Go rows. No formatting pass claimed. |
| `remediation-slide-import-generate-20261015-01` | `make generate` | Outer 2; tool summary failed after 8.890 s, installed-preparation context rejection. |
| `remediation-slide-import-generate-20261015-02` | `make generate` | Outer 0, 10.100 s; admits only exclusive `renderer-node` integration as the additional installed browser context. |
| `remediation-slide-import-format-20261015-06` | `make format` | Outer 0, 5.794 s. |
| `remediation-slide-import-packaged-browser-20261015-01` | Slidev row, installed adapter | Outer 0, 44.555 s including package preparation; the initial two-page content workload passed outside native confinement. |
| `remediation-slide-import-format-20261015-07` | `make format` | Outer 0, 5.942 s; adds closed module-failure diagnostics. |
| `remediation-slide-import-generate-20261015-03` | `make generate` | Outer 0, 10.390 s; updates the exact characterization title and three-page workload. |
| `remediation-slide-import-format-20261015-08` | `make format` | Outer 0, 5.770 s. |
| `remediation-slide-import-packaged-browser-20261015-02` | Installed native and packaged browser rows | Outer 2, 74.365 s; product assertion, 4/5 units passed, all 14 cleanup steps complete. The packaged three-page case passed; native execution classified the remaining failure as a missing module export. |
| `remediation-slide-import-packaged-browser-20261015-03` | Packaged browser row with embedded entrypoint shape | Outer 2, 36.224 s; product assertion, 3/4 units passed, all 13 cleanup steps complete. Deterministically reproduced `lz-string` missing its default export. |
| `remediation-slide-import-format-20261015-09` | `make format` | Outer 0, 5.845 s; explicit installed decoder preparation and caught-import failure injection. |

### Embedded entrypoint dependency preparation

The remaining native failure was reproduced outside confinement by giving the
packaged characterization the same embedded entrypoint shape. Slidev infers
global/local package preparation from `process.argv[1]`. A CLI path under
`node_modules` prepares its CommonJS `lz-string` decoder correctly; the native
`/runtime/renderer/src/stage-main.ts` entrypoint does not have that shape. Its
local nested-package optimizer selection cannot resolve from the isolated work
directory. The browser therefore received raw CommonJS and rejected the default
import needed by the Mermaid component. The fixed synthetic diagnostic names
that precise module in run `remediation-slide-import-packaged-browser-20261015-03`.

The adapter now resolves `lz-string` from the installed Slidev client and
explicitly configures its browser interop preparation. It neither falsifies the
production entrypoint nor adds a package-resolution fallback. The characterization
retains the embedded-entrypoint regression, restores the test process state at
test completion, uses the production browser argument profile, and checks three
completed slides with the manual image and automatic diagram on separate pages.
An injected invalid slide module verifies that Slidev's caught import error fails
before PDF capture even with an HTTP 200 response. These checks complement native
isolation; they cannot replace it. Source inventory records the exact renamed
test and preserved assertions.

### Renderer contract 15 repair validation

The repaired installed adapter passed both the packaged three-page content
characterization and eight independent native mixed-diagram attempts in
`remediation-slide-import-packaged-browser-20261015-04` (outer 0, 118.276 s,
5/5 units). The source identity was
`sha256:b53fd07c5a96feaff12c89454e5008022ad0be1e30334141533adcd4e7b79eee`;
the graph identity was
`sha256:2c4293d813d8053ff6f6920b825a5d94242597a3dfe146e664cd5c1e9af2e6e6`.
The 80.693 s queue interval reflects resource admission, not another additive
execution charge. This is a mechanism-specific repair result, not a statistical
claim that every future render is stable.

The temporary PDF-object probe was then removed and the native regression
returned to its ordinary two independent attempts. Safe-diagnostic tests cover
every new closed reason and exclude appended vendor/input text. The exact
post-repair validation roots are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-slide-import-format-20261015-10` | `make format` | Outer 0, 6.108 s. |
| `remediation-slide-import-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.782 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-slide-import-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 79.686 s; complete owner selection. |
| `remediation-slide-import-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 23.421 s; complete owner selection. |
| `remediation-slide-import-catalog-20261015-01` | `make test-slice OWNER=harness.test_catalog` | Outer 0, 31.442 s; complete owner selection. |
| `remediation-slide-import-lint-20261015-01` | `make lint` | Outer 0, 27.748 s. |
| `remediation-slide-import-drift-20261015-01` | `make generate-drift` | Outer 0, 20.265 s. |
| `remediation-slide-import-shapes-20261015-01` | `make json-shape-check` | Outer 0, 6.127 s. |
| `remediation-slide-import-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 0.975 s. |
| `remediation-slide-import-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-slide-import-module.reporting-20261015-01.json` | Outer 0, 2.876 s; exact compatible full-owner root. |
| `remediation-slide-import-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-slide-import-module.reportcomposition-20261015-01.json` | Outer 0, 2.838 s; exact compatible full-owner root. |
| `remediation-slide-import-markdown-20261015-01` | `make lint-markdown` | Outer 0, 9.600 s. |

The native PDF incident has a reproduced cause, retained failing evidence,
content-sensitive regressions and fresh execution in its triggering context.
It is closed for this reviewed workload. Earlier failures remain part of the
record; neither cached success nor a finite pass streak supplied the diagnosis.
Reporting/Composition adoption remains open for the independent model,
composition, workload and qualification obligations listed above.

## Core record and auxiliary source separation

Table 9-E requires Core envelope identities and Core record types. The old
projection emitted an additional record summary for each generic envelope,
mention, link and tag, and Artifact providers declared their subtype as the Core
type. Candidate REQ-RPT-043b now defines one summary per Core ID, preserves
source-family provenance, and selects the domain field projection when present.
An envelope-only Core record retains its explicit metadata projection. Fields
from the two projections are not merged by spelling: their row versions and
timestamps can have different meanings.

The implementation has a private source-group view reconstructed from section
fields and exact source-index bindings. It is not another wire inventory or
execution catalog. Subject, disclosure, support and redaction processing retain
auxiliary source groups even though those groups no longer appear in
`records[]`. Core summaries must match the complete frozen projection;
missing, duplicate, invented and altered summaries reject before rendering.
Auxiliary IDs do not share disclosure nodes with an unrelated Core ID.
Display labels are derived from the complete source group. No selector spelling
or current database lookup supplies provenance during replay.

Snapshot model v4 and render-input ExportModel v5 replace v3/v4. The owner
participant declaration and digest change with that wire identity; downstream
projections are regenerated through Make. Migration 85 rejects retained
snapshot/Job/import/preview state before mutation and forbids downgrade.
The cutover rehearsal includes predecessor 84. This does not reset any operator
environment, add a legacy reader or invent historical record identities.

Existing `TestSnapshotRecordMaterialization` and its catalog row now include
`source_groups_test.go` as a helper, covering Core/auxiliary separation,
envelope/domain version isolation, Artifact typing, ID collisions, missing and
extra summaries, frozen display tampering and conflicting envelope types.
The source-provider integration scenario compares the actual snapshot Core set
to its dedicated fixture's active envelopes. Records' existing provider test
now requires its declared Core type to match the source-owned type value.
The earlier subject-family test keeps distinct-family identity coverage with
distinct Core IDs and additionally rejects one envelope claiming two types.
No test entrypoint or execution context was removed.

The initial failed roots remain diagnostic evidence, not successful qualification:

| Run under `.cartulary/test-results/` | Result and disposition |
| --- | --- |
| `remediation-core-record-generate-20261015-01` | Outer 2, 1.840 s, harness/unknown_failure. Generation identified the missing migration owner allocation; the authored allocation now names Reporting. |
| `remediation-core-record-generate-20261015-02` | Outer 2, 4.460 s, harness/unknown_failure. Generation rejected the stale authored participant digest; its replacement was independently calculated from the exact canonical contract plus LF. |
| `remediation-core-record-generate-20261015-03` | Outer 0, 10.200 s; downstream projections regenerated. |
| `remediation-core-record-format-20261015-01` | Outer 0, 5.901 s. |
| `remediation-core-record-model-20261015-01` | Outer 2, 0.872 s, infra/preflight_error; a new test used the copy helper with an unsupported Go type. No selected test execution claimed. All 10 cleanup steps passed. |
| `remediation-core-record-format-20261015-02` | Outer 0, 6.298 s. |
| `remediation-core-record-model-20261015-02` | Outer 2, 3.668 s, product/test_assertion_failure. Revealed stale generic-type/path fixtures and error-boundary precedence. All 10 cleanup steps passed. |
| `remediation-core-record-format-20261015-03` | Outer 0, 6.042 s. |
| `remediation-core-record-model-20261015-03` | Outer 2, 3.393 s, product/test_assertion_failure. Remaining failures concerned dependent snapshot diagnostics and the new explicit envelope-type fixture field. All 10 cleanup steps passed. |
| `remediation-core-record-format-20261015-04` | Outer 0, 6.051 s. |
| `remediation-core-record-model-20261015-04` | Outer 0, 7.423 s; five selected model/redaction/admission rows. |
| `remediation-core-record-providers-20261015-01` | Outer 0, 54.580 s; source-provider, replay and cutover rows, including predecessor 84. |
| `remediation-core-record-envelope-20261015-01` | Outer 0, 15.815 s; Records provider contract. |
| `remediation-core-record-format-20261015-05` | Outer 0, 6.059 s; removes obsolete type fallback and adds valid-source missing-display denial. |
| `remediation-core-record-finalize-20261015-01` | Outer 0, 20.161 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-core-record-reporting-20261015-01` | Outer 0, 104.870 s; full Reporting owner, 24/24 units, all 28 cleanup steps complete. |
| `remediation-core-record-composition-20261015-01` | Outer 0, 45.361 s; full Composition owner. |
| `remediation-core-record-database-20261015-01` | Outer 0, 62.279 s; full database-migrations owner. |
| `remediation-core-record-lint-20261015-01` | Outer 0, 40.682 s. |
| `remediation-core-record-drift-20261015-01` | Outer 0, 14.951 s; generated artifact drift. |
| `remediation-core-record-shapes-20261015-01` | Outer 0, 5.209 s; JSON shapes. |
| `remediation-core-record-policy-20261015-01` | Outer 0, 0.930 s; generated artifact policy. |
| `remediation-core-record-reporting-audit-20261015-01` | Outer 0, 2.729 s; exact full-owner root via `.cartulary/evidence-inputs/remediation-core-record-module.reporting-20261015-01.json`. |
| `remediation-core-record-composition-audit-20261015-01` | Outer 0, 2.693 s; exact full-owner root via `.cartulary/evidence-inputs/remediation-core-record-module.reportcomposition-20261015-01.json`. |
| `remediation-core-record-extensions-20261015-01` | Outer 0, 1.863 s; coordinator registry and cross-owner descriptors. |
| `remediation-core-record-contractgen-20261015-01` | Outer 0, 0.860 s; contract decoder/validation row. |

The full Reporting run binds source
`sha256:8b3e519fdb4876bfad7df949c863d384957ca80d5a278b9f124ba5216d9cb00e`
and graph
`sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
These qualified owner results close this bounded source-normalization slice.
Concurrent owner invocations are functional validation, not comparable
performance-window observations.

Complete post-redaction public ExportModel construction, record-qualified
composition anchors, template expansion, lifecycle eligibility and the remaining
adoption obligations are separate work. This source-normalization slice does
not claim coordinated adoption or all-nine closure.

## Retained Core records and composition anchors

The retained projection now rebuilds Core summaries from selected field values.
Removed summaries cannot be restored through generic envelopes or auxiliary
sources, labels follow masking and removal, and ordinals and reference aggregates
are recomputed. The source-family display rule is shared with snapshot assembly;
the admitted summary is never a fallback for a missing retained display value.

Record-qualified composition anchors previously rejected every record target.
They now require one retained Core summary and resolve derived blocks through
explicit frozen source bindings. Nested matching examines the current field tree
after earlier operations, excluding authored insertion and removed children.
Shared auxiliary IDs and cached container reference unions cannot manufacture a
target. Several matching blocks remain ambiguous. This is an internal projection
and lookup change; no wire schema or historical reader was added.

The existing `TestRedactionProfilePrecedenceActionsAndManifest_Unit` and
`TestCompositionRenderBinding_Unit` entrypoints invoke the retained-record and
record-anchor helpers. Their assertions cover display masking/removal, empty
summary removal, envelope and auxiliary isolation, cancellation, source
immutability, digest-bound wire operations, zero and multiple matches, nested
membership, sequential removal and non-semantic target rejection. No test or
execution context was retired. The complete canonical post-redaction model,
lifecycle eligibility and template expansion remain distinct open obligations.

Validation roots are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-record-anchor-format-20261015-01` | `make format` | Outer 0, 6.245 s. |
| `remediation-record-anchor-model-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | Outer 0, 3.514 s; selected source-contract and redaction scenarios. |
| `remediation-record-anchor-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.960 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-record-anchor-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 43.158 s; full owner selection. |
| `remediation-record-anchor-lint-20261015-01` | `make lint` | Outer 0, 35.757 s. |
| `remediation-record-anchor-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 94.462 s; 24/24 units, including permitted cached evidence and fresh affected/native execution. |
| `remediation-record-anchor-drift-20261015-01` | `make generate-drift` | Outer 0, 13.477 s. |
| `remediation-record-anchor-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-record-anchor-module.reporting-20261015-01.json` | Outer 0, 2.673 s. |
| `remediation-record-anchor-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-record-anchor-module.reportcomposition-20261015-01.json` | Outer 0, 2.656 s. |

The Reporting invocation binds source
`sha256:69a29def3763bba3c67ad930cfe8b983ed4ba8c9d76af99a33a2d2d34ca0ec45`
and graph
`sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
These are functional owner results, not qualified performance-window samples.

## Canonical retained model projection

The private flat `redacted_export_model.v3` byte form is removed. The retained
wrapper now embeds the canonical ExportModel v5 shape. Its private decision
index, pre-removal addresses and admitted-input digest do not serialize.
The model starts from identity and explicit empty collections, then reconstructs
sections, Core records, Timeline events, subjects, relationships and indexes
from selected content. Canonical bytes and the complete post-redaction digest
are retained in `RedactionResult`; the admitted-input digest remains distinct.

Subject references and relationship labels follow selected fields. An edge
requires both retained Core endpoints and unchanged structural inputs; a
transformed ID cannot create a different edge. Source and support indexes retain
the necessary identity closure and immutable snapshot addresses. A citation
can retain a metadata-only source anchor without restoring removed fields or a
Core summary. Authored presentation references are narrowed to their selected
source fields. Deep copies prevent aliasing the snapshot. Reconstruction is
validated before publishing a passed export-model summary. Deck and Timeline
derivation use the canonical collections; remote-asset admission traverses the
same retained content rather than the private decision index.

`retained_model_test.go` is a helper of the existing redaction entrypoint. It
asserts the complete 29-member shape, explicit arrays, exact identity and digest,
absence of private facts and removed displays, immutable historical addresses,
masked endpoint labels, metadata-only citations, and rejection of missing,
foreign or altered references and field copies. The remote-asset case verifies
that changing the private index cannot bypass canonical-content inspection.
No test or execution context was removed. Existing direct renderer fixtures now
construct the canonical typed container explicitly.

This checkpoint does not close the full redaction census. Unsupported exact
container selectors still fail explicitly; section/record/reference rule
semantics, lifecycle eligibility, expansion, narrative slots, and complete
diagram/asset integration remain open. Source-provided diagram/asset objects
cannot bypass their later owning derivation. No new migration or legacy reader
is required: admitted input schemas and retained historical formats are
unchanged by this correction to an internal post-redaction representation.

Roots below are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-retained-model-format-20261015-01` | `make format` | Outer 0, 5.882 s. |
| `remediation-retained-model-format-20261015-02` | `make format` | Outer 0, 6.508 s. |
| `remediation-retained-model-model-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | Outer 0, 3.698 s. |
| `remediation-retained-model-format-20261015-03` | `make format` | Outer 0, 6.173 s. |
| `remediation-retained-model-format-20261015-04` | `make format` | Outer 0, 6.110 s. |
| `remediation-retained-model-format-20261015-05` | `make format` | Outer 0, 6.055 s. |
| `remediation-retained-model-model-20261015-02` | Same two exact owner rows | Outer 0, 3.612 s; adds reference closure, reconstruction validation and canonical runtime-asset inspection. |
| `remediation-retained-model-format-20261015-06` | `make format` | Outer 0, 5.853 s. |
| `remediation-retained-model-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.625 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-retained-model-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 44.697 s. |
| `remediation-retained-model-lint-20261015-01` | `make lint` | Outer 0, 37.096 s. |
| `remediation-retained-model-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 94.209 s; 24/24 units, permitted cached evidence and fresh affected/native execution. |
| `remediation-retained-model-drift-20261015-01` | `make generate-drift` | Outer 0, 13.662 s. |
| `remediation-retained-model-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-retained-model-module.reporting-20261015-01.json` | Outer 0, 2.728 s. |
| `remediation-retained-model-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-retained-model-module.reportcomposition-20261015-01.json` | Outer 0, 2.696 s. |

The Reporting invocation binds source
`sha256:dca113921693633a53fc7eff28f83d0a40a5b1454d622c743f0b42fe9dbe89ba`
and graph
`sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
Concurrent functional owner invocations do not constitute performance windows.
Coordinated adoption remains withheld for the open obligations above.

## Frozen record lifecycle and release eligibility

Snapshot v5 and ExportModel v6 supersede v4/v5. Private SourceFacts v2 adds the
complete incident-scoped deleted Core identity roster; ProviderOutput v4 assigns
that contribution exclusively to Records. Deleted values, actors and timestamps
are never collected into case content. The roster is frozen in the snapshot
transaction, independently validated against live source overlap, and excluded
from distributable output. Migration 86 rejects retained pre-cutover admissions
before mutation; it never reconstructs deletion history. Extensions' authored
contract/profile binding and generated projections use ExportModel v6.

Timeline `capture_state` and Decision `is_superseded` now establish lifecycle
from their explicit source owners. Executed-but-superseded decisions remain
superseded. Missing or invalid state and envelope-only stateful summaries fail
admission. All non-null source copies, including generic envelopes, carry the
superseded source state; nulls remain null. Replay validates rather than repairs.

Eligibility precedes Core redaction. Deleted dependencies are excluded in all
scopes; external superseded content requires both profile permission and a
matching opted-in section. Link, mention, tag and support dependencies inherit
restrictions without conflating auxiliary and Core IDs. Exclusions do not invent
Core drop decisions or structural-removal causes. Explicit deleted record,
Timeline and composition inclusions reject with the registered reason. Section
selection, diagram placement and post-composition source references enforce the
same boundary. A permission in one section cannot be reused in another.

The existing materialization/redaction rows call the new lifecycle helpers.
They cover owner state, null preservation, frozen tampering, internal/external
eligibility, both independent permissions, auxiliary identity collisions,
reference inheritance, exact-rule bypass, deleted inclusion, diagram placement,
post-operation content and source immutability. The Records integration row
checks incident scoping and the private roster. Existing Timeline fixtures now
supply their explicit owner state without retiring their original obligations.

This is a bounded lifecycle correction. Full container/record/reference Core
rule coverage, expansion, narrative slots, graph disclosure, canonical diagram
and asset integration, and coordinated adoption remain open. The later full
owner results below must qualify this source before implementation closure.

Roots below are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-lifecycle-generate-20261015-01` | `make generate` | Outer 0; refreshed authored contract, participant and migration projections. |
| `remediation-lifecycle-format-20261015-01` | `make format` | Outer 0, 6.281 s. |
| `remediation-lifecycle-facts-20261015-01` | Materialization, composition binding and redaction rows | Outer 0, 7.896 s. |
| `remediation-lifecycle-records-20261015-01` | `make service-backed-test-slice OWNER=module.records ROWS=module.records.reporting.envelope_facts` | Outer 0, 15.582 s. |
| `remediation-lifecycle-cutover-20261015-01` | `make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` | Retained summary reports pass; outer exit unavailable after session interruption. Not reused as independently verified command completion. |
| `remediation-lifecycle-format-20261015-02` | `make format` | Outer 0, 6.112 s. |
| `remediation-lifecycle-model-20261015-01` | Same three exact owner rows as facts run | Outer 0, 3.780 s. |
| `remediation-lifecycle-format-20261015-03` | `make format` | Outer 0, 6.108 s. |
| `remediation-lifecycle-model-20261015-02` | Same three exact owner rows | Outer 0, 3.887 s; lifecycle and reference adversarial cases. |
| `remediation-lifecycle-format-20261015-04` | `make format` | Outer 0, 6.207 s. |
| `remediation-lifecycle-model-20261015-03` | Same three exact owner rows | Outer 2, 3.702 s; `product/test_assertion_failure`. New composition fixture supplied null `label_overrides`, correctly rejected as `invalid_schema_id`. All 10 cleanup steps completed. Fixture corrected to explicit empty array; no production admission relaxation. |
| `remediation-lifecycle-format-20261015-05` | `make format` | Outer 0, 5.983 s. |
| `remediation-lifecycle-model-20261015-04` | Same three exact owner rows | Outer 0, 3.357 s; adds diagram placement and post-composition checks. |
| `remediation-lifecycle-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.214 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-lifecycle-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 36.088 s. |
| `remediation-lifecycle-migrations-20261015-01` | `make test-slice OWNER=module.database_migrations` | Outer 0, 66.682 s; 9/9 units. |
| `remediation-lifecycle-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 97.122 s; 24/24 units, permitted cached and fresh affected/native execution. |
| `remediation-lifecycle-lint-20261015-01` | `make lint` | Outer 2, 37.413 s; `harness/child_target_failure`, Go staticcheck S1021 in new lifecycle helper. All 20 cleanup steps completed; declaration merged without semantic change. |

Review after these runs found that one-hop lifecycle propagation was insufficient
for chained citations, and that validating bindings only against eligible
candidates would reject admitted null-preservation rules for excluded records.
Lifecycle now uses the existing monotone disclosure dependency graph with a
separate typed state; it is never encoded as a disclosure label. Cycles and
transitive dependencies preserve the strongest restriction. Exact rules bind to
the full admitted candidate inventory and remain inactive when lifecycle excludes
their target. Missing and unsupported targets still reject. Regression cases
include cyclic citations and the real built-in nullable-field profile. This is
additional source work requiring the later validation below; earlier results
are not silently promoted to current-source assurance.

| Run | Command | Result |
| --- | --- | --- |
| `remediation-lifecycle-format-20261015-06` | `make format` | Outer 0, 6.072 s. |
| `remediation-lifecycle-model-20261015-05` | Materialization, owned metadata, deterministic redaction and recipient-profile rows | Outer 0, 3.753 s. |
| `remediation-lifecycle-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.212 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-lifecycle-composition-20261015-02` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 42.229 s. |
| `remediation-lifecycle-reporting-20261015-02` | `make test-slice OWNER=module.reporting` | Outer 0, 97.630 s; 24/24 units, 11 permitted cache hits and fresh affected/native execution. |
| `remediation-lifecycle-lint-20261015-02` | `make lint` | Outer 0, 40.245 s. |
| `remediation-lifecycle-drift-20261015-01` | `make generate-drift` | Outer 0, 21.438 s. |
| `remediation-lifecycle-shapes-20261015-01` | `make json-shape-check` | Outer 0, 8.747 s. |
| `remediation-lifecycle-policy-20261015-01` | `make generated-artifact-policy-check` | Outer 0, 1.274 s. |
| `remediation-lifecycle-extensions-20261015-01` | `make test-slice OWNER=module.extensions ROWS=module.extensions.unit.coordinator_binding_admission_c2d8beffdb,module.extensions.unit.coordinator_registry_1dce9a539a` | Outer 0, 43.172 s. |
| `remediation-lifecycle-contractgen-20261015-01` | `make test-slice OWNER=harness.generated_artifacts ROWS=harness.generated_artifacts.unit.decode_contract_rejects_duplicate_jsonmembers_005d1b3b94` | Outer 0, 42.058 s. |
| `remediation-lifecycle-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-lifecycle-module.reporting-20261015-02.json` | Outer 0, 2.783 s. |
| `remediation-lifecycle-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-lifecycle-module.reportcomposition-20261015-02.json` | Outer 0, 2.702 s. |

The final Reporting invocation binds source
`sha256:82747fe8f0a82efd1e528495a53e709b7e3f1754215057aff637fb0b9f6b1f12`
and graph
`sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
These functional runs are not qualified performance windows. Lifecycle closure
does not close the remaining derivation obligations or authorize adoption.

`remediation-lifecycle-markdown-20261015-01` subsequently completed
`make lint-markdown` with outer exit 0, 16.040 s.

## Template expansion and Artifact narrative projection

Section derivation now emits deterministic Party and subject instances using
frozen identity dimensions, retained content, and stable semantic composition
keys. It does not use labels as identities. Internal Party dimensions survive
display-field removal; external dimensions intersect exact recipients. Party
instances require all field partitions to fit that instance, and subject
instances require exact retained subject-source membership. Missing dimensions,
required empty instances and multiplied section/block/field limits fail before
renderer launch. Selected tokens take label priority; retained eligible Party
labels and deterministic ordinal fallbacks never read raw display values.
Case-derived titles preserve their source/support/partition evidence.

Narrative slots select explicit Artifact field addresses, emit paragraphs in
deterministic record/field order, and preserve selected values and content class.
The admitted-source check rejects non-Artifact matches even if later redaction
would hide them. Removed, null and empty values cannot be restored. Required
empty slots and invalid scalar types reject. Limits count Unicode scalars per
slot instance without truncation. Deleted exact narrative bindings use the
existing lifecycle denial. No release-request text becomes case narrative.

Tables 21-C/E now close the previously unstated dimension membership, fallback,
ordering, field-address and over-limit rules. This is a proposed-amendment
clarification, not a retroactive claim that the adopted base specified those
details. Existing implemented singleton binding behavior is preserved. There
is no legacy reader or alternate template language. Private Party identities
and token lookups are reconstructed during replay and never serialized into the
canonical retained case model.

`template_derivation_test.go` is a helper of existing
`TestSlidevCanonicalSourceV2`, routed by the metadata/source-contract row. It
checks dimension ordering, exact anchors, recipient/multi-Party isolation,
token labels, removal fallbacks, immutable input, provenance, Unicode limits,
invalid and empty slot values, and render/replay equality through the existing
fixture runtime. The fixture-runtime evidence does not claim new native-browser
qualification. Existing entrypoints and assurance contexts remain.

Full canonical template-selected export-model assembly, non-section ordinary
bindings, diagram/asset integration, complete Core container-rule semantics and
graph disclosure still require their own closure. This slice does not promote
Reporting or Composition, or close Findings 2, 3, 4, 7 or 9 by inference.

Roots below are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-template-format-20261015-01` | `make format` | Outer 0, 6.312 s. |
| `remediation-template-model-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393` | Outer 2, 3.479 s; `product/test_assertion_failure`. New narrative fixture omitted its title from the complete fixture path census. All 10 cleanup steps completed; fixture corrected, no admission relaxation. Expansion assertions passed. |
| `remediation-template-format-20261015-02` | `make format` | Outer 0, 6.125 s. |
| `remediation-template-model-20261015-02` | Same exact owner row | Outer 0, 3.140 s. |
| `remediation-template-format-20261015-03` | `make format` | Outer 0, 6.156 s. |
| `remediation-template-model-20261015-03` | Same exact owner row | Outer 0, 3.634 s; adds render/replay, null, empty and invalid-type assertions. |
| `remediation-template-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.234 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-template-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 44.532 s. |
| `remediation-template-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 96.146 s; 24/24 units, permitted cached evidence and fresh affected/native execution. |
| `remediation-template-lint-20261015-01` | `make lint` | Outer 0, 38.628 s. |
| `remediation-template-drift-20261015-01` | `make generate-drift` | Outer 0, 26.815 s. |
| `remediation-template-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-template-module.reporting-20261015-01.json` | Outer 0, 2.781 s. |
| `remediation-template-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-template-module.reportcomposition-20261015-01.json` | Outer 0, 2.758 s. |

The Reporting invocation binds source
`sha256:9edf258e770a4cb7a39cf108bfc137bec791c7f60902be89b77fd245275e543c`
and graph
`sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
These functional invocations do not establish a qualified performance window.

`remediation-template-markdown-20261015-01` completed `make lint-markdown`
with outer exit 0 after this checkpoint's evidence rows.

## Canonical record and Timeline template bindings

Ordinary section selection now resolves field-valued bindings through typed
source sections, Core record summaries and Timeline views. Their shared source
identity prevents duplication when several addresses select the same value.
Core record selection never widens to a same-ID envelope or auxiliary source.
The lifecycle eligibility stage uses the same address relation, so a permitted
record/Timeline binding can select a superseded value only with both existing
permissions. Redaction rule and manifest addresses remain the original exact
canonical addresses; template aliases do not create another redaction policy.

Field-binding grammar rejects metadata scalars, traversal inside `value`,
numeric indexes, unsupported roots and query expressions at template admission,
including optional sections. Required absent fields continue to fail during
derivation. Tables 21-C/E remain the normative owners of those semantics.

`template_bindings_test.go` is another helper of
`TestSlidevCanonicalSourceV2`, with positive typed bindings, overlapping-view
deduplication, Core/envelope/auxiliary identity separation, admission denial and
fixture-render replay. The lifecycle helper adds both positive alias opt-in and
required-section denial without that opt-in. No entrypoint or evidence context
was removed. Canonical final model assembly and the broader adoption blockers
listed in the preceding checkpoint remain open.

| Run | Command | Result |
| --- | --- | --- |
| `remediation-template-bindings-format-20261015-01` | `make format` | Outer 0, 6.272 s. |
| `remediation-template-bindings-model-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | Outer 0, 3.828 s. |
| `remediation-template-bindings-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.394 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-template-bindings-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 93.495 s; 24/24 units, 11 permitted cache hits and fresh affected/native execution. |
| `remediation-template-bindings-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 40.947 s. |
| `remediation-template-bindings-lint-20261015-01` | `make lint` | Outer 0, 36.637 s. |
| `remediation-template-bindings-drift-20261015-01` | `make generate-drift` | Outer 0, 13.740 s. |
| `remediation-template-bindings-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-template-bindings-module.reporting-20261015-01.json` | Outer 0, 2.770 s. |
| `remediation-template-bindings-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-template-bindings-module.reportcomposition-20261015-01.json` | Outer 0, 2.768 s. |

The Reporting invocation binds source
`sha256:77d8a6dd7681b98fc76e68228b0f2a4d0833dfa52966a62aeb21fc94863f6aa1`
and graph
`sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
These are functional evidence, not qualified performance windows.

`remediation-template-bindings-markdown-20261015-01` completed
`make lint-markdown` with outer exit 0, 9.240 s.

## Template and composition source reachability

Source reachability now precedes Core actions and token allocation. Actual
section and narrative bindings select fields; the binding allowlist does not
select data. Expansion membership, Party display dependencies, authored subject
placeholders, exact Timeline diagram selections, relationship endpoints and
composition record anchors have explicit closure rules in Table 7-A3. The
selection never widens a field binding to sibling or same-ID auxiliary values.
Support-only citations retain anchors without restoring record fields.
Lifecycle exclusion applies before dependency expansion, so an ineligible
relationship cannot select active endpoint values as a side effect. Required
missing sections still fail at section derivation; an empty selection is not
misclassified as a malformed empty manifest. Null arrays remain invalid.

The source, retained records, evidence and token artifacts share the same
selection. Unselected values produce neither fabricated drop entries nor
structural-removal evidence. Profile paths still bind against the complete
admitted inventory. Frozen-source identity and the independently bound
composition remain unchanged. No compatibility reader or alternate renderer
input was added.

`template_reachability_test.go` is a helper of the existing Slidev source
entrypoint and metadata row. It checks unselected literal/token absence,
allowlist isolation, missing required selection, Party label dependencies,
exact Timeline selection, authored subject display dependencies, relationship
endpoints, support-only references, cancellation and immutable source bytes.
Fixture-runtime render/replay remains distinct from native evidence. Canonical
expanded-section assembly, complete container rules and graph disclosure still
need closure; this slice does not authorize coordinated adoption.

Roots below are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-reachability-format-20261015-01` | `make format` | Outer 0, 5.820 s. |
| `remediation-reachability-model-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | Outer 2, 3.761 s; `product/test_assertion_failure`. New replay fixture had unsorted binding allowlist; corrected fixture, no admission relaxation. All 10 cleanup steps completed. |
| `remediation-reachability-model-20261015-02` | Same two exact rows | Outer 2, 3.308 s; `product/test_assertion_failure`. New Party expansion fixture omitted the disclosure label required for dimension membership. Explicit public label added to synthetic source. All 10 cleanup steps completed. |
| `remediation-reachability-model-20261015-03` | Same two exact rows | Outer 0, 3.574 s. |
| `remediation-reachability-format-20261015-02` | `make format` | Outer 0, 6.154 s. |
| `remediation-reachability-model-20261015-04` | Same two exact rows | Outer 0, 3.785 s; adds relationship/support closure and cancellation. |
| `remediation-reachability-format-20261015-03` | `make format` | Outer 0, 6.382 s. |
| `remediation-reachability-model-20261015-05` | Same two exact rows | Outer 0, 3.257 s; eligibility before dependency expansion and regression assertion. |
| `remediation-reachability-finalize-20261015-01` | `make agent-finalize` | Outer 0, 19.937 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-reachability-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 42.267 s. |
| `remediation-reachability-lint-20261015-01` | `make lint` | Outer 0, 35.791 s. |
| `remediation-reachability-drift-20261015-01` | `make generate-drift` | Outer 0, 14.613 s. |
| `remediation-reachability-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 2, 93.762 s; 23/24 units passed, including native rendering. `product/test_assertion_failure` in decoder fixture: manual-layout denial requested a missing Timeline event. Source admission now correctly rejects that missing reference earlier. All 28 cleanup steps completed. Fixture given a valid synthetic event to isolate the intended layout denial; production precedence unchanged. |
| `remediation-reachability-format-20261015-04` | `make format` | Outer 0, 6.302 s. |
| `remediation-reachability-decoder-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.snapshot_reporting_route_decoders_reject_legacy_9d9f64130f` | Outer 0, 3.414 s. |

The corrected full-owner runs below replace the earlier source-incompatible
evidence. The failed runs are retained, not treated as passes or stability retries.

| Run | Command | Result |
| --- | --- | --- |
| `remediation-reachability-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.533 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-reachability-reporting-20261015-02` | `make test-slice OWNER=module.reporting` | Outer 0, 77.158 s; 24/24 units, 11 permitted cache hits and fresh affected/native execution. |
| `remediation-reachability-markdown-20261015-01` | `make lint-markdown` | Outer 0, 10.240 s, before these final evidence rows. |
| `remediation-reachability-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-reachability-module.reporting-20261015-02.json` | Outer 0, 2.658 s. |
| `remediation-reachability-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-reachability-module.reportcomposition-20261015-01.json` | Outer 2, 1.583 s; `artifact/artifact_error`: the earlier owner run predates the Reporting fixture correction and therefore differs from the audited executable snapshot. All 10 cleanup steps completed. No evidence override applied. |
| `remediation-reachability-composition-20261015-02` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 23.181 s. |
| `remediation-reachability-composition-audit-20261015-02` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-reachability-module.reportcomposition-20261015-02.json` | Retained summary passed, 2.574 s; retained-secret scan passed. The native session was unavailable after context recovery, so its outer exit is unavailable, not inferred from the summary. |

The corrected Reporting run binds source
`sha256:1abff939648864db7481c89e7493afc06d31c3eec0368d3bec8e77fda5e7ad82`
and graph
`sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
These functional results do not qualify a performance window or close the
remaining adoption conditions.

`remediation-reachability-markdown-20261015-02` completed
`make lint-markdown` with outer exit 0, 9.420 s.

## Canonical render-model assembly

Production now assembles expanded template sections, narrative blocks and full
retained diagrams before canonical export serialization. Diagram reference blocks
are canonical section content and count toward multiplied block/field limits.
Case-derived title references join section reference unions. Canonical-byte
limits apply after assembly. Snapshot field uniqueness remains strict; repeated
selected values in different emitted sections are deliberate occurrences.
Table 11-A now distinguishes stable source-address binding from final retained
assembly, avoiding a contradictory requirement to compute redaction-safe labels
before redaction.

Deck derivation and diagram source preparation read only the assembled model.
Their APIs no longer accept graph results or source-selection inputs. Private
assembly metadata contains template/model digests and semantic section origins,
not a parallel content tree. Consumers reject altered canonical values and a
different template, and receive independent copies before composition/chunking.
Source selection and pure assembly remain shared with retained-output replay.
The existing redaction deadline now covers section projection; cancellation
cannot return partial canonical bytes.

The new helper under `TestSlidevCanonicalSourceV2` checks canonical repeated
sections, diagram content and references, title evidence, immutable inputs and
consumer copies, private-index independence, model/template tampering,
cancellation, resource bounds including diagram references, and fixture-render
replay. Original low-level serializer fixtures use an explicit test-only adapter
for their intentionally incomplete or invalid boundary values. Production has
no alternate reader or unassembled-model fallback.

This closes the identified canonical section/diagram assembly defect. It does
not close graph disclosure/redaction, declared template assets, full Core
container-rule semantics, companion/recovery qualification or coordinated
adoption. Native validation below remains environment-specific.

Roots below are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-assembly-format-20261015-01` | `make format` | Outer 0, 6.186 s. |
| `remediation-assembly-model-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | Outer 2, 3.622 s; `product/test_assertion_failure`. Earlier low-level fixtures depended on pre-assembly boundaries: intentionally dropped required source section, empty Mermaid-only source, and incomplete synthetic Timeline metadata. All 10 cleanup steps completed. Corrected fixture boundaries; no production admission relaxation. |
| `remediation-assembly-format-20261015-02` | `make format` | Outer 0, 6.413 s. |
| `remediation-assembly-model-20261015-02` | Same two exact owner rows | Outer 0, 3.748 s; includes new canonical assembly checks. |
| `remediation-assembly-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.530 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-assembly-drift-20261015-01` | `make generate-drift` | Outer 0, 27.556 s. |

Full-owner verification and evidence audits are pending at this checkpoint.

| Run | Command | Result |
| --- | --- | --- |
| `remediation-assembly-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 45.544 s. |
| `remediation-assembly-lint-20261015-01` | `make lint` | Outer 0, 37.798 s. |
| `remediation-assembly-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 2, 96.340 s; 23/24 units passed, including native rendering. `product/test_assertion_failure`: moving template derivation under redaction changed undeclared-binding classification to generic invalid scalar. All 28 cleanup steps completed. Moved template contract validation before redaction, preserving the established typed reason. |
| `remediation-assembly-decoder-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.snapshot_reporting_route_decoders_reject_legacy_9d9f64130f,module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393` | Outer 0, 3.607 s. |
| `remediation-assembly-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.378 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-assembly-format-20261015-03` | `make format` | Outer 0, 5.974 s. |
| `remediation-assembly-finalize-20261015-03` | `make agent-finalize` | Outer 0, 20.544 s; maintenance skipped because `RESULTS_DIR` was unset. |

Final boundary review also found declared retained record, relationship, subject
and support-reference limits without enforcement. Assembly now checks those
collection counts; the new regression case exercises record and subject limits.
The final full-owner runs below include these changes and the classification
repair. Earlier owner evidence is source-specific and is not silently reused
for an exact-source evidence audit.

| Run | Command | Result |
| --- | --- | --- |
| `remediation-assembly-composition-20261015-02` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 34.856 s. |
| `remediation-assembly-reporting-20261015-02` | `make test-slice OWNER=module.reporting` | Outer 2, 81.221 s; 23/24 units passed, including native rendering. `product/test_assertion_failure`: the complete-model diagnostic fixture supplied the worker mock's empty template, so correct early template validation prevented reaching its seven intended model defects. All 28 cleanup steps completed. Fixture now supplies a valid template/output selection. Combined follow-up verification remains required. |

## Admitted and assembled model identity review

Assembly review exposed a separate identity discrepancy: the v2 bundle's
`export_model_sha256` contains the pre-redaction admitted model digest, while
Table 22-A describes the final canonical export-model digest. Bundle manifest v3
now binds both values separately. Completion and native recovery compare the
assembled digest against deterministic replay;
portable consumers can verify bundle integrity without being given private
source material. No canonical-model file is added to the shareable inventory:
private source/token metadata is not made public to solve a digest ambiguity.
External determinism now repeats materialization as well as renderer execution,
comparing canonical model, redaction manifest and token manifest bytes/digests
before the existing independent native render comparison. The comparisons use
existing typed mismatch reasons and never log source or replacement values.

Migration 87 rejects retained snapshots, jobs, bundles, imported artifacts or
preview attempts before mutation. The preflight recognizes the new bundle
constraint rather than treating the earlier lifecycle constraint as sufficient.
The database and shareable-bundle reader require v3 and both model identities;
retired v2, missing/null digests and rehashed substituted digests reject. Public
release admission metadata retains its original admitted-model meaning.
Table 22-A now enumerates the actual closed manifest, including explicit nulls
and external descriptor digests instead of a recursive self-hash member.

The seven-defect diagnostic test now supplies a valid template before exercising
its invalid model. This preserves early template validation and the complete
diagnostic census. No missing-template production fallback was added.

Roots below are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-model-identity-format-20261015-01` | `make format` | Outer 0, 6.150 s. |
| `remediation-model-identity-unit-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393,module.reporting.unit.frozen_release_payload,module.reporting.unit.reporting_job_worker_cancels_before_work_68d5024f75,module.reporting.unit.snapshot_reporting_route_decoders_reject_legacy_9d9f64130f` | Outer 0, 7.323 s. Later missing/null/retired-manifest cases require the subsequent full owner run. |
| `remediation-model-identity-format-20261015-02` | `make format` | Outer 0, 6.331 s. |
| `remediation-model-identity-generate-20261015-01` | `make generate` | Retained tool summary failed, tool exit 1, 1.780 s; native outer exit unavailable after context recovery. Migration 87 lacked an authored source-owner assignment. No product test failure. |
| `remediation-model-identity-finalize-20261015-01` | `make agent-finalize` | Outer 2, 5.023 s; `harness/child_target_failure` from the same missing assignment. All 10 cleanup steps completed. Added migration 87 to the authored Reporting allocation. |
| `remediation-model-identity-generate-20261015-02` | `make generate` | Outer 0, 10.360 s; regenerated catalog projections from authored ownership. |
| `remediation-model-identity-finalize-20261015-02` | `make agent-finalize` | Outer 0, 20.556 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-model-identity-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 35.046 s. |
| `remediation-model-identity-portability-20261015-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | Outer 0, 50.078 s; selected portability row, not full Incident Bundles qualification. |
| `remediation-model-identity-drift-20261015-01` | `make generate-drift` | Outer 0, 25.466 s. |
| `remediation-model-identity-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 93.958 s; 24/24 units, including native rendering and recovery; 11 permitted cache hits, fresh affected execution. |
| `remediation-model-identity-lint-20261015-01` | `make lint` | Outer 0, 34.088 s. |
| `remediation-model-identity-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-model-identity-module.reporting-20261015-01.json` | Outer 0, 2.684 s. |
| `remediation-model-identity-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-model-identity-module.reportcomposition-20261015-01.json` | Outer 0, 2.675 s. |

Reporting binds source
`sha256:bcfed6792b7a8de333c357875f0d1bb60fe7cf9fadf10ece1856a4aeb3d31745`
and graph
`sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
The identity and assembly corrections are validated within these owner contexts.
These changes do not establish coordinated adoption closure or qualify a cost
window. Other sandbox/LSM environments retain their earlier applicability limits.

## Typed graph redaction and canonical topology

Network Flow's frozen typed label candidates now enter the same exact-path,
content-class and default-rule selector as source and authored fields. Selection
retains only topology and decision addresses for later diagram assembly. No
diagram label generator receives raw Graph labels, properties or component
values. Canonical diagrams contain stable vertex/edge IDs, with ordinal Mermaid
identifiers assigned only by the serializer. This also makes manual layout and
structured override targets resolve against the actual selected graph identity.

Reporting §15 defines each pre-redaction component address, null source state,
manifest census and post-action partition check. Component drop preserves
topology and uses only the existing ordinal fallback; it does not pretend to
implement vertex/container drop. The Network Flow v2 handoff has no Core subject
mapping, so token-backed rules reject rather than infer a subject from text.
Non-tokenized graph overrides use the ordinary composition and label rules.
External Network Flow graph selection remains explicitly unsupported and fails
closed under the existing owner policy. General container-rule semantics remain
open; this slice does not remove that obligation.

The production materialization test covers mask, stub, drop, truncate, retained
canonical labels, value-free rule evidence, unselected-path rejection, unmapped
subject rejection, partition denial, immutable replay, manual geometry and
opaque override targets. Existing low-level graph cases keep their temporal,
absent/removed label, exact topology and external rejection assertions. Renderer
version advances to 17. Core 01's proposed built-in profile version reference is
reconciled with the already implemented version 3; no legacy alias is added.

All roots below are under `.cartulary/test-results/`:

| Run | Command | Result |
| --- | --- | --- |
| `remediation-graph-redaction-format-20261015-01` | `make format` | Outer 0, 6.424 s. |
| `remediation-graph-redaction-unit-20261015-01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.graph_v2_exact_selection_and_rendering_f54d8b1a62,module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393` | Outer 0, 3.917 s. Later manual/truncate cases are included in the full owner run. |
| `remediation-graph-redaction-format-20261015-02` | `make format` | Outer 0, 6.081 s. |
| `remediation-graph-redaction-finalize-20261015-01` | `make agent-finalize` | Outer 0, 20.001 s; maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-graph-redaction-reporting-20261015-01` | `make test-slice OWNER=module.reporting` | Outer 0, 95.583 s; 24/24 units, 11 permitted cache hits and fresh affected/native execution. |
| `remediation-graph-redaction-composition-20261015-01` | `make test-slice OWNER=module.reportcomposition` | Outer 0, 44.610 s. |
| `remediation-graph-redaction-lint-20261015-01` | `make lint` | Outer 0, 37.279 s. |
| `remediation-graph-redaction-reporting-audit-20261015-01` | `make test-evidence-audit OWNER=module.reporting` with `.cartulary/evidence-inputs/remediation-graph-redaction-module.reporting-20261015-01.json` | Outer 0, 2.712 s. |
| `remediation-graph-redaction-composition-audit-20261015-01` | `make test-evidence-audit OWNER=module.reportcomposition` with `.cartulary/evidence-inputs/remediation-graph-redaction-module.reportcomposition-20261015-01.json` | Outer 0, 2.682 s. |

Source: `sha256:267e1cf1e5951f231d21c714b1885a46bf3f32fbc89d0b8e291018a5d9b89f53`.
Graph: `sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`.
This qualifies the exercised graph behavior in the current native environment.
It is not a qualified execution-cost window, an additional LSM qualification,
or coordinated adoption closure. Container redaction, declared template assets,
companion review and the remaining product evidence remain open.

### Current packaged renderer qualification — October 15 continuation

`CARTULARY_TEST_RUN_ID=remediation-current-render-package-20261015-01
CARTULARY_OUTPUT_MODE=machine make standup-reference-pack-smoke` completed with
native Make exit `0`, 15/15 successful work units, 24 successful cleanup steps,
and 461,760 ms invocation wall time. The exact root is
`.cartulary/test-results/remediation-current-render-package-20261015-01`.
Its source digest is
`sha256:63732c750d0e1f173a9dec47fc7218ec4c51985917d810fcf04ed46248cfda00`,
and graph digest is
`sha256:3fdadeab1fbb0e72d04295c483b4e2a139f79d55c8d5c074fe8ad9ddcff7b9d6`.
This fresh packaged qualification includes the canonical model/bundle identity
and typed graph-redaction corrections. The source state is dirty at the retained
baseline commit; this is neither a qualified cost window nor a Core 05 claim.
Its confinement scope remains the exercised native-amd64/no-LSM environment;
unsupported LSM combinations are not qualified by inference.

### Entity assignment write dependency — implementation in progress

The existing Links source projection has no ordinary Host/Identity write path
in the baseline. Core 01 REQ-01-327a and Core 03 REQ-03-287a now explicitly
require the six default-hidden assignment collections through ordinary entity
creation/upsert, record patch, and inspector controls. Create/upsert accepts
additions only and still requires an entity identity seed. Links owns membership;
the entity owner coordinates authority, lifecycle, parent versions, history,
idempotency, and publication. The bounded source hydration uses Links facts and
Party-owned labels. No live recipient policy or Reporting mutation route is
introduced. Default visible columns and timed-grid geometry stay unchanged.

This is an implementation slice, not promotion evidence yet. Required evidence
includes same-incident/live-Party admission, wrong-family/scalar/unknown-member
rejection, atomic rollback, exact replay, independent field membership, parent
versioning, actual link history, snapshot isolation, inspector retained drafts,
and current authority. The ordinary source write must be demonstrated before
claiming that the companion adoption dependency is satisfied.

### Entity assignment validation — current continuation

The ordinary source path now has public-route integration evidence for both
entity families: create and exact-match reuse, exact replay, independent role
membership, one parent version advance, typed link history, atomic rejection,
current viewer denial and repeatable-read snapshot isolation. Core 02 REQ-02-231
also requires incoming-reference protection: each of the six assignment fields
independently blocks Party deletion, removal permits deletion, and a subsequent
assignment to that deleted Party rejects. No migration or alias is introduced.
The Records boundary grants only the Party-owned bounded label reader access
to current record envelopes; Host/Identity code does not read Party tables.

Retained roots below are under `.cartulary/test-results/`. Native exit codes are
separate from normalized test classification. They are correctness evidence,
not cost windows or a general stability claim.

| Root | Selection / result | Disposition |
| --- | --- | --- |
| `remediation-party-assignments-focused-20261015-01` | Focused entity integration, admission, registry, boundary and OpenAPI rows; native 0, 15.840 s, 5/5 units | Passed before the later deletion-guard addition |
| `remediation-party-assignments-ui-20261015-01` | Existing inspector component row with assignment scenario; native 0, 6.162 s | Initial draft coverage; later browser review found a missing round trip |
| `remediation-party-assignments-entities-20261015-01` | Full Entities owner; native 2, product/test_assertion_failure, 50/51 units, 190.982 s, 102 cleanup steps complete | Old create-shape assertion still expected 15 cells. Corrected to 18 and extended its atomic patch assertion to all eleven writable fields. All selected browser groups passed. |
| `remediation-party-assignments-lint-20261015-01` | Native 2, 43.490 s, 20 cleanup steps complete | Missing narrow Records read declaration and capitalized error string; both corrected. |
| `remediation-party-assignments-links-20261015-01` | Exact Party assignment snapshot-facts integration row; native 0, 33.417 s, 3/3 units | Snapshot-source evidence; no new live-policy lookup |
| `remediation-party-assignments-backend-20261015-02` | New assignment and existing entity origin-upsert integration rows; native 0, 39.469 s, 3/3 units | Includes all six independent deletion guards and existing auth/CSRF precedence assertions |
| `remediation-party-assignments-lint-20261015-02` | Native 0, 34.643 s, 11/11 units | Current backend boundary, Go and frontend checks passed before the draft-navigation correction |
| `remediation-party-assignments-browser-20261015-01` | Exact new ordinary browser scenario; native 2, product/test_assertion_failure, 150.911 s | Mode navigation incorrectly depended on whether the selected draft was attached. After add/remove/add, both mode buttons were disabled. The fix gates navigation on current authority and pending mutation, while keeping explicit Resume/review before submission. No timeout increase or action retry. |
| `remediation-party-assignments-ui-20261015-02` | Native 2, product/test_assertion_failure, 4.422 s | New regression used unavailable DOM matchers. Corrected to the existing native element-property assertion style. |
| `remediation-party-assignments-ui-20261015-03` | Native 0, 4.892 s, 2/2 units | Round-trip navigation remains available; detached submission remains disabled until explicit Resume. |

The successful backend/lint source is
`sha256:374ca85014a14875a2bdd48f96dd4e1a9366fedbf0369215bb66bef718bf2701`.
The failed browser graph is
`sha256:512961074526bc3b3ad8e73cad5457b844a30b7d3d117caae9e957c30f06c383`.
The final browser execution passed in
`remediation-party-assignments-browser-20261015-03` (native 0, 92.949 s,
12/12 fresh units). It exercises all three roles for both Hosts and Identities,
ordinary reference selection, add/remove/add retained drafts, explicit Resume
and stale-draft review, persisted removal/addition and reload. The preceding
`remediation-party-assignments-browser-20261015-02` also passed (native 0,
93.430 s); rendered inspection then prompted a visible selected-mode checkmark.
The final screenshots target the editor control group, avoiding a misleading
capture of the surrounding scrollable panel.

`remediation-party-assignments-lint-20261015-03` failed frontend type checking
because the new Testing Library query used Playwright's unsupported `exact`
option. Removing that option repaired the test typing. Full lint then passed
in `remediation-party-assignments-lint-20261015-04` (native 0, 28.507 s,
11/11 units). `remediation-party-assignments-drift-20261015-02` passed
(native 0, 20.386 s, 4/4 units). All three final runs bind dirty source
`sha256:e6c98895bf64d11f2337c4e89af4d4c38355c6e92a53112c9a8eb57eeec619fa`;
the browser graph remains the exact identity above.

The earlier multi-goal drift invocation used one run ID for several public
commands. Its first `generate-drift` target passed, but the next target correctly
rejected the nonempty run root (native 2); that invocation is not a passing run.
Separate `remediation-party-assignments-generated-policy-20261015-01` and
`remediation-party-assignments-json-shape-20261015-01` passed with native 0.
Final formatting and `agent-finalize` passed in their `format-20261015-10`
and `finalize-20261015-05` roots. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset.

The artifact-only UI review imported the exact two PNG attachments from browser
03, verified original bytes/digests and inspected both at native resolution.
Both show the active Remove mode, visible keyboard focus and unclipped editor
controls. Host: 24,809 bytes, SHA-256
`06f369a3714ac6ef7b86db6fe5182cfb1129e9520166c57abd5049c8902f20c0`;
Identity: 25,023 bytes, SHA-256
`5bf6c3e6e232f8a92d94ce6ec8b8c22d8c7ca12da5416bced6913c22bbc15022`.
This is a visual observation, not an accessibility audit or Core 05 claim.
Session `uireview-c73ad3e744636183c8e5c9427670a471` closed with native 0,
terminal state `closed`, status `ok` and cleanup `complete`; its durable
receipt is under `remediation-party-assignments-review-20261015-01/ui-review/terminal.json`.
Private image bundles expired at cleanup. Coordinated adoption remains open.

### Canonical container redaction — implementation and validation

The proposed Reporting/Core amendment makes canonical content ownership explicit.
Sections, blocks and fields receive Core path/class/default selection; classless
sections have no invented classification. Records, relationships, events,
subjects and indexes are derived views of those retained values, not independent
raw copies eligible for a second, potentially conflicting action. This resolves
the earlier demand for independent rules on every duplicate projection while
preserving disclosure qualification of every emitted element. New independent
values still require admitted typed candidates. Reviewer: Codex, under the
approved implementation plan; this is not independent companion approval.

Container selection runs in preorder after lifecycle/reachability eligibility.
A selected drop suppresses its complete eligible subtree before field actions,
token assignment and retained projections. An allow is qualified from surviving
fields; empty pruning remains a distinct operation with immediate causal paths.
Exact selector binding still uses the complete admitted census, so a dropped
parent cannot hide a stale selector. Non-scalar masking, stubbing and truncation
reject rather than stringify a container or manufacture a paragraph.

Manifest v5 separates field decisions, container decisions, suppressed paths and
consequential removals. Closed validation checks path kind, nullable class,
selected trace, action/outcome, sorted unique suppression, complete ancestry,
exclusive decision ownership and exact retained partition unions. Native replay
must still recompute the full admitted census; portable structural validation
does not prove completeness of unseen inputs. Built-in profile version 4 freezes
exact allow rules for typed containers while retaining its field policies.
Renderer contract 18 consumes this evidence. Migration 88 and application
preflight reject incompatible retained history before mutation; there is no
legacy reader, provenance inference or in-place downgrade.

The new helper test is reached through existing
`TestRedactionProfilePrecedenceActionsAndManifest_Unit` and catalog row
`module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d`.
It exercises section/block drop, path/class/default precedence, outside-partition
rejection, nested suppression, cancellation and tampered evidence. Existing
subject-token tests now assert that a section drop creates neither tokens nor
reveal entries. Exact-path tests positively exercise supported container drops
while retaining stale/unsupported-target denial. Scalar-action fixtures now
explicitly retain their containers; this is fixture policy, not a production
fallback. Cutover integration extends its pre-migration seed through version 87.

Retained execution roots are under `.cartulary/test-results/`:

| Root | Result and disposition |
| --- | --- |
| `remediation-container-redaction-unit-20261015-01` | Native 2, product/test_assertion_failure, 3.597 s. Scalar fixtures relied on ignored container transforms, and the old path test still expected supported container drops to reject. Repaired explicit fixture policy and expectations. |
| `remediation-container-redaction-unit-20261015-02` | Native 2, product/test_assertion_failure, 3.533 s. The separate subject-token fixture retained an implicit container mask. Repaired its explicit envelope policy. |
| `remediation-container-redaction-unit-20261015-03` | Native 0, 3.247 s, three exact redaction/profile/disclosure rows. Subsequent nested/token/retired-version assertions still require final validation. |
| `remediation-container-redaction-generate-20261015-01` | Native 0, 10.420 s; authored migration ownership regenerated through Make. |

Full owner, portability/recovery, migration, generated-drift and final lint
validation remain pending for this slice. This record does not close coordinated
adoption, declared template assets, the remaining timed criteria or the cost
portfolio.

The first full-owner run, `remediation-container-redaction-owner-20261015-01`,
exited 2 with product/test_assertion_failure (94.923 s, 21/24 units passed).
It exposed the still-active database check requiring profile revision 3; the new
migration now replaces that check with revision 4. Public-route admission had
returned 500 before this correction. Additional scalar/token/template fixtures
were updated to explicitly retain containers. The native Mermaid, Slidev,
list-click and candidate-pipeline cases passed; the authored-text native case
failed at its fixture's implicit container mask and requires revalidation.

`remediation-container-redaction-lint-20261015-01` exited 2 with
harness/child_target_failure (39.823 s): staticcheck found two unused intermediate
fixture profile digests. Those redundant assignments were removed.
`remediation-container-redaction-drift-20261015-01` passed (native 0, 30.405 s),
and `remediation-container-redaction-markdown-20261015-01` passed before this
addition. The first finalizer passed (native 0, 20.762 s); retained-run maintenance
was skipped because `RESULTS_DIR` was unset.

The focused follow-up `remediation-container-redaction-unit-20261015-04` exited
2 (product/test_assertion_failure, 3.978 s), isolating the remaining source-class
fixture's implicit envelope transform. Its scalar classification policy now
explicitly retains typed containers. Container drop evidence was also corrected
to use the actual subject/reference-derived field restrictions from the eligible
census, rather than the earlier cached container union. A subject-drop assertion
requires those derived labels, and closed-object cases reject missing/unknown
container evidence members. Final validation remains pending.

The final container slice validates dirty source
`sha256:c53a48d1fed646fba1d4ee6534094512aa813c9163105790ea418fec7edef6a8`:

| Exact retained run | Native result / applicability |
| --- | --- |
| `remediation-container-redaction-owner-20261015-02` | 0, 95.813 s, 24/24 Reporting units; 11 compatible cached units, 13 freshly executed. Native renderer, release/preview recovery, public admission and migration rejection passed; 28/28 cleanup operations complete. Graph `sha256:89827a38013fd9351535c9da25215e08599adc0cb7b551ef020199bbf273aa04`. |
| `remediation-container-redaction-composition-20261015-01` | 0, 43.969 s, 5/5 fresh units, 18/18 cleanup complete; complete Composition owner scope. |
| `remediation-container-redaction-portability-20261015-01` | 0, 56.240 s, 3/3 fresh units, 14/14 cleanup complete; exact existing Incident Bundle reference-catalog/native-snapshot/superseded-Timeline round-trip row. Not a full Incident Bundles owner pass. |
| `remediation-container-redaction-lint-20261015-02` | 0, 35.579 s, 11/11 units. |
| `remediation-container-redaction-drift-20261015-02` | 0, 15.610 s, 4/4 units. |
| `remediation-container-redaction-shape-20261015-01` | 0, 5.388 s, 3/3 units. |
| `remediation-container-redaction-generated-policy-20261015-01` | 0, 1.032 s, 3/3 units. |

`remediation-container-redaction-finalize-20261015-02` failed (native 2,
harness/child_target_failure, 4.885 s) because migration 88's corrected authored
SQL had not yet regenerated its history hash. `make generate` in
`remediation-container-redaction-generate-20261015-02` passed, followed by
`remediation-container-redaction-finalize-20261015-03` (native 0, 20.644 s).
Retained-run maintenance remained skipped because `RESULTS_DIR` was unset.
These runs retain the failed attempt and its actual repair; no old run was reused.

Owner evidence input manifests are
`.cartulary/evidence-inputs/remediation-container-redaction-module.reporting-20261015-01.json`
and
`.cartulary/evidence-inputs/remediation-container-redaction-module.reportcomposition-20261015-01.json`.
They select exactly the full-owner roots above. Container selection and its
versioned evidence are implemented and exercised at this source. This does not
qualify a cost window, a Core 05 claim, untested deployment confinement profiles,
or the remaining declared-asset path. Coordinated adoption remains open.

Both explicit owner audits passed with native 0:
`remediation-container-redaction-reporting-audit-20261015-01` (2.686 s) and
`remediation-container-redaction-composition-audit-20261015-01` (2.650 s).
The audit confirms the selected operational evidence scope and compatibility;
it does not establish specification adoption.

## Frozen template images and preview receipts

Reporting now owns a copied, immutable template registry separate from Reference
Packs. Release and preview admission retain the complete selected manifest and
exact declared file bytes. Migration 89 adds an independent manifest digest to
the job row, after rejecting incompatible pre-production history before mutation.
Replay, completion and recovery validate this original binding; they do not
resolve a current registry or infer missing historical provenance. The runtime
assembly accepts only trusted typed pack values. It exposes no public installer,
filesystem resolver, URL loader or raw Evidence-image capability.

The implemented pack media are PNG, JPEG and passive SVG under
`assets/{asset_id}/{filename}`. Closed declarations enforce exact size/hash,
unique IDs/source paths/bundle paths, ordering, requiredness and operative
limits. Raster headers must match the declared media and bound decoded geometry;
SVG uses the existing passive subset. Asset decisions retain typed public
template-boilerplate declarations and support allow/drop only. Dropped bytes
never enter the bundle. Canonical source encodes UTF-8 path segments as URL
octets, while transport and bundle identities retain exact NFC paths.

Renderer contract 19 and private stage v4 transport exact binary bytes as
canonical base64. Asset framing v2 binds path, media and raw octets. Private
attempt storage admits only declared assets/manual diagrams, with independent
count/size/total bounds. Retired v3/content transports reject. Native replay
recomputes asset membership and bytes; native persistence, portability and
recovery share role/path/media and self-contained-content admission. Table 22-B
now explicitly identifies the already-produced profile-view, deck-model and
diagram-model artifacts. A role alone never authorizes arbitrary bytes or paths.

Local CSS/font themes remain an open implementation obligation. The image
implementation and default installed Slidev theme do not demonstrate that
general template feature. Other promotion conditions remain open; this
checkpoint cannot adopt Reporting. External Network Flow graph denial is the
explicit current §15 security policy, exercised by the graph cases, rather than
an obligation to add an external allow rule. Its distinct new acceptance ID is
`RPT-AC-GRAPH-006`; the existing selection-order ID `RPT-AC-GRAPH-004` remains
unchanged. The previously duplicated ID was a specification defect.

Public asset validation exposed a separate existing Composition defect. The
generic extension finalizer replaced the immutable preview response with a
terminal common-job resource. The old test accepted either response shape.
Composition now owns a receipt reconciler that compares the original response
with its immutable preview admission. Application assembly explicitly routes
preview receipts to that owner and ordinary snapshot/release receipts to their
common-job response adapter. Unknown routes reject. Terminal success, failure,
cancellation and timeout use this same owner boundary. The preview route
continues to return its original Table 6-I response; execution status belongs to
the job route. There is no response alias or legacy reader.

The stronger preview test compares the complete first/replayed response and
rejects changed attempt/snapshot identities, omitted nullable members, unknown
job-response members and altered output options. Its ordinary application
fixture exercises authenticated release/preview admission, idempotency, frozen
image bytes, tampered manifest/digest/bytes rejection and backup/restore with no
current template registry. It is not native browser rendering evidence.

Retained roots below are under `.cartulary/test-results/`.

| Root | Native result and applicability |
| --- | --- |
| `remediation-bundle-members-unit-20261015-01` | 0, 4.705 s; exact source-contract row, role/path/media admission. |
| `remediation-template-assets-unit-20261015-01` | 2, 7.185 s, product/test_assertion_failure; native expected-members comparison omitted local assets and two old hostile fixtures expected the weaker boundary. Repaired comparison and earlier-rejection expectations. |
| `remediation-template-assets-unit-20261015-02` | 0, 3.893 s; metadata/source, frozen payload and stage-admission rows. |
| `remediation-template-assets-owner-20261015-01` | 2, 101.071 s, product/test_assertion_failure; native Unicode asset URL comparison and old TS stage fixture shape. 22/24 units passed; all 39 cleanup steps completed. |
| `remediation-template-assets-lint-20261015-01` | 2, 40.288 s; old manual-SVG fixture lacked v4 media/binary fields. Fixed the fixture. |
| `remediation-template-assets-owner-20261015-02` | 0, 94.887 s; full Reporting owner, including native PNG and Unicode SVG rendering. Later preview-receipt changes require the current owner run below. |
| `remediation-template-assets-lint-20261015-02` | 0, 37.934 s; complete lint. |
| `remediation-template-assets-composition-20261015-01` | 2, 30.858 s; existing recovery helper omitted the independent template digest while restoring a deleted retained input. Restored its original value; no inferred digest. |
| `remediation-template-assets-composition-20261015-02` | 2, 21.166 s, product/test_assertion_failure; exposed the preview receipt overwrite. Four other units passed; all 18 cleanup steps completed. |
| `remediation-template-assets-composition-20261015-03` | 0, 38.834 s; all five owner units passed after receipt repair, all 18 cleanup steps completed. |

The successful Composition source is
`sha256:f105d76ae6854328850315095bf0e6f221ef3f81ff4fdab1f7ff32be09fe9b74`,
graph `sha256:8ab26554f1dd6b277b3432100a6c02004ccd8e103a404be009ac565ad91b0fee`.
Commands use `CARTULARY_TEST_RUN_ID=<root> CARTULARY_OUTPUT_MODE=ci make
test-slice OWNER=module.reportcomposition` or `OWNER=module.reporting` for the
corresponding complete owner, and `make lint` for lint roots. These are dirty
source development checks, not qualified performance windows.

Generation passed in `remediation-template-assets-generate-20261015-01`
(native 0, 10.480 s). Finalizers 01–04 all passed; 03 took 22.504 s and 04
19.847 s. Each used a fresh run ID and `make agent-finalize`; retained-run
maintenance was skipped because `RESULTS_DIR` was unset. Format 07 has a pass
terminal summary but its native process exit was not recovered after context
loss. Format 08 completed with native 0 (6.063 s); it does not retroactively
establish format 07's outer exit. Earlier failed roots remain investigative
evidence and do not establish intermittency across changed sources.

Current-source Reporting, generated projections, portability, lint and explicit
owner-evidence audits remain pending at this checkpoint. A native external-image
case has now been added to exercise the separate fresh-attempt determinism gate;
it has not yet passed.

### Image and receipt validation completed

The subsequent current-source runs below supersede that pending validation
statement. The native external-image case passed through the real installed
renderer and its fresh-attempt byte-comparison gate. It is implementation
determinism evidence, not a Core 05 publication claim or custom-theme evidence.

The Incident Bundles historical-artifact fixture now creates a release with a
frozen public SVG, exports it, imports it into an isolated destination without
the source template registry, and re-exports the exact original bytes. It rejects
altered image content without publishing partial incident or artifact state.
Existing sensitive-artifact omission, reference binding, historical identity,
read-only recovery and immutable-row assertions remain. Both public application
tests share one static pack fixture in `internal/testutil/reportingfixture`;
the native PNG/Unicode-SVG fixture remains independent. No assertion was removed.

Final source:
`sha256:cbbe5f788d8060aa1699abbd21feadf98c4bf9c18a4b83844dc10c2fcc58b9a1`.
All following native exits are 0:

| Exact run root | Scope / result |
| --- | --- |
| `remediation-template-assets-owner-20261015-04` | Complete Reporting owner, 90.672 s, 24/24 units, all 28 cleanup steps complete. |
| `remediation-template-assets-composition-20261015-05` | Complete Composition owner, 35.181 s, 5/5 units, all 18 cleanup steps complete. |
| `remediation-template-assets-portability-20261015-02` | Exact Incident Bundles row `module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8`, including image round-trip; 42.484 s, 3/3 units, all 14 cleanup steps complete. |
| `remediation-template-assets-catalog-20261015-01` | Complete `harness.test_catalog` owner; 36.341 s, 2/2 units, all 11 cleanup steps complete. |
| `remediation-template-assets-portability-lint-20261015-01` | `make lint`, 29.972 s. |
| `remediation-template-assets-reporting-audit-20261015-02` | Exact final Reporting evidence-root audit, 3.216 s. |
| `remediation-template-assets-composition-audit-20261015-02` | Exact final Composition evidence-root audit, 2.831 s. |

Reporting and Composition retain their graph digests above. The portability
graph is `sha256:59c16aea585a7828bcabf89eaaeb447be6ecfbc7ab15da31c3e6b60c63176afe`;
catalog graph is `sha256:22e69bdf7826839811e7b37f694b3a894ad9e659d3d7b4e152bc7e8e52a6d0fc`.
The audit commands used `make test-evidence-audit OWNER=<owner>
EVIDENCE_ROOTS_FILE=<manifest>` with exact manifests
`.cartulary/remediation-evidence/template-assets-final-reporting-20261015.json`
and `template-assets-final-composition-20261015.json`. Cache reuse, where
permitted, is not another fresh rendering or timing observation.

Before the shared fixture extraction, source
`sha256:e2b89aacf5e5f743a44846db72d80174b1b9dff1b653b364aa25542fe847a0e2`
also passed full Reporting (owner 03, 105.884 s), Composition (04, 25.863 s),
portability (01, 37.557 s), lint (03, 34.601 s), `generate-drift` (01, 15.120 s),
`json-shape-check` (01, 5.509 s), and `generated-artifact-policy-check` (01,
1.018 s). Generated inputs did not change in the fixture extraction. Finalizer
05 passed in 19.684 s; the subsequent `portability-finalize-20261015-01` passed
in 19.946 s after format (6.036 s). Both skipped retained-run maintenance because
`RESULTS_DIR` was unset. Markdown root 01 passed in 35.210 s before the later
evidence text and duplicate-ID correction.

The native and portable image paths, receipt repair, and corresponding
adversarial assertions are demonstrated within these contexts. Packaged
deployment qualification completed in
`remediation-template-assets-package-20261015-01`: native exit 0, 363.742 s,
on the same `cbbe5f…` source and graph
`sha256:e9c9e8506c541000eb380536820902f4147296e1c6c8fb6962190b44c4a79870`.
The command was `make standup-package-smoke` with that exact run ID and
`CARTULARY_OUTPUT_MODE=ci`. This qualifies the installed Linux amd64/no-LSM
context exercised by the smoke; it does not qualify unsupported LSM hosts.
Coordinated adoption, remaining
timed workflows, custom themes, browser consolidation and the qualified cost
portfolio are still open. Failed predecessor roots retain their classifications;
these deterministic fixture, URL and receipt defects are not evidence of flakiness.

## Concurrent route admission

The route review exposed an implementation discrepancy under the existing Core
idempotency contract and Composition Table 6-A2: two simultaneous first requests
could both miss their receipt. The controlled regression holds an owner write
until both authenticated requests reach the database, then releases the barrier.
Before the repair, concurrent snapshot admission returned HTTP 500 for one
request instead of the accepted job. This is a reproduced product concurrency
defect, not test intermittency or a new product requirement.

Composition's create, update, retire, freeze and preview paths now acquire the
existing actor/route/scope/transaction admission lock before receipt lookup and
owner mutation guards. Reporting release admission and lifecycle actions use
that same lock. Snapshot admission acquires the identity before starting its
repeatable-read transaction, transfers the lock to that transaction, and drops
the temporary session lock before returning. A failed or uncertain handoff
discards its connection. Unfinished work rolls back with an independent bounded
cleanup context. There are no product action retries or response aliases.

This ordering preserves one consistent source view and observes the preceding
committed receipt. A transaction-level lock SELECT alone would establish a stale
repeatable-read view before waiting, as described by PostgreSQL's
[transaction isolation contract](https://www.postgresql.org/docs/14/transaction-iso.html).
Only Auth owns this lock primitive; the PostgreSQL adapter remains the narrow DB
port. The existing domain-separated identity framing is unchanged.

| Observable obligation | Implementation owner | Exact exercising test and context | Assertion / review disposition |
| --- | --- | --- | --- |
| Simultaneous snapshot/release requests admit one job | Reporting store, Auth admission | `TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration`, existing Composition preview-delegation row; ordinary authenticated application, isolated PostgreSQL and application protocol renderer fixture | Both requests succeed with one job identity; jobs settle successfully. The before-repair snapshot case failed as described above. |
| Concurrent draft mutations replay before stale-version/state checks | Composition store | Same public integration entrypoint, private `assertConcurrentReportingAdmissions` helper | Create/update/freeze/preview/retire return the identical committed response; draft version advances once. Release invalidation similarly replays its original response. |
| Cancellation cannot leak admission locks or commit unfinished receipts | Auth primitive | `TestRouteIdempotencyActorScopedUniqueness`, exact row `module.auth.integration.route_idempotency_actor_scoped_uniqueness_afb930c9c4`; isolated PostgreSQL | Cancel a proven blocked waiter; cancel an acquired unfinished transaction; require no receipt, successful next acquisition, independent actor admission, and zero advisory locks after commit. |

These helpers are support for their named entrypoints, not new executable tests
or separate coverage identities. No existing assertions or browser contexts were
removed. Behavior review follows the existing route contract; execution review
retains the existing owner/runner/fixture routing. Tests and runtime never read
this record. Native image/sandbox evidence remains in its separate contexts.

Retained roots below use `.cartulary/test-results/`; commands use
`CARTULARY_TEST_RUN_ID=<root> CARTULARY_OUTPUT_MODE=ci make <target>`.

| Root | Target / result |
| --- | --- |
| `remediation-concurrent-admission-before-20261015-01` | Exact Composition preview row; native 2, 14.472 s; `product/test_assertion_failure`, concurrent snapshot HTTP 500. Source `sha256:dcdada158d11af4b13996b49b90205def5c993370ab036b8e5661eee6997c395`. |
| `remediation-concurrent-admission-after-20261015-01` | Same row after repair; native 0, 58.208 s. Source `sha256:1e166a95676410c642cee3a2243475a0346a84f5d58ea5264d4c20df81857c27`. |
| `remediation-concurrent-admission-auth-20261015-01` | Exact Auth row above; native 0, 9.753 s. |
| `remediation-concurrent-admission-reporting-20261015-01` | Complete Reporting owner; native 0, 115.239 s. |
| `remediation-concurrent-admission-composition-20261015-01` | Complete Composition owner; native 0, 43.643 s. |
| `remediation-concurrent-admission-lint-20261015-01` | Complete lint; native 0, 44.323 s. |

Final validation source:
`sha256:397c6c50de74133763bca78d2e603fb829805094c9a749f6f66bf9f4bf50a998`.
Reporting and Composition use their full-owner graph digests recorded above.
The exact Auth graph is
`sha256:dff381777da5eaf2808620ce02bf943fc7793472535c0f1afdca8f0e95734340`;
the selected Composition graph is
`sha256:dee4e7969c76cbd49abdbddafab4b286d05e20cfafd868747c957a35d970f3be`.
Format 01 and 02 passed (6.571 s and 6.605 s). Finalizer 01 passed (20.186 s)
before full-owner validation; retained-run maintenance was skipped because
`RESULTS_DIR` was unset. The failed predecessor remains retained; changed-source
passes do not establish an intermittent failure rate. Required adoption and the
other unresolved remediation workstreams remain open.

The final explicit-root evidence audits also passed:
`remediation-concurrent-admission-reporting-audit-20261015-01` (native 0,
3.021 s) and `remediation-concurrent-admission-composition-audit-20261015-01`
(native 0, 2.971 s). They used `make test-evidence-audit OWNER=<owner>
EVIDENCE_ROOTS_FILE=<manifest>` and exact manifests
`.cartulary/remediation-evidence/concurrent-reporting-20261015.json` and
`concurrent-composition-20261015.json`. Catalog owner validation passed in
`remediation-concurrent-admission-catalog-20261015-01` (native 0, 33.089 s);
the helpers introduce no orphaned entrypoints. Markdown passed in
`remediation-concurrent-admission-markdown-20261015-01` (native 0, 10.130 s).

## Local theme profile — implementation awaiting qualification

Reporting REQ-RPT-103e now defines `reporting_theme_css_v1`: a closed palette,
typography and optional logo profile over the installed default theme. Template
authors select presentation tokens; the renderer owns selectors and layout
integration. This intentionally excludes arbitrary CSS, package-defined
components, imports, executable setup and network dependencies. It retains
useful branding while avoiding a second application styling surface coupled to
Slidev internals. No public route supplies template files or runtime capabilities.

The implementation admits exact immutable CSS, WOFF2 and passive image bytes,
requires complete dependencies, and retains all required theme decisions as
public template boilerplate. A denied required member fails closed. Native
rendering loads the declared stylesheet and every declared font before PDF
production; malformed compressed font data must fail in the confined browser.
Header checks bound file/table/decoded sizes without introducing an unconfined
font decoder. Source-only artifacts do not establish browser font-decoding or
rendering evidence.

Toolchain snapshot v2 adds the complete ordered font/configuration inventory:
installed entries and exact retained template stylesheet/font entries share one
canonical content-manifest digest. Historical validation checks complete
template membership, rather than trusting a digest of installed fonts alone.
Private stage v5 and renderer contract 20 identify the expanded behavior.
Migration 90 rejects incompatible retained history before mutation, then marks
the current renderer boundary; no old toolchain reader or inferred inventory is
introduced. The existing clean-cutover/paired-rollback procedure remains.

Adversarial cases cover CSS imports, selectors, hidden-content properties,
expressions, escapes, duplicate declarations, remote/traversal/encoded URLs,
font fallback lists, missing dependencies, malformed WOFF2 headers, altered font
inventories and required-member redaction. Public release/preview and portability
fixtures now include stylesheet and font bytes. Native internal and external
theme rendering plus corrupted-font rejection have been added to the existing
installed-sandbox entrypoint but have not yet passed at this checkpoint.

The inert fixture font is copied from
`apps/web/public/assets/fonts/atkinson-hyperlegible/AtkinsonHyperlegible-Regular.woff2`
into `internal/testutil/reportingassets/`, with its adjacent `OFL.txt` retained.
Its exact SHA-256 is
`2df4ba17804bc7a36f123127966075d8427bff2df58d0d76820c1130bb1a4150`
(23,196 bytes). The shared support package embeds only the font bytes and returns
a defensive copy; it does not consume license prose or other documentation.
Native and application fixtures remain distinct execution contexts.

Initial validation roots under `.cartulary/test-results/`:

| Root | Result |
| --- | --- |
| `remediation-template-themes-format-20261015-01` | Native 0, 6.587 s. |
| `remediation-template-themes-format-20261015-02` | Native 0, 6.533 s. |
| `remediation-template-themes-generate-20261015-01` | Native 0, 10.120 s; migration ownership and downstream projections. |
| `remediation-template-themes-unit-20261015-01` | Native 2, 26.575 s; theme output exposed a missing `local_theme` case in retained presentation binding. The fix compares its exact declared bytes alongside the other prepared artifacts. |
| `remediation-template-themes-lint-20261015-01` | Native 2, 43.471 s; staticcheck S1009 found a redundant nil check in the font inventory validator. All 20 cleanup steps completed. |
| `remediation-template-themes-format-20261015-03` | Native 0, 6.504 s after both repairs. |

The failed unit/lint source was
`sha256:21fdd85d2a6ebd94eafd0041a896e7457643b1a80f76f53eda55e2f41e60d0fb`.
Unit selection used the existing Reporting metadata, frozen release payload,
stage admission and stage result rows. Current unit rerun, lint, native rendering,
ordinary application/recovery, portability, migration, catalog and finalizer
qualification remain required. These implementation edits do not close adoption.

### Initial theme qualification and final review

The repaired selected unit rows passed in
`remediation-template-themes-unit-20261015-02` (native 0, 3.807 s). Finalizer
`remediation-template-themes-finalize-20261015-01` passed in 19.752 s, with
retained-run maintenance skipped because `RESULTS_DIR` was unset. The subsequent
full Reporting owner passed in `remediation-template-themes-reporting-20261015-01`
(native 0, 113.322 s), full Composition in
`remediation-template-themes-composition-20261015-01` (native 0, 49.365 s), and
lint in `remediation-template-themes-lint-20261015-02` (native 0, 33.875 s).
Source: `sha256:388db525f5834019bd46108d50b0b7015169b0d3b920ac5739786bd10e117f52`.
The native owner exercised internal/external theme rendering, the external
fresh-attempt determinism gate and corrupted-font rejection. The application
fixture exercised immutable theme admission, replay and actual backup/restore;
it remains distinct from native rendering evidence.

Final review identified installed package webfonts and CSS as part of the usable
font surface. Toolchain inventory now conservatively includes those files as
well as the system font/configuration directory, independent of enumeration
order. The exact v2 contract identifies both sets. Theme members now explicitly
reject at the `asset_ref` serializer as well as section-binding admission. The
packaged renderer probe uses private stage v5, and stage transport requires a
complete bounded theme file set. These subsequent changes require the current
qualification recorded next; they are not covered merely by the prior pass.

The metadata row passed after the font inventory/serializer review in
`remediation-template-themes-unit-20261015-03` (native 0, 3.964 s). Format 04
passed in 6.883 s. After the stage file-set check, format
`remediation-template-themes-format-20261016-01` passed in 6.021 s. Current
full-owner, lint, portability, catalog, generated checks and packaged smoke
validation remain pending.

### Final theme qualification

Final theme source:
`sha256:7458bc6e28e85db1244214a37bf54c78118842ebbe8103ec2ead0b998d1a0001`.
Every command below returned native exit 0. The three native sessions temporarily
lost across compaction were recovered from this chat's exact launch records and
drained: package 20425, Reporting audit 94000, Composition audit 47998. No
observer result was substituted for an owning command's exit, and no rerun was
launched to recover that outcome.

Roots below share `.cartulary/test-results/remediation-template-themes-` and
suffix `-20261016-01`:

| Root component | Public target / scope | Duration |
| --- | --- | ---: |
| `finalize` | `agent-finalize`; retained-run maintenance skipped because `RESULTS_DIR` was unset | 19.699 s |
| `reporting` | `test-slice OWNER=module.reporting`, complete owner; 24 units and 39 cleanup steps completed | 123.463 s |
| `composition` | `test-slice OWNER=module.reportcomposition`, complete owner | 48.069 s |
| `portability` | Exact Incident Bundles historical Reporting round-trip row; 3 units and 14 cleanup steps completed | 61.018 s |
| `lint` | `lint` | 39.727 s |
| `drift` | `generate-drift` | 16.027 s |
| `shape` | `json-shape-check` | 5.584 s |
| `policy` | `generated-artifact-policy-check` | 0.942 s |
| `catalog` | `test-slice OWNER=harness.test_catalog`, complete owner | 37.200 s |
| `package` | `standup-package-smoke`; 15 units and 24 cleanup steps completed | 365.270 s |
| `reporting-audit` | `test-evidence-audit OWNER=module.reporting` | 3.166 s |
| `composition-audit` | `test-evidence-audit OWNER=module.reportcomposition` | 3.101 s |

Audits consumed the explicit manifests
`.cartulary/remediation-evidence/themes-reporting-20261016.json` and
`themes-composition-20261016.json`, each naming its exact complete owner root.
Native/application, portability and packaged execution remain distinct contexts.
The package run qualifies the native Linux amd64/no-LSM deployment context only;
it does not qualify another architecture or LSM policy. The source was dirty;
these results are not a clean-source cost portfolio or Core 05 publication.
Later schema changes need their own applicability review and execution evidence.

## Structural redaction simplification

The final clause review found that `split_mixed_block_v1` and its retained
`split_from_block_id` member were never produced by the implementation. Adding a
second fragment identity hierarchy would duplicate the existing declared-field
selection boundary and couple redaction to presentation regrouping. The current
field/child selection already supports independent exact Core rules, ordered
retention, source provenance and empty-container removal. The durable correction
is to make that model the explicit owner contract.

REQ-RPT-061 now specifies `retain_declared_structure_v1`: select against original
addresses, apply each action to the complete scalar or scalar-array field,
reconstruct the same hierarchy, recompute unions, remove empty containers, and
renumber only after selection. It never selects a drop implicitly, splits text
with NLP, or weakens the partition subset predicate. Serializers still reject
an invalid surviving table instead of synthesizing cells or labels. Typed graph
labels keep their separate redaction boundary. The obsolete split algorithm,
member and unused mixed-content reason are retired; no existing executable
split behavior or assurance is removed.

Snapshot v6, export model v7, block v3 and renderer 21 identify the closed shape.
The Extensions participant projection and Core 02 vocabulary are aligned.
Migration 91 rejects incompatible retained history before any mutation; fresh
initialization is the only cutover, with matching application/data rollback.
The previous source-split assertion is replaced by rehashed portable-admission
rejection of both null and non-null retired members, reached through the same
Reporting unit owner context. A new mixed-partition regression checks exact rule
addresses, surviving values/structure/unions and denial of implicit
reclassification. Behavioral review: Reporting §§9, 12 and 13. Execution review:
existing `TestRedactionProfilePrecedenceActionsAndManifest_Unit` and source
validation routing; no browser/native context was retired. This is a proposed
specification simplification under the user's approved scope, not evidence that
the earlier splitting promise was implemented. Current validation is pending.

### Removal of inactive aggregate declassification

The same review found no aggregate-public proof implementation, no admitted
aggregate computation input, and no positive proof tests. The advertised block
flags, template allowlists and profile time-bucketing setting were therefore
ineffective declarations. Implementing a new privacy declassification subsystem
is not necessary to preserve source facts, ordinary redaction, rendering or
portable history. Under the approved preference for simpler durable contracts,
this amendment retires that optional declassification surface and requires the
conservative contributor union for all current output.

REQ-RPT-060–060c, 103c and the current schema tables now reject all retired
members, including false, null and empty values. The serializers cannot treat an
author assertion or a content digest as disclosure proof. Graph Projection and
Network Flow aggregation remain unchanged and do not gain declassification
permission. No current declassification claim is carried forward, and this is a
scope decision, not a deferred implementation obligation or a claim that the
old proof algorithm passed.

The clean shape is template manifest v3 / built-in template `@4`, authored
redaction profile v3 / normalized view v2 / built-in revision `5`, together with
the snapshot/export/block versions above. Migration 91 replaces the preceding
current-shape constraints atomically after its no-history preflight. It neither
converts history nor relaxes the retained admission binding. Core 01's proposed
companion paragraph now identifies the correct built-in tuple; its adopted
predecessor remains authoritative until coordinated promotion.

Retired scenarios `RPT-FIX-006`, `035`, `070` and acceptance IDs
`RPT-AC-PART-002`, `RPT-AC-TIME-002`, `RPT-AC-TEMPLATE-003` are replaced by the
explicit current denial/retention cases `RPT-FIX-081..083`,
`RPT-AC-PART-006`, `RPT-AC-TIME-003` and `RPT-AC-TEMPLATE-005`. The invalid
calendar assertion remains required. The previous source-aggregate and excluded
field assertions survive as rehashed unknown-member rejection in
`testRetainedMixedContent`; `testNormalizedRedactionProfile` and
`testTemplateManifestContract` reject retired settings through portable admission.
These private helpers keep their existing exact owner entrypoints; no test
selection or native/browser context is removed.

Initial structural-redaction roots use `.cartulary/test-results/` and prefix
`remediation-structural-redaction-`, dated `20261016`:

| Root component | Result |
| --- | --- |
| `format-01` | Native 0, 6.011 s. |
| `generate-01` | Native 2, tool exit 1, 4.126 s; `harness/unknown_failure`. Authored Extensions contribution retained the prior participant digest after its output schema changed. Recomputed the canonical authored binding; no validator bypass. |
| `unit-01` | Native 0, 4.142 s; three selected Reporting rows. This ran after incomplete generation and is limited unit evidence, not generated-contract closure. |
| `generate-02` | Native 0, 9.940 s after binding repair. |
| `format-02` | Native 0, 6.507 s after removal of inactive settings. |
| `generate-03` | Native 0, 10.050 s. |
| `unit-02` | Native 0, 4.403 s; five selected Reporting rows, including metadata, strict decoder and retained-structure cases. |

The later database-constraint alignment requires final generation, finalization,
current full-owner and portability qualification. Earlier failed generation is
retained with its exact tool summary; it is not an intermittent outcome.

### Structural redaction validation

The database constraints were aligned before the following frozen-source window:
`sha256:e3b2a2aa96d83000a2cbc8230cdd27ad62d60191f903e9833f84b483fbd85618`.
Roots have prefix `.cartulary/test-results/remediation-structural-redaction-`
and suffix `-20261016-01`, except the final generation root numbered `04`.
All commands below returned native exit 0:

| Root component | Public command and result |
| --- | --- |
| `generate-04` | `make generate`, 10.160 s. |
| `finalize` | `make agent-finalize`, 20.326 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `reporting` | Complete `make test-slice OWNER=module.reporting`, 128.054 s; 24 units, all 28 cleanup steps completed. Changed Go rows ran freshly; compatible unaffected cache hits remain cache evidence. |
| `composition` | Complete `make test-slice OWNER=module.reportcomposition`, 51.423 s. |
| `portability` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8`, 60.507 s. |
| `extensions` | Complete `make test-slice OWNER=module.extensions`, 79.048 s. |
| `lint` | `make lint`, 40.460 s. |
| `catalog` | Complete `make test-slice OWNER=harness.test_catalog`, 37.700 s. |
| `drift` | `make generate-drift`, 17.395 s. |
| `shape` | `make json-shape-check`, 6.222 s. |
| `policy` | `make generated-artifact-policy-check`, 0.925 s. |
| `reporting-audit` | `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/structural-reporting-20261016.json`, 3.192 s. |
| `composition-audit` | `make test-evidence-audit OWNER=module.reportcomposition EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/structural-composition-20261016.json`, 3.167 s. |

These roots establish current schema, native renderer, ordinary application,
retained-input/recovery and portability behavior in their selected contexts.
They do not establish a clean-source cost portfolio, Core 05 publication, or
coordinated adoption. The package root is separately recorded after completion.

The same-source `remediation-structural-redaction-package-20261016-01`
subsequently completed `make standup-package-smoke` in 379.990 s, native exit 0.
It qualifies the packaged Linux amd64/no-LSM context, including the actual
renderer and fresh initialization. It does not qualify another platform or LSM
policy. The Composition root completed all 18 cleanup operations and the
portability root all 14. Their evidence roots and cache/fresh distinctions remain
explicit; no missing process result was inferred from an observer.

### Current output capability review

The current implementation rejects rendered PNG/PPTX requests, but template
admission still allowed an unsupported capability declaration. The candidate
specification now makes the current renderer support boundary explicit: PDF,
SVG and the valid internal source-only path. PNG/PPTX remain false capability
members with deterministic rejection, not promised execution paths. Existing
PNG/JPEG/SVG template images are unaffected. A future renderer profile must own
execution, output validation and determinism before enabling another format.

`RPT-FIX-052` and `RPT-AC-SLIDEV-003` are replaced by `RPT-FIX-084` and
`RPT-AC-SLIDEV-004`; the source-byte invariance obligation is preserved for
supported variants. Template admission now rejects unsupported PNG/PPTX declarations. The source
projection regression compares supported PDF and internal-source-only bytes. The Core 01 release hash rule
now imports the exact adopted Reporting manifest owner instead of copying an
already stale version. The duplicated proposed `RPT-AC-TEMPLATE-004` identifier
was corrected to `RPT-AC-TEMPLATE-005`; the pre-existing timeout criterion keeps
its identity. None of these editorial repairs implies adoption.

### Output capability and native diagram validation

All following public commands returned native exit 0. Roots share prefix
`.cartulary/test-results/remediation-output-capabilities-` and suffix
`-20261016-01`:

| Root component | Command / result |
| --- | --- |
| `format` | `make format`, 6.615 s. |
| `generate` | `make generate`, 10.240 s. |
| `unit` | `make test-slice OWNER=module.reporting ROWS=module.reporting.unit.build_export_model_uses_reporting_owned_metadata_bdda4b2393,module.reporting.unit.frozen_output_options`, 4.041 s. |
| `finalize` | `make agent-finalize`, 20.395 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `lint` | `make lint`, 28.395 s. |
| `native` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.renderer_installed_sandbox`, 127.327 s. |

The final source binding is
`sha256:4b505db939be2d30616c0ca07412e67c9146163a7be819c38340a3096b90f1b9`.
Native graph identity is
`96ee898e2349e6e09c0d834136e3fe952e66e34cbde498c7236bb09bc5dbb156`.
The installed-sandbox entrypoint now exercises external Mermaid bundles with
one and three diagrams through the real installed renderer, checks complete
source/SVG membership and passive SVG output, validates exact candidate
binding, and requires external deterministic bundle identity. This supplements
the earlier package context; it does not claim another architecture or LSM.
Reporting and Composition remain proposed pending the complete coordinated
promotion review and remaining applicable product assurance.

## Current owner qualification before coordinated adoption

The full current owner slices passed at executable source
`sha256:5deb44873aba69184685552f34fd227f60b97fc91bef51e59ded210528b04658`:

| Run root under `.cartulary/test-results` | Public operation | Native result |
| --- | --- | --- |
| `remediation-adoption-reporting-20261016-01` | `make test-slice OWNER=module.reporting` | Exit 0, 128.013 s, 24/24 work units, all 39 cleanup operations completed, zero cache hits. |
| `remediation-adoption-composition-20261016-01` | `make test-slice OWNER=module.reportcomposition` | Exit 0, 39.635 s, 5/5 work units, all 18 cleanup operations completed, zero cache hits. |
| `remediation-adoption-reporting-audit-20261016-01` | Explicit Reporting root manifest `.cartulary/remediation-evidence/adoption-reporting-20261016.json` | Exit 0, 2.736 s. |
| `remediation-adoption-composition-audit-20261016-01` | Explicit Composition root manifest `.cartulary/remediation-evidence/adoption-composition-20261016.json` | Exit 0, 2.710 s. |
| `remediation-adoption-harness-20261016-01` | `make harness-contract` | Exit 0, 52.508 s, 9/9 work units. |
| `remediation-adoption-markdown-20261016-01` | `make lint-markdown` | Exit 0; this precedes this evidence annotation. |

The current output-scope editorial review also removes misleading PNG/PPTX
execution wording from the algorithm pipelines and headmatter/page-count rules.
The bundle-role entries now explicitly state future-only rejection. This aligns
the text with the previously adopted candidate decision and current negative
request/template tests; no new renderer capability or compatibility path was
added. PNG/JPEG/SVG template images remain separate supported assets.

These owner/audit results do not substitute for the complete companion-clause
review, the governing immutable adoption record, or outstanding Core performance
criteria. Reporting 1.4.0 and Composition 1.3.0 remain proposed at this checkpoint.

## Reporting browser workflow — implementation qualification in progress

Core 01's candidate redaction paragraph was still projecting authored schema v2
and built-in version 4. It now projects schema v3 / built-in version 5 and rejects
retired aggregate/time-bucket fields even when false, null or empty. This is a
candidate companion correction, not an independent adoption.

Core 03 REQ-03-287b and Design §16.5 specify the limited nonmodal snapshot/internal
report job surface. The current implementation adds one retained Reporting owner
to the fixed application composition, exact claimed-profile/contract-major route
admission, generated snapshot/release HTTP clients, captured request recovery,
monotonic same-incident/actor Job validation, exact cancellation and verified
snapshot binding before rendering. It uses default template version 4 and
admission-resolved internal redaction. No external publication, builder, artifact
download route or compatibility reader is introduced. Drawer detachment pauses
observation; reattachment does not retry a command or resume reads automatically.
A stalled status read can be interrupted locally to dispatch cancellation.

Initial retained validation roots under `.cartulary/test-results`:

| Run root | Native result and interpretation |
| --- | --- |
| `remediation-reporting-workflow-generate-20261016-01` | Exit 2; new verification IDs were not ASCII sorted. No valid generation claim. |
| `remediation-reporting-workflow-generate-20261016-02` | Exit 2; new collaborator IDs were not ASCII sorted. Fixed authored order. |
| `remediation-reporting-workflow-generate-20261016-03` and `-04` | Exit 0; generated bindings/topology after owner corrections and added integration tests. |
| `remediation-reporting-workflow-format-20261016-01` | Exit 2; test callback return and non-null assertion lint findings. Fixed test source. |
| `remediation-reporting-workflow-format-20261016-02` and `-03` | Exit 0; 6.166 s and 6.146 s. |
| `remediation-reporting-workflow-unit-20261016-01` | Exit 0; six retained-controller tests, 3.548 s. |
| `remediation-reporting-workflow-types-20261016-01` | Exit 0; frontend typecheck, 20.966 s; precedes the added drawer tests. |
| `remediation-reporting-workflow-unit-20261016-02` | Exit 2; controller tests passed, drawer tests used absent Chai DOM matchers. This is a test implementation defect, not intermittent product behavior. All 12 cleanup operations completed. |
| `remediation-reporting-workflow-authority-20261016-01` | Exit 0; both drawer authority/lifetime tests after replacing unsupported matchers with direct DOM assertions, 3.164 s. |

Real browser timing, rendered inspection and complete integration validation are
still required. These unit results do not close AC-030, AC-046 or coordinated
adoption. The exact recovery command's active generated-Make transport regression
also passed in `remediation-exact-recovery-active-command-20261016-01` (exit 0,
46.264 s); the old private wrapper-spelling assertion was not used as evidence.

Workflow integration follow-up (all roots under `.cartulary/test-results`):

| Run root | Native result |
| --- | --- |
| `remediation-reporting-workflow-finalize-20261016-01` / `-02` | Exit 0, 20.595 / 20.347 s. Retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-reporting-workflow-composition-20261016-01` | Exit 0, 5.587 s; application composition exercises all nine retained owners. |
| `remediation-reporting-workflow-boundary-20261016-01` | Exit 0, 4.721 s; exact source ownership checks. |
| `remediation-reporting-workflow-lint-20261016-01` | Exit 2; generated HTTP types crossed the frontend adapter boundary. Moved the client into `services` and projected its cancellation type through that adapter; no boundary exemption. Other 10 units passed and all 20 cleanup operations completed. |
| `remediation-reporting-workflow-format-20261016-04` | Exit 0, 6.092 s. |
| `remediation-reporting-workflow-lint-20261016-02` | Exit 0, 23.573 s, 11/11 units including TypeScript and import-boundary validation. |
| `remediation-reporting-workflow-unit-20261016-03` | Exit 0, 5.567 s; all eight Reporting controller and drawer tests. |
| `remediation-reporting-workflow-ui-contract-20261016-01` | Exit 0, 4.165 s, 8/8 units; UI contract owner including the new incident-controls selector. |
| `remediation-reporting-workflow-availability-20261016-01` | Exit 2; exact packaged-support test expected only the two earlier profiles. Its expected membership now explicitly includes Reporting major 1, retaining rejection of capability facts. |

The proposed Core 04 companion now defines independent snapshot and report
admission predicates with ordinary grid editing, 25 authenticated sessions,
fixed geometry, matching captured request/job identities and two visible frames.
No qualifying observation has yet been produced for these new predicates.

`remediation-reporting-workflow-availability-20261016-02` passed (native exit 0,
2.331 s), retaining exact standard client profile membership and capability rejection.

### Asynchronous snapshot admission — additional coordinated candidate

Ordinary large-grid browser evidence exposed a synchronous admission defect:
source collection and model construction preceded the returned Job. Core 01
REQ-01-477 and Reporting now distinguish request admission from worker capture.
The request's exact Reference Pack set is still pinned atomically at admission.
The worker reads one repeatable snapshot, records that transaction's capture time,
and persists its complete model under execution fencing. Replay preserves an
already retained capture. An explicit source boundary is never silently advanced;
a change between admission and capture fails the Job with a boundary conflict.
Omission and an explicit boundary are distinct idempotent requests in this new
protocol. Pending requests retain no fabricated capture timestamp or model.

Migration 92 extends the existing Reporting payload row with a separately
validated admission and digest. It rejects pre-cutover history before mutation.
The owner remains included in the existing authoritative Recovery contribution;
there is no additional lifecycle table. The retired snapshot Job kind is v2;
the replacement is `snapshot_reporting.snapshot_create_v3`.

Validation so far:

- `remediation-snapshot-admission-unit-20261016-01`: native exit 0, 8.587 s,
  frozen-payload and worker execution/cancellation rows.
- `remediation-snapshot-admission-integration-20261016-01`: native exit 0,
  65.345 s, all four harness units passed, no cached execution. Selected rows
  include the new source-lock admission/cancellation regression, immutable
  replay/provenance, Reference Pack retention, migration rejection through
  migration 92, and the payload/worker unit rows. Source identity:
  `sha256:fa79c9f8157e7b94044055e027a3613cc0df9aa8cc15ae78f3f1300f3002cb54`.

The initial browser failure remains retained and unqualified. New browser timing,
full affected-owner recovery/concurrency validation, and promotion review remain
open. This record does not adopt Reporting or Composition.

### 2026-10-16 source-capture correction

The large-grid Reporting preparation failed in
`remediation-reporting-capture-size-browser-20261016-01` at capture persistence:
the canonical full-model Job payload was 1,135,460,461 bytes, exceeding the
pinned PostgreSQL driver's message bound. The retained diagnostic contains only
that size and closed stage/cause tokens. The preceding diagnosis run
`remediation-reporting-render-diagnosis-browser-20261016-01` identified the same
stage. Both commands exited 2; all 28 cleanup activities in each completed.
Neither root qualifies render-progress assurance.

The candidate cutover now retains `cartulary.reporting_snapshot_capture.v1`
source facts once, in the originating Job's capture. The published snapshot
independently binds `source_capture_sha256`; sections, record summaries, and
Timeline presentation are derived. Migration 93 rejects incompatible history
before mutation; the snapshot Job kind advances to v4. The snapshot resource
replaces its misleading export-model digest field, and portable artifacts
advance to v3 with `source-capture.json`. No legacy reader is introduced.

Validation so far: `remediation-source-capture-unit-20261016-02` passed (3.681 s);
`remediation-source-capture-integration-20261016-02` passed (77.069 s), selecting
the snapshot integration row and its owned cutover, capture cancellation,
Reference Pack retention, replay and recovery checks. The preceding integration
root failed on the obsolete exact API field assertion, corrected together with
the authored OpenAPI owner. `remediation-source-capture-finalize-20261016-02`
passed (20.310 s) after recording the three reviewed API changes in the existing
2.0.0 change set. Retained-run maintenance was skipped because `RESULTS_DIR` was
unset. These are slice results, not coordinated-adoption or large-grid render
qualification. Portable owner validation and asynchronous render preparation
remain required.

### Render preparation separation in progress

The new closed `render_admission.go` and `render_preparation.go` codecs are
currently **not wired into production persistence or dispatch**. The current
release/preview admission path still expands the snapshot synchronously. Do not
claim render-progress closure from these codecs. The authored redaction policy
selector has been separated from its snapshot-dependent rule materialization;
existing materialized profiles remain byte-compatible in the exercised fixture.

`remediation-render-recipe-unit-20261016-05` passed (3.990 s): the recipe derives
the same render model and digest as the existing fixture, does not mutate its
authored admission, rejects substituted capture/identity/request fields and
unknown algorithms, checks its independent digest and preparation binding, and
honors cancellation before preparation. Earlier recipe roots 02 and 03 exposed,
respectively, private transport-byte serialization and an overbroad test check
that confused template sections with expanded case sections; both are corrected.
Root 04 passed admission checks; root 05 adds derivation checks. Root 01 compiled
the new code and exercised the existing codecs, before the new subtest existed.

The remaining integration must keep admission small: retain the authored policy,
source-capture digest, template bytes, composition and graph bindings, and fixed
identity/time before returning the Job. Persist only a derived preparation seal,
not another expanded source model. Compare that seal on restart. Perform source
materialization and candidate validation outside the Jobs transition lock, then
recheck independent admission/preparation bindings inside publication. A failure
before a materialized profile/model exists must retain Job diagnostics without
inventing those digests or a release artifact. Update release/preview Job epochs,
clean-cutover constraints, cancellation/recovery tests and companion owner text
together. These pending decisions do not promote either proposed subsystem.

### Render recipes wired into production — 2026-10-16

This checkpoint supersedes the preceding unwired-codec status. Release and
preview v4 Jobs now retain a closed recipe before returning admission. Source
expansion, recipient membership, materialized-profile validation and model
hashing run under cancelable worker supervision. Preparation retains only an
independently hashed seal. One total-attempt deadline includes preparation,
participant invocation, rendering, candidate validation and publication.

Publication accepts an opaque detached candidate validated before the Jobs lock;
commit rechecks recipe/preparation bindings. Unprepared failure/cancellation
retains owner diagnostics without inventing a release or model/profile digest.
Prepared failures retain the sealed values. Recovery validates pending,
unprepared, prepared and terminal states separately. Migration 94 rejects
incompatible retained history before mutation. Rollback requires a matching
application/data pair. No database reset was performed.

The retired full-model release/preview readers were removed. The existing
`TestReportingFrozenReleasePayload` operational identity now exercises the
recipe, independent Jobs bindings, capture/row identity, exact request and
policy digests, expiry semantics and explicit rejection of retired full-model
inputs. Materialized-model tampering remains checked by source-binding,
participant and candidate validation. No test entrypoint was removed.

Fresh evidence (native exits observed; roots under `.cartulary/test-results/`):

| Run | Scope and result |
| --- | --- |
| `remediation-render-recipe-unit-20261016-08` | PASS, 3.671 s; recipe/capture codecs. |
| `remediation-render-recipe-unit-20261016-09` | PASS, 5.384 s; wired lifecycle compiles, codec assertions. |
| `remediation-render-recipe-unit-20261016-10` | PASS, 7.292 s; worker preparation drain/cancellation and shared deadline checks. |
| `remediation-render-recipe-unit-20261016-11` | PASS, 7.348 s; retired reader removal and replacement assertions. |
| `remediation-render-recipe-integration-20261016-01` | FAIL, 56.989 s, product/test_assertion_failure; graph-digest tamper fixture rejected by the new database constraint before its expected recovery check. All 17 cleanup steps succeeded. |
| `remediation-render-recipe-integration-20261016-02` | PASS, 62.642 s; snapshot provenance, cutover through version 93, and live execution boundaries. |
| `remediation-render-recipe-preview-20261016-01` | FAIL, 23.584 s, product/test_assertion_failure; template-digest tamper fixture hit the same earlier constraint rejection. All 14 cleanup steps succeeded. |
| `remediation-render-recipe-preview-20261016-02` | PASS, 26.549 s; real release/preview rendering, tokenized/private artifacts, concurrent replay, template images, backup/recovery and unprepared failure/cancellation. |
| `remediation-render-recipe-finalize-20261016-01` | PASS, 20.501 s; generated projections refreshed. `RESULTS_DIR` unset, so retained-run maintenance skipped. |

Both tamper fixtures accept database rejection or require recovery rejection
when a mutation can be stored. Successful storage-level tampering cases remain
and restore exact original bytes. The next checks add detached-candidate and
changed-seal-at-commit assertions, full affected owner qualification, and a fresh
large-grid Reporting browser measurement. These slice results do not establish
coordinated adoption or close the remaining timing/cost findings.

The next standard service run, `remediation-render-recipe-reporting-services-20261016-01`,
failed after 76.030 s (product/test_assertion_failure; 20 successful cleanup
steps). Its remaining failure was the old omitted-versus-explicit snapshot
boundary replay expectation. The asynchronous capture owner explicitly makes
those distinct requests. The replacement assertion requires conflict for the
changed request, successful original-Job replay for the exact omitted request,
and a stale-boundary conflict without another capture for a new explicit request.
`remediation-render-recipe-idempotency-20261016-01` passed in 12.929 s.

`remediation-render-recipe-reporting-local-20261016-01` passed in 123.896 s,
covering every current Reporting row without service dependencies, including
native rendering, process confinement, renderer admission, frontend workflow,
codec, redaction and worker supervision. Native Make exits were observed.
The selected row list is retained in its run manifest. A subsequently found
contract-generator assertion still named snapshot Job v2; it is corrected to
the adopted candidate v4 identity and awaits its exact owner check. This is a
generated-artifact expectation repair, not a compatibility reader.


## Snapshot publication validation outside the terminal lock — 2026-10-16

The compact capture and render recipe passed current integration before the
next large-grid browser run. `remediation-render-recipe-reporting-services-20261016-02`
passed all five work units in 81.187 seconds with all 20 cleanup steps complete.
Source: `sha256:c6ce939ec3bcd77df730d2d692b9a74d71e58c3938e1199e46a5bc28c0794a99`.
The exact selected rows cover source capture/provenance, execution fencing,
Evidence export, Graph leases, authorization, idempotency and singleton reads.

`remediation-source-capture-portability-20261016-01` failed because the native
reference round-trip assertion queried the retired `export_model_json` column.
The assertion now compares exported references with the independently retained
snapshot admission. Historical artifact and hostile bundle cases passed in the
failed root; that does not make the aggregate successful. The corrected root
`remediation-source-capture-portability-20261016-02` passed all three work units
in 33.898 seconds with all 14 cleanup steps complete. The preview row in
`remediation-render-recipe-preview-20261016-03` passed all three work units in
27.791 seconds with all 14 cleanup steps complete, including zero candidate,
caller-buffer isolation and changed preparation-seal rejection. Both used source
`sha256:6fec885c5a58a3a44fcb25ae85e4536c1aa9ae4a4467332b8d415e16e9bd6f61`.
Subsequent formatting changed source bytes, not the reviewed behavior.

The contract generator regression passed in
`remediation-render-recipe-contracts-20261016-01` (0.849 seconds); Reporting
codec/worker checks passed in `remediation-render-recipe-unit-20261016-12`
(7.407 seconds). Finalizers 02, 03 and 04 passed in 20.149, 20.439 and 20.508
seconds. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
Markdown root `remediation-render-recipe-markdown-20261016-01` failed on an
incomplete failure-table row; corrected root 02 passed in 9.510 seconds. Native
outer exits were observed separately for these runs.

`remediation-reporting-recipe-progress-browser-20261016-01` failed in 288.994
seconds, outer exit 2, `product` / `test_assertion_failure`; 15 of 17 work units
passed and all 32 cleanup steps completed. Snapshot progress passed. Report
progress never started: its preparatory snapshot terminated with
`job_handler_attempts_exhausted`. This root supplies no report-progress claim.
Its source matches the integrated service root above. No action retry, timeout
increase, threshold relaxation or fixture reduction was applied.

Review found incident-sized decoding and repeated capture validation inside
`CompleteSnapshotCreateJobTx`, under the Jobs transition lock and five-second
terminal budget. The repair moves full retained-input validation into supervised
worker preparation. An opaque `SnapshotCandidate` carries only bounded metadata
and a transient PostgreSQL tuple-version fence. Publication locks and compares
that exact version, rechecks small independent admission/Jobs bindings, and
inserts or validates snapshot metadata. Tuple identities are implementation-local
concurrency guards, never retained provenance or public model identities. Changed
payloads invalidate the candidate, including restored original bytes. Recovery
continues to validate the complete capture and independent snapshot digest.

Regression coverage is attached to the existing provenance and worker entries:
empty candidate and changed identity denial, payload substitution with unchanged
embedded hash, stale-proof denial after restoration, fresh-proof acceptance,
independent snapshot-digest denial, and cancellation/drain before publication.
Current validation of this repair remains pending below. Reporting and Composition
remain proposed until the remaining measurements and promotion evidence close.


`remediation-snapshot-publication-integration-20261016-01` passed in 67.917
seconds (outer exit 0), exercising current capture/provenance, candidate
replay/tampering, migration rejection and execution fencing. Unit root 01 failed
because the new snapshot cancellation fixture injected cancellation on the second
observation, before snapshot preparation started. Its successor fixture triggers
cancellation only after the preparation-start channel closes, independent of
observation count; this was a deterministic test setup error, not instability.


## Qualified Reporting progress after snapshot publication repair

Both ordinary Reporting mechanisms now have fresh passing run roots on source
`sha256:13b270115003b640e3881c2f375cb3821d95ee394026de51efdccaf75577e831`:

| Predicate | Exact run | Observation | Execution |
| --- | --- | --- | --- |
| `perf.reporting_render.visible_with_cancel.v1` | `remediation-snapshot-publication-report-progress-20261016-01` | 200.8 ms, one sample, qualified, scheduler overlap zero | Native exit 0, 209.781 seconds, 15/15 units, all 28 cleanup steps complete |
| `perf.reporting_snapshot.visible_with_cancel.v1` | `remediation-snapshot-publication-snapshot-progress-20261016-01` | 177.4 ms, one sample, qualified, scheduler overlap zero | Native exit 0, 198.905 seconds, 15/15 units, all 28 cleanup steps complete |

Both use the ordinary claimed-profile UI, authentication and authorization,
large-grid fixture, 25 authenticated sessions, prescribed background traffic,
anchored Timeline edit and visible two-frame progress/cancel predicate. Report
preparation creates its real large-grid snapshot before measurement. Neither
observation retries the measured action. Cancellation and terminal cleanup are
retained separately. These observations establish admission responsiveness and
continued editing, not complete-render latency, arbitrary-size export support,
a clean-source execution-cost portfolio or Core 05 publication qualification.

Snapshot publication integration passed in 67.917 seconds; corrected worker and
codec regression root `remediation-snapshot-publication-unit-20261016-02`
passed in 7.135 seconds. Finalizer
`remediation-snapshot-publication-finalize-20261016-01` passed in 20.761 seconds;
maintenance was skipped because `RESULTS_DIR` was unset. Generated drift,
artifact policy and JSON shape roots `remediation-render-recipe-{drift,policy,shape}-20261016-01`
passed in 14.689, 0.897 and 5.214 seconds respectively. All native exits were 0.

## Render recipe owner qualification

The final Reporting and Composition owner runs and their explicit evidence audits
passed at executable source
`sha256:dbc9f7227f91a94844ff24e63d47f926c6af091b4f6d6163783339a104453834`.
Every native exit below was observed separately from the published summary.
All roots are under `.cartulary/test-results/`.

| Exact run ID | Public operation | Result and applicability |
| --- | --- | --- |
| `remediation-render-recipe-cleanup-generate-20261016-01` | `make generate` | Exit 0, 10.500 s; removed the retired full-model SQL query projection. |
| `remediation-render-recipe-cleanup-format-20261016-01` | `make format` | Exit 0, 6.715 s. |
| `remediation-render-recipe-cleanup-finalize-20261016-01` | `make agent-finalize` | Exit 0, 20.730 s; retained-run maintenance skipped because `RESULTS_DIR` was unset. |
| `remediation-render-recipe-cleanup-lint-20261016-01` | `make lint` | Exit 0, 42.747 s; includes the exact backend import-boundary policy. |
| `remediation-render-recipe-cleanup-reporting-20261016-01` | `make test-slice OWNER=module.reporting ROWS=<all non-browser rows>` | Exit 0, 145.283 s; 27/27 work units and all 31 cleanup operations complete. Exact selected identities remain in the run manifest. |
| `remediation-render-recipe-cleanup-composition-20261016-01` | `make test-slice OWNER=module.reportcomposition` | Exit 0, 52.507 s; 5/5 work units and all 18 cleanup operations complete. |
| `remediation-reporting-complete-owner-20261016-01` | `make test-slice OWNER=module.reporting` | Exit 0, 193.300 s; 39/39 work units and all 41 cleanup operations complete. Snapshot and report progress qualified at 173.7 ms and 179.1 ms respectively, each one fresh observation and zero measured overlap. |
| `remediation-render-recipe-reporting-audit-20261016-01` | `make test-evidence-audit OWNER=module.reporting EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/render-recipe-reporting-20261016.json` | Exit 0, 2.902 s; exact full-owner source, selection, results and cleanup audit. |
| `remediation-render-recipe-composition-audit-20261016-01` | `make test-evidence-audit OWNER=module.reportcomposition EVIDENCE_ROOTS_FILE=.cartulary/remediation-evidence/render-recipe-composition-20261016.json` | Exit 0, 2.710 s; exact full-owner audit. |
| `remediation-render-recipe-final-portability-20261016-01` | `make service-backed-test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | Exit 0, 37.486 s; native snapshot Reference Pack bindings and portable retained artifacts. |

The preceding packaged smoke,
`remediation-render-recipe-package-20261016-01`, passed with native exit 0 in
396.639 s, 15/15 work units and all 24 cleanup operations complete. It exercised
fresh initialization and actual packaged rendering on Linux amd64 under WSL2,
with the declared no-LSM confinement profile. Its executable source was
`sha256:13b270115003b640e3881c2f375cb3821d95ee394026de51efdccaf75577e831`.
Later changes removed dead private query helpers, simplified equivalent struct
conversions, admitted exact existing owner-port imports after file extraction,
and corrected a test-only relational Composition fixture. They do not change
packaged renderer behavior. This is a reviewed applicability statement, not a
claim that this older root satisfies a current-source machine evidence audit.
No broader platform, formal cryptographic certification, or Core 05 publication
claim follows.

The failed roots remain retained. `remediation-snapshot-publication-composition-20261016-01`
failed because a relational tuple fixture still inserted the retired
`export_model_json` column; it now uses `source_capture_sha256` without inventing
historical content. `remediation-snapshot-publication-lint-20261016-01` and
`remediation-snapshot-publication-boundary-diagnosis-20261016-01` identified unused
private query helpers, equivalent struct conversions, and missing exact import
permissions after the owner extraction. The policy now names only the actual
Reference Data facade and Revisions boundary consumers; no wildcard or direct
foreign-table access was added. These deterministic repairs do not establish
intermittent failures or a retry-to-green rate.

This checkpoint closes current Reporting execution and progress evidence. It
does not by itself adopt the documents, close companion review, qualify a cost
portfolio, or establish the remaining Evidence and Reference Pack predicates.
