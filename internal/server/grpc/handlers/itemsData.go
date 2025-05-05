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

	"github.com/mikhaylov123ty/GophKeeper/internal/domain"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

// ItemsDataHandler handles requests for item data and metadata operations by implementing gRPC server methods.
// It embeds an unimplemented gRPC server and utilizes injected itemDataCreator and itemDataProvider interfaces.
type ItemsDataHandler struct {
	pb.UnimplementedItemDataHandlersServer
	itemDataCreator  itemDataCreator
	itemDataProvider itemDataProvider
}

// itemDataCreator defines an interface for saving item data and associated metadata.
// It ensures atomic persistence operations.
type itemDataCreator interface {
	SaveItemData(*domain.ItemData, *domain.Meta) error
}

// itemDataProvider defines methods for retrieving item data by unique identifier.
type itemDataProvider interface {
	GetItemDataByID(uuid.UUID) (*domain.ItemData, error)
}

// NewItemsDataHandler creates a new instance of ItemsDataHandler with provided itemDataCreator and itemDataProvider dependencies.
// It initializes the handler to support operations for managing item data and metadata.
func NewItemsDataHandler(itemDataCreator itemDataCreator, itemDataProvider itemDataProvider) *ItemsDataHandler {
	return &ItemsDataHandler{
		itemDataCreator:  itemDataCreator,
		itemDataProvider: itemDataProvider,
	}
}

// PostItemData processes and stores item data and metadata provided in the request, returning a response with IDs and timestamps.
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

	metaData := domain.Meta{
		ID:          metaID,
		Title:       request.GetMetaData().Title,
		Description: request.GetMetaData().Description,
		Type:        request.GetMetaData().DataType,
		DataID:      dataID,
		UserID:      userID,
		Created:     time.Now(),
		Modified:    time.Now(),
	}

	itemData := domain.ItemData{
		ID:   dataID,
		Data: request.GetData(),
	}

	if err = h.itemDataCreator.SaveItemData(&itemData, &metaData); err != nil {
		slog.ErrorContext(ctx, "could not save data", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PostItemDataResponse{
			DataId:   dataID.String(),
			Created:  metaData.Created.Format(time.RFC3339),
			Modified: metaData.Modified.Format(time.RFC3339),
		},
		status.Errorf(codes.OK, "data registered")
}

// GetItemData retrieves item data associated with a given data ID from the request and returns it in the response.
func (h *ItemsDataHandler) GetItemData(ctx context.Context, request *pb.GetItemDataRequest) (*pb.GetItemDataResponse, error) {
	dataID, err := uuid.Parse(request.GetDataId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id %s", request.GetDataId())
	}

	item, err := h.itemDataProvider.GetItemDataByID(dataID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.ErrorContext(ctx, "no data found", slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, err.Error())
		}
		slog.ErrorContext(ctx, "could not get data", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetItemDataResponse{
			Data: item.Data},
		status.Errorf(codes.OK, "data gathered")
}
