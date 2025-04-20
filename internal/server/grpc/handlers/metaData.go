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
	//	metaDataCreator  metaDataCreator
	metaDataProvider metaDataProvider
}

//type metaDataCreator interface {
//	SaveMetaData(*models.Meta) error
//}

type metaDataProvider interface {
	GetMetaDataByUser(uuid.UUID, string) ([]*models.Meta, error)
}

func NewMetaDataHandler(metaDataProvider metaDataProvider) *MetaDataHandler {
	return &MetaDataHandler{
		//	metaDataCreator:  metaDataCreator,
		metaDataProvider: metaDataProvider,
	}
}

//	func (m *MetaDataHandler) PostMetaData(ctx context.Context, request *pb.PostMetaDataRequest) (*pb.PostMetaDataResponse, error) {
//		var metaData models.Meta
//		var metaID uuid.UUID
//
//		if request.GetId() == "" {
//			metaID = uuid.New()
//		} else {
//			id, err := uuid.Parse(request.GetId())
//			if err != nil {
//				return nil, status.Errorf(codes.InvalidArgument, "invalid id %s", request.GetId())
//			}
//			metaID = id
//		}
//
//		dataID, err := uuid.Parse(request.GetDataId())
//		if err != nil {
//			return nil, status.Errorf(codes.InvalidArgument, "invalid data_id %s", request.GetDataId())
//		}
//
//		userID, err := uuid.Parse(request.GetUserId())
//		if err != nil {
//			return nil, status.Errorf(codes.InvalidArgument, "invalid user_id %s", request.GetUserId())
//		}
//
//		metaData.ID = metaID
//		metaData.Created = time.Now()
//		metaData.Modified = time.Now()
//		metaData.Title = request.GetTitle()
//		metaData.Description = request.GetDescription()
//		metaData.Type = request.GetDataType()
//		metaData.DataID = dataID
//		metaData.UserID = userID
//
//		if err = m.metaDataCreator.SaveMetaData(&metaData); err != nil {
//			slog.ErrorContext(ctx, "could not save metaData", slog.String("error", err.Error()))
//			return nil, status.Error(codes.Internal, err.Error())
//		}
//
//		var response pb.PostMetaDataResponse
//		response.Id = metaID.String()
//
//		return &response, status.Errorf(codes.OK, "meta registered")
//	}
func (m *MetaDataHandler) GetMetaData(ctx context.Context, request *pb.GetMetaDataRequest) (*pb.GetMetaDataResponse, error) {
	userID, err := uuid.Parse(request.GetUserId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id %s", request.GetUserId())
	}

	metaDataItems, err := m.metaDataProvider.GetMetaDataByUser(userID, request.GetDataType())
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
