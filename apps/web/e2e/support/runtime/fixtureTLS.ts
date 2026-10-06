import { readFileSync } from "node:fs";
import { createServer, get } from "node:https";
import path from "node:path";

export async function withTLSProbe(
  kind: "trusted" | "untrusted" | "wrong-name",
  action: (origin: string) => Promise<void>,
): Promise<void> {
  const runtime = process.env.CARTULARY_WEB_E2E_RUNTIME_ROOT;
  if (!runtime || !path.isAbsolute(runtime)) {
    throw new Error("TLS probe requires the owned browser runtime");
  }
  const identity = kind === "trusted" ? "frontend" : kind;
  const server = createServer(
    {
      cert: readFileSync(path.join(runtime, "tls", `${identity}.crt`)),
      key: readFileSync(path.join(runtime, "tls", `${identity}.key`)),
      minVersion: "TLSv1.3",
      maxVersion: "TLSv1.3",
    },
    (_request, response) => response.end("isolated TLS fixture"),
  );
  try {
    await new Promise<void>((resolve, reject) => {
      server.once("error", reject);
      server.listen(0, "127.0.0.1", resolve);
    });
    const address = server.address();
    if (!address || typeof address === "string") {
      throw new Error("TLS probe has no owned TCP listener");
    }
    await action(`https://127.0.0.1:${address.port}`);
  } finally {
    server.closeAllConnections();
    await new Promise<void>((resolve, reject) =>
      server.close((error) => (error ? reject(error) : resolve())),
    );
  }
}

export async function requestTLS12(
  endpoint: "frontend" | "backend",
): Promise<void> {
  const origin =
    endpoint === "frontend"
      ? process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN
      : process.env.CARTULARY_WEB_E2E_API_ORIGIN;
  const root = process.env.CARTULARY_WEB_E2E_TLS_ROOT_CERTIFICATE;
  if (!origin || !root)
    throw new Error("TLS probe requires an admitted fixture");
  await new Promise<void>((resolve, reject) => {
    const request = get(
      `${origin}/healthz`,
      {
        ca: readFileSync(root),
        minVersion: "TLSv1.2",
        maxVersion: "TLSv1.2",
        agent: false,
      },
      (response) => {
        response.resume();
        response.once("end", resolve);
      },
    );
    request.setTimeout(5000, () =>
      request.destroy(new Error("TLS probe timed out")),
    );
    request.once("error", reject);
  });
}
