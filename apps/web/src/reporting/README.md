# reporting/

[Source overview / parent](../README.md)

Reporting owns captured snapshot and internal-draft render commands, exact common
Job observation, cancellation, and presentation. The application adapter owns
session/profile authority and lifetime. Drawer detachment pauses local observation;
server work continues. Returning to the drawer requires explicit refresh or exact
request recovery. Replacing the account or incident retires protected state.

| File | Responsibility |
| --- | --- |
| [ReportingController.ts](ReportingController.ts) | Immutable request admission, bounded reads, monotonic Job receipts, exact cancellation and snapshot binding. |
| [ReportingController.node.test.ts](ReportingController.node.test.ts) | Uncertain replay, identity rejection, authority retirement, cancellation and bounded observation. |
| [ReportingPanel.tsx](ReportingPanel.tsx) | Nonmodal snapshot/internal-draft controls and retained Job status. |

Internal drafts are not approvals or publication. This surface does not implement
report composition authoring, external release approval, or artifact downloading.

Public transport is owned by [reportingClient](../services/reportingClient.ts).
