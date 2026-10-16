import { createHash } from "node:crypto";
import { writeFile } from "node:fs/promises";
import path from "node:path";

import { chromium } from "playwright-chromium";

import browserProfile from "./browser-profile.json" with { type: "json" };
import { RenderFailure } from "./failure.ts";
import { renderMermaidSVG } from "./mermaid.ts";
import { renderSlidevPDF } from "./slidev.ts";

export const stageRequestSchema = "cartulary.reporting_render_stage_request.v5";
export const stageResultSchema = "cartulary.reporting_render_stage_result.v5";

import {
  type StageAsset,
  stageAssetsDigest,
  validateStageAssets,
} from "./stage-assets.ts";

export type { StageAsset } from "./stage-assets.ts";
export { stageAssetsDigest } from "./stage-assets.ts";
export const stageRequestLimit = 768 * 1024 * 1024;

interface StageIdentity {
  schema_id: typeof stageRequestSchema;
  source: string;
  admitted_at: string;
  identity_sha256: string;
}
export type StageRequest = StageIdentity &
  (
    | { operation: "slidev_pdf"; expected_pages: number; assets?: StageAsset[] }
    | { operation: "mermaid_svg" }
  );

// This private protocol contains data only. Paths, executable selection, launch
// arguments, environment and network policy are installed-runtime capabilities.
export function decodeStageRequest(bytes: Uint8Array): StageRequest {
  if (bytes.byteLength === 0 || bytes.byteLength > stageRequestLimit)
    throw new RenderFailure("slidev_source_invalid", "stage_request_invalid");
  const decoder = new TextDecoder("utf-8", { fatal: true, ignoreBOM: true });
  let value: unknown;
  let text = "";
  try {
    text = decoder.decode(bytes);
    value = JSON.parse(text);
  } catch {
    throw new RenderFailure("slidev_source_invalid", "stage_request_invalid");
  }
  if (!value || typeof value !== "object" || Array.isArray(value))
    throw new RenderFailure("slidev_source_invalid", "stage_request_invalid");
  const input = value as Record<string, unknown>;
  const keys = [
    "schema_id",
    "operation",
    "source",
    "admitted_at",
    "identity_sha256",
  ];
  if (input.operation === "slidev_pdf") keys.push("expected_pages");
  if (input.operation === "slidev_pdf" && Object.hasOwn(input, "assets"))
    keys.push("assets");
  if (
    Object.keys(input).length !== keys.length ||
    keys.some((key) => !Object.hasOwn(input, key)) ||
    input.schema_id !== stageRequestSchema ||
    !["slidev_pdf", "mermaid_svg"].includes(String(input.operation)) ||
    typeof input.source !== "string" ||
    input.source.length === 0 ||
    Buffer.byteLength(input.source) > 64 * 1024 * 1024 ||
    Buffer.from(input.source).toString("utf8") !== input.source ||
    typeof input.admitted_at !== "string" ||
    !/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,6})?Z$/.test(
      input.admitted_at,
    ) ||
    !Number.isFinite(Date.parse(input.admitted_at)) ||
    new Date(input.admitted_at).toISOString().slice(0, 19) !==
      input.admitted_at.slice(0, 19) ||
    typeof input.identity_sha256 !== "string" ||
    !/^[a-f0-9]{64}$/.test(input.identity_sha256) ||
    (input.operation === "slidev_pdf" &&
      (typeof input.expected_pages !== "number" ||
        !Number.isSafeInteger(input.expected_pages) ||
        input.expected_pages < 1 ||
        input.expected_pages > 9840))
  )
    throw new RenderFailure("slidev_source_invalid", "stage_request_invalid");
  if (Object.hasOwn(input, "assets")) validateStageAssets(input.assets);
  // JSON.parse silently overwrites duplicate keys. Track each object separately,
  // including nested assets; the full shape is validated before this scan.
  const token =
    /\s*("(?:[^"\\]|\\(?:["\\/bfnrt]|u[0-9a-fA-F]{4}))*"|-?(?:0|[1-9][0-9]*)(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?|[{}[\]:,])/y;
  const objects: Set<string>[] = [];
  let previous = "";
  let offset = 0;
  const end = text.trimEnd().length;
  while (offset < end) {
    token.lastIndex = offset;
    const match = token.exec(text);
    if (!match)
      throw new RenderFailure("slidev_source_invalid", "stage_request_invalid");
    const current = match[1] ?? "";
    if (current === "{") objects.push(new Set());
    if (current === "}") objects.pop();
    if (current === ":") {
      const key = JSON.parse(previous) as string;
      const observedKeys = objects.at(-1);
      if (!observedKeys || observedKeys.has(key))
        throw new RenderFailure(
          "slidev_source_invalid",
          "stage_request_invalid",
        );
      observedKeys.add(key);
    }
    previous = current;
    offset = token.lastIndex;
  }
  return input as unknown as StageRequest;
}

// The supervisor allocates this port in its private network namespace. Source
// input cannot request another endpoint, and the stage never probes for one.
export function decodeStagePort(value: string | undefined): number {
  if (value === undefined || !/^[1-9][0-9]{3,4}$/.test(value))
    throw new RenderFailure(
      "security_sandbox_violation",
      "invalid_loopback_binding",
    );
  const port = Number(value);
  if (port < 1024 || port > 65535)
    throw new RenderFailure(
      "security_sandbox_violation",
      "invalid_loopback_binding",
    );
  return port;
}

// Called only after the outer supervisor installed the private namespaces and
// runtime mounts. Browser observations below do not claim native-attempt audit.
export async function executeStage(
  input: StageRequest,
  workDirectory: string,
  allocatedPort: number,
) {
  const port = decodeStagePort(String(allocatedPort));
  const browser = await chromium
    .launch({
      chromiumSandbox: false,
      // No driver-injected defaults or random profile paths enter the actual
      // browser argv. Production workDirectory is always /work; direct adapter
      // tests use their explicit disposable directory outside qualification.
      ignoreDefaultArgs: true,
      args: browserProfile.browser_launch_args.map((argument) =>
        argument === "--user-data-dir=/work/browser-profile"
          ? `--user-data-dir=${path.join(workDirectory, "browser-profile")}`
          : argument,
      ),
      ...(process.env.CARTULARY_RENDER_CHROMIUM
        ? { executablePath: process.env.CARTULARY_RENDER_CHROMIUM }
        : {}),
    })
    .catch((error: unknown) => {
      const message = error instanceof Error ? error.message : "";
      const reason = message.includes("error while loading shared libraries")
        ? "stage_browser_library_missing"
        : message.includes("chrome_crashpad_handler")
          ? "stage_browser_crashpad_failed"
          : message.includes("Operation not permitted")
            ? "stage_browser_operation_denied"
            : "stage_browser_launch_failed";
      throw new RenderFailure("security_sandbox_violation", reason);
    });
  try {
    const runtime = {
      workDirectory,
      endpoint: { port, portRef: "render-loopback-1" },
      browser,
    };
    const rendered =
      input.operation === "slidev_pdf"
        ? await renderSlidevPDF(
            {
              source: input.source,
              admittedAt: input.admitted_at,
              identitySHA256: input.identity_sha256,
              expectedPages: input.expected_pages,
              assets: input.assets ?? [],
            },
            runtime,
          )
        : await renderMermaidSVG(input.source, runtime);
    const artifact = {
      path: input.operation === "slidev_pdf" ? "deck.pdf" : "diagram.svg",
      media_type:
        input.operation === "slidev_pdf" ? "application/pdf" : "image/svg+xml",
      bytes: rendered.bytes.byteLength,
      sha256: createHash("sha256").update(rendered.bytes).digest("hex"),
    };
    if (rendered.bytes.byteLength > 500 * 1024 * 1024)
      throw new RenderFailure(
        "bundle_resource_limit_exceeded",
        "rendered_pdf_bytes_exceeded",
      );
    await writeFile(path.join(workDirectory, artifact.path), rendered.bytes, {
      flag: "wx",
      mode: 0o600,
    });
    return {
      schema_id: stageResultSchema,
      operation: input.operation,
      identity_sha256: input.identity_sha256,
      source_sha256: createHash("sha256")
        .update(input.source, "utf8")
        .digest("hex"),
      assets_sha256: stageAssetsDigest(
        input.operation === "slidev_pdf" ? (input.assets ?? []) : [],
      ),
      artifact,
      browser_observations: rendered.browserObservations,
    };
  } finally {
    await browser.close();
  }
}
