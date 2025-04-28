package models

import (
	"time"

	"github.com/google/uuid"
)

type UserData struct {
	ID       uuid.UUID `json:"id"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

type CredsData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TextData struct {
	Text string `json:"text"`
}

type BankCardData struct {
	CardNum string `json:"card_num"`
	Expiry  string `json:"expiry"`
	CVV     string `json:"cvv"`
}

type Binary struct {
	Binary   BinaryData `json:"binary_data"`
	FilePath string     `json:"name"`
}

type BinaryData struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Content     []byte `json:"content"`
}

type Meta struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"data_type"`
	DataID      uuid.UUID `json:"data_id"`
	UserID      uuid.UUID `json:"user_id"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

type ItemData struct {
	ID   uuid.UUID `json:"id"`
	Data []byte    `json:"data"`
}

//TODO add OTP Data
