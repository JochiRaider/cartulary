# IR fixture adoption implementation handoff

## Status and scope

The production Network Flow export restriction has been replaced by the typed
portability participant. Current contracts are Network Flow public major **10**,
retained state **7**, Core bundle source-port major **3**, Incident Bundle **6**,
and Recovery generation `recovery.incident_portability.v1`. ZIP is exported;
ZIP, TAR and TAR.GZ import the same current logical format. Generic blocking
remains available for profiles that have no portability participant.

The sixth broad CI run passed all **2,432 selected test rows**, including the
connected A → B → operator backup → C lifecycle, actual browser uploads, historical
graphs, retained reports and the repaired browser synchronization cases. All
**558 cleanup steps** and the retained-secret scan passed. The invocation still
failed overall: **439 of 440 work units passed**, and the vulnerability gate
reported the sole blocking advisory **GO-2026-6909** in the existing SAML
dependency. The gate is unwaived, so complete acceptance remains open. The third
run's cancelled browser cleanup also remains unresolved; later clean runs do
not erase that earlier failure. See
[final qualification](#final-qualification-and-open-blockers) for exact identities,
results, limitations and remaining work.

The connected full-corpus API scenario passed in
`ir-portability-nf-transfer37` (native exit 0; 220,840 ms). It imports 1,500 Timeline
records, 30 hosts, 20 identities and all three Cisco sources; applies 222 alias
source associations; creates the synthetic investigation; selects day 1 before
later imports; freezes and renders a report; transfers A → B with distinct local
key material; and invokes the production operator to back up B and restore C.
Public reads compare the semantic inventory, selected graphs, source provenance,
attachment bytes/history and permitted report files at B and C. Source sessions
cannot authenticate to B/C; B's prior sessions cannot authenticate after restore.
Fresh destination authentication follows the Recovery account policy.

This is implementation-slice evidence, not final-source acceptance. The run's
source digest is
`sha256:588d86bdb0f9298ad60b15e63066fb1804a76583218e291a01e11d810b494995`.
The checkout began at `f9650a0b9d41c752016b59660fc015bd4bcdb150` and remains
uncommitted. Later changes require compatible requalification; no baseline commit
or selection plan substitutes for a passing retained run.

Additional focused evidence:

| Run ID | Selection | Native exit / outcome |
| --- | --- | --- |
| `ir-portability-browser-import01` | Actual browser upload/mapping journey for Core and all three flow files | 0 / pass; 141,543 ms |
| `ir-portability-history03` | `module.incidentbundles.integration.network_flow_historical_portability` | 0 / pass; 18,390 ms; rejected diagnostics, renamed/deleted source, retired and active selected graphs, exact re-export and unrelated destination graph isolation |
| `ir-portability-restore-browser03` | `module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0` | 0 / pass; 86,165 ms; production operator backup/restore, serving admission, fresh browser login and workbook query |
| `ir-portability-schema09` | `make source-check CHECK=schema` | 0 / pass; 7,548 ms |
| `ir-portability-generate67` | `make generate` | 0 / pass; 10,290 ms |
| `ir-portability-format89` | `make format` | 0 / pass; 6,973 ms |

The independently initialized transferred/restored Network Flow browser journeys
passed in `ir-portability-lifecycle-browser05` (native exit 0, 113,663 ms),
including fresh login, exact selected history without refresh, and downloads of
the approved manifest and explicitly selected diagram. Its source digest is
`sha256:ce3c6cf6f334acc18951d03bfc0dd8b7a4e38b1ec6acd6a42f445a9dcd46274e`.
Visual review then found a misleading job-less graph label and low-contrast
report links. Both are corrected; the successor browser assertions passed in
`ir-portability-final-ci02`. Artifact-mode review of that run's transferred
investigation captures confirmed the selected-result label and readable report
links with visible keyboard focus. This was image inspection at 1280×720, with
no DOM or accessibility observations; it is not a general accessibility claim.
No visual goldens were changed.

Additional passing development runs (all native exit 0):

| Run ID | Exact command after `CARTULARY_TEST_RUN_ID=<id>` | Result |
| --- | --- | --- |
| `ir-portability-attribution-transfer01` | `make test-slice OWNER=module.imports ROWS=module.imports.integration.committed_source_transfer` | Public transfer/re-export and subsequent native edit preserve immutable creation/source attribution and use the destination updater |
| `ir-portability-bulk-transfer02` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.large_member_public_transfer` | 65 MiB streamed public evidence transfer with exact source/destination digest |
| `ir-portability-harness-runtime01` | `make test-slice OWNER=harness.test_catalog ROWS=harness.test_catalog.behavior.routing_contract` | Declared browser operator binary dependency and partition routing |
| `ir-portability-nf-crash01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.process.the_packaged_standalone_server_composes_the_netw_400a31ad27` | Packaged runtime crash/replay |
| `ir-portability-restored-claim01` | `make test-slice OWNER=module.recovery ROWS=module.recovery.process.inactive_current_restore_preserves_network_flow_7c63ddd6aa` | Inactive restore preserves state; compatible profile admission reconstructs it |
| `ir-portability-attribution-mutations01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.actor_export_combines_native_users_and_imported_f4f3ed2ecc` | No-op, rollback, same-importer edit and reassignment-away-and-back |
| `ir-portability-sourcechecks10` | `make source-check CHECK=all` | Generated, schema, catalog, surface, identity, toolchain and migration-input checks |
| `ir-portability-generate69` | `make generate` | Current authored contracts project successfully |

Browser development failures `lifecycle-browser01`–`04` were test assertions:
Graph mode navigation, remote-browser download handling, then confusing a file's
raw hash with the domain-framed render-manifest hash. Run 05 corrected those
assertions without changing expected product output. `bulk-transfer01` omitted
the evidence owner's transition to available; run 02 exercises that public
transition. `snapshot01` supplied a retained approved mapping to the candidate
mapping route; the test now reuses the independently admitted original candidate.
These failed invocations exited 2 and remain retained, with cleanup evidence.
Final adversarial coverage, compatible final portfolio, acceptance-matrix closure
and finalization remain outstanding.

## Successor implementation

Committed provenance covers every successfully applied Core and Network Flow
unit. Exact uploaded bytes, including unmapped content, remain independently
retained; distinct applications retain distinct provenance while sharing blobs by
content digest. Public metadata and streamed raw-source inspection preserve the
source owner's authorization. Failed apply and exact replay add no provenance.
Live Import sessions, operational receipts and jobs are not transfer dependencies.

Saved graphs retain refresh intent separately from immutable selected-source
bindings. Transfer and Recovery recompute against the exact retained query,
table metadata/fingerprints, algorithm identities and result binding. Rename,
source soft deletion and declaration retirement preserve historical selections.
Incident import publishes only that incident's derived Graph results; Recovery
retains its separate deployment-wide publication and serving-admission lifecycle.

Bulk member transport uses operation-owned sealed streams throughout archive
admission, Core source ports, extensions, Evidence and retained Reporting files.
Canonical NDJSON, bounded control catalogs and digest-addressed blobs replace the
inline-only whole-archive path. The 64 MiB control ceiling and existing archive
ceilings remain. Source owners own validation; bundle transport does not parse
owner SQL. Transaction-keyed validation relations replace temporary-table DDL
without expanding serving-role privileges.

All five Network Flow retained families participate, including deleted tables,
rejected diagnostics, bindings and retired graph declarations. Export shares one
consistent source snapshot; typed actor references close the historical roster.
Preparation binds sealed inputs to the operation and owner. One final transaction
publishes Core state, provenance, Network Flow, reconstructed graphs, permitted
Reporting artifacts, inert attribution, object ownership and the local receipt.
Uncertain commit outcomes resolve receipts before cleanup. Replaced-format
readers, conversion/backfill and fabricated workflow jobs are not admitted.

Retained report files have authenticated incident-scoped inventory and streamed
content routes, plus a browser panel. Their hashes bind original approved inputs
and outputs; transfer does not create native releases, approvals or trust.
Sensitive reveal maps are excluded. Native edits retire only the imported actor
override for the changed action, preserving creation/history attribution and
unchanged assignments in the same transaction.

No developer or production environment has been reset. All runtime qualification
uses harness-owned databases and object stores. Historical access keeps the prior
executable and its data together; state 7 has no in-place downgrade path.

## Historical implementation-slice evidence

The following retained results are implementation-slice evidence, not final-source
qualification. All roots are under `.cartulary/test-results/`; source identities
are recorded by their run manifests.

| Run ID | Public command | Outcome |
| --- | --- | --- |
| `ir-portability-provenance-20261105-02` | `make test-slice OWNER=module.imports ROWS=module.imports.integration.ir_corpus_core` | Passed; public Core provenance survives workflow cleanup and replay |
| `ir-portability-state7-unit-20261105-02` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.saved_graph_receipt_integrity,module.networkflow.unit.table_lifecycle_contract` | Passed |
| `ir-portability-state7-store-20261105-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.store.graph_view_declaration_persistence` | Passed |
| `ir-portability-history-20261105-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.selected_source_history,module.networkflow.store.graph_view_declaration_persistence` | Failed: multi-table selected-source validation confused workspace order with sorted query identity; store case passed |
| `ir-portability-history-20261105-02` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.selected_source_history` | Passed after membership/order correction |
| `ir-portability-stream-20261105-01` | `make test-slice OWNER=module.stagedobjects ROWS=module.stagedobjects.unit.lifecycle_cleanup_outcomes_7f6ad80724,module.stagedobjects.integration.allocation_publication_cleanup_7f6ad80724` | Passed |
| `ir-portability-archives-20261105-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.member_stream_archives,module.incidentbundles.unit.incident_bundle_verification_accepts_safe_zip_an_892d158923` | Passed |
| `ir-portability-cutover-20261105-01` | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.production_ddl_v3_recurrence` | Passed; predecessor rejection leaves schema and ledger unchanged |
| `ir-portability-generate-20261105-17` | `make generate` | Passed |
| `ir-portability-format-20261105-07` | `make format` | Passed |
| `ir-portability-schema-20261105-01`, `-02`, `-03`, `-04` | `make source-check CHECK=schema` | Failed: stale graph-schema pattern, duplicated major-9 assertions, then stale generated topology; corrected and regenerated |
| `ir-portability-schema-20261105-05` | `make source-check CHECK=schema` | Passed |
| `ir-portability-storage-20261105-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.incident_bundle_storage_is_atomic_and_contained_2bfb18f3a9,module.incidentbundles.unit.incident_bundle_verification_accepts_safe_zip_an_892d158923` | Failed to compile: Recovery still used a digest validator removed from storage; moved it to the Recovery adapter |
| `ir-portability-storage-20261105-02` | Same selection as `storage-01` | Passed, including 65 MiB actual confined filesystem staging/publication and interrupted-write cleanup |
| `ir-portability-record-stream-20261105-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.member_stream_records,module.incidentbundles.integration.incident_bundle_import_upload_envelope_failures_bf5449e3d4,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` | Passed |
| `ir-portability-strict-links-20261105-01` | `make test-slice OWNER=module.links ROWS=module.links.unit.strict_portability_preparation_fails_closed_8617ab9352` | Passed |
| `ir-portability-export-stream-20261105-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.incident_bundle_archive_construction_emits_deter_d848a55dc0,module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e,module.incidentbundles.integration.actor_export_combines_native_users_and_imported_f4f3ed2ecc,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` | Passed against sealed-member export assembly |
| `ir-portability-generate-20261105-18`, `-19`, `-20`, `-21` | `make generate` | Passed |
| `ir-portability-format-20261105-08`, `-11`, `-12`, `-13` | `make format` | Passed |

Additional successor export-slice evidence (run IDs below intentionally omit the
older `20261105` segment):

| Run ID | Exact public command after `CARTULARY_TEST_RUN_ID=<id>` | Outer exit / outcome |
| --- | --- | --- |
| `ir-portability-sourceport3-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.cross_owner_import_descriptor_is_closed_9f6d848f63,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` | 2 / compile failure: Links validation still called the replaced buffered export API |
| `ir-portability-sourceport3-02` | Same selection as `sourceport3-01` | 2 / compile failure: shared Links test support still called that API |
| `ir-portability-sourceport3-03` | Same selection as `sourceport3-01` | 0 / pass after adapting validation and test support |
| `ir-portability-build01` | `make build COMPONENT=server-harness` | 2 / frontend prerequisite failed: authored protocol entrypoint named three removed v2 table validators; imports corrected to v3, complete build requalification still outstanding |
| `ir-portability-owner-stream01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.member_stream_records,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` | 2 / compile failure: unused Records import after removing its buffer |
| `ir-portability-owner-stream02` | Same selection as `owner-stream01` | 0 / pass, including producer streaming beyond 64 MiB, joined cancellation, short-write rejection, row ceiling, and public Core round trip |
| `ir-portability-records-stream01` | `make test-slice OWNER=module.records ROWS=module.records.portability.envelope_invariants` | 0 / pass, including 257 envelopes across keyset pages with exact attribution |
| `ir-portability-savedviews-stream01` | `make test-slice OWNER=module.savedviews ROWS=module.savedviews.integration.imported_saved_views_preserve_portable_owners_se_45009255f0,module.savedviews.integration.saved_view_export_emits_the_exact_canonical_elev_fb78b01a43` | 0 / pass |
| `ir-portability-indicators-stream01` | `make test-slice OWNER=module.indicators ROWS=module.indicators.portability.indicator_source_rows_are_deterministic_complete_68b42b02f6` | 0 / pass |
| `ir-portability-revisions-stream01` | `make test-slice OWNER=module.revisions ROWS=module.revisions.portability.contract_and_attribution` | 0 / pass |
| `ir-portability-format14`, `ir-portability-format15`, `ir-portability-format16`, `ir-portability-format17`, `ir-portability-format18` | `make format` | 0 / pass |
| `ir-portability-generate22`, `ir-portability-generate23`, `ir-portability-generate24` | `make generate` | 0 / pass |

Additional import-foundation evidence (same run-root convention):

| Run ID | Exact public command after `CARTULARY_TEST_RUN_ID=<id>` | Outer exit / outcome |
| --- | --- | --- |
| `ir-portability-upload-binding01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.member_stream_archives,module.incidentbundles.unit.incident_bundle_archive_construction_emits_deter_d848a55dc0,module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` | 0 / pass; admission digest verification, all containers, >64 MiB member, deterministic export and Core round trip |
| `ir-portability-record-import01` | `make test-slice OWNER=module.records ROWS=module.records.portability.envelope_invariants,module.records.unit.subtype_presence_catalog` | 0 / pass; streamed envelope preparation and exact subtype inventories |
| `ir-portability-record-member01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.member_stream_records,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` | 0 / pass; descriptor/byte tampering, cancellation, record streams and public Core round trip |
| `ir-portability-fixed-members01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.cross_owner_import_descriptor_is_closed_9f6d848f63,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` | 0 / pass; fixed owner imports use member references; typed composite identity order remains declared |
| `ir-portability-attribution01` | Same exact selection as `ir-portability-fixed-members01` | 0 / pass; transaction-owned historical attribution and streamed Saved Views |
| `ir-portability-savedviews-import01` | `make test-slice OWNER=module.savedviews ROWS=module.savedviews.integration.imported_saved_views_preserve_portable_owners_se_45009255f0,module.savedviews.unit.strict_saved_views_preparation_rejects_illegal_i_d1e9382c70,module.savedviews.unit.strict_saved_views_preparation_rejects_unbounded_0ac27d4f10` | 0 / pass; preserved owners, rollback, strict framing and semantic validation |
| `ir-portability-revisions-import01` | `make test-slice OWNER=module.revisions ROWS=module.revisions.portability.contract_and_attribution` | 2 / regression: malformed Links history was rejected after transaction-local change-set insertion; restored owner validation before writes |
| `ir-portability-revisions-import02` | Same exact selection as `ir-portability-revisions-import01` | 2 / regression: streamed ordering checks preempted the required missing-reference invariant; retained bounded flags and restored reference precedence |
| `ir-portability-revisions-import03` | Same exact selection as `ir-portability-revisions-import01` | 0 / pass after restoring validation-before-write and invariant precedence |
| `ir-portability-parties-import01` | `make test-slice OWNER=module.parties ROWS=module.parties.portability.party_portable_apply_preserves_exact_state_claim_7a428a3abd,module.parties.portability.party_portable_preparation_enforces_strict_shape_fdfc4a86c9` | 0 / pass; streamed exact rows and claim inventory |
| `ir-portability-tasks-import01` | `make test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.incident_portability.source_port_invariants_and_atomicity` | 0 / pass |
| `ir-portability-links-import01` | `make test-slice OWNER=module.links ROWS=module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1,module.links.unit.strict_portability_preparation_fails_closed_8617ab9352` | 2 / decoder error classification regression and extra newline in new test fixture; corrected both |
| `ir-portability-links-import02` | Same exact selection as `ir-portability-links-import01` | 0 / pass; exact round trip and collection failures leave no authoritative rows or attribution |
| `ir-portability-entities-import01` | `make test-slice OWNER=module.entities ROWS=module.entities.support_integration.entity_incident_bundle_invariant_closure_92db429a94` | 0 / pass; streamed preparations, claim reconstruction, strict and collection invariants |
| `ir-portability-indicators-import01` | `make test-slice OWNER=module.indicators ROWS=module.indicators.portability.partition_deterministic,module.indicators.portability.state_invariants,module.indicators.vocabulary.contract_parity_and_portable_admission` | 2 / compile failure: old test bundle helper lacked member-reader methods; replaced with shared test representation |
| `ir-portability-indicators-import02` | Same exact selection as `ir-portability-indicators-import01` | 0 / pass; active uniqueness, same-incident references and preserved repeated observations |
| `ir-portability-core-stream-import01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755,module.incidentbundles.unit.cross_owner_import_descriptor_is_closed_9f6d848f63` | 2 / serving-role failure: temporary table creation conflicts with the adopted no-TEMPORARY privilege policy; replaced runtime DDL with transaction-keyed scratch DML and deferred empty-at-commit guards |
| `ir-portability-generate26` | `make generate` | 2 / changed Recovery catalog digest left the rebuild-result fixture on the prior implementation-binding hash; corrected authored binding |
| `ir-portability-core-stream-import02` | Same exact selection as `ir-portability-core-stream-import01` | 0 / pass under the serving role; transaction-keyed scratch relations replace runtime temporary-table DDL |
| `ir-portability-generate27` | `make generate` | 0 / pass; successor Recovery inventory contains 160 schema tables, of which 119 remain authoritative-required |
| `ir-portability-format40` | `make format` | 0 / pass |
| `ir-portability-scratch-roles01` | `make test-slice OWNER=platform.postgres ROWS=platform.postgres.integration.purpose_identity_roles_and_acl` | 2 / catalog incorrectly classified immutable source tables as updateable; corrected the authored privilege class |
| `ir-portability-scratch-links01` | `make test-slice OWNER=module.links ROWS=module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1,module.links.unit.source_state_contributions_remain_exact_8d3f67c521` | 2 / commit-isolation test was incorrectly routed to a single rollback transaction; extracted a dedicated database row |
| `ir-portability-format41` | `make format` | 0 / pass |
| `ir-portability-generate28` | `make generate` | 0 / pass |
| `ir-portability-scratch-roles02` | Same exact selection as `ir-portability-scratch-roles01` | 0 / pass; no runtime UPDATE grant added |
| `ir-portability-scratch-links02` | `make test-slice OWNER=module.links ROWS=module.links.store.bundle_validation_commit_isolation,module.links.unit.source_state_contributions_remain_exact_8d3f67c521` | 0 / pass; real independent transactions and deferred commit guard |
| `ir-portability-format42` | `make format` | 0 / pass |
| `ir-portability-generate29` | `make generate` | 2 / successor snapshot fixture omitted the new required Reporting chunk table; corrected before regeneration |
| `ir-portability-generate30` | `make generate` | 0 / pass; 161 authored tables, 120 required authoritative tables |
| `ir-portability-reporting-stream01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | 0 / pass; public Reporting artifact round trip and re-export, source snapshot and reference round trips. Launched before generate30; this is focused implementation evidence, not final-source acceptance |
| `ir-portability-format43`, `ir-portability-format44`, `ir-portability-format45` | `make format` | 0 / pass |
| `ir-portability-generate31` | `make generate` | 0 / pass |
| `ir-portability-reporting-storage01` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.portable_artifact_stream_retention` | 0 / pass; 65 MiB+17 bytes, bounded reads/chunks, exact replay, missing/changed chunks, foreign operation, and rollback |
| `ir-portability-member-worker01` | Same exact selection as `ir-portability-upload-binding01` | 2 / compile failure: unused io import after replacing archive buffering; no behavior execution |
| `ir-portability-member-worker02` | Same exact selection as `ir-portability-upload-binding01` | 2 / archive and worker units passed; public import cleanup assertion raced post-terminal member abandonment; replaced immediate stat with bounded cleanup observation and explicit zero-live-member assertion |
| `ir-portability-format32`, `ir-portability-format33`, `ir-portability-format34`, `ir-portability-format35`, `ir-portability-format36`, `ir-portability-format37`, `ir-portability-format38`, `ir-portability-format39` | `make format` | 0 / pass |
| `ir-portability-format20`, `ir-portability-format22`, `ir-portability-format23`, `ir-portability-format24`, `ir-portability-format25`, `ir-portability-format26`, `ir-portability-format27`, `ir-portability-format28`, `ir-portability-format29`, `ir-portability-format30`, `ir-portability-format31` | `make format` | 0 / pass |
| `ir-portability-format21` | `make format` | 2 / configuration admission: new subtype test selector was not ASCII-sorted; corrected before `ir-portability-format22` |
| `ir-portability-generate25` | `make generate` | 0 / pass |

Additional extension and committed-source implementation evidence:

| Run ID | Exact command after `CARTULARY_TEST_RUN_ID=<id>` | Outer exit / outcome |
| --- | --- | --- |
| `ir-portability-member-worker03` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.member_stream_archives,module.incidentbundles.unit.incident_bundle_archive_construction_emits_deter_d848a55dc0,module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e,module.incidentbundles.unit.incident_bundle_verification_accepts_safe_zip_an_892d158923,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755,module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413` | 2 / archive and Core round-trip cases passed; Evidence helper raced terminal cleanup |
| `ir-portability-member-cleanup04`, `ir-portability-member-cleanup05` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_import_z_evidence_cleanup_a3d619b413` | 2 / private staging was incorrectly counted as published state; then Records swallowed the incident-scope invariant as an envelope error. Corrected ownership accounting, bounded cleanup observation and error propagation |
| `ir-portability-member-cleanup06` | Same exact selection as `member-cleanup04` | 0 / all Evidence failure and cleanup variants passed |
| `ir-portability-extension-stream01`, `ir-portability-extension-stream02` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.extension_portability_matrix_and_atomic_import_1b27fc2d91,module.incidentbundles.integration.actor_export_combines_native_users_and_imported_f4f3ed2ecc,module.incidentbundles.unit.cross_owner_import_descriptor_is_closed_9f6d848f63,module.incidentbundles.unit.incident_bundle_contract_registries_use_exactly_1233d2b3f3,module.incidentbundles.unit.incident_bundle_worker_terminal_success_transiti_66f6ace21e` | 2 / first compile stopped on an unused import; second passed member streams, actor and worker cases but found a stale error-registry expectation |
| `ir-portability-extension-registry03` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.incident_bundle_contract_registries_use_exactly_1233d2b3f3` | 0 / registry corrected to include the adopted encoded-archive ceiling error |
| `ir-portability-generate32`, `ir-portability-generate33`, `ir-portability-generate35`, `ir-portability-generate37`, `ir-portability-generate38`, `ir-portability-generate39` | `make generate` | 0 / pass |
| `ir-portability-generate34` | `make generate` | Native return unavailable after continuation; retained step exit 1 / stale committed-source OpenAPI change fingerprint; corrected declaration |
| `ir-portability-openapi-fields01` | `make openapi-compatibility-check` | 2 / same stale fingerprint; retained report identifies the new metadata shape |
| `ir-portability-generate36` | `make generate` | 2 / new test row used the wrong family; corrected routing |
| `ir-portability-format46`, `ir-portability-format47`, `ir-portability-format48`, `ir-portability-format49`, `ir-portability-format51`, `ir-portability-format52`, `ir-portability-format53`, `ir-portability-format54`, `ir-portability-format57`, `ir-portability-format58`, `ir-portability-format59`, `ir-portability-format60` | `make format` | 0 / pass |
| `ir-portability-format50` | `make format` | 2 / malformed test braces; corrected |
| `ir-portability-format55`, `ir-portability-format56` | `make format` | 2 / routing admission failed before retained run publication: unsorted row, then family mismatch; corrected before format57 |
| `ir-portability-committed-port01` | `make test-slice OWNER=module.imports ROWS=module.imports.unit.committed_source_portability,module.imports.integration.ir_corpus_core` | 0 / exact committed-source admission and full Core corpus, replay and workflow cleanup |
| `ir-portability-extension-schema01` | `make test-slice OWNER=harness.generated_artifacts ROWS=harness.generated_artifacts.unit.decode_contract_rejects_duplicate_jsonmembers_005d1b3b94` | 0 / member-stream contract generation and closed contract checks |
| `ir-portability-bundle6-01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.cross_owner_import_descriptor_is_closed_9f6d848f63,module.incidentbundles.unit.incident_bundle_archive_construction_emits_deter_d848a55dc0,module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755` | 2 / test-only duplicate version-map key after cutover |
| `ir-portability-bundle6-02` | Same exact selection as `bundle6-01` | 2 / public Core round trip passed; stale required-member/order expectations and direct version-5 helper calls failed. Corrected current expectations and made unsupported helper versions fail closed |
| `ir-portability-bundle6-03` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.cross_owner_import_descriptor_is_closed_9f6d848f63,module.incidentbundles.unit.incident_bundle_archive_construction_emits_deter_d848a55dc0,module.incidentbundles.unit.incident_bundle_storage_is_atomic_and_contained_2bfb18f3a9` | 2 / archive and streamed storage tests passed, including verified published reads; one retired-version test still classified version 6 as unsupported |
| `ir-portability-bundle6-04` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.unit.cross_owner_import_descriptor_is_closed_9f6d848f63` | 0 / closed current source catalog, prepared bindings and predecessor rejection |
| `ir-portability-history-fields01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.selected_source_history` | 0 / selected history survives rename, expanded refresh, deletion, failure and retirement under retained source row references |
| `ir-portability-committed-transfer01` | `make test-slice OWNER=module.imports ROWS=module.imports.integration.committed_source_transfer,module.imports.unit.committed_source_portability` | 2 / test requested raw capture for Notes, whose target disallows it. Switched the source interpretation to Timeline and reused live mapping semantics during portable mapping admission |
| `ir-portability-committed-transfer02` | Same exact selection as `committed-transfer01` | 0 / independent source transfer through public upload, mapping, apply, export-content, import, metadata/content reads, exact retry and re-export |
| `ir-portability-format61` | `make format` | 0 / pass |

These results precede subsequent import-path edits. No complete import-streaming,
Network Flow transfer, final-source audit or deployment recovery qualification is
implied. The earlier `schema-20261105-05` and
`generate-20261105-18` retained summaries report pass, but their native command
return handles were not preserved across context continuation; that limitation
qualifies the blanket native-exit statement below for those two invocations.

Format attempts `ir-portability-format-20261105-09` and `-10` exited 2
before creating a run root: a new catalog row and then its test selectors were
not ASCII-sorted. `make target-plan TARGET=format JSON=1` identified each
configuration error. The authored ordering was corrected before `format-11`.

Each command used `CARTULARY_TEST_RUN_ID=<listed-id>`. Passing invocations above
exited 0; failed Make invocations exited 2. At that historical point finalization and the connected lifecycle were outstanding.
The current status and later passing lifecycle evidence are recorded above;
these earlier runs are not final-source acceptance.

## Corpus identity and expectation provenance

`fixtures/ir-incident/v1/manifest.json` is revision 1. Its SHA-256 is
`1e6dfc0cf194a53c8bb4b15799a79371668730004147bd77a8ec6b71bd0abd71`.

| Original | Bytes | SHA-256 |
| --- | ---: | --- |
| `aliases.csv` | 15005 | `d6bbe6f78045a4e7c08d5004b85adddcde5f1ada5323fbcfdec0a0898dd19c7f` |
| `hosts.csv` | 5113 | `beaa09d6a77fb1f3c04bc0118ca95dc94a537a3575378d4466874e3987bb34e1` |
| `identities.csv` | 4404 | `0dfb5ed9f845392a96a7c26108db63276c29431cd91e7735cb322e2e10cb9a33` |
| `timeline.csv` | 1602175 | `628ed84a8a53b5982b1036d64397a322fb11c86641f6bd081a8a175b9cbbcc2d` |
| `timeline.tsv` | 1602175 | `54e570f84e1aafc4e705f1c77330377cc7693d983c2fd20e49644a2baf37426c` |
| `cisco-day1.csv` → dependency `day1.csv` | 1876042 | `091dcbc340b2c75713be8dec069e8e301d4ff1cfcd3320b66150a561416d4683` |
| `cisco-day2.csv` → dependency `day2.csv` | 1879293 | `038d92cbd120e1bafdff8d5d7be6ae9d251177d6a3dce9c1efcd48101f26959f` |
| `cisco-overlap.csv` → dependency `overlap.csv` | 467196 | `65d23cc5249247937332168c3207626b130c79853c5bcb5bb970c973a8ed4cdd` |

The pinned `network-flow-orientation` dependency uses expectation revision 1,
manifest SHA-256
`b03c5eabba8732d596cf4ed42c24bbff1bafd25bb1e502c3dcbc1d65d4dce3dd`.
Its day1/day2/overlap bytes are reused, not duplicated. Source member hashes
and semantic graph expectations remain owned by that manifest.

The reviewed IR `expected.json` SHA-256 is
`da798f47dc4ec8525e5402e8c88f499203fc96e6c0d907aa78f35dbe519aac47`.
`internal/testutil/irfixtures/expectations.go` independently recomputes it
through the selectable admission/expectation test. It uses standard-library
CSV, timestamps and counters, never product normalization, matching, graph or
digest code. Initial facts were independently calculated with Python's CSV and
datetime libraries. Tests compare the frozen output; they never record an
application response as an expectation or rewrite the expected file.

The independent Go builder's SHA-256 at this handoff is
`f7a01d81af0d78395eb03d87672b214256e56646e49d3fb76d0c4878740260ea`.
Successor transport expectations for the existing Network Flow fixture families
were separately recalculated from frozen normalized rows, rejected diagnostics
and mapping facts, using standard UUIDv5, SHA-256 and canonical JSON operations.
The reviewed Graph grammar is
`fixtures/network-flow-orientation/graph-revision-oracle.py`, SHA-256
`2d10ca39b54c7d6ae1041587dc130d57f715f69d6a28252d70b31c7b01b2f3d9`.
The successor calculation changes the source-snapshot domain to v2 and binds each
table to its immutable portable fingerprint instead of its mapping alone. It
does not change orientation, aggregation, row multiplicity or original CSV bytes.
No captured application response supplied these new expected identities.

The revised frozen families are: NF-FIX-001/11, 002/8, 004/9, 006/6, 007/7,
008/6, 009/6, 013/6, 014/7, 015/7, 016/12, 017/6, 018/7, 020/5, 023/6,
024/7, 025/6, 026/8, 027/6, 028/6 and 029/8 (family/revision). Each family's
manifest admits its source, expectation and transcript hashes. The separate IR
and Cisco corpus revisions and all numerical acceptance anchors remain unchanged.

Original CRLF bytes are protected from Git text conversion. The pre-existing
Cisco checkout matched its hashes, but its committed blobs had LF conversion;
the five existing Cisco CSV modifications repair that fresh-checkout issue.
They do not alter the approved orientation or aggregation semantics.

## Requirement and coverage matrix

Each row names its primary owner through its catalog selector. Collaborators are
registered in the authored catalog. Final CI row results below share the source
identity recorded in final qualification. Passing rows within a failed invocation
do not establish complete-owner or release acceptance. W1–W10 refer to the approved
Complete IR portability and fixture remediation plan.

`CI06` denotes `.cartulary/test-results/ir-portability-final-ci06`; individual
results are `rows/<catalog-row-id>.json`. A leading dot in a selector continues
the preceding owner and test-layer prefix.

| Requirement | Workstream | Concrete assertion / catalog row | CI06: passed |
| --- | --- | --- | --- |
| Current versions, closed contracts, no executable Markdown | W1/W10 | `make source-check CHECK=all`; generated and catalog boundaries | CI06: passed |
| Immutable originals, admitted Cisco dependency, independent expectations | W7/W10 | `module.imports.unit.ir_corpus_admission`; `web.architecture.boundary_support.ir_fixture_admission` | CI06: passed |
| 1,500/30/20, exact cells, 1,450 consistent/50 empty time pairs, 222 alias associations | W7 | `module.imports.integration.ir_corpus_core`; `module.incidentbundles.integration.network_flow_portability` | CI06: passed |
| No inferred evidence, relationships, indicators or accounts; ambiguous aliases remain explicit | W7/W9 | Same Core/full-lifecycle rows, public empty inventories and explicit supplement linking | CI06: passed |
| Clipboard 499/500, replay; 501/full-TSV atomic rejection | W7/W9 | `module.timeline.integration.ir_public_clipboard`; `module.timeline.unit.ir_clipboard_boundaries` | CI06: passed |
| Actual upload and mapping for Core plus all three flows | W7 | `module.imports.browser.ir_corpus_import` | CI06: passed |
| Every successful unit has immutable provenance; cleanup, deduplication, authorization and exact replay | W2 | `module.imports.integration.ir_corpus_core`; `.committed_source_transfer`; `module.imports.unit.committed_source_portability` | CI06: passed |
| All five flow families, 4,500 observations, overlap provenance, semantic anchors | W5/W7 | `module.incidentbundles.integration.network_flow_portability`; `module.networkflow.unit.orientation_corpus`; `module.networkflow.integration.orientation_corpus` | CI06: passed |
| Selected history separate from refresh intent; exact reconstruction after deletion/retirement | W3/W5 | `module.networkflow.integration.selected_source_history`; `module.incidentbundles.integration.network_flow_historical_portability` | CI06: passed |
| Import leaves unrelated incident Graph state untouched | W3/W6 | Historical portability row compares unrelated graph before/after | CI06: passed |
| Bounded streams, ZIP/TAR/TAR.GZ, short writes, checksums, archive limits | W4/W9 | `module.incidentbundles.unit.member_stream_archives`; `.member_stream_records`; `module.stagedobjects.integration.allocation_publication_cleanup_7f6ad80724`; `module.incidentbundles.unit.incident_bundle_storage_is_atomic_and_contained_2bfb18f3a9` | CI06: passed |
| Bulk member exceeds 64 MiB without weakening ceilings | W4/W9 | `module.incidentbundles.integration.large_member_public_transfer` plus instrumented stream unit row | CI06: passed |
| Historical actors include extension/provenance-only references; native mutation replaces only its action | W5 | `module.incidentbundles.integration.actor_export_combines_native_users_and_imported_f4f3ed2ecc`; `.extension_portability_matrix_and_atomic_import_1b27fc2d91`; `module.imports.integration.committed_source_transfer` | CI06: passed |
| Prepared bindings, atomic publication, unavailable destination, rollback and lost acknowledgement | W6/W9 | `module.incidentbundles.integration.extension_portability_matrix_and_atomic_import_1b27fc2d91`; `.incident_bundle_import_final_publication_recheck_b7b07b3e93`; `module.crossownertransaction.integration.postgres_atomicity_7f6ad80724`; full-lifecycle fault variants | CI06: passed |
| Evidence replacement/history, pending request, dependent tasks, decision, assessment, indicator, party, note, saved view | W7 | `module.incidentbundles.integration.network_flow_portability`: public supplement and exact semantic inventory | CI06: passed |
| Explicit day-1 graph composition, frozen snapshot/release, retained report inputs/output | W3/W7 | Full-lifecycle API and independent destination/restored browser rows | CI06: passed |
| A → B → operator backup → C, independent keys, session isolation, serving admission | W8 | Full-lifecycle integration and `module.recovery.browser.ir_transferred_investigation` / `.ir_restored_investigation` | CI06: passed |
| Old persisted state/Recovery generations reject before mutation | W1/W8 | `module.database_migrations.integration.production_preflight_state_matrix`; `.production_ddl_v3_recurrence`; `.recovery_target_control_distinguishes_fresh_init_d379ae733e`; `module.recovery.unit.recovery_generation_selection_drives_verificatio_2a8734f884` | CI06: passed |
| Concurrent import, rename, delete, refresh and retirement share one export snapshot | W5/W9 | `module.incidentbundles.integration.network_flow_export_snapshot` | CI06: passed |
| Malformed CSV/UTF-8, duplicate-header ordinal identity and inert documentary text | W9 | `module.networkflow.unit.network_flow_selector_covers_empty_csv_forbidden_421ddebc7c`; `module.networkflow.unit.network_flow_import_facade_publishes_complete_bi_cae079f700`; `module.timeline.unit.timeline_admission_decoding_and_hashing_46ac663c5c` | CI06: passed |
| Missing actors/sources, altered mappings/fingerprints and changed selected graph bindings | W2/W3/W5/W9 | `module.networkflow.unit.portable_member_admission`; `module.imports.unit.committed_source_portability` | CI06: passed |
| Duplicate graph names remain distinct without bypassing active quotas | W3/W5/W9 | `module.networkflow.unit.portable_member_admission`; `module.incidentbundles.integration.network_flow_historical_portability` | CI06: passed |
| Table review and lost-acknowledgement replay retain exact request identity | W7/W9 | `module.networkflow.browser.table_lifecycle_continuity` | CI06: passed |
| Catalog uniqueness, fixture/cache identity, operator dependency and runtime profiles | W10 | `harness.test_catalog.behavior.routing_contract`; CI source-check projections | CI06: passed |
| Final source, compatible owner union, visual review and operational handoff | W10 | Finalization, final portfolio and explicit evidence audits | CI06: security gate failed; complete-owner audit unperformed; final-source UI review passed |

The main lifecycle is one connected incident; independently initialized browser
journeys supply focused destination and restore usability evidence. The comparison
exclusion list is closed: `data/object_blobs.ndjson` excludes only `storage_key`,
with attachment bytes and hashes checked separately through public APIs.
Public flow-table comparison excludes only `source_filename_digest` and
`source_filename_digest_key_id`; it also asserts that the destination regenerates
the keyed digest. Public graph comparison requires transferred `latest_job_id`,
`last_failure_code` and `last_failed_at` to be null. Pagination cursors are
deployment-local continuations; the compared rows and page termination must agree.
All five portable Network Flow members compare byte-for-byte. Stable identities,
timestamps, source hashes, ordering, revisions,
selected graph bindings and retained report files are compared without broad
normalization. The 500 overlapping observations remain separate applications.

## Final qualification and open blockers

`CARTULARY_TEST_RUN_ID=ir-portability-finalize01 make agent-finalize` exited 0;
its retained summary reports no generated changes. Retained warm-check maintenance
was skipped because no compatible complete warm-check root was supplied.
`CARTULARY_TEST_RUN_ID=ir-portability-markdown01 make lint-markdown` exited 0.
The executable source frozen for the broader run is
`sha256:e0072a381171a816b24311aeaacb2a4931c5e113ddd4c24cb05c2b2dd6400f20`.

`CARTULARY_TEST_RUN_ID=ir-portability-final-ci01 make ci` selects 2,432 catalog
rows, including the connected lifecycle, actual upload journey, independent
transfer/restore browser journeys, snapshot-concurrency and 65 MiB transfer.
It terminated with native exit 2 after 2,573,082 ms: 440 work units completed,
404 passed and 36 failed, with no cancelled or skipped units. All 760 cleanup
steps passed. This failed run is retained and does not provide passing acceptance.
Diagnosis identified successor fixture/resource expectations, exact public-port
allowlists, a retained-state admission test, frontend resource fixtures, and two
dependency advisories. These are being repaired against their exact retained
diagnostics; semantic corpus totals and archive ceilings are unchanged.

Focused repair evidence (each command uses `CARTULARY_TEST_RUN_ID=<run ID>`):

| Run ID | Command / selection | Native exit / outcome |
| --- | --- | --- |
| `ir-portability-lint02` | `make lint` | 0; all 14 work units, including frontend/backend boundaries and type checks |
| `ir-portability-bundle-repair01` | Bundle member-stream archive, public API and staging-cleanup rows | 0; 59,087 ms; ZIP/TAR/TAR.GZ integrity cases and compression accounting |
| `ir-portability-db-repair02` | `make test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.recovery_target_control_distinguishes_fresh_init_d379ae733e` | 0; 14,424 ms; predecessor state rejected before control initialization |
| `ir-portability-imports-repair02` | Imports behavior/composition contracts and `module.imports.integration.network_flow_default_limit_100001_rows` | 0; 46,216 ms |
| `ir-portability-nf-repair01` | 29 previously failing Network Flow rows | 2; 181,765 ms; successor expectation families largely passed; four test defects diagnosed |
| `ir-portability-nf-repair02` | Original Cisco fixture, query-normalization fixture and saved-graph lifecycle | 2; 85,690 ms; first two passed; unsupported-state injection still used the validating store API |
| `ir-portability-web-repair01` | `make test-slice OWNER=web.networkflow` | 2; 30,234 ms; 3/4 work units passed; workspace mocks omitted current provenance and selected-source fields |
| `ir-portability-bootstrap02` | `make bootstrap` | 2; harness prerequisite build required the new dependency checksum before bootstrap could run |
| `ir-portability-bootstrap03` | `GOFLAGS=-mod=mod make bootstrap` | 0; Make-owned dependency acquisition updated checksums |
| `ir-portability-web-repair02` | `make test-slice OWNER=web.networkflow` | 0; 25,866 ms; current resources, permissions, saved-graph operations and recovery |
| `ir-portability-nf-repair03` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.saved_graph_lifecycle_v2` | 0; 222,881 ms; explicit unsupported-state injection, historical selections, reporting and reconstruction |

The large import's retained `network-flow-import-source-passes.json` records
about 3.1 seconds for validation and 23.3 seconds for validation plus retained-row
writes. Combined source-read time was below 0.4 seconds. This explains limited
deadline headroom; it is not a throughput qualification. No deadline was raised.

The compression dependency advances from v1.18.7 to the published v1.20.2 fix
for GO-2026-6865. The official GO-2026-6909 advisory, inspected on November 6,
2026, lists no fixed SAML version. The vulnerability gate remains required and
unwaived; final qualification must report this external blocker separately from
portability test results.

`CARTULARY_TEST_RUN_ID=ir-portability-finalize02 make agent-finalize` exited 0
after 26,003 ms and reported zero generated changes. Optional retained warm-check
maintenance was unselected because no compatible complete root was supplied.
The second broader command is
`CARTULARY_TEST_RUN_ID=ir-portability-final-ci02 make ci`, with executable source
`sha256:ad90e8dd4dd002e91bd8a4ad550e0e1ba3e6a4da3b5a749ca5ec12352891b8b6`
and graph
`sha256:e5d5996c1855f9f196281be768aaafd657a5e91258fef229dd840b95a9e128d1`.
It selected 2,432 rows in 440 work units and terminated with native exit 2 after
2,880,698 ms: 437 units passed, three failed, and none were cancelled or skipped.
All 762 cleanup steps passed. The full connected API transfer/restore scenario,
historical graph variants, actual browser import, independent destination and
restored browser journeys, 65 MiB transfer, and 100,001-row import all passed.
These are passing assertions within a failed invocation, not complete acceptance.

The three failures and subsequent dispositions are:

| Failure | Evidence and disposition |
| --- | --- |
| Recovery metadata fixture | The current-generation invalid-semantic-state fixture still seeded extension state 6. It now seeds state 7, preserving the owner-validator call and rejection assertions. `ir-portability-recovery-repair03` ran the exact metadata and crypto-format rows with native exit 0 (42,505 ms). |
| Native browser fixture lifecycle | The owned Vite process could not bind port 19002. Its live cleanup supervisor caused the failure to be reported as a timeout. The readiness classifier now reports the definite bind failure immediately as `resource_conflict`; `ir-portability-bind-diagnostic02 make harness-contract SUITE=lifecycle` passed (native exit 0, 7,360 ms). The exact native lifecycle row passed in `ir-portability-native-lifecycle02` (native exit 0, 140,121 ms). No startup retry or deadline increase was added. |
| Vulnerability gate | Compression advisory GO-2026-6865 is resolved. GO-2026-6909 remains the sole blocking advisory, with no published fixed SAML version in the official advisory. The gate is not waived. |

The exact competitor for port 19002 was not established. The host's observed
ephemeral range, 10240–32767, overlaps the configured browser range, so an outbound
socket is a possible explanation rather than a proved cause. No host setting or
foreign process was changed. The later pass does not establish that environmental
collisions are impossible. Failed selection attempts using retired
`harness-smoke-web-e2e-lifecycle` and unknown `test-status` returned Make exit 2
before test execution; current public targets were used for the retained checks.

The CI plan excludes explicit-only visual and release-measurement rows. This task
does not claim new throughput or latency qualification. Changed rendered behavior
is separately reviewed through the UI-review workflow. Complete-owner audits must
identify any such omitted partitions rather than treating CI selection as complete
owner acceptance.

The last catalog cleanup replaced the misleading
`module.networkflow.integration.saved_graph_cutover_v6` selector with
`module.networkflow.integration.saved_graph_current_admission`, preserving its
assertions. `ir-portability-generate74 make generate` passed with native exit 0
(10,400 ms); `ir-portability-format101 make format` passed with native exit 0
(6,840 ms). `ir-portability-finalize03 make agent-finalize` passed with native
exit 0 (25,972 ms), zero generated changes, and optional retained warm-check
maintenance unselected. No complete compatible warm-check root was available.

`CARTULARY_TEST_RUN_ID=ir-portability-final-ci03 make ci` uses source
`sha256:9b3c77fe2c08d04633fb8b74408421d28f664890ac2843baf04cca620cb9332d`
and graph
`sha256:f93337dd58b3a50b6eed9dff9a837ea763cced5b4ae452ae8149023fa7ed7519`.
It selected 2,432 rows in 440 units. Manual normative review found that portable
graph preparation rejected duplicate display names, although Network Flow permits
them. The verified native session was interrupted before editing executable
source. It terminated with native exit 130 after 1,023,880 ms: 79 units passed,
361 were cancelled, and none failed. These partial results do not qualify the
subsequent source.

Cleanup recorded 169 steps, with two failures and one blocked dependency:
browser lease `lease-000038` release and provider close failed with
`harness/cleanup_error` / `unsafe`; service close was blocked. The exact retained
run has no published private runtime root/lease identity required by
`test-run-recover`. No path scan, PID inference or broad reset was used to guess
that authority. Cleanup remains unresolved; the original evidence is preserved.

The duplicate-name restriction is removed while the active-graph quota remains.
Regression coverage includes two distinct unmaterialized declarations with the
same name, quota rejection, and public transfer/re-export of two active selected
graphs with the same name alongside a retired selection. Formatting passed in
`ir-portability-format102` (native exit 0, 6,933 ms). Both focused commands passed
with native exit 0:

| Run ID | Exact command after `CARTULARY_TEST_RUN_ID=<id>` | Outcome |
| --- | --- | --- |
| `ir-portability-duplicate-graphs-unit01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.portable_member_admission` | Pass, 10,537 ms; same-name declarations stay distinct and the quota still rejects excess active declarations |
| `ir-portability-duplicate-graphs-transfer01` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.network_flow_historical_portability` | Pass, 25,664 ms; public transfer/re-export retains same-name selected graphs, retired history, diagnostics and unrelated incident isolation |

`CARTULARY_TEST_RUN_ID=ir-portability-finalize04 make agent-finalize` passed
with native exit 0 (26,290 ms), zero generated changes and no supplied compatible
complete warm-check root. Optional retained-evidence maintenance was unselected.
The source was then frozen at
`sha256:6ec7e2aaf18ba9cbde15c7c896dcf72e57682cfaeced5386dfa4a4ea997e3f37`.
`CARTULARY_TEST_RUN_ID=ir-portability-final-ci04 make ci` selects 2,432 rows in
440 work units, graph
`sha256:f93337dd58b3a50b6eed9dff9a837ea763cced5b4ae452ae8149023fa7ed7519`.
It terminated with native exit 2 after 2,962,353 ms: 436 work units passed,
four failed, none were skipped or cancelled. All 761 cleanup steps and the
retained-secret scan passed. The four failed units represent two causes:
GO-2026-6909 and the table-lifecycle browser assertion; the latter also failed
both browser target-summary units. The original Cisco browser row was skipped
as a dependency consequence, not as an independent failure.

The table-lifecycle trace records only the peer rename and normalized `Café`
rename, both successful. The next `Recovered table` PATCH was never dispatched:
the captured UI retained a version-3 draft requiring review after the previous
rename's delayed invalidation. The test now observes that exact table's rename
notification through the existing owned socket monitor before opening the next
draft. No deadline, retry, review safeguard or product behavior changes.
`ir-portability-format103 make format` passed (native exit 0, 7,068 ms).

Passing lifecycle assertions within this failed invocation are:

| Catalog row | Terminal state / assertion duration |
| --- | --- |
| `module.incidentbundles.integration.network_flow_portability` | Passed / 256,670 ms; complete A → B → operator backup → C and exact semantic comparisons |
| `module.incidentbundles.integration.network_flow_historical_portability` | Passed / 4,800 ms; same-name graphs, deletion, retirement, diagnostics and unrelated incident isolation |
| `module.incidentbundles.integration.network_flow_export_snapshot` | Passed / 2,420 ms |
| `module.incidentbundles.integration.large_member_public_transfer` | Passed / 6,010 ms |
| `module.imports.browser.ir_corpus_import` | Passed / 46,113 ms |
| `module.recovery.browser.ir_transferred_investigation` | Passed / 3,066 ms |
| `module.recovery.browser.ir_restored_investigation` | Passed / 2,651 ms |

These row results do not turn the failed invocation into passing complete-owner
or release evidence. The browser synchronization repair changes executable test
source and requires a new compatible final portfolio.

`CARTULARY_TEST_RUN_ID=ir-portability-table-continuity-repair01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser.table_lifecycle_continuity,module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179`
passed with native exit 0 (103,646 ms). The repaired lost-acknowledgement scenario
and the previously dependency-skipped original Cisco upload/graph scenario both
passed, without changing production behavior, deadlines or expectations.

`CARTULARY_TEST_RUN_ID=ir-portability-finalize05 make agent-finalize` passed
with native exit 0 (26,375 ms), zero generated changes, and optional retained
warm-check maintenance unselected because no compatible complete root was
available. Final source is frozen at
`sha256:cfe754b4e26e2961eff03115a892f47a859ca276cd1319d6b71951e2f3b399cf`.
`CARTULARY_TEST_RUN_ID=ir-portability-final-ci05 make ci` terminated with native
exit 2 after 2,785,774 ms: 436 work units passed, four failed, none were skipped
or cancelled. All 558 cleanup steps and the retained-secret scan passed. The
four failed units represent GO-2026-6909, the IR upload browser cohort and its
two propagated target summaries. The IR browser row failed after 47,322 ms;
the table-lifecycle and original Cisco browser rows were dependency-skipped.

The exact retained trace shows overlap apply admitted at 11:19:08.891 UTC,
running polls, and successful completion observed at 11:19:12.978 UTC. The
five-second panel-absence assertion covered approval, selection, apply and
workspace handoff together. The repaired test observes the unique imported
table using the unchanged journey deadline, then verifies panel closure. It
does not change product behavior, expected records or whole-test limits.

The full A → B → backup → C API lifecycle passed (240,580 ms), as did historical
transfer (5,510 ms), snapshot consistency (2,820 ms), the 65 MiB public transfer
(6,640 ms), and independent transferred/restored browser assertions
(3,097 / 3,084 ms). These rows are passing assertions inside a failed invocation,
not complete-owner acceptance.

`CARTULARY_TEST_RUN_ID=ir-portability-upload-handoff-repair01 make test-slice OWNER=module.imports ROWS=module.imports.browser.ir_corpus_import`
passed with native exit 0 (124,057 ms; browser assertion 45,804 ms). All 23
cleanup steps completed. The real upload/mapping/apply journey and all final
Core/Network Flow record checks passed. Source identity was
`sha256:7248d77db4eb53fe8df7edafb95b17abb2fc3335e8aad34689c10f8ba25b5651`.
`CARTULARY_TEST_RUN_ID=ir-portability-format104 make format` also passed
(native exit 0, 6,928 ms), followed by a passing `git diff --check`.

`CARTULARY_TEST_RUN_ID=ir-portability-finalize06 make agent-finalize` passed
(native exit 0, 26,278 ms), with zero generated changes and no compatible complete
warm-check root supplied. Optional retained-evidence maintenance was unselected.
The source remains
`sha256:7248d77db4eb53fe8df7edafb95b17abb2fc3335e8aad34689c10f8ba25b5651`.
`CARTULARY_TEST_RUN_ID=ir-portability-final-ci06 make ci` terminated with native
exit **2** after **2,722,513 ms** (45 minutes 23 seconds). The source above is
unchanged; its CI graph is
`sha256:f93337dd58b3a50b6eed9dff9a837ea763cced5b4ae452ae8149023fa7ed7519`.
All **2,432 selected rows passed**. Work-unit totals are **439 passed, one failed,
zero skipped and zero cancelled**; cache accounting is 205 hits, six misses and
229 bypasses. All **558 cleanup steps completed**, and the retained-secret scan
passed across 6,877 files. Cached passes are not fresh timing measurements.

The sole failed unit is `operation:harness.go_vulncheck`. Its retained findings
are at
`.cartulary/test-results/ir-portability-final-ci06/unit-artifacts/operation-harness.go_vulncheck/govulncheck-findings.json`.
They identify GO-2026-6909 as symbol-reachable through
`ProductionSAMLVerifier.VerifyACS` → `ServiceProvider.ParseXMLResponse`, with
no fixed version recorded. The [upstream report](https://github.com/golang/vulndb/issues/6909)
still lists no fixed version when rechecked on 2026-11-06; the latest upstream
commit remains `346540312f721498fc75e69637d9250dd89f230b` from 2025-05-09.
No waiver, authentication disablement or local SAML fork was introduced.

Final-source lifecycle assertions are:

| Catalog row | Terminal state / assertion duration |
| --- | --- |
| `module.incidentbundles.integration.network_flow_portability` | Passed / 240,370 ms; full corpus, supplement, A → B transfer, B → backup → C restore, exact public semantic comparisons and session isolation |
| `module.incidentbundles.integration.network_flow_historical_portability` | Passed / 4,820 ms; diagnostics, same-name declarations, renamed/deleted source, retired selection and unrelated incident isolation |
| `module.incidentbundles.integration.network_flow_export_snapshot` | Passed / 2,430 ms |
| `module.incidentbundles.integration.large_member_public_transfer` | Passed / 6,010 ms; 65 MiB member |
| `module.imports.browser.ir_corpus_import` | Passed / 49,774 ms; original Core and all three Cisco uploads |
| `module.networkflow.browser.table_lifecycle_continuity` | Passed / 4,610 ms; exact request continuity and lost-acknowledgement replay |
| `module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179` | Passed / 14,853 ms |
| `module.recovery.browser.ir_transferred_investigation` | Passed / 3,158 ms |
| `module.recovery.browser.ir_restored_investigation` | Passed / 3,068 ms |

Every final row result is retained under
`.cartulary/test-results/ir-portability-final-ci06/rows/<catalog-row-id>.json`.
The invocation's `run-manifest.json`, `run-summary.json`, `cleanup-results.json`
and `retained-secret-scan.json` provide source, outcome and cleanup identities.
Complete-owner evidence audits remain unperformed: this failed invocation cannot
be promoted into a compatible passing root. Explicit-only visual and release
measurement partitions were not selected; rendered changes received the separate
artifact review below, and no throughput qualification is claimed.

Final-source artifact-mode review used `make ui-review UI_MODE=artifacts`, two
`CARTULARY_OUTPUT_MODE=machine make ui-capture UI_SESSION=<printed-session> UI_REQUEST=<private-request>`
operations and `make ui-review-stop UI_SESSION=<printed-session>`. Each command
and the owning foreground process exited 0. Its retained terminal receipt is
`.cartulary/test-results/20261106T121739Z-p73431/ui-review/terminal.json`, with
`cleanup=complete` and two inspected images. The exact CI06 transferred graph
and restored report captures were selected from their passed row/cohort and
test-authored output paths; original byte lengths and hashes were verified before
viewing. At 1280×720, the retained graph is labeled “Selected result”, and the
report links are readable with a visible focus indicator. This was image-only
inspection: no DOM/axe or general accessibility claim, and no golden refresh.
Private imported images and caller-owned request files were cleaned up.

Documentation validation used
`CARTULARY_TEST_RUN_ID=ir-portability-final-markdown06 make lint-markdown`
(native exit 0, 24,030 ms), and `git diff --check` passed. The final phase-status
annotation is rechecked with
`CARTULARY_TEST_RUN_ID=ir-portability-final-markdown07 make lint-markdown`;
its exact result is retained at
`.cartulary/test-results/ir-portability-final-markdown07/adhoc/lint-markdown/tool-run-summary.json`.
These are documentation checks and do not waive the CI security finding.

The artifact-mode UI review at
`.cartulary/test-results/20261106T090439Z-p12306/ui-review/terminal.json`
closed with native exit 0 and `cleanup=complete`. It inspected the exact
transferred-investigation graph and report captures from `final-ci02`; original
image lengths and hashes were verified before viewing. The review established
readability and selected-result labeling for those states, with the image-only
limitations stated above. Private image imports and caller request files were
removed through their owners; no golden or source capture was modified.

## Phase exits

| Phase | Final-source evidence and remaining admission |
| --- | --- |
| 0 — Contract closure | Successor owner amendments, authored contracts, generated projections and current-format source checks are implemented; selected checks passed. Human review of normative fidelity remains a review responsibility. |
| 1 — Durable state and streaming | Committed provenance, selected history, fresh-state admission and bounded member-stream assertions passed. |
| 2 — Complete portability | Public full-corpus A → B transfer, all retained families, historical variants and atomic-publication/failure assertions passed. Network Flow no longer declares a portability block. |
| 3 — Connected lifecycle | Investigation, immutable reporting, actual operator backup/restore to C and independent destination/restored browser assertions passed. |
| 4 — Adversarial qualification | Selected all-container rejection, resource bounds, 65 MiB transfer, cancellation/replay and coherent-snapshot assertions passed. The separate SAML dependency security gate remains failed. |
| 5 — Finalization and handoff | Finalization, all selected test rows, current-run cleanup, retained-secret scan and artifact UI review passed. Overall CI and complete-owner acceptance remain open on the security blocker; the earlier cancelled-run cleanup evidence remains unresolved. |

These entries distinguish passing assertions from overall acceptance. The complete
plan is not closed by selecting tests or by combining historical passing runs.

## Historical fixture-only verification record

All named run roots are under `.cartulary/test-results/` in this checkout.
Each contains the harness's retained summaries and referenced logs. A Make
exit of 2 below is the observed outer command failure; normalized failure
classification remains in the retained summary.

Focused development passes before finalization:

| Run ID | Exact command after `CARTULARY_TEST_RUN_ID=<run-id>` | Outer exit |
| --- | --- | ---: |
| `ir-core-20261105-04` | `make test-slice OWNER=module.imports ROWS=module.imports.unit.ir_corpus_admission,module.imports.integration.ir_corpus_core` | 0 |
| `ir-clipboard-20261105-01` | `make test-slice OWNER=module.timeline ROWS=module.timeline.unit.ir_clipboard_boundaries` | 0 |
| `ir-orientation-20261105-01` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.orientation_corpus,module.networkflow.integration.orientation_corpus,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a` | 0 |
| `ir-support-20261105-01` | `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.ir_fixture_admission,web.architecture.boundary_support.node_fixture_admission_seals_verified_inputs_and_d6cec88240` | 0 |
| `ir-browser-20261105-03` | `make test-slice OWNER=module.imports ROWS=module.imports.browser.ir_corpus_import` | 0 |
| `ir-harness-20261105-01` | `make harness-contract SUITE=contract` | 0 |
| `ir-schema-20261105-02` | `make source-check CHECK=schema` | 0 |
| `ir-catalog-20261105-01` | `make source-check CHECK=catalog` | 0 |

The earlier fixture-only portfolio below used executable source digest
`sha256:1e2c0cc233a8e5592633f065f8ce75019f53be45bdf9af33993256e3b33dafc9`.
It predates the production portability implementation. Its passes establish
neither successor-contract acceptance nor complete incident transfer and restore.
The current qualification record appears above; these historical results remain
separate and are not substituted for missing final-source evidence.

| Run ID | Exact command after `CARTULARY_TEST_RUN_ID=<run-id>` | Outer exit / outcome |
| --- | --- | --- |
| `ir-finalize-20261105-04` | `make agent-finalize` | 0 / pass; generated files unchanged |
| `ir-final-core-20261105-02` | `make test-slice OWNER=module.imports ROWS=module.imports.unit.ir_corpus_admission,module.imports.integration.ir_corpus_core` | 0 / pass |
| `ir-final-flow-20261105-02` | `make test-slice OWNER=module.networkflow ROWS=module.networkflow.unit.orientation_corpus,module.networkflow.integration.orientation_corpus,module.networkflow.unit.fixture_admission_and_materialization_preserve_i_7c4b8b432a` | 0 / pass |
| `ir-final-clipboard-20261105-02` | `make test-slice OWNER=module.timeline ROWS=module.timeline.unit.ir_clipboard_boundaries` | 0 / pass |
| `ir-final-support-20261105-02` | `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.ir_fixture_admission,web.architecture.boundary_support.node_fixture_admission_seals_verified_inputs_and_d6cec88240,web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19` | 0 / pass |
| `ir-final-harness-20261105-02` | `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.behavior.scheduler_contract` | 0 / pass |
| `ir-final-browser-20261105-01` | `make test-slice OWNER=module.imports ROWS=module.imports.browser.ir_corpus_import` | 0 / pass |
| `ir-final-types-20261105-02` | `make lint CHECK=types` | 0 / pass |
| `ir-final-go-lint-20261105-02` | `make lint CHECK=go` | 0 / pass |
| `ir-final-frontend-lint-20261105-02` | `make lint CHECK=frontend` | 0 / pass |
| `ir-final-generated-20261105-02` | `make source-check CHECK=generated` | 0 / pass |
| `ir-final-schema-20261105-02` | `make source-check CHECK=schema` | 0 / pass |

The named final manifests, terminal summaries, cleanup results and retained
secret scans were inspected. These checks report no cleanup failures, no
retained-secret failures and no cache hits. Finalization did not receive a
`RESULTS_DIR`; its optional retained-evidence and scheduler-drift audits were
therefore unselected. The complete-owner evidence audit remains unperformed.

Generation used `CARTULARY_TEST_RUN_ID=ir-generate-20261105-02 make generate`
(exit 0). Earlier `ir-final-*-20261105-01` passes preceded the cache regression
test-setup repair and are superseded by the final roots above. The repaired
scheduler test also passed in `ir-harness-repair-20261105-01` before the last
finalization. Repetition was needed to establish a common final source after
that repair; the unchanged expensive portfolios were not widened.

Documentation validation used
`CARTULARY_TEST_RUN_ID=ir-final-markdown-20261105-01 make lint-markdown`
(exit 0, pass). Its summary is
`.cartulary/test-results/ir-final-markdown-20261105-01/adhoc/lint-markdown/tool-run-summary.json`.
`git diff --check` also passed. These are documentation/working-tree checks,
not additional product or lifecycle evidence.

Diagnosed development failures, retained rather than deleted:

| Run | Target / outer exit | Diagnosis and correction |
| --- | --- | --- |
| `ir-core-20261105-01` | `test-slice` / 2 | Tiny-example five-second apply wait and read-only temporary cleanup. Full-corpus wait now uses the existing harness deadline; retained copies avoid temporary cleanup ownership. |
| `ir-core-20261105-02` | `test-slice` / 2 | Independent Go CSV reader folded quoted CRLF. Product retained the source correctly; oracle now preserves quoted CR. |
| `ir-browser-20261105-01` | `test-slice` / 2 | Mapping locator did not select the accessible combobox. Corrected role/name selector. |
| `ir-browser-20261105-02` | `test-slice` / 2 | Five-second assertion expired while Timeline apply ran. Assertion now waits within the unchanged whole-test budget. |
| `ir-schema-20261105-01` | `source-check` / 2 | Redundant source-policy assertion itself contained a restricted documentation-root literal. Removed; source-policy enforcement remains. |
| `ir-types-20261105-01` | `lint CHECK=types` / 2 | Strict indexed-access errors in new fixture support/browser test. Added tuple types and explicit guards. |
| `ir-types-20261105-02` | `lint CHECK=types` / 2 | Removed constant was still used for Network Flow produced-artifact names. Restored owner-local validation constant. |
| `ir-finalize-20261105-01` | `agent-finalize` / 2 | Generated topology inputs were stale after the cache implementation changed. Ran `make generate`; finalization subsequently passed. |
| `ir-final-harness-20261105-01` | `test-slice` / 2 | Cache regression setup attempted to compile an unregistered synthetic row. Changed only the test setup to supply the synthetic resolver unit and catalog directly. |
| `20261105T204412Z-p52421` | `generate` / 2 | New clipboard row used the wrong family ID. Corrected to the Timeline unit family; subsequent generation passed. |

An earlier generation attempt rejected unsupported native-schema regex syntax
before publishing a run root. The schema now uses the supported confined-path
pattern. Selection errors (unknown owner `module.network_flow`, incorrect
diagnostic target/detail combinations) were corrected before execution; they
are not product failures or passing evidence.

## Operational handoff

No live developer or production environment was reset. Application tests use
harness-owned databases and object stores. No live deployment or database cutover
was performed. The successor implementation has no replaced-format reader or
retained-data migration, backfill or conversion layer; generic extension blocking remains legitimate for
profiles without a participant.

For a new deployment, provision separate empty database and object-store bindings,
run the existing database administrator/provisioning and migration entrypoints
with their purpose-specific credentials, and start the successor executable with
compatible Network Flow support explicitly enabled. For local initialization,
the [development loop](../guides/cartulary-dev-guide.md#73-local-development-loop)
uses `make services-up`, `make db-migrate`, and `make dev` against those separately
selected bindings. Production initialization uses the existing `migrate up`
entrypoint; no reset command is part of this cutover. Complete the existing
bootstrap/MFA and serving-admission workflow before incident import. Do not point
these initialization steps at a prior retained deployment. Its executable and
storage remain together for separately managed historical access. State 7 has no
in-place downgrade; rollback means returning to that separately retained prior
deployment, not opening successor data with an older executable.

Bundle import requires destination deployment administration and compatible
profiles. It preserves historical attribution without source accounts, roles,
credentials, memberships, operational jobs or receipts. Exact raw source bytes,
including unmapped content, are retained and disclosed only under their source
owner's inspection authorization. Optional retained reporting artifacts follow
the existing disclosure policy; reveal maps are excluded.

Recovery requires an admitted empty restore target, authorized backup decryption
keys, purpose-specific operator configuration and the current generation. It
preserves its deployment inventory and invalidates old sessions. After successful
source claim, Graph reconstruction and job reconciliation govern their separate
availability gate; unrelated admitted operations may already be available.
Qualification waits for exact selected graphs without requesting refresh. Operator backup/restore
has been exercised in the connected public scenario and independent browser
fixtures; neither database copy nor direct SQL reseeding supplies that evidence.
Use the existing [operator recovery procedure](../guides/cartulary-dev-guide.md#125-backup-restore-and-failure-modes)
for `operator backup create`, `operator restore latest`, and exact operation-ID
replay after a lost acknowledgement.

The final selected test portfolio passed on the frozen source, while CI remains
failed on the SAML vulnerability gate. Completing acceptance requires resolving
that dependency security blocker, requalifying any resulting source changes and
performing the compatible evidence audit. The earlier cancelled run's unresolved
cleanup evidence remains separately recorded. Network Flow export, exact selected
history, reporting, transfer and restore now have passing final-source assertions;
the complete plan is not represented as accepted while these blockers remain.
