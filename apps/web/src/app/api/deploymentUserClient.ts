import type {
  CreateDeploymentUserRequest,
  CreateDeploymentUserResponse,
  CreateEnterpriseAuthBindingRequest,
  CreateEnterpriseAuthBindingResponse,
  GetDeploymentUserResponse,
  ListDeploymentUsersResponse,
  PatchDeploymentUserRequest,
  PatchDeploymentUserResponse,
  ResetDeploymentUserPasswordRequest,
  ResetDeploymentUserPasswordResponse,
  ResetDeploymentUserTOTPRequest,
  ResetDeploymentUserTOTPResponse,
  RetireEnterpriseAuthBindingRequest,
  RetireEnterpriseAuthBindingResponse,
  RevokeAllDeploymentUserSessionsRequest,
  RevokeAllDeploymentUserSessionsResponse,
  RotateEnterpriseAuthBindingRequest,
  RotateEnterpriseAuthBindingResponse,
} from "@cartulary/protocol-ts/http";
import { clientTxnID, fetchHTTPOperation } from "../../services/browserApi";

export function createLocalUser(options: {
  apiBase?: string | undefined;
  clientTxnId?: string;
  displayName: string;
  email: string;
  initialPassword: string;
  isDeploymentAdmin: boolean;
  mfaRequired: boolean;
}) {
  const request = {
    client_txn_id:
      options.clientTxnId ?? clientTxnID("authentication-ui-user-create"),
    auth_kind: "local",
    email: options.email,
    display_name: options.displayName,
    initial_password: options.initialPassword,
    mfa_required: options.mfaRequired,
    is_deployment_admin: options.isDeploymentAdmin,
  } satisfies CreateDeploymentUserRequest;
  return fetchHTTPOperation<CreateDeploymentUserResponse>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      body: JSON.stringify(request),
    },
    operationID: "createDeploymentUser",
  });
}

export function listUsers(options?: {
  apiBase?: string | undefined;
  cursorToken?: string | null;
  isActive?: boolean | null | undefined;
  isDeploymentAdmin?: boolean | null | undefined;
  limit?: number | undefined;
  search?: string | undefined;
  signal?: AbortSignal | undefined;
}) {
  const query: Record<string, string | number> = {
    limit: options?.limit ?? 100,
  };
  const cursorToken = options?.cursorToken?.trim() ?? "";
  if (cursorToken !== "") {
    query.cursor_token = cursorToken;
  }
  const search = options?.search?.trim() ?? "";
  if (search !== "") {
    query.search = search;
  }
  if (typeof options?.isActive === "boolean") {
    query.is_active = String(options.isActive);
  }
  if (typeof options?.isDeploymentAdmin === "boolean") {
    query.is_deployment_admin = String(options.isDeploymentAdmin);
  }
  const requestInit =
    typeof options?.signal === "undefined"
      ? undefined
      : {
          signal: options.signal,
        };
  return fetchHTTPOperation<ListDeploymentUsersResponse>({
    apiBase: options?.apiBase,
    init: requestInit,
    operationID: "listDeploymentUsers",
    query,
  });
}

export function loadUser(options: {
  apiBase?: string | undefined;
  userId: string;
}) {
  return fetchHTTPOperation<GetDeploymentUserResponse>({
    apiBase: options.apiBase,
    operationID: "getDeploymentUser",
    pathParameters: { user_id: options.userId },
  });
}

export function patchLocalUser(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  displayName: string;
  email: string;
  isActive: boolean;
  isDeploymentAdmin: boolean;
  mfaRequired: boolean;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    email: options.email,
    display_name: options.displayName,
    mfa_required: options.mfaRequired,
    is_active: options.isActive,
    is_deployment_admin: options.isDeploymentAdmin,
  } satisfies PatchDeploymentUserRequest;
  return fetchHTTPOperation<PatchDeploymentUserResponse>({
    apiBase: options.apiBase,
    init: {
      method: "PATCH",
      body: JSON.stringify(request),
    },
    operationID: "patchDeploymentUser",
    pathParameters: { user_id: options.userId },
  });
}

export function createEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  providerKey: string;
  providerSubject: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id:
      options.clientTxnId ??
      clientTxnID("authentication-ui-auth-binding-create"),
    provider_key: options.providerKey,
    provider_subject: options.providerSubject,
    reason: options.reason,
  } satisfies CreateEnterpriseAuthBindingRequest;
  return fetchHTTPOperation<CreateEnterpriseAuthBindingResponse>({
    apiBase: options.apiBase,
    operationID: "createEnterpriseAuthBinding",
    pathParameters: { user_id: options.userId },
    init: { method: "POST", body: JSON.stringify(request) },
  });
}

export function rotateEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  authBindingId: string;
  baseUserVersion: number;
  clientTxnId?: string;
  newProviderSubject: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id:
      options.clientTxnId ??
      clientTxnID("authentication-ui-auth-binding-rotate"),
    new_provider_subject: options.newProviderSubject,
    reason: options.reason,
  } satisfies RotateEnterpriseAuthBindingRequest;
  return fetchHTTPOperation<RotateEnterpriseAuthBindingResponse>({
    apiBase: options.apiBase,
    operationID: "rotateEnterpriseAuthBinding",
    pathParameters: {
      user_id: options.userId,
      auth_binding_id: options.authBindingId,
    },
    init: { method: "POST", body: JSON.stringify(request) },
  });
}

export function retireEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  authBindingId: string;
  baseUserVersion: number;
  clientTxnId?: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id:
      options.clientTxnId ??
      clientTxnID("authentication-ui-auth-binding-retire"),
    reason: options.reason,
  } satisfies RetireEnterpriseAuthBindingRequest;
  return fetchHTTPOperation<RetireEnterpriseAuthBindingResponse>({
    apiBase: options.apiBase,
    operationID: "retireEnterpriseAuthBinding",
    pathParameters: {
      user_id: options.userId,
      auth_binding_id: options.authBindingId,
    },
    init: { method: "DELETE", body: JSON.stringify(request) },
  });
}

export function adminResetPassword(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  newPassword: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id:
      options.clientTxnId ?? clientTxnID("authentication-ui-password-reset"),
    new_password: options.newPassword,
    reason: options.reason,
  } satisfies ResetDeploymentUserPasswordRequest;
  return fetchHTTPOperation<ResetDeploymentUserPasswordResponse>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      body: JSON.stringify(request),
    },
    operationID: "resetDeploymentUserPassword",
    pathParameters: { user_id: options.userId },
  });
}

export function adminResetTotp(options: {
  apiBase?: string | undefined;
  baseUserVersion: number;
  clientTxnId?: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id:
      options.clientTxnId ?? clientTxnID("authentication-ui-totp-reset"),
    reason: options.reason,
  } satisfies ResetDeploymentUserTOTPRequest;
  return fetchHTTPOperation<ResetDeploymentUserTOTPResponse>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      body: JSON.stringify(request),
    },
    operationID: "resetDeploymentUserTOTP",
    pathParameters: { user_id: options.userId },
  });
}

export function adminRevokeAllSessions(options: {
  apiBase?: string | undefined;
  clientTxnId?: string;
  reason: string;
  userId: string;
}) {
  const request = {
    client_txn_id:
      options.clientTxnId ?? clientTxnID("authentication-ui-revoke-all"),
    reason: options.reason,
  } satisfies RevokeAllDeploymentUserSessionsRequest;
  return fetchHTTPOperation<RevokeAllDeploymentUserSessionsResponse>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      body: JSON.stringify(request),
    },
    operationID: "revokeAllDeploymentUserSessions",
    pathParameters: { user_id: options.userId },
  });
}
