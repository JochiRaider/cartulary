# Source-linked Note creation and recovery handoff

## Baseline, authority, and scope

Execution baseline revalidated: clean `main` at
`92f26e03bffbead7a03ed0c1cb32f9c8621b18d2`. Only the root `AGENTS.md` applies.
Core 01 REQ-01-305/311/329/330/615, inspector registry §7.4.1A, creation,
discovery, normalization and transaction replay own the interface. Core 02
§10.1 and relationship/revision owners preserve shared artifacts, attribution,
tags and source-to-artifact direction. Core 03 continuity, memory-local work,
REQ-03-299/100/301 and Core 04 authorization govern lifetime and concealment.
`docs/domain.md` and `docs/design.md` apply within their declared scopes.
The localized digest read order, NLSpec research and completed Task/Decision,
Timeline Evidence and revocation handoffs are advisory or implementation
evidence, not authority. No executable artifact depends on Markdown.

Scope: four contextual Note actions and Notes-sheet linked authoring; ordinary
Notes creation shares necessary authoring primitives. Existing-Note editing,
general relationship management, uploads, coordination creation and broad
workflow changes remain excluded. No commit, push, deployment, persistent
browser storage, dependency/lockfile or digest changes.

## Execution tracker

| Workstream | Dependency | State | Binary exit |
| --- | --- | --- | --- |
| LNC-01 baseline and owner reconciliation | none | DONE | One owner-backed model and bounded paths; projections characterized. |
| LNC-02 cohesive authoring | LNC-01 | DONE | All entry paths share retained fields and explicit source semantics. |
| LNC-03 atomic submission and recovery | LNC-02 | DONE | Immutable attempts, exact replay and retained acceptance pass. |
| LNC-04 integration and validation | LNC-03 | DONE | Security, service, browser, accessibility and reviewed visual evidence pass. |
| LNC-05 final validation and handoff | LNC-04 | DONE | Final checks cover completed handoff bytes and implementation. |

Only the current row may be IN_PROGRESS. Record paths, commands/results, run
roots, risks, disposition and next action before advancing; never advance past
a blocked dependency.

## Entry and field inventory

| Entry | Current binding and owner | Seed and intended source | Verification owner |
| --- | --- | --- | --- |
| Timeline `create_related.note` | `view_row_create/view_row_create_route`; Timeline related adapter | Empty field seeds incorrectly omit selected source | web.workbook; module.workbook |
| Hosts `create_related.note` | Same binding; shared inspector related workflow | Same defect for selected Host | web.workbook; module.workbook |
| Identities `create_related.note` | Same binding; shared inspector related workflow | Same defect for selected Identity | web.workbook; module.workbook |
| Evidence `create_related.note` | Same binding; generic inspector workflow | Same defect for selected Evidence | web.workbook; module.workbook |
| Notes-sheet linked draft | `createRecordLinkedNote`, generic mutation port | Explicit dropdown source from Timeline/Hosts/Identities/Evidence | web.workbook; module.artifacts; module.workbook |
| Ordinary Notes draft | `createViewRow` | No source | web.workbook; module.artifacts; module.workbook |

Every entry targets `cartulary.view.notes.v1` and an `artifact_type=note`
artifact. Declared create fields are `note.title` (`single_line_title_v1`),
`note.body` (`multiline_body_v1`) and `note.tags` (tag collection actions).
`create_inputs=[]`. At least one normalized title/body is required; whitespace,
tags and source alone never qualify. Title limit is 512 Unicode scalar values,
body limit 16384; body preserves interior LF/TAB. Attribution, timestamps,
linked count and technical identity are server-owned.

Source changes are explicit staged authoring actions. Applying another
authorized source preserves all raw fields; clearing source preserves them and
selects ordinary unlinked creation for the next fresh action (user-confirmed).
Changing grid selection, source version, sheet or inspector attachment does not
retarget or discard work. Selected identity survives missing/current pages.

## Owner correction recorded before application

1. REQ-01-615 adds `record_linked_note_create_route` for existing
   `POST /api/v1/records/{record_id}/linked-notes`; no new endpoint.
2. §7.4.1A specializes all four `create_related.note` features as
   `record_action/record_linked_note_create_route`, action key equal to feature
   key, workflow panel, editor role, mutating, no confirmation, selected-row
   preservation and same-shell failure. Disabled conditions are no selection,
   incident closed, authorization lost, row version changed and record deleted.
3. Specialized bindings omit `target_view_schema_id` and use empty
   `seed_bindings`: selected subject supplies editable route source context,
   not a new writable Note field. The action always authors a Notes artifact.
4. REQ-01-305/329 clarify retained/replaced/cleared source behavior; REQ-01-311
   identifies the explicitly chosen source as the invoking source for this
   action. Direction remains source -> created artifact, `references_artifact`,
   explicit-route `field_key=null`, derived by the server.
5. Core 04 AC-068 expects the chosen association unless explicitly cleared.
   Shared artifact shape, minima, field inputs, authorization, revision and
   attribution semantics are unchanged. Core 02/03 require no owner changes.

Authored projections: `contracts/view-inspector/index.json`, four source view
schemas under `contracts/view-schemas`, and discovery OpenAPI under
`contracts/openapi-source/owners/platform.viewschema`. Derivatives are generated
through Make, never hand-edited. Historical OpenAPI releases remain immutable.

## Gap ledger

| Gap and disposition | Remediation and areas | Rationale and long-term benefit | Compatibility and unresolved risk | Binary completion |
| --- | --- | --- | --- | --- |
| G1 confirmed routing/owner mismatch | Adopt correction above; typed discovery and semantic Note dispatch | One explicit operation prevents source omission and frontend exceptions | Same endpoint, shape and schema identity; generated client/server move together | Four canonical actions dispatch chosen atomic source; no ordinary fallback with a source |
| G2 confirmed component-local draft destruction | Note-specific retained owner and separate attachments | Navigation/version changes cannot lose or retarget text | Existing memory lifetime; security fencing needs tests | All five paths retain raw values and source; replacement requires discard |
| G3 confirmed inconsistent authoring/discovery | Declared normalization/minima; staged paged source picker | One authoring model preserves valid input and stable references | Existing fields/inputs only; paging and invalid input need tests | Title/body-only accepted; tags/whitespace-only rejected; off-page identity retained |
| G4 confirmed new ID on each invocation | Immutable exact operation/body/path/actor/incident attempt and synchronous guard | Uncertainty cannot produce a second logical creation | Fresh corrected action gets new ID; uncertain rejections cannot prove no commit | Duplicate activation sends once; response-loss replay is byte-identical |
| G5 confirmed insufficient completion evidence | Full receipt before effects; independent read-only reconciliation debt | Refresh failure cannot erase accepted creation | Never fabricate saved rows or regress versions | Late acceptance and failed refresh retain receipt; refresh retry sends reads only |
| H1 race/security hypotheses | Lifecycle, autosave/conflict, source deletion, revocation and stale-callback coverage | Preserve current authorization and concealment across all presentations | No wider security decisions authorized | Source/role/lifecycle/account matrix and browser evidence pass |

## Bounded implementation paths

- New Note owner/model/form/picker/recovery and tests under
  `apps/web/src/workbook/features/notes`; Note transport/discovery under adapters.
- Necessary bindings in workbook inspector, Timeline and generic/entity
  presentation, mutation runtime, shell infrastructure and collaboration.
- Existing Artifacts linked-note and Workbook route/discovery tests; only
  demonstrated Note-specific implementation violations justify backend edits.
- Semantic browser cases and support under `apps/web/e2e`, UI selectors if
  needed, authored frontend source ownership and test-family/verification
  routing under `tools` and `contracts/verification`.
- Generated outputs declared by generated-artifact policy; reviewed visual
  goldens/manifests only through public visual update target.

## LNC-01 execution

Planning checks passed: focused inspector model/workflow `make test-slice`
at `.cartulary/test-results/20260913T024021Z-p320`; Artifacts linked-note
atomicity `make service-backed-test-slice` at
`.cartulary/test-results/20260913T024103Z-p1007`. These characterize the unchanged
baseline, not the completed implementation. Baseline and root instructions were
revalidated again before this handoff was written.

Disposition: owner correction documented; apply projections and characterization
next. Risk: new binding must preserve exact feature coverage and generated
compatibility; no broader owner prerequisite identified.

## Compatibility, rollback and acceptance

Deployable outputs must keep owners, projections, generated contracts and
implementation consistent. Rollback reverses this seam's source changes as a
unit, never deletes committed Notes or links, and introduces no data migration.
No runtime or tests inspect prose to infer behavior.

Digest dispositions: ADOPT keyboard/focus, local error/recovery and accessibility
concerns; ADAPT compact source discovery and retained recovery to existing
workbook containment. REJECT generic modal/wizard, new workflow, design-system
or security prescriptions. Applicable acceptance includes A001 authority,
A010 exact inspector coverage, A011 continuity, A012 transaction identity,
A014 editing, A015 conflict, A019 accessibility and the visual/generation
boundaries. Final evidence and remaining skips will be recorded below.

Generation initially failed at OpenAPI compatibility in
`.cartulary/test-results/20260913T041741Z-p39029`: adding the route owner
changed the reviewed enum fingerprint. The compatibility report classifies
this as additive. Updated only that entry in the unpublished
`contracts/openapi-releases/2.0.0.change-set.json`; immutable 1.0.0 remains
untouched. This is necessary downstream compatibility maintenance.


LNC-01 final disposition: DONE. Core 01/Core 04 correction and authored
projections are coherent. Public discovery tests now assert the four atomic
bindings and disabled conditions without a writable source field.

- `make generate`: PASS, `20260913T041928Z-p40811`.
- `make test-slice OWNER=platform.viewschema`: PASS,
  `20260913T042010Z-p43948`.
- `make test-slice OWNER=package.view_contracts ROWS=package.view_contracts.frontend_unit.contracts`:
  PASS, `20260913T042010Z-p43953`.

All run roots are under `.cartulary/test-results/`. Task guides for requested
owners and projection owners were inspected. Remaining risk is implementation
migration; next action: LNC-02 retained Note authoring and paged source control.

## LNC-02 execution

DONE. Added `features/notes` owner, raw-presence model, shared form, staged
source control, presentation attachment, sheet adapter and retained-draft
recovery. Runtime/shell owns the lifetime; inspectors and generic Notes grid
attach to the same author. Semantic `note_create` dispatch follows the adopted
route. Neutral discovery reads are shared without workflow-owner dependencies.
Title/body omission and explicit blank clearing remain distinguishable; tag
removal only changes local additions. Source selection is paged, authorized,
staged and independent of retained identity.

Paths: `apps/web/src/workbook/features/notes/**`, Note discovery adapter,
`features/generic/useGenericCreateDraft.ts`, generic/entity/Timeline inspector
adapters, shared inspector routing/presentation, runtime/shell infrastructure,
`tools/frontend_source_ownership.json` and `tools/test_families/web.workbook.json`.

Commands/results:

- Initial authoring slice routing failed because test titles were not sorted;
  corrected the authored manifest before rerunning.
- Authoring slice failed at `.cartulary/test-results/20260913T043316Z-p48340`:
  an inline error changed the Title accessible name. Explicit control naming
  and description fixed it.
- Initial frontend type check failed at
  `.cartulary/test-results/20260913T043316Z-p48422`; corrected nullable identities,
  imports, and moved generic defaults away from untouched Note field presence.
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260913T043534Z-p51798` (2/2 units).
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.note_create_authoring`:
  PASS, `.cartulary/test-results/20260913T043724Z-p54349` (2/2 units; eight tests).
- `make format-frontend`: initial lint errors were corrected (control validation,
  label binding, dependency lifetime and unused scaffolding); final invocation
  passed. Intermediate `make lint-biome` failure is retained at
  `.cartulary/test-results/20260913T043548Z-p52316` for final-check comparison.

Risks/disposition: authoring behavior is covered; dispatch controls are handed
straight to LNC-03 for immutable submission, not considered mutation evidence.
No storage, endpoint, editing contract or neighboring owner behavior changed.
Next action: LNC-03 atomic transport, replay and acceptance/reconciliation tests.

## LNC-03 execution

DONE. Added immutable Note operation/attempt/receipt types and a Note transport;
the neutral record-mutation sender now admits linked-note POST alongside
ordinary create and patch. Fresh submission reserves synchronously, coordinates
source saves, verifies current discovery/authority, and checks the captured
source identity/version. Explicit source/access review retains all fields.
Uncertain replay uses the captured operation, API base, path/parameters, exact
body and original secure transaction ID. It rechecks edit authority without
fresh-source or open-incident requirements. Replay rejection cannot disprove an
earlier commit, so it retains uncertainty rather than unlocking a duplicate.

Complete validated acceptance precedes draft clearing and effects. Receipts
survive navigation, suspension and refresh failure; retirement fences callbacks.
Refresh debt sends reads only, including mounted surface and history reads.
No second relationship mutation follows linked-note creation.

Paths: Note owner/form/recovery/sheet adapters and new
`features/notes/noteCreateOperation.ts`, `adapters/createNoteCreateTransport.ts`,
`adapters/sendWorkbookRecordMutation.ts`, neutral authoring-record reader,
runtime/shell/collaboration adapters, and ownership/routing manifests. Renamed
the existing Timeline source-save coordination hook to
`timeline/hooks/useTimelineSourceWriteCoordination.ts`, with a structural neutral
port shared by its existing consumer and Notes; no Evidence workflow dependency.

Commands/results:

- `make format-frontend`: PASS after implementation.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.note_create_recovery,web.workbook.regression.note_create_authoring`:
  PASS, `.cartulary/test-results/20260913T044817Z-p58806` (3/3 units; twenty tests).
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260913T044817Z-p58869` (2/2 units).

Risks/disposition: unit evidence covers atomic route capture, source clearing,
same-frame activation, exact replay, late acceptance, malformed envelopes,
known rejection/new logical action, replay rejection, source versions/deletion,
source-save blocks, suspended acceptance, account retirement and refresh-only
recovery. Service/browser and full security/presentation integration are the
next dependency, not inferred from these tests. Existing server semantics are
unchanged. Next action: LNC-04 source/security/semantic integration and user-facing
evidence, including neighboring regression boundaries.

## LNC-04 execution

The completed execution below exercises complete entry paths against services
and browsers, including authorization, keyboard/accessibility, compact
presentation, and neighboring regression boundaries.

### LNC-04 additional confirmed receipt projection defect

Browser run `.cartulary/test-results/20260913T045858Z-p85734` passed all four
source-route creations but failed source-edit completion and exact-replay
completion. Captured real responses contain the existing `source_record_id`
and `link_type=references_artifact` fields. Discovery projected linked-note
success through the closed ordinary `ViewMutationData` schema, so generated
validation rejected actual atomic success envelopes. Earlier unit fixtures
omitted those fields and did not characterize this mismatch.

Exact correction recorded before application: within the Notes-specific Core 01
§7.4.1A routing clarification, specify that atomic linked-note success retains
the existing standard mutation receipt plus the chosen `source_record_id` and
server-derived `link_type=references_artifact`. Project that existing response
through a dedicated linked-note data/envelope schema for only this endpoint;
do not relax ordinary mutation envelopes. Validate receipt source equality in
the Note transport/owner and preserve the complete envelope. Add real-shape
transport and service/browser assertions. This is a Note-specific routing and
receipt clarification, with no new endpoint, request field, response behavior,
storage, authorization or relationship direction. Candidate compatibility
metadata and generated protocol outputs must move with the projection.

Remediation benefit: complete atomic receipts become usable recovery evidence
and detect wrong-source responses. Binary completion: real linked-note success
and exact replay clear only their matching draft, retain the full source
receipt, and recover read debt without another mutation. Residual risk: final
compatibility and projection drift checks remain required.

Compatibility characterization: the new linked-note response schemas are
classified additive, and replacing the two response `$ref` values is classified
breaking by the structural checker. The existing unshipped 2.0.0 candidate
ledger records both classifications honestly. Existing runtime response bytes
and the immutable 1.0.0 release remain unchanged. Generation initially stopped
on the four unrecorded changes at
`.cartulary/test-results/20260913T050344Z-p23516`; the candidate ledger was
updated from that actual report. A verbose rerun attempt was rejected by the
public input guard because `generate` does not admit `CARTULARY_OUTPUT_MODE`;
subsequent generation uses only its public inputs.

### LNC-04 additional confirmed Note create admission defect

The expanded atomic service slice failed at
`.cartulary/test-results/20260913T050655Z-p29258`: explicit blank title plus valid
body was rejected by the shared legacy line helper. Inspection also confirmed
that this helper lacked the declared title/body scalar limits and rejected
Unicode format characters beyond the adopted C0/C1 boundary. Corrected only
Note create admission: title/body normalization and limits follow Core 01
REQ-01-490/491; optional normalized-empty values enter creation as omitted
values and therefore store authoritative null. Note create tags follow
REQ-01-494 without broadening the shared helper or existing-row editing.
Patch/conflict callers keep their previous collection admission path. No owner
change is required: this repairs implementation against existing string owners.
The existing atomicity row now characterizes four source types, deleted-source
replay, fresh rejection after deletion, normalization/minima/clearing/lengths,
and rollback without partial artifacts, links, revisions or receipts.


### LNC-04 disposition and execution evidence

DONE. Every contextual action and the Notes sheet use the incident-owned Note
author. Semantic dispatch retains exactly-once feature coverage and disabled
rendering. Source autosave/conflict coordination, current authority review,
collaboration observations, monotonic history versions, reference-cache
invalidation, accepted receipts and read-only projection recovery are integrated.
No generic or Timeline legacy Note submission branch remains.

Additional exercised defects: unrelated discovery invalidation remounted the
staged source picker, and narrow recovery positioning clipped controls. The
picker now refreshes candidates while retaining staged sheet/identity and
scrubbing stale labels; the Note recovery panel stays within the viewport.
Successful contextual creation returns focus to its originating action when
still attached; completed recovery returns focus to the current workbook
surface. Neither callback moves focus into concealed or retired presentation.
Tests cover these fixes alongside implementation.

Paths: the bounded Note authoring and runtime/inspector adapters; new
`apps/web/e2e/linked-note-create.spec.ts` and
`apps/web/e2e/support/workbook/noteCreate.ts`; existing sentinel, accessibility
and visual rows; Artifact admission/atomicity tests; Workbook response
projection and owner-contract tests; source/test ownership manifests. The visual
inspector fixture now counts the specialized Note action semantically alongside
seven ordinary creation actions. No neighboring workflow behavior changed.

Commands/results (all from repository root):

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.note_create_authoring,web.workbook.regression.note_create_recovery`:
  PASS, `.cartulary/test-results/20260913T053305Z-p64965`, 3/3 units,
  23 tests. Covers raw presence/normalization/minima, paged/staged retention,
  actual inspector/grid attachments, double activation, immutable replay,
  malformed/wrong-source receipts, late acceptance, read debt, closure/reopen,
  role restoration, source deletion/newer removal evidence, suspension and
  account retirement.
- The same slice plus contextual Task/Decision authoring/discovery/recovery and
  Timeline Evidence authoring/recovery: PASS,
  `.cartulary/test-results/20260913T052437Z-p44101`, 8/8 units.
- `make service-backed-test-slice OWNER=module.artifacts ROWS=module.artifacts.linked_notes.artifact_linked_note_creation_is_atomic_across_f_bce1ac123b`:
  PASS, `.cartulary/test-results/20260913T051411Z-p61631`, 3/3 units.
- `make test-slice OWNER=module.artifacts ROWS=module.artifacts.source_mutations.artifact_admission_and_replay_hashing_are_owner_local`:
  PASS, `.cartulary/test-results/20260913T051411Z-p61639`, 1/1 unit.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.note_create_sources,module.workbook.browser.note_create_source_edit,module.workbook.browser.note_create_recovery`:
  PASS, `.cartulary/test-results/20260913T053224Z-p32552`, 11/11 units.
  Exercises four sources, title/body-only creation, tags/minima, exact direction
  and count, editable/cleared source, sheet navigation, commit response loss,
  exact replay and read-only refresh, including deterministic completion focus.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.note_create_revocation`:
  PASS, `.cartulary/test-results/20260913T052548Z-p49166`, 11/11 units.
  A real committed response held until after membership deletion cannot reopen
  the concealed workbook; the authenticated directory remains visible.
- `make service-backed-test-slice OWNER=module.links ROWS=module.links.browser.the_notes_tab_supports_browser_visible_creation_ae6bdcc2da`:
  PASS, `.cartulary/test-results/20260913T051624Z-p12619`, 11/11 units.
- `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.related_evidence_link_replay,module.timeline.browser.related_evidence_original_source`:
  PASS, `.cartulary/test-results/20260913T052715Z-p86962`, 11/11 units.
- `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.browser_stateful.incident_access_revocation`:
  PASS, `.cartulary/test-results/20260913T053024Z-p99324`, 11/11 units.
- `make test-slice OWNER=web.application ROWS=web.application.regression.app_session_lifecycle_d770146af2`:
  PASS, `.cartulary/test-results/20260913T053023Z-p99011`, 2/2 units.
- `make test-slice OWNER=module.workbook ROWS=module.workbook.unit.openapi_record_mutation_contract_12a2ffaaf6,module.workbook.catalog.linked_notes_create_contextual_artifact_links_9a09b2f951`:
  PASS, `.cartulary/test-results/20260913T052712Z-p82488`, 4/4 units.
- `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.generated_http_operation_bindings,package.protocol_ts.frontend_unit.family_registries_and_account_types_i2`:
  PASS, `.cartulary/test-results/20260913T052714Z-p83448`, 3/3 units.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.note_create_authoring_recovery`:
  PASS, `.cartulary/test-results/20260913T052107Z-p77805`, 11/11 units;
  keyboard, names/errors, picker Escape/focus, 390/768/1280 widths and 200% zoom.
- `make browser-e2e-visual`: ordinary characterization at
  `.cartulary/test-results/20260913T052711Z-p82328` reached 241 captures;
  236 existing goldens resolve, zero orphans/ambiguous mappings, all 28 registered
  fixtures resolve. Only the five expected new Note goldens are missing.
  Authoring, staged source and recovery actual images were inspected; compact
  layout, focus, scrolling and containment are acceptable. Promotion and two
  ordinary passes are the LNC-05 maintenance dependency.
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260913T053225Z-p32827`, 2/2 units.
- `make lint-biome`: PASS,
  `.cartulary/test-results/20260913T053306Z-p65364`, 2/2 units.
- `make frontend-import-boundary-check`: PASS,
  `.cartulary/test-results/20260913T053307Z-p65737`, 2/2 units.
- `make generate`: PASS,
  `.cartulary/test-results/20260913T052438Z-p44302`; final generation follows
  the completed routing inputs in LNC-05.

Earlier failures were related and resolved: combined accessibility/visual
`.cartulary/test-results/20260913T051623Z-p12384` found narrow clipping and new
missing baselines; full visual `.cartulary/test-results/20260913T052127Z-p7324`
found the stale ordinary-route count; browser
`.cartulary/test-results/20260913T052922Z-p66640` found staged discovery reset
and completion focus assertions. Type checks
`.cartulary/test-results/20260913T052551Z-p49770` and
`.cartulary/test-results/20260913T052923Z-p66938` caught invalid socket test
fixture variants; authoring `.cartulary/test-results/20260913T053226Z-p33042`
caught eager initial label scrubbing. Corrected fixtures and revision-only
scrubbing pass their subsequent checks. No tolerances, sleeps or golden updates
concealed these failures.

Risks/disposition: all five gap-ledger remediations meet their binary behavior
criteria in focused evidence. Atomicity and complete HTTP acceptance are
service-backed; security lifetime and stale callbacks are tested at owner and
browser boundaries. No persistence, new endpoint or authorization rule was
introduced. Next action: LNC-05 final source generation, finalization, reviewed
visual promotion and terminal verification covering this handoff's final bytes.

## LNC-05 execution

DONE. Owner guides rechecked for web.workbook, module.artifacts,
module.workbook and package.protocol_ts; earlier projection guides remain
applicable. RESULTS_DIR remains unset: no successful full warm run matches this
source, so retained-run maintenance will be reported as skipped.

Visual promotion intent: the adopted Note authoring/routing clarification adds
five new, nonregistry captures owned by
`module.workbook.visual.note_create_authoring_recovery` in
`apps/web/e2e/workbook.visual.spec.ts`. Existing goldens remain active. The full
ordinary reconciliation above accounts for their absence without ambiguity.
The new filenames are `linked-note-authoring-linux.png`,
`linked-note-authoring-narrow-linux.png`, `linked-note-source-narrow-linux.png`,
`linked-note-recovery-narrow-linux.png`, and `linked-note-recovery-linux.png`.
They use the existing pinned renderer, full viewport scope, 1280×720 and 768×640,
100% zoom, compact workbook density, existing dynamic masks and explicit drawer
or recovery anchors. No renderer, font, tolerance, fixture registry or existing
capture scope change is proposed. Only the public update target may promote
these candidates and their managed manifest.


### LNC-05 final verification evidence

Generation: `make generate` PASS,
`.cartulary/test-results/20260913T053450Z-p67320`.
`make agent-finalize` PASS before terminal verification,
`.cartulary/test-results/20260913T053508Z-p70324` (1/1 unit).
RESULTS_DIR was unset; retained-run duration/baseline maintenance was SKIPPED.
No successful full warm run was reused or claimed for this source.

| Command / selected scope | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make frontend-typecheck` | PASS, 2/2 | `20260913T053536Z-p74395` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `20260913T053536Z-p74414` |
| `make lint-biome` | PASS, 2/2 | `20260913T053536Z-p74438` |
| `make generate-drift` | PASS, 4/4 | `20260913T053536Z-p74181` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `20260913T053536Z-p74189` |
| `make json-shape-check` | PASS, 3/3 | `20260913T053536Z-p74197` |
| `make openapi-compatibility-check` | PASS, 4/4 | `20260913T053536Z-p74221` |
| `make test-slice OWNER=web.workbook` with the eleven rows below | PASS, 12/12 | `20260913T053536Z-p74297` |
| `make service-backed-test-slice OWNER=module.workbook` with four Note browser rows and Note accessibility | PASS, 13/13 | `20260913T053553Z-p7841` |
| `make service-backed-test-slice OWNER=module.artifacts` with the LNC-04 atomicity row | PASS, 3/3 | `20260913T053553Z-p7871` |
| `make test-slice OWNER=module.artifacts` with the LNC-04 admission row | PASS, 1/1 | `20260913T053553Z-p7891` |
| `make test-slice OWNER=platform.viewschema` | PASS, 1/1 | `20260913T053605Z-p39979` |
| `make test-slice OWNER=package.view_contracts ROWS=package.view_contracts.frontend_unit.contracts` | PASS, 2/2 | `20260913T053605Z-p40029` |
| `make test-slice OWNER=module.workbook` with both LNC-04 contract rows | PASS, 4/4 | `20260913T053605Z-p40097` |
| `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.generated_http_operation_bindings` | PASS, 2/2 | `20260913T053605Z-p40141` |

The final web.workbook ROWS selection was:
`web.workbook.regression.note_create_authoring`,
`web.workbook.regression.note_create_recovery`,
`web.workbook.regression.contextual_task_decision_authoring`,
`web.workbook.regression.contextual_task_decision_recovery`,
`web.workbook.regression.timeline_related_evidence_authoring`,
`web.workbook.regression.timeline_related_evidence_recovery`,
`web.workbook.regression.decision_supersession_canonical_composition`,
`web.workbook.regression.timeline_inspector_feature_controller_a38c2d6f71`,
`web.workbook.regression.inspector_create_related_workflow_stale_results_31f0cbe2d1`,
`web.workbook.regression.inspector_related_record_model_20a3b73f6d`, and
`web.workbook.boundary_support.workbooksurfaceownershippolicy_suite_85a208f2dd`.

The final module.workbook service ROWS selection was
`module.workbook.browser.note_create_sources`,
`module.workbook.browser.note_create_source_edit`,
`module.workbook.browser.note_create_recovery`,
`module.workbook.browser.note_create_revocation`, and
`module.workbook.accessibility.note_create_authoring_recovery`.
Other selected row IDs are recorded verbatim in LNC-04 above.

Advisory digest acceptance dispositions:

- A001–A003, A010–A012, A024–A027: ADOPT/PASS through the owner correction,
  semantic coverage, memory-local continuity, secure replay identity, source
  boundaries, generated drift and this handoff. Typed projections remain
  executable authority inputs; prose is human review only.
- A007, A014–A017, A019, A023: ADOPT/PASS for the bounded Note seam through
  minima/normalization, retained invalid input, source-save admission, distinct
  discovery/recovery states, keyboard/error/focus assertions, real service
  mutations and semantic source/record selectors.
- A004–A006, A008–A009: ADAPT/PASS using current workbook controls, color tokens,
  density, inspector containment and responsive shell. The compact Note form and
  viewport-bounded recovery use existing layout conventions; no theme, density,
  renderer or token registry was introduced.
- A013, A018, A020–A021: unchanged regression boundaries. This seam does not
  redesign queued cell edits, Evidence access, component variants or grid
  virtualization. Selected neighboring tests and full visual fixtures remain
  the evidence boundary; no new broad claim is made.
- A022: ADOPT/PASS. Reviewed capture accounting, promoted baselines and two
  fresh ordinary validation results are recorded below.

Compatibility remains the existing atomic endpoint and source-to-artifact
relationship direction. Explicit source clearing uses existing unlinked Note
creation. The only authored owner amendments are the bounded Core 01 Note
routing/context/receipt clarification and Core 04 AC-068 association wording.
The candidate OpenAPI ledger records structural response-reference changes;
no released contract history, storage schema, existing-Note edit behavior or
account-wide authorization rule changed.

Broad `make check`, release/conformance publication, benchmark/measurement,
security scans, migration replay, and deployment are not claimed: the task used
owner-selected checks and real browser/service fixtures for changed inputs.
There are no migration, dependency, platform or runtime-toolchain changes.
The digest and lockfiles remain unchanged. No commit, push or deployment ran.


Visual update: `make browser-e2e-visual-update` PASS,
`.cartulary/test-results/20260913T053536Z-p74509`, 12/12 units, 45 visual
scenarios. The five new PNGs and managed golden manifest were promoted only
through that target. All five promoted images were inspected individually.
No existing golden bytes changed. Reconciliation v3 reports 241 captures,
241 active goldens, zero missing/orphan/ambiguous mappings, and all 28 registered
fixtures resolved. The new row's stable nonregistry scenario is
`scenario_e55e1a363e58`; project `chromium`. No D-VFIX identity was reassigned.

Post-promotion `make generate-drift`, `make generated-artifact-policy-check`
and `make json-shape-check` all PASS at
`.cartulary/test-results/20260913T054059Z-p89171`,
`.cartulary/test-results/20260913T054059Z-p89202`, and
`.cartulary/test-results/20260913T054059Z-p89228` respectively.
`make backend-module-boundary-check` PASS at
`.cartulary/test-results/20260913T053716Z-p84266` (3/3 units).
Initial handoff `make lint-markdown` PASS at
`.cartulary/test-results/20260913T053842Z-p85117`; completion bytes receive a
fresh lint/whitespace/scope check below.

The attempted `make task-guide ROLE=module-author OWNER=harness.visual` was
rejected as an unknown active test owner. That ID is a collaborator label, not
an active test partition. The actual `harness.browser` and `web.design` guides
were used, with the catalog's browser-support architecture/profile rows and
full visual fixture matrix selected. This routing correction changes no owner
inputs or verification claims.


Final browser-support boundary/profile slice:
`make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131,harness.browser.boundary_support.visual_presentation_profiles`
PASS, `.cartulary/test-results/20260913T054231Z-p58919` (3/3 units).
Final scope audit confirms `main` and HEAD remain
`92f26e03bffbead7a03ed0c1cb32f9c8621b18d2`. The worktree contains only this seam's
64 tracked changes and 23 new source/test/handoff/golden files (including the
neutral Timeline hook rename). Digest, lockfiles, db, cmd and configs are
unchanged. `git diff --check` passes. Existing visual golden hashes are unchanged.


### LNC-05 completion

Two fresh ordinary `make browser-e2e-visual` runs after promotion PASS:

- `.cartulary/test-results/20260913T054057Z-p87357`: 12/12 units, 45 scenarios.
- `.cartulary/test-results/20260913T054057Z-p87373`: 12/12 units, 45 scenarios.

Each full run reconciles all 241 active captures/goldens, zero
missing/orphan/ambiguous mappings, and all 28 registered fixtures. They used
independent private browser stacks against the same completed source and
promoted manifest. No source, golden or renderer change occurred between them.

Disposition: DONE. All LNC-01 through LNC-05 binary exits are satisfied. Final
Markdown lint, whitespace and scope checks are repeated over these completed
handoff bytes, including the DONE status. The retained public lint run and
terminal audit are execution evidence; no product artifact reads this Markdown.
There is no unresolved implementation or owner prerequisite. Broad gates and
retained-run maintenance remain explicitly unclaimed/skipped as recorded above.

Rollback restores the Core 01/04 clarification, authored inspector/source-view
and OpenAPI projections, 2.0 candidate ledger, Make-generated derivatives,
Note runtime/presentation/admission changes, ownership/routing, tests and new
visual baselines together. Existing endpoint semantics and committed analyst
Notes, links and revisions must remain intact; rollback requires no data delete
or migration. No commit, push or deployment was performed.

Next action: user review of this completed seam. Do not begin coordination
artifact creation, general relationship work or another refactor automatically.
