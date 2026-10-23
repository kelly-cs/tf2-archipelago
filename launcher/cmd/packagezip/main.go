// Command packagezip builds the small ZIP artifacts the standalone launcher
// needs: the packaged apworld and the bot mod's file tree. Standard library
// only, so a WSL checkout builds tf2ap.exe without Docker or a zip command.
//
// Usage:
//
//	packagezip apworld <source dir> <output.apworld> -container-version 7
//	packagezip tree <source dir> <output.zip> [-exclude-suffix .so]...
package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "packagezip:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: packagezip apworld|tree <source> <output> [flags]")
	}
	mode, source, output, rest := args[0], args[1], args[2], args[3:]
	flags := flag.NewFlagSet("packagezip "+mode, flag.ContinueOnError)
	switch mode {
	case "apworld":
		version := flags.Int("container-version", 0, "the APWorldContainer format version to stamp")
		if err := flags.Parse(rest); err != nil {
			return err
		}
		if *version == 0 {
			return fmt.Errorf("apworld needs -container-version")
		}
		return buildAPWorld(source, output, *version)
	case "tree":
		var excluded suffixes
		flags.Var(&excluded, "exclude-suffix", "leave out files ending in this (repeatable)")
		if err := flags.Parse(rest); err != nil {
			return err
		}
		return buildTree(source, output, excluded)
	default:
		return fmt.Errorf("unknown mode %q: apworld or tree", mode)
	}
}

type suffixes []string

func (s *suffixes) String() string     { return fmt.Sprint([]string(*s)) }
func (s *suffixes) Set(v string) error { *s = append(*s, v); return nil }
