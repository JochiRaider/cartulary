# Workbook pending replay

[Runtime owner](../README.md)

`workbookPendingQueue.ts` owns the memory-local, incident/client-scoped autosave
queue: exactly 64 units, FIFO admission, contiguous coalescing, immutable uncertain
replay, authorization gates, overflow refusal, and conflict settlement.

The runtime retains one model independently of mounted surfaces. Source drivers
own domain payloads and transport; the queue does not persist drafts or interpret
presentation order as dispatch order. `workbookPendingQueue.test.ts` retains its
Workbook and Collaboration verification rows and semantic test titles.
