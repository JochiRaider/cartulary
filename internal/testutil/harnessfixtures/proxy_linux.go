//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"maps"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/objectstoreproxy"
	"golang.org/x/sys/unix"
)

// A detached physical proxy remains a waited child. Its leader is kept reserved
// until exact group cleanup; an exited/replaced PID cannot become a kill target.
func startProxy(p ProductionOptions, binary string, args []string, environment map[string]string, log string) (func(context.Context) error, error) {
	fd, err := unix.Open(log, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0600)
	if err != nil {
		return nil, err
	}
	output := os.NewFile(uintptr(fd), "proxy-log")
	defer output.Close()
	cmd := exec.Command(binary, args...)
	cmd.Dir = p.Root
	cmd.Stdout = output
	cmd.Stderr = output
	for _, key := range slices.Sorted(maps.Keys(environment)) {
		cmd.Env = append(cmd.Env, key+"="+environment[key])
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	close, err := harnessadmission.InheritWorkspace(cmd)
	if err != nil {
		return nil, err
	}
	defer close()
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	target, proofErr := harnesssession.CaptureProcess(cmd.Process.Pid)
	if proofErr == nil {
		proofErr = p.Runtime.RecordDependentResource("helper_process", target, "acquired", p.Dependencies)
	}
	terminal := make(chan error, 1)
	go func() {
		for {
			var info unix.Siginfo
			err := unix.Waitid(unix.P_PID, cmd.Process.Pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
			if err == unix.EINTR {
				continue
			}
			terminal <- err
			return
		}
	}()
	var once sync.Once
	var stopErr error
	stop := func(ctx context.Context) error {
		once.Do(func() {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
			timer := time.NewTimer(time.Second)
			defer timer.Stop()
			var waitErr error
			select {
			case waitErr = <-terminal:
			case <-timer.C:
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				waitErr = <-terminal
			case <-ctx.Done():
				_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
				waitErr = <-terminal
			}
			settleCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			settleErr := harnesssession.StopProcess(settleCtx, target, 0)
			if settleErr == nil {
				settleErr = p.Runtime.RetireDetailProducers(target)
			}
			cancel()
			_ = cmd.Wait()
			if proofErr == nil && settleErr == nil {
				stopErr = p.Runtime.RecordResource("helper_process", target, "released")
			}
			stopErr = errors.Join(waitErr, settleErr, stopErr)
		})
		return stopErr
	}
	if proofErr != nil {
		ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		return nil, errors.Join(proofErr, stop(ctx))
	}
	return stop, nil
}

type proxyProvider struct {
	options  ProductionOptions
	suite    *SuiteOwner
	build    sync.Once
	binary   string
	buildErr error
}

func (p *proxyProvider) Close(context.Context) error { return nil }
func (p *proxyProvider) Acquire(ctx context.Context, _ Request) (allocation *Allocation, result error) {
	options := p.options
	suite, err := p.suite.Ensure(ctx)
	if err != nil {
		return nil, err
	}
	options.Dependencies = []string{suite.ResourceRef}
	environment := options.environment()
	maps.Copy(environment, suite.Environment)
	root, err := options.Runtime.RecoveryPath("object-store-proxy")
	if err != nil {
		return nil, err
	}
	directory, err := harnesslocal.PrivateDirectory(root)
	if err != nil {
		return nil, err
	}
	directory.Close()
	detail, err := options.Runtime.PrivatePath("object-store-proxy")
	if err != nil {
		return nil, err
	}
	detailDirectory, err := harnesslocal.PrivateDirectory(detail)
	if err != nil {
		return nil, err
	}
	detailDirectory.Close()
	p.build.Do(func() {
		p.binary = filepath.Join(detail, "s3corsproxy")
		buildEnv := maps.Clone(environment)
		for source, target := range map[string]string{"GO_CACHE_DIR": "GOCACHE", "GO_MOD_CACHE_DIR": "GOMODCACHE", "GO_TMP_DIR": "GOTMPDIR"} {
			if environment[source] == "" {
				p.buildErr = &Failure{Class: "config", Reason: "configuration_error"}
				return
			}
			buildEnv[target] = environment[source]
		}
		if environment["GO"] == "" {
			p.buildErr = &Failure{Class: "config", Reason: "configuration_error"}
			return
		}
		_, p.buildErr = options.command()(ctx, environment["GO"], []string{"build", "-o", p.binary, "./tools/s3corsproxy"}, buildEnv)
	})
	if p.buildErr != nil {
		return nil, p.buildErr
	}
	id := strings.ReplaceAll(identifier(), "-", "")
	instance := filepath.Join(root, id)
	directory, err = harnesslocal.PrivateDirectory(instance)
	if err != nil {
		return nil, err
	}
	directory.Close()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	listen := listener.Addr().String()
	if err := listener.Close(); err != nil {
		return nil, err
	}
	attempt, lease, log := filepath.Join(instance, "start-attempt.json"), filepath.Join(instance, "ready-lease.json"), filepath.Join(detail, id+".log")
	// The proxy owns a trust snapshot even when the suite is borrowed. Recovery
	// never acquires retention authority over the external suite owner's files.
	trust := filepath.Join(instance, "root.pem")
	trustBytes, err := harnesslocal.ReadFile(suite.Environment["SSL_CERT_FILE"], 262144, false)
	if err != nil {
		return nil, artifactFailure(err)
	}
	if err := harnesslocal.PublishNew(trust, trustBytes); err != nil {
		return nil, artifactFailure(err)
	}
	common := []string{"--listen", listen, "--upstream", "https://" + suite.Environment["CARTULARY_S3TEST_ENDPOINT"], "--origin", "https://localhost:5173", "--tls-certificate", suite.Environment["CARTULARY_S3TEST_PROXY_CERTIFICATE_PATH"], "--tls-private-key", suite.Environment["CARTULARY_S3TEST_PROXY_PRIVATE_KEY_PATH"], "--tls-root-certificate", trust}
	for i := 1; i < len(common); i += 2 {
		if common[i] == "" || common[i] == "https://" {
			return nil, &Failure{Class: "config", Reason: "configuration_error"}
		}
	}
	command := func(ctx context.Context, operation string, args ...string) error {
		if operation == "stop" {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
		}
		_, err := options.command()(ctx, p.binary, append(append([]string{operation}, common...), args...), environment)
		return err
	}
	if err := command(ctx, "attempt", "--attempt-file", attempt, "--instance-id", id, "--log-path", log); err != nil {
		return nil, err
	}
	if err := options.Runtime.RecordDependentResource("object_store_proxy", attempt, "pending", options.Dependencies, trust); err != nil {
		return nil, artifactFailure(err)
	}
	stop, err := startProxy(options, p.binary, append(append([]string{"serve"}, common...), "--attempt-file", attempt, "--instance-id", id, "--log-path", log), environment, log)
	if err != nil {
		return nil, err
	}
	ready := false
	resourceTarget := attempt
	var once sync.Once
	var closeErr error
	close := func(ctx context.Context) error {
		once.Do(func() {
			closeErr = options.closeResource(ctx, "object_store_proxy", resourceTarget, func(ctx context.Context) error {
				closeErr = options.Runtime.AttemptCleanup(ctx, "object_store_proxy", lease, func(ctx context.Context) error {
					if ready {
						proof := lease
						data, err := harnesslocal.ReadFile(proof, 1048576, false)
						if err != nil {
							closeErr = errors.Join(err, stop(ctx))
							return closeErr
						}
						snapshot := proof + ".cleanup-" + identifier()
						if err := harnesslocal.PublishNew(snapshot, data); err != nil {
							closeErr = errors.Join(err, stop(ctx))
							return closeErr
						}
						closeErr = objectstoreproxy.Stop(snapshot)
						closeErr = errors.Join(closeErr, removePrivate(snapshot))
					}
					closeErr = errors.Join(closeErr, stop(ctx))
					if closeErr != nil {
						return closeErr
					}
					for _, proof := range []string{lease, attempt} {
						if err := options.Runtime.RecordResource("object_store_proxy", proof, "released"); err != nil {
							closeErr = err
							return closeErr
						}
						if err := removePrivate(proof); err != nil {
							closeErr = err
							return closeErr
						}
					}
					return closeErr
				})
				return closeErr
			})
		})
		return CleanupFailure(closeErr)
	}
	defer func() {
		if result != nil {
			cleanup, done := context.WithTimeout(context.Background(), 30*time.Second)
			defer done()
			result = errors.Join(result, close(cleanup))
		}
	}()
	for i := 0; i < 200; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if command(ctx, "status", "--state-file", attempt) == nil {
			ready = true
			break
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	if !ready {
		return nil, &Failure{Class: "infra", Reason: "service_readiness_timeout"}
	}
	// Promotion is transactional; pending intent preserves interruption before the
	// ready proof becomes visible. Keep the attempt until release acknowledgement.
	if err := options.Runtime.RecordDependentResource("object_store_proxy", lease, "pending", options.Dependencies, trust); err != nil {
		return nil, artifactFailure(err)
	}
	if err := command(ctx, "promote", "--attempt-file", attempt, "--lease-file", lease); err != nil {
		ready = false
		return nil, err
	}
	if err := options.Runtime.RecordResource("object_store_proxy", lease, "acquired"); err != nil {
		return nil, artifactFailure(err)
	}
	if err := options.Runtime.RecordResource("object_store_proxy", attempt, "released"); err != nil {
		return nil, artifactFailure(err)
	}
	resourceTarget = lease
	resourceRef, err := options.Runtime.ResourceReference("object_store_proxy", lease)
	if err != nil {
		return nil, err
	}
	attached := maps.Clone(suite.Environment)
	attached["CARTULARY_HARNESS_RESOURCE_DEPENDENCIES"] = resourceDependencies([]string{resourceRef})
	maps.Copy(attached, map[string]string{"OBJECT_STORE_ENDPOINT": listen, "SEAWEEDFS_S3_ACCESS_KEY_ID": suite.Environment["CARTULARY_S3TEST_ACCESS_KEY_ID"], "SEAWEEDFS_S3_SECRET_ACCESS_KEY": suite.Environment["CARTULARY_S3TEST_SECRET_ACCESS_KEY"], "OBJECT_STORE_SECURE": "true", "OBJECT_STORE_CORS_ORIGIN": "https://localhost:5173"})
	if err := options.Runtime.RegisterEnvironment(attached); err != nil {
		return nil, artifactFailure(err)
	}
	return &Allocation{Ownership: "owned", ResourceIDs: []string{"suite:object_store_namespace", "object-store-proxy:" + id}, Environment: attached, Release: func(ctx context.Context, _ bool) error { return close(ctx) }}, nil
}
