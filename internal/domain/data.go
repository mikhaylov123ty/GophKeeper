package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserData represents a user in the system with unique ID, login credentials, and timestamps for creation and modification.
type UserData struct {
	ID       uuid.UUID `json:"id"`
	Login    string    `json:"login"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
	Modified time.Time `json:"modified"`
}

// Meta represents metadata associated with an item or resource, including identification, ownership, and timestamps.
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

// ItemData represents an entity containing a unique identifier and associated byte data.
type ItemData struct {
	ID   uuid.UUID `json:"id"`
	Data []byte    `json:"data"`
}

//TODO add OTP Data
