package main

import (
	"os"

	"github.com/JochiRaider/cartulary/internal/app/rendersandbox"
)

func main() {
	os.Exit(rendersandbox.Run(os.Args[1:], os.Environ(), os.Stderr))
}
