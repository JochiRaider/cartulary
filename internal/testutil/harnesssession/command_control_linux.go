//go:build linux

package harnesssession

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
)

// The private control directory is a capability, never a PID lookup service.
// One native parent owns launch, exit observation, settlement and reaping. Shell
// clients consume facts; they cannot detach the command or supply signal targets.
type commandReceipt struct {
	Parent ProcessProof `json:"parent"`
	PID    int          `json:"pid"`
	Exit   int          `json:"exit"`
}

// Preserve the shell lifecycle's ten-second cooperative grace, then bound
// forced settlement to five seconds (the former shell could wait twenty).
const commandGrace = 10 * time.Second
const commandSettlementBudget = commandGrace + 5*time.Second
const commandProtocolBudget = commandSettlementBudget + time.Second

func commandSocket(directory *os.File) string {
	return fmt.Sprintf("/proc/self/fd/%d/control", directory.Fd())
}

func readCommandReceipt(directory *os.File) (commandReceipt, error) {
	var receipt commandReceipt
	data, err := harnesslocal.ReadFileAt(directory, "settled", 4096, false)
	if err != nil {
		return receipt, err
	}
	value, err := harnessschema.Decode(data)
	fields, ok := value.(map[string]any)
	if err != nil || !ok || len(fields) != 3 {
		return receipt, ErrUnresolved
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&receipt) != nil || decoder.Decode(new(any)) != io.EOF || receipt.PID < 0 || receipt.PID == 1 || receipt.Exit < 0 || receipt.Exit > 255 || !DescendsFrom(os.Getpid(), receipt.Parent) {
		return receipt, ErrUnresolved
	}
	return receipt, nil
}

func commandReply(operation string, pid, code int, settled bool) (string, int) {
	switch operation {
	case "ready":
		if pid == 0 {
			return "", 12
		}
		return strconv.Itoa(pid), 0
	case "running":
		if settled {
			return "", 1
		}
		return "", 0
	case "result":
		if settled {
			return "", code
		}
	case "stop":
		if settled {
			return "", 0
		}
	}
	return "", 12
}

func commandClient(directory *os.File, operation string, output io.Writer) int {
	deadline := time.Now().Add(8 * time.Second)
	if operation == "stop" {
		deadline = time.Now().Add(commandProtocolBudget)
	}
	for {
		if receipt, err := readCommandReceipt(directory); err == nil {
			value, code := commandReply(operation, receipt.PID, receipt.Exit, true)
			if value != "" {
				fmt.Fprintln(output, value)
			}
			return code
		} else if !errors.Is(err, os.ErrNotExist) {
			return 12
		}
		connection, err := net.DialTimeout("unix", commandSocket(directory), time.Second)
		if err != nil {
			// Only initial acquisition can wait for a not-yet-bound socket.
			if operation == "ready" && time.Now().Before(deadline) {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			// Publication may race socket closure; require the actual receipt.
			if _, err := readCommandReceipt(directory); err == nil {
				continue
			}
			return 12
		}
		connection.SetDeadline(deadline)
		_, err = fmt.Fprintln(connection, operation)
		data, readErr := io.ReadAll(io.LimitReader(connection, 128))
		connection.Close()
		fields := strings.SplitN(strings.TrimSuffix(string(data), "\n"), " ", 2)
		if err != nil || readErr != nil || len(fields) != 2 {
			// The leader may settle after connect but before accept. A durable
			// acknowledgement is still authoritative; EOF alone never is.
			if _, err := readCommandReceipt(directory); err == nil {
				continue
			}
			return 12
		}
		code, err := strconv.Atoi(fields[0])
		if err != nil || code < 0 || code > 255 {
			return 12
		}
		if fields[1] != "" {
			fmt.Fprintln(output, fields[1])
		}
		return code
	}
}

func commandPeer(connection *net.UnixConn, parent ProcessProof) bool {
	var credential *unix.Ucred
	var credentialErr error
	raw, err := connection.SyscallConn()
	if err != nil {
		return false
	}
	err = raw.Control(func(fd uintptr) {
		credential, credentialErr = unix.GetsockoptUcred(int(fd), unix.SOL_SOCKET, unix.SO_PEERCRED)
	})
	return err == nil && credentialErr == nil && credential.Uid == uint32(os.Getuid()) && DescendsFrom(int(credential.Pid), parent)
}

func serveOwnedCommand(directory *os.File, command *exec.Cmd, parent ProcessProof, diagnostic io.Writer) int {
	// A capability has exactly one acquisition. An old terminal receipt cannot
	// acknowledge a second workload or hide its live native supervisor.
	entries, err := directory.Readdirnames(1)
	if len(entries) != 0 || !errors.Is(err, io.EOF) {
		return 12
	}
	// Bind the parent's kernel identity before any workload can be acquired.
	parentFD, err := commandParent(parent)
	if err != nil {
		return 12
	}
	defer unix.Close(parentFD)
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: commandSocket(directory), Net: "unix"})
	if err != nil {
		return 12
	}
	defer listener.Close()
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	failedAcquisition := func() int {
		data, _ := json.Marshal(commandReceipt{Parent: parent, Exit: 12})
		_ = harnesslocal.ReplaceFileAt(directory, "settled", data)
		fmt.Fprintln(diagnostic, "owned command acquisition failed")
		return 12
	}
	releaseAdmission, err := harnessadmission.InheritWorkspace(command)
	if err != nil {
		return failedAcquisition()
	}
	defer releaseAdmission()
	owner, err := StartOwnedCommand(command, nil)
	if err != nil {
		return failedAcquisition()
	}
	proof, err := json.Marshal(owner.proof)
	if err == nil {
		err = harnesslocal.ReplaceFileAt(directory, "proof", proof)
	}
	settle := func() int {
		settlement, stop := context.WithTimeout(context.Background(), commandSettlementBudget)
		code, failure := owner.Settle(settlement, commandGrace)
		stop()
		if failure != nil {
			_ = json.NewEncoder(diagnostic).Encode(DescribeCleanupFailure(failure))
			return 12
		}
		data, failure := json.Marshal(commandReceipt{Parent: parent, PID: owner.proof.PID, Exit: code})
		if failure == nil {
			failure = harnesslocal.ReplaceFileAt(directory, "settled", data)
		}
		if failure != nil {
			fmt.Fprintln(diagnostic, "owned command acknowledgement failed")
			return 12
		}
		return 0
	}
	if err != nil {
		settle()
		return 12
	}
	for {
		parentGone, pollErr := commandParentGone(parentFD)
		if parentGone || pollErr != nil {
			status := settle()
			if pollErr != nil {
				_ = json.NewEncoder(diagnostic).Encode(DescribeCleanupFailure(AtCleanupStage("process_identity", pollErr)))
				return 12
			}
			if status == 0 && parentGone {
				listener.Close()
				if RemoveOwnedTree(directory, directory.Name()) != nil {
					return 12
				}
			}
			return status
		}
		select {
		case <-owner.exited:
			return settle()
		case <-ctx.Done():
			return settle()
		default:
		}
		listener.SetDeadline(time.Now().Add(100 * time.Millisecond))
		connection, err := listener.AcceptUnix()
		if err != nil {
			if timeout, ok := err.(net.Error); ok && timeout.Timeout() {
				continue
			}
			settle()
			return 12
		}
		connection.SetDeadline(time.Now().Add(time.Second))
		operation := ""
		if commandPeer(connection, parent) {
			operation, _ = bufio.NewReader(io.LimitReader(connection, 16)).ReadString('\n')
			operation = strings.TrimSuffix(operation, "\n")
		}
		if operation == "stop" {
			connection.SetDeadline(time.Now().Add(commandProtocolBudget))
			status := settle()
			fmt.Fprintf(connection, "%d \n", status)
			connection.Close()
			return status
		}
		value, status := commandReply(operation, owner.proof.PID, 0, false)
		fmt.Fprintf(connection, "%d %s\n", status, value)
		connection.Close()
	}
}

// OwnedCommandMain is an internal physical adapter. The allocator returns a
// fresh owner-only capability; serve is a direct child of the invoking shell.
func OwnedCommandMain(args []string, output, diagnostic io.Writer) int {
	if len(args) > 0 && args[0] == "pipe" {
		return pipeCommandMain(args[1:], false)
	}
	if len(args) > 0 && args[0] == "capture" {
		return pipeCommandMain(args[1:], true)
	}
	if len(args) == 1 && args[0] == "allocate" {
		base := filepath.Join(os.TempDir(), "cartulary-owned-commands-"+strconv.Itoa(os.Getuid()))
		parent, err := harnesslocal.PrivateDirectory(base)
		if err != nil {
			return 12
		}
		defer parent.Close()
		name := identifier()
		if unix.Mkdirat(int(parent.Fd()), name, 0700) != nil || harnesslocal.CheckDirectory(parent, base) != nil {
			return 12
		}
		fmt.Fprintln(output, filepath.Join(base, name))
		return 0
	}
	if len(args) < 2 {
		return 2
	}
	operation, name := args[0], args[1]
	base := filepath.Join(os.TempDir(), "cartulary-owned-commands-"+strconv.Itoa(os.Getuid()))
	if filepath.Dir(name) != base || !uuidPattern.MatchString(filepath.Base(name)) {
		return 12
	}
	directory, err := harnesslocal.OpenDirectory(name, true)
	if err != nil {
		return 12
	}
	defer directory.Close()
	if operation == "serve" && len(args) >= 3 {
		parent, err := CaptureProcess(os.Getppid())
		if err != nil || parent.PID < 2 {
			return 12
		}
		command := exec.Command(args[2], args[3:]...)
		command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
		return serveOwnedCommand(directory, command, parent, diagnostic)
	}
	if len(args) != 2 {
		return 2
	}
	if operation == "release" {
		if _, err := readCommandReceipt(directory); err != nil || RemoveOwnedTree(directory, name) != nil {
			return 12
		}
		return 0
	}
	switch operation {
	case "ready", "running", "result", "stop":
		return commandClient(directory, operation, output)
	default:
		return 2
	}
}
