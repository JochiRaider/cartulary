package main

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnesscollect"
	"os"
)

func main() {
	if len(os.Args) != 1 {
		os.Exit(2)
	}
	os.Exit(harnesscollect.Main(os.Stdin, os.Stdout))
}
