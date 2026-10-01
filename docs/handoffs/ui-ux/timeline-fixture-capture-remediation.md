# Timeline fixture and capture-contract remediation

Follow-up: [all Timeline shell captures](timeline-shell-capture-expansion.md) expands
the rich dataset beyond the three primary shells and records its own promotion evidence.

Status: complete. All six phase exits are satisfied; two fresh ordinary visual runs pass against the final manifest.
Baseline: `6456013e8` (current main). Earlier review authority: `84424f70`.
Renderer: existing Playwright 1.63.0 / Chromium 1243 profile, unchanged.
This handoff is supporting material; executable behavior never depends on it.

## Decisions and owner mapping

| Concern | Adopted owner | Projection / implementation |
| --- | --- | --- |
| Ten source fields, default order, 24-field contract | Core 01 §7.4.1 | Timeline v2, unchanged |
| Representative content, capture identity, framing and source preservation | Design §15, D-VFIX-001, D-AC-064 | Existing visual registry v6; shell scenario |
| Recipe admission, isolated adapters, owner workflows | Testing Harness §8.0A, REQ-818 / AC-134 | `tools/harness/fixtures/timeline-investigation/` and two adapters |
| Metadata-only normalization, readiness and restoration | Testing Harness §8.0A, REQ-819 / AC-135 | Normalization policy v1, receipts v1, E2E preparation |
| Lifecycle, timestamps, evidence and mentions | Existing Core 01–03 | Public owner routes; semantic read-back |
| Authorization and request-time loss | Core 04 | Existing Timeline role integration and review-profile scenarios |

Dataset `timeline-investigation-rich-v1`, revision 1, uses recipe schema v1 and
`cartulary.view.timeline.v2`. No production API, service, dependency, database,
saved-view or application-data migration is required. Source values, associations
and workflow outcomes must agree for the common core; runtime IDs, timestamps
and incidental versions do not need to be equal across disposable incidents.

## Investigation narrative

Mira Chen and Jordan Ellis investigate unexpected activity by a fictional archive
service account. The ordered source events are authentication-anomaly,
unexpected-token-use, script-execution, scheduled-task-created,
outbound-connection, staging-activity, source-correlation, containment-requested,
host-isolated, evidence-collected, analyst-review and corrected-assessment.
The recipe is the literal narrative source: distinct authored UTC and local
strings, source excerpts, varied summaries, documentation addresses and reserved
example domains. Its first RAW excerpt deliberately includes
`2026-04-18T14:12:34Z` and `123e4567-e89b-42d3-a456-426614174000`.
Script and formula-looking source remains inert text.

The common twelve rows precede 36 visual or 54 review continuation rows. Optional
associations are distributed deliberately. `source-correlation` is reviewed with
one available evidence association; `evidence-collected` has two available
files, several tags and mixed relationships. `analyst-review` is superseded by
`corrected-assessment`. Creating fully populated rows remains rough.

A dismissed mention seeded before navigation is absent from active collections.
Design §13's existing inspector contract shows only **session-observed** dismissed
items in Relationships and uses History for durable changes. This revision does
not create a persistent dismissed-mention list or replay an already completed
action to manufacture browser session state.

## Field-to-scenario matrix

Every field below is enumerated by the executable assertion fixture and checked
against the actual view schema. Technical envelope `record_id` and `row_version`
are verified separately. Hidden derived fields do not add golden columns.

| Actual field key | Required evidence |
| --- | --- |
| `timeline.date_entered_text` | All core rows meaningful; unchanged authored date; separate null/empty owner cases |
| `timeline.analyst_text` | Mira Chen and Jordan Ellis preserved as source, independent of account metadata |
| `timeline.mitre_stage_text` | Varied authored investigation stages |
| `timeline.device_object_text` | Multiple meaningful host/object descriptions |
| `timeline.ip_address_text` | Documentation-range addresses; exact source preservation |
| `timeline.activity_utc_text` | Twelve distinct authored instants; generated/preserved/mismatch/unavailable/empty owner variants |
| `timeline.activity_local_text` | Authored local strings with base conversion disabled; explicit clear intent |
| `timeline.raw_activity_text` | Multiline excerpts; inert HTML/formula; embedded ISO and UUID survive preparation |
| `timeline.activity_synopsis_text` | Short, medium and long source summaries; no cosmetic fill |
| `timeline.data_source_text` | Varied authentication, process, network, collection and analyst sources |
| `timeline.host_refs` | Zero/one/multiple; manual and alias-driven automatic resolution; unresolved and dismissed exclusion; ordered source tokens |
| `timeline.identity_refs` | Zero/one/multiple; actual resolved semantic identity and unresolved source tokens |
| `timeline.tags` | Zero/one/multiple; tags-only does not enrich |
| `timeline.attached_evidence_ids` | Zero/one/two owner-created associations; semantic identities resolved before comparison |
| `timeline.recorded_at` | Valid system time; never writable seed input |
| `timeline.edited_at` | Valid system time at or after creation; stable during capture preparation |
| `timeline.activity_sort_ts` | Exact parsed authored instant; core then continuation ordering; invalid/absent owner specimens |
| `timeline.date_entered_sort_day` | Owner derivation from authored day; invalid/absent owner specimens |
| `timeline.activity_time_pair_state` | Base disabled; separate empty/generated/preserved/mismatch/unavailable specimens |
| `timeline.capture_state` | Rough creation, material enrichment, tag-only preservation, authorized review and supersession |
| `timeline.replacement_record_id` | Correct semantic replacement direction or null |
| `timeline.evidence_count` | Zero/one/two available finalized files; owner derivation rejects incomplete/metadata-only inflation |
| `timeline.has_evidence` | Owner-grounded boolean; independently asserted |
| `timeline.has_unresolved_mentions` | Only active unresolved items contribute; resolved/dismissed excluded |

## Phase ledger

| Phase | Dependencies and risk | Exit / state |
| --- | --- | --- |
| 0 Baseline | Qualified current renderer; avoid attributing old failures to this work | Passed unchanged ordinary visual run `20261001T125324Z-p97418`, 12/12 units |
| 1 Contract adoption | Correct ownership; no Markdown executable dependency | Design and harness binding text adopted; supporting matrix here |
| 2 Data and adapters | Admission before mutation; stale versions, evidence and alias workflows | Passed recipe admission, 48/66 shared core, both seeded profiles, real browser operations, expanded owner cases and lease cleanup |
| 3 Capture preparation | Exact metadata rules; source preservation; focus; preserve sparse cases | Broad helpers retired; negative readiness/source tests and six rich inspector observations pass; final incidental capture review complete |
| 4 Promotion | Complete joins and explained differences; no hidden product regression | Passed complete 255-capture reconciliation and reviewed 72-image transactional promotion |
| 5 Handoff | Same final manifest; owned cleanup; no false conformance claim | Passed two fresh ordinary visual runs, exact joins, all supporting attachments and owned cleanup; finalizer and focused checks pass |

The adapters use existing public operations, re-read current versions, and verify
all final values. Their enclosing browser/service leases own success, failure and
cancellation cleanup; no borrowed service or data is deleted by the recipe.
Credentials, capabilities and runtime mappings remain private transient state.
Structural receipts contain dataset identity, revision, row totals and semantic
digest. Capture preparation restores metadata before subsequent observations.

## Evidence ledger

All run roots below are relative to `.cartulary/test-results/`.

| Command / purpose | Run root | Result |
| --- | --- | --- |
| Unchanged `make browser-e2e-visual` | `20261001T125324Z-p97418` | Pass, 12/12 |
| Recipe admission and 24-field classification slice | `20261001T140231Z-p51867` | Pass, 2/2 |
| Real seed, source/API preservation, normalization and three-frame anchor slice | `20261001T154606Z-p29642` | Pass, 11/11 |
| Both seeded review profiles | `20261001T145731Z-p19303` | Pass, 4/4; final shared-core read-back after auxiliary setup; private cleanup complete |
| Timeline role authorization, lifecycle and derivation slices | `20261001T132753Z-p13833` | Pass, 4/4 |
| Expanded null/empty/omitted/time-pair specimens | `20261001T133556Z-p47051` | Pass, 3/3 |
| Explicit time clear / clipboard recovery | `20261001T135113Z-p59212` | Pass, 1/1 |
| Evidence association create / patch owner coverage | `20261001T135112Z-p57980` | Pass, 3/3 |
| Existing inspector presentation unit slices | `20261001T153428Z-p72642` | Pass, 3/3 |
| Shell and inspector accessibility, without normalization | `20261001T142327Z-p26556` | Pass, 11/11 |
| Lifecycle presentation and characterization unit slices | `20261001T143359Z-p29421` | Pass, 3/3 |
| `make frontend-typecheck` | `20261001T154532Z-p24908` | Pass, 2/2 |
| `make frontend-import-boundary-check` | `20261001T141549Z-p49188` | Pass, 2/2 |
| `make lint-biome` | `20261001T154532Z-p24928` | Pass, 2/2 |
| `make lint-markdown` | `20261001T155252Z-p37469` | Pass; final handoff and refresh documentation |
| `make lint-scripts` | `20261001T150007Z-p3263` | Pass, 2/2 |
| `make json-shape-check` | `20261001T142327Z-p26486` | Pass, 3/3 |
| `make generate-drift` | `20261001T142703Z-p5034` | Pass, 4/4 |
| `make generated-artifact-policy-check` | `20261001T142703Z-p5050` | Pass, 3/3 |
| `make agent-finalize` | `20261001T154530Z-p24576` | Pass, 1/1; RESULTS_DIR unset, retained-run maintenance skipped |
| Timeline action / contextual creation visual slice after admission fix | `20261001T153428Z-p72602` | Pass, 11/11 against promoted images |
| Final ordinary `make browser-e2e-visual`, run 1 | `20261001T154606Z-p29821` | Pass, 12/12; 255 captures; zero retries, skips or join errors |
| Final ordinary `make browser-e2e-visual`, run 2 | `20261001T154606Z-p29828` | Pass, 12/12; same manifest; 255 captures; zero retries, skips or join errors |
| Complete ordinary candidate, before final incidental fixes | `20261001T142327Z-p26729` | 255/255 joined captures, no preparation failures; 72 image differences; no promotion |

Early failures exposed an incorrect import depth, missing automatic-resolution
alias, an editor-opening selection gesture, stale generated routing, and remaining
callers of retired normalization helpers. Each was corrected at its source.
No comparison tolerance or renderer setting was changed. The first rich inspector
check incorrectly assumed seeded dismissals appear as session-observed items;
the existing design owner clarified the check.

## Compatibility, risk and handoff

Existing scenario purposes and adopted D-VFIX identities remain. No new committed
goldens were added. The rich inspector's six review attachments and twelve
layout studies are supporting observations, not product defaults. Sparse and
error/pending/conflict/recovery/permission fixtures remain independent.

Retired helpers: body-wide `maskVisualDynamicText`, paragraph-wide
`maskIncidentIdentity`, and their wrapper. Unsupported recipe versions and unknown
keys fail closed; there is no legacy reader or compatibility shim.

The [refresh record](timeline-fixture-capture-refresh.md) records every affected
image, catalog row/scenario/fixture identity, normalization/framing disposition,
manifest digest and review outcome. The finalizer runs before broader final checks.
Retained-run maintenance is skipped when `RESULTS_DIR` is unset. Two fresh ordinary
visual passes against the promoted manifest completed successfully.

Rollback restores source, fixtures, projections, goldens and manifest together.
Disposable incidents require no data migration. Passing screenshots do not imply
accessibility, release, product-wide or Core 05 publication conformance.

## Review findings and deliberate capture choices

The six `rich-inspector-{details,relationships,evidence}-{1440x900,768x640}`
attachments show authored Details and multiline RAW, resolved and unresolved
Hosts/Identities with tags, and two available Evidence files with Preview/Download
controls. The narrow inspector occupies the available overlay; content scrolls
vertically without horizontal overflow. Required content and geometry are asserted
before attachment. The existing twelve current/capture/review layout studies
remain supporting attachments. Neither group adds a committed golden or D-VFIX ID.

The canonical shell shows the twelve core events before continuation rows, selects
`authentication-anomaly` and focuses its visible Date Entered cell at zero offsets.
Narrow and compact shells retain their saved-view/group/control pressure states.
Default field widths and ordinary cell truncation remain governed by the existing
product contract; this revision does not widen columns to fit the fixture.

Removing broad normalization also exposes authored dates, UUID-shaped fixture
values and source text in older scenarios. These are intentional preservation
changes. Evidence source fixtures explicitly author Requested/Received timestamps;
the read-only `evidence.edited_at` surface has a narrowly permitted metadata target.
The guard still rejects every writable grid field, authored inspector value and all historical values. Exact read-only Timeline Recorded/Edited and Evidence Edited metadata surfaces are admitted only by their field identity, write kind and explicit capture rule.

Three auxiliary scenarios use ordinary column selection: Decision supersession
and Task Requests omit Owner; Handoff reference authoring omits Outgoing Owner
while retaining the Incoming Owner picker. This avoids rewriting runtime user
references in source cells or editors. Production defaults and owner workflows
are unchanged. The former broad/magenta owner-input mask is retired.

The shared sparse inspector fixture and its reading/geometry inputs are unchanged.
After its sparse captures, the visual History scenario makes an explicit RAW
source edit and selects that exact supported rollback event. Its before/after
values include authored ISO/UUID literals and remain untouched. This replaces
selection of the first rollbackable association event, whose runtime references
were previously hidden by broad normalization. Audit incident IDs and accepted
supersession change-set IDs have explicit leaf metadata markers and exact counts.

The Network Flow contributor panel's short grid and read-only notice were compared
with the existing baseline and are unchanged. This fixture revision does not claim
to resolve that pre-existing presentation limitation or establish Network Flow
publication conformance.

Review used the repository's `cartulary-ui-review` skill and Make-owned artifact
imports. Every opened component passed byte-length and SHA-256 verification.
Private review bundles are temporary; no credentials, capabilities, private
mappings, screenshots or expired bundle links are committed in this handoff.

## Final integration evidence

Shared semantic digest (both independently seeded consumers verify the same
source values and semantic associations):
`ea764b9f54835155ab881ca90863bfa8c239c4cbe0345a70bbb6bfce57e8b169`.
The review adapter re-queries and verifies again after auxiliary preparation,
including its role and sample checks, before returning its receipt.

Promotion `make browser-e2e-visual-update` passed 12/12 units at
`20261001T144454Z-p79181`. Exactly 72 existing images changed; no image outside
the reviewed inventory changed. Sixty-eight promoted files matched the final
reviewed candidate byte-for-byte. Four additional promoted-byte observations
(account menu, Decision accepted recovery, creation zoom, Timeline authoring)
were reviewed and accepted with unchanged content and usable geometry.

Final golden manifest SHA-256:
`ce1508c7630322fd903941cf70c6d1abdc0487ff390fff947193ec9925e226bd`.

Both final runs used the manifest above and normalization policy SHA-256
`a8e78deac29c1cbbe57bdb437e57d3b84b033c7d4a979fb91b877d8cef4bca24`.
Each resolves 255 active goldens and all 29 registry fixtures with zero orphan,
missing or ambiguous joins; each retains all twelve layout studies and six rich
inspector attachments. All 47 Playwright scenarios passed without skips, retries
or flaky outcomes. Both owned cleanup receipts passed. All temporary UI-review
sessions reached `closed`; caller-owned scratch is removed at handoff.

Rejected setup attempts were retained as diagnostics and never promoted:

| Run root | Rejection / correction |
| --- | --- |
| `20261001T131226Z-p19338` | Early helper/import and preparation failures; corrected source contracts |
| `20261001T132640Z-p78173` | Undeclared capture policy entry; explicit identity added |
| `20261001T134026Z-p36866` | Source changed while a browser run was active; discarded run |
| `20261001T134607Z-p47054` | Wrong cell focus target and stale worker imports during edits; corrected, then froze source during runs |
| `20261001T135308Z-p90660` | History unit-reference cardinality mismatch; exact mounted counts corrected |
| `20261001T141600Z-p51381` | Lifecycle uncertainty and Decision recovery metadata counts; corrected against actual states |
| `20261001T142327Z-p26729` | Complete comparison with 72 differences; image review exposed historical association values needing a source-edit fixture |
| `20261001T143356Z-p28904` | Final ordinary candidate: 72 reviewed differences, zero preparation failures |
| `20261001T145710Z-p18683` and `20261001T145755Z-p37965` | Finalizer / JSON-shape rejected stale generated fingerprint after final seeder read-back edit; `make generate` refreshed it at `20261001T145944Z-p99672` |

The first post-promotion run, `20261001T150421Z-p8362`, completed all 255
captures with no preparation errors but rejected two image comparisons. The
Timeline action editor had a fractional anchor shift; contextual Decision recovery
had a footer text-paint difference. Metadata admission was reading descendant
rectangles inside closed disclosures. Admission now counts mounted leaves without
layout reads; the geometry owner remains responsible for readiness. The focused
visual and source-preservation slices passed, but full repeat
`20261001T152449Z-p26583` still found one Timeline authoring mismatch. Further
inspection identified variable line wrapping in the generated Recorded/Edited
Details values before the History action editor. The three Timeline action
captures now declare those two read-only metadata targets explicitly. Source
fields, source-derived sort values and historical values remain protected;
regression coverage verifies read-only presentation restoration and rejects
normalization of the authored Analyst field. No tolerance, viewport, persisted
value or targeting identifier was changed.

The independent full run `20261001T153910Z-p79456` passed 12/12. Its companion
`20261001T153613Z-p43988` exposed the same generated-timestamp wrapping variation
in Evidence affordances and three inspector History captures. All four differences
were inspected through digest-verified artifact imports. Their contents and
geometry remained usable; the differences were fractional text placement. Exact
Recorded/Edited rules now cover all 30 Timeline inspector capture identities with
mounted fields, and one Evidence inspector Edited rule covers its affordance
capture. This closes the preparation cause rather than retrying until two lucky
runs pass. Sparse source records, source-derived timestamps, historical values,
and comparison tolerances remain unchanged.

Early artifact-review sessions hit their bounded capacity or were interrupted;
ownership cleanup completed. A final import rejected a changed registry fingerprint;
the original registry source was restored before importing the matching run.
No unsupported artifact metadata was edited to bypass admission.

Broad `make check`, release checks and Core 05 publication checks are outside this
fixture-only verification claim. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset. No product behavior migration or renderer upgrade occurred.

## Exact focused verification invocations

These are task-specific invocations, not a second public target inventory.
For current routing, use `make task-guide` and the catalog. Each command maps
to the matching run root in the evidence ledger.

```bash
make test-slice OWNER=harness.browser \
  ROWS=harness.browser.boundary_support.timeline_investigation_recipe
```

```bash
make service-backed-test-slice OWNER=harness.browser \
  ROWS=harness.browser.boundary_support.timeline_investigation_capture
```

```bash
make service-backed-test-slice OWNER=harness.browser \
  ROWS=harness.browser.integration.ui_review_seeded_default,harness.browser.integration.ui_review_seeded_network_flow_claimed
```

```bash
make service-backed-test-slice OWNER=module.timeline \
  ROWS=module.timeline.integration.incident_role_authorization_review_and_supersede_62f3d384f6,module.timeline.integration.rough_and_uncertain_timeline_capture_preserves_n_a868f4aa14,module.timeline.support_integration.integration_route_idempotency_is_actor_scoped_ab7992bb7d,module.timeline.unit.timeline_projection_contract_shaping_exposes_the_0a2232ebe0
```

```bash
make service-backed-test-slice OWNER=module.timeline \
  ROWS=module.timeline.support_integration.integration_route_idempotency_is_actor_scoped_ab7992bb7d
```

```bash
make test-slice OWNER=module.timeline \
  ROWS=module.timeline.support_unit.clipboard_paste_parsing_mapping_provenance_and_b_43e8f0fd4f
```

```bash
make service-backed-test-slice OWNER=module.evidence \
  ROWS=module.evidence.store.attached_evidence_create_and_patch_aa33ea0168
```

```bash
make test-slice OWNER=web.workbook \
  ROWS=web.workbook.regression.inspector_history_semantic_detail,web.workbook.regression.inspector_saved_reading
```

```bash
make service-backed-test-slice OWNER=module.workbook \
  ROWS=module.workbook.accessibility.verify_inspector_tabs_relationship_links_evidenc_9ee9fd9ea2,module.workbook.accessibility.verify_keyboard_open_close_panel_navigation_esc_42b98cf08e,module.workbook.accessibility.verify_shell_regions_tabs_switchers_menus_inspec_c481421159
```

```bash
make test-slice OWNER=module.incidents \
  ROWS=module.incidents.frontend.lifecycle_product_surface,module.incidents.frontend.lifecycle_characterization
```

## Changed-file inventory

The 72 changed PNGs are enumerated individually in the refresh record. Other
changed/new files are listed below; generated projections were written only by
Make-owned generators.

- `apps/web/e2e/support/timeline/timelineInvestigation.test.ts`
- `apps/web/e2e/support/timeline/timelineInvestigation.ts`
- `apps/web/e2e/support/visual/metadataNormalization.ts`
- `apps/web/e2e/support/visual/richTimelineInspector.ts`
- `apps/web/e2e/support/visual/timelineShellAnchor.ts`
- `apps/web/e2e/support/workbook/contextualCreate.ts`
- `apps/web/e2e/visual-harness.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/src/app/AccountApplicationMenu.tsx`
- `apps/web/src/app/IncidentAdminPanel.tsx`
- `apps/web/src/app/IncidentLanding.tsx`
- `apps/web/src/app/IncidentLifecyclePanel.tsx`
- `apps/web/src/app/IncidentMembershipAuditPanel.tsx`
- `apps/web/src/app/LandingAdminLayout.tsx`
- `apps/web/src/networkFlow/NetworkFlowMappingPanel.tsx`
- `apps/web/src/networkFlow/NetworkFlowSavedGraphPanel.tsx`
- `apps/web/src/shared/workbookRecoveryNavigation.ts`
- `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`
- `apps/web/src/workbook/components/WorkbookActiveSurfaceFrame.tsx`
- `apps/web/src/workbook/components/WorkbookIncidentIdentityDisclosure.tsx`
- `apps/web/src/workbook/components/WorkbookRecoveryPanel.tsx`
- `apps/web/src/workbook/components/WorkbookSameFieldConflictResolver.tsx`
- `apps/web/src/workbook/features/coordination/DecisionSupersessionEditor.tsx`
- `apps/web/src/workbook/features/coordination/WorkbookDecisionSupersessionRecovery.tsx`
- `apps/web/src/workbook/inspector/WorkbookInspectorSavedDetails.tsx`
- `apps/web/src/workbook/inspector/presentation/WorkbookHistoryPresentation.tsx`
- `apps/web/src/workbook/inspector/presentation/WorkbookInspectorFeedback.tsx`
- `apps/web/src/workbook/inspector/presentation/workbookInspectorPresentationModel.ts`
- `apps/web/src/workbook/timeline/actions/TimelineCaptureRecovery.tsx`
- `apps/web/src/workbook/timeline/actions/TimelineSupersessionEditor.tsx`
- `apps/web/src/workbook/timeline/presentation/TimelineWorkbookInspectorRegion.tsx`
- `docs/design.md`
- `docs/guides/cartulary_visual_golden_maintenance.md`
- `docs/handoffs/ui-ux/timeline-fixture-capture-refresh.md`
- `docs/handoffs/ui-ux/timeline-fixture-capture-remediation.md`
- `docs/testing-harness-nlspec.md`
- `internal/modules/timeline/timeline_event_integration_test.go`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_visual_fixture_registry.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/frontend_visual_normalization.json`
- `tools/harness/browser/design-review-seed.mjs`
- `tools/harness/fixtures/timeline-investigation/expectations.json`
- `tools/harness/fixtures/timeline-investigation/index.d.mts`
- `tools/harness/fixtures/timeline-investigation/index.mjs`
- `tools/harness/fixtures/timeline-investigation/recipe.json`
- `tools/harness/generated-artifacts/check-json-shapes.mjs`
- `tools/harness_helper_ownership.json`
- `tools/harness_schema_attachments.json`
- `tools/schemas/cartulary.frontend_visual_normalization.v1.schema.json`
- `tools/schemas/cartulary.frontend_visual_normalization_receipt.v1.schema.json`
- `tools/schemas/cartulary.timeline_investigation_expectations.v1.schema.json`
- `tools/schemas/cartulary.timeline_investigation_recipe.v1.schema.json`
- `tools/test_families/harness.browser.json`
