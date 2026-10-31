# networkflow Module Refactoring Tracker and Handoff

## Descriptor sample-count correction — S82–S84 and S67/V59 complete

The maintainer authorized S82–S84 and renewed S67/V59 qualification on
2026-10-31, without commit or deployment. NF-REQ-089 already requires every
column to sample the same structurally valid preview records. The shared
reader omitted cross-column count equality, and the fixed-bound positive control
incorrectly expanded only the first column. F71/V65, F72/V66, F73/V59 and S67/V59
were reopened and now close on the evidence below; F70/V64 remains resolved. Prior passing results below retain their
original scope and do not establish this invariant. The implementation baseline
is `/tmp/cartulary-sample-count-baseline-ustxslam`; all 7,477 product files matched
the prior freeze at entry and pre-existing worktree changes are preserved.

| Finding / obligation | Required remediation | Status | Acceptance |
| --- | --- | --- | --- |
| F70 / V64 | Reject aliases before typed decoding at every mapping object boundary. | DONE | Alias-only and conflicting dual spellings fail in either order; exact escaped spellings pass. |
| F71 / V65 | Complete approved descriptor admission with equal sample counts; preserve specified candidate defaults. | DONE | Equal counts, fixed bounds and contextual counts remain distinct; no approved repair. |
| F72 / V66 | Correct the fixed-bound control and add sample-count regressions to approved/preview/apply/inspection tests. | DONE | Valid controls, retained red evidence, exact errors, no forbidden effects and stable replay. |
| F73 / V59 | Replace the overbroad closure claim and requalify frozen source. | DONE | All 37 product rows, generator, source/lint checks, frozen comparisons and handoff pass. |

| Slice | Dependencies | Status | Exit / principal risk |
| --- | --- | --- | --- |
| S82 — reconciliation and characterization | Revalidated baseline | DONE | Valid zero/one/50 controls and retained mismatch failures; avoid unrelated invalid samples/counters. |
| S83 — shared-reader correction | S82 | DONE | Intrinsic equality without preview-only requirements or repair; preserve specific errors. |
| S84 — boundary qualification | S83 | DONE | Exact apply errors, zero source access/publication, stable replay and contextual counts. |
| S67 / V59 — renewed qualification and handoff | S84 | DONE | Exact 37 rows, generator and checks on frozen source; cleanup/scans and unchanged comparisons. |

### S82 characterization evidence

`nf-samples-s82-format-20261031-01 make format` passed native 0. The genuine
50-record controls for both Cisco profiles sample every column, with differing
valid empty-value counts. Zero/one/50 and single-column controls remain valid.

With the pinned environment and cache off, these commands ran with the named
fresh `CARTULARY_TEST_RUN_ID`, rooted under `.cartulary/test-results`:

| Run ID | Exact command | Native exit |
| --- | --- | --- |
| `nf-samples-s82-unit-red-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.approved_mapping_variants_reject_malformed_membe_c346d73f00,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.source_inspection_header_only` | 2 |
| `nf-samples-s82-integration-red-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.approved_mapping_apply_rejects_forbidden_variant_c2e96f179f` | 2 |

Both report `product/test_assertion_failure` at source
`sha256:8de07a319565dd434dfc4ae92dba28e0035cc8f2fae98b4793571387d193f34f`.
Approved admission accepted mismatched counts, and apply preparation opened the
source. Public apply accepted malformed approvals and published tables, rows and
audit, instead of returning the required owner error. Preview and source
inspection rows passed, confirming their existing contextual count checks.
Bounded diagnostic inspection established the intended product defect; earlier
alias/presence behavior was not weakened. Cleanup completed all 10 unit and 14
integration steps, and both retained-secret scans passed. Native failures remain
retained; later passing evidence will not replace them.

### S83 shared correction

The shared descriptor reader now derives the expected sample count from the
first column position and rejects every unequal subsequent count. Zero remains
valid. Fixed bounds, per-column empty-value counts, preview actual-count checks
and inspection zero-count checks remain independent. No reader signature,
serializer, fingerprint, schema, generated projection or error translation changed.

`nf-samples-s83-unit-20261031-01` ran the exact S82 three-row unit command and
passed native 0 at source `sha256:0b11427c205c1ca7b42b91b690da62357ed05b19928b6179f5a3b7c0d8e13401`. All three rows passed;
cleanup completed all 10 steps and retained-secret scanning passed. The approved
corpus now rejects inconsistent first/middle/last counts and rejects before
source access while keeping the valid original fingerprint.

### S84 boundary qualification

The exact focused commands below passed native 0, at the same source as S83,
with fresh run IDs, cache off and the pinned environment:

| Run ID | Exact command |
| --- | --- |
| `nf-samples-s84-unit-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771` |
| `nf-samples-s84-integration-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.approved_mapping_apply_rejects_forbidden_variant_c2e96f179f,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388` |

All three rows passed, including 42 public apply mutation cases across both
Cisco profiles. Inconsistent samples fail with the specified terminal owner
error before source opens, leave Network Flow publication state unchanged, and
replay stably. The preview corpus covers outer/inner/both mappings and all three
descriptor copies; contextual record counts and inspection's zero counts remain
enforced. Cleanup completed all 10 unit and 14 integration steps; scans passed.
S67 now formats and finalizes before freezing and qualifying the full portfolio.

### S67 descriptor-count source freeze

`nf-samples-final-format-20261031-01 make format` and
`nf-samples-finalize-20261031-01 make agent-finalize` passed native 0. Post-finalize
inspection found only the shared reader and three admission/inspection test files
changed from the implementation baseline. Authored schemas, generated projections,
routing, frozen fixtures and dependency files are unchanged. The 7,477-file
product snapshot is `/tmp/cartulary-sample-count-baseline-ustxslam/freeze.json`.
All four owner task guides passed native 0. The final portfolio reuses the exact
37-row selection and generator row recorded below, through machine selections
retained in the baseline directory; no executable verification reads Markdown.

### S67 descriptor-count final qualification

All twelve final roots are
`.cartulary/test-results/nf-samples-final-<group>-20261031-01`.
Commands used `CARTULARY_PREPARATION_POLICY=installed_only`,
`CARTULARY_HARNESS_CACHE_MODE=off`, fresh `CARTULARY_TEST_RUN_ID` values matching
the roots, and pinned Go 1.27.3, Node 24.15.0 and pnpm 10.33.0. PATH selected
`/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin`.

| Group | Exact command | Native exit | Rows | Cleanup steps |
| --- | --- | --- | --- | --- |
| unit | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.approved_mapping_variants_reject_malformed_membe_c346d73f00,module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.mapping_policy_admission_preserves_json_presence_d4e3e0dd31,module.networkflow.unit.network_flow_selector_covers_cisco_sna_required_66f0e756eb,module.networkflow.unit.network_flow_selector_covers_cisco_sna_v1_permit_a2c32abbbf,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.network_flow_selector_covers_every_network_flow_13fc533c41,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.unit.mapping_candidate_admission_rejects_aliases_and_3f0f806bc6,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771` | 0 | 18 | 10 |
| integration | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.approved_mapping_apply_rejects_forbidden_variant_c2e96f179f,module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.mapping_candidate_public_preview_and_approval_re_145ec783fa,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388` | 0 | 12 | 14 |
| browser | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179,module.networkflow.browser.import_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619` | 0 | 3 | 23 |
| client | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 0 | 3 | 12 |
| protocol | `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.mapping_policy_contracts_distinguish_candidate_d_5f524025f1` | 0 | 1 | 11 |
| generator | `make test-slice OWNER=harness.generated_artifacts ROWS=harness.generated_artifacts.unit.decode_contract_rejects_duplicate_jsonmembers_005d1b3b94` | 0 | 1 | 10 |
| generated | `make source-check CHECK=generated` | 0 | 2 | 14 |
| schema | `make source-check CHECK=schema` | 0 | 0 | 12 |
| catalog | `make source-check CHECK=catalog` | 0 | 0 | 11 |
| go | `make lint CHECK=go` | 0 | 0 | 15 |
| types | `make lint CHECK=types` | 0 | 0 | 11 |
| boundary | `make lint CHECK=backend-boundary` | 0 | 1 | 13 |

All **37 product rows** passed: 18 unit, 12 integration, three browser, three
client and one protocol. The generator regression and all required source/lint
checks also passed. Manifest selections match the exact published row outcomes.
Every root completed cleanup and passed retained-secret scans, with cache off
and zero hits. Native command exits were observed independently of summaries.
All twelve roots share:

- Source: `sha256:0b11427c205c1ca7b42b91b690da62357ed05b19928b6179f5a3b7c0d8e13401`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- Environment: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
- Commit: `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`, dirty worktree preserved.

Both NF-FIX-029 revision-5 comparisons passed with unchanged frozen inputs and
expectations:

- `.cartulary/test-results/nf-samples-final-integration-20261031-01/NF-FIX-029-cisco-sna-flow-search-4113496340/comparison-summary.json`.
- `.cartulary/test-results/nf-samples-final-browser-20261031-01/NF-FIX-029-cisco-sna-flow-search-node-kuuVTx/comparison-summary.json`.

Both retain manifest SHA-256
`01998415c8e08ad6d41e4411743a90af8a3bab509df66d41574696cda4c72e01`,
source bundle `896cc8b54d8c0c5675e74848c423f14852f5f207b6bc3674e156e639429a7b83`
and expected bundle `c8baf09449d135a288a382e36749cb2856fe4a3397fab62365f785f3c3272286`.
The final 7,477-file comparison matches the freeze, and staged Git entries match
the pre-session index. Exact commands/native session exits and the operator's
artifact inspection are retained in `native-invocations.json` and
`qualification-review.json` under `/tmp/cartulary-sample-count-baseline-ustxslam`.
The native run artifacts remain the authoritative execution evidence.

F71/V65 maps to the shared collection invariant and approved-reader controls;
F72/V66 maps to S82's retained failures and S83/S84's corrected controls,
pre-source apply rejection, contextual preview/inspection counts and replay;
F73/V59 maps to the reconciled tracker and this compatible final-source portfolio.
F70/V64 remains resolved, and its existing admission coverage passed again.

The fixed-bound control now comes from a real 50-record preview; columns may
have different values and valid empty-value counts. Schema IDs, major 8/state 5,
valid canonical bytes and fingerprints are unchanged. Malformed pending
approvals fail closed and must be recreated from original source through the
supported flow. There is no migration, repair, fallback or reset. Rollback can
restore defective admission and is not a data-repair strategy.

No required product check was skipped. Complete-owner evidence audits, full CI,
release/performance qualification and golden updates remain outside this scoped
portfolio. Optional finalizer retained-run maintenance was skipped because no
compatible full warm-check root was supplied; no redundant run was created.
Product qualification and tracker reconciliation are complete. Closing Markdown
passed native 0 in `nf-samples-final-docs-20261031-01` through
`make lint-markdown`; its tool summary is
`.cartulary/test-results/nf-samples-final-docs-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
F71/V65, F72/V66, F73/V59, S82–S84 and S67/V59 are DONE for this selected
boundary. The final status-only documentation checkpoint is
`nf-samples-final-docs-20261031-02 make lint-markdown`; its native outcome is
retained under that exact run ID. No product files changed after the freeze.
Commit/deployment were not done.

### Prior mapping-admission implementation and qualification

The maintainer authorized implementation of the complete mapping-admission plan
on 2026-10-31, including S79–S81 and final S67/V59 qualification, without commit
or deployment. Baseline is dirty `main` at
`d936506e2b851a22e8c9fbddd983ab4fd6aa21de`; content and index snapshots are retained
in `/tmp/cartulary-mapping-admission-baseline-ncqvfp9n`. All four affected owner
`make task-guide` calls passed native 0. Existing changes are preserved.

The [adopted Network Flow NLSpec](../network-flow-activity-nlspec.md),
NF-REQ-009, 011a/011b, 081, 088a, 089, 090, 092/093 and 097 already govern exact
decoded names, required presence, closed variants and candidate-only defaults.
No new normative behavior, schema ID, contract major or state version is needed.
F66–F69 retain their original demonstrated scope; their prior closure does not
cover member-name aliases or the additional presence/defaulting defects here.

| Slice | Dependencies | Status | Exit / principal risk |
| --- | --- | --- | --- |
| S79 — contract reconciliation and characterization | Revalidated baseline | DONE | Valid controls and targeted failing raw-input evidence; avoid unrelated fixture failures. |
| S80 — complete boundary correction | S79 | DONE | Shared strict readers and focused unit passes; preserve valid identities and specific errors. |
| S81 — public and projection qualification | S80 | DONE | Preview/approval/apply/replay and generated validators pass. |
| S67 / V59 — prior qualification and handoff | S81 | DONE | Compatible final-source portfolio, cleanup, scans, unchanged fixtures and current handoff. |

The approved plan's final portfolio is the historical 33 product rows below,
two new candidate-admission rows, the ordinal fixture unit row and privacy
fixture integration row, plus the existing generator regression and required
source/lint/Markdown checks. Qualification uses fresh run IDs and cache off.
Malformed pending approvals fail closed and must be recreated from original
source through the supported flow; no rewrite, migration, reset or alias reader
is introduced. Valid canonical bytes/fingerprints and NF-FIX-029 revision 5 stay
unchanged. The historical results below are not current completion evidence.

### S79 characterization evidence

Registered `TestMappingCandidateClosedJSONAdmission` as
`module.networkflow.unit.mapping_candidate_admission_rejects_aliases_and_3f0f806bc6`
and `TestMappingCandidatePublicAdmission` as
`module.networkflow.integration.mapping_candidate_public_preview_and_approval_re_145ec783fa`.
Generation and format passed native 0 in `nf-admission-s79-generate-20261031-01`
and `nf-admission-s79-format-20261031-01` under `.cartulary/test-results`.

The exact red roots are `.cartulary/test-results/nf-admission-s79-unit-red-20261031-01`
and `.cartulary/test-results/nf-admission-s79-integration-red-20261031-01`.
Both native Make commands exited 2, classified `product/test_assertion_failure`.
Unit selected the new candidate row plus approved-variant and preview-admission
rows; integration selected only the new public candidate row. Valid preview,
approved round-trip and apply controls succeeded. Alias/presence mutations were
accepted or lost the required shape-error identity; public cases returned 200
instead of 400, including an explicit null policy overridden by a later alias.
The reverse policy order rejected with the wrong policy-specific reason.
The existing preview reader rejected the newly exercised malformed shapes.

Both roots share source
`sha256:0bd494b424b65b6c2003c44580c6cc1d61c2232c2ef6960a06e34e6fa62d9c81`;
cache is off, cleanup completed (10 unit and 14 integration steps), and retained
secret scans passed. Manifests, row results, run summaries and bounded diagnostic
excerpts establish product reproduction, not setup failure. Diagnostic invocation
attempts without TARGET or with a non-public child target were usage errors;
`make explain-run ... TARGET=test-slice DETAIL=logs` resolved the published logs.

### S80 focused correction

Raw admission now lives in one private mapping reader. Candidate and approved
envelopes retain nested JSON until exact closed-shape admission; only absent
candidate defaults materialize. Stored approvals and both preview copies use
`decodeApprovedMapping(raw, limits)`, including the shared descriptor/sample
reader. Removed wire fallback repairs; internal timestamp construction remains
separate. No serializer, fingerprint, authored schema or frozen fixture changed.

`nf-admission-s80-format-20261031-01 make format` passed native 0. The first
eight-row focused unit run, `.cartulary/test-results/nf-admission-s80-unit-20261031-01`,
exited native 2 with one product failure: the Cisco policy-default test supplied
empty required target IDs and omitted required transform/combinability members.
Seven other rows passed, including the full alias/presence corpus, all timestamp
variants, preview admission, approved apply probes and streaming/source checks.
The invalid Cisco controls were corrected to supply those required members;
the intended empty-policy omission remains the only default request under test.
No product admission rule was weakened to accommodate the controls.

Format 02 and the narrow corrected Cisco row rerun
`nf-admission-s80-unit-20261031-02` passed native 0. The rerun source is
`sha256:7ae8b20bcde57e91de48311a9cc7a63fbf8942fa5f4a8660d21e703b39eb9ac4`.
Both focused unit invocations completed cleanup and passed retained-secret
scanning. Their distinct source identities remain recorded; the final 37-row
portfolio will requalify all obligations together after finalization.

### S81 boundary and projection evidence

All focused invocations below passed native 0 with cache off. Exact commands,
selected rows, summaries and identities are retained in each named root under
`.cartulary/test-results`:

| Root | Selection / coverage |
| --- | --- |
| `nf-admission-s81-integration-20261031-01` | Four rows: new public candidate admission; approved apply admission (30 cases); existing Cisco policy-presence and Flow Search policy admission. |
| `nf-admission-s81-protocol-20261031-01` | Existing protocol mapping-policy row, extended to exact nested names, presence, null/type and escaped canonical names. |
| `nf-admission-s81-unit-20261031-01` | Candidate, approved-variant and import-preview rows after whole-object/null/bool/array/string mutations and duplicate-corpus simplification. |
| `nf-admission-s81-unit-20261031-02` | Approved-variant and preview rows after shared descriptor-limit, fixed sample-bound and permitted mapping-row reordering controls. |

The public integration and protocol roots share source
`sha256:bdb4d421aec5b2072362fbfdaaeaa6a57887070fa1dd8181a898733bfd6145a8`.
The last unit root records
`sha256:818607aa2b0d3871910356a4fd9de7d582398d94180bdb8edacdeee27cae167b`.
The last two edits only expanded unit obligations; these focused results are not
asserted to form one frozen-source portfolio. All focused cleanup completed and
retained-secret scans passed. S81 format invocations 01–03 passed native 0.

Public candidate rejection preserves durable state and permits corrected approval
with the rejected transaction ID. Stored-approved mutation keeps the original
fingerprint, opens no source, publishes no Network Flow table/row/diagnostic/audit,
produces the specified terminal owner failure and replays the failed job stably.
Preview outer/inner/both copies reject the common raw mutation corpus. Stored
descriptor admission enforces effective header/column limits and the fixed
50-sample bound; preview additionally enforces its actual record count.

Review against the session snapshot found only the intended four new private/test
files and sixteen changed files (including tracker and generated routing index).
Authored schemas, normative specification, serializer/fingerprint implementation
and frozen fixtures match the pre-session content. Staged entries match the
saved index (`sha256:7cb5f6edab7410a8b52be2d1ab469e78b782685000eabf02dc6c0563741cf588`);
the index file's metadata bytes were refreshed by ordinary Git inspection.

### S67 qualification repair and renewed freeze

Finalization 01 passed native 0. The first complete 37-product-row portfolio and
generator regression passed, as did generated/schema/catalog checks and Go and
backend-boundary lint. Roots are
`.cartulary/test-results/nf-admission-final-<group>-20261031-01`.
Type lint alone exited native 2 (`config/configuration_error`): its published
log identifies TS2304 at three protocol-test calls to `structuredClone`, which
is unavailable in that package's declared TypeScript environment. All original
runs completed cleanup and passed retained-secret scans. The failure is in the
new test helper, not product admission; it is retained at the exact `types` root.

After the first portfolio drained, the test helper was changed to a JSON clone
for JSON-only fixtures. No compiler environment or product rule was broadened.
The first 7,477-file product freeze remains in
`/tmp/cartulary-mapping-admission-baseline-ncqvfp9n/freeze-01.json`. Because this
repair changes the harness source identity, the complete specified portfolio
must be renewed after focused protocol/type checks and finalization; passing
roots from freeze 01 will not be combined with corrected-source roots.

The focused `nf-admission-s81-repair-protocol-20261031-01` test slice and
`nf-admission-s81-repair-types-20261031-01 make lint CHECK=types` passed native 0.
Format and `nf-admission-finalize-20261031-02 make agent-finalize` passed native 0.
Freeze 02 changes only the protocol test file relative to freeze 01. The reviewed
7,477-file snapshot is `freeze-02.json` in the same baseline directory.

### S67 final frozen-source evidence

All final roots use `.cartulary/test-results/nf-admission-final-<group>-20261031-02`.
Every native command below exited 0. Execution used
`CARTULARY_PREPARATION_POLICY=installed_only`,
`CARTULARY_HARNESS_CACHE_MODE=off`, a fresh `CARTULARY_TEST_RUN_ID` matching
the root, and pinned Go 1.27.3, Node 24.15.0 and pnpm 10.33.0. PATH selected
`/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin`.

| Group | Exact command | Native exit |
| --- | --- | --- |
| unit | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.approved_mapping_variants_reject_malformed_membe_c346d73f00,module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.mapping_policy_admission_preserves_json_presence_d4e3e0dd31,module.networkflow.unit.network_flow_selector_covers_cisco_sna_required_66f0e756eb,module.networkflow.unit.network_flow_selector_covers_cisco_sna_v1_permit_a2c32abbbf,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.network_flow_selector_covers_every_network_flow_13fc533c41,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.unit.mapping_candidate_admission_rejects_aliases_and_3f0f806bc6,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771` | 0 |
| integration | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.approved_mapping_apply_rejects_forbidden_variant_c2e96f179f,module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.mapping_candidate_public_preview_and_approval_re_145ec783fa,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388` | 0 |
| browser | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179,module.networkflow.browser.import_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619` | 0 |
| client | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 0 |
| protocol | `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.mapping_policy_contracts_distinguish_candidate_d_5f524025f1` | 0 |
| generator | `make test-slice OWNER=harness.generated_artifacts ROWS=harness.generated_artifacts.unit.decode_contract_rejects_duplicate_jsonmembers_005d1b3b94` | 0 |
| generated | `make source-check CHECK=generated` | 0 |
| schema | `make source-check CHECK=schema` | 0 |
| catalog | `make source-check CHECK=catalog` | 0 |
| go | `make lint CHECK=go` | 0 |
| types | `make lint CHECK=types` | 0 |
| boundary | `make lint CHECK=backend-boundary` | 0 |

The exact selected product rows all passed: 18 unit, 12 integration, three
browser, three client and one protocol, totaling **37 product rows**, plus the
existing generator regression. Selection was checked against both declared
manifest inputs and published row outcomes. Source checks and lint also passed.
Every root completed all cleanup steps and passed retained-secret scanning;
cache mode is off with zero harness cache hits. Native exits were observed
separately from summaries. Each exact root retains manifests, summaries, row
results, cleanup and scan artifacts. All twelve roots share:

- Source: `sha256:05cfd6835de8c5ce211d4e10e5d7949047d6b90e4cee2f3ca3e530ff7746ff41`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- Environment: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
- Commit: `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`, original dirty worktree preserved.

Both NF-FIX-029 revision-5 comparisons passed with unchanged frozen inputs and
expectations:

- `.cartulary/test-results/nf-admission-final-integration-20261031-02/NF-FIX-029-cisco-sna-flow-search-2138698809/comparison-summary.json`.
- `.cartulary/test-results/nf-admission-final-browser-20261031-02/NF-FIX-029-cisco-sna-flow-search-node-Nbcd1R/comparison-summary.json`.

Both retain manifest SHA-256
`01998415c8e08ad6d41e4411743a90af8a3bab509df66d41574696cda4c72e01`,
source bundle `896cc8b54d8c0c5675e74848c423f14852f5f207b6bc3674e156e639429a7b83`
and expected bundle `c8baf09449d135a288a382e36749cb2856fe4a3397fab62365f785f3c3272286`.
Valid approved bytes and canonical fingerprints passed unchanged. The final
7,477-file comparison matches freeze 02 exactly; staged Git entries match the
pre-session index. The operator's artifact-inspection record is
`/tmp/cartulary-mapping-admission-baseline-ncqvfp9n/qualification-review.json`;
the native Make artifacts remain the authoritative execution evidence.

F70/V64 maps to S79–S81 and the new candidate, expanded approved/preview and
protocol rows. F71/V65 maps to the shared raw readers, omission/default controls,
all timestamp variants, descriptor limits and the same admission rows. F72/V66
maps to retained red characterization, both public candidate routes, apply
failure/replay and the expanded generated-validator row. F73/V59 maps to the
controlling tracker correction and this compatible final-source portfolio.

No required product check was skipped. Complete-owner evidence audits, full CI,
release/performance qualification and visual golden updates were outside this
specified portfolio. Optional retained-run maintenance was skipped because no
compatible full warm-check root exists; no extra run was launched to obtain one.
Product qualification and final source comparison are complete. The Markdown
check `nf-admission-final-docs-20261031-01 make lint-markdown` passed native 0;
its summary is
`.cartulary/test-results/nf-admission-final-docs-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
F70–F73/V64–V66, S79–S81 and S67/V59 were marked DONE at that checkpoint.
The subsequent descriptor-count finding reopens the obligations identified above.
The documentation closure checkpoint uses `nf-admission-final-docs-20261031-02`
and the same Make command and summary suffix. Product source remains frozen.

### Handoff and compatibility

Schema IDs, contract major 8, state version 5, valid approvals and published
tables remain unchanged. Only specified candidate omissions receive defaults;
approved admission never repairs input. Malformed pending approvals fail closed
and must be recreated through supported import from original source. There is
no migration, repair script, automatic normalization/reset, alias reader or
compatibility flag. Rollback can restore defective admission; it does not repair
malformed stored input. Release should keep application, assets and environment
matched. Commit and deployment remain separate, unperformed actions.

## Historical approved mapping variant admission — original scope complete

F69/V63 closes the approved-reader gap: a policy member on ignored
or system-derived mappings previously passed stored-mapping admission while the
generated validator and preview-result admission rejected it. Canonical fingerprints omit
that foreign member. NF-REQ-092/093/097 already require closed approved rows;
this correction shares one strict approved-row decoder across both readers.

The maintainer authorized implementation of S76–S78 and S67 requalification on
2026-10-31, without commit or deployment. Baseline is main at
`d936506e2b851a22e8c9fbddd983ab4fd6aa21de`, with 53 existing/relevant files captured
in `/tmp/cartulary-approved-variants-baseline-b2nn0zb8` before edits. Four owner
`make task-guide` calls passed native 0. F66–F68 remain resolved within their
original scope; their historical 31-row evidence does not cover F69.

| Finding / slice | Dependencies | Status | Exit / principal risk |
| --- | --- | --- | --- |
| F69 / V63 | NF-REQ-092/093/097 | DONE | Both approved readers and apply reject malformed variants before source reads or publication. |
| S76 — clarification and characterization | Revalidated baseline | DONE | Valid controls succeed and negative cases reproduce the gap; avoid invalid controls. |
| S77 — shared approved-row decoder | S76 | DONE | Shape checks share one owner; preserve error precedence and valid identities. |
| S78 — boundary validation | S77 | DONE | Generated contracts, public apply rejection and failed replay pass. |
| S67 / V59 — qualification and handoff | S78 | DONE | Exact 31 existing product rows plus two new rows, generator and required checks pass on frozen source. |

Approved wire schema IDs, major 8/state 5, valid bytes/fingerprints and frozen
fixtures remain unchanged. Malformed approved objects fail closed without
migration, rewrite, fallback or reset; recreate affected pending imports through
the supported source flow. This historical S67 qualification records only its
original scope; current acceptance is controlled by the complete mapping-admission
section above. Original results and coverage limits are retained.

### S76 retained characterization

`make author-test-row-id` registered the new selectors through the authored
Network Flow family:

- `module.networkflow.unit.approved_mapping_variants_reject_malformed_membe_c346d73f00` → `TestApprovedMappingFieldVariants`.
- `module.networkflow.integration.approved_mapping_apply_rejects_forbidden_variant_c2e96f179f` → `TestApprovedMappingApplyAdmission`.

`nf-variants-s76-generate-20261031-01 make generate` failed native 2 because the
authored row list was not ASCII-sorted. Its exact failure root is
`.cartulary/test-results/nf-variants-s76-generate-20261031-01`. The retained tool summary identifies
that ordering error; sorting the authored list resolved it.
`nf-variants-s76-generate-20261031-02 make generate` and
`nf-variants-s76-format-20261031-01 make format` passed native 0.

The red unit root `.cartulary/test-results/nf-variants-s76-unit-red-20261031-01`
selected the new unit row plus
`module.networkflow.unit.import_preview_admission_f0dc5f2d8b` through
`make test-slice OWNER=module.networkflow ROWS=<those exact two rows>`.
The new row failed native 2; existing preview admission, including the new
shared mutation corpus, passed. Valid approved round-trip and apply controls
succeeded; malformed approved fields were accepted or lost shape-error identity.

The red integration root
`.cartulary/test-results/nf-variants-s76-integration-red-20261031-01` selected
only the new integration row through `make test-slice OWNER=module.networkflow`.
Native exit was 2. All twelve otherwise-valid cases (two profiles, two non-source
variants, null/empty/default) incorrectly published tables/rows. Their unchanged
fingerprints did not detect the forbidden member. This is product reproduction,
not a setup failure. Both red roots share source
`sha256:ca67adc66701386ccef5ab15da1f8d8db0baed6aea71f1e8b9210facc26a4688`,
completed all cleanup actions and passed retained-secret scans.

The shared test fixture now supplies required source-row `combinability`;
its previous omission was an invalid approved control, not a compatibility
requirement. No frozen expected artifact was changed.

### S77–S78 focused correction

One private `decodeApprovedFieldMappings` now admits raw closed rows for both
approved decoding and preview-result validation. The wire envelope retains rows
as raw JSON; approved domain structs and serialization remain concrete. Required
presence, nullability, member types, positive ordinals, combinability and ignore
reason are checked before semantic admission. Profile policies, target/transform
compatibility, source accounting and derivation semantics remain in the existing
semantic validator. Unknown string discriminators retain `mapping_kind_unavailable`;
shape failures use `variant_member_conflict`; candidate errors remain unchanged.

The production change removes the preview-only copy of the row-shape rules.
Canonicalization/fingerprinting and authored approved schemas are unchanged.
New tests cover both profiles, every approved row variant, raw foreign-member
presence, required members, early apply rejection, all three outer/inner preview
mutation combinations, real Imports failure publication and replay, and generated
candidate/approved rejection of foreign policies. The apply seam changes only
owner mapping JSON after normal approval and keeps the original fingerprint.

`nf-variants-s77-format-20261031-01 make format` passed native 0. All three
`nf-variants-focused-<group>-20261031-01` invocations passed native 0:

- unit: new approved-variant row plus existing import-preview, streaming-apply,
  mapping-policy-presence and Cisco Flow Search profile rows.
- integration: new approved-mapping apply row (12 cases).
- protocol: existing mapping-policy contract row with expanded foreign-policy cases.

Exact selections and identities are retained in each run's manifest; summaries,
rows, cleanup and secret scans are retained under `.cartulary/test-results`.
These focused passes precede finalization and do not close S67.

### S67 frozen-source qualification

`nf-variants-finalize-20261031-01 make agent-finalize` passed native 0. Inspection
found only the intended production, tests, specification, routing and handoff
changes; the sole generated change for this slice is the topology render index.
The authored schemas, approved serializer, canonical fingerprint code and frozen
fixture content remain unchanged. Product freeze records 51 files at
`/tmp/cartulary-approved-variants-baseline-b2nn0zb8/freeze.json`.

Every final root is `.cartulary/test-results/nf-variants-final-<group>-20261031-01`.
All commands used pinned Go 1.27.3, Node 24.15.0 and pnpm 10.33.0, with
`CARTULARY_PREPARATION_POLICY=installed_only`,
`CARTULARY_HARNESS_CACHE_MODE=off`, and the corresponding fresh
`CARTULARY_TEST_RUN_ID`. Go/PATH pointed at the installed pinned toolchain
`/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin`.

| Group | Exact command | Native exit |
| --- | --- | --- |
| unit | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.approved_mapping_variants_reject_malformed_membe_c346d73f00,module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.mapping_policy_admission_preserves_json_presence_d4e3e0dd31,module.networkflow.unit.network_flow_selector_covers_cisco_sna_required_66f0e756eb,module.networkflow.unit.network_flow_selector_covers_cisco_sna_v1_permit_a2c32abbbf,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.network_flow_selector_covers_every_network_flow_13fc533c41,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b` | 0 |
| integration | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.approved_mapping_apply_rejects_forbidden_variant_c2e96f179f,module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015` | 0 |
| browser | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179,module.networkflow.browser.import_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619` | 0 |
| client | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 0 |
| protocol | `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.mapping_policy_contracts_distinguish_candidate_d_5f524025f1` | 0 |
| generator | `make test-slice OWNER=harness.generated_artifacts ROWS=harness.generated_artifacts.unit.decode_contract_rejects_duplicate_jsonmembers_005d1b3b94` | 0 |
| generated | `make source-check CHECK=generated` | 0 |
| schema | `make source-check CHECK=schema` | 0 |
| catalog | `make source-check CHECK=catalog` | 0 |
| go | `make lint CHECK=go` | 0 |
| types | `make lint CHECK=types` | 0 |
| boundary | `make lint CHECK=backend-boundary` | 0 |

All twelve native commands exited 0. The exact selected rows all passed:
16 unit, 10 integration, three browser, three client and one protocol, for
33 product rows, plus the generator regression row. Every run completed its
cleanup actions and passed retained-secret scanning; no harness cache hit was
used. Manifests, summaries, row outcomes, cleanup and scans are retained in each
exact root. All twelve runs share:

- Source: `sha256:bd2b112fe886ef03caddc4768006bc987020187d7671a5b58d1b8eb73ec054ba`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- Environment: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
- Commit: `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`, original dirty worktree retained.

NF-FIX-029 revision 5 comparisons passed at:

- `.cartulary/test-results/nf-variants-final-integration-20261031-01/NF-FIX-029-cisco-sna-flow-search-3975181371/comparison-summary.json`.
- `.cartulary/test-results/nf-variants-final-browser-20261031-01/NF-FIX-029-cisco-sna-flow-search-node-aLgAME/comparison-summary.json`.

Both retain manifest SHA-256 `01998415c8e08ad6d41e4411743a90af8a3bab509df66d41574696cda4c72e01`, source bundle
`896cc8b54d8c0c5675e74848c423f14852f5f207b6bc3674e156e639429a7b83`, and expected bundle
`c8baf09449d135a288a382e36749cb2856fe4a3397fab62365f785f3c3272286`. Original CSV and every frozen fixture file
match their pre-session hashes. Valid approved byte/fingerprint expectations
passed unchanged. Source freeze comparison found no product changes after
qualification; the Git index matches its pre-session snapshot.

F69/V63 acceptance now has passing evidence for raw approved variants, both
preview-result copies, both source profiles, early rejection before source reads,
unchanged approved identities, exact owner failure classification, no owner
publication and stable failed-job replay. Imports' normal failed-job/receipt
bookkeeping remains distinct from prohibited Network Flow publication.

Compatibility remains schema IDs/major 8/state 5. Valid approvals and published
tables remain valid. Malformed pending approvals fail closed; recreate them
through supported import from original source. There is no migration, automatic
rewrite/reset, compatibility flag or legacy reader. Use matching application,
assets and environment on release/rollback; rollback would reintroduce the
admission defect rather than repair malformed data.

No required product check was skipped. Full CI, whole-owner evidence audits,
release/performance qualification and visual golden updates were outside this
specified portfolio. Optional retained-run maintenance was skipped because no
compatible full warm-check root exists. Commit and deployment remain unperformed.
Final documentation validation `nf-variants-final-docs-20261031-01 make
lint-markdown` passed native 0. Its summary is
`.cartulary/test-results/nf-variants-final-docs-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
F69/V63, S76–S78 and S67/V59 were marked DONE at that historical checkpoint.
The evidence covers the original variant-policy scope. It did not qualify
member-name aliases or the additional presence/defaulting defects; S67/V59 was
therefore reopened for F70–F73 in the controlling section above.
The completion-status checkpoint uses `nf-variants-final-docs-20261031-02`
and `make lint-markdown`; its outcome is retained at the same summary suffix
under that exact run root. Product source remains frozen.

## Historical S73–S75 — candidate policy presence and contract correction

The maintainer authorized implementation of F67/F68 and S73–S75 on 2026-10-31,
including final S67/V59 requalification, without commit or deployment. The source
baseline is `main` at `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`, with 47
pre-existing modified/untracked files retained in the local comparison snapshot
`/tmp/cartulary-policy-baseline-d_qd6yvt`. Earlier sections describe historical
qualification, not coverage of these two newly identified cases. S72/F66's
profile-permission correction remains valid and is not reopened.

| Finding / validation | Required correction | Status |
| --- | --- | --- |
| F67 / V61 | Preserve candidate policy presence; reject explicit malformed values before defaulting; prove public-route rejection is inert. | DONE |
| F68 / V62 | Separate candidate and approved field contracts; admit omission/profile-default only in candidates; preserve approved bytes and fingerprints. | DONE |
| V59 | Requalify the exact historical 28 product rows, new regressions, generator regression and required checks at compatible final source. | DONE |

| Slice | Prerequisites | Status | Binary exit / main risk |
| --- | --- | --- | --- |
| S73 Specification and characterization | Revalidated baseline | DONE | Candidate/approved policy matrix is explicit and both defects reproduced with valid controls; avoid unrelated invalid fixtures. |
| S74 Candidate admission | S73 | DONE | V61 passes with one presence-aware conversion and unchanged approved identities; avoid premature defaulting. |
| S75 Contract projection | S73 | DONE | V62 passes against generated types and validators; approved and closed-variant requirements remain strict. |
| S67 Final qualification and handoff | S74, S75 | DONE | V59, cleanup, comparison and Markdown evidence complete; no stale pass substitutes for changed-source qualification. |

NF-REQ-097 and Tables 10-B0/10-D1/10-F clarify the already-adopted boundary.
Defaults and profile permissions remain separate. Candidate JSON null/empty
values gain no compatibility fallback. Valid stored mappings remain unchanged;
malformed approved mappings are rejected. No SQL migration, state/major bump,
registry translator or automatic data reset is introduced. Expected generated
artifact hashes may change; approved serialization, fingerprints and frozen
NF-FIX-029 revision 5 must not.

Discovery passed native 0: `make help ALL=1` and `make task-guide` for owners
`module.networkflow`, `web.networkflow`, `package.protocol_ts` and
`harness.generated_artifacts`. Execution uses the pinned Go/PATH, installed-only,
cache-off prefix recorded in the historical section below, fresh `nf-presence-*`
run IDs, native sessions and public Make targets. Tests/generators consume only
machine inputs and fixtures, never this tracker or other Markdown.


### S73–S75 execution evidence

Registered selectors (all three `make author-test-row-id` invocations passed native 0):

| Obligation | Actual registered row | Exact selector |
| --- | --- | --- |
| V61 unit | `module.networkflow.unit.mapping_policy_admission_preserves_json_presence_d4e3e0dd31` | `go:./internal/modules/networkflow:TestMappingPolicyJSONPresence` |
| V61 public routes | `module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4` | `go:./internal/modules/networkflow:TestCiscoPolicyJSONPresenceAdmission` |
| V62 generated contracts | `package.protocol_ts.frontend_unit.mapping_policy_contracts_distinguish_candidate_d_5f524025f1` | Vitest `node-unit`, `packages/protocol-ts/src/network-flow-mapping.node.test.ts`, title `Network Flow mapping policy contracts distinguishes candidate defaults from concrete approved policies` |

Each new row was authored in its owner family and generated before execution.
`nf-presence-s73-format-20261031-01 make format` and
`nf-presence-s73-routing-20261031-01 make generate` passed native 0.
The first unit reproduction, `nf-presence-s73-unit-red-20261031-01`, failed
before policy admission because its header-only CSV raised
`network_flow_no_data_rows`; it is a test setup failure, not defect evidence.
The fixture was corrected to contain one valid nonempty record before rerunning.

| Reproduction run | Exact command | Native exit and observation |
| --- | --- | --- |
| `nf-presence-s73-unit-red-20261031-02` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.mapping_policy_admission_preserves_json_presence_d4e3e0dd31` | 2; explicit null/empty values accepted across both profiles; non-string policies lacked the required field-specific reason. |
| `nf-presence-s73-contract-red-20261031-02` | `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.mapping_policy_contracts_distinguish_candidate_d_5f524025f1` | 2; explicit candidate/approved controls passed, then omitted candidate policy was rejected. |
| `nf-presence-s73-route-red-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4` | 2; all ten null/empty requests returned 200 instead of 400 after successful valid-source previews. |

An earlier contract reproduction (`nf-presence-s73-contract-red-20261031-01`,
same command/native 2) had the same omission failure before explicit positive
controls were moved ahead of it. Reproduction source digests respectively:
unit `sha256:16a4c75857b679d564b6097fcd2df6384dd48b351428da4a9cbf6abb2feb8ff0`,
contract `sha256:5c2dfdbb1bbb55dc0c9107dff8dfccf725dfb1e01b168798623cb235e2268b58`,
route `sha256:f7370e6e85bc1dcb780fec98556dcc6d8f5797f34fa97c50bb309682a334093c`.
Exact roots are `.cartulary/test-results/<run-id>`; each retains
`run-manifest.json`, `run-summary.json`, row results, bounded logs,
`cleanup-results.json` and `retained-secret-scan.json`. Reproductions completed
cleanup and passed retained scanning. Initial `explain-run DETAIL=logs` attempts
without a target and with runner-family targets were rejected as usage errors;
`TARGET=test-slice` correctly exposed the owning diagnostics. No test was
rerun for those observer errors.

S74 uses a private `candidateFieldMapping` with `json.RawMessage` for the policy.
Only omission and exact `profile_default` resolve defaults during conversion to
approved fields. Explicit null/empty/non-string values fail with the field-specific
policy error. The existing registry permission check handles concrete tokens;
ignored/system-derived variants cannot carry a candidate policy. Approved
representation, decoding, serialization and fingerprint code retain their
concrete-policy boundary. Test-only typed candidate builders were adapted.

S75 introduces `CandidateFieldMapping` in the authored schema and shares the
unchanged ignored/system-derived definitions with `FieldMapping`. The approved
schema expands identically to the pre-session version. Candidate policy alone
becomes optional and admits `profile_default`; transform/combinability requirements
remain unchanged. `nf-presence-s74-format-20261031-01 make format` and
`nf-presence-s75-generate-20261031-01 make generate` passed native 0.
No generator exception or generated-file hand edit was needed.

| Focused run | Exact command | Native exit |
| --- | --- | --- |
| `nf-presence-focused-unit-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.mapping_policy_admission_preserves_json_presence_d4e3e0dd31,module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0` | 0 |
| `nf-presence-focused-contract-20261031-01` | `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.mapping_policy_contracts_distinguish_candidate_d_5f524025f1` | 0 |
| `nf-presence-focused-types-20261031-01` | `make lint CHECK=types` | 0 |
| `nf-presence-focused-route-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4,module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556,module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488` | 0 |

These focused runs share source
`sha256:68c247ac3717d46527ef3c68fb68403353c3eabeadb53ca5514ff69a13d45200`.
The route controls were subsequently strengthened to provide schema-required
transform and combinability members explicitly, followed by a focused rerun
of the two admission rows. Final qualification will cover this strengthened
source and the complete independent example comparison again.

`nf-presence-focused-route-20261031-02` passed native 0 with
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4,module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556`.
All focused runs completed cleanup and retained scanning. S73–S75 are complete;
S67/V59 remain open for finalization and fresh qualification. No compatible full
warm-check root exists for optional retained-run maintenance; skip that operation
without launching a redundant full check.

S75 documentation checkpoint `nf-presence-s75-docs-20261031-01 make
lint-markdown` passed native 0; summary is
`adhoc/lint-markdown/tool-run-summary.json` beneath its exact run root.
The first finalization `nf-presence-finalize-20261031-01 make agent-finalize`
passed native 0. Subsequent source review limited the non-source policy-member
check to the two supported non-source variants, preserving
`mapping_kind_unavailable` for `constant`/`derived`; the unit row now asserts
that precedence. A focused unit rerun and a second finalization precede the
final source freeze. The first finalization is therefore historical preparation,
not the final qualification baseline.

### S67 frozen-source qualification

The final focused unit rerun `nf-presence-focused-unit-20261031-02` passed
native 0 with the same two unit selectors above. Second finalization
`nf-presence-finalize-20261031-02 make agent-finalize` passed native 0.
Source review found only the intended candidate admission, authored schema,
test/routing and generated projection changes. Approved mapping representation
and fingerprint implementation remain unchanged; frozen fixture files match
their pre-session hashes. The final product source is frozen; only handoff
documentation changes follow.

Exact final roots are `.cartulary/test-results/nf-presence-final-<group>-20261031-01`.
All commands use the pinned prefix recorded below. The portfolio preserves the
previous exact 28 product rows and adds three registered regressions: 15 unit,
nine integration, three browser, three client and one protocol row. Generator
regression is an additional row. Source/lint obligations remain unchanged.

| Group | Exact Make command | Native exit |
| --- | --- | --- |
| unit | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_cisco_sna_required_66f0e756eb,module.networkflow.unit.network_flow_selector_covers_cisco_sna_v1_permit_a2c32abbbf,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.network_flow_selector_covers_every_network_flow_13fc533c41,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.unit.mapping_policy_admission_preserves_json_presence_d4e3e0dd31` | 0 |
| integration | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556,module.networkflow.integration.mapping_policy_null_and_empty_requests_reject_wi_c863b856b4` | 0 |
| browser | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179,module.networkflow.browser.import_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619` | 0 |
| client | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 0 |
| protocol | `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.mapping_policy_contracts_distinguish_candidate_d_5f524025f1` | 0 |
| generator | `make test-slice OWNER=harness.generated_artifacts ROWS=harness.generated_artifacts.unit.decode_contract_rejects_duplicate_jsonmembers_005d1b3b94` | 0 |
| generated | `make source-check CHECK=generated` | 0 |
| schema | `make source-check CHECK=schema` | 0 |
| catalog | `make source-check CHECK=catalog` | 0 |
| go | `make lint CHECK=go` | 0 |
| types | `make lint CHECK=types` | 0 |
| boundary | `make lint CHECK=backend-boundary` | 0 |

All twelve native sessions exited 0. Every selected row passed; each run's
cleanup actions completed without failures and its retained-secret scan passed.
Final identity is identical across all twelve manifests:

- Commit: `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`, dirty worktree preserved.
- Source: `sha256:29bb731b540e23d1204876e90d2c61452b7da0edf0266e040b16cec11b421447`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- Environment: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

The exact roots retain `run-manifest.json`, `run-summary.json`,
`target-summaries/`, selected `rows/`, `cleanup-results.json` and
`retained-secret-scan.json`. Final NF-FIX-029 comparisons passed at:

- `.cartulary/test-results/nf-presence-final-integration-20261031-01/NF-FIX-029-cisco-sna-flow-search-4253886926/comparison-summary.json`.
- `.cartulary/test-results/nf-presence-final-browser-20261031-01/NF-FIX-029-cisco-sna-flow-search-node-1tURWR/comparison-summary.json`.

Both retain freeze revision 5, manifest SHA-256
`01998415c8e08ad6d41e4411743a90af8a3bab509df66d41574696cda4c72e01`,
source bundle SHA-256
`896cc8b54d8c0c5675e74848c423f14852f5f207b6bc3674e156e639429a7b83`,
and expected bundle SHA-256
`c8baf09449d135a288a382e36749cb2856fe4a3397fab62365f785f3c3272286`.
Original CSV and every frozen fixture file match their pre-session content.
Approved bytes/fingerprints remain stable across omission, profile-default and
explicit-default candidates; the original-profile and full example regressions
retain independent expectations.

V61 confirms both profiles reject explicit null/empty/malformed candidate
policies, approved decoding rejects unresolved policies, and both public routes
preserve durable owner/session/unit/receipt/job state on repeated rejection.
Corrected approval reuses the rejected transaction ID and replays unchanged.
V62 confirms generated candidate optionality, approved requiredness, closed
variants and unchanged requirements for other candidate members. S72's original
permission matrix still passes. No final product or source/lint check failed.

Source freeze comparison at
`/tmp/cartulary-policy-baseline-d_qd6yvt/freeze.json` found no product-file changes
after qualification. The pre-session content comparison found only the intended
authored, test/routing, documentation and generated projection changes; the index
remains unstaged and unrelated prior changes remain untouched.

Compatibility: major 8/state 5 and approved wire identities remain unchanged.
Malformed candidate requests now fail; callers must omit the policy, use
`profile_default`, or provide a permitted concrete token. Valid stored approvals
remain valid, including already-materialized results of earlier requests whose
original presence cannot be reconstructed. Malformed approved objects are
rejected, never repaired. Generated registry identities changed normally;
matching server/browser/assets must be published together. Any in-flight
workflow rejected by existing registry admission must be recreated from source;
no translator, migration, automatic reset or legacy reader was added.
Rollback uses a previous matching application/assets/environment combination.

Full CI, whole-owner evidence audits, release/performance qualification and
visual golden updates were not selected: this task's complete affected portfolio
and explicit checks define the acceptance claim. Optional retained maintenance
was skipped because no compatible full warm-check root exists. Commit,
deployment and data reset remain unperformed.

Final documentation closure `nf-presence-final-docs-20261031-01 make
lint-markdown` passed native 0; its retained summary is
`.cartulary/test-results/nf-presence-final-docs-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
F67/V61, F68/V62, S73–S75 and S67/V59 are DONE. No related residual gap remains.
The final completion-status checkpoint uses run ID
`nf-presence-final-docs-20261031-02` and `make lint-markdown`, with its result
published at the same summary suffix under that exact root. Product source
remains frozen; the completion-status edit changes only this handoff.

## Historical S72 — Flow Search policy permission correction

The maintainer authorized F66 correction without committing on 2026-10-31.
Baseline is `main` at `d936506e2b851a22e8c9fbddd983ab4fd6aa21de` with the
pre-existing Cisco implementation and review changes preserved. This entry
reopened and now closes S67/V59; prior qualifications below remain historical evidence.

| Finding / obligation | Requirement and exit | Status |
| --- | --- | --- |
| F66 / P2 | NF-REQ-080c/080d/097: reject explicit null-empty policies on Flow Search's seven non-port targets at mapping admission. | CLOSED |
| V60 | Candidate, approved/decode and real Imports preview/approval negative cases; no rejection side effects; corrected transaction-ID reuse; valid mapping/fingerprint stability. | PASS |
| V59 | Requalify the existing 27-row product portfolio plus the new admission row, affected generator row and required checks on compatible final source. | PASS |

| Workstream | Dependency | Status | Exit / principal risk |
| --- | --- | --- | --- |
| S72 Policy admission correction | Revalidated dirty baseline | DONE | Explicit registry permissions, shared admission enforcement and V60 pass without changing valid identities or frozen revision-5 bytes; closing Markdown passed before S67. |
| S67 Final validation and handoff | S72 | DONE | Finalized source qualifies all 28 product rows, generator regression and required checks; comparisons, cleanup, Markdown and handoff are complete. |

Registry permissions are independent of defaults and target nullability. Original
source-mappable fields retain both currently valid policy tokens; Flow Search
ports permit invalid/null, its other seven targets permit invalid only, and
unsupported/system-derived fields permit no source policy. Generated Go/TS
metadata stays outside approved mappings and fingerprint bytes. No API, SQL,
state-version, migration, conversion, compatibility layer or row-normalization
change was introduced by S72. NF-FIX-029 revision 5 and the original CSV remain unchanged.

Execution uses the pinned Go/PATH, installed-only and cache-off prefix recorded
below, fresh run IDs, public Make targets, and the test-operations skill. Record
commands, actual selectors, source/toolchain/environment identities, native exits,
summary/comparison paths and cleanup here. Update S72 and pass Markdown before
beginning S67. Register the new selector through authored routing before running
it; proposed names are not available or passing IDs. Commit/deployment excluded.

Selection discovery passed native 0: `nf-policy-guide-backend-20261031-01 make
task-guide OWNER=module.networkflow`, `nf-policy-guide-web-20261031-01 make
task-guide OWNER=web.networkflow`, and `nf-policy-guide-generator-20261031-01
make task-guide OWNER=harness.generated_artifacts`.

S72 reproduction: `nf-policy-s72-unit-red-20261031-01 make test-slice
OWNER=module.networkflow
ROWS=module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0`
failed native 2 as intended: all seven forbidden policies were accepted by
candidate materialization, approved validation and decoding. The exact root
under `.cartulary/test-results/` retains `run-summary.json`, the selected row
and bounded Go logs. This is the failing prerequisite, not qualification.

`nf-policy-s72-row-20261031-01 make author-test-row-id
FAMILY_ID=module.networkflow.integration
CLAIM='Flow Search policy admission rejects forbidden null policies without durable effects'
SELECTOR_KEY='go:./internal/modules/networkflow:TestCiscoSNAFlowSearchPolicyAdmission'`
passed native 0 and authored the actual row
`module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556`.
The authored family now registers that selector. Generation precedes its first
execution. Each subcase uses valid nonempty CSV and confirms successful preview
before changing only the forbidden policy.

Routing generation `nf-policy-s72-routing-20261031-01 make generate` passed
native 0. The new route row then failed native 2 in
`nf-policy-s72-route-red-20261031-01 make test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556`:
all seven valid-source previews returned 200 where 400 was required. Both red
runs completed cleanup without failures. F66 is reproduced at the real serving
boundary; S72 now applies the approved registry-driven correction.

S72 exit: NF-REQ-080c/080d/097 now separate defaults, permissions and target
nullability. Every authored registry field declares its allowed policy list;
the closed shape validator rejects unknown/duplicate tokens, empty permissions
on source targets, defaults outside their list, and permissions on unsupported
or derived targets. Make projects the metadata to Go and TypeScript. Shared
approved-mapping validation checks membership after field/transform admission,
preserving Imports error translation and approved serialization. No parser or
row-normalization change was needed.

The existing profile row tests both policy tokens on every source-mappable field
of both profiles at candidate, approved-validation and decoding boundaries.
Omitted/profile-default/explicit-default equivalence, blank/null/zero/malformed
ports and unsupported interfaces remain covered. The new integration row uses
seven isolated valid-source cases, exact Core/owner errors, repeated preview and
approval rejection, complete session/unit/receipt/job snapshots and owner durable
state. Corrected approval reuses the rejected transaction ID; replay preserves
its response and state. The stored owner mapping equals the valid preview.

| Run ID | Exact Make command | Native exit / evidence |
| --- | --- | --- |
| `nf-policy-s72-generate-20261031-01` | `make generate` | Tool summary pass/0; native exit was not retained. Finalization and generated-source qualification below independently cover this output. |
| `nf-policy-s72-unit-green-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0` | 0; profile unit row passes. |
| `nf-policy-s72-integration-green-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556,module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488` | 0; both rows pass. |
| `nf-policy-s72-format-20261031-01` | `make format` | 0; authored Go formatting only. |

All commands use the following prefix, substituting the exact run ID above:

```sh
PATH=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin:$PATH \
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go \
CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_HARNESS_CACHE_MODE=off \
CARTULARY_TEST_RUN_ID=<run-id>
```

Focused green test source:
`sha256:a5c92944bb7e1514065d3bbe96c665f2fc6ae4f8d00f0442ffe7c8d47ee81f71`.
Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`;
environment: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Exact roots are `.cartulary/test-results/<run-id>`; graph runs retain
`run-manifest.json`, `run-summary.json`, selected `rows/`,
`cleanup-results.json` and `retained-secret-scan.json`. Both green runs passed
retained scanning and completed every cleanup action without failures.
Generation retains `generate/tool-run-summary.json`.

Complete example comparison:
`nf-policy-s72-integration-green-20261031-01/NF-FIX-029-cisco-sna-flow-search-147752387/comparison-summary.json`.
Revision 5, manifest SHA-256
`01998415c8e08ad6d41e4411743a90af8a3bab509df66d41574696cda4c72e01`,
source/expected bundle hashes and original CSV SHA-256 remain the previously
recorded values. No frozen bytes or independent fingerprint expectations changed.
No new runtime defect emerged. F66 awaits final source qualification, with no
remaining implementation work or migration. S72 closing checkpoint:
`nf-policy-s72-close-20261031-01 make lint-markdown`.

S72 closing Markdown passed native 0; summary:
`.cartulary/test-results/nf-policy-s72-close-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
S67 begins with `nf-policy-finalize-20261031-01 make agent-finalize`.
The final portfolio preserves the previous 27 exact selections and adds the
registered admission row, plus
`harness.generated_artifacts.unit.decode_contract_rejects_duplicate_jsonmembers_005d1b3b94`.
Source/check/lint obligations are unchanged. No compatible full warm-check root
exists for retained-run maintenance; that optional operation is skipped without
creating a redundant qualification run. All product/check evidence will be fresh
at the finalized source; development and prior final passes remain historical.

Finalization passed native 0 with completed cleanup and retained scanning. Diff
inspection found only the expected generated extension-integrity hash refresh
for formatted `tools/contractgen/main.go`; no further runtime, fixture or authored
input changed. Source is frozen for final qualification. Exact final commands,
identities, comparisons and cleanup are reconciled below after native exits.

S67 final product/check evidence: all 28 product rows (14 unit, eight
integration, three browser and three client), the affected generator row and
six source/lint checks pass natively at the same frozen source. All eleven
runs completed every cleanup action without failures and passed retained-secret
scanning. No final test or check failed; the two intentional red reproductions
above remain recorded failures and are not qualification.

Common final identities from every exact run's `run-manifest.json`:

- Commit: `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`, dirty worktree preserved.
- Source: `sha256:8cb3c06dbfebc2381d272a8937397248c94f5ccc471336e740aac42923b648b0`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- Environment: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

The exact result root for each group is
`.cartulary/test-results/nf-policy-final-<group>-20261031-01`.
Each retains the manifest, `run-summary.json`, `target-summaries/`, selected
`rows/` where applicable, `cleanup-results.json` and
`retained-secret-scan.json`. Commands use the pinned prefix above.

| Group | Exact Make command | Native exit |
| --- | --- | --- |
| unit | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_cisco_sna_required_66f0e756eb,module.networkflow.unit.network_flow_selector_covers_cisco_sna_v1_permit_a2c32abbbf,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.network_flow_selector_covers_every_network_flow_13fc533c41,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b` | 0 |
| integration | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.flow_search_policy_admission_rejects_forbidden_n_acaa564556` | 0 |
| browser | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179,module.networkflow.browser.import_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619` | 0 |
| client | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 0 |
| generator | `make test-slice OWNER=harness.generated_artifacts ROWS=harness.generated_artifacts.unit.decode_contract_rejects_duplicate_jsonmembers_005d1b3b94` | 0 |
| generated | `make source-check CHECK=generated` | 0 |
| schema | `make source-check CHECK=schema` | 0 |
| catalog | `make source-check CHECK=catalog` | 0 |
| go | `make lint CHECK=go` | 0 |
| types | `make lint CHECK=types` | 0 |
| boundary | `make lint CHECK=backend-boundary` | 0 |

The final NF-FIX-029 comparisons both pass:

- `nf-policy-final-integration-20261031-01/NF-FIX-029-cisco-sna-flow-search-1133829094/comparison-summary.json`.
- `nf-policy-final-browser-20261031-01/NF-FIX-029-cisco-sna-flow-search-node-qiE6L6/comparison-summary.json`.

Integration retains the full `produced/actual.json` and thirteen targeted
wrong-expectation controls. Browser evidence includes the existing complete
saved-graph reopen/refresh assertions through a fresh page; import recovery and
alias suggestion also pass. Each comparison summary records its materialized
input and produced artifact paths. These pass with unchanged frozen revision 5:

- Manifest SHA-256: `01998415c8e08ad6d41e4411743a90af8a3bab509df66d41574696cda4c72e01`.
- Source bundle SHA-256: `896cc8b54d8c0c5675e74848c423f14852f5f207b6bc3674e156e639429a7b83`.
- Expected bundle SHA-256: `c8baf09449d135a288a382e36749cb2856fe4a3397fab62365f785f3c3272286`.
- Original CSV SHA-256: `b03c98cf0124f426e05d1b87f6ca1cf38c5f8a145e22c7dec6d3f527b99e70c3`
  (1,275,537 bytes, exact match to the supplied file).

V60 closes the policy-admission and compatibility obligations: all seven
forbidden policies fail at candidate/approved/decode and public preview/approval
boundaries with the expected field-specific error; repeated rejection is inert;
corrected transaction-ID reuse and successful approval replay work. Valid
mappings and independent fixture fingerprints remain unchanged, including the
original-profile vectors. Permissions do not enter serialized approved mappings.

Compatibility/handoff: no public API, durable mapping shape, SQL, migration,
state-version, row normalization, identity algorithm or translator changed in
S72. Network Flow remains major 8/state 5. Invalid unreleased Flow Search mappings
are rejected at admission/decoding, rather than converted; valid stored mappings
retain their meaning and fingerprints. Rollback is a source/test reversion;
databases with earlier Flow Search imports still require software supporting
that profile. No reset or data conversion is authorized or needed.

All required product/check obligations are satisfied. Final Markdown closure
`nf-policy-final-docs-20261031-01 make lint-markdown` passed native 0, retained at
`adhoc/lint-markdown/tool-run-summary.json` beneath that exact run root.
F66/V60 and S67/V59 are closed. The completion-status documentation checkpoint
`nf-policy-final-docs-20261031-02 make lint-markdown` passed native 0, with the same summary
suffix beneath its own exact root. No product source changed after the freeze.
No related residual gap is
known; historical unrelated debt remains historical. Full CI, complete-owner
qualification, all-29 behavioral reruns, visual certification and performance
publication remain outside this affected-portfolio claim. No commit or
deployment was performed.

## Cisco SNA export example — review corrections complete

The maintainer authorized implementation of the F63–F65 remediation plan on
2026-10-31. Baseline: `main` at `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`,
with the pre-existing Cisco example changes listed by `git status --short`.
Those changes and all historical evidence are preserved. S64's default-policy
obligation, S66's complete comparisons/reopen proof, and S67 closure are reopened.
The earlier S36 remediation remains complete; prior S67 passes establish only
their recorded coverage, not the missing obligations below.

| Finding | Required correction | Validation | Workstreams |
| --- | --- | --- | --- |
| F63 / P2 — CLOSED | NF-REQ-178: complete applicable fixture outcomes, independent expectations, v2 adapter input and targeted wrong-expectation controls. | V56 PASS | S70 |
| F64 / P2 — CLOSED | NF-REQ-080d/097: explicit per-profile defaults, equivalent omission/profile-default materialization and unchanged old identities. | V57 PASS | S68, S69 |
| F65 / P2 — CLOSED | Fixture 029: restore the persisted saved graph in a fresh page before explicit refresh. | V58 PASS | S71 |

| Workstream | Dependency | Status | Binary exit / principal risk |
| --- | --- | --- | --- |
| S68 Baseline and specification closure | Revalidated dirty baseline | DONE | NF-REQ-097 now defines the per-profile matrix; NF-REQ-080d links it. V56 applicability is recorded above; Markdown passed. |
| S69 Default-policy verification | S68 | DONE | V57 passes equivalent approved mappings/fingerprints, blank/zero/invalid ports, required empty values, explicit overrides, unsupported interfaces and both client profiles. No runtime change. |
| S70 Complete fixture comparison | S69 | DONE | V56 passes complete outcomes, authenticated continuation, adapter equality, thirteen negative controls and frozen revision-5 admission. |
| S71 Fresh-page saved-graph proof | S70 | DONE | V58 passes fresh-page complete persisted declaration/result equality, no reopen mutations, distinct refresh job, generation advancement and guaranteed page cleanup. |
| S67 Final validation and handoff, reopened | S68–S71 and required corrections | DONE | V59 passes all 27 affected rows, required source/lint checks and Markdown; compatible evidence, cleanup and handoff are complete. |

V56 comparison matrix: source inspection and preview (all descriptors/samples),
approval and selection, terminal import result, complete table resources and
row/diagnostic pages, authenticated cursor payload and continuation, durable
rows/publication/import/audit state, full graph response, v2 adapter input and
trusted invocation facts. All 1,200 persisted rows retain independently computed
raw-provenance hashes; public rows retain complete raw provenance. Indicator
link results, error payloads and resource-limit failures are not applicable to
this successful import/graph scenario; existing dedicated bundles own them.
Actual graph limit metadata remains compared. Safe digest fields and absence
of raw source material in graph metadata remain exact assertions.

The final selection is the existing 25-row S67 portfolio below plus
`module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2`
and `module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d`.
Use fresh explicit run IDs, pinned tools and cache-off as below. Record command,
selection, source/environment identities, native exit, artifacts and cleanup.
After each workstream update this tracker and run Markdown before its successor.
Run finalization before the final source freeze. No public API, SQL migration,
state-version change, production hook or compatibility translator is planned.
Only fixture 029 advances: the initial revision-3 authoring and subsequent
corrections below produce qualified revision 5. Original CSV bytes and earlier
bundles remain unchanged. Any further frozen-byte change advances again.

S68 exit: `nf-sna-review-s68-docs-20261031-01 make lint-markdown` passed
native 0; summary is its run root's `adhoc/lint-markdown/tool-run-summary.json`.
`nf-sna-review-guide-backend-20261031-01 make task-guide OWNER=module.networkflow`
and `nf-sna-review-guide-web-20261031-01 make task-guide OWNER=web.networkflow`
both passed native 0. All roots are beneath `.cartulary/test-results/`.
S69 is ready; no owner contradiction or migration is required. Closing tracker
checkpoint: `nf-sna-review-s68-close-20261031-01 make lint-markdown`.

S68 closing Markdown passed native 0. S69 extended the existing registered
profile and client-model rows; no new selector or generated input was needed.
Focused native-0 evidence: `nf-sna-review-s69-unit-20261031-01 make test-slice
OWNER=module.networkflow` selected
`module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0`,
`module.networkflow.unit.network_flow_selector_covers_cisco_sna_required_66f0e756eb`,
and `module.networkflow.unit.network_flow_selector_covers_cisco_sna_v1_permit_a2c32abbbf`;
`nf-sna-review-s69-client-20261031-01 make test-slice OWNER=web.networkflow
ROWS=web.networkflow.regression.networkflowimportmodel_suite_12249943c5` also passed.
The runs retain manifests, target summaries and cleanup beneath their exact roots.
F64/V57 are resolved pending final compatible qualification. S70 is ready after
`nf-sna-review-s69-close-20261031-01 make lint-markdown`.

S69 closing Markdown passed native 0. S70 now compares the publication outcome
without dropping contract fields, authenticates and compares two page cursors,
exercises continuation, and compares captured production composition with the
actual graph route. Complete adapter input/invocation facts are retained without
duplicating the already-compared graph result. Thirteen targeted negative
controls cover the former omissions, closed members, ordering and null/zero.

Expectation authoring uses original CSV bytes, previously independent row
vectors and the static contract shapes of fixtures 001/014/015; it never reads
recorded product output. Independent SHA-256/HMAC/length-framing calculations
produce new identities and graph bindings. During review of initially authored
revision 3, the oracle's generic key sorting was corrected to preserve Network
Flow's closed aggregation/sort-member order; the table GET schema was corrected
to its adopted spelling. The bundle advances again to revision 4 before its
first execution. Revision 3 was not qualification. Original CSV remains exact.

S70 focused run `nf-sna-review-s70-integration-20261031-01` selected the
example, graph-adapter and paging rows with `make test-slice OWNER=module.networkflow`.
Native exit 2: graph/paging passed; the example comparison failed on 27 absent-port
selectors. Diagnosis: NF-REQ-191's presence/value variant, projected by the closed
`GraphDefaultEdgeSelectorV2` schema, omits `destination_port` when presence is
false; the independently authored expectation incorrectly supplied null. The
oracle now emits the absent variant; null graph properties/public row ports remain
compared exactly. No runtime defect or requirement amendment. Frozen revision 5
replaces 4; original bytes remain exact. Cleanup completed without failures.
Failure evidence: the exact root's `run-summary.json`, fixture 029
`comparison-summary.json` and `produced/actual.json`. Retry is limited to the
same three rows after this diagnosed expectation correction.

S70 integration run 02 passed native 0 for all three rows, including all thirteen
negative controls. A subsequent example-only operation-facts check
`nf-sna-review-s70-opaque-20261031-01` failed native 2: the new assertion treated
retained `json.RawMessage` envelopes as decoded maps and therefore collected no
request IDs. It now decodes that existing representation before checking audit
correlation. No expectation or product behavior changed. Cleanup completed.
`make explain-run` diagnosis attempts 01/02 rejected incomplete/invalid target
selection (native 2); attempt 03 with `TARGET=test-slice DETAIL=logs` passed
native 0 and identified the assertion. The example-only retry closes this
additional validation; prior graph/cursor passes remain development evidence.

S70 exit: `nf-sna-review-s70-opaque-20261031-02 make test-slice
OWNER=module.networkflow
ROWS=module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488`
passed native 0. Job UUID, attribution/scope/route, ordered operation times,
retention, distinct request/audit IDs and audit-request correlation are validated
before narrow substitutions. Existing apply replay compares full responses and
durable effects. Both the complete outcome and targeted negative controls pass;
all cleanup steps completed without failures. V56/F63 are resolved pending final
source qualification. Manifest SHA-256:
`01998415c8e08ad6d41e4411743a90af8a3bab509df66d41574696cda4c72e01`;
expected bundle SHA-256:
`c8baf09449d135a288a382e36749cb2856fe4a3397fab62365f785f3c3272286`.
S71 is ready after `nf-sna-review-s70-close-20261031-01 make lint-markdown`.
All run source/toolchain/system identities are retained in each exact root's
`run-manifest.json`; summaries and produced comparison artifacts are linked
by its fixture `comparison-summary.json`.

S70 closing Markdown passed native 0. S71 begins with a fresh authenticated
page in the same incident; complete declaration/result equality precedes a
distinct refresh job and generation. The page is closed through `finally`.

S71 exit: `nf-sna-review-s71-browser-20261031-01 make test-slice
OWNER=module.networkflow
ROWS=module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179`
passed native 0. The fresh page selects the stable graph ID, observes actual GET
declaration/result responses, compares the entire persisted response, and proves
no import/upload/saved-graph mutation occurred during reopening. Explicit refresh
then succeeds with a distinct job and generation +1, unchanged semantic query,
complete selected-result binding and established totals. The Playwright
`cisco-saved-graph-reopen` JSON attachment retains the proof; fixture revision 5
comparison passed. All service cleanup completed and the extra page closed.
The walkthrough now describes this verified sequence. F65/V58 are resolved
pending final compatible qualification. No runtime defect was exposed.
S67 is ready after `nf-sna-review-s71-close-20261031-01 make lint-markdown`.

S71 closing Markdown passed native 0. S67 now formats authored changes, runs
`make agent-finalize`, inspects any generated changes, and freezes source for
the 27-row affected portfolio defined above. Development passes do not substitute
for compatible final evidence. Optional retained-run maintenance is skipped
because no qualifying full warm-check root exists; no full check is requested.

S67 preparation: `nf-sna-review-s67-format-20261031-01 make format` and
`nf-sna-review-s67-finalize-20261031-01 make agent-finalize` passed native 0.
Reviewed generated changes remain downstream Cisco-profile projections and
verification routing; this correction adds no authored projection or runtime
change. Frozen revision-5 manifest hash is unchanged after formatting/finalize.
Source is now frozen for final product qualification; only tracker evidence
updates follow. Final invocations use `nf-sna-review-final-<group>-20261031-01`
for groups unit, integration, browser, client, generated, schema, catalog, go,
types and boundary. Unit/browser/client use the exact historical selection table;
integration adds the two graph/cursor rows named in this iteration. Source checks
use `make source-check CHECK=<group>`; lint uses `make lint CHECK=go`, `types`,
or `backend-boundary`. All use the pinned cache-off prefix above.

### Final V59 qualification and handoff

All ten final invocations passed native 0. The 27 distinct selected product rows
(14 unit, 7 integration, 3 browser, 3 client) passed. All reported cleanup steps
completed without failure, and retained-secret scanning passed. No missing
selected row, comparison, check or cleanup remains.
`nf-sna-review-final-docs-20261031-01 make lint-markdown` passed native 0;
summary: `.cartulary/test-results/nf-sna-review-final-docs-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
The completion-status checkpoint
`nf-sna-review-final-docs-20261031-02 make lint-markdown` passed native 0, with the same summary
suffix in its exact run root. It changes no product source or qualification.

Common final identities, from each run's `run-manifest.json`:

- Commit: `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`, dirty worktree preserved.
- Source: `sha256:69da28f82623ad7f287bcaa4766a869603e758ba8e17efed7d263ad0b1081085`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- Environment/system: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

Exact roots are `.cartulary/test-results/nf-sna-review-final-<group>-20261031-01`
for each group below. Each contains `run-manifest.json`, `run-summary.json`,
`target-summaries/`, `rows/` where applicable, `cleanup-results.json`, and
`retained-secret-scan.json`. Commands use the exact pinned prefix above.

| Group | Exact Make command / selection | Native exit |
| --- | --- | --- |
| unit | Historical portfolio table's exact unit command (14 rows) | 0 |
| integration | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.effective_resource_limit_discovery,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d` | 0 |
| browser | Historical portfolio table's exact browser command (3 rows) | 0 |
| client | Historical portfolio table's exact client command (3 rows) | 0 |
| generated | `make source-check CHECK=generated` | 0 |
| schema | `make source-check CHECK=schema` | 0 |
| catalog | `make source-check CHECK=catalog` | 0 |
| go | `make lint CHECK=go` | 0 |
| types | `make lint CHECK=types` | 0 |
| boundary | `make lint CHECK=backend-boundary` | 0 |

Final fixture 029 comparison summaries:

- Integration: `nf-sna-review-final-integration-20261031-01/NF-FIX-029-cisco-sna-flow-search-2917067486/comparison-summary.json`.
- Browser: `nf-sna-review-final-browser-20261031-01/NF-FIX-029-cisco-sna-flow-search-node-tkJEpg/comparison-summary.json`.
  Its summary records the materialized input and evidence paths. The browser
  report is `browser-e2e-webserver-backed/browser-groups/cohort-6f32232ace264f5c1e9fb1ec7eb55eebb213ab3d3f792ab481714b58ea4299d2/playwright-report.json`
  within that exact root. It retains the complete
  `cisco-saved-graph-reopen` JSON attachment, including the restored response,
  declaration, absence of reopen mutations, refreshed declaration and job.

Both summaries pass at frozen revision 5 with manifest SHA-256
`01998415c8e08ad6d41e4411743a90af8a3bab509df66d41574696cda4c72e01`.
The integration's `produced/actual.json` and `produced/negative-controls.json`
close the V56 matrix above, including all thirteen rejected incorrect
expectations. Source bundle SHA-256:
`896cc8b54d8c0c5675e74848c423f14852f5f207b6bc3674e156e639429a7b83`;
expected bundle SHA-256:
`c8baf09449d135a288a382e36749cb2856fe4a3397fab62365f785f3c3272286`.
The original CSV remains 1,275,537 bytes, SHA-256
`b03c98cf0124f426e05d1b87f6ca1cf38c5f8a145e22c7dec6d3f527b99e70c3`,
identical to the supplied file. Earlier fixture bundles are unchanged.

S64's reopened default obligation, S66's comparisons/reopen obligations and S67
closure are complete. No production correction was required by the stronger
assertions. Public routes, schemas, SQL columns, identity algorithms, existing
approved mappings and Network Flow major 8/state 5 are unchanged by this review
correction. No migration, conversion, reset or runtime compatibility layer was
introduced. Rollback is a source/test reversion; preserve software support for
the earlier new-profile adoption in databases containing those imports.

The updated [walkthrough](../guides/network-flow-cisco-sna-example.md) provides
the verified upload/save/reopen/refresh sequence. No related residual gap is
known. Historical unrelated debt and earlier qualifications remain historical.
Full CI, complete-owner qualification, all-29 behavioral reruns, visual
certification, performance publication, commit and deployment are outside this
acceptance claim. No commit or deployment was performed.

### Earlier Cisco example implementation and evidence

The maintainer requested adoption of the synthetic 1,200-row Cisco export and
explicitly selected support for its original format plus repository examples and
automated verification. The completed remediation below is historical evidence
for commit `d936506e2`; this follow-up begins from its clean `main` checkout.
No application population, deployment or new commit is requested.

The supplied `temp/cisco_sna_netflow_example_1200_rows.csv` has 116 columns,
1,200 data records, 1,275,537 bytes and SHA-256
`b03c98cf0124f426e05d1b87f6ca1cf38c5f8a145e22c7dec6d3f527b99e70c3`.
Preserve those bytes. Its compact numeric timestamp offsets and nullable ICMP
ports need an explicit source-profile contract. Subject/peer ordering is a
projection convention, not inferred client/server or packet direction; map
connection totals once and retain additional source values as unmapped raw.

| Workstream | Dependency | Status | Exit and validation |
| --- | --- | --- | --- |
| S64 Export-example specification | Clean committed remediation | DONE | Earlier adoption evidence retained; reopened default-policy obligation closed by S68/S69 and final V57. |
| S65 Profile implementation | S64 | DONE | Registry-driven profile admission/defaults, strict compact-offset mode, nullable ports and browser mapping agree; authored contracts generate successfully; focused old/new profile tests pass. |
| S66 Example and graph proof | S65 | DONE | Complete comparisons and fresh-page reopen proof closed by S70/S71 and final V56/V58. |
| S67 Validation and handoff | S66 | DONE | Current V59 closure is recorded above; earlier evidence below remains historical. |

Each workstream records its changes, validation, risks and exit here before its
successor starts. Existing source-profile fingerprints, historical migrations
and frozen fixture bytes remain stable. The new profile is additive; no data
migration or implicit reinterpretation of an existing approved mapping is
planned. Final selection includes new example/parser/profile tests, existing
import and graph regressions, the upload/saved-graph browser workflow, generated,
schema and catalog checks, Go/types/boundary lint and Markdown. Prior 90-row
evidence is not current qualification for changed source.

Selection discovery: `nf-sna-example-guide-20261031-01 make task-guide
OWNER=module.networkflow` passed native 0 with the previously recorded pinned
Go/PATH and installed-only prefix. New selectors will be authored and recorded
before execution. Development and final evidence follow below.

S64 exit: `nf-sna-example-s64-close-20261031-01 make lint-markdown` passed
native 0; retained summary:
`.cartulary/test-results/nf-sna-example-s64-close-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
An earlier `nf-sna-example-s64-docs-20261031-01` invocation did not retain its
native session exit in the conversation; the closing run supplies complete
command evidence. The new source profile preserves all original source values,
uses nullable ports and connection totals, and explicitly defines graph direction.
It requires no historical migration edits, target-column changes, row-ID algorithm
changes or state-version changes. S65 must replace assumptions that the first
registry profile governs every mapping; original-profile vectors must remain
stable. Closing tracker check uses `nf-sna-example-s64-checkpoint-20261031-01`.

S65 execution: the new exact row is
`module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0`,
authored with `make author-test-row-id` in `nf-sna-example-s65-row-20261031-01`
(native 0). The existing import-model row now includes the Flow Search case.
Generation `nf-sna-example-s65-generate-20261031-02` failed (native 2) because
that new client case initially lacked catalog coverage; its authored selector
is now registered. Generation 01 passed before that case existed. The earlier
format invocation has no retained native outcome and is not qualification.
Focused selection: new profile, exact timestamp grammar, preview admission and
client import model; final checks will include metadata, old fixtures and types.

S65 exit: authored registry and schema projections now include the separate
Flow Search profile. Backend mapping defaults/required fields/empty values and
public discovery use the selected profile; discovery advertises only actually
mappable optional fields. The UI resets suggestions and timestamp defaults when
changing profiles. Compact offsets remain isolated from strict RFC3339.
No migration or original-profile fingerprint change is introduced.

Focused evidence (all under `.cartulary/test-results/`, cache off, pinned prefix):

| Run ID | Command / selection | Native exit |
| --- | --- | --- |
| `nf-sna-example-s65-generate-20261031-03` | `make generate` | 0 |
| `nf-sna-example-s65-unit-20261031-02` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.unit.import_preview_admission_f0dc5f2d8b` | 0 |
| `nf-sna-example-s65-client-20261031-01` | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 0 |
| `nf-sna-example-s65-types-20261031-02` | `make lint CHECK=types` | 0 |

Unit run 01 failed to compile an old test call after the private helper gained
its profile argument; the call is corrected. Types run 01 found a misplaced JSX
paragraph; the markup is corrected. Both failures exited 2, retained their
artifacts, and completed cleanup without failures. These are development runs,
not final frozen-source qualification. S66 remains responsible for the unchanged
full export, persisted provenance, graph aggregates and saved-graph workflow.
S65 documentation checkpoint: `nf-sna-example-s65-docs-20261031-01 make lint-markdown`.

S65 Markdown passed native 0. S66 adds fixture 029 (revision 2 after correcting
its independently authored system-derivation fingerprint before execution),
with unchanged original bytes. New selectors authored through Make, native 0:
`module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488`
and `module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179`
(`nf-sna-example-s66-row-20261031-01` and `-02`). The oracle uses independent
CSV parsing, integer sums, UTC formatting and the published digest transcripts;
it was not captured from the implementation. Comparison covers all persisted
row members (raw provenance by a complete sorted-key JSON digest), all endpoint
and edge properties, preview counts, durable publication and unchanged replay.
The browser supplies the upload and saved-graph proof with the same admitted file.

S66 review correction: Table 17-B2 still described discovery as one profile.
It now projects the two adopted profiles with per-profile modes. Discovery is
sorted by ID independently of the registry's original default-profile order;
the unit row asserts ordering and supported optional fields. This closes the
missed S64/S65 discovery projection before S66 acceptance. The first complete
example backend run passed native 0 (`nf-sna-example-s66-backend-20261031-01`),
including independent row/digest/provenance and graph comparisons. Subsequent
final qualification will cover the discovery correction on frozen source.

S66 browser run 01 exited 2 after the new test timed out waiting for a
redundant result fetch after a successful refresh. Retained network evidence
shows refresh accepted (202), completed declaration reads and the existing
successful immutable result; no second result download is required. The
assertion now observes the next successful materialization generation and
retained result identity, and compares saved/live vertex and edge properties.
Cleanup completed with 23 steps and zero failed/blocked steps. No product
change or timeout increase was made for this test assumption.

S66 browser run 02 exited 2 at the added live/saved property comparison:
projection-specific IDs impose different array order. The assertion now joins
by endpoint value / Network Flow edge ID before comparing every property;
ordering inside either result remains governed by Graph Projection. Cleanup
again completed 23 steps with no failure or blockage. Source-profile discovery
unit run `nf-sna-example-s66-unit-20261031-01` passed native 0.

S66 browser run 03 passed saved/live property equality but exited 2 while the
new refresh predicate confused declaration state (`active`) with job status
(`succeeded`). The corrected assertion observes both the next active generation
and its distinct successful job. Cleanup passed all 23 steps. The required
public declaration/job distinction is preserved; no runtime changes or timeout
extension were introduced. `nf-sna-example-s66-format-20261031-01 make format`
passed native 0; formatting did not alter frozen fixture bytes.

S66 exit: `nf-sna-example-s66-browser-20261031-04` passed native 0 through
original-file upload, 50/50 preview, 1,200/0 apply, 173 vertices / 944 edges,
exact counter totals, all saved/live properties and a separate successful
refresh job with generation advancement. Backend run 01 independently matches
all persisted row values, identities, digests and provenance plus graph
properties and unchanged import replay. Fixture revision 2 manifest SHA-256:
`f00d4dae65e750e79e1dc39e3522aa117554879a434af6ed427c41965e38ac6a`.
No original source bytes, old profile data or historical fixture bundles changed.
The [upload walkthrough](../guides/network-flow-cisco-sna-example.md) documents
subject-to-peer semantics, all mappings, totals and compatibility. No running
application was populated. Required final verification remains S67.

S66 documentation checkpoint:
`nf-sna-example-s66-docs-20261031-01 make lint-markdown`.

S66 Markdown passed native 0; browser cleanup completed all 23 steps and its
fixture summary reports `passed`. S67 uses the following affected portfolio
(25 selected product rows), plus generated/schema/catalog source checks,
Go/types/backend-boundary lint and Markdown. The new example and its dependency
changes require fresh qualification after finalization; earlier runs above are
development evidence. Existing unaffected fixture completions remain historical.

| Group | Exact selection |
| --- | --- |
| unit | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.cisco_flow_search_profile_admits_compact_timesta_3d18779db0,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.network_flow_selector_covers_cisco_sna_required_66f0e756eb,module.networkflow.unit.network_flow_selector_covers_cisco_sna_v1_permit_a2c32abbbf,module.networkflow.unit.network_flow_selector_covers_every_network_flow_13fc533c41,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb` |
| integration | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.effective_resource_limit_discovery` |
| browser | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179,module.networkflow.browser.import_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619` |
| client | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.networkflowimportmodel_suite_12249943c5,web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a` |

All commands use this environment prefix, with the run ID listed per invocation:

```sh
PATH=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin:$PATH \
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go \
CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_HARNESS_CACHE_MODE=off \
CARTULARY_TEST_RUN_ID=<listed-run-id> make <listed-target-and-selection>
```

`make agent-finalize` runs after formatting and before this portfolio. No full
warm-check root is available, so optional retained-run maintenance is skipped;
no redundant full check is planned solely to obtain such a root.

S67 preparation: format 01 passed native 0. Finalize 01 exited 2 at its
source-shape preflight because formatting the authored JSON-shape checker
invalidated its generated topology input digest. No finalization mutation was
performed; all ten cleanup steps completed. Regenerate through Make before
finalize 02, then inspect and freeze source. Failure summary:
`.cartulary/test-results/nf-sna-example-s67-finalize-20261031-01/unit-artifacts/finalize-summary.json`.

S67 source freeze: generate 01 and finalize 02 passed native 0. Generated
changes were inspected: the additional profile/schema changes flow through
Network Flow and Imports contract projections; browser scheduling and topology
hashes reflect only the authored selector/checker changes. Historical migrations
and all earlier frozen bundles are unchanged. The copied CSV remains byte-for-byte
identical to the supplied file. Final test groups use run IDs
`nf-sna-example-s67-{unit,integration,browser,client}-20261031-01` with the exact
selections above. Required source/lint checks and final evidence reconciliation
follow; no final completion claim is made until their native exits and cleanup
are known.

S67 first frozen portfolio: integration, client, generated/schema/catalog,
Go and backend-boundary checks passed native 0 at source
`sha256:62e0f4c410d2b447a70f2d82985c780ede9c37840c9e0167523d288348fbaed6`.
Unit and types runs exited 2. The corpus integrity helper embedded historical
remediation assumptions (`revision >= 3`, test-name prefix `TestNetworkFlow`)
that are not manifest requirements. It now admits any positive registered
revision and Go test name while preserving exact manifest/registry revision
agreement, unique test bindings, full inventory equality and file integrity.
The UI label condition was accidentally applied to the uptime export-mode
selector as well; that unrelated selector again uses its original labels.
These are the only post-freeze repairs. Unit/types cleanup completed all 10/11
steps, respectively, with no failed or blocked cleanup. Finalization and the
same affected portfolio will be repeated on the repaired source; first runs
are retained without being represented as final qualification.

S67 final code qualification: format 02 and finalize 03 passed native 0.
The repaired source is frozen at
`sha256:59ee80bce70dbd538c558ee43a094dbb10833c1bc3705bdd01205e4798e53602`,
dirty relative to `d936506e2b851a22e8c9fbddd983ab4fd6aa21de`.
All ten final invocations below share that source, toolchain
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system identity
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Their exact roots are `.cartulary/test-results/<run-id>/`; each contains
`run-manifest.json`, `run-summary.json`, `target-summaries/<target>.json`,
selected `rows/` and `cleanup-results.json`. Native exits were collected
separately from those artifacts. All listed cleanup steps completed with zero
failed or blocked steps.

| Run ID | Command / selected group above | Native exit | Selected product rows | Cleanup steps |
| --- | --- | ---: | ---: | ---: |
| `nf-sna-example-s67-unit-20261031-02` | Unit selection, `make test-slice` | 0 | 14/14 | 10 |
| `nf-sna-example-s67-integration-20261031-02` | Integration selection, `make test-slice` | 0 | 5/5 | 14 |
| `nf-sna-example-s67-browser-20261031-02` | Browser selection, `make test-slice` | 0 | 3/3 | 23 |
| `nf-sna-example-s67-client-20261031-02` | Client selection, `make test-slice` | 0 | 3/3 | 12 |
| `nf-sna-example-s67-generated-20261031-02` | `make source-check CHECK=generated` | 0 | n/a | 14 |
| `nf-sna-example-s67-schema-20261031-02` | `make source-check CHECK=schema` | 0 | n/a | 12 |
| `nf-sna-example-s67-catalog-20261031-02` | `make source-check CHECK=catalog` | 0 | n/a | 11 |
| `nf-sna-example-s67-go-20261031-02` | `make lint CHECK=go` | 0 | n/a | 15 |
| `nf-sna-example-s67-boundary-20261031-02` | `make lint CHECK=backend-boundary` | 0 | n/a | 13 |
| `nf-sna-example-s67-types-20261031-02` | `make lint CHECK=types` | 0 | n/a | 11 |

Fixture 029's current behavioral summaries are:

- `.cartulary/test-results/nf-sna-example-s67-integration-20261031-02/NF-FIX-029-cisco-sna-flow-search-3319697475/comparison-summary.json`
- `.cartulary/test-results/nf-sna-example-s67-browser-20261031-02/NF-FIX-029-cisco-sna-flow-search-node-hUoAfg/comparison-summary.json`

Both report `passed` for frozen revision 2 and the manifest hash recorded above.
The backend summary references the complete retained transport/state and
independent comparison output; its comparator also rejects a deliberately
incorrect expected value. Current integrity checks cover all **29** registered
bundles. Behavioral requalification here is the affected subset; no claim is
made that every historical fixture or the complete owner was rerun.

Compatibility and handoff: the new profile, compact-offset mode and discovery
entry are additive. Original-profile registry values and all earlier frozen
bundles are unchanged; old publication/interface/diagnostic fixtures pass.
Discovery also stops advertising unsupported optional targets. Server and
generated browser contracts must be deployed together for the new profile.
No SQL columns, historical migrations, row-identity algorithms, major 8/state 5,
or existing approved mappings changed. Before any new-profile durable import,
rollback is a source/test reversion. Once such imports exist, preserve support
for their immutable mappings; no automatic data conversion, reset or downgrade
is provided. The walkthrough is the operational upload/saved-graph handoff.
No user application was populated, and no commit or deployment was performed.

No known required runtime correction remains. Full CI, complete-owner
qualification, visual/performance certification, release/deployment and optional
warm-check retained-run maintenance are outside this affected portfolio.
Closing documentation check:
`nf-sna-example-s67-docs-20261031-01 make lint-markdown`.

S67 exit: all 25 selected product rows, required repository checks, fixture
comparisons and cleanup passed on the final source above. Closing Markdown run
`nf-sna-example-s67-docs-20261031-01` passed native 0; summary:
`.cartulary/test-results/nf-sna-example-s67-docs-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
S64–S67 are DONE. The final status-only documentation checkpoint is
`nf-sna-example-s67-docs-20261031-02 make lint-markdown`; product source remains
frozen. The original CSV hash and revision-2 manifest hash were rechecked after
finalization, and `git diff --check` reports no whitespace errors.

## 1. Scope and Source Posture

### Current implementation iteration — IMP-08

**COMPLETE through S36, including F59–F62 review corrections.** The maintainer
authorized implementation on 2026-10-31 and confirmed migration 34 remains
unapplied. S58–S63 and reopened S36 are DONE. All 90 selected rows, all 28 fixture
comparisons, required source/lint checks, final Markdown and the migration build
pass at source
`sha256:986a41d8eb5b59c7af5e02b8da00c1a3e26de4ffcc6ab65377c03318d9dee961`.
[§29.7](#297-s36--final-evidence-ledger-and-compatibility-handoff) is the current
qualification and operational handoff. F36–F62 are resolved within the agreed
affected scope. The earlier §27.64 88-row qualification remains historical.
Migration 34 was tested in isolated fixtures, not deployed to a user database.
No commit, reset, release claim or complete-owner qualification was performed.

The maintainer explicitly authorized the complete remediation plan, including
all 28 fixture families, required runtime corrections, fixture execution
mechanics, and import-boundary cleanup. [§27](#27-imp-08--complete-fixture-and-import-remediation)
records the earlier execution; §29 controls the review corrections. DOC-03's eight-bundle limit and T20 fixture
deferral are superseded. Earlier results retain their original source and scope.
The first execution action was the tracker update. Each workstream recorded its
validation and tracker checkpoint before its successor began; S36 closes the
final validation and handoff slice. No commit, deployment, reset, release
qualification, or complete-owner audit is included.

### Historical planning iteration — DOC-03

Temporal language and authorization limits in this historical section refer to
DOC-03 only; they are superseded by the completed IMP-08 execution above.

The current request authorizes a **document update only**. DOC-03 records the
next bounded iteration for legacy cleanup and import hardening; S32–S36 are
future implementation work, all with status **TODO**. No production, test,
fixture, contract, generated, migration, dependency, deployment or historical
data-conversion changes are authorized by this invocation.

The current baseline is clean `main` at
`98adcc82961456bcb74bbecd0023136485afbf00`. The target remains
`internal/modules/networkflow`: 138 tracked Go files, comprising 80 non-test
and 58 test files. Current inspection is targeted, not an exhaustive module or
conformance audit. [§26](#26-doc-03--legacy-cleanup-and-import-hardening) contains
the authoritative DOC-03 inventory, F36–F38, S32–S36, V29–V33, readiness gates,
and handoff. Sections 2–12 retain DOC-01 history with links to their current
supplements; §§13–25 retain the subsequent implementation history.

The agreed sequence is **S32 -> S33 -> S34 -> S35 -> S36**. Each future slice
requires passing validation and a tracker checkpoint before its successor.
S32 is technically ready for a later authorized task; S33–S36 are blocked by
their preceding evidence gates. Planning completeness, technical readiness and
implementation authorization are separate outcomes.

Preserve adopted features, public interfaces, major 8/state 5, current versioned
schemas, rejection guards and historical migrations. Retire the materialized
apply-validation policy retained for tests, separate normalized row input from
persisted identity/provenance, and correct behavioral execution for fixture
bundles 001–005, 020, 021 and 023. Broader fixture fidelity remains explicit
production-readiness debt. Acceptance is focused; no full CI, complete-owner
qualification, release or deployment claim is included.

### Historical implementation iteration — IMP-07

**DONE:** S29–S31 and S15-R4 close F34/V28 OAuth scope-description classification
and F35/V27 independent preview-retention validation. These obligations and the
corresponding S15-R3 acceptance portions were reopened and are now qualified by
the affected-gap portfolio in §25. Earlier passes remain evidence of their
recorded coverage; the broader S15-R3 integration/browser portfolio is historical.

### Historical implementation iteration — IMP-06

**DONE:** S25–S28 and S15-R3 implement and validate the maintainer's byte-only
CSV, complete security accounting, and absent-field choices. F32/V27 and F33/V28 pass the frozen portfolio in §24;
S15-R3 supersedes the affected S15-R2 acceptance claims within that explicit
scope. Earlier passing evidence remains unchanged and supports only its actual
coverage. No broader current Network Flow owner qualification is inferred.

### Historical implementation iteration — IMP-05

**DONE:** S21–S24 and S15-R2 implement and validate approval integrity, explicit
route semantics and legacy retirement. F29–F31 and the affected specification,
mapping-approval and final-acceptance claims are reconciled in §23. Prior-format
compatibility obligations are DROPPED under the adopted fresh-environment
decision. Earlier implementation and retained runs remain historical evidence
of their actual coverage.

The selected boundary requires a fresh database initialized through the canonical
migration chain. Earlier pre-production import plans are unsupported; their
registry translation and exclusive fixtures/harness support are removed.
Every public operation declares mutation semantics explicitly. Network Flow
major 8/state 5 and the coordinated API candidate remain unchanged. No database
reset, commit or deployment is authorized. Each workstream must pass its required
checks and record a checkpoint before its successor begins. See §23.

### Historical implementation iteration — IMP-04

**Historical completion claim, affected exits reopened by IMP-05:** the four-gap remediation authorized on 2026-10-29 completed
S16–S20 and S15-R. F25–F28 and the reopened S09, S12-I/V19, S13, S14 and S15
acceptance obligations are closed by §22's compatible focused evidence and
operational handoff. Earlier passing and failing runs remain historical evidence
of their actual coverage. S10, S11, S12-M and S12 retain their completed
implementation posture with affected final regressions passed.

The controlling sequence is **S16 -> S17 -> S18 -> S19 -> S20 -> S15-R** in §22.
Existing uncommitted changes and S00–S15 history were preserved. The user adopted
target-neutral CSV admission: source custody precedes target-specific
interpretation, and target selection remains after upload. Explicit owner
adoption preceded dependent implementation. Transaction-ownership corrections
implement the adopted Imports commit rule. Major 8/state 5 remains the release
boundary; commit and deployment are separate unperformed actions.

### Historical implementation iteration — IMP-03

**Historical completion claim, superseded by IMP-04:** S09–S15 were recorded DONE,
including S12-M/S12-I, with F18–F24 and RB-004/RB-005 closed. §22 reopens the
affected acceptance claims without erasing §21's owner adoption, source identity,
focused validation results, or operational handoff. Deployment remains separate.

On 2026-10-29 the user requested implementation of the revised remediation plan. This authorizes specification preparation, characterization, implementation, authored migrations, generated projections and focused validation. The first execution change is this tracker checkpoint. Existing DOC-02 changes and S00–S08 history are preserved. The implementation baseline is `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`; this tracker is the sole pre-existing modified file, with no staged or non-ignored untracked files.

The current sequence is **S09 -> S10 -> S11 -> S12-M -> S12 -> S12-I -> S13 -> S14 -> S15**. Each workstream requires completed validation and a tracker checkpoint before its successor begins. §21 controls this execution and supersedes DOC-02's tracker-only authorization, no-migration assumption, unchanged internal-source-port assumption and target-independent source-row admission. Human owner adoption remains a distinct S09 gate before dependent implementation; deployment is not authorized. Tests and generated/runtime evidence must not consume Markdown.

### Historical planning iteration — DOC-02

This is the sole controlling plan and execution record for Network Flow. The current request authorizes a **document update only**: record the next iteration for legacy removal and bounded-resource hardening. S09–S15 are future implementation work; this invocation authorizes no production, test, contract, generated, migration, deployment, deletion, or conversion changes. The user explicitly included the necessary Imports source-reader dependency and excluded a full CI rerun.

DOC-02 baseline: clean `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, revalidated before editing. The current target has 132 Go files: the original 129-file inventory plus three files added by the completed remediation. Current inspection coverage, owner decisions, F18–F23, S09–S15, validation obligations and handoff are in §§16–20. That targeted inspection is not a new exhaustive module or conformance audit.

The current plan is `S09 -> S10 -> S11 -> S12 -> S13 -> S14 -> S15`. Each future implementation slice is a separate workstream. Its validation and tracker checkpoint must finish before its successor begins. Planning completion is separate from implementation readiness and authorization; the current readiness and completion records are in §§18 and 20.

Major 8 / document 8.0.0 / durable state 5 / minimum admitted state 5 remains the coordinated release boundary. Human owner adoption and deployment remain separate gates. Older data remains untouched; original CSV sources and mapping decisions are required for re-import. No new public major, durable state, parser-profile ID, compatibility alias, or migration is proposed merely to remove dead code or correct implementation against an unchanged owner definition.

### Completed remediation history — S00–S08

The remainder of this section records the completed remediation; §§2–12 retain DOC-01 planning history, and §§13–15 retain its implementation evidence and handoff. Their original baselines, statuses and authorization statements are historical, not current instructions or acceptance evidence for S09–S15. Current planning decisions supersede earlier preservation restrictions only where explicitly identified below, including F18's targeted retirement of the inert contracts retained under F10.

On 2026-10-29 the user explicitly authorized the previous complete remediation plan, including owner amendments, tests, implementation, generated projections, a diagnostic storage migration, and a coordinated fresh-state cutover. That work completed as S00–S08. Deployment, destructive resets, automatic deletion, and historical data conversion were not authorized.

### Completed remediation execution

Baseline revalidated: `main` at `a182673ae3b460de9c0d078d3315f17e18c4f36a`; tracked files clean; this tracker is the sole pre-existing untracked file. The original 129-file inventory and DOC-01 evidence remain historical baseline evidence, not current test results.

The user approved implementation of major 8 / document 8.0.0 / durable state 5 / minimum state 5. Owner amendments remain **proposed** pending human owner review and explicit coordinated adoption; implementation authorization, passing tests and deployment adoption are separate gates. Prior state is left untouched and fails admission. Original sources are required for re-import into fresh state.

Execute exactly `S00 -> S03 -> S01 -> S02 -> S04 -> S05 -> S06 -> S07 -> S08`. Update this tracker after each workstream's validation and before beginning its successor. Tests, generators, runtime metadata, conformance and release evidence must never consume Markdown. Source requirements and projection fidelity are reviewed by humans.

| Slice / workstream | Status | Depends on | Goal / binary exit |
| --- | --- | --- | --- |
| S00 Specification closure and tracker revision | DONE | baseline | Coordinated owners and complete finding/validation traceability |
| S03 Canonical tracker validation | DONE | S00 | Module-tracker glob and passing Make Markdown lint |
| S01 Contract characterization | DONE | S03 | Dedicated registered preview baseline and safe Imports failure tests pass |
| S02 Transport-independent semantic boundary | DONE | S01 | Shared strict decoder and transitive import-root guard pass |
| S04 Diagnostic semantics and storage identity | DONE | S02 | Correct identities/messages/privacy and table-scoped uniqueness pass |
| S05 Complete previews and accurate retention | DONE | S04 | Complete bounded preview; exact ordered apply retention and rejected-row counts pass |
| S06 Strict preview-result admission | DONE | S05 | Closed typed results, nested values and mapping consistency pass |
| S07 Coordinated major/state/client cutover | DONE | S06 | Major 8/state 5 publication, fresh admission and unchanged old-state rejection pass |
| S08 Validation and handoff completion | DONE | S07 | Finalized source, user-revised focused portfolio and complete handoff |

The original S01/S02 IDs are retained. Active statuses in this table and §§14–15 supersede the explicitly historical planning statuses; S08 reconciled every current finding and blocker without erasing DOC-01 evidence. F03/RB-002 is resolved by retaining owner-local harness controls with ordinary-build isolation and runtime scan inclusion. F04-F07 and F09 remain supported boundaries; F08's zero-generated-diff expectation is superseded by S07; F10 does not authorize cosmetic historical renaming.

| New finding | Required remediation / areas | Rationale, benefit and unresolved risk | Compatibility / validation / slice |
| --- | --- | --- | --- |
| F13 Incomplete preview-result admission | Exact required presence, closed nested types, mapping and diagnostic equality; implementation/tests/owner clarification | One trustworthy owner boundary; otherwise malformed results and inconsistent approvals can escape | Valid schemas unchanged; malformed results reject safely; S06 |
| F14 Diagnostic construction drift | Owner ID bytes, closed reasons, messages, attribution and redaction; implementation/tests/fixtures/docs | Stable provenance and privacy; otherwise collisions, misleading values or disclosure | Correct erroneous emitted values under fresh-state cutover; fixed vectors and preview/apply parity; S04 |
| F15 Preview/retention accounting conflated | Separate preview completeness and apply retention; count rejected records independently; implementation/tests/spec | Predictable bounded validation; otherwise low limits disable preview and exact-limit flags/counts lie | Correct only nonconforming behavior; zero/exact/overflow and 50/51 cases; S05 |
| F16 Global diagnostic uniqueness | New authored migration to table+diagnostic identity; implementation/SQL/tests | Same source can appear in distinct tables; otherwise legitimate imports fail | Physical schema change, no data rewriting; cross-table success/same-table rejection; S04 |
| F17 Retained diagnostic admission absent | State-5 owner validator and coordinated current-only publication; spec/contracts/code/tests/docs | Prevent incompatible facts from being served | Major 8/state 5; old state rejected unchanged and fresh/restored valid state admitted; S07 |

F01/F02 -> S02; F11/G01/RB-001 -> S01; F12/RB-003 -> S03. Each workstream retains exact commands, exits and artifacts in the execution log below. No historical success is promoted into current evidence.

### Remediation execution log

| Session / slice | Changes / decisions | Commands / results / evidence | Next gate |
| --- | --- | --- | --- |
| 2026-10-29 / S00 start | Baseline and authorization revalidated; controlling tracker updated first | Read-only Git and applicable instructions; no product verification yet | Owner closure, then S03 |
| 2026-10-29 / S00 completion | Prepared user-approved major 8/state 5 amendments across NF, Extensions and Core; explicit diagnostic contracts and current-only policy | Agent source review; human owner adoption not established; product execution not claimed; tracker lint selection pending S03 | S03 |
| 2026-10-29 / S03 completion | Module-tracker glob covers both trackers; RB-003 resolved | `CARTULARY_TEST_RUN_ID=nf-s03-markdown-20261029-01 make lint-markdown` exit 0; `.cartulary/test-results/nf-s03-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json` | S01 |
| 2026-10-29 / S01 completion | Dedicated preview admission row registered; G01/RB-001 closed for decoder movement | Make generate/format passed; `test-slice` NF preview/error/boundary rows passed at `.cartulary/test-results/nf-s01-baseline-20261029-01`; Imports safe owner translation unit/integration rows passed at `.cartulary/test-results/nf-s01-imports-20261029-01`; outer exits 0 | S02 |
| 2026-10-29 / S02 completion | F01/F02 fixed; strictjson used directly and import roots guarded | Focused NF rows exit 0 at `.cartulary/test-results/nf-s02-focused-20261029-01`; backend boundary lint exit 0 at `.cartulary/test-results/nf-s02-boundary-20261029-01` | S04 |
| 2026-10-29 / S04 validation repair | New duplicate-rejection test poisoned its borrowed rollback transaction; isolate expected SQL failure in a nested transaction | `nf-s04-focused-20261029-01` failed product assertion (outer 2); unit identity passed, schema check passed at `nf-s04-schema-20261029-01`; migration-drift not reached by command chain | Rerun affected store row and migration drift |


### Historical planning baseline

The following baseline scope table and §§2–12 preserve DOC-01 exactly as planning history. Their restrictions, readiness states, skipped checks and completion claims are superseded by the active remediation table and execution evidence. They do not describe this implementation's current authorization or acceptance.

| Scope item | Recorded value |
| --- | --- |
| Supplied and canonical target | `internal/modules/networkflow`, resolving to `/home/jochi/code/cartulary/internal/modules/networkflow` |
| Derived label | `networkflow`: last target-path component, normalized to lowercase kebab case; matches `[a-z0-9]+(?:-[a-z0-9]+)*` |
| Output / sole allowed write | `docs/handoffs/networkflow-module-refactor-tracker.md` |
| Framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, read first; doctrine/template, not current-state proof |
| Baseline | Branch `main`; commit `a182673ae3b460de9c0d078d3315f17e18c4f36a`; initially clean index/worktree; no nonignored untracked files; tracker absent |
| Admission | Target and output parent are canonical repository directories, with no symlink traversal; no target symlinks, nested AGENTS, ignored or nonignored untracked target files found |
| File universe | 129 tracked Go files: 115 in the package root (67 production/support, 48 tests), 14 in `harnesscontrol` (8 support, 6 tests); 75 non-test and 54 test files total |
| Allowed changes | This tracker only, including its planning decisions and session history |
| Non-goals | Code movement, patches, test execution/addition, generated edits, dependency changes, migrations, runtime changes, harness changes, frontend redesign, release or conformance claims |
| Planning outcome | Complete for the declared target inventory and bounded structural recommendation; broader behavioral conformance and exhaustive caller coverage are not asserted |
| Slice readiness | S01 ready for a later authorized characterization task after baseline revalidation; S02 blocked by RB-001 |
| Implementation authorization | Not granted by this planning invocation; “implement this plan” means create this tracker |

Authority is applied by named scope, in this order: adopted subsystem NLSpecs; Core 00–04 implementation-conformance requirements; Core 05 only for claim-bearing timed/fixture-sensitive publication; domain vocabulary and implementation-support guidance; current code/tests as implementation evidence; prior plans/handoffs/framework as evidence only. Typed contracts and test rows are downstream projections. A matching test does not prove specification completeness. Unresolved applicable primary-owner conflict must be recorded as **BLOCKED: owner contradiction**; none was established in the inspected clauses. A supporting-file mismatch is not automatically such a contradiction.

| Evidence ID | Owner/support source inspected | Authority and inspection scope |
| --- | --- | --- |
| E01-NF | `docs/network-flow-activity-nlspec.md` | Adopted/current 7.0.0, public major 7, durable state 4, minimum migratable 3. Scope NF-REQ-001–006; imports/rows/query/graph/link/security/routes/saved graphs; §§27–29, especially NF-REQ-210–212. Existing facade/application boundary is required, not a new redesign proposal. |
| E01-GP | `docs/graph_projection_nlspec.md` | Adopted/current 2.2.0; scope/native V2 boundary and result/lease/restore ownership. NF imports §§8–9, while its own §29 governs flow-source contributions. |
| E01-EXT | `docs/extension-subsystem-nlspec.md` | Adopted/current 0.12.0, contract major 2; status/scope and discovery/claim coordination. Historical revision explanations do not override the current NF major-7 owner. |
| E01-RPT | `docs/reporting-subsystem-nlspec.md` | Adopted/current 1.4.0; status and scope separation. NF §29 defines the source-side exact graph lease/read/release obligations used here. |
| E01-HAR | `docs/testing-harness-nlspec.md` | Adopted/current; relevant Network Flow control families and §12.2.9. Registry tests alone cannot prove real product rollback/audit behavior. |
| E01-C00 | `docs/spec/00_document_set_status_and_precedence.md` | Precedence, adoption and scope allocation; generic graph and analytical import owner boundaries. |
| E01-C01 | `docs/spec/01_architecture_storage_and_view_contracts.md` | Required modules, current-record ownership, HTTP/discovery/view/import/restore interfaces; source-owner interfaces remain narrow. |
| E01-C02 | `docs/spec/02_domain_model_schema_and_history.md` | Core record/indicator terminology and indicator participant REQ-02-074C; analytical rows cannot become Core observation source records. |
| E01-C03 | `docs/spec/03_workbook_interaction_collaboration_and_workflows.md` | Workbook/import/collaboration and save-state sections; durable resource invalidation belongs to Collaboration. |
| E01-C04 | `docs/spec/04_security_deployment_and_conformance.md` | Route admission, including REQ-04-105A, and transaction/audit/privacy obligations including REQ-04-135–138. |
| E01-SUP | `docs/domain.md`; `docs/research/nlspec-spec.md`; supplied framework | Vocabulary, source posture and planning structure only. Core 05 applicability was assessed; no timed/publication claim is made and no capacity certification was run. |
| E01-PROC | `AGENTS.md`, `internal/AGENTS.md`, `tools/AGENTS.md`; refactor-tracker skill and its `references/tracker-format.md`; cartulary-test-ops skill and command recipes | Repository procedure, boundary and harness mechanics; planning-only scope takes precedence over implementation examples in these sources. |

Repository files inspected are accounted individually by E02-001–E02-129 in §2. The external inspection ledger E03–E09 is in §10. “Inspected” means the exact file was opened: imports/declarations plus relevant bodies or selected test assertions. It does not mean every line was audited. The proposed edit seam, its decoder implementation, existing semantic guard, and related Imports call sites received direct body-level inspection. No inventory file is excluded or left metadata-only.

The next discovery boundary is intentionally narrow: revalidate the baseline and S01/S02 files and their Imports callers. A future move of state, worker, security, or harness responsibilities must deepen that family's characterization first; this tracker does not authorize such movement.

## 2. Current-State Repository Inventory

DOC-03 current supplement: [§26.2 Current inventory and inspection depth](#262-current-inventory-and-inspection-depth). The material below remains historical DOC-01 evidence.

All paths in the first column are inside the canonical target. D followed by a number records the inspected file's line count, not a claim that every line was behaviorally audited. Inbound “reference candidates” and “source references” are bounded lexical cross-references between directly opened files, not a resolved Go call graph or coverage report. Private receiver methods with exported spellings are not automatically supported public APIs. Contract IDs resolve in §4. No target file is generated; generated dependencies are recorded as consumed surfaces.

Tests listed here were inspected, **not executed**. A TODO for an exact direct test mapping is an entry gate if that file is later selected for change; the contract-level posture in §4 supports this bounded plan without claiming exhaustive coverage.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| [api.go](/home/jochi/code/cartulary/internal/modules/networkflow/api.go) | Profile IDs, owner validation errors and effective resource limits | ProfileID, RouteContributionID, WorkspaceContributionID, EffectiveLimits | same-package reference candidates: `configuration.go`, `extension_state.go`, `graph_restore_source.go` | `github.com/google/uuid`; 1 standard-library imports | source references: `application_composition_test.go`, `graph_capacity_test.go`, `graph_query_application_test.go`; +6 other source references | C01, C03, C16; no generated file authored here | Network Flow | medium | E02-001; D194 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [application.go](/home/jochi/code/cartulary/internal/modules/networkflow/application.go) | Serialized table rename/soft-delete commands, admission/replay and atomic effects | private tableApplication, tableMutationOutcome | same-package reference candidates: `table_http.go` | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/incidents/admission`, `internal/platform/authn`; 2 standard-library imports | indirect tests in C04, C08, C09; TODO: exact direct call-to-test mapping before editing this file | C04, C08, C09; no generated file authored here | Network Flow | high | E02-002; D82 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [application_composition.go](/home/jochi/code/cartulary/internal/modules/networkflow/application_composition.go) | One retained private application bundle from the module dependency set | private activeApplications; prepareActiveApplications | same-package reference candidates: `graph_result_cleanup_dispatcher.go`, `module.go`, `routes.go` | `internal/modules/graphprojection/postgresresult`, `internal/modules/imports`, `internal/modules/incidents/admission`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 2 standard-library imports | indirect tests in C01; TODO: exact direct call-to-test mapping before editing this file | C01; no generated file authored here | Network Flow | high | E02-003; D59 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [application_composition_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/application_composition_test.go) | Construction probes for coherent dependencies, inactive/active contributions and retained applications | test-only `RegisterHandler`, `constructionModule`; +1 other source references | query_authoring_test.go via production surface | `internal/modules/crossownertransaction`, `internal/platform/httpapi`, `internal/platform/jobs`, `internal/platform/postgres`; 3 standard-library imports | helper/bridge; query_authoring_test.go via production surface | C01; no generated file authored here | Network Flow owner tests | medium | E02-004; D139 (declarations/imports and selected assertions); not executed |
| [binding_store.go](/home/jochi/code/cartulary/internal/modules/networkflow/binding_store.go) | Owner binding identity lookup/insert and audit in caller transaction | private store binding methods | same-package reference candidates: `indicator_link.go`, `indicator_link_receipts.go`, `transaction_participants.go` | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/indicators`; 9 standard-library imports | indirect tests in C07, C09; TODO: exact direct call-to-test mapping before editing this file | C07, C09; no generated file authored here | Network Flow binding source; Indicators owns target | high | E02-005; D275 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [collaboration_producer.go](/home/jochi/code/cartulary/internal/modules/networkflow/collaboration_producer.go) | Append table/graph resource invalidation intents within owner transaction | ResourceIntent; ResourceIntentAppender | store.go; saved_graph_application.go; app/server/collaboration_intents.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 4 standard-library imports | store_test.go; routes_integration_test.go | C08, C09; no generated file authored here | Network Flow producer / Collaboration persistence | medium | E02-006; D84 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [configuration.go](/home/jochi/code/cartulary/internal/modules/networkflow/configuration.go) | Closed configuration overlay, limits, key-ring findings and normalization | Configuration, ResourceLimitOverrides, ConfigurationFinding; normalization/clone/overlay functions | app/configassembly/{configuration,deployment,extension_admission}.go | 7 standard-library imports | source references: `configuration_test.go`, `graph_capacity_test.go` | C01, C16; no generated file authored here | Network Flow configuration contribution | medium | E02-007; D458 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [configuration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/configuration_test.go) | Configuration overlays, typed findings and effective limit assertions | test-only `TestNetworkFlowConfigurationContribution_Unit`, `TestNetworkFlowEffectiveResourceLimits_Unit` | Go test entrypoints | `github.com/google/uuid`; 4 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C01, C16; no generated file authored here | Network Flow owner tests | medium | E02-008; D239 (declarations/imports and selected assertions); not executed |
| [csv_parser.go](/home/jochi/code/cartulary/internal/modules/networkflow/csv_parser.go) | Bounded headered CSV parsing, field normalization, accepted rows and rejected diagnostics | private parser and row validators | same-package reference candidates: `graph.go`, `import_facade.go`, `mapping.go`; +3 other source references | 13 standard-library imports | source references: `network_flow_contract_test.go`, `network_flow_unit_test.go`, `store_test.go` | C03, C05; no generated file authored here | Network Flow analytical import target | medium | E02-009; D577 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [digest.go](/home/jochi/code/cartulary/internal/modules/networkflow/digest.go) | Canonical JSON, source/row/mapping digests and deterministic flow IDs | private identity/canonicalization helpers | same-package reference candidates: `csv_parser.go`, `extension_state.go`, `graph.go`; +11 other source references | `github.com/google/uuid`; 11 standard-library imports | source references: `graph_capacity_test.go`, `graph_streaming_test.go`, `graph_temporal_test.go`; +4 other source references | C03, C05, C06; no generated file authored here | Network Flow | medium | E02-010; D300 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [entropy.go](/home/jochi/code/cartulary/internal/modules/networkflow/entropy.go) | Instance-owned entropy reader for table IDs | private entropyReader | same-package reference candidates: `store.go` | 3 standard-library imports | indirect tests in C03, C14; TODO: exact direct call-to-test mapping before editing this file | C03, C14; no generated file authored here | Network Flow; harness injects test entropy | medium | E02-011; D26 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [extension_state.go](/home/jochi/code/cartulary/internal/modules/networkflow/extension_state.go) | Owner durable-family counts and validation | ExtensionStateReader; ExtensionStateFamilyCounters; ValidateExtensionState | saved_graph_admission.go; app/configassembly/extension_admission.go | `internal/platform/extensionstore`; 4 standard-library imports | source references: `extension_state_v3_integration_test.go`, `saved_graph_admission_integration_test.go` | C12; no generated file authored here | Network Flow extension contribution | medium | E02-012; D139 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [extension_state_v3_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/extension_state_v3_integration_test.go) | Reject obsolete saved semantic bytes without rewriting retained state | test-only `TestNetworkFlowExtensionStateV4RejectsV1WithoutRewritingBytes_Integration` | Go test entrypoint | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/networkflow`, `internal/platform/extensionstore`, `internal/platform/postgres`, `internal/testutil/revisionsupport`; 5 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C12; no generated file authored here | Network Flow owner tests | medium | E02-013; D149 (declarations/imports and selected assertions); not executed |
| [graph.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph.go) | Flow graph semantic admission, aggregates, projection input and contributor selectors | private graph request/composition types and helpers | same-package reference candidates: `extension_state.go`, `graph_query_application.go`, `graph_query_presentation.go`; +11 other source references | `github.com/google/uuid`; 8 standard-library imports | source references: `graph_capacity_test.go`, `graph_query_application_test.go`, `graph_streaming_test.go`; +6 other source references | C06, C10; no generated file authored here | Network Flow flow-specific semantics | medium | E02-014; D1278 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_capacity_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_capacity_test.go) | Capacity workload and host/runtime evidence capture; separate claim-bearing gate | test-only `TestNetworkFlowGraphCapacityCertification_Integration` | Go test entrypoint; Core 05 only if publishing timed/fixture claims | `github.com/google/uuid`; 11 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C06, C15; no generated file authored here | Network Flow owner tests | medium | E02-015; D412 (declarations/imports and selected assertions); not executed |
| [graph_cleanup_deadline_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_cleanup_deadline_test.go) | Stalled DB probe verifies cleanup deadline and rollback | test-only `BeginTx`, `QueryRow`; +3 other source references | graph cleanup integration assertions | `internal/platform/postgres`, `github.com/jackc/pgx/v5`; 4 standard-library imports | helper/bridge; graph cleanup integration assertions | C11; no generated file authored here | Network Flow owner tests | medium | E02-016; D64 (declarations/imports and selected assertions); not executed |
| [graph_materialization_payload.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_materialization_payload.go) | Strict shared materialization payload admission | private decodeGraphViewMaterializationPayload | same-package reference candidates: `graph_restore_source.go`, `graph_view_jobs.go`, `saved_graph_admission.go` | `internal/platform/strictjson`, `github.com/google/uuid`; 4 standard-library imports | source references: `graph_materialization_payload_test.go` | C10, C12; no generated file authored here | Network Flow | medium | E02-017; D36 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_materialization_payload_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_materialization_payload_test.go) | Malformed payload corpus and startup/restore admission probes | test-only `assertMaterializationPayloadAdmission`, `ObserveExecution`; +9 other source references | saved_graph_admission_integration_test.go / worker assertions | `internal/platform/authn`, `internal/platform/jobs`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 7 standard-library imports | helper/bridge; saved_graph_admission_integration_test.go / worker assertions | C10, C12; no generated file authored here | Network Flow owner tests | medium | E02-018; D139 (declarations/imports and selected assertions); not executed |
| [graph_materialization_timeout_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_materialization_timeout_test.go) | Deadline becomes terminal before publication; finalizer probes | test-only `TestGraphViewMaterializationDeadlineIsTerminalBeforePublication_Unit` | Go test entrypoint | `github.com/google/uuid`, `internal/platform/jobs`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C10; no generated file authored here | Network Flow owner tests | medium | E02-019; D75 (declarations/imports and selected assertions); not executed |
| [graph_projection_adapter.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_projection_adapter.go) | Single native Graph Projection V2 invocation and safe provider errors | private graphProjectionPort/adapter | same-package reference candidates: `graph_restore_source.go`, `graph_source_composer.go`, `module.go` | `internal/modules/graphprojection`; 11 standard-library imports | source references: `graph_capacity_test.go`, `graph_materialization_payload_test.go`, `graph_materialization_timeout_test.go`; +5 other source references | C06; no generated file authored here | Network Flow adapter / Graph Projection engine | medium | E02-020; D95 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_query_application.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_query_application.go) | Graph read admission, composition, contributors and completion audit | private graphQueryApplication; graphQueryComposer | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go`, `saved_graph_read_application.go` | `internal/modules/incidents/admission`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 4 standard-library imports | source references: `saved_graph_read_application_test.go` | C02, C06, C09; no generated file authored here | Network Flow | high | E02-021; D179 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_query_application_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_query_application_test.go) | Graph success/failure/audit/session completion with DB and writer probes | test-only `TestGraphQueryCompletion_Unit` | Go test entrypoint | `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/httpauth`, `internal/platform/postgres`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 7 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C06, C09; no generated file authored here | Network Flow owner tests | medium | E02-022; D176 (declarations/imports and selected assertions); not executed |
| [graph_query_presentation.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_query_presentation.go) | Graph and saved-graph contributor response resources | private response helpers | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go` | 0 standard-library imports | indirect tests in C06, C10; TODO: exact direct call-to-test mapping before editing this file | C06, C10; no generated file authored here | Network Flow response projection | medium | E02-023; D23 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_response_v2.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_response_v2.go) | Exact V2 result metadata, selector identity and response conversion | private graph result conversion | same-package reference candidates: `graph_routes.go`, `graph_source_composer.go`, `graph_view_routes.go`; +1 other source references | `internal/modules/graphprojection`; 2 standard-library imports | source references: `graph_streaming_test.go`, `graph_temporal_test.go` | C06, C10; no generated file authored here | Network Flow response projection | medium | E02-024; D201 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_restore_source.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_restore_source.go) | Restore source registration and bounded restored-job reconciliation | NewGraphRestoreSourceRegistration; ReconcileGraphRestoreJobsTx | app/recoveryassembly/graphprojection_restore.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/restore`, `internal/platform/jobs`, `internal/platform/postgres`; 3 standard-library imports | source references: `application_composition_test.go`, `graph_materialization_payload_test.go`, `graph_view_authority_integration_test.go`; +1 other source references | C12; no generated file authored here | Network Flow source contribution / Recovery orchestration | high | E02-025; D143 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_result_cleanup.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup.go) | Bounded result candidate sweep, lease cleanup and selected-result safety | private graphResultCleanupService | same-package reference candidates: `graph_result_cleanup_dispatcher.go` | `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/platform/postgres`; 4 standard-library imports | source references: `graph_cleanup_deadline_test.go`, `graph_result_cleanup_dispatcher_test.go`, `graph_result_cleanup_integration_test.go`; +1 other source references | C11; no generated file authored here | Network Flow selection / Graph Projection result storage | high | E02-026; D178 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_result_cleanup_dispatcher.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_dispatcher.go) | Periodic instance-owned cleanup scheduling and lifecycle | GraphResultCleanupDispatcher; NewGraphResultCleanupDispatcher; Start/Close | private package use; no distinct direct caller established in lexical map | `internal/modules/graphprojection`; 4 standard-library imports | source references: `application_composition_test.go`, `graph_result_cleanup_dispatcher_test.go`, `graph_telemetry_test.go` | C11; no generated file authored here | Network Flow | high | E02-027; D193 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_result_cleanup_dispatcher_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_dispatcher_test.go) | Scheduler lifecycle, nonoverlap and fatal-loss assertions | test-only `TestGraphResultCleanupDispatcherLifecycleSchedulingAndFatalLoss_Unit` | Go test entrypoint | `internal/modules/graphprojection`; 5 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C11; no generated file authored here | Network Flow owner tests | medium | E02-028; D231 (declarations/imports and selected assertions); not executed |
| [graph_result_cleanup_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_integration_test.go) | Selected/leased results protected; owner scoping, sweep bounds and commit failure | test-only `TestNetworkFlowGraphResultCleanupIsOwnerScopedSelectedSafeAndBounded_Integration` | Go test entrypoint | `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/networkflow`, `internal/platform/postgres`, `internal/testutil/revisionsupport`; 5 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C11; no generated file authored here | Network Flow owner tests | medium | E02-029; D206 (declarations/imports and selected assertions); not executed |
| [graph_result_cleanup_race_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_race_integration_test.go) | Publication and lease races against candidate deletion | test-only `TestNetworkFlowGraphCleanupPublicationAndLeaseRaces_Integration` | Go test entrypoint | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgxpool`, `internal/modules/auth/testsupport/storetest`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/networkflow`, `internal/testutil/appsupport`, `internal/testutil/networkflowsupport`, `internal/testutil/pgtest`; 3 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C10, C11; no generated file authored here | Network Flow owner tests | medium | E02-030; D319 (declarations/imports and selected assertions); not executed |
| [graph_result_cleanup_test_bridge_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_result_cleanup_test_bridge_test.go) | Test-only constructors/bounds for private cleanup service | test-only `NewGraphResultCleanupService`, `SetGraphResultCleanupTestBounds` | cleanup integration tests | `internal/platform/postgres`; 1 standard-library imports | helper/bridge; cleanup integration tests | C11; no generated file authored here | Network Flow owner tests | medium | E02-031; D19 (declarations/imports and selected assertions); not executed |
| [graph_routes.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_routes.go) | HTTP graph/query contributor extraction and response framing | private routeService handlers | private package use; no distinct direct caller established in lexical map | `internal/platform/httpapi`; 1 standard-library imports | source references: `graph_query_application_test.go` | C02, C06; no generated file authored here | Network Flow transport adapter | medium | E02-032; D70 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_source_composer.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_source_composer.go) | Shared route/worker/restore flow source traversal and projection composition | private graphSourceComposer; graphSourceReader | same-package reference candidates: `graph_query_application.go`, `graph_restore_source.go`, `graph_view_jobs.go`; +3 other source references | `github.com/google/uuid`; 4 standard-library imports | source references: `graph_query_application_test.go`, `graph_telemetry_test.go`, `graph_temporal_test.go`; +1 other source references | C06, C10, C12; no generated file authored here | Network Flow source owner | medium | E02-033; D317 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_streaming_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_streaming_test.go) | Aggregate retained-state bounds, limit outcomes, native V2 identity and SQL error propagation | test-only `TestStreamingDefaultGraphV2Metadata_Unit`, `TestStreamingGraphRetainedStateFollowsAggregateCardinality_Unit`; +3 other source references | Go test entrypoints | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/platform/postgres`; 7 standard-library imports | 5 named Test entrypoint(s) in this file; assertions inspected | C06; no generated file authored here | Network Flow owner tests | medium | E02-034; D223 (declarations/imports and selected assertions); not executed |
| [graph_telemetry.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_telemetry.go) | Safe phase/result/cleanup observations and failure vocabulary | GraphTelemetryObserver; GraphPhaseTelemetryObservation; GraphResultTelemetryObservation; GraphCleanupTelemetryObservation | app/server/network_flow_telemetry.go; source composer/cleanup | `internal/modules/graphprojection`; 3 standard-library imports | source references: `graph_telemetry_test.go` | C16; no generated file authored here | Network Flow contribution / telemetry adapter | medium | E02-035; D205 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_telemetry_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_telemetry_test.go) | Observer panic containment, closed outcomes and semantic observation boundaries | test-only `TestNetworkFlowGraphTelemetryBoundaryContainsObserverFailure_Unit`, `TestNetworkFlowGraphTelemetryOutcomeVocabulary_Unit`; +1 other source references | Go test entrypoints | `internal/modules/graphprojection`; 4 standard-library imports | 3 named Test entrypoint(s) in this file; assertions inspected | C16; no generated file authored here | Network Flow owner tests | medium | E02-036; D194 (declarations/imports and selected assertions); not executed |
| [graph_temporal.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_temporal.go) | Time buckets, conservation, canonical timestamp/selector and digest semantics | private temporal composition helpers | same-package reference candidates: `graph.go`, `graph_response_v2.go`, `graph_source_composer.go`; +3 other source references | `github.com/google/uuid`; 4 standard-library imports | source references: `graph_capacity_test.go`, `graph_temporal_test.go` | C06; no generated file authored here | Network Flow | medium | E02-037; D229 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_temporal_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_temporal_test.go) | Bucket arithmetic/property seeds, conservation, projection identity and digest fixtures | test-only `TestTimeBucketArithmeticFixturesAndLimit_Unit`, `TestTimeBucketAdmissionAndCanonicalTimestampErrors_Unit`; +4 other source references | Go test entrypoints and FuzzTimeBucketArithmetic | 5 standard-library imports | 6 named Test entrypoint(s) in this file; assertions inspected | C06; no generated file authored here | Network Flow owner tests | medium | E02-038; D364 (declarations/imports and selected assertions); not executed |
| [graph_view_authority_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_authority_integration_test.go) | Submitter role revalidation and notification-after-commit probes | test-only `TestSavedGraphMaterializationRevalidatesSubmitterRole_Integration` | Go test entrypoint / saved-graph integration | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgxpool`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/platform/authn`, `internal/platform/jobs`, `internal/platform/postgres`, `internal/testutil/appsupport`, `internal/testutil/collaborationsupport`, `internal/testutil/httptestx`; 7 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C09, C10; no generated file authored here | Network Flow owner tests | medium | E02-039; D269 (declarations/imports and selected assertions); not executed |
| [graph_view_authority_test_bridge_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_authority_test_bridge_test.go) | Test-only worker handler and saved-create command construction | test-only `NewGraphViewAuthorityTestHandler`, `NewSavedGraphCreateCommandForTest` | graph_view_authority_integration_test.go | `internal/modules/incidents/admission`, `internal/platform/authn`, `github.com/google/uuid`, `internal/platform/jobs`; 3 standard-library imports | helper/bridge; graph_view_authority_integration_test.go | C09, C10; no generated file authored here | Network Flow owner tests | medium | E02-040; D35 (declarations/imports and selected assertions); not executed |
| [graph_view_jobs.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_jobs.go) | Materialization execution, authority recheck, finalization and selected bindings | GraphViewMaterializationJobKind; GraphViewJobTransactions/Manager/Runner/Mutation/SuccessFinalization/FailureFinalization/Finalizer | Module.RegisterGraphViewWorker; app/extensionassembly/network_flow_jobs.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/incidents/admission`, `internal/platform/jobs`; 4 standard-library imports | source references: `graph_materialization_payload_test.go`, `graph_materialization_timeout_test.go`, `graph_view_receipts_test.go` | C09, C10; no generated file authored here | Network Flow worker / Jobs lifecycle | high | E02-041; D360 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_view_receipts.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_receipts.go) | Final idempotency reconciliation and saved declaration receipt integrity | GraphViewReceiptReconciler | same-package reference candidates: `graph_view_routes.go`, `indicator_link.go`, `routes.go`; +5 other source references | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/jobs`; 7 standard-library imports | source references: `graph_view_receipts_test.go` | C10, C12; no generated file authored here | Network Flow receipt contribution | high | E02-042; D195 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_view_receipts_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_receipts_test.go) | Name-byte constraints and historical receipt integrity/comparison | test-only `TestGraphViewNameByteContract_Unit`, `TestGraphViewReceiptIntegrityAndComparison_Unit` | Go test entrypoints | `github.com/google/uuid`, `internal/platform/jobs`; 6 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C10; no generated file authored here | Network Flow owner tests | medium | E02-043; D122 (declarations/imports and selected assertions); not executed |
| [graph_view_routes.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_routes.go) | Saved-graph HTTP commands, reads, refresh and contributors | private saved-graph route handlers/response projection | same-package reference candidates: `reporting_graph_source.go`, `routes.go`, `saved_graph_application.go`; +2 other source references | `internal/modules/incidents/admission`, `github.com/google/uuid`, `internal/modules/graphprojection`, `internal/platform/authn`, `internal/platform/httpapi`; 4 standard-library imports | source references: `graph_view_receipts_test.go` | C02, C10; no generated file authored here | Network Flow transport adapter | medium | E02-044; D339 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_view_store.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_store.go) | Saved declarations, generations, selected binding publication and invalidation | private store declaration methods | same-package reference candidates: `application.go`, `extension_state.go`, `graph_restore_source.go`; +7 other source references | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgtype`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`; 11 standard-library imports | source references: `extension_state_v3_integration_test.go`, `graph_result_cleanup_integration_test.go`, `graph_result_cleanup_race_integration_test.go`; +6 other source references | C10, C11; no generated file authored here | Network Flow persistence adapter | high | E02-045; D623 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [graph_view_store_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/graph_view_store_test.go) | Declaration persistence and generation-checked publication with rename | test-only `TestNetworkFlowGraphViewDeclarationPersistence_Integration`, `TestNetworkFlowGraphViewPublicationRejectsStaleGenerationAndAllowsRename_Integration` | Go test entrypoints | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection/postgresresult`, `internal/modules/networkflow`, `internal/testutil/revisionsupport`; 5 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C10; no generated file authored here | Network Flow owner tests | medium | E02-046; D227 (declarations/imports and selected assertions); not executed |
| [harness_audit_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harness_audit_integration_test.go) | Real committed audit consumers and exact count checks | test-only `assertNetworkFlowCommittedAuditConsumers` | network flow integration wrappers | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow/harnesscontrol`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/postgres`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `internal/testutil/networkflowsupport`, `internal/testutil/pgtest`, `github.com/google/uuid`; 8 standard-library imports | helper/bridge; network flow integration wrappers | C09, C14; no generated file authored here | Network Flow owner tests | medium | E02-047; D226 (declarations/imports and selected assertions); not executed |
| [harness_authorization_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harness_authorization_integration_test.go) | Real authority-transition consumers and resource/cursor rechecks | test-only `assertNetworkFlowAuthorizationConsumers` | network flow integration wrappers | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/modules/networkflow/harnesscontrol`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `internal/testutil/networkflowsupport`, `github.com/google/uuid`; 4 standard-library imports | helper/bridge; network flow integration wrappers | C09, C14; no generated file authored here | Network Flow owner tests | medium | E02-048; D159 (declarations/imports and selected assertions); not executed |
| [harness_randomness_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harness_randomness_integration_test.go) | Table entropy collision/retry transaction consumers | test-only `assertNetworkFlowTableEntropyTransactions` | network flow integration wrappers | `internal/modules/networkflow`, `internal/modules/networkflow/harnesscontrol`, `internal/platform/httpapi`, `internal/testutil/httptestx`, `internal/testutil/revisionsupport`, `github.com/google/uuid`, `github.com/jackc/pgx/v5/pgconn`; 7 standard-library imports | helper/bridge; network flow integration wrappers | C03, C14; no generated file authored here | Network Flow owner tests | medium | E02-049; D97 (declarations/imports and selected assertions); not executed |
| [harness_worker_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harness_worker_integration_test.go) | Real worker and transaction fault consumers | test-only `assertNetworkFlowWorkerFaultConsumers` | network flow integration wrappers | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/modules/networkflow/harnesscontrol`, `internal/platform/authn`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `github.com/google/uuid`; 5 standard-library imports | helper/bridge; network flow integration wrappers | C10, C14; no generated file authored here | Network Flow owner tests | medium | E02-050; D99 (declarations/imports and selected assertions); not executed |
| [harnesscontrol/control_json.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/control_json.go) | Bounded strict control JSON object decoding | private decodeControlRequest | same-package reference candidates: `harnesscontrol/network_flow_audit_assertion.go`, `harnesscontrol/network_flow_auth_transition.go`, `harnesscontrol/network_flow_fault.go`; +1 other source references | `internal/platform/strictjson`; 5 standard-library imports | indirect tests in C14; TODO: exact direct call-to-test mapping before editing this file | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-051; D42 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/control_json_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/control_json_test.go) | Malformed/duplicate/trailing/required-field control request matrix | test-only `assertStrictNetworkFlowControlRequests` | control route test wrappers | 4 standard-library imports | helper/bridge; control route test wrappers | C14; no generated file authored here | Testing Harness support tests | medium | E02-052; D33 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/controls.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/controls.go) | Instance registries and guarded harness contribution assembly | Controls; NewControls; Clear; Contribution | app/server/server_profile_harness.go (cartulary_harness build) | `internal/platform/harnessruntime`, `internal/platform/httpapi`; 0 standard-library imports | source references: `harness_audit_integration_test.go`, `harness_authorization_integration_test.go`, `harness_randomness_integration_test.go`; +2 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-053; D56 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/fault_consumers.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/fault_consumers.go) | Instance decorators for transaction/import/worker fault consumption | Controls.WrapDatabase/WrapImportFacade/ConfigureNetworkFlow | harness Controls composition; server NetworkFlowComposition | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/imports`, `internal/modules/networkflow`, `internal/platform/jobs`, `internal/platform/postgres`; 3 standard-library imports | source references: `application_composition_test.go`, `graph_materialization_payload_test.go`, `graph_materialization_timeout_test.go`; +1 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-054; D205 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_audit_assertion.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion.go) | Bound exact committed audit-count assertion registry and admission route | NetworkFlowAuditAssertionRegistry/Assertion/Scope; constructor; RegisterNetworkFlowAuditAssertionRoutes; fixture/consumption methods | same-package reference candidates: `harnesscontrol/controls.go`, `harnesscontrol/network_flow_auth_transition.go`, `harnesscontrol/network_flow_fault.go`; +1 other source references | `github.com/google/uuid`, `internal/platform/httpapi`; 6 standard-library imports | source references: `harness_audit_integration_test.go`, `harness_authorization_integration_test.go`, `harnesscontrol/control_json_test.go`; +4 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-055; D398 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_audit_assertion_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_audit_assertion_integration_test.go) | Disabled/default, harness authorization, count arming, duplicates and clearing | test-only `TestNetworkFlowAuditAssertionRouteDisabledByDefault`, `TestNetworkFlowAuditAssertionRouteRequiresHarnessAuthorization`; +4 other source references | Go test entrypoints; registry evidence only | `internal/platform/httpapi`; 3 standard-library imports | 6 named Test entrypoint(s) in this file; assertions inspected | C14; no generated file authored here | Testing Harness support tests | medium | E02-056; D218 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/network_flow_auth_transition.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_auth_transition.go) | Bound authority-transition registry and admission route | NetworkFlowAuthTransitionRegistry/NetworkFlowAuthTransition; constructor; route/consume/bind methods and closed tokens | same-package reference candidates: `harnesscontrol/controls.go`, `harnesscontrol/network_flow_audit_assertion.go`, `harnesscontrol/network_flow_fault.go`; +1 other source references | `github.com/google/uuid`, `internal/platform/httpapi`; 6 standard-library imports | source references: `harness_audit_integration_test.go`, `harness_authorization_integration_test.go`, `harnesscontrol/control_json_test.go`; +4 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-057; D342 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_auth_transition_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_auth_transition_integration_test.go) | Disabled/default, admission, bound transition keys, invalid requests and clearing | test-only `TestNetworkFlowAuthTransitionRouteDisabledByDefault`, `TestNetworkFlowAuthTransitionRouteRequiresHarnessAuthorization`; +4 other source references | Go test entrypoints; registry evidence only | `internal/platform/httpapi`; 3 standard-library imports | 6 named Test entrypoint(s) in this file; assertions inspected | C14; no generated file authored here | Testing Harness support tests | medium | E02-058; D197 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/network_flow_fault.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_fault.go) | One-shot correlation/boundary fault arming and guarded route | NetworkFlowFaultRegistry/Fault; constructor; RegisterNetworkFlowFaultRoutes; ConsumeNetworkFlowFault/For; boundary/effect tokens | same-package reference candidates: `harnesscontrol/controls.go`, `harnesscontrol/fault_consumers.go`, `harnesscontrol/network_flow_audit_assertion.go`; +2 other source references | `github.com/google/uuid`, `internal/platform/httpapi`; 6 standard-library imports | source references: `harness_worker_integration_test.go`, `harnesscontrol/control_json_test.go`, `harnesscontrol/network_flow_audit_assertion_integration_test.go`; +3 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-059; D297 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_fault_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_fault_integration_test.go) | Disabled/default, admission, one-shot effects, duplicate arming and invalid requests | test-only `TestNetworkFlowFaultRouteDisabledByDefault`, `TestNetworkFlowFaultRouteRequiresHarnessAuthorization`; +4 other source references | Go test entrypoints; registry evidence only | `internal/platform/httpapi`; 3 standard-library imports | 6 named Test entrypoint(s) in this file; assertions inspected | C14; no generated file authored here | Testing Harness support tests | medium | E02-060; D188 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/network_flow_randomness.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_randomness.go) | Table-ID UUID collision stream registry, strict arming and exhaustion | NetworkFlowRandomnessRegistry/NetworkFlowRandomnessState; constructor/route/consume methods; UUID tokens; exhaustion errors | same-package reference candidates: `harnesscontrol/controls.go`, `harnesscontrol/network_flow_audit_assertion.go`, `harnesscontrol/network_flow_auth_transition.go`; +2 other source references | `github.com/google/uuid`, `internal/platform/httpapi`; 5 standard-library imports | source references: `harness_randomness_integration_test.go`, `harnesscontrol/control_json_test.go`, `harnesscontrol/network_flow_audit_assertion_integration_test.go`; +3 other source references | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-061; D289 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/network_flow_randomness_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/network_flow_randomness_integration_test.go) | Disabled/default, admission, deterministic collisions, retired token rejection and clearing | test-only `TestNetworkFlowRandomnessRouteDisabledByDefault`, `TestNetworkFlowRandomnessRouteRequiresHarnessAuthorization`; +6 other source references | Go test entrypoints; registry evidence only | `internal/platform/httpapi`; 5 standard-library imports | 8 named Test entrypoint(s) in this file; assertions inspected | C14; no generated file authored here | Testing Harness support tests | medium | E02-062; D209 (declarations/imports and selected assertions); harness-only composition, F03 |
| [harnesscontrol/randomness_consumers.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/randomness_consumers.go) | Entropy adapter for an instance's armed randomness stream | Controls.TableIDEntropy and private adapter | harness Controls.ConfigureNetworkFlow | 3 standard-library imports | source references: `harness_randomness_integration_test.go` | C14; no generated file authored here | Testing Harness owner-local support | medium | E02-063; D33 (declarations/imports and relevant bodies); harness-only composition, F03 |
| [harnesscontrol/test_helpers_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/harnesscontrol/test_helpers_test.go) | HTTP test requests/body and assertion helpers for guarded control tests | test-only `testRuntimeEnabledEnv`, `testHTTPDependencies`; +7 other source references | harnesscontrol route tests | `internal/platform/httpapi`, `internal/testutil/httpapiextensions`; 5 standard-library imports | helper/bridge; harnesscontrol route tests | C14; no generated file authored here | Testing Harness support tests | medium | E02-064; D102 (declarations/imports and selected assertions); harness-only composition, F03 |
| [import_facade.go](/home/jochi/code/cartulary/internal/modules/networkflow/import_facade.go) | Imports owner binding, preview/result validation and transactional analytical apply | private importFacade implements imports.ExtensionImportFacade; Binding/PrepareImportUnitMapping/ValidateImportUnitMappingResult/ApplyImportUnitTx | application_composition.go; modules/imports/{mapping,application_operations}.go through ExtensionImportFacade; harness decorator | `github.com/jackc/pgx/v5`, `internal/modules/imports`, `internal/platform/httpapi`, `internal/platform/jobs`; 7 standard-library imports | network_flow_contract_test.go (binding/error assertions); network_flow_behavior_test.go (import helper wrappers); preview helper in network_flow_unit_test.go. Gap G01 | C03; no generated file authored here | Network Flow analytical target | high | E02-065; D439 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [import_owner_errors.go](/home/jochi/code/cartulary/internal/modules/networkflow/import_owner_errors.go) | Closed owner error union, validation and Imports translation | private facade TranslateImportUnitError/ValidateImportUnitError | same-package reference candidates: `import_facade.go` | `internal/modules/imports`; 2 standard-library imports | network_flow_contract_test.go: TestImportOwnerErrorTranslationUsesClosedRegisteredUnion | C03; no generated file authored here | Network Flow owner errors / Imports framing | medium | E02-066; D208 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link.go) | Typed explicit linking commands, selectors and owner participant coordination | private indicator-link application/outcomes | same-package reference candidates: `indicator_link_admission.go`, `indicator_link_http.go`, `indicator_link_receipts.go`; +1 other source references | `github.com/google/uuid`, `internal/modules/crossownertransaction`, `internal/modules/incidents/admission`, `internal/modules/indicators`, `internal/platform/authn`; 7 standard-library imports | source references: `indicator_link_boundary_test.go`, `network_flow_contract_test.go` | C07, C09; no generated file authored here | Network Flow linking / Indicators target | medium | E02-067; D470 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link_admission.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_admission.go) | Source row/target admission and canonical confirmation | private indicator-link admission | same-package reference candidates: `binding_store.go`, `indicator_link.go`, `indicator_link_graph_admission.go`; +4 other source references | `github.com/google/uuid`; 6 standard-library imports | source references: `indicator_link_boundary_test.go`, `indicator_link_fixture_test.go` | C07; no generated file authored here | Network Flow source semantics | high | E02-068; D300 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link_boundary_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_boundary_test.go) | Public facade/participant boundary assertion for indicator linking | test-only `TestIndicatorLinkPublicBoundary` | Go test entrypoint | 6 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C07; no generated file authored here | Network Flow owner tests | medium | E02-069; D123 (declarations/imports and selected assertions); not executed |
| [indicator_link_fixture_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_fixture_test.go) | Test-only link store construction and request/result fixtures | test-only `decodeIndicatorLinkFixture` | network_flow_contract_test.go and route tests | `internal/platform/httpapi`; 1 standard-library imports | helper/bridge; network_flow_contract_test.go and route tests | C07; no generated file authored here | Network Flow owner tests | medium | E02-070; D16 (declarations/imports and selected assertions); not executed |
| [indicator_link_graph_admission.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_graph_admission.go) | Graph selector/source admission for indicator linking | private graph-linked source validation | same-package reference candidates: `indicator_link_admission.go` | 1 standard-library imports | source references: `indicator_link_boundary_test.go` | C06, C07; no generated file authored here | Network Flow | high | E02-071; D102 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link_http.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_http.go) | HTTP indicator-link decoding, command invocation and translation | private route handler/HTTP adapter | private package use; no distinct direct caller established in lexical map | `internal/platform/authn`, `internal/platform/httpapi`, `github.com/google/uuid`; 2 standard-library imports | indirect tests in C02, C07; TODO: exact direct call-to-test mapping before editing this file | C02, C07; no generated file authored here | Network Flow transport adapter | medium | E02-072; D49 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [indicator_link_receipts.go](/home/jochi/code/cartulary/internal/modules/networkflow/indicator_link_receipts.go) | Indicator-link idempotent outcome replay and durable receipts | private indicator receipt adapter | same-package reference candidates: `indicator_link_http.go` | `github.com/jackc/pgx/v5`, `internal/modules/indicators`, `internal/platform/authn`, `internal/platform/extensionstore`, `github.com/google/uuid`; 6 standard-library imports | indirect tests in C07, C09; TODO: exact direct call-to-test mapping before editing this file | C07, C09; no generated file authored here | Network Flow | high | E02-073; D191 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [integration_surface_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/integration_surface_test.go) | Test-only aliases/constructors expose private module internals to integration package | test-only `parseKeyRingsWithDefaultRegistry`, `lifecycleStates`; +2 other source references | external-package networkflow integration tests | `internal/platform/authn`, `internal/platform/secretpurpose`, `internal/testutil/fixtures`; 2 standard-library imports | helper/bridge; external-package networkflow integration tests | C01, C15; no generated file authored here | Network Flow owner tests | medium | E02-074; D106 (declarations/imports and selected assertions); not executed |
| [keyring.go](/home/jochi/code/cartulary/internal/modules/networkflow/keyring.go) | Purpose-separated cursor/safe-digest keys, manifest validation and rotation | KeyRings; KeyRingsOverrideKey; KeyRingManifestMaximumSize; ParseKeyRingsWithRegistry | configuration.go; app/configassembly/deployment.go | `internal/platform/secretpurpose`; 12 standard-library imports | source references: `integration_surface_test.go`, `keyring_test.go`, `network_flow_contract_test.go` | C05, C16; no generated file authored here | Network Flow security contribution | high | E02-075; D403 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [keyring_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/keyring_test.go) | Key purpose separation, cursor rotation, inactive-epoch purge and invalid/null manifests | test-only `TestNetworkFlowKeyRingsAndCursorRotation`, `TestNetworkFlowSafeDigestRingPurgesInactiveEpoch`; +1 other source references | Go test entrypoints | `internal/platform/authn`; 6 standard-library imports | 3 named Test entrypoint(s) in this file; assertions inspected | C05, C16; no generated file authored here | Network Flow owner tests | medium | E02-076; D218 (declarations/imports and selected assertions); not executed |
| [mapping.go](/home/jochi/code/cartulary/internal/modules/networkflow/mapping.go) | Closed source/mapping profiles and approved-mapping materialization | private mapping types/decoders and registry projections | same-package reference candidates: `csv_parser.go`, `import_facade.go`, `query_filter.go`; +1 other source references | `internal/gen/networkflowmapping`, `internal/platform/strictjson`; 4 standard-library imports | source references: `network_flow_contract_test.go`, `network_flow_unit_test.go`, `timestamp_test.go` | `gen/networkflowmapping`; C03 | Network Flow | medium | E02-077; D539 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [module.go](/home/jochi/code/cartulary/internal/modules/networkflow/module.go) | Coherent public module construction, active contribution admission and registration | Module; ModuleDependencies; ImportSourcePort; NewModule; RegisterRoutes/ImportOwner/RegisterGraphViewWorker/InstallCrossOwnerCoordinator/TransactionCapabilities | app/server/runtime_assembly.go; app/server/extension_admission.go; app/extensionassembly/cross_owner_transaction.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/gen/networkflowroutes`, `internal/modules/crossownertransaction`, `internal/modules/imports`, `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/postgres`; 6 standard-library imports | application_composition_test.go via production_surface_test.go / query_authoring_test.go | `gen/networkflowroutes`; C01, C02 | Network Flow facade | medium | E02-078; D249 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [names.go](/home/jochi/code/cartulary/internal/modules/networkflow/names.go) | Display-name derivation, Unicode normalization and collision naming | private naming functions | same-package reference candidates: `csv_parser.go`, `graph_view_store.go`, `import_facade.go`; +5 other source references | `internal/gen/networkflowunicode`; 3 standard-library imports | source references: `graph_view_receipts_test.go`, `network_flow_contract_test.go`, `network_flow_unit_test.go`; +2 other source references | `gen/networkflowunicode`; C03, C04 | Network Flow | medium | E02-079; D179 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [network_flow_behavior_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/network_flow_behavior_test.go) | Integration wrappers execute contract helpers for envelopes, imports, cursor invalidation, authorization, redaction and graph/link behavior | test-only `TestNetworkFlow_CoreEnvelopeAdmissionEvidence_Integration`, `TestNetworkFlow_ReservedSourceProfileEvidence_Integration`; +16 other source references | Go test entrypoints | 1 standard-library imports | 18 named Test entrypoint(s) in this file; assertions inspected | C02, C03, C04, C05, C06, C07, C09, C16; no generated file authored here | Network Flow owner tests | medium | E02-080; D78 (declarations/imports and selected assertions); not executed |
| [network_flow_contract_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/network_flow_contract_test.go) | Reusable behavioral assertion bodies plus closed Imports error translation and binding metadata tests | test-only `TestImportOwnerErrorTranslationUsesClosedRegisteredUnion`, `TestNetworkFlowImportFacadePublishesCompleteBinding` | network_flow_behavior_test.go / own entrypoints | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/imports`, `internal/modules/incidents/admission`, `internal/platform/httpapi`; 10 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C02, C03, C04, C05, C06, C07, C09, C16; no generated file authored here | Network Flow owner tests | medium | E02-081; D720 (declarations/imports and selected assertions); not executed |
| [network_flow_unit_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/network_flow_unit_test.go) | Parser/mapping/resource/identity fixtures and reusable unit assertion helpers | test-only `TestNetworkFlow_CiscoSNARequiredFields_Unit`, `TestNetworkFlow_CSVParserEdgeOutcomes_Unit`; +12 other source references | Go test entrypoints / contract helpers | `github.com/google/uuid`, `internal/modules/graphprojection`; 14 standard-library imports | 14 named Test entrypoint(s) in this file; assertions inspected | C03, C05, C06, C15; no generated file authored here | Network Flow owner tests | medium | E02-082; D813 (declarations/imports and selected assertions); not executed |
| [operation_admission.go](/home/jochi/code/cartulary/internal/modules/networkflow/operation_admission.go) | Operation-policy-based current incident authorization | private operationAdmission; admitOperation | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go`, `indicator_link_http.go`; +1 other source references | `internal/modules/incidents/admission`, `internal/platform/httpapi`, `internal/platform/httpauth`, `github.com/google/uuid`; 3 standard-library imports | indirect tests in C02, C09; TODO: exact direct call-to-test mapping before editing this file | C02, C09; no generated file authored here | Network Flow operations / Incidents admission authority | high | E02-083; D72 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [operation_admission_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/operation_admission_integration_test.go) | Real major-7 route admission precedence matrix | test-only `assertMajor7AdmissionMatrix` | table_lifecycle_integration_test.go and route integration wrappers | `internal/gen/networkflowroutes`, `internal/platform/authn`, `internal/platform/postgres`, `internal/testutil/httptestx`, `github.com/google/uuid`; 4 standard-library imports | helper/bridge; table_lifecycle_integration_test.go and route integration wrappers | `gen/networkflowroutes`; C02, C09 | Network Flow owner tests | medium | E02-084; D78 (declarations/imports and selected assertions); not executed |
| [portability_state.go](/home/jochi/code/cartulary/internal/modules/networkflow/portability_state.go) | Retained owner-state presence for incident portability | PortabilityStateQuery; PortabilityStateBinding; NewPortabilityStateBinding | app/extensionassembly/incident_portability.go | `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 3 standard-library imports | source references: `configuration_test.go` | C12; no generated file authored here | Network Flow source contribution / Incident Portability | medium | E02-085; D56 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [ports.go](/home/jochi/code/cartulary/internal/modules/networkflow/ports.go) | Consumer-owned incident lock, audit and indicator participant ports | IncidentLockPort; AdministrativeAuditPort; IndicatorParticipationPort | module.go; store.go; app/server assembly | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/indicators`; 1 standard-library imports | indirect tests in C01, C07, C09; TODO: exact direct call-to-test mapping before editing this file | C01, C07, C09; no generated file authored here | Network Flow narrow ports | medium | E02-086; D24 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [production_surface_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/production_surface_test.go) | AST positive public API and private implementation surface guard | test-only `assertProductionSurface` | query_authoring_test.go | 6 standard-library imports | helper/bridge; query_authoring_test.go | C01; no generated file authored here | Network Flow owner tests | medium | E02-087; D154 (declarations/imports and selected assertions); not executed |
| [query.go](/home/jochi/code/cartulary/internal/modules/networkflow/query.go) | Typed accepted/rejected query decoding, sort/filter/cursor bindings and normalization | private query request/cursor/scope types | same-package reference candidates: `graph.go`, `graph_query_application.go`, `graph_source_composer.go`; +13 other source references | `internal/platform/strictjson`; 13 standard-library imports | source references: `indicator_link_fixture_test.go`, `network_flow_contract_test.go`, `network_flow_unit_test.go`; +2 other source references | C05; no generated file authored here | Network Flow | medium | E02-088; D922 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [query_authoring_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/query_authoring_test.go) | Direct read authority, public/semantic guards and filter normalization/error detail cases | test-only `TestQueryAuthoringAdmission_Unit` | Go test entrypoint | `internal/modules/incidents/admission`, `github.com/google/uuid`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C01, C05, C09; no generated file authored here | Network Flow owner tests | medium | E02-089; D78 (declarations/imports and selected assertions); not executed |
| [query_filter.go](/home/jochi/code/cartulary/internal/modules/networkflow/query_filter.go) | Closed filter grammar, scalar/range normalization and diagnostic tokens | private filter admission helpers | same-package reference candidates: `query.go` | `internal/gen/contractnetworkflow`; 7 standard-library imports | source references: `query_authoring_test.go` | `gen/contractnetworkflow`; C05 | Network Flow | medium | E02-090; D310 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [query_sql.go](/home/jochi/code/cartulary/internal/modules/networkflow/query_sql.go) | Parameterized source-owned row/diagnostic SQL with keyset pagination | private store QueryRowsPage/QueryRejectedDiagnosticsPage and SQL helpers | same-package reference candidates: `store.go`, `table_query_application.go` | `github.com/google/uuid`; 5 standard-library imports | source references: `network_flow_contract_test.go`, `query_sql_test_bridge_test.go` | C05; no generated file authored here | Network Flow persistence adapter | medium | E02-091; D409 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [query_sql_test_bridge_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/query_sql_test_bridge_test.go) | Test-only live SQL keyset boundary assertions | test-only `AssertLiveSQLKeysetBoundaries` | network_flow_contract_test.go | `github.com/google/uuid`; 4 standard-library imports | helper/bridge; network_flow_contract_test.go | C05; no generated file authored here | Network Flow owner tests | medium | E02-092; D50 (declarations/imports and selected assertions); not executed |
| [recovery_bindings.go](/home/jochi/code/cartulary/internal/modules/networkflow/recovery_bindings.go) | Owner authoritative-table inventory contribution | RecoveryPostgresTables | app/extensionassembly/recovery.go | 1 standard-library imports | indirect tests in C12; TODO: exact direct call-to-test mapping before editing this file | C12; no generated file authored here | Network Flow / Recovery registration | medium | E02-093; D24 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [recovery_state.go](/home/jochi/code/cartulary/internal/modules/networkflow/recovery_state.go) | Owner recovery family contribution | RecoveryStateContribution | app/recoveryassembly/state_catalog.go | `internal/platform/recoverystate`; 0 standard-library imports | indirect tests in C12; TODO: exact direct call-to-test mapping before editing this file | C12; no generated file authored here | Network Flow / Recovery registration | medium | E02-094; D13 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [reporting_graph_source.go](/home/jochi/code/cartulary/internal/modules/networkflow/reporting_graph_source.go) | Exact selected graph leasing, renewal/read, label conversion and release | ReportingGraphSource; Module.ReportingGraphSource; SourceOwnerID/ValidateAndLeaseResultTx/ReadAndRenewLeasedResultTx/ReleaseJobLeasesTx/ReleaseJobLeases | Module construction; app/server runtime composition; Reporting source interface | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/reporting/graphsourcecontract`, `internal/platform/postgres`; 6 standard-library imports | source references: `application_composition_test.go`, `graph_result_cleanup_race_integration_test.go`, `reporting_graph_source_integration_test.go` | C11, C12; no generated file authored here | Network Flow source / Reporting consumer / Graph Projection leases | medium | E02-095; D304 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [reporting_graph_source_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/reporting_graph_source_integration_test.go) | Exact lease/read/release lifecycle, scope and result stability | test-only `TestReportingGraphSourceValidatesLeasesReadsAndReleasesExactResult_Integration` | Go test entrypoint | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/graphprojection`, `internal/modules/graphprojection/postgresresult`, `internal/modules/networkflow`, `internal/modules/reporting/graphsourcecontract`, `internal/testutil/revisionsupport`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C11, C12; no generated file authored here | Network Flow owner tests | medium | E02-096; D229 (declarations/imports and selected assertions); not executed |
| [request_decoding.go](/home/jochi/code/cartulary/internal/modules/networkflow/request_decoding.go) | Transport-independent typed table and saved-graph request admission | private decode...RequestValue functions | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go`, `routes.go` | 1 standard-library imports | source references: `network_flow_contract_test.go` | C02, C04, C06, C10; no generated file authored here | Network Flow semantics | medium | E02-097; D303 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [resources.go](/home/jochi/code/cartulary/internal/modules/networkflow/resources.go) | Owner table/row/diagnostic/binding/profile resource projection | private resource functions | same-package reference candidates: `graph.go`, `graph_query_presentation.go`, `graph_response_v2.go`; +8 other source references | 2 standard-library imports | source references: `configuration_test.go`, `graph_streaming_test.go`, `graph_temporal_test.go`; +2 other source references | C02, C03, C05, C07; no generated file authored here | Network Flow | medium | E02-098; D218 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [routes.go](/home/jochi/code/cartulary/internal/modules/networkflow/routes.go) | Table/profile HTTP adapters, session framing and safe response errors | private routeService; no public service constructor | same-package reference candidates: `application.go`, `graph_routes.go`, `graph_view_receipts.go`; +9 other source references | `github.com/google/uuid`, `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/httpauth`; 6 standard-library imports | source references: `network_flow_contract_test.go`, `table_lifecycle_test.go` | C02, C04, C05, C09; no generated file authored here | Network Flow transport adapter | medium | E02-099; D373 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [routes_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/routes_integration_test.go) | Claimed/unclaimed routes, query pagination, graph/indicator links, saved lifecycle and recovery composition | test-only `TestNetworkFlowPaginationRecovery_Integration`, `TestNetworkFlowRoutesRemainUnclaimedByDefault`; +6 other source references | Go test entrypoints | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/app/recoveryassembly`, `internal/app/server`, `internal/modules/auth/testsupport/flowtest`, `internal/modules/collaboration/protocol`, `internal/modules/graphprojection`, `internal/modules/graphprojection/restore`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/modules/networkflow/harnesscontrol`, `internal/modules/recovery/restorecontract`, `internal/modules/reporting`, `internal/platform/authn`, `internal/platform/httpapi`, `internal/platform/jobs`, `internal/platform/postgres`, `internal/platform/renderprocess`, `internal/testutil/appsupport`, `internal/testutil/authcookietest`, `internal/testutil/collaborationsupport`, `internal/testutil/collaborationsupport/incidentwstest`, `internal/testutil/fixtures`, `internal/testutil/httptestx`, `internal/testutil/pgtest`, `internal/testutil/reportingfixture`; 12 standard-library imports | 8 named Test entrypoint(s) in this file; assertions inspected | C01, C02, C04, C05, C06, C07, C08, C10, C12; no generated file authored here | Network Flow owner tests | medium | E02-100; D2232 (declarations/imports and selected assertions); not executed |
| [saved_graph_admission.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_admission.go) | Readiness/cutover validation of declarations, jobs, receipts and selected proofs | SavedGraphCutoverAlgorithmID; ValidateSavedGraphAdmission; ValidateRetainedExtensionState | app/configassembly/extension_admission.go; app/server/extension_admission.go | `internal/modules/graphprojection/postgresresult`, `internal/platform/authn`, `internal/platform/extensionstore`, `internal/platform/jobs`, `github.com/google/uuid`; 7 standard-library imports | source references: `extension_state_v3_integration_test.go`, `graph_materialization_payload_test.go`, `routes_integration_test.go`; +1 other source references | C10, C12; no generated file authored here | Network Flow retained-state admission | high | E02-101; D355 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [saved_graph_admission_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_admission_integration_test.go) | Retained state cutover admission and unchanged invalid bytes | test-only `TestSavedGraphCutoverAdmission_Integration` | Go test entrypoint | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/platform/authn`, `internal/platform/extensionstore`, `internal/platform/jobs`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgxpool`; 8 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C10, C12; no generated file authored here | Network Flow owner tests | medium | E02-102; D192 (declarations/imports and selected assertions); not executed |
| [saved_graph_application.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_application.go) | Typed create/refresh/rename/retire commands, capacity, jobs, audit and after-commit notification | private savedGraphApplication/outcomes | same-package reference candidates: `graph_view_routes.go` | `internal/modules/incidents/admission`, `internal/platform/authn`, `internal/platform/jobs`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 5 standard-library imports | source references: `graph_view_authority_test_bridge_test.go` | C09, C10; no generated file authored here | Network Flow | high | E02-103; D359 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [saved_graph_read_application.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_read_application.go) | Saved declaration/result/contributor reads and unavailable-result classification | private savedGraphReadApplication; reader ports | same-package reference candidates: `graph_query_application.go`, `graph_routes.go`, `graph_view_routes.go` | `internal/modules/graphprojection`, `internal/platform/jobs`, `github.com/google/uuid`; 4 standard-library imports | source references: `saved_graph_read_application_test.go` | C09, C10, C11; no generated file authored here | Network Flow | high | E02-104; D252 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [saved_graph_read_application_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_read_application_test.go) | Unavailable result classification distinguishes missing Jobs history from operational failure | test-only `TestSavedGraphReadFailures_Unit` | Go test entrypoint | `internal/platform/jobs`, `github.com/google/uuid`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C09, C10; no generated file authored here | Network Flow owner tests | medium | E02-105; D101 (declarations/imports and selected assertions); not executed |
| [saved_graph_receipts.go](/home/jochi/code/cartulary/internal/modules/networkflow/saved_graph_receipts.go) | Saved command receipt adapter and immutable outcome serialization | private savedGraphReceiptAdapter/failure | same-package reference candidates: `graph_view_routes.go` | `internal/platform/authn`, `github.com/jackc/pgx/v5`; 5 standard-library imports | indirect tests in C10; TODO: exact direct call-to-test mapping before editing this file | C10; no generated file authored here | Network Flow | high | E02-106; D90 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [security.go](/home/jochi/code/cartulary/internal/modules/networkflow/security.go) | Authenticated opaque cursor encoding/binding/expiry and retired-key purge | private cursor codec/protector/payload | same-package reference candidates: `module.go` | `internal/platform/cryptography`; 9 standard-library imports | source references: `keyring_test.go`, `network_flow_contract_test.go` | C05, C09, C16; no generated file authored here | Network Flow cursor semantics | high | E02-107; D233 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [semantic_boundary_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_boundary_test.go) | Semantic import allowlist/root guard and safe failure assertions | test-only `assertSemanticBoundary` | query_authoring_test.go | 8 standard-library imports | helper/bridge; query_authoring_test.go | C01, C02, C03; no generated file authored here | Network Flow owner tests | medium | E02-108; D87 (declarations/imports and selected assertions); not executed |
| [semantic_failure.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_failure.go) | Closed private failures, safe typed details and wrapped causes | private semanticFailure/failureKind/failureDetails | same-package reference candidates: `graph.go`, `graph_query_application.go`, `graph_response_v2.go`; +11 other source references | `internal/modules/incidents/admission`; 2 standard-library imports | semantic_boundary_test.go via query_authoring_test.go | C02, C09; no generated file authored here | Network Flow | medium | E02-109; D106 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [semantic_http_adapter.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_http_adapter.go) | HTTP translation for semantic request admission and graph-stale failure | private networkFlowAPIError and HTTP decoder wrappers | same-package reference candidates: `graph_view_routes.go`, `operation_admission.go`, `routes.go` | `internal/platform/httpapi`; 1 standard-library imports | indirect tests in C02; TODO: exact direct call-to-test mapping before editing this file | C02; no generated file authored here | Network Flow transport adapter | medium | E02-110; D29 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [semantic_http_errors.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_http_errors.go) | Closed semantic failure to registered HTTP error mapping | private semanticHTTPError | same-package reference candidates: `graph_routes.go`, `graph_view_routes.go`, `indicator_link_http.go`; +3 other source references | `internal/platform/httpapi`; 2 standard-library imports | source references: `indicator_link_boundary_test.go`, `indicator_link_fixture_test.go`, `network_flow_unit_test.go`; +2 other source references | C02; no generated file authored here | Network Flow transport adapter | medium | E02-111; D138 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [semantic_reference_boundary_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/semantic_reference_boundary_test.go) | Transitive AST semantic reference guard, including aliases and reachable declarations | test-only `semanticReferenceViolations`, `assertSemanticReferences` | semantic_boundary_test.go | 9 standard-library imports | helper/bridge; semantic_boundary_test.go | C01, C02, C03; no generated file authored here | Network Flow owner tests | medium | E02-112; D174 (declarations/imports and selected assertions); not executed |
| [store.go](/home/jochi/code/cartulary/internal/modules/networkflow/store.go) | Owner table/row/diagnostic SQL, immutable inserts, lifecycle, limits and atomic audit/intents | private store and records; no exported store constructor | same-package reference candidates: `application.go`, `binding_store.go`, `graph_query_application.go`; +9 other source references | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `github.com/jackc/pgx/v5/pgtype`, `internal/platform/postgres`; 8 standard-library imports | source references: `graph_streaming_test.go`, `network_flow_contract_test.go`, `routes_integration_test.go`; +2 other source references | C03, C04, C05, C08, C09; no generated file authored here | Network Flow persistence adapter | high | E02-113; D1125 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [store_fixture_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/store_fixture_test.go) | Test-only transaction wrappers around private store operations | test-only `CreateTable`, `RenameTable`; +2 other source references | store/lifecycle/route/worker integration tests | `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 1 standard-library imports | helper/bridge; store/lifecycle/route/worker integration tests | C03, C04; no generated file authored here | Network Flow owner tests | medium | E02-114; D51 (declarations/imports and selected assertions); not executed |
| [store_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/store_test.go) | Real persistence, immutable rows, active/retained counts, naming and intent rollback assertions | test-only `TestNetworkFlowResourceIntentFailureRollsBackSourceMutation_Integration`, `TestNetworkFlow_TableLifecycleVocabularyHasNoRenamedState_Unit`; +1 other source references | Go test entrypoints / helper assertions | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/app/indicatorassembly`, `internal/app/projectionassembly`, `internal/modules/auth/testsupport/storetest`, `internal/modules/incidents`, `internal/modules/incidents/admission`, `internal/modules/incidents/testsupport/storetest`, `internal/modules/indicators`, `internal/modules/networkflow`, `internal/modules/records`, `internal/modules/reference_data/testsupport`, `internal/modules/revisions`, `internal/platform/authn`, `internal/platform/postgres`, `internal/testutil/collaborationsupport`, `internal/testutil/networkflowsupport`, `internal/testutil/revisionsupport`; 6 standard-library imports | 3 named Test entrypoint(s) in this file; assertions inspected | C03, C04, C08, C09; no generated file authored here | Network Flow owner tests | medium | E02-115; D584 (declarations/imports and selected assertions); not executed |
| [table_http.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_http.go) | Table typed outcome to HTTP and route command adapter | private tableOutcomeHTTP and commit route wrappers | same-package reference candidates: `routes.go` | `internal/modules/incidents/admission`, `internal/platform/httpapi`, `github.com/google/uuid`; 2 standard-library imports | indirect tests in C02, C04; TODO: exact direct call-to-test mapping before editing this file | C02, C04; no generated file authored here | Network Flow transport adapter | medium | E02-116; D29 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [table_lifecycle_integration_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_lifecycle_integration_test.go) | Current admission before replay, serialized lifecycle and historical response invariants | test-only `TestTableLifecycleAdmissionAndReplay_Integration` | Go test entrypoint | `internal/modules/auth/testsupport/flowtest`, `internal/modules/incidents/testsupport/scenariotest`, `internal/modules/networkflow`, `internal/platform/authn`, `internal/testutil/appsupport`, `internal/testutil/httptestx`, `github.com/google/uuid`, `github.com/jackc/pgx/v5`; 8 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C04, C09; no generated file authored here | Network Flow owner tests | medium | E02-117; D191 (declarations/imports and selected assertions); not executed |
| [table_lifecycle_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_lifecycle_test.go) | Name control precedence and mutation digest assertions | test-only `TestTableLifecycleNameControlPrecedence_Unit`, `TestTableLifecycleMutationDigest_Unit` | Go test entrypoints | 4 standard-library imports | 2 named Test entrypoint(s) in this file; assertions inspected | C04; no generated file authored here | Network Flow owner tests | medium | E02-118; D36 (declarations/imports and selected assertions); not executed |
| [table_query_application.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_query_application.go) | Read admission, active table scope, rows/diagnostics and cursor continuation | private tableQueryApplication and typed outcomes | same-package reference candidates: `graph_query_application.go`, `graph_source_composer.go`, `graph_view_jobs.go`; +5 other source references | `internal/modules/incidents/admission`, `github.com/google/uuid`; 5 standard-library imports | source references: `network_flow_contract_test.go`, `query_authoring_test.go` | C05, C09; no generated file authored here | Network Flow | high | E02-119; D425 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [table_query_presentation.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_query_presentation.go) | Accepted row and rejected diagnostic response projection | private acceptedRowsResource/diagnosticsResource | same-package reference candidates: `routes.go` | 0 standard-library imports | indirect tests in C05; TODO: exact direct call-to-test mapping before editing this file | C05; no generated file authored here | Network Flow | medium | E02-120; D32 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [table_receipts.go](/home/jochi/code/cartulary/internal/modules/networkflow/table_receipts.go) | Table command replay/save and stable response serialization | private tableReceiptAdapter/storedMutationReceipt | same-package reference candidates: `table_http.go` | `internal/platform/authn`, `github.com/jackc/pgx/v5`; 3 standard-library imports | indirect tests in C04; TODO: exact direct call-to-test mapping before editing this file | C04; no generated file authored here | Network Flow | high | E02-121; D43 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [textcanon_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/textcanon_test.go) | Pinned Unicode canonicalization assertions | test-only `TestUnicode17TextCanonicalization` | Go test entrypoint | `internal/gen/networkflowunicode`; 1 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | `gen/networkflowunicode`; C03, C04 | Network Flow owner tests | medium | E02-122; D38 (declarations/imports and selected assertions); not executed |
| [timestamp.go](/home/jochi/code/cartulary/internal/modules/networkflow/timestamp.go) | Exact timestamp variants, epoch/RFC3339/uptime parsing and DST ambiguity checks | private timestamp profile/error/decoder helpers | same-package reference candidates: `csv_parser.go`, `mapping.go`, `query_filter.go` | 6 standard-library imports | source references: `network_flow_unit_test.go`, `timestamp_test.go` | C03; no generated file authored here | Network Flow | medium | E02-123; D314 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [timestamp_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/timestamp_test.go) | Closed timestamp variants, precise grammar, timezone transitions and uptime ordinals | test-only `TestTimestampProfileExactGrammarPrecisionAndZoneTransitions`, `TestTimestampProfileClosedJSONVariantsRejectNullMissingAndCrossVariantMembers`; +1 other source references | Go test entrypoints | 4 standard-library imports | 3 named Test entrypoint(s) in this file; assertions inspected | C03; no generated file authored here | Network Flow owner tests | medium | E02-124; D122 (declarations/imports and selected assertions); not executed |
| [timezone.go](/home/jochi/code/cartulary/internal/modules/networkflow/timezone.go) | Pinned timezone lookup | private loadPinnedTimezone | same-package reference candidates: `timestamp.go` | `internal/gen/networkflowtz`; 6 standard-library imports | indirect tests in C03; TODO: exact direct call-to-test mapping before editing this file | `gen/networkflowtz`; C03 | Network Flow using generated timezone data | medium | E02-125; D51 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [transaction.go](/home/jochi/code/cartulary/internal/modules/networkflow/transaction.go) | Owner-local transaction begin/rollback/commit lifecycle | private withinTransaction | same-package reference candidates: `application.go`, `graph_query_application.go`, `graph_view_jobs.go`; +1 other source references | `github.com/jackc/pgx/v5`, `internal/platform/postgres`; 2 standard-library imports | transaction_test.go | C04, C09, C10; no generated file authored here | Network Flow persistence adapter | high | E02-126; D25 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [transaction_participants.go](/home/jochi/code/cartulary/internal/modules/networkflow/transaction_participants.go) | Import apply and explicit indicator-link participants in cross-owner transaction | ImportApplyParticipantID; IndicatorLinkParticipantID; private transactionCapability/participant | same-package reference candidates: `import_facade.go`, `indicator_link.go` | `github.com/google/uuid`, `github.com/jackc/pgx/v5`, `internal/modules/crossownertransaction`, `internal/modules/imports`, `internal/modules/incidents/admission`, `internal/modules/indicators`, `internal/platform/authn`; 6 standard-library imports | indirect tests in C03, C07, C09; TODO: exact direct call-to-test mapping before editing this file | C03, C07, C09; no generated file authored here | Network Flow participant / crossownertransaction coordinator | high | E02-127; D299 (declarations/imports and relevant bodies); source inspection, not conformance proof |
| [transaction_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/transaction_test.go) | Begin/commit/rollback lifecycle probe cases | test-only `TestTransactionLifecycle` | Go test entrypoint | `github.com/jackc/pgx/v5`, `internal/platform/postgres`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | C09; no generated file authored here | Network Flow owner tests | medium | E02-128; D123 (declarations/imports and selected assertions); not executed |
| [v3_contract_projection_test.go](/home/jochi/code/cartulary/internal/modules/networkflow/v3_contract_projection_test.go) | Current major-7 generated graph contract projection assertions despite historical filename | test-only `TestNetworkFlowV7GraphContractProjection_Unit` | Go test entrypoint | `internal/gen/contractgraphprojection`, `internal/gen/contractnetworkflow`, `internal/gen/networkflowroutes`; 4 standard-library imports | 1 named Test entrypoint(s) in this file; assertions inspected | `gen/contractgraphprojection`, `gen/contractnetworkflow`, `gen/networkflowroutes`; C13 | Network Flow owner tests | medium | E02-129; D223 (declarations/imports and selected assertions); not executed |

## 3. Module Boundary Diagnosis

DOC-03 current supplement: [§26.3 Boundary diagnosis](#263-boundary-diagnosis). The material below remains historical DOC-01 evidence.

The target is a **legitimate source-owner module with a small application/service facade**, private query/command orchestration, transport adapters and persistence adapters. It is a mutation coordinator and a flow-specific graph orchestration layer. Its breadth follows the adopted Network Flow scope; the directory's existence alone did not establish that conclusion. It is not a frontend controller or grid-vendor layer. No evidence supports replacing the existing facade or moving flow semantics wholesale into generic projections, Imports, Indicators, or Core saved views.

The main exception is a narrow transport dependency in import preview validation (F01). The harness subpackage adds a distinct testing responsibility, but its runtime isolation and security-scan accounting make relocation a deferred architectural question (F03), not an implementation decision.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Coherent facade and retained applications | `module.go`, `application_composition.go`, private application files | Network Flow | keep | E01-NF §29.2; E02; E03 | Active bundle validation, one DB/clock/port set, single-assignment coordinator and error-returning ImportOwner already exist. No new facade required. |
| Strict preview-result object admission through HTTP helper | `import_facade.go:validatePreviewDiagnosticShape` | NF owns semantic admission; platform/strictjson owns syntax decoding | split | E01-NF §29; E04; F01 | Remove only HTTP-helper dependency in a later slice; do not move owner validation into generic Imports or change rejection behavior. |
| CSV, mappings, immutable rows and rejected diagnostics | `csv_parser.go`, `mapping.go`, `digest.go`, `timestamp.go`, `store.go` | Network Flow analytical import target | keep | NF-REQ-001/002; Imports facade call sites | Imports owns upload/session/unit workflow; NF owns analytical target semantics. |
| Table/read command coordination and local SQL | `application.go`, `table_query_application.go`, `query_sql.go`, `store.go`, `transaction.go` | Network Flow, behind existing facade/private adapters | keep | NF-REQ-210; internal/AGENTS; E02 | SQL/pgx presence is not itself a boundary defect. Preserve transaction ownership and source-owned relations. |
| Flow graph semantics and source composition | `graph*.go`, `graph_source_composer.go` | Network Flow; generic projection engine remains Graph Projection | keep | E01-NF §§14,26,29; native V2 adapter; E05 | Shared route/worker/restore composer is already present. No generic projection engine migration into NF. |
| Saved graph declarations, lifecycle and refresh | `saved_graph_*.go`, `graph_view_*.go` | Network Flow; Jobs/Graph own their substrates | keep | E01-NF §§19.1,28,29; E02/E05 | These are not Core saved views. Versions, generations and selected bindings have distinct meanings. |
| Exact result leases, cleanup and restore contributions | cleanup, reporting and restore files | NF source contributions; Graph Projection persistence/leases; Recovery/Reporting orchestration | keep | E01-NF §29; E03/E05 | Preserve exact result after declaration refresh/retirement and all lease scopes. |
| Canonical indicator creation/reuse plus source binding | indicator link, binding store and transaction participant files | Indicators owns canonical target; NF owns source selector/binding | keep | E01-C02 REQ-02-074C/075; E02/E04 | No analytical-row-to-Core-record or observation conversion. |
| Durable resource invalidation production | `collaboration_producer.go`; server translator | NF semantic event producer; Collaboration persistence/delivery | keep | E01-C03; E03/E06 | Owner transaction appends intent; generic WebSocket sequence/auth/resume stays outside NF. |
| HTTP authentication/framing/error translation | route files, table/indicator HTTP adapters, semantic HTTP adapters | NF transport adapters using platform HTTP and Incidents admission | keep | E01-NF NF-REQ-210; E01-C04; E02 | Reusable semantic code must remain independent. Do not relocate authorization authority into decoders. |
| Fault/randomness/authority/audit controls | `harnesscontrol/*`; harness-tagged server assembly; `internal/testutil/networkflowsupport` | Testing Harness support; exact package destination unresolved | defer | E01-HAR §12.2.9; E07; F03 | Candidate shared test support requires preserving runtime security scanning and guarded owned-process consumers. |
| Browser shell/controller and grid integration | `apps/web/src/networkFlow`; protocol and UI adapter consumers | Existing frontend feature owners, shell, Collaboration session and grid adapter | keep | E08; E01-NF NF-REQ-004 | No frontend code lives in the target. No frontend redesign slice is supported by this backend finding. |
| Timeline, evidence, links, Core records/revisions/view schemas | No owned implementation found in this target's inspected source surfaces | Existing Core owners | defer | NF-REQ-002; Core 01/02; target imports and route inventory | Not applicable to the proposed seam. Absence of a direct import is not proof about all transitive behavior. No move or API is proposed. |

## 4. Public Contract and Behavior Freeze Map

DOC-03 current supplement: [§26.4 Authority and contract decisions](#264-authority-and-contract-decisions). The material below remains historical DOC-01 evidence.

The freeze preserves owner-aligned behavior. It does not promote a latent defect into a requirement. If characterization exposes an owner mismatch, record it separately; any behavior correction **requires later authorization**, including the owner clause, compatibility implications and new acceptance evidence.

All “existing tests” below mean source/selected assertions inspected. None was executed in this planning task. C03/G01 is the concrete missing characterization gate for S02. Other rows require preserving their named tests if later slices expand into those responsibilities.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| C01 — Public Go facade, active/inactive construction and contribution acquisition | Network Flow; Core Extensions publishes claims | E01-NF NF-REQ-210–212; module/composition and application assembly | `query_authoring_test.go` invokes `production_surface_test.go`; `application_composition_test.go` probes retained applications; routes test unclaimed default | Preserve construction/no-IO/API guard. Deepen only if dependencies or registrations change. | high for redesign; low for S02 | No supported public symbol/type change planned. |
| C02 — 19 HTTP operations; request/success/error envelopes, omission/defaults, admission precedence, safe error detail and session handling | NF route semantics; Core HTTP/auth/Incidents authority | E01-NF §§17,21,29; E01-C04 REQ-04-105A; E02 route/admission/error files; E09 route registry | `routes_integration_test.go`; `operation_admission_integration_test.go:assertMajor7AdmissionMatrix`; `network_flow_contract_test.go:AssertJSONAdmissionAndErrorDetails`; graph completion unit probes | Preserve current method/path/status/envelope and multi-invalid-field precedence assertions. No new route tests required solely for replacing a forwarding decoder. | high | Full route inventory below; invalid preview result still becomes Imports `owner_preview_validation_failed`. |
| C03 — Import binding/preview/apply, exact mapping fingerprint and diagnostics, immutable imported rows and replay | Imports workflow and transaction framework; NF analytical target | NF §§7–12,18,29; E04; binding/schema artifacts in E09 | `TestNetworkFlowImportFacadePublishesCompleteBinding` checks metadata; `TestImportOwnerErrorTranslationUsesClosedRegisteredUnion`; CSV/diagnostic vectors; behavior wrappers for source change/replay | **G01/S01:** direct diagnostic raw-object rejection matrix plus public facade validation of a valid prepared result, count/mapping/fingerprint errors and preserved closed error strings. Preserve existing import integration helpers. | high; direct S02 contract | No direct targeted `ValidateImportUnitMappingResult`/diagnostic-shape test found by symbol search in NF/Imports. This is a coverage finding, not proof of total absence of indirect tests. |
| C04 — Table rename/soft-delete admission before replay, immutable receipts, no-op/version behavior, retained-name/limit semantics | Network Flow | `application.go`, `table_receipts.go`, `store.go`; NF §§28–29.1 | `TestTableLifecycleAdmissionAndReplay_Integration`; name/digest unit tests; store lifecycle tests | Preserve replay after lifecycle change, conflicting transaction bodies, version errors, audit/intent counts and rollback when any participant fails. | high | Flow rows are immutable; table metadata mutation is not Core record patch/revision behavior. |
| C05 — Accepted/rejected row queries, filter/sort/keyset semantics, opaque cursors and revocation | Network Flow; Incidents current authority | NF §§12–13; query/filter/SQL/security/read app files | `TestQueryAuthoringAdmission_Unit`; contract helper keyset/cursor/crypto assertions; `TestNetworkFlowPaginationRecovery_Integration`; cursor invalidation wrapper | Preserve equal sort values/null ordering, active-table snapshots, expired/tampered cursor rejection and current authority/lifecycle rechecks. | high | Core entity/view row and test-util mutation/save-state contracts are not alternate owners for NF immutable rows. |
| C06 — Native Graph V2 identity, temporal buckets, limits, selectors, contributors and query completion | NF flow source semantics; Graph Projection generic engine | NF §§14,26–29; E05; generated graph semantics | Streaming/default identity tests; temporal arithmetic/conservation/digest tests; `TestGraphQueryCompletion_Unit`; route graph contributors tests | Preserve deterministic IDs/digests/counter strings, bounded aggregate state, safe cancellation/provider failure and one complete outcome/audit. | high | No projection refresh or graph schema redesign. |
| C07 — Explicit indicator link admission, canonical target visibility, duplicate/replay and atomic binding | NF source binding; Indicators canonical identity/participant; cross-owner transaction coordinator | NF §15; Core 02 REQ-02-074C/075; ports/participants/link stores | `TestIndicatorLinkPublicBoundary`; `AssertIndicatorLinkContractBoundary`; graph/link routes and existing-indicator behavior wrapper | Preserve current target visibility on replay, source selector validation, dedupe, one transaction and empty observation refs where no Core source record exists. | high | Not a generic links module responsibility; no automatic observation creation. |
| C08 — Durable `extension_resource_changed` intent and UI invalidation | NF producer; Collaboration durable stream/WebSocket lifecycle; frontend interpreter | `collaboration_producer.go`; server translator; E08 interpreter; Core 03 | `TestNetworkFlowResourceIntentFailureRollsBackSourceMutation_Integration`; route resource-change helpers; frontend extension-event tests located | Preserve exact payload/profile/resource/change/reason, stable intent identity, transaction rollback and no extra replay event. Before changing wire delivery, inspect Collaboration route/session tests beyond this plan. | high | NF registers no WebSocket route. Do not invent a NF socket path or migrate sequence/auth/resume into it. |
| C09 — Authorization, atomic audit, commit ordering and transaction capabilities | Core Incidents/auth/audit/coordinator; NF operation policy and source changes | E01-C04; operation admission, ports, transaction helpers/participants | Admission matrix, `TestTransactionLifecycle`, deployment-admin-no-bypass and redaction helpers; real auth/audit harness consumers | Preserve current admission before replay, fresh transaction checks, actor/incident attribution, no extra no-op/replay audit and failure rollback. | high | Privileged deployment role is not an incident-access bypass. Fixture SQL is confined to test support, not production authority. |
| C10 — Incident-shared saved graph declaration lifecycle, refresh/jobs, generation/version split, selected binding and retained admission | Network Flow; Jobs/finalizer substrate | NF §§19.1,27–29; saved graph commands/receipts/reads/store/jobs | Saved lifecycle/time-bucket routes; publication-stale-generation test; submitter-role revalidation; materialization timeout/payload; `TestSavedGraphReadFailures_Unit`; cutover admission tests | Preserve admission/replay order, postcommit notification, stale publication rejection, selected/desired snapshot identity, strict payload and safe missing-history classification. | high | Separate from Core saved views and view-schema persistence; no compatibility decoder or migration proposed. |
| C11 — Exact results, owner-scoped leases and bounded cleanup safety | Graph Projection storage/leases; NF declaration selection and source contribution | NF §29; cleanup files; reporting source; E05 | Owner-scoped selected-safe cleanup, publication/lease race tests, dispatcher tests and `TestReportingGraphSourceValidatesLeasesReadsAndReleasesExactResult_Integration` | Preserve lookup→renew→exact read, expiry/races/rollback, lease owner `snapshot_reporting`, purpose `render`, resource job ID; scoped release and valid empty release. | high | Refresh/retire cannot redirect a previously leased exact result. |
| C12 — Recovery/portability and authoritative extension state | NF source facts; Recovery/Incident Portability/Extensions orchestration; Graph restore | NF §§27–29; recovery files; E03/E05 | Obsolete semantic bytes rejected without rewrite; saved-state admission; routes restore composition; payload/startup helpers; reporting lease tests | Before movement, prove real restore source use, sorted pagination at 256 boundaries, entire borrowed-transaction rollback, defaults/no observer and selected+desired snapshot equality. Existing helpers are not a claim of exhaustive coverage. | high | Five source tables: graph views, indicator bindings, rejected diagnostics, rows, tables. Incompatible retained bytes remain unchanged. |
| C13 — Generated protocol, mapping, route, presentation and frontend entrypoint surfaces | Authored NF contracts projected by repository generators | E09 index/policy/topology; generated roots; current major-7 projection test | `TestNetworkFlowV7GraphContractProjection_Unit`; protocol compile surface; frontend boundary policy | Preserve authored contract bytes and generated outputs for S01/S02. Run canonical drift check only in later implementation. | high | No hand edits to generated files; old filenames such as `v3_contract_projection_test.go` do not define runtime major. |
| C14 — Four guarded test control surfaces and real consumers | Testing Harness; owner-local support and harness-tagged server | E01-HAR §12.2.9; E07; harnesscontrol files | Four route suites verify default disabled/auth/arming/rejection; root harness integration helpers exercise real operations | Before any relocation, preserve strict controls, consumption/correlation, owned-process crash, current DB authority, committed audit counts, security scan inclusion and default absence. | high | Randomness schema observed as v3; other control schemas observed as v2. Do not infer versions from family labels. |
| C15 — Test rows, owner slices, evidence routing and runtime/support scan accounting | Testing Harness/tooling accounting | E07/E09; 99 module.networkflow rows and 65 web.networkflow rows | Existing tests linked by manifest selectors; source assertions inspected, no run artifacts created | Keep existing selectors/rows for subtest additions; if new top-level selectors are needed, inspect authored owner inputs and regenerate only in an authorized implementation task. | medium | Phase/family maps are evidence accounting, not runtime boundaries. Capacity test presence does not authorize Core 05 claims. |
| C16 — Configuration, pinned canonicalization, key purposes, safe telemetry and raw-value privacy | NF source configuration/security/telemetry; platform secret/telemetry contracts | configuration/keyring/digest/timestamp/timezone/telemetry files; E01-NF §§16,20,29 | Configuration/effective-limit, key-ring rotation/purpose, Unicode/timestamp and telemetry containment tests; redaction helper | Preserve key separation, limit discovery, stable canonical bytes, non-disclosure of raw causes/secrets and observer failure isolation. | high | No config knobs, key rotation, fixture or generated timezone changes proposed. |
| C17 — Frontend selectors, save-state/controller lifetime and grid adapter consumption | Frontend NF feature; workbook shell/Collaboration session/grid adapter; NF browser contract | E08; NF NF-REQ-004 and browser workflows | Frontend boundary policy assertions inspected; controller/extension-event/selector tests located in owner map, not all assertions audited | No UI movement required. If later scope expands, inspect affected controller/event and browser assertions first and reselect web owner verification. | medium | No direct grid vendor code found in the Go target. NF presentation contracts are not Core view schemas. |

C02 route inventory: prepend `/api/v1/incidents/{incident_id}/network-flow` to each suffix.

| Method | Route suffix | Current handler family |
| --- | --- | --- |
| GET | `/source-profiles` | source profile read |
| GET | `/tables` | table list |
| GET | `/tables/{network_flow_table_id}` | table get |
| PATCH | `/tables/{network_flow_table_id}` | table rename |
| DELETE | `/tables/{network_flow_table_id}` | table soft-delete |
| POST | `/tables/{network_flow_table_id}/query` | accepted rows |
| POST | `/tables/{network_flow_table_id}/rejected-rows/query` | diagnostics |
| POST | `/rows/query` | cross-table accepted rows |
| POST | `/graphs/query` | graph query |
| POST | `/graphs/contributors/query` | graph contributors |
| GET | `/graph-views` | saved graph list |
| POST | `/graph-views` | saved graph create |
| GET | `/graph-views/{graph_view_id}` | saved graph get |
| PATCH | `/graph-views/{graph_view_id}` | saved graph rename |
| DELETE | `/graph-views/{graph_view_id}` | saved graph retire |
| POST | `/graph-views/{graph_view_id}/refresh` | saved graph refresh |
| GET | `/graph-views/{graph_view_id}/result` | selected result |
| POST | `/graph-views/{graph_view_id}/contributors/query` | saved result contributors |
| POST | `/indicator-links` | explicit indicator link |

Imports also calls the facade from its own mapping approval and `/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview` workflow (E04/E09); this is not a twentieth NF route. C14 consists of POST routes under `/api/v1/test/runtime/`: `network-flow-faults`, `network-flow-randomness`, `network-flow-auth-transitions`, and `network-flow-audit-assertions`. They are available only through admitted harness composition.

## 5. Coupling and Boundary Findings

DOC-03 current supplement: [§26.5 Findings](#265-findings). The material below remains historical DOC-01 evidence.

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| F01 — Preview diagnostic validation reaches transport package solely for strict object decoding | `import_facade.go:validatePreviewDiagnosticShape` calls `httpapi.DecodeStrictJSONObject`; `platform/httpapi/json_object.go` simply delegates to `strictjson.DecodeObject`; NF §29 forbids transport helpers in reusable semantics | medium dependency leak; high if validation behavior accidentally changes | must_fix | NF semantic validation + platform/strictjson | Scoped to S02 acceptance: replace dependency after G01 characterization, preserving decoder and all local result/error semantics. |
| F02 — Semantic boundary roots omit Imports facade/error adapter | `semantic_boundary_test.go` explicit roots and application glob exclude `import_facade.go` and `import_owner_errors.go`; transitive AST checker is already present | recurrence of F01 | should_fix | NF boundary tests | Extend existing semantic root guard in S02; inspect reachable functions; do not relax the guard to allow HTTP. |
| F03 — Harness controls are a distinct owner-local support responsibility; relocation changes security accounting | `test_support_inventory.json` declares owner_local/runtime_scan included; harness-tagged server import; backend allowlist; gosec test requires harnesscontrol not excluded, while internal/testutil is excluded | high if moved mechanically | defer | Testing Harness, with NF-specific consumers | RB-002: design preserved scan/build/consumer coverage before choosing destination. No move slice is approved. |
| F04 — Large package contains legitimate source state, adapters and orchestration behind an existing facade | E01-NF NF-REQ-001/002/210/211; module/application/ports; production surface guard | high for speculative redesign | intentional/no_action | Network Flow | Keep module boundary. Framework facade-extraction examples are not proof that this repository still needs extraction. |
| F05 — SQL, pgx and transaction helper live within owner-local private adapters | store/query SQL/transaction and participant bodies; internal instructions permit owner adapters | high if transaction/lock ownership changes | intentional/no_action | Network Flow persistence and caller transaction ports | Do not move SQL to generic platform/domain helpers merely to reduce imports. No shared-helper extraction has semantic justification here. |
| F06 — Graph, Reporting and restore integrations are narrow source-owner contributions | native V2 adapter, one source composer, reporting exact leases, restore registration; E03/E05 | high identity/lease/publication risk | intentional/no_action | NF source; Graph/Reporting/Recovery respective owners | Preserve contribution contracts and separate defaults/observer behavior by consumer. |
| F07 — Mutation effects cross owners explicitly through ports and app translators | NF ResourceIntentAppender, Indicators participation and coordinator, Jobs finalizer; source mutation rollback test | high atomicity/auth risk | intentional/no_action | Source owner + Collaboration/Indicators/Jobs | Preserve transaction boundaries, audit/intent identity and after-commit notification. No hidden Core revision/change-set rewrite discovered in this seam. |
| F08 — Generated artifacts are projections, not edit targets or authority | generated-artifact policy/topology; contracts/network-flow index; protocol entrypoint | high drift/hand-edit risk | intentional/no_action | Contract authors/generators | S01/S02 expect zero contract/generated changes. Reassess if any arise; use authored inputs and Make only under later authorization. |
| F09 — No duplicated Core view-schema or grid-vendor implementation established in target | target imports/source surfaces, NF scope and frontend contract adapter | medium risk of conflating ownership | intentional/no_action | NF analytical contracts; existing view-schema/grid owners | Keep Core saved views/records and NF saved graphs/immutable rows distinct. No frontend/grid slice. |
| F10 — Historical filenames/family labels differ from current runtime version | `v3_contract_projection_test.go` asserts V7; extension-state “v3” file asserts V4 rejection behavior; saved cutover ID remains v6 | medium if names drive design | intentional/no_action | NF compatibility owner | Record repository/framework mismatch: current major7/state4/V2 source contracts supersede historical plan naming. Do not rename for cosmetic consistency. |
| F11 — Preview validator's exact rejection matrix is not directly characterized in located tests | binding metadata and owner-error tests inspected; targeted symbol search; Imports callers depend on validation outcome | medium structural-change uncertainty | must_fix | NF tests | G01/RB-001 gates S02; S01 adds meaningful characterization, not tests mirroring implementation structure. |
| F12 — Required Markdown check has write side effects and default globs omit this tracker | Make prerequisite/run-step chain; markdownlint scripts; `.markdownlint-cli2.jsonc` enumerates selected handoffs but not this file | documentation verification limitation | defer | Documentation tooling | V08/RB-003: do not execute incompatible command or edit tooling in this task; record skipped required check and supported-selection/redirection discovery need. |

No adopted primary-owner contradiction was found in the bounded review. Generic “move test helpers to test-util” guidance versus explicit current owner-local harness/security policy is a planning tension recorded in F03, not proof of conflicting product owners. No claim is made that all owner documents or runtime behavior were exhaustively reconciled.

## 6. Refactor Workstreams

DOC-03 current supplement: [§26.6 Workstreams and dependencies](#266-workstreams-and-dependencies). The material below remains historical DOC-01 evidence.

These are planning workstreams. The DONE entries in §9 mean their documentary deliverables exist, not that a refactor or verification ran. Successors below are the inverse dependency edges.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Source/bootstrap and tracker admission | root | none | WF-01 | Establish scope, baseline and owner posture | framework, AGENTS, owner documents, tracker | Read-only identity/Git/source checks | Canonical target/output admitted; only tracker write authorized |
| WF-01 | File-level inventory | chain | WF-00 | WF-02 | Account for all 129 source/test files and direct interfaces | E02 target files; E03–E08 relevant callers | Enumeration reconciled to §2; depth explicit | 129 unique rows, no silent exclusions |
| WF-02 | Contract-owner map | chain | WF-01 | WF-03, WF-04 | Separate required scope from observed implementation | E01 owners; routes/contracts and consumers | C01–C17 owner/test posture reviewed | Every discovered risk assigned owner and applicability |
| WF-03 | Characterization gap analysis | parallel | WF-02 | WF-05, WF-07 | Identify concrete pre-move evidence gaps | import facade, contract tests, semantic guard; family maps | G01 plus preserved tests; V01/V02 discovery | S01 entry/exit and RB-001 explicit; no fake execution |
| WF-04 | Boundary/coupling diagnosis | parallel | WF-02 | WF-05 | Judge transport, SQL, contributor and harness placement | E02/E03/E04/E05/E07 | F01–F12 traced to source | Narrow seam selected; harness relocation deferred |
| WF-05 | Ownership/facade recommendation | chain | WF-03, WF-04 | WF-06 | Keep supported facade; choose only justified structural change | module, composition, import facade, guard | NF §29; freeze map | No facade redesign, API change or generic-module migration assumed |
| WF-06 | Small-slice sequencing | chain | WF-05 | WF-07, WF-08 | Separate characterization from dependency cleanup | S01/S02 files | Entry gates, commands, rollback and binary exits | Each slice behavior-preserving and readiness stated |
| WF-07 | Harness/accounting plan | chain | WF-06, WF-03 | WF-08 | Retain selector coverage and security scan posture | family manifests, test support inventory, boundary policy | E07/E09; V01–V07 | Existing rows reusable; no harness edits now; F03 deferred |
| WF-08 | Documentation audit and final handoff | chain | WF-06, WF-07 | none | Reconcile rows/dependencies/traceability and preserve baseline | tracker only | §12 document audit; final Git/content/index comparison; V08 skip recorded | Planning artifact current, limits explicit and only tracker changed |

## 7. Proposed Refactor Slice Plan

DOC-03 current supplement: [§26.7 Future implementation slices](#267-future-implementation-slices). The material below remains historical DOC-01 evidence.

These are future actions, **not implementation patches**. The current authorized output is the tracker. S01 is ready only as a later test-characterization task after source revalidation. S02 is **blocked: missing characterization** (RB-001), even though the wrapper's forwarding body makes behavior preservation plausible. No state, transport route, API/type, generated contract, migration, frontend, or harness relocation slice is proposed.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S01 | WF-08 handoff; later implementation authorization and baseline revalidation | Characterize existing preview-result admission before changing its dependency; add subcases to the existing facade contract test so its selector remains valid | `internal/modules/networkflow/network_flow_contract_test.go`; inspect `import_facade.go`, mapping/diagnostic resource helpers; no production edit required | C03; G01; Imports mapping preview/approval error translation | Raw diagnostic object: valid complete 15-member shape including legitimate nullable values; malformed/non-object/null/duplicate nested or top-level/trailing input; extra, missing and same-count wrong members. Public facade: valid prepared result roundtrip; schema/count/fingerprint/source-column failures. Preserve binding/error union and CSV preview assertions. | V01 facade/preview rows; V01 error row; V05 boundary baseline as appropriate to later task | Revert only newly added subcases; preserve baseline test names/fixtures and source behavior | Existing implementation passes meaningful accepted/rejected cases and exact error/result assertions; retained command/run evidence identified; no new production behavior or contract bytes |
| S02 | S01 passing evidence; RB-001 closed; later implementation authorization | Replace the import facade's forwarding HTTP decoder dependency with `platform/strictjson.DecodeObject`; add `import_facade.go` and `import_owner_errors.go` to existing semantic guard roots without weakening checks | `internal/modules/networkflow/import_facade.go`, `semantic_boundary_test.go`; inspect `semantic_reference_boundary_test.go` reachability and `query_authoring_test.go`; shared decoder and HTTP wrapper are read-only dependencies | C01/C02/C03/C13; identical strict syntax rejection and existing local validation/error semantics required | All S01 cases plus existing semantic import/reference guard, binding/error union, CSV preview and relevant import integration consumers | V01 selected rows; V02 service slice for Imports path if scope/risk guide requires; V04 generated drift and V05 backend boundary; reselect with task-guide in authorized implementation | Revert the dependency/root-list change together; no data rollback, contract version or migration expected | No HTTP import/reference from the admitted facade roots; same results/errors for characterization corpus; required selected checks pass; zero unexpected generated/contract/API changes |

S01 should test the validator's current shape responsibility, not strengthen it to validate every diagnostic member's value. The helper currently checks strict object syntax and the exact member set; changing that contract is outside S02. The three local diagnostic error forms remain “not an object”, “shape mismatch”, and “missing &lt;member&gt;” with their existing full prefixes.

A behavior discrepancy discovered during characterization must become a separate **requires later authorization** task; it must not be silently fixed in S01/S02. F03 has no ready implementation slice: destination, scan coverage and actual consumers must first be resolved by the Testing Harness owner.

## 8. Validation Plan

DOC-03 current supplement: [§26.8 Validation obligations and command routing](#268-validation-obligations-and-command-routing). The material below remains historical DOC-01 evidence.

Discovery commands were executed successfully; product validation commands below were **not run**. “Required before implementation?” means an entry baseline before the corresponding production change, not authorization to execute checks during this documentation task. Re-run owner guidance when implementation is authorized because selectors and scopes may change.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| unit — V01 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700,module.networkflow.unit.import_owner_error_translation_rs08_3d3f02e050,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.query_authoring_admission` | Existing facade binding, closed error translation, preview and semantic/API guard selectors | yes, before S02 | Future baseline and post-change evidence. S01 subtests must remain reachable through these selectors. Command formed using discovered public ROWS selector and exact authored row IDs; not executed. |
| integration — V02 | `make test-slice OWNER=module.networkflow SCOPE=services` | Service-backed owner rows, including actual source mutation/import/lifecycle consumers | yes when entering affected behavior; baseline before production change | Discovered from task-guide. Registry-only support tests cannot substitute. Inspect Imports-owned row coverage before widening scope; do not infer from family names. |
| e2e/browser — V03 | `make test-slice OWNER=web.networkflow` | Frontend owner rows selected by its manifest | no for isolated decoder dependency change; yes if browser contracts expand | Discovered from task-guide. `make browser-e2e SUITE=visual` is an explicit-only candidate, not a default gate for S02. No visual run or UI claim here. |
| generated drift — V04 | `make source-check CHECK=generated` | Repository generated-artifact drift | no before S01; required after S02 in authorized implementation | Expected zero authored/generated contract changes. This command is not presumed read-only; use implementation workflow prerequisites. |
| import-boundary/static — V05 | `make lint CHECK=backend-boundary` | Backend package/import policy | yes baseline before S02; repeat after changes | Semantic file/reference guard additionally runs through V01. Generic boundary lint alone will not establish the preview contract. |
| frontend static — V06 | `make lint CHECK=frontend-boundary` | Frontend package boundary policy | no for S01/S02 | Only if later task expands into browser adapters/controllers. |
| full check — V07 | `make check`; `make ci` only when required by later scope/owner guide | Broader repository gates | no before narrow change | Discoverable broader targets, not promised execution. Later implementation follows repository `make agent-finalize` procedure before broader end-of-run verification. Neither is run for documentation. |
| documentation — V08 | `make lint-markdown` | Canonical documentation check | no for production; repository-required for this documentation change | **Skipped**: inspected path writes cache/retained artifacts and has readiness/dependency prerequisites; complete side-effect redirection was not established. Default configured globs also omit this tracker. TODO: discover an admitted Make-owned selection/redirection path before claiming file-specific lint coverage. |
| discovery — V09 | `make help ALL=1`; `make task-guide OWNER=module.networkflow`; `make task-guide OWNER=web.networkflow` | Public command/owner routing discovery | completed planning prerequisite | All three returned exit 0 in prior exploration; no product test/verification result follows from that. |

V08 skip is required by the sole-write boundary and the [refactor-tracker skill](/mnt/c/Users/jochi/.codex/skills/refactor-tracker/SKILL.md): “Do not run commands that may write outside the tracker”. The inspected path includes `tools/task_surface.generated.mk`, readiness/preflight and frontend prerequisites, `tools/harness/execution/run-step.sh`, `tools/harness/static-analysis/markdownlint.sh` and its runner. Cache and result directory overrides alone do not establish that remaining prerequisites are safe. No direct linter invocation, run-then-restore, generated edit or tooling-config edit was used as a workaround.

The existing Markdown config is `fix: false`, but that does not prevent wrapper/cache writes. Its handoff globs list selected files rather than all `docs/handoffs/*.md`; this tracker is absent from that list. A future generic successful invocation alone would therefore not establish that this file was linted. No command is invented to bypass that selection gap.

## 9. Top-Level Work Tracker

DOC-03 current supplement: [§26.9 Current work tracker](#269-current-work-tracker). The material below remains historical DOC-01 evidence.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| T01 | Admit canonical target, safe label/output and baseline | WF-00 | DONE | none | §1; baseline record E10 | Target exists, output absent/admitted, scope fixed |
| T02 | Account for 75 non-test and 54 test Go files | WF-01 | DONE | T01 | E02-001–129 / §2 | Exactly 129 unique rows and explicit inspection depth |
| T03 | Map major7/state4 owners and 19 HTTP operations | WF-02 | DONE | T02 | E01; C01–C17; route table | Owner and test posture for each discovered contract risk |
| T04 | Identify preview validator characterization gap | WF-03 | DONE | T03 | G01/F11; S01 | Direct missing-case matrix and located indirect tests distinguished |
| T05 | Diagnose existing facade, SQL, graph and harness boundaries | WF-04 | DONE | T03 | F01–F12; §3 | Supported keep/split/defer findings, no speculative redesign |
| T06 | Select narrow decoder seam and preserve existing owner architecture | WF-05 | DONE | T04, T05 | F01/F02; no-change recommendations | No public API or owner-state redesign included |
| T07 | Sequence characterization then seam cleanup | WF-06 | DONE | T06 | S01/S02 | Entry gates, rollback, commands and binary exits documented |
| T08 | Map selector reuse and harness security-scan constraint | WF-07 | DONE | T04, T07 | E07/E09; V01; RB-002 | Test accounting separated from runtime design |
| T09 | Add/run meaningful preview-result characterization | WF-03 | TODO | T07; later authorization | S01/G01 | S01 passes on pre-refactor source with retained evidence |
| T10 | Remove facade transport decoder dependency and guard it | WF-06 | BLOCKED | T09; later authorization | S02/RB-001 | Same C03 behavior, admitted semantic roots, selected checks pass |
| T11 | Decide harness-control relocation with security coverage | WF-07 | DEFERRED | separate scoped authorization; RB-002 | F03 | Destination and actual build/scan/consumer coverage justified before any move |
| T12 | Establish admitted file-specific Markdown lint path | WF-08 | DEFERRED | RB-003; documentation tooling scope | V08/F12 | Canonical check covers tracker without unauthorized writes |
| T13 | Reconcile document and final baseline preservation | WF-08 | DONE | T07, T08 | §12; E10; CMD-05/06 | Read-back reconciled tables/IDs; content/index comparison confirms only tracker changed |

## 10. Session Handoff Log

DOC-03 current supplement: [§26.10 Session and evidence handoff](#2610-session-and-evidence-handoff). The material below remains historical DOC-01 evidence.

This is a newly created tracker. No prior on-disk handoff history existed. P00 below retains the useful prior read-only planning result; DOC-01 is the newly appended documentation session. Times use UTC where known; no exact timestamp is invented for the prior turn. “Touched” always means this tracker only. Named source files were read, not edited.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29, exact time not retained | Codex / P00 | Read-only plan; tracker absent | Framework first, E01 owner/support sources, AGENTS/skills; none touched | Git/inventory/source reads; V09 | Main baseline recorded; no writes or refactor | Full per-file inventory still pending then | DOC-01 completes file accounting |
| 2026-10-29 | Codex / DOC-01 | Canonical target/output and authority revalidated | E01; all E02; tracker created | Read-only identity/Git checks and source inspection | Major7 facade remains owner-required; no primary-owner contradiction established | None blocking this bounded plan | Revalidate baseline and governing clauses before later implementation |

### Backend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Existing module/application architecture retained; narrow decoder seam selected | All E02; E03–E06; touched tracker only | `rg`, exact source reads, read-only inventory extraction | 129 file rows; F01/F02 supported; SQL/source contributions intentional | RB-001 gates S02 | Authorize S01 separately, characterize before changing dependency |

### Frontend module boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Consumer freeze mapped; no frontend refactor recommended | E08 client, protocol entrypoint, table owner, Collaboration interpreter and boundary tests; touched tracker only | Frontend reference searches and selected exact reads; P00 web task-guide | Shell/session/controller ownership remains outside backend target; no Core view/grid migration inferred | No frontend slice; deeper UI behavior audit would need scoped inspection | Preserve C08/C13/C17; use web owner guidance only if scope expands |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Route/schema/binding/native graph surfaces frozen | E09 contracts/policies; E02 current projection test; touched tracker only | Read JSON indexes, generated-root policy/topology and route handlers | 19 NF HTTP operations; no proposed API/type/schema/generated changes | No generated drift check run | Expect zero contract/generated diff in S01/S02; never hand-edit generated paths |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29, exact time not retained | Codex / P00 | Public verification routes discovered | Make/task surface, family manifests, test-ops guidance; none touched | V09 commands | Exit 0 discovery only; no product verification | No passing product evidence | Carry commands forward as candidates |
| 2026-10-29 | Codex / DOC-01 | Test assertions and support/security accounting inspected; G01 defined | 54 E02 test files; E07/E09; touched tracker only | Source/selector searches and exact assertion reads | Existing guard reusable; no direct preview-validation case found; harness move deferred | RB-001; RB-002; V08 skipped | S01 uses stable existing selector; preserve real-consumer evidence distinct from registry tests |

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Current authority, replay, audit and guarded control boundaries mapped | E01-C04; operation admission/transactions/ports/security/keyring; E07 server/fixture/scan sources; touched tracker only | Exact source reads and import/reference searches | Production code uses injected ports; test fixture mutation lives in test support; ordinary server build excludes harness profile | Security scan preservation unresolved only for possible relocation | Keep C09/C14/C16 fixed; do not alter controls or authorization in S02 |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | Codex / DOC-01 | Tracker prepared for final editorial/preservation audit | Tracker plus baseline metadata; touched tracker only | Read-only document reconciliation and Git/content/index comparison pending final entry | Planning conclusions bounded; implementation not authorized | RB-001 and deferred RB-002/RB-003 | Complete E10, then hand off S01 as the first later authorized task |
| 2026-10-29 13:34:58 UTC | Codex / DOC-01 final audit | Tracker created and read back; planning complete for the declared scope | Tracker; repository content/index identities; touched tracker only | CMD-05/06 read-only audit | Twelve sections; 129 exact file rows; nine acyclic workflows with reciprocal edges; local paths resolve; baseline outside tracker unchanged | RB-001 gates S02; RB-002/RB-003 deferred; Markdown lint skipped | Revalidate baseline, then pursue S01 only under later implementation authorization |

### Inspected-file ledger outside the target

Grouped paths below are repository-relative; braces enumerate named files, not inferred wildcard coverage. The entire target's individual ledger is §2. Search-only discoveries do not imply body inspection.

| Evidence ID | Exact files or bounded surfaces inspected | Depth and relevance |
| --- | --- | --- |
| E03 | `internal/app/server/{runtime_assembly,extension_admission,network_flow_telemetry,collaboration_intents,server_profile_harness}.go`; `internal/app/configassembly/{configuration,deployment,extension_admission}.go`; `internal/app/extensionassembly/{cross_owner_transaction,incident_portability,recovery,network_flow_jobs}.go`; `internal/app/recoveryassembly/{graphprojection_restore,state_catalog}.go` | Relevant construction/contribution bodies and import references; harness profile, Collaboration translator and restore participant directly read. Not a full server audit. |
| E04 | `internal/modules/imports/{extension_facade,mapping,application_operations}.go`; `internal/platform/httpapi/json_object.go`; `internal/platform/strictjson/object.go` | Closed facade surface, real preview/approval validation calls, exact forwarding wrapper and shared strict decoder bodies. Indicators boundary traced through NF ports/participants and Core owner rather than auditing all Indicators internals. |
| E05 | NF graph/projection/store/reporting/restore files in §2; `internal/app/extensionassembly/network_flow_jobs.go`; `internal/app/recoveryassembly/graphprojection_restore.go` | Native projection invocation, Graph result-reader construction, lease read/release and restore borrowed transaction paths. Downstream Graph implementation not exhaustively audited. |
| E06 | NF Collaboration producer and store/commands; `internal/app/server/collaboration_intents.go` | Source-owned intent projection into Collaboration appender; no separate NF WebSocket server found. |
| E07 | `internal/testutil/networkflowsupport/{authorization,audit,intents,graph_source}.go`; `tools/test_support_inventory.json`; `tools/backend_module_boundaries.json`; `tools/harness/static-analysis/tests/test-go-gosec-targeted.mjs`; harness-profile server file in E03 | Fixture mutation/committed audit consumers, build admission and explicit runtime/support scan posture. Registry tests cannot substitute for real product assertions. |
| E08 | `apps/web/src/networkFlow/{networkFlowClient.ts,networkFlowCollaborationInterpreter.ts,useNetworkFlowTableOwner.ts,networkFlowBoundaryPolicy.node.test.ts}`; `packages/protocol-ts/src/entrypoints/network-flow.ts` | Exact selected bodies/imports and boundary assertions. Frontend test/selector files also located via family maps and targeted `rg --files`; those location-only findings are not coverage assertions. |
| E09 | `contracts/network-flow/{index.json,routes.v1.json,schemas.v3.json,import-facade-binding.v1.json,frontend-entrypoints.v7.json}`; `tools/{generated_artifact_policy.json,execution_topology_manifest.json,task_surface_owner.json,task_surface.generated.mk,task_surface.runtime.generated.mk}`; `tools/test_families/{module.networkflow,web.networkflow}.json`; `Makefile`; `.markdownlint-cli2.jsonc`; `tools/harness/static-analysis/{markdownlint,markdownlint-runner}.sh`; `tools/harness/execution/run-step.sh` | Contract versions/routes/registered inputs, protected generated roots and execution outputs, owner selectors, public targets and check side effects. Other contract registries referenced by the index were inspected at registry/dependency level; generated bodies are downstream projections, not semantic authority. |
| E10 | Git HEAD/index and tracked/nonignored worktree content/file identities; canonical output identity; completed tracker | Baseline and final preservation audit, described below. No product runtime reads or hashes Markdown as part of this work. This is documentation inspection only. |

Generated dependencies identified in source/policy include `internal/gen/contractnetworkflow`, `networkflowroutes`, `networkflowmapping`, `networkflowtz`, `networkflowunicode`, and `packages/protocol-ts/src/generated/network-flow-*`. Their generated posture was inspected; no generated file edit or regeneration is planned.

### Command and preservation ledger

| ID | Commands or inspection method actually used | Result and limits |
| --- | --- | --- |
| CMD-01 | `git --no-optional-locks status --short`, branch/HEAD/index inspection, `git --no-optional-locks ls-files`, diff inspection | Initial repository clean on stated main commit; optional index writes disabled. Target count and no extra target entries established. |
| CMD-02 | `rg --files`, `rg -n`/`rg -l`, `cat`, `sed -n`, bounded `head`/`tail`; read-only Python source/JSON/path inspection | Opened exact sources, extracted imports/declarations/test entrypoints and inspected selected bodies/assertions. Searches were extended along relevant dependencies and caller/contract/test paths only. Search hits were not used as edit sources. |
| CMD-03 | V09's three Make discovery commands | Each exited 0 during P00; establishes available commands/owner routing, not product validation. |
| CMD-04 | Read-only Python content/file-identity and index hashing; no comparison files written | Baseline covers 7,622 tracked files and zero nonignored untracked files, excluding admitted tracker; baseline fingerprints below. Target ignored/untracked scan empty. No mutating validation command run. |
| CMD-05 | Read-back/editorial table, ID, file-universe and workflow dependency audit | Completed: twelve ordered sections, 129 unique inventory rows exactly matching live Go files, consistent table widths, nine acyclic workflows with reciprocal predecessor/successor edges, valid work statuses, existing local file links and ledger paths, no trailing whitespace. This is not canonical Markdown lint or product verification. |
| CMD-06 | Final content/file-identity/index/HEAD comparison and Git diff/status | Completed: tracked and nonignored-untracked digests excluding tracker, index hash/inode and HEAD exactly match E10 baseline. Git reports only the new tracker; tracked diff is empty. Tracker is a regular file with one link. |
| CMD-07 | `make lint-markdown`, tests, formatting, generators, `agent-finalize`, source checks and full checks | **Not executed.** V08 is the skipped required documentation check for the recorded side-effect/selection reasons. Product checks are outside this documentation task. No test run root or result artifact exists for this session. |

E10 baseline: tracked content/identity digest `d0d83ba5294f1f12b5433bc52ba871364b0c541dc83eb4e9eb81e7ff6b2b5f06`; untracked digest `4f53cda18c2baa0c0354bb5f9a3ecbe5ed12ab4d8e11ba873c2f11161202b945`; index SHA-256 `7b50e63df03e37bfbb74005623a8673e0b2c81b7ffb40e10a88e6bf637e50f16`, inode `110884`. The read-only comparison includes path, mode, inode, link count, content size/hash for tracked/nonignored files and index bytes, excluding only this tracker. Relevant ignored command-output locations were considered through invocation-path inspection; no artifact-producing command was executed. This is not a claim to have hashed every unrelated ignored cache in the repository.

## 11. Open Questions and Blockers

DOC-03 current supplement: [§26.11 Gates and follow-up debt](#2611-gates-and-follow-up-debt). The material below remains historical DOC-01 evidence.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | S02 lacks executed direct preview-result/diagnostic-shape characterization | Forwarding-wrapper equivalence does not prove complete caller behavior or test coverage | Later authorized S01; retained passing V01 baseline and preserved Imports contract behavior | BLOCKED — applies to S02 only; planning can finish with this explicit gate |
| RB-002 | Can harness controls move to shared test support while preserving ordinary-build isolation, runtime security scan inclusion and real consumers? | Current internal/testutil exclusion would change scan coverage under a naive relocation | Testing Harness owner decision plus updated scan/build/consumer evidence in a separate authorized scope | DEFERRED — no relocation proposed; does not block S01/S02 |
| RB-003 | What admitted Make-owned invocation both covers this tracker and avoids writes outside its allowed output? | Required docs check has cache/results/prerequisite writes; default globs omit this handoff | Tooling-supported selection and complete safe redirection, or a later task with broader write authorization | DEFERRED — required lint skipped and not claimed successful |

Unknown direct per-file test mappings in §2 are gates only if those additional files are selected for implementation; none expands S01/S02. No unresolved owner contradiction blocks the chosen architectural diagnosis. If later source review establishes one, record the exact clauses as **BLOCKED: owner contradiction** and block dependent decisions without choosing a side.

## 12. Binary Completion Criteria

DOC-03 current supplement: [§26.12 Completion and acceptance limits](#2612-completion-and-acceptance-limits). The material below remains historical DOC-01 evidence.

Completion here is documentary, not implementation or conformance certification.

| Criterion | Status | Evidence / binary test |
| --- | --- | --- |
| Every target file inventoried or explicitly excluded | DONE | 129 unique E02 rows; 75 non-test + 54 test; no exclusions or unreadable entries; direct inspection depth declared |
| Every discovered public contract risk has owner and test posture | DONE | C01–C17, exact 19-route inventory; direct/indirect tests and future characterization distinguished |
| Every workflow has dependencies and exit criteria | DONE | WF-00–08 predecessors/successors/checkpoints; no intended cycles or dangling workflow references |
| Every implementation slice preserves behavior or is explicitly marked as later-authorized behavior correction | DONE | S01/S02 structural scope; any discovered correction requires later authorization and separate task |
| Validation commands discovered or TODO with reason | DONE | V01–V09; V08 selection/redirection gap explicit; no fabricated product result |
| Owner contradictions marked if found | DONE | None established in inspected owner scopes; required blocker wording retained |
| Repository/framework mismatches recorded | DONE | F03/F04/F10/F12: current facade, guarded support accounting, current versions versus historical names, docs validation coverage |
| Handoff sufficient for next agent without repeating broad discovery | DONE | Per-file ledger, owner/contract/finding links, S01 entry and S02 gate, external evidence/command ledger |
| Tracker read back and inventory/dependency/traceability audit completed | DONE | CMD-05: exact live file-set reconciliation, table/ID/path checks and acyclic reciprocal workflow graph |
| Content, file identity and index preserved outside tracker | DONE | CMD-06: E10 baseline digests/index/HEAD unchanged; only new tracker appears in Git status |
| Required documentation check result accurately reported | DONE | V08 skipped with exact limitation; no claim that lint passed or covered the file |

Final outcome: planning completeness is **complete** for this bounded target; S01 readiness is **ready** for a later authorized task after revalidation; S02 readiness is **blocked** by RB-001; implementation authorization is **not granted by this planning invocation**. Harness relocation remains deliberately deferred. Only this tracker was created; no production refactor was performed. V08 remains skipped, and no product validation success is claimed.

## 13. Remediation execution evidence

S04 repair evidence: `nf-s04-store-20261029-02` failed (Make exit 2, runner exit 3); `run-summary.json` and the retained Go stderr identify a test compilation error (`DB.Begin` instead of the owned `BeginTx` port). Corrected the test transaction setup; migration drift was not reached by that invocation.

S04 repair evidence: `nf-s04-store-20261029-03` failed (Make exit 2, runner exit 10) because the new two-table fixture reused an unrelated accepted-row primary key. Its retained `run-summary.json` and Go stderr identify `network_flow_rows_pkey`. Corrected the accepted-row fixture identity while keeping the diagnostic identity identical.

S04 completion (2026-10-29): F14/F16 implemented. Parsing now supplies structural facts; semantic validation stamps the owner NUL-delimited identity using source digest and mapping fingerprint. Diagnostics use complete message tokens, nullable row attribution, and numeric-only samples. Owner Table 21 now closes the previously unspecified mapped-interface failure with `network_flow_invalid_text`; request/resource schema identifiers remain unchanged. Migration 00030 changes only diagnostic uniqueness to `(network_flow_table_id, diagnostic_id)`. Core 01-661 and migration allocation now permit an owner-declared increment beyond the unchanged 29-file baseline. No historical migration was edited. Historical illustrative fixture diagnostics are not used as semantic acceptance evidence; the independent fixed vector and live producer tests are.

S04 validation: `nf-s04-diagnostics-20261029-02` focused preview/identity/privacy (exit 0); `nf-s04-store-20261029-04` storage lifecycle including repeated diagnostic identity across two tables and same-table SQLSTATE 23505 (exit 0); `nf-s04-migration-20261029-01` migration drift (exit 0); `nf-s04-schema-20261029-01` schema checks (exit 0); `nf-s04-generate-20261029-02` generation (exit 0); `nf-s04-markdown-20261029-01` Markdown lint (exit 0). Each run root is `.cartulary/test-results/<run-id>`; graph runs retain `run-summary.json`, tool runs retain their printed tool summary. Corrected test-setup failures are recorded above. F15 retention/counts remain explicitly pending S05; strict semantic admission remains pending S06/S07.

S05 validation repair: `nf-s05-focused-20261029-01` failed (Make exit 2, runner exit 10; retained `run-summary.json` and unit stderr). New retention cases and storage passed; the existing all-rejected error-translation test exposed its validator still requiring `int` after the explicit count became `int64`. Updated that owner-local validator; no wire type changes.

S05 completion: F15 fixed with an explicit row-validation result and separate preview/apply retention policies. Per-row sorting precedes retention; rejected rows and bounded all-rejected samples are independent of retained diagnostic multiplicity. Storage receives the actual rejected-row count. Tests cover limits 0/2/3/4 for three diagnostics, mixed structural/field ordering, preview/apply ID parity, multiple errors per row, all-rejected sample accounting, and 50/51 input records. `nf-s05-focused-20261029-02` focused preview/error tests passed; storage passed in `nf-s05-focused-20261029-01`; format and `nf-s05-markdown-20261029-01` passed. All successful owning exits are 0; earlier translation failure and repair are recorded above. Next S06 tightens the admission boundary without weakening these producer guarantees.

S06 validation repair: `nf-s06-focused-20261029-01` failed compilation (Make exit 2, runner exit 3; `run-summary.json` and unit stderr) after moving validation left two unused imports in the facade. Removed those imports before rerunning.

S06 validation repair: `nf-s06-focused-20261029-02` (Make exit 2, runner exit 10) passed the result corruption corpus but the dependency guard rejected the newly introduced reflection import. Replaced reflection with existing canonical JSON comparison and explicitly admitted standard-library `slices`; transport prohibitions and transitive traversal remain unchanged.

S06 completion: F13 fixed. The private strict decoder validates closed outer/nested objects, required presence and nullability, descriptors and sample digests/counts, mapping variants/default materialization, canonical outer/inner mapping agreement, fingerprints, bounded complete diagnostic counts, attribution, identity, messages, reasons and owner order. It shares the S04 diagnostic validator with retained-state admission planned in S07. The semantic guard includes both new private validation files. `nf-s06-focused-20261029-03` preview corruption corpus and transitive boundary passed; `nf-s06-imports-20261029-01` safe owner translation unit/integration consumers passed; `nf-s06-boundary-20261029-01` backend boundary and `nf-s06-markdown-20261029-01` Markdown lint passed (all exits 0). Both Imports publication and approval continue invoking the owner admission method before returning/storing owner output; no Network Flow semantics were added to generic Imports. Prior failures and repairs remain recorded. Next S07 publishes this coordinated breaking correction and validates retained diagnostics.

S07 generation repair: `nf-s07-generate-20261029-01` failed (Make exit 2, generator exit 1; `generate/tool-run-summary.json`, `generate/generate-contracts/stderr.log`) on the authored initialization digest. The Extensions canonical digest includes a terminal newline; corrected the authored digest and retained the old 3→4 definition as an inert historical ledger record, with no executable migration.

S07 validation scheduling repair: `nf-s07-format-20261029-01` exited 2 at `runtime_plan/configuration_error` before a run manifest or formatting execution; only `harness-runtime-build/tool-run-summary.json` was published (its build passed). Authored selectors had changed since generation. Regenerate the routing projections before invoking graph-backed checks; generation/schema checks chained after format were not reached.

S07 routing repair: `nf-s07-generate-20261029-03` failed (Make exit 2, generator exit 1; `generate/tool-run-summary.json`, `generate/generate-task-surface-and-execution-topology/stderr.log`) because renamed selector names were not ASCII-sorted. Sorted the authored selector array. This also explains the preceding format planning failure; format/schema had not run. The browser adapter now derives its sole supported major from the generated descriptor, eliminating its duplicate hard-coded major.

S07 routing repair: `nf-s07-generate-20261029-04` failed (Make exit 2, generator exit 1; `generate/tool-run-summary.json`, topology step stderr) because the browser test title changed to major 8 before its authored selector. Updated the selector and active graph projection test name together; historical row IDs remain stable.

S07 validation repairs: `nf-s07-schema-20261029-01` (Make exit 2, runner exit 10) identified unregistered retained historical schemas; registered both historical and current schemas without adding runtime major aliases. `nf-s07-state-nf-20261029-01` (Make exit 2, runner exit 10) passed preview/projection/binding cases but the new restore corruption fixture hit the correct immutable-update trigger; changed fixture setup to transactional delete/reinsert representing restore, followed by rollback. `nf-s07-state-ext-20261029-01` (Make exit 2, runner exit 10) passed state tests but an adoption-matrix assertion still expected major 7; corrected that projection expectation. Each failure is retained under its named `.cartulary/test-results` root with `run-summary.json` and unit logs. `nf-s07-client-20261029-01` passed current/unsupported client-major and transport tests (exit 0).

S07 schema repair: `nf-s07-schema-20261029-02` failed (Make exit 2, runner exit 10; `run-summary.json`, JSON-shape unit stderr) on the current contract index schema's stale major-7 constant. Updated that authored constant. Publication review also found approved display-name overrides were not normalized before strict result admission; materialization now uses the existing owner normalizer, with a valid-input roundtrip regression. The corruption corpus was moved outside the retention-limit loop to avoid redundant execution.

Operational cutover procedure (deployment remains pending): retain the previous environment, database, object storage, keys and matching server/browser release as a coherent rollback unit. Obtain original CSV sources and mapping decisions before scheduling re-import; retained tables and diagnostics cannot reconstruct source files. Provision a separate normal empty state-5 environment, configure the intended claims and keys, apply the authored database migrations through `make db-migrate` against that new environment, and let normal admission initialize metadata. Publish the server and browser assets from the same major-8 build, verify discovery/availability and readiness, then re-import originals through Imports and verify counts/diagnostics. Admission of old metadata or old authoritative state is expected to fail without conversion or deletion. Roll back by returning to the untouched previous environment; never downgrade a database after state-5 writes, remove metadata, or delete data to force admission. No deployment or database reset is performed by this task.

S07 schema repair: `nf-s07-schema-20261029-03` failed (Make exit 2, runner exit 10; `run-summary.json`, JSON-shape unit stderr) on hard-coded major-7 semantic assertions in the existing shape checker. Updated all four active Network Flow checks to major 8. Retained-state review found Recovery's declarative post-restore algorithm was not executed: the catalog now requires a bound owner callback, invoked once per profile/algorithm inside the authoritative restore transaction. Network Flow checks state-5 metadata and diagnostic semantics before commit; empty owner state may still initialize normally. Removed the obsolete state-1 pristine restore allowance. These are F17 implementation obligations, not a general Recovery redesign.

S07 schema repair: `nf-s07-schema-20261029-04` failed (Make exit 2, runner exit 10; JSON-shape unit stderr) because the semantic error inventory omitted the newly specified optional-interface diagnostic. Added that row to the existing exact inventory. Owner-guide discovery for `app.extensionassembly` returned unknown owner (exit 2); its verification routes belong to the already selected `app.server` owner. `module.recovery` guidance passed.

S07 validation repairs: `nf-s07-schema-20261029-05` failed (Make exit 2, runner exit 10) because the edited shape checker required topology regeneration; rerun generation before schema verification. `nf-s07-restore-20261029-01` failed compilation (Make exit 2, runner exit 3) because a same-package Recovery test imported Network Flow through Imports back to Recovery. The atomic-restore test now exercises the callback boundary with a transaction-reading validator; Network Flow's real validator is exercised in its owner integration test. `nf-s07-server-20261029-01` publication/discovery and `nf-s07-state-nf-20261029-03` real owner state/restore validation passed. Final producer review corrected an overly strict empty-policy admission check: a null policy for a non-nullable current-profile scalar must reject that row, not panic or reject an otherwise well-formed mapping. Added the owner-required roundtrip regression.

S07 completion: F08/F10/F17 now use major 8, document 8.0.0, state 5/minimum 5, frontend entrypoint v8 and operation policy v2. Resource/request schemas, Graph semantic query v2, Jobs payload v1 and diagnostic algorithm v1 remain unchanged. Historical contract files and migration ledger facts remain inert history; active server/browser publication admits only major 8 and no legacy state migration. Retained diagnostics are validated through the shared owner boundary at startup and through a bound callback before Recovery's authoritative transaction commits. No prior data was converted or deleted.

S07 child evidence: `nf-s07-state-ext-20261029-01` state-1/2/3/4 rejection, initialization and state-runtime cases passed, but the overall invocation failed on its separate stale registry assertion. That assertion passed in `nf-s07-registry-20261029-02`. S07 successful invocations (owning exits 0): `nf-s07-repaired-20261029-01` validates real current/restored diagnostics, mapping roundtrips and the actual Recovery catalog binding. `nf-s07-restore-20261029-02` validates callback execution/rollback in the authoritative restore transaction and the catalog contract. `nf-s07-server-20261029-01` covers publication/discovery; `nf-s07-client-20261029-01` covers supported and unsupported majors. `nf-s07-generate-20261029-09`, `nf-s07-schema-20261029-06`, `nf-s07-generated-20261029-01`, `nf-s07-boundary-20261029-01` and `nf-s07-markdown-20261029-01` passed. Earlier failures remain explicit diagnostic history, not final acceptance evidence.

S08 entry (initial plan reading; migration coverage corrected below): focused remediation gates passed. Reviewed the canonical CI plan (`make target-plan TARGET=ci DETAIL=summary` and `DETAIL=graph`; exits 0; read-only plan saved at `.cartulary/nf-ci-target-plan.txt`). CI selects 424 work units including backend/frontend verification, build/ordinary-versus-harness policy, runtime security scans, schemas, generated policy and owner boundaries. Its graph has no standalone `migration-drift` operation; S04 migration drift remains slice evidence and the final portfolio will run that missing obligation after source freeze. No extra visual or benchmark qualification is selected. Run `make agent-finalize`, review mutations, then freeze product source and run `make ci`; S08 remains incomplete until final acceptance and handoff reconciliation pass.

S08 source review: `nf-s08-finalize-20261029-01` passed (exit 0). Review found only intended generated changes. Added a second independent diagnostic identity vector for null column/field attribution and cleaned import grouping and the migration allocation variable name before the final freeze; these review changes require finalization again.

S08 review correction: the earlier S00 wording incorrectly treated implementation authorization as completed human owner adoption. New amendments are now explicitly proposed in Network Flow, Extensions and affected Core declarations. Unchanged owner requirements retain their adopted baseline authority. Human review of the coordinated amendment and projection fidelity is an explicit release handoff gate, not a test result or a deployment action performed here.

S08 first final portfolio: `CARTULARY_TEST_RUN_ID=nf-s08-finalize-20261029-02 make agent-finalize` exited 0. Frozen product source was `sha256:1a3244d9a3946782afa70bbd1a40b92b3021bf6ce8616695865af61925ad0bee`. `CARTULARY_TEST_RUN_ID=nf-s08-migration-20261029-01 make migration-drift` exited 0 (8/8 units, 19 cleanup steps passed). `CARTULARY_TEST_RUN_ID=nf-s08-ci-20261029-01 make ci` exited 2 (owning graph exit 10, `product/test_assertion_failure`; 420/424 units passed, four failed, none skipped/cancelled; 726 cleanup steps passed). Evidence is in the named `.cartulary/test-results/<run-id>/run-summary.json`, `target-summaries/` and `cleanup-results.json`. CI is not accepted.

S08 failure diagnosis before repair: CI rows `module.networkflow.unit.query_authoring_admission` and `module.extensions.integration.admission_preflight_v6` each exited 10. Retained unit logs `go-0a1eb56aa3639a55-219fa29a18aa/stderr.log` and `go-8121b7778a618816-c071127b976b/stderr.log` identify the missing reviewed export entry for the new real restore validator and an obsolete expectation of one executable migration. The `harness.lint_go_format` projection identifies three unformatted test files; `harness.lint_go_staticcheck` identifies the pre-existing unused `requireWritableStringNormalization` helper in Incidents request tests (confirmed in baseline HEAD). All four failures require repair; the unused helper can be removed without changing product behavior. Failed observer usage attempts (`DETAIL=failures`, logs without a target, and an internal target name) are diagnosis-command errors, not additional product failures; supported summary/children and retained structured results established the findings.

S08 portfolio correction: the earlier plan inspection incorrectly concluded CI omitted migration drift. The actual retained CI projection includes passing `migration-drift` and its eight units. The separate passing drift run is redundant historical evidence; do not repeat it in the repaired final portfolio. CI also passed schema/generated checks, browser suites, security scans, builds, store tests and Recovery tests. These are earlier-source results and do not close final acceptance after repair.

S08 source-review obligations: admit Table 10-A2's zero-record shape (while retaining the parser's separate header-only error); replace active legacy diagnostic test placeholders with valid owner diagnostics and exercise the same identity across incidents; run Make formatting. Documentation review also reconciled remaining active major-7/state-4 adoption and acceptance declarations. Re-run affected focused rows, finalize again, freeze source and execute CI before handoff completion.

S08 repair scheduling: `nf-s08-format-20261029-01` passed (exit 0). The attempted `test-slice ROWS=...` under `nf-s08-repair-20261029-01` exited 2 at public input admission because exactly one of OWNER/SUITE is required; no product row ran. Split the mixed-owner repair selection into exact Network Flow and Extensions owner invocations. The earlier `agent-finalize` successes did not imply Go formatting had run.

S08 user-directed portfolio revision: after the first CI run, the user explicitly instructed, “Don't re-rrun a full CI, validate with focused tests.” This supersedes the planned repaired full-CI invocation. Final acceptance will use focused affected behavior, static checks, generated/schema checks and documentation lint on finalized source. The failed earlier CI remains broad diagnostic evidence, not a final-source pass or complete-owner audit. No second full CI is authorized or required for this handoff.

S08 focused repair diagnosis: `nf-s08-repair-nf-20261029-01` exited 2 (runner 3, `infra/preflight_error`; 18 cleanup steps passed) because new fixture code referenced private owner symbols from the external `networkflow_test` package. Its unit stderr identifies the inaccessible diagnostic helpers and request type. `nf-s08-lint-go-20261029-01` exited 2 (lint units exit 1) for the same compile defect in vet/staticcheck; Go formatting passed. Use an independently constructed valid test diagnostic, the existing exported retained-state validator, and a test-build request alias rather than adding production exports. `nf-s08-repair-ext-20261029-01` and `nf-s08-markdown-20261029-01` passed with owning exit 0. Each exact root is `.cartulary/test-results/<run-id>`.

## 14. Completed remediation finding dispositions and phase gates

This section supersedes historical findings' scheduling and completion states without erasing their evidence. The user-revised focused final portfolio and documentation checkpoint passed; S08 is complete. The scope remains bounded remediation, not a complete Network Flow conformance audit.

| Finding | Current disposition and areas | Owner / rationale and durable benefit | Compatibility and risk if unresolved | Required validation / evidence |
| --- | --- | --- | --- | --- |
| F01 | Fixed in S02: use `strictjson.DecodeObject` directly; implementation | NF §29, NF-REQ-210; semantic code depends on the existing shared primitive, preserving one-way ownership | No independent wire/storage change; transport coupling otherwise spreads into reusable semantics | Dedicated preview characterization and transitive semantic-boundary tests; S02 evidence, S08 portfolio |
| F02 | Fixed in S02/S06: guard facade, error adapter and both new private validation files; tests | NF §29; guard actual public contribution entrypoints and local reachability rather than an incomplete root sample | No compatibility change; incomplete roots otherwise conceal architectural regression | Existing direct/alias/local-reference guard and boundary lint; S02/S06 evidence, S08 portfolio |
| F03 / RB-002 | Resolved, no relocation: keep harness controls owner-local; architecture/documentation | Testing Harness §12.2.9 and E07/E09 build/scan evidence; guarded runtime controls need owner context and runtime security scanning | Avoids relocation coupling and accidental scan exclusion; unreviewed movement could weaken security evidence | Ordinary-build exclusion, runtime scan inclusion and real rollback/audit/auth consumers passed in CI-01; registry presence alone is insufficient |
| F04 | Resolved, retain module/application boundary; architecture/documentation | NF-REQ-001/002/210/211; the source owner already has a coherent facade and retained application bundle | No compatibility change; speculative splitting would fragment ownership and lifecycle | Construction/source tests passed in CI-01; the repaired public-surface guard passed in the final Network Flow selection |
| F05 | Resolved, retain private SQL/transaction adapters; architecture/documentation | NF persistence and caller transaction ports, backend procedure; cohesive storage and lock ownership | No compatibility change; moving owner SQL into generic helpers adds coupling and obscures atomicity | Store lifecycle, transaction rollback and boundary checks in CI-01 |
| F06 | Resolved, retain distinct Graph/Reporting/Recovery owners; narrow F17 restore callback added in S07 | NF §§14/29, Graph V2, Reporting exact lease contracts, Recovery authoritative restore; owners exchange narrow capabilities | Unchanged Graph semantic query v2; unchecked restore otherwise publishes invalid owner state | Projection/lease contracts plus actual owner callback and restore commit/rollback tests; S07 evidence, S08 portfolio |
| F07 | Resolved, retain explicit mutation ports and application translators | NF §29.1 and Core transaction/audit/intent ownership; one commit boundary and owner-specific effects | No ownership transfer; implicit cross-owner effects would threaten rollback and authorization | Replay, rollback, audit and real consumer tests in CI-01 |
| F08 | Superseded: coordinated authored and generated updates in S01/S04/S07 | Core 00 precedence and generated-artifact policy; projections follow reviewed owner decisions | Zero-generated-diff restriction no longer applies; hand edits or partial publication would drift | Make generation, schema/generated checks and finalized CI; no historical migration edits |
| F09 | Resolved, retain Core saved views/records versus NF saved graphs/immutable rows | NF scope and Core view-schema/grid ownership; no duplicate frontend subsystem introduced | Current client major changes only; conflating resources would burden later extension | Browser adapter/publication and native Graph contracts in CI-01; no visual redesign obligation |
| F10 | Resolved in S07: update active contracts/selectors, retain historical filenames and stable row IDs | NF-REQ-212 / EXT-REQ-170; version truth resides in active declarations, not cosmetic filenames | Active major 8/state 5 only; inert v7/v1 contract files and old migration ledger facts remain history | Discovery/client/registry/schema/generated checks; historical names are not active aliases |
| F11 / G01 / RB-001 | Fixed in S01: dedicated preview admission selector; tests/routing | NF Table 10-A2 and Core Imports publication/approval boundary; meaningful public behavior independent of binding metadata assertions | Test-only baseline change; preserving malformed acceptance would obstruct corrections | S01 valid/framing/safe-translation baseline; S04–S07 correction corpus; S08 portfolio |
| F12 / RB-003 | Fixed in S03: module-tracker glob; tooling/docs | Canonical Markdown maintenance; future module trackers receive ordinary lint coverage | No product change; omitted controlling artifacts otherwise escape lint | `make lint-markdown`, selecting both module trackers; S03 and final handoff check |
| F13 | Fixed in S06: private typed strict result admission; implementation/tests/spec clarification | NF Table 10-A2, Tables 10-A/10-B and §12.4; one closed boundary verifies descriptors, canonical mappings, fingerprints and diagnostics | Malformed internal output now rejects through `owner_preview_validation_failed`; unchecked output could disclose unsafe data or approve inconsistent mappings | Required/unknown/duplicate/null/type and semantic corruption corpus; valid producer roundtrips; both Imports consumers; S06/S07 evidence and S08 portfolio |
| F14 | Fixed in S04: source/mapping-derived identities and shared diagnostic construction/validation; implementation/tests/contracts/spec | NF §12.4, Tables 12-E/12-F and 21-A/A1; stable provenance, closed message/reason relationships and safe scalar treatment | Erroneous old values change under fresh-state cutover; unresolved identities collide or misattribute sensitive data | Independent scalar and null-attribution vectors, preview/apply parity, sensitive/row-level samples and retained-state corruption tests; S08 portfolio |
| F15 | Fixed in S05: explicit validation result and separate preview/apply policies; implementation/tests/spec | NF Table 10-A2, §12.4, NF-AC-063; complete bounded preview, ordered retention and independent rejected-record accounting | Corrected low/exact-limit behavior; conflation otherwise disables preview or misreports rejection/truncation | Limits 0/N/N+1, mixed structural/field failures, multiple diagnostics per row, all-rejected sample and 50/51-record cases; S08 portfolio |
| F16 | Fixed in S04: migration 00030 adds table-scoped diagnostic key; SQL/implementation/tests/spec | NF §12.4/NF-REQ-212 and Core REQ-01-661 increment; physical uniqueness matches table-independent semantic identity | New physical schema; no ID algorithm fork or old data rewrite; otherwise identical valid source imports collide across tables | Cross-table duplicate identity accepted, same-table duplicate rejected, incident isolation/pagination/atomic rollback and migration drift; S08 portfolio plus drift |
| F17 | Fixed in S07: current-only state-5 admission and transactional restore validation; spec/contracts/code/tests/docs | NF-REQ-212, §29.2 restore obligation, Extensions state ownership and EXT-REQ-170; corrected runtime never silently serves incompatible retained facts | Breaking major-8/state-5 cutover; old state remains untouched and blocks admission; originals required for re-import | Fresh initialization, unchanged old-state rejection, malformed state-5 rejection, valid restore and atomic invalid-restore rollback; S07 evidence and S08 portfolio |

| Phase / workstream | Dependency | Principal risk and control | Exit / handoff |
| --- | --- | --- | --- |
| S00 | Revalidated baseline | Mistaking defects for requirements or implying human adoption; consistent proposed amendments and explicit review gate | Owner decisions and all finding dispositions recorded; proposed status retained for adoption review |
| S03 | S00 | Unbounded historical lint expansion; constrain selection to module-tracker glob | Canonical lint covers both trackers and passes; RB-003 closed |
| S01 | S03 | Tests copying implementation assumptions; use public facade invariants and independent valid/framing cases | Passing pre-change characterization retained; G01/RB-001 closed |
| S02 | S01 | Decoder movement changes admission/error translation; retain characterization and transitive guard | Shared primitive, safe public failure translation and boundary checks pass |
| S04 | S02 | Identity/privacy drift or global collisions; one typed validator, independent byte vectors, additive migration | Identity/privacy and storage uniqueness checks pass |
| S05 | S04 | Parser/semantic ordering loses early errors or preview grows unbounded; merge in source-row order and retain only under explicit mode | Exact retention/count/order and bounded-preview cases pass |
| S06 | S05 | Strict boundary rejects supported producer output; paired producer roundtrips and independently corrupted results | Closed result/mapping/diagnostic corpus and both Imports consumers pass |
| S07 | S06 | Partial major publication or invalid restored facts; generated current-only registry and owner validator inside restore transaction | State/client/publication/schema/generated checks pass; operational cutover documented |
| S08 | S07 | Treating historical, cached or partial evidence as complete; inspect exact final roots, source identity, owning exits and cleanup | All required portfolio checks pass and tracker is reconciled; human adoption/deployment remain explicit handoff gates |

No phase may skip its exit and begin a dependent phase. S08 repairs reopen only the affected obligations and require renewed finalization/source qualification before completion. Historical evidence remains useful for diagnosis but cannot replace final compatible acceptance.

## 15. Completed remediation validation and operational handoff

### Source, environment and acceptance scope

Baseline and current HEAD are `a182673ae3b460de9c0d078d3315f17e18c4f36a` on `main`; implementation changes remain in the working tree. The pre-existing untracked tracker and its original inventory, IDs and historical evidence were preserved. Final executable-source identity is `sha256:3c532611ae42c1f142d42688f6d60984bbb3704e795109904004923d84fc4122`, recorded identically by all final focused manifests. Documentation is outside executable-source identity and was edited only for the human handoff after the product freeze.

Execution environment: Linux `6.6.114.1-microsoft-standard-WSL2`, x86_64, GNU/Linux; repository `/home/jochi/code/cartulary`; execution date 2026-10-29. The pin registry specifies Go 1.27.3, Node 24.15.0 and pnpm 10.33.0; CI-01's toolchain check passed. Final manifests record toolchain digest `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f` and system digest `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

The user changed final acceptance to focused validation after CI-01. No second CI ran. CI-01 remains a failed earlier-source invocation: its four failures are individually repaired and requalified by the final focused checks. Passing earlier-source publication, browser, build, scan, Imports, Recovery and migration projections are identified as such; they are not a full-source current acceptance audit. No full Network Flow conformance, benchmark qualification, deployment adoption or complete-owner audit is claimed.

### Final commands and evidence

Every command below ran from the repository root. All completed product commands have owning exit 0. Each run root is `.cartulary/test-results/<run-id>`; graph results use `run-summary.json`, `run-manifest.json`, `target-summaries/` and `cleanup-results.json`.

| Run ID | Exact Make selection | Result / purpose |
| --- | --- | --- |
| `nf-s08-format-20261029-02` | `make format` | PASS; includes authored untracked Go files |
| `nf-s08-finalize-20261029-03` | `make agent-finalize` | PASS; generated refresh unchanged, zero rewritten files; `unit-artifacts/finalize-summary.json` |
| `nf-s08-final-nf-20261029-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.query_authoring_admission,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.integration.pagination_recovery,module.networkflow.integration.resource_change_intent_replay_and_privacy_f1a2b3c4d5,module.networkflow.integration.extension_state_v4_v1_rejection` | PASS, 6/6 work units; preview/diagnostics/retention/strict admission, dependency and export guards, storage identity and incident isolation, pagination/replay/privacy, current-state and real restore binding |
| `nf-s08-final-ext-20261029-01` | `make test-slice OWNER=module.extensions ROWS=module.extensions.integration.admission_preflight_v6,module.extensions.integration.network_flow_state_v4` | PASS, 3/3 work units; no executable prior-state migration, fresh initialization and unchanged old-state rejection |
| `nf-s08-lint-go-20261029-02` | `make lint CHECK=go` | PASS, 6/6 work units; formatting, vet and staticcheck, including the removed unused baseline helper |
| `nf-s08-schema-20261029-01` | `make source-check CHECK=schema` | PASS, 3/3 work units; authored schema/contract consistency |
| `nf-s08-generated-20261029-01` | `make source-check CHECK=generated` | PASS, 5/5 work units; generated consistency after finalization |
| `nf-s08-handoff-markdown-20261029-01` | `make lint-markdown` | PASS, owning exit 0; summary under `adhoc/lint-markdown/tool-run-summary.json` |

Each invocation set `CARTULARY_TEST_RUN_ID` to the exact table value. These work-unit counts are not test counts. Final Network Flow and Extensions invocations used no execution-result cache hits. Their cleanup checks passed (21 and 14 steps respectively); Go lint/schema/generated cleanup passed (15/12/14 steps), with none failed or blocked. No selected final product work was skipped or cancelled. Earlier slice evidence and exact failure roots remain in §13.

`agent-finalize` did not select retained-run maintenance because no compatible successful full warm-check root was supplied. Its optional retained-evidence and scheduler checks therefore report `results-dir-not-provided`; no redundant warm check was run to obtain one. Full CI was not repeated at the user's explicit direction. No extra visual or benchmark qualification was selected. No required check under the revised focused portfolio is waived. The final documentation checkpoint passed.

### Delivered changes and compatibility boundary

The semantic import entrypoints now use the shared strict decoder under the transitive owner guard. New private preview and diagnostic validators enforce required presence, closed nested objects, canonical mapping equality, deterministic identities, safe messages/samples and complete preview accounting. Preview diagnostics are bounded by 50 records; apply retention uses the first configured N in owner order, with actual omission and independent rejected-record counts. Active integration fixtures now use independently constructed owner-valid diagnostic bytes, including the same identity in two tables of one incident and a third table in another incident; duplicate identities inside a table still fail.

Authored migration `db/migrations/00030_network_flow_diagnostic_identity.sql` changes diagnostic uniqueness to `(network_flow_table_id, diagnostic_id)`. Versions 1–29 remain unchanged. The allocation/catalog/schema projections were regenerated through Make. This is a physical key change, not a diagnostic-ID format fork or historical data conversion. Migration drift passed both its recorded standalone invocation and CI-01; the SQL and allocation inputs were unchanged by final repairs.

Publication is coordinated at Network Flow document/public major 8, durable state 5 and minimum admitted/migratable state 5. Current frontend entrypoint is v8 and operation policy v2. Existing request/resource schema IDs, diagnostic algorithm v1, Graph semantic query v2 and Jobs payload v1 remain. Historical filenames and ledger facts remain inert history. The active prior-state migration route and dual-major compatibility are absent. Startup and authoritative restore validate diagnostic semantics; restore rejection rolls back before publication.

Harness controls remain in their owner-local package, preserving ordinary-build isolation and runtime scanning. Graph, Reporting, Recovery, Indicators and Collaboration retain their respective responsibilities. The sole unrelated source cleanup removes a pre-existing unused Incidents test helper that blocked required Go static analysis; no Incidents behavior changed. `docs/domain.md` required no edit because vocabulary and owner navigation are unchanged.

### Adoption and rollout instructions

Human owners must review the proposed Network Flow 8.0.0 / Extensions 0.13.0 / coordinated Core amendments, including the diagnostic reason registry and migration increment, and explicitly adopt them before release. Reviewers must also verify their fidelity to authored projections. Tests cannot establish those judgments. Deployment adoption is separate from implementation completion.

Older Network Flow data was **neither converted nor deleted**. Incompatible retained metadata or authoritative state blocks admission without rewriting it. Original CSV source files and mapping decisions are prerequisites for re-import; retained tables and diagnostics do not reconstruct those sources.

Retain the previous environment's database, object store, keys and matching server/browser assets as one rollback unit. Provision a separate empty environment, apply the authored migrations through `make db-migrate` there, configure claims/keys, and let normal admission initialize state 5. Publish the matching major-8 server and browser assets together; verify discovery, supported client behavior and readiness before re-importing original sources. Verify imported row counts, retained diagnostics and cross-incident isolation before operational acceptance.

Rollback returns traffic to the untouched previous environment and matching release. Do not downgrade a database after state-5 writes, remove metadata, convert old rows, or delete data to force admission. This effort performs no deployment or destructive reset.

### Completion checklist

| Criterion | Status / evidence |
| --- | --- |
| S00/S03/S01/S02/S04/S05/S06/S07 completed in dependency order with tracker checkpoints | PASS; §1 and §13 |
| F01–F17 and G01/RB-001/RB-002/RB-003 explicitly fixed, superseded or resolved without unnecessary relocation | PASS; §14 |
| Required focused final behavior and static/generated/schema checks pass on finalized source | PASS; exact roots and source identity above |
| Earlier CI failures diagnosed, repaired and focused checks passed; no false full-CI claim | PASS; §13 and final portfolio |
| Historical inventory/evidence retained; compatibility, original-source requirement and rollback documented | PASS; §§2–13 and this handoff |
| Human owner adoption and deployment separated from implementation | PASS; amendments proposed, human adoption/release pending |
| Final controlling-artifact Markdown checkpoint | PASS; `nf-s08-handoff-markdown-20261029-01`, owning exit 0 |

S08 completion (2026-10-29): implementation, validation and handoff are complete under the user-revised focused scope. G01/RB-001 is closed, RB-002 is resolved by retaining owner-local controls, and RB-003 is closed by canonical glob coverage. No remediation blocker remains. Human owner adoption and operational deployment remain explicit future gates; older Network Flow data was neither converted nor deleted. The completion-status edit is followed by one final documentation-only lint invocation, `CARTULARY_TEST_RUN_ID=nf-s08-completion-markdown-20261029-01 make lint-markdown`; no product checks or finalization are repeated for that Markdown-only checkpoint.

## 16. DOC-02 baseline, inventory and owner decisions

### Current source and permitted changes

Artifact mode: **normal**, for the bounded cleanup and hardening plan requested by the user. Repository root is `/home/jochi/code/cartulary`; canonical target remains `internal/modules/networkflow`, with related Imports source-reader, contract publication and verification dependencies explicitly included. The sole permitted tracked output is `docs/handoffs/networkflow-module-refactor-tracker.md`. The planning framework remains `docs/handoffs/cartulary_modular_refactor_planning_framework.md`.

Revalidation found clean `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, with no staged, unstaged or non-ignored untracked changes. This is the same baseline as the immediately preceding read-only planning turn. S00–S08 changes are now committed; §15's working-tree description is the historical S08 posture, not the DOC-02 posture. Prior test results remain historical results, not checks of this iteration's future implementation.

The tracker and its parent directories are not symlinks. The tracker is a regular file with one hard link; initial identity was device `2096`, inode `110366`. Its pre-edit SHA-256 was `e72ca49d3645047236d4a209af5c61a0492228c91bcf34f443cd261b16e565f9`; the Git index SHA-256 was `2e73fc3ff1e2a15ffe8c9b3e37443b524125bef093ea439e7827f9aec4e871d8`. These are manual document-preservation observations, never inputs to product checks, generators, conformance or release evidence.

The user authorized canonical Markdown lint and its ordinary ignored cache/result artifacts. That explicit instruction supersedes DOC-01's narrower command-write restriction and its historical V08 skip. No other managed verification, generator, formatter, dependency change or service operation is part of this documentation update.

### Inventory delta and inspection coverage

The current universe is all tracked and non-ignored untracked files under the literal target: **132 Go files, 77 non-test and 55 test files**. Every one of the original 129 paths in §2 still exists. The following three paths complete the current inventory; no original file was removed. No target path is generated. Generated projections outside the target remain explicitly accounted dependencies.

| Evidence ID / added path | Current responsibility and owner | Current inspection disposition | Finding / validation |
| --- | --- | --- | --- |
| E02-130 / `internal/modules/networkflow/diagnostic_validation.go` | Private Network Flow diagnostic identity, message, attribution, reason and limit validation | Directly inspected relevant validator bodies; retained as the shared semantic boundary | F23; C21; V13/V14 |
| E02-131 / `internal/modules/networkflow/import_preview_test.go` | Network Flow preview construction/admission and diagnostic regression assertions | Directly inspected relevant assertions, including the forwarding-helper caller; no test executed | F19/F23; C18/C21; V10/V14 |
| E02-132 / `internal/modules/networkflow/import_preview_validation.go` | Private typed preview-result decoder and admission | Directly inspected decoder, wrapper and nested admission bodies | F19/F23; C18/C21; V10/V14 |

Current direct-inspection set, including the preceding planning turn on the identical baseline: `application_composition.go`, `csv_parser.go`, `diagnostic_validation.go`, `import_facade.go`, `import_preview_test.go`, `import_preview_validation.go`, `indicator_link.go`, `mapping.go`, `network_flow_contract_test.go`, `store.go`, and `transaction_participants.go`. These 11 paths received relevant-body or selected-assertion inspection, not an assertion that every line was audited. Every other target path in the §2 inventory is **metadata-only for DOC-02**: 121 files enumerated and accounted, with historical DOC-01 inspection retained separately. No file is silently excluded or claimed unreadable. This partition, plus the three delta rows, gives every current target file a disposition without rewriting historical evidence.

The next discovery boundary is the named files and real consumers for each workstream. Consumer searches are lexical evidence, not proof that every dynamic or external consumer is absent. S10 must establish consumer closure before removing artifacts; S12 must characterize source-port consumers; S13/S14 must establish focused baseline evidence before replacing the parser/apply path. Other module families are outside this bounded iteration's conformance claim.

### Current evidence ledger

| Evidence ID | Exact inspected sources or observation | Finding supported / limits |
| --- | --- | --- |
| E11 | Read-only Git baseline, canonical-path/file-identity observations, tracker fingerprint, and current inventory comparison against §2 | Current clean baseline and 129 + 3 inventory; no product validation result |
| E12 | The 11 directly inspected Network Flow files listed above; `rg` caller/reference searches | Wrapper used by a test, `io.ReadAll`, materialized `Records`/`Rows`, late accepted-row limit, and `getAcceptedRowForLink` calling full-table `ListRows`; F19/F21/F22/F23 |
| E13 | `internal/modules/imports/source_streams.go`, `extension_facade.go`, `apply_coordination.go`; relevant `unit_outcomes.go` references; `internal/platform/postgres/db.go` | Source port loads `source_bytes` into memory; capability lookup also uses that loader; Imports owns the borrowed transaction, terminal outcome, cancellation and commit barrier; F20/F21 |
| E14 | `contracts/network-flow/index.json`, `frontend-entrypoints.v7.json`, `operation-policy.v1.json`; current frontend entrypoint reference in `contracts/protocol-ts/frontend-entrypoints.v2.json`; matching schemas and `tools/harness_schema_attachments.json`; generated artifact references in `internal/gen/contractnetworkflow/artifacts_gen.go` | Active major 8 uses operation-policy v2/entrypoint v8, while obsolete v7/v1 artifacts remain attached and packaged; F18. Generated output was read, never edited. |
| E15 | `internal/modules/extensions/catalogs.go` and `state.go`, historical ledger declarations in `contracts/extensions/fragments/network_flow_activity.profile.json`; existing semantic guard and test-family references | Ledger definitions authenticate committed rows without making migrations executable; retain them. Test/guard presence supplies routing evidence, not a fresh pass. |
| E16 | Current NF header and §§8.2/9.2/10/12.4/20/29; Core 01 analytical binding/unit contract; Core 04 database privilege rules; `docs/domain.md`, `docs/design.md`, `docs/research/nlspec-spec.md`; root/backend/tooling instructions and planning framework | Owner requirements, proposed-amendment status, vocabulary and supporting design doctrine. Supporting documents do not override behavioral owners. No presentation redesign or domain vocabulary change is needed. |
| E17 | Prior planning turn: `make help`, `make task-guide OWNER=module.networkflow`, `make task-guide OWNER=module.imports`, all exit 0 on the current baseline; `.markdownlint-cli2.jsonc`, generated Make entry and Markdown runner inspection | Public focused selectors and current canonical tracker glob. Discovery success is not test success; only Markdown lint is executed by this document-update step. |
| E18 | DOC-02 command, preservation and completion log in §20 | Documentation validation and final change-boundary evidence; no implementation acceptance |

### Contract decisions and adoption posture

The NF 8.0.0 header remains **proposed**, with unchanged requirements retaining adopted 7.0.0 baseline authority. DOC-02 records recommendations and does not adopt or edit any owner document. Owner-specific requirements govern behavior; machine contracts and tests remain projections. Human review establishes fidelity. No unresolved contradiction between different primary owners was established; the raw-cell diagnostic extension below is an explicit clarification/adoption gate, not a claim that code or a test can resolve owner policy.

| Contract ID / owner | Required or proposed decision | Disposition and dependent work |
| --- | --- | --- |
| C18 / NF §29, NF-REQ-210–212; Core 00 precedence and active contract publication | Keep private semantic validation in Network Flow; retain the active major-8 contract set and remove only demonstrably obsolete packaging obligations | Structural cleanup; F18/F19, S10. F10's historical decision to leave inert v7/v1 files is superseded only for those files and their exclusive projections. |
| C19 / Core 01 analytical binding Table 17.2-F and unit commit contract; NF §10 source capability | Imports owns opaque source access and generic orchestration; Network Flow receives bytes and source identity without Imports SQL or storage details | Preserve owner boundary and public capability/wire shape; F20, S12. Source integrity must be established before a successful preview/apply result. |
| C20 / NF §9.2 Table 9-B, NF-REQ-073/074 and §10 Table 10-A2 | Preview interprets the first `min(50, record_count)` complete data records. It may drain raw bytes for identity verification but must not report CSV/encoding errors confined to later records | Correct current whole-source normalization behavior in S13. Raw source read or digest failure is an integrity failure, not a later-record CSV diagnostic. |
| C21 / NF §12.4, NF-REQ-111–114, Tables 12-E/12-F; §20 and §21 | Enforce raw-cell limits for mapped, unmapped and ignored cells; preserve closed error tokens, exact scalar counts, attribution, hashes, redaction and ordered retention | Existing limit requirement plus proposed diagnostic-production clarification below; F23, S09/S13. RB-004 gates the new clarification. |
| C22 / NF §8.2, NF-REQ-057–062b; Core Imports atomic unit and cancellation contract | Validate before allocating a domain table. Reopen the source for bounded insertion into the existing borrowed transaction; source/summary mismatch, cancellation before commit or insertion failure rolls back all effects | F21, S14. Keep exactly one successful table and existing outcome/replay/commit ownership; no `creating`/`failed` table state. |
| C23 / NF indicator-link ownership and §29; NF-REQ-058; Core authorization/disclosure and transaction ownership | Read accepted rows by incident/table/row identity, retain supplied-reference checks and revalidate selected sources at commit | Structural retrieval change; F22, S11. Authorization and error precedence are characterization obligations, not permission to disclose hidden state. |
| C24 / Core 04 database role/privilege contract; NF §8.2 | Runtime has no DDL or temporary-table privilege. Do not add temporary tables, filesystem spooling, a durable staging format, or broader database grants | S12/S14 must work within current source/storage capabilities. No schema migration or state-version bump is planned. |

The proposed C21 clarification is: field-count mismatch emits its single row-level diagnostic first and skips target validation; otherwise emit one `cell_scalar_limit_exceeded` diagnostic per oversized source column, skip target validation for those cells, and validate remaining targets normally. Attribute a target only where the approved mapping supplies one; preserve exact decoded-scalar hashes and null safe samples. Run cross-field checks only when their operands validated successfully. Final diagnostic order remains NF-REQ-113. The conservative preview bound becomes `50 * (max_columns_per_csv + mapped_target_count + 1)`, with all terms bounded by the owner contracts. This replaces the insufficient target-field-only bound for the new raw-cell diagnostics; it neither changes diagnostic-ID bytes nor invents a parallel resource schema. Count and hash oversized scalars incrementally without retaining their full text. Adoption of this clarification is required before relying on it in S13.

Keep historical migration-ledger definitions because they authenticate retained ledger rows; keep owner-local harness controls with ordinary-build isolation/runtime scan inclusion; keep live Graph iterators, current request/resource schemas and stable test IDs. Historical numeric suffixes alone justify no removal or renaming. Graph, Reporting, Recovery, Indicators and Collaboration retain their distinct responsibilities. No broader compatibility decoder, migration route, generic validation framework or speculative module split is proposed.

## 17. DOC-02 finding dispositions

All F18–F23 statuses are **TODO**. Each is `must_fix` for its named workstream's exit; that classification does not authorize unrelated cleanup. No finding below is marked fixed by this document update. F01–F17 remain resolved as recorded in §14; F18 explicitly supersedes only F10's retention of obsolete packaged contracts.

| Finding / evidence / owner clauses | Remediation and affected areas | Rationale and expected long-term benefit | Compatibility or migration impact | Risk if unresolved | Validation / workstream |
| --- | --- | --- | --- | --- | --- |
| F18 — Superseded contracts remain packaged; E14; C18 | Remove inactive `frontend-entrypoints.v7.json` and `operation-policy.v1.json`, their exclusive schema attachments/schemas, and generated copies through authored inputs and Make. Contracts, tooling, generated outputs and tests. | One active publication set reduces misleading declarations, validator/package weight and unnecessary compatibility maintenance. | Active major-8/public resource contracts remain unchanged. Preserve history in Git and this tracker; do not remove historical migrations or live ledger facts. | Old declarations continue to look supported and impose validation/packaging obligations despite having no active dispatch role. | V10/V16; S10. Consumer closure, active resolution, schema/generated checks, and absence of obsolete packaged copies. |
| F19 — Test-only forwarding helper in production; E12/E02-131/132; C18 | Remove `validatePreviewDiagnosticShape`; retarget its assertions to the real private decoder and public preview admission. Implementation and tests. | Eliminates a redundant production surface and keeps tests coupled to meaningful owner behavior. | No wire, storage or error-translation change. Do not remove the real strict decoder or weaken framing/null/member coverage. | A convenience test continues to sustain unnecessary production code and obscures the actual validation boundary. | V10/V14; S10. Valid/invalid shape cases and public safe failure translation remain covered without the helper. |
| F20 — Source-stream access materializes whole blobs; E13; C19/C24 | Read bounded byte chunks behind Imports' existing port; split metadata-only capability/precondition reads from content retrieval. Implementation, port documentation and tests. | Bounds application buffering at the actual upstream boundary and removes repeated blob allocation for metadata checks. | Public source capability/wire shape and existing storage remain; private reader integrity/closure obligations become explicit. No data conversion, new storage backend or privilege change. | Network Flow streaming still inherits a whole-file allocation, increased concurrent memory use and repeated unnecessary hashing/loading. | V12/V15; S12. Digest/size integrity, cancellation, close/error cleanup, isolated snapshots and bounded read sizes. |
| F21 — Apply retains complete parsed and accepted collections; E12/E13; C22/C24 | Replace full `Records`/accepted `Rows` retention with a summary-producing validation pass and a reopened-source insertion pass. Network Flow implementation, private store interface and integration tests. | Application memory depends on configured fields/diagnostics and bounded insertion, rather than total source/accepted row count. Retains one understandable atomic commit owner. | Preserve row/diagnostic/provenance identities, accepted order, counts, audit, replay and cancellation semantics. Extra I/O is an explicit tradeoff; no durable format or schema change. | Large imports retain several representations, enforce limits late and create avoidable memory pressure during transactions. | V13/V15; S14. Two-pass source/summary agreement, bounded buffering, complete rollback on late failure and correct committed-success recovery. |
| F22 — One-row linking reads a whole table; E12; C23 | Add private incident/table/row lookup and remove production `ListRows` after migrating its test-only consumers. Implementation and tests. | Narrows the store surface and makes lookup work proportional to selected rows instead of the enclosing table. | Same active-table, authorization, disclosure, supplied-reference and commit-time checks; no route or resource shape change. Live graph streaming remains. | A small authorized request can allocate and scan a large table; repeated bounded selectors multiply this cost. | V11; S11. Positive/missing/wrong-scope/deleted/stale-reference cases and deletion race retain their contract, with no full-table enumeration. |
| F23 — Parser/limits drift from owner behavior; E12/E16; C20/C21 | Use one private incremental CSV decoder for preview/apply; correct dialect, preview-tail handling, raw-cell limits, count timing and limit details. Owner clarification, implementation, tests and documentation. | Predictable bounded parsing and deterministic provenance/diagnostics; limits become effective before full collections are retained. | Correct erroneous behavior under existing owner definitions; proposed C21 clarification needs review/adoption. Existing state-5 tables are not rewritten; corrected interpretation requires original-source re-import. | Preview can fail on forbidden later-record errors; blank/CR-delimited input can be interpreted incorrectly; oversized cells and late limits defeat bounded processing. | V13/V14/V15; S09/S13. Independent dialect/provenance vectors, split-byte cases, exact bounds, privacy, preview publication and apply parity. |

## 18. DOC-02 workstreams, sequencing and readiness

Implementation sequence: **S09 -> S10 -> S11 -> S12 -> S13 -> S14 -> S15**. The graph is acyclic and each workstream has one immediate predecessor. All workstream statuses are **TODO**. None is implementation-complete because its plan is written. Later implementation authorization and baseline revalidation are common entry conditions; readiness below describes technical prerequisites separately from that authorization.

| Workstream | Depends on / finding | Current readiness | Principal risk | Binary exit and successor |
| --- | --- | --- | --- | --- |
| S09 Contract closure and characterization | DOC-02 handoff; F18–F23 | ready to start under later authorization | Freezing defects as intended behavior or treating a proposal as adopted | Owner decisions/classification, registered characterization and explicit review gate recorded; V10–V15 entry evidence scoped; checkpoint before S10 |
| S10 Retire obsolete surfaces | S09; F18/F19 | blocked by S09 and missing characterization/consumer closure, RB-005 | An overlooked consumer depends on a removed artifact | V10/V16 pass; obsolete copies absent and active contracts resolve; checkpoint before S11 |
| S11 Narrow accepted-row retrieval | S10; F22 | blocked by predecessor and V11 baseline, RB-005 | Disclosure or missing-row precedence changes | Exact lookup with retained transaction checks and passing V11; checkpoint before S12 |
| S12 Bound Imports source reading | S11; F20 | blocked by predecessor and V12 source-port characterization, RB-005 | Connection leaks, integrity translation drift or inefficient database chunking | V12 and affected Imports consumers pass; application bound and database cost distinguished; checkpoint before S13 |
| S13 Incremental CSV and bounded preview | S12; F23 | blocked by predecessor, RB-004 and V13/V14 characterization | Parser/provenance changes or invalid preview bounds | C21 clarification gate satisfied; V13/V14 pass, V15 apply regression subset passes; checkpoint before S14 |
| S14 Streaming atomic apply | S13; F21 | blocked by predecessor and V15 entry evidence, RB-005 | Extra I/O/transaction duration or partial effects after late failure | V15 passes with bounded storage insertion, source/summary agreement and atomic outcomes; checkpoint before S15 |
| S15 Focused validation and handoff | S14; F18–F23 | blocked by all preceding exits | Reusing historical checks as final acceptance | Finalized source, missing focused obligations completed, findings/evidence/compatibility reconciled; no remediation blocker remains |

### S09 — Contract closure and characterization

Re-run applicable owner guidance and revalidate the exact named sources before editing. Record C18–C24 dispositions in the behavioral owners only where clarification is necessary; keep amendments proposed until the owner-review/adoption gate is satisfied. In particular, resolve RB-004's raw-cell diagnostic production and preview bound without changing the existing ID algorithm or resource shape. Keep Core Imports' source/transaction responsibility and Core 04 privilege restrictions explicit.

Register focused, clearly named behavioral selectors through authored verification inputs and Make generation. Capture passing entry evidence for retained valid preview/result admission, source integrity/error translation, row lookup disclosure and atomic import/replay/cancellation behavior. Add correction-specific cases with their correcting slices; do not encode erroneous permissiveness, whole-file allocation, or the wrong CSV dialect as permanent expectations. Test public contracts and meaningful resource invariants, not arbitrary private error text.

Exit requires consistent owner decisions, an explicit proposed-versus-adopted distinction, registered characterization obligations, and a recorded adoption gate for dependent changes. S09 may record an outstanding adoption gate while completing independent characterization; S13 remains blocked until RB-004 closes. Failure evidence and the S09 checkpoint precede S10.

### S10 — Retire obsolete surfaces

Establish references for the two obsolete contract files, their matching schemas, schema attachments, generators, generated packaging and active Go/browser consumers. Remove the authored obsolete artifacts and exclusive attachments/schemas; regenerate, never hand-edit generated copies. Keep entrypoint v8, operation-policy v2, current request/resource IDs, stable test rows, historical SQL and ledger facts. Do not create aliases or dual-major dispatch.

Remove only the forwarding `validatePreviewDiagnosticShape` production wrapper; move its assertions to `decodePreviewDiagnostic` or the public facade as appropriate. Preserve strict framing, forbidden-null/member checks, valid roundtrips and safe Imports failure translation. V10/V16 cover consumer closure, compilation, publication, schema/generated consistency and boundary behavior. Any discovered live consumer must be addressed within the active owner contract before deletion; absence from one lexical search is insufficient.

Exit is a smaller active artifact/production surface with no unresolved consumer and passing focused checks. Rollback is the coordinated authored-input and generated-output source change; no data rollback is required. Record the checkpoint before S11.

### S11 — Narrow accepted-row retrieval

Replace full-table retrieval with a private exact lookup constrained by incident ID, table ID and row ID. Preserve authorization before state disclosure, active-table admission, supplied source-row-number/mapping-fingerprint checks, bounded selector limits and commit-time source revalidation. Keep the existing indicator-link application and transaction owners.

Migrate test callers that used bulk retrieval for convenience to scoped assertions or test-only helpers, then remove the now-unused production `ListRows`. Do not replace live graph streaming paths or move SQL into a generic module. V11 covers success, wrong incident/table, absent row, soft-deleted table, stale supplied reference and deletion races, and demonstrates that unrelated rows are not retrieved. Exit and checkpoint precede S12; rollback is source-only and preserves stored data.

### S12 — Bound Imports source reading

Implement a context-aware reader over bounded database byte chunks from one consistent query snapshot, inside Imports. Metadata and byte chunks must refer to the same admitted source identity; separate capability/precondition lookup must not select or hash the full blob. Keep the existing public capability and wire contracts, SQL ownership and storage format. Use fixed internal chunk bounds rather than a new public tuning surface.

Verify exact byte count and digest before successful consumption completes. EOF, read error, cancellation and explicit close must release query/connection resources; close must be safe after EOF and must not falsely certify an incompletely consumed source. Preview will drain raw bytes before success; failed parsing may close early without treating that source as verified. Distinguish safe source-integrity failures from CSV semantics through existing owner error boundaries without exposing raw bytes or causes.

V12 covers fragmented/empty reads, size/digest mismatch, early close, repeated close, cancellation, concurrent snapshots and cleanup. Include focused real Imports consumers so moving integrity checking into consumption cannot silently weaken admission. Demonstrate bounded application buffering and inspect database-side chunk cost separately; a bounded Go reader is not evidence of bounded PostgreSQL memory or acceptable repeated blob decompression. No broad benchmark certification is claimed. Exit/checkpoint precedes S13. No storage migration or privilege expansion is permitted by this design.

### S13 — Incremental CSV and complete bounded previews

Use one Network Flow-private decoder for both modes. Implement Table 9-B rather than assuming the standard CSV reader has the same dialect: CR/LF/CRLF record delimiters, blank logical records, quoted line breaks, opening/closing/escaped quotes, BOM placement and incremental UTF-8 validation. Preserve exact decoded field content, logical record numbering, header descriptors and source/provenance digests. Handle boundaries inside UTF-8 sequences, BOMs, quotes and delimiters.

Enforce column/header/raw-cell/row limits at owner-defined stages and use the exact closed error details. For oversized fields, retain bounded content only, continue counting/hash computation as required for the exact scalar diagnostic, and reject the row without persisting the oversized value. A malformed unterminated record while seeking the next logical record retains the Table 9-B failure precedence. Stop count processing once one excess item establishes the count limit; do not scan for an unnecessary total.

After RB-004 closes, implement C21's diagnostic production, attribution and revised bound through the existing shared validator. Preview keeps only bounded records, descriptors, samples and diagnostics; it stops semantic parsing after the 50th complete data record. Drain the remainder only as raw bytes for source identity. Later malformed CSV, misplaced BOM or invalid UTF-8 must not become preview diagnostics, while read/identity failures must still prevent success. Preview remains side-effect free and always reports complete in-scope diagnostics with `diagnostics_truncated=false`.

V13/V14 use independent expected dialect/digest vectors, chunk boundaries, malformed input, 50/51 cases, limits at/below/above the bound, oversized-cell privacy and valid public-facade roundtrips. Preserve the existing apply path's contracts while it consumes the shared decoder until S14 removes its remaining materialization; do not claim end-to-end apply memory completion at S13. Checkpoint before S14.

### S14 — Streaming atomic apply

Replace whole-file `Records`/accepted `Rows` collections and `preparedImportApply`'s retained parser payload with a private validation summary. First pass parses and validates the source, retaining accepted/rejected counts, descriptor/source/mapping identities, the configured diagnostic prefix, actual omission state and the bounded all-rejected diagnostic sample. No domain table is allocated until this pass succeeds and admits at least one row. Accepted-row limits are enforced during validation, not after a full collection has been built.

Second pass reopens the same capability and streams accepted rows in source order through the private storage adapter into Imports' existing borrowed transaction. Keep insertion buffering bounded independently of total rows; do not collect input merely to sort rows already produced in owner order. Recompute source identity, mapping/descriptor agreement and summary agreement before returning success. The same approved mapping and pure validation path must govern both passes. A changed source, read/insert failure, summary mismatch or cancellation before commit must roll back table, rows, diagnostics, audits and outcome publication together.

Retain Imports' final commit/cancellation barrier and replay ownership. A proven committed success wins over later cancellation; no Network Flow facade may independently commit the borrowed transaction. Preserve count/diagnostic limits, IDs, provenance and audit behavior. Use no temporary tables, filesystem spool, durable staging format or broader database privileges. Additional reading/validation and time spent in the transaction are explicit tradeoffs for bounded application memory and simpler lifecycle ownership.

V15 covers zero/exact/overflow limits, partial/all-rejected input, multiple diagnostics per row, source changes between passes, source read errors, insertion failure, cancellation before commit, committed-result recovery and exact replay. Resource checks must demonstrate that retained row collections do not grow with source length, separately from the bounded diagnostic allowance and database costs. Exit/checkpoint precedes S15. Reverting source does not justify modifying already committed state or changing the release's admission boundary.

### S15 — Focused validation and handoff completion

Audit actual obligation coverage rather than rerunning entire suites by name. After focused slice checks, run repository-required `make agent-finalize`, review any generated changes, freeze source and execute only the remaining necessary focused selections. Reassess affected obligations after any repair. **Do not run full CI.** Add no unrelated visual, benchmark, release or full-owner qualification.

Reconcile F18–F23, RB-004/RB-005, all slice statuses, exact commands/owning exits/run roots, environment, finalized source identity, generated changes and current compatibility notes. Every deletion needs consumer evidence; bounded application-resource claims need direct evidence; database-side cost and validation limits must be stated. Historical evidence supports diagnosis but cannot substitute for final compatible acceptance.

Exit requires all preceding slices complete, required focused validation passing, no remediation blocker remaining, and a handoff that separates implementation completion from human adoption and operational deployment. Preserve the §15 operational original-source and untouched-environment rollback instructions, with the current release boundary restated. No destructive reset, deployment or data conversion is an S15 task.

## 19. DOC-02 validation obligations and evidence routing

Only V18 is executed for this document update. V10–V17 are future implementation obligations, with **no current pass claimed**. Run public Make targets from the repository root. Re-run owner guidance when entering an affected slice; select exact rows through `make target-plan` and `make test-slice` instead of running complete owners by default. New cases must be registered through authored verification inputs and Make generation. Tests, generators, runtime metadata, conformance and release evidence must never read, stat or hash Markdown.

| Validation ID | Required cases / evidence | Selection and owning workstreams |
| --- | --- | --- |
| V10 Active surfaces and dead-helper removal | Active major-8 resolution, obsolete artifact consumer closure and generated removal; real strict decoder and safe preview failure translation still exercised | S09/S10; NF preview/admission and binding rows, affected contract publication rows; source/schema/generated checks under V16 |
| V11 Exact accepted-row lookup | Positive row, wrong incident/table, missing row, soft-deleted table, stale supplied reference, deletion race; no unrelated row enumeration | S09/S11; focused NF indicator-link, storage and transaction consumers |
| V12 Imports source-port lifecycle | Bounded chunking, metadata-only lookup, digest and byte-size mismatch, fragmented/empty input, close/EOF/error/cancellation release, concurrent source snapshot isolation and existing consumer error translation | S09/S12; focused Imports unit/service tests plus analytical-owner consumers; distinguish application memory from database costs |
| V13 Incremental parser, limits and provenance | Independent CR/LF/CRLF, blank/quoted/escaped-record and digest vectors; UTF-8/BOM/delimiter splits; malformed framing; header/column/raw-cell/row/accepted-row bounds and exact error details | S09/S13/S14; focused NF parser/owner-error tests, with correction cases added alongside implementation |
| V14 Complete bounded preview and diagnostic admission | 50/51 records, later-record errors suppressed, raw integrity failures still rejected, complete diagnostics including raw-cell failures, correct null attribution/hash/redaction, 0/N/N+1 apply retention parity, both Imports preview publication and approval consumers | S09/S10/S13; dedicated NF preview admission and semantic guard rows plus focused Imports consumers |
| V15 Atomic streaming apply | Valid/partial/all-rejected imports, counts independent of diagnostic multiplicity, two-pass identity/summary agreement, source mutation/read failure/insertion failure/cancellation rollback, committed-success recovery, replay and bounded row buffering | S09/S12/S14; focused NF and Imports service-backed atomic-unit/outcome tests; no temporary-table/DDL privilege requirement |
| V16 Generated/publication/boundary consistency | Authored removals regenerate cleanly, active Go/browser resolution remains, no compatibility aliases, semantic source entrypoints retain dependency direction | S10 and later affected slices: `make generate`, `make source-check CHECK=schema`, `make source-check CHECK=generated`, `make lint CHECK=backend-boundary`; generated files never hand-edited |
| V17 Final focused acceptance | Finalization changes reviewed, source frozen, missing focused obligations selected, exact failure/repair evidence and cleanup outcomes reconciled | S15: `make agent-finalize`, then only required focused checks identified by current plans; `make ci` excluded by user instruction |
| V18 Documentation checkpoint | Tracker selected by canonical module-tracker glob; structural Markdown lint passes; only admitted tracked file changed | DOC-02 and each documentation checkpoint: `make lint-markdown`; current command/root/result recorded in §20 |

Existing selector anchors inspected on this baseline include `module.networkflow.unit.import_preview_admission_f0dc5f2d8b`, `module.networkflow.unit.query_authoring_admission`, `module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99`, `module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`, `module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`, `module.imports.integration.uncommitted_unit_cancellation`, and `module.imports.integration.unit_commit_crash_recovery_108cd054a6`. They are candidate evidence routes, not proof that every new case is already covered. The authored family inputs are `tools/test_families/module.networkflow.json` and `tools/test_families/module.imports.json`; no row is invented or changed during DOC-02.

For future failures record the exact target/selection, owning exit, run root or summary artifact, and relationship to the change **before** repair or rerun. A failed exit leaves its slice incomplete and blocks successors. Broaden only for an identified missing obligation. No standalone migration check is needed merely for this planned source-only iteration; if an unexpected required schema change emerges, record and resolve the scope/owner decision before proceeding instead of silently adding a migration. Human owner review remains independent of all automated results.

## 20. DOC-02 blockers, completion and planning handoff

### Gates and authorization

| Gate | Status | Dependent work / closure evidence |
| --- | --- | --- |
| RB-004 Raw-cell diagnostic-production clarification and revised preview bound | BLOCKED — proposed owner decision awaits review/adoption | S09 documents and presents C21; S13 and dependent S14/S15 cannot rely on the new rule until the owner gate is satisfied. Existing unchanged owner requirements remain authoritative. |
| RB-005 Missing executed characterization for the new structural/streaming seams | BLOCKED — tests were inspected, not run for this iteration | S09 establishes and records baseline obligations; each dependent S10–S14 slice consumes its exact compatible evidence. Code similarity and prior S08 passes do not close this gate. |
| Major-8/state-5 owner adoption and coordinated deployment | DEFERRED — human/operational release gates | Carry forward §15's proposed-amendment review and coordinated server/browser publication. Implementation completion alone cannot close these gates. |
| Future implementation authorization | Not granted by DOC-02 | This invocation updates the tracker only. S09–S15 require a later implementation task; the authorization column is not a work-status claim. |

Planning completeness is **complete for the declared bounded iteration**: V18 passed and the preservation audit confirms the tracker is the sole changed tracked file. S09 is ready to start under later authorization and baseline revalidation; subsequent readiness is blocked by the named predecessors, characterization and adoption gates. These known future prerequisites do not imply that implementation has started. There is no claim of exhaustive module conformance, production deployment readiness, full-owner acceptance or a new full-CI pass.

### Planning session and command log

| Session / action | Evidence and result | Next action |
| --- | --- | --- |
| DOC-02 read-only planning turn, 2026-10-29 | Current baseline inspected; user selected cleanup plus hardening and included Imports reader changes. `make help` and both `make task-guide OWNER=module.networkflow` / `OWNER=module.imports` returned 0. No files changed or product tests ran in that turn. | Authorized tracker-only update |
| DOC-02 update admission, 2026-10-29 | E11: baseline unchanged, clean `main`, safe regular tracker identity; original inventory comparison yields 129 retained paths plus E02-130–132. Root/backend/tooling guidance, skills and relevant owners consulted. | Write current scope, findings and workstreams while retaining historical evidence |
| DOC-02 tracker checkpoint | §§1/14/15 distinguish completed history; §§16–20 record current inventory, F18–F23, C18–C24, S09–S15, V10–V18 and gates. All implementation statuses remain TODO. | V18 and final preservation audit |
| DOC-02 V18 | PASS, owning exit 0: `CARTULARY_TEST_RUN_ID=nf-doc02-plan-markdown-20261029-01 make lint-markdown` from repository root. Emitted root `.cartulary/test-results/nf-doc02-plan-markdown-20261029-01`; summary `adhoc/lint-markdown/tool-run-summary.json` reports pass, zero failures. The module-tracker glob selects this artifact. | Record result and perform final post-log documentation check |
| DOC-02 preservation audit | `git --no-optional-locks diff --check` returned 0. Diff/status show only this tracker changed; no staged changes or non-ignored untracked files. Index SHA-256 still equals E11's starting value. Historical inventory and evidence retained; no production refactor performed. | Later authorized S09 after baseline revalidation |

The completion-log edit is followed by the same documentation-only target with fresh ID: `CARTULARY_TEST_RUN_ID=nf-doc02-plan-markdown-20261029-02 make lint-markdown`. Its exact evidence location is `.cartulary/test-results/nf-doc02-plan-markdown-20261029-02/adhoc/lint-markdown/tool-run-summary.json`; the owning terminal result is reported with the handoff. This final confirmation adds no product verification or finalization.

No product tests, formatter, generator, `agent-finalize`, full CI, visual suite, benchmark or deployment is run for this document-only update. These are intentionally outside the current execution scope, not skipped implementation acceptance checks. Ordinary ignored lint cache and retained-run artifacts are permitted. No source, tests, contracts, migrations, generated files or dependencies may be changed by DOC-02.

### Binary document-completion criteria

| Criterion | Status / evidence |
| --- | --- |
| Current clean baseline and 132-file universe recorded without presenting targeted review as exhaustive | DONE; E11, §16 |
| Original inventory, F01–F17, S00–S08, failures and operational handoff preserved as history | DONE; §§1–15 retained with explicit historical labels |
| F18–F23 include owners, remediation areas, rationale/benefit, compatibility, unresolved risk and validation | DONE; §§16–17 |
| S09–S15 form an acyclic sequence with separate risks, readiness, exits and tracker checkpoints | DONE; §18 |
| Proposed amendments, human adoption, technical readiness and implementation authorization separated | DONE; C21/RB-004 and this section |
| Focused validation and no-full-CI instruction carried forward; no Markdown-dependent product verification introduced | DONE; §19 |
| Canonical Markdown lint and final only-tracker preservation audit | DONE; V18/E18, owning lint exit 0 and unchanged index/only-tracker diff |

The later implementation handoff must explicitly state that older Network Flow data was **neither converted nor deleted**, that original CSV sources and mapping decisions are required for re-import, and that matching major-8 server/browser assets with state 5 are one coordinated release boundary. Correcting source interpretation does not rewrite existing state-5 tables. Rollback uses the untouched previous environment with its matching database, object store, keys and assets; do not downgrade a database after state-5 writes, remove metadata or delete data to force admission. Human adoption and deployment remain separate from engineering completion.

## 21. IMP-03 controlling implementation plan and execution

### Decisions and finding delta

The user approved the revised plan for implementation, including Imports storage migrations and analytical source authority based on original bytes and approved owner mapping. Explicit human owner adoption precedes dependent code. Keep major 8/state 5, focused acceptance and S15 completion; no full CI, deployment, destructive reset or historical Network Flow conversion is part of this effort.

F18–F23 retain §17's remediation, rationale, risks and validation, with these amendments:

- F20: migrate `import_source_streams.source_bytes` to PostgreSQL `EXTERNAL` storage and physically reserialize existing values without changing logical bytes. Use private fixed 64 KiB reads, metadata-only capability queries, owned consistent snapshots for ordinary reads and transaction-aware borrowed reads for apply. This replaces the no-migration/no-internal-port-change assumptions in C19/C24. Runtime receives no additional privilege.
- F21: resource acceptance includes the real Imports orchestration path, not merely the Network Flow facade.
- F23/C21: obtain adoption for raw-cell diagnostic production and the complete preview bound `50 * (max_columns_per_csv + mapped_target_count + 1)` before dependent correction.
- F24 (new, must_fix): Imports' `getApplyUnits`, `insertApplyUnitPlansTx` and `lockApplyUnitTx` fetch/retain generic `source_rows_json` before analytical application. Use ordered scheduling metadata, target-specific admission and lazy generic-row loading. Analytical content authority is its original source stream/digest and approved owner mapping/fingerprint plus target binding. Migrate analytical plan `source_rows_sha256` to null; retain the required digest for generic targets. Areas: specification, implementation, migrations, tests and documentation. Benefit: eliminate hidden full-source allocations and duplicate interpretation. Unresolved risk: F20/F21 cannot establish bounded analytical apply, and analytical content remains coupled to generic parsing. Compatibility: no public response change; migrate internal plan metadata with workers quiesced, preserve outcomes/replay identities, and prohibit old binaries on the migrated schema. V19 requires no analytical generic-row reads, source/mapping failure coverage, generic regression, migrated-plan behavior and cancellation/recovery.

### Workstreams and binary exits

| Workstream | Status | Depends on | Implementation / principal risk | Binary exit |
| --- | --- | --- | --- | --- |
| S09 Specification closure and characterization | DONE | Baseline and this checkpoint | Candidate explicitly adopted by the user on 2026-10-29; focused characterization passed. | Explicit owner adoption; baseline evidence and later obligations mapped; validated tracker checkpoint. |
| S10 Retire obsolete surfaces | DONE | S09 | Remove inactive v7/v1 publication, exclusive schemas/attachments and diagnostic forwarding wrapper after consumer closure; regenerate. Risk: overlooked live consumer. | V10/V16 pass; active publication intact; all deletions evidenced. |
| S11 Narrow accepted-row retrieval | DONE | S10 | Exact incident/table/row lookup and scoped tests; retain commit revalidation. Risk: disclosure or deletion-race drift. | V11 passes; production bulk retrieval removed; graph iteration intact. |
| S12-M Source storage migration | DONE | S11 | Imports-owned EXTERNAL-storage migration with byte-preserving rewrite and regenerated ownership/projections. Risk: locks, disk/WAL use and surviving compression. | Empty/populated migration, byte/digest equality, uncompressed storage, region copying, Recovery compatibility and migration drift pass. |
| S12 Bounded source readers | DONE | S12-M | Fixed 64 KiB reads; metadata lookup; owned snapshot and `OpenSourceStreamTx`; digest/size validation and complete resource lifecycle. Risk: leaked connections, lock order, false EOF or changed error translation. | V12 passes, including small-pool/concurrent/cancellation cases and database cost evidence. |
| S12-I Analytical admission and scheduling | DONE | S12 | Metadata-only scheduling; separate analytical/generic admission and unit reads; nullable analytical plan row digest migration. Risk: weaker admission or replay/mixed-version failure. | V19 passes; analytical SQL does not fetch/hash generic rows; generic and migration regressions pass. |
| S13 Incremental CSV and complete preview | DONE | S12-I; adopted C21 | Shared owner-dialect decoder, bounded fields, exact scalar/hash accounting, full in-scope diagnostics and raw-only tail drain. Risk: boundaries, provenance, precedence and privacy. | V13/V14 plus affected apply regression pass; no end-to-end apply claim yet. |
| S14 Streaming atomic apply | DONE | S13 | Summary validation followed by protected-source, one-row-at-a-time insertion in borrowed transaction. Risk: longer transactions and partial late-failure effects. | V15 passes on real Imports path; source/mapping/descriptor/summary agreement and atomic outcomes demonstrated. |
| S15 Focused validation and handoff completion | DONE | S14 | Finalize, review generation, freeze source and fill focused acceptance gaps. Risk: stale evidence presented as current acceptance. | All findings/workstreams reconciled, required checks pass, no engineering blocker, complete operational handoff. |

S09's diagnostic rule: field-count mismatch emits one row-level diagnostic and skips targets; otherwise emit one raw-cell-limit diagnostic per oversized column (mapped, unmapped or ignored), skip those cells' target validation and validate remaining targets normally. Attribute targets only through approved mappings. Cross-field checks require valid operands. Preserve diagnostic IDs, exact decoded scalar hashes, null unsafe samples and owner ordering.

S12 ordinary readers own a read-only consistent snapshot; borrowed readers hold source protection but never commit/roll back the caller transaction. Each bounded query finishes before returning data, allowing writes on the same connection. Successful EOF verifies size/digest. Error, cancellation, EOF and close release owned resources; early close is not integrity success. Preserve existing safe error boundaries.

S13 implements CR/LF/CRLF, blank records, literal/quoted/escaped/closing quotes, quoted line breaks, incremental UTF-8/BOM and logical numbering. Excess scalar text is discarded while counting/hashing continues. Stop semantic preview parsing after record 50; drain only raw bytes for integrity. Later CSV/encoding defects are outside preview diagnostics; integrity/read failures still prevent success.

S14 retains only counts, identities, descriptors/mapping agreement, configured diagnostic prefix and bounded all-rejected sample in validation. Enforce accepted limits while validating; allocate no table before validation succeeds with an accepted row. Reopen protected source and insert in canonical source order without collecting/sorting all rows or querying within COPY callbacks. Verify both-pass identity/summary agreement before success. All precommit failures roll back table/rows/diagnostics/audit/outcome; Imports owns commit, cancellation barrier and committed-success recovery.

### Validation, evidence and handoff requirements

V10–V18 remain defined in §19; V12 additionally includes physical storage migration and transaction-aware readers; V19 is the F24 admission/scheduling/migration obligation above. Discover exact current selections via `make task-guide` and `make target-plan`, then run `make test-slice` with fresh run IDs. Register tests in authored inputs and regenerate via Make. Required affected checks include schema/generated consistency, backend boundaries, Go checks and `make migration-drift`. Run `make lint-markdown` at checkpoints. Run `make agent-finalize`, review generated changes and freeze source before final focused acceptance. No full CI by default.

Every checkpoint records source identity, exact commands/selections, owning exits, run roots, cleanup, failures/repairs and skipped checks/reasons. Failure leaves its workstream incomplete. Separate application-memory evidence from database cost observations. No product evidence consumes Markdown.

Operational handoff requires quiesced import workers and a recoverable database snapshot before new migrations; publish matching server/browser assets and prohibit mixed workers around the changed plan schema. Existing Network Flow tables are neither reinterpreted nor deleted. Physical source migration preserves bytes; plan migration removes obsolete analytical metadata. Corrected interpretation requires original CSV and mappings for re-import. Upload/discovery and Recovery-wide memory are outside this preview/apply resource claim. Rollback uses the previous matching environment snapshot, never downgrade/delete-to-admit after new writes. Deployment remains separate from S15 engineering completion.

### Execution log

| Checkpoint | Changes and evidence | Result / next gate |
| --- | --- | --- |
| IMP-03 initial tracker update, 2026-10-29 | Revalidated `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`; preserved pre-existing tracker changes and all history. Added F24, S12-M/S12-I, revised source/storage decisions and explicit adoption gate. | S09 IN_PROGRESS; no product changes or validation passes claimed. |
| S09 owner candidate | Prepared Network Flow §9.2/§12.4 and Core 01 REQ-01-620b amendments; Core 00 §1.3 names the complete coordinated adoption candidate and companions. Added fragmented-reader, exact-digest and source-read-error characterization under existing registered `TestImportPreviewAdmission`. | Reviewable candidate prepared; human adoption still pending. No dependent implementation started. |
| S09 discovery | `make help ALL=1`, `make task-guide OWNER=module.networkflow`, `make task-guide OWNER=module.imports`, and `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s09-unit-selection.json DETAIL=rows` / `INPUTS=/tmp/cartulary-nf-s09-imports-selection.json`, all exit 0. | Existing semantic rows cover baseline obligations; new correction cases accompany their slices. |
| S09 format and documentation | `CARTULARY_TEST_RUN_ID=nf-s09-format-20261029-01 make format` and `CARTULARY_TEST_RUN_ID=nf-s09-markdown-20261029-01 make lint-markdown`, owning exits 0. Roots `.cartulary/test-results/nf-s09-format-20261029-01` and `.cartulary/test-results/nf-s09-markdown-20261029-01`; Markdown summary `adhoc/lint-markdown/tool-run-summary.json`. | PASS; subsequent status/document edits receive a final checkpoint lint. |
| S09 Network Flow baseline | `CARTULARY_TEST_RUN_ID=nf-s09-baseline-nf-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99`; owning exit 0, root `.cartulary/test-results/nf-s09-baseline-nf-20261029-01`. | PASS; preview admission/privacy/integrity, semantic boundaries and current CSV baseline. |
| S09 Imports baseline | `CARTULARY_TEST_RUN_ID=nf-s09-baseline-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6`; owning exit 0, root `.cartulary/test-results/nf-s09-baseline-imports-20261029-01`. | PASS; real preview/approval, authorization, unit atomicity, safe translation, cancellation and committed-result recovery. |
| S09 lookup baseline | `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s09-lookup-selection.json DETAIL=rows`, exit 0; `CARTULARY_TEST_RUN_ID=nf-s09-baseline-lookup-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37`, owning exit 0, root `.cartulary/test-results/nf-s09-baseline-lookup-20261029-01`. | PASS; table lifecycle/limits, indicator admission, graph contributors and bounded-source regressions provide S11 entry evidence. |
| S09 evidence review | `make explain-run RESULTS_DIR=.cartulary/test-results/nf-s09-baseline-nf-20261029-01`, corresponding `baseline-imports` and `baseline-lookup` roots, all exit 0. NF and Imports source identity `sha256:7c4fc60cf4e93e53779c3214bd00d426295e932c8eac8214206f25eb6ed77ded`; NF 1/1 and Imports 5/5 work units, no failed/skipped/canceled units; their cleanup has 10/20 steps with zero failed or blocked. | Baseline passes are characterization, not implementation acceptance. Exact lookup run identity/cleanup remains in its named artifacts. |
| S09 adoption review gate | Prepared candidate linked to the user with an explicit adoption request after baseline Network Flow/Imports passes. `CARTULARY_TEST_RUN_ID=nf-s09-review-markdown-20261029-01 make lint-markdown` passed with owning exit 0; summary `.cartulary/test-results/nf-s09-review-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json`. | BLOCKED pending human adoption under the user's S09 gate. S10 and successors remain TODO. No production behavior, migration or generated artifact changed. |

S09 resumption: obtain an explicit adoption answer for Core 00 §1.3's prepared
candidate; if adopted, reconcile candidate/companion status headers and record
the human decision before completing S09 and beginning S10. If changes are
requested, revise the concrete candidate and its affected characterization.
Do not infer adoption from elapsed time, passing tests, or the implementation
request itself. No repeated product baseline run is required without a relevant
source change. The final checkpoint runs
`CARTULARY_TEST_RUN_ID=nf-s09-checkpoint-markdown-20261029-01 make lint-markdown`;
its summary is
`.cartulary/test-results/nf-s09-checkpoint-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json`.
Generation, migrations, `agent-finalize` and final acceptance are pending their
authorized later slices, not claimed complete or waived. Full CI is excluded by
the focused plan. All pre-existing tracker text remains preserved as history.


### S09 completion and adoption checkpoint

The user explicitly answered **“Adopt the prepared candidate and continue
implementation”** to the S09 adoption request on 2026-10-29. Core 00 §1.3 and
Network Flow/Extensions/Core companion headers now record adoption. RB-004 is
closed; RB-005 entry characterization is satisfied by the three named baseline
roots (same product source identity, owning exits 0). Lookup baseline has 5/5
work units and 18 cleanup steps, all passing. The historical pending-review
entries above are superseded by this explicit human decision, not erased.
S09 is DONE. The S09 checkpoint Markdown command passed with exit 0; the
adoption status update is checked by
`CARTULARY_TEST_RUN_ID=nf-s09-adoption-markdown-20261029-01 make lint-markdown`
before S10 starts. Deployment remains unauthorized and separate.


### S10 execution

S09 adoption Markdown checkpoint passed (owning exit 0). Consumer closure:
active contract index selects operation-policy v2; protocol entrypoint selects
v8; removed v7/v1 references were confined to their authored declarations,
exclusive schema attachments and generated packaging. Contract generation
walks the authored Network Flow artifact directory; JSON shape checking uses
those schema attachments. Historical ledger/state inputs are retained.
Removed the two obsolete declarations, their exclusive schemas/attachments and
the test-only diagnostic forwarding wrapper. Tests now call the real decoder;
public preview admission and safe Imports translation remain selected evidence.


S10 exit: V10/V16 PASS. All commands used explicit `CARTULARY_TEST_RUN_ID`:
`nf-s10-generate-20261029-01` / `make generate`;
`nf-s10-nf-20261029-01` / `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700`;
`nf-s10-translation-20261029-01` / `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`;
`nf-s10-schema-20261029-01` / `make source-check CHECK=schema`;
`nf-s10-generated-20261029-01` / `make source-check CHECK=generated`;
`nf-s10-boundary-20261029-01` / `make lint CHECK=backend-boundary`.
Each owning exit is 0; roots are `.cartulary/test-results/<run-id>`.
Generation removed obsolete packaged copies/validators and renumbered generated
AJV identifiers downstream of the deleted schemas; this is expected generated
churn. No authored or generated active reference remains to the removed schema
IDs or helper. F18/F19 are fixed. Source rollback is coordinated authored and
generated changes, with no data rollback. Checkpoint:
`CARTULARY_TEST_RUN_ID=nf-s10-checkpoint-markdown-20261029-01 make lint-markdown`.


### S11 execution

S10 checkpoint Markdown passed (exit 0). Replaced `ListRows` with an exact
incident/table/row lookup constrained to active tables; missing rows retain
selector errors and table deletion is rechecked on a missing read. Supplied
references and commit-time source validation remain with the existing link
application/transaction owner. Migrated all three test-only bulk consumers to
known row identities. Graph streaming remains unchanged. Scoped missing,
wrong-incident and wrong-table cases extend the existing registered store row.


V11 passed with `CARTULARY_TEST_RUN_ID=nf-s11-lookup-20261029-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37`, owning exit 0. Root: `.cartulary/test-results/nf-s11-lookup-20261029-02`; all 5 work units passed, cleanup 18 steps without failures. The first invocation (`nf-s11-lookup-20261029-01`, same command/selection) exited 2 with harness `infra/preflight_error`: a new test misspelled local variable `racing` as `tracing`, preventing compilation. Corrected that typo before the fresh successful invocation; no failed assertion was hidden. Both roots retain their source identities and diagnostics. Formatting runs `nf-s11-format-20261029-01` and `nf-s11-format-20261029-02` passed `make format`.

The database guard test rejects enumerating queries and simulates deletion between table admission and row retrieval. Production `ListRows` references are absent. F22 is fixed; graph iteration, supplied-reference admission and commit-time revalidation remain covered. No remaining S11 engineering risk blocks S12-M. Source rollback restores the prior implementation without data conversion. Checkpoint: `CARTULARY_TEST_RUN_ID=nf-s11-checkpoint-markdown-20261029-01 make lint-markdown`.


### S12-M execution

S11 checkpoint passed (owning exit 0). The physical migration retains the existing source table, grants, references, digest and Recovery identity. Only compressed existing bytea values require reserialization; EXTERNAL is the policy for future writes. Operational application requires quiesced workers, a recoverable snapshot, and disk/WAL headroom; no deployment is performed here.


S12-M exit: PASS. Migration 31 is Imports-owned, sets EXTERNAL storage and reserializes only compressed legacy values by bytea concatenation. Existing source references, bytes, digests, grants and table identity are preserved. The populated migration test begins at version 30 with genuinely compressed binary content, checks exact bytes/digest/size after migration, and checks an operator-region-style INSERT SELECT copy. Empty bootstrap and repeated migration also pass. Recovery database coverage and actual operator-region replay remain valid. Migration locks and disk/WAL headroom remain rollout considerations, not unverified engineering behavior.

Commands (all owning exits 0; roots `.cartulary/test-results/<run-id>`):

- `nf-s12m-generate-20261029-01`: `make generate`; reviewed migration ownership/history projections.
- `nf-s12m-format-20261029-01`: `make format`.
- `nf-s12m-migration-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap`.
- `nf-s12m-drift-20261029-01`: `make migration-drift`.
- `nf-s12m-generated-20261029-01`: `make source-check CHECK=generated`.
- `nf-s12m-region-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`.
- `nf-s12m-recovery-20261029-01`: `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.database_coverage_rejects_an_unclassified_recove_a609c7d3a3`.

Selections were discovered with `make task-guide OWNER=module.database_migrations`, `make task-guide OWNER=module.imports`, and `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s12m-selection.json DETAIL=rows` (also `-region.json` and `-recovery.json`); all exited 0. Exact source identities, work-unit and cleanup results are retained in each run manifest/summary. No required check failed; no broad CI or deployment was run. Checkpoint: `CARTULARY_TEST_RUN_ID=nf-s12m-checkpoint-markdown-20261029-01 make lint-markdown`.


### S12 execution

S12-M checkpoint passed (exit 0). Separating metadata lookup from bounded content access; ordinary readers own read-only consistent snapshots and apply readers borrow the existing protected transaction. Upload/discovery and Recovery inventory memory remain outside this claim.


S12 exit: PASS. Metadata queries no longer select source bytes. Content readers issue completed substring queries of at most 64 KiB, validate declared length and SHA-256 before EOF, and release owned read-only repeatable-read snapshots on EOF/error/close/cancellation. Borrowed readers take a source SHARE lock, use the supplied transaction connection and never end it; Network Flow apply now uses that path. XLSX region discovery explicitly consumes the reader into its required complete archive and remains outside the bounded analytical claim. Application buffer bounds and PostgreSQL buffer observations are separate evidence.

V12 extends the registered upload/replay integration selector with empty/exact/boundary/multiple-chunk sizes, size corruption, digest mismatch, injected read errors, repeated/early close, idle cancellation, consistent snapshot mutation, single-connection borrowed execution and mutation exclusion after reader close. Database cost observations compare a 64 KiB slice from 512 KiB and 4 MiB uncompressed sources, checked in the selected test. Successful test log lines are not retained by this harness, so no numeric observations are claimed from those logs. A final page-access assertion requires at most 64 shared pages for each fixed slice; the 4 MiB whole value requires over 500 pages. This is a database access bound, not a timing benchmark. No public error envelope or capability identity changed.

Commands (roots `.cartulary/test-results/<run-id>`):

- `nf-s12-format-20261029-01`, `-02`, `-03`: `make format`, each exit 0.
- `nf-s12-boundary-20261029-01`: `make lint CHECK=backend-boundary`, exit 0.
- `nf-s12-readers-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`, exit 2, harness/tool diagnostic failure while compiling a remaining XLSX-region caller of the removed helper; selected product cases did not execute. Repaired that caller to consume the bounded source reader.
- `nf-s12-readers-20261029-02`: same selection plus `module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`, exit 0; 4/4 work units, cleanup 17/17.
- `nf-s12-readers-20261029-03`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation`, exit 0 after adding fault/cost cases and preserving cleanup errors.

Source and graph identities are retained in each manifest. Long snapshots consume one connection each; cancellation and deterministic cleanup bound their lifecycle. Source decoding and generic-row orchestration remain S13/S14 and S12-I obligations, respectively; no end-to-end bounded apply claim is made yet. Checkpoint: `CARTULARY_TEST_RUN_ID=nf-s12-checkpoint-markdown-20261029-01 make lint-markdown`.


S12 checkpoint follow-up, before S12-I: successful-test logs were not retained, so strengthened database-cost evidence to an explicit page-access assertion and reopened S12 for its focused verification.


S12 database-cost follow-up passed: `CARTULARY_TEST_RUN_ID=nf-s12-cost-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2`, owning exit 0, root `.cartulary/test-results/nf-s12-cost-20261029-01`. Both EXTERNAL source sizes satisfied the fixed-slice bound of at most 64 shared pages. S12 is DONE. Renewed checkpoint: `CARTULARY_TEST_RUN_ID=nf-s12-checkpoint-markdown-20261029-02 make lint-markdown`.


### S12-I execution

S12 renewed checkpoint passed (exit 0). Removing job-wide row payloads and splitting generic versus analytical admission/execution. Migration 32 will require the generic row digest only for workbook plans and remove obsolete analytical digests; mixed old/new workers are unsupported.


S12-I exit: PASS; F24 is fixed. Job scheduling loads only ordered unit identifiers. Admission reads mapping/source metadata and hashes generic discovered rows only in the generic branch; analytical admission protects and checks the original source identity. Locked analytical execution never fetches or decodes generic rows, and verifies source capability, mapping fingerprint/digest and admitted target binding. Generic execution retains discovered-row digest admission. Migration 32 clears analytical row digests, enforces target-dependent nullability and preserves every other plan field. Downgrade with analytical plans fails instead of synthesizing obsolete metadata; snapshot rollback remains required.

V19 evidence includes the real Network Flow mapping/admission/apply/replay path with poisoned generic JSON and database SELECT access to that column denied to runtime, generic source/mapping mutation regressions, analytical source/mapping mutation after admission, cancellation and committed-result recovery. Populated migration tests preserve source bytes and all plan metadata except the obsolete analytical digest, retain generic digests, and reject either invalid nullability shape.

Commands (all owning exits 0; roots `.cartulary/test-results/<run-id>`):

- `nf-s12i-generate-20261029-01`: `make generate`.
- `nf-s12i-format-20261029-01`, `-02`: `make format`.
- `nf-s12i-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.apply_transaction_current_state_bc8ad0ac32,module.imports.integration.unit_commit_crash_recovery_108cd054a6,module.imports.integration.uncommitted_unit_cancellation`.
- `nf-s12i-current-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.apply_transaction_current_state_bc8ad0ac32`, after extending the selector with analytical source/mapping mutations.
- `nf-s12i-migration-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap`.
- `nf-s12i-drift-20261029-01`: `make migration-drift`.
- `nf-s12i-generated-20261029-01`: `make source-check CHECK=generated`.
- `nf-s12i-schema-20261029-01`: `make source-check CHECK=schema`.

Selection: `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s12i-selection.json DETAIL=rows`, exit 0. Manifests retain exact source/graph identities; all selected work units and cleanup passed. No engineering failure remains for this slice. Operational risk is mixed workers against the changed schema: quiesce workers and take a recoverable snapshot before migration. S13 follows the checkpoint: `CARTULARY_TEST_RUN_ID=nf-s12i-checkpoint-markdown-20261029-01 make lint-markdown`.


### S13 execution

S12-I checkpoint passed (exit 0). Implementing the adopted owner dialect incrementally, with bounded per-cell retained text, exact scalar/hash accounting, field-count precedence, complete raw-cell diagnostics and raw-only preview tail verification. Apply collection removal remains S14.


S13 exit: PASS; F23 is fixed. Replaced whole-source loading and standard-library CSV interpretation with the adopted incremental decoder. It retains bounded cell text, hashes decoded values while discarding excess text, preserves exact scalar counts, counts blank/logical records correctly and accepts CR/LF/CRLF without changing quoted line breaks. Preview stops semantic decoding after record 50 and drains raw bytes for integrity. Field-count mismatch has sole row-level precedence; otherwise mapped/unmapped/ignored oversized cells produce exact, redacted diagnostics, skip dependent target validation and preserve other target checks. Preview admission accepts the adopted column-aware bound and rejects nonconforming limit attribution/privacy. Apply still retains complete record/accepted collections until S14.

V13/V14 cover independently authored dialect cells at byte splits, UTF-8/BOM/framing failures, source and row digest vectors, 50/51 boundaries with malformed/invalid-encoding tails, raw read/hash failures, header/column/row limits, exact scalar counts/hashes after discarded text, mapped/unmapped/ignored diagnostics, field-count precedence and a complete 1,500-diagnostic preview. Existing diagnostics-retention and apply/error regressions pass.

Commands (all owning exits 0; roots `.cartulary/test-results/<run-id>`):

- `nf-s13-format-20261029-01`, `-02`, `-03`: `make format`.
- `nf-s13-parser-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.query_authoring_admission`.
- `nf-s13-parser-20261029-02`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a`, after adding full-preview bound and valid ignored-column cases.
- `nf-s13-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`.

All selected work units and cleanup passed; source identities remain in the manifests. Existing tables are unchanged. Corrected interpretation applies to new original-source imports and does not reinterpret stored tables. No remaining S13 blocker; checkpoint before S14: `CARTULARY_TEST_RUN_ID=nf-s13-checkpoint-markdown-20261029-01 make lint-markdown`.


### S14 execution

S13 checkpoint passed (exit 0). Replacing complete apply collections with bounded summary validation followed by row-at-a-time insertion through the borrowed source transaction. No table is allocated until validation succeeds with accepted rows; Imports retains rollback/commit/cancellation ownership.


S14 exit: PASS; F21 is fixed. First-pass validation retains counts, source and
mapping identities, bounded descriptors, the configured diagnostic prefix and
at most 50 all-rejected diagnostics. Accepted-row limits are enforced before
allocating a table. The second pass inserts one accepted row at a time in source
order on the borrowed transaction, then compares source, descriptor, count and
full validation-summary digests. Production apply no longer collects parsed or
accepted rows; materialized apply helpers exist only in test fixtures. Imports
passes the complete locked source capability and retains final commit,
cancellation and committed-result recovery ownership.

V15 covers increasing 10/1,000/10,000-row inputs with fixed retained diagnostic
bounds and 64 KiB reads, retention 0/N/N+1, bounded all-rejected sampling,
accepted limits, multiple diagnostics per rejected row, second-pass read/digest/
insertion/cancellation failures and summary disagreement. The real Imports path
still runs with generic-row SELECT access denied. A database trigger proves
failure on the second insert rolls back the first row and table; a later
outcome-write failure rolls back table, rows, diagnostics and audit effects.
This establishes bounded application structures and query widths, not a measured
heap/RSS ceiling. Database slice-page assertions remain separate S12 evidence.

Commands (roots `.cartulary/test-results/<run-id>`; explicit
`CARTULARY_TEST_RUN_ID` for each):

- `nf-s14-format-20261029-01` through `-05`: `make format`, each exit 0.
- `nf-s14-generate-20261029-01`: `make generate`, exit 0; registered the new
  streaming validation row through the authored Network Flow family.
- `nf-s14-initial-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`,
  owning exit 2, product/test_assertion_failure, 2/3 work units passed, cleanup
  14/14. Full-capability validation exposed Imports' previously fabricated
  partial capability; repaired the facade to pass its already-validated locked
  capability. Source digest `sha256:9adbba202eac4a8895a232cb5dac47f16f5ad0a7794d1cdc58e18c3b98996aae`.
- `nf-s14-streaming-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232`,
  exit 0, 4/4 work units, cleanup 15/15.
- `nf-s14-imports-20261029-02`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6`,
  exit 0, 5/5 work units, cleanup 20/20. These two repaired runs share source
  digest `sha256:cd27701e28ea6cd81766884aa9b38896be8e11d3f93ac3d86e2e391e67c64155`.
- `nf-s14-atomic-20261029-03`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873`,
  exit 0, 3/3 work units, cleanup 14/14 after extending late-failure assertions;
  source digest `sha256:46a34642d19edb452c056d509625f9e0354175759ac19705edba63b15cdb3b6e`.
- `nf-s14-go-20261029-01`: `make lint CHECK=go`, exit 0.
- `nf-s14-boundary-20261029-01`: `make lint CHECK=backend-boundary`, exit 0.
- `nf-s14-regression-20261029-02`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.import_preview_admission_f0dc5f2d8b`,
  exit 0, 1/1 work unit, cleanup 10/10. Final review replaced zero-time sentinel
  checks with explicit operand validation, preserving cross-field validation
  for valid year-0001 timestamps. Source digest
  `sha256:a0feefbfcdd484f8b48b31e25dd53d9b2ffaeaee4118d5423feaf9be1c57aee7`.

`make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s14-selection.json DETAIL=rows`
and `make explain-run` for the five named test roots exited 0. Source commit is
`b6d073974802489edfeb892d39f48b128b9d00c4`; exact graph/toolchain identities remain
in each manifest. No cleanup failures, skipped or cancelled units remain in the
passing runs. Longer transactions and two source passes are intentional
operational tradeoffs; pool lifecycle and cancellation are covered. S15 will
finalize and qualify the final source. Checkpoint before S15:
`CARTULARY_TEST_RUN_ID=nf-s14-checkpoint-markdown-20261029-01 make lint-markdown`.


### S15 execution

S14 checkpoint Markdown passed (owning exit 0). Finalizing generated artifacts,
reviewing the complete change and then freezing product source for the focused
acceptance portfolio. Prior slice evidence remains historical development
qualification; final source identity and selected acceptance roots will be
recorded below. Full CI and deployment remain excluded. No compatible full
warm-check root exists, so retained-run maintenance is omitted from
`agent-finalize`; no redundant warm check is launched to manufacture one.


S15 source freeze: `nf-s15-format-20261029-01` / `make format` and
`nf-s15-finalize-20261029-01` / `make agent-finalize` passed, owning exits 0.
Finalizer summary `unit-artifacts/finalize-summary.json` reports zero updated
files, successful schema/catalog/tier validation and generated refresh. Reviewed
migration projections, active major-8 Go/browser publication, removed obsolete
schemas and expected generated AJV identifier renumbering. `git diff --check`
passes. Product source is frozen at commit
`b6d073974802489edfeb892d39f48b128b9d00c4`, dirty source digest
`sha256:e7ae7590b06382933642898e7e906bcc0c893683b2e15d87ae20f58d956188fb`.
Toolchain digest is
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`;
environment/capability digest is
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Only this non-executable tracker changes after freeze unless a documented repair
requires a new finalization and evidence decision.


### S15 final acceptance evidence

Current `make task-guide OWNER=module.networkflow`, `OWNER=module.imports`,
`OWNER=module.database_migrations`, `OWNER=module.recovery` and
`OWNER=web.networkflow` all exited 0. Exact selections were planned with
`make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s15-<selection>.json DETAIL=rows`
for selections `nf`, `imports`, `migration`, `recovery` and `web`, each exit 0.
The commands below preserve the selected rows directly so the temporary planning
files are not required for reproduction. Every invocation uses its named
`CARTULARY_TEST_RUN_ID`, repository-root Make and default result parent
`.cartulary/test-results`. Every owning command exited 0.

- `nf-s15-nf-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.unit.network_flow_selector_covers_diagnostic_discover_6f2d284ad8`; PASS, 6/6 work units and 21/21 cleanup steps.
- `nf-s15-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.apply_transaction_current_state_bc8ad0ac32,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6,module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`; PASS, 5/5 work units and 20/20 cleanup steps.
- `nf-s15-migration-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap`; PASS, 7/7 work units and 18/18 cleanup steps.
- `nf-s15-recovery-20261029-01`: `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.database_coverage_rejects_an_unclassified_recove_a609c7d3a3`; PASS, 3/3 work units and 14/14 cleanup steps.
- `nf-s15-web-20261029-01`: `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.networkflowclient_requires_a_claimed_profile_and_65c861c0d2,web.networkflow.regression.saved_graph_major5_decoder_and_lifecycle`; PASS, 2/2 work units and 11/11 cleanup steps.

| Run ID | Exact Make command | Passed work units | Cleanup steps |
| --- | --- | --- | --- |
| `nf-s15-drift-20261029-01` | `make migration-drift` | 8/8 | 19/19 |
| `nf-s15-schema-20261029-01` | `make source-check CHECK=schema` | 3/3 | 12/12 |
| `nf-s15-generated-20261029-01` | `make source-check CHECK=generated` | 5/5 | 14/14 |
| `nf-s15-go-20261029-01` | `make lint CHECK=go` | 6/6 | 15/15 |
| `nf-s15-types-20261029-01` | `make lint CHECK=types` | 2/2 | 11/11 |
| `nf-s15-boundary-20261029-01` | `make lint CHECK=backend-boundary` | 4/4 | 13/13 |

`make explain-run RESULTS_DIR=.cartulary/test-results/<run-id>` exited 0 for
all eleven acceptance roots. All report the exact frozen source, toolchain and
environment identities above, with zero failed, skipped or cancelled work units
and zero failed or blocked cleanup steps. Per-command graph digests, selected
rows, timestamps and fixture profiles are retained in each `run-manifest.json`;
`run-summary.json` and `cleanup-results.json` preserve the results. Work-unit
counts are not test counts. No S15 repair or source drift occurred. These are
focused obligation passes, not complete-owner or full-CI acceptance claims.

### Finding and gate reconciliation

| Finding / gate | Final decision and evidence |
| --- | --- |
| F18 Obsolete contracts | CLOSED, S10/V10/V16: authored and packaged v7/v1 artifacts and exclusive schemas removed after consumer closure; current v8/v2 Go/browser resolution, schema, generation and TypeScript acceptance pass. Historical migration ledger remains intact. |
| F19 Forwarding helper | CLOSED, S10/V10: production helper removed; direct strict decoder, public preview admission and safe Imports translation pass. |
| F20 Whole-source reads | CLOSED, S12-M/S12/V12: byte-preserving EXTERNAL migration, metadata-only access, 64 KiB integrity-checked source readers, owned/borrowed lifecycle, cancellation, source protection and bounded database page access pass. |
| F21 Materialized apply | CLOSED, S14/V15: bounded first-pass summaries and second-pass source-order insertion; accepted limits, identity agreement, late-failure rollback and real Imports cancellation/recovery pass. |
| F22 Whole-table row lookup | CLOSED, S11/V11: scoped incident/table/row retrieval, isolation, active-table and deletion-race checks pass; graph streaming retained. |
| F23 Divergent CSV/limits | CLOSED, S13/S14/V13/V14: shared incremental dialect, exact decoded hashes/counts, raw-cell privacy/precedence, complete bounded previews, explicit cross-field operand validity and 50/51 semantic boundary pass. |
| F24 Generic analytical rows | CLOSED, S12-I/V19: metadata-only scheduling and analytical admission/apply; real PostgreSQL generic-column access denial, generic regressions, nullable-plan migration and identity/replay checks pass. |
| RB-004 Owner adoption | CLOSED by the user's explicit S09 adoption; coordinated owner headers record the human decision. Tests do not supply specification authority. |
| RB-005 Characterization and implementation evidence | CLOSED by S09 characterization plus the implementing slices and compatible frozen-source acceptance above. |
| V17 Final focused acceptance | PASS: finalizer reviewed, product source frozen, exact required portfolio passed without repairs or cleanup failures. |
| V18 Tracker/handoff | Completion text is checked by the final Markdown commands recorded below. |

S09, S10, S11, S12-M, S12, S12-I, S13 and S14 are DONE. Their individual
checkpoint results precede their successors. No engineering remediation blocker
remains; S15 closes after the final documentation checkpoint. The historical
S00–S08 and DOC-02 evidence, including prior failures and obsolete authorization
postures, remain preserved and explicitly superseded by this execution record.

### Operational handoff and acceptance limits

1. Before applying migrations 31/32, quiesce import workers and prevent new apply
   scheduling. Take and verify a recoverable database/environment snapshot.
   Estimate source-storage expansion, temporary rewrite space, WAL volume and
   lock duration from the deployment's source inventory; reserve headroom and
   a maintenance window. Migration 31 changes storage policy and rewrites
   compressed values, preserving source references, logical bytes and digests.
2. Apply the authored migrations with the repository's normal migration owner.
   Migration 32 removes only obsolete analytical plan row digests and enforces
   target-specific source authority; mapping snapshots, source identities,
   outcomes and replay identifiers survive. Prohibit mixed old/new workers and
   old binaries against the migrated admission-plan schema. Both empty and
   populated migration paths and migration drift passed engineering validation.
3. Publish matching server/browser assets at the existing **major-8/state-5**
   boundary, then perform the separately authorized operational acceptance before
   resuming workers. Internal storage changes do not create another public major
   or Network Flow state version. Existing Network Flow tables are neither
   reinterpreted nor deleted. To obtain corrected CSV interpretation, re-import
   original source bytes with the approved mapping decisions.
4. Observe import failures, cancellation/connection cleanup, transaction duration
   and database storage/WAL use after deployment. Apply deliberately performs two
   source passes and row-at-a-time inserts; larger inputs can hold the transaction
   longer. Resource evidence establishes fixed query widths and retained state
   bounded by configured cell/column/diagnostic limits, including analytical
   orchestration without generic rows. It does not claim a measured heap/RSS
   maximum, timing improvement, bounded upload/discovery, or Recovery-wide memory.
5. If rollback is needed, return to the previous matching application/environment
   and database snapshot, including its matching assets, object store and keys.
   Do not downgrade a database after new writes, synthesize analytical generic
   metadata, or delete metadata/data to force admission. Migration 32 intentionally
   cannot reverse populated analytical plans without the matching snapshot.

No deployment, destructive reset or broad historical conversion occurred during
these slices. Operational deployment is a separate, unperformed action; this
handoff is its prerequisite, not its authorization. Full CI, visual suites,
release benchmarks and complete-owner audits were not selected because the
approved plan requires focused acceptance and no presentation behavior changed.
No required focused check is waived. Finalizer retained-run maintenance was
intentionally skipped because no compatible full warm-check root was supplied.

Final documentation checkpoint:
`CARTULARY_TEST_RUN_ID=nf-s15-checkpoint-markdown-20261029-01 make lint-markdown`.
Its root is `.cartulary/test-results/nf-s15-checkpoint-markdown-20261029-01`,
with summary `adhoc/lint-markdown/tool-run-summary.json`.


S15 exit: **DONE**. The final documentation checkpoint passed with owning exit 0
and zero failures. V18 and the operational handoff are complete; every preceding
workstream and finding is reconciled. No required engineering check remains.
All edits are left uncommitted for review, preserving the original tracker
changes and completed history. This completion-status edit receives the final
confirmation `CARTULARY_TEST_RUN_ID=nf-s15-completion-markdown-20261029-01 make lint-markdown`;
its retained summary is
`.cartulary/test-results/nf-s15-completion-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json`.
No product source changed after finalization or acceptance.

## 22. IMP-04 controlling remediation and execution

### Findings, decisions, and acceptance corrections

The user authorized implementation of this plan on 2026-10-29. The working tree
already contains IMP-03 source, specification, generated, test, and migration
changes, including untracked migrations 31/32 and bounded-reader/parser files.
They are preserved. The implementation baseline remains `main` at
`b6d073974802489edfeb892d39f48b128b9d00c4`. No IMP-04 check has run at this initial
checkpoint. The review was read-only and its retained S15 passes do not prove
the missing obligations.

| Finding | Remediation and areas | Rationale / long-term benefit | Compatibility and unresolved risk | Required validation |
| --- | --- | --- | --- | --- |
| F25 Generic discovery blocks owner-valid analytical input | Admit CSV bytes and a source-only unit; defer Workbook projection until generic preview/approval; add owner source inspection and owner-derived analytical descriptors. Specification, implementation, storage migration, contracts, browser, tests, docs. | Imports owns custody/lifecycle; each target owns interpretation. New targets no longer inherit Workbook dialect/limits. | Nullable tabular metadata and coordinated API/browser publication; preserve source, approval, outcome and replay identities. Without correction valid sources cannot reach the owner. | V20 public upload/inspection/preview/approval/apply, literal quotes, 50/51 boundary, default 100001 rows, CSV byte-only limits, generic failure isolation, authorization/integrity/cancellation/replay. |
| F26 Unit failures admitted as row diagnostics | Limit row resource diagnostics to cell_scalar_limit_exceeded at shared preview/retained-state admission. Implementation, tests, affected projections, docs. | One semantic boundary preserves failure scope and privacy. | Tighten already-invalid payloads without altering valid envelopes or rewriting data. Otherwise forged consistent unit failures can appear as row rejection. | V21 negative admission for header/column/row/accepted-row limits; positive mapped/unmapped/ignored cell diagnostics and safe unit translation. |
| F27 Contradictory transaction ownership | Reconcile Network Flow §10.1 with §9.2 and Core: Imports owns transaction lifecycle/final commit; owner writes within borrowed transaction. Specification/docs. | Single commit authority makes atomicity and recovery enforceable. | No intended runtime/storage change. Otherwise future owners can split commits or publish early. | V22 human normative consistency review plus borrowed-transaction/atomicity regression; no executable Markdown dependency. |
| F28 Migrated execution and replay unproved | Valid version-30 generic/analytical plans, migrate same databases through current head, attach normal runtime, execute/cancel/recover/replay. Tests, harness, routing, docs; repair exposed defects. | Proves workflow survival instead of only SQL shape preservation. | Preserve durable identities; explicit reviewed registry translation for equivalent apply bindings only. Otherwise pending plans can strand or duplicate effects after upgrade. | V23 queued/interrupted execution, precommit cancellation, before/after-upgrade committed outcomes, postcommit publication crash, replay and mismatch rejection for both target families. |

CSV upload performs envelope/auth/byte classification/byte ceiling/hash/source
admission only. Discovery creates one csv_file unit without interpreting CSV.
Its rectangle, header/data row references and inferred counts form an all-null
or materialized group; unknown values are never fabricated. Generic preview is
read-only; generic approval atomically materializes its projection and mapping.
Generic apply retains row-digest protection. Imports row/column/cell/archive
limits remain XLSX-specific. Existing generic projections remain usable.

The registered analytical source-inspection route accepts target_kind and
extension_profile_id, derives authorized source identity, and dispatches to its
owner. Network Flow parses only its header using the owner decoder and drains
remaining raw bytes for integrity, returning canonical descriptors/digest with
no mapping requirement or data-validation claim. Mapping preview/approval and
the browser use owner-derived descriptors, never generic rows/columns.

Add an Imports migration for absent tabular projections and removal of obsolete
analytical plan rectangles. Preserve original bytes, generic projections,
approvals, jobs, outcomes and replay identifiers. Registry changes require an
explicit old-to-new migration only for reviewed apply-equivalent bindings;
unknown/mismatched bindings fail closed, with no runtime legacy reader.

### Workstreams and authoritative dependencies

| Workstream | Status | Depends on | Scope / principal risk | Binary exit |
| --- | --- | --- | --- | --- |
| S16 Reopen acceptance and close specifications | DONE | Initial IMP-04 tracker checkpoint | Correct F27; prepare coordinated source-admission amendments and map obligations. Risk: treating proposals as adopted or erasing evidence. | F27 reconciled, new amendments explicitly adopted, decisions/validation mapped, Markdown checkpoint passes. |
| S17 Source admission and target-owned interpretation | DONE | S16 | Nullable unit/projection migration, byte-only CSV admission, lazy generic projection, owner inspection and mapping admission, authored contracts/generation. Risk: weakened admission, projection races, registry drift. | Focused backend, migration and generated checks pass; analytical content remains independent of generic projection. |
| S18 Browser and public workflow integration | DONE | S17 | Owner inspection for mapping, honest unavailable metadata, stable source identity, public pipeline/browser tests. Risk: generic fallback or stale descriptors. | V20 passes including real 100001-row default-limit apply; browser no longer requests generic preview for Network Flow. |
| S19 Diagnostic admission | DONE | S18 | Shared row-level reason restriction and adversarial tests. Risk: rejecting valid cell diagnostics or changing safe unit translation. | V21 passes at preview and retained-state boundaries. |
| S20 Migrated workflow execution | DONE | S19 | Scratch-runtime harness capability, valid legacy fixtures, execution/cancellation/recovery/replay and exposed repairs. Risk: invalid fixtures, binding drift, duplicated effects or cleanup leaks. | V23 and reopened V19 pass for both targets; migration drift and affected harness checks pass. |
| S15-R Final validation and handoff | DONE | S20 and all reopened obligations | Finalize/review/freeze, compatible focused portfolio, tracker reconciliation and operations handoff. Risk: stale evidence or unclosed obligations. | All findings/exits reconciled, required checks pass, handoff complete, no engineering blocker. |

Every successor waits for its predecessor's validation and tracker checkpoint.
Record exact commands/selections, source identity, owning exits, run roots,
cleanup, failures/repairs, skipped checks/reasons, remaining risks and exit
decision. A failed required check leaves the workstream incomplete.

### Validation and handoff obligations

V20 covers literal quotes, malformed framing/UTF-8 within versus beyond preview
record 50, full-apply rollback, a real 100001-row valid CSV under default limits,
CSV independence from generic row/column/cell limits, owner limits, byte-overflow
no-side-effects, generic-failure isolation, XLSX/generic regressions, source and
mapping integrity, authorization, cancellation and idempotency. Register the
large fixture with a declared resource budget; scaled examples do not replace it.

V21 uses otherwise-valid forged diagnostics for all four unit-only limit reasons
and positive privacy/attribution/hash cases. V22 is a human owner-text audit plus
existing executable transaction tests, never a Markdown-based product check.
V23 uses genuine legacy mappings and frozen old registry/binding identities at
schema 30, migrates the same database through 31/32/new head, and executes real
workers/recovery. Cover outcomes committed before upgrade and after migration,
precommit cancellation, publication interruption, repeated replay, source/mapping/
generic-row/binding mismatches, and analytical generic-column access denial.
Retain physical byte/compression/region-copy/Recovery migration coverage.

Discover current guidance with make task-guide and exact selections with
make target-plan; register authored verification inputs and regenerate via Make.
Run focused test-slice, schema/generated, Go/TypeScript, backend/frontend boundary,
migration-drift, affected browser and harness checks. Run agent-finalize, review
changes, freeze source, then final acceptance; recompute compatibility after
repairs. Run lint-markdown at tracker checkpoints. Full CI is excluded by default.
Tests, generators, runtime metadata and release evidence never consume Markdown.

Operational handoff requires quiesced workers, a verified recoverable snapshot,
migration preflight and matching server/browser publication; mixed workers are
unsupported. Existing Network Flow tables remain untouched. Corrected results
require original-source re-import. Upload buffering, generic materialization and
Recovery-wide memory are outside the bounded analytical claim. Rollback restores
the matching application/database environment snapshot, never downgrade or
metadata deletion after writes. Commit/deployment are separate unperformed actions.

### Execution log

| Checkpoint | Changes / evidence | Exit / next action |
| --- | --- | --- |
| IMP-04 initial tracker edit | Preserved IMP-03 working-tree changes and history; reopened affected completion claims; registered F25–F28, V20–V23 and S16–S20/S15-R. | S16 IN_PROGRESS; prepare concrete owner amendments before adoption request. |
| S16 owner candidate preparation | Core 00 §1.4 records proposed adoption scope; Core 01 §17.2 and Core 03 §11 define byte-only CSV source admission, nullable metadata, lazy generic projection and target-owned inspection/approval; Network Flow §10.1 defines the closed header-only inspection operation and corrects final commit ownership; Extensions and Core 04 are coordinated. Existing §1.3 adoption remains intact. | F27 text corrected; source-admission candidate awaits explicit human adoption. No dependent implementation begun. |
| S16 command discovery and initial document check | `make task-guide OWNER=module.imports` and `make task-guide OWNER=module.networkflow` exited 0. `CARTULARY_TEST_RUN_ID=nf-s16-candidate-markdown-20261029-01 make lint-markdown` exited 0; root `.cartulary/test-results/nf-s16-candidate-markdown-20261029-01`, summary `adhoc/lint-markdown/tool-run-summary.json`, zero failures. | A final candidate checkpoint follows the small owner-scope clarification and this log entry. Product tests, generation and finalization are intentionally deferred until the implementation workstreams. |
| S16 coordinated review checkpoint | `CARTULARY_TEST_RUN_ID=nf-s16-review-markdown-20261029-01 make lint-markdown` exited 0, zero failures; root `.cartulary/test-results/nf-s16-review-markdown-20261029-01`, summary `adhoc/lint-markdown/tool-run-summary.json`. Human-oriented source review reconciled §10.1 with §9.2 and Core transaction ownership; analytical versus generic descriptor authority is explicit. Source inspection samples are explicitly empty because no data records are interpreted. | Candidate ready for explicit owner adoption; S16 remains IN_PROGRESS and S17 has not begun. |

S16 migration baseline: the existing generated import-target registry digest is
`6ed462c191ff86e499efefc153e222f2bc0b7ee7098c6e018443bfcdacb5937b`.
This is an observed executable identity for later legacy fixtures/migration
review, not a document-derived runtime input. Capture legacy machine fixtures
from the authored/generated contract baseline, never from this Markdown record.
No product source, test, migration or generated output has been changed by
IMP-04/S16. Required product checks are deferred to their dependent slices;
agent-finalize is deferred to S15-R. The final prepared-document checkpoint is
`CARTULARY_TEST_RUN_ID=nf-s16-adoption-candidate-markdown-20261029-01 make lint-markdown`.

That prepared-candidate check passed with owning exit 0 and zero failures; root
`.cartulary/test-results/nf-s16-adoption-candidate-markdown-20261029-01`, summary
`adhoc/lint-markdown/tool-run-summary.json`. `git diff --check -- docs` also
exited 0. The explicit owner-adoption question has been presented for Core 00
§1.4; no answer has yet been recorded. S16 cannot exit and S17 cannot begin
until that required decision arrives. The user already adopted the distinct
bounded-import candidate in §1.3; that decision remains in force. This evidence
and gate-status log receives
`CARTULARY_TEST_RUN_ID=nf-s16-pending-adoption-markdown-20261029-01 make lint-markdown`.


### S16 adoption and exit

On 2026-10-29 the user explicitly answered "Adopt the prepared candidate and
continue implementation" to the Core 00 §1.4 adoption request, including the
nullable source-unit metadata, owner source inspection, lazy Workbook
interpretation, coordinated browser/API changes and migration rules. Core 00,
Network Flow, Extensions and Core 03 now record the adopted posture. This closes
the owner gate; no further adoption request is required for these decisions.
The pending-adoption Markdown checkpoint above passed with owning exit 0 and
zero failures. F27 is reconciled; transaction regressions remain required in
final acceptance. S16 final adoption checkpoint:
`CARTULARY_TEST_RUN_ID=nf-s16-adopted-markdown-20261029-01 make lint-markdown`.
No product code or generated input changed during S16.


S16 exit: **DONE**. The adopted-document checkpoint passed with owning exit 0,
zero failures, root `.cartulary/test-results/nf-s16-adopted-markdown-20261029-01`,
summary `adhoc/lint-markdown/tool-run-summary.json`. Explicit owner adoption,
F27 correction, interface/migration decisions and validation mappings are
complete. Product checks remain allocated to implementation/final acceptance;
no passing product evidence is claimed for S16. Before S17, this status checkpoint
uses `CARTULARY_TEST_RUN_ID=nf-s16-exit-markdown-20261029-01 make lint-markdown`.


S16 status checkpoint passed (owning exit 0, zero failures) at
`.cartulary/test-results/nf-s16-exit-markdown-20261029-01`. S17 begins with
authored contracts, original-source admission and nullable generic projections.

### S17 implementation and focused qualification (in progress)

Source identity remains HEAD `b6d073974802489edfeb892d39f48b128b9d00c4`
plus the preserved IMP-03 changes and current IMP-04 working tree. CSV admission
now creates a source-only unit; generic preview derives its projection without
persisting it and generic approval persists it atomically. Analytical inspection
and mapping use the owner decoder and source capability. Migration 33 introduces
the nullable projection group, removes analytical plan rectangles, and translates
only the known old registry and reviewed selector/binding pairs. Machine contract
comparison found no changes to mapping/create/apply/error/commit contracts;
unknown identities remain subject to normal fail-closed admission. Browser work
and migrated execution proof remain assigned to S18 and S20.

Execution evidence so far (all roots under `.cartulary/test-results/`):

| Exact command | Run ID / owning exit | Result / repair |
| --- | --- | --- |
| `make generate` with explicit `CARTULARY_TEST_RUN_ID` | `nf-s17-generate-20261029-01`, exit 2; tool exit 1 | Failed: release ledger covered 439/449 changes. Added the ten adopted nullable-metadata/inspection changes to the authored 2.0.0 change set. |
| Same command | `nf-s17-generate-20261029-02`, exit 2; tool exit 1 | Failed: duplicated source-column schema ID in inspection helper. Removed the helper's duplicate public identity. |
| Same command | `nf-s17-generate-20261029-03`, exit 2; tool exit 1 | Failed: new public inspection schemas missing from owner index. Registered both authored schema IDs. |
| Same command | `nf-s17-generate-20261029-04`, exit 0 | Passed, zero failures. |
| Same command | `nf-s17-generate-20261029-05`, exit 2; tool exit 1 | Failed: newly registered test rows were not ASCII sorted. Sorted authored row registries. |
| Same command | `nf-s17-generate-20261029-06`, exit 0 | Passed, zero failures, including updated case routing. |
| `make target-plan TARGET=test-slice DETAIL=rows OWNER=module.networkflow` | exit 2 | Selection usage error: OWNER belongs in INPUTS. No product execution. |
| `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-nf-s17-plan.json DETAIL=rows` | exit 0 | Corrected selection: Network Flow preview admission row only. |
| `make format` with explicit run ID | `nf-s17-format-20261029-01`, exit 0 | Passed. |
| Same command | `nf-s17-format-20261029-02`, exit 2 | Configuration failure from unsorted authored rows above; no product result. |

Generator summaries are `<run>/generate/tool-run-summary.json`; failures retain
their original harness/unknown_failure classification and concrete causes above.
Formatting summaries are graph-owned. These runs allocate no product services.
No product pass or S17 exit is claimed yet; focused backend, migration, schema
and generated validation remain required before S18 begins.

S17 follow-up evidence and repairs:

- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s17-imports-selection.json DETAIL=rows`,
  `/tmp/nf-s17-nf-selection.json`, and `/tmp/nf-s17-migration-selection.json`
  each exited 0. Selections are the seven Imports rows listed below, the two
  Network Flow rows, and migration schema bootstrap respectively.
- `CARTULARY_TEST_RUN_ID=nf-s17-format-20261029-03 make format` passed, exit 0.
- `CARTULARY_TEST_RUN_ID=nf-s17-networkflow-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.import_preview_admission_f0dc5f2d8b`
  passed, exit 0. New inspection vectors cover malformed/invalid-encoding data
  tails, source digest mismatch, invalid headers and strict response admission.
- `CARTULARY_TEST_RUN_ID=nf-s17-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.unit.behavior_contract,module.imports.unit.module_composition_contract,module.imports.unit.target_neutral_csv_source_admission,module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912`
  passed, exit 0. This covers source-only metadata, read-only generic preview,
  lazy generic approval/apply, analytical inspection/approval/apply under denied
  generic-row access, read-role rejection, replay and XLSX operator regions.
- `CARTULARY_TEST_RUN_ID=nf-s17-migrations-20261029-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap`
  and `CARTULARY_TEST_RUN_ID=nf-s17-migration-drift-20261029-01 make migration-drift`
  passed, owning exits 0. Preserved bytes/compression/region copies and plan
  metadata tests now include analytical rectangle removal and grouped-null
  projection constraints. These are storage proofs, not S20 execution proof.
- The four runs above used source digest
  `sha256:6a7065863dd3314e0d7deace15f486c550fcb878d4b20c8f1d75b8563c3ceace`.
  Their cleanup records show respectively 10, 18, 18 and 19 completed steps,
  no failures. Roots are `.cartulary/test-results/<run-id>`; graph acceptance and
  cleanup are in `run-summary.json` and `cleanup-results.json`.
- `make source-check CHECK=schema`, run `nf-s17-schema-20261029-01`, failed
  with owning exit 2 / graph product test_assertion_failure: the authored route
  shape validator lacked source inspection. Updated its exact keys/operation
  model. Run `nf-s17-schema-20261029-02` then failed on stale topology metadata;
  `make source-check CHECK=generated`, run `nf-s17-generated-20261029-01`,
  independently detected that same stale generated index. All cleanup completed.
- `CARTULARY_TEST_RUN_ID=nf-s17-generate-20261029-07 make generate` passed,
  exit 0. Subsequent schema run `nf-s17-schema-20261029-03` and generated run
  `nf-s17-generated-20261029-02` passed, exits 0, cleanup 12/14 completed steps.
  Both used source digest
  `sha256:6b1e8bff6c4e58b4894b229ff5569c78074ec276b19529cf3ba00fb0358c84c0`.
- `make lint CHECK=backend-boundary`, run `nf-s17-boundary-20261029-01`,
  exited 2 at runtime planning while tooling regeneration was in progress;
  no product execution or acceptance is claimed. After regeneration, run
  `nf-s17-boundary-20261029-02` passed, exit 0, cleanup 13 completed steps.
- `make lint CHECK=go`, run `nf-s17-go-20261029-01`, failed, exit 2:
  staticcheck S1016 required the direct private selector conversion. Repaired
  it; run `nf-s17-go-20261029-02` passed, exit 0, cleanup 15 completed steps.
  The repaired boundary/Go runs used source digest
  `sha256:8e9795fd214cce6562820a6ba00bb20994bdf8d9b534c7082194a4240ffd9f41`.
- Two diagnosis-only invocations were malformed: explain-run DETAIL=failures
  is unsupported, and DETAIL=logs requires TARGET; both exited 2 and executed
  no tests. Corrected bounded diagnosis used TARGET plus DETAIL=summary/logs.
- `CARTULARY_TEST_RUN_ID=nf-s17-markdown-progress-20261029-01 make lint-markdown`
  passed, exit 0. `git diff --check` passed.

Source capability lookup additionally rejects streams attached to a different
session/unit. The affected seven-row Imports selection is being requalified in
`nf-s17-imports-20261029-02` with the same exact command/ROWS as above. Historical
checks retain their actual scope; final-source compatibility is reassessed in
S15-R. Full CI, browser acceptance and migration execution/recovery remain
intentionally deferred to their assigned slices. S17 exit is still pending the
requalification and final Markdown checkpoint.


### S17 exit checkpoint

The repeated seven-row Imports selection passed with owning exit 0 at
`.cartulary/test-results/nf-s17-imports-20261029-02`, source digest
`sha256:8e9795fd214cce6562820a6ba00bb20994bdf8d9b534c7082194a4240ffd9f41`,
all 18 cleanup steps completed. S17 implementation, focused backend checks,
source/schema regeneration, Go/boundary validation and migration checks are
complete. **S17 exit: DONE**, subject to the Markdown checkpoint below before
S18 starts. F25 end-to-end closure remains open for S18; F28 remains open for
S20. Matching browser publication, known-registry translation execution,
large-input public scenarios and final evidence compatibility remain explicit
later obligations. No commit or deployment occurred.

Checkpoint command:
`CARTULARY_TEST_RUN_ID=nf-s17-exit-markdown-20261029-01 make lint-markdown`.

S17 Markdown exit checkpoint passed, owning exit 0 and zero failures. S18 now
begins; S19/S20/S15-R have not begun.

### S18 implementation and qualification (in progress)

Network Flow's browser controller and mapping model now consume the registered
owner source-inspection result. Approval echoes owner descriptors with the
owner's fixed row references. Generic Workbook preview accepts initially absent
metadata and supplies its own projection references on approval. Unit source
identity excludes derived rectangles. The Import Assistant displays unavailable
tabular metadata explicitly. Browser fixtures use source-only units, and the
workspace workflow asserts inspection is called and generic preview is not.

Two public backend rows are registered: target-neutral source admission exercises
literal quotes, early/late malformed framing and encoding, generic-preview
failure followed by analytical success, generic row/column/cell settings,
owner column limits, byte overflow, atomic apply failure and request replay.
The separate default-limit row imports exactly 100001 records under default
owner/import limits. Its declared `backend_capacity_isolated` profile reserves
1024 MiB, one CPU/IO/process, plus required service minima; its test bounds apply
observation at two minutes. This profile is an admission reservation, not a
measurement of application or database resource use.

- `make task-guide OWNER=web.networkflow` passed, exit 0.
- `CARTULARY_TEST_RUN_ID=nf-s18-generate-20261029-01 make generate` passed,
  exit 0; summary `generate/tool-run-summary.json` at the explicit run root.
- `CARTULARY_TEST_RUN_ID=nf-s18-types-20261029-01 make lint CHECK=types`
  failed, exit 2. Causes were a duplicated validator argument and test fixtures
  still shaped as generic previews/non-null units. Repaired the authored
  entrypoint and fixtures; `nf-s18-types-20261029-02` passed, exit 0.
- `CARTULARY_TEST_RUN_ID=nf-s18-format-20261029-01 make format` failed,
  exit 2, because the inspection local lacked an explicit type. Added the
  owner-decoder return type; `nf-s18-format-20261029-02` passed, exit 0.
- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s18-web.networkflow-selection.json DETAIL=rows`,
  `/tmp/nf-s18-module.imports-selection.json`, and `/tmp/nf-s18-public-selection.json`
  all passed, exit 0. The first selects four Network Flow browser/controller
  rows; the second selects seven Imports transport/lifecycle/presentation rows;
  the third selects the two new public backend rows.

These exact slices are running at `nf-s18-web-networkflow-20261029-01`,
`nf-s18-web-imports-20261029-01`, and `nf-s18-public-20261029-01` under
`.cartulary/test-results/`. Their eventual owning exits, source identities,
cleanup, failures and repairs must be reconciled before S18 exits. S19 has not
begun. No acceptance claim is made for pending runs.

S18 public and browser-unit results:

- `CARTULARY_TEST_RUN_ID=nf-s18-public-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.target_neutral_public_source_admission,module.imports.integration.network_flow_default_limit_100001_rows`
  passed, owning exit 0, all 17 cleanup steps completed. The default-limit case
  successfully committed exactly 100001 rows through public upload, inspection,
  mapping preview, approval, selection and apply. Late framing/encoding failures
  permitted preview but atomically failed apply; repeated requests retained the
  same job. Generic preview failure did not block owner interpretation.
- `CARTULARY_TEST_RUN_ID=nf-s18-web-networkflow-20261029-01 make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkflowimportmodel_suite_12249943c5,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a`
  passed, exit 0, all 12 cleanup steps completed.
- `CARTULARY_TEST_RUN_ID=nf-s18-web-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.frontend.lifecycle,module.imports.frontend.operator_region_assistant_039d399324,module.imports.frontend.operator_region_coordinator_eb1902fb6e,module.imports.frontend.presentation_recovery,module.imports.frontend.sanitize_import_transport_errors_through_the_sha_1cdfb222a8,module.imports.frontend.skip_reselect_mapping_retention_522ba6591e,module.imports.frontend.typed_transport`
  failed, owning exit 2 / product test_assertion_failure, with all 12 cleanup
  steps completed. Only lifecycle failed: its overlap fixture used two CSV
  units. Corrected the fixture to two overlapping XLSX units and kept the
  source-change/terminal-session assertions. The other six selected rows passed.
- These three runs used source digest
  `sha256:95ebb5cec161068c3e04e63898b31a3e87f58f3ab9aefb9a499e8bf137933759`.
  Detailed failure artifact:
  `nf-s18-web-imports-20261029-01/unit-logs/row-module.imports.frontend.lifecycle/vitest-failure-details.json`.
- `CARTULARY_TEST_RUN_ID=nf-s18-lifecycle-20261029-02 make test-slice OWNER=module.imports ROWS=module.imports.frontend.lifecycle`
  passed, owning exit 0. The focused repair rerun does not rewrite the earlier
  failure or constitute complete-owner acceptance.
- `CARTULARY_TEST_RUN_ID=nf-s18-format-20261029-03 make format`,
  `CARTULARY_TEST_RUN_ID=nf-s18-boundary-20261029-01 make lint CHECK=frontend-boundary`,
  and `CARTULARY_TEST_RUN_ID=nf-s18-protocol-20261029-01 make lint CHECK=protocol`
  passed with owning exits 0. Type/format run roots retain their manifest source
  digests; no cross-source final acceptance is inferred.
- Both exact browser plans passed, exit 0:
  `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s18-e2e-nf-selection.json DETAIL=rows`
  and `/tmp/nf-s18-e2e-imports-selection.json`. The Network Flow plan selects
  import, recovery and retained-draft scenarios; the Imports plan selects the
  claimed production workflow. Recovery's deliberate source-read interruption
  now targets source-inspection, preserving the original recovery obligation.

Real-browser runs `nf-s18-e2e-networkflow-20261029-01` and
`nf-s18-e2e-imports-20261029-01` remain pending. S18 is incomplete until their
outcomes/cleanup and the final tracker checkpoint are recorded. All run roots
above are under `.cartulary/test-results`; graph summaries and cleanup records
remain the owning evidence. No full CI, commit or deployment has occurred.


### S18 exit checkpoint

Both real-browser slices passed with owning exit 0:

- `CARTULARY_TEST_RUN_ID=nf-s18-e2e-networkflow-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_importing_nf_fix_00_094b0f779b,module.networkflow.browser.import_recovery,module.networkflow.browser.import_retained_draft`.
- `CARTULARY_TEST_RUN_ID=nf-s18-e2e-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.browser.claimed_import_assistant_production_workflow_986c2117ee`.

Both used source digest
`sha256:300b8b0e64c917fa8668b76a6a4bd1e17c491bb0f6ab05da8156219df01705cb`.
Their exact roots retain run summaries, browser row results and cleanup records.
The repaired lifecycle and frontend boundary/protocol checks also passed on that
source. F25's required public reachability and browser operation are now proved,
including large-input default limits; S15-R must still requalify the final-source
portfolio. F26 and F28 remain open, allocated to S19 and S20. Existing historical
S09/S12-I/S13/S14/S15 claims remain reopened until their remaining obligations
are reconciled. No commit, deployment, broad conversion or reset occurred.

**S18 exit: DONE**, subject to final Markdown checkpoint before S19 begins:
`CARTULARY_TEST_RUN_ID=nf-s18-exit-markdown-20261029-01 make lint-markdown`.

S18 progress and exit Markdown checks passed, owning exits 0. Both browser
runs completed all 23 cleanup steps; repaired lifecycle/boundary/protocol runs
completed all 11/11/13 steps respectively, without failures. S19 now begins.

### S19 — shared row-diagnostic boundary (complete)

F26/V21 is implemented at the shared semantic admission boundary used by preview
and retained-state/Recovery admission. Resource row diagnostics permit only
`cell_scalar_limit_exceeded`, with positive configured limit, actual greater than
limit, exact limit name, attributed column and both hashes, and no raw sample.
The authored rejected-row schema carries the same conditional restrictions;
unit-level error reasons remain available in the unit error contract.

The existing registered preview admission row now forges all four prohibited
reasons with coherent identities/messages/limit fields and proves rejection by
both semantic validation and public owner-result admission. The existing
registered state-5 admission row repeats these forgeries inside restored-state
transactions and proves rejection and rollback. Existing mapped/unmapped/ignored
cell tests and Imports safe-error translation remain green. No stored diagnostics
are rewritten to accommodate invalid content.

Schema regeneration changed the registry digest to
`b7bd50637cb33fb417d0e490f52d8add3203c45bee130a6d35cd4c449beda859`.
Migration 33's explicit known-binding translation now targets that identity.
The change adds source inspection and rejects previously forbidden row payloads;
valid mapping/apply contracts, outcomes and commit ownership remain equivalent.
S20 must prove that translation with frozen legacy identities.

Exact commands and retained roots (all owning exits 0):

- `make task-guide OWNER=module.networkflow`; `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s19-selection.json DETAIL=rows`; `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s19-imports-selection.json DETAIL=rows` selected the rows below.
- Initial `HARNESS_RUN_ID=nf-s19-generate-20261029-01 make generate` passed; that variable is not the run-ID input, so the actual root is `.cartulary/test-results/20261029T221829Z-p13567`. No root was renamed or reused.
- `CARTULARY_TEST_RUN_ID=nf-s19-format-20261029-01 make format`.
- `CARTULARY_TEST_RUN_ID=nf-s19-admission-20261029-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.integration.extension_state_v4_v1_rejection`.
- `CARTULARY_TEST_RUN_ID=nf-s19-schema-20261029-01 make source-check CHECK=schema`.
- `CARTULARY_TEST_RUN_ID=nf-s19-generate-20261029-02 make generate` regenerated after the explicit migration digest update.
- `CARTULARY_TEST_RUN_ID=nf-s19-imports-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.unit.canonical_write_null_and_safe_owner_translation_rs08_3f58f0c338`.
- `CARTULARY_TEST_RUN_ID=nf-s19-generated-20261029-01 make source-check CHECK=generated`.
- `CARTULARY_TEST_RUN_ID=nf-s19-drift-20261029-01 make migration-drift`.

Named run roots are `.cartulary/test-results/<CARTULARY_TEST_RUN_ID>`; graph
summaries/manifests and cleanup results are retained there. Admission/schema
source identity was `sha256:8ae0a1181f07c2fccafb5f5f730b71b8f94c1c8ef5ed8c9d63521f38a88a20a9`;
post-migration-digest source identity was
`sha256:6f1e96c35bcb8a04f8d8e3f3060aad5a024f9d3aee7fe61f9288903abf60743f`.
Admission/schema/generated/format cleanup completed all 15/12/14/11 steps.
`nf-s19-imports-20261029-01` cleanup completed 15/15 steps.
`nf-s19-drift-20261029-01` cleanup completed 19/19 steps.

No selected validation failed. No full CI, deployment, commit, or stored-data
conversion occurred. Final compatibility checks and F28 remain allocated to
S20/S15-R. **S19 exit: DONE**, pending the documentation checkpoint below; S20
starts only after its successful owning exit.

`CARTULARY_TEST_RUN_ID=nf-s19-exit-markdown-20261029-01 make lint-markdown`.

S19 documentation checkpoint passed with owning exit 0 and zero failures. S20
now begins; the migration fixture will retain its own database/role/cleanup
authority and admit runtime attachment only after current-head readiness.

### S20 — migrated workflow execution (in progress)

The migration-scratch capability now retains its issued fixture/lease and can
lend that same database to normal application composition only after canonical
head readiness and application crypto-format checks through the runtime role.
It does not initialize missing identity, accept arbitrary handles, or transfer
cleanup ownership. REQ-810/811 were reviewed manually; executable inputs do not
read the specification. Unit checks reject unissued capabilities; service-backed
cases exercise pre-head and missing-crypto rejection and normal startup.

Authored JSON fixtures freeze schema-30 source bytes, mappings, fingerprints,
parser identities and registry `6ed462c191ff86e499efefc153e222f2bc0b7ee7098c6e018443bfcdacb5937b`.
The new registered `module.imports.integration.migrated_workflows` row creates
real jobs/plans/idempotency state at version 30, applies 31/32/33 in the same
database, and runs normal workers. Both targets cover queued/interrupted jobs,
precommit cancellation, explicit before/after-commit faults, pre-upgrade committed
outcomes, two restarts/replays, source/mapping/row-digest/binding failures and
analytical execution without generic-row read permission. A nontransactional
PostgreSQL sequence witnesses owner write attempts independently of row counts.
The row uses exclusive `postgres_migration` isolation and existing standard
resources; it is functional migration evidence, not a resource measurement.

Evidence so far (all roots `.cartulary/test-results/<run-id>`):

- `CARTULARY_TEST_RUN_ID=nf-s20-generate-20261029-01 make generate` failed (owning exit 2): new collaborator IDs were not sorted. `nf-s20-generate-20261029-02` then failed (exit 2) on an unknown collaborator ID. The authored row was corrected to sorted registered owners; `nf-s20-generate-20261029-03` passed, exit 0. Task-guide/target-plan attempts made before that repair failed at the same configuration boundary; no product execution occurred in those attempts.
- `CARTULARY_TEST_RUN_ID=nf-s20-format-20261029-01 make format` failed (exit 2) at that configuration boundary. Format runs `nf-s20-format-20261029-02`, `-03`, `-04`, `-05` passed (exit 0).
- `make task-guide OWNER=module.imports`, `make task-guide OWNER=module.database_migrations`, `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s20-selection.json DETAIL=rows`, and `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s20-harness-selection.json DETAIL=rows` subsequently passed (exit 0).
- `CARTULARY_TEST_RUN_ID=nf-s20-workflows-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.migrated_workflows` failed (owning exit 2): fixture membership history had to precede the membership and use `last_version`. Source `sha256:151eaa96fb731c923a00e2e060427efc1d89ab8979144ce30294ce16b868af50`; cleanup 14/14 completed.
- The same exact selection with run ID `nf-s20-workflows-20261029-02` failed (exit 2): the analytical fixture lacked required RFC3339 null/policy members; successful-path jobs also did not finish. Source `sha256:bd7024a832a51eec82a4b03f75a92b3f3ce7d7d80128e305b161e405367b1f4f`; cleanup 14/14 completed.
- `nf-s20-workflows-20261029-03` failed (exit 2): the bounded diagnostic wait showed handler retries, disproving the initial timing hypothesis. The fixtures lacked the frozen Jobs extension ownership/idempotency identity needed for final success publication. Those fields were added; no product timeout or retry policy changed. Source `sha256:a27ffedecb84d873942c5e2baf662632cf5c8337d845c8220fae3e36f81c5648`; cleanup 14/14 completed.
- Failure diagnosis used `make explain-run RESULTS_DIR=.cartulary/test-results/<exact-run-id> TARGET=test-slice DETAIL=summary` and the bounded referenced Go logs. Earlier failures remain retained.
- `CARTULARY_TEST_RUN_ID=nf-s20-harness-20261029-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.test_harness_targeted_operation_validation` passed (exit 0); additional nil/canceled/missing-lease assertions were subsequently added and require revalidation.

The expanded workflow run `nf-s20-workflows-20261029-04` is pending. Committed
fixture response fidelity and capability negatives were further strengthened;
passing earlier runs will not be represented as final-source acceptance. S20
remains IN_PROGRESS; S15-R has not begun.

S20 qualification and repairs:

- `nf-s20-workflows-20261029-04` failed, owning exit 2: the committed generic fixture omitted its record envelope, and analytical cancellation could not settle. Source `sha256:46832623708390a074a0c82de74955559f0d3a0d8342c7608efaaf1c990d9c92`; cleanup 14/14. The envelope was added without changing product admission.
- Cancellation was split into the exact registered row `module.imports.integration.migrated_cancellation`; the workflow row retains 24 other cases across both families. `CARTULARY_TEST_RUN_ID=nf-s20-cancellation-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.migrated_cancellation` failed, exit 2, source `sha256:105daa67dab418d2b0daab0915fc9641d07621dcfd9179a2bbd64dee97229f92`; cleanup 14/14. The generic case passed; analytical cancellation remained blocked.
- The same cancellation selection at `nf-s20-cancellation-20261029-02` failed, exit 2, source `sha256:324028b22dced463bc8569e694eff79689d6bc38a1c32601443d92c0c5ab3040`; cleanup 14/14. A five-second request deadline and PostgreSQL lock evidence identified the actual dependency: the owner's incident `FOR UPDATE` lock blocked cancellation's foreign-key reference while owner insertion waited for the test gate. This was a product cancellation defect, not an insufficient runner deadline.
- Incidents now locks immutable incident identity `FOR NO KEY UPDATE`. Owner writes, lifecycle changes and publication still serialize; independent key-share references can proceed so cancellation can interrupt the borrowed transaction. A focused incident test verifies both allowed references and conflicting writes. The owner still never commits or rolls back the borrowed transaction.
- `nf-s20-workflows-20261029-05` passed the 24-case workflow row, exit 0, on source `sha256:324028b22dced463bc8569e694eff79689d6bc38a1c32601443d92c0c5ab3040`; cleanup 14/14. After the lock repair, `nf-s20-cancellation-20261029-03` passed both live cancellation cases, exit 0; cleanup 14/14.
- `CARTULARY_TEST_RUN_ID=nf-s20-incidents-20261029-01 make test-slice OWNER=module.incidents ROWS=module.incidents.integration.participant_reference_lock,module.incidents.integration.incident_lifecycle_close_reopen_authorization_re_6ac1f84d20` passed, exit 0, cleanup 14/14.
- `CARTULARY_TEST_RUN_ID=nf-s20-bundles-20261029-01 make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.network_flow_retained_state_blocks_bundle_export_7dcd6f85c2` passed, exit 0, cleanup 14/14. These two runs and cancellation-03 used source `sha256:a5473f6c97ebde1116476cce587371dd81e928f7eb8e3bd7b84c7715edcd31f7`.
- `CARTULARY_TEST_RUN_ID=nf-s20-go-20261029-01 make lint CHECK=go` failed, exit 2, on that source: Staticcheck SA1012 flagged the deliberately nil context in the capability rejection test. Added a local explanation/suppression for that negative test; no production behavior changed. Cleanup 15/15.
- Generate runs `nf-s20-generate-20261029-04`, `-05`, `-06` and format runs `nf-s20-format-20261029-06`, `-07`, `-08`, `-09` passed with owning exits 0. Exact plans `/tmp/nf-s20-complete-selection.json`, `/tmp/nf-s20-cancellation-selection.json`, `/tmp/nf-s20-incidents-selection.json`, and `/tmp/nf-s20-bundles-selection.json` passed through `make target-plan TARGET=test-slice INPUTS=<file> DETAIL=rows`. `nf-s20-progress-markdown-20261029-01` passed, exit 0.

The strengthened write witness observes the first owner write (generic record or
Network Flow table); cancellation blocks the child insertion to prove parent
rollback. Assertions include parent/child rows, generic journal/change sets,
one outcome/commit identity, preserved outcome bytes, idempotency identity and
two normal-runtime restarts/replays. Already-committed outcomes cause zero owner
write attempts. Unknown bindings/registries and source/mapping/row changes fail
closed. Runtime attachment rejects old schema and absent crypto identity without
repairing either; all fixture and runtime cleanup remains harness-owned.

On source `sha256:072a3ff4185ae9a445c319556c799ff4b9913138da682cb5f94002ec27873cb6`,
the following commands passed with owning exits 0. Each run uses
`CARTULARY_TEST_RUN_ID=<run-id>` and retains its manifest, summaries and cleanup
under `.cartulary/test-results/<run-id>`:

| Run ID | Exact Make command | Completed cleanup |
| --- | --- | --- |
| `nf-s20-final-workflows-20261029-01` | `make test-slice OWNER=module.imports ROWS=module.imports.integration.migrated_workflows,module.imports.integration.migrated_cancellation` | 14/14 |
| `nf-s20-harness-20261029-02` | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.test_harness_targeted_operation_validation` | 10/10 |
| `nf-s20-go-20261029-02` | `make lint CHECK=go` | 15/15 |
| `nf-s20-drift-20261029-01` | `make migration-drift` | 19/19 |
| `nf-s20-schema-20261029-01` | `make source-check CHECK=schema` | 12/12 |
| `nf-s20-generated-20261029-01` | `make source-check CHECK=generated` | 14/14 |
| `nf-s20-boundary-20261029-01` | `make lint CHECK=backend-boundary` | 13/13 |

A final fixture review aligned the committed generic journal's mutation reference
with its already-correct response reference. This changes test data only;
`nf-s20-final-workflows-20261029-02` repeats the exact two-row selection on that
source. Its owning result is required before S20 exits. No full CI, commit,
deployment, or production data conversion has occurred.

The S20 normative consistency audit also found related stale wording in
NF-REQ-057/062a beyond the original §10.1 sentence. Corrected those passages to
allow validation-summary/streamed insertion without mandatory staging and to
label Imports' cancellation barrier, atomic outcome commit and rollback
explicitly. This implements the already-adopted division of responsibility;
there is no new owner decision or runtime contract. V22's review now includes
§8, §9.2, §10.1, Core source access and the Extensions restatement. Product
checks do not consume these Markdown amendments.

### S20 exit checkpoint

`CARTULARY_TEST_RUN_ID=nf-s20-final-workflows-20261029-02 make test-slice OWNER=module.imports ROWS=module.imports.integration.migrated_workflows,module.imports.integration.migrated_cancellation`
passed with owning exit 0 on source
`sha256:2e16f7f5b4698b1539a69774f96f067953ade5c050b54b81b5c8b6b6590b44e6`.
All 14 cleanup steps completed. The 26 scenarios close V23 and the outstanding
S12-I/V19 migrated-execution obligation. F28 is resolved; V22 also includes the
corrected §8 ownership sequence and NF-BLOCK-011's three-operation restatement.
No engineering blocker remains for S20. Operational snapshot, quiescence,
coordinated publication and final-source acceptance remain S15-R obligations.

**S20 exit: DONE**, subject to the documentation checkpoint before S15-R:
`CARTULARY_TEST_RUN_ID=nf-s20-exit-markdown-20261029-01 make lint-markdown`.

The S20 Markdown checkpoint passed with owning exit 0 and zero failures.

### S15-R finalization and source freeze

Finalizing through `CARTULARY_TEST_RUN_ID=nf-s15r-finalize-20261029-01 make agent-finalize`.
There is no compatible successful full warm-check root, so retained-run
maintenance is skipped; no redundant warm check is launched to obtain one.
After review, the focused final portfolio will cover current public admission,
owner parsing/diagnostics/atomicity, migrated execution and cancellation,
generic/XLSX regression, source readers/storage/Recovery, browser integration,
incident lock participants, and affected static/generated/harness obligations.
Earlier source runs remain slice evidence, not final-source acceptance.

Finalization passed, owning exit 0, with zero generated files changed. Reviewed
the finalizer summary and active generated operation/schema/migration projections;
`git diff --check` passed. Product source is frozen at dirty digest
`sha256:2e16f7f5b4698b1539a69774f96f067953ade5c050b54b81b5c8b6b6590b44e6`,
HEAD `b6d073974802489edfeb892d39f48b128b9d00c4`. Toolchain digest remains
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`;
environment/capability digest remains
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Only non-executable documentation may change without a documented repair.
The exact S20 two-row run `nf-s20-final-workflows-20261029-02` has this same
source/toolchain/environment identity and current graph/selection. It supplies
migrated workflow/cancellation evidence without another redundant execution;
other S20 checks used the earlier fixture identity and are requalified below.

Current task guides for module.networkflow, module.imports,
module.database_migrations, module.recovery, web.networkflow, module.incidents
and module.incidentbundles passed with owning exits 0. Exact current plans use
`make target-plan TARGET=test-slice INPUTS=/tmp/nf-s15r-<selection>-selection.json DETAIL=rows`
for nf, imports, migration, recovery, web, web-imports, incidents, bundles,
browser-nf and browser-imports. The completion record below preserves their
selected rows independently of temporary files.

### IMP-04 operational handoff

Deployment is a separate action. Before applying migrations 31–33, operators
must stop admission and quiesce all import workers, verify a recoverable matching
database/environment snapshot, and preflight canonical migration history,
application crypto identity, registry identities and available disk/WAL capacity.
Migration 31 rewrites compressed source values; schedule its locks and storage
growth within the maintenance window. Migrations 32/33 change admission-plan
and source-unit metadata. Do not run old workers against the migrated shape.

Publish matching server/browser assets at the existing major-8/state-5 boundary.
Confirm current schema head and active target registration before resuming
workers. Verify a source-only CSV through inspection, mapping and apply, a
generic Workbook import, and pending-job recovery using the normal application
workflow. Retain the snapshot until operational verification is complete.

Migration 33 translates only predecessor registry
`6ed462c191ff86e499efefc153e222f2bc0b7ee7098c6e018443bfcdacb5937b`
to current registry
`b7bd50637cb33fb417d0e490f52d8add3203c45bee130a6d35cd4c449beda859`
for the explicitly listed selector/binding pairs with equivalent valid apply
contracts. Unknown identities or mismatched bindings remain untransformed and
fail closed. Preflight must identify those pending plans for investigation;
never edit their metadata merely to force execution. The migrated workflow
tests prove both target families, including failures, recovery and replay.

Existing Network Flow tables are neither reinterpreted nor deleted. Logical
source bytes, source/mapping identities, valid approvals, committed outcomes and
replay identities are preserved. Existing generic projections remain usable;
new CSV units honestly expose absent tabular metadata until generic approval.
Analytical plans lose obsolete generic-row digest and rectangle metadata.
Corrected CSV results require original-source re-import with mapping decisions;
there is no automatic historical conversion or compatibility parser.

The bounded-resource claim applies to analytical preview/apply and its Imports
orchestration. Upload buffering, generic Workbook materialization and Recovery-wide
memory are outside it. Two-pass apply increases source reads and transaction
duration; each ordinary reader occupies a snapshot connection until EOF, close,
error or cancellation. Fixed-slice database page assertions are distinct from
application retained-state bounds and are not a performance benchmark.

Rollback restores the previous matching application/database/object-store/key
environment snapshot and assets. Do not downgrade a database after new writes,
mix old/new workers, delete metadata, or reinterpret existing tables to regain
admission. Engineering work here performs no deployment, destructive reset,
commit or production data conversion. Operational rollout must be recorded
separately after this handoff.


### S15-R final focused acceptance

All ten exact target plans passed, owning exits 0. The following fresh
invocations passed with owning exits 0 on the frozen source/toolchain/environment
identities above. Each command is prefixed with
`CARTULARY_TEST_RUN_ID=<run-id>`; every root is
`.cartulary/test-results/<run-id>`. Selected rows are recorded explicitly so the
temporary planning files are not execution dependencies.

- `nf-s15r-nf-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.unit.network_flow_selector_covers_diagnostic_discover_6f2d284ad8,module.networkflow.unit.source_inspection_header_only`; PASS, 6/6 work units, 21/21 cleanup steps.
- `nf-s15r-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.apply_transaction_current_state_bc8ad0ac32,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6,module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912,module.imports.integration.xlsx_import_discovery_accepts_a_bounded_open_xml_d7c25ef6c8,module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88,module.imports.integration.target_neutral_public_source_admission,module.imports.integration.network_flow_default_limit_100001_rows,module.imports.unit.target_neutral_csv_source_admission,module.imports.unit.canonical_write_null_and_safe_owner_translation_rs08_3f58f0c338`; PASS, 7/7 work units, 24/24 cleanup steps.
- `nf-s15r-migration-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap,module.database_migrations.unit.test_harness_targeted_operation_validation`; PASS, 8/8 work units, 19/19 cleanup steps.
- `nf-s15r-recovery-20261029-01`: `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.database_coverage_rejects_an_unclassified_recove_a609c7d3a3`; PASS, 3/3 work units, 14/14 cleanup steps.
- `nf-s15r-web-20261029-01`: `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkflowimportmodel_suite_12249943c5,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowclient_requires_a_claimed_profile_and_65c861c0d2,web.networkflow.regression.saved_graph_major5_decoder_and_lifecycle`; PASS, 3/3 work units, 12/12 cleanup steps.
- `nf-s15r-web-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.frontend.lifecycle,module.imports.frontend.operator_region_assistant_039d399324,module.imports.frontend.operator_region_coordinator_eb1902fb6e,module.imports.frontend.presentation_recovery,module.imports.frontend.sanitize_import_transport_errors_through_the_sha_1cdfb222a8,module.imports.frontend.skip_reselect_mapping_retention_522ba6591e,module.imports.frontend.typed_transport`; PASS, 3/3 work units, 12/12 cleanup steps.
- `nf-s15r-incidents-20261029-01`: `make test-slice OWNER=module.incidents ROWS=module.incidents.integration.participant_reference_lock,module.incidents.integration.incident_lifecycle_close_reopen_authorization_re_6ac1f84d20`; PASS, 3/3 work units, 14/14 cleanup steps.
- `nf-s15r-bundles-20261029-01`: `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.network_flow_retained_state_blocks_bundle_export_7dcd6f85c2`; PASS, 3/3 work units, 14/14 cleanup steps.
- `nf-s15r-browser-nf-20261029-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_importing_nf_fix_00_094b0f779b,module.networkflow.browser.import_recovery,module.networkflow.browser.import_retained_draft`; PASS, 12/12 work units, 23/23 cleanup steps.
- `nf-s15r-browser-imports-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.browser.claimed_import_assistant_production_workflow_986c2117ee`; PASS, 12/12 work units, 23/23 cleanup steps.

| Run ID | Exact Make command | Passed work units | Completed cleanup |
| --- | --- | --- | --- |
| `nf-s15r-go-20261029-01` | `make lint CHECK=go` | 6/6 | 15/15 |
| `nf-s15r-types-20261029-01` | `make lint CHECK=types` | 2/2 | 11/11 |
| `nf-s15r-backend-boundary-20261029-01` | `make lint CHECK=backend-boundary` | 4/4 | 13/13 |
| `nf-s15r-frontend-boundary-20261029-01` | `make lint CHECK=frontend-boundary` | 2/2 | 11/11 |
| `nf-s15r-protocol-20261029-01` | `make lint CHECK=protocol` | 4/4 | 13/13 |
| `nf-s15r-schema-20261029-01` | `make source-check CHECK=schema` | 3/3 | 12/12 |
| `nf-s15r-generated-20261029-01` | `make source-check CHECK=generated` | 5/5 | 14/14 |
| `nf-s15r-drift-20261029-01` | `make migration-drift` | 8/8 | 19/19 |

`make explain-run RESULTS_DIR=.cartulary/test-results/<run-id>` exited 0 for
all eighteen new acceptance roots and the compatible S20 final-workflows-02
root. Manifests agree on source, toolchain and environment identities; their
per-command graph/selection identities remain explicit. All selected units
passed, none failed/skipped/cancelled, and all cleanup completed without failures
or blocked steps. Work-unit counts are not test counts. Across these slices and
the compatible S20 root, 50 registered verification rows cover the selected
obligations. This is focused acceptance, not a full-owner, full-CI or release
performance claim. No S15-R product-source repair or drift occurred.

### IMP-04 finding and reopened-exit reconciliation

| Finding / gate | Final disposition and evidence |
| --- | --- |
| F25 / V20 | CLOSED: target-neutral source custody, nullable metadata, lazy generic interpretation, registered owner inspection and browser integration. Public admission proves literal quotes, early/late framing/encoding, 50/51 scope, default 100001-row apply, CSV byte-only admission, generic failure isolation and existing safety/replay. Backend and real-browser final rows pass. |
| F26 / V21 | CLOSED: shared semantic/schema row-diagnostic boundary accepts only the cell limit reason. Coherent header/column/row/accepted-row forgeries fail preview and retained-state admission; valid cell and unit-error translation regressions pass. |
| F27 / V22 | CLOSED: Core, Extensions, Network Flow §8/§9.2/§10.1 and NF-BLOCK-011 consistently assign transaction lifecycle, final barrier/commit and recovery to Imports. Optional preparation is not mandatory durable staging. Human document review is supported by borrowed-reader, atomicity and cancellation tests; executable checks never read Markdown. |
| F28 / V23 / outstanding V19 | CLOSED: frozen valid schema-30 generic/analytical state migrates through 31/32/33, normal workers execute/cancel/recover/replay, pre-upgrade outcomes remain unchanged, duplicate owner writes are excluded, and mismatches fail closed. The compatible S20 final root covers all 26 scenarios; final migration/harness/static checks pass. |
| Reopened S09 | CLOSED: adopted source-admission amendments and transaction-ownership cleanup are reconciled. The user's explicit adoption remains the authority; passing checks are not a substitute for it. |
| Reopened S12-I/V19 | CLOSED: analytical orchestration avoids generic content; migrated workflow evidence now proves the previously missing execution/recovery obligation. |
| Reopened S13 | CLOSED: owner parsing/preview and row-diagnostic boundaries pass independent vectors and public-path checks. |
| Reopened S14 | CLOSED: public reachability, bounded validation/insertion, source protection, atomic rollback and Imports commit/cancellation are qualified on compatible source. |
| S10/S11/S12-M/S12 retained implementation | CONFIRMED by active-publication, scoped-row/graph, byte-preserving storage/bootstrap, bounded-reader and Recovery/operator-region regressions. Prior historical results retain their original scope. |
| S15 / V17–V18 / S15-R | CLOSED: finalization, compatible focused acceptance, documentation checkpoint and operational handoff complete. |

No required engineering check is outstanding. Full CI was intentionally not run under the user's focused-validation
instruction. Pixel-golden suites, independent performance/release qualification,
retained warm-check maintenance, commit and deployment are not selected and are
not claimed. All earlier failed runs and their repairs remain preserved above.

Final documentation checkpoint:
`CARTULARY_TEST_RUN_ID=nf-s15r-exit-markdown-20261029-01 make lint-markdown`.

That checkpoint passed with owning exit 0 and zero failures. A final
`git diff --check` found only an extra blank line at this document's EOF;
the completion edit removed it. This documentation-only repair does not affect
the frozen product source or acceptance compatibility.

**S15-R exit: DONE.** All workstreams/findings and reopened acceptance obligations
are reconciled; no engineering blocker remains. Migration, compatibility,
cancellation/recovery, rollback and operational handoff are complete. The work
remains uncommitted and undeployed. Final record check:
`CARTULARY_TEST_RUN_ID=nf-s15r-exit-markdown-20261029-02 make lint-markdown`.

Final record check passed with owning exit 0 and zero failures; its retained
summary is `nf-s15r-exit-markdown-20261029-02/adhoc/lint-markdown/tool-run-summary.json`.
The final whitespace check also passed. S15-R remains DONE.

## 23. IMP-05 — Approval integrity, explicit route semantics and legacy retirement

### Authorization, baseline and supersession

The user explicitly requested implementation of the complete S21–S24/S15-R2
plan after selecting fresh-environment admission and whole-catalog explicit
mutation semantics. This tracker update is the first implementation change.
Baseline: `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, with IMP-03/IMP-04
tracked and untracked changes preserved. No product checks have run in IMP-05.
The current instruction supersedes old-plan preservation/translation and
schema-30 execution obligations, not current-version recovery or replay.

| Finding | Remediation / areas | Rationale, benefit and unresolved risk | Compatibility / validation |
| --- | --- | --- | --- |
| F29 / V24 | Protect and compare source identity in one private transaction operation; implementation and integration tests | Prevent stale approval; retain one transaction owner and explicit protection lifetime | No wire/storage change; deterministic generic/analytical contention, rollback, cancellation and replay |
| F30 / V25 | Complete normative inventory; explicit boolean mutation metadata for all operations; contracts, tooling, runtime, tests and generated outputs | Remove method inference and conflicting security metadata; missing declarations fail closed | Same public request/response shapes; complete catalog, parity, CSRF, assembly and compatibility checks |
| F31 / V26 | Retire prior-format plan translation, fixtures and exclusive harness support; replace useful scenarios with current public admission | Remove unnecessary compatibility burden and invalid test state without losing atomicity/recovery coverage | Fresh database required; bootstrap, current previews/workflows, storage, Recovery and migration drift |

| Workstream | Status | Dependency | Risk / exit |
| --- | --- | --- | --- |
| S21 Authority and acceptance reconciliation | DONE | User decisions | Owner documents and tracker agree; Markdown checkpoint passed |
| S22 Protected approval | DONE | S21 checkpoint | V24 and focused approval regressions passed |
| S23 Explicit mutation semantics | DONE | S22 checkpoint | V25 and generated/compatibility checks passed |
| S24 Legacy retirement/current workflows | DONE | S23 checkpoint | Current workflow matrix, bootstrap, drift and harness qualification passed |
| S15-R2 Validation and handoff | DONE | All preceding checkpoints | Finalization, compatible acceptance and fresh-environment handoff complete |

Affected S09/S16 specification, S17 mapping-admission, S20 compatibility,
S12-I/V19 migrated-execution and S15/S15-R completion declarations are reopened.
Legacy execution obligations will be marked DROPPED, superseded by the selected
fresh-environment policy; their passing historical evidence is preserved.
F25–F28 implementation coverage unrelated to these findings remains retained.

### Execution and validation contract

Use public Make task-guide/target-plan/test-slice workflows; register authored
verification inputs and regenerate through Make. Required checks include
applicable schema/generated, Go, boundary, compatibility, migration drift,
browser/client and harness checks. Run agent-finalize, review changes, freeze
source, then execute compatible final acceptance. Run lint-markdown at every
checkpoint. Full CI and visual-golden updates are not selected by default.
Record exact commands/selections, source identity, owning exits, run roots,
cleanup, failures/repairs, skipped checks and exit decisions. A required failure
leaves its workstream incomplete. Executable evidence never consumes Markdown.

The final handoff requires a fresh canonical-head database and matching
server/browser publication. Prior pending imports and pre-production database
snapshots are not upgrade inputs. Current-format restart/cancellation/recovery/
replay remain supported. Resubmit original sources when needed. Rollback restores
the previous matching application/database environment; it does not downgrade
the new database. Provisioning/reset, commit and deployment remain separate.

### S21 exit checkpoint

Core 00 §1.5 records the user's explicit fresh-environment/whole-catalog decisions
and supersedes only predecessor compatibility obligations. Core 01's twelve-route
Imports inventory now includes source inspection and all six read-only operations;
it requires explicit boolean mutation semantics for the complete public catalog.
Extensions now agrees with fresh-environment admission and current recovery.

Commands: `make task-guide OWNER=module.imports` (exit 0, diagnostic only);
`CARTULARY_TEST_RUN_ID=nf-s21-markdown-20261029-01 make lint-markdown` (owning exit
0, pass, zero failures, run root `.cartulary/test-results/nf-s21-markdown-20261029-01`,
summary `adhoc/lint-markdown/tool-run-summary.json`). No product checks or fixture
resources were required for this documentation slice. Source is baseline HEAD
with preserved dirty implementation plus the specified owner/tracker edits.
**S21 exit: DONE.** F29–F31 remain open for their implementation slices. Final
checkpoint command: `CARTULARY_TEST_RUN_ID=nf-s21-checkpoint-20261029-01 make lint-markdown`.

### S22 exit checkpoint

`protectMappingSourceTx` now locks the selected source before validating the
returned capability and ownership; ordinary metadata lookups remain non-locking.
Approval retains session/unit/source protection through projection, mapping and
idempotency publication. The registered mapping-source-protection integration
row exercises eight generic/analytical contention, cancellation and lifetime
cases using observable database lock waits. No production test hook was added.

Commands (all owning exit 0):

- `CARTULARY_TEST_RUN_ID=nf-s22-generate-20261029-01 make generate`.
- `CARTULARY_TEST_RUN_ID=nf-s22-format-20261029-01 make format`.
- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s22-selection.json DETAIL=rows JSON=1`.
- `CARTULARY_TEST_RUN_ID=nf-s22-approval-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.mapping_source_protection,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6`.
- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s22-regression-selection.json DETAIL=summary`.
- `CARTULARY_TEST_RUN_ID=nf-s22-regressions-20261029-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.upload_admission_replay_complete_b8684935f2`.

Both test roots under `.cartulary/test-results/` passed 4/4 work units, with
17 cleanup steps each and zero failed/blocked cleanup. Exact `make explain-run
RESULTS_DIR=<root> DETAIL=summary` diagnostics also exited 0. Compatible product
source: `sha256:3bf1e18fd2fa4baa097a25766bd9a15563dfedc7d23ef154a8734d767e318c3d`.
No test failure or repair was needed. Full CI remains unselected; final static
and compatibility obligations remain S23/S15-R2 work. F29/V24 is resolved.
**S22 exit: DONE**, subject to the checkpoint command below before S23 begins.
`CARTULARY_TEST_RUN_ID=nf-s22-checkpoint-20261029-01 make lint-markdown`.

### S23 exit checkpoint

All 130 public operations now declare `x-cartulary-state-changing` (47 read-only,
83 state-changing). Assembly and generation reject missing/non-boolean values;
there is no method fallback. Imports resolves method/classification from generated
operations and rejects missing/duplicate/foreign/unbound entries before serving.
Parity checks now compare the generated flag directly. Compatibility treats the
extension as security-sensitive and the authored 2.0.0 change set is reconciled.

Human annotation review includes the six read-only POSTs (Imports preview and
inspection, Workbook query and locate, composition validation, Evidence metadata),
plus auth-completion GETs and download-handle consumption. Core REQ-01-328a's
read-only Evidence metadata retains its independent cookie-CSRF requirement;
auth callbacks/handle redemption retain their protocol protections. No security
policy was inferred from mutation metadata outside Imports. The narrow operation
catalog test package was added to the closed harness roots, schema, registry and
normative harness inventory; arbitrary tool roots remain excluded.

Generation failures and repairs are preserved: `nf-s23-generate-20261029-01`
failed on stale change-set fingerprints; 02 failed on unsorted test selectors;
03 failed on the unregistered operation-catalog test root; 04 failed on the
remaining runner-registry mirror. Each owning Make exit was 2, underlying step
exit 1. Exact-root `make explain-run` plus bounded referenced logs identified each
cause. Authored inputs were repaired; `nf-s23-generate-20261029-05` passed (exit 0).
No product-test failure occurred. `nf-s23-format-20261029-01` also passed.

`make task-guide OWNER=platform.openapi` and target plans for test-slice inputs
`/tmp/nf-s23-{openapi,imports,workbook,composition}.json` and harness-contract input
`/tmp/nf-s23-harness.json` all exited 0. The following exact invocations use
`CARTULARY_TEST_RUN_ID=<run-id>`; roots are `.cartulary/test-results/<run-id>`.

| Run ID | Make command | Result / cleanup |
| --- | --- | --- |
| nf-s23-openapi-20261029-01 | `make test-slice OWNER=platform.openapi` | PASS 5/5; 14 cleanup steps |
| nf-s23-imports-20261029-01 | `make test-slice OWNER=module.imports ROWS=module.imports.unit.behavior_contract,module.imports.integration.target_neutral_public_source_admission,module.imports.integration.upload_admission_replay_complete_b8684935f2` | PASS 5/5; 18 cleanup steps |
| nf-s23-workbook-20261029-01 | `make test-slice OWNER=module.workbook ROWS=module.workbook.catalog.locator_public_contract,module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120` | PASS 3/3; 14 cleanup steps |
| nf-s23-composition-20261029-01 | `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.public_lifecycle_authority` | PASS 3/3; 14 cleanup steps |
| nf-s23-schema-20261029-01 | `make source-check CHECK=schema` | PASS 3/3; 12 cleanup steps |
| nf-s23-generated-20261029-01 | `make source-check CHECK=generated` | PASS 5/5; 14 cleanup steps |
| nf-s23-go-20261029-01 | `make lint CHECK=go` | PASS 6/6; 15 cleanup steps |
| nf-s23-boundary-20261029-01 | `make lint CHECK=backend-boundary` | PASS 4/4; 13 cleanup steps |
| nf-s23-harness-20261029-01 | `make harness-contract SUITE=contract` | PASS; cleanup checked in exact-root summary |

All owning exits were 0; exact-root summary diagnostics exited 0 and cleanup had
zero failed/blocked steps. Compatible source:
`sha256:86ab1eceab050d310c4383f13b50a1abe6fa3eb224c20d24969b2733b06a8402`.
Full CI remains unselected; final browser/client checks remain S15-R2 obligations.
F30/V25 is resolved. **S23 exit: DONE**, subject to the documentation checkpoint
before S24: `CARTULARY_TEST_RUN_ID=nf-s23-checkpoint-20261029-01 make lint-markdown`.

### S24 implementation and qualification

Migrations 31–33 retain canonical numbering and structural constraints but no
longer reserialize predecessor values, clear old analytical plan metadata or
translate old registry identities. Core 01 REQ-01-620b now states the adopted
fresh-environment boundary directly. The frozen schema-30 fixtures and exclusive
migration-scratch/runtime attachment capability are removed, including their
private fields and tests. General migration isolation and canonical bootstrap
remain. Current-format binding mismatches continue to fail closed.

The replacement workflow matrix obtains source, mapping, selection and apply
admission through public routes on a fresh current database. It covers both
target families across queued/interrupted execution, cancellation, source/mapping/
binding/registry mismatches, generic-row digest failure, analytical execution
without generic-row read privileges, precommit and postcommit faults, restart,
and repeated replay. Nontransactional write witnesses detect attempted duplicate
owner effects. Generic preview asserts exact column/cell members before mapping,
after projection materialization, and after successful restarts. Bootstrap verifies
nullable projection invariants, EXTERNAL storage, byte/digest equality and source
copy behavior; current workflows verify target-dependent plan constraints.

Legacy F28/V23 and the migrated-execution part of V19, S20 and S12-I are
**DROPPED—superseded by the fresh-environment decision**. The compressed
predecessor rewrite and upgrade preservation claims in S12-M/S12-I are likewise
dropped. Their historical runs remain unchanged and do not qualify this build.
Current recovery/replay, bounded readers, byte integrity and canonical migration
isolation remain required, mapped to V26 and final acceptance below.

Commands so far (roots under `.cartulary/test-results/`, exact run ID supplied
through `CARTULARY_TEST_RUN_ID`):

- `nf-s24-generate-20261029-01`: `make generate`, owning exit 0.
- `nf-s24-format-20261029-01`: `make format`, owning exit 0.
- `make task-guide OWNER=module.database_migrations`, exit 0.
- `make target-plan TARGET=test-slice INPUTS=/tmp/nf-s24-current.json DETAIL=summary`
  and `/tmp/nf-s24-migration.json`, both exit 0.
- `nf-s24-bootstrap-20261029-01`: `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap,module.database_migrations.unit.test_harness_targeted_operation_validation`, PASS 8/8 work units, 19 cleanup steps.
- `nf-s24-drift-20261029-01`: `make migration-drift`, PASS 8/8, 19 cleanup steps.
- `nf-s24-go-20261029-01`: `make lint CHECK=go`, PASS 6/6, 15 cleanup steps.
- `nf-s24-current-20261029-01`: `make test-slice OWNER=module.imports ROWS=module.imports.integration.current_workflows,module.imports.integration.current_cancellation`, FAIL 2/3 work units; Make exit 2, runner exit 10, product/test_assertion_failure; 14 cleanup steps all completed. Both binding-change cases expected target unavailability instead of the existing identity-mismatch `source_changed` result. The runtime failed closed correctly; the test expectation was repaired. Cancellation passed. No deadlock or runtime repair occurred.

The first four qualification roots use source
`sha256:11be81376fce1b8b4ba8131aadeac54138ee63d0cbda87d38a0dc693c0b9c5d3`.
Successful checks had owning exit 0 and no failed/blocked cleanup. Exact-root
`make explain-run RESULTS_DIR=<root> DETAIL=summary` diagnosed each result with
exit 0. The repaired workflow rerun is `nf-s24-current-20261029-02`, using the
same exact command and source
`sha256:12915ad34d9d20eda9ffc679e76da01dcd0c2f14718b91806d1c67791497e0fc`.
The repaired run passed all 3 work units with owning exit 0, 14 completed cleanup
steps and zero failed/blocked cleanup. Both registered rows passed, covering all
26 scenarios. The retirement scan found no schema-30 workflow, translation or
runtime-attachment references in affected active code, migrations and routing;
`git diff --check` passed. No required S24 check remains failed. F31/V26 is
resolved. **S24 exit: DONE**, subject to the following checkpoint before S15-R2:
`CARTULARY_TEST_RUN_ID=nf-s24-checkpoint-20261029-01 make lint-markdown`.
Full CI is unselected; final static/browser and compatible regression evidence
remain S15-R2 work.

### S15-R2 finalization, frozen acceptance and handoff

S21, S22, S23 and S24 Markdown checkpoint commands all passed with owning exit 0
and zero failures. Each checkpoint completed before its successor began.

`CARTULARY_TEST_RUN_ID=nf-s15r2-finalize-20261029-01 make agent-finalize`
passed with owning exit 0, 1/1 work units and 10 completed cleanup steps.
Generated changes were reviewed against the pre-finalization diff/status;
finalization introduced no additional changes. Source was then frozen:

- HEAD: `b6d073974802489edfeb892d39f48b128b9d00c4`, dirty working tree preserved.
- Source: `sha256:12915ad34d9d20eda9ffc679e76da01dcd0c2f14718b91806d1c67791497e0fc`.
- Toolchain: `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`.
- Environment: `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

The S24 current-workflow rerun has exactly these identities and remains compatible
acceptance for its 26 scenarios. No product repair followed it. Final source/diff
review found no drift. Subsequent tracker edits are documentation only.

`make task-guide OWNER=module.networkflow` and `OWNER=module.recovery` exited 0.
Every final selection below was resolved successfully with
`make target-plan TARGET=<target> INPUTS=/tmp/nf-s15r2-<selection>.json DETAIL=summary`,
using the exact inputs reproduced in its command. Selection names are the run-ID
portion between `nf-s15r2-` and `-20261029-01`. Planning is not test evidence.

Each command below was executed with `CARTULARY_TEST_RUN_ID=<run-id>`.
Every owning exit was 0. Roots are `.cartulary/test-results/<run-id>`;
all selected work units passed, with zero failed/skipped/cancelled units and
zero failed/blocked cleanup steps. Work-unit counts are not test counts.

| Run ID | Exact Make command and selection | Work units / completed cleanup |
| --- | --- | --- |
| nf-s15r2-imports-20261029-01 | `make test-slice OWNER=module.imports ROWS=module.imports.unit.behavior_contract,module.imports.integration.mapping_source_protection,module.imports.integration.upload_admission_replay_complete_b8684935f2,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.apply_transaction_current_state_bc8ad0ac32,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6,module.imports.integration.xlsx_operator_region_durable_replay_7b0a58c912,module.imports.integration.xlsx_import_discovery_accepts_a_bounded_open_xml_d7c25ef6c8,module.imports.integration.csv_import_discovery_exhaustive_mapping_approval_fc99e8ea88,module.imports.integration.target_neutral_public_source_admission,module.imports.integration.network_flow_default_limit_100001_rows,module.imports.unit.target_neutral_csv_source_admission,module.imports.unit.canonical_write_null_and_safe_owner_translation_rs08_3f58f0c338` | 7/7; 24 steps |
| nf-s15r2-nf-20261029-01 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700,module.networkflow.unit.network_flow_selector_covers_diagnostic_discover_6f2d284ad8,module.networkflow.unit.source_inspection_header_only` | 1/1; 10 steps |
| nf-s15r2-openapi-20261029-01 | `make test-slice OWNER=platform.openapi` | 5/5; 11 steps |
| nf-s15r2-migration-20261029-01 | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.schema_bootstrap,module.database_migrations.unit.test_harness_targeted_operation_validation` | 8/8; 19 steps |
| nf-s15r2-recovery-20261029-01 | `make test-slice OWNER=module.recovery ROWS=module.recovery.integration.database_coverage_rejects_an_unclassified_recove_a609c7d3a3` | 3/3; 14 steps |
| nf-s15r2-browser-nf-20261029-01 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_importing_nf_fix_00_094b0f779b` | 12/12; 23 steps |
| nf-s15r2-browser-imports-20261029-01 | `make test-slice OWNER=module.imports ROWS=module.imports.browser.claimed_import_assistant_production_workflow_986c2117ee` | 12/12; 23 steps |
| nf-s15r2-harness-20261029-01 | `make harness-contract SUITE=contract` | 13/13; 22 steps |
| nf-s15r2-workbook-20261029-01 | `make test-slice OWNER=module.workbook ROWS=module.workbook.catalog.locator_public_contract,module.workbook.integration.registry_derived_query_and_create_surface_coverage_7bd7f5a120` | 3/3; 14 steps |
| nf-s15r2-composition-20261029-01 | `make test-slice OWNER=module.reportcomposition ROWS=module.reportcomposition.integration.public_lifecycle_authority` | 3/3; 14 steps |
| nf-s15r2-web-nf-20261029-01 | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 2/2; 11 steps |
| nf-s15r2-web-imports-20261029-01 | `make test-slice OWNER=module.imports ROWS=module.imports.frontend.lifecycle,module.imports.frontend.typed_transport` | 3/3; 12 steps |
| nf-s15r2-go-20261029-01 | `make lint CHECK=go` | 6/6; 15 steps |
| nf-s15r2-types-20261029-01 | `make lint CHECK=types` | 2/2; 11 steps |
| nf-s15r2-backend-boundary-20261029-01 | `make lint CHECK=backend-boundary` | 4/4; 13 steps |
| nf-s15r2-frontend-boundary-20261029-01 | `make lint CHECK=frontend-boundary` | 2/2; 11 steps |
| nf-s15r2-protocol-20261029-01 | `make lint CHECK=protocol` | 4/4; 13 steps |
| nf-s15r2-schema-20261029-01 | `make source-check CHECK=schema` | 3/3; 12 steps |
| nf-s15r2-generated-20261029-01 | `make source-check CHECK=generated` | 5/5; 14 steps |
| nf-s15r2-drift-20261029-01 | `make migration-drift` | 8/8; 19 steps |

For each exact root, `make explain-run RESULTS_DIR=<root> DETAIL=summary`
exited 0. Run manifests agree on the frozen identities; command-specific graph,
row, fixture and case identities remain in their original manifests. This is
focused acceptance of the stated obligations, not a full-CI or whole-repository
conformance claim. Compatible cache reuse, where admitted by the harness, is not
claimed as a fresh performance sample.

### IMP-05 finding and acceptance reconciliation

| Finding / reopened claim | Disposition |
| --- | --- |
| F29 / V24 / S17 source approval | CLOSED. Protected capability comparison and complete association checks precede durable approval. Generic/analytical race, unchanged-content contention, cancellation, lock lifetime and replay cases pass, including no projection/mapping/selection/idempotency effects after rejection. |
| F30 / V25 / S09/S16 specification closure | CLOSED. Core inventories twelve Imports operations with six read-only and six mutations. All 130 public operations declare boolean semantics; assembly/generation fail closed, generated flags feed Imports, independent CSRF requirements remain explicit, and security-sensitive compatibility accounting passes. |
| F31 / V26 / S20 compatibility | CLOSED under the adopted fresh-environment boundary. Old translation, frozen workflows and exclusive runtime attachment are absent. Fresh bootstrap, source storage/copy, current public preview and both target families' execution/cancellation/recovery/replay pass. |
| F28 / V23 / legacy part of V19/S12-I/S20 | DROPPED—superseded by the fresh-environment decision. Historical migrated-workflow runs remain historical; no prior-format upgrade guarantee is carried forward. |
| Current V19/S12-I and S14 guarantees | CONFIRMED. Analytical execution succeeds without generic-row SELECT authority; current binding/source/mapping/row-digest changes fail closed; atomic rollback and committed-result recovery/replay preserve one set of owner effects. |
| F25–F27 / S13 and affected S14 regressions | CONFIRMED. Source admission, strict diagnostic admission, borrowed transaction ownership, 50/51 preview scope and full apply, default 100001-row public import, generic/XLSX paths and both browser workflows pass. |
| S15/S15-R/S15-R2 final acceptance | CLOSED by compatible focused evidence and this operational handoff, subject to the final Markdown record check below. |

### Fresh-environment operational handoff

The revised application requires a **fresh database initialized to canonical
head through the normal migration chain**. Earlier pre-production databases,
pending import plans and snapshots are not upgrade inputs for this build.
Original sources must be resubmitted through current upload, mapping approval
and apply when prior development results are needed. Existing development data
is neither reset nor automatically deleted by these changes.

Publish matching server/browser assets together at Network Flow major 8/state 5
and the existing coordinated API candidate. Do not mix old/new workers or point
this application at a predecessor database. Canonical bootstrap and migrations
31–33 establish source storage and current plan/projection constraints; they do
not perform historical plan conversion. Current-format restart, cancellation,
Recovery and replay remain supported. Recovery after a committed unit does not
repeat owner writes.

Upload buffering, generic Workbook materialization and Recovery-wide memory are
outside the bounded analytical preview/apply claim. CSV interpretation changes
require original-source re-import when different results are desired. No existing
Network Flow table is reinterpreted by this implementation.

Rollback restores the previous matching application/database environment snapshot;
it never downgrades the new database after writes or deletes metadata to force
admission. Provisioning, database resets, commit and deployment are separate
unperformed operational actions.

All selected engineering checks passed. Full CI, visual-golden updates,
independent performance/release qualification and retained warm-check maintenance
were not selected; no compatible full warm-check root was supplied to finalization.
No required check was skipped, and no engineering blocker remains. Historical
failures and their repairs remain recorded. Final record command:
`CARTULARY_TEST_RUN_ID=nf-s15r2-checkpoint-20261029-01 make lint-markdown`.

The final record check passed with owning exit 0 and zero failures; its retained
summary is `nf-s15r2-checkpoint-20261029-01/adhoc/lint-markdown/tool-run-summary.json`.
A whitespace review found and removed one extra blank line at this document's
EOF. Final documentation-only confirmation:
`CARTULARY_TEST_RUN_ID=nf-s15r2-checkpoint-20261029-02 make lint-markdown`.
**S15-R2 exit: DONE.** All implementation, validation, tracker reconciliation
and operational handoff obligations are complete. Commit and deployment remain
unperformed.

## 24. IMP-06 resource safety and security accounting

Baseline: `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, with the existing
IMP-05 dirty tracked and untracked work preserved. Full file fingerprints,
prior working-tree patch and index patch are retained outside the repository at
`/tmp/cartulary-nf-remediation-h8v4nkoe`. This invocation implements the user-approved
plan; the earlier planning-only restriction is superseded. No commit, deployment,
database reset or historical data conversion is included.

| Finding | Areas / remediation | Benefit and unresolved risk | Compatibility / validation |
| --- | --- | --- | --- |
| F32 / P1 | Core owners, shared lexer, Imports preview/materialization/apply, client contracts and tests: incremental preview, sparse actual cells, cancellation | Prevent rectangular memory amplification and fabricated source values; unresolved small inputs can exhaust memory | Missing trailing CSV fields are absent; coordinated fresh-environment API/client candidate; V27 |
| F33 / P2 | Common API owner, semantic compatibility traversal, end-to-end tests and release review records | Security accounting follows semantic surfaces even under structural changes; unresolved declarations bypass review | Additional exact security fingerprints in existing 2.0.0 candidate; immutable release baseline retained; V28 |

| Workstream | Status | Depends on | Risk / binary exit |
| --- | --- | --- | --- |
| S25 Owner and tracker closure | DONE | authorized plan | Adopt concrete coordinated owners; Markdown checkpoint passes |
| S26 Characterization and registration | DONE | S25 | Small regressions expose both defects without oversized baseline allocation; registered Make coverage |
| S27 CSV architecture and consumers | DONE | S26 | V27: sparse geometry, cancellation, source protection, mapping, replay and browser coverage pass |
| S28 Security accounting | DONE | S27 | V28: Compare matrix, exact review records and production gate pass |
| S15-R3 Frozen acceptance and handoff | DONE | S27, S28 | Finalize, freeze source, compatible required portfolio, successful cleanup and final record |

V27 requires the 10,000-column / 10,000 one-field-row case, first-50 retention,
late width/framing, explicit empty versus absent, incremental cancellation,
public source/mapping/replay and client alignment. The isolated builder allocation
budget is 64 MiB per operation, excluding fixture/harness setup; it is not a runtime
admission limit. V28 requires whole-path, operation and property add/remove/change,
false and empty-array presence, root/operation security, schemes, replacement,
escaping, deterministic deduplication, descriptive/unrelated-property negatives,
and version/change-set rejection. Tests and release records never consume Markdown.

The read-only planning observations of `nf-s15r2-imports-20261029-01`,
`nf-s24-current-20261029-02`, `nf-s15r2-openapi-20261029-01`, and
`nf-s15r2-browser-imports-20261029-01` report historical passes. Their source is
`sha256:12915ad34d9d20eda9ffc679e76da01dcd0c2f14718b91806d1c67791497e0fc`;
those observations did not audit current source or observe owning exits.
They do not cover the new findings. Earlier outer exits and cleanup remain
recorded in §23, without alteration.

### IMP-06 execution record

| Workstream / event | Change or evidence | Result / next gate |
| --- | --- | --- |
| S25 start | User policy selections and implementation request recorded in Core 00 §1.6; Core 01 separates inspection/materialization and adopts absent CSV fields; common API owns nested security accounting; Core 04 limits are explicit about XLSX | Owner changes prepared under the approved concrete plan; Markdown checkpoint next |
| S25 checkpoint | `CARTULARY_TEST_RUN_ID=nf-s25-markdown-20261029-01 make lint-markdown` | PASS, owning exit 0; summary `.cartulary/test-results/nf-s25-markdown-20261029-01/adhoc/lint-markdown/tool-run-summary.json` |
| S26 regressions | `make test-slice OWNER=module.imports ROWS=module.imports.unit.target_neutral_csv_source_admission` at `nf-s26-imports-baseline-20261029-01`; `make test-slice OWNER=platform.openapi ROWS=platform.openapi.unit.release_compatibility` at `nf-s26-openapi-baseline-20261029-01` | Both expected product/test_assertion_failure, owning exits 2; 10 cleanup steps each, none failed/blocked. Small ragged rows expose padding; Compare exposes missing nested declarations. Exact-root summary diagnosis exits 0. The oversized fixture was not executed against vulnerable code. |
| S26 checkpoint | `CARTULARY_TEST_RUN_ID=nf-s26-markdown-20261029-01 make lint-markdown` | PASS, owning exit 0 |
| S27 initial focused checks | `nf-s27-unit-20261029-01` and `nf-s27-tabular-20261029-01` | Both product assertion failures, owning exits 2; all 10 cleanup steps passed each. Ragged resource test passed (preview 4,759,088 allocated bytes; approval 12,709,448). Corrected a non-clearable test field and restored exact clipboard lexical error text; no limit was relaxed. |
| S27 browser diagnosis | `nf-s27-browser-20261029-01`, Imports browser/lifecycle/typed-transport selection | Product assertion failure, owning exit 2; 12/14 work units passed, all 25 cleanup steps completed. Existing transport validation still required rectangular CSV rows, hiding the mapping UI. Corrected CSV-only sparse admission and added ordinal/width/empty-row negatives; XLSX remains rectangular. |
| S27 checkpoint | Unit/behavior `nf-s27-unit-20261029-02`; shared kernel `nf-s27-tabular-20261029-02`; seven integration rows `nf-s27-integration-20261029-01`; browser/transport `nf-s27-browser-20261029-02`; Markdown `nf-s27-markdown-20261029-01` | PASS, all owning exits 0. Public sparse preview→approval→apply and restart/replay passed; 17 integration cleanup steps completed. Browser confirms absent fields render without padding. Focused evidence precedes final frozen acceptance. |
| S28 accounting review | `nf-s28-generate-20261029-01` and explicit report `nf-s28-accounting-20261029-01` | Expected gate failure (owning exits 2): previous change set covered 559 of 600 records. Reviewed 40 newly visible security records for 20 existing operation additions against their owner inputs and prior structural rationales, plus the sparse-preview description. Preserved all 559 previous approvals, immutable 1.0.0 release and 2.0.0 candidate; added exact owner/value-specific records. All 13 compatibility-run cleanup steps completed. |
| S28 checkpoint | `nf-s28-generate-20261029-02`, `nf-s28-openapi-20261029-01` | PASS, owning exits 0. Canonical generation and complete OpenAPI owner selection cover nested Compare matrices, false-positive exclusions, exact review/version negatives, assembly, runtime parity, and production release gate. |

The final frozen portfolio is gap-scoped: complete Tabular Ingest and OpenAPI
owner selections; Imports resource/behavior/nullability unit rows, public source
admission, mapping/source protection, current-workflow recovery/cancellation,
CSV/XLSX/region/reimport controls, the public 100,001-row Network Flow admission
case, Imports browser/transport/lifecycle, and Network Flow preview/source
inspection plus its discovery browser workflow. This is not a new complete
Network Flow owner qualification or a full CI/performance/visual-baseline claim.
Final source/generated/schema/catalog, Go/type/protocol/boundary lint and Markdown
checks complete the portfolio. Any source repair invalidates the frozen cohort;
its affected acceptance must be re-established before closure.

### IMP-06 frozen validation and handoff

Frozen executable source:
`sha256:78e53e3695ccbdca74cfd3bde529c372319af6373e565cce90b1320f8a0b335a`.
Every final product/source-check/lint/compatibility run below reports this same
identity. Native owning exits were observed as 0, independently of summary
status. Final main product selections have no missing, skipped or failed selected
rows; all published cleanup steps completed. OpenAPI reused five compatible
cached work units; its complete-owner audit validates retained selection,
provenance, current source and cleanup. Tabular Ingest's complete-owner audit also
passes. Imports and Network Flow are deliberately exact-row selections, not
complete-owner audit claims.

All run roots below are under `.cartulary/test-results/`. The retained
`run-manifest.json` contains each exact command input/row selection;
`run-summary.json`, `target-summaries/`, `rows/`, and `cleanup-results.json` retain
outcomes and cleanup. No Markdown is consumed as executable evidence.

| Command / scope | Run ID | Observed result |
| --- | --- | --- |
| `make agent-finalize` | `nf-s15r3-finalize-20261029-01` | PASS; generation/finalization completed before frozen acceptance |
| `make test-slice OWNER=module.imports ROWS=…` (17 rows: the five unit, nine integration and three client/browser rows in its manifest) | `nf-s15r3-imports-20261029-01` | PASS; 17/17 rows, 19/19 work units; 36 cleanup steps completed |
| `make test-slice OWNER=module.networkflow ROWS=…` (three preview/inspection unit rows and production discovery browser row) | `nf-s15r3-networkflow-20261029-01` | PASS; 4/4 rows; 24 cleanup steps completed |
| `make test-slice OWNER=module.tabularingest` | `nf-s15r3-tabular-20261029-01` | PASS; both complete-owner rows; 10 cleanup steps completed |
| `make test-slice OWNER=platform.openapi` | `nf-s15r3-openapi-20261029-01` | PASS; all six owner rows; five compatible cache hits; nine cleanup steps completed |
| `make test-evidence-audit OWNER=module.tabularingest EVIDENCE_ROOTS_FILE=…` | `nf-s15r3-tabular-audit-20261029-01` | PASS complete-owner current-source audit; manifest `/tmp/cartulary-nf-remediation-h8v4nkoe/tabular-evidence.json` |
| `make test-evidence-audit OWNER=platform.openapi EVIDENCE_ROOTS_FILE=…` | `nf-s15r3-openapi-audit-20261029-01` | PASS complete-owner current-source audit; manifest `/tmp/cartulary-nf-remediation-h8v4nkoe/openapi-evidence.json` |
| `make source-check CHECK=generated` | `nf-s15r3-generated-20261029-01` | PASS |
| `make source-check CHECK=schema` | `nf-s15r3-schema-20261029-01` | PASS |
| `make source-check CHECK=catalog` | `nf-s15r3-catalog-20261029-01` | PASS |
| `make lint CHECK=go` | `nf-s15r3-go-20261029-01` | PASS |
| `make lint CHECK=types` | `nf-s15r3-types-20261029-01` | PASS |
| `make lint CHECK=protocol` | `nf-s15r3-protocol-20261029-01` | PASS |
| `make lint CHECK=backend-boundary` | `nf-s15r3-backend-boundary-20261029-01` | PASS |
| `make lint CHECK=frontend-boundary` | `nf-s15r3-frontend-boundary-20261029-01` | PASS |
| `make openapi-compatibility-check` | `nf-s15r3-compatibility-20261029-01` | PASS; exact 600-record review set, immutable baseline 1.0.0, candidate 2.0.0 |
| `make lint-markdown` | `nf-s15r3-markdown-20261029-01` | PASS, owning exit 0; final handoff documentation checkpoint |

F32/V27 closes the reopened generic resource portion of F25/V20: preview retains
only header/column metadata and the first 50 actual rows, scans the remainder for
geometry/framing/integrity, and materializes approval separately. Missing trailing
fields never become blank cells or unknown-value warnings; explicit empty fields
retain their mapped empty-value policy. The browser consumes sparse rows, while
XLSX remains rectangular. The 10,000-column/10,000-row adversarial regression
passes the separate 64 MiB builder budgets. Focused measured totals were about
4.8 MB preview and 12.7 MB materialization; these are fixture allocation checks,
not a production RSS bound or performance qualification. Admission remains
byte-only for CSV, including the passing public 100,001-row analytical case.

F33/V28 closes the reopened F30/V25/S23 accounting portion: structural records
remain, while added/removed/replaced paths and operations expose exact nested
mutation/security declarations. Root requirements and security schemes receive
the same treatment. Structural matching excludes unrelated schema properties and
examples, preserves descriptive-only changes, and supports reusable path items,
webhooks and callbacks. Compare tests establish absence versus false/empty,
escaping, deterministic unique records, version policy and exact review rejection.
The 40 newly visible security entries are accounting for already adopted
operations, not new route behavior.

Compatibility handoff: ship the coordinated server/client candidate together.
Missing trailing CSV fields are intentionally absent; there is no rectangular
compatibility mode. The established fresh-database migration boundary, Network
Flow major 8/state 5 and API candidate 2.0.0 remain in force. Old pre-production
imports require explicit reimport; no historical projection converter or extra
major is introduced. No database reset, commit, deployment or release publication
was performed. Original dirty work and index were preserved; the six paths absent
at entry remain absent. The task delta is 32 files, including authored owners,
tests, documentation and canonical generated projections.

Full CI, complete Network Flow owner qualification, visual-golden refresh,
performance publication and deployment are outside this gap-scoped acceptance.
Historical passes and failed intermediate runs remain available with their
original classifications. All required product and documentation checks passed;
none were skipped. The acceptance claims reopened by IMP-06 are closed within
the frozen gap-scoped portfolio above.

## 25. IMP-07 descriptive security changes and preview retention

Baseline: `main` at `b6d073974802489edfeb892d39f48b128b9d00c4`, with existing dirty
tracked/untracked work and index preserved. Entry fingerprints, source copies,
and prior/index patches are retained at `/tmp/cartulary-imp07-kmxt5qho`.
The maintainer authorized implementation after selecting affected-gap final
qualification. Core 01 §17.1 and Core 04 acceptance text clarify the already
adopted descriptive-change and bounded-preview rules; they do not introduce a new
runtime profile, OAuth authentication capability, CSV admission limit, migration,
API version or state converter. This implementation request supersedes the earlier
planning-only boundary.

| Finding / obligation | Remediation / areas | Rationale, benefit and unresolved risk | Compatibility / exit |
| --- | --- | --- | --- |
| F34 / V28 | Clarify owner; make security classification change-aware; correct and expand Compare tests; reconcile tracker | Distinguish scope identity from its description without weakening security additions/removals; otherwise prose edits require major releases and a wrong test expectation becomes policy | No public interface change; description-only fixture changes become non-behavioral. Production has no OAuth2 scheme; the existing 600-record report/change set must remain unchanged |
| F35 / V27 | Clarify acceptance; add narrow 10,000/100,000-row preview tests with a 1 MiB per-call allocation bound and a full-materialization negative control; register tests and document evidence | Separate bounded-retention protection from the existing 64 MiB dense-padding guard; otherwise full sparse materialization followed by truncation would pass | No planned parser/runtime change. Exact first-50 results, full-tail validation and discriminating allocation bounds must pass |

| Phase | Status | Depends on | Principal risk / binary exit |
| --- | --- | --- | --- |
| S29 Owner and tracker clarification | DONE | authorized plan | Clarified owners; Markdown checkpoint passed; earlier evidence preserved |
| S30 Characterization | DONE | S29 | Corrected OAuth expectations failed on baseline; preview passed and full-materialization control exceeded its bound; selectors regenerated |
| S31 Classifier correction and focused checks | DONE | S30 | Exact structural/type/change-kind guards implemented; both focused selections passed |
| S15-R4 Frozen qualification and handoff | DONE | S31 | Finalized and frozen; affected-gap portfolio, complete OpenAPI audit, exact production report equality, cleanup and documentation checkpoint passed |

V28 covers all four OAuth flow kinds; ordinary, reserved-looking and escaped
scheme/scope names; empty descriptions; scope add/remove/rename; malformed value
replacements; flow/requirement/scheme changes; structural records and exact
version/change-set enforcement. V27 retains the 10,000-wide ragged test and adds
1 MiB narrow-preview budgets at 10,000 and 100,000 data rows. A separate 10,000-row
full-materialization control must exceed 1 MiB. Fixture construction/assertions
are outside sequential TotalAlloc measurements; no timing/RSS gate, GC suppression
or runtime quota is introduced. Tail width, malformed quote, terminal I/O/integrity
failure and cancellation must still prevent a truncated scan from passing.

The final portfolio is complete `platform.openapi`, the exact Imports
`module.imports.unit.target_neutral_csv_source_admission` row, compatibility,
generated/schema/catalog checks, Go/backend-boundary lint, Markdown, and the
complete OpenAPI owner evidence audit. The broader S15-R3 passes remain historical
at `sha256:78e53e3695ccbdca74cfd3bde529c372319af6373e565cce90b1320f8a0b335a`.
No browser/integration rerun, complete Imports/Network Flow owner qualification,
performance publication, commit, database reset or deployment is part of IMP-07.
Tests, generators and release evidence must not consume Markdown.

### IMP-07 execution record

| Event | Command / evidence | Result / next gate |
| --- | --- | --- |
| Entry | Existing source and owner guidance inspected; immutable compatibility baseline copied; `make task-guide OWNER=platform.openapi` and `OWNER=module.imports` | Read-only discovery passed; owner/tracker clarification prepared |
| Owner checkpoint | `make lint-markdown`, `nf-imp07-owners-markdown-20261030-01` | Passed, outer exit 0 |
| Characterization preparation | `make generate`, `nf-imp07-characterization-generate-20261030-01` | Passed, outer exit 0; generated topology render index updated through Make |
| Expected baseline failure | `make test-slice OWNER=platform.openapi ROWS=platform.openapi.unit.release_compatibility`, `nf-imp07-characterization-openapi-20261030-01` | Product assertion failure, outer exit 2; existing scope descriptions incorrectly classified security-sensitive. One failed work unit; 10 cleanup steps succeeded. Failure retained, not acceptance evidence |
| Retention characterization | `make test-slice OWNER=module.imports ROWS=module.imports.unit.target_neutral_csv_source_admission`, `nf-imp07-characterization-imports-20261030-01` | Passed, outer exit 0; narrow previews stayed within 1 MiB, full 10,000-row materialization exceeded 1 MiB, complete-tail checks passed; 10 cleanup steps succeeded |
| Correction and formatting | Change-aware classifier and structural/type guards; `make format`, `nf-imp07-format-20261030-01` | Passed, outer exit 0; no CSV runtime change or release approval refresh |
| Focused checks | Same two exact row selections, `nf-imp07-focused-openapi-20261030-01` and `nf-imp07-focused-imports-20261030-01` | Both passed, outer exits 0; ready for finalization |

Run IDs above resolve beneath `.cartulary/test-results/`. Characterization source
digest: `sha256:2a196e58529817829c684212c2c66c12a15f562aed87dc40211130d2987f61bc`.
`make explain-run ... DETAIL=summary` confirmed both characterization outcomes;
`DETAIL=logs TARGET=test-slice` confirmed the expected comparator assertions.
An initial observer-only `DETAIL=logs LIMIT=8` request was rejected because LIMIT
belongs to critical-path queries; removing LIMIT succeeded. Neither observer
request reran product tests or changed the retained results.

### S15-R4 frozen affected-gap portfolio

`make agent-finalize` passed at `nf-imp07-finalize-20261030-01`, outer exit 0,
with 10 successful cleanup steps. Source was then frozen at
`sha256:e9d96b0779b713b92a9ea8a3ad18789584ce621c254a8056b18d3990268330d8`;
toolchain `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system `sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`
match throughout. No compatible successful full warm-check root was available;
retained warm-check maintenance was skipped, without launching a redundant check.

All runs below passed with outer exit 0 and zero failed/blocked cleanup steps.
Their roots are `.cartulary/test-results/nf-imp07-final-<suffix>-20261030-01`.
The work-unit counts are execution units, not test counts.

| Command / selection | Suffix | Passed work units | Successful cleanup steps |
| --- | --- | --- | --- |
| `make test-slice OWNER=platform.openapi` (all six rows) | `openapi` | 5 | 13 |
| `make test-slice OWNER=module.imports ROWS=module.imports.unit.target_neutral_csv_source_admission` | `imports` | 1 | 9 |
| `make openapi-compatibility-check` | `compatibility` | 4 | 13 |
| `make source-check CHECK=generated` | `generated` | 5 | 14 |
| `make source-check CHECK=schema` | `schema` | 3 | 12 |
| `make source-check CHECK=catalog` | `catalog` | 2 | 11 |
| `make lint CHECK=go` | `go` | 6 | 15 |
| `make lint CHECK=backend-boundary` | `backend-boundary` | 4 | 13 |
| `make test-evidence-audit OWNER=platform.openapi EVIDENCE_ROOTS_FILE=/tmp/cartulary-imp07-kmxt5qho/openapi-evidence.json` | `openapi-audit` | 1 | 10 |

Canonical exact-root `make explain-run ... DETAIL=summary` observations confirmed
these outcomes, source identities and cleanup. Both characterization rows and
the corrected focused comparator row executed fresh. The final OpenAPI selection
executed four work units and accepted one valid cached comparator unit; the final
Imports selection accepted its valid cached unit. The audit accepted all six
OpenAPI rows at the frozen source, with no rejected or unused input roots. Imports
is an exact-row qualification, not a complete-owner audit.

The production compatibility report remains exactly equal to the entry report:
600 records, identical classifications, pointers, hashes and fingerprints. The
2.0.0 change set and immutable released baseline remain unchanged; no blanket
reapproval, version increment or migration was needed. Preview runtime source is
unchanged. The source review found only the intended 11 changed/new files relative
to entry, including the Make-generated topology render index. Existing unrelated
tracked/untracked contents and the index remain preserved; executable fingerprints
still match the post-finalize snapshot in `/tmp/cartulary-imp07-kmxt5qho`.

V28's follow-up implementation accepts only string-to-string `change` records at
exact OAuth scope locations in the four defined flows. Tests cover scope identity
changes, reserved/escaped names, empty text, invalid types, unknown flows, nested
non-description values, endpoints, whole schemes, patch-versus-major version
policy and exact review failures. Prior nested operation/security requirements,
uniqueness, determinism and false-positive tests still pass. V27's new independent
1 MiB guard passes for 10,000 and 100,000 rows, and the 10,000-row full-materialization
control exceeds the bound. Full-tail width, malformed quote, terminal integrity
failure and cancellation checks pass alongside the original wide-ragged test.
The bound is fixture-specific allocation evidence, not a universal memory guarantee.

**S15-R4 exit: DONE.** F34/V28 and F35/V27 are closed within this handoff scope.
S15-R4 supersedes only the affected descriptive-classification and retention
acceptance claims. Earlier S15-R3 integration/browser and owner evidence remains
historical at its recorded source. Broader browser/integration/Network Flow/Tabular
qualification was not repeated because IMP-07 changes the comparator and resource
tests, not those runtime surfaces. No CI-wide/current complete Imports or Network
Flow owner claim, publication, commit, database reset or deployment is made.

Final documentation checkpoint: `CARTULARY_TEST_RUN_ID=nf-imp07-final-markdown-20261030-01 make lint-markdown`
passed, outer exit 0. The status-only closure above is followed by the same
command with run ID `nf-imp07-final-markdown-20261030-02`; its terminal result is
retained in that root's `adhoc/lint-markdown/tool-run-summary.json`.

## 26. DOC-03 — Legacy cleanup and import hardening

### 26.1 Scope, baseline and authorization

DOC-03 implements the maintainer-approved **tracker update**, not S32–S36.
The refactor-tracker skill and repository planning framework govern this
handoff; adopted owners govern future behavior. The selected audience is the
next implementation agent and its reviewer. The intended improvement is one
current import-validation path, an explicit normalized-row write boundary,
and trustworthy evidence for the affected import fixtures.

| Scope item | Current decision |
| --- | --- |
| Supplied tracker / canonical target | `docs/handoffs/networkflow-module-refactor-tracker.md`; existing contents establish `internal/modules/networkflow`, resolving inside `/home/jochi/code/cartulary`; label `networkflow` |
| Output admission | Existing regular file, link count 1; no symlink in output or parent components; target resolves inside the repository |
| Framework | `docs/handoffs/cartulary_modular_refactor_planning_framework.md`, available and inspected; doctrine, not behavior authority |
| Artifact mode and coverage | Normal bounded iteration plan; whole-module behavioral inspection remains partial |
| Allowed change | This tracker only; the explicitly requested Markdown check may produce its ordinary cache/result artifacts |
| Implementation authorization | **Not granted by this planning invocation**; S32–S36 require a later authorized task |
| Chosen scope | Preserve adopted features; focused acceptance; remediate only bundles 001–005, 020, 021 and 023 in this iteration |
| Non-goals | Feature retirement, full fixture-corpus overhaul, broad package splitting, generic validation framework, new compatibility aliases, schema/state/public-major changes, database conversion/reset, commit or deployment |
| Supersession | DOC-03 controls future planning. IMP-07 and earlier completions retain only their recorded source/coverage; F36–F38 do not erase those results or establish new runtime failures |

| Baseline item | Starting evidence / identity | Final comparison / outcome | Limits |
| --- | --- | --- | --- |
| Repository / branch / commit | `/home/jochi/code/cartulary`; `main`; `98adcc82961456bcb74bbecd0023136485afbf00` | Unchanged; E24 | Revalidate at the next invocation |
| Index, unstaged and untracked content | Clean index/worktree; no non-ignored untracked files; 7,653 tracked/non-ignored paths fingerprinted with file identity and index-byte digest | Only tracker content differs; other path contents/identities and index bytes unchanged; no added/removed paths, E24 | Comparison material outside repository |
| Tracker identity/content | Initial SHA-256 `70c567e4541b1896d69a7aaba4ad65843d32ed455a69df087ed855179ac0228c`; canonical regular file, link count 1 | Rechecked unchanged immediately before DOC-03 write; final content intentionally differs | Original retained outside repository |
| Retained comparison material | `/tmp/cartulary-nf-doc03-6fj3ffog/baseline.json` and `tracker-before.md` | `preservation-before-checkpoint.json` records the passing comparison; final recheck uses `preservation-final.json`, E24 | Local session evidence, not a product input |
| Relevant ignored outputs | Markdown lint uses `.cache/cartulary/static-analysis` and its exact `.cartulary/test-results/<run-id>/adhoc/lint-markdown` root | Ordinary declared lint outputs admitted by the approved plan | No whole ignored-tree preservation claim; no product/service invocation |
| Prior-baseline comparison | Planning and write-session HEAD/tracker fingerprints agree; the current commit contains the prior import remediation | Revalidated selected production bodies, tests, owners and command routing | Older source digests/results remain historical, not current passes |

### 26.2 Current inventory and inspection depth

Universe: tracked and non-ignored untracked entries under the literal target.
There are 138 tracked authored Go files: 124 at the package root and 14 in
`harnesscontrol`; 80 non-test files and 58 test files. No generated, vendored,
binary, symlink or submodule entry was found in this target universe. Ignored
build/cache outputs and the repository outside this target are not part of its
file count. No new exhaustive caller-graph or conformance claim is made.

The following ledger accounts for every current file: 23 directly inspected
through relevant bodies or assertions, and 115 metadata-only. Direct inspection
includes the preceding planning exchange at the identical source baseline;
it does not claim every line was audited. Metadata-only files retain their
historical §2 descriptions only as historical evidence. Any such file selected
for future editing must be opened and its contracts/consumers revalidated first.

| Path relative to target | Kind | Accounting disposition / depth | Evidence or next gate |
| --- | --- | --- | --- |
| `api.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `application.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `application_composition.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `application_composition_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `binding_store.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `collaboration_producer.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `configuration.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `configuration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `csv_decoder.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `csv_decoder_test.go` | authored test | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `csv_parser.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `diagnostic_validation.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `digest.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `entropy.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `extension_state.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `extension_state_v3_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `graph_capacity_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_cleanup_deadline_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_materialization_payload.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_materialization_payload_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_materialization_timeout_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_projection_adapter.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_query_application.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_query_application_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_query_presentation.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_response_v2.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_restore_source.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_result_cleanup.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_result_cleanup_dispatcher.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_result_cleanup_dispatcher_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_result_cleanup_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_result_cleanup_race_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_result_cleanup_test_bridge_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_routes.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_source_composer.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `graph_streaming_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_telemetry.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_telemetry_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_temporal.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_temporal_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_view_authority_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_view_authority_test_bridge_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_view_jobs.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_view_receipts.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_view_receipts_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_view_routes.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_view_store.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `graph_view_store_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harness_audit_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harness_authorization_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harness_randomness_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harness_worker_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/control_json.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/control_json_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/controls.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/fault_consumers.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/network_flow_audit_assertion.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/network_flow_audit_assertion_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/network_flow_auth_transition.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/network_flow_auth_transition_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/network_flow_fault.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/network_flow_fault_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/network_flow_randomness.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/network_flow_randomness_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/randomness_consumers.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `harnesscontrol/test_helpers_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `import_facade.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `import_owner_errors.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `import_preview_test.go` | authored test | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `import_preview_validation.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `import_source_inspection.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `import_source_inspection_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `import_validation.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `import_validation_test.go` | authored test | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `indicator_link.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `indicator_link_admission.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `indicator_link_boundary_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `indicator_link_fixture_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `indicator_link_graph_admission.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `indicator_link_http.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `indicator_link_receipts.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `integration_surface_test.go` | authored test | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `keyring.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `keyring_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `mapping.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `module.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `names.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `network_flow_behavior_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `network_flow_contract_test.go` | authored test | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `network_flow_unit_test.go` | authored test | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `operation_admission.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `operation_admission_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `portability_state.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `ports.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `production_surface_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `query.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `query_authoring_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `query_filter.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `query_sql.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `query_sql_test_bridge_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `recovery_bindings.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `recovery_state.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `reporting_graph_source.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `reporting_graph_source_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `request_decoding.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `resources.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `routes.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `routes_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `saved_graph_admission.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `saved_graph_admission_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `saved_graph_application.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `saved_graph_read_application.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `saved_graph_read_application_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `saved_graph_receipts.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `security.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `semantic_boundary_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `semantic_failure.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `semantic_http_adapter.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `semantic_http_errors.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `semantic_reference_boundary_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `store.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `store_fixture_test.go` | authored test | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `store_test.go` | authored test | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `table_http.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `table_lifecycle_integration_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `table_lifecycle_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `table_query_application.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `table_query_presentation.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `table_receipts.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `textcanon_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `timestamp.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `timestamp_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `timezone.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `transaction.go` | authored production/support | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `transaction_participants.go` | authored production/support | directly inspected; relevant bodies/assertions | E20–E22 as applicable |
| `transaction_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |
| `v3_contract_projection_test.go` | authored test | metadata-only; enumerated, no current behavior claim | E19; reinspection required before movement |

External inspection covered the adopted Network Flow and Core 00 clauses,
domain owner navigation, framework, applicable backend/tooling instructions,
test-operations procedure, authored Network Flow/Imports verification rows,
generated-artifact policy, Make command paths, fixture-schema validation and
selected expected/scenario/mapping artifacts. The eight bundles' entire source
and consumer closure is **not yet established**: S32 must inspect every input,
transcript, expected artifact and direct consumer before S33. Bundles outside
the selected eight remain unqualified by DOC-03.

### 26.3 Boundary diagnosis

| Responsibility | Current location / observation | Owner candidate | Decision | Finding / contract / evidence |
| --- | --- | --- | --- | --- |
| CSV decoding and per-record normalization | Private decoder and `validateCSVRecord`; preview stops at its boundary, apply scans all records | Network Flow | Keep the shared decoder and per-record semantics | F36; C03/C21/C22; E20 |
| Preview versus apply accounting | `validateRows` carries preview and materialized-apply policy; production apply uses `validateImportPass` | Network Flow | Split accounting by actual consumer; remove obsolete production apply branch | F36; C03/C22; E20 |
| Import-row construction | Normalizer, insertion visitor and persisted reads all use `flowRow`; `materializeRow` accepts optional caller-supplied identity/provenance | Network Flow private writer | Split normalized input from persisted record; derive identity once | F37; C03/C05/C06/C09; E21 |
| Fixture execution | Frozen-corpus test checks integrity but also calls an error-discarding pseudo-behavior traversal | Network Flow tests; Harness owns routing only | Preserve integrity checks; replace selected behavioral coverage through real adapters | F38; C25; E22 |
| Public facade and cross-owner coordination | Module composition and borrowed Imports transaction remain live | Network Flow / Imports / application assembly | Keep; no new package or public port | C01/C03/C09/C24; E20/E21 |
| Other graph, query, linking and harness internals | Current implementations remain consumers of persisted rows | Existing owner boundaries | Defer unrelated cleanup; current schema suffixes and rejection guards are not dead-code evidence | C05–C17; E19 |

### 26.4 Authority and contract decisions

Core 00 §1.3 explicitly adopts Network Flow 8.0.0/major 8 and Extensions 0.13.0;
§1.4 adopts the target-neutral source-admission additions. Network Flow's
current header and §1 retain durable state/minimum state 5. Current owner
adoption is established; earlier proposal language in historical tracker
entries is not a new approval gate. No owner amendment is proposed here.

| Contract | Observed implementation and current owner requirement | Disposition / prerequisites | Slice / validation |
| --- | --- | --- | --- |
| C03/C21/C22 — Preview, apply and diagnostic semantics | NF §§9.2,10,12.4; NF-REQ-062a/b; NF-AC-081,083,084,088,089,091,107. Production apply is two-pass streaming; preview's complete bounded diagnostics differ intentionally from apply retention | Preserve first 50 complete records, raw-only tail integrity, complete apply validation, exact diagnostic ordering/privacy, counts and source/mapping agreement. Do not substitute generic Workbook CSV tail semantics | S32/S34; V29/V30 |
| C03/C05/C06 — Row identity and downstream consumption | NF-REQ-039–041; §9/Table 9-D; NF-AC-085,089,090. Normalization produces source digests and semantic values, while the writer currently conditionally preserves supplied row IDs/provenance | Preserve canonical bytes and public nullable fields; construct ownership, mapping fingerprint, identity and provenance from validated import context. No HTTP, schema, storage-column or read-port change | S32/S35; V31/V32 |
| C09/C24 — Commit and authorization | NF-REQ-062a/b and §10.2 retain Imports' final commit/cancellation/recovery authority; Network Flow receives a borrowed transaction | Preserve no table before successful validation, all-or-nothing owner effects, audit, source rechecks and replay; no staging system or added database privilege | S32/S35/S36; V31/V32 |
| C25 — Fixture fidelity | NF-REQ-177–178 require independently executable bundles and comparison of produced behavior with committed expected artifacts/state. Inspected older mappings/expected shapes and discarded failures do not satisfy that obligation | Separate owner-required test/fixture correction in S33, requiring later implementation authorization. Advance only the eight selected freeze revisions to 3; no runtime compatibility translation | S32/S33; V29 |
| C01/C02/C04/C07–C17 — Adjacent surfaces | Module facade, HTTP/authorization, table lifecycle, graph results, indicator links, Collaboration intents, Recovery and generated interfaces retain their adopted owners | Preserve; no new routes, WebSocket events, Core saved views, frontend selectors, public Go types or migration. Broaden inspection only if a planned change actually touches one | S35/S36; V31/V32 |

No unresolved contradiction between primary owners was established in these
inspected clauses. Fixture/projection drift is resolved through the adopted
owner, not by treating existing expected output as authority. A newly exposed
runtime mismatch becomes a separately described correction with compatibility
impact; it is not silently fixed during a structural slice.

### 26.5 Findings

All three findings are **TODO** and `must_fix` only for their named exits.

| Finding | Evidence / affected contracts | Required action and benefit | Risk / scoped acceptance | Slice / validation |
| --- | --- | --- | --- | --- |
| F36 — Materialized-apply policy retained by tests | E20; C03/C21/C22. `validateRows` branches on `parsedCSV.Mode`; its only production caller is preview preparation. Test-only `parseCSVApply` sustains apply retention/sample fields | Make preview accounting explicit; route apply assertions to streaming preparation/validation. Remove accepted-row collection from preview accounting, obsolete mode-dependent policy, apply-only fields and the unused error return | Duplicated policy can pass tests while the serving path diverges. Exit requires complete preview diagnostics and exact streaming retention/sample parity without that production branch | S32/S34; V30 |
| F37 — Persisted identity is accepted through normalized-row input | E21; C03/C05/C06/C09. `flowRow` carries both normalized values and stored identity; `materializeRow` preserves nonempty RowID, MappingFingerprint and ObservationSourceRef; store creation defaults an omitted parser profile | Introduce a private normalized-row input; derive persisted ownership/identity/provenance unconditionally. Require the validated parser profile; keep test conveniences outside production | Fixture overrides obscure the invariant and allow future callers to carry inconsistent identity. This is a private-boundary hardening finding, not proof that today's public import accepts forged IDs | S32/S35; V31 |
| F38 — Fixture integrity is presented as runtime behavior | E22; C25. `AssertAllFixtureRuntimeBehavior` drops parse/decode/validation failures and compares only fixture identity/revision; inspected expected artifacts contain obsolete IDs/tokens and mappings | Replace selected fixture execution with real adapter/state assertions and independently reviewed current expectations; preserve integrity checks and Git history | A passing traversal does not establish NF-REQ-177–178. Exit requires actual comparisons, wrong-expectation rejection and explicit accounting of the remaining corpus | S32/S33; V29 |

Retain live versioned schemas and algorithm IDs, state rejection guards,
historical migration-ledger inputs and owner-local harness controls. Do not
manufacture deletion candidates from old-looking names. Broader fixture
remediation and full production/release qualification remain deferred scope,
not completed findings.

### 26.6 Workstreams and dependencies

This table is the authoritative dependency list. All workflows are sequential;
there are no optional peers or hidden prerequisites. Each checkpoint includes
the exact validation result and tracker update before advancing.

| Workflow | Class | Prerequisite | Goal / likely subsystem | Validation | Binary exit / handoff |
| --- | --- | --- | --- | --- | --- |
| WF-14 / S32 | root | Later implementation authorization and baseline revalidation | Characterize current imports and complete eight-bundle discovery | V29/V30/V31 baseline | Selected inputs/callers inspected, owner/expected-output matrix reviewed, actual baseline outcomes retained |
| WF-15 / S33 | chain | WF-14 | Replace selected legacy fixture execution and advance freeze revisions | V29 | Eight bundles execute current adapters with real expected-output/state comparisons; integrity and negative control pass |
| WF-16 / S34 | chain | WF-15 | Retire materialized-apply policy from production | V30 | Preview and streaming apply use their intended accounting paths; no obsolete production branch remains |
| WF-17 / S35 | chain | WF-16 | Separate normalized row input from persisted identity/provenance | V31 | Canonical persisted output is unchanged and test overrides cannot enter the production write input |
| WF-18 / S36 | chain | WF-17 | Finalize, freeze and qualify the affected portfolio | V32/V33 | Required focused checks and cleanup pass at compatible source; exact evidence and remaining debt recorded |

### 26.7 Future implementation slices

All slice work statuses remain **TODO**. Readiness does not authorize execution.

| Slice / kind | Intended change | Entry / readiness | Tests and exit | Rollback / migration |
| --- | --- | --- | --- | --- |
| S32 — Characterization and fixture scope | Inspect all eight bundles and their consumers; characterize preview/apply, identities, source protection, borrowed transactions and replay; register narrowly owned assertions | **ready** after baseline revalidation in a later authorized task; E19–E23 establish the bounded plan | V29–V31 baseline records distinguish observed defects from valid preserved behavior; identify every required expected artifact and current consumer | Test/fixture preparation only; no runtime or database change |
| S33 — Separate owner-required fixture correction | Advance selected bundles 001–005,020,021,023 from freeze revision 2 to 3; use current mapping/response contracts, deterministic inputs and reviewed expectations; include required graph output for 001–002 | **blocked**, RB-006/RB-007 until S32 characterization and bundle inspection complete; requires later authorization | V29: actual adapter/state comparisons, unexpected-error failures, wrong-expectation negative control, exact artifact integrity, mixed revisions; remove error-discarding traversal only after selected replacement coverage exists | Prior artifacts remain in Git; no legacy converter or blanket snapshot approval; no owner/runtime schema version change |
| S34 — Structural preservation | Replace `validateRows` with preview-only counts/complete diagnostics; remove accepted-row retention, apply-only fields/policy, preview container's obsolete mode and unused error return; preserve decoder modes | **blocked**, RB-006; S33 and current compatible V29 baseline required | V30: move apply assertions to `prepareImportApply`/`validateImportPass`; decoder-only collectors may remain in test files without duplicating apply policy; preserve required wire `diagnostics_truncated=false` | Source/test rollback only; unchanged parser profile and public output |
| S35 — Structural preservation and private-boundary hardening | Private normalized-row stream excludes RowID, incident/table IDs, audit attribution, MappingFingerprint and ObservationSourceRef; materialize these once from validated context; remove parser default and provenance overrides | **blocked**, RB-006; S34 and compatible identity/transaction characterization required | V31: unchanged exact persisted values, source ordering and downstream reads; malformed private input fails before writes; construction errors propagate rather than producing empty provenance | No data rewrite, migration or public Go interface change; update test fixtures to canonical construction |
| S36 — Acceptance / operational handoff | Run finalization, review outputs, freeze source, execute affected portfolio and document compatibility/cleanup/remaining debt | **blocked**, RB-006 until S35 checkpoint | V32/V33: exact selection/source/outer exits/artifacts recorded, no missing required selected obligation, no unexpected source changes | No deployment/reset/commit; no complete-owner or release claim |

S33 must preserve the fixture manifest schema and integrity policy. Replace the
test's global `freeze.revision == 2` assumption with valid positive revisions
and manifest-to-artifact revision agreement, allowing selected revision 3 and
untouched revision 2 bundles. Update the selected manifests' authored file sizes
and hashes with the revised artifacts; never add acceptance IDs, test selectors,
specification paths or Markdown-derived hashes to executable inputs.

Expected outputs must be reviewed against adopted behavior and independent
identity/digest vectors, not regenerated from the implementation and accepted
without review. Pin fixture actor/incident identities, time and entropy where
needed for exact comparison. Unexpected decoding or execution errors fail the
behavioral test. Other bundles are not considered behaviorally qualified merely
because their integrity checks still pass.

S35 retains existing nullable field meanings and source-derived digest bytes.
The normalized type remains private and contains semantic values, source row
number and source/normalized digests. The persisted read model remains private
`flowRow`; graph/query consumers continue receiving stored rows. No generic
row framework, broad package move or extra public facade is needed.

### 26.8 Validation obligations and command routing

Commands below are future obligations except V33. Run from the repository root,
using a fresh explicit `CARTULARY_TEST_RUN_ID` for every execution. Keep exact
commands, selected rows, source/environment identity, native exit and cleanup
artifacts together. The current row names were read from authored manifests and
the public owner task guides; revalidate registration after test changes.
Tests and evidence must never read, stat or hash Markdown.

| Validation | Finding / contract / stage | Exact command or selection | Required environment / expected outcome / failure gate |
| --- | --- | --- | --- |
| V29-A / fixture integrity | F38/C25; pre-move and post-S33 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847` | Go harness; integrity covers current mixed revisions. A pass alone is not behavioral evidence |
| V29-B / selected bundle behavior | F38/C25; S32 preparation, S33 and final | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.current_import_fixture_behavior` | **Planned new row, not yet registered.** S32 authors its routing and test selector; eight selected bundles run through current adapters with PostgreSQL/claimed-profile composition as required. Wrong expectation must fail comparison. Missing row or incomplete comparison blocks S33 |
| V30-A / preview and decoder | F36/C03/C21; pre-move and post-S34 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c,module.networkflow.unit.source_inspection_header_only` | Go harness; first 50 including blank/mismatch rows, post-boundary malformed bytes ignored semantically, raw-tail integrity failure rejects, full preview diagnostics and privacy |
| V30-B / streaming accounting | F36/C03/C22; pre-move and post-S34 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd` | Go harness; zero/exact/overflow retention, rejected-row multiplicity, all-rejected sample independence, bounded two-pass processing and late failures; add migrated apply assertions here |
| V31-A / identity and private construction | F37/C03/C05/C06; pre-move and post-S35 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a` | Go harness; extend existing identity tests for canonical persisted construction, nullable fields, duplicate-row locator identity and invalid private context; no fixture-supplied ID/provenance override |
| V31-B / persistence and state admission | F37/C03/C04/C09; pre-move and post-S35 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.integration.pagination_recovery,module.networkflow.integration.bounded_graph_contributor_pipeline` | Managed PostgreSQL/claimed server where selected; store lifecycle, state-5 admission/old-state rejection, accepted-row paging and graph/link consumers retain behavior |
| V31-C / atomic public apply | F36/F37/C03/C09; pre-move and post-S35 | `make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.network_flow_default_limit_100001_rows` | Managed PostgreSQL/object store and claimed runtime as selected; exact table/row/diagnostic/audit effects, safe errors and large streaming import; no partial publication |
| V31-D / cancellation and replay | F37/C09; pre-move and post-S35 | `make test-slice OWNER=module.imports ROWS=module.imports.integration.current_workflows,module.imports.integration.current_cancellation,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6` | Managed services; late read/digest/insertion failure and cancellation roll back; postcommit crash/replay publishes one success without repeated owner writes |
| V32-A / browser handoff | F36/F37/C03; final regression | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.import_recovery` | Installed Chromium and claimed runtime; mapping/approval/apply and published table recovery remain coherent |
| V32-B / source preparation | F36–F38; before final source freeze | `make agent-finalize` | Later implementation only; inspect generated changes before freezing. Supply retained warm-check roots only if compatible; do not launch a redundant check to obtain one |
| V32-C / final source checks | F36–F38; final regression | `make source-check CHECK=generated`; `make source-check CHECK=schema`; `make source-check CHECK=catalog`; `make lint CHECK=go`; `make lint CHECK=backend-boundary` | Pinned installed toolchain/readiness; projections/routing and boundaries pass. Generated outputs refreshed through Make, never manually edited |
| V32-D / frozen portfolio | F36–F38; final regression | Exact V29–V31 selections above plus V32-A/C, after V32-B | Combine selections by owner where useful; require compatible final-source evidence without automatically duplicating satisfied obligations. Source repairs require reevaluating affected acceptance |
| V33 / documentation | F36–F38; planning-session check and future checkpoints | `make lint-markdown` | Authored Markdown only; expected exit 0 and exact `adhoc/lint-markdown/tool-run-summary.json`; failure blocks document closure |

V29-B is an explicitly planned registration, not a command claimed available or
executed today. Author its selector and owner row in S32; if repository row-ID
validation requires the canonical authoring utility, derive the admitted ID
through `make author-test-row-id` and record that replacement before execution.
This procedural registration gate does not permit reusing an unrelated row or
claiming the old integrity row compares behavior.

Focused acceptance covers the selected fixtures and affected import/persistence
surfaces. Full CI, complete-owner audits, visual-golden refresh and independent
performance/release qualification are outside this portfolio. The browser row
is functional validation; it does not authorize visual baseline changes.
No frontend type/protocol check is added solely for a private Go type change;
expand only if actual ownership or interface changes justify it.

| Invocation-path / session item | Admission and side effects | Execution / applicability |
| --- | --- | --- |
| `make help ALL=1`; `make task-guide OWNER=module.networkflow`; `make task-guide OWNER=module.imports` | Inspected Make/read-only Node diagnostic path; existing workspace lock; no service or generation | Passed, exits 0 in the planning exchange at the same source; command discovery only, E23 |
| V33 Make path | Generated Make recipe → readiness/prerequisite checks → run-step → `markdownlint.sh` → runner; `fix=false`; tracker glob is selected. Ordinary result/cache writes are explicitly admitted by the approved document plan | PASS, owning exit 0, run `nf-doc03-markdown-20261030-01`; installed readiness avoided dependency installation; E24 |
| V29–V32 | Future product/setup/generation/service effects exceed this invocation | Not run; no product pass inferred from inspected tests or historical §25 evidence |
| Format, generation, agent-finalize and product verification now | Documentation-only task excludes these repository workflows | Skipped as required by root procedure; not missing document checks |

### 26.9 Current work tracker

| Work ID | Item | Workflow | Status | Dependency / traceability | Binary exit |
| --- | --- | --- | --- | --- | --- |
| T14 | DOC-03 tracker update | Planning only | DONE | E19–E24; F36–F38 | V33 passed; preservation comparison passed; current/historical claims separated |
| T15 | Characterization and eight-bundle discovery | WF-14 | TODO | S32; C03/C25; V29–V31 | Scoped baseline and review matrix retained |
| T16 | Current fixture execution | WF-15 | TODO | T15; S33/F38/C25/V29 | Eight revision-3 bundles compare actual behavior; remaining corpus limits explicit |
| T17 | Preview/apply accounting cleanup | WF-16 | TODO | T16; S34/F36/V30 | Obsolete materialized-apply production policy removed |
| T18 | Canonical import-row write boundary | WF-17 | TODO | T17; S35/F37/V31 | Normalized input cannot override persisted identity/provenance |
| T19 | Frozen focused acceptance and handoff | WF-18 | TODO | T18; S36/V32/V33 | Compatible selected evidence, cleanup and remaining production debt recorded |

### 26.10 Session and evidence handoff

#### Scope and authority

| Time / session | Baseline and current state | Files / commands / evidence | Supersession / next action |
| --- | --- | --- | --- |
| 2026-10-30 / DOC-03 | Maintainer selected cleanup/hardening, focused acceptance and affected import fixtures, then authorized this document update only | Root/backend/tooling instructions, refactor framework, skills, Core 00 and NF owner clauses; E19/E23 | DOC-03 controls future planning; owner adoption already established; S32 requires a later implementation task |

#### Backend boundary

| Time / session | Current state | Inspected or touched / evidence | Next action |
| --- | --- | --- | --- |
| 2026-10-30 / DOC-03 | Streaming apply is live; old materialized policy remains for tests; write input mixes normalized and stored facts | §26.2 direct production rows; E20/E21; tracker is sole authored edit | S32 characterization, then S34/S35; no broad module split |

#### Frontend boundary

| Time / session | Current state | Inspection limit / evidence | Next action |
| --- | --- | --- | --- |
| 2026-10-30 / DOC-03 | No frontend change proposed; public table/import contracts preserved | Browser row routing inspected; no rendered UI inspection or current browser execution, E23 | Preserve import-recovery workflow through V32-A after implementation |

#### Contract and codegen

| Time / session | Current state | Inspected or touched / evidence | Next action |
| --- | --- | --- | --- |
| 2026-10-30 / DOC-03 | Major 8/state 5 and current schemas remain; selected fixture freeze revisions will change, not owner versions | C03/C25 owner clauses, fixture samples and schema-validator body; E22 | S32/S33 revise authored fixture/routing inputs; Make owns generated projections; no contract/codegen edit now |

#### Tests and harness

| Time / session | Current state | Commands / artifacts / evidence | Next action |
| --- | --- | --- | --- |
| 2026-10-30 / DOC-03 | Current tests inspected, not executed; fixture integrity is distinct from runtime comparison | E20/E22; public help and two task guides passed, E23; V33 passed with owning exit 0, E24 | Register planned V29-B before any execution; retain exact source/selection/native exits and cleanup |

#### Security and authorization

| Time / session | Current state | Evidence / limits | Next action |
| --- | --- | --- | --- |
| 2026-10-30 / DOC-03 | Imports keeps commit/cancellation/recovery ownership; row/diagnostic privacy and audit unchanged | NF-REQ-062a/b, §10.2; E20/E21; no security test execution claimed | Characterize complete rollback and replay; do not introduce provenance overrides or raw-value leakage |

#### Open risks and next session

| Time / session | Current state | Commands / artifacts / evidence | Next action |
| --- | --- | --- | --- |
| 2026-10-30 / DOC-03 | Missing executed characterization and incomplete selected-bundle consumer inspection gate future movement | RB-006/RB-007; baseline retained at `/tmp/cartulary-nf-doc03-6fj3ffog`; E24 confirms tracker-only authored changes and unchanged index | Revalidate affected source and instructions before S32; retain broader fixture fidelity as separate debt |

| Evidence ID | Source / inspected surface | Baseline / observation | Applicability |
| --- | --- | --- | --- |
| E19 | Git status/HEAD, target enumeration, path/link checks, SHA-256/content/file-identity snapshot; §26.1 | Clean `98adcc82961456bcb74bbecd0023136485afbf00`; 138 module entries; initial tracker fingerprint retained | Current document baseline; not a test result |
| E20 | `csv_parser.go`, `csv_decoder.go`, `import_facade.go`, `import_validation.go`, relevant preview/decoder/streaming/contract tests | Production preview alone calls `validateRows`; streamed apply has its own summaries/visitor; test helper sustains materialized policy | Supports F36 and V30; no executed characterization |
| E21 | `store.go`, `transaction_participants.go`, normalizer, module/composition/ports, fixture bridge/store tests, graph consumers and extension-state admission | Conditional persisted fields and parser fallback inspected; production stream and fixture override callers traced | Supports F37; private API correction, no demonstrated public exploit or changed owner contract |
| E22 | Frozen-corpus/traversal tests; expected artifacts for 001–005/020/021/023, selected scenario/mapping files; NF §22 and shape-validator body | Traversal drops errors and checks identity/revision; inspected legacy fixtures contain obsolete shapes; manifest validator permits positive freeze revisions | Supports F38; complete bundle/input/consumer inspection still belongs to S32 |
| E23 | Make/readiness/command paths, generated policy, authored Network Flow/Imports test families, test-operations skill/recipes, owner clauses and framework | `make help ALL=1` and both owner task guides exited 0; exact existing rows resolved | Selection evidence only; V29-B explicitly future registration |
| E24 | `CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-doc03-markdown-20261030-01 make lint-markdown`; content/file-identity/index comparison against E19 | PASS, owning exit 0; `.cartulary/test-results/nf-doc03-markdown-20261030-01/adhoc/lint-markdown/tool-run-summary.json`; only tracker changed, no added/removed paths, index unchanged | Document validation only. Pre-checkpoint proof retained outside repository; final record confirmation and final preservation recheck described below |

The initial `git diff --check` found one extra blank line at this document's
EOF. The final record removes it and updates only DOC-03 completion/evidence
fields. Final record confirmation uses
`CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-doc03-markdown-20261030-02 make lint-markdown`;
its terminal result is retained at
`.cartulary/test-results/nf-doc03-markdown-20261030-02/adhoc/lint-markdown/tool-run-summary.json`.
The final preservation recheck also verifies that removing DOC-03 additions and
restoring the IMP-07 heading reproduces the original tracker bytes exactly.
No production refactor, fixture rewrite, test registration or owner amendment
was performed by DOC-03.

### 26.11 Gates and follow-up debt

| ID | Gate / affected scope | Needed evidence | Status | Resolution condition |
| --- | --- | --- | --- | --- |
| RB-006 | Missing current executed characterization; S33–S36 | Compatible V29–V31 baseline and each predecessor's passing checkpoint | BLOCKED | S32 establishes baseline; later slices establish their required exits. Historical passes cannot substitute without compatibility proof |
| RB-007 | Selected-bundle input/consumer and expected-output matrix incomplete; S33 | Open all eight bundles and relevant production-facing adapters; register V29-B; review independent expected outputs and mixed-revision integrity | BLOCKED | S32 completes concrete fixture matrix and registration before S33 |
| T20 | Broader frozen fixture fidelity / full production-readiness qualification | Separate inspection of remaining 20 fixture families and applicable acceptance/publication obligations | DEFERRED | Later scoped task; current integrity passes do not qualify their behavior |

No owner contradiction or new owner-adoption gate has been found for this plan.
The eight selected fixtures' correction implements existing NF-REQ-177–178.
If comparison exposes an additional runtime defect, name its owner clause and
compatibility consequence before a separately authorized correction; do not
weaken expectations or silently expand S34/S35.

### 26.12 Completion and acceptance limits

| Outcome | DOC-03 statement |
| --- | --- |
| Planning completeness | **Complete for the bounded next-iteration document plan.** Whole-module behavioral audit and complete selected-bundle inspection are not claimed; the latter is the explicit S32 prerequisite |
| Per-slice readiness | S32 **ready** for later authorized characterization after baseline revalidation. S33 **blocked** by RB-006/RB-007; S34–S36 **blocked** by RB-006 and their predecessor checkpoints |
| Implementation authorization | **Not granted by this planning invocation**; all S32–S36 and T15–T19 remain TODO |
| Session write boundary | **Verified**, E24: only this tracker changed among recorded authored paths; index unchanged, no added/removed paths; ordinary admitted lint cache/result outputs excluded from that comparison |
| Document validation | **Passed**, V33/E24; no product checks executed; final record confirmation has its own exact run root |
| Production readiness | This is a plan for measurable progress, not a production, complete-owner, performance, release or deployment qualification |

| Document acceptance criterion | Current result / evidence |
| --- | --- |
| Canonical target/output, baseline and inspection depth recorded | Met; E19 and §§26.1–26.2 |
| Adopted behavior separated from legacy implementation and fixture drift | Met; C03/C25, F36–F38; no feature retirement or owner amendment |
| Findings trace to acyclic slices and validation with explicit gates | Met; WF-14–WF-18, S32–S36, V29–V33, RB-006/RB-007 |
| Future statuses and historical evidence are not promoted to completion | Met; §§1,26.9,26.12; IMP-07 history retained |
| Exact current row selections and planned registration distinguished | Met; E23/V29-B; no invented passing command |
| Required Markdown check passes | Met; V33/E24, owning exit 0 |
| Final content/index comparison preserves all other authored files | Met; E24, with exact retained baseline and comparison artifacts |

## 27. IMP-08 — Complete fixture and import remediation

### 27.1 Authority, baseline and execution contract

The maintainer's implementation request authorizes the complete agreed plan.
The target is `internal/modules/networkflow`, with its fixture corpus and the
necessary Imports, test-support, schema and routing consumers. Core 00's adopted
Network Flow 8.0.0/major 8/state 5 decisions remain authoritative. NF-REQ-177–178
and Testing Harness §8.2 govern behavioral fixtures and their execution; tests
and machine projections never consume Markdown. F39 corrects an obsolete
fixture reference to the already-adopted §14.4 v2 input, without restoring v1.

Entry baseline: `main`, `98adcc82961456bcb74bbecd0023136485afbf00`; the tracker
is the sole pre-existing modified path, with no staged or non-ignored untracked
files. Its entry SHA-256 is
`79ff7b6bc400bff1024a7ec0167cc825a899257fbc591e0315df12ac623a98c4`.
Content, file identities, index digest and original tracker are retained under
`/tmp/cartulary-nf-imp08-vm452p0o`. Existing DOC-03 edits are preserved.

Public interfaces, storage columns, row/digest algorithms and major/state
versions have no planned change. Imports retains final commit, cancellation and
replay authority. No legacy converter, generic behavior interpreter, broad
package split or production determinism hook is introduced. Necessary discovered
owner violations receive separate findings and correction workstreams. Such a
workstream depends on its failing reproduction, not completion of the blocked
fixture group; the group's exit depends on the correction's passing evidence.

### 27.2 Findings and remediation decisions

| Finding | Areas / remediation | Rationale and long-term benefit | Compatibility / unresolved risk | Completion evidence |
| --- | --- | --- | --- | --- |
| F36 | Implementation/tests/docs: preview-only accounting; streaming apply assertions; remove materialized apply policy | Serving paths own their accounting and share record normalization | Private Go change; duplicate policy otherwise masks serving-path defects | S34; V30 and revised fixture regressions |
| F37 | Implementation/tests/docs: private normalized-row input and unconditional persisted identity/provenance construction; require parser profile; propagate errors | Types express write ownership and prevent override coupling | No planned migration or digest change; inconsistent future provenance otherwise remains possible | S35; V31 independent identity, state, transaction and consumer evidence |
| F38 | Fixtures/tests/routing/docs and required runtime corrections: real comparisons for every bundle | Trustworthy independent regression boundaries | Revision 2 advances to 3 on first correction; increment again for later frozen-byte changes; no runtime legacy reader | S33-A–F; expanded V29, all 28 executable |
| F39 | Specification/fixtures/tests/docs: Table 22-A fixture 015 points to §14.4 v2 | One authoritative adapter definition | Editorial owner alignment; obsolete reference otherwise encourages v1 restoration | S37 source review; S33-D adapter comparison; V34 |
| F40 | Test support/harness/schema/tests/docs: strict admission, run-local read-only copies and bounded retained comparison evidence | Identifiable immutable inputs and reviewable outcomes | Manifest/scenario v2 retained; unsafe inputs and incomplete evidence otherwise persist | S38/S33-F; V35 admission/materialization/schema and all fixture summaries |
| F41 | Implementation/tests, durable compatibility and docs: replace permissive canonical serialization with the exact NF-REQ-031/032a/032b encoding, including closed-object order and invalid-number rejection | Fixture 019 independently demonstrates four mismatches; one serialization authority must serve digests and normalized values | Correcting bytes changes mapping fingerprints and dependent persisted row identities; compatibility/migration must be explicit before runtime changes | S39; fixture 019 and independent mapping/row/graph/idempotency vectors, migration/replay/state admission |
| F45 | Specification/tooling/tests/docs: distinguish procedural absence guards from forbidden permissive DDL | Prevents migration policy from rejecting consistency checks | No product compatibility change; false rejection otherwise blocks valid migrations | S43 exact policy negatives/positive guard, generation and Markdown |
| F46 | Tooling/projections/tests: migration catalog collects object kinds out of source order and resurrects indexes after their table is dropped | REQ-01-661 exact physical inventory; PostgreSQL parity fails on migration 34's temporary-in-transaction indexes | No runtime compatibility change; generated permissions/recovery inventory must not contain absent objects | S44 source-ordered projection, generated inventory and actual PostgreSQL parity |
| F47 | Specification/docs: Core §2.1A still says rewrite existing files and head 29, conflicting with REQ-01-661's adopted append-only increments | Align lineage reset rules with current-line owner-authorized increments and NF-REQ-032c | Clarifies existing migration obligations; no new lineage or production-baseline claim | S45 owner consistency and Markdown |

| F44 | Implementation/tests/docs: typed numeric mapping tuple ordering under NF-042 | Independently distinguishes ignored ordinals 2 and 10; prevents runtime string representations defining identity | Corrected fingerprints join S39 migration; leaving it open creates inconsistent identities | S42 passing 019/preview/identity vectors; S39 migration |
| F42 | Implementation/tests/docs: Network Flow must use Unicode-preserving strict JSON admission before semantic validation | NF-REQ-020 step 2 and Table 21-C require rejection before unknown-member validation; fixture 010 proves an unpaired surrogate is silently replaced | Malformed UTF-8/surrogate requests reject earlier; valid request and durable formats remain unchanged. No migration | S40; fixture 010, mapping/preview boundary regressions and affected browser query workflow |
| F43 | Harness implementation/tests/docs: retained scanning rejects owner-only read-only fixture inputs and changes their permissions | TH-HARNESS-REQ-266 requires immutable run-local input; REQ-603 requires private creation before sealing. S40 focused run passes comparisons but exits with artifact error at final scan | Accept sealed owner-only 0400 files/0500 directories in retained scanning while preserving all privacy/content checks. No runtime or durable data change | S41; native session scan negatives and complete fixture Make exit |

### 27.3 Workstreams and authoritative dependencies

All work statuses below are execution statuses, independent of historical
planning completion. S33 is an aggregate over S33-A–F. Every completed slice
records commands, native exits, source identity, artifacts, cleanup and remaining
risks here before its successor starts. Documentation checkpoints use V33.

| Slice / workstream | Depends on | Status | Binary exit |
| --- | --- | --- | --- |
| S37 Specification and tracker closure | Revalidated baseline | DONE | F39 reference consistent; F36–F40 traceable; V33 passes |
| S32 Characterization and corpus inventory | S37 | DONE | Complete 28-bundle adapter/comparison matrix and retained baseline outcomes |
| S38 Fixture admission and evidence mechanics | S32 | DONE | Strict integrity/path admission, immutable materialization, summary and schema/routing tests pass |
| S33-A Parsing, admission and normalization | S38 and S39–S51 correction exits | DONE | 003/005/010/011/012/019/021/022 execute every required case and compare |
| S43 Migration policy guard classification | Failing S39 generation reproduction | DONE | DDL negative controls, procedural guard, generation and Markdown pass |
| S42 Numeric mapping tuple ordering | Failing 019 revision 4 reproduction | DONE | Independent multi-digit ordinal fingerprint, preview and source identity checks pass |
| S44 Migration catalog event ordering | Failing S39 DDL catalog-parity reproduction | DONE | Source order preserves automatic dependent-object removal; generated inventory and PostgreSQL parity pass |
| S45 Incremental migration owner consistency | S44; Core §2.1A/REQ-01-661 contradiction | DONE | Core distinguishes reset-only lineage transition from current-line increments and agrees with NF-REQ-032c; Markdown passes |
| S39 Canonical serialization correction | Failing fixture 019 reproduction and S42–S45, not S33-A completion | DONE | Exact adopted bytes and invalid-value rejection; independently validated identity compatibility/migration; dependent fixtures pass |
| S41 Retained read-only input scan correction | Failing S40 focused native exit and scan policy inspection; independent of S33-A completion | DONE | Private sealed inputs retain permissions, secret/path negatives pass, and fixture Make native exit is zero |
| S40 Lossless JSON admission correction | Failing fixture 010 reproduction and S41; not S33-A/S39 completion | DONE | Shared Unicode validation used at Network Flow JSON entry points; exact 010 results and boundary regressions pass |
| S46 Ignored-column provenance | Failing 003 reproduction and S39 | DONE | Mapping-selected ignored values are absent; migration and exact identities pass |
| S47 Typed retained-raw canonical boundary | Failing ordinal 2/10 reproduction | DONE | Numeric ordinal order and closed raw-value shape agree across runtime/migration |
| S48 Timestamp grammar and precision | Failing 022 reproduction | DONE | Exact grammar/UTC/precision and malformed retained-time rejection pass |
| S49 Node fixture admission and evidence | S38 and S48 | DONE | Client admission, immutable inputs, isolation, negative controls and summaries pass |
| S44-R Independent schema-checker event ordering | Remaining F46 schema-check reproduction | DONE | Source-order object inventory agrees across both projections and PostgreSQL |
| S50 NFC aliases and historical approval boundary | Failing 011 backend reproduction | DONE | New approvals use NFC keys while committed historical facts remain immutable |
| S51 Explicit client alias confirmation | S50 and failing 011 client reproduction | DONE | Suggestions require explicit disposition confirmation; client/browser comparisons pass |
| S52 Complete all-rejected diagnostics | Failing 004 comparison | DONE | Complete safe resources, closed admission, streaming/preview/schema checks and replay pass |
| S33-B Import publication | S33-A and S52 | DONE | 001/002/004/020/023 compare rows/digests/mappings/diagnostics/graph/source/atomic effects |
| S53 Query admission/error fidelity | S33-C failing 014/024 reproduction | DONE | Adopted cursor/scope reasons, complete safe error details, exact fixture regressions and affected query workflow pass |
| S54 Import display-name error fidelity | S33-C failing 013 reproduction | DONE | Owner registry, preview/apply errors, migration compatibility and naming comparisons agree |
| S55 Deleted-table errors and cursor invalidation | S33-C failing 027 reproduction | DONE | Complete safe lookup context and typed cursor invalidation pass with retention and lifecycle regressions |
| S33-C Query and lifecycle | S33-B, S53, S54 and S55 | DONE | 013/014/024/027 compare actual query, name, cursor, authorization and retention behavior |
| S56 Contributor scope invalidation | Failing S33-D 009 reproduction | DONE | Initial graph staleness and continuation cursor policy pass with unchanged retained state and contributor workflows |
| S57 Contributor query-limit admission | Failing S33-D 025 reproduction | DONE | Shared omission/default and explicit-limit policy, schema projections, exact contributor comparison and affected workflows pass |
| S33-D Graph and resource bounds | S33-C, S56 and S57 | DONE | 006/008/009/015/018/025/028 compare current v2 graph, contributors, limits and stale-state behavior |
| S33-E Linking, privacy and audit | S33-D | DONE | 007/016/017/026 compare links, authorization, redaction, safe digests and exact replay/audit effects |
| S33-F Corpus closure | S33-E | DONE | All 28 qualified; misleading traversal removed; inventory equals executable coverage |
| S33 Fixture corpus aggregate | S33-A–F and required correction exits | DONE | All 28 bundles have executable behavioral evidence and corpus closure |
| S34 Preview/apply accounting | S33-F | DONE | V30 and affected fixtures pass without obsolete materialized policy |
| S35 Canonical row construction | S34 | DONE | V31 passes; normalized input cannot supply persisted identity/provenance |
| S36 Validation and handoff completion | S35 and required correction exits | DONE | Finalized source, compatible affected portfolio, cleanup and final handoff complete |

### 27.4 Validation and fixture rules

S32 re-enumerated all 28 directory-scoped manifests and opened all 135 listed
source/expected/transcript files (163 files including manifests). Each bundle is
revision 2. Source mappings use obsolete label-based fixture wrappers, scenario
files are descriptive metadata, and JSONL families contain narrative placeholders.
Expected artifacts are template objects rather than faithful current outputs.
These are fixture defects, not required compatibility behavior. Current consumers
are `network_flow_unit_test.go`'s integrity/traversal helper and the minimal CSV
consumer; existing production-facing contract, streaming, store and route tests
provide reusable composition but do not qualify these committed expectations.

The comparison matrix below is the complete required fixture scope. Production
semantics stay in Network Flow; the support layer only admits and retains files.
Every row includes positive behavior and its named rejection boundaries.

| Fixture suffix | Workstream / actual adapter boundary | Required independent comparison |
| --- | --- | --- |
| 001 cisco-sna-minimal | S33-B; import facade/borrowed transaction and graph composer | Table, approved mapping/fingerprint, exact row IDs/digests/provenance, zero diagnostics and graph result |
| 002 cisco-sna-interface-fields | S33-B; import facade/writer and graph composer | Mapped interface text/null, unsupported fields null, exact rows and graph |
| 003 duplicate-headers | S33-A; source inspection/preview/mapping | Contiguous ordinal descriptors and distinct duplicate-header selections |
| 004 rejected-rows | S33-B; streaming validation/atomic apply | IP/port/protocol/timestamp/counter/width/end-before-start diagnostics in exact order; partial/all-rejected outcomes |
| 005 csv-parser-edges | S33-A; incremental CSV decoder | Terminal newline, blank record, quoted newline, escaped quote, malformed quote and source locators |
| 006 cross-table-graph | S33-D; real source composer/Graph v2 | Two tables, shared endpoint, aggregate edge, exact query/snapshot/edge identities |
| 007 indicator-linking | S33-E; indicator-link application and participant | Existing/create/duplicate results and exact binding/observation effects |
| 008 large-limits | S33-D; configuration/import/graph boundaries | Table and graph threshold errors with valid timestamps; no publication qualification |
| 009 soft-delete-stale-graph | S33-D; lifecycle/query/graph routes | Active graph, terminal deletion, stale graph and cursor rejection |
| 010 json-admission | S33-A; production strict request decoders | Exact duplicate/null/unknown/malformed/non-object rejection, including nested/raw-byte cases |
| 011 alias-collision | S33-A; mapping/alias/preview | Alias keys and warnings, explicit ordinal approval and reused-source conflict |
| 012 sys-uptime-timestamps | S33-A; timestamp profile/record normalization | Exact derived UTC values, export/uptime/start/end inputs and wrap rejection |
| 013 filename-display | S33-C; naming and table lifecycle | Path stripping, hidden/trailing-dot stems, override rejection, suffixing and deleted-name reuse |
| 014 cursor-pagination | S33-C; database query/cursor routes | Default sort tail, exact positions/pages, terminal null, actor mismatch and rename survival |
| 015 graph-adapter-input | S33-D; source composer and actual Graph v2 adapter | Complete §14.4 v2 input, trusted context, graph-view/source identity, property definitions and safe metadata |
| 016 redaction | S33-E; diagnostics/metadata/audit boundaries | Safe null/integer samples, exact scalar hashes and safe digests; forbidden values absent from restricted outputs |
| 017 indicator-link-mismatch | S33-E; link admission/application | Non-IP field, ambiguous edge, exact-value mismatch, existing target mismatch and Core dependency failure |
| 018 resource-limits | S33-D; decoder/import/config/graph | Parser and accepted-row bounds, retention truncation, graph/counter bounds and invalid config |
| 019 canonical-json-unicode | S33-A; canonical JSON/text/digest primitives | Exact escapes/scalar ordering/NFC/Unicode whitespace, null framing and unpaired surrogate rejection |
| 020 atomic-import-commit | S33-B; Imports-owned public unit commit | Every final-step injection leaves no ghost table/row/diagnostic/binding/domain audit; successful replay occurs once |
| 021 preview-boundaries | S33-A; preview/apply decoders and source integrity | Header/control rules, exact 50 complete records including blank/mismatch, ignored semantic tail, raw integrity and limit+1 |
| 022 timestamp-rulesets | S33-A; mapping/timestamp parser | Closed variants, RFC3339 grammar, pinned timezone folds/gaps, epoch integers, uptime bounds and distinct ordinals |
| 023 import-facade-source-change | S33-B; public facade/source port | Exact requests/results/descriptors, all source-change reasons, no path leakage or partial publication |
| 024 query-normalization-cursors | S33-C; query normalization and routes | Closed scopes, duplicate filters, initial/continuation separation, 4096-byte/expiry bounds, independent row/diagnostic positions |
| 025 graph-contributors | S33-D; graph/contributor routes | Complete response, vertex/edge pages, current authorization, stale digest and exclusion of rejected rows |
| 026 audit-and-replay | S33-E; mutation receipts/audit/graph completion | Created/reused binding events, safe-digest key IDs, exact replay with zero second audit, success-only graph audit and truncated-ref count |
| 027 retention-soft-delete | S33-C; lifecycle/query/incident routes | Terminal deletion, retained source/counts across incident closure, query/cursor rejection and Core audit retention; no purge behavior |
| 028 graph-aggregate-bounds | S33-D; composer/Graph adapter | Arbitrary-precision sums of valid uint64 rows, exact digit threshold, vertex/edge/bytes/packets failure order and zero partial output |

V29–V33 retain §26.8's exact existing selections. V29-B's proposed eight-bundle
row is superseded by registered behavior-group selectors added with their tests.
Record actual row IDs before execution; proposed selectors are never evidence.
V34 covers F39's owner/reference review and current adapter comparison. V35
covers manifest/schema closure, traversal/symlink/duplicate/missing/extra-file
rejection, size/hash/bundle identity, draft blocking, materialization isolation,
read-only inputs and accurate success/failure summaries.

Each bundle runs independently using typed owner adapters and concrete inputs;
scenario summaries remain descriptive. Compare complete applicable responses
and durable effects. Compare canonical bytes, identities, digests, nullability
and ordering exactly. Pin supported test time/entropy/identities and safe-digest
keys. Review expectations independently of the tested implementation and prove
wrong expectations fail. Correct invalid large-limit timestamps and invalid
aggregate source counters so cases reach their intended boundaries. Preserve
manifest/scenario schemas and update all file/bundle hashes with each revision.

Final acceptance is the compatible union of all corpus comparisons, V30/V31,
affected graph/query/lifecycle/link/security tests, import-recovery browser row,
generated/schema/catalog source checks, Go/backend-boundary lint, relevant
fixture-support/schema tests and Markdown lint. Additional browser workflows
are selected only for relevant runtime changes. Run `make agent-finalize`, review
its output, freeze source and then qualify missing obligations. Do not repeat
compatible satisfied coverage by habit or infer complete-owner/release status.

### 27.5 Execution and evidence log

**S41 dependency insertion before correction work:**
`nf-imp08-s40-focused-20261030-01` returned native Make exit 2 (graph artifact
exit 11), although both units/five rows passed and all 11 cleanup steps completed.
`run-summary.json` says pass, but there is no successful retained-secret scan;
this is not passing invocation evidence. `ScanRetained` requires exactly
0600/0700, repairs the first sealed fixture input permission, then fails.
Fixture support already creates bytes privately at 0600 before sealing inputs
0400/0500, so REQ-603 secure creation and REQ-266 read-only execution are
compatible; the scanner's writable-only predicate is the implementation defect.
S41 will admit only the existing owner-only modes plus their sealed equivalents,
preserve sealed modes, and retain every ownership, link, content and mutation
check. Tests must cover read-only safe inputs and read-only secret/unsafe inputs.
S40 qualification and all fixture execution now depend on S41's passing exit.
No accepted expectation is changed. The import transaction regressions in
`nf-imp08-s40-imports-20261030-01` passed with native exit 0.


**S40 dependency insertion before correction work:** fixture 010 in
`nf-imp08-s33a-parsing-20261030-01` fails the unpaired-surrogate case: current
admission substitutes the invalid escape and reports `unknown_member` for
`visible_label`; NF-REQ-020 step 2 requires lossless UTF-8 JSON admission first.
Use the shared strict JSON Unicode primitive through an explicit lossless object
decoder, retaining existing callers' API and body-size ownership. Update all
Network Flow uses of that admission boundary, including mapping/preview inputs.
This is an input rejection correction, without identity, storage or migration
change. S33-A/010 depends on S40; S40 does not depend on S39 or the blocked group.
Expected fixture bytes remain unchanged.

Reproduction command used the standard explicit pinned `GO`/installed-only
environment and `CARTULARY_TEST_RUN_ID=nf-imp08-s33a-parsing-20261030-01` with
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632`.
Native exit 2; product/test-assertion failure; ten clean cleanup steps. Both CSV
and preview comparisons passed, but this aggregate does not qualify JSON. Exact
010 result: `.cartulary/test-results/nf-imp08-s33a-parsing-20261030-01/NF-FIX-010-json-admission-1830427104/produced/actual.json`;
the sibling comparison summary retains verified input identity and failed status.

S33-A subsequent authoring: fixtures **005, 010 and 021** advance revision 2 → 3.
005 has ten concrete CSV edge cases, 010 has ten actual request bodies, and 021
has nineteen preview/apply decoder cases. Full descriptors, samples, logical
records, cell hashes, raw source hashes and complete typed errors are compared.
Malformed/invalid UTF-8 inputs are stored byte-exactly. Expected records are
explicitly authored; SHA-256s are independently calculated from those values.
Each comparison uses S38 materialization and retains actual output and a summary.
The exact output comparator also rejects a deliberately incorrect expectation.
No source policy, materialized apply accounting, or persisted write input changed.

Actual additional registered rows (each derived through `make author-test-row-id`
with its exact test name as the selector key before family registration):

- `module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093`
  → `TestNetworkFlowCSVFixtureComparison`.
- `module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7`
  → `TestNetworkFlowJSONFixtureComparison`.
- `module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632`
  → `TestNetworkFlowPreviewFixtureComparison`.

These are independent portions of the already-open S33-A workstream; S39 remains
open for canonical encoding and persisted identity compatibility. Inventory
tracing confirms effects reach Imports approvals/plans/outcomes, normalized rows,
diagnostics, binding row-ref arrays/JSON, graph semantic/query hashes and receipts.
A partial serializer flip would leave mixed durable identities. Historical SQL
migrations remain unchanged and no migration has been executed.

**S39 dependency insertion before correction work:** S33-A's fixture 019 comparison
has a real failure against NF-REQ-031, NF-REQ-032a and NF-REQ-032b. S39 depends on
the retained failing reproduction below; S33-A acceptance depends on S39. This
does not invent an adoption gate or permit weakening expected bytes. Other
S33-A authoring may proceed independently, but no later group is ready.

The first 019 execution (`nf-imp08-s33a-canonical-20261030-01`, native exit 2)
exposed an adapter mistake: whitespace primitive cases were sent through table
name admission, which correctly rejects control characters before trimming.
The adapter now exercises the actual whitespace and NFC primitives; fixture
expectations and revision-3 bytes were unchanged. The second invocation used
`GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s33a-canonical-20261030-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721`.
It returned native exit 2, product/test-assertion failure, with ten successful
cleanup steps and zero failed/blocked cleanup. Source digest:
`sha256:3596714f7f4e097e9ea51c33cf37347d419898087bb2d25544e98e90d4222544`;
toolchain/system match S32. Exact evidence root:
`.cartulary/test-results/nf-imp08-s33a-canonical-20261030-02/`.
The materialization and comparison summary are at
`NF-FIX-019-canonical-json-unicode-416193815/comparison-summary.json`, with
actual output at the sibling `produced/actual.json`. Its four unequal outcomes:

- `json-01`: HTML characters and U+2028/U+2029 are escaped instead of emitted
  as literal UTF-8. Expected SHA-256
  `e2f13288a3137e2b6d907561554f2f2ebe77c2301daddbb4aa9509747e747853`.
- `json-06`: floating-point `1.5` is serialized rather than rejected.
- `digest-04`: the escaping defect changes normalized-row digest; expected
  `8c618c7ced537ae616b984bb93e3204942bc3d9a2c31083cdb05b56fac716d4c`.
- `closed-profile`: RFC3339 timestamp members are alphabetically sorted rather
  than ordered as Table 9-G; expected canonical SHA-256
  `599134f0ed17d1176ed4f8d226377c5f97bf9e8157158c0f030d9a1ef3f747bc`.

Risk/compatibility: a local escaping patch would leave ordering/type failures
and silently alter persisted digests. Trace mapping approval, accepted rows,
diagnostics, row references, binding references, graph snapshots, receipts and
state admission before selecting the authored migration. No data reset, legacy
translator, database mutation, or runtime serializer change has occurred.

S33-A authoring begins with fixture 019, advanced from frozen revision 2 to 3.
Its narrative JSONL, expected template and transcript were replaced by concrete
Unicode/escape, scalar-order, whitespace/NFC, invalid-number/surrogate, closed
timestamp-profile and null/present digest vectors. Expected bytes/hashes were
authored independently using the stated framing and schema order, not captured
from the implementation. New selector
`module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721`
was derived by `make author-test-row-id FAMILY_ID=module.networkflow.unit CLAIM='Canonical fixture compares exact bytes Unicode and digest framing' SELECTOR_KEY='internal/modules/networkflow:TestNetworkFlowCanonicalFixtureComparison'`
and added to the owner family before execution. The historical integrity check
now selects the exact revision per converted fixture; no general revision waiver
or behavioral completion claim was introduced. Remaining seven S33-A families
are unconverted and S33-A is not complete.

S38 authoring: new actual row
`module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a`
was derived with `make author-test-row-id FAMILY_ID=module.networkflow.unit CLAIM='Fixture admission and materialization preserve immutable comparison evidence' SELECTOR_KEY='internal/testutil/networkflowfixtures:TestFixtureAdmission,TestFixtureExecution'`
and registered in the authored Network Flow family input before execution.
It covers file admission/evidence mechanics, not product fixture qualification.
The first `GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s38-generate-20261030-01 make generate`
failed native exit 2 because the native schema generator does not support
`unevaluatedProperties`. The manifest schema's three file variants were flattened
to equivalent closed objects, retaining v2's accepted instance format; this avoids
adding a second schema interpreter or general evaluation tracking to the generator.
The exact existing fixture-path regex gains the same explicit predicate as the
native reference-path projections. Generation must pass before evidence runs.

| Slice / checkpoint | Changes and authority | Commands / results / retained evidence | Next gate |
| --- | --- | --- | --- |
| S37 start | Implementation authorization and baseline revalidated; tracker updated first; prior DOC-03 bytes retained | Read-only Git/source inspection; `/tmp/cartulary-nf-imp08-vm452p0o` | Correct F39 and run V33 |
| S37 completion | Table 22-A fixture 015 now references adopted §14.4 v2; no new owner behavior or compatibility alias | `CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s37-markdown-20261030-01 make lint-markdown`: PASS, native exit 0; `.cartulary/test-results/nf-imp08-s37-markdown-20261030-01/adhoc/lint-markdown/tool-run-summary.json`. Public help and both owner task guides exit 0 | S32 characterization and full corpus matrix |
| S32 baseline setup diagnosis | Installed-only selected the local Go 1.27.1 launcher despite cached 1.27.3; cryptographic admission rejected it before tests. No product defect established | `nf-imp08-s32-networkflow-20261030-01`, native exit 2, no terminal product summary. Pinned executable `/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go` reports 1.27.3; archive SHA-256 matches `tools/toolchain_pins.json` | Repeat the identical selection with explicit pinned `GO`; preserve failed root |

### 27.6 Completion and handoff

This is a historical S40/S41 checkpoint. The current completion and operational
handoff is [§27.64](#2764-s36--final-affected-portfolio-evidence-and-operational-handoff).

**S41 and S40 completion checkpoint:** S41 preserves owner-only sealed input
permissions during scanning, still rejects unsafe ownership/links/modes and
secret content, and removes unsafe bytes from sealed directories while restoring
the directory's sealed mode. No specification relaxation was required: inputs
are created at 0600 under 0700 directories before sealing, as required by
TH-HARNESS-REQ-603 and REQ-266. The new tests exercise safe immutable bytes,
registered secrets and syntactic secrets; existing attack tests remain selected.
S40 routes Network Flow JSON object admission through shared lossless Unicode
validation without changing body limits or unrelated strict-JSON callers.
Fixture 010 advances **3 → 4** for independently authored invalid-UTF-8 and valid
surrogate-pair cases; its original expected rejection remains unchanged.

Current behavioral comparisons: 005 revision 3 (10 cases), 010 revision 4
(12 cases), 021 revision 3 (19 cases), with complete expected/transcript
comparisons and wrong-expectation controls. Fixture 019 revision 3 remains a
failing F41 reproduction; 003/011/012/022 have not been rebuilt. S33-A is not
complete, and the 24 later/unconverted families have no current comparison claim.
F42's selected backend correction exit and F43's retained-execution exit are
satisfied. S36 must additionally run the affected browser row
`module.networkflow.browser.query_authoring_integrity`; the original
`module.networkflow.browser.import_recovery` obligation remains.

All commands below use
`GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only`.
Each retained root is `.cartulary/test-results/<run-id>/`; terminal evidence is
`run-summary.json`, cleanup is `cleanup-results.json`, and the successful retained
scan is `retained-secret-scan.json`. Source commit remains
`98adcc82961456bcb74bbecd0023136485afbf00` with dirty authored changes.
Toolchain identity remains
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system identity remains
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.

- `CARTULARY_TEST_RUN_ID=nf-imp08-s40-imports-20261030-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0`: PASS, native exit 0, 3 units; 14 cleanup steps completed, no failures. Source `sha256:f2a7ed922e42417fe932152e55947ce0024026a3729fe2af59c381dc286cfea9`; cache hits 0.
- `CARTULARY_TEST_RUN_ID=nf-imp08-s41-scan-20261030-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_session`: PASS, native exit 0, 1 units; 10 cleanup steps completed, no failures. Source `sha256:2e90b90e9c92af07a2260e0c27d9e63c308200d64e0af62d40aec6fbed260c96`; cache hits 0.
- `CARTULARY_TEST_RUN_ID=nf-imp08-s41-networkflow-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a`: PASS, native exit 0, 3 units; 12 cleanup steps completed, no failures. Source `sha256:2e90b90e9c92af07a2260e0c27d9e63c308200d64e0af62d40aec6fbed260c96`; cache hits 0.
  Comparison evidence: `.cartulary/test-results/nf-imp08-s41-networkflow-20261030-01/NF-FIX-005-csv-parser-edges-3207793137/comparison-summary.json`.
  Comparison evidence: `.cartulary/test-results/nf-imp08-s41-networkflow-20261030-01/NF-FIX-010-json-admission-2884310954/comparison-summary.json`.
  Comparison evidence: `.cartulary/test-results/nf-imp08-s41-networkflow-20261030-01/NF-FIX-021-preview-boundaries-3455646476/comparison-summary.json`.

Dependency documentation checks `nf-imp08-s40-dependency-markdown-20261030-01`
and `nf-imp08-s41-dependency-markdown-20261030-01` both passed native exit 0.
Both correction `make format` runs (`nf-imp08-s40-format-20261030-01`,
`nf-imp08-s41-format-20261030-01`) passed native exit 0. This checkpoint uses
`CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s41-checkpoint-20261030-01 make lint-markdown`;
it passed with native exit 0 and retained `adhoc/lint-markdown/tool-run-summary.json` before successor work. Failed outer run
`nf-imp08-s40-focused-20261030-01` remains non-qualifying despite passing unit
summaries. Diagnostic `make explain-run ... DETAIL=artifacts` was an invalid
read-only invocation (native exit 2); supported summary inspection and direct
source inspection established the scanner cause before retry.


S38 completed its support-layer exit. `internal/testutil/networkflowfixtures`
owns strict manifest/scenario validation, exact inventories and bundle hashes,
revision/draft admission, read-only isolated copies and bounded summaries with
artifact hashes. Owner callbacks run only after complete admission; failure and
blocked summaries remain distinguishable. No product services or semantics moved
into the helper. Authored schema attachments/support inventory/catalog registration
and generated projections were updated. Generator numbering changes account for
the wider validator diff; generated files were produced only by Make. F40 stays
open until S33-F confirms every behavioral fixture consumes this boundary.

S38 failures were investigated before retry: generation `...-02` rejected an
unsorted authored row list; schema check `...schema-20261030-01` rejected an
unsorted attachment list (native exit 2, product/test-assertion classification,
12 clean cleanup steps). Both were corrected in authored inputs and regenerated.
Generation `...generate-20261030-03` and `...-04` passed native exit 0; `...-04`
is current. Format `...format-20261030-01` passed native exit 0.

The following exact commands all returned native exit 0 at source digest
`sha256:2a5f244b20c822082d95a23657a19332cf307f8a6afaa9afdbc4975d9cea58d8`.
They use the same toolchain/system identities as S32. Each run retains
`run-summary.json` and `cleanup-results.json` beneath
`.cartulary/test-results/<run-id>/`; cleanup failed/blocked counts are zero.

| Run ID | Command after the shared environment below | Cleanup steps |
| --- | --- | --- |
| nf-imp08-s38-admission-20261030-02 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a` | 10 |
| nf-imp08-s38-parity-20261030-02 | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_schema` | 10 |
| nf-imp08-s38-schema-20261030-02 | `make source-check CHECK=schema` | 12 |
| nf-imp08-s38-catalog-20261030-01 | `make source-check CHECK=catalog` | 11 |
| nf-imp08-s38-generated-20261030-01 | `make source-check CHECK=generated` | 14 |
| nf-imp08-s38-boundary-20261030-01 | `make lint CHECK=backend-boundary` | See retained cleanup result |

For each command the exact environment prefix was
`GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=<run-id>`.
Support/admission and parity `...-01` passed before registry sorting and remain
historical focused evidence, superseded by `...-02`. Public task guides confirmed
the new actual selector; no proposed selector is represented as passing.
S38 documentation closure uses
`CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s38-markdown-20261030-01 make lint-markdown`.
Its native exit must pass before S33-A edits start.

S32 completed after both baseline commands below returned native exit 0. Imports
reported six work units passing and 23 successful cleanup steps, zero failed or
blocked, under `.cartulary/test-results/nf-imp08-s32-imports-20261030-01/`;
its terminal `run-summary.json` and `cleanup-results.json` retain the result.
Its source/toolchain/system identities match the Network Flow baseline. The
all-28 matrix is complete; fixture behavior remains unqualified pending S33-A–F.
Structural preview/write-input changes remain behind those comparison exits.
S32 closure documentation check is recorded under
`nf-imp08-s32-checkpoint-20261030-01` (see its native result before advancing).

S32 baseline commands (the `GO` override selects the already installed pinned
toolchain; it does not change repository pins or cryptographic policy):

```sh
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s32-networkflow-20261030-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.integration.pagination_recovery,module.networkflow.integration.bounded_graph_contributor_pipeline
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s32-imports-20261030-01 make test-slice OWNER=module.imports ROWS=module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.network_flow_default_limit_100001_rows,module.imports.integration.current_workflows,module.imports.integration.current_cancellation,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6
```

The Network Flow command passed with native exit 0, six work units and 21
successful cleanup steps (zero failed/blocked), under
`.cartulary/test-results/nf-imp08-s32-networkflow-20261030-02/`.
Canonical `make explain-run RESULTS_DIR=.cartulary/test-results/nf-imp08-s32-networkflow-20261030-02 DETAIL=summary`
reported source digest
`sha256:5567d2258b29691a01515029517e83298c9f09e5637cf2e07a84c49578b60fea`,
toolchain digest
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system digest
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
The source identity excludes Markdown by design; the dirty controlling tracker
is preserved separately. Existing fixture traversal remains **non-qualifying**
despite this pass. S32 corpus matrix Markdown checkpoint passed native exit 0:
`CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s32-markdown-20261030-01 make lint-markdown`,
summary `adhoc/lint-markdown/tool-run-summary.json` beneath that exact run root.

Implementation is **IN_PROGRESS**. No current product pass or corpus qualification
is claimed at entry. RB-006/RB-007 remain evidence gates until S32 and the relevant
fixture exits satisfy them; T20's fixture deferral is superseded by S33-A–F.
Unrelated production/release qualification remains outside this portfolio.
S36 must reconcile every finding and required correction, source identity,
exact evidence roots, compatibility/rollback, cleanup and any residual debt.

**S39 compatibility investigation update (no canonical runtime edit or migration yet):**
The correction affects more than new imports. Mapping fingerprints are embedded
in retained Imports approvals/plans/outcomes and committed Network Flow tables;
normalized digests feed row identities, binding row references and observation
provenance. Saved graph semantic digests and mutation receipt request hashes also
use the permissive serializer. The original migrations make rows and diagnostics
immutable, so an authored correction migration must explicitly preserve that
boundary outside its controlled transaction. Imports retains source units through
foreign keys, which supplies approved-mapping context for investigating repair.
Do not assume an idempotency request can be reconstructed from its stored hash:
receipt compatibility, nonterminal work, old cursors, saved projections, exported
state and rollback all require an explicit supported transition before changing
canonical bytes. No reset, translator, version change or irreversible conversion
has been implemented. The existing user authorization covers a necessary authored
correction migration once its compatibility design is complete. S39 remains
IN_PROGRESS; the failing expected bytes stay unchanged.

**Historical pre-S39 qualification boundary:** S37, S32, S38, S40 and S41 have their recorded
slice exits. S33-A is partial. S33-B–F, S34, S35 and S36 remain TODO; F36–F40 are
not collectively closed. The obsolete materialized apply validator and permissive
persisted-row construction remain. Four fixture bundles have been rewritten,
three have passing current behavioral evidence, and fixture 019 exposes F41.
The remaining 24 bundles still require replacement. Final `agent-finalize`,
source freeze, the complete affected portfolio, browser obligations, final
compatibility guidance and S36 handoff completion have not been performed.
Earlier source/schema/catalog/boundary passes retain their recorded historical
source identities and are not final qualification for the current working tree.

**S39 resumed compatibility decision:** canonical serialization will be checked
and type-directed: dynamic objects sort Unicode scalars, closed owner types use
schema order, and invalid strings/numbers/types return errors rather than
substitute bytes. Callers must propagate encoding failure before publication.
An authored transactional migration will recompute live table mapping identities,
normalized row digests/IDs, diagnostic identities, observation provenance and
binding references from retained approved mappings and normalized values. It
will advance affected table versions so old row cursors and graph snapshots
cannot claim unchanged source identity. Historical SQL remains immutable.

Committed receipts, apply outcomes and audit events are historical facts and
must not be rewritten to fabricate a different original response. Receipt replay
keeps current authorization and hash comparison; a formerly noncanonical request
whose corrected digest differs returns `client_txn_conflict`, never another
mutation. Clients must reload live resources after upgrade; historical response
row references are not aliases for corrected identities. No legacy serializer,
identity alias, automatic reset or request translator will be introduced.
Nonterminal Network Flow apply/materialization work must drain before migration;
unapplied approvals need renewed admission under the corrected fingerprint.
Migration preflight must reject incomplete retained mapping context atomically.
Old exported state that fails corrected owner validation requires controlled
restore-and-upgrade through its original database version before a new export;
current admission will not waive digest checks. Major 8/state 5 remain unchanged.
Rollback after migration requires the pre-upgrade database backup with the prior
binary; source-only rollback is insufficient once identities have changed.

Validation must independently cover exact canonical bytes, migration atomicity,
live reference preservation, stale cursors/projections, historical receipt
non-duplication and new-operation replay. This supersedes the earlier assumption
that the whole iteration requires no migration, not the original no-reset rule.


**F44 / S42 — numeric mapping tuple ordering (IN_PROGRESS):** NF-042 sorts
mapping tuples by source-column ordinal as an integer. `mappingSortKey` currently
sorts decimal strings, putting ignored column 10 before column 2. Fixture 019
revision 4 adds five independently authored mapping vectors to its existing
canonical comparisons. After S39 repaired closed system-derivation property
order, only `ignored-ordinal-order` fails: actual fingerprint
`42538f648b8a702e5b764bb0be568ee3f063ddfd9b1521a28398f896ec6558b1`, expected
`c3c995c1a521ec3e9de9acac58c1da19a1d2a205eb2d94d1f8233139acb65186`.
The reproduction is `nf-imp08-s39-mapping-vectors-20261030-02`: source
`sha256:3638fc7a4b3372aab791c3b865d457ab4ce4adf1dee822a99a2e0cc82a88fb52`,
native exit 2, product assertion failure, ten successful cleanup steps. Its exact
command used the pinned GO/installed-only prefix recorded above, this explicit
run ID, and `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721`.
The prior `...mapping-vectors-20261030-01` run (source
`sha256:fa5d0bae54717667eaefdb871ec2b12b83b26f3c02fcabeda737432d208e141a`)
also failed native 2 with ten successful cleanup steps, exposing closed property
order plus this ordering defect. Exact roots are `.cartulary/test-results/`
followed by those run IDs; comparison output is retained beneath their fixture
materialization directories and native summaries under `backend-unit/`.

S42 depends on the isolated failing reproduction, not completion of S33-A. It
will replace string-key ordering with a typed tuple comparator shared by
fingerprinting and preview mapping comparison. Its exit is the unchanged
independent vector plus affected preview/identity regressions passing with
cleanup. S39 migration depends on S42 so fingerprints are corrected once;
S33-A requires both exits. The compatibility impact is a fingerprint change for
mappings whose otherwise-equal tuple prefixes contain multi-digit ordinals;
S39's authored migration must incorporate numeric ordering. Leaving F44 open
would persist a second, incorrect canonical identity. No public schema or
legacy mapping translator is introduced.

**S39 receipt compatibility refinement:** immutable historical responses, apply
outcomes and audits remain unchanged. Saved graph admission recomputes request
hashes from retained receipt responses, so its reconstructible hash projections
must be corrected together with the matching retained job and commit-proof
request hashes. The migration must verify the original receipt identity first,
including same-name rename's two allowed base-version interpretations; it must
abort on ambiguous or inconsistent retained facts. This refines the earlier
blanket receipt-hash preservation statement. Unreconstructible mutation inputs
retain their original hashes and reject changed canonical replay hashes with
`client_txn_conflict`. Affected active saved graphs retain historical selected
results as stale; refresh is required before an export that needs current-source
reconstruction. No projection result is silently rewritten or fabricated.


**S42 completion checkpoint:** replaced decimal-string tuple keys with a typed
comparator used by mapping fingerprints and preview comparison. F44's runtime
correction is complete; durable transition remains a required S39 exit. The
unchanged independent fixture 019 revision 4 now passes, along with preview and
source identity regressions. Native exit 0, one work unit, ten successful cleanup
steps, no failed/blocked cleanup:

```sh
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s42-regression-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a
```

Evidence root: `.cartulary/test-results/nf-imp08-s42-regression-20261030-01/`;
source digest `sha256:46908e7d49066b59bc276cd13c4dfba34b3edf4efa056ac6d757ad9b14a08706`,
with the recorded pinned toolchain/system identities. Fixture execution summary
and produced `actual.json` are retained in its isolated `NF-FIX-019-*` directory.
S42 entry Markdown passed native 0 under `nf-imp08-s42-entry-20261030-01`.
Closure checkpoint uses `nf-imp08-s42-close-20261030-01 make lint-markdown` with
the same installed-only environment. S39 can resume after that native pass;
its migration, receipt and state-admission risks remain open. S33-A–F, S34,
S35 and S36 are not complete and no final-source qualification is claimed.


**S39 specification compatibility decision:** completing the requested canonical
repair requires a narrow exception to historical receipt byte preservation.
NF-REQ-032c now defines the authorized offline identity repair; NF-REQ-209's
read-only startup admission remains strict and does not run that repair.
Saved-graph receipts may contain a semantic-query checksum made with escaped
non-control Unicode. Repairing only request hashes would still leave those
receipts incompatible with canonical admission. The migration therefore
recomputes that derived checksum as well, after verifying the original hash,
without changing admitted values, versions, timestamps, target/job identities,
status, effects, audits or immutable projection bytes. This supersedes the
prior blanket statement that every response byte remains unchanged. Replay
returns the repaired historical receipt. No runtime legacy branch or dual-hash
acceptance is introduced. Other non-reconstructible request hashes remain
historical and conflict safely if the corrected client request differs.

The source owner now explicitly records this compatibility consequence and
backup-based rollback. The latest maintainer request authorizes resolving the
canonical defect and its necessary migration impact; this is the scoped
correction, not an unrelated change to read-only cutover behavior.

Migration registration added actual row
`module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`
through `make author-test-row-id` and authored family inputs. Generation run
`nf-imp08-s39-migration-generate-20261030-01` failed native 2 because allocation
still ended at 33; authored allocation/profile 34 corrected it, and `...-02`
passed native 0. Migration run `nf-imp08-s39-migration-20261030-01` failed native
2 during template preparation, before product callbacks: migration roles have
no TEMP privilege. Its source identity is
`sha256:2bc092f40bf732d4788433ed985db9cac1e403b9db70a8016c825000c4f06e58`;
all 12 cleanup steps passed. Helpers now use ordinary private objects created
and dropped within the migration transaction; privileges are not broadened.
Generation `nf-imp08-s39-migration-generate-20261030-03` passed native 0.
These are intermediate S39 diagnostics, not a completed migration exit.


**F45 / S43 — migration policy distinguishes DDL from procedural guards:**
`nf-imp08-s39-migration-generate-20261030-07` failed native 2 because the
migration policy's unqualified `IF NOT EXISTS` search labels a PL/pgSQL
existence guard as `permissive_create`. This is a false rejection of a required
commit-proof consistency check, not permissive object creation. S43 depends on
that failing reproduction and will restrict the projection's prohibition to
DDL clauses, retaining negative controls for conditional CREATE/ADD while
allowing procedural existence guards. S39 generation and migration qualification
now depend on S43. This tooling correction changes no product compatibility;
leaving it unresolved would force misleading SQL or block valid authored
migrations. Its exit requires the policy's exact selected unit row, generation
and documentation checks to pass before S39 resumes.

S39's closed graph semantic/time-range/aggregation projections now preserve
schema order after transport/jsonb round trips. Fixture 019 revision 5 adds
independent default and temporal semantic bytes, SHA-256, length-framed graph
digests and mutation digests; earlier vectors remain unchanged. Graph Projection
transport uses ordinary JSON, keeping its own normalization authority separate
from Network Flow's canonical digest codec. Current saved-graph fixtures now
use canonical bytes; explicit legacy bytes remain in the migration fixture.

The migration is authored as `00034_network_flow_canonical_identity.sql` with
allocation/profile registered through Make generation. Native run
`nf-imp08-s39-migration-20261030-06` passed the migration's repair, incomplete
mapping, late bad-receipt rollback and inflight-apply cases, including real
post-migration declaration and receipt admission: three work units, 14 cleanup
steps passed, source
`sha256:0984691f22ea25caa889af7679b4b5628b581415662c01e12e9e3627e4a1310a`.
It selected only
`module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`
with the pinned GO/installed-only prefix and that explicit run ID. Retained-job
and commit-proof cases are now being added; the earlier pass is not their
qualification. S39 is still IN_PROGRESS, and S33-A has not closed.


**S43 owner clarification and completion:** Core §2.1A, REQ-01-661 used a
blanket `IF NOT EXISTS` prohibition. Review clarified that wording to prohibit
conditional creation/alteration DDL while permitting procedural query guards.
This is an explicit specification cleanup, not a claim that tests overrule the
old wording. Generator and structural tests now match that scope. Negative
controls retain rejection of conditional tables, indexes and added columns;
a PL/pgSQL query guard is accepted. No runtime privilege, DDL idempotence or
schema admission boundary was loosened.

```sh
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s43-generate-20261030-01 make generate
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s43-policy-20261030-01 make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.production_ddl_v3_recurrence
```

Both commands passed native 0. Their exact roots are `.cartulary/test-results/`
plus the explicit run IDs; generation summary is `generate/tool-run-summary.json`,
and the selected policy row's native summary is `run-summary.json`. Entry Markdown
`nf-imp08-s43-entry-20261030-01` passed native 0. The closure command is
`CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s43-close-20261030-01 make lint-markdown`.
S43 becomes DONE only on that native pass, after which S39 may resume. Its
outstanding work is migration proof/tombstone, full compatibility regressions
and operational documentation; this checkpoint does not close S39 or S33-A.


S43 closure Markdown passed native 0. The policy run's source digest is
`sha256:f6388a24c7dc7f52206fd69054760fdd94b9f911d10ffeb42dfa89a4fc2b8d29`,
with the recorded pinned toolchain/system; its one work unit and all ten cleanup
steps passed. F45/S43 are resolved. This status-record checkpoint uses
`CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s43-close-20261030-02 make lint-markdown`
before resuming S39.

V36 records F41/F44's independent canonical vectors, authored migration,
rollback and owner admission, graph receipt/proof/tombstone and replay checks.
The final affected browser portfolio also includes actual row
`module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle`, because
canonical correction affects saved-graph creation and replay. This is a required
selection, not a passing claim. V37 records F45's policy positives/negatives and
migration catalog generation. Existing V29–V35 selections remain unchanged.


**S39 Imports authority refinement:** terminal unit approvals are the original
admission identity and must remain consistent with immutable apply plans and
outcomes. The draft's earlier update of applied `import_units.mapping_fingerprint`
is removed before closure. Network Flow tables acquire corrected fingerprints;
terminal Imports approvals/plans/outcomes retain their original fingerprints.
Unapplied approvals are invalidated for current admission. The migration tests
must compare these distinct authorities explicitly. Disposable empty-database
Down is a no-op, preserving Core REQ-01-661's harness cleanup contract; retained
affected data blocks Down and requires the documented backup/binary rollback.


**S44/S45 dependency insertion before correction work:** S39's migration and
complete retained-fact comparison passed `nf-imp08-s39-migration-20261030-09`
(native 0). `nf-imp08-s39-ddl-20261030-01` returned native 2: all bootstrap,
contamination, prerequisite and empty-database Down cases passed, but PostgreSQL
catalog parity found two nonexistent `nf34_rows_*` indexes in the generated
manifest. Cleanup completed all 15 steps. The projector processes all relation
events before indexes, incorrectly adding indexes after the owning table's drop.
F46/S44 will apply typed object events in authored source order, retaining its
existing dependency cascade, then regenerate through Make. The real migration
and PostgreSQL catalog comparison are the regression oracle; no SQL workaround
or weakened inventory expectation is permitted. S39's exit depends on S44.

Inspection also found F47: Core §2.1A's old reset-only/rewrite/head-29 prose
contradicts the already adopted REQ-01-661 baseline plus contiguous increments.
S45 will align that prose, preserving reset-only rejection across DDL lineages
and the explicit within-line NF-REQ-032c repair. This documentation correction
precedes S39's closure and requires its own Markdown checkpoint.


**S44 implementation exit evidence:** the catalog projector now gathers typed
create/drop events and applies them in source order, using its dependency cascade
for drops. It no longer resurrects objects dropped within the same migration.
No migration workaround, new parser framework or semantic interpreter was added.
`nf-imp08-s44-generate-20261030-01` (`make generate`) passed native 0;
`nf-imp08-s44-ddl-20261030-01` ran
`make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.production_ddl_v3_recurrence,module.database_migrations.integration.production_ddl_v3_recurrence`
and passed native 0, four units and all 14 cleanup steps. Source identity is
`sha256:99cf029d3a2a24c4b3187c6962985bdaf5e1022ffe2e3ffaf95f2c58ca7af92e`;
toolchain/system identities and the explicit GO/preparation prefix are unchanged.
Exact roots are `.cartulary/test-results/<run-id>/`, with generation
`generate/tool-run-summary.json` and test `run-summary.json`/`cleanup-results.json`.
PostgreSQL confirms exact inventory, prerequisites, contamination rejection and
zero-residue Down. F46's product risk is resolved; S44 awaits its documentation
checkpoint `nf-imp08-s44-close-20261030-01 make lint-markdown` before S45 starts.
V38 retains these DDL recurrence selectors and the final generated check.


S44's `nf-imp08-s44-close-20261030-01` Markdown checkpoint passed native 0.
S44/F46 are DONE; S45 is ready. The status-only checkpoint is
`nf-imp08-s44-close-20261030-02 make lint-markdown`; S45 begins after its pass.


**S45 documentation correction:** S44's status checkpoint passed native 0.
Core §2.1A now agrees with REQ-01-661: the v3 baseline and authored increments
remain unchanged; current-line owner-authorized changes append allocated
versions. The machine catalog determines the current head. Reset-only rejection
still governs transitions from another lineage, while NF-REQ-032c explicitly
requires preservation for this within-line identity repair. The pre-production
manifest boundary remains zero; no production-history or deployment claim is
introduced. F47's only changed areas are specification and this tracker. V39
requires owner-reference consistency and Markdown. The completion checkpoint is
`nf-imp08-s45-close-20261030-01 make lint-markdown`; no product/generation check
is required for this documentation-only slice.


S45's documentation checkpoint passed native 0; F47/S45 are DONE. Its summary
is `.cartulary/test-results/nf-imp08-s45-close-20261030-01/adhoc/lint-markdown/tool-run-summary.json`.
The status checkpoint `nf-imp08-s45-close-20261030-02 make lint-markdown` precedes
resuming S39. S39 retains only its final migration/compatibility validation and
checkpoint; no corpus completion or S36 qualification is implied.


S45's status checkpoint passed native 0; S39 resumes. Its final compatibility
wording preserves committed Imports approvals and all historical attempt plans/
outcomes. Current selected/mapped/ready approvals require renewal even if a
previous terminal attempt had a plan; active apply work still blocks migration.
The temporary draft exclusion for units with prior plans is removed. Repaired
historical graph receipts are now compared in full through the actual serving
receipt adapter and Auth storage reader; changed request hashes must conflict.


**S39 retained execution ledger (intermediate source states):** the commands below use
the recorded pinned GO and installed-only prefix and their exact explicit run IDs.
Native exits were observed separately from these public manifests. Source identity
and declared row selections are transcribed from each retained `run-manifest.json`;
all listed cleanup records completed with zero failed or blocked steps. Each root
is `.cartulary/test-results/<run-id>/`, retaining `run-summary.json`,
`run-manifest.json`, `cleanup-results.json` and owner-native result artifacts.
These are focused correction evidence, not the final S36 frozen-source portfolio.

| Run ID / native exit / cleanup steps | Exact Make command | Source digest |
| --- | --- | --- |
| nf-imp08-s39-canonical-20261030-01; exit 2; cleanup 11 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.import_preview_admission_f0dc5f2d8b` | `sha256:5fb8cfd60ecb7260880461ec38d60209126b3fd8d657884032fc2aa06d7ef267` |
| nf-imp08-s39-canonical-20261030-02; exit 2; cleanup 11 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.import_preview_admission_f0dc5f2d8b` | `sha256:b2ab55f5a86bd197bee9d533a6c9e7d6a4f8c7a96ce8077db1a7f82af20258d2` |
| nf-imp08-s39-canonical-20261030-03; exit 2; cleanup 11 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.import_preview_admission_f0dc5f2d8b` | `sha256:9de865f392cbbb1fa55416df08635306cabda87150a4c88d5d17a721fc3286cd` |
| nf-imp08-s39-canonical-20261030-04; exit 0; cleanup 11 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.import_preview_admission_f0dc5f2d8b` | `sha256:9353fc72220c02dff8c46386a23096aa94d4c0ccec138ebcc140e0636db9379d` |
| nf-imp08-s39-migration-20261030-01; exit 2; cleanup 12 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca` | `sha256:2bc092f40bf732d4788433ed985db9cac1e403b9db70a8016c825000c4f06e58` |
| nf-imp08-s39-migration-20261030-02; exit 2; cleanup 12 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca` | `sha256:898f5173e6610bb83709a9cce754d52d484c4fa31ef3fac6e78e944a8a62b60d` |
| nf-imp08-s39-migration-20261030-03; exit 2; cleanup 11 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.saved_graph_receipt_integrity` | `sha256:5418e6f29f4c69ffe25fec6ed416500a7f6775c80172b4edf4a15175020383a7` |
| nf-imp08-s39-migration-20261030-04; exit 2; cleanup 15 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.saved_graph_receipt_integrity` | `sha256:e7352291243b8ed4f197cf417b077559bfbba0d254c10401a2e19a05b6e7e025` |
| nf-imp08-s39-migration-20261030-05; exit 2; cleanup 19 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.saved_graph_receipt_integrity` | `sha256:0b8691d57c79dc3067b5da51e59fb5cbef98bd148f0e96a599e1485355833587` |
| nf-imp08-s39-migration-20261030-06; exit 0; cleanup 14 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca` | `sha256:0984691f22ea25caa889af7679b4b5628b581415662c01e12e9e3627e4a1310a` |
| nf-imp08-s39-migration-20261030-07; exit 0; cleanup 14 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca` | `sha256:f6388a24c7dc7f52206fd69054760fdd94b9f911d10ffeb42dfa89a4fc2b8d29` |
| nf-imp08-s39-migration-20261030-08; exit 0; cleanup 19 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.saved_graph_receipt_integrity` | `sha256:693ca94543c9cf3f2490933cd89f695a555f7278e1697048ff8ebad0d1a5269a` |
| nf-imp08-s39-migration-20261030-09; exit 0; cleanup 14 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca` | `sha256:55b20940514ff7595cca736907e56601f2f8846e72e05c8a8e7149019b1e54f6` |
| nf-imp08-s39-regression-20261030-01; exit 0; cleanup 10 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.network_flow_selector_covers_duplicate_csv_heade_61bc2b1f54,module.networkflow.unit.network_flow_selector_covers_interface_fields_st_6333a351b3,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.unit.network_flow_selector_covers_logs_telemetry_audi_61f27d1e10,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.table_lifecycle_contract,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b` | `sha256:9353fc72220c02dff8c46386a23096aa94d4c0ccec138ebcc140e0636db9379d` |
| nf-imp08-s39-completion-20261030-01; exit 2; cleanup 19 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.network_flow_selector_covers_duplicate_csv_heade_61bc2b1f54,module.networkflow.unit.network_flow_selector_covers_interface_fields_st_6333a351b3,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.unit.network_flow_selector_covers_logs_telemetry_audi_61f27d1e10,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.table_lifecycle_contract,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b` | `sha256:8ba9442e3189093e02aa06c3f54ffa5cbafe0ada7f9c47f31c66e9414f4d7ced` |
| nf-imp08-s39-ddl-20261030-01; exit 2; cleanup 15 | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.production_ddl_v3_recurrence,module.database_migrations.unit.production_ddl_v3_recurrence` | `sha256:55b20940514ff7595cca736907e56601f2f8846e72e05c8a8e7149019b1e54f6` |


**S39 final implementation and compatibility:** one checked canonical encoder
now preserves declared closed-object order, Unicode scalar order for dynamic
maps, shortest escaping, exact integers and explicit nulls. Invalid Unicode,
unsupported/fractional numeric values, duplicate raw JSON members and cycles
return errors; identity and graph/mutation callers propagate them. Closed graph
semantic/time-range/aggregation projections survive JSON transport without
losing their ordering. Graph Projection transport stays with its owning codec.
F44's numeric mapping tuple order is incorporated. Fixture 019 revision 5 adds
independent default/temporal graph, semantic checksum and mutation vectors.

The authored version-34 migration validates original mapping, row, diagnostic,
provenance, graph receipt, job and proof identities before atomic repair. It
preserves source identity, nullability, attribution/times, committed Imports
approvals, historical plans/outcomes, audit occurrences and immutable graph
results. It updates dependent live row/binding references, table versions and
reconstructible graph receipt/job/proof hashes, including expired-job proofs.
Missing mappings/proofs, tampered receipts/proofs/diagnostics and active apply
work roll back all repair writes and restore immutability protections. Helper
objects disappear before commit. Full retained-fact comparison excludes only
explicitly authorized derived changes. The serving receipt adapter compares
complete repaired responses and rejects changed request hashes.

Compatibility: stop prior writers and drain affected jobs, retain a recoverable
pre-upgrade backup, apply the authored current-line migration, then start the
matching binary. Unapplied approvals need renewed admission; corrected table
versions invalidate old cursors and graph source snapshots. Refresh selected
graphs before exports requiring current source reconstruction. Old exported
state that fails the corrected validator needs controlled restore/upgrade and
re-export; validation is never waived. Nonreconstructible historical mutation
hashes stay unchanged, and changed canonical replay conflicts without duplicate
effects. Major 8/state 5 and public shapes remain unchanged. Populated rollback
requires the matching backup and prior binary; Down supports only disposable
empty-state cleanup. No runtime legacy serializer, alias or automatic reset is
introduced. Version 1–33 SQL bytes remain unchanged.

Failure reconciliation: canonical runs 01–03 exposed private call-site signature
and transport-fixture assumptions before run 04 passed. Migration runs 01/02
failed fixture preparation because the purpose role lacks TEMP and an authored
SQL expression had an extra parenthesis; helpers now exist only inside the
Goose transaction, with PUBLIC execution revoked. Run 03 caught an int/int64
projection mismatch. Run 04 found obsolete graph receipt property ordering;
run 05 found an incorrect fixture path identity. Both fixtures were corrected
against their owner contracts. Runs 06–09 passed progressively expanded real
migration, admission, proof/tombstone and retained-fact checks. DDL run 01 exposed
F46, now closed by S44. Completion run 01 passed the migration/replay and state
rows but failed two old graph digest constants; independent Python SHA-256 and
unsigned-length framing reproduced their obsolete alphabetical ordering, then
computed the corrected closed-object vectors with valid 32-hex table IDs.
Completion run 02 passed both graph rows plus frozen fixture 019 (native 0).
Completion run 03 supplies a standalone native pass for the expanded migration/
replay obligation; its result is still pending. S39 remains IN_PROGRESS until
that result and its documentation checkpoint pass.

Generation runs `nf-imp08-s39-migration-generate-20261030-01..10` used `make generate`:
01 failed the stale allocation count, 07 failed F45's policy false positive;
02–06 and 08–10 passed native 0. The two failures returned native 2 and were
investigated before repair. S43/S44 retain their separate generator/policy fixes.
Format runs `nf-imp08-s39-format-20261030-01..04` used `make format` and passed
native 0. Generation summaries are under each exact root's
`generate/tool-run-summary.json`; format retains the graph run summaries.
No intermediate generation/format pass is final source qualification.


S39 completion run 03 passed native 0 with the full migrated-receipt replay
comparison. Final preflight review narrows the Imports job drain check to jobs
whose immutable unit plans include Network Flow, avoiding a maintenance
prerequisite on unrelated view imports. The Network Flow job and applying-unit
checks remain. Working storage and validation scans are proportional to retained
Network Flow rows; operators must budget the offline window and backup/storage
space. This is a one-time authored migration, not a runtime compatibility layer.
The changed SQL requires generation 11 and focused migration completion 04.


**S39 binary validation exit:** completion 04 returned native 0: three units,
14 completed cleanup steps, no failed/blocked cleanup. Its source digest is
`sha256:a9b845c86aa152b248c5d8fbd32a7665502175bf57a84b609d3ffcd7d2ce090b`;
commit/toolchain/system identity and the pinned command prefix remain unchanged.
Generation 11 also passed native 0. Migration, replay and rollback results are
retained at `.cartulary/test-results/nf-imp08-s39-completion-20261030-04/`.
The exact command was
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`.
Graph/semantic independent comparisons, prior boundary/state checks and S44's
actual DDL inventory/Down checks complete the focused correction obligations.
S39 awaits `nf-imp08-s39-close-20261030-01 make lint-markdown`; after that pass
F41 is resolved and S33-A can resume its four remaining fixture families.

The generic generation envelopes retain the exact command/pinned environment,
outputs and native outcome but do not publish a source digest; their paired
post-generation graph runs above identify the resulting source. They are not
used as frozen-source qualification. Final generated/schema/catalog checks are
still required after S36 finalization and source freeze.

| Additional run ID / native exit / cleanup | Exact Make command | Source digest |
| --- | --- | --- |
| nf-imp08-s39-completion-20261030-02; exit 0; cleanup 10 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.time_bucket_graph_backend` | `sha256:1f5aad161bed98b087345654370b0790be1f994b698f46a633c52e3dc7c6ca67` |
| nf-imp08-s39-completion-20261030-03; exit 0; cleanup 14 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca` | `sha256:1f5aad161bed98b087345654370b0790be1f994b698f46a633c52e3dc7c6ca67` |
| nf-imp08-s39-completion-20261030-04; exit 0; cleanup 14 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca` | `sha256:a9b845c86aa152b248c5d8fbd32a7665502175bf57a84b609d3ffcd7d2ce090b` |


S39 final lineage review adds explicit preflight rejection for a row's incident/
locator mismatch and a binding's missing, cross-table/cross-incident or inconsistent
indexed row references. The repair must not silently carry an unmatched reference
forward. A deliberately corrupted binding now exercises atomic rejection alongside
the existing identity/proof negatives. This completes the same S39 preflight
obligation; generation 12 and focused completion 05 validate the final SQL.


**S39 / F41 completed:** the final lineage negatives, repair, complete serving
replay, state admission and rollback passed native 0 in
`nf-imp08-s39-completion-20261030-05`: three units, all 14 cleanup steps completed,
source `sha256:550b26c7f12bc28b1b15f1aceb2934db3c7faa37393fd691b956c35568fade35`.
The exact command and selected row are unchanged from completion 04 above;
retained root is `.cartulary/test-results/nf-imp08-s39-completion-20261030-05/`.
Format 05 and generation 12 passed native 0 with the same pinned environment.
S39's first Markdown closure passed native 0; the final status/lineage checkpoint
is `nf-imp08-s39-close-20261030-02 make lint-markdown`, which must pass before
resuming S33-A. The compatibility and operational requirements above are part
of the handoff. Production migration/deployment is a separate activity.

**Current implementation boundary after S39:** S37/S32/S38/S39–S45 have their
recorded exits. Four bundles (005, 010, 019 and 021) have behavioral comparisons;
S33-A still requires 003, 011, 012 and 022. The other 20 bundles, corpus closure,
F36/S34, F37/S35 and S36 remain outstanding. F38/F40 close only after the entire
corpus consumes admission and evidence mechanics. Final source freeze, affected
browser rows, generated/schema/catalog checks and the complete affected portfolio
remain S36 obligations; intermediate passes are not final qualification.

### 27.18 S33-A resumption after S39

S39's final documentation checkpoint, `nf-imp08-s39-close-20261030-02 make
lint-markdown`, returned native 0. Its summary is
`.cartulary/test-results/nf-imp08-s39-close-20261030-02/adhoc/lint-markdown/tool-run-summary.json`.
HEAD remains `98adcc82961456bcb74bbecd0023136485afbf00`; all pre-existing working
changes and historical evidence remain preserved. S33-A now resumes 003, 011,
012 and 022. The comparison adapters will exercise actual source inspection,
mapping preparation, normalization and client alias selection. Inspection found
that ignored columns may still enter `unmapped_raw`; fixture 003 will establish
an independent failing reproduction before any runtime correction starts.

### 27.19 F48 / S46 — honor explicit ignored-column retention

**Finding F48:** Table 10-C and NF-REQ-090/091 require explicit ignored columns
to be excluded from preserved `unmapped_raw`. `unmappedRawValues` only consults
source-field mappings and therefore retains ignored values. Fixture 003's real
facade comparison reproduces exactly two unwanted raw entries (ordinals 4 and 3)
and their incorrect normalized digests; every other expected field matches.
This is a retention/privacy and identity defect, not a compatibility feature.

**S46 is IN_PROGRESS**, prerequisite the failing fixture reproduction, with
S33-A acceptance dependent on its passing exit. Repair the normalization boundary
using the complete approved mapping's source dispositions. Extend the **unreleased,
untracked migration 34 authored in this iteration**, before final source freeze,
to remove explicitly ignored retained values and atomically recompute affected
normalized identities/binding references. Migrations 0–33 remain untouched.
Consolidating the repair avoids inventing support for an intermediate, never
released migration 34 and avoids two offline scans of the same identity graph.
The prior S39 passes remain evidence for their exact earlier source only.

Areas: implementation, authored migration, migration/fixture tests, specification
repair procedure and tracker. The mapping fingerprint, source bytes/digests,
committed Imports facts, audit and attribution remain unchanged except for S39's
already documented canonical repair. Remove only provenance explicitly disallowed
by the original approved mapping. No row alias, automatic reset, or runtime
legacy translator is introduced. Cached row references and graph snapshots can
become stale; table versions advance and users refresh affected saved graphs.
Rollback remains backup plus matching prior binary; preserving prohibited values
would continue violating user-selected retention and conceal the defect in IDs.

**V40 / binary exit:** fixture 003 passes complete independent comparisons; the
migration fixture uses actual ordinal-keyed provenance, proves removal of ignored
values, exact new row identities and binding updates, preservation of unrelated
facts, and rollback on inconsistent context. Existing canonical migration/replay
checks and focused DDL projection checks pass. Update this checkpoint and lint
Markdown before resuming S33-A. Separate preserved-provenance canonical ordering
and owner prose will be checked by subsequent independent corpus cases.

Reproduction command (usual pinned GO/preparation environment):
`CARTULARY_TEST_RUN_ID=nf-imp08-s33a-ordinal-repro-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771`.
Native exit 2; product assertion failure; all 10 cleanup steps completed. Source
`sha256:d8590af3afc7ee08f5a0caabf0cbae8bad87e5151615115ad0db7ad028423253`.
Root `.cartulary/test-results/nf-imp08-s33a-ordinal-repro-20261030-01/` retains
`run-manifest.json`, `run-summary.json`, cleanup, and the fixture's
`NF-FIX-003-duplicate-headers-1408606065/comparison-summary.json` and
`produced/actual.json`. The registered row ID was generated through
`make author-test-row-id FAMILY_ID=module.networkflow.unit CLAIM='Ordinal fixture compares source inspection mapping approval and normalized rows' SELECTOR_KEY=TestNetworkFlowOrdinalFixtureComparison`.

S33-A resumption Markdown, ordinal generation/format and the post-generation
`make task-guide OWNER=module.networkflow` passed native 0. An initial task-guide
started before row registration failed catalog completeness (native 2); registration
and generation resolved it. An attempted `nf-imp08-s33a-generate-20261030-01`
invocation was rejected before execution because that historical root already
exists; it was preserved. Fresh `nf-imp08-s33a-ordinal-gen-20261030-01` passed.
No check was widened or failed expectation weakened.

**S46 validation exit:** source dispositions now exclude both mapped and explicitly
ignored ordinals from retained provenance. Migration 34 validates the original
closed ordinal/header/value records, removes only explicitly ignored entries,
then derives new digests/row IDs and updates binding references in its existing
atomic repair. The migration vectors now contain genuine ordinal-keyed raw
provenance; an additional malformed-provenance case proves rollback. Complete
retained-fact comparisons exclude only the authorized raw removal, which has its
own exact comparison; failed migration preserves the original raw records.
NF-REQ-032c documents this narrow repair exception and the existing offline
backup/refresh compatibility requirements remain in force.

`nf-imp08-s46-correction-20261030-01` passed native 0, four units and all 15 cleanup
steps, using `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`.
Source `sha256:bd2b90a6c471da6799be0cac11768f31db4de3d2c3d99fbb2f30e4f9a565b565`;
root `.cartulary/test-results/nf-imp08-s46-correction-20261030-01/` retains the
run/cleanup summaries and fixture 003 comparison evidence.

`nf-imp08-s46-ddl-20261030-01` passed native 0 using
`make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.production_ddl_v3_recurrence,module.database_migrations.integration.production_ddl_v3_recurrence`.
Source `sha256:bd2b90a6c471da6799be0cac11768f31db4de3d2c3d99fbb2f30e4f9a565b565`; all 15 cleanup steps completed.

Its exact root is `.cartulary/test-results/nf-imp08-s46-ddl-20261030-01/`.
S46 format/generation and entry Markdown also passed native 0. All commands use
the established pinned GO and installed-only environment; source commit,
toolchain and system identities remain unchanged.

S46/F48 are DONE subject to status checkpoint
`nf-imp08-s46-close-20261030-01 make lint-markdown`; S33-A resumes after that
native pass. The affected final portfolio includes V40 and the existing import
recovery browser obligation. No production deployment or data migration has
been performed by these disposable-database tests.

S46's closing Markdown checkpoint returned native 0; S33-A resumed. Fixture 003
now advances from frozen revision 3 to 4 to add a preserved-provenance comparison
with ordinals 2 and 10, independently authored schema-table value ordering and
Unicode cell values. This checks the adjacent canonical retention boundary;
revision 3's successful ignored-column comparison remains historical evidence.

### 27.20 F49 / S47 — canonical retained-provenance boundary

**Finding F49:** fixture 003 revision 4 exposes a second canonical defect:
`unmapped_raw` is serialized as nested dynamic maps, losing NF-REQ-106's numeric
ordinal order and Table 12-A1's closed member order. Only the normalized digest
differs; complete facade responses and semantic raw values match. The independent
expected digest is `089eca2754b8a65d1fbf48855f09dd5d37618c2b9676bb78ba456e3315199118`;
the erroneous digest is `a005dba994eeadc9a9d8554674fe1ef58c5af073ceeb6f58d61c7c00526bee16`.

The same boundary has stale owner prose: NF-REQ-105's broad exclusion from
identity contradicts the explicit `normalized_row_digest_v1` framing in §6.5.
The remediation clarifies NF-105 against that precise framing: retained
provenance participates in integrity without becoming a semantic flow field.
NF-032b will explicitly distinguish NF-106's typed ordinal object from generic
Unicode-ordered dynamic dictionaries. This preserves the adopted digest inputs;
it does not add a provenance promotion feature or an alternate algorithm.

**S47 is IN_PROGRESS**, depending on this failed reproduction; S33-A acceptance
now also depends on S47. Introduce a private typed retained-value object and
ordinal-keyed collection, with strict shape/hash/ordinal admission and explicit
canonical projection. Use it in normalization and digest construction. Extend
the unreleased migration 34 with the same numeric/closed serialization and a
second independently authored migration fixture proving preservation (alongside
S46's ignored-value removal). Historical migrations remain immutable.

Areas: specification clarification, implementation, authored migration, fixtures,
tests and documentation. Benefit: one explicit provenance type owns its shape,
ordering and integrity; generic maps cannot silently reinterpret it. Existing
nonconforming retained identities are repaired offline with the already required
backup, reference updates and table-version invalidation. No runtime dual codec,
new public interface or state-version change. Leaving this unresolved makes
persisted identities depend on implementation-specific map order and weakens
state reconstruction. V41 requires exact fixture digest/bytes, malformed typed
value negatives, preserved and ignored migration outcomes, rollback, and DDL
projection checks before the closing Markdown checkpoint and S33-A resumption.

Reproduction: `CARTULARY_TEST_RUN_ID=nf-imp08-s33a-preserved-repro-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771`, with the established pinned environment.
Native exit 2, product assertion failure, all 10 cleanup steps completed; source
`sha256:083f0c6a56b95cbb962e066ad1e1ebddccbd8fd764e64d839c4977791515139d`.
Root `.cartulary/test-results/nf-imp08-s33a-preserved-repro-20261030-01/` contains
run identity/summary/cleanup and fixture directory
`NF-FIX-003-duplicate-headers-1369048688/` with comparison summary and produced
actual output. No expected value was copied from the implementation.

**S47 validation exit:** normalization now produces `retainedRawValues` containing
closed `retainedRawValue` records. Its canonical projection orders numeric ordinals
and schema-table members explicitly. Strict decoding rejects missing/extra/null
members, noncanonical ordinals, duplicate members and inconsistent value/header
hashes. The normalized digest consumes that type directly; generic dictionary
ordering remains unchanged. NF-032b/NF-105 clarify the precise existing digest
and typed-object rules, and NF-032c includes this offline repair scope.

Migration 34 now uses the same typed ordinal/closed projection. Independently
framed fixtures exercise both ignored-value removal and preserved-value repair,
with complete unrelated-fact preservation, row/binding/diagnostic identities,
serving receipt replay, state admission and atomic rollback. No runtime legacy
codec or additional intermediate migration was added.

`nf-imp08-s47-correction-20261030-01` passed native 0, four units and all 15 cleanup
steps. Exact command: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`.
Source `sha256:a356e1b0e4c22b97f8e2c26525b1fe69d40ca7b5817604b66c0b65330cd2a87e`;
root `.cartulary/test-results/nf-imp08-s47-correction-20261030-01/` retains run
summaries, cleanup, and fixture 003 revision 4 / fixture 019 revision 5 evidence.

`nf-imp08-s47-ddl-20261030-01` passed native 0 with
`make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.production_ddl_v3_recurrence,module.database_migrations.integration.production_ddl_v3_recurrence`.
Source `sha256:a356e1b0e4c22b97f8e2c26525b1fe69d40ca7b5817604b66c0b65330cd2a87e`; all 15 cleanup steps completed.

Exact root `.cartulary/test-results/nf-imp08-s47-ddl-20261030-01/`.
Format/generation and entry Markdown passed native 0. The established source
commit, pinned toolchain/system and installed-only environment apply throughout.

S47/F49 are DONE subject to `nf-imp08-s47-close-20261030-01 make lint-markdown`.
After its native pass, S33-A resumes 011, 012 and 022. V41 joins the final affected
portfolio; the final migration/source freeze and qualification remain S36 work.

S47's closing Markdown checkpoint passed native 0. S33-A resumes with uptime
012 and timestamp/ruleset 022, each advancing from revision 2 to 3. Expectations
are independently framed from explicit CSV cells and owner rules. Before deriving
New York transition cases, the vendored `tzdata2026c.tar.gz` was verified against
the authored provenance artifact: 475694 bytes, SHA-256
`e4a178a4477f3d0ea77cc31828ff72aa38feff8d61aa13e7e99e142e9d902be4`.
Its US rules establish March `Sun>=8` and November `Sun>=1` transitions; the
fixture will retain that archive/provenance and execute the actual pinned parser.
Registered IDs from `make author-test-row-id` are
`module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352`
and `module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb`.
They are registrations, not passing evidence.

### 27.21 F50 / S48 — timestamp scalar and precision admission

**Finding F50:** independent fixtures 012/022 expose six mismatched cases:
uptime accepts an RFC3339 export clock finer than its declared millisecond
precision; explicit-offset and IANA-zone conversion admit UTC year 0000 or
10000; canonical epoch integers beyond uint64 are mislabeled `invalid_syntax`
instead of `out_of_range`. All other current case outputs match. Owners are
Table 6-D `timestamp_utc_v1`, NF-REQ-082–084, and the timestamp reason registry.

**S48 is IN_PROGRESS**, prerequisite the failing comparisons; S33-A acceptance
depends on S48. Enforce the UTC scalar range at the common timestamp entry,
propagate numeric overflow classification, and apply declared uptime precision
to its RFC3339 export clock. Extend migration 34's fail-closed context checks to
reject retained timestamps inconsistent with approved precision. Invalid absolute
years already fail its canonical timestamp construction; validate that rollback
explicitly. Do not round timestamps, delete rows, reinterpret source bytes, or
fabricate replacement approvals. Invalid historical data blocks upgrade and
requires authoritative source/context remediation under a separately reviewed
repair, with the pre-upgrade snapshot retained; it cannot be silently made valid.

Areas: implementation, fixtures/tests, authored migration, compatibility guidance
and tracker. Benefit: one UTC admission boundary and consistent precision prevent
invalid values from reaching storage or later state reconstruction. Valid values,
public schemas, digests and state versions remain unchanged. Newly rejected
inputs were outside the adopted contract; diagnostic reason correction affects
only oversized epoch failures. Unresolved, invalid rows can be committed and
prevent reliable graph/query/state operation and migration. V42 requires exact
012/022 comparisons, closed timestamp variants and uptime ordinal/bounds cases
inside 022 itself, migration invalid-context rollback and focused DDL checks.
The existing import recovery and query-authoring browser obligations cover the
affected workflows in S36.

Reproduction (established pinned environment):
`CARTULARY_TEST_RUN_ID=nf-imp08-s33a-time-repro-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb`.
Native exit 2, product assertions, all 10 cleanup steps completed. Source
`sha256:fb508a88018277fafc7d0d426006f5294bc784d37a784a10cd91c173c71a8207`;
root `.cartulary/test-results/nf-imp08-s33a-time-repro-20261030-01/` retains the
run/cleanup summaries and each bundle's comparison summary/actual output.
Time fixture format and generation passed native 0. Expectations were not
weakened. Fixture 022 will advance to revision 4 for additional self-contained
uptime/closed-variant coverage required by its registry; fixture 012 remains 3.

S48's serving correction passes all 15 uptime cases and all 54 timestamp cases
(revision 4 now executes its own uptime/closed-variant scenarios). Before adding
the migration precision guard, an independently coherent legacy uptime dataset
proved the existing migration incorrectly accepted finer-than-approved retained
timestamps. `nf-imp08-s48-migration-repro-20261030-01` returned native 2 solely
for that intended negative; the prior repair/rollback cases and new invalid-UTC
negatives passed. All 15 cleanup steps completed. Exact command:
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`.
Source `sha256:ce762dd63b369b8b07a8c29ec4b8d6e50a41b1ba9278baa34eeb98239cd13d91`; retained root

`.cartulary/test-results/nf-imp08-s48-migration-repro-20261030-01/`.
Migration-only permissive diagnostic stubs were also replaced with valid closed
IP diagnostics and independent IDs. Failed migration now compares the entire
affected durable-state set, including Imports journals/outcomes, receipts,
proofs, audit and migration ledger, before and after the attempt.

**S48 validation exit:** the common timestamp entry now validates the resulting
UTC year, including explicit offsets and pinned IANA-zone conversion. Epoch
numeric overflow retains its range classification. Uptime's RFC3339 export
clock uses the declared millisecond precision. Valid normalization and existing
closed-variant/ordinal regressions pass unchanged.

Migration 34 now rejects stored timestamp precision inconsistent with the
approved profile; its existing UTC-range rejection is covered explicitly.
Independent legacy uptime vectors first demonstrated erroneous migration success,
then passed the new fail-closed expectation. Complete affected-state snapshots
prove failed attempts preserve rows, provenance, bindings, Imports state,
receipts/proofs/jobs, audit and ledger. The operator must resolve invalid source
context through an explicit reviewed repair before retrying upgrade; no rounding,
row removal, reset or replacement approval is performed.

`nf-imp08-s48-correction-20261030-01` passed native 0, four units, all 15 cleanup
steps. Exact command: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.unit.network_flow_selector_covers_timestamp_profile_p_bdf924da13,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`.
`nf-imp08-s48-ddl-20261030-01` passed native 0, four units, all 15 cleanup steps,
using `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.production_ddl_v3_recurrence,module.database_migrations.integration.production_ddl_v3_recurrence`.
Both source digests are
`sha256:9abdd701e8ffb9aaab67faf71ea5dc5fd283ba4678058a0bae3fde5f6a24a8ba`.
Each exact `.cartulary/test-results/<run-id>/` retains identity, run summary,
cleanup and applicable fixture comparisons. Format 02, generation 01 and entry
Markdown passed native 0 in the established pinned environment.

S48/F50 are DONE subject to `nf-imp08-s48-close-20261030-01 make lint-markdown`.
S33-A then resumes fixture 011's backend/client alias comparison. Seven of its
eight bundles have behavioral comparisons (003, 005, 010, 012, 019, 021, 022);
this does not close the remaining corpus groups or final qualification.
V42 joins S36's affected portfolio. No production migration was executed.

### 27.22 S49 — client fixture admission and evidence

S48's closing Markdown checkpoint returned native 0. Fixture 011 requires the
actual web mapping model as well as backend admission; Go-only fixture support
cannot qualify that client comparison. S49 extends F40's narrow mechanics to the
existing Node test runner before S33-A resumes 011. This is test infrastructure,
not a new product finding or service coordinator.

Implement a private frontend test-support interface for strict manifest/scenario
schema admission, exact inventories/order/sizes/hashes, immutable run-local input
copies, bounded produced artifacts and schema-validated retained summaries.
Reuse the authored strict JSON parser and generated JavaScript schema validators.
Register support source ownership and exact selectors through authored inputs,
then generate with Make. Keep alias semantics entirely in owner tests.

Dependencies: S38's evidence/schema contract and S48's completed checkpoint.
Risks: filesystem escape or race, callback after failed admission, shared output
paths, and evidence publication falsely reported as success. Binary exit: direct
negative controls cover traversal/absolute paths, symlinks, duplicates, missing/
extra files, ordering/size/hash/bundle mismatches, draft/revision rejection,
readonly copies, isolated concurrent callbacks and publication failure; schema,
source-ownership and affected routing checks pass. V35 gains this actual selector
when authored. Close S49 and lint Markdown before 011's behavioral work resumes.
Compatibility: test-support additions only; existing manifest/scenario v2 and
execution-summary v1 remain unchanged. Reversion removes the support/selector
additions without touching durable product data.

**S49 checkpoint before dependent correction:** the Node support and its controls
are implemented. The authored selector is
`web.architecture.boundary_support.node_fixture_admission_seals_verified_inputs_and_d6cec88240`
(author invocation `nf-imp08-s49-author-20261030-01`, native 0). Existing
`web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`
was selected alongside it. `nf-imp08-s49-controls-20261030-01` ran
`make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.node_fixture_admission_seals_verified_inputs_and_d6cec88240,web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`
and passed native 0 (two units). `nf-imp08-s49-types-20261030-01`
(`make lint CHECK=types`) and `nf-imp08-s49-catalog-20261030-01`
(`make source-check CHECK=catalog`) also passed native 0. The selected source is
`sha256:09af7ebaebdc72ec18caf623105625b6ebc3eacb8cb49a8059b40f036f15136c`;
HEAD, toolchain/system identities and explicit GO/preparation prefix remain as
recorded above. These are intermediate evidence, not S36 qualification.

`nf-imp08-s49-format-20261030-01` failed at runtime-plan configuration (native 2,
no published summary). `nf-imp08-s49-generate-20261030-01` failed native 2;
its retained generator log identified unsorted authored family rows. Sorted that
authored array, then `nf-imp08-s49-generate-20261030-02` and
`nf-imp08-s49-format-20261030-02` passed native 0. The failed roots remain intact.
`nf-imp08-s49-entry-20261030-01` and owner task-guide also passed native 0.

`nf-imp08-s49-schema-20261030-01` (`make source-check CHECK=schema`) failed
native 2 with all 12 cleanup steps complete: the independent schema shape check
still processes relation drops before index creates, reporting absent staging
indexes `public.nf34_rows_old_id` and `public.nf34_rows_row_id` as missing from
the correctly generated inventory. This is a second affected consumer of F46,
not a new migration defect. S49's exit remains pending this required correction.
Exact artifacts use `.cartulary/test-results/<run-id>/`; graph runs publish
`run-manifest.json`, `run-summary.json`, `cleanup-results.json`; generation uses
`generate/tool-run-summary.json` and the referenced nested logs.

### 27.23 S44-R — finish F46 across the independent schema checker

Reopen F46/S44 for this consumer; the earlier generator/PostgreSQL passes remain
historical evidence. Prerequisite is the retained S49 schema failure, not S33-A
completion. Apply the checker's existing typed create/drop events in source order
so dependent objects disappear with their table. Preserve independent authored-
SQL verification and PostgreSQL parity; do not add nonexistent ownership entries
or change SQL to accommodate the checker. This is tooling/projection repair only,
with no additional migration or product compatibility effect.

Binary exit: schema and generated-projection checks pass against migration 34,
and existing production DDL unit/PostgreSQL recurrence rows still pass. Risk:
incorrect ordering could misclassify other object kinds; the complete catalog
comparison remains required. V38 now includes the schema check explicitly.
Close this correction and run Markdown before resuming S49's final checkpoint.

**S44-R exit:** the independent checker now queues each existing typed object's
mutation and applies the complete DDL stream in source order, retaining its
relation-dependent removal semantics. No SQL or ownership entries changed.
Entry Markdown and format runs (`nf-imp08-s44r-entry-20261030-01`,
`nf-imp08-s44r-format-20261030-01`) passed native 0. The first schema/generated
checks failed native 2 because the checker edit invalidated the generated input
index (12/14 cleanup steps complete); the index mismatch was investigated before
`nf-imp08-s44r-generate-20261030-01 make generate`, which passed native 0.

At resulting source
`sha256:ab4fbf6f11f9ca6e743421813680ab8ccfe507549155c14777bc7b857c6b27ef`,
all of these commands passed native 0 with the established GO/preparation prefix:

| Run ID | Exact Make selection | Cleanup |
| --- | --- | --- |
| `nf-imp08-s44r-schema-20261030-02` | `make source-check CHECK=schema` | 12 completed |
| `nf-imp08-s44r-generated-20261030-02` | `make source-check CHECK=generated` | 14 completed |
| `nf-imp08-s44r-ddl-20261030-02` | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.production_ddl_v3_recurrence,module.database_migrations.integration.production_ddl_v3_recurrence` | 14 completed |

DDL run 01 also passed native 0 at pre-regeneration source
`sha256:466362e3cc58552621fe271dab77ffe1cad2e82ecffff6457514d19db8fddabc`
(15 cleanup steps); run 02 establishes compatible evidence after regeneration.
Graph evidence remains at `.cartulary/test-results/<run-id>/run-summary.json`
and accompanying manifests/cleanup records. Toolchain/system identities are
unchanged. S44-R/F46 are resolved, subject to the closing documentation checkpoint
`nf-imp08-s44r-close-20261030-01 make lint-markdown`. No additional compatibility
or migration impact remains. S49 resumes only after that checkpoint passes.

### 27.24 S49 completion checkpoint

S44-R's closing Markdown check passed native 0. The Node support remains private
to `web.testing`; its executable support/ownership rows belong to
`web.architecture`. It reuses the strict JSON parser and generated schema registry,
opens verified regular files, copies in bounded chunks, seals inputs 0400/0500,
keeps each callback's outputs separate, and publishes a bounded v1 summary.
Product behavior stays in owner adapters. Synthetic controls use fixture 999,
which is outside the 28-bundle product inventory.

Review added cross-inventory duplicate and linked-corpus controls. It also exposed
Node's default BOM stripping: `nf-imp08-s49-bom-repro-20261030-01` ran the support
row alone and failed native 2 because invalid input qualified (11 cleanup steps
complete). The decoder now preserves the BOM for strict JSON rejection. This
correction is confined to the new S49 helper; it changes no product decoder.
`nf-imp08-s49-format-20261030-03 make format` passed native 0.

At source
`sha256:164649e9e9d9af9b7a05a646a2739a2f72cf4208fa37d4d8662fed0feb8d4f1e`,
the following commands passed native 0, using the established explicit
GO/preparation prefix and unchanged HEAD/toolchain/system identities:

| Run ID | Exact Make selection | Units / cleanup |
| --- | --- | --- |
| `nf-imp08-s49-controls-20261030-02` | `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.node_fixture_admission_seals_verified_inputs_and_d6cec88240,web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` | 2 / 11 completed |
| `nf-imp08-s49-types-20261030-02` | `make lint CHECK=types` | 2 / 11 completed |
| `nf-imp08-s49-schema-20261030-02` | `make source-check CHECK=schema` | 3 / 12 completed |
| `nf-imp08-s49-catalog-20261030-02` | `make source-check CHECK=catalog` | 2 / 11 completed |

Exact evidence roots are `.cartulary/test-results/<run-id>/`, with
`run-manifest.json`, `run-summary.json`, `cleanup-results.json` and per-control
`NF-FIX-999-admission-controls-node-*/comparison-summary.json`. The intentional
publication-failure control proves failure without fabricating a success summary.
Mutable attack corpora were removed; verified copies and comparison artifacts
remain. No product migration or interface impact was introduced.

S49's binary exit is satisfied, subject to
`nf-imp08-s49-close-20261030-01 make lint-markdown`. F40 remains open for complete
corpus integration in S33-F. S33-A resumes fixture 011 after this checkpoint;
seven of its eight bundles currently have independent comparisons. S33-B through
S36 remain pending. These focused results do not establish final qualification.

### 27.25 S33-A fixture 011 reproduction and correction dependencies

S49's closing Markdown checkpoint passed native 0. Fixture 011 now uses real
backend source inspection/mapping preparation/record normalization and the actual
web mapping model. Its independent expectations cover exact source descriptors,
NFC/whitespace/ASCII-only alias keys, fingerprints, retained raw values, source
reuse rejection, lowest-ordinal suggestions, explicit collision dispositions,
takeover reconfirmation and field-ordered preview candidates. Client expectations
and transcripts are separate manifest-listed artifacts consumed by the Node
owner adapter; the backend adapter compares its complete applicable artifacts.

Registered with `make author-test-row-id` (native 0):

- `module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0`
- `web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309`

The corresponding author run IDs are
`nf-imp08-s33a-alias-server-author-20261030-01` and
`nf-imp08-s33a-alias-client-author-20261030-01`. Authored family/source-ownership
inputs were updated and `nf-imp08-s33a-alias-generate-20261030-01 make generate`
passed native 0. Both owner task-guide invocations and format runs 01/02 passed
native 0.

Reproduction 01 failed native 2 in both adapters. Investigation found a fixture
stimulus mistake: an embedded BOM in the CSV was correctly rejected before alias
normalization. Removed that invalid CSV stimulus, retained the BOM match-key
vector at the direct algorithm boundary, and advanced frozen revision 3 to 4.
The client corpus excludes that unreachable CSV-header case. Revision 4 contains
20 backend and 17 client comparisons; no product expectation was weakened.

At source
`sha256:89f12974e4031916703bc8809df8fb44afeb1631ba009ed6c2a74066382fd7d2`,
these exact reproductions failed native 2 with all cleanup complete:

| Run ID | Exact Make selection | Cleanup |
| --- | --- | --- |
| `nf-imp08-s33a-alias-server-repro-20261030-02` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0` | 10 completed |
| `nf-imp08-s33a-alias-client-repro-20261030-02` | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309` | 11 completed |

Retained comparisons are beneath the exact roots in
`NF-FIX-011-alias-collision-4184409308/` (server) and
`NF-FIX-011-alias-collision-node-IxYz4z/` (client), with
`comparison-summary.json` and referenced produced artifacts. Server differences
show missing NFC for Kelvin-sign `PacKets`, decomposed `é` and Angstrom-sign
headers, then the consequent mapping/normalized digest differences. Client
outputs additionally omit the deterministic default, serialize source order,
and lose required collision reconfirmation after reassignment.

| Finding | Adopted owner and consequence | Correction / validation |
| --- | --- | --- |
| F51 | NF-REQ-080a/080b and Table 10-B: backend alias keys omit NFC. New approvals can receive incorrect descriptors and fingerprints. | S50: pinned Unicode 17 NFC before exact whitespace/ASCII folding; fixture 011 backend comparison; historical-context migration/replay and state regression. |
| F52 | NF-REQ-080b Table 9-F1 steps 5–7 / NF-AC-074: client aliases omit NFC, duplicate defaults/order are wrong, and explicit dispositions are incompletely tracked. | S51: cohesive client suggestion/confirmation state, field-order serialization/display, actual modal/browser workflow and client comparison. |

### 27.26 S50 — canonical alias keys and historical approval boundary

Prerequisite: the valid-source server reproduction above. S33-A acceptance now
depends on S50 and S51; neither correction waits for S33-A completion. S50 repairs
new source inspection and approval using the existing pinned Unicode normalizer.
Fingerprint framing itself remains unchanged: it consumes the approved mapping's
recorded descriptors, rather than reinterpreting historical approvals. NF-REQ-032c
requires committed Imports approvals/outcomes to remain immutable; migration 34
already invalidates unapplied approvals for current readmission. Add an independent
retained non-NFC approval case to verify this boundary and the existing canonical
identity repair/replay. Do not add a database Unicode normalizer or rewrite
historical descriptors merely to match new suggestions.

Compatibility: new approvals for affected headers receive correct match keys and
therefore correct fingerprints. Existing applied facts retain their original
approved inputs; no new public shape, digest framing version or state version is
planned. Confirm this decision with migration/state/replay evidence before closing.
Risk: accidentally normalizing historical fingerprint inputs would drift retained
identities or invalidate replay. Binary exit: all 20 backend comparisons pass;
independent migration identities and immutable historical inputs/replay pass;
state admission and relevant existing mapping tests pass. V43 tracks F51.

### 27.27 S51 — client alias suggestions and explicit confirmation

Planned after S50's checkpoint, using the retained client reproduction. Preserve
raw headers and ordinal identity while applying the adopted alias algorithm.
Keep the lowest ordinal as a suggestion while requiring explicit confirmation of
all duplicate dispositions; re-open confirmation when reassignment changes a
previously confirmed disposition. Provide a direct UI action to confirm the
current suggested or unmapped choice. Serialize/display suggestions in the owner
field order without turning column labels into identities. Keep product logic in
the owner model/panel and fixture mechanics in the S49 support.

Compatibility: private draft-state changes and corrected mapping-modal behavior;
no wire/durable-state shape change. Risks: accidental approval, moving selection
without renewed confirmation, and UI state disagreeing with serialized candidates.
Binary exit: all client fixture comparisons, affected model/panel regressions and
an actual browser alias-resolution workflow pass; no automatic import effect
occurs before approval. Author any new browser selector before execution. V44
tracks F52. Update this workstream with concrete implementation/evidence before
resuming S33-A closure. S33-B through S36 remain pending.

### 27.28 S50 completed — new NFC keys, immutable historical approvals

S50 is **DONE**, F51 resolved and V43 passing for this source. `mapping.go` now
uses the existing pinned Unicode 17 NFC implementation before the exact whitespace
trim and ASCII-only fold. The semantic import boundary explicitly admits that
pure generated dependency. No fingerprint framing, SQL migration, durable state
version or public schema changed in this correction.

Added an independently authored retained-context migration vector with a real
CSV source, decomposed-accent and Kelvin-sign headers, duplicate accepted records,
Unicode raw values and a rejected record. Historical approved descriptors retain
their old non-NFC keys. Independent legacy/current framing pins repaired IDs and
digests; the actual facade separately confirms the new approval fingerprint.
All ten migration scenarios run for this vector: repair/replay and nine abort/
rollback cases. Existing retained-context vectors and state admission also pass.
Committed approvals/outcomes remain unchanged, while migration 34's existing
pending-approval reset provides readmission under corrected normalization.

Commands use the §27 environment prefix and exact run roots
`.cartulary/test-results/<run-id>/`; source is
`sha256:d94d12222e8fe33a50642f255c9f7dd7eeb2b571bb40643f57854cb41fbf18c7`.
Toolchain/system identities remain those recorded for IMP-08. All following native
exits were observed as 0; target summaries and cleanup records are retained:

| Run ID | Exact Make command | Cleanup |
| --- | --- | --- |
| `nf-imp08-s50-format-20261030-01` | `make format` | completed |
| `nf-imp08-s50-guide-20261030-01` | `make task-guide OWNER=module.networkflow` | selection only |
| `nf-imp08-s50-core-20261030-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.unicode17_text_canonicalization_3de608c3dc` | 18 completed |
| `nf-imp08-s50-preview-20261030-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.import_preview_admission_f0dc5f2d8b` | 10 completed |
| `nf-imp08-s50-boundary-20261030-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.query_authoring_admission` | 10 completed |

The core root retains all 20 backend 011 comparisons in
`NF-FIX-011-alias-collision-1872095402/comparison-summary.json` and its referenced
actual artifact. Exact-root `make explain-run ... TARGET=test-slice DETAIL=summary`
confirmed pass identities and cleanup for core, preview and boundary runs; the
corresponding observer run IDs use `s50-<selection>-explain-20261030-01`.
The earlier `nf-imp08-s50-entry-20261030-01 make lint-markdown` has a passing tool
summary, but its native session exit was lost across continuation; it is not
claimed as an observed native pass. The closing checkpoint below replaces that
missing documentation evidence.

Compatibility risk is bounded to newly corrected match keys for affected new
approvals. Historical inputs are intentionally not normalized again. The existing
migration-34 drain, backup, fail-closed admission and rollback guidance remains
applicable. S51 is ready after the closing Markdown checkpoint; S33-A still waits
for client behavior and its browser workflow. S33-B through S36 remain pending.
Closing command: `nf-imp08-s50-close-20261030-01 make lint-markdown`; outcome is
recorded at the next checkpoint before S51 implementation.

S50 closing Markdown check passed with observed native exit 0 and retained
`adhoc/lint-markdown/tool-run-summary.json`. S51 is now **IN_PROGRESS** against
the existing revision-4 client reproduction and V44 selections. No changes to
fixture expected values are authorized merely to match the implementation.

S51 focused implementation evidence (not closure): default selection, NFC/exact
whitespace matching, persistent collision membership/reconfirmation, explicit-null
handling and field-ordered candidates now pass all 17 client fixture cases plus
existing model, controller/recovery and workspace tests. The UI uses a direct
confirmation action and stable suggestion order. The browser selector now checks
reassignment, default/null confirmation, out-of-source-order display, NFC aliases
and unchanged durable table listing before apply.

At source `sha256:02ace684a1982dec953c5890e592a6359495eae9ec449aaa089c00f02f245912`,
`nf-imp08-s51-model-20261030-01 make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.networkflowimportmodel_suite_12249943c5,web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a`
passed native 0 (12 cleanup steps); retained client comparison:
`NF-FIX-011-alias-collision-node-NExCQ6/comparison-summary.json` beneath that root.
`nf-imp08-s51-types-20261030-01 make lint CHECK=types` failed native 2 (11 cleanup
steps): the new sorter incorrectly used the full wire mapping union, which also
allows system derivations. Repaired by ordering source columns before constructing
wire mappings. `nf-imp08-s51-browser-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619`
failed native 2, normalized product/assertion failure, all 23 cleanup steps
completed. After confirmation, preview remained disabled; added explicit checks
for each remaining collision to localize the UI path before another correction.
The browser group result references its exact Playwright report and failure
context; no timeout widening, expectation weakening or golden refresh occurred.
An observer attempt using the child projection name as `TARGET` failed usage
validation; `TARGET=test-slice` summary/accounting and referenced artifacts were
used instead. Format runs 01–04 and owner task-guide passed native 0.

After the type repair, `nf-imp08-s51-types-20261030-02 make lint CHECK=types`
and `nf-imp08-s51-model-20261030-02` with the same five-row selection passed native
0. Browser reproduction 02 is pending; S51 and S33-A remain open.

### 27.29 S51 completed — explicit collision choices in the actual modal

S51 is **DONE**, F52 resolved; V44 passes for source
`sha256:28fd610dd48f521c4b174966767e4f85eaf6b5a16e32f4128cb5ae810a85a58a`.
The final model uses NFC/exact Unicode whitespace/ASCII folding, keeps the lowest
ordinal suggestion, retains original collision membership and reopens displaced
choices. Explicit unmapped choices resolve under preserve/ignore policies;
reject policy still blocks them. Wire mappings and stable suggestion display use
the authored field order followed by ordinal. Confirmation moves focus to the
existing target select before its button disappears, preserving an actionable
control during the next choice. Raw labels and ordinal identities remain intact.

Browser reproduction 02 failed native 2 with 23 cleanup steps complete. Its
more precise assertion showed the mapped confirmation disappearing but the
unmapped confirmation remaining. The final focus-preserving handler passes the
same full workflow, including both confirmations, takeover/reconfirmation, NFC,
field order, unchanged durable catalog before apply and successful explicit
publication. Earlier failures remain evidence, not passing qualification or a
claim of flakiness. No public/durable schema, digest or migration changed in S51.

All commands below use the §27 environment prefix; all native exits were observed
as 0. Exact roots are `.cartulary/test-results/<run-id>/`, with
`target-summaries/<target>.json`, `cleanup-results.json` and referenced row reports.
Toolchain/system identities remain unchanged.

| Run ID | Exact Make command | Cleanup |
| --- | --- | --- |
| `nf-imp08-s51-model-20261030-03` | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.networkflowimportmodel_suite_12249943c5,web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a` | 12 completed |
| `nf-imp08-s51-types-20261030-03` | `make lint CHECK=types` | 11 completed |
| `nf-imp08-s51-browser-20261030-03` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619` | 23 completed |

Format runs 05/06 passed native 0. Exact-root summary observers for the three
passing runs use `nf-imp08-s51-<selection>-explain-20261030-03`; their native 0
means diagnosis succeeded. Browser status observers 01–03 retained their own
limitations and were not substituted for native exits. Client comparison evidence
is `NF-FIX-011-alias-collision-node-Mb1mTV/comparison-summary.json` beneath the
model root. No fixture bytes or expected values changed during this correction.

Remaining risk: subsequent work changes source identity and requires the final
compatible portfolio; these focused passes do not establish S36 or full-owner
qualification. S33-A is ready to resume and close its eight-bundle comparison
portfolio after this documentation checkpoint. Closing command:
`nf-imp08-s51-close-20261030-01 make lint-markdown`; record its native outcome
before resuming S33-A. S33-B through S36 remain pending.

S51 closing Markdown checkpoint passed with observed native exit 0. S33-A resumes
at the same source identity to run its combined eight-bundle backend selection;
its client obligation is the compatible passing S51 model/browser evidence above.

### 27.30 S33-A completed — parsing, admission and normalization

S33-A is **DONE**. At source
`sha256:28fd610dd48f521c4b174966767e4f85eaf6b5a16e32f4128cb5ae810a85a58a`,
`nf-imp08-s33a-closure-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb`
passed with observed native exit 0 and all 10 cleanup steps complete. The exact
root is `.cartulary/test-results/nf-imp08-s33a-closure-20261030-01/`; its target
summary, row reports and cleanup are retained. The same environment identities
apply. `nf-imp08-s33a-explain-20261030-01` confirmed the exact-root summary.

| Bundle | Frozen revision | Backend outcomes | Retained comparison directory |
| --- | --- | --- | --- |
| 003 | 4 | 4 | `NF-FIX-003-duplicate-headers-3145116218` |
| 005 | 3 | 10 | `NF-FIX-005-csv-parser-edges-3224738318` |
| 010 | 4 | 12 | `NF-FIX-010-json-admission-537273948` |
| 011 | 4 | 20 | `NF-FIX-011-alias-collision-120025348` |
| 012 | 3 | 15 | `NF-FIX-012-sys-uptime-timestamps-2200901232` |
| 019 | 5 | 47 | `NF-FIX-019-canonical-json-unicode-3771358696` |
| 021 | 3 | 19 | `NF-FIX-021-preview-boundaries-2537009686` |
| 022 | 4 | 55 | `NF-FIX-022-timestamp-rulesets-4026232802` |

Each directory contains `comparison-summary.json`, immutable admitted inputs and
referenced actual artifacts. Every backend outcome compares its complete
applicable expected artifact and concrete transcript; wrong-expectation controls
reject mismatches. The compatible S51 model run adds 17 independently compared
client outcomes for 011, and its actual browser workflow passes. Timestamp source
provenance is one of 022's 55 outcomes. These totals describe this group, not the
remaining corpus. All corrections required by S33-A have passing exits above.

Compatibility and residual risks are those of S39–S51, especially migration 34's
retained-data preflight and immutable approval boundary. Revised fixture history
remains in Git; there is no legacy mapping translator. S33 aggregate remains
**IN_PROGRESS**, with 8 of 28 families behaviorally rebuilt. F38/F40 stay open
until S33-F. S33-B (001/002/004/020/023) is ready after this closing documentation
check; S33-C–F, S34/F36, S35/F37 and S36 remain pending.
Closing command: `nf-imp08-s33a-close-20261030-01 make lint-markdown`; record its
native result before beginning S33-B implementation.

S33-A closing Markdown checkpoint passed with observed native exit 0. S33-B is
now IN_PROGRESS: rebuild 001/002/004/020/023 through production import publication
and durable-state comparisons, retaining independent fixture inputs and expected
identities. Its binary exit remains the live workstream table above.

### 27.31 S33-B publication adapter — in progress

S33-A's closing checkpoint is complete. S33-B has begun with fixture 001. The
owner integration adapter admits/materializes the bundle before starting a
claimed application. Supported Auth fixtures establish a fixed actor/session;
the initial database context contains an existing incident and one discovered,
source-only CSV unit. It seeds no approval, apply plan, outcome, or Network Flow
row. Actual Imports inspection, preview, approval, selection, apply, and replay
routes own those subsequent effects. Existing Network Flow randomness controls
pin the table UUID; the existing route fixture pins the clock and digest keys.
Opaque Jobs IDs are validated and checked across replay, with complete transport
responses retained separately; Network Flow identities remain exact expectations.

Fixture 001 now contains one real Cisco row, a current mapping, and independently
calculated expected rows, provenance, HMAC, and complete v2 graph output. Revision
3 was authored, then revision 4 corrected the independently reviewed Graph
Projection direction registry before the first execution. No fixture acceptance
is claimed yet. Subsequent publication families remain pending.

Registered through `make author-test-row-id` with family
`module.networkflow.integration`, claim
`Minimal fixture compares Imports publication provenance graph and replay`, and
selector `TestNetworkFlowMinimalImportFixtureComparison`:
`module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0`.
The authored family includes Imports and Graph Projection collaborators.

All commands use the established Go 1.27.3/installed-only environment and fresh
`CARTULARY_TEST_RUN_ID` values. Native exits observed as 0:

- `nf-imp08-s33b-minimal-row-20261030-01`: the row-authoring command above.
- `nf-imp08-s33b-generate-20261030-01`: `make generate`.
- `nf-imp08-s33b-format-20261030-01`: `make format`.
- `nf-imp08-s33b-guide-20261030-01`: `make task-guide OWNER=module.networkflow`.

First comparison command:
`CARTULARY_TEST_RUN_ID=nf-imp08-s33b-minimal-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0`.
Its native outcome, source identity, retained summary, and cleanup accounting are
pending reconciliation. S33-B remains IN_PROGRESS; S33-C is not ready.

The first comparison exited natively 2 (test assertion), with all 14 cleanup
steps passing. `explain-run ... TARGET=test-slice DETAIL=summary` and `DETAIL=logs`
under observer IDs `nf-imp08-s33b-minimal-explain-20261030-01` and
`nf-imp08-s33b-minimal-logs-20261030-01` both exited 0. Source identity was
`sha256:f15afefa8f81ed6e7e834bc819995ed2e0a2d42414541cd7611ed0a5e723679b`.
The source-inspection request returned 401: the seeded July session was expired
against the application test clock. The route helper's dependency clock does not
set that application clock. The adapter now explicitly uses the existing
`httptestx.SetClockFixed` control before seeding the session. This is a fixture
setup repair, not a product requirement change. No fixture bytes changed.

S33-B investigation supersedes the preceding clock setup. Runs
`nf-imp08-s33b-minimal-20261030-02` through `-06` each exited natively 2 with
14 successful cleanup steps. The selected row and command were unchanged from
`-01`. Run `-06` traced the failure to Imports' authoritative PostgreSQL lease
fence: freezing the application in July produced a lease already expired against
`clock_timestamp()`. The final adapter pins only Network Flow's existing injected
business clock; authentication and Jobs retain wall time. Temporary database and
facade tracing wrappers have been removed. This changes fixture setup only.

Run `nf-imp08-s33b-minimal-20261030-07` completed publication and then rejected
an incorrect replay assertion (native 2, cleanup 14/14). Core 01 §3.3.9 and the
Imports job contract return the original job with its current authoritative
state. The adapter now compares the entire replay response with the terminal
job and independently requires unchanged durable effects. Run `-08` reached the
complete comparison (native 2, cleanup 14/14); every independent digest, row,
provenance value, and graph identity matched. Two authored labels differed:
Network Flow §18 requires `cartulary.network_flow.import_unit_result.v1`, and
§26's selector contract requires `default_edge`. Those expectation corrections
advance fixture 001 to revision 6, including the intervening configured safe
key ID correction at revision 5. No product behavior was changed to fit data.
All runs retain their own source/environment manifests, summaries and cleanup
artifacts under `.cartulary/test-results/<run-id>/`. Per-run `explain-run` logs
commands with observer IDs `nf-imp08-s33b-minimal-explain-20261030-02` through
`-08` exited 0; `-01` used the separately recorded observer pair above.

Fixture 002 revision 3 is authored independently for trimmed/null interfaces,
IPv6 normalization, duplicate raw records with distinct identities, full graph
aggregates, and replay. Its registered selector is
`module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923`
(`TestNetworkFlowInterfacesImportFixtureComparison`), created by
`nf-imp08-s33b-interfaces-row-20261030-01` with `make author-test-row-id` (native 0).
`nf-imp08-s33b-generate-20261030-02 make generate` exited 0. Neither fixture is
claimed qualified yet. S33-B remains IN_PROGRESS.

### 27.32 S33-B reproduction checkpoint and S52 correction entry

The 001/002 comparison command was
`CARTULARY_TEST_RUN_ID=nf-imp08-s33b-publication-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923`.
Native exit 0; cleanup 14/14; source
`sha256:0bf156f82999bc9c48d6cba0d195c7d7ed52aa5fb3c90ea8c7355b4af163e03e`.
The established toolchain/system identities are unchanged. Evidence root:
`.cartulary/test-results/nf-imp08-s33b-publication-20261030-01/`, including
`target-summaries/test-slice.json` and both comparison summaries. This qualifies
001 revision 6 and 002 revision 3 at that source only. Their current revisions
7/4 add exact durable error fields to the common comparison and need rerunning.

The shared adapter now isolates each case in a subtest and retains transport
artifacts on failure. Replay checks distinguish the finalized success receipt
from an unchanged failed-operation admission receipt, then compare a fresh GET
with the complete terminal resource and require unchanged durable effects.
This narrows the earlier statement about current-state replay: failed admission
receipts retain their initial resource; they are not authoritative job-status
reads. No runtime replay behavior was changed.

Fixture 004 revision 3 first executed under
`nf-imp08-s33b-rejected-20261030-01` (native 2; cleanup 14/14), exposing that
fixture replay assumption. Source:
`sha256:70545edb418d483b057776f64bc571e307c54f1c66b33434d21b012e1584ed8d`.
After correcting the assertion, run
`nf-imp08-s33b-rejected-20261030-02` reached the complete comparison (native 2;
cleanup 14/14), at source
`sha256:5b19a32930dec3f61d5ce6b0e6bf711802593bc8b6e0a3bc876d4407128b10c5`.
Both used `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015`.
Both `explain-run` summary/log observers, named
`nf-imp08-s33b-rejected-{summary,logs}-20261030-{01,02}`, exited 0.
The exact failing comparison evidence is
`.cartulary/test-results/nf-imp08-s33b-rejected-20261030-02/NF-FIX-004-rejected-rows-2257363317/comparison-summary.json`,
with admitted/materialized inputs and retained `produced/actual.json` plus both
case transports. The partial-import output matches independently authored
expectations. The all-rejected job and durable outcome each omit ten required
members from both sampled diagnostics.

One separate expected-value error is corrected in fixture 004 revision 4:
failed outcomes retain an empty owner-result object, not JSON null, consistent
with the existing immutable migration 19 object/array constraint. This does not
alter the missing-diagnostic-field expectation or remedy its reproduction.

**F53 — All-rejected error samples lose the diagnostic resource.** Adopted
Network Flow §8.2/NF-REQ-061, §12.4/Table 12-E, and §21.2/Table 21-B require the
complete safe diagnostic resource, including explicit nullable members. The
runtime reduces it to five members, and the downstream owner-error schema
repeats that incomplete shape. Consequences include missing stable identities,
redaction hashes, safe numeric samples, message keys, and limit metadata.

**S52 — Complete all-rejected diagnostic publication**, IN_PROGRESS after this
reproduction/checkpoint. Dependency: the failing 004 comparison above. S33-B
acceptance now depends on S52's passing exit; S33-C remains not ready.

- Remediation/areas: reuse the existing complete diagnostic presenter; share
  closed diagnostic decoding and semantic/privacy validation; correct the
  authored owner-error schema and regenerate projections; remove the fabricated
  legacy diagnostic from the union test; add missing-member, unsafe-value,
  unknown-token and ordering negatives.
- Rationale/benefit: one diagnostic representation and validation boundary for
  previews, retained rows and failed imports prevents divergent policy.
- Compatibility: new errors gain fields already required by the adopted v1
  contract. No version, SQL-column or migration change is planned. The live
  `ValidateImportUnitError` hook is called only during owner-error translation;
  it is not a retained-outcome rewrite path. Existing failed receipts and Jobs
  error summaries remain historical snapshots. Verify replay/read compatibility
  before closing; do not rewrite historical proof hashes or outcomes.
- Risk: relaxing admission could leak raw source values; tightening it against
  the obsolete shape could silently replace the useful owner error with Core's
  generic failure. Keep 004's complete independent expectation and validate the
  public job plus stored outcome through Imports.
- V45/binary exit: fixture 004 passes both cases, including complete error
  samples, no partial table, stable replay and exact diagnostic counts;
  closed-union, preview-admission and streaming-retention regressions pass;
  authored/generated schema checks and documentation checkpoint pass. The
  affected import-recovery browser row remains required in the final portfolio.

V45 uses the actual registered 004 row above and existing rows
`module.networkflow.unit.import_owner_error_translation_rs08_3d3f02e050`,
`module.networkflow.unit.import_preview_admission_f0dc5f2d8b`, and
`module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd`.
Row authoring (`nf-imp08-s33b-rejected-row-20261030-01`), generation
(`nf-imp08-s33b-generate-20261030-03`), format
(`nf-imp08-s33b-format-20261030-09`), and task guide
(`nf-imp08-s33b-guide-20261030-03`) all exited 0. Workstream exit remains open.

S52 entry documentation check `nf-imp08-s52-entry-20261030-01 make lint-markdown` exited 0.

S52 compatibility refinement before changing the migration guard: regenerating
the corrected owner-error schema changes the complete Imports registry digest
from `b7bd50637cb33fb417d0e490f52d8add3203c45bee130a6d35cd4c449beda859`
to `660bd3742382a8e67bdbbcc12b66fede41736b11149a77004dfd451b27536ba9`.
`lockApplyUnitTx` compares admitted plans with that complete digest, including
non-Network-Flow targets. The operational prerequisite must therefore drain or
cancel **all** import-apply jobs, plus Network Flow jobs, before the upgrade.
Extend the guard in this iteration's new, unreleased migration 34 to enforce
that prerequisite. Migrations 0–33 remain immutable. No second migration or
stored-outcome conversion is needed; committed facts and receipts remain
untouched. Add a guard-specific negative control without a Network Flow plan,
verify the complete database rolls back, and rerun the migration row and
schema/DDL checks as part of V45. This supersedes S52's initial assumption that
no migration-file change would be needed; it does not change SQL columns,
major 8/state 5, or the backup-and-matching-binary rollback requirement.

### 27.33 S52 validation and correction exit

The runtime now reuses `diagnosticResource` for all-rejected errors. Closed
import-object decoding is shared by source inspection, preview and diagnostic
admission; diagnostic semantics/privacy checks are shared with retained-state
validation. The owner-error admission additionally enforces nonempty bounded
samples, strict ordering, distinct diagnostics and consistent rejected-record
counts. The authored schema declares every Table 12-E member. Union negatives
cover every missing member, raw/unsafe values, malformed identities/hashes,
unknown field/error/reason tokens, ordering, duplicates, empty/overflow samples
and count mismatches. The fabricated `network_flow_invalid_ipv4` positive vector
was replaced with a valid diagnostic and retained only as a rejection control.

Fixture 004 is revision 5. Its three isolated serving cases compare partial
publication, all-rejected failure, and 26 rejected records producing 52 complete
preview diagnostics and exactly 50 ordered apply-error diagnostics with
truncation. Complete job errors, durable outcomes and stable replay are compared;
failed apply publishes no table. Its successful admission/comparison summary is
`.cartulary/test-results/nf-imp08-s52-validation-20261030-03/NF-FIX-004-rejected-rows-345169627/comparison-summary.json`.
The summary binds manifest hash
`f76dd447bd267c62de770ceed0c507595c5058535ca35d2b47908775464bb04f`,
materialized inputs, complete actual results and all three concrete transports.

Migration 34 now rejects every nonterminal `import.apply_v1` job, as well as
Network Flow jobs and applying Network Flow units, before identity repair. The
new ordinary-apply guard case asserts that exact prerequisite failure and full
rollback. Existing repair, invalid-context, immutability, historical-alias,
retained-proof and replay checks remain passing. No historical migration,
committed outcome, historical error or receipt is rewritten by this correction.
Inspection of Imports translation and Jobs/receipt reads confirms that strict
owner-error admission applies to newly returned owner errors; historical Core
error details remain opaque stored snapshots. Previously incomplete samples do
not acquire fabricated missing fields or become current-format qualification.

All commands below used the established explicit Go 1.27.3 and installed-only
preparation environment. The two compilation failures were caused by leftover
private decoder names during the extraction, investigated through published
logs and repaired before another run:

| Run ID | Native result / cleanup | Source identity |
| --- | --- | --- |
| `nf-imp08-s52-validation-20261030-01` | 2; compile/preflight failure; 11/11 cleanup | `sha256:1dd8d65ad1ede2ade0c85667adb08d744f5533bafc22a3fe881222fd6639b379` |
| `nf-imp08-s52-validation-20261030-02` | 2; remaining test decoder references; 18/18 cleanup | `sha256:171e319df2605c24875274bbdf866d8a82f982122063e375cd7c78e9b7ae2a08` |
| `nf-imp08-s52-validation-20261030-03` | 0; all five selected rows pass; 18/18 cleanup | `sha256:c091438fbddde91f96e158fd32e67ab746b865658f737665e69f611752f05816` |

All three used the exact selection:
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.unit.import_owner_error_translation_rs08_3d3f02e050,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`.

Additional passing evidence at source `c091438f…`, with unchanged toolchain and
system identities from §27.32:

| Run ID | Exact Make command | Native / cleanup |
| --- | --- | --- |
| `nf-imp08-s52-generated-20261030-01` | `make source-check CHECK=generated` | 0 / 14 completed |
| `nf-imp08-s52-schema-20261030-01` | `make source-check CHECK=schema` | 0 / 12 completed |
| `nf-imp08-s52-catalog-20261030-01` | `make source-check CHECK=catalog` | 0 / 11 completed |
| `nf-imp08-s52-ddl-20261030-01` | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.unit.production_ddl_v3_recurrence,module.database_migrations.integration.production_ddl_v3_recurrence` | 0 / 15 completed |

Each root is `.cartulary/test-results/<run-id>/`, with its run identity,
`target-summaries/<target>.json`, row evidence and `cleanup-results.json`.
The summary observers `nf-imp08-s52-summary-20261030-{01,02,03}` and
`nf-imp08-s52-{generated,schema,catalog,ddl}-summary-20261030-01` exited 0, as did
both compilation-log observers. Generation runs
`nf-imp08-s52-generate-20261030-{01,02}`, format runs
`nf-imp08-s52-format-20261030-{01,02}`, and the `platform.postgres` and
`module.database_migrations` task-guide runs exited 0.

V45's implementation, comparison, migration and projection exits are satisfied.
S52/F53 await only this closing Markdown checkpoint. Once it passes, S33-B may
resume with 020/023 and publication/audit closure. S33-C, S33-F, F36/F37 and S36
remain open; none of this evidence represents final-source qualification.

S52 closing documentation check `nf-imp08-s52-close-20261030-01 make lint-markdown` passed native 0. S52 is DONE, F53 resolved, and V45 passes at the recorded source. S33-B resumes; its 020/023 comparisons and audit/publication closure remain outstanding.


### 27.34 S33-B — complete import publication comparisons

All five publication bundles now execute independently through actual Imports
routes, Network Flow's facade, borrowed transactions, storage and query routes.
001/002 additionally compare the complete Graph Projection v2 result. Frozen
revisions are 001=8, 002=5, 004=6, 020=3 and 023=3. Each frozen-byte change
advanced its revision; prior forms remain historical Git evidence.

The shared typed adapter compares source inspection, repeatable side-effect-free
preview, approval, selection, terminal job result/error, complete persisted
rows/diagnostics/provenance, table/query resources, and replay. Publication checks
bind the frozen plan and outcome to source, mapping, actor, incident, job and
unit-commit identity, require one apply receipt/job/outcome and no generic record
journal, and compare session/unit state. Audit semantic members are exact;
opaque event IDs/timestamps/request IDs are validated, retained in full, and
required to remain byte-for-byte stable on replay. Time/entropy use existing
supported controls; no production determinism hook was added.

020 exercises all four existing commit fault boundaries with an accepted row and
a rejected row producing two diagnostics. Before/after-owner errors publish only
the Core failure outcome. Before-commit failure exposes zero table, row,
diagnostic, binding, applied-outcome and domain-audit effects before real Jobs
recovery; recovery publishes once. Lost commit acknowledgement also resolves to
one successful publication. Expected recovery behavior is fixed, not selected
from the observed failure count.

023 contains 15 isolated source-change cases: expected digest, mapping
fingerprint, all six capability members, first/second-pass header and content
changes, first/second-pass filename changes, and second-pass capability change.
Existing composition/source ports inject changes after real approval. Exact safe
`network_flow_source_changed` translation, complete rollback, stable replay and
absence of the private replacement path are required. Inputs/expectations were
authored from literal semantic values and independent SHA/HMAC/framing rules;
no implementation-generated snapshot was accepted as an oracle.

The new selectors were authored through `make author-test-row-id`, registered in
`tools/test_families/module.networkflow.json`, and projected through Make:

- `module.networkflow.integration.atomic_fixture_compares_every_commit_boundary_ro_a4cdb920d4`
- `module.networkflow.integration.source_changed_fixture_compares_every_source_ver_213b5f7eab`

All commands use the established explicit Go 1.27.3 and installed-only environment.
`nf-imp08-s33b-atomic-validation-20261030-01` ran the first four rows below,
returned native 0 and completed all 14 cleanup steps at source
`sha256:0fa6b402f8f85b01d9e3f5afff16ba7fe1f3dc4909357a7f433ece36715a12f5`.
After adding the source port adapter,
`nf-imp08-s33b-source-validation-20261030-01` ran this exact command, returned
native 0 and completed all 14 cleanup steps:

`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.atomic_fixture_compares_every_commit_boundary_ro_a4cdb920d4,module.networkflow.integration.source_changed_fixture_compares_every_source_ver_213b5f7eab`

The second invocation qualifies all five bundles/24 concrete cases at source
`sha256:0831220341ca2f0ee016ece49eaf88205b50fb97b5b7afe0faa0c8ddad2f775b`,
HEAD `98adcc82961456bcb74bbecd0023136485afbf00`; toolchain/system identities
remain those in §27.33. Its root is
`.cartulary/test-results/nf-imp08-s33b-source-validation-20261030-01/`, with
`target-summaries/test-slice.json`, selected-row evidence and
`cleanup-results.json`. Comparison summaries bind each manifest, materialized
input and complete produced comparison/transports:

| Bundle | Relative comparison summary | Manifest SHA-256 |
| --- | --- | --- |
| NF-FIX-001 | `NF-FIX-001-cisco-sna-minimal-2777033800/comparison-summary.json` | `d16582b1b7a79b62f05a7794873af5954efc8432449b2dc5b005cf4ebfade648` |
| NF-FIX-002 | `NF-FIX-002-cisco-sna-interface-fields-1484193153/comparison-summary.json` | `9ab9720e77c9612a1c3d747e61ee9e9f5a46a664f4ae743630bcf9ea2318c1c9` |
| NF-FIX-004 | `NF-FIX-004-rejected-rows-609075431/comparison-summary.json` | `deee3c770b3d9345def7727d1c100af1b8be57bac6dc59e075767aa5e3093313` |
| NF-FIX-020 | `NF-FIX-020-atomic-import-commit-1245772776/comparison-summary.json` | `292e7ae738b7b2a58f853610bd79ef7347b0ec7a1562cb49ff14cbf9384f1dfa` |
| NF-FIX-023 | `NF-FIX-023-import-facade-source-change-801799697/comparison-summary.json` | `539e97258c94a6f01bd54caae7a8c323ed6ac03dcd14eaa1c9b23f32d986d555` |

Both exact-root summary observers exited 0. Supporting run IDs
`nf-imp08-s33b-{atomic,source}-{row,generate,format,guide}-20261030-01` all exited
0: row authorship used the selectors/claims recorded in the authored family;
generation used `make generate`, formatting `make format`, and selection
`make task-guide OWNER=module.networkflow`. No failed verification remains in
this continuation. Deliberately wrong expectations remain required by every
comparison adapter.

S33-B's implementation and comparison exits are satisfied, pending the closing
Markdown check below. There is no additional public interface or migration
change in this workstream. S52's migration-34 drain prerequisite remains in
force. S33-C is next after the checkpoint; S33-F, F36/F37 and final-source S36
qualification remain open. Earlier-source passes are historical evidence.

S33-B closing check `nf-imp08-s33b-close-20261030-01 make lint-markdown` passed native 0. S33-B is DONE. S33-C is now IN_PROGRESS; its four comparisons are the next dependency.


### 27.35 S33-C query comparison checkpoint and S53 correction entry

The typed fixture adapter now supports additional real imports in one runtime,
separate route transaction IDs, session-bound cursor inspection and complete
lifecycle audit comparison. It still seeds only discovered source prerequisites;
all approvals, plans, rows and outcomes use Imports. 014 exercises byte-exact
semantic cursor payloads, full keysets, another authorized actor, rename survival
and terminal null cursors. 024 adds all three scope variants, closed-variant
negatives, normalized duplicate filters and `in` values, mixed query bodies,
4096/4097-byte cursor boundaries, exact expiry and independent diagnostic tuples.
The crypto primitive opens authenticated payloads using fixture keys; no product
cursor policy is reimplemented and no randomness hook was added.

Retained cursor capabilities require redaction under the existing harness policy.
`nf-imp08-s33c-paging-validation-20261030-01` returned native 2 after its product
summary passed: finalization failed with `artifact/artifact_error`, while all
14 cleanup steps completed. Source was
`sha256:dd96501773b7436378283cd5778b34778b7bf65fad2106f6fe9957fff5939faa`.
Inspection of the published run and retained-scan predicate identified the
unredacted `next_cursor_token` fields, including symbolic fixture placeholders.
The retained adapter now hashes/redacts cursor fields after in-memory checks,
and the evidence encoder preserves the literal `<redacted>` marker. Actual
payloads and transport bindings remain reviewable; capabilities are not retained.
No scanner exception or production change was introduced.

`nf-imp08-s33c-paging-validation-20261030-02` then passed native 0, including
finalization and 14 cleanup steps, at source
`sha256:cc36e1cf3d35670344d3f64d7825d01f1a1f2d14b7939ee06a56616631379df0`.
Both runs selected:
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0`.
That pass is historical: subsequent owner-clause review found missing required
error members in the fixture expectation as well as runtime. It does not close
014 or S33-C.

**F54 — Query admission loses required reason and detail fidelity.** NF-REQ-115a,
118, 119 and 175, Tables 21-A1/21-B and retry-action rules require meaningful
scope/mixed-query reasons and present required detail members, including nullable
unknown values. The current runtime omits `cursor_scope`; scope errors omit
`mode`, `table_ids`, `limit_key` and recovery advice, expose an unrelated generic
`field`, and collapse duplicate/over-limit selections into `empty_resolved_scope`.
Mixed initial/continuation requests use `semantic_query_mismatch` rather than the
registered specific mixed-variant reason.

014 revision 5 and 024 revision 4 now encode the owner requirements independently.
The new selectors were authored and registered before execution:

- `module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d`
- `module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367`

`nf-imp08-s33c-query-reproduction-20261030-01` ran:
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367`.
Native exit was 2; both comparison rows failed with product/test-assertion
classification; all 14 cleanup steps completed and retained scanning passed.
Source is `sha256:756acc360cff808c907d1fecc362e8ba3ce796ebcd070370df02f158cfea5bc4`;
HEAD/toolchain/system remain those recorded in §27.34. Exact root:
`.cartulary/test-results/nf-imp08-s33c-query-reproduction-20261030-01/`.
Its comparison summaries are:

- `NF-FIX-014-cursor-pagination-2882833233/comparison-summary.json`:
  exactly the missing `cursor_scope` member.
- `NF-FIX-024-query-normalization-cursors-462871193/comparison-summary.json`:
  scope details/classifications and mixed/cursor detail failures. Complete row,
  diagnostic, normalized-filter, keyset and expiry-success expectations otherwise
  agree. The first bounded difference review covered 32/36 paths; remaining
  paths stay part of the correction's complete-output exit.

S53 is now a prerequisite of S33-C. Its remediation belongs in implementation,
tests and tracker: use owner-specific closed scope admission with typed failure
context, preserve duplicate/limit classifications, and emit complete safe scope
and cursor details through the single semantic HTTP presenter. Reuse scope
resolution where needed to avoid divergent row/graph policy; do not weaken the
new comparisons. Long-term benefit is one explicit admission vocabulary and
consistent client recovery. Without correction, clients cannot reliably
classify scope/continuation failures and tests conceal the adopted detail gap.
Compatibility is correction of documented reasons and addition of required
nullable/context members; unsupported generic scope `field` is removed. No SQL,
state-major, durable identity or migration change is planned. V46 requires the
014/024 comparisons, affected query/admission/graph regressions, and the existing
query-authoring browser workflow. A runtime failure against an adopted owner
requires no additional adoption gate. S53's binary exit precedes resuming C.

All commands use the established pinned Go/installed-only environment. Row,
generate, format and task-guide commands used fresh
`nf-imp08-s33c-{paging,query}-*-20261030-*` IDs and exited 0. Exact-root observers
`nf-imp08-s33c-paging-observe-20261030-{01,02}`,
`nf-imp08-s33c-paging-logs-20261030-01`,
`nf-imp08-s33c-query-observe-20261030-01` and
`nf-imp08-s33c-query-logs-20261030-01` exited 0. All roots retain their native
result separately from target summaries. 013/027 have not yet been rebuilt;
S33-D–F, F36/F37 and S36 remain open.

### 27.36 S53 implementation and verification

S53 replaces divergent row/graph scope admission with a private two-method read
interface and shared resolver. Selection preserves workspace order, rejects
explicit duplicates, retains missing selected IDs for a proper lookup failure,
and bounds all-active scopes as well as explicit selections. Typed semantic
failure context carries the admitted mode, selected IDs and applicable limit.
The HTTP presenter supplies the required nullable cursor/scope members and
recovery action, and mixed initial/continuation bodies receive their specific
registered reason. The last four previously unreviewed comparison differences
were the same cursor-detail/mixed-variant omissions; no expectation was weakened.

The first correction execution, `nf-imp08-s53-validation-20261030-01`, exited 2
with infrastructure/preflight compilation failure: the new resolver initially
required the concrete store while Graph uses its private read interface, and
an obsolete `sort` import remained. All 11 cleanup steps completed. Source was
`sha256:3a0e218132643651c3c77c9ef723f11eb790c8732ec2cdbb46914f785ef80258`.
The narrow `tableScopeReader` interface and explicit effective limits repair
that coupling without broadening Graph's storage access.

`nf-imp08-s53-validation-20261030-02` compiled, passed both full fixture
comparisons, and exited 2 with product assertion failures in two old unit
expectations that classified duplicate IDs as an empty scope. All 15 cleanup
steps completed. Source was
`sha256:48819664d0eb94cfd05267a8e20eaea875bc887332282b081176fc85c61b30d4`.
Both duplicate-specific expectations now require `duplicate_table_id`.

`nf-imp08-s53-validation-20261030-03` passed native 0 with all 15 cleanup steps
complete, source
`sha256:6affdcacb34a293e6a317a1e502fe1fec15e04f25bb8f6349828ff8d92770a2a`.
All three executions used the exact command:
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_selector_covers_a_table_query_uses_6caf43b23c,module.networkflow.unit.network_flow_selector_covers_every_network_flow_43bd8f4992,module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58`.
Each root is `.cartulary/test-results/<run-id>/`, with native result, run manifest,
`target-summaries/test-slice.json` and `cleanup-results.json`. HEAD, toolchain and
system identities remain those in §27.34. Formatting runs
`nf-imp08-s53-format-20261030-{01,02,03}` and exact-root observers
`nf-imp08-s53-observe-20261030-{01,02,03}` exited 0. Published failure-log
observers also exited 0; their success is diagnostic evidence only.

The existing browser row is selected separately as the remaining V46 obligation:
`nf-imp08-s53-browser-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.query_authoring_integrity`.
Its terminal result and closing Markdown checkpoint are recorded below before
S33-C resumes. These are affected-scope passes, not final-source qualification.

The first browser run exited native 2, product/test-assertion failure, with all
23 cleanup steps complete. Its source matched the passing focused run. Published
Playwright evidence identifies `apps/web/e2e/network-flow.spec.ts:333`: expected
8 contributing rows, received 6. Source review confirms query-a accepts five
rows and the rebuilt NF-FIX-001 source accepts one; the obsolete expectation
assumed the former three-row minimal input. The browser now asserts both source
accepted-count contributions `[1,5]` and exactly six edge contributors. This is
fixture-consumer maintenance, not acceptance of a runtime discrepancy.
`nf-imp08-s53-format-20261030-04 make format` passed. The same browser selector
runs under fresh `nf-imp08-s53-browser-20261030-02`; no wider suite was launched.

Passing comparison evidence from focused execution 03 is retained at:

- `NF-FIX-014-cursor-pagination-2333893895/comparison-summary.json`, manifest
  `0eabf0e75af97983bf82b18dec17bd1b381aedf54c6f80f0f01be758984b5134`.
- `NF-FIX-024-query-normalization-cursors-3345858373/comparison-summary.json`,
  manifest `b58b8d5d9dd26222149f309cf4641d4878df7ce6936c6b8d72075cc9c9fbda10`.

The browser's published report is
`browser-e2e-webserver-backed/browser-groups/cohort-230b9903c97ed05f15391d2ab1c0c7beb49e9d37a93cbdc178d32ec9afc663ea/playwright-report.json`
under its exact run root. Summary/log observers
`nf-imp08-s53-browser-{observe,logs}-20261030-01` passed native 0.

Browser execution 02 passed native 0 with all 23 cleanup steps complete at
`sha256:ca5be7e7ea5f3d8c1ea02a8425a74294ff3cfa6b9b753396ed946be0c148d259`.
The source difference from the focused backend pass is the documented browser
expectation repair; S36 will apply the harness's compatibility admission rather
than asserting cross-source qualification. S53/F54 and V46's correction exit are
satisfied. No schema, durable-state version or migration changes resulted.
Rollback is the source/test revert; the risk of incomplete scope/cursor error
classification is resolved for the reproduced paths. Unrelated error families
remain subject to their upcoming fixture workstreams.

S53 is DONE, pending its closing Markdown check below. S33-C remains IN_PROGRESS;
013 and 027 resume only after that checkpoint. S33-D–F, F36/F37 and S36 remain
open; neither this pass nor any historical pass is final-source qualification.

S53 closing check `nf-imp08-s53-close-20261030-01 make lint-markdown` passed native 0. S33-C resumes with 013 and 027.


### 27.37 S33-C naming reproduction and S54 correction entry

013 now contains twelve independently isolated cases, including a real
four-attempt import sequence: default name, deterministic suffix, rejected
explicit duplicate, and reuse after deleting the original. It compares full
responses, source/mapping/row digests, provenance, publication, rollback, replay,
retained state and audit. Additional cases cover Windows/Unix path stripping,
NFC/control removal, empty filenames, hidden/trailing-dot stems, 64-scalar default
bounds, normalized explicit names, and empty/control/overlong override rejection.
The revision is 4: revision 3 was advanced after correcting the independently
reviewed Core fallback error message. No runtime output was copied into an oracle.

The selector was authored with
`make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Naming fixture compares imports filename normalization explicit rejection duplicate allocation and deleted name reuse' SELECTOR_KEY=TestNetworkFlowNamesFixtureComparison`
under `nf-imp08-s33c-names-row-20261030-02`, then registered through the authored
family and `make generate`. The actual ID is
`module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d`.
An initial row-authoring invocation ending `-01` rejected the unsupported
`FAMILY` input, native 2/configuration error; it authored nothing. Generate,
format 01/02 and task-guide runs under `nf-imp08-s33c-names-*-20261030-*` passed.

Both reproduction executions used
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d`.
`nf-imp08-s33c-names-reproduction-20261030-01` exited 2/product assertion at source
`sha256:4ebb2bc0085af37a240701a16ec01f6ab138d18ddec84a029cd659a3f14e1f99`,
with all 14 cleanup steps complete: the fixture tried to re-arm an existing
randomness stream. The adapter now uses the existing supported registry Clear
operation between completed imports before arming the next input; no production
determinism hook was introduced.

Execution `nf-imp08-s33c-names-reproduction-20261030-02` exited 2/product assertion
at `sha256:776d1a7d3e7965ab29206747042f74c753baf98504f37d0024d2be0c31ec853d`,
with all 14 cleanup steps complete. All scenario operations ran. Full comparison
has exactly 15 differences, all reviewed: four appearances of the duplicate
failure use the wrong owner code and omit both length members; three preview
errors lose owner detail entirely. All other exact expected responses and
persistence effects agree. Retained comparison summary:
`.cartulary/test-results/nf-imp08-s33c-names-reproduction-20261030-02/NF-FIX-013-filename-display-856834157/comparison-summary.json`.
Exact-root summary/log observers 01/02 passed native 0. Native results, manifests,
target summaries and cleanup records remain in each exact run root. HEAD,
toolchain and system identities remain those in §27.34.

**F55 — Import display-name failures lose the adopted identity and safe detail.**
NF-REQ-065b explicitly requires `network_flow_invalid_display_name` for an
explicit duplicate; NF-REQ-066 defines empty/control/length rejection. The
NF-REQ-088d closed import owner-error registry omits this required case, and the
implementation maps duplicate apply failures to generic mapping-invalid while
preview loses the typed error. S54 reconciles the registry with the specific
naming requirement, projects the closed safe error shape, and uses one private
name-error conversion for both preview and apply. The shape contains only
`reason_code`, `max_length=64` and nonnegative `normalized_length`; no supplied
name, filename, path or storage identity is exposed. Core retains its existing
preview/apply envelope and transaction authority.

This is a specification, implementation, schema, test and documentation
correction. It preserves useful field-specific recovery and avoids a second
name-policy path. Leaving it unresolved conceals the cause of rejected names
and contradicts the adopted import behavior. Compatibility: the owner-error
union gains the already-required code, and clients must consume the matching
contract projection. Major 8/state 5 and SQL remain unchanged. The import-registry
hash changes; migration 34's existing requirement to drain/cancel every
`import.apply_v1` job remains mandatory. Completed outcomes retain historical
error bytes; no receipt rewrite or legacy translation layer is added. Record the
new registry hash after generation and validate migration preservation/drain
coverage. Rollback remains the matching source/schema revert plus S39's
pre-migration backup procedure for an applied migration 34.

S54 is now a prerequisite of S33-C, entered before correction edits. V47 requires
013, closed owner-error admission and negatives, preview/table-name regressions,
canonical migration checks, affected schema/generated/catalog checks and the
existing import-recovery browser workflow. Its binary exit is complete matching
naming comparisons and safe closed translation with those obligations passing.
S33-C remains IN_PROGRESS; 027 and S33-D–F/F36/F37/S36 remain outstanding.

S54 entry check `nf-imp08-s54-entry-20261030-01 make lint-markdown` passed native 0.
The owner registry now explicitly includes the name-error case required by
NF-REQ-065b/066, and the authored schema projects its closed safe shape. The
private conversion is shared by `facadeError` and `storeApplyError`; admission
rejects unknown reasons, omitted/extra members and invalid numeric bounds.
Existing registered owner errors remain accepted. Unit controls exercise both
serving conversion paths and deliberately unsafe detail objects.

`nf-imp08-s54-generate-20261030-01 make generate` and
`nf-imp08-s54-format-20261030-01 make format` passed native 0. The projected
Imports registry hash changes from
`660bd3742382a8e67bdbbcc12b66fede41736b11149a77004dfd451b27536ba9` to
`e546bb797cc033d43d24b3eaa8579e1405283ced322c4290043d1071aea54b0b`.
Migration 34 already fences all unfinished Core apply jobs, so no further SQL or
historical migration edit is needed. Its existing preservation and drain
negative controls are selected again for this contract change.

The V47 correction portfolio now running uses these exact commands and fresh IDs:

| Run ID | Command |
| --- | --- |
| `nf-imp08-s54-validation-20261030-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d,module.networkflow.unit.import_owner_error_translation_rs08_3d3f02e050,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.table_lifecycle_contract,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca` |
| `nf-imp08-s54-browser-20261030-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.import_recovery` |
| `nf-imp08-s54-schema-20261030-01` | `make source-check CHECK=schema` |
| `nf-imp08-s54-generated-20261030-01` | `make source-check CHECK=generated` |
| `nf-imp08-s54-catalog-20261030-01` | `make source-check CHECK=catalog` |

All use the pinned Go/installed-only environment already recorded. Results are
not claimed until native completion and retained finalization are observed.


### 27.38 S54 validation and correction exit

Every V47 command listed in §27.37 passed native 0, including retained
finalization. All five share source
`sha256:cef6e1328f0b7d52ab2fdfde9aa21c0bc432eef289ef0ed473353783dd810f06`,
HEAD `98adcc82961456bcb74bbecd0023136485afbf00`, and the established toolchain and
system digests. Cleanup completed without failures or blocked steps: focused
validation 18, browser 23, schema 12, generated 14, catalog 11. Exact roots are
`.cartulary/test-results/<run-id>/`; each retains its manifest, native result,
`target-summaries/{test-slice,source-check}.json` as appropriate and
`cleanup-results.json`. Corresponding exact-root `*-observe-20261030-01`
commands passed native 0; these are diagnostic operations, not separate tests.

013 revision 4 now passes every complete comparison and deliberate wrong-output
control. Its summary is
`NF-FIX-013-filename-display-4034759134/comparison-summary.json` beneath
`nf-imp08-s54-validation-20261030-01`; manifest hash is
`63a7afe2d29cd8c99d732d7084f4860ba0ca321b93171c71302c6ea822c8adf6`.
The migration row passes preservation, malformed-context rollback, Network Flow
in-flight work rejection and non-Network-Flow Core apply-job drain coverage.
Import-recovery browser behavior passes with the new registry.

S54/F55 are resolved and V47's correction exit is satisfied. The generated owner
projection and runtime use the same safe name-error vocabulary; no stored outcome
was rewritten, no historical migration changed, and migration 34's all-apply-job
drain requirement remains explicit. Remaining risk is deployment without the
matching schema/runtime or required drain; the handoff must retain both conditions.
S54 is DONE, pending the closing Markdown checkpoint below. S33-C resumes with
027 after it passes; C's aggregate closure and S33-D–F/F36/F37/S36 remain open.

S54 closing check `nf-imp08-s54-close-20261030-01 make lint-markdown` passed native 0. S33-C resumes with fixture 027.


### 27.39 S33-C retention reproduction and S55 correction entry

027 revision 4 replaces the placeholder with two isolated production workflows:
soft delete and actual Core incident closure. Each imports two accepted rows and
one rejected record with two diagnostics, creates a real Network Flow binding to
an owner-seeded existing Indicator, and obtains independent real row/diagnostic
cursors. Viewer rename/delete attempts, successful mutations, exact replay,
post-transition reads/linking/rename, retained counts and full owner state are
compared. Core closure audit is compared as well as Network Flow audit. Positive
binding retention prevents a vacuous zero-binding assertion. This establishes
retention at the adopted boundaries without claiming a private purge feature.

Canonical row identities, digests, source refs and business timestamps are exact
independent vectors. The existing controls expose table UUIDs but no binding-ID
entropy override. Random binding IDs are retained verbatim, checked for their
fixed format and exact equality across response/storage/audit/replay, then aliased
only in frozen semantic comparisons. Core closure timestamps are retained,
checked against the operation interval and each other, and likewise aliased.
No production determinism hook or global crypto substitution was introduced.
The generic error observer validates/retains the complete actual envelope;
expected status, code and every detail remain owned by the frozen comparison.

The selector was authored before execution:
`make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Retention fixture compares soft delete incident closure positive binding retention authorization cursors and audit replay' SELECTOR_KEY=TestNetworkFlowRetentionFixtureComparison`,
run `nf-imp08-s33c-retention-row-20261030-01`, native 0. The actual row is
`module.networkflow.integration.retention_fixture_compares_soft_delete_incident_570ae9035a`.
Authored routing includes Indicators and Incidents collaborators. Generate,
format and task-guide runs `nf-imp08-s33c-retention-{generate,format,guide}-20261030-01`
all passed native 0. Revision 3 advanced to 4 before execution to correct an
independent HMAC literal's byte framing; no product result was used as an oracle.

`nf-imp08-s33c-retention-reproduction-20261030-01` executed
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.retention_fixture_compares_soft_delete_incident_570ae9035a`.
Native exit 2, product/test-assertion failure, all 14 cleanup steps complete;
source `sha256:80e224110a32abea8898f0b5a570a1d77d0b94654fb2b23fee7f058c5326188b`.
HEAD/toolchain/system remain those in §27.34. Both scenario operations completed.
Exactly 35 difference paths were reviewed: all concern deleted-table error
presentation and cursor invalidation. Every import, persisted row/diagnostic,
positive binding, audit, count, authorization rejection and Core closure
expectation otherwise agrees. Retained summary:
`.cartulary/test-results/nf-imp08-s33c-retention-reproduction-20261030-01/NF-FIX-027-retention-soft-delete-2188818921/comparison-summary.json`.
The exact root contains native result, manifest, target summary and cleanup;
summary/log observers ending `-observe-20261030-01` / `-logs-20261030-01` passed 0.

**F56 — Deleted-table failures omit owner context and bypass cursor policy.**
NF-REQ-175/Table 21-B requires table ID, status, allowed states and recovery advice
for disclosed table lookup/state failures. Current GET, row query, diagnostic
query and rename paths omit those facts and expose an unsupported generic
`field`; linking adds required facts but still carries that generic field.
NF-REQ-125/Table 13-G explicitly requires `network_flow_cursor_invalid` after a
cursor-scoped table is soft-deleted, but both accepted-row and diagnostic
continuations return `network_flow_table_not_active` instead.

S55 is entered before runtime edits and is now a prerequisite of S33-C.
Remediation: bind known table identity at the private error boundary, present
lookup/state facts through one policy, and translate typed scope failures on
continuation to safe `scope_stale` cursor failures. Preserve Core route admission
and hidden-resource precedence; do not expose sealed payloads or weaken failed
reads. This belongs in implementation, tests and documentation. It gives callers
stable recovery semantics and keeps read/mutation/link paths consistent. Without
it, clients cannot distinguish a restartable continuation from a fresh table
lookup and required errors remain incomplete.

Compatibility is correction to adopted error statuses/codes and required details,
with removal of the generic lookup `field`. No SQL, owner registry, major/state
version or migration changes are planned. Rollback is source/test reversion;
existing migration-34 operational prerequisites remain unchanged. V48 requires
all four C comparisons, focused pagination/lifecycle/graph-link regressions, and
the existing table-lifecycle browser workflow. Exit is exact passing deleted-table
and cursor errors with unchanged retained data, authorization and replay effects.
S33-C remains IN_PROGRESS; S33-D–F, F36/F37 and S36 remain open.

S55 entry documentation check `nf-imp08-s55-entry-20261030-01 make lint-markdown`
passed native 0. The correction binds validated table IDs when constructing typed
lookup failures; malformed supplied IDs produce a required null identity rather
than echoing untrusted text. The semantic HTTP presenter now owns complete lookup
and state details for reads, table mutations and links. Redundant legacy HTTP and
link-specific lookup presenters were removed. Verified row/diagnostic
continuations translate only typed missing/deleted scope failures to
`network_flow_cursor_invalid/scope_stale`; infrastructure failures retain their
original classification. Existing Core authentication/incident admission stays
before owner request processing. Focused negative controls ensure malformed IDs
do not escape and internal errors are not mislabeled as stale cursors.

`nf-imp08-s55-format-20261030-01 make format` passed native 0. The final obsolete
link-helper reference was removed during static review before execution; there
was no failed compilation or speculative rerun. Current V48 commands are:

- `nf-imp08-s55-validation-20261030-01`:
  `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367,module.networkflow.integration.retention_fixture_compares_soft_delete_incident_570ae9035a,module.networkflow.unit.query_authoring_admission,module.networkflow.integration.pagination_recovery,module.networkflow.integration.table_lifecycle_admission_replay,module.networkflow.integration.bounded_graph_contributor_pipeline`.
- `nf-imp08-s55-browser-20261030-01`:
  `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.table_lifecycle_continuity`.

Both use the established pinned Go/installed-only environment and exact roots
under `.cartulary/test-results/<run-id>/`. Terminal results follow before closure.

S55 focused execution 01 exited native 2/product assertion at source
`sha256:5d5970a8c4015da8ade77b3bf46eb37eba1a43be5b07b5911615cda9ce52d544`;
all 15 cleanup steps completed. All four C comparisons and the query, pagination
and graph/link regressions passed. The sole failing leaf is the old functional
authorization control in `harness_authorization_integration_test.go:116`, which
expects HTTP 409/table-not-active after deleting a cursor's table. The actual
400/cursor-invalid response matches NF-REQ-125 and the new 027 comparison. Its
parent failures are the same assertion, not additional defects. The legacy
expectation will be corrected before the focused rerun; no runtime rollback or
fixture-oracle weakening is warranted. Exact-root summary/log observers
`nf-imp08-s55-{observe,logs}-20261030-01` exited 0.

### 27.40 S55 correction exit

F56 is repaired. Private table lookup failures now carry only validated table
identity; the shared HTTP presenter supplies required status, allowed states
and refresh action. Reads, mutations and linking use this same presentation.
Both row and diagnostic continuations convert unavailable scoped tables into
`network_flow_cursor_invalid` / `scope_stale`. Initial requests retain the
appropriate lookup error; internal/cancellation failures are not reclassified.
Core authorization and closed-incident precedence remain intact. The existing
query-authoring row tests these distinctions and suppresses malformed IDs.
The obsolete authorization-control assertion now expects the adopted cursor
response. No schema, registry hash, migration or public version changed.

Focused rerun `nf-imp08-s55-validation-20261030-02` used exactly
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.table_lifecycle_admission_replay`:
native 0, all 14 cleanup steps successful, source
`sha256:673bb81faafdcb35e52e8e5d47eb39aa8c156b137387877601ce599927ee0962`.
Its summary is
`.cartulary/test-results/nf-imp08-s55-validation-20261030-02/target-summaries/test-slice.json`.
Only the failed row was rerun after its test-only expectation repair; the seven
other selected rows had already passed in execution 01. That failed aggregate
run is not represented as passing. Exact-root summary observer
`nf-imp08-s55-observe-20261030-02` exited 0.

`nf-imp08-s55-browser-20261030-01` executed the previously recorded
`module.networkflow.browser.table_lifecycle_continuity` selection: native 0,
23 successful cleanup steps, source matching execution 01 (`5d5970a8…`).
Summary:
`.cartulary/test-results/nf-imp08-s55-browser-20261030-01/target-summaries/test-slice.json`.
Observer `nf-imp08-s55-browser-observe-20261030-01` exited 0. Environment identities
are unchanged from §27.1. These workstream observations do not assert final
source compatibility; S36 must qualify the final frozen source.

All required S55 behaviors have passing observations; no dependent correction
remains open. Compatibility deliberately changes stale-cursor and incomplete
lookup responses to the adopted contract. Clients must refresh cursor state.
Rollback is source reversion only, subject to migration 34's already recorded
backup/binary constraint. S55 closure awaits its documentation checkpoint;
S33-C may then close, and S33-D remains unstarted.

S55 closing `make lint-markdown`, run `nf-imp08-s55-close-20261030-01`,
exited native 0; retained summary:
`.cartulary/test-results/nf-imp08-s55-close-20261030-01/adhoc/lint-markdown/tool-run-summary.json`.
S55 is DONE and F56 resolved before S33-C resumes.

### 27.41 S33-C query and lifecycle exit

All four independently admitted C bundles now compare actual Imports publication,
Network Flow HTTP resources, cursor payloads and durable effects. Fixture 013
covers filename/default-name normalization, explicit names, collisions, deletion
and reuse through real imports. Fixture 014 compares exact keyset pages, sort
completion, actor binding and rename continuity. Fixture 024 covers closed scope
admission, canonical filter identity, incompatible request variants, token size,
expiry and distinct diagnostic continuation tuples. Fixture 027 compares positive
linking, forbidden mutations, soft deletion, Core incident closure, replay,
retained rows/diagnostics/bindings/audit and both invalidated continuations.

Shared test setup admits only import prerequisites and supported Core indicator
records; no fabricated Network Flow rows replace product publication. Generated
binding IDs and real Core closure times are independently validated and
cross-correlated before semantic aliasing; raw resources remain retained.
Business time and table identity use existing supported controls. The transport
retention boundary redacts opaque cursor values while the adapters execute their
original values and compare verified plaintext payloads. Errors are complete
response comparisons, including the S53–S55 corrections; expectations were not
changed to preserve the defects.

Passing comparison summaries in
`.cartulary/test-results/nf-imp08-s55-validation-20261030-01/` are:

| Bundle | Revision | Summary directory (append `/comparison-summary.json`) | Manifest SHA-256 |
| --- | --- | --- | --- |
| 013 | 4 | `NF-FIX-013-filename-display-3743095276` | `63a7afe2d29cd8c99d732d7084f4860ba0ca321b93171c71302c6ea822c8adf6` |
| 014 | 5 | `NF-FIX-014-cursor-pagination-2702695518` | `0eabf0e75af97983bf82b18dec17bd1b381aedf54c6f80f0f01be758984b5134` |
| 024 | 4 | `NF-FIX-024-query-normalization-cursors-2288931224` | `b58b8d5d9dd26222149f309cf4641d4878df7ce6936c6b8d72075cc9c9fbda10` |
| 027 | 4 | `NF-FIX-027-retention-soft-delete-3593359091` | `c6665c6e121fd898fa8efe2050dfc3a292f6821ac9db1f9bf74487df8c74d698` |

The four registered row IDs and exact commands are recorded in §27.39; all four
row-result artifacts are `passed`. The aggregate execution's unrelated obsolete
lifecycle assertion, its narrow passing rerun, browser result, source/environment
identities and cleanup are recorded in §27.40. These are retained workstream
observations, not a claim that a failed aggregate passed or that final S36
qualification already exists. Existing earlier failed reproductions remain
historical evidence.

No additional C correction is open. Revisions supersede prior frozen bytes
through Git history without a runtime compatibility adapter. Source/test
reversion is the fixture rollback; deployed identity changes retain migration
34's independent backup/binary rule. Remaining risk is regression during later
graph/link and normalization work, addressed by final affected verification.
Seventeen of 28 families now have substantive passing comparisons. S33-C's
binary behavior exit is satisfied; its closing Markdown check is the remaining
checkpoint before S33-D starts. S33 aggregate remains IN_PROGRESS.

S33-C closing `make lint-markdown`, run `nf-imp08-s33c-close-20261030-01`,
exited native 0; summary:
`.cartulary/test-results/nf-imp08-s33c-close-20261030-01/adhoc/lint-markdown/tool-run-summary.json`.
S33-C is DONE before the next workstream begins.

### 27.42 S33-D graph and resource-bound entry

S33-D now owns 006/008/009/015/018/025/028. Its prerequisites S33-C and all
required correction exits are complete. Typed adapters will publish valid
source records through Imports, compare v2 projection inputs/results and graph
identities, exercise contributor/query/state/authorization boundaries, and
prove resource-limit failure ordering and absence of partial output. Frozen
expectations must derive from owner rules, not observed snapshots. Runtime
violations will receive separate reproductions/correction entries before any
runtime edit; S33-D acceptance will depend on their passing exits. F39's registry
reference correction is already adopted. F38/F40 remain open until corpus
closure. No D behavior is yet qualified.

S33-D first adapters are authored for 006/015, both revision 3. A test-only
private-boundary observer wraps the real SQL source iterator and delegates the
actual trusted invocation to Graph Projection v2. The public route result must
also equal that private composition result. Independent expected bytes include
all source properties, definitions, metadata, identifiers, examples and semantic
audit effects; unsaved declarations/results must remain absent. Optional fixture
resource limits use the existing deployment configuration interface, not a new
production test hook. Existing publication fixtures retain their defaults.

New rows were authored with `make author-test-row-id` using these exact claims
and selectors (runs `nf-imp08-s33d-author-{cross,adapter}-20261030-01`, native 0):

| Actual registered row | Claim | Selector |
| --- | --- | --- |
| `module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9` | Cross table graph fixture compares exact imported aggregation and Graph Projection input output identities | `TestNetworkFlowCrossTableGraphFixtureComparison` |
| `module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2` | Graph adapter fixture compares v2 input trusted context bounded examples nullable properties and exact pure result | `TestNetworkFlowGraphAdapterFixtureComparison` |

Both use `FAMILY_ID=module.networkflow.integration`. Authored family registration
was regenerated with `nf-imp08-s33d-generate-20261030-01 make generate` (native 0).
`nf-imp08-s33d-format-20261030-01 make format` and
`nf-imp08-s33d-guide-20261030-01 make task-guide OWNER=module.networkflow` also
exited 0. Planned first comparison execution (not yet a pass):

```sh
CARTULARY_TEST_RUN_ID=nf-imp08-s33d-graph-reproduction-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2
```

All commands retain the explicit Go/preparation environment from §27.1. S33-D
remains IN_PROGRESS; no successor workstream is authorized by these first two
adapters alone.

First graph execution exited native 2/product assertion at source
`sha256:b42c1f6f065dd3393152ce4d2f3116942850643cc5171bdd951b3230c1604a77`;
all 14 cleanup steps completed. Both adapters stopped in the new test setup:
its resource-limit resolver supplied a relative dummy key-ring manifest path,
which the production configuration validator correctly rejects. The helper now
uses the same existing absolute fixture key-ring path as server assembly. This
is a fixture setup repair, not a product defect or an expected-value change.
The exact-root summary/log observers `nf-imp08-s33d-{observe,logs}-20261030-01`
exited 0. The root retains both failed admission/comparison summaries. The
next focused run also selects the existing minimal publication row to verify
unchanged defaults in the shared setup.

Setup repair format `nf-imp08-s33d-format-20261030-02` exited native 0.
Focused rerun command:

```sh
CARTULARY_TEST_RUN_ID=nf-imp08-s33d-graph-reproduction-20261030-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0
```

Graph execution 02 exited native 2/product comparison at source
`sha256:5ef0f258b95b19bc568bc5129b2d3dc2e606c9dd1292f0a7a2be0e4bb3168268`,
14 successful cleanup steps. Fixture 006 and the unchanged minimal import row
passed. Fixture 015 produced complete evidence but exposed an oracle mistake:
§9/Table 9-A and the explicit claimable-profile rule require non-null Cisco
source/destination ports. Nullable graph property definitions bind future
profiles; they do not authorize accepting empty Cisco ports. Both production
preview and apply correctly reject the final row with two field-specific
`missing_or_empty` diagnostics. All 115 differences were inspected and derive
from that incorrectly expected accepted row and its graph aggregate.

015 advances to revision 4. Its independent reference now applies the current
source-profile restriction, compares both exact diagnostics and excludes the
rejected row from graph composition. The same six source records remain, and
nullable property-definition policy is still compared exactly. This repairs
an expectation contradicting the owner; no runtime behavior is weakened or
changed. Failed summary:
`.cartulary/test-results/nf-imp08-s33d-graph-reproduction-20261030-02/NF-FIX-015-graph-adapter-input-3919477708/comparison-summary.json`.
Exact-root observers `nf-imp08-s33d-{observe,logs}-20261030-02` exited 0.
Next command selects only the failed adapter row:

```sh
CARTULARY_TEST_RUN_ID=nf-imp08-s33d-graph-reproduction-20261030-03 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2
```

015 revision 4 passed in execution 03: native 0, 14 successful cleanup steps,
source `sha256:cfbe4951555a2b477d7905665950f7e947319f8009d089af32b9f95461240ba8`.
Passing summary:
`.cartulary/test-results/nf-imp08-s33d-graph-reproduction-20261030-03/NF-FIX-015-graph-adapter-input-917159516/comparison-summary.json`.
006 revision 3's passing summary is
`.cartulary/test-results/nf-imp08-s33d-graph-reproduction-20261030-02/NF-FIX-006-cross-table-graph-2714910451/comparison-summary.json`.
Observer `nf-imp08-s33d-observe-20261030-03` exited 0. F39 now has substantive v2
adapter/input/output evidence as well as its specification correction; final
qualification remains S36. Five D families remain uncompleted.

009 revision 3 now has an independent two-table lifecycle comparison. It
captures exact graph input/output and row/contributor cursor payloads, deletes
one table through the route, checks replay and retained state, rejects stale
requests, then recomputes over the remaining table. Expectations distinguish
initial stale graph selection (NF-REQ-141c/191, Table 21-B `scope_stale`) from
cursor invalidation (NF-REQ-125/Table 13-G); an explicit new graph request naming
a deleted table retains the ordinary authorized table lookup error.

`nf-imp08-s33d-author-stale-20261030-01 make author-test-row-id` used
`FAMILY_ID=module.networkflow.integration`,
`CLAIM='Stale graph fixture compares soft deletion contributor and row cursor invalidation retained sources and recomputation'`
and `SELECTOR_KEY=TestNetworkFlowStaleGraphFixtureComparison`, native 0.
Actual row:
`module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49`.
Authored registration was projected by `nf-imp08-s33d-generate-20261030-02`;
format 03 and task-guide 02 also exited 0. Exact reproduction command:

```sh
CARTULARY_TEST_RUN_ID=nf-imp08-s33d-stale-reproduction-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49
```

### 27.43 S33-D stale graph reproduction and S56 correction entry

009 reproduction exited native 2/product comparison at source
`sha256:4945687250b8b341d4782fa7f879754ee57966fc65a6b729a94f7864a1744195`,
with all 14 cleanup steps successful. Summary:
`.cartulary/test-results/nf-imp08-s33d-stale-reproduction-20261030-01/NF-FIX-009-soft-delete-stale-graph-1242238029/comparison-summary.json`.
Both exact-root observers `nf-imp08-s33d-stale-{observe,logs}-20261030-01`
exited 0. All 16 differing values were inspected: they are confined to the
two contributor failures after deletion. Imported state, both cursor payloads,
initial graph, deletion/replay, row cursor invalidation, remaining-table graph,
audit and absence of graph publication all matched the independent expectation.

**F57 — Contributor scope failures bypass graph/cursor lifecycle policy.**
The contributor application forwards private table lookup failures unchanged.
An initial selector against deleted graph scope must instead report
`network_flow_graph_query_stale` / `scope_stale` under NF-REQ-141c/191 and
Tables 21-A/21-B/21-C. A contributor continuation must report
`network_flow_cursor_invalid` / `scope_stale` under NF-REQ-125/Table 13-G.
The source table's identity and lifecycle details are not members of those
closed recovery errors. Current behavior obscures the required client recovery
and differs from the corrected row/diagnostic continuation policy.

S56 is entered before runtime edits, with the failing 009 comparison as its
prerequisite; it does not depend on completion of S33-D. Remedy: put the
context-specific conversion at the contributor serving/application boundary,
sharing existing typed failure construction and preserving every unrelated
failure, authorization precedence and durable effect. No Graph Projection
algorithm, public schema ID, registry hash, durable-state version or migration
changes. Compatibility intentionally corrects these error variants. Rollback is
source/test reversion within the previously recorded migration-34 constraints.
The main risk is accidentally disguising cancellation/internal failures as
staleness; focused negative assertions will forbid that.

**V49** adds the 009 comparison plus existing
`module.networkflow.unit.query_authoring_admission` and
`module.networkflow.integration.bounded_graph_contributor_pipeline` rows.
The affected browser workflows are the registered selectors
`module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519`
and
`module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_v_171429ad9e`.
These cover the current ordered edge drawer and stable vertex selector workflows.
S56 exits only after exact 009 behavior, typed failure distinctions, existing
contributor regressions, affected browser workflows, cleanup and its closing
Markdown check pass. S33-D then resumes; 009 expectations remain unchanged.

S56 entry Markdown run `nf-imp08-s56-entry-20261030-01` exited native 0 before
runtime edits. The contributor application now converts typed missing/inactive
table failures using graph-selection or live-cursor context. Existing typed
cursor construction is reused; source composition and other consumers remain
unchanged. The registered query-authoring row now compares both exact closed
error objects and proves nil, internal cancellation, projection cancellation,
actor mismatch and digest mismatch retain their original classifications.

S56 `make format` and `make task-guide OWNER=module.networkflow`, runs
`nf-imp08-s56-{format,guide}-20261030-01`, exited native 0. V49 commands on the
resulting source (explicit toolchain/preparation environment unchanged):

```sh
CARTULARY_TEST_RUN_ID=nf-imp08-s56-validation-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49,module.networkflow.unit.query_authoring_admission,module.networkflow.integration.bounded_graph_contributor_pipeline
CARTULARY_TEST_RUN_ID=nf-imp08-s56-browser-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_v_171429ad9e
```

S56 validation 01 exited native 2/product assertion at source
`sha256:3420ecadfe8cb7638d7f37993ea307ee20642279de044d6359fb31504020e981`,
with all 15 cleanup steps successful. The unchanged 009 revision-3 comparison
and the query-authoring unit row passed. The sole failed row is the existing
bounded contributor pipeline: `routes_integration_test.go:1191` still requires
`network_flow_table_not_active` / `soft_deleted` for an initial stale graph
selector. Actual `network_flow_graph_query_stale` matches the adopted contract
and the exact fixture. The legacy assertion will be corrected, then only that
failed row rerun. Summary/log observers `nf-imp08-s56-{observe,logs}-20261030-01`
exited 0. The aggregate failure is retained, not reported as a pass.

Review of browser coverage identifies one further directly affected existing
workflow: `module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf`
(`Verify Network Analysis clears protected grid, inspector, graph, contributor, and selection state after lifecycle loss and refetches after recovery.`).
V49 adds it because it exercises deletion, graph-stale recovery and protected
state clearing beyond the two ordinary selector workflows. No new selector or
browser behavior is authored for this addition.

The two selector browser rows completed native 0 in
`nf-imp08-s56-browser-20261030-01` before the legacy Go expectation edit.
Follow-up commands on the resulting test-only source change:

```sh
CARTULARY_TEST_RUN_ID=nf-imp08-s56-validation-20261030-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline
CARTULARY_TEST_RUN_ID=nf-imp08-s56-browser-20261030-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf
```

The first S56 browser root has 23 successful cleanup steps and source identity
matching validation 01 (`3420ecad…`); exact-root browser observer 01 exited 0.
The passing 009 summary in that backend root is
`NF-FIX-009-soft-delete-stale-graph-3510209282/comparison-summary.json`, manifest
SHA-256 `4ec59085d9817392cf69bb9f7ce53896416605eccc33b170bd123bd3c63024c0`.
Ordinary Git whitespace inspection reports only the deliberate trailing spaces
in fixture 002's raw interface cells (rows 2/4). Those frozen bytes exercise
ASCII-space trimming and must remain intact; this is not a whitespace-check
pass or a reason to rewrite the source corpus.

### 27.44 S56 correction exit

F57 is repaired in implementation, tests and handoff documentation; the adopted
owner requirements needed no amendment. The contributor serving boundary now
owns graph-selection versus cursor recovery, reusing typed lifecycle failures.
Closed error details no longer inherit irrelevant table context. This keeps
storage lookup semantics useful to other consumers while making contributor
recovery consistent and independently testable. No persisted identity, source
snapshot, Graph Projection output, import contract, migration or registry hash
changed. Clients receive the already-adopted graph refresh or query restart
signal. Leaving the former behavior would preserve misleading recovery and
inconsistent cursor lifecycle; the exact fixture now guards both branches.

Validation 02's narrow failed-row rerun and browser 02's protected-state recovery
workflow both exited native 0 at source
`sha256:1e44f7c0604a8913ecaa7a93ed0add8dbbd4e0729085a6bba6c78ba57c3f7b29`.
Cleanup completed all 14 backend and all 23 browser steps. Exact summaries:

- `.cartulary/test-results/nf-imp08-s56-validation-20261030-02/target-summaries/test-slice.json`
- `.cartulary/test-results/nf-imp08-s56-browser-20261030-02/target-summaries/test-slice.json`

Both corresponding exact-root observers exited 0. The two ordinary selector
browser rows passed in browser 01, and 009 plus typed failure assertions passed
in validation 01; their exact earlier identities and the failed aggregate are
retained in §27.43. Only the obsolete test expectation changed between those
sources. This is workstream evidence, not a claim of final S36 compatibility.
V49's required behaviors and cleanup have passing observations. No correction
remains open for 009. S56 awaits its closing documentation checkpoint; S33-D
then resumes with 008/018/025/028 outstanding. Source reversion remains the
correction rollback, subject to the previously recorded migration-34 rule.

S56 closing `make lint-markdown`, run `nf-imp08-s56-close-20261030-01`,
exited native 0; retained summary:
`.cartulary/test-results/nf-imp08-s56-close-20261030-01/adhoc/lint-markdown/tool-run-summary.json`.
S56 is DONE and F57 resolved before S33-D resumes. Twenty families now have
substantive passing comparisons. S33-D next authors 025's exact contributor
pages, current authorization, stale digest and rejected-row exclusion, followed
by the three remaining resource-bound families.

### 27.45 S33-D contributor comparison entry

S56 is closed and S33-D has resumed. Fixture 025 advances from revision 2 to
revision 3 with real Imports publication into two tables, four accepted rows and
one rejected source row. Independent expectations cover the complete graph/v2
adapter input, bounded examples, four ordered edge contributor pages and their
cursor payloads, complete vertex contributors, viewer access, omitted-limit
defaulting, stale digest and absent endpoint failures, and membership revocation
and restoration through the existing supported authorization control. Rejected
rows cannot contribute. Unauthorized requests must preserve owner state; restoring
membership must recover the same live continuation. Scenario prose remains
non-executable. Existing manifest/scenario schemas remain v2.

The adapter retains actual errors in the complete comparison; it does not convert
unexpected failures into expected successes or skip dependent state assertions.
The membership control changes only fixture authorization setup. Product routes
perform the admission and current-scope checks. Fixture bytes, independent
expected values, actual HTTP envelopes and durable state remain separately
identifiable in retained evidence.

The newly authored and generated selector is:
`module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2`
(`TestNetworkFlowContributorsFixtureComparison`). The authored collaborator set
includes Incidents for live membership transitions. This registration is available;
no passing execution is claimed at entry.

With the §27.5 pinned `GO` and `installed_only` preparation prefix, these commands
completed with native exit 0 before comparison execution:

- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-author-contributors-20261030-01 make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Contributor fixture compares exact graph vertex edge pages live authorization stale digest and rejected row exclusion' SELECTOR_KEY=TestNetworkFlowContributorsFixtureComparison`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-format-20261030-04 make format`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-generate-20261030-03 make generate`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-guide-20261030-03 make task-guide OWNER=module.networkflow`

S33-D remains **IN_PROGRESS**. Twenty fixture families have passing behavioral
comparisons; 025 awaits its first execution, and 008/018/028 still require rebuilt
resource-bound comparisons. S33-E is not ready. Any adopted-contract violation
exposed by 025 receives a reproduction and a separately tracked correction before
runtime editing; the fixture expectation must not be weakened to preserve it.

The first launch command used an unsupported selector variable:
`CARTULARY_TEST_RUN_ID=nf-imp08-s33d-contributors-reproduction-20261030-01 make test-slice OWNER=module.networkflow ROW_IDS=module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2`.
Native exit 2 was workspace configuration admission (`unknown input ROW_IDS`),
before product execution. No fixture result or cleanup qualification is claimed.
The public recipes and prior successful exact commands use `ROWS`; the corrected
launch retains the identical single-row scope under a fresh run identity:

```sh
CARTULARY_TEST_RUN_ID=nf-imp08-s33d-contributors-reproduction-20261030-02 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2
```

### 27.46 S33-D contributor reproduction and S57 correction entry

The corrected single-row reproduction command in §27.45 exited native 2 with
`product/test_assertion_failure`. Its source identity is
`sha256:077f8484f5d0a4b4178a26c1a7b563c830690134cebe3c56a34b55c8c1ac09e0`;
HEAD/toolchain/system remain the recorded IMP-08 values. All 14 cleanup steps
passed. The exact target summary is
`.cartulary/test-results/nf-imp08-s33d-contributors-reproduction-20261030-02/target-summaries/test-slice.json`.
The retained comparison is
`NF-FIX-025-graph-contributors-2389202327/comparison-summary.json` beneath that
root, manifest SHA-256
`ab99bf1b67599a68e6e9c7ada35ddaba792bd25306039ca9ae060e36601bd73f`.
The native-0 observer was
`CARTULARY_TEST_RUN_ID=nf-imp08-s33d-contributors-observe-20261030-01 make explain-run RESULTS_DIR=.cartulary/test-results/nf-imp08-s33d-contributors-reproduction-20261030-02 TARGET=test-slice DETAIL=summary`
with the same pinned environment.

Complete expected/actual comparison has exactly three differences: omitted-limit
status is 400 instead of 200, its data is absent instead of the four complete
contributors, and an unexpected `network_flow_invalid_request` error reports
`field=limit`, `reason_code=missing_member`. Every other value, including all
pagination payloads, current authorization, recovery, graph input/result, audit
and durable state, matches the independent oracle. The fixture remains revision
3; these expectations require no change.

**F58 — Contributor admission incorrectly requires the optional query limit.**
NF-REQ-141c explicitly applies Table 13-A's omission default and explicit-value
bounds; Table 20-A applies that same policy to query admission. The current v2
selector amendment (NF-REQ-191) and major-8 precedence (NF-REQ-212) do not revoke
it. Both ephemeral and saved contributor decoders duplicate limit validation and
reject omission, and both downstream request schemas incorrectly require the
member. This is an implementation/projection conflict with adopted behavior.

S57 starts from the failing 025 reproduction, independently of S33-D completion.
Its remedy is one private limit admission function shared by paginated request
decoders, preserving each route's existing field-validation order. Only omission
receives `min(200, effective max_query_limit)`; explicit null, non-integers and
out-of-range values remain failures and are never clamped. Correct the two
contributor request projections in authored schema inputs and regenerate through
Make. Add focused checks for both contributor variants and existing row/diagnostic
routes, bounded defaults, exact explicit-limit failures and closed continuation
admission. No schema ID, contract major, durable version or migration changes
are needed. Existing explicit valid requests retain their behavior; omitted
limits gain the adopted bounded default. The long-term benefit is one clear
query-limit policy without another mandatory client compatibility field.

Leaving F58 open perpetuates contradictory schemas and inconsistent pagination.
Correction risks are accidentally treating null as omission, changing error
precedence, or accepting explicit values above the deployment bound. Tests must
reject these changes. Rollback is source/projection reversion subject to the
already recorded migration-34 deployment constraints.

**V50** includes these existing registered rows and the new 025 row:

- `module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2`
- `module.networkflow.unit.query_authoring_admission`
- `module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58`
- `module.networkflow.integration.bounded_graph_contributor_pipeline`
- `module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519`
- `module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_v_171429ad9e`
- `module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle`

Affected generated/schema checks and the closing Markdown check are also required.
S57 exits only after these obligations and cleanup pass. S33-D waits for that exit;
no later fixture workstream has begun.

S57 entry Markdown (`nf-imp08-s57-entry-20261030-01`) and formatting
(`nf-imp08-s57-format-20261030-01`) exited native 0. Before runtime edits,
`CARTULARY_TEST_RUN_ID=nf-imp08-s57-unit-reproduction-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.query_authoring_admission`
reproduced the same omission defect for both contributor decoders at deployment
maxima 32, 500 and 1000. Native exit was 2/product assertion, source
`sha256:36d3cef90b01cc8967177c5621c5bfeee6cf94932ded27311978b2a5525a9320`;
all ten cleanup steps passed. Exact summary:
`.cartulary/test-results/nf-imp08-s57-unit-reproduction-20261030-01/target-summaries/test-slice.json`.
Observer `nf-imp08-s57-unit-observe-20261030-01` used
`make explain-run RESULTS_DIR=.cartulary/test-results/nf-imp08-s57-unit-reproduction-20261030-01 TARGET=test-slice DETAIL=summary`
and exited 0. Bounded published stderr inspection confirmed all six omission
failures; no test expectation is being changed to accept missing-member errors.

S57 formatting 02, generation 01 and task-guide 01 passed native 0. Generated
TypeScript correctly makes `limit` optional on both contributor requests. The
schema edit also changes the derived complete Imports registry digest from
`e546bb797cc033d43d24b3eaa8579e1405283ced322c4290043d1071aea54b0b` to
`6feb41d9fde5fc1633c18abc54293c102e7f957ab2dc4d5c72de99479b71ce60`.
This is an additional publication compatibility consequence, not a new storage
format. Migration 34 already fences every unfinished Core apply job and resets
pending Network Flow approvals; its SQL requires no further amendment. Historical
migration files remain untouched. V50 therefore also selects the existing
`module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`
and `module.networkflow.browser.import_recovery` rows to validate that boundary
against the newly projected registry. No frozen fixture bytes reference the
superseded registry hash, so fixture revisions do not change.

The exact correction validation commands, all with the pinned environment, are:

```sh
CARTULARY_TEST_RUN_ID=nf-imp08-s57-validation-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca
CARTULARY_TEST_RUN_ID=nf-imp08-s57-browser-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_v_171429ad9e,module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle,module.networkflow.browser.import_recovery
CARTULARY_TEST_RUN_ID=nf-imp08-s57-generated-20261030-01 make source-check CHECK=generated
CARTULARY_TEST_RUN_ID=nf-imp08-s57-schema-20261030-01 make source-check CHECK=schema
```

Results remain unqualified until native completion, retained summaries and cleanup
are reconciled. The original F58 reproduction remains historical failed evidence.

### 27.47 S57 correction validation and exit

All four V50 execution commands in §27.46 passed native 0, including retained
finalization. They share source
`sha256:e1e0875cc7c2cb1165f28ebbf68c06646759512d20807821eaed2a0c1bb5e519`,
the unchanged baseline HEAD, and the recorded toolchain/system digests. Cleanup
completed without failed or blocked steps: backend validation 18, browser 24,
generated 14, schema 12. Exact roots are
`.cartulary/test-results/nf-imp08-s57-<kind>-20261030-01/`, where kind is
`validation`, `browser`, `generated`, or `schema`. Backend/browser summaries are
`target-summaries/test-slice.json`; source-check summaries are
`target-summaries/source-check.json`. Each root retains the exact selected-row
results, execution manifest, cleanup results and secret-scan finalization.

The unchanged 025 revision-3 fixture passes at
`NF-FIX-025-graph-contributors-2782782897/comparison-summary.json` beneath the
validation root, with manifest SHA-256
`ab99bf1b67599a68e6e9c7ada35ddaba792bd25306039ca9ae060e36601bd73f`.
It compares every admitted expected value and all four accepted contributors;
omitted limit now returns the complete vertex page with effective limit 200.
The focused unit comparison proves the bounded default of 32 and default 200 at
larger deployments on all five paginated route variants. Explicit invalid values
remain rejected. The schema projection checks and both existing contributor
pipeline and migration/drain regressions pass. Edge, vertex, saved exact-result
lifecycle and import-recovery browser workflows all pass.

Observers used fresh IDs `nf-imp08-s57-observe-<kind>-20261030-01` and
`make explain-run RESULTS_DIR=.cartulary/test-results/nf-imp08-s57-<kind>-20261030-01 TARGET=<target> DETAIL=summary`,
with target `test-slice` for validation/browser and `source-check` for
generated/schema. All four observers exited native 0. Their success establishes
inspection only; native execution exits and retained results above establish
V50's passing phase evidence.

F58 is repaired in runtime, authored schemas, generated projections, tests and
handoff documentation. The current Imports registry digest is
`6feb41d9fde5fc1633c18abc54293c102e7f957ab2dc4d5c72de99479b71ce60`.
Migration 34's previously recorded all-Core-apply drain requirement remains
mandatory; no SQL amendment or public version change was introduced. Original
failed reproductions remain failed historical evidence. Final frozen-source
qualification still belongs to S36.

The only remaining S57 exit action is its documentation checkpoint:
`CARTULARY_TEST_RUN_ID=nf-imp08-s57-close-20261030-01 make lint-markdown` with the
same pinned environment. After native success, record S57 DONE/F58 resolved and
resume S33-D. Twenty-one fixture families then have passing behavioral evidence;
008/018/028 remain in S33-D, and S33-E–F/S34–S36 remain open.

S57 closing Markdown passed native 0 at
`.cartulary/test-results/nf-imp08-s57-close-20261030-01/adhoc/lint-markdown/tool-run-summary.json`.
S57 is DONE and F58 resolved before S33-D resumes. Its remaining resource-bound
fixtures are 008, 018 and 028; twenty-one families now have passing comparisons.

### 27.48 S33-D aggregate-bound comparison

After S57 closure, fixture 028 advances from revision 2 to revision 3. Its six
independent real-import cases replace out-of-range source counters with legal
uint64 values. They compare an exact twenty-digit sum above uint64 range, bytes
and packets overflow at the next aggregate digit, and simultaneous
vertex-before-edge-before-bytes-before-packets failures during ordered iteration.
The captured production composer verifies exact visited-row counts, no projection
invocation or partial result on failure, complete successful v2 input/output,
unchanged imported rows and absence of failed-query publication/audit effects.
The independent oracle uses decimal integer arithmetic and authored failure
expectations; it reads no product snapshot.

The registered selector is
`module.networkflow.integration.aggregate_bounds_fixture_compares_legal_source_c_7934db1cd2`
(`TestNetworkFlowAggregateBoundsFixtureComparison`). Before execution, these
commands passed native 0 with the pinned environment:

- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-author-aggregate-20261030-01 make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Aggregate bounds fixture compares legal source counters exact big integer sums digit limits failure priority and absent partial graph output' SELECTOR_KEY=TestNetworkFlowAggregateBoundsFixtureComparison`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-format-20261030-05 make format`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-generate-20261030-04 make generate`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-guide-20261030-04 make task-guide OWNER=module.networkflow`

The first comparison command is
`CARTULARY_TEST_RUN_ID=nf-imp08-s33d-aggregate-reproduction-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.aggregate_bounds_fixture_compares_legal_source_c_7934db1cd2`.
No passing result is claimed at entry. S33-D remains IN_PROGRESS and later
workstreams remain unstarted.

The 028 comparison passed native 0 with all six cases and all 14 cleanup steps.
Source identity is
`sha256:de5383d97e81f943d313c69f877d10ac6774bbf12a3b7f8205f995851ae0fe4c`;
HEAD/toolchain/system remain unchanged. The exact target summary is
`.cartulary/test-results/nf-imp08-s33d-aggregate-reproduction-20261030-01/target-summaries/test-slice.json`.
Its passing comparison is
`NF-FIX-028-graph-aggregate-bounds-1789152320/comparison-summary.json`, manifest
SHA-256 `d40a5d3c30854f291cc3e5512b9e63153d4588364e64f8b4207c42380209551e`,
beneath that root. Observer `nf-imp08-s33d-aggregate-observe-20261030-01` used
`make explain-run RESULTS_DIR=.cartulary/test-results/nf-imp08-s33d-aggregate-reproduction-20261030-01 TARGET=test-slice DETAIL=summary`
and exited 0. No product correction or fixture expectation relaxation was needed.
Twenty-two families have passing behavioral comparisons. S33-D remains open for
008 and 018; S33-E is not yet ready.

### 27.49 S33-D large-input and table-quota comparison

Fixture 008 advances from revision 2 to revision 3. A 64-record source uses valid
UTC timestamps through the hour rollover; its independent expectations compare
exactly 50 preview records and source-column samples, all 64 persisted rows,
complete graph/v2 input, bounded examples, and graph rejection after the 33rd
contributing row when the request lowers the limit to 32. Separate lower vertex
and edge limits establish their first-excess observations without invalid source
values masking graph behavior.

Two isolated quota cases import a real first table and attempt another through
Imports. One reaches the active-table cap; the other soft-deletes the first table,
replays deletion without new effects, and reaches the retained-table cap. They
compare complete publication/failure objects, retained rows, table status, audit,
receipts and counts. Core-facing errors follow NF-REQ-088d's closed
`network_flow_target_unavailable` shape with `reason_code`; direct graph errors
retain their full limit/actual/phase details. No unsupported owner-error members
or new wire codes are invented. These fixtures make engineering claims only.

The authored/generated selector is
`module.networkflow.integration.large_limits_fixture_compares_valid_imported_row_470de5552b`
(`TestNetworkFlowLargeLimitsFixtureComparison`). These pinned-environment commands
passed native 0 before execution:

- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-author-large-20261030-01 make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Large limits fixture compares valid imported rows preview bounds bounded graph scans and active retained table quotas through Imports' SELECTOR_KEY=TestNetworkFlowLargeLimitsFixtureComparison`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-format-20261030-06 make format`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-generate-20261030-05 make generate`
- `CARTULARY_TEST_RUN_ID=nf-imp08-s33d-guide-20261030-05 make task-guide OWNER=module.networkflow`

First comparison:
`CARTULARY_TEST_RUN_ID=nf-imp08-s33d-large-reproduction-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.large_limits_fixture_compares_valid_imported_row_470de5552b`.
S33-D remains IN_PROGRESS pending 008/018 qualification and its group exit.

The 008 comparison passed native 0 with all 14 cleanup steps successful at source
`sha256:8a775a8edb1049cc14555ff52d3a4cfdb444197c2cceed98a121e05e40372476`.
HEAD/toolchain/system remain unchanged. The exact target summary is
`.cartulary/test-results/nf-imp08-s33d-large-reproduction-20261030-01/target-summaries/test-slice.json`.
The passing comparison is
`NF-FIX-008-large-limits-949177351/comparison-summary.json`, manifest SHA-256
`bdee707a60c85fad3019760a0e4909647065b6baf49bd47036eded7ed9a45146`, beneath that root.
Observer `nf-imp08-s33d-large-observe-20261030-01` used
`make explain-run RESULTS_DIR=.cartulary/test-results/nf-imp08-s33d-large-reproduction-20261030-01 TARGET=test-slice DETAIL=summary`
and exited 0. The 50/64 preview/apply distinction, complete graph, early stopping,
both table quotas, deletion/replay and retained failure effects matched without
runtime changes. Twenty-three fixture families have passing behavioral evidence.
Only 018 remains before S33-D's group exit; S33-E is not yet ready.


### 27.50 S33-D — 018 resource-limit comparison preparation

Fixture 018 is now revision 3, pending its first behavioral run. Its typed adapters
use the actual Core configuration loader, production CSV decoders, Imports
publication and Graph Projection route/composer. Eleven configuration cases compare
defaults, meaningful zero values, invalid objects/members, numeric limits,
cross-limit relationships and unclaimed configuration. Five decoder cases compare
column, Unicode scalar and row bounds plus exact boundaries and preview isolation.
Seven publication cases compare accepted/raw-row caps, zero/exact/overflow retained
diagnostics, all-rejected sampling independent of retention, and overlong raw cells.
Two graph cases compare deployment graph/counter limits and absence of partial output.
Expected bytes use the independent reference authorship described in §27.48–49.
The column case includes five columns against a limit of two. Table 9-B requires
column checks after complete header decoding; its observed header width is five.
NF-REQ-172 first-excess streaming accounting is separately exercised by the row
case (limit two, actual three), rather than overriding header decoding precedence.

Fixture support adds a narrow listed-only `Bundle.InputPath` accessor for file-based
product readers. The Core loader reads the admitted read-only TOML copy, never the
repository fixture. Support negatives cover absolute, traversal and unlisted paths;
source mutation after admission cannot alter the file seen by the reader. No service
orchestration or product policy moves into support. `platform.config` is a declared
collaborator on the new owner row.

`nf-imp08-s33d-author-resource-20261030-01` ran
`make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Resource limits fixture compares application configuration decoder bounds streaming import retention and graph failures with retained effects' SELECTOR_KEY=TestNetworkFlowResourceLimitsFixtureComparison`
and exited 0, returning actual row
`module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f`.
Registration is authored; generation, guide confirmation and behavioral execution
remain pending. The first selection also includes
`module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a`
and `module.networkflow.unit.configuration_claim_and_manifest_admission`.
S33-D remains in progress and S33-E is not ready.


Before first execution, owner review against NF-REQ-073/Table 9-B corrected two
fixture assumptions and advanced 018 to revision 4. Header bounds follow complete
header decoding, while row limits apply while streaming in either mode. The preview
case now stops at 50 before a configured row cap of 51; the Imports failure uses 52
records against that cap so it reaches apply after a valid preview. Revision 3 had no
behavioral qualification. No runtime change or expectation derived from captured
product output was involved.

Preparation `make format`, `make generate` and `make task-guide OWNER=module.networkflow`
passed with run IDs `nf-imp08-s33d-format-20261030-07`,
`nf-imp08-s33d-generate-20261030-06` and `nf-imp08-s33d-guide-20261030-06`.
The resource row is registered and available. Its first execution is:

```sh
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s33d-resource-reproduction-20261030-01 make test-slice ROWS=module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a,module.networkflow.unit.configuration_claim_and_manifest_admission
```


The `resource-reproduction-20261030-01` launch exited 2 during configuration
admission because `test-slice` requires exactly one of `OWNER` or `SUITE`. No product
callback ran and no behavioral result was produced. The corrected launch uses fresh
ID `nf-imp08-s33d-resource-reproduction-20261030-02` and the same `ROWS` selection,
with `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a,module.networkflow.unit.configuration_claim_and_manifest_admission`.


The corrected 018/support/configuration selection passed native 0, with all 16
cleanup steps successful, at source
`sha256:30666b95de4049f6812d7b7b7d9c20671b4db4f0898b0af8d71bd48da48dca75`.
HEAD/toolchain/system remain unchanged. Summary:
`.cartulary/test-results/nf-imp08-s33d-resource-reproduction-20261030-02/target-summaries/test-slice.json`.
Comparison: `NF-FIX-018-resource-limits-4189476578/comparison-summary.json` beneath
that root, manifest SHA-256
`3a722b13a07ef323c1a68cb2d95702ca2d68687d660375b19745554ac1ba66c0`.
Observer `nf-imp08-s33d-resource-observe-20261030-01` ran
`make explain-run RESULTS_DIR=.cartulary/test-results/nf-imp08-s33d-resource-reproduction-20261030-02 TARGET=test-slice DETAIL=summary`
and exited 0. All 25 cases matched independently authored outputs without runtime
changes. Twenty-four families now have passing comparisons.

### 27.51 S33-D group exit validation

The remaining six D families are rerun because the listed-input accessor now serves
all fixture reads. The unchanged current-source 018/support/configuration pass above
is retained without automatic duplication. Authored routing additions also require
generated/catalog consistency. The following commands use the established explicit
Go 1.27.3 and installed-only environment:

| Fresh run ID | Exact Make command |
| --- | --- |
| `nf-imp08-s33d-group-20261030-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49,module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2,module.networkflow.integration.aggregate_bounds_fixture_compares_legal_source_c_7934db1cd2,module.networkflow.integration.large_limits_fixture_compares_valid_imported_row_470de5552b` |
| `nf-imp08-s33d-generated-20261030-01` | `make source-check CHECK=generated` |
| `nf-imp08-s33d-catalog-20261030-01` | `make source-check CHECK=catalog` |

S33-D remains in progress pending these outcomes and the documentation checkpoint.


All three exit executions passed native 0 at the same source as the 018 pass:
`sha256:30666b95de4049f6812d7b7b7d9c20671b4db4f0898b0af8d71bd48da48dca75`.
The group, generated and catalog runs respectively completed 14, 14 and 11 cleanup
steps with no failures or blocked cleanup. Their exact summary paths are beneath
`.cartulary/test-results/<run-id>/target-summaries/`, named `test-slice.json` for the
group and `source-check.json` for the other two. Observers
`nf-imp08-s33d-observe-group-20261030-01`,
`nf-imp08-s33d-observe-generated-20261030-01` and
`nf-imp08-s33d-observe-catalog-20261030-01` each ran
`make explain-run RESULTS_DIR=.cartulary/test-results/<corresponding-run-id> TARGET=<corresponding-target> DETAIL=summary`
and exited 0. Source commit, toolchain and system identities are unchanged.

The group retained these passing comparisons beneath its exact run root:

| Fixture | Comparison summary | Manifest SHA-256 |
| --- | --- | --- |
| NF-FIX-006-cross-table-graph | `NF-FIX-006-cross-table-graph-1973735903/comparison-summary.json` | `eeae13f8a8de031e8c4f4c83adfe733c35187d4f08534e920030f8997662f1fc` |
| NF-FIX-008-large-limits | `NF-FIX-008-large-limits-3171974930/comparison-summary.json` | `bdee707a60c85fad3019760a0e4909647065b6baf49bd47036eded7ed9a45146` |
| NF-FIX-009-soft-delete-stale-graph | `NF-FIX-009-soft-delete-stale-graph-3708622529/comparison-summary.json` | `4ec59085d9817392cf69bb9f7ce53896416605eccc33b170bd123bd3c63024c0` |
| NF-FIX-015-graph-adapter-input | `NF-FIX-015-graph-adapter-input-2311273442/comparison-summary.json` | `bdd09489ddada650316702a0bbb8060ec761734ce4b0cf8d67bb90a2ded1fc79` |
| NF-FIX-025-graph-contributors | `NF-FIX-025-graph-contributors-2362124465/comparison-summary.json` | `ab99bf1b67599a68e6e9c7ada35ddaba792bd25306039ca9ae060e36601bd73f` |
| NF-FIX-028-graph-aggregate-bounds | `NF-FIX-028-graph-aggregate-bounds-4014316785/comparison-summary.json` | `d40a5d3c30854f291cc3e5512b9e63153d4588364e64f8b4207c42380209551e` |


S33-D's behavioral exit is satisfied: all seven families execute their registered
scenarios, compare exact graph/adapter/cursor identities and authorized durable
state, enforce bounds using valid imported values, and reject wrong expectations.
S56/F57 and S57/F58 corrections are closed with their separately retained regressions.
F39's corrected v2 registry reference is exercised by fixture 015. F38/F40 remain open
until all-28 corpus closure. F36/F37 and S36 remain required. No additional database
migration or runtime compatibility layer was introduced by D's remaining fixture work.
All these are phase evidence, not final frozen-source qualification.

The documentation checkpoint is `nf-imp08-s33d-close-20261030-01 make lint-markdown`
with the established environment. S33-D remains pending that result; S33-E begins only
after the tracker records completion.


The S33-D Markdown checkpoint passed native 0. S33-D is DONE; its dependencies,
behavioral exit, generated/catalog checks and cleanup are satisfied. S33-E is now
IN_PROGRESS for 007/016/017/026. This checkpoint precedes all S33-E changes.


### 27.52 S33-E — linking, privacy and audit preparation

S33-E covers the remaining 007/016/017/026 families against NF-REQ-142–151a,
Tables 15-A–F/16-A–C and the current public route contracts. The first family,
007, is now revision 3 with two isolated real-import setups and nine concrete link
requests. It covers existing targets, real Core creation and canonical dedupe,
new/reused binding tuples, exact replay, caller-ref ordering, graph vertex/edge
selection, deterministic truncated contributors, and reuse across selector kinds.
The adapter compares full link responses, persisted bindings and receipts, Core
canonical identities, row versions and attribution, exact audit effects, and zero
observation/lifecycle mutations. Every replay compares raw durable state and IDs.

Time, source/table identities and safe-digest keys use established controls. As in
027, opaque cryptographic binding/Core IDs and wall-clock receipt/seed metadata are
validated and retained verbatim, then aliased only in frozen semantic comparisons.
Derived row/graph identities, request digests and every source-ref member remain exact.
No global random-source replacement or production determinism hook is introduced.
Core's private dedupe-key representation is retained and checked unchanged on replay;
the frozen Network Flow comparison asserts canonical values and dedupe effects rather
than duplicating Reference Data's private algorithm implementation.

A typed POST observer now retains successful creation as well as query responses;
unexpected status/data/error variants still fail the frozen comparison. Product
behavior and the migration remain unchanged. Linking adapters stay in owner tests.

`nf-imp08-s33e-author-linking-20261030-01` ran
`make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Indicator linking fixture compares Core create and dedupe binding identities exact replay source references and atomic durable effects' SELECTOR_KEY=TestNetworkFlowIndicatorLinkingFixtureComparison`
and returned actual row
`module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9`,
native 0. Its authored collaborators include Indicators and Cross-owner Transactions.
Two offline authorship errors (an undefined HMAC helper and a nested selector lookup)
were corrected before writing any revised frozen bytes. The interrupted registration
left `nf-imp08-s33e-format-20261030-01 make format` failing native 2 at runtime-plan
configuration admission. No format/test work graph or product comparison was produced;
only the successful harness bootstrap summary exists in that run root. Observer
`nf-imp08-s33e-format-observe-20261030-01` exited 0 and confirmed missing graph evidence.
Registration is now complete; a fresh preparation run follows. Generation, guide
confirmation and first behavioral execution remain pending.


007 preparation passed: `nf-imp08-s33e-format-20261030-02 make format`,
`nf-imp08-s33e-generate-20261030-01 make generate`, and
`nf-imp08-s33e-guide-20261030-01 make task-guide OWNER=module.networkflow`, all native 0.
The registered first comparison uses:

```sh
GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only CARTULARY_TEST_RUN_ID=nf-imp08-s33e-linking-reproduction-20261030-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9
```


The first 007 run exited 2, with all 14 cleanup steps successful, at source
`sha256:d6ca953287095126b645474fb0b730ce4c5068d9043071c6eba381cc8779d777`.
The exact target summary is
`.cartulary/test-results/nf-imp08-s33e-linking-reproduction-20261030-01/target-summaries/test-slice.json`;
its failed comparison is `NF-FIX-007-indicator-linking-3071632167/comparison-summary.json`
beneath that root, revision 3 manifest SHA-256
`5b55e503991207cb38735cc6f5f03c292faa9cb6c58b4eea638944688c27ba37`.
Observer `nf-imp08-s33e-linking-observe-20261030-01` ran
`make explain-run RESULTS_DIR=.cartulary/test-results/nf-imp08-s33e-linking-reproduction-20261030-01 TARGET=test-slice DETAIL=summary`
and exited 0. HEAD/toolchain/system are unchanged.

Investigation found exactly five differing occurrences of two row-ref request hashes;
all nine operations, graph results, replays and durable effects otherwise matched.
The oracle incorrectly applied dynamic-object ordering to closed nested row refs.
NF-REQ-031/Table 6-C and NF-REQ-107/Table 12-B require the declared order
`network_flow_table_id`, `network_flow_row_id`, `source_row_number`,
`mapping_fingerprint`. Independently encoding that order gives
`c2cc573c45a53002abfb1ea9557c027338fd5ab969a5953b93117b656658832d`
and `4017e63a0a9eb84cedcb6422f42bb6ac78ae4a7c5ab4e0c52cf77fab8ffaa3be`,
matching the retained receipts. This is an oracle correction, not a runtime defect
or a relaxation of identity checks. Revision 4 freezes the corrected independently
computed hashes; no production change or migration is required.

The fresh run is `nf-imp08-s33e-linking-validation-20261030-01`, with the same exact
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9`
selection and established environment. S33-E remains in progress.


Revision 4's 007 comparison passed native 0 with all 14 cleanup steps successful
at source `sha256:2f712c8a972d3827cef54ea8f74707dc6796199e9ae0f35ce06c0fe4237cb15a`.
The target summary is
`.cartulary/test-results/nf-imp08-s33e-linking-validation-20261030-01/target-summaries/test-slice.json`.
The passing comparison is `NF-FIX-007-indicator-linking-2564575883/comparison-summary.json`
beneath that root, manifest SHA-256
`884524611efe26879556c742cb6afbb08dc29476d00748012713ccdd53733c6f`.
Observer `nf-imp08-s33e-linking-observe-20261030-02` used the same observer command
with that validation root and exited 0. Twenty-five families have passing behavioral
evidence. S33-E next completes 017 dependency/mismatch cases, then 016/026; it remains
in progress. No runtime defect was found in 007.


### 27.53 S33-E — 017 mismatch and Core rollback comparison

Revision 3 now contains eighteen concrete operations after real import and graph
construction. Cases cover non-IP/unmapped fields, exact and noncanonical confirmation
mismatches, missing/rejected locators, duplicate refs, conflicting candidates, target
value/type/visibility, wrong graph endpoint, stale digest and a viewer. Two controlled
failures use the existing Indicator participation port: before creation, and after
real Core creation is visible inside the borrowed transaction. The latter asserts
actual creation before injecting failure, then compares rollback of Core rows,
bindings, receipts and audit. A retry of that same uncommitted intent must succeed;
exact replay is unchanged, and changed committed intent must conflict.

The private fixture composition helper accepts test-only decorators of existing
module ports. No production hook, public interface or service orchestration is added.
Every failed request compares its entire pre/post domain state. Optional safe `field`
location details are retained where supplied: Table 21-B specifies required link
context members, while that location distinguishes the invalid selector/target member
from the candidate's `field_key`.

`nf-imp08-s33e-author-mismatch-20261030-01` ran
`make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Indicator mismatch fixture compares authorization selector and target failures and rollback of real Core creation before successful retry' SELECTOR_KEY=TestNetworkFlowIndicatorMismatchFixtureComparison`
and returned actual row
`module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505`,
native 0. An offline selector-key typo was corrected before the initial revision was
written. Registration is authored; preparation and comparison follow.


017 exit PASS (26/28 families): preparation IDs `nf-imp08-s33e-format-20261030-03`,
`nf-imp08-s33e-generate-20261030-02`, and `nf-imp08-s33e-guide-20261030-02`
ran format/generate/task-guide, native 0. Validation
`nf-imp08-s33e-mismatch-reproduction-20261030-01` ran
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9`,
native 0, cleanup 14/14. Source digest:
`77674fcf0f52f427f90e5dde079049acc6867c938fc265a0efb0485fa87e8f35`.
The shared-port decorator change is also covered by the passing 007 comparison.
Under `.cartulary/test-results/nf-imp08-s33e-mismatch-reproduction-20261030-01/`:

- `NF-FIX-017-indicator-link-mismatch-135263875/comparison-summary.json`, manifest
  `043b81f172a532312c0d5a276c5101f836f7c425b294d4925099c4008c194ca5`.
- `NF-FIX-007-indicator-linking-2407759972/comparison-summary.json`, unchanged
  revision 4 manifest `884524611efe26879556c742cb6afbb08dc29476d00748012713ccdd53733c6f`.

`nf-imp08-s33e-mismatch-observe-20261030-01` used `make explain-run` with that
root, `TARGET=test-slice DETAIL=summary`, native 0. No runtime defect or new
migration impact was exposed. S33-E remains in progress for 016 and 026.


### 27.54 S33-E — 016 privacy and observable-channel comparison

The authored adapter captures real OpenTelemetry SDK records emitted by the application
observer, standard logger output, and all semantic telemetry payloads through a
delegating existing port. It compares the complete Network Flow records, validates
wall-clock timestamps/nonnegative durations before aliasing only those values, and
retains the raw capture. A bounded capture fails on overflow. Every captured log,
telemetry payload, all projection metadata and durable audit are scanned for the
fixture's source filename, raw IP/text/formula values and unkeyed hashes. A deliberate
forbidden-value insertion must be rejected. Authorized rows and diagnostics still
compare their complete outputs, including numeric safe samples and canonical values.

Revision 3 was authored, then advanced to 4 during pre-execution review to use the
adopted `invalid_syntax` diagnostic reason (consistent with 004), replacing an
incorrect authoring label. No runtime comparison generated these expectations.
`nf-imp08-s33e-author-privacy-20261030-01` ran
`make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Privacy fixture compares real preview diagnostics safe digests graph metadata telemetry and emitted logs without raw value disclosure' SELECTOR_KEY=TestNetworkFlowPrivacyFixtureComparison`,
native 0, yielding actual row
`module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388`.
Preparation and comparison remain required.


016 preparation (`nf-imp08-s33e-format-20261030-04`,
`nf-imp08-s33e-generate-20261030-03`, `nf-imp08-s33e-guide-20261030-03`)
passed, native 0. First selection
`nf-imp08-s33e-privacy-reproduction-20261030-01` ran
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388`,
native 2, cleanup 14/14. Source digest
`9563d9e06f3e3acb50079f869c714e97a8dd1ba49c888e8897667eacd3ebf5a0`.
The failure was fixture admission: the descriptive scenario summary was 548 characters,
exceeding its unchanged v2 schema maximum of 500. No product callback ran. Evidence:
`.cartulary/test-results/nf-imp08-s33e-privacy-reproduction-20261030-01/NF-FIX-016-redaction-3672172228/comparison-summary.json`,
manifest `0b578d6347ca6c392fe45b55255058b0de4deb1e3f03dbce867cdd687825619d`.
Observer `nf-imp08-s33e-privacy-observe-20261030-01` used that run root with
`make explain-run TARGET=test-slice DETAIL=summary`, native 0. Revision 5 shortens
only the scenario summary and advances the frozen envelope revision. Product behavior
and schema limits remain unchanged; comparison must rerun.

### 27.55 S33-E — 026 audit and replay comparison

The final family independently imports a source, creates and reuses a Core indicator
binding, and repeats both exact requests. An additional real successful graph request
must add exactly one graph-success audit; a missing-table query and changed committed
link intent must leave all state unchanged. Complete audit attribution, key IDs and
safe digests are frozen, with explicit raw-value exclusion. This is a typed family
adapter with concrete request data, not an executable scenario-summary interpreter.

`nf-imp08-s33e-author-audit-20261030-01` ran
`make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Audit fixture compares publication binding and graph success events exact replay and rejection without duplicate durable audit' SELECTOR_KEY=TestNetworkFlowAuditReplayFixtureComparison`,
native 0, producing actual row
`module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b`.
Revision 3 was advanced to 4 during pre-execution review: public graph requests use
the request v2 object and selected-table scope, while the response's semantic-query
v2 object is a separate contract. Registration is authored; no 026 comparison has run.


The combined 016/026 reproduction `nf-imp08-s33e-privacy-audit-reproduction-20261030-01`
ran `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b`,
native 2, cleanup 14/14, source
`89018afe65052ee7a88d1a95a174a5f5e7dc4feee3d9bd410502eefece1eeaa6`.
Preparation format/generate/task-guide IDs ending `format-20261030-05`,
`generate-20261030-04`, `guide-20261030-04` used the `nf-imp08-s33e-` prefix and
passed native 0. The observer `nf-imp08-s33e-privacy-audit-observe-20261030-01`
passed native 0. Under the reproduction root:

- `NF-FIX-016-redaction-2845141242/comparison-summary.json`, manifest
  `aea472515e4f47168685e0c58a0fae494e1a0762037d59843ca9b61e37ae28f3`.
  The retained transport establishes correct rejection of unsupported Cisco-profile
  `exporter_id` mapping. Revision 6 uses Table 9-F's supported input/output interface
  text fields for the same raw-text/formula privacy assertions.
- `NF-FIX-026-audit-and-replay-3969352234/comparison-summary.json`, manifest
  `14f4af0595be999a7defd73b6b7c673746016427fbffaf12831262e173edd51e`.
  The adapter had used `/graph/query`; the actual public route is `/graphs/query`.
  Its plain 404 body also exposed a test-evidence limitation: embedding an invalid
  `json.RawMessage` prevented transport publication. The shared observer now retains
  non-JSON bodies as exact base64 bytes and continues to fail the envelope assertion.
  Passing JSON observations and frozen 026 data are unchanged.

These are adapter/stimulus corrections, not runtime defects or migration changes.
Both comparisons remain required; there is no passing qualification for 016/026 yet.


`nf-imp08-s33e-format-20261030-06` passed native 0. The next selection,
`nf-imp08-s33e-privacy-audit-validation-20261030-01`, used the same two explicit
rows, native 2, cleanup 14/14, source
`6a5541f9e56aa8251711530922fed20911948e3816175922233ee4d24efce521`.
Both product workflows completed, then the privacy scans rejected
`network_flow_table_created.after_json.source_content_sha256`. Owner review establishes
that Table 16-B **requires that exact unkeyed source-integrity digest**; it is distinct
from the raw cell-value SHA prohibited in administrative channels by NF-REQ-149.
The independent expected audit objects already contained that required safe field.
Thus the extra exclusion was an incorrect test classification, not a product defect.

016 revision 7 keeps all raw-text/filename/candidate and cell-hash exclusions, separately
excludes the source-integrity hash from observability/graph metadata, and compares its
required exact audit value. 026 revision 5 removes only the erroneous source-integrity
hash exclusion from audit. Raw restricted-channel and final audit state are now
retained before scans, including on failure. Exact graph responses and all durable
state comparisons remain required.

The run's summaries are
`NF-FIX-016-redaction-521659492/comparison-summary.json` and
`NF-FIX-026-audit-and-replay-2604474685/comparison-summary.json` beneath
`.cartulary/test-results/nf-imp08-s33e-privacy-audit-validation-20261030-01/`.
`nf-imp08-s33e-privacy-audit-observe-20261030-02` observed that root with the same
summary command, native 0. No new runtime finding or migration consequence is claimed.


`nf-imp08-s33e-format-20261030-07` passed native 0. Validation
`nf-imp08-s33e-privacy-audit-validation-20261030-02` used the same two rows,
native 2, cleanup 14/14, source
`5d486df5afe99236a21115fb2f49d40d5dd8a94d9d515eb6649c9fcaad76557c`.
Observer `nf-imp08-s33e-privacy-audit-observe-20261030-03` passed native 0.
The complete actual comparisons are now retained. Under that run root:

- `NF-FIX-026-audit-and-replay-1187777053/comparison-summary.json`, manifest
  `4879b1624be9bb7d0e422c2ffcf3930e69c85fd9706681a41f66d7391f080782`.
  The sole difference was the adapter taking `link_final_state` after the additional
  graph request. That named stage snapshot now occurs immediately after link steps;
  `audit_final_state` still captures the later complete workflow. Frozen 026 bytes
  remain unchanged; no product behavior changed.
- `NF-FIX-016-redaction-3585057192/comparison-summary.json`, manifest
  `89de18807a52875ddb041523630b0411feeb5576abd661cf2b1bae46ec9fae99`.
  Every publication, row, diagnostic, graph, metadata and durable audit value matched.
  The initial test-installed SDK provider had been replaced by normal application
  telemetry bootstrap, so no logs were captured. The startup cleanup observation
  was also omitted from the initial expected capture.

016 revision 8 qualifies the actual configured OTLP/HTTP TLS exporter against a
bounded local collector. It uses existing deployment configuration, the platform's
qualified TLS test identities, and the existing SDK force-flush API. A subprocess
isolates immutable system-root admission, following the existing platform transport
test pattern; host trust is unchanged. The private publication helper accepts an
unexported environment override, so fixture JSON cannot configure infrastructure.
Full exported records and semantic telemetry remain retained; graph and asynchronous
startup cleanup are compared in separate complete sequences. Service version and
instance identity are pinned through supported configuration. No new production
hook or replacement application logger is introduced. Revalidation remains required.


`nf-imp08-s33e-format-20261030-08` passed native 0. Validation
`nf-imp08-s33e-privacy-audit-validation-20261030-03` used the two existing rows,
native 2, cleanup 14/14, source
`2cf1e93c2a072e493a120f542fad6b890a5cc8f1a84719140e0b4aa530fe7902`.
026's complete comparison passed at
`NF-FIX-026-audit-and-replay-1232117621/comparison-summary.json`, unchanged revision 5
manifest `4879b1624be9bb7d0e422c2ffcf3930e69c85fd9706681a41f66d7391f080782`.
The enclosing selection failed in 016's object-store readiness: the child had selected
only the collector CA, excluding the managed service CA. Its failed summary is
`NF-FIX-016-redaction-766771370/comparison-summary.json`, revision 8 manifest
`e47f0a9dfe58aec80c2d7e3b3b94d45d2be8a579b2ee5551998ab355a3cc98f2`.
All paths are beneath `.cartulary/test-results/nf-imp08-s33e-privacy-audit-validation-20261030-03/`.
Observer `nf-imp08-s33e-privacy-audit-observe-20261030-04` passed native 0.

The isolated child now preserves the managed CA and adds its collector CA to one
explicit process-local trust bundle before any TLS client starts. It neither changes
host trust nor changes roots after admission. This corrects the fixture setup, with
no product or frozen-data change. The 016 row additionally names its actual
`platform.cryptography` collaborator. Passing 026 comparison evidence is retained,
but the required combined/native validation still must pass.


Preparation `nf-imp08-s33e-format-20261030-09`,
`nf-imp08-s33e-generate-20261030-05`, `nf-imp08-s33e-guide-20261030-05` passed,
native 0. `nf-imp08-s33e-privacy-validation-20261030-04` selected only the actual
privacy row with `make test-slice OWNER=module.networkflow
ROWS=module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388`,
native 2, cleanup 14/14, source
`0415ffc3ab9ac41e2ff6ca9fbc7090f01473697cf4c061aa7fb630a3dac96cb9`.
The TLS exporter, privacy scans, complete source/row/diagnostic/graph/audit comparisons,
three graph logs, startup cleanup log and semantic telemetry all matched except eight
occurrences of two environment labels. `httptestx` loads the authored
`internal/testutil/fixtures/config/valid.toml`: `disconnected`, with import, incident
portability, Reference Pack and snapshot reporting claimed; the fixture additionally
claims Network Flow. Revision 9 corrects the expected resource values to those declared
inputs and their canonically ordered claim string. It does not alter application
claims or derive an expected snapshot from runtime output.

Evidence is
`.cartulary/test-results/nf-imp08-s33e-privacy-validation-20261030-04/NF-FIX-016-redaction-213655199/comparison-summary.json`,
unchanged revision 8 manifest
`e47f0a9dfe58aec80c2d7e3b3b94d45d2be8a579b2ee5551998ab355a3cc98f2`.
Observer `nf-imp08-s33e-privacy-observe-20261030-05` passed native 0. Revalidation
and the S33-E close checks remain required.


### 27.56 S33-E — behavioral exit and boundary closure

`nf-imp08-s33e-validation-20261030-01` ran
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b,module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9`,
native 0, cleanup 14/14. Source digest
`23ac7e0b279ef591b7c8bedbae796e71e3b121612189902260983c49d39549e1`;
toolchain/system identities remain the established IMP-08 identities. Observer
`nf-imp08-s33e-validation-observe-20261030-01` passed native 0.
All four E comparisons passed under that run root:

| Bundle | Revision | Comparison summary directory | Manifest SHA-256 |
| --- | --- | --- | --- |
| 007 | 4 | `NF-FIX-007-indicator-linking-2896729651` | `884524611efe26879556c742cb6afbb08dc29476d00748012713ccdd53733c6f` |
| 016 | 9 | `NF-FIX-016-redaction-2703392038` | `707ae16b1986902e39fcb83c5602281a721c906f5fd22f59093715e9a42b9328` |
| 017 | 3 | `NF-FIX-017-indicator-link-mismatch-3009828445` | `043b81f172a532312c0d5a276c5101f836f7c425b294d4925099c4008c194ca5` |
| 026 | 5 | `NF-FIX-026-audit-and-replay-2251552072` | `4879b1624be9bb7d0e422c2ffcf3930e69c85fd9706681a41f66d7391f080782` |

Each directory contains `comparison-summary.json` beneath
`.cartulary/test-results/nf-imp08-s33e-validation-20261030-01/`, with exact source and
expected bundle hashes, sealed inputs and actual/transport artifacts. All 28 families
now have passing behavioral phase evidence; final frozen-source qualification is
still S36's obligation.

`nf-imp08-s33e-generated-20261030-01 make source-check CHECK=generated` and
`nf-imp08-s33e-catalog-20261030-01 make source-check CHECK=catalog` passed native 0.
The additional `nf-imp08-s33e-boundary-20261030-01 make lint CHECK=backend-boundary`
failed native 2, cleanup 13/13, at the same source. The retained
`unit-logs/row-module.records.architecture.boundary/stdout.log` identifies exactly one
violation: 018's owner test directly imported `internal/platform/config`, forbidden by
`production-modules-and-http-no-deployment-config`. This is test-adapter coupling,
not a runtime requirement defect. The adapter now uses the existing
`configtest.DiagnosticsJSON` boundary, retaining the complete typed diagnostic and
failing on unexpected error types. No production rule, expectation or frozen byte is
weakened. Observer `nf-imp08-s33e-boundary-observe-20261030-01` passed native 0.
018 and the boundary check must pass before the documentation checkpoint and E exit.


The existing diagnostic helper deliberately exposes Core's complete JSON diagnostic
envelope, rather than the private error struct that 018 had serialized. Revision 5
of 018 now compares that public envelope (`error.code`, `error.details.items`) with
the same independently authored, ordered diagnostic items. Source files, stimuli,
limits, production responses and all remaining expected values are unchanged. This
removes the module test's deployment-config type coupling instead of adding a
boundary exception or a new support API. `nf-imp08-s33e-format-20261030-10` passed
native 0 before the literal revision update.


The boundary correction and full E group now pass together. Final E phase commands,
all with the established environment and source digest
`d4ec4f943b9dfbc035dd49d4edc2aea07cfa10eb9e76285af09fe75d5abcc002`:

- `nf-imp08-s33e-validation-20261030-02`:
  `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b,module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9,module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f`,
  native 0, cleanup 14/14.
- `nf-imp08-s33e-boundary-20261030-02`: `make lint CHECK=backend-boundary`,
  native 0, cleanup 13/13.
- Observers `nf-imp08-s33e-validation-observe-20261030-02` and
  `nf-imp08-s33e-boundary-observe-20261030-02` used their exact roots with
  `make explain-run`, respective targets and `DETAIL=summary`; both native 0.

The validation root contains these passing `comparison-summary.json` files:
`NF-FIX-007-indicator-linking-1612230977`, `NF-FIX-016-redaction-1650516875`,
`NF-FIX-017-indicator-link-mismatch-2028920680`,
`NF-FIX-018-resource-limits-1298184544`, and
`NF-FIX-026-audit-and-replay-4022576100`. E manifests are unchanged from the preceding
passing E table. 018's new revision 5 manifest is
`ab81f596972992948f53c434ec1a7aba5a3530215fa4f171a369924ef17b94b8`.
Generated/catalog phase checks passed before this test-only diagnostic projection
change; no authored generator/routing input changed afterward.

S33-E's behavioral and boundary exits are satisfied, pending its documentation
checkpoint. No E runtime correction or migration was required. The remaining risks
are the obsolete traversal and incomplete inventory/coverage closure (S33-F), obsolete
preview/apply accounting (S34/F36), writable identity/provenance inputs (S35/F37), and
final frozen-source qualification/handoff (S36). F38/F40 remain open until S33-F.
The next command is `nf-imp08-s33e-close-20261030-01 make lint-markdown`.


`nf-imp08-s33e-close-20261030-01 make lint-markdown` passed, native 0;
summary `adhoc/lint-markdown/tool-run-summary.json`. Generated/catalog cleanup was
14/14 and 11/11 respectively, with no failed or blocked operations. S33-E is DONE.
S33-F is now IN_PROGRESS; this update precedes its implementation. S33-F will retain
integrity separately, remove the error-discarding traversal, establish exact inventory
to executable coverage equality, integrate remaining browser fixture consumers, and
run all 28 comparisons with admission and wrong-expectation controls. F36/F37 and S36
remain required afterward.


### 27.57 S33-F — executable corpus closure

Removed the error-discarding `AssertAllFixtureRuntimeBehavior` traversal. V29-A's
existing row remains the independent integrity/inventory gate. Its inventory now
matches an explicit 28-family revision/test registry; each behavioral entry rejects
an unregistered caller/revision and empty outcomes. There is no revision-2 fallback.
Negative controls preserve fixture identity, revision, requests and case labels while
changing an existing observable scalar. The alias client comparison receives the
same stronger control. The isolated graph unit vector no longer reads frozen bytes
outside admission.

Browser consumers of 001 now receive a per-test admitted payload. Admission precedes
the browser callback, inputs remain sealed beneath the selected run root, and the
comparison summary records the actual browser test outcome. Listing tests performs
no fixture execution. Each test owns its workspace; imports require explicit input.
No product interface or frozen bundle bytes changed in this workstream.

`nf-imp08-s33f-format-20261030-01 make format` failed native 2, cleanup 11/11,
source `66df0a1f60c8dc5985c564eef37e33475a94f0222e740fe2fdf7bb83d63be747`.
The exact `format` summary and bounded frontend log identify the new empty fixture
argument pattern. Playwright requires literal destructuring to declare dependencies;
a targeted lint explanation preserves a dependency-free admission fixture. Existing
nonfatal warnings remain visible. The observer
`nf-imp08-s33f-format-observe-20261030-01 make explain-run` used this exact root,
`TARGET=format DETAIL=summary`, native 0. No product defect or additional migration
is implicated. Required exits remain all 28 comparisons, support controls, browser
admission integration, affected projections/types and closing Markdown.


Format 02 and generation 01 passed native 0. `nf-imp08-s33f-types-20261030-01`
(`make lint CHECK=types`) failed native 2, cleanup 11/11, source
`a9be7347427990ec10158b96564569c2e519b20ae41302bb6d1cf55e1d145230`.
The frontend typecheck reports only the new negative control's unchecked first array
entry. Added an explicit nonempty-outcome guard; no production change was needed.
The exact-root types observer passed native 0. Task-guide 01 passed; coverage detail
02 is used to confirm the full row selection rather than the summary's abbreviated
recommendations.


Format 03 and coverage task-guide 02 passed native 0; all 30 selected backend rows
are present in coverage output. The S33-F corpus command is
`nf-imp08-s33f-corpus-20261030-01`, with the established GO/preparation prefix:

```sh
make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.aggregate_bounds_fixture_compares_legal_source_c_7934db1cd2,module.networkflow.integration.atomic_fixture_compares_every_commit_boundary_ro_a4cdb920d4,module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b,module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2,module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9,module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.large_limits_fixture_compares_valid_imported_row_470de5552b,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f,module.networkflow.integration.retention_fixture_compares_soft_delete_incident_570ae9035a,module.networkflow.integration.source_changed_fixture_compares_every_source_ver_213b5f7eab,module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49,module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352
```

Concurrent missing obligations use separate fresh identities:

- `nf-imp08-s33f-types-20261030-02`: `make lint CHECK=types`.
- `nf-imp08-s33f-client-20261030-01`:
  `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309`.
- `nf-imp08-s33f-node-20261030-01`:
  `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.node_fixture_admission_seals_verified_inputs_and_d6cec88240,web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`.

These are workstream checks, not S36 qualification. The existing import-recovery
browser row will validate per-test admission through an actual browser workflow;
visual baselines are unchanged and no golden refresh is authorized or needed.


S33-F corpus validation passed native 0: exactly 28 unique passing product-family
summaries match the repository inventory; synthetic support controls are separate.
All wrong-observable-expectation controls passed. Client, Node support/ownership,
types, generated and catalog checks also passed native 0. Generated/catalog commands
were `make source-check CHECK=generated` and `make source-check CHECK=catalog`.
Native runs retain the unchanged HEAD/toolchain/system identities from §27.1:

| Run ID | Source digest | Completed cleanup |
| --- | --- | --- |
| `nf-imp08-s33f-corpus-20261030-01` | `sha256:8e725fa888c19499e92f3d8f752868a5aa026d7a61f07e325d0d1a78c92cfba5` | 16/16 |
| `nf-imp08-s33f-client-20261030-01` | `sha256:8e725fa888c19499e92f3d8f752868a5aa026d7a61f07e325d0d1a78c92cfba5` | 11/11 |
| `nf-imp08-s33f-node-20261030-01` | `sha256:8e725fa888c19499e92f3d8f752868a5aa026d7a61f07e325d0d1a78c92cfba5` | 11/11 |
| `nf-imp08-s33f-types-20261030-02` | `sha256:8e725fa888c19499e92f3d8f752868a5aa026d7a61f07e325d0d1a78c92cfba5` | 11/11 |
| `nf-imp08-s33f-generated-20261030-01` | `sha256:8e725fa888c19499e92f3d8f752868a5aa026d7a61f07e325d0d1a78c92cfba5` | 14/14 |
| `nf-imp08-s33f-catalog-20261030-01` | `sha256:8e725fa888c19499e92f3d8f752868a5aa026d7a61f07e325d0d1a78c92cfba5` | 11/11 |

Each artifact below is `comparison-summary.json` beneath the named directory in
`.cartulary/test-results/nf-imp08-s33f-corpus-20261030-01/`. These phase comparisons
are current at that source and are not a substitute for S36's final portfolio.

| Bundle | Revision | Manifest SHA-256 | Evidence directory |
| --- | --- | --- | --- |
| NF-FIX-001-cisco-sna-minimal | 8 | `d16582b1b7a79b62f05a7794873af5954efc8432449b2dc5b005cf4ebfade648` | `NF-FIX-001-cisco-sna-minimal-1489573851` |
| NF-FIX-002-cisco-sna-interface-fields | 5 | `9ab9720e77c9612a1c3d747e61ee9e9f5a46a664f4ae743630bcf9ea2318c1c9` | `NF-FIX-002-cisco-sna-interface-fields-1655858709` |
| NF-FIX-003-duplicate-headers | 4 | `a770fa3d5375a48b1e34eb421481f2f5e912b7010011e6917c569e8ad69f5252` | `NF-FIX-003-duplicate-headers-2027851132` |
| NF-FIX-004-rejected-rows | 6 | `deee3c770b3d9345def7727d1c100af1b8be57bac6dc59e075767aa5e3093313` | `NF-FIX-004-rejected-rows-2471697443` |
| NF-FIX-005-csv-parser-edges | 3 | `64756221ede6b99faa01f802e5652b65e22139ca1479c0fd3f61a9ff4ef6b616` | `NF-FIX-005-csv-parser-edges-1934421893` |
| NF-FIX-006-cross-table-graph | 3 | `eeae13f8a8de031e8c4f4c83adfe733c35187d4f08534e920030f8997662f1fc` | `NF-FIX-006-cross-table-graph-1559653142` |
| NF-FIX-007-indicator-linking | 4 | `884524611efe26879556c742cb6afbb08dc29476d00748012713ccdd53733c6f` | `NF-FIX-007-indicator-linking-3642536177` |
| NF-FIX-008-large-limits | 3 | `bdee707a60c85fad3019760a0e4909647065b6baf49bd47036eded7ed9a45146` | `NF-FIX-008-large-limits-3125499098` |
| NF-FIX-009-soft-delete-stale-graph | 3 | `4ec59085d9817392cf69bb9f7ce53896416605eccc33b170bd123bd3c63024c0` | `NF-FIX-009-soft-delete-stale-graph-1146745292` |
| NF-FIX-010-json-admission | 4 | `69de015263c02bb89be3f451d2ff554c671dca55e43a93afd4e1f212308e693f` | `NF-FIX-010-json-admission-3043581537` |
| NF-FIX-011-alias-collision | 4 | `94cce451a5b10f5aa26b18f428693c09082fead8d7f01112e89d38d0249a1ad1` | `NF-FIX-011-alias-collision-4038121195` |
| NF-FIX-012-sys-uptime-timestamps | 3 | `8c3c793cf01dc53f9191055333dc93e77257896e3e06cebfee64ac454133742e` | `NF-FIX-012-sys-uptime-timestamps-129453160` |
| NF-FIX-013-filename-display | 4 | `63a7afe2d29cd8c99d732d7084f4860ba0ca321b93171c71302c6ea822c8adf6` | `NF-FIX-013-filename-display-157032837` |
| NF-FIX-014-cursor-pagination | 5 | `0eabf0e75af97983bf82b18dec17bd1b381aedf54c6f80f0f01be758984b5134` | `NF-FIX-014-cursor-pagination-1590554316` |
| NF-FIX-015-graph-adapter-input | 4 | `bdd09489ddada650316702a0bbb8060ec761734ce4b0cf8d67bb90a2ded1fc79` | `NF-FIX-015-graph-adapter-input-363852864` |
| NF-FIX-016-redaction | 9 | `707ae16b1986902e39fcb83c5602281a721c906f5fd22f59093715e9a42b9328` | `NF-FIX-016-redaction-2255581443` |
| NF-FIX-017-indicator-link-mismatch | 3 | `043b81f172a532312c0d5a276c5101f836f7c425b294d4925099c4008c194ca5` | `NF-FIX-017-indicator-link-mismatch-1042493670` |
| NF-FIX-018-resource-limits | 5 | `ab81f596972992948f53c434ec1a7aba5a3530215fa4f171a369924ef17b94b8` | `NF-FIX-018-resource-limits-3975301562` |
| NF-FIX-019-canonical-json-unicode | 5 | `69ceebd5a51bce7de6ad2b19a92187a0bf582cf8a7c5e8589d72d37c322208d4` | `NF-FIX-019-canonical-json-unicode-2603325265` |
| NF-FIX-020-atomic-import-commit | 3 | `292e7ae738b7b2a58f853610bd79ef7347b0ec7a1562cb49ff14cbf9384f1dfa` | `NF-FIX-020-atomic-import-commit-2534207203` |
| NF-FIX-021-preview-boundaries | 3 | `d837da92e3e41f4263ddada780c0a9f2871916070493217c4799072420fb7fb7` | `NF-FIX-021-preview-boundaries-812623438` |
| NF-FIX-022-timestamp-rulesets | 4 | `75f4dea266a516e9b3e697c80d020b87c2d6505634b757b7305793ef54d3b6bd` | `NF-FIX-022-timestamp-rulesets-198777445` |
| NF-FIX-023-import-facade-source-change | 3 | `539e97258c94a6f01bd54caae7a8c323ed6ac03dcd14eaa1c9b23f32d986d555` | `NF-FIX-023-import-facade-source-change-964012810` |
| NF-FIX-024-query-normalization-cursors | 4 | `b58b8d5d9dd26222149f309cf4641d4878df7ce6936c6b8d72075cc9c9fbda10` | `NF-FIX-024-query-normalization-cursors-3913051213` |
| NF-FIX-025-graph-contributors | 3 | `ab99bf1b67599a68e6e9c7ada35ddaba792bd25306039ca9ae060e36601bd73f` | `NF-FIX-025-graph-contributors-2825074312` |
| NF-FIX-026-audit-and-replay | 5 | `4879b1624be9bb7d0e422c2ffcf3930e69c85fd9706681a41f66d7391f080782` | `NF-FIX-026-audit-and-replay-941390242` |
| NF-FIX-027-retention-soft-delete | 4 | `c6665c6e121fd898fa8efe2050dfc3a292f6821ac9db1f9bf74487df8c74d698` | `NF-FIX-027-retention-soft-delete-546845942` |
| NF-FIX-028-graph-aggregate-bounds | 3 | `d40a5d3c30854f291cc3e5512b9e63153d4588364e64f8b4207c42380209551e` | `NF-FIX-028-graph-aggregate-bounds-2830791121` |

`nf-imp08-s33f-browser-20261030-01` selects
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.import_recovery`.
Its result and the closing documentation checkpoint remain required before S33-F
can close F38/F40 and release S34. No frozen bytes changed during corpus closure.


The browser check failed native 2, cleanup 23/23, at the same `8e725f…` source.
Its exact-root observer `nf-imp08-s33f-browser-observe-20261030-01` passed native 0.
The retained Playwright report under
`browser-e2e-webserver-backed/browser-groups/cohort-175b9060a6112fc0e67622992c2f577f0dc174279bfea6100b40bc3d5a5885a5/`
reports `SyntaxError: Unexpected token 'export'` while `createRequire` loads the
existing strict-JSON ES module. No browser test or product comparison executed.
S33-F's new browser integration exposed a loader assumption in the Node fixture
helper. Replace that require with native asynchronous ES-module loading, propagating
asynchrony through admission and its private decoder consumers. Keep the existing
strict parser, schemas and callback admission order. This is fixture-support repair,
with no product requirement, migration or public-interface change. Node controls,
client comparison, types and browser execution must pass after the repair.


The loader repair passed `nf-imp08-s33f-format-20261030-04`, native 0. At source
`7d2f9076b0b0b107caf1ca6a8178c8534f8eaccba9aba6f6bd9da70d830d0e38`,
`nf-imp08-s33f-types-20261030-03`, `nf-imp08-s33f-client-20261030-02`, and
`nf-imp08-s33f-node-20261030-02` passed native 0, each with 11/11 cleanup completed,
using the same respective selections above. Browser run 02 uses the unchanged exact
import-recovery selection and remains pending. Earlier `8e725f…` corpus evidence is
retained as phase history; final corpus closure will use compatible source evidence.


Browser run 02 passed native 0, cleanup 23/23, source `7d2f9076…` above. Its admitted
001 revision-8 comparison is retained at
`NF-FIX-001-cisco-sna-minimal-node-JCMqrD/comparison-summary.json`; the listed
`produced/browser-comparison.json` records the passing import-recovery test.
Manifest SHA-256 is `d16582b1b7a79b62f05a7794873af5954efc8432449b2dc5b005cf4ebfade648`.

To close at the repaired source, corpus/generated/catalog runs 02 use the exact
respective commands from run 01. This is the distinct corpus-closure qualification;
earlier source evidence remains historical. No service orchestration or product
interpreter was added to fixture support. S33-F remains pending these exits and
its closing Markdown checkpoint.


Corpus run 02 passed native 0 (14/14 cleanup), but the retained-evidence audit found
only the 20 integration comparison directories. The harness accepted cached unit-row
results (unit exits null; row durations zero), so the eight unit-family callbacks did
not execute or publish new fixture summaries. `nf-imp08-s33f-corpus-observe-20261030-02`
used the exact root with `TARGET=test-slice DETAIL=accounting`, native 0. This is a
cache/evidence distinction, not a failed product comparison. Do not infer fresh
materialization from a cached row pass.

`nf-imp08-s33f-unit-evidence-20261030-01` selects precisely the ten unit rows from
the recorded 30-row corpus command (eight families, integrity, support controls),
with `CARTULARY_HARNESS_CACHE_MODE=off` and the established environment. The 20 current
integration comparisons will be combined with these missing same-source outputs;
no integration rerun is required. S36's fixture-comparison qualification must likewise
disable execution-result reuse unless retained comparison artifacts are explicitly
available through admitted evidence. This uses the existing public cache control,
without changing global harness policy or weakening comparison requirements.

Closing Markdown run 01 passed native 0, but its attempted closing-record preparation
stopped at the missing-summary assertion; it did not mark S33-F complete. A new
closing checkpoint follows completion of this evidence obligation.


### 27.58 S33-F — closing evidence

The missing unit evidence passed native 0. Its exact command, with the established
GO/preparation prefix, was:

```sh
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352
```

At source `7d2f9076b0b0b107caf1ca6a8178c8534f8eaccba9aba6f6bd9da70d830d0e38`,
the closing corpus and source checks retain these completed cleanups:

| Run ID | Native exit | Cleanup |
| --- | --- | --- |
| `nf-imp08-s33f-corpus-20261030-02` | 0 | 14/14 |
| `nf-imp08-s33f-unit-evidence-20261030-01` | 0 | 11/11 |
| `nf-imp08-s33f-generated-20261030-02` | 0 | 14/14 |
| `nf-imp08-s33f-catalog-20261030-02` | 0 | 11/11 |

Exactly 28 unique passing summaries equal the directory inventory; all manifest
hashes/revisions match the preceding table. Each directory below contains
`comparison-summary.json`. Root **I** is
`.cartulary/test-results/nf-imp08-s33f-corpus-20261030-02/`; root **U** is
`.cartulary/test-results/nf-imp08-s33f-unit-evidence-20261030-01/`.

| Bundle | Root | Current evidence directory |
| --- | --- | --- |
| NF-FIX-001-cisco-sna-minimal | I | `NF-FIX-001-cisco-sna-minimal-936266731` |
| NF-FIX-002-cisco-sna-interface-fields | I | `NF-FIX-002-cisco-sna-interface-fields-3542136572` |
| NF-FIX-003-duplicate-headers | U | `NF-FIX-003-duplicate-headers-1927152433` |
| NF-FIX-004-rejected-rows | I | `NF-FIX-004-rejected-rows-3438640018` |
| NF-FIX-005-csv-parser-edges | U | `NF-FIX-005-csv-parser-edges-2713798389` |
| NF-FIX-006-cross-table-graph | I | `NF-FIX-006-cross-table-graph-2419823218` |
| NF-FIX-007-indicator-linking | I | `NF-FIX-007-indicator-linking-3581609821` |
| NF-FIX-008-large-limits | I | `NF-FIX-008-large-limits-3438277402` |
| NF-FIX-009-soft-delete-stale-graph | I | `NF-FIX-009-soft-delete-stale-graph-3900639733` |
| NF-FIX-010-json-admission | U | `NF-FIX-010-json-admission-4014334719` |
| NF-FIX-011-alias-collision | U | `NF-FIX-011-alias-collision-926714892` |
| NF-FIX-012-sys-uptime-timestamps | U | `NF-FIX-012-sys-uptime-timestamps-2600050274` |
| NF-FIX-013-filename-display | I | `NF-FIX-013-filename-display-3071834910` |
| NF-FIX-014-cursor-pagination | I | `NF-FIX-014-cursor-pagination-1982644443` |
| NF-FIX-015-graph-adapter-input | I | `NF-FIX-015-graph-adapter-input-3835913172` |
| NF-FIX-016-redaction | I | `NF-FIX-016-redaction-1898632762` |
| NF-FIX-017-indicator-link-mismatch | I | `NF-FIX-017-indicator-link-mismatch-901398912` |
| NF-FIX-018-resource-limits | I | `NF-FIX-018-resource-limits-1408619570` |
| NF-FIX-019-canonical-json-unicode | U | `NF-FIX-019-canonical-json-unicode-456845012` |
| NF-FIX-020-atomic-import-commit | I | `NF-FIX-020-atomic-import-commit-1799704478` |
| NF-FIX-021-preview-boundaries | U | `NF-FIX-021-preview-boundaries-2483651197` |
| NF-FIX-022-timestamp-rulesets | U | `NF-FIX-022-timestamp-rulesets-3002834024` |
| NF-FIX-023-import-facade-source-change | I | `NF-FIX-023-import-facade-source-change-2250014226` |
| NF-FIX-024-query-normalization-cursors | I | `NF-FIX-024-query-normalization-cursors-3739003977` |
| NF-FIX-025-graph-contributors | I | `NF-FIX-025-graph-contributors-1929942113` |
| NF-FIX-026-audit-and-replay | I | `NF-FIX-026-audit-and-replay-1120017451` |
| NF-FIX-027-retention-soft-delete | I | `NF-FIX-027-retention-soft-delete-3867789223` |
| NF-FIX-028-graph-aggregate-bounds | I | `NF-FIX-028-graph-aggregate-bounds-494649214` |

Together with same-source types 03, client 02, Node controls/ownership 02, and browser
02, the executable corpus, admission, immutable materialization, actual retained
comparisons and wrong-expectation exits pass. F38/F40's implementation exits are
satisfied; F39's corrected owner reference and actual v2 adapter agree. S33's aggregate
exit is satisfied pending `nf-imp08-s33f-close-20261030-02 make lint-markdown`.

Remaining risks are open F36/S34, F37/S35 and final S36 qualification. S33-F adds no
production compatibility or migration impact. Migration 34 and all prior correction
compatibility decisions remain required. Source/test reversion is this workstream's
rollback; historical frozen revisions remain in Git. These phase passes do not claim
S36 acceptance. Record the closing checkpoint and successor readiness before S34.


`nf-imp08-s33f-close-20261030-02 make lint-markdown` passed native 0; summary
`adhoc/lint-markdown/tool-run-summary.json`. S33-F and the S33 aggregate are DONE.
F38/F39/F40 are resolved at their implementation/workstream exits, with S36 final
qualification still required. S34 is now IN_PROGRESS. This checkpoint precedes its
code changes. F36 will remove materialized apply accounting and accepted-row
collection from preview while preserving decoder boundaries and meaningful canonical
encoding error propagation introduced by S39. V30 and affected fixtures are its exit.


### 27.59 S34 — preview-only accounting and streaming assertions

Replaced `validateRows`/`rowValidationResult` with `validatePreview`: accepted and
rejected counts plus complete diagnostics, with no accepted-row collection, retention
sample, diagnostic-total field, truncation state or parse-mode branch. The facade
emits `diagnostics_truncated=false`. Removed the parsed container's mode field;
production decoder preview/apply modes remain intact.

S39 made normalization/canonical encoding errors meaningful. Therefore the earlier
proposal to remove an unused error return is superseded: `validatePreview` propagates
those real failures. Discarding or substituting them would reintroduce F41.

Moved retention, exact diagnostic-prefix agreement and all-rejected sampling onto
`prepareImportApply` and its streaming visitor. The all-rejected assertion now uses
an actually all-rejected source and the production owner error, replacing the former
fabricated error from a mixed source. Existing import characterization helpers now
exercise production prepare/visit policy; the graph unit vector uses per-record
normalization without implementing apply accounting. The test-only full decoder
collector remains a decoder helper, with no validation policy.

Format 01 and coverage task-guide 01 passed native 0. No selectors or generated
contracts changed. The focused V30/affected unit command is
`nf-imp08-s34-unit-20261030-01`, with the established environment and cache disabled
for retained fixture comparisons:

```sh
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c,module.networkflow.unit.network_flow_selector_covers_import_apply_replay_8b8c7de503,module.networkflow.unit.network_flow_selector_covers_import_facade_previ_2371e1f671,module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352
```

This preserves the exact V30 rows and adds their changed graph/import callers and
the eight parsing/normalization fixture comparisons. Publication, rejected-row and
privacy fixture regressions follow. No digest, schema, public protocol or migration
change is intended; complete preview diagnostics and streaming retention remain
required. S34 remains IN_PROGRESS until these checks and its Markdown exit pass.


The first V30 run failed native 2, cleanup 10/10, source
`3883fed956cb95b8df9ca49c3e4abbf832e546613ad2994be3a0bf5723f33dda`.
Only the new preview/apply agreement subcase failed: its independently constructed
mapping omitted materialized source-field `combinability`, producing a different
mapping fingerprint and therefore different diagnostic identities. Counts and sample
retention were correct. The test now passes the exact approved preview mapping and
fingerprint to apply, as Imports does; it retains complete diagnostic equality.
No production expectation was weakened and no new runtime defect was found.
The exact-root unit observer passed native 0 with `TARGET=test-slice DETAIL=summary`.

At that source, `nf-imp08-s34-fixtures-20261030-01` passed native 0 with cache disabled,
selecting the registered 001/002/004/016/020/023 comparison rows; generated 01
(`make source-check CHECK=generated`) and boundary 01
(`make lint CHECK=backend-boundary`) also passed native 0. Their evidence is retained
as phase history while the corrected test setup is checked.


The corrected V30 run 02 passed native 0. Prior fixtures/generated/boundary cleanup
was respectively 14/14, 14/14 and 13/13. Final review also separated the apply
all-rejected sample bound from the preview-record bound: apply validation and owner
error publication share `allRejectedDiagnosticSampleLimit=50`; preview retains its
own 50-record decoder limit. This removes accidental policy coupling without changing
values or behavior. Existing exact sample/retention tests cover the separation.
Final S34 evidence will combine the 17 unit rows and six publication/privacy fixture
rows at the resulting source; cache is disabled to retain all comparison artifacts.


Format 03 passed native 0. Closing S34 selection
`nf-imp08-s34-validation-20261030-01`, with the established environment:

```sh
CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.atomic_fixture_compares_every_commit_boundary_ro_a4cdb920d4,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.source_changed_fixture_compares_every_source_ver_213b5f7eab,module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c,module.networkflow.unit.network_flow_selector_covers_import_apply_replay_8b8c7de503,module.networkflow.unit.network_flow_selector_covers_import_facade_previ_2371e1f671,module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352
```

`nf-imp08-s34-generated-20261030-02` and `nf-imp08-s34-boundary-20261030-02`
use the same generated/boundary commands above at this source. All remain pending.


### 27.60 S34 — closing evidence

All closing commands passed native 0 with unchanged HEAD/toolchain/system identities.
Exact selections are recorded above; current source and cleanup are:

| Run ID | Source digest | Cleanup |
| --- | --- | --- |
| `nf-imp08-s34-validation-20261030-01` | `sha256:c8548ccb858c66518aac05e9096465759ee1b9c4b3ac6fdc245d04ac17a4ee72` | 15/15 |
| `nf-imp08-s34-generated-20261030-02` | `sha256:c8548ccb858c66518aac05e9096465759ee1b9c4b3ac6fdc245d04ac17a4ee72` | 14/14 |
| `nf-imp08-s34-boundary-20261030-02` | `sha256:c8548ccb858c66518aac05e9096465759ee1b9c4b3ac6fdc245d04ac17a4ee72` | 13/13 |

The intermediate unit run 02 passed at `sha256:05128b45baf9e5294ff5d3af210c0b15b50b913892fb22965a05165c6a9e8a4c`, cleanup 10/10; it is superseded by closing validation.

All 14 affected fixture comparisons retained passing `comparison-summary.json` under
these directories in the closing validation root:

- `NF-FIX-001-cisco-sna-minimal-3034795406`.
- `NF-FIX-002-cisco-sna-interface-fields-2418383732`.
- `NF-FIX-003-duplicate-headers-3863312733`.
- `NF-FIX-004-rejected-rows-2118905210`.
- `NF-FIX-005-csv-parser-edges-1583888077`.
- `NF-FIX-010-json-admission-930187915`.
- `NF-FIX-011-alias-collision-4024466572`.
- `NF-FIX-012-sys-uptime-timestamps-2987876471`.
- `NF-FIX-016-redaction-1799221333`.
- `NF-FIX-019-canonical-json-unicode-3517811964`.
- `NF-FIX-020-atomic-import-commit-2704065089`.
- `NF-FIX-021-preview-boundaries-3382944999`.
- `NF-FIX-022-timestamp-rulesets-224146972`.
- `NF-FIX-023-import-facade-source-change-655877277`.

V30 preserves 50/51 boundaries, malformed/raw tails, complete private-safe preview
diagnostics, zero/exact/overflow retention, multiplicity, actual all-rejected samples,
large bounded processing, source verification and late failures. The retired validator,
materialized apply policy and parsed mode field are absent. The test-only decoder
collector has no apply policy. No public bytes, durable schema, identity, state version
or migration changed; canonical errors still propagate. F36's implementation exit is
satisfied pending `nf-imp08-s34-close-20261030-01 make lint-markdown`.

Rollback is source/test reversion. F37's identity/provenance write-boundary risk remains
open, and all phase evidence remains distinct from S36 final qualification. S35 may
begin only after the documentation checkpoint and successor-status update.


`nf-imp08-s34-close-20261030-01 make lint-markdown` passed native 0;
summary `adhoc/lint-markdown/tool-run-summary.json`. S34/F36 are DONE at the workstream
exit. S35/F37 is now IN_PROGRESS; this update precedes its changes. It will restrict
normalization/streaming write inputs to semantic values, locators and digests, require
validated import context before writes, derive persisted identity/provenance once,
and preserve Imports' borrowed-transaction authority. No additional migration or
digest change is planned beyond the already-required migration 34. S36 remains open.


### 27.61 S35 — canonical persisted-row ownership

F37's implementation and V31 exits now pass. The private `normalizedRow` contains
semantic fields, source locator and source/normalized digests only. Normalization,
streaming visitors and the storage fixture writer accept this type; `flowRow`
remains the graph/query/read model. Incoming rows cannot supply table/incident/row
identity, mapping fingerprint, attribution or observation provenance.

`validatedRowImportContext` rejects missing UUIDs, malformed source/mapping/safe
filename digests, invalid digest-key identity, unavailable source profiles,
missing/mismatched approved parser profile and missing/out-of-range import time
before transaction access. Profile admission uses the authored mapping registry.
There is no storage parser fallback. One constructor derives row identity,
context mapping fingerprint, attribution and all eight provenance members
unconditionally; invalid locators, digests, table IDs or retained raw objects
return errors before their row insert. No empty-provenance substitution remains.
The store returns failures to Imports, which retains transaction authority.

Persistence fixtures now provide normalized inputs and derive expected IDs from
an independent NF-040 framed stream. The obsolete suffix argument and arbitrary
ID overrides are removed. Existing broader nullable storage/read-model scenarios
remain separate from claimable CSV normalization, whose complete contracts are
compared by the 28 admitted fixtures. Fixed independent row-ID vectors, duplicate
semantic values at different locators, exact provenance/nullability, year-0001
flow values and malformed contexts/constructions extend V31-A. V31-B verifies
that a malformed second row rolls back table, rows, diagnostics, audit and
idempotency effects; a subsequent valid import succeeds without a reserved name.

The first construction portfolio exposed two test assumptions: fabricated ID
suffixes in graph iteration and missing explicit time in a randomness constraint
fixture. Both fixtures now use independently derived identities/explicit import
context. This was test migration within S35, not a new owner defect. The original
failure remains retained; its constructor/rollback subtests already passed.

No public API, SQL column, canonical digest, durable-state version or additional
migration changes in S35. The already-authored migration 34 remains required for
F41 and related adopted corrections. Rollback for this slice is source/test
reversion. The remaining risk is final portfolio compatibility after preparation;
these phase passes do not claim S36 qualification.

All commands below use the recorded pinned GO and installed-only environment.
Their exact roots are `.cartulary/test-results/<run-id>/`, with `run-manifest.json`,
`run-summary.json`, `cleanup-results.json` and target/native summaries. Native
exits were observed separately. Test selections used cache mode `off` so that
behavioral summaries were produced. Toolchain/system identities remain those in
§27.1. Every cleanup result listed below is `completed`, none failed/blocked.

| Run ID / native exit | Exact Make command (plus cache mode `off` for test slices) | Source digest | Cleanup |
| --- | --- | --- | --- |
| `nf-imp08-s35-unit-20261030-01`; exit 0 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a` | `sha256:7ec61d9d50a8de5ea71aa0813b3fd1a6e657bad250bb583b9d47ff1e38c2f244` | 10/10 |
| `nf-imp08-s35-construction-20261030-01`; exit 2 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.integration.pagination_recovery,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a` | `sha256:7e1998ba3d1249e704f3f79d5afb3623824e27e8780ed4e5aac6c013b500678b` | 21/21 |
| `nf-imp08-s35-construction-20261030-02`; exit 0 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.integration.pagination_recovery,module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a` | `sha256:339445a0719b9200c65fb0aef1f37ed585ca136168a5995787d9909fbe8bcf8e` | 21/21 |
| `nf-imp08-s35-imports-20261030-01`; exit 0 | `make test-slice OWNER=module.imports ROWS=module.imports.integration.current_cancellation,module.imports.integration.current_workflows,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_default_limit_100001_rows,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6` | `sha256:339445a0719b9200c65fb0aef1f37ed585ca136168a5995787d9909fbe8bcf8e` | 23/23 |
| `nf-imp08-s35-corpus-20261030-01`; exit 0 | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.aggregate_bounds_fixture_compares_legal_source_c_7934db1cd2,module.networkflow.integration.atomic_fixture_compares_every_commit_boundary_ro_a4cdb920d4,module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b,module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2,module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9,module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.large_limits_fixture_compares_valid_imported_row_470de5552b,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f,module.networkflow.integration.retention_fixture_compares_soft_delete_incident_570ae9035a,module.networkflow.integration.source_changed_fixture_compares_every_source_ver_213b5f7eab,module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49,module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352` | `sha256:339445a0719b9200c65fb0aef1f37ed585ca136168a5995787d9909fbe8bcf8e` | 16/16 |
| `nf-imp08-s35-boundary-20261030-01`; exit 0 | `make lint CHECK=backend-boundary` | `sha256:339445a0719b9200c65fb0aef1f37ed585ca136168a5995787d9909fbe8bcf8e` | 13/13 |
| `nf-imp08-s35-generated-20261030-01`; exit 0 | `make source-check CHECK=generated` | `sha256:339445a0719b9200c65fb0aef1f37ed585ca136168a5995787d9909fbe8bcf8e` | 14/14 |

The closing substantive source is
`sha256:339445a0719b9200c65fb0aef1f37ed585ca136168a5995787d9909fbe8bcf8e`.
V31-C/D passed all seven exact Imports rows, including 100,001 records, safe
owner errors, public atomic publication, cancellation and crash/replay. The
closing corpus run also passes V30's exact five rows, integrity and admission
controls, and independently emits all 28 passing fixture comparisons:

| Fixture / revision | Exact summary path beneath `nf-imp08-s35-corpus-20261030-01/` |
| --- | --- |
| `NF-FIX-001-cisco-sna-minimal` / 8 | `NF-FIX-001-cisco-sna-minimal-421228271/comparison-summary.json` |
| `NF-FIX-002-cisco-sna-interface-fields` / 5 | `NF-FIX-002-cisco-sna-interface-fields-3078976659/comparison-summary.json` |
| `NF-FIX-003-duplicate-headers` / 4 | `NF-FIX-003-duplicate-headers-2286645567/comparison-summary.json` |
| `NF-FIX-004-rejected-rows` / 6 | `NF-FIX-004-rejected-rows-3378186231/comparison-summary.json` |
| `NF-FIX-005-csv-parser-edges` / 3 | `NF-FIX-005-csv-parser-edges-3316446891/comparison-summary.json` |
| `NF-FIX-006-cross-table-graph` / 3 | `NF-FIX-006-cross-table-graph-1784085704/comparison-summary.json` |
| `NF-FIX-007-indicator-linking` / 4 | `NF-FIX-007-indicator-linking-378789972/comparison-summary.json` |
| `NF-FIX-008-large-limits` / 3 | `NF-FIX-008-large-limits-1705925416/comparison-summary.json` |
| `NF-FIX-009-soft-delete-stale-graph` / 3 | `NF-FIX-009-soft-delete-stale-graph-145803068/comparison-summary.json` |
| `NF-FIX-010-json-admission` / 4 | `NF-FIX-010-json-admission-4187545884/comparison-summary.json` |
| `NF-FIX-011-alias-collision` / 4 | `NF-FIX-011-alias-collision-233583607/comparison-summary.json` |
| `NF-FIX-012-sys-uptime-timestamps` / 3 | `NF-FIX-012-sys-uptime-timestamps-1933185651/comparison-summary.json` |
| `NF-FIX-013-filename-display` / 4 | `NF-FIX-013-filename-display-134632266/comparison-summary.json` |
| `NF-FIX-014-cursor-pagination` / 5 | `NF-FIX-014-cursor-pagination-4239031459/comparison-summary.json` |
| `NF-FIX-015-graph-adapter-input` / 4 | `NF-FIX-015-graph-adapter-input-2196991294/comparison-summary.json` |
| `NF-FIX-016-redaction` / 9 | `NF-FIX-016-redaction-548845573/comparison-summary.json` |
| `NF-FIX-017-indicator-link-mismatch` / 3 | `NF-FIX-017-indicator-link-mismatch-2147602737/comparison-summary.json` |
| `NF-FIX-018-resource-limits` / 5 | `NF-FIX-018-resource-limits-3742929210/comparison-summary.json` |
| `NF-FIX-019-canonical-json-unicode` / 5 | `NF-FIX-019-canonical-json-unicode-1016926087/comparison-summary.json` |
| `NF-FIX-020-atomic-import-commit` / 3 | `NF-FIX-020-atomic-import-commit-486645510/comparison-summary.json` |
| `NF-FIX-021-preview-boundaries` / 3 | `NF-FIX-021-preview-boundaries-3482406769/comparison-summary.json` |
| `NF-FIX-022-timestamp-rulesets` / 4 | `NF-FIX-022-timestamp-rulesets-3118201820/comparison-summary.json` |
| `NF-FIX-023-import-facade-source-change` / 3 | `NF-FIX-023-import-facade-source-change-2480687519/comparison-summary.json` |
| `NF-FIX-024-query-normalization-cursors` / 4 | `NF-FIX-024-query-normalization-cursors-1882879445/comparison-summary.json` |
| `NF-FIX-025-graph-contributors` / 3 | `NF-FIX-025-graph-contributors-2168936415/comparison-summary.json` |
| `NF-FIX-026-audit-and-replay` / 5 | `NF-FIX-026-audit-and-replay-3370122815/comparison-summary.json` |
| `NF-FIX-027-retention-soft-delete` / 4 | `NF-FIX-027-retention-soft-delete-2848509671/comparison-summary.json` |
| `NF-FIX-028-graph-aggregate-bounds` / 3 | `NF-FIX-028-graph-aggregate-bounds-263993250/comparison-summary.json` |

Each summary retains the manifest and source/expected bundle hashes, materialized
inputs, produced artifacts and comparison outcome. Frozen revisions are unchanged
from S33-F. Supporting preparation was `make format` under
`nf-imp08-s35-format-20261030-01`, `-02`, `-03` (native 0 each). Coverage discovery
used `make task-guide OWNER=module.networkflow DETAIL=coverage JSON=1` and
`OWNER=module.imports` under `nf-imp08-s35-task-guide-20261030-01` and
`nf-imp08-s35-import-guide-20261030-01`; their exact JSON is retained locally in
`/tmp/nf-s35-task-guide.json` and `/tmp/nf-s35-import-guide.json`. No new selector
was needed: constructor and rollback cases extend the exact V31-A/B rows.

S35 completion is pending the documentation checkpoint
`nf-imp08-s35-close-20261030-02 make lint-markdown`. Only after its native pass and
status update may S36 finalize, inspect generated changes, freeze source and
qualify the complete affected portfolio. No required runtime correction is open.

The initial local ledger attempts stopped on evidence-field assumptions and an authoring-script syntax error; no tracker changes were written. `nf-imp08-s35-close-20261030-01` passed native 0 against the preceding document and is not this checkpoint. Run 02 validates the corrected ledger. Product evidence was unchanged.


`nf-imp08-s35-close-20261030-02 make lint-markdown` passed native 0;
summary `adhoc/lint-markdown/tool-run-summary.json`. S35/F37 is DONE at its
workstream exit. This status update precedes S36 work. All implementation and
correction workstreams now have passing exits; final source preparation and the
agreed affected-portfolio qualification remain outstanding.

### 27.62 S36 — finalization and frozen-source qualification

S36 is IN_PROGRESS. Preparation will remove remaining unused test helpers, run
`make agent-finalize`, inspect its changes and freeze source. Final coverage is
the exact V29–V33 selections plus the adopted V34–V50 correction obligations,
all 28 comparisons and support/schema/routing checks. Behavioral fixture runs
will disable the result cache to retain actual comparison summaries. This is an
affected portfolio; complete-owner, release and deployment qualification remain
separate. Migration 34 requires the operational handoff already recorded by S39
and its later registry refinements. Final source/evidence/cleanup reconciliation
and the final Markdown checkpoint must pass before S36 becomes DONE.


`nf-imp08-s36-finalize-20261030-01 make agent-finalize` passed native 0.
Its `unit-artifacts/finalize-summary.json` reports generated structure unchanged,
zero updated files and no mutation rollback. Reviewed preparation therefore
introduces no generated-byte delta; the only final cleanup removes the unused
`rowIDSuffix` test helper. Schema/catalog/tier preparation passed. Retained-run
maintenance was skipped because no compatible successful full warm-check root
was supplied; no redundant full check was launched to obtain one.

Source is now frozen at **`sha256:60ad5add6add321e704cc6e75bf90e86f863ee80991c67844966a1585a413c76`**, dirty `main` over
`98adcc82961456bcb74bbecd0023136485afbf00`, with the pinned toolchain/system in
§27.1. Current `make task-guide OWNER=<owner> DETAIL=coverage JSON=1` discovery
confirmed every selected row for six owners (`nf-imp08-s36-guide-1-20261030-01`
through `-6-20261030-01`, all native 0). The affected portfolio comprises 70
Network Flow rows (62 backend and 8 browser/stateful workflows), seven Imports,
two migration-policy/parity, five Network Flow client, two web admission/ownership
and two native evidence-schema/session rows. It includes all exact V29–V33 and
V34–V50 obligations and all 28 fixture comparisons. This is 88 selected rows,
not complete-owner acceptance. The remaining non-row checks are generated,
schema, catalog, Go/types/backend-boundary lint and final Markdown.

Final fixture comparisons require actual execution (`CARTULARY_HARNESS_CACHE_MODE=off`),
retained input/expected hashes and comparison summaries. Earlier phase results
remain historical because this frozen source differs. All terminal runs must
retain successful cleanup and native exits before final reconciliation.

The initial backend target-plan request (`nf-imp08-s36-plan-backend-20261030-01`) was rejected before execution because its exact ROWS string exceeded the 4096-character input bound. The unchanged 70-row Network Flow selection is split into unit/store, integration and browser slices. No tests ran in that rejected request and coverage was neither widened nor reduced.


### 27.63 S36 — validation cleanup before final qualification

The first frozen-source unit/store run (`nf-imp08-s36-unit-20261030-01`, 37 rows)
and integration run (`nf-imp08-s36-integration-20261030-01`, 25 rows) passed native
0. The initial Go lint run (`nf-imp08-s36-go-20261030-01`) failed native 2 because
staticcheck's nested `go` process resolved PATH's Go 1.27.1 despite the explicit
pinned GO executable. This is an execution-environment mismatch, not a product
failure; the harness retains its normalized `product/test_assertion_failure`
classification. Its 15 cleanup steps completed. Subsequent commands prepend the
pinned toolchain's `bin` directory to PATH as well as setting GO.

Go lint run 02 reached analysis and failed native 2 on four unused test helpers
in `canonical_success_test.go` and SA4000 for `if <-paths == <-paths` in fixture
support. The channel expression receives two distinct values, but naming those
values makes the isolation assertion explicit and avoids the analyzer ambiguity.
Its 15 cleanup steps completed. The four obsolete wrappers have no callers.
This is S36 test cleanup, not a new runtime defect or owner-contract change.

Current independent Imports/browser executions will drain before source edits.
Then remove those wrappers, name the received paths, run formatting and Go lint,
repeat finalization, and establish a new source freeze. Earlier S36 passes remain
retained intermediate evidence; final qualification will use only compatible
runs admitted by the harness. No expectation, row selection, frozen fixture byte,
production behavior, migration or public interface changes in this cleanup.

The independent first-freeze Imports and browser runs both passed native 0 at
`sha256:60ad5add6add321e704cc6e75bf90e86f863ee80991c67844966a1585a413c76`,
with respectively 23/23 and 24/24 completed cleanup steps. Unit/store and integration
cleanup completed 17/17 and 20/20. All first-freeze tests passed; only the documented
static-analysis cleanup changed source. The four wrappers are now removed and the
concurrency assertion compares named received paths. `nf-imp08-s36-format-20261030-01`
passed native 0, followed by `nf-imp08-s36-go-20261030-03 make lint CHECK=go` (native 0)
at `sha256:2983c43dc1630a1056faeb39e2fba0ef0071043b22c071bd174c5a6264ca0776`. The repeated finalizer is
`nf-imp08-s36-finalize-20261030-02 make agent-finalize`; no product semantics or
fixture revisions changed.

Repeated finalization passed native 0, again reporting generated structure unchanged
and zero updated files. The final freeze is now **`sha256:2983c43dc1630a1056faeb39e2fba0ef0071043b22c071bd174c5a6264ca0776`**.
Go lint run 03 is already compatible with that source. The same exact 88-row
portfolio is being qualified; fixture and browser/client comparisons disable
cache reuse, while non-fixture obligations may use harness-admitted normal cache
mode. First-freeze roots remain historical and are not manually treated as final
evidence. All subsequent commands use the pinned GO and PATH prefix.


### 27.64 S36 — final affected-portfolio evidence and operational handoff

All 88 selected rows and every required non-row product check passed native 0
at final source **`sha256:2983c43dc1630a1056faeb39e2fba0ef0071043b22c071bd174c5a6264ca0776`**. HEAD remains
`98adcc82961456bcb74bbecd0023136485afbf00` on dirty `main`; no commit or deployment
was performed. The source includes the four dead-wrapper removals and explicit
channel-receive comparison from §27.63. Both finalizers reported zero generated
changes; the second finalizer and all final roots agree on this source.

Toolchain identity is
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`;
system identity is
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
The exact command prefix is:

```sh
PATH=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin:$PATH GO=/home/jochi/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.3.linux-amd64/bin/go CARTULARY_PREPARATION_POLICY=installed_only
```

Add `CARTULARY_TEST_RUN_ID=<literal Run ID below>` to that prefix. Each exact
root is `/home/jochi/code/cartulary/.cartulary/test-results/<Run ID>/`. Each retains
`run-manifest.json`, `run-summary.json`, `cleanup-results.json`, its
`target-summaries/<target>.json` and native unit artifacts. Native Make exits were
observed independently of summary publication. Every cleanup step below is
completed, with zero failed/blocked steps. The table records the actual admitted
row selection, not proposed selectors or an inferred complete-owner union.

| Run ID / native exit | Exact command after prefix and run ID | Obligation | Cleanup |
| --- | --- | --- | --- |
| `nf-imp08-s36-unit-20261030-02` / 0 | `CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.networkflow ROWS=module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.configuration_claim_and_manifest_admission,module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.import_owner_error_translation_rs08_3d3f02e050,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_a_table_query_uses_6caf43b23c,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_duplicate_csv_heade_61bc2b1f54,module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.network_flow_selector_covers_every_network_flow_43bd8f4992,module.networkflow.unit.network_flow_selector_covers_import_apply_replay_8b8c7de503,module.networkflow.unit.network_flow_selector_covers_import_facade_previ_2371e1f671,module.networkflow.unit.network_flow_selector_covers_interface_fields_st_6333a351b3,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.unit.network_flow_selector_covers_logs_telemetry_audi_61f27d1e10,module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.network_flow_selector_covers_timestamp_profile_p_bdf924da13,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.table_lifecycle_contract,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.unit.unicode17_text_canonicalization_3de608c3dc,module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352` | V29-A/B, V30, V31-A/B, affected unit/store corrections | 17/17 |
| `nf-imp08-s36-integration-20261030-02` / 0 | `CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.aggregate_bounds_fixture_compares_legal_source_c_7934db1cd2,module.networkflow.integration.atomic_fixture_compares_every_commit_boundary_ro_a4cdb920d4,module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2,module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9,module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.large_limits_fixture_compares_valid_imported_row_470de5552b,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d,module.networkflow.integration.pagination_recovery,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f,module.networkflow.integration.retention_fixture_compares_soft_delete_incident_570ae9035a,module.networkflow.integration.source_changed_fixture_compares_every_source_ver_213b5f7eab,module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49,module.networkflow.integration.table_lifecycle_admission_replay` | V29-B, V31-B, canonical migration and query/graph/link corrections | 20/20 |
| `nf-imp08-s36-imports-20261030-02` / 0 | `CARTULARY_HARNESS_CACHE_MODE=normal make test-slice OWNER=module.imports ROWS=module.imports.integration.current_cancellation,module.imports.integration.current_workflows,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_default_limit_100001_rows,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6` | V31-C/D | 23/23 |
| `nf-imp08-s36-browser-20261030-02` / 0 | `CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.import_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_v_171429ad9e,module.networkflow.browser.query_authoring_integrity,module.networkflow.browser.table_lifecycle_continuity,module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle,module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf` | V32-A and all affected correction workflows | 24/24 |
| `nf-imp08-s36-ddl-20261030-01` / 0 | `CARTULARY_HARNESS_CACHE_MODE=normal make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.production_ddl_v3_recurrence,module.database_migrations.unit.production_ddl_v3_recurrence` | V37/V38 migration policy and physical catalog parity | 15/15 |
| `nf-imp08-s36-client-20261030-01` / 0 | `CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | V29/V44 client alias comparisons and model/workspace regressions | 12/12 |
| `nf-imp08-s36-node-20261030-01` / 0 | `CARTULARY_HARNESS_CACHE_MODE=off make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.node_fixture_admission_seals_verified_inputs_and_d6cec88240,web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` | V35 Node admission/materialization and source ownership | 11/11 |
| `nf-imp08-s36-evidence-20261030-01` / 0 | `CARTULARY_HARNESS_CACHE_MODE=normal make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_schema,harness.evidence_accounting.behavior.native_session` | V35 native evidence schema/session | 11/11 |
| `nf-imp08-s36-go-20261030-03` / 0 | `make lint CHECK=go` | V32-C Go format/vet/staticcheck | 15/15 |
| `nf-imp08-s36-types-20261030-01` / 0 | `make lint CHECK=types` | Affected frontend type boundaries | 11/11 |
| `nf-imp08-s36-boundary-20261030-01` / 0 | `make lint CHECK=backend-boundary` | V32-C backend import boundaries | 13/13 |
| `nf-imp08-s36-generated-20261030-01` / 0 | `make source-check CHECK=generated` | V32-C generated projections | 14/14 |
| `nf-imp08-s36-schema-20261030-01` / 0 | `make source-check CHECK=schema` | V32-C/V35/V38 schema projections | 12/12 |
| `nf-imp08-s36-catalog-20261030-01` / 0 | `make source-check CHECK=catalog` | V32-C row catalog and routing | 11/11 |

The final eight test-slice selections are disjoint within each owner and equal
the recorded required union: Network Flow 70, Imports 7, migrations 2, client 5,
web admission/ownership 2, and native evidence 2. The 70 Network Flow rows split
into 37 unit/store, 25 integration and 8 browser/stateful rows. All exact V29–V33
rows and V34–V50 correction obligations are included. `explain-run ...
TARGET=test-slice DETAIL=accounting` for every final test root reports passed
units, native unit exit 0, zero missing outputs and completed cleanup. Observer
success is not substituted for the independently observed owning command exits.
The final comparisons use cache mode `off`; normal-mode Imports, migration and
native evidence runs had zero cache hits as well. No prior-source root is
represented as final qualification.

The final fixture evidence is the union of **U** =
`nf-imp08-s36-unit-20261030-02` and **I** =
`nf-imp08-s36-integration-20261030-02`, beneath the exact result parent above.
Every inventory ID appears once, has a passing comparison, matches its current
frozen manifest/revision, and retains nonempty produced artifacts. Each listed
summary also carries the source/expected bundle hashes, materialized input root,
produced-artifact paths/hashes, comparison outcome and failure phase. No frozen
bytes changed during S34–S36.

| Fixture / revision | Root / exact comparison summary | Manifest SHA-256 |
| --- | --- | --- |
| `NF-FIX-001-cisco-sna-minimal` / 8 | I: `NF-FIX-001-cisco-sna-minimal-2783885623/comparison-summary.json` | `d16582b1b7a79b62f05a7794873af5954efc8432449b2dc5b005cf4ebfade648` |
| `NF-FIX-002-cisco-sna-interface-fields` / 5 | I: `NF-FIX-002-cisco-sna-interface-fields-3407566715/comparison-summary.json` | `9ab9720e77c9612a1c3d747e61ee9e9f5a46a664f4ae743630bcf9ea2318c1c9` |
| `NF-FIX-003-duplicate-headers` / 4 | U: `NF-FIX-003-duplicate-headers-3865760952/comparison-summary.json` | `a770fa3d5375a48b1e34eb421481f2f5e912b7010011e6917c569e8ad69f5252` |
| `NF-FIX-004-rejected-rows` / 6 | I: `NF-FIX-004-rejected-rows-1691477616/comparison-summary.json` | `deee3c770b3d9345def7727d1c100af1b8be57bac6dc59e075767aa5e3093313` |
| `NF-FIX-005-csv-parser-edges` / 3 | U: `NF-FIX-005-csv-parser-edges-3093439341/comparison-summary.json` | `64756221ede6b99faa01f802e5652b65e22139ca1479c0fd3f61a9ff4ef6b616` |
| `NF-FIX-006-cross-table-graph` / 3 | I: `NF-FIX-006-cross-table-graph-2863526772/comparison-summary.json` | `eeae13f8a8de031e8c4f4c83adfe733c35187d4f08534e920030f8997662f1fc` |
| `NF-FIX-007-indicator-linking` / 4 | I: `NF-FIX-007-indicator-linking-258945877/comparison-summary.json` | `884524611efe26879556c742cb6afbb08dc29476d00748012713ccdd53733c6f` |
| `NF-FIX-008-large-limits` / 3 | I: `NF-FIX-008-large-limits-3234023691/comparison-summary.json` | `bdee707a60c85fad3019760a0e4909647065b6baf49bd47036eded7ed9a45146` |
| `NF-FIX-009-soft-delete-stale-graph` / 3 | I: `NF-FIX-009-soft-delete-stale-graph-1761419338/comparison-summary.json` | `4ec59085d9817392cf69bb9f7ce53896416605eccc33b170bd123bd3c63024c0` |
| `NF-FIX-010-json-admission` / 4 | U: `NF-FIX-010-json-admission-2051393043/comparison-summary.json` | `69de015263c02bb89be3f451d2ff554c671dca55e43a93afd4e1f212308e693f` |
| `NF-FIX-011-alias-collision` / 4 | U: `NF-FIX-011-alias-collision-105062040/comparison-summary.json` | `94cce451a5b10f5aa26b18f428693c09082fead8d7f01112e89d38d0249a1ad1` |
| `NF-FIX-012-sys-uptime-timestamps` / 3 | U: `NF-FIX-012-sys-uptime-timestamps-42205694/comparison-summary.json` | `8c3c793cf01dc53f9191055333dc93e77257896e3e06cebfee64ac454133742e` |
| `NF-FIX-013-filename-display` / 4 | I: `NF-FIX-013-filename-display-1645682974/comparison-summary.json` | `63a7afe2d29cd8c99d732d7084f4860ba0ca321b93171c71302c6ea822c8adf6` |
| `NF-FIX-014-cursor-pagination` / 5 | I: `NF-FIX-014-cursor-pagination-2491662639/comparison-summary.json` | `0eabf0e75af97983bf82b18dec17bd1b381aedf54c6f80f0f01be758984b5134` |
| `NF-FIX-015-graph-adapter-input` / 4 | I: `NF-FIX-015-graph-adapter-input-1345761393/comparison-summary.json` | `bdd09489ddada650316702a0bbb8060ec761734ce4b0cf8d67bb90a2ded1fc79` |
| `NF-FIX-016-redaction` / 9 | I: `NF-FIX-016-redaction-3776089172/comparison-summary.json` | `707ae16b1986902e39fcb83c5602281a721c906f5fd22f59093715e9a42b9328` |
| `NF-FIX-017-indicator-link-mismatch` / 3 | I: `NF-FIX-017-indicator-link-mismatch-60525988/comparison-summary.json` | `043b81f172a532312c0d5a276c5101f836f7c425b294d4925099c4008c194ca5` |
| `NF-FIX-018-resource-limits` / 5 | I: `NF-FIX-018-resource-limits-1726720185/comparison-summary.json` | `ab81f596972992948f53c434ec1a7aba5a3530215fa4f171a369924ef17b94b8` |
| `NF-FIX-019-canonical-json-unicode` / 5 | U: `NF-FIX-019-canonical-json-unicode-1100060235/comparison-summary.json` | `69ceebd5a51bce7de6ad2b19a92187a0bf582cf8a7c5e8589d72d37c322208d4` |
| `NF-FIX-020-atomic-import-commit` / 3 | I: `NF-FIX-020-atomic-import-commit-2246569142/comparison-summary.json` | `292e7ae738b7b2a58f853610bd79ef7347b0ec7a1562cb49ff14cbf9384f1dfa` |
| `NF-FIX-021-preview-boundaries` / 3 | U: `NF-FIX-021-preview-boundaries-2018046681/comparison-summary.json` | `d837da92e3e41f4263ddada780c0a9f2871916070493217c4799072420fb7fb7` |
| `NF-FIX-022-timestamp-rulesets` / 4 | U: `NF-FIX-022-timestamp-rulesets-3073319093/comparison-summary.json` | `75f4dea266a516e9b3e697c80d020b87c2d6505634b757b7305793ef54d3b6bd` |
| `NF-FIX-023-import-facade-source-change` / 3 | I: `NF-FIX-023-import-facade-source-change-2310669054/comparison-summary.json` | `539e97258c94a6f01bd54caae7a8c323ed6ac03dcd14eaa1c9b23f32d986d555` |
| `NF-FIX-024-query-normalization-cursors` / 4 | I: `NF-FIX-024-query-normalization-cursors-1937730281/comparison-summary.json` | `b58b8d5d9dd26222149f309cf4641d4878df7ce6936c6b8d72075cc9c9fbda10` |
| `NF-FIX-025-graph-contributors` / 3 | I: `NF-FIX-025-graph-contributors-3095580296/comparison-summary.json` | `ab99bf1b67599a68e6e9c7ada35ddaba792bd25306039ca9ae060e36601bd73f` |
| `NF-FIX-026-audit-and-replay` / 5 | I: `NF-FIX-026-audit-and-replay-1503484169/comparison-summary.json` | `4879b1624be9bb7d0e422c2ffcf3930e69c85fd9706681a41f66d7391f080782` |
| `NF-FIX-027-retention-soft-delete` / 4 | I: `NF-FIX-027-retention-soft-delete-792127358/comparison-summary.json` | `c6665c6e121fd898fa8efe2050dfc3a292f6821ac9db1f9bf74487df8c74d698` |
| `NF-FIX-028-graph-aggregate-bounds` / 3 | I: `NF-FIX-028-graph-aggregate-bounds-2359160686/comparison-summary.json` | `d40a5d3c30854f291cc3e5512b9e63153d4588364e64f8b4207c42380209551e` |

The final browser root additionally retains seven admitted 001 summaries, and
the client root retains the admitted 011 comparison. The Node support root
retains 31 summary artifacts from positive and deliberately failing admission,
comparison and publication scenarios; expected failed summaries are negative
controls, not failed qualification. Go/Node controls prove admission failure
prevents product invocation, materialization is isolated/read-only, and both
success and failure are retained. All 28 owner comparisons reject deliberately
wrong observable expectations. Integrity remains separate from behavior.

**Finding and workstream reconciliation.** F36 is closed by preview-only
accounting and production streaming assertions; F37 by the normalized write
input, validated context and canonical construction; F38 by all 28 actual
comparisons and removal of the error-discarding traversal; F39 by the v2 owner
reference and exact 015 adapter comparison; F40 by Go/Node admission,
materialization, bounded schema-validated evidence and negative controls.
Required discovered corrections F41–F58 are closed: canonical serialization and
migration (S39), lossless JSON (S40), sealed-input scanning (S41), numeric mapping
tuples (S42), DDL guards (S43), catalog/checker event order (S44/S44-R), Core
migration prose (S45), ignored provenance (S46), typed raw canonical values
(S47), timestamp grammar/precision (S48), NFC aliases and immutable historical
approvals (S50), explicit client confirmations (S51), complete all-rejected
samples (S52), query error fidelity (S53), import naming errors (S54), deletion
and cursor errors (S55), contributor staleness (S56), and optional contributor
limits (S57). S49 completes the Node side of F40. Their reproduction, owner
clauses, compatibility decisions and phase exits remain in the chronological
record; this final portfolio supplies current qualification. S33 is DONE as the
aggregate of S33-A–F. No required correction remains open.

**Compatibility and migration handoff.** Public HTTP/WebSocket/extension and
SQL-column shapes remain unchanged, and Network Flow stays major 8/state 5.
Canonical fixes change derived identity bytes and require the authored,
unreleased `db/migrations/00034_network_flow_canonical_identity.sql`. Historical
migrations 1–33 are unchanged. The current generated Imports registry hash is
`6feb41d9fde5fc1633c18abc54293c102e7f957ab2dc4d5c72de99479b71ce60`.
S35 itself adds no migration or digest change. The final upgrade requirements
below supersede narrower job-drain statements in early S39 checkpoints.

1. Before upgrade, stop prior writers and obtain a recoverable, consistent
   pre-upgrade database backup while preserving its referenced retained
   artifacts and matching prior binary. Budget an offline window and working
   storage proportional to retained Network Flow data; this effort makes no
   production performance claim.
2. Drain or cancel **all nonterminal `import.apply_v1` jobs**, including non-Network
   Flow imports, as well as Network Flow-owned jobs and applying Network Flow
   units. The global registry identity changed. Use the owning recovery workflow;
   do not delete jobs, fabricate completion or bypass the migration guard.
3. Apply authored migration 34 through the supported migration workflow, then
   start the matching new binary. Goose owns one atomic transaction. Invalid or
   incomplete retained mapping/provenance/time/receipt/proof context aborts the
   whole migration. The migration's revoked helper objects disappear before
   commit; no runtime legacy codec, alias translator or automatic reset remains.
4. Re-admit pending Network Flow mappings/approvals. Completed Imports approvals,
   plans/outcomes, source identity, audit occurrences and immutable graph results
   remain historical facts. Repair updates derived row/binding references,
   versions and reconstructible receipt/job/proof identities. Do not reapply
   completed imports to compensate for the repair.
5. Refresh invalidated cursors and source-dependent graph snapshots. Refresh
   selected saved graphs before exports requiring current source reconstruction.
   Old exports that fail corrected admission require controlled restore/upgrade
   and re-export; never waive state validation. Changed canonical replay may
   conflict safely while preserving the original committed effect.
6. Before migration deployment, rollback is source/test reversion. After a
   populated migration, restore the matching pre-upgrade backup and prior binary
   together. Down intentionally rejects retained affected state; it is not a
   general inverse. No data reset or legacy runtime compatibility layer is
   authorized by this handoff.

**Operational boundaries and residual work.** Migration was exercised only in
owned test databases; no production migration, release or deployment occurred.
The existing revision-2 fixtures remain available in Git history, and retained
phase inputs preserve the intermediate unreleased revisions. Final revised
bundles and all source changes are reviewable in the working tree. Complete-owner,
full-CI, release/deployment and performance-publication qualification remain
separate activities, not missing gates in this agreed affected portfolio.
Finalizer retained-run maintenance was deliberately skipped because no compatible
successful full warm-check root was supplied; no redundant check was run merely
to obtain one. All required test, comparison, generated/schema/catalog and
boundary checks are present and passing, with no remaining cleanup obligation.

**Documentation exit.** S36's only remaining gate is
`CARTULARY_TEST_RUN_ID=nf-imp08-s36-close-20261030-01 make lint-markdown`, using the
pinned prefix above. Its exact summary will be
`.cartulary/test-results/nf-imp08-s36-close-20261030-01/adhoc/lint-markdown/tool-run-summary.json`.
After native success, record the result and change S36 to DONE. Historical
successes and failures remain historical; none substitutes for the final roots.


`nf-imp08-s36-close-20261030-01 make lint-markdown` passed native 0;
its exact summary is `adhoc/lint-markdown/tool-run-summary.json` beneath that
run root. **S36 is DONE**, all required binary exits are satisfied, and no required
selected check, comparison, correction or cleanup remains missing. The tracker
is the final controlling handoff; source remains frozen at the identity above.
The final status-record documentation check is
`nf-imp08-s36-close-20261030-02 make lint-markdown` with the same pinned prefix.

For completeness, the preparation roots below retain their own source identity
and successful native exit 0; they are preparation/history, not replacement
qualification. The finalizers' detailed artifacts are
`unit-artifacts/finalize-summary.json`; format roots retain graph summaries.

| Preparation run ID | Command | Recorded source digest |
| --- | --- | --- |
| `nf-imp08-s35-format-20261030-01` | `make format` | `sha256:d3ac1eaa221a4fa49bd5f09b940f880e68414ad9d85135ac537ef61ccd324e5a` |
| `nf-imp08-s35-format-20261030-02` | `make format` | `sha256:c6ff26460265f66047ecc2520dcac669dc75df5271273baea3e6fce14fcca66f` |
| `nf-imp08-s35-format-20261030-03` | `make format` | `sha256:67b10f53ab67489e6c93f7403ce56c6c579b03f1ffc4a750981bdd554e4fd0d5` |
| `nf-imp08-s36-format-20261030-01` | `make format` | `sha256:28a10c3aa975de8956d709237f8bec0a6449609160a861ee31e3c90387dc7725` |
| `nf-imp08-s36-finalize-20261030-01` | `make agent-finalize` | `sha256:60ad5add6add321e704cc6e75bf90e86f863ee80991c67844966a1585a413c76` |
| `nf-imp08-s36-finalize-20261030-02` | `make agent-finalize` | `sha256:2983c43dc1630a1056faeb39e2fba0ef0071043b22c071bd174c5a6264ca0776` |

Read-only selection/diagnostic commands were not product qualification:
`make help ALL=1` (`nf-imp08-s35-help-20261030-01`), the six S36 task guides,
the rejected oversized backend plan and three passing bounded target plans,
S35/S36 `explain-run` diagnostics, and the S35 Imports live status observation.
Their exact targets and roots are recorded in the corresponding checkpoints or
native command history. Final `explain-run` accounting used
`nf-imp08-s36-accounting-unit-20261030-01`,
`nf-imp08-s36-accounting-2-20261030-01` through `-7-20261030-01`, and
`nf-imp08-s36-accounting-imports-20261030-01`, all native 0. The local ledger
transcription initially assumed a flat manifest revision field; it stopped
before writing, was corrected to the manifest's existing `freeze.revision`, and
then recorded the unchanged passing evidence. No test or source repair was
needed for that authoring-only correction.


Final status-record check `nf-imp08-s36-close-20261030-02 make lint-markdown`
passed native 0. Its retained summary is
`.cartulary/test-results/nf-imp08-s36-close-20261030-02/adhoc/lint-markdown/tool-run-summary.json`.
S36 remains DONE; this terminal evidence note changes documentation only.


## 28. Pre-production migration 34 hard cutover

The maintainer subsequently authorized applying migration 34 as a pre-production
hard cutover on 2026-10-31. This authorizes the requested environment operation;
S36's completed implementation/qualification remains historical and unchanged.
The selected database, its current migration state and active writers must be
identified before applying SQL. No database reset, migration-history rewrite or
object-store deletion has been performed or inferred from this instruction.

**OP-01 status: target configuration required.** Read-only preflight found the
repository-supported `make db-migrate` path. It applies current embedded
migrations without resetting database or object storage. The current shell has
no `CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN`, runtime DSN, deployment
configuration override or development TLS configuration. `docker compose -f
docker-compose.dev.yml ps --format '{{.Name}} {{.Service}} {{.State}} {{.Ports}}'`
returned native 1 at configuration interpolation because required TLS/object-store
variables were missing; it did not query or mutate services. The environment-file
or configuration path was requested. No target was guessed from retained test
runs, and no migration has been applied to a user database.

Independent preparation is `make build COMPONENT=migrate`, with run ID
`nf34-cutover-build-20261031-01`, pinned Go 1.27.3 in GO and PATH, and
`CARTULARY_PREPARATION_POLICY=installed_only`. Result and artifact identity will
be recorded after native completion. Once the target is identified, inspect its
actual version and active work, perform the authorized cutover through the
supported migration workflow, and retain the resulting migration/verification
outcome here. Pre-production status does not establish whether this particular
database contains state to retain; target inspection will determine the concrete
operation. No additional compatibility implementation was requested or made.


`nf34-cutover-build-20261031-01 make build COMPONENT=migrate` passed native 0.
The run root is `.cartulary/test-results/nf34-cutover-build-20261031-01/`, retaining
`run-manifest.json`, `run-summary.json` and `cleanup-results.json`; cleanup completed
12/12 steps. Source remains
`sha256:2983c43dc1630a1056faeb39e2fba0ef0071043b22c071bd174c5a6264ca0776`,
with the same pinned toolchain/system as S36. Prepared artifact:
`build/bin/migrate`, SHA-256 `60c81ca732d9613fc18617c8575131407e2e40968907fe51e40bb780f67ccdbc`. This prepares the executable; it does not
apply a migration or establish a target database version. Documentation check:
`nf34-cutover-docs-20261031-01 make lint-markdown`. OP-01 remains awaiting the
selected environment configuration, with no database mutation performed.


## 29. IMP-08 review corrections — F59–F61 and S36 requalification

### 29.1 S58 — specification and tracker reopening

The maintainer explicitly requested implementation of the proposed correction
plan. They confirmed migration 34 remains unapplied and selected migration-only
canonical reordering of unchanged saved-query filter objects. Amend pending 34;
preserve migrations 1–33. No environment cutover or reset is part of this work.
OP-01 remains separate and its previously prepared executable must be superseded.

Entry: dirty `main`, HEAD `98adcc82961456bcb74bbecd0023136485afbf00`, 668
modified/untracked paths, no staged changes at the planning baseline. Preserve
all pre-existing work. The earlier S36 source is
`sha256:2983c43dc1630a1056faeb39e2fba0ef0071043b22c071bd174c5a6264ca0776`;
its 88 selected rows and 28 fixture comparisons are historical evidence. The
first repository mutation of this continuation is this tracker reopening.

| Finding | Owner and required correction | Validation | Risk if unresolved |
| --- | --- | --- | --- |
| F59 / P1 | NF-REQ-032c, Table 12-B: validate complete original live references, repair row ID and fingerprint atomically, and enforce live lineage during read-only state admission. Historical receipts remain historical. | V51: complete vectors, malformed-reference rollback, retained-state admission and reference reuse. | Stale provenance and broken linking concealed by incomplete test fixtures. |
| F60 / P1 | NF-REQ-121a and the approved NF-REQ-032c amendment: canonical permutation of unchanged filter objects in declarations and reconstructible receipts; verify old facts before repairing dependent hashes. | V52: independent old/new order/hash vectors, replay/admission/refresh and atomic rejection. | Successful SQL upgrade leaves saved graphs unable to pass serving admission. |
| F61 / P2 | Table 6-D, NF-REQ-034: check the UTC year range at the shared timestamp scalar boundary, including filters and retained semantics. | V53: endpoint and offset-crossing cases through parser/query routes; no effects on rejection. | Invalid normalized bounds reach storage or cannot be reconstructed. |

These changes cover specification, implementation, tests and handoff documentation.
Complete references make provenance coherent; temporary migration conversion
keeps one permanent serializer; shared scalar admission prevents caller bypass.
Public shapes, SQL columns and major 8/state 5 remain unchanged. Invalid retained
context aborts upgrade; no inference, clamping, dual hashes or runtime repair is
introduced. Populated rollback restores the matching snapshot and binary.

| Workstream | Dependency | Status | Binary exit / principal risk |
| --- | --- | --- | --- |
| S58 specification/tracker | Revalidated entry | DONE | Owner amendment, findings, obligations and dependencies recorded; Markdown passes. Risk: stale qualification. |
| S59 complete bindings | S58 | DONE | V51 passes. Risk: applying current lineage to historical receipts. |
| S60 canonical graph repair | S59 | DONE | V52 passes with only permitted historical fields changed. Risk: excess receipt mutation. |
| S61 shared timestamp boundary | S60 | DONE | V53 and revised 024 pass; valid bytes unchanged. Risk: precision/timezone regressions. |
| S63 binding identity projection | S62 failing reproduction | DONE | V55 passes migration, sorted identity admission and public reuse. Risk: conflating reference order and set identity. |
| S62 combined upgrade serving | S59, S60, S61, S63 | DONE | V54 passes actual startup, linking, replay, refresh, immutable preservation and rollback. Risk: SQL-only evidence. |
| S36 final qualification/handoff | S62, S63 | DONE | Agreed 88-row portfolio plus new obligations, required checks, cleanup and handoff complete on compatible source. Risk: stale evidence/executable. |

Every workstream records its changes, failures, exact commands, source/environment
identity, artifacts, cleanup, risks and successor readiness here, then passes
Markdown before its successor starts. Focused failing reproductions precede each
correction. New runtime defects receive a separate finding/workstream.

V54 adds a registered Network Flow integration test using existing borrowed
migration-database/application fixtures and actual service dependencies. Obtain
its row ID through `make author-test-row-id`; author routing then generate via
Make. No proposed selector is available or passing evidence. V51–V53 extend
existing migration/state, receipt, timestamp and query rows. Changed frozen
bundles increment revision and retain complete input/output comparisons.

After focused exits, run `make agent-finalize`, inspect projections and freeze
source. Qualify the §27.64 88-row portfolio and V51–V54, all 28 fixtures,
generated/schema/catalog source checks, Go/types/backend-boundary lint and
Markdown. Reuse only harness-compatible evidence. Rebuild the migration binary
and record its new hash. Every invocation uses a fresh explicit run ID; native
exit and cleanup are recorded separately from summary publication. Full CI,
complete-owner, deployment and performance claims remain outside acceptance.


**S58 exit:** NF-REQ-032c now specifies complete live-reference repair and
read-only lineage admission separately from historical receipts. Its approved
exception permits only canonical filter permutation and dependent hashes;
NF-REQ-034 explicitly covers UTC conversion in filters and retained semantics.
No public/state-version change. `nf-review-s58-docs-20261031-01 make
lint-markdown` passed native 0; summary is
`.cartulary/test-results/nf-review-s58-docs-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
`nf-review-s58-guide-20261031-01 make task-guide OWNER=module.networkflow`
passed native 0 and confirmed owner-slice selection. All invocations use the
§27.64 pinned Go/PATH and installed-only prefix unless recorded otherwise.
Content/inode baseline after the two authorized documentation edits is retained
in `/tmp/cartulary-nf-review-k7r3vc_a/content-baseline.json`; it covers all
tracked/nonignored untracked files. S58 is DONE; S59 is ready after closing
Markdown `nf-review-s58-close-20261031-01`. Historical §27.64 remains unchanged.


### 29.2 S59 — complete live binding references

S58 closing Markdown passed native 0 at
`.cartulary/test-results/nf-review-s58-close-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
S59 starts with independently corrected four-member fixture references and
negative controls for missing/wrong fingerprints, source locators, extra fields
and duplicate references. Full owner admission follows successful migration.


S59 reproduction `nf-review-s59-repro-20261031-01` selected
`module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca`
through `make test-slice OWNER=module.networkflow ROWS=<that exact ID>`, cache
off. Native Make exit 2; row/unit exit 10, product/test_assertion_failure; 14/14
cleanup. Source `sha256:645e421360a1c702e0dc75a1ca9dde934b39fd7f43c186809629af7601589f73`.
Root `.cartulary/test-results/nf-review-s59-repro-20261031-01/` retains the
manifest, target and row results; native logs are empty, so no assertion excerpt
is claimed. Independent corrected references expose the known SQL mismatch.
`nf-review-s59-explain-20261031-01 make explain-run` with that exact RESULTS_DIR,
TARGET=test-slice DETAIL=summary passed as diagnosis only.

The correction now validates all four original reference fields, uniqueness and
ordered ID projection, updates row ID/fingerprint together, and applies the
same live-lineage invariant during read-only state admission. Historical receipt
validation remains separate. Full admission and actual linking row lookup were
added to the migration test; fixture expectations remain independently authored.


**S59 exit:** V51 passes. Migration and live admission reject incomplete or
inconsistent references; corrected references pass the actual linking lookup,
full retained-state admission, corruption controls and soft-deleted retention.
Historical receipts remain unjoined to current identities. No additional owner
change or compatibility layer was needed.

`nf-review-s59-validate-20261031-01` passed native 0 with cache off and verbose
output: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9`.
Source `sha256:4a8acd306e25e60cd46cb7d8e6c42b62ec14090cbc0a49f70007f521f6c0708c`; toolchain/system match §27.64.
Root `.cartulary/test-results/nf-review-s59-validate-20261031-01/` retains
manifest, target/row summaries and cleanup results. All cleanup completed.
`nf-review-s59-format-20261031-01` and `-02` (`make format`) passed native 0.
S59 is DONE; S60 is ready after `nf-review-s59-close-20261031-01 make
lint-markdown`. Remaining risks are F60/F61 and combined upgrade coverage.


### 29.3 S60 — canonical filter order in retained graphs

S59 closing Markdown passed native 0. Its GO override contained a transcription
typo (`linux-ad64`); this frontend-only check did not invoke Go. Product runs
retain the correct pinned prefix. S60 begins with an independent seven-filter
vector covering HTML-sensitive values, Unicode separators and an ordinary
non-ASCII control. Authoring calculations first matched existing reviewed
historical/current hash controls; no product serializer generated expectations.
The vector is `canonical_filter_order_migration.json`. Preservation checks now
compare filter multisets; exact order and hashes remain independent assertions.


S60 reproduction `nf-review-s60-repro-20261031-01` selected the canonical
migration row through `make test-slice OWNER=module.networkflow`; native Make
exit 2, product/test_assertion_failure. Its exact selection and source identity
are retained in that root's run-manifest; prior S59 passed on its prior source.
The correction now computes explicit C-collated canonical filter permutations,
rejects duplicates, writes declaration arrays/checksums together, and repairs
receipt arrays/checksums/request identities from their historical response only.
Original hashes are validated before identity publication; helper functions are
revoked and dropped within the migration transaction.


**S60 exit:** V52's migration, receipt-integrity and authoring-admission rows pass.
The new vectors initially overescaped bytea request hashes; therefore the
reproduction and first validation failures do not establish the filter-order
regression by themselves. The corrected independent vectors establish different
historical/current orders and exact repaired responses, hashes and admission.
Duplicate filters and tampered receipts/proofs abort without retained changes.
Public refresh and actual startup remain V54 obligations in S62.

`nf-review-s60-validate-20261031-01` failed native 2 with product/test assertion
failure on the migration row; the two unit rows passed. Source
`sha256:1a154d3942ea1cb758e01707ead4bc0f840b861df4b16770e61e3e08898b61a7`.
After fixing the fixture bytea encoding, `nf-review-s60-validate-20261031-02`
passed native 0, cache off, source
`sha256:ac9b5934ce820e9b05973810c75a440f009cdb47792020e1fcc41444f23b7d18`.
Both used the pinned prefix and exact command
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.query_authoring_admission`.
Their roots beneath `.cartulary/test-results/` retain manifests, row/unit results,
summary and completed cleanup. Formatting runs `nf-review-s60-format-20261031-01`
and `-02` passed native 0. S60 is DONE; S61 follows closing Markdown
`nf-review-s60-close-20261031-01 make lint-markdown`. No runtime compatibility
codec, public version change or edits to migrations 1–33 were introduced.


### 29.4 S61 — shared UTC timestamp scalar admission

S60 closing Markdown passed native 0. Add independent year-boundary cases to the
existing timestamp row, actual row/graph query routes in revised fixture 024,
and historical invalid-bound migration vectors before closing V53.


`nf-review-s61-repro-20261031-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b`
failed native 2 (product/test assertion), source
`sha256:1434d286ddfffa1e3b677a7af1b3994a152b627494383ac0f4c9216a75001a55`,
cleanup 10/10. The independent tests exercise both offset crossings directly at
the shared scalar boundary. The correction centralizes UTC admission for every
successful RFC3339 parse and record timestamp result. Migration 34 now rejects
noncanonical/out-of-range retained timestamp bounds without changing scalars.
Fixture 024 revision 5 adds sixteen actual endpoint probes (both fields, both
extremes, four routes), with complete errors and unchanged durable-state checks.
No expected response was generated by the product. Formatting
`nf-review-s61-format-20261031-01 make format` passed native 0.


**S61 exit:** V53 passes; valid endpoint/offset timestamps remain accepted,
invalid conversions fail scalar/record/filter admission, and all sixteen route
errors compare exactly without durable mutation. Independently hashed invalid
retained filters abort migration atomically. Fixture 024 revision 5 passes.
`nf-review-s61-validate-20261031-01` passed native 0 with pinned prefix, cache off:
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367`.
Source `sha256:c419520e76137836394e84d6be56c9bdf7c7d12689a922a3a127d7bfc5e5a453`.
The exact run root retains all selected rows, fixture comparison summary,
run-manifest, run-summary and completed cleanup. S61 is DONE; S62 follows
`nf-review-s61-close-20261031-01 make lint-markdown`. Combined startup, immutable
result preservation, reference reuse and public replay/refresh remain V54.


### 29.5 S62 — actual application on upgraded retained state

S61 closing Markdown passed native 0. V54 will borrow the same harness-issued
migration scratch database for application startup; its original fixture retains
ownership and cleanup. TH-HARNESS-REQ-810/359 continue to govern source identity,
exclusive isolation and full migration replay. A narrow descriptor accessor is
test support only; it does not expose a production migration or connection API.
The test includes a real published/selected immutable Graph Projection result
and lease before upgrade, exact preservation, public link reuse/replay, historical
graph replay, refresh publication, and rejection rollback with the same context.


V54 selector authored by `nf-review-s62-author-20261031-01 make author-test-row-id FAMILY_ID=module.networkflow.integration CLAIM='Canonical migration preserves immutable results and admits public linking replay and refresh' SELECTOR_KEY=TestNetworkFlowCanonicalUpgradeServing`:
`module.networkflow.integration.canonical_migration_preserves_immutable_results_c092f26ac0`.
Authored routing declares exclusive postgres_migration and object_store dependencies;
`nf-review-s62-generate-20261031-01 make generate` and
`nf-review-s62-format-20261031-01 make format` passed native 0.
`nf-review-s62-guide-20261031-01 make task-guide OWNER=module.networkflow`
confirmed selection. The first validation command omitted OWNER/SUITE and was
rejected before execution (native 2, usage_error); it supplies no test evidence.
The corrected invocation selects V54 under OWNER=module.networkflow. Support
validation separately selects the existing
`module.database_migrations.unit.test_harness_targeted_operation_validation`
under OWNER=module.database_migrations; `nf-review-s62-support-20261031-01`
passed native 0. All use the pinned prefix and cache off. The new test retains
bounded stage/outcome summaries `canonical-upgrade-true.json` and `-false.json`
inside its selected run root, without connection strings or service credentials.


**S62 failure and dependency correction:** `nf-review-s62-validate-20261031-02`
failed native 2 at application crypto admission: the SQL-only vector omitted
`application_crypto_format`. The serving fixture now seeds that historical marker
and uses the supported auth actor fixture. `nf-review-s62-validate-20261031-03`
then passed actual startup and failed reference reuse: HTTP 201 instead of 200.
Both roots retain native assertion logs, bounded stage summaries and completed
cleanup. This is new product evidence, not a reason to weaken the assertion.

**F62 / P1 — binding identity order after row-ID repair.** NF-REQ-144a and
Table 15-D define identity using sorted row IDs, independent of ordered public
references. S59 incorrectly checked the identity array in reference order, and
migration preserved old identity order even when canonical row IDs reverse it.
A valid repaired link can therefore produce a duplicate binding. The S58 wording
must explicitly defer to the sorted-set contract; historical S59 qualification
is incomplete for this case.

**S63 — canonical binding identity projection (DONE).** Depends on S62's
failing public reproduction, not completion of S62. Repair the specification
clarification, original sorted-identity admission, migration's new sorted identity
projection, current read-only admission and independent fixture identity arrays.
Preserve public reference order, binding IDs, historical receipts and audit facts.
No new migration: 34 remains pending. V55 selects the migration row, V54 serving
row and existing linking fixture. Exit: exact sorted identity assertions, invalid
projection rollback/admission, public duplicate reuse and no replay effects pass.
Risk: confusing display/reference order with set identity. S62 and S36 depend on
this passing exit; no current final qualification is claimed.


**S63 exit:** `nf-review-s63-validate-20261031-01` passed native 0, source
`sha256:51e92c0ce4028cada04bbca4b0eda352c817a01f65b2b092e51c49f34e3d0f9d`,
with the pinned prefix and cache off:
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.canonical_migration_preserves_immutable_results_c092f26ac0,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9`.
All cleanup completed. Exact identity arrays, original/current permutation
rejections, complete linking fixture and real post-upgrade duplicate reuse pass.
The combined serving stages reached `complete` and `rollback_verified`; immutable
projection/lease facts, historical graph responses and replay effects compare.
`nf-review-s63-format-20261031-01` and `-02` passed native 0. An authoring script
initially assumed the negative precision vector had successful expectations; it
stopped before the test edit, was corrected, and no validation ran on that partial
edit. F62 is resolved. S62 resumes after closing Markdown
`nf-review-s63-close-20261031-01 make lint-markdown`.

S62 earlier source/cleanup accounting: `-validate-02` source
`sha256:2113ea72209be3358046d6eff157ccc4f1c23f42c3ae6509bc2ec57d7b1444b8`,
14/14 cleanup; `-validate-03` source
`sha256:955183dd0dde5c1d247689f209e49c4efb27e6cf85cdcefe7075dacd0a5ce2af`,
14/14 cleanup. Both used `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_migration_preserves_immutable_results_c092f26ac0`.
The support run used the first source and completed 10/10 cleanup.


**S62 exit:** V54 passes actual startup, exact repaired-reference duplicate reuse,
link/graph/refresh replay without additional effects, successful refresh, complete
retained owner admission, immutable result/lease preservation, and atomic failed
upgrade. A separately authored historical link receipt retains old references
and bytes through upgrade/serving; it is not joined to repaired live row IDs.
No runtime admission bypass or new production test hook was added.

`nf-review-s62-validate-20261031-04` passed native 0 with pinned prefix and cache off:
`make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.canonical_migration_preserves_immutable_results_c092f26ac0`.
Source `sha256:a9f8735239bd2a6f0d16cb368ae6be5508986a560c2528c674e243ac3e4c3427`; cleanup 14/14 completed.
The run root retains exact manifest/row results and both passing bounded upgrade
summaries. `nf-review-s62-format-20261031-02 make format` passed native 0.
S62 is DONE; S36 follows `nf-review-s62-close-20261031-01 make lint-markdown`.

Final affected selection is the exact §27.64 88-row union plus V54 and
`module.database_migrations.unit.test_harness_targeted_operation_validation`:
90 rows. Finalization regenerates the updated migration digest, then source is
frozen before these rows and required non-row checks run. Only compatible roots
may qualify completion. Historical successes retain their historical status.


### 29.6 S36 — finalization and affected qualification

S62 closing Markdown passed native 0. Run `nf-review-s36-finalize-20261031-01
make agent-finalize` with the pinned prefix, installed-only policy and cache off.
There is no compatible successful full warm-check root; retained-run maintenance
is skipped rather than launching an unrelated full check. Inspect all generated
changes, preserve migrations 1–33 and existing unrelated work, then freeze source.
Final acceptance is the 90-row affected portfolio, 28 fixture summaries, required
source/lint checks, migration binary identity, cleanup and documentation handoff.
OP-01 deployment remains outside this iteration.


Finalizer `nf-review-s36-finalize-20261031-01` stopped before mutation (native 2,
harness/child_target_failure): its schema precheck found migration 34's generated
history hash stale after S63. The authored SQL hash is
`6ad004256b5e2b55d0f7d1d73f6ee78218a553921199bb8a86650685a3fddc5a`;
the projection still contained the pre-S63 hash. Refresh through
`nf-review-s36-generate-20261031-01 make generate`, then run a fresh finalizer.
The failure summary is the exact root's `unit-artifacts/finalize-summary.json`.
No generated file is hand-edited and no broader final qualification has begun.


`nf-review-s36-generate-20261031-01 make generate` and
`nf-review-s36-finalize-20261031-02 make agent-finalize` passed native 0.
Inspection found only the expected migration-history hash refresh beyond the
pre-finalization source (plus tracker edits). The finalizer itself changed zero
files. Migrations 1–33 are byte-unchanged. Retained full warm-check maintenance
remains explicitly skipped; no full-check claim is made.

Source is now frozen at
`sha256:986a41d8eb5b59c7af5e02b8da00c1a3e26de4ffcc6ab65377c03318d9dee961`.
The existing dirty HEAD remains unchanged. The final 90-row selection is derived
from the eight exact §27.64 manifests, adding only the two registered rows
specified above. All final invocations use explicit fresh IDs and cache off.


Final interim checkpoint: Network Flow unit/integration/browser, DDL/support,
client, Node fixture support and native evidence portfolios passed native 0:
83 selected rows. All 28 owner comparison summaries have passing outcomes,
current frozen revision/manifest hashes and matching retained-artifact hashes.
Generated/schema/catalog checks and Go/types/backend-boundary lint passed.
The seven-row Imports root remains running; S36 is not closed yet. All these
roots have the frozen source and the §27.64 toolchain/system identities.

`nf-review-s36-migrate-build-20261031-01 make build COMPONENT=migrate` passed
native 0. Prepared `build/bin/migrate` SHA-256 is
`1d87c6d0ff681988837b5e916d830bc1b3863deaa9ca1dbc8803eb4e81fc294d`.
This supersedes OP-01's earlier prepared executable; no user database was changed.
Read-only inspection confirms source remained frozen except for tracker prose.
`git diff --check` reports only pre-existing intentional trailing field spaces in
frozen fixture 002's CSV rows 2/4; byte-faithful parser stimuli are preserved.
Touched correction paths passed the narrower whitespace check.

The local handoff-index authoring initially looked for `status` instead of the
execution schema's `comparison_status`; it stopped before writing, was corrected,
and reconciled all 28 unchanged passing summaries. This was a documentation
transcription correction, not a test failure or replacement verification.

### 29.7 S36 — final evidence ledger and compatibility handoff

All 90 selected rows passed native 0 at the frozen source above. All required
non-row checks and the migration build passed native 0. The exact common prefix
is §27.64's pinned PATH/GO plus `CARTULARY_PREPARATION_POLICY=installed_only`
and `CARTULARY_HARNESS_CACHE_MODE=off`; add each literal
`CARTULARY_TEST_RUN_ID` from this table. Every root is
`/home/jochi/code/cartulary/.cartulary/test-results/<Run ID>/`, retaining
`run-manifest.json`, `run-summary.json`, `cleanup-results.json`, target summaries,
row/unit results and native logs. All cleanup completed; no missing outputs.

| Run ID | Exact command after common prefix/run ID | Rows | Native exit | Cleanup |
| --- | --- | ---: | ---: | --- |
| `nf-review-s36-unit-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.store.network_flow_selector_covers_only_active_and_sof_400db83232,module.networkflow.unit.alias_fixture_compares_normalized_header_keys_an_600fafb6f0,module.networkflow.unit.bounded_graph_aggregation_and_selectors,module.networkflow.unit.canonical_fixture_compares_exact_bytes_unicode_a_a519d58721,module.networkflow.unit.configuration_claim_and_manifest_admission,module.networkflow.unit.csv_fixture_compares_logical_records_exact_bytes_74b37b1093,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a,module.networkflow.unit.graph_telemetry_boundary,module.networkflow.unit.import_owner_error_translation_rs08_3d3f02e050,module.networkflow.unit.import_preview_admission_f0dc5f2d8b,module.networkflow.unit.json_fixture_compares_actual_strict_request_admi_1249b283f7,module.networkflow.unit.network_flow_selector_covers_a_table_query_uses_6caf43b23c,module.networkflow.unit.network_flow_selector_covers_csv_preview_consume_66f3f08c99,module.networkflow.unit.network_flow_selector_covers_duplicate_csv_heade_61bc2b1f54,module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c,module.networkflow.unit.network_flow_selector_covers_every_fixture_row_h_0f0889c847,module.networkflow.unit.network_flow_selector_covers_every_network_flow_43bd8f4992,module.networkflow.unit.network_flow_selector_covers_import_apply_replay_8b8c7de503,module.networkflow.unit.network_flow_selector_covers_import_facade_previ_2371e1f671,module.networkflow.unit.network_flow_selector_covers_interface_fields_st_6333a351b3,module.networkflow.unit.network_flow_selector_covers_linking_an_endpoint_cfa3b46c37,module.networkflow.unit.network_flow_selector_covers_logs_telemetry_audi_61f27d1e10,module.networkflow.unit.network_flow_selector_covers_selecting_multiple_861eb26e58,module.networkflow.unit.network_flow_selector_covers_source_row_digest_v_c4ff94923a,module.networkflow.unit.network_flow_selector_covers_timestamp_profile_p_bdf924da13,module.networkflow.unit.ordinal_fixture_compares_source_inspection_mappi_9efd893771,module.networkflow.unit.preview_fixture_compares_fifty_record_boundaries_5f51ca2632,module.networkflow.unit.query_authoring_admission,module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.source_inspection_header_only,module.networkflow.unit.streaming_apply_validates_bounded_summaries_and_94281c36fd,module.networkflow.unit.table_lifecycle_contract,module.networkflow.unit.time_bucket_graph_backend,module.networkflow.unit.timestamp_fixture_compares_grammar_precision_utc_d5c6e48edb,module.networkflow.unit.timestamp_profile_exact_grammar_precision_and_zo_2f4a52432b,module.networkflow.unit.unicode17_text_canonicalization_3de608c3dc,module.networkflow.unit.uptime_fixture_compares_auxiliary_ordinal_admiss_0dd963f352` | 37 | 0 | 17/17 |
| `nf-review-s36-integration-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.aggregate_bounds_fixture_compares_legal_source_c_7934db1cd2,module.networkflow.integration.atomic_fixture_compares_every_commit_boundary_ro_a4cdb920d4,module.networkflow.integration.audit_fixture_compares_publication_binding_and_g_5d2f1cd15b,module.networkflow.integration.bounded_graph_contributor_pipeline,module.networkflow.integration.canonical_identity_migration_repairs_retained_ro_14035fcdca,module.networkflow.integration.canonical_migration_preserves_immutable_results_c092f26ac0,module.networkflow.integration.contributor_fixture_compares_exact_graph_vertex_91ea7945d2,module.networkflow.integration.cross_table_graph_fixture_compares_exact_importe_be02f892b9,module.networkflow.integration.extension_state_v4_v1_rejection,module.networkflow.integration.graph_adapter_fixture_compares_v2_input_trusted_fcfebf06a2,module.networkflow.integration.indicator_linking_fixture_compares_core_create_a_84092b62e9,module.networkflow.integration.indicator_mismatch_fixture_compares_authorizatio_eb91e61505,module.networkflow.integration.interface_fixture_compares_nullable_fields_dupli_7b4e71f923,module.networkflow.integration.large_limits_fixture_compares_valid_imported_row_470de5552b,module.networkflow.integration.minimal_fixture_compares_imports_publication_pro_73de2c8ef0,module.networkflow.integration.naming_fixture_compares_imports_filename_normali_d840a6c98d,module.networkflow.integration.pagination_recovery,module.networkflow.integration.paging_fixture_compares_exact_keyset_cursor_bind_2697ec393d,module.networkflow.integration.privacy_fixture_compares_real_preview_diagnostic_a0f7197388,module.networkflow.integration.query_normalization_fixture_compares_scopes_filt_bae5383367,module.networkflow.integration.rejected_fixture_compares_exact_diagnostics_part_3d34882015,module.networkflow.integration.resource_limits_fixture_compares_application_con_77d0adda0f,module.networkflow.integration.retention_fixture_compares_soft_delete_incident_570ae9035a,module.networkflow.integration.source_changed_fixture_compares_every_source_ver_213b5f7eab,module.networkflow.integration.stale_graph_fixture_compares_soft_deletion_contr_5ee08e6e49,module.networkflow.integration.table_lifecycle_admission_replay` | 26 | 0 | 23/23 |
| `nf-review-s36-imports-20261031-01` | `make test-slice OWNER=module.imports ROWS=module.imports.integration.current_cancellation,module.imports.integration.current_workflows,module.imports.integration.network_flow_atomic_unit_commit_6b6ded6873,module.imports.integration.network_flow_default_limit_100001_rows,module.imports.integration.network_flow_safe_owner_error_translation_rs08_ef0e4eafa0,module.imports.integration.uncommitted_unit_cancellation,module.imports.integration.unit_commit_crash_recovery_108cd054a6` | 7 | 0 | 23/23 |
| `nf-review-s36-browser-20261031-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.import_recovery,module.networkflow.browser.network_flow_selector_covers_alias_suggestions_u_891ee3e619,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_e_346ec21519,module.networkflow.browser.network_flow_selector_covers_selecting_a_graph_v_171429ad9e,module.networkflow.browser.query_authoring_integrity,module.networkflow.browser.table_lifecycle_continuity,module.networkflow.browser_stateful.saved_graph_exact_result_lifecycle,module.networkflow.browser_stateful.verify_protected_network_analysis_state_is_disca_21a5de1ebf` | 8 | 0 | 24/24 |
| `nf-review-s36-ddl-20261031-01` | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.production_ddl_v3_recurrence,module.database_migrations.unit.production_ddl_v3_recurrence,module.database_migrations.unit.test_harness_targeted_operation_validation` | 3 | 0 | 16/16 |
| `nf-review-s36-client-20261031-01` | `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.alias_fixture_compares_deterministic_suggestions_9f95bc6309,web.networkflow.regression.analytical_import_ownership,web.networkflow.regression.analytical_import_recovery,web.networkflow.regression.networkanalysisworkspace_previews_an_ordinal_awa_07f315d42a,web.networkflow.regression.networkflowimportmodel_suite_12249943c5` | 5 | 0 | 12/12 |
| `nf-review-s36-node-20261031-01` | `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.node_fixture_admission_seals_verified_inputs_and_d6cec88240,web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` | 2 | 0 | 11/11 |
| `nf-review-s36-evidence-20261031-01` | `make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.native_schema,harness.evidence_accounting.behavior.native_session` | 2 | 0 | 11/11 |
| `nf-review-s36-generated-20261031-01` | `make source-check CHECK=generated` | 0 | 0 | 14/14 |
| `nf-review-s36-schema-20261031-01` | `make source-check CHECK=schema` | 0 | 0 | 12/12 |
| `nf-review-s36-catalog-20261031-01` | `make source-check CHECK=catalog` | 0 | 0 | 11/11 |
| `nf-review-s36-go-20261031-01` | `make lint CHECK=go` | 0 | 0 | 15/15 |
| `nf-review-s36-types-20261031-01` | `make lint CHECK=types` | 0 | 0 | 11/11 |
| `nf-review-s36-boundary-20261031-01` | `make lint CHECK=backend-boundary` | 0 | 0 | 13/13 |
| `nf-review-s36-migrate-build-20261031-01` | `make build COMPONENT=migrate` | 0 | 0 | 12/12 |

All final roots share toolchain identity
`sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`
and system identity
`sha256:9b915d7b93e8e24ac07de37b2ee4a4869ef3a27d42a23e3a42f7986ee64cefb1`.
Selections equal the historical 88-row affected union plus V54 and the existing
fixture-capability row. Each test root was inspected with
`make explain-run RESULTS_DIR=<exact root> TARGET=test-slice DETAIL=accounting`,
using IDs `nf-review-s36-accounting-<group>-20261031-01` for groups unit,
integration, imports, browser, ddl, client, node and evidence; all observers
returned native 0. They are diagnosis/accounting, not complete-owner audits.

Final fixture evidence uses **U** = `nf-review-s36-unit-20261031-01` and
**I** = `nf-review-s36-integration-20261031-01`. Inventory equality is exact:
28 IDs, each with current manifest/revision, a passing comparison, and nonempty
produced artifacts whose sizes/hashes match the retained bytes. The summaries
also identify materialized inputs, bundle hashes, outcomes and failure phases.

| Fixture | Revision | Root / comparison summary | Manifest SHA-256 |
| --- | ---: | --- | --- |
| `NF-FIX-001-cisco-sna-minimal` | 8 | I: `NF-FIX-001-cisco-sna-minimal-1396631393/comparison-summary.json` | `d16582b1b7a79b62f05a7794873af5954efc8432449b2dc5b005cf4ebfade648` |
| `NF-FIX-002-cisco-sna-interface-fields` | 5 | I: `NF-FIX-002-cisco-sna-interface-fields-807978531/comparison-summary.json` | `9ab9720e77c9612a1c3d747e61ee9e9f5a46a664f4ae743630bcf9ea2318c1c9` |
| `NF-FIX-003-duplicate-headers` | 4 | U: `NF-FIX-003-duplicate-headers-1702681146/comparison-summary.json` | `a770fa3d5375a48b1e34eb421481f2f5e912b7010011e6917c569e8ad69f5252` |
| `NF-FIX-004-rejected-rows` | 6 | I: `NF-FIX-004-rejected-rows-77114399/comparison-summary.json` | `deee3c770b3d9345def7727d1c100af1b8be57bac6dc59e075767aa5e3093313` |
| `NF-FIX-005-csv-parser-edges` | 3 | U: `NF-FIX-005-csv-parser-edges-293565602/comparison-summary.json` | `64756221ede6b99faa01f802e5652b65e22139ca1479c0fd3f61a9ff4ef6b616` |
| `NF-FIX-006-cross-table-graph` | 3 | I: `NF-FIX-006-cross-table-graph-4021349183/comparison-summary.json` | `eeae13f8a8de031e8c4f4c83adfe733c35187d4f08534e920030f8997662f1fc` |
| `NF-FIX-007-indicator-linking` | 4 | I: `NF-FIX-007-indicator-linking-3835880255/comparison-summary.json` | `884524611efe26879556c742cb6afbb08dc29476d00748012713ccdd53733c6f` |
| `NF-FIX-008-large-limits` | 3 | I: `NF-FIX-008-large-limits-3635066914/comparison-summary.json` | `bdee707a60c85fad3019760a0e4909647065b6baf49bd47036eded7ed9a45146` |
| `NF-FIX-009-soft-delete-stale-graph` | 3 | I: `NF-FIX-009-soft-delete-stale-graph-3255006283/comparison-summary.json` | `4ec59085d9817392cf69bb9f7ce53896416605eccc33b170bd123bd3c63024c0` |
| `NF-FIX-010-json-admission` | 4 | U: `NF-FIX-010-json-admission-2777538398/comparison-summary.json` | `69de015263c02bb89be3f451d2ff554c671dca55e43a93afd4e1f212308e693f` |
| `NF-FIX-011-alias-collision` | 4 | U: `NF-FIX-011-alias-collision-1132752194/comparison-summary.json` | `94cce451a5b10f5aa26b18f428693c09082fead8d7f01112e89d38d0249a1ad1` |
| `NF-FIX-012-sys-uptime-timestamps` | 3 | U: `NF-FIX-012-sys-uptime-timestamps-2832038374/comparison-summary.json` | `8c3c793cf01dc53f9191055333dc93e77257896e3e06cebfee64ac454133742e` |
| `NF-FIX-013-filename-display` | 4 | I: `NF-FIX-013-filename-display-3490382392/comparison-summary.json` | `63a7afe2d29cd8c99d732d7084f4860ba0ca321b93171c71302c6ea822c8adf6` |
| `NF-FIX-014-cursor-pagination` | 5 | I: `NF-FIX-014-cursor-pagination-204898842/comparison-summary.json` | `0eabf0e75af97983bf82b18dec17bd1b381aedf54c6f80f0f01be758984b5134` |
| `NF-FIX-015-graph-adapter-input` | 4 | I: `NF-FIX-015-graph-adapter-input-1844043812/comparison-summary.json` | `bdd09489ddada650316702a0bbb8060ec761734ce4b0cf8d67bb90a2ded1fc79` |
| `NF-FIX-016-redaction` | 9 | I: `NF-FIX-016-redaction-706918909/comparison-summary.json` | `707ae16b1986902e39fcb83c5602281a721c906f5fd22f59093715e9a42b9328` |
| `NF-FIX-017-indicator-link-mismatch` | 3 | I: `NF-FIX-017-indicator-link-mismatch-3658929199/comparison-summary.json` | `043b81f172a532312c0d5a276c5101f836f7c425b294d4925099c4008c194ca5` |
| `NF-FIX-018-resource-limits` | 5 | I: `NF-FIX-018-resource-limits-979855597/comparison-summary.json` | `ab81f596972992948f53c434ec1a7aba5a3530215fa4f171a369924ef17b94b8` |
| `NF-FIX-019-canonical-json-unicode` | 5 | U: `NF-FIX-019-canonical-json-unicode-3456294415/comparison-summary.json` | `69ceebd5a51bce7de6ad2b19a92187a0bf582cf8a7c5e8589d72d37c322208d4` |
| `NF-FIX-020-atomic-import-commit` | 3 | I: `NF-FIX-020-atomic-import-commit-4026919411/comparison-summary.json` | `292e7ae738b7b2a58f853610bd79ef7347b0ec7a1562cb49ff14cbf9384f1dfa` |
| `NF-FIX-021-preview-boundaries` | 3 | U: `NF-FIX-021-preview-boundaries-4193762092/comparison-summary.json` | `d837da92e3e41f4263ddada780c0a9f2871916070493217c4799072420fb7fb7` |
| `NF-FIX-022-timestamp-rulesets` | 4 | U: `NF-FIX-022-timestamp-rulesets-812850950/comparison-summary.json` | `75f4dea266a516e9b3e697c80d020b87c2d6505634b757b7305793ef54d3b6bd` |
| `NF-FIX-023-import-facade-source-change` | 3 | I: `NF-FIX-023-import-facade-source-change-2295930216/comparison-summary.json` | `539e97258c94a6f01bd54caae7a8c323ed6ac03dcd14eaa1c9b23f32d986d555` |
| `NF-FIX-024-query-normalization-cursors` | 5 | I: `NF-FIX-024-query-normalization-cursors-2473025900/comparison-summary.json` | `5ce13d85d20c24df1f80d811e2da8d83b739419b30d6ae6e556c13a3a37f3f8d` |
| `NF-FIX-025-graph-contributors` | 3 | I: `NF-FIX-025-graph-contributors-1617441352/comparison-summary.json` | `ab99bf1b67599a68e6e9c7ada35ddaba792bd25306039ca9ae060e36601bd73f` |
| `NF-FIX-026-audit-and-replay` | 5 | I: `NF-FIX-026-audit-and-replay-1795392006/comparison-summary.json` | `4879b1624be9bb7d0e422c2ffcf3930e69c85fd9706681a41f66d7391f080782` |
| `NF-FIX-027-retention-soft-delete` | 4 | I: `NF-FIX-027-retention-soft-delete-4250933087/comparison-summary.json` | `c6665c6e121fd898fa8efe2050dfc3a292f6821ac9db1f9bf74487df8c74d698` |
| `NF-FIX-028-graph-aggregate-bounds` | 3 | I: `NF-FIX-028-graph-aggregate-bounds-2672277262/comparison-summary.json` | `d40a5d3c30854f291cc3e5512b9e63153d4588364e64f8b4207c42380209551e` |

**Finding reconciliation.** F59 is closed by complete original/live lineage,
atomic row-ID/fingerprint repair and retained reference reuse. F60 is closed by
canonical filter permutation and dependent identities with full historical replay
and refresh. F61 is closed by shared post-conversion UTC admission and actual
query route/migration vectors. F62 is closed by the sorted identity projection
required by NF-REQ-144a, with exact public duplicate reuse and unchanged replay
effects. Historical link receipts retain their old references and bytes and
pass admission independently of repaired live lineage. F36–F58 retain their
implemented fixes and are covered by the fresh agreed affected portfolio.

S58–S63 and S36 have passing exits, including the closing documentation check. The 28-fixture
corpus, import boundaries, state admission, recovery, authorization and affected
browser workflows all have current selected evidence. Full CI, complete-owner,
release, deployment and performance qualification remain separate activities.

**Compatibility and operational handoff.** No public shape, SQL-column, major
version or state-version change; Network Flow remains major 8/state 5. Pending
migration 34 now repairs both binding reference fields and sorted identity arrays,
permutes unchanged saved-graph filters, and rejects invalid retained UTC bounds.
Migrations 1–33 are unchanged. There is no runtime legacy serializer, translator,
dual hash, automatic reset or new runtime determinism hook. Fixture 024 advances
from revision 4 to 5; all other frozen bundles retain their prior revisions.

The offline upgrade/rollback procedure in §27.64 remains controlling: preserve
a matching pre-upgrade database/artifact snapshot and binary, stop prior writers,
drain all nonterminal Imports apply and Network Flow jobs, apply 34 atomically,
start the matching new server, re-admit pending approvals and refresh stale
cursors/saved graphs. Preserve completed Imports outcomes, audit facts and
historical immutable results. Invalid retained context aborts upgrade; do not
infer missing provenance or waive admission. Before deployment rollback is a
source/test reversion. After a populated upgrade, restore the matching snapshot
and old binary together; Down is not an inverse data conversion.

Prepared executable: `build/bin/migrate`, SHA-256
`1d87c6d0ff681988837b5e916d830bc1b3863deaa9ca1dbc8803eb4e81fc294d`.
This replaces OP-01's earlier prepared hash; OP-01 still needs the selected
environment configuration. No user database, object store, commit or deployment
was changed. Only isolated owned test fixtures were migrated and cleaned up.

No required product check remains missing. Retained warm-check maintenance was
skipped because no compatible full warm-check root was supplied; the finalizer
records that omission. Intentional frozen CSV whitespace remains unchanged.
Prior failures are retained and classified above; prior-source passes are never
substituted for this source's qualification. Final documentation command:
`nf-review-s36-docs-20261031-01 make lint-markdown` with the common prefix.


**S36 terminal checkpoint:** `nf-review-s36-docs-20261031-01 make lint-markdown`
passed native 0. Its summary is
`.cartulary/test-results/nf-review-s36-docs-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
All workstreams and required corrections are DONE. No selected row, comparison,
required non-row check or owned cleanup remains missing. The prior residual
operational boundaries in §27.64 remain separate from this implementation exit.
Final status-record check: `nf-review-s36-close-20261031-01 make lint-markdown`
with the same pinned prefix; exact retained root follows the convention above.


Final status-record check `nf-review-s36-close-20261031-01 make lint-markdown`
passed native 0; summary:
`.cartulary/test-results/nf-review-s36-close-20261031-01/adhoc/lint-markdown/tool-run-summary.json`.
Final read-only content inspection found only tracker evidence changes since the
source freeze; the prepared executable hash still matches. The initial local
inventory comparison included already-deleted tracked fixture paths; filtering
to existing files, consistently with the frozen inventory, confirmed no new or
removed source files. No source repair or repeated product qualification was
needed. S36 is DONE.
