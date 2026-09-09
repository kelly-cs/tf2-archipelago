/*
Package spa serves the browser interface out of the binary.

launcher/web's `ng build` writes straight into dist/ beside this file, and
//go:embed all:dist snapshots it at compile time, so one file ships the whole
launcher. `make web-build` is what puts a real build there; `make
embed-placeholders` leaves a holding page saying so, because Go refuses to
compile a package whose embed pattern matches nothing.

Routing: a file that exists under dist/ is served as itself, and every other
path gets index.html, which is what makes a refresh on /session work. The
launcher answers on loopback for one player, so there is no cache to reason
about beyond the two rules the build's own file names ask for: a hashed asset
never changes, and index.html always may have.
*/
package spa

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var embedded embed.FS

// immutableMax is a year, the longest max-age RFC 9111 asks a cache to accept.
// It is only ever sent for a file whose name carries a hash of its contents.
const immutableMax = "public, max-age=31536000, immutable"

// Handler serves the embedded app. Register it last on the mux: the Connect
// paths are matched first and everything the browser asks for lands here.
func Handler() http.Handler {
	root, err := fs.Sub(embedded, "dist")
	if err != nil {
		// The embed directive proves dist/ exists at compile time, so this is
		// unreachable. Answering rather than panicking keeps a launcher that
		// somehow got here able to say what is wrong.
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "the browser interface is not in this binary", http.StatusInternalServerError)
		})
	}
	return handlerFor(root)
}

func handlerFor(root fs.FS) http.Handler {
	files := http.FileServerFS(root)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name != "" && exists(root, name) {
			if hashed(name) {
				w.Header().Set("Cache-Control", immutableMax)
			}
			files.ServeHTTP(w, r)
			return
		}
		// A client-side route. The app owns it, so serve the page that boots it.
		w.Header().Set("Cache-Control", "no-cache")
		index := r.Clone(r.Context())
		index.URL.Path = "/"
		files.ServeHTTP(w, index)
	})
}

func exists(root fs.FS, name string) bool {
	file, err := root.Open(name)
	if err != nil {
		return false
	}
	_ = file.Close()
	return true
}

// hashed reports whether the build named this file after its contents, which is
// what makes it safe to keep forever. Angular's outputHashing writes
// main-AMM3UY6G.js and styles-QAEZWQS6.css; index.html never carries one.
func hashed(name string) bool {
	base := name[strings.LastIndex(name, "/")+1:]
	dot := strings.LastIndex(base, ".")
	if dot < 0 {
		return false
	}
	dash := strings.LastIndex(base[:dot], "-")
	if dash < 0 {
		return false
	}
	stamp := base[dash+1 : dot]
	if len(stamp) < 8 {
		return false
	}
	for _, r := range stamp {
		if (r < 'A' || r > 'Z') && (r < '0' || r > '9') {
			return false
		}
	}
	return true
}
