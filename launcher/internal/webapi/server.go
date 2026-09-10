package webapi

import (
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/validate"

	"github.com/m-this/tf2-archipelago/launcher/internal/gen/tf2ap/launcher/v1/launcherv1connect"
	"github.com/m-this/tf2-archipelago/launcher/internal/spa"
)

/*
Handler is the whole interface: the three Connect services, a health check and
the embedded app under everything else.

Loopback only, no auth and no CSRF token: the four passwords never cross a
network, and this is one player's own machine. What the mux does check is that
the request is for the address it bound. Comparing Origin with Host is not
enough on its own, because a DNS-rebinding attacker controls both, so Host is
pinned to the authority the listener actually took. An absent Origin stays
allowed: a local tool with no browser behind it is a reasonable thing to be.
*/
func (a *App) Handler(authority string) http.Handler {
	options := connect.WithInterceptors(validate.NewInterceptor())

	mux := http.NewServeMux()
	mux.Handle(launcherv1connect.NewLauncherServiceHandler(LauncherRPC{App: a}, options))
	mux.Handle(launcherv1connect.NewSettingsServiceHandler(SettingsRPC{App: a}, options))
	mux.Handle(launcherv1connect.NewFilesServiceHandler(NewFilesRPC(a), options))
	mux.Handle(grpchealth.NewHandler(grpchealth.NewStaticChecker(
		launcherv1connect.LauncherServiceName,
		launcherv1connect.SettingsServiceName,
		launcherv1connect.FilesServiceName,
	)))
	mux.HandleFunc("GET /ws", a.serveStream(authority))
	mux.Handle("/", spa.Handler())

	return guard(authority, mux)
}

/*
policy is what the page is allowed to do.

Everything comes from this origin and nothing is fetched from anywhere: the
fonts are bundled, there is no analytics and no CDN. script-src has no
unsafe-inline, so nothing injected into a log line or a mission name could run.

style-src does, and it has to. Angular writes an element's style attribute for
a style binding, and the virtual scroll the log view is built on moves its
content by writing a transform there on every frame. Without unsafe-inline the
browser drops those writes and the log stops scrolling, which is what running
the real binary is for.
*/
const policy = "default-src 'self'; " +
	"script-src 'self'; " +
	"style-src 'self' 'unsafe-inline'; " +
	"connect-src 'self'; " +
	"img-src 'self' data:; " +
	"font-src 'self'; " +
	"base-uri 'self'; " +
	"form-action 'none'; " +
	"frame-ancestors 'none'"

func guard(authority string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != authority {
			http.Error(w, "wrong host", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+authority {
			http.Error(w, "wrong origin", http.StatusForbidden)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", policy)
		next.ServeHTTP(w, r)
	})
}
