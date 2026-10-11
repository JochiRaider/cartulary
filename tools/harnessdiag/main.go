package main

import (
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"os"
)

func main() { os.Exit(harnessdiag.Main(os.Args[1:], os.Stdout, os.Stderr)) }
