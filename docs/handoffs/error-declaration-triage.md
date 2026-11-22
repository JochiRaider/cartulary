# Error declaration audit triage

The S-1150 report recorded 224 operation/code pairs across 91 operations. All 136 catalog operations had a call-graph root. The status below comes from the error registry. An observed response would require independent evidence of that status and code on the named operation.

The evidence root is `tcr-s1150-errdecl02`, artifact `error-declaration-audit.json`. Producer names below are relative to the Go internal directory. The owning team is the operation owner from the catalog. This review has no runtime or CI consumers.

**Disposition.** Seven operations now use the mutation decoder directly. Source tracing excluded the remaining 217 pairs as public writes on their reported operations. The qualified check has its own proof requirement. No operation declaration was widened.

## module.auth

**Simplified 1.** The user mutation decoder now returns invalid_mutation_payload directly. Its callers previously replaced the discarded auth decoder error.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createDeploymentUser` | 400 `invalid_auth_request` | `modules/auth.invalidAuthRequest` |
| `patchCurrentAccountProfile` | 400 `invalid_auth_request` | `modules/auth.invalidAuthRequest` |
| `patchDeploymentUser` | 400 `invalid_auth_request` | `modules/auth.invalidAuthRequest` |
| `putCurrentAccountPreferences` | 400 `invalid_auth_request` | `modules/auth.invalidAuthRequest` |
| `resetDeploymentUserPassword` | 400 `invalid_auth_request` | `modules/auth.invalidAuthRequest` |
| `resetDeploymentUserTOTP` | 400 `invalid_auth_request` | `modules/auth.invalidAuthRequest` |
| `revokeAllDeploymentUserSessions` | 400 `invalid_auth_request` | `modules/auth.invalidAuthRequest` |

**Excluded 2.** Create looks up the provider by key. Rotate and retire look up the addressed binding by user and binding ID. The sibling not-found sentinel does not arise in this transaction.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createEnterpriseAuthBinding` | 404 `auth_binding_not_found` | `(*modules/auth.Service).writeEnterpriseBindingResult` |
| `retireEnterpriseAuthBinding` | 404 `auth_provider_not_found` | `(*modules/auth.Service).writeEnterpriseBindingResult` |
| `rotateEnterpriseAuthBinding` | 404 `auth_provider_not_found` | `(*modules/auth.Service).writeEnterpriseBindingResult` |

**Excluded 3.** CompleteSAMLEnterpriseAuthTransaction validates completion and browser binding. The state/nonce sentinels consumed by enterpriseCompletionError come from the OIDC path.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `finishEnterpriseSAML` | 409 `provider_response_rejected` | `modules/auth.providerResponseRejected` |

## module.entities

**Excluded 1.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `listEntityCandidates` | 403 `authorization_denied` | `modules/entities.authorizationDeniedError` |

## module.evidence

**Excluded 1.** concealedRole intercepts DenialNotVisible. Its exact callback selects evidenceRecordNotFound for attach/issue, blobNotVisible for verify, handleNotFoundOrRevoked for redeem, and objectUploadNotFoundOrRevoked for upload.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `attachBlobToEvidenceRecord` | 404 `handle_not_found_or_revoked` | `modules/evidence.handleNotFoundOrRevoked` |
| `attachBlobToEvidenceRecord` | 404 `incident_not_found` | `modules/evidence.evidenceAdmissionResult` |
| `attachBlobToEvidenceRecord` | 404 `object_upload_not_found_or_revoked` | `modules/evidence.objectUploadNotFoundOrRevoked` |
| `issueEvidenceDownloadHandle` | 409 `evidence_attach_rejected` | `modules/evidence.evidenceAttachRejected` |
| `issueEvidenceDownloadHandle` | 404 `handle_not_found_or_revoked` | `modules/evidence.handleNotFoundOrRevoked` |
| `issueEvidenceDownloadHandle` | 404 `incident_not_found` | `modules/evidence.evidenceAdmissionResult` |
| `issueEvidenceDownloadHandle` | 404 `object_upload_not_found_or_revoked` | `modules/evidence.objectUploadNotFoundOrRevoked` |
| `issueEvidencePreviewHandle` | 409 `evidence_attach_rejected` | `modules/evidence.evidenceAttachRejected` |
| `issueEvidencePreviewHandle` | 404 `handle_not_found_or_revoked` | `modules/evidence.handleNotFoundOrRevoked` |
| `issueEvidencePreviewHandle` | 404 `incident_not_found` | `modules/evidence.evidenceAdmissionResult` |
| `issueEvidencePreviewHandle` | 404 `object_upload_not_found_or_revoked` | `modules/evidence.objectUploadNotFoundOrRevoked` |
| `redeemEvidenceHandle` | 409 `evidence_attach_rejected` | `modules/evidence.evidenceAttachRejected` |
| `redeemEvidenceHandle` | 404 `evidence_record_not_found` | `modules/evidence.evidenceRecordNotFound` |
| `redeemEvidenceHandle` | 404 `incident_not_found` | `modules/evidence.evidenceAdmissionResult` |
| `redeemEvidenceHandle` | 404 `object_upload_not_found_or_revoked` | `modules/evidence.objectUploadNotFoundOrRevoked` |
| `uploadObjectBlobContent` | 409 `evidence_attach_rejected` | `modules/evidence.evidenceAttachRejected` |
| `uploadObjectBlobContent` | 404 `evidence_record_not_found` | `modules/evidence.evidenceRecordNotFound` |
| `uploadObjectBlobContent` | 404 `handle_not_found_or_revoked` | `modules/evidence.handleNotFoundOrRevoked` |
| `uploadObjectBlobContent` | 404 `incident_not_found` | `modules/evidence.evidenceAdmissionResult` |
| `verifyObjectBlob` | 404 `evidence_record_not_found` | `modules/evidence.evidenceRecordNotFound` |
| `verifyObjectBlob` | 404 `handle_not_found_or_revoked` | `modules/evidence.handleNotFoundOrRevoked` |
| `verifyObjectBlob` | 404 `incident_not_found` | `modules/evidence.evidenceAdmissionResult` |
| `verifyObjectBlob` | 404 `object_upload_not_found_or_revoked` | `modules/evidence.objectUploadNotFoundOrRevoked` |

**Excluded 2.** The key syntax check passes a local PersistedObjectBlobStorageKey validation error to the dependency classifier. It performs no object-store request at that site.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createObjectBlobSlot` | 503 `object_store_access_rejected` | `modules/evidence.objectStoreDependencyAPIError` |
| `createObjectBlobSlot` | 503 `object_store_unavailable` | `modules/evidence.objectStoreUnavailableAPIError` |

**Excluded 3.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `issueEvidenceDownloadHandle` | 403 `authorization_denied` | `modules/evidence.evidenceAdmissionResult` |
| `issueEvidencePreviewHandle` | 403 `authorization_denied` | `modules/evidence.evidenceAdmissionResult` |
| `redeemEvidenceHandle` | 403 `authorization_denied` | `modules/evidence.evidenceAdmissionResult` |

## module.imports

**Excluded 1.** startApply loads a session selection. requireSelectedUnitsReadyTx maps a missing selected unit to import_apply_blocked with unit_not_ready.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `applyImportSession` | 404 `import_unit_not_found` | `modules/imports.writeImportStoreError` |

**Excluded 2.** admitImportSession intercepts DenialNotVisible and returns import_session_not_found before calling importAdmissionResult.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `applyImportSession` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `createImportUnitRegion` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `getImportSession` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `getImportUnit` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `getImportUnitPreview` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `inspectImportUnitExtensionSource` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `listImportUnits` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `previewImportUnitExtensionMapping` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `putImportUnitMapping` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `selectImportUnit` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |
| `skipImportUnit` | 404 `incident_not_found` | `modules/imports.importAdmissionResult` |

**Excluded 3.** The named admission decoder belongs to a different operation. Follow the selected job-kind contribution to the route-specific request reader.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createImportSession` | 400 `invalid_release_request` | `modules/reporting.DecodeCreateReleaseRequest`, `modules/reporting.invalidReleaseRequest`, `modules/reporting.materializeOutputOptions` |
| `createImportSession` | 400 `invalid_request` | `modules/reportcomposition.DecodePreviewRequest`, `modules/reportcomposition.schemaFieldError`, `modules/reportcomposition.validateGraphRefs` |

**Excluded 4.** The named producer is reporting.internalAPIError. The actual handler translates its own synchronous admission/store result. Even composition preview maps errors from its reporting port through reportcomposition.writeStoreError.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createImportSession` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |

**Excluded 5.** createOperatorRegion checks session state, source bounds and collisions. It neither selects units nor invokes the apply readiness checks.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createImportUnitRegion` | 409 `import_apply_blocked` | `modules/imports.writeImportStoreError` |

**Excluded 6.** regionSourceAPIError admits import_source_rejected and translates other source failures to internal_error.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createImportUnitRegion` | 409 `import_source_unsupported` | `modules/imports.importSourceUnsupported` |

**Excluded 7.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `getImportSession` | 403 `authorization_denied` | `modules/imports.importAdmissionResult` |
| `getImportUnit` | 403 `authorization_denied` | `modules/imports.importAdmissionResult` |
| `getImportUnitPreview` | 403 `authorization_denied` | `modules/imports.importAdmissionResult` |
| `listImportUnits` | 403 `authorization_denied` | `modules/imports.importAdmissionResult` |

**Excluded 8.** These editing actions use withUnitMutation. The lifecycle guard in startApply is outside that path.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `putImportUnitMapping` | 409 `incident_closed` | `modules/imports.writeImportStoreError` |
| `selectImportUnit` | 409 `incident_closed` | `modules/imports.writeImportStoreError` |
| `skipImportUnit` | 409 `incident_closed` | `modules/imports.writeImportStoreError` |

**Excluded 9.** skipUnit changes selection and refreshes session status. The overlap/readiness checks that construct applyBlockedError run on selection/apply.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `skipImportUnit` | 409 `import_apply_blocked` | `modules/imports.writeImportStoreError` |

## module.incidentbundles

**Excluded 1.** The named admission decoder belongs to a different operation. Follow the selected job-kind contribution to the route-specific request reader.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `importIncidentBundle` | 400 `invalid_import_request` | `modules/imports.invalidImportRequest` |
| `importIncidentBundle` | 400 `invalid_release_request` | `modules/reporting.DecodeCreateReleaseRequest`, `modules/reporting.invalidReleaseRequest`, `modules/reporting.materializeOutputOptions` |
| `importIncidentBundle` | 400 `invalid_request` | `modules/reportcomposition.DecodePreviewRequest`, `modules/reportcomposition.schemaFieldError`, `modules/reportcomposition.validateGraphRefs` |

**Excluded 2.** The named producer is reporting.internalAPIError. The actual handler translates its own synchronous admission/store result. Even composition preview maps errors from its reporting port through reportcomposition.writeStoreError.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `importIncidentBundle` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |

## module.incidents

**Excluded 1.** The operation supplies its own failure constructor to mapAdmissionError. The named sibling decoder belongs to another mutation.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `closeIncident` | 400 `invalid_incident_create` | `modules/incidents/httpapi.invalidIncidentCreate` |
| `closeIncident` | 400 `invalid_incident_patch` | `modules/incidents/httpapi.invalidIncidentPatch` |
| `closeIncident` | 400 `invalid_mutation_payload` | `modules/incidents/httpapi.invalidMutationPayload` |
| `createIncident` | 400 `invalid_incident_lifecycle_request` | `modules/incidents/httpapi.invalidIncidentLifecycleRequest` |
| `createIncident` | 400 `invalid_incident_patch` | `modules/incidents/httpapi.invalidIncidentPatch` |
| `createIncident` | 400 `invalid_mutation_payload` | `modules/incidents/httpapi.invalidMutationPayload` |
| `createIncidentMembership` | 400 `invalid_incident_create` | `modules/incidents/httpapi.invalidIncidentCreate` |
| `createIncidentMembership` | 400 `invalid_incident_lifecycle_request` | `modules/incidents/httpapi.invalidIncidentLifecycleRequest` |
| `createIncidentMembership` | 400 `invalid_incident_patch` | `modules/incidents/httpapi.invalidIncidentPatch` |
| `deleteIncidentMembership` | 400 `invalid_incident_create` | `modules/incidents/httpapi.invalidIncidentCreate` |
| `deleteIncidentMembership` | 400 `invalid_incident_lifecycle_request` | `modules/incidents/httpapi.invalidIncidentLifecycleRequest` |
| `deleteIncidentMembership` | 400 `invalid_incident_patch` | `modules/incidents/httpapi.invalidIncidentPatch` |
| `patchIncident` | 400 `invalid_incident_create` | `modules/incidents/httpapi.invalidIncidentCreate` |
| `patchIncident` | 400 `invalid_incident_lifecycle_request` | `modules/incidents/httpapi.invalidIncidentLifecycleRequest` |
| `patchIncident` | 400 `invalid_mutation_payload` | `modules/incidents/httpapi.invalidMutationPayload` |
| `patchIncidentMembership` | 400 `invalid_incident_create` | `modules/incidents/httpapi.invalidIncidentCreate` |
| `patchIncidentMembership` | 400 `invalid_incident_lifecycle_request` | `modules/incidents/httpapi.invalidIncidentLifecycleRequest` |
| `patchIncidentMembership` | 400 `invalid_incident_patch` | `modules/incidents/httpapi.invalidIncidentPatch` |
| `reopenIncident` | 400 `invalid_incident_create` | `modules/incidents/httpapi.invalidIncidentCreate` |
| `reopenIncident` | 400 `invalid_incident_patch` | `modules/incidents/httpapi.invalidIncidentPatch` |
| `reopenIncident` | 400 `invalid_mutation_payload` | `modules/incidents/httpapi.invalidMutationPayload` |

**Excluded 2.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `listIncidentMemberships` | 403 `authorization_denied` | `modules/incidents/httpapi.authorizationDeniedError` |

## module.indicators

**Excluded 1.** Append lifecycle validates the addressed indicator and supporting envelopes. A missing support is ErrInvalidCreateRequest. It never applies an observation transition or addresses an observation/source.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `appendIndicatorStateInterval` | 409 `illegal_transition` | `modules/indicators/httpapi.mutationError` |
| `appendIndicatorStateInterval` | 404 `indicator_observation_not_found` | `modules/indicators/httpapi.observationNotFoundError` |
| `appendIndicatorStateInterval` | 404 `indicator_source_record_not_found` | `modules/indicators/httpapi.indicatorSourceNotFoundError` |
| `appendIndicatorStateInterval` | 404 `resolved_indicator_not_found` | `modules/indicators/httpapi.resolvedIndicatorNotFoundError` |

**Excluded 2.** CreateIndicatorObservation validates a source and optional resolved indicator before inserting. It neither loads an addressed observation nor changes an existing resolution state.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createManualIndicatorObservation` | 409 `illegal_transition` | `modules/indicators/httpapi.mutationError` |
| `createManualIndicatorObservation` | 404 `indicator_not_found` | `modules/indicators/httpapi.indicatorNotFoundError` |
| `createManualIndicatorObservation` | 404 `indicator_observation_not_found` | `modules/indicators/httpapi.observationNotFoundError` |

**Excluded 3.** The command addresses an observation. validatePriorObservationDependencies conceals missing or deleted dependencies as observation not found. Only resolve validates a new resolved indicator.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `dismissIndicatorObservation` | 404 `indicator_not_found` | `modules/indicators/httpapi.indicatorNotFoundError` |
| `dismissIndicatorObservation` | 404 `indicator_source_record_not_found` | `modules/indicators/httpapi.indicatorSourceNotFoundError` |
| `dismissIndicatorObservation` | 409 `record_deleted_use_restore` | `modules/indicators/httpapi.recordDeletedUseRestoreError` |
| `dismissIndicatorObservation` | 404 `resolved_indicator_not_found` | `modules/indicators/httpapi.resolvedIndicatorNotFoundError` |
| `resolveIndicatorObservation` | 404 `indicator_not_found` | `modules/indicators/httpapi.indicatorNotFoundError` |
| `resolveIndicatorObservation` | 404 `indicator_source_record_not_found` | `modules/indicators/httpapi.indicatorSourceNotFoundError` |
| `resolveIndicatorObservation` | 409 `record_deleted_use_restore` | `modules/indicators/httpapi.recordDeletedUseRestoreError` |
| `restoreIndicatorObservation` | 404 `indicator_not_found` | `modules/indicators/httpapi.indicatorNotFoundError` |
| `restoreIndicatorObservation` | 404 `indicator_source_record_not_found` | `modules/indicators/httpapi.indicatorSourceNotFoundError` |
| `restoreIndicatorObservation` | 409 `record_deleted_use_restore` | `modules/indicators/httpapi.recordDeletedUseRestoreError` |
| `restoreIndicatorObservation` | 404 `resolved_indicator_not_found` | `modules/indicators/httpapi.resolvedIndicatorNotFoundError` |

**Excluded 4.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `listIndicatorObservations` | 403 `authorization_denied` | `modules/indicators/httpapi.requireIncidentAdmission` |
| `listIndicatorStateIntervals` | 403 `authorization_denied` | `modules/indicators/httpapi.requireIncidentAdmission` |
| `listSourceRecordIndicatorObservations` | 403 `authorization_denied` | `modules/indicators/httpapi.requireIncidentAdmission` |

**Excluded 5.** The list handler selects its not-found constructor by the addressed record type. Its list query has no observation mutation error path.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `listIndicatorObservations` | 404 `indicator_observation_not_found` | `modules/indicators/httpapi.observationNotFoundError` |
| `listIndicatorObservations` | 404 `indicator_source_record_not_found` | `modules/indicators/httpapi.indicatorSourceNotFoundError` |
| `listIndicatorStateIntervals` | 404 `indicator_observation_not_found` | `modules/indicators/httpapi.observationNotFoundError` |
| `listIndicatorStateIntervals` | 404 `indicator_source_record_not_found` | `modules/indicators/httpapi.indicatorSourceNotFoundError` |
| `listSourceRecordIndicatorObservations` | 404 `indicator_not_found` | `modules/indicators/httpapi.indicatorNotFoundError` |
| `listSourceRecordIndicatorObservations` | 404 `indicator_observation_not_found` | `modules/indicators/httpapi.observationNotFoundError` |

## module.reference_data

**Excluded 1.** Verification rejection is a verification-result/job failure. Administrative activate maps failed retained integrity to ActivationRejection. Disable/remove do not verify, and refresh/reverify enqueue work.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `activateReferencePackVersion` | 409 `reference_pack_verification_failed` | `modules/reference_data.coordinatorAPIError` |
| `disableReferencePackVersion` | 409 `reference_pack_verification_failed` | `modules/reference_data.coordinatorAPIError` |
| `refreshReferencePacks` | 409 `reference_pack_verification_failed` | `modules/reference_data.coordinatorAPIError` |
| `removeReferencePackVersion` | 409 `reference_pack_verification_failed` | `modules/reference_data.coordinatorAPIError` |
| `reverifyReferencePackVersion` | 409 `reference_pack_verification_failed` | `modules/reference_data.coordinatorAPIError` |

**Excluded 2.** Only the activate case in applyAdministrativeAction creates ActivationRejection. The selected action uses a different case or job.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `disableReferencePackVersion` | 409 `reference_pack_activation_rejected` | `modules/reference_data.referencePackActivationRejected` |
| `importReferencePack` | 409 `reference_pack_activation_rejected` | `modules/reference_data.referencePackActivationRejected` |
| `refreshReferencePacks` | 409 `reference_pack_activation_rejected` | `modules/reference_data.referencePackActivationRejected` |
| `removeReferencePackVersion` | 409 `reference_pack_activation_rejected` | `modules/reference_data.referencePackActivationRejected` |
| `reverifyReferencePackVersion` | 409 `reference_pack_activation_rejected` | `modules/reference_data.referencePackActivationRejected` |

**Excluded 3.** The named admission decoder belongs to a different operation. Follow the selected job-kind contribution to the route-specific request reader.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `importReferencePack` | 400 `invalid_import_request` | `modules/imports.invalidImportRequest` |
| `importReferencePack` | 400 `invalid_release_request` | `modules/reporting.DecodeCreateReleaseRequest`, `modules/reporting.invalidReleaseRequest`, `modules/reporting.materializeOutputOptions` |
| `importReferencePack` | 400 `invalid_request` | `modules/reportcomposition.DecodePreviewRequest`, `modules/reportcomposition.schemaFieldError`, `modules/reportcomposition.validateGraphRefs` |

**Excluded 4.** The named producer is reporting.internalAPIError. The actual handler translates its own synchronous admission/store result. Even composition preview maps errors from its reporting port through reportcomposition.writeStoreError.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `importReferencePack` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |

## module.reportcomposition

**Excluded 1.** CreateDraft inserts a new composition after replay lookup. Existing-resource, retired-resource, draft-version and version-bound errors are produced by other store methods.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createReportComposition` | 409 `conflict` | `(*modules/reportcomposition.Service).writeStoreError` |
| `createReportComposition` | 404 `not_found` | `(*modules/reportcomposition.Service).writeStoreError` |

**Excluded 2.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `getReportComposition` | 403 `authorization_denied` | `modules/reportcomposition.reportCompositionAdmissionResult` |
| `getReportCompositionVersion` | 403 `authorization_denied` | `modules/reportcomposition.reportCompositionAdmissionResult` |
| `listReportCompositions` | 403 `authorization_denied` | `modules/reportcomposition.reportCompositionAdmissionResult` |
| `previewReportComposition` | 403 `authorization_denied` | `modules/reportcomposition.reportCompositionAdmissionResult` |
| `validateReportComposition` | 403 `authorization_denied` | `modules/reportcomposition.reportCompositionAdmissionResult` |

**Excluded 3.** The named admission decoder belongs to a different operation. Follow the selected job-kind contribution to the route-specific request reader.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `previewReportComposition` | 400 `invalid_release_request` | `modules/reporting.DecodeCreateReleaseRequest`, `modules/reporting.invalidReleaseRequest`, `modules/reporting.materializeOutputOptions` |

**Excluded 4.** The named producer is reporting.internalAPIError. The actual handler translates its own synchronous admission/store result. Even composition preview maps errors from its reporting port through reportcomposition.writeStoreError.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `previewReportComposition` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |

## module.reporting

**Excluded 1.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `approveRelease` | 403 `authorization_denied` | `modules/reporting.reportingAdmissionResult` |
| `getRelease` | 403 `authorization_denied` | `modules/reporting.reportingAdmissionResult` |
| `getSnapshot` | 403 `authorization_denied` | `modules/reporting.reportingAdmissionResult` |

**Excluded 2.** The admission result passes through snapshotVisibilityError or releaseVisibilityError, which replaces an invisible incident with the addressed resource error.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `approveRelease` | 404 `incident_not_found` | `modules/reporting.reportingAdmissionResult` |
| `createRelease` | 404 `incident_not_found` | `modules/reporting.reportingAdmissionResult` |
| `getRelease` | 404 `incident_not_found` | `modules/reporting.reportingAdmissionResult` |
| `getSnapshot` | 404 `incident_not_found` | `modules/reporting.reportingAdmissionResult` |
| `invalidateRelease` | 404 `incident_not_found` | `modules/reporting.reportingAdmissionResult` |
| `publishRelease` | 404 `incident_not_found` | `modules/reporting.reportingAdmissionResult` |

**Excluded 3.** The typed pack failure originates in snapshot/render admission and render preparation. This read or release-state action does not perform those operations.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `approveRelease` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |
| `getRelease` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |
| `getRetainedReportArtifactContent` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |
| `getRetainedReportArtifacts` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |
| `getSnapshot` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |
| `invalidateRelease` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |
| `publishRelease` | 409 `required_reference_pack_unavailable` | `modules/reporting.internalAPIError` |

**Excluded 4.** The named admission decoder belongs to a different operation. Follow the selected job-kind contribution to the route-specific request reader.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createRelease` | 400 `invalid_request` | `modules/reportcomposition.DecodePreviewRequest`, `modules/reportcomposition.schemaFieldError`, `modules/reportcomposition.validateGraphRefs` |

**Excluded 5.** Listing supplies an empty requested artifact. The missing-artifact response requires a nonempty requested artifact.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `getRetainedReportArtifacts` | 404 `report_artifact_not_found` | `(*modules/reporting.Service).serveRetainedReportArtifacts` |

**Excluded 6.** The approval policy runs in ApproveRelease. This action binds PublishRelease or InvalidateRelease.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `invalidateRelease` | 409 `release_approval_rejected` | `modules/reporting.releaseApprovalRejected` |
| `publishRelease` | 409 `release_approval_rejected` | `modules/reporting.releaseApprovalRejected` |

## module.revisions

**Excluded 1.** The addressed command selects its own decoder. The deleting flag selects its own transition, preconditions, and blocked error. Only restore uses the NOWAIT lock in the shared delete/restore store.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `deleteRecord` | 400 `invalid_rollback_request` | `modules/revisions/httpapi.invalidRollbackRequest` |
| `deleteRecord` | 409 `record_locked` | `modules/revisions/httpapi.recordLockedError` |
| `deleteRecord` | 409 `record_not_deleted` | `modules/revisions/httpapi.recordNotDeletedError` |
| `deleteRecord` | 409 `record_restore_blocked` | `modules/revisions/httpapi.recordRestoreBlockedError` |
| `restoreRecord` | 400 `invalid_rollback_request` | `modules/revisions/httpapi.invalidRollbackRequest` |
| `restoreRecord` | 409 `record_already_deleted` | `modules/revisions/httpapi.recordAlreadyDeletedError` |
| `restoreRecord` | 409 `record_delete_blocked` | `modules/revisions/httpapi.recordDeleteBlockedError` |
| `rollbackRecord` | 400 `invalid_mutation_payload` | `modules/revisions/httpapi.invalidMutationPayload` |

## module.savedviews

**Excluded 1.** getVisible and deleteVisible do not compare a base version. Only patchVisible constructs savedViewVersionConflictError.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `deleteIncidentSavedView` | 409 `saved_view_version_conflict` | `modules/savedviews.savedViewVersionConflictAPIError` |
| `getIncidentSavedView` | 409 `saved_view_version_conflict` | `modules/savedviews.savedViewVersionConflictAPIError` |

**Excluded 2.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `getIncidentSavedView` | 403 `authorization_denied` | `modules/savedviews.authorizationDeniedError` |

## module.timeline

**Excluded 1.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `getTimelineTimeConversionProfile` | 403 `authorization_denied` | `(*modules/timeline/admission.Service).requireIncidentRole` |

**Excluded 2.** MarkReviewed delegates to applyAction. A stale base produces newRowVersionConflict before the transition. Field-level merge classification runs on patch.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `markTimelineRecordReviewed` | 409 `same_field_conflict` | `modules/timeline/admission.sameFieldConflictAPIError` |

**Excluded 3.** PutTimeConversionProfile locks the profile, checks incident openness and base profile version, and upserts it. It performs no record transition, field merge, delete or client transaction replay.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `putTimelineTimeConversionProfile` | 409 `client_txn_conflict` | `platform/httpapi.ClientTxnConflictError` |
| `putTimelineTimeConversionProfile` | 409 `illegal_transition` | `modules/timeline/admission.illegalTransitionError` |
| `putTimelineTimeConversionProfile` | 409 `record_deleted_use_restore` | `modules/timeline/admission.recordDeletedUseRestoreError` |
| `putTimelineTimeConversionProfile` | 409 `same_field_conflict` | `modules/timeline/admission.sameFieldConflictAPIError` |

## module.workbook

**Excluded 1.** The only bulk provider builds Timeline scalar fill or tag assignment cells. These cannot request ambiguous entity resolution. Batch target classification maps deleted rows to not found, and per-cell stale conflicts go into the batch result.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `applyWorkbookBulkMutation` | 409 `entity_match_conflict` | `modules/workbook.entityMatchConflictError` |
| `applyWorkbookBulkMutation` | 409 `record_deleted_use_restore` | `modules/workbook.recordDeletedAPIError` |
| `applyWorkbookBulkMutation` | 409 `same_field_conflict` | `modules/workbook.mutationFailureAPIError` |

**Excluded 2.** Workbook assembly binds bulk to Timeline, clipboard to Timeline/Entities, and supersede to Timeline/Decisions. None uses the Party or Evidence mutation classifier.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `applyWorkbookBulkMutation` | 409 `evidence_attach_rejected` | `modules/workbook.mutationFailureAPIError` |
| `applyWorkbookBulkMutation` | 409 `party_match_conflict` | `modules/workbook.mutationFailureAPIError` |
| `pasteWorkbookClipboard` | 409 `evidence_attach_rejected` | `modules/workbook.mutationFailureAPIError` |
| `pasteWorkbookClipboard` | 409 `party_match_conflict` | `modules/workbook.mutationFailureAPIError` |
| `supersedeRecord` | 409 `evidence_attach_rejected` | `modules/workbook.mutationFailureAPIError` |
| `supersedeRecord` | 409 `party_match_conflict` | `modules/workbook.mutationFailureAPIError` |

**Excluded 3.** The only bound linked-note provider calls Artifacts.CreateContextualNote. Its failure translator has no Entity, Party or Evidence subtype.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createRecordLinkedNote` | 409 `entity_match_conflict` | `modules/workbook.entityMatchConflictError` |
| `createRecordLinkedNote` | 409 `evidence_attach_rejected` | `modules/workbook.mutationFailureAPIError` |
| `createRecordLinkedNote` | 409 `party_match_conflict` | `modules/workbook.mutationFailureAPIError` |

**Excluded 4.** CreateContextualNote delegates to Artifacts create with a locked source and creates the note at version one. There is no base version, field merge or lifecycle action in this transaction.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createRecordLinkedNote` | 409 `illegal_transition` | `modules/workbook.illegalTransitionAPIError` |
| `createRecordLinkedNote` | 409 `row_version_conflict` | `modules/workbook.rowVersionConflictError` |
| `createRecordLinkedNote` | 409 `same_field_conflict` | `modules/workbook.mutationFailureAPIError` |

**Excluded 5.** Ordinary create addresses a view and incident. Artifacts.Create passes no linked-note source. Its optional coordination source converts deletedAt to a validation error. Timeline create uses batch reference scope. Entity matches exclude deleted rows, and the Party, Evidence and Task/Decision create paths never call the addressed-record deletion classifier.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createViewRow` | 409 `record_deleted_use_restore` | `modules/workbook.recordDeletedAPIError` |

**Excluded 6.** The create providers do not admit a base row version or run a patch conflict window. Entity and Party upserts match or reuse records under their own identity and idempotency rules. The named conflict subtypes come from update/resolution.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `createViewRow` | 409 `row_version_conflict` | `modules/workbook.rowVersionConflictError` |
| `createViewRow` | 409 `same_field_conflict` | `modules/workbook.mutationFailureAPIError` |

**Excluded 7.** The operation admits RolesMember. The insufficient-role branch in the shared admission translator requires a restricted role set.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `getCurrentUserWorkbookPreferences` | 403 `authorization_denied` | `(*modules/workbook.service).requireIncidentRole` |
| `getIncidentDefaultWorkbookPreferences` | 403 `authorization_denied` | `(*modules/workbook.service).requireIncidentRole` |
| `getIncidentWorkbookStartup` | 403 `authorization_denied` | `(*modules/workbook.service).requireIncidentRole` |
| `locateWorkbookViewRecord` | 403 `authorization_denied` | `(*modules/workbook.service).requireIncidentRole` |
| `putCurrentUserWorkbookPreferences` | 403 `authorization_denied` | `(*modules/workbook.service).requireIncidentRole` |
| `queryWorkbookView` | 403 `authorization_denied` | `(*modules/workbook.service).requireIncidentRole` |

**Excluded 8.** noteAssociationReadFailure returns only TargetNotFoundFailure. The route admits all member roles and translates other read failures internally.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `listNoteAssociations` | 403 `authorization_denied` | `(*modules/workbook.service).requireIncidentRole`, `modules/workbook.mutationFailureAPIError` |
| `listNoteAssociations` | 409 `client_txn_conflict` | `platform/httpapi.ClientTxnConflictError` |
| `listNoteAssociations` | 409 `entity_match_conflict` | `modules/workbook.entityMatchConflictError` |
| `listNoteAssociations` | 409 `evidence_attach_rejected` | `modules/workbook.mutationFailureAPIError` |
| `listNoteAssociations` | 409 `illegal_transition` | `modules/workbook.illegalTransitionAPIError` |
| `listNoteAssociations` | 409 `incident_closed` | `modules/workbook.incidentClosedError` |
| `listNoteAssociations` | 400 `invalid_mutation_payload` | `modules/workbook.invalidMutationPayloadWithDetails` |
| `listNoteAssociations` | 409 `party_match_conflict` | `modules/workbook.mutationFailureAPIError` |
| `listNoteAssociations` | 409 `record_deleted_use_restore` | `modules/workbook.recordDeletedAPIError` |
| `listNoteAssociations` | 409 `row_version_conflict` | `modules/workbook.rowVersionConflictError` |
| `listNoteAssociations` | 409 `same_field_conflict` | `modules/workbook.mutationFailureAPIError` |

**Excluded 9.** NoteAssociations.Apply checks the note base version and validates link actions. Its conflict is a row-version error. It runs no field merge or lifecycle transition.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `mutateNoteAssociations` | 409 `entity_match_conflict` | `modules/workbook.entityMatchConflictError` |
| `mutateNoteAssociations` | 409 `evidence_attach_rejected` | `modules/workbook.mutationFailureAPIError` |
| `mutateNoteAssociations` | 409 `illegal_transition` | `modules/workbook.illegalTransitionAPIError` |
| `mutateNoteAssociations` | 409 `party_match_conflict` | `modules/workbook.mutationFailureAPIError` |
| `mutateNoteAssociations` | 409 `same_field_conflict` | `modules/workbook.mutationFailureAPIError` |

**Excluded 10.** Timeline batch scope maps a deleted target to not found and returns per-cell conflicts in the result. Host/Identity clipboard uses identifier upsert without addressed-row patch conflict admission.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `pasteWorkbookClipboard` | 409 `record_deleted_use_restore` | `modules/workbook.recordDeletedAPIError` |
| `pasteWorkbookClipboard` | 409 `same_field_conflict` | `modules/workbook.mutationFailureAPIError` |

**Excluded 11.** The Evidence adapter shares evidenceMutationFailure across owner commands. The blob precondition and AttachRejectedError construction belong to Evidence create/attachment, outside patch and field resolution.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `patchRecord` | 409 `evidence_attach_rejected` | `modules/workbook.mutationFailureAPIError` |
| `resolveRecordSameFieldConflict` | 409 `evidence_attach_rejected` | `modules/workbook.mutationFailureAPIError` |

**Excluded 12.** The two providers perform Timeline and Decision lifecycle actions. They address an existing record and check its base version. Neither action resolves entity identifiers or opens a scalar conflict window.

| Operation | Registry status and code | Producers |
| --- | --- | --- |
| `supersedeRecord` | 409 `entity_match_conflict` | `modules/workbook.entityMatchConflictError` |
| `supersedeRecord` | 409 `same_field_conflict` | `modules/workbook.mutationFailureAPIError` |
