package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

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
	SaveText(*models.TextData) error
}

type textDataProvider interface {
	GetTextByID(uuid.UUID) (*models.TextData, error)
}

func NewTextHandler(textDataCreator textDataCreator, textDataProvider textDataProvider) *TextHandler {
	return &TextHandler{
		textDataCreator:  textDataCreator,
		textDataProvider: textDataProvider,
	}
}

func (t *TextHandler) PostTextData(ctx context.Context, request *pb.PostTextDataRequest) (*pb.PostTextDataResponse, error) {
	var textID uuid.UUID

	if request.GetText() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty text")
	}

	if request.GetTextId() == "" {
		textID = uuid.New()
	} else {
		id, err := uuid.Parse(request.GetTextId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid id %s", request.GetTextId())
		}
		textID = id
	}

	textData := models.TextData{
		ID:   textID,
		Text: request.GetText(),
	}

	if err := t.textDataCreator.SaveText(&textData); err != nil {
		slog.ErrorContext(ctx, "could not save text", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PostTextDataResponse{DataId: textID.String()}, status.Errorf(codes.OK, "text registered")
}
func (t *TextHandler) GetTextData(ctx context.Context, request *pb.GetTextDataRequest) (*pb.GetTextDataResponse, error) {
	textID, err := uuid.Parse(request.GetTextId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id %s", request.GetTextId())
	}

	textItem, err := t.textDataProvider.GetTextByID(textID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.ErrorContext(ctx, "no text found", slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, err.Error())
		}
		slog.ErrorContext(ctx, "could not get text", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetTextDataResponse{
			Text: textItem.Text},
		status.Errorf(codes.OK, "text gathered")
}
