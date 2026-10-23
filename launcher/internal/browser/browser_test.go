package browser

import "testing"

// Getting this wrong is not a crash: it is a browser opening inside the
// distribution, where the player cannot see it, and a launcher that looks like
// it did nothing.
func TestWSLIsSpottedByEitherAnswer(t *testing.T) {
	for name, spotted := range map[string]struct {
		distro, version string
		want            bool
	}{
		"the shell set the distribution": {distro: "Ubuntu", want: true},
		"only /proc/version names it": {
			version: "Linux version 6.6.87.2-microsoft-standard-WSL2 (root@build)",
			want:    true,
		},
		"the word WSL alone is enough": {version: "Linux version 5.15.0-wsl2", want: true},
		"an ordinary Linux": {
			version: "Linux version 6.12.105+deb13-amd64 (debian-kernel@lists.debian.org)",
			want:    false,
		},
		"nothing to go on": {want: false},
	} {
		if got := UnderWSL(spotted.distro, spotted.version); got != spotted.want {
			t.Errorf("%s: UnderWSL(%q, %q) = %v, want %v", name, spotted.distro, spotted.version, got, spotted.want)
		}
	}
}
