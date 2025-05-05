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

// MetaDataHandler provides methods to handle metadata operations such as retrieval and deletion.
// It embeds pb.UnimplementedMetaDataHandlersServer for forward compatibility.
// Utilizes metaDataProvider for fetching metadata and dataRemover for metadata removal.
type MetaDataHandler struct {
	pb.UnimplementedMetaDataHandlersServer
	metaDataProvider metaDataProvider
	dataRemover      dataRemover
}

// metaDataProvider defines an interface for retrieving metadata associated with a given user ID.
// GetMetaDataByUser retrieves metadata associated with the provided user UUID and returns a slice of Meta objects or an error.
type metaDataProvider interface {
	GetMetaDataByUser(uuid.UUID) ([]*domain.Meta, error)
}

// dataRemover defines methods to delete item and metadata by their unique identifier.
// DeleteItemDataByID removes the associated item data using a UUID.
// DeleteMetaDataByID removes the metadata associated with a UUID.
type dataRemover interface {
	DeleteItemDataByID(uuid.UUID) error
	DeleteMetaDataByID(uuid.UUID) error
}

// NewMetaDataHandler creates and initializes a new MetaDataHandler with the provided metaDataProvider and dataRemover.
func NewMetaDataHandler(metaDataProvider metaDataProvider, dataRemover dataRemover) *MetaDataHandler {
	return &MetaDataHandler{
		metaDataProvider: metaDataProvider,
		dataRemover:      dataRemover,
	}
}

func (m *MetaDataHandler) GetMetaData(ctx context.Context, request *pb.GetMetaDataRequest) (*pb.GetMetaDataResponse, error) {
	userID, err := uuid.Parse(request.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id %s", request.GetUserId())
	}

	metaDataItems, err := m.metaDataProvider.GetMetaDataByUser(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.ErrorContext(ctx, "no metaData found", slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, err.Error())
		}
		slog.ErrorContext(ctx, "could not get metaData", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoItems := make([]*pb.MetaData, len(metaDataItems))
	for i, v := range metaDataItems {
		protoItems[i] = &pb.MetaData{
			Id:          v.ID.String(),
			Title:       v.Title,
			Description: v.Description,
			DataType:    v.Type,
			DataId:      v.DataID.String(),
			UserId:      v.UserID.String(),
			Modified:    v.Modified.Format(time.RFC3339),
			Created:     v.Created.Format(time.RFC3339),
		}
	}

	return &pb.GetMetaDataResponse{
			Items: protoItems},
		status.Errorf(codes.OK, "meta gathered")
}

// DeleteMetaData removes metadata and associated data by their unique IDs parsed from the request and returns a response.
func (m *MetaDataHandler) DeleteMetaData(ctx context.Context, request *pb.DeleteMetaDataRequest) (*pb.DeleteMetaDataResponse, error) {
	metaDataID, err := uuid.Parse(request.GetMetadataId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id %s", request.GetMetadataId())
	}

	dataID, err := uuid.Parse(request.GetDataId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dataId %s", request.GetDataId())
	}

	if err = m.dataRemover.DeleteItemDataByID(dataID); err != nil {
		slog.ErrorContext(ctx, "could not delete bank card", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = m.dataRemover.DeleteMetaDataByID(metaDataID); err != nil {
		slog.ErrorContext(ctx, "could not delete metaData", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.DeleteMetaDataResponse{}, nil
}
