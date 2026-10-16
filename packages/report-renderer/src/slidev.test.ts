import { createHash } from "node:crypto";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { createServer as createTCPServer } from "node:net";
import { tmpdir } from "node:os";
import path from "node:path";

import { parser } from "@slidev/cli";
import { PDFDocument } from "pdf-lib";
import { chromium } from "playwright-chromium";
import { expect, it, onTestFinished, vi } from "vitest";
import browserProfile from "./browser-profile.json" with { type: "json" };
import { RenderFailure } from "./failure.ts";
import { renderSlidevPDF } from "./slidev.ts";

it("Slidev parses canonical Reporting source without implicit or blank slides", async () => {
  const fixture = JSON.parse(
    await readFile(
      new URL(
        "../../../contracts/reporting/fixtures/slidev-source.v2.json",
        import.meta.url,
      ),
      "utf8",
    ),
  ) as {
    deck: { title: string; slides: { layout_id: string }[] };
    template: { author: string };
    source_lines: string[];
  };
  const parsed = parser.parseSync(
    fixture.source_lines.join("\n"),
    "synthetic-slides.md",
  );
  expect(parsed.slides).toHaveLength(fixture.deck.slides.length);
  expect(parsed.slides.map((slide) => slide.frontmatter.layout)).toEqual(
    fixture.deck.slides.map((slide) => slide.layout_id),
  );
  expect(parsed.slides[0]?.frontmatter).toEqual({
    title: fixture.deck.title,
    author: fixture.template.author,
    theme: "default",
    download: false,
    export: { format: "pdf", withClicks: true },
    clickAnimation: false,
    lineNumbers: false,
    monaco: false,
    twoslash: false,
    layout: "cover",
  });
  expect(parsed.slides[0]?.content).toContain(
    "# First \\& literal {{ title }}",
  );
  expect(parsed.slides[1]?.content).toContain("# Second page");
  expect(parsed.slides[1]?.content).toContain("**Total:** 2");
  expect(parsed.slides[1]?.content).toContain("- Literal \\| \\<item>");
  expect(parsed.slides[1]?.content).toContain(
    "| Name |\n| --- |\n| Alpha |\n| Beta |",
  );
  expect(parsed.slides[0]?.note).toBe("Private presenter note");
});

// Test-owned port allocation. Production allocates inside the isolated attempt's
// network namespace. strictPort makes an allocation race fail, never rebind.
async function allocatePort() {
  const server = createTCPServer();
  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", resolve);
  });
  const address = server.address();
  if (!address || typeof address === "string")
    throw new Error("missing test port");
  await new Promise<void>((resolve, reject) => {
    server.close((error) => (error ? reject(error) : resolve()));
  });
  return address.port;
}

const source = `---
title: "Cartulary renderer characterization"
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

# Real rendered content

An immutable admitted render.

---
layout: default
---

# A second page

Distinct content must produce a real second PDF page.
`;

it("Slidev renders packaged slides assets and diagrams with repeatable admitted metadata", async () => {
  // Characterize the deployed dependency graph as well as the checkout adapter
  // used by the other cases. Native confinement is exercised by its Go owner;
  // this boundary exposes actual print DOM and synthetic vendor diagnostics.
  const installation = new URL(
    "../../../build/report-renderer/",
    import.meta.url,
  );
  const receipt = JSON.parse(
    await readFile(new URL("installation.json", installation), "utf8"),
  ) as { schema_id: string; root: string; manifest_sha256: string };
  expect(receipt.schema_id).toBe("cartulary.reporting_runtime_receipt.v1");
  expect(receipt.manifest_sha256).toMatch(/^[a-f0-9]{64}$/);
  expect(receipt.root).toBe(`${receipt.manifest_sha256}.root`);
  const installedSource = new URL(
    `${receipt.root}/runtime/renderer/src/`,
    installation,
  );
  // An embedded native stage has no node_modules segment in argv[1]. Vendor
  // CLI heuristics must not make dependency preparation depend on that segment.
  const originalEntry = process.argv[1];
  if (!originalEntry) throw new Error("missing test process entrypoint");
  process.argv[1] = new URL("stage-main.ts", installedSource).pathname;
  onTestFinished(() => {
    process.argv[1] = originalEntry;
  });
  const { renderSlidevPDF } = (await import(
    new URL("slidev.ts", installedSource).href
  )) as typeof import("./slidev.ts");
  const { RenderFailure: InstalledRenderFailure } = (await import(
    new URL("failure.ts", installedSource).href
  )) as typeof import("./failure.ts");
  const profileDirectory = await mkdtemp(
    path.join(tmpdir(), "cartulary-installed-browser-"),
  );
  const browser = await chromium.launch({
    ignoreDefaultArgs: true,
    args: browserProfile.browser_launch_args.map((argument) =>
      argument === "--user-data-dir=/work/browser-profile"
        ? `--user-data-dir=${profileDirectory}`
        : argument,
    ),
  });
  const manualSVG =
    '<svg xmlns="http://www.w3.org/2000/svg" width="100" height="50"><text x="5" y="20">Manual diagram</text></svg>';
  const assets = [
    {
      path: "diagrams/manual.svg",
      media_type: "image/svg+xml",
      bytes: Buffer.from(manualSVG).toString("base64"),
      sha256: createHash("sha256").update(manualSVG).digest("hex"),
    },
  ];
  let failAssetFetch = false;
  let failSlideImport = false;
  // Diagnostics are collected only for this authored synthetic input, never by
  // the production adapter or from incident-bound presentation data.
  const diagnostics: string[] = [];
  const renderedHeadings: string[][] = [];
  const renderedImages: {
    source: string | null;
    width: number;
    height: number;
  }[][] = [];
  const renderedDiagrams: {
    count: number;
    id: string | null;
    foreignObjects: number;
    labels: string[];
  }[] = [];
  const createContext = browser.newContext.bind(browser);
  vi.spyOn(browser, "newContext").mockImplementation(async (options) => {
    const context = await createContext(options);
    context.on("page", (page) => {
      if (failSlideImport)
        void page.route("**/@slidev/slides/2/md*", (route) =>
          route.fulfill({
            status: 200,
            contentType: "text/javascript",
            body: "export default (",
          }),
        );
      if (failAssetFetch)
        void page.route("**/diagrams/manual.svg", (route) =>
          route.fulfill({ status: 404, body: "" }),
        );
      const print = page.pdf.bind(page);
      vi.spyOn(page, "pdf").mockImplementation(async (options) => {
        // The actual print DOM must contain the finished asynchronous SVG,
        // not an empty Mermaid shadow host or a transient render container.
        const diagrams = page.locator(".print-slide-container .mermaid svg");
        const count = await diagrams.count();
        renderedDiagrams.push({
          count,
          id: count === 1 ? await diagrams.getAttribute("id") : null,
          foreignObjects: await diagrams.locator("foreignObject").count(),
          labels: await diagrams.locator("text").allTextContents(),
        });
        renderedHeadings.push(
          await page.locator(".print-slide-container h1").allTextContents(),
        );
        renderedImages.push(
          await page
            .locator(".print-slide-container img")
            .evaluateAll((images) =>
              images.map((image) => ({
                source: image.getAttribute("src"),
                width: (image as HTMLImageElement).naturalWidth,
                height: (image as HTMLImageElement).naturalHeight,
              })),
            ),
        );
        diagnostics.unshift(
          JSON.stringify(
            await page.evaluate(() =>
              ["html", "body", "#page-root", ".print-slide-container"].map(
                (selector) => ({
                  selector,
                  boxes: [...document.querySelectorAll(selector)].map(
                    (node) => {
                      const box = node.getBoundingClientRect();
                      const style = getComputedStyle(node);
                      return {
                        x: box.x,
                        y: box.y,
                        width: box.width,
                        height: box.height,
                        breakAfter: style.breakAfter,
                        display: style.display,
                      };
                    },
                  ),
                }),
              ),
            ),
          ),
        );
        return print(options);
      });
      page.on("pageerror", (error) =>
        diagnostics.push(error.message.slice(0, 1000)),
      );
      page.on("console", (message) => {
        if (message.type() === "error")
          diagnostics.push(message.text().slice(0, 1000));
      });
    });
    return context;
  });
  const directories: string[] = [profileDirectory];
  try {
    const renders = [];
    for (let index = 0; index < 2; index += 1) {
      const directory = await mkdtemp(
        path.join(tmpdir(), "cartulary-renderer-test-"),
      );
      directories.push(directory);
      renders.push(
        await renderSlidevPDF(
          {
            source:
              source.replace(
                "twoslash: false",
                "twoslash: false\nlayout: cover",
              ) +
              "\n![Manual diagram](diagrams/manual.svg)\n" +
              '\n---\nlayout: default\n---\n\n# Automatic diagram\n\n```mermaid\nflowchart TD\n n1["Diagram first"] --> n2["Diagram second"]\n```\n',
            assets,
            admittedAt: "2026-10-14T01:02:03.123456Z",
            identitySHA256: "a".repeat(64),
            expectedPages: 3,
          },
          {
            workDirectory: directory,
            endpoint: { port: await allocatePort(), portRef: "renderer.http" },
            browser,
          },
        ),
      );
    }
    expect(renderedDiagrams).toEqual(
      Array.from({ length: 2 }, () => ({
        count: 1,
        id: "cartulary-diagram",
        foreignObjects: 0,
        labels: ["Diagram first", "Diagram second"],
      })),
    );
    expect(renders[0]?.bytes).toEqual(renders[1]?.bytes);
    expect(renderedImages, JSON.stringify(renderedImages)).toEqual(
      Array.from({ length: 2 }, () => [
        { source: "/diagrams/manual.svg", width: 100, height: 50 },
      ]),
    );
    for (const render of renders) {
      expect(render.bytes.byteLength).toBeGreaterThan(1000);
      const document = await PDFDocument.load(render.bytes, {
        updateMetadata: false,
      });
      expect(document.getPageCount()).toBe(3);
      expect(document.getPage(0).node.Contents()).toBeDefined();
      expect(document.getPage(1).node.Contents()).toBeDefined();
      expect(document.getPage(2).node.Contents()).toBeDefined();
      expect(render.browserObservations.length).toBeGreaterThan(0);
      expect(render.browserObservations.every((entry) => entry.allowed)).toBe(
        true,
      );
    }
    expect(browser.contexts()).toHaveLength(0);
    expect(renderedHeadings).toEqual([
      ["Real rendered content", "A second page", "Automatic diagram"],
      ["Real rendered content", "A second page", "Automatic diagram"],
    ]);
    const failedDirectory = await mkdtemp(
      path.join(tmpdir(), "cartulary-renderer-failed-diagram-"),
    );
    directories.push(failedDirectory);
    await expect(
      renderSlidevPDF(
        {
          source: source + "\n```mermaid\nnot a diagram\n```\n",
          admittedAt: "2026-10-14T01:02:03.123456Z",
          identitySHA256: "a".repeat(64),
          expectedPages: 2,
        },
        {
          workDirectory: failedDirectory,
          endpoint: { port: await allocatePort(), portRef: "renderer.http" },
          browser,
        },
      ),
    ).rejects.toThrow("renderer_fonts_or_paint_failed");
    expect(renderedDiagrams).toHaveLength(2); // Failed diagrams never reach PDF capture.
    expect(browser.contexts()).toHaveLength(0);
    failAssetFetch = true;
    const missingAssetDirectory = await mkdtemp(
      path.join(tmpdir(), "cartulary-renderer-missing-asset-"),
    );
    directories.push(missingAssetDirectory);
    await expect(
      renderSlidevPDF(
        {
          source: source + "\n![Manual diagram](diagrams/manual.svg)\n",
          assets,
          admittedAt: "2026-10-14T01:02:03.123456Z",
          identitySHA256: "a".repeat(64),
          expectedPages: 2,
        },
        {
          workDirectory: missingAssetDirectory,
          endpoint: { port: await allocatePort(), portRef: "renderer.http" },
          browser,
        },
      ),
    ).rejects.toThrow("renderer_resource_missing");
    expect(renderedDiagrams).toHaveLength(2);
    expect(browser.contexts()).toHaveLength(0);
    failAssetFetch = false;
    failSlideImport = true;
    const failedImportDirectory = await mkdtemp(
      path.join(tmpdir(), "cartulary-renderer-failed-import-"),
    );
    directories.push(failedImportDirectory);
    await expect(
      renderSlidevPDF(
        {
          source,
          admittedAt: "2026-10-14T01:02:03.123456Z",
          identitySHA256: "a".repeat(64),
          expectedPages: 2,
        },
        {
          workDirectory: failedImportDirectory,
          endpoint: { port: await allocatePort(), portRef: "renderer.http" },
          browser,
        },
      ),
    ).rejects.toThrow("browser_slide_import_failed");
    expect(renderedDiagrams).toHaveLength(2);
    expect(browser.contexts()).toHaveLength(0);
  } catch (error) {
    if (
      !(error instanceof RenderFailure) &&
      !(error instanceof InstalledRenderFailure)
    )
      throw error;
    throw new Error(
      `Synthetic renderer characterization: ${String(error)}; ${diagnostics.slice(0, 4).join("; ")}`,
    );
  } finally {
    await browser.close();
    await Promise.all(
      directories.map((directory) =>
        rm(directory, { recursive: true, force: true }),
      ),
    );
  }
});

it("Slidev exports each click state and rejects an incorrect expected page count", async () => {
  const browser = await chromium.launch();
  const directories: string[] = [];
  const clickSource = source.replace(
    "An immutable admitted render.",
    '<v-click at="1">\n\nA revealed point.\n\n</v-click>',
  );
  try {
    for (const expectedPages of [3, 2]) {
      const directory = await mkdtemp(
        path.join(tmpdir(), "cartulary-renderer-click-test-"),
      );
      directories.push(directory);
      const result = renderSlidevPDF(
        {
          source: clickSource,
          admittedAt: "2026-10-14T01:02:03Z",
          identitySHA256: "b".repeat(64),
          expectedPages,
        },
        {
          workDirectory: directory,
          endpoint: { port: await allocatePort(), portRef: "renderer.http" },
          browser,
        },
      );
      if (expectedPages === 3) {
        const document = await PDFDocument.load((await result).bytes);
        expect(document.getPageCount()).toBe(3);
      } else {
        await expect(result).rejects.toMatchObject({
          reasonCode: "click_export_page_count_mismatch",
          pageCounts: { expected: 2, observed: 3, source: "slides" },
        });
      }
      expect(browser.contexts()).toHaveLength(0);
    }
  } finally {
    await browser.close();
    await Promise.all(
      directories.map((directory) =>
        rm(directory, { recursive: true, force: true }),
      ),
    );
  }
});

it("Slidev renders generated expressions entities HTML and URLs as literal text", async () => {
  const browser = await chromium.launch();
  const directory = await mkdtemp(
    path.join(tmpdir(), "cartulary-renderer-literal-test-"),
  );
  const title = "{{ 7 * 7 }}";
  const body =
    "{{ (() => { globalThis.__literalProbe = 1; return 49 })() }} &lt; <script>alert('text')</script> https://example.invalid/plain";
  // Exact v2 source escaping; the Go owner has independent byte vectors.
  const generatedSource = source
    .replace("Real rendered content", "{{ 7 \\* 7 }}")
    .replace(
      "An immutable admitted render.",
      "{{ (() => { globalThis.\\_\\_literalProbe = 1; return 49 })() }} \\&lt; \\<script>alert('text')\\</script> https://example.invalid/plain",
    );
  const createContext = browser.newContext.bind(browser);
  let inspected = false;
  vi.spyOn(browser, "newContext").mockImplementation(async (options) => {
    const context = await createContext(options);
    context.on("page", (page) => {
      const print = page.pdf.bind(page);
      vi.spyOn(page, "pdf").mockImplementation(async (options) => {
        const first = page.locator(".print-slide-container").first();
        expect(await first.locator("h1").textContent()).toBe(title);
        expect((await first.locator("p").textContent())?.trim()).toBe(body);
        expect(await first.locator("a, script").count()).toBe(0);
        expect(
          await page.evaluate(() =>
            Object.hasOwn(globalThis, "__literalProbe"),
          ),
        ).toBe(false);
        inspected = true;
        return print(options);
      });
    });
    return context;
  });
  try {
    const output = await renderSlidevPDF(
      {
        source: generatedSource,
        admittedAt: "2026-10-14T01:02:03Z",
        identitySHA256: "c".repeat(64),
        expectedPages: 2,
      },
      {
        workDirectory: directory,
        endpoint: { port: await allocatePort(), portRef: "renderer.http" },
        browser,
      },
    );
    expect(inspected).toBe(true);
    expect((await PDFDocument.load(output.bytes)).getPageCount()).toBe(2);
  } finally {
    await browser.close();
    await rm(directory, { recursive: true, force: true });
  }
});
