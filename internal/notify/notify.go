package notify

import domain "parser/internal/domain/models"

type Notifier interface {
	// `args` are specific for every Notifier impl.
	// Firstly, look into the concrete implementation's args
	// e.g. TelegramNotifier has it's own args rules...
	Notify(ad *domain.Advert, args ...interface{}) error
}
