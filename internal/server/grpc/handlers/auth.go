package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"time"
)

type AuthHandler struct {
	pb.UnimplementedUserHandlersServer
	userCreator  userCreator
	userProvider userProvider
}

type userCreator interface {
	SaveUser(*models.UserData) error
}

type userProvider interface {
	GetUserByLogin(string) (*models.UserData, error)
}

func NewAuthHandler(userCreator userCreator, userProvider userProvider) *AuthHandler {
	return &AuthHandler{
		userCreator:  userCreator,
		userProvider: userProvider,
	}
}

func (a *AuthHandler) PostUserData(ctx context.Context, request *pb.PostUserDataRequest) (*pb.PostUserDataResponse, error) {
	var res pb.PostUserDataResponse
	if request.GetLogin() == "" || request.GetPassword() == "" {
		slog.Error("failed to get request parameters: login or password is empty")
		res.Error = fmt.Sprintf("login or password is empty")
		return &res, status.Error(codes.InvalidArgument, "login or password is empty")
	}

	storageUser, err := a.userProvider.GetUserByLogin(request.GetLogin())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to get user by login", slog.String("error", err.Error()))
		res.Error = err.Error()
		return &res, status.Error(codes.InvalidArgument, err.Error())
	}

	if storageUser == nil {
		storageUser = &models.UserData{
			ID:       uuid.New(),
			Login:    request.Login,
			Password: request.Password,
			Created:  time.Now(),
			Modified: time.Now(),
		}
		if err = a.userCreator.SaveUser(storageUser); err != nil {
			slog.Error("failed to save user", slog.String("error", err.Error()))
			res.Error = err.Error()
			return &res, status.Error(codes.Internal, err.Error())
		}
	} else {
		if storageUser.Password != request.Password {
			slog.Error("password not match")
			res.Error = fmt.Sprintf("login or password is incorrect")
			return &res, status.Error(codes.PermissionDenied, "login or password is incorrect")
		}
	}
	//todo create jwt
	res.UserId = storageUser.ID.String()

	return &res, nil
}
