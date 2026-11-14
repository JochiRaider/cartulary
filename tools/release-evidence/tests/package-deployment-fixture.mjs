import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
export const borrowedDeploymentID = "f".repeat(64);
export function createDeploymentFixture(directory, scenario) {
 fs.mkdirSync(directory, { recursive: true, mode: 0o700 });
 const bin = path.join(directory, "bin"); fs.mkdirSync(bin);
 const state = path.join(directory, "state.json"), control = path.join(directory, "control.json"), log = path.join(directory, "calls.jsonl");
 fs.writeFileSync(state, JSON.stringify({ container: { [borrowedDeploymentID]: { Id: borrowedDeploymentID, Name: "/borrowed", Config: { Labels: { "com.docker.compose.project": "borrowed" } } } },
  volume: { borrowed_volume: { Name: "borrowed_volume", Labels: { "com.docker.compose.project": "borrowed" } } },
  network: { [borrowedDeploymentID]: { Id: borrowedDeploymentID, Name: "borrowed", Labels: { "com.docker.compose.project": "borrowed" } } } }));
 fs.writeFileSync(control, JSON.stringify({ scenario })); fs.writeFileSync(log, "");
 fs.writeFileSync(path.join(bin, "docker"), `#!${process.execPath}\nimport { runDeploymentFixture } from ${JSON.stringify(import.meta.url)};\nrunDeploymentFixture();\n`, { mode: 0o700 });
 return { state, control, log, env: { PATH: `${bin}:${process.env.PATH}`, PACKAGE_DEPLOYMENT_STATE: state, PACKAGE_DEPLOYMENT_CONTROL: control, PACKAGE_DEPLOYMENT_LOG: log } };
}
export function populateDeploymentFixture(file, allocation) {
 const state = JSON.parse(fs.readFileSync(file)); state.projects = [allocation.sourceProject, allocation.destinationProject];
 for (const [index, project] of state.projects.entries()) {
  const id = String(index + 1).repeat(64), network = String(index + 3).repeat(64), volume = `${project}_data`;
  state.container[id] = { Id: id, Name: `/${project}-app-1`, Config: { Labels: { "com.docker.compose.project": project } } };
  state.volume[volume] = { Name: volume, Labels: { "com.docker.compose.project": project } };
  state.network[network] = { Id: network, Name: `${project}_default`, Labels: { "com.docker.compose.project": project } };
 }
 fs.writeFileSync(file, JSON.stringify(state));
 fs.writeFileSync(path.join(allocation.workspace, "private"), "disposable fixture bytes");
}
export function runDeploymentFixture() {
 const args = process.argv.slice(2), state = JSON.parse(fs.readFileSync(process.env.PACKAGE_DEPLOYMENT_STATE)), { scenario } = JSON.parse(fs.readFileSync(process.env.PACKAGE_DEPLOYMENT_CONTROL));
 fs.appendFileSync(process.env.PACKAGE_DEPLOYMENT_LOG, JSON.stringify(args) + "\n");
 const save = () => fs.writeFileSync(process.env.PACKAGE_DEPLOYMENT_STATE, JSON.stringify(state));
 const output = value => process.stdout.write((typeof value === "string" ? value : JSON.stringify(value)) + "\n");
 const fail = () => process.exit(1);
 if (args[0] === "info") { output("fixture-engine"); return; }
 if (args[0] === "compose" && args[1] === "version") { if (scenario === "setup_failure") fail(); output("fixture-compose"); return; }
 if (["container", "volume", "network"].includes(args[0])) {
  const family = args[0], collection = state[family];
  if (args[1] === "ls") {
   const filter = args[args.indexOf("--filter") + 1];
   if (scenario === "collision" && filter.startsWith("name=")) { output(borrowedDeploymentID); return; }
   const matches = Object.values(collection).filter(value => {
    if (filter.startsWith("label=")) { const project = filter.slice("label=com.docker.compose.project=".length); return (value.Config?.Labels ?? value.Labels)?.["com.docker.compose.project"] === project; }
    const name = value.Name.replace(/^\//u, ""), pattern = filter.slice(5);
    return pattern.startsWith("^/") ? name === pattern.slice(2, -1) : name.includes(pattern);
   });
   output(matches.map(value => family === "volume" ? value.Name : value.Id).join("\n")); return;
  }
  if (args[1] === "inspect") { assert.ok(collection[args[2]]); output([collection[args[2]]]); return; }
  if (args[1] === "rm") {
   const id = args.at(-1), value = collection[id]; assert.ok(value);
   const project = (value.Config?.Labels ?? value.Labels)?.["com.docker.compose.project"];
   assert.notEqual(project, "borrowed", "fixture must never lose borrowed resources");
   if (family === "container" && scenario === "source_failure" && project === state.projects?.[0]) fail();
   if (family !== "container") assert.equal(Object.values(state.container).some(container => container.Config.Labels["com.docker.compose.project"] === project), false, "parents cannot be removed before their consumers");
   delete collection[id]; save(); return;
  }
 }
 if (args[0] === "create") {
  assert.equal(Object.values(state.container).some(value => state.projects?.includes(value.Config.Labels["com.docker.compose.project"])), false, "workspace repair cannot run while deployments remain");
  const labels = Object.fromEntries(args.flatMap((arg, index) => arg === "--label" ? [args[index + 1].split("=")] : []));
  const id = "e".repeat(64), name = args[args.indexOf("--name") + 1]; assert.equal(state.container[id], undefined);
  state.container[id] = { Id: id, Name: `/${name}`, Image: args.at(-5), Config: { Labels: labels } }; save();
  if (scenario === "lost_helper_reply") fail();
  output(id); return;
 }
 if (args[0] === "start") {
  assert.ok(state.container[args.at(-1)]);
  if (scenario === "repair_failure") fail();
  if (scenario === "repair_death") { process.kill(Number(process.env.PACKAGE_QUALIFICATION_OWNER_PID), "SIGKILL"); process.exit(0); }
  return;
 }
 throw new Error("unknown fixture command");
}
export const deploymentWorker = `
import fs from "node:fs";
import { createRuntime } from ${JSON.stringify(new URL("../../harness/runtime/session-client.mjs", import.meta.url).href)};
import { reservePackageDeployment } from ${JSON.stringify(new URL("../package-qualification-allocation.mjs", import.meta.url).href)};
import { populateDeploymentFixture } from ${JSON.stringify(import.meta.url)};
const options=JSON.parse(process.env.PACKAGE_DEPLOYMENT_OPTIONS),runtime=createRuntime(options.runtime);
fs.writeFileSync(options.identity,JSON.stringify({root:runtime.root,leaseID:runtime.leaseID}));
Object.assign(process.env,runtime.producerEnvironment());process.env.PACKAGE_QUALIFICATION_OWNER_PID=String(process.pid);
let failed=false,failures=[],allocation,pending=0;
try {
 const handle=options.scenario==="admission_reply"?{...runtime,record(resource){const result=runtime.record(resource);if(resource.state==="pending"&&++pending===2)throw new Error("reply lost");return result;}}:runtime;
 allocation=reservePackageDeployment({runtime:handle,repoRoot:options.runtime.repoRoot,runRoot:options.runtime.runRoot,helperImage:"sha256:"+"c".repeat(64)});
 fs.writeFileSync(options.allocation,JSON.stringify({workspace:allocation.workspace,sourceProject:allocation.sourceProject,destinationProject:allocation.destinationProject,dependencies:allocation.dependencies}));
 populateDeploymentFixture(process.env.PACKAGE_DEPLOYMENT_STATE,allocation);
 if(options.scenario==="death") process.kill(process.pid,"SIGKILL");
 failures=await allocation.releaseAll();failed=failures.length>0;
 if(!failed && (await allocation.releaseAll()).length) throw new Error("repeat release failed");
} catch { failed=true; }
try {runtime.close();} catch {failed=true;}
fs.writeFileSync(options.result,JSON.stringify({failed,failures:failures.map(value=>value.cleanup_step)}));
process.exitCode=failed?1:0;
`;
