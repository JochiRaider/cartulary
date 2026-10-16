import { readFile } from "node:fs/promises";
import { createServer } from "node:net";

import { chromium } from "playwright-chromium";
import { expect, it, vi } from "vitest";

import { renderMermaidSVG, validateMermaidSource } from "./mermaid.ts";
import { materializePassiveSVG } from "./passive-svg.ts";

const flowchart =
  'flowchart TD\n  n0001["First node"]\n  n0002["Second node"]\n  n0001 -->|"Connects"| n0002\n';
const sequence =
  "sequenceDiagram\n  participant p0001 as First participant\n  participant p0002 as Second participant\n  p0001->>p0002: Message\n";

it("Mermaid admission accepts only canonical bounded flowchart and sequence source", () => {
  for (const source of [flowchart, sequence])
    expect(() => validateMermaidSource(source)).not.toThrow();
  for (const source of [
    "pie\n",
    `${flowchart}  click n0001 "https://example.invalid"\n`,
    `${flowchart}  style n0001 fill:red\n`,
    flowchart.replace("n0002[", "n0003["),
    flowchart.replace("n0002\n", "n0003\n"),
    flowchart.replace("First node", "<b>HTML</b>"),
    flowchart.replace("First node", "x".repeat(241)),
    flowchart.replace('"Connects"', '""'),
    `%%{init: {}}%%\n${flowchart}`,
    flowchart.replaceAll("\n", "\r\n"),
    flowchart.slice(0, -1),
    sequence.replace("->>", "-->>"),
    `${sequence}  Note over p0001: note\n`,
  ])
    expect(() => validateMermaidSource(source)).toThrow(
      "invalid_mermaid_construct",
    );
});

async function allocatePort() {
  const server = createServer();
  await new Promise<void>((resolve, reject) => {
    server.once("error", reject);
    server.listen(0, "127.0.0.1", resolve);
  });
  const address = server.address();
  if (!address || typeof address === "string")
    throw new Error("missing_test_port");
  await new Promise<void>((resolve, reject) =>
    server.close((error) => (error ? reject(error) : resolve())),
  );
  return address.port;
}

it("Mermaid renders real graph geometry arrowheads and sequence messages as repeatable passive SVG", async () => {
  const browser = await chromium.launch();
  const diagnostics: string[] = [];
  const newContext = browser.newContext.bind(browser);
  // Only authored synthetic fixtures enter this diagnostic capture. Production
  // never retains vendor exceptions containing incident text or destinations.
  vi.spyOn(browser, "newContext").mockImplementation(async (options) => {
    const context = await newContext(options);
    context.on("page", (page) => {
      const evaluate = page.evaluate.bind(page);
      vi.spyOn(page, "evaluate").mockImplementation(async (...args) => {
        try {
          return await evaluate(...args);
        } catch (error) {
          diagnostics.push(String(error).slice(0, 1000));
          diagnostics.push(
            JSON.stringify(
              await evaluate(() =>
                [...document.querySelectorAll("#output svg *")]
                  .filter((element) => {
                    const style = getComputedStyle(element);
                    return (
                      style.filter !== "none" ||
                      style.clipPath !== "none" ||
                      style.maskImage !== "none"
                    );
                  })
                  .slice(0, 8)
                  .map((element) => ({
                    tag: element.tagName,
                    filter: getComputedStyle(element).filter,
                    clip: getComputedStyle(element).clipPath,
                    mask: getComputedStyle(element).maskImage,
                    attrs: [...element.attributes].map(({ name, value }) => [
                      name,
                      value,
                    ]),
                  })),
              ),
            ),
          );

          throw error;
        }
      });
    });
    return context;
  });
  try {
    const vectors = JSON.parse(
      await readFile(
        new URL(
          "../../../contracts/reporting/fixtures/mermaid-source.v3.json",
          import.meta.url,
        ),
        "utf8",
      ),
    ) as { cases: { source_lines: string[]; visible_labels: string[] }[] };
    const cases = [
      { source: flowchart, labels: ["First node", "Second node", "Connects"] },
      {
        source: sequence,
        labels: ["First participant", "Second participant", "Message"],
      },
      ...vectors.cases.map((vector) => ({
        source: vector.source_lines.join("\n"),
        labels: vector.visible_labels,
      })),
    ];
    for (const { source, labels: expectedLabels } of cases) {
      const outputs: Buffer[] = [];
      for (let attempt = 0; attempt < 2; attempt += 1) {
        const result = await renderMermaidSVG(source, {
          browser,
          endpoint: {
            port: await allocatePort(),
            portRef: "render-loopback-1",
          },
        });
        expect(result.browserObservations.length).toBeGreaterThan(0);
        expect(
          result.browserObservations.every(
            (item) => item.allowed && item.stage === "mermaid_render",
          ),
        ).toBe(true);
        outputs.push(result.bytes);
      }
      expect(outputs[0]).toEqual(outputs[1]);
      const svg = outputs[0]?.toString("utf8") ?? "";

      const page = await browser.newPage();
      try {
        await page.setContent(svg);
        expect(await page.locator("svg").count()).toBe(1);
        expect(
          await page.locator("style, marker, foreignObject, script").count(),
        ).toBe(0);
        // Reference syntax in inert label text is allowed. Inspect the actual
        // element attributes, which can grant rendering capabilities.
        expect(
          await page
            .locator("svg, svg *")
            .evaluateAll((elements) =>
              elements.flatMap((element) =>
                [...element.attributes]
                  .filter(
                    (attribute) =>
                      [
                        "style",
                        "class",
                        "id",
                        "marker-end",
                        "stroke-dasharray",
                      ].includes(attribute.name) ||
                      /url\(/iu.test(attribute.value),
                  )
                  .map((attribute) => attribute.name),
              ),
            ),
        ).toEqual([]);
        expect(
          await page.locator("svg path, svg line").count(),
        ).toBeGreaterThan(1);
        expect(await page.locator('g[transform*="rotate("]').count()).toBe(1);
        expect(await page.locator("svg text").count()).toBeGreaterThanOrEqual(
          expectedLabels.length,
        );
        // Mermaid uses positioned word spans. Join leaf text with its rendered
        // word boundaries rather than treating adjacent SVG text nodes as prose.
        const labels = (
          await page
            .locator("svg text:not(:has(tspan)), svg tspan:not(:has(tspan))")
            .allTextContents()
        )
          .join(" ")
          .replace(/\s+/g, " ");
        for (const label of expectedLabels) expect(labels).toContain(label);
      } finally {
        await page.close();
      }
    }
  } catch (error) {
    throw new Error(
      `${String(error)}; synthetic fixture diagnostics: ${diagnostics.join("; ")}`,
    );
  } finally {
    await browser.close();
  }
}, 30_000);

it("passive SVG construction rejects active elements references and unsupported effects", async () => {
  const browser = await chromium.launch();
  try {
    const page = await browser.newPage();
    for (const content of [
      "<foreignObject><div>HTML</div></foreignObject>",
      '<use href="#target"/>',
      '<rect width="10" height="10" onclick="void 0"/>',
      '<rect width="10" height="10" style="filter: blur(1px)"/>',
      '<path d="M 0 0 L 10 10" style="marker-end:url(https://example.invalid/marker)"/>',
    ]) {
      await page.setContent(
        `<svg xmlns="http://www.w3.org/2000/svg">${content}</svg>`,
      );
      await expect(
        page.evaluate(materializePassiveSVG, "svg"),
      ).rejects.toThrow();
    }
  } finally {
    await browser.close();
  }
});
