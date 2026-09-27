import path from "node:path";
import sharp from "sharp";
import { artifact, bundleBase, publishBundle } from "./bundles.mjs";
import { digest, inputPath, jsonBytes, readInput } from "./session-files.mjs";
import { limits, ReviewFailure } from "./contract.mjs";
import { decodePNG } from "./png.mjs";
import { pageBinding, readJSON, containedFile } from "./source.mjs";
import { repoRoot } from "./toolchain.mjs";

const stamp = () => new Date().toISOString();
const stableJSON = (value) => JSON.stringify(value, (_key, item) => typeof item === "number" ? Math.round(item * 64) / 64 : item);
export function scopeRectangle(scope, visible, viewport, image) {
  const scale_x = image.width / viewport.width, scale_y = image.height / viewport.height;
  let rectangle;
  if (scope.kind === "viewport") rectangle = { x: 0, y: 0, width: image.width, height: image.height };
  else if (scope.kind === "region") rectangle = { x: scope.x, y: scope.y, width: scope.width, height: scope.height };
  else {
    if (!visible) throw new ReviewFailure("target_unavailable");
    const x = Math.max(0, Math.floor(visible.x * scale_x)), y = Math.max(0, Math.floor(visible.y * scale_y));
    rectangle = { x, y, width: Math.min(image.width, Math.ceil((visible.x + visible.width) * scale_x)) - x, height: Math.min(image.height, Math.ceil((visible.y + visible.height) * scale_y)) - y };
  }
  if (rectangle.width <= 0 || rectangle.height <= 0 || rectangle.x < 0 || rectangle.y < 0 || rectangle.x + rectangle.width > image.width || rectangle.y + rectangle.height > image.height) throw new ReviewFailure("invalid_request");
  return { rectangle, transform: { origin_x: rectangle.x / scale_x, origin_y: rectangle.y / scale_y, scale_x, scale_y } };
}
async function pageCapture(session, request, operationID) {
  const browser = session.browser, page = browser.page;
  const started_at = stamp(), started = performance.now(), epoch = browser.epoch, generation = browser.generation;
  const bound = pageBinding(session, request.binding);
  const targets = [...request.targets];
  if (request.scope.kind === "element" && !targets.some((target) => JSON.stringify(target) === JSON.stringify(request.scope.target))) targets.push(request.scope.target);
  const entries = [], temporary = [];
  let documentNode, focused;
  try {
    for (const target of targets) {
      const handle = await browser.resolve(target); entries.push({ target, handle });
      if (target.kind !== "element_ref") temporary.push(handle);
    }
    if (bound) {
      const fonts = readJSON(path.join(repoRoot, "apps/web/public/assets/fonts/FONT_MANIFEST.json")).value.families.filter((font) => font.active_by_default).map((font) => font.family);
      await page.evaluate((families) => Promise.all(families.map((family) => document.fonts.load(`400 12px "${family}"`))), fonts);
    }
    await page.evaluate(() => document.fonts.ready);
    documentNode = await page.evaluateHandle(() => document);
    focused = await page.evaluateHandle(() => ({ node: document.activeElement, generation: 0, scrollIDs: new WeakMap(), nextScrollID: 0 }));
    const arguments_ = { nodes: entries.map((entry) => entry.handle), documentNode, focused };
    // The same function is used for the three-frame preparation and the
    // immediate post-screenshot check; it performs observation only.
    const geometry = ({ nodes, documentNode: original, focused: state }) => {
      const rect = (node) => { const r = node.getBoundingClientRect(); return [r.x, r.y, r.width, r.height]; };
      const scrolling = new Set([document.scrollingElement]);
      for (const node of nodes) for (let ancestor = node; ancestor; ancestor = ancestor.parentElement) if (ancestor.scrollHeight > ancestor.clientHeight || ancestor.scrollWidth > ancestor.clientWidth) scrolling.add(ancestor);
      const zoom = Number.parseFloat(getComputedStyle(document.documentElement).zoom);
      if (state.node !== document.activeElement) { state.node = document.activeElement; state.generation++; }
      const scroll = [...scrolling].filter(Boolean).map((node) => {
        if (!state.scrollIDs.has(node)) state.scrollIDs.set(node, ++state.nextScrollID);
        return [state.scrollIDs.get(node), node.scrollLeft, node.scrollTop, node.clientWidth, node.clientHeight, node.scrollWidth, node.scrollHeight];
      });
      return { url: location.href, same_document: document === original, focus_generation: state.generation, viewport: { width: innerWidth, height: innerHeight }, device_scale_factor: devicePixelRatio, visual_viewport_scale: visualViewport?.scale ?? 1, css_zoom: Number.isFinite(zoom) && zoom > 0 ? zoom : null, theme: document.querySelector("[data-cartulary-theme]")?.getAttribute("data-cartulary-theme") ?? null, density: document.querySelector("[data-cartulary-density]")?.getAttribute("data-cartulary-density") ?? null, rectangles: nodes.map((node) => ({ connected: node.isConnected, rect: rect(node) })), scroll };
    };
    let previous, consecutive = 0, prepared;
    while (consecutive < 3) {
      if (performance.now() - started >= limits.operation - 100) throw new ReviewFailure("operation_expired");
      await page.evaluate(() => new Promise(requestAnimationFrame));
      prepared = await page.evaluate(geometry, arguments_);
      if (!prepared.same_document || prepared.rectangles.some((entry) => !entry.connected) || browser.epoch !== epoch || browser.generation !== generation) throw new ReviewFailure("unstable_capture");
      const encoded = stableJSON(prepared); consecutive = encoded === previous ? consecutive + 1 : 1; previous = encoded;
    }
    if (bound?.profile) {
      const profile = bound.profile;
      if (`${prepared.viewport.width}x${prepared.viewport.height}` !== profile.viewport_css_px || prepared.device_scale_factor !== profile.device_scale_factor || prepared.css_zoom * 100 !== profile.browser_zoom_percent || prepared.theme !== profile.theme_id || (profile.density_id !== null && prepared.density !== profile.density_id)) throw new ReviewFailure("unstable_capture");
    }
    if (bound) {
      const fonts = readJSON(path.join(repoRoot, "apps/web/public/assets/fonts/FONT_MANIFEST.json")).value.families.filter((font) => font.active_by_default).map((font) => font.family);
      const ready = await page.evaluate((families) => families.every((family) => Array.from(document.fonts).some((face) => face.family.replaceAll('"', "") === family && face.status !== "error") && document.fonts.check(`400 12px "${family}"`)), fonts);
      if (!ready) throw new ReviewFailure("unstable_capture");
    }
    const observed = await browser.observe(entries.slice(0, request.targets.length));
    if (request.include_axe) {
      const { observeAxe } = await import("./axe.mjs"); observed.axe = await observeAxe(page);
    } else observed.axe.status = "disabled";
    const screenshot = await page.screenshot({ type: "png", fullPage: false, animations: "allow", caret: "initial", scale: "device", timeout: Math.max(1, limits.operation - (performance.now() - started)) });
    const after = await page.evaluate(geometry, arguments_);
    if (stableJSON(after) !== previous || browser.epoch !== epoch || browser.generation !== generation) throw new ReviewFailure("unstable_capture");
    const image = await decodePNG(screenshot);
    let visible;
    if (request.scope.kind === "element") {
      const index = entries.findIndex((entry) => JSON.stringify(entry.target) === JSON.stringify(request.scope.target));
      const [x, y, width, height] = prepared.rectangles[index].rect;
      const left = Math.max(0, x), top = Math.max(0, y), right = Math.min(prepared.viewport.width, x + width), bottom = Math.min(prepared.viewport.height, y + height);
      visible = right > left && bottom > top ? { x: left, y: top, width: right - left, height: bottom - top } : null;
    }
    const { rectangle, transform } = scopeRectangle(request.scope, visible, prepared.viewport, image);
    const png = request.scope.kind === "viewport" ? screenshot : await sharp(screenshot).extract({ left: rectangle.x, top: rectangle.y, width: rectangle.width, height: rectangle.height }).png().toBuffer();
    const bundle = bundleBase(session, operationID), files = new Map([["original.png", png], ["observations.json", jsonBytes(observed)]]);
    bundle.components.original = artifact("original.png", png, "image/png");
    bundle.components.observations = artifact("observations.json", files.get("observations.json"), "application/json");
    if (session.mode === "seeded") {
      await session.seeded.check();
      const stack = readJSON(session.seeded.attached.CARTULARY_WEB_E2E_STACK_JSON_FILE, "cartulary.web_e2e_stack.v7").value;
      const receipt = readJSON(containedFile(session.seeded.runRoot, stack.frontend.build_artifact_ref), "cartulary.frontend_build_artifact.v1", stack.frontend.build_receipt_sha256);
      if (receipt.value.run_id !== session.runID || receipt.value.source_digest.replace(/^sha256:/u, "") !== session.workspaceDigest) throw new ReviewFailure("invalid_artifact");
      files.set("frontend-receipt.json", receipt.bytes);
      bundle.source = { kind: "sealed_review", workspace_digest: session.workspaceDigest, served_source_digest: receipt.value.source_digest.slice(7), frontend_receipt: artifact("frontend-receipt.json", receipt.bytes, "application/json"), browser_version: browser.browser.version(), runtime_profile_id: session.input.REVIEW_PROFILE };
    } else bundle.source = { kind: "live_unattested", workspace_digest: session.workspaceDigest, browser_version: browser.browser.version() };
    bundle.binding = bound?.binding ?? null;
    bundle.observation = { epoch, started_at, finished_at: stamp(), duration_ms: Math.floor(performance.now() - started), viewport: prepared.viewport, image_dimensions: { width: rectangle.width, height: rectangle.height }, device_scale_factor: prepared.device_scale_factor, visual_viewport_scale: prepared.visual_viewport_scale, css_zoom: prepared.css_zoom, theme: prepared.theme, density: prepared.density, document_generation: generation, scope: request.scope, coordinate_transform: transform };
    bundle.limitations = ["no_trace", "rendered_nodes_only", ...(session.mode === "dev" ? ["live_unattested"] : []), ...(!request.include_axe ? ["no_axe"] : []), ...(observed.console.truncated ? ["truncated_console"] : []), ...(observed.network.truncated ? ["truncated_network"] : [])].sort();
    return publishBundle(session, bundle, files);
  } catch (cause) {
    if (browser.generation !== generation) throw new ReviewFailure("unstable_capture", { cause });
    if (cause.name === "TimeoutError") throw new ReviewFailure("operation_expired", { cause });
    throw cause;
  } finally { await Promise.allSettled([...temporary, documentNode, focused].filter(Boolean).map((handle) => handle.dispose())); }
}
export async function execute(session, request, operationID) {
  if (request.source === "page") return pageCapture(session, request, operationID);
  if (request.source === "canonical_visual") { const { importCanonical } = await import("./canonical-import.mjs"); const imported = await importCanonical(request); return publishBundle(session, { ...bundleBase(session, operationID), ...imported.metadata }, imported.files); }
  const file = inputPath(request.path, ".png"), bytes = readInput(file, { extension: ".png", maximum: limits.png, privateFile: false });
  await decodePNG(bytes);
  const bundle = bundleBase(session, operationID);
  bundle.source = { kind: "reference_image", import_ref: { input_path: file, input_sha256: digest(bytes), metadata: null } };
  bundle.components.original = artifact("original.png", bytes, "image/png");
  bundle.limitations = ["no_axe", "no_dom", "no_trace", "reference_only"];
  return publishBundle(session, bundle, new Map([["original.png", bytes]]));
}
