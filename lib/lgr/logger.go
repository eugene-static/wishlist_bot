package lgr

import (
	"fmt"
	"io"
	"log/slog"
)

type Log struct {
	*slog.Logger
	errors map[int]string
}

func New(w io.Writer, env string) *Log {
	var handler slog.Handler
	if env == "debug" {
		handler = slog.NewTextHandler(w, &slog.HandlerOptions{
			AddSource:   false,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		})

	} else {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			AddSource:   false,
			Level:       slog.LevelInfo,
			ReplaceAttr: nil,
		})
	}
	return &Log{
		Logger: slog.New(handler),
		errors: make(map[int]string),
	}
}

func (l *Log) Errorf(text string, err error, attr ...any) {
	l.With(attr...).Error(text, slog.Any("error", err))
}

func (l *Log) Set(key int, text string) {
	l.errors[key] = text
}

func (l *Log) get(key int) string {
	if err, ok := l.errors[key]; ok {
		return err
	}
	return "Unknown error"
}

func (l *Log) ErrorCode(code int, err error) error {
	return fmt.Errorf(l.get(code), err)
}

func (l *Log) With(attr ...any) *Log {
	return &Log{
		Logger: l.Logger.With(attr...),
		errors: l.errors,
	}
}
