package ekbot

import (
	"context"

	"github.com/eugene-static/wishlist_bot/internal/session"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler interface {
	Serve(ctx context.Context, user *session.User) tgbotapi.MessageConfig
}

type HandlerFunc func(ctx context.Context, user *session.User) tgbotapi.MessageConfig

type List struct {
	m map[string]HandlerFunc
}

func NewList() *List {
	return &List{make(map[string]HandlerFunc)}
}

func (l *List) Handle(name string, handler HandlerFunc) {
	l.m[name] = handler
}

func (f HandlerFunc) Serve(ctx context.Context, user *session.User) tgbotapi.MessageConfig {
	return f(ctx, user)
}
