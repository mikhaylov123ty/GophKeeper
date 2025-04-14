package handlers

import (
	"context"
	"github.com/google/uuid"
	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (m *MetaDataHandler) PostMetaData(context.Context, *pb.PostMetaDataRequest) (*pb.PostMetaDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method PostMetaData not implemented")
}
func (m *MetaDataHandler) GetMetaData(context.Context, *pb.GetMetaDataRequest) (*pb.GetMetaDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method GetMetaData not implemented")
}
