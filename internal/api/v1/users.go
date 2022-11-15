package v1

import (
	"context"
	"log"
	"strings"

	db "github.com/golang-unitied-school/useragent/internal/interfaces"
	"github.com/golang-unitied-school/useragent/internal/models"
	global "github.com/golang-unitied-school/useragent/internal/pkg/utils"
	"github.com/sethvargo/go-password/password"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserAgent struct {
	UnimplementedUserAgentServer
	DBConn db.UserDataManager
}

// inner func for check user by creds
func (agent *UserAgent) findUserByEmail(email string) (bool, error) {

	if email == "" {
		return false, global.ErrorEmptyLogin
	} else {
		if global.CheckEmail(email) {
			_, err := agent.DBConn.GetByEmail(email)

			if err == global.ErrorUserNotFound {
				return false, nil
			}

			if err != nil {
				return false, err
			}

		} else {
			return false, global.ErrorInvalidEmailFormat
		}
	}

	return true, nil
}

func (agent *UserAgent) findUserByUUID(userId string) (bool, error) {
	_, err := agent.DBConn.GetById(userId)

	if err == global.ErrorUserNotFound {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (agent *UserAgent) checkPrerequisites(req *CreateUserRequest) error {

	if req.GetName() == "" || req.GetSurname() == "" {
		return status.Error(codes.InvalidArgument, global.ErrorEmptyCredentials.Error())
	}

	hasUser, err := agent.findUserByEmail(req.GetEmail())
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if hasUser {
		return status.Error(codes.AlreadyExists, global.ErrorUserExists.Error())
	}

	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, global.ErrorEmptyPass.Error())
	}

	if !global.ValidatePassword(req.GetPassword()) {
		return status.Error(codes.InvalidArgument, global.ErrorBadPassword.Error())
	}

	return nil
}

func (agent *UserAgent) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {

	if err := agent.checkPrerequisites(req); err != nil {
		return nil, err
	}

	userID, err := agent.DBConn.Create(
		req.GetName(),
		req.GetSurname(),
		req.GetEmail(),
		req.GetPassword(),
		req.GetRole())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &CreateUserResponse{UserId: userID}, nil
}

// update some user`s fields by id
func (agent *UserAgent) UpdateUser(ctx context.Context, req *ChangeUserRequest) (*emptypb.Empty, error) {

	hasUser, err := agent.findUserByUUID(req.GetUserID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !hasUser {
		return nil, status.Error(codes.NotFound, global.ErrorUserNotFound.Error())
	}

	if req.GetEmail() != "" {
		if !global.CheckEmail(req.GetEmail()) {
			return nil, status.Error(codes.InvalidArgument, global.ErrorInvalidEmailFormat.Error())
		}
	}

	err = agent.DBConn.Update(
		req.GetUserID(),
		req.GetName(),
		req.GetSurname(),
		req.GetEmail(),
		req.GetRole())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// delete user by id
func (agent *UserAgent) DeleteUser(ctx context.Context, req *DeleteUserRequest) (*emptypb.Empty, error) {
	hasUser, err := agent.findUserByUUID(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !hasUser {
		return nil, status.Error(codes.NotFound, global.ErrorUserNotFound.Error())
	}

	err = agent.DBConn.Delete(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// find user by uuid
func (agent *UserAgent) GetUserById(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {

	rowUser, err := agent.DBConn.GetById(req.GetUserId())
	if err != nil {
		if err == global.ErrorUserNotFound {
			return nil, status.Error(codes.NotFound, global.ErrorUserNotFound.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &GetUserResponse{
		UserId:    rowUser.Id.String(),
		Name:      rowUser.Name,
		Surname:   rowUser.Surname,
		Email:     rowUser.Email,
		Role:      rowUser.Role,
		CreatedAt: timestamppb.New(rowUser.CreatedAt),
		IsDeleted: rowUser.IsDeleted,
	}, nil
}

// find user by email
func (agent *UserAgent) GetUserByEmail(ctx context.Context, req *GetUserByEmailRequest) (*GetUserByEmailResponse, error) {

	if !global.CheckEmail(req.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, global.ErrorInvalidEmailFormat.Error())
	}

	rowUser, err := agent.DBConn.GetByEmail(req.GetEmail())
	if err != nil {
		if err.Error() == global.ErrorUserNotFound.Error() {
			return nil, status.Error(codes.NotFound, global.ErrorUserNotFound.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &GetUserByEmailResponse{
		UserID:    rowUser.Id.String(),
		Name:      rowUser.Name,
		Surname:   rowUser.Surname,
		Email:     rowUser.Email,
		Role:      rowUser.Role,
		CreatedAt: timestamppb.New(rowUser.CreatedAt),
		IsDeleted: rowUser.IsDeleted,
	}, nil
}

func (agent *UserAgent) CheckPassword(ctx context.Context, req *CheckPasswordRequest) (*CheckPasswordResponse, error) {

	var (
		user models.User
		err  error
	)

	if !global.CheckEmail(req.GetEmail()) {
		return nil, status.Error(codes.InvalidArgument, global.ErrorInvalidEmailFormat.Error())
	}

	user, err = agent.DBConn.GetByEmail(req.GetEmail())

	if err != nil {
		if err == global.ErrorUserNotFound {
			return nil, status.Error(codes.Unauthenticated, global.ErrorUnauthenticated.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	hash, err := agent.DBConn.GetPassword(user.Id.String())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	verify := global.ComparePasswords(req.GetPassword(), hash)

	if !verify {
		return nil, status.Error(codes.Unauthenticated, global.ErrorUnauthenticated.Error())
	}

	return &CheckPasswordResponse{Verified: true}, nil
}

func (agent *UserAgent) ChangePassword(ctx context.Context, req *ChangePasswordRequest) (*emptypb.Empty, error) {

	hasUser, err := agent.findUserByUUID(req.GetUserID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !hasUser {
		return nil, status.Error(codes.NotFound, global.ErrorUserNotFound.Error())
	}

	hash, err := agent.DBConn.GetPassword(req.GetUserID())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !global.ComparePasswords(req.OldPassword, hash) {
		return nil, status.Error(codes.FailedPrecondition, global.ErrorPasswordNotMatched.Error())
	}

	if strings.EqualFold(req.GetOldPassword(), req.GetNewPassword()) {
		return nil, status.Error(codes.FailedPrecondition, global.ErrorNewOldPassMatched.Error())
	}

	if !global.ValidatePassword(req.GetNewPassword()) {
		return nil, status.Error(codes.InvalidArgument, global.ErrorBadPassword.Error())
	}

	err = agent.DBConn.SetPassword(req.GetUserID(), req.GetNewPassword())
	if err != nil {
		if err.Error() == global.ErrorOldPassInvalid.Error() {
			return nil, status.Error(codes.InvalidArgument, global.ErrorOldPassInvalid.Error())
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &emptypb.Empty{}, nil
}

func (agent *UserAgent) ResetPassword(ctx context.Context, req *ResetPasswordRequest) (*emptypb.Empty, error) {
	newPass, err := password.Generate(10, 3, 3, false, false)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	log.Println(newPass)
	///here must be a code, which send new pass to user
	err = agent.DBConn.SetPassword(req.GetUserID(), newPass)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// TEST FUNC
func (agent *UserAgent) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	input := req.Message
	if strings.EqualFold(input, "Hello!") {
		return &HelloResponse{Message: "Hello? Guy!"}, nil
	} else if strings.EqualFold(input, "Give me error") {
		return nil, status.Error(codes.Unknown, "Get it!")
	}

	return &HelloResponse{Message: "Hey, say hi, be polite!"}, nil
}
