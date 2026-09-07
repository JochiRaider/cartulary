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
import { fetchHTTPOperation } from "../../services/browserApi";
import type {
  AccountPreferencesResource,
  AccountProfileResource,
  DensityMode,
} from "./publicHttpTypes";

type TotpContext =
  | { authMode: "bootstrap"; bootstrapToken: string }
  | { authMode: "session"; bootstrapToken?: never };
type MutationOptions = {
  apiBase?: string | undefined;
  clientTxnId: string;
  signal?: AbortSignal;
};
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

function totpEnrollmentRequestInit(
  options: TotpContext & {
    body: Record<string, unknown>;
    signal?: AbortSignal;
  },
): RequestInit {
  return {
    method: "POST",
    credentials: options.authMode === "bootstrap" ? "omit" : "include",
    body: JSON.stringify(options.body),
    ...(options.signal === undefined ? {} : { signal: options.signal }),
    ...(options.authMode === "bootstrap"
      ? {
          headers: { Authorization: `Bearer ${options.bootstrapToken.trim()}` },
        }
      : {}),
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
  signal?: AbortSignal;
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
      ...(options.signal === undefined ? {} : { signal: options.signal }),
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
  signal?: AbortSignal;
  providerKey: string;
  returnTo?: string | undefined;
}) {
  return fetchHTTPOperation<BeginEnterpriseAuthResponse>({
    apiBase: options.apiBase,
    operationID: "beginEnterpriseAuth",
    pathParameters: { provider_key: options.providerKey },
    init: {
      method: "POST",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
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
    init: {
      method: "POST",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
    },
  });
}

export function beginTotpEnrollment(
  options: MutationOptions &
    (
      | {
          authMode: "bootstrap";
          bootstrapToken: string;
          currentPassword?: never;
          currentFactorCode?: never;
        }
      | {
          authMode: "session";
          bootstrapToken?: never;
          currentPassword: string;
          currentFactorCode?: string;
        }
    ),
) {
  const secondFactor = secondFactorPayload(options.currentFactorCode ?? "");
  const requestInit = totpEnrollmentRequestInit({
    ...options,
    body: {
      client_txn_id: options.clientTxnId,
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

export function completeTotpEnrollment(
  options: MutationOptions &
    TotpContext & { code: string; enrollmentId: string },
) {
  const requestInit = totpEnrollmentRequestInit({
    ...options,
    body: {
      client_txn_id: options.clientTxnId,
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
  clientTxnId: string;
  signal?: AbortSignal;
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
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify({
        client_txn_id: options.clientTxnId,
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
