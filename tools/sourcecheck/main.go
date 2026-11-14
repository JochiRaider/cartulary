// sourcecheck is the private batch parser used by harness inspection.
package main

import (
	"fmt"
	"os"

	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "sourcecheck accepts a JSON batch on stdin")
		os.Exit(1)
	}
	if err := sourcecheck.InspectBatch(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
