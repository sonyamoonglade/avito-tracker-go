package notify

import domain "parser/internal/domain/models"

type Notifier interface {
	Notify(ad *domain.Advert, args ...interface{}) error
}
