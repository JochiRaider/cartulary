package renderprocess

import (
	"errors"
	"runtime"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Linux asm-generic/ioctls.h; this file admits only the native amd64 ABI.
const ioctlFIONBIO = 0x5421

// ExecNetworkChild is the installed launcher's sole operation. It is called
// after namespace/filesystem isolation, before any renderer code executes.
// channelFD is an inherited private socket to the supervising process. The
// listener is transferred, then every copy of that capability is closed before
// exec. Only this OS thread installs the filter; exec destroys all other Go
// runtime threads and descendants inherit the filter irreversibly.
func ExecNetworkChild(channelFD int, command []string, environment []string) error {
	if channelFD < 3 || len(command) == 0 || command[0] == "" || command[0][0] != '/' {
		return ErrConfiguration
	}
	runtime.LockOSThread()
	// Do not unlock after installing a thread-local security boundary, even on
	// failure. The launcher must terminate when this function returns.
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return ErrStart
	}
	filter := networkFilter()
	program := unix.SockFprog{Len: uint16(len(filter)), Filter: &filter[0]}
	fd, _, errno := unix.Syscall(unix.SYS_SECCOMP, unix.SECCOMP_SET_MODE_FILTER, unix.SECCOMP_FILTER_FLAG_NEW_LISTENER, uintptr(unsafe.Pointer(&program)))
	runtime.KeepAlive(filter)
	if errno != 0 {
		return ErrStart
	}
	listener := int(fd)
	err := unix.Sendmsg(channelFD, []byte("render-network-v1"), unix.UnixRights(listener), nil, 0)
	closeErr := errors.Join(unix.Close(listener), unix.Close(channelFD))
	if err != nil || closeErr != nil {
		return ErrStart
	}
	if err := unix.Exec(command[0], command, environment); err != nil {
		return ErrStart
	}
	return nil
}

// The filter admits only the native amd64 ABI. Internet sockets are TCP only;
// connect/bind are emulated by the supervisor using copied addresses and
// duplicated socket descriptions. No decision uses SECCOMP_USER_NOTIF_FLAG_CONTINUE.
// AF_UNIX socket pairs remain available for browser process IPC. Alternate
// network paths (raw/datagram sockets, io_uring, fast-open) cannot bypass connect.
func networkFilter() []unix.SockFilter {
	load := func(offset uint32) unix.SockFilter {
		return unix.SockFilter{Code: unix.BPF_LD | unix.BPF_W | unix.BPF_ABS, K: offset}
	}
	jeq := func(value uint32, yes, no uint8) unix.SockFilter {
		return unix.SockFilter{Code: unix.BPF_JMP | unix.BPF_JEQ | unix.BPF_K, K: value, Jt: yes, Jf: no}
	}
	jset := func(value uint32, yes, no uint8) unix.SockFilter {
		return unix.SockFilter{Code: unix.BPF_JMP | unix.BPF_JSET | unix.BPF_K, K: value, Jt: yes, Jf: no}
	}
	ret := func(value uint32) unix.SockFilter { return unix.SockFilter{Code: unix.BPF_RET | unix.BPF_K, K: value} }
	allow, notify := ret(unix.SECCOMP_RET_ALLOW), ret(unix.SECCOMP_RET_USER_NOTIF)
	deny := ret(unix.SECCOMP_RET_ERRNO | uint32(unix.EPERM))
	filter := []unix.SockFilter{load(4), jeq(unix.AUDIT_ARCH_X86_64, 1, 0), ret(unix.SECCOMP_RET_KILL_PROCESS), load(0), jset(0x40000000, 0, 1), ret(unix.SECCOMP_RET_KILL_PROCESS)}
	// Lock namespace membership before any renderer code executes. This avoids
	// writing namespace sysctls through a container's protected /proc/sys. The
	// filter is inherited by descendants and cannot be replaced. Unknown high
	// clone flags are also denied; clone3 cannot bypass scalar argument checks.
	const namespaceFlags = unix.CLONE_NEWUSER | unix.CLONE_NEWNS | unix.CLONE_NEWNET | unix.CLONE_NEWPID | unix.CLONE_NEWIPC | unix.CLONE_NEWUTS | unix.CLONE_NEWCGROUP | unix.CLONE_NEWTIME
	filter = append(filter, jeq(unix.SYS_CLONE, 0, 7), load(20), jeq(0, 1, 0), notify, load(16), jset(namespaceFlags, 0, 1), notify, load(0))
	for _, call := range []uint32{unix.SYS_CONNECT, unix.SYS_BIND} {
		filter = append(filter, jeq(call, 0, 1), notify)
	}
	// Replacing/unsharing a thread's descriptor table would defeat resolving a
	// notification TID to its process pidfd. clone3 falls back to checked clone.
	filter = append(filter, jeq(unix.SYS_CLONE3, 0, 1), ret(unix.SECCOMP_RET_ERRNO|uint32(unix.ENOSYS)))
	// libuv probes epoll batching independently of UV_USE_IO_URING. No ring may
	// be created, but a denied capability probe is not an outbound attempt.
	// The launcher inherits no ring descriptors; submission/registration remain
	// fatal if reached. No pointer or operation is delegated to the kernel here.
	filter = append(filter, jeq(unix.SYS_IO_URING_SETUP, 0, 1), ret(unix.SECCOMP_RET_ERRNO|uint32(unix.ENOSYS)))
	for _, call := range []uint32{unix.SYS_UNSHARE, unix.SYS_SETNS, unix.SYS_IO_URING_ENTER, unix.SYS_IO_URING_REGISTER, unix.SYS_SECCOMP, unix.SYS_PTRACE, unix.SYS_USERFAULTFD} {
		filter = append(filter, jeq(call, 0, 1), notify)
	}
	filter = append(filter, jeq(unix.SYS_PRCTL, 0, 4), load(16), jeq(unix.PR_SET_SECCOMP, 0, 1), notify, allow, load(0))
	// A duplicated socket shares its file status flags. Prevent another child
	// thread from clearing NONBLOCK between the supervisor's check and connect.
	filter = append(filter, jeq(unix.SYS_FCNTL, 0, 6), load(24), jeq(unix.F_SETFL, 0, 3), load(32), jset(unix.O_NONBLOCK, 1, 0), notify, allow, load(0))
	filter = append(filter, jeq(unix.SYS_IOCTL, 0, 4), load(24), jeq(ioctlFIONBIO, 0, 1), notify, allow, load(0))
	filter = append(filter, jeq(unix.SYS_CLONE, 0, 6), load(16), jset(unix.CLONE_THREAD, 0, 3), jset(unix.CLONE_FILES, 2, 0), deny, allow, allow, load(0))
	filter = append(filter, jeq(unix.SYS_CLOSE_RANGE, 0, 4), load(32), jset(unix.CLOSE_RANGE_UNSHARE, 0, 1), deny, allow, load(0))
	// Flags are scalar syscall arguments, never mutable userspace pointers.
	for _, call := range []struct{ number, flagsOffset uint32 }{{unix.SYS_SENDTO, 40}, {unix.SYS_SENDMSG, 32}, {unix.SYS_SENDMMSG, 40}} {
		filter = append(filter, jeq(call.number, 0, 4), load(call.flagsOffset), jset(unix.MSG_FASTOPEN, 0, 1), notify, allow, load(0))
	}
	// AF_UNIX is local filesystem/IPC scoped by the private mount namespace.
	// Netlink is not needed by the renderer. Deny it without claiming an IP
	// destination; neither a route nor an external namespace can be installed.
	filter = append(filter, jeq(unix.SYS_SOCKET, 0, 14), load(16), jeq(unix.AF_UNIX, 11, 0), jeq(unix.AF_INET, 1, 0), jeq(unix.AF_INET6, 0, 8), load(24), unix.SockFilter{Code: unix.BPF_ALU | unix.BPF_AND | unix.BPF_K, K: 0xf}, jeq(unix.SOCK_STREAM, 1, 0), notify, load(32), jeq(0, 3, 0), jeq(unix.IPPROTO_TCP, 2, 0), notify, deny, allow, load(0))
	filter = append(filter, jeq(unix.SYS_SOCKETPAIR, 0, 4), load(16), jeq(unix.AF_UNIX, 0, 1), allow, deny, allow)
	return filter
}
