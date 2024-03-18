package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"time"
)

// TraceIDFunc is a declaration for a function that returns
// the trace id from the passed context
type TraceIDFunc func(ctx context.Context) string

// Logger is responsible for logging the messages
type Logger struct {
	handler slog.Handler
	traceIDFunc TraceIDFunc
}

// NewWithEvents constructs a Logger for application use with events.
func NewWithEvents(w io.Writer, minLevel Level, serviceName string, traceIDFunc TraceIDFunc, events Events) *Logger {
	return new(w, minLevel, serviceName, traceIDFunc, events)
}

// Info logs at LevelInfo with the given context.
func (log *Logger) Info(ctx context.Context, message string, args ...any) {
	log.write(ctx, LevelInfo, 3, message, args...)
}

func (log *Logger) write(ctx context.Context, level Level, caller int, message string, args ...any) {
	slogLevel := slog.Level(level)

	if !log.handler.Enabled(ctx, slogLevel) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:])

	r := slog.NewRecord(time.Now(), slogLevel, message, pcs[0])

	if log.traceIDFunc != nil {
		args = append(args, "trace_id", log.traceIDFunc(ctx))
	}
	r.Add(args...)

	log.handler.Handle(ctx, r)
}

// =============================================================================

func new(w io.Writer, minLevel Level, serviceName string, traceIDFunc TraceIDFunc, events Events) *Logger {

	// Function that formats the source of the event to file.ext:line
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	// Construct the slog JSON handler
	handler := slog.Handler(slog.NewJSONHandler(w, &slog.HandlerOptions{AddSource: true, Level: slog.Level(minLevel), ReplaceAttr: f}))

	if events.Debug != nil || events.Error != nil || events.Info != nil || events.Warn != nil {
		handler = newLogHandler(handler, events)
	}

	// Attributes to add to every log message.
	attributes := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}

	handler = handler.WithAttrs(attributes)

	return &Logger{
		handler: handler,
		traceIDFunc: traceIDFunc,
	}
}

