package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const (
	httpAddrEnv     = "CARTULARY_HTTP_ADDR"
	httpListenFDEnv = "CARTULARY_HTTP_LISTEN_FD"
)

func main() {
	listenAddr := flag.String("listen", "", "TCP listen address to reserve for the child server")
	flag.Parse()

	if strings.TrimSpace(*listenAddr) == "" {
		_, _ = fmt.Fprintln(os.Stderr, "webstacklisten requires --listen")
		os.Exit(2)
	}
	if flag.NArg() == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "webstacklisten requires a child command after --")
		os.Exit(2)
	}

	listener, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "reserve backend listener %s: %v\n", *listenAddr, err)
		os.Exit(1)
	}
	defer listener.Close()

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		_, _ = fmt.Fprintf(os.Stderr, "reserved listener is %T, expected *net.TCPListener\n", listener)
		os.Exit(1)
	}
	listenerFile, err := tcpListener.File()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "duplicate backend listener fd: %v\n", err)
		os.Exit(1)
	}
	defer listenerFile.Close()

	args := flag.Args()
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = append(os.Environ(),
		httpAddrEnv+"="+listener.Addr().String(),
		httpListenFDEnv+"=3",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.ExtraFiles = []*os.File{listenerFile}
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	if err := cmd.Start(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "start backend child: %v\n", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	if err == nil {
		return
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		os.Exit(exitError.ExitCode())
	}
	_, _ = fmt.Fprintf(os.Stderr, "wait backend child: %v\n", err)
	os.Exit(1)
}
