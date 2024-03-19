package v1

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/hpetrov29/restapi/internal/logger"
	"github.com/hpetrov29/restapi/internal/web"
)

// APIMuxConfig contains all mandatory systems required by handlers.
type APIMuxConfig struct {
	Build    string
	Shutdown chan os.Signal
	Log      *logger.Logger
	DB       *sql.DB
}

func NewAPIMux(config APIMuxConfig) http.Handler {
	app := web.NewApp(config.Shutdown, nil)
	return app
}