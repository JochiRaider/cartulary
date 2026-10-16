// Package rendersandbox is the composition facade for cmd/render-sandbox.
package rendersandbox

import (
	"fmt"
	"io"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

// Run is entered only inside the supervisor's namespace and mount boundary.
// It has no application configuration, database, credentials or module wiring.
// Executable paths are installation capabilities, never stage-request fields.
func Run(arguments, environment []string, diagnostic io.Writer) int {
	if len(arguments) != 2 || arguments[0] != "/runtime/node/bin/node" || arguments[1] != "/runtime/renderer/src/stage-main.ts" {
		_, _ = fmt.Fprintln(diagnostic, "report-render-sandbox: invalid installation command")
		return 2
	}
	if err := renderprocess.ExecNetworkChild(3, arguments, environment); err != nil {
		_, _ = fmt.Fprintln(diagnostic, "report-render-sandbox: confinement unavailable")
		return 1
	}
	return 0
}
