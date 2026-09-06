# Account Settings Editing Refactor Handoff

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
