package notify

import (
	"errors"
	"fmt"
	domain "parser/internal/domain/models"
	"parser/internal/telegram"
)

var (
	ErrNoArgs = errors.New("args could not be empty")

	ErrNoChatIdentified = errors.New("missing userID or chatID in args")
	ErrInvalidIDFormat  = errors.New("ID should be int64")

	ErrNoMessage            = errors.New("missing msg in args")
	ErrInvalidMessageFormat = errors.New("message should be string")
)

type notifyArgs struct {
	chatIdentifier int64
	message        string
}

type telegramNotifier struct {
	tg telegram.Telegram
}

func NewTelegramNotifier(telegram telegram.Telegram) Notifier {
	return &telegramNotifier{tg: telegram}
}

// args[0] - userID, chatID to sent message to (int64)
// args[1] - message that's sent to end user (string)
//
func (tn *telegramNotifier) Notify(target *domain.Advert, args ...interface{}) error {
	nargs, err := tn.validateArgs(&args)
	if err != nil {
		return err
	}

	err = tn.tg.SendMessage(nargs.chatIdentifier, nargs.message)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}

	return nil
}

func (tn *telegramNotifier) validateArgs(args *[]interface{}) (*notifyArgs, error) {
	if len(*args) == 0 {
		return nil, ErrNoArgs
	}

	var chatIdentifier int64
	var message string

	if len(*args) > 0 {
		candidateID := (*args)[0]

		chatIDInt, ok := candidateID.(int64)
		if !ok {
			return nil, ErrInvalidIDFormat
		}

		chatIdentifier = chatIDInt
	}

	if len(*args) > 1 {
		candidateMsg := (*args)[1]

		msgStr, ok := candidateMsg.(string)
		if !ok {
			return nil, ErrInvalidMessageFormat
		}
		message = msgStr
	}

	return &notifyArgs{
		chatIdentifier: chatIdentifier,
		message:        message,
	}, nil
}
