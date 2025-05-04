package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mikhaylov123ty/GophKeeper/internal/models"
	pb "github.com/mikhaylov123ty/GophKeeper/internal/proto"
	"github.com/mikhaylov123ty/GophKeeper/internal/server/config"
)

// AuthHandler handles user authentication and implements the gRPC UserHandlersServer interface.
// It relies on userCreator to save user data and userProvider to retrieve user details.
type AuthHandler struct {
	pb.UnimplementedUserHandlersServer
	userCreator  userCreator
	userProvider userProvider
}

// userCreator defines a contract for saving user data to a storage system.
type userCreator interface {
	SaveUser(*models.UserData) error
}

// userProvider defines the contract for retrieving user information by their login credentials.
type userProvider interface {
	GetUserByLogin(string) (*models.UserData, error)
}

// NewAuthHandler initializes and returns a new instance of AuthHandler with the provided userCreator and userProvider dependencies.
func NewAuthHandler(userCreator userCreator, userProvider userProvider) *AuthHandler {
	return &AuthHandler{
		userCreator:  userCreator,
		userProvider: userProvider,
	}
}

// authClaims defines the structure for JWT claims related to authentication, including UserID and registered claims.
type authClaims struct {
	UserID string `json:"userID"`
	jwt.RegisteredClaims
}

// PostUserData processes the authentication or registration of a user based on the provided login and password details.
// It generates and returns a JWT if authentication is successful or registers the user if they don't already exist.
// An error is returned if the login or password is invalid, mismatched, or if token generation fails.
func (a *AuthHandler) PostUserData(ctx context.Context, request *pb.PostUserDataRequest) (*pb.PostUserDataResponse, error) {
	var res pb.PostUserDataResponse
	if request.GetLogin() == "" || request.GetPassword() == "" {
		slog.Error("failed to get request parameters: login or password is empty")
		res.Error = "login or password is empty"
		return &res, status.Error(codes.InvalidArgument, "login or password is empty")
	}

	storageUser, err := a.userProvider.GetUserByLogin(request.GetLogin())
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		slog.Error("failed to get user by login", slog.String("error", err.Error()))
		res.Error = err.Error()
		return &res, status.Error(codes.InvalidArgument, err.Error())
	}

	if storageUser == nil {
		pass, err := bcrypt.GenerateFromPassword([]byte(request.GetPassword()), 10)
		if err != nil {
			slog.Error("failed to generate password for user", slog.String("error", err.Error()))
			res.Error = err.Error()
			return &res, status.Error(codes.InvalidArgument, err.Error())
		}

		storageUser = &models.UserData{
			ID:       uuid.New(),
			Login:    request.Login,
			Password: string(pass),
			Created:  time.Now(),
			Modified: time.Now(),
		}
		if err = a.userCreator.SaveUser(storageUser); err != nil {
			slog.Error("failed to save user", slog.String("error", err.Error()))
			res.Error = err.Error()
			return &res, status.Error(codes.Internal, err.Error())
		}
	} else {
		if bcrypt.CompareHashAndPassword([]byte(storageUser.Password), []byte(request.Password)) != nil {
			slog.Error("password not match")
			res.Error = "login or password is incorrect"
			return &res, status.Error(codes.PermissionDenied, "login or password is incorrect")
		}
	}

	// Create the Claims
	claims := authClaims{
		UserID: storageUser.ID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(config.GetKeys().JWTKey))
	if err != nil {
		slog.Error("failed to sign token", slog.String("error", err.Error()))
		res.Error = "failed to sign token"
		return &res, status.Error(codes.PermissionDenied, "failed to sign token")
	}

	res.UserId = storageUser.ID.String()
	res.Jwt = ss

	return &res, nil
}
