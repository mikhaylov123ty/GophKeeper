package handlers

import (
	"context"
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
	metaDataCreator  metaDataCreator
	metaDataProvider metaDataProvider
}

type metaDataCreator interface {
	SaveMetaData(*models.Meta) error
}

type metaDataProvider interface {
	GetMetaData(uuid.UUID) (*models.Meta, error)
}

func NewMetaDataHandler(metaDataCreator metaDataCreator, metaDataProvider metaDataProvider) *MetaDataHandler {
	return &MetaDataHandler{
		metaDataCreator:  metaDataCreator,
		metaDataProvider: metaDataProvider,
	}
}

func (m *MetaDataHandler) PostMetaData(ctx context.Context, request *pb.PostMetaDataRequest) (*pb.PostMetaDataResponse, error) {
	var metaData models.Meta
	var metaID uuid.UUID

	if request.GetId() == "" {
		metaID = uuid.New()
	} else {
		metaID = uuid.MustParse(request.GetId())
	}

	metaData.ID = metaID
	metaData.Created = time.Now()
	metaData.Modified = time.Now()
	metaData.Title = request.GetTitle()
	metaData.Description = request.GetDescription()
	metaData.Type = request.GetDataType()

	if err := m.metaDataCreator.SaveMetaData(&metaData); err != nil {
		slog.ErrorContext(ctx, "could not save metaData", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	var response pb.PostMetaDataResponse
	response.Id = metaID.String()

	return &response, status.Errorf(codes.OK, "meta registered")
}
func (m *MetaDataHandler) GetMetaData(ctx context.Context, request *pb.GetMetaDataRequest) (*pb.GetMetaDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method GetMetaData not implemented")
}
