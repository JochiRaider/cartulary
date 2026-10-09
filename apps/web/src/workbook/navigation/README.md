# workbook/navigation/

[Workbook / parent](../README.md)

Navigation admits semantic destinations and coordinates their committed
presentation. It does not own query pages, source authoring, resource operations,
or authorization. [Core 03 §2.5](../../../../../docs/spec/03_workbook_interaction_collaboration_and_workflows.md#25-investigation-workbench-navigation-and-commands)
owns navigation, pins and Return (REQ-03-311–318);
[Core 01](../../../../../docs/spec/01_architecture_storage_and_view_contracts.md)
owns bounded locator reads (REQ-01-680–684), and
[Core 04 §2.2](../../../../../docs/spec/04_security_deployment_and_conformance.md)
owns account/incident concealment and retirement (REQ-04-169–170).

| File | Responsibility |
| --- | --- |
| [WorkbookNavigationHost.ts](WorkbookNavigationHost.ts) | Explicit owner capabilities consumed by navigation and adapted by shell composition. |
| [WorkbookNavigationIntent.ts](WorkbookNavigationIntent.ts) | Supported destination, pin and Return variants and semantic intent equality. |
| [WorkbookNavigationPresentation.ts](WorkbookNavigationPresentation.ts) | Acceptance, committed selection, Grid Adapter focus and explicit inspector sequencing. |
| [WorkbookWorkbenchContext.tsx](WorkbookWorkbenchContext.tsx) | Semantic navigation actions and optional workbench presentation context. |
| [useWorkbookWorkbench.ts](useWorkbookWorkbench.ts) | Staged query/locator admission, saved-view observation, Return and inspector handoff. |
| [WorkbookSessionNavigation.ts](WorkbookSessionNavigation.ts) | Memory-local pins, bounded Return origins and cancellable attempt identity. |
| [WorkbookWorkPanel.tsx](WorkbookWorkPanel.tsx) | Work, Return and stable footer navigation feedback. |

Admission and presentation completion are distinct. The coordinator observes an
opaque query-owner acceptance handle; source selection commits before Grid
Adapter focus; inspector focus requires the current committed attachment. It
stores semantic identities and handles, never query rows or checkpoints.
Selection and mounted-presentation tokens fence stale acknowledgements. Replacing
only a grid handle cancels its pending focus without undoing committed selection.
New deliberate interaction and
authority loss fence obsolete intent. Navigation never submits or discards a
source-owner draft and does not duplicate source rows or receipts.

An explicitly requested inspector completes through its own committed focus
binding. Its presence can make the grid inaccessible; that is not a failed
inspector handoff. Grid-only destinations still require grid presentation
eligibility. Query acceptance, mounted presentation, grid eligibility and
inspector readiness are independent signals to the coordinator. Shell input
capture calls one navigation cancellation capability without consuming the event.

The shell supplies only the host capabilities in `WorkbookNavigationHost`.
Extend this boundary for a real consumed decision, not the shell's aggregate
return type. Saved-view resource operations remain in [savedviews](../savedviews/README.md),
query pages in [query](../query/README.md), and runtime lifetime in
[runtime](../runtime/README.md).

Note association links and saved-view browser selections require the narrow
`WorkbookNavigationActions.open(target, inspect?)` capability from composition.
An absent optional presentation context never selects a page-scanning or direct
resource-activation algorithm. Unsupported locators preserve the origin and
report a read failure without scanning. Same-ID saved-view selection reads the
current authorized resource before application.

Saved views issues an observation handle with accepted, rejected and aborted
outcomes. Each attempt releases only its own handle; selected-resource retention
is established before that release. Rejected results preserve the owner problem
classification: operational/contract failures offer read retry, unavailability
uses only authorized fallback, and authority failures use owner recovery.
Aborted observations create no failure notice.

Session pins and Return origins contain semantic identity, compatible query/layout,
saved-view comparison version and invoker metadata. They hold no rows, drafts,
receipts or authorization cache. Uncertainty conceals and cancels pending intent;
same-account recovery can reveal retained state. Account replacement and incident
departure/access loss retire it. Incident closure alone does not revoke reads.

Equivalent active semantic intents coalesce through admission and incomplete
presentation. Equality ignores object member order and labels, while retaining
meaningful array order, anchors, entry mode, Inspect and captured Return data.
After terminal completion a new activation starts a fresh observation. Session
phase is the pending-feedback source; trail push/pop happens once at admission.

Tests: `WorkbookNavigationPresentation.test.ts` covers commit/focus sequencing,
replacement, cancellation and delayed attachment;
`useWorkbookWorkbench.test.tsx` covers staged admission and presentation;
`WorkbookSessionNavigation.test.ts` covers pins, trail and cancellation;
`WorkbookNavigationStatus.test.tsx` covers feedback and keyboard access. Use
`make task-guide ROLE=module-author OWNER=web.workbook` for current verification;
browser evidence is routed through `module.workbook`.
