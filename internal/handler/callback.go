package handler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/eugene-static/wishlist_bot/internal/session"
	"github.com/eugene-static/wishlist_bot/lib/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (h *Handle) onCallback(ctx context.Context, user *session.User) tgbotapi.MessageConfig {
	log := h.log.With(
		slog.Int64("user_id", user.ID),
		slog.String("username", user.Name),
		slog.String("data", user.Request),
	)
	log.Info("new callback update")
	mc := &messageConfig{userID: user.ID}
	switch user.Request {
	case actionBack:
		mc.text = text(textGreetings)
		mc.level = lvlStart
	case actionAdd:
		mc.text = text(textAddWish)
		mc.level = lvlEdit
	case actionDelete:
		mc.text = text(textDeleteWish)
		mc.level = lvlEdit
	case actionPassword:
		mc.text = text(textEnterPassword)
		mc.level = lvlEdit
	case actionShowUser:
		mc.text = text(textEnterUsername)
		mc.level = lvlUser
	case actionShowMe:
		list, err := h.service.GetWishlistByID(ctx, user.ID)
		if err != nil {
			log.Errorf(lgr.ErrGetList, err)
			return mc.error(lgr.ErrGetList)
		}
		mc.level = lvlMe
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
		mc.text = wishes.String()
	}
	user.Action = user.Request
	return mc.build()
}
