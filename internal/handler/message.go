package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/eugene-static/wishlist_bot/internal/bot"
	"github.com/eugene-static/wishlist_bot/internal/entity"
	"github.com/eugene-static/wishlist_bot/lib/lgr"
	"github.com/eugene-static/wishlist_bot/lib/random"
)

func (h *Handle) start(ctx context.Context, r *bot.Request) {
	user, err := h.getUser(ctx, r)
	if err != nil {
		h.error(nil, err)
		return
	}
	h.log.Debug("new user", slog.Int64("user_id", user.ID), slog.String("username", user.Name))
	h.send(user, lvlStart, textGreetings)
}

func (h *Handle) message(ctx context.Context, r *bot.Request) {
	user, err := h.getUser(ctx, r)
	if err != nil {
		h.error(nil, err)
		return
	}
	h.log.Debug("new message update",
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Name),
		slog.String("action", user.Action),
	)
	user.Request = r.Data
	r.Data = user.Action
	h.mux.ServeBot(ctx, r)
}

func (h *Handle) add(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, r *bot.Request) {
		user, err := h.getUser(ctx, r)
		if err != nil {
			h.error(nil, err)
			return
		}
		if err = h.service.AddWish(ctx, &entity.Wish{
			ID:      random.String(16),
			Content: user.Request,
			UserID:  user.ID,
		}); err != nil {
			h.errorCode(lgr.ErrAddWish, user, err)
			return
		}
		next(ctx, r)
	}
}

func (h *Handle) delete(next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, r *bot.Request) {
		user, err := h.getUser(ctx, r)
		if err != nil {
			h.error(nil, err)
			return
		}
		nums := strings.Fields(user.Request)
		if len(nums) > 0 {
			for i, num := range nums {
				index, err := strconv.Atoi(num)
				if err != nil || index > len(user.IDList) || index <= 0 {
					h.send(user, lvlEmpty, textWrongRequest)
					return
				}
				nums[i] = user.IDList[index-1]
			}
			if err = h.service.DeleteWishes(ctx, nums); err != nil {
				h.errorCode(lgr.ErrDelWish, user, err)
				return
			}
		}
		next(ctx, r)
	}
}

func (h *Handle) password(ctx context.Context, r *bot.Request) {
	user, err := h.getUser(ctx, r)
	if err != nil {
		h.error(nil, err)
		return
	}
	if user.Request == DeletePassword {
		user.Request = user.Name
	} else if strings.ContainsRune(user.Request, ' ') {
		h.send(user, lvlEmpty, textNoSpace)
		return
	}
	hashedPass, err := hash(user.Request)
	if err != nil {
		h.errorCode(lgr.ErrChangePass, user, err)
		return
	}
	if err = h.service.UpdateUser(ctx, user.ID, user.Name, hashedPass); err != nil {
		h.errorCode(lgr.ErrChangePass, user, err)
		return
	}
	h.send(user, lvlService, textSuccess)
}

func (h *Handle) show(me bool) bot.HandlerFunc {
	return func(ctx context.Context, r *bot.Request) {
		user, err := h.getUser(ctx, r)
		if err != nil {
			h.error(nil, err)
			return
		}
		var (
			id    int64
			level int
		)
		if !me {
			level = lvlUser
			req := strings.Fields(user.Request)
			switch len(req) {
			case 1:
				req = append(req, user.Name)
			case 2:
				break
			default:
				h.send(user, level, textWrongRequest)
				return
			}
			username, password := strings.TrimPrefix(req[0], "@"), []byte(req[1])
			reqUser, err := h.service.GetUserByUsername(ctx, username)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					h.send(user, level, textUserNotFound)
					return
				}
				h.errorCode(lgr.ErrGetUser, user, err)
				return
			}
			if compare(reqUser.Password, password) != nil {
				h.send(user, level, textWrongPassword)
				return
			}
			id = reqUser.ID
		} else {
			level = lvlEmptyList
			id = user.ID
		}
		list, err := h.service.GetWishlistByID(ctx, id)
		if err != nil {
			h.errorCode(lgr.ErrGetList, user, err)
			return
		}
		if list == nil {
			h.send(user, level, textNoWishes)
			return
		}
		if me {
			level = lvlMe
			user.IDList = make([]string, len(list))
		}
		var wishes strings.Builder
		for i, wish := range list {
			if me {
				user.IDList[i] = fmt.Sprintf("'%s'", wish.ID)
			}
			_, _ = wishes.WriteString(fmt.Sprintf("%d. %s\n", i+1, wish.Content))
		}
		h.bot.Config.SetReplyMessage(textWishList, wishes.String())
		h.send(user, level, textWishList)
	}
}
