package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eugene-static/wishlist_bot/internal/handler"
	"github.com/eugene-static/wishlist_bot/internal/service"
	"github.com/eugene-static/wishlist_bot/internal/session"
	"github.com/eugene-static/wishlist_bot/internal/storage"
	"github.com/eugene-static/wishlist_bot/lib/config"
	"github.com/eugene-static/wishlist_bot/lib/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Server struct {
	cfg *config.Config
	log *lgr.Log
}

func New(cfg *config.Config) *Server {
	var output *os.File
	if cfg.Logger.Internal {
		output = os.Stdout
	} else {
		var err error
		output, err = os.Open(cfg.Logger.ExternalPath)
		if err != nil {
			panic(err)
		}
		defer output.Close()
	}
	return &Server{
		cfg: cfg,
		log: lgr.New(output, cfg.Logger.Level),
	}
}

func (s *Server) Start() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	appStorage, err := storage.New(ctx, &s.cfg.Storage)
	if err != nil {
		s.log.Errorf(lgr.ErrStorageInit, err)
		return
	}
	appHandler := handler.New(s.log, session.New(), service.New(appStorage))
	bot, err := tgbotapi.NewBotAPI(s.cfg.Bot.Token)
	if err != nil {
		s.log.Errorf(lgr.ErrCreateBot, err)
		return
	}
	s.log.Info("authorized", slog.String("admin", bot.Self.UserName))
	bot.Debug = s.cfg.Bot.DebugMode
	u := tgbotapi.NewUpdate(s.cfg.Bot.UpdateOffset)
	u.Timeout = s.cfg.Bot.UpdateTimeout
	updates := bot.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			mc := appHandler.MessageConfig(ctx, &update)
			if mc == nil {
				continue
			}
			if _, err = bot.Send(mc); err != nil {
				s.log.Errorf(lgr.ErrSendMessage, err)
			}
		}
	}()
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	long := make(chan struct{}, 1)
	go func() {
		bot.StopReceivingUpdates()
		if err = appStorage.Close(); err != nil {
			s.log.Errorf(lgr.ErrStorageClose, err)
		}
		long <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		s.log.Errorf(lgr.ErrShutdown, ctx.Err())
	case <-long:
		s.log.Info("The app is shut down successfully")
	}
}
