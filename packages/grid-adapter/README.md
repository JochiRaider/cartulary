# Grid Adapter

The package owns semantic grid interaction and its production RDG binding.
Workbook query requests, cursors, retained work and server normalization stay with
Workbook and its source owners.

## Bounded workbook windows

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

Use `make task-guide ROLE=module-author OWNER=package.grid_adapter` from the
repository root to select current semantic, production-binding, browser,
accessibility and visual evidence. Application fakes cannot establish RDG focus,
virtualization or scrolling behavior. See the frontend implementation testing
guide and visual golden maintenance guide under `docs/guides/`.

Nested editor popups declare `data-grid-editor-interaction` on their owned controls. Capture handlers defer keyboard navigation and dismissal to that boundary; ordinary cell typing and commit/navigation retain their existing semantics. Keep popups inside editor DOM ownership so internal focus movement does not blur-commit the cell.
