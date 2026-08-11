package biz

import (
	"context"

	authv1 "ride-hailing/services/auth-service/api/auth/v1"
	"ride-hailing/services/gateway-service/internal/data"
)

type AuthUsecase struct {
	client data.AuthClient
}

func NewAuthUsecase(client data.AuthClient) *AuthUsecase {
	return &AuthUsecase{client: client}
}

func (uc *AuthUsecase) SendLoginCode(ctx context.Context, req *authv1.SendLoginCodeRequest) (*authv1.SendLoginCodeReply, error) {
	return uc.client.SendLoginCode(ctx, req)
}

func (uc *AuthUsecase) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginReply, error) {
	return uc.client.Login(ctx, req)
}

func (uc *AuthUsecase) VerifyToken(ctx context.Context, authorization string) (*authv1.VerifyTokenReply, error) {
	return uc.client.VerifyToken(ctx, authorization)
}

func (uc *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string) (*authv1.LoginReply, error) {
	return uc.client.RefreshToken(ctx, refreshToken)
}

func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) (*authv1.LogoutReply, error) {
	return uc.client.Logout(ctx, refreshToken)
}

func (uc *AuthUsecase) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	return uc.client.CheckPermission(ctx, userID, permissionCode)
}
