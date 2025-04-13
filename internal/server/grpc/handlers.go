package grpc

import (
	"context"
	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedHandlersServer
	storageCommands *StorageCommands
}

type StorageCommands struct {
	userDataProvider
}

type userDataProvider interface {
	SaveUser(*models.UserData) error
}

func NewHandler(storageCommands *StorageCommands) *Handler {
	return &Handler{
		storageCommands: storageCommands,
	}
}

func NewStorageCommands(userDataProvider userDataProvider) *StorageCommands {
	return &StorageCommands{
		userDataProvider: userDataProvider,
	}
}

func (h *Handler) PostUserData(context.Context, *pb.PostUserDataRequest) (*pb.PostUserDataResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostUserData not implemented")
}
func (h *Handler) PostTextData(context.Context, *pb.PostTextDataRequest) (*pb.PostTextDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method PostTextData not implemented")
}
func (h *Handler) GetTextData(context.Context, *pb.GetTextDataRequest) (*pb.GetTextDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method GetTextData not implemented")
}
func (h *Handler) PostBankCardData(context.Context, *pb.PostBankCardDataRequest) (*pb.PostBankCardDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method PostBankCardData not implemented")
}
func (h *Handler) GetBankCardData(context.Context, *pb.GetBankCardDataRequest) (*pb.GetBankCardDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method GetBankCardData not implemented")
}
func (h *Handler) PostMetaData(context.Context, *pb.PostMetaDataRequest) (*pb.PostMetaDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method PostMetaData not implemented")
}
func (h *Handler) GetMetaData(context.Context, *pb.GetMetaDataRequest) (*pb.GetMetaDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method GetMetaData not implemented")
}
