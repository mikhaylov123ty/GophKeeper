package tui

import (
	"context"
	"fmt"
	grpcLib "google.golang.org/grpc"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/screens"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

const (
	mB           = 1048576
	messageLimit = 60 * mB
)

// ItemsManager is responsible for managing metadata items, gRPC client interactions, and user authentication data.
type ItemsManager struct {
	metaItems  map[string][]*models.MetaItem
	grpcClient *grpc.Client
	userID     string
}

// NewItemManager - Initializes and returns a new instance of an `ItemManager`,
// which is responsible for managing TUI interactions and connecting with gRPC services.
func NewItemManager(grpcClient *grpc.Client) (*models.Model, error) {
	im := ItemsManager{
		metaItems:  map[string][]*models.MetaItem{},
		grpcClient: grpcClient,
	}

	mainMenu := screens.NewMainMenu([]string{
		screens.TextCategory,
		screens.CredsCategory,
		screens.FileCategory,
		screens.CardCategory,
		screens.ExitCategory,
	}, &im)

	auth := screens.NewAuthScreen(mainMenu, &im)

	return auth, nil
}

// GetMetaData retrieves metadata items associated with a specific category.
func (im *ItemsManager) GetMetaData(category string) []*models.MetaItem {
	return im.metaItems[category]
}

// SaveMetaItem saves a new metadata item into the `ItemsManager` under a specific category.
func (im *ItemsManager) SaveMetaItem(category string, newItem *models.MetaItem) {
	im.metaItems[category] = append(im.metaItems[category], newItem)
}

// PostItemData sends item data and metadata to the associated gRPC service after encrypting the data.
func (im *ItemsManager) PostItemData(data []byte, dataID string, metaData *pb.MetaData) (*pb.PostItemDataResponse, error) {
	encryptedData, err := utils.EncryptData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	metaData.UserId = im.userID

	resp, err := im.grpcClient.Handlers.ItemDataHandler.PostItemData(context.Background(),
		&pb.PostItemDataRequest{
			Data:     encryptedData,
			DataId:   dataID,
			MetaData: metaData,
		},
		grpcLib.MaxCallRecvMsgSize(messageLimit),
		grpcLib.MaxCallSendMsgSize(messageLimit),
	)
	if err != nil {
		return nil, fmt.Errorf("post item failed:  %w,", err)
	}

	return resp, err
}

// GetItemData retrieves the item data associated with the given data ID,
// decrypts it using the utility functions, and returns the decrypted data as a string.
func (im *ItemsManager) GetItemData(dataID string) (string, error) {
	response, err := im.grpcClient.Handlers.ItemDataHandler.GetItemData(context.Background(), &pb.GetItemDataRequest{
		DataId: dataID,
	},
		grpcLib.MaxCallRecvMsgSize(messageLimit),
		grpcLib.MaxCallSendMsgSize(messageLimit),
	)
	if err != nil {
		return "", fmt.Errorf("could not get text data: %w", err)
	}

	decryptedData, err := utils.DeryptData(response.Data)
	if err != nil {
		return "", fmt.Errorf("failed decrypt data: %w", err)
	}

	return string(decryptedData), nil

}

// PostUserData sends user credentials to the authentication service, retrieves a user ID and JWT token, and stores them.
func (im *ItemsManager) PostUserData(login string, password string) error {
	res, err := im.grpcClient.Handlers.AuthHandler.PostUserData(context.Background(), &pb.PostUserDataRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("failed login: %w", err)
	}

	if res.Error != "" {
		return fmt.Errorf("failed login: %s", res.Error)
	}

	if res.UserId == "" {
		return fmt.Errorf("failed login: empty user id")
	}

	im.userID = res.UserId
	im.grpcClient.JWTToken = res.Jwt

	return nil
}

// SyncMeta synchronizes metadata by retrieving and storing metadata items for the current user from the gRPC service.
func (im *ItemsManager) SyncMeta() error {
	metaItems, err := im.grpcClient.Handlers.MetaDataHandler.GetMetaData(context.Background(),
		&pb.GetMetaDataRequest{UserId: im.userID})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				return nil
			}
		} else {
			return fmt.Errorf("failed to get meta data: %w", err)
		}
	}

	for _, metaItem := range metaItems.Items {
		id, err := uuid.Parse(metaItem.GetId())
		if err != nil {
			return fmt.Errorf("invalid meta item id: %s", metaItem.GetId())
		}
		im.metaItems[metaItem.DataType] = append(im.metaItems[metaItem.DataType], &models.MetaItem{
			ID:          id,
			Title:       metaItem.GetTitle(),
			Description: metaItem.GetDescription(),
			DataID:      metaItem.GetDataId(),
			Created:     metaItem.GetCreated(),
			Modified:    metaItem.GetModified(),
		})
	}

	return nil
}

// DeleteItem removes a metadata item by its ID, category, and data ID, and updates the local metadata cache.
func (im *ItemsManager) DeleteItem(metaItemID uuid.UUID, category string, dataID string) error {
	resp, err := im.grpcClient.Handlers.MetaDataHandler.DeleteMetaData(context.Background(), &pb.DeleteMetaDataRequest{
		MetadataId:   metaItemID.String(),
		MetadataType: category,
		DataId:       dataID,
	})
	if err != nil && resp.GetError() != "" {
		return fmt.Errorf("could not delete meta data: %w", err)
	}

	for i, v := range im.metaItems[category] {
		if v.ID == metaItemID {
			im.metaItems[category] = append(im.metaItems[category][:i], im.metaItems[category][i+1:]...)
		}
	}

	return nil
}
