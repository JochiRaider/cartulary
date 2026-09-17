# Grid Adapter

The package owns semantic grid interaction and its production RDG binding.
Workbook query requests, cursors, retained work and server normalization stay with
Workbook and its source owners.

## Bounded workbook windows

`GridHandle.cancelEdit` clears the matching editor without moving focus from an
external recovery panel or another control. Keyboard cancellation inside the
grid retains its normal cell focus.

`GridHandle.detachEdit` separates editor presentation from the caller's retained
draft. `getActiveCell` exposes the semantic anchor used by Workbook when explicitly
browsing or leaving a sheet. Appending rows preserves captured selection;
eviction prunes selected identities and invalidates ranges whose original members
are no longer contiguous. Focus fallback preserves the eligible field and does
not steal focus from a newer interaction or a browsing control. Core-record
fallback does not replace extension-owned page focus. Operational-state actions
retain their return anchor across deferred acceptance and temporary document-body
focus caused by a disabled control.

Grouping descriptors may supply `compareValues` for owner-defined bucket order.
The adapter preserves incoming row order within each bucket and never turns that
comparison into a server query sort. The package's generic default remains
insertion order for callers that do not supply a comparator.

Select-all names its loaded-record scope. Workbook adds the loaded-window
accessible description and explicit continuation controls outside the grid.

## Verification

Column sizing uses `columnSizing.ts` for normalization and cancellable intrinsic
measurement of production DOM. `useGridColumnSizing.ts` binds viewport/font and
presentation lifetime; `ColumnResizeHandle.tsx` translates pointer deltas into
CSS pixels. Semantic bounds come from callers. The neutral port returns
measurements without storing widths or defining saved-view behavior. Editors and
owner-declared local-value presentations are excluded. `columnSizing.test.tsx`
covers command translation, bounds, unavailable geometry and cancellation;
production geometry requires browser evidence.

Compiled `GridColumn.width` is the only width input. The adapter does not seed a
second vendor width map: a grouped remount or saved configuration replacement
must render the owner's current widths. Viewport observation follows replacement
grid roots and clipping ancestors. Completed measurements remain fixed until
another explicit command; observer and font events only invalidate pending work
and update availability.

Use `make task-guide ROLE=module-author OWNER=package.grid_adapter` from the
repository root to select current semantic, production-binding, browser,
accessibility and visual evidence. Application fakes cannot establish RDG focus,
virtualization or scrolling behavior. See the frontend implementation testing
guide and visual golden maintenance guide under `docs/guides/`.

Nested editor popups declare `data-grid-editor-interaction` on their owned controls. Capture handlers defer keyboard navigation and dismissal to that boundary; ordinary cell typing and commit/navigation retain their existing semantics. Keep popups inside editor DOM ownership so internal focus movement does not blur-commit the cell.

## Clipboard representation boundary

[clipboardCodec.ts](src/clipboardCodec.ts) owns portable copy representations,
MIME precedence, inert HTML extraction and strict delimited decoding. It returns
scalar, rectangular table, no-op or explicit failure before semantic planning.
Cartulary-marked HTML carries version and data intent only; identifiers and field
rights never come from the clipboard. Formula-safe plain text is the fallback.
Single-line plain text is scalar; multiline/tab text is TSV. Explicit CSV requires
its MIME representation. Native editors keep their own clipboard behavior.

Workbook owns schema-header recognition, field contracts and mutation admission.
The grid publishes a destination range only after synchronous source acceptance;
asynchronous scalar completion does not move a newer selection. Native event
identity deduplicates delivery, while separate gestures remain distinct actions.
The authored JSON corpus under `contracts/tabularingest` is shared with Go tests.

## Contiguous cell interaction

`GridCellRange` remains the only completed selection. `gridInteractionController`
classifies stationary release, drag and Shift extension; `gridInteractionDom`
owns capture, current registered-cell geometry and bounded edge scrolling.
`useGridInteraction` binds their mounted lifetime. Tentative outlines are private
and never authorize clipboard or mutation operations. Keyboard and pointer
extension share `extendSemanticCellRange`; explicit semantic navigation replaces
selection, while vendor focus notifications only acknowledge the active cell.
Admitted destinations use the semantic focus owner even when the vendor already
considers the endpoint active. Entering a native recordless draft clears the
completed range without creating a record or taking ownership of its text.

An explicitly authorized consumer enables `cellRangeSelection` with kind
`contiguous` and an opaque accepted-surface `scopeKey`. Omission preserves existing
keyboard selection and ordinary editing without enabling pointer ranges. No
consumer supplies pointer IDs, coordinates, DOM nodes or a second selection store.
Scope/authority changes cancel intentions; captured ordered membership is retained
only when `retainGridCellRange` accepts it. Value/version refreshes and appends
outside the range can preserve selection. Group headers, collapsed rows, hidden
fields and drafts are excluded. Read-only members gain no mutation rights.

The existing editor session remains the commit gate. A rejected departure retains
the exact draft; superseding gestures abandon only destinations, never write
settlement. The old mouse-down/window mouse-up timer, vendor click activation and
unconditional vendor range collapse have been retired. Native editors, embedded
actions, column sizing/reordering and fill handles keep their separate ownership.

## Semantic Find seam

`GridHandle.presentation` publishes ordered visible semantic field keys and
presented row identities, including virtualized membership and excluding collapsed
records. It contains no cell values. `navigateToCell` departs through the active
scalar editor's deduplicated commit gate, then invokes the existing cancellable
reveal/focus owner. `requestFocus` remains a restoration command and forwards its
AbortSignal without adding an editor commit. New interactions cancel pending
destinations, never authoritative settlement. `ownsNavigationFocus` keeps callers
out of vendor DOM details when deciding focus-scoped application shortcuts.

Sources provide the optional `findMatch` semantic state. Its accessible marker
and current-match indication coexist with primary state, focus and selection.
