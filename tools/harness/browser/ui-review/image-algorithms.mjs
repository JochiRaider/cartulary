import sharp from "sharp";
import { ReviewFailure, limits } from "./contract.mjs";
import { decodePNG, encodePNG } from "./png.mjs";

export function fraction(count, total) {
  return Number((BigInt(count) * 2000000n + BigInt(total)) / (2n * BigInt(total))) / 1000000;
}
export function exactDifference(left, right) {
  if (left.width !== right.width || left.height !== right.height) throw new ReviewFailure("invalid_artifact");
  const data = Buffer.alloc(left.data.length); let different_pixels = 0;
  for (let offset = 0; offset < data.length; offset += 4) {
    if ([0, 1, 2, 3].some((channel) => left.data[offset + channel] !== right.data[offset + channel])) {
      different_pixels++; data[offset] = 255; data[offset + 2] = 255; data[offset + 3] = 255;
    }
  }
  const total_pixels = left.width * left.height;
  return { image: { width: left.width, height: left.height, data }, different_pixels, total_pixels, different_fraction: fraction(different_pixels, total_pixels) };
}
export function cropImage(image, rectangle) {
  const { x, y, width, height } = rectangle;
  if (![x, y, width, height].every(Number.isSafeInteger) || x < 0 || y < 0 || width < 1 || height < 1 || x + width > image.width || y + height > image.height) throw new ReviewFailure("invalid_artifact");
  const data = Buffer.alloc(width * height * 4);
  for (let row = 0; row < height; row++) image.data.copy(data, row * width * 4, ((y + row) * image.width + x) * 4, ((y + row) * image.width + x + width) * 4);
  return { width, height, data };
}
export function overlayImage(image, rectangles, transform) {
  if (!transform || !rectangles.length) throw new ReviewFailure("invalid_artifact");
  const data = Buffer.from(image.data); let drawn = 0;
  for (const rect of rectangles) {
    const left = Math.max(0, Math.floor((rect.x - transform.origin_x) * transform.scale_x));
    const top = Math.max(0, Math.floor((rect.y - transform.origin_y) * transform.scale_y));
    const right = Math.min(image.width, Math.ceil((rect.x + rect.width - transform.origin_x) * transform.scale_x));
    const bottom = Math.min(image.height, Math.ceil((rect.y + rect.height - transform.origin_y) * transform.scale_y));
    if (right <= left || bottom <= top) continue;
    drawn++;
    for (let y = top; y < bottom; y++) for (let x = left; x < right; x++) {
      if (x >= left + 2 && x < right - 2 && y >= top + 2 && y < bottom - 2) continue;
      const offset = (y * image.width + x) * 4; data[offset] = 255; data[offset + 1] = 0; data[offset + 2] = 255; data[offset + 3] = 255;
    }
  }
  if (!drawn) throw new ReviewFailure("invalid_artifact");
  return { width: image.width, height: image.height, data };
}
export async function contactSheet(images) {
  if (!images.length || images.length > 19) throw new ReviewFailure("observation_limit");
  const width = 4 * 320 + 3 * 8, rows = Math.ceil(images.length / 4), height = rows * 240 + (rows - 1) * 8;
  const data = Buffer.alloc(width * height * 4, 255);
  for (let index = 0; index < images.length; index++) {
    const image = await decodePNG(images[index]);
    const resized = await sharp(image.data, { raw: { width: image.width, height: image.height, channels: 4 } })
      .resize({ width: 320, height: 240, fit: "inside", withoutEnlargement: true, kernel: "lanczos3" })
      .flatten({ background: { r: 255, g: 255, b: 255 } }).ensureAlpha().raw().toBuffer({ resolveWithObject: true });
    const x = index % 4 * 328 + Math.floor((320 - resized.info.width) / 2), y = Math.floor(index / 4) * 248 + Math.floor((240 - resized.info.height) / 2);
    for (let row = 0; row < resized.info.height; row++) resized.data.copy(data, ((y + row) * width + x) * 4, row * resized.info.width * 4, (row + 1) * resized.info.width * 4);
  }
  return { width, height, data };
}
export async function computeImages(job) {
  if (!Array.isArray(job.operations) || job.primary.length > limits.png || (job.secondary?.length ?? 0) > limits.png) throw new ReviewFailure("invalid_artifact");
  const image = await decodePNG(job.primary), outputs = [], cropped = [];
  if (job.operations.includes("crop")) for (const rectangle of job.crops) {
    const bytes = await encodePNG(cropImage(image, rectangle)); cropped.push(bytes); outputs.push({ kind: "crop", rectangle, bytes });
  }
  if (job.operations.includes("overlay")) outputs.push({ kind: "overlay", rectangle: null, bytes: await encodePNG(overlayImage(image, job.rectangles, job.transform)) });
  let comparison = null, diff;
  if (job.operations.includes("exact_diff")) {
    const computed = exactDifference(image, await decodePNG(job.secondary));
    diff = await encodePNG(computed.image); outputs.push({ kind: "exact_diff", rectangle: null, bytes: diff });
    comparison = { width: image.width, height: image.height, different_pixels: computed.different_pixels, total_pixels: computed.total_pixels, different_fraction: computed.different_fraction };
  }
  if (job.operations.includes("contact_sheet")) {
    const inputs = [...job.contactInputs, ...cropped];
    // A requested exact diff is a newly derived diagnostic, not the source
    // bundle's optional diff channel. The specified sheet uses source channels.
    outputs.push({ kind: "contact_sheet", rectangle: null, bytes: await encodePNG(await contactSheet(inputs)) });
  }
  return { outputs, comparison };
}
