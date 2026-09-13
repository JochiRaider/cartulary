# collaboration/

[Source overview / parent](../README.md)

Incident-scoped browser WebSocket lifetime and typed collaboration event publication.

This layer owns transport lifetime and typed publication. The
[workbook coordinator](../workbook/collaboration/README.md),
[Timeline bindings](../workbook/timeline/collaboration/README.md), and
[Network Flow interpreters](../networkFlow/README.md) own feature effects.

## Files

| File | Responsibility |
| --- | --- |
| [IncidentCollaborationSession.tsx](IncidentCollaborationSession.tsx) | Incident-scoped WebSocket provider for hello/resume, private resume state, one sequence high-water mark, reconnect, heartbeat, safe decoding, presence publication, reset, revocation, and closure events. |
| [incidentCollaborationSessionPlan.ts](incidentCollaborationSessionPlan.ts) | Pure decoded-message plan for handshake, heartbeat, replay de-duplication, sequence gaps, reset suppression, and terminal session outcomes. |

## Tests

| File | Responsibility |
| --- | --- |
| [IncidentCollaborationSession.test.tsx](IncidentCollaborationSession.test.tsx) | Tests for single-socket lifetime, surface presence changes, replay deduplication/gaps, heartbeat, and unknown-message safety. |
| [incidentCollaborationSessionPlan.test.ts](incidentCollaborationSessionPlan.test.ts) | Tests handshake and terminal-message classification, sequence deduplication, gaps, and reset suppression. |
