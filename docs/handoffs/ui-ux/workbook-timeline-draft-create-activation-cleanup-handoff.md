# Timeline trailing-draft Create activation cleanup

## Repository state and scope

- Checkout: `/home/jochi/code/cartulary`, branch `main`.
- Starting commit: `8a62966a095084bfd33589d80a74563d4aeb9635`.
- Starting tracked and untracked state: clean, rechecked at implementation start.
- Authorized slice: native activation of Timeline's existing draft-gutter Create
  button, focused regression coverage, authored verification routing and this handoff.
- No commit, push or deployment. No dependency, endpoint, schema, persistence,
  token, theme, layout or public component-interface change; no migration.
- The digest's September 13 snapshot was advisory navigation. Current React
  19.2.5 and Grid Adapter vendor 7.0.0-beta.59, source ownership, import policy,
  active view projection and current verification catalogs were inspected again.

## Authority and ownership

| Concern | Exact governing owner | Current source / verification boundary |
| --- | --- | --- |
| Create shape, zero-field create, idempotency and original receipt replay | Core 01 §3.3.5, REQ-01-057, REQ-01-069, REQ-01-070 | `timelineMutationIntents`, scalar-save command and retained Timeline mutation driver; `module.timeline` |
| Active create capability and inline-create policy | Core 01 REQ-01-288 and §7.4; `contracts/view-schemas/cartulary.view.timeline.v2.json` projects `create_capable=true` and `permits_zero_field_create=true` | Existing view-contract facade and Grid Adapter editable draft admission |
| Autosave, deduplication, raw retention and recovery | Core 03 §§3–4, REQ-03-035, REQ-03-087–089, REQ-03-099–100 | Editor draft registry, `useTimelineMutationCommands`, retained queue and driver; `module.timeline` and `web.workbook` |
| Incomplete facts and fast capture | Core 03 §7, REQ-03-111–115 | Existing scalar input capture and collection authoring |
| Embedded controls, editing, semantic identity and keyboard ownership | Core 03 §13, REQ-03-217–220, REQ-03-298, REQ-03-300; design §§8.4–8.6, 14.1 | `TimelineDraftRowActions`, production Grid Adapter, editor and continuity owners |
| Closed/read-only and authorization admission | Core 03 REQ-03-287–288, REQ-03-299/100 | Existing runtime authority and editable draft presentation |

Source `web.workbook` owns the changed production component under
`tools/frontend_source_ownership.json`. Direct vendor integration remains solely
inside `packages/grid-adapter`, as required by `tools/frontend_import_boundaries.json`.
`contracts/verification/registry.json`, the two verification-owner contracts,
`tools/test_catalog_owner.json` and `tools/test_families` independently select
verification; they do not establish required interaction behavior.

Inspected the requested starting files and followed their current component,
composition, editing, presentation, mutation and Grid Adapter guides. In particular,
reviewed scalar input/blur, collection blur, editor materialization and accepted
draft identity, pending admission, row-menu focus borrowing and deferred continuity.
`docs/domain.md` supplies vocabulary/navigation; `docs/research/nlspec-spec.md`
supplies background on specifications. Neither creates an additional implementation
task or supersedes the adopted behavioral owners.

## Characterization and selection rubric

| Rubric item | Assessment |
| --- | --- |
| Confirmed defect | Mouse-down submitted before release, including middle/right presses. Click-only activation did nothing. Enter/Space manually submitted from key-down rather than native button activation. |
| Structural weakness | One helper coupled submission with focus suppression. The payload regression encoded press-as-activation. The existing blank-create browser scenario was not independently selected by the owner catalog. |
| Hypotheses investigated | Removing focus suppression could blur-submit raw collection text before click, replace draft identity or alter continuation. Scalar typing already starts fast capture and must be tested separately from unsent collection authoring. |
| Remediation / boundary | Correct only the button's activation adapter and its verification. The hidden decision is how native activation invokes the existing create command; no shared mutation or button framework is needed. |
| Rationale / benefit | A completed native activation has one submission path, while focus preservation has an explicit non-submitting purpose. Tests can distinguish activation from pending admission and authoritative acceptance. |
| Future extension | Other input methods that produce native click can use the same button command without another mutation path. No speculative extension is implemented. |
| Carried-forward value | Raw drafts preserve incomplete facts; runtime deduplication protects one logical create; captured bytes permit exact uncertain replay; semantic continuation preserves fast capture while yielding to newer intent. |
| Retirement / retention | Retire mouse-down submission and manual Enter/Space submission. Retain press-time focus suppression, key propagation isolation, the existing create callback, disabled projection and all retained operation owners. |
| Compatibility / cutover | Same accessible name, selectors, `type=button`, props and request shape. Toolbar Add row and Evidence remain distinct. Native click replaces the incidental mouse-down activation contract; no alias or migration. |
| Risk if left unchanged | Press cancellation cannot prevent an already submitted create, auxiliary buttons create unexpectedly, and click-only assistive activation has no command path. |
| Delivery / rollback | Characterize first; correct the leaf; run owner slices; finalize generated routing and checks; assess acceptance. Roll back this component/test/catalog change and regenerate routing together; no data rollback is required. |

### Event and ownership trace

The production Grid Adapter recognizes embedded buttons as native action targets
and does not start scalar editing or a range gesture for their presses. The
draft gutter differs from ordinary data/action cells: it has no local keyboard
wrapper, so Enter/Space propagation isolation remains on this button.

On an enabled Create press, mouse-down cancels only default focus movement and
propagation. The current input therefore does not blur or submit on that press.
Releasing outside or cancelling contact has no activation command. On a completed
primary click, the button invokes the unchanged `handleCreateBlankDraftRow`.
Keyboard Enter and Space retain browser default activation, which reaches click.
DOM click-only activation reaches that same handler without a pointer prelude.

The command resolves the current row and calls scalar save with
`allowZeroFieldCreate=true`, `continueOnFreshDraft=true`, `surface=grid` and
`preserveInputFocus=false`. The save owner materializes retained scalar and
collection drafts, resolves accepted draft identity and deduplicates authoring
revisions synchronously. The retained queue owns pending-create coalescing,
captured request identity/bytes, rejection and uncertainty. No state is copied
into the button.

Accepted row projection provides one fresh trailing draft. Existing deferred
continuity owns its focus request and cancels restoration for newer pointer,
keyboard, input, wheel, focus, scope or authority intent. Its recent menu and
continuity changes are retained.

Preserved source chain, relative to `apps/web/src/workbook/`:

- `timeline/composition/useTimelineInteractionComposition.ts`:
  `handleCreateBlankDraftRow` resolves the current draft and requests explicit create.
- `timeline/hooks/useTimelineMutationCommands.ts`: `queueScalarSave` materializes
  authoring and performs synchronous revision-based admission.
- `timeline/editing/useTimelineEditorDraftRegistry.ts`: retained raw values,
  captured authoring revisions and accepted draft identity.
- `timeline/mutations/createTimelineMutationDriver.ts` and
  `utils/workbookPendingQueue.ts`: retained operation admission, dispatch,
  captured bytes, authoritative settlement and recovery.
- `timeline/models/timelineAcceptedProjection.ts` and
  `timeline/models/timelineAcceptedMutationEffects.ts`: one fresh draft and
  semantic continuation intent.
- `timeline/presentation/useTimelineWorkbookPresentation.tsx`: current production
  grid projection and distinct toolbar focus, gutter Create and Evidence actions.

## Advisory query and dispositions

Ran the narrow offline query from the repository root:

```bash
PYTHONDONTWRITEBYTECODE=1 python3 -B docs/cartulary-ui-ux-refactor-digest/upstream/ui-ux-pro-max/scripts/search.py \
  "semantic button keyboard focus" --domain ux -n 3 --json
```

It returned three focus results; no fallback was needed. Apply existing
`rules.tsv` R002 (`ADOPT`) within design-owned focus requirements; enhanced AAA
advice does not become a new profile claim. R001/R004/R013 (`ADOPT`) support
native keyboard operation and semantic, named controls. R008 (`ADOPT`) retains
owner-defined recovery. R012 (`ADAPT`) keeps state with the existing retained
owner instead of adding button-local pending state. R033/R034 (`REJECT`) exclude
advisory behavior authority and incidental selectors. No new rule or TSV schema
is necessary; these dispositions are recorded against the maintained rule IDs.

## Verification record

Commands ran at repository root. Make invocations used command-local
`PATH="/home/jochi/code/cartulary/tmp/node-runtime/bin:$PATH"`. No `RESULTS_DIR`
was supplied: these narrow selections are not eligible full warm-check evidence.
No product check reads or depends on this handoff, the digest or other Markdown.

### Selection and before/after evidence

Both requested discovery commands passed:

```bash
make task-guide ROLE=module-author OWNER=web.workbook
make task-guide ROLE=module-author OWNER=module.timeline
```

The following run identifiers are directories under `.cartulary/test-results/`.
Slice results have `run-summary.json` and `unit-results/`. Browser artifacts are
under `browser-e2e-webserver-backed/browser-groups/`, including each selected
group's `playwright-report.json` and failure context/trace attachments.

| Run | Command / selection | Result and disposition |
| --- | --- | --- |
| `20260922T131207Z-p4902` | Initial baseline `make test-slice` | Infrastructure failure: scheduler could not launch `node` from PATH. Recovered with command-local bundled runtime PATH; unrelated to Create behavior. |
| `20260922T131431Z-p6661` | Baseline Timeline frontend payload selection | PASS, two rows. Existing tests established payload behavior, not native activation. |
| `20260922T131632Z-p40624` | Baseline workbook regression selection | PASS, four rows. |
| `20260922T131516Z-p7607` | Baseline Timeline service-backed creation/continuity selection | PASS, two rows. |
| `20260922T134955Z-p52804` | New activation and corrected blank payload frontend rows against original production handlers | Expected FAIL: pressing produced a create (two fetches instead of query only); click-only produced none (one fetch instead of two). |
| `20260922T135248Z-p55561` | `make generate` after authored catalog edits | PASS. Updated browser batch selection and topology render index. An earlier browser preflight refused missing generated groups before executing tests; generation repaired this routing prerequisite. |
| `20260922T135305Z-p58645` | New browser activation/cancellation rows against original production handlers | Expected FAIL, both at first mouse-down: one POST instead of zero, including raw collection authoring. |
| `20260922T135535Z-p91870` | Corrected Timeline frontend slice below | PASS, three rows, 4/4 work units. |
| `20260922T135628Z-p24792` | Workbook regression slice below | PASS, six rows, 7/7 work units. |
| `20260922T135547Z-p92875` | Ten-row Timeline browser slice below | Seven scenarios passed; three test assumptions failed. Native activation, cancellation, rejection, captured replay, ordinary create and existing uncertainty/continuity passed. |
| `20260922T140233Z-p33535` | Rerun continuity, controls, explicit blank browser rows | Controls and blank PASS. An additional pending scalar replacement probe failed; baseline investigation below. |
| `20260922T140637Z-p68549` | Pending scalar replacement probe with corrected button | FAIL: a second fill while create was held reverted to the first captured text. |
| `20260922T140853Z-p2113` | Same replacement probe with original HEAD button restored temporarily | Same failure, now asserted immediately after the second fill, before any button interaction. Confirmed pre-existing and outside the activation adapter; corrected component restored afterward. |
| `20260922T141104Z-p36239` | Final pending-create/continuation browser row | PASS, 11/11 work units. Captured text and one-create admission survive attempts while disabled; newer account-menu focus intent wins after accepted explicit creation. |
| `20260922T141136Z-p66427` | `make agent-finalize` before broader final checks | PASS. JSON shape, catalog/tier coverage and generated drift passed; generated files unchanged. Retained evidence/scheduler/performance maintenance skipped because `RESULTS_DIR` was unset. Detailed artifact: `unit-artifacts/finalize-summary.json`. |

The ten-row run exposed three fixture assumptions: the synopsis draft needed the
semantic virtualized-grid scroll helper; toolbar Add row focuses the first
create-writable visible field (Date Entered here); and saved empty synopsis is
displayed as `—`. The newly catalog-routed existing blank scenario already asserted
that display later, so its wait now uses the same expected value. Measurement
predicates and budgets were not altered.

The additional pending scalar replacement probe confirmed a separate existing
authoring defect: after holding the first create and filling synopsis a second
time, the input reverted to its original captured text even before any button
interaction. This happened with the original HEAD component as well as the
correction. It does not establish a button regression. The final activation test
retains the required already-saving case: Create is disabled, DOM and pointer
attempts do not alter its captured text or add another logical create, and the
accepted server value is unchanged. The existing owner-selected changed-authoring
autosave regression also passes. No mutation or editor owner was weakened to
accommodate this newly identified issue.

### Exact final slice commands

```bash
make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.draft_create_activation,module.timeline.frontend.the_workbook_blank_row_action_submits_exactly_on_ad8dcf13e4,module.timeline.frontend.the_workbook_create_payload_builder_omits_zero_f_deaef46bf2

make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_deferred_continuity,web.workbook.regression.timeline_row_menu_lifetime,web.workbook.regression.workbookshell__autosave_suppresses_duplicate_pen_293b7b0803,web.workbook.regression.workbookshell__autosave_queues_a_follow_up_scala_3b8fa7ce27,web.workbook.regression.workbookshell__grid_preserves_an_in_flight_draft_d12ffd720f,web.workbook.regression.workbookshell__grid_preserves_draft_row_edits_ac_cd490037ce

make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.draft_create_activation,module.timeline.browser.draft_create_cancellation,module.timeline.browser.draft_create_continuity,module.timeline.browser.draft_create_controls,module.timeline.browser.draft_create_rejection,module.timeline.browser.draft_create_uncertainty,module.timeline.browser.explicit_blank_create,module.timeline.browser.a_user_creates_a_timeline_row_on_the_real_workbo_62ae8a3a07,module.timeline.browser.spreadsheet_uncertain_create,module.timeline.browser.deferred_continuity

make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.draft_create_continuity,module.timeline.browser.draft_create_controls,module.timeline.browser.explicit_blank_create

make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.draft_create_continuity
```

These Chromium cases exercise the production grid and real PostgreSQL/object-store
backed application. The browser-native cancellation case uses a CDP touch start
and cancel, observes `pointercancel`, and verifies no synthesized activation or
input blur. Test observation attributes record native events; they are not new
product selectors. Request holds separate activation from acceptance and allow
same-task and later pending activation checks.

## Changed files and compatibility

- `apps/web/src/workbook/timeline/components/TimelineDraftRowActions.tsx`:
  sole production change, click-only command dispatch plus non-submitting
  mouse-down focus suppression and native-key propagation isolation.
- `apps/web/src/workbook/WorkbookShell.payload.test.tsx`: replace incidental
  mouse-down submissions with same-frame and pending clicks; reject presses,
  cancellation and auxiliary clicks without cancelling native keyboard defaults.
- `apps/web/e2e/timeline-workbook.spec.ts`: six focused browser scenarios;
  correct the existing empty display wait without changing its payload guarantee.
- `tools/test_families/module.timeline.json`: one frontend and seven browser
  catalog rows, including the existing explicit-blank scenario.
- `tools/browser_e2e_batch_manifest.json` and
  `tools/execution_topology_render_index.json`: generated routing only.
- This handoff: advisory dispositions, authority, characterization and results.

No production mutation, authoring, query, authorization, menu, continuity,
performance or shared Grid Adapter owner changed. No manual keyboard submission
or mouse-down submission remains on Create. No new API, pending state,
dependency or framework was introduced.

## Limits and rollback

DOM click-only coverage proves event-path support, not complete screen-reader
compatibility. No manual screen-reader, browser/assistive-technology matrix or
visual focus-ring review was performed. Automated focus, accessible-name, native
Enter/Space and pending-state assertions cover this change's applicable behavior;
they do not constitute a new accessibility conformance claim.

No performance improvement is claimed. Existing capture performance marks,
predicates and budgets are unchanged; this leaf activation correction does not
change shared capture behavior, so measurement and visual suites were not rerun.
Unrelated backend/release/full-check suites are outside this bounded source change.

Known pre-existing limitation: replacing scalar text during a held initial fast
capture can revert the replacement before any Create activation. Reproduce by
opening an empty Timeline, revealing Activity Synopsis, holding its create POST,
filling `Incomplete fact Ω`, waiting for the request, then filling
`Incomplete fact Ω with newer text`. The second value reverted in both baseline
and corrected runs listed above. This needs a separately scoped authoring-owner
investigation; the approved production change is confined to the button. The
activation cleanup does not claim to repair all pending scalar authoring behavior.

Rollback consists of reverting this component correction, its corresponding test
and authored catalog changes, and regenerating downstream routing. No persisted
data or schema rollback is needed. No commit, push or deployment was performed.

## Acceptance assessment

Assessments apply to this activation slice. A PASS below does not assert that an
untouched subsystem's entire requirement family was exhaustively reverified. The
separate pre-existing scalar replacement defect is explicitly retained above;
it is not classified as a passing authoring behavior.

| ID | Assessment | Evidence or scope rationale |
| --- | --- | --- |
| A001 | PASS | Exact Core/design requirements and independent source/verification ownership are mapped above; advisory guidance supplied no new behavioral requirement. |
| A002 | PASS | One leaf activation decision; selection rubric identifies correction, retention, retirement, risk and rollback. No general abstraction or duplicate state. |
| A003 | PASS | Branch/HEAD/clean baseline, current manifests, source guides, ownership, import boundary and generated routing inspected. Digest is dated; current source and catalog are controlling. Final dirty scope is only the seven listed files. |
| A004 | PASS | Production diff introduces no design literal, token or theme/density registry. Existing styles are byte-for-byte retained. |
| A005 | N/A | No theme selection, palette, styling or theme fixture changed. |
| A006 | N/A | No density, geometry, padding, typography or editor layout change; density/measurement suites are outside this event-only correction. |
| A007 | PASS | Active view declares zero-field create; payload tests and real browser blank, raw collection, fast capture, viewer and closed cases pass. Add row focuses without creating; paperclip opens its independent chooser. Other producers/contextual create routes are untouched. |
| A008 | N/A | Responsive thresholds, viewport fallbacks and inspector geometry are unchanged. |
| A009 | N/A | Scroll/overflow ownership and shell layout are unchanged; account-menu focus remains reachable in the scoped continuation case. |
| A010 | N/A | No inspector dispatcher, feature route, review lifetime or attachment lifetime change. |
| A011 | PASS | Production-grid fresh-draft focus, one trailing draft, raw collection preservation and newer-intent focus yielding pass. Existing deferred continuity and row-menu lifetime rows also pass. Other producer/source transitions are untouched. |
| A012 | PASS | Same-frame and later pending activation admit one create; blank payload has only `client_txn_id`; four independent accepted actions use distinct IDs. Lost response replays identical bytes/transaction and returns the same record/change-set receipt. Web Crypto identity owner unchanged. |
| A013 | PASS | Definitive invalid raw create retains authoring and Discard, without re-key Retry; uncertainty retains captured request and one logical record. Existing uncertain-create browser row passes. General queue conflict/acknowledged-refresh mechanisms are unchanged. |
| A014 | PASS | Within activation scope, unsent raw collection text survives cancelled/auxiliary presses and completed create; captured scalar text survives attempts while saving; native key activation and continuation pass. The separate replacement-before-activation defect remains an explicit limitation, not a passing claim. Existing changed-authoring autosave regression passes. |
| A015 | N/A | No saved-cell conflict locus or conflict-value presentation changes. Draft terminal recovery is assessed under A013. |
| A016 | PASS | Applicable interaction admission passes for editable, pending, viewer and closed states. Query-data state production and its broader matrix are unchanged. |
| A017 | N/A | No refresh, revocation, reauthentication, account replacement or retained authorization-scope implementation changes. Viewer/closed Create admission is covered under A016. |
| A018 | N/A | No Evidence lifecycle/overlay/preview changes. Independent paperclip chooser behavior is exercised under A007. |
| A019 | PASS | Applicable button semantics, accessible name, native Enter/Space, click-only event support, no Space-down submission and focused fresh continuation pass through the production grid. Styling/live regions/motion remain unchanged; manual assistive-technology limits are stated. |
| A020 | PASS | Enabled, pending-disabled and absent read-only/closed action states behave deterministically. Styles/compound visual precedence are unchanged; no new density/zoom/overflow claim. |
| A021 | PASS | Real virtualized Grid Adapter is used; tests reveal fields with the semantic scroll helper and assert saved/draft identities and focus. No virtualization algorithm, result size or shared capture-performance predicate changed. |
| A022 | N/A | No renderer geometry, color, crop, viewport or golden change; visual fixtures would not independently prove native activation. |
| A023 | PASS | Existing UI-contract selectors, roles/names, field keys, record IDs and owner states identify controls. Native event observation attributes are test-only observations. No selector facade changed, so a package-wide selector slice was unnecessary. |
| A024 | PASS | Changed test/catalog inputs depend on production contracts and existing executable fixtures only; no executable consumer of docs/Markdown was added. Offline advisory query and human review are separate from product verification. |
| A025 | PASS | Authored Timeline catalog changed first; Make generation produced only browser batch manifest and topology render index. Finalizer catalog/shape/tier and generated-drift checks pass. |
| A026 | PASS | Accessible name, native type, disabled projection, props, selectors, payload and existing command ownership preserved. No API/schema/storage/lifecycle/migration change. Removed incidental event paths and rollback are explicit. |
| A027 | PASS | This handoff records scope, exact owners, characterization, commands/artifacts, failures and dispositions, limits, acceptance and rollback. No commit/push/deploy. Follow-up is the separately scoped pre-existing scalar replacement issue, not a hidden completion claim. |

## Final checks

After `make agent-finalize`:

| Command | Result | Artifact |
| --- | --- | --- |
| `make frontend-typecheck` | PASS, 2/2 work units | `.cartulary/test-results/20260922T141231Z-p73109/run-summary.json` |
| `make lint-biome` | PASS, 2/2 work units | `.cartulary/test-results/20260922T141231Z-p73131/run-summary.json` |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260922T141407Z-p74710/adhoc/lint-markdown/tool-run-summary.json` |
| `git diff --check` | PASS | Final working diff, no whitespace errors. |

`make format-frontend` also passed during implementation; its final run summary is
`.cartulary/test-results/20260922T141036Z-p35686/format-frontend/tool-run-summary.json`.

The final source selections total three Timeline frontend rows, six workbook
regression rows and ten browser scenarios with passing evidence across targeted
runs. Failed characterization/probe runs remain documented and are not reported
as successful runs. Retained-run maintenance was skipped with `RESULTS_DIR` unset.


## Subsequent resolution of pending-capture authoring

The separate scalar replacement limitation recorded above is addressed by the
[Timeline pending-capture authoring remediation](workbook-timeline-pending-capture-authoring-remediation-handoff.md).
That slice retains the exact replacement-without-keydown reproduction, moves
capture identity and promotion into the retained owner, and verifies revision-owned
successors, detachment and native input. This activation handoff remains historical
evidence; consult the linked record for current verification and manual limits.
