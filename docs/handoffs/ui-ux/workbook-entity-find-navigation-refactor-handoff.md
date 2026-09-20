# Hosts and Identities Find and match navigation

## Execution control

Execution starts on clean `main`, HEAD
`601aeb6c6f04923ffedd5d0cad979d8fd018c28d`. Root AGENTS.md applies. The user
authorized the complete implementation plan and bounded Core 03/design adoption.
No commits, pushes, deployment, analyst-data changes, dependencies, endpoints,
durable search, additional surfaces, Replace or automatic page scanning belong here.

| Workstream | Status | Binary exit |
| --- | --- | --- |
| EFN-01 Characterization and adopted contract | DONE | Both production renderers characterized; owner amendments adopted; Markdown passes. |
| EFN-02 Source integration and shared boundary | DONE | Focused text, integration, Timeline, type and boundary evidence passes; superseded path removed. |
| EFN-03 Interaction, continuity and authority | DONE | Required production browser and affected regressions pass. |
| EFN-04 Final verification and handoff | DONE | Finalizer, terminal checks, reviewed visuals and acceptance complete. |

Record actual evidence and save DONE before starting each dependent. Applicable
BLOCKED evidence prevents completion. Historical handoffs and the September 13
digest localization are advisory, never current product passes.

## Authority and baseline

Planning inspected AGENTS.md and the digest README, START_HERE, LOCAL_AGENT_PROMPT,
REPO_MAP, OWNER_MAP, rules, acceptance and QUERY_RECIPES. Core 01 owns query,
identity, field and mutation contracts; Core 03 §§2.3A,3–4,13.5,14.9,16 and
REQ-03-298/299/100 own interaction/retention; Core 04 §§1–2 own authorization.
Design §§7–8,10,14 own presentation. Domain §13.11 supplies vocabulary/navigation.
The NLSpec research essay is methodological context, not another request.

Current source ownership is `web.workbook`; verification is independently routed
through `web.workbook`, `module.workbook`, `package.grid_adapter` and affected
`module.timeline`, architecture, design and UI rows. Authored source/import,
catalog, verification and topology manifests were inspected. Generated policy
also includes `packages/view-contracts/src/generated`; generated roots and
topology are updated only through Make. RDG imports remain Adapter-private.
React 19.2.5 and RDG 7.0.0-beta.59 are the current authored versions.

Later completed query-continuation, grid-edit/autosave, recovery-navigation,
column-sizing/frozen-layout, Timeline Find, collection-input and autosave-feedback
handoffs were reconciled with current source. Relevant localization drift:
independent Entity browsers, accepted-window membership, independent Timeline
grid/Inspector collection revisions, interrupted continuity restoration,
read authorization independent of replay, and query failures independent of save
feedback are already implemented. They must survive this slice.

Fresh planning evidence under `.cartulary/test-results/`:

- Find matching/control/Timeline text/integration: 5/5 at `20260920T034228Z-p42911`.
- Adapter semantic navigation/focus: 3/3 at `20260920T034343Z-p44632`.
- Import boundaries: 2/2 at `20260920T034343Z-p44718`.
- Make help/help-all and task guides for the four primary owners passed.
- Prescribed offline UX focus/accessibility and React virtualized-state queries
  used `PYTHONDONTWRITEBYTECODE=1 python3 -B`; both returned eight results, no fallback.

## Selection rubric and decisions

The following ledger covers observed weakness, remediation/change areas, owner,
rationale/lasting benefit, extension path, capability value, retirement,
compatibility, risk and binary validation for every material gap.

| Gap and classification | Remediation / owner / benefit | Retention, retirement and extension | Compatibility / risk / binary validation |
| --- | --- | --- | --- |
| G1 Capability extension: Entity Find absent | Adopt Core 03 §13.5 and bounded design amendments before integration; explicit loaded scope avoids incident-search ambiguity. | Retain existing literal matching and cell counts for predictable navigation. Another surface requires adoption and can supply the same capabilities; do not enable it now. | No wire/data change. Leaving scope implicit risks misleading absence claims. Both entity surfaces expose the adopted interaction and no other surface gains it. |
| G2 Structural weakness: Timeline hook combines common lifecycle and source settlement | Extract one Workbook binding for authority, accepted/presented intersection, cancellation, focus and panel lifetime. Source adapters retain text and editor semantics. | Retain controller/control and Timeline registry/save callbacks; remove superseded Timeline lifecycle implementation and migrate callers. Future sources supply capabilities, not branches or a registry. | Internal TS cutover only. Duplication risks inconsistent retirement and races. Shared and Timeline regressions pass with one Find owner. |
| G3 New presentation seam: entity fallbacks/collections differ from Timeline | Share entity committed presentation decisions with the renderer; safe independent fragments exclude metadata/placeholders. | Retain displayed fallback precedence, aliases and readable identifier values. Retire inline entity formatting only after renderer/text consumers migrate. No duplicate row or search store. | No record/layout migration. Raw JSON, internal fallbacks or draft scanning would disclose false matches. Exact renderer-derived counts/exclusions/order pass for both schemas. |
| G4 Confirmed coupling: active-cell callback changes Inspector subject | Fence Entity Find-driven focus/restoration from Inspector retargeting while preserving ordinary selection. Existing Adapter editor gate owns scalar departure. | Retain independent Inspector/collection drafts and explicit submission, as selected by user; retain source focus handles without copying text. No new mutation/settlement owner. | No new authoring policy. Coupling risks hidden draft detachment and unwanted reads. Opening/typing/Close cause no writes; rejection retains raw work; admitted movement preserves Inspector subject. |
| G5 Browser evidence gap and lifecycle hypotheses | Production fixtures cover two-axis virtualization, pending navigation, query/layout changes, authority and geometry; current owner routing. | Retain meaningful semantic assertions, density/focus tokens and current visual harness. No speculative index or performance framework. | No conformance claim inferred from fixtures. Late focus/protected state/duplicate writes are concrete risks to test. Required browser cases and affected regressions pass, with actual artifacts. |

EFN-03 exposed additional material gaps in existing owners (conformance repairs,
separate from the authorized Find capability extension):

| Gap and classification | Remediation / owner / benefit | Retention, retirement and extension | Compatibility / risk / binary validation |
| --- | --- | --- | --- |
| G6 Recovery attachment race: a confirmed viewer role is repeatedly invalidated after a cancelled source read | Collaboration retains the confirmed authorization plan and retries the owed accepted read; publish recovered role after retiring old write admission. | Retain the source query guard, recovery scheduler and mutation pause. Reuse the existing delay; remove the cancelled-read full-authorization loop. Other surfaces use the same accepted-read capability. | No authority bypass or wire change. Premature read admission/replay is the risk. Viewer reauthentication on both entities and coordinator cancellation/late-outcome regressions must pass. |
| G7 Adapter callback assumes the selected column survives configuration replacement | Adapter treats a missing vendor column as a retired semantic anchor. | Retain vendor-private translation and semantic selection; retire the unsafe dereference. The same guard handles any supported column replacement. | No new navigation policy. A saved view hiding the selected column previously threw and interrupted the shell. Production saved-view replacement and Adapter regressions must pass. |

The failed-refresh/newer-input investigation additionally tests G4/G5: input must
reach the source editor before Find publishes outside-context collapse. No draft
text or mutation state is transferred into Find. Deletion delivery exposed G8 below; completion remains pending.

| Gap and classification | Remediation / owner / benefit | Retention, retirement and extension | Compatibility / risk / binary validation |
| --- | --- | --- | --- |
| G8 Existing publication defect: empty public field keys become JSON null | Collaboration publication preserves an empty array, as Core 01 REQ-01-267 and the authored WS schema already require. | Retain strict browser decoding and source-owned removal refresh; retire nil-slice cloning for this wire field. The fix also covers restore/invalidate events without cell deltas. | No schema amendment or migration. Invalid events silently leave stale membership. Exact serialized array regression and live deletion on both Entity renderers must pass. |

Material advisory classification: R001–R004/R006–R010/R013–R015 ADOPT for keyboard,
focus, non-color meaning, local feedback, virtualization and owner-scoped recovery;
R012 ADAPT for state placement through existing lifetimes; R017–R019/R035 ADAPT
through desktop geometry and displayed fragments. R029/R033/R034 REJECT generated
design authority, generic behavior prescriptions and incidental test selectors.

Entity Inspector scalar/alias drafts retain explicit submission. Find only settles
an active Grid Adapter editor; Timeline retains its existing source settlement.
Readable stale rows remain searchable; authority loss clears Find independently
of retained work. Same-schema configuration changes recompute; schema departure
retires Find. No separate owner contradiction is identified.

## Evidence and acceptance

Execution evidence, final A001–A027 assessment and exact changed-file inventory
are recorded below as work completes. No acceptance is implied by planning passes.

## Compatibility, rollback and next action

No endpoints, data/dependency or saved-layout migration. Roll back this slice's
coordinated owner, authored source/routing, generated and reviewed golden changes;
preserve unrelated work and accepted writes. Current next action: review the completed uncommitted slice and handoff.

### EFN-01 evidence

- `make generate`: PASS at `20260920T034905Z-p46709`; only authored browser
  routing generated topology changes. `make format`: PASS at `20260920T034937Z-p49848`.
- `make frontend-typecheck`: PASS 2/2 at `20260920T034944Z-p54376`.
- `make lint-markdown`: PASS at `20260920T034944Z-p54396`.
- First characterization run failed at `20260920T034954Z-p56183` because the
  fixture attempted an unmounted horizontal actions-column button. Inspected
  `make explain-run` and the Playwright report. Use the existing visible Inspector
  toolbar entry after selecting the original cell; no product assertion changed.
- Second characterization failed at `20260920T035230Z-p89264`: reused fixture
  SAM account name resolved multiple Identity creates to one record/version.
  Made that declared identifier unique per fixture record; captured observation
  text while its virtualized cell is mounted. No product change was required.

EFN-01 exit: production characterization passed 11/11 execution units at
`20260920T035355Z-p21948`. The group report retains
`entity-presentation-observations`: each schema renders committed names, aliases
and its current primary fallback; empty reusable identifiers display `None`.
Raw grid work survives focus borrowing, Escape cancels locally, and Inspector
alias work remains explicitly submitted. Core 03 §13.5/design scope and fragment
rules are adopted; domain changes only navigation. No owner contradiction remains.
Next action: EFN-02 shared Find binding and entity presentation integration.

### EFN-02 evidence

The common `find/useWorkbookFind.ts` boundary now owns attachment, accepted /
presented membership, cooperative scan freshness, synchronous authority retirement,
destination cancellation, keyboard/panel lifetime and semantic focus restoration.
Timeline retains text, registry, revision-specific source settlement and continuity
interruption in `useTimelineFindSource.ts`; its old 462-line lifecycle hook is
removed and composition/tests migrated with no compatibility alias.

Entity renderers and Find share `entityCellPresentation`; generic collection label
precedence is retained in one presentation decision. Readable fragments exclude
technical fallback IDs, arbitrary JSON and placeholders. Entity query browsers,
rows, mutation/recovery owners and explicitly submitted inspector drafts remain
independent. Find-admitted focus suppresses only inspector retargeting; ordinary
selection and bulk context retain their existing owners. Control IDs now use
React instance identity. The generated help copy describes displayed collection
items accurately. No wire, dependency, persistence or data migration occurs.

- Initial focused/type/format routing failed before execution: new selector titles
  were not ASCII-sorted. Corrected the authored rows, then generated normally.
- `make generate`: PASS `20260920T040433Z-p62638` (design derivative and routing).
- Focused six-row run: 6/7 at `20260920T040443Z-p65722`; new binding fixture
  omitted required overrides in `fullWorkbookViewRow`. Same fixture error caused
  typecheck 1/2 at `20260920T040443Z-p65836`. Inspected retained diagnostics and
  supplied the missing fixture argument; no product assertion was relaxed.
- Focused entity text, shared binding and Timeline integration: PASS 4/4 at
  `20260920T040526Z-p72418`; controller/control/Timeline text passed in the prior
  six-row run. Import boundaries PASS 2/2 at `20260920T040443Z-p65862`.
- Typecheck PASS 2/2 at `20260920T040528Z-p72704`; format PASS 2/2 at
  `20260920T040509Z-p68007`.
- Existing Timeline surface-lifetime browser expectation now checks Hosts has a
  fresh empty Find, reflecting the adopted capability rather than absence.

EFN-02 exit recorded complete before EFN-03 production interaction additions.
Next action: both-surface production focus, editing, query and authority proof.

### EFN-03 evidence and repair decisions

All roots below are under `.cartulary/test-results/`. Execution-unit totals
include prerequisites and summaries; scenario counts are read from each group's
`playwright-report.json`, not inferred from those totals.

- Real navigation originally cancelled during React imperative-handle replacement.
  The shared binding now checks the captured stable Adapter presentation port,
  including during the transient null ref. The binding regression simulates that
  commit boundary; production tests establish physical focus and virtualization.
- Saved views hiding a selected column exposed the Adapter missing-column callback
  (G7). Viewer recovery exposed the cancelled-read authorization loop (G6).
  Live deletion exposed `changed_field_keys: null` from empty-slice cloning (G8).
  These repairs stay with their existing owners; Find has no workaround/read path.
- Timeline's superseded lifecycle hook is removed. The now-unused exported
  `collectionItemLabels` helper is also retired after its rendering caller and
  test migrate to the shared generic presentation decision. No compatibility alias
  or duplicate text/search state remains.
- Browser input simulation must first return focus to a borrowed editor before
  entering newer work. The failed-refresh case now does so, preserves every raw
  value/focus assertion, and checks one accepted write followed only by reads.
  Find's outside-input listener runs after the editor's React handler while still
  synchronously cancelling a pending destination in that event.

| Command / selected boundary | Result / retained root |
| --- | --- |
| Entity initial focus/browser proof | FAIL `20260920T041122Z-p86061`, then broader discovery FAIL `20260920T041612Z-p31795`; stable presentation-port fix and fixture corrections described above. |
| All nine existing Timeline Find scenarios | PASS 11/11 units `20260920T041613Z-p32226`; later combined current-source run below. |
| `make service-backed-test-slice OWNER=module.timeline` collection-input authoring/characterization/lifetime/promotion and autosave feedback characterization/query-overlap | PASS 13/13 units `20260920T041755Z-p97742`. |
| `make service-backed-test-slice OWNER=module.workbook` grid autosave refresh/role/session/surface-matrix/exact-replay/availability and query continuation Hosts/Identities/authority/live-recovery/session/Timeline | PASS 13/13 units `20260920T041756Z-p97961`. |
| `make test-slice OWNER=web.workbook` committed-grid-autosave, query browsing/controls, recovery navigation/model, column sizing and frozen layout | PASS 8/8 `20260920T041757Z-p98185`. |
| Entity focused browser investigations | FAIL `20260920T042333Z-p68564`, `20260920T042426Z-p932`, `20260920T042724Z-p35281`, `20260920T043159Z-p69109`, `20260920T043332Z-p2086`, `20260920T044010Z-p36479`, `20260920T044236Z-p70373`. Group reports and traces retain exact failing assertions. Last run passes newer-work/read-only recovery and proves deletion's invalid null array. |
| Corrected live deletion and saved-view replacement, both schemas | PASS 11/11 units `20260920T044407Z-p3900`. |
| `make test-slice OWNER=module.collaboration ROWS=module.collaboration.support_unit.semantic_hub_and_payload_ownership_d1e2f3a4b5` | PASS 1/1 `20260920T044406Z-p3643`; serialized empty arrays for remove/invalidate, nil/empty owner inputs. |
| `make test-slice OWNER=web.collaboration ROWS=web.collaboration.regression.workbookcollaborationcoordinator_suite_514be174a8` | PASS 2/2 `20260920T044112Z-p69007`; deferred viewer read confirmation, no repeated role invalidation or replay. |
| `make test-slice OWNER=package.grid_adapter` | PASS 56/56 `20260920T044530Z-p43241`; all routed local Adapter rows including semantic focus/navigation and editor retention. |
| `make test-slice OWNER=web.workbook` entity text, shared binding, control, matching, generic labels, Timeline binding/text | PASS 8/8 `20260920T044557Z-p7948`. |
| `make frontend-typecheck` | FAIL 1/2 `20260920T044507Z-p42440`: redundant cancellation after task reset narrowed to null; removed it. PASS 2/2 `20260920T044558Z-p8236`. |
| `make format` | PASS 2/2 `20260920T044454Z-p37635`. |

Earlier routing-only failures used singular scenario IDs/dynamic test titles,
then an incorrect `.regression` browser row prefix. Corrected authored selectors
and Make arguments before execution; no assertions or tolerances changed.
Diagnostics were removed from production source after localization. Find never
persists or logs search terms. Browser fixtures use isolated test incidents and
local visual preferences; no analyst or shared account presentation data changed.

EFN-03 final evidence:

- Combined Entity/Timeline run `20260920T044528Z-p43009`: all nine Timeline
  scenarios and ten of eleven Entity scenarios PASS. The added creation check
  incorrectly assumed an Entity creation pin is rendered. Current query source
  deliberately admits receipts only for existing members; unlike Timeline, no
  Entity out-of-query pin exists. Corrected that characterization assertion to
  require an accepted receipt, absent out-of-query row and zero Find matches,
  followed by one matching cell after accepted filter removal. Shared binding
  tests additionally exclude synthetic presentation-only pins.
- Window/creation exclusion and both viewer suspension/reauthentication plus
  incident-membership revocation scenarios PASS, 11/11 execution units at
  `20260920T044936Z-p22072`. The revocation scenarios verify incident departure
  and protected search retirement independently of account-session recovery.
- Both-surface exact primary renderer search (including Identity email fallback
  in UPN and Email cells), paste/origin and paste recovery, frozen columns and
  column-layout surface regressions PASS, 15/15 units at `20260920T045129Z-p93689`.
- Current-source grid autosave refresh/role/session/surface-matrix/exact-replay/
  availability and query-continuation Hosts/Identities/authority/live-recovery/
  session/Timeline rerun PASS, 13/13 units at `20260920T045007Z-p56216`.
- Inspected retained 768px/text-spacing Find and frozen/offscreen Identity focus
  images; labelled controls, scope feedback, focus and non-color cues are visible.
  Extracted review copies live in `.cartulary/entity-find-review/`; originals are
  embedded attachments in the Entity group report at `20260920T044528Z-p43009`.
- `git diff --check` PASS; no superseded hook caller or collection helper remains.
  The retained Timeline test filename is an existing verification route, with its
  caller migrated to the source adapter. Design §14 explicitly names the already
  adopted surfaces for its existing Find announcement rules.

EFN-03 binary exit is PASS and marked DONE before EFN-04 starts. Every required
interaction is backed by current production browser evidence; unit fixtures are
supplemental. The three owner repairs have regression evidence and no unresolved
owner contradiction. Next action: run `make agent-finalize`, with `RESULTS_DIR`
unset because no qualifying successful full warm-check root exists, then the
selected terminal checks and visual workflow.

### EFN-04 terminal verification

`make agent-finalize` PASS 1/1 at `20260920T045334Z-p27990`, before all terminal
commands below. Its `unit-artifacts/finalize-summary.json` records retained-run
maintenance **SKIPPED because RESULTS_DIR was unset** (`results-dir-not-provided`).
No qualifying full warm-check root was supplied and no retained timing baseline
was refreshed. Generated structure maintenance introduced no unrelated diff.

Terminal verification and corrective checks are recorded below. The ordinary
visual run passed without a golden mutation. This evidence is implementation
support, not Core 05 publication or a claim of complete Base Profile conformance.

Terminal verification found G9: the new browser test used a raw editor test-ID
template. The architecture selector gate correctly rejected it. A shared
`workbookGridEditorTestId` now owns the existing exact record/field identity in
`package.ui`, with both production editor variants and the new browser consumer
migrated. This preserves selector bytes and existing supported consumers, adds no
product behavior or compatibility alias, and removes the new duplicated template.
The lasting benefit is one validated cross-boundary selector construction; future
ordinary-editor tests can use it without local templates. Leaving the duplication
would violate the authored selector boundary. Binary exit: architecture selector,
UI package and production editing/Find checks pass. Existing unrelated test
selector migrations are outside this slice. This is an EFN-04 verification repair,
not a reopened capability decision; the previously passed interaction assertions
are retained.

Terminal results recorded so far:

| Public command | Result / root |
| --- | --- |
| `make generated-artifact-policy-check` | PASS 3/3 `20260920T045404Z-p31949`. |
| `make generate-drift` | PASS 4/4 `20260920T045404Z-p31941`. |
| `make json-shape-check` | PASS 3/3 `20260920T045404Z-p31956`. |
| `make backend-module-boundary-check` | PASS 3/3 `20260920T045404Z-p32196`. |
| `make frontend-unit` | 654/655 PASS, target FAIL `20260920T045404Z-p32119`; sole failure is G9's new raw editor selector. No other frontend row failed. |
| Corrective `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.selectorcontractpolicy_keeps_cross_boundary_sele_61868c7039` | PASS 2/2 `20260920T045808Z-p2620`; same gate, unchanged assertions. |
| `make test-slice OWNER=package.ui` after selector cutover | PASS 10/10 `20260920T045930Z-p19274`. |
| `make frontend-typecheck` | PASS 2/2 `20260920T045902Z-p16321`. |
| `make frontend-import-boundary-check` | PASS 2/2 `20260920T045902Z-p16330`. |
| `make lint-biome` | FAIL 1/2 `20260920T045902Z-p16337`: new test's non-null assertion. Replaced it with an explicit missing-fixture guard; rerun below. |

The full frontend run's 654 passing units plus the unchanged failed gate's
corrective pass establish the selected corpus without rerunning unrelated tests.
The shared selector preserves bytes; current Entity production regressions are
rerun after that cutover. No failure is hidden by changing assertions or goldens.

| Remaining terminal command | Result / root |
| --- | --- |
| Current full Entity Find browser selection (eleven scenarios) | PASS 11/11 execution units `20260920T045930Z-p19271`, including both entity renderers and all interaction assertions after selector cutover. |
| `make lint-biome` corrective run | PASS 2/2 `20260920T050034Z-p81960`. |
| `make lint-markdown` | PASS `20260920T050034Z-p81980`. |
| `make browser-e2e-visual` | PASS 12/12 units `20260920T045404Z-p32268`; 44 catalog rows, 47 Playwright scenarios. |
| `make browser-e2e-a11y` | FAIL 18/20 units `20260920T045404Z-p32412`; 65 scenarios passed, one unrelated Reference Pack observation-retry scenario failed. All affected Workbook/grid scenarios passed. |
| Unchanged Reference Pack corrective isolation: `make service-backed-test-slice OWNER=web.design ROWS=web.design.accessibility.reference_pack_administration` | PASS 11/11 units `20260920T050019Z-p56330`. No reference-pack source, assertion or timeout changed. The broad-run failure did not reproduce; its cause is not established. |
| Four Timeline measurement rows through `make service-backed-test-slice OWNER=module.timeline` | 18/20 units at `20260920T045435Z-p85474`: focus p95 30 ms, selection 40.3 ms and typing 38 ms pass their 100 ms limits. Blank creation p95 156.9 ms exceeds 150 ms; isolated rerun follows. |

Visual reconciliation PASS: `browser-e2e-visual/frontend-visual-reconciliation.json`
accounts for all 252 capture intents and committed goldens, 29 registered fixtures,
zero orphan/missing/ambiguous entries and zero unresolved registered fixtures.
Renderer profile, viewport, masks, scroll normalization, screenshot scope,
tolerances and golden bytes are unchanged. The scoped new Find chrome was reviewed
in retained production screenshots. Ordinary comparisons already pass, so the
guide requires retaining those goldens: update and the two-post-refresh-run rule
are N/A because no renderer or golden refresh occurred. No snapshot was refreshed
to hide a failed assertion.

Measurement uses the existing 100-sample AC-043 fixture, background sessions and
traffic. The first run overlapped other verification; that is a possible source
of variability, not an established explanation. The isolated corrective run uses
the unchanged fixture, predicate and 150 ms threshold. No sample, threshold,
assertion, traffic or warmup was changed.

## Acceptance assessment

Each row applies only to this adopted capability and affected boundaries. The
following evidence abbreviations refer to exact roots above: **Entity** = final
11-scenario run `20260920T045930Z-p19271`; **Timeline** = nine Find scenarios in
`20260920T044528Z-p43009`; **Adapter** = 56/56 local units; **Recovery** = current
12-scenario autosave/query selection plus coordinator and recovery-navigation
units; **A11y** = affected passing rows in the ordinary accessibility run;
**Visual** = ordinary visual run and reconciled 252 captures; **Types/boundaries**
= final type/import/backend checks. Historical evidence is not substituted.

| ID | Disposition | Evidence / bounded rationale |
| --- | --- | --- |
| A001 | PASS | Core 03 §13.5 / shortcut matrix, §2.3A, REQ-03-298/299/100 and design §§7–8,10,14 map Find, drafts, focus and authority; Core 01 REQ-01-267 maps the publication repair. Source and verification owners are separately recorded. |
| A002 | PASS | G1–G9 rubric: one matching engine/control and shared attachment/lifecycle boundary; source adapters retain renderer and settlement meaning. Superseded hook and collection helper removed with migrated callers. |
| A003 | PASS | Clean baseline main/HEAD recorded; final dirty state is this slice. Authored source/import/generated/catalog/topology inputs and later handoffs revalidated; RDG remains Adapter-private. |
| A004 | PASS | Existing Find chrome, styles and tokens reused; no new design literal, theme, density registry or component geometry. |
| A005 | PASS | Existing dark_graphite presentation only; Visual and Entity screenshots. |
| A006 | PASS | Entity compact/default/comfortable density, virtualized frozen columns and 2× zoom; Adapter and Visual density fixtures. |
| A007 | PASS | Entity recordless draft remains exact with no Find submission; accepted out-of-query creation excluded until query admission. Existing paste/create regressions preserve owner capabilities. No new create mode. |
| A008 | PASS | Entity 1024×720, 768×640, large viewport, zoom and spacing checks; existing column viewport/A11y/Visual responsive and fallback rows. Responsive owner unchanged. |
| A009 | PASS | Entity panel bounds/internal scrolling, grid viewport, status and account-menu focus; A11y/Visual shell overflow. |
| A010 | PASS | Entity Inspector subject and independent scalar/alias drafts survive navigation and focus borrowing; explicit submission and current attachment restoration remain source-owned. Existing inspector A11y and frontend regressions. |
| A011 | PASS | Entity actual offscreen focus, accepted/rejected/superseded moves, saved-view replacement, grouping, collapse, frozen/hidden columns and surface departure; Timeline and Recovery. |
| A012 | PASS | Find produces no reads/writes or transaction IDs. Pending movement and Close/new intent produce one source-owned write; existing exact-replay and duplicate-activation rows pass. Crypto/request identity owners unchanged. |
| A013 | PASS | Entity accepted-write/failed-refresh preserves newer raw work and recovers through reads with exactly one write; Recovery verifies retained request and owner-specific recovery. |
| A014 | PASS | Exact raw scalar, alias and creation drafts, rejection, accepted settlement, native editor keys, IME, Escape and restoration through production Entity and Timeline renderers. |
| A015 | PASS | Source editor retains local validation/conflict feedback; independent committed search excludes raw draft. Existing autosave/recovery and A11y conflict state evidence passes. |
| A016 | PASS | Entity query pending/empty/stale/read-only states compose independently; stale retained rows stay searchable and failed replacement does not substitute requested membership. Types and existing operational-state regressions pass. |
| A017 | PASS | Entity suspension, recovered viewer and incident revocation clear protected Find independently of retained authoring. Recovery confirms accepted reads before replay; source read guards remain intact. |
| A018 | N/A | No Evidence lifecycle, overlay, preview or download behavior changes. Existing Timeline Evidence-summary text and affected rendering regressions still pass. |
| A019 | PASS | Entity keyboard ownership, labelled control, non-color cues, actual unobscured focus, announcements without terms and menu recovery; Timeline and affected A11y rows. Unrelated Reference Pack failure/repass disclosed above. |
| A020 | PASS | Existing component precedence and styles retained; Entity geometry/text spacing/zoom and Visual cover the affected compound states. |
| A021 | PASS | Entity 85-row two-axis focus and 405-row bounded-window/eviction cases, Adapter and Timeline pass. All four measurement predicates pass; isolated blank creation p95 is 77.5 ms against 150 ms, with 100 unchanged samples. |
| A022 | PASS | Visual production renderer, pinned profile and reconciliation pass; 252 active goldens, no missing/orphan/ambiguous entries, no golden changes. New Entity screenshots manually inspected. |
| A023 | PASS | Shared record/field editor selector, unchanged bytes; architecture selector corrective gate and package.ui 10/10, followed by full Entity browser run. |
| A024 | PASS | Executable changes consume authored contracts and source capabilities only. No Markdown/digest dependency added; frontend architecture/import and generated-policy/drift checks pass. Documentation lint remains separate maintenance. |
| A025 | PASS | Only design presentation and authored routing generate derivatives through Make; finalizer, policy, JSON shape and generated drift pass. No hand-edited generated roots or lockfiles. |
| A026 | PASS | Explicit internal hook cutover and helper retirement; no data, endpoint, schema, dependency, layout or persistence migration. Existing selector bytes and publication wire contract preserved. Rollback below. |
| A027 | PASS | All four workstreams complete; final measurement, Markdown and diff checks pass. This handoff records all changed paths, owner decisions, evidence, failures/corrections, skipped checks, limitations, compatibility, rollback and next action. |

## Limitations, skipped work and rollback

- Find intentionally covers only accepted loaded rows on Timeline, Hosts and
  Identities. Other surfaces, exhaustive incident/server search, additional page
  scans, Replace, fuzzy matching and durable terms remain outside scope.
- Entity primary display fallback is deliberately preserved: Identity UPN's
  current renderer prefers Email when both exist. Find follows that displayed
  value; it does not silently change entity formatting or identifier semantics.
- The full frontend and accessibility commands each had one failure; exact
  corrective passes and retained failed roots are disclosed above. No remaining
  applicable failure is accepted or hidden. The first creation measurement
  failed its original threshold; isolated correction passes at 77.5 ms. The
  overlap explanation remains an inference, not a diagnosed product defect.
- Full `make check`, release/CI, unrelated server/network performance suites,
  deployment and Core 05 publication were not selected. This is owner-scoped
  implementation evidence, not a release or exhaustive conformance claim.
- `RESULTS_DIR` was unset; retained-run maintenance and duration-baseline refresh
  were skipped. Visual update/two-post-refresh passes are N/A because the ordinary
  visual comparison passes without changing any golden or renderer.
- No commit, push, deployment, endpoint, dependency, database/saved-layout
  migration, analyst-data change, search logging or persistence was performed.
  Test incidents and fault injection stayed inside Make-owned isolated fixtures.
- Rollback removes this slice's coordinated specs/design/domain navigation,
  shared Find/source adapter/entity presentation integration, selector cutover,
  three integration-exposed owner repairs, tests/routing and generated design/
  topology derivatives. Restore the previous Timeline hook and its callers as
  one change. Preserve unrelated work and all already accepted writes; no data
  rollback or migration is needed. No visual golden rollback is needed.

## Changed-file inventory

| State | Path |
| --- | --- |
| Modified | [apps/web/e2e/timeline-find.spec.ts](../../../apps/web/e2e/timeline-find.spec.ts) |
| Modified | [apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.test.ts](../../../apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.test.ts) |
| Modified | [apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts](../../../apps/web/src/workbook/collaboration/WorkbookCollaborationCoordinator.ts) |
| Modified | [apps/web/src/workbook/collaboration/workbookAuthorizationRecoveryMachine.ts](../../../apps/web/src/workbook/collaboration/workbookAuthorizationRecoveryMachine.ts) |
| Modified | [apps/web/src/workbook/components/EntityWorkbookSurface.tsx](../../../apps/web/src/workbook/components/EntityWorkbookSurface.tsx) |
| Modified | [apps/web/src/workbook/components/WorkbookGridEditorControl.tsx](../../../apps/web/src/workbook/components/WorkbookGridEditorControl.tsx) |
| Modified | [apps/web/src/workbook/features/entities/useEntityWorkbookInspectorComposition.tsx](../../../apps/web/src/workbook/features/entities/useEntityWorkbookInspectorComposition.tsx) |
| Modified | [apps/web/src/workbook/find/README.md](../../../apps/web/src/workbook/find/README.md) |
| Modified | [apps/web/src/workbook/find/WorkbookFindControl.tsx](../../../apps/web/src/workbook/find/WorkbookFindControl.tsx) |
| Modified | [apps/web/src/workbook/find/WorkbookFindController.ts](../../../apps/web/src/workbook/find/WorkbookFindController.ts) |
| Modified | [apps/web/src/workbook/models/README.md](../../../apps/web/src/workbook/models/README.md) |
| Modified | [apps/web/src/workbook/models/genericWorkbookModel.test.ts](../../../apps/web/src/workbook/models/genericWorkbookModel.test.ts) |
| Modified | [apps/web/src/workbook/models/genericWorkbookModel.ts](../../../apps/web/src/workbook/models/genericWorkbookModel.ts) |
| Modified | [apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx](../../../apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx) |
| Modified | [apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts](../../../apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts) |
| Modified | [apps/web/src/workbook/timeline/hooks/README.md](../../../apps/web/src/workbook/timeline/hooks/README.md) |
| Removed | `apps/web/src/workbook/timeline/hooks/useTimelineFind.ts` |
| Modified | [apps/web/src/workbook/timeline/useTimelineFind.test.tsx](../../../apps/web/src/workbook/timeline/useTimelineFind.test.tsx) |
| Modified | [contracts/design/presentation.v1.json](../../../contracts/design/presentation.v1.json) |
| Modified | [docs/design.md](../../../docs/design.md) |
| Modified | [docs/domain.md](../../../docs/domain.md) |
| Modified | [docs/spec/03_workbook_interaction_collaboration_and_workflows.md](../../../docs/spec/03_workbook_interaction_collaboration_and_workflows.md) |
| Modified | [internal/modules/collaboration/publication.go](../../../internal/modules/collaboration/publication.go) |
| Modified | [internal/modules/collaboration/record_change_intent_test.go](../../../internal/modules/collaboration/record_change_intent_test.go) |
| Modified | [packages/grid-adapter/src/SemanticDataGrid.tsx](../../../packages/grid-adapter/src/SemanticDataGrid.tsx) |
| Modified | [packages/ui-contracts/src/generated/design-presentation.ts](../../../packages/ui-contracts/src/generated/design-presentation.ts) |
| Modified | [packages/ui-contracts/src/gridSelectors.ts](../../../packages/ui-contracts/src/gridSelectors.ts) |
| Modified | [packages/ui-contracts/src/index.ts](../../../packages/ui-contracts/src/index.ts) |
| Modified | [tools/browser_e2e_batch_manifest.json](../../../tools/browser_e2e_batch_manifest.json) |
| Modified | [tools/execution_topology_render_index.json](../../../tools/execution_topology_render_index.json) |
| Modified | [tools/frontend_source_ownership.json](../../../tools/frontend_source_ownership.json) |
| Modified | [tools/test_families/module.workbook.json](../../../tools/test_families/module.workbook.json) |
| Modified | [tools/test_families/web.workbook.json](../../../tools/test_families/web.workbook.json) |
| Added | [apps/web/e2e/entity-find.spec.ts](../../../apps/web/e2e/entity-find.spec.ts) |
| Added | [apps/web/src/workbook/find/useWorkbookFind.test.tsx](../../../apps/web/src/workbook/find/useWorkbookFind.test.tsx) |
| Added | [apps/web/src/workbook/find/useWorkbookFind.ts](../../../apps/web/src/workbook/find/useWorkbookFind.ts) |
| Added | [apps/web/src/workbook/models/entityCellPresentation.test.ts](../../../apps/web/src/workbook/models/entityCellPresentation.test.ts) |
| Added | [apps/web/src/workbook/models/entityCellPresentation.ts](../../../apps/web/src/workbook/models/entityCellPresentation.ts) |
| Added | [apps/web/src/workbook/timeline/hooks/useTimelineFindSource.ts](../../../apps/web/src/workbook/timeline/hooks/useTimelineFindSource.ts) |
| Added | [docs/handoffs/ui-ux/workbook-entity-find-navigation-refactor-handoff.md](../../../docs/handoffs/ui-ux/workbook-entity-find-navigation-refactor-handoff.md) |

## Final delivery record

Isolated `make service-backed-test-slice OWNER=module.timeline
ROWS=module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13`
PASS 14/14 execution units at `20260920T050227Z-p95143`: 100 measured samples,
p50 64.6 ms and p95 **77.5 ms**, threshold 150 ms. Its
`browser-e2e-measurement/browser-groups/measurement-measurement-timeline-grid-afddd2ce13/frontend-measurement-observation.v2.json`
retains the result. The other three predicates passed in the original four-row
run; no measurement owner, input, sample or threshold was changed.

Final source review: the sole matching engine and existing control remain;
Timeline callers use the source adapter and the superseded hook is absent.
Entities consume their existing accepted query and shared renderer decisions;
Find has no endpoint, reads, writes, term logging or persistence. Inspector
subject suppression is limited to Find-admitted focus. Recovery, editor and
publication fixes stay in their existing owners. Generated differences trace to
the authored presentation and verification manifests. The digest, lockfiles,
analyst data, visual goldens and renderer configuration are untouched.

Final branch remains `main`, HEAD
`601aeb6c6f04923ffedd5d0cad979d8fd018c28d`. There were no pre-existing dirty files;
the final 40-file dirty set is this uncommitted slice, including seven new files
and the retired Timeline hook. No commit, push or deployment was performed.
`git diff --check` passes. Final documentation check: `make lint-markdown` PASS at
`20260920T050652Z-p29222` (`adhoc/lint-markdown/tool-run-summary.json`).

Final next action: review the uncommitted implementation and this completed
handoff. No implementation follow-up or migration is required for the adopted
Hosts/Identities capability; broader release/claim publication is outside this task.

EFN-04 binary exit: **PASS**. A001–A027 are complete: 26 PASS and one justified
N/A (A018); no applicable BLOCKED row or unresolved implementation work remains.
