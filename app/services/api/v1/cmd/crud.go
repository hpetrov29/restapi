package cmd

import (
	"github.com/hpetrov29/restapi/app/services/api/v1/handlers/users"
	v1 "github.com/hpetrov29/restapi/business/web/v1"
	"github.com/hpetrov29/restapi/internal/web"
)

// Routes constructs the add value which provides the implementation of
// of RouteAdder for specifying what routes to bind to this instance.
func Routes() add {
	return add{}
}

type add struct{}

// Add implements the RouterAdder interface.
func (add) Add(app *web.App, cfg v1.APIMuxConfig) {
	users.Routes(app, users.Config{
		Log:  cfg.Log,
		Auth: cfg.Auth,
		DB:   cfg.DB,
	})
}
