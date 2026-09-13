# workbook/services/

[Parent](../README.md) · [Source overview](../../README.md)

Instance-scoped reference-query sharing, cancellation, invalidation, and disposal.

The reference broker shares mechanics and in-flight reads within one workbook
context. Surface policies choose reference requirements; the broker does not
choose domain references or authorize access.

## Files

| File | Responsibility |
| --- | --- |
| [referenceQueryBroker.ts](referenceQueryBroker.ts) | Creates instance-scoped, incident/authorization-bound reference-query ports with shared-consumer deduplication, typed invalidation, abort ownership, and idempotent disposal. |

## Tests

| File | Responsibility |
| --- | --- |
| [referenceQueryBroker.test.ts](referenceQueryBroker.test.ts) | Tests in-flight deduplication, shared-consumer cancellation, two-shell isolation, context binding, invalidation, teardown, and late-result rejection. |
