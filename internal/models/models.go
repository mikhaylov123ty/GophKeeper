package models

import (
	"github.com/google/uuid"
	"time"
)

type UserData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type TextData struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

type BankCardData struct {
	ID      uuid.UUID `json:"id"`
	CardNum string    `json:"card_num"`
	Expiry  time.Time `json:"expiry"`
	CVV     string    `json:"cvv"`
}

type BinaryData struct {
	ID     uuid.UUID `json:"id"`
	Binary []byte    `json:"binary"`
	Name   string    `json:"name"`
}

type Meta struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"data_type"`
	DataID      uuid.UUID `json:"data_id"`
	Created     time.Time `json:"created"`
	Modified    time.Time `json:"modified"`
}

//TODO add OTP Data
