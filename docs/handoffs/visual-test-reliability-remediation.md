# Visual test reliability remediation

Implemented G1–G4 and completed acceptance. Both final full ordinary visual runs
passed with exact reconciliation; `make check` passed all 786 work units. The
reviewed update changes 26 goldens. Current evidence, schema cutovers, fault
injections, and the remaining historical-cause uncertainty are recorded below.

## Evidence baseline

Implementation began at `8ddb16b0` with a clean worktree. Historical runs below
used dirty source snapshots and are investigation evidence, not validation of
this implementation.

- `20260908T234914Z-p3597126`: Membership audit committed comfortable density,
  reloaded into nine JavaScript asset 404s, and waited about 56 seconds for
  navigation. Page-bound preference restoration then failed after context
  closure. Later Metadata and Lifecycle traces read comfortable density.
- `20260908T235153Z-p3652197`: a concurrent Make build ran approximately
  23:51:55.931–23:52:03.590 UTC, overlapping the failed reload. Mutable output
  replacement is strongly supported; controlled publication-isolation results
  are recorded below.
- `20260908T235436Z-p3706625`: Lifecycle compact differed by 17,237 pixels (4%).
  Image inspection shows roughly 6px of internal vertical displacement without
  changed content or drawer bounds. Its exact historical trigger is unproven.
- Network Analysis passed its visual row. That observation does not establish
  the causes of either independent failure.

## Implementation and verification

The Testing Harness NLSpec owns requirements; this handoff records execution.
`docs/domain.md` is unchanged. No product API or database migration is required.

- Frontend readiness and browser lifecycle now build directly into private
  current-run staging, seal a profile-specific artifact and receipt, and serve
  that exact artifact through the existing suite runtime lease. Conventional
  packaging publication uses the completed artifact. Embedded preparation hashes
  and reads the same private source. Production and measurement remain distinct.
- Visual authentication support owns one page-local generated-protocol preference
  resource. All seven density loops and auxiliary visual pages use it. Account
  Settings presentation faults delegate to that route owner; real persistence
  coverage remains independent.
- Shared capture preparation establishes readiness and normalization before
  semantic drawer anchors. Observation-only geometry verifies three consecutive
  frames, including legal positioning and focus. Lifecycle, Membership audit,
  Membership management, Metadata, and workbook preferences declare framing.
  Existing workbook-grid framing remains explicit.
- UI contracts expose the drawer body scrollport; the application and workbook
  shells expose resolved presentation for observation. Capture evidence checks
  declared, observed, and registered profiles. Asset and preparation diagnostics
  retain their primary failure classification without preference identities or
  response bodies.
- Authored harness catalog rows cover artifact publication overlap, preference
  isolation, geometry perturbation, missing assets, and profile rejection.
  Generated routing was refreshed through Make.

The compatibility cutover is atomic: frontend build receipt v1 is new; stack v6
becomes v7; capture intent v1 becomes v2; reconciliation v2 becomes v3; fixture
registry v5 becomes v6; browser group result v5 becomes v6; browser target result
v3 becomes v4. Group v6 records stack v7 and accepts artifact failure classes.
The existing diagnostic-only historical registry archives unchanged group v5
bytes to close its older target schema reference; current readers cannot use it.
Registry profiles are
per exact golden path. Backend generation v1 references stack v7. Old readers and
schema files are removed; historical evidence is investigation-only. Public Make
targets, product routes, row identities, screenshot tolerances, and retries are
unchanged.

Final validation results and finalizer-generated changes are recorded below.

## Reviewed refresh request

Accepted trigger: the harness intentionally changed post-normalization drawer
framing. Ordinary review run
`.cartulary/test-results/20260909T015810Z-p262274` accounted for all 210 captures,
210 active goldens, and 26 resolved registry fixtures. No capture was missing,
orphaned, or ambiguous. Its only reconciliation error was the failed comparison
attempt. All functional, profile, and geometry assertions completed.

The following images were inspected as expected/actual pairs. Content, density,
viewport, zoom, spacing, masks, and screenshot scope are preserved; declared
scroll and focus preparation explains each listed difference. Earlier candidates
that hid review controls at zoom were rejected and corrected before this review.
The Evidence timestamp differences were fixed in preparation and are excluded.
The accepted refresh was produced only by `make browser-e2e-visual-update` in
`20260909T022005Z-p482026`. Exactly the 21 listed PNGs changed; the manifest was
generated by that target. Post-promotion comparison matched the reviewed pixels
exactly for 19 files. Lifecycle compact and workbook preferences confirmed-stale
differed from their reviewed candidates at only 26 and 12 pixels respectively,
with a maximum channel delta of 2 and unchanged geometry/content. No additional
golden was accepted. Ordinary post-promotion runs are recorded below.

The first update attempt (`20260909T021029Z-p358049`) completed its checks but
post-update review rejected the candidate. It also rewrote 19 images for small
pixel differences, changed the bottom border of Membership audit zoom, and
captured the grouped grid at the wrong horizontal position. All candidate images
and the generated manifest were rolled back to the original committed bytes.
Grid setup now waits for font/presentation readiness and checks its declared
offset rather than adopting a later observed offset. Update mode retains images
that already pass the ordinary comparison contract. No tolerance changed.

The three focused invocations at `20260909T022507Z` exposed another isolated
failure boundary. Membership audit passed in all three. Lifecycle's drawer
geometry matched, but 296–863 pixels differed in the account-menu label: the full
suite had inherited the long display name persisted by the preceding account-menu
scenario. The menu scenario now uses the existing local profile fixture while
preserving actual application memberships and administration context. Its long
label still receives visual coverage without changing the shared account. The
resulting neutral account-label golden changes require the additional ordinary
review recorded below.

| Golden filename | Owner row | Fixture ID | Reviewed reason |
| --- | --- | --- | --- |
| `membership-audit-stale-linux.png` | `web.design.visual.membership_audit_browsing` | None (exact nonregistry capture) | Audit panel start is re-established after normalization; stale values and retry remain visible. |
| `membership-management-pending-linux.png` | `web.design.visual.membership_management_visual` | None (exact nonregistry capture) | Membership panel start replaces incidental post-click scroll; pending status and disabled action remain visible. |
| `membership-management-uncertain-linux.png` | `web.design.visual.membership_management_visual` | None (exact nonregistry capture) | Membership panel start replaces incidental post-click scroll; observation and recovery remain visible. |
| `membership-management-removal-zoom-linux.png` | `web.design.visual.membership_management_visual` | None (exact nonregistry capture) | The document explicitly reveals the oversized drawer end, then centers Confirm removal within its legal scroll range. |
| `metadata-loading-linux.png` | `web.design.visual.metadata_editing` | None (exact nonregistry capture) | Promoted-fields panel start replaces incidental initial padding/scroll. |
| `metadata-dirty-linux.png` | `web.design.visual.metadata_editing` | None (exact nonregistry capture) | Promoted-fields panel start is established after normalized dynamic text; exact input and focus remain. |
| `metadata-review-zoom-linux.png` | `web.design.visual.metadata_editing` | None (exact nonregistry capture) | The document explicitly reveals the oversized drawer end, then centers Use this version within its legal scroll range. |
| `metadata-compact-linux.png` | `web.design.visual.metadata_editing` | None (exact nonregistry capture) | The compact promoted-fields panel start is explicit; the exact severity draft and focus remain. |
| `metadata-comfortable-linux.png` | `web.design.visual.metadata_editing` | None (exact nonregistry capture) | The comfortable promoted-fields panel start is explicit; the exact severity draft and focus remain. |
| `lifecycle-review-zoom-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Confirm Close is centered in the visible drawer intersection; outer document scroll resets, retaining the header. |
| `lifecycle-review-spacing-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Confirm Close is centered after expanded text spacing settles. |
| `lifecycle-pending-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | The pending confirmation control receives its declared centered composition without forcing disabled focus. |
| `lifecycle-uncertain-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Replay original action is centered after recovery state and normalized text settle. |
| `lifecycle-confirmed-refresh-failure-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Refresh current incident is centered after the confirmation state settles. |
| `lifecycle-closed-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | The named Lifecycle panel start replaces inherited recovery-button scroll. |
| `lifecycle-compact-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Confirm Reopen is centered using fractional border-box geometry, replacing the old incidental position. |
| `workbook-preferences-uncertain-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | Keep observed value is centered within the legal drawer range, making the recovery controls explicit. |
| `workbook-preferences-uncertain-narrow-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | Keep observed value uses the same legal centered anchor at narrow width. |
| `workbook-preferences-confirmed-stale-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | The startup preferences panel start replaces the prior outcome scroll. |
| `workbook-preferences-compact-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | The compact startup preferences panel start is explicit. |
| `workbook-preferences-comfortable-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | The comfortable startup preferences panel start is explicit. |

## Account-label refresh review

Ordinary run `20260909T023048Z-p669351` retained 202 captures and all 210 goldens.
Eight account-menu goldens had no capture because profile-only session mocking
had frozen memberships before incident creation. That setup defect also prevented
the `20260909T023628Z-p711839` update from promoting any candidates. Its final
headline reported a second defect: copied candidate PNGs retained source-file
permissions. The underlying product assertion remains in its browser-group result.
Profile-only mocking now reads live session context; snapshot staging makes a
complete private copy before publishing its directory. All 17 label differences
below were reviewed as expected/actual pairs. The shared long account name becomes
the declared worker baseline, normalized without its numeric slot. Drawer contents,
controls, focus, density, and framing remain unchanged. Outside the header, at most
27 pixels differ per image, with a maximum channel delta of 3. No new mask,
tolerance, or retry was added. The missing account-menu captures required narrow
verification after the session fix before the update could proceed.

That narrow account-menu verification passed in `20260909T024603Z-p755800`,
covering all eight unchanged goldens with live session context and local profile
writes. Execution smoke coverage also passed with a world-readable source PNG
copied to an owner-only candidate, candidate reuse, and promotion rollback.

The final update, `20260909T024730Z-p795125`, passed all 12 work units with 210
captures, 210 active goldens, 26 resolved registered fixtures, and zero orphans,
missing goldens, or ambiguous mappings. Exactly 26 reviewed PNGs differ from the
original checkout. The 17 label candidates match their reviewed pixels except
Lifecycle compact (20 pixels, maximum channel delta 1) and preferences
confirmed-stale (12 pixels, maximum delta 2). All other updated images retain the
previously reviewed framing. No unreviewed image was promoted.

| Golden filename | Owner row | Fixture ID | Reviewed reason |
| --- | --- | --- | --- |
| `lifecycle-closed-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-comfortable-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-compact-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-confirmed-refresh-failure-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-pending-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-reason-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-review-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-review-narrow-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-review-spacing-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-review-zoom-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `lifecycle-uncertain-linux.png` | `web.design.visual.lifecycle` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `workbook-preferences-comfortable-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `workbook-preferences-compact-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `workbook-preferences-confirmed-stale-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `workbook-preferences-uncertain-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `workbook-preferences-uncertain-narrow-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |
| `workbook-preferences-unset-linux.png` | `module.workbook.visual.preferences` | None (exact nonregistry capture) | Remove the inherited account-menu long-label state; retain the neutral worker label. |

## Verification scope corrections

The first broad `make check`,
[20260909T025944Z-p941148](../../.cartulary/test-results/20260909T025944Z-p941148),
finished with 781 of 786 work units passing. Two failures were introduced by this
remediation: a synthetic browser-group row ID violated the canonical grammar,
and capture geometry assembled a selector instead of using the UI-contract
builder. Both are corrected; diagnostic element identities are plain structural
identifiers rather than assembled selectors.

Three existing Network Analysis verification assumptions also needed maintenance.
The exact source-ownership inventory omitted 23 already-existing files; those
files now have their existing directory owners. The import identity test now
checks the session-owned `NetworkFlowImportController`, which calls the shared
claim-gated adapter. The keyboard recovery scenario observes its completion
status instead of waiting on heading copy. The browser artifact scanner no longer
uses a bundler-selected chunk filename as a feature boundary: the authored import
policy restricts the Network Flow protocol entrypoint to its generated-contract
adapter, and the scan invokes that policy. Protected Audit/Revisions and removed
module scans remain intact. The scanner resolves the current run's sealed
production artifact when a runtime lease is present. These changes repair
verification projections; they do not change Network Analysis product behavior.

All five failed units passed their focused reruns:
[routing](../../.cartulary/test-results/20260909T031402Z-p1213688),
[source and selector ownership](../../.cartulary/test-results/20260909T031402Z-p1213703),
[import identity](../../.cartulary/test-results/20260909T031402Z-p1213721), and
[browser artifact reachability](../../.cartulary/test-results/20260909T031402Z-p1213736).
[Frontend typechecking](../../.cartulary/test-results/20260909T031402Z-p1213938)
also passed. These focused results precede the final broad acceptance below.

Finalization passed in
[20260909T031440Z-p1216758](../../.cartulary/test-results/20260909T031440Z-p1216758).
Its machine summary reports no generated changes. Retained-run maintenance was
skipped because `RESULTS_DIR` was unset; no eligible successful full warm check
was supplied. Earlier finalizer failures from stale topology helper hashes were
resolved by `make generate`, not by editing generated projections.

## Forced failure and worker replacement

Two deliberately failing, separate Make invocations close the distinction between
caught fixture failures and actual Playwright worker replacement:

- [Assertion injection](../../.cartulary/test-results/20260909T031835Z-p1252873):
  Membership audit failed with the injected assertion in worker 0; Lifecycle then
  passed in worker 1.
- [Whole-test timeout injection](../../.cartulary/test-results/20260909T031954Z-p1284445):
  Membership audit reached Playwright's `timedOut` state in worker 0; Lifecycle
  then passed in worker 1.

Both used the requested Lifecycle/Membership audit slice. The temporary injection
read the server preference before the density loop, selected and reloaded local
comfortable density, asserted the rendered density, and compared the server
preference data again. The `density-fault-precondition` attachment records only
`server_value_and_version_unchanged=true`, `rendered_density=comfortable`, and the
fault kind. It contains no preference identity or response body. The assertion
case then used `expect("injected density assertion failure").toBe("unreachable")`.
The timeout case instead reduced that test's budget to 1ms and awaited an
unresolved promise, forcing teardown and worker replacement. The primary errors
remain the intentional assertion and timeout respectively, with no replacement
by a cleanup error. Lifecycle's ordinary geometry, profile, and golden assertions
passed after each replacement.

The temporary source changes were restored byte-for-byte before final acceptance.
These runs have their own source snapshots and failed target status; they are
expected-failure evidence and do not count as successful ordinary runs. No timeout
change or fault switch is present in the delivered visual scenarios.

Artifact rejection was also checked for a dangling publication symlink and a hard
link to an inode outside the candidate. The latter must fail before changing the
linked file's permissions. These extend the existing incomplete-build, profile,
provenance, tampering, ownership, containment, and idempotent-release cases.

## Ownership and completion evidence

| Gap | Implementation and contract owners | Acceptance evidence |
| --- | --- | --- |
| G1: frontend artifact lifetime | `tools/harness/readiness/frontend-artifact.mjs`, Make build recipes, browser startup/session evidence, fixture broker, work-graph compiler, Playwright configuration; frontend receipt v1 and stack v7 | Artifact unit faults cover partial input, wrong profile/provenance, tampering, symlinks, hard links, containment, publication replacement, premature release, and idempotent cleanup. Browser overlap checks repeated reloads and lazy asset digests during independent production and measurement builds. Missing-asset injection produces the artifact failure classification. |
| G2: visual preference isolation | `apps/web/e2e/support/auth/visualPreferences.ts`, visual fixtures, Account Settings presentation adapter, seven workbook density loops, auxiliary page initialization | A comfortable server seed cannot change the declared nullable baseline. Both density orders render correctly; local variants preserve server value/version. Page closure and the separate real assertion/timeout injections prove the next page or worker starts from its declaration. The real Account Settings integration row remains unmocked. |
| G3: deterministic capture geometry | Shared visual capture preparation and workbook captures; Lifecycle/Metadata reachability helpers; UI-contract drawer scrollport selector | Correct anchoring is required across three animation frames. Perturbation tests cover viewport, zoom, scroll, focus, and delayed geometry. Negative cases retain clipping, missing-scrollport, focus-loss, and late-movement failures. Ordinary visual coverage includes density, narrow viewport, zoom, and text spacing. |
| G4: capture evidence | Visual profile helper, capture declarations/diagnostics, fixture registry v6, capture intent v2, reconciliation v3, browser row adapter, group v6 and target v4 | Invalid, empty, unknown, mismatched, or inapplicable profiles are rejected. Registry and exact nonregistry captures reconcile with their rows/scenarios/projects. Ordinary and update modes use the same preparation path. Asset, product assertion, timeout, and secondary cleanup cases preserve the primary classification. |

The changed product observation surface is limited to resolved shell presentation
attributes and the drawer-body UI-contract hook. Browser fixtures, build/readiness
and artifact lifetime, evidence schemas/reconciliation, snapshot staging,
verification routing/policy, and harness documentation are the other changed
subsystems. No data migration, product endpoint change, removed visual scenario,
new Playwright retry, raised timeout, or relaxed image tolerance is included.

The final source was regenerated through
[Make generation](../../.cartulary/test-results/20260909T032130Z-p1316127), followed
by execution smoke validation and
[finalization](../../.cartulary/test-results/20260909T032226Z-p1326628).
The finalizer again reports `updated_file_count=0`; retained-run maintenance was
skipped because `RESULTS_DIR` was unset. Authored changes and Make-generated
routing/manifests are therefore separate from finalizer output: the finalizer
contributed no file edits.

The final implementation source digest is
`sha256:fbafbb16b9eea4d2b9f0e51b7e08b612dab693d8f672a00a4008627be8213fc4`.
Markdown-only handoff edits are deliberately outside executable source identity.

## Validation commands and results

The focused visual slice was run in three fresh ordinary invocations after the
reviewed golden promotion. All passed (11/11 work units each):
[run 1](../../.cartulary/test-results/20260909T025232Z-p837351),
[run 2](../../.cartulary/test-results/20260909T025232Z-p837365), and
[run 3](../../.cartulary/test-results/20260909T025232Z-p837372).
The subsequent verification-policy and artifact-rejection refinements did not
change their reviewed framing; the final full ordinary runs below cover the
final source.

```sh
make service-backed-test-slice OWNER=web.design ROWS=web.design.visual.lifecycle,web.design.visual.membership_audit_browsing
```

| Command / selected scope | Result and retained run |
| --- | --- |
| `make check` | PASS, 786/786 units; [final full check](../../.cartulary/test-results/20260909T032253Z-p1330759), on the final source digest. |
| `make run-harness-smoke-execution` | PASS, including final artifact rejection and private snapshot staging cases; standalone exit status 0. |
| `make lint-markdown` | PASS; [specification, guide, and handoff lint](../../.cartulary/test-results/20260909T033132Z-p1657346). This result link was recorded after the check. |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.verify_continuous_workbook_shell_composition_top_96ea2f5084` | PASS, 11/11 units; [real account preference persistence](../../.cartulary/test-results/20260909T030204Z-p1067384). The visual preference mock does not apply to this row. |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95` | PASS, 14/14 units; [measurement profile and attachment/finalization](../../.cartulary/test-results/20260909T031501Z-p1219921). Run separately from CPU-heavy broad validation. |
| `make service-backed-test-slice OWNER=web.design ROWS=web.design.accessibility.lifecycle,web.design.accessibility.metadata_editing` | PASS, 11/11 units; [shared reachability helpers](../../.cartulary/test-results/20260909T032253Z-p1330258). |
| `make service-backed-test-slice OWNER=module.imports ROWS=module.imports.browser.keyboard_recovery` | PASS, 11/11 units; [keyboard recovery completion observation](../../.cartulary/test-results/20260909T032253Z-p1330300). |
| `make harness-contract` | PASS; [contract checks](../../.cartulary/test-results/20260909T032253Z-p1330667). |
| `make generate-drift` | PASS; [generated projection drift](../../.cartulary/test-results/20260909T032314Z-p1470060). |
| `make generated-artifact-policy-check` | PASS; [generated artifact policy](../../.cartulary/test-results/20260909T032326Z-p1517719). |
| `make json-shape-check` | PASS; [JSON shape validation](../../.cartulary/test-results/20260909T032327Z-p1518141). |
| `make lint-scripts` | PASS; [script lint](../../.cartulary/test-results/20260909T032331Z-p1518615). |
| `make lint-shell` | PASS; [shell lint](../../.cartulary/test-results/20260909T032332Z-p1519008). |
| `make lint-biome` | PASS; [frontend lint](../../.cartulary/test-results/20260909T032333Z-p1519875). |
| `make frontend-import-boundary-check` | PASS; [import ownership](../../.cartulary/test-results/20260909T032335Z-p1520314). |

The real persistence and measurement rows preceded only later harness permission
rejection/policy refinements; their product code and tested preference/measurement
behavior are unchanged. The final two full ordinary visual runs, overlap probes,
and full check use the exact final source digest above.

The four final harness browser regression rows passed in
[20260909T032253Z-p1330237](../../.cartulary/test-results/20260909T032253Z-p1330237)
(11/11 work units). The overlap attachment records **14 reloads, 11 unchanged
JavaScript asset digests, and no resource failures**, during separate public Make
[production](../../.cartulary/test-results/20260909T032356Z-p1537380) and
[measurement](../../.cartulary/test-results/20260909T032356Z-p1537378) builds.
Those builds overlapped both full ordinary visual runs and the full check.
The preference attachment records three page lifetimes, both variant orders,
and unchanged persisted preference value/version. Geometry and missing-asset
negative cases also passed.

```sh
make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.frontend_artifact_lifetime,harness.browser.boundary_support.visual_anchor_geometry,harness.browser.boundary_support.visual_asset_failure,harness.browser.boundary_support.visual_preference_lifetime
```

The final golden manifest file digest is
`sha256:9987c07cc77944bc0045f196b9857dca9d14ffcd01451ed447b8b72884b8a47c`.
There were no promotions after the reviewed `20260909T024730Z-p795125` update.

Both final `make browser-e2e-visual` invocations passed, 12/12 work units each:
[ordinary run 1](../../.cartulary/test-results/20260909T032253Z-p1330603) and
[ordinary run 2](../../.cartulary/test-results/20260909T032253Z-p1330617).
Their reconciliation v3 artifacts each report 210 capture intents, 210 active
committed goldens, 26 resolved registered fixtures, no orphans, no missing
goldens, no ambiguous mappings, and an empty error list. Both attest the final
source digest and golden manifest digest recorded above. The ordinary runs are
independent fresh Make invocations; neither is an update-mode or expected-failure
run.

## Remaining uncertainty and handoff

The precise historical Lifecycle 6px displacement trigger was **not reproduced**.
The structural vulnerability—normalization after incidental focus/scroll setup—
has been eliminated, and controlled geometry perturbations plus both final
ordinary visual passes validate the declared composition. No fixed pixel
correction or unexplained golden difference remains.

The historical Membership audit asset-replacement race remains strongly supported
by its nine 404s and overlapping build timing. Current regression evidence proves
that active browser assets survive independent publications; it does not claim
to recreate the exact historical scheduling interval. Density leakage and the
additional account-label mutation were removed at their fixture owners. Fault
injections prove that subsequent worker execution does not inherit the selected
visual density.

There is no product API/database migration. Current producers and consumers must
ship together with the schema cutovers; old current-schema readers are removed.
Retained historical runs remain investigation-only. Future visual scenarios must
use the shared preference resource and capture declaration, and future frontend
profiles must declare their producer and entries through authored topology.
Consult the existing visual golden maintenance guide for any future image update.
All changes remain uncommitted in the shared worktree.
