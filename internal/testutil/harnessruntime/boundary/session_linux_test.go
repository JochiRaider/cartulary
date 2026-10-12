//go:build linux

package boundary

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// Exercise the retained Node adapter against the installed native owner. Expected
// authority and state transitions are assertions, never another runtime algorithm.
func TestStandaloneNativeSession(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	node := os.Getenv("NODE_BIN")
	if node == "" {
		t.Fatal("Make-owned Node required")
	}
	runRoot := t.TempDir()
	if err := os.Chmod(runRoot, 0700); err != nil {
		t.Fatal(err)
	}
	command := exec.CommandContext(ctx, node, "--input-type=module", "-e", `
import assert from 'node:assert/strict';
import {existsSync} from 'node:fs';
import {spawnSync} from 'node:child_process';
import path from 'node:path';
import {createRuntime,borrowRuntime,recoverRuntime} from './tools/harness/runtime/session-client.mjs';
const repoRoot=process.cwd(), runRoot=process.argv[1], scratchRoot=path.join(runRoot,'..','private');
const owned=createRuntime({repoRoot,runRoot,runID:'standalone',scratchRoot});
const environment={...process.env,CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:owned.root,CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:owned.leaseID,CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID:owned.runID};
const borrowed=borrowRuntime({repoRoot,runRoot,environment});
borrowed.registerSecret('standalone-boundary-value');
assert.ok(owned.forbiddenValues().includes('standalone-boundary-value'));
borrowed.close(); assert.equal(existsSync(owned.root),true);
assert.throws(()=>borrowRuntime({repoRoot,runRoot,environment:{...environment,CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:'00000000-0000-4000-8000-000000000000'}}));
assert.equal(owned.scan().status,'pass'); owned.close(); owned.close(); assert.equal(existsSync(owned.root),false);
const source="import {createRuntime} from './tools/harness/runtime/session-client.mjs';const r=createRuntime({repoRoot:process.cwd(),runRoot:process.argv[1],runID:'lost-owner',scratchRoot:process.argv[2]});r.record({kind:'managed_suite',target:r.privatePath('pending.json'),state:'pending'});process.stdout.write(JSON.stringify({root:r.root,leaseID:r.leaseID,runID:r.runID}));";
const child=spawnSync(process.execPath,['--input-type=module','-e',source,runRoot,scratchRoot],{cwd:repoRoot,encoding:'utf8'});
assert.equal(child.status,0,child.stderr);
const identity=JSON.parse(child.stdout);
assert.throws(()=>recoverRuntime({repoRoot,runRoot,...identity})); assert.equal(existsSync(identity.root),true);
const recovery=borrowRuntime({repoRoot,runRoot,environment:{...process.env,CARTULARY_HARNESS_SUITE_RUNTIME_ROOT:identity.root,CARTULARY_HARNESS_SUITE_RUNTIME_LEASE_ID:identity.leaseID,CARTULARY_HARNESS_SUITE_RUNTIME_RUN_ID:identity.runID}});
assert.equal(recovery.resources().length,1);
recovery.record({...recovery.resources()[0],state:'released'});
recoverRuntime({repoRoot,runRoot,...identity}); assert.equal(existsSync(identity.root),false);
`, runRoot)
	command.Dir = root
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("standalone native boundary: %v %s", err, out)
	}
}
