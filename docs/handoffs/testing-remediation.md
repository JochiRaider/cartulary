# Testing Remediation — Hard Pre-production Cutover

Document date: 2026-10-24, America/New_York.

Status: WF-00–WF-04, WF-06 and WF-07 COMPLETE; human owner/interface approval recorded in E-026; WF-05/S-006 is IN PROGRESS after E-036. Execution completion remains open.

Scope: the audited 88 isolated browser rows, their supporting harness, and backend verification needed to preserve their production coverage.

Current action boundary: the user explicitly authorized implementation of the complete bounded cutover on 2026-10-24. That authorization includes the planned specification, implementation, test, projection, validation, and handoff work. Historical planning-session restrictions below describe those earlier sessions only. Existing unrelated work and independently owned service lifetimes MUST be preserved.

Normative terms: **MUST** and **MUST NOT** define implementation and acceptance obligations within this handoff's scope. **MAY** permits only the explicitly stated option and omission behavior. Paragraphs labeled **Rationale** are explanatory and MUST NOT override requirements. Existing owner requirements remain authoritative until a coordinated amendment is adopted.

## 1. Scope and Source Posture

The cutover MUST reduce test execution overhead while preserving production correctness, security, determinism, integration coverage, accessibility, and regression detection. Existing test bodies, titles, row counts, fixture arrangements, and private helper APIs are not preservation requirements.

The audit baseline is commit `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1` on `main`. Static inspection found 88 isolated browser rows selecting 97 cases across 26 files. The working tree was clean before handoff creation. No tests, builds, generators, or service operations were executed during the audit or handoff-writing session.

The repository root is `/home/jochi/code/cartulary`. The canonical output is `docs/handoffs/testing-remediation.md`. Before creation, the destination was absent and its existing parent resolved inside the repository without a symlink component. This handoff follows the [modular refactor planning framework](cartulary_modular_refactor_planning_framework.md).

**TR-REQ-001 — Authority.** This handoff MUST govern implementation sequencing and acceptance only. Lasting harness behavior MUST be specified in the adopted Testing Harness NLSpec and projected into authored machine contracts. Product behavior MUST remain governed by its adopted owner.

| Source | Authority and interaction |
|---|---|
| [Core 00](../spec/00_document_set_status_and_precedence.md) | Document precedence, profile boundaries, and primary-owner allocation. |
| [Domain vocabulary](../domain.md) | Vocabulary and owner navigation; not an independent runtime or test-plan authority. |
| [Testing Harness NLSpec](../testing-harness-nlspec.md) | Selection, resource ownership, cohorts, cleanup, accounting, and evidence. REQ-805/806 and REQ-678 are particularly relevant. |
| [Core 01](../spec/01_architecture_storage_and_view_contracts.md), [Core 02](../spec/02_domain_model_schema_and_history.md), [Core 03](../spec/03_workbook_interaction_collaboration_and_workflows.md), [Core 04](../spec/04_security_deployment_and_conformance.md), and adopted subsystem NLSpecs | Production interfaces, authorization, persistence, collaboration, recovery, and feature acceptance requirements. |
| [Core 05](../spec/05_claim_publication_and_benchmark_reproducibility.md) | Applies only when its claim-publication predicate is active. This remediation makes no benchmark-publication claim. |
| Earlier remediation handoffs | Historical context only. Their execution authorizations, performance thresholds, and completion claims MUST NOT be inherited. |

**TR-REQ-002 — Scope boundary.** The implementation MUST NOT broaden into general Go, Vitest, shell, scheduler, or production-feature redesign. Backend changes are limited to test composition and coverage required by browser restructuring. Production APIs, storage semantics, authorization rules, time semantics, and supported profiles MUST remain unchanged.

**TR-REQ-003 — Documentation boundary.** Tests, generators, runtime metadata, conformance, and release evidence MUST NOT read, stat, hash, or otherwise consume this handoff or other Markdown. Human coverage mapping and executable selection metadata MUST remain separate.

## 2. Current-State Repository Inventory

The following is the frozen baseline triage, not current runtime qualification. E-036 records the qualified successor inventory; historical labels below remain attributable to their audit source. Inspection covered selected scenarios and consequential helpers; it does not establish exhaustive inspection of unrelated cases in the same files.

| Inventory group | Current isolated rows | Current disposition |
|---|---:|---|
| Incident lifecycle, metadata, membership management, membership audit, and imports | 21 | Candidates for owned-state execution. |
| Authentication/directory, shell support, and administrative audit, excluding the last-admin scenario | 12 | Candidates; require authentication and shared-listing interference checks. |
| Collaboration, incident administration, preferences, extension preferences, grid closure, Timeline, and Evidence role changes | 20 | Candidates; secondary-context and asynchronous cleanup require qualification. |
| Reference Pack presentation and visual-harness data/preference checks | 4 | Candidates; distinguish local presentation state from real registry mutation. |
| Accessibility and visual scenarios excluding server-fault consumers | 20 | Candidates; preserve browser-native and visual assertions. |
| Clock-dependent scenarios | 4 | Restructure as specified in Section 7. |
| Last-admin browser scenario | 1 | Remove deployment-wide setup from browser execution. |
| Real Reference Pack lifecycle | 1 | Retain a compact dedicated integration workflow. |
| Server public-error-fault consumers | 3 | Replace presentation fault injection and retire the obsolete control endpoint. |
| Restore and concurrent frontend publication | 2 | Retain dedicated resource lifecycles. |
| **Total** | **88** | **77 candidates, 9 restructuring targets, 2 specialized lifecycles.** |

The implementation baseline MUST enumerate the exact 88 row IDs, their 97 selected cases, authored selectors, owning verification contracts, and directly used fixture helpers. Every baseline case MUST receive a disposition of retained, split, consolidated, or replaced, with its production obligations mapped to successors.

Enumeration MUST NOT silently expand to subsequently added rows. Changes made by the other development chat MUST be reconciled explicitly before implementation starts.

The baseline selection is the active Playwright rows declaring `browser_state_scope=isolated` in the authored family manifests at the audit commit. Current manifests MUST be compared with that frozen selection rather than treated as an interchangeable baseline. Generated projections and ignored run outputs were not inspected as behavioral authority. Retained runtime results remain a WF-00 prerequisite.

### WF-00 frozen case inventory

E-011: Authored family manifests were compared directly with Git commit `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1` and match. The following 88 rows select exactly 97 cases in 26 files. Each row binds the file, project, stage, scenario IDs and title selectors below. All rows start `isolated`; proposed dispositions do not authorize sharing. The default runtime applies unless `network_flow_claimed` is stated. Verification IDs refer to the matching authored file under `contracts/verification/owners/`; these are projections, not owner requirements. Selected assertions and consequential helpers are mapped in Section 10, including the multiple-case rows; baseline runtime characterization remains in progress. Candidate qualification is not inferred from baseline results.

#### B-001 — `harness.browser.boundary_support.frontend_artifact_lifetime`

- Source: `apps/web/e2e/visual-harness.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `harness.browser.verification.support_contract`.
- Disposition: **retained**; successor: S-005: dedicated publication lifecycle.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-023); candidate sharing/lifecycle qualification pending.
- Current qualification: **isolated** for the documented real expiry or specialized runtime obligation; prior scoped evidence is retained in E-028/E-030/E-034. Fresh final-suite acceptance remains S-006.

  - `scenario_1fc4932fab42`: visual harness retains loaded assets across an independent frontend publication

#### B-002 — `harness.browser.boundary_support.timeline_investigation_capture`

- Source: `apps/web/e2e/visual-harness.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `harness.browser.verification.support_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-023); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `harness.browser`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_timeline_investigation_capture`: rich Timeline recipe preserves owner states and source text during capture preparation

#### B-003 — `harness.browser.boundary_support.visual_preference_lifetime`

- Source: `apps/web/e2e/visual-harness.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `harness.browser.verification.support_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-023); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `harness.browser`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_8c56bafc97b7`: visual harness preferences remain local across density order and page closure

#### B-004 — `module.auth.accessibility.verify_session_mfa_incident_forbidden_loading_an_e8f5d5f3a9`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_24af174d2595`: a11y.incident-selection.row-01 anonymous login after initial session_required reaches login controls and authenticated landing
  - `scenario_518ccc2daf28`: a11y.incident-selection.row-01 authenticated landing exposes account, admin, incident, retry, and visible incident controls
  - `scenario_6b62ae5a8f33`: a11y.incident-selection.row-01 deferred session loading exposes progress and keeps recovery controls keyboard reachable
  - `scenario_7e9876d4d78d`: a11y.incident-selection.row-01 forbidden access-denied public envelope is announced and exposes recovery without private diagnostics
  - `scenario_9e9180c613d0`: a11y.incident-selection.row-01 generic public error envelope renders safe diagnostics and keyboard error recovery
  - `scenario_b631edcb35d1`: a11y.incident-selection.row-01 incident empty, list, selected, stale-selection, and incident-error states expose keyboard recovery
  - `scenario_c05759d45cdc`: a11y.incident-selection.row-01 mfa_required challenge is keyboard reachable, visibly focused, named, and safely announced
  - `scenario_d9203af9002f`: a11y.incident-selection.row-01 mfa_setup_required enrollment is keyboard reachable and public errors hide private setup diagnostics
  - `scenario_fcec515eb61b`: a11y.incident-selection.row-01 revoked session after prior authentication announces session end and supports re-authentication

#### B-005 — `module.auth.browser.a_reviewer_visible_incident_edit_control_fails_w_b6248bf4c8`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_5ef1cd3e0693`: observes current-role authorization on a stale reviewer edit through the public incident error envelope

#### B-006 — `module.auth.browser.administrative_audit_read_recovery`

- Source: `apps/web/e2e/administrative-audit.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_054c68190f9e`: administrative audit fences delayed pages and errors with keyboard filter recovery and focus continuity

#### B-007 — `module.auth.browser.administrative_audit_real_browsing`

- Source: `apps/web/e2e/administrative-audit.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_58e388dabe02`: administrative audit browses more than one hundred real exact matching events and timestamp boundaries

#### B-008 — `module.auth.browser.deployment_admin_revoke_all_invalidates_a_target_a09ed5d06f`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_cc16b99b2c52`: returns a revoked target browser to login and allows re-authentication with unchanged incident membership

#### B-009 — `module.auth.browser.the_ordinary_shell_keeps_deployment_user_adminis_8a57567508`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_55fc11b5a1fc`: keeps deployment-user administration on deployment-admin sessions and hides it from an incident admin on the ordinary shell

#### B-010 — `module.auth.browser.verify_ordinary_login_incident_directory_entry_o_d66ea23dc1`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_33e24429c797`: Verify ordinary login, incident entry, and current-role refresh stay on public browser routes.

#### B-011 — `module.auth.browser_stateful.browser_deployment_user_administration_works_on_d6d7ff0b86`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `stateful`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **split**; successor: S-002: owned browser administration and local last-admin presentation; real backend HTTP rejection/state/session witnesses.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_e2a04c9d54c7`: lets deployment admins create and patch users, rejects stale versions, and shows the last-admin guard on the ordinary shell

#### B-012 — `module.auth.browser_stateful.idle_expiry_fails_closed_after_the_shared_clock_6525f19d9d`

- Source: `apps/web/e2e/clock.spec.ts`; project `chromium`; stage `stateful`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-002: compact real HTTP expiry bridge.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **isolated** for the documented real expiry or specialized runtime obligation; prior scoped evidence is retained in E-028/E-030/E-034. Fresh final-suite acceptance remains S-006.

  - `scenario_0098c4cf976a`: advances the shared clock past idle expiry and requires a fresh login afterwards

#### B-013 — `module.auth.browser_support.supplemental_browser_verifies_ordinary_shell_sel_f8a9da1957`

- Source: `apps/web/e2e/auth.support.spec.ts`; project `chromium`; stage `support`; runtime `default`.
- Verification: `module.auth.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-012); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.auth`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_3925df4ef78d`: Verify bootstrap route selectors and error-state selectors use stable test-id builders.

#### B-014 — `module.collaboration.browser.concurrent_edits_to_different_fields_on_the_same_7ee5cad0ab`

- Source: `apps/web/e2e/collaboration.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.collaboration.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-015); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.collaboration`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_dad2a534f922`: auto-merges different-field concurrent edits and requires explicit same-field resolution

#### B-015 — `module.collaboration.browser.live_updates_never_retarget_pending_local_edits_11699e9b2d`

- Source: `apps/web/e2e/collaboration.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.collaboration.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-015); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.collaboration`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_7425a87e156f`: keeps live updates conflict markers and presence markers anchored to record_id and field_key

#### B-016 — `module.collaboration.browser.verify_conflict_resolver_actions_submit_public_m_0c77fae932`

- Source: `apps/web/e2e/collaboration.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.collaboration.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-015); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.collaboration`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_e36420c2e84a`: Verify conflict resolver actions submit public mutations and refresh rows without losing focus or pending queue ordering.

#### B-017 — `module.collaboration.browser_stateful.deterministic_route_owned_browser_revocation_sou_a134c3e26b`

- Source: `apps/web/e2e/session-recovery.spec.ts`; project `chromium`; stage `stateful`; runtime `default`.
- Verification: `module.collaboration.verification.behavior_contract`.
- Disposition: **split**; successor: S-002: independent owned revoke-all/logout cases and dedicated socket idle-expiry bridge.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-015); candidate sharing/lifecycle qualification pending.
- Current qualification: real socket-expiry successor remains **isolated**; new logout and revoke-all successors are **test_owned**, passed in E-035/E-036. E-030 maps pending-work/replay and backend obligations.

  - `scenario_f645b23950cc`: preserves unsaved local work after socket revocation and re-authentication

#### B-018 — `module.collaboration.browser_stateful.verify_multi_client_live_row_update_presence_anc_4c65a275c1`

- Source: `apps/web/e2e/collaboration.spec.ts`; project `chromium`; stage `stateful`; runtime `default`.
- Verification: `module.collaboration.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-015); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.collaboration`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_cd147372eead`: Verify multi-client live row update, presence anchoring, reset/invalidate handling, stale-row requery, and same-field conflict resolver through /ws/v1/ and /api/v1/.

#### B-019 — `module.evidence.accessibility.verify_evidence_icon_buttons_blocked_states_erro_0ec80f90e2`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `module.evidence.verification.behavior_contract`.
- Disposition: **retained**; successor: S-003: local public-error presentation; preserve real Evidence access/preview/download and accessibility assertions.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-016); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.evidence`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_bd0842a0bd8b`: a11y.evidence-workflow.row-01 Verify evidence icon buttons, blocked states, error states, preview controls, and download controls have names, focus, contrast, and non-color-only distinctions.

#### B-020 — `module.evidence.browser.timeline_related_role_loss`

- Source: `apps/web/e2e/timeline-related-evidence.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.evidence.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-016); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.evidence`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_9ba6d3e6be57`: Timeline Evidence rechecks a real role loss between commits and preserves the accepted record

#### B-021 — `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `module.evidence.verification.behavior_contract`.
- Disposition: **retained**; successor: S-003: local public-error presentation; preserve meaningful visual assertions.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-016); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.evidence`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_2bff3f0ed45f`: Capture evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures.

#### B-022 — `module.incidents.browser.a_selected_incident_whose_membership_is_removed_21054ae899`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_cb4fe9d1ef4c`: clears a stale selected incident after membership removal while preserving the account session

#### B-023 — `module.incidents.browser.incident_admins_manage_memberships_on_the_ordina_f26145a162`

- Source: `apps/web/e2e/incident-administration.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_40fddb90fe45`: lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell

#### B-024 — `module.incidents.browser.incident_creation_native_keyboard_payload_and_fi_b6b4631f5f`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_incident_creation_keyboard`: creates an incident with native keyboard submission and associated field errors

#### B-025 — `module.incidents.browser.incident_creation_preserves_chosen_navigation_af_24cb9bd4de`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_incident_creation_navigation`: retains a confirmed creation when account navigation precedes its delayed response

#### B-026 — `module.incidents.browser.incident_creation_uncertain_replay_and_committed_985931d51a`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_incident_creation_replay`: replays a committed incident after transport loss without duplicate dispatch or recreation

#### B-027 — `module.incidents.browser.lifecycle_acknowledgement`

- Source: `apps/web/e2e/incident-lifecycle.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_527e9b72ca42`: lifecycle confirms a real action independently of failed reads and preserves newer input during delayed acknowledgement

#### B-028 — `module.incidents.browser.lifecycle_boundaries`

- Source: `apps/web/e2e/incident-lifecycle.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_96fddba6266b`: lifecycle validates normalized reason boundaries and requires fresh review after a real version conflict

#### B-029 — `module.incidents.browser.lifecycle_departure`

- Source: `apps/web/e2e/incident-lifecycle.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_1414a3b14f43`: lifecycle metadata and membership retained work receive sequential departure review with cancellation and history

#### B-030 — `module.incidents.browser.lifecycle_replay`

- Source: `apps/web/e2e/incident-lifecycle.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_f9e8cae8d508`: lifecycle recovers a committed lost Close after another actor reopens without publishing its historical receipt

#### B-031 — `module.incidents.browser.lifecycle_workbook`

- Source: `apps/web/e2e/incident-lifecycle.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_bd1b5cd5fe8b`: lifecycle closure retains rejected workbook work and allowed configuration while reopening requires fresh source action

#### B-032 — `module.incidents.browser.membership_audit_access_recovery`

- Source: `apps/web/e2e/incident-membership-audit.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_0b61b135c791`: membership audit distinguishes temporary observation failure role downgrade hidden incident and session loss

#### B-033 — `module.incidents.browser.membership_audit_read_recovery`

- Source: `apps/web/e2e/incident-membership-audit.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_d8f8c46dbe24`: membership audit fences delayed reads across query drawer section and visibility with explicit cursor recovery

#### B-034 — `module.incidents.browser.membership_audit_real_browsing`

- Source: `apps/web/e2e/incident-membership-audit.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_da52a9fc1592`: membership audit browses 105 real exact role changes with accepted query pagination and instant boundaries

#### B-035 — `module.incidents.browser.membership_audit_response_inspection`

- Source: `apps/web/e2e/incident-membership-audit.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_da73fe10b909`: membership audit rejects malformed scope and preserves additive inert JSON inspection

#### B-036 — `module.incidents.browser.membership_management_departure`

- Source: `apps/web/e2e/incident-membership-management.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_b0e3c392d612`: membership management retains pending work on close and requires deliberate route departure

#### B-037 — `module.incidents.browser.membership_management_live_pages`

- Source: `apps/web/e2e/incident-membership-management.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_6f0db0ba0dc9`: membership management traverses live service pages and adds changes removes with audit visibility

#### B-038 — `module.incidents.browser.membership_management_mutation_recovery`

- Source: `apps/web/e2e/incident-membership-management.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_485c30a13c76`: membership management separates confirmed refresh failure exact create replay and explicit conflict review

#### B-039 — `module.incidents.browser.membership_management_self_actions`

- Source: `apps/web/e2e/incident-membership-management.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_81622c56baf7`: membership management self-demotion and self-removal recover current authority on closed incidents

#### B-040 — `module.incidents.browser.metadata_authority`

- Source: `apps/web/e2e/incident-metadata-editing.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_1dd13df75af6`: metadata uses current roles distinguishes closure and clears protected state after session or incident loss

#### B-041 — `module.incidents.browser.metadata_departure`

- Source: `apps/web/e2e/incident-metadata-editing.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_87fec4f8f1a5`: metadata and membership retained work receive sequential departure review including Stay and history

#### B-042 — `module.incidents.browser.metadata_live`

- Source: `apps/web/e2e/incident-metadata-editing.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_d262742ed34b`: metadata edits each live field sparsely clears normalizes and reviews a real version conflict

#### B-043 — `module.incidents.browser.metadata_pending_departure`

- Source: `apps/web/e2e/incident-metadata-editing.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_d74e78db396f`: metadata pending transport survives drawer return and cannot publish after deliberate incident departure

#### B-044 — `module.incidents.browser.metadata_recovery`

- Source: `apps/web/e2e/incident-metadata-editing.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`, `module.incidents.verification.lifecycle_openapi_alignment`, `module.incidents.verification.membership_audit_alignment`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_b40bd7809b2d`: metadata retains newer input and distinguishes confirmed failed reads from uncertain observation

#### B-045 — `module.incidents.browser.the_ordinary_authenticated_landing_screen_create_fa80fbfef5`

- Source: `apps/web/e2e/auth-and-incident-directory.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_d218b2939740`: creates an incident from the landing screen, lists it, and opens the workbook as incident admin

#### B-046 — `module.incidents.browser.the_ordinary_shell_exposes_incident_discovery_ra_59c190a9b5`

- Source: `apps/web/e2e/incident-administration.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_7caf936917c3`: shows incident discovery, raw querystring deep-link retrieval, and promoted-field-only patching on the ordinary incident shell

#### B-047 — `module.incidents.browser.workbook_home_and_incident_default_preferences_c_6591154e0f`

- Source: `apps/web/e2e/incident-administration.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.incidents.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-014); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.incidents`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_e4bfa26e46dc`: confirms workbook home and incident default preferences in a ready workbook shell

#### B-048 — `module.networkflow.browser_stateful.pagination_recovery`

- Source: `apps/web/e2e/network-flow-pagination.spec.ts`; project `chromium`; stage `stateful`; runtime `network_flow_claimed`.
- Verification: `module.networkflow.verification.behavior_contract`.
- Disposition: **split**; successor: S-002: owned real pagination/local recovery matrix and real rows-route cursor-expiry bridge; backend all-route expiry/scope coverage.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-017); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.networkflow`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `network_flow_pagination_recovery`: Network Analysis preserves committed pages and recovers real expired cursors across all exploration surfaces

#### B-049 — `module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0`

- Source: `apps/web/e2e/restore.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.recovery.verification.behavior_contract`.
- Disposition: **retained**; successor: S-005: dedicated restore lifecycle.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-018); candidate sharing/lifecycle qualification pending.
- Current qualification: **isolated** for the documented real expiry or specialized runtime obligation; prior scoped evidence is retained in E-028/E-030/E-034. Fresh final-suite acceptance remains S-006.

  - `scenario_90b26af20a25`: restore recovers workbook surface and executes a built-in workbook query

#### B-050 — `module.reference_data.browser.administration_async_recovery`

- Source: `apps/web/e2e/reference-pack.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.reference_data.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-019); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.reference_data`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_7c2843c7b21f`: supports asynchronous exact actions and retained selection with independent observation recovery

#### B-051 — `module.reference_data.browser.administration_real_lifecycle`

- Source: `apps/web/e2e/reference-pack.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.reference_data.verification.behavior_contract`.
- Disposition: **split**; successor: S-002: compact real lifecycle; presentation and repeated keyboard/recovery combinations move to owned presentation cases.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-019); candidate sharing/lifecycle qualification pending.
- Current qualification: signed global lifecycle remains **isolated** (E-030); repeated presentation obligations moved to the qualified owned B-050 successor (E-035/E-036).

  - `scenario_6d7791e02f0f`: imports and replays real pack bytes then activates disables reverifies and refreshes exact scope by keyboard

#### B-052 — `module.reference_data.browser.deployment_admin_browser_sessions_expose_referen_8c10ae0a9c`

- Source: `apps/web/e2e/reference-pack.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.reference_data.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-019); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.reference_data`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_448f02eff5ba`: shows Reference Pack progress and cancel controls without blocking landing interaction

#### B-053 — `module.timeline.browser.a_reviewer_session_browser_flow_visibly_performs_c3aaf62d89`

- Source: `apps/web/e2e/timeline-workbook.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.timeline.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-020); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.timeline`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_78d2b7279147`: drives review, demotion, and supersede through the visible workbook surface

#### B-054 — `module.timeline.browser.draft_create_controls`

- Source: `apps/web/e2e/timeline-workbook.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.timeline.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-020); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.timeline`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_timeline_draft_create_controls`: Timeline Add row and Evidence stay independent of Create and closed or viewer admission

#### B-055 — `module.workbook.accessibility.preferences`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_d11d07cc9caa`: a11y.preferences independent clear and recovery controls preserve keyboard focus and responsive reachability

#### B-056 — `module.workbook.browser.grid_autosave_closure`

- Source: `apps/web/e2e/workbook-grid-autosave.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_grid_autosave_closure`: Committed grid closure retains copyable original-cell drafts and reopening requires explicit activation

#### B-057 — `module.workbook.browser.ordinary_create_closure`

- Source: `apps/web/e2e/workbook-ordinary-security.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_a6b333fbcfb5`: Ordinary closure preserves copyable next authoring and requires explicit exact recovery before fresh creation

#### B-058 — `module.workbook.browser.preferences_1`

- Source: `apps/web/e2e/workbook-preferences.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_ec4568d88f1c`: preferences persist exact base and saved identities and explicit clears through subsequent startup fallback

#### B-059 — `module.workbook.browser.preferences_2`

- Source: `apps/web/e2e/workbook-preferences.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_a40d8ce0dce6`: preferences retain uncertain committed writes until explicit observed-value review and separate confirmed refresh failure

#### B-060 — `module.workbook.browser.preferences_3`

- Source: `apps/web/e2e/workbook-preferences.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_2432797dd6c5`: preferences independently publish delayed reads and guard shortcut admission across changed surfaces

#### B-061 — `module.workbook.browser.preferences_4`

- Source: `apps/web/e2e/workbook-preferences.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_b87e438937ad`: preferences preserve personal isolation and viewer boundaries on a closed incident

#### B-062 — `module.workbook.browser.reference_admission_lifecycle`

- Source: `apps/web/e2e/workbook-grid-autosave.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_reference_admission_lifecycle`: Existing reference closure role changes and access loss separate readable choices from write admission

#### B-063 — `module.workbook.browser.verify_continuous_workbook_shell_composition_top_96ea2f5084`

- Source: `apps/web/e2e/incident-administration.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_c94e91b52ee4`: updates workbook density from Account Settings while the workbook remains open

#### B-064 — `module.workbook.browser_stateful.preferences_extension`

- Source: `apps/web/e2e/extensions.stateful.spec.ts`; project `chromium`; stage `stateful`; runtime `network_flow_claimed`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_16880851da3c`: preferences store and clear the claimed authorized extension workspace without changing active identity

#### B-065 — `module.workbook.browser_stateful.query_continuation_session`

- Source: `apps/web/e2e/workbook-query-browsing.spec.ts`; project `chromium`; stage `stateful`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **replaced**; successor: S-002: local expiry response, real replacement login/paging, old protected-state/cursor retirement.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `workbook_query_continuation_session`: Workbook expiry clears continuation and account replacement starts a readable closed incident without an old cursor

#### B-066 — `module.workbook.browser_stateful.verify_timeline_public_mutations_remount_safe_tr_84f645d6c5`

- Source: `apps/web/e2e/timeline-public-route.spec.ts`; project `chromium`; stage `stateful`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-003: local public-error presentation; preserve real Timeline mutation/replay assertions.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_481272c41da9`: Recover Timeline client transaction blockers through remount-safe IDs, retry, discard, and same-field resolver handoff.
  - `scenario_da2d76b86dc8`: Verify rough Timeline row creation, inline edit, paste, pending save, refresh, and replay through /api/v1/ route contracts.

#### B-067 — `module.workbook.visual.preferences`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `module.workbook.verification.behavior_contract`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-021); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `module.workbook`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_6f57bd00c7a7`: Capture workbook preferences inspection uncertainty confirmation responsive and density.

#### B-068 — `web.design.accessibility.administrative_audit_browsing`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_db2f95a87d11`: a11y.administrative-audit exact filters paging inspection and recovery stay reachable

#### B-069 — `web.design.accessibility.deployment_user_draft_credential_dialogs_ac3e5d7802`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_admin_draft_credential_accessibility`: a11y.deployment-users guarded drafts and credential dialogs remain keyboard reachable

#### B-070 — `web.design.accessibility.incident_creation_accessible_keyboard_feedback_a_f2a19fbfdb`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_incident_creation_accessibility`: a11y.incident-creation keyboard feedback and constrained recovery remain reachable

#### B-071 — `web.design.accessibility.incident_import_workflow`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_c2edc8d7c5ec`: a11y.incident-import upload recovery progress and cancellation remain keyboard reachable

#### B-072 — `web.design.accessibility.lifecycle`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_2afeb52183ff`: a11y.lifecycle multiline reason inline review exact recovery and departure retain keyboard focus

#### B-073 — `web.design.accessibility.membership_audit_browsing`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_270d8abd7971`: a11y.membership-audit exact filters inspection paging recovery and drawer focus remain reachable

#### B-074 — `web.design.accessibility.membership_management_accessibility`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_ca8f73401a99`: a11y.membership-management draft review keyboard recovery and removal focus remain reachable

#### B-075 — `web.design.accessibility.metadata_editing`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_efa66f198be9`: a11y.metadata native inputs inline review discard and drawer focus remain reachable

#### B-076 — `web.design.accessibility.reference_pack_administration`

- Source: `apps/web/e2e/workbook.a11y.spec.ts`; project `chromium`; stage `accessibility`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_b1682646994d`: a11y.reference-packs keyboard selection upload observation and cancellation recovery remain reachable

#### B-077 — `web.design.browser.account_application_menu_keyboard_navigation_and_8cba739220`

- Source: `apps/web/e2e/incident-administration.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_18fd149b1431`: operates account application menus with keyboard focus and viewport containment

#### B-078 — `web.design.browser.incident_import_action_access_hardening`

- Source: `apps/web/e2e/incident-import.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_76b94e0a582f`: checks import actions and recovers access explicitly before accepting current profile loss

#### B-079 — `web.design.browser.incident_import_observation_recovery`

- Source: `apps/web/e2e/incident-import.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_4d254f0c49a1`: recovers import observation and cancellation without confusing terminal outcomes

#### B-080 — `web.design.browser.incident_import_real_workflow`

- Source: `apps/web/e2e/incident-import.spec.ts`; project `chromium`; stage `webserver_backed`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_48a95e7fc1d0`: imports real active and closed bundles with exact recovery and explicit workbook launch

#### B-081 — `web.design.visual.administrative_audit_browsing`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_2da065feffeb`: Capture Administrative audit loading empty unavailable paging and inspected values.

#### B-082 — `web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_46ea092e7f2b`: Capture incident creation form validation pending recovery and confirmed handoff states.

#### B-083 — `web.design.visual.incident_import_workflow`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_49044213afe8`: Capture incident import admission observation cancellation and result recovery states.

#### B-084 — `web.design.visual.lifecycle`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_4c28f69d060d`: Capture Lifecycle review pending exact recovery confirmation responsive and density.

#### B-085 — `web.design.visual.membership_audit_browsing`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_d4214d0e8b76`: Capture Membership audit loading inspected stale empty cursor recovery and density.

#### B-086 — `web.design.visual.membership_management_visual`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_71b349a3a8ce`: Capture Membership management drafts pending uncertainty confirmed recovery and density.

#### B-087 — `web.design.visual.metadata_editing`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_7462a6753b02`: Capture Metadata editing loading dirty conflict uncertainty confirmation responsive and density.

#### B-088 — `web.design.visual.reference_pack_administration`

- Source: `apps/web/e2e/workbook.visual.spec.ts`; project `chromium`; stage `visual`; runtime `default`.
- Verification: `web.design.verification.readiness_direction`.
- Disposition: **retained**; successor: S-004: same case/obligations, owned-state qualification required.
- Inspection/qualification: catalog and selected assertions reconciled against the frozen source; owner/obligation mapping in Section 10; baseline passed (E-022); candidate sharing/lifecycle qualification pending.
- Current qualification: **test_owned**, all selected cases passed in E-035/E-036 under `web.design`; canonical row/scenario identities retained. Final-source acceptance remains S-006.

  - `scenario_034f30c7dcac`: Capture Reference Pack catalog selection admission recovery and all job states.

### Direct fixture/helper discovery boundary

File-level local imports below identify the next direct-consumer inspection boundary. They are discovery metadata, not proof that every import is used by every selected case. All selected bodies and consequential helpers must be inspected before WF-00 closes.

| Source | Direct local imports |
|---|---|
| `apps/web/e2e/administrative-audit.spec.ts` | `./fixtures`, `./pages/deploymentAdministration`, `./support/administrativeAudit`, `./support/auth/deploymentUsers`, `./support/auth/sessions`, `./support/transport/publicHttpOperationClient` |
| `apps/web/e2e/auth-and-incident-directory.spec.ts` | `./fixtures`, `./pages/accountSettings`, `./pages/authGateway`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./support/auth/browserSession`, `./support/auth/deploymentUsers`, `./support/auth/sessions`, `./support/auth/storageState`, `./support/auth/suiteAdmin`, `./support/incidents/creation`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/publicHttpOperationClient`, `./support/transport/publicJsonClient` |
| `apps/web/e2e/auth.support.spec.ts` | `./fixtures`, `./pages/accountSettings`, `./pages/deploymentAdministration`, `./pages/incidentDirectory` |
| `apps/web/e2e/clock.spec.ts` | `./fixtures`, `./pages/accountSettings`, `./pages/authGateway`, `./support/auth/deploymentUsers`, `./support/auth/sessions`, `./support/runtime/fixtureIdentity`, `./support/runtime/testClock` |
| `apps/web/e2e/collaboration.spec.ts` | `./fixtures`, `./pages/incidentDirectory`, `./support/auth/browserSession`, `./support/auth/sessions`, `./support/collaboration/replay`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/incidentSocket`, `./support/workbook/mutationAnchors`, `./support/workbook/query`, `./support/workbook/recovery` |
| `apps/web/e2e/extensions.stateful.spec.ts` | `./fixtures`, `./pages/deploymentAdministration`, `./support/collaboration/replay`, `./support/extensions/network_flow_activity/workspace`, `./support/runtime/configuration` |
| `apps/web/e2e/incident-administration.spec.ts` | `./fixtures`, `./pages/accountSettings`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./support/auth/accountEditingFixture`, `./support/auth/browserSession`, `./support/auth/deploymentUsers`, `./support/extensions/network_flow_activity/workspace`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/requestInterception`, `./support/workbook/query`, `./support/workbook/rowMutations`, `./support/workbook/savedViews` |
| `apps/web/e2e/incident-import.spec.ts` | `./fixtures`, `./pages/deploymentAdministration`, `./support/incidents/creation`, `./support/incidents/import` |
| `apps/web/e2e/incident-lifecycle.spec.ts` | `./fixtures`, `./pages/incidentDirectory`, `./support/administrativeAudit`, `./support/auth/browserSession`, `./support/incidentLifecycle`, `./support/incidentMembershipManagement`, `./support/incidentMetadata`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/publicHttpOperationClient`, `./support/transport/publicJsonClient`, `./support/workbook/query`, `./support/workbook/recovery`, `./support/workbook/rowMutations`, `./support/workbook/savedViews` |
| `apps/web/e2e/incident-membership-audit.spec.ts` | `./fixtures`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./support/administrativeAudit`, `./support/incidentMembershipAudit`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/fixtureIdentity`, `./support/transport/publicHttpOperationClient` |
| `apps/web/e2e/incident-membership-management.spec.ts` | `./fixtures`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./support/administrativeAudit`, `./support/auth/browserSession`, `./support/auth/deploymentUsers`, `./support/incidentMembershipManagement`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/publicHttpOperationClient`, `./support/transport/publicJsonClient` |
| `apps/web/e2e/incident-metadata-editing.spec.ts` | `./fixtures`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./support/administrativeAudit`, `./support/auth/browserSession`, `./support/incidentMembershipManagement`, `./support/incidentMetadata`, `./support/incidents/fixtures`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/publicHttpOperationClient`, `./support/transport/publicJsonClient`, `./support/transport/requestInterception` |
| `apps/web/e2e/network-flow-pagination.spec.ts` | `./fixtures`, `./support/extensions/network_flow_activity/workspace`, `./support/runtime/testClock` |
| `apps/web/e2e/reference-pack.spec.ts` | `./fixtures`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./support/auth/sessionPresentation`, `./support/referencePacks` |
| `apps/web/e2e/restore.spec.ts` | `../../../tools/harness/browser/index.mjs`, `./fixtures`, `./measurement/timingSupport`, `./support/auth/browserSession` |
| `apps/web/e2e/session-recovery.spec.ts` | `./fixtures`, `./support/auth/browserSession`, `./support/auth/sessions`, `./support/collaboration/replay`, `./support/runtime/configuration`, `./support/runtime/testClock` |
| `apps/web/e2e/timeline-public-route.spec.ts` | `./fixtures`, `./pages/incidentDirectory`, `./pages/workbookInspector`, `./support/auth/browserSession`, `./support/auth/sessions`, `./support/collaboration/replay`, `./support/incidents/fixtures`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/publicHttpOperationClient`, `./support/transport/requestInterception`, `./support/transport/testControlEnvironment`, `./support/workbook/query`, `./support/workbook/recovery`, `./support/workbook/rowMutations` |
| `apps/web/e2e/timeline-related-evidence.spec.ts` | `./fixtures`, `./pages/incidentDirectory`, `./support/auth/browserSession`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/publicHttpOperationClient`, `./support/transport/publicJsonClient`, `./support/workbook/history`, `./support/workbook/query`, `./support/workbook/recovery`, `./support/workbook/rowMutations` |
| `apps/web/e2e/timeline-workbook.spec.ts` | `./fixtures`, `./measurement/timingSupport`, `./pages/incidentDirectory`, `./pages/workbookInspector`, `./support/auth/browserSession`, `./support/entities/mentions`, `./support/incidentLifecycle`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/requestInterception`, `./support/workbook/collections`, `./support/workbook/history`, `./support/workbook/ordinaryCreate`, `./support/workbook/query`, `./support/workbook/recovery`, `./support/workbook/rowMutations`, `./support/workbook/timelineCaptureActions` |
| `apps/web/e2e/visual-harness.spec.ts` | `./fixtures`, `./pages/incidentDirectory`, `./support/auth/visualPreferences`, `./support/incidents/fixtures`, `./support/runtime/applicationReadiness`, `./support/runtime/fixtureIdentity`, `./support/runtime/fixtureTLS`, `./support/runtime/frontendBuild`, `./support/runtime/visualRenderer`, `./support/timeline/timelineInvestigation`, `./support/visual/capture`, `./support/visual/metadataNormalization`, `./support/visual/timelineShellAnchor`, `./support/visual/timelineVisualFixture`, `./support/workbook/history`, `./support/workbook/rowMutations` |
| `apps/web/e2e/workbook-grid-autosave.spec.ts` | `./fixtures`, `./support/auth/browserSession`, `./support/auth/sessions`, `./support/auth/storageState`, `./support/incidentLifecycle`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/incidentSocket`, `./support/transport/publicHttpOperationClient`, `./support/transport/publicJsonClient`, `./support/workbook/history`, `./support/workbook/ordinaryCreate`, `./support/workbook/query`, `./support/workbook/savedViews` |
| `apps/web/e2e/workbook-ordinary-security.spec.ts` | `./fixtures`, `./support/incidentLifecycle`, `./support/incidents/memberships`, `./support/runtime/fixtureIdentity`, `./support/transport/incidentSocket`, `./support/workbook/ordinaryCreate`, `./support/workbook/query` |
| `apps/web/e2e/workbook-preferences.spec.ts` | `./fixtures`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./support/auth/browserSession`, `./support/incidentLifecycle`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/workbook/savedViews` |
| `apps/web/e2e/workbook-query-browsing.spec.ts` | `./fixtures`, `./support/auth/browserSession`, `./support/auth/sessions`, `./support/incidentLifecycle`, `./support/incidents/fixtures`, `./support/incidents/memberships`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/runtime/testClock`, `./support/transport/incidentSocket`, `./support/transport/publicHttpOperationClient`, `./support/transport/publicJsonClient`, `./support/transport/requestInterception`, `./support/workbook/literalSetFilters`, `./support/workbook/ordinaryCreate`, `./support/workbook/query`, `./support/workbook/rowMutations`, `./support/workbook/savedViews` |
| `apps/web/e2e/workbook.a11y.spec.ts` | `./fixtures`, `./pages/accountSettings`, `./pages/authGateway`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./support/administrativeAudit`, `./support/auth/accountEditingFixture`, `./support/auth/browserSession`, `./support/auth/deploymentUsers`, `./support/auth/sessions`, `./support/auth/storageState`, `./support/auth/suiteAdmin`, `./support/collaboration/replay`, `./support/entities/mentions`, `./support/evidence/fixtures`, `./support/evidence/uploads`, `./support/extensions/network_flow_activity/workspace`, `./support/incidentLifecycle`, `./support/incidentMembershipAudit`, `./support/incidentMembershipManagement`, `./support/incidentMetadata`, `./support/incidents/creation`, `./support/incidents/fixtures`, `./support/incidents/import`, `./support/incidents/memberships`, `./support/referencePacks`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/transport/incidentSocket`, `./support/transport/requestInterception`, `./support/transport/testControlEnvironment`, `./support/workbook/collections`, `./support/workbook/contextualCreate`, `./support/workbook/coordinationCreate`, `./support/workbook/decisionSupersession`, `./support/workbook/history`, `./support/workbook/indicatorCanonicalCreate`, `./support/workbook/indicatorLifecycle`, `./support/workbook/indicatorObservations`, `./support/workbook/noteCreate`, `./support/workbook/ordinaryCreate`, `./support/workbook/query`, `./support/workbook/querySummary`, `./support/workbook/recovery`, `./support/workbook/references`, `./support/workbook/rowMutations`, `./support/workbook/saveStatus`, `./support/workbook/timelineCaptureActions`, `./support/workbook/timelineRelatedEvidence` |
| `apps/web/e2e/workbook.visual.spec.ts` | `./fixtures`, `./pages/accountSettings`, `./pages/deploymentAdministration`, `./pages/incidentDirectory`, `./pages/workbookInspector`, `./support/administrativeAudit`, `./support/auth/accountEditingFixture`, `./support/auth/browserSession`, `./support/auth/visualPreferences`, `./support/collaboration/replay`, `./support/entities/mentions`, `./support/evidence/fixtures`, `./support/evidence/uploads`, `./support/extensions/network_flow_activity/workspace`, `./support/incidentLifecycle`, `./support/incidentMembershipAudit`, `./support/incidentMembershipManagement`, `./support/incidentMetadata`, `./support/incidents/creation`, `./support/incidents/fixtures`, `./support/incidents/import`, `./support/incidents/memberships`, `./support/referencePacks`, `./support/runtime/applicationReadiness`, `./support/runtime/configuration`, `./support/runtime/fixtureIdentity`, `./support/runtime/visualRenderer`, `./support/timeline/inspectorReadingFixture`, `./support/timeline/timelineInvestigation`, `./support/transport/incidentSocket`, `./support/transport/requestInterception`, `./support/transport/testControlEnvironment`, `./support/visual/capture`, `./support/visual/fixtures`, `./support/visual/metadataNormalization`, `./support/visual/profile`, `./support/visual/richTimelineInspector`, `./support/visual/timelineShellAnchor`, `./support/visual/timelineVisualFixture`, `./support/visual/workbookLayoutStudies`, `./support/workbook/collections`, `./support/workbook/contextualCreate`, `./support/workbook/coordinationCreate`, `./support/workbook/decisionSupersession`, `./support/workbook/history`, `./support/workbook/indicatorLifecycle`, `./support/workbook/indicatorObservations`, `./support/workbook/noteCreate`, `./support/workbook/ordinaryCreate`, `./support/workbook/query`, `./support/workbook/recovery`, `./support/workbook/rowMutations`, `./support/workbook/savedViews`, `./support/workbook/timelineCaptureActions`, `./support/workbook/timelineRelatedEvidence` |

### Expanded ownership-consumer boundary

Expanded static name search includes session capture and login, in addition to the narrower E-013 cleanup search. It finds 51 TypeScript files. This is the required migration/check boundary for S-001, including existing shareable and measurement consumers outside the 26-file baseline; it is not an enlarged baseline selection or a claim that those consumers have passed. An unchanged facade name is acceptable only when its implementation delegates all ownership to the single scope, without preserving a competing cleanup model.

| Consumer category | Exact source paths |
|---|---|
| Entry points | `apps/web/e2e/auth-and-incident-directory.spec.ts`, `apps/web/e2e/clock.spec.ts`, `apps/web/e2e/collaboration.spec.ts`, `apps/web/e2e/contextual-coordination-create.spec.ts`, `apps/web/e2e/coordination-public-route.spec.ts`, `apps/web/e2e/enterprise-auth.spec.ts`, `apps/web/e2e/entity-find.spec.ts`, `apps/web/e2e/evidence-integration.spec.ts`, `apps/web/e2e/evidence.spec.ts`, `apps/web/e2e/fixtures.ts`, `apps/web/e2e/incident-administration.spec.ts`, `apps/web/e2e/incident-lifecycle.spec.ts`, `apps/web/e2e/incident-membership-audit.spec.ts`, `apps/web/e2e/incident-metadata-editing.spec.ts`, `apps/web/e2e/incident.support.spec.ts`, `apps/web/e2e/indicator-canonical-create.spec.ts`, `apps/web/e2e/inspector-actions.spec.ts`, `apps/web/e2e/keyboard.spec.ts`, `apps/web/e2e/linked-note-create.spec.ts`, `apps/web/e2e/measurement/evidence-inspector.spec.ts`, `apps/web/e2e/measurement/evidence-verification.spec.ts`, `apps/web/e2e/measurement/timeline-anchor-trace.spec.ts`, `apps/web/e2e/mentions.resolve.spec.ts`, `apps/web/e2e/timeline-auto-resolution-feedback.spec.ts`, `apps/web/e2e/timeline-autosave-feedback.spec.ts`, `apps/web/e2e/timeline-bulk-tag.spec.ts`, `apps/web/e2e/timeline-collection-input.spec.ts`, `apps/web/e2e/timeline-find.spec.ts`, `apps/web/e2e/timeline-grid-entry.spec.ts`, `apps/web/e2e/timeline-public-route.spec.ts`, `apps/web/e2e/timeline-range-selection.spec.ts`, `apps/web/e2e/timeline-related-evidence.spec.ts`, `apps/web/e2e/timeline-scalar-clipboard.spec.ts`, `apps/web/e2e/timeline-workbook.spec.ts`, `apps/web/e2e/timeline.support.spec.ts`, `apps/web/e2e/workbook-destination-handoff.spec.ts`, `apps/web/e2e/workbook-grid-autosave.spec.ts`, `apps/web/e2e/workbook-inspector-edit.spec.ts`, `apps/web/e2e/workbook-ordinary-security.spec.ts`, `apps/web/e2e/workbook-preferences.spec.ts`, `apps/web/e2e/workbook-query-browsing.spec.ts`, `apps/web/e2e/workbook-save-status.spec.ts`, `apps/web/e2e/workbook.a11y.spec.ts`, `apps/web/e2e/workbook.visual.spec.ts` |
| Helpers and support tests | `apps/web/e2e/pages/incidentDirectory.ts`, `apps/web/e2e/support/auth/sessions.ts`, `apps/web/e2e/support/collaboration/replay.ts`, `apps/web/e2e/support/performance/ac043Snapshot.ts`, `apps/web/e2e/support/runtime/cleanup.test.ts`, `apps/web/e2e/support/runtime/cleanup.ts`, `apps/web/e2e/support/transport/requestInterception.ts` |

Pinned Playwright source inspection confirms that its test instrumentation applies effective context settings before both browser and request-context creation. The new scope should delegate acquisition to that instrumented browser/request client and validate inheritance; duplicating the framework's entire option-merging implementation would add avoidable coupling. The current fixture's explicit option subset alone is not evidence of a settings defect.

Direct acquisition search additionally identifies `apps/web/e2e/support/auth/suiteAdmin.ts`. All raw browser/request acquisition sites in the listed consumers must be classified as test-owned or borrowed control-plane work. Worker bootstrap/provisioner requests retain their worker boundary; ordinary and secondary test contexts must be admitted by the common scope before the acquisition can escape. Restored-deployment sessions retain their separate runtime boundary.

Additional gate families are `responseBarrier`, `auditBrowserBarrier`, `referencePackBarrier`, `accountResponseGate`, collaboration patch-controller holds and local read/write/admission gates in metadata, membership, import and reference fixtures. Bounded timer/readiness waits are not automatically response gates; classify their cancellation needs by their dependency rather than replacing every Promise.

### Fault-control retirement boundary

The frozen-baseline authored/runtime search identified exactly these then-active source consumers (retired or migrated in E-028/E-032): `apps/web/e2e/timeline-public-route.spec.ts`, `apps/web/e2e/workbook.a11y.spec.ts`, `apps/web/e2e/workbook.visual.spec.ts`, `internal/app/server/server_profile_harness.go`, `internal/app/server/runtime_assembly.go`, `internal/platform/httpapi/httpapi.go`, `internal/platform/harnessruntime/controls.go`, `internal/platform/harnessruntime/public_error_fault.go`, `internal/platform/harnessruntime/public_error_fault_integration_test.go`, `tools/test_families/harness.browser.json`, `tools/harness_schema_attachments.json`, and `tools/schemas/cartulary.test.public_error_fault.v1.schema.json`. Generated schema-validator/catalog references are regenerated from their authored inputs. Negative retirement witnesses and historical migration records may name the retired route; they must not register or implement it.

Owner closure covers Testing Harness REQ-450's route list, Section 8's schema inventory, Section 12.2.4/REQ-454, and AC-035's fault-specific assertions. REQ-453's token guards and REQ-460–464's clock behavior remain authoritative, including the clock's independently justified process-replacement rule. Network Flow controls are not part of this retirement.

## 3. Module Boundary Diagnosis

| Responsibility | Required owner | Disposition |
|---|---|---|
| Browser contexts, request clients, route barriers, and session cleanup | Browser test-support lifecycle owner | Consolidate lifetime management. |
| Feature-specific fixture data and browser choreography | Existing feature support modules | Retain semantic ownership; remove duplicated lifetime machinery. |
| Cohort compatibility, packing, resets, and resource admission | Existing harness scheduler/browser owners | Preserve the existing execution architecture. |
| Session expiry, last-admin enforcement, cursor validity, and error serialization | Existing production owners | Verify through real backend boundaries. |
| Browser recovery, focus, accessibility, layout, and safe presentation | Browser scenarios | Preserve browser execution with owned fixtures. |
| Global clock, Reference Pack activation, restore, and publication overlap | Explicit dedicated scenarios | Retain only the isolation their behavior requires. |

**Rationale:** Resource reuse and test-state isolation are different concerns. A shared database or renderer is compatible with independent tests when mutable records, credentials, contexts, and asynchronous work have explicit owners.

## 4. Public Contract and Behavior Freeze Map

**TR-REQ-004 — Coverage preservation.** A test MUST NOT be removed solely because another test exercises the same route or feature name. Consolidation requires equivalent coverage of the relevant preconditions, transitions, negative outcomes, durable effects, and observation boundaries.

| Contract | Required assurance after cutover |
|---|---|
| Authentication and administration | Real credential/session behavior, authorization denial, revocation effects, version conflicts, and last-admin protection; browser feedback and recovery remain covered separately. |
| Incident/workbook mutations | Real persistence, lifecycle and membership admission, replay identity, conflict handling, retained authoring, and stale-response fencing. |
| Collaboration | Real multi-client publication, presence, conflicts, session effects, and recovery. |
| Queries and listings | Real pagination and authorization; browser continuation, focus, recovery, and account replacement; unrelated data MUST NOT alter scoped assertions. |
| Evidence and imports | Real commit and job behavior, partial success, access restrictions, preview/download boundaries, and recovery. |
| Reference Packs | Real verification, activation, disablement, re-verification, selected refresh, and refresh-all effects. |
| Accessibility and visual behavior | Existing meaningful keyboard, focus, accessible-name, semantic-state, responsive-layout, and screenshot obligations. |
| Harness accounting | Exact selection, truthful attribution, settled cleanup, visible failures, quarantine, and absence of product retries. |

A browser mock MUST NOT be credited as proof of server enforcement or persistence. A backend test MUST NOT be credited as proof of browser focus, navigation, rendering, or recovery.

Production selectors and interfaces MUST remain unchanged unless an independently authorized production correction requires otherwise.

The coverage map MUST cite the applicable adopted owner clauses. In particular, last-admin enforcement remains subject to Core 04 AC-177 and its Core 01/Core 02 dependencies; Network Flow cursor expiry and scope remain subject to the [Network Flow Activity NLSpec](../network-flow-activity-nlspec.md); Reference Pack behavior remains subject to the [Reference Pack Subsystem NLSpec](../reference-pack-subsystem-nlspec.md). These pointers do not replace the exact obligation mapping required by V-001.

## 5. Coupling and Boundary Findings

| ID | Evidence-backed finding | Required action |
|---|---|---|
| F-001 | Many isolated scenarios already use fresh actors, contexts, and incident-scoped data. | Qualify compatible successors for `test_owned`; remove unnecessary singleton declarations. |
| F-002 | Session tracking does not own every secondary context; some helpers create raw contexts directly. | Introduce one fixture-owned browser resource lifetime and migrate all consumers of changed helpers. |
| F-003 | Response gates can remain unresolved on failure; some fulfillment errors are swallowed. | Make barriers cancellation-aware and require settled cleanup with visible errors. |
| F-004 | Long browser scenarios combine ordinary workflows with global clock or administrative mutation. | Separate browser interaction coverage from global enforcement checks. |
| F-005 | Public-error injection has a normative runtime-global pending-fault restriction. | Retire this endpoint through an owner-first cutover, rather than introducing another fault registry. |
| F-006 | Test actors and other durable data survive session cleanup. | Require scoped assertions, terminal asynchronous work, and bounded cohort lifetimes. |
| F-007 | Browser cohort packing already exists. | Use existing compatibility and admission mechanisms; do not add a second runner or scheduler. |
| F-008 | The lifetime/cancellation contract and coordinated fault retirement are not closed in the adopted owner. | Amend the Testing Harness NLSpec and reviewed projections, including obsolete fault-specific process-replacement wording. |
| F-009 | The original handoff lacks the exact case inventory, separate slice workstreams, mandatory checkpoints, and a completion slice. | Record obligation-level successors and retained evidence; update this tracker between every workstream. |

F-001 through F-006 and F-008/F-009 are `must_fix` for the affected conversion slices. F-007 is `intentional/no_action` for scheduler replacement.

### Accepted remediation decisions

| Finding | Change areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if unresolved | Completion evidence |
|---|---|---|---|---|---|
| F-001 | Tests, authored verification/catalog contracts, generated projections, documentation | Explicit state boundaries safely reuse infrastructure and make future admission predictable. | Execution grouping changes; retain canonical identities and record retired-row successors. | Repeated startup/reset cost and inherited unnecessary isolation. | V-001, V-003, V-007, V-008 |
| F-002 | Harness specification, support implementation, fixtures, tests, documentation | One owner closes partial-acquisition and fixture-ordering leaks. | Migrate all changed-helper consumers together; remove competing ownership implementations. | Leaked contexts/credentials and cross-test interference. | V-002, V-004 |
| F-003 | Harness specification, browser support, tests, diagnostics | Explicit cancellation settles teardown without approving delayed mutations. | Gate consumers handle cancellation; repeated disposal returns the same failed or successful completion. | Hung teardown, hidden route failures, unintended mutations, false passes. | V-002, V-007; credential-safe diagnostics |
| F-004 | Browser/backend tests, verification routing, documentation | Separate browser assurance from real enforcement; keep global setup compact. | Split/consolidate with predecessor-to-successor accounting; preserve owner-defined product behavior. | Serialized routine coverage or mocks replacing real enforcement evidence. | V-001, V-004, V-005, V-006 |
| F-005 | Specification, runtime assembly/middleware, schemas, tests, catalogs, documentation | Remove shared fault state and a test-only generic HTTP dependency. | Delete the test endpoint/internal interfaces without an alias, replacement server API, or dormant code. | Global interference and unnecessary security-sensitive controls. | V-001, V-005, V-006, V-009; absent route even with harness controls enabled |
| F-006 | Feature fixtures, browser tests, lifecycle specification, documentation | Owned assertions and terminal work permit bounded durable residue without per-test deletion. | Change assertions/teardown; no production deletion API, shared actor pool, or outer transaction around HTTP work. | Cardinality/order dependence, identity collisions, background interference. | V-002, V-003, V-004 |
| F-007 | Existing implementation/regressions and decision documentation | Retain one compatibility, admission, reset, and accounting mechanism. | No scheduler replacement. | An unnecessary rewrite creates divergent lifecycle policy. | V-007 |
| F-008 | Specification, reviewed machine projections, documentation | Lasting behavior has one adopted owner independent of this tracker. | Coordinated owner revision; unrelated Network Flow controls and product Core requirements remain intact. | Implementation satisfies this handoff but contradicts its adopted owner. | Human owner/projection review; V-009 |
| F-009 | This tracker, verification routing where needed, retained evidence | Resumption from verified state without reconstructing decisions or mistaking test presence for execution. | Preserve historical IDs/failures while expanding evidence and checkpoint records. | Lost coverage, false completion, stale evidence supporting acceptance. | V-001, V-008, V-009 |

| Finding | Slice | Validation |
|---|---|---|
| F-001 | S-004 | V-001, V-003, V-007, V-008 |
| F-002, F-003 | S-001, S-005 | V-002, V-004, V-007 |
| F-004 | S-002 | V-001, V-004, V-005, V-006 |
| F-005 | S-003 | V-001, V-005, V-006, V-009 |
| F-006 | S-001, S-004 | V-002, V-003, V-004, V-008 |
| F-007 | Preserve existing owner boundaries | V-007; no scheduler-replacement slice |
| F-008 | WF-01, S-001, S-003 | V-009 and owner/projection review |
| F-009 | WF-00, all workstream checkpoints, S-006 | V-001, V-008, V-009 |

## 6. Refactor Workstreams

The dependency column is authoritative. Workstreams MUST NOT be merged into the production branch independently.

| Workflow | Slice | Dependencies | Main risk | Binary exit criterion |
|---|---|---|---|---|
| WF-00 — Freeze coverage and baseline | Baseline characterization | None | Counts or test presence mistaken for coverage proof. | All 88 rows and 97 cases, helpers, selectors, owners, obligations and dispositions are accounted for; baseline correctness/resource evidence is captured. |
| WF-01 — Amend harness contracts | Specification cleanup | WF-00 | Tracker becomes a second specification. | Complete owner/interface amendments and acceptance criteria are reviewed for the coordinated cutover. |
| WF-02 — Establish resource ownership | S-001 | WF-01 | Partial-acquisition leaks or hidden cleanup failures. | V-002/V-004 pass; all changed-helper consumers share one lifetime. |
| WF-03 — Restructure production coverage | S-002 | WF-02 | Presentation mocks credited as server enforcement. | Inspected successors preserve every removed assertion at the correct assurance layer; V-001/V-004–006 pass. |
| WF-06 — Retire public-error fault control | S-003 | WF-03 | Partial retirement or missing coverage. | Successors, retirement checks and authored/generated closure pass. |
| WF-07 — Integrate specialized lifecycles | S-005 | WF-06 | Early dependency closure or leaked fixture processes. | V-002/V-005 pass for startup/body failure, interruption and cleanup failure. |
| WF-04 — Admit cohorts and consolidate tests | S-004 | WF-07 | Unsafe sharing or duplicate/missing execution. | V-003/V-007 pass; every promoted case has proof and all remaining isolation is justified. |
| WF-05 — Qualify and hand off | S-006 | WF-04 | Stale evidence or displaced execution cost. | V-001–009 pass on compatible final-source evidence and all completion records are present. |

**TR-REQ-010 — Mandatory checkpoints.** Execution follows the table order. After each workstream, and before beginning the next one, this tracker MUST record its status, source revision, substantive changes, exact commands, result roots, primary/cleanup failures, coverage dispositions, remaining risks, and next entry gate. Blocked work MUST be marked `BLOCKED`; dependent work MUST NOT advance. Tests or catalog declarations alone do not establish completion. Existing workflow IDs are retained; WF-06/WF-07 separate formerly combined implementation work.

Intermediate commits MAY support review, but MUST NOT introduce a supported dual-mode architecture. Omission of intermediate commits changes no acceptance requirement.

## 7. Proposed Refactor Slice Plan

### S-001 — One browser resource lifetime

**TR-REQ-005 — Ownership interface.** Introduce one test-scoped `OwnedBrowserScope` in existing browser runtime support. Existing fixtures MUST use it through these capabilities:

| Capability | Required semantics |
|---|---|
| `registerActor(userId)` | Establish owned-actor session cleanup before attempting login, including failed or incompletely observed acquisition. |
| `newContext(options)` | Create and register ownership before exposing the context; preserve required project context options. |
| `ownRequestContext(context)` | Register an independently created request client for disposal. |
| `registerSession(snapshot)` | Register an owned session immediately after acquisition; borrowed provisioner sessions MUST NOT be registered as test subjects. |
| `barrier()` | Return a test-local waitable barrier with idempotent release and an explicit cancelled outcome. |
| `dispose()` | Share one completion promise, including rejection, across repeated calls; reject new acquisitions/registrations after disposal starts. |

In-flight acquisitions MUST be registered before awaiting them; a context that resolves during disposal MUST NOT escape ownership. Feature-specific data and job behavior remain with feature fixtures, whose work MUST settle before their dependencies are released. Changed-helper consumers MUST migrate together without retaining competing lifetime implementations.

Teardown MUST:

1. Stop new registrations and cancel outstanding test-owned waits.
2. Settle feature-owned work, then begin draining every page/context route owner before awaiting their collective settlement, while request clients remain usable.
3. Close every owned browser context, including secondary clients.
4. Revoke and verify owned sessions using the still-live provisioner.
5. Dispose remaining owned request clients.
6. Preserve the primary failure and report additional cleanup failures using redacted diagnostics that exclude credentials and private values.

A cancelled barrier MUST NOT release a delayed mutation as if the scenario had approved it. Unexpected route, revocation, or disposal errors MUST NOT be swallowed. Cleanup MUST remain within existing deadlines.

The worker provisioner remains control-plane-only. Tests MUST NOT inspect or mutate its production identity as their subject.

Durable test data MAY remain until the cohort reset when it is uniquely owned and observationally isolated. Omission of per-test deletion MUST NOT permit pending jobs, active sessions, open contexts, or callbacks to survive. Global deletion, truncation, per-test deployment reset, and an outer transaction around independently committed HTTP work are prohibited substitutes for ownership.

### S-002 — Restructure the constrained scenarios

| Scenario | Required replacement |
|---|---|
| Last-admin browser workflow | Retain real owned-user creation, patching, stale-version handling, and navigation. Exercise `last_deployment_admin` browser presentation with a page-local response. Preserve real backend demotion and deactivation rejection, unchanged durable state, and absence of unintended revocation. Remove the browser helper that demotes every other administrator. |
| HTTP idle expiry | Retain one compact dedicated browser scenario proving actual expiry, rejected access, and fresh login. |
| Collaboration session recovery | Split revoke-all, logout, and idle expiry into independent cases. Revoke-all and logout MUST use owned users and shared infrastructure. Retain dedicated clock ownership only for actual idle-expiry/socket recovery. |
| Workbook continuation/account replacement | Remove global-clock manipulation. Use a scoped expired-session response to verify continuation retirement and protected-state removal, then real replacement-account login and real paging. Assert that the replacement starts without the old cursor. |
| Network Flow pagination | Retain real pagination and browser recovery across rows, diagnostics, and contributors in a shareable scenario. Use exact page-local expiry responses for the recovery matrix. Retain one compact real cursor-expiry browser bridge; backend integration tests MUST continue to cover expiry boundaries and scope rules across all three routes. |
| Reference Pack lifecycle | Retain real signed import/replay, activation, disablement, re-verification, selected refresh, and refresh-all in a dedicated scenario. Move repeated presentation, keyboard, and observation-recovery combinations into existing owned presentation scenarios. |

The implementation MUST NOT introduce a per-user clock, shortened production TTL, mutable shared actor pool, or generic database-mutation control API.

### S-003 — Retire public-error fault control

**TR-REQ-006 — Error-boundary replacement.** The Timeline, Evidence accessibility, and Evidence visual consumers MUST use context-local, exact-method/path error responses for presentation checks.

Backend tests MUST independently verify the production serializer, status/code/details behavior, request correlation, and private-diagnostic suppression. Existing real Evidence preview/download and mutation workflows MUST remain.

The same merged change MUST retire:

- The public-error-fault runtime route and registration.
- Its registry, middleware dependency, browser clients, response schema, and executable schema references.
- Tests whose only remaining purpose is verifying that retired control.
- Testing Harness §12.2.4, REQ-454, and affected acceptance/schema inventory entries, replacing them with explicit retirement and successor coverage.

The current single-pending-fault requirement remains authoritative until that coordinated amendment is adopted. The implementation MUST NOT retain an alias, compatibility parser, replacement server fault API, or dormant implementation.

### S-004 — Cohort admission and test consolidation

**TR-REQ-007 — Sharing admission.** Every promoted row MUST prove ownership of its actors, sessions, records, preferences, routes, and asynchronous work. A multi-case row MUST satisfy this requirement for every selected case.

Compatible successors MUST use existing bounded packing. Runtime profile, stage, project, trust, fixture, PostgreSQL isolation, service dependencies, affinity, and resource claims MUST remain part of compatibility. Default and `network_flow_claimed` runtimes MUST remain separate.

One worker per cohort, existing deadlines, failure-stop behavior, quarantine, and zero product retries MUST remain unchanged.

Exact listing assertions MUST filter by owned identities. Directory helpers MUST navigate explicitly when they require the directory; they MUST NOT depend on raw-root cardinality or residue from preceding tests.

The post-cutover isolation inventory MUST be limited to these justified classes:

1. Actual HTTP session expiry.
2. Actual socket/session-expiry recovery.
3. The real Network Flow cursor-expiry bridge.
4. Global Reference Pack lifecycle.
5. Restore lifecycle.
6. Concurrent frontend publication.

This is an architectural target, not permission to force an unsafe promotion. An additional irreducible isolation requirement MUST block acceptance until its production obligation and resource boundary are documented and the plan is amended.

Consolidation MUST preserve distinct owners, postconditions, runtime profiles, and behavior-sensitive fixtures. Canonical surviving identities MUST follow REQ-678. Retired row identities MUST use existing migration accounting without becoming executable aliases.

### S-005 — Specialized lifecycles

Restore MUST retain disposable source/target databases, runtime roots, and process cleanup. Publication overlap MUST retain independently owned builds and immutable loaded-asset verification.

These scenarios MUST use the common browser cleanup interface, but this cutover MUST NOT pool their mutable deployment resources or weaken their process-lifetime assertions.

Browser dependencies MUST close before their fixture runtimes retire. Sessions belonging to a restored deployment remain in its disposable runtime boundary and MUST NOT be revoked through the ordinary deployment's provisioner.

### S-006 — Validation and handoff completion

Finalize authored/generated consistency, then run final correctness, lifecycle, browser, evidence and resource qualification. Record complete baseline-to-successor coverage, the final isolation inventory, retirement accounting, compatible V-001–009 results, owner/projection review, exact commands/revisions/environments/result roots, cleanup and retained-scan outcomes, and migration/rollback notes. Preserve failed and superseded evidence. No skipped required gate may be represented as complete. Planning completeness, slice readiness, authorization and execution completion remain independent outcomes.

### Slice entry and completion gates

| Slice | Entry gate | Current readiness | Binary completion criterion |
|---|---|---|---|
| S-001 | WF-00 characterization and WF-01 owner/interface review | COMPLETE: E-028 | V-002 passes; changed helper consumers use the common lifetime; V-004 and V-007 remain correct. |
| S-002 | WF-02 ownership closure and inspected production-obligation mapping | COMPLETE: E-030 | V-001, V-004, V-005, and V-006 pass for every replacement. |
| S-003 | WF-03 checkpoint and coordinated fault-control retirement review | COMPLETE: E-032 | Successor coverage passes and retired runtime/schema/client paths are absent. |
| S-004 | WF-07 checkpoint and per-candidate interference evidence | COMPLETE: E-036 | V-003 and V-007 pass; generated selection and remaining-isolation accounting agree. |
| S-005 | WF-06 checkpoint and existing lifecycle characterization | COMPLETE: E-034 | V-002 and V-005 pass without weakened process/resource assertions. |
| S-006 | WF-04 checkpoint | READY: E-036; final V-001–009 acceptance is required | V-001–009 and the complete final handoff pass. |

The user has authorized all slices; their evidence/readiness gates still apply. Their common rollback is the complete change-set reversion defined by TR-REQ-009; no slice establishes a supported intermediate mode.

## 8. Validation Plan

Execution below is authorized, subject to workstream entry gates and an available host. No validation may interfere with independently owned work in this workspace.

| ID | Obligation | Acceptance criterion |
|---|---|---|
| V-001 | Coverage reconciliation | Every baseline production obligation has an inspected successor and correct assurance layer; no unexplained removal. |
| V-002 | Scope lifecycle | Failure immediately after acquisition, pending response, login, job admission, and context creation settles resources without hidden errors. Repeated disposal is safe. |
| V-003 | Sharing interference | Compatible cases pass without intervening resets in forward and reverse order, with unrelated users, incidents, audit events, and jobs present. |
| V-004 | Security witnesses | Revoking one test's sessions leaves unrelated sessions and the provisioner usable; old owned sessions are unauthorized. |
| V-005 | Real integration | Last-admin, HTTP/socket expiry, cursor expiry, persistence/replay, Reference Packs, restore, and publication retain real-boundary evidence. |
| V-006 | Browser assurance | Keyboard, focus, accessibility, screenshots, stale-response fencing, account replacement, and recovery remain correct. |
| V-007 | Harness accounting | Exact cases execute once; early failure records unexecuted work; cleanup failure quarantines the allocation; incompatible tuples do not share. |
| V-008 | Performance | The protocol below establishes reproducibly lower elapsed time and CPU use. |
| V-009 | Final closure | Required checks, explicit suites, owner evidence audits, cleanup, and retained scans pass on the final source. |

Forward/reverse qualification MUST start from separate clean invocations. Each selected case runs once per cohort generation. The implementation MUST NOT replay deterministic actor identities repeatedly against an uncleared database.

Qualification scheduling MUST be exercised through the existing harness regression path; it MUST NOT introduce a public shuffle flag, second runner, or production ordering dependency.

### Performance protocol

**TR-REQ-008 — Measurement gate.**

- Freeze baseline and candidate revisions, a production-coverage equivalence mapping, host capacity, toolchain, lane count, and command portfolio before measurement.
- Measure the five ordinary browser suites: `functional`, `stateful`, `support`, `accessibility`, and `visual`.
- Use four paired observations, alternating baseline/candidate order. Builds and dependencies MUST be prepared consistently; execution-result reuse MUST be disabled.
- Use `HARNESS_DIAGNOSTICS=basic`, `HARNESS_ACCOUNTING=required`, fresh run IDs, and identical capacity settings. Required accounting prerequisites MUST be satisfied; sampled CPU or RSS MUST NOT silently substitute for charged CPU and peak accounting.
- Sum elapsed time and charged CPU across the same five-suite portfolio for each observation.
- Candidate elapsed time and charged CPU MUST each improve in at least three of four pairs, and each median paired improvement MUST exceed the reporting resolution. No fixed percentage improvement is required.
- Candidate maximum observed peak memory MUST NOT exceed the baseline maximum across those observations. No OOM, leaked resource, timeout increase, reduced selection, or added retry is acceptable.
- Functional failures invalidate acceptance. Repairs require a new frozen candidate and a new comparison set; failed runs MUST remain retained.
- The changed graph/source identities MUST be recorded honestly. The built-in equivalent-run comparison MUST NOT be forced to accept incompatible revisions.

A full-CI comparison MUST also check for displaced cost outside the browser portfolio. A reproducible CI regression blocks acceptance pending diagnosis.

### Canonical command families

Commands MUST run from the repository root through public Make targets:

- Discover exact selections with `make task-guide OWNER=<owner> DETAIL=coverage` and `make target-plan`.
- Validate changed support, backend, and harness rows through exact `make test-slice` selections.
- Validate authored/generated closure through `make generate` and applicable `make source-check CHECK=generated`, `schema`, `catalog`, `identity`, and `surface`.
- Run `make agent-finalize` before broader final verification.
- Run `make check`, `make ci`, and the explicit browser suites required above; visual coverage MUST NOT be inferred from CI.
- Run affected harness lifecycle/regression and UI-review contracts.
- Audit affected owners through `make test-evidence-audit` using explicitly recorded compatible result roots.
- Inspect performance through `make explain-run ... DETAIL=performance`.
- Run `make lint-markdown` after documentation changes.

Exact invocation inputs and run identities MUST be recorded before execution using the [test-operations command recipes](../../.agents/skills/cartulary-test-ops/references/command-recipes.md). Command discovery and static source inspection do not constitute passing verification. `RESULTS_DIR` maintenance MUST identify a successful full warm check root; otherwise the implementation report MUST state that maintenance was skipped because `RESULTS_DIR` was unset.

Goldens MUST NOT be refreshed to conceal a refactor regression. No rendering change is intended.

### Handoff-writing session checks

No Make target was invoked in the handoff-writing session. Inspection of the public Markdown lint path found cache-stamp and retained-output writes outside the sole admitted handoff. Canonical Markdown lint was therefore skipped to preserve the current write boundary; this is not a lint pass. A bounded read-only inspection of the new document checks its structure and local links. Product verification, generation, formatting, and `agent-finalize` are not applicable to this document-only session.

## 9. Top-Level Work Tracker

| Item | Status | Completion gate |
|---|---|---|
| T-001 — Capture exact baseline and coverage mapping | DONE | WF-00 checkpoint E-024; all 88 rows / 97 cases characterized. V-001 successor acceptance remains pending. |
| T-002 — Adopt coordinated harness amendments | DONE | WF-01 E-025/E-026: amendment/mapping approved by the user; documentation validation passed. |
| T-003 — Implement owned browser lifetimes | DONE | WF-02 / E-028; S-001-specific V-002/V-004/V-007 witnesses pass. |
| T-004 — Replace mixed global/browser scenarios | DONE | WF-03 / V-004–006; E-030 records successor closure. |
| T-005 — Promote qualified cohorts and remove obsolete paths | DONE | WF-04 / E-036; ownership, order, exact-case, compatibility and regression qualification passed. |
| T-006 — Measure and close final evidence | IN PROGRESS | WF-05 / V-008–009 |
| T-007 — Save the approved implementation handoff | DONE | This document created; runtime work remains pending. |
| T-008 — Retire public-error fault control | DONE | WF-06 / S-003; E-032 closes retirement and successor gates. |
| T-009 — Integrate specialized lifecycles | DONE | WF-07 / S-005; E-034 records implementation, native results and limitations. |
| T-010 — Complete validation and final handoff | IN PROGRESS | WF-05 / S-006; all V-001–009 |

No item is complete merely because its tests exist or its catalog declarations have changed.

## 10. Session Handoff Log

| Evidence | Observation | Limit |
|---|---|---|
| E-001 — Catalog inspection | 88 isolated rows, 97 selected cases, 26 source files; 86 rows use the default runtime and two use `network_flow_claimed`. | Static baseline only. |
| E-002 — Fixture inspection | Fresh primary actors/contexts exist; secondary-context ownership and gated-route settlement are incomplete. | No runtime failure reproduction performed. |
| E-003 — Backend coverage inspection | Existing last-admin, session-expiry, cursor-expiry, and error-boundary tests support restructuring. | Presence does not establish complete successor coverage or a passing result. |
| E-004 — Harness-owner inspection | Cohort compatibility already exists; global public-error-fault behavior is explicitly normative. | Retirement requires coordinated owner amendment. |
| E-005 — User decisions | Scope is 88 rows plus supporting harness; performance gate is reproducible improvement without a fixed percentage. | The planning invocation authorized no cutover execution. |
| E-006 — Audit workspace inspection | Baseline commit unchanged and working tree reported clean; destination absent during planning. | Historical planning-session observation; no runtime evidence collected. |
| E-007 — Handoff creation, 2026-10-24 | The same commit and clean initial working-tree posture were rechecked. Only this handoff was created from the approved plan. | No implementation, tests, builds, generation, service operations, or canonical lint were run. |
| E-008 — Read-only document inspection | Twelve numbered sections, nine unique requirement IDs, and 33 existing local-link targets were confirmed; no trailing whitespace, tabs, inconsistent table-column counts, or missing final newline were found. Git reported only this new handoff, empty tracked/staged diffs, and the unchanged baseline commit. | These checks do not substitute for canonical Markdown lint or runtime validation. |
| E-009 — Implementation authorization and first checkpoint, 2026-10-24 | User explicitly requested implementation of the complete bounded cutover. HEAD remains `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`; initial tracked/staged state is clean and this handoff is the sole untracked input. Updated this tracker first with accepted F-008/F-009, separate workstreams, S-006, and mandatory checkpoints. | WF-00 is in progress; no implementation slice or runtime result is complete. Historical planning-only restrictions are superseded for this execution. |
| E-010 — Public command discovery | `make help ALL=1` and `make task-guide OWNER=harness.browser DETAIL=coverage JSON=1` both exited 0. Current public surface includes exact slices, live status, required accounting, lifecycle/regression/UI-review contracts and finalization. Catalog enumeration still yields 88 isolated rows and 97 cases. | Discovery establishes selection, not runtime success. |

### WF-00 invocation ledger

All commands below run from `/home/jochi/code/cartulary`. Native session handles are recorded when returned. Each run uses a fresh identity and preserves any failure output.

| Run ID / command | Intended selection | Result root | Native session / outcome |
|---|---|---|---|
| `tr-baseline-support-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-support-20261024-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.cleanup_suite_9036a615b1,harness.browser.boundary_support.sessions_suite_4d7bc21669 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Existing cleanup and session-support characterization; required-accounting admission | `.cartulary/test-results/tr-baseline-support-20261024-01` | Native session `7741`, Make exit 2; `config/configuration_error`, stage `accounting_executable_identity`; no product tests executed. Make rebuilt the native harness after admission. `harness-runtime-build/tool-run-summary.json` confirms successful preparation. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-support-20261024-01 DETAIL=accounting` exited 0 and reported none; no complete resource claim. |
| `tr-baseline-support-20261024-02`: same exact support selection and settings as attempt 01, with `CARTULARY_TEST_RUN_ID=tr-baseline-support-20261024-02` | Authorized retry after successful native preparation; unchanged source and selection | `.cartulary/test-results/tr-baseline-support-20261024-02` | Native session `63317`, Make exit 0; graph passed in 3567 ms. Exact two-row support characterization passes; not browser or full-owner qualification. |
| `tr-baseline-auth-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-auth-20261024-01 make test-slice OWNER=module.auth ROWS=module.auth.accessibility.verify_session_mfa_incident_forbidden_loading_an_e8f5d5f3a9,module.auth.browser.a_reviewer_visible_incident_edit_control_fails_w_b6248bf4c8,module.auth.browser.administrative_audit_read_recovery,module.auth.browser.administrative_audit_real_browsing,module.auth.browser.deployment_admin_revoke_all_invalidates_a_target_a09ed5d06f,module.auth.browser.the_ordinary_shell_keeps_deployment_user_adminis_8a57567508,module.auth.browser.verify_ordinary_login_incident_directory_entry_o_d66ea23dc1,module.auth.browser_stateful.browser_deployment_user_administration_works_on_d6d7ff0b86,module.auth.browser_stateful.idle_expiry_fails_closed_after_the_shared_clock_6525f19d9d,module.auth.browser_support.supplemental_browser_verifies_ordinary_shell_sel_f8a9da1957 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 10 frozen authentication rows / 18 browser cases, including the last-admin and real-expiry baselines | `.cartulary/test-results/tr-baseline-auth-20261024-01` | Native session `33917`, Make exit 0; graph passed in 195885 ms. Exact 10-row / 18-case baseline passed; final-source qualification remains pending. |

| `tr-baseline-incidents-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-incidents-20261024-01 make test-slice OWNER=module.incidents ROWS=module.incidents.browser.a_selected_incident_whose_membership_is_removed_21054ae899,module.incidents.browser.incident_admins_manage_memberships_on_the_ordina_f26145a162,module.incidents.browser.incident_creation_native_keyboard_payload_and_fi_b6b4631f5f,module.incidents.browser.incident_creation_preserves_chosen_navigation_af_24cb9bd4de,module.incidents.browser.incident_creation_uncertain_replay_and_committed_985931d51a,module.incidents.browser.lifecycle_acknowledgement,module.incidents.browser.lifecycle_boundaries,module.incidents.browser.lifecycle_departure,module.incidents.browser.lifecycle_replay,module.incidents.browser.lifecycle_workbook,module.incidents.browser.membership_audit_access_recovery,module.incidents.browser.membership_audit_read_recovery,module.incidents.browser.membership_audit_real_browsing,module.incidents.browser.membership_audit_response_inspection,module.incidents.browser.membership_management_departure,module.incidents.browser.membership_management_live_pages,module.incidents.browser.membership_management_mutation_recovery,module.incidents.browser.membership_management_self_actions,module.incidents.browser.metadata_authority,module.incidents.browser.metadata_departure,module.incidents.browser.metadata_live,module.incidents.browser.metadata_pending_departure,module.incidents.browser.metadata_recovery,module.incidents.browser.the_ordinary_authenticated_landing_screen_create_fa80fbfef5,module.incidents.browser.the_ordinary_shell_exposes_incident_discovery_ra_59c190a9b5,module.incidents.browser.workbook_home_and_incident_default_preferences_c_6591154e0f HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 26 frozen incident rows / cases | `.cartulary/test-results/tr-baseline-incidents-20261024-01` | Native session `49667`, Make exit 0; graph passed in 379003 ms. All 26 selected incident rows passed; no candidate-source claim. |

| `tr-baseline-module-collaboration-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-module-collaboration-20261024-01 make test-slice OWNER=module.collaboration ROWS=module.collaboration.browser.concurrent_edits_to_different_fields_on_the_same_7ee5cad0ab,module.collaboration.browser.live_updates_never_retarget_pending_local_edits_11699e9b2d,module.collaboration.browser.verify_conflict_resolver_actions_submit_public_m_0c77fae932,module.collaboration.browser_stateful.deterministic_route_owned_browser_revocation_sou_a134c3e26b,module.collaboration.browser_stateful.verify_multi_client_live_row_update_presence_anc_4c65a275c1 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 5 frozen rows for `module.collaboration` | `.cartulary/test-results/tr-baseline-module-collaboration-20261024-01` | Native session `4085`, Make exit 0; graph passed in 150498 ms. All five selected cases passed; see E-015 for accounting and cleanup. |
| `tr-baseline-module-evidence-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-module-evidence-20261024-01 make test-slice OWNER=module.evidence ROWS=module.evidence.accessibility.verify_evidence_icon_buttons_blocked_states_erro_0ec80f90e2,module.evidence.browser.timeline_related_role_loss,module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 3 frozen rows for `module.evidence` | `.cartulary/test-results/tr-baseline-module-evidence-20261024-01` | Native session `10287`, Make exit 0; graph passed in 108354 ms. All three Evidence cases passed; see E-016. |
| `tr-baseline-module-networkflow-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-module-networkflow-20261024-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.pagination_recovery HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 1 frozen rows for `module.networkflow` | `.cartulary/test-results/tr-baseline-module-networkflow-20261024-01` | Native session `35594`, Make exit 0; graph passed in 91928 ms. Exact Network Flow case passed; see E-017. |
| `tr-baseline-module-recovery-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-module-recovery-20261024-01 make test-slice OWNER=module.recovery ROWS=module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 1 frozen rows for `module.recovery` | `.cartulary/test-results/tr-baseline-module-recovery-20261024-01` | Native session `10510`, Make exit 0; graph passed in 92579 ms. Restore case passed; see E-018. |
| `tr-baseline-module-reference_data-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-module-reference_data-20261024-01 make test-slice OWNER=module.reference_data ROWS=module.reference_data.browser.administration_async_recovery,module.reference_data.browser.administration_real_lifecycle,module.reference_data.browser.deployment_admin_browser_sessions_expose_referen_8c10ae0a9c HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 3 frozen rows for `module.reference_data` | `.cartulary/test-results/tr-baseline-module-reference_data-20261024-01` | Native session `8760`, Make exit 0; graph passed in 104178 ms. All three Reference Pack cases passed; see E-019. |
| `tr-baseline-module-timeline-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-module-timeline-20261024-01 make test-slice OWNER=module.timeline ROWS=module.timeline.browser.a_reviewer_session_browser_flow_visibly_performs_c3aaf62d89,module.timeline.browser.draft_create_controls HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 2 frozen rows for `module.timeline` | `.cartulary/test-results/tr-baseline-module-timeline-20261024-01` | Native session `6845`, Make exit 0; graph passed in 93736 ms. Both selected Timeline cases passed; see E-020. |
| `tr-baseline-module-workbook-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-module-workbook-20261024-01 make test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.preferences,module.workbook.browser.grid_autosave_closure,module.workbook.browser.ordinary_create_closure,module.workbook.browser.preferences_1,module.workbook.browser.preferences_2,module.workbook.browser.preferences_3,module.workbook.browser.preferences_4,module.workbook.browser.reference_admission_lifecycle,module.workbook.browser.verify_continuous_workbook_shell_composition_top_96ea2f5084,module.workbook.browser_stateful.preferences_extension,module.workbook.browser_stateful.query_continuation_session,module.workbook.browser_stateful.verify_timeline_public_mutations_remount_safe_tr_84f645d6c5,module.workbook.visual.preferences HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 13 frozen rows for `module.workbook` | `.cartulary/test-results/tr-baseline-module-workbook-20261024-01` | Native session `3667`, Make exit 0; graph passed in 267559 ms. All 13 rows / 14 Workbook cases passed; see E-021. |
| `tr-baseline-web-design-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-web-design-20261024-01 make test-slice OWNER=web.design ROWS=web.design.accessibility.administrative_audit_browsing,web.design.accessibility.deployment_user_draft_credential_dialogs_ac3e5d7802,web.design.accessibility.incident_creation_accessible_keyboard_feedback_a_f2a19fbfdb,web.design.accessibility.incident_import_workflow,web.design.accessibility.lifecycle,web.design.accessibility.membership_audit_browsing,web.design.accessibility.membership_management_accessibility,web.design.accessibility.metadata_editing,web.design.accessibility.reference_pack_administration,web.design.browser.account_application_menu_keyboard_navigation_and_8cba739220,web.design.browser.incident_import_action_access_hardening,web.design.browser.incident_import_observation_recovery,web.design.browser.incident_import_real_workflow,web.design.visual.administrative_audit_browsing,web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde,web.design.visual.incident_import_workflow,web.design.visual.lifecycle,web.design.visual.membership_audit_browsing,web.design.visual.membership_management_visual,web.design.visual.metadata_editing,web.design.visual.reference_pack_administration HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 21 frozen rows for `web.design` | `.cartulary/test-results/tr-baseline-web-design-20261024-01` | Native session `23619`, Make exit 0; graph passed in 289915 ms. All 21 design cases passed; see E-022. |
| `tr-baseline-harness-browser-20261024-01`: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-baseline-harness-browser-20261024-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.frontend_artifact_lifetime,harness.browser.boundary_support.timeline_investigation_capture,harness.browser.boundary_support.visual_preference_lifetime HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Exact 3 frozen rows for `harness.browser` | `.cartulary/test-results/tr-baseline-harness-browser-20261024-01` | Native session `71519`, Make exit 0; graph passed in 103733 ms. All three browser-support cases passed; see E-023. |

Documentation validation after baseline reconciliation: `CARTULARY_TEST_RUN_ID=tr-baseline-docs-20261024-01 make lint-markdown` exited 0 (native session `79780`, duration 20050 ms). Summary: `.cartulary/test-results/tr-baseline-docs-20261024-01/adhoc/lint-markdown/tool-run-summary.json`. This is documentation lint, not product or performance acceptance.

### WF-00 assertion and assurance observations

The entries below record inspected assertions, not successor acceptance. Retained cases keep their complete observable assertions unless an explicit obligation-level replacement is recorded. A passing baseline is not evidence for changed source.

| Baseline cases | Inspected obligations and owner | Successor / assurance boundary | Outstanding work |
|---|---|---|---|
| B-001–003 | Testing Harness: immutable loaded JS asset hashes during overlapping normal/measurement publication; rich Timeline capture rejects unready/unsafe normalization and preserves source/history; local visual preferences do not persist across density order/page closure. | B-001 remains dedicated S-005; B-002/003 retain browser assertions under S-004. | Specialized cleanup/failure qualification and compatible runtime results. |
| B-004 (all nine cases), B-005, B-008–010, B-012–013 | Core 01 §3.3.2 and Core 04 §1: anonymous/loading/MFA/setup/forbidden/revoked/error states, keyboard names/focus/contrast, real login and revoke-all, current-role denial, public credential-action authorization, real HTTP expiry and fresh login. | Retain each case; only B-012 keeps clock isolation. Broad local response patterns require exact method/path admission when migrated. | Scope ownership before login, including failed/partial acquisitions; interference proof. |
| B-006/007 | Core 01 §3.3.5.1A: 105 exact actor/action/target events, 100/5 paging, inclusive/exclusive timestamps, cursor-limit mismatch; stale responses cannot replace accepted filters; inert published JSON, recovery and focus. | Same browser cases; real HTTP reads and generated audit mutations remain. | Cancel owned gates and verify unrelated deployment audit residue. |
| B-011 | Core 01 §3.3.5.1 / Core 04 administration: real user create/patch/version conflict, native departure protection, retained drafts and history navigation; last-admin demotion and deactivation presentation. | S-002 retains owned CRUD/navigation and local guard presentation. Real HTTP successors must prove both guard paths, unchanged state and still-valid sessions. | Existing store-handler tests prove both guards/state/no hub revocation; existing HTTP integration proves demotion status only. Deactivation plus HTTP durable/session witnesses are an explicit gap, not already satisfied coverage. |
| B-014–018 | Core 03 §§3–4 / Core 01 §3.3.10: real different-field merge, all three same-field choices, fresh conflict after stale resolution, FIFO pending replay, live tombstone/restore invalidation, record/field anchoring under virtualization, legacy conflict-token rejection/recovery; owned revoke-all/logout and real idle expiry. | Retain distinct browser obligations; split B-017 into three cases. Deduplicate replay only after equivalent trigger/pending-work assertions are mapped. | Scope all secondary contexts and transport gates; actual socket expiry remains isolated. |
| B-019–021 | Core 01 §16 / Core 03 §8 / Core 04 evidence access: real uploads, preview/download, closed-incident readable handles, accessible state/announcement/focus/contrast, visual states; role loss between Evidence commit and Timeline link preserves accepted Evidence without unauthorized link/history change. | S-003 replaces only server-fault presentation with exact local responses; real workflows and B-020 two-commit authorization remain. | Real serializer witness and complete fault retirement; cancellation of pending issuance after dismiss. |
| B-022–026, B-045/046 | Core 01 §3.3.5.3: creation payload/normalization/uniqueness, native keyboard validation, exact replay after committed response loss, no duplicate dispatch, delayed success respects account navigation, stale membership selection clears while session remains; deep-link and promoted-field persistence. | Same cases with explicit directory navigation and scope barriers. | Retain owned incident identities and all real request/state witnesses. |
| B-023, B-032–039 | Core 01 §§3.3.5.1/3.3.5.1A: real membership paging/add/patch/delete/version/audit, non-admin denial, self-demotion/removal on closed incidents; exact replay versus confirmed read failure, explicit conflict review, sequential departure, accepted-query paging, timestamps, scope validation and inert additive JSON. | Same cases; real and local presentation layers remain distinct. | Secondary contexts and gates; no session/route survives test end. |
| B-027–031 | Core 01 §3.3.5.3.2 / Core 03 §4.3.2: close/reopen replay cannot overwrite newer accepted state; uncertainty versus confirmation; Unicode bounds/NFC and conflicts; ordered departure; closed workbook retains copyable drafts without auto-resubmission after reopen. | Retain five cases and their real lifecycle/persistence witnesses. | Cancel barriers without dispatching held writes; owned cleanup. |
| B-040–044 | Core 01 §3.3.5.3.1: sparse normalized patch/null clear/no-op, real stale-version conflict, explicit observed-version review, newer draft protection, authority loss and hidden/session cleanup, pending completion fenced across departure/other incident. | Retain five cases; local presentation remains separate from real patch/history evidence. | Remove swallowed fulfillment failures; settle route owners together. |
| B-047, B-055–065, B-067 | Core 01 §3.3.5.2 / Core 03 workbook: base/saved/extension identity preferences and clears, startup fallback, no unwanted navigation/query side effects, independent observations, uncertainty review, personal isolation/viewer denial on closed incidents; original-cell and reference drafts, exact create replay; density/layout/accessibility/visual assertions. | Retain cases. B-065 uses local expiry plus real replacement login/paging; old protected rows/cursors must disappear. | Retain separate Network Flow runtime; scope secondary actors/gates; actual expiry remains covered by B-012/B-017. |
| B-048 | Network Flow Activity §§13/13.5: real 200-row and 500-contributor pages; failed next/previous retain committed page; retry/refresh, single-admission double click, inspector/focus and graph retention; 900-second expiry restarts forward/backward on all three routes. | Shareable three-route real paging/local recovery; dedicated real rows-expiry bridge; backend boundary/scope witnesses on all three routes. | Inspected real HTTP coverage proves expiry at 900 seconds (valid at 899), restart, final-page termination and rename stability on all three routes. Codec tests prove actor mismatch, tamper and key retirement, but the complete HTTP route/actor/session/incident/table/filter/sort/limit mismatch matrix remains an S-002 gap. Old combined browser coverage cannot be removed first. |
| B-049 | Core 01 §12.2 / Core 04 §9.14: retained backup identity and restored incident, restored authenticated workbook and real query of recovered source content. | Dedicated S-005 with disposable databases, runtime roots and processes. | Browser must close before restored runtime; target sessions never use ordinary provisioner cleanup. |
| B-050–052 | Reference Pack owner: real signed bytes and exact replay, staged import, activate/disable/reverify, selected/all refresh and safe validation summary; local asynchronous acknowledgment/selection, late list fencing, job cancel/progress and current-admin loss. | Preserve dedicated real lifecycle; owned presentation cases keep keyboard/recovery combinations. | Feature jobs must settle on acquisition/body failure, including admission before response delivery. |
| B-053/054 | Core 02 Timeline / Core 03: reviewer review → material-edit demotion → supersession with exact version/reason/replacement; Add row and canceled file choice create nothing, viewer/closed admission forbids create/attach. | Same real browser cases. | Own reviewer/viewer contexts before exposing pages. |
| B-066 (both cases) | Core 01 public mutation/envelope and Core 03 §§3–4: rough creation excludes server-owned fields; exact PATCH receipt replay, divergent transaction and invalid change-set rejection, scalar/tabular paste with scope validation and grouped conflict, authored unparseable dates, remount-safe IDs, explicit blocker rekey/discard, fresh same-field resolver, real revoke-all FIFO replay and unknown-error presentation. | Preserve real mutation/recovery coverage for both cases; S-003 replaces only the unknown-error injection and separately verifies serialization. | Own all temporary contexts/controllers; no mocked response receives credit for real persistence/replay. |
| B-068–088 | Core interaction owners plus design accessibility/presentation: native form/dialog keyboard behavior, visible focus and restoration, named controls, narrow/zoom/text spacing, exact errors and uncertainty/confirmation states; existing visual captures across density. Import cases additionally prove exact multipart replay, one job identity, current authority/profile checks and explicit workbook launch. | Keep visual/accessibility layers; local presentation does not establish backend authorization/job enforcement. | Selected test bodies and their presentation/import helpers inspected; baseline execution remains pending. Final unchanged visual qualification and owned job settlement remain required. |

E-012: `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-auth-20261024-01 DETAIL=performance JSON=1` exited 0. Native command exit 0 was independently collected. Terminal receipt reports passed retained-secret scan, complete execution and empty invocation cgroup settlement. Invocation duration: 202316.148463 ms; charged CPU: 359275382 µs; memory peak: 11332182016 bytes. Source digest: `sha256:1910139510676bbc33f9759c09f8e259dabc402f3bfbf15ab7fbff9d212b8fae`; graph digest: `sha256:0232f251f2509dc8976df5e07a33ebdeda95dfddfae7d72de26bc0d1a7e31b93`. These are characterization results, not a completed four-pair comparison.

E-013: Direct changed-helper name search found consumers in 34 TypeScript files, including files outside the frozen 26-file inventory. S-001 must trace transitive consumers and migrate affected helper calls together; those consumers do not become additional baseline rows. Existing `holdBrowserRequest.dispose()` releases a held request and suppresses one route-error string. Existing `OwnedSessionTracker.cleanup()` consumes failure on repeated calls. Neither behavior meets the new cancellation/completion contract.

E-014: Incident characterization `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-incidents-20261024-01 DETAIL=performance JSON=1` exited 0, following independently collected Make exit 0. Execution is complete, retained-secret scan passed, and cgroup settlement is empty. Duration: 385030.750037 ms; charged CPU: 675690966 µs; memory peak: 6086696960 bytes. Same source digest as E-012; graph digest: `sha256:6deeb4f4e5b0feff00215f16a96a09878b035bbef4c65f289cdcf741f6456b53`. No product or cleanup failure reported. Baseline evidence only.

E-015: Collaboration characterization completed with independently collected Make exit 0 (native session `4085`, graph 150498 ms). `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-module-collaboration-20261024-01 DETAIL=performance JSON=1` exited 0: execution complete, retained-secret scan passed, settlement empty; duration 155655.055929 ms, charged CPU 235053100 µs, peak memory 6700339200 bytes. Source digest matches E-012; graph digest is `sha256:adc1fc8b9b115a388de8ef9e7430a6c1abde1bbe1696c7c49049f5d7867f0154`. All five selected cases pass; no product or cleanup failure reported. Evidence characterization was then launched as native session `10287` using its exact recorded command; its outcome is pending.

Characterization environment: Linux host exposes 24 logical CPUs and 32571412 KiB total memory; default browser lane count is four. Authored pin registry declares Go 1.27.3, Node 24.15.0, pnpm 10.33.0, Playwright 1.63.0 and axe-core 4.13.0. Commands use `CGO_ENABLED=0`, required accounting, basic diagnostics and disabled execution-result reuse. These observations are baseline context; the later paired protocol must freeze and verify both environments independently.

E-016: Evidence characterization completed with independently collected Make exit 0. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-module-evidence-20261024-01 DETAIL=performance JSON=1` exited 0: complete execution, no evidence absences, passed retained-secret scan and empty settlement. Duration 113306.78353 ms; charged CPU 186362658 µs; peak memory 6374154240 bytes. Source digest matches E-012; graph digest `sha256:36774b7b9fa44dc4f7e33d6d903ee94e121c68882c5d77b9761c9b6355f2e23d`. No product or cleanup failure reported. Network Flow characterization then launched with its recorded command as native session `35594`; outcome pending.

E-017: Network Flow characterization completed with independently collected Make exit 0. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-module-networkflow-20261024-01 DETAIL=performance JSON=1` exited 0: complete execution, no evidence absences, passed retained-secret scan and empty settlement. Duration 96755.446036 ms; charged CPU 119094363 µs; peak memory 2531696640 bytes. Source digest matches E-012; graph digest `sha256:bb6a11c9247bb361bf1f66260387ee8646259014c9fceffdd9622a4bc5ebe60a`. No product or cleanup failure reported. Restore characterization then launched as native session `10510`; outcome pending.

E-018: Restore characterization completed with independently collected Make exit 0. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-module-recovery-20261024-01 DETAIL=performance JSON=1` exited 0: complete execution, no evidence absences, passed retained-secret scan and empty settlement. Duration 97435.718873 ms; charged CPU 122651941 µs; peak memory 3359072256 bytes. Source digest matches E-012; graph digest `sha256:9d1f7107e19ca297d9d2464ea2d5398060bc80bb36807560be9384e7c5d11ebc`. No product or cleanup failure reported. Reference Pack characterization then launched as native session `8760`; outcome pending.

E-019: Reference Pack characterization completed with independently collected Make exit 0. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-module-reference_data-20261024-01 DETAIL=performance JSON=1` exited 0: complete execution, no evidence absences, passed retained-secret scan and empty settlement. Duration 109026.190941 ms; charged CPU 159493325 µs; peak memory 5883432960 bytes. Source digest matches E-012; graph digest `sha256:824ab86c4f0823235dad56b00c4e9eb54a4f3ccb583950ffa0f53a5e970e26a2`. No product or cleanup failure reported. Timeline characterization then launched as native session `6845`; outcome pending.

E-020: Timeline characterization completed with independently collected Make exit 0. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-module-timeline-20261024-01 DETAIL=performance JSON=1` exited 0: complete execution, no evidence absences, passed retained-secret scan and empty settlement. Duration 98501.931914 ms; charged CPU 135678324 µs; peak memory 4270014464 bytes. Source digest matches E-012; graph digest `sha256:4e81c9a6dbf986ab4c8cc0ac2736a4913d794f1beeb4d7a602c7b04cf093abd0`. No product or cleanup failure reported. Workbook characterization then launched as native session `3667`; outcome pending.

E-021: Workbook characterization completed with independently collected Make exit 0. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-module-workbook-20261024-01 DETAIL=performance JSON=1` exited 0: complete execution, no evidence absences, passed retained-secret scan and empty settlement. Duration 273310.061371 ms; charged CPU 536068876 µs; peak memory 6501425152 bytes. Source digest matches E-012; graph digest `sha256:ba399d8373da9622f3aa7ec88658dad92456592e7ce789d7701459ae79a3ef3d`. No product or cleanup failure reported. Design characterization then launched as native session `23619`; outcome pending.

E-022: Design characterization completed with independently collected Make exit 0. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-web-design-20261024-01 DETAIL=performance JSON=1` exited 0: complete execution, no evidence absences, passed retained-secret scan and empty settlement. Duration 295783.134329 ms; charged CPU 576803915 µs; peak memory 5891825664 bytes. Source digest matches E-012; graph digest `sha256:9947c3f9763ce3e86d7cffefec146cc9a9c957cc2497779d511647d38baa9671`. No product or cleanup failure reported. The three frozen browser-support cases then launched as native session `71519`; outcome pending.

E-023: Browser-support characterization completed with independently collected Make exit 0. `make explain-run RESULTS_DIR=.cartulary/test-results/tr-baseline-harness-browser-20261024-01 DETAIL=performance JSON=1` exited 0: complete execution, no evidence absences, passed retained-secret scans for the outer graph and both independent frontend-publication builds, and empty invocation settlement. Duration 108825.345187 ms; charged CPU 238360249 µs; peak memory 8500154368 bytes. Source digest matches E-012; outer graph digest `sha256:c2f122afe2f752a4813df08d8295ef54a1f49a6f3beb8c09839390e131713e0c`; nested build graph digests `sha256:a144f031bc22e28db53cd644d5db0f9f740fb737cd11bf60daa83f4916a0e9e0` and `sha256:8f57d2948cb8581fae3c2ce1ac5026c0b9865e440ad61358bb6fe91ffb4dee98`. No product or cleanup failure reported.

### WF-00 multiple-case and replacement accounting

Each scenario listed under B-001–088 is a distinct baseline obligation bundle. A retained case keeps every inspected assertion in that bundle unless the following table explicitly assigns an assertion to a different successor. The owner/assurance table above supplies its requirement boundary. Neither a shared row nor a shared file can discharge another case's bundle. Candidate case identifiers and compatible results must replace the planned successor descriptions during the implementing workstream; none are qualified yet.

| Baseline bundle | Inspected obligation and planned successor | Assurance / completion gate |
|---|---|---|
| B-004 `scenario_24af174d2595` | Anonymous session-required response exposes keyboard login; real login reaches the authenticated landing. Retain this case. | Browser; S-001 ownership and S-004 sharing. |
| B-004 `scenario_518ccc2daf28` | Authenticated account/admin/incident/retry controls remain named, reachable and visibly focused. Retain this case. | Browser accessibility; S-004. |
| B-004 `scenario_6b62ae5a8f33` | Deferred session loading presents progress and reachable recovery. Retain with cancellable response barrier. | Browser accessibility plus scope cancellation; S-001/S-004. |
| B-004 `scenario_7e9876d4d78d` | Local forbidden envelope announces denied access, recovery and safe public details. Retain exact local interception. | Browser presentation, not authorization enforcement; S-004. |
| B-004 `scenario_9e9180c613d0` | Local generic error renders safe diagnostics and keyboard recovery. Retain without disclosing private fields. | Browser presentation; serializer evidence remains separate. |
| B-004 `scenario_b631edcb35d1` | Real empty/list/selected incident transitions, stale selection and current-role denial retain keyboard recovery and reject unauthorized writes. Retain owned incidents/users. | Browser and real HTTP admission; S-001/S-004. |
| B-004 `scenario_c05759d45cdc` | Real TOTP challenge and completion preserve naming, focus, announcements and secret handling. Retain actor ownership before login/enrollment. | Real authentication plus browser accessibility; S-001/S-004. |
| B-004 `scenario_d9203af9002f` | Setup-required enrollment remains keyboard reachable without prematurely granting a session; invalid enrollment shows only safe diagnostics. Retain. | Real authentication plus browser accessibility; S-001/S-004. |
| B-004 `scenario_fcec515eb61b` | Real revocation after prior authentication announces session end and permits fresh login. Retain with real session revocation witness. | Browser and real HTTP; S-001/S-004. |
| B-011: owned administration | Preserve real user creation/editing/version conflict and draft/history navigation in the existing canonical case. Replace only last-admin presentation with exact local responses. | Browser interaction and real owned CRUD; S-002. |
| B-011: last-admin enforcement | Add real HTTP demotion and deactivation rejection with unchanged durable user version/state and continued validity of the protected administrator's sessions. Store-handler witnesses alone do not close this obligation. | Auth integration; must pass before old global browser setup is removed. |
| B-012 | Keep actual shared-clock HTTP expiry, denied access and fresh login in the compact isolated case. | Real HTTP/browser expiry bridge; S-002/S-004. |
| B-017: three triggers | Split revoke-all, logout and idle expiry into independent cases, each preserving pending edit admission, recovery and replay assertions. Revoke-all/logout use owned users; only actual socket/session expiry retains clock isolation. | Real auth/collaboration plus browser recovery; S-002. |
| B-048: ordinary paging/recovery | Preserve accepted pages, rows/diagnostics/contributors, failed next/previous retention, explicit retry/refresh, single admission, focus/inspector and graph retention. Local expiry response matrix becomes shareable. | Browser recovery with real paging; S-002. |
| B-048: expiry/scope | Keep one real rows-route expiry browser bridge. Retain real backend 899/900-second boundary/restart/rename tests and add missing scope mismatches for all three routes. | Real HTTP and one isolated browser bridge; S-002. |
| B-051 | Keep real signed import and exact replay, activation, disablement, re-verification, selected refresh and refresh-all. Observe each admitted real job to a terminal state even after body failure. | Dedicated Reference Pack integration workflow; S-002. |
| B-050/B-052 and B-076/B-088 | Preserve local asynchronous acknowledgment, selection, job progress/cancellation, identity/late-response fencing, current-admin loss and keyboard/visual presentation without mutating the shared registry. | Browser presentation; no credit for registry enforcement. |
| B-065 | Replace clock control with a local expired-session response; retain real replacement-account login and paging. Assert old protected rows and continuation tokens disappear before replacement state is accepted. | Browser identity/continuation fence; real expiry remains B-012/B-017. |
| B-066 `scenario_481272c41da9` | Retain remount-safe fresh transaction IDs, explicit blocker retry/rekey from latest version, discard without writes and subsequent real same-field resolver handoff. | Real Timeline mutations plus browser recovery; S-001/S-004. |
| B-066 `scenario_da2d76b86dc8` | Retain rough create, sparse/versioned PATCH, exact receipt replay, malformed/divergent transactions, scalar/tabular paste, foreign scope rejection, grouped conflicts, authored invalid dates, remount persistence and real revoke-all FIFO replay. Replace only the unknown-error presentation response. | Real HTTP/persistence/replay plus local presentation; S-003. |
| B-019/B-021 and B-066 public errors | Replace the three server-fault consumers with exact method/path responses in the owning browser context. Preserve real Evidence upload, preview, download and mutation. Separately strengthen serializer status/code/details/request-correlation/private-diagnostic checks. | Browser presentation versus real HTTP serializer; S-003. |
| B-001/B-049 | Preserve immutable loaded-asset overlap and retained-backup/restore workbook-query assertions with their disposable runtimes and process evidence. | Specialized lifecycle; S-005. |

All other B bundles retain their canonical case and complete inspected assertion set. Consolidation is permitted only after an obligation-by-obligation equivalence record identifies the surviving canonical identity and records retired identities through existing migration accounting. No consolidation is presently accepted.

### E-024 — WF-00 completion checkpoint

**Status: COMPLETE.** Recorded before beginning WF-01. Source revision remains `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`; only this controlling tracker is changed. The frozen inventory remains 88 rows, 97 selected cases and 26 files, with 86 default-runtime rows and two `network_flow_claimed` rows. All 97 cases passed their exact baseline selections with native exit 0. E-012 and E-014–023 retain the commands, roots, source/graph bindings and charged resource observations; E-011 and the Section 10 tables record every case disposition and its assurance boundary. Changed-helper discovery includes 51 direct consumers plus the separately classified suite-admin acquisition helper.

Substantive work: reconciled authored selection against the frozen commit; inspected selected bodies and consequential lifecycle/transport/feature helpers; separated production enforcement from presentation obligations; identified missing last-admin HTTP durable/session witnesses and the full Network Flow HTTP cursor-scope matrix; recorded complete fault retirement and specialized process boundaries. No case was removed or promoted, and no production, test, schema or generated source changed.

Validation: all eleven exact browser owner selections passed, as did the two selected cleanup/session-support rows and `make lint-markdown`. Required-accounting terminal receipts show passed retained scans and empty invocation cgroups; this does not prove per-test ownership inside a shared cohort. The first support invocation failed native executable-identity admission after preparation rebuilt the harness; it ran no product tests, remains retained, and was followed by the documented successful fresh invocation. No baseline product or cleanup failure was reported. Discovery used `make help ALL=1`, `make task-guide OWNER=harness.browser DETAIL=coverage JSON=1`, and the same task-guide command for each product/reporting owner listed in the invocation ledger. These results are characterization only, not the four-pair comparison or final acceptance.

Coverage dispositions: retain canonical cases except the explicitly mapped splits/local-presentation replacements. Both multi-case rows are accounted for individually. No consolidation is accepted yet. Existing backend coverage gaps must close before their predecessor browser assertions are removed.

Remaining risks: common ownership and cancellation have not been implemented or failure-qualified; all baseline browser rows still run isolated; owner amendments are unreviewed; real successor enforcement, order/interference, retirement and performance evidence remain absent. Required final checks are pending their phase, not passed or waived. No eligible full warm-check `RESULTS_DIR` is available; retained-run maintenance has not run.

**Next entry gate:** WF-01 may prepare the owner/interface amendment and authored-projection mapping. WF-02/S-001 must not begin until the human owner-review gate closes. The complete cutover remains one merge/rollback unit.

### E-025 — WF-01 owner amendment and projection review packet

**Workstream checkpoint: BLOCKED on human review (RB-006).** Owner amendment, interface/projection mapping and documentation validation are complete; no dependent workstream has begun. Fidelity review retained the existing AC-088 scheduling/reset/accounting obligations while adding scope semantics. Source base remains `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`; the proposed working-tree owner change is in `docs/testing-harness-nlspec.md`. No implementation slice has begun. The amendment is proposed for the coordinated cutover, not evidence that its new behavior exists.

| Owner amendment | Concrete downstream projection / implementation | Assurance boundary and completion evidence |
|---|---|---|
| REQ-806 / AC-088: one test resource owner, admission, effective context settings and in-flight acquisition | S-001 replaces competing lifetimes in `apps/web/e2e/fixtures.ts` and `support/runtime/cleanup.ts`; existing cleanup/session-support suite rows retain their owner attribution and gain partial/late-acquisition and inherited-settings cases. All Section 2 consumers migrate together. | V-002/V-004: fail immediately after actor creation/login/context/request acquisition, reject new admission during disposal, close late contexts, invalidate owned sessions while unrelated/provisioner sessions survive. |
| REQ-806 / AC-088: barrier terminal outcomes and shared disposal completion | `OwnedBrowserScope.barrier()` exposes released/cancelled outcomes; `dispose()` returns one promise retaining its rejection. S-001 migrates feature gates and transport helpers, removing release-on-dispose and swallowed route failures. | V-002/V-007: cancellation before/during interception, no delayed mutation dispatch, failed fulfillment, concurrent drains, repeated rejected disposal, primary/cleanup failure preservation and credential-safe diagnostics. |
| REQ-806 / AC-088: feature work and bounded residue | Feature fixtures retain job/data ownership and settle in-flight admission plus admitted jobs before route/context retirement. S-002/S-004 scope lists to owned identities, navigate required surfaces explicitly and use deterministic identities once per generation. | V-003/V-005: unrelated residue, terminal jobs, separate clean forward/reverse invocations and exact execution accounting. |
| REQ-806 / AC-088: assurance separation | Existing product-owner verification contracts remain the authority projections for auth, collaboration, Network Flow, Reference Packs, Timeline, Evidence and workbook behavior. S-002 adds real last-admin and cursor-scope witnesses and records each split in authored family selectors before retiring predecessor assertions. | V-001/V-004–006: real authorization/state/session/expiry/persistence/replay witnesses plus separate browser recovery/focus/accessibility/visual assertions. |
| REQ-481 / AC-035: public-error control retirement | S-003 deletes the exact source/schema/attachment dependencies in Section 2 and replaces the six obsolete control tests selected by `harness.browser.unit.public_error_fault_route_disabled_by_default_0910c3af84` with explicit retirement accounting and a route-absence witness. Existing HTTP serializer coverage is strengthened under its current owner. | V-005/V-006/V-009: negative route witness with controls disabled/enabled; actual serializer status/code/details/request correlation/private suppression; all three local browser successors. |
| REQ-806 / AC-088: specialized lifetimes | S-005 uses the common browser cleanup boundary with the existing restore and publication process owners. Restored sessions never enter ordinary-provisioner revocation. Publication overlap remains a real independent build and immutable loaded-asset assertion. | V-002/V-005: startup/body/interruption/cleanup failures, browser-before-runtime closure, exact owned process/database/runtime settlement and retained evidence. |
| REQ-805/806/678 and AC-088: existing cohort architecture | S-004 changes only qualified authored `browser_state_scope` declarations and exact successor selectors in `tools/test_families/**`; canonical survivors and retired identities use `tools/test_catalog_row_migrations.json`. Existing compatibility, bounded packing, reset, admission and accounting stay authoritative. | V-003/V-007: every selected case qualifies, incompatible tuples separate, failure stops/quarantines, no retries/deadline increase, no missing/duplicate execution. |
| Schema inventory and REQ-450 route list | The proposed owner removes the retired schema/route from active inventories. S-003 removes `tools/harness_schema_attachments.json`'s corresponding entry and authored schema; S-004 regenerates catalog projections from authored selections through Make. | Generated/schema/catalog/identity/surface checks and explicit retirement search in S-003/S-004/S-006; no executable consumer reads Markdown. |

The existing `harness.browser.verification.support_contract` and `source_contracts` already admit the required evidence kinds; this change needs no parallel verification registry, new public command, scheduler, state-scope enum or product API/schema migration. Authored executable changes belong to their separately gated implementation slices; this packet specifies their proposed owner fidelity without making premature catalog or generated changes.

The private interface remains the S-001 contract: `registerActor(userId)`, `newContext(options)`, `ownRequestContext(context)`, `registerSession(snapshot)`, `barrier()`, and `dispose()`. Acquisition registration must include pending promises before resources are exposed. Feature-specific settlement uses feature-owned callbacks/fixtures; it must not transfer feature policy into the common scope. Independent request clients and borrowed control-plane clients must remain distinguishable. Calling a retained helper name is permissible only if it uses this single ownership model, with no independent legacy cleanup path.

Compatibility and migration: deliberate private-helper signature changes across all consumers; removal of the test-only fault endpoint/schema without aliases; exact case splits/migrations; no production storage migration. REQ-454 is retained only as an explicitly retired historical identifier, while new REQ-481 owns retirement. AC-035 retains token/clock edge coverage and AC-088 now closes scope-lifecycle obligations. REQ-464's separate clock process-replacement rule and all Network Flow controls remain intact. Merge/rollback covers the whole cutover.

Review needed: confirm the owner wording, assurance-layer mapping, lifecycle/feature boundary and planned authored projections. The user's plan explicitly blocks implementation until this review gate closes; root `AGENTS.md` also assigns projection-fidelity review to humans. No human review is recorded yet. S-001 remains BLOCKED on RB-006.

Validation: `CARTULARY_TEST_RUN_ID=tr-owner-docs-20261024-01 make lint-markdown` exited 0 (native session `8995`, 10160 ms), summary `.cartulary/test-results/tr-owner-docs-20261024-01/adhoc/lint-markdown/tool-run-summary.json`. A subsequent fidelity edit explicitly preserved the existing AC-088 scheduling/reset/accounting acceptance clauses and AC-035 weak-token negative case. Final documentation check `CARTULARY_TEST_RUN_ID=tr-owner-docs-20261024-02 make lint-markdown` exited 0 (native session `67655`, 10070 ms), summary `.cartulary/test-results/tr-owner-docs-20261024-02/adhoc/lint-markdown/tool-run-summary.json`. No primary or cleanup failure occurred in these documentation checks. Baseline coverage dispositions remain unchanged; no successor has been runtime-qualified. The next entry gate is explicit human review of this amendment/mapping, then the WF-01 completion checkpoint before S-001. Generation, `agent-finalize`, broader final product checks, owner evidence audits and the four-pair comparison are pending the later implementation/final slices. Retained-run maintenance is not run because no compatible successful full warm-check `RESULTS_DIR` is available.

### E-026 — WF-01 completion checkpoint and human approval

**Status: COMPLETE.** The user explicitly answered “Approve owner amendment and mapping” to the E-025 review request on 2026-10-24. This closes RB-006 before WF-02 begins. Source base remains `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`; changed files are this tracker and `docs/testing-harness-nlspec.md`. The substantive owner changes, exact projection mapping and passing lint commands/result roots are recorded in E-025. No production/test/generated source has changed, and no primary or cleanup failure occurred in owner-document validation.

Coverage dispositions remain the reviewed baseline mapping; the scope and retirement rules now have an approved owner amendment for the coordinated cutover. Implementation/runtime fidelity and all candidate acceptance evidence remain to be established. Remaining risks are partial acquisition, delayed work cancellation, feature/session boundaries, helper migration outside the frozen baseline, and specialized lifecycle ordering. No baseline result is promoted to candidate evidence.

**Next entry gate:** WF-02/S-001 is READY. Implement the common scope and migrate changed-helper consumers; run V-002/V-004 and relevant V-007 checks. Update this tracker with their actual outcomes before WF-03. The complete change remains one merge/rollback unit.

### E-027 — S-001 implementation progress (not a completion checkpoint)

WF-02/S-001 is IN PROGRESS; dependent workstreams remain blocked. The approved E-026 gate preceded all implementation edits. Source base remains `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1` with uncommitted candidate changes. The common scope now owns primary/secondary contexts, independent clients, actors and session snapshots; it records in-flight acquisitions, cancels barriers, settles feature callbacks, drains routes, closes contexts, revokes/verifies sessions, and disposes independent clients. Session helper facades delegate to that scope. The old tracker and context-cleanup implementation are removed. Inline gate discovery expanded the migration beyond the original 51-file named-helper search; these are support-consumer changes, not an enlarged frozen baseline.

New support witnesses cover late acquisition, concurrent/repeated disposal, cancellation before/during interception, route failure, drain ordering and redacted errors. Existing fixture-state coverage is being extended with real context-setting inheritance, cancelled mutation, uncaptured-login revocation, secondary closure and unrelated control-plane session checks. Test presence does not establish runtime qualification. Feature job admission/settlement, the complete consumer audit, real browser witnesses, affected-owner runs, cleanup-failure quarantine and primary-failure preservation still require closure before S-001 can complete.

Execution ledger (all commands from the repository root; no product retries):

| Command | Native result | Evidence and disposition |
|---|---|---|
| `CARTULARY_TEST_RUN_ID=tr-s001-types-20261024-01 make lint CHECK=types` | Session `86126`, exit 2 before TypeScript. | Runtime-plan configuration failure; only harness-runtime-build summary exists under `.cartulary/test-results/tr-s001-types-20261024-01`. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-types-20261024-02 make lint CHECK=types` | Session `3134`, exit 2 before TypeScript. | Same planning failure. `make target-plan TARGET=lint INPUTS=/tmp/cartulary-tr-types-input.json DETAIL=graph JSON=1` identified the stale `OwnedSessionTracker` title selection. Authored cleanup-suite titles were reconciled while retaining its row ID. The CGO setting did not repair that selection error. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-types-20261024-03 make lint CHECK=types` | Session `97153`, exit 2; graph 21634 ms. | `.cartulary/test-results/tr-s001-types-20261024-03`: unused imports/fixture parameter and a mock signature error; cleanup 11/11 successful. This intermediate run overlapped subsequent source edits and is diagnostic only. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-types-20261024-04 make lint CHECK=types` | Session `45466`, exit 2; graph 22521 ms. | `.cartulary/test-results/tr-s001-types-20261024-04`: gate migration type/control-flow errors, subsequently repaired; cleanup 11/11 successful. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-types-20261024-05 make lint CHECK=types` | Session `49291`, exit 2; graph 21035 ms. | `.cartulary/test-results/tr-s001-types-20261024-05`: one unused import, subsequently removed; cleanup 11/11 successful. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-support-20261024-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.cleanup_suite_9036a615b1,harness.browser.boundary_support.sessions_suite_4d7bc21669 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Session `49567`, exit 0; graph 3523 ms. | `.cartulary/test-results/tr-s001-support-20261024-01`: both selected rows passed, execution complete, retained scan passed, settlement empty. Source `sha256:895e805e852a7c6e2d2c5288e3da28fee11aff5a5ffad4f6946a1bdcc65e21cb`; graph `sha256:cf2eee8c3b51bc16925b3dfccfd166c06d2646f7a2f04a7cf647c5bad949274b`. Subsequent formatting/repairs require compatible final-source validation. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-format-20261024-01 make format` | Session `84493`, exit 2; graph 8142 ms. | `.cartulary/test-results/tr-s001-format-20261024-01`: formatted 80 files; a promise truthiness lint error and unused/constant-binding warnings remained. Cleanup 11/11 successful. Repairs are in progress. |


| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-format-20261024-02 make format` | Session `10547`, exit 0; graph 7842 ms. | Formatting and earlier lint repairs passed; superseded by later consumer changes. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-types-20261024-06 make lint CHECK=types` | Session `25703`, exit 0; graph 22039 ms. | Type checking passed before the route-settlement repair. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-ownership-20261024-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_a_a_577b9aa596,harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_a_b_e9fdc93f9d,harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_b_a_a194a3225b,harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_b_b_ab488bf040 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Session `37648`, exit 2; graph 93862 ms. | `.cartulary/test-results/tr-s001-ownership-20261024-01`: one failed case, three early-stop skips, zero retries; infrastructure cleanup 23/23 passed. Native selective unroute released interception before the cancelled handler finished aborting, producing “Route is already handled”; the controller and scope retained cleanup failures. Source `sha256:80ecf43e8311dd24247dd34626a20eccf4f7b74c4e41bba3e53c4f726edb8078`, graph `sha256:a3687a85471b1317e3114d04cbc587ab8cbff203d479b214cfaaf35ade599df8`. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-format-20261024-03 make format` | Native session/exit were not retained across context compaction; **not acceptance evidence**. | Exact-root `make test-run-status RESULTS_DIR=.cartulary/test-results/tr-s001-format-20261024-03 JSON=1` reported terminal pass, source `sha256:7e7afd85eef745b78382eebeb7fbd41ba25e16bd27de453db3ce9ccfaa8a510b`. The reader explicitly reports outer exit unobserved. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-types-20261024-07 make lint CHECK=types` | Session `58390`, exit 0; graph 22508 ms. | Type checking passed after selective route settlement was repaired. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-ownership-20261024-02 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_a_a_577b9aa596,harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_a_b_e9fdc93f9d,harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_b_a_a194a3225b,harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_b_b_ab488bf040 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Session `31539`, exit 0; graph 86373 ms. | `.cartulary/test-results/tr-s001-ownership-20261024-02`: four cases passed; context inheritance/override, cancelled mutation without dispatch, secondary closure, uncaptured-login revocation and unrelated/provisioner session survival witnessed. Cleanup 23/23 passed. Source `sha256:a23eb8eea0548ece689e1418b2d1b36b95b01644c46baa8665b4c8c442ca1804`; graph unchanged from ownership-01. Later feature and route migration changes require compatible validation. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-format-20261024-04 make format` | Session `26427`, exit 0; graph 7954 ms. | Formatting passed for the wider route/finalizer migration. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-types-20261024-08 make lint CHECK=types` | Session `70507`, exit 2; graph 20855 ms. | `.cartulary/test-results/tr-s001-types-20261024-08`: Playwright route registration returns a Disposable; one gate cancel binding and one imported alias needed repair. Cleanup 11/11 passed. These source errors were repaired before the next invocation. |
| `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-support-20261024-02 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.cleanup_suite_9036a615b1,harness.browser.boundary_support.sessions_suite_4d7bc21669 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Session `29760`, exit 0; graph 3648 ms. | `.cartulary/test-results/tr-s001-support-20261024-02`: cleanup and session support passed. Later context feature settlement and Job ownership changes supersede this source for final acceptance. |

The selective route owner now settles active handlers before unregistering and retains failures through the scope. Ordinary route registrations/removals migrate together; failure finalizers cancel barriers, and successful paths release explicitly. Feature Jobs are being attached through a separate feature fixture, including in-flight admission, lost upstream replies, original dispatch authority, and early context closure. Context initialization remains a generic scope hook. Existing per-feature deadlines are preserved. Acquisition/route/feature diagnostics retain bounded operation categories instead of private dependency messages. Remaining S-001 gates include these new Job witnesses, full consumer validation, primary/cleanup failure preservation and harness quarantine qualification.

Failure diagnosis used public `make explain-run` summary/log views and bounded referenced diagnostics. No primary/cleanup failures were discarded; the two planning failures produced no product evidence. No browser case has been promoted or removed. The next entry gate is still the S-001 exit, not WF-03. `agent-finalize`, generated closure, broader final qualification, owner evidence audits and performance acceptance remain pending; no compatible successful full warm-check `RESULTS_DIR` is available for retained-run maintenance.

Additional S-001 observations (all result roots are `.cartulary/test-results/<run-id>`; all source remains uncommitted on the frozen base):

| Run ID and exact command suffix after `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run-id>` | Native result | Evidence and disposition |
|---|---|---|
| `tr-s001-format-20261024-05`: `make format` | Session `77581`, exit 2 before graph execution. | Configuration failure: new cleanup-suite titles were not ASCII sorted. `make target-plan TARGET=format DETAIL=graph JSON=1` identified the authored catalog constraint; titles were sorted, retaining the row ID and all tests. |
| `tr-s001-format-20261024-06`: `make format` | Session `4852`, exit 0; graph 7959 ms. | Feature-job ownership and its unit/browser witnesses formatted successfully. |
| `tr-s001-types-20261024-09`: `make lint CHECK=types` | Session `16451`, exit 2; graph 22538 ms. | A self-referential context initializer needed an explicit scope type. Cleanup 11/11 passed; annotation repaired. |
| `tr-s001-support-20261024-03`: same two-row cleanup/session command as support-02 | Session `23614`, exit 0; graph 4330 ms. | Expanded cleanup/Job tests passed. |
| `tr-s001-ownership-20261024-03`: same exact four-row command as ownership-02 | Session `93409`, exit 0; graph 87243 ms. | All four browser cases passed, now including real Evidence Job admission, injected failure after admission, terminal settlement before secondary context closure, and subsequent session revocation. Cleanup 23/23 passed. Source `sha256:fe45d8c3ee7c37e1ced3482029007fb98cad48f9b5e7be6b3c9b6660622bfa2e`; graph `sha256:a3687a85471b1317e3114d04cbc587ab8cbff203d479b214cfaaf35ade599df8`. |
| `tr-s001-format-20261024-07`: `make format` | Session `72736`, exit 0; graph 7977 ms. | Socket observer retirement and removal of an orphaned pending-queue race formatted successfully. |
| `tr-s001-types-20261024-10`: `make lint CHECK=types` | Session `84127`, exit 2; graph 20661 ms. | Restored the Playwright Response type import still used elsewhere in row mutation support. Cleanup 11/11 passed. |
| `tr-s001-format-20261024-08`: `make format` | Session `17281`, exit 0; graph 8055 ms. | Worker control-plane finalization and socket observer tests formatted successfully. |
| `tr-s001-types-20261024-11`: `make lint CHECK=types` | Session `43749`, exit 0; graph 23468 ms. | Types passed on source `sha256:0870b326d6edb4921ad8a083ef6675f7d3a8be5dc5b99fed1ba8858617d1f944`. |
| `tr-s001-support-20261024-04`: complete command below | Session `10239`, exit 2; graph 19169 ms. | One failure: architecture policy found the unused exported `TrackedSessionSnapshot` type. Other selected support rows passed, including socket cancellation, fixtures and Job ownership; cleanup 11/11 passed. Source as types-11; graph `sha256:72422ced5e030d63652a17d153cee133c9793ec2c7877c4eebb72d16b2daf41f`. This remains a failed invocation, not full support completion. |

The widened support selection was discovered with `make task-guide OWNER=harness.browser DETAIL=coverage JSON=1` and includes all 24 Vitest rows:

```sh
CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s001-support-20261024-04 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131,harness.browser.boundary_support.atomic_private_harness_state_publication_and_val_6e8cd66ca6,harness.browser.boundary_support.autosave_feedback_diagnostics,harness.browser.boundary_support.cleanup_suite_9036a615b1,harness.browser.boundary_support.evidenceupload_suite_6f27a1c9e4,harness.browser.boundary_support.fixtureidentity_suite_827c446bfe,harness.browser.boundary_support.fixtures_suite_b570a8a829,harness.browser.boundary_support.harnessstate_suite_a4c7c5d935,harness.browser.boundary_support.incidentsocket_suite_6db375c4ec,harness.browser.boundary_support.mentions_suite_e2e13939bd,harness.browser.boundary_support.mutationanchors_suite_ba8fab1244,harness.browser.boundary_support.ownercharacterization_suite_3242ca020b,harness.browser.boundary_support.public_http_diagnostic_privacy,harness.browser.boundary_support.publicjsonclient_suite_a3aedb76f9,harness.browser.boundary_support.record_stable_collaboration_summary_polling_604f761c28,harness.browser.boundary_support.savedviews_suite_79e8036a1a,harness.browser.boundary_support.sessions_suite_4d7bc21669,harness.browser.boundary_support.source_publicjsonclient_test_ts_a6c5c0d17c,harness.browser.boundary_support.source_sessions_test_ts_24e61441d9,harness.browser.boundary_support.testcontrolclient_suite_f60081d8f5,harness.browser.boundary_support.timeline_investigation_recipe,harness.browser.boundary_support.timingsupport_dom_suite_6614d14ea5,harness.browser.boundary_support.visual_preference_isolation,harness.browser.boundary_support.visual_presentation_profiles HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off
```

Further S-001 validation and consumer-audit results follow. Unless explicitly stated otherwise, commands use `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run-id>` from the repository root, and result roots are `.cartulary/test-results/<run-id>`. Source remains uncommitted on the frozen base. Browser selections use `make test-slice OWNER=<owner> ROWS=<rows> HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`. These concurrent correctness observations are not performance comparison pairs.

| Run ID / command selection | Native result | Evidence and disposition |
|---|---|---|
| `tr-s001-ownership-20261024-04`: exact four-row ownership command above | Session `67868`, exit 0; graph 91848 ms. | Four cases and 23 cleanup steps passed. Source `sha256:0870b326d6edb4921ad8a083ef6675f7d3a8be5dc5b99fed1ba8858617d1f944`; ownership graph unchanged. Includes worker control-plane and socket cleanup changes. |
| `tr-s001-runtime-20261024-01`: `make test-slice OWNER=harness.command_surface ROWS=harness.command_surface.unit.native_runtime HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Session `82511`, exit 0; graph 65809 ms. | Native lifecycle, browser failure continuation, failure precedence, cleanup quarantine and retry accounting pass; 10 cleanup steps passed. Source as ownership-04; graph `sha256:20a82a20da09fc95d1bbe7b1d9509e477288990fd3f3b72b983300a48aa39189`. No scheduler implementation changed. |
| `tr-s001-support-20261024-05`: exact 24-row support-04 command with fresh run ID | Session `42509`, exit 0; graph 34276 ms. | All 24 rows and 11 cleanup steps passed after making the unused snapshot type private. Source `sha256:56b076429e1fb67f9314ae7d009bcb23e206fd6451812458cf069b681022d580`; support graph unchanged. |
| `tr-s001-auth-20261024-01`: owner `module.auth`; rows `module.auth.browser_stateful.idle_expiry_fails_closed_after_the_shared_clock_6525f19d9d` | Session `64903`, exit 0; graph 168696 ms. | Actual HTTP expiry and fresh-login case passed; 23 cleanup steps passed. Source as support-05; graph `sha256:d9a020f25a3f8d802fa94bedcc0c7f1fe400f20352ce0edbc845ac287056d26b`. |
| `tr-s001-collaboration-20261024-01`: owner `module.collaboration`; rows `module.collaboration.browser.within_one_browser_runtime_queued_unsent_writes_3efbd48054,module.collaboration.browser.concurrent_edits_to_different_fields_on_the_same_7ee5cad0ab` | Session `44074`, exit 0; graph 168759 ms. | Both cases and 26 cleanup steps passed. Source as support-05; graph `sha256:ffde0dd24e3d2e9786607c36105d913724d41b5cef974cfa08e57c26a6247099`. |
| `tr-s001-reference-data-20261024-01`: owner `module.reference_data`; rows `module.reference_data.browser.administration_async_recovery,module.reference_data.browser.administration_real_lifecycle` | Session `34409`, exit 0; graph 235525 ms. | Both cases and 26 cleanup steps passed. Source as support-05; graph `sha256:970ed9dadefa2973dda51ea7aab120eb9da149e3010ed462f962eb198f9119a2`. |
| `tr-s001-imports-20261024-01`: owner `web.design`; rows `web.design.browser.incident_import_observation_recovery,web.design.browser.incident_import_real_workflow` | Session `92154`, exit 0; graph 205712 ms. | Both cases and 26 cleanup steps passed. Source as support-05; graph `sha256:1e7e530029152cc76fa06ce6fa8dcd9e0e9a0956b5e4c130838eb27afbb58201`. |
| `tr-s001-evidence-20261024-01`: owner `module.evidence`; rows `module.evidence.browser.file_stage_recovery,module.evidence.browser.an_inline_safe_type_receives_a_preview_handle_an_076fefa625,module.evidence.browser.verification_cancellation_race` | Session `5213`, exit 0; graph 215175 ms. | All three cases and 23 cleanup steps passed. Source as support-05; graph `sha256:b9339176b5d5fd5e83b79a5369b6f532b14afbe659619bb55cfd09738f3e2aa0`. |
| `make lint-markdown` (no explicit run ID) | Session `62185`, exit 0; duration 14260 ms. | Documentation passed; exact root `.cartulary/test-results/20261024T183947Z-p30593`, summary `adhoc/lint-markdown/tool-run-summary.json`. |
| `tr-s001-format-20261024-09`: `make format` | Session `10943`, exit 0; graph 7986 ms. | First finalizer audit formatted successfully. |
| `tr-s001-types-20261024-12`: `make lint CHECK=types` | Session `18086`, exit 0; graph 22171 ms. | Types passed after the first finalizer audit. |
| `tr-s001-support-20261024-06`: exact 24-row support command | Session `98789`, exit 0; graph 17979 ms. | All support rows passed after the first finalizer audit. |
| `tr-s001-workbook-20261024-01`: owner `module.workbook`; rows `module.workbook.browser.requested_filters_stay_editable_through_held_and_598bebf8f3` | Session `53943`, exit 0; graph 85825 ms. | Both selected cases passed. Source `sha256:b2b90c2bc6d918eaed841f4004b5df5e8f5f38f7cfd63290e515924d206f4939`; graph `sha256:e8a19e2d6ec2974887f7e9661b4e661ba858e2e74f11084e85407b5a3fd3b455`. |
| `tr-s001-format-20261024-10`: `make format` | Session `49095`; native exit unavailable after session boundary. | Public summary reports pass, 7950 ms, 11 successful cleanup steps. This is not confirmed native success. Source `sha256:26eec0fa60d533a7f32d2eaeab670ef162feae7f457a6d5a1df25d4aece9631e`; graph `sha256:7c96f2e1859b1cb545a8a6fc15d378fdb6a57a16808265c4d6f3aeafe8031a6f`. |
| `tr-s001-types-20261024-13`: `make lint CHECK=types` | Session `92410`, exit 0; graph 21846 ms. | Types passed after the final consumer cleanup audit. |
| `tr-s001-support-20261024-07`: exact 24-row support command | Session `19914`, exit 0; graph 17216 ms. | All support rows passed after the final consumer cleanup audit. |

The final consumer audit removed competing route/context/client retirement from failure-path finalizers. Successful paths still retire resources when the scenario needs an early boundary; on failure the shared scope cancels and reports cleanup independently of the primary assertion. The last held Workbook response now handles cancellation and exposes fulfillment failure. Feature Jobs own admission before dispatch and settle before their context request transport closes; socket observers detach their listeners and cancel pending waits. Ordinary session/helper consumers use the common scope; restore/publication-specific runtimes remain reserved for S-005. No baseline browser identity, product obligation or isolation declaration has been removed or promoted. The five final-source browser selections are running before the S-001 completion checkpoint; WF-03 remains blocked until those results are reconciled.

### E-028 — WF-02 completion checkpoint

**Status: COMPLETE.** S-001 closes on uncommitted source based on `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`, executable source digest `sha256:4d1d83341c8057a6d99e946fa9950bbcb41ef6279324f6dbc2b415c0dd428100`. E-027 records the implementation, intermediate failures, repairs and commands. One common scope now owns all ordinary browser contexts, independent clients, actors, acquired sessions, response barriers and route failures. Feature fixtures own Job admission/terminal observation and socket/background callbacks; the common scope orders their settlement before transport retirement. Late acquisition, repeated failed disposal and cleanup errors remain observable without credential-bearing diagnostics. Specialized restore/publication resources retain their separate S-005 entry gate.

The final consumer audit and type/support runs cover the expanded helper boundary, including callers outside the frozen inventory. Browser selections below qualify the changed lifetime against real HTTP session revocation, live provisioner survival, cancelled writes, secondary closure, terminal Jobs, collaboration replay, signed Reference Pack operations, import workflows, Evidence preview/mutation/cancellation, and Workbook held-response recovery. The unchanged native runtime selection in E-027 qualifies failure precedence, cleanup quarantine, early-stop and zero-retry behavior. This is S-001-specific V-002/V-004/V-007 closure, not final acceptance of V-001–009.

All commands retain the exact selections/settings recorded in E-027, changing only `CARTULARY_TEST_RUN_ID` as shown. All result roots are `.cartulary/test-results/<run-id>`. All six browser runs below use the same source digest above and the graph digests recorded for their matching earlier selection. Native exits were collected independently of public summaries; no primary, cleanup or retained-scan failure occurred.

| Final run ID | Native result | Coverage and cleanup |
|---|---|---|
| `tr-s001-collaboration-20261024-02` | Session `96334`, exit 0; 201915 ms | Both selected cases passed; 26/26 cleanup steps passed. |
| `tr-s001-reference-data-20261024-02` | Session `12268`, exit 0; 201871 ms | Both selected cases passed; 26/26 cleanup steps passed. |
| `tr-s001-imports-20261024-02` | Session `20439`, exit 0; 225951 ms | Both selected cases passed; 26/26 cleanup steps passed. |
| `tr-s001-evidence-20261024-02` | Session `67909`, exit 0; 241228 ms | All three selected cases passed; 23/23 cleanup steps passed. |
| `tr-s001-ownership-20261024-05` | Session `96851`, exit 0; 201827 ms | All four fixture cases passed, including real revocation and Job settlement witnesses; 23/23 cleanup steps passed. |
| `tr-s001-workbook-20261024-02` | Session `77628`, exit 0; 91458 ms | Both selected cases passed; 23/23 cleanup steps passed. |
| `tr-s001-types-20261024-13` | Session `92410`, exit 0; 21846 ms | Complete TypeScript check passed. |
| `tr-s001-support-20261024-07` | Session `19914`, exit 0; 17216 ms | All 24 selected support rows passed; 11/11 cleanup steps passed. |
| `tr-s001-docs-20261024-01`: `CARTULARY_TEST_RUN_ID=tr-s001-docs-20261024-01 make lint-markdown` | Session `49102`, exit 0; 11330 ms | Tracker/specification lint passed before this completion record; final Markdown validation remains required in S-006. |

**Coverage dispositions:** all 88 frozen rows / 97 cases and their production obligations remain accounted for. No browser case was removed, consolidated or promoted; canonical surviving identities remain unchanged. The authored cleanup-support row gained the actual ownership witnesses with its existing identity. Full-suite, measurement, accessibility, visual and final-owner audits remain S-006 gates; these narrow selections do not claim those results. No scheduler, production schema, production deadline, retry policy or server fault-control interface changed.

**Remaining risks:** scenario-global mutations still require S-002; public-error machinery still requires S-003; specialized runtime/process dependencies require S-005; sharing needs S-004 interference/accounting qualification; generated closure, broad final validation, owner audits and four-pair performance acceptance require S-006. The preexisting borrowed bootstrap/worker control plane remains outside test subject ownership and under worker/global harness cleanup. Failed/unsupported resource retirement quarantines its allocation; it is not silently accepted. `agent-finalize` is pending the implementation end-of-run boundary, and retained-run maintenance has not run because a compatible successful full warm-check `RESULTS_DIR` is unavailable.

**Next entry gate:** WF-03/S-002 is READY. Inspect and establish each real backend successor before removing the corresponding browser enforcement assertions, then record the replacement mapping and actual validation results here before WF-06. The complete cutover remains one merge/rollback unit.

### E-029 — S-002 implementation progress (not a completion checkpoint)

WF-03/S-002 began only after E-028 closed WF-02. Source base remains `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`. The first changes strengthen real HTTP successors before removing any browser enforcement assertion. `TestUserAdminLifecycle_Integration` now checks last-admin demotion and deactivation rejection, exact unchanged durable user state, two surviving target sessions and an unrelated usable session. `TestNetworkFlowPaginationRecovery_Integration` retains exact 899/900-second expiry boundaries for rows, diagnostics and contributors, and adds real actor/session binding, cross-route token rejection and forbidden continuation overrides on all three routes. Core 01 §3.3.5.1, Core 04 §1.1.1, and Network Flow §13 remain authoritative; no production semantics or TTL changes are planned. Existing unit/store witnesses remain retained.

No browser assertion has yet been removed and no isolation declaration has changed. Backend successor validation is pending; dependent workstreams remain BLOCKED. Runtime results and the detailed replacement accounting will be appended before the WF-03 exit checkpoint.

S-002 backend characterization (all result roots `.cartulary/test-results/<run-id>`; commands from the repository root with `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run-id>`):

| Run ID and command | Native result | Evidence / disposition |
|---|---|---|
| `tr-s002-format-20261024-01`: `make format` | Session `88432`, exit 0; 8135 ms | Backend successor additions formatted. |
| `tr-s002-auth-backend-20261024-01`: `make test-slice OWNER=module.auth ROWS=module.auth.integration.user_create_applies_defaults_stores_argon2id_pas_8af88020c6 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Session `49646`, exit 0; 27202 ms | Real last-admin demotion/deactivation, exact durable-state preservation and target/unrelated session witnesses passed. Cleanup 14/14 passed. Source `sha256:2f06691bb20a62ce372574dd5589da7cf50af162e81ad656f86e80b88224d57f`; graph `sha256:767a097c2706264264921be46a91f28fc696e1994ef736727e3a509a3320e117`. |
| `tr-s002-network-backend-20261024-01`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.pagination_recovery HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Session `26655`, exit 2; 25779 ms | All three routes rejected semantic overrides with generic `network_flow_invalid_request`, contradicting Network Flow §13.1's required `network_flow_cursor_invalid`. Actor/session/cross-route checks passed before the assertion. Cleanup 14/14 passed. Source as auth-backend-01; graph `sha256:e6d6f49b99be41baa75860a4d26803c1346fa6d1e30d5583464dcd6b7e622344`. |
| `tr-s002-format-20261024-02`: `make format` | Session `15027`, exit 0; 7983 ms | Shared request-variant decoder and unit matrix formatted. |
| `tr-s002-network-backend-20261024-02`: `make test-slice OWNER=module.networkflow ROWS=module.networkflow.integration.pagination_recovery,module.networkflow.unit.network_flow_selector_covers_every_network_flow_43bd8f4992,module.networkflow.unit.query_authoring_admission HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | Session `93058`, exit 2; 24574 ms | Decoder/unit checks and rows/diagnostics passed. The alternate-session witness expired before the contributor case because rejected requests correctly do not renew idle time. Added a successful authorized initial-query control before each binding denial; no TTL change. Cleanup 15/15 passed. Source `sha256:f31033705174c5eb26254584136a8f9e74575ee9870d80cdcf21e2e359b3a5dc`; graph `sha256:d5a403e00b8f08d3011bdb9711535fb6c28bfd3d8aa6d11a84b2803828f2e64c`. |

The discovered decoder gap is repaired in the Network Flow owner module through one closed-variant admission function shared by accepted rows, rejected diagnostics and contributors. It rejects initial-plus-cursor and continuation-plus-semantics with `network_flow_cursor_invalid`, and malformed continuation tokens with the same owner error family. HTTP status remains 400; the error-code correction is an intentional conformance repair, with no schema/data migration or compatibility alias. Existing initial-request errors and token cryptography remain unchanged. All failures remain retained; replacement browser coverage still awaits implementation and validation.

Further S-002 progress: `tr-s002-network-backend-20261024-03` repeated the exact three-row network-backend-02 command after adding successful authorized controls. Native session `29795` exited 0, graph 17423 ms, cleanup 15/15 passed; source `sha256:baa99c97fc08ad64cd6749d39e4399bdc6b1dc7a7f43bab969e47d4e6b288ee5`, graph unchanged from network-backend-02. This closed the backend successor gate before browser enforcement assertions were replaced.

The browser edits now replace B-011's deployment-wide demotion with two exact PATCH presentation responses, while retaining owned CRUD, stale versions and navigation. The obsolete demotion helper, its facade and its two private-helper tests are removed; surviving generated-operation tests exercise active lookup/load/patch helpers. B-017 splits into three independently owned cases; the original row/scenario identity carries actual idle expiry, with new `session_recovery_revoke_all` and `session_recovery_logout` rows carrying their unchanged replay helper assertions. B-048 retains its canonical matrix row with local exact expiry responses and real pages; new `module.networkflow.browser_stateful.pagination_expiry_bridge` supplies actual rows-route expiry. B-065 retains its canonical real account-replacement/paging case with a local exact query failure. B-051 retains real signed bytes/replay and lifecycle enforcement; file-selection navigation, keyboard retry, observation recovery and detail disclosure move into existing B-050 presentation coverage. B-012's actual HTTP expiry case remains unchanged. The frozen baseline remains 88 rows / 97 cases; these splits add three successor cases, without retiring baseline row identities. No case is promoted to shared execution yet; S-004 remains the admission gate.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s002-format-20261024-03 make format` exited 0 (session `83359`, 8058 ms). `tr-s002-support-20261024-01` used E-027's exact 24-row support command with the new run ID and exited 0 (session `84063`, 18937 ms), including the exact method/path/error hit-count witness. Result roots follow the run IDs. Browser successor validation is pending; no completed S-002 checkpoint is claimed.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s002-types-20261024-01 make lint CHECK=types` exited 0 (session `30791`, 23050 ms). Five browser selections then launched with fresh `tr-s002-<name>-browser-20261024-01` IDs, the common basic/required accounting/cache-off settings, and these exact owner/row arguments:

| Name / native session | Owner | Rows | Initial outcome |
|---|---|---|---|
| auth / `94768` | `module.auth` | `module.auth.browser_stateful.browser_deployment_user_administration_works_on_d6d7ff0b86` | Running |
| collaboration / `63927` | `module.collaboration` | `module.collaboration.browser_stateful.deterministic_route_owned_browser_revocation_sou_a134c3e26b,module.collaboration.browser_stateful.session_recovery_logout,module.collaboration.browser_stateful.session_recovery_revoke_all` | Exit 2 before runtime execution: generated groups missing for the two new rows. |
| network / `73932` | `module.networkflow` | `module.networkflow.browser_stateful.pagination_expiry_bridge,module.networkflow.browser_stateful.pagination_recovery` | Exit 2 before runtime execution: generated group missing for the new bridge row. |
| workbook / `36034` | `module.workbook` | `module.workbook.browser_stateful.query_continuation_session` | Running |
| reference-data / `14450` | `module.reference_data` | `module.reference_data.browser.administration_async_recovery,module.reference_data.browser.administration_real_lifecycle` | Running |

Public diagnosis used `make target-plan TARGET=test-slice INPUTS=/tmp/cartulary-s002-collaboration-plan.json DETAIL=graph JSON=1` and the corresponding `network` file. Inputs contain exactly the table's OWNER/ROWS and the common diagnostics/accounting/cache settings. Both report missing generated browser groups; neither establishes product failure or a complete resource claim. `make generate` is required after current source-bound runs drain, before retrying those selections. An attempted `make help TARGET=target-plan` exited 2 because `help` does not accept TARGET; it produced no verification claim. These failures and superseded evidence remain retained.

Browser-01 terminal reconciliation: auth passed (session `94768`, exit 0, 136176 ms, cleanup 23/23), source `sha256:eb31769e8f5944d43eaf532eed4c0725e28a2f5e3b1d9a5535209c0d5cd78fcd`, graph `sha256:7184124367a960f599e4acf0dbbea47bc8b07d303ce22d548d2ae44ea5635e9e`. Workbook failed (session `36034`, exit 2, 140104 ms, cleanup 23/23): the local query 401 alone correctly triggered an authoritative session read which still returned the real valid session, so the expected login shell did not appear. The presentation fixture now also supplies one exact `GET /api/v1/auth/session` 401 before genuine replacement login; both hits are asserted. This preserves the application's authoritative-session revalidation behavior.

Reference-data browser-01 failed (session `14450`, exit 2, 156347 ms, cleanup 26/26). The moved presentation case passed. In the real lifecycle case, replay-capture equality ran before the second request arrived after its prior observation retry was removed. Equality and exact admission count now follow the already-required committed-state wait; no deadline or retry policy changes. Both failures used the same source as auth-browser-01. Canonical `explain-run` children/log views led to the affected `browser-group-result.json` and its declared `playwright-report.json`; bounded structured errors identified Workbook line 2906 and Reference Pack line 85. Attempts to use unsupported `DETAIL=failures` exited 2 and establish no evidence. No cleanup failure was hidden and neither failed invocation counts as successor completion.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s002-generate-20261024-01 make generate` passed (session `80226`, native exit 0, 11750 ms; summary `generate/tool-run-summary.json`). It regenerated the browser batch manifest and execution-topology render index from authored inputs. `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s002-format-20261024-04 make format` passed (session `71740`, native exit 0, 8010 ms). After those repairs and generation, the exact collaboration/network/workbook/reference-data browser selections above relaunched with run suffix `-02` and native sessions `70091`, `88632`, `8653`, and `54829`, respectively. Type checking (`tr-s002-types-20261024-02`, session `90519`) and `make source-check CHECK=catalog` (`tr-s002-catalog-20261024-01`, session `74420`) are also running. Commands retain `CGO_ENABLED=0` and their explicit run IDs. No new source edits occur while those source-bound checks execute.

### E-030 — WF-03 / S-002 completion checkpoint

**Status: COMPLETE.** The backend successor gate passed before replacement of browser enforcement assertions (E-029). The final browser successors all passed with native exit 0. Source base remains `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`; final browser-02 source is `sha256:64cab350e114480badd07d4e3da7df3c459c056785ed825e754902d6ed40df13`. Toolchain digest is `sha256:54f13740b73fafe0723d9ab2b9c8e06e21e9df6ed39e2b4bc3ba3f744116f88f`; system digest is `sha256:008fe637b622cd8932b55cc1d8be5346bae02c09ab35d7518fc9f72b66eb6c21`.

Commands use exactly E-029's owner/row selections, `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run-id> make test-slice OWNER=<owner> ROWS=<rows> HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`. Roots are `.cartulary/test-results/<run-id>`; published run summaries and cleanup artifacts retain complete native evidence.

| Run ID | Native session / exit | Graph elapsed | Cleanup | Graph digest |
|---|---|---|---|---|
| `tr-s002-collaboration-browser-20261024-02` | `70091` / 0 | 216237 ms | 31/31 passed | `sha256:e58b9311776990ed9521605c6cb2000e4efbda1f1073ecd8f8fa9171cb4ab48c` |
| `tr-s002-network-browser-20261024-02` | `88632` / 0 | 208026 ms | 27/27 passed | `sha256:d8e4e2953467de4a36c04db542a00b3b679dd3c78eac762d04f4a51bd221f496` |
| `tr-s002-workbook-browser-20261024-02` | `8653` / 0 | 186585 ms | 23/23 passed | `sha256:23a908edfc295236d840201959407f16f2f9b4ba3ce5ea431f35626c30e31fc3` |
| `tr-s002-reference-data-browser-20261024-02` | `54829` / 0 | 186621 ms | 26/26 passed | `sha256:970ed9dadefa2973dda51ea7aab120eb9da149e3010ed462f962eb198f9119a2` |

Additional native passes (same environment/run-root convention): `tr-s002-types-20261024-02 make lint CHECK=types` (session `90519`, exit 0, 27769 ms); `tr-s002-catalog-20261024-01 make source-check CHECK=catalog` (`74420`, exit 0, 2076 ms); `tr-s002-generated-20261024-01 make source-check CHECK=generated` (`69961`, exit 0, 19659 ms). E-029 retains the passing auth browser, backend, support, format and generation evidence and every failed observation. No failing primary or cleanup result is represented as a pass.

**Coverage dispositions:** B-011 retains real CRUD/version/navigation and two local guard presentations, backed by real HTTP demotion/deactivation rejection, unchanged durable state and surviving target/unrelated sessions. B-012's real HTTP expiry remains retained. B-017's original identity carries actual socket idle expiry; two new independently owned cases carry revoke-all/logout and the same pending-work replay assertions. B-048's canonical matrix keeps real paging and local recovery across rows/diagnostics/contributors; one new real rows-expiry bridge complements backend 899/900-second and scope enforcement. B-065 keeps real account replacement/paging and verifies old protected state/cursors are discarded; both query denial and authoritative session denial are exact local responses. B-051 keeps real signed import/replay and all lifecycle operations; presentation/keyboard/navigation/retry assertions now pass in B-050. No local response is credited as backend enforcement. All 97 baseline cases retain their obligation mappings; splits add three successors (91 rows / 100 cases before S-004), with no baseline row retirement or sharing promotion.

The Network Flow successor exposed an existing conformance gap: forbidden continuation overrides returned generic `invalid_request`. Shared query-variant decoding now returns the owner-required `cursor_invalid` family for semantic overrides or malformed tokens on the four query variants. The real integration and contract tests pass. This deliberate error-code correction changes no cursor lifetime, authorization, cryptography, schema or persistence. No production migration is required. The browser-wide administrator demotion helper and its two private-helper-only tests are retired; active lookup/load/patch coverage remains.

**Validation boundary:** the S-002 portions of V-001/V-004–006 are closed by inspected assertion mapping and these results. Final compatible-source acceptance of V-001–009 is still required by S-006. No deadlines, retries or golden expectations were increased. Remaining risks are S-003 fault-control retirement, S-005 specialized resource lifetimes, S-004 interference/accounting qualification, and final correctness/performance qualification. `agent-finalize` has not yet run; no compatible successful full warm-check `RESULTS_DIR` is available for retained-run maintenance.

**Next entry gate:** WF-06/S-003 is READY. E-026 already records human approval of the coordinated owner/interface retirement mapping. Replace the three presentation consumers, strengthen the real serializer witness, remove all runtime/schema/client dependencies, regenerate and verify retirement before recording the next checkpoint. WF-07 remains blocked on that checkpoint. Merge and rollback remain the complete cutover.

### E-031 — S-003 implementation and validation in progress

WF-06 began after the E-030 checkpoint. All three browser consumers now use the common exact-method/path local response helper. Timeline retains the unknown-code/private-diagnostic presentation assertion and all real mutation/replay coverage; Evidence retains real upload, preview, download, focus/accessibility and visual assertions. The runtime-global registry, six implementation-only tests, response schema and attachment, generic HTTP dependency/middleware/types, assembly injection, and browser control clients are removed. The now-stateless generic `Controls` object is also removed; route composition retains clock, runtime identity and Network Flow contributions.

The old harness row is retired through `tools/test_catalog_row_migrations.json`, with successors `harness.browser.unit.retired_public_error_control` and `app.server.unit.internal_error_boundary`. The negative witness exercises controls disabled/enabled and missing/valid authorization, with a successful enabled identity-route positive control. Serializer witnesses use a real HTTP server and the production handler/request-ID middleware; they compare complete status/code/message/details/retryable/correlation envelopes, exercise supplied/generated correlation, and retain private internal-diagnostic suppression. Core 04 AC-413's obsolete route-list reference is reconciled with the already-approved Testing Harness retirement; its general security requirements are unchanged.

Initial `tr-s003-format-20261024-01` failed configuration before execution (session `5377`, native exit 2): the new migration replacement IDs were not ASCII-sorted. Three attempted `make task-guide OWNER=<harness.browser|module.evidence|module.workbook> DETAIL=coverage JSON=1` calls reported that same authored-input error and exited 2. Sorting the migration fixed the prerequisite. `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s003-generate-20261024-01 make generate` then passed (session `75131`, exit 0, 11560 ms; `generate/tool-run-summary.json`); `tr-s003-format-20261024-02 make format` passed (`40402`, exit 0, 8062 ms). Subsequent public task-guide discovery passed for all affected owners.

All run roots use `.cartulary/test-results/<run-id>`. Commands below use `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run-id> make test-slice OWNER=<owner> ROWS=<rows> HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`.

| Run ID | Owner / exact rows | Native session | Current result |
|---|---|---|---|
| `tr-s003-serializer-20261024-01` | `app.server` / `app.server.unit.internal_error_boundary` | `80764` | Exit 0; 3595 ms |
| `tr-s003-retirement-20261024-01` | `harness.browser` / `harness.browser.unit.retired_public_error_control,harness.browser.boundary_support.cleanup_suite_9036a615b1` | `29707` | Exit 0; 14348 ms |
| `tr-s003-evidence-browser-20261024-01` | `module.evidence` / `module.evidence.accessibility.verify_evidence_icon_buttons_blocked_states_erro_0ec80f90e2,module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | `54914` | Running |
| `tr-s003-timeline-browser-20261024-01` | `module.workbook` / `module.workbook.browser_stateful.verify_timeline_public_mutations_remount_safe_tr_84f645d6c5` | `72050` | Running |

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s003-types-20261024-01 make lint CHECK=types` is running (`52386`). The five `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s003-<check>-20261024-01 make source-check CHECK=<check>` invocations have: generated running (`41054`); schema pass (`75853`, exit 0, 7627 ms); catalog pass (`66462`, exit 0, 2453 ms); identity pass (`41140`, exit 0, 1145 ms); surface pass (`95347`, exit 0, 3132 ms).

An authored/generated source search now finds the retired control only in its negative witness and historical row-migration record. Historical handoff inventories remain attributable to their frozen source; active owner inventory is reconciled. No browser row has been retired or promoted. Final browser outcomes, cleanup, source compatibility and remaining retirement closure must be reconciled before WF-06 completion; WF-07 remains BLOCKED.

### E-032 — WF-06 / S-003 completion checkpoint

**Status: COMPLETE.** All E-031 successor selections passed with native exit 0. Source base is `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`, dirty candidate digest `sha256:ed2ebf0a192b104663bb30354ffeab2408344a61f77b6a7b24ce6fcb8f95b90a`; toolchain/system digests match E-030. The exact commands, selections, run roots and native sessions are recorded in E-031.

| Run ID | Native exit / elapsed | Cleanup | Graph digest |
|---|---|---|---|
| `tr-s003-serializer-20261024-01` | 0 / 3595 ms | 10/10 passed | `sha256:2919096d681eb12afbe7156e9c1af4eb37d2cc80e3971191fbf88a732935335e` |
| `tr-s003-retirement-20261024-01` | 0 / 14348 ms | 12/12 passed | `sha256:66270fe0b0d0cb3784e184e1509a8a07cb0ba0186910b772c6efa023275cf178` |
| `tr-s003-evidence-browser-20261024-01` | 0 / 127712 ms | 27/27 passed | `sha256:94415949609ecccbffdd787ca992f7f35752fa6e11a6474cebed5a433b89def4` |
| `tr-s003-timeline-browser-20261024-01` | 0 / 131584 ms | 23/23 passed | `sha256:f79805ee5dfb2ee2d53bcd7a30f794e24118e803665348a6a4bcc085be11d3bf` |

The remaining E-031 checks also passed: `tr-s003-generated-20261024-01` (`41054`, native exit 0, 20641 ms) and `tr-s003-types-20261024-01` (`52386`, native exit 0, 30531 ms). `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s003-markdown-20261024-01 make lint-markdown` passed (`10616`, native exit 0, 10900 ms; `adhoc/lint-markdown/tool-run-summary.json`). `git diff --check` passed. The initial configuration failure remains retained; `explain-run` confirms its root has only prerequisite tool summaries and no run identity, so it establishes no runtime or resource result.

**Coverage and retirement accounting:** B-019/B-021 retain their canonical cases and all real Evidence behavior/accessibility/visual assertions; their error presentation is local. Both B-066 cases pass, preserving real Timeline persistence, replay and recovery; only the unknown-error presentation response changed. No browser baseline identity or obligation was retired. The six obsolete registry/control-only tests and their single harness row were deliberately replaced with the negative control witness and strengthened real serializer row, using authored migration accounting. No alias or replacement server API exists. Source searches across `apps`, `internal`, `tools` and `contracts` find the retired name only in the negative witness and historical migration record; generated validators and attachment indexes no longer expose its schema. Clock and Network Flow controls remain unchanged. Approved Testing Harness REQ-481/AC-035 and Core 04 AC-413 inventories agree.

**Validation boundary and risks:** S-003's V-001/V-005/V-006/V-009 obligations are closed on compatible source. Final-source V-001–009, retained evidence audits and performance acceptance remain S-006 obligations. There were no product or cleanup failures in the successor runs, no golden updates, retries, deadline increases or production schema migrations. Removal is a deliberate test-only API break, merged and rolled back only with the complete cutover. Specialized lifetimes and sharing remain unqualified; no sharing admission is implied. `agent-finalize` and retained-run maintenance remain pending as recorded in E-030.

**Next entry gate:** WF-07/S-005 is READY. Integrate restore/publication browser ownership with their existing disposable runtimes, preserve process/build/loaded-asset evidence, and validate startup/body/interruption/cleanup failure paths before the next checkpoint. WF-04 remains BLOCKED until S-005 is complete.

### E-033 — S-005 specialized lifetime implementation in progress

WF-07 began after E-032. Restore's fixture now disposes the complete browser scope before stopping the disposable source/target process, replacing `about:blank` navigation. It still attempts process retirement after browser cleanup failure and preserves its existing redacted lifecycle attachment and primary/startup/interruption classifications. Restored credentials remain inside that disposable runtime boundary; they are never registered against the ordinary provisioner. Source/target databases, stores, runtime roots, process ownership and fixture deadlines remain unchanged.

Publication build ownership now resides in a test fixture using `OwnedFrontendPublications`. It records every acquired independent build, closes browser dependencies first, attempts every build's retirement despite cleanup failures, rejects new starts after disposal, and shares failed disposal completion. Browser fixture teardown preserves the primary test failure independently of cleanup diagnostics. The production process helper now preserves both a failed build and failed cleanup instead of overriding the former in `finally`. The restore process helper observes early readiness rejection while retaining the original rejection for its dependent fixture. The two remaining direct routes in visual harness support now use common owned interception. Immutable asset digests, repeated reloads and independent ordinary/measurement build evidence remain intact.

The common cleanup row adds partial-startup ownership, browser-before-build settlement, concurrent/repeated failed disposal and private-diagnostic suppression witnesses. Existing adapter regressions retain real startup/forced retirement/cleanup failure/body/interruption evidence; the early-readiness rejection test now delays its dependent await. Frontend process regression adds unsuccessful build and absent executable cases, retaining cancellation and repeated retirement coverage.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s005-format-20261024-01 make format` failed (`29926`, native exit 2, 8120 ms, cleanup 11/11): the new owner's optional Promise was used directly as a condition, rejected by `noMisusedPromises`. It now checks `!== undefined`. The exact format command with suffix `-02` passed (`74140`, exit 0, 7917 ms). Public `explain-run` summary/logs established the diagnostic; an initial logs request without TARGET exited 2 and supplied no verification evidence. A source-edit script also used the wrong selector key (`test_names`); it changed no catalog data before raising, and the authored `titles` list was then updated correctly.

All roots are `.cartulary/test-results/<run-id>`. Public task-guide discovery covered `harness.browser`, `harness.test_catalog` and `module.recovery`. The following commands use `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<run-id> make test-slice OWNER=<owner> ROWS=<rows> HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`:

| Run ID | Owner / exact rows | Native session | Current result |
|---|---|---|---|
| `tr-s005-support-20261024-01` | `harness.browser` / `harness.browser.boundary_support.cleanup_suite_9036a615b1,harness.browser.boundary_support.playwright_adapter,harness.browser.unit.recovery_fixture_server_lifecycle` | `87507` | Exit 0; 9346 ms |
| `tr-s005-publication-browser-20261024-01` | `harness.browser` / `harness.browser.boundary_support.frontend_artifact_lifetime,harness.browser.boundary_support.timeline_investigation_capture,harness.browser.boundary_support.visual_asset_failure` | `58603` | Running |
| `tr-s005-restore-browser-20261024-01` | `module.recovery` / `module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0` | `10342` | Running |

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s005-types-20261024-01 make lint CHECK=types` passed (`63712`, exit 0, 24766 ms). `make target-plan TARGET=harness-contract INPUTS=/tmp/cartulary-s005-regression-plan.json DETAIL=graph JSON=1` identifies the existing public regression suite as the owner of the changed frontend process test; inputs select SUITE=regression and basic/required/cache-off settings. No public shuffle, new scheduler or private runner is introduced. Runtime/browser outcomes, generated closure and regression evidence are pending; WF-04 remains BLOCKED.

S-005 browser terminal results: publication (`58603`) exited 0 in 245310 ms with cleanup 29/29; restore (`10342`) exited 0 in 234076 ms with cleanup 23/23. Support cleanup was 13/13. All three use source `sha256:675cfcdc302e668ad764d5361167cfb755e52a3b91b1053b985f6e5ae8180efd` on the frozen base, with E-030's toolchain/system digests. Publication graph is `sha256:22cbf3a6eb5e88eb9a4cb5cd22a0588040d72acdd8d65bfab5c470b0eddb58dc`; restore graph `sha256:9d1f7107e19ca297d9d2464ea2d5398060bc80bb36807560be9384e7c5d11ebc`; support graph `sha256:3816bfb8023ec7136c6a419fa7910f7616f16a5142a0c13758751eec85e91cb7`.

Published target summaries reference each group result and its Playwright report. The restore lifecycle attachment records exit 0, closed=true, forced=false, no failures, successful source/target database and store cleanup, and a successful terminal fixture event. The publication attachment records 24 reloads preserving all nine loaded asset digests, with two successful independent builds under `frontend-publications/frontend-publication-18712e0c-f33d-4df9-aead-444e54a37bb2` and `frontend-publications/frontend-publication-caad355d-0ad7-40cf-910f-6e88223eccee` within its run root. These are correctness/lifecycle observations, not paired performance acceptance.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s005-regression-20261024-01 make harness-contract SUITE=regression HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` is running in native session `88160`. No source mutations occur during this source-bound regression run. S-005 remains IN PROGRESS until regression and generated closure are reconciled.

S-005 regression-01 completed with native exit 2 (`88160`, 274036 ms), 45/46 units passed and cleanup 55/55. `harness.assurance_json_shapes` rejected stale generated topology inputs for `fixture-process.mjs`, `frontend-publication.mjs` and the authored harness family. Changed process and lifecycle regressions passed, but this invocation is not an overall pass. Source/graph digests are E-033's source and `sha256:180b0ecd03b016fd92fa2d07b91124918f9c86170228513be5583b3874921d77`. Native output also reports incomplete invocation accounting, so no complete resource claim is made. A performance view could not open the absent `invocation-accounting.json` (exit 2); that limitation remains recorded. An attempted logs view using the internal target name was rejected; the public `harness-contract` logs view identified the exact stale-input diagnostic.

After every source-bound run drained, publication settlement was hardened to capture synchronous stop exceptions while still attempting all independent stops; the existing redaction/failure test now exercises that path. `tr-s005-format-20261024-03 make format` passed (`65578`, exit 0, 7959 ms). `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s005-generate-20261024-01 make generate` passed (`36342`, exit 0, 11210 ms; `generate/tool-run-summary.json`). The generated render index now matches authored inputs.

Fresh validation uses `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=<id>` and the following exact commands/roots under `.cartulary/test-results/<id>`:

| Run ID | Make command | Native session / current result |
|---|---|---|
| `tr-s005-regression-20261024-02` | `make harness-contract SUITE=regression HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | `32035`, running |
| `tr-s005-support-20261024-02` | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.cleanup_suite_9036a615b1 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | `19410`, running |
| `tr-s005-publication-browser-20261024-02` | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.frontend_artifact_lifetime HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` | `36102`, running |
| `tr-s005-types-20261024-02` | `make lint CHECK=types` | `25605`, exit 0, 26111 ms |
| `tr-s005-generated-20261024-01` | `make source-check CHECK=generated` | `65028`, exit 0, 21786 ms |
| `tr-s005-catalog-20261024-01` | `make source-check CHECK=catalog` | `12633`, exit 0, 2072 ms |

Restore/process/visual-auxiliary implementations are unchanged from their passing S-005 evidence; the final browser rerun is restricted to the publication owner affected by synchronous-stop hardening. No next-workstream entry is claimed yet.

S-005 rerun terminal reconciliation: support-02 passed (`19410`, native exit 0, 62801 ms, cleanup 11/11) and publication-browser-02 passed (`36102`, native exit 0, 174001 ms, cleanup 23/23). Both use source `sha256:fa5c7b35c822edd094139beec46e6e7ef4fdc2c40dae36cf48e144b3a2a6844f`; support graph `sha256:715d7b30cf38318f9100628e54a7ec5bb6c5f00cf2752f7400ce1c67fae73ea8`; publication graph `sha256:7a5bdc1f54b20f1201703a0242665d8418a7ddfae25456dbece430dfe02e50a8`. Regression-02 reported all 46/46 units passed and cleanup 55/55 in 257137 ms, but native session `32035` exited 2 because outer invocation accounting remained incomplete. It is explicitly NOT an overall successful invocation. Its source is the same and graph is unchanged from regression-01.

The accounting controller's existing diagnostic did not identify whether scope settlement, controller identity or receipt/accounting publication failed. A bounded diagnostic amendment in `internal/testutil/harnessaccount/command_linux.go` now reports only `scope_settled`, `controller_current`, a closed terminal stage and `terminal_failed`; raw errors/private values remain excluded. Admission, settlement, evidence validation and failure policy are unchanged. This is a validation repair, not a scheduler or accounting redesign. `tr-s005-format-20261024-04 make format` passed (`34711`, exit 0, 7916 ms); `tr-s005-generate-20261024-02 make generate` passed (`92615`, exit 0, 11200 ms). Both retain the usual `CGO_ENABLED=0` and explicit run ID settings.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s005-accounting-20261024-01 make test-slice OWNER=harness.evidence_accounting ROWS=harness.evidence_accounting.behavior.invocation_accounting_preserves_complete_kernel_74dff0b392,harness.evidence_accounting.behavior.invocation_accounting_rejects_malformed_counters_5eb3794701 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` is running (`78757`). The exact E-033 full regression command is running again as `tr-s005-regression-20261024-03` (`47455`) to identify the finalization boundary. Required performance observations will retain strict accounting; neither failed regression observation is accepted as a resource or final success claim. WF-07 remains IN PROGRESS and WF-04 remains BLOCKED.

The focused accounting selection exposed a command-mode prerequisite: `TestAccountingKernelLifecycle` explicitly rejects an enclosing accounted invocation because its own delegated kernel-lifecycle probe must not escape that scope. `tr-s005-accounting-20261024-01` therefore failed (`78757`, exit 2, 4034 ms, cleanup 10/10); signal/counter/short-lived-child assertions passed, and the kernel-lifecycle assertion identified this prerequisite. The same exact two-row selection ran as `tr-s005-accounting-20261024-02` with `HARNESS_ACCOUNTING=off`, retaining basic diagnostics and cache-off, and passed (`44418`, native exit 0, 4092 ms, cleanup 10/10). Its source is `sha256:7bfeccf86e27180ecd1572a51886d063c20be8fc67d931855c32a663322fa3df`; graph `sha256:f601ae89efb4535f80f581dc494334a71172c146fa3745331c0d40e1920b11dc`. This verifies the accounting controller's own contract; it is not a resource observation. No host privilege or accounting boundary was changed. Regression-03 remains running to diagnose its independent terminal-accounting failure.

Regression-03 (`47455`) again passed all 46 units and cleanup 55/55 in 229937 ms, but native exit was 2. The new diagnostic identifies `scope_settled=true controller_current=true terminal_stage=receipts terminal_failed=true`: the kernel scope settled, while terminal receipt publication rejected the regression invocation. Its exact receipt-rejection cause remains unresolved; no complete accounting/resource claim is made. Source is `sha256:7bfeccf86e27180ecd1572a51886d063c20be8fc67d931855c32a663322fa3df`, graph unchanged. This is kept separate from product and cleanup outcomes.

Strict outer resource accounting was an additional diagnostic for this regression portfolio, not an S-005 functional acceptance requirement. Functional qualification is now running through the public default accounting mode: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s005-regression-20261024-04 make harness-contract SUITE=regression HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=off CARTULARY_HARNESS_CACHE_MODE=off` (`71160`). This does not waive any required performance gate: all five-suite paired observations must still have complete required accounting. `tr-s005-generated-20261024-02 make source-check CHECK=generated` (`27085`) and `tr-s005-markdown-20261024-01 make lint-markdown` (`46580`) are also running with `CGO_ENABLED=0` and their explicit IDs. S-005 remains IN PROGRESS until native functional completion is observed.

### E-034 — WF-07 / S-005 completion checkpoint

**Status: COMPLETE.** `tr-s005-regression-20261024-04` finished with native exit 0 (`71160`, 230671 ms), all 46/46 units passed and cleanup 55/55. Its exact command is in E-033; source base remains `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`, candidate digest `sha256:7bfeccf86e27180ecd1572a51886d063c20be8fc67d931855c32a663322fa3df`, graph `sha256:180b0ecd03b016fd92fa2d07b91124918f9c86170228513be5583b3874921d77`. Toolchain/system digests match E-030. Native success is observed independently of the passing summary.

Final closure checks passed: `tr-s005-generated-20261024-02` (`27085`, exit 0, 20069 ms) and `tr-s005-markdown-20261024-01` (`46580`, exit 0, 11220 ms, `adhoc/lint-markdown/tool-run-summary.json`). Exact commands, prior successful type/catalog/support/browser/process results, retained attachment references, native exits and source digests are in E-033. `git diff --check` passed. No source-bound check remains running.

**Coverage dispositions:** B-052 retains the canonical restore case, disposable source/target databases and runtime roots, restored workbook/query assertions, and process/startup/cleanup evidence. Browser dependencies now fully dispose before fixture process retirement; restored sessions remain inside the disposable deployment boundary. B-005 retains both independent frontend builds, immutable loaded-asset assertions and reload overlap, with fixture-owned build settlement after browser cleanup. The two affected visual harness support cases retain their identities and assertions while routes use common ownership. No baseline browser case was retired or promoted. Existing process regressions and the extended common cleanup row cover startup/body/interruption/cleanup classification, partial acquisition, late readiness rejection, synchronous/asynchronous cleanup failure, repeated disposal and redaction. The passing real restore attachment records closed=true, forced=false and successful terminal/source/target cleanup; the passing publication attachment retains nine unchanged asset digests across 24 reloads.

The final diagnostic-only accounting change does not alter browser/runtime behavior; its own signal/counter/kernel-lifecycle tests pass in their required unaccounted invocation. The prior passing browser and ownership results remain compatible with that bounded stderr-only change and the regenerated topology inputs. Full final-source qualification remains S-006's responsibility.

**Accounting limitation:** regression-01–03 remain failed outer invocations, including the two whose 46 test units passed. Regression-03 proves scope settlement and current controller identity, but terminal receipt publication failed. Their precise receipt-rejection cause remains unresolved. The successful functional regression uses outer accounting off; this is not performance evidence. The kernel-delegation self-test similarly requires outer accounting off by its explicit guard. No scope, validation or resource-accounting rule was weakened. Required V-008 browser pairs must still use complete strict accounting; any failure in those observations blocks performance acceptance. Ordinary browser selections recorded in E-033 retain their successful required-accounting commands.

**Remaining risks and next entry gate:** WF-04/S-004 is READY to qualify sharing, scope residue-sensitive assertions and consolidate only equivalent cases. RB-002 still blocks promotion until per-case ownership, independent order/interference qualification, exact execution and quarantine evidence close. Restore and publication remain isolated. Final V-001–009, explicit owner audits, broad product/UI checks, four performance pairs and CI cost comparison remain outstanding. `agent-finalize` has not yet run; retained-run maintenance still lacks a compatible full warm-check `RESULTS_DIR`. The complete cutover remains one merge/rollback unit. Update this tracker with the WF-04 result before entering S-006.

### E-035 — S-004 admission qualification in progress

WF-04 began after E-034. The candidate authored families now declare 85 formerly isolated rows / 94 cases `test_owned`; these declarations are provisional, not completion evidence. The six retained isolated rows are actual HTTP expiry (B-012), socket expiry (B-017), the new Network Flow rows expiry bridge, Reference Pack lifecycle (B-051), restore (B-049), and publication (B-001). Together the successor portfolio remains 91 rows / 100 cases. No baseline browser row identity is retired. S-002 already removed repeated Reference Pack presentation combinations from the real lifecycle; the remaining cases have distinct postconditions, so no further deletion is justified merely by similar setup.

Ownership inspection uses the frozen per-case obligations in section 2, not row counts as proof. Common actors, sessions, independent clients, secondary contexts, routes, barriers and feature jobs use S-001 ownership. The nine-case authentication accessibility row has nine fresh contexts/actors: anonymous/session loading, authenticated landing, forbidden, generic error, incident-state recovery, MFA challenge, enrollment, and revocation remain separate presentation obligations. The two-case Timeline row independently owns its incident, transactions, pending queues and routes in both cases. Neither multi-case row may qualify through only one selected case.

| Candidate source boundary | Owned state and residue independence |
|---|---|
| Authentication, deployment administration and incident directory | Fresh actors; unique created users/incidents; selected user-ID mutations and local guard responses; no deployment-wide administrator mutation. Directory entry now explicitly opens the Incidents surface after `/`, including startup that restores a workbook. |
| Administrative and membership audit | Real pagination queries bind actor/action/target or owned incident/user; timestamp/cursor checks retain their exact matching result sets. Presentation histories and delayed reads belong to one context. |
| Membership, metadata and lifecycle | Fresh incident plus owned member users; fixed page counts are incident-scoped. Live roles, drafts, replay, history and departure listeners are scoped to those identities/pages. |
| Collaboration, session recovery, Timeline and Evidence | Owned incidents, rows, sessions and secondary pages; revoke-all/logout affect owned users. Only actual socket expiry retains global clock isolation. Evidence jobs settle before context/session release. |
| Network Flow and workbook continuation | Owned incident/query/graph state; local expiry responses for recovery; real paging and account replacement retained. Claimed runtime compatibility remains separate from default. |
| Workbook preferences, extensions and shell | Per-actor/account and per-incident preference keys; claimed extension runtime retained; no worker provisioner preferences used as subjects. |
| Reference Pack presentation | Context-local catalog/job fixtures; only the real signed lifecycle changes global catalog state and remains isolated. |
| Accessibility, visual and harness capture | Fresh owned contexts, local presentation fixtures, local visual preference routes, fixed owned render fixtures and unique artifact identities. No golden changes. Publication remains isolated. |

The real fixture-state witnesses now seed an unrelated owned actor, incident, audit effects and completed verification Job before each case. Global incident-list emptiness is no longer an assertion: fresh actor memberships/preferences and exact owned-incident filtering establish the boundary. All live seed dependencies settle through the scope; durable records remain until cohort reset. Forward A→B and reverse B→A will run in separate clean public invocations, once per deterministic case identity/generation. Existing graph and pinned Playwright regression mechanisms remain the qualification path; no public shuffle input, scheduler or alternate runner is added.

Initial commands: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-format-20261024-01 make format` (native session `33895`); `make task-guide OWNER=<owner> DETAIL=coverage JSON=1` for harness.browser, module.auth, module.collaboration, module.evidence, module.incidents, module.networkflow, module.reference_data, module.timeline, module.workbook and web.design, retained under `/tmp/cartulary-s004-<owner>-guide.json`. Runtime results, compatibility/accounting, source/generated closure and final per-case dispositions are pending. S-006 remains BLOCKED.

S-004 promoted-row qualification command ledger (pending launch/results; each ID names its exact root under `.cartulary/test-results/`):

- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-harness-browser-20261024-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.timeline_investigation_capture,harness.browser.boundary_support.visual_preference_lifetime HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-module-auth-20261024-01 make test-slice OWNER=module.auth ROWS=module.auth.accessibility.verify_session_mfa_incident_forbidden_loading_an_e8f5d5f3a9,module.auth.browser.a_reviewer_visible_incident_edit_control_fails_w_b6248bf4c8,module.auth.browser.administrative_audit_read_recovery,module.auth.browser.administrative_audit_real_browsing,module.auth.browser.deployment_admin_revoke_all_invalidates_a_target_a09ed5d06f,module.auth.browser.the_ordinary_shell_keeps_deployment_user_adminis_8a57567508,module.auth.browser.verify_ordinary_login_incident_directory_entry_o_d66ea23dc1,module.auth.browser_stateful.browser_deployment_user_administration_works_on_d6d7ff0b86,module.auth.browser_support.supplemental_browser_verifies_ordinary_shell_sel_f8a9da1957 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-module-collaboration-20261024-01 make test-slice OWNER=module.collaboration ROWS=module.collaboration.browser.concurrent_edits_to_different_fields_on_the_same_7ee5cad0ab,module.collaboration.browser.live_updates_never_retarget_pending_local_edits_11699e9b2d,module.collaboration.browser.verify_conflict_resolver_actions_submit_public_m_0c77fae932,module.collaboration.browser_stateful.session_recovery_logout,module.collaboration.browser_stateful.session_recovery_revoke_all,module.collaboration.browser_stateful.verify_multi_client_live_row_update_presence_anc_4c65a275c1 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-module-evidence-20261024-01 make test-slice OWNER=module.evidence ROWS=module.evidence.accessibility.verify_evidence_icon_buttons_blocked_states_erro_0ec80f90e2,module.evidence.browser.timeline_related_role_loss,module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-module-incidents-20261024-01 make test-slice OWNER=module.incidents ROWS=module.incidents.browser.a_selected_incident_whose_membership_is_removed_21054ae899,module.incidents.browser.incident_admins_manage_memberships_on_the_ordina_f26145a162,module.incidents.browser.incident_creation_native_keyboard_payload_and_fi_b6b4631f5f,module.incidents.browser.incident_creation_preserves_chosen_navigation_af_24cb9bd4de,module.incidents.browser.incident_creation_uncertain_replay_and_committed_985931d51a,module.incidents.browser.lifecycle_acknowledgement,module.incidents.browser.lifecycle_boundaries,module.incidents.browser.lifecycle_departure,module.incidents.browser.lifecycle_replay,module.incidents.browser.lifecycle_workbook,module.incidents.browser.membership_audit_access_recovery,module.incidents.browser.membership_audit_read_recovery,module.incidents.browser.membership_audit_real_browsing,module.incidents.browser.membership_audit_response_inspection,module.incidents.browser.membership_management_departure,module.incidents.browser.membership_management_live_pages,module.incidents.browser.membership_management_mutation_recovery,module.incidents.browser.membership_management_self_actions,module.incidents.browser.metadata_authority,module.incidents.browser.metadata_departure,module.incidents.browser.metadata_live,module.incidents.browser.metadata_pending_departure,module.incidents.browser.metadata_recovery,module.incidents.browser.the_ordinary_authenticated_landing_screen_create_fa80fbfef5,module.incidents.browser.the_ordinary_shell_exposes_incident_discovery_ra_59c190a9b5,module.incidents.browser.workbook_home_and_incident_default_preferences_c_6591154e0f HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-module-networkflow-20261024-01 make test-slice OWNER=module.networkflow ROWS=module.networkflow.browser_stateful.pagination_recovery HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-module-reference_data-20261024-01 make test-slice OWNER=module.reference_data ROWS=module.reference_data.browser.administration_async_recovery,module.reference_data.browser.deployment_admin_browser_sessions_expose_referen_8c10ae0a9c HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-module-timeline-20261024-01 make test-slice OWNER=module.timeline ROWS=module.timeline.browser.a_reviewer_session_browser_flow_visibly_performs_c3aaf62d89,module.timeline.browser.draft_create_controls HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-module-workbook-20261024-01 make test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.preferences,module.workbook.browser.grid_autosave_closure,module.workbook.browser.ordinary_create_closure,module.workbook.browser.preferences_1,module.workbook.browser.preferences_2,module.workbook.browser.preferences_3,module.workbook.browser.preferences_4,module.workbook.browser.reference_admission_lifecycle,module.workbook.browser.verify_continuous_workbook_shell_composition_top_96ea2f5084,module.workbook.browser_stateful.preferences_extension,module.workbook.browser_stateful.query_continuation_session,module.workbook.browser_stateful.verify_timeline_public_mutations_remount_safe_tr_84f645d6c5,module.workbook.visual.preferences HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-web-design-20261024-01 make test-slice OWNER=web.design ROWS=web.design.accessibility.administrative_audit_browsing,web.design.accessibility.deployment_user_draft_credential_dialogs_ac3e5d7802,web.design.accessibility.incident_creation_accessible_keyboard_feedback_a_f2a19fbfdb,web.design.accessibility.incident_import_workflow,web.design.accessibility.lifecycle,web.design.accessibility.membership_audit_browsing,web.design.accessibility.membership_management_accessibility,web.design.accessibility.metadata_editing,web.design.accessibility.reference_pack_administration,web.design.browser.account_application_menu_keyboard_navigation_and_8cba739220,web.design.browser.incident_import_action_access_hardening,web.design.browser.incident_import_observation_recovery,web.design.browser.incident_import_real_workflow,web.design.visual.administrative_audit_browsing,web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde,web.design.visual.incident_import_workflow,web.design.visual.lifecycle,web.design.visual.membership_audit_browsing,web.design.visual.membership_management_visual,web.design.visual.metadata_editing,web.design.visual.reference_pack_administration HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off`


Initial qualification launches:

- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-order-ab-20261024-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_a_b_e9fdc93f9d,harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_b_a_a194a3225b HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` (native session `88737`).
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-order-ba-20261024-01 make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_b_b_ab488bf040,harness.browser.boundary_support.browser_fixture_state_remains_test_owned_for_a_a_577b9aa596 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=required CARTULARY_HARNESS_CACHE_MODE=off` (native session `48084`).
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-types-20261024-01 make lint CHECK=types` (native session `28191`).
- `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-regression-20261024-01 make harness-contract SUITE=regression HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=off CARTULARY_HARNESS_CACHE_MODE=off` (native session `72272`).

The promoted incident and web.design selections from the command ledger are also running (sessions `34251` and `84950`). Source mutations are paused until these bound invocations drain. Static inspection identified a stale graph regression expectation: its former isolated Network Flow recovery row is now a provisional owned candidate. The fixture must retain the new real expiry bridge as its isolated successor while continuing to test cross-owner packing and reset adjacency; any failure of the old expectation will remain recorded before repair.

S-004 initial terminal outcomes: format (`33895`) passed, native exit 0, 8102 ms; generate (`43986`) passed, native exit 0, 11250 ms; types (`28191`) passed, native exit 0, 26935 ms. All owner guides exited 0. A first functional plan read failed because its `/tmp` input file had not yet been created; after writing `{ "SUITE": "functional", "BROWSER_LANES": "4" }`, `make target-plan TARGET=browser-e2e DETAIL=graph INPUTS=/tmp/cartulary-s004-functional-plan.json JSON=1` passed (`44701`) and reports 13 cohorts. This static projection is not a performance result.

A→B (`88737`) and B→A (`48084`) both passed with native exit 0 in 334734 and 334746 ms, respectively, with cleanup 23/23 each. Their graph digests are `sha256:34d9031afa91878f6142b30b1e989c5cdbbcb90724c7ab2991eafead215af94d` and `sha256:cf935b824f62e6d32184b9365c9ceba9fde720e7b081643a4b8e82a8186a32df`. Published target → group → Playwright report references prove each invocation had exactly one group, no reset units, two passed cases, worker index 0 throughout and retry 0 for both cases. Actual start order is A-before-B at 20:05:15.178Z then B-after-A at 20:05:18.052Z; the separate run starts B-before-A at 20:05:35.767Z then A-after-B at 20:05:37.425Z on 2026-10-24. Each identity executes once in its clean generation. These are the existing real shared-fixture order witnesses; candidate product assertions additionally require their own selected-case results.

The 21-row web.design selection (`84950`) passed, native exit 0, 295406 ms, cleanup 31/31; graph `sha256:42554576106ea662347d786ba5c336782a85ab78e93f21146b41b2818e5e1530`. All three browser invocations use dirty candidate source `sha256:e62cc32b742a37b156bf660f447b677d1fe990fe4d27f83a89a89a3f423632d4` at the frozen base, with the unchanged E-034 toolchain/system digests. Authentication (`89431`) and workbook (`1783`) are now running the ledger commands. Incident and regression invocations remain active. An early regression `explain-run` was unavailable before terminal publication; attempted `DETAIL=summary JSON=1` readers were rejected because JSON is supported only for performance detail. Correct plain summary/children readers then succeeded for the completed runs; reader exits do not substitute for native test exits.

S-004 incident selection (`34251`) passed, native exit 0, 412246 ms. Regression (`72272`) exited 2, 472016 ms: the retained public log confirms the single browser-work-graph assertion at line 1494 expected two cohorts but correctly received one after recovery admission (`actual: 1`, `expected: 2`). This is a stale fixture assumption, not a scheduler defect. The failed invocation remains retained and does not establish V-007 closure. The repair will add the real expiry bridge to this existing selection and retain assertions for one shared cohort plus one isolated cohort and exactly one reset.

The six-row collaboration (`70362`) and one-row claimed Network Flow recovery (`49165`) ledger commands are now running. Authentication and workbook are still active; source remains fixed until this qualification batch drains.

Workbook (`1783`) passed, native exit 0, 219727 ms, cleanup 38/38, graph `sha256:1f9b5ffa4e3837dce3ec5f470fc6c3692c0d8f038e7dfc5e61a6f0b13b8b2397`. Incident cleanup was 23/23, graph `sha256:544ae446245e50b43650cc454aa19ad886a9b6f859a87a6af08adfbe641afebb`. Regression completed all 46 units (45 passed, one stale-expectation failure), with cleanup 55/55 and graph `sha256:180b0ecd03b016fd92fa2d07b91124918f9c86170228513be5583b3874921d77`. The remaining ledger commands were launched for harness.browser (`17835`), module.evidence (`84820`), module.reference_data (`62194`), and module.timeline (`61599`). All S-004 selections are now launched; pending native exits are not passes.

Authentication (`89431`), collaboration (`70362`) and Network Flow recovery (`49165`) passed with native exit 0 in 279156, 135135 and 173716 ms. These cover all 17 selected authentication cases (including the nine-case accessibility row), six collaboration cases and the owned claimed-runtime recovery case. Harness capture/preference, Evidence, Reference Pack presentation and Timeline selections are still running. No implementation repair has occurred between these browser observations.

All promoted browser selections completed with native exit 0. Inspection follows only published run-summary → target-summary → group → Playwright-report references. It reconciles exact authored file/title selections: 85 rows / 94 cases, each with exactly one passed result and retry 0, including all nine authentication accessibility cases and both Timeline mutation/recovery cases. An initial inspection joined describe ancestors to catalog leaf titles and therefore mismatched nine names; correcting that read-only comparison to the authored leaf-title contract reconciled the complete set. This inspection supports coverage disposition; it does not replace the final public evidence audit.

| Owner | Rows / cases / cohorts | Native session (exit 0) | Elapsed ms | Graph digest |
|---|---|---|---|---|
| `harness.browser` | 2 / 2 / 1 | `17835` | 220187 | `sha256:02b857b38a449c11c4e7f77098ff1ee0a89694d03cd6508e0c75504ea87c86b8` |
| `module.auth` | 9 / 17 / 4 | `89431` | 279156 | `sha256:69a2b80520a4cf0b93af2cc760c7e339831c83a3a99d11f88c37cc2f9de137c2` |
| `module.collaboration` | 6 / 6 / 2 | `70362` | 135135 | `sha256:e153e05e0c691d65e758ea548820195510199113ee77cb53f657f5fdfd911b1b` |
| `module.evidence` | 3 / 3 / 3 | `84820` | 243651 | `sha256:cba0c94e01021d084b27c73e190254a80c9e1aab4d175a3d217576a0a4964970` |
| `module.incidents` | 26 / 26 / 1 | `34251` | 412246 | `sha256:544ae446245e50b43650cc454aa19ad886a9b6f859a87a6af08adfbe641afebb` |
| `module.networkflow` | 1 / 1 / 1 | `49165` | 173716 | `sha256:bd0de13c4a5a124ef2dc27381193d4703eaf5161293dbc1f2d08a4ccae42c5c0` |
| `module.reference_data` | 2 / 2 / 1 | `62194` | 220116 | `sha256:cb09d4e13dd841671ee3985638d26defb0d2739d59a633254a610b532fbd9b8a` |
| `module.timeline` | 2 / 2 / 1 | `61599` | 220753 | `sha256:bb88e9da46be285ebcd608dd2cd462e31c08d303117541f96676896bdd071212` |
| `module.workbook` | 13 / 14 / 5 | `1783` | 219727 | `sha256:1f9b5ffa4e3837dce3ec5f470fc6c3692c0d8f038e7dfc5e61a6f0b13b8b2397` |
| `web.design` | 21 / 21 / 3 | `84950` | 295406 | `sha256:42554576106ea662347d786ba5c336782a85ab78e93f21146b41b2818e5e1530` |

Final small-owner cleanup outcomes are harness.browser 23/23, Evidence 31/31, Reference Pack presentation 23/23 and Timeline 23/23. Authentication cleanup is 35/35, collaboration 27/27 and Network Flow 23/23. No cleanup failures or product retries occurred.

After all browser invocations drained, the existing graph fixture added `pagination_expiry_bridge` to its claimed-runtime selection. Its assertions now require exactly four compatible owned rows together, the bridge alone, and one reset between these selected cohorts. No scheduler or production implementation changed. `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-format-20261024-02 make format` passed (`72881`, exit 0, 8025 ms); `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-generate-20261024-02 make generate` passed (`57285`, exit 0, 11490 ms). `git diff --check` passed.

Closure commands: `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-regression-20261024-02 make harness-contract SUITE=regression HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=off CARTULARY_HARNESS_CACHE_MODE=off` is running (`61506`). `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-generated-20261024-01 make source-check CHECK=generated` passed (`18772`, exit 0, 21245 ms); `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-catalog-20261024-01 make source-check CHECK=catalog` passed (`31518`, exit 0, 1941 ms). `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s004-markdown-20261024-01 make lint-markdown` passed (`23265`, exit 0, 11080 ms; `adhoc/lint-markdown/tool-run-summary.json`). Later checkpoint prose still requires final Markdown validation. S-004 remains IN PROGRESS pending the native regression exit; S-006 remains BLOCKED.

### E-036 — WF-04 / S-004 completion checkpoint

**Status: COMPLETE.** The qualified successor inventory is **85 test-owned rows / 94 cases plus six isolated rows / six cases**. The frozen 88-row / 97-case inventory retains every canonical browser row; S-002 adds two collaboration recovery rows and the real Network Flow expiry bridge. Baseline obligations, split successors and assurance layers remain in sections 2 and 7 and E-030. No browser identity is retired or aliased. The only row retirement is the obsolete server fault-control test already accounted for by E-032's authored migration.

**Final isolation inventory:**

| Row | Required boundary |
|---|---|
| `module.auth.browser_stateful.idle_expiry_fails_closed_after_the_shared_clock_6525f19d9d` | Actual HTTP session expiry and fresh login. |
| `module.collaboration.browser_stateful.deterministic_route_owned_browser_revocation_sou_a134c3e26b` | Actual socket/session idle expiry and pending-work recovery. |
| `module.networkflow.browser_stateful.pagination_expiry_bridge` | Real rows cursor expiry at the owner-defined boundary; all-route expiry/scope enforcement remains backend-owned. |
| `module.reference_data.browser.administration_real_lifecycle` | Global signed import/replay, activation, disablement, re-verification and refresh. |
| `module.recovery.browser.a_harness_owned_browser_fixture_captures_a_retai_d8973973b0` | Disposable restored deployment, databases, process and storage. |
| `harness.browser.boundary_support.frontend_artifact_lifetime` | Independent concurrent publications and immutable loaded assets. |

**Evidence:** E-035 records every exact command/root/native outcome, per-owner row/case/cohort totals, graph digests, failure and cleanup disposition. All 94 promoted cases passed once with retry 0. Independent A→B and B→A real fixture invocations used one worker and one cohort each, no intervening reset, and unrelated durable users/incidents/audit effects/terminal Jobs. Per-case source inspection and product selections establish owned selectors, preferences, local routes and settled work; the two order witnesses qualify the common resource boundary. Changed directory helpers explicitly open the requested surface; no assertion depends on deployment-wide emptiness. No extra consolidation removes distinct postconditions.

The corrected full harness regression passed (`61506`, native exit 0, 232601 ms, 46/46 units, cleanup 55/55), source `sha256:00b0012b3b7d49c658f095f27d764bda6b50b027d7778f40076b950edd4ba20e`, graph `sha256:180b0ecd03b016fd92fa2d07b91124918f9c86170228513be5583b3874921d77`, frozen base and unchanged E-035 toolchain/system. It preserves bounded packing, mixed-owner tuple compatibility, default/claimed separation, exact selection, early-stop/unexecuted accounting, cleanup/quarantine and zero retries. Generated and catalog checks pass on that source. The only difference from E-035's passing browser source is the corrected regression selection and its generated input projection; browser/support/product implementations and admissions are unchanged. Final broad acceptance still requires compatible evidence after finalization. The failed first regression and all rejected diagnostic invocations remain preserved; no successful rerun erases them.

**Compatibility and rollback:** 85 declarations change grouping through the existing scheduler. No product schema migration, TTL, retry, deadline, resource claim, golden or runtime-profile boundary changes. Durable owned records remain until the existing bounded cohort reset. Merge/rollback remains the complete cutover. E-034's strict outer regression-accounting limitation remains explicit; its default-mode functional regression is not a resource measurement.

**Next entry gate:** WF-05 / S-006 is READY. Finalize generated state, run `agent-finalize` (retained maintenance skipped if `RESULTS_DIR` remains unset), broad checks and explicit suites, lifecycle/regression/UI-review contracts, compatible owner evidence audits, four frozen baseline/candidate performance pairs and full-CI cost comparison. Retained scans, final isolation/coverage dispositions, obsolete guidance and migration/handoff closure must all be reconciled. RB-004 remains BLOCKED for acceptance until actual required-accounting performance evidence passes; it does not prevent starting the work that produces that evidence. No required final gate is waived.

### E-037 — S-006 final qualification in progress

WF-05 began only after E-036. Historical handoffs now carry explicit bounded-cutover supersession notes for the retired fault route/schema, former browser cleanup helper and obsolete isolation reasons. Their original inventories, failures, measurements and source-specific completion claims are preserved; they do not become alternate controlling trackers or candidate acceptance evidence. The current tracker labels its old consumer inventory as frozen-baseline history.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-generate-20261024-01 make generate` is running (`10787`). Applicable generated/schema/catalog/identity/surface checks follow, then `agent-finalize` before broad verification. No compatible successful full warm-check `RESULTS_DIR` is currently available: retained-run maintenance will be explicitly skipped because it is unset, not represented as passed. Required correctness, resource comparison, explicit owner audits and retained scans remain open.

Initial S-006 generation passed (`10787`, native exit 0, 10880 ms). Source-check commands use `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-<check>-20261024-01 make source-check CHECK=<check>`: generated (`61040`, 18454 ms), schema (`98265`, 6851 ms), catalog (`5876`, 1883 ms), identity (`97201`, 899 ms), surface (`87623`, 2531 ms), all native exit 0. `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-markdown-20261024-01 make lint-markdown` passed (`63560`, exit 0, 11170 ms).

A managed baseline worktree was created and attached at `/home/jochi/.codex/worktrees/testing-remediation-baseline/cartulary`, exact clean commit `5e2ed4a9cda0104c7c5755049f49c6713fe2fad1`. Its repository instructions and public help were read; no source files changed. This checkout is reserved for consistent preparation and the frozen baseline observations. Candidate finalization and baseline bootstrap are preparation, not performance samples.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-finalize-20261024-01 make agent-finalize` passed (`99516`, native exit 0, 28352 ms, cleanup 10/10), source `sha256:00b0012b3b7d49c658f095f27d764bda6b50b027d7778f40076b950edd4ba20e`, graph `sha256:12ea589f64ef55010707f0f2168419c3cd4263f0393cffebe61c7b3a98575479`. `RESULTS_DIR` was unset; retained-run maintenance is skipped, not passed. This precedes broad end-of-run verification.

In the clean baseline worktree, `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-baseline-bootstrap-20261024-01 make bootstrap` failed (`79949`, native exit 2) at the pinned browser trust-tool download. Earlier substep pass summaries are not overall success. Its declared `bootstrap-full/playwright-install/stderr.log` identifies `browser trust tool download failed`; the suggested `TARGET=bootstrap-full` reader is rejected as an undeclared public target, so the explicitly published bounded log was inspected directly. The existing candidate download is exactly the baseline's unchanged pin (615170 bytes, SHA-256 `e5b42390b02c21851bc04e9557ec48539a913f9b09637220c4b1db0d1386b3fb`). Those verified bytes were copied as a private cache file into the baseline worktree. No dependency version, source file, privilege, trust policy or checksum validation changed. `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-baseline-bootstrap-20261024-02 make bootstrap` is now running (`23298`) and will revalidate the archive through the public bootstrap path. Both attempts remain retained; neither is a performance observation.

Baseline bootstrap-02 passed (`23298`, native exit 0), with the canonical `bootstrap/tool-run-summary.json` retained in that worktree. No source changes were needed. Candidate implementation is frozen for broad qualification at commit `e39ba9ce9` on `codex/testing-remediation`; this is an intermediate review commit, not acceptance or merge. The complete cutover remains one merge/rollback unit. Documentation-only checkpoint updates may follow without changing executable inputs.

`CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-candidate-check-20261024-01 make check BROWSER_LANES=4 HARNESS_DIAGNOSTICS=basic HARNESS_ACCOUNTING=off CARTULARY_HARNESS_CACHE_MODE=off` is launched from the candidate root. This is correctness/preparation evidence, not a paired resource sample. Broader invocations with possible measurement intervals will run sequentially to avoid deliberate workload interference. Required five-suite performance samples will use strict accounting, identical lanes/preparation and fresh IDs after correctness closure.

Candidate check native session is `99656`; frozen implementation commit is `e39ba9ce90669dcbe396ff21d0f31b9f36133f50`. Its emitted source digest is `sha256:00b0012b3b7d49c658f095f27d764bda6b50b027d7778f40076b950edd4ba20e`, graph `sha256:879dc4dc5c07bae1e0b4d81eee172e432e2d4bd7625c7c0a14b83f4aa3ebc4b6`. The current public observation is still executing, without a native terminal result.

Candidate check-01 completed (`99656`, native exit 2, 638910 ms): 312/313 work units passed, with only `operation:harness.lint_biome` failing; cleanup 522/522 succeeded. Public `TARGET=lint DETAIL=logs` reports three warnings: two `useConst` warnings in `authoring-candidates.spec.ts` and `noNonNullAssertion` in the worker cleanup fixture. The current primary classification remains product/test_assertion_failure as emitted by the harness. This is not an overall successful full warm-check root and cannot authorize retained-run maintenance.

The correction binds barrier release callbacks directly to constants and captures the optional worker actor in a narrowed constant before its verification closure. It also removes a redundant initial no-op release assignment. Runtime ownership/order/cancellation behavior is unchanged. Candidate `e39ba9ce9` is superseded for acceptance; no performance observations had started. `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-format-20261024-01 make format` is running, followed by narrow lint/type validation, generation/finalization and a new frozen candidate. Failed check evidence remains retained.

Lint-repair preparation passed: format (`76555`, native exit 0, 8312 ms) and `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-generate-20261024-02 make generate` (`83209`, native exit 0, 11220 ms). A speculative task-guide lookup for `web.authoring` was rejected as an unknown active owner; no test selection was launched from it. This syntax-only correction requires the discovered frontend/type lint selections, followed by required broad verification; no redundant implementation-mirroring tests are added.

The lint repair passed `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-frontend-20261024-01 make lint CHECK=frontend` (`87525`, native exit 0, 2012 ms) and the corresponding `tr-s006-types-20261024-01 make lint CHECK=types` (`87883`, exit 0, 20664 ms). `CGO_ENABLED=0 CARTULARY_TEST_RUN_ID=tr-s006-finalize-20261024-02 make agent-finalize` passed (`40735`, exit 0, 26311 ms); `RESULTS_DIR` remains unset because the full check failed, so retained-run maintenance remains skipped. `git diff --check` passed. Only the two TypeScript files and this tracker changed since the first freeze; generation made no additional tracked changes.

The next required full check uses normal compatible-result reuse for correctness; previously passed unchanged units may be reused honestly. It is not a performance observation. Required resource comparisons and their CI comparison runs will disable execution-result reuse.

### Source evidence navigation

| Boundary | Inspected source evidence | Revalidation obligation |
|---|---|---|
| Frontend lifetime and authentication | [Browser fixtures](../../apps/web/e2e/fixtures.ts), [cleanup support](../../apps/web/e2e/support/runtime/cleanup.ts), [incident-directory helpers](../../apps/web/e2e/pages/incidentDirectory.ts), [session support](../../apps/web/e2e/support/auth/sessions.ts) | Trace all consumers of changed ownership helpers; qualify failure and interruption paths. |
| Scoped data and presentation | [Administrative audit](../../apps/web/e2e/administrative-audit.spec.ts), [metadata presentation](../../apps/web/e2e/support/incidentMetadata.ts), [Evidence role loss](../../apps/web/e2e/timeline-related-evidence.spec.ts) | Confirm exact actor/resource boundaries and route settlement. |
| Backend enforcement | [Authentication integration](../../internal/modules/auth/integration_test.go), [last-admin store coverage](../../internal/modules/auth/store_test.go), [Network Flow routes](../../internal/modules/networkflow/routes_integration_test.go), [public error boundary](../../internal/platform/httpapi/internal_error_test.go) | Inspect the exact successor obligations before replacing browser assertions. |
| Harness and code generation | [Cohort packing](../../tools/harness/browser/browser-cohorts.mjs), [compatibility](../../tools/harness/browser/browser-compatibility.mjs), [browser graph](../../tools/harness/scheduler/work-graph/browser.mjs), [row migration accounting](../../tools/test_catalog_row_migrations.json) | Preserve selected-case identity and authored/generated consistency. |
| Fault-control retirement | [Negative control witness](../../internal/platform/harnessruntime/retired_controls_test.go), [HTTP serializer](../../internal/platform/httpapi/internal_error_test.go), Testing Harness §12.2.4 | Confirm the removed registry/middleware/schema/client path stays absent and real serializer coverage remains independent. |
| Specialized lifecycles | [Restore scenario](../../apps/web/e2e/restore.spec.ts), [restore fixture](../../internal/testutil/recoverybrowsertest/fixture.go), [visual harness](../../apps/web/e2e/visual-harness.spec.ts), [frontend publication](../../tools/harness/browser/frontend-publication.mjs) | Preserve resource admission, process ownership, and cleanup evidence. |

The implementation MUST append exact commands, revisions, result roots, cleanup outcomes, and coverage dispositions to the handoff. Prior failures and superseded evidence MUST remain identifiable. Concurrent changes MUST be reported and reconciled without automatically reverting another task's work.

## 11. Open Questions and Blockers

No unresolved product-design choice remains in this plan. WF-00 has closed baseline characterization and the selected assertion/consumer inventory. Human owner review is closed by E-026; candidate implementation, successor coverage, sharing qualification and final acceptance remain separate gates.

| Blocker | Affected work | Resolution | Status |
|---|---|---|---|
| RB-001 — Baseline characterization | Entry to owner amendments | Closed by E-024: unchanged-source 88-row / 97-case characterization and coverage dispositions. | DONE |
| RB-002 — Cohort promotion qualification | Cohort promotion | E-036 closes ownership, independent order/residue witnesses, all 94 selected successor cases and V-007 regression. Final-source acceptance remains in S-006. | DONE |
| RB-003 — Fault-control retirement | Cutover closure | Owner/interface review approved in E-026; implementation, projections and successor/retirement evidence closed in E-032. Merge remains the complete cutover. | DONE |
| RB-006 — Human owner/interface review | WF-01 closure and S-001 entry | User explicitly approved the E-025 owner amendment and mapping; recorded in E-026. | DONE |
| RB-004 — Performance evidence is absent | Cutover acceptance | Complete V-008 with valid required accounting. | BLOCKED |
| RB-005 — Plan mode prohibited file creation | Saving this handoff | Resolved when Plan mode ended and the requested handoff was created. Cutover implementation remains a separate task. | DONE |

These are explicit execution prerequisites, not permission to weaken acceptance criteria.

## 12. Binary Completion Criteria

**TR-REQ-009 — Hard cutover.** The change MUST merge as one coordinated pre-production cutover. It MUST contain no feature flag, legacy execution mode, duplicate active catalog, compatibility alias, or fallback to the retired fault-control path. Rollback means reverting the complete change set.

Acceptance requires all of the following:

- Every baseline production obligation has verified successor coverage.
- Shared tests have explicit ownership and no demonstrated order dependence.
- Remaining isolation matches documented production/resource requirements.
- Retired helpers, endpoint machinery, schemas, assertions, and selectors are removed.
- Owner requirements, authored contracts, generated projections, and active guidance agree.
- Correctness, accessibility, visual, integration, cleanup, and evidence checks pass.
- Measured elapsed time and CPU use improve under the defined protocol.
- No cost reduction comes from reduced assurance, hidden failures, retries, increased deadlines, or resource under-accounting.

| Outcome | Current statement |
|---|---|
| Planning completeness | COMPLETE for the bounded plan: exact baseline, selected assertion/owner mapping, helper migration boundary, dispositions, sequencing and exit gates are recorded. This does not qualify candidate behavior. |
| Per-slice readiness | S-001 is COMPLETE at E-028; S-002 is COMPLETE at E-030; S-003 is COMPLETE at E-032; S-005 is COMPLETE at E-034; S-004 is COMPLETE at E-036; S-006 is READY for final qualification. |
| Implementation authorization | Explicitly granted by the user's implementation request on 2026-10-24; evidence gates remain mandatory. |
| Execution completion | NOT COMPLETE. WF-00–WF-04, WF-06 and WF-07 are complete; S-006 and V-001–009 final acceptance remain incomplete. |
| Session write boundary | The bounded cutover is authorized. The first execution edit updates this controlling tracker; subsequent edits and validation results are recorded at workstream checkpoints. |

Handoff creation, baseline characterization, owner review, ownership qualification and scenario restructuring are complete. Fault retirement is complete. Specialized lifecycles are complete. Cohort qualification and S-006 final completion remain outstanding.
