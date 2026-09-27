import { chromium } from "playwright";
import { randomBytes } from "node:crypto";
import { accessSync, constants } from "node:fs";
import { ReviewFailure, limits, schemaID, validate } from "./contract.mjs";
import { reviewTotp } from "../design-review-seed.mjs";

export function browserReady() {
  try { accessSync(chromium.executablePath(), constants.X_OK); }
  catch (cause) { throw new ReviewFailure("tool_configuration", { cause }); }
}
function boundedText(value, maximum = 4096) {
  const bytes = Buffer.from(value);
  if (bytes.length <= maximum) return { value, truncated: false };
  let end = maximum;
  while (end > 0 && (bytes[end] & 0xc0) === 0x80) end--;
  return { value: bytes.subarray(0, end).toString("utf8"), truncated: true };
}
export const unavailableAxe = (status = "unavailable") => ({ status, engine_version: null, scope: "main_document", violations: [], incomplete: [], unassessed_frames: null });
export class ReviewBrowser {
  constructor({ origin, mode, actors = {}, onLost = () => {} }) {
    this.origin = origin; this.mode = mode; this.actors = actors; this.onLost = onLost;
    this.epoch = 0; this.generation = 0; this.references = new Map(); this.sequence = 0; this.needsSnapshot = false; this.closing = false;
    this.channels = { console: { records: [], truncated: false }, network: { records: [], truncated: false } };
  }
  async start() {
    browserReady();
    try { this.browser = await chromium.launch({ headless: true, timeout: 30000, env: { PATH: "/usr/bin:/bin", HOME: process.env.HOME ?? "/tmp", LANG: "en_US.UTF-8" } }); }
    catch (cause) { throw new ReviewFailure("startup_failed", { cause }); }
    this.browser.on("disconnected", () => { if (!this.closing) this.onLost(new ReviewFailure("session_lost")); });
    await this.newContext();
    await this.page.goto(this.origin, { waitUntil: "domcontentloaded", timeout: 30000 });
    this.checkFault();
    return this;
  }
  async newContext() {
    await this.context?.close(); this.references.clear(); this.fault = null;
    this.context = await this.browser.newContext({ viewport: { width: 1440, height: 900 }, deviceScaleFactor: 1, locale: "en-US", colorScheme: "light", acceptDownloads: false, serviceWorkers: "block" });
    await this.context.route("**/*", async (route) => {
      const request = route.request();
      if (request.isNavigationRequest() && new URL(request.url()).origin !== this.origin) { this.fault = new ReviewFailure("navigation_boundary"); await route.abort(); }
      else await route.continue();
    });
    this.page = await this.context.newPage(); this.page.setDefaultTimeout(limits.action);
    // Playwright routes only the first URL of a server redirect chain. Chromium's
    // request-stage interception applies the same origin boundary to every hop.
    // This stays private to the pinned Playwright adapter.
    const navigation = await this.context.newCDPSession(this.page);
    await navigation.send("Fetch.enable", { patterns: [{ resourceType: "Document", requestStage: "Request" }] });
    navigation.on("Fetch.requestPaused", async ({ requestId, request }) => {
      try {
        if (new URL(request.url).origin !== this.origin) {
          this.fault = new ReviewFailure("navigation_boundary");
          await navigation.send("Fetch.failRequest", { requestId, errorReason: "BlockedByClient" });
        } else await navigation.send("Fetch.continueRequest", { requestId });
      } catch { if (!this.closing) this.fault ??= new ReviewFailure("target_unavailable"); }
    });
    this.page.on("popup", (popup) => { this.fault = new ReviewFailure("target_unavailable"); void popup.close(); });
    this.page.on("dialog", () => { this.fault = new ReviewFailure("target_unavailable"); });
    this.page.on("crash", () => this.onLost(new ReviewFailure("session_lost")));
    this.page.on("framenavigated", (frame) => {
      if (frame !== this.page.mainFrame()) return;
      this.generation++; this.references.clear();
      this.channels = { console: { records: [], truncated: false }, network: { records: [], truncated: false } };
      if (frame.url() !== "about:blank" && new URL(frame.url()).origin !== this.origin) { this.fault = new ReviewFailure("navigation_boundary"); this.onLost(this.fault); }
    });
    this.page.on("console", (message) => {
      const types = { error: "error", warning: "warning", info: "info", debug: "debug", log: "info" };
      const text = boundedText(message.text());
      this.record("console", { level: types[message.type()] ?? "debug", text: text.value }, text.truncated);
    });
    this.page.on("pageerror", (error) => { const text = boundedText(error.message); this.record("console", { level: "error", text: text.value }, text.truncated); });
    this.page.on("response", (response) => { const url = boundedText(response.url()); this.record("network", { method: response.request().method(), url: url.value, status: response.status(), outcome: "response" }, url.truncated); });
    this.page.on("requestfailed", (request) => { const url = boundedText(request.url()); this.record("network", { method: request.method(), url: url.value, status: null, outcome: "failed" }, url.truncated); });
  }
  record(channel, record, truncated) {
    const collection = this.channels[channel]; collection.records.push({ sequence: ++this.sequence, ...record });
    if (collection.records.length > 200) { collection.records.shift(); collection.truncated = true; }
    collection.truncated ||= truncated;
  }
  checkFault() { if (this.fault) { const error = this.fault; this.fault = null; throw error; } }
  validateAction(request) {
    if (request.expected_epoch !== this.epoch || (request.parameters.target?.kind === "element_ref" && request.parameters.target.epoch !== this.epoch)) throw new ReviewFailure("session_mismatch");
    if (this.needsSnapshot && request.action !== "snapshot") throw new ReviewFailure("session_mismatch");
    if (request.action === "authenticate" && this.mode !== "seeded") throw new ReviewFailure("invalid_request");
    if (request.parameters.target?.kind === "element_ref") {
      const ref = this.references.get(request.parameters.target.value);
      if (!ref || ref.epoch !== this.epoch || ref.generation !== this.generation) throw new ReviewFailure("session_mismatch");
    }
  }
  async resolve(target) {
    if (target.kind === "element_ref") {
      const ref = this.references.get(target.value);
      if (target.epoch !== this.epoch || !ref || ref.epoch !== this.epoch || ref.generation !== this.generation) throw new ReviewFailure("session_mismatch");
      if (!await ref.handle.evaluate((node) => node.isConnected)) throw new ReviewFailure("target_unavailable");
      return ref.handle;
    }
    const locator = target.kind === "test_id" ? this.page.getByTestId(target.value) : this.page.getByRole(target.role, { name: target.name, exact: true });
    if (await locator.count() !== 1) throw new ReviewFailure("target_unavailable");
    const handle = await locator.elementHandle();
    if (!handle || !await handle.evaluate((node) => node.isConnected)) throw new ReviewFailure("target_unavailable");
    return handle;
  }
  async action(request) {
    this.validateAction(request);
    const { action, parameters } = request;
    let target;
    try { target = parameters.target ? await this.resolve(parameters.target) : null; }
    finally { if (action !== "snapshot") { this.epoch++; await this.clearReferences(target); } }
    try {
      if (target && (!await target.isVisible() || !await target.isEnabled())) throw new ReviewFailure("target_unavailable");
      switch (action) {
        case "snapshot": return await this.snapshot();
        case "navigate": await this.page.goto(new URL(parameters.path, this.origin).href, { timeout: limits.operation, waitUntil: "domcontentloaded" }); break;
        case "click": await target.click({ timeout: limits.action }); break;
        case "fill": await target.fill(parameters.text, { timeout: limits.action }); break;
        case "select": await target.selectOption(parameters.values.map((value) => ({ value })), { timeout: limits.action }); break;
        case "focus": await target.focus(); if (!await target.evaluate((node) => document.activeElement === node)) throw new ReviewFailure("target_unavailable"); break;
        case "press": if (target) await target.press(parameters.key, { timeout: limits.action }); else await this.page.keyboard.press(parameters.key); break;
        case "scroll": if (target) await target.evaluate((node, position) => { node.scrollLeft = position.x; node.scrollTop = position.y; }, parameters); else await this.page.evaluate(({ x, y }) => window.scrollTo({ left: x, top: y, behavior: "instant" }), parameters); break;
        case "resize": await this.page.setViewportSize({ width: parameters.width, height: parameters.height }); break;
        case "authenticate": await this.authenticate(parameters.actor); break;
        default: throw new ReviewFailure("invalid_request");
      }
      this.checkFault();
      if (!target && ["scroll", "press"].includes(action)) target = (await this.page.evaluateHandle((kind) => kind === "scroll" ? document.scrollingElement : document.activeElement, action)).asElement();
      return await this.observe(target ? [{ target: parameters.target, handle: target }] : []);
    } catch (cause) {
      this.checkFault();
      if (cause instanceof ReviewFailure) throw cause;
      if (cause.name === "TimeoutError") { this.needsSnapshot = true; throw new ReviewFailure("operation_expired", { cause }); }
      throw new ReviewFailure("target_unavailable", { cause });
    } finally { if (action !== "snapshot") await target?.dispose(); }
  }
  async authenticate(actor) {
    const account = this.actors[actor]; if (!account) throw new ReviewFailure("invalid_request");
    await this.newContext();
    // The normal authentication endpoint enforces the same password/MFA policy;
    // the isolated context receives its cookies, never another actor's storage.
    const body = { username: account.email, password: account.password, ...(account.totp_setup_key ? { second_factor: { kind: "totp", assertion: { code: reviewTotp(account.totp_setup_key) } } } : {}) };
    const response = await this.context.request.post(`${this.origin}/api/v1/auth/login`, { data: body, timeout: limits.action, maxRedirects: 0 });
    if (!response.ok()) throw new ReviewFailure("target_unavailable");
    await this.page.goto(this.origin, { timeout: limits.action, waitUntil: "domcontentloaded" });
  }
  async snapshot() {
    await this.clearReferences();
    const selected = await this.page.evaluateHandle(() => {
      const result = [];
      for (const node of document.querySelectorAll('button,a[href],input,select,textarea,[role],h1,h2,h3,h4,h5,h6,[tabindex]')) {
        if (result.length === 64) break;
        const rect = node.getBoundingClientRect();
        if (rect.width > 0 && rect.height > 0 && getComputedStyle(node).visibility !== "hidden") result.push(node);
      }
      return result;
    });
    const properties = await selected.getProperties(); const observed = [...properties.values()].map((handle) => ({ handle: handle.asElement() })).filter((entry) => entry.handle);
    await selected.dispose();
    const value = await this.observe(observed, { allocate: true }); this.needsSnapshot = false; return value;
  }
  async clearReferences(except) {
    const previous = [...this.references.values()]; this.references.clear();
    await Promise.all(previous.filter((entry) => entry.handle !== except).map((entry) => entry.handle.dispose()));
  }
  async observe(entries = [], { allocate = false } = {}) {
    const elements = [];
    for (const entry of entries) {
      const id = `element-${randomBytes(16).toString("hex")}`;
      if (allocate) this.references.set(id, { handle: entry.handle, epoch: this.epoch, generation: this.generation });
      const target = entry.target ?? { kind: "element_ref", value: id, epoch: this.epoch };
      const detail = await entry.handle.evaluate((node) => {
        const rect = node.getBoundingClientRect(); const style = getComputedStyle(node);
        const x = Math.max(0, rect.x), y = Math.max(0, rect.y), right = Math.min(innerWidth, rect.right), bottom = Math.min(innerHeight, rect.bottom);
        const names = ["font-family", "font-size", "font-weight", "line-height", "color", "background-color", "padding", "gap", "overflow-x", "overflow-y"];
        return { role: node.getAttribute("role"), name: node.getAttribute("aria-label"), text: node.textContent, rect: { x: rect.x, y: rect.y, width: rect.width, height: rect.height }, visible_rect: right > x && bottom > y ? { x, y, width: right - x, height: bottom - y } : null, focused: document.activeElement === node, disabled: node.matches(":disabled") || node.getAttribute("aria-disabled") === "true", scroll: { left: node.scrollLeft, top: node.scrollTop, client_width: node.clientWidth, client_height: node.clientHeight, scroll_width: node.scrollWidth, scroll_height: node.scrollHeight }, style: Object.fromEntries(names.map((name) => [name.replaceAll("-", "_"), style.getPropertyValue(name)])), overflow_candidate: node.scrollWidth > node.clientWidth + 1 || node.scrollHeight > node.clientHeight + 1 };
      });
      const text = detail.text === null ? null : boundedText(detail.text);
      elements.push({ target, resolved_ref: id, ...detail, text: text?.value ?? null, text_truncated: text?.truncated ?? false });
    }
    const accessibility_snapshot = await this.page.locator("body").ariaSnapshot({ timeout: limits.action });
    if (Buffer.byteLength(accessibility_snapshot) > limits.snapshot) throw new ReviewFailure("observation_limit");
    const fonts = await this.page.evaluate(() => Array.from(document.fonts, ({ family, style, weight, status }) => ({ family, style, weight, status })));
    if (fonts.length > 128 || fonts.some((font) => Object.values(font).some((value) => Buffer.byteLength(value) > 1024))) throw new ReviewFailure("observation_limit");
    const axe = { ...unavailableAxe("disabled"), unassessed_frames: await this.page.locator("iframe").count() };
    return validate("observations", { schema_id: schemaID("observations"), elements, fonts, accessibility_snapshot, axe, ...structuredClone(this.channels) });
  }
  async close() { this.closing = true; await this.browser?.close(); }
}
