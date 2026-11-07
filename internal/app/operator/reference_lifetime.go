package operator

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/processlifecycle"
)

func referenceVerificationLifetime(grace time.Duration, stderr io.Writer) (*reference_data.VerificationLifetime, error) {
	lifecycle := processlifecycle.New()
	return reference_data.NewVerificationLifetime(grace, func(error) {
		if lifecycle.Fatal("operation_cancellation_grace_exhausted") {
			_, _ = fmt.Fprintln(stderr, `{"code":"extension_integrity_failure","reason_code":"operation_cancellation_grace_exhausted"}`)
		}
		// A blocked operation cannot unwind to the ordinary CLI result path.
		os.Exit(70)
	})
}
