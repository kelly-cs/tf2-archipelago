package main

import (
	"fmt"
	"os"

	"github.com/m-this/tf2-archipelago/launcher/internal/sigmodpatch"
)

func main() {
	if len(os.Args) != 3 || os.Args[2] != "linux" {
		fmt.Fprintln(os.Stderr, "usage: sigmodpatch <unpacked-sigmod-directory> linux")
		os.Exit(2)
	}
	if err := sigmodpatch.Patch(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
