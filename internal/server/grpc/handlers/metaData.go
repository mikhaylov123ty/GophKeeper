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

type MetaDataHandler struct {
	pb.UnimplementedMetaDataHandlersServer
	metaDataProvider metaDataProvider
}

type metaDataProvider interface {
	GetMetaDataByUser(uuid.UUID) ([]*models.Meta, error)
}

func NewMetaDataHandler(metaDataProvider metaDataProvider) *MetaDataHandler {
	return &MetaDataHandler{
		metaDataProvider: metaDataProvider,
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
