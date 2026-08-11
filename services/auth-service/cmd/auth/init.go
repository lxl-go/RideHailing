package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"go.uber.org/zap"

	"ride-hailing/services/auth-service/internal/biz"
	"ride-hailing/services/auth-service/internal/conf"
	"ride-hailing/services/auth-service/internal/data"
	"ride-hailing/services/auth-service/internal/server"
	"ride-hailing/services/auth-service/internal/service"
)

func initApp(cfg *conf.Bootstrap, logger *zap.Logger, registrar registry.Registrar) (*kratos.App, error) {
	db, err := data.NewDB(cfg.Data, logger)
	if err != nil {
		return nil, err
	}
	rdb, err := data.NewRedis(cfg.Data)
	if err != nil {
		return nil, err
	}
	tokens := data.NewTokenManager(cfg.Auth)
	smsSender := data.NewSMSSender(cfg.Auth)
	authOptions := data.NewAuthOptions(cfg.Auth)
	repo := data.NewAccountRepo(db, logger)
	permissions := data.NewRBACRepo(db)
	smsCodes := data.NewSMSCodeRepo(rdb, logger)
	sessions := data.NewSessionRepo(db, logger)
	uc := biz.NewAuthUsecase(logger, repo, smsCodes, sessions, smsSender, permissions, tokens, authOptions)
	authSvc := service.NewAuthService(uc)
	httpSrv := server.NewHTTPServer(cfg.Server, authSvc)
	grpcSrv := server.NewGRPCServer(cfg.Server, authSvc)
	return newApp(httpSrv, grpcSrv, registrar), nil
}
