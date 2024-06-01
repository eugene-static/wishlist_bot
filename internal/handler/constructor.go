package handler

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	buttonMyWishlist = "Мой вишлист"
	buttonFindUser   = "Найти пользователя"
	buttonAdd        = "Добавить"
	buttonDelete     = "Удалить"
	buttonPassword   = "Пароль"
	buttonBack       = "Назад"
	buttonCancel     = "Отмена"
	buttonOK         = "ОК"
)

const admin = "@eugene_static"

type messageConfig struct {
	userID int64
	text   string
	level  int
}

func newConfig(userID int64) *messageConfig {
	return &messageConfig{
		userID: userID,
		level:  lvlEmpty,
	}
}

func (mc *messageConfig) build() tgbotapi.MessageConfig {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           mc.userID,
			ReplyToMessageID: 0,
		},
		Text:                  mc.text,
		ParseMode:             tgbotapi.ModeHTML,
		DisableWebPagePreview: true,
	}
	markUp := tgbotapi.NewInlineKeyboardMarkup()
	switch mc.level {
	case lvlStart:
		markUp = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonMyWishlist, actionShowMe),
				tgbotapi.NewInlineKeyboardButtonData(buttonFindUser, actionShowUser)))
	case lvlMe:
		markUp = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonAdd, actionAdd),
				tgbotapi.NewInlineKeyboardButtonData(buttonDelete, actionDelete),
				tgbotapi.NewInlineKeyboardButtonData(buttonPassword, actionPassword),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonBack, actionBack),
			))
	case lvlUser:
		markUp = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonBack, actionBack)))
	case lvlService:
		markUp = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonOK, actionShowMe)))
	case lvlEdit:
		markUp = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonCancel, actionShowMe)))
	case lvlServiceExt:
		markUp = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(buttonCancel, actionBack),
				tgbotapi.NewInlineKeyboardButtonData(buttonOK /*TODO*/, "nil")))
	default:
		return msg
	}
	msg.ReplyMarkup = markUp
	return msg
}

func (mc *messageConfig) error(code int) tgbotapi.MessageConfig {
	mc.text = fmt.Sprintf("В работе бота возникла ошибка. Код %03o\nПопробуйте снова позже или же обратитесь к %s за помощью",
		code, admin)
	msg := tgbotapi.NewMessage(mc.userID, mc.text)
	return msg
}
