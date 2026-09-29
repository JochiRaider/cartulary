import { artifact, bundleBase } from "./bundles.mjs";
import { digest, inputPath, readInput } from "./session-files.mjs";
import { limits, ReviewFailure } from "./contract.mjs";

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
export async function execute(operation, request, operationID) {
  const { decodePNG } = await import("./png.mjs");
  if (request.source === "page") {
    const { bundle, visible, inputNames } = structuredClone(operation.observation);
    const screenshot = operation.readStage("screenshot.png"), image = await decodePNG(screenshot);
    const { rectangle, transform } = scopeRectangle(request.scope, visible, bundle.observation.viewport, image);
    const sharp = (await import("sharp")).default;
    const png = request.scope.kind === "viewport" ? screenshot : await sharp(screenshot).extract({ left: rectangle.x, top: rectangle.y, width: rectangle.width, height: rectangle.height }).png().toBuffer();
    const files = new Map(inputNames.map((name) => [name, operation.readStage(name)])); files.set("original.png", png);
    bundle.components.original = artifact("original.png", png, "image/png");
    bundle.observation.image_dimensions = { width: rectangle.width, height: rectangle.height }; bundle.observation.coordinate_transform = transform;
    return operation.store.publishBundle(bundle, files);
  }
  if (request.source === "canonical_visual") { const { importCanonical } = await import("./canonical-import.mjs"); const imported = await importCanonical(request); return operation.store.publishBundle({ ...bundleBase(operation, operationID), ...imported.metadata }, imported.files); }
  const file = inputPath(request.path, ".png"), bytes = readInput(file, { extension: ".png", maximum: limits.png, privateFile: false });
  await decodePNG(bytes);
  const bundle = bundleBase(operation, operationID);
  bundle.source = { kind: "reference_image", import_ref: { input_path: file, input_sha256: digest(bytes), metadata: null } };
  bundle.components.original = artifact("original.png", bytes, "image/png");
  bundle.limitations = ["no_axe", "no_dom", "no_trace", "reference_only"];
  return operation.store.publishBundle(bundle, new Map([["original.png", bytes]]));
}
