# Account Settings Editing Refactor Handoff

AS-01–AS-05 and their evidence below are preserved as completed history, including
their original baseline and end-of-iteration working-tree statements. The
[next-iteration plan](#next-iteration--account-frontend-cleanup-and-production-hardening)
records the current inspection baseline and future work separately.

## Baseline and scope

- Branch `main`; HEAD `77868f53dad19dfec308fe022be7f2b88bb0df6f`; clean at entry.
- Root AGENTS.md is the only applicable instruction file. No reset or unrelated edits.
- Scope: Profile/Appearance editing ownership, minimal App/session integration,
  directly dependent adapters, selectors, routed tests, visual evidence, and this handoff.
- React 19 / TypeScript 6 / Vite 8 / pnpm 10.33; no dependency changes.
- Session/bootstrap, directory/creation, account menu, Security, density geometry,
  save-status, presence, and workbook recovery remain regression baselines.

## Tracker

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| AS-01 — Baseline, ownership, characterization | DONE | Four failures reproduced; disposal guard already passes. |
| AS-02 — Resource/draft/attempt lifecycle | DONE | Twelve lifecycle scenarios and existing session tests pass. |
| AS-03 — Integration and presentation | DONE | Native forms/radios, retention, guarded publication, typecheck and imports pass. |
| AS-04 — Browser, accessibility, visual | DONE | Browser/a11y regressions and two fresh visual passes; all new images reviewed. |
| AS-05 — Final validation and handoff | DONE | Acceptance, validation, review, limitations, rollback and scope recorded. |

Only the current row is IN_PROGRESS. Save its terminal evidence before advancing.
An adopted-owner contradiction means BLOCKED: owner contradiction.

## Owner-to-change map

| Behavior | Owner | Implementation / verification |
| --- | --- | --- |
| Current-account fields, versions, replay | Core 01 §3.3.2.3, REQ-01-597–602 | Account panels/controller; web.application and module.auth |
| Display-name normalization | Core 01 §18, REQ-01-489; Core 02 REQ-02-255 | Profile validation and HTTP boundary |
| Deployment-local account state | Core 02 REQ-02-255 | Existing resources; no storage or API change |
| Actor and session lifetime | Core 04 §1.1.1 and REQ-04-114 | Existing AppSessionController; web.application |
| Composition, density, feedback, keyboard | Design §§3.9, 4.4, 12, 14 | Panels/styles; web.design and package.ui |
| Full workbook density | Core 01 density defaults; design §3.9 | Existing pipeline; module.workbook browser row |

Digest README and localized read order, rules, acceptance and owner maps were
read during planning. Domain owns vocabulary/navigation; design owns its bounded
presentation scope. Completed incident-directory and account-menu handoffs are
implementation evidence. No inspected owner contradiction.

## Inspection and compatibility decisions

Inspected AccountSettingsPanels, App, authAccountClient, publicHttpTypes,
appSessionController, useAppSession, landingAdminStyles, App.landing tests,
shellHttpClients route tests, accountSettings browser helper, account/density/a11y/
visual fixtures, source ownership, catalog owners and relevant test families.

Each panel has one production caller in App. Appearance's uncontrolled mode has
no production caller; remove it with the old callback interfaces. Each account
mutation adapter has one production caller; the operation owner will supply its ID.
Profile and Appearance retain independent saves. No generic form framework.

User decisions: section switching and close/reopen retain in-memory drafts and
attempts within the current lifetime. Discard resets the editable draft only;
pending/uncertain attempts remain recoverable and block fresh writes. Every
retired application lifetime clears settings state, including same-account
reauthentication. Ordinary refresh preserves it. No browser storage.

Initial source observations: panels use message strings for operational state; profile
completion replaces newer input; Appearance treats every 409 as a refresh and
overwrites its draft; neither panel guards overlapping submissions. These are
inspection hypotheses, not claims of reproduced failures or cross-account leaks.

## Verification and work log

Planning resolved `make help`, `make help-all`, and task guides for web.application,
module.auth, web.design and package.ui. Source owner web.app is distinct from
verification owner web.application. Full-density browser evidence routes to
module.workbook, not an inferred filename owner.

Selected baseline `make test-slice OWNER=web.application ROWS=...` passed three
cached rows (account entry, density publication, session lifetime) at
`.cartulary/test-results/20260906T031055Z-p77004`. The module.auth HTTP-boundary
and account-support slice passed two cached rows at
`.cartulary/test-results/20260906T031055Z-p77005`. These are regression evidence,
not new race characterization. `git diff --check` passed.

AS-01 completed before production edits. New characterization lives in
`apps/web/src/app/accountSettingsEditing.test.tsx`, registered under web.app and
`web.application.regression.account_settings_preserve_drafts_and_guard_submi_e245fc5a0c`.
Generation passed at `20260906T032301Z-p84857` after correcting an unsorted
authored title list (initial generation `20260906T032148Z-p81252` and the first
slice admission failed on that harness input).

The characterization slice failed as expected at `20260906T032401Z-p88847`:
duplicate dispatch sends two writes, save completion replaces newer text,
section switching loses the draft, and a version-conflict refresh replaces the
Appearance draft. Four failures are product assertions against unchanged source.
The disposal scenario already passes because the session owner rejects work
after disposal; no cross-account leak is claimed. An initial conflict test passed
too early; its refreshed read is now gated and awaited before asserting retention.
The initial run is `20260906T032315Z-p88079` (three failures, two passes).

| State / event | Saved resource | Draft / attempt | Recovery |
| --- | --- | --- | --- |
| Initial load failure | Unavailable | No writable base | Retry load |
| Refresh / refresh failure | Retain usable accepted value | Preserve draft and reviewed base | Retry refresh |
| Saving / newer edit | Retain accepted value | Immutable attempt; newer revision stays editable | Await result |
| Version conflict | Read separately; never regress version | Retain draft; review required | Use saved or review edit |
| Transaction conflict | No automatic write | Retain draft; distinct rejection | Explicit new-attempt review |
| Uncertain / malformed success | No unvalidated publication | Retain exact attempt | Replay exact request |
| Confirmed / propagation failure | Mutation remains confirmed | Acknowledge captured revision only | Retry publication read only |
| Close / section switch | Retain in memory | Retain work | Reopen |
| Lifetime retirement | Clear protected resources | Clear drafts, attempts, callbacks | Existing authentication owner |

Next: begin AS-02; no owner contradiction or blocked dependency.

AS-02 completed. Added the bounded accountSettingsModel controller and twelve
deterministic lifecycle tests, with independent Profile/Appearance dispatch locks,
immutable attempts, 30-second deadlines, exact replay, conflict review, monotonic
resource acceptance, revision acknowledgement, explicit discard, and lifetime
clearing. Appearance's saved snapshot is a projection of the session owner.
The session acceptance boundary now rejects regressing/invalid preferences and
retains usable preferences through refresh/failure. Profile publication uses a
guarded, typed session observation. Account mutation adapters require captured
transaction IDs and accept observation signals; Security adapters are unchanged.

`make generate` passed at `20260906T033200Z-p91230`.
`make test-slice OWNER=web.application ROWS=web.application.regression.account_edit_resource_draft_attempt_lifecycle_9420b2e411,web.application.regression.app_session_lifecycle_d770146af2`
passed at `20260906T033220Z-p94474` (three graph units). Tests cover both resource
kinds, rejection categories, exact replay, ignored-abort timeout, stale versions,
newer edits, account/logout/reauthentication clearing, density modes, display-name
normalization and publication-only recovery. Production panel migration and full
typecheck belong to AS-03; the old panels have not been given a compatibility shim.
Next: AS-03 integration and presentation.

AS-03 completed. App owns the settings controller above panel mounts; existing
session retirement/disposal clears it. Lifetime-bound panel commands reject old
callbacks. Profile and Appearance are required bindings, with no uncontrolled
Appearance path or panel-owned HTTP effects. Native forms preserve focus, expose
field-associated display-name validation and explicit Save, show saved/draft values,
and provide local load, conflict, replay, discard and publication recovery. Density
is the required native segmented radiogroup; null remains surface default. The
existing settings dialog, menu, Security callbacks and workbook geometry remain.

Updated accountSettings browser helper, full-density browser scenario and long-label
fixtures to use semantic radios and field readiness instead of Save readiness.
No hidden select, new dependency, browser storage or shared design registry.

Passing evidence under `.cartulary/test-results/`:

- `make format`: `20260906T033949Z-p16895`.
- `make frontend-typecheck`: `20260906T033959Z-p21417`.
- `make frontend-import-boundary-check`: `20260906T033959Z-p21422`.
- Application/controller/characterization/density/session slices:
  `20260906T033607Z-p979`, `20260906T033859Z-p14495`.
- Module.auth HTTP-boundary and account-support slice: `20260906T033859Z-p14512`.

Resolved intermediate failures: typecheck `20260906T033606Z-p605` rejected
Testing Library's unsupported `exact` option and an untyped mock; format
`20260906T033803Z-p2517` identified effect dependency syntax; import boundary
`20260906T033859Z-p14643` required protocol resource validation to live in the
existing authAccountClient adapter layer. No checks or boundaries were weakened.
Next: AS-04 real-browser conflict/replay, keyboard, accessibility and visual review.

AS-04 evidence in progress (run IDs below are under `.cartulary/test-results/`):

- Added authored web.design browser rows
  `web.design.browser.account_settings_real_conflict_replay_46b2ec9831` and
  `web.design.browser.account_settings_keyboard_responsive_781bd74a33`.
  `make service-backed-test-slice OWNER=web.design ROWS=...` passed both at
  `20260906T034613Z-p31629`. The first commits against the real isolated backend,
  loses the response, retains a newer edit across section/dismissal transitions,
  and verifies byte-identical replay plus a single server version increment.
  The second uses deterministic HTTP faults for Appearance conflict/replay and
  native radio keyboard behavior at constrained widths and heights.
- Added accessibility row
  `web.design.accessibility.account_settings_forms_radios_recovery_71b288da44`;
  its service-backed slice passed at `20260906T034613Z-p31641`. Evidence includes
  accessible names, field feedback association, focus, contrast, native radio
  arrow/wrap behavior, keyboard recovery, Escape restoration and reduced motion.
- Existing full-density browser row
  `module.workbook.browser.verify_continuous_workbook_shell_composition_top_96ea2f5084`
  passed at `20260906T034721Z-p26624`, including null and every explicit mode.
- Existing Security and non-admin account-navigation module.auth browser rows
  passed at `20260906T035509Z-p54410`; no Security implementation changes.
- `make generate` passed at `20260906T034524Z-p24000`;
  `make frontend-typecheck` passed at `20260906T034613Z-p31752`.
  Latest `make format` passed at `20260906T034938Z-p18083`.
- Controller/component/session tests passed at `20260906T035509Z-p54377`
  after adding same-version timestamp validation and accepting a successful
  session publication that already contains a later external display-name edit.
  Ordinary session refresh preserves settings drafts; retired-lifetime bindings
  cannot submit, publish, or initiate publication recovery.
- `make test-slice OWNER=package.ui` passed at `20260906T035622Z-p17568`.
- `env -u RESULTS_DIR make agent-finalize` passed at
  `20260906T034645Z-p18745`, before broader verification. Retained-run maintenance
  was skipped because RESULTS_DIR was unset; no full warm check evidence claimed.

Ordinary `make browser-e2e-visual` at `20260906T034722Z-p26740` failed solely
because the twelve new captures have no committed goldens yet. All functional
capture assertions and existing visual rows passed. The retained
`browser-e2e-visual/frontend-visual-reconciliation.json` accounts for 122 active
capture intents, 110 existing goldens, 26 registered fixtures, zero orphans and
zero ambiguous mappings. The twelve missing captures all resolve to the new
authored row `web.design.visual.account_settings_editing_states_5af731dc29`;
they are new active captures, not lost or unexplained existing goldens. Their
explicit creation is the intended update. Existing 110 goldens remain unchanged.
No D-VFIX identity was invented; these captures use exact nonregistry row/scenario
accounting. Next: inspect the generated images and obtain two ordinary passes.

Visual review of update `20260906T035456Z-p24003` (12/12 graph units passed)
identified a panel layout defect: the inherited two-column form put recovery
feedback apart from its field and constrained the radios unnecessarily. At 200%
CSS zoom, the viewport-unit dialog bounds clipped content. These captures are
not accepted final evidence. Profile/Appearance now use the existing one-column
form grid. The existing dialog uses its available grid area for maximum size;
this is a minimal containment correction, without changing navigation, dismissal,
Security content or focus policy. Disabled Save uses existing muted/secondary
tokens. Browser assertions now check the whole dialog and radio group bounds,
plus focus reachability of enabled Save, Refresh, Close and wrapped radio choices.

The update also rewrote six unrelated existing PNGs despite their passing ordinary
comparisons. They were restored to HEAD (presence markers, timeline evidence badge,
creation expanded-details/zoom, Network Analysis inspector, pending replay status).
`make generate` rebuilt their manifest through the authored generator at
`20260906T040024Z-p15108`; no generated manifest was hand-edited.

Full `make browser-e2e-a11y` passed at `20260906T035623Z-p17862`, and the existing
account-menu keyboard browser row passed at `20260906T035622Z-p17556`.
The strengthened zoom browser test at `20260906T040025Z-p15751` exposed a fixture
error: it tried to focus disabled Save without first editing. The scenario now
selects a different density before checking Save reachability. Typecheck passed
at `20260906T040025Z-p15834`.

An additional deferred characterization at `20260906T040153Z-p11388` reproduced
an obsolete Appearance replay incorrectly clearing the local refreshing indicator
while the canonical session owner kept a newer read in flight. The controller now
preserves the session-owned read state, and the test checks both the in-flight
state and the later accepted resource. No request was duplicated or resource
version regressed. `make format` passed at `20260906T040219Z-p12117`.

The corrected controller/component slice passed at `20260906T040234Z-p16419`,
and the strengthened responsive browser scenario passed at
`20260906T040234Z-p16424`. The second ordinary visual run,
`20260906T040050Z-p62203`, failed only on the twelve intended layout differences;
28 existing workbook visual tests and the Network Analysis visual group passed.
Its reconciliation has 122 active goldens/captures, with no missing, orphan or
ambiguous entries. Reviewed actual images confirm the single-column form,
distinct saved/draft values, muted disabled Save, local conflict/replay controls,
unclipped dialog at zoom and vertical panel scrolling with reachable controls.

Update `20260906T040400Z-p66321` stopped with `artifact_error`: restored unrelated
PNG files had mode 0644, while retained harness artifacts require owner-only mode.
It did not promote the candidate or manifest. Restored those six source PNGs to
0600 without changing bytes. No harness policy was changed. The text-spacing
capture is now 640x640 to include wrapped choices; the long-name capture remains
768x640. This fixture-only viewport adjustment is an intentional refresh trigger.

`env -u RESULTS_DIR make agent-finalize` passed again at
`20260906T040322Z-p62237`, retaining the same maintenance skip. Broader checks:

- Full `make test-slice OWNER=web.application`: 64/64 graph units passed at
  `20260906T040414Z-p1131`.
- `make frontend-typecheck`: `20260906T040414Z-p1377`, pass.
- `make frontend-import-boundary-check`: `20260906T040414Z-p1401`, pass.
- `make lint-biome`: `20260906T040414Z-p1433`, pass.
- Module.auth HTTP-boundary/account-support slice:
  `20260906T040534Z-p21378`, pass.

Final-source browser regressions pass: full accessibility at
`20260906T040534Z-p21608`; full-density at `20260906T040643Z-p11848`;
Security/non-admin account navigation at `20260906T040643Z-p11847`;
real backend conflict/replay and account-menu keyboard at
`20260906T040759Z-p6307`.

Final visual update passed at `20260906T040610Z-p64196` (12/12 graph units,
29 workbook visual scenarios plus the claimed Network Analysis scenario).
It promoted a complete reconciled candidate. Three unrelated rewritten PNGs
(presence markers, creation expanded details, pending replay status) were restored
to their baseline bytes with mode 0600. `make generate` passed at
`20260906T040936Z-p51353`, regenerating the manifest. All 110 pre-existing PNGs
are byte-identical to HEAD; only twelve new account-settings PNGs remain added.

### Visual refresh record

Accepted trigger: owner-required account editing behavior and segmented radios;
reviewed single-column forms, explicit disabled Save, and corrected containment
under constrained width/zoom. This is implementation-support evidence, not a
Core 05 claim or a new design authority. Agent inspection reviewed every new
image; nine final PNGs exactly match reviewed ordinary actuals, and the final
Profile conflict/recovery plus wrapped text-spacing PNGs were separately inspected.

All captures belong to authored row
`web.design.visual.account_settings_editing_states_5af731dc29`, scenario
`scenario_4301c1db00ce`, project `chromium`. No stable registered D-VFIX ID applies
to these twelve new nonregistry captures. Existing registry entries remain exact
and unchanged. Files are under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`;
each filename below has suffix `-linux.png`.

| Filename stem | Viewport / zoom | Reviewed state / decision |
| --- | --- | --- |
| account-settings-profile-dirty | 1280x720 / 100% | Saved name, editable draft, focus, explicit Save; accept |
| account-settings-profile-pending | 1280x720 / 100% | Editable field, muted Save, pending text; accept |
| account-settings-profile-conflict | 1280x720 / 100% | Saved/draft distinction and explicit review choices; accept |
| account-settings-profile-recovery | 1280x720 / 100% | Uncertainty, exact replay action, discard meaning; accept |
| account-settings-appearance-dirty | 1280x720 / 100% | Ordered named radios, checked/focus cues, saved density; accept |
| account-settings-appearance-pending | 1280x720 / 100% | Retained selection and pending action; accept |
| account-settings-appearance-conflict | 1280x720 / 100% | Current Default versus Compact draft, local review; accept |
| account-settings-appearance-recovery | 1280x720 / 100% | Exact replay and retained selection; accept |
| account-settings-appearance-short | 640x480 / 100% | Header/tabs remain visible; body scrolls; accept |
| account-settings-appearance-zoom | 1280x720 / 200% CSS zoom | Dialog fits, named choices visible, body scrolls; accept |
| account-settings-appearance-text-spacing | 640x640 / 100% | Increased spacing, wrapped Comfortable segment; accept |
| account-settings-profile-long-name | 768x640 / 100% | 252-scalar unbroken draft scrolls inside native input; accept |

Classification for each final image: accepted intentional state, under design
§§4.4, 12 and 14. Theme is the existing dark_graphite; density is the fixture's
declared saved value and never an unsaved global preview. Viewport captures use
the existing font-readiness, renderer, scroll-normalization and focus/caret
contract. No new masks, selector crops, screenshot tolerance, browser pins, or
fixture identities. CSS zoom follows the existing suite's emulation convention.
Two fresh ordinary `make browser-e2e-visual` runs passed against the final manifest:
`20260906T041003Z-p54792` and `20260906T041003Z-p54785`, each 12/12 graph units.
Both reconcile 122 active captures/goldens and 26 registered fixtures, with zero
missing, orphan, ambiguous or unresolved entries. Browser execution was fresh;
these are not cached visual assertions.

AS-04 complete. Expanded the existing deterministic model cases to cover
Appearance initial failure/retry, dirty successful refresh, newer editing during
a pending Appearance save, malformed Appearance success, and failed preference
refresh after confirmed publication with no duplicate mutation. The final focused
controller/component slice passes at `20260906T041342Z-p59872`. Typecheck and
Biome pass at `20260906T041342Z-p59948` / `20260906T041342Z-p59969`; format passes
at `20260906T041322Z-p55477`. Runtime and visual fixtures were unchanged by this
last test expansion.

JSON shape, generated policy and drift checks passed respectively at
`20260906T041013Z-p3537`, `20260906T041013Z-p3519`, and
`20260906T041013Z-p3498`. Final import-boundary verification passed at
`20260906T041013Z-p4010`. Next: AS-05 final handoff and scope audit.

## Compatibility and rollback

Frontend interaction remediation on existing contracts; no API or stored-data
migration. Production controller/panels/session integration, dependent tests and
selectors, generated routing, and reviewed visual goldens form one atomic rollback
boundary. Do not commit, push, or begin another seam.

## Final behavior and ownership

Before, panel-local state mixed saved values and editable input, generated fresh
transaction IDs per invocation, discarded work on unmount, and interpreted
Appearance's entire 409 class as a version conflict. Four concrete product
failures were reproduced before remediation; a cross-account leak was not.

After, App owns one bounded account-settings controller above section/dialog
mounts. Profile's accepted resource belongs there; accepted preferences remain
canonical in AppSessionController. Each resource independently owns its draft,
edit revision, reviewed base version, read outcome, operation and immutable
captured attempt. UI feedback derives from typed state. Profile and Appearance
can save independently; duplicate dispatch and duplicate replay are blocked at
the synchronous operation boundary.

An uncertain request retains its original actor/lifetime, kind, desired value,
version and ID. Retry save observes that exact request again; neither timeout nor
discard asserts server cancellation. A definitive version conflict refreshes the
resource separately and preserves the draft. Review my edit explicitly accepts
the displayed saved version as the next base; Save then creates a new request.
Use saved value replaces the draft. Transaction conflicts, validation rejection,
authorization denial, malformed success and session loss have distinct outcomes.

Confirmed saves acknowledge only the captured revision. Later edits stay intact.
Older reads/replays cannot replace newer accepted resources or retire a newer
preferences read. Confirmed Profile saves use guarded session observation for
current labels/presence; publication failure offers read-only Retry refresh and
does not repeat the write. Confirmed Appearance saves pass validated resources
through the existing session acceptance boundary. Failed preference refresh
retains usable density. No historical display-name snapshots are rewritten.

Section switching, Close and Escape retain in-memory work during the current
lifetime. Reopen refreshes saved resources while keeping drafts and reviewed bases.
Discard changes the draft only while a request is pending/uncertain. Every retired
application lifetime clears protected settings state, timers and callbacks,
including same-account reauthentication. Ordinary session refresh does not retire
the lifetime. The existing session controller remains the only authentication and
bootstrap owner.

## Final changed-path inventory

All paths are relative to the repository root. Final scope is 34 paths:

- Runtime: `apps/web/src/app/accountSettingsModel.ts`,
  `apps/web/src/app/AccountSettingsPanels.tsx`, `apps/web/src/app/App.tsx`,
  `apps/web/src/app/appSessionController.ts`,
  `apps/web/src/app/api/authAccountClient.ts`,
  `apps/web/src/app/landingAdminStyles.ts`.
- Deterministic tests: `apps/web/src/app/accountSettingsModel.test.ts`,
  `apps/web/src/app/accountSettingsEditing.test.tsx`,
  `apps/web/src/app/App.landing.test.tsx`,
  `apps/web/src/app/api/shellHttpClients.routeBoundary.test.ts`.
- Browser support/evidence: `apps/web/e2e/pages/accountSettings.ts`,
  `apps/web/e2e/support/auth/accountEditingFixture.ts`,
  `apps/web/e2e/incident-administration.spec.ts`,
  `apps/web/e2e/workbook.a11y.spec.ts`, `apps/web/e2e/workbook.visual.spec.ts`.
- Authored verification inputs: `tools/frontend_source_ownership.json`,
  `tools/test_families/web.application.json`, `tools/test_families/web.design.json`.
- Make-generated projections: `tools/browser_e2e_batch_manifest.json`,
  `tools/execution_topology_render_index.json`,
  `tools/frontend_visual_golden_manifest.json`.
- The twelve new PNGs listed in the visual record, and this handoff.

No backend, public schema/type shape, API route, dependencies, lockfiles, browser
storage, generated runtime code, design registry or digest files changed.
The existing Appearance test ID now identifies the semantic radiogroup; helper
callers select radios by accessible name. No hidden select or compatibility mode
remains. Unrelated select controls elsewhere keep their existing semantics.

## Acceptance dispositions

PASS below is scoped to this seam and cited regression evidence. N/A identifies
an unchanged workstream; it does not claim new verification of its entire owner.
These advisory IDs remain human review references only, with no executable link
to Markdown or digest content.

| ID | Disposition | Evidence / boundary |
| --- | --- | --- |
| A001 | PASS | Adopted owner-to-change map; digest never became authority. |
| A002 | PASS | Profile/Appearance only; model ownership and behavior corrections are separately recorded. |
| A003 | PASS | Actual branch/HEAD/clean baseline, stack, caller inventory, sources and routing inspected; no grid imports added. |
| A004 | PASS | Existing tokens/styles only; no second theme, typography or density registry. |
| A005 | PASS | Existing dark_graphite visual suite; no theme behavior change. |
| A006 | PASS, regression | Null/all explicit modes, no draft preview, retained preferences on refresh failure, full-density browser row. |
| A007 | N/A | Incident/workbook creation owners unchanged; existing visual baseline retained. |
| A008 | PASS, scoped | Whole-dialog bounds and control reachability at 1280, 1024, 768, 640 and 480 widths; zoom/text spacing; inspector geometry unchanged. |
| A009 | PASS, scoped | Account navigation and Close remain reachable; settings body scrolls; workbook overflow owner unchanged. |
| A010 | N/A | Inspector dispatcher, features and authorization restrictions unchanged. |
| A011 | PASS, scoped | Native field/radio identity and focus survive refresh/completion; account-menu focus regressions pass. |
| A012 | PASS | Existing Web Crypto ID helper; immutable attempt and byte-identical real backend replay. |
| A013 | N/A | Workbook pending queue/recovery excluded; settings use their own exact replay and draft-only discard. |
| A014 | PASS, scoped | Native Enter/Tab/arrows/Space, validation, retained drafts and existing Escape dismissal; workbook paste/edit model unchanged. |
| A015 | PASS, adapted locus | Profile/Appearance panel shows saved value separately from retained draft and offers local review/replay. |
| A016 | PASS, scoped | Typed initial/refresh failures, validation/transaction/version/auth outcomes, uncertainty and confirmed/publication outcomes. |
| A017 | PASS, adapted resource | Usable saved profile/preferences survive background refresh/failure; directory rows unchanged. |
| A018 | N/A | Evidence lifecycle/overlay/preview unchanged. |
| A019 | PASS | Full a11y suite plus native form/radio/recovery browser row, names, focus, contrast, reduced motion and single local announcement. |
| A020 | PASS, scoped | Twelve deterministic dirty/pending/conflict/recovery/responsive captures and controller/component compound states. |
| A021 | N/A | Workbook virtualization/performance mechanisms unchanged. |
| A022 | PASS | Two fresh ordinary full visual runs; all existing registered fixtures reconcile; new nonregistry row identities exact. |
| A023 | PASS | Semantic roles/names and existing account selector contract; package.ui owner slice passes. |
| A024 | PASS | No production/test/generator/evidence dependency on docs or digest added; documentation is handoff only. |
| A025 | PASS | Authored catalogs/source ownership feed Make generation; JSON shape, generated policy and drift pass. |
| A026 | PASS | Existing HTTP/string/session contracts; no new auth policy, storage or workbook recovery behavior. |
| A027 | PASS | Ownership, paths, commands, failures, review, exclusions and rollback recorded here. |

Advisory classifications retained: ADOPT R001–R008 and R010–R015 where applicable
to local forms, focus, retained data, recovery, existing tokens and verification;
R009 virtualization remains an unchanged baseline. ADAPT R016–R025 and R035 to
the existing desktop/token/owner boundaries: no touch-size inflation, global CTA,
new icons, skeleton records, animation, or discard-confirmation ceremony. Native
form feedback stays local; text and radio choices wrap without altering workbook
geometry. REJECT R026–R034 as directed: no generated visual identity, marketing
layout, decorative motion, replacement behavior authority or incidental selectors.

## Limitations, exclusions and rollback procedure

Recovery is in memory only. A full reload or lifetime retirement clears settings
work; there is no persistence or automatic background replay. Closing the dialog
does not cancel an active server mutation. The UI observes for 30 seconds before
offering exact replay. A later external display-name edit may already be present
in the accepted session observation; the session owner remains authoritative.

Browser evidence uses the owned Chromium renderer. Zoom uses the existing CSS
zoom convention; no native browser-chrome zoom or separate assistive-technology
listening session is claimed. Visual review is agent inspection of actual pixels,
not separate human design approval or Core 05 publication evidence.

Retained-run maintenance was skipped because RESULTS_DIR was unset. A full warm
`make check` for this exact source was not claimed. Broad backend/release targets
(`make check`, `make ci`, `make release-check`) and `make test-evidence-audit` were
not run: this seam uses narrow routed owners and browser/design evidence, and
the complete check/support roots required by that audit are absent. No remaining
product/test failure is intentionally accepted; resolved failures are logged above.

Rollback the runtime, dependent unit/browser helpers and scenarios, authored
verification inputs, generated projections, twelve new goldens and their manifest
together. Restore the baseline behavior as one unit, then use `make generate` and
the same owner slices to check projections and behavior. Do not revert independent
session/density/Security baselines or user changes added after this work. There is
no API or stored-data migration to reverse. No commit, push or follow-on seam is
part of this work; next action is user review of this completed bounded diff.

AS-05 complete. Final branch remains `main` at
`77868f53dad19dfec308fe022be7f2b88bb0df6f`; the only dirty paths are the 34 listed
above (17 modified, 17 new). The scoped path audit and `git diff --check` passed;
all 110 existing goldens are unchanged. Final caller inventory still has one App
caller per panel. Added executable sources and changed implementation/catalog
lines contain no docs/digest dependencies, browser storage or workbook queue
imports. Excluded workstreams, dependency files and generated runtime roots are
untouched.

Final `env -u RESULTS_DIR make agent-finalize` passed at
`20260906T041438Z-p61448`; retained-run maintenance remains skipped as described
above. `make lint-markdown` passed at `20260906T041709Z-p65177` with summary
`adhoc/lint-markdown/tool-run-summary.json`. The terminal tracker edit is followed
by another Markdown lint, diff check and the same exact-path scope audit before
returning the implementation. No commit or push was made.

---

## Next iteration — Account frontend cleanup and production hardening

### Planning baseline and authority

Inspection baseline, revalidated on 2026-09-06: clean `main`, HEAD
`e444439c74781624edde15115a4c6988709baeb3`. Root AGENTS.md remains the only
applicable instruction file. No reset is planned. Revalidate branch, HEAD and
dirty state before implementation and preserve unrelated changes.

Implementation authorized on 2026-09-06 and completed on 2026-09-07.
AR-01–AR-06 executed sequentially, with terminal tracker updates between slices;
the existing staged handoff edit is preserved. AS-01–AS-05 and earlier test results remain
historical evidence, not verification of this iteration. Source observations
about races or stale state are hypotheses until characterization reproduces them.

Adopted Core owners remain authoritative: Core 01 account/authentication routes,
versions and string contracts; Core 02 deployment-local account meaning; and
Core 04 current-account, credential and session-lifetime boundaries. Use
[domain.md](../domain.md) for vocabulary and owner navigation, and
[nlspec-spec.md](../research/nlspec-spec.md) for precise interfaces, explicit
defaults, conceptual fidelity and economical documentation. The latter guides
writing; this handoff remains implementation-support planning, not an adopted
NLSpec. Design direction and the UI/UX digest retain their existing bounded or
advisory roles. An adopted-owner contradiction blocks dependent work.

### Scope and structural decisions

The planned scope is the frontend account subsystem: Profile, Appearance,
settings composition, Security, sign-in/bootstrap, deployment-user administration,
adapters and direct integration with the existing session owner. Backend internals
and public wire contracts remain unchanged. Targeted Core 01/Core 04 browser
owner amendments and design clarification precede implementation. Prefer structural simplification that makes
future phases easier to understand, test and extend. Carry behavior forward when
it has a required or material future role; remove unnecessary compatibility burden.

The inspected implementation surfaces are `AccountAdministrationPanels.tsx`,
`AccountSettingsPanels.tsx`, `accountSettingsModel.ts`, `AuthGateway.tsx`, `App.tsx`,
`appSessionController.ts`, `useAppSession.ts`, and the account/deployment-user HTTP
adapters under `apps/web/src/app/`. AR-01 must recheck their callers, exports and
state readers against its actual implementation baseline.

| Candidate | Inspection classification | Planned decision |
| --- | --- | --- |
| Both administration-panel imperative handles and `forwardRef` wrappers | Dead interfaces: callers supply no refs | Remove handles and wrappers; migrate any caller discovered during revalidation. |
| Deployment-user command-state callback, types and effect | Dead interface: no consumer found | Remove the callback and state used exclusively by the removed interfaces. |
| Target identity/version presentation and equivalent action predicates | Redundant derivation from the accepted selected user | Derive directly; preserve concurrency tokens and authorization checks. |
| Editing-model confirmed-resource field and internal-only exports | Unread field and unnecessary public surface | Remove the field and narrow exports to actual consumers. |
| Broad editor resource/value unions and internal shape discrimination | Structural duplication | Replace with resource-specific editor types and explicit transitions. |
| Enterprise-navigation test hook | Live test use, undesirable mutable global | Replace with an instance-injected navigation port and migrate tests. |
| Submission, pagination, stale-callback and propagation paths | Suspected defects, not reproduced by this planning update | Characterize before changing behavior. |

Profile publication state belongs only to Profile; accepted preferences remain
canonical in the existing session owner. Resource-specific types should remove
internal duck typing, while retaining validation of untrusted HTTP responses.
Preserve the shipped current-account recovery model and independent saves.

Separate Security, deployment-user administration and authentication models from
rendering. Extract settings presentation from App while keeping navigation and
session acceptance under their existing owners. Make boundaries feature-specific;
introduce no generic form framework, global mutation manager, compatibility facade
or new workspace package.

Make internal adapter inputs explicit: discriminated bootstrap/session TOTP
contexts, abort signals where observation needs them, and caller-owned transaction
IDs only on routes that declare them. Preserve wire shapes, response validation
and route-specific idempotency. Keep observation mechanisms separate when dispatch
timing or cancellation semantics differ; sharing an asynchronous helper is not an
objective by itself.

Current bootstrap/session modes, enterprise-profile gating, accessibility features,
recovery controls, administration visibility/loading policy and density defaults
are live-required behavior, not removal candidates. Preserve null surface defaults
and the complete existing density pipeline.

### Sequential workstreams

| Workstream | Status | Planned work and exit |
| --- | --- | --- |
| AR-01 — Ownership and characterization | DONE | Inventory callers, exports, state readers, adapters and authored verification routes. Classify candidates as dead, redundant, live-required or suspected defect. Characterize submission races, stale callbacks, pagination ordering and propagation failures before production edits; record failures separately from passing baselines and observations. |
| AR-02 — Dead interfaces and boundaries | DONE | Remove confirmed unused interfaces and derived-state duplication. Separate feature models from presentation, extract settings presentation and replace global navigation-test mutation. Exit with migrated callers, unchanged behavior and passing narrow tests. |
| AR-03 — Current-account model cleanup | DONE | Introduce resource-specific editor states and narrow session ports; remove unread state and duplicated presentation decisions. Exit with exact replay, drafts, monotonic acceptance, publication recovery and lifetime clearing preserved by focused tests. |
| AR-04 — Authentication and Security hardening | DONE | Establish explicit operation ownership, synchronous dispatch guards, typed outcomes and obsolete-completion checks. Keep bootstrap, credential changes and session establishment distinct. Fix characterized failures on existing contracts; verify secret handling and uncertainty recovery. |
| AR-05 — Deployment-user hardening | DONE | Separate query/pagination state, selected server resource, editable draft and captured operation. Fence pagination by accepted query generation; prevent duplicate actions, stale target publication and refresh-driven draft replacement. Verify confirmed mutations remain distinct from subsequent refresh failures. |
| AR-06 — Verification and closure | DONE | Owner-routed tests, browser/accessibility evidence and two fresh full ordinary visual passes are complete. Gap dispositions, deletion/ownership inventory, compatibility, rollback, current evidence and limitations are recorded below. |

During implementation, mark only the current row `IN_PROGRESS`. Complete its exit
checks, mark it `DONE` or `BLOCKED`, and append paths, commands, results, decisions,
risks and next action before starting the next row. Do not advance past a blocked
dependency. For contradictory adopted owners, record `BLOCKED: owner contradiction`
and request direction rather than inventing new behavior.

### Recovery and lifetime defaults

Preserve Profile/Appearance drafts and attempts across section switches and
close/reopen within one application lifetime, including exact uncertain replay.
Retire settings state on every retired lifetime, including same-account
reauthentication; ordinary same-lifetime refresh preserves it. Discard resets the
editable draft only and cannot erase a pending or uncertain operation. Add no
persistence or automatic background replay.

Security, bootstrap and administrator credential forms use these distinct defaults:

- Clear passwords, factor codes, bootstrap tokens and setup secrets when their
  form/flow closes or retires. Retain no credential-bearing replay queue.
- Clear submitted factor codes and provisioning passwords after dispatch. Retain
  login passwords only through the active MFA challenge, for at most five minutes
  from initial dispatch. Retain bootstrap tokens and setup seeds only through the
  active enrollment and its server expiry. Clear all secrets on flow retirement. This
  does not imply that transport-owned request bytes can be erased or recalled.
- Fence confirmations by actor, lifetime and operation identity. Closing a form
  does not mean server cancellation, and an obsolete callback cannot publish state.
- Recover uncertain outcomes through safe resource/session observation or fresh
  authentication before an explicitly confirmed new action. Do not automatically
  repeat writes or report an uncertain request as rolled back.
- Keep login non-idempotent. Preserve existing session confirmation after uncertain
  login; bootstrap completion must never become automatic sign-in.
- Preserve exact password handling, required factors, revocation effects and
  current validation contracts. Confirmed mutations remain confirmed when a
  subsequent read or publication fails; recovery must not issue a duplicate write.

### Verification and closure plan

Resolve current public Make targets and task guides before execution. Route by
authored ownership rather than filenames:

| Change or evidence | Verification route |
| --- | --- |
| Application composition and feature models | `web.application`; application sources remain under `web.app` ownership |
| Authentication and HTTP adapters | `module.auth` |
| Presentation, browser and accessibility evidence | Their authored owners, including `web.design` where applicable |
| Shared selectors | `package.ui` |
| Full-density browser regression | Preserve its existing `module.workbook` row |

Start with narrow owner slices and controlled deferred responses/request gates.
Cover duplicate activation, pending edits, malformed responses, obsolete lifetimes,
secret clearing, query/pagination ordering, version and transaction conflicts,
uncertainty, and confirmed-write/failed-refresh separation. Preserve the shipped
Profile/Appearance replay, account navigation, Security, bootstrap, enterprise-auth
and full-density browser regressions. Run representative keyboard, race and
recovery scenarios in a real browser; mocked component evidence alone is insufficient.

Use Make-owned typecheck, import-boundary, lint, generation and applicable browser
checks. Update authored source/routing inputs before generating projections; never
hand-edit generated outputs or make executable evidence depend on Markdown.
Follow the [visual maintenance guide](../guides/cartulary_visual_golden_maintenance.md)
for intentional presentation changes: reconcile normally, inspect changed images,
update only justified goldens and obtain the required fresh ordinary passes. Keep
unrelated goldens unchanged.

Before broader final verification of the implementation, run `make agent-finalize`.
Leave `RESULTS_DIR` unset unless a qualifying successful full warm run exists for
the exact source; report retained-run maintenance skips and all other omitted or
failed checks honestly. Finish AR-06 with the deletion inventory, before/after
ownership, compatibility, evidence and limitations. After final handoff edits, run
Markdown lint, `git diff --check` and the final scope audit.

Compatibility is frontend structural cleanup and interaction hardening on existing
contracts. No API or stored-data migration, backend change, dependency addition,
new workspace package, commit or push is planned. Record atomic rollback boundaries
that keep each production change with its migrated callers, dependent tests,
authored routing, generated projections and any reviewed goldens; preserve unrelated
changes and the completed AS baseline. Scope-specific evidence must not be described
as repository-wide production readiness.

### Historical document-update validation

The original planning update was limited to this handoff. Validate with `make lint-markdown`,
`git diff --check` and an exact one-file scope audit; preserve the completed
AS-01–AS-05 history unchanged. Runtime tests, generation, catalogs and golden updates
are outside this document-only step. Implementation remains pending, with every
AR workstream `PLANNED`; next action is AR-01 when implementation is undertaken.

### Authorized execution decisions and gap register

The accepted remediation plan expands this iteration through owner cleanup,
characterization, structural extraction, hardening, browser validation and handoff.
Each workstream must save terminal evidence before the next starts. No commit or
push is requested. No historical AS evidence establishes AR completion.

| Gaps | Workstream | Required disposition |
| --- | --- | --- |
| G01–G02 | AR-01 | Owner lifecycle/recovery closure and characterized boundaries. |
| G03–G04 | AR-02 | Delete unused interfaces; separate feature models and presentation; inject navigation per instance. |
| G05–G06 | AR-03 | Resource-specific editor states, narrow session ports and canonical preference publication. |
| G07–G09 | AR-04 | Synchronous auth/Security ownership, bounded secrets, uncertainty and exact adapter/input contracts. |
| G10–G12 | AR-05 | Query generations, one reviewed draft, sparse PATCH, exact safe replay and guarded session effects. |
| G13–G14 | AR-06 | Accessible browser recovery, reviewed visual evidence and complete handoff. |

One non-secret administrator draft requires Stay, Discard and leave, or Save and
leave before voluntary departure. Authorization loss bypasses that guard. Profile
and Appearance retain independent drafts and exact attempts in the current lifetime.
No browser persistence or automatic write retry is added.

Safe administrator TOTP reset, revoke-all and binding attempts retain their exact
request and transaction ID for explicit replay within the original actor/lifetime.
User create, password reset/change, login and TOTP enrollment retain no
credential-bearing replay record. Administrator PATCH has no transaction ID and
requires refreshed state and explicit review after uncertainty. Resource/session
observation establishes current state, never a receipt for an earlier write.
Confirmed writes remain confirmed through failed propagation; retry only the read.

Observations settle within 30 seconds. Application-owned observation and underlying
transport settlement remain distinct; cancellation cannot promise rollback or
credential erasure. Session-establishing requests cannot overlap an outstanding
authentication transport. Protected state clears before replacement authentication.

### AR execution log

AR-01 started at `e444439c74781624edde15115a4c6988709baeb3` on `main`.
Only the controlling handoff had staged changes at entry. Static inspection confirms
unused imperative handles and command callback, redundant selected-user derivation,
unread confirmed resource, broad editor unions, global navigation test mutation,
full-field administrator PATCH, and login validation stricter than `email_address_v1`.
Submission, pagination, stale completion and propagation behaviors are characterized
before production edits; existing target-action locks are retained where valid.

AR-01 complete. Core 01 browser acceptance/recovery and Users draft/query rules,
Core 04 secret lifetime and AC-414, and design recovery composition now specify
the approved guarantees. No backend, wire contract or domain-vocabulary change.
New `accountFrontendBoundaries.test.tsx` and its authored web.application row
characterize behavior against unchanged production source. The initial run
`20260907T031725Z-p13027` had five product failures and one invalid selector;
corrected selector and independently split races run at `20260907T031916Z-p18641`.
All nine cases fail product assertions: valid deployment email blocked, duplicate
login and password dispatch, stale Security completion, stale pagination append,
full-field PATCH, lost newer draft, retained closed-form password, and unavailable
publication recovery. The final case also exposes an unhandled refresh rejection.
These are expected characterization failures, not passing implementation evidence.
Existing account-model/session baseline passes at `20260907T031735Z-p13690`.
Generation passes at `20260907T031846Z-p15640`; Markdown lint passes at
`20260907T031736Z-p13983` (before the final acceptance-criterion sentence).
Next: AR-02 structural extraction. Models are first separated without redesigning
state transitions; concrete operation controllers replace their extracted hook
bindings in the owning AR-04/AR-05 slices. This sequencing keeps extraction reviewable.

AR-02 complete. Removed unconsumed imperative handles/wrappers and command-state
publication. Selected identity/version and action eligibility now derive from the
accepted selected user. Extracted `authenticationModel`, `accountSecurityModel`,
`deploymentUsersModel` and `AccountSettingsDialog`; App retains session/navigation
composition. Enterprise navigation is instance-injected through AppRoot/App.
Source ownership and generated topology were updated through Make.

Typecheck passes at `20260907T032344Z-p34120`, import boundaries at
`20260907T032246Z-p28081`, support lifecycle at `20260907T032345Z-p34478`
(the companion Security row initially failed only its aggregate fetch count),
Security at `20260907T032455Z-p36043`, and account editing/navigation at
`20260907T032432Z-p35313`. The stable dialog eliminates a redundant read; the
Security test now asserts exactly one password write rather than total unrelated
fetches, while retaining all body/revocation assertions. Intermediate extraction
type errors and missing moved style bindings were corrected; no boundaries weakened.
Generation `20260907T032225Z-p24602` and format `20260907T032328Z-p29859` pass.
Next: AR-03 resource-specific editor state and narrow session interfaces.

AR-03 complete. `accountSettingsModel.ts` now exposes distinct Profile and Appearance
states and bindings over typed scalar editor mechanics. Unloaded drafts are
`undefined`, distinct from the loaded null density override; the unread confirmed
resource is removed. Publication exists only on Profile. Session-owned observation
and preference acceptance enter through `AccountSettingsSessionPort`; presentation
uses model-derived action eligibility. Tests explicitly assert the absence of
Appearance publication and preserve lifetime-clearing assertions.

Typecheck passes at `20260907T033626Z-p48681`; account-model, editing and session
rows pass at `20260907T033626Z-p48611`. Browser conflict/exact replay passes at
`20260907T033648Z-p49814`; the complete workbook density regression passes at
`20260907T033649Z-p50053`. Generation passes at `20260907T033757Z-p140662`.
The initial typecheck identified an overly broad presentation command type;
removing its unused change capability corrected it. Initial tests at
`20260907T033527Z-p43416` found only obsolete unloaded/publication representation
assertions, which were migrated without weakening lifecycle guarantees.
No API, persistence or density migration. Next: AR-04 authentication and Security.

AR-04 complete. Authentication and Security now use application-composed concrete
controllers with snapshots, subscriptions, synchronous admission and explicit
form/lifetime retirement. `accountOperation.ts` bounds observation separately from
transport settlement and retains only sanitized diagnostic metadata. Authentication
and logout reserve the session owner's narrow cookie-changing transport slot until
actual settlement, including across retirement. This is not a write retry manager.
`refreshSessionForAccountOperation` names the guarded session publication port used
by Profile and Security; the old internal method name has no compatibility alias.

MFA passwords expire five minutes from initial dispatch. Bootstrap tokens/seeds
expire with enrollment; credential inputs clear on dispatch or the specified flow
boundary. No credential replay record, persistence or automatic write retry was
introduced. Provider-discovery failures and safe credential-read failures remain
visible and retryable. Credential resources/results must match the captured actor.
Confirmed login/Security writes preserve confirmation after failed propagation and
retry only session/resource observation. Native Security and bootstrap forms expose
field-associated validation and operation-specific recovery.

Both adapter families now require caller-owned transaction IDs where declared and
propagate observation signals. TOTP contexts exclude mixed bootstrap/session inputs;
bootstrap bearer requests omit cookies and CSRF. `accountInputValidation.ts` owns
browser projections of email, display-name and exact provisioning-password rules.
The deployment-user model's new transaction IDs are captured at its existing call
sites pending AR-05's immutable-attempt implementation. No public route/body/schema
or stored-data migration occurred.

Auth/Security characterization is now routed separately from the five pending
administrator failures. New lifecycle tests cover duplicate activation, MFA expiry,
bootstrap expiry, obsolete completion/navigation, timeout/transport separation,
logout admission across retirement, safe reads, actor mismatches, exact password
bytes and confirmed-write/failed-publication recovery. Typecheck passes at
`20260907T041132Z-p420807`; new lifecycle tests at `20260907T041132Z-p420693`;
characterization/session rows at `20260907T041041Z-p414380`; seven auth/support/
adapter rows at `20260907T041041Z-p414385`. Generation passes at
`20260907T041132Z-p420634`, import boundaries at `20260907T040208Z-p291244`,
and Biome at `20260907T040924Z-p362592`.

Representative real-browser local login (including duplicate activation), MFA,
bootstrap, Security and enterprise initiation pass at `20260907T040924Z-p362429`.
The latest session-transport composition also passes focused session/model and
application integration tests; AR-06 reruns integrated browser coverage on the final
source. Earlier browser failure `20260907T035539Z-p176381` expected a server error
for a now locally rejected missing factor; its replacement asserts the associated
field error and absence of a write. Password clearing is also asserted in-browser.

Intermediate failures were resolved: TypeScript discriminant/unused-import errors
(`20260907T034708Z-p150148`, `20260907T035154Z-p157319`), ASCII test-title ordering
(`20260907T035152Z-p156724` generation), obsolete read counts/public-message
expectations and expired enrollment fixtures (`20260907T035250Z-p167451`), mismatched
actor fixtures (`20260907T040208Z-p291116`), and one missing logout mock injection
(`20260907T040924Z-p362398`). Enrollment fixtures now use active expiry times and
re-enter cleared credentials; private-message fixtures retain their negative
assertions. Biome's non-null assertion finding at `20260907T040449Z-p351082` was
removed with explicit lifetime narrowing. Next: AR-05 deployment-user hardening.


### AR-05 execution — completed 2026-09-07

G10–G12 are implemented in `deploymentUsersModel.ts`, the deployment-user HTTP
adapter, render-only administration panels, `DeploymentUserActionDialog.tsx`,
`DeploymentUserLeaveDialog.tsx`, App composition, the session controller and the
existing route owner. There is one accepted query and collection, one selected
resource, one reviewed non-secret draft, and one captured mutation. Paging is
fenced by query generation, request identity and cursor. Dirty drafts survive
refresh and pending Save; PATCH contains only changed mutable fields and is
suppressed when empty. Exact manual replay retains only non-credential inputs
for TOTP reset, revoke-all and the three binding routes. Credential creates and
resets clear submitted passwords; their uncertain outcomes require observation
and explicit review. A confirmed write remains confirmed if session propagation
fails. The accepted selected resource supplies target identity/version.

App owns the narrow leave guard for panel/application navigation; indexed browser
history supports Back/Forward without replacing the accepted draft before a leave
decision. Full-document departure uses native unload protection. Same-lifetime
capability reduction retires administration state before the reduced session is
published. Credential dialogs use native forms/modal containment and the existing
overlay focus owner. Closing clears secrets immediately. Binding subjects retain
exact post-JSON code points, including whitespace and decomposed Unicode, under
Core 01 REQ-513/538/539; the old trimming was an additional inspected contract gap.
No wire/schema/storage changes or compatibility facade were introduced.

Verification (run roots below are under `.cartulary/test-results/`):

- `make generate`: `20260907T044712Z-p457464`, pass; authored source ownership
  and the ten model-test titles were added to the owner catalog.
- `make frontend-typecheck`: `20260907T044948Z-p521546`, pass.
- `make frontend-import-boundary-check`: `20260907T044902Z-p516171`, pass.
- `make format`: `20260907T044810Z-p506707`, pass; later formatting is included
  in final validation.
- `make test-slice OWNER=web.application` selecting administrator characterization,
  operation lifetime and target-loading rows: `20260907T044949Z-p521913`, pass.
  Coverage includes all five replay families, query replacement/failure/cursor
  rejection, sparse/no-op PATCH, pending edits, review, Save/Discard/Stay,
  credential clearing, propagation-only retry, self-revocation and obsolete
  success/401/403 after capability loss and reauthentication.
- `make test-slice OWNER=module.auth` selecting administrator integration,
  public-error adapters and account-support/binding lifecycle rows:
  `20260907T044950Z-p522152`, pass.
- `make service-backed-test-slice OWNER=module.auth` selecting stateful ordinary
  deployment-user administration and revoke-all rows:
  `20260907T044725Z-p461911`, pass. The former now also proves real Back/Forward,
  Escape/Stay, Save-and-leave, retained draft review and last-admin rejection.

Intermediate related failures were resolved: initial extraction type errors
(`20260907T043551Z-p443044`), stale no-op PATCH/test presentation assumptions
(`20260907T044024Z-p445193`, `20260907T044025Z-p445473`), formatting dependencies
(`20260907T044552Z-p452327`), new fixture typing/imports and a reused consumed
Response (`20260907T044623Z-p456720`, `20260907T044713Z-p458044`,
`20260907T044827Z-p514963`), and raw server prose expectations
(`20260907T044714Z-p458772`). An initial authored-row rejection required ASCII
sorting the new titles before generation. No product requirement was weakened.

Compatibility/rollback: internal controller and adapter callers move atomically
with their tests, route integration and authored/generated ownership. The workbook
startup URL replacement preserves the route owner's opaque history state. Backend
idempotency, last-admin enforcement and deployment capability remain authoritative.
The new UI requires explicit review/leave decisions and fresh credentials where
replay is unsafe. No attempt survives a new application lifetime. AR-06 remains
responsible for final integrated readiness evidence, accessibility and visual review.

### AR-06 execution — completed 2026-09-07

`make agent-finalize` passed before broader verification at
`.cartulary/test-results/20260907T045049Z-p523669`. `RESULTS_DIR` was unset:
retained successful-run maintenance was skipped because there was no qualifying
full warm `check` run. The final scope remains this account/frontend iteration;
these results do not assert backend-wide conformance or release readiness.

The final interaction audit added native modal containment and standard tab
navigation to Account settings, native forms to administrator dialogs, and a
small `useAccountDialogFocus.ts` adapter to the existing registered-overlay
navigation helper for Tab/Escape. It has no form values, HTTP operations or
mutation authority. It preserves native text editing and checked-radio Tab
behavior. Dialog closure releases native modal containment before restoring
trigger focus. Validation messages sit outside labels and are associated through
`aria-describedby`, preserving stable accessible names. Percentage-based dialog
bounds preserve constrained viewport behavior under browser zoom.

An additional administrator accessibility row covers Stay/Discard/Save choices,
keyboard wrapping, validation descriptions, secret clearing on Escape, reduced
motion, narrow/short viewports, text spacing, and focus restoration. The existing
settings row now explicitly checks containment and arrow-key tab navigation.
Final model review also covers duplicate resource versions inside a page and
normalized no-op Save-and-leave without manufacturing a new draft revision.

### Final gap dispositions and acceptance map

All implementation dispositions below are adopted-owner projections, not new
authority. The execution/evidence index records terminal validation.

| Gap | Remediation and affected areas | Rationale and long-term benefit | Compatibility or migration impact | Risk if omitted | Completion criteria and evidence owner |
| --- | --- | --- | --- | --- | --- |
| G01 | Core 01 browser lifecycle/recovery, Core 04 secret/session rules, bounded Design guidance, controlling handoff. Specification and documentation. | Explicit certainty and credential boundaries let future factors and phases extend one coherent policy. | Browser-only guarantees; no route, schema, backend or stored-data migration. | Conflicting recovery semantics, unnecessary credential retention and false receipts. | Owner sections reviewed against all approved policy tables; no Markdown consumed by executable evidence. AR-01. |
| G02 | Reproduced deferred-response cases, passing baselines and authored routing. Tests and documentation. | Characterization distinguishes live protections from actual failures and makes structural changes reviewable. | Internal test additions; historical AS results remain historical. | Refactoring sound guards while leaving real races. | Nine baseline defects classified; fixed cases pass in their owning slices. `web.application`/`module.auth`. |
| G03 | Delete both account-administration panel handles, imperative wrappers and unused command-state plumbing; derive identity/version from accepted resources. Implementation, tests and documentation. | Fewer public surfaces and invalid state combinations reduce future synchronization work. | Repository callers migrate atomically; no shim. | Obsolete interfaces attract new coupling or stale state. | Caller searches, typecheck and owner regressions; deletion inventory below. AR-02. |
| G04 | Concrete application-owned Authentication, Security and deployment-user controllers; render-only panels; extracted settings dialog; instance navigation port. Implementation, tests, routing and documentation. | Operation lifetime is explicit and independently testable as features grow. | Internal composition changes only; no dependency/package/framework added. | Mount-owned operations, stale closures and shared test mutation. | Instance isolation, lifecycle/StrictMode and session integration tests; App retains route/session composition. AR-02/04/05. |
| G05 | Profile and Appearance states/bindings/attempts encode their own resources and scalar drafts; unloaded Appearance differs from null override. Implementation and tests. | Invalid cross-resource editor states become unrepresentable; future fields need deliberate transitions. | TypeScript migration only; density values unchanged. | Casts and shape checks spread across new editors. | Typecheck plus complete current-account lifecycle and density regressions. AR-03. |
| G06 | Narrow session ports, session-canonical preferences, Profile-only publication recovery and model-derived action availability. Implementation and tests. | Draft, attempt, accepted resource and session publication remain distinct facts. | Independent saves, exact replay and in-memory lifetime retention preserved. | Competing caches, lost newer edits or duplicate writes on propagation failure. | Profile/Appearance conflict, replay, ignored aborts, same-version inconsistency, reauthentication and full-density browser evidence. AR-03/06. |
| G07 | Synchronous admission, flow/operation identities, lifecycle guards and bounded observations, with transport settlement separate from UI timeout. Implementation, tests and specification. | Future authentication phases inherit explicit ownership without a generic mutation manager. | Duplicate activation dispatches once; obsolete effects are inert; no new route idempotency. | Duplicate sessions, stale redirects and replacement-session corruption. | Deferred duplicate/retirement/timeout/session-replacement tests and authentication/Security browsers. AR-04. |
| G08 | Bounded MFA password retention, immediate submitted-code/password clearing, enrollment expiry, safe credential resources, separate provider discovery failures and sanitized errors. Multiple areas. | Sensitive state has a short, understandable protocol lifetime and honest recovery. | Terminal/uncertain actions require fresh credentials; local login survives discovery failure. | Secret over-retention and misleading empty/current/confirmed states. | Model, component, adapter and browser secret/loading/failure/expiry assertions; visual changes reflect cleared inputs. AR-04/06. |
| G09 | Required caller transaction IDs, exclusive TOTP contexts/signals and owner-defined email/name/password validation. Implementation and tests. | Typed routes prevent malformed combinations and admit valid deployment-local identities. | Internal callers updated atomically; login/PATCH remain without transaction IDs; exact password bytes preserved. | Rejected valid accounts, accidental keys and incorrect bootstrap cookie/CSRF behavior. | Adapter body/header/credentials tests, scalar validation and malformed-response cases. AR-04. |
| G10 | Query draft/accepted query/rows/cursor/read state separated and fenced; explicit first-page recovery. Implementation and tests. | Additional filters and collection growth remain bounded by one query generation. | Existing debounce/page size retained; last accepted rows remain during refresh. | Mixed-query pages, stale cursors and duplicate resources. | Deferred ordering, duplicate activation/resources, query failure, invalid cursor and retirement tests. AR-05. |
| G11 | One accepted target, one reviewed draft and captured operation; sparse/no-op PATCH; explicit review and Save/Discard/Stay navigation. Multiple areas. | New fields stay isolated from unrelated changes; refresh cannot become overwrite. | Deliberate navigation/review interactions replace silent draft loss; no per-user cache. | Lost edits, broader writes and brittle concurrency. | Pending/newer-edit, conflict, target/Clear, real Back/Forward, unload and accessible leave decisions. AR-05/06. |
| G12 | Exact safe replay, captured target/version/binding/key, confirmed-write versus propagation state, guarded self/capability effects. Multiple areas. | New administrator actions can extend clear outcome ownership without a queue. | Credential operations never replay; subjects preserve exact identity code points. | Wrong-target publication, duplicate revocation and stale privileges. | All five replay families, transaction conflict, self effects, last-admin/binding routes, obsolete 401/403 and read-only retry tests. AR-05. |
| G13 | Native forms/modals, associated errors, keyboard/focus recovery and constrained presentation. Implementation, browser tests and reviewed visuals. | Recovery remains reachable and understandable as states expand. | Semantic selectors preserved; deliberate changed pixels require reviewed goldens. | Correct models with inaccessible or clipped recovery controls. | Routed browser/accessibility scenarios and ordinary visual reconciliation; changed images reviewed individually. AR-06. |
| G14 | Current owner-routed verification, gap/deletion/ownership records, evidence index and atomic rollback guidance. Tests, routing and documentation. | The next phase starts from an auditable baseline. | Rollback includes source, callers, tests, authored/generated routing and reviewed goldens. | Partial execution mistaken for readiness or inconsistent rollback. | All six rows terminal, no related required check unresolved, Markdown/whitespace/staged/scope audits complete. AR-06. |

### Deletion and ownership inventory

Removed: `AccountSecurityPanelHandle`, `DeploymentUsersPanelHandle`,
`DeploymentUsersPanelCommandState`, `onCommandStateChange`, their effects and
imperative wrappers; duplicated selected target identity/version state; the
editing model's unread confirmed-resource cache; the mutable global enterprise
navigation test setter; optional transaction-key generation in account adapters;
wide session dependencies in account editors; and PATCH's unconditional mutable
fields. Other subsystems' imperative interfaces were outside this iteration.

| Before | After | Extension rule |
| --- | --- | --- |
| AuthGateway owns requests, secrets, discovery and redirects in render state. | `AuthenticationController` owns the flow; AuthGateway renders snapshots and commands; App supplies an instance navigation port. | New factors add explicit flow states and secret lifetimes. |
| Security panel owns HTTP and render-closure session callbacks. | `AccountSecurityController` owns credential operations and observations; App provides guarded session effects. | Add credential actions inside its serialized lifetime, not another global slot. |
| Deployment-user panel combines reads, selection, edits, requests and refresh. | `DeploymentUsersController` owns query/target/draft/operation state; dialogs/panel own presentation; App/route owner owns departure. | Add typed changed fields or exact safe action variants, with explicit replay policy. |
| Profile/Appearance share broad value/resource unions and session coupling. | Resource-specific states/bindings; Profile publication and session-canonical preferences. | Add a resource editor deliberately without widening unrelated publication state. |
| App contains settings presentation and unfenced administration effects. | App composes controllers and narrow ports; settings presentation and overlay focus are local components/helpers. | Keep session acceptance and application navigation under their existing owners. |

### Compatibility, rollback and remaining boundaries

No backend, public HTTP inventory, idempotency semantics, database data, lockfile,
dependency or workspace package changed. Internal callers and tests must move
with controller and adapter changes. Login now accepts owner-valid addresses
such as `operator@deployment`. Administrator PATCH is sparse, clean drafts do
not dispatch, and leaving a dirty draft requires an explicit choice. Safe replay
uses the original target/key/body even after later reads. Fresh authentication
retires all drafts/attempts; normal application-lifetime Profile/Appearance
retention remains intact. Secrets are never put in browser persistence.

Rollback units are AR-02 extraction/composition; AR-03 current-account model;
AR-04 authentication/Security/adapters/session admission; AR-05 administration,
route guard and opaque history preservation; and AR-06 presentation/evidence.
Apply dependent rollback units in reverse order. Each unit includes production
source, migrated callers/tests, authored ownership/routing and Make-generated
projections; AR-06 includes any promoted PNGs and their manifest. Owner documents
must continue to describe the behavior intentionally retained. Preserve the
preexisting staged handoff contribution and completed AS history during rollback;
use scoped diffs, not workspace reset.

Browser abort/deadline completion is not proof of server cancellation. Unsettled
session-establishing transports keep authentication admission closed and offer
reload guidance. Safe state observations are not receipts. JavaScript reference
release does not claim secure erasure of strings or transport bytes. There is no
reload persistence, cross-reauthentication replay, automatic write retry, generic
form framework or global mutation manager. Unknown/unowned same-document history
entries use a bounded leave decision without inferring a traversal distance;
full-document departure uses native unload protection. These are intentional
policy boundaries, not deferred implementation defects.

### AR-06 current evidence index

Run roots in this index are under `.cartulary/test-results/`. Detailed row results
are in each run's `rows/`; target/unit summaries and browser reports retain the
underlying terminal evidence. Successful rows inside a failed combined run are
identified explicitly rather than treating that run as successful.

| Verification | Result and run root |
| --- | --- |
| `make agent-finalize`, before final broader verification | Pass, `20260907T050826Z-p1519650`; `RESULTS_DIR` unset, retained-run maintenance skipped. |
| `make frontend-typecheck` | Pass, `20260907T050925Z-p1523554`. |
| `make frontend-import-boundary-check` | Pass, `20260907T050925Z-p1523561`. |
| `make lint-biome` | Pass, `20260907T050925Z-p1523567`; later e2e-only formatting uses Make. |
| `make build-web` | Pass, `20260907T050925Z-p1523627`. |
| `make frontend-unit` | Pass, all 456 routed units, `20260907T050926Z-p1523953`. |
| `make generate-drift` | Pass, `20260907T050928Z-p1524090`. |
| `make generated-artifact-policy-check` | Pass, `20260907T050928Z-p1524092`. |
| `make json-shape-check` | Pass, `20260907T050928Z-p1524094`. |
| Application session, authentication/Security and administrator operation-lifetime slices | Pass, `20260907T050653Z-p1369823`, including anonymous confirmation that preserves its current flow and fences replacement. |
| Administrator/binding/public-error integration slices | Pass, `20260907T050301Z-p1314494`. |
| Authentication, MFA/bootstrap, Security, provider discovery, enterprise landing, session expiry and authentication accessibility browser slices | Pass, all 17 units, `20260907T051110Z-p1651533`. |
| Account-settings keyboard/responsive and settings/administrator accessibility slices | Pass, `20260907T050227Z-p1263838`. Administrator accessibility also passes 200% zoom at `20260907T050734Z-p1419226`. |
| Administrator real navigation, native unload, panel Stay, sparse/conflicting PATCH and last-admin protection | Pass, `20260907T051016Z-p1551382`. Native reload is initiated in the page; its unload dialog is dismissed and the draft is verified intact. |
| Administrator revoke-all and ordinary stateful administration | Pass, `20260907T050302Z-p1314731`. |
| Account real conflict/replay, account-menu navigation and menu accessibility | Each named row passed in `20260907T045209Z-p531497`; that combined run failed the subsequently repaired responsive case. Final ordinary visuals also cover all account-menu/settings states. |
| Workbook full-density regression | Pass, `20260907T045245Z-p614182`; all explicit densities and the null/default override remain covered. |
| Full ordinary visual reconciliation before promotion | `20260907T051017Z-p1552357`: 122 active captures / 122 committed goldens, zero orphan/missing/ambiguous mappings, all 26 registered fixtures resolved. Only the five intentional authentication comparisons failed. |

Additional final-audit fixes: session confirmation of an already-anonymous login
flow must not retire that flow when discovery returns `session_required`. The
session owner now distinguishes authentication confirmation from ordinary shell
or caller observation; it retains the uncertainty message while still fencing
replacement and accepting a validated successful session. The unit/session and
visual/browser fixtures exercise this boundary. Password visibility controls no
longer contribute to the password input's accessible name.

Resolved AR-06 failures: the initial full unit run
`20260907T045208Z-p531342` exposed one remaining clean-draft Save assumption;
`20260907T045209Z-p531497` exposed native-dialog bounds under zoom;
`20260907T045210Z-p532023` exposed an accessibility fixture trying to focus a
now-disabled login button. Administrator/settings accessibility iterations
`20260907T045448Z-p823654`, `20260907T045621Z-p870291`,
`20260907T045738Z-p996953` and `20260907T050051Z-p1162950` drove label association,
Tab wrapping and focus restoration fixes. The last also contained a browser
session-start fixture error; the same selection passed in
`20260907T050227Z-p1263838` without an infrastructure change. The native unload
fixture in `20260907T050732Z-p1418486` waited on Playwright's canceled reload;
initiating reload inside the browser verified the same unload decision without
that automation wait, and `20260907T051016Z-p1551382` passed. A formatting-only
failure at `20260907T050052Z-p1163816` was resolved through Make formatting.

Early visual runs `20260907T045212Z-p532264`, `20260907T045244Z-p611266`,
`20260907T045737Z-p996040`, `20260907T050140Z-p1212275` and
`20260907T050655Z-p1371675` were not accepted for promotion: zoom clipping,
fixtures relying on retained terminal credentials or missing bootstrap expiry,
the anonymous-confirmation defect and an ambiguous error selector were resolved
first. Unreached captures in those failed runs did not justify deleting goldens.
The final pre-promotion run reached every capture and failed only reviewed pixel
changes. No unrelated golden, renderer, viewport, zoom, mask, scroll-normalization
or screenshot-scope changes are authorized by this iteration.

### AR-06 visual refresh record

Accepted trigger: the adopted browser lifecycle and secret-retention guarantees
intentionally change authentication action availability, cleared input values and
uncertain-outcome presentation. The ordinary run above reconciled every capture
before mutation. `make browser-e2e-visual-update` passed all 12 work units at
`20260907T051441Z-p1724837`, including functional assertions and reconciliation.

The update emitted 22 changed candidates. Twelve settings candidates differed in
the native modal's background dimming, within the existing ordinary comparison;
five other candidates contained only 1–13 changed antialiasing pixels. Those 17
unnecessary candidates were discarded by restoring their original committed bytes.
`make generate` regenerated the complete manifest at
`20260907T052201Z-p1773704`. Only the five justified authentication images below
remain changed; no PNG was manually drawn or edited.

All five captures belong to authored owner row
`module.auth.visual.capture_the_anonymous_auth_gateway_across_initia_755030aa99`,
scenario `scenario_b04591a66827`, project `chromium`. They are active nonregistry
captures: the reconciler reports no registered fixture IDs for these images.
Their stable capture intents and exact golden names provide the capture identity;
no fixture ID or ownership was inferred from a filename.

| Stable capture intent | Changed golden under `apps/web/e2e/workbook.visual.spec.ts-snapshots/` | Individual visual review |
| --- | --- | --- |
| `auth-loading` | `auth-loading-linux.png` | Sign in remains disabled while initial session discovery owns admission. Layout and text remain intact. |
| `auth-invalid-credentials` | `auth-invalid-credentials-linux.png` | Terminal authentication failure clears the password; the field and associated error remain reachable. |
| `auth-invalid-mfa` | `auth-invalid-mfa-linux.png` | The submitted factor is cleared; the password remains only in the bounded active challenge. Focus and field error remain visible. |
| `auth-mfa-setup-required` | `auth-mfa-setup-required-linux.png` | Completion is disabled before enrollment begins; bootstrap state is represented without displaying the token. |
| `auth-service-unavailable` | `auth-service-unavailable-linux.png` | The password is cleared and the uncertainty message gives explicit retry/settlement/reload guidance, with no clipping. |

Each retained PNG was inspected after promotion as well as its failed-run actual
before promotion. Viewport remains 1440 × 900, browser zoom 100%, device scale 1,
and the pinned renderer is unchanged. Masks, scroll normalization, screenshot
scope, capture identities and comparison tolerances did not change. The existing
settings/menu and unrelated goldens remain byte-identical.

After manifest regeneration, `make generate-drift` passed at
`20260907T052233Z-p1776957`, `make generated-artifact-policy-check` passed at
`20260907T052233Z-p1776996`, and `make json-shape-check` passed at
`20260907T052233Z-p1776995`.

Both required fresh ordinary full visual runs passed all 12 work units against
the final five-image manifest: `20260907T052233Z-p1777383` and
`20260907T052412Z-p1878194`, each invoked with
`make browser-e2e-visual CARTULARY_HARNESS_CACHE_MODE=off`. Both reconcile all 122
active captures with zero orphan, missing or ambiguous mappings and all 26
registered fixtures resolved. No further golden promotion is needed.

The additional `make test-slice OWNER=web.design` at
`20260907T052302Z-p1823989` passed all visual and functional browser rows, but its
accessibility group and aggregate failed because the settings fixture pressed
Enter before Save became enabled after a keyboard tab refresh. Its retained
Playwright trace shows the disabled submit control at that transition; the final
failure snapshot shows the preserved draft without a submission. The fixture now
asserts Save readiness before native Enter, preserving all keyboard, field-error
and recovery assertions without a sleep, retry or production change. The same
owner selection is rerun for terminal evidence. That run also supplies fresh
successful per-row evidence for account real conflict/replay, account navigation,
responsive settings, workbook geometry, administrator/menu accessibility and the
visual fixture matrix.

The rerun `make test-slice OWNER=web.design` passed all 17 work units at
`20260907T052652Z-p1928401`, including all five accessibility rows, the visual
fixture matrix/catalog checks and all four functional browser rows. The final
fixture edit also passes `make lint-biome` at `20260907T052726Z-p1971403`.
The final typecheck passes at `20260907T052930Z-p1979904`.

### Completion and handoff

All AR-01–AR-06 workstreams and G01–G14 dispositions are complete. No related
required implementation, test, accessibility review or visual review remains
unresolved. The working tree contains 47 intentional paths: account/frontend
production source and tests, five authentication PNGs, Core 01/Core 04 and Design
amendments, this handoff, authored source/test ownership and three generated
topology/visual-manifest projections. The 12 new account source/test files all
have authored ownership. Repository caller searches find none of the deleted
panel handles, command-state callback, global enterprise-navigation setter or
unread confirmed-resource field. No temporary diagnostics entered tracked scope.

The complete historical AS body is byte-identical to HEAD. The existing index
still contains only the original 175-line handoff addition; no staging, commit or
push was performed. Branch and HEAD remain `main` at
`e444439c74781624edde15115a4c6988709baeb3`. The explicit final path audit excludes
backend, public contracts, authored SQL, dependencies, lockfiles and unrelated
goldens. `git diff --check` and `git diff --cached --check` pass. The terminal
tracker edit is followed by Markdown lint and the same final diff/scope checks.
`make lint-markdown` passes on that completed handoff at
`20260907T052930Z-p1979929` (`adhoc/lint-markdown/tool-run-summary.json`);
the final evidence-reference edit is checked again before return.

Omitted broader checks are intentional: no full `make check`, `make ci`,
`make release-check` or backend-wide suite was needed for this frontend-only
change; these results make no whole-repository conformance or release claim.
Shared selectors did not change, so a separate `package.ui` migration/slice was
unnecessary. Retained-run maintenance was skipped by `make agent-finalize` because
`RESULTS_DIR` was unset and no qualifying full warm check existed. Current owner
slices, the full frontend unit suite, build, type/import/lint checks, generation
policies, applicable browser/accessibility regressions and both full ordinary
visual passes supply this iteration's evidence. The intentional browser lifetime
and uncertainty boundaries above remain the constraints for future expansion.

Next action: review the cohesive implementation and owner amendments using this
evidence index. Integrate source, tests, authored/generated ownership, reviewed
goldens and this handoff together; use the scoped rollback units above if needed.
