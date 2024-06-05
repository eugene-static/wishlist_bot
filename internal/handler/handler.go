package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/eugene-static/wishlist_bot/internal/bot"
	"github.com/eugene-static/wishlist_bot/internal/entity"
	"github.com/eugene-static/wishlist_bot/internal/session"
	"github.com/eugene-static/wishlist_bot/lib/lgr"
	"golang.org/x/crypto/bcrypt"
)

type User interface {
	GetUser(ctx context.Context, userID int64) (*entity.User, error)
	GetUserByUsername(ctx context.Context, username string) (*entity.User, error)
	AddUser(ctx context.Context, user *entity.User) error
	UpdateUser(ctx context.Context, id int64, username string, new []byte) error
}

type List interface {
	AddWish(ctx context.Context, wish *entity.Wish) error
	GetWishlistByID(ctx context.Context, id int64) ([]*entity.Wish, error)
	DeleteWishes(ctx context.Context, ids []string) error
}

type Service interface {
	User
	List
}

type Handle struct {
	log     *lgr.Log
	service Service
	mgr     *session.Manager
	bot     *bot.Bot
	mux     *bot.Mux
}

func New(log *lgr.Log, service Service, mgr *session.Manager, b *bot.Bot, mux *bot.Mux) *Handle {
	return &Handle{
		log:     log,
		service: service,
		mgr:     mgr,
		bot:     b,
		mux:     mux,
	}
}

func (h *Handle) Register() {
	h.mux.Handle(bot.DefaultMessage, h.message)
	h.mux.Handle(messageAdd, h.add(h.show(true)))
	h.mux.Handle(messageDelete, h.delete(h.show(true)))
	h.mux.Handle(messageShowUser, h.show(false))
	h.mux.Handle(messagePassword, h.password)
	h.mux.Handle(messageStart, h.start)
	h.mux.Handle(actionShowMe, h.show(true))
	h.mux.Handle(actionBack, h.callback(textGreetings, lvlStart, messageStart))
	h.mux.Handle(actionAdd, h.callback(textAddWish, lvlEdit, messageAdd))
	h.mux.Handle(actionDelete, h.callback(textDeleteWish, lvlEdit, messageDelete))
	h.mux.Handle(actionPassword, h.callback(textEnterPassword, lvlEdit, messagePassword))
	h.mux.Handle(actionShowUser, h.callback(textEnterUsername, lvlUser, messageShowUser))
}

func (h *Handle) send(user *session.User, configKey int, messageKey int) {
	_, err := h.bot.Send(user.ID, configKey, messageKey)
	if err != nil {
		h.error(user, err)
	}

}

func (h *Handle) errorCode(code int, user *session.User, err error) {
	err = h.log.Sprint(code, err)
	log := h.log.With(
		slog.Any("error", err),
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Name),
		slog.String("request", user.Request),
	)
	log.Error("error building message")
	h.bot.Config.SetReplyMessage(
		textError,
		fmt.Sprintf("В работе бота возникла ошибка. Код %03o\nПопробуйте снова позже или же обратитесь к %s за помощью", code, admin))
	if _, err = h.bot.Send(user.ID, lvlEmpty, textError); err != nil {
		log.Errorf("error sending message", err)
	}
}

func (h *Handle) error(user *session.User, err error) {
	if user == nil {
		h.log.Errorf("error sending message", err)
	} else {
		h.log.Error("error building message",
			slog.Any("error", err),
			slog.Int64("user_id", user.ID),
			slog.String("username", user.Name),
			slog.String("request", user.Request),
		)
	}
}

func (h *Handle) getUser(ctx context.Context, r *bot.Request) (*session.User, error) {
	user := h.mgr.GetUser(r.Chat.ID)
	if user == nil {
		log := h.log.With(
			slog.Int64("user_id", r.Chat.ID),
			slog.String("username", r.Chat.UserName),
		)
		log.Debug("not such user, searching in db")
		userData, err := h.service.GetUser(ctx, r.Chat.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Debug("not such user, adding in db")
				hashedPass, err := hash(r.Chat.UserName)
				if err != nil {
					return nil, fmt.Errorf("error generating password hash: %w", err)
				}
				userData = &entity.User{
					ID:       r.Chat.ID,
					Name:     r.Chat.UserName,
					Password: hashedPass,
				}
				err = h.service.AddUser(ctx, userData)
				if err != nil {
					return nil, fmt.Errorf("error adding user to db: %w", err)
				}
			} else {
				return nil, fmt.Errorf("error getting user from db: %w", err)
			}
		}
		if userData.Name != r.Chat.UserName {
			if err = h.service.UpdateUser(ctx, userData.ID, r.Chat.UserName, nil); err != nil {
				return nil, fmt.Errorf("error updating user in db: %w", err)
			}
		}
		log.Debug("adding user in session manager")
		user = h.mgr.AddUser(r.Chat.ID, r.Chat.UserName)
	}
	return user, nil
}

func hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func compare(password1, password2 []byte) error {
	return bcrypt.CompareHashAndPassword(password1, password2)
}
