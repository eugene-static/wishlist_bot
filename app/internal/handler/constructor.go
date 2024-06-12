package handler

import (
	"github.com/eugene-static/wishlist_bot/app/internal/bot"
	"github.com/eugene-static/wishlist_bot/app/lib/format"
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

const (
	actionAdd       = "/add"
	actionDelete    = "/delete"
	actionPassword  = "/password"
	actionShowUser  = "/show_user"
	actionShowMe    = "/show_me"
	actionBack      = "/back"
	messageStart    = "/start"
	messageAdd      = "/message_add"
	messageDelete   = "/message_delete"
	messageShowUser = "/message_show_user"
	messagePassword = "/message_password"
)

const (
	deleteAllWishes = "Удалить всё"
	deletePassword  = "Удалить пароль"
)

const (
	lvlEmpty = iota
	lvlStart
	lvlEmptyList
	lvlMe
	lvlUser
	lvlEdit
	lvlService
	lvlServiceExt
)

const (
	textGreetings = 100 + iota
	textAddWish
	textDeleteWish
	textEnterPassword
	textWrongPassword
	textNoSpace
	textEnterUsername
	textSuccess
	textWishList
	textNoWishes
	textUserNotFound
	textWrongRequest
	textDefaultMessage
	textError
)

const (
	errGetUser = iota
	errGetList
	errAddWish
	errDelWish
	errChangePass
)

func (h *Handle) Register() {
	h.mux.Handle(bot.DefaultMessage, h.message)
	h.mux.Handle(messageAdd, h.add(h.showMe))
	h.mux.Handle(messageDelete, h.delete(h.showMe))
	h.mux.Handle(messageShowUser, h.showUser)
	h.mux.Handle(messagePassword, h.password)
	h.mux.Handle(messageStart, h.start)
	h.mux.Handle(actionShowMe, h.showMe)
	h.mux.Handle(actionBack, h.callback(textGreetings, lvlStart, messageStart))
	h.mux.Handle(actionAdd, h.callback(textAddWish, lvlEdit, messageAdd))
	h.mux.Handle(actionDelete, h.callback(textDeleteWish, lvlEdit, messageDelete))
	h.mux.Handle(actionPassword, h.callback(textEnterPassword, lvlEdit, messagePassword))
	h.mux.Handle(actionShowUser, h.callback(textEnterUsername, lvlUser, messageShowUser))
}

func (h *Handle) SetConfig() {
	msg := bot.NewConfig(bot.ModeHTML)
	h.bot.Config.Set(lvlEmpty, msg)
	msg.ReplyMarkup = bot.NewMarkup(
		bot.NewRow(
			bot.NewButton(buttonMyWishlist, actionShowMe),
			bot.NewButton(buttonFindUser, actionShowUser)))
	h.bot.Config.Set(lvlStart, msg)
	msg.ReplyMarkup = bot.NewMarkup(
		bot.NewRow(
			bot.NewButton(buttonAdd, actionAdd),
			bot.NewButton(buttonDelete, actionDelete),
			bot.NewButton(buttonPassword, actionPassword),
		),
		bot.NewRow(
			bot.NewButton(buttonBack, actionBack),
		))
	h.bot.Config.Set(lvlMe, msg)
	msg.ReplyMarkup = bot.NewMarkup(
		bot.NewRow(
			bot.NewButton(buttonAdd, actionAdd),
			bot.NewButton(buttonPassword, actionPassword),
		),
		bot.NewRow(
			bot.NewButton(buttonBack, actionBack),
		))
	h.bot.Config.Set(lvlEmptyList, msg)
	msg.ReplyMarkup = bot.NewMarkup(
		bot.NewRow(
			bot.NewButton(buttonBack, actionBack)))
	h.bot.Config.Set(lvlUser, msg)
	msg.ReplyMarkup = bot.NewMarkup(
		bot.NewRow(
			bot.NewButton(buttonOK, actionShowMe)))
	h.bot.Config.Set(lvlService, msg)
	msg.ReplyMarkup = bot.NewMarkup(
		bot.NewRow(
			bot.NewButton(buttonCancel, actionShowMe)))
	h.bot.Config.Set(lvlEdit, msg)
	//
	h.bot.Config.SetReplyMessage(textGreetings, "Итак, чем займемся?")
	h.bot.Config.SetReplyMessage(textAddWish, "Введи описание и/или ссылку и отправь в чат одним сообщением:")
	h.bot.Config.SetReplyMessage(textDeleteWish, "Введи через пробелы номера желаний из списка, которые нужно удалить. Например:\n"+
		format.Format("1 3 10 6\n", format.Monotype)+
		"Если хочешь удалить весь список, введи "+format.Format(deleteAllWishes, format.Monotype))
	h.bot.Config.SetReplyMessage(textEnterPassword, "Пароль необходим для того, чтобы к твоему вишлисту был доступ только у тех, кто знает пароль. "+
		"Им ты можешь делиться лично с кем-то или же опубликовать в своем профиле. "+
		"Пароль может быть в любой форме, но не должен содержать пробелы. Например:"+
		"🐈‍⬛💥💽\n"+
		"Чтобы сбросить пароль и сделать вишлист общедоступным, введи:\n"+
		format.Format(deletePassword, format.Monotype))
	h.bot.Config.SetReplyMessage(textWrongPassword, "Неверный пароль. Поищи пароль в профиле пользователя либо же обратись к нему лично")
	h.bot.Config.SetReplyMessage(textEnterUsername, "Введи юзернейм пользователя, чей вишлист ты хочешь посмотреть. Юзернейм можно найти в профиле пользователя.\n"+
		"Если вишлист выбранного пользователя защищён паролем, то через пробел введи пароль. Например:\n"+
		format.Format(admin+" пароль", format.Monotype))
	h.bot.Config.SetReplyMessage(textSuccess, "Успешно")
	h.bot.Config.SetReplyMessage(textNoWishes, "Здесь нет ни одного желания...")
	h.bot.Config.SetReplyMessage(textUserNotFound, "Похоже, у этого пользователя нет вишлиста")
	h.bot.Config.SetReplyMessage(textWrongRequest, "В запросе ошибка, попробуй снова")
	h.bot.Config.SetReplyMessage(textNoSpace, "В пароле не должно содержаться пробелов. Попробуй другой")
	h.bot.Config.SetReplyMessage(textDefaultMessage, "Не могу обработать сообщение")
}

func (h *Handle) SetErrors() {
	h.log.Set(errGetUser, "getting user error")
	h.log.Set(errGetList, "getting list error")
	h.log.Set(errAddWish, "adding wish error")
	h.log.Set(errDelWish, "deleting wish error")
	h.log.Set(errChangePass, "changing password error")
}
