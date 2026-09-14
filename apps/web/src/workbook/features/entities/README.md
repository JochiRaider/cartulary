# workbook/features/entities/

[Parent](../README.md) · [Source overview](../../../README.md)

Hosts and Identities inspector composition, Entity merge ownership, and clipboard paste.

Merge review and operation lifetime remain separate from inspector rendering.
The merge owner coordinates survivor and loser writes and retains exact recovery
attempts. Paste execution consumes the shared private
[clipboard adapter](../../adapters/README.md).

## Files

| File | Responsibility |
| --- | --- |
| [entityMergeOperation.ts](entityMergeOperation.ts) | Captured Entity merge attempts, receipts, transport outcomes, and owner bindings. |
| [entityMergeReview.ts](entityMergeReview.ts) | Entity merge authority and reviewed survivor/loser identities and versions. |
| [EntityWorkbookInspector.tsx](EntityWorkbookInspector.tsx) | Entity inspector facade over feature-owned composition and presentation. |
| [useEntityClipboardPasteController.ts](useEntityClipboardPasteController.ts) | Preserves scalar editing and admits Entity-origin table plans to retained batch ownership, coordinated with existing merge/write reservations. |
| [useEntityMergeController.ts](useEntityMergeController.ts) | Entity merge selection, eligibility, review, and confirmation controller. |
| [useEntityWorkbookInspectorComposition.tsx](useEntityWorkbookInspectorComposition.tsx) | Assembles Entity inspector sections, merge actions, and owner-local workflows. |
| [WorkbookEntityMergeOwner.ts](WorkbookEntityMergeOwner.ts) | Entity merge participant reservation, retained operation, acknowledgement, and recovery ownership. |
| [WorkbookEntityMergeRecovery.tsx](WorkbookEntityMergeRecovery.tsx) | Retained Entity merge recovery with replay and reconciliation actions. |

## Tests

| File | Responsibility |
| --- | --- |
| [entityInspectorEditing.test.tsx](entityInspectorEditing.test.tsx) | Selected-record editing, dirty refresh, explicit draft return and detached completion evidence. |
| [useEntityMergeController.test.tsx](useEntityMergeController.test.tsx) | Tests single merge confirmation and synchronous invalidation when reviewed inputs change. |
| [WorkbookEntityMergeOwner.test.ts](WorkbookEntityMergeOwner.test.ts) | Tests pair reservation, reviewed versions, coordinated writes, and exact uncertain merge replay. |
| [WorkbookEntityMergeRecovery.test.tsx](WorkbookEntityMergeRecovery.test.tsx) | Tests keyboard recovery, receipt completion, refresh-only retry, and protected-state concealment. |

## Ordinary creation seam

| File | Responsibility |
| --- | --- |
| [entityOrdinaryCreate.ts](entityOrdinaryCreate.ts) | Entity-origin ordinary preparation and direct-seed minima; exact reuse remains server owned. |
