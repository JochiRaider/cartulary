# Timeline shell capture expansion

Status: complete; two fresh ordinary visual comparisons pass against the same final manifest.

## Scope and outcome

Trigger: the three primary shells already used the rich investigation, but other
Timeline captures still owned independent sparse seed records. The user requested
all populated Timeline captures, including backgrounds behind menus, inspectors,
authoring and recovery panels. This follow-up starts from the completed
[fixture remediation](timeline-fixture-capture-remediation.md) and its
[initial refresh](timeline-fixture-capture-refresh.md), on the same uncommitted
`main` baseline `6456013e8`.

Dataset `timeline-investigation-rich-v1`, content revision 1 and recipe schema v1
remain unchanged. The browser adapter now accepts an explicit continuation count:
0 gives the common twelve-event core; 36 retains the primary shell's 48 records.
The interactive-review adapter remains at 66 records. The core semantic digest is
`ea764b9f54835155ab881ca90863bfa8c239c4cbe0345a70bbb6bfce57e8b169`.
All ten default-visible source fields have meaningful values. Column order,
widths, viewport sizes and product defaults remain unchanged; narrower captures
naturally leave later columns outside the frame.

The executable profile inventory contains 146 exact capture bindings: 123 rich,
13 intentional empty/state, and 10 sparse-inspector captures. Rich scenarios use
the complete core, plus explicitly keyed fully populated mutable targets where
required. Queries still filter and group the records. Separate incidents keep
Timeline enrichment out of Evidence-only datasets. Existing helper consumers keep
their original setup when supplied context is absent.

## Implementation and requirement mapping

- `docs/design.md` §15.2 owns the full-dataset rule and permitted exceptions.
  Testing Harness TH-HARNESS-REQ-818 owns admission, supplied contexts, profile rejection,
  pre-fault source verification and final core preservation.
- `apps/web/e2e/support/visual/timelineVisualFixture.ts` owns incident admission,
  semantic target lookup, scenario creation, mounted-source checks and final
  authoritative core verification. Capture-time checks make no API requests
  while intentional transport failures are installed.
- `apps/web/e2e/support/timeline/timelineScenarioFields.ts` composes complete
  explicit scenario source values from the shared fixture and rejects derived,
  unknown, null or empty values in rich records. Missing-value tests retain their
  separate fixtures.
- `tools/frontend_visual_timeline_profiles.json` and its v1 schema declare exact
  capture identities and exception purposes. JSON-shape checks join each entry
  to the capture policy. Undeclared Timeline captures fail before comparison.
- `apps/web/e2e/workbook.visual.spec.ts` migrates every populated Timeline capture
  family. Shared workbook, mention and incident-panel helpers accept supplied
  contexts without changing their functional-test defaults.
- `apps/web/e2e/visual-harness.spec.ts` verifies profile admission, semantic mapping,
  source preservation and complete core read-back. The recipe unit slice covers
  complete scenario composition and rejection of invalid source/profile inputs.
- Registry v6 keeps its wire format and identities. Review found incomplete
  catalog row lists in `visual.fixture.edit_cell` and
  `visual.fixture.save_state_strip`; the missing Timeline, grid-adapter and
  collaboration catalog rows are now explicit.

The follow-up source inventory is:

- Browser entry points: `workbook.visual.spec.ts`, `visual-harness.spec.ts`.
- Timeline support: `timelineInvestigation.ts`, `timelineInvestigation.test.ts`,
  `timelineScenarioFields.ts`; visual support: `timelineVisualFixture.ts` and
  `timelineDataProfiles.ts` under `apps/web/e2e/support/`.
- Supplied-context helper consumers under `apps/web/e2e/support/`:
  `entities/mentions.ts`, `incidentMembershipAudit.ts`,
  `incidentMembershipManagement.ts`, and workbook `query.ts`,
  `timelineCaptureActions.ts`, `timelineRelatedEvidence.ts`, `noteCreate.ts`,
  `coordinationCreate.ts`, `contextualCreate.ts`, `indicatorObservations.ts`.
  `evidence/fixtures.ts` adds the optional authored Received value.
  `workbook/decisionSupersession.ts` adds optional authored Decided values.
- Machine inputs: `tools/frontend_visual_timeline_profiles.json`, its v1 schema,
  `tools/harness_schema_attachments.json`,
  `tools/harness/generated-artifacts/check-json-shapes.mjs`,
  `tools/test_families/harness.browser.json`, and the visual fixture registry.
- Generated projections: browser batch manifest and topology render index;
  promoted artifacts: the 112-image refresh inventory and golden manifest.
- Human owners and evidence: design §15, Testing Harness TH-HARNESS-REQ-818, this handoff,
  the follow-up refresh record and links from the preceding remediation records.

No production API, schema, database, saved-view or renderer migration is needed.
No additional golden or adopted design identity is introduced. The earlier
24-field semantic coverage matrix remains in the fixture remediation handoff.

## Capture review findings

The initial ordinary candidate executed all 255 capture identities and produced
146 data-profile receipts, with no setup, workflow or core-preservation failure.
It differed in 111 images. Every candidate was inspected at useful resolution
through digest-verified Make UI-review imports.

The two active-edit images were rejected: an existing left-scroll reset hid the
focused Synopsis editor. Those captures now preserve the editor's naturally
established horizontal position and assert full viewport visibility, focus and
stable scroll around capture. The canonical shell retains Date Entered focus at
top-left. This fixes capture composition without changing keyboard behavior.

The unchanged baseline also exposed an intermittent Evidence inspector difference
(641 pixels). Attachment finalized a file with wall-clock `received_at`, which is
source data. The visual evidence helper now explicitly supplies and reads back
`2026-05-01T10:45:00Z`; ordinary helper callers remain unchanged. No broad source
normalization or comparison-tolerance change was used.

The first post-promotion pair passed once, then exposed a separate Decision review
difference (6,769 pixels). Its writable Decided values were also omitted and
therefore clock-generated. Variable timestamp length changed the inspector's
review-section scroll framing. Visual setup now supplies and reads back fixed
target/replacement values, `2026-04-18T14:44:08Z` and
`2026-04-18T15:42:09Z`. Existing functional helper setup is unchanged when these
values are absent. The focused Decision scenario passes against the already
promoted images; this correction required no further promotion.

The next pair exposed an independent sparse-inspector preparation defect. Both
runs differed only in History, rollback preview and public-error captures; all
rich Timeline captures passed. Their scroll anchors had been established before
metadata normalization, so metadata wrapping above them changed the final offset.
History and public error now use the existing asserted start anchor after
normalization and match their accepted images. Rollback uses an explicit centered
confirmation anchor with Cancel focused. Its framing correction is separately
reviewed; the sparse source fields, actions and geometry assertions retain their
purpose. No pixel offset or comparison tolerance substitutes for the anchor.

The remaining reviewed differences show richer source cells and additional rows
behind the existing foreground UI. Filtered saved views intentionally show one
matching scenario record or the three reviewed records. Constrained panels retain
their scroll positions, focus and actions; core relationships, dismissed mentions,
review and supersession remain produced by real owner workflows. Image review
establishes visual observations, not accessibility or release conformance.

## Verification ledger

All run IDs below are under `.cartulary/test-results/`. Graph results include owned
setup/cleanup units; counts are not individual assertion totals.

| Command / scope | Result | Run |
| --- | --- | --- |
| `make browser-e2e-visual`, unchanged working-tree baseline | Failed 10/12 units; only Evidence affordance image differed, 641 pixels | `20261001T163900Z-p54178` |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.timeline_investigation_recipe` | Passed 2/2; six unit cases | `20261001T165522Z-p17046` |
| `make generate` | Passed | `20261001T165339Z-p2959` |
| `make frontend-import-boundary-check` | Passed 2/2 | `20261001T165449Z-p15687` |
| `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.timeline_investigation_capture` | Passed 11/11 | `20261001T165620Z-p22211` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.coordination_create_matrix,module.workbook.browser.note_create_sources,module.workbook.browser.indicator_observations_paging_source_edits` | Passed 13/13 | `20261001T165620Z-p22222` |
| `make service-backed-test-slice OWNER=module.incidents ROWS=module.incidents.browser.membership_audit_response_inspection,module.incidents.browser.membership_management_mutation_recovery` | Passed 12/12 | `20261001T170016Z-p53580` |
| `make browser-e2e-visual`, expanded candidate | Failed image comparison only, 111 differences; 255 captures reconciled | `20261001T170003Z-p32794` |
| `make format` | Passed 2/2 | `20261001T172412Z-p48698` |
| `make frontend-typecheck` | Passed 2/2 | `20261001T172508Z-p56608` |
| `make json-shape-check` | Passed 3/3 | `20261001T172510Z-p56754` |
| `make generated-artifact-policy-check` | Passed 3/3 | `20261001T171806Z-p35820` |
| `make lint-biome` | Passed 2/2 | `20261001T172536Z-p58655` |
| `make agent-finalize` | Passed; retained-run maintenance skipped because `RESULTS_DIR` was unset | `20261001T172524Z-p57526` |
| `make lint-markdown`, before final handoff edits | Passed | `20261001T170016Z-p53869` |
| `make browser-e2e-visual`, first post-promotion comparison | Passed 12/12 | `20261001T173422Z-p1461` |
| `make browser-e2e-visual`, second post-promotion comparison | Failed 10/12; only Decision review image differed, 6,769 pixels; fixed authored timestamps afterward | `20261001T173424Z-p1655` |
| `make format`, after Decision fixture correction | Passed 2/2 | `20261001T174523Z-p81047` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.visual.decision_supersession_review_recovery` | Passed 11/11 against the promoted manifest | `20261001T174542Z-p85823` |
| `make frontend-typecheck`, after Decision fixture correction | Passed 2/2 | `20261001T174543Z-p86085` |
| `make lint-biome`, after Decision fixture correction | Passed 2/2 | `20261001T174852Z-p20418` |
| `make agent-finalize`, before final pair | Passed; retained-run maintenance skipped because `RESULTS_DIR` was unset | `20261001T174851Z-p20166` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.decision_supersession_review_recovery` | Passed 11/11; existing helper defaults, without metadata normalization | `20261001T175055Z-p94826` |
| `make browser-e2e-visual`, next ordinary pair | Both failed 10/12; same three sparse inspector images differed; all rich captures passed | `20261001T174920Z-p24845`, `20261001T174921Z-p25044` |
| Sparse inspector visual owner slice, first anchor correction | Image comparison only: History and public error passed, rollback still inherited a variable offset | `20261001T175957Z-p35592` |
| Sparse inspector visual owner slice, explicit centered rollback anchor | Preparation and all other images passed; only reviewed rollback framing differs | `20261001T180203Z-p73394` |
| `make format`, after final inspector anchor correction | Passed 2/2 | `20261001T180140Z-p68568` |
| `make frontend-typecheck`, after inspector anchor correction | Passed 2/2 | `20261001T180204Z-p73802` |
| `make json-shape-check`, after inspector registry clarification | Passed 3/3 | `20261001T180205Z-p73993` |
| `make lint-biome`, after inspector anchor correction | Passed 2/2 | `20261001T180327Z-p8444` |
| `make agent-finalize`, before anchor promotion | Passed; retained-run maintenance skipped because `RESULTS_DIR` was unset | `20261001T180326Z-p8190` |
| `make browser-e2e-visual-update`, final anchor promotion | Passed 12/12; only the reviewed rollback PNG changed from the first promotion | `20261001T180405Z-p12672` |
| `make browser-e2e-visual`, final ordinary run 1 | Passed 12/12; 47 scenarios, 255 captures, no retries/flaky/skipped scenarios | `20261001T181153Z-p49823` |
| `make browser-e2e-visual`, final ordinary run 2 | Passed 12/12 against the same manifest; 47 scenarios, 255 captures, no retries/flaky/skipped scenarios | `20261001T181154Z-p50017` |
| `make lint-markdown`, final handoff and owner documentation | Passed | `20261001T182038Z-p22234` |

An earlier expanded attempt (`20261001T165620Z-p22399`) failed Playwright discovery
because the new hook did not destructure its fixture argument. The hook syntax
was corrected before the complete candidate run. Initial type/lint failures were
corrected and superseded by the passing results above.

## Preserved exceptions

These exact captures preserve their original data and purpose. Foreground panels
named “empty” or “loading” are not exceptions when their Timeline background is
populated.

| Capture | Identity | Profile |
| --- | --- | --- |
| `design-delayed-initial-loading` | `visual.capture.3f34e7b4236421b3e863` | empty |
| `design-grid-unavailable-initial-load` | `visual.capture.c8259c0780b175f21e6b` | empty |
| `design-immediate-initial-loading` | `visual.capture.a8043c7522df282f104c` | empty |
| `timeline-mutation-empty-timeline-query` | `visual.capture.a19f6bdc0eb0997fcda9` | empty |
| `workbook-inspector-attached-edit` | `visual.capture.bec04aee967bd2e5d893` | sparse |
| `workbook-inspector-compact-actions` | `visual.capture.26d4c9bb1434fa99f29b` | sparse |
| `workbook-inspector-destructive-confirmation` | `visual.capture.06903c2d94c98d2d2eaf` | sparse |
| `workbook-inspector-details` | `visual.capture.1124de6a22d3e44e0f5f` | sparse |
| `workbook-inspector-history` | `visual.capture.21a59ffec6a398738353` | sparse |
| `workbook-inspector-narrow-technical-details` | `visual.capture.f5ca1ba5368f9538cbb9` | sparse |
| `workbook-inspector-public-error` | `visual.capture.1aebc4c7b4cfc2965a5f` | sparse |
| `workbook-inspector-relationships` | `visual.capture.e0929813825f6088fd2b` | sparse |
| `workbook-inspector-retained-draft` | `visual.capture.e31aaf9c5514ea71e39f` | sparse |
| `workbook-inspector-rollback-preview` | `visual.capture.3451d6d508d3782f0df3` | sparse |
| `workbook-query-empty-closed-read-only` | `visual.capture.23f73af839fc55869642` | empty |
| `workbook-query-empty-compact` | `visual.capture.e5c167baf7ec64225366` | empty |
| `workbook-query-empty-density-comfortable` | `visual.capture.30d95ea87bb7ba5570a4` | empty |
| `workbook-query-empty-density-compact` | `visual.capture.69f7321c2878fabf6c61` | empty |
| `workbook-query-empty-narrow` | `visual.capture.8600c4d30dc3b5cefd5b` | empty |
| `workbook-query-empty-successful-query` | `visual.capture.4f6f9418cf30dff3e93d` | empty |
| `workbook-query-empty-text-spacing` | `visual.capture.83841467fb122ec391d1` | empty |
| `workbook-query-empty-zoom-200` | `visual.capture.1906a5e6fafb74a2acc9` | empty |
| `workbook-query-filtered-empty` | `visual.capture.d7d60d718df7c7a56852` | empty |

## Renderer and final evidence

Renderer remains `visual.renderer.playwright_1_63_0_chromium_1243_linux_amd64`,
Chromium 153.0.8010.12, en-US, scale 1, with existing presentation profiles.
Container digest: `bc6ab0d6d44ff4826e4cb8c1e6d801e185bfc42bb0753f8e2a30efc70db054c7`.
Font digest: `c21f8663e6c8fe72681b2be644aa8398538afc59a0f0cda06b94d46d5fbba5fe`.
Starting manifest: `ce1508c7630322fd903941cf70c6d1abdc0487ff390fff947193ec9925e226bd`.

The first promotion passed 12/12 units at `20261001T172554Z-p62130`, refreshing
111 PNGs. Its manifest was
`8c5b07c83844d1a95098c7066e1701af559fc5dcd2d3bfb0fa73d6b87cb8690f`.
The final anchor promotion passed 12/12 at `20261001T180405Z-p12672` and changed
only the reviewed rollback preview, bringing the follow-up total to 112 PNGs.
Final manifest: `785598423f199c91a965c89ae8aa648f1b3b0653aefc2983cefbdb02f8d7a2fc`.
All 23 exception specimens retain their source inputs and purpose; 22 images
remain byte-identical, and the rollback preview has the documented anchor correction.
Exact capture identities, catalog
and fixture mappings, and every image disposition are in the
[follow-up refresh record](timeline-shell-capture-refresh.md).
Promotion retained all twelve current/capture/review layout-study images and six
rich-inspector Details/Relationships/Evidence review images. It produced all 146
data-profile receipts and reconciled 255 captures to 255 goldens and 29 registered
fixtures, with no missing, orphaned, ambiguous or unresolved mapping.
Two fresh ordinary visual runs, `20261001T181153Z-p49823` and
`20261001T181154Z-p50017`, passed against that exact manifest. Each retains all
146 profile receipts, twelve layout studies and six rich-inspector images, with
255 exact capture/golden joins and no reconciliation error. No applicable
implementation or image-review requirement remains unresolved.

## Cleanup, limits and rollback

Artifact review used bounded sessions and exact selected run/capture identities.
A session storage-capacity rejection was resolved by closing the consumed session
and starting a fresh one; no limit was bypassed. A canonical import rejected the
incomplete edit-fixture catalog mapping. The exact image was inspected as a
standalone reference, with no DOM or accessibility claim; the registry is repaired
before promotion. All completed review sessions returned cleanup `complete` and
foreground exit 0. Transient credentials and private mappings are not in this record.

Completed artifact-review sessions were `20261001T165634Z-p73951`,
`20261001T171052Z-p13233`, `20261001T171657Z-p25748`,
`20261001T171840Z-p37939`, `20261001T172342Z-p46852`,
`20261001T173448Z-p61653`, `20261001T174311Z-p79891`,
`20261001T175740Z-p29761`, and `20261001T180141Z-p68965`.
Each retained `ui-review/terminal.json` reports
cleanup complete; foreground processes exited successfully. Session operation
failures remain recorded separately and are not converted to successful imports.
One inspector import initially rejected a caller request with non-private file
permissions. Correcting that request to mode 0600 admitted the same exact source;
no artifact safety check was bypassed.
The private caller scratch was removed after the final evidence checks.

The unchanged review adapter and previous 24-field semantic workflow coverage are
supported by the preceding remediation ledger. This follow-up does not rerun a
full release, accessibility or Core 05 publication claim. Restore the coherent
source, projections, profiles, golden images and manifest together to roll back;
seeded incidents are disposable, so no application-data migration is required.
