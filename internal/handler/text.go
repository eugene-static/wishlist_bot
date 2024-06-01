package handler

import "github.com/eugene-static/wishlist_bot/lib/format"

const (
	textGreetings = 100 + iota
	textAddWish
	textDeleteWish
	textEnterPassword
	textWrongPassword
	textNoSpace
	textEnterUsername
	textSuccess
	textNoWishes
	textUserNotFound
	textWrongRequest
	textDefaultMessage
)

func text(code int) (text string) {
	switch code {
	case textGreetings:
		text = `Итак, чем займемся?`
	case textAddWish:
		text = `Введи описание и/или ссылку и отправь в чат одним сообщением:`
	case textDeleteWish:
		text = "Введи через пробелы номера желаний из списка, которые нужно удалить. Например:\n" +
			format.Format("1 3 10 6", format.Monotype)
	case textEnterPassword:
		text = "Пароль необходим для того, чтобы к твоему вишлисту был доступ только у тех, кто знает пароль. " +
			"Им ты можешь делиться лично с кем-то или же опубликовать в своем профиле. " +
			"Пароль может быть в любой форме, но не должен содержать пробелы. Например:" +
			"🐈‍⬛💥💽\n" +
			"Чтобы сбросить пароль и сделать вишлист общедоступным, введи:\n" +
			format.Format(DeletePassword, format.Monotype)
	case textWrongPassword:
		text = "Похоже, вишлист пользователя защищен паролем. Поищи пароль в его в профиле, либо же обратись к нему лично"
	case textEnterUsername:
		text = "Введи юзернейм пользователя, чей вишлист ты хочешь посмотреть. Юзернейм можно найти в профиле пользователя.\n" +
			"Если вишлист выбранного пользователя защищён паролем, то через пробел введи пароль. Например:\n" +
			format.Format(admin+" пароль", format.Monotype)
	case textSuccess:
		text = "Успешно"
	case textNoWishes:
		text = "Здесь нет ни одного желания..."
	case textUserNotFound:
		text = "Похоже, у этого пользователя нет вишлиста"
	case textWrongRequest:
		text = "В запросе ошибка, попробуй снова"
	case textNoSpace:
		text = "В пароле не должно содержаться пробелов. Попробуй другой"
	case textDefaultMessage:
		text = "Не могу обработать сообщение"
	}
	return text
}
