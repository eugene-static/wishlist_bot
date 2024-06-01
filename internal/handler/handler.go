package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/eugene-static/wishlist_bot/internal/entity"
	"github.com/eugene-static/wishlist_bot/internal/session"
	"github.com/eugene-static/wishlist_bot/lib/ekbot"
	"github.com/eugene-static/wishlist_bot/lib/lgr"
	"github.com/eugene-static/wishlist_bot/lib/random"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	actionStart     = "/start"
	actionAdd       = "/add"
	actionDelete    = "/delete"
	actionPassword  = "/password"
	actionShowUser  = "/show_user"
	actionShowMe    = "/show_me"
	actionBack      = "/back"
	messageAdd      = "message_add"
	messageDelete   = "message_delete"
	messageShowUser = "message_show_user"
	messagePassword = "message_password"
	DeletePassword  = "Удалить пароль"
)

const (
	lvlEmpty = iota
	lvlStart
	lvlMe
	lvlUser
	lvlEdit
	lvlService
	lvlServiceExt
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
	mgr     *session.Manager
	service Service
	list    *ekbot.List
}

func New(log *lgr.Log, mgr *session.Manager, service Service, list *ekbot.List) *Handle {
	return &Handle{
		log:     log,
		mgr:     mgr,
		service: service,
		list:    list,
	}
}

func (h *Handle) MessageConfig(ctx context.Context, u *tgbotapi.Update) *tgbotapi.MessageConfig {
	chat := u.FromChat()
	log := h.log.With(
		slog.Int64("user_id", chat.ID),
		slog.String("username", chat.UserName),
	)
	user := h.mgr.GetUser(chat.ID)
	if user == nil {
		log.Debug("not such user, searching in db")
		userData, err := h.service.GetUser(ctx, chat.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Debug("not such user, adding in db")
				hashedPass, err := bcrypt.GenerateFromPassword([]byte(chat.UserName), bcrypt.DefaultCost)
				if err != nil {
					log.Errorf(lgr.ErrChangePass, err)
					return nil
				}
				userData = &entity.User{
					ID:       chat.ID,
					Name:     chat.UserName,
					Password: hashedPass,
				}
				err = h.service.AddUser(ctx, userData)
				if err != nil {
					log.Errorf(lgr.ErrAddUser, err)
					return nil
				}
			} else {
				log.Errorf(lgr.ErrGetUser, err)
				return nil
			}
		}
		if userData.Name != chat.UserName {
			if err = h.service.UpdateUser(ctx, userData.ID, chat.UserName, nil); err != nil {
				log.Errorf(lgr.ErrUpdateUsername, err)
				return nil
			}
		}
		log.Debug("adding user in session manager")
		user = h.mgr.AddUser(chat.ID, chat.UserName)
	}
	var mc tgbotapi.MessageConfig
	if u.Message != nil {
		user.Request = u.Message.Text
		h.log = h.log.With(slog.String("request", user.Request))
		mc = h.onMessage(ctx, user)
	} else if u.CallbackQuery != nil {
		user.Request = u.CallbackQuery.Data
		mc = h.onCallback(ctx, user)
	} else {
		log.Debug("unknown update")
		return nil
	}
	return &mc
}

func (h *Handle) Start(_ context.Context, user *session.User) tgbotapi.MessageConfig {
	mc := newConfig(user.ID)
	mc.text = text(textGreetings)
	mc.level = lvlStart
	return mc.build()
}

func (h *Handle) Add(next ekbot.HandlerFunc) ekbot.HandlerFunc {
	return func(ctx context.Context, user *session.User) tgbotapi.MessageConfig {
		mc := newConfig(user.ID)
		if err := h.service.AddWish(ctx, &entity.Wish{
			ID:      random.String(16),
			Content: user.Request,
			UserID:  user.ID,
		}); err != nil {
			h.Error(lgr.ErrAddWish, user, err)
			return mc.error(lgr.ErrAddWish)
		}
		return next(ctx, user)
	}
}

func (h *Handle) Delete(next ekbot.HandlerFunc) ekbot.HandlerFunc {
	return func(ctx context.Context, user *session.User) tgbotapi.MessageConfig {
		mc := newConfig(user.ID)
		nums := strings.Fields(user.Request)
		if len(nums) > 0 {
			for i, num := range nums {
				index, err := strconv.Atoi(num)
				if err != nil || index > len(user.IDList) || index <= 0 {
					mc.text = text(textWrongRequest)
					return mc.build()
				}
				nums[i] = user.IDList[index-1]
			}
			if err := h.service.DeleteWishes(ctx, nums); err != nil {
				h.Error(lgr.ErrDelWish, user, err)
				return mc.error(lgr.ErrDelWish)
			}
		}
		return next(ctx, user)
	}
}

func (h *Handle) Password(ctx context.Context, user *session.User) tgbotapi.MessageConfig {
	mc := newConfig(user.ID)
	mc.level = lvlService
	if user.Request == DeletePassword {
		user.Request = user.Name
	} else if strings.ContainsRune(user.Request, ' ') {
		mc.level = lvlEmpty
		mc.text = text(textNoSpace)
		return mc.build()
	}
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.Request), bcrypt.DefaultCost)
	if err != nil {
		h.Error(lgr.ErrChangePass, user, err)
		return mc.error(lgr.ErrChangePass)
	}
	if err = h.service.UpdateUser(ctx, user.ID, user.Name, hashedPass); err != nil {
		h.Error(lgr.ErrChangePass, user, err)
		return mc.error(lgr.ErrChangePass)
	}
	mc.text = text(textSuccess)
	return mc.build()
}

func (h *Handle) Show(me bool) ekbot.HandlerFunc {
	return func(ctx context.Context, user *session.User) tgbotapi.MessageConfig {
		mc := newConfig(user.ID)
		var id int64
		if !me {
			mc.level = lvlUser
			req := strings.Fields(user.Request)
			switch len(req) {
			case 1:
				req = append(req, user.Name)
			case 2:
				break
			default:
				mc.text = text(textWrongRequest)
				return mc.build()
			}
			username, password := strings.TrimPrefix(req[0], "@"), []byte(req[1])
			reqUser, err := h.service.GetUserByUsername(ctx, username)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					mc.text = text(textUserNotFound)
					return mc.build()
				}
				h.Error(lgr.ErrGetUser, user, err)
				return mc.error(lgr.ErrGetUser)
			}
			if bcrypt.CompareHashAndPassword(reqUser.Password, password) != nil {
				mc.text = text(textWrongPassword)
				return mc.build()
			}
			id = reqUser.ID
		} else {
			mc.level = lvlMe
			id = user.ID
		}
		list, err := h.service.GetWishlistByID(ctx, id)
		if err != nil {
			h.Error(lgr.ErrGetList, user, err)
			return mc.error(lgr.ErrGetList)
		}
		if list == nil {
			mc.text = text(textNoWishes)
			return mc.build()
		}
		var wishes strings.Builder
		for i, wish := range list {
			if me {
				user.IDList = append(user.IDList, fmt.Sprintf("'%s'", wish.ID))
			}
			_, _ = wishes.WriteString(fmt.Sprintf("%d. %s\n", i+1, wish.Content))
		}
		mc.text = wishes.String()
		return mc.build()
	}
}

func (h *Handle) Service(code int, level int, action string) ekbot.HandlerFunc {
	return func(ctx context.Context, user *session.User) tgbotapi.MessageConfig {
		mc := newConfig(user.ID)
		mc.text = text(code)
		mc.level = level
		user.Action = action
		return mc.build()
	}
}

func (h *Handle) Register() {
	h.list.Handle(messageAdd, h.Add(h.Show(true)))
	h.list.Handle(messageDelete, h.Delete(h.Show(true)))
	h.list.Handle(messageShowUser, h.Show(false))
	h.list.Handle(messagePassword, h.Password)
	h.list.Handle(actionStart, h.Start)
	h.list.Handle(actionShowMe, h.Show(true))
	h.list.Handle(actionAdd, h.Service(textAddWish, lvlEdit, actionAdd))
	h.list.Handle(actionBack, h.Service(textGreetings, lvlStart, actionBack))
	h.list.Handle(actionDelete, h.Service(textDeleteWish, lvlEdit, actionDelete))
	h.list.Handle(actionPassword, h.Service(textEnterPassword, lvlEdit, actionPassword))
	h.list.Handle(actionShowUser, h.Service(textEnterUsername, lvlUser, actionShowUser))
}

func (h *Handle) Error(code int, user *session.User, err error) {
	h.log.Errorf(code, err,
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Name),
		slog.String("request", user.Request),
	)
}
