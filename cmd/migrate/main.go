package main

import (
	"os"

	"github.com/JochiRaider/cartulary/internal/app"
)

func main() {
	os.Exit(app.RunMigrateCLI(os.Args[1:], os.Stderr))
}
