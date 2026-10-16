import { createHash } from "node:crypto";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { createServer } from "node:net";
import { tmpdir } from "node:os";
import path from "node:path";

import { PDFDocument } from "pdf-lib";
import { expect, it } from "vitest";

import {
  decodeStagePort,
  decodeStageRequest,
  executeStage,
  stageAssetsDigest,
  stageRequestSchema,
} from "./stage.ts";
import { stageAssetURL } from "./stage-assets.ts";
import { themeStyleHTML } from "./theme-style.ts";

const request = {
  schema_id: stageRequestSchema,
  operation: "slidev_pdf",
  source: `---
title: Stage protocol test
author: ""
theme: default
download: false
export:
  format: pdf
  withClicks: true
clickAnimation: false
lineNumbers: false
monaco: false
twoslash: false
fonts:
  provider: none
  sans: sans-serif
  local: [sans-serif]
---

# An admitted stage

Synthetic, inert presentation text.
`,
  admitted_at: "2026-04-05T12:00:00Z",
  identity_sha256: "e".repeat(64),
  expected_pages: 1,
};

it("render stage admission rejects unknown capabilities malformed identities and duplicate keys", () => {
  for (const port of [
    undefined,
    "0",
    "80",
    "1023",
    "65536",
    "032123",
    "32123x",
    "3e4",
  ])
    expect(() => decodeStagePort(port)).toThrow("invalid_loopback_binding");
  expect(decodeStagePort("32123")).toBe(32123);
  const encode = (value: unknown) => Buffer.from(JSON.stringify(value));
  expect(decodeStageRequest(encode(request))).toEqual(request);
  const mermaid = {
    schema_id: stageRequestSchema,
    operation: "mermaid_svg",
    source: 'flowchart TD\n  n0001["A"]\n',
    admitted_at: request.admitted_at,
    identity_sha256: request.identity_sha256,
  };
  expect(decodeStageRequest(encode(mermaid))).toEqual(mermaid);
  const content =
    '<svg xmlns="http://www.w3.org/2000/svg"><text>é</text></svg>';
  const asset = {
    path: "diagrams/a.svg",
    media_type: "image/svg+xml",
    bytes: Buffer.from(content).toString("base64"),
    sha256: createHash("sha256").update(content).digest("hex"),
  };
  const withAssets = {
    ...request,
    assets: [asset, { ...asset, path: "diagrams/b.svg" }],
  };
  expect(decodeStageRequest(encode(withAssets))).toEqual(withAssets);
  const binary = Buffer.from([0, 255, 128, 1]);
  const local = {
    path: "assets/image/café.png",
    media_type: "image/png",
    bytes: binary.toString("base64"),
    sha256: createHash("sha256").update(binary).digest("hex"),
  };
  expect(
    decodeStageRequest(encode({ ...request, assets: [local, asset] })),
  ).toEqual({ ...request, assets: [local, asset] });
  expect(stageAssetURL("assets/image/café !'()*+$&=.png")).toBe(
    "assets/image/caf%C3%A9%20%21%27%28%29%2A%2B%24%26%3D.png",
  );
  const css = Buffer.from(
    ':root { --cartulary-report-text: #183650; --cartulary-report-logo: url("images/brand.svg"); }',
  );
  const theme = {
    path: "theme/style.css",
    media_type: "text/css; charset=utf-8",
    bytes: css.toString("base64"),
    sha256: createHash("sha256").update(css).digest("hex"),
  };
  const logo = { ...asset, path: "theme/images/brand.svg" };
  expect(
    decodeStageRequest(encode({ ...request, assets: [logo, theme] })),
  ).toEqual({ ...request, assets: [logo, theme] });
  const html = themeStyleHTML([logo, theme]);
  expect(html).toContain('href="/theme/style.css"');
  expect(html).toContain("color:var(--cartulary-report-text) !important");
  expect(html).toContain('url("/theme/images/brand.svg")');
  expect(html).not.toContain("font-family:");
  expect(html).not.toContain(css.toString());
  expect(themeStyleHTML([])).toBe("");
  expect(stageAssetsDigest([asset])).toBe(
    createHash("sha256")
      .update(
        `cartulary.reporting_render_assets.v2\x0014:diagrams/a.svg13:image/svg+xml61:${content}`,
      )
      .digest("hex"),
  );
  for (const assets of [
    null,
    {},
    [asset, asset],
    [{ ...asset, path: "../a.svg" }],
    [{ ...asset, path: "diagrams/../a.svg" }],
    [{ ...asset, path: "/diagrams/a.svg" }],
    [{ ...asset, path: "diagrams/a.svg?x" }],
    [{ ...asset, sha256: "f".repeat(64) }],
    [{ ...asset, bytes: "" }],
    [{ ...asset, bytes: `${asset.bytes}\n` }],
    [{ ...asset, content, bytes: undefined }],
    [{ ...asset, media_type: "text/html" }],
    [{ ...local, path: "assets/image/cafe\u0301.png" }],
    [{ ...local, path: "assets/image/%2e%2e.png" }],
    [{ ...local, path: "assets/image/a.js" }],
    [{ ...local, path: "assets/image/a.png?x" }],
    [{ ...local, path: "assets/image/a.png#x" }],
    [{ ...asset, command: "forbidden" }],
    [{ ...theme, path: "theme/../style.css" }],
    [{ ...theme, path: "theme/setup.js" }],
    [{ ...theme, path: "theme/style.css?x" }],
    [{ ...theme, path: "theme/fonts/a.ttf", media_type: "font/ttf" }],
    [{ ...theme, path: "theme/fonts/%61.woff2", media_type: "font/woff2" }],
    [{ ...theme, media_type: "application/javascript" }],
    [logo],
    [{ ...logo, path: "theme/images/a.svg" }, logo, theme],
  ]) {
    expect(() => decodeStageRequest(encode({ ...request, assets }))).toThrow(
      "stage_request_invalid",
    );
  }
  expect(() =>
    decodeStageRequest(encode({ ...mermaid, assets: [asset] })),
  ).toThrow("stage_request_invalid");
  const duplicateAsset = JSON.stringify({
    ...request,
    assets: [asset],
  }).replace(
    '"path":"diagrams/a.svg"',
    '"path":"diagrams/a.svg","\\u0070ath":"diagrams/a.svg"',
  );
  expect(() => decodeStageRequest(Buffer.from(duplicateAsset))).toThrow(
    "stage_request_invalid",
  );
  for (const invalid of [
    null,
    [],
    {},
    { ...request, source: "" },
    { ...request, source: null },
    { ...request, operation: "shell" },
    { ...request, schema_id: "cartulary.reporting_render_stage_request.v3" },
    { ...request, schema_id: "cartulary.reporting_render_stage_request.v4" },
    { ...request, schema_id: "retired" },
    { ...request, admitted_at: "2026-02-31T12:00:00Z" },
    { ...request, identity_sha256: "not-a-digest" },
    { ...request, expected_pages: 0 },
    { ...request, expected_pages: 1.5 },
    { ...request, expected_pages: 9841 },
    { ...request, command: ["/bin/sh"] },
    { ...request, environment: { SECRET: "forbidden" } },
    { ...request, work_directory: "/outside" },
    { ...request, loopback_port: 8080 },
    { ...request, schema_id: "cartulary.reporting_render_stage_request.v1" },
    { ...request, schema_id: "cartulary.reporting_render_stage_request.v2" },
    { ...mermaid, expected_pages: 1 },
  ])
    expect(() => decodeStageRequest(encode(invalid))).toThrow(
      "stage_request_invalid",
    );
  for (const key of Object.keys(request)) {
    const invalid = { ...request } as Record<string, unknown>;
    delete invalid[key];
    expect(() => decodeStageRequest(encode(invalid))).toThrow(
      "stage_request_invalid",
    );
  }
  for (const prefix of ['"expected_pages":2,', '"\\u0065xpected_pages":2,']) {
    const duplicate = `{${prefix}${JSON.stringify(request).slice(1)}`;
    expect(() => decodeStageRequest(Buffer.from(duplicate))).toThrow(
      "stage_request_invalid",
    );
  }
  expect(() => decodeStageRequest(new Uint8Array([0xff]))).toThrow(
    "stage_request_invalid",
  );
  expect(() =>
    decodeStageRequest(Buffer.from(`\ufeff${JSON.stringify(request)}`)),
  ).toThrow("stage_request_invalid");
});

it("render stage produces a digest-bound real PDF with observed browser traffic", async () => {
  const directory = await mkdtemp(
    path.join(tmpdir(), "cartulary-render-stage-"),
  );
  const listener = createServer();
  await new Promise<void>((resolve, reject) => {
    listener.once("error", reject);
    listener.listen(0, "127.0.0.1", resolve);
  });
  const address = listener.address();
  if (!address || typeof address === "string")
    throw new Error("missing_test_port");
  await new Promise<void>((resolve, reject) =>
    listener.close((error) => (error ? reject(error) : resolve())),
  );
  try {
    const result = await executeStage(
      decodeStageRequest(Buffer.from(JSON.stringify(request))),
      directory,
      address.port,
    );
    const bytes = await readFile(path.join(directory, result.artifact.path));
    expect(result.artifact.media_type).toBe("application/pdf");
    expect(result.artifact.bytes).toBe(bytes.length);
    expect(result.artifact.sha256).toBe(
      createHash("sha256").update(bytes).digest("hex"),
    );
    const document = await PDFDocument.load(bytes, { updateMetadata: false });
    expect(document.getPageCount()).toBe(1);
    expect(document.getCreationDate()?.toISOString()).toBe(
      "2026-04-05T12:00:00.000Z",
    );
    expect(result.browser_observations.length).toBeGreaterThan(0);
    for (const observation of result.browser_observations) {
      expect(observation.allowed).toBe(true);
      expect(observation.loopback_port_ref).toBe("render-loopback-1");
    }
    expect(JSON.stringify(result)).not.toContain("Synthetic, inert");
    expect(JSON.stringify(result)).not.toContain(directory);
  } finally {
    await rm(directory, { recursive: true, force: true });
  }
});
