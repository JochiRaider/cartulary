import type { StageAsset } from "./stage-assets.ts";

// Template CSS declares only the admitted presentation tokens. Selectors and
// their relationship to Slidev are renderer-owned, so templates cannot depend
// on private component markup or hide/reorder case content with arbitrary CSS.
export function themeStyleHTML(assets: readonly StageAsset[]): string {
  const sheet = assets.find((asset) => asset.path === "theme/style.css");
  if (!sheet) return "";
  const css = Buffer.from(sheet.bytes, "base64").toString("utf8");
  const rules: string[] = [];
  const mappings: readonly [string, string, string][] = [
    ["background", ".slidev-layout", "background-color"],
    ["text", ".slidev-layout", "color"],
    ["font-family", ".slidev-layout", "font-family"],
    [
      "heading-font-family",
      ".slidev-layout h1,.slidev-layout h2,.slidev-layout h3",
      "font-family",
    ],
    ["font-size", ".slidev-layout", "font-size"],
    [
      "accent",
      ".slidev-layout h1,.slidev-layout h2,.slidev-layout h3",
      "color",
    ],
    ["accent", ".slidev-layout th", "border-color"],
  ];
  const present = (key: string) =>
    new RegExp(`--cartulary-report-${key}\\s*:`).test(css);
  for (const [key, selector, property] of mappings)
    if (present(key))
      rules.push(
        `${selector}{${property}:var(--cartulary-report-${key}) !important}`,
      );
  const logo = assets.find((asset) => asset.path.startsWith("theme/images/"));
  if (present("logo") && logo)
    rules.push(
      '.slidev-layout{position:relative}.slidev-layout::after{content:"";position:absolute;top:12px;right:12px;width:100px;height:48px;background-image:url(' +
        JSON.stringify("/" + logo.path) +
        ");background-size:contain;background-repeat:no-repeat;background-position:right top;pointer-events:none}",
    );
  return (
    '<link rel="stylesheet" href="/theme/style.css"><style>' +
    rules.join("") +
    "</style>"
  );
}
