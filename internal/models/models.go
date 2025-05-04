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
