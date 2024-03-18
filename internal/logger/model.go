package logger

import (
	"context"
	"log/slog"
	"time"
)

// level represents different different log categories
type Level slog.Level

// logging categories in use
const (
	LevelDebug = Level(slog.LevelDebug)
	LevelInfo = Level(slog.LevelInfo)
	LevelWarn = Level(slog.LevelWarn)
	LevelError = Level(slog.LevelError)
)

// the data to be logged
type Record struct {
	Time time.Time
	Message string
	Level Level
	Attributes map[string]any
}

// transforms slog.Record to Record
func toRecord(r slog.Record) Record {
	attributes := make(map[string]any, r.NumAttrs())

	f := func(attr slog.Attr) bool {
		attributes[attr.Key] = attr.Value.Any()
		return true
	}

	r.Attrs(f)

	return Record{
		Time: r.Time,
		Message: r.Message,
		Level: Level(r.Level),
		Attributes: attributes,
	}
}

// EventFunc is a function to be executed when configured against a log level.
type EventFunc func(ctx context.Context, r Record)

// Events contains an assignment of an event function to a log level.
type Events struct{
	Debug EventFunc
	Info EventFunc
	Warn EventFunc
	Error EventFunc
}