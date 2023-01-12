package model

type Message struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}
