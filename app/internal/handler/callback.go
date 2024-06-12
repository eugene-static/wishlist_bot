package handler

import (
	"context"
	"log/slog"

	"github.com/eugene-static/wishlist_bot/app/internal/bot"
)

func (h *Handle) callback(code int, level int, action string) bot.HandlerFunc {
	return func(ctx context.Context, r *bot.Request) {
		user, err := h.getUser(ctx, r)
		if err != nil {
			h.error(nil, err)
			return
		}
		user.Action = action
		h.log.Debug("new callback update",
			slog.Int64("user_id", user.ID),
			slog.String("username", user.Name),
			slog.String("data", r.Data),
		)
		h.send(user, level, code)
	}
}
