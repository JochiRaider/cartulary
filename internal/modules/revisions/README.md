# Revisions History projection

Core 01 REQ-01-049/052A owns the public History representation. Core 02
REQ-02-218/265 owns source contributions and canonical retained snapshots.
These documents are human owner references; runtime and verification inputs
remain authored contracts and source code.

The existing application-composed provider catalog requires a pure
`HistoryProjector` for every record and non-row target contribution. Source
owners select public fields and interpret their retained facts. Revisions
loads canonical snapshots and collection mutations, validates their admission,
combines row detail with collection events when necessary, preserves logical
item identities, decorates imported attribution, and applies current rollback
eligibility. Generic History never switches on source record types.

`historycontract` supplies closed public value types and pure projection
mechanics. Absent, explicit null, false, zero, empty text, and string collections
remain distinct. Required canonical members must exist. Unsupported versions,
missing projectors, malformed values, and unsafe public shapes fail the read;
there is no successful missing-detail result. Source projectors allow public
fields explicitly. Storage snapshots, mutation identifiers, object content,
credentials, upload tokens, and access handles must never be serialized.

## Deployment and rollback

History requests use a typed limit and ordering position at the application
boundary. HTTP validates the protected cursor after record/incident admission and
before invoking History. The repository selects at most `limit + 1` logical event
descriptors; the extra descriptor determines continuation and is never projected.
Only selected mutation/revision snapshots are loaded. Complete association metadata
determines whether the first mutation in a change set also contains supplemental
row detail, including when a row mutation lies outside the page. Revision-only
events remain addressable. Ambiguous selected mutation/revision associations fail
the read instead of emitting duplicate logical identities.

Current reversal eligibility can require all mutations of a selected change set
and their source dependencies. This cost is distinct from projecting unrelated
retained history. The bounded paging regression records database row counts,
projector calls, allocation observations and query plans as history grows. Existing
indexes remain sufficient unless measured plans justify an additive index.

This change belongs to the pending OpenAPI 2.0.0 release boundary. The historical
1.0.0 baseline is immutable. Deploy the server and its bundled browser from the
same release; retain that coherent previous bundle for rollback. `/api/v1`,
mutation bodies, idempotency, History item identities, and opaque reversal
selectors retain their existing contracts. There is no legacy response adapter.

`history.position.v1` versions the ordering position inside the unchanged
`pagination.cursor.v1` protected envelope. Old History anchor cursors restart via
`invalid_pagination_request`; there is no legacy decoder. Rolling back the coherent
application bundle also restarts newer disposable cursors without clearing local
drafts, captured requests or receipts. Paging changes alone do not alter semantic
representation generation, retained content or selectors.

`data.representation_generation` identifies the semantic projection, including
empty pages. Change its value whenever a released projection changes committed
item content. Browsing and action lookup compare generations before comparing
items. A continuation from another generation requires a fresh first-page read.
That restart affects disposable reads only: it must preserve drafts, captured
request bytes, unresolved operations, and acknowledged receipts. Never require
a page reload as recovery from this condition.

Supported retained data consists of canonical snapshots and conformant imports.
This change requires no SQL migration or history rewrite. Schema-less legacy
databases remain subject to the existing operator-controlled reset boundary;
no application path infers historical shapes, backfills history, or resets a
database automatically. Imported source attribution comes from the portability
owner; obtaining an account-directory name is not a History dependency.

Use the current public task guide for `module.revisions` and affected source
owners. Semantic source fixtures cover all seventeen current record
configurations and seven non-row targets; catalog tests require the ten record
snapshot types and fourteen mutation target kinds. Browser validation includes
generation transitions, paging, current action eligibility, and retained
operation recovery. The controlling inspector handoff records exact execution
evidence and release gates.
