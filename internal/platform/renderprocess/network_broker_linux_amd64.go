package renderprocess

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

type networkNotification struct {
	ID    uint64
	PID   uint32
	Flags uint32
	Data  struct {
		Number             int32
		Architecture       uint32
		InstructionPointer uint64
		Args               [6]uint64
	}
}

type networkResponse struct {
	ID    uint64
	Value int64
	Error int32
	Flags uint32
}

func networkIOCTL(fd int, operation uint, data unsafe.Pointer) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(operation), uintptr(data))
	if errno != 0 {
		return errno
	}
	return nil
}

func receiveNetworkListener(channel int) (int, error) {
	message, control := make([]byte, 64), make([]byte, unix.CmsgSpace(4))
	n, oob, flags, _, err := unix.Recvmsg(channel, message, control, unix.MSG_CMSG_CLOEXEC)
	if err != nil {
		return -1, ErrNetworkObservation
	}
	messages, err := unix.ParseSocketControlMessage(control[:oob])
	var descriptors []int
	for _, item := range messages {
		fds, decodeErr := unix.ParseUnixRights(&item)
		if decodeErr != nil {
			err = decodeErr
		}
		descriptors = append(descriptors, fds...)
	}
	if err != nil || flags&(unix.MSG_TRUNC|unix.MSG_CTRUNC) != 0 || string(message[:n]) != "render-network-v1" || len(messages) != 1 || len(descriptors) != 1 {
		for _, fd := range descriptors {
			_ = unix.Close(fd)
		}
		return -1, ErrNetworkObservation
	}
	return descriptors[0], nil
}

// The supervisor never resumes a pointer-bearing syscall after inspection. It
// copies the address, duplicates the open socket description through a pidfd,
// then performs only a permitted operation on those owned values. Target
// threads can change their pointers or descriptor table without widening access.
func superviseNetwork(ctx context.Context, channel int, port uint16) ([]NetworkObservation, error) {
	defer unix.Close(channel)
	if err := waitNetworkReadable(ctx, channel, false); err != nil {
		return nil, err
	}
	listener, err := receiveNetworkListener(channel)
	if err != nil {
		return nil, fmt.Errorf("%w: listener_channel_invalid", err)
	}
	defer unix.Close(listener)
	// Future kernel ABI changes are rejected rather than silently reading a
	// truncated notification or inventing absent fields.
	var sizes struct{ Notification, Response, Data uint16 }
	_, _, errno := unix.Syscall(unix.SYS_SECCOMP, unix.SECCOMP_GET_NOTIF_SIZES, 0, uintptr(unsafe.Pointer(&sizes)))
	if errno != 0 || uintptr(sizes.Notification) != unsafe.Sizeof(networkNotification{}) || uintptr(sizes.Response) != unsafe.Sizeof(networkResponse{}) || sizes.Data != 64 {
		return nil, ErrNetworkObservation
	}
	observations := make([]NetworkObservation, 0, 3)
	observe := func(kind, class string, allowed bool) {
		for index := range observations {
			if observations[index].Kind == kind && observations[index].DestinationClass == class {
				observations[index].Count++
				return
			}
		}
		observations = append(observations, NetworkObservation{Kind: kind, DestinationClass: class, Allowed: allowed, Count: 1})
	}
	for {
		if err := waitNetworkReadable(ctx, listener, true); err != nil {
			if errors.Is(err, errNetworkFinished) {
				return observations, nil
			}
			return nil, err
		}
		var request networkNotification
		if err := networkIOCTL(listener, unix.SECCOMP_IOCTL_NOTIF_RECV, unsafe.Pointer(&request)); err != nil {
			if errors.Is(err, unix.EINTR) || errors.Is(err, unix.ENOENT) {
				continue
			}
			return nil, ErrNetworkObservation
		}
		response, observation, err := respondNetwork(listener, request, port)
		if observation != nil {
			observe(observation.Kind, observation.DestinationClass, observation.Allowed)
		}
		if err != nil {
			return observations, err
		}
		if err := networkIOCTL(listener, unix.SECCOMP_IOCTL_NOTIF_SEND, unsafe.Pointer(&response)); err != nil && !errors.Is(err, unix.ENOENT) {
			return nil, ErrNetworkObservation
		}
	}
}

var errNetworkFinished = errors.New("network filter finished")

func waitNetworkReadable(ctx context.Context, fd int, filter bool) error {
	for ctx.Err() == nil {
		poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		if _, err := unix.Poll(poll, 50); err != nil && !errors.Is(err, unix.EINTR) {
			return ErrNetworkObservation
		}
		// Unlike a socket, a dead filter can report readable without a pending
		// notification. RECV after its final task exits can block indefinitely.
		if filter && poll[0].Revents&unix.POLLHUP != 0 {
			return errNetworkFinished
		}
		if poll[0].Revents&unix.POLLIN != 0 {
			return nil
		}
		if poll[0].Revents&unix.POLLHUP != 0 {
			return errNetworkFinished
		}
		if poll[0].Revents&(unix.POLLERR|unix.POLLNVAL) != 0 {
			return ErrNetworkObservation
		}
	}
	return ctx.Err()
}

func notificationCurrent(listener int, id uint64) bool {
	return networkIOCTL(listener, unix.SECCOMP_IOCTL_NOTIF_ID_VALID, unsafe.Pointer(&id)) == nil
}

func respondNetwork(listener int, request networkNotification, port uint16) (networkResponse, *NetworkObservation, error) {
	response := networkResponse{ID: request.ID, Error: -int32(unix.EACCES)}
	if request.PID == 0 || request.Flags != 0 || request.Data.Architecture != unix.AUDIT_ARCH_X86_64 {
		return response, nil, ErrNetworkObservation
	}
	if !notificationCurrent(listener, request.ID) {
		return response, nil, nil
	}
	if request.Data.Number == unix.SYS_FCNTL && request.Data.Args[1] == unix.F_SETFL {
		fd, err := copyNetworkSocket(listener, request)
		if errors.Is(err, unix.ENOENT) {
			return response, nil, nil
		}
		if err != nil {
			return response, nil, ErrNetworkObservation
		}
		defer unix.Close(fd)
		allowed, err := descriptorMayBlock(fd)
		if err != nil {
			return response, nil, ErrNetworkObservation
		}
		if !allowed {
			return response, nil, nil
		}
		if _, err := unix.FcntlInt(uintptr(fd), unix.F_SETFL, int(request.Data.Args[2])); err != nil {
			return response, nil, ErrNetworkObservation
		}
		response.Error = 0
		return response, nil, nil
	}
	if request.Data.Number == unix.SYS_IOCTL && request.Data.Args[1] == ioctlFIONBIO {
		value, err := copyNetworkMemory(listener, request, request.Data.Args[2], 4)
		if errors.Is(err, unix.ENOENT) {
			return response, nil, nil
		}
		if err != nil {
			return response, nil, ErrNetworkObservation
		}
		// Clearing NONBLOCK is forbidden even when another thread concurrently
		// changes the argument. Only the copied value is acted upon.
		mode := binary.LittleEndian.Uint32(value)
		if mode > 1 {
			return response, nil, nil
		}
		fd, err := copyNetworkSocket(listener, request)
		if errors.Is(err, unix.ENOENT) {
			return response, nil, nil
		}
		if err != nil {
			return response, nil, ErrNetworkObservation
		}
		defer unix.Close(fd)
		if mode == 0 {
			allowed, err := descriptorMayBlock(fd)
			if err != nil {
				return response, nil, ErrNetworkObservation
			}
			if !allowed {
				return response, nil, nil
			}
		}
		if err := unix.IoctlSetPointerInt(fd, ioctlFIONBIO, int(mode)); err != nil {
			return response, nil, ErrNetworkObservation
		}
		response.Error = 0
		return response, nil, nil
	}
	violation := func(kind, class string) (networkResponse, *NetworkObservation, error) {
		return response, &NetworkObservation{Kind: kind, DestinationClass: class, Count: 1}, &NetworkError{Kind: kind, DestinationClass: class}
	}
	if request.Data.Number == unix.SYS_SOCKET {
		// Creating an unsupported socket has no destination and cannot transmit.
		// Deny the capability itself; do not fabricate an outbound observation.
		// No unsupported socket may be inherited from the launcher either.
		response.Error = -int32(unix.EAFNOSUPPORT)
		return response, nil, nil
	}
	if request.Data.Number != unix.SYS_BIND && request.Data.Number != unix.SYS_CONNECT {
		reason := "unsupported_native_operation"
		switch request.Data.Number {
		case unix.SYS_IO_URING_SETUP, unix.SYS_IO_URING_ENTER, unix.SYS_IO_URING_REGISTER:
			reason = "io_uring_blocked"
		case unix.SYS_PRCTL, unix.SYS_SECCOMP:
			reason = "filter_replacement_blocked"
		case unix.SYS_USERFAULTFD:
			reason = "userfault_blocked"
		case unix.SYS_CLONE, unix.SYS_UNSHARE, unix.SYS_SETNS:
			reason = "namespace_change_blocked"
		case unix.SYS_SENDMSG, unix.SYS_SENDTO, unix.SYS_SENDMMSG:
			reason = "fast_open_blocked"
		case unix.SYS_PTRACE:
			reason = "tracing_blocked"
		}
		return response, nil, fmt.Errorf("%w: %s", ErrNetworkObservation, reason)
	}
	address, err := copyNetworkAddress(listener, request)
	if errors.Is(err, unix.ENOENT) {
		return response, nil, nil
	}
	if err != nil {
		return response, nil, ErrNetworkObservation
	}
	if len(address) < 2 {
		return response, nil, ErrNetworkObservation
	}
	family := binary.LittleEndian.Uint16(address)
	// Unix-domain connects/binds are unnecessary for the renderer's pipe IPC.
	// Their paths are never copied into the host namespace or retained evidence.
	if family == unix.AF_UNIX {
		return response, nil, nil
	}
	var destination unix.Sockaddr
	var requestedPort uint16
	loopback := false
	switch family {
	case unix.AF_INET:
		if len(address) != unix.SizeofSockaddrInet4 {
			return response, nil, ErrNetworkObservation
		}
		requestedPort = binary.BigEndian.Uint16(address[2:4])
		var ip [4]byte
		copy(ip[:], address[4:8])
		loopback = ip == [4]byte{127, 0, 0, 1}
		destination = &unix.SockaddrInet4{Port: int(requestedPort), Addr: ip}
	case unix.AF_INET6:
		if len(address) != unix.SizeofSockaddrInet6 {
			return response, nil, ErrNetworkObservation
		}
		requestedPort = binary.BigEndian.Uint16(address[2:4])
		var ip [16]byte
		copy(ip[:], address[8:24])
		loopback = ip == [16]byte{15: 1} && binary.LittleEndian.Uint32(address[24:28]) == 0
		destination = &unix.SockaddrInet6{Port: int(requestedPort), Addr: ip}
	default:
		return response, nil, ErrNetworkObservation
	}
	if !loopback {
		return violation("non_loopback_blocked", "tcp")
	}
	if requestedPort != port {
		return violation("loopback_blocked", "loopback")
	}
	fd, err := copyNetworkSocket(listener, request)
	if errors.Is(err, unix.ENOENT) {
		return response, nil, nil
	}
	if err != nil {
		return response, nil, ErrNetworkObservation
	}
	defer unix.Close(fd)
	domain, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_DOMAIN)
	if err != nil || domain != int(family) {
		return response, nil, ErrNetworkObservation
	}
	kind, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_TYPE)
	if err != nil || kind != unix.SOCK_STREAM {
		return response, nil, ErrNetworkObservation
	}
	if !notificationCurrent(listener, request.ID) {
		return response, nil, nil
	}
	var observation *NetworkObservation
	if request.Data.Number == unix.SYS_BIND {
		err = unix.Bind(fd, destination)
	} else {
		// Node and Chromium use nonblocking TCP. Never allow an untrusted socket
		// to block the supervisor, nor mutate its shared file status flags.
		flags, flagErr := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
		if flagErr != nil || flags&unix.O_NONBLOCK == 0 {
			return response, nil, ErrNetworkObservation
		}
		err = unix.Connect(fd, destination)
		observation = &NetworkObservation{Kind: "loopback_allowed", DestinationClass: "loopback", Allowed: true, Count: 1}
	}
	response.Error = 0
	if err != nil {
		var errno unix.Errno
		if !errors.As(err, &errno) {
			return response, nil, ErrNetworkObservation
		}
		response.Error = -int32(errno)
	}
	return response, observation, nil
}

// The network broker must never block on an Internet socket. Pipes, ordinary
// files and local Unix IPC can keep their normal blocking semantics (including
// Node's stdio restoration). Check and mutate the same duplicated description;
// do not resume a child syscall against its mutable descriptor table.
func descriptorMayBlock(fd int) (bool, error) {
	domain, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_DOMAIN)
	if errors.Is(err, unix.ENOTSOCK) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return domain == unix.AF_UNIX, nil
}

func copyNetworkAddress(listener int, request networkNotification) ([]byte, error) {
	length := request.Data.Args[2]
	if length < 2 || length > unix.SizeofSockaddrUnix {
		return nil, ErrNetworkObservation
	}
	return copyNetworkMemory(listener, request, request.Data.Args[1], int(length))
}

func copyNetworkMemory(listener int, request networkNotification, pointer uint64, length int) ([]byte, error) {
	// Open the process memory capability first; validate the notification before
	// and after its use, avoiding both PID reuse and abandoned syscall memory.
	file, err := os.Open(fmt.Sprintf("/proc/%d/mem", request.PID))
	if err != nil {
		return nil, ErrNetworkObservation
	}
	defer file.Close()
	if !notificationCurrent(listener, request.ID) {
		return nil, unix.ENOENT
	}
	data := make([]byte, length)
	if pointer > uint64(^uint64(0)>>1) {
		return nil, ErrNetworkObservation
	}
	if _, err := file.ReadAt(data, int64(pointer)); err != nil {
		return nil, ErrNetworkObservation
	}
	if !notificationCurrent(listener, request.ID) {
		return nil, unix.ENOENT
	}
	return data, nil
}

func copyNetworkSocket(listener int, request networkNotification) (int, error) {
	// Notifications name a TID; pidfd_getfd accepts a process pidfd on supported
	// kernels. The filter prevents thread-local descriptor-table unsharing.
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", request.PID))
	if err != nil {
		return -1, ErrNetworkObservation
	}
	process := 0
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Tgid:\t") {
			process, err = strconv.Atoi(strings.TrimPrefix(line, "Tgid:\t"))
			break
		}
	}
	if err != nil || process <= 0 || request.Data.Args[0] > uint64(^uint32(0)>>1) {
		return -1, ErrNetworkObservation
	}
	pidfd, err := unix.PidfdOpen(process, 0)
	if err != nil {
		return -1, ErrNetworkObservation
	}
	defer unix.Close(pidfd)
	if !notificationCurrent(listener, request.ID) {
		return -1, unix.ENOENT
	}
	fd, err := unix.PidfdGetfd(pidfd, int(request.Data.Args[0]), 0)
	if err != nil {
		return -1, ErrNetworkObservation
	}
	if !notificationCurrent(listener, request.ID) {
		_ = unix.Close(fd)
		return -1, unix.ENOENT
	}
	return fd, nil
}
