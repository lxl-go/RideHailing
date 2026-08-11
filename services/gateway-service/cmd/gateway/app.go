package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/registry"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
)

const serviceName = "gateway-service"

func newApp(httpSrv *khttp.Server, registrar registry.Registrar) *kratos.App {
	opts := []kratos.Option{
		kratos.Name(serviceName),
		kratos.Server(httpSrv),
	}
	if registrar != nil {
		opts = append(opts, kratos.Registrar(registrar))
	}
	return kratos.New(opts...)
}
