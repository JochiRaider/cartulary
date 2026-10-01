# Workbook pending replay

[Runtime owner](../README.md)

`workbookPendingQueue.ts` owns the memory-local, incident/client-scoped autosave
queue: exactly 64 units, FIFO admission, contiguous coalescing, immutable uncertain
replay, authorization gates, overflow refusal, and conflict settlement.

The runtime retains one model independently of mounted surfaces. Source drivers
own domain payloads and transport; the queue does not persist drafts or interpret
presentation order as dispatch order. `workbookPendingQueue.test.ts` retains its
Workbook and Collaboration verification rows and semantic test titles.

Queue lifecycle is private: open, closed, reopened with retained work, or retired.
Authentication remains an independent replay gate. Closure retains pending units;
reopening cannot replay them. Settlement or explicit discard removes those units,
and only an empty reopened queue permits fresh work through ordinary authority
checks. Retirement is permanent and cannot be reversed by late callbacks.

Closure recovery follows the current queued head after settlement, including a
late retryable failure or transfer to the independent same-field conflict queue.
It cannot refer to removed work or skip an earlier returned unit. Specific terminal
failures retain their own recovery semantics; the lifecycle is not a second
authorization, conflict, or save-state owner.
