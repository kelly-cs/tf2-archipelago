package tfschema

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf16"
)

// tokenLine is one `"token" "text"` pair, which is every line of the
// localisation file that names something.
var tokenLine = regexp.MustCompile(`^\s*"([^"\\]+)"\s+"((?:\\.|[^"\\])*)"`)

// English reads tf_english.txt into lower-cased token to text. The file is
// UTF-16 with a byte order mark, and the schema names tokens in either case.
func English(path string) (map[string]string, error) {
	body, err := os.ReadFile(path) //nolint:gosec // the path is a flag on a maintainer tool
	if err != nil {
		return nil, err
	}
	text, err := decodeUTF16(body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	names := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(text))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		if m := tokenLine.FindStringSubmatch(scanner.Text()); m != nil {
			names[strings.ToLower(m[1])] = strings.ReplaceAll(m[2], `\"`, `"`)
		}
	}
	return names, scanner.Err()
}

func decodeUTF16(body []byte) (string, error) {
	if len(body) < 2 || len(body)%2 != 0 {
		return "", fmt.Errorf("not UTF-16: %d bytes", len(body))
	}
	var order binary.ByteOrder
	switch {
	case body[0] == 0xFF && body[1] == 0xFE:
		order = binary.LittleEndian
	case body[0] == 0xFE && body[1] == 0xFF:
		order = binary.BigEndian
	default:
		return "", fmt.Errorf("not UTF-16: no byte order mark")
	}
	units := make([]uint16, 0, (len(body)-2)/2)
	for i := 2; i < len(body); i += 2 {
		units = append(units, order.Uint16(body[i:]))
	}
	return string(utf16.Decode(units)), nil
}
