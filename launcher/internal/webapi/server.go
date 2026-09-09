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
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self'; connect-src 'self'; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}
