# Playwright upgrade and agent diagnostics handoff

This handoff records the implementation and qualification of the Playwright
1.63.0 upgrade and private diagnostics amendment. Work started at clean commit
`331ad2e46`. Product APIs, persisted data, authentication policy and domain
vocabulary have no migration. Implementation and qualification on the available
WSL2 host and canonical Linux amd64 container are complete. Separate native Linux
host qualification remains blocked by the unavailable environment; this handoff
does not claim that support-matrix acceptance.

## Acceptance evidence

This table identifies the latest relevant evidence; the ledger below also keeps
failed attempts and superseded successes for diagnosis. Run roots are under
`.cartulary/test-results/`.

| Gate | Result | Run root |
| --- | --- | --- |
| Ordinary visual comparison, first | Pass 12/12 | `20261001T053336Z-p39651` |
| Ordinary visual comparison, second | Pass 12/12 | `20261001T053926Z-p74823` |
| Stateful browser workflows | Pass 42/42 | `20261001T044430Z-p40485` |
| Measurement | Pass 24/24; unchanged budgets | `20261001T043414Z-p57258` |
| Accessibility | Pass 20/20 | `20261001T050350Z-p63058` |
| Concurrent private diagnostics | Three passes, 1/1 each | `20261001T053336Z-p39414`, `20261001T053423Z-p50421`, `20261001T053655Z-p7751` |
| Process-lease shutdown regression | Pass 1/1 | `20261001T053117Z-p4345` |
| Real PostgreSQL lease integration | Pass 3/3 | `20261001T053123Z-p4758` |
| Entities interrupted-batch rerun | Pass 3/3 | `20261001T055329Z-p62272` |
| Timeline Undo scroll fixture | Pass 11/11 | `20261001T055852Z-p96639` |
| Export cancellation authorization fixture | Pass 3/3 | `20261001T061347Z-p99538` |
| Host doctor | Pass | `20261001T060531Z-p21162` |
| Final integrated check | Pass 984/984 | `20261001T061512Z-p22265` |
| Final harness contract | Pass 2/2 | `20261001T062202Z-p7899` |
| Final agent finalization | Pass 1/1; retained-run maintenance skipped | `20261001T061452Z-p18193` |
| Final webserver browser workflows | Pass 144/144 | `20261001T062232Z-p11652` |
| Final Markdown maintenance | Pass | `20261001T063610Z-p99331` |

Both final visual comparisons reconcile all 255 captures to 255 committed
goldens, with 29 registered fixtures and zero missing, orphaned or ambiguous
entries. Both use golden-manifest SHA-256
`5155ea46750774fda1900edbc92690c8fecf923bbb4b85e5ab119a1516282d82`.
The only later frontend change is a nonvisual fixture's wheel-completion fence;
it changes no production source, renderer input or accepted image.

The canonical Linux amd64 container and Ubuntu 26.04 WSL2 host have runtime
evidence. No independent native Linux host is available, so that separate host
support claim remains blocked. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset for finalization; no interrupted run was supplied as
successful warm-check evidence.

## Authority and implementation map

The Testing Harness NLSpec owns this change through amendment
`ui-review-diagnostics-3`, Sections 1.1, 4, 5, 8, 11.11, 14.1 and 17,
including acceptance cases AC-132 and AC-133. Human review connects those clauses
to the authored projections and tests. No executable check reads Markdown.
`docs/design.md` and `docs/domain.md` were inspected; no direction or vocabulary
change is needed.

| Gap | Implementation and acceptance owner |
| --- | --- |
| G1 | Action request v3 and renderer profile v2; seven review command identities and v2 results/receipts remain. Superseded action schemas have no compatibility reader. |
| G2 | Toolchain pins, both package manifests and pnpm-generated lockfile align at 1.63.0. Installed readiness resolves the root and web dependency chains and checks the installation receipt. Explicit helper-only `frontend-lockfile-update` serializes with frozen installation. The Ubuntu platform override is removed. |
| G3 | `visual-renderer-profile.mjs` validates the active authored profile. `playwright-packages.mjs` uses manifest-rooted Node resolution; no pnpm versioned store path remains. Installer, renderer lease, configuration, assertions, imports and golden metadata consume the active profile. |
| G4 | Lease verifies image digest/platform, copied package versions, packaged Chromium descriptor and observed browser version. Attestation v1 retains the observed image ID, package/browser identities and font hashes alongside profile v2. Local font files are hashed; capture checks served manifest/font bytes and successful font loading. |
| G5 | Exact container removal is bounded, confirmed and retryable. Partial acquisition uses a private name/label identity. Publication and signal paths retain cleanup ownership; earlier failures remain primary. |
| G6 | `diagnostic_snapshot` takes empty parameters through existing `ui-browser`, uses the session lock/epoch, binds the owned browser and verifies the same Chromium page target before CLI attachment. API references remain actionable; CLI labels are diagnostic text. |
| G7 | Fixed attach/snapshot/detach adapter, explicit private config, allowlisted environment, private file modes, bounded output and resource admission. Pre-spawn operation nonce plus boot/start/UID proof covers detached daemon acquisition; exact PID proofs and minimal stop recovery need no Playwright installation. |
| G8 | Fresh old-renderer baseline passed. All 61 changed PNGs received private visual review and transactional promotion. Two ordinary comparisons pass against the same final manifest; all 255 captures and goldens reconcile with no unexplained difference. Qualification-exposed product and fixture fixes have separate corrective evidence below. |
| G9 | Existing evidence adapters remain version-specific and schema-closed. Qualification includes report/import, privacy, PNG, accessibility and lifecycle owners. Workflow, golden-maintenance guide and review skill explain diagnostic limits and installation recovery. |

## Approved identities

| Component | Identity |
| --- | --- |
| Playwright, playwright-core, web @playwright/test | `1.63.0` |
| Renderer profile | `visual.renderer.playwright_1_63_0_chromium_1243_linux_amd64` |
| Image | `mcr.microsoft.com/playwright@sha256:bc6ab0d6d44ff4826e4cb8c1e6d801e185bfc42bb0753f8e2a30efc70db054c7` |
| Platform | `linux/amd64` |
| Chromium revision/version | `1243` / `153.0.8010.12` |
| Font manifest SHA-256 | `c21f8663e6c8fe72681b2be644aa8398538afc59a0f0cda06b94d46d5fbba5fe` |
| Host used for qualification | Ubuntu 26.04 LTS, Linux x86_64, WSL2 kernel `6.6.114.1-microsoft-standard-WSL2` |

The font files and digest are unchanged. The pnpm lockfile diff changes only the
Playwright packages, the axe peer resolution, and removal of Playwright's obsolete
optional fsevents dependency. No global CLI package or upstream skill is installed.

## Evidence ledger

Paths below are repository-relative retained run roots. Historical September 29
visual results are background only. They do not establish this implementation's
baseline or acceptance.

| Check | Result and retained root |
| --- | --- |
| Fresh unchanged visual baseline | Pass 12/12, `.cartulary/test-results/20261001T010151Z-p70520` |
| Structural generation | Pass, `.cartulary/test-results/20261001T011454Z-p98171` |
| Structural toolchain drift | Pass 2/2, `.cartulary/test-results/20261001T011612Z-p7283` |
| Explicit lockfile preparation | Pass, `.cartulary/test-results/20261001T011753Z-p9425`; initial helper label was subsequently corrected from adhoc to its public target identity. |
| Frozen frontend installation | Pass, `.cartulary/test-results/20261001T011808Z-p9728` |
| Browser/image installation | Pass, `.cartulary/test-results/20261001T011836Z-p10345` |
| Generation after contract migration | Pass, `.cartulary/test-results/20261001T013135Z-p46691` |
| Frontend typecheck | Pass 2/2, `.cartulary/test-results/20261001T013234Z-p50467` |
| Script lint | Pass 2/2, `.cartulary/test-results/20261001T013234Z-p50479` |
| Shell lint | Pass 4/4, `.cartulary/test-results/20261001T013234Z-p50485` |
| Command-surface owner slice | Pass 1/1, `.cartulary/test-results/20261001T013254Z-p84819` |
| Review skill validation | External skill-creator quick validator passed. |

## Failed attempts and corrective accounting

| Target and root | Classification and relationship | Correction |
| --- | --- | --- |
| generate, `20261001T011233Z-p8405` | Task-owned registry omission | Added observability/help registration for the new helper; later generation passes. |
| test-slice harness.browser, `20261001T011307Z-p11695` | Mixed preparation failures during implementation | Old Playwright did not support the removed Ubuntu override; generation overlapped source sealing. The upgrade removes that obsolete requirement. Later qualification freezes executable inputs. The corrected lifecycle fixtures subsequently passed in the complete 48/48 browser-owner run `20261001T024523Z-p7201`. |
| harness-ui-review-contract, `20261001T011633Z-p8210` | Task-owned test expectation | Missing package can report ENOENT after Node resolution caching; assertion now accepts both missing-package forms. Focused rerun passed. |
| generate, `20261001T012215Z-p12075` | Task-owned projection ordering | Sorted the diagnostic resource profile. |
| generate, `20261001T012520Z-p20808` and `20261001T012637Z-p29462` | Task-owned closed resource registry projection | Registered and ordered the diagnostic resource profile in all closed projections; later generation passes. |
| harness-ui-review-workflow, `20261001T012252Z-p15564` and `20261001T012416Z-p17965` | Task-owned Unix socket path limit | Short directory-descriptor socket paths retain private session storage; direct and public workflow reruns passed. |
| frontend-typecheck, `20261001T013137Z-p47297` | Task-owned missing module declarations | The lease now passes its validated profile as private JSON to a typed browser-side consumer. This also preserves the frontend dependency boundary; corrective typecheck passes. |

An attempted direct Make binding for the internal check
`harness-smoke-browser-work-graph` had no target; that check is dispatched by the
harness tier. An `explain-run DETAIL=logs` attempt without TARGET was rejected;
the exact retained logs were then inspected. Neither attempt produced test evidence.

## Migration and rollback

Stop live review sessions through their exact locator with the old implementation
before installation cutover. Replace manifests, lockfile, renderer profile and
generated projections together; recreate installation proof with frozen
installation and install the matching browser/image. Start fresh sessions.
Migrate request callers directly to action v3 and renderer consumers to profile
v2. No live-state translation, database migration or user-data migration is needed.

Rollback package manifests, lockfile, renderer profile, projections, and any
accepted PNG/manifest transaction together. The CLI diagnostic capability can be
reverted separately if its request schema and guidance move together. Preserve
independently validated product fixes. Never delete unresolved cleanup proof.

Private observations, daemon output, caches, registries and endpoints expire with
session detail. Only safe structural receipts belong in retained evidence. The
helper's success makes no product, accessibility, release or publication claim.

## Visual review before promotion

The first upgraded ordinary run, `20261001T013235Z-p50838`, accounted for
255 active captures and 255 goldens, with no orphan, missing or ambiguous mapping.
It failed screenshot comparisons in 17 tests and emitted 59 changed actual images.
Every changed capture was imported by its exact capture ID through private Make-owned
artifact review sessions. All 59 expected/actual/diff sheets were inspected; native
Network Flow and incident-creation components were also inspected. No interactive
HTML report was used. All review sessions were stopped through their exact locators
and returned successful cleanup. One interrupted import batch was resumed by exact
capture IDs; no image was selected by filesystem recency.

Accepted changes are Chromium rendering of native control borders, select arrows,
textarea resize handles and isolated antialiased pixels. Text, state, geometry,
focus location, overflow and responsive composition remain consistent. No product
CSS, vendored font, masking or screenshot tolerance was changed. This review accepts
the following exact candidates for the transactional update; promotion and two
ordinary comparison runs remain separate gates. Hashes below are SHA-256 prefixes;
the retained report, reconciliation and final golden manifest carry full identities.

| Capture intent | Exact capture ID | Previous → reviewed candidate SHA-256 |
| --- | --- | --- |
| `network-flow-analysis-mapping-dialog` | `visual.capture.88c91acf15734aa774a9` | `79f91cc60b25` → `d0743467ff02` |
| `network-flow-analysis-rejected-diagnostics` | `visual.capture.da3392816b3c13320b9d` | `2cef05f69614` → `f484cf910617` |
| `network-flow-analysis-graph-contributors` | `visual.capture.fe36571af4277f846e89` | `c2086d1b06d2` → `d2899a15ce09` |
| `incident-create-expanded-details` | `visual.capture.10c44722b7aa59411f31` | `fc46497cb37a` → `c67f47f648ea` |
| `incident-create-required-errors` | `visual.capture.38385abd2fd326e6b0ca` | `f12175c220ad` → `55aa54a8541e` |
| `incident-create-pending` | `visual.capture.555eb9a04b92ff7930c3` | `754515e6ffdd` → `03ef5f574e6f` |
| `incident-create-confirmed-handoff-failure` | `visual.capture.79e70bdf00ce420831dc` | `2715eb5c6b5c` → `90739deec2c4` |
| `incident-create-recovery-short` | `visual.capture.8c4a9ca3c4fc73840423` | `6ecfaeadfde6` → `7b58df5aacef` |
| `incident-create-initial` | `visual.capture.bbef5adba7dcd870c318` | `9ed5b9f92b1a` → `79895f09dd22` |
| `incident-create-recovery-text-spacing` | `visual.capture.d62db66130daec6e0428` | `1db48f2b383e` → `4f7f34fcd6d8` |
| `account-menu-directory-root` | `visual.capture.15139237bba7908f4c58` | `8f10746df6a0` → `889549fab97c` |
| `account-menu-deployment-root` | `visual.capture.17e0b9ae241d550045ee` | `6e2f0ad301cf` → `bc0a77980a4f` |
| `reference-pack-selection` | `visual.capture.615dce7640c6fa6e27f3` | `73687a1b3cd0` → `360effb64c11` |
| `reference-pack-catalog` | `visual.capture.99cf31a0eceac0687928` | `b3a8b22a3d89` → `e44348eef0b2` |
| `reference-pack-submitted` | `visual.capture.ac2982250ef26c4e4be7` | `a7f43e1b7951` → `abdc7258fb99` |
| `incident-import-canceled` | `visual.capture.57b2cde9b1146309d22c` | `d8c0ff2d5301` → `4b4e3aa65b75` |
| `incident-import-recovery-text-spacing` | `visual.capture.b436c140dac7fb4fb4ab` | `347d57b9a5a3` → `79021827a98e` |
| `membership-audit-inspected-spacing` | `visual.capture.5574ccaedbb0e2586598` | `19b6bbf21852` → `bfd375617742` |
| `membership-management-pending` | `visual.capture.278029016c6b60a9745a` | `fc916ab3d11a` → `1d64f5eed17d` |
| `membership-management-confirmed-refresh-failure` | `visual.capture.65f37803d46189a11a85` | `d676418ea05b` → `7483affc9d77` |
| `membership-management-loading` | `visual.capture.7a5985aa2e474a629d10` | `0550bb2496cc` → `ac0bd14eece8` |
| `membership-management-role` | `visual.capture.97d5ba9d8a7cde127c5c` | `7deaa88a617a` → `5530091a8a9c` |
| `membership-management-uncertain` | `visual.capture.b11f46b442198c4d0995` | `0bd8e1219008` → `3c634766b300` |
| `membership-management-comfortable` | `visual.capture.b6fae0a185503be4b49e` | `6e26d3a5c4ba` → `9e339b49c174` |
| `membership-management-compact` | `visual.capture.be1865e801a7fdab4caa` | `a82c46248109` → `1ded676ce0c3` |
| `membership-management-removal-spacing` | `visual.capture.f94b83236deaeebca37d` | `3f63293f6550` → `3690e043ae51` |
| `metadata-saving` | `visual.capture.048441e5821b032d0f67` | `ccc3d8e7ffea` → `22d31611a49b` |
| `metadata-loading` | `visual.capture.3b6710f5709fc093a054` | `60fd4ed30792` → `056e69d0d9e7` |
| `metadata-dirty` | `visual.capture.4030c6a79227345943e0` | `5614bae9650f` → `16d644cfaecf` |
| `metadata-confirmed-refresh-failure` | `visual.capture.59b95b912d49f30a9759` | `45e298ce3521` → `5cad43e76a4b` |
| `metadata-review-spacing` | `visual.capture.68e190a75befd7f69f33` | `ee86953e3707` → `e653ee7a2c86` |
| `metadata-compact` | `visual.capture.e2466c4700ff818cf4aa` | `1e354d9dfa7b` → `984e182f411d` |
| `metadata-comfortable` | `visual.capture.f2768bd50460059330f7` | `3ba4087b6963` → `5c9677574734` |
| `lifecycle-reason` | `visual.capture.048bb8a82f5679c1acf3` | `a03c16a3b944` → `e8b296bb46f7` |
| `lifecycle-confirmed-refresh-failure` | `visual.capture.35f4b8ca4813aa639b4d` | `33ca985bc1b6` → `466778b82224` |
| `lifecycle-closed` | `visual.capture.864872af69315a1b0719` | `114bc258a1a8` → `fe62102df91b` |
| `timeline-supersession-authoring` | `visual.capture.5fb1a56e0f004366997f` | `71493d2fae2e` → `1290d2bdb7ef` |
| `indicator-lifecycle-authoring` | `visual.capture.24cdae4ee2f0566b9bad` | `7e770ba93f88` → `884772d328eb` |
| `indicator-lifecycle-authoring-narrow` | `visual.capture.24d4875734ba06b1914f` | `5e1de808c906` → `3cbcf465ee17` |
| `indicator-observation-authoring-narrow` | `visual.capture.42d65cdf1d6d6cc7a834` | `6bff834c7f2c` → `5ccb7a8484a1` |
| `indicator-observation-authoring` | `visual.capture.e084a3414b80dbabab3d` | `6d90b3590561` → `c5f404135665` |
| `linked-note-authoring` | `visual.capture.9bb90770e53a11be45c3` | `1358522b82c5` → `8b9dfafb4a52` |
| `linked-note-authoring-narrow` | `visual.capture.f49e759c383c4e0f1f5a` | `66d3553081cd` → `13875b60421b` |
| `coordination-status_review-authoring` | `visual.capture.ca30bc4813c00e464e4a` | `c5a03f72a462` → `c58cb09a8242` |
| `incident-directory-narrow-desktop-workbook-shell` | `visual.capture.474f3c17ecf77ef9d8b2` | `4f4121b61a10` → `c3f17f4e4f1d` |
| `incident-directory-compact-desktop-workbook-shell` | `visual.capture.5895c38a72f7e019fbed` | `77873eb2256d` → `860be9f3fec1` |
| `workbook-view-bar-maximum-pressure-narrow` | `visual.capture.000d8edec05917c8680f` | `ca87188e60d4` → `28f69d3306e1` |
| `workbook-view-bar-long-columns` | `visual.capture.1f0c5e239a5b50226642` | `41b88a864cfd` → `35b8590f56a1` |
| `workbook-view-bar-maximum-pressure-base` | `visual.capture.255f84878da966bb6a2f` | `f0848435bc99` → `2c0603086699` |
| `workbook-view-bar-maximum-pressure-compact` | `visual.capture.31051420009eef38c992` | `f4fce7f84792` → `88ca585cdc53` |
| `workbook-view-bar-saved-view-modified` | `visual.capture.41ad8eafac6d081ac683` | `f0848435bc99` → `2c0603086699` |
| `workbook-view-bar-saved-view-clean` | `visual.capture.4f5e321cc67de64bdcfa` | `a5bde41ebdee` → `bb3e7607d25e` |
| `workbook-view-bar-text-spacing` | `visual.capture.82fbeaf5595c1c0b0e18` | `8223d943d750` → `c0e1eaa35a34` |
| `workbook-view-bar-filter-editing-overflow` | `visual.capture.9dece7e0dfcaa9014a74` | `73a38bd36858` → `13a5689df56b` |
| `workbook-query-saved-view-query-controls` | `visual.capture.b6524c89920690198ffb` | `f0848435bc99` → `2c0603086699` |
| `workbook-view-bar-ordered-maximum-sort` | `visual.capture.d994111394c03b53cee4` | `fd39582f8d71` → `8e6b4328758f` |
| `workbook-view-bar-saved-view-actions` | `visual.capture.f08b4d814d30429d51f3` | `ecd362b57439` → `c53cdc411f9a` |
| `workbook-inspector-attached-edit` | `visual.capture.bec04aee967bd2e5d893` | `1ca4cb7330a8` → `3af1f676dfbf` |
| `workbook-inspector-narrow-technical-details` | `visual.capture.f5ca1ba5368f9538cbb9` | `a2423948fee5` → `6dc0f9fe22c5` |

## Additional qualification findings

The full 1.63.0 visual report exposed two evidence-adapter defects: upstream
normalizes/truncates named screenshot attachments, and aggregate reports/opaque
traces were incorrectly constrained by the 8 MiB private JSON-component cap.
The exact test/capture join is retained; the pinned adapter maps upstream names
within that result and rejects collisions. Producer reports and opaque traces
have explicit 32 MiB bounds; bundle/session limits stay unchanged. No ZIP
extraction or historical-reader fallback was added.

Frontend qualification at `20261001T015811Z-p58770` found pre-existing ownership
bookkeeping omissions in unchanged files: `WorkbookIncidentIdentityDisclosure.test.tsx`
was absent from the source owner projection, and authoring tests used unowned
selector literals. The owner projection now registers the existing file. Tests
use accessible named controls and semantic fixture groups, preserving assertion
intent without expanding selector exceptions or changing product behavior.
That run was interrupted after 346/690 successful units and is not acceptance
evidence. Its replacement run must finish before handoff.

`make agent-finalize` passed at `20261001T020701Z-p44010`. Retained-run maintenance
was skipped because RESULTS_DIR was unset; no eligible full warm check was
available at that point.

## Promotion review and deterministic focus correction

The first transaction passed 12/12 at `20261001T020740Z-p51596`, promoting 58 PNGs.
Fifty-five matched the previously reviewed bytes exactly. A second private Make
artifact session compared the three differing candidates as explicit image
references: import recovery spacing differed by 7 pixels, Indicator authoring by
8, and Network Flow contributors by 74. Native exact diffs showed isolated
antialias/control-edge differences, with no content or geometry change. These
candidate variations were reviewed and accepted without changing tolerances.
The session stopped with successful cleanup.

The canceled-import image was not promoted in that transaction. Its 797-pixel
variation included the Refresh job status focus ring, revealing nondeterministic
focus after replacement of a previously disabled control. This supersedes the
initial assumption that every difference was browser rasterization. The visual
fixture now explicitly establishes keyboard focus on that recovery control and
asserts focus and enabled state before and after capture. The corrected fixture
required the further promotion and two ordinary passes recorded below. It did
not alter product focus behavior, hide the control, or weaken screenshot checks.


## Later qualification and recovery ledger

| Target and retained root | Finding and corrective status |
| --- | --- |
| `test-slice OWNER=harness.browser`, `20261001T013237Z-p51029` | 46/48 passed. The lifecycle fixture's engine-block hook accidentally matched the new local resolver filename; it now matches package imports. The seeded Network Flow preparation failure passed on focused rerun; its original cause was not established. |
| `service-backed-test-slice OWNER=harness.browser`, `20261001T013909Z-p82222` | Passed 17/17. |
| `browser-e2e-visual`, `20261001T013235Z-p50838` | First upgraded ordinary comparison: 9/12 units, 59 changed captures across 17 tests. Reconciliation accounted for all 255 active captures and goldens. Pixel review and promotion are recorded above. |
| `lint-biome`, `20261001T013912Z-p82687` | Formatting failures corrected by the repository formatter. |
| `harness-contract`, `20261001T013911Z-p82493` | Importing the harness MJS profile reader into browser TypeScript broadened the source closure. Replaced with the typed private profile-data boundary. |
| `harness-contract`, `20261001T014423Z-p88087`, `20261001T015150Z-p30284` | Generated module hashes were stale during implementation; regenerated after source changes. |
| `harness-ui-review-contract`, `20261001T015845Z-p72126` | The new font-file schema excluded brackets and commas used by existing variable-font filenames. Corrected the structural pattern; contract rerun passed at `20261001T020043Z-p18561`. |
| `frontend-unit`, `20261001T015811Z-p58770` | Interrupted after 346/690 passing units. Confirmed pre-existing selector/source-owner bookkeeping gaps were corrected separately. Interruption origin was not established. |
| `format`, `20261001T020434Z-p32239` | Accessibility lint rejected generic fixture divs with group roles. Replaced them with named fieldsets; formatter passed at `20261001T020609Z-p39435`. |
| `frontend-unit`, `20261001T020740Z-p51511` | 689/690 passed. One corrected selector used listbox for a single-select control; changed it to combobox. Focused owner rerun passed 2/2 at `20261001T021200Z-p83841`. |
| `frontend-unit`, `20261001T021536Z-p17027` | Interrupted after 540/690 passing units, without a reproduced assertion failure. This is incomplete evidence, not acceptance. |
| `harness-ui-review-workflow`, `20261001T021521Z-p14384` | All four direct diagnostic cases passed. The public workflow hit resource admission contention during concurrent bulk qualification. Isolated rerun passed at `20261001T022440Z-p88001`. |
| `browser-e2e-webserver-backed`, `20261001T021538Z-p17474` | Range-scroll and file-picker assertions failed; other groups were interrupted. An interrupted runner left unpacked temporary trace resources in retained output, correctly rejected by the privacy scan. Process-group cleanup now drains descendants and removes only pinned producer scratch before retention. |
| Timeline range-scroll owner slice, `20261001T022630Z-p94085` | Passed 11/11 unchanged. The earlier timing failure was not reproduced in isolation; final integrated qualification is still required. |
| Evidence picker-promotion owner slice, `20261001T022630Z-p94082` | Reproduced editor disappearance: the picker action's horizontal scrolling unmounted the virtualized synopsis editor. The fixture now places its relevant column beside the action through existing user controls. |
| Evidence picker-promotion owner slice, `20261001T023447Z-p73498` | Corrective rerun passed 11/11, preserving focus, no premature write, one promoted row, one zero-byte upload, and attachment to the invoking source. |
| `frontend-typecheck`, `20261001T023512Z-p4268` | The new Testing Library role selectors mistakenly supplied Playwright's `exact` option. Removed that unsupported option; Testing Library matches string names exactly by default. Corrective typecheck passed 2/2 at `20261001T024005Z-p48299`. |

The temporary-trace cleanup has behavior checks for idempotence, exact producer
names, preservation of completed attachments, and rejection of symlinked roots.
The shared process owner retains PID/boot/start proof and drains a detached group
after its leader exits; recovery tests reject reused PIDs and verify descendant
termination. No global process, container, CLI-session, or trace cleanup is used.

At this point the canonical Linux amd64 container and Ubuntu 26.04 WSL2 host have
runtime evidence. An independent native Linux host is unavailable; this remains
an explicit blocker for any claim of separate native-host qualification.


The corrected picker-promotion fixture passed at `20261001T023447Z-p73498`.
Typecheck, Biome, script lint and shell lint passed at respectively
`20261001T024005Z-p48299`, `20261001T024028Z-p48850`,
`20261001T024032Z-p49349` and `20261001T024035Z-p49787`.
Lifecycle qualification passed at `20261001T024040Z-p50962`, including the new
failed-publication and missing-executable acquisition cases.

The second promotion passed 12/12 at `20261001T023542Z-p7272`. Private reference
comparisons reviewed both changed PNGs: canceled import (787 changed pixels,
principally the expected focus ring) and unavailable observation (768 pixels,
only the Retry observation focus ring). This exposed the same disabled-control
replacement race in the unavailable state. All refreshed import captures now
explicitly establish keyboard focus on their observation recovery control. The
final promotion and ordinary comparisons recorded below establish acceptance for
that complete fixture correction. The private review session's terminal receipt
reported successful exact-session cleanup.

Generation passed at `20261001T024157Z-p55068`; `agent-finalize` passed at
`20261001T024208Z-p58026`. RESULTS_DIR was unset, so retained-run maintenance was
skipped. These precede the final focus-fixture consolidation.


The full browser harness owner slice passed 48/48 at
`20261001T024523Z-p7201`. This includes direct diagnostic and seeded/default/
Network Flow/borrowed/artifact workflow tests, renderer/profile/package mismatch
checks, private import/evidence tests, and lifecycle/recovery cases.


## Final reviewed visual transaction

The consolidated promotion passed 12/12 at `20261001T024359Z-p71781`.
Only `incident-import-running-determinate-linux.png` changed from the prior
reviewed transaction: 809 pixels, principally the declared Refresh job status
focus ring with a few isolated antialiasing edge pixels. A private reference
comparison and native-size source inspection accepted it; the other images were
byte-identical to their already reviewed candidates. Its exact review session
stopped successfully and purged private outputs.

The final manifest SHA-256 is
`5155ea46750774fda1900edbc92690c8fecf923bbb4b85e5ab119a1516282d82`.
All 61 changed PNGs are covered by that manifest and the review records above.
The accepted trigger is the pinned Chromium/image upgrade plus the explicitly
corrected import-observation keyboard-focus setup. Viewports, masks, tolerance,
comparison settings, fonts and product behavior are unchanged. Reconciliation
accounts for 255 active captures and 255 committed goldens, with no orphan,
missing or ambiguous mapping. The two ordinary acceptance run roots are recorded
below; update-mode success alone is not acceptance.

Affected authored rows and fixture mappings from that exact reconciliation:

| Row ID | Changed PNGs | Fixture IDs |
| --- | ---: | --- |
| `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | 3 | `visual.fixture.claimed_network_analysis_workspace_states` |
| `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | 11 | `visual.fixture.saved_view_query_controls_and_grouped_result` |
| `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | 2 | `visual.fixture.compact_desktop_workbook_shell`, `visual.fixture.narrow_desktop_workbook_shell` |
| `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | 2 | `visual.fixture.base_inspector`, `visual.fixture.inspector_narrow_technical_details` |
| `module.workbook.visual.coordination_create_authoring_recovery` | 1 | `visual.fixture.contextual_coordination_creation` |
| `module.workbook.visual.indicator_lifecycle_authoring` | 2 | `visual.fixture.indicator_lifecycle_authoring` |
| `module.workbook.visual.indicator_observations_authoring` | 2 | `visual.fixture.indicator_observations_authoring` |
| `module.workbook.visual.note_create_authoring_recovery` | 2 | None; exact active capture mapping |
| `module.workbook.visual.timeline_capture_actions` | 1 | None; exact active capture mapping |
| `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | 2 | None; exact active capture mapping |
| `web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde` | 7 | None; exact active capture mapping |
| `web.design.visual.incident_import_workflow` | 4 | None; exact active capture mapping |
| `web.design.visual.lifecycle` | 3 | None; exact active capture mapping |
| `web.design.visual.membership_audit_browsing` | 1 | None; exact active capture mapping |
| `web.design.visual.membership_management_visual` | 8 | None; exact active capture mapping |
| `web.design.visual.metadata_editing` | 7 | None; exact active capture mapping |
| `web.design.visual.reference_pack_administration` | 3 | None; exact active capture mapping |

The two captures added by focus qualification are:

- `incident-import-observation-unavailable`: `visual.capture.547fe3643cd6bf2d1efd`.
- `incident-import-running-determinate`: `visual.capture.757739eb08f67f662e7e`.


## Earlier gate evidence before integrated corrections

| Gate | Result | Exact retained root |
| --- | --- | --- |
| Browser harness owner | 48/48 passed | `20261001T024523Z-p7201` |
| Command-surface owner | 1/1 passed | `20261001T024824Z-p22001` |
| Generation | Passed | `20261001T024954Z-p25709` |
| Toolchain drift | 2/2 passed | `20261001T025005Z-p28720` |
| JSON shape | 3/3 passed | `20261001T025007Z-p29140` |
| Generated-artifact policy | 3/3 passed | `20261001T025014Z-p30322` |
| Generation drift | 4/4 passed | `20261001T025018Z-p30812` |
| Harness contract | 2/2 passed | `20261001T025033Z-p35063` |
| Agent finalization | 1/1 passed; RESULTS_DIR unset, retained-run maintenance skipped | `20261001T025103Z-p42104` |
| Ordinary visual comparison 1 | 12/12 passed | `20261001T025059Z-p38578` |
| Markdown maintenance | Passed; helper uses the existing adhoc summary label | `20261001T025227Z-p77942` |
| Ordinary visual comparison 2 | 12/12 passed | `20261001T025549Z-p80860` |

Both ordinary runs reconcile the exact same final manifest SHA-256 shown above:
255 captures, 255 active goldens, zero missing/orphan/ambiguous mappings, and all
29 registered fixtures resolved. Screenshot comparisons were enabled; browser
units are uncached by policy. These passes precede the functional corrections below; final visual acceptance
will be repeated on the corrected source. Each successful
browser group includes verified exact renderer release; no cleanup error was
suppressed as success.

## Private review cleanup

All ten exact review sessions terminated with `cleanup=complete`. Their private
image bundles and reports expired at stop. Caller-owned request and comparison
scratch is removed after consumption; canonical run evidence and committed
goldens remain. Terminal receipt identities:

| Review session | Status | Receipt SHA-256 |
| --- | --- | --- |
| `pw-upgrade-review-1` | `ok`; `complete` | `05ca719a44e65216160695904f4fc03f85c46fc03cd0777c9eb12606d5098c4b` |
| `pw-upgrade-review-2` | `ok`; `complete` | `c4fed14e43c5a414b3c1937f0b75e0e412985d544a07d39c5628e1d86e1576d9` |
| `pw-upgrade-review-3` | `ok`; `complete` | `5e79fe32637f8545195814500f9ba49be61843f66d3f0dfe211b88ffba6bcaf6` |
| `pw-upgrade-review-4` | `ok`; `complete` | `aa9279659bd86871fc4b2a4a2cd43ebf9e1d7f7eb0aa062ef356e57c7de3492f` |
| `pw-upgrade-review-5` | `ok`; `complete` | `7ad468ec49d34c7696da0b37602087d85666db4ad4ca61a40b7fd760423cd1a1` |
| `pw-upgrade-review-6` | `ok`; `complete` | `8f5f69a9522c153d9eefebe46dde37c1198b12f0b14dea665ad1f1676aa4feb9` |
| `pw-upgrade-review-7` | `ok`; `complete` | `cfb42b99d9d89a475a8bab5479a0ab5f1de15e71897142cfdde3b683fa02db4c` |
| `pw-promotion-review` | `ok`; `complete` | `69be7a79b80556ea38444ad54e6a5ba17481a01c36e868c63fc24753d9c71e30` |
| `pw-corrective-review` | `ok`; `complete` | `831dbb0b0cad071482409abe8d4affcbb4aae03e1f4fad69fc0bd99e0c63a9a6` |
| `pw-final-promotion-review` | `ok`; `complete` | `8dd0defa4965b8a0a63788fc741f06ba586865c515f6c0822a5c17246930a647` |


## Integrated functional qualification corrections

The full webserver-backed run `20261001T030104Z-p17572` completed 129/144
units and reproduced 24 failed rows. This is separate functional evidence, not
proof that every failure was introduced by the browser upgrade. Exact row
results and producer reports remain under that root. Current corrections:

- Shared overlay navigation preserves ownership through native disabled-control
  blur, while intentional focus departure still dismisses the overlay. Saved-view
  recovery enables the existing successor reconciliation. All three saved-view
  browser scenarios pass at `20261001T031926Z-p71388` (11/11 units).
- Inspector actions reveal the control without moving saved-value context under
  the pointer before click. Primary editors still reveal their field context.
  Timeline Find passes at `20261001T032105Z-p13307`; Party editing passes at
  `20261001T032203Z-p46839` (both 11/11). Clipboard continuity/composition also
  passed inside `20261001T032808Z-p34756`, whose batch-history and disclosure
  groups still failed.
- The Timeline pending-create fixture explicitly re-reveals its virtualized
  synopsis column after native action hover. Its exact text and write-count
  assertions pass at `20261001T030630Z-p71432` (11/11).
- Canonical Indicator linking distinguishes a pending request from an uncertain
  terminal outcome; the held-request regression and recovery/inspector unit
  slices pass at `20261001T032829Z-p66848` (4/4). Reference-selection and canonical
  create browser slices pass at `20261001T032642Z-p43732` (12/12); the pending
  feedback correction needs the final integrated rerun.
- Mention discovery fixtures migrate native-select assumptions to the current
  searchable chooser, including exact selected target, stale-selection fencing,
  read retry and no-write assertions. Read-only tag disclosure retains a
  programmatically focusable note; mutable disclosure focuses its Remove control.
- History preview focus and recovery continuation corrections and focused evidence
  are recorded below.

Additional attempted checks: `web.shared` was not an active test owner;
registration resolves to `web.application`. The shared focus unit slice passed
at `20261001T031756Z-p61613`. The new inspector geometry test first failed on a
read-only jsdom prototype method at `20261001T031855Z-p63047`; instance property
fixtures corrected it, passing at `20261001T031950Z-p5301`. Generation passed at
`20261001T031915Z-p68387`.

History-focused attempts `20261001T032505Z-p80363`,
`20261001T032744Z-p87467`, and `20261001T033016Z-p83850` reproduced preview
cancellation. Entity discovery/merge run `20261001T032607Z-p13456` failed on
retired discovery controls and the same preview cancellation. The Evidence slice
`20261001T032740Z-p83508` failed source sealing because source changed during
build; it is invalid qualification, not a product failure. Formatter attempt
`20261001T033127Z-p17356` reported a source lint error during these edits.


History preview cancellation is corrected by observing native disabled-control
blur in the existing History focus owner, retaining the exact invoking element,
and yielding the shell fallback to a detail-owned successor. React does not
reliably deliver that blur during the disabling commit. The unchanged Entity
history recovery scenario passes at `20261001T033828Z-p30166` (11/11). Temporary
focus probes were removed. All four merge recovery scenarios also pass within
`20261001T034015Z-p70784`; that run failed only its still-stale discovery fixture.

Discovery now intercepts the dedicated GET entity-candidates API and uses its
candidate envelope and search/listbox controls. Its real invalid cursor response
also exposed an implementation gap: `invalid_pagination_request` has no validated
reason-code registry, so continuation recovery must recognize the public error
code itself. The owner and its regression fixture now do so. All discovery
scenarios pass at `20261001T034305Z-p15882` (11/11), including actual target
identity, independent retries, revalidation and read-only paging.

The mutable tag disclosure focuses its existing Remove action; a read-only tag
now has a programmatic focus destination without adding a tab stop. The Undo
continuity fixture has sufficient rows for the exact user scroll offset to remain
reachable after notice removal increases the grid height; its equality and
newer-focus assertions remain unchanged.

History, canonical linking, recovery navigation and inspector unit slices pass
6/6 at `20261001T034043Z-p1608`. Formatting and generation pass at
`20261001T034248Z-p8305` and `20261001T034255Z-p12887`; frontend typechecking
passes 2/2 at `20261001T034355Z-p49033`.

Diagnostic-only attempts `20261001T033326Z-p22494` and
`20261001T033526Z-p56278` reproduced and isolated the native blur issue.
`lint-biome` at `20261001T033327Z-p22786` rejected an unnecessary React blur
handler on an unlabelled section; that handler was removed in favor of the
existing native-event owner. Discovery attempt `20261001T033640Z-p89299` still
used the retired endpoint and failed; `20261001T034015Z-p70784` exposed the
public pagination-code handling above. Neither is acceptance evidence.


The final focused Timeline browser batch passes 13/13 at
`20261001T034528Z-p88506` (batch-history review, disclosure/read-only focus,
and all Undo continuity scenarios). Evidence recovery still reproduced at
`20261001T034402Z-p49765`: its conflict resolver unmounts after settlement while
the selected parent creation remains. The prior unmount guard only allowed a
removed selection and canceled the valid return-to-grid continuation. The
ownership check now accepts the unchanged activation with either its retained
parent or removed selection, while preserving all newer-intent cancellation.
The added retained-parent unit case failed before the fix at
`20261001T034735Z-p36686`. Its initial test-edit syntax error at
`20261001T034709Z-p35759` was corrected before that behavioral reproduction.

Additional successful checks: candidate recovery unit slice at
`20261001T034411Z-p60992` (2/2); collection inspection at
`20261001T034420Z-p76702` (2/2); lint at `20261001T034419Z-p75725`;
import boundaries at `20261001T034424Z-p80002`. Agent finalization at
`20261001T034510Z-p84561` passed with retained-run maintenance skipped because
RESULTS_DIR was unset. Toolchain drift, JSON shape, generated policy, generation
drift and harness contract passed respectively at `20261001T034530Z-p88756`,
`20261001T034533Z-p93165`, `20261001T034540Z-p5830`,
`20261001T034544Z-p14989`, and `20261001T034600Z-p24577`.


The retained-parent unit correction passes at `20261001T034825Z-p45256`
(3/3 with candidate recovery). Browser attempts `20261001T034834Z-p46003`
and instrumented `20261001T035307Z-p79684` isolated a second cause: the original
Evidence collection field is hidden, so its cell cannot receive focus. The
resolver now uses the existing active-grid fallback required by Design §8/§10.4
only after an unavailable cell result and a fresh continuation/viewport check.
Cancellation and newer user intent still prevent restoration. The expanded unit
matrix passes at `20261001T035513Z-p20713` (2/2). All temporary browser probes
were removed. Formatting and generation passed at `20261001T035457Z-p13309`
and `20261001T035503Z-p17714`.

The corrected Evidence browser slice passes 11/11 at
`20261001T035521Z-p21276`, including conflict settlement, retained target
identity, exact retry and history accounting. Finalization passes at
`20261001T035528Z-p28848`; retained-run maintenance remains skipped because
RESULTS_DIR is unset.


The first integrated rerun at `20261001T035548Z-p55696` passed 141/144
work units. Two browser groups failed (plus their aggregate summary): native
scalar Undo continuity and Party source-link recovery. These are fixture races,
not accepted implementation regressions. Undo now uses enough filler rows for
its original exact scroll offset to remain reachable after notice removal.
Party recovery waits for the first authoritative link receipt before recording
its pre-replay history; request arrival alone did not prove that write committed.
Exact scroll, local editing, replay bytes, receipt identity and history equality
assertions remain intact. The chained stateful, measurement and accessibility
targets did not run after that failed webserver gate.

Additional qualification uses two concurrent public review workflows with
distinct private sentinels, proving separate CLI attachments and retained-detail
isolation. The canonical browser font-readiness check is now a shared browser
function in `visualRenderer.ts`; the harness exercises real valid font loads,
a malformed FontFace that reaches error state despite FontFaceSet.ready, and
missing faces. The same function gates ordinary visual captures.


Formatting and generation pass at `20261001T040956Z-p49043` and
`20261001T041003Z-p53462`. Typechecking passes 2/2 at
`20261001T041013Z-p56499`; the corrected Party recovery browser slice passes
11/11 at `20261001T041018Z-p56923`. The enhanced private workflow target passes
at `20261001T041037Z-p84025`, including concurrent exact-page diagnostics with
distinct sentinels and actual failed/missing browser font-load rejection.


The corrected complete Undo characterization browser slice passes 11/11 at
`20261001T041126Z-p1307`. Agent finalization passes at
`20261001T041149Z-p31437` (RESULTS_DIR unset; retained-run maintenance skipped).
The subsequent webserver/stateful runs exposed two further fixture gaps: the
sentinel sampled virtualized rows immediately after mutation-substrate readiness,
and the stateful mention lifecycle still used native-select commands against the
searchable candidate chooser. Their failed roots and corrective evidence are
recorded with the final gate results below.


Stateful qualification at `20261001T041343Z-p76085` passed 40/42 work
units. The only failed browser group is mention lifecycle (plus its aggregate
summary): its three native-select interactions target the current searchable
combobox. Authentication, collaboration, recovery, Evidence integration,
coordination/public routes, workbook querying and all Network Flow groups pass.
Measurement and accessibility did not run in that failed chain and were started
separately afterward.


The webserver rerun `20261001T041208Z-p36697` passed 142/144 units:
only the sentinel virtual-row readiness race and its summary failed. The fixture
now awaits its two saved rows before reading their exact identities. The three
stateful mention-lifecycle selection steps now use the current searchable chooser
and still assert the exact resolved record IDs in public mutation receipts.

Measurement attempt `20261001T042223Z-p68614` passed 19/24 units; all four
Timeline predicates failed while establishing background presence, before any
measurement sample. Its retained browser console identifies
`ERR_BLOCKED_BY_LOCAL_NETWORK_ACCESS_CHECKS`. Chromium treated the intercepted
synthetic navigation as lacking a loopback document's address-space proof.
[Chrome's network-access model](https://developer.chrome.com/blog/local-network-access?hl=en)
explains the boundary; no performance budget was changed. The measurement build
now seals a minimal `presence.html` entry that the owned frontend
actually serves. It mounts no application and needs no permission override or
security-disabling browser flag. Background contexts enter cleanup ownership
immediately after acquisition, including login/handshake failures.

The concurrent diagnostic workflow's row now reserves the dedicated
`ui_review_browser_pair` resource profile: two browsers, two CPU/IO units,
3072 MiB and four helper process units. Its projection is explicit in the authored
topology/catalog and closed schemas, preventing undeclared parallel consumption
when integrated checks schedule it beside other browser owners.


Generation attempt `20261001T043029Z-p73378` rejected the initial hyphenated
presence-entry name under the existing closed HTML-entry grammar. The entry is
now `presence.html`; the grammar was preserved. Generation passes at
`20261001T043150Z-p76677`, typechecking passes 2/2 at
`20261001T043200Z-p79717`, and the concurrent diagnostic workflow passes through
its scheduler-owned row at `20261001T043221Z-p80229` (1/1). The corrected
sentinel browser slice passes 11/11 at `20261001T043225Z-p80683`.


The corrected stateful mention lifecycle passes 11/11 at
`20261001T043320Z-p24521`. Isolated measurement qualification passes all 24/24
work units at `20261001T043414Z-p57258` with unchanged performance budgets,
including all four Timeline predicates and Network Flow measurements. This
confirms that the real served presence document resolves the Chromium network
boundary without expanding browser permissions.

Accessibility qualification at `20261001T043904Z-p9425` passed 18/20 units.
The saved-view preference shortcut lost focus while its asynchronous request was
pending; its aggregate summary also failed. This is a task-owned regression from
enabling overlay reconciliation. Connected `aria-disabled` controls retain DOM
focus by design, unlike removed or natively disabled controls. The focused-button
assertion remains authoritative and is not relaxed.

The webserver attempt `20261001T044430Z-p40476` also exposed a fixture cleanup
race in the Timeline supersession handoff: selective route removal returned while
an asynchronous fetch/fulfill handler was still active. Cleanup must drain those
handlers before browser teardown. The public mutation and focus assertions are
unchanged; this is fixture lifecycle correction, not a product behavior change.
That run finished 142/144: the affected browser group and its aggregate summary
were the only failures. The fixture now uses Playwright's documented
[waiting route cleanup](https://playwright.dev/docs/api/class-page#page-unroute-all)
before context teardown.

Integrated qualification before those last two corrections passed `make check`
at `20261001T044054Z-p51160` (983/983), including the complete frontend unit
inventory, static/contract checks and both seeded review workflows. Its preceding
`agent-finalize` passed at `20261001T044033Z-p46598`; retained-run maintenance
was skipped because RESULTS_DIR was unset. The stateful browser target passed
42/42 at `20261001T044430Z-p40485`.

The final overlay correction distinguishes activation availability from focus
ownership. A connected, focused `aria-disabled` control keeps focus while its
request settles. Keyboard navigation retains its position in the form, skips
unavailable destinations, and respects deliberate focus departure. Native
disabling and removal still reconcile to an eligible successor. A shared-hook
regression covers pending and settled states, both Tab directions, boundary
wrapping and deliberate departure; the original browser accessibility assertion
is unchanged.

Renderer attestation equality now compares typed values rather than serialized
property order. A reordered profile and font-file object remains valid, while
changed identities or font hashes still fail. This removes incidental JSON
serialization coupling without weakening active-profile validation.

These final corrections pass formatting (`20261001T050310Z-p53754`, 2/2),
generation (`20261001T050316Z-p58345`), the focused overlay unit row
(`20261001T050326Z-p61343`, 2/2), and the direct Make-owned renderer/review
contract check. The corrected supersession browser slice passes 11/11 at
`20261001T050349Z-p62848`.

Accessibility now passes all 20/20 work units at `20261001T050350Z-p63058`,
including the unchanged saved-view focused-control assertion. Finalization
passes 1/1 at `20261001T050734Z-p35932`; RESULTS_DIR remains unset and
retained-run maintenance is explicitly skipped. Executable inputs were frozen
after these corrections for final integrated and visual acceptance.

The integrated attempt `20261001T050907Z-p41149` reproduced an intermittent
`target_unavailable` failure in one of the concurrent public diagnostic sessions.
Its direct same-page diagnostic cases passed. The original private detail expired
with exact-session cleanup, so its precise failing stage was not retained. Review
identified a concrete admission race: a short-lived CLI command can exit while
its exact PID guard waits for the admission transaction, which rejects dead
guards. The diagnostic adapter must distinguish a proven expired guard from a
failed registration of a live process; it must not retry against a reused PID or
relax the general admission contract.

Additional fault coverage addresses asynchronous spawn failure before PID
publication, publication callback failure, and expired versus live admission
guards. One bounded latest private failure record supplies stage and controlled
error categories for future diagnosis; raw errors and command output remain
private and expire with session stop. These lifecycle corrections do not change
product behavior, canonical capture, the reviewed golden transaction or public
result envelopes.

That integrated attempt finished 982/983; the concurrent diagnostic workflow
was its only failed unit. The chained harness-contract target did not run.
Both fresh ordinary visual gates passed: `20261001T050907Z-p41104` and
`20261001T051545Z-p67049`, 12/12 each against the unchanged final manifest.

The webserver attempt `20261001T050907Z-p41085` exposed a separate database
reset failure at the indicator-lifecycle → enterprise-auth boundary. The exact
reset attempt records `recovery_reset_target_release_failed`, no SQLSTATE and
no timeout; reset state verification had completed before release failed. The
process-lease monitor cancels its in-flight PostgreSQL proof query during stop,
which can invalidate the connection or transition ownership immediately before
unlock. The remedy drains that proof under its existing loss-detection deadline
and then releases the lease; actual failed proof/release remains a failure.
This is a qualification-exposed lifecycle correction under `app.server`, with
no database schema or public API migration.

The webserver attempt finished 142/144; the reset boundary and its aggregate
summary were the only failed units. New deterministic process-lease tests
reproduced the bug at `20261001T053003Z-p95529`: shutdown canceled the owned
query, release reported an invalid transition, and the proof's independent
deadline was replaced by cancellation. The implementation now drains in-flight
proof under that deadline before releasing. It still rejects genuinely failed
proof or unlock.

The private browser-binding registry now uses a stable session directory,
separate from disposable per-operation CLI scratch, matching upstream's cached
registry-root lifetime. The registry remains owner-only and expires on session
stop. No additional browser or public diagnostic interface is introduced.

The corrected lease unit row passes at `20261001T053117Z-p4345` (1/1),
and both real PostgreSQL lease rows pass at `20261001T053123Z-p4758` (3/3).
The diagnostic workflow, including spawn/publication faults and expired/live
admission guards, passes at `20261001T053211Z-p23473` (1/1). Formatting and
generation pass at `20261001T053100Z-p96255` and `20261001T053106Z-p1334`.
Finalization passes at `20261001T053239Z-p34076` (1/1); RESULTS_DIR is unset,
so retained-run maintenance remains skipped. The final integrated runs include
these changes; earlier successful results remain useful historical evidence.

Three additional scheduler-owned concurrent diagnostic rounds pass 1/1 each:
`20261001T053336Z-p39414`, `20261001T053423Z-p50421`, and
`20261001T053655Z-p7751`. Each round includes isolated same-page, hostile ambient
configuration, spawn/publication failure, daemon failure, cancellation, expired/live
guard and two simultaneous public-session workflows. No CLI command retry or API
fallback was introduced. The first ordinary visual comparison after the lifecycle
corrections passes 12/12 at `20261001T053336Z-p39651`.

The concurrent diagnostic workflow and renderer/review contract rows also pass
inside the final integrated check `20261001T053336Z-p39701`, exercising the
corrected admission path beside the full verification workload.

The second ordinary visual comparison after the lifecycle corrections passes
12/12 at `20261001T053926Z-p74823`. Both final reconciliations account for
255 captures and 255 committed goldens with no missing or ambiguous entries.

Integrated attempt `20261001T053336Z-p39701` finished 983/984. One Entities
Go batch received SIGKILL after about 44 seconds, with no stdout, stderr or
assertion output. It did not reach the configured ten-minute unit timeout.
Kernel and cgroup evidence shows no OOM kill; the signal's source is unresolved.
The harness's generic product/test-assertion classification is not evidence of
an assertion failure in this case. The four selected Entities rows pass on the
immediate focused service-backed rerun `20261001T055329Z-p62272` (3/3 work
units), without changing Entities code. The chained harness-contract command
was skipped when check failed. A fresh standalone integrated run is required.

Webserver attempt `20261001T053336Z-p39621` finished 142/144. Its only failed
browser group and aggregate summary concern the late Undo body-focus fixture:
it sampled scrollTop 685, then observed 678 after acceptance. The fixture issued
a wheel event and only checked an already-positive scrollTop before sampling.
[Playwright's wheel contract](https://playwright.dev/docs/api/class-mouse#mouse-wheel)
does not wait for native scrolling to finish. The corrected fixture installs a
bounded scrollend observer before dispatch, waits for completion, and moves away
from the bottom clamp before releasing the held Undo response. It proves the
wheel changed position and preserves the exact post-acceptance scroll and body
focus assertions. This corrects fixture synchronization; it introduces no
product scrolling change, tolerance relaxation, screenshot refresh or sleep.

The corrected Timeline Undo browser slice passes 11/11 at
`20261001T055852Z-p96639`; formatting and generation pass at
`20261001T055834Z-p88925` and `20261001T055841Z-p93532`. Finalization passes
1/1 at `20261001T060132Z-p30544`, again with RESULTS_DIR unset. Host doctor
passes at `20261001T060531Z-p21162`.

The subsequent standalone check `20261001T060152Z-p34648` finished 983/984.
The previously interrupted Entities batch passes. Its only failure is
`TestExportJobAuthorizationReDerivesIncidentMembership_Integration`: a tiny
export completed before the test's authorized cancellation, correctly returning
409 `job_cancel_rejected` / `already_terminal` instead of the fixture's expected
200. This is a pre-existing integration-test race, unrelated to browser behavior
or the process-lease fix. The fixture now holds a bounded source-read transaction
barrier, proves the real worker is running, checks visibility and current
membership authorization, requests cancellation, then releases the barrier.
All original response, terminal-state, cancellation-observation and no-success-proof
assertions remain exact. Production job behavior and cancellation policy are
unchanged. The focused corrected slice passes 3/3 at
`20261001T061347Z-p99538`; formatting and generation pass at
`20261001T061330Z-p91916` and `20261001T061336Z-p96609`. The chained
harness-contract and webserver commands did not run after the failed check.

Final integrated qualification passes: `make check` at
`20261001T061512Z-p22265` completes 984/984, and `make harness-contract` at
`20261001T062202Z-p7899` completes 2/2. This includes the corrected export
cancellation fixture and the formerly interrupted Entities batch. The preceding
`make agent-finalize` at `20261001T061452Z-p18193` passes 1/1 with
RESULTS_DIR unset. Check includes the complete fast inventory, frontend units,
typechecking, lint/import checks, typed/generated contract checks and review
workflows; `make explain-target TARGET=test-fast DETAIL=summary` confirms its
inclusion in check, so no redundant standalone fast run was required.

Final webserver qualification passes 144/144 at
`20261001T062232Z-p11652`, including the repaired Timeline Undo group and
all reset boundaries. No browser group, terminal aggregate or cleanup outcome
failed. Markdown maintenance passes at `20261001T063610Z-p99331` after the
completion record was written. Final whitespace validation also passes.

All nine remediation gaps now have implementation and executable validation
on the available host. The baseline is established by the fresh unchanged
renderer run, all accepted image changes are reviewed, and no task-owned defect
or unexplained pixel difference remains. The unexplained earlier SIGKILL is
retained as an interruption with a passing focused rerun and passing final
integrated run, not attributed to an unproven root cause. Separate native Linux
host qualification is the only outstanding support-matrix acceptance item.
No publication or release-conformance claim is made by the helper diagnostics.
