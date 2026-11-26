//go:build linux

package artifactstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const installationReceiptSchemaID = "cartulary.installation_receipt.v1"

// Installation is a producer that installs in place: the Node runtime, the
// workspace's frontend dependencies or the Playwright browsers. Tools write
// and read these at fixed paths, so they cannot be immutable store roots.
// Instead, a receipt written last beside the installation records what it was
// installed from, and is its only key record: a receipt describes what is
// installed now, so an installation replaced by another never reuses a stale
// record.
type Installation interface {
	ID() string
	// Current reports whether the installation in place matches its inputs.
	Current() (bool, error)
	// Install installs in place and writes the receipt last.
	Install() error
}

// Install brings inst up to date while holding lock: it reuses a current
// installation unless the controls bypass reuse. The installed-only
// preparation policy installs nothing, so a missing or stale installation
// fails instead. It reports whether the installation was reused.
func Install(inst Installation, lock string, controls Controls, installedOnly bool) (bool, error) {
	if err := os.MkdirAll(filepath.Dir(lock), 0o755); err != nil {
		return false, err
	}
	unlock, err := lockFile(lock)
	if err != nil {
		return false, err
	}
	defer unlock()
	if installedOnly || !controls.Disable && !controls.Force {
		current, err := inst.Current()
		if err != nil {
			return false, fmt.Errorf("check the %s installation: %w", inst.ID(), err)
		}
		if current {
			return true, nil
		}
	}
	if installedOnly {
		return false, fmt.Errorf("the %s installation is missing or stale, and the installed-only policy installs nothing; run make bootstrap", inst.ID())
	}
	return false, inst.Install()
}

type installationReceipt struct {
	SchemaID string `json:"schema_id"`
	Producer string `json:"producer"`
	Key      string `json:"key"`
}

// receiptMatches reports whether the receipt at name records key for producer.
func receiptMatches(name, producer, key string) bool {
	data, err := os.ReadFile(name)
	if err != nil {
		return false
	}
	var r installationReceipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(&r) == nil && r == installationReceipt{SchemaID: installationReceiptSchemaID, Producer: producer, Key: key}
}

// writeReceipt records key for producer at name, replacing it by rename.
func writeReceipt(name, producer, key string) error {
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(installationReceipt{SchemaID: installationReceiptSchemaID, Producer: producer, Key: key})
	if err != nil {
		return err
	}
	return writeAtomic(name, append(data, '\n'))
}

// nodeEnvironment is env with the Node runtime's executables first on PATH
// and its Corepack home, as every pnpm invocation needs.
func nodeEnvironment(env []string, runtime string) []string {
	return append(env,
		"PATH="+filepath.Join(runtime, "bin")+string(os.PathListSeparator)+environmentValue(env, "PATH"),
		"COREPACK_HOME="+filepath.Join(runtime, "corepack"))
}

// executableIn resolves name on env's PATH, where the command it names runs,
// rather than on this process's own.
func executableIn(env []string, name string) string {
	for _, directory := range filepath.SplitList(environmentValue(env, "PATH")) {
		candidate := filepath.Join(directory, name)
		if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
			return candidate
		}
	}
	return name
}

// probe runs a readiness probe and reports whether it succeeded. A probe that
// cannot start is an error; one that runs and fails means "not current".
func probe(command *exec.Cmd) (bool, error) {
	var stderr bytes.Buffer
	command.Stderr = &stderr
	err := command.Run()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%s: %w: %s", strings.Join(command.Args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return true, nil
}
