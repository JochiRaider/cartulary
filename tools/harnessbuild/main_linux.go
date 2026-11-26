//go:build linux

// Command harnessbuild produces build artifacts through the harness artifact
// store (P5-3b). Make runs it with `go run`, which needs no prebuilt binary, so
// it also builds the harness runtime itself.
//
//	harnessbuild go-binary [--store DIR] --package PATH --output FILE [--tags TAGS] [--label LABEL]
//
// The store defaults to the workspace layout's artifact_store directory.
// The binary and its receipt are published at FILE and FILE.crypto.json. With
// --label and RUN_STEP_SCRIPT, a build runs as that step; a reused root runs
// none.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/artifactstore"
)

// collection bounds: records unused for 30 days go, and one build removes at
// most 16 entries.
const (
	collectAge   = 30 * 24 * time.Hour
	collectLimit = 16
)

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 || args[0] != "go-binary" {
		fmt.Fprintln(os.Stderr, "usage: harnessbuild go-binary [--store DIR] --package PATH --output FILE [--tags TAGS] [--label LABEL]")
		return 2
	}
	flags := flag.NewFlagSet("go-binary", flag.ContinueOnError)
	store := flags.String("store", "", "artifact store directory")
	pkg := flags.String("package", "", "package to build")
	output := flags.String("output", "", "path the binary is published at")
	tags := flags.String("tags", "", "build tags")
	label := flags.String("label", "", "step label for a build")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 || *pkg == "" || *output == "" {
		fmt.Fprintln(os.Stderr, "usage: harnessbuild go-binary [--store DIR] --package PATH --output FILE [--tags TAGS] [--label LABEL]")
		return 2
	}
	if *store == "" {
		directory, err := layoutStore()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		*store = directory
	}
	if err := goBinary(*store, *pkg, *output, *tags, *label); err != nil {
		fmt.Fprintln(os.Stderr, err)
		reportFailure()
		return 1
	}
	return 0
}

func goBinary(store, pkg, output, tags, label string) error {
	module, err := os.Getwd()
	if err != nil {
		return err
	}
	node := os.Getenv("NODE_BIN")
	if node == "" {
		return errors.New("NODE_BIN is required")
	}
	goExecutable := os.Getenv("GO")
	if goExecutable == "" {
		goExecutable = "go"
	}
	env := os.Environ()
	for name, variable := range map[string]string{"GOCACHE": "GO_CACHE_DIR", "GOMODCACHE": "GO_MOD_CACHE_DIR", "GOTMPDIR": "GO_TMP_DIR"} {
		directory := os.Getenv(variable)
		if directory == "" {
			return fmt.Errorf("%s is required", variable)
		}
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return err
		}
		env = append(env, name+"="+directory)
	}
	producer := artifactstore.GoBinary{Module: module, Package: pkg, Tags: tags, Name: filepath.Base(output), Go: goExecutable, Node: node, Env: env}
	if script := os.Getenv("RUN_STEP_SCRIPT"); label != "" && script != "" {
		if os.Getenv("CARTULARY_TEST_TARGET") == "" {
			return errors.New("a labelled build requires CARTULARY_TEST_TARGET")
		}
		producer.Env = append(producer.Env, "CARTULARY_SUPPRESS_CHILD_SUCCESS=1")
		producer.Wrap = func(command []string) []string { return append([]string{"bash", script, label, "--"}, command...) }
	}
	name := producer.Name
	s := artifactstore.Store{Root: store}
	if _, err := s.Produce(producer, map[string]string{name: output, name + ".crypto.json": output + ".crypto.json"}, artifactstore.Controls{
		Disable: os.Getenv("CARTULARY_BUILD_CACHE_DISABLE") == "1",
		Force:   os.Getenv("CARTULARY_FORCE_REBUILD") == "1",
	}); err != nil {
		return err
	}
	_, err = s.Collect(time.Now(), collectAge, collectLimit)
	return err
}

// layoutStore reads the store's location from the authored workspace layout,
// so the first build of the harness runtime needs no generated Make include.
func layoutStore() (string, error) {
	data, err := os.ReadFile("tools/workspace_layout.json")
	if err != nil {
		return "", err
	}
	var layout struct {
		Directories []struct {
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"directories"`
	}
	if err := json.Unmarshal(data, &layout); err != nil {
		return "", err
	}
	for _, directory := range layout.Directories {
		if directory.ID == "artifact_store" && directory.Path != "" && !filepath.IsAbs(directory.Path) {
			return filepath.Abs(directory.Path)
		}
	}
	return "", errors.New("the workspace layout declares no artifact_store directory")
}

// reportFailure records a build failure in the command's failure context, as
// the shell build helper did.
func reportFailure() {
	if os.Getenv("CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT") == "" || os.Getenv("NODE_BIN") == "" {
		return
	}
	command := exec.Command(os.Getenv("NODE_BIN"), "tools/harness/runtime/command-failure-cli.mjs", "harness", "tool_diagnostic_failure")
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	_ = command.Run()
}
