# Browser design readiness overhaul record

Date: 2026-09-23. Implementation began from clean revision
`f2314e68867d3334259003e4e877af36c5191f1f`; verification covers the working-tree
changes described here. This is an implementation handoff and findings record,
not human design acceptance or release/conformance evidence.

## Scope and resulting workflow

The [readiness guide](../guides/cartulary_browser_design_readiness_workflow.md)
now covers current application navigation, grid/inspector behavior, recovery,
presentation settings, and all six current extension profiles. It distinguishes
live application observations, simulated presentation fixtures, canonical visual
comparisons, accessibility checks, and service/API evidence.

`make browser-design-review` starts a fresh isolated, seeded browser environment;
`REVIEW_PROFILE=default` retains the existing four claimed extensions and omits
Network Flow, while the default `network_flow_claimed` includes it. Both leave
Enterprise Authentication unclaimed. `browser-design-review-smoke` verifies the
sample workflows and closes. The authored task surface registers both as helpers
outside verification gates; generated task/topology projections follow it.

The setup uses the existing suite runtime, service broker, browser lifecycle,
v7 attachment validator and sealed frontend build. API seeding establishes two
incidents, zero/one/multiple-incident actors, viewer/editor/admin roles, Timeline
history, Evidence with a real object attachment and Timeline relationship,
coordination records and a shared saved view. The claimed profile imports the
existing Network Flow fixture through the application. Sample files use current
machine contracts and existing fixture assets, including the intentional v4
incident-bundle format. Credentials and authenticator keys remain private and
are removed at teardown. Diagnostic screenshots are not committed goldens.

Owner reconciliation changed only the stale single-incident design acceptance
row, the design registry reference to v6, and active-stack wording to v7. Core 01
explicit incident selection remains authoritative. The draft Reference Pack
NLSpec adds no new setup or product requirements.

## Observation context

Live setup observations use Chromium from the pinned Playwright installation,
Linux/WSL2, a 1440×900 viewport, 100% zoom, dark graphite, cleared density override,
and the application's resolved density. No text-spacing or motion override is
applied to the live seed. The finite smoke also opens 1280×720 default browser contexts
for the three non-admin roles. Canonical visual/accessibility rows use their own
existing declared presentation profiles. Each retained root records its source
and runtime identities; review roots additionally retain `review-session.json`,
samples, screenshots, startup/lease evidence, cleanup evidence and a secret scan.

## Findings and dispositions

| ID / classification | Reproduction, observation and owner | Evidence / disposition |
| --- | --- | --- |
| BR-01 — coverage gap | Inspect the claimed Snapshot/Reporting profile and current frontend route/workspace composition. Reporting and Report Composition have adopted API/backend capabilities; there is no current frontend composition workspace. Owners: adopted Reporting/Report Composition NLSpecs; `module.reporting`, `module.reportcomposition`. | Static source/owner inspection. **Defer** product workspace implementation; disclose the gap in the six-profile matrix. Backend success must not be reported as browser coverage. |
| BR-02 — coverage gap | Review enterprise sign-in presentation fixtures under `module.auth`; provider discovery/begin is simulated. Neither review profile configures a real OIDC/SAML provider. | Simulated presentation evidence only. **Defer** external-provider round-trip verification until an explicit provider configuration is supplied. Guide records optional deployment configuration; no IdP is bundled. |
| BR-03 — accepted intentional state | Log in as the zero-, one-, and multiple-incident review accounts and open `/`. All stay in the directory until an explicit Open action. Owner: Core 01 incident navigation. | Live browser/API smoke. **Accept** the owner behavior; repair stale design D-AC-076 restatement. |
| BR-04 — guide/setup issue | Start review sessions in both profiles, import sample inputs, then interrupt. During implementation, final scanning rejected screenshot/copied-sample permissions. Owner: Testing Harness runtime/evidence isolation. | Corrected all authored sample/screenshot writes to mode 0600; final validation below supersedes the intermediate failures. **Change before acceptance**, implemented. |
| BR-05 — fixture issue | Run lifecycle smoke against the existing fixture that supplied obsolete stage `functional` to the current attachment validator. Owner: `harness.browser`. | The fixture now supplies `webserver-backed`; obsolete v4 assertion labels were also corrected to v7. **Change before acceptance**, implemented without changing production attachment rules. |
| BR-06 — guide/setup issue | Run graph verification in a shell without a global `node`, despite the pinned repo runtime being installed. Row processes fail to spawn as `service_start_error`. Owner: harness graph execution/toolchain setup. | Initial owner-slice roots `20260923T215802Z-p35598` and `20260923T215803Z-p35823`; both pass with the pinned bin directory on PATH. **Defer** unrelated runner behavior; document the environment prerequisite. |
| BR-07 — fixture issue | Invoke standalone `make harness-command-surface-contract`. Its Vitest diagnostics subtest requires a private suite runtime that this standalone wrapper does not provide. Owner: harness command-surface/execution test setup. | Root `20260923T215912Z-p77780`, retained stdout. **Defer** unrelated wrapper repair; the supported command-surface owner slice provides a suite runtime and passes. |
| BR-08 — coverage gap | Run `make lint-scripts` after the scripts-to-tools relocation. Its tracked-file scope still selects only `scripts/**/*.mjs`, so it does not lint the new `tools/` helpers. Owner: harness static analysis. | `tools/harness/static-analysis/scripts-biome.sh`; **defer** repository-wide lint-scope repair. Do not interpret this target's success as lint coverage of the new helpers. Executed helper/lifecycle validation is separate evidence. |

## Validation

Verification commands are run from the repository root. Where necessary,
`PATH="$PWD/tmp/node-runtime/bin:$PATH"` supplies the pinned runtime to graph
children. Final run outcomes and remaining limitations are recorded below.

| Command / observation | Outcome and retained root under `.cartulary/test-results/` |
| --- | --- |
| `make browser-design-review-smoke` | Pass: `design-review-1790201070329-287fff37`; real workbook, Evidence bytes uploaded/downloaded, CSV import, reference-pack import, incident-bundle access/handoff, role controls and Network Analysis import. `review-smoke.json` passes; lease released/completed, secret scan passes, private runtime removed. |
| `make browser-design-review-smoke REVIEW_PROFILE=default` | Pass: `design-review-1790201206050-bf18717b`; same shared workflows with Network Analysis omitted. Closed state, released/completed lease, passing secret scan and removed private runtime. |
| `make browser-design-review REVIEW_PROFILE=default`, then Ctrl-C | Pass lifecycle observation: `design-review-1790201084312-bc98e1d2`; ready → closed, released/completed lease, passing secret scan and removed private runtime. Make itself can return an interruption status even when the review helper closes normally. |
| `make browser-design-review`, then Ctrl-C with asynchronous attachment monitoring | Pass: `design-review-1790201460233-7a68aeeb`; closed terminal, released/completed lease, passing secret scan and removed private runtime. |
| Interrupt `make browser-design-review` during prerequisite build | Expected interruption: `design-review-1790201207249-d5efd40b`; terminated build process group, stopped preparation, retained primary interruption, passing secret scan, no private runtime left. |
| Concurrent claimed/default review sessions and subsequent frontend builds | Distinct ports, database names and object buckets confirmed in their v7 stacks. The default session remained ready while later builds/smokes ran, using its original sealed artifact. Supporting `harness.browser` frontend-artifact lifetime row also passes. |
| `make browser-design-review REVIEW_PROFILE=invalid` | Expected `usage_error`, exit 2 before review preparation. |
| `make test-slice OWNER=harness.browser` | Pass, 38/38 graph units: `20260923T220013Z-p95877`. Includes affected frontend/browser-support rows and immutable artifact checks. |
| `make test-slice OWNER=harness.command_surface` | Pass, 1/1 graph unit: `20260923T220012Z-p95658`. Standalone limitation remains BR-07. |
| `make browser-e2e-visual` | Pass, 12/12 graph units: `20260923T220239Z-p92909`. Ordinary validation; no golden refresh or committed image changes. |
| `make browser-e2e-a11y` | Pass, 20/20 graph units: `20260923T220249Z-p15714`. Existing row-owned accessibility profiles and fixtures. |
| `make generate` | Pass, latest projection regeneration: `20260923T221058Z-p31870`. |
| `make generate-drift` | Pass, 4/4 graph units: `20260923T221201Z-p63624`. |
| `make generated-artifact-policy-check` | Pass, 3/3 graph units: `20260923T220456Z-p30561`. |
| `make agent-finalize` | Pass: `20260923T221142Z-p58000`; includes schema/catalog and tier validation. Initial finalize attempts correctly rejected stale helper-source digests; regenerating projections resolved them. |
| `make json-shape-check` | Pass, 3/3 graph units: `20260923T220933Z-p25482`. |
| `make run-harness-smoke-lifecycle` | Pass: `20260923T221113Z-p45757`; browser lifecycle and review preparation assertions pass. |
| `make lint-markdown` | Pass: `20260923T221320Z-p68293`. |
| `make lint-scripts` | Pass with the coverage limitation BR-08: `20260923T220649Z-p62315`. |
| `make lint-shell` | Pass, 4/4 graph units: `20260923T220649Z-p62322`. |

The lifecycle smoke reuses the existing v7 readiness/profile/stale-build and
immutable-publication tests and adds review profile validation, independent sample
construction, private sample permissions, TOTP calculation, ordered cleanup, and
primary-failure preservation across acquisition, seeding, hold and cleanup errors.
Health checks run asynchronously; interruption during a check or between checks
stops monitoring without being misclassified as a stale attachment.
Actual successful and failed session runs complement these setup tests; their
screenshots are live observations and do not manufacture product row results.

Retained-run maintenance, performance-evidence maintenance and retained-run drift
checks are explicitly skipped: `RESULTS_DIR` is unset and no full successful warm
`check` run was selected. Full release/conformance checks and optional real-IdP
verification were not requested as acceptance evidence for this helper. Product
readiness still requires human disposition of observations and the documented
coverage gaps. No production API or UI behavior changed.
