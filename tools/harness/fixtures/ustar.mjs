// Bounded USTAR fixture encoding shared by synthetic archive producers.
export function fixtureTar(entries) {
  const chunks = [];
  for (const [name, bytes] of entries) {
    if (Buffer.byteLength(name) >= 100) throw new Error("Review TAR path exceeds header capacity");
    const header = Buffer.alloc(512);
    header.write(name, 0, 100);
    const octal = (value, offset, width) => header.write(`${value.toString(8).padStart(width - 1, "0")}\0`, offset, width);
    octal(0o644, 100, 8); octal(0, 108, 8); octal(0, 116, 8);
    octal(bytes.length, 124, 12); octal(0, 136, 12);
    header.fill(32, 148, 156); header.write("0", 156);
    header.write("ustar\0", 257); header.write("00", 263);
    header.write(`${header.reduce((sum, byte) => sum + byte, 0).toString(8).padStart(6, "0")}\0 `, 148, 8);
    chunks.push(header, bytes, Buffer.alloc((512 - bytes.length % 512) % 512));
  }
  return Buffer.concat([...chunks, Buffer.alloc(1024)]);
}
