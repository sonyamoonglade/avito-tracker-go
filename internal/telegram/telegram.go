package telegram

import (
	"errors"
	"fmt"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	ErrNoToken = errors.New("token must not be empty")
)

const (
	pollTimeout int = 60
)

type Telegram interface {
	// TODO: ctx
	SendMessage(chatIdentifier int64, msg string) error

	// Starts the bot to poll telegram api and receive updates
	Connect(token string) error
	Close()
}

type telegram struct {
	client *tg.BotAPI
}

func NewTelegram() Telegram {
	return &telegram{
		client: nil,
	}
}

func (t *telegram) Connect(token string) error {
	if token == "" {
		return ErrNoToken
	}

	bot, err := tg.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("unable to connect: %w", err)
	}
	bot.Debug = true

	// todo: custom timeout
	updates := bot.GetUpdatesChan(tg.UpdateConfig{
		Timeout: pollTimeout,
	})

	t.client = bot

	for update := range updates {
		fmt.Printf("ID: %d\n", update.SentFrom().ID)
	}

	return nil
}

func (t *telegram) Close() {
	t.client.StopReceivingUpdates()
}

// TODO: ctx
// TODO: logger
func (t *telegram) SendMessage(chatIdentifier int64, msg string) error {
	m := t.newEmptyMessage(chatIdentifier, msg)
	err := t.send(m)
	if err != nil {
		return fmt.Errorf("unable to send message: %w", err)
	}

	return nil
}

func (t *telegram) send(ch tg.Chattable) error {
	_, err := t.client.Send(ch)
	return err
}

func (t *telegram) newEmptyMessage(chatIdentifier int64, text string) tg.MessageConfig {
	return tg.NewMessage(chatIdentifier, text)
}
