# services/

[Source overview / parent](../README.md)

Browser transport, generated-contract adapters, bounded observation, and shared operation clients.

Transport handles credentials, CSRF, cancellation, decoding, and public errors.
App and feature controllers own workflow decisions. Extension discovery and
availability live in [extensions](../extensions/README.md).

## Files

| File | Responsibility |
| --- | --- |
| [asyncObservation.ts](asyncObservation.ts) | Abortable bounded reads, operation observation, and delays over an injectable clock. |
| [browserApi.ts](browserApi.ts) | Owner-neutral browser JSON, multipart, and generated-operation transport with path/query derivation, response validation, CSRF, and public-error extraction. |
| [clientTransactionId.ts](clientTransactionId.ts) | Secure prefixed client transaction ID generation using Web Crypto UUIDs or RFC 4122 v4 fallback formatting. |
| [commonJobContract.ts](commonJobContract.ts) | Shared common-job validation, terminal-state classification, and non-regression checks. |
| [httpTransport.ts](httpTransport.ts) | Same-origin JSON and multipart transport mechanics for credentials, CSRF, cancellation, parsing, optional runtime decoding, and sanitized contract failures. |
| [importClient.ts](importClient.ts) | Typed workbook Import transport with captured writes, validated reads, receipts, and failures. |
| [importContractAdapter.ts](importContractAdapter.ts) | Thin generated-protocol alias facade for Import and common-job operations and workflow resource types. |
| [importJobContract.ts](importJobContract.ts) | Validates workbook import jobs and extracts correlated session references from receipts. |
| [importTargetContractAdapter.ts](importTargetContractAdapter.ts) | Thin generated-protocol facade for Import target discovery and target-specific mapping contracts. |
| [networkFlowContractAdapter.ts](networkFlowContractAdapter.ts) | Thin post-decode Network Flow presentation type and decoder facade; contains no handwritten wire model. |
| [networkFlowIndicatorAdapter.ts](networkFlowIndicatorAdapter.ts) | Adapts Core Indicator target discovery and atomic-IP compatibility for Network Flow linking. |
| [publicErrorIdentity.ts](publicErrorIdentity.ts) | Validates structured public-error reason codes, including Evidence access reasons. |
| [referencePacks.ts](referencePacks.ts) | Reference Pack catalog, lifecycle, import, and job transport with captured command outcomes. |
| [workbookEvidence.ts](workbookEvidence.ts) | Evidence upload/attach client helpers and evidence public-error mapping. |

## Tests

| File | Responsibility |
| --- | --- |
| [browserApi.test.ts](browserApi.test.ts) | Tests for browser API base/path helpers. |
| [clientTransactionId.test.ts](clientTransactionId.test.ts) | Tests for platform UUID use, secure-random fallback formatting, prefixes, and unavailable-crypto failure. |
| [clientTransactionIdPolicy.test.ts](clientTransactionIdPolicy.test.ts) | Static policy test excluding counters, clocks, and insecure randomness from browser mutation IDs. |
| [importClient.test.ts](importClient.test.ts) | Tests exact workbook upload replay, acceptance/resource separation, and selection receipts. |
| [importTargetContractAdapter.test.ts](importTargetContractAdapter.test.ts) | Tests for strict Import target contract decoding and generated-protocol alignment. |
| [networkFlowContractAdapter.test.ts](networkFlowContractAdapter.test.ts) | Network Flow claimed-profile and compiled-contract-major admission tests. |
| [referencePacks.test.ts](referencePacks.test.ts) | Tests Reference Pack inline/asynchronous outcomes and exact selected-scope replay. |
| [workbookEvidence.test.ts](workbookEvidence.test.ts) | Tests for evidence client helpers and error mapping. |
