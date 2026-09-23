# Timeline attachment feedback cleanup

## Scope and execution record

Started on clean `main` at `e1762903c6135519597a698e4b4e8e4f56e203c1`.
No pre-existing tracked or untracked user changes were present. The branch and
commit remain unchanged; the final dirty state consists of this slice. No commit,
push, deployment, dependency, endpoint, persistence or data migration is included.

Read repository `AGENTS.md`, the digest's `START_HERE.md`, `LOCAL_AGENT_PROMPT.md`,
`REPO_MAP.tsv`, `OWNER_MAP.tsv`, `rules.tsv`, `acceptance.tsv` and
`QUERY_RECIPES.md`; followed current Workbook/Timeline/Evidence source guides.
The digest is dated advisory navigation. Current authored source ownership,
import boundaries, generated policy and test-family routing were revalidated.
Current app inputs declare React 19.2.5, TypeScript 6.0.2, Vite 8.0.8,
Playwright 1.59.1 and Vitest 4.1.4. Direct vendor grid integration remains solely
in `packages/grid-adapter`. No new direct grid import or localization registry
was introduced. The exhaustive generated registry remains
`tools/generated_artifact_policy.json`.

| Sequential exit | Status | Evidence |
| --- | --- | --- |
| Characterize before production edits | DONE | Baseline browser captures and observations below. |
| Implement owner projection and local presentation | DONE | Narrow owner and component tests; source and API review. |
| Verify production behavior | DONE | Recovery, targeting, continuity and clipping-aware geometry passed; ordinary visual differences reconciled. |
| Finalize and handoff | DONE | Golden maintenance, two ordinary passes, final checks and acceptance below. |

## Authority, owners and selection rubric

| Change | Governing authority | Source placement | Verification routing |
| --- | --- | --- | --- |
| Retained operation classification and identity | Core 03 §8.1; §§3–4, REQ-03-089; Core 01 §3.3.5 and Evidence transaction contracts | `web.workbook`, existing file owner | `module.evidence` retained recovery/draft/targeting rows |
| Compact local summary, bounded detail and completed history | Design §§11.4, 12.7; Core 03 §8.1 retention | Timeline components and presentation under `web.workbook` | `web.workbook` feedback regression; `module.evidence` browser, accessibility and visual rows |
| Focus, announcements and detachment | Design §§12.7, 14.2; Core 03 REQ-03-100/299 | Focused Timeline leaf and Inspector detail | Component focus/announcement tests; production recovery/continuity/accessibility |
| New file and test routing | Current source/import manifests; adopted harness routing | Authored ownership and test-family JSON | `web.architecture`, generation and policy checks |

The completed Recovery-navigation handoff explicitly retained local file feedback.
This slice preserves that boundary and does not contribute to the shell Recovery
catalog. Design clarification is recorded at its existing Evidence, Inspector and
announcement owners. It does not change upload, write or authorization contracts.

| Rubric decision | Assessment |
| --- | --- |
| Observed weakness | Confirmed expanded completed cards consume additional grid height and offer discard merely to remove success. Structural weakness: file/admission subscriptions rebuild grid presentation. Focus loss during new completion presentation was found and repaired in browser verification. |
| Remediation and change areas | Authorized behavior correction: compact summary, explicit bounded recovery disclosure, collapsed completed history without discard. Structural movement: leaf subscriptions and one Timeline detail/action binding. Narrow read-only owner facts support both. Tests, routing, guides and bounded design direction follow those changes. |
| Common decision and boundary | The file owner decides progress, uncertainty, completion and permitted commands. Timeline decides disclosure, detail placement, focus and announcements. No second operation machine, notification store or queue. |
| Rationale and long-term benefit | Retained success no longer grows the resting viewport; stage recovery and receipts survive independently. Grid presentation no longer observes file publications directly. Shared Timeline action binding prevents grid/Inspector divergence. |
| Future extension path | Additional owner-issued stages can extend the read-only projection and existing detail renderer. No speculative cross-surface framework was implemented. |
| Capability value | Exact replay, original-source review, read-only refresh, native capture and draft promotion remain useful recovery and capture capabilities. Stable work identity supports same-source replacement without suppressing new work. |
| Retirement | Removed the hook's expanded recovery mapping, its file/admission subscriptions and duplicated Inspector action mapping. Removed Timeline visual/a11y fixture discard-on-success behavior. |
| Retention and compatibility | Existing `EvidenceFileRecovery` API and existing-Evidence consumer remain unchanged. File/session/finalization owners, Inspector attention navigation and authorization concealment remain. Owner admission snapshot now supplies an event identity; all actual consumers were migrated. |
| Migration | Memory-local projection/presentation only; no data or persistent-state migration. Existing owner replacement of a completed same-source entry is retained, not expanded into archival history. |
| Risk and validation | Uncertainty must never become success, and collapse must never mutate owner work. Tests cover typed facts, exact request bodies, retained receipts, no presentation commands, focus and production geometry. Render diagnostics establish propagation, not latency. |

## Characterization before production changes

Baseline narrow unit run: `.cartulary/test-results/20260923T174513Z-p47113`.
Baseline uncertain-stage/accessibility run:
`.cartulary/test-results/20260923T174558Z-p48238`. Both passed after exposing the
repository-installed Node runtime on `PATH`. Earlier `service_start_error`
attempts could not resolve `node`; no repository repair was needed.

The added characterization scenario first failed to locate a virtualized cell;
the fixture was corrected to use the semantic grid scroll helper. Baseline
characterization then passed in
`.cartulary/test-results/20260923T194953Z-p61308`. Stage and source-review
characterization passed in `.cartulary/test-results/20260923T194736Z-p25122`.
JSON and PNG attachments are retained inside each browser group's
`playwright-report.json`.

At 1280×720, the work area was `(0,88,1280,572)`. Grid height after one completion
was 484.203125 px, with 87.796875 px occupied above it. After two and three
completions it was 429 px, with 143 px occupied. Pending upload, admission error
and four completions retained that expanded 143 px strip. Every completed card
offered `Discard retained file work`.

| State | Observed baseline controls and semantics |
| --- | --- |
| Several completed attachments | Expanded filename/source/success cards and discard per retained source; internal overflow at the ceiling. |
| Upload in progress | Stage explanation and owner stop/discard command above grid; editing remained admitted. |
| Uncertain preparation/transfer/finalization | Exact-stage explanation and Resume; uncertain transfer recovery finalized without transferring again. |
| Accepted Evidence awaiting association | Saved Evidence remained distinct from linked Timeline; original-source review and Resume remained available after remount. |
| Original-source review | Review original source, reviewed source text/version, Use reviewed source and Resume; unrelated Inspector selection did not retarget it. |
| Accepted attachment, failed refresh | Attachment accepted with Refresh pending; Refresh sent reads and did not repeat writes. |
| Stopped uncertain work | Stopped explanation retained uncertainty and the exact replay path. |
| Admission rejection | Choose one file at a time was immediately visible; no upload began. |

The synthetic drop-only geometry fixture left focus on BODY throughout; that is
not evidence of user focus loss. Existing native editor and accessibility tests
provide deliberate focused-control evidence. Baseline live-region observations
captured Syncing/Saved/query refresh; the initial observer did not establish that
newly inserted error regions were announced. Source inspection showed duplicate
recovery copies already used `aria-live=off`.

Baseline presentation/grid render observations were 40 after one completion,
63 after two, 83 after three, 91 during held upload, 92 after isolated admission
rejection and 108 after completion. The isolated rejection changed both counts
without changing the capped geometry. These are diagnostics, not a latency
measurement. Slow typing was a hypothesis, not a measured defect.

## Implementation and retained behavior

- `WorkbookTimelineFileOwner.ts` projects unconditional `workId`, stable
  `outcomeIdentity`, typed `feedbackKind` and admission event identity. Uncertainty
  takes precedence over ordinary progress. Stopped work remains attention;
  completion requires accepted/present association plus complete refresh.
  New-file admission uses that same completion fact, so a source association
  verified by a read also permits explicit replacement after refresh; it never
  needs a fabricated link receipt or presentation discard to unlock the row.
  Projection contains no File bytes, upload capability or captured request.
- `TimelineAttachmentFeedback.tsx` owns the two subscriptions, compact count
  trigger, immediate admission message, internal scrolling, local focus fallback
  and one assertive/polite announcement source. Background events never open it.
  Completed details omit discard and are separately collapsed. No dismissal
  feature or suppression store exists.
- `TimelineWorkbookGrid.tsx` passes a stable owner to the leaf and establishes
  the work area as the shared container-relative geometry boundary.
  `useTimelineWorkbookPresentation.tsx` retains accepted-row rendering and no
  longer maps/subscribes retained files.
- `TimelineEvidencePanel.tsx` shares source-specific detail and commands. Its
  local fallback ref stays attached while Inspector registry callbacks change;
  completion does not drop focus to BODY.
- Existing targeting, gesture capture, clipboard, draft promotion, mutation
  owners and explicit Inspector editing are retained. Generic existing-Evidence
  recovery is not subjected to Timeline's disclosure policy.

The new component/unit test is mapped in `tools/frontend_source_ownership.json`.
New selectors are authored in `tools/test_families/{web.workbook,module.evidence}.json`;
browser grouping/topology outputs were regenerated with `make generate`, never
edited directly. Browser changes cover the new scenario, existing recovery,
accessibility, visual setup and the shared explicit-disclosure helper. Local
source READMEs describe the new boundary.

## Advisory dispositions

The narrow offline UX query `loading error recovery focus disclosure` returned
five results without fallback. The corresponding existing `rules.tsv` evidence
cells record this slice: ADOPT R002 local focus preservation/fallback, R006 local
accessible error feedback, R008 stage-specific exact replay/read-only refresh;
ADAPT R024 field placement to a bounded workbook disclosure. Enhanced AAA advice
was not imported. No new palette, tokens or upstream design system was generated.
Product tests and generators do not consume Markdown.

## Verification and artifacts

All repository verification commands use public Make targets with
`PATH="$PWD/tmp/node-runtime/bin:$PATH"`. Current task guides selected
`web.workbook`, `module.evidence` and, for the new boundary, `web.architecture`.

| Command/selection | Result and artifact |
| --- | --- |
| `make generate` | PASS; `20260923T194700Z-p21953`, `20260923T200748Z-p57715`. |
| Evidence retained recovery, draft recovery, picker targeting unit rows | PASS; `20260923T203615Z-p50004`, 4/4 work units, including authoritative-association replacement after read-only refresh. |
| Workbook feedback and stateless presentation rows | PASS; `20260923T201116Z-p56349`, 3/3 work units. |
| `make test-slice OWNER=web.architecture` | PASS; `20260923T200739Z-p56109`, 12/12 work units. |
| Stage recovery and original-source browser rows | PASS in `20260923T200627Z-p18372`; overall run failed other new assertions. |
| Recovery accessibility | PASS in `20260923T200912Z-p61428`; overall run failed new geometry assertion. |
| Production geometry matrix | PASS; `20260923T201842Z-p63206`, 11/11 work units; all three densities, four viewport sizes, 200% zoom and text spacing. |
| Targeting/native editing/draft ordering/stage recovery | Eight selected rows PASS in `20260923T201059Z-p6506`; three old selectors required explicit disclosure/admission scoping. |
| Corrected background/delayed-source/picker-lifetime and recovery accessibility | PASS; `20260923T201533Z-p84256`, 13/13 work units. |
| `make frontend-typecheck` | PASS; `20260923T201809Z-p58363`, 2/2 work units. |
| `make lint-markdown` | PASS; `20260923T203725Z-p54821`. |
| `make browser-e2e-visual` before refresh | Expected comparison failure; `20260923T201101Z-p6808`, five changed goldens in three rows, all other scenarios pass. |
| `make agent-finalize` with `RESULTS_DIR` unset | PASS; `20260923T202053Z-p97869` and final reruns `20260923T203012Z-p96248`, `20260923T203614Z-p49786`. Retained-run/performance maintenance explicitly skipped: no eligible successful full warm check supplied. |
| Final type/import/lint checks | PASS; `frontend-typecheck` `20260923T203725Z-p54773`, `frontend-import-boundary-check` `20260923T202142Z-p2496`, `lint-biome` `20260923T203725Z-p54792`. |
| Final JSON/policy/drift checks | PASS; `json-shape-check` `20260923T202142Z-p2438`, `generated-artifact-policy-check` `20260923T202142Z-p2436`, `generate-drift` `20260923T202142Z-p2434`; catalog check also passed. |
| Geometry with isolated accepted-row diagnostics | PASS; `20260923T202242Z-p40374`, 11/11 work units. Held-write and accepted-row observations are separately captured. |
| `make browser-e2e-visual-update` | PASS; `20260923T202141Z-p2333`, 12/12 work units; exactly five goldens and their generated manifest changed. |
| Final feedback unit row | PASS; `20260923T202806Z-p60872`, including coincident admission/failure aggregation and silent retained remount. |
| Two fresh `make browser-e2e-visual` runs | PASS; `20260923T202658Z-p86600` and `20260923T202658Z-p86599`, each 12/12 work units and 47 browser executions, zero unexpected/flaky/skipped. Both reconcile all 255 active goldens. |
| Existing-Evidence compatibility browser row | PASS in `20260923T202808Z-p61353`; the additional accessibility worker could not reach service readiness, so that combined run failed infrastructure before its scenario started. Isolated accessibility rerun PASS in `20260923T203228Z-p4008`, 11/11 work units; no product fix was needed for the readiness timeout. |

Run roots above are under `.cartulary/test-results/`. Graph runs expose
`run-summary.json`; generation and Markdown maintenance expose the
`tool-run-summary.json` subpath reported in command output. Final review also found a pre-existing completed-association admission mismatch:
a read-verified association was completed for presentation but `begin` still
required a local link receipt. The owner now uses its common completed fact for
explicit same-source replacement. Regression evidence covers blocked replacement
while refresh debt remains, read-only refresh, no invented link write/receipt and
fresh work identity afterward.

Intermediate errors included authored catalog family/title
ordering, JSDOM native-details selectors, TypeScript test/ref-cleanup types and
semantic-element lint. They were repaired without weakening operation tests.
Browser verification exposed a real Inspector fallback-ref gap and a ceiling
measured against an ancestor rather than the work area; both were repaired.
Manual screenshot review then found nested scrollport starvation at 200% zoom.
One bounded internal scroll owner and token-based scroll padding now preserve
full control visibility, verified with intersection/clipping checks rather than
viewport rectangles alone.
The first newly visible admission changes geometry and may legitimately render
the grid; repeated admission with unchanged geometry isolates status propagation.

### Production geometry and announcements

The final geometry scenario retains a 28 px resting summary after one, two and
three completions at 1280×720, leaving a 544 px grid in a 572 px work area. In the
compact diagnostic, the first visible admission adds 28.296875 px and one layout
render; the repeated identical event leaves presentation/grid counts unchanged
(48/48 before and after). Accepted row changes continue through normal rendering.
The held-write observation was 75/75; acceptance and row reconciliation raised
it to 81/81 without a geometry change. These counts make no speed claim.
Each rejection creates one new assertive emission even when text repeats;
disclosure/history interaction creates no outcome or request. Ordinary Saved
remains with save status. Filename/stage progress remains polite and Inspector
copies remain off. Coincident rejection and actionable failure share one
assertive emission without losing either explanation. Protected live copy is
concealed in the same render as an unavailable owner snapshot. The scroll/keyboard scenarios preserve the active trigger on
background completion; native-editor scenarios separately preserve text, caret,
selection and IME behavior during transfer.

Opened feedback occupies at most 143 px at 1280×720 and 1024×720, 123 px at
768×640, and 83 px at 390×480 in the recorded compact profile. Long filenames,
all three owner densities, 200% zoom, text spacing and focus-ring padding are
covered. Screenshots were manually inspected; the 390 px below-minimum layout
is resilience evidence, not a new mobile conformance claim.

### Visual refresh record

Accepted trigger: authorized compact retained completion replaces fixture-driven
discard after successful attachment. The prior ordinary run's reconciliation v3
accounts for 255 capture intents and 255 active goldens, zero orphan, missing or
ambiguous mappings and all 29 registered fixtures. Its sole reconciliation error
is the failed comparison target. Five actual images were manually reviewed;
changes follow the retained 28 px strip and corresponding grid height. No
viewport, zoom, mask, scroll anchor, screenshot scope or tolerance was changed.

| Authored row | Fixture IDs | Changed golden filenames (snapshot directory) |
| --- | --- | --- |
| `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.default_timeline_workbook_shell`, `visual.fixture.narrow_desktop_workbook_shell`, `visual.fixture.compact_desktop_workbook_shell` | `incident-directory-default-timeline-workbook-shell-linux.png`, `incident-directory-narrow-desktop-workbook-shell-linux.png`, `incident-directory-compact-desktop-workbook-shell-linux.png` |
| `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830` | Active nonregistry capture `visual.capture.47ccc6b491fd5e186555` | `evidence-grid-timeline-evidence-badge-linux.png` |
| `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4` | Active nonregistry capture `visual.capture.b780497df76862578d0c` | `evidence-timeline-evidence-count-linux.png` |

The module.workbook guide was consulted for the changed shell fixture.
The update passed in `20260923T202141Z-p2333`; all five promoted images were
manually inspected. The final disclosure chevron and native completed-history
marker use existing icon/control contracts. Reconciliation is PASS with all 255
active captures and no errors. The two Evidence images retain their existing
grid crop and therefore record resulting grid geometry; the new production
browser attachments separately show the local feedback controls. Two fresh
ordinary validation runs passed in `20260923T202658Z-p86600` and
`20260923T202658Z-p86599`. Both reconciliations pass with 255 active goldens, no
orphans, missing goldens, ambiguous mappings or unresolved registered fixtures.
Generated-policy/JSON/drift checks after promotion also pass in
`20260923T202659Z-p86770`, `20260923T202659Z-p86772` and
`20260923T202659Z-p86768`.

## Acceptance assessment

Assessment is limited to this slice. PASS does not assert unmodified subsystem
completeness; N/A identifies an untouched boundary.

| Row | Status | Evidence or scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Exact Core/design owner mapping above; typed projections and verification routing remain downstream. |
| A002 Scope | PASS | Selection rubric identifies one presentation/projection boundary, behavior corrections, retirement and retained consumers. |
| A003 Repository state | PASS | Clean baseline, current commit/branch, source guides, authored manifests, toolchain inputs and direct-grid boundary inspected. |
| A004 Tokens | PASS | Existing quiet command, typography, spacing, icon, focus and work-area ceiling tokens reused; no second registry. |
| A005 Theme | PASS | Existing dark_graphite colors only; production screenshot review and ordinary visual scenarios. |
| A006 Density | PASS | All three production preference densities explicitly asserted in geometry scenario; shared grid/typography geometry retained; visual density variants pass. |
| A007 Creation | PASS | Draft creation orderings, picker promotion, zero-byte attachment and original-source tests retain existing capabilities/payloads. No creation policy change. |
| A008 Responsive | PASS | 1280×720, 1024×720, 768×640, 390×480, vertical resize, 200% zoom and text-spacing captures; controls tested against clipping ancestors. Responsive threshold algorithm untouched. |
| A009 Overflow | PASS | Feedback remains inside quarter work area with one internal scroll owner; grid/status/safe navigation remain visible in reviewed screenshots. |
| A010 Inspector | PASS | Shared source-specific detail and existing attention navigation; review and late-source tests pass; completed history excludes unfinished attention. |
| A011 Continuity | PASS | Native editor/IME, caret/selection, original-source, filtering/detachment and same-row replacement production scenarios. |
| A012 Transactions | PASS | Stage recovery asserts exact slot/create/link bodies and one transfer; no new transaction generation path. |
| A013 Recovery | PASS | Uncertain replay, retained receipts, accepted-write/failed-refresh and read-only refresh assertions; collapse calls no owner actions or requests. Queue recovery itself unchanged. |
| A014 Editing | PASS | Native scalar/collection editing, draft orderings and picker cancellation pass. Explicit Inspector editor remains untouched. |
| A015 Conflict | N/A | Cell conflict rendering/resolution is outside this slice; no cell-error or local-authoring semantics changed. |
| A016 Query and interaction | PASS | Typed feedback remains separate from query/interaction state; owner admission still decides eligibility. Existing row rendering and producer/architecture checks retained. |
| A017 Authorization | PASS | Owner tests conceal snapshots/admission under suspension and retain authority-governed recovery. No request or authorization/security-lifetime checks changed; source-hidden boundary remains. |
| A018 Evidence | PASS | Stage-specific retained recovery and source identity; lifecycle/access/preview contracts and generic consumer unchanged, covered by Evidence visual scenarios. |
| A019 Accessibility | PASS | Production keyboard/focus fallback, long filenames, clipping-aware reachability, repeated-event live-region tests and silent Inspector copies; existing contrast/reduced-motion checks pass. |
| A020 Components | PASS | Compact summary, unfinished detail, completed group and owner-permitted controls tested through density, zoom/text spacing and source-state variants. |
| A021 Virtualization | PASS | Semantic grid targeting helpers, filtering/detachment and original-source continuity pass; grid row identity, virtualization and timing budgets untouched. |
| A022 Visual fixtures | PASS | Five promoted images manually reviewed; maintenance update and two fresh ordinary visual runs pass, with complete reconciliation and unchanged capture profiles. |
| A023 Selectors | PASS | Tests use semantic roles, owner work identity and existing view/record/field builders; architecture selector policy passes. Fiber observations are diagnostics only. |
| A024 Test authority | PASS | Source/dependency diff audit: no product runtime, test, generator or release input reads/stats/hashes Markdown; documentation lint remains maintenance. |
| A025 Generated artifacts | PASS | Authored ownership/routing updated first; Make generation, policy and drift pass; golden candidates use the maintenance target. |
| A026 Compatibility | PASS | Existing-Evidence renderer/consumer preserved; no routes/schema/endpoints/dependencies/persistence changed; explicit no-migration rollback. |
| A027 Handoff | PASS | Owner map, sequential exits, changed/retained/retired paths, failures and resolutions, command artifacts, limitations, skipped maintenance and rollback recorded. |

## Limitations and rollback

Retention remains memory-local under the existing account/incident owner. This
does not promise recovery after reload or persistent completed history. Browser
live-region evidence establishes DOM priority/deduplication and keyboard/focus
behavior, not a manual assistive-technology certification. No speed claim or
performance-budget change is made. No unresolved product blocker remains. Full
backend/release/performance suites were not selected for this frontend slice;
retained full-warm-run maintenance was skipped with `RESULTS_DIR` unset.

Final working-tree review contains only the 29 intended authored/generated/doc
and golden files listed by this slice. Nothing was committed, pushed or deployed.
The next action is user review of the implementation and handoff.

Rollback is reverting this presentation/projection slice and its associated
tests, authored routing, generated routing, docs and affected golden refresh.
There is no data migration. Required Files, receipts and replay semantics remain
under the original owners throughout.
