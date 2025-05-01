package models

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Test the MetaItem.FilterValue method
func TestMetaItem_FilterValue(t *testing.T) {
	item := MetaItem{
		ID:          uuid.New(),
		Title:       "Test Title",
		Description: "Test Description",
		DataID:      "Data123",
		Created:     "2024-01-01",
		Modified:    "2024-01-02",
	}

	assert.Equal(t, "", item.FilterValue(), "FilterValue should return empty string")
}

// Helper to create a MetaItem for rendering
func createMetaItem(title, description, created, modified string) *MetaItem {
	return &MetaItem{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		DataID:      "DataID",
		Created:     created,
		Modified:    modified,
	}
}

// Test the MetaItemDelegate.Render method
func TestMetaItemDelegate_Render(t *testing.T) {
	delegate := MetaItemDelegate{}

	// Create a list.Model with an index
	l := list.New(nil, delegate, 10, 10)
	l.Select(0)

	// Create a MetaItem
	item := createMetaItem("Title1", "Description1", "2024-01-01", "2024-01-02")

	// Prepare a buffer to capture output
	var buf bytes.Buffer

	// Call Render with index 0 (selected)
	delegate.Render(&buf, l, 0, item)

	output := buf.String()

	// Check if output contains expected strings
	assert.Contains(t, output, "Title1")
	assert.Contains(t, output, "Description1")
	assert.Contains(t, output, "2024-01-01")
	assert.Contains(t, output, "2024-01-02")
	assert.Contains(t, output, "[x]") // selected item indicator

	// Now test with an unselected index
	buf.Reset()
	l.Select(1)
	delegate.Render(&buf, l, 0, item) // index 0, but selected index is 1
	output2 := buf.String()

	// It should not have the "[x]" indicator since it's not selected
	assert.NotContains(t, output2, "[x]")
}

// Test the Render method with a non-MetaItem item (should do nothing)
func TestMetaItemDelegate_Render_NonMetaItem(t *testing.T) {
	delegate := MetaItemDelegate{}

	var buf bytes.Buffer
	l := list.New(nil, delegate, 10, 10)

	// Should not panic
	delegate.Render(&buf, l, 0, nil)
	assert.Empty(t, buf.String())
}
