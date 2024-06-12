package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const ModeHTML = tgbotapi.ModeHTML

type Sender interface {
	Send(int, int64, string) (int, error)
}

type Bot struct {
	Config *Config
	bot    *tgbotapi.BotAPI
}

func NewBot(bot *tgbotapi.BotAPI) *Bot {
	return &Bot{
		Config: &Config{
			config:       map[int]tgbotapi.MessageConfig{},
			replyMessage: make(map[int]string),
		},
		bot: bot,
	}
}

func (b *Bot) Send(id int64, configKey int, messageKey int) (int, error) {
	c := b.Config.get(configKey, messageKey)
	c.ChatID = id
	m, err := b.bot.Send(c)
	if err != nil {
		return -1, err
	}
	return m.MessageID, nil
}

type Config struct {
	config       map[int]tgbotapi.MessageConfig
	replyMessage map[int]string
}

func NewConfig(parseMode string) tgbotapi.MessageConfig {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ReplyToMessageID: 0,
		},
		ParseMode:             parseMode,
		DisableWebPagePreview: true,
	}
	return msg
}

func (c *Config) Set(key int, cfg tgbotapi.MessageConfig) {
	c.config[key] = cfg
}

func (c *Config) SetReplyMessage(key int, message string) {
	c.replyMessage[key] = message
}

func (c *Config) get(configKey int, messageKey int) tgbotapi.MessageConfig {
	if cfg, ok := c.config[configKey]; ok {
		cfg.Text = c.getReplyMessage(messageKey)
		return cfg
	}
	return tgbotapi.MessageConfig{}
}

func (c *Config) getReplyMessage(key int) string {
	if m, ok := c.replyMessage[key]; ok {
		return m
	}
	return "Невозможно обработать запрос"
}

func NewButton(name string, data string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(name, data)
}

func NewRow(buttons ...tgbotapi.InlineKeyboardButton) []tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardRow(buttons...)
}

func NewMarkup(rows ...[]tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
