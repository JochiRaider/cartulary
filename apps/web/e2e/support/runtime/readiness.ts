import type { APIRequestContext, Page } from "@playwright/test";

import { createEnvironmentTestControlClient } from "../transport/testControlEnvironment";
import { apiBase } from "./configuration";

export async function waitForAPIReady(authRequests: APIRequestContext) {
  for (let attempt = 0; attempt < 60; attempt += 1) {
    try {
      const response = await authRequests.get("/readyz");
      if (response.ok()) {
        return;
      }
    } catch (error) {
      if (!isConnectionRefused(error)) {
        throw error;
      }
    }
    await sleep(500);
  }
  throw new Error(`timed out waiting for API readiness at ${apiBase}/readyz`);
}

export async function waitForPageRequestAPIReady(page: Page) {
  for (let attempt = 0; attempt < 60; attempt += 1) {
    try {
      const response = await page.request.get(`${apiBase}/readyz`);
      if (response.ok()) {
        return;
      }
    } catch (error) {
      if (!isConnectionRefused(error)) {
        throw error;
      }
    }
    await sleep(500);
  }
  throw new Error(`timed out waiting for API readiness at ${apiBase}/readyz`);
}

export async function verifyOwnedHarnessRuntime() {
  if (
    !process.env.CARTULARY_TEST_ROUTE_TOKEN_FILE?.trim() &&
    !process.env.CARTULARY_TEST_ROUTE_TOKEN?.trim()
  ) {
    return;
  }
  const request = {
    fetch: async (
      url: string,
      options: {
        data?: unknown;
        headers?: Record<string, string>;
        method: "DELETE" | "GET" | "PATCH" | "POST" | "PUT";
      },
    ) => {
      const response = await fetch(url, {
        ...(options.data === undefined
          ? {}
          : { body: JSON.stringify(options.data) }),
        ...(options.headers === undefined ? {} : { headers: options.headers }),
        method: options.method,
        signal: AbortSignal.timeout(5000),
      });
      return {
        headers: () => Object.fromEntries(response.headers.entries()),
        json: () => response.json(),
        ok: () => response.ok,
        status: () => response.status,
      };
    },
  };
  const response = await createEnvironmentTestControlClient(request, {
    endpointOrigin: apiBase,
  }).request({
    method: "GET",
    path: "/api/v1/test/runtime/identity",
  });
  if (!response.ok) {
    throw new Error(
      `browser harness identity check failed for ${apiBase}: HTTP ${response.status}`,
    );
  }
  const body = response.body as {
    data?: {
      schema_id?: unknown;
      runtime_marker?: unknown;
      server_pid?: unknown;
      test_routes_enabled?: unknown;
    };
  };
  const data = body.data;
  if (
    data?.schema_id !== "cartulary.test.runtime_identity.v1" ||
    data.runtime_marker !== "harness-owned" ||
    data.test_routes_enabled !== true ||
    !Number.isInteger(data.server_pid)
  ) {
    throw new Error(
      `browser harness identity check received an unexpected payload from ${apiBase}`,
    );
  }
}

function isConnectionRefused(error: unknown) {
  return error instanceof Error && error.message.includes("ECONNREFUSED");
}

function sleep(milliseconds: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, milliseconds);
  });
}
