# links Module Refactoring Tracker and Handoff

## Active remediation execution — 2026-11-09

The maintainer explicitly approved implementation of the Links remediation plan on
2026-11-09, including the Core 02 merge clarification and Reporting assignment
inclusion. This execution supersedes the historical planning-only restrictions,
readiness decisions, and deferrals below; historical evidence remains attributed
to its original session. Baseline: main at
`a43248fbe3f762b0f448993126f74172f597f101`; no staged or tracked changes;
this tracker was the only untracked file (SHA-256
`aaa0882665e6db8a4d623dbabae08547e8b7360feaae98f4086ad1cd4cf136af`).

Each slice is a separate workstream. Update this table and the execution log after
validation and before starting the next slice. Product tests and generators must
not consume this tracker or any specification prose. No production DDL, public
wire shape, bundle version, or canonical history format change is planned.

| Slice / workstream | Findings | Dependencies | Status | Exit criteria |
| --- | --- | --- | --- | --- |
| S-08 Specification and tracker cleanup | F-04, F-06, F-12; F-05/F-07–10 no action | baseline | DONE | Owner clarifications adopted, stale restrictions superseded, Markdown lint passes. |
| S-00 Baseline and characterization | F-01/F-02/F-04/F-06 | S-08 | DONE | Passing structural baseline and owner-aligned defect reproductions. |
| S-06 Remove redundant inverse metadata | F-02 | S-00 | DONE | Shared unused field and all four producer mappings removed; rollback field/atomicity checks pass. |
| S-01 Ordinary Records access | F-01 | S-00 | DONE | Ordinary validation uses Records, preserving active/type/incident/errors and transaction/lock semantics. |
| S-02 Rollback Records access | F-01 | S-01, S-06 | DONE | Retained/deleted endpoint lookup uses Records; rollback/history/events pass. |
| S-03 Exact SQL guards | F-11 | S-01, S-02 | DONE | Exact remaining statements declared; dynamic predicate removed; negative/positive guard, schema and generation checks pass. |
| S-04 Merge correction | F-04, F-12 | S-00, S-01, adopted clarification | DONE | Retained provenance and distinct attribution survive import/merge/dedup/inverse; ordinary admission remains narrow. |
| S-05 Reporting correction | F-06 | S-00, S-08 | DONE | Assignments contribute exact support output; capture remains immutable and disclosure fails closed. |
| S-07 Fixture ownership cleanup | F-03 | S-00 | DONE | Links receipts own links/tags; aggregate mention validation retained; versioned fixtures regenerated and lifecycle verified. |
| S-09 Validation and handoff completion | all | all preceding slices | DONE | Finalize, user-directed targeted portfolio, complete Links evidence audit, preservation review and Markdown lint passed. |

F-12 (new): baseline merge recreation substituted owner for creator and decision
time for creation time. Preserve each original field independently; record the
merge actor/time in the change set. This is implemented with F-04, not a history
rewrite. F-02 is resolved by removing the unused non-row `ChangedFieldKeys`
contract, not moving its mapping. F-03 removes Links' duplicate mention accounting
while preserving application aggregate validation. F-11 uses exact normalized SQL
statements in authored boundary contracts. F-06 includes all eligible assignment
links, as confirmed by the maintainer. No new Links subsystem specification,
frontend split, generic repository, compatibility shim, or transaction owner is
introduced. Domain §12 already points to the correct behavior owners and needs no
vocabulary amendment.

The adopted plan's rationale is responsibility cohesion: Records owns envelopes,
Links owns canonical relationship effects, Revisions derives public row deltas,
Reporting owns support/disclosure adaptation, and application test support owns
aggregate fixture validation. Retained history and snapshots stay immutable;
corrected behavior applies to new merges/captures. Fixture identities change and
incompatible cached fixtures are rebuilt rather than translated.

### Execution evidence and checkpoints

| Checkpoint | Changes / decision | Commands and result | Evidence / next gate |
| --- | --- | --- | --- |
| Start | Revalidated baseline and instructions; recorded implementation authorization and owner clarification adoption. | Git status/rev-parse; Make help; task-guide selection started. | Historical RB-005 tracker-only lint restriction superseded. No implementation pass claimed. |
| S-08 complete | Core 00 records maintainer adoption; Core 02 separates retained merge admission and preserves attribution; Reporting explicitly includes assignment links. Domain owner pointers already correct. | `make lint-markdown` PASS (exit 0). | `.cartulary/test-results/20261109T161735Z-p86526/adhoc/lint-markdown/tool-run-summary.json`; begin S-00. |

S-00 completed: structural runs `links-s00-baseline-20261109b` (Links T-09/10/11/13),
`links-s00-records-20261109` (V-10), `links-s00-revisions-20261109` (V-11/12,
transaction failure, indicator children and merge rollback), and
`links-s00-unit-20261109` (T-01–08) passed with outer exit 0. All run roots are
under `.cartulary/test-results/<run-id>` with `run-summary.json`. The initial
OWNER+SUITE invocation was rejected before tests (exit 2); current Make requires
exactly one selector. Use OWNER plus exact ROWS, not the historical V-02 form.

Expected regression failures: `links-s00-defects-20261109` proves both assignment
support lists contain zero instead of three references. Its merge setup first
failed source timestamp admission; corrected the fixture to the required
PostgreSQL `+00:00` byte form. `links-s00-merge-repro-20261109` then proves retained
import/system/rollback are rejected and manual/auto_match overwrite creator/time.
Both failed runs exited 2 with product/test_assertion_failure; cleanup passed.
These owner-aligned regressions remain red until S-04/S-05, and are excluded from
intermediate structural selections rather than skipped or weakened. RB-001's
characterization gate is closed; RB-002/RB-003 have reproduced required outcomes.
Next workstream: S-06. S-06 compilation exposed a fourth producer in
Entities' Host/Identity collection provider, omitted by the original plan's
three-producer inventory. Include its unused mapping removal atomically while
retaining its identifier-claim preparation test under an accurate name.
Initial S-06 Revisions/mention runs failed compilation after an edit left a
stray statement; repaired it. The next Revisions run exposed this fourth
producer. These are change-related failures, not environmental blockers;
retained roots are `links-s06-revisions-20261109`, `links-s06-mentions-20261109`,
and `links-s06-revisions-fixed-20261109`. Indicator admission and repaired mention
unit selections passed. Requalify Revisions after completing the contract update.

S-06 complete: removed unused inverse field and four provider mappings, plus
obsolete private-mapping tests; retained claim preparation under its accurate
name and updated authored selectors. `make generate` passed at
`20261109T162559Z-p13302`; `make format` passed at `20261109T162350Z-p3405`.
Revisions selection `links-s06-complete-20261109` passed (exit 0), as did
`links-s06-collections-20261109`, `links-s06-mentions-fixed-20261109`, and
`links-s06-indicators-20261109`. No wire/data migration. Public delta calculation
remains in Revisions. Next workstream: S-01.

S-01 complete: ordinary endpoints, supersedes type checks, and collection targets
use Records `LoadEnvelopesTx(..., false)` in the caller transaction. Links retains
incident, deletion, and type eligibility. No constructor/locking/wire change.
`make test-slice OWNER=module.links ROWS=` T-09,T-10 reported PASS at
`links-s01-20261109/run-summary.json`; native outer exit was lost at context
compaction, so this is summary evidence only, superseded by S-02/final execution.
No standalone Records SQL remains in store/collection validation. Next: S-02.

S-02 complete: inverse description batches retained envelopes through Records with
no locks; deleted endpoints remain eligible, absent/foreign endpoints remain
`ErrTargetNotFound`. `make task-guide OWNER=module.links` and
`make task-guide OWNER=module.revisions` refreshed selection. T-09/T-10 passed
at `links-s02-20261109`; the five-row Revisions portfolio passed at
`links-s02-revisions-20261109` (both outer exit 0, `run-summary.json`). This also
requalifies S-01. No wire/data migration. Next: S-03.

S-03 first qualification: schema passed (`links-s03-schema-20261109`, exit 0).
Boundary lint failed (`links-s03-boundary-20261109`, exit 2): new import-to-merge
test needs an exact source-port test permission. Contract suite failed
(`links-s03-contract-20261109`, exit 2): 105/106 boundary assertions passed,
including the new SQL matrix; two existing runner imports bypass the Test Catalog
facade (confirmed present at baseline HEAD). Repair those two imports through the
existing facade, without weakening policy. Cleanup passed. S-03 remains open.

S-03 second qualification: boundary lint, exact SQL matrix, schema, generation and
field-reference integration passed. Contract run `links-s03-contract-fixed-20261109`
failed its routing row (exit 2): baseline smoke membership includes a fourth
Markdown assurance row, contrary to adopted TH-HARNESS-REQ-053 and unchanged
assertions. Remove only that smoke membership; retain its full aggregate and
regression coverage. This failure was also present in the first contract run.
No assertion is weakened. Requalify the required contract suite after generation.

S-03 complete: replaced five Links path exemptions with three exact normalized
statement declarations; fixed the optional field-reference predicate. Checker
retains identifiers/literals/operators/parameters/statement boundaries, accepts
comments/formatting, and rejects changed/extra statements, stale grants, dynamic
composition, Records writes and locks. Links-owned writes remain permitted.
`make harness-contract SUITE=contract` passed at `links-s03-contract-closed-20261109`;
`make lint CHECK=backend-boundary` passed at `links-s03-boundary-fixed-20261109`;
`make source-check CHECK=schema` passed at `links-s03-schema-20261109`;
`make source-check CHECK=generated` passed at `links-s03-generated-closed-20261109`;
T-09/T-10 passed at `links-s03-fields-20261109` (all outer exit 0).
`make generate` refreshed validators/topology at `20261109T164003Z-p24697`.
RB-004 is closed. Two baseline harness defects were repaired without widening
policy or weakening assertions. No product migration. Next: S-04.

S-04 focused Links and actual Entities merge-route/store tests passed at
`links-s04-20261109` and `links-s04-entities-20261109` (exit 0), including the
post-tombstone failure injection. `links-s04-revisions-20261109` failed (exit 2,
product/test_assertion_failure) because the new timestamp fixture lacked an
explicit timestamptz cast before interval arithmetic. Transaction failure tests
passed; repair the fixture and rerun merge inverse. Cleanup passed.

S-04 complete: merge validates canonical retained values separately from active
endpoint eligibility. Recreation preserves owner, creator, decision/creation times,
provenance, confidence and field key. Full tuple collisions and self-links retain
deterministic tombstone behavior. Expanded import-to-merge coverage exercises all
five provenances, legal null/boundary confidence cases, rejects illegal tuples,
checks persistence/canonical metadata, and applies inverse mutations. The real
Entities route verifies an injected recreation failure rolls back sources,
projections, versions and history, then successfully retries the same request.
Whole-change-set inverse restores exact system/rollback link rows.
`links-s04-revisions-fixed-20261109` passed (exit 0); together with the preceding
Links/Entities passes and transaction-failure row this closes S-04/RB-002.
Future merges change intentionally; ordinary provenance admission, existing
history, bundle v8, wire schemas and DDL remain unchanged. Next: S-05.

S-05 complete: all active eligible link types now contribute support references,
including the six assignment fields. Adapter tests preserve order/repetitions,
logical targets/fallback, non-null empty output and all-or-nothing failures.
Production adapter tests prove repeatable-read capture and live deletion
separation. Snapshot/replay integration replaces the obsolete exclusion assertion
with an exact indexed assignment target and unchanged captured fields after live
assignment deletion. Disclosure tests prove target restrictions propagate and a
missing recipient fails without partial output; support does not grant access.
`links-s05-adapter-20261109` (T-11) and `links-s05-reporting-20261109` (six-row
adapter/materialization/frozen payload/redaction/snapshot-replay portfolio) passed
with outer exit 0. RB-003 closed. New snapshots intentionally gain references;
retained snapshots/releases and derivation profile remain unchanged. Next: S-07.

S-07 complete: Links fixture expectations, SQL and receipts contain links/tags
only; obsolete mention-bearing descriptor shapes fail admission. All three
authored profiles and dependent contributions use `links.timeline_associations.v2`;
Make regenerated their descriptors and source-contract identities. Aggregate
mention semantics are unchanged. A negative returned mention count fails aggregate
validation after all owner checks, without a semantic success digest. Retained
artifacts with incompatible contribution versions fail admission.
`make generate` passed at `20261109T164946Z-p92846`; the five-row fixture portfolio
passed at `links-s07-fixtures-20261109`, and generated drift passed at
`links-s07-generated-20261109` (outer exit 0). The lifecycle run built one fresh
production-owner snapshot using its bounded profile, read it back, created four
isolated clones, and cleaned them. Do not translate or reuse old receipt/key
artifacts; future profile consumers build under the new generated identity.
No product data or fixture semantic dimensions changed. Next: S-09, finalize
then freeze executable inputs and run CI; add only uncovered required obligations.

### S-09 coverage and disposition review

Finalization passed at `20261109T165115Z-p10921` (outer exit 0), with no generated
files changed. Schema, catalog and tier validation passed. Retained-run maintenance
was not selected: no compatible full warm-check root was supplied, and no redundant
check was launched to manufacture one. Product inputs are frozen at
`sha256:92df12a6c2a10799118efd5b120fa574bb77782a697aa461dce22a7002601b68`.
`make target-plan TARGET=ci DETAIL=rows JSON=1` selected 2,510 rows, including all
14 Links rows, Notes creation, both entity merge/recovery browser rows, and fixture
lifecycle coverage. This is selection evidence only. `CARTULARY_TEST_RUN_ID=links-s09-ci-20261109 make ci`
was launched and subsequently cancelled at the maintainer's request, as recorded
below. Final acceptance uses the completed targeted portfolio.

First CI qualification reported a product/test_assertion_failure in
`module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8`
(`unit-results/go-9d3cb97910f61dfa-3db46bc3a259.json`). At that checkpoint the
invocation was draining; outer exit and cleanup were still pending.
The retained stdout tail does not contain the original assertion, and the row
has no typed failure diagnostic. Reproduce only this exact row after drain to
recover the cause before repair or broader requalification. No source repair or
flakiness conclusion is justified by the available evidence. S-09 stays open.

The maintainer then explicitly directed: "End the CI run addressed the issue and
use targeted tests." This supersedes S-09's full-CI completion requirement.
Cancelled that exact invocation; its published summary reports cancelled,
68 passed / 1 failed / 395 cancelled work units, with all 153 cleanup actions
completed and no cleanup failures. The original product assertion failure remains
recorded separately. Native outer completion is still pending. A fresh isolated
reproduction is running at `links-s09-bundle-repro-20261109` using
`make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8`.
Do not rerun full CI or claim its cancellation as passing evidence. Final acceptance
uses the targeted remediation portfolio and explicitly records the revised scope.

CI native outer exit is now observed: 130. The isolated bundle reproduction exited
2 with product/test_assertion_failure; its retained stderr identifies
`embedded_destination_verification`: "canonical destination expectation differs".
The authored owner expectation pins manifest `bdfa6425…`, whereas the independent
portable input pins and contains `5ecf0c00…`. Both files are unchanged from baseline
HEAD, establishing a pre-existing fixture mismatch unrelated to Links behavior.
Correct the authored expectation (including its fixture version), improve the
assertion's diagnostics, regenerate its embedded projection through Make, and
requalify that row. Preserve the exact manifest/payload/health assertions.

The fixture repair passed: `links-s09-bundle-fixed-20261109` closes the exact
bundle row (including historical artifacts, destination verification, cohort
publication/commit rollback and supersedes portability); `links-s09-reference-fixture-20261109`
closes `module.reference_data.unit.canonical_fixture_manifests`. Both exited 0.
Generation passed at `20261109T171006Z-p51194`. The authored fixture advances to
version 4; its generated embedded projection was regenerated, and the assertion
now reports actual/expected identities on failure. No runtime import rule changed.
Re-finalize and freeze this source before the final targeted portfolio.

Final-source `make agent-finalize` passed at `20261109T171134Z-p57230` (outer
exit 0). Its `unit-artifacts/finalize-summary.json` reports unchanged generated
outputs, passing schema/catalog/tier checks and generated refresh validation.
Frozen source:
`sha256:8294b04ac1522691c12356b925eb2c9d0a64a615b38d421326b10f18a21de6be`.
Retained warm-check maintenance remains unselected for the same reason recorded
above. The final targeted runs below must match this digest; the cancelled CI and
earlier-source runs are historical evidence, not final acceptance substitutes.

### Compatibility and operational handoff

- No production schema migration, HTTP/WebSocket shape change, bundle-version
  transition, or canonical history translation is required. Deploy the internal
  rollback contract and its four producer updates together; no compatibility alias
  is retained.
- New merges accept valid retained import/system/rollback provenance and preserve
  independent attribution. Ordinary command admission is unchanged. Existing
  incorrect historical attribution is not guessed or rewritten; the merge change
  set continues to identify the transforming actor and time.
- New snapshots include eligible assignment support references. This can change
  restrictive disclosure outcomes; support never grants recipient permission.
  Stored snapshots and releases retain their captured payloads, including on
  export/import and replay. No live-state backfill is performed.
- All three fixture profiles use the version-2 Links contribution. Their generated
  source-contract identities, keys and receipts invalidate version-1 cached state.
  Consumers build a fresh snapshot through the existing Make-managed lifecycle;
  old artifacts must not be translated or relabeled. The bounded lifecycle test
  has exercised a fresh build and isolated clones. Production data and aggregate
  mention-count expectations do not change.
- Reverts must keep each slice's code, authored contracts, generated projections
  and tests together. They may change future behavior or fixture identities, but
  must never rewrite retained product history, snapshots or releases.

Run shorthand in the finding map expands to `links-s09-<name>-20261109`;
all roots are under `.cartulary/test-results/` and have `run-summary.json`.

| Finding | Responsible owner / remediation | Slice | Required evidence / contract  Final run IDs (all passed) |
| --- | --- | --- | ---  --- |
| F-01 | Records envelope capability; Links eligibility and retained inverse semantics | S-01/S-02 | T-09/T-10; V-10/V-11/V-12; Revisions failure/merge/indicator rows; C-01–07  `final-links`, `final-records`, `final-revisions` |
| F-02 | Revisions owns live-row deltas; delete unused metadata from four source providers | S-06 | Changed-field completeness, mention/collection/indicator admission, transaction failures; C-04/C-07  `final-revisions`, `final-entities`, `final-indicators` |
| F-03 | Links fixture links/tags; application aggregate mention count; versioned descriptors | S-07 | V-14 and fixture runtime/key/lifecycle rows; negative aggregate count, receipt rejection  `final-fixtures`, `final-generated` |
| F-04 | Links canonical retained-value admission and deterministic merge identity | S-04 | T-13, Entities route/store, whole-change-set inverse; C-01/C-04/C-07/C-08  `final-links`, `final-entities`, `final-revisions` |
| F-05 | Retain cohesive Links persistence/codecs; no broad split or generic facade | S-08/S-09 no-action | T-01–10/T-13, boundary lint; owner-local SQL and canonical values remain  `final-links`, `final-boundary` |
| F-06 | Reporting includes assignment support independently from assignment interpretation | S-05 | T-11, adapter/materialization/frozen/replay/redaction rows; C-09  `final-links`, `final-reporting` |
| F-07 | Retain caller-owned authorization, transactions, parent history/publication; Links supplies effects and projection facts | S-08/S-09 no-action | T-10, Artifacts/Entities/Revisions integration; C-04/C-05/C-10/C-11  `final-artifacts`, `final-entities`, `final-revisions` |
| F-08 | Retain existing frontend/grid/saved-view owners; no frontend extraction | S-08/S-09 no-action | Notes creation and entity merge/recovery browser rows; C-12/C-13  `final-links`, `final-entities` |
| F-09 | Preserve authored harness/verification ownership; select exact rows and regenerate downstream projections | S-03/S-09 no-action on product boundaries | All 14 Links rows, catalog/tier/schema/generated validation; execution metadata governs evidence  `final-audit`, `final-contract`, `final-schema`, `final-generated` |
| F-10 | Replace stale framework assumptions with target-specific slices and current Make commands; retain domain owner navigation | S-08/S-09 no-action on module shape | No generic facade split; current Markdown lint and task-guide selection recorded  `final-audit`; final documentation gate below |
| F-11 | Exact SQL template grants replacing blanket Links Records permissions | S-03 | Boundary positive/negative matrix, lint, schema, generated drift; V-06/V-07  `final-boundary`, `final-contract`, `final-schema`, `final-generated` |
| F-12 | Preserve independent owner/creator and decision/creation attribution on recreation | S-04 | Imported provenance matrix, persisted/canonical metadata, exact inverse restoration  `final-links`, `final-entities`, `final-revisions` |

No-action findings keep justified owner boundaries rather than freezing incidental
behavior. Their unresolved risk was misleading planning scope; the active plan
now explicitly prohibits the unnecessary extraction/coupling described above.
The adopted corrections intentionally replace the historical F-02/F-03 deferrals
and F-04/F-06 compatibility assumptions. Historical sections below are evidence
of earlier planning, not current implementation status or executable inputs.

### Exact focused execution ledger

Commands below use the repository root and default harness settings. Each retained
run directory is `.cartulary/test-results/<run-id>/`; `run-manifest.json` records
HEAD, full source/graph/toolchain identities and normalized inputs;
`run-summary.json` and target/row projections record the outcome. These intermediate
passes describe their own source snapshots; the final targeted portfolio below
establishes final-source acceptance under the maintainer's revised scope. Initial
expected regressions and repair failures remain in the
checkpoints above.

| Slice | Exact passing command | Run ID / outer exit | Source digest |
| --- | --- | --- | --- |
| S-00 | `make test-slice OWNER=module.links ROWS=module.links.integration.active_links_and_tags_views_v1_contract_eaf4aac3b7,module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906,module.links.integration.party_assignment_source_facts_preserve_snapshot_963c86bbbb,module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1` | `links-s00-baseline-20261109b` / 0 | `sha256:682bfc492ba2f29821b9c0afc97fe5bebf795087daeadb8db24cf44d992aed8e` |
| S-00 | `make test-slice OWNER=module.records ROWS=module.records.envelope.store` | `links-s00-records-20261109` / 0 | `sha256:682bfc492ba2f29821b9c0afc97fe5bebf795087daeadb8db24cf44d992aed8e` |
| S-00 | `make test-slice OWNER=module.revisions ROWS=module.revisions.integration.covers_delete_restore_single_entry_rollback_dire_b348439e6d,module.revisions.integration.indicator_observation_create_and_resolution_targ_6249c556e2,module.revisions.integration.merge_aware_change_set_rollback_restores_survivo_130bd55b07,module.revisions.rollback.changed_field_completeness,module.revisions.rollback.transaction_failure_injection_8c5e3f2a11` | `links-s00-revisions-20261109` / 0 | `sha256:682bfc492ba2f29821b9c0afc97fe5bebf795087daeadb8db24cf44d992aed8e` |
| S-00 | `make test-slice OWNER=module.links ROWS=module.links.unit.canonical_valuecodec_is_pure_and_exact_9181a7cf12,module.links.unit.closed_vocabulary_exact_round_trips_3b5f179c21,module.links.unit.immutable_mutation_value_isolated_96e35c81c9,module.links.unit.links_item_refs_are_canonical_ade3870b93,module.links.unit.links_revision_contribution_owns_current_target_08116fb575,module.links.unit.source_state_contributions_remain_exact_8d3f67c521,module.links.unit.source_state_manifest_is_immutable_and_exact_2e516fd3a4,module.links.unit.strict_portability_preparation_fails_closed_8617ab9352` | `links-s00-unit-20261109` / 0 | `sha256:c52f623bf33378a23e4d511e31283a21358ff6577c41eb35438afad0b135f6b4` |
| S-06 | `make test-slice OWNER=module.revisions ROWS=module.revisions.integration.covers_delete_restore_single_entry_rollback_dire_b348439e6d,module.revisions.integration.indicator_observation_create_and_resolution_targ_6249c556e2,module.revisions.integration.merge_aware_change_set_rollback_restores_survivo_130bd55b07,module.revisions.rollback.changed_field_completeness,module.revisions.rollback.transaction_failure_injection_8c5e3f2a11` | `links-s06-complete-20261109` / 0 | `sha256:8ccb62733dbeb79e2e8e8e91c9b632f4fe11161aba22601c96e783e0b00fc731` |
| S-06 | `make test-slice OWNER=module.entities ROWS=module.entities.unit.host_identity_rollback_contract_411e7e4300` | `links-s06-collections-20261109` / 0 | `sha256:8ccb62733dbeb79e2e8e8e91c9b632f4fe11161aba22601c96e783e0b00fc731` |
| S-06 | `make test-slice OWNER=module.entities ROWS=module.entities.unit.mention_rollback_contract_99130c8734` | `links-s06-mentions-fixed-20261109` / 0 | `sha256:7950f8eb7b1f704f23f1cd2015db268e9d6ed108643d0f64006a4571a16509fb` |
| S-06 | `make test-slice OWNER=module.indicators ROWS=module.indicators.vocabulary.rollback_admission` | `links-s06-indicators-20261109` / 0 | `sha256:5db1d7149be8fa251aaf38f6a5d5acc3bebfb57f81886f989eecc9b7915a8ccd` |
| S-01/S-02 | `make test-slice OWNER=module.links ROWS=module.links.integration.active_links_and_tags_views_v1_contract_eaf4aac3b7,module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906` | `links-s02-20261109` / 0 | `sha256:a037e7db255f139dcd67f675513d9dbffa210a2634140ef0ff335372fdb9f769` |
| S-02 | `make test-slice OWNER=module.revisions ROWS=module.revisions.integration.covers_delete_restore_single_entry_rollback_dire_b348439e6d,module.revisions.integration.indicator_observation_create_and_resolution_targ_6249c556e2,module.revisions.integration.merge_aware_change_set_rollback_restores_survivo_130bd55b07,module.revisions.rollback.changed_field_completeness,module.revisions.rollback.transaction_failure_injection_8c5e3f2a11` | `links-s02-revisions-20261109` / 0 | `sha256:a037e7db255f139dcd67f675513d9dbffa210a2634140ef0ff335372fdb9f769` |
| S-03 | `make harness-contract SUITE=contract` | `links-s03-contract-closed-20261109` / 0 | `sha256:e2101c477969e0e540f2408ad05f409cd0ea8447f37f7778f7a67836a856f98a` |
| S-03 | `make lint CHECK=backend-boundary` | `links-s03-boundary-fixed-20261109` / 0 | `sha256:fc7c3de2a2a70498a0d8809ce26d2c9fa83e223b531fe6d83c11e06d8d2463e8` |
| S-03 | `make source-check CHECK=schema` | `links-s03-schema-20261109` / 0 | `sha256:3c7d1854aff27008b743c77d80021551b85407d110a6cb1ca094d5fbcd6e71ef` |
| S-03 | `make source-check CHECK=generated` | `links-s03-generated-closed-20261109` / 0 | `sha256:e2101c477969e0e540f2408ad05f409cd0ea8447f37f7778f7a67836a856f98a` |
| S-03 | `make test-slice OWNER=module.links ROWS=module.links.integration.active_links_and_tags_views_v1_contract_eaf4aac3b7,module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906` | `links-s03-fields-20261109` / 0 | `sha256:fc7c3de2a2a70498a0d8809ce26d2c9fa83e223b531fe6d83c11e06d8d2463e8` |
| S-04 | `make test-slice OWNER=module.links ROWS=module.links.integration.active_links_and_tags_views_v1_contract_eaf4aac3b7,module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1` | `links-s04-20261109` / 0 | `sha256:210186268a7c0456ce66e9e80db6534ffdb1d0c343d1783712b606e665123d79` |
| S-04 | `make test-slice OWNER=module.entities ROWS=module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044,module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310` | `links-s04-entities-20261109` / 0 | `sha256:210186268a7c0456ce66e9e80db6534ffdb1d0c343d1783712b606e665123d79` |
| S-04 | `make test-slice OWNER=module.revisions ROWS=module.revisions.integration.merge_aware_change_set_rollback_restores_survivo_130bd55b07` | `links-s04-revisions-fixed-20261109` / 0 | `sha256:3266841971dcbc1642ec9860354184456770948fda14114646f009fb9098cdb6` |
| S-05 | `make test-slice OWNER=module.links ROWS=module.links.integration.party_assignment_source_facts_preserve_snapshot_963c86bbbb` | `links-s05-adapter-20261109` / 0 | `sha256:a9bb4a25db33127321cb9c43c6e8d42f190e999b03efbbd3bbbd39afb0bbc13a` |
| S-05 | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c,module.reporting.unit.frozen_snapshot_payload,module.reporting.unit.logical_support_target_provider_composition_62b6d3190a,module.reporting.unit.snapshot_record_materialization,module.reporting.unit.snapshot_reporting_recipient_redaction_profiles_75c820718c,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | `links-s05-reporting-20261109` / 0 | `sha256:a9bb4a25db33127321cb9c43c6e8d42f190e999b03efbbd3bbbd39afb0bbc13a` |
| S-07 | `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.performance_fixture_snapshot_lifecycle,harness.browser.unit.performance_fixture_lifecycle,harness.browser.unit.performance_fixture_runtime_and_assembler,harness.browser.unit.performance_fixture_snapshot_key,harness.browser.unit.source_owner_contribution_assembler` | `links-s07-fixtures-20261109` / 0 | `sha256:92df12a6c2a10799118efd5b120fa574bb77782a697aa461dce22a7002601b68` |
| S-07 | `make source-check CHECK=generated` | `links-s07-generated-20261109` / 0 | `sha256:92df12a6c2a10799118efd5b120fa574bb77782a697aa461dce22a7002601b68` |

### Final targeted validation and handoff ledger

All commands below passed with native outer exit 0 on the final frozen source
`sha256:8294b04ac1522691c12356b925eb2c9d0a64a615b38d421326b10f18a21de6be`.
The portfolio covers 53 distinct catalog rows; every selected row passed, with no
missing, skipped or cancelled rows and no cleanup failure in these accepted runs.
One fixture work unit reused admitted cached evidence; other accepted invocations
reported no cache hits. The fixture lifecycle integration executed a fresh bounded
snapshot build and clone checks; this is not performance publication.

`test-evidence-audit` admitted the complete Links owner (all 14 original T-rows),
including Notes creation, against current source, graph, selection and cleanup.
Its report is `links-s09-final-audit-20261109/test-evidence-audit/audit.json`.
Other owner selections are intentionally scoped to the remediation obligations;
no full-owner or full-CI claim is made for them. Both entity merge/recovery browser
rows passed in `links-s09-final-entities-20261109`. The bundle reproduction and
repair roots remain separate, and the cancelled CI is never an acceptance input.

Commands use normalized Make inputs from each retained `run-manifest.json`.
Prefix each command with `CARTULARY_TEST_RUN_ID=<run-id>` from its row. Each root
contains `run-summary.json`, `target-summaries/`, selected `rows/` and
`cleanup-results.json`; these are the exact retained outcome/evidence paths.

| Run ID | Passing command | Selected rows |
| --- | --- | --- |
| `links-s09-final-links-20261109` | `make test-slice OWNER=module.links` | 14 |
| `links-s09-final-entities-20261109` | `make test-slice OWNER=module.entities ROWS=module.entities.browser.entity_merge_exact_recovery,module.entities.browser.the_browser_inspector_merges_duplicate_entities_d0cb95b021,module.entities.integration.the_explicit_merge_route_repoints_live_fan_out_p_0ec76e8044,module.entities.store.explicit_merge_store_behavior_repoints_live_ment_0560122310,module.entities.unit.host_identity_rollback_contract_411e7e4300,module.entities.unit.mention_rollback_contract_99130c8734` | 6 |
| `links-s09-final-revisions-20261109` | `make test-slice OWNER=module.revisions ROWS=module.revisions.integration.covers_delete_restore_single_entry_rollback_dire_b348439e6d,module.revisions.integration.indicator_observation_create_and_resolution_targ_6249c556e2,module.revisions.integration.merge_aware_change_set_rollback_restores_survivo_130bd55b07,module.revisions.rollback.changed_field_completeness,module.revisions.rollback.transaction_failure_injection_8c5e3f2a11` | 5 |
| `links-s09-final-records-20261109` | `make test-slice OWNER=module.records ROWS=module.records.envelope.store` | 1 |
| `links-s09-final-indicators-20261109` | `make test-slice OWNER=module.indicators ROWS=module.indicators.vocabulary.rollback_admission` | 1 |
| `links-s09-final-reporting-20261109` | `make test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c,module.reporting.unit.frozen_snapshot_payload,module.reporting.unit.logical_support_target_provider_composition_62b6d3190a,module.reporting.unit.snapshot_record_materialization,module.reporting.unit.snapshot_reporting_recipient_redaction_profiles_75c820718c,module.reporting.unit.snapshot_reporting_redaction_applies_determinist_7f142ce89d` | 6 |
| `links-s09-final-artifacts-20261109` | `make test-slice OWNER=module.artifacts ROWS=module.artifacts.linked_notes.note_association_resource,module.artifacts.source_mutations.note_association_admission` | 2 |
| `links-s09-final-fixtures-20261109` | `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.performance_fixture_snapshot_lifecycle,harness.browser.unit.performance_fixture_lifecycle,harness.browser.unit.performance_fixture_runtime_and_assembler,harness.browser.unit.performance_fixture_snapshot_key,harness.browser.unit.source_owner_contribution_assembler` | 5 |
| `links-s09-final-boundary-20261109` | `make lint CHECK=backend-boundary` | 1 |
| `links-s09-final-schema-20261109` | `make source-check CHECK=schema` | 0 |
| `links-s09-final-generated-20261109` | `make source-check CHECK=generated` | 2 |
| `links-s09-final-contract-20261109` | `make harness-contract SUITE=contract` | 8 |
| `links-s09-final-audit-20261109` | `make test-evidence-audit OWNER=module.links EVIDENCE_ROOTS_FILE=.cartulary/audit-inputs/links-s09-final.json` | 0 |
| `links-s09-bundle-fixed-20261109` | `make test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | 1 |
| `links-s09-reference-fixture-20261109` | `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests` | 1 |

RB-001–RB-004 are CLOSED: baseline/endpoint characterization and Records wiring,
retained-provenance merge and attribution, Reporting assignment inclusion with
restrictive disclosure, and exact SQL guard qualification all have passing final
evidence above. RB-005 is SUPERSEDED by this implementation authorization and the
canonical documentation lint, not silently deferred. Historical TODO/DEFERRED
labels below apply only to the preserved planning session.

The original untracked tracker content was reconstructed and its SHA-256 matched
`aaa0882665e6db8a4d623dbabae08547e8b7360feaae98f4086ad1cd4cf136af` exactly.
HEAD remains `a43248fbe3f762b0f448993126f74172f597f101`; the index remains unchanged.
There were no unrelated tracked user edits at baseline, and no commit, staging,
history rewrite or remote publication was performed. `git diff --check` passed.

S-09 is complete: planning, implementation, product/tooling validation and handoff
review are complete for the maintainer's revised targeted scope. No unresolved
in-scope implementation blocker or required targeted check remains. Documentation
lint passed with outer exit 0 at
`links-s09-doc-check-20261109/adhoc/lint-markdown/tool-run-summary.json`.
The final post-update handoff command is
`CARTULARY_TEST_RUN_ID=links-s09-handoff-20261109 make lint-markdown`; its retained
result is `.cartulary/test-results/links-s09-handoff-20261109/adhoc/lint-markdown/tool-run-summary.json`.
Full CI completion was explicitly withdrawn by the maintainer; visual golden
changes, performance publication and release acceptance were never selected.
Retained warm-check maintenance was not run because no compatible successful full
warm-check root was available. These scope limits are not deferred remediation.


## 1. Scope and Source Posture

Target path: internal/modules/links. Target label: links, derived from the final path component and normalized to lowercase kebab case; it contains no separators, spaces, or unsafe filename characters. Output: docs/handoffs/links-module-refactor-tracker.md, under repository root /home/jochi/code/cartulary.

This is a planning-only tracker. The only authorized persistent write is this file. Implementation requires a later authorized task. No production code, tests, public APIs, schemas, contracts, generated output, migrations, dependency files, or harness files may be changed by this task. The slice proposals below are reviewable planning candidates, not approved implementation decisions. Non-goals include redesigning HTTP or WebSocket behavior, moving responsibilities merely to match directory names, introducing another service, and changing fixture or evidence claims.

Session LNK-2026-11-09-02 continued the preceding read-only planning turn. The preceding turn reported main at ce30337c032a68d145c7ddc061979f655c908951. This session inspected main at a43248fbe3f762b0f448993126f74172f597f101 with an initially clean worktree/index. The target still contains exactly 45 regular tracked Go files and no symlinks. The tracker was absent at admission. All 45 target files were opened for behavioral inspection; adjacent files were inspected to the depths stated below. Target files and governing excerpts were unchanged across those two commits; the inspected task-surface Make projection changed, so command discovery was refreshed.

Authority is: (1) adopted subsystem NLSpecs, only within their named scope; (2) Core 00–04 for implementation conformance, using Core 00 primary-owner allocation; (3) Core 05 only for timed or fixture-sensitive publication claims; (4) domain vocabulary and implementation-support guidance; (5) live code and tests as implementation evidence; (6) framework and historical handoffs as planning evidence. Tests, typed contracts, routing and allowlists cannot override an owner. No publication claim is made and Core 05 is not a runtime design input. No primary-owner contradiction was established for the proposed envelope-read slices. If one is established later, record BLOCKED: owner contradiction and stop that dependent slice without choosing a side.

The supplied framework was read first. Its instructions are used as the requested planning doctrine; its examples do not authorize implementation or establish current state. Repository documentation was read by this planning session only; no product test, generator, runtime metadata or conformance mechanism was made to consume it.

Evidence register; paths are repository-relative unless absolute. “Excerpt” explicitly limits the inspection claim.

| Evidence ID | Documents/files inspected | Inspection depth and use |
| --- | --- | --- |
| E-01 | docs/handoffs/cartulary_modular_refactor_planning_framework.md; AGENTS.md; internal/AGENTS.md; tools/AGENTS.md; /mnt/c/Users/jochi/.codex/skills/refactor-tracker/SKILL.md and references/tracker-format.md; .agents/skills/cartulary-test-ops/SKILL.md and references/command-recipes.md | Framework and applicable instruction chains read. Nested AGENTS search found no additional instructions under the target or docs. Tracker-only authorization takes precedence over generic implementation workflow advice. |
| E-02 | docs/spec/00_document_set_status_and_precedence.md; docs/spec/01_architecture_storage_and_view_contracts.md | Owner/status and allocation excerpts; Core 01 §§2, 3.3.4.2, 3.3.5, 7.4.5A, 8.6, 12.3.7 and current portability amendment. Particularly REQ-01-649 (Records), REQ-01-658/662 (projection/fact boundary), REQ-01-675 (Notes associations). Historical adoption references are not substituted for current clauses. |
| E-03 | docs/spec/02_domain_model_schema_and_history.md; docs/spec/03_workbook_interaction_collaboration_and_workflows.md; docs/spec/04_security_deployment_and_conformance.md; docs/domain.md | Relevant excerpts: Core 02 merge §9, tags/links and retained history, assignment REQ-02-143a.2; Core 03 saved/local view state and mutation/save/conflict interaction; Core 04 current incident authorization; domain Links vocabulary. |
| E-04 | docs/reporting-subsystem-nlspec.md; docs/report-composition-nlspec.md; docs/testing-harness-nlspec.md | Adopted/current status and scope/adoption declarations inspected. Reporting 1.5.0 §§1 and 7.1.2 (REQ-RPT-026b–d) owns consumer adaptation, not Links mutations; Report Composition 1.3.1 §1 establishes the companion boundary. Testing Harness owns execution/accounting mechanics. No Links-specific adopted NLSpec was identified in Core 00's owner allocation. |
| E-05 | Every path in §2 | Complete target-file content inspection, including private packages, all tests and fixture support. Import/symbol searches supplement, rather than replace, source reads. |
| E-06 | internal/app/timelineassembly/relationship_adapters.go and assembly.go; internal/app/timelinefactassembly/links.go; internal/app/tasksdecisionassembly/link_facts.go; internal/app/entitymergeassembly/links.go; internal/app/reportingassembly/links.go | Adapter bodies inspected; assembly construction excerpts. These are concrete consumers of Links facts/commands and consumer-owned ports. |
| E-07 | internal/app/workbookassembly/note_associations.go; internal/modules/artifacts/mutation_collections.go; internal/modules/tasksdecisions/mutation_collections.go; internal/modules/entities/hostidentity/party_assignments.go; internal/app/importassembly/owner_registry.go | Full narrow adapters and relevant collection/composition excerpts. Source owners admit actions, coordinate parent history and map errors; import composes existing owners. |
| E-08 | internal/app/revisionassembly/revisions.go; internal/modules/revisions/rollbackcontract/contract.go; rollback_apply_nonrow.go; rollback_apply_effects.go; rollback_apply_row.go; internal/modules/records/store.go | Revisions contribution/composition, non-row application and effects inspected; row rollback excerpts. Records envelope lookup/batch lookup and locking methods inspected. Paths without a repeated prefix in this cell are under internal/modules/revisions. |
| E-09 | internal/app/projectionassembly/catalog.go; internal/modules/projections/adapters/adapters.go; internal/modules/projections/internal/runtime/link_effects.go; internal/app/incidentportabilityassembly/catalog.go; internal/app/recoveryassembly/state_catalog.go | Projection composition excerpts and linked-record refresh implementation; complete relevant source-port/recovery catalog construction. |
| E-10 | internal/modules/workbook/routes.go and routes_mutation.go; internal/modules/collaboration/routes.go and socket_transport.go; contracts/openapi-source/owners/module.workbook/openapi.json and module.revisions/openapi.json; contracts/view-schemas/cartulary.view.timeline.v2.json and cartulary.view.notes.v1.json | Actual route registration, mutation authorization excerpts, socket transport admission excerpts, authored path/operation and field-schema objects. Generated clients were not edited or treated as authority. |
| E-11 | internal/modules/artifacts/note_associations_integration_test.go; internal/modules/revisions/integration_test.go and rollback_changed_fields_test.go; apps/web/e2e/notes.authoring.spec.ts | Assertion bodies/excerpts: Notes replay/roles/closed/revoked access; non-row rollback endpoint versions/history/events; changed-cell completeness; full Notes browser scenario. Inspected assertions are not passing-run evidence. |
| E-12 | internal/testutil/appsupport/performancefixture/assembly.go, production.go and assembly_test.go; tools/test_support_inventory.json | Relevant fixture assembly/validation and assertion bodies; exact Links owner-local, non-service-starting support inventory entry. Fixture accounting is not production architecture. |
| E-13 | Makefile; tools/task_surface_owner.json; tools/task_surface.generated.mk; tools/task_surface.help-all.generated.json; tools/execution_topology_manifest.json; tools/generated_artifact_policy.json; tools/backend_module_boundaries.json; tools/toolchain_pins.json; tools/harness/static-analysis/markdown-biome.sh and markdown-biome-runner.sh | Make entry/recipe and relevant input/side-effect definitions inspected, not executed. Current Markdown lint uses Biome and cache stamps. Exact generated-output registry and Records SQL allowlist inspected. |
| E-14 | tools/test_families/module.links.json; contracts/verification/owners/module.links.json; relevant rows in tools/test_families/module.records.json, module.revisions.json and harness.browser.json | Current owner, 14 Links rows, verification IDs, selectors, collaborators and service prerequisites inspected. Cross-owner rows are selection candidates; source inspection is separately identified in E-11/E-12. |

Repository files found only through wider caller searches are not claimed fully reviewed. In particular, all peer admission branches, saved-view handlers, merge transaction orchestration, and Reporting materialization tests were not exhaustively audited. Their dependent characterization/discovery gates remain explicit below.

## 2. Current-State Repository Inventory

All paths in this table are relative to internal/modules/links. Coverage is 45/45 files; none is excluded. Public means exported Go surface, not a new HTTP API. Private-package exports remain restricted by Go internal visibility. Dependency cells list material repository/external boundaries, omitting routine standard-library imports; pgx and UUID use are retained where material. T-IDs resolve to exact harness rows in §8. C-IDs resolve to §4. “Tests” denotes inspected assertions or mapped coverage, never execution success.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| active_facts.go | Read ordered active link/tag facts in caller transaction | RecordLinkFact, RecordTagFact, ActiveFacts, FactReader.LoadIncidentTx/LoadRecordTx, FactReadError | E-06 Reporting, Timeline and Tasks/Decisions adapters | pgx, UUID; active_record_links_v1, active_record_tags_v1 | portability_test.go; T-10, T-13; consumer assertions | C-05/C-09 source-fact shape; no generated file | Links | high | Full read. Record scope includes outgoing links and incoming supersedes; explicit empty arrays, safe typed error/no partial facts. |
| association_lookup.go | Null-field directional association lookup, keyset paging and ensure | AssociationLookup, AssociationLink; Store.ListAssociationsTx, LoadAssociationTx, EnsureAssociationTx | Artifacts NoteAssociations via E-07 app composition | pgx, own record_links and upsert command | E-11 Notes tests; T-09 indirectly | C-06 association identities/paging | Links storage; Artifacts semantic policy | high | Full read. Internal limit 1–501 supports lookahead; public limit is owned above. Existing association retains provenance. |
| collection_actions.go | Typed collection input and target validation | RecordRef/PartyRef/TagCollectionValidation and Command; TagCollectionAdd, RecordTagRef, CollectionValidationError; Validate*CollectionTx | Artifacts/TasksDecisions collections, Entities party assignments, local mutation methods | pgx; standalone records reads | T-09, T-11, T-13 | C-02/C-03 validation failures | Links; Records for envelope access | high | Full read. F-01. Ordinary field non-emptiness is not full view-schema admission; callers own it. |
| collection_mutations.go | Field-ref synchronization and ref/party/tag changes producing immutable mutations | CollectionMutationResult.Mutations; SyncFieldReferenceWithMutationValuesTx; ApplyRecordRef/PartyRef/TagCollectionWithMutationValuesTx | E-06/E-07 owner adapters and mutation facades | mutationvalue, valuecodec, own SQL, ProjectionRows | T-10, T-11, T-13 | C-01–05/C-07 | Links | high | Full read. Add/remove ordering and no-op semantics; refs invoke projection hook, tags rely on coordinating owner. |
| commands.go | Typed command/result vocabulary | LinkType, LinkProvenance, Mutation aliases; parsers; UpsertLinkCommand, SyncFieldReferenceCommand, InsertSupersedesCommand, TombstoneActiveLinkCommand, RecordLinkCommandResult | Root methods, E-06 adapters, peer capabilities | private vocabulary and mutationvalue | T-01–03, T-09, T-13 | C-01/C-04 exact mutation payload | Links | high | Full read. Result accessor returns isolated mutation values. |
| field_refs.go | Field-aware link persistence | Private lookup/list/upsert/tombstone helpers | collection_mutations.go | pgx; own record_links; source/envelope join on tombstone | T-13 | C-01 nullable field tuple | Links | high | Full read. Necessary own-source join differs from F-01 standalone reads; do not remove every SQL allowance indiscriminately. |
| history_projection.go | Link/tag history facet projection | Private projectLinkHistory/projectTagHistory | revision_provider_contribution.go callbacks | revisions/historycontract | T-05, T-10; E-11 | C-04 history collection labels/IDs | Links semantic projection; Revisions presentation envelope | high | Full read. attached_evidence has distinct history semantics. |
| incident_bundle_source_port.go | Root portability contribution constructor | NewIncidentBundleSourcePort | E-09 incidentportabilityassembly; test support | incidentbundles/sourceport; private sourcestate | T-06, T-12, T-13 | C-08 sourceport descriptor | Links contribution; IncidentBundles orchestration | medium | Full read. Thin constructor; name alone gives no reason to move it. |
| internal/mergeeffects/merge_effects.go | Deterministic repoint/deduplicate link/tag effects | Internal RepointLinksTx/RepointTagsTx, commands/results, LinkDependencies | root merge_effects.go only | pgx; valuecodec, mutationvalue, vocabulary; own tables | T-13 merge assertions | C-01/C-04/C-07 | Links | high | Full read. Retains full field-aware tuple; validator callback creates F-04 provenance question. |
| internal/mutationvalue/mutation.go | Defensive typed mutation grammar/value isolation | Internal Value, New, Copy and getters | root command/collection/merge facades, private mergeeffects | Standard library only | T-03 | C-04 before/after payloads | Links | medium | Full read. Record-link UUID and composite record-tag identities; recursive copies. |
| internal/mutationvalue/mutation_test.go | Value grammar and copy tests | TestMutationConstructorRejectsInvalidGrammar; TestMutationRecursivelyIsolatesConstructionAndReaders | T-03 Go runner | mutationvalue; testing | Self | T-03 accounting | Links tests | low | Full assertion read; no runtime verification performed. |
| internal/revisionprovider/provider.go | Describe and apply non-row inverse mutations; affected-record/field facts | Internal Provider, NewProvider, DescribeTx, ApplyInverseTx | root RevisionProviderContribution; Revisions catalog | rollbackcontract, valuecodec, pgx; records and artifact-type lookup | T-05/T-10; E-11 rollback integration | C-04/C-07; F-01/F-02 | Links target semantics; Records envelope reads | high | Full read. Distinguish active-write checks from retained/deleted endpoint existence during rollback. |
| internal/revisionprovider/repository.go | Inverse-effect link/tag persistence and value loading | Private helpers | provider.go | valuecodec, vocabulary, pgx; own tables | T-10; E-11 | C-01/C-04 retained state | Links | high | Full read. Own-table SQL is not automatically misplaced. |
| internal/sourcestate/manifest.go | Authored exact Links source-state descriptor | Private manifest loader/validator/defensive projections | sourcestate port and prepare | incidentbundles/sourceport | T-07, T-06 | C-08 version 8, 3 paths, 7 invariant IDs | Links | high | Full read. Two source tables; tags.ndjson is validation rows, not a third source table. |
| internal/sourcestate/manifest_test.go | Exact descriptor, defensive copy and malformed authoring matrix | TestManifestProjectsExactSourceStateSurfaceAndDefensiveCopies; TestManifestRejectsMalformedAuthoringMatrix | T-07 | Private manifest | Self | T-07 accounting | Links tests | medium | Full assertion read. |
| internal/sourcestate/port.go | Streaming export, prepare/apply validation, recovery table contribution | Internal NewSourcePort, RecoveryStateContribution | Root portability/recovery constructors | incidentportability, sourceport, recoverystate; pgx and own scratch/source SQL | T-06/T-12/T-13 | C-08 retained source bytes and scratch cleanup | Links contribution | high | Full read. Borrows transaction; three scratch tables excluded/invalidate on recovery. |
| internal/sourcestate/prepare.go | Strict NDJSON members, canonical typed source preparation and invariant ordering | Private prepared values, sourceStateError, helpers | Source-port callbacks | sourceport, incidentportability, valuecodec, platform/fieldnorm | T-08/T-13 | C-08 exact field/time/tag normalization | Links | high | Full read. No SQL during preparation; owner-local validation is not generic import orchestration. |
| internal/sourcestate/prepare_test.go | Strict parsing, failure precedence and opaque copies | TestPrepareImportEnforcesStrictNDJSONAndExactMembers; TestPreparedValueIsOpaqueCopiedAndWrongTypesDoNotPanic; TestValidationPriorityIsFixedAcrossMultipleDefects | T-08 | portabilitysupport, sourceport, incidentportability | Self | T-08 accounting | Links tests | medium | Full assertion read; historical test labels are not accepted bundle versions. |
| internal/valuecodec/valuecodec.go | Closed canonical source/history values and transition validation | Internal RecordLink/TagMutationValue, Identity/Input; build/decode/parse/history validation and Map methods | Root commands, history validator, merge, revision and portability code | vocabulary, UUID, standard library; no SQL/pgx | T-01, T-08, T-13 | C-01/C-04/C-08 canonical maps | Links | high | Full read. Retained/import provenance grammar is wider than ordinary command admission; preserve distinction. |
| internal/valuecodec/valuecodec_test.go | Exact members, invalid shapes and mutation operation matrix | Six tests selected by T-01 | T-01 | Private valuecodec | Self | T-01 accounting | Links tests | medium | Full assertion read. Reject legacy/narrative shapes and preserve canonical fields. |
| internal/vocabulary/vocabulary.go | Exact closed typed token vocabulary | Internal LinkType/LinkProvenance, Parse*, String, MarshalText | Public aliases, private codec/repository/merge | Standard library | T-02; T-09 | C-01 tokens | Links | medium | Full read. Ten types and five provenances; no normalization or unknown-token fallback. |
| internal/vocabulary/vocabulary_test.go | Token roundtrip and rejection | TestLinkVocabularyExactRoundTrips; TestLinkVocabularyRejectsNonExactTokens | T-02 | Private vocabulary | Self | T-02 accounting | Links tests | low | Full assertion read. |
| item_refs.go | Canonical collection item references | RecordRefItemRef/ParseRecordRefItemRef; PartyRefItemRef/ParsePartyRefItemRef; RecordTagItemRef/ParseRecordTagItemRef | Artifacts, TasksDecisions, Entities action/value adapters; root tag validation | UUID, string grammar | T-04; T-10/T-11/T-13 | C-02/C-03 item_ref wire values | Links | medium | Full read. Exact canonical UUIDs; record_tag includes record and assignment identities. |
| item_refs_test.go | Item-ref positive/negative grammar | TestLinksItemRefsAreCanonical; TestLinksItemRefParsingIsStrict | T-04 | Root item-ref helpers | Self | T-04 accounting | Links tests | low | Full assertion read. |
| merge_effects.go | Root transaction-scoped merge capability | RepointMergedLinks/TagsCommand and Result; Store.RepointMergedLinksTx/RepointMergedTagsTx | internal/app/entitymergeassembly/links.go | Private mergeeffects/mutationvalue; root validation callback | T-13; peer entity merge rows discovered | C-07 merge mutation/count outputs | Links | high | Full read. Copies returned mutation maps and link-type lists; does not own merge admission. |
| merge_test.go | Field-aware deduplication characterization | TestMergeDeduplicatesOnlyFullFieldAwareTupleCharacterization | T-13 | appsupport and owner test fixtures | Self | C-01/C-07; T-13 | Links tests | high | Full assertion read. Manual provenance coverage does not resolve non-manual case F-04. |
| party_assignment_facts.go | Six exact Host/Identity assignment field recognizers and fail-closed fact reads | IsPartyAssignmentField; FactReader.LoadPartyAssignmentsTx/LoadPartyAssignmentsForRecords | Entities hostidentity; E-06 Reporting adapter | partyassignment; pgx; active links/endpoint type facts | T-11 | C-03/C-09 assignment source facts | Links, explicitly Core-allocated | high | Full read. Bounded reads; no inference from arbitrary references or labels. |
| party_assignment_facts_test.go | Assignment identity/digest and rejection matrix | TestPartyAssignmentSnapshotSource | T-11 | Entities/Parties fixtures, postgres, appsupport, conflicttest | Self | C-03/C-09; T-11 | Links tests | high | Full assertion read. Snapshot consistency, fixed digest, invalid provenance/type/duplicates, deleted endpoint exclusion, scalar bypass rejection. |
| partyassignment/fact.go | Canonical assignment state and opaque source token | State, Fact, NewFact, Fact.Validate | FactReader; Reporting exportprovider/source_facts.go; assignment tests | platform/canonicaljson, UUID, SHA-256 | T-11 | C-03 record_party_assignment.v1 and rpa_ token | Links | high | Full read. Exact UTC microseconds and explicit nulls; hash is not a revision counter. |
| portability_test.go | Fail-closed malformed import and active-fact error behavior | TestLinksIncidentBundleRejectsUnknownLinkMembersBeforeMutation | T-13 | Sourceport, incidentportability, portabilitysupport, appsupport | Self | C-08/C-09; T-13 | Links tests | high | Full assertion read. Unknown members cause no source/attribution writes; closed transaction yields typed error/no partial facts. |
| projection_dependencies.go | Discover relationship endpoints needing rebuild | Store.ProjectionDependentsTx | internal/app/revisionassembly/revisions.go | pgx; own nondeleted record_links | E-11 rollback; T-10 indirect | C-05/C-07 dependency scope | Links fact contribution | high | Full read. Deliberately does not exclude deleted endpoints needed for restore; do not substitute active-fact reader. |
| projection_mutations.go | Derive affected endpoints from before/after and call injected refresh | Private publishProjectionEffectsTx | Root command/ref mutations | valuecodec, ProjectionRows; caller pgx.Tx | T-10/T-13 | C-05 atomic refresh | Links effects; Projections execution | high | Full read. Distinct sorted endpoints; no hook/no mutation is a no-op. Not a generic row renderer. |
| record_link_commands.go | Upsert/replay/update/tombstone/supersedes source mutations | Store.UpsertLinkCommandTx, InsertSupersedesCommandTx, TombstoneActiveLinkCommandTx | E-06 Timeline; source-owner adapters and note associations | mutationvalue/valuecodec; pgx; own SQL and validation/projection helpers | T-09/T-10/T-13 | C-01/C-04/C-05 | Links | high | Full read. Same tuple replays; metadata changes yield patch; missing tombstone is no-op; exact mutation snapshots. |
| recovery_state.go | Root recovery contribution constructor | RecoveryStateContribution | internal/app/recoveryassembly/state_catalog.go | sourcestate, platform/recoverystate | T-06 | C-08 five-table recovery contribution | Links source catalog; Recovery lifecycle | medium | Full read. Keep semantic ownership separate from backup orchestration. |
| revision_provider_contribution.go | Register Links non-row providers and history facet validators | RevisionProviderContribution | internal/app/revisionassembly/revisions.go | revisions, private revisionprovider/valuecodec | T-05/T-10; E-11 | C-04 two target kinds | Links contribution; Revisions orchestration | high | Full read. Exactly record_link and record_tag; no extra peer-owned targets. |
| revision_provider_contribution_test.go | Exact two-target contribution contract | TestRevisionProviderContributionOwnsLinkAndTagHistoryFacet_Unit | T-05 | Root Links and Revisions contribution types | Self | T-05 accounting | Links tests | medium | Full assertion read. Provider and history validator presence checked. |
| route_projection_history_test.go | Service/route characterization of typed links/tags, projections/history/query/events/rollback | TestTypedLinksAndTags_Unit; TestLinkTagProjectionHistoryQuery_Integration; private route helpers | T-13 and T-10 respectively | Timeline/app assemblies, auth/records fixtures, HTTP/socket/history helpers | Self | C-02/C-04/C-05/C-10/C-11 | Links integration tests | high | Full assertion read. “_Unit” is a name, not execution class; both require integration fixtures per map. |
| source_state_contributions_test.go | Exact bundle/recovery contribution assertions | TestSourceStateContributionsRemainExact_Unit | T-06 | Root constructors | Self | C-08/T-06 | Links tests | medium | Full assertion read. Version 8, roles/identities, seven invariants, 2 authoritative + 3 transient tables. |
| source_state_portability_integration_test.go | Strict roundtrip, atomic invalid input and concurrent scratch isolation | TestLinksSourceStateStrictRoundTripAndAtomicFailure_Integration; TestLinksBundleValidationCommitIsolation_Integration | T-13/T-12 | IncidentBundles, portability helpers, Links fixtures, appsupport | Self | C-08/T-12/T-13 | Links tests | high | Full assertion read. Export/import/re-export bytes; failure precedence; no invalid writes; dirty scratch commit rejected, other transaction isolated. |
| store.go | Root storage/mutation construction, validation and supersedes lock | Store, ProjectionRows, NewStore, NewMutationStore, constants; HasActiveIncomingSupersedesLinkForUpdateTx | E-06/E-07/E-08 assemblies and peer capabilities | pgx, vocabulary; own links and standalone records queries | T-09/T-10/T-13 | C-01/C-05/C-07 | Links with Records lookup seam | high | Full read. Existing constructors deliberately distinguish orchestrated projection scope; F-01 is narrower than a facade rewrite. |
| store_test.go | Active views, endpoint admission and identity/metadata characterization | TestActiveLinksAndTagsViewsV1Contract; TestRecordLinkOwnerValidation; TestFieldAwareLinkIdentityAndMergeCharacterization | T-09/T-13 | Auth, Timeline/Entities fixtures, appsupport | Self | C-01/C-02/C-04; T-09/T-13 | Links tests | high | Full assertion read. Exact columns, deleted endpoint filtering, confidence, tuple/null-field distinctions, isolated maps. |
| test_helpers_test.go | SQL/history/canonical assertions for package tests | Private helpers; no production surface | Root Links test files | platform/postgres, testing and SQL helpers | T-09/T-10/T-11/T-12/T-13 support | Test-only expectations | Links tests | low | Full read. Not an application service or runtime package. |
| testsupport/links.go | Deterministic link/tag fixtures and assertions | LinkFixture, fixed IDs, SeedRecordLink/SeedRecordTag, LookupActiveLink, RequireActiveLink | Links portability and peer Entities/Revisions/SavedViews tests; import search | platform/postgres and database/sql; testing | T-13 plus discovered peer tests | C-12 fixture identity | Links owner-local testsupport | medium | Full read. Machine inventory marks non-service-starting support; no production import found in searched callers. |
| testsupport/performancefixture/production.go | Read-only fixture association-count validator | ProductionApplication, NewProductionApplication, ValidateFixtureAssociations | E-12 app fixture composition | platform/postgres; record_links, record_tags, entity_mentions | E-12 assembler tests | C-12 source-owner receipt expectations | Links fixture facts; Test-util aggregate candidate | medium | Full read. F-03 cross-owner mentions count; not evidence of production test-util leakage. |
| testsupport/performancefixture/provider.go | Descriptor-based fixture contribution/receipt | Expectations, Application, Provider, New, Descriptor, Apply | E-12 closed fixture assembler | internal/testutil/performancefixture | E-12 assembler tests | C-12 generated fixture descriptor input | Links fixture contribution | medium | Full read. Validates prebuilt fixture; copies descriptor; does not itself create source rows. |

## 3. Module Boundary Diagnosis

Links is a legitimate source-state owner and a transaction-scoped mutation facade with persistence adapters. It is not merely a thin forwarding service: identity, validation, canonical values, history semantics, active facts, merge effects and source portability are substantive responsibilities supported by Core and the adopted Reporting boundary. The directory is not evidence for its own legitimacy; those source/consumer contracts are.

It contains a specific cross-owner envelope-read defect (F-01) and some coupling requiring characterization. It is not presently established as an accidental catch-all, frontend shell, grid-vendor integration, generic view orchestration service, or transport adapter. Projection effect coordination is present, while physical projection refresh and parent mutation orchestration remain elsewhere.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Typed relationship/tag identity, retained source state and transaction effects | Root Store/commands/collections and private codecs | Links | keep | E-02/E-03/E-05; C-01–04 | No rename or wholesale module dissolution warranted. |
| Standalone current-envelope lookup and type/existence validation | store.go, collection_actions.go, internal/revisionprovider/provider.go | Records for lookup; Links for interpretation | split | Core 01 REQ-01-649; E-08; F-01 | Candidate seam only; leave joined source/envelope reads that meet the declared-read rule. |
| Source-owned rollback/merge contribution | revision_provider_contribution.go, internal/revisionprovider; merge_effects.go/internal/mergeeffects | Links source effects; Revisions/Entities orchestration | keep | Core 01 history owner split; Core 02 §9; E-06/E-08 | Neither contribution's name justifies transfer to a generic coordinator. |
| Per-record-type changed-field mapping during non-row inverse application | internal/revisionprovider/provider.go | Links semantic contribution with Revisions public result derivation | defer | E-08; F-02 | Determine independent consumers and failure semantics before deleting or moving mappings. |
| Relationship effect set and projection refresh port | projection_dependencies.go, projection_mutations.go, Store.ProjectionRows | Links dependency facts; Projections physical rebuild | keep | REQ-01-658/662; E-09 | Preserve same transaction and refresh failures; no duplicate view-schema engine found in target. |
| Source-bundle preparation/apply and recovery catalog | internal/sourcestate and root contribution constructors | Links contribution; IncidentBundles/Recovery lifecycle | keep | E-05/E-09; C-08 | No generic tabular-import lifecycle in target. |
| Assignment facts and opaque version token | party_assignment_facts.go, partyassignment/fact.go | Links, explicitly REQ-02-143a.2 | keep | E-03/E-04/E-07 | Entities coordinates Host/Identity writes; Reporting consumes source facts. |
| Notes association storage versus source/evidence/related-note policy | association_lookup.go versus Artifacts and workbookassembly | Links storage; Artifacts policy; Workbook transport | keep | REQ-01-675; E-07/E-10/E-11 | Do not put admission or Notes labels into Links. |
| Reporting field/support transformation | internal/app/reportingassembly/links.go, outside target | Application composition/Reporting | keep | REQ-RPT-026b–d; E-06 | Existing separation is intentional; F-06 identifies a consumer-policy question. |
| Owner-local fixtures plus aggregate mention count | testsupport/performancefixture | Links fixture facts; test-util application fixture assembler for aggregate checks | defer | E-12; F-03 | No move until descriptor/receipt consumer scope and duplicate checks are characterized. |
| HTTP/authentication/WebSocket behavior | Workbook, source-owner facades, Revisions, Collaboration outside target | Existing transport/application/security owners | keep | E-07–11 | Links borrows caller transaction and actor context; it does not authorize independently. |
| Saved views, view contracts, frontend save/conflict state and grid selectors | Authored schemas, source-owner projections and web consumers outside target | SavedViews, view-contracts, frontend/controller and grid-adapter owners | defer | E-03/E-10/E-11 | No target-owned implementation found to relocate. No frontend slice proposed; preserve consumer contracts. |

## 4. Public Contract and Behavior Freeze Map

“Observed” describes inspected code/assertions. “Required” describes the named owner. Disagreement creates a finding; freezing current behavior is not permission to endorse a normative defect. All tests listed here were inspected or mapped, not run. HTTP success/error envelope structure, request identity/idempotency, status classification and transaction atomicity must survive a structural refactor.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| C-01 Directed link identity, exact vocabulary/provenance/confidence, null versus field-owned tuple, active/deleted state | Links; Core 02 link/history rules | E-05 store/commands/field_refs/valuecodec; E-03 | T-01/T-02/T-09/T-13 | Same incident vs foreign/missing/deleted endpoints; full tuple replay/metadata changes; unchanged canonical snapshots and safe errors through Records lookup | high | Observed ten type tokens and five stored provenance tokens. Ordinary command permits manual with null confidence and entity-observation auto_match with 100; retained codec is wider. F-04 must not silently broaden admission. No per-link row_version or narrative fields. |
| C-02 Collection action and tag identity: record_ref/party_ref/record_tag item refs, add/remove/no-op and tag display/normalized values | Links source effects; source-owner admission and field schema | E-05 collection/actions/item_refs; E-07; authored view fields E-10 | T-04/T-10/T-13; Notes E-11 | Preserve invalid item-ref, duplicate add, missing remove, wrong-record removal, field error classification and transaction failure; no new implicit normalization in Store | high | Public tags preserve trimmed display and normalized dedupe asserted in route tests. Store receives RawText/NormalizedText; portability independently validates normalization. Schema admission is above the generic capability. |
| C-03 Host/Identity assignment collections and record_party_assignment.v1/rpa_ source token | Links source identity; Entities writes; Reporting capture | REQ-02-143a.2; E-04/E-05/E-07 | T-11 fixed digest, snapshot and invalid matrix | Preserve all six field/type combinations; no scalar bypass; same-snapshot facts; missing/foreign/deleted Party and read-failure atomicity after lookup change | high | Required exact field prefixes and three suffixes; token equality only, UTC microseconds and explicit nulls. No invented assignment inference from arbitrary references. |
| C-04 Revision/change-set values, immutable before/after maps, history_entry_ref meaning, inverse effect and affected-record versions | Links owns non-row semantics; Revisions owns history/rollback transaction; Records owns versions | E-05 contribution/provider/codecs; E-08; Core 01 §3.3.4.2 | T-01/T-03/T-05/T-10; E-11 endpoint rollback/history | Deleted-but-retained endpoints vs missing/foreign identity; failure before/after inverse must leave no partial mutations/revisions/projections/events; preserve affected record IDs and changed keys | high | GET /api/v1/records/{record_id}/history; POST /api/v1/records/{record_id}/rollback; DELETE record and POST restore affect link visibility/rebuild. Do not reinterpret providers as owning HTTP authorization. |
| C-05 Projection dependency closure, refresh timing, failure propagation, row/query cells | Links dependency/effect facts; Projections execution; source owners' parent mutation | E-05 projection_*; E-09; REQ-01-658/662 | T-10 tag/evidence cells, row_version, query and rollback; T-09 active views | Inject refresh/read failure; verify mutation/history/event atomic rollback and unchanged sorted endpoint set; preserve restored/deleted endpoint dependency distinctions | high | NewMutationStore requires refresh port; NewStore is deliberately used by broader coordinators. Tag-only refresh belongs to parent coordinator. No async refresh redesign. |
| C-06 Notes association direction/kind/paging/replay | Artifacts policy, Links storage, Workbook routes | REQ-01-675; association_lookup.go; E-07/E-10 | E-11 Notes integration; T-14 contextual linked-note browser scenario | Preserve null-field lookup, ordering/cursor tie-break, public limits, duplicate add no-op, replay with changed role/closed state and current visibility | high | GET and POST /api/v1/records/{note_record_id}/note-associations; POST /api/v1/records/{record_id}/linked-notes. Internal 501 lookahead must not become a public limit. |
| C-07 Merge/rollback graph and tag repointing in one change set | Entities merge and Revisions rollback coordinators; Links source effects | Core 02 REQ-02-066; E-05 mergeeffects; E-06/E-08 | T-13 full field-aware dedup; E-11 rollback; peer merge rows discovered | Characterize manual, auto_match, import/system/rollback retained provenance; collisions/self-links; unchanged dedup counts, history and inverse graph | high | F-04 observed validator path may reject valid retained non-manual links during repoint. No correction authorized; do not preserve an inferred defect as the desired contract. |
| C-08 Bundle v8 source roundtrip, strict members/invariant order, attribution and recovery exclusion | Links sourceport; IncidentBundles orchestration; Recovery lifecycle | E-05 sourcestate; E-09 catalogs; Core 01 portability owner | T-06/T-07/T-08/T-12/T-13 | Preserve exact export bytes/order, empty collections, retained tombstones, actor attribution, no writes on invalid preparation, scratch commit isolation | high | data/record_links.ndjson, data/tags.ndjson, data/record_tags.ndjson; two authoritative tables, three transient scratch tables. No new legacy compatibility or migration. |
| C-09 Active fact ordering, safe errors, Reporting field/support conversion and Timeline fact conversion | Links facts; app adapters; Reporting/Timeline consumer policies | REQ-RPT-026b–d; E-05/E-06; REQ-01-662 | T-13 non-nil empty/no-partial facts; T-11 assignments; consumer tests require targeted discovery | Direct adapter tests for ordered repeated targets, missing target fallback, assignment exclusion question F-06 and error/no-partial output; retain Timeline incoming supersedes selection | high | Reporting paths/DTOs remain outside Links. F-06 is observed consumer divergence requiring owner review, not a Links module move. |
| C-10 HTTP row/query/mutation envelopes and authorization | Workbook transport; Timeline/Artifacts/Entities/etc. source facades; incident admission | E-07/E-10; Core 01/Core 04 | T-10 HTTP projection/history; E-11 Notes role/closed/revoked tests | Preserve foreign-incident/not-visible vs invalid-target failures, editor admission, replay, base_row_version/client_txn_id and rejected mutation atomicity | high | POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows and /query; PATCH /api/v1/records/{record_id}; /bulk-mutations, /clipboard-paste, record /supersede and /conflicts/{conflict_token}/resolve are discovered indirect Workbook entry points. Their full branches were not audited; select owner tests if touched. |
| C-11 WebSocket record_changed path, fields, versions and publication | Collaboration transport/stream; source mutation/Revisions intent production | E-10 socket route/transport; E-08; T-10 | T-10 checks changed_field_keys after tags/evidence and rollback; E-11 affected endpoint events | Preserve same /ws/v1/incidents/{incident_id} path, current authorization/revocation, event timing after atomic changes, changed keys and row versions on all affected records | high | Links registers no socket and must not publish independently or change event envelope/authorization. |
| C-12 Test-util row/query/mutation/save/conflict/inspector support and harness evidence accounting | Existing test-util adapters and harness; production contracts retain source owners | E-11/E-12/E-14; route helpers in target | T-10/T-14; fixture assembler row in §8 | Preserve fixture IDs/association counts/receipts, row and query assertions, selector ownership and mapping if tests move; discover additional save/conflict/inspector rows only for a touched consumer | medium | Harness rows and historical phases are accounting only. No test-util mutation/save-state/inspector implementation exists inside Links. No test-only runtime dependency found. |
| C-13 View schema/saved-view behavior, generated protocol and UI selectors/grid integration | View-contract sources, SavedViews, frontend shell/controllers and grid adapter | E-03/E-10/E-11; E-13 generated policy | T-10 query cells; T-14 Notes selectors/imported UI/view contracts | Confirm no schema/selector/route generated diff for backend-only slice; if behavior/columns change, stop and map affected SavedViews/grid owner tests before proceeding | medium | timeline.tags, timeline.attached_evidence_ids, note.tags are collection action fields; note.linked_record_count is read-only. No direct grid vendor import or frontend implementation in target. Generated protocol/view/UI packages are indirect consumers, never hand-edit candidates. |

## 5. Coupling and Boundary Findings

Classifications prioritize later work; they do not authorize it. A machine allowlist differing from Core is implementation/projection drift, not BLOCKED: owner contradiction.

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| F-01 Standalone Records lookups inside Links violate the Records port allocation | REQ-01-649; store.go endpoint count and records-only self-join; collection_actions.go target EXISTS; revisionprovider/provider.go endpoint/type queries; E-08 Records.LoadEnvelopeTx/LoadEnvelopesTx | high: active/deleted, same-incident and error behavior can change during extraction | must_fix | Records lookup; Links semantic validation | Plan S-01–S-03. Preserve no-lock reads where current SQL does not lock; retain caller transaction, current errors and active-vs-retained distinctions. Existing backend allowlist admits these paths, so a passing static check alone would not establish conformance. |
| F-02 Source-owned inverse provider contains record-type/view-field switches and artifact subtype lookup | revisionprovider/provider.go; Revisions rollback_apply_nonrow.go derives public changed keys from before/after live rows | high if removed as “duplicate” without checking error/API effects | defer | Links source contribution with Records/Artifacts facts and Revisions public delta calculation | TODO: inspect all ApplyInverseResult.ChangedFieldKeys consumers and artifact lookup failure cases before a separate cleanup. Core assigns source history meaning to source owners; do not move all rollback logic to Revisions. Not a gate for S-01; preserve in S-02. |
| F-03 Links fixture validation reads entity_mentions and overlaps aggregate fixture semantics | testsupport/performancefixture/production.go and E-12 app fixture production/assembly | medium: count, descriptor and receipt drift | defer | Links owner-local fixtures; application test-util assembler for aggregate checks | Keep current contribution. A later fixture task must prove whether the cross-owner mention count is redundant, identify exact receipt consumers and retain the assembler row. No production runtime import from test support was found. |
| F-04 Merge re-creation reuses ordinary-write provenance admission | merge_effects.go passes validateRecordLinkCommand into private mergeeffects; re-create retains source provenance; validator accepts manual/auto_match only, codec/portability admits import/system/rollback; Core 02 REQ-02-066 | high: inferred rejected merge for admitted retained source; not reproduced | should_fix | Links merge source effect; Entities merge orchestration | RB-002: characterize reachability and exact outcomes before deciding repair. S-04 is a separate behavior correction requiring later authorization; never broaden ordinary write validation as an incidental refactor. |
| F-05 Existing SQL, platform primitives and cross-owner contribution interfaces are not blanket defects | Own record_links/record_tags SQL; canonicaljson assignment hash; fieldnorm portability validation; recoverystate/sourceport/historycontract/rollbackcontract; E-05/E-08/E-09 | high if “purity” cleanup changes source meaning or transaction ownership | intentional/no_action | Links with narrowly allocated shared contracts | Keep private codec SQL-free, source adapters owner-local and app composition responsible for peer DTO conversion. Do not introduce a generic shared domain helper or transfer portability to tabular Imports. |
| F-06 Reporting support adapter excludes recognized assignment fields before accepting references_record links | internal/app/reportingassembly/links.go CollectSupportRefsTx; adopted REQ-RPT-026d says exactly the active eligible link types, ordered repeated targets and fallback; REQ-02-143a.2 allocates assignment source semantics | high consumer-policy discrepancy; source facts still emitted separately | should_fix | Reporting consumer contract and app reportingassembly | RB-003: review the assignment exclusion against the adopted owner and relevant materialization tests. Observed code differs from the inspected eligible-type rule; no conflicting primary-owner clause was found. Keep this out of structural slices; any changed selection is S-05, requires later authorization. |
| F-07 Auth, version/history, projection and publication duties are split across callers and Links capabilities | NewStore/NewMutationStore comments and methods; E-07 owner facades; E-08 revisionassembly; E-09 projection runtime | high if an omitted hook is mistaken for a missing feature or auth added in storage | intentional/no_action | Source application owner/incident admission; Links effects; Projections; Revisions; Collaboration | Freeze C-04/C-05/C-10/C-11 with failure characterization. Do not make Store start transactions, authorize HTTP users, append parent history or publish sockets. |
| F-08 No direct grid-vendor integration, frontend shell state or saved-view runtime in target | Complete 45-file inventory/import scan; E-10 authored view fields and E-11 browser | medium indirect consumer risk | intentional/no_action | Existing frontend/grid/view/SavedViews owners | Preserve C-12/C-13. No frontend extraction or grid change justified. Test-utils/browser labels do not imply Links owns those workflows. |
| F-09 Generated boundaries and harness accounting are downstream; labels can mislead | E-13 generated policy/topology; E-14 maps TestTypedLinksAndTags_Unit to integration and browser Notes scenario to module.links; incident_portability verification has machine profile base | medium evidence misclassification | intentional/no_action | Authored harness/verification sources and their generators | Use exact selectors and execution metadata, not test suffixes, historical phases or profile labels, for selection. Machine base allocation does not redefine extension conformance. No generated edits; S-03 inventories any needed authored mapping changes. |
| F-10 Framework candidate ownership/workflow list is broader than actual target; prior command evidence aged | E-01 framework; actual Store/private contributions; session HEAD comparison; current markdown-biome recipe | medium planning drift | intentional/no_action | This tracker | Record mismatch: target has no local HTTP, frontend/grid or saved-view implementation, and no generic one-file service facade to split. Use target-specific slices; current lint is Biome, not the earlier recipe. No historical framework assumption is promoted to behavior. |
| F-11 Exact-path SQL read allowances may be too coarse to demonstrate the required exact read shape | E-13 records-current-envelope-access permits entire Links files and sourcestate subtree; REQ-01-649 requires exact path and read shape | medium: future standalone reads could slip through | should_fix | Backend boundary authored contracts/checker; Records allocation | S-03 must inspect existing guard expressiveness and narrow obsolete exemptions after migration. TODO: determine whether remaining joined reads already have sufficient separate shape declarations before proposing a checker/schema change. |

## 6. Refactor Workstreams

This is the planning dependency graph, not a runtime module hierarchy:
WF-00 → WF-01 → WF-02 → (WF-03 and WF-04) → WF-05 → WF-06 → WF-07 → WF-08.
Independent analysis branches join before ownership and sequencing. Completion of these planning workflows does not complete any implementation slice.

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WF-00 | Session/source bootstrap and tracker initialization | root | none | WF-01 | Admit exact output, source hierarchy, clean baseline and write restriction | E-01–04; this tracker | Read-only Git/path/owner inspection | DONE when path identity, authority and baseline are recorded; §1/§10. |
| WF-01 | Complete Links inventory | chain | WF-00 | WF-02 | Account for all 45 files including private/support code | Every §2 path; targeted callers E-06–12 | File enumeration against table; exact source reads | DONE when 45/45 paths have responsibility, calls/dependencies, contracts and test posture. |
| WF-02 | Contract-owner mapping | chain | WF-01 | WF-03, WF-04 | Distinguish source effects from consumer/runtime owners | E-02–11 | Compare observed C-01–13 with owner clauses | DONE when each discovered contract has owner, evidence and test-gap disposition; contradictions explicitly blocked if found. |
| WF-03 | Characterization test gap analysis | parallel | WF-02 | WF-05 | Identify minimum tests before a Records seam and separate behavior questions | Links tests, Records/Revisions/Artifacts tests, E-14 | Static assertion review; future V-01–V-05/V-10–V-12 | DONE as analysis when each gap has a contract/slice/entry gate; passing evidence remains TODO. |
| WF-04 | Coupling/boundary scan | parallel | WF-02 | WF-05 | Classify standalone SQL, contribution dependencies, fixture and consumer risks | E-05–09/E-12/E-13 | Trace imports/SQL/callers and primary owners | DONE when F-01–F-11 have classification, proposed owner and next action. |
| WF-05 | Facade/ownership recommendation | chain | WF-03, WF-04 | WF-06 | Recommend a Records lookup seam while retaining source-owned contributions | Store/collection_actions/revisionprovider; Records store | Compare alternatives against C-01–13 and F findings | DONE when narrow recommendation and no-action/defer alternatives are explicit; no public signature change presumed. |
| WF-06 | Slice sequencing | chain | WF-05 | WF-07 | Separate characterization, ordinary-read and inverse-read changes from behavior repairs | §7 candidates | Dependency/readiness/rollback audit | DONE when each slice has entry gates, exact scope and binary exit. No implementation authorized. |
| WF-07 | Harness/test/accounting update plan | chain | WF-06 | WF-08 | Assign canonical commands and preserve coverage identity | E-13/E-14; any later authored boundary/test maps | Command/row definitions inspected; V-06/V-07 planned | DONE when known rows/commands and unresolved selection work are recorded; do not encode architecture as phases. |
| WF-08 | Validation and final handoff | chain | WF-07 | none | Audit tracker coverage and allowed-write boundary | This tracker; read-only content/index baseline | Structural audit, read-back, Git diff check; V-09 explicitly skipped | DONE when §12 passes as documentation accounting and final baseline shows only tracker addition; product readiness remains separate. |

## 7. Proposed Refactor Slice Plan

These are candidate plans, not implementation patches. All implementation, including tests and authored guard changes, requires a later authorized task. No public API, type, schema or generated artifact change is authorized now. Readiness is separate: S-00 is defined for a later authorized characterization task; S-01/S-02 cannot start before its passing baseline/gaps; S-03 depends on both; S-04/S-05 are separate behavior work and blocked by their named evidence gates.

Recommended design direction for F-01: consume the existing Records-owned caller-transaction envelope lookup capability and leave Links responsible for link/collection interpretation. Records.LoadEnvelopesTx already supports caller pgx.Tx and lock=false, returns existing envelopes and omits absent IDs; Links must translate those results to its existing active/same-incident/type checks and error categories. Preserve transaction and constructor behavior; if a narrow private adapter can use the existing root Records capability without changing signatures, prefer it. TODO: confirm the smallest constructor/composition wiring during S-00; a new public capability or signature is not presumed. Do not substitute locking reads, active-only queries for rollback existence, or active-fact readers for restore dependency closure.

“No broader refactor warranted” is an acceptable outcome for codecs, source-state contributions, projection ports, assignment facts and test support. If new evidence removes the need for a candidate, mark it DROPPED with evidence rather than implementing cosmetic motion.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| S-00 | Later implementation authorization; current source recheck | Characterize envelope validation/rollback boundaries and capture passing selected baseline before changing access. Decide minimal internal wiring and record exact consumer coverage. Ready for authorized test work; not executed. | Links store_test.go, party_assignment_facts_test.go, route_projection_history_test.go; Revisions tests; Records store contract; app adapters only for review | C-01–05/C-07/C-10/C-11 error/transaction semantics | Add missing foreign/missing/deleted/wrong-type endpoint cases and lookup/refresh failure atomicity; preserve existing tuple/confidence/copy and row/event assertions. Add tests only for uncovered behavior, not implementation mirroring. | V-01, V-02, V-03, V-04, V-10, V-11; discover additional owner selections via task-guide if needed | Revert newly authored characterization/accounting together if the expected behavior was wrong; retain diagnosis in tracker. No source-state changes. | Passing pre-move evidence covers each planned lookup's success/error/deleted-state behavior; exact constructor and consumer selection recorded; RB-001 closed for next slice. |
| S-01 | S-00 | Replace standalone envelope reads in ordinary Links endpoint/collection validation with Records-owned lookup; retain Links validation, error mapping and existing constructor behavior. Behavior-preserving candidate. | internal/modules/links/store.go, collection_actions.go; private Links adapter if necessary; internal/modules/records/store.go consumed, not redesigned; composition sites only if S-00 proves needed | C-01–03/C-05/C-10; no extra row locks or wider authorization | Preserve T-09/T-11/T-13 and T-10; S-00 cases must continue to pass, including unknown ordinary field policy being admitted upstream | V-02, V-03, V-04, V-06; V-10 for reused Records capability | Revert lookup/wiring as one slice, preserving characterization. No migrations or data rollback. | No standalone records SQL remains in these two Links files; owner port and tests preserve errors, tuple behavior, refresh scope and transaction ownership. |
| S-02 | S-01; rollback characterization from S-00 | Replace standalone current-envelope reads in private non-row inverse provider with Records port. Keep source-owned inverse logic and existing changed-field/artifact-subtype behavior. Behavior-preserving candidate. | internal/modules/links/internal/revisionprovider/provider.go; root revision_provider_contribution.go wiring only if needed; E-08 composition and existing Records capability | C-04/C-05/C-07/C-11, especially retained/deleted endpoints, affected IDs and failure rollback | Preserve T-05/T-10 plus Revisions endpoint-history/event assertions, changed-field completeness and S-00 missing/foreign/deleted cases; inspect fault-injection row before selecting if changed failure path requires it | V-02, V-04, V-06, V-10, V-11; V-12 where the selected rollback delta applies | Revert provider lookup and any constructor wiring together; retain tests; no history rewrites. | No standalone records lookup remains in provider; retained inverse state and public event/history behavior unchanged with passing evidence. F-02 cleanup remains deferred. |
| S-03 | S-01, S-02 | Narrow obsolete Records SQL exemptions and reconcile any genuinely changed test/caller accounting. Inspect remaining joined-read declarations before changing guard expressiveness. No runtime behavior change. | tools/backend_module_boundaries.json; existing authored boundary guard/test inputs discovered at implementation; tools/test_families/module.links.json only if selectors actually change; generated derivatives only through Make | C-12/C-13; accidental exemption removal or misrouted tests | Preserve owner allocations, all 14 Links rows and required collaborator coverage; demonstrate standalone reads rejected while valid own-source joins remain permitted | V-06, V-07; V-01 after authored mapping changes; V-08 only in final implementation portfolio | Revert authored contract and its Make-generated projections together; never hand-edit a generated rollback. | Obsolete exemptions removed/narrowed, remaining read shapes accounted for, no dropped tests, generated drift clean. If checker change is needed, identify its owner/tests before implementation (RB-004). |
| S-04 | RB-002 resolved; separate later behavior authorization | Characterize and, only if confirmed against owner requirements, correct merge re-creation of admitted non-manual provenance without broadening ordinary-write admission. requires later authorization | Links merge_effects.go/internal/mergeeffects; merge_test.go; Entities merge orchestration/tests if transaction outcome affected | C-01/C-04/C-07; behavior correction, not structural equivalence | Retained/imported manual/auto_match/import/system/rollback matrices, field-null/non-null duplicate and inverse graph/history cases | V-03 plus make task-guide OWNER=module.entities; TODO: exact additional merge rows after characterization/consumer inspection | Keep correction separate from structural slices; revert code/tests as a unit without mutating historical rows. | Reproduced case, owner-reviewed expected outcome, independent authorization and exact tests pass. Otherwise remain DEFERRED or DROP with evidence. |
| S-05 | RB-003 resolved; separate later behavior authorization | Resolve Reporting assignment support-ref selection against adopted REQ-RPT-026d and characterize exact output; any selection correction requires later authorization | internal/app/reportingassembly/links.go; Reporting/app adapter tests; Links source facts preserved | C-03/C-09; snapshot support policy and disclosure consumers | Exact order, repeated target retention/fallback, assignment vs ordinary reference output, no-partial failure; relevant immutable capture consumers | make task-guide OWNER=module.reporting; TODO: exact test-slice rows require inspection of affected adapter/materialization tests | Separate behavioral patch; no rewrite of frozen snapshots or retained output. | Owner disposition and reproduced adapter expectations recorded, exact test scope selected, separate authorization obtained before behavior changes. |

No slice moves frontend state, changes transport routes, hand-edits generated clients, renames an authored migration, changes existing bundle versions, or promotes a test fixture into runtime architecture.

## 8. Validation Plan

Commands were discovered from Make-owned authored/generated definitions and the test-operations skill; none below was executed in this documentation task. Run future commands from /home/jochi/code/cartulary. V-IDs are references to commands, not shell aliases. Product runs must retain the emitted run/summary paths and exit status; inspected assertions alone are not validation. Before a later implementation session, refresh selection with V-01 and make help ALL=1, and follow current test-ops instructions for run identity/observation.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| selection (V-01) | make task-guide OWNER=module.links | Current owner rows, collaborators and narrow portfolio | yes | Discovered, not run. Use DETAIL=rows or DETAIL=coverage only when needed. Other owners must be discovered before expanding. |
| unit (V-02) | make test-slice OWNER=module.links SUITE=backend-unit | Current Links unit selection, T-01–T-08 | yes, baseline before moving behavior | No run. SUITE enum verified in authored command inputs. Tests added later need matching authored selectors. |
| integration/store (V-03) | make test-slice OWNER=module.links ROWS=module.links.integration.active_links_and_tags_views_v1_contract_eaf4aac3b7,module.links.integration.party_assignment_source_facts_preserve_snapshot_963c86bbbb,module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1 | T-09/T-11/T-13 endpoint, assignment, tuple/merge/portability coverage | yes for S-01 and relevant baseline | No run. Service needs/isolation come from row map; no direct go test substitute. |
| integration (V-04) | make test-slice OWNER=module.links ROWS=module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906 | T-10 route/projection/history/query/socket/rollback | yes before structural movement | No run. Dedicated PostgreSQL and object-store fixture per map. |
| e2e/browser (V-05) | make test-slice OWNER=module.links ROWS=module.links.browser.the_notes_tab_supports_browser_visible_creation_ae6bdcc2da | T-14 linked-note creation in Chromium | no before a pure lookup edit; required in applicable final consumer portfolio | No run. Preserve browser row; do not infer ownership from visible Notes tab. No visual golden changes planned. |
| import-boundary/static (V-06) | make lint CHECK=backend-boundary | Authored backend dependency/SQL rules | yes as baseline; after S-01–03 | No run. Existing allowlist can pass despite F-01; code/owner comparison remains required. |
| generated drift (V-07) | make source-check CHECK=generated | Generated projection consistency | no for tracker; after authorized authored map changes and final slice | No run. Regenerate authored changes with make generate only in later authorized work; never hand-edit generated policy/topology outputs. |
| full check (V-08) | make check | Broad implementation portfolio, subject to task-guide coverage | no before tracker; later end-of-implementation if needed | No run. Implementation procedure uses make agent-finalize before broader final verification; it may rewrite artifacts and is prohibited here. Avoid redundant check/ci portfolios; incident_portability verification's public_target ci is an accounting obligation to evaluate, not proof every local slice must run both. |
| documentation (V-09) | make lint-markdown | Repository Markdown lint | no before drafting; normally required for docs completion | SKIPPED: current recipe writes .cache/cartulary/static-analysis/outputs/lint-markdown.ok, cache metadata and run artifacts and has readiness prerequisites. Explicit tracker-only restriction prevents this invocation. No substitute claimed equivalent. |
| Records integration (V-10) | make test-slice OWNER=module.records ROWS=module.records.envelope.store | Existing envelope port contract | yes before using it in S-01/S-02 | No run. Refresh current owner selection first; do not rewrite Records semantics to accommodate Links. |
| Revisions integration (V-11) | make test-slice OWNER=module.revisions ROWS=module.revisions.integration.covers_delete_restore_single_entry_rollback_dire_b348439e6d | Existing non-row inverse/history/affected-record event path, E-11 | yes for S-02 baseline | No run. Additional merge/fault rows discovered but not automatically selected; inspect changed scope first. |
| Revisions changed fields (V-12) | make test-slice OWNER=module.revisions ROWS=module.revisions.rollback.changed_field_completeness | Public-cell delta completeness | yes for S-02 baseline | No run. Complements integration; does not prove every provider error path. |
| portability isolation (V-13) | make test-slice OWNER=module.links ROWS=module.links.store.bundle_validation_commit_isolation | T-12 concurrent scratch commit isolation | no for envelope-only edit unless source-state path changes; retain final owner coverage | No run. No source-state refactor proposed. |
| fixture accounting (V-14) | make test-slice OWNER=harness.browser ROWS=harness.browser.unit.source_owner_contribution_assembler | Deterministic closed assembler/receipts | no; only if F-03 becomes authorized work | No run. Test bodies inspected in E-12; not permission for fixture modification or performance claims. |

The exact current Links map is below. All rows are active with implementation claim posture. These are evidence/routing identities, not module or phase boundaries. T-01–08 are unit; T-09–13 are integration even when their IDs contain “store”; T-14 is browser. No passing result is attached to any row.

| Test ID | Exact row ID | Assertion/selection posture |
| --- | --- | --- |
| T-01 | module.links.unit.canonical_valuecodec_is_pure_and_exact_9181a7cf12 | Six codec tests: canonical fresh maps, strict link/tag negative matrices, reject legacy shapes, retain every canonical field, history operation matrix. |
| T-02 | module.links.unit.closed_vocabulary_exact_round_trips_3b5f179c21 | Exact token roundtrip/rejection. |
| T-03 | module.links.unit.immutable_mutation_value_isolated_96e35c81c9 | Constructor grammar and recursive isolation. |
| T-04 | module.links.unit.links_item_refs_are_canonical_ade3870b93 | Canonical/strict item refs. |
| T-05 | module.links.unit.links_revision_contribution_owns_current_target_08116fb575 | Two Links revision targets and history validators. |
| T-06 | module.links.unit.source_state_contributions_remain_exact_8d3f67c521 | Exact portability/recovery contributions. |
| T-07 | module.links.unit.source_state_manifest_is_immutable_and_exact_2e516fd3a4 | Exact/copy-safe descriptor and malformed authoring matrix. |
| T-08 | module.links.unit.strict_portability_preparation_fails_closed_8617ab9352 | Strict NDJSON, opaque copies and validation priority. |
| T-09 | module.links.integration.active_links_and_tags_views_v1_contract_eaf4aac3b7 | TestActiveLinksAndTagsViewsV1Contract and TestRecordLinkOwnerValidation. |
| T-10 | module.links.integration.link_and_tag_mutations_atomically_update_project_03f8283906 | TestLinkTagProjectionHistoryQuery_Integration. |
| T-11 | module.links.integration.party_assignment_source_facts_preserve_snapshot_963c86bbbb | TestPartyAssignmentSnapshotSource. |
| T-12 | module.links.store.bundle_validation_commit_isolation | TestLinksBundleValidationCommitIsolation_Integration. |
| T-13 | module.links.store.typed_links_and_tags_persist_as_structured_rows_8daeea1ad1 | Field-aware identity; unknown-member import rejection; strict source roundtrip; merge dedup; TestTypedLinksAndTags_Unit. |
| T-14 | module.links.browser.the_notes_tab_supports_browser_visible_creation_ae6bdcc2da | apps/web/e2e/notes.authoring.spec.ts: Notes tab creates artifact-backed linked notes; scenario_f62103626873, Chromium, webserver_backed. |

contracts/verification/owners/module.links.json allocates three exact IDs: module.links.verification.behavior_contract (go_test/playwright), module.links.verification.source_contracts (go_test), and module.links.verification.incident_portability (go_test, public_target ci). All currently have machine profile base; that is reported as allocation, not an extension-conformance conclusion. Preserve mappings unless authored changes are justified and generated through Make.

The [refactor-tracker skill](/mnt/c/Users/jochi/.codex/skills/refactor-tracker/SKILL.md) explicitly says, “Do not run commands that may write outside the tracker.” Together with the user's sole-file restriction, this is why the normally required documentation lint is skipped; this is not a lint pass.

Command-discovery limitations: TODO: exact future Reporting materialization/adapter and additional Entities merge selectors require source inspection after RB-002/RB-003 disposition. TODO: exact checker tests for a possible read-shape guard expansion require discovery under RB-004. Searches covered Makefile, task_surface_owner.json, current Make/help projections, execution topology, test-family and verification owner maps. No guessed command or direct implementation script is proposed as canonical verification.

## 9. Top-Level Work Tracker

Statuses describe the named work item, not overall product conformance. Planning DONE does not mean the corresponding code was changed or its tests passed.

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| W-01 | Admit links output/authority and reconcile advanced HEAD | WF-00 | DONE | none | §1, E-01–04 | Correct safe path, baseline and write scope recorded. |
| W-02 | Inventory all 45 Links files and direct consumer roles | WF-01 | DONE | W-01 | §2, E-05–12 | 45/45 rows; no unexplained exclusions. |
| W-03 | Freeze C-01–13 with primary owners and test posture | WF-02 | DONE | W-02 | §4 | Every discovered risk has evidence and coverage/gap disposition. |
| W-04 | Identify Records seam characterization gaps | WF-03 | DONE | W-03 | S-00 and RB-001 | Explicit entry tests and commands recorded; execution remains W-10. |
| W-05 | Classify SQL/rollback/fixtures/Reporting coupling | WF-04 | DONE | W-03 | F-01–11 | Severity, candidate owner and next action per finding. |
| W-06 | Recommend retained source boundary and limited envelope seam | WF-05 | DONE | W-04, W-05 | §3/§7 | Clear keep/split/defer findings without unauthorized API redesign. |
| W-07 | Sequence structural slices separately from corrections | WF-06 | DONE | W-06 | S-00–05 | Dependencies, entry gates, validation and rollback notes complete. |
| W-08 | Map Make validation and preserve all 14 Links rows | WF-07 | DONE | W-07 | V-01–14, T-01–14 | Exact known commands and explicit unresolved selector discovery. |
| W-09 | Create/read back tracker and verify sole output | WF-08 | DONE | W-08 | This file; baseline/audit in §10 | Final artifact and repository/index comparison recorded. |
| W-10 | Obtain passing pre-move characterization evidence | WF-03 | TODO | Later authorized implementation task | S-00; RB-001 | Selected baseline and uncovered cases pass with retained results. |
| W-11 | Route ordinary and inverse standalone reads through Records | WF-06 | BLOCKED | W-10 | S-01/S-02; F-01 | Later authorization, preserved behavior evidence and no standalone Records SQL in named sites. |
| W-12 | Narrow allowance/accounting projections after access change | WF-07 | BLOCKED | W-11 | S-03; F-11; RB-004 | Exact remaining read shapes and intact tests verified; no hand-edited generated files. |
| W-13 | Characterize non-manual merge provenance | WF-03 | DEFERRED | Separate authorized work | F-04; RB-002; S-04 | Reproduce reachable case and establish owner-required result before correction. |
| W-14 | Resolve Reporting assignment support-ref discrepancy | WF-02 | DEFERRED | Separate authorized work | F-06; RB-003; S-05 | Adopted owner disposition plus exact consumer tests, before any selection change. |
| W-15 | Evaluate redundant changed-field mapping and fixture aggregate checks | WF-04 | DEFERRED | Separate evidence-backed task | F-02/F-03 | Prove consumer/error/receipt effects before proposing another slice. |
| W-16 | Generic facade rename, frontend/grid extraction or broad module split | WF-05 | DROPPED | W-06 | §3; F-05/F-08/F-10 | No supporting target evidence; reopen only with new ownership evidence. |
| W-17 | Run Markdown lint under tracker-only write scope | WF-08 | BLOCKED | A task permitting required lint artifacts | V-09; RB-005 | Canonical lint can run without violating authorized write boundary. |

## 10. Session Handoff Log

Time entries use America/New_York (UTC−05:00). LNK-2026-11-09-01 is the supplied prior-turn handoff, not a recovered disk log; no tracker existed then. LNK-2026-11-09-02 is this continuation. “Touched” always means this tracker only. Other paths are read-only inspection. No prior handoff history was overwritten.

Scope and authority:

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-09, prior turn; exact time not retained | Codex / LNK-2026-11-09-01 | Read-only initial plan at ce30337c032a68d145c7ddc061979f655c908951 | Framework/instructions, domain/Core excerpts; no touched file | File reads/searches and Git inspection, per supplied handoff | 45 tracked target files, clean tree; no tracker written in Plan Mode | Full source inspection unfinished then | Continue in authorized documentation mode. |
| 2026-11-09, continuation | Codex / LNK-2026-11-09-02 | Framework first; current main a43248fbe3f762b0f448993126f74172f597f101; output absent before write | E-01–04; touched only this tracker | cat/sed; rg/rg --files; git --no-optional-locks status/rev-parse/ls-files/diff; inline read-only Python | Sources/authority reconciled; target unchanged since prior HEAD; no target-relevant primary-owner contradiction established | RB-001 concerns implementation readiness, not tracker authorship | Recheck HEAD/authority at next session; stop dependent work if primary owners contradict. |

Backend module boundary:

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-09, continuation | Codex / LNK-2026-11-09-02 | 45/45 files reviewed; source owner retained; exact standalone Records queries isolated | All §2 paths, E-06–09/E-13; touched only tracker | Exact file reads; import/symbol/SQL searches | F-01 supports narrow Records seam; contribution/private-package names do not justify moving them | RB-001, RB-004; F-02/F-03 deferred | Start S-00 only in later implementation task; preserve nonlocking vs retained-state distinctions. |

Frontend module boundary, indirect consumers only:

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-09, continuation | Codex / LNK-2026-11-09-02 | No frontend/controller/grid implementation found in target | E-10 view fields/routes; E-11 browser; target imports; touched only tracker | Source/schema reads and rg import search | C-12/C-13 freeze query cells, save/conflict/inspector consumers, selectors and generated surfaces; no frontend slice | No blocker to backend planning; full saved-view/grid behavior not audited | Preserve T-14; open owner-specific tests only if a later slice touches those surfaces. |

Contract and codegen:

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-09, continuation | Codex / LNK-2026-11-09-02 | Canonical mutation/source contracts and authored route/schema surfaces mapped | E-02–05/E-09/E-10/E-13; touched only tracker | Exact reads plus read-only JSON extraction | No generated target files; indirect protocol/view/UI consumers frozen; bundle v8 and retained history shapes recorded | RB-002/RB-003 only for separate behavior decisions | Keep schemas/bytes unchanged in structural slices; author inputs and use Make if later projections must change. |

Tests and harness:

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-09, continuation | Codex / LNK-2026-11-09-02 | 14 Links rows mapped; assertion review separated from execution | E-05/E-11–14; touched only tracker | Reads/searches/JSON inspection; no Make target, product test, generator, format or agent-finalize | V-01–14 discovered. No retained run/summary exists for this session. Failed speculative file reads were corrected through rg --files (Core 02 filename and verification/support-map locations); no mutation resulted. | RB-001 baseline; RB-005 lint write restriction | Later use task-guide and exact rows; record emitted run roots/exits. Do not claim inspected tests passed. |

Security and authorization:

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-09, continuation | Codex / LNK-2026-11-09-02 | Incident visibility/roles and request/socket admission remain above Links | Core 04 excerpts; E-07/E-08/E-10/E-11; touched only tracker | Source/owner/negative-assertion reads | C-10/C-11 preserve editor admission, foreign/not-visible outcomes, closed/revoked behavior and current socket authorization; Links must not gain ad hoc ACL policy | No authorization blocker to planning; runtime role/socket suites unrun | Characterize lookup error equivalence and failure atomicity before source movement. |

Open risks and next session:

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-11-09, continuation | Codex / LNK-2026-11-09-02 | Planning draft assembled; implementation unauthorized | This tracker; source/index baselines read | In-memory drafting; read-only baseline hash of 9,473 tracked/nonignored files and index; no cache-producing command | Initial aggregate SHA-256 004f01b6e1c3cbe865481a2fcb59b91941a47cb78f3621854684216f3d0e96ec; index SHA-256 b08906fe70f4b62c329b2820def871bc147824df6228b18d86380b7f15d5acf8. Final comparison pending below. | RB-001–05 with limited scopes in §11 | Complete read-back/coverage and final allowed-write audit, then hand off S-00 as later work. |
| 2026-11-09 11:03 −05:00 | Codex / LNK-2026-11-09-02 | Tracker created; planning accounting complete | Read back this tracker; sole file touched | Inline read-only Python structural/coverage audit; Git status and content/index comparison; git --no-optional-locks diff --no-index --check -- /dev/null docs/handoffs/links-module-refactor-tracker.md | 12 sections, exact 45-file inventory, 14 mapped Links rows, 7 handoff tables, consistent table columns/statuses and no unresolved ID references. Outside-tracker content/identity aggregate, index and HEAD exactly match baseline; status shows only this new file. Initial whitespace check reported a blank EOF line (exit 3); corrected. Recheck emitted no whitespace diagnostics (exit 1 indicates the new-file diff). This is not a Make lint result. | RB-001–04 remain future implementation gates; RB-005 explains skipped lint. No product check run. | Begin later authorized S-00; keep behavior corrections separate. |

Baseline method: enumerate Git tracked plus nonignored untracked paths, hash file content and identity metadata (mode, inode, link count, size, nanosecond mtime), aggregate in path order, and hash Git index bytes. Exclude only this output from final comparison. Repository .git internals other than index and ignored runtime/cache contents were not exhaustively fingerprinted; no command capable of creating their harness outputs was invoked. This is an edit-scope audit, not runtime verification or release evidence.

## 11. Open Questions and Blockers

These entries block specific implementation/verification actions, not completion of this planning artifact. No broad wishlist is included. No owner contradiction has been confirmed for the proposed structural work; future genuine primary-owner conflicts must be entered literally as BLOCKED: owner contradiction.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | TODO: passing pre-move baseline, uncovered endpoint/failure characterization and exact minimal Records wiring | S-01/S-02 cannot safely preserve active/deleted/type/error/transaction behavior from static review alone | Later implementation authorization; S-00 results for V-02/V-03/V-04/V-10/V-11, exact constructor/caller review | TODO; blocks implementation readiness, not planning completion |
| RB-002 | TODO: can a sourceport-admitted non-manual link reach merge re-creation, and what exact outcome results? | Shared ordinary-write validator may reject preserved historical provenance; correcting it changes behavior | F-04 reproduction through admitted source + merge path; Core 02 merge/source requirements; separate correction authorization | DEFERRED; blocks S-04 only |
| RB-003 | TODO: resolve assignment-field exclusion in Reporting support refs against adopted REQ-RPT-026d | Changing selection can affect snapshot support/disclosure output; existing code is not authority | Reporting owner disposition, adapter/materialization assertion review and reproduction; separate behavior authorization | DEFERRED; blocks S-05 only; not an established owner contradiction |
| RB-004 | TODO: do existing boundary contracts declare sufficient exact shapes for remaining necessary source/envelope joins? | Removing file exemptions safely requires distinguishing authorized joins from standalone lookups | Inspect authored checker/shape rules and tests under tools guidance after S-01/S-02 scope is fixed; discover exact narrow checker row | TODO; gates S-03 guard expansion, not the Records ownership finding |
| RB-005 | Canonical Markdown lint writes outside the sole permitted tracker path | Running it would violate this task's explicit write restriction | A later task permitting lint's cache/readiness/run artifacts; V-09 current recipe | BLOCKED for this session; required documentation lint explicitly skipped |

## 12. Binary Completion Criteria

These checks measure the tracker, not production correctness. Final outcomes are recorded after the read-back audit.

| Criterion | Binary result | Evidence / limit |
| --- | --- | --- |
| Every target file inventoried or explicitly excluded | PASS | Read-back audit matched all 45 exact tracked paths to §2; no missing/extra paths or exclusions. |
| Every discovered public contract risk has owner and test posture | PASS | C-01–13, E-IDs, T-IDs and explicit characterization gaps. Wider consumer audit limits are stated rather than guessed. |
| Every workflow has dependencies and exit criteria | PASS | WF-00–08 and W-01–17; analysis branches join before recommendation. |
| Every proposed structural slice preserves behavior; behavior changes explicitly require later authorization | PASS | S-00–03 have entry/validation/rollback gates; S-04/S-05 explicitly marked requires later authorization. No patch or public redesign implemented. |
| Validation commands discovered or TODO with reasons | PASS | V-01–14; remaining selector/checker discovery is scoped to RB-002–04. No validation success inferred. |
| Primary-owner contradictions explicitly blocked if found | PASS | No target-relevant primary-owner contradiction established. F-01/F-06 concern code/projection divergence, not competing owners. |
| Framework/repository mismatches recorded | PASS | F-10 and session HEAD/lint-recipe reconciliation; candidate framework concerns not fabricated in target. |
| Workstream handoffs sufficient to continue without rebuilding inventory | PASS | Seven §10 categories, evidence register, exact rows, blockers, entry gates and commands. |
| Only authorized tracker content changed; index unchanged | PASS | All 9,473 baseline paths match content/identity aggregate after excluding this tracker; index and HEAD unchanged. Git status shows only the tracker addition. |

Planning completeness: complete as a planning and evidence-accounting document. Implementation slice readiness: S-00 defined for a later authorized task; S-01–03 gated; S-04/S-05 deferred. Implementation authorization: absent. Product validation: not run. Canonical Markdown lint: skipped for RB-005, not passed.
