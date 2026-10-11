package main

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"os"
)

func main() { os.Exit(harnessadmission.Main(os.Args[1:], os.Stdout)) }
