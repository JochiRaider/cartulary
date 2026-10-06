// A renderer proof is published before Docker acquisition, so recovery can
// identify a container even when creation succeeded but its reply was lost.
export function validateRendererProof(proof) {
  if (!proof || Object.keys(proof).sort().join(",") !== "daemon_id,image_id,name,token" ||
      Object.values(proof).some((value) => typeof value !== "string") ||
      !/^[A-Za-z0-9:._-]{1,128}$/u.test(proof.daemon_id) ||
      !/^sha256:[a-f0-9]{64}$/u.test(proof.image_id) ||
      !/^[a-f0-9]{32}$/u.test(proof.token) ||
      !/^cartulary-visual-[1-9][0-9]*-[a-f0-9]{12}$/u.test(proof.name) ||
      !proof.name.endsWith(`-${proof.token.slice(0, 12)}`)) throw new Error("unsafe renderer ownership proof");
  return proof;
}

export function removeOwnedRenderer(proof, run) {
  validateRendererProof(proof);
  if (run(["info", "--format", "{{.ID}}"]) !== proof.daemon_id) throw new Error("renderer daemon identity mismatch");
  const query = ["ps", "--all", "--no-trunc", "--filter", `name=^/${proof.name}$`, "--filter", `label=cartulary.visual-owner=${proof.token}`, "--format", "{{.ID}}"];
  const id = run(query);
  if (!id) return;
  if (!/^[a-f0-9]{64}$/u.test(id)) throw new Error("ambiguous renderer ownership");
  const values = JSON.parse(run(["inspect", id]));
  const value = values[0];
  if (values.length !== 1 || value.Id !== id || value.Name !== `/${proof.name}` ||
      value.Image !== proof.image_id || value.Config?.Labels?.["cartulary.visual-owner"] !== proof.token) throw new Error("renderer ownership mismatch");
  let removalError;
  try { run(["rm", "--force", id]); } catch (error) { removalError = error; }
  if (run(query)) throw removalError ?? new Error("renderer removal not confirmed");
}
