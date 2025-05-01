package screens

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type itemScreen struct {
	itemsManager models.ItemsManager
	category     string
	newTitle     string
	newDesc      string
	cursor       int
	backScreen   models.Screen
	selectedItem *models.MetaItem
}

func (is *itemScreen) postItemData(itemData []byte) error {
	var id uuid.UUID
	var dataID string
	if is.selectedItem != nil {
		id = is.selectedItem.ID
		dataID = is.selectedItem.DataID
	} else {
		id = uuid.New()
	}

	newItem := models.MetaItem{
		ID:          id,
		Title:       is.newTitle,
		Description: is.newDesc,
	}

	metaData := pb.MetaData{
		Id:          newItem.ID.String(),
		Title:       newItem.Title,
		Description: newItem.Description,
		DataType:    is.category,
	}

	resp, err := is.itemsManager.PostItemData(itemData, dataID, &metaData)
	if err != nil {
		return fmt.Errorf("failed to post item data: %w", err)
	}

	if is.selectedItem != nil {
		is.selectedItem.Title = is.newTitle
		is.selectedItem.Description = is.newDesc
		is.selectedItem.Modified = resp.Modified
	} else {
		newItem.DataID = resp.DataId
		newItem.Created = resp.Created
		newItem.Modified = resp.Modified

		is.itemsManager.SaveMetaItem(is.category, &newItem)
	}

	return nil
}
