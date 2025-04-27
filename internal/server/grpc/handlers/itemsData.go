package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type ItemsDataHandler struct {
	pb.UnimplementedItemDataHandlersServer
	itemDataCreator  itemDataCreator
	itemDataProvider itemDataProvider
}

// TODO ADD TX
type itemDataCreator interface {
	SaveItemData(*models.ItemData) error
	SaveMetaData(*models.Meta) error
}

// TODO separate metadata to interface

type itemDataProvider interface {
	GetItemDataByID(uuid.UUID) (*models.ItemData, error)
}

func NewTextHandler(itemDataCreator itemDataCreator, itemDataProvider itemDataProvider) *ItemsDataHandler {
	return &ItemsDataHandler{
		itemDataCreator:  itemDataCreator,
		itemDataProvider: itemDataProvider,
	}
}

func (h *ItemsDataHandler) PostItemData(ctx context.Context, request *pb.PostItemDataRequest) (*pb.PostItemDataResponse, error) {
	var dataID uuid.UUID

	if len(request.GetData()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty item data")
	}

	if request.GetDataId() == "" {
		dataID = uuid.New()
	} else {
		id, err := uuid.Parse(request.GetDataId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid id %s", request.GetDataId())
		}
		dataID = id
	}

	if request.GetMetaData() == nil {
		return nil, status.Error(codes.InvalidArgument, "empty meta data")
	}

	metaID, err := uuid.Parse(request.GetMetaData().Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id %s", request.GetMetaData().Id)
	}

	userID, err := uuid.Parse(request.GetMetaData().UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id %s", request.GetMetaData().UserId)
	}

	metaData := models.Meta{
		ID:          metaID,
		Title:       request.GetMetaData().Title,
		Description: request.GetMetaData().Description,
		Type:        request.GetMetaData().DataType,
		DataID:      dataID,
		UserID:      userID,
		Created:     time.Now(), // Current time
		Modified:    time.Now(), // Current time
	}

	itemData := models.ItemData{
		ID:   dataID,
		Data: request.GetData(),
	}

	// TODO open TX
	if err = h.itemDataCreator.SaveItemData(&itemData); err != nil {
		slog.ErrorContext(ctx, "could not save text", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = h.itemDataCreator.SaveMetaData(&metaData); err != nil {
		slog.ErrorContext(ctx, "could not save meta data", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PostItemDataResponse{
			DataId:   dataID.String(),
			Created:  metaData.Created.Format(time.RFC3339),
			Modified: metaData.Modified.Format(time.RFC3339),
		},
		status.Errorf(codes.OK, "text registered")
}
func (h *ItemsDataHandler) GetItemData(ctx context.Context, request *pb.GetItemDataRequest) (*pb.GetItemDataResponse, error) {
	dataID, err := uuid.Parse(request.GetDataId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id %s", request.GetDataId())
	}

	item, err := h.itemDataProvider.GetItemDataByID(dataID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.ErrorContext(ctx, "no text found", slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, err.Error())
		}
		slog.ErrorContext(ctx, "could not get text", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetItemDataResponse{
			Data: item.Data},
		status.Errorf(codes.OK, "text gathered")
}
