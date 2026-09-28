# Six `make check` failures: remediation handoff

## Baseline, authority, and scope

Work began on `main` at `4daa23ed50ac96d77fd3ed3a8a5d90352020e56c` with the ordinary reference-picker implementation already staged. Those staged changes were preserved. The earlier full run `.cartulary/test-results/20260928T180220Z-p91269` passed 964/970 units; `.cartulary/test-results/20260928T200034Z-p28950` reproduced the Timeline and WorkbookShell failures. The six failures were handled as separate ownership seams, without changing picker semantics.

Core 01 §16 and Core 03 §8.4 govern opaque same-origin Evidence handles and local blocked feedback. The Collaboration boundary decision owns intent persistence, while Entities owns candidate reads. The Testing Harness NLSpec owns direct `check` graph accounting for a promoted Fallow unit. The OpenTelemetry NLSpec §4.5 owns emitted-shape classification. Existing product owners required no amendment. The harness owner was updated to state the direct Fallow rule; authored verification inputs were regenerated. Tests, runtime, contracts, and release evidence remain independent of this Markdown handoff.

## Corrections and owner decisions

| Failure | Correction and ownership |
| --- | --- |
| Evidence raw handle and missing feedback | The existing `resolvePublicEvidenceHandleHref` rule is applied at the effect boundary as well as transport. A rejected handle leaves no iframe or download effect and shows local “Preview unavailable” feedback. Row-initiated issuance tolerates selecting that same row; selection of a different record, replacement, closure, or access loss retires the ticket. The Evidence hook and capability port retain request and feedback state. |
| OpenTelemetry advisory | The coordinated Go train moved to API/SDK/exporters `v1.45.0` and logs `v0.21.0`; the authored source snapshot records resolved versions. Log attribute calls use the new `attribute` API. Cartulary constructs explicit HTTP signal paths, explicit gRPC TLS credentials, and retains `TraceIDRatioBased`. Existing conformance and golden corpus passed unchanged, so the §4.5 classification is `registry_equivalent`; no NLSpec migration note or golden refresh was needed. |
| Fallow ownership | The unused preview-probe symbol and its now-unused exported type are private. The global Fallow scan is one direct `check` target, with frontend readiness; the package grid-adapter proxy row and its owner verification reference were retired together. The Testing Harness NLSpec now states that direct accounting. |
| Selector policy | Authoring picker tests use the Records listbox role and name. Column-sizing browser tests use `dataTestIdSelector` from `@cartulary/ui-contracts` before browser evaluation. No policy exception or product selector changed. |
| Entities/Collaboration boundary | Entities retains candidate-read and `change_sets` assertions. A Collaboration-owned integration test calls the public candidate endpoint and checks its private intent table for non-publication. No production port or cross-owner test utility was added. |
| Timeline status isolation | `WorkbookExplicitPatchOwner.configure` publishes a genuine capability change. Timeline collection cells observe `canSubmit` and pending block state for their own record, Timeline inspector details select their row's patch state, and explicit-patch recovery observes only matching entries. Notes status publications now leave unrelated collection cells and grid identities stable, while relevant drafts and conflicts still render. |

No route, wire, schema, database, storage, or data conversion is involved. Roll back each bounded source/test/routing change with its generated projection; keep the previously staged picker slice as its separate change.

## UI digest assessment

| Rows | Assessment |
| --- | --- |
| A001–A003 | PASS. Adopted owners, staged baseline, source boundaries, authored routing, and independent test evidence were checked. |
| A011, A014, A016, A018–A019 | PASS for affected journeys. Evidence feedback, stale target retirement, Timeline editor continuity, candidate selector semantics, and picker browser focus remain covered by focused rows. |
| A023–A027 | PASS subject to final integrated verification. Authored selectors and test-family inputs route checks independently of documentation; generation and drift passed. |
| A004–A010, A012–A013, A015, A017, A020–A022 | N/A to this remediation. Theme, density, creation, responsive chrome, mutation replay, conflict UI, and golden publication were not changed. |

## Verification

Commands ran at the repository root through public Make targets. Focused successful evidence includes:

| Command | Result and run root |
| --- | --- |
| `make generate` | PASS `.cartulary/test-results/20260928T203309Z-p44720` |
| `make generate-drift` | PASS `.cartulary/test-results/20260928T203428Z-p52903` |
| `make harness-contract` | PASS `.cartulary/test-results/20260928T203641Z-p65137` |
| `make otel-conformance` | PASS `.cartulary/test-results/20260928T203322Z-p47577` |
| `make go-vulncheck` | PASS `.cartulary/test-results/20260928T203714Z-p94025`; no blocking `GO-2026-6508` |
| `make frontend-fallow-static` | PASS `.cartulary/test-results/20260928T202403Z-p25352` |
| `make service-backed-test-slice OWNER=module.collaboration ROWS=module.collaboration.integration.entity_candidate_discovery_does_not_publish` | PASS `.cartulary/test-results/20260928T203641Z-p64851` |
| `make test-slice OWNER=module.records ROWS=module.records.architecture.boundary` | PASS `.cartulary/test-results/20260928T203641Z-p64811` |
| `make backend-module-boundary-check` | PASS `.cartulary/test-results/20260928T202211Z-p21386` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.column_sizing_gestures_saved,module.workbook.browser.column_sizing_surfaces,module.workbook.browser.column_sizing_viewport,module.workbook.browser.reference_selection_recovery` | PASS `.cartulary/test-results/20260928T203707Z-p86079` |
| `make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.selectorcontractpolicy_keeps_cross_boundary_sele_61868c7039` | PASS `.cartulary/test-results/20260928T203428Z-p52997` |
| `make frontend-import-boundary-check` | PASS `.cartulary/test-results/20260928T203714Z-p93452` |
| `make lint-biome` | PASS `.cartulary/test-results/20260928T203714Z-p93554` |
| `make json-shape-check` | PASS `.cartulary/test-results/20260928T203713Z-p93181` |
| `make generated-artifact-policy-check` | PASS `.cartulary/test-results/20260928T203714Z-p93253` |
| `make frontend-typecheck` | PASS `.cartulary/test-results/20260928T203849Z-p31229` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.authoring_candidate_discovery,web.workbook.regression.evidence_access_request_lifetime_ea01,web.workbook.regression.explicit_task_patch_recovery,web.workbook.regression.reference_controls,web.workbook.regression.reference_selection,web.workbook.regression.timeline_autosave_subscription_isolation,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e` | PASS `.cartulary/test-results/20260928T203907Z-p31900` |
| `make format` | PASS `.cartulary/test-results/20260928T203838Z-p26723` |
| `make agent-finalize` | PASS `.cartulary/test-results/20260928T203348Z-p48647`; `RESULTS_DIR` unset |
| `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.collection_inspection_preserves_draft_text_and_s_f5c8c4c211` | PASS `.cartulary/test-results/20260928T205234Z-p74066` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_linked_evidence_preview_resolution_975344b0d0,web.workbook.regression.timeline_autosave_subscription_isolation` | PASS `.cartulary/test-results/20260928T205234Z-p74081` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.evidence_access_request_lifetime_ea01` | PASS `.cartulary/test-results/20260928T205454Z-p17131`; alternate-port Preview and Download reject external, wrong-route, and encoded malformed handles without navigation |
| `make frontend-typecheck` | PASS `.cartulary/test-results/20260928T205252Z-p75525` after final Timeline selector change |
| `make lint-biome` | PASS `.cartulary/test-results/20260928T205252Z-p75559` |
| `make generate-drift` | PASS `.cartulary/test-results/20260928T205252Z-p75336` |
| `make format` | PASS `.cartulary/test-results/20260928T205608Z-p40451` |
| `make check` | PASS 970/970 `.cartulary/test-results/20260928T205346Z-p84344`; expanded alternate-port test was edited during the run and passed separately in the focused row above |
| `make check` | PASS 970/970 `.cartulary/test-results/20260928T210530Z-p83887` from the stable final source and test state; direct Fallow and security units passed |

The first post-change web slice `.cartulary/test-results/20260928T203428Z-p52944` exposed an Evidence retarget/shared-feedback regression; the ticket identity was corrected and the focused Evidence row passed at `.cartulary/test-results/20260928T203601Z-p59494`. Two TypeScript runs exposed shared-hook type constraints and informed the optional context shape. An intermediate `make generate` failed at `.cartulary/test-results/20260928T203240Z-p41555` because the new title was out of ASCII order; authored ordering was corrected and generation passed. An initial Fallow run `.cartulary/test-results/20260928T202211Z-p21423` found the exported probe-result type after the function was made private; that type is now private and Fallow passes.

The first integrated `make check` at `.cartulary/test-results/20260928T204221Z-p40077` passed 968/970 units. Its two failures were `module.timeline.frontend.collection_inspection_preserves_draft_text_and_s_f5c8c4c211` and `web.workbook.regression.timeline_linked_evidence_preview_resolution_975344b0d0`. The first showed that tag removal still needs its own record's pending block state; the scoped selector now includes that fact. The second used `/opaque/unused`, which the adopted handle route correctly rejects; the synthetic fixture now uses `/api/v1/evidence-handles/unused`. Their focused reruns pass above. The direct `target:frontend-fallow-static` unit passed in this integrated run.

`make lint-markdown` passed for the handoff draft at `.cartulary/test-results/20260928T204037Z-p37882`, its updated version at `.cartulary/test-results/20260928T210358Z-p77939`, and the integrated-result version at `.cartulary/test-results/20260928T211547Z-p70285`. `make agent-finalize` passed at `.cartulary/test-results/20260928T210505Z-p79923` with `RESULTS_DIR` unset before the final integrated run. No eligible successful full warm check existed at that time, so retained-run maintenance was skipped. No visual golden was refreshed. The prior seeded picker review remains in the picker handoff; no new visual styling was introduced by these fixes.
