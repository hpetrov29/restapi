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

	db "github.com/hpetrov29/restapi/business/data/dbsql/mysql"
	v1 "github.com/hpetrov29/restapi/business/web/v1"
	"github.com/hpetrov29/restapi/business/web/v1/auth"
	"github.com/hpetrov29/restapi/internal/keystore"
	"github.com/hpetrov29/restapi/internal/logger"
	"github.com/hpetrov29/restapi/internal/web"
)

func Main(routeAdder v1.RouteAdder) {
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

	run(ctx, log, "v1", routeAdder)
}

func run(ctx context.Context, log *logger.Logger, build string, routeAdder v1.RouteAdder) error {

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
		DB struct {
			User         string `conf:"default:root"`
			Password     string `conf:"default:"`
			Host         string `conf:"default:localhost:3306"`
			Name         string `conf:"default:golang_api"`
			MaxIdleConns int    `conf:"default:2"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:true"`
		}
		Auth struct {
			KeysFolder string
			Issuer string
		}
	}{}

	config.Version.Build = build
	config.Version.Description = ""

	config.Web.APIHost = "localhost:3000"
	config.Web.ReadTimeout = time.Duration(5)*time.Second
	config.Web.WriteTimeout = time.Duration(10)*time.Second
	config.Web.IdleTimeout = time.Duration(120)*time.Second
	config.Web.ShutdownTimeout = time.Duration(20)*time.Second

	config.DB.User = os.Getenv("DB_USER")
	config.DB.Password = os.Getenv("DB_PASSWORD")
	config.DB.Host = os.Getenv("DB_HOST")
	config.DB.Name = os.Getenv("DB_NAME")
	config.DB.MaxIdleConns = 2
	config.DB.MaxOpenConns = 0
	config.DB.DisableTLS = true

	config.Auth.KeysFolder = "zarf/keys"
	config.Auth.Issuer = "service"

	// -------------------------------------------------------------------------
	// Set up database client conneciton

	log.Info(ctx, "DB startup", "status", "initializing database support", "host", config.DB.Host)

	dbClient, err := db.Open(db.Config{
		User:         config.DB.User,
		Password:     config.DB.Password,
		Host:         config.DB.Host,
		Name:         config.DB.Name,
		MaxIdleConns: config.DB.MaxIdleConns,
		MaxOpenConns: config.DB.MaxOpenConns,
		DisableTLS:   config.DB.DisableTLS,
	})
	if err != nil {
		fmt.Println("error connecting to db: ", err)
	}
	defer func() {
		log.Info(ctx, "DB shutdown", "status", "stopping database support", "host", config.DB.Host)
		dbClient.Close()
	}()

	err = db.StatusCheck(ctx, dbClient); if err != nil {
		fmt.Println("error database status check: ", err)
	}

	// -------------------------------------------------------------------------
	// Initialize authentication support

	log.Info(ctx, "Auth startup", "status", "initializing authentication support")

	keystore, err := keystore.NewFS(os.DirFS(config.Auth.KeysFolder))
	if err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	auth, err := auth.New(auth.Config{
		Log:       log,
		DB:        dbClient,
		Issuer:    config.Auth.Issuer,
		Vault: 	   keystore,
	})
	if err != nil {
		return fmt.Errorf("constructing auth: %w", err)
	}

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
		Auth: auth,
		DB: dbClient,
	}

	apiMux := v1.NewAPIMux(muxConfig, routeAdder)

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
			log.Info(ctx, "API shutdown", "status", "shutdown started", "signal", sig)
			defer log.Info(ctx, "API shutdown", "status", "shutdown complete", "signal", sig)
	
			ctx, cancel := context.WithTimeout(ctx, config.Web.ShutdownTimeout)
			defer cancel()
	
			if err := api.Shutdown(ctx); err != nil {
				api.Close()
				return fmt.Errorf("could not stop server gracefully: %w", err)
			}
	}

	return nil
}