# Appendix E: Roadmap, Open Questions, and Decision Backlog

This appendix is **non-normative**.

It preserves roadmap, historical open-question material, and editorial-backlog context from the exploratory source artifact for the current normative core.

## E.1 Editorial decision backlog

The source artifact mixed current requirements, roadmap positioning, and open questions in a few places. The current normative core closes current-profile contract decisions. This appendix now serves three non-normative purposes: preserving historical source-question context, recording future-roadmap material, and tracking bounded editorial-hardening backlog. Closed rows below point only to current owner sections and future-only follow-on scope. Appendix H carries separate non-normative operating-model guidance for operator practice.

| Decision ID | Topic | Current owner section(s) | Closure status | Future-only scope |
| --- | --- | --- | --- | --- |
| `DQ-002` | File-based structured import closure | Core 01 §2.1 and §17.2; Core 03 §11; Core 04 §9.2 | Resolved in core. | Future work MAY add richer parser coverage or operator conveniences without widening the base workbook surface. |
| `DQ-003` | Immutable snapshots and report exports status | Core 01 §10; Core 04 §2.1 and §9.3 | Resolved in core. | Future work MAY add delivery or authoring conveniences without changing snapshot-bound release controls. |
| `DQ-004` | Reference-pack refresh and distribution status | Core 01 §11; Core 04 §4.1 and §9.4 | Resolved in core. | Future work MAY add mirroring, signing, or distribution tooling without changing the disconnected-pack contract. |
| `DQ-005` | Incident portability contract | Core 01 §12.3; Core 04 §4.2 and §9.11 | Resolved in core. | Future work MAY add migration helpers or transport wrappers without changing the current portability contract. |
| `DQ-042` | Operational backup and coherent restore contract | Core 01 §12.1-§12.2; Core 04 §12 and §9.14 | Resolved in core. | Future work MAY add offsite replication or broader restore profiles without reinterpreting the base recovery contract. |
| `DQ-006` | Restricted evidence visibility versus export-scoped withholding | Core 01 §10.5; Core 04 §2 and §4.2 | Resolved in core. | Future work MAY add narrow live sensitive-evidence controls only through an explicit later profile. |
| `DQ-007` | Saved-view scope and startup/default selection | Core 01 §3.3.5.2 and §6; Core 03 §2.3-§2.4 and §14; Core 04 §9.10 | Resolved in core. | Future work MAY add curation, seeding, or suggestion tooling without changing scope or startup/default semantics. |
| `DQ-008` | Generated presentation depth | Core 01 §10; Core 04 §2.1 and §9.3 | Resolved in core. | Future work MAY add richer presentation families without relaxing the current evidence-versus-presentation boundary. |
| `DQ-009` | `task_request` and `decision` legal transition matrices | Core 02 §10.4.1.1 and §10.4.4; Core 03 §16.4; Core 04 §9.9 | Resolved in core. | Future work MAY add workflow presentation or convenience behavior without changing base lifecycle semantics. |
| `DQ-010` | Public interface surface contract | Core 01 §3.3 and §3.3.10.1; Core 04 §9.10 | Resolved in core. | Future work MAY add additive route families or messages only if the current major-version contracts remain intact. |
| `DQ-039` | Definition of Done separation: implementation conformance vs corpus-maintenance vs claim publication | Core 00 §1 and §4; Core 04 §9; Core 05 §1-§4 | Resolved in core. | Future work MAY refine benchmark-publication tooling or editorial linting without reintroducing those concerns into implementation conformance. |
| `DQ-020` | Extension-surface parity | Core 01 §17 and §20; Core 04 §9.2-§9.5 and §9.11 | Resolved in core. | Future work MAY expand claimed extension families additively without changing current route ownership or durable-state contracts. |
| `DQ-032` | Enterprise-auth identity-binding lifecycle | Core 01 §20; Core 04 §1.2 and §9.5 | Resolved in core. | Future work MAY add helper tooling or bulk provisioning that compiles to the same binding semantics. |
| `DQ-033` | Destructive-operation concurrency and public contention failure | Core 01 §3.3.5.0; Core 03 §5; Core 04 §9.1 | Resolved in core. | Future work MAY widen the destructive-operation family only through an explicit later profile that preserves fail-fast contention behavior. |
| `DQ-037` | Local pending queue behavior and persistence boundary | Core 03 §4.2 and §4.4; Core 04 §9.1 | Resolved in core. | Future work MAY add durable local draft persistence as a separate feature or profile. |
| `DQ-038` | History retention and rollback horizon | Core 01 §3.3.4.2 and §12.3; Core 02 §15.3.1; Core 04 §9.11 | Resolved in core. | Future work MAY define history purge or hard-delete only through an explicit later profile. |
| `DQ-021` | Row-create contract completeness | Core 01 §3.3.5 and §7.4; Core 04 §9.0 and §9.1 | Resolved in core. | Future work MAY specialize create-time policy per view without changing base replay and owner rules. |
| `DQ-022` | Incident-scoped party model | Core 02 §10.4, §18, and §19; Core 03 §16.4 and §20; Core 04 §2 and §9.1 | Resolved in core. | Future work MAY add party merge or broader dedupe only through an explicit later profile. |
| `DQ-011` | Blob-slot create contract completeness | Core 01 §3.3.8 and §16; Core 04 §9.10 | Resolved in core. | Future work MAY add integrity hints or convenience behavior without changing two-step attachment semantics. |
| `DQ-012` | Session lifecycle boundaries | Core 04 §1.1.1; Core 01 §3.3.2.1 | Resolved in core. | Future work MAY refine session UX or diagnostics without changing current expiry and revocation boundaries. |
| `DQ-013` | Deployment-local user-account and incident-membership administration | Core 01 §3.3.2, §3.3.5.1, and §18; Core 04 §2, §3, and §9.10 | Resolved in core. | Future work MAY add administrative tooling without changing the deployment-local versus incident-scoped split. |
| `DQ-014` | Collection-review mutation wire contract | Core 01 §3.3.5 and §7.4; Core 03 §3.3 and §13; Core 04 §9.6 | Resolved in core. | Future work MAY add collection families or field bindings without changing the current wire family. |
| `DQ-015` | Timeline `capture_state` transition triggers | Core 03 §6; Core 04 §9.9 | Resolved in core. | Future work MAY add lifecycle presentation or diagnostics without changing current capture-state triggers. |
| `DQ-016` | Hosts, Identities, and Evidence per-field view-schema registries | Core 01 §7.4 and §19; Core 03 §20; Core 04 §9.1 | Resolved in core. | Future work MAY add fields or views additively without changing current per-field contracts. |
| `DQ-017` | Incident create contract completeness | Core 01 §3.3.5.3; Core 04 §9.10 | Resolved in core. | Future work MAY add incident fields or helper flows without changing current create-route ownership. |
| `DQ-018` | Entity-mention explicit action route contract | Core 01 §3.3.5; Core 02 §6 and §7; Core 03 §9 | Resolved in core. | Future work MAY add suggestion or bulk-resolution tooling without changing the explicit-action route contract. |
| `DQ-019` | `text_compare_merge` operational semantics | Core 03 §3.3.3.1; Core 04 §9.6 | Resolved in core. | Future work MAY add presentation refinements without changing line-based same-field merge semantics. |
| `DQ-022` | Session resource shape | Core 01 §3.3.2.1; Core 04 §1.1.1 and §9.10 | Resolved in core. | Future work MAY add additive safe fields without changing the current session resource contract. |
| `DQ-023` | Writable-string normalization matrix | Core 01 §18; Core 04 §9.0 and §9.1 | Resolved in core. | Future work MAY add writable-string contracts or bindings additively without changing current normalization rules. |
| `DQ-024` | Coordination and optional artifact-backed surface closure | Core 02 §10.1 and §10.4.4A-§10.4.6; Core 01 §7.3-§7.4 and §19; Core 03 §2.1-§2.2 and §16.4; Core 04 §9.1 | Resolved in core. | Future profiles MAY promote one or more coordination surfaces to first-class record types only through an explicit compatibility-preserving profile or later NLSpec. Supporting guidance covers operator practice such as tracker hygiene, handoff quality, status-review cadence, debrief discipline, workload redistribution, and challenge/escalation practice. |
| `DQ-025` | Deployment configuration surface | Core 04 §12; Core 01 §14; Core 04 §9.12 | Resolved in core. | Future work MAY add deployment controls or packaging examples without changing the current config-artifact contract. |
| `DQ-026` | Direct-scalar timestamp write contract | Core 01 §18A; Core 03 §13.1; Core 04 §9.0 and §9.1 | Resolved in core. | Future work MAY add temporal contracts or bindings additively without changing current clear semantics. |
| `DQ-027` | Local-account login identifier | Core 01 §3.3.2 and §3.3.5.1; Core 04 §1.1 and §9.10 | Resolved in core. | Future work MAY add distinct usernames or login aliases only through an explicit profile or major-version change. |
| `DQ-028` | Direct-reference scalar write contract | Core 01 §18B; Core 04 §9.0 and §9.1 | Resolved in core. | Future work MAY add direct-reference contracts or bindings additively without changing current clear semantics. |
| `DQ-029` | Timeline supersede replacement relation | Core 01 §3.3.5 and §7.4.1; Core 02 §12.1-§12.3; Core 03 §6; Core 04 §9.11 | Resolved in core. | Future work MAY add derived terminal-successor or reviewer affordances without changing the direct replacement relation. |
| `DQ-030` | Define-once create-time defaults and initial-value ownership | Core 00 §5.1; Core 01 §7.4 and §19 | Resolved in core. | Future work MAY add corpus-lint or editorial tooling, but current owner-scoped defaults are closed. |
| `DQ-031` | Binary AC wording and closed allowlists | Core 04 §9; owner sections as applicable | Resolved in core. | Future work MAY add authoring or lint tooling without moving behavioral meaning into appendices. |
| `DQ-034` | View-query sort and group contract completeness | Core 01 §3.3.4 and §7.4; Core 03 §14; Core 04 §9.1 | Resolved in core. | Future work MAY add multi-level grouping or richer sort controls only through an explicit later profile or major-version change. |
| `DQ-035` | Saved-view object grammar and public `view_schema` resource shape | Core 01 §3.3.5.2, §6, and §7.4; Core 03 §2.3 and §14; Core 04 §9.1 and §9.10 | Resolved in core. | Future work MAY add discovery metadata or layout capabilities additively without changing current resource shapes. |
| `DQ-036` | Filter operand canonicalization and full-text token semantics | Core 01 §3.3.4.1; Core 04 §9.10 | Resolved in core. | Future work MAY add phrase, fuzzy, or language-aware search only through new operators or later profiles. |
| `DQ-040` | Manual link confidence default | Core 01 §7.4; Core 02 §12.3; Core 03 §13.1; Core 04 §9.1 | Resolved in core. | Future work MAY add scored manual-link workflows only through an explicit later profile. |
| `DQ-041` | Deterministic post-merge identifier retention for host and identity reuse | Core 01 §3.3.3 and §3.3.5.4; Core 02 §7.3, §8, §9, and §15.4; Core 03 §5 and §16.2; Core 04 §9.1 | Resolved in core. | Future work MAY add broader reusable-identifier administration or party-merge semantics only through an explicit later profile. |
### E.1.1 Editorial sweep rule

The owner/default sweep described in this subsection has been applied to the current corpus. The bullets below now remain as retained editorial-lint criteria for future edits and corpus-maintenance checks.

- In normative core docs, phrases such as `defaults to`, `when omitted on create`, and `server-generated when omitted on create` should appear only in the Core 01 owner section for the affected surface, except in Core 04 acceptance criteria.
- Core 02 should state field minima, closed vocabularies, immutability-after-first-commit, lifecycle guards, and server-managed post-commit fields, but not omitted-on-create defaults or initial-value policy.
- Unnumbered non-owner normative prose that restates owner-held create policy should be removed or replaced with an owner reference.
- Core 04 owner prose must not use open-ended equivalence language when the owner section can name the exact acceptable audit or observable shape; any Core 04 §9 criterion that verifies that owner prose should reference the closed owner wording rather than restating an open class.
- Core 04 §9 acceptance criteria must not use open-ended equivalence language. Any intentionally allowed alternative observable outcomes must be enumerated by exact identifiers in the owning Core 01, Core 02, or Core 03 section and referenced from the acceptance criterion.

### E.1.2 Binary AC hardening crosswalk

| Acceptance criterion | Owner section(s) | Exact closure applied |
| --- | --- | --- |
| `AC-035` | Core 01 §11.3.1 and Core 01 §3.3.6.2 | Replaced open-ended integrity-mismatch prose with `error.code='reference_pack_verification_failed'` plus the exact `reason_code` allowlist. |
| `AC-056` | Core 01 §10.2 and Core 02 §10.5 | Replaced `or equivalent frozen source boundary` with the exact snapshot-boundary field `source_change_set_high_watermark`. |
| `AC-074` | Core 02 §10.2 | Replaced `or an equivalent stable canonical identity` with incident plus deterministic type-specific dedupe key. |
| `AC-085`, `AC-112` | Core 01 view-schema registry and Core 03 workbook interaction | Replaced `blank-row or equivalent grid-native` phrasing with the exact owner term `inline create from the sheet itself` and preserved no-modal workbook-surface behavior. |
| `AC-093` | Core 01 §11.3 | Closed reference-pack ingress to exactly two entry points: configured storage root placement or `POST /api/v1/reference-packs/import`. |
| `AC-103` | Core 01 §3.3.6.2, Core 02 §13, and Core 03 §8.3-§8.4 | Replaced umbrella wording with the exact evidence-access `reason_code` registry and exact blocked-state behavior. |
| `AC-108`, `AC-110` | Core 04 §9.9 | Replaced generic prose with matrix-backed fixtures keyed by closed `machine_id` values and exact expected outcomes. |
| `AC-144` | Core 02 §10.4.2.1 | Replaced `equivalent computed indicator` with exact read-side `decision.is_superseded=true` while keeping persisted `status='executed'`. |
| `AC-156` | Core 04 §1.1.1 and Core 01 auth routes | Replaced `401 or equivalent session_expired error` with exact `401` failure behavior on the next request after idle expiry. |
| `AC-168` | Core 01 §12.3.5 | Replaced `equivalent historical actor descriptors` with the exact two-case imported-actor or historical-descriptor outcome. |
| `AC-190` | Core 03 §9 and Core 03 §14 | Replaced `equivalent unresolved-resolution queue views` with exact `timeline.has_unresolved_mentions=true` filter and grouping behavior. |
| `AC-250` | Core 01 §3.3.2 | Replaced provider-specific assertion catch-all prose with the exact forbidden-member set and unknown-top-level-member rejection rule. |
| `AC-252` | Core 03 §8.4 | Replaced `equivalent same-surface region` with exact same-workbook-surface no-full-page-navigation behavior. |
| `AC-278` | Core 03 §20 | Replaced `equivalent same-surface enrichment flow` with the exact closed set `inspector or Parties view`. |
| `AC-352` | Core 01 §3.3.5.1, Core 01 §20, and Core 04 §3 | Replaced `or equivalent secret-bearing protocol material` with the exact closed forbidden set in `REQ-04-096` and bound returned safe user resources plus `auth_bindings[]` summaries to the exact owner-defined member sets in `REQ-01-115` and `REQ-01-116`. |
| `AC-404` | Core 01 §1 | Replaced `and equivalent network-adjacent infrastructure` with the closed allowlist `reverse proxies`, `TLS terminators`, `managed ingress`, and `load balancers`. |
| `AC-405` | Core 01 §4.1-§4.3 | Replaced `or equivalent structured-store export` with `a Postgres logical backup or logical dump of the authoritative structured data store`. |
| `AC-406` | Core 02 §5 | Replaced `source-preserving text, or equivalent raw capture` with the exact rough-capture vocabulary `unresolved host or account text`, unstructured `details`, and unstructured `source_text`. |
| `AC-408` | Core 02 §14.3 and Core 04 §3 | Replaced `before/after values or equivalent patch data` with a reference to the exact mutation-target family enumeration in `REQ-04-037`, keyed to Core 02 §14.3. |

Rationale remains non-normative. The normative closure is that `decision.is_superseded` stays visible while persisted `decision.status` remains `executed`, and that `source_change_set_high_watermark` is the only current-profile frozen source boundary term.

### E.1.3 Repo-local editorial gate placeholder

The current corpus also needs a repo-local editorial gate. Filename and CI location remain `TODO:` until repository conventions are available.

Until those conventions are supplied, the gate contract is:

- fail when owner-only create-default phrases appear outside Core 01 owner sections and allowed Core 04 verification text,
- fail when Core 04 owner prose or Core 04 §9 acceptance criteria use open-ended `equivalent` or `similar` wording where the owner section can name an exact acceptable audit or observable shape,
- fail when any Core 00 §5.1 contract-owner-matrix row leaves `Requirement ID`, `Profiles`, or `Verified by` blank,
- fail when appendix text states current-profile runtime behavior without an exact owner-section restatement or an explicit non-normative boundary,
- fail when a Core 04 acceptance criterion is hardened to close open-ended wording but Appendix E §E.1.2 is not updated with the corresponding crosswalk row.
- fail when any of Core 01 §3.3.2, §3.3.4, §3.3.5, §17, or §20 lacks an owner-local route-index table for its dense public route family,
- fail when a dense public route family in those owner sections lacks owner-local compact tables for request and omission behavior, success and durable-resource members, and error or reason mapping,
- fail when an appendix example or preservation extract becomes the only place that states request shape, omission or explicit-`null` semantics, route-scoped idempotency, or family-specific terminal success mapping for one of those owner families,
- fail when a row-create or patch quick-reference table duplicates per-field create defaults or create-time writeability already owned by Core 01 §7.4 or §19 instead of referencing those registries.

This gate is corpus-maintenance support only. It does not define implementation conformance.

## E.2 Source roadmap extract

## 15. Recommended MVP and phased roadmap

### MVP

The MVP must include the spreadsheet-like UX on day one, or it is not an MVP for this problem.

Must-have on day one:

- browser workbook with Timeline, Hosts, Identities, Evidence, Notes tabs
- contract-backed system views for indicators and compromise assessments, with canonical indicators modeled as first-class records, source-bound indicator observations, and incident-scoped lifecycle intervals while preserving raw source text in Timeline, Notes, and Evidence
- inline grid editing with keyboard navigation
- multi-row clipboard paste
- blank-row quick entry
- unresolved mention capture for hosts and identities
- canonical host/identity records and manual resolution flow
- incident-scoped compromise assessment history for hosts and identities
- screenshot/file attachment to rows and standalone evidence requests
- typed links and tags
- data-driven type registries for host/evidence presentation, with local overrides
- row-centric history, attribution, and rollback with single-entry and whole-change-set restore paths
- multi-user live updates and presence
- local auth with TOTP MFA
- background jobs for evidence processing and other non-hot-path work required by the base profile, with grid editing and row creation remaining responsive while those jobs run
- container deployment with Postgres + MinIO

### Phase 2

- OIDC integration
- richer host/identity merge workflows
- richer saved-view curation, seeding, and suggestion tooling beyond the base scope and startup-default contract
- richer workflow presentation, automation, curation, and operating-model conveniences on top of the already standardized `task_request`, `decision`, `comm_log`, `handoff`, `status_review`, and `lesson` workbook surfaces
- stronger evidence previews and metadata extraction
- incident import assistant for XLSX files, prioritizing Timeline, Systems/Hosts, Accounts/Identities, Indicators, Evidence Tracker, and VERIS-like sheets when present; preserving unknown columns in `raw_capture` or `custom_attrs`
- immutable incident snapshots and self-contained report exports with stable identifiers, versioned recipient-specific redaction profiles, and supported templates
- integrity-checked ATT&CK/D3FEND/VERIS/reference-pack distribution for disconnected deployments
- dedicated framework workbook-surface extension profiles if the product later standardizes ATT&CK, D3FEND, VERIS, or similar workbook surfaces, with per-framework or tightly bounded-family `view_schema_id`, field registry, required pack keys, startup eligibility, filter/sort/group semantics, writeability, degradation behavior, and export/redaction behavior
- narrow live sensitive-evidence access controls only if repeated real-world incidents show export-scoped withholding is insufficient

### Phase 3

- SAML support
- richer portability tooling, compatibility helpers, and convenience packaging around the Incident Portability Extension Profile
- duplicate detection and resolution suggestions
- cross-record correlation helpers and graph-oriented exploration
- promote `hypothesis` to a first-class `record_type` only if repeated use shows artifact-backed tracking is insufficient for competing explanations, support or contradiction sets, state transitions, or reviewer-visible history
- richer presentation outputs built from immutable snapshots, with a clear boundary between source evidence and generated narrative material
- optional integrations for ingesting external exports

What can safely wait: advanced merge suggestions, enterprise SSO, richer importers, live restricted-evidence policies, and cross-incident analytics. What cannot wait: the grid, paste, evidence attach, progressive structuring, history, collaboration, and the contracts that keep workbook views, overlays, and exports aligned.


## E.3 Historical source open-questions extract

Remaining current-profile open questions:

None in the current profile. This includes pack-dependent framework workbook surfaces, which are closed in the current profile by prohibition unless a later profile standardizes them. History retention and rollback horizon are now explicitly closed by Core 01 §3.3.4.2 and §12.3, Core 02 §15.3.1, and Core 04 §9.1 and §9.11.

## 16. Source open-questions extract

Resolved in this revision:

The bullets below are archival closure summaries of source questions. They do not create new normative requirements; the cited core sections remain authoritative.

- Typed host/account alias auto-resolution is no longer open for MVP. Core 03 §12 now closes owner-scoped eligibility with one derived `auto_resolution_candidate_text`, `mention_token_text_v1` normalization followed by locale-independent Unicode case folding for comparison, a closed suppressor grammar (`?`, `~`, `maybe`, `prob`, `probably`, `approx`, `approximately`), explicit forbidden rewrites, explicit negative write effects that preserve unresolved mentions and create no `provenance='auto_match'` link, and negative conformance cases for suppressor and forbidden-rewrite inputs. Existing same-surface disclosure, direct undo, and reversible history remain part of the success path. Multilingual or locale-specific uncertainty lexicons remain future-only and require a later explicit profile or NLSpec.
- History granularity and rollback target selection are no longer open for MVP. Core 01 and Core 02 now require history at `change_set` plus mutation-target granularity. Core 01 now fixes `GET /api/v1/records/{record_id}/history` as the source of stable rollback selectors, including opaque `history_entry_ref` values for individually reversible logical history items, plus `change_set_id`, `revision_no`, `reversible`, and `available_rollback_actions[]`. `POST /api/v1/records/{record_id}/rollback` now uses a typed `target.kind` contract with `history_entry`, `change_set`, and `row_restore`; the reviewer UI remains row-centric, exposes single-entry rollback plus whole-row restore and whole-change-set rollback, and does not require arbitrary field-picker rollback from historical snapshots in MVP.
- History retention and rollback horizon are no longer open for the current profile. Core 02 now defines retained history for an extant incident record as the full authoritative history substrate needed to materialize row history and rollback semantics, retains that history for the full life of the incident record within the current deployment, and defines no age-based, closure-based, per-record, or operator-configurable purge workflow. Core 01 now makes `/history` visibility and `history_entry_ref` stability independent from current rollback eligibility, and the Incident Portability Extension Profile now preserves logical history and rollback scope without requiring byte-identical `history_entry_ref` values across deployments.
- Same-field conflict resolution affordance is no longer open for MVP. See Core 03 §3.3 and Core 04 AC-037 through AC-042. Same-field conflicts now resolve through a same-surface compare drawer or equivalent same-surface panel anchored to the conflicted cell, with the saved value retained in the grid, the unsaved local draft kept client-local until explicit resolution, contract-declared `conflict_resolution_class`, and grouped paste conflict handling.
- The public client/server interface surface is no longer open for the current profile. Core 01 now fixes a versioned HTTP+JSON API rooted at `/api/v1/` plus the canonical base-profile WebSocket subscription route `GET /ws/v1/incidents/{incident_id}` and the bounded v1 collaboration stream carried on that route, including a concrete v1 collaboration wire contract with a common envelope, typed `hello` and `resume` handshakes, typed presence payloads, replayable-versus-ephemeral delivery classes, bounded resume and reset behavior, application-level heartbeats, and per-incident ordering; Core 03 aligns same-field conflicts, presence, evidence upload, evidence access, and Timeline read or write flows to that contract; Core 04 fixes the session-based authentication surface, CSRF expectations, and conformance checks for query, mutation, conflict, blob-slot, job, and live-update behavior.
- The evidence-access handle surface is no longer split across route enumeration, UI text, and security criteria. Core 01 now owns request-body rules, response members, redemption, lifetime, revocation, filename and disposition, and route-specific error codes; Core 02 owns exact token spellings; Core 03 owns same-surface blocked-state UI; and Core 04 owns authorization re-derivation and active-content controls.
- The incident create contract is no longer implicit or split across retrieval and patch semantics. Core 01 now fixes `POST /api/v1/incidents` as the authoritative create contract, including required and optional request members, explicit create-only versus patchable incident fields, server-managed initial values, creator bootstrap side effects, normalized idempotency comparison, and explicit create-route error codes; Core 02 aligns persistence minima; Core 04 adds conformance checks; and Appendix C now makes clear which create-time validation rules remain service-layer obligations rather than raw SQL checks.
- Session lifecycle boundaries for authenticated HTTP and WebSocket access are no longer implicit. Core 04 now fixes a 30-minute qualifying-activity idle timeout, a 12-hour absolute lifetime, a 5-session human-user concurrency cap with least-recently-used non-current eviction, immediate current-session logout versus revoke-all triggers, and incident-membership loss that closes only the affected incident subscription. Core 01 now exposes `authenticated_at`, `idle_expires_at`, `absolute_expires_at`, and `session_expires_at` on `GET /api/v1/auth/session`, keeps `resume_token` as a replay-only token bound to `{session, incident, client_instance}`, and forbids heartbeats, passive push, reconnect or replay, and `GET /api/v1/auth/session` from renewing idle expiry; Core 03 preserves the client-local pending queue and same-field drafts across auth expiry until re-authentication; Core 04 adds conformance checks for timeout, revocation, and concurrent-session eviction. Core 03 also now closes the local pending queue boundary directly: same-runtime memory locality, exact `64`-unit capacity per `(incident_id, client_instance_id)`, FIFO replay, bounded coalescing, fail-closed overflow, explicit reload/crash loss in the base profile, and durable local persistence remaining future-only.
- Blob-slot create, replay, mismatch, and timeout behavior are no longer implicit. Core 01 now fixes required and optional `POST /api/v1/object-blobs` request members, omitted-versus-`null` handling for optional members, route-scoped idempotency keyed by `(actor_user_id, incident_id, client_txn_id)`, `accepted_contract` echo, `invalid_blob_create_request`, and same-request replay that returns the same expired slot rather than refreshing it; Core 02 adds accepted-contract and observed-object metadata, terminal size and expected-hash mismatch reasons, and structured expiry, failure, and cleanup metadata; Core 03 aligns the workbook flow with accepted-contract echo, lost-response replay, a 60-minute upload-target expiry, 24-hour pending-slot expiry, a single-upload lease, and terminal mismatch handling outside the ordinary 3-attempt retry budget; Core 04 adds conformance checks for malformed create requests, same-scope idempotency conflicts, expired-slot replay, terminal mismatch behavior, timeout, retry exhaustion, and orphaned-byte cleanup.
- Resource and security limits are no longer numerically open. Core 04 now owns a deployment-level resource-limit registry with exact defaults for blob-create ceilings, CSV and XLSX source-byte ceilings, workbook-shape ceilings, archive extracted-byte ceilings, compression-ratio and member-count ceilings, reference-pack and incident-bundle overrides, and preview ceilings. Core 01 now binds blob create, preview issuance, structured import, reference-pack import or refresh, and incident-bundle import to those keys with explicit acceptance or rejection behavior, while Core 04 adds conformance checks for exact-limit acceptance, one-byte-over rejection, oversized-preview blocking, archive-bomb rejection, and the distinct reference-pack versus incident-bundle extracted-byte overrides.
- The performance envelope for large-grid and evidence-heavy incidents is now closed in the normative core through an explicit split: Core 04 owns only implementation-facing thresholds and observable timed-state definitions, Core 05 owns claim-bearing benchmark reproducibility and publication policy, and corpus traceability plus spec-lint rules are editorial only. The earlier fixture-plus-`AC-043..AC-047` wording overstated closure because it mixed implementation outcomes with publication procedure.
- The remaining current-profile equivalence leak in performance and cursor-expiry conformance is no longer open. Core 04 now binds `perf.typing_ack.v1` to one exact Timeline summary typing acknowledgment surface and removes `equivalent deterministic test hook` language from AC-375, while Core 01 clarifies what `cursor_snapshot_unavailable` means when the cursor is otherwise valid.
- The timeline-sheet grouping-key whitelist is no longer open for the current profile. Core 03 now fixes the current five-key whitelist as sufficient for GA, forbids any sixth base-profile grouping key, and explicitly excludes raw `timeline.event_type` as a base-profile grouping key. Core 04 now makes that whitelist boundary explicit in conformance. See `### Timeline grouping-key whitelist` for the gate a future proposal must satisfy before widening the whitelist.
- Timeline `capture_state` transition triggers are no longer implicit. Core 03 now defines `capture_state` as a system-managed persisted Timeline lifecycle machine with `rough` on row creation, automatic server-side promotion to `enriched` on the first later `capture-state-material` mutation and on later material edits to a reviewed row, explicit reviewer-only `mark-reviewed` and `supersede` actions, and a terminal `superseded` state for ordinary workflow. Core 01 makes `timeline.capture_state` read-only to generic create and patch flows, adds dedicated record-scoped action routes, and requires future Timeline write surfaces to declare whether they are `capture-state-material`. Core 04 adds conformance checks for initial `rough`, review, supersede, reviewed-to-enriched regression, and independence from derived flags such as `timeline.has_unresolved_mentions`.
- The base-profile built-in and system view registry is no longer implicit. Core 01 now fixes fourteen mandatory pack-independent `view_schema_id` entries for Timeline, Hosts, Identities, Evidence, Notes, Indicators, Assessments, Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson, each with an ordered field set, deterministic default sort tuple, filter whitelist, and explicit write-back contract. Core 01 §7.4 and §19 now act as exhaustive per-field registries for those surfaces, including the new coordination-artifact addenda, while Findings, Investigative Queries, and Forensic Keywords now have standardized optional artifact-backed workbook surface contracts when exposed. Core 02 keeps exact-match reuse precedence, closed vocabularies, and artifact-backed storage authoritative for the relevant surfaces, and Core 04 AC-116 through AC-118 plus AC-281 through AC-287 make these registry contracts inspectable and fail closed on undeclared or read-only writes.
- The incident-specific custom-metadata promotion question is no longer open for the current profile. Core 02 now defines the promotion rule: a field leaves `custom_attrs` only when it recurs across independent SoD exemplars and is operationally load-bearing for sort, filter, join, lifecycle, dedupe, queueing, or reporting. The current profile now fixes the minimum promoted field sets for incident metadata, task requests, hosts, identities, canonical indicators, and evidence records; keeps framework overlays, enrichment metadata, presentation flags, and low-frequency governance text flexible; requires multi-valued relationships to use typed links or child tables rather than JSONB; and keeps authoritative custody narrative in append-only custody events rather than a single top-level text field.
- Snapshot and report release controls are no longer open for the current profile. Core 01 now fixes the immutable release tuple, `content_class`, `release_scope`, versioned template contracts, versioned redaction profiles, fail-closed redaction, `support_refs[]` on externally releasable curated narrative, and redaction manifests. The former advisory default for direct-source note and coordination-derived export text is now closed normatively in Core 01 as fail-closed `working_material` classification at export-model derivation time and verified in Core 04. Core 04 now defines the bounded artifact-release gate and approval invalidation rules.
- Generated presentation depth is no longer open for the current profile. Core 01 now allows arrangement and deterministic templating over snapshot facts, forbids invented facts or synthesized unobserved operator activity, allows `mermaid` and `slidev` external release only within the chosen release scope, and keeps `reenactment` outputs internal-review-only with `generated_presentation=true`.
- Whole-incident portability is no longer open for the current profile. Core 00 now adds the Incident Portability Extension Profile; Core 01 fixes the logical bundle layout, manifest, deterministic JSON/NDJSON bundle format, authoritative-source-only export boundary, portable actor descriptors, optional embedded snapshot/reference-pack sections, and fail-closed import phases; Core 04 adds the corresponding trust-boundary and conformance checks.
- Clipboard paste versus XLSX adoption is no longer open as a single yes-or-no question. The current core keeps clipboard paste on the base workbook interaction path and closes richer file-based onboarding as a Phase 2 Workbook Import Assistant inside the Import Extension Profile behind a dedicated imports module. Whole-workbook import is now defined as an orchestrated batch of explicit `import_unit` objects within one `import_session`; bounded CSV plus discovered XLSX used ranges, tables, eligible named ranges, and operator-selected regions are the current bridge for brownfield workbook migration, while workbook semantics remain downgraded metadata or rejection cases rather than first-class logic.
- Restricted evidence visibility inside a single incident workspace is no longer open for the current profile. Live workbook views remain incident-scoped for authenticated incident participants. Recipient-specific withholding is handled only at snapshot, render, and release time through versioned redaction profiles and optional `disclosure_partition_refs[]`. Future live sensitive-evidence controls remain out of scope unless repeated real-world incidents show that export-scoped withholding is insufficient.
- Shared/private saved views and workbook startup defaults are no longer open for the current profile. Core 01 now adds saved-view delete plus explicit saved-view and workbook-preference API contracts; Core 02 defines the saved-view schema contract plus separate per-user and incident-wide `sheet_ref` preference objects; Core 03 defines the `private`, `shared`, and `system` scope model, immutable `view_schema_id`, owner/admin mutability rules, and deterministic startup fallback from explicit `sheet_ref` to user home to incident default to Timeline; Core 04 adds authorization and conformance checks that saved-view scope affects only discoverability of the saved-view object, not underlying incident-data visibility.
- The dedicated Notes tab is no longer open for the current profile. Core 01 now fixes Notes as a built-in sheet keyed by stable `view_schema_id` and backed by `artifact_grid_projection` filtered to `artifact_type='note'`; Core 02 keeps note storage artifact-backed and links contextual notes through generic `record_links`; Core 04 adds acceptance criteria for direct and contextual note creation and for excluding raw note working material from `external_release`. The current core now separates analyst-work coordination objects from notes by using distinct `record_type` or `artifact_type` assignments rather than overloading `artifact_type='note'`. A future NLSpec MAY revisit whether Notes remains first-class only after those richer objects ship and adoption evidence shows the dedicated sheet is redundant; it is not a current-profile open question.
- Indicator storage promotion is no longer open for the current profile. Core 01 now fixes the Indicators system view as a projection over canonical indicator records and makes `cartulary.view.indicators.v1` explicit about existing-row mutability: there are no existing-row writable fields, the exact identity-defining immutable field set is `indicator.indicator_type`, `indicator.value_kind`, `indicator.display_value`, and `indicator.normalized_value`, plus `indicator.hash_algorithm` and `indicator.hash_value` when populated and used by the canonical dedupe key, and `indicator.stix_pattern` plus `indicator.defanged_value` remain create-only but non-identity fields. Core 02 requires first-class canonical indicators, source-bound `indicator_observation` rows, and append-only indicator lifecycle intervals; Core 03 keeps indicator capture embedded in raw source fields and places indicator linking in the same-surface enrichment flow; Core 04 adds acceptance criteria for distinct observations, canonical dedupe, lifecycle history, stable one-row-per-indicator system-view behavior, and fail-closed rejection of existing-row indicator rekey attempts.
- The minimum disconnected reference-pack bundle is no longer open for the current profile. Core 01 now fixes the smallest supported disconnected bundle as exactly three preinstalled type-registry packs, `type_registry.host`, `type_registry.evidence`, and `type_registry.indicator`; keeps framework, enrichment, template, and separately distributed `view_contract` packs outside that minimum bundle; and requires offline staged import plus explicit activation. Core 02 now requires structured activation-state and attestation metadata. Core 04 adds fail-closed verification, retained prior active version, no-live-fetch disconnected operation, and conformance criteria for the pack lifecycle.
- Assessment vocabulary and confidence model are no longer open for the current profile. Core 01 now fixes the Compromise Assessments system view as a projection over append-only assessment records; Core 02 fixes closed compromise-assessment states `unknown`, `suspected`, `confirmed`, `disproven`, and `cleared`, separates assessment judgment from response posture, and defines nullable `confidence_score` plus derived `confidence_band`; Core 03 makes interactive assessment entry band-first and keeps filtering separate by `assessment_state` and `confidence_band`; Core 04 adds acceptance criteria for distinct `disproven` versus `cleared` behavior and for keeping response actions separate from assessment state.
- The analyst-work tracking boundary is no longer open for the current profile. Core 00 now places `task_request`, `decision`, and structured coordination artifacts inside the current system boundary; Core 01 adds task and decision system views plus mandatory workbook-native coordination surfaces for communications, handoff, status-review, and lesson without adding built-in tabs; Core 02 defines the promotion rule, first-class `task_request` and `decision` records, ownership as a field on coordination objects, artifact-backed `comm_log`, `handoff`, `status_review`, and `lesson`, and current hypothesis tracking through `artifact_type='finding'` plus `finding.kind='hypothesis'`; Core 03 keeps routine timeline capture free of mandatory owner or approval fields while making the coordination surfaces field-closed and surface-closed in the base profile; and Findings, Investigative Queries, and Forensic Keywords are now standardized optional workbook surface contracts rather than vague future placeholders.

### Analyst-work tracking boundary

Historical source topic retained for lineage only.

Current authority: Core 03 §2.2 and §16.4; Core 02 §10.4.3-§10.4.5; Core 01 §7.4 and §19; Core 04 §2, §4.2, and §9.1.

Supporting guidance: Appendix H covers tracker hygiene, companion findings-document discipline, handoff quality, status-review cadence, debrief follow-through, workload redistribution, and challenge/escalation practice.

Future-only scope: later profiles MAY promote additional coordination surfaces or first-class escalation/risk objects only through an explicit compatibility-preserving profile or later NLSpec.

### Performance envelope for large incidents and evidence-heavy incidents

This subsection is historical rationale only. Core 04 now owns only implementation-facing thresholds and observable timed-state definitions, while Core 05 owns claim-bearing benchmark publication and reproducibility.

This revision closes the performance question with two reference fixtures rather than a single vague "large incident" threshold. That keeps grid pressure and evidence-path pressure separable, while leaving claim-bearing publication procedure to Core 05 instead of the implementation-facing core.

**Fixture A: large-grid incident**

- 20,000 timeline rows
- 1,000 host rows
- 1,000 identity rows
- 25 concurrently connected analyst sessions on one incident, with presence enabled and representative live row-update traffic
- representative tags, mentions, and links, but not evidence-heavy per row

**Fixture B: evidence-heavy incident**

- 5,000 timeline rows
- 10,000 evidence records
- tens of GB of binary evidence stored in object storage
- at least one timeline row linked to 100 evidence records
- Core 05 now allows stubbed evidence blobs for throughput tests while still requiring real evidence metadata, counts, attachment state, and preview handles

**Immediate hot path**

- selection move, focus move, and typing acknowledgment: at or below 100 ms p95
- blank-row creation in the timeline sheet with one non-empty user-entered value: at or below 150 ms p95
- row identity and selection remain anchored to `record_id` during live updates, filtering, sorting, and grouping

**Short-delay interactive**

- sort, filter, and grouping changes: first useful viewport at or below 250 ms p95
- full stable viewport and result ordering: at or below 1.0 s p95
- inspector metadata shell for a row linked to 100 evidence records: at or below 300 ms p95

**Staged/background**

- imports, evidence processing, projection rebuilds, snapshot generation, report generation, and reference-pack refresh remain background jobs
- progress and cancellation appear within 1 second of job start
- grid editing and row creation remain responsive while those jobs run

### Timeline grouping-key whitelist

This revision closes the timeline grouping-key question by keeping the base-profile whitelist at exactly five keys: `timeline.occurred_day`, `timeline.recorded_day`, `timeline.capture_state`, `timeline.has_evidence`, and `timeline.has_unresolved_mentions`. The current whitelist is sufficient for GA.

That whitelist is only the Timeline-specific part of the now-generic grouped-surface contract. Any workbook surface whose active `view_schema` declares `grouping_fields` now uses its declared key list in discovery order, omission-only `group_by` for `Group: None`, derived group headers from `group_values[group_by]`, and the generic comparator rule unless the Timeline-specific override applies.

Raw `timeline.event_type` remains out of scope for base-profile grouping because it would couple grouping behavior to evolving source vocabularies and imported taxonomies. If a later revision adds one more grouping key, the next candidate should be a derived bucket such as `timeline.event_family`, not raw `timeline.event_type`.

A future proposal to widen the base-profile whitelist is eligible only if all of the following gates are met:

- the candidate key is a scalar contract-backed value already available from `timeline_grid_projection` or an equivalent projection, with no scan of free text, no join against blob or evidence metadata, and no dependence on visible labels
- the candidate key uses a stable closed vocabulary or deterministic bucketization, not raw imported labels
- analyst testing shows that sort, filter, saved views, or dedicated derived views are materially worse for the target task
- the added key preserves the sort, filter, and grouping interaction envelope defined in Core 04

