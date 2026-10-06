// browserpki provisions only disposable Make-owned browser fixture identities.
package main

import (
	"flag"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/browserpki"
	"os"
)

func main() {
	directory := flag.String("directory", "", "new private TLS directory beneath the owned browser runtime")
	flag.Parse()
	if err := browserpki.Provision(*directory); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "browser fixture TLS provisioning failed:", err)
		os.Exit(1)
	}
}
