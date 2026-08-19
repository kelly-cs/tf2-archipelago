//go:build windows

package assets

import _ "embed"

//go:embed embedded/sm-ripext-windows.zip
var ripextZip []byte

//go:embed embedded/defender-bots-windows.zip
var defenderBotsZip []byte
