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

type TextHandler struct {
	pb.UnimplementedTextHandlersServer
	textDataCreator  textDataCreator
	textDataProvider textDataProvider
}

// TODO ADD TX
type textDataCreator interface {
	SaveText(*models.TextData) error
	SaveMetaData(*models.Meta) error
}

// TODO separate metadata to interface

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
		DataID:      textID,
		UserID:      userID,
		Created:     time.Now(), // Current time
		Modified:    time.Now(), // Current time
	}

	textData := models.TextData{
		ID:   textID,
		Text: request.GetText(),
	}

	// TODO open TX
	if err = t.textDataCreator.SaveText(&textData); err != nil {
		slog.ErrorContext(ctx, "could not save text", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err = t.textDataCreator.SaveMetaData(&metaData); err != nil {
		slog.ErrorContext(ctx, "could not save meta data", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.PostTextDataResponse{
			DataId:   textID.String(),
			Created:  metaData.Created.Format(time.RFC3339),
			Modified: metaData.Modified.Format(time.RFC3339),
		},
		status.Errorf(codes.OK, "text registered")
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
