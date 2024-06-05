package lgr

import (
	"fmt"
	"io"
	"log/slog"
)

const (
	ErrGetUser = iota
	ErrAddUser
	ErrGetList
	ErrAddWish
	ErrDelWish
	ErrChangePass
	ErrUpdateUsername
	ErrSendMessage
	ErrStorageInit
	ErrStorageClose
	ErrCreateBot
	ErrGetMessageConfig
	ErrShutdown
)

type Log struct {
	*slog.Logger
}

func New(w io.Writer, env string) *Log {
	if env == "debug" {
		handler := slog.NewTextHandler(w, &slog.HandlerOptions{
			AddSource:   false,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		})
		return &Log{slog.New(handler)}
	}
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelInfo,
		ReplaceAttr: nil,
	})
	return &Log{slog.New(handler)}
}

func (l *Log) Errorf(text string, err error, attr ...any) {
	l.With(attr...).Error(text, slog.Any("error", err))
}

func match(code int) string {
	var text string
	switch code {
	case ErrGetList:
		text = "Get list error"
	case ErrAddWish:
		text = "Add wish error"
	case ErrDelWish:
		text = "Delete wish error"
	case ErrChangePass:
		text = "Change password error"
	case ErrUpdateUsername:
		text = "Update username error"
	case ErrAddUser:
		text = "Add user error"
	case ErrGetUser:
		text = "GetUser error"
	case ErrSendMessage:
		text = "Send message error"
	case ErrStorageInit:
		text = "Storage init error"
	case ErrStorageClose:
		text = "Storage close error"
	case ErrCreateBot:
		text = "Creating bot error"
	case ErrGetMessageConfig:
		text = "Get message config error"
	case ErrShutdown:
		text = "Shutdown error"
	default:
		text = "Unknown error"
	}
	return text
}

func (l *Log) Sprint(code int, err error) error {
	return fmt.Errorf(match(code), err)
}

func (l *Log) With(attr ...any) *Log {
	return &Log{l.Logger.With(attr...)}
}
