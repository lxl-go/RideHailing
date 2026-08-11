package service

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "ride-hailing/services/auth-service/api/auth/v1"
	"ride-hailing/services/auth-service/internal/biz"
)

type AuthService struct {
	authv1.UnimplementedAuthServiceServer
	uc *biz.AuthUsecase
}

func NewAuthService(uc *biz.AuthUsecase) *AuthService {
	return &AuthService{uc: uc}
}

func (s *AuthService) SendLoginCode(ctx context.Context, req *authv1.SendLoginCodeRequest) (*authv1.SendLoginCodeReply, error) {
	err := s.uc.SendLoginCode(ctx, biz.SendLoginCodeCommand{
		Mobile: req.Mobile,
		Role:   biz.Role(req.Role),
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.SendLoginCodeReply{Sent: true}, nil
}

func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginReply, error) {
	session, err := s.uc.Login(ctx, biz.LoginCommand{
		Principal: req.Principal,
		Role:      biz.Role(req.Role),
		Code:      req.Code,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.LoginReply{
		UserId:       session.UserID,
		Role:         string(session.Role),
		AccessToken:  session.AccessToken,
		TokenType:    session.TokenType,
		ExpiresIn:    session.ExpiresIn,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *AuthService) VerifyToken(ctx context.Context, req *authv1.VerifyTokenRequest) (*authv1.VerifyTokenReply, error) {
	claims, err := s.uc.VerifyToken(ctx, req.Authorization)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.VerifyTokenReply{
		UserId: claims.UserID,
		Role:   string(claims.Role),
		Jti:    claims.JTI,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.LoginReply, error) {
	session, err := s.uc.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.LoginReply{
		UserId:       session.UserID,
		Role:         string(session.Role),
		AccessToken:  session.AccessToken,
		TokenType:    session.TokenType,
		ExpiresIn:    session.ExpiresIn,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutReply, error) {
	if err := s.uc.Logout(ctx, req.RefreshToken); err != nil {
		return nil, mapError(err)
	}
	return &authv1.LogoutReply{LoggedOut: true}, nil
}

func (s *AuthService) CheckPermission(ctx context.Context, req *authv1.CheckPermissionRequest) (*authv1.CheckPermissionReply, error) {
	allowed, err := s.uc.CheckPermission(ctx, req.UserId, req.PermissionCode)
	if err != nil {
		return nil, mapError(err)
	}
	return &authv1.CheckPermissionReply{Allowed: allowed}, nil
}

func mapError(err error) error {
	switch {
	case errors.Is(err, biz.ErrInvalidPrincipal), errors.Is(err, biz.ErrInvalidRole):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrAccountDisabled), errors.Is(err, biz.ErrInvalidToken), errors.Is(err, biz.ErrInvalidSMSCode), errors.Is(err, biz.ErrSessionExpired), errors.Is(err, biz.ErrSessionRevoked):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, biz.ErrSMSCodeTooFrequent), errors.Is(err, biz.ErrSMSLoginLocked):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, biz.ErrSMSSendFailed), strings.Contains(err.Error(), "ihuyi sms"):
		return status.Error(codes.Internal, "验证码发送失败，请检查短信服务配置或稍后重试")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
