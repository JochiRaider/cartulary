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
import { fetchHTTPOperation } from "../../services/browserApi";

export function createLocalUser(options: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  clientTxnId: string;
  displayName: string;
  email: string;
  initialPassword: string;
  isDeploymentAdmin: boolean;
  mfaRequired: boolean;
}) {
  const request = {
    client_txn_id: options.clientTxnId,
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
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
    operationID: "createDeploymentUser",
  });
}

export function listUsers(options?: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  cursorToken?: string | null;
  isActive?: boolean | null | undefined;
  isDeploymentAdmin?: boolean | null | undefined;
  limit?: number | undefined;
  search?: string | undefined;
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
  signal?: AbortSignal;
  userId: string;
}) {
  return fetchHTTPOperation<GetDeploymentUserResponse>({
    apiBase: options.apiBase,
    operationID: "getDeploymentUser",
    init: options.signal === undefined ? undefined : { signal: options.signal },
    pathParameters: { user_id: options.userId },
  });
}

export type DeploymentUserChanges = Partial<
  Pick<
    PatchDeploymentUserRequest,
    | "email"
    | "display_name"
    | "mfa_required"
    | "is_active"
    | "is_deployment_admin"
  >
>;
export function patchLocalUser(options: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  baseUserVersion: number;
  changes: DeploymentUserChanges;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    ...options.changes,
  } satisfies PatchDeploymentUserRequest;
  return fetchHTTPOperation<PatchDeploymentUserResponse>({
    apiBase: options.apiBase,
    init: {
      method: "PATCH",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
    operationID: "patchDeploymentUser",
    pathParameters: { user_id: options.userId },
  });
}

export function createEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  baseUserVersion: number;
  clientTxnId: string;
  providerKey: string;
  providerSubject: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id: options.clientTxnId,
    provider_key: options.providerKey,
    provider_subject: options.providerSubject,
    reason: options.reason,
  } satisfies CreateEnterpriseAuthBindingRequest;
  return fetchHTTPOperation<CreateEnterpriseAuthBindingResponse>({
    apiBase: options.apiBase,
    operationID: "createEnterpriseAuthBinding",
    pathParameters: { user_id: options.userId },
    init: {
      method: "POST",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
  });
}

export function rotateEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  authBindingId: string;
  baseUserVersion: number;
  clientTxnId: string;
  newProviderSubject: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id: options.clientTxnId,
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
    init: {
      method: "POST",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
  });
}

export function retireEnterpriseAuthBinding(options: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  authBindingId: string;
  baseUserVersion: number;
  clientTxnId: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id: options.clientTxnId,
    reason: options.reason,
  } satisfies RetireEnterpriseAuthBindingRequest;
  return fetchHTTPOperation<RetireEnterpriseAuthBindingResponse>({
    apiBase: options.apiBase,
    operationID: "retireEnterpriseAuthBinding",
    pathParameters: {
      user_id: options.userId,
      auth_binding_id: options.authBindingId,
    },
    init: {
      method: "DELETE",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
  });
}

export function adminResetPassword(options: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  baseUserVersion: number;
  clientTxnId: string;
  newPassword: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id: options.clientTxnId,
    new_password: options.newPassword,
    reason: options.reason,
  } satisfies ResetDeploymentUserPasswordRequest;
  return fetchHTTPOperation<ResetDeploymentUserPasswordResponse>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
    operationID: "resetDeploymentUserPassword",
    pathParameters: { user_id: options.userId },
  });
}

export function adminResetTotp(options: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  baseUserVersion: number;
  clientTxnId: string;
  reason: string;
  userId: string;
}) {
  const request = {
    base_user_version: options.baseUserVersion,
    client_txn_id: options.clientTxnId,
    reason: options.reason,
  } satisfies ResetDeploymentUserTOTPRequest;
  return fetchHTTPOperation<ResetDeploymentUserTOTPResponse>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
    operationID: "resetDeploymentUserTOTP",
    pathParameters: { user_id: options.userId },
  });
}

export function adminRevokeAllSessions(options: {
  apiBase?: string | undefined;
  signal?: AbortSignal;
  clientTxnId: string;
  reason: string;
  userId: string;
}) {
  const request = {
    client_txn_id: options.clientTxnId,
    reason: options.reason,
  } satisfies RevokeAllDeploymentUserSessionsRequest;
  return fetchHTTPOperation<RevokeAllDeploymentUserSessionsResponse>({
    apiBase: options.apiBase,
    init: {
      method: "POST",
      ...(options.signal === undefined ? {} : { signal: options.signal }),
      body: JSON.stringify(request),
    },
    operationID: "revokeAllDeploymentUserSessions",
    pathParameters: { user_id: options.userId },
  });
}
