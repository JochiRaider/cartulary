import type { MermaidConfig } from "mermaid";

export const mermaidFailureAttribute = "data-cartulary-mermaid-failed";

// One profile governs standalone SVG and Slidev's embedded diagrams. Returning
// a fresh object prevents Mermaid's initialization from changing later stages.
export function mermaidProfile(): MermaidConfig {
  return {
    startOnLoad: false,
    securityLevel: "strict",
    htmlLabels: false,
    deterministicIds: true,
    deterministicIDSeed: "cartulary",
    fontFamily: "sans-serif",
    theme: "default",
    look: "classic",
    flowchart: { htmlLabels: false, curve: "linear", useMaxWidth: false },
    sequence: {
      useMaxWidth: false,
      actorFontFamily: "sans-serif",
      messageFontFamily: "sans-serif",
      noteFontFamily: "sans-serif",
    },
  };
}
