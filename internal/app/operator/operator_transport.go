package operator

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type operatorPostgresPool interface {
	postgres.DB
	Close()
}

type operatorTransport struct {
	stdout io.Writer
	stderr io.Writer
}

func (transport operatorTransport) encodeJSON(payload any) error {
	encoder := json.NewEncoder(normalizeOperatorWriter(transport.stdout))
	if err := encoder.Encode(payload); err != nil {
		return fmt.Errorf("encode operator JSON: %w", err)
	}
	return nil
}

func (transport operatorTransport) logger() *slog.Logger {
	return operatorLogger(transport.stderr)
}

func operatorLogger(stderr io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(normalizeOperatorWriter(stderr), &slog.HandlerOptions{Level: slog.LevelError}))
}

func normalizeOperatorWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}
	return writer
}
