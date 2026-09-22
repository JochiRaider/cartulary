# Workbook composition

The application supplies the `WorkbookMutationRuntimeRegistry`. The shell borrows
its active runtime. `WorkbookMutationRuntimeBoundary` acquires it in a layout
effect, so abandoned renders cannot retire or reconfigure a live lifetime.

`createWorkbookMutationInfrastructure` creates the retained runtime, all Timeline
owners and their dispatch/read capabilities once, before publication. The common
runtime stays independent of Timeline implementation. Its accepted authority
controls protected reads independently of a source owner's local denial.

`attachWorkbookMutationPresentation` binds mounted reconciliation and authority
notifications as one typed lease. Replacing the attachment invalidates its old
callbacks immediately; old cleanup cannot remove the replacement. Detaching
presentation leaves admitted requests, receipts, drafts and transport intact.
Late reads retain recovery debt and cannot restore obsolete focus. The application
registry owns account/incident replacement and terminal disposal.

The attachment preserves the source-specific reconciliation functions and shared
surface refresh-debt owner. It introduces no persistence, cross-tab coordination,
plugin registry or public workflow API.

`WorkbookMutationPresentation.test.tsx` covers abandoned rendering, Strict Mode
mounting, replacement cleanup, late refresh fencing and detached acknowledgement.

[Workbook](../README.md) · [Runtime](../runtime/README.md)
