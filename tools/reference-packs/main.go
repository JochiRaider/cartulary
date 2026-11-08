// The producer executable only composes the owner-local offline command.
package main

import (
	"context"
	"os"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
)

func main() {
	os.Exit(producer.RunCommand(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
