# Reference Pack remediation implementation

This ledger tracks the coordinated cutover authorized on 2026-10-02. It is implementation support, not a normative owner, an adoption record, or a conformance claim. Tests, generators, runtime metadata, and release evidence must not depend on it.

## Current boundary

**The authorized pre-production remediation is implemented and validated.** HTTP administration and lifecycle Jobs call the canonical application coordinator. The legacy checksum-labelled verifier, mutable store implementation and empty built-in generator have been removed. The live verifier uses bounded confined storage, strict JSON and offline Ed25519 verification. Production HTTP and browser acceptance fixtures use canonical signed containers. The final coordinated release run passed all 1,259 execution units, with no failures, skips or cancellations.

Migrations 00046–00063 implement the immutable stores, exact Reporting binding, identity preflight, audit/diagnostic history, portable catalogs and preparation, retired-store removal, and immutable imported Reporting files. They have been exercised only in disposable PostgreSQL databases. No real deployment has been migrated or reset; no incident identities or historical provenance have been converted. Reference Pack profile major 2, lifecycle Jobs v2, configuration v2, incident-bundle format 5 and recovery projections are implemented.

The production consumer boundary now serves all sixteen profiles and nine algorithm families. Indicators, observations, Network Flow, import and revision mutations use that boundary and transactional usage guards. Native snapshots pin exact sets through rendering. Incident portability validates embedded bytes under destination trust in its parent finalizer, reuses exact local content without changing activation, and retains/re-exports immutable source Reporting artifacts. Recovery validates historical signatures, successful-time freshness and retained source bindings. Physical collection and backup retention now serialize across processes and preserve successful containers, pins and removal history.

The final visual and release gates are complete. Production conformance separately requires owner adoption and the target deployment's cryptographic compliance disposition; this handoff does not claim either. The project owner explicitly removed independent recreation and interchangeability reports from this pre-production cutover on 2026-10-04; those reports are not outstanding implementation work. The focused implementer security review is recorded separately. The [obligation evidence index](reference-pack-obligation-evidence.md), [limit evidence index](reference-pack-limit-evidence.md), [owner/version inventory](reference-pack-owner-inventory.md) and [security disposition](reference-pack-security-review.md) record the implemented boundaries and the distinction between executed evidence and approval. There are 163 fixture manifests across all 27 required families. Reporting identity and admission binding, live concurrency and commit faults, complete reachable boundary recipes and all five closed consumer machine contracts have focused and coordinated release evidence. A fixture count does not establish exhaustive scenario coverage. The subsystem NLSpec remains draft; its formal production conformance gates are unchanged. Dated checkpoints below are historical; this current boundary supersedes their descriptions of missing implementation or validation.

## Fixed decisions

- Replace the legacy pack format without a compatibility reader or dual writes.
- Correct never-adopted schema identifiers in place; increment adopted contracts when observable behavior changes.
- Evaluate freshness once at verification-attempt start, excluding Jobs queue time.
- Reverify and refresh require a successful retained envelope; never-successful candidates require explicit reimport.
- Verification preserves administrative disablement; explicit activation can clear it.
- Complete exact-set consumer, reporting, portability, recovery, and removal integration before claiming completion.
- Never silently discard incident data, invent historical trust, or rewrite indicator identities during migration.

## Current obligation disposition

All rows have passed the final coordinated implementation gates. “Implemented” below identifies live production code, not a production deployment approval or conformance claim. The project owner has scoped this work to pre-production. Exact test routes, retained results and evidence limits are in the linked indexes.

| Plan gap | Implemented boundary / evidence | Pre-production disposition and scope |
| --- | --- | --- |
| A1 | Companion amendments, profile/configuration/Job/public/portability/recovery version inventory and pending OpenAPI change set. | Exact companion content revisions and typed projections are recorded; production adoption is unclaimed. |
| A2 | Strict original JSON, closed nested schemas, explicit null/default policy, normalized alias rejection; nested mutation evidence for manifests, profiles, envelopes, provenance and consumer results. | Nested admission, nullable fields and union branches reviewed against the typed projections; focused and final coordinated checks passed. |
| A3 | All 49 live ordered checks, applicability, bounded exact deduplication, safe diagnostics and durable summaries. | Projection/precedence/privacy review complete; fast/check pass and complete release passed. |
| B1 | Bounded confined ZIP/ustar/GZIP, original names/types, collisions, exact ratios and crash cleanup. Complete structural and configured-limit fixtures are indexed. | Focused implementer security review complete; no additional independent audit is required by the plan. |
| B2 | Live offline Ed25519, root union/both thresholds, 128 root-update signatures, exact historical replay, rollback and complete signer sets. | Focused trust review complete. RP-GATE-010 remains a target-production compliance decision. |
| B3 | Execution-start freshness, clock assertion, one monotonic cohort budget, cancellation/recovery and Extensions finality; queue/expiry/commit-fault evidence. | Fast/check and complete release regression passed; production owner adoption unclaimed. |
| C1 | Candidates, attempts, immutable successful envelopes, independent health/disablement, eligibility and interrupted work. | Transition/projection review and focused plus broad regression passed. |
| C2 | Atomic mixed cohort outcomes, preserved failed renewal, invalidation/fallback, trust/attestation/audit publication and lost-acknowledgement classification. | Shared-finalizer regression passed; focused review includes parent transaction changes. |
| C3 | Durable cohort/dependency/revision captures and ordered guards; live shared-root, activation, usage and pin/removal race evidence. | Focused race evidence and complete release regression passed. |
| C4 | Fixed-point dependency pruning and mandatory Base fallback; startup/content-loss and effective-registry evidence. | Current Host/Evidence owners expose no custom-token assignment fields and Indicator admits the nine Base-preserved types. The normative scope argument is explicit; future assignment owners must adopt token-preserving degradation before expanding that boundary. |
| C5 | Immutable tuples, envelopes, roots, sets, first-success provenance and operation-idempotent attestations; collision/replay/collection evidence. | Database/recovery, fast/check and complete release regression passed. |
| D1 | Versioned event IDs with rightmost parsing and 266/192/184-byte profile bounds; longest identities pass actual Get/lookup. | All sixteen profiles pass the actual consumer boundary and complete release regression. |
| D2 | One evaluator and transactional usage port across Indicators, observations, Network Flow, import and revision mutations; nine families and explicit identity preflight. | Caller and functional browser regression passed. |
| D3 | Five consumer operations with authored closed request/result schemas, shared validation, staged precedence and confidential 900-second cursors. | Closed result unions, downstream callers and complete release regression passed. |
| E1 | Three original nonempty release-bound Base packs, sixteen implemented profiles, all deployment-mode reconciliation and imported-selection revalidation. | Distribution inventory reviewed: project-owned Base and synthetic optional fixtures; no external corpus is redistributed. |
| E2 | Admission-pinned snapshots and frozen render identities, incident-bundle v5, destination trust, exact inactive reuse and retained source artifacts. | Reporting/portability regression passed; exact companion content revisions recorded; production adoption unclaimed. |
| E3 | Historical-time recovery and retained-reference inventory; cross-process backup/collection serialization and exact reimport protection. | Recovery/process regression passed; exact companion content revisions recorded; production adoption unclaimed. |
| E4 | HTTP/Job/operator coordinator, typed metadata/errors, administration details and common audit/telemetry publication. | Client/type/import-boundary/unit/functional/a11y checks, reviewed visual refresh, two ordinary visual comparisons and final release passed. |
| F1 | Obligation and per-limit evidence indexes, 163 manifests/27 families, independent vectors, explicit unreachability arguments and disposable reset rehearsal. | Independent vectors, limit arguments, pre-production scope disposition and complete final validation are recorded. |

## Final coordinated verification

All run roots below are under `.cartulary/test-results/`. Each retains `run-summary.json`; the full release also retains its source manifest, unit results, target summaries and visual reconciliation. Failed and canceled diagnostic runs remain recorded in the historical checkpoints.

| Command | Final result | Retained run |
| --- | --- | --- |
| `make format` | Passed 2/2 after the last browser fixture preparation change. | `20261004T063334Z-p30649` |
| `make generate` | Passed after the last authored source change. | `20261004T063341Z-p35615` |
| `make agent-finalize` | Passed 1/1 before the complete release rerun. | `20261004T063610Z-p76079` |
| `make test-fast` | Passed 737/737 on the final production implementation; subsequent changes affected browser fixtures, goldens and documentation. | `20261004T044530Z-p34338` |
| `make check` | Passed 990/990 on the final production implementation. | `20261004T045139Z-p26860` |
| `make browser-e2e-visual` | Passed 12/12; first ordinary comparison after reviewed golden promotion. | `20261004T055013Z-p41727` |
| `make release-check` | Passed 1259/1259 in 2148.180 seconds; zero failed, skipped or canceled units. Includes the second ordinary visual comparison and final browser fixture changes. | `20261004T063630Z-p80139` |
| `make lint-markdown` | Passed after the completed handoff edits and again after the result-only ledger/owner-inventory additions. | `20261004T071627Z-p84809`; follow-up `20261004T071911Z-p87560` |
| `git diff --check` | Passed after the handoff edits. | Local whitespace check; no harness run root. |

The release source manifest identifies uncommitted source digest `sha256:a0078d33960af5195218b0d56d7e856c5fe739e7034126596b2e24330efc2785`, based on commit `78a4effc23aac4b46ca414e11b91e17892f790c7`. This is the tested working tree, not a new commit. Final handoff edits are documentation only. Both ordinary visual runs use identical renderer identity and golden-manifest SHA-256 `8fbbe83aab213faba870618041d0609d31d5d3a1d4563f37fa6025912eca50d9`; both reconcile all 255 active captures and 29 registered fixtures with zero orphan, missing, ambiguous or unresolved mappings.

Retained-run maintenance was skipped because `RESULTS_DIR` is unset. No mandatory execution unit was skipped in the successful release run. No real deployment migration/reset, commit, push or production conformance publication was performed.

## Phase exits and handoff

| Phase | Completed pre-production exit |
| --- | --- |
| 0 — Contracts | Defaults, omission/null semantics, checks, transitions and owner boundaries reconciled; exact companion content revisions and typed projections recorded. Formal owner adoption is not inferred. |
| 1 — Persistence | Appended migrations 46–63, immutable constraints and fail-without-mutation preflight verified. Explicit reset rehearsed twice against a disposable database and cleaned up. |
| 2 — Verification and operations | Live canonical verifier, retained trust, frozen admission and atomic finalization validated with adversarial, timing, concurrency and commit-fault evidence. |
| 3 — Base and consumers | All deployment branches establish Base; five consumer operations, sixteen profiles and nine algorithm families execute through owner interfaces. Current custom-token reachability limits are documented. |
| 4 — Integration | Exact Reporting binding, format-5 portability, historical recovery, retention/collection, operator administration and UI pass focused and complete release validation. |
| 5 — Evidence | Retired production paths removed; projections regenerated; obligation/limit evidence, licensing inventory, focused implementer review and operational handoff recorded. Recreation/interchangeability reports are outside the owner's pre-production scope. |

Use the [owner/version inventory](reference-pack-owner-inventory.md) for changed contracts and retired behavior, and the [development cutover procedure](../guides/reference-pack-development-cutover.md) for preflight, explicit reset, backup and recovery. Historical backups require their matching historical application; reset does not convert them. No unresolved implementation finding is carried into this pre-production handoff. Production adoption and deployment compliance remain separately unclaimed.

## Files and ownership

- `docs/reference-pack-subsystem-nlspec.md`: draft boundary closure, ordered checks, state/freshness/portability decisions, acceptance amendments. This document has not been adopted.
- `contracts/reference-packs/`: authored schemas, profiles, checks, limits, pinned SPDX identifiers, and independent vectors. The generated Go projection is produced through the existing contract-family generator.
- `internal/modules/reference_data/internal/packformat/`: pure archive, trust, manifest, license, identity, profile, normalization, indexing and diagnostic logic.
- `internal/modules/reference_data/internal/packstate/`: pure eligibility, frozen cohort, transition, fallback and revision-comparison logic. These are proposals, not durable publication.
- `internal/platform/canonicaljson/`: strict original-byte JSON admission before canonicalization; duplicate keys and malformed Unicode are rejected without echoing input.
- `internal/modules/reference_data/storage_port.go` and `internal/app/referenceassembly/verification_storage.go`: streaming, root-free verification storage; fresh private workspaces with confined cleanup.
- `internal/platform/rootedfs/`: random-access regular readers and identity-guarded empty-directory removal, with no recursive deletion.
- Contract-family, generation-scratch and test-routing authored inputs are updated; generated outputs are refreshed using Make.

## Historical verification record — foundations (2026-10-02)

Read-only owner discovery completed for `module.reference_data` and `platform.rootedfs`. Focused verification passed for strict JSON, archives, trust, manifests/licenses/set identity, all declared profiles, indicators, diagnostics, pure transitions, workspace isolation/cleanup, and rooted filesystem containment.

Selected successful run roots under `.cartulary/test-results/`:

| Run | Evidence |
| --- | --- |
| `20261002T194459Z-p55946` | Strict canonical JSON admission and existing RFC 8785 vectors. |
| `20261002T202140Z-p42489` | Storage workspace row passed; the combined archive row failed due to the ratio fixture described below. This is not a successful whole run. |
| `20261002T202227Z-p48543` | Rooted filesystem containment and empty-directory rejection. |
| `20261002T202325Z-p54064` | Archive, trust, manifest, indicator and profile rows after ratio-fixture correction. |
| `20261002T202550Z-p59877` | Diagnostic ordering, deterministic retention and operational-abort semantics. |
| `20261002T203151Z-p71770` | Pure verification transitions, refresh freezing, dependency fallback and revision guards. |
| `20261002T203444Z-p77971` | Trust combined-fault precedence and profile validation. |
| `20261002T203658Z-p79788` | Generated contract and topology outputs. |
| `20261002T203911Z-p84435` | Markdown lint. |

A newly authored invalid-text fixture initially used an invalid UTF-8 byte instead of the intended Unicode C1 control; it was corrected to U+0085. The compression-ratio regression fixture initially used a hard-coded ratio that did not straddle its intended two-member boundary; the test now chooses the effective ratio from a fixed 150,000-byte threshold and the complete fixture size. Production limits were not enlarged. Both failures were related to newly authored tests and passed after correction.

`agent-finalize` initially failed because test-routing edits made the topology inputs stale. After regeneration, `json-shape-check` exposed the separate contract-family allowlist; the authored validator now includes `reference-packs`. That validator edit itself requires topology regeneration before finalization. Regeneration and finalization subsequently passed; see the final checkpoint below.

Retained-run maintenance is skipped because `RESULTS_DIR` is unset. No full service-backed, frontend, migration, `check`, `release-check`, or security-review disposition has been established for the coordinated cutover. Those mandatory release obligations are **unfulfilled**, not waived or represented by the focused foundation tests.

## Completed pre-production acceptance sequence

1. Reviewed and transactionally promoted thirteen explained Reference Pack goldens. Preserved the Evidence golden after repairing its capture anchor.
2. Passed two fresh ordinary visual comparisons, including the complete release run after `agent-finalize`; exact results are recorded above.
3. Recorded obligation, version, migration/reset, licensing and focused security-review dispositions. Production conformance and deployment-specific compliance approval remain distinct; no recreation or interchangeability report is required for this pre-production work.

## Historical foundation checkpoint verification

- `make generate`: passed, `.cartulary/test-results/20261002T204057Z-p21931`.
- `make agent-finalize`: passed, `.cartulary/test-results/20261002T204108Z-p27395`; retained-run maintenance skipped because `RESULTS_DIR` was unset.
- `make test-slice OWNER=module.reference_data`: passed, 23/23 execution units, `.cartulary/test-results/20261002T204009Z-p87002`.
- `make test-slice OWNER=platform.rootedfs`: passed, 2/2 execution units, `.cartulary/test-results/20261002T204120Z-p28688`.
- `make test-catalog-check`: passed.
- `git diff --check`: passed before the final documentation update.

These checks validate the current foundation changes and regression boundary. They do not establish that the unimplemented cutover, integrations, or adoption gates are complete.

The broad `make test-fast` regression passed all 736 execution units at `.cartulary/test-results/20261002T204329Z-p32720`. Subsequent archive review added a draft fixed maximum of 10,000 empty directory markers, independent of regular-file counts. This closes collision-index memory growth through logically empty directories without changing logical identity or compression-ratio accounting. Complete TAR fixtures exercise equality and one-over. The subsequent focused archive/profile check covers that follow-up.

The directory-bound follow-up passed generation at `.cartulary/test-results/20261002T205101Z-p27971`, the focused archive/profile slice at `.cartulary/test-results/20261002T205116Z-p35624`, and `agent-finalize` at `.cartulary/test-results/20261002T205121Z-p36073`. `git diff --check` also passed. The 736-unit fast run preceded this follow-up; the focused slice covers its changed production code. The post-ledger Markdown lint passed at `.cartulary/test-results/20261002T204330Z-p32923`; a final lint includes this evidence update.

## Reachability arguments requiring separate guard evidence

- **Pack-set member ceiling:** the current closed catalog contains 16 pack keys and a set permits at most one version per key. A valid current set therefore cannot contain 64 members. The 64-member ceiling must remain a future-compatible structural guard; an accepted current-profile 64-member fixture would contradict key uniqueness or the closed catalog. Use complete fixtures through the reachable 16-key catalog and a separate cardinality-guard test at 64/65. The independent complete 16-key vector and isolated 64/65 guard pass at `20261003T015302Z-p88217`; all 16 profiles pass live consumer integration at `20261003T015842Z-p15435`.
- **Framework relationship ceiling:** every canonical relationship contains a 68-byte `relationship_id` (`rpr_` plus 64 hexadecimal digits), even before JSON framing or any other required field. Five million relationships therefore require more than 340,000,000 bytes, exceeding the 268,435,456-byte payload-file ceiling. A complete valid fixture cannot reach five million relationships. Verify the isolated count guard at 5,000,000/5,000,001 and use complete payload fixtures for reachable file-size boundaries. The isolated count guard passes at `20261003T015302Z-p88217`. A complete uncompressed archive now reaches the 268,435,456-byte file ceiling using a streamed additional notice alongside the valid Base manifest, original license and registry payload; manifest inventory, profile content, exact notice digest and streaming notice admission are checked. A complete one-byte-over archive fails before destination writes. This passes in the Reference Data slice at `20261003T210850Z-p53778`; no production limits were enlarged.

Final `make lint-markdown` passed at `.cartulary/test-results/20261002T205247Z-p40364`. The final whitespace check also passed. No migrations, resets, commits, pushes, or release claims were performed.


## Persistence, Base and consumer implementation checkpoint

This checkpoint supersedes the earlier foundation-only status and its claim that no migration was authored. It does not close any complete A–F workstream.

- Appended migration `00046_reference_pack_immutable_state.sql` introduces candidates, attempts, immutable versions/envelopes/root history/sets/provenance, retained objects, indexes, revisions, operations, pins and release bindings. The application preflight reports incompatible categories before migration; the migration repeats the check under locks. Neither path converts or deletes retained data.
- Added bounded stream staging/publication and a canonical verification engine over confined per-attempt workspaces. The engine is **not yet the live HTTP/Jobs verifier**, and its phase integration and full diagnostic enumeration remain incomplete.
- Base reconciliation now runs before readiness in every deployment mode. Three nonempty compiled registries carry manifests, notices and release digests. Claimed startup admits configured offline bootstrap roots and persists exact root bytes; retained roots cannot be replaced through a configuration edit. Asset provenance/licensing disposition and complete release-binding review remain open.
- Added exact-set consumer operations, immutable first-success provenance, transactional current-set pin acquisition, and authenticated confidential lookup cursors. Request admission is being hardened before downstream owners are connected. Indicators, observations, Network Flow and Reporting still require migration to these interfaces.
- Configuration now owns `reference_packs.clock_trusted` (default false), the required claimed-profile bootstrap path, and the 1,800-second execution budget bounded by 60–86,400. The Extensions configuration contract advances to v4 with an explicit configuration namespace; the Reference Pack configuration contract advances to major 2. Profile major and persisted lifecycle job cutover remain pending.
- Recovery state catalog advances to v2 and the changed graph restore implementation binding to v5; the graph restore algorithm remains v4. The Reference Data contribution inventories retained object references rather than a health allowlist. Historical signature/freshness restore, index reconstruction and collection serialization are still incomplete.

Current validation evidence:

| Command | Result and run root |
| --- | --- |
| `make generate` | Passed, `20261002T222018Z-p25624`; subsequent source/test edits require final regeneration. |
| `make test-slice OWNER=platform.config` | Passed 2/2 units, `20261002T222138Z-p30416`. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3 units, `20261002T222139Z-p32168`: fresh Base, pins, immutable provenance, restart reuse, read-only preflight, migration rejection without mutation. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers` | Earlier consumer checkpoint passed, `20261002T215733Z-p83854`; new request-decoding edits require a rerun. |

The broader Reference Data slice failed at `20261002T221420Z-p56807` (20/23 units) while the configuration adapter still assumed that a profile ID was its namespace. That assumption has been corrected. Legacy integration fixtures and routes still require the coordinated production cutover; no broad passing result is claimed. A concurrent service run at `20261002T221421Z-p57037` failed on a shared harness warm-stamp race and supplies no product evidence; the later sequential persistence run above passed.

The earlier 736-unit fast run predates these changes. Current `agent-finalize`, broad fast/check/release, browser, focused security review and companion-owner adoption remain outstanding. `RESULTS_DIR` is unset, so retained-run maintenance remains skipped. No reset, commit, push or conformance claim has been made.

### Publication preparation follow-up

The draft migration now also retains private index generations and immutable operation-member/repository captures. Index rows are prepared in bounded batches before publication; the public candidate points only to a complete generation bound to exact manifest/payload digests. Per-file index hashes are compared with the signed payload. The version sequence uniqueness scope is `(repository_id, pack_key, pack_release_sequence)`, with built-in sequences separate from operator repositories. A previous per-key-only projection was corrected before any real deployment migration.

Frozen admission and publication guard functions now use repository/key/current-set/usage lock order. Pending work is checked at the selected key boundary before missing-envelope eligibility. Cohort members reference immutable envelopes instead of copying all content into a large Job payload. These functions are not yet integrated with production Jobs admission or finalization; acceptance coverage and coordinator wiring remain mandatory.

Base configuration identity now includes the clock assertion and verification/archive limits. A restart with unchanged configuration should preserve its revision; a policy change must invalidate affected frozen input. Recovery catalog counts advance with the new authored tables; the catalog remains v2 within this unpublished cutover.

- Consumer JSON admission and the independent complete signed-container test passed at `20261002T225513Z-p43403`. The test verifies exact expected digests/signers, start-time expiry equality, untrusted-clock rejection and workspace ownership. Earlier failures identified mistakes in the newly authored fixture (root threshold, hint schema ID, target inventory and required source timestamp); the fixture was corrected without weakening admission.
- The shared Extensions finalizer now consults an immutable success proof after a commit-acknowledgement error. A matching final commit identity, timestamp, result and digest proves success; absence or inability to read the proof does not prove rollback. Its focused service-backed row passed at `20261002T224314Z-p8758`. This is not complete cancellation/deadline or failed-outcome reconciliation evidence.
- The index-generation persistence checkpoint passed at `20261002T223528Z-p69529`; subsequent admission/configuration/sequence changes require the next persistence run.
- Generation passed at `20261002T230013Z-p55007`; later schema and source changes still require regeneration.

The authoritative release block is unchanged: live legacy import verification, the full outcome coordinator, consumer injection, portability, restore, collection, administration projections, owner adoption and focused security disposition are incomplete.

### Coordinator execution checkpoint

The application coordinator now admits immutable operation inputs, starts attempts under the Jobs execution lock, streams verification into private index generations, and retains individual member results in `reference_pack_attempt_members`. It processes a frozen cohort one member at a time. The publication transaction merges compatible root proposals, publishes successful envelopes and failed verdicts, restores mandatory Base registries where needed, retains first-success set provenance, and finalizes the attempt and operation through Extensions. Cancellation and operational aborts use a separate mutation that preserves established health. No transport has yet been switched to this coordinator.

The migration now retains 25 new Reference Data tables. The complete recovery catalog classifies 140 authored tables, including 109 required authoritative tables. Frozen operation headers, cohort rows, terminal attempts and member results have database mutation guards. These are still unpublished changes to migration 00046, exercised only in disposable databases.

Queued imports now finish an immutable container object before admission. The operation's object reference commits with the Job and is included in recovery inventory; temporary upload staging is removed. Successful envelopes reuse that exact retained container. Uncertain admission keeps potentially referenced objects. This closes the earlier temporary-staging reference gap without inventing trust or historical provenance.

The coordinator uses the Auth-owned transaction lock for an idempotency identity before creating a Job or taking owner mutation locks. Attempt startup validates the current Jobs lease in the same transaction that records interruption of an older attempt. Frozen trust loading is bounded independently of accumulated repository history. Reverify checks every retained logical member before reading the successful container; it cannot silently reconstruct lost payload bytes.

Evidence completed before the latest input-retention and attempt-guard follow-up:

- `make generate`: passed at `20261002T234801Z-p52000`.
- `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence`: passed at `20261002T234130Z-p20427`. This covers signed coordinator publication, disabled-state preservation, failed-renewal preservation, active invalidation/fallback, immutable historical sets, leased import, terminal replay, queued expiry and cancellation. It does not establish mixed-cohort, all deadline faults, or live HTTP coverage.
- The preceding leased test failed at `20261002T233928Z-p91752` because its fixed signing-vector clock produced an already-expired PostgreSQL Job lease. Jobs now uses the real lease clock in that fixture while Reference Data uses the independent verification clock. No production lease validation was weakened.
- Shared finalizer cancellation, lost success/failure acknowledgements, proven absence and indeterminate outcome coverage passed at `20261002T231204Z-p14992` under the Extensions finalization row.
- Generation at `20261002T233057Z-p51149` failed because the updated recovery catalog's canonical artifact digest had not been propagated into the generation registry. The authored registry was corrected, and generation passed at `20261002T233321Z-p58653`.

Current mandatory gaps remain live coordinator routing, complete ordered diagnostics, all public/version projections, activation/disable/removal, consumer injection, exact reporting binding, incident portability, historical restore, collection, UI/operator integration, focused security review, companion adoption and release validation. No full-plan completion or conformance claim is made.

### Live administration and retired verifier checkpoint

- Removed the production legacy verifier, mutable store and empty built-in generator. HTTP imports, reverify and refresh now use canonical coordinator admission and Jobs execution; activation, disablement and removal use its administration operations.
- Administrative resources expose health separately from disablement, explicit nullable success metadata, complete signer IDs and current verification timestamps. Core 01 section 17.4 and the pending OpenAPI release describe the changed contract. Historical OpenAPI baselines remain unchanged.
- Administrative JSON admission rejects duplicate keys, malformed Unicode and trailing input before ordinary decoding. Requests have a 65,536-byte bound; normalized reasons have a 4,096-scalar bound. Removal requires a reason. Public diagnostics no longer echo arbitrary unknown member names or internal storage errors.
- Added removal through the existing administration command/replay controller and updated generated operation bindings. Runtime and recovery grants are explicit for the new tables; the initial full owner slice exposed missing runtime grants, which have been corrected in the unpublished migration.
- Focused canonical persistence passed 3/3 at `.cartulary/test-results/20261003T005834Z-p46872`; frontend type checks passed 2/2 at `.cartulary/test-results/20261003T005835Z-p47128`. Generation passed at `20261003T005803Z-p43734` before subsequent HTTP fixture edits. The complete Reference Data slice at `20261003T005539Z-p6721` failed 19/23 with missing runtime table grants and a retired disabled-state expectation; a full rerun remains required.
- Canonical signed HTTP fixtures and lifecycle expectations are being replaced. This work does not close browser, cross-owner consumer, audit, telemetry, security-review or adoption obligations. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.

### Exact binding, portability shape and historical recovery checkpoint

This checkpoint records implementation through 2026-10-03 02:30 UTC. The full A–F remediation remains incomplete and release remains blocked.

- Removed the whole-byte storage API. The confined streaming port is the only live Reference Pack storage path. The Extensions finalizer now classifies synchronous administration commit acknowledgement loss through a durable operation proof, including cancellation after commit.
- Reporting export/derivation/job contracts and render-bundle manifest advance to v2. Snapshot admission acquires a transactional exact-set pin and retains first-success provenance; release/render jobs retain the binding. Missing required content is the closed `required_reference_pack_unavailable` conflict. Committed admission replay precedes fresh content availability checks. Migration 00047 rejects the retired render manifest representation; migration 00046 preflight additionally inventories retained portability Jobs and historical incident-bundle exports without deleting them.
- Incident-bundle format 5 is current-only. The authored source catalog and every source owner accept format 5. The required `reference_pack_refs.v1` object has strict nested schemas, canonical sets, unique/comprehensive version rows and independent fixtures. This closes representation admission, **not** full portability: actual artifact reference discovery, embeddings, destination verification/reuse, pins and redistribution scenarios are still open.
- All 16 profiles now pass actual HTTP import/activation followed by Get/lookup through the consumer interface. All nine indicator families pass independent evaluation/idempotence vectors through that boundary. Indicators, observations and Network Flow still require injection of this interface and retirement of their private canonicalizers.
- Historical validation verifies root transitions against both predecessor/successor thresholds, retained envelope identity, signed bytes at the recorded successful time, logical inventories, Base release bindings, immutable set provenance and required indexes. Restore rebuilds indexes without changing health, disablement or provenance. Startup invokes required-state validation before readiness. Additional corruption and concurrency evidence remains open.
- Recovery now has a separate confined Reference Pack object source/restore port. It restores exact opaque references with streaming digest/size verification, rejects occupied destinations and confines disposable-target cleanup. Restore marker v3 and restore-verification artifact v3 include the Reference Pack root; the unpublished journal v4 completion retains all target binding digests for replay. The full operator acceptance run is in progress; the older process sentinel still needs its legacy fixture updated.

Selected evidence (all run roots are under `.cartulary/test-results/`):

| Command / boundary | Result |
| --- | --- |
| Extensions service slice, finalization cancellation/reconciliation row | Passed 3/3, `20261003T012931Z-p63504`. |
| Reference Data consumer/profile HTTP service rows | Passed 3/3, `20261003T015842Z-p15435`. |
| Reference Data request/manifest unit rows, including portable references and isolated limits | Passed 2/2, `20261003T015302Z-p88217`. |
| Reporting snapshot binding service row | Passed 3/3, `20261003T020439Z-p39461`; includes missing-content rejection and committed replay. |
| Canonical persistence service row with historical expiry/index rebuild and required notice loss | Passed 3/3, `20261003T021251Z-p86460`; later set-provenance validation changes need the next rerun. |
| Reference Pack confined storage row with backup/restore identity and cleanup | Passed 1/1, `20261003T022013Z-p22173`. |
| Recovery marker and verification artifact unit rows | Passed 2/2, `20261003T022610Z-p96854`; later terminal binding evidence changes need rerun. |
| Generation | Passed `20261003T022418Z-p93316`; subsequent journal/schema edits require regeneration. |

Failures retained for attribution:

- `20261003T015819Z-p14118`: generation rejected one unclassified pending OpenAPI description change. The Reporting-owned change was classified; generation passed at `20261003T020400Z-p36259`.
- `20261003T015301Z-p87991`: the new Windows lookup fixture expected one result for omitted version. The normative result includes both versions; only the assertion changed before the passing profile run.
- `20261003T014638Z-p58586`: a new preflight query named the wrong incident-bundle export table. Corrected before the passing migration/Reporting runs. Current migration schema hash is `b11048c0cac06f34b352ad3085ffb02ee5b6c2534c64aea82b12bf9d96eef282`.
- `20261003T020620Z-p59872`: incident-bundle service failures exposed a remaining Timeline format-4 switch. It now admits format 5. `20261003T022015Z-p22521` then passed 5/6 units; the remaining restart fixture reused the database with a fresh Reference Pack root. The fixture now retains that root; rerun is required.
- `20261003T022237Z-p48686`: the packaged-server recovery sentinel failed before readiness because its process fixture omitted the now-required trust bootstrap. That fixture is corrected; its old backup representation still requires migration to the current recovery path.

No real deployment migration, reset, commit or push occurred. Current broad `agent-finalize`, `test-fast`, `check`, `release-check`, browser evidence, focused security disposition, asset licensing disposition and companion-owner adoption are outstanding. The old 736-unit fast run is not evidence for the live cutover. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.


### Recovery round-trip and admission checkpoint

Implementation through 2026-10-03 03:00 UTC remains incomplete overall. The live operator and packaged-server recovery paths now have passing Reference Pack evidence:

- The unpublished migration 00046 no longer seeds an empty current-set singleton. Base establishes it in the same transaction as the first usable set. This leaves fresh migration and disposable-reset targets empty without a restore-admission exception. The current migration source hash is `2513be78d24b2419129665b9416f24bc1f831f52aa547ea08a3827d4e87ad6d1`; only disposable databases were used.
- PostgreSQL restore supplies writable columns explicitly and checks generated values against the backup after PostgreSQL recomputes them. The focused round trip exposed the attempted write to `reference_pack_lookup_keys.lookup_sha256`; no database integrity constraint was weakened. Restore now also rejects missing or additional row columns, and capture admits original JSON without duplicate-key or Unicode repair.
- The process sentinel uses current catalog-driven backup and restore, including Base objects in their separate root, historical validation, derived index rebuilding, graph/workbook reconstruction and actual server readiness. Operator terminal replay retains the root bindings already captured in journal v4.
- Root admission checks filesystem overlap across source packs, generic objects, backups and destination storage. Current journal decoding requires explicit nullable members, validates nested typed shape and rejects lossy JSON. These latest admission changes have focused unit coverage; the preceding operator/process pass predates them and requires a follow-up service run.

| Command | Result and retained run |
| --- | --- |
| `make generate` | Passed `20261003T023832Z-p68594`; later source changes require regeneration. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3, `20261003T024838Z-p8943`, including Base backup/restore, pinned provenance and readiness. Predates the final exact-column shape check. |
| `make service-backed-test-slice OWNER=module.recovery ROWS=module.recovery.process.canonical_operator_process_evidence_maps_the_imp_9808fdd4e9,module.recovery.process.the_packaged_server_serves_the_coherently_restor_6b1731590e` | Passed 9/9, `20261003T024953Z-p28590`. |
| `make service-backed-test-slice OWNER=module.incidentbundles` | Passed 6/6, `20261003T025136Z-p67889`. This validates existing format-5 scenarios, not the unfinished pack embedding/import behavior. |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.closed_typed_failure_catalog_2d89c71ab4,module.recovery.unit.exact_recovery_state_catalog_and_synthetic_goose_774b28fd25,module.recovery.unit.graph_restore_v4_contract_projection,module.recovery.unit.owner_registered_workbook_probe_and_v2_artifact_a881cb6053,module.recovery.unit.restore_target_marker_v2_validates_purpose_gener_492d40e257,module.recovery.unit.runtime_resource_lifetimes_9c37e10ad6,module.recovery.unit.vnext_parallel_capture_restore_codecs_8d6a20d34c` | Passed 3/3, `20261003T025153Z-p82166`. |
| `make test-slice OWNER=module.recovery ROWS=module.recovery.unit.restore_target_marker_v2_validates_purpose_gener_492d40e257,module.recovery.unit.graph_restore_v4_contract_projection,module.recovery.unit.closed_typed_failure_catalog_2d89c71ab4` | Passed 2/2, `20261003T025709Z-p94402`, including current journal nested admission and cross-root overlap. |

Earlier failures were change-related and fixed: `20261003T023131Z-p27602` rejected the seeded empty singleton; `20261003T023922Z-p76495` and `20261003T024614Z-p84624` exposed generated-column restore; `20261003T024236Z-p21209` used an insufficient retention duration in the new process fixture; `20261003T024458Z-p65593` caught an unused test import; `20261003T025610Z-p93520` caught an incomplete storage-binding test fixture. The later passes above supersede those specific failures. Required broad checks, security review and owner adoption are still outstanding. `RESULTS_DIR` remains unset, so retained-run maintenance remains skipped.


### Shared identity and registry transaction checkpoint

Implementation through 2026-10-03 04:20 UTC remains incomplete overall.

- Indicators now obtains canonical evaluation from Reference Data for explicit creation, observations, Network Flow participants, import, rollback and incident portability. Private type aliases and identity canonicalizers are retired. Raw display and supplied normalized text reach owner evaluation without generic line normalization. All nine algorithm families and a 6,000-character URL are covered through live HTTP.
- Source transactions acquire mandatory registry keys, current-set state and usage guards before incident or record locks. Application composition supplies the guarded transaction opener to Revisions, Imports and the Extensions cross-owner backend. Source owners increment only affected registry usage; unchanged identities and unrelated edits do not create a new usage revision. Dedicated concurrency coverage is still required beyond the current guard/rollback tests.
- Migration 00048 introduces non-null canonical values, type-specific dedupe constraints and bounded digest indexes with exact logical comparisons. The generated active-identity digest is recomputed during restore. No historical key is reinterpreted. Preflight now also rejects retained Indicator identities in intermediate migration-47 databases. The current authored migration source hash is `05bcbeb5a0b6af91ec0a8657c52d4b1a242ec85a6da21536a1f7f5f5600e5ee9`.
- Indicator portable rows and browser creation constraints advance to v2; the Network Flow find-or-create participant is v2. The source-state catalog names exact row schema IDs rather than synthesizing `.v1`. Browser value-kind choices derive from the Reference Pack algorithm catalog. Public raw-value requirements and bounds are recorded in the pending OpenAPI release.
- Administrative metadata projection is being completed with release sequence, source profile/version/time, license, redistribution and repository identity. Signers use an explicit empty array before success. Core and subsystem prose are reconciled on that omission rule. These latest projection changes still need generation and acceptance evidence.

| Command | Result and retained run |
| --- | --- |
| `make test-slice OWNER=module.indicators` | Passed 20/20, `20261003T041702Z-p50745`, after source-schema v2 correction. |
| `make service-backed-test-slice OWNER=module.indicators` | Passed 8/8, `20261003T035910Z-p6775`; predates the final portable-row v2 and browser projections. |
| `make generate` | Passed `20261003T041627Z-p47354`; latest administrative projection changes require regeneration. |
| `make frontend-typecheck` | Passed 2/2, `20261003T040451Z-p35860`; latest administrative fields require rerun. |
| `make frontend-unit` | Failed 687/691, `20261003T040642Z-p63075`. Four change-related failures: retired pack fixtures, derived Indicator defaults/client validation, and source-ownership inventory. Corrections are authored. |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.indicator_canonical_transport` | Passed 2/2, `20261003T041702Z-p50749`. |
| `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.generated_http_operation_bindings` | Passed 2/2, `20261003T041702Z-p50768`; predates the additional metadata fields. |
| `make service-backed-test-slice OWNER=module.networkflow` | Failed 15/36 units, `20261003T040642Z-p63009`. Backend behavior rows passed except a database cleanup timeout. Browser startup lacked the required trust bootstrap; its harness configuration is corrected and awaits rerun. |
| Canonical persistence service row | `20261003T041702Z-p50742` failed before product execution because concurrent owner slices raced on the shared service-image warm stamp. This supplies no product evidence; service-backed runs must be sequential. |

No real data reset or migration occurred. Full diagnostic execution, historical artifact portability, collection/backup serialization, custom-token degradation, operator import, complete public/audit/telemetry evidence, focused security review, licensing disposition and owner adoption remain open. Current finalization and broad release gates have not run. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.

### Live verification, browser and consumer integrity checkpoint

Implementation through 2026-10-03 05:40 UTC remains incomplete overall.

- Browser fixtures now produce canonical Ed25519/TUF containers, with a confined harness bootstrap and an explicit trusted-clock assertion. The real administration lifecycle passed. The duplicate administration unit fixture now uses the shared typed fixture. The closed administrative schema is checked for exact parity with the authored OpenAPI resource, including all eight source/release/license/repository fields added in the cutover.
- Mixed refresh acceptance now covers one successful renewal and one missing active member: one transaction renews the healthy envelope while preserving disablement, marks the missing member, applies Base fallback and retains a failed Job outcome. Replaying the operation creates neither another attempt nor another member attestation; historical set provenance remains unchanged.
- Verification separates manifest shape, logical identity, payload identity, content shape, content semantics and compatibility. Release-sequence and logical-version checks run before content validation; historical restore supplies exact retained identity instead of current release-sequence policy. Every content-schema pass finishes before semantic checks or index construction. Combined-fault HTTP fixtures exercise these winners. This does **not** close the complete 49-check runner, enumerable finding collection or diagnostic safety obligation.
- Consumer cursors retain subsecond timestamp precision for their exact 900-second lifetime. Missing or incomplete indexes fail explicitly. Consumer-detected missing or altered authoritative bytes now invoke coordinator invalidation with revision comparison and atomic fallback. Transactional consumers roll back the complete source transaction before independent invalidation, avoiding a publication-lock deadlock and preventing source writes from committing after content loss. Repeated detection preserves the first event. Required Base loss with no healthy fallback, complete normative attestations/audit and broader races still need closure.
- Recovery rejects the retired whole-snapshot representation before reading its artifacts. The live restore path requires the current state catalog and admitted operation identity. Catalog-driven readiness, metadata-store, tampered-binding and missing-artifact fixtures replace retired restore fixtures. The old whole-snapshot capture helpers still require retirement; this checkpoint does not claim they are gone.

All retained roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| Reference Data live administration browser row (`make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.browser.administration_real_lifecycle`) | Passed 11/11, `20261003T043822Z-p30392`. |
| Recovery browser row (`module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0`) through `make service-backed-test-slice` | Passed 11/11, `20261003T042734Z-p54690`. |
| Recovery service rows for readiness, selected-backup validation, missing artifact and durable metadata creation | Passed 4/4, `20261003T045458Z-p27107`. Exact row selectors are retained in that run's plan. |
| `make generate` | Passed `20261003T053605Z-p94987`, after preceding passes `20261003T050206Z-p47760` and `20261003T051054Z-p78107`. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.generated_open_api_and_error_registry_artifacts_0843af138f,module.reference_data.frontend.the_deployment_admin_reference_pack_panel_starts_1cf94e4f79,module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` | Passed 6/6, `20261003T051614Z-p29302`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T052050Z-p54258`; predates consumer invalidation. |
| `make test-slice OWNER=module.reference_data` | Failed 22/23, `20261003T053638Z-p3186`. New consumer invalidation acceptance passed. One retired symlink expectation was corrected; rerun required. |
| `make frontend-typecheck` | Passed 2/2, `20261003T053638Z-p3262`. |
| `make format` | Passed 2/2, `20261003T053625Z-p98194`; subsequent edits need formatting. |

Earlier change-related failures and repairs remain attributable: browser `20261003T043547Z-p95929` omitted required source time; recovery `20261003T045049Z-p3328` used an obsolete evidence binding and misplaced service selector; Reference Data `20261003T051149Z-p86498` exposed stale administrative fixtures and a combined-fault fixture that unintentionally reused TUF metadata version 1; frontend typecheck `20261003T052051Z-p54519` found an unused import after fixture replacement. Those specific defects were corrected before the corresponding passes above. Generation `20261003T044825Z-p91423` rejected selector ordering; corrected generation passed at `20261003T044955Z-p94699`. The current generation result above supersedes the ordering failure.

No deployment reset, migration, commit or push occurred. Release/security/adoption gates remain open. `RESULTS_DIR` is unset; retained-run maintenance remains skipped.

### Canonical evidence, audit and release binding checkpoint

This checkpoint covers work through 2026-10-03 06:31 UTC. Full-plan completion remains blocked.

- Replaced incomplete lifecycle event objects with the exact closed `cartulary.reference_pack_attestation.v1` shape and content-derived identity. Every required nullable member is emitted. Failed verification does not borrow previous successful signer or freshness evidence. Historical validation checks canonical bytes, identity, relational event binding and operation attribution. An independently hashed fixture exercises identity and nested schema rejection.
- Lifecycle publication now appends canonical activation/rollback, verification, disablement/removal, consumer invalidation, safety fallback, dependency invalidation, exact reimport and trust-update evidence. Rollback derives from prior successful activation or lower sequence in the same repository/key scope. First-success set provenance remains unchanged.
- Migration 00049 and the authored audit registry allow the exact Reference Pack actions. Admission, terminal attempts and semantic publication append raw/projection audit pairs in their owner transactions. Audit carries operation identity and opaque version references, with no payload, source identifier, user reason, path or raw error. Replay and repeated consumer detection do not duplicate events. Bootstrap-root import and the complete rejected-admission audit matrix remain open.
- Core 01's retired table-based lifecycle, disabled reverify rule, blanket failure behavior and open pack-key extension clauses were replaced with the coordinated contract. Adoption gates explicitly include Extensions, incident portability and recovery; exact approved revisions remain unrecorded, so the subsystem remains draft.
- Packaged Base assets now carry the exact eight-member release binding; startup and restore validate its schema, sequence, profile and digest identity. Packaging wrappers remain separate. Reconciliation checks conflicting retained bindings rather than ignoring them. The browser fixture reads the new wrapper. This is a correction of unpublished packaging, not a compatibility reader.
- A narrow operation observer is now wired into live import, verification, refresh/reverify, lifecycle, Base reconciliation and all consumer paths. The telemetry owner and typed registry declare ten exact operation names and one duration histogram. Only operation/result tokens cross the port. These latest telemetry changes require verification before this obligation can be closed.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make test-slice OWNER=module.reference_data` | Passed 23/23 at `20261003T055032Z-p58622` and `20261003T061012Z-p38528`; the latter includes live audit, fallback and rollback assertions. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest` | Passed 1/1 at `20261003T055456Z-p4980`, including the independent attestation fixture. |
| `make service-backed-test-slice OWNER=module.extensions ROWS=module.extensions.integration.job_finalization_cancellation_and_reconciliation_2c712826c2` | Passed 3/3 at `20261003T055547Z-p5606`, including the Base-only mutation finalizer and lost acknowledgement. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13 at `20261003T061253Z-p75768`, before the exact Base-binding follow-up. |
| `make test-slice OWNER=platform.audit` | Passed 4/4 at `20261003T061434Z-p10296`. |
| `make test-slice OWNER=module.reference_data` after Base binding changes | Failed 21/23 at `20261003T062021Z-p39302`; all Go checks passed, but the browser fixture still read three retired wrapper fields. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.browser.administration_real_lifecycle` after fixture repair | Passed 11/11 at `20261003T062254Z-p76628`. |
| `make generate` | Passed at `20261003T060113Z-p27294`, `20261003T061852Z-p30996` and `20261003T063104Z-p13843`. Earlier `20261003T060008Z-p26097` failed because migration 49 lacked an owner assignment; that authored catalog defect was corrected. |
| `make format` | Passed 2/2 at `20261003T061950Z-p34400`; latest telemetry changes still require formatting and verification. |

No real deployment has been reset or migrated. Security authority, licensing disposition, full portability, garbage collection, operator import, complete check precedence/findings, companion adoption and broad release evidence remain mandatory open work. `RESULTS_DIR` is unset, so retained-run maintenance remains skipped.

### Archive, operational failures and bootstrap audit checkpoint

This checkpoint covers work through 2026-10-03 06:58 UTC. Full-plan completion remains blocked.

- ZIP admission now parses one central record at a time instead of constructing an unbounded metadata inventory through `archive/zip.Reader`. It skips comments, bounds names before allocation, checks local/central metadata and intervals, and independently verifies decoded size, CRC and exact compressed-stream termination. A valid 32-member fixture with 1.9 MB of comments proves bounded metadata reads. Malformed CRC, size and compressed-trailer fixtures are rejected.
- The draft and typed limits now bound the terminating TAR zero run to 2–20 blocks, including the mandatory pair. This also applies after GZIP decompression, closing work expansion through non-payload padding. Complete TAR and GZIP fixtures exercise 2, 20 and 21 blocks. Regular-file byte accounting and logical identity remain unchanged.
- Bootstrap root installation retains a terminal attributed operation and administrative audit in the same transaction as the exact root. Reconciliation replay creates neither another operation nor another audit event. Operator/user admission rejection coverage remains open.
- Retained-content checks now distinguish definite absence/prohibited filesystem replacement from permission or device/read failures. Operational failures return unavailable to the caller without condemning established content or publishing fallback. Fault injection covers permission and mid-read failures through live consumers, and failures during archive parsing through the production verification boundary. A decoder cannot erase the underlying I/O failure into a content verdict.
- Proven import-admission replay remains successful if cleanup of its unused newly uploaded copy fails; that copy remains cleanup work. Restart/orphan collection is still incomplete.
- Migration generation and schema ownership validation now import one authored owner allocation. This removes the duplicated maps that disagreed about migrations 47–49.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.operation_telemetry` | Passed 1/1, `20261003T063621Z-p23329`. |
| `make test-slice OWNER=platform.telemetry` | Passed 2/2, `20261003T063644Z-p24623`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T063718Z-p26402`, including exact Base binding restore and mixed-refresh telemetry. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_archive_admission` | Passed 1/1, `20261003T064311Z-p72665`, including bounded metadata and TAR padding. |
| `make test-slice OWNER=module.reference_data` | Passed 23/23, `20261003T064533Z-p79107`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T064720Z-p16130`, including bootstrap-root audit replay. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2,module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` | Passed 4/4, `20261003T065547Z-p67426`, including operational read fault classification and the live lifecycle. |
| `make generate` | Passed `20261003T065453Z-p59435`. |
| `make json-shape-check` | Passed 3/3, `20261003T065656Z-p87861`. Earlier failure `20261003T064853Z-p50473` exposed the duplicated migration-owner map; `20261003T065253Z-p57979` then required topology regeneration. Both are superseded by this pass. |
| `make migration-drift` | Passed 5/5, `20261003T065708Z-p88374`. |
| `make format` | Passed 2/2, `20261003T065527Z-p62705`. |

No deployment reset, commit, push or conformance claim occurred. Portability still emits empty references in its live exporter; actual snapshot section transport, permitted pack embedding and destination publication must be implemented together. Garbage collection, local operator import, complete diagnostics, human adoption/licensing/security disposition and full release evidence remain mandatory. `RESULTS_DIR` is unset; retained-run maintenance remains skipped.

### Closed envelopes, historical retention and invalidation evidence

This checkpoint covers work through 2026-10-03 08:05 UTC. The full plan remains incomplete.

- Added the closed successful-envelope projection, including required nullable success fields, exact nested trust snapshots/proposals, canonical Base64 metadata, root-history keys, sorted complete signer sets, transitions and target binding. All production envelope reads and writes use the same strict admission boundary. Historical restore checks envelope identity against its relational row and requires exact canonical root-transition evidence. Independent signed fixtures exercise nested omission, null, type, unknown-member and binding failures. Signed TUF bytes themselves are unchanged.
- Separated immutable set retention from active-set publication. Retaining a healthy disabled member does not clear disablement or change the active-set revision. The owner retention interface can pin a named historical set and export its exact closed reference catalog; duplicate uses coalesce to one set/version row. Pin acquisition locks the same ordered pack-key guards as removal. The live Incident Bundles exporter still requires integration and must not be described as complete.
- Migration 00050 adds a publication fact to attempt members for first invalidation of established logical content. A narrow column grant and trigger permit only false-to-true publication before attempt completion; prepared verification facts, repeated publication and terminal evidence remain immutable. Refresh/reverify retain distinct payload-invalidation audit and attestation evidence with the complete resulting set. Failed import renewal and operational aborts do not assert invalidation. The current independent migration source hash is `61885889af6e2cbeb9e9d68b8f47807a720441920a8bac28772afadc753a80a5`.
- Updated stale bundle-version, Extensions Job/worker, owner API disposition and migration-evidence fixtures. The administrative audit filter now includes the twenty Reference Pack actions and operation target in both the authored OpenAPI owner and existing UI model. Its breaking parameter fingerprint is recorded in the pending release; historical baselines remain untouched.
- Corrected remaining Core 02 prose that described the retired Reporting v1 model/derivation and a legacy reader. The companion now names the current v2 exact-set binding and explicit cutover rejection. No owner adoption or security/licensing authorization is inferred from these edits.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make test-fast` before compatibility-fixture corrections | Failed 725/736, `20261003T065957Z-p96820`. |
| `make test-fast` after the first fixture corrections | Failed 734/736, `20261003T071755Z-p5303`; the two remaining failures were an obsolete incident-bundle rejection vector and the administrative audit filter model. Both are corrected. A current broad rerun remains required. |
| Incidents source-codec row through `make test-slice` | Passed 1/1, `20261003T072338Z-p87149`. |
| `make test-slice OWNER=web.application ROWS=web.application.regression.administrative_audit_exact_query` | Passed 2/2, `20261003T072449Z-p88292`. |
| Reference Data verifier row through `make test-slice` | Passed 1/1, `20261003T073639Z-p4452`, including closed-envelope mutations; subsequent metadata/linkage follow-ups are included in the persistence passes below. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3, `20261003T073852Z-p11245` (envelopes and historical validation), then 3/3, `20261003T074844Z-p38917` (historical retention, catalog export and pin/removal guard). |
| `make generate` | Passed `20261003T080230Z-p6754`, including migration 50 and its owner assignment. |
| `make service-backed-test-slice OWNER=module.reference_data` | Failed 12/13, `20261003T075849Z-p70250`: the initial migration-50 projection did not permit its new publication field through the existing immutable-preparation boundary. The narrow grant/guard correction is authored and regenerated. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.reference_pack_activation_prior_active_retention_af4f8ee94a` | Passed 3/3, `20261003T080426Z-p15191`, after that correction; includes mixed refresh, first-invalidation replay and terminal-evidence mutation rejection. |
| `make format` | Passed 2/2, `20261003T080413Z-p10334`. |

Earlier change-related failures are retained: generation at `20261003T071353Z-p93561` required the pending OpenAPI fingerprint update; `20261003T073310Z-p92187` required test-selector ordering; operator evidence at `20261003T075719Z-p68660` required the new migration-catalog golden. The current golden is independently calculated from the exact emitted JSON, the appended SQL bytes and regenerated manifest; a fresh test pass remains required. A read-only `explain-target` output piped into an early-closing pager produced EPIPE; it did not run product checks or change product state.

No real migration/reset, data conversion, commit or push occurred. Full live portability, collection/backup serialization, local operator import, complete diagnostic execution/enumeration, remaining concurrency/crash evidence and exact companion adoption remain open. Focused security authority and licensing disposition have not been obtained. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.

### Closed extension members and broad fast verification

The canonical manifest and every profile object now require `extensions={}` and reject arbitrary namespaced members. This corrects a remaining draft contradiction with the approved closed-schema decision. Only the explicitly permitted TUF `signed` extension exception remains; a mutation test proves that these bytes participate in signature verification. Existing canonical fixtures and their signatures/digests remain unchanged.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make agent-finalize` | Passed 1/1, `20261003T080834Z-p35451`; retained-run maintenance skipped because `RESULTS_DIR` is unset. |
| `make test-fast` | Passed 736/736, `20261003T080900Z-p39874`, including the corrected migration-50 golden. This run precedes the extension-schema correction. |
| `make generate` | Passed `20261003T081822Z-p24580`. |
| `make format` | Passed 2/2, `20261003T081839Z-p27520`. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_trust` | Passed 1/1 execution unit, `20261003T081853Z-p32210`, covering all three selected rows. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T081918Z-p32718`, superseding the earlier migration-50 publication failure. |

Full-plan completion remains unclaimed. The mandatory production and adoption gaps listed above remain open.

### Retained provenance and broad integration repairs

This checkpoint covers work through 2026-10-03 09:08 UTC. The full remediation remains incomplete.

- Added the closed twenty-member provenance projection and strict canonical admission for retained set anchors. Publication, consumer reads, transactional pin capture and historical restore now share that boundary. Validation binds every anchor to the exact ordered member, profile authority, verification method and successful-time expiry; operator signer sets are nonempty, unique and sorted, while built-in expiry and signers retain their distinct rules. Nested omission, null, type, unknown-member and identity mutations run through independently signed fixtures. Portable source evidence still cannot confer destination trust.
- Fixed transaction-backed Indicator import to finish and close its row stream before consulting the Reference Data evaluator. This removes a same-connection query conflict during incident import. Rollback compares canonical identity without substituting the creation default for an explicitly null retained defanged presentation. Obsolete underscore-domain fixtures were rebuilt rather than accepted through a compatibility canonicalizer.
- Required Base loss now maps to the closed consumer/reporting unavailable error even if its consequence transaction cannot find a healthy fallback. Durable health publication and readiness consequences when Base itself is lost remain open; this error correction does not close that obligation.
- Disposable application restarts now preserve the Reference Pack storage root along with their database. Tests still create independent credential roots. Bootstrap audit assertions select their own action because Base/trust reconciliation legitimately emits additional audit records.
- Completed migration 46's explicit foreign-key names, supporting indexes, composite-type ACLs and append-only table privilege projections. The finalizer's one-way invalidation field has a separate narrow column grant. Migrations 46–50 remain the appended, undeployed cutover sequence; migrations 1–45 were not rewritten. Reversible-catalog tests now cover the pre-cutover sequence through 45, and independent tests prove attempted downgrade of 46, 47 and 48 fails without changing schema or version state. No real deployment was migrated or reset.
- The independently calculated migration source digest is `cc3746b50b09108486aca979383a1b7b41fb9680c89504878e3475bad4964fac`. Migration projections and exact CLI evidence fixtures were regenerated/checked against the authored SQL bytes. Removed unused legacy helpers identified by static analysis.
- Updated the exact consumer import allowlists after reviewing each boundary, preserving owner table and storage confinement. Added Reference Pack duration inventory to the telemetry checker and aligned the two new public error classifications with the telemetry owner and projection.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make agent-finalize` | Passed 1/1, `20261003T082117Z-p68261`; `RESULTS_DIR` unset. |
| `make check` before the integration repairs | Failed 973/988, `20261003T082129Z-p70212`. Focused replacement evidence follows; a fresh broad run is still required. |
| Database Migrations recurrence and preflight rows through `make test-slice` | Passed 3/3, `20261003T084703Z-p36584`. |
| Incident Bundles affected round-trip/atomicity rows through `make test-slice` | Passed 3/3, `20261003T084811Z-p56439`. |
| Reporting pinned-content row through `make test-slice` | Passed 3/3, `20261003T085013Z-p96716`. |
| Imports crash-recovery row through `make test-slice` | Passed 3/3, `20261003T085014Z-p96938`. |
| Revisions affected rollback/restart rows through `make test-slice` | Passed 3/3, `20261003T085251Z-p40397`. |
| Application/serverprocess six affected startup and restart rows through `make test-slice` | Passed 9/9, `20261003T085250Z-p40169`. |
| `make test-slice OWNER=platform.postgres ROWS=platform.postgres.integration.purpose_identity_roles_and_acl` | Passed 3/3, `20261003T090200Z-p42991`. |
| Reference Data verifier/provenance row through `make test-slice` | Passed 1/1, `20261003T090159Z-p42763`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T090249Z-p81582`. |
| Evidence attach/failure-mapping row through `make test-slice` | Passed 3/3, `20261003T090320Z-p13868`. |
| `make test-slice OWNER=app.operator ROWS=app.operator.unit.migration_evidence_cli_transport_and_redaction_906768574c` | Passed 1/1, `20261003T090320Z-p13874`. |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.integration.ui_review_seeded_network_flow_claimed` | Passed 3/3, `20261003T090355Z-p31481`; the earlier readiness timeout did not reproduce. No timeout was enlarged. |
| `make test-slice OWNER=module.records ROWS=module.records.architecture.boundary` | Passed 3/3, `20261003T090750Z-p92377`. |
| `make otel-conformance` | Passed 6/6, `20261003T090750Z-p92271`. |
| `make generate` | Passed, `20261003T090103Z-p34509`. |
| `make format` | Passed 2/2, `20261003T090146Z-p37812`. |

The broad testutil execution unit has not yet been independently re-proved. Earlier failed generation, lint, operator-golden, privilege, boundary and telemetry runs remain retained; the focused results above supersede their corrected assertions, not the complete broad gate. Full release validation, owner adoption, security review and licensing disposition remain mandatory.

Operator import has an additional explicit Core/Jobs dependency: RP-REQ-182 requires an admitted background Job, while the existing Jobs resource, persistence and route-scoped identity require a human `submitted_by_user_id`. Implementing this requires an adopted, typed local-operator attribution variant across Jobs, Extensions finalization, recovery and public Job projection. Fabricating a user UUID or introducing a synchronous second verifier would violate the approved contracts. This dependency remains open along with live snapshot/pack portability, complete diagnostic execution and enumeration, collection/backup serialization, and the remaining concurrency/crash scenarios.

### Verification precedence and relevant usage guards

This checkpoint covers work through 2026-10-03 09:44 UTC. It does not close the full plan.

- Manifest attribution no longer selects an error before trust selection. Missing or oversized hints use `bundle_shape`; unknown repositories precede TUF metadata admission. The bounded member reader now preserves each member class's declared byte-limit error instead of mapping manifest and hint limits to archive structure.
- Live TUF admission validates present descriptor types, numeric domains, digest encoding and closed nested members before signature checks. Missing role files use the metadata-schema verdict. Missing integrity fields are checked across all roles and targets before link, expiry or inventory mismatches. Root filename errors no longer preempt metadata schema or canonical-byte errors. Cartulary target repository binding is checked with metadata links; exact manifest binding remains at its later payload check. Independent signed-vector mutations and isolated at-limit byte guards cover these changes. Complete issue enumeration, bounded deduplication and public validation summaries remain open.
- Registry-usage captures now represent actual read dependencies. Normal activation retains and checks usage revisions; verification and safety fallback preserve assignments and omit that dependency. Integration evidence proves a selected-key assignment change does not abort refresh, while a captured activation usage change still rejects as stale. Other key, trust, configuration and set guards remain in force.
- Added focused routing for the HTTP test harness's restart scenario. Its repeated runtimes now retain pack storage at the parent test's lifetime while reusing their database.

All run roots below are relative to `.cartulary/test-results/`:

| Command / boundary | Result |
| --- | --- |
| `make agent-finalize` | Passed 1/1, `20261003T091702Z-p97020`; retained-run maintenance skipped because `RESULTS_DIR` is unset. |
| `make check` before this checkpoint's repairs | Failed 987/988, `20261003T091724Z-p1428`. The only failure was `raw_go:backend-integration-testutil`; its aggregate log tail omitted the original failing assertion. The repeated HTTP harness storage lifetime defect was identified by source review and repaired, with focused evidence below. A new broad pass remains required. |
| `make generate` | Passed, `20261003T093633Z-p41868`. |
| `make format` | Passed 2/2, `20261003T093644Z-p44998` and `20261003T094046Z-p71482`. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_trust,module.reference_data.unit.reference_pack_verifier_accepts_disconnected_loc_9a0b2298c2` | Passed 2/2, `20261003T093716Z-p49820`. |
| `make service-backed-test-slice OWNER=app.server ROWS=app.server.integration.http_harness_restart_storage` | Passed 3/3, `20261003T093716Z-p49828`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T094151Z-p76852`, including relevant-usage capture and live verification changes. |
| `make lint-go` | Passed in the same command session as the Markdown check; this target emitted no retained summary location. |
| `make lint-markdown` | Passed, `20261003T094212Z-p8022`, summary `adhoc/lint-markdown/tool-run-summary.json`. |

Read-only routing discovery also confirmed that `backend-integration-testutil` is an execution unit rather than a public Make target; asking `explain-target` for that name failed with `usage_error` and ran no tests. The public owner slice above now supplies narrow evidence for its application-restart boundary.

### Retained verification order and broad gate findings

This checkpoint begins with the broad verification performed through 2026-10-03 10:11 UTC. Full remediation and release adoption remain incomplete.

- `make agent-finalize` passed 1/1 at `20261003T094549Z-p30254`; retained-run maintenance was skipped because `RESULTS_DIR` is unset.
- `make test-fast` passed 736/736 at `20261003T094610Z-p34537`.
- `make check` failed with 988/989 units passed at `20261003T095132Z-p17119`. The prior raw HTTP harness restart failure is resolved. The remaining failure was `harness.browser.boundary_support.ui_review_lifecycle`: its cancellation test expected a browser action to start within two seconds of spawning Make. It observed `ready` instead of `busy`. The test now synchronizes on the borrowed server receiving the actual navigation before measuring the unchanged stop deadline.
- `make release-check` at `20261003T100105Z-p73199` was interrupted after detecting the change-related `module.artifacts.linked_notes.note_association_resource` failure. The retained Links round-trip fixture still named bundle format 4; it now names format 5. The runner emitted `status=cancelled`, with 954/1258 units passed before interruption. This is not successful release evidence; unexecuted and canceled units remain required.

The next source revision separates retained presence, length and hash checks at their declared verification ranks. Reverify and refresh cannot select a retained hash failure before an earlier trust error. Consumer integrity checks use the same presence/length/hash order. Retained length failures now use `target_length_mismatch/member_length` and participate in atomic first-invalidation evidence. Combined-fault tests cover missing retained bytes, altered bytes, trust expiry, and the absence of indexing after rejection.

Root-update metadata now admits every required nested object and field type before canonical-byte and rotation checks. The signed-field extension exception does not permit arbitrary nested key, role or binding members. Tests mutate omissions, nulls, types and unknown members while also supplying later canonical and filename faults. Cryptographic algorithm policy, derived key identities, role-key references and threshold sufficiency remain separate trust/rotation checks.

The local operator dependency is now an explicit `RP-GATE-016` promotion condition. No Job attribution workaround, synthetic user identity, local synchronous verifier or conformance claim has been introduced.

`make format` passed 2/2 at `20261003T101218Z-p69734`; `make generate` passed at `20261003T101225Z-p74525`. The focused canonical-consumer and trust rows passed 2/2 at `20261003T101248Z-p77595`. The Reference Data service slice passed 13/13 at `20261003T101248Z-p77617`; the Artifacts Note association row passed 3/3 at `20261003T101248Z-p77634`. `make harness-ui-review-lifecycle` exited successfully but emitted no retained result location. A later broad gate must still cover its aggregate scheduling behavior.

### Closed preparation and destination catalog retention

This checkpoint records continuing implementation on 2026-10-03. All mandatory completion gates remain in force.

- Durable prepared results now use a closed canonical schema, strict manifest/envelope admission, exact complete ordered inventory and internally consistent object identities. Operator publication accepts only persisted objects owned by the current operation, or the exact captured envelope's container. It cannot manufacture an object from a descriptor. The completed index generation must belong to the same operation. Tests cover nested malformed preparations and attempted publication of absent or unrelated stored objects.
- The preparation ceiling is 71,303,168 bytes. A complete valid at-ceiling fixture is unreachable under the nested limits: the 67,108,864-byte envelope plus a 1,398,104-character padded manifest and at most 68 descriptors is below 70,608,000 bytes even allowing six-byte JSON escaping for every bounded path/reference byte and conservative framing. The isolated guard is tested at 71,303,168/71,303,169; production limits were not raised. Schema references resolve only to packaged relative leaf projections, reject external references and sibling keywords, and detect cycles.
- Migration 00051 adds immutable source reference catalogs and destination-envelope version pins. A catalog may retain absent optional versions without fabricating trusted versions. Exact local collisions fail; available members receive individual pins, and complete sets receive immutable set retention and set pins. The live Incident Bundles module now requires this owner interface for import and export. Publication shares the incident final transaction, uses ordered key guards, preserves activation/disablement, and replays committed attribution before new revision checks.
- Recovery contributions include both new tables, advancing this unpublished cutover's catalog to 142 authored tables and 111 required tables. Required-state validation includes version pins and validates catalog, result, operation and pin linkage. Core 01 and the draft Reference Pack owner describe the retained evidence and its closed typed input/result projections. This is not owner adoption.
- Native Reporting portability, new embedded-pack verification, required-artifact resolution, collection, operator attribution, complete diagnostic enumeration, security review and licensing disposition remain incomplete. The new catalog boundary must not be presented as completion of E2 or E3.

Evidence before the new portability integration:

| Command / boundary | Result |
| --- | --- |
| `make format` / `make generate` | Passed at `20261003T102147Z-p41598` / `20261003T102153Z-p46439`. |
| Canonical consumer and trust rows | Passed 2/2 at `20261003T102226Z-p49476`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13 at `20261003T102226Z-p49503`, including prepared-object ownership checks. |
| `make json-shape-check` | Passed 3/3 at `20261003T102226Z-p49431`; `make lint-go` also passed without emitting a result location. |
| `make format` / `make generate` | Passed at `20261003T102832Z-p6253` / `20261003T102839Z-p11087`. |
| Canonical consumer and trust rows, including isolated prepared bound and schema-reference confinement | Passed 2/2 at `20261003T103558Z-p15998`. An earlier invocation used abbreviated row names and failed before execution with `usage_error`; the corrected command used complete owner row IDs. |

During portability implementation, generation failed at `20261003T104914Z-p28214` and `20261003T105030Z-p30773` because dependent authored recovery binding/result fixtures still carried the prior catalog or binding digest. Those inputs were updated together; generation then passed at `20261003T105118Z-p32726`, `20261003T105356Z-p41581` and `20261003T105457Z-p49826`. Format passed 2/2 at `20261003T105450Z-p45020`. Focused portability and persistence acceptance is in progress; no pass is claimed until its result is retained.

### Portability retention acceptance and diagnostic execution

The full remediation remains incomplete. These checkpoints do not close the portability, diagnostic, security-review, or adoption gates.

The imported-catalog boundary passed its owner and HTTP tests. The exact unavailable tuples survive import and re-export without activation changes. Migration 51 now rejects older imported incidents whose source reference catalogs were never retained; category counts are reported before mutation. Portable catalogs have explicit inclusive bounds of 64 MiB, 1,024 sets and 16,384 versions. A complete independently hashed 16,801,852-byte catalog reaches both collection ceilings. A valid empty catalog with admitted surrounding JSON whitespace reaches the raw-byte ceiling; the next byte is rejected before decoding.

| Command / boundary | Result |
| --- | --- |
| `make test-slice OWNER=module.incidentbundles` | Passed 8/8, `20261003T105530Z-p52800`. |
| Reference Data canonical-persistence service row | Initial compile failure at `20261003T105651Z-p94577` used an unnecessarily broad database interface. After narrowing it to the read it needs, passed 3/3 at `20261003T105829Z-p19178`. |
| Incident Bundles HTTP round-trip service row | Initial run `20261003T105531Z-p53024` failed before product execution because concurrent warm-image setup shared one temporary stamp name. The readiness helper now uses an owned `mktemp` file and atomic replacement. The row passed 3/3 at `20261003T105932Z-p38917`. |
| `make format` / `make generate` | Passed at `20261003T110800Z-p61341` / `20261003T110807Z-p66220`. |
| Canonical manifest unit row, including complete portable bounds | Passed 1/1, `20261003T110845Z-p69292`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T110851Z-p70572`, including catalog/pin linkage and migration preflight. |
| `make test-slice OWNER=module.recovery` | Passed 24/24, `20261003T110846Z-p69521`. |
| `make lint-shell` | Passed 4/4, `20261003T111021Z-p52379`. |
| Operator migration-evidence transport row | Failed `20261003T111015Z-p51863` on the old source-catalog digest. Exact SQL/manifest hashes, ledger, binding and findings were independently checked before updating the golden; passed 1/1 at `20261003T112653Z-p67226`. This pass predates migration 52. |

The diagnostic collector no longer retains an unbounded issue map or receives a complete finding slice. It streams fixed 4,096-digest chunks into confined runs, merges with two readers, deletes obsolete runs immediately and retains only the first 1,000 issue objects. Safe path/member names and detail tokens are checked against typed projections. Unknown input names use the fixed sentinel. Tests include 40,002 unique findings with repeats across spill runs, reverse traversal, hostile fields, caller mutation, cancellation and scratch faults. Initial collector/trust selection passed 1/1 at `20261003T112653Z-p67220` after format/generation at `20261003T112618Z-p59524` / `20261003T112626Z-p64135`.

Live content-schema verification now uses the same schema evaluator as admission and enumerates every safely bounded row. Migration 52 retains canonical failed summaries and their content-derived references. Committed Jobs and attestations bind those references; the new authenticated read operation exposes the closed object. Private preparation and operationally aborted attempts remain unreadable. Historical validation checks issue identity, check mapping, ordering, counts and attestation linkage. Missing old findings cannot be reconstructed: both read-only and transactional cutover guards reject incompatible failed-attempt history.

The 16 MiB retained-summary ceiling is conservative and unreachable for a valid 1,000-issue object: admitted paths need at most twice their 4,096-byte bound after escaping, admitted IDs at most twice 512 bytes, and fixed registry tokens, bounded related identities and framing keep each complete issue below 12 KiB. This bound does not authorize larger issue counts. Remaining verification phases still produce aggregate root findings, and the full registry program is not yet the live orchestration authority. A3 remains open.

Format passed at `20261003T114317Z-p76707`. Generation at `20261003T114325Z-p81541` required four additive summary-operation/schema entries in the pending OpenAPI release; generation then passed at `20261003T114459Z-p83047`. The first owner slice at `20261003T114601Z-p91202` passed 20/23 execution units but failed compilation of the root package because the new HTTP test redeclared an existing response variable. That test is corrected; a fresh pass is pending. No release gate is claimed from these focused runs. No real migration, reset, commit, push or owner adoption occurred. `RESULTS_DIR` remains unset, so retained-run maintenance remains skipped.


### Diagnostic administration and local operator execution

The full remediation and mandatory release/adoption gates remain incomplete. The diagnostic administration UI now fetches only authenticated content-derived summary references, checks the summary against its terminal Job, and keeps diagnostic read retries separate from mutation replay. Concealment, dismissal, stale responses and authority loss fence protected results. Focus stays on the existing action control. This work used the [Cartulary UI refactor skill](/mnt/c/Users/jochi/.codex/skills/cartulary-ui-ux-refactor/SKILL.md), specifically its existing-controller, local-feedback, retry and focus rules; no redesign or dependency was introduced. Rendered browser evidence for this new panel is still required.

The collector now hashes scratch runs and verifies their lengths and record counts during bounded merge. The canonical-persistence test also checks that local import renewal preserves disablement. The initial assertion expected two audit records; the actual required admission, import and verification-completion sequence produces three. That assertion is corrected and awaits the current focused rerun.

| Command / boundary | Result |
| --- | --- |
| Reference Data full service slice | Passed 13/13, `20261003T115232Z-p80016`, including migration 52 and committed diagnostics. |
| Operator migration-evidence transport | Passed 1/1, `20261003T115232Z-p80017`; predates migration 53. |
| Diagnostic and UI focused rows | Initial run `20261003T120830Z-p37685` passed 4/5. A malformed edit to an existing controller assertion caused the fifth failure; corrected controller row passed 2/2 at `20261003T121025Z-p46416`. Scratch fault, transport and panel tests passed in the initial run. |
| Frontend typecheck / import boundary | Typecheck passed 2/2 at `20261003T121025Z-p46495` after fixing the same accidental argument edit; import boundary passed 2/2 at `20261003T120852Z-p40247`. Latest nullable Job submitter typecheck passed 2/2 at `20261003T122659Z-p44830`. |
| Jobs and Extensions focused owners | Passed 6/6 and 24/24 at `20261003T122541Z-p68336` / `20261003T122541Z-p68344`; these predate the latest durable operator attribution column. |
| Reference Data canonical persistence | `20261003T122541Z-p68352` passed 2/3; the only failure was the audit-count assertion above. Local admission, worker execution, null human submitter, proof and preserved disablement were reached successfully. |
| Format / generation | Passed at `20261003T124403Z-p55074` / `20261003T124409Z-p59905`. Earlier generation failures required the pending nullable-submitter OpenAPI entry and the identity-schema catalog metadata; historical baselines were not edited. |

Migration 53 adds nullable human submission and a durable local operator operation UUID. It refuses retired import attribution before schema mutation. The full `.v2` identity schema is closed and admits either a human actor or a local operation; only import v2 adopts it, while unchanged human-only operations retain their declared v1 identity. The same Extensions finalizer validates this identity and publishes Job outcomes. Local operations omit human Auth receipt reconciliation and retain the common proof/Job identity. Expiry may clear request material without erasing the local operation UUID.

The operator facade admits bounded incoming bytes through the same coordinator, then observes the deployment worker. It holds the existing serving lease, creates no listener and does not duplicate verification execution. Interruption stops observation, leaving the durable Job. Command, confinement, worker, cutover and expiry evidence is still being completed. RP-GATE-016 stays open until that evidence and owner adoption are recorded. No real migration, reset, commit or push occurred; `RESULTS_DIR` is unset.


The operator worker scenario now passed 3/3 at `20261003T124854Z-p15639`; its earlier run `20261003T124602Z-p87705` failed on a test-only obsolete `active` column assertion after both actual command outcomes succeeded. The assertion now reads current-set membership and `verified_available` health. Operator owner tests passed 12/12 at `20261003T124854Z-p15649`. The Jobs service run `20261003T124854Z-p15646` passed 4/5; its new expiry fixture needed an explicit UUID cast for a parameter shared with JSON text. The fixture is corrected and its focused rerun is pending. Canonical persistence and cutover preflight passed 3/3 at `20261003T124452Z-p63370`. Migration 53's source hash and migration-evidence output digest were independently reconstructed from the SQL/manifest bytes before updating goldens.

Native Reporting reference discovery is now connected to incident export through a narrow owner interface. It validates stored snapshot model identity and digest, then Reference Data validates the exact binding and unions it with the retained imported catalog. Duplicate facts coalesce, conflicting tuples fail, and unknown imported optional tuples are retained. This closes the empty-native-catalog production path once current integration evidence passes. It does not implement embedded new-pack verification, required embedded artifacts or optional snapshot payload transport; E2 remains incomplete. Focused native export/import, missing-binding, catalog-union and expired operator attribution evidence is in progress.


### Required Base loss, native reference export and schema diagnostics

The coordinated remediation remains incomplete. Definitive required-registry loss now commits its health and attestation even when no healthy Base fallback exists. The same transaction clears the current selection and advances its revision, preserves immutable historical sets, and makes the live readiness probe fail. System integrity detection remains admissible with an explicitly null frozen set; current-set consumers and new administrative operations fail through their declared errors. A healthy imported replacement remains current after inactive Base loss. Disabling the final healthy replacement is rejected without mutation. Core 01 and the Reference Pack draft now describe this boundary explicitly.

Native Reporting snapshot discovery and the retained imported catalog are now exported as one exact reference graph. The previously dormant Reporting boundary tests have explicit owner routing; their stale allowance list was reconciled with existing narrow Incidents admission and Graph source interfaces. Direct Reference Data table access is prohibited. The union test now checks both native and unavailable imported references instead of incorrectly expecting the unmerged source catalog. The real browser scenario covers rejected-pack diagnostic reading, transport failure, read retry and unchanged mutation count. The fixture declaration now exposes its admitted invalid-manifest option to TypeScript.

| Validation | Result location under `.cartulary/test-results/` |
| --- | --- |
| Real administration browser, including diagnostic retry | Passed 11/11, `20261003T130245Z-p80787`. |
| Jobs local-operator compaction | Passed 3/3, `20261003T125625Z-p3685`. |
| Native snapshot incident-bundle round trip | Passed 3/3, `20261003T125625Z-p3689`. |
| Initial Reporting ownership / persistence / frontend checks | Failed at `20261003T130245Z-p80761`, `20261003T130245Z-p80770`, `20261003T130245Z-p80936` on the stale boundary allowance, incorrect union assertion and missing fixture declaration described above. |
| Reporting boundary rerun | Failed at `20261003T131345Z-p55704` on remaining existing Graph source imports; complete corrected row passed 1/1 at `20261003T131830Z-p29135`. |
| Incident Bundles public API boundary | Passed 1/1, `20261003T131851Z-p43144`. |
| Canonical persistence and lifecycle, including absent current selection | Passed 4/4, `20261003T131345Z-p55720`. |
| Complete Reference Data unit/frontend owner slice | Passed 23/23, `20261003T131345Z-p55719`. |
| Frontend typecheck | Passed 2/2, `20261003T131515Z-p19696`. |
| Lifecycle plus healthy replacement / rejected unsafe disablement | Passed 3/3, `20261003T131830Z-p29156`. |
| Markdown / JSON shape | Passed at `20261003T131830Z-p29298` / `20261003T132017Z-p55135`. |
| Manifest enumeration, diagnostics and canonical consumers | Passed 2/2, `20261003T132743Z-p65498`; an earlier invocation named a nonexistent verifier row and failed routing before execution. The actual verifier tests belong to `canonical_consumers`. |

Manifest and bundle-hint schema diagnostics now share ordinary admission rules, enumerate known structural fields safely, and collect independently applicable cross-field findings before canonical-byte checks. TUF schema projections and complete metadata shape enumeration are being connected and need their next focused validation. All-check registry orchestration, remaining diagnostic phases, embedded new-pack portability, complete retention/collection races, licensing disposition, focused security review and owner adoption remain open. The existing Reporting canonical-object helper was also found to use ordinary JSON marshaling and unprefixed hashes despite its owner's distinct canonical/hash rules; do not treat native reference round-trip acceptance as Reporting canonicalization conformance. This needs an explicit companion-contract remediation before adoption.

No real migration, reset, commit, push or owner adoption occurred. `RESULTS_DIR` remains unset; retained-run maintenance remains skipped.


### TUF registry execution checkpoint

The full Reference Data owner slice passed 23/23 at `20261003T133606Z-p83983`, its service-backed slice passed 13/13 at `20261003T133607Z-p84213`, and JSON shape passed 3/3 at `20261003T133743Z-p58001`. These runs cover the typed TUF metadata schema additions, complete shape diagnostics and required-Base loss behavior.

The next implementation replaces the TUF/target verification sequence with one complete registry-keyed program used by both live and historical readers. It collects complete winning-check findings for canonical metadata, independent roles, supplied signatures and thresholds, rollback scopes, integrity descriptors, links, expiry and target inventory. Rejected rotation chains do not guess a predecessor for later thresholds. The live error adapter now rejects an ambiguous code mapping instead of choosing the first matching registry row. A malformed Go field tag in retained metadata was corrected without changing its intended JSON member.

Initial `make format` calls failed before execution with a harness `artifact_error`: the newly routed test name was not ASCII-sorted. `make explain-target TARGET=format DETAIL=summary` exposed the exact catalog error. Correcting authored routing restored format and generation, passing at `20261003T135114Z-p67520` and `20261003T135122Z-p72155`. Acceptance of this new execution program is pending. Remaining archive precedence, non-TUF diagnostic phases, portability, retention and adoption gaps are not closed by this checkpoint.


### Archive preflight and exact extraction

The registry-keyed TUF program passed its focused canonical trust/consumer/diagnostic selection, 2/2 at `20261003T135232Z-p75535`, and the full Reference Data service slice, 13/13 at `20261003T135259Z-p76823`. An earlier focused invocation named the nonexistent `canonical_diagnostics` row; it failed routing and ran no tests. The correct row is `canonical_pack_diagnostics`.

Archive admission now uses the registry range from format identification through member type. It performs a bounded framing/decompression scan without named writes, enumerates paths and collision participants in stable original-name order, then extracts only the exact admitted headers and bytes. Changes between scans abort operationally. Format precedes container size; later framing errors precede earlier path faults when both are safely observable. Resource guards and opaque prohibited bodies stop further interpretation as explicitly described by the owner. ZIP comments remain skipped with bounded reads in each pass; production byte/member/ratio limits were not raised. The derived header ceiling uses overflow-safe comparison.

`make format` and `make generate` passed at `20261003T140326Z-p27585` / `20261003T140333Z-p32219`. The full Reference Data owner slice passed 23/23 at `20261003T140344Z-p35319`, including real administration browser coverage, traversal permutations, complete collision findings, no writes from rejected names and changed-source rejection. A subsequent TAR boundary repair validates original name/prefix padding and header checksum before the standard reader or type dispatch; its acceptance is pending. This checkpoint does not close all A3 phases, crash cleanup, portability, retention, licensing, security review, adoption or release gates.


The final TAR name/prefix/checksum tests and canonical consumers passed 2/2 at `20261003T140945Z-p88740`; the service slice passed 13/13 at `20261003T140945Z-p88751`. JSON shape passed 3/3 at `20261003T140945Z-p88701`; Markdown passed at `20261003T140953Z-p94861`. Latest archive generation passed at `20261003T140900Z-p85590` after format at `20261003T140854Z-p80956`.

`make lint-go` first failed Go vet at `20261003T141104Z-p34224`: Network Flow's retained-Job test fixture still assigned a string to the now-nullable submitter. The fixture now supplies the human submitter pointer. The next lint run passed formatting and vet, then failed staticcheck at `20261003T141250Z-p58835` on the schema evaluator's obsolete unused `inBounds` helper; that helper is removed. The subsequent lint run passed formatting, vet and staticcheck at `20261003T141410Z-p61809`, `20261003T141414Z-p66436` and `20261003T141425Z-p76628`. The affected Network Flow unit row passed 1/1 at `20261003T141514Z-p78391`.

Base asset source inspection found only project-owned data: host and evidence contain the required `unknown` fallback entry, and indicator contains the nine declared types. Their source-profile digest is `243087a41dfefd3ad706b756f44f4cdc96e66554edb2aff568e6823b090645e2`. All three license declarations are `Apache-2.0`, redistribution `allowed`, with no LicenseRef bindings. Each packaged notice is exactly the repository Apache license plus the final LF required for a notice member; the repository file itself lacks that LF. This records the actual included material and transformation. It does not authorize redistributing any optional upstream dataset or close companion-owner adoption, independent review or release gates.


Retained logical-member checks now enumerate all missing, length-mismatched or hash-mismatched objects in their winning check, using ordinals over the immutable logical inventory. Storage failures still abort without a content verdict. The canonical-persistence service selection passed 3/3 at `20261003T141830Z-p85384` after format at `20261003T141823Z-p80679`. A dependency audit then found that live verification checks manifest policy but does not resolve retained dependency tuples or their reachable graph before indexing; activation checks do not satisfy this verification obligation. The next implementation closes this gap and its admission/publication revision guard.


### Dependency verification and admission closure

Live verification now resolves every exact retained dependency, checks its canonical member bytes, traverses reachable usable versions, rejects cycles by strongly connected component and checks candidate self/dependency conflicts before registry compatibility or indexing. Healthy disabled versions remain eligible. Missing, removed, failed or digest-mismatched tuples produce bounded identifier diagnostics; operational reads abort. Historical verification ignores later administrative health/removal changes while requiring exact retained identities and bytes. Current dependency state is protected by separate immutable admission read guards; it does not falsely assert pending mutation work. Admission rechecks the reachable key closure after ordered locking; preparation and publication reject relevant changed revisions.

Migration 00054 appends those read guards and fails before mutation on nonterminal verification inputs or preexisting dependent content lacking this evidence. The application preflight reports the same two categories. Fresh-state initialization and both rejection categories passed in disposable databases. Recovery includes the guard table (143 authored, 112 required); its catalog digest is `8ce98d4607faab2fbc75e99e132a0a2fd3c245af452effdf1215db679aa87682`. No real deployment was migrated. Core 02's mutable local-override sentence is replaced with the immutable complete-pack rule; no adoption approval is claimed.

Format passed at `20261003T143900Z-p26618`, generation at `20261003T143907Z-p31507`, focused pure manifest/dependency/consumer tests 2/2 at `20261003T143939Z-p34756`, and canonical-persistence plus live verification/activation service tests 4/4 at `20261003T143939Z-p34770`. Earlier generation failures at `20261003T143508Z-p16924` and `20261003T143700Z-p24296` caught stale recovery fixture bindings and exact catalog-count checks; both were updated from their authored inputs. The migration source digest is independently calculated as `3012414fce2cc3d206a6a8ada629dc9c78c3364b8ddaafc9031b5a18d84b8302`; the exact operator evidence stdout digest is `3023fe8ca0f90e49ec3d89c098a5f93a0852c3c09229d374c142a427225bf974`. Fresh migration-owner/operator/recovery evidence is pending. Full A–F completion, collection, portability, security review and owner adoption remain open.


### Inventory, semantic diagnostics and cleanup classification

The coordinated remediation remains **incomplete**. This checkpoint records production behavior and focused evidence; it does not close portability, physical collection, the complete verification program, security review, adoption or release gates.

The live payload inventory now reports every missing, extra, length-mismatched and hash-mismatched member. Retained-member and container findings combine only for the same winning check; no truncated summary is expanded into an invented full count. Target binding reports all independent mismatches. Strict JSON checks whole-document syntax before duplicate detection, preventing an early duplicate from hiding later malformed syntax. Manifest duplicate diagnostics enumerate declared paths and collapse undeclared subtrees to a fixed safe segment.

Semantic validators now share one field-predicate implementation between ordinary admission, index construction and diagnostics. The diagnostic pass visits every schema-admitted row, including independent profile fields, aliases, source references, external references, relationship identities/endpoints, replacement chains, license bindings, notices and registry policies. Unknown indicator types do not fabricate expected algorithms. A live combined-fault import retains all four independently specified semantic findings ahead of a later compatibility failure and creates no index. Source text and unknown field names are absent from diagnostics.

Review found a cleanup classification defect: joining cleanup errors with a content rejection left the rejection discoverable through `errors.As`. Verification now returns the operational cleanup failure without a publishable content verdict. Historical validation also closes its verification workspace before successful index publication. Fault tests cover extraction-workspace cleanup and a 5,000-finding diagnostic spill workspace; neither returns a content rejection.

| Command / boundary | Result and retained run root |
| --- | --- |
| Full Reference Data service slice after dependency cutover | Passed 13/13, `20261003T144333Z-p67876`. |
| Migration drift | Passed 5/5, `20261003T144508Z-p9916`. |
| Focused inventory/consumer unit rows | Passed 2/2, `20261003T145001Z-p23435`. |
| Canonical persistence and live verification service rows | Passed 4/4, `20261003T145010Z-p26424`. |
| Recovery catalog/parallel restore unit rows | Passed 3/3, `20261003T145001Z-p23427`. |
| Migration public surface and production catalog unit rows | Passed 1/1, `20261003T145001Z-p23445`. |
| Operator migration evidence transport | Passed 1/1, `20261003T145012Z-p27319`. |
| Strict JSON combined-fault unit row | Passed 1/1, `20261003T145817Z-p56253`. |
| Semantic/profile/manifest/consumer unit rows | Passed 2/2, `20261003T150839Z-p66062`. |
| Strict JSON plus semantic/duplicate/consumer rows | Passed 3/3, `20261003T151141Z-p73462`. |
| Live verification/activation service row | Passed 3/3, `20261003T151142Z-p73685`. |
| Cleanup faults plus profile/manifest/consumer unit rows | Passed 2/2, `20261003T151440Z-p99852`. |
| Latest format | Passed 2/2, `20261003T151424Z-p94898`. |

The recovery/migration failures at `20261003T144333Z-p67854` and `20261003T144333Z-p67860` were stale expected catalog counts/digests after migration 54; the focused passes above supersede those failures. Latest inventory generation passed at `20261003T144922Z-p20240`. No source projection input changed during the subsequent semantic/cleanup implementation, but fresh final generation/drift and broader verification remain required. No real deployment migration, destructive reset, commit, push or adoption approval occurred. `RESULTS_DIR` is unset; retained-run maintenance remains skipped.


### Private workspace restart recovery

Server readiness and local operator import now reconcile abandoned private verification workspaces under descriptor-confined cross-process directory leases. A parent lease serializes creation and collection; each live workspace retains an inode lease. Process death releases ownership without PID or age heuristics. Enumeration uses bounded batches, accepts only the private workspace filename vocabulary, rejects links and special files, and never descends into authoritative storage. Descriptor duplication uses atomic close-on-exec. This closes private scratch restart cleanup only; retained-object collection and retention-reference serialization remain open.

Format passed at `20261003T153038Z-p81165`; generation passed at `20261003T153113Z-p86270`. RootedFS passed 2/2 at `20261003T153150Z-p89428`, including independent capabilities, canceled waits, symlink rejection, streaming enumeration and actual killed-process lease release. Reference Data passed 23/23 at `20261003T153150Z-p89439` and its full service slice passed 13/13 at `20261003T153150Z-p89469`. Operator passed 12/12 at `20261003T153150Z-p89468`. Workspace recovery tests preserve live work, staged inputs and published artifacts and prove repeated sweeping is idempotent. Full A–F status remains incomplete.


### Complete live verification program

The live verifier now binds all 49 callbacks before reading source bytes and executes them through the typed ordered registry. Archive, trust, manifest and dependency owners supply complete deferred ranges; missing, nil, unknown and overlapping implementations fail operationally. No content-verdict selection remains in transport orchestration. Retained inventory summaries are combined only within their proven 67-object bound; truncated summaries are never expanded into invented counts. A source failure, cleanup failure or misplaced content verdict cannot escape as a publishable rejection.

Manifest admission now gives raw/escaped Unicode and the top-level-object requirement their correct encoding/JSON ranks. Missing and oversized manifests have explicit applicability rules; oversized content is not decoded. The current closed profiles exclude active roles before `disallowed_content`, whose data-only allowlist is still asserted. The typed registry names each applicability class, and the owner records why this particular failure rank is unreachable after earlier admission. Inert examples are never classified by guessed executable syntax.

Generation passed at `20261003T155019Z-p11588` after format at `20261003T155000Z-p6584`. The six selected verification/consumer owner rows coalesced to two passing Go units at `20261003T155038Z-p14551`. Go format/vet/staticcheck passed at `20261003T155038Z-p14828`, `20261003T155042Z-p19326` and `20261003T155058Z-p35614`. JSON shape passed 3/3 at `20261003T155121Z-p38239`; the full Reference Data service slice passed 13/13 at `20261003T155121Z-p38292`. Subsequent retained-trust-guard and misplaced-verdict tests passed with consumers and strict JSON, 2/2 at `20261003T155426Z-p79698`, after format at `20261003T155335Z-p74752`. The independent signed fixture proves exact execution of every rank; removing each callback rejects before storage access. A mistakenly requested `platform.canonicaljson` owner does not exist and ran no tests; its actual routing is the Reference Data strict-JSON row used in the passing selection. Further portability, collection, security review, owner adoption and final release evidence remain open.


### Exact portability requirements and migration 55

Incident-bundle v5 now admits `reference_pack_content.v1`: a closed original-byte-validated inventory of exact operator-container identities and required set/member pairs. Container paths derive exclusively from their admitted manifest digests. Unknown, duplicate, unreferenced or aliased paths fail before owner preparation. Optional unavailable tuples remain historical references; required unavailable tuples fail before owner state is published. Required pairs survive references-only re-export. Proven exact replay precedes fresh byte availability and revision checks, and replay cannot erase or change required-member classification.

Migration 00055 adds the retained-input constraint only after rejecting prior portability operations without this evidence. The read-only preflight reports `unclassified_portable_content`; the transactional migration repeats rejection and preserves both rows and schema head. The cutover guide documents that category. No deployed database was migrated or reset.

| Command / boundary | Result and retained run root |
| --- | --- |
| Generation after migration 55 | Passed, `20261003T161955Z-p16704`. |
| Reference Data manifest and consumer units | Passed 2/2, `20261003T162456Z-p21157`. |
| Incident Bundles unit slice | Passed 8/8, `20261003T162456Z-p21165`. |
| Reference Data full service slice | Passed 13/13, `20261003T162810Z-p82737`; includes required-member failure, exact replay without storage, re-export preservation and disposable migration rejection. |
| Incident Bundles service slice | Passed 6/6 execution units, `20261003T162810Z-p82750`. |

The initial manifest test at `20261003T161025Z-p98535` used an operator fixture where its new assertion expected a built-in. The fixture was corrected; the focused pass above supersedes it. Operator runs `20261003T162456Z-p21173` (11/12) and `20261003T162810Z-p82726` (0/1) failed the stale independent migration-evidence digest after adding migration 55. The expected schema hash is now `a6cc19b74ad44b89e959f12ccde266c63ae49755df8208a38a0be630bba9433c`; the reviewed output digest is `05cceebd83114d8d4fd865408f4bb91cfdf1eb98b07398bc1cc2073eb7ce5d49`. Its focused rerun passed 1/1 at `20261003T163511Z-p49719`. Migration drift passed 5/5 at `20261003T163511Z-p49673`.

Subsequent changes preserve staged incident-bundle input until authoritative terminal Job observation, retain it across interrupted attempts or uncertain outcomes, and add independent complete embedded-content boundary fixtures. The input-retention changes passed with the Incident Bundles unit slice (8/8 at `20261003T163511Z-p49706`); the independent complete content boundaries passed with Reference Data manifest/consumer rows (2/2 at `20261003T163511Z-p49697`). Source-side exact-container export was added afterward and is under focused verification. New embedded-container verification/publication, snapshot payload transport, physical collection, security review, owner adoption and complete release evidence remain open. `RESULTS_DIR` is unset, so retained-run maintenance remains skipped.


### Source-side embedded container export

Incident export now selects exact referenced operator containers through a narrow Reference Data stream. It checks immutable manifest/envelope/object linkage, honors the manifest's `allowed` redistribution classification, skips built-ins and absent optional container bytes, and hashes the stream before and during copying. Restricted/prohibited classifications remain references-only; the implementation does not infer legal permission from the SPDX expression. Generated content inventory names only actually copied containers and preserves retained required pairs. Incident Bundles bounds allocation before reading a container and checks complete archive bytes/member counts, including generated manifest and checksums, before ZIP construction.

Reference Data's full service slice passed 13/13 at `20261003T164346Z-p85825`; it exercises the independently signed container through the exporter, exact bytes/digests, refs-only behavior, short sinks, sink failure, missing/altered bytes and mutation during copying. Incident Bundles passed 8/8 at `20261003T164346Z-p85827`, including at-limit/over-limit allocation and complete-archive metadata accounting. Staticcheck passed at `20261003T164346Z-p86362`. Format passed 2/2 at `20261003T164324Z-p80954`. Subsequent review preserves staged outputs immediately after proven incident commit, before inspecting returned participant values, and preserves cancellation as an operational export outcome. Those final small changes require rerun.

This is source-side embedding only. Destination verification/publication of new embedded containers, source snapshot artifact/provenance transport, physical retention collection, complete fault/concurrency evidence, security review, companion adoption and final release checks remain mandatory and incomplete. The next embedded-import slice must freeze one durable cohort, use the existing finalizer for incident and pack publication, and retain inputs across Jobs recovery. Submitting independent Reference Pack Jobs from Incident Portability would violate the required atomic publication boundary and is not an acceptable substitute.


Final source-export follow-up validation passed: format `20261003T164826Z-p49631`, generation `20261003T164840Z-p54392`, Incident Bundles services 6/6 `20261003T164947Z-p57928`, JSON shape 3/3 `20261003T164947Z-p57860`, and Markdown lint `20261003T164947Z-p58143`. Go-format validation also passed. No final release/conformance claim follows from these focused checks.

### Destination cohort work in progress

Migration 00056 and the closed portable verification-context projection are being added for immutable destination preparation. The tables retain exact source selections, original reusable envelopes and per-container operation-member bindings; they do not publish content or alter legacy catalogs. Recovery contribution inventory includes these facts. This foundation is not yet connected to destination verification or the incident finalizer, and its migration/projection validation is pending. The A–F status remains incomplete.


### Destination foundation and terminal recovery checkpoint

The A–F remediation remains **incomplete**. Migration 00056 and portable-context generation passed at `20261003T170519Z-p7490`; canonical schema tests passed 1/1 at `20261003T171000Z-p11614`, canonical persistence/live coordinator services passed 3/3 at `20261003T171001Z-p11842`, and recovery-catalog validation passed 1/1 at `20261003T171002Z-p12087`. The operator digest failure at `20261003T171003Z-p12349` was the expected stale fingerprint after migration 56. Its reviewed replacement is `1bdbf96d2b0182a1751f487b857fce85404f5adbe82bb39c24b9927dda94680c`; the focused rerun passed 1/1 at `20261003T171909Z-p50899`. The new preparation tables still need explicit runtime/recovery privileges; that correction and live destination wiring remain open.

Jobs now has an injected transaction-scoped terminal-effects port. Production composition supplies Reference Data's participant, so ordinary finalization is a no-op after successful owner publication, while exhausted retries and inactive-profile cancellation close pending owner work atomically. The participant preserves established health, disablement, trust and activation, never publishes private preparation, and does not invent a verification attempt for a queued abort. Source and Jobs owner amendments describe the same boundary. Focused tests are running; startup repair of previously orphaned terminal work and full adoption evidence remain open. Generation passed at `20261003T171827Z-p47343` after format `20261003T171801Z-p42256`. `RESULTS_DIR` remains unset.


### Atomic terminal effects and startup repair evidence

Jobs terminal transitions now invoke the required owner participant before returning from the same transaction. Reference Data's participant closes exhausted or inactive-profile work, keeps never-started verification without an attempt timestamp, preserves established health and disablement, and rejects success without owner publication. Startup repair uses a narrow Jobs lifecycle projection that remains available after public expiry, preserves recoverable nonterminal work, and uses the existing mutation finalizer and exact owner receipt. It performs no verification or content/trust publication.

The new `attempt_result.v1` projection closes result shapes and count/outcome relationships. Migration 00057 appends the missing immutable-preparation runtime and recovery grants without rewriting migration 56. The schema fingerprint is `609576f1a8b7679e76b475a27193e4fa4ddbddd380b97008646d1fcec519af8d`. No real deployment migration or reset occurred.

| Validation | Result / retained root |
| --- | --- |
| Initial terminal-effect Reference Data services | Passed 3/3, `20261003T171906Z-p50432`. |
| Complete Jobs service slice | Passed 5/5, `20261003T171908Z-p50654`. |
| Generation after migration 57 and closed results | Passed, `20261003T172520Z-p8578`. |
| Closed Reference Data schema unit row | Passed 1/1, `20261003T172601Z-p11717`. |
| Reference Data persistence, live verification, terminal faults and startup repair | Passed 3/3, `20261003T172603Z-p11947`. |
| Production migration catalog/privilege/rollback recurrence | Passed 3/3, `20261003T172604Z-p12190`. |

The operator transport golden at `20261003T172605Z-p12482` failed only its stale migration-56 digest. The reviewed migration-57 output digest is `1a7f5a6eb7f7778be20631d75484bba689648edbff67d5595c69335a9f1ca7a9`; rerun is pending. The full A–F status and destination-cohort wiring remain incomplete. No security approval, companion adoption, conformance claim, or final release validation is implied. `RESULTS_DIR` is unset; retained-run maintenance is skipped.


### Retained lifecycle validation and completed checkpoint

Restore/readiness now validates closed frozen lifecycle inputs, canonical attempt results, stored outcome agreement, admitted cohort counts, complete published content results, and matching successful operation/attempt evidence. Terminal operations cannot hide unfinished attempts. Corruption tests change outcomes, canonical bytes, counts and frozen inputs inside rolled-back transactions and require rejection. Common Job success without terminal owner publication is also rejected in the live integration path.

Latest focused results:

| Command | Result / retained root |
| --- | --- |
| `make test-slice OWNER=module.reference_data` | Passed 23/23, `20261003T173505Z-p76194`. |
| `make service-backed-test-slice OWNER=module.reference_data` | Passed 13/13, `20261003T173506Z-p76422`. |
| `make service-backed-test-slice OWNER=module.incidentbundles` | Passed 6/6, `20261003T173507Z-p76650`. |
| `make migration-drift` | Passed 5/5, `20261003T173508Z-p76889`. |
| Operator migration transport row | Passed 1/1, `20261003T173509Z-p78377`, superseding the stale golden failure above. |

Latest generation passed at `20261003T173053Z-p62444`; latest format at `20261003T173420Z-p71221`. Subsequent documentation moved attempt-result ownership into persistence §24.2 and records restore obligations. New destination verification is still unwired. Portability artifact transport, physical collection, full obligation evidence, security review, companion adoption and final release gates remain open. No approval or conformance claim has been fabricated. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset.


### Next destination integration boundary

The single live coordinator preparation method is now extracted as `prepareVerificationMember`; ordinary lifecycle execution still calls it. The extraction itself is awaiting focused rerun. Destination integration must use that same ordered verification/index/artifact path and publish within the parent Incident Portability transaction.

Before wiring the cohort, resolve three concrete boundary details in its contract and persistence projection: an embedded source-reference tuple alone cannot justify a candidate before safe manifest attribution; several newly embedded versions of one key must not acquire traversal-dependent release-sequence rollback outcomes; and redistribution denial must remain a transport eligibility decision rather than a fabricated cryptographic verdict. The current preparation-to-operation-member foreign key assumes every freshly verified container is attributable. Either explicitly close unattributable transport admission or correct that private model before accepting such input; do not create a false candidate to satisfy the foreign key. Required/optional classification remains exact and failures cannot silently publish partial trust.

Shared deadline/cancellation observation must continue through the remaining source preparation and common finalizer. Per-container independent Jobs, separate commits, fresh current-set resolution, retry rebasing and nil required checks remain prohibited. The parent Job terminal-effects participant is now available to close an aborted portable cohort without promoting its private preparation. These notes record remaining implementation work, not completed behavior or adopted evidence.

### Destination cohort production integration checkpoint

The A–F remediation remains **incomplete**. The destination path now connects admitted embedded containers to the production verifier and the parent Incident Portability transaction. This supersedes the unwired-foundation notes above, but does not close complete portability, recovery, security review or owner adoption.

- Transport admission streams exact descriptor-bound inputs, separates redistribution/unattributable denial from content verdicts, rejects contradictory source identity and prioritizes operational cleanup faults. Missing optional inventory members are not opened. The independent `portable-input.v1.json` fixture includes allowed and restricted transport variants.
- Durable preparation captures original local selections, retained input objects, canonical source catalog, dependency schedule, repository/key revisions and configuration. The parent Jobs lease authorizes admission and replay; retries do not reopen source containers or rebase onto later local versions. Pending work rejects before clock eligibility. Rejected admission removes only its uncommitted private objects; uncertain preparation commit retains potentially referenced bytes.
- The shared live verifier/index/artifact path verifies every new cohort member. Only fully verified predecessors may resolve fresh dependencies. Restarted attempts preserve interrupted history and use a new freshness instant. Publication merges compatible roots first, orders successful versions by numeric release sequence and retains destination envelopes, attestations, exact sets and pins within the incident transaction. Active selection and administrative disablement remain unchanged.
- The parent execution scope carries one budget and cancellation observer through source preparation and finalization. Its abort classification uses the adopted Extensions policy and finalizer, preserves recoverable lease/process interruptions and records typed operational outcomes. Restore now validates preparation identity, selection and object-reference consistency and terminal-history linkage.

The new focused tests cover transport contradictions and I/O faults, stable dependency ordering including cycles, expired execution budget and cancellation/deadline equality, durable replay, conflicting work, stale publication, interrupted-attempt freshness, real signed destination verification, transaction rollback, exact replay and unchanged activation. The service scenario now calls the public Reference Data `PrepareImport` boundary for the final successful import. It is not yet a complete HTTP embedded-pack acceptance matrix.

| Command / evidence | Result |
|---|---|
| `make generate` | Passed at `20261003T181105Z-p72637`, including the authored portable-input fixture. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers,module.reference_data.unit.canonical_pack_manifest` | Passed 2/2 at `20261003T184243Z-p8261`. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3 at `20261003T184244Z-p8491`. |
| `make service-backed-test-slice OWNER=module.incidentbundles` | Passed 6/6 at `20261003T184245Z-p8732`. |

During this slice, the initially mistyped test-row selector executed no tests. Subsequent failures were a test constant spelling, a missing reloaded input logical path, a fixture bootstrap-field spelling and a fixture hash that incorrectly included the transport newline. The corrected rows above pass. No production limit was increased and no historical data was rewritten. Migration head remains 57.

Remaining destination evidence includes HTTP embedded import under destination trust, missing/invalid required versus optional bytes, multiple fresh versions and dependent cohorts, conflicting roots, source-publication/commit faults, timeout/cancellation through the common finalizer and corrupt preparation recovery. Complete snapshot/artifact transport, authoritative-byte collection, obligation-level release evidence, independent security-review disposition and companion-owner adoption remain open. Prior broad check/release evidence is stale; no new release pass or adoption claim is made. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset.


### Destination rejection and recovery evidence

Required embedded content rejection now commits attributable failed-attempt and candidate findings with the failed parent Job, without publishing successful sibling content, trust proposals, a catalog or the incident. Operational aborts remain separate. Rejection publication retains the original monotonic budget and frozen revision guards; proven cancellation, expiry or stale state reclassifies uncommitted work without re-verifying or rebasing. Restore admits a catalog-free content rejection only when it exactly matches a completed rejected attempt.

| Command / evidence | Result |
|---|---|
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_consumers` | Passed 1/1, `20261003T191629Z-p2666`. |
| `make service-backed-test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.incident_bundle_export_and_import_preserve_an_ac_02b4ecb4b8` | Passed 3/3, `20261003T191632Z-p3173`. Includes live HTTP untrusted clock, missing required input, required/optional invalid signature, valid destination verification, unchanged activation, exact reuse and re-export. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | Passed 3/3, `20261003T191849Z-p56439`. Includes catalog-free rejection recovery and deliberate corrupted-history rejection. |
| `make lint-go-staticcheck` | Passed, `20261003T191633Z-p3680`. |

The preceding Reference Data service run `20261003T191631Z-p2890` failed because its intentional corruption fixture was stopped by the immutable-operation trigger. The corrected test disables that trigger only inside its disposable rolled-back corruption transaction, leaving the production guard unchanged. An earlier format run `20261003T191251Z-p87057` caught a missing source newline; the subsequent formats passed.

HTTP test setup failures preceding this checkpoint were an isolated-server fixture sharing its parent's object-store prefix (the run `20261003T185345Z-p72234` was canceled) and a synthetic bootstrap below the adopted root threshold (`20261003T185924Z-p6611`). The fixture now uses a distinct prefix and three root keys with threshold two; the passing HTTP rows above supersede both. The renewal test producer is not counted as independent cryptographic-vector evidence.

### Legacy schema retirement in progress

Appended migration 00058 removes only the four empty retired mutable stores after locked row-count preflight. It does not rewrite migration 26 or convert trust/history. Read-only cutover reporting now includes legacy activation and audit rows even on an otherwise current schema. The current recovery projection removes these four stores, with 141 authored tables and 110 required authoritative tables. No real deployment migration or reset has occurred. Generation and focused retirement/preflight validation are pending.

The full remediation remains **incomplete**. Historical snapshot/artifact transport, authoritative-byte collection, the remaining cohort/finalization/race evidence, focused security review and companion-owner adoption remain mandatory. Earlier broad check/release runs remain stale. Retained-run maintenance is skipped because `RESULTS_DIR` is unset.


### Legacy retirement validation checkpoint

Migration 00058 and recovery generation now pass initialization and current-schema drift checks. The safe preflight scenarios cover each retired table, preserve category counts, leave all four tables present on rejection and keep migration head 57. Source code uses the canonical stores only; the retired stores occur solely in historical migrations and explicit preflight/retirement evidence.

| Command | Result / retained root |
|---|---|
| `make generate` | Passed, `20261003T192643Z-p85550`. |
| `make migration-drift` | Passed 5/5, `20261003T192715Z-p88889`. |
| `make service-backed-test-slice OWNER=module.database_migrations ROWS=module.database_migrations.integration.production_ddl_v2_recurrence` | Passed 3/3, `20261003T192715Z-p88933`. |
| Recovery frozen-catalog unit row | Passed 1/1, `20261003T192715Z-p88930`. |
| Reference Data canonical-persistence service row | Passed 3/3, `20261003T193103Z-p63578`. |
| Operator migration-evidence transport row | Passed 1/1, `20261003T193103Z-p63583`. |
| `make json-shape-check` | Passed 3/3, `20261003T193103Z-p63501`. |
| `make lint-markdown` | Passed, `20261003T193103Z-p63847`, before this checkpoint addition. |

Generation initially failed at `20261003T192513Z-p83456` because dependent Graph restore fixtures still named the old recovery catalog binding. The authored fixtures were updated together. The first retirement test run `20261003T192715Z-p88924` passed 12/13 units; the failed new fixture attempted a restricted replication setting. It now constructs valid FK parents with normal migration-role privileges. The stale operator digest at `20261003T192715Z-p88946` was updated to the reviewed schema-58 output and passed above. The broader migration unit slice `20261003T193103Z-p63636` passed 8/9 and exposed a separate stale schema fingerprint in pgtest; that expectation is corrected. The complete migration owner rerun passed 9/9 at `20261003T193356Z-p10162`.

Current schema SHA-256 is `df58d0d882dcfd7cbf12b1b4d97851e3908f2d6836922c1ab704080e91deb2cb`. Recovery semantic catalog digest is `e0b2195e6dff2bb482a058c60d8088faf9400b4627fbb558a42939ebe6c005d3`; canonical artifact digest is `1b00cd1c3e2347e961d96794d323023ba6c7e8e247dc7b03db43e6107809029b`. The Graph restore binding digest is `38798b2c3711d452e28cfc3a3731036df419d92e1a0e2c30d1f39f6a53f47347`, and operator transport output digest is `d2b667e4af077b7470e5d9238bb179a3dba8cd1834ae9ba3d83cb89103984358`.

The recovery catalog schema now uses the existing format bound of 1–4096 tables; the current generation still fixes exact counts at 141 authored and 110 required tables. The prior 113–114 schema range was stale. No production table limit was raised. Legacy schema retirement does not close physical byte collection, historical artifact transport, independent security review, owner adoption or final release validation. Overall status remains incomplete.


### Historical artifact transport implementation in progress

Reporting now owns a closed optional artifact catalog, strict byte/identity/binding admission, explicit sensitive-map omissions, and transactional immutable source-evidence retention. Incident Bundles delegates through a narrow artifact port and does not read Reporting tables. Migration 00059 adds the retained files; the recovery catalog becomes 142 authored and 111 required tables. Startup validates restored source artifacts through the Reporting and Reference Data boundaries. These changes are awaiting focused evidence; the A–F remediation remains incomplete. No security approval or owner adoption is implied.

The first generation attempt at `20261003T195518Z-p41792` rejected migration 59's missing explicit foreign-key actions, constraint name and routine search path. Those declarations were corrected before rerunning generation. No deployment database was changed.


### Historical artifact transport first evidence

The optional snapshot path now exports native immutable models and completed render members, preserves source provenance, retains imported bytes in the incident transaction and re-exports them unchanged. It excludes sensitive reveal maps with exact omission records. The first live round trip also verifies no destination native snapshots, releases or approvals are fabricated, no active set changes, and bad hashes, undeclared files, altered models and supplied sensitive files publish no partial incident state. A subsequent recovery-corruption scenario is added and awaiting rerun.

| Command | Result / retained root |
|---|---|
| `make generate` | Passed, `20261003T200040Z-p51721`. |
| `make test-slice OWNER=module.reporting` | Passed 7/7, `20261003T200244Z-p60465`, including nested catalog closure and isolated cardinality/path guards. |
| Migration production-DDL recurrence service row | Passed 3/3, `20261003T200244Z-p60493`. |
| Incident Bundles artifact and embedded-pack round-trip service row | Passed 3/3, `20261003T200534Z-p37979`. |
| Recovery frozen-catalog unit row | Passed 1/1, `20261003T200534Z-p38001`. |
| `make lint-go-staticcheck` | Passed, `20261003T200245Z-p61314`, before the final recovery-test edits. |

The first HTTP fixture at `20261003T200244Z-p60489` incorrectly tried to patch the immutable incident title; it now changes the mutable phase and proves the snapshot still retains the earlier phase. JSON-shape checks at `20261003T200244Z-p60402` and `20261003T200534Z-p37887` found the new Reporting family missing from the fixed family registry and then its generated topology fingerprint awaiting regeneration. The Incident Bundles unit slice at `20261003T200534Z-p37986` passed 7/8; its export-surface guard exposed the earlier cancellation-finalization type missing from the allowlist. The operator golden at `20261003T200534Z-p38018` needs the reviewed migration-59 fingerprint. These projections are corrected; focused reruns are pending.

Migration head is 59. Schema SHA-256 is `17571ef18d83b6a9d163b7a54f7a69bc75fb4f03912e079e67d6f644a652b176`. Recovery catalog digest is `ff769efe796c3ac9194cbe6e531e82084d39d3d48b181d63b43f07840f40264b`; canonical catalog artifact digest is `7a8df5eebeaf7e1cdbcd826bd9d2e8183fa2fc5e035b81b2d63dbfa9667d2318`. The Graph restore binding is `01a6b736bc33c211fb246bdf64fabc19b115cb431c0cc3f2c6902ee6f889a314`; operator transport digest is `b92d01e607131301f04916a933c271503bd824e5e3c6ef84fb8fdaee692a5727`.

The full remediation remains incomplete. Physical authoritative-byte collection, the remaining multi-member/cohort/finalization/race evidence, Reporting canonicalization discrepancies, obligation-level completeness, focused security review, companion adoption and fresh final release gates remain open. Retained-run maintenance is skipped because `RESULTS_DIR` is unset. No deployment migration or reset has occurred.


## Historical artifact transport — completed focused reruns

All result roots below are under `.cartulary/test-results/`. These are focused implementation evidence, not adoption or release approval.

- Incident Bundle artifact and pack HTTP round trips: `20261003T201149Z-p97400`, 3/3 passed.
- Incident Bundle public surface: `20261003T201149Z-p97409`, 1/1 passed.
- Operator migration transport golden: `20261003T201149Z-p97432`, 1/1 passed.
- JSON shape: `20261003T201149Z-p97364`, 3/3 passed.
- Migration drift: `20261003T201149Z-p97358`, 5/5 passed.
- Reporting snapshot/render integration: `20261003T201149Z-p97501`, 3/3 passed.

Migration 59 retains source Reporting artifacts separately from native snapshots, releases and approvals. Startup and restore validate their closed catalog, exact source pack bindings, checksums and member inventory. Sensitive reveal maps remain explicitly omitted. Current schema inventory is 142 authored tables / 111 required authoritative tables. No deployment database was migrated.


## Authoritative byte collection and backup retention

Production collection now runs after terminal-Job reconciliation before readiness and retries durable removal tombstones once per minute. The storage port returns a shared publication namespace lease; the registering transaction holds a shared database retention guard. Collection acquires the exclusive filesystem namespace before its exclusive database guard, retires unreferenced objects before physical deletion, and retains bytes on an uncertain retirement commit. Backup capture takes its guard before opening the repeatable-read snapshot and keeps it through object streaming.

Removed extracted copies are reclaimed only while the version is still removed, inactive, unpinned and free of pending work. Successful envelopes retain their exact signed containers; full restore authenticates those historical bytes at the successful instant and does not recreate removed indexes. Terminal preparations may release byte ownership only with another retained owner or a recorded unavailable generation. Exact reimport publishes distinct physical identities and keeps removal history. Unknown publication leftovers are collectible only after all live publisher leases and database registrations have resolved.

Focused evidence currently includes passing scenarios for live publication exclusion, lost retirement acknowledgement retaining bytes, backup/snapshot lock order, capture-wide exclusion, unregistered orphan collection, removed extraction collection, full historical verification, queued exact reimport, subsequent collection and Base integrity. The encompassing persistence row still awaits a clean rerun after tightening terminal portable-history ownership checks; do not treat these scenarios as a complete row pass yet.

- `make generate`: passed, `20261003T203529Z-p59508`.
- `make test-slice OWNER=platform.rootedfs`: 2/2 passed, `20261003T203551Z-p62757`.
- `make test-slice OWNER=module.reference_data`: 22/23 passed, `20261003T203551Z-p62756`; failed portable replay compared a process-local publication lease. Corrected to compare durable identity.
- `make test-slice OWNER=module.recovery`: first 21/24, then 23/24 at `20261003T204140Z-p93922`. Stale catalog golden/count expectations and missing Base fixtures were corrected. The remaining operator process rows passed 8/8 in focused rerun `20261003T204711Z-p90865`.
- Incident Bundle artifact/pack HTTP round trips: 3/3 passed, `20261003T204140Z-p93946`.
- `make lint-markdown`: passed, `20261003T204140Z-p94146`.
- `make lint-go-staticcheck`: passed, `20261003T204711Z-p91394`.

No migration, reset, deployment, commit or push occurred. Generation and focused checks do not close owner adoption, licensing, security review or the final release gates. `RESULTS_DIR` remains unset, so retained-run maintenance is skipped.


The complete persistence rerun passed 3/3 at `20261003T204937Z-p57894` using `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence`. This supersedes the collection checkpoint's pending-row status. The preceding failures exposed and corrected process-local replay comparison, terminal ownership validation, and a nullable SQL projection for selections with no input object. Historical corruption rejection remains covered.


## Administrative eligibility and reachable limit evidence

Administrative resources now expose required, non-null `pending_work` and `reproducibility_pinned` booleans from the same database snapshot. Pending work is selected-key-wide. The administration model blocks known ineligible actions and displays health independently of disablement, retention blockers, and current verification separately from historical snapshot provenance. The coordinator still rechecks every mutation transactionally. Core 01 and the pending OpenAPI 2.0.0 change set carry the new fields; the historical baseline is unchanged.

The Reference Pack draft now uses Core's single activation/removal error registry, including common operation rejections and pending-before-envelope precedence. It distinguishes definitive missing bytes from operational read failure and closes timeout equality. Current record owners expose no custom Host/Evidence registry assignment, and Indicators requires exactly the nine Base tokens; §18 documents why that custom-token fallback scenario is currently unreachable without inventing a record field.

All result roots are under `.cartulary/test-results/`:

| Command | Result / retained root |
|---|---|
| `make generate` | Passed, `20261003T210312Z-p96369`. |
| `make test-slice OWNER=module.reference_data` | Passed 23/23, `20261003T210850Z-p53778`, including administration, browser and the complete streamed file boundary. |
| `make frontend-typecheck` | Passed 2/2, `20261003T210850Z-p53855`. |
| `make frontend-import-boundary-check` | Passed 2/2, `20261003T210850Z-p53863`. |
| `make openapi-compatibility-check` | Passed 4/4, `20261003T210413Z-p4932`. |
| `make json-shape-check` | Passed 3/3, `20261003T210413Z-p4915`. |
| `make lint-markdown` | Passed, `20261003T210413Z-p5120`, before this ledger update. |

Superseded failures: generation `20261003T210101Z-p90986` and compatibility `20261003T210151Z-p92427` detected the newly required release fingerprints; generation `20261003T210230Z-p93177` detected unsorted authored test selectors. Reference Data `20261003T210413Z-p4957` passed 20/23 while the public required-member assertion and browser fixtures still had the retired resource shape; frontend type checking `20261003T210413Z-p5071` identified that same fixture omission. The later passes above cover those corrections.

Reporting canonical serialization is under repair and is not covered by those earlier passes. Full remediation and adoption remain incomplete. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset.


## Reporting canonical serialization repair

Reporting now implements its own UTF-8 byte-ordered, shortest-escape, integer-only canonical JSON. It rejects duplicate keys, malformed Unicode, fractions/exponents/negative zero, unsafe integers and opaque binary producer values before hashing. This remains distinct from Reference Data's RFC 8785 canonicalizer. Independent literal vectors cover the differing Unicode order, scalar boundaries, idempotence and the schema-plus-LF hash domain. A complete model also survives JSON persistence without changing canonical bytes. The obsolete golden for Go struct declaration order was replaced with those independent byte/hash vectors and a full model persistence check.

Export models, redacted models, profile views, redaction/token/reveal manifests and render metadata now use schema-domain object digests. Physical file and Extensions output checksums remain raw SHA-256; the render participant validates and translates those separate identities explicitly. Portable source-object admission now applies Reporting's number and Unicode rules to original bytes before reconstructing JSONB values. Its catalog continues using its separately declared RFC 8785 transport serialization. The private release Job codec carries normalized request bytes as base64 and does not use the canonical-object encoder.

Incident artifact/pack HTTP round trips passed 3/3 at `20261003T212032Z-p49114`. Reporting's first reruns exposed a removed encoder import (`20261003T211537Z-p98449` and companion Incident run `20261003T211537Z-p98462`), the private binary Job codec and obsolete golden (`20261003T211728Z-p22142`), and an integration assertion that conflated object digests and file checksums (`20261003T212032Z-p49106`, 6/7). Those issues are corrected; a complete Reporting rerun is pending. Staticcheck at `20261003T212032Z-p49639` found an unused fixture manifest assignment in the new file-boundary test; that assignment is corrected.

This closes neither the separately identified generated-ID/timestamp discrepancy nor full Reporting owner conformance. Reference Pack release/adoption gates and fresh broad final evidence remain incomplete. `RESULTS_DIR` is unset.

The complete Reporting slice subsequently passed 7/7 at `20261003T212347Z-p99337`, covering the corrected source object/hash boundaries and native snapshot/release flows. Staticcheck passed in the corresponding final focused rerun. These supersede the Reporting and fixture failures above; generated identifiers/timestamps and the remaining full-plan obligations are still open.


## Broad verification and fallback attribution checkpoint

The full remediation remains incomplete. `make test-fast` completed 735/737 units in `20261003T212557Z-p24748`; stale fixtures omitted the server collection lifecycle field and required administration blocker fields. Those fixtures are corrected, with focused passes in `20261003T213314Z-p18563` (server, 1/1) and `20261003T213322Z-p19236` (protocol, 2/2).

`make agent-finalize` passed in `20261003T213420Z-p20029`. `make check` then completed 988/990 units in `20261003T213440Z-p24271`. Its failures identified missing exact importer registrations for the new owner-port wiring and a direct execute grant on Reporting's private artifact trigger. The authored import registry now lists the reviewed files; migration 00060 revokes the unnecessary grant without editing migration 00059. `release-check` did not execute because `check` failed. Retained-run maintenance was skipped because `RESULTS_DIR` is unset. These corrections require fresh evidence.

Migration 00061 and the coordinator now persist the displaced version when safety fallback selects Base. The required nullable `fallback_from_version` administrative field is separate from health and immutable provenance. Unchanged claimed-profile reconciliation preserves it; different selection and unclaimed reconciliation clear current attribution. Preflight rejects intermediate fallback history whose current cause cannot be reconstructed safely. HTTP, restart/reconciliation, explicit replacement, schema, UI and preflight acceptance additions are pending verification.

The fixture-manifest contract required by RP-REQ-256 remains absent. Obligation-level evidence and fixture registration must be completed explicitly; the broad passing units do not close that conformance requirement. Reporting identity/timestamp reconciliation, remaining concurrency/finalization cases, security authority and companion adoption remain open.


Fallback acceptance passed in the complete Reference Data slice, 23/23 at `20261003T215904Z-p98469`, and artifact portability passed 3/3 at `20261003T215904Z-p98532`. Frontend types passed 2/2 at `20261003T215904Z-p98692`, importer boundaries passed 3/3 at `20261003T215904Z-p98721`, and migration drift passed 5/5 at `20261003T215904Z-p98441`. Generation passed at `20261003T215817Z-p95017` after recording the new field in the pending OpenAPI release.

PostgreSQL parity initially exposed missing append-only classifications for the two portable preparation tables and imported Reporting artifact files (`20261003T215904Z-p98492`); the authored generator now projects their existing intended immutable grants. The operator golden initially retained the old migration inventory (`20261003T215904Z-p98491`). After both fixes and generation (`20261003T220123Z-p91277`), PostgreSQL parity passes 3/3 at `20261003T220205Z-p94753`, operator evidence passes 1/1 at `20261003T220205Z-p94768`, and the recovery operator slice passes 8/8 at `20261003T220205Z-p94782`. Migration source digest is `5c4e74637434b98fbd36b56b9f84208dd016f107d4fe5b8e7fb6dc0f199d7d97`.

The administration resource now additionally projects exact verified dependency tuples, using explicit null before successful verification and an empty array for a verified empty list. Expandable UI details show source/profile, dates, logical digests, prior activation, missing reason, and dependencies. Keyboard browser acceptance and safe text rendering tests are added; this subsequent metadata slice requires its own verification.


### Administrative metadata, fixture registration and rendered review checkpoint

- The administrative projection now includes exact dependencies (`null` before success, an explicit array afterward) and retained Base fallback attribution. The UI separates fallback from health and expands source/profile dates, digests, prior activation, missing reason and dependencies. Core 01 and the pending OpenAPI release carry the same required fields; generated clients were refreshed.
- Reference Data passed all 23 units at `20261003T221141Z-p69963`, including the new metadata and keyboard browser assertions. The combined invocation then rejected an inapplicable `OWNER` input for the frontend target; the separate frontend type and boundary invocations passed at `20261003T221330Z-p14222` and `20261003T221330Z-p14230`.
- The private seeded UI review used the cartulary-ui-review skill. Both inspected captures showed wrapping digests, readable details and visible keyboard focus; both axe scans completed with zero violations or incomplete findings. The terminal receipt at `20261003T221321Z-p9692/ui-review/terminal.json` reports two captures, zero console errors/failed requests and complete cleanup. Private screenshots and app observations were disposed by the harness. This is supporting review evidence, not a conformance claim. The first launch rejected machine-output mode before session creation; the corrected foreground invocation succeeded.
- Added the closed harness fixture-manifest schema and the first three expectation-bound manifests (independent manifest identity, complete set, signed container). Added the manually projected 264-requirement/54-criterion range registry. The test validates explicit null/omission/type/unknown fields, sorted/disjoint effects, root-confined regular inputs, symlink rejection and expectation identity. It rejects prose paths before filesystem access. The projection consistency report is `unmapped=0 unknown_requirement=0 unknown_acceptance_criterion=0`; this does not mean all obligations have implementation evidence.
- The fixture and traceability test passed at `20261003T222525Z-p57780`. JSON-shape initially rejected the schema attachment's validation token at `20261003T222526Z-p58002`; it was corrected to the registry's required `json-shape-check` token. A rerun is required. Registration of all remaining scenario fixtures and all 27 required families remains incomplete.
- Corrected the stale RP-REQ-262 clock assumption to reference RP-REQ-075's adopted design decision, preserving Base, historical-use and empty-refresh exceptions. No time-service inference or future subsystem is needed for the operator assertion.

The security-authority compliance-posture decision has been requested from the user. No approval, licensing determination for external datasets, companion-owner adoption, or full-plan completion is inferred from this work. No deployment database was migrated or reset. `RESULTS_DIR` remains unset.


## Cohort publication and boundary checkpoint

The full `make test-fast` passed 737/737 units at `.cartulary/test-results/20261003T223007Z-p67231`. The following `make check` failed 988/990 at `.cartulary/test-results/20261003T223007Z-p67243`; `release-check` did not run. Both failures were related to this work: a test-fixture setup triggered the generated-root write heuristic, and migration 61's partial fallback FK index did not satisfy the adopted full-index policy. Temporary fixture creation now lives in a separate test helper. Appended migration 62 replaces the partial index with a full leading-column index; applied migration bytes remain unchanged.

- `make backend-module-boundary-check`: passed 3/3, `.cartulary/test-results/20261003T225332Z-p52973`.
- Reference Data `make test-slice`: passed 23/23, `.cartulary/test-results/20261003T225332Z-p52700`, including request-shape precedence and content-count reachability.
- Database Migrations service-backed `production_ddl_v2_recurrence`: passed 3/3, `.cartulary/test-results/20261003T225332Z-p52743`.
- Incident Bundles service-backed round-trip row: passed 3/3, `.cartulary/test-results/20261003T230150Z-p95227`. Two fresh historical versions share one verification instant and publish in release-sequence order even when lexical order differs. Compatible root proposals advance trust once, publish both envelopes and pins, retain four attributable events and leave activation unchanged. Individually valid conflicting root bytes roll back all semantic effects, retain terminal operational-abort evidence and replay the same Job.

The new HTTP scenario initially failed at `.cartulary/test-results/20261003T225332Z-p52719`, `.cartulary/test-results/20261003T225735Z-p43317` and `.cartulary/test-results/20261003T225955Z-p69441`. Portability incorrectly copied the container digest onto root-update attestations; only the verification event now carries it, matching ordinary lifecycle publication and the closed attestation contract. Temporary bounded diagnostic instrumentation used to isolate that failure was removed.

Generation failed at `.cartulary/test-results/20261003T225235Z-p48573` because migration 62 had not yet received an authored source-owner allocation. After that allocation, `make generate` passed at `.cartulary/test-results/20261003T225311Z-p49413`. The operator digest row failed at `.cartulary/test-results/20261003T225332Z-p52747` because its expected catalog digest predated migration 62; its expected bytes have been updated and await rerun. The schema hash is now `94f4cf02311af1bed7c0ff4f8602148a6f3e27f0c337a11ced9a8bb40f76b53a`.

Forty-eight separately registered content fixtures now provide valid, malformed-structure and semantic-error inputs for all sixteen profiles. Their expected issue preimages/digests are calculated independently with Python JSON/SHA-256, never by the production verifier; the fixture runner must compare the exact winning error and ordered issues. These additions await focused validation. They do not complete every required fixture family or owner adoption.


The 48 content-profile fixture manifests passed the canonical fixture/traceability row at `.cartulary/test-results/20261003T230642Z-p25227`. The updated operator digest row passed at `20261003T230642Z-p25234`; JSON-shape passed 3/3 at `20261003T230642Z-p25187`, migration drift 5/5 at `20261003T230642Z-p25199`, module boundaries 3/3 at `20261003T230642Z-p25367`, and Markdown lint passed at `20261003T230642Z-p25390`. Generation passed at `20261003T230610Z-p21894`. These are focused checks, not a complete release pass.

## Final-commit classification and live concurrency checkpoint

The late parent-publication fault scenario passed 3/3 at `.cartulary/test-results/20261003T231602Z-p90553`. A test-only nontransactional sequence proves that the incident, two successful envelopes, four attestations, two pins and root advancement were visible inside the parent transaction before the injected catalog failure. The entire semantic mutation rolls back; terminal attempt evidence remains attributable without condemning content. Earlier runs `20261003T231037Z-p39815` and `20261003T231324Z-p65444` exposed a missing grant on that disposable fixture sequence; only the fixture grant was corrected.

The shared Incident Bundles transaction previously classified every lost commit acknowledgement as unknown without consulting the already-written Job proof. It now registers the existing Extensions owner finalizer's exact-success classifier during terminal publication. The physical classifier is shared by success, failure, cancellation and synchronous owner mutations; proof reads use a bounded context independent of caller cancellation. Missing or unreadable proof still does not prove rollback. Proven server rollback errors preserve their original classification, and unresolved mutation still enters the fatal integrity path. Server composition preserves the admitted database port when applying instrumentation, allowing the narrow transaction decorator to exercise the real parent transaction without a production fault toggle.

- Incident Bundles HTTP cohort, lost acknowledgement and explicit commit rollback: passed 3/3 at `20261003T232611Z-p18441` and again after classifier consolidation at `20261003T232958Z-p84905`.
- Extensions Job finalization: passed 3/3 at `20261003T232627Z-p29710`; the expanded success/failure/timeout/cancellation matrix passed 3/3 at `20261003T233233Z-p30452`. Every committed outcome survives context cancellation after commit. Run `20261003T232957Z-p84677` failed compilation because the new test incorrectly named a separate timeout status; Jobs correctly exposes timeout as failed, and the fixture was corrected.
- Shared cross-owner PostgreSQL protocol: passed 3/3 at `20261003T232627Z-p29712`.
- Reference Data live lifecycle row: passed 3/3 at `20261003T233907Z-p64651`. A paused import retains root-1 inputs while a different key commits root 2; it verifies its frozen bytes, rejects stale publication, preserves activation, replays its original failed Job and accepts only a new explicit import under root 2. Concurrent activation requests both capture the prior set; exactly one commits, the other returns stale state, and a new explicit request preserves both selections. Historical pin acquisition blocks a live HTTP removal, which rejects as pinned after the pin commits.
- Module boundaries: passed 3/3 at `20261003T233907Z-p64747`. Generation passed at `20261003T233603Z-p50968`.

Routing discovery initially used the nonexistent owner `platform.extensions`; it was corrected to the actual `module.extensions` and `module.crossownertransaction` owners before verification. Full remediation, Reporting identity/timestamp closure, complete fixture/evidence registration, security-authority disposition and companion adoption remain incomplete. `RESULTS_DIR` remains unset; retained-run maintenance is skipped.


## Full-check checkpoint and admitted Reporting identity work

Before the subsequent Reporting identity correction, `make agent-finalize` passed 1/1 at `20261003T234052Z-p85331`, `make test-fast` passed 737/737 at `20261003T234131Z-p89465`, and `make check` passed 990/990 at `20261003T234131Z-p89477`. `make release-check` was deliberately interrupted to resume required implementation; it completed 958/1259 units and reports `cancelled` at `20261003T234131Z-p89481`. This is not release evidence or a product failure. The earlier completed suites do not cover the subsequent Reporting edits. `RESULTS_DIR` remains unset.

Reporting now separates immutable snapshot materialization from admitted release/preview models. Snapshot admission allocates its real ID before constructing nested source references. Release admission freezes the exact release ID and a microsecond timestamp; canonical export, deck and token identifiers use their declared schema-domain tuples. Previews use distinct hash domains with the companion's actual attempt ID. A failed render retains its admitted release identity, and execution cannot fill a missing admission time from its current clock. Generated render artifacts retain the same identity/time. These changes require focused acceptance before being considered complete.

Appended migration 63 and Reporting's read-only preflight reject incompatible retained snapshot, job, imported-artifact and preview history without reconstructing it. This has not been run against a real deployment. Reporting 1.4.0 and Report Composition 1.3.0 are explicitly proposed amendments over their adopted bases; no owner approval is inferred. The implementing-agent security review is recorded in `docs/handoffs/reference-pack-security-review.md`; independent authority/compliance disposition remains open. Full remediation and adoption remain incomplete.


The first Reporting identity slice failed 6/7 at `20261004T001610Z-p80401`; its focused diagnostic run `20261004T001910Z-p2657` identified a three-segment test path that created no record. The corrected fixture reached the intended nested-source assertion. The next slice failed 6/7 at `20261004T002706Z-p30945` because the new migration fixture omitted required Job metadata. After correcting that setup, Reporting passed 7/7 at `20261004T003053Z-p98066`, including independent ID vectors, exact timestamp admission, immutable participant input, live snapshot/release metadata and safe migration rejection. Report Composition passed 3/3 at `20261004T002706Z-p30954`; Incident Bundle round trips passed 3/3 at `20261004T002706Z-p30976`.

The new preflight's dynamic relation spelling was rejected by the SQL ownership checker (`20261004T002706Z-p31147`). It now uses explicit owner-table queries. Module boundaries passed 3/3 at `20261004T003053Z-p98240`, JSON shape passed 3/3 at `20261004T003053Z-p98001`, and migration DDL recurrence passed 3/3 at `20261004T003053Z-p98107`. The migration source hash is `9052cea15514ace7ede6cbc4d214c585c27c3ced742595b7b9bfd5269e6176f3`. The operator fingerprint run `20261004T003053Z-p98087` failed against its prior 62-migration inventory; the expected 63-migration digest is now `9925ebfb54dc9883d3fca57c01b9c0512c4a2cadd32d4ab18cf0f3e5d810890f` and requires rerun.

A subsequent schema-version correction advances changed redaction, token, reveal-map and deck artifacts independently to v2, as required for the release/preview identity union. This correction requires fresh focused validation; it does not invalidate the preceding findings about behavior. Full release evidence, complete obligation registration, reset rehearsal, security-authority disposition and companion adoption remain open.


## Disposable reset rehearsal

The public reset procedure was rehearsed against unique Compose project `cartulary-reference-pack-reset-k6w7meiq`, bound only to `127.0.0.1:54215`. Its PostgreSQL data used a 1 GiB tmpfs, with no named database volume. The selected Compose input lived at `/tmp/cartulary-reference-pack-reset-k6w7meiq/compose.yml`; no normal development service or retained incident database was selected.

- `make db-reset CARTULARY_CLEANUP_DRY_RUN=1` with that explicit Compose/port/runtime environment passed at `20261004T003507Z-p50193`. The retained log names the isolated Compose path.
- `make db-reset CARTULARY_DESTRUCTIVE_CONFIRM=db-reset` passed first initialization at `20261004T003609Z-p50844` and a second deliberate reset of the initialized database at `20261004T003700Z-p51853`.
- Read-only inspection after the second reset confirmed migration head 63, zero snapshots and zero Reference Pack versions before application startup. This is schema/reset evidence; Base readiness remains covered by the separate startup acceptance tests.
- Cleanup with the same isolated environment and `make services-down` passed at `20261004T003723Z-p52539`; read-only inspection confirmed no project container remained. Historical backup conversion was neither attempted nor implied. The guide still requires preserving a matching historical application and backup before resetting affected disposable data.


## Reporting and canonical-admission verification

The final independent artifact-version correction passed Reporting 7/7 (`20261004T003936Z-p58677`), Incident Bundles 3/3 (`20261004T003936Z-p58688`), and the operator inventory row 1/1 (`20261004T003936Z-p58690`). The changed render artifacts use their own v2 identifiers; source snapshots have a separate model identity, release/preview identities have distinct domains, and all retained render timestamps derive from admission. Module boundaries passed 3/3 (`20261004T003937Z-p58878`), JSON shape 3/3 (`20261004T003936Z-p58631`), migration drift 5/5 (`20261004T003936Z-p58647`), and Markdown lint passed (`20261004T003937Z-p58912`). These close the specific model-identity/timestamp mismatch, not full Reporting owner adoption.

Nested profile admission now mutates every present nested object field and array item in every canonical profile row through the full content validator. Omission, disallowed null, wrong type and unknown member must reject before indexing. Explicit nullable-field expectations are independent of the compiled schemas. The expanded profile row passed 1/1 at `20261004T004731Z-p16509`.

Forty-nine additional canonical admission manifests retain independently produced archive bytes and original JSON errors. Equivalent ZIP, ustar and GZIP-ustar containers carry the same signed operator pack and pass the full verifier with exact content, signer and expiry identity. Archive attacks retain exact public errors and all ordered issue tuples; invalid input writes no extracted member. JSON vectors preserve malformed Unicode, duplicate members and noncanonical input that ordinary decoding could erase. Independent Python standard-library archive/JSON/SHA-256 construction supplies the expected bytes and digests; the Go implementation does not generate its own oracle. Both fixture and live-engine rows passed 2/2 at `20261004T005431Z-p34922`, with 100 registered fixture manifests. An earlier invocation used an invalid row selector and ran no tests; the corrected command was `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_consumers`.

Nineteen additional trust manifests passed for exact root/signer proposals, retained replay, rollback, all-role signature rejection, root thresholds, expiry equality and target substitution at `20261004T010225Z-p46051`. Eleven license/source-profile manifests passed at `20261004T010517Z-p56019`. These validate runtime provenance admission, not rights to redistribute external datasets. The overall remediation remains incomplete.


## Lifecycle, dependency and startup closure

The canonical fixture runner now has a separate narrow row, `module.reference_data.unit.canonical_fixture_manifests`, so a failed canonical expectation remains visible without being displaced by unrelated nested-schema output. Its 155-manifest checkpoint passed at `20261004T014402Z-p69789`; JSON shape passed 3/3 at `20261004T014402Z-p69751`. The inventory includes isolated 64/65 set guards, the traceability projection, registry-key/algorithm/alias failures, and missing, cyclic, conflicting and transitive dependency inputs. Dependency graphs are isolated authenticated-resolution inputs; they are not represented as independently signed containers. Expected diagnostic order follows the declared tuple, including the issue-ID tie break.

Fourteen canonical lifecycle steps now run through the real coordinator, PostgreSQL, Jobs and the shared finalizer: import, same-transaction replay, activation, active reverify, disablement, envelope renewal, refresh, cancellation, empty refresh, removal, exact reimport, reactivation, and claimed/unclaimed Base reconciliation. They assert independently calculated set identities, administrative state, temporary cleanup and declared side effects. The integration row passed 3/3 at `20261004T013408Z-p28442`. Initial fixture failures were an overstrict no-attestation expectation on unchanged startup and an omitted registry-usage test port; neither was a production authorization bypass.

A subsequent production audit found that startup could preserve an active imported version without rechecking its retained bytes. Base reconciliation now captures one coherent selection and relevant revisions, checks retained bytes and current content/runtime compatibility outside publication locks, and compares captured revisions before mutation. Definitive loss atomically publishes health, fallback/dependency consequences, attestations and the resulting set. Storage execution failures publish none of those effects. Freshness is not reapplied; healthy expired metadata remains usable at startup. Six database-backed scenarios cover healthy expiry, missing bytes, altered bytes, operational read failure, stale relevant state and unrelated revisions. The expanded persistence row passed 3/3 at `20261004T014402Z-p69801`.

Canonical owner scenarios now register snapshot/report binding, embedded destination verification and historical required-content loss. With these additions all 27 fixture families have a registered execution boundary, across 158 manifests. Registration is not proof of exhaustive coverage within a family. Snapshot admission now has a real publication-barrier test that changes activation before worker completion, checks rendering and rerendering against the admitted pin, and rejects loss of the pinned Base manifest while an imported replacement remains active. Current validation is in progress. Destination verification passed 3/3 at `20261004T015240Z-p3875`; the canonical manifest inventory passed 1/1 at `20261004T015240Z-p3849`.

Related new-test failures at `20261004T015240Z-p3863` and `20261004T015240Z-p3903` respectively selected the redaction manifest instead of the render-bundle manifest and incorrectly expected an active pointer to survive definitive required Base loss. Both expectations were corrected. The boundary checker at `20261004T015240Z-p4131` required the exact new Reporting test importer to be registered for Reference Data's test-support port; the production import boundary was not widened. Fresh validation follows below.

## Final boundary-fixture audit checkpoint

The [obligation index](reference-pack-obligation-evidence.md) now links the [per-limit evidence inventory](reference-pack-limit-evidence.md). Neither document is executable authority or an adoption record. The canonical inventory contains 163 manifests across all 27 declared families. The latest five cover complete longest Windows-event, CVE and SID identities and alias/source-ref excesses. Production Get/lookup exercises the longest identities, including a colon-bearing event provider, omitted version and explicit numeric version.

Additional compact complete recipes cover a 1048576-byte indexed NDJSON line, 64 profile-specific elements, 66 declared files, 64 source artifacts/dependencies/conflicts, exact original path/segment bounds, an exact compression ratio of 100 with a one-byte-over numerator, 64 authorized ordinary signers, two-MiB TUF files, and eight-MiB aggregate metadata and deployment bootstrap. The independent 128-signature disjoint-root vector remains the cryptographic cross-language oracle. Configured container/extracted/member limits also execute against a complete independently signed container at equality and one-over. Lookup defaults/endpoints and exact 1000/1001 diagnostic retention now have direct assertions. Unreachable canonical object ceilings use explicit guards and constraint arguments; production limits were not increased.

| Command | Result / retained run |
| --- | --- |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.all_profiles_consumer_boundary` | PASS 3/3, `20261004T020443Z-p30137`; all sixteen profiles and longest identity round trips. |
| `make service-backed-test-slice OWNER=module.reporting ROWS=module.reporting.integration.snapshot_creation_resolves_and_persists_an_immut_afe792169c` | PASS 3/3, `20261004T020006Z-p78396`; admission barrier, activation change, rerender and required pinned-content loss. |
| `make service-backed-test-slice OWNER=module.reference_data ROWS=module.reference_data.integration.canonical_persistence` | PASS 3/3, `20261004T020006Z-p78407`; startup/lifecycle/history and exact owner-fixture assertions. |
| `make backend-module-boundary-check` | PASS 3/3, `20261004T020006Z-p78578`; exact Reporting test-support exception is registered. |
| `make lint-markdown` | PASS, `20261004T020006Z-p78615`; predates this limit-index update. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_fixture_manifests,module.reference_data.unit.canonical_consumers` | PASS 2/2, `20261004T021226Z-p62233`; 163 fixtures and complete configured archive boundaries. |
| `make test-slice OWNER=module.reference_data ROWS=module.reference_data.unit.canonical_pack_manifest,module.reference_data.unit.canonical_pack_profiles,module.reference_data.unit.canonical_pack_trust,module.reference_data.unit.canonical_pack_diagnostics` | PASS 1/1 grouped execution unit, `20261004T022751Z-p7856`; new complete boundary recipes and exact retention assertions. |
| `make generate` | PASS, `20261004T022619Z-p4548`; latest authored source and fixture projection generation. |
| `make format` | PASS 2/2, `20261004T022605Z-p99221`; subsequent numeric recipe correction requires no formatting change. |

Related failures are retained: `20261004T020443Z-p30128` had an incorrect expected diagnostic omitting the declared 64/65 count tokens; the independently specified expectation and its canonical issue hash were corrected. `20261004T021922Z-p82982` rejected the new test producer's one-key bootstrap because the adopted root policy requires three keys and threshold two; the producer now uses a valid 64-key root. `20261004T022352Z-p97534` failed compilation because two local recipe variables used the same declaration; the second now assigns the existing variable. A separate invocation using the nonexistent selector `module.reference_data.unit.canonical_diagnostics` was rejected before execution and corrected to the registered `canonical_pack_diagnostics` row. None of these failures is represented as passing evidence.

Owner adoption, final independent security review and the full final-tree gates remain unrecorded at this checkpoint. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset. No deployment reset or data migration was performed outside disposable test databases.


## Consumer machine-contract closure checkpoint

The five consumer operations now have authored closed request/result schemas and an operation catalog defining scalar check order. Native callers and strict original-JSON adapters use the same compiled request projection. Required nullable result arms remain explicit, and success/error are mutually exclusive. UTF-8 byte bounds and Unicode-scalar bounds are distinct owner keywords. Actual Get/lookup results are additionally checked against their selected profile in integration evidence; ordinary production reads retain typed constructors over verified immutable indexes rather than reparsing every result.

The focused live all-profile slice passed 3/3 at `20261004T025627Z-p4834`, including longest identities and all five result projections. Consumer units passed 1/1 at `20261004T025627Z-p4828`. JSON shape passed 3/3 at `20261004T025425Z-p95335`; generation passed at `20261004T025557Z-p1786`. The first unit invocation, `20261004T025425Z-p95367`, failed compilation on an unused import introduced by this refactor; that import was removed before the passing rerun. Owner boundary checks passed 3/3 at `20261004T030245Z-p30047`.

The prior broad runs were deliberately canceled after this missing projection was identified: `test-fast` at `20261004T023224Z-p14613` completed 141/737 units, `check` at `20261004T023224Z-p14660` completed 231/990, and `release-check` at `20261004T023224Z-p14683` completed 173/1259. No failure was reported before cancellation, but none of these runs is complete passing evidence. Their prerequisite `agent-finalize` passed at `20261004T023155Z-p10216` and is superseded by the next finalization.

Consumer result mutation evidence now also checks explicit nulls and array elements using independently stated nullable fields and deterministic traversal. Formatting passed at `20261004T030342Z-p30741` and generation at `20261004T030358Z-p35572`; the final consumer unit rerun passed 1/1 at `20261004T030413Z-p38440`. Fresh broad gates follow. Retained-run maintenance remains skipped because `RESULTS_DIR` is unset. No human adoption or security approval is inferred.


## Full-suite failure diagnosis and correction checkpoint

`make agent-finalize` passed 1/1 at `20261004T030610Z-p40218`; retained-run maintenance was skipped because `RESULTS_DIR` was unset. Markdown lint passed at `20261004T030610Z-p40332`, and whitespace validation passed before the broad runs.

- `make test-fast` passed 737/737 at `20261004T030640Z-p46384`.
- `make check` finished 989/990 at `20261004T030640Z-p46428`. The sole failed execution unit was Recovery's `selected_backup_restore_fails_before_readiness_w_3a6ccb7d7a` row: all restore assertions passed, but force-dropping a disposable PostgreSQL database exceeded the fixture cleanup deadline. No production restore failure was observed. The exact row is being repeated without concurrent broad suites; neither the cleanup deadline nor the acceptance requirement is widened.
- `make release-check` was deliberately stopped after confirmed failures so corrections could be applied. `20261004T030640Z-p46445` retains 1139 passed, 4 failed and 116 canceled units of 1259; it is not complete release evidence. The completed failed groups were timeline range scrolling, timeline keyboard support, incident import and canonical Indicator creation.
- The isolated timeline scrolling rerun also failed at `20261004T031959Z-p98398` (9/11 execution units, including its failed group summary). The range ended 52 pixels short, compared with 30 in the full run, while the five-second wall-clock assertion expired. The production scroller intentionally caps work per animation frame. Functional acceptance now uses Playwright's controlled animation clock, preserving exact range, no-query and no-document-scroll assertions. It does not increase product speed or turn the test into a performance claim.
- The existing keyboard fixture expected obsolete retention text. It now checks the named Resume/Discard controls, activates Resume through the keyboard, verifies the exact retained draft and checks Escape/focus recovery. No production inspector behavior changed.
- The browser incident-import producer still emitted bundle 4 and an unversioned empty reference array. It now selects catalog version 5 and emits the canonical empty `reference_pack_refs.v1` object with deterministic file ordering. Server integration separately continues to reject versions 1–4 without semantic effects.
- Canonical Indicator browser and accessibility fixtures still supplied `NEW[.]EXAMPLE`/`ALPHA[.]EXAMPLE` as valid domain input. RP-REQ-123 admits ASCII domain labels; defanging is a separate output operation. Positive fixtures now supply `NEW.EXAMPLE`/`ALPHA.EXAMPLE`, and an independently stated negative domain vector retains bracketed-dot rejection. No refanging alias was introduced.

A read-only audit also identified the original JSON decoder's previously unnamed depth limit. RP-REQ-029 and the fixed typed limit now define 10000 open object/array containers, top-level container depth 1, with scalar leaves adding no depth. A complete authenticated TUF signed-extension fixture reaches equality; 10001 is rejected during original JSON admission. Byte bounds remain unchanged. Format passed at `20261004T034826Z-p77805`, generation at `20261004T034844Z-p82674`, and trust/indicator units passed 1/1 at `20261004T034914Z-p85619`. Focused browser, Recovery and final sequential full-suite evidence follow.

The candidate definition of done additionally requires recreatability and interchangeable independent implementations. The checked-in cross-language vectors establish their named byte/digest/signature boundaries; they are not evidence that a second complete subsystem implementation was recreated. That broader adoption proof remains unrecorded, alongside exact companion-owner adoption and the security-authority compliance decision. This checkpoint does not promote the draft or claim completion.


The targeted corrections passed: timeline functional scrolling and keyboard retained-draft recovery, 13/13 at `20261004T034933Z-p86329`; canonical Indicator creation and accessibility, 13/13 at `20261004T035132Z-p47611`; Recovery's exact previously failed row, 3/3 at `20261004T035031Z-p26546`; JSON shape, 3/3 at `20261004T035031Z-p26493`; frontend types, 2/2 at `20261004T035031Z-p26672`. The Recovery cleanup failure did not recur without concurrent broad suites. No cleanup deadline was changed. Whitespace validation passes. The incident-import browser row subsequently passed 11/11 at `20261004T035241Z-p87079`. Sequential final full-tree gates follow the diagnostic-limit closure below.


### Diagnostic limit contract closure

The final projection review found draft Table 26-C used names that differed from the typed limit catalog, and several streaming guards omitted numeric measurements. The draft now uses the catalog's single vocabulary; the bootstrap, source-artifact, dependency and conflict limits are also present in that catalog. Typed schema annotations attach the registered limit ID to array/object cardinality findings. Shared bounded numeric constructors preserve safe counts across archive, manifest, TUF and NDJSON guards, including the distinct 128-signature root-update limit. Count strings admit unsigned 64-bit measurements without widening signed JSON numbers or path indexes. The depth guard explicitly uses the enclosing document's original-JSON failure. Failed initial candidate retention is consistent with the outcome matrix. New focused evidence and the sequential final gates remain pending until recorded below.

The numeric diagnostic changes passed `make format` at `20261004T040533Z-p27451`, `make generate` at `20261004T040547Z-p32315`, the combined diagnostic/manifest/trust row (1/1 execution units) at `20261004T040606Z-p35294`, and `make json-shape-check` (3/3) at `20261004T040606Z-p35256`. The regression checks assert actual emitted IDs and measurements, including 64/65 ordinary signatures, 128/129 root-update signatures, manifest arrays, complete container byte admission, path/segment lengths, NDJSON line stop, and unsigned-count overflow without widening path indexes. The complete owner slice and final gates remain pending.

The full Reference Data owner run at `20261004T040652Z-p36726` passed 22/23 units, including production integration and administration browser rows. The sole failure was the independent SID alias/source-reference overflow fixture still expecting `limit_id=null`; the focused manifest runner reproduced it at `20261004T040927Z-p78618`. Both authored expectations now name their registered limits, with issue identities computed from the specification's seven-field hash domain using Python JSON/SHA-256, not copied from the verifier result. Markdown lint passed at `20261004T040652Z-p36834` before this evidence addition. A subsequent owner run validates the corrected fixture expectations.

The corrected fixture row passed 1/1 at `20261004T041051Z-p82212`; regeneration passed at `20261004T041033Z-p79220`. Fresh `make agent-finalize` passed 1/1 at `20261004T041051Z-p82202`. Retained-run maintenance was skipped because `RESULTS_DIR` is unset. The source tree is now frozen for sequential `test-fast`, `check` and `release-check`.

For clarity, earlier checkpoints used “independent security review” too broadly. The requested focused implementation review and the security-authority compliance decision are distinct. The former is recorded in the security review; RP-GATE-010 explicitly requires the latter. This handoff does not add a separate mandatory external audit. The owner's explicit independent-implementation/recreatability proof and companion adoption gates remain as written.

### Sequential final gate evidence

`make test-fast` passed **737/737** execution units at `20261004T041155Z-p88329` in 325.388 seconds on the tree including the limit-diagnostic closure and corrected independent fixture expectations. `make check` and `make release-check` follow sequentially. No conformance or adoption claim follows from this fast-suite result.

`make check` passed **990/990** units at `20261004T041748Z-p71111` in 633.086 seconds, including the previously failing Recovery row. A subsequent read-only configuration review found the development example still claimed Reference Packs while omitting the newly required bootstrap. The example now explicitly uses `reference_pack.claimed=false`, preserving required Base startup; the guide explains explicit post-adoption opt-in with a real operator-selected bootstrap. It also removes obsolete “overlay” wording. No new capability toggle was added. This example/documentation follow-up is covered by the next focused configuration and final release runs; the two broad results above precede it.

The development-default follow-up passed platform configuration 2/2 at `20261004T042914Z-p58703` and fresh finalization 1/1 at `20261004T042914Z-p58689`. The release attempt `20261004T042948Z-p63192` exposed a browser harness coupling: it injected fixture trust settings while inheriting the development example's claim flag, producing `extension_config_without_claim` during startup. That run was deliberately canceled after diagnosis: **209 passed, 56 failed, 994 canceled** in 280.676 seconds. Most failures were fixture admission/reset cascades; the Jobs group also reported three disposable-database cleanup timeouts during the failed startup churn. The harness now explicitly claims Reference Packs alongside its existing fixture-only trust settings. The development default stays unclaimed. No production claim rule or cleanup deadline was weakened. Focused browser/Base startup evidence and a fresh release run follow.

The three Jobs cleanup cases passed 3/3 at `20261004T043802Z-p43324` without deadline changes. Generation passed at `20261004T043720Z-p40216`, shell lint 4/4 at `20261004T043802Z-p43486`, and finalization 1/1 at `20261004T043802Z-p43310`. Startup diagnostics also exposed a remaining namespace/identity bug: inactive Reference Pack settings reported `profile_id=reference_packs` rather than the registered `reference_pack`. The neutral configuration adapter now takes identity from the admitted claim registration instead of splitting the settings path. A composition-level regression asserts the exact owner identity for both rejected trust settings. These subsequent source changes require refreshed validation.

The repaired browser harness and Base startup slice passed **12/12** at `20261004T044205Z-p76698`, covering administration recovery and the disconnected/optional failure modes. The first new configuration regression at `20261004T044205Z-p76693` failed only because the fixture supplied one trust setting while asserting two findings; the actual diagnostic already reported the correct registered identity. The fixture now explicitly supplies both the clock assertion and a must-not-be-read bootstrap path. Finalization at `20261004T044205Z-p76688` passed before that test-input correction.


### Pre-production scope clarification and final visual correction

The project owner confirmed that this is pre-production and that no recreation or interchangeability reports exist or are required. The draft owner §32 now reflects that scope: independent vectors, behavioral checks, safe cutover and focused security review remain required; a second full implementation does not. No production compliance approval or formal owner adoption is inferred from that clarification. Core 00's remaining profile-major-1 table row was corrected to match the intended major-2 companion contract, with its pre-production status explicit.

Fresh finalization passed 1/1 at `20261004T044452Z-p29892`; configuration passed 2/2 at `20261004T044452Z-p29898`. `make test-fast` passed 737/737 at `20261004T044530Z-p34338` and `make check` passed 990/990 at `20261004T045139Z-p26860`. The completed release run `20261004T050233Z-p18405` passed 1257/1259 units in 2013.918 seconds. Its two failed units were the visual group and its required summary: the group had two failing screenshot scenarios and 44 passing scenarios. No backend, functional browser, stateful, accessibility, recovery or migration unit failed in that run.

Artifact review used the UI-review skill in artifact mode at `20261004T053640Z-p45892`. All 14 imported capture bundles were inspected, including expected/actual/diff Evidence framing and all 13 Reference Pack actual states; image digests were verified before viewing. Reconciliation accounts for all 255 active captures/goldens, with no orphan, missing, ambiguous or unresolved registered fixture. The Reference Pack differences follow the planned health, current-verification, immutable-provenance, details and removal UI, replacing the retired checksum-labelled verifier. The unrelated Evidence difference is a retained horizontal offset at a sticky gutter. Its anchor now establishes the declared left edge before capture rather than inheriting prior interaction scroll. No production grid behavior, tolerance, renderer, mask or viewport changed. UI-review cleanup completed successfully; artifact-only review supplied no DOM or accessibility observations.

Format and generation passed at `20261004T053942Z-p50028` and `20261004T053949Z-p54982`. Focused framing verification and the maintained transactional visual refresh follow; final release evidence remains incomplete until those results are recorded.


### Reviewed visual refresh

Accepted trigger: the existing Reference Pack golden images described retired checksum verification and omitted the implemented health/current verification, historical-provenance explanation, version details and removal controls. The already-passing functional administration scenarios establish those changes. `make browser-e2e-visual-update` passed 12/12 at `20261004T054154Z-p95826`, promoting only the thirteen Reference Pack goldens below and the maintained manifest. Each promoted PNG is byte-identical to the corresponding actual image inspected through artifact-mode UI review. The renderer, fonts, viewport, zoom, density, masks and screenshot scopes are unchanged.

Affected semantic owner row: `web.design.visual.reference_pack_administration`, scenario `scenario_034f30c7dcac`. This is an active nonregistry visual scenario, so it has no separate stable registry fixture ID; its thirteen exact capture intents are the filename stems below. Ordinary-run reconciliation accounts for them through their catalog/scenario/capture IDs. It has zero orphan, ambiguous, missing or unresolved fixture mappings. No invented fixture ID or filename-based ownership was used.

Changed files under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`:

- `reference-pack-catalog-linux.png`
- `reference-pack-selection-linux.png`
- `reference-pack-submitted-linux.png`
- `reference-pack-recovery-short-linux.png`
- `reference-pack-recovery-zoom-linux.png`
- `reference-pack-recovery-spacing-linux.png`
- `reference-pack-queued-linux.png`
- `reference-pack-running-linux.png`
- `reference-pack-observation-recovery-linux.png`
- `reference-pack-cancel-requested-linux.png`
- `reference-pack-canceled-linux.png`
- `reference-pack-succeeded-linux.png`
- `reference-pack-failed-linux.png`

The separate Evidence anchor repair belongs to `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`, scenario `scenario_2bff3f0ed45f`, capture intent `evidence-timeline-evidence-count`. It passed 11/11 at `20261004T054018Z-p58587` with its existing golden; no Evidence image was refreshed. The repaired helper establishes the declared left edge for the sticky-gutter anchor before capture and is also exercised by the other evidence-action scenario in the full visual group. Fresh finalization passed 1/1 at `20261004T054953Z-p37655`. Two ordinary visual comparisons remain pending, with the second scheduled inside the final release run. `RESULTS_DIR` is unset, so retained-run maintenance was skipped.

The first fresh ordinary visual run passed **12/12** at `20261004T055013Z-p41727` in 436.454 seconds. Reconciliation again passed all 255 captures, with no missing, ambiguous, orphan or unresolved mapping. The second ordinary visual run is part of the final full `make release-check`, started after this pass with the same promoted manifest.


### Functional clock-driving follow-up

The next full release attempt, `20261004T055737Z-p82919`, was deliberately canceled after the timeline auto-scroll scenario exhausted its unchanged 60-second overall test budget. The trace shows 13.002 real seconds spent advancing 2 simulated seconds and 27.055 real seconds advancing another 8 simulated seconds; the test expired while advancing the horizontal interval, before its final boundary assertion could inspect the page. Eighteen other tests in that browser group passed. This is retained as a failed/canceled run, not successful release evidence.

The functional test now advances one 16 ms frame at a time, stops immediately when each declared scroll boundary is reached, and retains a bounded maximum frame count. It no longer executes several seconds of unrelated idle frames after reaching an edge. Production auto-scroll, production deadlines, the overall test timeout, exact selected membership, focus, query-count and document-scroll assertions are unchanged. This repairs the incomplete earlier fixed-interval clock test approach; focused and full reruns are required. The accepted visual manifest is unchanged.

The canceled run totals were **437 passed, 1 failed, 821 canceled**, duration 405.639 seconds. Format/generation passed at `20261004T060502Z-p95740` and `20261004T060509Z-p1046`. The repaired functional scrolling selector passed **11/11** at `20261004T060525Z-p3967`; no assertion or timeout was removed or weakened. Finalization and the complete release gate follow sequentially.

The final editorial version review also corrected Core 00's capability prohibition to name Extensions contract major 2, matching the existing Extensions owner. Required capability arrays remain empty. The exact Core 00 and Reference Pack content revision bindings were refreshed manually; no runtime or generated artifact consumes those Markdown bytes.


### Undo-continuity fixture preparation

The following full release run, `20261004T060707Z-p45314`, passed the repaired timeline range group but found an independent fixture precondition error in the existing auto-resolution Undo scenario. The test never established its initial grid offset and always wheeled upward by 120 pixels; depending on the preceding relationship-authoring layout, that could validly clamp to zero before the test checked that it remained positive. Fourteen other scenarios in that group passed. The run was canceled after diagnosis to repair the fixture and restart validation; it is not successful release evidence.

The scenario now establishes a middle scroll position before Undo captures its owner state, verifies that the position is scrollable, then dispatches a real native wheel movement for half the observed offset while the request is held. It still requires a positive newer position, a decrease from the initial position, BODY focus, one request and exact preservation after late acceptance. Product scrolling/focus logic, the timeout and comparison tolerances are unchanged. This is a test-preparation repair; the accepted visual manifest remains unchanged.

Canceled-run totals: **1154 passed, 1 failed, 104 canceled** in 1336.966 seconds.

The combined range-scrolling and auto-resolution owner selection passed **12/12** at `20261004T063044Z-p89867`. A final setup refinement waits for two animation frames after the initial programmatic scroll, before Undo admission, so setup cannot be mistaken for post-admission user intent. Format/generation passed at `20261004T063334Z-p30649` and `20261004T063341Z-p35615`; the focused follow-up and final release run remain pending.

The final Undo fixture follow-up passed **11/11** at `20261004T063352Z-p38466` in 125.264 seconds. The production tree and accepted golden manifest are frozen for fresh finalization and the complete release rerun.

### Completed pre-production handoff — 2026-10-04

Fresh `make agent-finalize` passed at `20261004T063610Z-p76079`. The final complete `make release-check` passed **1259/1259** at `20261004T063630Z-p80139`, including both repaired Timeline browser groups and the second ordinary visual comparison. Its source and golden identities, complete counts and phase exits are recorded in the current sections above. This closes the earlier pending implementation and validation statements for the authorized pre-production scope. The prior failed/canceled runs remain diagnostic history.

Final handoff Markdown lint passed at `20261004T071627Z-p84809`; its follow-up after the ledger and owner-inventory result additions passed at `20261004T071911Z-p87560`. `git diff --check` passed. This final record adds only the successful follow-up result. Retained-run maintenance was skipped because `RESULTS_DIR` is unset. No commit, push, real deployment reset or production conformance claim was made.
