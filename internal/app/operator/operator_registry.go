package operator

import (
	"context"
	"fmt"
	"io"
	"strings"
)

type operatorCommandHandler func(context.Context, []string) int

type operatorCommandDescriptor struct {
	Tokens           []string
	Usage            string
	Run              operatorCommandHandler
	InvalidNamespace operatorCommandHandler
}

type operatorCommandRegistry struct {
	commands           []operatorCommandDescriptor
	invalidByNamespace map[string]operatorCommandHandler
	stderr             io.Writer
}

func newOperatorCommandRegistry(stderr io.Writer, commands []operatorCommandDescriptor) (operatorCommandRegistry, error) {
	registry := operatorCommandRegistry{
		commands:           append([]operatorCommandDescriptor(nil), commands...),
		invalidByNamespace: make(map[string]operatorCommandHandler),
		stderr:             normalizeOperatorWriter(stderr),
	}
	seen := make(map[string]struct{}, len(registry.commands))
	for index, command := range registry.commands {
		if len(command.Tokens) == 0 || strings.TrimSpace(command.Usage) == "" || command.Run == nil {
			return operatorCommandRegistry{}, fmt.Errorf("operator command descriptor %d is incomplete", index)
		}
		for _, token := range command.Tokens {
			if strings.TrimSpace(token) == "" || strings.ContainsAny(token, " \t\n") {
				return operatorCommandRegistry{}, fmt.Errorf("operator command descriptor %d has invalid token", index)
			}
		}
		key := strings.Join(command.Tokens, "\x00")
		if _, exists := seen[key]; exists {
			return operatorCommandRegistry{}, fmt.Errorf("duplicate operator command path %q", strings.Join(command.Tokens, " "))
		}
		for _, previous := range registry.commands[:index] {
			if commandPathPrefixes(command.Tokens, previous.Tokens) || commandPathPrefixes(previous.Tokens, command.Tokens) {
				return operatorCommandRegistry{}, fmt.Errorf("prefix-ambiguous operator command paths %q and %q", strings.Join(previous.Tokens, " "), strings.Join(command.Tokens, " "))
			}
		}
		seen[key] = struct{}{}
		if command.InvalidNamespace != nil {
			registry.invalidByNamespace[command.Tokens[0]] = command.InvalidNamespace
		}
	}
	return registry, nil
}

func (registry operatorCommandRegistry) run(ctx context.Context, args []string) int {
	for _, command := range registry.commands {
		if commandPathPrefixes(command.Tokens, args) {
			return command.Run(ctx, args)
		}
	}
	if len(args) > 0 {
		if invalid := registry.invalidByNamespace[args[0]]; invalid != nil {
			return invalid(ctx, args)
		}
	}
	_, _ = fmt.Fprintln(registry.stderr, registry.usage())
	return 2
}

func (registry operatorCommandRegistry) usage() string {
	lines := make([]string, 0, len(registry.commands)+1)
	lines = append(lines, "usage:")
	for _, command := range registry.commands {
		lines = append(lines, "  "+command.Usage)
	}
	return strings.Join(lines, "\n")
}

func commandPathPrefixes(prefix []string, tokens []string) bool {
	if len(prefix) > len(tokens) {
		return false
	}
	for index := range prefix {
		if prefix[index] != tokens[index] {
			return false
		}
	}
	return true
}
