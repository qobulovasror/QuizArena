// Package telegram — Telegram bot: /start → Mini App ochish tugmasi.
package telegram

import (
	"context"
	"log/slog"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

// Run — bot'ni ishga tushiradi (token bo'sh bo'lsa skip). Bloklaydi — goroutine'da chaqiring.
func Run(ctx context.Context, token, miniAppURL string, logger *slog.Logger) {
	if token == "" {
		logger.Info("telegram bot o'chiq (TELEGRAM_BOT_TOKEN yo'q)")
		return
	}
	bot, err := telego.NewBot(token)
	if err != nil {
		logger.Error("telegram: bot yaratish", "err", err)
		return
	}
	updates, err := bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		logger.Error("telegram: updates", "err", err)
		return
	}
	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		logger.Error("telegram: handler", "err", err)
		return
	}
	defer func() { _ = bh.Stop() }()

	// /start → Mini App ochuvchi inline tugma.
	bh.HandleMessage(func(c *th.Context, message telego.Message) error {
		kb := tu.InlineKeyboard(tu.InlineKeyboardRow(
			tu.InlineKeyboardButton("🎮 O'ynash").WithWebApp(&telego.WebAppInfo{URL: miniAppURL}),
		))
		_, err := c.Bot().SendMessage(c, tu.Message(
			tu.ID(message.Chat.ID),
			"QuizArena'ga xush kelibsiz! 🎮\nO'ynash uchun pastdagi tugmani bosing.",
		).WithReplyMarkup(kb))
		return err
	}, th.CommandEqual("start"))

	logger.Info("telegram bot ishga tushdi", "miniApp", miniAppURL)
	if err := bh.Start(); err != nil {
		logger.Error("telegram: bot to'xtadi", "err", err)
	}
}
