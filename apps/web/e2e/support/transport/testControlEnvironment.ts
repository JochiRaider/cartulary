import { readFileSync } from "node:fs";

import type { JsonRequestContextLike } from "./publicJsonClient";
import { TestControlClient } from "./testControlClient";

function tokenFromEnvironment(): string {
  const tokenFile = process.env.CARTULARY_TEST_ROUTE_TOKEN_FILE?.trim() ?? "";
  if (tokenFile !== "") {
    try {
      const token = readFileSync(tokenFile, "utf8").trim();
      if (token !== "") {
        return token;
      }
    } catch {
      throw new Error("configured test-control token file is unreadable");
    }
  }
  const token = process.env.CARTULARY_TEST_ROUTE_TOKEN?.trim() ?? "";
  if (token === "") {
    throw new Error("test-control token is not configured");
  }
  return token;
}

export function createEnvironmentTestControlClient(
  request: JsonRequestContextLike,
  options: {
    readonly endpointOrigin: string;
    readonly requestOrigin?: string | undefined;
  },
) {
  const requestOrigin =
    options.requestOrigin ??
    process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN?.trim() ??
    options.endpointOrigin;
  return new TestControlClient({
    approvedEndpointOrigins: [options.endpointOrigin],
    approvedRequestOrigins: [requestOrigin],
    endpointOrigin: options.endpointOrigin,
    request,
    requestOrigin,
    token: tokenFromEnvironment(),
  });
}
