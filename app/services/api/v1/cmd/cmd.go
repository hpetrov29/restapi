package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"

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

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "API", traceIDFunc, events)

	ctx := context.Background()

	run(ctx, log, "v1")
}

func run(ctx context.Context, log *logger.Logger, build string) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS
	
	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration


	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", build)
	defer log.Info(ctx, "shutdown complete")

	api := &http.Server{
        Addr:    ":8080", // Listen on port 8080
        Handler: nil,     // No handler initially
    }

	go func() {
		if err := api.ListenAndServe(); err != nil {
			fmt.Printf("Error starting server: %s", err)
		}
	}()

	//shutdown := make(chan os.Signal, 1)

	select{}
	//return nil
}