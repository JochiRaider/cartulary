import AxeBuilder from "@axe-core/playwright";
import { ReviewFailure, limits } from "./contract.mjs";

export const axeTags = Object.freeze(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "best-practice"]);
export function normalizeAxe(value, unassessedFrames) {
  try {
    if (value?.testEngine?.version !== "4.13.0" || !Number.isSafeInteger(unassessedFrames) || unassessedFrames < 0 || !Array.isArray(value.violations) || !Array.isArray(value.incomplete)) throw new Error("unsupported engine result");
    let occurrences = 0;
    const normalize = (entries) => entries.map((entry) => {
      if (typeof entry.id !== "string" || !entry.id || Buffer.byteLength(entry.id) > 1024 || ![null, "minor", "moderate", "serious", "critical"].includes(entry.impact) || !Array.isArray(entry.nodes) || !entry.nodes.length) throw new Error("malformed finding");
      occurrences += entry.nodes.length;
      if (occurrences > 1000) throw new ReviewFailure("observation_limit");
      const node_refs = entry.nodes.map((node) => {
        const validTarget = (target, depth = 0) => depth < 16 && Array.isArray(target) && target.length > 0 && target.length <= 64 && target.every((item) => typeof item === "string" ? item.length > 0 : validTarget(item, depth + 1));
        if (!validTarget(node.target)) throw new Error("malformed target");
        const reference = node.target.length === 1 && typeof node.target[0] === "string" ? node.target[0] : JSON.stringify(node.target);
        if (Buffer.byteLength(reference) > 4096) throw new ReviewFailure("observation_limit");
        return reference;
      });
      return { rule_id: entry.id, impact: entry.impact, node_refs };
    });
    const result = { status: "completed", engine_version: "4.13.0", scope: "main_document", violations: normalize(value.violations), incomplete: normalize(value.incomplete), unassessed_frames: unassessedFrames };
    if (Buffer.byteLength(JSON.stringify(result)) > limits.component) throw new ReviewFailure("observation_limit");
    return result;
  } catch (cause) { if (cause instanceof ReviewFailure) throw cause; throw new ReviewFailure("invalid_artifact", { cause }); }
}
export async function observeAxe(page) {
  let raw;
  try {
    // Explicitly exclude frame elements so the integration never recurses into
    // child browsing contexts, in addition to disabling engine traversal.
    raw = await new AxeBuilder({ page }).exclude("iframe").withTags([...axeTags]).options({ iframes: false, resultTypes: ["violations", "incomplete"] }).analyze();
  } catch (cause) {
    throw new ReviewFailure(cause.name === "TimeoutError" ? "operation_expired" : "analysis_failed", { cause });
  }
  return normalizeAxe(raw, await page.locator("iframe").count());
}
