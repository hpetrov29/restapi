package v1

import (
	"net/http"
	"os"

	"github.com/hpetrov29/restapi/business/web/v1/auth"
	"github.com/hpetrov29/restapi/internal/logger"
	"github.com/hpetrov29/restapi/internal/web"
	"github.com/jmoiron/sqlx"
)

// APIMuxConfig contains all mandatory systems required by handlers.
type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
	Auth	*auth.Auth
	DB       *sqlx.DB
}

// RouteAdder defines behavior that sets the routes to bind for an instance
// of the service.
type RouteAdder interface {
	Add(app *web.App, cfg APIMuxConfig)
}

func NewAPIMux(config APIMuxConfig, routeAdder RouteAdder) http.Handler {
	app := web.NewApp(config.Shutdown, nil)

	routeAdder.Add(app, config)

	return app
}