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

This change belongs to the pending OpenAPI 2.0.0 release boundary. The historical
1.0.0 baseline is immutable. Deploy the server and its bundled browser from the
same release; retain that coherent previous bundle for rollback. `/api/v1`,
mutation bodies, idempotency, History item identities, and opaque reversal
selectors retain their existing contracts. There is no legacy response adapter.

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
