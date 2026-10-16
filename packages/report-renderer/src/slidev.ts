import { mkdir, writeFile } from "node:fs/promises";
import { createServer as createHTTPServer } from "node:http";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  createDataUtils,
  createServer,
  resolveOptions,
  type ViteSlidevPlugin,
} from "@slidev/cli";
import type { Browser } from "playwright-chromium";

import {
  type BrowserEndpoint,
  installBrowserPolicy,
} from "./browser-policy.ts";
import { RenderFailure } from "./failure.ts";
import { renderLiteralText } from "./literal-text.ts";
import { mermaidFailureAttribute, mermaidProfile } from "./mermaid-profile.ts";
import { normalizeRenderedPDF, type PDFBinding } from "./pdf.ts";
import {
  type StageAsset,
  stageAssetURL,
  validateStageAssets,
} from "./stage-assets.ts";
import { themeStyleHTML } from "./theme-style.ts";

export interface SlidevPDFInput extends PDFBinding {
  source: string;
  assets?: readonly StageAsset[];
}

// The caller owns an isolated, empty work directory, the declared endpoint,
// browser process, OS sandbox and stage deadline. This adapter owns Slidev's
// server and browser context and always closes both. It is not a sandbox launcher.
export async function renderSlidevPDF(
  input: SlidevPDFInput,
  runtime: {
    workDirectory: string;
    endpoint: BrowserEndpoint;
    browser: Browser;
  },
) {
  const entry = path.join(runtime.workDirectory, "slides.md");
  await writeFile(entry, input.source, { flag: "wx", mode: 0o600 });
  // This is renderer-owned code, generated exclusively from installed module
  // identity and the closed profile. No incident input becomes JavaScript.
  // Slidev's supported setup boundary avoids its default random IDs and HTML
  // labels, while canonical Markdown continues to contain Mermaid fences.
  const setup = path.join(runtime.workDirectory, "setup");
  await mkdir(setup, { mode: 0o700 });
  await writeFile(
    path.join(setup, "mermaid-renderer.mjs"),
    `import engine from ${JSON.stringify(fileURLToPath(import.meta.resolve("mermaid/dist/mermaid.esm.mjs")))};\n` +
      `export default () => async (source) => {\n` +
      `  engine.initialize(${JSON.stringify(mermaidProfile())});\n` +
      `  try { return (await engine.render("cartulary-diagram", source)).svg; }\n` +
      `  catch {\n` +
      `    document.documentElement.setAttribute(${JSON.stringify(mermaidFailureAttribute)}, "true");\n` +
      `    throw new Error("mermaid_render_failed");\n` +
      `  }\n` +
      `};\n`,
    { flag: "wx", mode: 0o600 },
  );
  if (input.assets?.length) {
    validateStageAssets(input.assets);
    for (const asset of input.assets) {
      // Admission limits destinations to the private public/assets and
      // public/diagrams namespaces. All bytes were frozen before rendering.
      const destination = path.join(
        runtime.workDirectory,
        "public",
        asset.path,
      );
      await mkdir(path.dirname(destination), { recursive: true, mode: 0o700 });
      await writeFile(destination, Buffer.from(asset.bytes, "base64"), {
        flag: "wx",
        mode: 0o600,
      });
    }
  }
  const theme = path.dirname(
    fileURLToPath(import.meta.resolve("@slidev/theme-default/package.json")),
  );
  const options = await resolveOptions(
    { entry, theme, routerMode: "history", withoutNotes: true },
    "export",
  );
  // Font selection is an installed render-profile capability. Theme defaults
  // must not inject a remote font provider into canonical source that carries
  // no font settings. Rebuild derived HTML/defines from the admitted local set.
  options.data.config.fonts = {
    ...options.data.config.fonts,
    provider: "none",
    webfonts: [],
    sans: ["sans-serif"],
    serif: ["serif"],
    mono: ["monospace"],
    local: ["sans-serif", "serif", "monospace"],
  };
  options.data.config.canvasWidth = 1280;
  options.data.config.aspectRatio = 16 / 9;
  options.utils = await createDataUtils(options);
  const themeHTML = themeStyleHTML(input.assets ?? []);
  if (themeHTML)
    options.utils.indexHtml = options.utils.indexHtml.replace(
      "</head>",
      themeHTML + "</head>",
    );
  const width = options.data.config.canvasWidth;
  // Pinned Slidev rounds its canvas height up in client/env.ts. Use the same
  // geometry: a fractional page height would spill the final slide onto a page.
  const height = Math.ceil(width / options.data.config.aspectRatio);
  if (
    !Number.isFinite(width) ||
    !Number.isFinite(height) ||
    width <= 0 ||
    height <= 0
  ) {
    throw new RenderFailure("slidev_export_failed", "invalid_canvas_geometry");
  }
  const context = await runtime.browser.newContext({
    viewport: {
      width: Math.ceil(width),
      height: Math.ceil(height),
    },
    deviceScaleFactor: 1,
    locale: "en-US",
    timezoneId: "UTC",
    colorScheme: "light",
    serviceWorkers: "block",
  });
  let server: Awaited<ReturnType<typeof createServer>> | undefined;
  const endpoint = createHTTPServer(async (request, response) => {
    if (!server) return void response.writeHead(503).end();
    if (request.method !== "GET") return void response.writeHead(405).end();
    if (request.url === "/print?print=clicks") {
      try {
        // Middleware hosting must apply Vite's HTML transform explicitly. It
        // installs the pinned client environment defines before Slidev starts.
        const html = await server.transformIndexHtml(
          request.url,
          options.utils.indexHtml,
        );
        response
          .writeHead(200, { "content-type": "text/html; charset=utf-8" })
          .end(html);
      } catch {
        response.writeHead(500).end();
      }
    } else server.middlewares(request, response);
  });
  try {
    const policy = await installBrowserPolicy(
      context,
      runtime.endpoint,
      "slidev_export",
    );
    let failureReason = "renderer_server_start_failed";
    try {
      // Slidev reads its plugin options from this Vite extension. Use the
      // exported plugin's parameter type rather than weakening the config to any.
      const serverOptions: NonNullable<Parameters<typeof createServer>[1]> & {
        slidev: NonNullable<Parameters<typeof ViteSlidevPlugin>[1]>;
      } = {
        logLevel: "silent",
        cacheDir: path.join(runtime.workDirectory, "vite-cache"),
        // Slidev's CLI guesses global/local dependency preparation from argv.
        // Our embedded entrypoint has neither CLI shape. Resolve its required
        // CommonJS decoder from the admitted client package and explicitly
        // prepare browser interop, independent of cwd or the launch path.
        resolve: {
          alias: [
            {
              find: /^lz-string$/,
              replacement: createRequire(
                path.join(options.clientRoot, "package.json"),
              ).resolve("lz-string"),
            },
          ],
        },
        optimizeDeps: { include: ["lz-string"], needsInterop: ["lz-string"] },
        slidev: {
          // Authoring inspectors own extra RPC endpoints and are outside the
          // closed render profile, including when Vite runs as middleware.
          unocss: { inspector: false },
          // Assets are declared HTTP resources of this isolated attempt, not
          // JavaScript imports. Prevent Vite from rewriting them to data URLs
          // or consulting a virtual slide's filesystem location.
          vue: { template: { transformAssetUrls: false } },
          markdown: {
            markdownOptions: { linkify: false, typographer: false },
            markdownSetup(md) {
              const renderImage = md.renderer.rules.image;
              const declaredAssets = new Set(
                input.assets?.map((asset) => stageAssetURL(asset.path)),
              );
              md.renderer.rules.image = (tokens, index, options, env, self) => {
                const token = tokens[index];
                const source = token?.attrGet("src");
                if (
                  !token ||
                  !source ||
                  !declaredAssets.has(source) ||
                  !renderImage
                )
                  throw new RenderFailure(
                    "slidev_source_invalid",
                    "stage_request_invalid",
                  );
                // Canonical bundle paths stay relative in slides.md. Vite's
                // Vue compiler otherwise interprets them as module imports
                // relative to a virtual slide, rather than public assets.
                // The browser receives only this attempt's admitted URL.
                token.attrSet("src", `/${source}`);
                return renderImage(tokens, index, options, env, self);
              };
              md.renderer.rules.text = (tokens, index) => {
                const token = tokens[index];
                if (!token)
                  throw new RenderFailure(
                    "slidev_source_invalid",
                    "invalid_markdown_text_token",
                  );
                return renderLiteralText(token.content);
              };
            },
          },
        },
        server: {
          // Vite's ordinary listen path probes wildcard addresses even with
          // strictPort. The supervisor owns the only endpoint, so use Vite as
          // middleware and bind that endpoint directly without discovery.
          middlewareMode: true,
          hmr: false,
          // Vite's client environment and Slidev's reactive print state use
          // its websocket transport even with HMR disabled. Share the exact
          // supervisor-owned HTTP endpoint; never allocate a second listener.
          ws: {
            server: endpoint,
            host: "127.0.0.1",
            port: runtime.endpoint.port,
            clientPort: runtime.endpoint.port,
          },
          host: "127.0.0.1",
          port: runtime.endpoint.port,
          strictPort: true,
          watch: null,
        },
      };
      server = await createServer(options, serverOptions);
      await new Promise<void>((resolve, reject) => {
        endpoint.once("error", reject);
        endpoint.listen(runtime.endpoint.port, "127.0.0.1", resolve);
      });
      const page = await context.newPage();
      let browserError: string | undefined;
      let resourceError: string | undefined;
      page.on("pageerror", () => {
        browserError ??= "browser_render_error";
      });
      // Slidev catches asynchronous slide-import failures and renders an error
      // layout. Such failures do not raise pageerror and must never become an
      // apparently successful PDF. Match the pinned loader's diagnostic rather
      // than unrelated authoring-plugin console messages; retain no raw text.
      page.on("console", (message) => {
        if (
          message.type() === "error" &&
          message.text().startsWith("slide failed to load")
        )
          browserError ??= message
            .text()
            .includes("does not provide an export named")
            ? "browser_module_export_missing"
            : "browser_slide_import_failed";
      });
      page.on("response", (response) => {
        if (response.status() >= 400) {
          resourceError ??=
            response.status() === 404
              ? "renderer_resource_missing"
              : response.status() === 500
                ? "renderer_resource_transform_failed"
                : "renderer_resource_load_failed";
        }
      });
      page.on("requestfailed", () => {
        resourceError ??= "renderer_resource_load_failed";
      });
      await page.emulateMedia({ media: "screen", colorScheme: "light" });
      failureReason = "renderer_page_load_failed";
      await page.goto(
        `http://127.0.0.1:${runtime.endpoint.port}/print?print=clicks`,
        { waitUntil: "networkidle" },
      );
      if (resourceError)
        throw new RenderFailure("slidev_export_failed", resourceError);
      if (browserError)
        throw new RenderFailure("slidev_export_failed", browserError);
      // This adapter is bound to the pinned Slidev print surface. Count real
      // expanded pages, not source headings or request completion.
      failureReason = "renderer_print_surface_missing";
      await page.locator(".print-slide-container").first().waitFor();
      await page
        .locator(".slidev-slide-loading")
        .waitFor({ state: "detached" });
      // Export uses a paginated block flow. Slidev's screen grid includes the
      // Twoslash side column and a viewport-height root; those screen containers
      // must not add a trailing PDF page after the final slide's page break.
      await page.addStyleTag({
        content: `
        @page { size: ${width}px ${height}px; margin: 0; }
        html.print, html.print body, html.print #app { height: auto; }
        html.print #page-root { display: block; }
        .print-slide-container:last-child { break-after: auto; }
      `,
      });
      // The pinned Slidev Mermaid component fills an initially empty shadow
      // root after its asynchronous renderer finishes. The slide loading
      // indicator covers component loading, not that rendering promise.
      // Await every mounted diagram's SVG before fonts, paint and PDF capture.
      failureReason = "renderer_fonts_or_paint_failed";
      await page.waitForFunction(
        (failureAttribute) =>
          document.documentElement.hasAttribute(failureAttribute) ||
          Array.from(
            document.querySelectorAll(".print-slide-container .mermaid"),
          ).every((host) => {
            const svg = host.shadowRoot?.querySelector("svg");
            if (!(svg instanceof SVGSVGElement)) return false;
            const box = svg.viewBox.baseVal;
            return (
              box.width > 0 &&
              box.height > 0 &&
              svg.querySelector("text, path, polygon, rect") !== null
            );
          }),
        mermaidFailureAttribute,
      );
      if (await page.locator("html").getAttribute(mermaidFailureAttribute))
        throw new RenderFailure(
          "slidev_export_failed",
          "renderer_fonts_or_paint_failed",
        );
      await page.evaluate(
        async ({ theme, fontCount, images }) => {
          if (theme) {
            const sheets = Array.from(document.styleSheets).filter((sheet) =>
              sheet.href?.endsWith("/theme/style.css"),
            );
            if (sheets.length !== 1 || !sheets[0]?.cssRules.length)
              throw new Error("theme_styles_missing");
            const faces = Array.from(document.fonts).filter((face) =>
              face.family.replace(/^"|"$/g, "").startsWith("rpt-"),
            );
            if (faces.length !== fontCount)
              throw new Error("theme_fonts_missing");
            await Promise.all(
              faces.map(async (face) => {
                await face.load();
                if (face.status !== "loaded")
                  throw new Error("theme_font_failed");
              }),
            );
            await Promise.all(
              images.map(async (source) => {
                const image = new Image();
                image.src = "/" + source;
                await image.decode();
                if (
                  !image.complete ||
                  image.naturalWidth < 1 ||
                  image.naturalHeight < 1
                )
                  throw new Error("theme_image_failed");
              }),
            );
          }
          await document.fonts.ready;
          await Promise.all(
            Array.from(
              document.querySelectorAll(".print-slide-container img"),
            ).map(async (image) => {
              if (!(image instanceof HTMLImageElement))
                throw new Error("invalid_diagram_image");
              await image.decode();
              if (
                !image.complete ||
                image.naturalWidth < 1 ||
                image.naturalHeight < 1
              )
                throw new Error("diagram_image_missing");
            }),
          );
          await new Promise<void>((resolve) => {
            requestAnimationFrame(() => requestAnimationFrame(() => resolve()));
          });
        },
        {
          theme: !!themeHTML,
          fontCount:
            input.assets?.filter((asset) => asset.media_type === "font/woff2")
              .length ?? 0,
          images:
            input.assets
              ?.filter((asset) => asset.path.startsWith("theme/images/"))
              .map((asset) => stageAssetURL(asset.path)) ?? [],
        },
      );
      policy.assertAllowed();
      if (resourceError)
        throw new RenderFailure("slidev_export_failed", resourceError);
      if (browserError)
        throw new RenderFailure("slidev_export_failed", browserError);
      const observedPages = await page
        .locator(".print-slide-container")
        .count();
      if (observedPages !== input.expectedPages) {
        throw new RenderFailure(
          "slidev_export_failed",
          "click_export_page_count_mismatch",
          {
            expected: input.expectedPages,
            observed: observedPages,
            source: "slides",
          },
        );
      }
      failureReason = "renderer_pdf_export_failed";
      const bytes = await page.pdf({
        width: `${width}px`,
        height: `${height}px`,
        margin: { top: 0, bottom: 0, left: 0, right: 0 },
        printBackground: true,
        preferCSSPageSize: true,
      });
      policy.assertAllowed();
      return {
        bytes: await normalizeRenderedPDF(bytes, input),
        browserObservations: policy.observations(),
      };
    } catch (error) {
      policy.assertNoViolation();
      if (error instanceof RenderFailure) throw error;
      throw new RenderFailure("slidev_export_failed", failureReason);
    }
  } finally {
    // Close print clients before the shared endpoint. HTTP close waits for
    // upgraded websocket connections; closing it first deadlocks teardown.
    // Every owner still gets its cleanup if an earlier close fails.
    try {
      await context.close();
    } finally {
      try {
        await server?.close();
      } finally {
        endpoint.closeAllConnections();
        await new Promise<void>((resolve) => endpoint.close(() => resolve()));
      }
    }
  }
}
