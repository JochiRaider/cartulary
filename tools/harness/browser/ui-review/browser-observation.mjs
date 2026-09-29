import { artifact, bundleBase } from "./bundles.mjs";
import { jsonBytes } from "./session-files.mjs";
import { limits, ReviewFailure } from "./contract.mjs";
import { freeze } from "./immutable.mjs";

const stamp = () => new Date().toISOString();
const stableJSON = (value) => JSON.stringify(value, (_key, item) => typeof item === "number" ? Math.round(item * 64) / 64 : item);

export async function captureObservation(browser, request, context, operationID, stage) {
  const page = browser.page;
  const started_at = stamp(), started = performance.now(), epoch = browser.epoch, generation = browser.generation;
  const bound = context.source.binding;
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
      const fonts = context.source.fonts;
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
      const fonts = context.source.fonts;
      const ready = await page.evaluate((families) => families.every((family) => Array.from(document.fonts).some((face) => face.family.replaceAll('"', "") === family && face.status !== "error") && document.fonts.check(`400 12px "${family}"`)), fonts);
      if (!ready) throw new ReviewFailure("unstable_capture");
    }
    const observed = structuredClone(await browser.observe(entries.slice(0, request.targets.length)));
    if (request.include_axe) {
      const { observeAxe } = await import("./axe.mjs"); observed.axe = await observeAxe(page);
    } else observed.axe.status = "disabled";
    const screenshot = await page.screenshot({ type: "png", fullPage: false, animations: "allow", caret: "initial", scale: "device", timeout: Math.max(1, limits.operation - (performance.now() - started)) });
    const after = await page.evaluate(geometry, arguments_);
    if (stableJSON(after) !== previous || browser.epoch !== epoch || browser.generation !== generation) throw new ReviewFailure("unstable_capture");
    let visible;
    if (request.scope.kind === "element") {
      const index = entries.findIndex((entry) => JSON.stringify(entry.target) === JSON.stringify(request.scope.target));
      const [x, y, width, height] = prepared.rectangles[index].rect;
      const left = Math.max(0, x), top = Math.max(0, y), right = Math.min(prepared.viewport.width, x + width), bottom = Math.min(prepared.viewport.height, y + height);
      visible = right > left && bottom > top ? { x: left, y: top, width: right - left, height: bottom - top } : null;
    }
    stage("screenshot.png", screenshot);
    const bundle = bundleBase(context.identity, operationID), files = new Map([["observations.json", jsonBytes(observed)]]);
    bundle.components.observations = artifact("observations.json", files.get("observations.json"), "application/json");
    const attested = await context.source.attest();
    bundle.source = attested.source;
    if (attested.receipt) files.set("frontend-receipt.json", attested.receipt);
    bundle.binding = bound?.binding ?? null;
    bundle.observation = { epoch, started_at, finished_at: stamp(), duration_ms: Math.floor(performance.now() - started), viewport: prepared.viewport, image_dimensions: null, device_scale_factor: prepared.device_scale_factor, visual_viewport_scale: prepared.visual_viewport_scale, css_zoom: prepared.css_zoom, theme: prepared.theme, density: prepared.density, document_generation: generation, scope: request.scope, coordinate_transform: null };
    bundle.limitations = ["no_trace", "rendered_nodes_only", ...(bundle.source.kind === "live_unattested" ? ["live_unattested"] : []), ...(!request.include_axe ? ["no_axe"] : []), ...(observed.console.truncated ? ["truncated_console"] : []), ...(observed.network.truncated ? ["truncated_network"] : [])].sort();
    for (const [name, bytes] of files) stage(name, bytes);
    return freeze({ bundle, visible, inputNames: [...files.keys()] });
  } catch (cause) {
    if (browser.generation !== generation) throw new ReviewFailure("unstable_capture", { cause });
    if (cause.name === "TimeoutError") throw new ReviewFailure("operation_expired", { cause });
    throw cause;
  } finally { await Promise.allSettled([...temporary, documentNode, focused].filter(Boolean).map((handle) => handle.dispose())); }
}
