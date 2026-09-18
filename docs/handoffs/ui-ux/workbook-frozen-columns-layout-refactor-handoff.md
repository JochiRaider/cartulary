# Workbook frozen columns and saved-layout evolution

## Execution control

Baseline revalidated: clean `main`, HEAD
`5297e5fdc0b611c1e950f42ddf24370c7136551c`, two commits ahead of
`origin/main`. Preserve those commits. No analyst data, digest, dependencies,
lockfiles, commit, push, or deployment changes are authorized. Scope is the
shared Workbook layout seam and its necessary contract/portability consumers.

| Workstream | Status | Exit |
| --- | --- | --- |
| WFC-01 Characterization and adopted contract | DONE | Adopted grammar, transitions, measured budget, migration and gap ledger |
| WFC-02 Canonical layout and portability | DONE | Strict legacy/current boundaries and lossless public/portable round trips |
| WFC-03 Shared controls and geometry | DONE | One layout owner and reachable production regions |
| WFC-04 Integrated evidence | DONE | Applicable interaction, security and compatibility acceptance passes |
| WFC-05 Terminal validation | DONE | Final-source checks and completed handoff |

Only the current row may be IN_PROGRESS. Save actual exit evidence and DONE
before beginning its dependent. Applicable BLOCKED dependencies prevent advance.

## Authority and characterization

Behavior: Core 01 REQ-01-143, saved-view routes, field identity, public
compatibility and Incident Portability; Core 03 REQ-03-026/295 and editing,
query continuity, ranges, Find and retained authoring; Core 04 authorization;
design Columns, viewport, focus and correction-access contracts. Core 02's
saved-layout restatement follows Core 01. Domain owns vocabulary/navigation.
The NLSpec research essay, localized digest and completed sizing, saved-view
authoring/discovery/startup, range-selection/entry, Find, correction-access,
clear-contents and batch-History handoffs are advisory or historical evidence.
Planning followed the digest README's localized read order.

Source placement follows frontend ownership/import manifests and established
backend owners. Verification routing follows contracts/verification, the owner
catalog and authored test families. No executable consumer may depend on Markdown.

Inspected boundaries:

- Workbook layout controller, layout materialization, Columns command/control,
  saved-view models/controller/adapters/observer and startup admission.
- Timeline, Entity, Assessment and Generic share applyWorkbookLayoutToColumns;
  Network Analysis does not and retains its own policy.
- Grid Adapter core, RDG compiler, semantic presentation/focus, pointer geometry,
  column measurement, editor reveal and private structural/group columns.
- Go viewschema layout normalization; Saved Views API/policy/store; Workbook
  startup resource projection; authored OpenAPI and immutable released baseline.
- Incident Bundle archive integrity/version dispatch, source catalog/ports,
  Saved Views preparation/export/apply/validation and recovery contribution.

## Adopted decisions and transitions

One runtime layout uses cartulary.layout.v2 with required nullable
frozen_through_field_key. The non-null boundary names a declared nontechnical
field in the complete column_order. Legacy v1 conversion exists only at
compatibility boundaries. Unknown versions/members/identities fail explicitly.
Existing sparse widths, canonical ordering and bounded additive default-hidden
read-only field evolution remain in force.

The user selected coordinated server/client upgrade on existing routes. Adopt
the narrow breaking-response exception in Core 01 and record the current API
compatibility disposition; immutable released baselines remain unchanged.
Reads normalize in memory, never write. Material writes/imports persist v2;
normalization alone is a no-op. Saved Views has optimistic concurrency and no
idempotent-create receipt contract. Bundle replay hashes and completed receipts
remain unchanged. Newly generated bundles use v4, v3 remains importable under
its original grammar, and bundle v1/v2 remain unsupported.

| Transition | Authored/effective consequence |
| --- | --- |
| Freeze / unfreeze | Set the semantic boundary / null; no record mutation |
| Reorder | Preserve boundary identity and recompute complete prefix |
| Hide / show boundary | Preserve identity; hidden fields render no cells |
| All hidden | Preserve configuration; zero frozen cells; Columns stays reachable |
| Reset Columns | Schema order, default visibility, empty width overrides, null boundary; preserve query/selection |
| Saved-view Reset | Selected saved configuration, or schema default when none is selected |
| Select / startup / replace | Apply admitted saved layout through existing authoring and query continuity owners |
| Create / update / duplicate | Capture working layout / capture working layout / copy selected saved configuration |
| Group / size / Inspector / Recovery / viewport | Recompute geometry; viewport changes never dirty or rewrite layout |
| Authority change | Existing current-authority and protected-content lifetimes; freezing grants no rights |

The minimum scrollable work area is 240 local CSS pixels. With clipped grid
width V, private structural frozen width S and complete visible requested prefix
width P, enable the entire data prefix iff floor(V) - ceil(S + P) >= 240.
Use the same rule for all-visible freezing. Unknown geometry suspends; an empty
visible prefix renders no frozen cells. Inputs do not depend on enabled state.
Never freeze a smaller prefix, shrink widths or change authored state. Resume
without focus theft. Oversized and below-minimum editors retain maximum useful
exposure, native text selection and individually reachable correction actions.

Fresh planning baseline: make service-backed-test-slice OWNER=module.workbook
ROWS=module.workbook.accessibility.grid_autosave passed 10 browser cases / 11
graph units at .cartulary/test-results/20260918T193657Z-p1926802. Actual editor
width was 220px with 2px outline plus 2px offset on each side; correction actions
measured about 116px, 62px and 58px. The 240px budget exceeds the 228px editor
extent. This does not establish frozen-column acceptance or native browser zoom.

## Gap and compatibility ledger

| Gap / classification | Remediation and affected areas | Rationale / durable benefit | Migration / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 New semantic capability | Adopt Core/design grammar and extend existing layout owner | Stable identity and one saved configuration | v1 boundary conversion; v2 needs coordinated readers | Order/visibility/width/freeze round trip exactly |
| G2 Confirmed permissive lower-level decoder; resource admission already guards many cases | Strict frontend and Go compatibility codecs | No malformed input repaired into savable defaults | Keep valid legacy/evolution support; audit every ingress | Unknown/malformed values fail explicitly |
| G3 Confirmed public closed-v1 grammar | Authored input/output schemas, startup projection and reviewed breaking disposition | Public contract matches actual writes/reads | Immutable baseline retained; old clients require upgrade | Generated clients and compatibility checks pass |
| G4 Confirmed bundle v3 and catalog version restrictions | v4 output, immutable v3 decoding, version-disjoint source bindings | Lossless portability with explicit integrity boundary | All source participants admit v4; old readers reject it | v3 import/v4 round trip, tamper rejection, atomicity pass |
| G5 Confirmed correction geometry excludes adjacent frozen headers | Shared region-aware Adapter geometry | Same work area for focus, pointer, sizing and correction | Frozen-region behavior is initially a browser hypothesis | Frozen/scrollable editors and correction actions reachable |
| G6 New viewport suspension | Derived Adapter placement and read-only status; 240px design projection | Authored state survives constrained viewports | Fractional/zoom/pane geometry needs production evidence | Threshold stable; bytes/dirty/drafts/focus unchanged |
| G7 Interaction regression risk | Reuse semantic order, range, editor and retained operation owners | No vendor positions or duplicate DOM semantics | Grouping/virtualization/native selection require browser checks | Correct focus/order/targets and exact request counts |
| G8 Security/continuity regression risk | Preserve current read, local-layout and saved-write authority | Configuration never changes record access | Revocation/account replacement must fence late geometry | No protected-content resurrection or unauthorized writes |
| G9 Confirmed during integrated draft/reorder evidence: vendor EDIT position could open the adjacent field and reclaim Columns focus | Admit editor rendering only for the semantic seed; retained remounts respect the existing external-action focus boundary in Grid Adapter | Prevent wrong-field authoring and preserve local command focus | No layout migration; existing explicit activation remains available for detached presentation | Freeze/unfreeze retain DOM identity; reorder retains exact draft and command focus; zero unrelated record writes |
| G10 Confirmed terminal startup regression from strict decoding | Preserve unavailable-schema reason after Saved Views visibility admission | Existing fallback survives strict canonical reads without exposing malformed layouts | No migration; malformed known-schema layouts still fail explicitly; resolved by startup service fixture | Unknown-schema preference falls back and clears only under existing startup rules |
| G11 Confirmed test-environment incompatibility | Make the jsdom scrollTo shim writable like the platform method | Specialized geometry fixtures can install their own observers | Test-only; no production geometry claim; resolved by focused Adapter tests | Pointer and correction test slices pass together |

Current v1 consumers/dispositions: Go viewschema and Saved Views normalize to
v2; Workbook startup projects normalized v2; frontend models/reads/dirty/operations
use current layout; authored OpenAPI and generated Go/TypeScript migrate through
Make; current API fixtures migrate while explicit legacy fixtures stay v1;
saved_views.row.v1 remains immutable for Bundle v3 and a new row schema serves
v4. Source-port versions and Timeline/Entity version checks admit both bundle
versions without changing record meaning. Released OpenAPI 1.0.0 stays immutable.

## Evidence and next action

Execution preflight: git status --short --branch, git rev-parse HEAD and git log
confirm the clean baseline and both preserved commits. Planning task guides
identified platform.openapi (contract.api was an invalid owner lookup).
WFC-01 exit: owner amendments adopt the grammar, transitions, geometry rule,
compatibility exception and portability policy above. make lint-markdown passed
at .cartulary/test-results/20260918T200122Z-p1966297; git diff --check passed.
No format or geometry policy decision remains deferred to implementation.
Next: begin WFC-02 strict codecs and compatibility projections.

## Recovery and completion policy

After current-format persistence/export, recovery requires a compatible repair
build retaining v2 layouts and v3/v4 bundle codecs. An older binary alone is not
a supported downgrade; never discard freezing to make it readable. Validate
recovery with disposable fixtures. Do not operate on analyst databases.
RESULTS_DIR remains unset: retained-run maintenance is skipped without qualifying
full warm evidence. Applicable acceptance requires PASS; N/A requires an owner
rationale. The execution-control table records current completion; the evidence sections
below record each workstream exit.

## WFC-02 exit evidence

Implemented strict viewschema and Workbook codecs, duplicate-member rejection,
read-time normalization, normalized PATCH comparison, current public responses
and legacy/current inputs. Existing Saved Views concurrency and bundle request
hash construction are unchanged. No receipt rewriting or database conversion.
Bundle v4 export and version-bound v3/v4 saved-row preparation preserve the
original integrity boundary; unchanged families explicitly admit both versions.
The original saved_views.row.v1 schema and released OpenAPI baseline are intact.

Public Make evidence (all roots below .cartulary/test-results):

| Command / selection | Result / run root |
| --- | --- |
| make generate | PASS 20260918T201130Z-p1980642 |
| make test-slice OWNER=platform.viewschema ROWS=platform.viewschema.unit.layout_version_compatibility | PASS 20260918T201211Z-p1983879 |
| Saved Views unit application/no-op and strict preparation slices | PASS 20260918T201211Z-p1983899; versioned preparation PASS 20260918T201501Z-p1996914 |
| Incident Bundles unit catalog and archive/version slices | PASS 20260918T201958Z-p2078036 |
| Workbook saved-view/model/column sizing slices | PASS except adapter projection omission found at 20260918T201444Z-p1995100; fixed adapter rerun PASS 20260918T201619Z-p2016896 |
| Saved-view hook and startup admission slices | PASS 20260918T201927Z-p2076262 and 20260918T201938Z-p2077313 |
| make frontend-typecheck | PASS 20260918T201619Z-p2017080 |
| Saved Views service-backed legacy independent read/no-op | PASS 20260918T201619Z-p2016919 |
| Saved Views service-backed create/patch/export/lifecycle/restore | PASS substantive rows 20260918T201501Z-p1996909; obsolete OpenAPI fixture corrected and create/restore rerun PASS 20260918T201808Z-p2053514 |
| Incident Bundles service-backed export/import/idempotency | PASS 20260918T201747Z-p2036079; imported non-null boundary asserted |
| make openapi-compatibility-check | PASS 20260918T201928Z-p2076469 |

Intermediate failures: generation correctly required five new reviewed API
changes (20260918T200857Z-p1975146); initial manifest ordering was corrected;
old version/schema expectations and one missing test import were corrected.
These failures are resolved, not acceptance evidence. Original v3 integrity
admission, strict legacy preparation and version mismatch rejection pass.
Disposable restore/re-export and rollback fixtures retain a non-null boundary.
Integrated historical-artifact replay and broader owner/security evidence remain
WFC-04 work. Next: extend the existing layout controller and Adapter geometry.


## WFC-03 exit evidence

Extended the existing WorkbookColumnLayoutController, Columns commands and mounted
capability binding. Timeline, Entity, Assessment and Generic project the semantic
visible prefix. Grid Adapter validates it, compiles private sticky flags, measures
CSS tracks independently of active state and publishes only derived status.
viewportGeometry now supplies structural, frozen-data and scrollable regions to
focus/restoration, hit testing, edge scrolling, measurement and correction access.
No second layout store, editor key, duplicate grid or mutation capability was added.
Updated Adapter and Workbook layout source guides.

Public Make evidence (roots below .cartulary/test-results):

| Command / selection | Result / run root |
| --- | --- |
| make generate | PASS 20260918T204011Z-p2200076 |
| make frontend-typecheck | PASS 20260918T203715Z-p2163272 |
| Adapter placement, correction geometry and column sizing test-slice | PASS 20260918T203550Z-p2125700 |
| Workbook frozen layout, sizing and Columns test-slice | PASS 20260918T203555Z-p2126476 |
| Adapter pointer/range policy test-slice | PASS selected rows 20260918T203737Z-p2163882 |
| Adapter production binding test-slice | PASS 20260918T203910Z-p2165134; explicit invalid-prefix extension PASS 20260918T204117Z-p2234590 |
| Workbook frozen layout production browser slice | PASS 20260918T203606Z-p2127261 |
| Existing correction accessibility and frozen layout browser slices | PASS 13/13 graph units 20260918T203911Z-p2165390 |
| Frozen and scrollable correction accessibility slice | PASS two browser cases, 11/11 graph units 20260918T204032Z-p2203093 |

The initial browser assertion incorrectly required an offscreen virtualized
header to remain mounted (203432Z-p2089948); it now asserts absence of all frozen
cells during suspension. Actual frozen placement, non-color boundary, saved v2
bytes, clean dirty state, zero record requests and identical editor DOM across
suspension/resumption pass. Adapter unit failures exposed missing jsdom scrollTo;
the test environment now provides a semantic-only shim. Browser evidence remains
the geometry authority. Initial test fixture type omissions were repaired.
Next: WFC-04 integrated transitions, portability replay and security evidence.


## WFC-04 visual refresh review

Read the visual golden maintenance and browser design readiness guides. Ordinary
make browser-e2e-visual at 20260918T205444Z-p2694182 completed all functional
preconditions and failed only the intended Columns screenshot comparison. Its
reconciliation accounts for all 252 captures and 252 active goldens, with zero
orphans, missing goldens, ambiguous mappings or unresolved registry fixtures.

Accepted trigger: adopted Columns controls and configuration feedback. Affected
owner row: module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc;
scenario_5974e42bb8fb; capture visual.capture.1f0c5e239a5b50226642. This active
capture has no registered fixture ID. Golden:
apps/web/e2e/workbook.visual.spec.ts-snapshots/workbook-view-bar-long-columns-linux.png.
Viewport, CSS/browser zoom, density, masks, scroll normalization and screenshot
scope are unchanged. Reviewed the actual image and the production active/suspended
attachments: new commands and feedback remain inside Columns; the main grid and
view-bar density remain intact. The boundary overlay preserves resize access.
Next: public Make refresh, inspect promoted bytes, then two ordinary validations.


## WFC-04 integrated evidence

Registered focused layout, threshold, correction and legacy replay cases in the
existing owner manifests; generated routing through Make. Existing range, Find,
autosave, saved-view, startup and History fixtures now exercise production frozen
columns. Their semantic targets and request-count assertions remain in place.

All run roots below are under .cartulary/test-results. Each run-manifest.json
records the exact public target, OWNER and ROWS arguments; row results distinguish
passing selections from resolved intermediate failures.

| Public command / selection | Actual result and root |
| --- | --- |
| make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.frozen_columns_layout | PASS 20260918T211250Z-p3098400; placement, cue, saved bytes, draft identity, layout commands, update, suspension and focus |
| make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.frozen_columns_threshold | PASS 20260918T204646Z-p2400076; 239/240/241px both directions, repeated observations, all-visible, hidden boundary, reorder, all-hidden and both resets |
| make service-backed-test-slice OWNER=module.timeline; range/entry/clear selections | PASS 20260918T204517Z-p2334764 and 20260918T204736Z-p2434773; 19 browser cases |
| make service-backed-test-slice OWNER=module.workbook; autosave role/session/availability/successive and correction selections | PASS 20260918T204735Z-p2434354 |
| make service-backed-test-slice OWNER=module.workbook; shared sizing gestures/saved operations and surface matrix | Selected rows PASS 20260918T204602Z-p2368476; Timeline, Entity, Assessment and Generic at three densities |
| make service-backed-test-slice OWNER=module.workbook; Find selections | Eight rows PASS across 20260918T204602Z-p2368476 and 20260918T204833Z-p2499412; corrected virtualized case PASS 20260918T205529Z-p2775868 |
| make service-backed-test-slice OWNER=module.workbook; frozen correction and startup preferences | Selected rows PASS 20260918T205340Z-p2592662; three frozen/scrollable/CSS-zoom correction cases and restored saved boundary |
| make service-backed-test-slice OWNER=module.timeline; batch History review | PASS 20260918T205341Z-p2592911 |
| Network Analysis navigation/accessibility service-backed slices | PASS 20260918T205343Z-p2593463; no policy changes |
| make service-backed-test-slice OWNER=module.workbook; baseline correction, Columns accessibility, viewport and frozen layout | PASS 20260918T205445Z-p2694309; text spacing, below-minimum correction and Inspector geometry |
| make service-backed-test-slice OWNER=module.incidentbundles ROWS=module.incidentbundles.integration.layout_legacy_conversion_replay | PASS 20260918T205035Z-p2539437 |
| Incident Bundle queued export recovery service-backed slice | PASS 20260918T205457Z-p2741050; new artifact is v4 |
| make service-backed-test-slice OWNER=module.timeline; four measurement rows | PASS 20260918T205654Z-p2811945; 20/20 graph units |
| Workbook strict admission/model/controller test-slice | PASS 20260918T205730Z-p2841973 |
| Adapter production binding test-slice after semantic editor admission fix | PASS 20260918T210705Z-p3028736 |
| Saved-view current-response browser rows | PASS 20260918T210043Z-p2925566; two obsolete v1 assertions repaired |
| make frontend-typecheck | PASS 20260918T210804Z-p3086793 |
| make frontend-import-boundary-check | PASS 20260918T205718Z-p2841535 |
| make lint-markdown | PASS 20260918T210805Z-p3088523 |

Compatibility fixtures verify original v3 archive bytes and integrity before
conversion, original row grammar, atomic rejection, current v2 persistence and
historical successful export/import replay without additional durable requests.
They preserve original job/receipt/artifact/hash values. The historical v3
fixture originates from an unfrozen legacy layout; it is not a downgrade codec.
The separate v4 recovery/restore fixtures retain non-null freezing.

Resolved findings: frozen headers remain mounted, so Find virtualization now
asserts the genuinely scrollable destination; reverse traversal follows the
fixture's authored order. A real boundary border initially obscured the vendor
resize edge; the non-interactive double-rule overlay preserves hit testing.
Draft/reorder evidence exposed G9: stale vendor EDIT position could open the
adjacent field. Semantic seed admission now prevents that, and a retained editor
respects external-action focus. The final browser case proves zero record writes,
exact draft recovery and Columns focus. Initial assertions expecting a display
cell while that retained editor was mounted were corrected. Failures at
20260918T205841Z-p2881710, 20260918T210215Z-p2995843 and
20260918T210706Z-p3029008 are superseded by the passing focused run above.

| Applicable acceptance group | Disposition |
| --- | --- |
| Strict null/one/several/all-visible/hidden/evolution/order/width contracts | PASS: strict codecs, controller tests and production threshold/sizing matrix |
| Every saved-view operation, startup, normalized no-op, concurrency and system reads | PASS: Saved Views service, controller and browser evidence; duplicate copies selected saved configuration |
| Legacy/current public and portable formats, integrity, replay, rollback and compatible recovery | PASS: WFC-02 service evidence and WFC-04 legacy/replay fixtures |
| Production virtualization, grouping, offscreen reveal, sizing, density, spacing and panels | PASS: range, Find, sizing, correction and measurement selections |
| Deterministic suspension and resumption without draft/configuration/focus changes | PASS: exact threshold and retained-editor production cases |
| Keyboard/pointer ranges, Enter/Tab/F2, native selection, clipboard/fill/clear/Find, trailing creation and History | PASS: production interaction selections with frozen fixtures |
| Pending/rejected/conflicting edits, delayed navigation, replacement and query admission | PASS: retained-authoring, range settlement, Find lifecycle and saved-view cases |
| Read-only interaction, write authority, revocation, suspension and account replacement | PASS: existing authority lifetimes exercised with frozen production fixtures |
| Shared consumers and unchanged Network Analysis navigation | PASS: surface matrix and dedicated Network Analysis evidence |
| Native browser zoom qualification | N/A: this seam's required fixtures use CSS zoom; no native browser-zoom claim is made |

G1–G9 have passing binary criteria and no unresolved implementation blocker.
Remaining work: complete the second ordinary golden validation, then terminal
checks against final source. Broader release/deployment qualification is outside
this implementation seam.


WFC-04 exit: make browser-e2e-visual-update passed at
20260918T210042Z-p2925377. Inspected the promoted Columns image; only that golden
and its generated manifest entry changed. Two fresh ordinary
make browser-e2e-visual validations passed, 12/12 units each, at
20260918T210751Z-p3061216 and 20260918T211300Z-p3117793. The explicit additive-field
fixture preserves its non-null boundary; make test-slice OWNER=platform.viewschema
passed at 20260918T211412Z-p3171102. All applicable integrated exits now pass.
Next: WFC-05 finalization and final-source verification.


## WFC-05 progress and terminal dispositions

Task guides verified web.workbook, package.grid_adapter, platform.viewschema,
module.savedviews, module.incidentbundles, module.timeline, module.workbook,
platform.openapi, package.protocol_ts, package.ui, web.design and
harness.generated_artifacts, plus every changed bundle source owner: artifacts,
assessments, entities, evidence, incidents, indicators, links, parties, records,
revisions and tasksdecisions. web.testing is a source-ownership ID, not an active
test owner; its shim is covered through Adapter/frontend evidence. The initial
harness.frontend and web.testing task-guide lookups were invalid and confer no
verification. The correct API owner is platform.openapi.

make agent-finalize passed at 20260918T211748Z-p3174470 before broader terminal
verification. RESULTS_DIR was unset; retained-run maintenance was skipped.
make generate-drift passed at 20260918T211815Z-p3178458; public compatibility,
generated policy and JSON shape passed at 20260918T211824Z-p3185606,
20260918T211824Z-p3185586 and 20260918T211824Z-p3185727. Frontend/backend import
boundaries passed at 20260918T211824Z-p3187058 and
20260918T211824Z-p3186668. Script lint passed at 20260918T211824Z-p3187062.

Terminal failures and remediation remain explicit:

- lint-biome at 20260918T211824Z-p3186950 rejected 15 non-null assertions;
  explicit fixture checks replaced them. Rerun PASS 20260918T212104Z-p3305288.
- Startup slice at 20260918T211844Z-p3204400 exposed G10. The existing
  unavailable-schema fallback now runs after Saved Views visibility admission;
  it does not admit an invalid layout. Focused startup rerun PASS
  20260918T212204Z-p3371089.
- backend-unit at 20260918T211815Z-p3178509 passed 149/150 units; the remaining
  Incident codec test still called bundle v4 unsupported. Its negative fixture
  now uses v5 and separately asserts v4 context binding.
- frontend-unit at 20260918T211815Z-p3178657 passed 651/653 units. Both failures
  came from the new test shim's non-writable property descriptor. Writable
  scrollTo fixes the specialized pointer/reveal fixtures; focused rerun PASS
  20260918T212346Z-p3505495. Production code did not change for this repair.
- Indicator portability at 20260918T212240Z-p3403074 retained an obsolete
  descriptor expectation [3]; it now asserts the adopted [3,4] admission.
- The broad concurrent Timeline browser slice at 20260918T212104Z-p3305156
  passed all selected interaction cases except timed edge scrolling. It reached
  638px within the unchanged 5s assertion while other heavy suites ran. Keep
  the acceptance threshold unchanged and rerun with heavy work quiescent;
  this failure is not treated as acceptance evidence.

Next: finish final-source reruns, isolated scrolling/measurement, remaining
maintenance and final scope review before marking WFC-05 DONE.


Final review of G9 also keeps the ordinary semantic cell content visible when a
vendor edit position is no longer admitted for that field. The shared cell
renderer supplies the fallback; no draft is moved and no second cell is created.
Focused Adapter placement, binding, pointer and reveal checks pass at
20260918T212916Z-p3668068. Frontend type checking passes at
20260918T212916Z-p3668262. The browser case additionally checks that the displaced
Date Entered cell remains rendered during Synopsis reordering.

| Terminal public command / owner selection | Result / run root |
| --- | --- |
| make backend-unit | PASS 150/150 units, 20260918T212453Z-p3549557 |
| make frontend-unit | PASS 653/653 units, 20260918T212453Z-p3549637; subsequent small renderer fallback covered by focused Adapter/browser/type checks |
| make service-backed-test-slice OWNER=module.incidentbundles | PASS 6/6 units, 20260918T211815Z-p3178532 |
| Saved Views service-backed Go rows (create, update, storage, read, export, restore) | PASS 7/7 units, 20260918T211843Z-p3204255 |
| Workbook startup fallback rerun | PASS 20260918T212204Z-p3371089; other preference rows passed in 20260918T211844Z-p3204400 |
| Entity source invariants | PASS 20260918T212240Z-p3403071 |
| Indicator source determinism and invariants rerun | PASS 20260918T212524Z-p3567874 |
| Party source apply and closed failures | PASS 20260918T212240Z-p3403132 |
| Records source envelope and rollback | PASS 20260918T212240Z-p3403036 |
| Revisions portability allocator, atomicity and attribution | PASS 20260918T212418Z-p3511711 |
| Tasks/Decisions portability invariants and atomicity | PASS 20260918T212418Z-p3511724 |
| Workbook production browser and accessibility matrix | PASS 20/20 units, 20260918T212104Z-p3305138 |
| Timeline range/entry/clear and all four batch History operations | All selected rows PASS in 20260918T212104Z-p3305156 except scrolling, which requires the isolated rerun recorded below |
| make lint-markdown | PASS 20260918T212712Z-p3623557; completed handoff will be linted again |

All six separately selected source portability owners retain unchanged record
schemas while admitting v4. Other changed source owners are covered by the full
backend unit suite and Incident Bundle extension/atomic-import matrix. No new
SQL migration is needed or run against analyst databases.


## Final scope, rollout and recovery

Substantive authored paths are the adopted Core 01/02/03/04 and design amendments;
internal/platform/viewschema/layout.go; Saved Views policy, store and portability;
Incident Bundles version dispatch and sourceport catalog; the source-owner bundle
admission declarations; authored OpenAPI and current compatibility dispositions;
apps/web/src/workbook/layout, models, saved-view/startup adapters and Columns;
and packages/grid-adapter/src compiler, measurement, viewport, pointer and editor
geometry. New files are the frozen-placement hook, saved_views.row.v2 schema and
this handoff. Tests and their authored owner manifests follow those same owners.
Generated OpenAPI/Go/TypeScript/design/routing derivatives were produced through
Make. The one reviewed Columns golden and its generated manifest are intentional.

No file path was retired. The permissive frontend saved-layout decoder behavior
was retired in favor of explicit failure. Layout v1 has no parallel runtime
controller: it survives only at validated compatibility boundaries. Original
saved_views.row.v1 and the released OpenAPI 1.0.0 baseline remain unchanged.
The digest, research document, domain vocabulary, dependency pins/lockfiles and
SQL migrations remain unchanged. No analyst data, browser persistence, record
capabilities, row-height policy or Network Analysis navigation policy changed.

Rollout is a coordinated server/client upgrade on the existing routes. Drain or
reload older clients before allowing writes against the upgraded deployment;
current responses use layout.v2 and newly produced exports use bundle v4.
Existing supported layouts normalize in memory without a bulk data conversion.
A completed historical export remains its original artifact, and unfinished
queued exports produce v4.

Recovery after any v2 persistence or v4 export requires a compatible repair build
retaining the current layout reader/writer and both admitted bundle readers.
Restoring a disposable v4 fixture and re-exporting preserves its non-null freeze
boundary; legacy v3 imports preserve their original-byte verification before
conversion. The saved-view restore, atomic rollback, historical replay and v4
round-trip fixtures cited above validate this recovery boundary. An older binary
alone is not a supported downgrade. No tool or procedure silently removes
freezing or rewrites an immutable archive to make that downgrade succeed.

Skipped/outside scope: retained-run maintenance (RESULTS_DIR unset); native browser
zoom qualification (CSS zoom is explicitly limited evidence); SQL migration drift
(no authored SQL changes); full release/check/conformance publication and
production deployment (this is bounded implementation evidence, not a release
claim). Applicable owner slices, full backend/frontend units, production browser,
accessibility, visual, measurement, security and maintenance checks are the
completion gates. No runtime, test, generator, conformance or release artifact depends on
Markdown; owner-document alignment is a human review responsibility.


Latest final-source confirmations: frozen layout, threshold and three correction
browser cases PASS 13/13 units at 20260918T212916Z-p3668095; generation drift PASS
20260918T213043Z-p3740421; Biome lint PASS 20260918T213043Z-p3740642; targeted Go
security PASS 20260918T213043Z-p3740756. G10 and G11 are resolved. Remaining
terminal scrolling, timing and visual results are recorded in the completion
entry below.

Final scope review reconfirms main at
5297e5fdc0b611c1e950f42ddf24370c7136551c, still two commits ahead of origin/main.
Both original commits remain intact. git diff --check passes; no deleted paths,
immutable baseline edits, dependency/lockfile edits, digest/research edits or
analyst-data changes are present. This work remains uncommitted for review.


Final ordinary make browser-e2e-visual PASS 12/12 units at
20260918T212916Z-p3668386, covering the final renderer fallback. Final generated
artifact policy PASS at 20260918T213315Z-p3772926. With other heavy suites stopped,
make service-backed-test-slice OWNER=module.timeline
ROWS=module.timeline.browser.range_scrolling passed unchanged, 11/11 units, at
20260918T213404Z-p3773804. The earlier concurrent timing failure is resolved;
no timeout, threshold, frame-rate limit or scrolling behavior was relaxed.


## WFC-05 completion evidence

The final four Timeline measurement rows passed through
make service-backed-test-slice OWNER=module.timeline with the exact ROWS recorded
in 20260918T213518Z-p3805330/run-manifest.json: typing acknowledgment, blank-row
creation, ArrowDown selection and Enter focus. All 20/20 graph units passed.
Production DOM geometry, focus/order, submitted request counts and saved bytes
are asserted in the frozen-column browser cases; application fakes are only
supplemental evidence.

All applicable acceptance groups above are PASS. All G1–G11 binary validations
pass; there is no applicable BLOCKED dependency or unresolved implementation gap.
The final documentation review aligned requirement backlinks and clarified that
the retired-codec prohibition concerns bundle v1/v2, preserving the adopted
v3/v4 policy. These are owner-document reconciliations, not runtime Markdown
inputs or new product behavior.

The supported limitation is coordinated rollout and compatible repair recovery;
this change does not qualify native browser zoom or publish a release/conformance
claim. The only observed timing sensitivity was the concurrent edge-scrolling
run, superseded by the unchanged isolated pass. No acceptance criterion was
weakened to obtain passing evidence.

Next action: review this uncommitted seam and use the coordinated upgrade policy
for any later rollout. Implementation stops here. No commit, push, deployment,
analyst-data modification or unrelated follow-on work was performed.


WFC-05 DONE: completed owner documents and handoff passed make lint-markdown at
20260918T214044Z-p3846306. Final git diff --check and branch/HEAD/scope review
passed. Every required implementation gate has passing evidence above; retained-run
maintenance remains explicitly skipped because RESULTS_DIR was unset.
