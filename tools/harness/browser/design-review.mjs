import path from "node:path";
import { fileURLToPath } from "node:url";
import { runPreparedReview } from "./review-preparation.mjs";
export { holdReviewSession, reviewProfile, withReviewResources } from "./review-preparation.mjs";

export function runDesignReview(options = {}) {
  return runPreparedReview({ target: process.argv.includes("--smoke") ? "browser-design-review-smoke" : "browser-design-review", ...options });
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const controller = new AbortController();
  const stop = () => controller.abort();
  process.on("SIGINT", stop);
  process.on("SIGTERM", stop);
  try {
    if (process.argv.slice(2).some((arg) => arg !== "--smoke")) throw new Error("usage: design-review.mjs [--smoke]");
    await runDesignReview({ signal: controller.signal, ...(process.argv.includes("--smoke") ? { hold: async () => {}, verifySamples: true } : {}) });
  }
  catch (error) {
    process.stderr.write(`Browser review failed: ${error.message}\n`);
    process.exitCode = controller.signal.aborted ? 130 : 1;
  } finally {
    process.off("SIGINT", stop);
    process.off("SIGTERM", stop);
  }
}
