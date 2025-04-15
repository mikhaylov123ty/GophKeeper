package handlers

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type TextHandler struct {
	pb.UnimplementedTextHandlersServer
	textDataCreator  textDataCreator
	textDataProvider textDataProvider
}

type textDataCreator interface {
	SaveText(string) error
}

type textDataProvider interface {
	GetText(uuid.UUID) (*models.TextData, error)
}

func NewTextHandler(textDataCreator textDataCreator, textDataProvider textDataProvider) *TextHandler {
	return &TextHandler{
		textDataCreator:  textDataCreator,
		textDataProvider: textDataProvider,
	}
}

func (t *TextHandler) PostTextData(context.Context, *pb.PostTextDataRequest) (*pb.PostTextDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method PostTextData not implemented")
}
func (t *TextHandler) GetTextData(context.Context, *pb.GetTextDataRequest) (*pb.GetTextDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method GetTextData not implemented")
}
