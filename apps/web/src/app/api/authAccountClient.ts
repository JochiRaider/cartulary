import type {
  BeginEnterpriseAuthRequest,
  BeginEnterpriseAuthResponse,
  BeginTOTPEnrollmentRequest,
  BeginTOTPEnrollmentResponse,
  ChangeCurrentPasswordRequest,
  ChangeCurrentPasswordResponse,
  CompleteTOTPEnrollmentRequest,
  CompleteTOTPEnrollmentResponse,
  GetCredentialStateResponse,
  GetCurrentAccountPreferencesResponse,
  GetCurrentAccountProfileResponse,
  GetCurrentSessionResponse,
  ListDeploymentExtensionsResponse,
  ListEnterpriseAuthProvidersResponse,
  LoginLocalUserRequest,
  LoginLocalUserResponse,
  LogoutCurrentSessionResponse,
  PatchCurrentAccountProfileRequest,
  PatchCurrentAccountProfileResponse,
  PutCurrentAccountPreferencesRequest,
  PutCurrentAccountPreferencesResponse,
} from "@cartulary/protocol-ts/http";
import { validateHTTPOperationResponse } from "@cartulary/protocol-ts/http";
import { clientTxnID, fetchHTTPOperation } from "../../services/browserApi";
import type {
  AccountPreferencesResource,
  AccountProfileResource,
  DensityMode,
} from "./publicHttpTypes";
export type TotpAuthMode = "bootstrap" | "session";
type ShellGetOptions = { apiBase?: string | undefined; signal?: AbortSignal };

function secondFactorPayload(code: string) {
  if (code.trim() === "") {
    return undefined;
  }
  return {
    kind: "totp" as const,
    assertion: {
      code,
    },
  };
}

function bootstrapAuthorizationHeader(options: {
  authMode: TotpAuthMode;
  bootstrapToken?: string | undefined;
}): HeadersInit | undefined {
  if (options.authMode !== "bootstrap") {
    return undefined;
  }
  const token = options.bootstrapToken?.trim() ?? "";
  if (token === "") {
    return undefined;
  }
  return {
    Authorization: `Bearer ${token}`,
  };
}

function totpEnrollmentRequestInit(options: {
  authMode: TotpAuthMode;
  bootstrapToken?: string | undefined;
  body: Record<string, unknown>;
}): RequestInit {
  const headers = bootstrapAuthorizationHeader(options);
  return {
    method: "POST",
    credentials: options.authMode === "bootstrap" ? "omit" : "include",
    body: JSON.stringify(options.body),
    ...(headers === undefined ? {} : { headers }),
  };
}

export function loadSession(options: ShellGetOptions = {}) {
  return fetchHTTPOperation<GetCurrentSessionResponse>({
    apiBase: options.apiBase,
    operationID: "getCurrentSession",
    init: options.signal === undefined ? undefined : { signal: options.signal },
  });
}

export function loadCredentialState(options: ShellGetOptions = {}) {
  return fetchHTTPOperation<GetCredentialStateResponse>({
    apiBase: options.apiBase,
    operationID: "getCredentialState",
    init: options.signal === undefined ? undefined : { signal: options.signal },
  });
}

export function loadAccountProfile(options?: ShellGetOptions) {
  return fetchHTTPOperation<GetCurrentAccountProfileResponse>({
    apiBase: options?.apiBase,
    init:
      typeof options?.signal === "undefined"
        ? undefined
        : {
            signal: options.signal,
          },
    operationID: "getCurrentAccountProfile",
  });
}

export function patchAccountProfile(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId: string;
  signal?: AbortSignal;
  displayName: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id: options.clientTxnId,
    display_name: options.displayName,
  } satisfies PatchCurrentAccountProfileRequest;
  return fetchHTTPOperation<PatchCurrentAccountProfileResponse>({
    apiBase: options.apiBase,
    init: {
      method: "PATCH",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
    operationID: "patchCurrentAccountProfile",
  });
}

export function loadAccountPreferences(options?: ShellGetOptions) {
  return fetchHTTPOperation<GetCurrentAccountPreferencesResponse>({
    apiBase: options?.apiBase,
    init:
      typeof options?.signal === "undefined"
        ? undefined
        : {
            signal: options.signal,
          },
    operationID: "getCurrentAccountPreferences",
  });
}

export function putAccountPreferences(options: {
  apiBase?: string | undefined;
  basePreferencesVersion: number;
  clientTxnId: string;
  signal?: AbortSignal;
  densityMode: DensityMode | null;
}) {
  const request = {
    base_preferences_version: options.basePreferencesVersion,
    client_txn_id: options.clientTxnId,
    density_mode: options.densityMode,
  } satisfies PutCurrentAccountPreferencesRequest;
  return fetchHTTPOperation<PutCurrentAccountPreferencesResponse>({
    apiBase: options.apiBase,
    init: {
      method: "PUT",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
    operationID: "putCurrentAccountPreferences",
  });
}

export function loadExtensions(options?: ShellGetOptions) {
  return fetchHTTPOperation<ListDeploymentExtensionsResponse>({
    apiBase: options?.apiBase,
    init:
      typeof options?.signal === "undefined"
        ? undefined
        : {
            signal: options.signal,
          },
    operationID: "listDeploymentExtensions",
  });
}

export function loginLocal(options: {
  apiBase?: string | undefined;
  password: string;
  secondFactorCode?: string;
  username: string;
}) {
  const secondFactor = secondFactorPayload(options.secondFactorCode ?? "");
  return fetchHTTPOperation<LoginLocalUserResponse>({
    apiBase: options.apiBase,
    operationID: "loginLocalUser",
    init: {
      method: "POST",
      body: JSON.stringify({
        username: options.username,
        password: options.password,
        ...(secondFactor === undefined
          ? {}
          : {
              second_factor: secondFactor,
            }),
      } satisfies LoginLocalUserRequest),
    },
  });
}

export function listEnterpriseAuthProviders(options: ShellGetOptions = {}) {
  return fetchHTTPOperation<ListEnterpriseAuthProvidersResponse>({
    apiBase: options.apiBase,
    operationID: "listEnterpriseAuthProviders",
    init: options.signal === undefined ? undefined : { signal: options.signal },
  });
}

export function beginEnterpriseAuth(options: {
  apiBase?: string | undefined;
  providerKey: string;
  returnTo?: string | undefined;
}) {
  return fetchHTTPOperation<BeginEnterpriseAuthResponse>({
    apiBase: options.apiBase,
    operationID: "beginEnterpriseAuth",
    pathParameters: { provider_key: options.providerKey },
    init: {
      method: "POST",
      body: JSON.stringify({
        return_to: options.returnTo ?? "/",
      } satisfies BeginEnterpriseAuthRequest),
    },
  });
}

export function logoutCurrentSession(options: ShellGetOptions = {}) {
  return fetchHTTPOperation<LogoutCurrentSessionResponse>({
    apiBase: options.apiBase,
    operationID: "logoutCurrentSession",
    init: { method: "POST" },
  });
}

export function beginTotpEnrollment(options: {
  apiBase?: string | undefined;
  authMode: TotpAuthMode;
  bootstrapToken?: string;
  clientTxnId?: string;
  currentFactorCode?: string;
  currentPassword?: string;
}) {
  const secondFactor = secondFactorPayload(options.currentFactorCode ?? "");
  const requestInit = totpEnrollmentRequestInit({
    authMode: options.authMode,
    bootstrapToken: options.bootstrapToken,
    body: {
      client_txn_id:
        options.clientTxnId ?? clientTxnID("authentication-ui-totp-begin"),
      ...(options.authMode === "session"
        ? {
            current_password: options.currentPassword ?? "",
            ...(secondFactor === undefined
              ? {}
              : { second_factor: secondFactor }),
          }
        : {}),
    } satisfies BeginTOTPEnrollmentRequest,
  });
  return fetchHTTPOperation<BeginTOTPEnrollmentResponse>({
    apiBase: options.apiBase,
    operationID: "beginTOTPEnrollment",
    init: requestInit,
  });
}

export function completeTotpEnrollment(options: {
  apiBase?: string | undefined;
  authMode: TotpAuthMode;
  bootstrapToken?: string;
  clientTxnId?: string;
  code: string;
  enrollmentId: string;
}) {
  const requestInit = totpEnrollmentRequestInit({
    authMode: options.authMode,
    bootstrapToken: options.bootstrapToken,
    body: {
      client_txn_id:
        options.clientTxnId ?? clientTxnID("authentication-ui-totp-complete"),
      enrollment_id: options.enrollmentId,
      code: options.code,
    } satisfies CompleteTOTPEnrollmentRequest,
  });
  return fetchHTTPOperation<CompleteTOTPEnrollmentResponse>({
    apiBase: options.apiBase,
    operationID: "completeTOTPEnrollment",
    init: requestInit,
  });
}

export function changePassword(options: {
  apiBase?: string | undefined;
  clientTxnId?: string;
  currentPassword: string;
  newPassword: string;
  secondFactorCode?: string;
}) {
  const secondFactor = secondFactorPayload(options.secondFactorCode ?? "");
  return fetchHTTPOperation<ChangeCurrentPasswordResponse>({
    apiBase: options.apiBase,
    operationID: "changeCurrentPassword",
    init: {
      method: "POST",
      body: JSON.stringify({
        client_txn_id:
          options.clientTxnId ??
          clientTxnID("authentication-ui-password-change"),
        current_password: options.currentPassword,
        new_password: options.newPassword,
        ...(secondFactor === undefined ? {} : { second_factor: secondFactor }),
      } satisfies ChangeCurrentPasswordRequest),
    },
  });
}

export function isAccountProfileResource(
  value: unknown,
): value is AccountProfileResource {
  return validateHTTPOperationResponse("getCurrentAccountProfile", {
    data: value,
    meta: { request_id: "account-resource-acceptance" },
  }).ok;
}

export function isAccountPreferencesResource(
  value: unknown,
): value is AccountPreferencesResource {
  return validateHTTPOperationResponse("getCurrentAccountPreferences", {
    data: value,
    meta: { request_id: "account-resource-acceptance" },
  }).ok;
}
