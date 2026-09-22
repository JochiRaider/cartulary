# app/

[Source overview / parent](../README.md)

Application shell, route entry surfaces, authentication, account settings, and incident/deployment administration.

Session acceptance belongs to [appSessionController.ts](appSessionController.ts);
[useAppSession.ts](useAppSession.ts) binds its snapshot and lifetime to React.
Session discovery and supporting reads have independent 30-second observation
bounds. Preference failures remain local, unresolved extension discovery does
not admit extension actions, and credential reads belong to account security.

Same-account reauthentication retains workbook pending edits while replay is
paused. Replacement-shell authorization and an accepted current-surface query
precede replay. Account replacement retires previous protected state and pending
work. Directory and incident-creation recovery have separate lifetimes.

Workbook internals are documented in the [workbook guide](../workbook/README.md).

## Subdirectories

| Directory | Responsibility |
| --- | --- |
| [api/](api/README.md) | App-shell operation adapters and private protocol types for authentication, accounts, and administration. |

## Shell and navigation

| File | Responsibility |
| --- | --- |
| [App.auth.support.test.tsx](App.auth.support.test.tsx) | Account security and deployment-user adapter integration tests. |
| [App.auth.test.tsx](App.auth.test.tsx) | Authentication app behavior tests for auth/account/route readiness. |
| [App.landing.test.tsx](App.landing.test.tsx) | Landing-surface tests for app startup and landing interactions. |
| [App.timeline-invalidation.support.test.tsx](App.timeline-invalidation.support.test.tsx) | Tests self-originated Timeline invalidation suppression without moving draft-row focus. |
| [App.tsx](App.tsx) | Top-level application composition and route/surface selection for the web app. |
| [AppRoot.tsx](AppRoot.tsx) | Root React wrapper that connects app-level providers and the rendered `App`. |
| [fontBundle.test.ts](fontBundle.test.ts) | Font bundle availability and packaging boundary tests. |
| [fontRoles.test.tsx](fontRoles.test.tsx) | Font-role presentation tests for app and workbook surfaces. |
| [IncidentLanding.tsx](IncidentLanding.tsx) | Incident directory landing panel, search/filter controls, and create-incident dialog. |
| [LandingAdminDisplay.tsx](LandingAdminDisplay.tsx) | Shared display helpers for landing/admin panels. |
| [LandingAdminLayout.tsx](LandingAdminLayout.tsx) | Landing and deployment-administration shell layout plus the account/application menu. |
| [landingAdminStyles.ts](landingAdminStyles.ts) | Shared style constants for landing/admin panels. |
| [landingAdminTypes.ts](landingAdminTypes.ts) | Shared landing/admin TypeScript model types. |
| [otelBoundary.test.ts](otelBoundary.test.ts) | OpenTelemetry import and runtime-boundary tests. |
| [routeState.test.ts](routeState.test.ts) | Route-state parsing and history write tests. |
| [routeState.ts](routeState.ts) | Pure app-route parsing and history URL construction helpers. |
| [useAppRouteRuntime.ts](useAppRouteRuntime.ts) | React hook for route state, popstate handling, and history writes. |

## Authentication and accounts

| File | Responsibility |
| --- | --- |
| [AccountApplicationMenu.test.tsx](AccountApplicationMenu.test.tsx) | Tests menu keyboard handling, nested navigation, focus handoff, and focus restoration. |
| [AccountApplicationMenu.tsx](AccountApplicationMenu.tsx) | Account and application menu with nested navigation and semantic focus handoff. |
| [accountAuthenticationLifecycle.test.ts](accountAuthenticationLifecycle.test.ts) | Tests bounded session observations, outstanding login admission, and security transport retirement. |
| [AccountDialog.tsx](AccountDialog.tsx) | Shared account-action dialog with dismissal and focus restoration. |
| [accountFrontendBoundaries.test.tsx](accountFrontendBoundaries.test.tsx) | Tests account dialog dismissal, focus restoration, and single-source error announcements. |
| [accountInputValidation.ts](accountInputValidation.ts) | Account email and provisioning-password validation shared by account controls. |
| [accountOperation.ts](accountOperation.ts) | Normalizes account-operation failures into safe public errors. |
| [accountPanelStyles.ts](accountPanelStyles.ts) | Shared account-panel form, card, action, and feedback styles. |
| [accountSecurityModel.ts](accountSecurityModel.ts) | Credential and security-operation controller bound to the current account session. |
| [AccountSecurityPanel.tsx](AccountSecurityPanel.tsx) | Account security controls for credentials, password changes, TOTP, and session actions. |
| [AccountSettingsDialog.tsx](AccountSettingsDialog.tsx) | Account-settings dialog that hosts profile, appearance, and security panels. |
| [accountSettingsEditing.test.tsx](accountSettingsEditing.test.tsx) | Tests profile retry, accessible field validation, and draft preservation during refresh. |
| [accountSettingsModel.test.ts](accountSettingsModel.test.ts) | Tests account edit admission, revision acknowledgements, conflicts, and explicit review. |
| [accountSettingsModel.ts](accountSettingsModel.ts) | Profile and appearance draft state, validation, review, and save coordination. |
| [AccountSettingsPanels.tsx](AccountSettingsPanels.tsx) | Account profile and appearance settings panels. |
| [appSessionController.test.tsx](appSessionController.test.tsx) | Tests session discovery, anonymous confirmation, supporting-read isolation, and replacement fencing. |
| [appSessionController.ts](appSessionController.ts) | Accepts session observations and coordinates account lifetime, preferences, and extension discovery. |
| [authenticationModel.ts](authenticationModel.ts) | Local and enterprise authentication controller with safe error and navigation handling. |
| [AuthGateway.tsx](AuthGateway.tsx) | Authentication-state gate around app content and login/account readiness. |
| [useAccountMenuLayout.ts](useAccountMenuLayout.ts) | Positions the account menu within the effective viewport while preserving DOM tab order. |
| [useAccountSecurity.ts](useAccountSecurity.ts) | React binding for the account security controller and its session lifetime. |
| [useAppSession.ts](useAppSession.ts) | Binds accepted session snapshots and controller lifetime to React. |
| [useAuthentication.ts](useAuthentication.ts) | React binding for authentication commands, local inputs, and TOTP normalization. |

## Shared app coordination

| File | Responsibility |
| --- | --- |
| [appDepartureReview.test.ts](appDepartureReview.test.ts) | Tests sequential owner departure review, Stay cancellation, and newly outstanding work. |
| [appDepartureReview.ts](appDepartureReview.ts) | Sequences outstanding owner reviews before app navigation can proceed. |
| [IncidentAdminPanel.test.tsx](IncidentAdminPanel.test.tsx) | Incident administration panel tests. |
| [IncidentAdminPanel.tsx](IncidentAdminPanel.tsx) | Incident administration panel UI for incident metadata, preferences, membership, and audit affordances. |
| [incidentResourceController.ts](incidentResourceController.ts) | Shared incident summary observation and reconciliation bound to incident authority. |

## Incident directory and creation

| File | Responsibility |
| --- | --- |
| [IncidentCreationForm.tsx](IncidentCreationForm.tsx) | Incident-creation form with field errors, pending feedback, and recovery actions. |
| [incidentCreationModel.test.ts](incidentCreationModel.test.ts) | Tests exact creation payloads, synchronous admission, uncertain replay, and late-response rejection. |
| [incidentCreationModel.ts](incidentCreationModel.ts) | Incident-creation drafts, request capture, operation recovery, and successful workbook handoff. |
| [incidentDirectoryModel.test.ts](incidentDirectoryModel.test.ts) | Tests search debounce, immediate filter application, continuation, and obsolete-response suppression. |
| [incidentDirectoryModel.ts](incidentDirectoryModel.ts) | Incident directory query controller with search, filtering, continuation, and read-state feedback. |
| [useIncidentCreation.ts](useIncidentCreation.ts) | React binding for incident-creation drafts, submission, recovery, and navigation handoff. |
| [useIncidentDirectory.ts](useIncidentDirectory.ts) | React binding for membership-scoped incident directory queries and continuation. |

## Incident metadata

| File | Responsibility |
| --- | --- |
| [incidentMetadataCharacterization.test.tsx](incidentMetadataCharacterization.test.tsx) | Tests sparse metadata saves, duplicate admission, and acknowledgement of captured revisions. |
| [incidentMetadataController.test.ts](incidentMetadataController.test.ts) | Tests sparse field construction, immutable save attempts, and authority-bound admission. |
| [incidentMetadataController.ts](incidentMetadataController.ts) | Incident metadata draft ownership, sparse saves, conflict review, and acknowledgement reconciliation. |
| [incidentMetadataLifecycle.test.tsx](incidentMetadataLifecycle.test.tsx) | Tests metadata drafts across drawer transitions, role changes, access loss, and session replacement. |
| [incidentMetadataModel.ts](incidentMetadataModel.ts) | Incident metadata values, field revisions, authority, drafts, and review models. |
| [IncidentMetadataPanel.tsx](IncidentMetadataPanel.tsx) | Incident metadata editing, save/review feedback, and departure-dialog composition. |
| [useIncidentMetadata.ts](useIncidentMetadata.ts) | React binding for metadata drafts, incident summary reconciliation, and departure review. |

## Incident lifecycle

| File | Responsibility |
| --- | --- |
| [incidentLifecycleCharacterization.test.tsx](incidentLifecycleCharacterization.test.tsx) | Tests lifecycle action deduplication, uncertain replay, and distinct conflict recovery paths. |
| [incidentLifecycleController.test.ts](incidentLifecycleController.test.ts) | Tests reviewed lifecycle admission, current authority, secure identity, and exact replay. |
| [incidentLifecycleController.ts](incidentLifecycleController.ts) | Incident close/reopen operation ownership, reviewed request capture, replay, and recovery. |
| [incidentLifecycleIntegration.test.tsx](incidentLifecycleIntegration.test.tsx) | Tests lifecycle draft retention across drawer transitions, role changes, and session replacement. |
| [incidentLifecycleModel.ts](incidentLifecycleModel.ts) | Lifecycle authority, action drafts, reviews, captured attempts, and problem-state models. |
| [IncidentLifecyclePanel.tsx](IncidentLifecyclePanel.tsx) | Incident lifecycle controls, action review, recovery, and departure-dialog composition. |
| [useIncidentLifecycle.ts](useIncidentLifecycle.ts) | React binding for incident lifecycle ownership, authority updates, and departure review. |

## Incident membership management

| File | Responsibility |
| --- | --- |
| [incidentMembershipManagementCharacterization.test.tsx](incidentMembershipManagementCharacterization.test.tsx) | Tests membership continuation, repeated-create admission, and newer-draft preservation. |
| [incidentMembershipManagementController.test.ts](incidentMembershipManagementController.test.ts) | Tests bounded membership browsing, request fencing, cursor cycles, and refresh recovery. |
| [incidentMembershipManagementController.ts](incidentMembershipManagementController.ts) | Membership browsing and reviewed mutation ownership with retained drafts and operation recovery. |
| [incidentMembershipManagementLifecycle.test.tsx](incidentMembershipManagementLifecycle.test.tsx) | Tests membership draft lifetime, role-loss retirement, and account or incident access boundaries. |
| [incidentMembershipManagementModel.ts](incidentMembershipManagementModel.ts) | Membership page, authority, draft, review, and mutation-attempt models. |
| [IncidentMembershipManagementPanel.test.tsx](IncidentMembershipManagementPanel.test.tsx) | Tests reviewed membership writes, permission-sensitive controls, and dirty-input departure review. |
| [IncidentMembershipManagementPanel.tsx](IncidentMembershipManagementPanel.tsx) | Membership browsing, add/edit/remove review, operation recovery, and departure controls. |
| [useIncidentMembershipManagement.ts](useIncidentMembershipManagement.ts) | React binding for membership browsing, reviewed mutations, and departure review. |

## Incident membership audit

| File | Responsibility |
| --- | --- |
| [incidentMembershipAuditCharacterization.test.tsx](incidentMembershipAuditCharacterization.test.tsx) | Tests audit loading presentation, accepted-query continuation, and duplicate-read coalescing. |
| [incidentMembershipAuditController.test.ts](incidentMembershipAuditController.test.ts) | Tests membership audit query replacement, stale-response fencing, and bounded cursor history. |
| [incidentMembershipAuditController.ts](incidentMembershipAuditController.ts) | Incident membership audit read ownership with applied filters and bounded page navigation. |
| [incidentMembershipAuditLifecycle.test.tsx](incidentMembershipAuditLifecycle.test.tsx) | Tests audit drawer lifetime, role-loss cleanup, and incident-scoped access retirement. |
| [incidentMembershipAuditModel.test.ts](incidentMembershipAuditModel.test.ts) | Tests membership audit vocabulary, timestamp precision, target validation, and authority binding. |
| [incidentMembershipAuditModel.ts](incidentMembershipAuditModel.ts) | Membership audit filter validation, paging, authority, and read-state models. |
| [IncidentMembershipAuditPanel.test.tsx](IncidentMembershipAuditPanel.test.tsx) | Tests incident audit placement, exact-filter validation, and stale or unavailable read feedback. |
| [IncidentMembershipAuditPanel.tsx](IncidentMembershipAuditPanel.tsx) | Incident membership audit filters, event browsing, paging, and read-state presentation. |
| [useIncidentMembershipAudit.ts](useIncidentMembershipAudit.ts) | React binding for membership audit activation, applied filters, and read lifetime. |

## Import workflows

| File | Responsibility |
| --- | --- |
| [incidentImportHardening.test.ts](incidentImportHardening.test.ts) | Tests malformed import outcomes, retained result actions, and fresh-read admission before opening. |
| [incidentImportModel.test.ts](incidentImportModel.test.ts) | Tests immutable import replay, uncertain admission, read recovery, and retained validated status. |
| [incidentImportModel.ts](incidentImportModel.ts) | Session-bound incident import admission, observation, cancellation, and result-access coordination. |
| [IncidentImportPanel.test.tsx](IncidentImportPanel.test.tsx) | Tests single import admission, uncertain-response recovery, and file-replacement locking. |
| [IncidentImportPanel.tsx](IncidentImportPanel.tsx) | Incident bundle import panel and import-job polling controls. |
| [incidentImportState.test.ts](incidentImportState.test.ts) | Tests immutable import transitions, terminal-result consistency, protected reset, and unavailable state. |
| [incidentImportState.ts](incidentImportState.ts) | Pure incident import state transitions, authority checks, and action availability. |
| [useIncidentImport.ts](useIncidentImport.ts) | React binding and presentation projection for incident import ownership and recovery. |
| [useNetworkFlowImport.test.tsx](useNetworkFlowImport.test.tsx) | Tests analytical import intent retention, claim withdrawal, and account/incident replacement fencing. |
| [useNetworkFlowImport.ts](useNetworkFlowImport.ts) | App-lifetime binding for Network Flow import authority, retained work, and surface handoff. |
| [useWorkbookImport.test.tsx](useWorkbookImport.test.tsx) | Tests workbook import retention on role loss and retirement on profile or session replacement. |
| [useWorkbookImport.ts](useWorkbookImport.ts) | App-lifetime binding for workbook import authority and protected workflow state. |

## Deployment-user administration

| File | Responsibility |
| --- | --- |
| [DeploymentUserActionDialog.tsx](DeploymentUserActionDialog.tsx) | Reviewed deployment-user action dialog for the selected administrative operation. |
| [DeploymentUserLeaveDialog.tsx](DeploymentUserLeaveDialog.tsx) | Departure review for unresolved deployment-user edits and operations. |
| [deploymentUsersModel.test.ts](deploymentUsersModel.test.ts) | Tests user browsing activation, uncertain-operation review, stale pages, and cursor recovery. |
| [deploymentUsersModel.ts](deploymentUsersModel.ts) | Deployment-user browsing and administrative-operation controller tied to session authority. |
| [DeploymentUsersPanel.tsx](DeploymentUsersPanel.tsx) | Deployment-user browsing, selection, editing, and administrative action controls. |
| [useDeploymentUsers.ts](useDeploymentUsers.ts) | React binding for deployment-user browsing and administrative operations. |

## Deployment audit

| File | Responsibility |
| --- | --- |
| [administrativeAuditController.test.ts](administrativeAuditController.test.ts) | Tests draft/applied/accepted audit state, stale-response suppression, and bounded paging. |
| [administrativeAuditController.ts](administrativeAuditController.ts) | Deployment audit read controller with applied queries, paging, and injected transport ports. |
| [administrativeAuditModel.test.ts](administrativeAuditModel.test.ts) | Tests exact audit query vocabulary, filter dependencies, and timestamp boundaries. |
| [administrativeAuditModel.ts](administrativeAuditModel.ts) | Deployment audit query, authority, paging, validation, and read-state models. |
| [DeploymentAuditPanel.test.tsx](DeploymentAuditPanel.test.tsx) | Tests applied audit filters, native form submission, continuation, and read feedback. |
| [DeploymentAuditPanel.tsx](DeploymentAuditPanel.tsx) | Deployment administrative audit panel and audit-event formatting. |
| [useAdministrativeAudit.ts](useAdministrativeAudit.ts) | React binding for deployment audit activation, query state, and page commands. |

## Reference Pack administration

| File | Responsibility |
| --- | --- |
| [referencePackAdminController.test.ts](referencePackAdminController.test.ts) | Tests Reference Pack operation capture, replay reconciliation, and capability-loss retirement. |
| [referencePackAdminController.ts](referencePackAdminController.ts) | Reference Pack catalog reads, upload/reload operations, job observation, and authority retirement. |
| [referencePackAdminModel.test.ts](referencePackAdminModel.test.ts) | Tests Reference Pack read generations, version identity, ordering, and search normalization. |
| [referencePackAdminModel.ts](referencePackAdminModel.ts) | Reference Pack query, resource identity, authority, paging, and operation-state models. |
| [ReferencePackAdminPanel.test.tsx](ReferencePackAdminPanel.test.tsx) | Reference-pack administration panel tests. |
| [ReferencePackAdminPanel.tsx](ReferencePackAdminPanel.tsx) | Reference-pack administration panel UI for import, reload, cancellation, and job-status controls. |
| [useReferencePackAdmin.ts](useReferencePackAdmin.ts) | React binding and presentation projection for Reference Pack administration. |

## Workbook preferences and saved views

| File | Responsibility |
| --- | --- |
| [useWorkbookPreferences.ts](useWorkbookPreferences.ts) | App-lifetime binding for workbook preference ownership and departure review. |
| [useWorkbookSavedViews.ts](useWorkbookSavedViews.ts) | App-lifetime binding for saved-view controller state and operations. |
| [WorkbookPreferenceDepartureDialog.tsx](WorkbookPreferenceDepartureDialog.tsx) | Departure review for outstanding workbook preference edits. |
| [workbookPreferenceIntegration.test.tsx](workbookPreferenceIntegration.test.tsx) | Tests preference work across drawer transitions, role changes, and session replacement. |

Incident presentation test compositions live in [testing](../testing/README.md).
