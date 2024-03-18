package logger

import (
	"context"
	"log/slog"
)

type logHandler struct {
	handler slog.Handler
	events Events
}

func newLogHandler(handler slog.Handler, events Events) *logHandler {
	return &logHandler{handler, events}
}

// Enabled reports whether the handler handles records at the given level. 
// The handler ignores records whose level is lower. It is called early, 
// before any arguments are processed, to save effort if the log event should be discarded. 
// If called from a Logger method, the first argument is the context passed to that method, 
// or context.Background() if nil was passed or the method does not take a context. 
// The context is passed so Enabled can use its values to make a decision.
func (h *logHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
} 

// Handle handles the Record. It will only be called when Enabled returns true. 
// The Context argument is as for Enabled. 
// It is present solely to provide Handlers access to the context's values. 
// Canceling the context should not affect record processing. (Among other things, 
// log messages may be necessary to debug a cancellation-related problem.)
func (h *logHandler) Handle(ctx context.Context, r slog.Record) error {
	switch r.Level {
	case slog.LevelDebug:
		if h.events.Debug != nil {
			h.events.Debug(ctx, toRecord(r))
		}

	case slog.LevelError:
		if h.events.Error != nil {
			h.events.Error(ctx, toRecord(r))
		}

	case slog.LevelWarn:
		if h.events.Warn != nil {
			h.events.Warn(ctx, toRecord(r))
		}

	case slog.LevelInfo:
		if h.events.Info != nil {
			h.events.Info(ctx, toRecord(r))
		}
	}

	return h.handler.Handle(ctx, r)
}

// WithAttrs returns a new JSONHandler whose attributes consists
// of h's (the old) attributes followed by attrs (the new attributes).
func (h *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logHandler{handler: h.handler.WithAttrs(attrs), events: h.events}
}

// WithGroup returns a new Handler with the given group appended to the receiver's
// existing groups. The keys of all subsequent attributes, whether added by With
// or in a Record, should be qualified by the sequence of group names.
func (h *logHandler) WithGroup(name string) slog.Handler {
	return &logHandler{handler: h.handler.WithGroup(name), events: h.events}
}