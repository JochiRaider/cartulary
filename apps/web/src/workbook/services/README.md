# workbook/services/

[Parent](../README.md) · [Source overview](../../README.md)

WorkbookReferenceSelection owns only picker browsing and staged identities:
one 100-row page, ten earlier request checkpoints, and at most 64 pending
collection actions. Parent grid and inspector drafts own accepted authoring.

workbookReferenceReader shares only in-flight reads within one authority instance,
preserves typed failures, and fences disposal using the existing 30-second
read deadline. Field target meaning comes from the view-contracts projection.
Membership discovery retains its separate route and ordering.

workbookReferenceSelection.test.ts verifies paging bounds, independent selections,
source replacement, continuation recovery, cancellation and authority concealment.
