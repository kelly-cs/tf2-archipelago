//go:build !windows

package assets

import _ "embed"

//go:embed embedded/sm-ripext-linux.zip
var ripextZip []byte

//go:embed embedded/defender-bots-linux.zip
var defenderBotsZip []byte
