# workbook/inspector/presentation/

[Parent](../README.md) · [Source overview](../../../README.md)

Shared inspector shells, panels, actions, feedback, and History event rendering.

These components render declared subjects, panels, action bindings, and safe
feedback. They receive feature commands and presentation models rather than
owning source mutations or request lifetimes.

## Files

| File | Responsibility |
| --- | --- |
| [WorkbookHistoryPresentation.tsx](WorkbookHistoryPresentation.tsx) | Shared History list and event rendering from presentation models. |
| [WorkbookInspectorActions.tsx](WorkbookInspectorActions.tsx) | Inspector action groups, contextual actions, and accessible action buttons. |
| [WorkbookInspectorFeedback.tsx](WorkbookInspectorFeedback.tsx) | Inspector metadata, technical details, safe errors, feedback, and confirmation presentation. |
| [workbookInspectorPresentationModel.ts](workbookInspectorPresentationModel.ts) | Inspector action bindings, disabled reasons, technical fields, and History event presentation types. |
| [WorkbookInspectorShell.tsx](WorkbookInspectorShell.tsx) | Shared inspector shell and declared panel-section layout. |

## Tests

| File | Responsibility |
| --- | --- |
| [WorkbookInspectorPresentation.test.tsx](WorkbookInspectorPresentation.test.tsx) | Tests valid subject boundaries, ordered panels, deleted-record History, and explicit creation content. |

The shell keeps its compact title, Close control, measured direct navigation or Sections disclosure, and optional
owner-contributed unfinished-work entry outside the single
`data-inspector-scroll-body`. Full record context and technical metadata belong
in that body. Semantic focus targets scroll within it; outer slot geometry,
separator behavior and focus restoration retain their layout owners.

`WorkbookInspectorDeclaredPanelList` admits sections once. Its descriptor sequence
supplies both the shell navigation and mounted body contributions. The layout's
bounded navigation selection survives closing the inspector for the same subject;
it carries no protected content, permissions, or authoring. Explicit focus wins
over passive section tracking. History marks its existing unrequested Open control
as the section entry; choosing the section focuses that control without invoking it.

Retained source rows use the incident, account and access epoch as their read
scope. A saved-view presentation reset can clear selection without retiring a
still-authorized canonical row. Actual access loss conceals and retires the old
observation; recovery requires a newly authorized source observation. Presentation
reset keys never substitute for that access boundary.

`WorkbookInspectorPanelContent.tsx` renders explicit data/access states without
owning requests or authoring. Contextual action groups share descriptions by
reason identity and parameters, never by wording. Feature owners retain closed
cause vocabularies; shared rendering knows only their presentation contributions.
Buttons use authored component tokens. Form controls share
`../../components/workbookFormStyles.ts`; grid-specific sizing remains separate.
That boundary also applies complete semantic typography roles. Narrative preview
limits come from the generated design-presentation facade, not local literals.

Saved Details preserve contract field order. Scalar read kinds use aligned rows;
multiline bodies and reason notes use full-width narratives. The closed design
presentation v2 projection overrides exactly eight Timeline metadata fields to
compact rows and keeps its two activity fields narrative. Unknown and collection read kinds stack safely.
Values distinguish unloaded, null, empty text, false and zero. The six-line
preview measures actual text layout and exposes an explicit expansion control.

Ordinary editing supplies content, actions, feedback and retained-draft slots.
Update and Close editor share an action row. The draft owner supplies original
field cues and review/discard commands; the component does not copy draft state.
An owner-admitted Resume attaches unchanged work in one activation. Changed
dependencies enter explicit review; original-field and authority checks stay
with the edit owner.
Timeline and Entity owners supply collection-management focus destinations.

Relationship summaries retain source mention identity and distinguish raw text,
resolution state and authorized target labels. Selected correction controls stay
immediately after their original item, including session-observed dismissed
items. Candidate read/filter/paging/retry controls remain inside the chooser. Collapsing them retains mounted picker/creation state;
source-owned operation receipts and recovery commands remain outside the
disclosure. Empty bounded target pages never imply a complete search.

Evidence uses separate accepted-information, file-access and attachment regions.
Preview/download eligibility does not derive from a linked-record count. Source
owners provide access results and file lifecycle feedback independently. A shared
attachment entry delegates picker, drop and paste to the same admitted command;
its Attach file button supplies keyboard access and a paste focus target.
Progress and recovery remain visible after the entry, with separate source-link
and saved-but-unrefreshed meanings supplied by the operation owner.

## History reading and action placement

`workbookHistoryEventPresentation` maps the closed public semantic units to
reading content. Display summaries and labels never decide action eligibility;
only source operation metadata and the existing permission owner do that.
Imported source attribution takes precedence over the local execution identity;
an already-authorized supplied name may replace its identifier label. No
account-directory query is introduced.

Each event shows attribution and absolute UTC time with a numeric offset.
`Event details` retains every typed before/after value and event-local reversal
choice. Absent, null, empty text, scalars and collections stay distinct; exact
field keys and references live in subordinate technical disclosure. Native disclosure preserves mounted
content. Nested Escape closes the nearest disclosure; confirmation handles its
own Escape first. Closing event detail cancels its unsubmitted review, including a pending read,
and clears disposable checking state. Captured submitted work stays with its owner.
Record deletion/restoration appears after events under `Record actions`.
Confirmation retains the safe initial focus and source-owned command/receipt
lifetime. Whole-change-set review explicitly includes other affected records.

The former raw-operation secondary paragraph, technical-only actor presentation,
global reversal confirmation and leading destructive-action block have no
remaining consumers. Shared History list/event presentation also serves batch
review; field/grid and operation recovery remain with their existing owners.

Contextual commands render once in canonical feature order as action rows. Their
result descriptions come from the authored presentation projection; semantic
capability and mutation intent select authoring or review, never label text.
Panel `featureContent` binds active authoring and feedback to the original
feature identity immediately after its command. Shared reasons retain
identity/parameter deduplication; repeated outcome descriptions are shared once
within the group. No-subject creation remains a source-owned region.
The contextual dispatcher is not a second route registry: source-local read,
field, relationship, lifecycle and History contributions retain their bindings.
Unsupported features retain Core 01’s `omit_feature` behavior; this presentation
slice does not invent query predicates or write commands for an unbound feature.

Saved-subject titles always take precedence over operation headings. Creation
without a saved subject declares creation mode; operation titles stay inside
Workflow or the source-local form. Assessment uses the shared read-only Details
facade because its source is append-only. Its explicit append, close and discard
commands remain on the Assessment owner.

Specialized Note, coordination, contextual Task/Decision, related-Evidence,
Assessment and Indicator forms consume shared field, group, heading, action and
text roles. The retired Observation/lifecycle style modules have no remaining
consumers. These are presentation primitives: each source still owns field
validation, draft identity, close/discard, source replacement and accepted receipts.


Attention contributions carry owner-issued work, subject/authority and outcome
identities. The shell admits, orders, deduplicates and navigates; it neither
classifies operations nor dispatches recovery. Navigation uses one admitted
section sequence for direct and chooser presentations and never opens lazy
History. Measurement is clipped independently from the visible scrollport.
