package tui

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/models"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/screens"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/app/tui/utils"
	"github.com/mikhaylov123ty/GophKeeper/internal/client/grpc"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ItemsManager struct {
	metaItems  map[string][]*models.MetaItem
	grpcClient *grpc.Client
	userID     string
}

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

func (im *ItemsManager) GetMetaData(category string) []*models.MetaItem {
	return im.metaItems[category]
}

func (im *ItemsManager) SaveMetaItem(category string, newItem *models.MetaItem) {
	im.metaItems[category] = append(im.metaItems[category], newItem)
}

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
		})
	if err != nil {
		return nil, fmt.Errorf("post item failed:  %w,", err)
	}

	return resp, err
}

func (im *ItemsManager) GetItemData(dataID string) (string, error) {
	response, err := im.grpcClient.Handlers.ItemDataHandler.GetItemData(context.Background(), &pb.GetItemDataRequest{
		DataId: dataID,
	})
	if err != nil {
		return "", fmt.Errorf("could not get text data: %w", err)
	}

	decryptedData, err := utils.DeryptData(response.Data)
	if err != nil {
		return "", fmt.Errorf("failed decrypt data: %w", err)
	}

	return string(decryptedData), nil

}

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

func (im *ItemsManager) SyncMeta() error {
	metaItems, err := im.grpcClient.Handlers.MetaDataHandler.GetMetaData(context.Background(),
		&pb.GetMetaDataRequest{UserId: im.userID})
	if err != nil {
		if e, ok := status.FromError(err); ok {
			if e.Code() == codes.NotFound {
				fmt.Println(`NOT FOUND`, e.Message())
				return nil
			} else {
				fmt.Println(e.Code(), e.Message())
			}
		} else {
			fmt.Printf("Не получилось распарсить ошибку %v", err)
			return err
		}
	}

	for _, metaItem := range metaItems.Items {
		id, err := uuid.Parse(metaItem.GetId())
		if err != nil {
			return fmt.Errorf("invalid meta item id: %s", metaItem.GetId())
		}
		im.metaItems[metaItem.DataType] = append(im.metaItems[metaItem.DataType], &models.MetaItem{
			Id:          id,
			Title:       metaItem.GetTitle(),
			Description: metaItem.GetDescription(),
			DataID:      metaItem.GetDataId(),
			Created:     metaItem.GetCreated(),
			Modified:    metaItem.GetModified(),
		})
	}

	return nil
}

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
		if v.Id == metaItemID {
			im.metaItems[category] = append(im.metaItems[category][:i], im.metaItems[category][i+1:]...)
		}
	}

	return nil
}
