"""Write the PE checksum the Go linker leaves at zero.

Windows only enforces this field for drivers and boot-time DLLs, so Go ignores
it. Scanners do not: every binary built by Microsoft's toolchain carries a
correct checksum, so zero is one more way this exe fails to look like the
software it is. signtool recomputes the field when it signs, which is the same
reason it is worth writing here while nothing signs the exe.
"""

import sys

import pefile

if len(sys.argv) != 2:
    sys.exit("usage: pe-checksum.py <exe>")

path = sys.argv[1]
pe = pefile.PE(path)
want = pe.generate_checksum()

if pe.OPTIONAL_HEADER.CheckSum == want:
    print(f"{path}: checksum already {want:#x}")
    sys.exit(0)

pe.OPTIONAL_HEADER.CheckSum = want
pe.write(path)
print(f"{path}: checksum written, {want:#x}")
