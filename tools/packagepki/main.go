// packagepki is a Make-owned disposable package fixture composition root.
package main

import (
	"flag"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/packagepki"
	"os"
)

func main() {
	directory := flag.String("directory", "", "new private TLS directory beneath an owned package workspace")
	flag.Parse()
	if err := packagepki.Provision(*directory); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "package fixture TLS provisioning failed:", err)
		os.Exit(1)
	}
}
