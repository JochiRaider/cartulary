//go:build linux

package harnessaccount

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

func TestAccountingSignalExit(t *testing.T) {
	if os.Getenv("CARTULARY_ACCOUNTING_SIGNAL_CHILD") == "1" {
		cmd := exec.Command("/bin/sh", "-c", "trap 'exit 143' TERM; printf 'ready\\n'; sleep 3")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Stdout = os.Stdout
		code, received, _, err := runChild(cmd)
		if err != nil || code != 143 || received != 15 {
			t.Fatalf("signal settlement: code=%d signal=%d err=%v", code, received, err)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestAccountingSignalExit$")
	cmd.Env = append(os.Environ(), "CARTULARY_ACCOUNTING_SIGNAL_CHILD=1")
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(output)
	if !scanner.Scan() || scanner.Text() != "ready" {
		_ = cmd.Wait()
		t.Fatalf("child not ready: %s", stderr.String())
	}
	if err = cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	for scanner.Scan() {
	}
	if err = cmd.Wait(); err != nil {
		t.Fatalf("signal child failed: %v %s", err, stderr.String())
	}
}

func TestAccountingCounterIntegrity(t *testing.T) {
	if !sameExecutable("/proc/self/exe", "/proc/self/exe") || sameExecutable("/proc/self/exe", "/bin/sh") || sameExecutable("/proc/self/exe", "/does-not-exist") {
		t.Fatal("executable identity check accepted missing or different content")
	}
	for _, input := range []string{"", "-1", "+1", "1.0", "NaN", "9007199254740992", "1 2"} {
		if _, err := unsigned(input); err == nil {
			t.Fatalf("accepted invalid counter %q", input)
		}
	}
	for _, input := range []string{"user_usec 5", "usage_usec 1\nusage_usec 2", "usage_usec nope", "usage_usec 1 extra"} {
		if _, err := counterField(input, "usage_usec"); err == nil {
			t.Fatalf("accepted invalid kernel record %q", input)
		}
	}
	if n, err := counterField("usage_usec 123\nuser_usec 100\nsystem_usec 23\n", "usage_usec"); err != nil || n != 123 {
		t.Fatalf("counter: %d %v", n, err)
	}
	if _, err := containerParentForMember("/cartulary-account-00000000-0000-0000-0000-000000000000.scope/measured/commands"); err == nil {
		t.Fatal("accepted an absent owner proof")
	}
	for _, invalid := range []string{"populated", "oom", "missing_cpu"} {
		t.Run(invalid, func(t *testing.T) {
			dir := t.TempDir()
			files := map[string]string{"cpu.stat": "usage_usec 1\n", "memory.peak": "4096\n", "memory.events": "oom_kill 0\n", "cgroup.events": "populated 0\n"}
			switch invalid {
			case "populated":
				files["cgroup.events"] = "populated 1\n"
			case "oom":
				files["memory.events"] = "oom_kill 1\n"
			case "missing_cpu":
				delete(files, "cpu.stat")
			}
			for name, data := range files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := os.Mkdir(filepath.Join(dir, "commands"), 0700); err != nil {
				t.Fatal(err)
			}
			root, err := os.OpenRoot(dir)
			if err != nil {
				t.Fatal(err)
			}
			defer root.Close()
			g := group{root: root, path: "/cartulary-test-unavailable/owned"}
			if g.release() == nil {
				t.Fatal("accepted uncertain accounting")
			}
			if _, err := root.Stat("commands"); err != nil {
				t.Fatal("uncertain release mutated owned hierarchy", err)
			}
		})
	}
}

func TestAccountingKernelLifecycle(t *testing.T) {
	scope := os.Getenv("CARTULARY_ACCOUNTING_TEST_SCOPE")
	if scope == "" {
		if parent, err := ContainerParent(); err != nil || parent != "" {
			t.Fatal("delegation probe cannot escape an accounted invocation")
		}
		id, err := os.ReadFile("/proc/sys/kernel/random/uuid")
		if err != nil {
			t.Fatal(err)
		}
		scope = "cartulary-account-" + strings.TrimSpace(string(id)) + ".scope"
		cmd := exec.Command("systemd-run", "--user", "--quiet", "--scope", "--property=Delegate=yes", "--unit="+scope, os.Args[0], "-test.run=^TestAccountingKernelLifecycle$", "-test.v")
		cmd.Env = append(os.Environ(), "CARTULARY_ACCOUNTING_TEST_SCOPE="+scope)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("delegated lifecycle probe: %v\n%s", err, output)
		}
		if !bytes.Contains(output, []byte("kernel-lifetime-complete")) {
			t.Fatalf("missing lifecycle proof: %s", output)
		}
		return
	}
	g, err := createGroup(scope)
	if err != nil {
		t.Fatal(err)
	}
	defer g.close()
	root, err := createPrivateRoot(scope)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)
	if err = os.Mkdir(filepath.Join(root, "invocation"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = harnessdiag.OpenInvocation(filepath.Join(root, "invocation")); err != nil {
		t.Fatal(err)
	}
	if err = writeBinding(root, g); err != nil {
		t.Fatal(err)
	}
	// Missing or replaced controller identity must not authorize another borrow,
	// even though the kernel membership still matches the original hierarchy.
	for _, file := range []string{"binding.json", "invocation/owner.json"} {
		name := filepath.Join(root, file)
		saved, err := harnesslocal.ReadFile(name, 4096, false)
		if err != nil {
			t.Fatal(err)
		}
		var broken map[string]any
		if err = json.Unmarshal(saved, &broken); err != nil {
			t.Fatal(err)
		}
		if file == "binding.json" {
			broken["inode"] = "0"
		} else {
			broken["start_ticks"] = "0"
		}
		data, _ := json.Marshal(broken)
		if err = harnesslocal.ReplaceFile(name, data); err != nil {
			t.Fatal(err)
		}
		child := exec.Command(os.Args[0], "-test.run=^TestAccountingShortLivedChild$")
		child.Env = append(os.Environ(), "CARTULARY_ACCOUNTING_TEST_CHILD=1", "CARTULARY_ACCOUNTING_REJECT_FENCE=1")
		child.SysProcAttr = &syscall.SysProcAttr{UseCgroupFD: true, CgroupFD: int(g.commands.Fd())}
		if output, err := child.CombinedOutput(); err != nil {
			t.Fatalf("stale owner fence: %v\n%s", err, output)
		}
		if err = harnesslocal.ReplaceFile(name, saved); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestAccountingShortLivedChild$")
	cmd.Env = append(os.Environ(), "CARTULARY_ACCOUNTING_TEST_CHILD=1")
	cmd.SysProcAttr = &syscall.SysProcAttr{UseCgroupFD: true, CgroupFD: int(g.commands.Fd())}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("short-lived child: %v\n%s", err, output)
	}
	// A child can exit while an orphaned descendant still owns work. Only the
	// fixture's exact pipe capability settles it; accounting must not kill it.
	input, release, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer release.Close()
	orphan := exec.Command("/bin/sh", "-c", "exec 3<&0; (read line <&3) </dev/null >/dev/null 2>&1 & exit 0")
	orphan.Stdin = input
	orphan.SysProcAttr = &syscall.SysProcAttr{UseCgroupFD: true, CgroupFD: int(g.commands.Fd())}
	err = orphan.Run()
	input.Close()
	if err != nil {
		t.Fatal(err)
	}
	if value, err := g.snapshot(); err != nil || !value.Populated {
		t.Fatal("orphan lifetime was lost", value, err)
	}
	if g.release() == nil {
		t.Fatal("retired a populated hierarchy")
	}
	if err = release.Close(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		value, err := g.snapshot()
		if err != nil {
			t.Fatal(err)
		}
		if !value.Populated {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("owned orphan did not settle")
		}
		time.Sleep(time.Millisecond)
	}
	value, err := g.snapshot()
	if err != nil || value.Populated || value.OOMKills != 0 || value.CPU == 0 || value.Peak < 16*1024*1024 {
		t.Fatalf("lost completed child accounting: %+v %v", value, err)
	}
	// An empty but unretired service generation still belongs to its lifecycle
	// owner. Accounting must neither adopt nor remove that generation.
	if err = g.root.Mkdir("unretired-service", 0700); err != nil {
		t.Fatal(err)
	}
	if g.release() == nil {
		t.Fatal("removed an unretired service group")
	}
	if _, err = g.root.Stat("unretired-service"); err != nil {
		t.Fatal("failed release destroyed ownership", err)
	}
	if err = g.root.Remove("unretired-service"); err != nil {
		t.Fatal(err)
	}
	if err = g.release(); err != nil {
		t.Fatal(err)
	}
	fmt.Println("kernel-lifetime-complete")
}

func TestAccountingShortLivedChild(t *testing.T) {
	if os.Getenv("CARTULARY_ACCOUNTING_TEST_CHILD") != "1" {
		return
	}
	parent, err := ContainerParent()
	if os.Getenv("CARTULARY_ACCOUNTING_REJECT_FENCE") == "1" {
		if err == nil || parent != "" {
			t.Fatal("accepted stale owner or hierarchy identity")
		}
		return
	}
	if err != nil || parent == "" {
		t.Fatal("child missing its live accounting fence", err)
	}
	if err = VerifyRuntimeExecutable(); err != nil {
		t.Fatal("current child executable rejected", err)
	}
	if VerifyContainer(parent, strings.Repeat("a", 64)) == nil {
		t.Fatal("accepted a foreign container")
	}
	data := make([]byte, 32*1024*1024)
	for i := range data {
		data[i] = byte(i)
	}
	until := time.Now().Add(25 * time.Millisecond)
	for time.Now().Before(until) {
		data[0]++
	}
	runtime.KeepAlive(data)
}
