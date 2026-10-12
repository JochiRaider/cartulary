import { stopOwnedProcess as stopProcess } from "../../runtime/owned-process.mjs";
import { ReviewFailure } from "./failure.mjs";

export { ownedProcess } from "../../runtime/owned-process.mjs";
export async function stopOwnedProcess(proof, options) {
  try { await stopProcess(proof, options); }
  catch (cause) { throw new ReviewFailure("cleanup_failed", { cause }); }
}

export async function boundedCleanup(release, milliseconds = 30000) {
  let timer;
  try {
    return await Promise.race([Promise.resolve().then(release), new Promise((_, reject) => { timer = setTimeout(() => reject(new ReviewFailure("cleanup_failed")), milliseconds); })]);
  } finally { clearTimeout(timer); }
}
