import { createServer } from "node:http";

import { type Browser, chromium } from "playwright-chromium";
import { expect, it } from "vitest";

import { installBrowserPolicy } from "./browser-policy.ts";

it("browser interception blocks external HTTP and undeclared loopback before connection", async () => {
  let served = 0;
  const server = createServer((_request, response) => {
    served += 1;
    response.writeHead(200, { "content-type": "text/html" });
    response.end(
      "<!doctype html><title>Synthetic renderer policy test</title>",
    );
  });
  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", resolve);
  });
  const address = server.address();
  if (!address || typeof address === "string")
    throw new Error("missing test port");
  let browser: Browser | undefined;
  try {
    browser = await chromium.launch();
    const blocked = [
      {
        url: "http://192.0.2.1:31973/private?token=synthetic-secret",
        reason: "outbound_request_observed",
        kind: "non_loopback_blocked",
      },
      {
        url: `http://127.0.0.1:${address.port === 65535 ? 65534 : address.port + 1}/private?token=synthetic-secret`,
        reason: "loopback_not_declared",
        kind: "loopback_blocked",
      },
    ];
    for (const item of blocked) {
      const context = await browser.newContext({ serviceWorkers: "block" });
      try {
        const policy = await installBrowserPolicy(
          context,
          { port: address.port, portRef: "renderer.http" },
          "slidev_export",
        );
        const page = await context.newPage();
        await page.goto(`http://127.0.0.1:${address.port}/`);
        policy.assertAllowed();
        await page.evaluate(async (url) => {
          await fetch(url).catch(() => undefined);
        }, item.url);
        expect(() => policy.assertAllowed()).toThrow(item.reason);
        expect(policy.observations()).toContainEqual({
          schema_id: "render_sandbox_observation.v1",
          network_policy_id: "render_sandbox_policy_v1",
          stage: "slidev_export",
          observation_kind: item.kind,
          destination_class:
            item.kind === "loopback_blocked" ? "loopback" : "http",
          allowed: false,
          attempt_count: 1,
          loopback_port_ref: null,
        });
        expect(JSON.stringify(policy.observations())).not.toMatch(
          /private|secret|192\.0\.2/,
        );
      } finally {
        await context.close();
      }
    }
    expect(served).toBeGreaterThanOrEqual(2);
  } finally {
    await browser?.close();
    await new Promise<void>((resolve, reject) => {
      server.close((error) => (error ? reject(error) : resolve()));
    });
  }
});
