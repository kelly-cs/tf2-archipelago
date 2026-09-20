package installer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// SigMod passes an empty phrase name to the final %t while drawing an
// unselected extra loadout item. SourceMod then returns the entire format
// string literally. The pinned native package can be corrected in place by
// changing that one placeholder to %s; both strings have the same length.
var (
	sigmodLoadoutMenuBroken = []byte("%t: %s %s %t\x00")
	sigmodLoadoutMenuFixed  = []byte("%t: %s %s %s\x00")
	sigmodTitleBroken       = []byte("Extra loadout items class\x00")
	sigmodTitleFixed        = append([]byte("Extra loadout items\x00"), make([]byte, 6)...)
)

type sigmodCallPatch struct{ old, fixed []byte }

// Both title call sites in the SHA-pinned 20250703 Linux package invoke the
// variadic translation formatter. It returns "%s" for this title. Calling
// the plain translation helper with the existing "Extra loadout items" key
// gives a readable, localized title. These byte sequences are unique in each
// pinned extension; an updated package must be assessed before patching.
var sigmodTitleCalls = map[string][]sigmodCallPatch{
	"x86": {
		{[]byte("\xe8\x5f\xc2\xd4\xff"), []byte("\xe8\x9f\xc1\xd4\xff")},
		{[]byte("\xe8\x55\x19\xd7\xff"), []byte("\xe8\x95\x18\xd7\xff")},
	},
	"x64": {
		{[]byte("\xe8\x4b\x85\xd2\xff"), []byte("\xe8\x7b\x84\xd2\xff")},
		{[]byte("\xe8\xa8\x06\xd5\xff"), []byte("\xe8\xd8\x05\xd5\xff")},
	},
}

func sigmodExtensionPaths(goos string) []string {
	base := "addons/sourcemod/extensions/"
	if goos == "windows" {
		return []string{base + "sigsegv.ext.2.tf2.dll"}
	}
	return []string{base + "sigsegv.ext.2.tf2.so", base + "x64/sigsegv.ext.2.tf2.so"}
}

func sigmodLoadoutMenuPatched(modDir, goos string) bool {
	for _, relative := range sigmodExtensionPaths(goos) {
		body, err := os.ReadFile(filepath.Join(modDir, filepath.FromSlash(relative)))
		if err != nil || bytes.Count(body, sigmodLoadoutMenuFixed) != 1 || bytes.Contains(body, sigmodLoadoutMenuBroken) {
			return false
		}
		if goos != "windows" {
			arch := "x86"
			if filepath.Dir(relative) == "addons/sourcemod/extensions/x64" {
				arch = "x64"
			}
			if bytes.Count(body, sigmodTitleFixed) != 1 || bytes.Contains(body, sigmodTitleBroken) {
				return false
			}
			for _, call := range sigmodTitleCalls[arch] {
				if bytes.Count(body, call.fixed) != 1 || bytes.Contains(body, call.old) {
					return false
				}
			}
		}
	}
	return true
}

func patchSigmodLoadoutMenu(modDir, goos string) error {
	for _, relative := range sigmodExtensionPaths(goos) {
		path := filepath.Join(modDir, filepath.FromSlash(relative))
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fixed := body
		switch {
		case bytes.Count(fixed, sigmodLoadoutMenuBroken) == 1 && !bytes.Contains(fixed, sigmodLoadoutMenuFixed):
			fixed = bytes.Replace(fixed, sigmodLoadoutMenuBroken, sigmodLoadoutMenuFixed, 1)
		case bytes.Count(fixed, sigmodLoadoutMenuFixed) == 1 && !bytes.Contains(fixed, sigmodLoadoutMenuBroken):
		default:
			return fmt.Errorf("SigMod extra loadout menu marker is missing or ambiguous in %s", relative)
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if goos != "windows" {
			arch := "x86"
			if filepath.Dir(relative) == "addons/sourcemod/extensions/x64" {
				arch = "x64"
			}
			switch {
			case bytes.Count(fixed, sigmodTitleBroken) == 1 && !bytes.Contains(fixed, sigmodTitleFixed):
				fixed = bytes.Replace(fixed, sigmodTitleBroken, sigmodTitleFixed, 1)
			case bytes.Count(fixed, sigmodTitleFixed) == 1 && !bytes.Contains(fixed, sigmodTitleBroken):
			default:
				return fmt.Errorf("SigMod title marker is missing or ambiguous in %s", relative)
			}
			for _, call := range sigmodTitleCalls[arch] {
				switch {
				case bytes.Count(fixed, call.old) == 1 && !bytes.Contains(fixed, call.fixed):
					fixed = bytes.Replace(fixed, call.old, call.fixed, 1)
				case bytes.Count(fixed, call.fixed) == 1 && !bytes.Contains(fixed, call.old):
				default:
					return fmt.Errorf("SigMod title call is missing or ambiguous in %s", relative)
				}
			}
		}
		if bytes.Equal(body, fixed) {
			continue
		}
		if err := os.WriteFile(path, fixed, info.Mode().Perm()); err != nil {
			return fmt.Errorf("cannot patch SigMod extra loadout menu in %s: %w", relative, err)
		}
	}
	return nil
}
