package operator

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestOperatorCommandRegistryRejectsDuplicateAndPrefixAmbiguousPaths(t *testing.T) {
	run := func(context.Context, []string) int { return 0 }
	for _, test := range []struct {
		name       string
		descriptor operatorCommandDescriptor
	}{
		{name: "empty tokens", descriptor: operatorCommandDescriptor{Usage: "operator object-store init", Run: run}},
		{name: "blank token", descriptor: operatorCommandDescriptor{Tokens: []string{"object-store", ""}, Usage: "operator object-store init", Run: run}},
		{name: "whitespace token", descriptor: operatorCommandDescriptor{Tokens: []string{"object-store", "init command"}, Usage: "operator object-store init", Run: run}},
		{name: "blank usage", descriptor: operatorCommandDescriptor{Tokens: []string{"object-store", "init"}, Usage: " ", Run: run}},
		{name: "nil handler", descriptor: operatorCommandDescriptor{Tokens: []string{"object-store", "init"}, Usage: "operator object-store init"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := newOperatorCommandRegistry(nil, []operatorCommandDescriptor{test.descriptor}); err == nil {
				t.Fatal("registry accepted an incomplete command descriptor")
			}
		})
	}
	for _, commands := range [][]operatorCommandDescriptor{
		{
			{Tokens: []string{"object-store", "init"}, Usage: "operator object-store init", Run: run},
			{Tokens: []string{"object-store", "init"}, Usage: "operator object-store init", Run: run},
		},
		{
			{Tokens: []string{"object-store"}, Usage: "operator object-store", Run: run},
			{Tokens: []string{"object-store", "init"}, Usage: "operator object-store init", Run: run},
		},
	} {
		if _, err := newOperatorCommandRegistry(nil, commands); err == nil {
			t.Fatal("registry accepted ambiguous command paths")
		}
	}
}

func TestOperatorCommandRegistryRoutesExactAndCanonicalNamespaceFailures(t *testing.T) {
	var stderr bytes.Buffer
	var exactArgs []string
	var invalidArgs []string
	exact := func(_ context.Context, args []string) int {
		exactArgs = append([]string(nil), args...)
		return 7
	}
	invalid := func(_ context.Context, args []string) int {
		invalidArgs = append([]string(nil), args...)
		return 2
	}
	registry, err := newOperatorCommandRegistry(&stderr, []operatorCommandDescriptor{
		{
			Tokens:           []string{"backup", "create"},
			Usage:            "operator backup create",
			Run:              exact,
			InvalidNamespace: invalid,
		},
	})
	if err != nil {
		t.Fatalf("build registry: %v", err)
	}
	if got := registry.run(context.Background(), []string{"backup", "create", "--progress=jsonl"}); got != 7 {
		t.Fatalf("exact exit code got %d want 7", got)
	}
	if got, want := strings.Join(exactArgs, " "), "backup create --progress=jsonl"; got != want {
		t.Fatalf("exact handler args got %q want %q", got, want)
	}
	if got := registry.run(context.Background(), []string{"backup", "unsupported"}); got != 2 {
		t.Fatalf("canonical namespace exit code got %d want 2", got)
	}
	if got, want := strings.Join(invalidArgs, " "), "backup unsupported"; got != want {
		t.Fatalf("invalid handler args got %q want %q", got, want)
	}
	if got := registry.run(context.Background(), []string{"unknown-root", "latest"}); got != 2 {
		t.Fatalf("unknown top-level exit code got %d want 2", got)
	}
	if !strings.Contains(stderr.String(), "usage:") || strings.Contains(stderr.String(), "unknown-root") {
		t.Fatalf("global usage did not reject unknown top-level name cleanly: %q", stderr.String())
	}
}

func TestOperatorCommandRegistryContainsExactlyEightCanonicalPaths(t *testing.T) {
	registry, err := (operatorRunner{}).commandRegistry()
	if err != nil {
		t.Fatalf("build operator registry: %v", err)
	}
	want := []string{
		"backup inspect latest",
		"backup create",
		"restore latest",
		"restore-verify latest",
		"restore-verify due",
		"migration-evidence capture",
		"object-store init",
		"collaboration requeue",
	}
	if len(registry.commands) != len(want) {
		t.Fatalf("operator command count got %d want %d", len(registry.commands), len(want))
	}
	for index, command := range registry.commands {
		if got := strings.Join(command.Tokens, " "); got != want[index] {
			t.Fatalf("operator command %d got %q want %q", index, got, want[index])
		}
	}
	usage := registry.usage()
	if strings.Count(usage, "\n  operator ") != len(want) {
		t.Fatalf("operator usage does not contain exactly eight command lines: %q", usage)
	}
	if !strings.Contains(usage, "\n  "+collaborationRequeueUsage) || strings.Contains(usage, "collaboration requeue --incident-id <uuid> [-config") {
		t.Fatalf("operator usage does not expose only the strict Collaboration v2 grammar: %q", usage)
	}
}
