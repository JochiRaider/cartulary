import { createHash } from "node:crypto";
import { mkdirSync, rmSync } from "node:fs";
import path from "node:path";
import { artifact, loadBundle } from "./bundles.mjs";
import { primary } from "./analysis.mjs";
import { pngHeader } from "./png.mjs";
import { limits, ReviewFailure } from "./contract.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile } from "../../runtime/secure-local-files.mjs";

const version = "1";
const sha = (bytes) => createHash("sha256").update(bytes).digest("hex");
export const escapeHTML = (value) => String(value).replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;").replaceAll("'", "&#39;");
const script = `document.querySelectorAll('[data-zoom]').forEach(input=>input.addEventListener('input',()=>{const image=document.getElementById(input.dataset.zoom);image.style.width=(Number(image.dataset.width)*Number(input.value)/100)+'px';input.nextElementSibling.textContent=input.value+'%';}));document.querySelectorAll('[data-slider]').forEach(input=>input.addEventListener('input',()=>{document.getElementById(input.dataset.slider).style.clipPath='inset(0 '+(100-Number(input.value))+'% 0 0)';}));`;
const style = `body{font:16px system-ui;margin:24px;background:#f4f4f4;color:#181818}h1,h2,h3{line-height:1.2}p{max-width:85ch}pre{white-space:pre-wrap;overflow-wrap:anywhere;background:white;padding:16px}figure{margin:12px 0;padding:12px;background:white;min-width:0}img{display:block;max-width:none;height:auto}label{display:block;margin:8px 0}.viewport{overflow:auto;max-height:85vh;border:1px solid #888}.pair{display:grid;grid-template-columns:1fr 1fr;gap:16px}.slider{position:relative}.slider .after{position:absolute;inset:0 auto auto 0;clip-path:inset(0 50% 0 0)}a{color:#174aa5}.missing{border-left:4px solid #987b14;padding-left:12px}`;

export function renderReport(entry) {
  const { bundle, files } = entry, assets = new Map(), references = [];
  const image = (ref, label) => {
    const bytes = files.get(ref.path), { width, height } = pngHeader(bytes), name = `image-${references.length + 1}.png`;
    assets.set(name, bytes); references.push({ ref, name, width, height });
    const id = `image-${references.length}`;
    return `<figure><figcaption>${escapeHTML(label)} · ${width} × ${height} · <a href="${name}" download>Original PNG</a></figcaption><label>Zoom <input type="range" min="25" max="400" step="25" value="100" data-zoom="${id}"><output>100%</output></label><div class="viewport"><img id="${id}" src="${name}" alt="${escapeHTML(label)}" width="${width}" height="${height}" data-width="${width}"></div></figure>`;
  };
  const sections = [];
  for (const kind of ["original", "expected", "actual", "diff"]) sections.push(bundle.components[kind] ? image(bundle.components[kind], kind) : `<p class="missing">${kind}: unavailable</p>`);
  for (const derived of bundle.derived) sections.push(image(derived.ref, `${derived.kind}${derived.rectangle ? ` ${JSON.stringify(derived.rectangle)}` : ""}`));
  let pair;
  if (bundle.analysis?.comparison) pair = bundle.derived.find((entry) => entry.kind === "exact_diff")?.source_refs;
  else if (bundle.components.expected && (bundle.components.actual || bundle.components.original)) pair = [bundle.components.expected, primary(bundle)];
  if (pair?.length === 2) {
    const left = pngHeader(files.get(pair[0].path)), right = pngHeader(files.get(pair[1].path));
    sections.push(`<h2>Side-by-side comparison</h2><div class="pair">${image(pair[0], "Left input")}${image(pair[1], "Right input")}</div>`);
    if (left.width === right.width && left.height === right.height) {
      const [before, after] = references.slice(-2);
      sections.push(`<h3>Equal-size before/after slider</h3><p>Pixel alignment is not a claim of semantic equivalence.</p><label>Reveal right image <input type="range" min="0" max="100" value="50" data-slider="after"></label><div class="viewport"><div class="slider" style="width:${left.width}px;height:${left.height}px"><img src="${before.name}" alt="Left input" width="${left.width}" height="${left.height}"><img id="after" class="after" src="${after.name}" alt="Right input" width="${left.width}" height="${left.height}"></div></div>`);
    } else sections.push('<p class="missing">Slider unavailable: unequal dimensions.</p>');
  }
  const observed = bundle.components.observations ? JSON.parse(files.get(bundle.components.observations.path)) : null;
  const data = (title, value) => `<h2>${title}</h2>${value === null ? '<p class="missing">Unavailable</p>' : `<pre>${escapeHTML(JSON.stringify(value, null, 2))}</pre>`}`;
  let trace = '<p class="missing">Trace: unavailable</p>';
  if (bundle.components.trace) { assets.set("trace.zip", files.get(bundle.components.trace.path)); trace = '<p><a href="trace.zip" download>Local diagnostic trace</a> (not interpreted)</p>'; }
  assets.set("bundle.json", Buffer.from(`${JSON.stringify(bundle, null, 2)}\n`));
  const hash = createHash("sha256").update(script).digest("base64");
  const html = Buffer.from(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; script-src 'sha256-${hash}'; base-uri 'none'; form-action 'none'; connect-src 'none'"><title>Private UI review</title><style>${style}</style><h1>Private UI review</h1><p>Temporary local diagnostic. Links expire when this session stops, fails, or expires. Findings and zero differences do not establish product, accessibility, visual, release, or publication success.</p><p>Source class: <strong>${escapeHTML(bundle.source.kind)}</strong> · Bundle: ${escapeHTML(bundle.bundle_id)}</p><p><a href="bundle.json" download>Bundle metadata</a></p>${data("Source and capture identity", { source: bundle.source, binding: bundle.binding, parents: bundle.parents, observation: bundle.observation })}${data("Limitations", bundle.limitations)}${sections.join("")}${data("Image analysis", bundle.analysis)}${data("Geometry, fonts and accessibility snapshot", observed ? { elements: observed.elements, fonts: observed.fonts, accessibility_snapshot: observed.accessibility_snapshot } : null)}${data("Accessibility observations", observed?.axe ?? { status: "unavailable", scope: "main_document", violations: [], incomplete: [], unassessed_frames: null })}<p>Automated main-document observations exclude iframe contents and closed shadow roots. Empty findings do not establish keyboard or screen-reader usability.</p>${data("Console observations", observed?.console ?? null)}${data("Network observations", observed?.network ?? null)}${trace}<script>${script}</script></html>`);
  if (html.length > limits.report) throw new ReviewFailure("observation_limit");
  assets.set("index.html", html);
  return assets;
}

export async function execute(session, id) {
  const entry = await loadBundle(session, id), key = `${entry.sha256}-v${version}`;
  session.reports ??= new Map();
  const cached = session.reports.get(key);
  if (cached) {
    for (const ref of cached.files) {
      const bytes = readLocalFile(path.join(cached.root, ref.path), { maximum: ref.bytes });
      if (bytes.length !== ref.bytes || sha(bytes) !== ref.sha256) throw new ReviewFailure("invalid_artifact");
    }
    return { bundle_id: id, private_refs: [{ kind: "report", absolute_path: path.join(cached.root, "index.html") }] };
  }
  const files = renderReport(entry), release = session.reserveBytes([...files.values()].reduce((sum, bytes) => sum + bytes.length, 0));
  const root = session.runtime.privatePath("reports", key); let owned = false;
  try {
    privateDirectory(path.dirname(root)); mkdirSync(root, { mode: 0o700 }); owned = true;
    for (const [name, bytes] of files) { session.abort.signal.throwIfAborted(); atomicLocalFile(path.join(root, name), bytes); }
    session.reports.set(key, { root, files: [...files].map(([name, bytes]) => artifact(name, bytes, "application/octet-stream")) });
    return { bundle_id: id, private_refs: [{ kind: "report", absolute_path: path.join(root, "index.html") }] };
  } catch (error) { release(); if (owned) rmSync(root, { recursive: true, force: true }); throw error; }
}
