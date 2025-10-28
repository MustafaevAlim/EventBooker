package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wb-go/wbf/zlog"
)

type TelegramBot struct {
	bot *tgbotapi.BotAPI
	m   sync.Mutex
}

const maxRetry = 5

type RetryMessage struct {
	ChatID   int64
	Attempts int
	Text     string
}

func NewTelegramBot(token string) *TelegramBot {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {

		log.Println(err)
		return nil
	}

	return &TelegramBot{bot: bot, m: sync.Mutex{}}
}

func (tg *TelegramBot) ListenUpdated(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := tg.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return

		case update, ok := <-updates:
			if !ok {
				zlog.Logger.Info().Msg("ListenUpdated: канал updates закрыт")
				return
			}

			if update.Message == nil {
				continue
			}

			if update.Message.IsCommand() {
				command := update.Message.Command()
				chatID := update.Message.Chat.ID

				switch command {
				case "start":
					username := update.Message.Chat.UserName

					zlog.Logger.Info().Msgf("Получено сообщение от %s [%d]: %s", username, chatID, update.Message.Text)

					msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Ваш ChatId: %d", chatID))
					if _, err := tg.bot.Send(msg); err != nil {
						zlog.Logger.Error().Msgf("Ошибка отправки сообщения: %v", err)
					}
				default:
					msg := tgbotapi.NewMessage(chatID, "Неизвестная команда")
					if _, err := tg.bot.Send(msg); err != nil {
						zlog.Logger.Error().Msgf("Ошибка отправки сообщения: %v", err)
					}
				}
			}

		}
	}
}

func (tg *TelegramBot) Send(chatId int64, text string) error {
	str := "Новое уведомление!\n" + text
	msg := tgbotapi.NewMessage(chatId, str)
	tg.m.Lock()
	defer tg.m.Unlock()
	_, err := tg.bot.Send(msg)
	if err != nil {
		zlog.Logger.Error().Msgf("service.TelegramBot.Send error: %v", err)
		return err
	}
	return nil
}

func (tg *TelegramBot) RetryWorker(ctx context.Context, retryQueue chan RetryMessage) {
	for rm := range retryQueue {
		if tg == nil {
			log.Println(rm.Text)
			continue
		}
		err := tg.Send(rm.ChatID, rm.Text)
		if err != nil {
			zlog.Logger.Error().Msgf("service.TelegramBot.RetryWorker error: %v", err)
			rm.Attempts++

			if rm.Attempts < maxRetry {
				go func(r RetryMessage) {
					select {
					case <-time.After(time.Second * 3 * time.Duration(r.Attempts)):
						select {
						case <-ctx.Done():
							return
						case retryQueue <- r:
						}
					case <-ctx.Done():
						return
					}
				}(rm)
			}

		}
	}
}
