package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	bot    *Bot
	router Handler
}

func NewServer(bot *Bot, router Handler) *Server {
	return &Server{
		bot:    bot,
		router: router,
	}
}

type Request struct {
	Chat *tgbotapi.Chat
	Data string
}

func (s *Server) Listen(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		r := new(Request)
		if update.Message != nil {
			r.Chat = update.Message.Chat
			r.Data = update.Message.Text
		} else if update.CallbackQuery != nil {
			r.Chat = update.CallbackQuery.Message.Chat
			r.Data = update.CallbackQuery.Data
		}
		go s.router.ServeBot(ctx, r)
	}
}
