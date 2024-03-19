package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	v1 "github.com/hpetrov29/restapi/business/web/v1"
	"github.com/hpetrov29/restapi/internal/logger"
	"github.com/hpetrov29/restapi/internal/web"
)

func Main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT ******")
		},
	}

	traceIDFunc := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	// Logger will disregard logs of category lower than the one specified here
	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "API", traceIDFunc, events)

	ctx := context.Background()

	run(ctx, log, "v1")
}

func run(ctx context.Context, log *logger.Logger, build string) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS
	
	log.Info(ctx, "service startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration

	config := struct {
		Version struct {
			Build string
			Description string
		}
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			// DebugHost       string        `conf:"default:0.0.0.0:4000"`
		}
	}{}

	config.Version.Build = build
	config.Version.Description = ""

	config.Web.APIHost = "localhost:3000"
	config.Web.ReadTimeout = time.Duration(5)*time.Second
	config.Web.WriteTimeout = time.Duration(10)*time.Second
	config.Web.IdleTimeout = time.Duration(120)*time.Second
	config.Web.ShutdownTimeout = time.Duration(20)*time.Second

	// -------------------------------------------------------------------------
	// Start API

	log.Info(ctx, "API startup", "version", build)

	// Only the signals explicitly provided (SIGINT and SIGTERM) will be captured.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	muxConfig := v1.APIMuxConfig{
		Build: build,
		Shutdown: shutdown,
		Log: log,
		DB: nil, //TO DO: Add db connection
	}

	apiMux := v1.NewAPIMux(muxConfig)

	api := &http.Server{
        Addr:    config.Web.APIHost,
        Handler: apiMux,
		ReadTimeout:  config.Web.ReadTimeout,
		WriteTimeout: config.Web.WriteTimeout,
		IdleTimeout:  config.Web.IdleTimeout,
    }

	serverErrors := make(chan error, 1)

	go func() {
		serverErrors <- api.ListenAndServe()
	}()

	select{
		case err := <-serverErrors:
			return fmt.Errorf("server error: %w", err)
		case sig := <-shutdown:
			log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
			defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)
	
			ctx, cancel := context.WithTimeout(ctx, config.Web.ShutdownTimeout)
			defer cancel()
	
			if err := api.Shutdown(ctx); err != nil {
				api.Close()
				return fmt.Errorf("could not stop server gracefully: %w", err)
			}
	}

	return nil
}