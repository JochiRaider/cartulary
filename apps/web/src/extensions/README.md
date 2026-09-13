# extensions/

[Source overview / parent](../README.md)

Client extension discovery, availability coordination, and stable extension workspace identities.

Feature implementations consume this availability boundary. Discovery and
packaged client support jointly determine availability; feature presentation
does not redefine extension claims or workspace identity.

## Files

| File | Responsibility |
| --- | --- |
| [extensionAvailability.ts](extensionAvailability.ts) | Decodes packaged support and discovered claims; coordinates generation-bound, fail-closed workspace availability. |
| [ExtensionAvailabilityContext.tsx](ExtensionAvailabilityContext.tsx) | React provider and required-consumer hook for the extension availability controller. |
| [extensionWorkspaceIdentities.ts](extensionWorkspaceIdentities.ts) | Stable extension profile, route-family, workspace, and sheet-reference identities. |

## Tests

| File | Responsibility |
| --- | --- |
| [extensionAvailability.test.ts](extensionAvailability.test.ts) | Tests for support-registry decoding, exact availability intersections, stale-response rejection, generation rollover, and fail-closed behavior. |
