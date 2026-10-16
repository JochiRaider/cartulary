import type { BrowserContext } from "playwright-chromium";

import { RenderFailure } from "./failure.ts";

export interface BrowserEndpoint {
  port: number;
  portRef: string;
}

export type DestinationClass =
  | "loopback"
  | "http"
  | "https"
  | "file"
  | "data"
  | "javascript"
  | "tcp";

export interface BrowserObservation {
  schema_id: "render_sandbox_observation.v1";
  network_policy_id: "render_sandbox_policy_v1";
  stage: "slidev_export" | "mermaid_render";
  observation_kind:
    | "loopback_allowed"
    | "loopback_blocked"
    | "non_loopback_blocked"
    | "url_scheme_blocked";
  destination_class: DestinationClass;
  allowed: boolean;
  attempt_count: number;
  loopback_port_ref: string | null;
}

export function browserDestination(
  rawURL: string,
  endpoint: BrowserEndpoint,
): Pick<
  BrowserObservation,
  "observation_kind" | "destination_class" | "allowed" | "loopback_port_ref"
> {
  let url: URL;
  try {
    url = new URL(rawURL);
  } catch {
    return {
      observation_kind: "url_scheme_blocked",
      destination_class: "tcp",
      allowed: false,
      loopback_port_ref: null,
    };
  }
  const web = ["http:", "https:", "ws:", "wss:"].includes(url.protocol);
  const loopback = ["127.0.0.1", "[::1]", "localhost"].includes(url.hostname);
  const allowed =
    ["http:", "ws:"].includes(url.protocol) &&
    loopback &&
    !url.username &&
    !url.password &&
    Number(url.port || (url.protocol === "http:" ? 80 : 443)) === endpoint.port;
  return {
    observation_kind: allowed
      ? "loopback_allowed"
      : !web
        ? "url_scheme_blocked"
        : loopback
          ? "loopback_blocked"
          : "non_loopback_blocked",
    destination_class: loopback
      ? "loopback"
      : url.protocol === "http:" || url.protocol === "https:"
        ? (url.protocol.slice(0, -1) as "http" | "https")
        : url.protocol === "file:" ||
            url.protocol === "data:" ||
            url.protocol === "javascript:"
          ? (url.protocol.slice(0, -1) as "file" | "data" | "javascript")
          : "tcp",
    allowed,
    loopback_port_ref: allowed ? endpoint.portRef : null,
  };
}

// Browser interception is defense in depth. It does not observe Node/native
// traffic and must never be presented as the operating-system sandbox evidence.
export async function installBrowserPolicy(
  context: BrowserContext,
  endpoint: BrowserEndpoint,
  stage: BrowserObservation["stage"],
) {
  if (
    !Number.isInteger(endpoint.port) ||
    endpoint.port < 1 ||
    endpoint.port > 65535 ||
    !/^[a-z][a-z0-9_.-]{0,127}$/.test(endpoint.portRef)
  ) {
    throw new RenderFailure(
      "security_sandbox_violation",
      "invalid_loopback_binding",
    );
  }
  const observations = new Map<string, BrowserObservation>();
  let failure: RenderFailure | undefined;
  function observe(url: string) {
    const destination = browserDestination(url, endpoint);
    const key = `${destination.observation_kind}:${destination.destination_class}`;
    const prior = observations.get(key);
    if (prior) prior.attempt_count += 1;
    else
      observations.set(key, {
        schema_id: "render_sandbox_observation.v1",
        network_policy_id: "render_sandbox_policy_v1",
        stage,
        ...destination,
        attempt_count: 1,
      });
    if (!destination.allowed && !failure) {
      failure =
        destination.observation_kind === "non_loopback_blocked"
          ? new RenderFailure(
              "remote_asset_reference",
              "outbound_request_observed",
            )
          : new RenderFailure(
              "security_sandbox_violation",
              destination.observation_kind === "loopback_blocked"
                ? "loopback_not_declared"
                : "external_reference_in_rendered_output",
            );
    }
    return destination.allowed;
  }
  await context.route("**/*", async (route) => {
    if (observe(route.request().url())) await route.continue();
    else await route.abort("blockedbyclient");
  });
  await context.routeWebSocket("**/*", (route) => {
    if (observe(route.url())) route.connectToServer();
    else route.close();
  });
  return {
    assertNoViolation() {
      if (failure) throw failure;
    },
    assertAllowed() {
      if (failure) throw failure;
      if (!observations.size)
        throw new RenderFailure(
          "security_sandbox_violation",
          "sandbox_observation_missing",
        );
    },
    observations() {
      return [...observations.entries()]
        .sort(([a], [b]) => a.localeCompare(b, "en"))
        .map(([, value]) => ({ ...value }));
    },
  };
}
