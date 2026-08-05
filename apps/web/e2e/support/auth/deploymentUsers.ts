import type {
  CreateDeploymentUserRequest,
  CreateDeploymentUserResponse,
} from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";

import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import {
  atJsonOrigin,
  type JsonRequestContextLike,
} from "../transport/publicJsonClient";
import { csrfHeaders } from "./browserSession";

export type DeploymentUserCreation = Omit<
  CreateDeploymentUserRequest,
  "auth_kind" | "client_txn_id"
> &
  Required<
    Pick<CreateDeploymentUserRequest, "is_deployment_admin" | "mfa_required">
  >;

export async function createDeploymentUser(
  authority: JsonRequestContextLike | Page,
  options: DeploymentUserCreation,
): Promise<CreateDeploymentUserResponse["data"]> {
  let request: JsonRequestContextLike;
  let headers: Record<string, string> | undefined;
  if ("fetch" in authority) {
    request = authority;
  } else {
    request = atJsonOrigin(authority.request, apiBase);
    headers = await csrfHeaders(authority);
  }
  const body: CreateDeploymentUserRequest = {
    ...options,
    auth_kind: "local",
    client_txn_id: uniqueTxn("deployment-user"),
  };
  const response = await publicHttpOperation({
    body,
    ...(headers === undefined ? {} : { headers }),
    operationID: "createDeploymentUser",
    request,
  });
  if (!response.ok) {
    throw new Error(
      `create deployment user failed with HTTP ${response.status}`,
    );
  }
  return response.payload.data;
}
