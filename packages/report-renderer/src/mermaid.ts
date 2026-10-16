import { realpath } from "node:fs/promises";
import { createServer } from "node:http";
import path from "node:path";
import { fileURLToPath } from "node:url";

import type { Browser } from "playwright-chromium";

import {
  type BrowserEndpoint,
  installBrowserPolicy,
} from "./browser-policy.ts";
import { RenderFailure } from "./failure.ts";
import { mermaidProfile } from "./mermaid-profile.ts";
import { readRendererModule, rendererModulePath } from "./module-assets.ts";
import { materializePassiveSVG } from "./passive-svg.ts";

// Defense at the language boundary: only owner-generated canonical source is
// admitted. Mermaid's general language includes executable presentation features
// that are not part of Reporting's contract, even with strict mode enabled.
export function validateMermaidSource(source: string): void {
  const invalid = () => {
    throw new RenderFailure("mermaid_invalid", "invalid_mermaid_construct");
  };
  if (
    !source.endsWith("\n") ||
    source.includes("\r") ||
    source.length > 256 * 1024
  )
    invalid();
  const lines = source.slice(0, -1).split("\n");
  const header = lines.shift();
  const flowchart = header === "flowchart TD" || header === "flowchart LR";
  if (!flowchart && header !== "sequenceDiagram") invalid();
  const nodes = new Set<string>();
  let edges = 0;
  let edgePhase = false;
  const label = (value: string) => {
    let text = "";
    const reserved = new Set([34, 35, 37, 38, 39, 59, 92, 96, 124]);
    for (let index = 0; index < value.length; ) {
      const character = value[index] ?? "";
      if (character === "#") {
        const entity = /^#(32|34|35|37|38|39|59|92|96|124);/.exec(
          value.slice(index),
        );
        if (!entity) invalid();
        text += String.fromCharCode(Number(entity?.[1]));
        index += entity?.[0].length ?? 1;
      } else {
        if (reserved.has(character.charCodeAt(0))) invalid();
        text += character;
        index += 1;
      }
    }
    const canonical = [...text]
      .map((character, index, characters) =>
        reserved.has(character.codePointAt(0) ?? 0) ||
        (character === " " && (index === 0 || index === characters.length - 1))
          ? `#${character.codePointAt(0)};`
          : character,
      )
      .join("");
    if (canonical !== value) invalid();
    if (
      /[<>\p{Cc}]/u.test(text) ||
      text.includes("```") ||
      text !== text.normalize("NFC") ||
      [...text].length > 240
    )
      invalid();
  };
  for (const line of lines) {
    const declaration = flowchart
      ? /^ {2}(n\d{4})\["((?:[^"\\]|\\["\\])*)"\]$/.exec(line)
      : /^ {2}participant (p\d{4}) as (.*)$/.exec(line);
    if (declaration) {
      const id = declaration[1] ?? "";
      if (
        edgePhase ||
        id !==
          `${flowchart ? "n" : "p"}${String(nodes.size + 1).padStart(4, "0")}`
      )
        invalid();
      nodes.add(id);
      label(declaration[2] ?? "");
      continue;
    }
    const unlabelled = flowchart
      ? /^ {2}(n\d{4}) --> (n\d{4})$/.exec(line)
      : null;
    if (unlabelled) {
      if (!nodes.has(unlabelled[1] ?? "") || !nodes.has(unlabelled[2] ?? ""))
        invalid();
      edges += 1;
      edgePhase = true;
      continue;
    }
    const edge = flowchart
      ? /^ {2}(n\d{4}) -->\|"((?:[^"\\]|\\["\\])*)"\| (n\d{4})$/.exec(line)
      : /^ {2}(p\d{4})->>(p\d{4}): ((?:[^"\\]|\\["\\])*)$/.exec(line);
    if (!edge) invalid();
    if (flowchart && edge?.[2] === "") invalid();
    const from = edge?.[1] ?? "";
    const to = edge?.[flowchart ? 3 : 2] ?? "";
    if (!nodes.has(from) || !nodes.has(to)) invalid();
    label(edge?.[flowchart ? 2 : 3] ?? "");
    edges += 1;
    edgePhase = true;
  }
  if (!nodes.size || nodes.size > (flowchart ? 80 : 24) || edges > 160)
    invalid();
}

// The caller owns the browser, OS sandbox, allocated endpoint and stage timer.
// This adapter owns only a closed local static server and its browser context.
export async function renderMermaidSVG(
  source: string,
  runtime: { browser: Browser; endpoint: BrowserEndpoint },
) {
  validateMermaidSource(source);
  const packageRoot = await realpath(
    path.dirname(fileURLToPath(import.meta.resolve("mermaid"))),
  );
  const server = createServer(async (request, response) => {
    if (request.method !== "GET") {
      response.writeHead(405).end();
    } else if (request.url === "/") {
      response.writeHead(200, { "content-type": "text/html; charset=utf-8" });
      response.end(
        '<!doctype html><html><head><meta charset="utf-8"></head><body><div id="output"></div><script type="module" src="/entry.mjs"></script></body></html>',
      );
    } else if (request.url === "/entry.mjs") {
      response.writeHead(200, {
        "content-type": "text/javascript; charset=utf-8",
      });
      response.end(
        'import mermaid from "/renderer/mermaid.esm.mjs"; globalThis.mermaid = mermaid;',
      );
    } else if (request.url && rendererModulePath(request.url) !== null) {
      try {
        const script = await readRendererModule(packageRoot, request.url);
        response.writeHead(200, {
          "content-type": "text/javascript; charset=utf-8",
        });
        response.end(script);
      } catch {
        response.writeHead(404).end();
      }
    } else response.writeHead(404).end();
  });
  const context = await runtime.browser.newContext({
    viewport: { width: 1280, height: 720 },
    deviceScaleFactor: 1,
    locale: "en-US",
    timezoneId: "UTC",
    colorScheme: "light",
    serviceWorkers: "block",
  });
  try {
    const policy = await installBrowserPolicy(
      context,
      runtime.endpoint,
      "mermaid_render",
    );
    try {
      await new Promise<void>((resolve, reject) => {
        server.once("error", reject);
        server.listen(runtime.endpoint.port, "127.0.0.1", resolve);
      });
      const page = await context.newPage();
      await page.goto(`http://127.0.0.1:${runtime.endpoint.port}/`, {
        waitUntil: "networkidle",
      });
      await page.evaluate(
        async ({ text, profile }) => {
          const engine = (
            globalThis as unknown as {
              mermaid: typeof import("mermaid").default;
            }
          ).mermaid;
          engine.initialize(profile);
          const rendered = await engine.render("cartulary-diagram", text);
          const output = document.querySelector("#output");
          if (!output) throw new Error("missing_output");
          output.innerHTML = rendered.svg;
          await document.fonts.ready;
          await new Promise<void>((resolve) =>
            requestAnimationFrame(() => requestAnimationFrame(() => resolve())),
          );
        },
        { text: source, profile: mermaidProfile() },
      );
      policy.assertAllowed();
      const svg = await page.evaluate(materializePassiveSVG, "#output > svg");
      policy.assertAllowed();
      return {
        bytes: Buffer.from(svg, "utf8"),
        browserObservations: policy.observations(),
      };
    } catch (error) {
      policy.assertNoViolation();
      if (error instanceof RenderFailure) throw error;
      throw new RenderFailure(
        "mermaid_render_failed",
        "renderer_svg_export_failed",
      );
    }
  } finally {
    try {
      server.closeAllConnections();
      if (server.listening)
        await new Promise<void>((resolve, reject) =>
          server.close((error) => (error ? reject(error) : resolve())),
        );
    } finally {
      await context.close();
    }
  }
}
