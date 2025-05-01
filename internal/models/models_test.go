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

func TestCredsDataFields(t *testing.T) {
	creds := CredsData{
		Login:    "user123",
		Password: "pass123",
	}

	assert.Equal(t, "user123", creds.Login)
	assert.Equal(t, "pass123", creds.Password)
}

func TestTextDataFields(t *testing.T) {
	text := TextData{
		Text: "Sample text",
	}

	assert.Equal(t, "Sample text", text.Text)
}

func TestBankCardDataFields(t *testing.T) {
	card := BankCardData{
		CardNum: "4111111111111111",
		Expiry:  "12/25",
		CVV:     "123",
	}

	assert.Equal(t, "4111111111111111", card.CardNum)
	assert.Equal(t, "12/25", card.Expiry)
	assert.Equal(t, "123", card.CVV)
}

func TestBinaryAndBinaryData(t *testing.T) {
	content := []byte{0x00, 0x01, 0x02}
	binaryData := BinaryData{
		Name:        "file.txt",
		ContentType: "text/plain",
		Content:     content,
	}
	binary := Binary{
		Binary:   binaryData,
		FilePath: "path/to/file.txt",
	}

	assert.Equal(t, "file.txt", binary.Binary.Name)
	assert.Equal(t, "text/plain", binary.Binary.ContentType)
	assert.Equal(t, content, binary.Binary.Content)
	assert.Equal(t, "path/to/file.txt", binary.FilePath)
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
