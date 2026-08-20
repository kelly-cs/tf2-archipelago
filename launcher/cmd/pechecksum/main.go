// Command pechecksum writes the PE header checksum into a Windows executable.
//
// Windows enforces the field only for drivers and boot-time DLLs, so the Go
// linker leaves it at zero. Scanners are less relaxed: everything built by
// Microsoft's toolchain carries a correct checksum, so zero is one more way a
// Go binary fails to look like the software it is. signtool recomputes the
// field when it signs, which is the same reason this exists while nothing
// signs the exe.
package main

import (
	"encoding/binary"
	"fmt"
	"os"
)

// checksumOffset is where CheckSum sits inside the optional header. The same
// offset in PE32 and PE32+, which is why the magic never has to be read.
const checksumOffset = 64

// peOffsetAt holds e_lfanew, the offset of the PE signature.
const peOffsetAt = 0x3c

// coffToOptional is the PE signature plus the COFF header.
const coffToOptional = 4 + 20

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: pechecksum <exe>")
		os.Exit(2)
	}
	path := os.Args[1]

	image, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	field, err := checksumField(image)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
		os.Exit(1)
	}

	want := checksum(image, field)
	if binary.LittleEndian.Uint32(image[field:]) == want {
		fmt.Printf("%s: checksum already %#x\n", path, want)
		return
	}

	binary.LittleEndian.PutUint32(image[field:], want)
	if err := os.WriteFile(path, image, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("%s: checksum written, %#x\n", path, want)
}

// checksumField returns the offset of CheckSum, or an error if the file is not
// a PE image. Every read is bounds-checked: this parses a file, and the file is
// the only thing that says how long it claims to be.
func checksumField(image []byte) (int, error) {
	if len(image) < peOffsetAt+4 || image[0] != 'M' || image[1] != 'Z' {
		return 0, fmt.Errorf("not a PE image: no MZ header")
	}
	pe := int(binary.LittleEndian.Uint32(image[peOffsetAt:]))
	field := pe + coffToOptional + checksumOffset
	if pe < 0 || field+4 > len(image) {
		return 0, fmt.Errorf("not a PE image: header at %#x runs past the file", pe)
	}
	if string(image[pe:pe+4]) != "PE\x00\x00" {
		return 0, fmt.Errorf("not a PE image: no PE signature at %#x", pe)
	}
	return field, nil
}

// checksum is the 16-bit ones-complement sum of the image with the CheckSum
// field read as zero, folded to 16 bits, plus the length of the file.
func checksum(image []byte, field int) uint32 {
	var sum uint32
	for i := 0; i+1 < len(image); i += 2 {
		if i == field || i == field+2 {
			continue
		}
		sum += uint32(binary.LittleEndian.Uint16(image[i:]))
		sum = (sum & 0xffff) + (sum >> 16)
	}
	if len(image)%2 == 1 {
		sum += uint32(image[len(image)-1])
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return sum + uint32(len(image))
}
