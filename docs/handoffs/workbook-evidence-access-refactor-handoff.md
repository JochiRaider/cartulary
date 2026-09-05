# Workbook Evidence Access Refactor Handoff

## Baseline and authorization

- Execution authorized by the user's complete Evidence access implementation plan.
- Branch: `main`; commit: `edf5c4911dc47c40cfe22ebe6563b55541efb4af`.
- Initial and execution-start Git state: clean. Only root `AGENTS.md` applies.
- No commits, dependencies, backend workflows, or stored-data migrations are authorized.
- The digest is advisory and read-only. This is a new frontend slice, not the old digest refresh.
- Planning baseline slices passed with compatible cached evidence at
  `.cartulary/test-results/20260904T215750Z-p57486` and
  `.cartulary/test-results/20260904T215750Z-p57487`; these are not corrected-behavior evidence.

## Authority and allowed scope

Core 00 sections 1–5 establish precedence. Core 01 section 16 owns opaque
same-origin handles, fresh issuance without transaction IDs, preview policy,
and error vocabulary. Core 02 section 13 owns separate Evidence/blob machines
and consistency guards. Core 03 section 8 owns attachment and inline access;
REQ-03-291/292 and continuity requirements own inspector subjects and focus;
REQ-03-287/288 own closed-incident read access and rejected source writes.
Core 04 section 2.0A owns current authorization and concealment. Design sections
3, 7–8, 10.8, 11–12, and 14 own density, component presentation, local feedback,
and accessibility. Core access guards are evaluated before design compatibility;
presentation never grants access. No owner contradiction has been found.

The adopted Testing Harness NLSpec owns execution mechanics. Machine routing
comes from `contracts/verification`, `tools/test_catalog_owner.json`, and
`tools/test_families`. Generated-artifact policy and the visual maintenance guide
were inspected. No executable artifact may depend on Markdown.

Revalidated owner paths: [Core 00](../spec/00_document_set_status_and_precedence.md),
[Core 01](../spec/01_architecture_storage_and_view_contracts.md),
[Core 02](../spec/02_domain_model_schema_and_history.md),
[Core 03](../spec/03_workbook_interaction_collaboration_and_workflows.md),
[Core 04](../spec/04_security_deployment_and_conformance.md),
[domain navigation](../domain.md), and [design](../design.md).
Implementation/routing owners are `web.workbook`, `module.evidence`,
`web.application`, `web.design`, and `package.ui`; final inventory verification
also touches the existing `web.architecture` owner. No owner boundary moved.

Allowed paths are Evidence-related frontend source and consumers under
`apps/web/src`, focused browser/unit tests under `apps/web`, authored selectors
under `packages/ui-contracts/src`, necessary authored test/visual routing under
`tools` and `contracts/verification`, Make-generated derivatives, reviewed
affected goldens, and this handoff. Existing design projections provide the
required error family; no owner amendment is planned.

Completed inspector, density/renderer, responsive, recovery, view-bar, and
grid-state work remain regression baselines. API/routes, authorization, storage,
lifecycle, saved-view behavior, dependencies, and grid-vendor boundaries remain
outside the implementation scope.

## Workstream ledger

| Workstream | Status | Exit condition |
| --- | --- | --- |
| EA-01 — Baseline and characterization | DONE | Owner map, exact scope, proven defects, and evidence routes recorded. |
| EA-02 — Typed outcomes and presentation | DONE | One typed interpretation, safe copy, owner-correct announcements, focused tests. |
| EA-03 — Production UI and preview lifetime | DONE | Compact row/detailed inspector views, safe async lifetime, preserved contracts. |
| EA-04 — Browser and visual regression | DONE | Production-backed evidence and reviewed affected goldens pass. |
| EA-05 — Terminal checks and closure | DONE | Required checks, scope audit, and completed handoff; closed bytes receive final lint/whitespace validation. |

Only one workstream may be `IN_PROGRESS`. Complete its exit checks and log
paths, commands, results, failures, disposition, and next action before opening
its successor. Never advance past a blocked dependency.

## Findings and implementation decisions

Source-proven defects: access reason identity is flattened into strings;
message regexes determine announcements; handle issuance claims loaded preview
bytes; requested Evidence with an available blob is incorrectly inconsistent;
available/released rows with invalid blobs do not consistently receive the
required inconsistency state. Screenshot-only CSS enlarges rows, shrinks text,
and hides file inputs. These assertions are characterization targets, not authority.

Structural duplication: row and inspector reuse one oversized stack; upload
failures cross exception-message boundaries; unused lifecycle count helpers count
only accessible records. Production Timeline already uses projected record counts.

Risks requiring runtime characterization: painted clipping, focus/Escape,
preview-close and retarget races, feedback ordering, and attachment controls
after permission or lifecycle changes. Existing reset/unmount generations and
Timeline attachment continuity/replay are retained and extended only as needed.

The selected row design retains visible Preview and Download labels, a compact
Attach action, and an inspectable state indicator. Detailed feedback uses the
existing inspector Evidence section. Preview failure preserves focus; download
requires its separate existing access invocation.

## Verification routing

Inspected `make help`, `make help-all`, and module-author task guides for
`web.workbook`, `module.evidence`, `web.design`, `web.application`, and
`package.ui`. Start with exact semantic rows through `make test-slice` and
`make service-backed-test-slice`. Broader terminal verification includes
frontend type/unit/import-boundary/Biome, relevant webserver-backed/stateful/
accessibility/measurement/visual targets, generation/drift when authored inputs
change, Markdown lint, and Git whitespace checks.

Run `make agent-finalize` before broader terminal verification. Leave
`RESULTS_DIR` unset unless a qualifying full warm successful check exists and
record skipped retained-run maintenance accurately. Visual updates require
ordinary-run reconciliation, Make-owned promotion, manual review of every
changed PNG, and two fresh ordinary visual validations.

## Advisory dispositions

| Acceptance IDs | Disposition and evidence scope |
| --- | --- |
| A001–A003 | ADOPT: authority, bounded scope, clean baseline, and path/routing audit. |
| A004–A005 | ADOPT: existing tokens and graphite theme; no second design system. |
| A006 | ADOPT: fixed compact/default/comfortable geometry and painted controls. |
| A007 | ADAPT: preserve existing creation; attachment-only integration, no new create flow. |
| A008–A009 | ADOPT: current responsive bands, shell scroll ownership, safe navigation. |
| A010–A011 | ADOPT: existing inspector dispatch, subject invalidation, focus and scroll continuity. |
| A012–A015 | ADAPT: preserve attachment transaction/replay and existing editing/recovery baselines. |
| A016–A017 | ADOPT: typed async states and authorized-data retention. |
| A018–A021 | ADOPT: separate Evidence dimensions, names, announcements, component states, virtualization. |
| A022–A023 | ADOPT: exact production captures and semantic context-scoped selectors. |
| A024–A026 | ADOPT: no Markdown dependency, Make-generated outputs only, owner boundaries. |
| A027 | ADOPT: completed execution/evidence/rollback record, no automatic next seam. |

R001–R015 and R013's native control semantics are adopted within owner limits;
R008 never invents unknown-error recovery. R016–R025 and R035 are adapted to
fixed-height desktop rows, local disclosure, current tokens, and owner routes.
R026–R034 remain rejected: no cybersecurity palette, decorative motion,
dashboard-card redesign, generated design system, or incidental test authority.
No acceptance entry is currently N/A; final results are recorded at closure.

## Execution log

| Workstream | Paths and commands | Result and disposition | Next action |
| --- | --- | --- | --- |
| EA-01 start | Git status/HEAD/branch; root instructions; planned source, owner, digest, routing, and visual-policy reads | Checkout unchanged from planning. This tracker is the first implementation artifact. | Add and run failing characterization before production edits. |

### EA-01 exit evidence

- `make generate`: PASS, `.cartulary/test-results/20260904T220906Z-p62512`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.evidence_access_presentation_maps_blocking_evide_4b2a60e66a,web.workbook.regression.evidence_access_request_lifetime_ea01`:
  expected FAIL at `.cartulary/test-results/20260904T220931Z-p66064` (assertive preview blocker and false loaded-byte claim).
- `make test-slice OWNER=module.evidence ROWS=module.evidence.frontend_unit.verify_evidence_state_view_models_cover_requeste_634061c1db`:
  expected FAIL at `.cartulary/test-results/20260904T220931Z-p66053` (legal requested/available combination).
- `make test-slice OWNER=web.application ROWS=web.application.regression.workbookevidence_keeps_raw_storage_details_out_o_68c8061c4e`:
  expected FAIL at `.cartulary/test-results/20260904T220943Z-p20826` (unknown implementation identifier displayed).
- `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.accessibility.verify_evidence_icon_buttons_blocked_states_erro_0ec80f90e2`:
  expected FAIL at `.cartulary/test-results/20260904T220931Z-p66085`; real painted Preview bounds exceed clipping ancestors. Browser trace and assertion are retained under `browser-e2e-a11y/browser-groups/a11y-workbook-a11y`.
- `make browser-e2e-visual`: PASS at `.cartulary/test-results/20260904T220931Z-p66356`.
  Inspected `browser-e2e-visual/frontend-visual-reconciliation.json`, schema v2,
  status pass, pinned renderer and golden manifest pass. Existing screenshot
  overrides explain why this success does not prove production row geometry.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.evidence_access_request_lifetime_ea01`:
  expected FAIL at `.cartulary/test-results/20260904T221202Z-p59504`; soft assertions
  independently prove enabled attachment despite a read-only input, false preview
  success copy, and an obsolete completion reviving a closed preview.

All failures above are intentional characterization of this seam. Source
inspection proves string identity loss and screenshot substitution. Authorization
and broader retarget/deletion behaviors remain regression scenarios, not assumed
additional defects. EA-01 exit checks are complete. Next action: EA-02.

EA-02 started after the EA-01 exit record was written: consolidate typed identity, lifecycle, safe presentation, and attachment feedback.

### EA-02 exit evidence

Typed identity now crosses the protocol boundary through the service adapter
`publicErrorIdentity.ts`; Workbook failures retain validated reasons only.
`evidenceLifecycleViewModel` keeps custody, blob overlay, consistency, and local
access eligibility separate. Its unused access-based counting path is removed;
Timeline still consumes projected record counts. `evidenceAccessPresentation`
owns safe copy and derives announcements from existing owner presentations.
Preview blocks are polite; unknown preview/download failures remain unknown,
with no invented actions. Transport upload failures are typed. Both attachment
paths share slot/upload validation while retaining their finalization sequence,
accepted contracts, committed versions, and uncertain row-create replay.

- `make frontend-typecheck`: PASS, `.cartulary/test-results/20260904T222200Z-p62747`.
- `make format`: PASS, `.cartulary/test-results/20260904T222225Z-p63414`.
- `make generate`: PASS, `.cartulary/test-results/20260904T222229Z-p67672`;
  authored test routing generated only the expected topology render-index changes.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.evidence_access_presentation_maps_blocking_evide_4b2a60e66a,web.workbook.regression.evidence_access_error_identity_ea02,web.workbook.regression.workbookz_mutation_runtime_semantic_command_ports_7b2f1d02a4`:
  PASS, `.cartulary/test-results/20260904T222304Z-p71151`.
- `make test-slice OWNER=module.evidence ROWS=module.evidence.frontend_unit.verify_evidence_state_view_models_cover_requeste_634061c1db,module.evidence.frontend_unit.private_object_blob_upload_hides_target_4b2f89cd31,module.evidence.frontend_unit.upload_retry_body_discard_697516c4d8,module.evidence.frontend_unit.atomic_timeline_attachment_row_retry_79bdc83a10`:
  PASS, `.cartulary/test-results/20260904T222304Z-p71161`.
- `make test-slice OWNER=web.application ROWS=web.application.regression.public_error_presentation_typed_mapping_1ea7e57fb4,web.application.regression.workbookevidence_keeps_raw_storage_details_out_o_68c8061c4e,web.application.regression.workbookevidence_accepts_only_public_evidence_ha_4ba022c249`:
  PASS, `.cartulary/test-results/20260904T222304Z-p71167`.
- `make frontend-import-boundary-check`: FAIL at
  `.cartulary/test-results/20260904T222304Z-p71359` because the registry adapter
  initially lived in shared presentation. Moved it to the permitted service
  adapter layer; PASS at `.cartulary/test-results/20260904T222523Z-p74236`.
- Git whitespace and removal searches pass. No message regex or lifecycle-based
  count authority remains in the Evidence seam. No design-owner projection
  change was necessary. EA-02 exit checks complete. Next action: EA-03.

EA-03 started after the EA-02 exit record was written.

### EA-03 execution (in progress)

The production views now share `EvidenceAccessActions` semantics with distinct row
and inspector selectors. Visible Preview/Download labels, a native file picker
behind Attach, and an inspectable state button replace the tall row stack.
The existing layout's Escape handler closes previews before the inspector;
the preview layer remains usable when invoked from an overlay inspector.
No grid row-height, virtualization, selection, or scroll policy changed.

Request tickets extend the existing lifetime with record/version and ordering.
Close, retarget, version replacement, deletion, surface change, unmount, and
protected-data loss invalidate pending previews. Download ordering suppresses
obsolete effects. Protected-data failures invalidate outstanding access work and
request the existing query/root refresh. Attachment still settles its existing
mutation lease, even when its presentation lifetime has ended.

Additional browser finding: `evidence.upload_state` is a record projection, not
proof of a linked blob. An ordinary requested record without a blob reports
`pending`; `projectioncontract/contribution.go`, the authored Evidence view schema,
and the existing lifecycle route verify this boundary. Core 02 section 13 permits
blobless requested/pending/received/quarantined records. The model now records
projected pending as **unverified** linkage, without claiming an upload slot or
inventing a server state. It preserves the record lifecycle label and blocks
access. Known linked-blob combinations still receive the strict Core guards.
The public projection cannot distinguish a blobless quarantined record from an
invalid linked-pending record; the UI does not claim that distinction. No owner
contradiction or API amendment is needed for this conservative presentation.

Commands and corrections:

- `make frontend-typecheck`: failed at
  `.cartulary/test-results/20260904T223752Z-p77803` on an unexported density type;
  corrected to the existing grid-adapter type. Passed at
  `.cartulary/test-results/20260904T224005Z-p91078`.
- `make format`: first rejected unsorted authored test titles, then an implicit
  variable type at `.cartulary/test-results/20260904T223901Z-p78991`;
  both corrected. Passed at `20260904T223940Z-p83453`, `20260904T224109Z-p92793`,
  and `20260904T224411Z-p43618` under `.cartulary/test-results`.
- `make generate`: passed at `.cartulary/test-results/20260904T223946Z-p87575`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.evidence_access_request_lifetime_ea01`:
  all five request/feedback/attachment-lifetime scenarios passed at
  `.cartulary/test-results/20260904T224005Z-p90978`.
- `make test-slice OWNER=module.evidence ROWS=module.evidence.frontend.the_ordinary_evidence_workbook_surface_issues_pr_6b6cd04008,module.evidence.frontend.a_workbook_invalidation_refresh_visibly_updates_fd7a0de472`:
  passed at `.cartulary/test-results/20260904T224005Z-p90983`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e`:
  failed at `.cartulary/test-results/20260904T224121Z-p97238` on old Timeline raw
  upload copy and a transient attachment label cleared by refreshed row version.
  Corrected assertions use safe feedback and the operation announcement.
- `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.accessibility.verify_evidence_icon_buttons_blocked_states_erro_0ec80f90e2`:
  failed at `.cartulary/test-results/20260904T224120Z-p97003` on the projected
  pending distinction above. This was an implementation assumption, not a fixture
  defect; the model was corrected and its owner-combination tests extended.

EA-03 remains in progress pending production browser and corrected integration
exit checks. No golden has been changed.

### EA-03 exit

Corrected exit checks passed:

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.evidence_access_request_lifetime_ea01,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e`
  at `.cartulary/test-results/20260904T224423Z-p47939`.
- `make test-slice OWNER=module.evidence ROWS=module.evidence.frontend_unit.verify_evidence_state_view_models_cover_requeste_634061c1db`
  at `.cartulary/test-results/20260904T224423Z-p47947`.
- `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.accessibility.verify_evidence_icon_buttons_blocked_states_erro_0ec80f90e2`
  at `.cartulary/test-results/20260904T224423Z-p47964`, including the production
  painted-row bound assertion that failed characterization, keyboard focus,
  distinct access actions, safe feedback, one polite announcement locus, contrast,
  and no focus trap.

EA-03 is DONE. EA-04 is now IN_PROGRESS: expand production profiles and races,
remove screenshot presentation substitution, reconcile and review affected goldens.

### EA-04 execution (in progress)

The Evidence capture no longer installs `installFeP6EvidenceAccessVisualStyle`
or injects Evidence fixture markup. Its declared selector targets the production
workbook work area and includes compact rows plus the detailed Evidence inspector.
The final intended viewport remains 1280×720, with the existing renderer, theme,
default density, masks, and right-edge grid scroll. Inspector scroll is normalized
to its Evidence section. Intermediate inspector candidates were rejected because
visual review exposed clipped row controls; no candidate has been promoted.

Verification and corrections so far:

- `make generate`: passed at `.cartulary/test-results/20260904T225021Z-p470`.
- `make frontend-typecheck`: failed at `20260904T225035Z-p3860` on the new
  accessibility fixture's missing CSRF helper import and a DOM Node type;
  corrected, then passed at `20260904T225356Z-p6406`.
- The eight-row Evidence `service-backed-test-slice` at
  `.cartulary/test-results/20260904T225034Z-p3724` passed all six webserver-backed
  scenarios and the stateful access/authorization scenario. The expanded a11y
  row failed on the missing CSRF helper. The selected rows are the complete
  `module.evidence` webserver-backed inventory, its stateful row, and its a11y row.
- `make browser-e2e-visual` at `.cartulary/test-results/20260904T225035Z-p3938`
  reached only the expected `evidence-affordance-states.png` pixel mismatch.
  Reconciliation: 93 capture intents, 94 committed goldens, zero missing goldens,
  zero ambiguous mappings, all 26 registered fixtures resolved. The one temporarily
  unconsumed Timeline golden follows the failed earlier assertion; retain it.
  Baseline full reconciliation already accounts for it. Actual and expected
  Evidence PNGs were manually inspected: production labels fit the normal rows;
  the old override image clips most state/attachment content.
- Expanded a11y at `.cartulary/test-results/20260904T225356Z-p6323` passed density
  and text-spacing profiles, then detected partially scrolled controls after a
  viewport resize. Corrected the test's horizontal normalization to align the
  full production action group, rather than only its Preview button.
- A11y/visual slice at `.cartulary/test-results/20260904T225602Z-p55809` passed
  painted controls, responsive inspector preview/Escape focus, 200% zoom, native
  file-picker activation, and held-response dismissal before the close fixture
  failed for an omitted existing `base_incident_version`. Corrected the fixture.
  Its visual comparison remained an expected pixel mismatch; the overlaid
  1280-pixel inspector candidate was rejected as described above.
- New expected-failure checks proved two additional lifetime integration defects:
  `.cartulary/test-results/20260904T225841Z-p8190` retained an obsolete row's
  `Pending` label after retargeting; `20260904T225841Z-p8200` proved the narrow
  inspector's polite feedback was inside the inert background overlay.
  Supersession/dismissal now removes only matching operation feedback, and the
  single announcement channel lives outside the inert overlay while remaining
  within the workbook lifetime. Corrected hook tests passed at
  `.cartulary/test-results/20260904T230058Z-p57804`.
- `make frontend-typecheck` at `20260904T230058Z-p57934` identified a missing
  `unsupported_preview` member in the test fault helper's reason union; correction
  is pending the current browser result. This is test typing, not runtime policy.

All abbreviated run directories above are under `.cartulary/test-results`.
No golden has yet been promoted. EA-04 remains in progress.

Further EA-04 evidence:

- `.cartulary/test-results/20260904T230058Z-p57823` verified the corrected narrow
  announcement channel, then proved that native disabling on a blocked preview
  dropped focus and prevented Escape from closing the inspector. Access blockers
  now use `aria-disabled` plus guarded activation, preserving the focused button;
  protected access loss still disables controls. The previously characterized
  superseded preview label is also cleared without erasing newer feedback.
- `.cartulary/test-results/20260904T230339Z-p11457` passed the hook and workbook
  integration rows; `20260904T230339Z-p11572` passed typecheck. A11y at
  `20260904T230339Z-p11476` passed the explicit focus/Escape and closed-incident
  assertions, then reached the generic focus-trap check before the final fixture
  reload had finished. An explicit workbook-ready/painted-control wait corrects
  that test lifecycle.
- Render diagnostics at `20260904T230339Z-p11476` establish the actual clipping
  dependency: the Evidence grid shell was 1020px wide, but its scrollport retained
  a 1248px minimum. The earlier attribution to the responsive band was incorrect;
  both 1280px and 1440px are in the adopted base band. The Evidence consumer now
  selects the adapter's existing `fillViewportInline` capability, as Timeline
  already does. This preserves the authored density and internal horizontal
  scrolling while allowing the Evidence pane to shrink beside its inspector.
  Other generic surfaces retain their existing configuration. The temporary
  1440px capture experiment is reverted; the original 1280×720 viewport remains.
- Production selector contexts are checked in the existing package selector row.
  `make test-slice OWNER=package.ui` passed at
  `.cartulary/test-results/20260904T231012Z-p75323`.
- The request-lifetime row passed again at
  `.cartulary/test-results/20260904T231012Z-p75318`, including suppression of
  retargeted download effects and cleanup of their obsolete pending feedback.

- A11y at `.cartulary/test-results/20260904T231012Z-p75343` exposed a closed-startup
  query cancellation: closure arrived during the initial Evidence query and
  aborted the read without replacing it. The trace established an authorized
  closed incident and the aborted query; this was not an authorization denial.
  Expected-failure unit evidence at `20260904T231533Z-p33702` independently proved
  the canceled read. Core 03 closed-incident read requirements justify the narrow
  correction in `useGenericSurfaceQuery`: incident closure no longer cancels a
  current authorized read; protected-data and other invalidations retain their
  cancellation behavior. No mutation authority changed.
- `make generate` passed at `20260904T231446Z-p30462` after routing the closed-read
  characterization under the existing generic query row with Evidence collaboration.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.use_generic_surface_query_owner_fab9b06dde,web.workbook.regression.evidence_access_request_lifetime_ea01,web.workbook.regression.evidence_access_presentation_maps_blocking_evide_4b2a60e66a`
  passed at `.cartulary/test-results/20260904T231728Z-p39104`.
- `make test-slice OWNER=module.evidence ROWS=module.evidence.frontend_unit.private_object_blob_upload_hides_target_4b2f89cd31,module.evidence.frontend_unit.upload_retry_body_discard_697516c4d8,module.evidence.frontend_unit.atomic_timeline_attachment_row_retry_79bdc83a10`
  passed at `.cartulary/test-results/20260904T231728Z-p39113`. The shared upload
  adapter now uses the existing error resolver for HTTP authentication failure;
  a 401 retains authentication identity without reading its response body or
  replaying the upload. Completed downloads retain their feedback across inspector
  selection; only outstanding retargeted download requests are canceled.

- The complete expanded Evidence a11y scenario passed at
  `.cartulary/test-results/20260904T231728Z-p39134`: all density/text-spacing
  profiles, responsive row and inspector controls, scrollport geometry, 200% zoom,
  polite blocked feedback outside inert content, focus preservation and Escape,
  native picker keyboard activation, held-response dismissal, closed-incident read
  controls, disabled attachment writes, accessible names, contrast, and focus traversal.
- The fresh full ordinary visual run at
  `.cartulary/test-results/20260904T231728Z-p39423` failed only the intended
  Evidence affordance comparison. Its reconciliation again has zero missing or
  ambiguous mappings and all 26 registered fixtures resolved. The subsequent
  Timeline capture is retained, not deleted. This completes pre-refresh accounting.
- `make browser-e2e-visual-update` is now running. Accepted triggers are corrected
  production action rendering/feedback and removal of the prior presentation
  substitution, with a declared crop widened from grid to workbook work area so
  the actual inspector is also captured. Viewport remains 1280×720, zoom 100%,
  scale factor 1, default density, dark_graphite, and the pinned renderer. Masks
  and grid right-edge normalization remain; the inspector scrolls to Evidence.
  The state disclosure's preserved focus is now explicitly declared by the fixture.

### EA-04 visual review and final profile evidence

- `make browser-e2e-visual-update`: PASS, 12/12 units, at
  `.cartulary/test-results/20260904T232039Z-p33647`.
  Manually reviewed every changed image. Accepted only
  `apps/web/e2e/workbook.visual.spec.ts-snapshots/evidence-affordance-states-linux.png`.
  Its production rows have readable Preview, Download, Attach, and state controls;
  the inspector shows lifecycle, file state, and typed upload feedback. The
  focused disclosure and selected row remain visible. No row/font overrides or
  screenshot-only state markup remain. The native picker is hidden by the
  production Attach component, whose real keyboard activation is browser-tested.
- The update also changed 15 unrelated pixels in
  `network-flow-analysis-accepted-inspector-linux.png`. Compared old and current
  images; rejected that fluctuation and restored the original PNG from HEAD.
  `make generate` then regenerated the manifest at
  `.cartulary/test-results/20260904T232640Z-p90639`. No unrelated golden is changed.
- Accepted semantic owner row:
  `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`;
  stable fixture `visual.fixture.evidence_affordance` (`D-VFIX006`);
  capture `visual.capture.39bce0844ede522ac2e8`; project `chromium`, Linux.
  Scenario: “Capture evidence count, affordance, available, requested, pending,
  blocked, failed, inconsistent, preview, and download-handle state fixtures.”
  Golden SHA-256 changed from
  `ed5118c112db0241d965d6da204c7b27fe1f48bf9b19cbb496cd6e2e2409c6de`
  to `9268edbf048874c2659d2c04837ff3ce137472e99ae5511d4f1584d1730018e7`.
  This is the manifest's only changed entry. Renderer profile remains
  `visual.renderer.playwright_1_59_1_chromium_1217_linux_amd64`.
- The final Evidence a11y row passed at
  `.cartulary/test-results/20260904T232730Z-p94022` with exact responsive profiles
  1440×900, 1024×720, and 768×640, 200% zoom, text spacing, all three densities,
  reduced motion, and actual closed-incident Preview and Download invocations.
  Attachment remains disabled on the closed incident.
- First fresh post-promotion `make browser-e2e-visual`: PASS, 12/12, at
  `.cartulary/test-results/20260904T232730Z-p94143`. Reconciliation accounts for
  all 94 captures, 94 goldens, and 26 registered fixtures; zero missing, ambiguous,
  unresolved, or orphan entries. Renderer attestation and golden manifest pass.
  Browser execution is fresh (cache bypass), not reused baseline evidence.
- Final import-boundary check passed at `20260904T232058Z-p75336`; formatting
  passed at `20260904T232512Z-p81937` and `20260904T232609Z-p86377`.
  The second fresh ordinary visual comparison is running. No EA-05 work has begun.

EA-04 exit: the second fresh `make browser-e2e-visual` passed, 12/12, at
`.cartulary/test-results/20260904T233219Z-p87320`. Its reconciliation again
accounts for 94 captures, 94 goldens, and 26 fixtures with no gaps; renderer and
manifest checks pass. Both ordinary runs validate the final promoted PNG and
manifest. Final registry review corrected one descriptive phrase from preview
failure to the actual captured upload failure; no capture settings, application
source, or image bytes changed. `make generate` passed at
`20260904T233446Z-p35193`, and `make test-slice OWNER=web.design` passed 15/15
at `20260904T233459Z-p38326`, validating the corrected fixture/catalog inputs.
All EA-04 exit checks are complete. Next action: EA-05 finalization and terminal
verification, with no additional seam or golden changes planned.

### EA-05 execution

EA-05 started only after the EA-04 exit record was complete. The remaining work
is finalization, broader checks justified by the shared layout/query boundaries,
source audit, rollback preparation, and final handoff-byte validation.

- `env -u RESULTS_DIR make agent-finalize`: PASS at
  `.cartulary/test-results/20260904T233623Z-p87634` before broader terminal checks.
  `unit-artifacts/finalize-summary.json` reports generated outputs unchanged,
  zero updated files, and no rollback needed. Retained-run selection, closure,
  scheduler, and warm-run performance maintenance were skipped with
  `results-dir-not-provided`; no qualifying successful full warm run exists.
- `make frontend-typecheck`: PASS at `20260904T233646Z-p91041`.
- `make generate-drift`: PASS at `20260904T233659Z-p74996`.
- `make generated-artifact-policy-check`: PASS at `20260904T233659Z-p75023`.
- `make json-shape-check`: PASS at `20260904T233659Z-p75049`.
- `make frontend-unit`: FAIL, 434/435 units, at
  `.cartulary/test-results/20260904T233646Z-p91048`. The sole failure was
  `web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`:
  its exact inventory omitted the new owner-local slot/upload request helper.
  Added that exact path to `apps/web/src/testing/sourceOwnershipPolicy.test.ts`;
  transaction-ID and wire-intent confinement checks are retained. Inspected
  `make task-guide ROLE=module-author OWNER=web.architecture`, then ran
  `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`:
  PASS at `20260904T234017Z-p30967`. The full frontend unit target is rerunning.
  This focused inventory test is the only additional final-validation path.
- `make frontend-import-boundary-check`: PASS at `20260904T233953Z-p87024`.
- `make lint-biome`: PASS at `20260904T233953Z-p87036`.
- `make browser-e2e-a11y`: PASS, 12/12 units, at
  `.cartulary/test-results/20260904T233646Z-p91365`.
- `make browser-e2e-stateful`: PASS, 34/34 units, at
  `.cartulary/test-results/20260904T233646Z-p91171`.
- Corrected `make frontend-unit`: PASS, 435/435 units, at
  `.cartulary/test-results/20260904T234104Z-p38689`.
- `make browser-e2e-webserver-backed`: PASS, 60/60 units (83 semantic rows), at
  `.cartulary/test-results/20260904T233646Z-p91144`.
- `make browser-e2e-measurement`: PASS, 22/22 units (seven measurement rows), at
  `.cartulary/test-results/20260904T233953Z-p87077`. These are the routed Timeline
  and Network Flow grid baselines; Evidence's production geometry is verified
  in its accessibility row rather than a fabricated measurement owner row.
- `make lint-biome`: PASS after the inventory correction at
  `20260904T234359Z-p92285`. Preliminary `make lint-markdown` passed at
  `20260904T234036Z-p32268`; final bytes will be checked after tracker closure.
- Final lifetime audit added an expected-failure case to the existing request
  lifetime row: a same-version row receives invalid blob-state information while
  a download response is pending. At `20260904T234609Z-p998`, the older completion
  still invoked the download (two anchor invocations instead of one). The existing
  download invalidation loop now checks current row access eligibility as well as
  subject retargeting, discarding both obsolete effect and matching pending
  feedback. No new workflow or server state was introduced. The same focused row
  passed at `20260904T234713Z-p6245`; `make format` passed at
  `20260904T234650Z-p1812`, and typecheck passed at `20260904T234713Z-p6404`.
  Frontend checks, the Evidence accessibility/stateful slice, and an ordinary
  visual comparison are rerunning for this last correction. The completed broad
  measurement and unrelated browser rows are retained because their inputs and
  behavior are unaffected by this Evidence request guard.
- Final corrected frontend checks passed: `frontend-unit` 435/435 at
  `20260904T234713Z-p6411`, `frontend-import-boundary-check` at
  `20260904T234713Z-p6420`, and `lint-biome` at `20260904T234713Z-p6446`.
- `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.accessibility.verify_evidence_icon_buttons_blocked_states_erro_0ec80f90e2,module.evidence.browser_stateful.verify_evidence_attach_preview_download_blocked_b2c37d6c7b`:
  PASS, 13/13 units, at `.cartulary/test-results/20260904T234713Z-p6275`.
- The additional final-source `make browser-e2e-visual` failed at
  `.cartulary/test-results/20260904T234713Z-p6552` before image comparison in the
  unchanged entity-linking fixture: its dismissed mention chip did not appear
  within the existing assertion deadline. Evidence changes do not touch that
  Timeline mention path; the previous two full ordinary runs passed it. No golden,
  timeout, or unrelated fixture is changed. An isolated ordinary rerun will
  establish whether this failure is transient before closure.
- Isolated final-source `make browser-e2e-visual`: PASS, 12/12 units, at
  `.cartulary/test-results/20260904T235117Z-p43868`. All 27 visual scenarios pass
  without retries; reconciliation accounts for 94 captures/goldens and 26
  registered fixtures, with zero missing, ambiguous, unresolved, or orphan entries.
  The unchanged mention fixture passes, so the previous attempt is retained as
  transient browser evidence rather than a source or golden correction. Renderer
  and manifest checks pass. No image changed during either ordinary attempt.

EA-05 exit: final source typecheck, all frontend units, import boundaries, Biome,
owner slices, production browser/accessibility/stateful/measurement routes,
visual reconciliation, generation/drift/policy checks, preliminary Markdown lint,
and scope/rollback/whitespace audits pass as recorded above. No blocked dependency
or unresolved product defect remains. All five workstreams are DONE. The final
post-closure commands are `make lint-markdown`, `git diff --check`, and
`git diff --cached --check`; their final result is reported with the completed
handoff, without another source edit. No subsequent seam is started.

## Resulting responsibility boundaries

All source paths below are relative to `apps/web/src` unless qualified.

| Paths | Responsibility after this slice |
| --- | --- |
| `services/publicErrorIdentity.ts`; `shared/publicErrorPresentation.ts`; `workbook/adapters/workbookOperationErrorPolicy.ts`; `workbook/mutations/workbookOperationOutcome.ts` | Validate projected public reason identity through the shared validator, retain it in typed outcomes, and select existing error presentation families. No arbitrary error payload is used for Evidence copy. |
| `workbook/models/evidenceLifecycleViewModel.ts`; `workbook/evidence/evidenceAccessPresentation.ts` | Separate record lifecycle, upload overlay, derived consistency, access preconditions, and transient outcomes. One resolver owns Evidence labels, safe copy, tone, announcements, and available access actions. Counts remain projected record counts. |
| `services/workbookEvidence.ts`; `workbook/features/evidence/createUploadedEvidenceBlob.ts`; `workbook/features/evidence/createEvidenceAttachmentPort.ts`; `workbook/timeline/adapters/createTimelineEvidenceAttachmentAdapter.ts`; `workbook/timeline/hooks/useTimelineEvidenceAttach.ts` | Typed private-upload feedback and shared slot/upload validation. Existing attachment/finalization, committed versions, atomic Timeline creation, uncertain replay, and transaction recovery remain with their existing owners. Opaque handle validation remains in the service adapter. |
| `workbook/features/evidence/EvidenceAccessActions.tsx`; `workbook/features/evidence/useEvidenceWorkbookBindings.tsx` | Distinct compact row and detailed inspector views; request ordering, lifetime invalidation, local feedback, single announcement channels, and workbook preview/focus integration. |
| `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`; `workbook/components/GenericWorkbookSurface.tsx`; `workbook/layout/WorkbookSurfaceLayout.tsx` | Existing subject/interaction authority and semantic continuity are wired into Evidence. Demonstrated dependencies use the existing grid fill capability, place announcements outside inert content, and close preview before inspector on Escape. |
| `workbook/query/useGenericSurfaceQuery.ts` | The characterized closed-startup read is allowed to finish. Closure still disables source writes; protected-data cancellation behavior is preserved. |
| `packages/ui-contracts/src/entityEvidenceSelectors.ts`; its `index.ts` and `workbook-interaction-selectors.test.ts` | Stable row/inspector selector contexts, with unchanged default row IDs. No new UI package or vendor-grid import. |
| `apps/web/e2e/evidence.spec.ts`; `evidence-integration.spec.ts`; `workbook.a11y.spec.ts`; `workbook.visual.spec.ts` | Production behavior, attachment, authorization, geometry, keyboard/focus, responsive/density profiles, and visual regression evidence. Screenshot layout substitution is removed. |
| `tools/test_families/{module.evidence,web.application,web.workbook}.json`; `tools/frontend_source_ownership.json` | Authored semantic verification and source ownership for the focused changes. |
| `testing/sourceOwnershipPolicy.test.ts` | Exact owner-local wire-intent inventory includes the extracted Evidence upload helper; existing confinement and identity assertions remain enforced. |
| `tools/frontend_visual_fixture_registry.json`; the single Evidence golden; `tools/frontend_visual_golden_manifest.json`; `tools/execution_topology_render_index.json` | Declared production capture and reviewed image; manifest and topology derivatives generated through Make. |
| This handoff | Controlling execution, owner decisions, advisory dispositions, evidence, limitations, and rollback. The digest stays read-only. |

Focused unit changes accompany the service, shared error resolver, operation
policy, presentation, lifecycle, handle command port, and generic query files.
The new `useEvidenceWorkbookBindings.test.tsx` characterizes request lifetimes;
`WorkbookShell.surfaces.test.tsx` verifies integrated Evidence/Timeline behavior.
These are implementation evidence, not new product authority.

## Compatibility and known limits

Proven corrections are distinct from structural consolidation: typed identity
loss, raw feedback, incorrect live priority, false loaded-byte claims, legal
lifecycle rejection, actual row clipping, stale preview revival/pending feedback,
inert inspector announcements, blocked-button focus loss, and closed-startup
query cancellation each have source or failing characterization evidence above.
Splitting views, consolidating slot/upload feedback, and removing unused
access-derived count helpers are behavior-preserving structural changes.

The existing Evidence query projection uses `pending` when no blob is linked.
It cannot distinguish that case from a linked pending blob. The UI therefore
marks this derived consistency as `unverified`, shows the actual record lifecycle
and “File pending,” and does not claim an upload slot or permit access. Direct
blob-state inputs still enforce the Core 02 consistency guards. This frontend
limitation requires no new server enum or API inference.

Handle issuance remains fresh per invocation, with empty request bodies and no
transaction ID. The client validates opaque same-origin handle paths without
treating their tokens as identity. Opening an iframe is not proof that bytes
loaded; download feedback claims only that the download was requested. Server
authorization remains the authority on every invocation. Unknown failures have
safe copy and no invented recovery action.

No API, route, schema, storage, authorization, server lifecycle, dependency,
saved-view, or stored-data migration changes are included. No commits are made.
No following seam is authorized or started.

Final scope audit: 42 source/test/routing/artifact files plus this handoff;
38 modified tracked files and four new source/test files. Only the Evidence
affordance PNG changed. Git remains on the original `main` commit with no staged
files. The digest, dependencies, backend/API/SQL, and generated runtime roots are
untouched. Only the Make-generated topology index and golden manifest changed
among generated artifacts. Added-source review finds no Markdown dependency,
vendor-grid import, raw-copy classifier, token persistence, or compatibility
wrapper. Git staged/unstaged whitespace checks and the complete reverse-patch
dry run pass.

The complete Go/backend, `check`, `ci`, and release/claim-publication gates were
not run: this bounded frontend change has no affected backend owner inputs, and
the selected full frontend/browser routes cover the demonstrated dependencies.
Visual/a11y outputs remain implementation evidence. Retained-run maintenance was
skipped as documented above, not represented as successful full warm evidence.

## Final advisory acceptance evidence

The dispositions above remain ADOPT or ADAPT within this bounded seam; none is
N/A. ADAPT entries preserve the already-completed creation/editing/recovery
baselines rather than opening new work. No recommendation establishes product
authority, and R026–R034 remain REJECT.

| Acceptance IDs | Completed evidence and interpretation |
| --- | --- |
| A001–A005 | Owner-to-change and source inventories above; no theme, token registry, dependency, or authority change. |
| A006 | Production painted-control assertions at compact/default/comfortable densities; existing fixed row metrics and shared grid density remain intact. |
| A007 | Attachment sequence and atomic Timeline creation tests; ordinary creation policy remains the existing baseline. |
| A008–A009 | Exact responsive bands, below-minimum 200% zoom, text spacing, scrollport bounds, shell navigation, and full a11y regression. |
| A010 | Existing inspector dispatcher and role/closed/interaction authority; detailed Evidence section and context-scoped controls; full stateful inspector regression. |
| A011 | Preview Close/Escape restores semantic focus; held completion, retarget/version/deletion, scroll, selection, and full continuity regressions. |
| A012–A015 | Existing transaction identity, accepted contracts, committed versions, uncertain replay, conflict, editing, and recovery tests remain in the routed frontend/stateful suites. No new queue or recovery workflow. |
| A016–A017 | Typed progress/outcomes, ordered feedback, authorized refresh retention, protected-data invalidation, and closed-startup read characterization. |
| A018 | Core-valid direct blob matrix and safe ambiguous workbook projection; separate lifecycle/upload/access fields; all eight access reasons, unsupported/oversized preview, failed/inconsistent states, and projected record counts. |
| A019 | Names, keyboard activation, painted focus, polite blockers, assertive upload/unknown failures, contrast, non-color state labels, reduced motion, and full a11y suite. |
| A020 | Shared typed resolver unit matrix plus actual row/inspector browser states and production golden; no screenshot-only component variant. |
| A021 | Stable row IDs and existing adapter virtualization remain unchanged; lifetime/selection/scroll tests and terminal measurement route retain the performance baseline. |
| A022 | Reviewed D-VFIX006 capture and exact settings above; two fresh ordinary 94-capture reconciliations pass against the promoted manifest. |
| A023 | Context-scoped owner selectors and full `package.ui` slice pass; no duplicate row/inspector identity. |
| A024–A026 | Source/import audit, generated-policy, JSON-shape, drift, and owner routing; no Markdown dependency or new product/API/storage/auth behavior. |
| A027 | Sequential ledger, failures/corrections, reviewed visual record, compatibility limitation, terminal results, and atomic rollback recorded here. Final byte validation closes EA-05. |

## Source-only rollback

Prepared `/tmp/cartulary-evidence-access-refactor-source.patch` from the complete
42-file source/test/routing/golden delta, including all four new source/test files.
It excludes this handoff so the audit record survives a rollback. SHA-256:
`cad782daced7e14d7b39cdc6e821ebd5f9eeff436fa0729b7983e2face38edd1`.
`git apply --reverse --check /tmp/cartulary-evidence-access-refactor-source.patch`
passed. The atomic reversal, if separately chosen, is
`git apply --reverse /tmp/cartulary-evidence-access-refactor-source.patch`.
It has not been applied. Recheck against any later edits before reversal; do not
use a blanket reset. There is no runtime or stored-data rollback step.
