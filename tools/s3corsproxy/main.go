package main

import (
	"github.com/JochiRaider/cartulary/internal/testutil/objectstoreproxy"
	"os"
)

func main() { os.Exit(objectstoreproxy.Main(os.Args[1:])) }
