import { buildReceipt, verifyArchive } from "./cryptographic-module.mjs";
try {
  const [operation, binary] = process.argv.slice(2);
  if (operation === "verify" && !binary) verifyArchive(process.env.GO);
  else if (operation === "receipt" && binary) buildReceipt(process.env.GO, binary);
  else throw new Error("invalid cryptographic build invocation");
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 2;
}
