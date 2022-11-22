package dto

type SubscribeRequest struct {
	TelegramID int64  `json:"telegram_id"`
	AdvertURL  string `json:"advert_url"`
}
