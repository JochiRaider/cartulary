# Workbook Find

Runtime-only search intent lives in `WorkbookFindController`. Sources provide a
current authorized snapshot, ordered semantic membership and committed readable
text fragments. The controller schedules bounded batches, publishes matching cell
identities atomically and fences movement against current scope and input. It
owns no rows, editor drafts, query requests, writes or persistent index.

`WorkbookFindControl` renders the scoped non-modal interaction from that owner's
projection. Typing changes results only. Explicit movement uses the injected
semantic navigation command and collapses the panel on acceptance; Close clears
state. The live status remains mounted to announce the result after collapse.

Timeline supplies the first source in `timeline/hooks/useTimelineFind.ts` and
`timeline/models/timelineFindText.ts`. The source intersects accepted query rows
with Adapter presentation membership. Collaboration supplies current read
authority independently of mutation replay admission. Collection summary and
scalar formatting remain Timeline-owned. Borrowed inspector/collection editors settle through
existing source save callbacks; Adapter scalar sessions use `navigateToCell`.

Another authorized grid can supply the same readable-text, membership and
navigation capabilities without another Find owner. Exhaustive server search
requires its own adopted boundary; do not scan additional pages here.

Verification is routed through `web.workbook`, `module.workbook` and
`package.grid_adapter`. Use the current Make task guides to select evidence.
Production focus and scrolling require browser scenarios, not unit fakes.
