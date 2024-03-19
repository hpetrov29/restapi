package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/go-chi/chi"
)

// Handler is a type definition that handles a http request within the mini framework.
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into the application and what configures our context
// object for each of our http handlers. Feel free to add any configuration
// data/logic on this App struct.
type App struct {
	mux http.Handler
	shutdown chan os.Signal
	middlewares []Middleware
}

// NewApp creates an App instance using the chi router.
func NewApp(shutdown chan os.Signal, middlewares ...Middleware) *App {

	// TO DO: Create an OpenTelemetry HTTP Handler which wraps our router.
	mux := chi.NewMux();

	return &App{
		mux: mux,
		shutdown: shutdown,
		middlewares: middlewares,
	}
}

// SignalShutdown is used to gracefully shut down the app when an integrity
// issue is identified.
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// ServeHTTP method implements the http.Handler interface for App. 
// It's the entry point for all http traffic.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}