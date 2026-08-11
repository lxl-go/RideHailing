package service

import (
	"context"

	authv1 "ride-hailing/services/auth-service/api/auth/v1"
	"ride-hailing/services/gateway-service/internal/biz"
)

type AuthService struct {
	uc *biz.AuthUsecase
}

func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) SendLoginCode(ctx context.Context, req *authv1.SendLoginCodeRequest) (*authv1.SendLoginCodeReply, error) {
	return s.uc.SendLoginCode(ctx, req)
}

func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginReply, error) {
	return s.uc.Login(ctx, req)
}

func (s *AuthService) VerifyToken(ctx context.Context, authorization string) (*authv1.VerifyTokenReply, error) {
	return s.uc.VerifyToken(ctx, authorization)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*authv1.LoginReply, error) {
	return s.uc.RefreshToken(ctx, refreshToken)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) (*authv1.LogoutReply, error) {
	return s.uc.Logout(ctx, refreshToken)
}

func (s *AuthService) CheckPermission(ctx context.Context, userID int64, permissionCode string) (bool, error) {
	return s.uc.CheckPermission(ctx, userID, permissionCode)
}
