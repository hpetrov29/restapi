package users

import (
	"net/http"

	"github.com/hpetrov29/restapi/business/web/v1/auth"
	"github.com/hpetrov29/restapi/business/web/v1/middleware"
	"github.com/hpetrov29/restapi/internal/logger"
	"github.com/hpetrov29/restapi/internal/web"
	"github.com/jmoiron/sqlx"
)

// Config contains all the mandatory systems required by handlers.
type Config struct {
	Log  *logger.Logger
	Auth *auth.Auth
	DB   *sqlx.DB
}

// Routes adds specific routes for this group.
func Routes(app *web.App, cfg Config) {

	const version = "v1"

	handlers := New(cfg.Auth)

	authenticated := middleware.Authenticate(cfg.Auth)
	_ = middleware.Authorize(cfg.Auth, auth.RuleAdminOnly)
	_ = middleware.Authorize(cfg.Auth, auth.RuleAdminOrSubject)

	// arguments: METHOD, version, path, controller, ...middlewares
	app.Handle(http.MethodGet, version, "/users/token/{kid}", handlers.Token)
	app.Handle(http.MethodGet, version, "/users", handlers.Query, authenticated)
}
