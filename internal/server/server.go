package server

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eugene-static/wishlist_bot/internal/bot"
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
		log: lgr.New(output, cfg.Logger.Env),
	}
}

func (s *Server) Start() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	appStorage, err := storage.New(ctx, &s.cfg.Storage)
	if err != nil {
		s.log.Errorf("storage initialization error", err)
		return
	}
	botapi, err := tgbotapi.NewBotAPI(s.cfg.Bot.Token)
	if err != nil {
		s.log.Errorf("bot creating error", err)
		return
	}
	botapi.Debug = s.cfg.Bot.DebugMode
	mux := bot.NewBotMux()
	b := bot.NewBot(botapi)
	appHandler := handler.New(s.log, service.New(appStorage), session.New(), b, mux)
	appHandler.Register()
	appHandler.Build()
	s.log.Info("authorized", slog.String("admin", botapi.Self.UserName))
	go bot.NewServer(b, mux).Listen(ctx, botapi.GetUpdatesChan(tgbotapi.UpdateConfig{
		Offset:  s.cfg.Bot.UpdateOffset,
		Limit:   s.cfg.Bot.UpdateLimit,
		Timeout: s.cfg.Bot.UpdateTimeout,
	}))
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	long := make(chan struct{}, 1)
	go func() {
		botapi.StopReceivingUpdates()
		if err = appStorage.Close(); err != nil {
			s.log.Errorf("closing storage error", err)
		}
		long <- struct{}{}
	}()
	select {
	case <-ctx.Done():
		s.log.Errorf("error during shutdown", ctx.Err())
	case <-long:
		s.log.Info("The app is shut down successfully")
	}
}
