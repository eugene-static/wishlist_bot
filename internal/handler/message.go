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
	"github.com/eugene-static/wishlist_bot/lib/lgr"
	"github.com/eugene-static/wishlist_bot/lib/random"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handle) onMessage(ctx context.Context, user *session.User) tgbotapi.MessageConfig {
	log := h.log.With(
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Name),
		slog.String("action", user.Action),
	)
	log.Info("new message update")
	mc := &messageConfig{userID: user.ID, level: lvlEmpty}
	switch user.Request {
	case actionStart:
		mc.text = text(textGreetings)
		mc.level = lvlStart
		return mc.build()
	}
	switch user.Action {
	case actionAdd:
		if err := h.service.AddWish(ctx, &entity.Wish{
			ID:      random.String(16),
			Content: user.Request,
			UserID:  user.ID,
		}); err != nil {
			log.Errorf(lgr.ErrAddWish, err)
			return mc.error(lgr.ErrAddWish)
		}
	case actionDelete:
		nums := strings.Fields(user.Request)
		if len(nums) > 0 {
			for i, num := range nums {
				index, err := strconv.Atoi(num)
				if err != nil || index > len(user.IDList) || index <= 0 {
					mc.text = text(textWrongRequest)
					return mc.build()
				}
				nums[i] = fmt.Sprintf("'%s'", user.IDList[index-1])
			}
			if err := h.service.DeleteWishes(ctx, nums); err != nil {
				log.Errorf(lgr.ErrDelWish, err)
				return mc.error(lgr.ErrDelWish)
			}
		}
	case actionPassword:
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
			log.Errorf(lgr.ErrChangePass, err)
			return mc.error(lgr.ErrChangePass)
		}
		if err = h.service.UpdateUser(ctx, user.ID, user.Name, hashedPass); err != nil {
			log.Errorf(lgr.ErrChangePass, err)
			return mc.error(lgr.ErrChangePass)
		}
		mc.text = text(textSuccess)
		return mc.build()
	case actionShowUser:
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
			log.Errorf(lgr.ErrGetUser, err)
			return mc.error(lgr.ErrGetUser)
		}
		if bcrypt.CompareHashAndPassword(reqUser.Password, password) != nil {
			mc.text = text(textWrongPassword)
			return mc.build()
		}
		list, err := h.service.GetWishlistByID(ctx, reqUser.ID)
		if err != nil {
			log.Errorf(lgr.ErrGetList, err)
			return mc.error(lgr.ErrGetList)
		}
		if list == nil {
			mc.text = text(textNoWishes)
			return mc.build()
		}
		var wishes strings.Builder
		for i, wish := range list {
			_, _ = wishes.WriteString(fmt.Sprintf("%d. %s\n", i+1, wish.Content))
		}
		mc.text = wishes.String()
		return mc.build()
	default:
		mc.text = text(textWrongRequest)
		mc.level = lvlUser
		return mc.build()
	}
	mc.level = lvlMe
	list, err := h.service.GetWishlistByID(ctx, user.ID)
	if err != nil {
		log.Errorf(lgr.ErrGetList, err)
		return mc.error(lgr.ErrGetList)
	}
	if list == nil {
		mc.text = text(textNoWishes)
		return mc.build()
	}
	user.IDList = make([]string, len(list))
	var wishes strings.Builder
	for i, wish := range list {
		user.IDList[i] = wish.ID
		_, _ = wishes.WriteString(fmt.Sprintf("%d. %s\n", i+1, wish.Content))
	}
	user.Action = actionStart
	mc.text = wishes.String()
	return mc.build()
}
