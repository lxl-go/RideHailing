package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const serviceName = "review-service"

func newApp(httpSrv *khttp.Server, grpcSrv *kgrpc.Server, registrar registry.Registrar) *kratos.App {
	opts := []kratos.Option{
		kratos.Name(serviceName),
		kratos.Server(httpSrv, grpcSrv),
	}
	if registrar != nil {
		opts = append(opts, kratos.Registrar(registrar))
	}
	return kratos.New(opts...)
}
