package handlers

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	GetBankCard(uuid.UUID) (*models.BankCardData, error)
}

func NewBankCardDataHandler(bankCardDataCreator bankCardDataCreator,
	bankCardDataProvider bankCardDataProvider) *BankCardDataHandler {
	return &BankCardDataHandler{
		bankCardDataCreator:  bankCardDataCreator,
		bankCardDataProvider: bankCardDataProvider,
	}
}

func (b *BankCardDataHandler) PostBankCardData(context.Context, *pb.PostBankCardDataRequest) (*pb.PostBankCardDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method PostBankCardData not implemented")
}
func (b *BankCardDataHandler) GetBankCardData(context.Context, *pb.GetBankCardDataRequest) (*pb.GetBankCardDataResponse, error) {

	return nil, status.Errorf(codes.Unimplemented, "method GetBankCardData not implemented")
}
