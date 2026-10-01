import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import type { Page } from "@playwright/test";

export type MetadataRule = {
  id: string;
  target: string;
  replacement: string;
  expected_count: number;
};
type Original = { element: HTMLElement; text: string; replacement: string };
type NormalizationWindow = Window & {
  __cartularyMetadataOriginals?: Original[];
};
const policySource = readFileSync(
  new URL(
    "../../../../../tools/frontend_visual_normalization.json",
    import.meta.url,
  ),
  "utf8",
);
const policyDigest = createHash("sha256").update(policySource).digest("hex");
const policy = JSON.parse(policySource) as {
  schema_id: string;
  captures: Record<string, MetadataRule[]>;
  bindings: Record<string, { capture_id: string; golden_path: string }>;
};

// Runs in the page. Validate the complete declaration before changing any text.
export function normalizeMetadataDocument(rules: MetadataRule[]) {
  const state = window as NormalizationWindow;
  if (state.__cartularyMetadataOriginals)
    throw new Error("Nested visual normalization");
  const protectedSource =
    '[data-grid-field-key]:not([data-grid-field-key="evidence.edited_at"]), [data-history-value], input, textarea, [contenteditable="true"], [data-source-value]';
  const selected: Original[] = [];
  const receipt: {
    id: string;
    target: string;
    expected_count: number;
    applied_count: number;
  }[] = [];
  for (const rule of rules) {
    if (
      !/^\[data-generated-metadata="[a-z][a-z-]*"\]$/.test(rule.target) ||
      !Number.isInteger(rule.expected_count) ||
      rule.expected_count < 0
    )
      throw new Error("Invalid metadata declaration");
    // Admission is over mounted metadata leaves, including closed disclosures.
    // Geometry measurement belongs to capture readiness, not target admission.
    const elements = [...document.querySelectorAll<HTMLElement>(rule.target)];
    if (elements.length !== rule.expected_count)
      throw new Error(
        `Metadata cardinality ${rule.id}: expected ${rule.expected_count}, observed ${elements.length}`,
      );
    for (const element of elements) {
      const inspectorField = element.closest("[data-inspector-saved-field]");
      const inspectorMetadataTargets: Record<string, string> = {
        "timeline.recorded_at": "timeline-recorded-at",
        "timeline.edited_at": "timeline-edited-at",
        "evidence.edited_at": "evidence-inspector-edited-at",
      };
      const metadataSurface =
        inspectorMetadataTargets[
          inspectorField?.getAttribute("data-inspector-saved-field") ?? ""
        ];
      const inspectorMetadata =
        inspectorField?.getAttribute("data-inspector-field-write-kind") ===
          "read_only" &&
        metadataSurface !== undefined &&
        rule.target === `[data-generated-metadata="${metadataSurface}"]`;
      const gridField = element
        .closest("[data-grid-field-key]")
        ?.getAttribute("data-grid-field-key");
      if (
        (inspectorField && !inspectorMetadata) ||
        (gridField &&
          (gridField !== "evidence.edited_at" ||
            rule.target !==
              '[data-generated-metadata="evidence-edited-at"]')) ||
        element.childElementCount ||
        element.closest(protectedSource) ||
        element.querySelector(protectedSource)
      )
        throw new Error(`Metadata source overlap ${rule.id}`);
      if (
        selected.some(
          (item) =>
            item.element === element ||
            item.element.contains(element) ||
            element.contains(item.element),
        )
      )
        throw new Error(`Metadata rule overlap ${rule.id}`);
      selected.push({
        element,
        text: element.textContent ?? "",
        replacement: rule.replacement,
      });
    }
    receipt.push({
      id: rule.id,
      target: rule.target,
      expected_count: rule.expected_count,
      applied_count: elements.length,
    });
  }
  state.__cartularyMetadataOriginals = selected;
  let index = 0;
  for (const rule of rules)
    for (let i = 0; i < rule.expected_count; i++) {
      const selectedItem = selected[index++];
      if (!selectedItem) throw new Error("Invalid metadata selection");
      selectedItem.element.textContent = rule.replacement;
    }
  return receipt;
}

export function restoreMetadataDocument() {
  const state = window as NormalizationWindow;
  for (const original of state.__cartularyMetadataOriginals ?? []) {
    // A React refresh invalidates capture readiness; never insert detached nodes.
    if (original.element.isConnected)
      original.element.textContent = original.text;
  }
  delete state.__cartularyMetadataOriginals;
}

export async function applyVisualMetadata(page: Page, capture: string) {
  if (
    policy.schema_id !== "cartulary.frontend_visual_normalization.v1" ||
    !Object.hasOwn(policy.captures, capture)
  )
    throw new Error(`Undeclared visual capture ${capture}`);
  const inventory = await page.evaluate(() => {
    const counts: Record<string, number> = {};
    for (const element of document.querySelectorAll<HTMLElement>(
      "[data-generated-metadata]",
    )) {
      const surface = element.getAttribute("data-generated-metadata") ?? "";
      counts[surface] = (counts[surface] ?? 0) + 1;
    }
    return counts;
  });
  const rules = policy.captures[capture];
  if (!rules) throw new Error(`Missing metadata rules for ${capture}`);
  const applied = await page.evaluate(normalizeMetadataDocument, rules);
  return {
    schema_id: "cartulary.frontend_visual_normalization_receipt.v1",
    capture,
    capture_id: policy.bindings[capture]?.capture_id,
    policy_sha256: policyDigest,
    inventory,
    rules: applied,
  };
}

export async function restoreVisualMetadata(page: Page) {
  if (!page.isClosed()) await page.evaluate(restoreMetadataDocument);
}

export function assertMetadataCaptureIdentity(
  capture: string,
  captureId: string,
  goldenPath: string,
) {
  const binding = policy.bindings[capture];
  if (
    !binding ||
    binding.capture_id !== captureId ||
    binding.golden_path !== goldenPath
  )
    throw new Error(`Metadata capture identity mismatch for ${capture}`);
}

export async function verifyVisualMetadata(page: Page) {
  await page.evaluate(() => {
    const originals = (window as NormalizationWindow)
      .__cartularyMetadataOriginals;
    if (
      !originals ||
      originals.some(
        (item) =>
          !item.element.isConnected ||
          item.element.textContent !== item.replacement,
      )
    )
      throw new Error("Metadata presentation changed during capture");
  });
}
