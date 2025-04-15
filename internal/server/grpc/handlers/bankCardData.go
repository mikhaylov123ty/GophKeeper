package handlers

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
)

type BankCardDataHandler struct {
	pb.UnimplementedBankCardHandlersServer
	bankCardDataCreator  bankCardDataCreator
	bankCardDataProvider bankCardDataProvider
}

type bankCardDataCreator interface {
	SaveBankCard(*models.BankCardData) error
}

type bankCardDataProvider interface {
	GetBankCardById(uuid.UUID) (*models.BankCardData, error)
}

func NewBankCardDataHandler(bankCardDataCreator bankCardDataCreator,
	bankCardDataProvider bankCardDataProvider) *BankCardDataHandler {
	return &BankCardDataHandler{
		bankCardDataCreator:  bankCardDataCreator,
		bankCardDataProvider: bankCardDataProvider,
	}
}

func (b *BankCardDataHandler) PostBankCardData(ctx context.Context, request *pb.PostBankCardDataRequest) (*pb.PostBankCardDataResponse, error) {
	var bankCardID uuid.UUID

	// TODO create better validation
	if request.GetCardNum() == "" || request.GetCvv() == "" || request.GetCvv() == "" {
		return nil, status.Error(codes.InvalidArgument, "empty cardNum or cvv")
	}

	if request.GetCardId() == "" {
		bankCardID = uuid.New()
	} else {
		id, err := uuid.Parse(request.GetCardId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid cardId")
		}
		bankCardID = id
	}

	cardData := models.BankCardData{
		ID:      bankCardID,
		CardNum: request.GetCardNum(),
		CVV:     request.GetCvv(),
		Expiry:  request.GetExpiry(),
	}

	if err := b.bankCardDataCreator.SaveBankCard(&cardData); err != nil {
		slog.ErrorContext(ctx, "cold not save card", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "save card")
	}

	return &pb.PostBankCardDataResponse{DataId: bankCardID.String()}, status.Errorf(codes.OK, "card registered ")
}
func (b *BankCardDataHandler) GetBankCardData(ctx context.Context, request *pb.GetBankCardDataRequest) (*pb.GetBankCardDataResponse, error) {
	cardID, err := uuid.Parse(request.GetCardId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid cardId")
	}

	cardItem, err := b.bankCardDataProvider.GetBankCardById(cardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.ErrorContext(ctx, "no text found", slog.String("error", err.Error()))
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.NotFound, "could not get card data")
	}

	return &pb.GetBankCardDataResponse{
			CardNum: cardItem.CardNum,
			Expiry:  cardItem.Expiry,
			Cvv:     cardItem.CVV},
		status.Errorf(codes.OK, "card data gathered")
}
