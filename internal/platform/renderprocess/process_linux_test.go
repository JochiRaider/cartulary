package renderprocess

import (
	"context"
	"debug/elf"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestRenderProcessIsolation(t *testing.T) {
	if index := slices.Index(os.Args, "render-network-launcher"); index >= 0 {
		if err := ExecNetworkChild(3, os.Args[index+1:], os.Environ()); err != nil {
			os.Exit(51)
		}
		os.Exit(52)
	}
	if slices.Contains(os.Args, "render-process-fixture") {
		runSandboxFixture()
		return
	}
	runtime := sandboxFixtureRuntime(t)
	t.Setenv("CARTULARY_RENDER_HOST_SECRET", "must-never-cross")
	host, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer host.Close()
	runtime.LoopbackPort = uint16(host.Addr().(*net.TCPAddr).Port)
	input, err := json.Marshal(map[string]string{"action": "isolation", "host_address": host.Addr().String()})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	result, err := runtime.Run(ctx, input, []Output{{Path: "result.bin", MaxBytes: 1024}})
	if err != nil {
		t.Fatal(err)
	}
	if string(result.Protocol) != "isolated" || string(result.Files["result.bin"]) != "rendered" {
		t.Fatalf("unexpected result: %#v", result)
	}
	if len(result.NetworkObservations) != 1 || result.NetworkObservations[0].Kind != "loopback_allowed" {
		t.Fatal("host isolation probe was not observed")
	}
	assertAttemptCleanup(t, runtime)
}

func TestRenderProcessCancellationAndOutputAdmission(t *testing.T) {
	runtime := sandboxFixtureRuntime(t)
	for _, action := range []string{"symlink", "hardlink", "fifo", "oversize", "protocol_overflow", "missing", "nonzero"} {
		t.Run(action, func(t *testing.T) {
			input, err := json.Marshal(map[string]string{"action": action})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			result, err := runtime.Run(ctx, input, []Output{{Path: "result.bin", MaxBytes: 8}})
			want := ErrOutput
			if action == "nonzero" {
				want = ErrExecution
			}
			if !errors.Is(err, want) || result.Files != nil || result.Protocol != nil {
				t.Fatalf("got result=%#v error=%v; want no result and %v", result, err, want)
			}
			assertAttemptCleanup(t, runtime)
		})
	}
	t.Run("terminate and kill", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() { _, err := runtime.Run(ctx, []byte(`{"action":"ignore_term"}`), nil); done <- err }()
		// Wait for this attempt's owned readiness file, rather than assuming the
		// child has installed its signal handler after an arbitrary delay.
		deadline := time.Now().Add(10 * time.Second)
		for {
			entries, err := os.ReadDir(runtime.TemporaryRoot)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) == 1 {
				if _, err := os.Stat(filepath.Join(runtime.TemporaryRoot, entries[0].Name(), "work", "ready")); err == nil {
					break
				}
			}
			select {
			case err := <-done:
				t.Fatalf("child exited before ready: %v", err)
			default:
			}
			if time.Now().After(deadline) {
				t.Fatal("child never became ready")
			}
			time.Sleep(10 * time.Millisecond)
		}
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("got %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("cancellation did not reap sandbox")
		}
		assertAttemptCleanup(t, runtime)
	})
}

func TestRenderProcessRejectsInvalidCapabilities(t *testing.T) {
	t.Run("safe namespace diagnostics", func(t *testing.T) {
		for _, item := range []struct{ input, want string }{
			{"bwrap: cannot open /proc/sys/user/max_user_namespaces: Read-only file system", "sandbox_namespace_lock_unavailable"},
			{"bwrap: No permissions to create a new namespace, likely because the kernel does not allow non-privileged user namespaces", "sandbox_namespace_create_denied"},
			{"bwrap: unshare user ns: Operation not permitted", "sandbox_nested_namespace_failed"},
			{"bwrap: private/path containing secret-value", "sandbox_setup_failed"},
			{"report-render-stage: renderer_resource_missing\nprivate/path containing secret-value", "renderer_resource_missing"},
			{"report-render-stage: renderer_resource_transform_failed\n", "renderer_resource_transform_failed"},
			{"report-render-stage: renderer_resource_load_failed\n", "renderer_resource_load_failed"},
			{"report-render-stage: browser_module_export_missing\n", "browser_module_export_missing"},
			{"report-render-stage: browser_slide_import_failed\n", "browser_slide_import_failed"},
		} {
			if got := safeExitReason([]byte(item.input)); got != item.want {
				t.Fatalf("safe diagnostic = %q, want %q", got, item.want)
			}
		}
	})
	runtime := sandboxFixtureRuntime(t)
	for _, path := range []string{"../escape", "/absolute", "a//b", "a/./b", "a\\b", "a:b", "a\n"} {
		if _, err := runtime.Run(context.Background(), nil, []Output{{Path: path, MaxBytes: 1}}); !errors.Is(err, ErrConfiguration) {
			t.Fatalf("accepted %q: %v", path, err)
		}
	}
	for _, root := range []string{"/", runtime.TemporaryRoot, filepath.Join(runtime.TemporaryRoot, "nested")} {
		candidate := runtime
		candidate.Root = root
		if _, err := candidate.Run(context.Background(), nil, nil); !errors.Is(err, ErrConfiguration) {
			t.Fatalf("accepted root %q: %v", root, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := runtime.Run(ctx, nil, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled run: %v", err)
	}
	assertAttemptCleanup(t, runtime)
}

func assertAttemptCleanup(t *testing.T, runtime Runtime) {
	t.Helper()
	entries, err := os.ReadDir(runtime.TemporaryRoot)
	if err != nil || len(entries) != 0 {
		t.Fatalf("attempts retained: %d, %v", len(entries), err)
	}
}

// The fixture copies only its own executable and exact ELF dependencies, not
// host library trees, the repository, or application configuration. This is a
// test fixture, not production runtime discovery or a sandbox fallback.
func sandboxFixtureRuntime(t *testing.T) Runtime {
	t.Helper()
	bubblewrap, err := exec.LookPath("bwrap")
	if err != nil {
		t.Fatal("the Linux renderer sandbox requires bubblewrap")
	}
	pinBytes, err := os.ReadFile("../../../tools/toolchain_pins.json")
	if err != nil {
		t.Fatal(err)
	}
	var pins struct {
		BubblewrapVersion string `json:"bubblewrap_version"`
	}
	if err := json.Unmarshal(pinBytes, &pins); err != nil || pins.BubblewrapVersion == "" {
		t.Fatal("missing renderer sandbox pin")
	}
	version, err := exec.Command(bubblewrap, "--version").Output()
	if err != nil || strings.TrimSpace(string(version)) != "bubblewrap "+pins.BubblewrapVersion {
		t.Fatal("installed bubblewrap does not match tools/toolchain_pins.json")
	}
	root := t.TempDir()
	for _, mountpoint := range []string{"proc", "dev", "work", "diagnostics", "tmp"} {
		if err := os.Mkdir(filepath.Join(root, mountpoint), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	copied := map[string]bool{}
	var copyELF func(string, string)
	copyELF = func(source, destination string) {
		if copied[destination] {
			return
		}
		copied[destination] = true
		file, err := elf.Open(source)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		for _, program := range file.Progs {
			if program.Type == elf.PT_INTERP {
				data, err := io.ReadAll(program.Open())
				if err != nil {
					t.Fatal(err)
				}
				interpreter := strings.TrimRight(string(data), "\x00")
				copyELF(interpreter, interpreter)
			}
		}
		libraries, err := file.ImportedLibraries()
		if err != nil {
			t.Fatal(err)
		}
		for _, library := range libraries {
			found := false
			for _, directory := range []string{"/lib/x86_64-linux-gnu", "/usr/lib/x86_64-linux-gnu", "/lib64"} {
				candidate := filepath.Join(directory, library)
				if _, err := os.Stat(candidate); err == nil {
					copyELF(candidate, candidate)
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("missing fixture library %q", library)
			}
		}
		data, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(root, destination)
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, data, 0o500); err != nil {
			t.Fatal(err)
		}
	}
	copyELF(executable, "/bin/probe")
	return Runtime{SandboxCommand: []string{bubblewrap}, Root: root, TemporaryRoot: t.TempDir(), Command: []string{"/bin/probe", "-test.run=^TestRenderProcessIsolation$", "--", "render-process-fixture"}, Launcher: []string{"/bin/probe", "-test.run=^TestRenderProcessIsolation$", "--", "render-network-launcher"}, LoopbackPort: 32123, Environment: map[string]string{"LANG": "C.UTF-8", "TZ": "UTC", "HOME": "/work/home", "TMPDIR": "/work/tmp"}}
}

func runSandboxFixture() {
	var input map[string]string
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		os.Exit(20)
	}
	write := func(name, value string) {
		if err := os.WriteFile(name, []byte(value), 0o600); err != nil {
			os.Exit(21)
		}
	}
	switch input["action"] {
	case "isolation":
		// This port is declared for the child, but only the parent has a listener.
		// The socket's private namespace must prevent reaching that host listener.
		connection, err := net.DialTimeout("tcp4", input["host_address"], time.Second)
		if err == nil {
			connection.Close()
			os.Exit(75)
		}
		if os.Getenv("CARTULARY_RENDER_HOST_SECRET") != "" || os.Getenv("HOME") != "/work/home" {
			os.Exit(22)
		}
		if _, err := os.Stat("/etc/passwd"); !os.IsNotExist(err) {
			os.Exit(23)
		}
		for _, name := range []string{"/dev/pts", "/dev/tty", "/dev/shm"} {
			if _, err := os.Stat(name); !os.IsNotExist(err) {
				os.Exit(90)
			}
		}
		if err := os.WriteFile("/dev/undeclared", []byte("forbidden"), 0o600); err == nil {
			os.Exit(91)
		}
		if err := os.WriteFile("/runtime-change", []byte("forbidden"), 0o600); err == nil {
			os.Exit(24)
		}
		write("/work/result.bin", "rendered")
		_, _ = os.Stdout.WriteString("isolated")
	case "declared_loopback":
		listener, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
		if err != nil {
			os.Exit(61)
		}
		defer unix.Close(listener)
		if unix.Bind(listener, &unix.SockaddrInet4{Port: 32123, Addr: [4]byte{127, 0, 0, 1}}) != nil || unix.Listen(listener, 1) != nil {
			os.Exit(61)
		}
		go func() {
			_, _ = unix.Poll([]unix.PollFd{{Fd: int32(listener), Events: unix.POLLIN}}, 1000)
			connection, _, err := unix.Accept4(listener, unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC)
			if err != nil {
				return
			}
			defer unix.Close(connection)
			_, _ = unix.Write(connection, []byte("private-loopback"))
		}()
		connection, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC, unix.IPPROTO_TCP)
		if err != nil {
			os.Exit(62)
		}
		defer unix.Close(connection)
		err = unix.Connect(connection, &unix.SockaddrInet4{Port: 32123, Addr: [4]byte{127, 0, 0, 1}})
		if err != nil && !errors.Is(err, unix.EINPROGRESS) {
			os.Exit(62)
		}
		_, _ = unix.Poll([]unix.PollFd{{Fd: int32(connection), Events: unix.POLLIN}}, 1000)
		data := make([]byte, 100)
		n, err := unix.Read(connection, data)
		if err != nil || string(data[:n]) != "private-loopback" {
			os.Exit(63)
		}
		write("/work/result.bin", "rendered")
	case "undeclared_loopback", "non_loopback":
		address := "127.0.0.1:32124"
		if input["action"] == "non_loopback" {
			address = "192.0.2.1:80"
		}
		connection, _ := net.DialTimeout("tcp4", address, time.Second)
		if connection != nil {
			connection.Close()
		}
		write("/work/result.bin", "must-not-publish")
	case "udp":
		fd, err := unix.Socket(unix.AF_INET, unix.SOCK_DGRAM|unix.SOCK_NONBLOCK, 0)
		if fd >= 0 || !errors.Is(err, unix.EAFNOSUPPORT) {
			os.Exit(76)
		}
		write("/work/result.bin", "rendered")
	case "raw_socket", "other_protocol":
		kind, protocol := unix.SOCK_RAW, unix.IPPROTO_TCP
		if input["action"] == "other_protocol" {
			kind, protocol = unix.SOCK_STREAM, unix.IPPROTO_SCTP
		}
		fd, err := unix.Socket(unix.AF_INET, kind|unix.SOCK_NONBLOCK, protocol)
		if fd >= 0 || !errors.Is(err, unix.EAFNOSUPPORT) {
			os.Exit(77)
		}
		write("/work/result.bin", "rendered")
	case "descriptor_mode":
		fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK, 0)
		if err != nil {
			os.Exit(64)
		}
		defer unix.Close(fd)
		if _, err := unix.FcntlInt(uintptr(fd), unix.F_SETFL, 0); !errors.Is(err, unix.EACCES) {
			os.Exit(65)
		}
		if err := unix.IoctlSetPointerInt(fd, 0x5421, 0); !errors.Is(err, unix.EACCES) {
			os.Exit(66)
		}
		if err := unix.IoctlSetPointerInt(fd, 0x5421, 1); err != nil {
			os.Exit(67)
		}
		flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
		if err != nil || flags&unix.O_NONBLOCK == 0 {
			os.Exit(68)
		}
		write("/work/result.bin", "rendered")
	case "clone_user_namespace", "clone_network_namespace", "clone_unknown_flags":
		flags := uintptr(unix.CLONE_NEWUSER | unix.SIGCHLD)
		if input["action"] == "clone_network_namespace" {
			flags = uintptr(unix.CLONE_NEWNET | unix.SIGCHLD)
		}
		if input["action"] == "clone_unknown_flags" {
			flags = uintptr(1<<32 | unix.SIGCHLD)
		}
		pid, _, _ := unix.RawSyscall(unix.SYS_CLONE, flags, 0, 0)
		if pid == 0 {
			unix.RawSyscall(unix.SYS_EXIT, 89, 0, 0)
		}
		os.Exit(89)
	case "unshare_namespace":
		_ = unix.Unshare(unix.CLONE_NEWUSER)
		os.Exit(89)
	case "pipe_mode":
		var descriptors [2]int
		if err := unix.Pipe2(descriptors[:], unix.O_CLOEXEC|unix.O_NONBLOCK); err != nil {
			os.Exit(71)
		}
		defer unix.Close(descriptors[0])
		defer unix.Close(descriptors[1])
		if _, err := unix.FcntlInt(uintptr(descriptors[0]), unix.F_SETFL, 0); err != nil {
			os.Exit(72)
		}
		if err := unix.IoctlSetPointerInt(descriptors[1], ioctlFIONBIO, 0); err != nil {
			os.Exit(73)
		}
		for _, fd := range descriptors {
			flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
			if err != nil || flags&unix.O_NONBLOCK != 0 {
				os.Exit(74)
			}
		}
		write("/work/result.bin", "rendered")
	case "unix_address":
		fd, err := unix.Socket(unix.AF_UNIX, unix.SOCK_STREAM, 0)
		if err != nil {
			os.Exit(78)
		}
		defer unix.Close(fd)
		if err := unix.Bind(fd, &unix.SockaddrUnix{Name: "/work/tmp/named-socket"}); !errors.Is(err, unix.EACCES) {
			os.Exit(79)
		}
		if err := unix.Connect(fd, &unix.SockaddrUnix{Name: "/work/tmp/named-socket"}); !errors.Is(err, unix.EACCES) {
			os.Exit(80)
		}
		write("/work/result.bin", "rendered")
	case "fast_open":
		fd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM|unix.SOCK_NONBLOCK, 0)
		if err != nil {
			os.Exit(69)
		}
		defer unix.Close(fd)
		_ = unix.Sendto(fd, []byte("forbidden"), unix.MSG_FASTOPEN, &unix.SockaddrInet4{Port: 80, Addr: [4]byte{192, 0, 2, 1}})
		write("/work/result.bin", "must-not-publish")
	case "io_uring":
		_, _, errno := unix.Syscall(unix.SYS_IO_URING_SETUP, 1, 0, 0)
		if errno != unix.ENOSYS {
			os.Exit(70)
		}
		write("/work/result.bin", "rendered")
	case "userfault":
		_, _, _ = unix.Syscall(unix.SYS_USERFAULTFD, 0, 0, 0)
		write("/work/result.bin", "must-not-publish")
	case "filter_replace":
		_, _, _ = unix.Syscall(unix.SYS_SECCOMP, unix.SECCOMP_SET_MODE_FILTER, 0, 0)
		write("/work/result.bin", "must-not-publish")
	case "descendant":
		child := exec.Command("/bin/probe", "-test.run=^TestRenderProcessIsolation$", "--", "render-process-fixture")
		child.Stdin = strings.NewReader(`{"action":"non_loopback"}`)
		_ = child.Run()
		write("/work/result.bin", "must-not-publish")
	case "symlink":
		if err := os.Symlink("/bin/probe", "/work/result.bin"); err != nil {
			os.Exit(26)
		}
	case "hardlink":
		write("/work/source", "data")
		if err := os.Link("/work/source", "/work/result.bin"); err != nil {
			os.Exit(27)
		}
	case "fifo":
		if err := syscall.Mkfifo("/work/result.bin", 0o600); err != nil {
			os.Exit(28)
		}
	case "oversize":
		write("/work/result.bin", "too many bytes")
	case "protocol_overflow":
		_, _ = os.Stdout.WriteString(strings.Repeat("x", protocolLimit+1))
	case "missing":
	case "nonzero":
		_, _ = os.Stderr.WriteString("SECRET MUST NOT ESCAPE")
		os.Exit(29)
	case "ignore_term":
		signal.Ignore(syscall.SIGTERM)
		write("/work/ready", "ready")
		for {
			time.Sleep(time.Second)
		}
	default:
		os.Exit(30)
	}
	os.Exit(0)
}

func TestRenderNetworkPolicy(t *testing.T) {
	for _, test := range []struct{ action, kind, class string }{
		{"declared_loopback", "loopback_allowed", "loopback"},
		{"undeclared_loopback", "loopback_blocked", "loopback"},
		{"non_loopback", "non_loopback_blocked", "tcp"},
		{"udp", "", ""},
		{"raw_socket", "", ""},
		{"other_protocol", "", ""},
		{"descendant", "non_loopback_blocked", "tcp"},
		{"descriptor_mode", "", ""},
		{"pipe_mode", "", ""},
		{"unix_address", "", ""},
		{"fast_open", "unavailable", ""},
		{"io_uring", "", ""},
		{"userfault", "unavailable", ""},
		{"filter_replace", "unavailable", ""},
		{"clone_user_namespace", "unavailable", ""},
		{"clone_network_namespace", "unavailable", ""},
		{"clone_unknown_flags", "unavailable", ""},
		{"unshare_namespace", "unavailable", ""},
	} {
		t.Run(test.action, func(t *testing.T) {
			runtime := sandboxFixtureRuntime(t)
			input, _ := json.Marshal(map[string]string{"action": test.action})
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			result, err := runtime.Run(ctx, input, []Output{{Path: "result.bin", MaxBytes: 100}})
			if test.kind == "" {
				if err != nil || string(result.Files["result.bin"]) != "rendered" || len(result.NetworkObservations) != 0 {
					t.Fatalf("descriptor flags not protected: %v", err)
				}
			} else if test.kind == "unavailable" {
				if !errors.Is(err, ErrNetworkObservation) || len(result.Files) != 0 || len(result.Protocol) != 0 {
					t.Fatalf("unobservable alternate path did not fail closed: %v", err)
				}
			} else if test.action == "declared_loopback" {
				if err != nil {
					t.Fatal(err)
				}
				if string(result.Files["result.bin"]) != "rendered" || len(result.NetworkObservations) != 1 || result.NetworkObservations[0] != (NetworkObservation{Kind: test.kind, DestinationClass: test.class, Allowed: true, Count: 1}) {
					t.Fatalf("missing actual loopback observation: %#v", result.NetworkObservations)
				}
			} else {
				var violation *NetworkError
				if !errors.As(err, &violation) || violation.Kind != test.kind || violation.DestinationClass != test.class || len(result.Files) != 0 || len(result.Protocol) != 0 {
					t.Fatalf("network attempt did not fail closed: result=%#v error=%v", result, err)
				}
			}
			assertAttemptCleanup(t, runtime)
		})
	}
}
