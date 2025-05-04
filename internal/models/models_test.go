package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserDataFields(t *testing.T) {
	now := time.Now()
	id := uuid.New()

	user := UserData{
		ID:       id,
		Login:    "testuser",
		Password: "securepassword",
		Created:  now,
		Modified: now,
	}

	assert.Equal(t, id, user.ID)
	assert.Equal(t, "testuser", user.Login)
	assert.Equal(t, "securepassword", user.Password)
	assert.Equal(t, now, user.Created)
	assert.Equal(t, now, user.Modified)
}

func TestMetaFields(t *testing.T) {
	now := time.Now()
	id := uuid.New()
	userID := uuid.New()
	dataID := uuid.New()

	meta := Meta{
		ID:          id,
		Title:       "Sample Title",
		Description: "A description",
		Type:        "type",
		DataID:      dataID,
		UserID:      userID,
		Created:     now,
		Modified:    now,
	}

	assert.Equal(t, id, meta.ID)
	assert.Equal(t, "Sample Title", meta.Title)
	assert.Equal(t, "A description", meta.Description)
	assert.Equal(t, "type", meta.Type)
	assert.Equal(t, dataID, meta.DataID)
	assert.Equal(t, userID, meta.UserID)
	assert.Equal(t, now, meta.Created)
	assert.Equal(t, now, meta.Modified)
}

func TestItemDataFields(t *testing.T) {
	dataBytes := []byte{0x10, 0x20, 0x30}
	item := ItemData{
		ID:   uuid.New(),
		Data: dataBytes,
	}

	assert.NotNil(t, item.ID)
	assert.Equal(t, dataBytes, item.Data)
}
